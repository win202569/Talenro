package statetimebinding

import (
	"bytes"
	"context"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/statepolicybinding"
	"talenro.local/platform/internal/nodecontrol/statetime"
)

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
func currentMetadataRef(in Input) statematerials.MaterialRef {
	mark := in.Policy.Binding.Main.Previous.Metadata.Mark
	return statematerials.MaterialRef{Kind: statematerials.Metadata, Payload: mark.Material.Payload, Envelope: mark.Material.Envelope}
}
func bindScope(ctx context.Context, in Input, policy statepolicybinding.Summary) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	main := in.Policy.Binding.Main
	old := in.Policy.Policy.Old
	target := in.Policy.Binding.Proposed.Target
	if policy.Status != statepolicybinding.ConditionalSelectedPolicyBound {
		return ErrBindingMismatch
	}
	if old == nil || in.Policy.Policy.Deployment.FreshInstall || main.Previous.Desired.Header.Coverage != statecontract.Present || main.Proposed.Desired.Header.Coverage != statecontract.Present || main.Previous.Scope.TimeFloor != main.Proposed.Scope.TimeFloor {
		return ErrUnsupportedPolicyShape
	}
	if !reflect.DeepEqual(main.Previous.Root, main.Proposed.Root) || !reflect.DeepEqual(main.Previous.Metadata, main.Proposed.Metadata) || len(main.RootTransition) != 0 {
		return ErrUnsupportedBindingShape
	}
	if len(old.Roots) != len(target.Links)+1 {
		return ErrUnsupportedBindingShape
	}
	tail := target.AnchorRoot
	if len(target.Links) > 0 {
		tail = target.Links[len(target.Links)-1].NewRoot
	}
	mark := main.Previous.Root.Mark
	if mark.Material.Kind != statecontract.RootPayload || tail != (statematerials.MaterialRef{Kind: statematerials.RootSet, Payload: mark.Material.Payload, Envelope: mark.Material.Envelope}) || policy.Previous.Root.RootSet != tail || policy.Proposed.Root.RootSet != tail || policy.Previous.Root.Version != mark.Version || policy.Proposed.Root.Version != mark.Version || policy.Previous.Root.Sequence != mark.Sequence || policy.Proposed.Root.Sequence != mark.Sequence || target.Metadata != currentMetadataRef(in) {
		return ErrBindingMismatch
	}
	if policy.Base.StateID != old.StateID || policy.Base.CounterID != old.CounterID || policy.Base.StateEpoch != old.StateEpoch || policy.Base.DeploymentSourceID != in.Policy.Policy.Deployment.SourceID || policy.Base.DeploymentObservationID != in.Policy.Policy.Deployment.ObservationID || policy.HistoryCount != len(in.Policy.Policy.History) {
		return ErrBindingMismatch
	}
	return ctx.Err()
}
func bindMetadata(ctx context.Context, in Input, policy statepolicybinding.Summary) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	// Missing observations precede interpretation of their metadata coordinates.
	// Full selected-policy evaluation and supported-scope checks still run first.
	switch in.Time.Metadata.Status {
	case statetime.Available:
	default:
		return ErrAssuranceMissing
	}
	switch in.Time.Time.Status {
	case statetime.Available:
	default:
		return ErrAssuranceMissing
	}
	switch in.Time.Request.Status {
	case statetime.Available:
	default:
		return ErrAssuranceMissing
	}
	switch in.Time.Checkpoint.Status {
	case statetime.Available:
	default:
		return ErrAssuranceMissing
	}
	prev := in.Policy.Binding.Main.Previous.Metadata.Mark
	next := in.Policy.Binding.Main.Proposed.Metadata.Mark
	ref := currentMetadataRef(in)
	obs := in.Time.Metadata
	if prev.Material.Kind != statecontract.Metadata || next.Material.Kind != statecontract.Metadata || prev != next || obs.Version != prev.Version || obs.Payload != ref.Payload || obs.Envelope != ref.Envelope || policy.Previous.Metadata.Version != prev.Version || policy.Previous.Metadata.Sequence != prev.Sequence || policy.Previous.Metadata.Payload != ref.Payload || policy.Proposed.Metadata.Version != next.Version || policy.Proposed.Metadata.Sequence != next.Sequence || policy.Proposed.Metadata.Payload != ref.Payload || !bytes.Equal(in.Time.MetadataEnvelope, in.Policy.Policy.Batch.MetadataEnvelope) {
		return ErrBindingMismatch
	}
	for _, table := range [2]statematerials.Input{in.Policy.Binding.Previous, in.Policy.Binding.Proposed.Materials} {
		m, e := findMaterial(ctx, table, ref)
		if e != nil {
			return e
		}
		if !bytes.Equal(m.Raw, in.Time.MetadataEnvelope) {
			return ErrBindingMismatch
		}
	}
	return ctx.Err()
}
