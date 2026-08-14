// Package readiness performs bounded dependency health aggregation.
package readiness

import (
	"context"
	"reflect"
	"sync"
	"time"
)

const (
	statusUp       = "up"
	statusDegraded = "degraded"
	statusDown     = "down"
)

// Probe is a named dependency health check.
type Probe interface {
	Name() string
	Ping(context.Context) error
}

// Checker runs validated probes within a shared timeout.
type Checker struct {
	timeout time.Duration
	policy  Policy
	probes  []checkedProbe
	redis   redisState
}

// New validates and constructs a checker using the local default policy.
// It is retained for callers that do not need the outbox health probe.
func New(timeout time.Duration, probes ...Probe) *Checker {
	checked := make([]checkedProbe, 0, len(probes))
	seen := make(map[string]struct{}, len(probes))
	for _, probe := range probes {
		if isNilValue(probe) {
			panic("readiness: nil probe")
		}

		name := probe.Name()
		if !isPublicProbeName(name) {
			panic("readiness: unsafe probe name")
		}
		if _, exists := seen[name]; exists {
			panic("readiness: duplicate probe name")
		}

		seen[name] = struct{}{}
		checked = append(checked, checkedProbe{name: name, kind: kindForName(name), probe: probe})
	}

	return &Checker{timeout: timeout, policy: DefaultPolicy(), probes: checked}
}

// Check runs all probes and returns only bounded public states.
func (c *Checker) Check(ctx context.Context) (bool, map[string]string) {
	checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	type result struct {
		name   string
		status string
	}

	checks := make(map[string]string, len(c.probes))
	results := make(chan result, len(c.probes))
	for _, probe := range c.probes {
		checks[probe.name] = initialStatus(probe.kind)

		go func(probe checkedProbe) {
			status := c.evaluate(checkCtx, probe)
			select {
			case results <- result{name: probe.name, status: status}:
			case <-checkCtx.Done():
			}
		}(probe)
	}

	ready := true
	for range c.probes {
		select {
		case <-checkCtx.Done():
			return false, checks
		case probeResult := <-results:
			if checkCtx.Err() != nil {
				return false, checks
			}
			checks[probeResult.name] = probeResult.status
			if probeResult.status == statusDown {
				ready = false
			}
		}
	}

	return ready, checks
}

func (c *Checker) evaluate(ctx context.Context, checked checkedProbe) string {
	switch checked.kind {
	case probePostgres:
		if checked.probe.Ping(ctx) != nil {
			return statusDown
		}
		return statusUp
	case probeRedis:
		return c.redis.record(checked.probe.Ping(ctx) == nil, c.policy)
	case probeNATS:
		if checked.probe.Ping(ctx) != nil {
			return statusDegraded
		}
		return statusUp
	case probeOutbox:
		backlog, oldest, err := checked.outbox.Snapshot(ctx)
		if err != nil || backlog < 0 || oldest < 0 {
			return statusDown
		}
		if backlog >= c.policy.OutboxDownBacklog || oldest >= c.policy.OutboxDownAge {
			return statusDown
		}
		if backlog >= c.policy.OutboxDegradedBacklog || oldest >= c.policy.OutboxDegradedAge {
			return statusDegraded
		}
		return statusUp
	default:
		return statusDown
	}
}

type checkedProbe struct {
	name   string
	kind   probeKind
	probe  Probe
	outbox OutboxHealthProbe
}

type probeKind uint8

const (
	probePostgres probeKind = iota + 1
	probeRedis
	probeNATS
	probeOutbox
)

func kindForName(name string) probeKind {
	switch name {
	case "postgres":
		return probePostgres
	case "redis":
		return probeRedis
	case "nats":
		return probeNATS
	default:
		return 0
	}
}

func initialStatus(kind probeKind) string {
	if kind == probeRedis || kind == probeNATS {
		return statusDegraded
	}
	return statusDown
}

type redisState struct {
	mu        sync.Mutex
	status    string
	failures  int
	successes int
}

func (state *redisState) record(success bool, policy Policy) string {
	state.mu.Lock()
	defer state.mu.Unlock()

	if state.status == "" {
		state.status = statusUp
	}
	if !success {
		state.failures++
		state.successes = 0
		if state.failures >= policy.RedisDownAfterFailures {
			state.status = statusDown
		} else {
			state.status = statusDegraded
		}
		return state.status
	}

	state.failures = 0
	if state.status == statusUp {
		state.successes = 0
		return statusUp
	}
	state.successes++
	if state.successes >= policy.RedisRecoverAfterSuccesses {
		state.status = statusUp
		state.successes = 0
		return statusUp
	}
	state.status = statusDegraded
	return statusDegraded
}

func isNilValue(value any) bool {
	if value == nil {
		return true
	}

	reflected := reflect.ValueOf(value)
	kind := reflected.Kind()
	isNilCapable := kind == reflect.Chan ||
		kind == reflect.Func ||
		kind == reflect.Interface ||
		kind == reflect.Map ||
		kind == reflect.Pointer ||
		kind == reflect.Slice
	return isNilCapable && reflected.IsNil()
}

func isPublicProbeName(name string) bool {
	switch name {
	case "postgres", "redis", "nats":
		return true
	default:
		return false
	}
}
