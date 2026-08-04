package readiness

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type fakeProbe struct {
	name string
	ping func(context.Context) error
}

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

	if ready {
		t.Fatal("expected unavailable")
	}
	want := map[string]string{"postgres": "ok", "redis": "unavailable"}
	if !reflect.DeepEqual(checks, want) {
		t.Fatalf("unexpected checks: got %#v, want %#v", checks, want)
	}
}

func TestCheckReturnsReadyWhenAllProbesSucceed(t *testing.T) {
	checker := New(time.Second,
		fakeProbe{name: "postgres"},
		fakeProbe{name: "redis"},
		fakeProbe{name: "nats"},
	)

	ready, checks := checker.Check(context.Background())

	if !ready {
		t.Fatal("expected ready")
	}
	want := map[string]string{"postgres": "ok", "redis": "ok", "nats": "ok"}
	if !reflect.DeepEqual(checks, want) {
		t.Fatalf("unexpected checks: got %#v, want %#v", checks, want)
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

func TestCheckUsesOneOverallTimeoutAndReleasesWorkers(t *testing.T) {
	const timeout = 30 * time.Millisecond
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
	want := map[string]string{"postgres": "unavailable", "redis": "unavailable"}
	if !reflect.DeepEqual(checks, want) {
		t.Fatalf("unexpected checks: got %#v, want %#v", checks, want)
	}
	if elapsed < timeout {
		t.Fatalf("check returned before timeout: elapsed %v, timeout %v", elapsed, timeout)
	}
	if elapsed >= 10*timeout {
		t.Fatalf("checks appear to have timed out serially: elapsed %v", elapsed)
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

	if ready {
		t.Fatal("expected unavailable")
	}
	want := map[string]string{"postgres": "ok", "redis": "unavailable"}
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

	if ready || checks["nats"] != "unavailable" {
		t.Fatalf("unexpected canceled check result: ready=%v checks=%#v", ready, checks)
	}
	if elapsed := time.Since(startedAt); elapsed >= 100*time.Millisecond {
		t.Fatalf("caller cancellation was not prompt: %v", elapsed)
	}
}
