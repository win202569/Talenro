package staterecovery

import (
	"context"
	"errors"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
)

func memoryPrepared(t *testing.T) *CommitModel {
	t.Helper()
	in := memoryScan()
	m, e := PrepareModel(context.Background(), in, memoryProposal(in))
	if e != nil {
		t.Fatal(e)
	}
	return m
}

// memoryEvent supplies assertions, not actual I/O acknowledgments.
func memoryEvent(m *CommitModel, kind EventKind) Event {
	s := m.Summary()
	e := Event{Kind: kind, Phase: s.Phase, AttemptID: s.AttemptID, EventID: digest(byte(s.Phase) + 40), SourceID: digest(11), Scope: s.Scope, TargetSlot: s.TargetSlot, TargetStateEpoch: s.TargetStateEpoch, ContentDigest: s.ProposedContentDigest, BlobDigest: s.CandidateBlobDigest}
	switch kind {
	case CandidateWritten:
		e.BlobDigest = digest(21)
	case BaseRechecked:
		e.SourceID = digest(4)
		e.Scan = freshScan(memoryScan(), 60)
		e.Scan.B = memorySlot(e.Scan, 8)
	case CounterAdvanced:
		e.SourceID = digest(4)
		e.CounterBefore = 7
		e.CounterAfter = 8
	case Failed, OutcomeUnknown:
		if s.Phase == CandidateDurable || s.Phase == ReadyToAdvance {
			e.SourceID = digest(4)
		}
	}
	return e
}
func memoryPhase(t *testing.T, p CommitPhase) *CommitModel {
	t.Helper()
	m := memoryPrepared(t)
	for _, k := range []EventKind{CandidateWritten, FileSynced, DirectorySynced, BaseRechecked, CounterAdvanced} {
		if m.Summary().Phase == p {
			return m
		}
		n, e := AdvanceModel(context.Background(), m, memoryEvent(m, k))
		if e != nil {
			t.Fatal(e)
		}
		m = n
	}
	if m.Summary().Phase != p {
		t.Fatal("invalid fixture phase")
	}
	return m
}
func requireAdvanceError(t *testing.T, m *CommitModel, e Event, want error) {
	t.Helper()
	before := m.Summary()
	n, err := AdvanceModel(context.Background(), m, e)
	if n != nil || !errors.Is(err, want) || m.Summary() != before {
		t.Fatalf("advance got %v want %v", err, want)
	}
}

func TestRecoveryAdvanceSequence(t *testing.T) {
	m := memoryPrepared(t)
	for i, k := range []EventKind{CandidateWritten, FileSynced, DirectorySynced, BaseRechecked, CounterAdvanced} {
		before := m.Summary()
		n, e := AdvanceModel(context.Background(), m, memoryEvent(m, k))
		if e != nil {
			t.Fatal(e)
		}
		if m.Summary() != before || n.Summary().Phase != []CommitPhase{Written, FileDurable, CandidateDurable, ReadyToAdvance, NeedsRecovery}[i] || n.Summary().CandidateBlobDigest != digest(21) {
			t.Fatalf("step %d", i)
		}
		m = n
	}
	for _, p := range []CommitPhase{Prepared, Written, FileDurable, CandidateDurable, ReadyToAdvance} {
		for _, k := range []EventKind{Failed, OutcomeUnknown} {
			m := memoryPhase(t, p)
			n, e := AdvanceModel(context.Background(), m, memoryEvent(m, k))
			if e != nil || n.Summary().Phase != NeedsRecovery {
				t.Fatalf("failure %d/%d: %v", p, k, e)
			}
		}
	}
	in := memoryScan()
	proposal := memoryProposal(in)
	proposal.ProposedContentDigest = digest(9)
	nochange, e := PrepareModel(context.Background(), in, proposal)
	if e != nil {
		t.Fatal(e)
	}
	requireAdvanceError(t, nochange, Event{}, ErrInvalidTransition)
	requireAdvanceError(t, m, Event{}, ErrInvalidTransition)
	requireAdvanceError(t, &CommitModel{}, Event{}, ErrStaleModel)
}
func TestRecoveryAdvanceRejections(t *testing.T) {
	phases := []CommitPhase{Prepared, Written, FileDurable, CandidateDurable, ReadyToAdvance}
	for idx, p := range phases {
		m := memoryPhase(t, p)
		for k := EventUnknown; k <= OutcomeUnknown+1; k++ {
			if k == []EventKind{CandidateWritten, FileSynced, DirectorySynced, BaseRechecked, CounterAdvanced}[idx] || k == Failed || k == OutcomeUnknown {
				continue
			}
			requireAdvanceError(t, m, memoryEvent(m, k), ErrInvalidTransition)
		}
	}
	for _, p := range phases {
		m := memoryPhase(t, p)
		kind := map[CommitPhase]EventKind{Prepared: CandidateWritten, Written: FileSynced, FileDurable: DirectorySynced, CandidateDurable: BaseRechecked, ReadyToAdvance: CounterAdvanced}[p]
		for n := 0; n < 12; n++ {
			e := memoryEvent(m, kind)
			want := ErrStaleModel
			switch n {
			case 0:
				e.Phase = PhaseUnknown
				want = ErrInvalidTransition
			case 1:
				e.AttemptID = digest(90)
			case 2:
				e.EventID = contracts.Digest{}
				want = ErrInvalidInput
			case 3:
				e.SourceID = contracts.Digest{}
				want = ErrInvalidInput
			case 4:
				e.SourceID = digest(90)
			case 5:
				e.Scope.GuardID = digest(90)
			case 6:
				e.TargetSlot = SlotA
			case 7:
				e.TargetStateEpoch = 9
			case 8:
				e.ContentDigest = digest(90)
			case 9:
				if p == Prepared {
					e.BlobDigest = digest(10)
				} else {
					e.BlobDigest = digest(90)
				}
			case 10:
				if kind == BaseRechecked {
					e.CounterBefore = 7
				} else {
					e.Scan = memoryScan()
				}
				want = ErrInvalidTransition
			case 11:
				if kind == CounterAdvanced {
					e.CounterAfter = 9
				} else {
					e.CounterAfter = 8
				}
				if kind != CounterAdvanced {
					want = ErrInvalidTransition
				}
			}
			requireAdvanceError(t, m, e, want)
		}
		if p != Prepared {
			for old := Prepared; old < p; old++ {
				e := memoryEvent(m, kind)
				e.EventID = digest(byte(old) + 40)
				requireAdvanceError(t, m, e, ErrInvalidTransition)
			}
		}
	}
	m := memoryPrepared(t)
	e := memoryEvent(m, Failed)
	e.BlobDigest = digest(21)
	requireAdvanceError(t, m, e, ErrStaleModel)
}
func TestRecoveryRecheckBindings(t *testing.T) {
	m := memoryPhase(t, CandidateDurable)
	for n := 0; n < 10; n++ {
		e := memoryEvent(m, BaseRechecked)
		switch n {
		case 0:
			e.Scan = freshScan(e.Scan, 6)
		case 1:
			e.Scan.Counter.Maximum = 99
		case 2:
			e.Scan.Counter.SourceID = digest(90)
		case 3:
			e.Scan.A.SourceID = digest(90)
		case 4:
			e.Scan.B.SourceID = digest(90)
		case 5:
			e.Scan.A.BlobDigest = digest(90)
		case 6:
			e.Scan.A.ContentDigest = digest(90)
		case 7:
			e.Scan.B.BlobDigest = digest(90)
		case 8:
			e.Scan.B.ContentDigest = digest(90)
		case 9:
			e.Scan.Counter.Value = 8
		}
		requireAdvanceError(t, m, e, ErrStaleModel)
	}
	e := memoryEvent(m, BaseRechecked)
	e.Scan.Counter.ObservationID = digest(91)
	e.Scan.A.ObservationID = digest(92)
	e.Scan.B.ObservationID = digest(93)
	n, err := AdvanceModel(context.Background(), m, e)
	if err != nil || n.Summary().Phase != ReadyToAdvance {
		t.Fatal(err)
	}
}
