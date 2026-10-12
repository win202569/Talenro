package stateformat

import (
	"strings"
	"testing"
)

func presenceSchema() *wireSchema {
	return &wireSchema{kind: 'o', fields: map[string]*wireSchema{
		"a": {kind: 'b'}, "b": {kind: 's', nullable: true},
		"c": {kind: 'a', min: 0, max: 2, element: &wireSchema{kind: 'o', fields: map[string]*wireSchema{"a": {kind: 's'}}}},
	}}
}

// Removing presence or decoded-name checks must make this test fail.
func TestStrictPresence(t *testing.T) {
	for _, s := range []string{`{"a":false,"b":null,"c":[]}`, `{"a":true,"b":"1","c":[{"a":"2"}]}`} {
		if err := strictCanonical([]byte(s), 1024, presenceSchema()); err != nil {
			t.Errorf("valid %s: %v", s, err)
		}
	}
	for _, s := range []string{
		`{"b":null,"c":[]}`, `{"a":false,"c":[]}`, `{"a":false,"b":null}`, `{"a":false,"b":null,"c":[{}]}`,
		`{"a":false,"b":null,"c":[{"a":"1","a":"2"}]}`, `{"a":false,"b":null,"c":[{"a":"1","\u0061":"2"}]}`,
		`{"a":false,"b":null,"c":[],"d":false}`, `{"A":false,"b":null,"c":[]}`,
		`{"a":null,"b":null,"c":[]}`, `{"a":false,"b":1,"c":[]}`, `{"a":false,"b":null,"c":null}`,
		`{"a":false,"b":null,"c":[]} {}`, `{"a":false,"b":null,"c":[{"a":"1"},{"a":"2"},{"a":"3"}]}`,
	} {
		t.Run(s, func(t *testing.T) {
			if err := strictCanonical([]byte(s), 1024, presenceSchema()); err != ErrInvalidFormat {
				t.Errorf("accepted %s: %v", s, err)
			}
		})
	}
}

func TestStrictUnicodeCanonical(t *testing.T) {
	schema := &wireSchema{kind: 'o', fields: map[string]*wireSchema{"a": {kind: 's'}, "b": {kind: 's'}}}
	for _, s := range []string{`{"a":"é","b":"😀"}`, `{"a":"é","b":"\\uD800"}`, `{"a":"\n","b":"\""}`} {
		if err := strictCanonical([]byte(s), 1024, schema); err != nil {
			t.Errorf("valid %s: %v", s, err)
		}
	}
	for _, s := range []string{
		`{"a":"\ud800","b":"x"}`, `{"a":"\udc00","b":"x"}`, `{"a":"\ud800\u0041","b":"x"}`,
		`{"a":"\ud83d\ude00","b":"x"}`, `{"a":"\u0061","b":"x"}`, `{"b":"x","a":"y"}`,
		` {"a":"x","b":"y"}`, "{\"a\":\"\xff\",\"b\":\"y\"}", `{"a":"x","b":"y"}` + "\n",
	} {
		if err := strictCanonical([]byte(s), 1024, schema); err != ErrInvalidFormat {
			t.Errorf("accepted %q: %v", s, err)
		}
	}
}

func TestStrictLimits(t *testing.T) {
	for _, depth := range []int{16, 17} {
		schema := &wireSchema{kind: 's'}
		for i := 0; i < depth; i++ {
			schema = &wireSchema{kind: 'a', min: 1, max: 1, element: schema}
		}
		body := []byte(strings.Repeat("[", depth) + `"x"` + strings.Repeat("]", depth))
		err := strictCanonical(body, len(body), schema)
		if (depth == 16 && err != nil) || (depth == 17 && err != ErrInvalidFormat) {
			t.Errorf("depth %d: %v", depth, err)
		}
	}
	schema := &wireSchema{kind: 's'}
	for _, limit := range []int{65536, 262144} {
		body := []byte(`"` + strings.Repeat("a", limit-2) + `"`)
		if err := strictCanonical(body, limit, schema); err != nil {
			t.Errorf("exact %d: %v", limit, err)
		}
		if err := strictCanonical(append(body, ' '), limit, schema); err != ErrInvalidFormat {
			t.Errorf("over %d: %v", limit, err)
		}
	}
	for _, body := range []string{"", `[`, `"unterminated`, `true false`} {
		if err := strictCanonical([]byte(body), 64, schema); err != ErrInvalidFormat {
			t.Errorf("malformed %q", body)
		}
	}
	if strictCanonical([]byte(`"x"`), 0, schema) != ErrInvalidFormat || strictCanonical([]byte(`"x"`), 64, nil) != ErrInvalidFormat {
		t.Fatal("invalid parameters")
	}
}
