// Package readiness performs bounded dependency health aggregation.
package readiness

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"

	nodeauthority "talenro.local/platform/internal/nodecontrol/authority"
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

var errInvalidAuthorityProbe = errors.New("readiness: invalid authority probe")

type authorityReadinessChecker interface {
	CheckReady(context.Context) (nodeauthority.Readiness, error)
}

type authorityProbe struct {
	checker authorityReadinessChecker
}

func newAuthorityProbe(checker authorityReadinessChecker) (*authorityProbe, error) {
	if isNilValue(checker) {
		return nil, errInvalidAuthorityProbe
	}
	return &authorityProbe{checker: checker}, nil
}

func (*authorityProbe) Name() string { return "authority" }

func (probe *authorityProbe) Ping(ctx context.Context) error {
	if probe == nil || isNilValue(probe.checker) || ctx == nil || ctx.Err() != nil {
		return nodeauthority.ErrAuthorityUnavailable
	}
	readiness, err := probe.checker.CheckReady(ctx)
	if err != nil || !readiness.Ready || readiness.Reason != nodeauthority.ReadinessReady {
		return nodeauthority.ErrAuthorityUnavailable
	}
	return nil
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
	overallContext, cancelOverall := context.WithTimeout(ctx, c.timeout)
	defer cancelOverall()
	probeContext, cancelProbes := context.WithTimeout(overallContext, c.timeout/2)
	defer cancelProbes()

	type result struct {
		name    string
		kind    probeKind
		outcome probeOutcome
	}

	checks := make(map[string]string, len(c.probes))
	results := make(chan result, len(c.probes))
	for _, probe := range c.probes {
		checks[probe.name] = initialStatus(probe.kind)

		go func(probe checkedProbe) {
			results <- result{name: probe.name, kind: probe.kind, outcome: c.evaluate(probeContext, probe)}
		}(probe)
	}

	ready := true
	for range c.probes {
		select {
		case <-overallContext.Done():
			return false, checks
		case probeResult := <-results:
			if overallContext.Err() != nil {
				return false, checks
			}
			status := probeResult.outcome.status
			if probeResult.kind == probeRedis {
				status = c.redis.record(probeResult.outcome.success, c.policy)
			}
			checks[probeResult.name] = status
			if status == statusDown {
				ready = false
			}
		}
	}

	return ready, checks
}

type probeOutcome struct {
	status  string
	success bool
}

func (c *Checker) evaluate(ctx context.Context, checked checkedProbe) probeOutcome {
	switch checked.kind {
	case probePostgres, probeAuthority:
		if checked.probe.Ping(ctx) != nil {
			return probeOutcome{status: statusDown}
		}
		return probeOutcome{status: statusUp}
	case probeRedis:
		return probeOutcome{success: checked.probe.Ping(ctx) == nil}
	case probeNATS:
		if checked.probe.Ping(ctx) != nil {
			return probeOutcome{status: statusDegraded}
		}
		return probeOutcome{status: statusUp}
	case probeOutbox:
		backlog, oldest, err := checked.outbox.Snapshot(ctx)
		if err != nil || backlog < 0 || oldest < 0 {
			return probeOutcome{status: statusDown}
		}
		if backlog >= c.policy.OutboxDownBacklog || oldest >= c.policy.OutboxDownAge {
			return probeOutcome{status: statusDown}
		}
		if backlog >= c.policy.OutboxDegradedBacklog || oldest >= c.policy.OutboxDegradedAge {
			return probeOutcome{status: statusDegraded}
		}
		return probeOutcome{status: statusUp}
	default:
		return probeOutcome{status: statusDown}
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
	probeAuthority
)

func kindForName(name string) probeKind {
	switch name {
	case "postgres":
		return probePostgres
	case "redis":
		return probeRedis
	case "nats":
		return probeNATS
	case "authority":
		return probeAuthority
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
	case "postgres", "redis", "nats", "authority":
		return true
	default:
		return false
	}
}
