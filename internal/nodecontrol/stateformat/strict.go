package stateformat

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"unicode/utf8"

	"github.com/gowebpki/jcs"
)

// Static package-owned schemas describe required fields, never caller policy.
type wireSchema struct {
	kind     byte
	nullable bool
	fields   map[string]*wireSchema
	element  *wireSchema
	min, max int
}

var wireString = &wireSchema{kind: 's'}

func equalCanonicalAndClear(body, canonical []byte) bool {
	defer clear(canonical)
	return bytes.Equal(body, canonical)
}

func strictCanonical(body []byte, limit int, schema *wireSchema) error {
	if len(body) == 0 || len(body) > limit || schema == nil || !utf8.Valid(body) || !validEscapedUnicode(body) {
		return ErrInvalidFormat
	}
	d := json.NewDecoder(bytes.NewReader(body))
	d.UseNumber()
	if !walkWire(d, schema, 0) {
		return ErrInvalidFormat
	}
	if _, err := d.Token(); err != io.EOF {
		return ErrInvalidFormat
	}
	canonical, err := jcs.Transform(body)
	equal := equalCanonicalAndClear(body, canonical)
	if err != nil || !equal {
		return ErrInvalidFormat
	}
	return nil
}

func walkWire(d *json.Decoder, s *wireSchema, depth int) bool {
	if s == nil {
		return false
	}
	t, err := d.Token()
	if err != nil {
		return false
	}
	if t == nil {
		return s.nullable
	}
	switch s.kind {
	case 's':
		_, ok := t.(string)
		return ok
	case 'b':
		_, ok := t.(bool)
		return ok
	case 'o':
		if t != json.Delim('{') || depth >= 16 {
			return false
		}
		seen := make(map[string]bool, len(s.fields))
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return false
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return false
			}
			child, ok := s.fields[name]
			if !ok {
				return false
			}
			seen[name] = true
			if !walkWire(d, child, depth+1) {
				return false
			}
		}
		if len(seen) != len(s.fields) {
			return false
		}
		end, err := d.Token()
		return err == nil && end == json.Delim('}')
	case 'a':
		if t != json.Delim('[') || depth >= 16 {
			return false
		}
		n := 0
		for d.More() {
			if n >= s.max || !walkWire(d, s.element, depth+1) {
				return false
			}
			n++
		}
		end, err := d.Token()
		return err == nil && end == json.Delim(']') && n >= s.min
	default:
		return false
	}
}

// encoding/json repairs lone surrogates. Reject them before it sees a string.
// The JSON decoder handles other lexical errors; this scan only tracks quoted
// escapes, consuming a paired surrogate together so escaped backslashes remain literal.
func validEscapedUnicode(body []byte) bool {
	quoted := false
	for i := 0; i < len(body); i++ {
		if body[i] == '"' {
			quoted = !quoted
			continue
		}
		if !quoted || body[i] != '\\' {
			continue
		}
		i++
		if i >= len(body) {
			return false
		}
		if body[i] != 'u' {
			continue
		}
		if i+4 >= len(body) {
			return false
		}
		v, err := strconv.ParseUint(string(body[i+1:i+5]), 16, 16)
		if err != nil {
			return false
		}
		i += 4
		if v >= 0xdc00 && v <= 0xdfff {
			return false
		}
		if v >= 0xd800 && v <= 0xdbff {
			if i+6 >= len(body) || body[i+1] != '\\' || body[i+2] != 'u' {
				return false
			}
			low, err := strconv.ParseUint(string(body[i+3:i+7]), 16, 16)
			if err != nil || low < 0xdc00 || low > 0xdfff {
				return false
			}
			i += 6
		}
	}
	return true
}
