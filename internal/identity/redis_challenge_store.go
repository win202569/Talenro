package identity

import (
	"context"
	"errors"
	"reflect"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	accountChallengeRedisPrefix  = "talenro:identity:account-challenge:"
	accountChallengeGetDEL       = "return redis.call('GETDEL', KEYS[1])"
	minimumChallengeRedisTimeout = 100 * time.Millisecond
	maximumChallengeRedisTimeout = time.Second
)

type redisChallengeExecutor interface {
	SetNX(context.Context, string, []byte, time.Duration) (bool, error)
	Run(context.Context, string, []string, ...any) (any, error)
}

type goRedisChallengeExecutor struct{ client goredis.UniversalClient }

func (executor goRedisChallengeExecutor) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	return executor.client.SetNX(ctx, key, value, ttl).Result()
}

func (executor goRedisChallengeExecutor) Run(ctx context.Context, script string, keys []string, arguments ...any) (any, error) {
	return goredis.NewScript(script).Run(ctx, executor.client, keys, arguments...).Result()
}

// RedisChallengeStore persists only bounded, single-use account challenge state.
type RedisChallengeStore struct {
	executor redisChallengeExecutor
	timeout  time.Duration
}

var _ ChallengeStore = (*RedisChallengeStore)(nil)

// NewRedisChallengeStore binds a go-redis client to the fail-closed adapter.
func NewRedisChallengeStore(client goredis.UniversalClient, timeout time.Duration) (*RedisChallengeStore, error) {
	if nilChallengeDependency(client) {
		return nil, ErrInvalidChallenge
	}
	return newRedisChallengeStoreWithExecutor(goRedisChallengeExecutor{client: client}, timeout)
}

func newRedisChallengeStoreWithExecutor(executor redisChallengeExecutor, timeout time.Duration) (*RedisChallengeStore, error) {
	if nilChallengeDependency(executor) || timeout < minimumChallengeRedisTimeout || timeout > maximumChallengeRedisTimeout {
		return nil, ErrInvalidChallenge
	}
	return &RedisChallengeStore{executor: executor, timeout: timeout}, nil
}

// Create stores one collision-resistant random challenge ID for exactly two minutes.
func (store *RedisChallengeStore) Create(ctx context.Context, record ChallengeRecord, ttl time.Duration) error {
	if nilChallengeDependency(ctx) || store == nil || nilChallengeDependency(store.executor) || ttl != accountChallengeTTL || !validChallengeRecord(record) {
		return ErrInvalidChallenge
	}
	if ctx.Err() != nil {
		return ErrChallengeUnavailable
	}
	wire, err := encodeChallengeRecord(record)
	if err != nil {
		return ErrInvalidChallenge
	}
	defer clear(wire)
	operationContext, cancel := context.WithTimeout(ctx, store.timeout)
	defer cancel()
	created, err := store.executor.SetNX(operationContext, accountChallengeRedisPrefix+record.ChallengeID, wire, ttl)
	if err != nil || operationContext.Err() != nil || !created {
		return ErrChallengeUnavailable
	}
	return nil
}

// Consume executes one atomic GETDEL and validates the private context digest.
func (store *RedisChallengeStore) Consume(ctx context.Context, challengeID string, contextDigest [32]byte) (ChallengeRecord, error) {
	if nilChallengeDependency(ctx) || store == nil || nilChallengeDependency(store.executor) || contextDigest == [32]byte{} || !validChallengeID(challengeID) {
		return ChallengeRecord{}, ErrInvalidChallenge
	}
	if ctx.Err() != nil {
		return ChallengeRecord{}, ErrChallengeUnavailable
	}
	operationContext, cancel := context.WithTimeout(ctx, store.timeout)
	defer cancel()
	result, err := store.executor.Run(operationContext, accountChallengeGetDEL, []string{accountChallengeRedisPrefix + challengeID})
	if operationContext.Err() != nil {
		return ChallengeRecord{}, ErrChallengeUnavailable
	}
	if errors.Is(err, goredis.Nil) || (err == nil && result == nil) {
		return ChallengeRecord{}, ErrChallengeNotFound
	}
	if err != nil {
		return ChallengeRecord{}, ErrChallengeUnavailable
	}
	var wire []byte
	switch value := result.(type) {
	case string:
		wire = []byte(value)
	case []byte:
		wire = append([]byte(nil), value...)
	default:
		return ChallengeRecord{}, ErrChallengeUnavailable
	}
	defer clear(wire)
	record, decodeErr := decodeChallengeRecord(challengeID, wire)
	if decodeErr != nil {
		return ChallengeRecord{}, ErrChallengeUnavailable
	}
	if !challengeContextMatches(record.ContextDigest, contextDigest) {
		return ChallengeRecord{}, ErrChallengeNotFound
	}
	return record, nil
}

func validChallengeID(value string) bool {
	record := ChallengeRecord{
		ChallengeID: value, ProtocolVersion: accountRotationProtocolVersion, Operation: accountRotationOperation,
		Challenge: [32]byte{1}, ContextDigest: [32]byte{1}, ExpiresAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	return validChallengeRecord(record)
}

func nilChallengeDependency(value any) bool {
	if value == nil {
		return true
	}
	representation := reflect.ValueOf(value)
	switch representation.Kind() { //nolint:exhaustive // Only nil-capable interface representations matter.
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return representation.IsNil()
	default:
		return false
	}
}
