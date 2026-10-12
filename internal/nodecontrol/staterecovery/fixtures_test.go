package staterecovery

import (
	"context"
	"errors"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
)

func digest(n byte) contracts.Digest { return contracts.Digest{n} }

// memoryScan is an in-process assumption fixture, never a real provider.
func memoryScan() RecoveryInput {
	s := Scope{NodeID: "123e4567-e89b-42d3-a456-426614174000", AuthorityEpoch: 1, GuardID: digest(1), SealKeyID: digest(2), DeploymentID: digest(3)}
	return RecoveryInput{ExpectedScope: s,
		Counter: CounterObservation{Status: CounterAvailable, SourceID: digest(4), ObservationID: digest(5), ScanID: digest(6), Scope: s, Value: 7, Maximum: 100},
		A:       SlotObservation{Status: SlotIntact, SourceID: digest(7), ObservationID: digest(8), ScanID: digest(6), Scope: s, StateEpoch: 7, ContentDigest: digest(9), BlobDigest: digest(10), ModelVersion: 1},
		B:       SlotObservation{Status: SlotAbsent, SourceID: digest(11), ObservationID: digest(12), ScanID: digest(6)}}
}
func memorySlot(in RecoveryInput, epoch uint64) SlotObservation {
	s := in.A
	s.SourceID = in.B.SourceID
	s.ObservationID = in.B.ObservationID
	s.StateEpoch = epoch
	s.ContentDigest = digest(20)
	s.BlobDigest = digest(21)
	return s
}
func freshScan(in RecoveryInput, id byte) RecoveryInput {
	in.Counter.ScanID = digest(id)
	in.A.ScanID = digest(id)
	in.B.ScanID = digest(id)
	return in
}
func requireRecoveryError(t *testing.T, in RecoveryInput, want error, reason SecurityReason) {
	t.Helper()
	m, err := RecoverModel(context.Background(), in)
	if m != nil || !errors.Is(err, want) {
		t.Fatalf("got model=%v err=%v want=%v", m, err, want)
	}
	if reason != ReasonUnknown {
		var se *SecurityError
		if !errors.As(err, &se) || se.Reason != reason {
			t.Fatalf("reason: %v", err)
		}
	}
}
