package ratelimit

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"talenro.local/platform/internal/config"
)

func TestRedisUsesOnePrivateAtomicScriptAndStrictResult(t *testing.T) {
	t.Parallel()

	digest := [32]byte{0x01, 0x02, 0x03, 0x04}
	policy := config.RateLimitPolicy{Limit: 5, Window: time.Hour}
	executor := &fakeRedisExecutor{result: []any{int64(5), int64(time.Hour / time.Millisecond)}}
	adapter, err := newRedisWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	allowed, err := adapter.Allow(context.Background(), Delivery, digest, policy)
	if err != nil || !allowed {
		t.Fatalf("Allow = %v, %v", allowed, err)
	}
	if executor.calls != 1 || len(executor.keys) != 1 || len(executor.arguments) != 1 {
		t.Fatalf("script calls/keys/arguments = %d/%d/%d", executor.calls, len(executor.keys), len(executor.arguments))
	}
	for _, exact := range []string{"redis.call('INCR', KEYS[1])", "if count == 1 then", "redis.call('PEXPIRE', KEYS[1], ARGV[1])", "redis.call('PTTL', KEYS[1])", "return {count, ttl}"} {
		if strings.Count(executor.script, exact) != 1 {
			t.Fatalf("Lua fragment %q count = %d", exact, strings.Count(executor.script, exact))
		}
	}
	if strings.Contains(executor.keys[0], "member@example") || executor.keys[0] != "talenro:ratelimit:0102030400000000000000000000000000000000000000000000000000000000" {
		t.Fatalf("Redis key was not digest-only: %q", executor.keys[0])
	}
	if got, ok := executor.arguments[0].(int64); !ok || got != int64(time.Hour/time.Millisecond) {
		t.Fatalf("PEXPIRE argument = %#v", executor.arguments[0])
	}
	if executor.deadlineClass != 250*time.Millisecond {
		t.Fatalf("Redis deadline class = %s", executor.deadlineClass)
	}

	executor.result = []any{int64(6), int64(3_599_999)}
	allowed, err = adapter.Allow(context.Background(), Delivery, digest, policy)
	if err != nil || allowed {
		t.Fatalf("over-limit Allow = %v, %v", allowed, err)
	}
}

func TestRedisRejectsInvalidPolicyResultAndCancellationValueFree(t *testing.T) {
	t.Parallel()

	digest := [32]byte{0x44}
	validPolicy := config.RateLimitPolicy{Limit: 5, Window: time.Hour}
	tests := []struct {
		name   string
		result any
		err    error
	}{
		{name: "not array", result: int64(1)},
		{name: "wrong length", result: []any{int64(1)}},
		{name: "string count", result: []any{"1", int64(1)}},
		{name: "zero count", result: []any{int64(0), int64(1)}},
		{name: "negative ttl", result: []any{int64(1), int64(-1)}},
		{name: "ttl beyond window", result: []any{int64(1), int64(time.Hour/time.Millisecond + 1)}},
		{name: "backend", err: errors.New("SECRET_REDIS_CANARY")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := &fakeRedisExecutor{result: test.result, err: test.err}
			adapter, newErr := newRedisWithExecutor(executor, 250*time.Millisecond)
			if newErr != nil {
				t.Fatal(newErr)
			}
			_, allowErr := adapter.Allow(context.Background(), Delivery, digest, validPolicy)
			if !errors.Is(allowErr, ErrBackendUnavailable) || strings.Contains(allowErr.Error(), "SECRET_REDIS_CANARY") {
				t.Fatalf("error = %v, want value-free backend error", allowErr)
			}
		})
	}

	executor := &fakeRedisExecutor{}
	adapter, err := newRedisWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []struct {
		operation Operation
		policy    config.RateLimitPolicy
	}{
		{operation: Operation("arbitrary"), policy: validPolicy},
		{operation: Delivery, policy: config.RateLimitPolicy{Limit: 21, Window: time.Hour}},
		{operation: Delivery, policy: config.RateLimitPolicy{Limit: 5, Window: time.Minute}},
	} {
		if _, allowErr := adapter.Allow(context.Background(), invalid.operation, digest, invalid.policy); !errors.Is(allowErr, ErrInvalidInput) {
			t.Fatalf("invalid operation/policy error = %v", allowErr)
		}
	}
	if executor.calls != 0 {
		t.Fatal("invalid input reached Redis")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, allowErr := adapter.Allow(ctx, Delivery, digest, validPolicy); !errors.Is(allowErr, ErrCanceled) {
		t.Fatalf("pre-canceled error = %v", allowErr)
	}
}

func TestRedisConcurrentCountAllowsExactlyThePolicyLimit(t *testing.T) {
	t.Parallel()

	executor := &atomicRedisExecutor{}
	adapter, err := newRedisWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	policy := config.RateLimitPolicy{Limit: 5, Window: time.Hour}
	digest := [32]byte{0x77}
	var allowed atomic.Int32
	var workers sync.WaitGroup
	for range 100 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			ok, allowErr := adapter.Allow(context.Background(), Delivery, digest, policy)
			if allowErr != nil {
				t.Errorf("Allow: %v", allowErr)
				return
			}
			if ok {
				allowed.Add(1)
			}
		}()
	}
	workers.Wait()
	if got := allowed.Load(); got != 5 {
		t.Fatalf("allowed = %d, want 5", got)
	}
}

func TestRedisRejectsTypedNilExecutor(t *testing.T) {
	t.Parallel()

	var executor *fakeRedisExecutor
	if adapter, err := newRedisWithExecutor(executor, 250*time.Millisecond); !errors.Is(err, ErrInvalidInput) || adapter != nil {
		t.Fatalf("typed-nil executor = %#v, %v", adapter, err)
	}
	adapter, err := newRedisWithExecutor(&fakeRedisExecutor{}, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	var ctx *typedNilRedisContext
	if _, allowErr := adapter.Allow(ctx, Delivery, [32]byte{1}, config.RateLimitPolicy{Limit: 5, Window: time.Hour}); !errors.Is(allowErr, ErrInvalidInput) {
		t.Fatalf("typed-nil context error = %v", allowErr)
	}
}

type typedNilRedisContext struct{}

func (*typedNilRedisContext) Deadline() (time.Time, bool) { panic("typed nil context used") }
func (*typedNilRedisContext) Done() <-chan struct{}       { panic("typed nil context used") }
func (*typedNilRedisContext) Err() error                  { panic("typed nil context used") }
func (*typedNilRedisContext) Value(any) any               { panic("typed nil context used") }

type fakeRedisExecutor struct {
	result        any
	err           error
	script        string
	keys          []string
	arguments     []any
	calls         int
	deadlineClass time.Duration
}

func (fake *fakeRedisExecutor) Run(ctx context.Context, script string, keys []string, arguments ...any) (any, error) {
	fake.calls++
	fake.script = script
	fake.keys = append([]string(nil), keys...)
	fake.arguments = append([]any(nil), arguments...)
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		fake.deadlineClass = remaining.Round(time.Millisecond)
	}
	return fake.result, fake.err
}

type atomicRedisExecutor struct {
	mu    sync.Mutex
	count int64
}

func (fake *atomicRedisExecutor) Run(_ context.Context, _ string, _ []string, arguments ...any) (any, error) {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	fake.count++
	return []any{fake.count, arguments[0].(int64)}, nil
}
