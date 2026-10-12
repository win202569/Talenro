package statecomposition

import (
	"bytes"
	"context"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"testing"
)

func checkedOwned(t *testing.T, in Input) Input {
	t.Helper()
	m, e := statematerials.Evaluate(context.Background(), in.Materials)
	if e != nil {
		t.Fatal("fixture subset", e)
	}
	m.Clear()
	owned, e := ownInput(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { clearInput(&owned) })
	return owned
}
func TestCompositionSelection(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		in := fixture(t, n)
		owned := checkedOwned(t, in)
		s, e := selectBatch(context.Background(), owned)
		if e != nil {
			t.Fatal(n, e)
		}
		if s.rootVersion != uint64(n+1) || s.batch.ExpectedEpoch != 19 || s.batch.ExpectedNodeID != "12345678-1234-4234-8234-123456789abc" || len(s.batch.Rotations) != n || s.metadataVersion != 6 || s.desiredGeneration != 7 || s.desiredSequence != 9 {
			t.Fatal("coordinates", n)
		}
		if !bytes.Equal(s.batch.DesiredEnvelope, material(&owned, owned.Target.Desired).Raw) || !bytes.Equal(s.batch.MetadataEnvelope, material(&owned, owned.Target.Metadata).Raw) || !bytes.Equal(s.batch.AnchorEnvelope, owned.Anchor.Raw) {
			t.Fatal("selected raw")
		}
		active := owned.Anchor.Raw
		for i, l := range owned.Target.Links {
			raw := material(&owned, l.Rotation).Raw
			if !bytes.Equal(s.batch.Rotations[i], raw) {
				t.Fatal("rotation raw")
			}
			active = raw
		}
		if !bytes.Equal(s.batch.ActiveRootEnvelope, active) || s.finalRoot != tailRef(owned) {
			t.Fatal("chain tail")
		}
		before := bytes.Clone(owned.Anchor.Raw)
		s.Clear()
		s.Clear()
		if !bytes.Equal(owned.Anchor.Raw, before) {
			t.Fatal("borrow cleared")
		}
	}
	in := fixture(t, 1)
	disconnected(t, &in)
	owned := checkedOwned(t, in)
	s, e := selectBatch(context.Background(), owned)
	if e != nil {
		t.Fatal(e)
	}
	s.Clear()
	in.Target.Links = make([]RotationLink, 33)
	if _, e := ownInput(context.Background(), in); e != ErrLimitExceeded {
		t.Fatal(e)
	}
}
func TestCompositionAnchor(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*Input)
		want error
	}{
		{"payload", func(in *Input) { in.Anchor.Payload[0] ^= 1 }, ErrBindingMismatch}, {"envelope", func(in *Input) { in.Anchor.Envelope[0] ^= 1 }, ErrBindingMismatch}, {"zero", func(in *Input) { in.Anchor.Payload = contracts.Digest{} }, ErrInvalidInput},
		{"scope node", func(in *Input) { in.Anchor.Scope.NodeID = "22345678-1234-4234-8234-123456789abc" }, ErrBindingMismatch}, {"scope epoch", func(in *Input) { in.Anchor.Scope.Epoch++ }, ErrBindingMismatch},
		{"malformed", func(in *Input) { in.Anchor.Raw = []byte("bad") }, ErrInvalidMaterial},
		{"high version", func(in *Input) {
			p := object(t, frozen(t, "anchor.payload.json"))
			p["root_version"] = "100"
			in.Anchor.Raw = mint(t, "anchor", p, nil)
		}, ErrInvalidMaterial},
		{"different root", func(in *Input) { in.Target.AnchorRoot = in.Target.Links[0].NewRoot }, ErrBindingMismatch},
		{"valid other anchor", func(in *Input) {
			p := object(t, frozen(t, "anchor.payload.json"))
			p["authority_sequence"] = "4"
			in.Anchor.Raw = mint(t, "anchor", p, nil)
		}, ErrBindingMismatch},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := checkedOwned(t, fixture(t, 1))
			tc.mut(&in)
			s, e := selectBatch(context.Background(), in)
			if e != tc.want || s != nil {
				t.Fatalf("%v %v", s, e)
			}
		})
	}
}
func TestCompositionTargetRefs(t *testing.T) {
	setters := []func(*Input) *statematerials.MaterialRef{func(in *Input) *statematerials.MaterialRef { return &in.Target.AnchorRoot }, func(in *Input) *statematerials.MaterialRef { return &in.Target.Links[0].Rotation }, func(in *Input) *statematerials.MaterialRef { return &in.Target.Links[0].NewRoot }, func(in *Input) *statematerials.MaterialRef { return &in.Target.Metadata }, func(in *Input) *statematerials.MaterialRef { return &in.Target.Desired }}
	for i, set := range setters {
		for _, mode := range []string{"kind", "zero", "missing", "envelope"} {
			in := checkedOwned(t, fixture(t, 1))
			r := set(&in)
			want := ErrInvalidInput
			switch mode {
			case "kind":
				r.Kind = 0
			case "zero":
				r.Payload = contracts.Digest{}
			case "missing":
				r.Payload[0] ^= 1
				want = ErrMissingMaterial
			case "envelope":
				r.Envelope[0] ^= 1
				if r.Kind != statematerials.RootSet {
					want = ErrBindingMismatch
				}
			}
			s, e := selectBatch(context.Background(), in)
			if s != nil || e != want {
				t.Fatalf("%d %s: %v", i, mode, e)
			}
		}
	}
	for _, mode := range []string{"reverse", "duplicate", "gap", "new root", "history"} {
		in := fixture(t, 2)
		history := disconnected(t, &in)
		owned := checkedOwned(t, in)
		switch mode {
		case "reverse":
			owned.Target.Links[0], owned.Target.Links[1] = owned.Target.Links[1], owned.Target.Links[0]
		case "duplicate":
			owned.Target.Links[1] = owned.Target.Links[0]
		case "gap":
			owned.Target.Links = owned.Target.Links[1:]
		case "new root":
			owned.Target.Links[0].NewRoot = owned.Target.Links[1].NewRoot
		case "history":
			owned.Target.Links = []RotationLink{history}
		}
		s, e := selectBatch(context.Background(), owned)
		if e != ErrBindingMismatch || s != nil {
			t.Fatal(mode, e)
		}
	}
	in := fixture(t, 1)
	other := compatibleMetadata(t, &in, 7, false)
	owned := checkedOwned(t, in)
	owned.Target.Metadata = other
	s, e := selectBatch(context.Background(), owned)
	if e != ErrBindingMismatch || s != nil {
		t.Fatal("declared metadata", e)
	}
	in = fixture(t, 1)
	old := in.Target.Metadata
	raw := extraShare(t, material(&in, old).Raw, "signatures")
	nr := replaceEnvelope(t, &in, old, raw)
	owned = checkedOwned(t, in)
	s, e = selectBatch(context.Background(), owned)
	if e != ErrBindingMismatch || s != nil {
		t.Fatal("same payload different envelope", e)
	}
	owned.Target.Metadata = nr
	s, e = selectBatch(context.Background(), owned)
	if e != nil {
		t.Fatal(e)
	}
	s.Clear()
}
