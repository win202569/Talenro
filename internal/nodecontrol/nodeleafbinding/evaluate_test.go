package nodeleafbinding

import (
	"context"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/nodecerttime"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"testing"
	"time"
)

func nodeLowerError(in Input) (bool, error) {
	m, e := nodecerttime.Evaluate(context.Background(), in)
	if m != nil {
		m.Clear()
	}
	return e == nil, e
}
func certTimeExpiredError() error { return nodecerttime.ErrExpired }
func TestLeafBindingEqualSKIValue(t *testing.T) {
	in := leafBindingFixture(t, "base")
	s := assertLeafBindingLower(t, in)
	if !reflect.DeepEqual(s.Certificate.Leaf.SubjectKeyIdentifier[:], s.Certificate.Leaf.Leaf.SPKIDERDigest[:]) {
		t.Fatal("equal SKI numerical control")
	}
	m := leafBindingResult(t, in, nil)
	defer m.Clear()
	if m.Summary().Obligations != ([5]Obligation{IssuerAndSerialReferenceMapping, KeyHandleAndPrivateKeyPossession, IdentityEpochLineageHistoryAndRowStatus, MainStateAndObservationProvenance, FullIdentityCompositionCommitAndRuntimeAuthorization}) {
		t.Fatal("responsibilities")
	}
}
func TestLeafBindingUncoveredReferences(t *testing.T) {
	for index, mutate := range []func(*statecontract.IdentitySection){
		func(i *statecontract.IdentitySection) { i.IssuerID[0] ^= 128 },
		func(i *statecontract.IdentitySection) { i.Serial[0] ^= 1 },
		func(i *statecontract.IdentitySection) { i.KeyHandle[0] ^= 1 },
		func(i *statecontract.IdentitySection) { i.History.Payload[0] ^= 1 },
		func(i *statecontract.IdentitySection) { i.LineageUntil++ },
		func(i *statecontract.IdentitySection) { i.IdentityEpoch++ },
	} {
		in := leafBindingFixture(t, "base")
		mutate(&in.Time.Policy.Binding.Main.Previous.Identity)
		mutate(&in.Time.Policy.Binding.Main.Proposed.Identity)
		t.Logf("uncovered reference mutation index: %d", index)
		assertLeafBindingLower(t, in)
		m := leafBindingResult(t, in, nil)
		if len(m.Summary().Obligations) != 5 || m.Summary().Proposed != leafBindingExpected(t, "base").Projection {
			t.Fatal("uncovered reference changed guarantee")
		}
		m.Clear()
	}
}
func TestLeafBindingTemporalInheritance(t *testing.T) {
	for _, name := range []string{"base", "leaf-first", "ca-first", "history-0", "history-32"} {
		in := leafBindingFixture(t, name)
		s := assertLeafBindingLower(t, in)
		m := leafBindingResult(t, in, nil)
		defer m.Clear()
		if m.Summary().ObservedAt != s.ObservedAt || m.Summary().ValidUntil != s.ValidUntil || !reflect.DeepEqual(m.Summary().Evidence, s) {
			t.Fatal("exact inheritance")
		}
	}
	in := leafBindingFixture(t, "base")
	a := leafBindingResult(t, in, nil)
	defer a.Clear()
	old := a.Summary()
	in.Time.Time.Time.Now = in.Time.Time.Time.Now.Add(time.Nanosecond)
	in.Time.Policy.Policy.Time.Now = in.Time.Policy.Policy.Time.Now.Add(time.Nanosecond)
	assertLeafBindingLower(t, in)
	b := leafBindingResult(t, in, nil)
	defer b.Clear()
	if b.Summary().ObservedAt != old.ObservedAt.Add(time.Nanosecond) || a.Summary().ObservedAt != old.ObservedAt {
		t.Fatal("fresh independent observation")
	}
}
func TestLeafBindingErrorPriority(t *testing.T) {
	in := leafBindingFixture(t, "leaf-end-exact")
	in.Time.Policy.Binding.Main.Previous.Identity.PublicKey[0] ^= 1
	in.Time.Policy.Binding.Main.Proposed.Identity.PublicKey[0] ^= 1
	leafBindingResult(t, in, nodecerttime.ErrExpired)
	if m, e := Evaluate(nil, in); m != nil || e != ErrInvalidInput {
		t.Fatal("nil context", e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if m, e := Evaluate(ctx, Input{}); m != nil || e != context.Canceled {
		t.Fatal("pre-cancel", e)
	}
}

type evaluationContext struct {
	context.Context
	calls, stop int
	failure     error
	panicValue  any
}

func (c *evaluationContext) Err() error {
	c.calls++
	if c.stop > 0 && c.calls >= c.stop {
		if c.panicValue != nil {
			panic(c.panicValue)
		}
		return c.failure
	}
	return nil
}
func TestLeafBindingCancellationAndPanic(t *testing.T) {
	in := leafBindingFixture(t, "base")
	before := leafBindingJSONHash(t, in)
	probe := &evaluationContext{Context: context.Background()}
	m, e := Evaluate(probe, in)
	if e != nil || m == nil {
		t.Fatal(e)
	}
	m.Clear()
	t.Logf("evaluation cancellation/panic observation points: %d", probe.calls)
	for _, failure := range []error{context.Canceled, context.DeadlineExceeded} {
		for point := 1; point <= probe.calls; point++ {
			ctx := &evaluationContext{Context: context.Background(), stop: point, failure: failure}
			m, e := Evaluate(ctx, in)
			if m != nil || e != failure {
				if m != nil {
					m.Clear()
				}
				t.Fatal("stage error", point, e, failure)
			}
			if leafBindingJSONHash(t, in) != before {
				t.Fatal("caller changed", point)
			}
		}
	}
	marker := new(int)
	for point := 1; point <= probe.calls; point++ {
		func() {
			defer func() {
				if r := recover(); r != marker {
					t.Fatal("panic propagation", point, r)
				}
			}()
			ctx := &evaluationContext{Context: context.Background(), stop: point, panicValue: marker}
			m, e := Evaluate(ctx, in)
			if m != nil {
				m.Clear()
			}
			t.Fatal("panic missing", point, e)
		}()
		if leafBindingJSONHash(t, in) != before {
			t.Fatal("panic changed caller", point)
		}
	}
}
