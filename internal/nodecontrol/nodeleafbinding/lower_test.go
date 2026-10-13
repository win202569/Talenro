package nodeleafbinding

import (
	"context"
	"errors"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/nodecerttime"
	"talenro.local/platform/internal/nodecontrol/statetimebinding"
	"testing"
	"time"
)

func TestLeafBindingLowerErrors(t *testing.T) {
	for _, name := range []string{"bad-signature", "leaf-end-exact", "bad-nonce", "bad-metadata", "rollback", "missing-observation", "policy-failure"} {
		t.Run(name, func(t *testing.T) {
			in := leafBindingFixture(t, name)
			m, want := nodecerttime.Evaluate(context.Background(), in)
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
func TestLeafBindingLowerOwnership(t *testing.T) {
	in := leafBindingFixture(t, "leaf-first")
	expected := assertLeafBindingLower(t, in)
	if !expected.ValidUntil.Before(expected.Time.ValidUntil) {
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
	if e = validateEvidence(expected); e != nil {
		t.Fatal(e)
	}
	for _, mutate := range []func(*nodecerttime.Summary){
		func(s *nodecerttime.Summary) { s.Proof = 0 },
		func(s *nodecerttime.Summary) { s.Proof = 99 },
		func(s *nodecerttime.Summary) { s.ObservedAt = time.Time{} },
		func(s *nodecerttime.Summary) { s.ValidUntil = time.Time{} },
		func(s *nodecerttime.Summary) { s.ValidUntil = s.ObservedAt },
		func(s *nodecerttime.Summary) { s.ValidUntil = s.ObservedAt.Add(-time.Nanosecond) },
		func(s *nodecerttime.Summary) { s.ObservedAt = s.ObservedAt.Add(time.Nanosecond) },
		func(s *nodecerttime.Summary) {
			s.ObservedAt = s.ObservedAt.In(time.FixedZone("same-instant", 3600))
			if !s.ObservedAt.Equal(s.Time.Policy.Time.Now) || s.ObservedAt == s.Time.Policy.Time.Now {
				t.Fatal("same-instant different-location control")
			}
		},
		func(s *nodecerttime.Summary) { s.ValidUntil = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) },
		func(s *nodecerttime.Summary) { s.ObservedAt = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC) },
	} {
		bad := expected
		mutate(&bad)
		if e = validateEvidence(bad); e != ErrInvalidEvidence {
			t.Fatal(e)
		}
	}
}
func TestLeafBindingProjectionControls(t *testing.T) {
	in := leafBindingFixture(t, "base")
	assertLeafBindingLower(t, in)
	for _, mutate := range []func(*Input){
		func(i *Input) {
			i.Time.Policy.Binding.Main.Previous.Identity.PublicKey[0] ^= 1
			i.Time.Policy.Binding.Main.Proposed.Identity.PublicKey[0] ^= 1
		},
		func(i *Input) {
			i.Time.Policy.Binding.Main.Previous.Identity.Window.Issued++
			i.Time.Policy.Binding.Main.Proposed.Identity.Window.Issued++
		},
		func(i *Input) {
			i.Time.Policy.Binding.Main.Previous.Identity.Window.Until--
			i.Time.Policy.Binding.Main.Proposed.Identity.Window.Until--
		},
	} {
		trial := leafBindingFixture(t, "base")
		mutate(&trial)
		assertLeafBindingLower(t, trial)
	}
}
