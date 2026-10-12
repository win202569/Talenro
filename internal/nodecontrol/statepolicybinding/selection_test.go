package statepolicybinding

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"testing"
)

func TestPolicyBindingSelectedScope(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		for _, old := range []int{0, n} {
			in := policyBindingFixture(t, n, old)
			proof := checkBindingInput(t, in)
			checkPolicyInput(t, in)
			if e := bindSelectedScope(context.Background(), in, proof); e != nil {
				t.Fatal(n, old, e)
			}
		}
	}
	in := historicalPolicyBindingFixture(t)
	proof := checkBindingInput(t, in)
	if e := bindSelectedScope(context.Background(), in, proof); !errors.Is(e, ErrUnsupportedPolicyShape) {
		t.Fatal("historical target", e)
	}
	for name, change := range map[string]func(*Input){
		"uninitialized": func(in *Input) { in.Policy.Old = nil },
		"fresh_install": func(in *Input) { in.Policy.Deployment.FreshInstall = true },
		"new_main_floor": func(in *Input) {
			in.Binding.Main.Proposed.Scope.TimeFloor = statecontract.OptionalTime{Present: true, UnixSeconds: 1577750400}
		},
		"advanced_main_floor": func(in *Input) {
			in.Binding.Main.Previous.Scope.TimeFloor = statecontract.OptionalTime{Present: true, UnixSeconds: 1577750400}
			in.Binding.Main.Proposed.Scope.TimeFloor = statecontract.OptionalTime{Present: true, UnixSeconds: 1577750401}
		},
		"no_old_desired": func(in *Input) {
			x := &in.Binding.Main.Previous
			x.Desired = statecontract.DesiredSection{Header: statecontract.SectionHeader{Domain: statecontract.DomainDesired, Coverage: statecontract.NeverEstablished}}
			x.Supervisor.Verified = statecontract.AppliedRecord{Coverage: statecontract.NeverEstablished}
			x.Supervisor.Applied = statecontract.AppliedRecord{Coverage: statecontract.NeverEstablished}
			x.LKG = statecontract.LKGSection{Header: statecontract.SectionHeader{Domain: statecontract.DomainLKG, Coverage: statecontract.NeverEstablished}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			in := policyBindingFixture(t, 1, 1)
			change(&in)
			proof := checkBindingInput(t, in)
			if e := bindSelectedScope(context.Background(), in, proof); !errors.Is(e, ErrUnsupportedPolicyShape) {
				t.Fatal(e)
			}
		})
	}
}
func TestPolicyBindingBatch(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		in := policyBindingFixture(t, n, n)
		checkBindingInput(t, in)
		if e := bindBatch(context.Background(), in); e != nil {
			t.Fatal(n, e)
		}
	}
	for name, change := range map[string]func(*Input){
		"node":     func(in *Input) { in.Policy.Batch.ExpectedNodeID = "22345678-1234-4234-8234-123456789abc" },
		"epoch":    func(in *Input) { in.Policy.Batch.ExpectedEpoch++ },
		"anchor":   func(in *Input) { in.Policy.Batch.AnchorEnvelope[0] ^= 1 },
		"rotation": func(in *Input) { in.Policy.Batch.Rotations[0][0] ^= 1 },
		"count":    func(in *Input) { in.Policy.Batch.Rotations = in.Policy.Batch.Rotations[:1] },
		"order": func(in *Input) {
			in.Policy.Batch.Rotations[0], in.Policy.Batch.Rotations[1] = in.Policy.Batch.Rotations[1], in.Policy.Batch.Rotations[0]
		},
		"active_root": func(in *Input) { in.Policy.Batch.ActiveRootEnvelope = in.Policy.Batch.AnchorEnvelope },
		"metadata":    func(in *Input) { in.Policy.Batch.MetadataEnvelope[0] ^= 1 },
		"desired":     func(in *Input) { in.Policy.Batch.DesiredEnvelope[0] ^= 1 },
	} {
		t.Run(name, func(t *testing.T) {
			in := policyBindingFixture(t, 2, 2)
			change(&in)
			if e := bindBatch(context.Background(), in); !errors.Is(e, ErrBindingMismatch) {
				t.Fatal(e)
			}
		})
	}
	in := policyBindingFixture(t, 1, 1)
	replacement := policyBindingFixture(t, 0, 0)
	in.Policy = replacement.Policy
	checkBindingInput(t, in)
	checkPolicyInput(t, in)
	if e := bindBatch(context.Background(), in); !errors.Is(e, ErrBindingMismatch) {
		t.Fatal("independently valid other batch", e)
	}
	in = multiRootPolicyBindingFixture(t)
	checkBindingInput(t, in)
	checkPolicyInput(t, in)
	originalPayload := in.Binding.Proposed.Target.Metadata.Payload
	in.Policy.Batch.MetadataEnvelope = mint(t, "metadata", object(t, payload(t, in.Policy.Batch.MetadataEnvelope)), private(3))
	checkPolicyInput(t, in)
	if sha256.Sum256(payload(t, in.Policy.Batch.MetadataEnvelope)) != originalPayload {
		t.Fatal("replacement payload changed")
	}
	if bytes.Equal(in.Policy.Batch.MetadataEnvelope, material(&in.Binding.Proposed, in.Binding.Proposed.Target.Metadata).Raw) {
		t.Fatal("fixture not changed")
	}
	if e := bindBatch(context.Background(), in); !errors.Is(e, ErrBindingMismatch) {
		t.Fatal("same payload other envelope", e)
	}
}
