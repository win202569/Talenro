package statepolicybinding

import (
	"context"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
)

func bindStateProjection(ctx context.Context, table statematerials.Input, main statecontract.MainStateContentModel, state *statepolicy.StateModel) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if state == nil || state.Metadata.Version != main.Metadata.Mark.Version || state.Metadata.Sequence != main.Metadata.Mark.Sequence || state.Metadata.Payload != main.Metadata.Mark.Material.Payload || state.Desired.Version != main.Desired.Mark.Version || state.Desired.Sequence != main.Desired.Mark.Sequence || state.Desired.Payload != main.Desired.Mark.Material.Payload {
		return ErrBindingMismatch
	}
	return bindKeyProjection(ctx, table, main.Metadata.Mark.Material, state)
}
func bindKeyProjection(ctx context.Context, table statematerials.Input, metadata statecontract.ArtifactRef, state *statepolicy.StateModel) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if state == nil {
		return ErrBindingMismatch
	}
	if metadata.Kind != statecontract.Metadata {
		return ErrInvalidMaterial
	}
	ref, e := projectRef(metadata)
	if e != nil {
		return e
	}
	raw, e := findMaterial(ctx, table, ref)
	if e != nil {
		return e
	}
	parsed, e := stateformat.ParseEnvelope(stateformat.Metadata, raw.Raw)
	defer parsed.Clear()
	if e != nil {
		return ErrInvalidMaterial
	}
	material, e := parsed.MetadataMaterial()
	defer material.Clear()
	if e != nil {
		return ErrInvalidMaterial
	}
	if parsed.PayloadDigest() != ref.Payload || parsed.EnvelopeDigest() != ref.Envelope {
		return ErrBindingMismatch
	}
	if !reflect.DeepEqual(material.RevokedIDs, state.RevokedIDs) {
		return ErrBindingMismatch
	}
	for _, current := range material.Keys {
		matches := 0
		for _, known := range state.Keys {
			if e := ctx.Err(); e != nil {
				return e
			}
			if known.Material.KeyID == current.KeyID {
				matches++
				if !reflect.DeepEqual(known.Material, current) {
					return ErrBindingMismatch
				}
			}
		}
		if matches != 1 {
			return ErrBindingMismatch
		}
	}
	return ctx.Err()
}
func bindStateIdentity(ctx context.Context, in Input, state *statepolicy.StateModel, proposed bool) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if state == nil || in.Policy.Old == nil {
		return ErrBindingMismatch
	}
	a, b := in.Binding.Main.Previous.Scope, in.Binding.Main.Proposed.Scope
	d := in.Policy.Deployment
	batch := in.Policy.Batch
	if state.NodeID.String() != a.NodeID || state.NodeID.String() != b.NodeID || state.NodeID.String() != batch.ExpectedNodeID || state.NodeID != d.NodeID || state.Epoch != a.AuthorityEpoch || state.Epoch != b.AuthorityEpoch || state.Epoch != batch.ExpectedEpoch || state.Epoch != d.Epoch || state.AnchorPayload != in.Binding.Proposed.Anchor.Payload || state.AnchorPayload != d.AnchorPayload || state.DeploymentSourceID != d.SourceID || state.TimeProviderID != in.Policy.Time.ProviderID {
		return ErrBindingMismatch
	}
	if proposed && (state.Status != statepolicy.Unknown || state.StateID != in.Policy.Old.StateID || state.CounterID != in.Policy.Old.CounterID || state.StateEpoch != in.Policy.Old.StateEpoch || !state.TimeFloor.Equal(in.Policy.Old.TimeFloor)) {
		return ErrBindingMismatch
	}
	return ctx.Err()
}
