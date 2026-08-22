package readiness

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"talenro.local/platform/internal/store"
)

func TestPolicyPostgresFailureIsImmediatelyDown(t *testing.T) {
	postgres := &task18ScriptedProbe{name: "postgres", failures: []bool{true}}
	redis := &task18ScriptedProbe{name: "redis"}
	checker := task18PolicyChecker(t, postgres, redis, task18OutboxProbe{})

	ready, checks := checker.Check(context.Background())
	if ready || checks["postgres"] != "down" {
		t.Fatalf("postgres failure: ready=%v checks=%v", ready, checks)
	}
}

func TestPolicyDeadlineOnlyPostgresIsDownBeforeOverallDeadline(t *testing.T) {
	const timeout = 500 * time.Millisecond
	postgres := fakeProbe{name: "postgres", ping: func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}}
	checker, err := NewWithPolicy(timeout, DefaultPolicy(), postgres, fakeProbe{name: "redis"}, task18OutboxProbe{})
	if err != nil {
		t.Fatal(err)
	}

	startedAt := time.Now()
	ready, checks := checker.Check(context.Background())
	if ready || checks["postgres"] != "down" {
		t.Fatalf("deadline-only postgres failure: ready=%v checks=%v", ready, checks)
	}
	if elapsed := time.Since(startedAt); elapsed >= 3*timeout/4 {
		t.Fatalf("deadline-only postgres consumed aggregation margin: elapsed %v, timeout %v", elapsed, timeout)
	}
}

func TestPolicyRedisFailsDegradedThenDownAndNeedsTwoSuccessesToRecover(t *testing.T) {
	postgres := &task18ScriptedProbe{name: "postgres"}
	redis := &task18ScriptedProbe{name: "redis", failures: []bool{true, true, true, false, false}}
	checker := task18PolicyChecker(t, postgres, redis, task18OutboxProbe{})
	wantStatus := []string{"degraded", "degraded", "down", "degraded", "up"}
	wantReady := []bool{true, true, false, true, true}
	for index := range wantStatus {
		ready, checks := checker.Check(context.Background())
		if ready != wantReady[index] || checks["redis"] != wantStatus[index] {
			t.Fatalf("check %d: ready=%v redis=%q, want %v/%q", index+1, ready, checks["redis"], wantReady[index], wantStatus[index])
		}
	}
}

func TestPolicyDeadlineOnlyRedisRemainsDegradedTwiceThenDown(t *testing.T) {
	const timeout = 500 * time.Millisecond
	redis := fakeProbe{name: "redis", ping: func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}}
	checker, err := NewWithPolicy(timeout, DefaultPolicy(), fakeProbe{name: "postgres"}, redis, task18OutboxProbe{})
	if err != nil {
		t.Fatal(err)
	}

	wantStatus := []string{"degraded", "degraded", "down"}
	wantReady := []bool{true, true, false}
	for index := range wantStatus {
		ready, checks := checker.Check(context.Background())
		if ready != wantReady[index] || checks["redis"] != wantStatus[index] {
			t.Fatalf("check %d: ready=%v redis=%q, want %v/%q", index+1, ready, checks["redis"], wantReady[index], wantStatus[index])
		}
	}
}

func TestPolicyCanceledDeadlineOnlyRedisResultDoesNotAdvanceState(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		steps := make(chan func(context.Context) error, 5)
		steps <- func(context.Context) error { return nil }
		steps <- func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		}
		failure := func(context.Context) error { return errors.New("CANARY") }
		steps <- failure
		steps <- failure
		steps <- failure
		redis := fakeProbe{name: "redis", ping: func(ctx context.Context) error {
			return (<-steps)(ctx)
		}}
		checker, err := NewWithPolicy(500*time.Millisecond, DefaultPolicy(), fakeProbe{name: "postgres"}, redis, task18OutboxProbe{})
		if err != nil {
			t.Fatal(err)
		}

		ready, checks := checker.Check(context.Background())
		if !ready || checks["redis"] != "up" {
			t.Fatalf("initial Redis success: ready=%v checks=%v", ready, checks)
		}
		requireTask19RedisState(t, checker, "up", 0, 0)

		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		ready, _ = checker.Check(canceled)
		if ready {
			t.Fatal("canceled readiness check returned ready")
		}
		synctest.Wait()
		requireTask19RedisState(t, checker, "up", 0, 0)

		wantStatus := []string{"degraded", "degraded", "down"}
		wantReady := []bool{true, true, false}
		for index := range wantStatus {
			ready, checks = checker.Check(context.Background())
			if ready != wantReady[index] || checks["redis"] != wantStatus[index] {
				t.Fatalf("failure %d: ready=%v redis=%q, want %v/%q", index+1, ready, checks["redis"], wantReady[index], wantStatus[index])
			}
		}
	})
}

func TestPolicyLateRedisSuccessAfterOverallTimeoutDoesNotAdvanceState(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		steps := make(chan func(context.Context) error, 6)
		failure := func(context.Context) error { return errors.New("CANARY") }
		steps <- failure
		steps <- failure
		steps <- failure
		releaseLateSuccess := make(chan struct{})
		steps <- func(context.Context) error {
			<-releaseLateSuccess
			return nil
		}
		steps <- func(context.Context) error { return nil }
		steps <- func(context.Context) error { return nil }
		redis := fakeProbe{name: "redis", ping: func(ctx context.Context) error {
			return (<-steps)(ctx)
		}}
		checker, err := NewWithPolicy(500*time.Millisecond, DefaultPolicy(), fakeProbe{name: "postgres"}, redis, task18OutboxProbe{})
		if err != nil {
			t.Fatal(err)
		}

		wantFailureStatus := []string{"degraded", "degraded", "down"}
		wantFailureReady := []bool{true, true, false}
		for index := range wantFailureStatus {
			ready, checks := checker.Check(context.Background())
			if ready != wantFailureReady[index] || checks["redis"] != wantFailureStatus[index] {
				t.Fatalf("failure %d: ready=%v redis=%q, want %v/%q", index+1, ready, checks["redis"], wantFailureReady[index], wantFailureStatus[index])
			}
		}
		requireTask19RedisState(t, checker, "down", 3, 0)

		ready, _ := checker.Check(context.Background())
		if ready {
			t.Fatal("timed-out readiness check returned ready")
		}
		close(releaseLateSuccess)
		synctest.Wait()
		requireTask19RedisState(t, checker, "down", 3, 0)

		wantRecoveryStatus := []string{"degraded", "up"}
		for index := range wantRecoveryStatus {
			ready, checks := checker.Check(context.Background())
			if !ready || checks["redis"] != wantRecoveryStatus[index] {
				t.Fatalf("recovery %d: ready=%v redis=%q, want true/%q", index+1, ready, checks["redis"], wantRecoveryStatus[index])
			}
		}
	})
}

func TestPolicyNATSSocketStateAloneNeverDeterminesReadiness(t *testing.T) {
	checker := New(time.Second,
		fakeProbe{name: "postgres"},
		fakeProbe{name: "redis"},
		fakeProbe{name: "nats", ping: func(context.Context) error { return errors.New("CANARY") }},
	)
	ready, checks := checker.Check(context.Background())
	if !ready || checks["nats"] != "degraded" {
		t.Fatalf("NATS socket failure: ready=%v checks=%v", ready, checks)
	}
}

func TestPolicyOutboxThresholdsAreInclusive(t *testing.T) {
	tests := []struct {
		name    string
		backlog int64
		oldest  time.Duration
		status  string
		ready   bool
	}{
		{name: "below", backlog: 999, oldest: 59 * time.Second, status: "up", ready: true},
		{name: "degraded backlog", backlog: 1000, oldest: 0, status: "degraded", ready: true},
		{name: "degraded age", backlog: 0, oldest: 60 * time.Second, status: "degraded", ready: true},
		{name: "down backlog", backlog: 10000, oldest: 0, status: "down", ready: false},
		{name: "down age", backlog: 0, oldest: 300 * time.Second, status: "down", ready: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			checker := task18PolicyChecker(t,
				&task18ScriptedProbe{name: "postgres"},
				&task18ScriptedProbe{name: "redis"},
				task18OutboxProbe{backlog: test.backlog, oldest: test.oldest},
			)
			ready, checks := checker.Check(context.Background())
			if ready != test.ready || checks["outbox"] != test.status {
				t.Fatalf("outbox snapshot: ready=%v status=%q, want %v/%q", ready, checks["outbox"], test.ready, test.status)
			}
		})
	}
}

func TestPolicyExcludesSignerEmailAndReporterFromProbeSurface(t *testing.T) {
	checker := task18PolicyChecker(t,
		&task18ScriptedProbe{name: "postgres"},
		&task18ScriptedProbe{name: "redis"},
		task18OutboxProbe{},
	)
	ready, checks := checker.Check(context.Background())
	if !ready {
		t.Fatalf("healthy policy was not ready: %v", checks)
	}
	for _, excluded := range []string{"signer", "email", "reporter", "nats"} {
		if _, exists := checks[excluded]; exists {
			t.Fatalf("excluded async/provider component %q appeared as a policy probe", excluded)
		}
	}
}

func TestTask18PostgresOutboxHealthProbeUsesGeneratedSnapshot(t *testing.T) {
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	source := &task18OutboxSource{row: store.GetOutboxHealthRow{Backlog: 1234, OldestAgeSeconds: 61.25}}
	probe, err := newPostgresOutboxHealthProbe(source, task18ReadinessClock{now: now})
	if err != nil {
		t.Fatal(err)
	}
	backlog, oldest, err := probe.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if source.now != now || backlog != 1234 || oldest != 61250*time.Millisecond {
		t.Fatalf("snapshot = now:%s backlog:%d oldest:%s", source.now, backlog, oldest)
	}
}

func TestTask18PostgresOutboxHealthProbeRejectsInvalidOrPrivateResults(t *testing.T) {
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name  string
		row   store.GetOutboxHealthRow
		err   error
		clock task18ReadinessClock
	}{
		{name: "query", err: errors.New("CANARY postgres private detail"), clock: task18ReadinessClock{now: now}},
		{name: "negative backlog", row: store.GetOutboxHealthRow{Backlog: -1}, clock: task18ReadinessClock{now: now}},
		{name: "negative age", row: store.GetOutboxHealthRow{OldestAgeSeconds: -1}, clock: task18ReadinessClock{now: now}},
		{name: "NaN age", row: store.GetOutboxHealthRow{OldestAgeSeconds: math.NaN()}, clock: task18ReadinessClock{now: now}},
		{name: "infinite age", row: store.GetOutboxHealthRow{OldestAgeSeconds: math.Inf(1)}, clock: task18ReadinessClock{now: now}},
		{name: "zero clock", clock: task18ReadinessClock{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			probe, err := newPostgresOutboxHealthProbe(&task18OutboxSource{row: test.row, err: test.err}, test.clock)
			if err != nil {
				t.Fatal(err)
			}
			_, _, err = probe.Snapshot(context.Background())
			if !errors.Is(err, ErrOutboxHealth) {
				t.Fatalf("snapshot error = %v, want fixed %v", err, ErrOutboxHealth)
			}
			if err != nil && (strings.Contains(err.Error(), "CANARY") || len(err.Error()) > 64) {
				t.Fatalf("snapshot exposed private detail: %q", err)
			}
		})
	}
}

func task18PolicyChecker(t *testing.T, postgres, redis Probe, outbox OutboxHealthProbe) *Checker {
	t.Helper()
	checker, err := NewWithPolicy(time.Second, Policy{
		RedisDownAfterFailures: 3, RedisRecoverAfterSuccesses: 2,
		OutboxDegradedBacklog: 1000, OutboxDownBacklog: 10000,
		OutboxDegradedAge: 60 * time.Second, OutboxDownAge: 300 * time.Second,
	}, postgres, redis, outbox)
	if err != nil {
		t.Fatal(err)
	}
	return checker
}

type task18ScriptedProbe struct {
	name     string
	failures []bool
	calls    int
}

func (probe *task18ScriptedProbe) Name() string { return probe.name }

func (probe *task18ScriptedProbe) Ping(context.Context) error {
	failed := false
	if probe.calls < len(probe.failures) {
		failed = probe.failures[probe.calls]
	}
	probe.calls++
	if failed {
		return errors.New("CANARY dependency detail")
	}
	return nil
}

type task18OutboxProbe struct {
	backlog int64
	oldest  time.Duration
	err     error
}

func (probe task18OutboxProbe) Snapshot(context.Context) (int64, time.Duration, error) {
	return probe.backlog, probe.oldest, probe.err
}

type task18OutboxSource struct {
	row store.GetOutboxHealthRow
	err error
	now time.Time
}

func (source *task18OutboxSource) GetOutboxHealth(_ context.Context, now time.Time) (store.GetOutboxHealthRow, error) {
	source.now = now
	return source.row, source.err
}

type task18ReadinessClock struct{ now time.Time }

func (clock task18ReadinessClock) Now() time.Time { return clock.now }

func requireTask19RedisState(t *testing.T, checker *Checker, status string, failures, successes int) {
	t.Helper()
	checker.redis.mu.Lock()
	defer checker.redis.mu.Unlock()
	if checker.redis.status != status || checker.redis.failures != failures || checker.redis.successes != successes {
		t.Fatalf(
			"Redis state = %q/%d/%d, want %q/%d/%d",
			checker.redis.status, checker.redis.failures, checker.redis.successes,
			status, failures, successes,
		)
	}
}
