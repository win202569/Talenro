package outbox

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
	"talenro.local/platform/internal/buildinfo"
	contractevents "talenro.local/platform/internal/contracts/events"
	"talenro.local/platform/internal/errorreport"
	"talenro.local/platform/internal/store"
)

const (
	publisherClaimLimit = int32(100)
	publisherLease      = 30 * time.Second
	maximumPublishBytes = 264 * 1024
	minimumRetryDelay   = time.Second
	maximumRetryDelay   = 30 * time.Second
	outboxWorkerTraceID = "00000000000000000000000000000001"
)

var (
	// ErrPublisherConfiguration reports an unsafe publisher construction.
	ErrPublisherConfiguration = errors.New("outbox: invalid publisher configuration")
	// ErrPublisherStore reports a fixed generated-store failure.
	ErrPublisherStore = errors.New("outbox: publisher store failed")
	// ErrPublisherEvent reports a malformed or unregistered stored envelope.
	ErrPublisherEvent = errors.New("outbox: invalid publisher event")
	// ErrPublish reports a broker failure without retaining provider details.
	ErrPublish = errors.New("outbox: publish failed")
	// ErrPublisherCanceled reports a shutdown deadline without raw details.
	ErrPublisherCanceled = errors.New("outbox: publisher canceled")
)

// PublisherStore is the lease/mark/release persistence boundary.
type PublisherStore interface {
	Claim(context.Context, time.Time, int32, time.Time) ([]store.TransactionalOutbox, error)
	MarkPublished(context.Context, uuid.UUID, time.Time) error
	Release(context.Context, uuid.UUID, time.Time) error
}

// Broker publishes one validated envelope with a stable idempotency ID.
type Broker interface {
	Publish(context.Context, string, string, []byte) error
}

// PublisherClock supplies UTC publication time.
type PublisherClock interface {
	Now() time.Time
}

// Publisher owns at most one polling goroutine.
type Publisher struct {
	store        PublisherStore
	broker       Broker
	clock        PublisherClock
	random       io.Reader
	pollInterval time.Duration
	reporter     errorreport.Reporter

	mu      sync.Mutex
	started bool
	closed  bool
	cancel  context.CancelFunc
	done    chan struct{}
}

// NewPublisher validates the bounded publisher dependencies.
func NewPublisher(
	boundStore PublisherStore,
	broker Broker,
	clock PublisherClock,
	random io.Reader,
	pollInterval time.Duration,
	reporter errorreport.Reporter,
) (*Publisher, error) {
	if publisherNil(boundStore) || publisherNil(broker) || publisherNil(clock) || publisherNil(random) ||
		publisherNil(reporter) || pollInterval <= 0 || pollInterval > time.Minute {
		return nil, ErrPublisherConfiguration
	}
	return &Publisher{
		store: boundStore, broker: broker, clock: clock, random: random,
		pollInterval: pollInterval, reporter: reporter,
	}, nil
}

// PublishAvailable claims and processes one bounded batch.
func (publisher *Publisher) PublishAvailable(ctx context.Context) (int, error) {
	if publisher == nil || ctx == nil || ctx.Err() != nil {
		return 0, ErrPublisherCanceled
	}
	now := publisher.clock.Now()
	if !validPublisherTime(now) {
		return 0, ErrPublisherStore
	}
	rows, err := publisher.store.Claim(ctx, now, publisherClaimLimit, now.Add(publisherLease))
	if err != nil || len(rows) > int(publisherClaimLimit) {
		return 0, publisherContextOr(ctx, ErrPublisherStore)
	}
	processed := 0
	var batchErr error
	for index := range rows {
		row := rows[index]
		subject, eventErr := validatePublisherRow(row)
		if eventErr != nil {
			publisher.release(ctx, row, now)
			batchErr = ErrPublisherEvent
			continue
		}
		if err := publisher.broker.Publish(ctx, subject, row.EventID.String(), bytes.Clone(row.Payload)); err != nil {
			publisher.release(ctx, row, now)
			batchErr = publisherContextOr(ctx, ErrPublish)
			continue
		}
		if err := publisher.store.MarkPublished(ctx, row.EventID, now); err != nil {
			batchErr = publisherContextOr(ctx, ErrPublisherStore)
			continue
		}
		processed++
	}
	return processed, batchErr
}

// Start launches exactly one publisher goroutine.
func (publisher *Publisher) Start(owner context.Context) bool {
	if publisher == nil || owner == nil {
		return false
	}
	publisher.mu.Lock()
	defer publisher.mu.Unlock()
	if publisher.started || publisher.closed {
		return false
	}
	workerCtx, cancel := context.WithCancel(context.WithoutCancel(owner))
	publisher.started = true
	publisher.cancel = cancel
	publisher.done = make(chan struct{})
	go publisher.run(workerCtx, publisher.done)
	return true
}

// Close stops intake and waits only until the caller's shutdown deadline.
func (publisher *Publisher) Close(ctx context.Context) error {
	if publisher == nil {
		return nil
	}
	publisher.mu.Lock()
	publisher.closed = true
	cancel, done := publisher.cancel, publisher.done
	publisher.mu.Unlock()
	if cancel == nil || done == nil {
		return nil
	}
	cancel()
	if ctx == nil {
		<-done
		return nil
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ErrPublisherCanceled
	}
}

func (publisher *Publisher) run(ctx context.Context, done chan struct{}) {
	defer close(done)
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			_, err := publisher.PublishAvailable(ctx)
			if err != nil && ctx.Err() == nil {
				publisher.reportFailure()
			}
			timer.Reset(publisher.pollInterval)
		}
	}
}

func (publisher *Publisher) reportFailure() {
	if publisher == nil || publisher.reporter == nil {
		return
	}
	defer func() { _ = recover() }()
	publisher.reporter.TryReport(errorreport.Report{
		Event: errorreport.EventWorkerFailure, Category: errorreport.CategoryDependency,
		Component: errorreport.ComponentOutbox, Outcome: errorreport.OutcomeFailure,
		Fingerprint: errorreport.FingerprintDependency, BuildVersion: buildinfo.Current().Version,
		TraceID: outboxWorkerTraceID,
	})
}

func (publisher *Publisher) release(ctx context.Context, row store.TransactionalOutbox, now time.Time) {
	delay, err := publisher.retryDelay(row.Attempts)
	if err != nil {
		delay = retryBase(row.Attempts)
	}
	_ = publisher.store.Release(ctx, row.EventID, now.Add(delay))
}

func (publisher *Publisher) retryDelay(attempts int32) (time.Duration, error) {
	base := retryBase(attempts)
	var sample [8]byte
	if _, err := io.ReadFull(publisher.random, sample[:]); err != nil {
		return 0, ErrPublish
	}
	minimum := base * 8 / 10
	span := base*4/10 + 1
	offset := time.Duration(binary.BigEndian.Uint64(sample[:]) % uint64(span)) // #nosec G115 -- span is a validated positive duration no greater than twelve seconds.
	return minimum + offset, nil
}

func retryBase(attempts int32) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	delay := minimumRetryDelay
	for step := int32(1); step < attempts; step++ {
		if delay >= maximumRetryDelay/2 {
			return maximumRetryDelay
		}
		delay *= 2
	}
	if delay > maximumRetryDelay {
		return maximumRetryDelay
	}
	return delay
}

func validatePublisherRow(row store.TransactionalOutbox) (string, error) {
	if row.EventID == uuid.Nil || row.AggregateID == uuid.Nil || row.AggregateVersion < 1 ||
		len(row.Payload) == 0 || len(row.Payload) > maximumPublishBytes {
		return "", ErrPublisherEvent
	}
	envelope, err := contractevents.UnmarshalEnvelope(row.Payload)
	if err != nil || envelope.GetEventId() != row.EventID.String() || envelope.GetEventType() != row.EventType ||
		envelope.GetAggregateId() != row.AggregateID.String() || envelope.GetAggregateType() != row.AggregateType ||
		envelope.GetAggregateVersion() != uint64(row.AggregateVersion) || envelope.GetIdempotencyKey() != row.IdempotencyKey {
		return "", ErrPublisherEvent
	}
	subject, ok := SubjectForEventType(row.EventType)
	if !ok {
		return "", ErrPublisherEvent
	}
	return subject, nil
}

// SubjectForEventType maps the complete finite event registry to subjects.
func SubjectForEventType(eventType string) (string, bool) {
	switch eventType {
	case contractevents.AccountStateChangedType,
		contractevents.EmailDeliveryRequestedType,
		contractevents.DeviceAuthorizationChangedType,
		contractevents.DeviceTokenFamilyCompromisedType,
		contractevents.BundleIssuedType,
		contractevents.BundleAcknowledgedType:
		return eventType, true
	default:
		return "", false
	}
}

func validPublisherTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Year() >= 2020 && value.Year() <= 2100
}

func publisherContextOr(ctx context.Context, fallback error) error {
	if ctx != nil && ctx.Err() != nil {
		return ErrPublisherCanceled
	}
	return fallback
}

func publisherNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	kind := reflected.Kind()
	return (kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface ||
		kind == reflect.Map || kind == reflect.Pointer || kind == reflect.Slice) && reflected.IsNil()
}

// PostgresPublisherStore adapts the generated outbox queries.
type PostgresPublisherStore struct{ queries *store.Queries }

// NewPostgresPublisherStore binds the publisher to generated queries only.
func NewPostgresPublisherStore(database store.DBTX) (*PostgresPublisherStore, error) {
	if publisherNil(database) {
		return nil, ErrPublisherConfiguration
	}
	return &PostgresPublisherStore{queries: store.New(database)}, nil
}

// Claim acquires one generated-query lease batch.
func (bound *PostgresPublisherStore) Claim(ctx context.Context, now time.Time, limit int32, claimedUntil time.Time) ([]store.TransactionalOutbox, error) {
	if bound == nil || bound.queries == nil {
		return nil, ErrPublisherStore
	}
	rows, err := bound.queries.ClaimOutboxBatch(ctx, store.ClaimOutboxBatchParams{
		AvailableAt: now, Limit: limit, ClaimedUntil: sql.NullTime{Time: claimedUntil, Valid: true},
	})
	if err != nil {
		return nil, ErrPublisherStore
	}
	return rows, nil
}

// MarkPublished completes a publish transition idempotently.
func (bound *PostgresPublisherStore) MarkPublished(ctx context.Context, eventID uuid.UUID, now time.Time) error {
	if bound == nil || bound.queries == nil {
		return ErrPublisherStore
	}
	rows, err := bound.queries.MarkOutboxPublished(ctx, store.MarkOutboxPublishedParams{
		EventID: eventID, PublishedAt: sql.NullTime{Time: now, Valid: true},
	})
	if err != nil || rows < 0 || rows > 1 {
		return ErrPublisherStore
	}
	return nil
}

// Release schedules a failed claim for its next bounded retry.
func (bound *PostgresPublisherStore) Release(ctx context.Context, eventID uuid.UUID, availableAt time.Time) error {
	if bound == nil || bound.queries == nil {
		return ErrPublisherStore
	}
	rows, err := bound.queries.ReleaseOutboxClaim(ctx, store.ReleaseOutboxClaimParams{EventID: eventID, AvailableAt: availableAt})
	if err != nil || rows < 0 || rows > 1 {
		return ErrPublisherStore
	}
	return nil
}
