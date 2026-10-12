package nodecerttime

import (
	"context"
	"errors"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/nodeissuer"
	"talenro.local/platform/internal/nodecontrol/statetimebinding"
	"testing"
)

func TestCertTimeLowerErrorIdentity(t *testing.T) {
	for _, name := range []string{"bad-signature", "bad-nonce", "bad-metadata", "rollback", "missing-observation", "policy-failure"} {
		t.Run(name, func(t *testing.T) {
			in := certTimeFixture(t, name)
			c, want := nodeissuer.Evaluate(context.Background(), in.Certificate)
			if c != nil {
				defer c.Clear()
			}
			if want == nil {
				tm, e := statetimebinding.Evaluate(context.Background(), in.Time)
				if tm != nil {
					defer tm.Clear()
				}
				want = e
			}
			if want == nil {
				t.Fatal("negative lower control succeeded")
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
					t.Fatal("typed original lower error", got, want)
				}
			} else if got != want {
				t.Fatal("sentinel identity", got, want)
			}
			if o.valid || o.summary.Proof != None {
				t.Fatal("premature proof")
			}
		})
	}
}
func TestCertTimeStageError(t *testing.T) {
	original := errors.New("original")
	c, cancel := context.WithCancel(context.Background())
	cancel()
	d, stop := context.WithDeadline(context.Background(), certTimeFixture(t, "base").Time.Time.Time.Now)
	defer stop()
	for _, tc := range []struct {
		ctx       context.Context
		err, want error
	}{{nil, nil, ErrInvalidInput}, {nil, original, ErrInvalidInput}, {c, original, context.Canceled}, {d, original, context.DeadlineExceeded}, {context.Background(), original, original}, {context.Background(), nil, nil}} {
		if got := stageError(tc.ctx, tc.err); got != tc.want {
			t.Fatal(got, tc.want)
		}
	}
}
func TestCertTimeLowerProofs(t *testing.T) {
	in := certTimeFixture(t, "base")
	s := assertCertTimeLowerControls(t, in)
	if e := checkLowerProofs(s); e != nil {
		t.Fatal(e)
	}
	for _, mutate := range []func(*Summary){func(s *Summary) { s.Certificate.Proof = 0 }, func(s *Summary) { s.Certificate.Proof = 99 }, func(s *Summary) { s.Time.Status = 0 }, func(s *Summary) { s.Time.Status = 99 }} {
		bad := s
		mutate(&bad)
		if e := checkLowerProofs(bad); e != ErrInvalidEvidence {
			t.Fatal(e)
		}
	}
	o, e := ownInput(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer o.clear()
	if e = evaluateLower(context.Background(), o); e != nil || !reflect.DeepEqual(s, o.summary) {
		t.Fatal("real lower summary capture", e)
	}
	if o.valid || o.summary.Proof != None {
		t.Fatal("premature proof")
	}
}
