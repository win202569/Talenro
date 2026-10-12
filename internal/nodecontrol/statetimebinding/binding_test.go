package statetimebinding

import (
	"context"
	"crypto/sha256"
	"errors"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/statepolicybinding"
	"testing"
)

func TestTimeBindingScope(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		in := timeBindingFixture(t, n, 100)
		s := checkLowerModels(t, in)
		if e := bindScope(context.Background(), in, s); e != nil {
			t.Fatal(n, e)
		}
	}
	in := attachTime(t, policyBindingFixture(t, 1, 0), 100)
	lower, e := statepolicybinding.Evaluate(context.Background(), in.Policy)
	defer lower.Clear()
	if e != nil || lower == nil {
		t.Fatal("legal candidate control", e)
	}
	if !errors.Is(bindScope(context.Background(), in, lower.Summary()), ErrUnsupportedBindingShape) {
		t.Fatal("candidate admitted")
	}
}
func TestTimeBindingUnsupportedSections(t *testing.T) {
	in := timeBindingFixture(t, 1, 100)
	s := checkLowerModels(t, in)
	for _, change := range []func(*Input){
		func(x *Input) { x.Policy.Binding.Main.Proposed.Root.Threshold++ },
		func(x *Input) { x.Policy.Binding.Main.Proposed.Root.KeyIDs[0][0] ^= 1 },
		func(x *Input) { x.Policy.Binding.Main.Proposed.Metadata.KeyIDs[0][0] ^= 1 },
		func(x *Input) { x.Policy.Binding.Main.Proposed.Metadata.Window.Until++ },
		func(x *Input) { x.Policy.Binding.Main.Proposed.Metadata.RootPayload[0] ^= 1 },
		func(x *Input) { x.Policy.Binding.Main.Proposed.Metadata.Mark.Material.Envelope[0] ^= 1 },
	} {
		bad := cloneFixture(t, in)
		change(&bad)
		if !errors.Is(bindScope(context.Background(), bad, s), ErrUnsupportedBindingShape) {
			t.Fatal("partial comparison")
		}
	}
	for _, change := range []func(*Input){
		func(x *Input) { x.Policy.Binding.Main.Proposed.Metadata.History.Payload[0] ^= 1 },
		func(x *Input) { x.Policy.Binding.Main.Proposed.Metadata.RevokedSetDigest[0] ^= 1 },
	} {
		bad := cloneFixture(t, in)
		change(&bad)
		control := checkLowerModels(t, bad)
		if !errors.Is(bindScope(context.Background(), bad, control), ErrUnsupportedBindingShape) {
			t.Fatal("lower-valid section drift")
		}
	}
}
func TestTimeBindingMetadata(t *testing.T) {
	in := timeBindingFixture(t, 1, 100)
	s := checkLowerModels(t, in)
	if e := bindMetadata(context.Background(), in, s); e != nil {
		t.Fatal(e)
	}
	for _, change := range []func(*Input){
		func(x *Input) { x.Time.MetadataEnvelope[0] ^= 1 },
		func(x *Input) { x.Time.Metadata.Payload[0] ^= 1 },
		func(x *Input) { x.Time.Metadata.Envelope[0] ^= 1 },
		func(x *Input) { x.Time.Metadata.Version++ },
	} {
		bad := cloneFixture(t, in)
		change(&bad)
		if !errors.Is(bindMetadata(context.Background(), bad, s), ErrBindingMismatch) {
			t.Fatal("metadata mismatch")
		}
	}
	ref := in.Policy.Binding.Proposed.Target.Metadata
	table := in.Policy.Binding.Previous
	m, e := findMaterial(context.Background(), table, ref)
	if e != nil || m.Ref != ref {
		t.Fatal(e)
	}
	table.Materials = append(table.Materials, m)
	if _, e = findMaterial(context.Background(), table, ref); e != ErrDuplicateMaterial {
		t.Fatal("duplicate", e)
	}
	if _, e = findMaterial(context.Background(), statematerials.Input{}, ref); e != ErrMissingMaterial {
		t.Fatal("missing", e)
	}
	different := ref
	different.Envelope[0] ^= 1
	if _, e = findMaterial(context.Background(), in.Policy.Binding.Previous, different); e != ErrBindingMismatch {
		t.Fatal("E identity", e)
	}
	for _, side := range []bool{false, true} {
		bad := cloneFixture(t, in)
		tbl := &bad.Policy.Binding.Previous
		if side {
			tbl = &bad.Policy.Binding.Proposed.Materials
		}
		for i := range tbl.Materials {
			if tbl.Materials[i].Ref == ref {
				tbl.Materials[i].Raw[0] ^= 1
			}
		}
		if !errors.Is(bindMetadata(context.Background(), bad, s), ErrBindingMismatch) {
			t.Fatal("table raw")
		}
	}
}
func TestTimeBindingAlternateMetadataEnvelope(t *testing.T) {
	in := attachTime(t, multiRootPolicyBindingFixture(t), 100)
	checkLowerModels(t, in)
	replacement := cloneFixture(t, in)
	replacement.Time.MetadataEnvelope = mint(t, "metadata", object(t, payload(t, in.Time.MetadataEnvelope)), private(3))
	replacement.Time.Metadata.Envelope = sha256.Sum256(replacement.Time.MetadataEnvelope)
	s := checkLowerModels(t, replacement)
	if replacement.Time.Metadata.Payload != in.Time.Metadata.Payload || replacement.Time.Metadata.Envelope == in.Time.Metadata.Envelope {
		t.Fatal("control")
	}
	if !errors.Is(bindMetadata(context.Background(), replacement, s), ErrBindingMismatch) {
		t.Fatal("alternate valid E accepted")
	}
}
