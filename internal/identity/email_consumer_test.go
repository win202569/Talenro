package identity

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	identityv1 "talenro.local/platform/gen/go/talenro/identity/v1"
	contractevents "talenro.local/platform/internal/contracts/events"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/sensitive"
)

func TestEmailConsumerSendsOutsideTransactionThenAtomicallyRecordsAndClears(t *testing.T) {
	for _, template := range []TemplateID{VerifyEmailTemplate, ResetPasswordTemplate} {
		t.Run(string(template), func(t *testing.T) {
			fixture := newTask18EmailConsumerFixture(t, template)
			if err := fixture.consumer.Consume(context.Background(), fixture.envelope, fixture.now); err != nil {
				t.Fatal(err)
			}
			if got := fixture.repository.traceString(); got != "has,load,send,begin,record,clear:"+string(template)+",commit" {
				t.Fatalf("consumer trace = %q", got)
			}
			if fixture.sender.calledInTransaction {
				t.Fatal("external email provider was called while a transaction was open")
			}
			if len(fixture.sender.deliveries) != 1 {
				t.Fatalf("provider deliveries = %d, want 1", len(fixture.sender.deliveries))
			}
			delivery := fixture.sender.deliveries[0]
			if delivery.DeliveryID != fixture.deliveryID || delivery.TemplateID != template ||
				delivery.Locale != "en" || delivery.Recipient != "person@example.com" ||
				!bytes.Equal(delivery.OneTimeToken.Copy(), bytes.Repeat([]byte{0x44}, 32)) {
				t.Fatal("consumer did not reconstruct the exact private delivery")
			}
			if fixture.repository.recordEventID != fixture.eventID || fixture.repository.clearDeliveryID != fixture.deliveryID {
				t.Fatal("database transition did not bind the exact event and delivery IDs")
			}
		})
	}
}

func TestEmailConsumerSkipsProviderWhenEventAlreadyConsumed(t *testing.T) {
	fixture := newTask18EmailConsumerFixture(t, VerifyEmailTemplate)
	fixture.repository.consumed = true
	if err := fixture.consumer.Consume(context.Background(), fixture.envelope, fixture.now); err != nil {
		t.Fatal(err)
	}
	if got := fixture.repository.traceString(); got != "has" {
		t.Fatalf("consumed fast-path trace = %q, want has", got)
	}
	if len(fixture.sender.deliveries) != 0 {
		t.Fatal("already-consumed event reached the provider")
	}
}

func TestEmailConsumerProviderFailureDoesNotOpenTransactionOrLeakError(t *testing.T) {
	fixture := newTask18EmailConsumerFixture(t, VerifyEmailTemplate)
	fixture.sender.err = errors.New("smtp://secret@CANARY.internal/provider-body")
	err := fixture.consumer.Consume(context.Background(), fixture.envelope, fixture.now)
	if !errors.Is(err, ErrEmailConsumerProvider) || strings.Contains(err.Error(), "CANARY") {
		t.Fatalf("consumer error = %v, want fixed provider category", err)
	}
	if fixture.repository.beginCalls != 0 || fixture.repository.recordCalls != 0 || fixture.repository.clearCalls != 0 {
		t.Fatal("provider failure mutated account authority")
	}
}

func TestEmailConsumerCommitAmbiguityRetriesSameProviderIdempotencyKey(t *testing.T) {
	fixture := newTask18EmailConsumerFixture(t, ResetPasswordTemplate)
	fixture.repository.commitFailures = 1
	if err := fixture.consumer.Consume(context.Background(), fixture.envelope, fixture.now); !errors.Is(err, ErrEmailConsumerStore) {
		t.Fatalf("first consume error = %v, want fixed store category", err)
	}
	if err := fixture.consumer.Consume(context.Background(), fixture.envelope, fixture.now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if len(fixture.sender.deliveries) != 2 || fixture.sender.deliveries[0].DeliveryID != fixture.deliveryID ||
		fixture.sender.deliveries[1].DeliveryID != fixture.deliveryID {
		t.Fatal("ambiguous commit retry changed the provider idempotency key")
	}
	if fixture.sender.effects != 1 {
		t.Fatalf("provider effects = %d, want one idempotent effect", fixture.sender.effects)
	}
}

func TestEmailConsumerConcurrentAcceptanceHasOneDatabaseWinnerAndExactClear(t *testing.T) {
	fixture := newTask18EmailConsumerFixture(t, VerifyEmailTemplate)
	fixture.repository.recordBarrier = make(chan struct{})
	errorsOut := make(chan error, 2)
	for range 2 {
		go func() { errorsOut <- fixture.consumer.Consume(context.Background(), fixture.envelope, fixture.now) }()
	}
	deadline := time.Now().Add(time.Second)
	for fixture.repository.recordCallCount() < 2 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	close(fixture.repository.recordBarrier)
	for range 2 {
		if err := <-errorsOut; err != nil {
			t.Fatal(err)
		}
	}
	if fixture.repository.recordWinners != 1 || fixture.repository.clearCalls != 1 {
		t.Fatalf("database winners=%d clears=%d, want 1 each", fixture.repository.recordWinners, fixture.repository.clearCalls)
	}
	if fixture.sender.effects != 1 || len(fixture.sender.deliveries) != 2 {
		t.Fatalf("provider calls=%d effects=%d, want 2 calls and 1 effect", len(fixture.sender.deliveries), fixture.sender.effects)
	}
}

func TestEmailConsumerRejectsMalformedOrWrongEventBeforeDependencies(t *testing.T) {
	fixture := newTask18EmailConsumerFixture(t, VerifyEmailTemplate)
	for _, encoded := range [][]byte{
		[]byte("CANARY"),
		task18WrongEmailEnvelope(t, fixture.now),
	} {
		err := fixture.consumer.Consume(context.Background(), encoded, fixture.now)
		if !errors.Is(err, ErrEmailConsumerEvent) || strings.Contains(err.Error(), "CANARY") {
			t.Fatalf("invalid event error = %v", err)
		}
	}
	if fixture.repository.hasCalls != 0 || len(fixture.sender.deliveries) != 0 {
		t.Fatal("invalid event reached repository or provider")
	}
}

type task18EmailConsumerFixture struct {
	consumer   *EmailConsumer
	repository *task18EmailRepository
	sender     *task18EmailSender
	envelope   []byte
	deliveryID uuid.UUID
	eventID    uuid.UUID
	now        time.Time
}

func newTask18EmailConsumerFixture(t *testing.T, template TemplateID) task18EmailConsumerFixture {
	t.Helper()
	now := time.Date(2026, time.August, 13, 15, 0, 0, 0, time.UTC)
	deliveryID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	eventID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	lookupKey := secret.NewBytes(bytes.Repeat([]byte{0x11}, 32))
	encryptionKey := secret.NewBytes(bytes.Repeat([]byte{0x22}, 32))
	protector, err := sensitive.NewLocal(lookupKey, encryptionKey, 1)
	if err != nil {
		t.Fatal(err)
	}
	token := bytes.Repeat([]byte{0x44}, 32)
	plaintext := pendingDeliveryJSON([]byte("person@example.com"), token, template, "en")
	domain := emailVerificationDeliveryDomain
	if template == ResetPasswordTemplate {
		domain = passwordResetDeliveryDomain
	}
	protected, err := protector.Encrypt(domain, plaintext)
	clear(plaintext)
	if err != nil {
		t.Fatal(err)
	}
	repository := &task18EmailRepository{
		pending: EmailPendingDelivery{DeliveryID: deliveryID, TemplateID: template, Protected: protected, ExpiresAt: now.Add(time.Hour)},
	}
	sender := &task18EmailSender{repository: repository, accepted: make(map[uuid.UUID]struct{})}
	consumer, err := NewEmailConsumer(repository, protector, sender, 200*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	return task18EmailConsumerFixture{
		consumer: consumer, repository: repository, sender: sender,
		envelope:   task18EmailEnvelope(t, eventID, deliveryID, template, now),
		deliveryID: deliveryID, eventID: eventID, now: now,
	}
}

func task18EmailEnvelope(t *testing.T, eventID, deliveryID uuid.UUID, template TemplateID, now time.Time) []byte {
	t.Helper()
	payload, err := contractevents.MarshalPayload(contractevents.EmailDeliveryRequestedType, &identityv1.EmailDeliveryRequested{
		DeliveryId: deliveryID.String(), TemplateId: string(template), Locale: "en",
	})
	if err != nil {
		t.Fatal(err)
	}
	envelope := &eventsv1.EventEnvelope{
		EventId: eventID.String(), EventType: contractevents.EmailDeliveryRequestedType,
		OccurredAt: timestamppb.New(now), Producer: "identity", AggregateType: "email_delivery",
		AggregateId: deliveryID.String(), AggregateVersion: 1,
		IdempotencyKey: "email-delivery:test", Payload: payload,
	}
	encoded, err := contractevents.MarshalEnvelope(envelope)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func task18WrongEmailEnvelope(t *testing.T, now time.Time) []byte {
	t.Helper()
	payload, err := contractevents.MarshalPayload(contractevents.AccountStateChangedType, &identityv1.AccountStateChanged{
		PrincipalId: uuid.NewString(), State: "active", Version: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	envelope := &eventsv1.EventEnvelope{
		EventId: uuid.NewString(), EventType: contractevents.AccountStateChangedType,
		OccurredAt: timestamppb.New(now), Producer: "identity", AggregateType: "account",
		AggregateId: uuid.NewString(), AggregateVersion: 1,
		IdempotencyKey: "account:test", Payload: payload,
	}
	envelope.AggregateId = (&identityv1.AccountStateChanged{}).PrincipalId
	// Rebuild with a matching aggregate ID so the envelope itself is valid.
	principalID := uuid.NewString()
	payload, err = contractevents.MarshalPayload(contractevents.AccountStateChangedType, &identityv1.AccountStateChanged{
		PrincipalId: principalID, State: "active", Version: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	envelope.AggregateId, envelope.Payload = principalID, payload
	encoded, err := contractevents.MarshalEnvelope(envelope)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

type task18EmailRepository struct {
	mu              sync.Mutex
	trace           []string
	pending         EmailPendingDelivery
	consumed        bool
	hasCalls        int
	beginCalls      int
	recordCalls     int
	recordWinners   int
	clearCalls      int
	commitFailures  int
	recordBarrier   chan struct{}
	inTransaction   bool
	recordEventID   uuid.UUID
	clearDeliveryID uuid.UUID
}

func (repository *task18EmailRepository) HasConsumed(context.Context, string, uuid.UUID) (bool, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.hasCalls++
	repository.trace = append(repository.trace, "has")
	return repository.consumed, nil
}

func (repository *task18EmailRepository) LoadPending(_ context.Context, deliveryID uuid.UUID, template TemplateID) (EmailPendingDelivery, bool, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.trace = append(repository.trace, "load")
	if deliveryID != repository.pending.DeliveryID || template != repository.pending.TemplateID {
		return EmailPendingDelivery{}, false, nil
	}
	result := repository.pending
	result.Protected.Ciphertext = bytes.Clone(result.Protected.Ciphertext)
	return result, true, nil
}

func (repository *task18EmailRepository) WithinTransaction(ctx context.Context, operation func(context.Context, EmailDeliveryTransaction) error) error {
	repository.mu.Lock()
	consumedBefore := repository.consumed
	winnersBefore := repository.recordWinners
	clearsBefore := repository.clearCalls
	repository.beginCalls++
	repository.inTransaction = true
	repository.trace = append(repository.trace, "begin")
	repository.mu.Unlock()
	err := operation(ctx, repository)
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.inTransaction = false
	if err != nil {
		return err
	}
	if repository.commitFailures > 0 {
		repository.commitFailures--
		repository.consumed = consumedBefore
		repository.recordWinners = winnersBefore
		repository.clearCalls = clearsBefore
		return errors.New("CANARY commit failure")
	}
	repository.trace = append(repository.trace, "commit")
	return nil
}

func (repository *task18EmailRepository) RecordConsumed(_ context.Context, _ string, eventID uuid.UUID, _, _ time.Time) (bool, error) {
	repository.mu.Lock()
	repository.recordCalls++
	repository.recordEventID = eventID
	repository.trace = append(repository.trace, "record")
	barrier := repository.recordBarrier
	repository.mu.Unlock()
	if barrier != nil {
		<-barrier
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.consumed {
		return false, nil
	}
	repository.consumed = true
	repository.recordWinners++
	return true, nil
}

func (repository *task18EmailRepository) ClearPending(_ context.Context, template TemplateID, deliveryID uuid.UUID, _ time.Time) (int64, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.clearCalls++
	repository.clearDeliveryID = deliveryID
	repository.trace = append(repository.trace, "clear:"+string(template))
	if deliveryID != repository.pending.DeliveryID || template != repository.pending.TemplateID {
		return 0, nil
	}
	return 1, nil
}

func (repository *task18EmailRepository) traceString() string {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	return strings.Join(repository.trace, ",")
}

func (repository *task18EmailRepository) recordCallCount() int {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	return repository.recordCalls
}

type task18EmailSender struct {
	mu                  sync.Mutex
	repository          *task18EmailRepository
	deliveries          []Delivery
	accepted            map[uuid.UUID]struct{}
	effects             int
	calledInTransaction bool
	err                 error
}

func (sender *task18EmailSender) Send(_ context.Context, delivery Delivery) error {
	sender.mu.Lock()
	defer sender.mu.Unlock()
	sender.repository.mu.Lock()
	inTransaction := sender.repository.inTransaction
	sender.repository.trace = append(sender.repository.trace, "send")
	sender.repository.mu.Unlock()
	sender.calledInTransaction = sender.calledInTransaction || inTransaction
	token := delivery.OneTimeToken.Copy()
	sender.deliveries = append(sender.deliveries, Delivery{
		DeliveryID: delivery.DeliveryID, TemplateID: delivery.TemplateID, Locale: delivery.Locale,
		Recipient: delivery.Recipient, OneTimeToken: secret.NewBytes(token),
	})
	clear(token)
	if sender.err != nil {
		return sender.err
	}
	if _, exists := sender.accepted[delivery.DeliveryID]; !exists {
		sender.accepted[delivery.DeliveryID] = struct{}{}
		sender.effects++
	}
	return nil
}
