package statepolicybinding

import (
	"context"
	"errors"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
	"testing"
	"time"
)

func requireBindingEvaluation(t *testing.T, in Input, want error) {
	t.Helper()
	m, e := Evaluate(context.Background(), in)
	if want != nil {
		if m != nil || !errors.Is(e, want) {
			t.Fatalf("model=%v error=%v want=%v", m, e, want)
		}
		return
	}
	if e != nil || m == nil {
		t.Fatal(e)
	}
	defer m.Clear()
}
func TestPolicyBindingTime(t *testing.T) {
	for _, present := range []bool{false, true} {
		for _, floorHigher := range []bool{false, true} {
			in := policyBindingFixture(t, 1, 1)
			if present {
				f := statecontract.OptionalTime{Present: true, UnixSeconds: in.Policy.Old.TimeFloor.Unix()}
				in.Binding.Main.Previous.Scope.TimeFloor = f
				in.Binding.Main.Proposed.Scope.TimeFloor = f
			}
			if floorHigher {
				in.Policy.Time.Floor = in.Policy.Old.TimeFloor.Add(time.Nanosecond)
				in.Policy.Time.Now = in.Policy.Time.Floor.Add(25 * time.Hour)
			}
			checkBindingInput(t, in)
			checkPolicyInput(t, in)
			requireBindingEvaluation(t, in, nil)
		}
	}
	for _, sec := range []int64{-62135596801, 253402300800} {
		in := policyBindingFixture(t, 0, 0)
		in.Binding.Main.Previous.Scope.TimeFloor = statecontract.OptionalTime{Present: true, UnixSeconds: sec}
		in.Binding.Main.Proposed.Scope.TimeFloor = in.Binding.Main.Previous.Scope.TimeFloor
		if e := bindTime(context.Background(), in); !errors.Is(e, ErrInvalidInput) {
			t.Fatal("checked seconds", e)
		}
	}
	in := policyBindingFixture(t, 1, 1)
	f := statecontract.OptionalTime{Present: true, UnixSeconds: in.Policy.Old.TimeFloor.Unix() + 1}
	in.Binding.Main.Previous.Scope.TimeFloor = f
	in.Binding.Main.Proposed.Scope.TimeFloor = f
	checkBindingInput(t, in)
	if e := bindTime(context.Background(), in); !errors.Is(e, ErrBindingMismatch) {
		t.Fatal(e)
	}
	for _, change := range []func(*Input){
		func(x *Input) { x.Policy.Time.Floor = x.Policy.Old.TimeFloor.Add(-time.Nanosecond) },
		func(x *Input) { x.Policy.Time.Now = x.Policy.Time.Floor.Add(-time.Nanosecond) },
	} {
		x := policyBindingFixture(t, 1, 1)
		change(&x)
		checkBindingInput(t, x)
		assertPolicyFailure(t, x, statepolicy.SecurityFault, statepolicy.StateRollback)
	}
	for _, change := range []func(*Input){
		func(x *Input) { x.Policy.Time.Now = time.Time{} }, func(x *Input) { x.Policy.Time.Floor = time.Time{} },
		func(x *Input) { x.Policy.Time.Now = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) },
		func(x *Input) { x.Policy.Time.Now = fixtureTime(t, "2020-01-02T00:00:00Z") },
		func(x *Input) { x.Policy.Time.ObservationID = [32]byte{} },
	} {
		x := policyBindingFixture(t, 1, 1)
		change(&x)
		m, e := Evaluate(context.Background(), x)
		if m != nil || e == nil {
			t.Fatal("invalid/expired observation accepted")
		}
	}
	for _, change := range []func(*Input){
		func(x *Input) { x.Binding.Main.Proposed.Scope.TimeFloor.UnixSeconds++ },
		func(x *Input) { x.Binding.Main.Proposed.Scope.TimeFloor = statecontract.OptionalTime{} },
	} {
		x := policyBindingFixture(t, 1, 1)
		x.Binding.Main.Previous.Scope.TimeFloor = statecontract.OptionalTime{Present: true, UnixSeconds: x.Policy.Old.TimeFloor.Unix()}
		x.Binding.Main.Proposed.Scope.TimeFloor = x.Binding.Main.Previous.Scope.TimeFloor
		change(&x)
		m, e := Evaluate(context.Background(), x)
		if m != nil || e == nil {
			t.Fatal("floor change accepted")
		}
		if x.Binding.Main.Proposed.Scope.TimeFloor.Present {
			if !errors.Is(e, ErrUnsupportedPolicyShape) {
				t.Fatal(e)
			}
		} else if !errors.Is(e, ErrRollback) {
			t.Fatal("lower rollback lost", e)
		}
	}
}
func TestPolicyBindingIdentity(t *testing.T) {
	in := policyBindingFixture(t, 1, 1)
	if in.Binding.Main.Proposed.Scope.DeploymentID == in.Policy.Deployment.SourceID || in.Binding.Main.Proposed.Scope.TrustAnchorID == in.Policy.Old.AnchorPayload || in.Binding.Main.Proposed.Scope.Main.CounterID == in.Policy.Old.CounterID || in.Binding.Main.Proposed.Checkpoint.Sequence == in.Policy.Old.StateEpoch {
		t.Fatal("fixture opaque identities coincide")
	}
	requireBindingEvaluation(t, in, nil)
	for _, change := range []func(*Input){
		func(x *Input) { x.Policy.Old.NodeID[0] ^= 1 }, func(x *Input) { x.Policy.Old.Epoch++ }, func(x *Input) { x.Policy.Old.AnchorPayload[0] ^= 1 },
		func(x *Input) { x.Policy.Deployment.NodeID[0] ^= 1 }, func(x *Input) { x.Policy.Deployment.Epoch++ }, func(x *Input) { x.Policy.Deployment.AnchorPayload[0] ^= 1 },
		func(x *Input) { x.Policy.Deployment.SourceID[0] ^= 1 }, func(x *Input) { x.Policy.Old.TimeProviderID[0] ^= 1 }, func(x *Input) { x.Policy.Time.ProviderID[0] ^= 1 },
	} {
		x := copyPolicyBindingInput(t, in)
		change(&x)
		checkBindingInput(t, x)
		requireBindingEvaluation(t, x, ErrBindingMismatch)
	}
	state := copyPolicyBindingInput(t, in).Policy.Old
	state.Status = statepolicy.Unknown
	if e := bindStateIdentity(context.Background(), in, state, true); e != nil {
		t.Fatal(e)
	}
	for _, change := range []func(*statepolicy.StateModel){
		func(s *statepolicy.StateModel) { s.Status = statepolicy.Available }, func(s *statepolicy.StateModel) { s.StateID[0] ^= 1 }, func(s *statepolicy.StateModel) { s.CounterID[0] ^= 1 }, func(s *statepolicy.StateModel) { s.StateEpoch++ }, func(s *statepolicy.StateModel) { s.TimeFloor = s.TimeFloor.Add(time.Nanosecond) },
	} {
		s := copyPolicyBindingInput(t, in).Policy.Old
		s.Status = statepolicy.Unknown
		change(s)
		if e := bindStateIdentity(context.Background(), in, s, true); !errors.Is(e, ErrBindingMismatch) {
			t.Fatal(e)
		}
	}
}
