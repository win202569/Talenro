package staterecovery

import (
	"context"
	"errors"
	"testing"
)

func memoryInterrupted(t *testing.T, p CommitPhase, k EventKind) *CommitModel {
	t.Helper()
	m := memoryPhase(t, p)
	n, e := AdvanceModel(context.Background(), m, memoryEvent(m, k))
	if e != nil {
		t.Fatal(e)
	}
	return n
}
func memoryConfirmation(newState bool) ConfirmationInput {
	in := freshScan(memoryScan(), 70)
	in.B = memorySlot(in, 8)
	if newState {
		in.Counter.Value = 8
	}
	return ConfirmationInput{AttemptID: digest(31), Scan: in}
}
func requireConfirmError(t *testing.T, m *CommitModel, in ConfirmationInput, want error) {
	t.Helper()
	before := m.Summary()
	n, e := ConfirmModel(context.Background(), m, in)
	if n != nil || !errors.Is(e, want) || m.Summary() != before {
		t.Fatalf("confirm got %v want %v", e, want)
	}
}
func TestRecoveryConfirmOutcomes(t *testing.T) {
	for _, k := range []EventKind{Failed, OutcomeUnknown} {
		m := memoryInterrupted(t, ReadyToAdvance, k)
		for _, newState := range []bool{false, true} {
			in := memoryConfirmation(newState)
			n, e := ConfirmModel(context.Background(), m, in)
			if e != nil {
				t.Fatal(e)
			}
			want := NotCommitted
			slot := SlotA
			if newState {
				want = CommittedModel
				slot = SlotB
			}
			if n.Summary().Outcome != want || n.Summary().Recovery.SelectedSlot != slot || n.Summary().AttemptID != digest(31) {
				t.Fatalf("%+v", n.Summary())
			}
		}
	}
	m := memoryPhase(t, NeedsRecovery)
	n, e := ConfirmModel(context.Background(), m, memoryConfirmation(true))
	if e != nil || n.Summary().Outcome != CommittedModel {
		t.Fatal(e)
	}
	for _, p := range []CommitPhase{Prepared, Written, FileDurable, CandidateDurable} {
		m := memoryInterrupted(t, p, OutcomeUnknown)
		want := ErrStaleModel
		if p == Prepared {
			want = ErrAssuranceMissing
		}
		requireConfirmError(t, m, memoryConfirmation(true), want)
		n, e := ConfirmModel(context.Background(), m, memoryConfirmation(false))
		if e != nil || n.Summary().Outcome != NotCommitted {
			t.Fatal("early old", p, e)
		}
	}
}
func TestRecoveryConfirmAcknowledgedRollback(t *testing.T) {
	m := memoryPhase(t, NeedsRecovery)
	n, e := ConfirmModel(context.Background(), m, memoryConfirmation(false))
	var se *SecurityError
	if n != nil || !errors.Is(e, ErrSecurityFault) || !errors.As(e, &se) || se.Reason != CounterMismatch {
		t.Fatalf("rollback: %v", e)
	}
}
func TestRecoveryConfirmBindings(t *testing.T) {
	m := memoryInterrupted(t, ReadyToAdvance, OutcomeUnknown)
	for n := 0; n < 12; n++ {
		in := memoryConfirmation(true)
		want := ErrStaleModel
		switch n {
		case 0:
			in.AttemptID = digest(90)
		case 1:
			in.Scan = freshScan(in.Scan, 6)
		case 2:
			in.Scan = freshScan(in.Scan, 60)
		case 3:
			in.Scan.Counter.SourceID = digest(90)
		case 4:
			in.Scan.Counter.Maximum = 99
		case 5:
			in.Scan.A.SourceID = digest(90)
		case 6:
			in.Scan.B.SourceID = digest(90)
		case 7:
			in.Scan.B.BlobDigest = digest(90)
		case 8:
			in.Scan.B.ContentDigest = digest(90)
		case 9:
			in.Scan.A, in.Scan.B = in.Scan.B, in.Scan.A
			in.Scan.A.SourceID = digest(7)
			in.Scan.B.SourceID = digest(11)
		case 10:
			in.Scan.Counter.Value = 9
			in.Scan.B.StateEpoch = 9
			want = ErrSecurityFault
		case 11:
			in.Scan.B.Status = SlotCorrupt
			want = ErrSecurityFault
		}
		requireConfirmError(t, m, in, want)
	}
	in := memoryConfirmation(false)
	in.Scan.A.BlobDigest = digest(90)
	requireConfirmError(t, m, in, ErrStaleModel)
	in = memoryConfirmation(false)
	in.Scan.A.ContentDigest = digest(90)
	requireConfirmError(t, m, in, ErrStaleModel)
	in = memoryConfirmation(true)
	in.Scan.Counter.ObservationID = digest(91)
	in.Scan.A.ObservationID = digest(92)
	in.Scan.B.ObservationID = digest(93)
	n, e := ConfirmModel(context.Background(), m, in)
	if e != nil || n.Summary().Outcome != CommittedModel {
		t.Fatal(e)
	}
	requireConfirmError(t, memoryPrepared(t), memoryConfirmation(false), ErrInvalidTransition)
	p := memoryProposal(memoryScan())
	p.ProposedContentDigest = digest(9)
	nochange, e := PrepareModel(context.Background(), memoryScan(), p)
	if e != nil {
		t.Fatal(e)
	}
	requireConfirmError(t, nochange, memoryConfirmation(false), ErrInvalidTransition)
}
