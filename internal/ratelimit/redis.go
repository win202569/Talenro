package ratelimit

import (
	"context"
	"encoding/hex"
	"errors"
	"reflect"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"talenro.local/platform/internal/config"
)

const (
	minimumRedisTimeout = 100 * time.Millisecond
	maximumRedisTimeout = time.Second
	redisKeyPrefix      = "talenro:ratelimit:"
	redisAllowScript    = `local count = redis.call('INCR', KEYS[1])
if count == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
local ttl = redis.call('PTTL', KEYS[1])
return {count, ttl}`
)

var (
	// ErrCanceled reports caller cancellation without retaining context values.
	ErrCanceled = errors.New("ratelimit: canceled")
)

type redisExecutor interface {
	Run(context.Context, string, []string, ...any) (any, error)
}

type goRedisExecutor struct{ client goredis.Scripter }

func (executor goRedisExecutor) Run(ctx context.Context, script string, keys []string, arguments ...any) (any, error) {
	return goredis.NewScript(script).Run(ctx, executor.client, keys, arguments...).Result()
}

// Redis applies one atomic fixed-window script to a digest-only Redis key.
type Redis struct {
	executor redisExecutor
	timeout  time.Duration
}

var _ Limiter = (*Redis)(nil)

// NewRedis validates and binds a go-redis scripting client.
func NewRedis(client goredis.Scripter, timeout time.Duration) (*Redis, error) {
	if nilRedisValue(client) {
		return nil, ErrInvalidInput
	}
	return newRedisWithExecutor(goRedisExecutor{client: client}, timeout)
}

func newRedisWithExecutor(executor redisExecutor, timeout time.Duration) (*Redis, error) {
	if nilRedisValue(executor) || timeout < minimumRedisTimeout || timeout > maximumRedisTimeout {
		return nil, ErrInvalidInput
	}
	return &Redis{executor: executor, timeout: timeout}, nil
}

// Allow increments one operation/window digest and fails closed on ambiguity.
func (adapter *Redis) Allow(ctx context.Context, operation Operation, digest [32]byte, policy config.RateLimitPolicy) (bool, error) {
	if nilRedisValue(ctx) || adapter == nil || nilRedisValue(adapter.executor) || !validRedisPolicy(operation, policy) || digest == [32]byte{} {
		return false, ErrInvalidInput
	}
	if ctx.Err() != nil {
		return false, ErrCanceled
	}
	windowUnits := policy.Window / time.Millisecond
	if windowUnits <= 0 || time.Duration(windowUnits)*time.Millisecond != policy.Window {
		return false, ErrInvalidInput
	}
	operationContext, cancel := context.WithTimeout(ctx, adapter.timeout)
	defer cancel()
	key := redisKeyPrefix + hex.EncodeToString(digest[:])
	result, err := adapter.executor.Run(operationContext, redisAllowScript, []string{key}, int64(windowUnits))
	if err != nil {
		if ctx.Err() != nil {
			return false, ErrCanceled
		}
		return false, ErrBackendUnavailable
	}
	if ctx.Err() != nil {
		return false, ErrCanceled
	}
	values, ok := result.([]any)
	if !ok || len(values) != 2 {
		return false, ErrBackendUnavailable
	}
	count, countOK := values[0].(int64)
	ttl, ttlOK := values[1].(int64)
	if !countOK || !ttlOK || count < 1 || ttl < 1 || ttl > int64(windowUnits) {
		return false, ErrBackendUnavailable
	}
	return count <= int64(policy.Limit), nil
}

func validRedisPolicy(operation Operation, policy config.RateLimitPolicy) bool {
	switch operation {
	case Login:
		return policy.Limit >= 1 && policy.Limit <= 100 && policy.Window >= time.Minute && policy.Window <= 24*time.Hour
	case Delivery:
		return policy.Limit >= 1 && policy.Limit <= 20 && policy.Window >= 10*time.Minute && policy.Window <= 24*time.Hour
	case Challenge:
		return policy.Limit >= 5 && policy.Limit <= 100 && policy.Window >= time.Minute && policy.Window <= 30*time.Minute
	default:
		return false
	}
}

func nilRedisValue(value any) bool {
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
