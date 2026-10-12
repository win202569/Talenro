package staterecovery

import (
	"context"
	"errors"
	"testing"
)

// These independently constructed restart observations do not simulate real
// fsync, disk caches, TPM, or power loss. No durability probability is claimed.
func TestRecoveryCrashMatrix(t *testing.T) {
	cases := []struct {
		name    string
		counter uint64
		slot    SlotStatus
		epoch   uint64
		hint    int
		want    SlotID
		fault   error
	}{
		{"before write", 7, SlotAbsent, 0, 0, SlotA, nil},
		{"torn write", 7, SlotCorrupt, 0, 0, SlotUnknown, ErrSecurityFault},
		{"unsynced old bytes", 7, SlotIntact, 6, 0, SlotA, nil},
		{"unsynced absent", 7, SlotAbsent, 0, 0, SlotA, nil},
		{"unsynced future", 7, SlotIntact, 8, 0, SlotA, nil},
		{"unsynced corrupt", 7, SlotCorrupt, 0, 0, SlotUnknown, ErrSecurityFault},
		{"file and directory synced before increment", 7, SlotIntact, 8, 0, SlotA, nil},
		{"unknown increment old", 7, SlotIntact, 8, 0, SlotA, nil},
		{"unknown increment new", 8, SlotIntact, 8, 0, SlotB, nil},
		{"new counter old pointer", 8, SlotIntact, 8, 1, SlotB, nil},
		{"new counter no pointer", 8, SlotIntact, 8, 0, SlotB, nil},
		{"new counter missing candidate", 8, SlotAbsent, 0, 0, SlotUnknown, ErrSecurityFault},
		{"new counter corrupt candidate", 8, SlotCorrupt, 0, 0, SlotUnknown, ErrSecurityFault},
		{"duplicate epoch", 7, SlotIntact, 7, 0, SlotUnknown, ErrSecurityFault},
		{"missing observation", 7, SlotUnavailable, 0, 0, SlotUnknown, ErrAssuranceMissing},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := freshScan(memoryScan(), 70)
			in.Counter.Value = c.counter
			in.HasPointerHint = c.hint != 0
			in.PointerHint = c.hint
			if c.slot == SlotIntact {
				in.B = memorySlot(in, c.epoch)
			} else {
				in.B.Status = c.slot
			}
			m, e := RecoverModel(context.Background(), in)
			if c.fault != nil {
				if m != nil || !errors.Is(e, c.fault) {
					t.Fatal(e)
				}
				return
			}
			if e != nil || m.Summary().SelectedSlot != c.want || m.Summary().DiscardFuture != (c.counter == 7 && c.epoch == 8) {
				t.Fatalf("%v %v", m, e)
			}
		})
	}
	for _, field := range []int{0, 1, 2} {
		in := memoryConfirmation(true).Scan
		switch field {
		case 0:
			in.B.Scope.GuardID = digest(90)
		case 1:
			in.B.Scope.SealKeyID = digest(90)
		case 2:
			in.B.ScanID = digest(90)
		}
		want := ErrSecurityFault
		if field == 2 {
			want = ErrInvalidInput
		}
		requireRecoveryError(t, in, want, ReasonUnknown)
	}
}
