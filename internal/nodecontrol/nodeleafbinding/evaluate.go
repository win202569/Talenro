package nodeleafbinding

import "context"

// Evaluate binds the Proposed projection for this observation only. It does not
// authorize an identity commit, runtime action, or reuse at a later observation.
func Evaluate(ctx context.Context, in Input) (result *Model, err error) {
	o, err := ownInput(ctx, in)
	if err != nil {
		return nil, err
	}
	defer func() {
		if result == nil {
			o.clear()
		}
	}()
	if err = evaluateLower(ctx, o); err != nil {
		return nil, err
	}
	o.summary.Proposed, err = bindProjection(ctx, o.input, o.summary.Evidence)
	if err != nil {
		return nil, err
	}
	o.summary.Proof = SelectedLeafProjectionConditionalTimeBound
	o.summary.ObservedAt = o.summary.Evidence.ObservedAt
	o.summary.ValidUntil = o.summary.Evidence.ValidUntil
	o.summary.Obligations = [5]Obligation{
		IssuerAndSerialReferenceMapping,
		KeyHandleAndPrivateKeyPossession,
		IdentityEpochLineageHistoryAndRowStatus,
		MainStateAndObservationProvenance,
		FullIdentityCompositionCommitAndRuntimeAuthorization,
	}
	if err = stageError(ctx, nil); err != nil {
		return nil, err
	}
	o.valid = true
	return &Model{owned: o}, nil
}
