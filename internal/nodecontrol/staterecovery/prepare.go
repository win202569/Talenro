package staterecovery

import "context"

// PrepareModel assumes external business validation; a content digest is not
// proof of authorization, full-state coverage, or semantic monotonicity.
func PrepareModel(ctx context.Context, in RecoveryInput, p ProposalObservation) (*CommitModel, error) {
	r, e := RecoverModel(ctx, in)
	if e != nil {
		return nil, e
	}
	defer r.Clear()
	if !nonzero(p.ProposalID) || !nonzero(p.AttemptID) || !nonzero(p.BaseBlobDigest) || !nonzero(p.ProposedContentDigest) {
		return nil, ErrInvalidInput
	}
	if p.ExpectedScope != r.summary.Scope || p.BaseBlobDigest != r.summary.BlobDigest || p.BaseStateEpoch != r.summary.StateEpoch {
		return nil, ErrStaleModel
	}
	s := CommitSummary{Phase: NoChange, Scope: r.summary.Scope, ProposalID: p.ProposalID, AttemptID: p.AttemptID, BaseSlot: r.summary.SelectedSlot, BaseStateEpoch: r.summary.StateEpoch, BaseBlobDigest: p.BaseBlobDigest, ProposedContentDigest: p.ProposedContentDigest}
	if p.ProposedContentDigest != r.summary.ContentDigest {
		if in.Counter.Value == in.Counter.Maximum {
			return nil, ErrCounterExhausted
		}
		s.Phase = Prepared
		s.TargetSlot = otherSlot(s.BaseSlot)
		s.TargetStateEpoch = in.Counter.Value + 1
	}
	if e := checkContext(ctx); e != nil {
		return nil, e
	}
	return &CommitModel{valid: true, initial: r.snapshot, base: r.summary, summary: s}, nil
}
