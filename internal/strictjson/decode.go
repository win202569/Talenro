// Package strictjson decodes bounded JSON without exposing parser details.
package strictjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
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
	if err := validateStructure(body); err != nil {
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

func validateStructure(body []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	first, err := decoder.Token()
	if err != nil {
		return ErrInvalidJSON
	}
	if err := walkValue(decoder, first, 0); err != nil {
		return err
	}
	_, err = decoder.Token()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return ErrInvalidJSON
	}
	return ErrTrailingData
}

func walkValue(decoder *json.Decoder, token json.Token, depth int) error {
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
		members := make(map[string]struct{})
		for decoder.More() {
			memberToken, err := decoder.Token()
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
			valueToken, err := decoder.Token()
			if err != nil {
				return ErrInvalidJSON
			}
			if err := walkValue(decoder, valueToken, depth+1); err != nil {
				return err
			}
		}
	} else {
		for decoder.More() {
			valueToken, err := decoder.Token()
			if err != nil {
				return ErrInvalidJSON
			}
			if err := walkValue(decoder, valueToken, depth+1); err != nil {
				return err
			}
		}
	}

	closing, err := decoder.Token()
	if err != nil {
		return ErrInvalidJSON
	}
	closingDelimiter, ok := closing.(json.Delim)
	if !ok || (delimiter == '{' && closingDelimiter != '}') || (delimiter == '[' && closingDelimiter != ']') {
		return ErrInvalidJSON
	}
	return nil
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
