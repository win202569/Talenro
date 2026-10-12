package nodecerttime

import (
	"context"
	"testing"
	"time"
)

func certTimeEvaluateError(t testing.TB, in Input, want error) {
	t.Helper()
	m, e := Evaluate(context.Background(), in)
	if m != nil {
		defer m.Clear()
	}
	if e != want || (e != nil && m != nil) || (e == nil && m == nil) {
		t.Fatal("Evaluate result", e, want, m)
	}
}
func TestCertTimeWindowEdges(t *testing.T) {
	// Unit window controls are separate from real signed, public Evaluate controls.
	start := time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	for _, tc := range []struct {
		now  time.Time
		want error
	}{{start.Add(-time.Nanosecond), ErrNotYetValid}, {start, nil}, {start.Add(time.Nanosecond), nil}, {end.Add(-time.Nanosecond), nil}, {end, ErrExpired}, {end.Add(time.Nanosecond), ErrExpired}} {
		if e := certificateWindow(tc.now, start, end); e != tc.want {
			t.Fatal("unit half-open window", e, tc.want)
		}
	}
	for _, edge := range []string{"ca-start", "ca-end", "leaf-start", "leaf-end"} {
		for _, point := range []string{"before", "exact", "after"} {
			t.Run(edge+"-"+point, func(t *testing.T) {
				in := certTimeFixture(t, edge+"-"+point)
				assertCertTimeLowerControls(t, in)
				var want error
				if (edge == "ca-start" || edge == "leaf-start") && point == "before" {
					want = ErrNotYetValid
				}
				if (edge == "ca-end" || edge == "leaf-end") && point != "before" {
					want = ErrExpired
				}
				certTimeEvaluateError(t, in, want)
			})
		}
	}
}
func TestCertTimeYearOneEndpoints(t *testing.T) {
	zero := time.Time{}
	end := time.Date(1, 1, 2, 0, 0, 0, 0, time.UTC)
	now := zero.Add(time.Nanosecond)
	if e := certificateWindow(now, zero, end); e != nil {
		t.Fatal("valid zero certificate start", e)
	}
	for _, name := range []string{"ca-year-1", "leaf-year-1"} {
		in := certTimeFixture(t, name)
		assertCertTimeLowerControls(t, in)
		if name == "ca-year-1" {
			certTimeEvaluateError(t, in, nil)
		} else {
			certTimeEvaluateError(t, in, ErrExpired)
		}
	}
	// year-1 input is an issuer-positive control, not evidence of a year-1 time provider.
}
func TestCertTimeNecessaryTimes(t *testing.T) {
	s := assertCertTimeLowerControls(t, certTimeFixture(t, "base"))
	mutations := []func(*Summary){
		func(s *Summary) { s.Time.Policy.Time.Now = time.Time{} },
		func(s *Summary) { s.Time.ValidUntil = time.Time{} },
		func(s *Summary) { s.Time.Policy.Time.Now = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC) },
		func(s *Summary) { s.Time.ValidUntil = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) },
		func(s *Summary) { s.Time.Policy.Time.Now = s.Time.ValidUntil },
		func(s *Summary) { s.Time.Policy.Time.Now = s.Time.ValidUntil.Add(time.Nanosecond) },
		func(s *Summary) { s.Certificate.CA.NotBefore = s.Certificate.CA.NotAfter },
		func(s *Summary) { s.Certificate.CA.NotBefore = s.Certificate.CA.NotAfter.Add(time.Nanosecond) },
		func(s *Summary) { s.Certificate.Leaf.Leaf.NotBefore = s.Certificate.Leaf.Leaf.NotAfter },
		func(s *Summary) { s.Certificate.Leaf.Leaf.NotBefore = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC) },
		func(s *Summary) { s.Certificate.CA.NotAfter = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) },
	}
	for n, mutate := range mutations {
		bad := s
		mutate(&bad)
		if _, e := checkWindows(bad); e != ErrInvalidEvidence {
			t.Fatal("private necessary time", n, e)
		}
	}
}
func TestCertTimeDeadlineMinimum(t *testing.T) {
	now := time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name    string
		minutes int
	}{{"ca-first", 1}, {"leaf-first", 2}, {"time-first", 5}, {"ca-leaf-tie", 2}, {"ca-time-tie", 5}, {"leaf-time-tie", 5}, {"all-tie", 5}} {
		t.Run(tc.name, func(t *testing.T) {
			in := certTimeFixture(t, tc.name)
			s := assertCertTimeLowerControls(t, in)
			want := now.Add(time.Duration(tc.minutes) * time.Minute)
			got, e := checkWindows(s)
			if e != nil || got != want {
				t.Fatal("exact minimum helper", got, e, want)
			}
			m, e := Evaluate(context.Background(), in)
			if m != nil {
				defer m.Clear()
			}
			if e != nil || m == nil || m.Summary().ValidUntil != want {
				t.Fatal("exact minimum public", e)
			}
		})
	}
	certTimeEvaluateError(t, certTimeFixture(t, "ca-expired-leaf-future"), ErrExpired)
	certTimeEvaluateError(t, certTimeFixture(t, "ca-future-leaf-expired"), ErrNotYetValid)
}
func TestCertTimeNonContainingLeafWindow(t *testing.T) {
	in := certTimeFixture(t, "leaf-crosses-ca")
	s := assertCertTimeLowerControls(t, in)
	if !s.Certificate.Leaf.Leaf.NotBefore.Before(s.Certificate.CA.NotBefore) || !s.Certificate.CA.NotAfter.Before(s.Certificate.Leaf.Leaf.NotAfter) {
		t.Fatal("control does not cross")
	}
	m, e := Evaluate(context.Background(), in)
	if m != nil {
		defer m.Clear()
	}
	if e != nil || m == nil || m.Summary().ValidUntil != s.Certificate.CA.NotAfter {
		t.Fatal("no whole-window containment", e)
	}
}
func TestCertTimeFreshObservation(t *testing.T) {
	in := certTimeFixture(t, "base")
	a, e := Evaluate(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer a.Clear()
	old := a.Summary()
	in.Time.Time.Time.Now = in.Time.Time.Time.Now.Add(time.Nanosecond)
	in.Time.Policy.Policy.Time.Now = in.Time.Policy.Policy.Time.Now.Add(time.Nanosecond)
	assertCertTimeLowerControls(t, in)
	b, e := Evaluate(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer b.Clear()
	if b.Summary().ObservedAt != old.ObservedAt.Add(time.Nanosecond) || a.Summary().ObservedAt != old.ObservedAt {
		t.Fatal("fresh observation nanoseconds/owner independence")
	}
}
