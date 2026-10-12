package statebinding

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/statecomposition"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

// These tests catch omitted pre-copy budgets, borrowed slices, and clearing caller data.
func TestBindingBudget(t *testing.T) {
	for _, side := range []string{"Previous", "Proposed.Materials"} {
		for _, field := range []string{"Materials", "Expected"} {
			budgetSlice(t, side+"."+field, 128)
		}
		for path, limit := range map[string]int{
			"Root.KeyIDs": 5, "Metadata.KeyIDs": 64, "Metadata.RevokedIDs": 4096, "Desired.SlotIDs": 8,
		} {
			t.Run(side+"/inactive/"+path, func(t *testing.T) {
				for _, n := range []int{limit, limit + 1} {
					in := budgetInput()
					table := inputField(reflect.ValueOf(&in).Elem(), side)
					expected := table.FieldByName("Expected")
					expected.Set(reflect.MakeSlice(expected.Type(), 1, 1))
					v := inputField(expected.Index(0), path)
					v.Set(reflect.MakeSlice(v.Type(), n, n))
					want := error(nil)
					if n > limit {
						want = ErrLimitExceeded
					}
					if err := checkBudget(in); err != want {
						t.Fatalf("%d: %v, want %v", n, err, want)
					}
				}
			})
		}
		for _, field := range []string{"Materials", "Expected"} {
			t.Run(side+"/rotation/"+field, func(t *testing.T) {
				for _, n := range []int{32, 33} {
					in := budgetInput()
					v := inputField(reflect.ValueOf(&in).Elem(), side+"."+field)
					v.Set(reflect.MakeSlice(v.Type(), n, n))
					for i := 0; i < n; i++ {
						v.Index(i).FieldByName("Ref").FieldByName("Kind").SetUint(uint64(statematerials.RootRotation))
					}
					want := error(nil)
					if n > 32 {
						want = ErrLimitExceeded
					}
					if err := checkBudget(in); err != want {
						t.Fatalf("%d: %v", n, err)
					}
				}
			})
		}
		for _, tc := range []struct {
			kind  statematerials.MaterialKind
			limit int
		}{{statematerials.RootSet, 65536}, {statematerials.Metadata, 262144}} {
			for _, n := range []int{tc.limit, tc.limit + 1} {
				in := budgetInput()
				v := inputField(reflect.ValueOf(&in).Elem(), side+".Materials")
				v.Set(reflect.ValueOf([]statematerials.Material{{Ref: statematerials.MaterialRef{Kind: tc.kind}, Raw: make([]byte, n)}}))
				want := error(nil)
				if n > tc.limit {
					want = ErrLimitExceeded
				}
				if err := checkBudget(in); err != want {
					t.Fatalf("%s single %d: %v", side, n, err)
				}
			}
		}
		t.Run(side+"/table_raw", func(t *testing.T) {
			raw := make([]byte, 262144) // Count repeated aliases once per occurrence.
			for _, n := range []int{64, 65} {
				in := budgetInput()
				m := make([]statematerials.Material, n)
				for i := range m {
					m[i] = statematerials.Material{Ref: statematerials.MaterialRef{Kind: statematerials.Metadata}, Raw: raw}
				}
				inputField(reflect.ValueOf(&in).Elem(), side+".Materials").Set(reflect.ValueOf(m))
				want := error(nil)
				if n > 64 {
					want = ErrLimitExceeded
				}
				if err := checkBudget(in); err != want {
					t.Fatalf("%d occurrences: %v", n, err)
				}
			}
		})
		t.Run(side+"/table_raw_plus_one", func(t *testing.T) {
			raw := make([]byte, 262144)
			for _, extra := range []int{0, 1} {
				in := budgetInput()
				materials := make([]statematerials.Material, 64)
				for i := range materials {
					materials[i] = statematerials.Material{Ref: statematerials.MaterialRef{Kind: statematerials.Metadata}, Raw: raw}
				}
				if extra != 0 {
					materials = append(materials, statematerials.Material{Ref: statematerials.MaterialRef{Kind: statematerials.Metadata}, Raw: []byte{1}})
				}
				inputField(reflect.ValueOf(&in).Elem(), side+".Materials").Set(reflect.ValueOf(materials))
				want := error(nil)
				if extra != 0 {
					want = ErrLimitExceeded
				}
				if err := checkBudget(in); err != want {
					t.Fatalf("16 MiB + %d byte: got %v, want %v", extra, err, want)
				}
			}
		})
	}
	for _, side := range []string{"Main.Previous", "Main.Proposed"} {
		for _, tc := range mainCollections() {
			budgetSlice(t, side+"."+tc.path, tc.limit)
		}
	}
	budgetSlice(t, "Main.RootTransition", 32)
	budgetSlice(t, "Proposed.Target.Links", 32)
	budgetSlice(t, "Proposed.Anchor.Raw", 262144)
	t.Run("combined_raw", func(t *testing.T) {
		in := budgetInput()
		raw := make([]byte, 262144)
		for _, table := range []*statematerials.Input{&in.Previous, &in.Proposed.Materials} {
			table.Materials = make([]statematerials.Material, 64)
			for i := range table.Materials {
				table.Materials[i] = statematerials.Material{Ref: statematerials.MaterialRef{Kind: statematerials.Metadata}, Raw: raw}
			}
		}
		in.Proposed.Anchor.Raw = raw // 32 MiB + 262144, despite only one allocation.
		if err := checkBudget(in); err != nil {
			t.Fatal(err)
		}
		in.Proposed.Anchor.Raw = make([]byte, 262145)
		if err := checkBudget(in); err != ErrLimitExceeded {
			t.Fatal(err)
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	in := budgetInput()
	in.Main.RootTransition = make([]statecontract.RotationRef, 33)
	got, err := ownInput(ctx, in)
	if err != context.Canceled || !reflect.DeepEqual(got, Input{}) {
		t.Fatal("pre-cancel priority", err)
	}
}

func mainCollections() []struct {
	path  string
	limit int
} {
	return []struct {
		path  string
		limit int
	}{
		{"Root.KeyIDs", 5}, {"Metadata.KeyIDs", 64}, {"Metadata.RevokedIDs", 4096},
		{"ServerTrust.Bootstrap.Deauthorized", 128}, {"ServerTrust.Agent.Deauthorized", 128},
		{"Desired.SlotIDs", 8}, {"LKG.Entries", 8}, {"Recovery.Local", 64}, {"Recovery.Supervisor", 16},
	}
}
func budgetInput() Input { in := Input{}; in.Proposed.Anchor.Raw = []byte{1}; return in }
func inputField(v reflect.Value, path string) reflect.Value {
	start := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '.' {
			v = v.FieldByName(path[start:i])
			start = i + 1
		}
	}
	return v
}
func budgetSlice(t *testing.T, path string, limit int) {
	t.Helper()
	t.Run(path, func(t *testing.T) {
		for _, n := range []int{limit, limit + 1} {
			in := budgetInput()
			v := inputField(reflect.ValueOf(&in).Elem(), path)
			v.Set(reflect.MakeSlice(v.Type(), n, n))
			want := error(nil)
			if n > limit {
				want = ErrLimitExceeded
			}
			if err := checkBudget(in); err != want {
				t.Fatalf("%d: %v, want %v", n, err, want)
			}
		}
	})
}

func populatedInput() Input { in := Input{}; populateValue(reflect.ValueOf(&in).Elem()); return in }
func populateValue(v reflect.Value) {
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			populateValue(v.Field(i))
		}
	case reflect.Slice:
		v.Set(reflect.MakeSlice(v.Type(), 1, 1))
		populateValue(v.Index(0))
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			populateValue(v.Index(i))
		}
	case reflect.String:
		v.SetString("caller")
	case reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64:
		v.SetUint(1)
	case reflect.Int64:
		v.SetInt(1)
	case reflect.Bool:
		v.SetBool(true)
	}
}
func mutableSlices(v reflect.Value) []reflect.Value {
	var out []reflect.Value
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			out = append(out, mutableSlices(v.Field(i))...)
		}
	case reflect.Slice:
		out = append(out, reflect.ValueOf(v.Interface()))
		for i := 0; i < v.Len(); i++ {
			out = append(out, mutableSlices(v.Index(i))...)
		}
	}
	return out
}
func jsonInput(t *testing.T, in Input) []byte {
	t.Helper()
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestBindingOwnedInput(t *testing.T) {
	in := populatedInput()
	before := string(jsonInput(t, in))
	owned, err := ownInput(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if string(jsonInput(t, owned)) != before {
		t.Fatal("full input not preserved")
	}
	callerAliases := mutableSlices(reflect.ValueOf(in))
	for i := len(callerAliases) - 1; i >= 0; i-- {
		callerAliases[i].Clear()
	}
	if string(jsonInput(t, owned)) != before {
		t.Fatal("borrowed caller storage")
	}
	aliases := mutableSlices(reflect.ValueOf(owned))
	clearInput(&owned)
	clearInput(&owned)
	clearInput(nil)
	if !reflect.DeepEqual(owned, Input{}) {
		t.Fatal("owner not reset")
	}
	for _, v := range aliases {
		for i := 0; i < v.Len(); i++ {
			if !v.Index(i).IsZero() {
				t.Fatal("reachable owned data retained", v.Type())
			}
		}
	}
	in = populatedInput()
	before = string(jsonInput(t, in))
	owned, err = ownInput(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	clearInput(&owned)
	if string(jsonInput(t, in)) != before {
		t.Fatal("caller cleared")
	}
	if _, err := ownInput(nil, in); err != ErrInvalidInput {
		t.Fatal("nil context", err)
	}
	if _, err := ownInput(context.Background(), Input{}); err != ErrInvalidInput {
		t.Fatal("empty anchor", err)
	}
	normalized := budgetInput()
	normalized.Main.Previous.Desired.SlotIDs = []string{}
	normalized.Previous.Expected = []statematerials.Expectation{{Root: statematerials.RootExpectation{KeyIDs: []contracts.Digest{}}, Metadata: statematerials.MetadataExpectation{KeyIDs: []contracts.Digest{}, RevokedIDs: []contracts.Digest{}}, Desired: statematerials.DesiredExpectation{SlotIDs: []string{}}}}
	// Empty nested fields are covered without relying on a lower validator.
	normalized.Main.RootTransition = []statecontract.RotationRef{{Sequence: 2}, {Sequence: 1}}
	normalized.Proposed.Target.Links = []statecomposition.RotationLink{{Rotation: statematerials.MaterialRef{Payload: [32]byte{2}}}, {Rotation: statematerials.MaterialRef{Payload: [32]byte{1}}}}
	normalized.Previous.Materials = []statematerials.Material{{Ref: statematerials.MaterialRef{Kind: statematerials.Metadata, Payload: [32]byte{2}}}, {Ref: statematerials.MaterialRef{Kind: statematerials.RootSet, Payload: [32]byte{1}}}}
	owned, err = ownInput(context.Background(), normalized)
	if err != nil {
		t.Fatal(err)
	}
	defer clearInput(&owned)
	if owned.Main.Previous.Desired.SlotIDs != nil || owned.Previous.Expected[0].Root.KeyIDs != nil || owned.Previous.Expected[0].Metadata.KeyIDs != nil || owned.Previous.Expected[0].Metadata.RevokedIDs != nil || owned.Previous.Expected[0].Desired.SlotIDs != nil || owned.Main.RootTransition[0].Sequence != 2 || owned.Proposed.Target.Links[0].Rotation.Payload[0] != 2 || owned.Previous.Materials[0].Ref.Kind != statematerials.RootSet {
		t.Fatal("normalization scope")
	}
}

func TestBindingBorrowedAliases(t *testing.T) {
	in := populatedInput()
	in.Proposed.Materials = in.Previous
	owned, err := ownInput(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	defer clearInput(&owned)
	before := string(jsonInput(t, owned))
	owned.Previous.Materials[0].Raw[0] = 9
	owned.Previous.Expected[0].Root.KeyIDs[0][0] = 9
	if owned.Proposed.Materials.Materials[0].Raw[0] != 1 || owned.Proposed.Materials.Expected[0].Root.KeyIDs[0][0] != 1 || in.Previous.Materials[0].Raw[0] != 1 {
		t.Fatal("old/new snapshots share storage")
	}
	if string(jsonInput(t, owned)) == before {
		t.Fatal("test mutation not applied")
	}
}
