package statebinding

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/contracts"
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
		return statematerials.MaterialRef{}, ErrInvalidInput
	}
	if ref.Payload == (contracts.Digest{}) || ((kind == statematerials.RootSet) != (ref.Envelope == (contracts.Digest{}))) {
		return statematerials.MaterialRef{}, ErrInvalidInput
	}
	return statematerials.MaterialRef{Kind: kind, Payload: ref.Payload, Envelope: ref.Envelope}, nil
}
func findExpected(ctx context.Context, in statematerials.Input, ref statematerials.MaterialRef) (statematerials.Expectation, error) {
	if ctx == nil {
		return statematerials.Expectation{}, ErrInvalidInput
	}
	if err := ctx.Err(); err != nil {
		return statematerials.Expectation{}, err
	}
	foundMaterial, foundExpected := false, false
	var out statematerials.Expectation
	for _, m := range in.Materials {
		if err := ctx.Err(); err != nil {
			return statematerials.Expectation{}, err
		}
		if m.Ref.Kind == ref.Kind && m.Ref.Payload == ref.Payload {
			if m.Ref != ref {
				return statematerials.Expectation{}, ErrBindingMismatch
			}
			foundMaterial = true
		}
	}
	for _, e := range in.Expected {
		if err := ctx.Err(); err != nil {
			return statematerials.Expectation{}, err
		}
		if e.Ref.Kind == ref.Kind && e.Ref.Payload == ref.Payload {
			if e.Ref != ref {
				return statematerials.Expectation{}, ErrBindingMismatch
			}
			foundExpected = true
			out = e
		}
	}
	if !foundMaterial || !foundExpected {
		return statematerials.Expectation{}, ErrMissingMaterial
	}
	return out, nil
}
func equalSlice[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func bindProjection(ctx context.Context, in Input) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if in.Main.Proposed.Desired.Header.Coverage != statecontract.Present {
		return ErrInvalidInput
	}
	for i, x := range [2]statecontract.MainStateContentModel{in.Main.Previous, in.Main.Proposed} {
		table := in.Previous
		if i == 1 {
			table = in.Proposed.Materials
		}
		if x.Scope.NodeID != table.Scope.NodeID || x.Scope.AuthorityEpoch != table.Scope.Epoch {
			return ErrBindingMismatch
		}
		rootRef, err := projectRef(x.Root.Mark.Material)
		if err != nil {
			return err
		}
		root, err := findExpected(ctx, table, rootRef)
		if err != nil {
			return err
		}
		if rootRef.Kind != statematerials.RootSet || root.Epoch != x.Scope.AuthorityEpoch || root.Sequence != x.Root.Mark.Sequence || root.Root.Version != x.Root.Mark.Version || root.Root.Threshold != uint64(x.Root.Threshold) || !equalSlice(root.Root.KeyIDs, x.Root.KeyIDs) {
			return ErrBindingMismatch
		}
		metaRef, err := projectRef(x.Metadata.Mark.Material)
		if err != nil {
			return err
		}
		meta, err := findExpected(ctx, table, metaRef)
		if err != nil {
			return err
		}
		m := meta.Metadata
		if metaRef.Kind != statematerials.Metadata || meta.Epoch != x.Scope.AuthorityEpoch || meta.Sequence != x.Metadata.Mark.Sequence || m.Version != x.Metadata.Mark.Version || m.RootVersion != x.Metadata.RootVersion || m.RootPayload != x.Metadata.RootPayload || !equalSlice(m.KeyIDs, x.Metadata.KeyIDs) || !equalSlice(m.RevokedIDs, x.Metadata.RevokedIDs) || m.Window.Issued != x.Metadata.Window.Issued || m.Window.Until != x.Metadata.Window.Until {
			return ErrBindingMismatch
		}
		if x.Desired.Header.Coverage == statecontract.NeverEstablished {
			continue
		}
		if x.Desired.Header.Coverage != statecontract.Present {
			return ErrInvalidInput
		}
		desRef, err := projectRef(x.Desired.Mark.Material)
		if err != nil {
			return err
		}
		des, err := findExpected(ctx, table, desRef)
		if err != nil {
			return err
		}
		d := des.Desired
		if desRef.Kind != statematerials.Desired || des.Epoch != x.Scope.AuthorityEpoch || des.Sequence != x.Desired.Mark.Sequence || d.NodeID != x.Scope.NodeID || d.Generation != x.Desired.Mark.Version || d.InventoryVersion != x.Desired.InventoryVersion || d.ResourceVersion != x.Desired.ResourceVersion || d.ResourcePayload != x.Desired.ResourcePayload || d.RootPayload != x.Desired.RootPayload || d.MetadataPayload != x.Desired.MetadataPayload || d.SigningKeyID != x.Desired.SigningKeyID || d.Window.Issued != x.Desired.Window.Issued || d.Window.Until != x.Desired.Window.Until || !equalSlice(d.SlotIDs, x.Desired.SlotIDs) {
			return ErrBindingMismatch
		}
	}
	return ctx.Err()
}
