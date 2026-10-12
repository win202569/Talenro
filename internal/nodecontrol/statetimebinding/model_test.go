package statetimebinding

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"testing"
	"time"
)

func zeroMutableSlices(v reflect.Value) {
	if v.Type() == reflect.TypeOf(Summary{}.ValidUntil) {
		return
	}
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			zeroMutableSlices(v.Elem())
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).CanSet() {
				zeroMutableSlices(v.Field(i))
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			zeroMutableSlices(v.Index(i))
			v.Index(i).Set(reflect.Zero(v.Type().Elem()))
		}
	}
}
func TestTimeBindingSummaryOwnership(t *testing.T) {
	in := historyTimeInput(t)
	m := evaluated(t, in)
	before := jsonInput(t, m.input)
	s := m.Summary()
	detached := m.Summary()
	zeroMutableSlices(reflect.ValueOf(&detached).Elem())
	zeroMutableSlices(reflect.ValueOf(&in).Elem())
	if !reflect.DeepEqual(m.Summary(), s) || !bytes.Equal(before, jsonInput(t, m.input)) {
		t.Fatal("alias")
	}
	m.Clear()
	if s.Policy.Binding.Proposed.Target.Links[0].NewRoot.Payload == ([32]byte{}) {
		t.Fatal("detached cleared")
	}
}
func TestTimeBindingModelClear(t *testing.T) {
	in := historyTimeInput(t)
	before := jsonInput(t, in)
	m := evaluated(t, in)
	detached := m.Summary()
	expected := m.Summary()
	held := append(mutableSlices(reflect.ValueOf(m.input)), mutableSlices(reflect.ValueOf(m.summary))...)
	m.Clear()
	m.Clear()
	for _, slice := range held {
		for i := 0; i < slice.Len(); i++ {
			if !slice.Index(i).IsZero() {
				t.Fatal("owned backing store not zero", slice.Type())
			}
		}
	}
	if m.valid || !reflect.DeepEqual(m.input, Input{}) || !reflect.DeepEqual(m.summary, Summary{}) || !bytes.Equal(before, jsonInput(t, in)) || !reflect.DeepEqual(detached, expected) {
		t.Fatal("clear scope")
	}
	for _, zero := range []*Model{nil, {}} {
		zero.Clear()
		if !reflect.DeepEqual(zero.Summary(), Summary{}) {
			t.Fatal("zero owner")
		}
	}
	if _, e := ownInput(context.Background(), Input{}); e != ErrInvalidInput {
		t.Fatal("empty input")
	}
}

func legalReplacement(t *testing.T, m *Model, in Input) {
	t.Helper()
	evaluated(t, in)
	if e := Recheck(context.Background(), m, in); !errors.Is(e, ErrBindingMismatch) {
		t.Fatal("changed snapshot reused", e)
	}
}
func TestTimeBindingRecheck(t *testing.T) {
	in := historyTimeInput(t)
	m := evaluated(t, in)
	same := cloneFixture(t, in)
	for _, table := range []*statematerials.Input{&same.Policy.Binding.Previous, &same.Policy.Binding.Proposed.Materials} {
		for i, j := 0, len(table.Materials)-1; i < j; i, j = i+1, j-1 {
			table.Materials[i], table.Materials[j] = table.Materials[j], table.Materials[i]
		}
		for i, j := 0, len(table.Expected)-1; i < j; i, j = i+1, j-1 {
			table.Expected[i], table.Expected[j] = table.Expected[j], table.Expected[i]
		}
	}
	same.Time.Time.Now = same.Time.Time.Now.In(time.FixedZone("same", 3600))
	same.Policy.Policy.Time.Now = same.Time.Time.Now
	same.Time.Time.Floor = same.Time.Time.Floor.In(time.FixedZone("same", -3600))
	same.Policy.Policy.Time.Floor = same.Time.Time.Floor
	same.Policy.Policy.Old.TimeFloor = same.Policy.Policy.Old.TimeFloor.In(time.FixedZone("same", 7200))
	same.Policy.Binding.Main.Previous.Recovery.Local = []statecontract.LocalFaultBinding{}
	same.Policy.Binding.Main.Proposed.Recovery.Local = []statecontract.LocalFaultBinding{}
	if e := Recheck(context.Background(), m, same); e != nil {
		t.Fatal("equivalent", e)
	}
	for _, change := range []func(*Input){
		func(x *Input) { x.Policy.Policy.Old.CounterID[0] ^= 1 }, func(x *Input) { x.Policy.Policy.Old.StateEpoch++ },
		func(x *Input) { x.Policy.Policy.Deployment.ObservationID[0] ^= 1 },
		func(x *Input) { x.Policy.Policy.History[0].ObservationID[0] ^= 1 },
		func(x *Input) { x.Policy.Policy.History[0].RetiringKnown = !x.Policy.Policy.History[0].RetiringKnown },
		func(x *Input) {
			x.Policy.Policy.History[0], x.Policy.Policy.History[1] = x.Policy.Policy.History[1], x.Policy.Policy.History[0]
		},
		func(x *Input) {
			x.Policy.Binding.Main.Previous.Identity.History.Payload[0] ^= 1
			x.Policy.Binding.Main.Proposed.Identity.History = x.Policy.Binding.Main.Previous.Identity.History
		},
		func(x *Input) {
			x.Policy.Binding.Main.Previous.Metadata.History.Payload[0] ^= 1
			x.Policy.Binding.Main.Proposed.Metadata.History = x.Policy.Binding.Main.Previous.Metadata.History
		},
		func(x *Input) {
			x.Policy.Binding.Main.Previous.Metadata.RevokedSetDigest[0] ^= 1
			x.Policy.Binding.Main.Proposed.Metadata.RevokedSetDigest = x.Policy.Binding.Main.Previous.Metadata.RevokedSetDigest
		},
		func(x *Input) {
			x.Policy.Binding.Main.Previous.Scope.DeploymentID = d(98)
			x.Policy.Binding.Main.Proposed.Scope.DeploymentID = d(98)
		},
	} {
		x := cloneFixture(t, in)
		change(&x)
		legalReplacement(t, m, x)
	}
	x := cloneFixture(t, in)
	unselected := object(t, frozen(t, "metadata.payload.json"))
	unselected["metadata_version"] = "9"
	unselected["authority_sequence"] = "11"
	addEnvelope(t, &x.Policy.Binding.Proposed, statematerials.Metadata, mint(t, "metadata", unselected, private(1)), statematerials.MaterialRef{}, tailRef(x.Policy.Binding.Proposed), statematerials.MaterialRef{})
	legalReplacement(t, m, x)
	m2 := evaluated(t, x)
	y := cloneFixture(t, x)
	for _, mat := range y.Policy.Binding.Proposed.Materials.Materials {
		if mat.Ref.Kind == statematerials.Metadata && mat.Ref != y.Policy.Binding.Proposed.Target.Metadata {
			removeSelected(&y.Policy.Binding.Proposed, mat.Ref)
		}
	}
	unselected["metadata_version"] = "10"
	unselected["authority_sequence"] = "12"
	addEnvelope(t, &y.Policy.Binding.Proposed, statematerials.Metadata, mint(t, "metadata", unselected, private(1)), statematerials.MaterialRef{}, tailRef(y.Policy.Binding.Proposed), statematerials.MaterialRef{})
	legalReplacement(t, m2, y)
	if Recheck(nil, nil, in) != ErrInvalidInput || Recheck(context.Background(), nil, in) != ErrClearedModel {
		t.Fatal("nil owner/context")
	}
	m.Clear()
	if Recheck(context.Background(), m, in) != ErrClearedModel {
		t.Fatal("cleared")
	}
}
func TestTimeBindingObservationRecheck(t *testing.T) {
	in := historyTimeInput(t)
	m := evaluated(t, in)
	for _, change := range []func(*Input){
		func(x *Input) { x.Time.Metadata.SourceID[0] ^= 1 }, func(x *Input) { x.Time.Metadata.ObservationID[0] ^= 1 },
		func(x *Input) { x.Time.Checkpoint.SourceID[0] ^= 1 }, func(x *Input) { x.Time.Checkpoint.ObservationID[0] ^= 1 },
		func(x *Input) { x.Time.Request.RequestID[0] ^= 1 }, func(x *Input) { x.Time.Request.ObservationID[0] ^= 1 },
		func(x *Input) {
			x.Time.Request.ExpectedNonce[0] = 1
			signAttestation(t, x, 100, x.Time.Request.ExpectedNonce)
		},
		func(x *Input) {
			x.Policy.Policy.Old.StateID = d(99)
			x.Time.Metadata.StateID = d(99)
			x.Time.Checkpoint.StateID = d(99)
			for i := range x.Policy.Policy.History {
				x.Policy.Policy.History[i].BaseStateID = d(99)
			}
		},
		func(x *Input) {
			x.Time.Time.ProviderID = d(99)
			x.Policy.Policy.Time.ProviderID = d(99)
			x.Policy.Policy.Old.TimeProviderID = d(99)
			x.Policy.Binding.Main.Previous.Scope.TimeProviderID = d(99)
			x.Policy.Binding.Main.Proposed.Scope.TimeProviderID = d(99)
		},
	} {
		x := cloneFixture(t, in)
		change(&x)
		legalReplacement(t, m, x)
	}
}
func TestTimeBindingFixedTime(t *testing.T) {
	in := timeBindingFixture(t, 1, 100)
	m := evaluated(t, in)
	for _, change := range []func(*Input){
		func(x *Input) {
			x.Time.Time.Now = x.Time.Time.Now.Add(time.Nanosecond)
			x.Policy.Policy.Time.Now = x.Time.Time.Now
		},
		func(x *Input) {
			x.Time.Time.Floor = x.Time.Time.Floor.Add(time.Nanosecond)
			x.Policy.Policy.Time.Floor = x.Time.Time.Floor
		},
		func(x *Input) {
			x.Time.Time.ObservationID[0] ^= 1
			x.Policy.Policy.Time.ObservationID = x.Time.Time.ObservationID
		},
	} {
		x := cloneFixture(t, in)
		change(&x)
		legalReplacement(t, m, x)
	}
}

type traceTimeBindingContext struct {
	calls, at int
	failure   error
	panicAt   bool
}

func (c *traceTimeBindingContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c *traceTimeBindingContext) Done() <-chan struct{}       { return nil }
func (c *traceTimeBindingContext) Value(any) any               { return nil }
func (c *traceTimeBindingContext) Err() error {
	c.calls++
	if c.at > 0 && c.calls >= c.at {
		if c.panicAt {
			panic("trace panic")
		}
		return c.failure
	}
	return nil
}
func traceCount(t *testing.T, in Input, recheck bool, m *Model) int {
	t.Helper()
	ctx := &traceTimeBindingContext{}
	if recheck {
		if e := Recheck(ctx, m, in); e != nil {
			t.Fatal(e)
		}
	} else {
		model, e := Evaluate(ctx, in)
		defer model.Clear()
		if e != nil {
			t.Fatal(e)
		}
	}
	if ctx.calls == 0 {
		t.Fatal("no trace")
	}
	return ctx.calls
}
func TestTimeBindingCancellation(t *testing.T) {
	in := timeBindingFixture(t, 0, 100)
	before := jsonInput(t, in)
	m := evaluated(t, in)
	for _, e := range []error{context.Canceled, context.DeadlineExceeded} {
		for _, recheck := range []bool{false, true} {
			count := traceCount(t, in, recheck, m)
			for i := 1; i <= count; i++ {
				ctx := &traceTimeBindingContext{at: i, failure: e}
				if recheck {
					if got := Recheck(ctx, m, in); !errors.Is(got, e) {
						t.Fatal(i, got)
					}
				} else {
					got, err := Evaluate(ctx, in)
					if got != nil || !errors.Is(err, e) {
						t.Fatal(i, err)
					}
				}
			}
		}
	}
	if !bytes.Equal(before, jsonInput(t, in)) || !m.valid {
		t.Fatal("caller/prior model changed")
	}
}
func TestTimeBindingLateCancellation(t *testing.T) {
	in := timeBindingFixture(t, 1, 100)
	m := evaluated(t, in)
	for _, recheck := range []bool{false, true} {
		ctx := &traceTimeBindingContext{at: traceCount(t, in, recheck, m), failure: context.DeadlineExceeded}
		if recheck {
			if e := Recheck(ctx, m, in); e != context.DeadlineExceeded {
				t.Fatal(e)
			}
		} else {
			got, e := Evaluate(ctx, in)
			if got != nil || e != context.DeadlineExceeded {
				t.Fatal(e)
			}
		}
	}
}
func TestTimeBindingPanicPropagation(t *testing.T) {
	in := timeBindingFixture(t, 0, 100)
	before := jsonInput(t, in)
	m := evaluated(t, in)
	for _, recheck := range []bool{false, true} {
		count := traceCount(t, in, recheck, m)
		for i := 1; i <= count; i++ {
			func() {
				defer func() {
					if recover() != "trace panic" {
						t.Fatal("panic not propagated", i)
					}
				}()
				ctx := &traceTimeBindingContext{at: i, panicAt: true}
				if recheck {
					_ = Recheck(ctx, m, in)
				} else {
					got, e := Evaluate(ctx, in)
					if got != nil || e != nil {
						t.Fatal("panic returned result")
					}
				}
			}()
		}
	}
	if !bytes.Equal(before, jsonInput(t, in)) || !m.valid {
		t.Fatal("panic changed caller/prior")
	}
	// Nonreturned private owners are not directly observed: early-cleanup AST evidence is separate.
}
