package statecomposition

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/stateverify"
)

// Evaluate never accepts cached lower-layer evidence or an external Batch.
func Evaluate(ctx context.Context, in Input) (result *CompositionModel, err error) {
	owned, e := ownInput(ctx, in)
	if e != nil {
		return nil, e
	}
	defer func() {
		if result == nil {
			clearInput(&owned)
		}
	}()
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	subset, e := statematerials.Evaluate(ctx, owned.Materials)
	defer subset.Clear()
	if e = mapError(ctx, e); e != nil {
		return nil, e
	}
	selected, e := selectBatch(ctx, owned)
	defer selected.Clear()
	if e = mapSelectionError(ctx, e); e != nil {
		return nil, e
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	proof, e := stateverify.Verify(ctx, selected.batch)
	defer proof.Clear()
	if e = mapError(ctx, e); e != nil {
		return nil, e
	}
	if e := checkEvidence(ctx, owned, selected, proof); e != nil {
		return nil, e
	}
	summary, coordinates, finalRoot := subset.Summary(), proof.Coordinates(), selected.finalRoot
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	return &CompositionModel{valid: true, input: owned, subset: summary, coordinates: coordinates, finalRoot: finalRoot}, nil
}

// Selection emits only our fixed sentinels or actual context errors.
func mapSelectionError(ctx context.Context, err error) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	return err
}
