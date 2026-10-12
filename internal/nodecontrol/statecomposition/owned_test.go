package statecomposition

import (
	"bytes"
	"context"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"testing"
)

func TestCompositionBudget(t *testing.T) {
	base := func() Input { return Input{Anchor: AnchorInput{Raw: []byte{1}}} }
	cases := []struct {
		name  string
		limit int
		set   func(*Input, int)
	}{
		{"materials", 128, func(in *Input, n int) { in.Materials.Materials = make([]statematerials.Material, n) }},
		{"expected", 128, func(in *Input, n int) { in.Materials.Expected = make([]statematerials.Expectation, n) }},
		{"links", 32, func(in *Input, n int) { in.Target.Links = make([]RotationLink, n) }},
		{"material rotations", 32, func(in *Input, n int) {
			in.Materials.Materials = make([]statematerials.Material, n)
			for i := range in.Materials.Materials {
				in.Materials.Materials[i].Ref.Kind = statematerials.RootRotation
			}
		}},
		{"expected rotations", 32, func(in *Input, n int) {
			in.Materials.Expected = make([]statematerials.Expectation, n)
			for i := range in.Materials.Expected {
				in.Materials.Expected[i].Ref.Kind = statematerials.RootRotation
			}
		}},
		{"root raw", 65536, func(in *Input, n int) {
			in.Materials.Materials = []statematerials.Material{{Ref: statematerials.MaterialRef{Kind: statematerials.RootSet}, Raw: make([]byte, n)}}
		}},
		{"envelope raw", 262144, func(in *Input, n int) {
			in.Materials.Materials = []statematerials.Material{{Ref: statematerials.MaterialRef{Kind: statematerials.Metadata}, Raw: make([]byte, n)}}
		}},
		{"anchor raw", 262144, func(in *Input, n int) { in.Anchor.Raw = make([]byte, n) }},
		{"root keys", 5, func(in *Input, n int) {
			in.Materials.Expected = []statematerials.Expectation{{Root: statematerials.RootExpectation{KeyIDs: make([]contracts.Digest, n)}}}
		}},
		{"metadata keys", 64, func(in *Input, n int) {
			in.Materials.Expected = []statematerials.Expectation{{Metadata: statematerials.MetadataExpectation{KeyIDs: make([]contracts.Digest, n)}}}
		}},
		{"revoked", 4096, func(in *Input, n int) {
			in.Materials.Expected = []statematerials.Expectation{{Metadata: statematerials.MetadataExpectation{RevokedIDs: make([]contracts.Digest, n)}}}
		}},
		{"slots", 8, func(in *Input, n int) {
			in.Materials.Expected = []statematerials.Expectation{{Desired: statematerials.DesiredExpectation{SlotIDs: make([]string, n)}}}
		}},
		{"total raw", 16 * 1024 * 1024, func(in *Input, n int) {
			for n > 0 {
				part := n
				if part > 262144 {
					part = 262144
				}
				in.Materials.Materials = append(in.Materials.Materials, statematerials.Material{Ref: statematerials.MaterialRef{Kind: statematerials.Metadata}, Raw: make([]byte, part)})
				n -= part
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := base()
			tc.set(&in, tc.limit)
			if e := checkBudget(in); e != nil {
				t.Fatalf("equal: %v", e)
			}
			in = base()
			tc.set(&in, tc.limit+1)
			if e := checkBudget(in); e != ErrLimitExceeded {
				t.Fatalf("over: %v", e)
			}
			if got, e := ownInput(context.Background(), in); e != ErrLimitExceeded || !reflect.DeepEqual(got, Input{}) {
				t.Fatalf("copy before budget: %v", e)
			}
		})
	}
	if e := checkBudget(Input{}); e != ErrInvalidInput {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := ownInput(ctx, Input{}); e != context.Canceled {
		t.Fatal(e)
	}
	if _, e := ownInput(nil, base()); e != ErrInvalidInput {
		t.Fatal(e)
	}
}

func TestCompositionOwnedInput(t *testing.T) {
	in := Input{Anchor: AnchorInput{Raw: []byte{1, 2}}, Target: Target{Links: []RotationLink{{Rotation: statematerials.MaterialRef{Payload: contracts.Digest{9}}}}}, Materials: statematerials.Input{
		Materials: []statematerials.Material{{Ref: statematerials.MaterialRef{Kind: statematerials.Desired, Payload: contracts.Digest{2}}, Raw: []byte{3, 4}}, {Ref: statematerials.MaterialRef{Kind: statematerials.RootSet, Payload: contracts.Digest{1}}, Raw: []byte{5, 6}}},
		Expected:  []statematerials.Expectation{{Ref: statematerials.MaterialRef{Kind: statematerials.Desired, Payload: contracts.Digest{2}}, Root: statematerials.RootExpectation{KeyIDs: []contracts.Digest{{1}}}, Metadata: statematerials.MetadataExpectation{KeyIDs: []contracts.Digest{{2}}, RevokedIDs: []contracts.Digest{{3}}}, Desired: statematerials.DesiredExpectation{SlotIDs: []string{"slot-a"}}}},
	}}
	owned, e := ownInput(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	if owned.Materials.Materials[0].Ref.Kind != statematerials.RootSet {
		t.Fatal("not normalized")
	}
	in.Anchor.Raw[0] = 8
	if owned.Anchor.Raw[0] != 1 {
		t.Fatal("anchor alias")
	}
	in.Materials.Materials[0].Raw[0] = 8
	if owned.Materials.Materials[1].Raw[0] != 3 {
		t.Fatal("material alias")
	}
	in.Materials.Expected[0].Root.KeyIDs[0][0] = 8
	if owned.Materials.Expected[0].Root.KeyIDs[0][0] != 1 {
		t.Fatal("root alias")
	}
	in.Materials.Expected[0].Metadata.KeyIDs[0][0] = 8
	if owned.Materials.Expected[0].Metadata.KeyIDs[0][0] != 2 {
		t.Fatal("key alias")
	}
	in.Materials.Expected[0].Metadata.RevokedIDs[0][0] = 8
	if owned.Materials.Expected[0].Metadata.RevokedIDs[0][0] != 3 {
		t.Fatal("revoked alias")
	}
	in.Materials.Expected[0].Desired.SlotIDs[0] = "slot-b"
	if owned.Materials.Expected[0].Desired.SlotIDs[0] != "slot-a" {
		t.Fatal("slots alias")
	}
	in.Target.Links[0].Rotation.Payload[0] = 8
	if owned.Target.Links[0].Rotation.Payload[0] != 9 {
		t.Fatal("links alias")
	}
	anchor, raw := owned.Anchor.Raw, owned.Materials.Materials[0].Raw
	keys, revoked, slots, links := owned.Materials.Expected[0].Root.KeyIDs, owned.Materials.Expected[0].Metadata.RevokedIDs, owned.Materials.Expected[0].Desired.SlotIDs, owned.Target.Links
	clearInput(&owned)
	clearInput(&owned)
	clearInput(nil)
	if !reflect.DeepEqual(owned, Input{}) || !bytes.Equal(anchor, []byte{0, 0}) || !bytes.Equal(raw, []byte{0, 0}) || keys[0] != (contracts.Digest{}) || revoked[0] != (contracts.Digest{}) || slots[0] != "" || links[0] != (RotationLink{}) {
		t.Fatal("owned clear")
	}
	if in.Anchor.Raw[0] != 8 || in.Materials.Materials[1].Raw[0] != 5 {
		t.Fatal("caller cleared")
	}
	empty := Input{Anchor: AnchorInput{Raw: []byte{1}}, Target: Target{Links: []RotationLink{}}, Materials: statematerials.Input{Expected: []statematerials.Expectation{{Root: statematerials.RootExpectation{KeyIDs: []contracts.Digest{}}, Metadata: statematerials.MetadataExpectation{KeyIDs: []contracts.Digest{}, RevokedIDs: []contracts.Digest{}}, Desired: statematerials.DesiredExpectation{SlotIDs: []string{}}}}}}
	normalized, e := ownInput(context.Background(), empty)
	if e != nil {
		t.Fatal(e)
	}
	defer clearInput(&normalized)
	if normalized.Target.Links != nil || normalized.Materials.Expected[0].Root.KeyIDs != nil || normalized.Materials.Expected[0].Metadata.KeyIDs != nil || normalized.Materials.Expected[0].Metadata.RevokedIDs != nil || normalized.Materials.Expected[0].Desired.SlotIDs != nil {
		t.Fatal("empty normalization")
	}
}
