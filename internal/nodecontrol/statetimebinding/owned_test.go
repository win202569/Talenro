package statetimebinding

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"talenro.local/platform/internal/nodecontrol/statebinding"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
	"talenro.local/platform/internal/nodecontrol/statepolicybinding"
	"testing"
	"time"

	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/statecomposition"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

// These tests catch omitted pre-copy budgets, borrowed slices, and clearing caller data.
func TestTimeBindingBudget(t *testing.T) {
	TestTimeBindingPolicyBudgets(t)
	for _, side := range []string{"Policy.Binding.Previous", "Policy.Binding.Proposed.Materials"} {
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
	for _, side := range []string{"Policy.Binding.Main.Previous", "Policy.Binding.Main.Proposed"} {
		for _, tc := range mainCollections() {
			budgetSlice(t, side+"."+tc.path, tc.limit)
		}
	}
	budgetSlice(t, "Policy.Binding.Main.RootTransition", 32)
	budgetSlice(t, "Policy.Binding.Proposed.Target.Links", 32)
	budgetSlice(t, "Policy.Binding.Proposed.Anchor.Raw", 262144)
	t.Run("combined_raw", func(t *testing.T) {
		in := budgetInput()
		raw := make([]byte, 262144)
		for _, table := range []*statematerials.Input{&in.Policy.Binding.Previous, &in.Policy.Binding.Proposed.Materials} {
			table.Materials = make([]statematerials.Material, 64)
			for i := range table.Materials {
				table.Materials[i] = statematerials.Material{Ref: statematerials.MaterialRef{Kind: statematerials.Metadata}, Raw: raw}
			}
		}
		in.Policy.Binding.Proposed.Anchor.Raw = raw // 32 MiB + 262144, despite only one allocation.
		if err := checkBudget(in); err != nil {
			t.Fatal(err)
		}
		in.Policy.Binding.Proposed.Anchor.Raw = make([]byte, 262145)
		if err := checkBudget(in); err != ErrLimitExceeded {
			t.Fatal(err)
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	in := budgetInput()
	in.Policy.Binding.Main.RootTransition = make([]statecontract.RotationRef, 33)
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
func budgetInput() Input {
	in := Input{Policy: statepolicybinding.Input{Binding: statebinding.Input{}, Policy: statepolicy.ModelInput{Old: &statepolicy.StateModel{}}}}
	in.Policy.Binding.Proposed.Anchor.Raw = []byte{1}
	return in
}
func inputField(v reflect.Value, path string) reflect.Value {
	start := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '.' {
			if v.Kind() == reflect.Pointer {
				v = v.Elem()
			}
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
	if v.Type() == reflect.TypeOf(time.Time{}) {
		v.Set(reflect.ValueOf(time.Date(2020, 1, 1, 1, 0, 0, 123, time.UTC)))
		return
	}
	switch v.Kind() {
	case reflect.Pointer:
		v.Set(reflect.New(v.Type().Elem()))
		populateValue(v.Elem())
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
	if v.Type() == reflect.TypeOf(time.Time{}) {
		return nil
	}
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			out = append(out, mutableSlices(v.Elem())...)
		}
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

func TestTimeBindingOwnedInput(t *testing.T) {
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
	normalized.Policy.Binding.Main.Previous.Desired.SlotIDs = []string{}
	normalized.Policy.Binding.Previous.Expected = []statematerials.Expectation{{Root: statematerials.RootExpectation{KeyIDs: []contracts.Digest{}}, Metadata: statematerials.MetadataExpectation{KeyIDs: []contracts.Digest{}, RevokedIDs: []contracts.Digest{}}, Desired: statematerials.DesiredExpectation{SlotIDs: []string{}}}}
	// Empty nested fields are covered without relying on a lower validator.
	normalized.Policy.Binding.Main.RootTransition = []statecontract.RotationRef{{Sequence: 2}, {Sequence: 1}}
	normalized.Policy.Binding.Proposed.Target.Links = []statecomposition.RotationLink{{Rotation: statematerials.MaterialRef{Payload: [32]byte{2}}}, {Rotation: statematerials.MaterialRef{Payload: [32]byte{1}}}}
	normalized.Policy.Binding.Previous.Materials = []statematerials.Material{{Ref: statematerials.MaterialRef{Kind: statematerials.Metadata, Payload: [32]byte{2}}}, {Ref: statematerials.MaterialRef{Kind: statematerials.RootSet, Payload: [32]byte{1}}}}
	owned, err = ownInput(context.Background(), normalized)
	if err != nil {
		t.Fatal(err)
	}
	defer clearInput(&owned)
	if owned.Policy.Binding.Main.Previous.Desired.SlotIDs != nil || owned.Policy.Binding.Previous.Expected[0].Root.KeyIDs != nil || owned.Policy.Binding.Previous.Expected[0].Metadata.KeyIDs != nil || owned.Policy.Binding.Previous.Expected[0].Metadata.RevokedIDs != nil || owned.Policy.Binding.Previous.Expected[0].Desired.SlotIDs != nil || owned.Policy.Binding.Main.RootTransition[0].Sequence != 2 || owned.Policy.Binding.Proposed.Target.Links[0].Rotation.Payload[0] != 2 || owned.Policy.Binding.Previous.Materials[0].Ref.Kind != statematerials.RootSet {
		t.Fatal("normalization scope")
	}
}

func TestTimeBindingOwnedAliases(t *testing.T) {
	in := populatedInput()
	in.Policy.Binding.Previous.Materials[0].Ref.Kind = statematerials.Metadata
	in.Policy.Binding.Proposed.Materials = in.Policy.Binding.Previous
	// Same P in separate tables is not one envelope identity.
	in.Policy.Binding.Proposed.Materials.Materials = append([]statematerials.Material(nil), in.Policy.Binding.Proposed.Materials.Materials...)
	in.Policy.Binding.Proposed.Materials.Materials[0].Ref.Envelope[0] = 2
	in.Policy.Policy.Batch.AnchorEnvelope = in.Policy.Binding.Previous.Materials[0].Raw
	in.Policy.Policy.Batch.ActiveRootEnvelope = in.Policy.Policy.Batch.AnchorEnvelope
	in.Policy.Policy.Batch.Rotations = [][]byte{in.Policy.Policy.Batch.AnchorEnvelope}
	owned, err := ownInput(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	defer clearInput(&owned)
	before := string(jsonInput(t, owned))
	if owned.Policy.Binding.Previous.Materials[0].Ref.Envelope[0] != 1 || owned.Policy.Binding.Proposed.Materials.Materials[0].Ref.Envelope[0] != 2 {
		t.Fatal("cross-table envelope conflation")
	}
	owned.Policy.Binding.Previous.Materials[0].Raw[0] = 9
	owned.Policy.Binding.Previous.Expected[0].Root.KeyIDs[0][0] = 9
	if owned.Policy.Binding.Proposed.Materials.Materials[0].Raw[0] != 1 || owned.Policy.Binding.Proposed.Materials.Expected[0].Root.KeyIDs[0][0] != 1 || in.Policy.Binding.Previous.Materials[0].Raw[0] != 1 {
		t.Fatal("old/new snapshots share storage")
	}
	if owned.Policy.Policy.Batch.AnchorEnvelope[0] != 1 || owned.Policy.Policy.Batch.ActiveRootEnvelope[0] != 1 || owned.Policy.Policy.Batch.Rotations[0][0] != 1 {
		t.Fatal("batch alias borrowed")
	}
	owned.Policy.Policy.Batch.ActiveRootEnvelope[0] = 8
	if owned.Policy.Policy.Batch.AnchorEnvelope[0] != 1 || owned.Policy.Policy.Batch.Rotations[0][0] != 1 || in.Policy.Policy.Batch.AnchorEnvelope[0] != 1 {
		t.Fatal("each raw occurrence not independent")
	}
	if string(jsonInput(t, owned)) == before {
		t.Fatal("test mutation not applied")
	}
}

func TestTimeBindingNormalization(t *testing.T) {
	// Catches discarded nanoseconds, changed zero semantics, or unintended sorting.
	in := budgetInput()
	instant := time.Date(2020, 1, 1, 1, 0, 0, 123456789, time.FixedZone("offset", 3600))
	in.Policy.Policy.Time.Now = instant
	in.Policy.Policy.Time.Floor = instant
	in.Policy.Policy.Old.TimeFloor = instant
	in.Policy.Policy.Old.Keys = []statepolicy.KeyRecord{{Material: stateformat.MetadataKeyMaterial{NotBefore: instant, NotAfter: instant, RevokedAt: instant}, FirstRevokedAt: instant, LastSnapshotUntil: instant, LastCompositeDeadline: instant}}
	in.Policy.Policy.History = []statepolicy.KeyHistoryObservation{{ObservationID: contracts.Digest{2}, LastSnapshotUntil: instant, LastCompositeDeadline: instant}, {ObservationID: contracts.Digest{1}}}
	in.Policy.Policy.Old.Roots = []statepolicy.RootRecord{{Mark: statepolicy.HighWater{Version: 2}}, {Mark: statepolicy.HighWater{Version: 1}}}
	got, err := ownInput(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	defer clearInput(&got)
	if got.Policy.Policy.Time.Now.Location() != time.UTC || got.Policy.Policy.Time.Now.Nanosecond() != 123456789 || !got.Policy.Policy.Time.Now.Equal(instant) {
		t.Fatal("time changed")
	}
	if got.Policy.Policy.Old.Keys[0].Material.NotBefore.Location() != time.UTC || got.Policy.Policy.History[0].LastSnapshotUntil.Location() != time.UTC || got.Policy.Policy.History[0].ObservationID[0] != 2 || got.Policy.Policy.Old.Roots[0].Mark.Version != 2 {
		t.Fatal("nested time/order")
	}
	zero := time.Time{}.In(time.FixedZone("zero-offset", 3600))
	if normalizeTime(zero) != (time.Time{}) || normalizeTime(time.Time{}) != (time.Time{}) {
		t.Fatal("zero supplemented")
	}
	empty := budgetInput()
	empty.Policy.Policy.History = []statepolicy.KeyHistoryObservation{}
	empty.Policy.Policy.Batch.Rotations = [][]byte{}
	empty.Policy.Policy.Old.Roots = []statepolicy.RootRecord{}
	empty.Policy.Policy.Old.Keys = []statepolicy.KeyRecord{}
	empty.Policy.Policy.Old.RevokedIDs = []contracts.Digest{}
	norm, err := ownInput(context.Background(), empty)
	if err != nil {
		t.Fatal(err)
	}
	defer clearInput(&norm)
	if norm.Policy.Policy.History != nil || norm.Policy.Policy.Batch.Rotations != nil || norm.Policy.Policy.Old.Roots != nil || norm.Policy.Policy.Old.Keys != nil || norm.Policy.Policy.Old.RevokedIDs != nil {
		t.Fatal("empty normalization")
	}
}

func TestTimeBindingPolicyBudgets(t *testing.T) {
	for path, limit := range map[string]int{"Policy.Policy.Batch.Rotations": 32, "Policy.Policy.Batch.AnchorEnvelope": 262144, "Policy.Policy.Batch.ActiveRootEnvelope": 262144, "Policy.Policy.Batch.MetadataEnvelope": 262144, "Policy.Policy.Batch.DesiredEnvelope": 262144, "Policy.Policy.Old.Roots": 33, "Policy.Policy.Old.Keys": 4096, "Policy.Policy.Old.RevokedIDs": 4096, "Policy.Policy.History": 4096} {
		budgetSlice(t, path, limit)
	}
	for _, n := range []int{262144, 262145} {
		in := budgetInput()
		in.Policy.Policy.Batch.Rotations = [][]byte{make([]byte, n)}
		err := checkBudget(in)
		if (n == 262144 && err != nil) || (n > 262144 && !errors.Is(err, ErrLimitExceeded)) {
			t.Fatal("rotation raw", err)
		}
	}
	for _, n := range []int{5, 6} {
		in := budgetInput()
		in.Policy.Policy.Old.Roots = make([]statepolicy.RootRecord, 33)
		in.Policy.Policy.Old.Roots[32].Root.Keys = make([]stateformat.PublicKeyMaterial, n)
		err := checkBudget(in)
		if (n == 5 && err != nil) || (n > 5 && !errors.Is(err, ErrLimitExceeded)) {
			t.Fatal("every root keys", err)
		}
	}
	for _, tc := range []struct {
		path  string
		limit int
	}{{"Status", 8}, {"RevocationReason", 32}} {
		for _, n := range []int{tc.limit, tc.limit + 1} {
			in := budgetInput()
			in.Policy.Policy.Old.Keys = make([]statepolicy.KeyRecord, 1)
			v := reflect.ValueOf(&in.Policy.Policy.Old.Keys[0].Material).Elem().FieldByName(tc.path)
			v.SetString(strings.Repeat("a", n))
			err := checkBudget(in)
			if (n == tc.limit && err != nil) || (n > tc.limit && !errors.Is(err, ErrLimitExceeded)) {
				t.Fatal(tc.path, err)
			}
		}
	}
	for _, n := range []int{64, 65} {
		in := budgetInput()
		in.Policy.Policy.History = []statepolicy.KeyHistoryObservation{{IncidentKind: strings.Repeat("a", n)}}
		err := checkBudget(in)
		if (n == 64 && err != nil) || (n > 64 && !errors.Is(err, ErrLimitExceeded)) {
			t.Fatal("history bytes", err)
		}
	}
	raw := make([]byte, 262144)
	in := budgetInput()
	in.Policy.Binding.Proposed.Anchor.Raw = raw
	for _, table := range []*statematerials.Input{&in.Policy.Binding.Previous, &in.Policy.Binding.Proposed.Materials} {
		table.Materials = make([]statematerials.Material, 64)
		for i := range table.Materials {
			table.Materials[i] = statematerials.Material{Ref: statematerials.MaterialRef{Kind: statematerials.Metadata}, Raw: raw}
		}
	}
	in.Policy.Policy.Batch.AnchorEnvelope = raw
	in.Policy.Policy.Batch.ActiveRootEnvelope = raw
	in.Policy.Policy.Batch.MetadataEnvelope = raw
	in.Policy.Policy.Batch.DesiredEnvelope = raw
	in.Policy.Policy.Batch.Rotations = make([][]byte, 32)
	for i := range in.Policy.Policy.Batch.Rotations {
		in.Policy.Policy.Batch.Rotations[i] = raw
	}
	if err := checkBudget(in); err != nil {
		t.Fatal("occurrence total at limit", err)
	}
	in.Policy.Policy.Batch.Rotations[31] = make([]byte, 262145)
	if err := checkBudget(in); !errors.Is(err, ErrLimitExceeded) {
		t.Fatal("occurrence +1", err)
	}
	if out, err := ownInput(nil, in); !errors.Is(err, ErrInvalidInput) || !reflect.DeepEqual(out, Input{}) {
		t.Fatal("nil context", err)
	}
}

func TestTimeBindingFourMetadataAliases(t *testing.T) {
	in := timeBindingFixture(t, 1, 100)
	in.Policy.Binding.Previous = in.Policy.Binding.Proposed.Materials
	raw := material(&in.Policy.Binding.Proposed, in.Policy.Binding.Proposed.Target.Metadata).Raw
	in.Policy.Policy.Batch.MetadataEnvelope = raw
	in.Time.MetadataEnvelope = raw
	owned, e := ownInput(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer clearInput(&owned)
	ref := owned.Policy.Binding.Proposed.Target.Metadata
	a, e := findMaterial(context.Background(), owned.Policy.Binding.Previous, ref)
	if e != nil {
		t.Fatal(e)
	}
	b, e := findMaterial(context.Background(), owned.Policy.Binding.Proposed.Materials, ref)
	if e != nil {
		t.Fatal(e)
	}
	held := [][]byte{a.Raw, b.Raw, owned.Policy.Policy.Batch.MetadataEnvelope, owned.Time.MetadataEnvelope}
	before := append([]byte(nil), raw...)
	for i := range held {
		old := held[i][0]
		held[i][0] ^= 1
		for j := range held {
			if j != i && held[j][0] != old {
				t.Fatal("occurrence alias")
			}
		}
		if string(raw) != string(before) {
			t.Fatal("caller alias")
		}
		held[i][0] = old
	}
	clearInput(&owned)
	for _, r := range held {
		for _, b := range r {
			if b != 0 {
				t.Fatal("raw not cleared")
			}
		}
	}
	if string(raw) != string(before) {
		t.Fatal("caller cleared")
	}
}
