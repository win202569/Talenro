package staterecovery

import (
	"context"
	"errors"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
)

func memoryProposal(in RecoveryInput) ProposalObservation {
	return ProposalObservation{ProposalID: digest(30), AttemptID: digest(31), BaseBlobDigest: digest(10), ProposedContentDigest: digest(20), ExpectedScope: in.ExpectedScope, BaseStateEpoch: in.Counter.Value}
}
func TestRecoveryPrepareBindings(t *testing.T) {
	for _, swap := range []bool{false, true} {
		in := memoryScan()
		if swap {
			in.A, in.B = in.B, in.A
		}
		m, e := PrepareModel(context.Background(), in, memoryProposal(in))
		if e != nil {
			t.Fatal(e)
		}
		s := m.Summary()
		want := SlotB
		if swap {
			want = SlotA
		}
		if s.Phase != Prepared || s.TargetSlot != want || s.TargetStateEpoch != 8 || s.BaseBlobDigest != digest(10) || s.ProposedContentDigest != digest(20) {
			t.Fatalf("%+v", s)
		}
	}
	for n := 0; n < 7; n++ {
		in := memoryScan()
		p := memoryProposal(in)
		want := ErrInvalidInput
		switch n {
		case 0:
			p.ProposalID = contracts.Digest{}
		case 1:
			p.AttemptID = contracts.Digest{}
		case 2:
			p.BaseBlobDigest = contracts.Digest{}
		case 3:
			p.ProposedContentDigest = contracts.Digest{}
		case 4:
			p.BaseStateEpoch = 6
			want = ErrStaleModel
		case 5:
			p.BaseBlobDigest = digest(90)
			want = ErrStaleModel
		case 6:
			p.ExpectedScope.GuardID = digest(90)
			want = ErrStaleModel
		}
		m, e := PrepareModel(context.Background(), in, p)
		if m != nil || !errors.Is(e, want) {
			t.Fatalf("case %d: %v", n, e)
		}
	}
	in := memoryScan()
	in.B.Status = SlotCorrupt
	m, e := PrepareModel(context.Background(), in, memoryProposal(in))
	if m != nil || !errors.Is(e, ErrSecurityFault) {
		t.Fatal("prepare skipped recovery", e)
	}
}
func TestRecoveryPrepareNoChange(t *testing.T) {
	in := memoryScan()
	p := memoryProposal(in)
	p.ProposedContentDigest = digest(9)
	for _, hint := range []int{-1, 0, 2, 100} {
		in.HasPointerHint = true
		in.PointerHint = hint
		before := in
		m, e := PrepareModel(context.Background(), in, p)
		if e != nil {
			t.Fatal(e)
		}
		s := m.Summary()
		if s.Phase != NoChange || s.TargetSlot != SlotUnknown || s.TargetStateEpoch != 0 || s.CandidateBlobDigest != (contracts.Digest{}) || in != before {
			t.Fatalf("not zero action: %+v", s)
		}
	}
}
func TestRecoveryPrepareCapacity(t *testing.T) {
	for _, max := range []uint64{7, ^uint64(0)} {
		in := memoryScan()
		in.Counter.Value = max
		in.Counter.Maximum = max
		in.A.StateEpoch = max
		p := memoryProposal(in)
		m, e := PrepareModel(context.Background(), in, p)
		if m != nil || !errors.Is(e, ErrCounterExhausted) {
			t.Fatalf("max %d: %v", max, e)
		}
		p.ProposedContentDigest = digest(9)
		m, e = PrepareModel(context.Background(), in, p)
		if e != nil || m.Summary().Phase != NoChange {
			t.Fatal("nochange at capacity", e)
		}
	}
	in := memoryScan()
	in.Counter.Maximum = ^uint64(0)
	in.Counter.Value = ^uint64(0) - 1
	in.A.StateEpoch = in.Counter.Value
	m, e := PrepareModel(context.Background(), in, memoryProposal(in))
	if e != nil || m.Summary().TargetStateEpoch != ^uint64(0) {
		t.Fatal("checked increment", e)
	}
	in.Counter.Maximum = 0
	m, e = PrepareModel(context.Background(), in, memoryProposal(in))
	if m != nil || !errors.Is(e, ErrInvalidInput) {
		t.Fatal(e)
	}
}
