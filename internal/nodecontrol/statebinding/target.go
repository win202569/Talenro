package statebinding

import (
	"context"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/statecomposition"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

func bindTransitions(ctx context.Context, in Input) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	previous, proposed := in.Main.Previous.Root.Mark, in.Main.Proposed.Root.Mark
	if len(in.Main.RootTransition) == 0 {
		if previous.Version != proposed.Version || previous.Sequence != proposed.Sequence || previous.Material.Payload != proposed.Material.Payload {
			return ErrBindingMismatch
		}
		return nil
	}
	oldRef, err := projectRef(previous.Material)
	if err != nil {
		return err
	}
	old, err := findExpected(ctx, in.Proposed.Materials, oldRef)
	if err != nil {
		return err
	}
	if old.Root.Version != previous.Version || old.Sequence != previous.Sequence {
		return ErrBindingMismatch
	}
	version, sequence, payload := previous.Version, previous.Sequence, previous.Material.Payload
	for _, r := range in.Main.RootTransition {
		ref, err := projectRef(r.Material)
		if err != nil {
			return err
		}
		e, err := findExpected(ctx, in.Proposed.Materials, ref)
		if err != nil {
			return err
		}
		rot := e.Rotation
		if ref.Kind != statematerials.RootRotation || e.Epoch != in.Proposed.Materials.Scope.Epoch || e.Sequence != r.Sequence || rot.PreviousVersion != r.PreviousVersion || rot.NewVersion != r.NewVersion || rot.OldRootPayload != r.OldRootPayload || rot.NewRootPayload != r.NewRootPayload || r.PreviousVersion != version || r.NewVersion != version+1 || r.OldRootPayload != payload || r.Sequence <= sequence {
			return ErrBindingMismatch
		}
		next, err := findExpected(ctx, in.Proposed.Materials, statematerials.MaterialRef{Kind: statematerials.RootSet, Payload: r.NewRootPayload})
		if err != nil {
			return err
		}
		if next.Epoch != in.Proposed.Materials.Scope.Epoch || next.Root.Version != r.NewVersion || next.Sequence != r.Sequence {
			return ErrBindingMismatch
		}
		version, sequence, payload = r.NewVersion, r.Sequence, r.NewRootPayload
	}
	if proposed.Version != version || proposed.Sequence != sequence || proposed.Material.Payload != payload {
		return ErrBindingMismatch
	}
	return ctx.Err()
}

func bindTarget(ctx context.Context, in Input, proposed statecomposition.Summary) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if in.Main.Proposed.Desired.Header.Coverage != statecontract.Present || in.Proposed.Target.Desired.Kind != statematerials.Desired || in.Proposed.Target.Metadata.Kind != statematerials.Metadata || in.Proposed.Target.AnchorRoot.Kind != statematerials.RootSet {
		return ErrInvalidInput
	}
	desiredRef, err := projectRef(in.Main.Proposed.Desired.Mark.Material)
	if err != nil {
		return err
	}
	target := in.Proposed.Target
	if target.Desired != desiredRef || target.Metadata.Payload != in.Main.Proposed.Desired.MetadataPayload || target.AnchorRoot.Payload != in.Proposed.Anchor.Payload || in.Proposed.Anchor.Scope != in.Proposed.Materials.Scope {
		return ErrBindingMismatch
	}
	if proposed.Status != statecomposition.ConditionalTargetVerified || proposed.Scope != in.Proposed.Materials.Scope || proposed.Subset.Status != statematerials.SubsetConsistent || proposed.Subset.Scope != in.Proposed.Materials.Scope || !reflect.DeepEqual(proposed.Target, target) || proposed.Anchor.PayloadDigest != in.Proposed.Anchor.Payload || proposed.Anchor.EnvelopeDigest != in.Proposed.Anchor.Envelope {
		return ErrBindingMismatch
	}
	des, err := findExpected(ctx, in.Proposed.Materials, target.Desired)
	if err != nil {
		return err
	}
	meta, err := findExpected(ctx, in.Proposed.Materials, target.Metadata)
	if err != nil {
		return err
	}
	rootRef := target.AnchorRoot
	if len(target.Links) > 0 {
		rootRef = target.Links[len(target.Links)-1].NewRoot
	}
	if rootRef.Kind != statematerials.RootSet || rootRef != proposed.FinalRoot || rootRef.Payload != des.Desired.RootPayload || meta.Metadata.RootPayload != rootRef.Payload || des.Desired.MetadataPayload != target.Metadata.Payload {
		return ErrBindingMismatch
	}
	root, err := findExpected(ctx, in.Proposed.Materials, rootRef)
	if err != nil {
		return err
	}
	c := proposed.Coordinates
	if des.Epoch != in.Proposed.Materials.Scope.Epoch || meta.Epoch != in.Proposed.Materials.Scope.Epoch || root.Epoch != in.Proposed.Materials.Scope.Epoch || c.ExpectedNodeID.String() != in.Proposed.Materials.Scope.NodeID || c.ExpectedEpoch != in.Proposed.Materials.Scope.Epoch || c.FinalRootVersion != root.Root.Version || c.MetadataVersion != meta.Metadata.Version || c.DesiredGeneration != des.Desired.Generation || c.DesiredSequence != des.Sequence || c.DesiredKeyID != des.Desired.SigningKeyID {
		return ErrBindingMismatch
	}
	return ctx.Err()
}
