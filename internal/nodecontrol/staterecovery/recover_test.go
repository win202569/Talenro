package staterecovery

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
)

// Catches selection by pointer, first-good shortcuts, and implicit fallback.
func TestRecoverySelect(t *testing.T) {
	for _, epoch := range []uint64{0, 6, 8} {
		for _, swap := range []bool{false, true} {
			in := memoryScan()
			if epoch != 0 {
				in.B = memorySlot(in, epoch)
			}
			want := SlotA
			if swap {
				in.A, in.B = in.B, in.A
				want = SlotB
			}
			m, err := RecoverModel(context.Background(), in)
			if err != nil {
				t.Fatal(err)
			}
			got := m.Summary()
			if got.SelectedSlot != want || got.StateEpoch != 7 || got.BlobDigest != digest(10) || got.ContentDigest != digest(9) || got.Scope != in.ExpectedScope || got.DiscardFuture != (epoch == 8) {
				t.Fatalf("summary %+v", got)
			}
			wantFuture := SlotUnknown
			if epoch == 8 {
				wantFuture = SlotB
				if swap {
					wantFuture = SlotA
				}
			}
			if got.FutureSlot != wantFuture {
				t.Fatalf("epoch=%d swap=%t: future slot=%v want=%v", epoch, swap, got.FutureSlot, wantFuture)
			}
			m.Clear()
			m.Clear()
			if m.Summary() != (RecoverySummary{}) {
				t.Fatal("clear")
			}
		}
	}
	var m *RecoveryModel
	m.Clear()
	if m.Summary() != (RecoverySummary{}) {
		t.Fatal("nil summary")
	}
}

func TestRecoveryObservations(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*RecoveryInput)
		want   error
		reason SecurityReason
	}{
		{"counter unknown", func(i *RecoveryInput) { i.Counter.Status = CounterUnknown }, ErrAssuranceMissing, ReasonUnknown},
		{"counter unavailable", func(i *RecoveryInput) { i.Counter.Status = CounterUnavailable }, ErrAssuranceMissing, ReasonUnknown},
		{"counter rollback", func(i *RecoveryInput) { i.Counter.Status = CounterRollback }, ErrSecurityFault, CounterMismatch},
		{"counter enum", func(i *RecoveryInput) { i.Counter.Status = 99 }, ErrInvalidInput, ReasonUnknown},
		{"counter zero", func(i *RecoveryInput) { i.Counter.Value = 0 }, ErrInvalidInput, ReasonUnknown},
		{"capacity zero", func(i *RecoveryInput) { i.Counter.Maximum = 0 }, ErrInvalidInput, ReasonUnknown},
		{"capacity small", func(i *RecoveryInput) { i.Counter.Maximum = 6 }, ErrInvalidInput, ReasonUnknown},
		{"uuid nil", func(i *RecoveryInput) { i.ExpectedScope.NodeID = "00000000-0000-0000-0000-000000000000" }, ErrInvalidInput, ReasonUnknown},
		{"uuid case", func(i *RecoveryInput) { i.ExpectedScope.NodeID = "123E4567-e89b-42d3-a456-426614174000" }, ErrInvalidInput, ReasonUnknown},
		{"uuid variant", func(i *RecoveryInput) { i.ExpectedScope.NodeID = "123e4567-e89b-42d3-7456-426614174000" }, ErrInvalidInput, ReasonUnknown},
		{"authority zero", func(i *RecoveryInput) { i.ExpectedScope.AuthorityEpoch = 0 }, ErrInvalidInput, ReasonUnknown},
		{"authority overflow", func(i *RecoveryInput) { i.ExpectedScope.AuthorityEpoch = 1 << 63 }, ErrInvalidInput, ReasonUnknown},
		{"counter scope", func(i *RecoveryInput) { i.Counter.Scope.GuardID = digest(99) }, ErrSecurityFault, ScopeMismatch},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { in := memoryScan(); c.mutate(&in); requireRecoveryError(t, in, c.want, c.reason) })
	}
	for _, which := range []int{0, 1, 2} {
		for _, field := range []int{0, 1, 2} {
			in := memoryScan()
			var p *contracts.Digest
			if which == 0 {
				p = []*contracts.Digest{&in.Counter.SourceID, &in.Counter.ObservationID, &in.Counter.ScanID}[field]
			} else if which == 1 {
				p = []*contracts.Digest{&in.A.SourceID, &in.A.ObservationID, &in.A.ScanID}[field]
			} else {
				p = []*contracts.Digest{&in.B.SourceID, &in.B.ObservationID, &in.B.ScanID}[field]
			}
			*p = contracts.Digest{}
			requireRecoveryError(t, in, ErrInvalidInput, ReasonUnknown)
		}
	}
	for _, status := range []SlotStatus{SlotUnknownStatus, SlotUnavailable, SlotCorrupt, SlotUnsupported, 99} {
		for _, which := range []SlotID{SlotA, SlotB} {
			in := memoryScan()
			s := &in.A
			if which == SlotB {
				s = &in.B
			}
			s.Status = status
			want, reason := ErrSecurityFault, CorruptSlot
			if status == SlotUnknownStatus || status == SlotUnavailable {
				want, reason = ErrAssuranceMissing, ReasonUnknown
			}
			if status == SlotUnsupported {
				reason = UnsupportedVersion
			}
			requireRecoveryError(t, in, want, reason)
		}
	}
	for field := 0; field < 5; field++ {
		in := memoryScan()
		s := &in.A.Scope
		switch field {
		case 0:
			s.NodeID = "123e4567-e89b-42d3-a456-426614174001"
		case 1:
			s.AuthorityEpoch = 2
		case 2:
			s.GuardID = digest(90)
		case 3:
			s.SealKeyID = digest(90)
		case 4:
			s.DeploymentID = digest(90)
		}
		requireRecoveryError(t, in, ErrSecurityFault, ScopeMismatch)
	}
	for field := 0; field < 5; field++ {
		in := memoryScan()
		switch field {
		case 0:
			in.A.StateEpoch = 0
		case 1:
			in.A.ContentDigest = contracts.Digest{}
		case 2:
			in.A.BlobDigest = contracts.Digest{}
		case 3:
			in.A.ModelVersion = 0
		case 4:
			in.B.StateEpoch = 1
		}
		reason := CorruptSlot
		if field == 3 {
			reason = UnsupportedVersion
		}
		requireRecoveryError(t, in, ErrSecurityFault, reason)
	}
	in := memoryScan()
	in.B.ScanID = digest(99)
	requireRecoveryError(t, in, ErrInvalidInput, ReasonUnknown)
	for field := 0; field < 3; field++ {
		in = memoryScan()
		s := &in.ExpectedScope
		switch field {
		case 0:
			s.GuardID = contracts.Digest{}
		case 1:
			s.SealKeyID = contracts.Digest{}
		case 2:
			s.DeploymentID = contracts.Digest{}
		}
		requireRecoveryError(t, in, ErrInvalidInput, ReasonUnknown)
	}
}

func TestRecoveryDuplicateEpoch(t *testing.T) {
	in := memoryScan()
	in.B = in.A
	requireRecoveryError(t, in, ErrSecurityFault, DuplicateEpoch)
	in.B.BlobDigest = digest(99)
	requireRecoveryError(t, in, ErrSecurityFault, Fork)
	in.B.BlobDigest = in.A.BlobDigest
	in.B.ContentDigest = digest(99)
	requireRecoveryError(t, in, ErrSecurityFault, Fork)
}

// A complete blob identity includes its epoch. This catches accepting the same
// blob relabeled as either an older or a newer state, without needing raw bytes.
func TestRecoveryBlobIdentityEpoch(t *testing.T) {
	for _, counter := range []uint64{7, 8} {
		for _, sameContent := range []bool{false, true} {
			for _, swap := range []bool{false, true} {
				in := memoryScan()
				in.B = memorySlot(in, 8)
				in.B.BlobDigest = in.A.BlobDigest
				if sameContent {
					in.B.ContentDigest = in.A.ContentDigest
				}
				in.Counter.Value = counter
				if swap {
					in.A, in.B = in.B, in.A
				}
				requireRecoveryError(t, in, ErrSecurityFault, CorruptSlot)
			}
		}
	}
}

func TestRecoveryFutureAndPointer(t *testing.T) {
	in := memoryScan()
	in.B = memorySlot(in, 9)
	requireRecoveryError(t, in, ErrSecurityFault, UnexpectedFuture)
	in.B.StateEpoch = 8
	in.Counter.Maximum = 7
	requireRecoveryError(t, in, ErrSecurityFault, UnexpectedFuture)
	in = memoryScan()
	in.A.StateEpoch = 6
	in.B = memorySlot(in, 8)
	requireRecoveryError(t, in, ErrSecurityFault, CounterMismatch)
	in = memoryScan()
	in.B = memorySlot(in, 8)
	for _, hint := range []int{-1, 0, 1, 2, int(^uint(0) >> 1)} {
		in.HasPointerHint = true
		in.PointerHint = hint
		m, e := RecoverModel(context.Background(), in)
		if e != nil || m.Summary().SelectedSlot != SlotA {
			t.Fatalf("hint %d: %v", hint, e)
		}
	}
	in = memoryScan()
	in.Counter.Value = ^uint64(0)
	in.Counter.Maximum = ^uint64(0)
	in.A.StateEpoch = ^uint64(0)
	m, e := RecoverModel(context.Background(), in)
	if e != nil || m.Summary().StateEpoch != ^uint64(0) {
		t.Fatal(e)
	}
}
