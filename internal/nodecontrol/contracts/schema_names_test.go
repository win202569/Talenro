package contracts

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"talenro.local/platform/internal/strictjson"
)

type nameFixturePair struct {
	Schema     string `json:"schema"`
	ObjectName string `json:"object_name"`
}
type nameRegistryFixture struct {
	Schema   string            `json:"schema"`
	Coverage string            `json:"coverage"`
	Entries  []nameFixturePair `json:"entries"`
}
type nameGoldenFixture struct {
	Schema          string            `json:"schema"`
	Coverage        string            `json:"coverage"`
	ExpectedEntries []nameFixturePair `json:"expected_entries"`
	RejectedSchemas []string          `json:"rejected_schemas"`
}

const nameFixtureBase = "../../../testdata/c12/authority-v7/"

func validateNamePairs(pairs []nameFixturePair) error {
	if len(pairs) != 146 {
		return errors.New("expected exactly146 pairs")
	}
	schemas, objects := map[string]bool{}, map[string]bool{}
	for i, p := range pairs {
		if p.Schema == "" || p.ObjectName == "" || schemas[p.Schema] || objects[p.ObjectName] {
			return errors.New("empty/duplicate name")
		}
		if i > 0 && pairs[i-1].Schema >= p.Schema {
			return errors.New("unordered schema")
		}
		schemas[p.Schema], objects[p.ObjectName] = true, true
	}
	return nil
}
func decodeNameRegistry(body []byte, expected []nameFixturePair) ([]nameFixturePair, error) {
	var f nameRegistryFixture
	if err := strictjson.Decode(bytes.NewReader(body), 64<<10, &f); err != nil {
		return nil, err
	}
	if f.Schema != "talenro.c12.authority-v7.schema-name-registry.v1" || f.Coverage != "name-directory" {
		return nil, errors.New("invalid registry markers")
	}
	if err := validateNamePairs(f.Entries); err != nil {
		return nil, err
	}
	if expected != nil && !reflect.DeepEqual(f.Entries, expected) {
		return nil, errors.New("incorrect pair set")
	}
	return f.Entries, nil
}
func decodeNameGolden(body []byte) ([]nameFixturePair, []string, error) {
	var f nameGoldenFixture
	if err := strictjson.Decode(bytes.NewReader(body), 64<<10, &f); err != nil {
		return nil, nil, err
	}
	if f.Schema != "talenro.c12.authority-v7.schema-name-registry-golden.v1" || f.Coverage != "name-directory" || f.RejectedSchemas == nil {
		return nil, nil, errors.New("invalid golden markers/required array")
	}
	if err := validateNamePairs(f.ExpectedEntries); err != nil {
		return nil, nil, err
	}
	valid := map[string]bool{}
	for _, p := range f.ExpectedEntries {
		valid[p.Schema] = true
	}
	for i, s := range f.RejectedSchemas {
		if valid[s] || (i > 0 && f.RejectedSchemas[i-1] >= s) {
			return nil, nil, errors.New("invalid negative set")
		}
	}
	return f.ExpectedEntries, f.RejectedSchemas, nil
}
func readNameRegistryFixture() ([]nameFixturePair, error) {
	b, err := os.ReadFile(nameFixtureBase + "schema-name-registry.v1.json")
	if err != nil {
		return nil, err
	}
	return decodeNameRegistry(b, nil)
}
func readNameGoldenFixture() ([]nameFixturePair, []string, error) {
	b, err := os.ReadFile(nameFixtureBase + "schema-name-registry-golden.v1.json")
	if err != nil {
		return nil, nil, err
	}
	return decodeNameGolden(b)
}
func TestAuthorityV7NameFixtureInventory(t *testing.T) {
	entries, err := readNameRegistryFixture()
	if err != nil {
		t.Fatal(err)
	}
	expected, rejected, err := readNameGoldenFixture()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(entries, expected) {
		t.Fatal("independent pair sets differ")
	}
	want := map[string]bool{"": true, "AuthorityEffectCommitmentV1": true, "AuthorityEffectResolutionV1": true, "ActivationDecisionEvidenceV1": true, "talenro.c12.unknown.v1": true}
	for _, p := range expected {
		if p.ObjectName == "AuthorityEffectCommitmentV1" || p.ObjectName == "AuthorityEffectResolutionV1" || p.ObjectName == "ActivationDecisionEvidenceV1" {
			t.Fatal("historical object in v7 directory")
		}
		for _, s := range []string{strings.ToUpper(p.Schema), " " + p.Schema, p.Schema + " ", p.Schema + "\x00", strings.TrimSuffix(p.Schema, ".v1") + ".v0", strings.TrimSuffix(p.Schema, ".v1") + ".v2"} {
			want[s] = true
		}
	}
	got := map[string]bool{}
	for _, s := range rejected {
		got[s] = true
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("negative vector set incomplete")
	}
}
func TestAuthorityV7NameFixtureRejectsMalformed(t *testing.T) {
	expected, _, err := readNameGoldenFixture()
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"registry", "golden"} {
		t.Run(kind, func(t *testing.T) {
			filename := "schema-name-registry.v1.json"
			if kind == "golden" {
				filename = "schema-name-registry-golden.v1.json"
			}
			body, err := os.ReadFile(nameFixtureBase + filename)
			if err != nil {
				t.Fatal(err)
			}
			pairsField := "entries"
			if kind == "golden" {
				pairsField = "expected_entries"
			}
			check := func(b []byte) error {
				if kind == "registry" {
					_, err := decodeNameRegistry(b, expected)
					return err
				}
				pairs, _, err := decodeNameGolden(b)
				if err == nil && !reflect.DeepEqual(pairs, expected) {
					return errors.New("incorrect pair set")
				}
				return err
			}
			if err := check(body); err != nil {
				t.Fatalf("valid fixture refused: %v", err)
			}
			cases := map[string][]byte{
				"trailing":       append(append([]byte{}, body...), []byte("{}")...),
				"invalid-utf8":   {0xff},
				"oversize":       bytes.Repeat([]byte(" "), (64<<10)+1),
				"duplicate-root": []byte(strings.Replace(string(body), "\"coverage\":", "\"coverage\":\"name-directory\",\"coverage\":", 1)),
				"duplicate-pair": []byte(strings.Replace(string(body), "\"object_name\":", "\"object_name\":\"duplicate\",\"object_name\":", 1)),
			}
			mutate := func(name string, change func(map[string]any)) {
				var v map[string]any
				if err := json.Unmarshal(body, &v); err != nil {
					t.Fatal(err)
				}
				change(v)
				b, err := json.Marshal(v)
				if err != nil {
					t.Fatal(err)
				}
				cases[name] = b
			}
			var original map[string]any
			if err := json.Unmarshal(body, &original); err != nil {
				t.Fatal(err)
			}
			for key := range original {
				key := key
				mutate("missing-"+key, func(v map[string]any) { delete(v, key) })
				mutate("null-"+key, func(v map[string]any) { v[key] = nil })
				mutate("wrong-type-"+key, func(v map[string]any) { v[key] = 42 })
				mutate("case-"+key, func(v map[string]any) { v[strings.ToUpper(key)] = v[key]; delete(v, key) })
			}
			mutate("extra-root", func(v map[string]any) { v["unexpected"] = true })
			mutate("wrong-marker", func(v map[string]any) { v["schema"] = "wrong" })
			mutate("wrong-coverage", func(v map[string]any) { v["coverage"] = "metadata" })
			for _, key := range []string{"schema", "object_name"} {
				key := key
				mutate("missing-pair-"+key, func(v map[string]any) { delete(v[pairsField].([]any)[0].(map[string]any), key) })
				mutate("case-pair-"+key, func(v map[string]any) {
					p := v[pairsField].([]any)[0].(map[string]any)
					p[strings.ToUpper(key)] = p[key]
					delete(p, key)
				})
				for i, value := range []any{nil, 42, []any{}, map[string]any{}, ""} {
					value := value
					mutate(fmt.Sprintf("wrong-pair-%s-%d", key, i), func(v map[string]any) { v[pairsField].([]any)[0].(map[string]any)[key] = value })
				}
				mutate("duplicate-"+key, func(v map[string]any) {
					a := v[pairsField].([]any)
					a[1].(map[string]any)[key] = a[0].(map[string]any)[key]
				})
			}
			mutate("extra-pair", func(v map[string]any) { v[pairsField].([]any)[0].(map[string]any)["extra"] = true })
			mutate("null-pair", func(v map[string]any) { v[pairsField].([]any)[0] = nil })
			mutate("missing-entry", func(v map[string]any) { a := v[pairsField].([]any); v[pairsField] = a[1:] })
			mutate("additional-entry", func(v map[string]any) {
				a := v[pairsField].([]any)
				v[pairsField] = append(a, map[string]any{"schema": "zz-unknown.v1", "object_name": "UnknownV1"})
			})
			mutate("count-preserving-replacement", func(v map[string]any) {
				a := v[pairsField].([]any)
				a[len(a)-1] = map[string]any{"schema": "zz-unknown.v1", "object_name": "UnknownV1"}
			})
			mutate("mismatched-pairs", func(v map[string]any) {
				a := v[pairsField].([]any)
				x, y := a[0].(map[string]any), a[1].(map[string]any)
				x["object_name"], y["object_name"] = y["object_name"], x["object_name"]
			})
			mutate("unsorted", func(v map[string]any) { a := v[pairsField].([]any); a[0], a[1] = a[1], a[0] })
			names := make([]string, 0, len(cases))
			for name := range cases {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				t.Run(name, func(t *testing.T) {
					if err := check(cases[name]); err == nil {
						t.Fatal("malformed fixture accepted")
					}
				})
			}
		})
	}
}
