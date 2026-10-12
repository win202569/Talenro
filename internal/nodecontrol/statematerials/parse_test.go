package statematerials

import (
	"context"
	"reflect"
	"strings"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
)

func testContext() context.Context { return context.Background() }

// Every parsed expected scalar/list participates; ignoring any field must fail.
func TestSubsetParse(t *testing.T) {
	in := subsetFixture(t)
	m, e := Evaluate(testContext(), in)
	if e != nil {
		t.Fatal(e)
	}
	m.Clear()
	for index := range in.Expected {
		var mutate func(reflect.Value, string)
		mutate = func(v reflect.Value, path string) {
			for i := 0; i < v.NumField(); i++ {
				field := v.Field(i)
				name := v.Type().Field(i).Name
				full := path + "." + name
				if name == "Ref" {
					continue
				}
				if field.Kind() == reflect.Struct {
					mutate(field, full)
					continue
				}
				// Only active format facts here; external references belong to relation tests.
				if strings.Contains(full, "Payload") {
					continue
				}
				if field.IsZero() {
					continue
				}
				fresh := subsetFixture(t)
				target := reflect.ValueOf(&fresh.Expected[index]).Elem()
				for _, part := range strings.Split(full, ".")[1:] {
					target = target.FieldByName(part)
				}
				switch target.Kind() {
				case reflect.Uint64:
					target.SetUint(target.Uint() + 1)
				case reflect.Int64:
					target.SetInt(target.Int() + 1)
				case reflect.String:
					target.SetString(strings.ReplaceAll(target.String(), "12345678", "22345678"))
				case reflect.Slice:
					if target.Len() > 0 {
						target.Index(0).Index(0).SetUint(target.Index(0).Index(0).Uint() ^ 1)
					}
				case reflect.Array:
					target.Index(0).SetUint(target.Index(0).Uint() ^ 1)
				default:
					continue
				}
				bad, err := Evaluate(testContext(), fresh)
				if err == nil || bad != nil {
					t.Fatalf("ignored %d %s", index, full)
				}
			}
		}
		mutate(reflect.ValueOf(in.Expected[index]), "")
	}
	for i := range in.Materials {
		fresh := subsetFixture(t)
		fresh.Materials[i].Raw[0] ^= 1
		expectError(t, fresh, ErrInvalidMaterial)
	}
	fresh := subsetFixture(t)
	fresh.Expected[0].Ref.Payload[0] ^= 1
	expectError(t, fresh, ErrMissingMaterial)
	fresh = subsetFixture(t)
	fresh.Expected[3].Metadata.RevokedIDs = []contracts.Digest{{1}}
	expectError(t, fresh, ErrReferenceMismatch)
}
func TestSubsetInputShape(t *testing.T) {
	if m, e := Evaluate(nil, subsetFixture(t)); m != nil || e != ErrInvalidInput {
		t.Fatal("nil ctx")
	}
	expectError(t, Input{}, ErrInvalidInput)
	for _, tc := range []struct {
		name string
		mut  func(*Input)
		want error
	}{
		{"node", func(i *Input) { i.Scope.NodeID = "bad" }, ErrInvalidInput}, {"epoch", func(i *Input) { i.Scope.Epoch = 0 }, ErrInvalidInput},
		{"kind", func(i *Input) { i.Materials[0].Ref.Kind = 255 }, ErrInvalidInput}, {"root E", func(i *Input) { i.Materials[0].Ref.Envelope[0] = 1 }, ErrInvalidInput},
		{"signed E", func(i *Input) { i.Materials[2].Ref.Envelope = contracts.Digest{} }, ErrInvalidInput},
		{"union", func(i *Input) { i.Expected[0].Desired.Generation = 1 }, ErrInvalidInput},
		{"duplicate material", func(i *Input) { i.Materials = append(i.Materials, i.Materials[0]) }, ErrDuplicateMaterial},
		{"duplicate expectation", func(i *Input) { i.Expected = append(i.Expected, i.Expected[0]) }, ErrDuplicateMaterial},
		{"missing material", func(i *Input) { i.Materials = i.Materials[1:] }, ErrMissingMaterial},
		{"missing expectation", func(i *Input) { i.Expected = i.Expected[1:] }, ErrMissingMaterial},
		{"E mismatch", func(i *Input) { i.Expected[2].Ref.Envelope[0] ^= 1 }, ErrReferenceMismatch},
		{"duplicate keys", func(i *Input) {
			i.Expected[1].Root.KeyIDs = append(i.Expected[1].Root.KeyIDs, i.Expected[1].Root.KeyIDs[0])
		}, ErrInvalidInput},
		{"slots order", func(i *Input) { i.Expected[4].Desired.SlotIDs = []string{"z", "a"} }, ErrInvalidInput},
	} {
		t.Run(tc.name, func(t *testing.T) { in := subsetFixture(t); tc.mut(&in); expectError(t, in, tc.want) })
	}
}
func TestSubsetBudgets(t *testing.T) {
	for _, tc := range []struct {
		name string
		mut  func(*Input)
	}{
		{"materials", func(i *Input) { i.Materials = make([]Material, 129) }}, {"expected", func(i *Input) { i.Expected = make([]Expectation, 129) }},
		{"rotations", func(i *Input) {
			for n := 0; n < 33; n++ {
				i.Materials = append(i.Materials, Material{Ref: MaterialRef{Kind: RootRotation}})
			}
		}},
		{"root bytes", func(i *Input) { i.Materials[0].Raw = make([]byte, 65537) }}, {"envelope bytes", func(i *Input) { i.Materials[2].Raw = make([]byte, 262145) }},
		{"total", func(i *Input) {
			i.Materials = nil
			for n := 0; n < 65; n++ {
				size := 262144
				if n == 64 {
					size = 1
				}
				i.Materials = append(i.Materials, Material{Ref: MaterialRef{Kind: Metadata}, Raw: make([]byte, size)})
			}
		}},
		{"root keys", func(i *Input) { i.Expected[0].Root.KeyIDs = make([]contracts.Digest, 6) }}, {"metadata keys", func(i *Input) { i.Expected[3].Metadata.KeyIDs = make([]contracts.Digest, 65) }},
		{"revoked", func(i *Input) { i.Expected[3].Metadata.RevokedIDs = make([]contracts.Digest, 4097) }}, {"slots", func(i *Input) { i.Expected[4].Desired.SlotIDs = make([]string, 9) }},
	} {
		t.Run(tc.name, func(t *testing.T) { in := subsetFixture(t); tc.mut(&in); expectError(t, in, ErrLimitExceeded) })
	}
	// Boundary checks isolate budget from format/shape; boundary must not report overflow.
	for _, tc := range []struct {
		name string
		mut  func(*Input)
	}{
		{"128", func(i *Input) { i.Materials = make([]Material, 128); i.Expected = make([]Expectation, 128) }},
		{"32 rotations", func(i *Input) {
			i.Materials = make([]Material, 32)
			for n := range i.Materials {
				i.Materials[n].Ref.Kind = RootRotation
			}
		}},
		{"root 64KiB", func(i *Input) { i.Materials[0].Raw = make([]byte, 65536) }},
		{"envelope 256KiB", func(i *Input) { i.Materials[2].Raw = make([]byte, 262144) }},
		{"16MiB", func(i *Input) {
			i.Materials = make([]Material, 64)
			for n := range i.Materials {
				i.Materials[n] = Material{Ref: MaterialRef{Kind: Metadata}, Raw: make([]byte, 262144)}
			}
		}},
		{"collection limits", func(i *Input) {
			i.Expected[0].Root.KeyIDs = make([]contracts.Digest, 5)
			i.Expected[3].Metadata.KeyIDs = make([]contracts.Digest, 64)
			i.Expected[3].Metadata.RevokedIDs = make([]contracts.Digest, 4096)
			i.Expected[4].Desired.SlotIDs = make([]string, 8)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := subsetFixture(t)
			tc.mut(&in)
			if e := validateInput(in); e == ErrLimitExceeded {
				t.Fatal("boundary rejected")
			}
		})
	}
}
