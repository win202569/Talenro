package statetimebinding

import (
	"context"
	"errors"
	"talenro.local/platform/internal/nodecontrol/statetime"
	"testing"
	"time"
)

func TestTimeBindingObservations(t *testing.T) {
	in := timeBindingFixture(t, 1, 100)
	s := checkLowerModels(t, in)
	if e := bindObservations(context.Background(), in, s); e != nil {
		t.Fatal(e)
	}
	both := cloneFixture(t, in)
	both.Time.Metadata.StateID = d(99)
	both.Time.Checkpoint.StateID = d(99)
	control := checkLowerModels(t, both)
	if !errors.Is(bindObservations(context.Background(), both, control), ErrBindingMismatch) {
		t.Fatal("different base")
	}
	for _, change := range []func(*Input){
		func(x *Input) { x.Time.Metadata.NodeID[0] ^= 1 }, func(x *Input) { x.Time.Metadata.Epoch++ },
		func(x *Input) { x.Time.Request.NodeID[0] ^= 1 }, func(x *Input) { x.Time.Request.Epoch++ },
		func(x *Input) { x.Time.Checkpoint.NodeID[0] ^= 1 }, func(x *Input) { x.Time.Checkpoint.Epoch++ },
		func(x *Input) { x.Time.Metadata.StateID[0] ^= 1 }, func(x *Input) { x.Time.Checkpoint.StateID[0] ^= 1 },
	} {
		bad := cloneFixture(t, in)
		change(&bad)
		if !errors.Is(bindObservations(context.Background(), bad, s), ErrBindingMismatch) {
			t.Fatal("coordinate")
		}
	}
	for _, status := range []statetime.ObservationStatus{statetime.Unknown, statetime.Unavailable, statetime.Rollback, 99} {
		for n := 0; n < 4; n++ {
			bad := cloneFixture(t, in)
			switch n {
			case 0:
				bad.Time.Metadata.Status = status
			case 1:
				bad.Time.Time.Status = status
			case 2:
				bad.Time.Request.Status = status
			case 3:
				bad.Time.Checkpoint.Status = status
			}
			if !errors.Is(bindObservations(context.Background(), bad, s), ErrAssuranceMissing) {
				t.Fatal("missing observation")
			}
		}
	}
	independent := cloneFixture(t, in)
	independent.Time.Metadata.SourceID = d(96)
	independent.Time.Checkpoint.SourceID = d(97)
	checkLowerModels(t, independent)
	if e := bindObservations(context.Background(), independent, s); e != nil {
		t.Fatal("unjustified source equality", e)
	}
}
func TestTimeBindingTimeSnapshot(t *testing.T) {
	in := timeBindingFixture(t, 1, 100)
	s := checkLowerModels(t, in)
	for _, change := range []func(*Input){
		func(x *Input) { x.Time.Time.Now = x.Time.Time.Now.Add(time.Nanosecond) },
		func(x *Input) { x.Time.Time.Floor = x.Time.Time.Floor.Add(time.Nanosecond) },
		func(x *Input) { x.Time.Time.ObservationID[0] ^= 1 },
		func(x *Input) { x.Time.Time.ProviderID[0] ^= 1 },
	} {
		bad := cloneFixture(t, in)
		change(&bad)
		control := checkLowerModels(t, bad)
		if !errors.Is(bindObservations(context.Background(), bad, control), ErrBindingMismatch) {
			t.Fatal("different time")
		}
	}
	floors := cloneFixture(t, in)
	floors.Policy.Binding.Main.Previous.Scope.TimeFloor.Present = true
	floors.Policy.Binding.Main.Previous.Scope.TimeFloor.UnixSeconds = in.Policy.Policy.Old.TimeFloor.Add(-time.Hour).Unix()
	floors.Policy.Binding.Main.Proposed.Scope.TimeFloor = floors.Policy.Binding.Main.Previous.Scope.TimeFloor
	floors.Time.Time.Floor = floors.Time.Time.Floor.Add(time.Minute)
	floors.Policy.Policy.Time.Floor = floors.Time.Time.Floor
	control := checkLowerModels(t, floors)
	if e := bindObservations(context.Background(), floors, control); e != nil {
		t.Fatal("three floors", e)
	}
	if e := bindObservations(context.Background(), in, s); e != nil {
		t.Fatal(e)
	}
}
