package readiness

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	nodeauthority "talenro.local/platform/internal/nodecontrol/authority"
)

type fakeProbe struct {
	name string
	ping func(context.Context) error
}

type typedNilProbe struct{}

type fakeAuthorityReadinessChecker struct {
	readiness nodeauthority.Readiness
	err       error
}

func (checker *fakeAuthorityReadinessChecker) CheckReady(context.Context) (nodeauthority.Readiness, error) {
	return checker.readiness, checker.err
}

func (*typedNilProbe) Name() string { panic("typed nil probe name called") }

func (*typedNilProbe) Ping(context.Context) error { panic("typed nil probe ping called") }

func (p fakeProbe) Name() string { return p.name }

func (p fakeProbe) Ping(ctx context.Context) error {
	if p.ping == nil {
		return nil
	}

	return p.ping(ctx)
}

func TestCheckReportsEveryProbeWithoutLeakingErrors(t *testing.T) {
	checker := New(50*time.Millisecond,
		fakeProbe{name: "postgres"},
		fakeProbe{name: "redis", ping: func(context.Context) error {
			return errors.New("redis://user:secret@example.internal/account/device/node")
		}},
	)

	ready, checks := checker.Check(context.Background())

	if !ready {
		t.Fatal("degraded Redis must remain ready")
	}
	want := map[string]string{"postgres": "up", "redis": "degraded"}
	if !reflect.DeepEqual(checks, want) {
		t.Fatalf("unexpected checks: got %#v, want %#v", checks, want)
	}
}

func TestCheckReturnsReadyWhenAllProbesSucceed(t *testing.T) {
	checker := New(time.Second,
		fakeProbe{name: "postgres"},
		fakeProbe{name: "redis"},
		fakeProbe{name: "nats"},
		fakeProbe{name: "authority"},
	)

	ready, checks := checker.Check(context.Background())

	if !ready {
		t.Fatal("expected ready")
	}
	want := map[string]string{"postgres": "up", "redis": "up", "nats": "up", "authority": "up"}
	if !reflect.DeepEqual(checks, want) {
		t.Fatalf("unexpected checks: got %#v, want %#v", checks, want)
	}
}

func TestAuthorityProbeHasFixedNameAndValueFreeFailure(t *testing.T) {
	checker := &fakeAuthorityReadinessChecker{readiness: nodeauthority.Readiness{Ready: true, Reason: nodeauthority.ReadinessReady}}
	probe, err := newAuthorityProbe(checker)
	if err != nil {
		t.Fatal(err)
	}
	if probe.Name() != "authority" {
		t.Fatalf("Name = %q, want authority", probe.Name())
	}
	if err := probe.Ping(t.Context()); err != nil {
		t.Fatalf("ready Ping error = %v", err)
	}

	checker.readiness = nodeauthority.Readiness{
		ProviderHead: nodeauthority.Head{Epoch: 9223372036854775807},
		DatabaseHead: nodeauthority.DatabaseHead{Epoch: 123456789},
		Reason:       nodeauthority.ReadinessDatabaseBehindProvider,
	}
	if err := probe.Ping(t.Context()); err != nodeauthority.ErrAuthorityUnavailable {
		t.Fatalf("unready Ping error = %v, want fixed ErrAuthorityUnavailable", err)
	}
	checker.err = errors.New("postgres://private-host/account/device/secret")
	if err := probe.Ping(t.Context()); err != nodeauthority.ErrAuthorityUnavailable {
		t.Fatalf("dependency Ping error = %v, want fixed ErrAuthorityUnavailable", err)
	}
}

func TestCheckRunsProbesConcurrently(t *testing.T) {
	started := make(chan string, 2)
	release := make(chan struct{})
	probe := func(name string) fakeProbe {
		return fakeProbe{name: name, ping: func(context.Context) error {
			started <- name
			<-release
			return nil
		}}
	}
	checker := New(time.Second, probe("postgres"), probe("redis"))
	type result struct {
		ready  bool
		checks map[string]string
	}
	resultCh := make(chan result, 1)
	go func() {
		ready, checks := checker.Check(context.Background())
		resultCh <- result{ready: ready, checks: checks}
	}()

	for range 2 {
		select {
		case <-started:
		case <-time.After(time.Second):
			close(release)
			t.Fatal("probes did not start concurrently")
		}
	}
	close(release)

	select {
	case got := <-resultCh:
		if !got.ready {
			t.Fatalf("expected ready, got checks %#v", got.checks)
		}
	case <-time.After(time.Second):
		t.Fatal("check did not finish after probes were released")
	}
}

func TestCheckUsesShortProbeBudgetAndReleasesDeadlineOnlyWorkers(t *testing.T) {
	const timeout = 500 * time.Millisecond
	done := make(chan string, 2)
	blocked := func(name string) fakeProbe {
		return fakeProbe{name: name, ping: func(ctx context.Context) error {
			<-ctx.Done()
			done <- name
			return ctx.Err()
		}}
	}
	checker := New(timeout, blocked("postgres"), blocked("redis"))

	startedAt := time.Now()
	ready, checks := checker.Check(context.Background())
	elapsed := time.Since(startedAt)

	if ready {
		t.Fatal("expected unavailable")
	}
	want := map[string]string{"postgres": "down", "redis": "degraded"}
	if !reflect.DeepEqual(checks, want) {
		t.Fatalf("unexpected checks: got %#v, want %#v", checks, want)
	}
	if elapsed < timeout/2 {
		t.Fatalf("check returned before the probe budget elapsed: elapsed %v, timeout %v", elapsed, timeout)
	}
	if elapsed >= 3*timeout/4 {
		t.Fatalf("check did not preserve aggregation margin: elapsed %v, timeout %v", elapsed, timeout)
	}
	for range 2 {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("a probe worker remained blocked after timeout")
		}
	}
}

func TestCheckPreservesProbeAssociationsWhenCompletionOrderDiffers(t *testing.T) {
	releasePostgres := make(chan struct{})
	startedPostgres := make(chan struct{})
	checker := New(time.Second,
		fakeProbe{name: "postgres", ping: func(context.Context) error {
			close(startedPostgres)
			<-releasePostgres
			return nil
		}},
		fakeProbe{name: "redis", ping: func(context.Context) error {
			<-startedPostgres
			close(releasePostgres)
			return errors.New("sensitive dependency detail")
		}},
	)

	ready, checks := checker.Check(context.Background())

	if !ready {
		t.Fatal("degraded Redis must remain ready")
	}
	want := map[string]string{"postgres": "up", "redis": "degraded"}
	if !reflect.DeepEqual(checks, want) {
		t.Fatalf("completion order changed status association: got %#v, want %#v", checks, want)
	}
}

func TestCheckHonorsEarlierCallerCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	checker := New(time.Second, fakeProbe{name: "nats", ping: func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}})

	startedAt := time.Now()
	ready, checks := checker.Check(ctx)

	if ready || checks["nats"] != "degraded" {
		t.Fatalf("unexpected canceled check result: ready=%v checks=%#v", ready, checks)
	}
	if elapsed := time.Since(startedAt); elapsed >= 100*time.Millisecond {
		t.Fatalf("caller cancellation was not prompt: %v", elapsed)
	}
}

func TestNewRejectsNilProbeBeforeStartingChecks(t *testing.T) {
	pingCalled := false
	requirePanic(t, "readiness: nil probe", func() {
		New(time.Second,
			fakeProbe{name: "postgres", ping: func(context.Context) error {
				pingCalled = true
				return nil
			}},
			nil,
		)
	})

	if pingCalled {
		t.Fatal("invalid construction started a probe check")
	}
}

func TestNewRejectsTypedNilProbeBeforeCallingIt(t *testing.T) {
	var probe *typedNilProbe

	requirePanic(t, "readiness: nil probe", func() {
		New(time.Second, probe)
	})
}

func TestNewRejectsNamesOutsidePublicComponentAllowlist(t *testing.T) {
	for _, name := range []string{
		"redis://user:secret@example.internal/account/device/node",
		"10.0.0.1",
		"tenant-device-123",
	} {
		t.Run(name, func(t *testing.T) {
			pingCalled := false
			requirePanic(t, "readiness: unsafe probe name", func() {
				New(time.Second, fakeProbe{name: name, ping: func(context.Context) error {
					pingCalled = true
					return nil
				}})
			})

			if pingCalled {
				t.Fatal("invalid construction started a probe check")
			}
		})
	}
}

func TestNewRejectsDuplicateNamesBeforeOpposingChecksCanFinishOutOfOrder(t *testing.T) {
	pingCalls := 0
	releaseFirst := make(chan struct{})
	first := fakeProbe{name: "redis", ping: func(context.Context) error {
		pingCalls++
		<-releaseFirst
		return nil
	}}
	second := fakeProbe{name: "redis", ping: func(context.Context) error {
		pingCalls++
		close(releaseFirst)
		return errors.New("sensitive dependency detail")
	}}

	requirePanic(t, "readiness: duplicate probe name", func() {
		New(time.Second, first, second)
	})

	if pingCalls != 0 {
		t.Fatalf("invalid construction started %d probe checks", pingCalls)
	}
}

func TestNewAcceptsAllReviewedPublicComponentNames(t *testing.T) {
	checker := New(time.Second,
		fakeProbe{name: "postgres"},
		fakeProbe{name: "redis"},
		fakeProbe{name: "nats"},
		fakeProbe{name: "authority"},
	)

	ready, checks := checker.Check(context.Background())

	if !ready {
		t.Fatal("expected ready")
	}
	want := map[string]string{"postgres": "up", "redis": "up", "nats": "up", "authority": "up"}
	if !reflect.DeepEqual(checks, want) {
		t.Fatalf("unexpected checks: got %#v, want %#v", checks, want)
	}
}

func requirePanic(t *testing.T, want string, action func()) {
	t.Helper()

	defer func() {
		got := recover()
		if got == nil {
			t.Fatalf("expected panic %q", want)
		}
		if got != want {
			t.Fatalf("unexpected panic: got %q, want %q", got, want)
		}
	}()

	action()
}
