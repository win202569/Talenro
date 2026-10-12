package statepolicybinding

import (
	"context"
	"errors"
	"talenro.local/platform/internal/nodecontrol/statecomposition"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
	"testing"
)

func TestPolicyBindingRootPath(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		in := policyBindingFixture(t, n, n)
		checkBindingInput(t, in)
		path, e := buildRootPath(context.Background(), in)
		if e != nil {
			t.Fatal(e)
		}
		if len(path.items) != n+1 || path.items[0].projection.Source.Kind != InitialAnchor || path.items[0].projection.Source.Payload != in.Binding.Proposed.Anchor.Payload || path.items[0].projection.Source.Envelope != in.Binding.Proposed.Anchor.Envelope {
			t.Fatal("anchor identity")
		}
		for i := 1; i <= n; i++ {
			p := path.items[i].projection
			l := in.Binding.Proposed.Target.Links[i-1]
			if p.Source.Kind != RootRotationProof || p.Source.Payload != l.Rotation.Payload || p.Source.Envelope != l.Rotation.Envelope || p.RootSet != l.NewRoot || p.Source.Payload == p.RootSet.Payload || in.Policy.Old.Roots[i].Mark.Payload != p.Source.Payload {
				t.Fatal("rotation/root-set identity conflation")
			}
		}
		keys := path.items[0].material.Keys
		path.Clear()
		path.Clear()
		if len(path.items) != 0 || keys[0].KeyID != ([32]byte{}) || keys[0].PublicKey != ([32]byte{}) {
			t.Fatal("path retained material")
		}
	}
}
func TestPolicyBindingRoots(t *testing.T) {
	for _, tc := range []struct{ n, old int }{{0, 0}, {1, 0}, {1, 1}, {32, 0}, {32, 17}, {32, 32}} {
		in := policyBindingFixture(t, tc.n, tc.old)
		checkBindingInput(t, in)
		path, e := buildRootPath(context.Background(), in)
		if e != nil {
			t.Fatal(e)
		}
		prior, e := bindRoots(context.Background(), in, path, in.Policy.Old, false)
		if e != nil || prior.RootSet.Payload != in.Binding.Main.Previous.Root.Mark.Material.Payload {
			t.Fatal("old prefix", e)
		}
		// Independently derive complete roots from raw test data, not a Candidate.
		full := policyBindingFixture(t, tc.n, tc.n).Policy.Old
		next, e := bindRoots(context.Background(), in, path, full, true)
		if e != nil || next.RootSet.Payload != in.Binding.Main.Proposed.Root.Mark.Material.Payload {
			t.Fatal("full chain", e)
		}
		path.Clear()
	}
	for name, change := range map[string]func(*Input){
		"mark_proof": func(in *Input) {
			in.Policy.Old.Roots[1].Mark.Payload = in.Binding.Proposed.Target.Links[0].NewRoot.Payload
		},
		"sequence":   func(in *Input) { in.Policy.Old.Roots[0].Mark.Sequence++ },
		"version":    func(in *Input) { in.Policy.Old.Roots[0].Root.Version++ },
		"threshold":  func(in *Input) { in.Policy.Old.Roots[0].Root.Threshold++ },
		"key_id":     func(in *Input) { in.Policy.Old.Roots[0].Root.Keys[0].KeyID[0] ^= 1 },
		"public_key": func(in *Input) { in.Policy.Old.Roots[0].Root.Keys[0].PublicKey[0] ^= 1 },
		"position": func(in *Input) {
			in.Policy.Old.Roots[0], in.Policy.Old.Roots[1] = in.Policy.Old.Roots[1], in.Policy.Old.Roots[0]
		},
		"empty":    func(in *Input) { in.Policy.Old.Roots = nil },
		"excess":   func(in *Input) { in.Policy.Old.Roots = append(in.Policy.Old.Roots, statepolicy.RootRecord{}) },
		"main_ref": func(in *Input) { in.Binding.Main.Previous.Root.Mark.Material.Payload[0] ^= 1 },
	} {
		t.Run(name, func(t *testing.T) {
			in := policyBindingFixture(t, 2, 2)
			path, e := buildRootPath(context.Background(), in)
			if e != nil {
				t.Fatal(e)
			}
			defer path.Clear()
			change(&in)
			if _, e = bindRoots(context.Background(), in, path, in.Policy.Old, false); !errors.Is(e, ErrBindingMismatch) {
				t.Fatal(e)
			}
		})
	}
}
func TestPolicyBindingRootSuffix(t *testing.T) {
	for _, tc := range []struct{ n, old int }{{0, 0}, {1, 0}, {1, 1}, {32, 17}, {32, 32}} {
		in := policyBindingFixture(t, tc.n, tc.old)
		checkBindingInput(t, in)
		path, e := buildRootPath(context.Background(), in)
		if e != nil {
			t.Fatal(e)
		}
		if e = bindRootTransition(context.Background(), in, path); e != nil {
			t.Fatal(tc, e)
		}
		path.Clear()
	}
	for name, change := range map[string]func(*Input){
		"length":           func(in *Input) { in.Binding.Main.RootTransition = in.Binding.Main.RootTransition[:1] },
		"source_envelope":  func(in *Input) { in.Binding.Main.RootTransition[0].Material.Envelope[0] ^= 1 },
		"old_payload":      func(in *Input) { in.Binding.Main.RootTransition[0].OldRootPayload[0] ^= 1 },
		"new_payload":      func(in *Input) { in.Binding.Main.RootTransition[0].NewRootPayload[0] ^= 1 },
		"sequence":         func(in *Input) { in.Binding.Main.RootTransition[0].Sequence++ },
		"previous_version": func(in *Input) { in.Binding.Main.RootTransition[0].PreviousVersion++ },
		"new_version":      func(in *Input) { in.Binding.Main.RootTransition[0].NewVersion++ },
		"order": func(in *Input) {
			in.Binding.Main.RootTransition[0], in.Binding.Main.RootTransition[1] = in.Binding.Main.RootTransition[1], in.Binding.Main.RootTransition[0]
		},
	} {
		t.Run(name, func(t *testing.T) {
			in := policyBindingFixture(t, 3, 1)
			path, e := buildRootPath(context.Background(), in)
			if e != nil {
				t.Fatal(e)
			}
			defer path.Clear()
			change(&in)
			if e = bindRootTransition(context.Background(), in, path); !errors.Is(e, ErrBindingMismatch) {
				t.Fatal(e)
			}
		})
	}
	in := policyBindingFixture(t, 1, 1)
	path, e := buildRootPath(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer path.Clear()
	in.Binding.Proposed.Target.Links = append(in.Binding.Proposed.Target.Links, statecomposition.RotationLink{})
	if _, e := buildRootPath(context.Background(), in); !errors.Is(e, ErrMissingMaterial) {
		t.Fatal("unknown root chosen", e)
	}
}
