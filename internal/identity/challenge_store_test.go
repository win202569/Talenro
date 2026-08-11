package identity

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const task10ChallengeID = "d6df0ff2-b723-4b62-bc43-f63cc1bf012a"

func TestRedisChallengeStoreCreatesOpaqueKeyAndPrivateFixedValue(t *testing.T) {
	executor := &fakeChallengeExecutor{setAllowed: true}
	store, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	record := task10ChallengeRecord()
	if err = store.Create(context.Background(), record, 2*time.Minute); err != nil {
		t.Fatal(err)
	}
	if executor.setTTL != 2*time.Minute {
		t.Fatalf("TTL = %s", executor.setTTL)
	}
	if executor.setKey != "talenro:identity:account-challenge:"+task10ChallengeID {
		t.Fatal("Redis key was not derived solely from the random challenge ID")
	}
	for _, want := range [][]byte{[]byte("account-rotation-v1"), []byte("rotate_account_token")} {
		if !bytes.Contains(executor.setValue, want) {
			t.Fatalf("record omitted fixed contract %q", want)
		}
	}
	for _, forbidden := range [][]byte{[]byte("TOKEN-CANARY"), []byte("PRINCIPAL-CANARY"), []byte("DEVICE-CANARY")} {
		if bytes.Contains(executor.setValue, forbidden) || bytes.Contains([]byte(executor.setKey), forbidden) {
			t.Fatal("Redis challenge state exposed forbidden authority material")
		}
	}
}

func TestRedisChallengeStoreConsumesWithOneAtomicGetDEL(t *testing.T) {
	record := task10ChallengeRecord()
	wire, err := encodeChallengeRecord(record)
	if err != nil {
		t.Fatal(err)
	}
	executor := &fakeChallengeExecutor{runValues: []any{wire, nil}}
	store, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	consumed, err := store.Consume(context.Background(), record.ChallengeID, record.ContextDigest)
	if err != nil {
		t.Fatal(err)
	}
	if consumed != record {
		t.Fatalf("consumed record = %#v", consumed)
	}
	if executor.runCalls != 1 || executor.runScript != "return redis.call('GETDEL', KEYS[1])" {
		t.Fatal("consume did not use exactly one atomic GETDEL script")
	}
	_, err = store.Consume(context.Background(), record.ChallengeID, record.ContextDigest)
	if !errors.Is(err, ErrChallengeNotFound) {
		t.Fatalf("second consume error = %v", err)
	}
}

func TestRedisChallengeStoreFailsClosedOnMismatchAndAmbiguity(t *testing.T) {
	record := task10ChallengeRecord()
	wire, err := encodeChallengeRecord(record)
	if err != nil {
		t.Fatal(err)
	}
	backendCanary := errors.New("REDIS-BACKEND-CANARY")
	tests := []struct {
		name      string
		executor  *fakeChallengeExecutor
		digest    [32]byte
		wantError error
	}{
		{name: "context mismatch", executor: &fakeChallengeExecutor{runValues: []any{wire}}, digest: [32]byte{0xff}, wantError: ErrChallengeNotFound},
		{name: "backend ambiguity", executor: &fakeChallengeExecutor{runErrors: []error{backendCanary}}, digest: record.ContextDigest, wantError: ErrChallengeUnavailable},
		{name: "deadline ambiguity", executor: &fakeChallengeExecutor{runWaitForContext: true}, digest: record.ContextDigest, wantError: ErrChallengeUnavailable},
		{name: "malformed value", executor: &fakeChallengeExecutor{runValues: []any{[]byte("TOKEN-CANARY")}}, digest: record.ContextDigest, wantError: ErrChallengeUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store, newErr := newRedisChallengeStoreWithExecutor(test.executor, 250*time.Millisecond)
			if newErr != nil {
				t.Fatal(newErr)
			}
			_, consumeErr := store.Consume(context.Background(), record.ChallengeID, test.digest)
			if !errors.Is(consumeErr, test.wantError) {
				t.Fatalf("consume error = %v", consumeErr)
			}
			if bytes.Contains([]byte(consumeErr.Error()), []byte("CANARY")) {
				t.Fatal("challenge error disclosed backend or stored data")
			}
		})
	}
}

func TestRedisChallengeStoreCreateFailsClosedOnCollisionAndAmbiguity(t *testing.T) {
	backendCanary := errors.New("REDIS-CREATE-CANARY")
	tests := []struct {
		name     string
		executor *fakeChallengeExecutor
	}{
		{name: "collision", executor: &fakeChallengeExecutor{}},
		{name: "backend ambiguity", executor: &fakeChallengeExecutor{setAllowed: true, setError: backendCanary}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store, newErr := newRedisChallengeStoreWithExecutor(test.executor, 250*time.Millisecond)
			if newErr != nil {
				t.Fatal(newErr)
			}
			createErr := store.Create(context.Background(), task10ChallengeRecord(), accountChallengeTTL)
			if !errors.Is(createErr, ErrChallengeUnavailable) || bytes.Contains([]byte(createErr.Error()), []byte("CANARY")) {
				t.Fatalf("create error = %v", createErr)
			}
		})
	}
}

func TestRedisChallengeStoreRejectsTypedNilClient(t *testing.T) {
	var client *goredis.Client
	store, err := NewRedisChallengeStore(client, 250*time.Millisecond)
	if store != nil || !errors.Is(err, ErrInvalidChallenge) {
		t.Fatalf("typed-nil constructor result = %#v / %v", store, err)
	}
}

func task10ChallengeRecord() ChallengeRecord {
	return ChallengeRecord{
		ChallengeID: task10ChallengeID, ProtocolVersion: "account-rotation-v1", Operation: "rotate_account_token",
		Challenge: [32]byte{1, 2, 3}, ContextDigest: [32]byte{4, 5, 6},
		ExpiresAt: time.Date(2026, 8, 10, 12, 2, 0, 123456789, time.UTC),
	}
}

type fakeChallengeExecutor struct {
	setAllowed        bool
	setError          error
	setKey            string
	setValue          []byte
	setTTL            time.Duration
	runValues         []any
	runErrors         []error
	runCalls          int
	runScript         string
	runKeys           []string
	runArguments      []any
	runHook           func()
	runWaitForContext bool
}

var _ func(goredis.UniversalClient, time.Duration) (*RedisChallengeStore, error) = NewRedisChallengeStore

func (executor *fakeChallengeExecutor) SetNX(_ context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	executor.setKey = key
	executor.setValue = bytes.Clone(value)
	executor.setTTL = ttl
	return executor.setAllowed, executor.setError
}

func (executor *fakeChallengeExecutor) Run(ctx context.Context, script string, keys []string, arguments ...any) (any, error) {
	executor.runScript = script
	executor.runKeys = append([]string(nil), keys...)
	executor.runArguments = make([]any, len(arguments))
	for index, argument := range arguments {
		if value, ok := argument.([]byte); ok {
			executor.runArguments[index] = bytes.Clone(value)
			continue
		}
		executor.runArguments[index] = argument
	}
	index := executor.runCalls
	executor.runCalls++
	if executor.runHook != nil {
		hook := executor.runHook
		executor.runHook = nil
		hook()
	}
	if executor.runWaitForContext {
		<-ctx.Done()
		return nil, goredis.Nil
	}
	if index < len(executor.runErrors) && executor.runErrors[index] != nil {
		return nil, executor.runErrors[index]
	}
	if index >= len(executor.runValues) {
		return nil, nil
	}
	return executor.runValues[index], nil
}
