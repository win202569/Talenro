package identity

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"
	identityv1 "talenro.local/platform/gen/go/talenro/identity/v1"
	contractevents "talenro.local/platform/internal/contracts/events"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
	"talenro.local/platform/internal/strictjson"
)

const (
	emailConsumerName      = "identity.email-delivery.v1"
	emailConsumedRetention = 30 * 24 * time.Hour
	emailRollbackTimeout   = 2 * time.Second
)

var (
	// ErrEmailConsumerConfiguration reports unusable delivery dependencies.
	ErrEmailConsumerConfiguration = errors.New("identity: invalid email consumer configuration")
	// ErrEmailConsumerEvent reports malformed or non-email delivery events.
	ErrEmailConsumerEvent = errors.New("identity: invalid email delivery event")
	// ErrEmailConsumerStore reports a fixed generated-store failure.
	ErrEmailConsumerStore = errors.New("identity: email consumer store failed")
	// ErrEmailConsumerProvider reports a fixed email-provider failure.
	ErrEmailConsumerProvider = errors.New("identity: email provider failed")
	// ErrEmailConsumerCanceled reports bounded cancellation.
	ErrEmailConsumerCanceled = errors.New("identity: email consumer canceled")
)

// EmailPendingDelivery is the protected adapter value loaded before provider work.
type EmailPendingDelivery struct {
	DeliveryID uuid.UUID
	TemplateID TemplateID
	Protected  sensitive.EncryptedField
	ExpiresAt  time.Time
}

// EmailDeliveryRepository owns only the delivery worker's PostgreSQL transitions.
type EmailDeliveryRepository interface {
	HasConsumed(context.Context, string, uuid.UUID) (bool, error)
	LoadPending(context.Context, uuid.UUID, TemplateID) (EmailPendingDelivery, bool, error)
	WithinTransaction(context.Context, func(context.Context, EmailDeliveryTransaction) error) error
}

// EmailDeliveryTransaction atomically records acceptance and clears the exact pending ciphertext.
type EmailDeliveryTransaction interface {
	RecordConsumed(context.Context, string, uuid.UUID, time.Time, time.Time) (bool, error)
	ClearPending(context.Context, TemplateID, uuid.UUID, time.Time) (int64, error)
}

// EmailConsumer validates, opens, sends, and records one delivery event.
type EmailConsumer struct {
	repository      EmailDeliveryRepository
	protector       sensitive.Protector
	sender          EmailSender
	providerTimeout time.Duration
}

// NewEmailConsumer constructs the provider worker boundary.
func NewEmailConsumer(repository EmailDeliveryRepository, protector sensitive.Protector, sender EmailSender, providerTimeout time.Duration) (*EmailConsumer, error) {
	if nilIdentityValue(repository) || nilIdentityValue(protector) || nilIdentityValue(sender) ||
		providerTimeout < 100*time.Millisecond || providerTimeout > 10*time.Second {
		return nil, ErrEmailConsumerConfiguration
	}
	return &EmailConsumer{repository: repository, protector: protector, sender: sender, providerTimeout: providerTimeout}, nil
}

// Consume processes one raw, deterministic EmailDeliveryRequested envelope.
func (consumer *EmailConsumer) Consume(ctx context.Context, encoded []byte, consumedAt time.Time) error {
	if consumer == nil || nilIdentityValue(ctx) || ctx.Err() != nil {
		return ErrEmailConsumerCanceled
	}
	eventID, requested, err := decodeEmailDeliveryEvent(encoded)
	if err != nil {
		return ErrEmailConsumerEvent
	}
	if !validConsumerTime(consumedAt) {
		return ErrEmailConsumerEvent
	}
	deliveryID, err := uuid.Parse(requested.GetDeliveryId())
	template := TemplateID(requested.GetTemplateId())
	if err != nil || deliveryID == uuid.Nil || deliveryID.String() != requested.GetDeliveryId() || !validTemplateID(template) ||
		!validIdentityLocale(requested.GetLocale()) {
		return ErrEmailConsumerEvent
	}
	consumed, err := consumer.repository.HasConsumed(ctx, emailConsumerName, eventID)
	if err != nil {
		return emailConsumerContextOr(ctx, ErrEmailConsumerStore)
	}
	if consumed {
		return nil
	}
	pending, found, err := consumer.repository.LoadPending(ctx, deliveryID, template)
	if err != nil {
		return emailConsumerContextOr(ctx, ErrEmailConsumerStore)
	}
	if !found || !validPendingDelivery(pending, deliveryID, template, consumedAt) {
		clear(pending.Protected.Ciphertext)
		return ErrEmailConsumerEvent
	}
	domain := emailVerificationDeliveryDomain
	if template == ResetPasswordTemplate {
		domain = passwordResetDeliveryDomain
	}
	plaintext, err := consumer.protector.Decrypt(domain, pending.Protected)
	clear(pending.Protected.Ciphertext)
	if err != nil || len(plaintext) == 0 || len(plaintext) > maximumPendingDeliveryBytes {
		clear(plaintext)
		return ErrEmailConsumerEvent
	}
	delivery, err := decodePendingDelivery(plaintext, deliveryID, template, requested.GetLocale())
	clear(plaintext)
	if err != nil {
		return ErrEmailConsumerEvent
	}
	defer clearDelivery(&delivery)
	providerCtx, cancel := context.WithTimeout(ctx, consumer.providerTimeout)
	providerFailed := emailProviderFailed(providerCtx, consumer.sender, delivery)
	cancel()
	if providerFailed {
		return emailConsumerContextOr(ctx, ErrEmailConsumerProvider)
	}
	expiresAt := consumedAt.Add(emailConsumedRetention)
	if !expiresAt.After(consumedAt) || expiresAt.Year() > 2101 {
		return ErrEmailConsumerEvent
	}
	err = consumer.repository.WithinTransaction(ctx, func(transactionContext context.Context, transaction EmailDeliveryTransaction) error {
		if nilIdentityValue(transaction) {
			return ErrEmailConsumerStore
		}
		winner, recordErr := transaction.RecordConsumed(transactionContext, emailConsumerName, eventID, consumedAt, expiresAt)
		if recordErr != nil {
			return ErrEmailConsumerStore
		}
		if !winner {
			return nil
		}
		rows, clearErr := transaction.ClearPending(transactionContext, template, deliveryID, consumedAt)
		if clearErr != nil || rows < 0 || rows > 1 {
			return ErrEmailConsumerStore
		}
		return nil
	})
	if err != nil {
		return emailConsumerContextOr(ctx, ErrEmailConsumerStore)
	}
	return nil
}

type pendingDeliveryWire struct {
	Recipient string `json:"recipient"`
	Token     string `json:"token"`
	Template  string `json:"template"`
	Locale    string `json:"locale"`
}

func decodePendingDelivery(encoded []byte, deliveryID uuid.UUID, template TemplateID, locale string) (Delivery, error) {
	var wire pendingDeliveryWire
	if err := strictjson.Decode(bytes.NewReader(encoded), maximumPendingDeliveryBytes, &wire); err != nil ||
		wire.Template != string(template) || wire.Locale != locale {
		return Delivery{}, ErrEmailConsumerEvent
	}
	canonical, err := CanonicalizeEmail(wire.Recipient)
	if err != nil {
		return Delivery{}, ErrEmailConsumerEvent
	}
	canonicalBytes := canonical.Bytes()
	if string(canonicalBytes) != wire.Recipient {
		clear(canonicalBytes)
		return Delivery{}, ErrEmailConsumerEvent
	}
	clear(canonicalBytes)
	token, err := securitykit.DecodeOpaqueToken(wire.Token)
	if err != nil {
		return Delivery{}, ErrEmailConsumerEvent
	}
	return Delivery{DeliveryID: deliveryID, TemplateID: template, Locale: locale, Recipient: wire.Recipient, OneTimeToken: token}, nil
}

func decodeEmailDeliveryEvent(encoded []byte) (uuid.UUID, *identityv1.EmailDeliveryRequested, error) {
	envelope, err := contractevents.UnmarshalEnvelope(encoded)
	if err != nil || envelope.GetEventType() != contractevents.EmailDeliveryRequestedType {
		return uuid.Nil, nil, ErrEmailConsumerEvent
	}
	eventID, err := uuid.Parse(envelope.GetEventId())
	if err != nil || eventID == uuid.Nil || eventID.String() != envelope.GetEventId() {
		return uuid.Nil, nil, ErrEmailConsumerEvent
	}
	payload := new(identityv1.EmailDeliveryRequested)
	if err := (proto.UnmarshalOptions{DiscardUnknown: false}).Unmarshal(envelope.GetPayload(), payload); err != nil || len(payload.ProtoReflect().GetUnknown()) != 0 {
		return uuid.Nil, nil, ErrEmailConsumerEvent
	}
	canonical, err := contractevents.MarshalPayload(contractevents.EmailDeliveryRequestedType, payload)
	if err != nil || !bytes.Equal(canonical, envelope.GetPayload()) || payload.GetDeliveryId() != envelope.GetAggregateId() {
		return uuid.Nil, nil, ErrEmailConsumerEvent
	}
	return eventID, payload, nil
}

func validPendingDelivery(pending EmailPendingDelivery, deliveryID uuid.UUID, template TemplateID, now time.Time) bool {
	return pending.DeliveryID == deliveryID && pending.TemplateID == template && pending.Protected.KeyVersion > 0 &&
		len(pending.Protected.Ciphertext) >= 29 && len(pending.Protected.Ciphertext) <= maximumPendingDeliveryBytes &&
		validConsumerTime(pending.ExpiresAt) && now.Before(pending.ExpiresAt)
}

func validConsumerTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Year() >= 2020 && value.Year() <= 2100
}

func emailProviderFailed(ctx context.Context, sender EmailSender, delivery Delivery) (failed bool) {
	defer func() {
		if recover() != nil {
			failed = true
		}
	}()
	return sender.Send(ctx, delivery) != nil
}

func clearDelivery(delivery *Delivery) {
	if delivery == nil {
		return
	}
	delivery.OneTimeToken.Clear()
	delivery.Recipient = ""
}

func emailConsumerContextOr(ctx context.Context, fallback error) error {
	if ctx != nil && ctx.Err() != nil {
		return ErrEmailConsumerCanceled
	}
	return fallback
}

// EmailDeliveryDatabase combines generated query reads with transaction begin.
type EmailDeliveryDatabase interface {
	store.DBTX
	Begin(context.Context) (pgx.Tx, error)
}

// PostgresEmailDeliveryRepository is the generated-query-only worker adapter.
type PostgresEmailDeliveryRepository struct{ database EmailDeliveryDatabase }

// NewPostgresEmailDeliveryRepository binds email delivery to PostgreSQL.
func NewPostgresEmailDeliveryRepository(database EmailDeliveryDatabase) (*PostgresEmailDeliveryRepository, error) {
	if nilIdentityValue(database) {
		return nil, ErrEmailConsumerConfiguration
	}
	return &PostgresEmailDeliveryRepository{database: database}, nil
}

// HasConsumed performs the pre-provider read-only fast path.
func (repository *PostgresEmailDeliveryRepository) HasConsumed(ctx context.Context, consumer string, eventID uuid.UUID) (bool, error) {
	if repository == nil || nilIdentityValue(repository.database) {
		return false, ErrEmailConsumerStore
	}
	found, err := store.New(repository.database).HasConsumedEvent(ctx, store.HasConsumedEventParams{Consumer: consumer, EventID: eventID})
	if err != nil {
		return false, ErrEmailConsumerStore
	}
	return found, nil
}

// LoadPending reads only the exact finite template's encrypted delivery.
func (repository *PostgresEmailDeliveryRepository) LoadPending(ctx context.Context, deliveryID uuid.UUID, template TemplateID) (EmailPendingDelivery, bool, error) {
	if repository == nil || nilIdentityValue(repository.database) {
		return EmailPendingDelivery{}, false, ErrEmailConsumerStore
	}
	queries := store.New(repository.database)
	switch template {
	case VerifyEmailTemplate:
		row, err := queries.GetPendingEmailDelivery(ctx, uuid.NullUUID{UUID: deliveryID, Valid: true})
		if errors.Is(err, pgx.ErrNoRows) {
			return EmailPendingDelivery{}, false, nil
		}
		if err != nil || !row.VerificationDeliveryID.Valid || !row.VerificationDeliveryKeyVersion.Valid ||
			row.VerificationDeliveryKeyVersion.Int32 <= 0 {
			return EmailPendingDelivery{}, false, ErrEmailConsumerStore
		}
		return EmailPendingDelivery{
			DeliveryID: row.VerificationDeliveryID.UUID, TemplateID: template,
			Protected: sensitive.EncryptedField{KeyVersion: uint32(row.VerificationDeliveryKeyVersion.Int32), Ciphertext: bytes.Clone(row.VerificationDeliveryCiphertext)}, // #nosec G115 -- positive int32 is losslessly representable as uint32.
			ExpiresAt: row.VerificationExpiresAt.Time.UTC(),
		}, true, nil
	case ResetPasswordTemplate:
		row, err := queries.GetPendingPasswordResetDelivery(ctx, uuid.NullUUID{UUID: deliveryID, Valid: true})
		if errors.Is(err, pgx.ErrNoRows) {
			return EmailPendingDelivery{}, false, nil
		}
		if err != nil || !row.ResetDeliveryID.Valid || !row.ResetDeliveryKeyVersion.Valid || row.ResetDeliveryKeyVersion.Int32 <= 0 {
			return EmailPendingDelivery{}, false, ErrEmailConsumerStore
		}
		return EmailPendingDelivery{
			DeliveryID: row.ResetDeliveryID.UUID, TemplateID: template,
			Protected: sensitive.EncryptedField{KeyVersion: uint32(row.ResetDeliveryKeyVersion.Int32), Ciphertext: bytes.Clone(row.ResetDeliveryCiphertext)}, // #nosec G115 -- positive int32 is losslessly representable as uint32.
			ExpiresAt: row.ResetExpiresAt.Time.UTC(),
		}, true, nil
	default:
		return EmailPendingDelivery{}, false, ErrEmailConsumerEvent
	}
}

// WithinTransaction commits record+clear together and independently rolls back.
func (repository *PostgresEmailDeliveryRepository) WithinTransaction(ctx context.Context, operation func(context.Context, EmailDeliveryTransaction) error) (resultErr error) {
	if repository == nil || nilIdentityValue(repository.database) || operation == nil {
		return ErrEmailConsumerStore
	}
	transaction, err := repository.database.Begin(ctx)
	if err != nil || nilIdentityValue(transaction) {
		return ErrEmailConsumerStore
	}
	defer func() {
		if recover() != nil {
			resultErr = ErrEmailConsumerStore
		}
		if transaction != nil {
			rollbackEmailDelivery(ctx, transaction)
		}
	}()
	bound := &postgresEmailDeliveryTransaction{queries: store.New(transaction)}
	if err := operation(ctx, bound); err != nil {
		return err
	}
	if err := transaction.Commit(ctx); err != nil {
		return ErrEmailConsumerStore
	}
	transaction = nil
	return nil
}

type postgresEmailDeliveryTransaction struct{ queries *store.Queries }

func (transaction *postgresEmailDeliveryTransaction) RecordConsumed(ctx context.Context, consumer string, eventID uuid.UUID, consumedAt, expiresAt time.Time) (bool, error) {
	if transaction == nil || transaction.queries == nil {
		return false, ErrEmailConsumerStore
	}
	rows, err := transaction.queries.RecordConsumedEvent(ctx, store.RecordConsumedEventParams{
		Consumer: consumer, EventID: eventID, ConsumedAt: consumedAt, ExpiresAt: expiresAt,
	})
	if err != nil || rows < 0 || rows > 1 {
		return false, ErrEmailConsumerStore
	}
	return rows == 1, nil
}

func (transaction *postgresEmailDeliveryTransaction) ClearPending(ctx context.Context, template TemplateID, deliveryID uuid.UUID, updatedAt time.Time) (int64, error) {
	if transaction == nil || transaction.queries == nil {
		return 0, ErrEmailConsumerStore
	}
	switch template {
	case VerifyEmailTemplate:
		return transaction.queries.ClearPendingEmailDelivery(ctx, store.ClearPendingEmailDeliveryParams{
			VerificationDeliveryID: uuid.NullUUID{UUID: deliveryID, Valid: true}, UpdatedAt: updatedAt,
		})
	case ResetPasswordTemplate:
		return transaction.queries.ClearPendingPasswordResetDelivery(ctx, store.ClearPendingPasswordResetDeliveryParams{
			ResetDeliveryID: uuid.NullUUID{UUID: deliveryID, Valid: true}, UpdatedAt: updatedAt,
		})
	default:
		return 0, ErrEmailConsumerEvent
	}
}

func rollbackEmailDelivery(operationContext context.Context, transaction pgx.Tx) {
	if operationContext == nil {
		return
	}
	rollbackContext, cancel := context.WithTimeout(context.WithoutCancel(operationContext), emailRollbackTimeout)
	defer cancel()
	_ = transaction.Rollback(rollbackContext)
}
