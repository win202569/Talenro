package statetimebinding

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/statepolicybinding"
	"talenro.local/platform/internal/nodecontrol/statetime"
)

func finishSummary(in Input, policy statepolicybinding.Summary, evidence statetime.EvidenceSummary) Summary {
	until := policy.ValidUntil
	if evidence.ValidUntil.Before(until) {
		until = evidence.ValidUntil
	}
	return Summary{Status: ConditionalTimeEvidenceBound, Policy: policy, Time: evidence,
		Observations: ObservationBinding{BaseStateID: in.Policy.Policy.Old.StateID, MetadataSourceID: in.Time.Metadata.SourceID, MetadataObservationID: in.Time.Metadata.ObservationID, RequestObservationID: in.Time.Request.ObservationID, CheckpointSourceID: in.Time.Checkpoint.SourceID, CheckpointObservationID: in.Time.Checkpoint.ObservationID},
		ValidUntil:   until, Obligations: [5]Obligation{ObservationAndBaseProvenance, TrustedTimeAndTLS, AtomicRequestConsumption, DurableCheckpointCommit, FullMainStateAndRuntimeAuthorization}}
}

// Evaluate returns only a conditional fixed-snapshot model, never commit authority.
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
	policy, e := statepolicybinding.Evaluate(ctx, model.input.Policy)
	defer policy.Clear()
	if e != nil {
		return nil, mapError(ctx, e)
	}
	model.summary.Policy = policy.Summary()
	if e = bindScope(ctx, model.input, model.summary.Policy); e != nil {
		return nil, mapError(ctx, e)
	}
	if e = bindMetadata(ctx, model.input, model.summary.Policy); e != nil {
		return nil, mapError(ctx, e)
	}
	if e = bindObservations(ctx, model.input, model.summary.Policy); e != nil {
		return nil, mapError(ctx, e)
	}
	if e = bindOldCheckpoint(ctx, model.input); e != nil {
		return nil, mapError(ctx, e)
	}
	evidence, e := statetime.EvaluateModel(ctx, model.input.Time)
	defer evidence.Clear()
	if e != nil {
		return nil, mapError(ctx, e)
	}
	model.summary.Time = evidence.Summary()
	if e = bindEvidence(ctx, model.input, model.summary.Time); e != nil {
		return nil, mapError(ctx, e)
	}
	if model.summary.Policy.ValidUntil.IsZero() || model.summary.Time.ValidUntil.IsZero() {
		return nil, ErrInvalidTime
	}
	model.summary = finishSummary(model.input, model.summary.Policy, model.summary.Time)
	if e = ctx.Err(); e != nil {
		return nil, e
	}
	model.valid = true
	return model, nil
}
