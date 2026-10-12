package statepolicybinding

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/statebinding"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
	"time"
)

func makeSummary(in Input, binding statebinding.Summary, prior, next RootProjection, nextState *statepolicy.StateModel, noChange bool, validUntil time.Time) Summary {
	old := in.Policy.Old
	return Summary{Status: ConditionalSelectedPolicyBound, Binding: binding,
		Base:         PolicyBase{StateID: old.StateID, CounterID: old.CounterID, StateEpoch: old.StateEpoch, DeploymentSourceID: in.Policy.Deployment.SourceID, DeploymentObservationID: in.Policy.Deployment.ObservationID},
		Previous:     PolicyProjection{Root: prior, Metadata: old.Metadata, Desired: old.Desired, KeyCount: len(old.Keys), RevokedCount: len(old.RevokedIDs)},
		Proposed:     PolicyProjection{Root: next, Metadata: nextState.Metadata, Desired: nextState.Desired, KeyCount: len(nextState.Keys), RevokedCount: len(nextState.RevokedIDs)},
		HistoryCount: len(in.Policy.History), Time: TimeSnapshot{ProviderID: in.Policy.Time.ProviderID, ObservationID: in.Policy.Time.ObservationID, Now: in.Policy.Time.Now, OldFloor: old.TimeFloor, ObservedFloor: in.Policy.Time.Floor, MainFloor: in.Binding.Main.Proposed.Scope.TimeFloor},
		ValidUntil: validUntil, SelectedPolicyNoChange: noChange, Obligations: [3]Obligation{ExplicitObservationsAndBaseProvenance, TimeAttestationAndFreshness, OtherDomainPolicyCoverage}}
}

// Evaluate binds one explicit snapshot. Success remains conditional, not authority.
func Evaluate(ctx context.Context, in Input) (result *Model, err error) {
	owned, e := ownInput(ctx, in)
	if e != nil {
		return nil, mapError(ctx, e)
	}
	model := &Model{input: owned}
	defer func() {
		if result == nil {
			model.Clear()
		}
	}()
	binding, e := statebinding.Evaluate(ctx, model.input.Binding)
	defer binding.Clear()
	if e != nil {
		return nil, mapError(ctx, e)
	}
	model.summary.Binding = binding.Summary()
	if e = bindSelectedScope(ctx, model.input, model.summary.Binding); e != nil {
		return nil, mapError(ctx, e)
	}
	if e = bindBatch(ctx, model.input); e != nil {
		return nil, mapError(ctx, e)
	}
	path, e := buildRootPath(ctx, model.input)
	defer path.Clear()
	if e != nil {
		return nil, mapError(ctx, e)
	}
	if e = bindStateIdentity(ctx, model.input, model.input.Policy.Old, false); e != nil {
		return nil, mapError(ctx, e)
	}
	prior, e := bindRoots(ctx, model.input, path, model.input.Policy.Old, false)
	if e != nil {
		return nil, mapError(ctx, e)
	}
	if e = bindRootTransition(ctx, model.input, path); e != nil {
		return nil, mapError(ctx, e)
	}
	if e = bindStateProjection(ctx, model.input.Binding.Previous, model.input.Binding.Main.Previous, model.input.Policy.Old); e != nil {
		return nil, mapError(ctx, e)
	}
	if e = bindTime(ctx, model.input); e != nil {
		return nil, mapError(ctx, e)
	}
	candidate, e := statepolicy.EvaluateModel(ctx, model.input.Policy)
	defer candidate.Clear()
	if e != nil {
		return nil, mapError(ctx, e)
	}
	state := candidate.State()
	defer state.Clear()
	if e = bindStateIdentity(ctx, model.input, state, true); e != nil {
		return nil, mapError(ctx, e)
	}
	next, e := bindRoots(ctx, model.input, path, state, true)
	if e != nil {
		return nil, mapError(ctx, e)
	}
	if e = bindStateProjection(ctx, model.input.Binding.Proposed.Materials, model.input.Binding.Main.Proposed, state); e != nil {
		return nil, mapError(ctx, e)
	}
	model.summary = makeSummary(model.input, model.summary.Binding, prior, next, state, candidate.NoChange(), candidate.ValidUntil())
	if e = ctx.Err(); e != nil {
		return nil, e
	}
	model.valid = true
	return model, nil
}
