package readiness

import (
	"context"
	"time"
)

const (
	statusOK          = "ok"
	statusUnavailable = "unavailable"
)

type Probe interface {
	Name() string
	Ping(context.Context) error
}

type Checker struct {
	timeout time.Duration
	probes  []Probe
}

func New(timeout time.Duration, probes ...Probe) *Checker {
	return &Checker{timeout: timeout, probes: append([]Probe(nil), probes...)}
}

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
		name := probe.Name()
		checks[name] = statusUnavailable

		go func() {
			status := statusOK
			if probe.Ping(checkCtx) != nil {
				status = statusUnavailable
			}
			results <- result{name: name, status: status}
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
