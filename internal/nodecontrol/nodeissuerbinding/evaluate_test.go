package nodeissuerbinding

import (
	"context"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"testing"
	"time"
)

func TestIssuerBindingEvaluate(t *testing.T) {
	var api func(context.Context, Input) (*Model, error) = Evaluate
	for _, name := range []string{"base", "leaf-first", "ca-first", "history-0", "history-32", "serial-one-high", "serial-twenty", "ca-renewal", "other-ca-same-serial"} {
		in := issuerBindingFixture(t, name)
		lower := assertIssuerBindingLower(t, in)
		before := issuerBindingJSONHash(t, in)
		m, e := api(context.Background(), in)
		if e != nil || m == nil {
			t.Fatal(name, e)
		}
		s := m.Summary()
		want := issuerBindingExpected(t, name)
		if s.Proof != SelectedIssuerSerialReferencesConditionalTimeBound || s.Profile != selectedIssuerSerialProfile || !reflect.DeepEqual(s.Evidence, lower) || s.Proposed != (Projection{Leaf: lower.Proposed, IssuerID: want.Issuer, Serial: want.Serial}) || s.ObservedAt != lower.ObservedAt || s.ValidUntil != lower.ValidUntil || s.Obligations != ([5]Obligation{StoreIssuerMappingAndDeploymentProvenance, KeyHandleAndPrivateKeyPossession, IdentityEpochLineageHistoryAndRowStatus, MainStateAndObservationProvenance, FullIdentityCompositionCommitAndRuntimeAuthorization}) {
			t.Fatal("full conditional result", name)
		}
		copyModel := *m
		copyModel.Clear()
		if !reflect.DeepEqual(m.Summary(), Summary{}) || !reflect.DeepEqual(copyModel.Summary(), Summary{}) || issuerBindingJSONHash(t, in) != before || !reflect.DeepEqual(s.Evidence, lower) {
			t.Fatal("lifecycle", name)
		}
	}
}
func TestIssuerBindingTemporalInheritance(t *testing.T) {
	in := issuerBindingFixture(t, "base")
	a, e := Evaluate(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer a.Clear()
	old := a.Summary()
	in.Time.Time.Time.Now = in.Time.Time.Time.Now.Add(time.Nanosecond)
	in.Time.Policy.Policy.Time.Now = in.Time.Policy.Policy.Time.Now.Add(time.Nanosecond)
	lower := assertIssuerBindingLower(t, in)
	b, e := Evaluate(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer b.Clear()
	if b.Summary().ObservedAt != old.ObservedAt.Add(time.Nanosecond) || b.Summary().ValidUntil != lower.ValidUntil || !reflect.DeepEqual(b.Summary().Evidence, lower) || a.Summary().ObservedAt != old.ObservedAt {
		t.Fatal("fresh observation")
	}
}
func TestIssuerBindingErrorPriority(t *testing.T) {
	in := issuerBindingFixture(t, "leaf-end-exact")
	lower, err := nodeIssuerBindingLowerError(in)
	if lower || err == nil {
		t.Fatal("expired control")
	}
	in.Time.Policy.Binding.Main.Previous.Identity.IssuerID[0] ^= 1
	in.Time.Policy.Binding.Main.Proposed.Identity.IssuerID[0] ^= 1
	m, e := Evaluate(context.Background(), in)
	if m != nil || e != err {
		t.Fatal("lower error priority", e, err)
	}
	if m, e = Evaluate(nil, in); m != nil || e != ErrInvalidInput {
		t.Fatal("nil context", e)
	}
}
func TestIssuerBindingUncoveredReferences(t *testing.T) {
	for _, mutate := range []func(*statecontract.IdentitySection){
		func(i *statecontract.IdentitySection) { i.KeyHandle[0] ^= 1 },
		func(i *statecontract.IdentitySection) { i.History.Payload[0] ^= 1 },
		func(i *statecontract.IdentitySection) { i.LineageUntil++ },
		func(i *statecontract.IdentitySection) { i.IdentityEpoch++ },
	} {
		in := issuerBindingFixture(t, "base")
		mutate(&in.Time.Policy.Binding.Main.Previous.Identity)
		mutate(&in.Time.Policy.Binding.Main.Proposed.Identity)
		assertIssuerBindingLower(t, in)
		m, e := Evaluate(context.Background(), in)
		if e != nil {
			t.Fatal(e)
		}
		if len(m.Summary().Obligations) != 5 {
			t.Fatal("responsibilities removed")
		}
		m.Clear()
	}
}
func TestIssuerBindingCancellationAndPanic(t *testing.T) {
	in := issuerBindingFixture(t, "base")
	before := issuerBindingJSONHash(t, in)
	probe := &ownershipFailureContext{Context: context.Background()}
	m, e := Evaluate(probe, in)
	if e != nil {
		t.Fatal(e)
	}
	m.Clear()
	for _, failure := range []error{context.Canceled, context.DeadlineExceeded} {
		for stop := 1; stop <= probe.calls; stop++ {
			ctx := &ownershipFailureContext{Context: context.Background(), stop: stop, failure: failure}
			failed, err := Evaluate(ctx, in)
			if failed != nil || err != failure || issuerBindingJSONHash(t, in) != before {
				t.Fatal("cancellation", stop, err)
			}
		}
	}
	for stop := 1; stop <= probe.calls; stop++ {
		func() {
			defer func() {
				if recover() != "ownership test panic" || issuerBindingJSONHash(t, in) != before {
					t.Fatal("panic", stop)
				}
			}()
			Evaluate(&ownershipFailureContext{Context: context.Background(), stop: stop, panicAt: true}, in)
		}()
	}
	t.Logf("full evaluation cancellation/deadline/panic checkpoints: %d", probe.calls)
}
