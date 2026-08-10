package strictjson_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"talenro.local/platform/internal/strictjson"
)

const maxBodyBytes int64 = 64 << 10

type request struct {
	Name   string `json:"name"`
	Nested struct {
		Enabled bool `json:"enabled"`
	} `json:"nested"`
}

func TestDecodeRejectsInvalidBodiesWithFiniteErrors(t *testing.T) {
	t.Parallel()

	tooLarge := append([]byte(`{"name":"`), bytes.Repeat([]byte{'a'}, int(maxBodyBytes))...)
	tests := []struct {
		name string
		body []byte
		want error
	}{
		{name: "unknown member", body: []byte(`{"name":"ok","SECRET_CANARY_FIELD":true}`), want: strictjson.ErrUnknownMember},
		{name: "duplicate root member", body: []byte(`{"name":"first","name":"second"}`), want: strictjson.ErrDuplicateMember},
		{name: "duplicate nested member", body: []byte(`{"name":"ok","nested":{"enabled":true,"enabled":false}}`), want: strictjson.ErrDuplicateMember},
		{name: "invalid UTF-8", body: []byte{'{', '"', 'n', 'a', 'm', 'e', '"', ':', '"', 0xff, '"', '}'}, want: strictjson.ErrInvalidUTF8},
		{name: "two JSON values", body: []byte(`{"name":"first"}{"name":"second"}`), want: strictjson.ErrTrailingData},
		{name: "larger than limit", body: tooLarge, want: strictjson.ErrTooLarge},
		{name: "deeper than sixteen", body: []byte(strings.Repeat("[", 17) + "0" + strings.Repeat("]", 17)), want: strictjson.ErrTooDeep},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var target request
			err := strictjson.Decode(bytes.NewReader(test.body), maxBodyBytes, &target)
			if !errors.Is(err, test.want) {
				t.Fatalf("Decode() error = %v, want sentinel %v", err, test.want)
			}
			for _, forbidden := range [][]byte{test.body, []byte("name"), []byte("SECRET_CANARY_FIELD"), []byte("first")} {
				if len(forbidden) >= 4 && bytes.Contains([]byte(err.Error()), forbidden) {
					t.Fatalf("Decode() error disclosed input %q", forbidden)
				}
			}
		})
	}
}

func TestDecodeAcceptsOneObjectAtBodyLimitAndDepthLimit(t *testing.T) {
	t.Parallel()

	const envelopeBytes = len(`{"name":""}`)
	body := []byte(`{"name":"` + strings.Repeat("a", int(maxBodyBytes)-envelopeBytes) + `"}`)
	if int64(len(body)) != maxBodyBytes {
		t.Fatalf("fixture length = %d, want %d", len(body), maxBodyBytes)
	}
	var target request
	if err := strictjson.Decode(bytes.NewReader(body), maxBodyBytes, &target); err != nil {
		t.Fatalf("Decode() at byte limit error = %v", err)
	}
	if len(target.Name) != int(maxBodyBytes)-envelopeBytes {
		t.Fatalf("decoded name length = %d", len(target.Name))
	}

	depthBody := []byte(strings.Repeat("[", 16) + "0" + strings.Repeat("]", 16))
	var scalar any
	if err := strictjson.Decode(bytes.NewReader(depthBody), maxBodyBytes, &scalar); err != nil {
		t.Fatalf("Decode() at depth limit error = %v", err)
	}
}

func TestDecodeRequiresUsableReaderLimitAndTarget(t *testing.T) {
	t.Parallel()

	valid := strings.NewReader(`{"name":"ok"}`)
	tests := []struct {
		name   string
		reader io.Reader
		max    int64
		target any
	}{
		{name: "nil reader", reader: nil, max: maxBodyBytes, target: &request{}},
		{name: "zero max", reader: valid, max: 0, target: &request{}},
		{name: "oversized max", reader: valid, max: maxBodyBytes + 1, target: &request{}},
		{name: "nil target", reader: valid, max: maxBodyBytes, target: nil},
		{name: "typed nil target", reader: valid, max: maxBodyBytes, target: (*request)(nil)},
		{name: "non-pointer target", reader: valid, max: maxBodyBytes, target: request{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := strictjson.Decode(test.reader, test.max, test.target); !errors.Is(err, strictjson.ErrInvalidArgument) {
				t.Fatalf("Decode() error = %v, want ErrInvalidArgument", err)
			}
		})
	}
}

func TestDecodeSupportsTypedScalarWithExactEOF(t *testing.T) {
	t.Parallel()

	var target string
	if err := strictjson.Decode(strings.NewReader(`"value"`), maxBodyBytes, &target); err != nil {
		t.Fatalf("Decode() scalar error = %v", err)
	}
	if target != "value" {
		t.Fatalf("Decode() scalar = %q", target)
	}
	if err := strictjson.Decode(strings.NewReader(`"value" true`), maxBodyBytes, &target); !errors.Is(err, strictjson.ErrTrailingData) {
		t.Fatalf("Decode() trailing error = %v, want ErrTrailingData", err)
	}
}
