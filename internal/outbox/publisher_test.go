package outbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"talenro.local/platform/internal/buildinfo"
	contractevents "talenro.local/platform/internal/contracts/events"
	"talenro.local/platform/internal/errorreport"
	"talenro.local/platform/internal/store"
)

func TestPublisherClaimsAtMostOneHundredWithThirtySecondLeaseAndMarksAfterPublish(t *testing.T) {
	now := time.Date(2026, time.August, 13, 14, 0, 0, 0, time.UTC)
	row := task18PublishedRow(t, now, 1)
	storeFake := &task18PublisherStore{claimed: []store.TransactionalOutbox{row}}
	broker := &task18Broker{trace: storeFake}
	publisher, err := NewPublisher(storeFake, broker, task18PublisherClock{now: now}, bytes.NewReader(make([]byte, 64)), time.Second, newTask18PublisherReporter())
	if err != nil {
		t.Fatal(err)
	}

	processed, err := publisher.PublishAvailable(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if processed != 1 {
		t.Fatalf("processed = %d, want 1", processed)
	}
	if storeFake.limit != 100 || !storeFake.now.Equal(now) || !storeFake.claimedUntil.Equal(now.Add(30*time.Second)) {
		t.Fatalf("claim = limit %d now %s until %s", storeFake.limit, storeFake.now, storeFake.claimedUntil)
	}
	if broker.calls != 1 || broker.subject != contractevents.AccountStateChangedType || broker.messageID != row.EventID.String() || !bytes.Equal(broker.payload, row.Payload) {
		t.Fatal("publisher did not publish the validated raw envelope with event UUID message ID")
	}
	if storeFake.markCalls != 1 || storeFake.releaseCalls != 0 || storeFake.trace != "claim,publish,mark" {
		t.Fatalf("publisher transition trace = %q", storeFake.trace)
	}
}

func TestPublisherRejectsUnknownOrMismatchedEventBeforeBroker(t *testing.T) {
	now := time.Date(2026, time.August, 13, 14, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		mutate func(*store.TransactionalOutbox)
	}{
		{name: "unknown event type", mutate: func(row *store.TransactionalOutbox) { row.EventType = "talenro.unknown.CANARY.v1" }},
		{name: "row event mismatch", mutate: func(row *store.TransactionalOutbox) { row.EventID = uuid.New() }},
		{name: "malformed envelope", mutate: func(row *store.TransactionalOutbox) { row.Payload = []byte("CANARY") }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			row := task18PublishedRow(t, now, 1)
			test.mutate(&row)
			storeFake := &task18PublisherStore{claimed: []store.TransactionalOutbox{row}}
			broker := &task18Broker{trace: storeFake}
			publisher, err := NewPublisher(storeFake, broker, task18PublisherClock{now: now}, bytes.NewReader(make([]byte, 64)), time.Second, newTask18PublisherReporter())
			if err != nil {
				t.Fatal(err)
			}
			if _, err = publisher.PublishAvailable(context.Background()); !errors.Is(err, ErrPublisherEvent) || strings.Contains(err.Error(), "CANARY") {
				t.Fatalf("publisher error = %v, want fixed ErrPublisherEvent", err)
			}
			if broker.calls != 0 || storeFake.markCalls != 0 || storeFake.releaseCalls != 1 {
				t.Fatal("invalid event reached the broker or was not released")
			}
		})
	}
}

func TestPublisherFailureReleasesWithBoundedCryptoJitterAndNoRawError(t *testing.T) {
	now := time.Date(2026, time.August, 13, 14, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		attempt int32
		base    time.Duration
	}{
		{attempt: 1, base: time.Second},
		{attempt: 2, base: 2 * time.Second},
		{attempt: 3, base: 4 * time.Second},
		{attempt: 4, base: 8 * time.Second},
		{attempt: 5, base: 16 * time.Second},
		{attempt: 20, base: 30 * time.Second},
	} {
		t.Run(fmt.Sprintf("attempt-%d", test.attempt), func(t *testing.T) {
			row := task18PublishedRow(t, now, test.attempt)
			storeFake := &task18PublisherStore{claimed: []store.TransactionalOutbox{row}}
			broker := &task18Broker{err: errors.New("nats://secret@CANARY.internal"), trace: storeFake}
			publisher, err := NewPublisher(storeFake, broker, task18PublisherClock{now: now}, bytes.NewReader(bytes.Repeat([]byte{0xff}, 64)), time.Second, newTask18PublisherReporter())
			if err != nil {
				t.Fatal(err)
			}
			if _, err = publisher.PublishAvailable(context.Background()); !errors.Is(err, ErrPublish) || strings.Contains(err.Error(), "CANARY") {
				t.Fatalf("publisher error = %v, want fixed ErrPublish", err)
			}
			delay := storeFake.availableAt.Sub(now)
			if delay < test.base*8/10 || delay > test.base*12/10 || delay > 36*time.Second {
				t.Fatalf("attempt %d retry delay = %s, want %s ±20%%", test.attempt, delay, test.base)
			}
			if storeFake.trace != "claim,publish,release" || storeFake.markCalls != 0 || storeFake.releaseCalls != 1 {
				t.Fatalf("failure transition trace = %q", storeFake.trace)
			}
		})
	}
}

func TestPublisherWorkerStopsWithinCancellationDeadline(t *testing.T) {
	now := time.Date(2026, time.August, 13, 14, 0, 0, 0, time.UTC)
	storeFake := &task18PublisherStore{claimed: []store.TransactionalOutbox{task18PublishedRow(t, now, 1)}}
	broker := &task18Broker{block: make(chan struct{}), started: make(chan struct{}), trace: storeFake}
	reporter := newTask18PublisherReporter()
	publisher, err := NewPublisher(storeFake, broker, task18PublisherClock{now: now}, bytes.NewReader(make([]byte, 64)), 10*time.Millisecond, reporter)
	if err != nil {
		t.Fatal(err)
	}
	publisher.Start(context.Background())
	select {
	case <-broker.started:
	case <-time.After(time.Second):
		t.Fatal("publisher worker did not start")
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := publisher.Close(shutdown); err != nil {
		t.Fatalf("publisher close: %v", err)
	}
	if publisher.Start(context.Background()) {
		t.Fatal("closed publisher started a second worker")
	}
	select {
	case report := <-reporter.reports:
		t.Fatalf("publisher cancellation emitted a failure report: %#v", report)
	default:
	}
}

func TestPublisherWorkerReportsOnlyFixedPrivacySafeFailure(t *testing.T) {
	const canary = "nats://private-user:private-password@CANARY.internal/event-id-8472" // #nosec G101 -- an intentional privacy canary, not a credential.
	now := time.Date(2026, time.August, 13, 14, 0, 0, 0, time.UTC)
	storeFake := &task18PublisherStore{claimed: []store.TransactionalOutbox{task18PublishedRow(t, now, 1)}}
	broker := &task18Broker{err: errors.New(canary), trace: storeFake}
	reporter := newTask18PublisherReporter()
	publisher, err := NewPublisher(
		storeFake,
		broker,
		task18PublisherClock{now: now},
		bytes.NewReader(make([]byte, 64)),
		10*time.Millisecond,
		reporter,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !publisher.Start(context.Background()) {
		t.Fatal("publisher worker did not start")
	}
	t.Cleanup(func() { _ = publisher.Close(context.Background()) })

	select {
	case report := <-reporter.reports:
		want := errorreport.Report{
			Event:        errorreport.EventWorkerFailure,
			Category:     errorreport.CategoryDependency,
			Component:    errorreport.ComponentOutbox,
			Outcome:      errorreport.OutcomeFailure,
			Fingerprint:  errorreport.FingerprintDependency,
			BuildVersion: buildinfo.Current().Version,
			TraceID:      report.TraceID,
		}
		if report != want {
			t.Fatalf("worker report = %#v, want %#v", report, want)
		}
		if len(report.TraceID) != 32 || strings.Trim(report.TraceID, "0123456789abcdef") != "" {
			t.Fatalf("worker trace ID = %q, want fixed lowercase hex", report.TraceID)
		}
		if strings.Contains(fmt.Sprintf("%#v", report), canary) {
			t.Fatalf("worker report retained provider data: %#v", report)
		}
	case <-time.After(time.Second):
		t.Fatal("publisher worker did not submit a failure report")
	}
}

func TestSubjectForEventTypeIsFinite(t *testing.T) {
	for _, eventType := range []string{
		contractevents.AccountStateChangedType,
		contractevents.EmailDeliveryRequestedType,
		contractevents.DeviceAuthorizationChangedType,
		contractevents.DeviceTokenFamilyCompromisedType,
		contractevents.BundleIssuedType,
		contractevents.BundleAcknowledgedType,
	} {
		subject, ok := SubjectForEventType(eventType)
		if !ok || subject != eventType {
			t.Fatalf("registered event type %q has subject %q ok=%v", eventType, subject, ok)
		}
	}
	if subject, ok := SubjectForEventType("talenro.attacker.CANARY.v1"); ok || subject != "" {
		t.Fatal("attacker-controlled event type received a subject")
	}
}

func TestNATSBrokerRequiresBoundedJetStreamPubAckAndStableMessageID(t *testing.T) {
	connection := &task18JetStreamPublisher{ack: &nats.PubAck{Stream: EventStreamName, Sequence: 42}}
	broker, err := NewNATSBroker(connection, 200*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte{1, 2, 3}
	messageID := "11111111-1111-4111-8111-111111111111"
	if err := broker.Publish(context.Background(), contractevents.AccountStateChangedType, messageID, payload); err != nil {
		t.Fatal(err)
	}
	if connection.message == nil || connection.message.Subject != contractevents.AccountStateChangedType ||
		connection.message.Header.Get(nats.MsgIdHdr) != messageID || !bytes.Equal(connection.message.Data, payload) ||
		connection.calls != 1 {
		t.Fatal("NATS adapter did not synchronously publish the stable duplicate ID envelope")
	}
	deadline, bounded := connection.publishContext.Deadline()
	if !bounded || time.Until(deadline) <= 0 || time.Until(deadline) > 250*time.Millisecond {
		t.Fatalf("JetStream publish context deadline = %v bounded=%v", deadline, bounded)
	}
	payload[0] = 9
	if connection.message.Data[0] != 1 {
		t.Fatal("NATS message retained caller-owned payload bytes")
	}
}

func TestNATSBrokerAcceptsDuplicatePubAck(t *testing.T) {
	connection := &task18JetStreamPublisher{ack: &nats.PubAck{Stream: EventStreamName, Sequence: 7, Duplicate: true}}
	broker, err := NewNATSBroker(connection, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if err := broker.Publish(context.Background(), contractevents.AccountStateChangedType, uuid.NewString(), []byte{1}); err != nil {
		t.Fatalf("duplicate PubAck must be a successful at-least-once delivery: %v", err)
	}
}

func TestNATSBrokerRejectsInvalidPubAckProviderFailurePanicAndDeadline(t *testing.T) {
	tests := []struct {
		name       string
		connection *task18JetStreamPublisher
	}{
		{name: "nil ack", connection: &task18JetStreamPublisher{}},
		{name: "wrong stream", connection: &task18JetStreamPublisher{ack: &nats.PubAck{Stream: "CANARY", Sequence: 1}}},
		{name: "zero sequence", connection: &task18JetStreamPublisher{ack: &nats.PubAck{Stream: EventStreamName}}},
		{name: "provider error", connection: &task18JetStreamPublisher{err: errors.New("nats://secret@CANARY.internal")}},
		{name: "provider panic", connection: &task18JetStreamPublisher{panicPublish: true}},
		{name: "deadline", connection: &task18JetStreamPublisher{waitForContext: true}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			broker, err := NewNATSBroker(test.connection, 100*time.Millisecond)
			if err != nil {
				t.Fatal(err)
			}
			err = broker.Publish(context.Background(), contractevents.AccountStateChangedType, uuid.NewString(), []byte{1})
			if !errors.Is(err, ErrPublish) || strings.Contains(err.Error(), "CANARY") {
				t.Fatalf("JetStream provider error = %v, want fixed ErrPublish", err)
			}
		})
	}
}

func TestNATSBrokerPubAckControlsPublisherMarkOrRelease(t *testing.T) {
	now := time.Date(2026, time.August, 13, 14, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name        string
		ack         *nats.PubAck
		wantMark    int
		wantRelease int
		wantErr     error
		wantTrace   string
	}{
		{name: "durably acknowledged", ack: &nats.PubAck{Stream: EventStreamName, Sequence: 1}, wantMark: 1, wantTrace: "claim,publish,mark"},
		{name: "unacknowledged", wantRelease: 1, wantErr: ErrPublish, wantTrace: "claim,publish,release"},
	} {
		t.Run(test.name, func(t *testing.T) {
			storeFake := &task18PublisherStore{claimed: []store.TransactionalOutbox{task18PublishedRow(t, now, 1)}}
			connection := &task18JetStreamPublisher{ack: test.ack, trace: storeFake}
			broker, err := NewNATSBroker(connection, 100*time.Millisecond)
			if err != nil {
				t.Fatal(err)
			}
			publisher, err := NewPublisher(storeFake, broker, task18PublisherClock{now: now}, bytes.NewReader(make([]byte, 64)), time.Second, newTask18PublisherReporter())
			if err != nil {
				t.Fatal(err)
			}
			_, err = publisher.PublishAvailable(context.Background())
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("PublishAvailable error = %v, want %v", err, test.wantErr)
			}
			if storeFake.markCalls != test.wantMark || storeFake.releaseCalls != test.wantRelease || storeFake.trace != test.wantTrace {
				t.Fatalf("transition = mark:%d release:%d trace:%q", storeFake.markCalls, storeFake.releaseCalls, storeFake.trace)
			}
		})
	}
}

func TestNATSBrokerRejectsUnboundedConfiguration(t *testing.T) {
	connection := &task18JetStreamPublisher{ack: &nats.PubAck{Stream: EventStreamName, Sequence: 1}}
	for _, timeout := range []time.Duration{0, 99 * time.Millisecond, 10*time.Second + time.Nanosecond} {
		if broker, err := NewNATSBroker(connection, timeout); !errors.Is(err, ErrPublisherConfiguration) || broker != nil {
			t.Fatalf("NewNATSBroker timeout %s = %v, %v", timeout, broker, err)
		}
	}
}

func task18PublishedRow(t *testing.T, now time.Time, attempts int32) store.TransactionalOutbox {
	t.Helper()
	envelope := accountStateEnvelope(t)
	encoded, err := contractevents.MarshalEnvelope(envelope)
	if err != nil {
		t.Fatal(err)
	}
	return store.TransactionalOutbox{
		EventID: uuid.MustParse(envelope.EventId), EventType: envelope.EventType,
		AggregateType: envelope.AggregateType, AggregateID: uuid.MustParse(envelope.AggregateId),
		AggregateVersion: int64(envelope.AggregateVersion), IdempotencyKey: envelope.IdempotencyKey, // #nosec G115 -- the frozen test fixture uses aggregate version one.
		Payload: encoded, OccurredAt: now, AvailableAt: now, Attempts: attempts,
	}
}

type task18PublisherClock struct{ now time.Time }

func (clock task18PublisherClock) Now() time.Time { return clock.now }

type task18PublisherReporter struct {
	reports chan errorreport.Report
}

func newTask18PublisherReporter() *task18PublisherReporter {
	return &task18PublisherReporter{reports: make(chan errorreport.Report, 16)}
}

func (reporter *task18PublisherReporter) TryReport(report errorreport.Report) bool {
	select {
	case reporter.reports <- report:
		return true
	default:
		return false
	}
}

type task18PublisherStore struct {
	mu           sync.Mutex
	claimed      []store.TransactionalOutbox
	now          time.Time
	claimedUntil time.Time
	limit        int32
	availableAt  time.Time
	markCalls    int
	releaseCalls int
	trace        string
}

func (fake *task18PublisherStore) Claim(_ context.Context, now time.Time, limit int32, claimedUntil time.Time) ([]store.TransactionalOutbox, error) {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	fake.now, fake.limit, fake.claimedUntil = now, limit, claimedUntil
	fake.appendTrace("claim")
	claimed := append([]store.TransactionalOutbox(nil), fake.claimed...)
	fake.claimed = nil
	return claimed, nil
}

func (fake *task18PublisherStore) MarkPublished(context.Context, uuid.UUID, time.Time) error {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	fake.markCalls++
	fake.appendTrace("mark")
	return nil
}

func (fake *task18PublisherStore) Release(_ context.Context, _ uuid.UUID, availableAt time.Time) error {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	fake.releaseCalls++
	fake.availableAt = availableAt
	fake.appendTrace("release")
	return nil
}

func (fake *task18PublisherStore) appendTrace(value string) {
	if fake.trace != "" {
		fake.trace += ","
	}
	fake.trace += value
}

type task18Broker struct {
	mu        sync.Mutex
	calls     int
	subject   string
	messageID string
	payload   []byte
	err       error
	block     chan struct{}
	started   chan struct{}
	trace     *task18PublisherStore
}

type task18JetStreamPublisher struct {
	message        *nats.Msg
	publishContext context.Context
	ack            *nats.PubAck
	err            error
	panicPublish   bool
	waitForContext bool
	calls          int
	trace          *task18PublisherStore
}

func (connection *task18JetStreamPublisher) PublishMsg(message *nats.Msg, options ...nats.PubOpt) (*nats.PubAck, error) {
	if connection.panicPublish {
		panic("CANARY NATS publish panic")
	}
	connection.calls++
	connection.message = &nats.Msg{Subject: message.Subject, Header: make(nats.Header), Data: bytes.Clone(message.Data)}
	for key, values := range message.Header {
		connection.message.Header[key] = append([]string(nil), values...)
	}
	if len(options) == 1 {
		if option, ok := options[0].(nats.ContextOpt); ok {
			connection.publishContext = option.Context
		}
	}
	if connection.trace != nil {
		connection.trace.mu.Lock()
		connection.trace.appendTrace("publish")
		connection.trace.mu.Unlock()
	}
	if connection.waitForContext && connection.publishContext != nil {
		<-connection.publishContext.Done()
		return nil, connection.publishContext.Err()
	}
	return connection.ack, connection.err
}

func (broker *task18Broker) Publish(ctx context.Context, subject, messageID string, payload []byte) error {
	broker.mu.Lock()
	broker.calls++
	broker.subject, broker.messageID, broker.payload = subject, messageID, bytes.Clone(payload)
	if broker.started != nil && broker.calls == 1 {
		close(broker.started)
	}
	broker.mu.Unlock()
	if broker.trace != nil {
		broker.trace.mu.Lock()
		broker.trace.appendTrace("publish")
		broker.trace.mu.Unlock()
	}
	if broker.block != nil {
		select {
		case <-broker.block:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return broker.err
}
