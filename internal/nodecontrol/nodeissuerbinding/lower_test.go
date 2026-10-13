package nodeissuerbinding

import (
	"context"
	"errors"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/nodeleafbinding"
	"talenro.local/platform/internal/nodecontrol/statetimebinding"
	"testing"
	"time"
)

func TestIssuerBindingLowerErrors(t *testing.T) {
	for _, name := range []string{"bad-signature", "leaf-end-exact", "bad-nonce", "bad-metadata", "rollback", "missing-observation", "policy-failure"} {
		t.Run(name, func(t *testing.T) {
			in := issuerBindingFixture(t, name)
			m, want := nodeleafbinding.Evaluate(context.Background(), in)
			if m != nil {
				defer m.Clear()
			}
			if want == nil {
				t.Fatal("negative control succeeded")
			}
			o, e := ownInput(context.Background(), in)
			if e != nil {
				t.Fatal(e)
			}
			defer o.clear()
			got := evaluateLower(context.Background(), o)
			if name == "policy-failure" {
				var a, b *statetimebinding.PolicyFailure
				if !errors.As(got, &a) || !errors.As(want, &b) || a.Code != b.Code || a.Reason != b.Reason || errors.Unwrap(got) != errors.Unwrap(want) {
					t.Fatal("typed original error", got, want)
				}
			} else if got != want {
				t.Fatal("original sentinel", got, want)
			}
			if o.valid || o.summary.Proof != None {
				t.Fatal("premature publication")
			}
		})
	}
	original := errors.New("original")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	deadline, stop := context.WithDeadline(context.Background(), time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	defer stop()
	for _, c := range []struct {
		ctx       context.Context
		err, want error
	}{{nil, original, ErrInvalidInput}, {ctx, original, context.Canceled}, {deadline, original, context.DeadlineExceeded}, {context.Background(), original, original}, {context.Background(), nil, nil}} {
		if got := stageError(c.ctx, c.err); got != c.want {
			t.Fatal(got, c.want)
		}
	}
	if e := evaluateLower(context.Background(), nil); e != ErrInvalidInput {
		t.Fatal(e)
	}
}
func TestIssuerBindingLowerOwnership(t *testing.T) {
	in := issuerBindingFixture(t, "leaf-first")
	expected := assertIssuerBindingLower(t, in)
	if !expected.ValidUntil.Before(expected.Evidence.Time.ValidUntil) {
		t.Fatal("certificate earlier control")
	}
	o, e := ownInput(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer o.clear()
	if e = evaluateLower(context.Background(), o); e != nil || !reflect.DeepEqual(o.summary.Evidence, expected) {
		t.Fatal("protected full capture", e)
	}
	if o.valid || o.summary.Proof != None {
		t.Fatal("premature proof")
	}
	snapshot := ownEvidence(o.summary.Evidence)
	defer clearEvidence(&snapshot)
	o.clear()
	if !reflect.DeepEqual(snapshot, expected) {
		t.Fatal("lower or parent clear escaped")
	}
	if e = validateEvidence(in, expected); e != nil {
		t.Fatal(e)
	}
	for _, mutate := range []func(*nodeleafbinding.Summary){
		func(s *nodeleafbinding.Summary) { s.Proof = 0 },
		func(s *nodeleafbinding.Summary) { s.Proof = 99 },
		func(s *nodeleafbinding.Summary) { s.ObservedAt = time.Time{} },
		func(s *nodeleafbinding.Summary) { s.ValidUntil = time.Time{} },
		func(s *nodeleafbinding.Summary) { s.ValidUntil = s.ObservedAt },
		func(s *nodeleafbinding.Summary) { s.ValidUntil = s.ObservedAt.Add(-time.Nanosecond) },
		func(s *nodeleafbinding.Summary) { s.ObservedAt = s.ObservedAt.Add(time.Nanosecond) },
		func(s *nodeleafbinding.Summary) {
			s.ObservedAt = s.ObservedAt.In(time.FixedZone("same-instant", 3600))
			if !s.ObservedAt.Equal(s.Evidence.Time.Policy.Time.Now) || s.ObservedAt == s.Evidence.Time.Policy.Time.Now {
				t.Fatal("same-instant different-location control")
			}
		},
		func(s *nodeleafbinding.Summary) { s.ValidUntil = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) },
		func(s *nodeleafbinding.Summary) { s.ObservedAt = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC) },
	} {
		bad := expected
		mutate(&bad)
		if e = validateEvidence(in, bad); e != ErrInvalidEvidence {
			t.Fatal(e)
		}
	}
}

func TestIssuerBindingLowerConsistency(t *testing.T) {
	in := issuerBindingFixture(t, "base")
	expected := assertIssuerBindingLower(t, in)
	for _, mutate := range []func(*nodeleafbinding.Summary){
		func(s *nodeleafbinding.Summary) { s.Evidence.ObservedAt = s.Evidence.ObservedAt.Add(time.Nanosecond) },
		func(s *nodeleafbinding.Summary) { s.Evidence.ValidUntil = s.Evidence.ValidUntil.Add(time.Nanosecond) },
		func(s *nodeleafbinding.Summary) {
			s.Evidence.Time.Policy.Time.Now = s.Evidence.Time.Policy.Time.Now.Add(time.Nanosecond)
		},
		func(s *nodeleafbinding.Summary) { s.Proposed.NodeID += "x" },
		func(s *nodeleafbinding.Summary) { s.Proposed.Leaf[0] ^= 1 },
		func(s *nodeleafbinding.Summary) { s.Proposed.PublicKey[0] ^= 1 },
		func(s *nodeleafbinding.Summary) { s.Proposed.Window.Until-- },
		func(s *nodeleafbinding.Summary) { s.ValidUntil = s.ValidUntil.In(time.FixedZone("same", 3600)) },
	} {
		bad := ownEvidence(expected)
		mutate(&bad)
		if validateEvidence(in, bad) != ErrInvalidEvidence {
			t.Fatal("full lower consistency")
		}
		clearEvidence(&bad)
	}
}
