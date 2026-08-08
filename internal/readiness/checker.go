// Package readiness performs bounded dependency health aggregation.
package readiness

import (
	"context"
	"reflect"
	"time"
)

const (
	statusOK          = "ok"
	statusUnavailable = "unavailable"
)

// Probe is a named dependency health check.
type Probe interface {
	Name() string
	Ping(context.Context) error
}

// Checker runs validated probes within a shared timeout.
type Checker struct {
	timeout time.Duration
	probes  []checkedProbe
}

// New validates and constructs a readiness checker.
func New(timeout time.Duration, probes ...Probe) *Checker {
	checked := make([]checkedProbe, 0, len(probes))
	seen := make(map[string]struct{}, len(probes))
	for _, probe := range probes {
		if isNilProbe(probe) {
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
		checked = append(checked, checkedProbe{name: name, probe: probe})
	}

	return &Checker{timeout: timeout, probes: checked}
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
		checks[probe.name] = statusUnavailable

		go func() {
			status := statusOK
			if probe.probe.Ping(checkCtx) != nil {
				status = statusUnavailable
			}
			results <- result{name: probe.name, status: status}
		}()
	}

	ready := true
	for range c.probes {
		select {
		case <-checkCtx.Done():
			return false, checks
		default:
		}

		select {
		case <-checkCtx.Done():
			return false, checks
		case probeResult := <-results:
			if checkCtx.Err() != nil {
				return false, checks
			}
			checks[probeResult.name] = probeResult.status
			if probeResult.status != statusOK {
				ready = false
			}
		}
	}

	return ready, checks
}

type checkedProbe struct {
	name  string
	probe Probe
}

func isNilProbe(probe Probe) bool {
	if probe == nil {
		return true
	}

	value := reflect.ValueOf(probe)
	kind := value.Kind()
	isNilCapable := kind == reflect.Chan ||
		kind == reflect.Func ||
		kind == reflect.Interface ||
		kind == reflect.Map ||
		kind == reflect.Pointer ||
		kind == reflect.Slice
	return isNilCapable && value.IsNil()
}

func isPublicProbeName(name string) bool {
	switch name {
	case "postgres", "redis", "nats":
		return true
	default:
		return false
	}
}
