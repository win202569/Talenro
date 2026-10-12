package statetimebinding

import (
	"context"
	"errors"
	"talenro.local/platform/internal/nodecontrol/statetime"
	"testing"
)

func timeEvidence(t *testing.T, in Input) statetime.EvidenceSummary {
	t.Helper()
	m, e := statetime.EvaluateModel(context.Background(), in.Time)
	defer m.Clear()
	if e != nil || m == nil {
		t.Fatal(e)
	}
	return m.Summary()
}
func TestTimeBindingCheckpoint(t *testing.T) {
	for _, cp := range []uint64{100, 101} {
		in := timeBindingFixture(t, 1, cp)
		s := timeEvidence(t, in)
		if e := bindOldCheckpoint(context.Background(), in); e != nil {
			t.Fatal(e)
		}
		if e := bindEvidence(context.Background(), in, s); e != nil {
			t.Fatal(e)
		}
		for _, change := range []func(*statetime.EvidenceSummary){
			func(s *statetime.EvidenceSummary) { s.OldCheckpoint++ }, func(s *statetime.EvidenceSummary) { s.ProposedCheckpoint++ }, func(s *statetime.EvidenceSummary) { s.AdvanceRequired = !s.AdvanceRequired }, func(s *statetime.EvidenceSummary) { s.RequestID[0] ^= 1 }, func(s *statetime.EvidenceSummary) { s.Nonce[0] ^= 1 }, func(s *statetime.EvidenceSummary) { s.MetadataPayload[0] ^= 1 }, func(s *statetime.EvidenceSummary) { s.MetadataEnvelope[0] ^= 1 }, func(s *statetime.EvidenceSummary) { s.NodeID[0] ^= 1 }, func(s *statetime.EvidenceSummary) { s.Epoch++ },
		} {
			bad := s
			change(&bad)
			if !errors.Is(bindEvidence(context.Background(), in, bad), ErrBindingMismatch) {
				t.Fatal("evidence mismatch")
			}
		}
		bad := cloneFixture(t, in)
		bad.Time.Checkpoint.Sequence++
		if bindOldCheckpoint(context.Background(), bad) != ErrBindingMismatch {
			t.Fatal("old checkpoint")
		}
		bad = cloneFixture(t, in)
		bad.Policy.Binding.Main.Proposed.Checkpoint.Sequence++
		if bindEvidence(context.Background(), bad, s) != ErrBindingMismatch {
			t.Fatal("new checkpoint")
		}
		if in.Policy.Policy.Old.StateEpoch == in.Time.Checkpoint.Sequence {
			t.Fatal("fixture should distinguish roles")
		}
	}
}
