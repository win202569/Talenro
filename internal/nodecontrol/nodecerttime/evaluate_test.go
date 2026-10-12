package nodecerttime

import (
	"context"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/nodeissuer"
	"testing"
	"time"
)

func TestCertTimeEvaluateGolden(t *testing.T) {
	in := certTimeFixture(t, "base")
	want := assertCertTimeLowerControls(t, in)
	want.Proof = NodeCertificateConditionalTimeBound
	want.ObservedAt = in.Time.Policy.Policy.Time.Now
	want.ValidUntil = time.Date(2020, 1, 1, 1, 5, 0, 0, time.UTC)
	m, e := Evaluate(context.Background(), in)
	if m != nil {
		defer m.Clear()
	}
	if e != nil || m == nil || !reflect.DeepEqual(m.Summary(), want) {
		t.Fatal("full conditional summary", e)
	}
}
func TestCertTimeObligationsRetained(t *testing.T) {
	in := certTimeFixture(t, "base")
	lower := assertCertTimeLowerControls(t, in)
	m, e := Evaluate(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer m.Clear()
	s := m.Summary()
	if !reflect.DeepEqual(s.Certificate, lower.Certificate) || !reflect.DeepEqual(s.Time, lower.Time) {
		t.Fatal("all nested obligations retained")
	}
	if len(s.Certificate.CA.Pending) != 6 || len(s.Certificate.Leaf.Leaf.Pending) != 6 {
		t.Fatal("original pending arrays")
	}
}
func TestCertTimeUnboundIdentity(t *testing.T) {
	base := certTimeFixture(t, "base")
	for _, kind := range []string{"identity", "ca-self-signature"} {
		in := ownedTestClone(t, base)
		if kind == "identity" {
			in.Time.Policy.Binding.Main.Previous.Identity.Leaf.Payload[0] ^= 1
			in.Time.Policy.Binding.Main.Proposed.Identity.Leaf.Payload[0] ^= 1
		} else {
			in.Certificate.CA.DER[len(in.Certificate.CA.DER)-1] ^= 1
		}
		lower := assertCertTimeLowerControls(t, in)
		m, e := Evaluate(context.Background(), in)
		if e != nil {
			t.Fatal(kind, e)
		}
		defer m.Clear()
		s := m.Summary()
		if s.Proof != NodeCertificateConditionalTimeBound || !reflect.DeepEqual(s.Certificate.CA.Pending, lower.Certificate.CA.Pending) || !reflect.DeepEqual(s.Certificate.Leaf.Leaf.Pending, lower.Certificate.Leaf.Leaf.Pending) || !reflect.DeepEqual(s.Time.Obligations, lower.Time.Obligations) {
			t.Fatal("unbound provenance not discharged")
		}
	}
}
func TestCertTimeEvaluatePriority(t *testing.T) {
	in := certTimeFixture(t, "bad-signature")
	in.Certificate.CA.DeclaredSPKI = make([]byte, 8193)
	in.Certificate.Leaf.DER = nil
	certTimeEvaluateError(t, in, ErrLimitExceeded)
	in.Certificate.CA.DeclaredSPKI = make([]byte, 8192)
	certTimeEvaluateError(t, in, ErrInvalidInput)
	in = certTimeFixture(t, "bad-signature")
	in.Time.Time.MetadataEnvelope[0] = '!'
	c, want := nodeissuer.Evaluate(context.Background(), in.Certificate)
	if c != nil {
		defer c.Clear()
	}
	if want == nil {
		t.Fatal("bad issuer control")
	}
	certTimeEvaluateError(t, in, want)
	in = certTimeFixture(t, "ca-expired-leaf-future")
	in.Time = certTimeFixture(t, "other-node").Time
	assertCertTimeLowerControls(t, in)
	certTimeEvaluateError(t, in, ErrBindingMismatch)
	certTimeEvaluateError(t, certTimeFixture(t, "ca-expired-leaf-future"), ErrExpired)
	certTimeEvaluateError(t, certTimeFixture(t, "ca-future-leaf-expired"), ErrNotYetValid)
}
func TestCertTimeIndependentLifecycle(t *testing.T) {
	in := certTimeFixture(t, "base")
	original := certTimeJSONHash(t, in)
	a, e := Evaluate(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer a.Clear()
	b, e := Evaluate(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer b.Clear()
	s1, s2 := a.Summary(), a.Summary()
	s1.Time.Policy.Binding.Proposed.Target.Links[0].Rotation.Payload[0] ^= 1
	s1.Time.Policy.Binding.Proposed.Coverage[0].Ref.Payload[0] ^= 1
	if !reflect.DeepEqual(s2, a.Summary()) || !reflect.DeepEqual(s2, b.Summary()) {
		t.Fatal("summary independent backing")
	}
	wrapper := *a
	views := ownedTestViews(reflect.ValueOf(a.owned.input))
	views = append(views, ownedTestViews(reflect.ValueOf(a.owned.summary.Time))...)
	wrapper.Clear()
	assertOwnedTestZero(t, views)
	if !reflect.DeepEqual(a.Summary(), Summary{}) || !reflect.DeepEqual(wrapper.Summary(), Summary{}) || !reflect.DeepEqual(b.Summary(), s2) || original != certTimeJSONHash(t, in) {
		t.Fatal("shared clear/independent owner/input")
	}
	a.Clear()
	wrapper.Clear()
	var nilModel *Model
	nilModel.Clear()
	if !reflect.DeepEqual(nilModel.Summary(), Summary{}) {
		t.Fatal("nil model")
	}
}

type certTimeCountingContext struct {
	context.Context
	calls, at  int
	fail       error
	panicValue any
}

func (c *certTimeCountingContext) Err() error {
	c.calls++
	if c.at > 0 && c.calls >= c.at {
		if c.panicValue != nil {
			panic(c.panicValue)
		}
		return c.fail
	}
	return nil
}
func TestCertTimeCancellationAndPanic(t *testing.T) {
	in := certTimeFixture(t, "base")
	original := certTimeJSONHash(t, in)
	count := &certTimeCountingContext{Context: context.Background()}
	m, e := Evaluate(count, in)
	if e != nil {
		t.Fatal(e)
	}
	m.Clear()
	if count.calls < 10 {
		t.Fatal("control did not traverse stages")
	}
	t.Log("successful observable context callbacks", count.calls)
	token := &struct{ label string }{"original panic"}
	for at := 1; at <= count.calls; at++ {
		for _, failure := range []error{context.Canceled, context.DeadlineExceeded} {
			ctx := &certTimeCountingContext{Context: context.Background(), at: at, fail: failure}
			m, e := Evaluate(ctx, in)
			if m != nil || e != failure {
				t.Fatal("cancel/timeout priority", at, m, e, failure)
			}
		}
		func() {
			ctx := &certTimeCountingContext{Context: context.Background(), at: at, panicValue: token}
			defer func() {
				if got := recover(); got != token {
					t.Fatal("normal panic original propagation", at, got)
				}
			}()
			m, e := Evaluate(ctx, in)
			if m != nil {
				m.Clear()
			}
			t.Fatal("panic was swallowed", at, e)
		}()
		if original != certTimeJSONHash(t, in) {
			t.Fatal("original input changed", at)
		}
	}
	// Public failure owner is deliberately inaccessible. Reachable cleanup is tested
	// above; Task 5 AST checks its failure protection. This is not hard realtime proof.
}
