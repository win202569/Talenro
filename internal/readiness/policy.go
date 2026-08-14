package readiness

import (
	"context"
	"errors"
	"math"
	"time"

	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/store"
)

var errInvalidPolicy = errors.New("readiness: invalid policy")

// ErrOutboxHealth is the fixed outbox snapshot failure classification.
var ErrOutboxHealth = errors.New("readiness: outbox health unavailable")

// Policy contains the bounded readiness thresholds.
type Policy struct {
	RedisDownAfterFailures     int
	RedisRecoverAfterSuccesses int
	OutboxDegradedBacklog      int64
	OutboxDownBacklog          int64
	OutboxDegradedAge          time.Duration
	OutboxDownAge              time.Duration
}

// DefaultPolicy returns the reviewed local thresholds.
func DefaultPolicy() Policy {
	return Policy{
		RedisDownAfterFailures:     3,
		RedisRecoverAfterSuccesses: 2,
		OutboxDegradedBacklog:      1000,
		OutboxDownBacklog:          10000,
		OutboxDegradedAge:          60 * time.Second,
		OutboxDownAge:              300 * time.Second,
	}
}

// OutboxHealthProbe reports the unpublished backlog and oldest event age.
type OutboxHealthProbe interface {
	Snapshot(context.Context) (backlog int64, oldest time.Duration, err error)
}

type outboxHealthSource interface {
	GetOutboxHealth(context.Context, time.Time) (store.GetOutboxHealthRow, error)
}

// PostgresOutboxHealthProbe adapts the generated outbox health query.
type PostgresOutboxHealthProbe struct {
	source outboxHealthSource
	clock  securitykit.Clock
}

// NewPostgresOutboxHealthProbe binds the generated query to a UTC clock.
func NewPostgresOutboxHealthProbe(database store.DBTX, clock securitykit.Clock) (*PostgresOutboxHealthProbe, error) {
	if isNilValue(database) {
		return nil, errInvalidPolicy
	}
	return newPostgresOutboxHealthProbe(store.New(database), clock)
}

func newPostgresOutboxHealthProbe(source outboxHealthSource, clock securitykit.Clock) (*PostgresOutboxHealthProbe, error) {
	if isNilValue(source) || isNilValue(clock) {
		return nil, errInvalidPolicy
	}
	return &PostgresOutboxHealthProbe{source: source, clock: clock}, nil
}

// Snapshot returns a bounded conversion of the generated aggregate row.
func (probe *PostgresOutboxHealthProbe) Snapshot(ctx context.Context) (int64, time.Duration, error) {
	if probe == nil || isNilValue(probe.source) || isNilValue(probe.clock) || ctx == nil || ctx.Err() != nil {
		return 0, 0, ErrOutboxHealth
	}
	now := probe.clock.Now()
	if now.IsZero() || now.Location() != time.UTC || now.Year() < 2020 || now.Year() > 2100 {
		return 0, 0, ErrOutboxHealth
	}
	row, err := probe.source.GetOutboxHealth(ctx, now)
	maximumSeconds := float64(int64(^uint64(0)>>1)) / float64(time.Second)
	if err != nil || row.Backlog < 0 || row.OldestAgeSeconds < 0 || math.IsNaN(row.OldestAgeSeconds) ||
		math.IsInf(row.OldestAgeSeconds, 0) || row.OldestAgeSeconds > maximumSeconds {
		return 0, 0, ErrOutboxHealth
	}
	return row.Backlog, time.Duration(row.OldestAgeSeconds * float64(time.Second)), nil
}

// NewWithPolicy constructs the authoritative PostgreSQL/Redis/outbox checker.
func NewWithPolicy(timeout time.Duration, policy Policy, postgres, redis Probe, outbox OutboxHealthProbe) (*Checker, error) {
	if timeout <= 0 || !validPolicy(policy) {
		return nil, errInvalidPolicy
	}
	if isNilValue(postgres) || postgres.Name() != "postgres" {
		return nil, errors.New("readiness: invalid postgres probe")
	}
	if isNilValue(redis) || redis.Name() != "redis" {
		return nil, errors.New("readiness: invalid redis probe")
	}
	if isNilValue(outbox) {
		return nil, errors.New("readiness: invalid outbox probe")
	}

	return &Checker{
		timeout: timeout,
		policy:  policy,
		probes: []checkedProbe{
			{name: "postgres", kind: probePostgres, probe: postgres},
			{name: "redis", kind: probeRedis, probe: redis},
			{name: "outbox", kind: probeOutbox, outbox: outbox},
		},
	}, nil
}

func validPolicy(policy Policy) bool {
	return policy.RedisDownAfterFailures >= 1 && policy.RedisDownAfterFailures <= 10 &&
		policy.RedisRecoverAfterSuccesses >= 1 && policy.RedisRecoverAfterSuccesses <= 10 &&
		policy.OutboxDegradedBacklog >= 100 && policy.OutboxDegradedBacklog <= 10000 &&
		policy.OutboxDownBacklog > policy.OutboxDegradedBacklog && policy.OutboxDownBacklog <= 100000 &&
		policy.OutboxDegradedAge >= 10*time.Second && policy.OutboxDegradedAge <= 10*time.Minute &&
		policy.OutboxDownAge >= policy.OutboxDegradedAge+time.Second && policy.OutboxDownAge <= time.Hour
}
