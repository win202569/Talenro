package statepolicybinding

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/statebinding"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
	"testing"
	"time"
)

func evaluated(t *testing.T, in Input) *Model {
	t.Helper()
	m, e := Evaluate(context.Background(), in)
	if e != nil || m == nil {
		t.Fatal(e)
	}
	t.Cleanup(m.Clear)
	return m
}
func historyRecheckInput(t *testing.T) Input {
	in := policyBindingFixture(t, 1, 1)
	extraHistoricalKey(t, &in, 3, false)
	for i, k := range in.Policy.Old.Keys {
		in.Policy.History = append(in.Policy.History, statepolicy.KeyHistoryObservation{Status: statepolicy.Available, ObservationID: d(byte(86 + i)), BaseStateID: in.Policy.Old.StateID, MetadataPayload: in.Binding.Proposed.Target.Metadata.Payload, KeyID: k.Material.KeyID, RetiringKnown: true})
	}
	return in
}
func TestPolicyBindingSummary(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		in := policyBindingFixture(t, n, n)
		m := evaluated(t, in)
		s := m.Summary()
		proof := checkBindingInput(t, in)
		if !reflect.DeepEqual(s.Binding, proof) {
			t.Fatal("nested obligations or proof changed")
		}
		if s.Status != ConditionalSelectedPolicyBound || s.Obligations != [3]Obligation{ExplicitObservationsAndBaseProvenance, TimeAttestationAndFreshness, OtherDomainPolicyCoverage} || !s.SelectedPolicyNoChange {
			t.Fatal("status/duties")
		}
		wantBase := PolicyBase{StateID: in.Policy.Old.StateID, CounterID: in.Policy.Old.CounterID, StateEpoch: in.Policy.Old.StateEpoch, DeploymentSourceID: in.Policy.Deployment.SourceID, DeploymentObservationID: in.Policy.Deployment.ObservationID}
		wantTime := TimeSnapshot{ProviderID: in.Policy.Time.ProviderID, ObservationID: in.Policy.Time.ObservationID, Now: in.Policy.Time.Now, OldFloor: in.Policy.Old.TimeFloor, ObservedFloor: in.Policy.Time.Floor, MainFloor: in.Binding.Main.Proposed.Scope.TimeFloor}
		if s.Base != wantBase || s.Time != wantTime || s.HistoryCount != len(in.Policy.History) || s.ValidUntil != fixtureTime(t, "2020-01-02T00:00:00Z") {
			t.Fatal("fixed summary fields")
		}
		if s.Previous.Metadata != in.Policy.Old.Metadata || s.Previous.Desired != in.Policy.Old.Desired || s.Previous.KeyCount != len(in.Policy.Old.Keys) || s.Previous.RevokedCount != len(in.Policy.Old.RevokedIDs) || s.Previous != s.Proposed {
			t.Fatal("policy projections")
		}
		if s.Proposed.Root.RootSet != selectedTail(in) || s.Proposed.Root.Version != uint64(n+1) || s.Proposed.Root.Sequence != in.Policy.Old.Roots[n].Mark.Sequence || s.Proposed.Root.Source.Payload != in.Policy.Old.Roots[n].Mark.Payload {
			t.Fatal("typed root identity")
		}
		if n == 0 {
			if s.Proposed.Root.Source.Kind != InitialAnchor || s.Proposed.Root.Source.Envelope != in.Binding.Proposed.Anchor.Envelope {
				t.Fatal("anchor source")
			}
		} else {
			l := in.Binding.Proposed.Target.Links[n-1]
			if s.Proposed.Root.Source.Kind != RootRotationProof || s.Proposed.Root.Source.Payload != l.Rotation.Payload || s.Proposed.Root.Source.Envelope != l.Rotation.Envelope || s.Proposed.Root.Source.Payload == s.Proposed.Root.RootSet.Payload {
				t.Fatal("rotation source")
			}
		}
	}
	for _, m := range []*Model{nil, {}, evaluated(t, policyBindingFixture(t, 0, 0))} {
		m.Clear()
		if !reflect.DeepEqual(m.Summary(), Summary{}) {
			t.Fatal("zero/cleared summary")
		}
	}
}
func TestPolicyBindingNoWholeStateChange(t *testing.T) {
	in := policyBindingFixture(t, 1, 1)
	in.Binding.Main.Proposed.Checkpoint.Sequence++
	checkBindingInput(t, in)
	m := evaluated(t, in)
	if !m.Summary().SelectedPolicyNoChange || reflect.DeepEqual(in.Binding.Main.Previous, in.Binding.Main.Proposed) {
		t.Fatal("selected no-change conflated with whole state")
	}
	for _, method := range []string{"State", "Commit", "Authorize", "Ready", "WholeStateNoChange"} {
		if _, ok := reflect.TypeOf(m).MethodByName(method); ok {
			t.Fatal("authority API", method)
		}
	}
}
func zeroMutableSlices(v reflect.Value) {
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
func TestPolicyBindingModelOwnership(t *testing.T) {
	in := historyRecheckInput(t)
	in.Binding.Previous = in.Binding.Proposed.Materials
	in.Policy.Batch.AnchorEnvelope = in.Binding.Proposed.Anchor.Raw
	for i, l := range in.Binding.Proposed.Target.Links {
		in.Policy.Batch.Rotations[i] = material(&in.Binding.Proposed, l.Rotation).Raw
	}
	in.Policy.Batch.ActiveRootEnvelope = in.Policy.Batch.Rotations[len(in.Policy.Batch.Rotations)-1]
	in.Policy.Batch.MetadataEnvelope = material(&in.Binding.Proposed, in.Binding.Proposed.Target.Metadata).Raw
	in.Policy.Batch.DesiredEnvelope = material(&in.Binding.Proposed, in.Binding.Proposed.Target.Desired).Raw
	m := evaluated(t, in)
	before := jsonInput(t, m.input)
	summary := m.Summary()
	zeroMutableSlices(reflect.ValueOf(&in).Elem())
	if !bytes.Equal(before, jsonInput(t, m.input)) || !reflect.DeepEqual(m.Summary(), summary) {
		t.Fatal("caller mutation entered owner")
	}
	detached := m.Summary()
	zeroMutableSlices(reflect.ValueOf(&detached).Elem())
	if !reflect.DeepEqual(m.Summary(), summary) {
		t.Fatal("summary alias")
	}
	m.Clear()
	if !reflect.DeepEqual(summary.Binding.Status, statebinding.ConditionalProjectionBound) {
		t.Fatal("detached summary destroyed")
	}
}
func collectHeldSlices(v reflect.Value, out *[]reflect.Value) {
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			collectHeldSlices(v.Elem(), out)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).CanInterface() {
				collectHeldSlices(v.Field(i), out)
			}
		}
	case reflect.Slice:
		if v.Len() > 0 {
			*out = append(*out, reflect.ValueOf(v.Interface()))
		}
		for i := 0; i < v.Len(); i++ {
			collectHeldSlices(v.Index(i), out)
		}
	}
}
func TestPolicyBindingModelClear(t *testing.T) {
	in := historyRecheckInput(t)
	m := evaluated(t, in)
	caller := jsonInput(t, in)
	detached := m.Summary()
	detachedBefore := m.Summary()
	raw := m.input.Policy.Batch.ActiveRootEnvelope
	bindingRaw := m.input.Binding.Proposed.Materials.Materials[0].Raw
	rotations := m.input.Policy.Batch.Rotations
	rotationRaw := rotations[0]
	var allHeld []reflect.Value
	collectHeldSlices(reflect.ValueOf(m.input), &allHeld)
	collectHeldSlices(reflect.ValueOf(m.summary), &allHeld)
	keys := m.input.Policy.Old.Keys
	roots := m.input.Policy.Old.Roots
	public := roots[0].Root.Keys
	history := m.input.Policy.History
	links := m.summary.Binding.Proposed.Target.Links
	coverage := m.summary.Binding.Proposed.Coverage
	mainKeys := m.input.Binding.Main.Proposed.Root.KeyIDs
	m.Clear()
	m.Clear()
	for _, b := range [][]byte{raw, bindingRaw, rotationRaw} {
		for _, x := range b {
			if x != 0 {
				t.Fatal("reachable raw not cleared")
			}
		}
	}
	for _, held := range []any{keys, roots, public, history, links, coverage, mainKeys, rotations} {
		v := reflect.ValueOf(held)
		for i := 0; i < v.Len(); i++ {
			if !v.Index(i).IsZero() {
				t.Fatal("reachable collection not cleared")
			}
		}
	}
	for _, held := range allHeld {
		for i := 0; i < held.Len(); i++ {
			if !held.Index(i).IsZero() {
				t.Fatal("owned backing store not cleared", held.Type())
			}
		}
	}
	if !reflect.DeepEqual(m.input, Input{}) || !reflect.DeepEqual(m.summary, Summary{}) || m.valid || !bytes.Equal(caller, jsonInput(t, in)) || !reflect.DeepEqual(detached, detachedBefore) {
		t.Fatal("clear scope")
	}
}
func legalReplacement(t *testing.T, m *Model, in Input) {
	t.Helper()
	fresh := evaluated(t, in)
	if !fresh.valid {
		t.Fatal("replacement not independently valid")
	}
	if e := Recheck(context.Background(), m, in); !errors.Is(e, ErrBindingMismatch) {
		t.Fatal("identity change accepted", e)
	}
}
func TestPolicyBindingRecheck(t *testing.T) {
	in := historyRecheckInput(t)
	m := evaluated(t, in)
	equal := copyPolicyBindingInput(t, in)
	for _, table := range []*statematerials.Input{&equal.Binding.Previous, &equal.Binding.Proposed.Materials} {
		for i, j := 0, len(table.Materials)-1; i < j; i, j = i+1, j-1 {
			table.Materials[i], table.Materials[j] = table.Materials[j], table.Materials[i]
		}
		for i, j := 0, len(table.Expected)-1; i < j; i, j = i+1, j-1 {
			table.Expected[i], table.Expected[j] = table.Expected[j], table.Expected[i]
		}
	}
	equal.Policy.Time.Now = equal.Policy.Time.Now.In(time.FixedZone("same", 3600))
	equal.Policy.Time.Floor = equal.Policy.Time.Floor.In(time.FixedZone("same", -3600))
	equal.Policy.Old.TimeFloor = equal.Policy.Old.TimeFloor.In(time.FixedZone("same", 7200))
	equal.Binding.Main.Proposed.Recovery.Local = []statecontract.LocalFaultBinding{}
	equal.Binding.Main.Previous.Recovery.Local = []statecontract.LocalFaultBinding{}
	if e := Recheck(context.Background(), m, equal); e != nil {
		t.Fatal("normalization equivalence", e)
	}
	for _, change := range []func(*Input){
		func(x *Input) { x.Policy.Deployment.ObservationID[0] ^= 1 }, func(x *Input) { x.Policy.Old.CounterID[0] ^= 1 }, func(x *Input) { x.Policy.Old.StateEpoch++ },
		func(x *Input) {
			x.Policy.Old.StateID[0] ^= 1
			for i := range x.Policy.History {
				x.Policy.History[i].BaseStateID = x.Policy.Old.StateID
			}
		},
		func(x *Input) { x.Policy.History[0].ObservationID[0] ^= 1 }, func(x *Input) { x.Policy.History[0].RetiringKnown = !x.Policy.History[0].RetiringKnown },
		func(x *Input) { x.Policy.History[0], x.Policy.History[1] = x.Policy.History[1], x.Policy.History[0] },
	} {
		x := copyPolicyBindingInput(t, in)
		change(&x)
		legalReplacement(t, m, x)
	}
	if e := Recheck(nil, nil, in); !errors.Is(e, ErrInvalidInput) {
		t.Fatal(e)
	}
	if e := Recheck(context.Background(), nil, in); !errors.Is(e, ErrClearedModel) {
		t.Fatal(e)
	}
	m.Clear()
	if e := Recheck(context.Background(), m, in); !errors.Is(e, ErrClearedModel) {
		t.Fatal(e)
	}
}
func TestPolicyBindingUnmappedRecheck(t *testing.T) {
	in := policyBindingFixture(t, 1, 1)
	m := evaluated(t, in)
	for _, change := range []func(*Input){
		func(x *Input) { x.Binding.Main.Proposed.Checkpoint.Sequence++ },
		func(x *Input) {
			x.Binding.Main.Previous.Metadata.History.Payload[0] ^= 1
			x.Binding.Main.Proposed.Metadata.History = x.Binding.Main.Previous.Metadata.History
		},
		func(x *Input) {
			x.Binding.Main.Previous.Metadata.RevokedSetDigest[0] ^= 1
			x.Binding.Main.Proposed.Metadata.RevokedSetDigest = x.Binding.Main.Previous.Metadata.RevokedSetDigest
		},
		func(x *Input) {
			x.Binding.Main.Previous.Identity.History.Payload[0] ^= 1
			x.Binding.Main.Proposed.Identity.History = x.Binding.Main.Previous.Identity.History
		},
		func(x *Input) {
			x.Binding.Main.Previous.ServerTrust.Agent.DeauthorizedSetDigest[0] ^= 1
			x.Binding.Main.Proposed.ServerTrust.Agent.DeauthorizedSetDigest = x.Binding.Main.Previous.ServerTrust.Agent.DeauthorizedSetDigest
		},
		func(x *Input) {
			x.Binding.Main.Previous.Scope.DeploymentID = d(91)
			x.Binding.Main.Proposed.Scope.DeploymentID = x.Binding.Main.Previous.Scope.DeploymentID
		},
	} {
		x := copyPolicyBindingInput(t, in)
		change(&x)
		legalReplacement(t, m, x)
	}
	x := copyPolicyBindingInput(t, in)
	unselected := object(t, frozen(t, "metadata.payload.json"))
	unselected["metadata_version"] = "9"
	unselected["authority_sequence"] = "11"
	addEnvelope(t, &x.Binding.Proposed, statematerials.Metadata, mint(t, "metadata", unselected, private(1)), statematerials.MaterialRef{}, tailRef(x.Binding.Proposed), statematerials.MaterialRef{})
	legalReplacement(t, m, x)
	// Changing an existing legal unselected raw/expectation also cannot reuse an earlier model.
	m2 := evaluated(t, x)
	unselected["metadata_version"] = "10"
	unselected["authority_sequence"] = "12"
	y := copyPolicyBindingInput(t, x)
	for _, r := range y.Binding.Proposed.Materials.Materials {
		if r.Ref.Kind == statematerials.Metadata && r.Ref != y.Binding.Proposed.Target.Metadata {
			removeSelected(&y.Binding.Proposed, r.Ref)
		}
	}
	addEnvelope(t, &y.Binding.Proposed, statematerials.Metadata, mint(t, "metadata", unselected, private(1)), statematerials.MaterialRef{}, tailRef(y.Binding.Proposed), statematerials.MaterialRef{})
	legalReplacement(t, m2, y)
}
func TestPolicyBindingFixedTime(t *testing.T) {
	in := policyBindingFixture(t, 1, 1)
	m := evaluated(t, in)
	for _, change := range []func(*Input){func(x *Input) { x.Policy.Time.Now = x.Policy.Time.Now.Add(time.Nanosecond) }, func(x *Input) { x.Policy.Time.Floor = x.Policy.Time.Floor.Add(time.Nanosecond) }, func(x *Input) { x.Policy.Time.ObservationID[0] ^= 1 }} {
		x := copyPolicyBindingInput(t, in)
		change(&x)
		legalReplacement(t, m, x)
	}
}

type tracePolicyContext struct {
	calls, at int
	failure   error
	panicAt   bool
}

func (c *tracePolicyContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c *tracePolicyContext) Done() <-chan struct{}       { return nil }
func (c *tracePolicyContext) Value(any) any               { return nil }
func (c *tracePolicyContext) Err() error {
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
	ctx := &tracePolicyContext{}
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
func TestPolicyBindingCancellation(t *testing.T) {
	in := policyBindingFixture(t, 0, 0)
	before := jsonInput(t, in)
	m := evaluated(t, in)
	for _, e := range []error{context.Canceled, context.DeadlineExceeded} {
		for _, recheck := range []bool{false, true} {
			count := traceCount(t, in, recheck, m)
			for i := 1; i <= count; i++ {
				ctx := &tracePolicyContext{at: i, failure: e}
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
func TestPolicyBindingLateCancellation(t *testing.T) {
	in := policyBindingFixture(t, 1, 1)
	m := evaluated(t, in)
	for _, recheck := range []bool{false, true} {
		ctx := &tracePolicyContext{at: traceCount(t, in, recheck, m), failure: context.DeadlineExceeded}
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
func TestPolicyBindingPanicCleanup(t *testing.T) {
	in := policyBindingFixture(t, 0, 0)
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
				ctx := &tracePolicyContext{at: i, panicAt: true}
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
