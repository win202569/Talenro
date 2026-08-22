package deviceauth

import (
	"context"
	"errors"
	"reflect"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"talenro.local/platform/internal/securitykit"
)

const (
	deviceChallengeRedisPrefix   = "talenro:deviceauth:challenge:"
	deviceChallengeGetDEL        = "return redis.call('GETDEL', KEYS[1])"
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

// RedisChallengeStore persists only bounded single-use device challenge state.
type RedisChallengeStore struct {
	executor redisChallengeExecutor
	timeout  time.Duration
	clock    securitykit.Clock
}

var _ ChallengeStore = (*RedisChallengeStore)(nil)

// NewRedisChallengeStore binds a go-redis client to the fail-closed device challenge adapter.
func NewRedisChallengeStore(client goredis.UniversalClient, timeout time.Duration, clock securitykit.Clock) (*RedisChallengeStore, error) {
	if nilChallengeDependency(client) {
		return nil, ErrInvalidChallenge
	}
	return newRedisChallengeStoreWithExecutor(goRedisChallengeExecutor{client: client}, timeout, clock)
}

func newRedisChallengeStoreWithExecutor(
	executor redisChallengeExecutor,
	timeout time.Duration,
	clock securitykit.Clock,
) (*RedisChallengeStore, error) {
	if nilChallengeDependency(executor) || nilChallengeDependency(clock) ||
		timeout < minimumChallengeRedisTimeout || timeout > maximumChallengeRedisTimeout {
		return nil, ErrInvalidChallenge
	}
	return &RedisChallengeStore{executor: executor, timeout: timeout, clock: clock}, nil
}

// Create stores a random challenge ID for its bounded remaining lifetime using SETNX.
func (store *RedisChallengeStore) Create(ctx context.Context, record ChallengeRecord, ttl time.Duration) error {
	if nilChallengeDependency(ctx) || store == nil || nilChallengeDependency(store.executor) ||
		ttl <= 0 || ttl > deviceChallengeTTL || !validChallengeRecord(record) {
		return ErrInvalidChallenge
	}
	ttl = ttl.Truncate(time.Millisecond)
	if ttl <= 0 {
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
	created, err := safeRedisChallengeSetNX(operationContext, store.executor, deviceChallengeRedisPrefix+record.ChallengeID, wire, ttl)
	if err != nil || operationContext.Err() != nil || !created {
		return ErrChallengeUnavailable
	}
	return nil
}

// Consume atomically removes and validates one exact registration challenge.
func (store *RedisChallengeStore) Consume(
	ctx context.Context,
	challengeID string,
	grantDigest [32]byte,
	contextDigest [32]byte,
) (ChallengeRecord, error) {
	if nilChallengeDependency(ctx) || store == nil || nilChallengeDependency(store.executor) || nilChallengeDependency(store.clock) ||
		grantDigest == [32]byte{} || contextDigest == [32]byte{} || !validChallengeID(challengeID) {
		return ChallengeRecord{}, ErrInvalidChallenge
	}
	if ctx.Err() != nil {
		return ChallengeRecord{}, ErrChallengeUnavailable
	}
	operationContext, cancel := context.WithTimeout(ctx, store.timeout)
	defer cancel()
	result, err := safeRedisChallengeRun(operationContext, store.executor, deviceChallengeGetDEL, []string{deviceChallengeRedisPrefix + challengeID})
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
	record, err := decodeChallengeRecord(challengeID, wire)
	if err != nil {
		return ChallengeRecord{}, ErrChallengeUnavailable
	}
	now, validClock := safeChallengeClockNow(store.clock)
	if !validClock {
		return ChallengeRecord{}, ErrChallengeUnavailable
	}
	if !record.ExpiresAt.After(now) || !challengeDigestsMatch(record, grantDigest, contextDigest) {
		return ChallengeRecord{}, ErrChallengeNotFound
	}
	return record, nil
}

func safeChallengeClockNow(clock securitykit.Clock) (now time.Time, valid bool) {
	defer func() {
		if recover() != nil {
			now = time.Time{}
			valid = false
		}
	}()
	now = clock.Now().UTC()
	return now, !now.IsZero()
}

func safeRedisChallengeSetNX(ctx context.Context, executor redisChallengeExecutor, key string, wire []byte, ttl time.Duration) (created bool, err error) {
	defer func() {
		if recover() != nil {
			created = false
			err = ErrChallengeUnavailable
		}
	}()
	return executor.SetNX(ctx, key, wire, ttl)
}

func safeRedisChallengeRun(ctx context.Context, executor redisChallengeExecutor, script string, keys []string) (result any, err error) {
	defer func() {
		if recover() != nil {
			result = nil
			err = ErrChallengeUnavailable
		}
	}()
	return executor.Run(ctx, script, keys)
}

func validChallengeID(value string) bool {
	record := ChallengeRecord{
		ChallengeID: value, Kind: ChallengeRegistration, ProtocolVersion: deviceProofProtocolVersion, Operation: registerDeviceOperation,
		Challenge: [32]byte{1}, GrantDigest: [32]byte{1}, ContextDigest: [32]byte{1},
		ExpiresAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
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
