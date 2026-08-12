package trust

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"reflect"
	"strconv"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/gowebpki/jcs"
	"talenro.local/platform/internal/strictjson"
)

// CanonicalizeJSON strictly validates one bounded typed JSON value before JCS
// transformation and then strictly decodes the exact canonical output again.
func CanonicalizeJSON(body []byte, target any) ([]byte, error) {
	if len(body) < 1 || len(body) > maximumCanonicalPayloadBytes || !validDecodeTarget(target) {
		return nil, ErrInvalidArgument
	}
	if err := strictjson.Decode(bytes.NewReader(body), maximumCanonicalPayloadBytes, target); err != nil {
		return nil, ErrInvalidJSON
	}
	if !validIJSON(body) {
		return nil, ErrInvalidJSON
	}
	canonical, err := jcs.Transform(body)
	if err != nil || len(canonical) < 1 || len(canonical) > maximumCanonicalPayloadBytes {
		return nil, ErrInvalidJSON
	}
	redecoded := reflect.New(reflect.TypeOf(target).Elem()).Interface()
	if err := strictjson.Decode(bytes.NewReader(canonical), maximumCanonicalPayloadBytes, redecoded); err != nil {
		return nil, ErrInvalidJSON
	}
	return bytes.Clone(canonical), nil
}

// CanonicalizePayloadV1 validates and RFC 8785-canonicalizes one exact payload.
func CanonicalizePayloadV1(value PayloadV1) ([]byte, error) {
	if err := ValidatePayloadV1(value); err != nil {
		return nil, err
	}
	body, err := json.Marshal(value)
	if err != nil {
		return nil, ErrInvalidPayload
	}
	var decoded PayloadV1
	canonical, err := CanonicalizeJSON(body, &decoded)
	if err != nil || ValidatePayloadV1(decoded) != nil {
		return nil, ErrInvalidPayload
	}
	return canonical, nil
}

func validDecodeTarget(target any) bool {
	if target == nil {
		return false
	}
	value := reflect.ValueOf(target)
	return value.Kind() == reflect.Pointer && !value.IsNil()
}

func validIJSON(body []byte) bool {
	if !utf8.Valid(body) || !validUnicodeEscapes(body) {
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	for {
		token, err := decoder.Token()
		if err != nil {
			return errors.Is(err, io.EOF)
		}
		if number, ok := token.(json.Number); ok {
			value, parseErr := strconv.ParseFloat(string(number), 64)
			if parseErr != nil || math.IsInf(value, 0) || math.IsNaN(value) {
				return false
			}
		}
	}
}

func validUnicodeEscapes(body []byte) bool {
	inString := false
	escaped := false
	for index := 0; index < len(body); index++ {
		character := body[index]
		if !inString {
			if character == '"' {
				inString = true
			}
			continue
		}
		if escaped {
			escaped = false
			if character != 'u' {
				continue
			}
			first, ok := decodeHexQuad(body, index+1)
			if !ok {
				return false
			}
			index += 4
			if first >= 0xdc00 && first <= 0xdfff {
				return false
			}
			if first < 0xd800 || first > 0xdbff {
				continue
			}
			if index+6 >= len(body) || body[index+1] != '\\' || body[index+2] != 'u' {
				return false
			}
			second, secondOK := decodeHexQuad(body, index+3)
			if !secondOK || second < 0xdc00 || second > 0xdfff || utf16.DecodeRune(rune(first), rune(second)) == utf8.RuneError {
				return false
			}
			index += 6
			continue
		}
		if character == '\\' {
			escaped = true
			continue
		}
		if character == '"' {
			inString = false
		}
	}
	return !inString && !escaped
}

func decodeHexQuad(body []byte, start int) (uint16, bool) {
	if start < 0 || start+4 > len(body) {
		return 0, false
	}
	var value uint16
	for _, character := range body[start : start+4] {
		value <<= 4
		switch {
		case character >= '0' && character <= '9':
			value += uint16(character - '0')
		case character >= 'a' && character <= 'f':
			value += uint16(character-'a') + 10
		case character >= 'A' && character <= 'F':
			value += uint16(character-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}
