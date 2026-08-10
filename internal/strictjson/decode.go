// Package strictjson decodes bounded JSON without exposing parser details.
package strictjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	maximumBodyBytes int64 = 64 << 10
	maximumDepth           = 16
)

var (
	// ErrInvalidArgument reports an unusable reader, target, or byte limit.
	ErrInvalidArgument = errors.New("strictjson: invalid argument")
	// ErrTooLarge reports a body exceeding the configured byte limit.
	ErrTooLarge = errors.New("strictjson: body too large")
	// ErrInvalidUTF8 reports a body that is not valid UTF-8.
	ErrInvalidUTF8 = errors.New("strictjson: invalid UTF-8")
	// ErrInvalidJSON reports malformed JSON or a typed decoding mismatch.
	ErrInvalidJSON = errors.New("strictjson: invalid JSON")
	// ErrDuplicateMember reports a repeated member in any object.
	ErrDuplicateMember = errors.New("strictjson: duplicate member")
	// ErrTooDeep reports JSON nested beyond the supported depth.
	ErrTooDeep = errors.New("strictjson: nesting too deep")
	// ErrUnknownMember reports a member absent from the typed target.
	ErrUnknownMember = errors.New("strictjson: unknown member")
	// ErrTrailingData reports a second JSON value after the first.
	ErrTrailingData = errors.New("strictjson: trailing data")
)

// Decode reads and validates exactly one bounded JSON value into target.
func Decode(reader io.Reader, maxBytes int64, target any) error {
	if isNil(reader) || maxBytes <= 0 || maxBytes > maximumBodyBytes || !validTarget(target) {
		return ErrInvalidArgument
	}

	body, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return ErrInvalidJSON
	}
	if int64(len(body)) > maxBytes {
		return ErrTooLarge
	}
	if !utf8.Valid(body) {
		return ErrInvalidUTF8
	}
	if err := validateStructure(body, reflect.TypeOf(target).Elem()); err != nil {
		return err
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		if strings.HasPrefix(err.Error(), "json: unknown field ") {
			return ErrUnknownMember
		}
		return ErrInvalidJSON
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return ErrTrailingData
		}
		return ErrInvalidJSON
	}
	return nil
}

type structureValidator struct {
	decoder *json.Decoder
	schemas map[reflect.Type]structSchema
}

type structSchema struct {
	fields map[string]reflect.Type
}

type fieldCandidate struct {
	typeOf reflect.Type
	depth  int
	tagged bool
}

type embeddedType struct {
	typeOf reflect.Type
	depth  int
}

var jsonUnmarshalerType = reflect.TypeFor[json.Unmarshaler]()

func validateStructure(body []byte, targetType reflect.Type) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	validator := structureValidator{
		decoder: decoder,
		schemas: make(map[reflect.Type]structSchema),
	}
	first, err := validator.decoder.Token()
	if err != nil {
		return ErrInvalidJSON
	}
	if err := validator.walkValue(first, 0, targetType); err != nil {
		return err
	}
	_, err = validator.decoder.Token()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return ErrInvalidJSON
	}
	return ErrTrailingData
}

func (validator *structureValidator) walkValue(token json.Token, depth int, expectedType reflect.Type) error {
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}
	if delimiter != '{' && delimiter != '[' {
		return ErrInvalidJSON
	}
	if depth >= maximumDepth {
		return ErrTooDeep
	}

	if delimiter == '{' {
		schema, mapElementType, validatesMembers := validator.objectExpectation(expectedType)
		members := make(map[string]struct{})
		seenFields := make(map[string]struct{})
		for validator.decoder.More() {
			memberToken, err := validator.decoder.Token()
			if err != nil {
				return ErrInvalidJSON
			}
			member, ok := memberToken.(string)
			if !ok {
				return ErrInvalidJSON
			}
			if _, exists := members[member]; exists {
				return ErrDuplicateMember
			}
			members[member] = struct{}{}

			memberType := mapElementType
			if validatesMembers {
				acceptedType, exists := schema.fields[member]
				if !exists {
					if canonical, unique := schema.foldedMatch(member); unique {
						if _, alreadySeen := seenFields[canonical]; alreadySeen {
							return ErrDuplicateMember
						}
					}
					return ErrUnknownMember
				}
				seenFields[member] = struct{}{}
				memberType = acceptedType
			}

			valueToken, err := validator.decoder.Token()
			if err != nil {
				return ErrInvalidJSON
			}
			if err := validator.walkValue(valueToken, depth+1, memberType); err != nil {
				return err
			}
		}
	} else {
		elementType := arrayElementType(expectedType)
		for validator.decoder.More() {
			valueToken, err := validator.decoder.Token()
			if err != nil {
				return ErrInvalidJSON
			}
			if err := validator.walkValue(valueToken, depth+1, elementType); err != nil {
				return err
			}
		}
	}

	closing, err := validator.decoder.Token()
	if err != nil {
		return ErrInvalidJSON
	}
	closingDelimiter, ok := closing.(json.Delim)
	if !ok || (delimiter == '{' && closingDelimiter != '}') || (delimiter == '[' && closingDelimiter != ']') {
		return ErrInvalidJSON
	}
	return nil
}

func (validator *structureValidator) objectExpectation(expectedType reflect.Type) (structSchema, reflect.Type, bool) {
	typeOf, opaque := concreteJSONType(expectedType)
	if opaque || typeOf == nil {
		return structSchema{}, nil, false
	}
	if typeOf.Kind() == reflect.Struct {
		schema, exists := validator.schemas[typeOf]
		if !exists {
			schema = buildStructSchema(typeOf)
			validator.schemas[typeOf] = schema
		}
		return schema, nil, true
	}
	if typeOf.Kind() == reflect.Map {
		return structSchema{}, typeOf.Elem(), false
	}
	return structSchema{}, nil, false
}

func arrayElementType(expectedType reflect.Type) reflect.Type {
	typeOf, opaque := concreteJSONType(expectedType)
	if opaque || typeOf == nil {
		return nil
	}
	if typeOf.Kind() == reflect.Array || typeOf.Kind() == reflect.Slice {
		return typeOf.Elem()
	}
	return nil
}

func concreteJSONType(typeOf reflect.Type) (reflect.Type, bool) {
	for typeOf != nil {
		if typeOf.Implements(jsonUnmarshalerType) {
			return nil, true
		}
		if typeOf.Kind() != reflect.Pointer && reflect.PointerTo(typeOf).Implements(jsonUnmarshalerType) {
			return nil, true
		}
		if typeOf.Kind() != reflect.Pointer {
			return typeOf, false
		}
		typeOf = typeOf.Elem()
	}
	return nil, false
}

func buildStructSchema(root reflect.Type) structSchema {
	candidates := make(map[string][]fieldCandidate)
	queue := []embeddedType{{typeOf: root}}
	visitedDepth := make(map[reflect.Type]int)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if previousDepth, visited := visitedDepth[current.typeOf]; visited && previousDepth < current.depth {
			continue
		}
		if previousDepth, visited := visitedDepth[current.typeOf]; !visited || current.depth < previousDepth {
			visitedDepth[current.typeOf] = current.depth
		}

		for index := range current.typeOf.NumField() {
			field := current.typeOf.Field(index)
			fieldType := field.Type
			embeddedBase := fieldType
			if embeddedBase.Kind() == reflect.Pointer {
				embeddedBase = embeddedBase.Elem()
			}
			if field.Anonymous {
				if !field.IsExported() && embeddedBase.Kind() != reflect.Struct {
					continue
				}
			} else if !field.IsExported() {
				continue
			}

			tag := field.Tag.Get("json")
			if tag == "-" {
				continue
			}
			name, _, _ := strings.Cut(tag, ",")
			if name != "" && !validJSONTag(name) {
				name = ""
			}
			tagged := name != ""
			if name == "" {
				name = field.Name
			}

			if tagged || !field.Anonymous || embeddedBase.Kind() != reflect.Struct {
				candidates[name] = append(candidates[name], fieldCandidate{
					typeOf: fieldType,
					depth:  current.depth,
					tagged: tagged,
				})
				continue
			}
			queue = append(queue, embeddedType{typeOf: embeddedBase, depth: current.depth + 1})
		}
	}

	fields := make(map[string]reflect.Type, len(candidates))
	for name, namedCandidates := range candidates {
		minimumDepth := namedCandidates[0].depth
		for _, candidate := range namedCandidates[1:] {
			if candidate.depth < minimumDepth {
				minimumDepth = candidate.depth
			}
		}
		atDepth := make([]fieldCandidate, 0, len(namedCandidates))
		for _, candidate := range namedCandidates {
			if candidate.depth == minimumDepth {
				atDepth = append(atDepth, candidate)
			}
		}
		selected, ok := dominantCandidate(atDepth)
		if ok {
			fields[name] = selected.typeOf
		}
	}
	return structSchema{fields: fields}
}

func dominantCandidate(candidates []fieldCandidate) (fieldCandidate, bool) {
	if len(candidates) == 1 {
		return candidates[0], true
	}
	var tagged fieldCandidate
	taggedCount := 0
	for _, candidate := range candidates {
		if candidate.tagged {
			tagged = candidate
			taggedCount++
		}
	}
	if taggedCount == 1 {
		return tagged, true
	}
	return fieldCandidate{}, false
}

func (schema structSchema) foldedMatch(member string) (string, bool) {
	match := ""
	for field := range schema.fields {
		if strings.EqualFold(field, member) {
			if match != "" {
				return "", false
			}
			match = field
		}
	}
	return match, match != ""
}

func validJSONTag(name string) bool {
	if name == "" {
		return false
	}
	for _, character := range name {
		if unicode.IsLetter(character) || unicode.IsDigit(character) || strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", character) {
			continue
		}
		return false
	}
	return true
}

func validTarget(target any) bool {
	if target == nil {
		return false
	}
	value := reflect.ValueOf(target)
	return value.Kind() == reflect.Pointer && !value.IsNil()
}

func isNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	// Only nil-capable kinds can represent a typed-nil interface value.
	//nolint:exhaustive
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
