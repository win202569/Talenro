package statepolicybinding

import (
	"bytes"
	"context"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/statebinding"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

func projectRef(ref statecontract.ArtifactRef) (statematerials.MaterialRef, error) {
	var kind statematerials.MaterialKind
	switch ref.Kind {
	case statecontract.RootPayload:
		kind = statematerials.RootSet
	case statecontract.RootRotation:
		kind = statematerials.RootRotation
	case statecontract.Metadata:
		kind = statematerials.Metadata
	case statecontract.DesiredSnapshot:
		kind = statematerials.Desired
	default:
		return statematerials.MaterialRef{}, ErrInvalidMaterial
	}
	return statematerials.MaterialRef{Kind: kind, Payload: ref.Payload, Envelope: ref.Envelope}, nil
}
func findMaterial(ctx context.Context, table statematerials.Input, ref statematerials.MaterialRef) (statematerials.Material, error) {
	if ctx == nil {
		return statematerials.Material{}, ErrInvalidInput
	}
	var out statematerials.Material
	found := false
	for _, m := range table.Materials {
		if e := ctx.Err(); e != nil {
			return statematerials.Material{}, e
		}
		if m.Ref.Kind == ref.Kind && m.Ref.Payload == ref.Payload {
			if found {
				return statematerials.Material{}, ErrDuplicateMaterial
			}
			out = m
			found = true
		}
	}
	if !found {
		return statematerials.Material{}, ErrMissingMaterial
	}
	if out.Ref != ref {
		return statematerials.Material{}, ErrBindingMismatch
	}
	return out, nil
}
func findExpected(ctx context.Context, table statematerials.Input, ref statematerials.MaterialRef) (statematerials.Expectation, error) {
	if ctx == nil {
		return statematerials.Expectation{}, ErrInvalidInput
	}
	var out statematerials.Expectation
	found := false
	for _, m := range table.Expected {
		if e := ctx.Err(); e != nil {
			return statematerials.Expectation{}, e
		}
		if m.Ref.Kind == ref.Kind && m.Ref.Payload == ref.Payload {
			if found {
				return statematerials.Expectation{}, ErrDuplicateMaterial
			}
			out = m
			found = true
		}
	}
	if !found {
		return statematerials.Expectation{}, ErrMissingMaterial
	}
	if out.Ref != ref {
		return statematerials.Expectation{}, ErrBindingMismatch
	}
	return out, nil
}
func selectedTail(in Input) statematerials.MaterialRef {
	t := in.Binding.Proposed.Target
	if len(t.Links) == 0 {
		return t.AnchorRoot
	}
	return t.Links[len(t.Links)-1].NewRoot
}
func bindSelectedScope(ctx context.Context, in Input, proof statebinding.Summary) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	comp := in.Binding.Proposed
	main := in.Binding.Main
	if proof.Status != statebinding.ConditionalProjectionBound || proof.Contract.Scope != main.Proposed.Scope || proof.Proposed.Scope != comp.Materials.Scope || !reflect.DeepEqual(proof.Proposed.Target, comp.Target) || proof.Proposed.FinalRoot != selectedTail(in) || proof.Proposed.Anchor.PayloadDigest != comp.Anchor.Payload || proof.Proposed.Anchor.EnvelopeDigest != comp.Anchor.Envelope {
		return ErrBindingMismatch
	}
	if in.Policy.Old == nil || in.Policy.Deployment.FreshInstall || main.Previous.Desired.Header.Coverage != statecontract.Present || main.Proposed.Desired.Header.Coverage != statecontract.Present || main.Previous.Scope.TimeFloor != main.Proposed.Scope.TimeFloor {
		return ErrUnsupportedPolicyShape
	}
	metadata, e := projectRef(main.Proposed.Metadata.Mark.Material)
	if e != nil {
		return e
	}
	currentRoot, e := projectRef(main.Proposed.Root.Mark.Material)
	if e != nil {
		return e
	}
	if metadata != comp.Target.Metadata || currentRoot != selectedTail(in) {
		return ErrUnsupportedPolicyShape
	}
	desired, e := projectRef(main.Proposed.Desired.Mark.Material)
	if e != nil {
		return e
	}
	if desired != comp.Target.Desired {
		return ErrBindingMismatch
	}
	return ctx.Err()
}
func bindBatch(ctx context.Context, in Input) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	b := in.Policy.Batch
	c := in.Binding.Proposed
	if b.ExpectedNodeID != c.Materials.Scope.NodeID || b.ExpectedEpoch != c.Materials.Scope.Epoch || b.ExpectedNodeID != in.Binding.Main.Proposed.Scope.NodeID || b.ExpectedEpoch != in.Binding.Main.Proposed.Scope.AuthorityEpoch || !bytes.Equal(b.AnchorEnvelope, c.Anchor.Raw) || len(b.Rotations) != len(c.Target.Links) {
		return ErrBindingMismatch
	}
	active := c.Anchor.Raw
	for i, l := range c.Target.Links {
		m, e := findMaterial(ctx, c.Materials, l.Rotation)
		if e != nil {
			return e
		}
		if !bytes.Equal(b.Rotations[i], m.Raw) {
			return ErrBindingMismatch
		}
		active = m.Raw
	}
	if !bytes.Equal(b.ActiveRootEnvelope, active) {
		return ErrBindingMismatch
	}
	meta, e := findMaterial(ctx, c.Materials, c.Target.Metadata)
	if e != nil {
		return e
	}
	if !bytes.Equal(b.MetadataEnvelope, meta.Raw) {
		return ErrBindingMismatch
	}
	desired, e := findMaterial(ctx, c.Materials, c.Target.Desired)
	if e != nil {
		return e
	}
	if !bytes.Equal(b.DesiredEnvelope, desired.Raw) {
		return ErrBindingMismatch
	}
	return ctx.Err()
}
