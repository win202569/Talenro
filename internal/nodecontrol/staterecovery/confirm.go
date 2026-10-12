package staterecovery

import "context"

type ConfirmationModel struct {
	valid    bool
	snapshot RecoveryInput
	summary  ConfirmationSummary
}

func (m *ConfirmationModel) Summary() ConfirmationSummary {
	if m == nil || !m.valid {
		return ConfirmationSummary{}
	}
	return m.summary
}
func (m *ConfirmationModel) Clear() {
	if m != nil {
		*m = ConfirmationModel{}
	}
}

// ConfirmModel independently recovers a new scan. Neither CommittedModel nor
// NotCommitted is a provider receipt or a permission to retry an operation.
func ConfirmModel(ctx context.Context, m *CommitModel, in ConfirmationInput) (*ConfirmationModel, error) {
	if e := checkContext(ctx); e != nil {
		return nil, e
	}
	if m == nil || !m.valid {
		return nil, ErrStaleModel
	}
	if m.summary.Phase != NeedsRecovery {
		return nil, ErrInvalidTransition
	}
	if in.AttemptID != m.summary.AttemptID {
		return nil, ErrStaleModel
	}
	r, e := RecoverModel(ctx, in.Scan)
	if e != nil {
		return nil, e
	}
	defer r.Clear()
	scan := r.snapshot
	if !sameSources(scan, m.initial) || scan.Counter.ScanID == m.initial.Counter.ScanID || scan.Counter.ScanID == m.recheckScanID {
		return nil, ErrStaleModel
	}
	outcome := ConfirmationUnknown
	switch scan.Counter.Value {
	case m.initial.Counter.Value:
		if m.acknowledged {
			return nil, security(CounterMismatch)
		}
		if r.summary.SelectedSlot != m.summary.BaseSlot || !sameBlob(slotAt(scan, m.summary.BaseSlot), slotAt(m.initial, m.summary.BaseSlot)) {
			return nil, ErrStaleModel
		}
		outcome = NotCommitted
	case m.summary.TargetStateEpoch:
		if !nonzero(m.summary.CandidateBlobDigest) {
			return nil, ErrAssuranceMissing
		}
		if !m.acknowledged && m.interruptedPhase != ReadyToAdvance {
			return nil, ErrStaleModel
		}
		if r.summary.SelectedSlot != m.summary.TargetSlot || !candidateMatches(slotAt(scan, m.summary.TargetSlot), m) {
			return nil, ErrStaleModel
		}
		outcome = CommittedModel
	default:
		return nil, security(CounterMismatch)
	}
	if e := checkContext(ctx); e != nil {
		return nil, e
	}
	return &ConfirmationModel{valid: true, snapshot: scan, summary: ConfirmationSummary{Outcome: outcome, AttemptID: in.AttemptID, Recovery: r.summary}}, nil
}
