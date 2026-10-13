package nodeissuerbinding

import "context"

// Evaluate establishes only conditional local references for this observation.
// It does not authenticate production mappings, history, or runtime authority.
func Evaluate(ctx context.Context, in Input) (*Model, error) {
	o, e := ownInput(ctx, in)
	if e != nil {
		return nil, e
	}
	transferred := false
	defer func() {
		if !transferred {
			o.clear()
		}
	}()
	if e = evaluateLower(ctx, o); e != nil {
		return nil, e
	}
	projection, e := bindProjection(ctx, o.input, o.summary.Evidence)
	if e != nil {
		return nil, e
	}
	o.summary.Proof = SelectedIssuerSerialReferencesConditionalTimeBound
	o.summary.Profile = selectedIssuerSerialProfile
	o.summary.Proposed = projection
	o.summary.ObservedAt = o.summary.Evidence.ObservedAt
	o.summary.ValidUntil = o.summary.Evidence.ValidUntil
	o.summary.Obligations = [5]Obligation{
		StoreIssuerMappingAndDeploymentProvenance,
		KeyHandleAndPrivateKeyPossession,
		IdentityEpochLineageHistoryAndRowStatus,
		MainStateAndObservationProvenance,
		FullIdentityCompositionCommitAndRuntimeAuthorization,
	}
	if e = stageError(ctx, nil); e != nil {
		return nil, e
	}
	o.valid = true
	transferred = true
	return &Model{owned: o}, nil
}
