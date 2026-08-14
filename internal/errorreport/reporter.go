// Package errorreport provides a privacy-safe, bounded asynchronous error reporter.
package errorreport

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"time"
)

const forcedWorkerWait = 100 * time.Millisecond

// Event is the finite report-event registry.
type Event string

// Category is the finite public failure category registry.
type Category string

// Component is the finite reporting component registry.
type Component string

// Outcome is the finite operation outcome registry.
type Outcome string

// Fingerprint is a static code-defined error fingerprint.
type Fingerprint string

// DeliveryResult is the finite reporter-delivery metric result.
type DeliveryResult string

//nolint:revive // These closed report registries are documented by their exported types and self-describing names.
const (
	// EventRequestFailure and the following constants are the finite report registries.
	EventRequestFailure Event = "request_failure"
	// EventWorkerFailure reports a bounded asynchronous worker failure.
	EventWorkerFailure Event = "worker_failure"
	// EventDependencyFailure reports a bounded dependency failure.
	EventDependencyFailure Event = "dependency_failure"

	CategoryAuthentication Category = "authentication"
	CategoryDependency     Category = "dependency"
	CategoryCryptography   Category = "cryptography"
	CategoryInternal       Category = "internal"

	ComponentControlAPI Component = "control_api"
	ComponentIdentity   Component = "identity"
	ComponentDeviceAuth Component = "deviceauth"
	ComponentTrust      Component = "trust"
	ComponentOutbox     Component = "outbox"
	ComponentEmail      Component = "email"
	ComponentReporter   Component = "reporter"

	OutcomeFailure   Outcome = "failure"
	OutcomeRecovered Outcome = "recovered"

	FingerprintAuthentication Fingerprint = "authentication"
	FingerprintDependency     Fingerprint = "dependency"
	FingerprintCryptography   Fingerprint = "cryptography"
	FingerprintInternal       Fingerprint = "internal"

	ResultSent           DeliveryResult = "sent"
	ResultDropped        DeliveryResult = "dropped"
	ResultProviderFailed DeliveryResult = "provider_failed"
)

// ErrInvalidConfiguration reports an unsafe reporter construction.
var ErrInvalidConfiguration = errors.New("errorreport: invalid configuration")

// Report is the complete fixed report DTO. It intentionally has no raw error,
// map, request, body, header, or identifier field.
type Report struct {
	Event        Event
	Category     Category
	Component    Component
	Outcome      Outcome
	Fingerprint  Fingerprint
	BuildVersion string
	TraceID      string
}

// Reporter is the nonblocking application-facing reporting boundary.
type Reporter interface {
	TryReport(Report) bool
}

// Provider delivers one bounded report batch to an adapter-owned backend.
type Provider interface {
	Send(context.Context, []Report) error
}

// Observer records only finite component/result pairs.
type Observer interface {
	Record(Component, DeliveryResult)
}

// AsyncReporter owns one bounded queue and one provider worker.
type AsyncReporter struct {
	provider Provider
	observer Observer
	queue    chan Report
	batchMax int
	timeout  time.Duration

	mu        sync.RWMutex
	accepting bool
	closeOnce sync.Once
	done      chan struct{}
	workerCtx context.Context
	cancel    context.CancelFunc
}

var _ Reporter = (*AsyncReporter)(nil)

// New validates and starts one bounded reporter worker.
func New(owner context.Context, provider Provider, queueCapacity, batchMaximum int, timeout time.Duration, observer Observer) (*AsyncReporter, error) {
	if owner == nil || nilValue(provider) || nilValue(observer) || queueCapacity < 10 || queueCapacity > 1000 ||
		batchMaximum < 1 || batchMaximum > 100 || batchMaximum > queueCapacity ||
		timeout < 100*time.Millisecond || timeout > 2*time.Second {
		return nil, ErrInvalidConfiguration
	}
	workerCtx, cancel := context.WithCancel(context.WithoutCancel(owner))
	reporter := &AsyncReporter{
		provider: provider, observer: observer, queue: make(chan Report, queueCapacity),
		batchMax: batchMaximum, timeout: timeout, accepting: true,
		done: make(chan struct{}), workerCtx: workerCtx, cancel: cancel,
	}
	go reporter.run(workerCtx)
	return reporter, nil
}

// TryReport validates and enqueues without waiting for provider work.
func (reporter *AsyncReporter) TryReport(report Report) bool {
	if reporter == nil || !validReport(report) {
		return false
	}
	reporter.mu.RLock()
	defer reporter.mu.RUnlock()
	if !reporter.accepting {
		return false
	}
	select {
	case reporter.queue <- report:
		return true
	default:
		reporter.record(report.Component, ResultDropped)
		return false
	}
}

// Close rejects new reports, drains until ctx expires, and then drops the
// remaining queue without recursively reporting provider failure.
func (reporter *AsyncReporter) Close(ctx context.Context) {
	if reporter == nil {
		return
	}
	reporter.closeOnce.Do(func() {
		reporter.mu.Lock()
		reporter.accepting = false
		close(reporter.queue)
		reporter.mu.Unlock()
	})
	if ctx == nil {
		reporter.cancel()
		<-reporter.done
		return
	}
	select {
	case <-reporter.done:
		return
	case <-ctx.Done():
		reporter.cancel()
	}
	select {
	case <-reporter.done:
	case <-time.After(forcedWorkerWait):
	}
}

func (reporter *AsyncReporter) run(ctx context.Context) {
	defer close(reporter.done)
	defer reporter.cancel()
	for {
		first, ok := <-reporter.queue
		if !ok {
			return
		}
		batch := make([]Report, 1, reporter.batchMax)
		batch[0] = first
		queueClosed := false
		for len(batch) < reporter.batchMax {
			select {
			case report, open := <-reporter.queue:
				if !open {
					queueClosed = true
				} else {
					batch = append(batch, report)
				}
			default:
				queueClosed = false
			}
			if queueClosed || len(batch) == reporter.batchMax || len(reporter.queue) == 0 {
				break
			}
		}
		reporter.deliver(ctx, batch)
		if queueClosed {
			return
		}
	}
}

func (reporter *AsyncReporter) deliver(ctx context.Context, batch []Report) {
	if ctx.Err() != nil {
		for _, report := range batch {
			reporter.record(report.Component, ResultDropped)
		}
		return
	}
	providerCtx, cancel := context.WithTimeout(ctx, reporter.timeout)
	failed := providerFailed(providerCtx, reporter.provider, batch)
	cancel()
	if failed {
		result := ResultProviderFailed
		if ctx.Err() != nil {
			result = ResultDropped
		}
		for _, report := range batch {
			reporter.record(report.Component, result)
		}
		return
	}
	for _, report := range batch {
		reporter.record(report.Component, ResultSent)
	}
}

func providerFailed(ctx context.Context, provider Provider, batch []Report) (failed bool) {
	defer func() {
		if recover() != nil {
			failed = true
		}
	}()
	return provider.Send(ctx, append([]Report(nil), batch...)) != nil
}

func (reporter *AsyncReporter) record(component Component, result DeliveryResult) {
	defer func() { _ = recover() }()
	reporter.observer.Record(component, result)
}

func validReport(report Report) bool {
	return validEvent(report.Event) && validCategory(report.Category) && validComponent(report.Component) &&
		validOutcome(report.Outcome) && validFingerprint(report.Fingerprint) &&
		validPublicScalar(report.BuildVersion, 1, 64) && validTraceID(report.TraceID)
}

func validEvent(value Event) bool {
	switch value {
	case EventRequestFailure, EventWorkerFailure, EventDependencyFailure:
		return true
	default:
		return false
	}
}

func validCategory(value Category) bool {
	switch value {
	case CategoryAuthentication, CategoryDependency, CategoryCryptography, CategoryInternal:
		return true
	default:
		return false
	}
}

func validComponent(value Component) bool {
	switch value {
	case ComponentControlAPI, ComponentIdentity, ComponentDeviceAuth, ComponentTrust, ComponentOutbox, ComponentEmail, ComponentReporter:
		return true
	default:
		return false
	}
}

func validOutcome(value Outcome) bool {
	return value == OutcomeFailure || value == OutcomeRecovered
}

func validFingerprint(value Fingerprint) bool {
	switch value {
	case FingerprintAuthentication, FingerprintDependency, FingerprintCryptography, FingerprintInternal:
		return true
	default:
		return false
	}
}

func validPublicScalar(value string, minimum, maximum int) bool {
	if len(value) < minimum || len(value) > maximum || strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' ||
			character >= '0' && character <= '9' || character == '.' || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
}

func validTraceID(value string) bool {
	if len(value) != 32 {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' && character < 'a' || character > 'f' {
			return false
		}
	}
	return true
}

func nilValue(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	kind := reflected.Kind()
	return (kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface ||
		kind == reflect.Map || kind == reflect.Pointer || kind == reflect.Slice) && reflected.IsNil()
}
