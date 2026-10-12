package nodecerttime

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/statetimebinding"
	"testing"
	"time"
)

func TestCertTimeCrossNodeControls(t *testing.T) {
	a, b := certTimeFixture(t, "base"), certTimeFixture(t, "other-node")
	sa, sb := assertCertTimeLowerControls(t, a), assertCertTimeLowerControls(t, b)
	a.Time = b.Time
	s := Summary{Certificate: sa.Certificate, Time: sb.Time}
	if e := bindSnapshots(context.Background(), a, s); e != ErrBindingMismatch {
		t.Fatal("valid separate lower cross node", e)
	}
}
func TestCertTimeSnapshotRelations(t *testing.T) {
	// Private helper-only synthetic summary/input controls. Not a public trust path.
	base := certTimeFixture(t, "base")
	s := assertCertTimeLowerControls(t, base)
	if e := bindSnapshots(context.Background(), base, s); e != nil {
		t.Fatal(e)
	}
	inputMutations := []func(*Input){
		func(i *Input) { i.Certificate.Leaf.NodeID[0] ^= 1 },
		func(i *Input) { i.Time.Time.Metadata.NodeID[0] ^= 1 },
		func(i *Input) { i.Time.Time.Request.NodeID[0] ^= 1 },
		func(i *Input) { i.Time.Time.Checkpoint.NodeID[0] ^= 1 },
		func(i *Input) { i.Time.Policy.Binding.Main.Previous.Scope.NodeID = "different" },
		func(i *Input) { i.Time.Policy.Binding.Main.Proposed.Scope.NodeID = "different" },
		func(i *Input) { i.Time.Time.Time.ProviderID[0] ^= 1 },
		func(i *Input) { i.Time.Time.Time.ObservationID[0] ^= 1 },
		func(i *Input) { i.Time.Time.Time.Now = i.Time.Time.Time.Now.Add(time.Nanosecond) },
		func(i *Input) { i.Time.Time.Time.Floor = i.Time.Time.Time.Floor.Add(time.Nanosecond) },
		func(i *Input) { i.Time.Policy.Policy.Time.ProviderID[0] ^= 1 },
		func(i *Input) { i.Time.Policy.Policy.Time.ObservationID[0] ^= 1 },
		func(i *Input) { i.Time.Policy.Policy.Time.Now = i.Time.Policy.Policy.Time.Now.Add(time.Nanosecond) },
		func(i *Input) { i.Time.Policy.Policy.Time.Floor = i.Time.Policy.Policy.Time.Floor.Add(time.Nanosecond) },
	}
	for n, mutate := range inputMutations {
		in := certTimeFixture(t, "base")
		mutate(&in)
		if e := bindSnapshots(context.Background(), in, s); e != ErrBindingMismatch {
			t.Fatal("private input relation", n, e)
		}
	}
	summaryMutations := []func(*Summary){
		func(s *Summary) { s.Certificate.Leaf.Leaf.NodeID[0] ^= 1 },
		func(s *Summary) { s.Time.Time.NodeID[0] ^= 1 },
		func(s *Summary) { s.Time.Policy.Time.ProviderID[0] ^= 1 },
		func(s *Summary) { s.Time.Policy.Time.ObservationID[0] ^= 1 },
		func(s *Summary) { s.Time.Policy.Time.Now = s.Time.Policy.Time.Now.Add(time.Nanosecond) },
		func(s *Summary) {
			s.Time.Policy.Time.ObservedFloor = s.Time.Policy.Time.ObservedFloor.Add(time.Nanosecond)
		},
	}
	for n, mutate := range summaryMutations {
		bad := s
		mutate(&bad)
		if e := bindSnapshots(context.Background(), base, bad); e != ErrBindingMismatch {
			t.Fatal("private summary relation", n, e)
		}
	}
}
func TestCertTimeTimeTupleLowerControls(t *testing.T) {
	for n, mutate := range []func(*Input){
		func(i *Input) { i.Time.Time.Time.ProviderID[0] ^= 1 },
		func(i *Input) { i.Time.Time.Time.ObservationID[0] ^= 1 },
		func(i *Input) { i.Time.Time.Time.Now = i.Time.Time.Time.Now.Add(time.Nanosecond) },
		func(i *Input) { i.Time.Time.Time.Floor = i.Time.Time.Time.Floor.Add(time.Nanosecond) },
	} {
		in := certTimeFixture(t, "base")
		assertCertTimeLowerControls(t, in)
		mutate(&in)
		m, e := statetimebinding.Evaluate(context.Background(), in.Time)
		if m != nil {
			defer m.Clear()
		}
		if e == nil {
			t.Fatal("real lower tuple mismatch accepted", n)
		}
	}
}
