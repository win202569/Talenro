package staterecovery

import "context"

func sameSources(a, b RecoveryInput) bool {
	return a.ExpectedScope == b.ExpectedScope && a.Counter.SourceID == b.Counter.SourceID && a.A.SourceID == b.A.SourceID && a.B.SourceID == b.B.SourceID && a.Counter.Maximum == b.Counter.Maximum
}
func sameBlob(a, b SlotObservation) bool {
	return a.Status == b.Status && a.Scope == b.Scope && a.StateEpoch == b.StateEpoch && a.ContentDigest == b.ContentDigest && a.BlobDigest == b.BlobDigest && a.ModelVersion == b.ModelVersion
}
func candidateMatches(s SlotObservation, m *CommitModel) bool {
	c := m.summary
	return s.Status == SlotIntact && s.ModelVersion == 1 && s.Scope == c.Scope && s.StateEpoch == c.TargetStateEpoch && s.ContentDigest == c.ProposedContentDigest && s.BlobDigest == c.CandidateBlobDigest
}
func recheckBase(ctx context.Context, m *CommitModel, in RecoveryInput) error {
	r, e := RecoverModel(ctx, in)
	if e != nil {
		return e
	}
	defer r.Clear()
	if !sameSources(in, m.initial) || in.Counter.ScanID == m.initial.Counter.ScanID || in.Counter.Value != m.initial.Counter.Value || r.summary.SelectedSlot != m.base.SelectedSlot || !sameBlob(slotAt(in, m.summary.BaseSlot), slotAt(m.initial, m.summary.BaseSlot)) || !candidateMatches(slotAt(in, m.summary.TargetSlot), m) {
		return ErrStaleModel
	}
	return nil
}

// AdvanceModel returns a value-owned successor, not a performed operation.
// A lost response must enter NeedsRecovery; callers must never blindly retry
// a hardware increment based on a model phase.
func AdvanceModel(ctx context.Context, m *CommitModel, e Event) (*CommitModel, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	if m == nil || !m.valid {
		return nil, ErrStaleModel
	}
	p := m.summary.Phase
	if p < Prepared || p > ReadyToAdvance || e.Phase != p {
		return nil, ErrInvalidTransition
	}
	expected := map[CommitPhase]EventKind{Prepared: CandidateWritten, Written: FileSynced, FileDurable: DirectorySynced, CandidateDurable: BaseRechecked, ReadyToAdvance: CounterAdvanced}[p]
	if e.Kind != expected && e.Kind != Failed && e.Kind != OutcomeUnknown {
		return nil, ErrInvalidTransition
	}
	if !nonzero(e.EventID) || !nonzero(e.SourceID) {
		return nil, ErrInvalidInput
	}
	if m.eventCount >= uint8(len(m.eventIDs)) {
		return nil, ErrInvalidTransition
	}
	for _, id := range m.eventIDs[:m.eventCount] {
		if id == e.EventID {
			return nil, ErrInvalidTransition
		}
	}
	s := m.summary
	if e.AttemptID != s.AttemptID || e.Scope != s.Scope || e.TargetSlot != s.TargetSlot || e.TargetStateEpoch != s.TargetStateEpoch || e.ContentDigest != s.ProposedContentDigest {
		return nil, ErrStaleModel
	}
	if e.Kind == CandidateWritten {
		if !nonzero(e.BlobDigest) || e.BlobDigest == s.BaseBlobDigest {
			return nil, ErrStaleModel
		}
	} else if e.BlobDigest != s.CandidateBlobDigest {
		return nil, ErrStaleModel
	}
	source := slotAt(m.initial, s.TargetSlot).SourceID
	if p == CandidateDurable || p == ReadyToAdvance {
		source = m.initial.Counter.SourceID
	}
	if e.SourceID != source {
		return nil, ErrStaleModel
	}
	if e.Kind != BaseRechecked && e.Scan != (RecoveryInput{}) {
		return nil, ErrInvalidTransition
	}
	if e.Kind != CounterAdvanced && (e.CounterBefore != 0 || e.CounterAfter != 0) {
		return nil, ErrInvalidTransition
	}
	next := *m
	switch e.Kind {
	case CandidateWritten:
		next.summary.CandidateBlobDigest = e.BlobDigest
		next.summary.Phase = Written
	case FileSynced:
		next.summary.Phase = FileDurable
	case DirectorySynced:
		next.summary.Phase = CandidateDurable
	case BaseRechecked:
		if err := recheckBase(ctx, m, e.Scan); err != nil {
			return nil, err
		}
		next.recheckScanID = e.Scan.Counter.ScanID
		next.summary.Phase = ReadyToAdvance
	case CounterAdvanced:
		if e.CounterBefore != m.initial.Counter.Value || e.CounterAfter != s.TargetStateEpoch {
			return nil, ErrStaleModel
		}
		next.acknowledged = true
		next.interruptedPhase = p
		next.summary.Phase = NeedsRecovery
	case Failed, OutcomeUnknown:
		next.interruptedPhase = p
		next.summary.Phase = NeedsRecovery
	}
	next.eventIDs[next.eventCount] = e.EventID
	next.eventCount++
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	return &next, nil
}
