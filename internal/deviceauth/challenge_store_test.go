package deviceauth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const task12ChallengeID = "d6df0ff2-b723-4b62-bc43-f63cc1bf012a"

func TestChallengeRedisStoreCreatesOnlyOpaqueBoundedState(t *testing.T) {
	t.Parallel()

	executor := &fakeDeviceChallengeExecutor{setAllowed: true}
	challengeStore, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	record := task12ChallengeRecord()
	if err := challengeStore.Create(context.Background(), record, 2*time.Minute); err != nil {
		t.Fatal(err)
	}
	if executor.setKey != "talenro:deviceauth:challenge:"+task12ChallengeID || executor.setTTL != 2*time.Minute {
		t.Fatalf("Redis key/TTL = %q/%s, want random ID and exact two minutes", executor.setKey, executor.setTTL)
	}
	for _, fixed := range [][]byte{[]byte("registration"), []byte("device-pop-v1"), []byte("register_device")} {
		if !bytes.Contains(executor.setValue, fixed) {
			t.Fatalf("Redis record omitted fixed contract %q", fixed)
		}
	}
	for _, forbidden := range [][]byte{
		[]byte("GRANT-TOKEN-CANARY"), []byte("PRINCIPAL-CANARY"), []byte("DISPLAY-NAME-CANARY"),
		[]byte("IDEMPOTENCY-CANARY"), []byte("PRIVATE-KEY-CANARY"), []byte("PUBLIC-KEY-CANARY"),
	} {
		if bytes.Contains(executor.setValue, forbidden) || strings.Contains(executor.setKey, string(forbidden)) {
			t.Fatalf("Redis state exposed forbidden value %q", forbidden)
		}
	}
}

func TestChallengeSingleUseConsumesWithOneAtomicGetDEL(t *testing.T) {
	t.Parallel()

	record := task12ChallengeRecord()
	wire, err := encodeChallengeRecord(record)
	if err != nil {
		t.Fatal(err)
	}
	executor := &fakeDeviceChallengeExecutor{runValues: []any{wire, nil}}
	challengeStore, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	consumed, err := challengeStore.Consume(context.Background(), record.ChallengeID, record.GrantDigest, record.ContextDigest)
	if err != nil {
		t.Fatal(err)
	}
	if consumed != record {
		t.Fatal("consumed challenge did not preserve the exact private record")
	}
	if executor.runCalls != 1 || executor.runScript != "return redis.call('GETDEL', KEYS[1])" {
		t.Fatal("challenge consume did not perform exactly one atomic GETDEL")
	}
	if _, err := challengeStore.Consume(context.Background(), record.ChallengeID, record.GrantDigest, record.ContextDigest); !errors.Is(err, ErrChallengeNotFound) {
		t.Fatalf("second consume error = %v, want finite not-found", err)
	}
}

func TestChallengeConsumeFailsClosedOnDigestMismatchAndAmbiguity(t *testing.T) {
	t.Parallel()

	record := task12ChallengeRecord()
	wire, err := encodeChallengeRecord(record)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name          string
		executor      redisChallengeExecutor
		grantDigest   [32]byte
		contextDigest [32]byte
		want          error
	}{
		{name: "grant mismatch", executor: &fakeDeviceChallengeExecutor{runValues: []any{wire}}, grantDigest: [32]byte{0xff}, contextDigest: record.ContextDigest, want: ErrChallengeNotFound},
		{name: "context mismatch", executor: &fakeDeviceChallengeExecutor{runValues: []any{wire}}, grantDigest: record.GrantDigest, contextDigest: [32]byte{0xff}, want: ErrChallengeNotFound},
		{name: "malformed record", executor: &fakeDeviceChallengeExecutor{runValues: []any{[]byte("TOKEN-CANARY")}}, grantDigest: record.GrantDigest, contextDigest: record.ContextDigest, want: ErrChallengeUnavailable},
		{name: "backend failure", executor: &fakeDeviceChallengeExecutor{runErrors: []error{errors.New("REDIS-BACKEND-CANARY")}}, grantDigest: record.GrantDigest, contextDigest: record.ContextDigest, want: ErrChallengeUnavailable},
		{name: "provider panic", executor: panicDeviceChallengeExecutor{}, grantDigest: record.GrantDigest, contextDigest: record.ContextDigest, want: ErrChallengeUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			challengeStore, err := newRedisChallengeStoreWithExecutor(test.executor, 250*time.Millisecond)
			if err != nil {
				t.Fatal(err)
			}
			_, err = challengeStore.Consume(context.Background(), record.ChallengeID, test.grantDigest, test.contextDigest)
			if !errors.Is(err, test.want) || strings.Contains(err.Error(), "CANARY") {
				t.Fatalf("consume error = %v, want sanitized %v", err, test.want)
			}
		})
	}
}

func TestChallengeRecordRedactsFormattingAndRejectsJSON(t *testing.T) {
	t.Parallel()

	record := task12ChallengeRecord()
	rendered := fmt.Sprintf("%+v", record)
	if strings.Contains(rendered, record.ChallengeID) || strings.Contains(rendered, "device-pop-v1") {
		t.Fatalf("challenge record formatting leaked fields: %q", rendered)
	}
	if _, err := json.Marshal(record); err == nil {
		t.Fatal("challenge record JSON serialization succeeded")
	}
	var client *goredis.Client
	if challengeStore, err := NewRedisChallengeStore(client, 250*time.Millisecond); challengeStore != nil || !errors.Is(err, ErrInvalidChallenge) {
		t.Fatalf("typed-nil Redis constructor = %#v / %v", challengeStore, err)
	}
}

func task12ChallengeRecord() ChallengeRecord {
	return ChallengeRecord{
		ChallengeID: task12ChallengeID, Kind: ChallengeRegistration, ProtocolVersion: "device-pop-v1", Operation: "register_device",
		Challenge: [32]byte{1, 2, 3}, GrantDigest: sha256.Sum256([]byte("opaque-grant-digest")),
		ContextDigest: sha256.Sum256([]byte("opaque-public-context")), ExpiresAt: time.Now().UTC().Add(time.Minute),
	}
}

type fakeDeviceChallengeExecutor struct {
	setAllowed bool
	setError   error
	setKey     string
	setValue   []byte
	setTTL     time.Duration
	runValues  []any
	runErrors  []error
	runCalls   int
	runScript  string
	runKeys    []string
}

func (executor *fakeDeviceChallengeExecutor) SetNX(_ context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	executor.setKey = key
	executor.setValue = bytes.Clone(value)
	executor.setTTL = ttl
	return executor.setAllowed, executor.setError
}

func (executor *fakeDeviceChallengeExecutor) Run(_ context.Context, script string, keys []string, _ ...any) (any, error) {
	executor.runScript = script
	executor.runKeys = append([]string(nil), keys...)
	index := executor.runCalls
	executor.runCalls++
	if index < len(executor.runErrors) && executor.runErrors[index] != nil {
		return nil, executor.runErrors[index]
	}
	if index >= len(executor.runValues) {
		return nil, nil
	}
	return executor.runValues[index], nil
}

type panicDeviceChallengeExecutor struct{}

func (panicDeviceChallengeExecutor) SetNX(context.Context, string, []byte, time.Duration) (bool, error) {
	panic("REDIS-SET-CANARY")
}
func (panicDeviceChallengeExecutor) Run(context.Context, string, []string, ...any) (any, error) {
	panic("REDIS-RUN-CANARY")
}

var _ redisChallengeExecutor = (*fakeDeviceChallengeExecutor)(nil)
var _ redisChallengeExecutor = panicDeviceChallengeExecutor{}
