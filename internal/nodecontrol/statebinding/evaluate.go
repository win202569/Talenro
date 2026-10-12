package statebinding

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/statecomposition"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

func Evaluate(ctx context.Context, in Input) (result *BindingModel, err error) {
	owned, e := ownInput(ctx, in)
	if e != nil {
		return nil, e
	}
	model := &BindingModel{input: owned}
	defer func() {
		if result == nil {
			model.Clear()
		}
	}()
	contract, e := statecontract.EvaluateModel(ctx, owned.Main)
	defer contract.Clear()
	if e = mapError(ctx, e); e != nil {
		return nil, e
	}
	previous, e := statematerials.Evaluate(ctx, owned.Previous)
	defer previous.Clear()
	if e = mapError(ctx, e); e != nil {
		return nil, e
	}
	proposed, e := statecomposition.Evaluate(ctx, owned.Proposed)
	defer proposed.Clear()
	if e = mapError(ctx, e); e != nil {
		return nil, e
	}
	if e := bindProjection(ctx, owned); e != nil {
		return nil, mapError(ctx, e)
	}
	if e := bindTransitions(ctx, owned); e != nil {
		return nil, mapError(ctx, e)
	}
	// Summary owns detached mutable fields. Store them directly under the
	// already registered model cleanup before any further context callback.
	model.contractSummary = contract.Summary()
	model.previousSummary = previous.Summary()
	model.proposedSummary = proposed.Summary()
	if model.contractSummary.Status != statecontract.ContractShapeValid || model.contractSummary.Scope != owned.Main.Proposed.Scope || model.previousSummary.Status != statematerials.SubsetConsistent || model.previousSummary.Scope != owned.Previous.Scope {
		return nil, ErrBindingMismatch
	}
	if e := bindTarget(ctx, owned, model.proposedSummary); e != nil {
		return nil, mapError(ctx, e)
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	model.valid = true
	return model, nil
}
