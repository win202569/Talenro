// Package outbox provides transaction-bound event persistence and consumption.
package outbox

import (
	"context"
	"errors"
	"math"

	"github.com/google/uuid"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	contractevents "talenro.local/platform/internal/contracts/events"
	"talenro.local/platform/internal/store"
)

var (
	// ErrInvalidEvent reports invalid event input without retaining its values.
	ErrInvalidEvent = errors.New("outbox: invalid event")
	// ErrStore reports a transaction-bound persistence failure.
	ErrStore = errors.New("outbox: store failed")
)

// Repository appends validated events through one caller-owned transaction.
type Repository interface {
	Append(context.Context, *eventsv1.EventEnvelope) error
}

type outboxStore interface {
	InsertOutboxEvent(context.Context, store.InsertOutboxEventParams) error
}

type repository struct{ store outboxStore }

var _ Repository = (*repository)(nil)

// NewRepository binds all appends to the supplied caller-owned DBTX.
func NewRepository(db store.DBTX) (Repository, error) {
	if db == nil {
		return nil, ErrInvalidEvent
	}
	return newRepositoryWithStore(store.New(db)), nil
}

func newRepositoryWithStore(boundStore outboxStore) *repository {
	return &repository{store: boundStore}
}

func (repository *repository) Append(ctx context.Context, envelope *eventsv1.EventEnvelope) error {
	if ctx == nil || repository == nil || repository.store == nil {
		return ErrInvalidEvent
	}
	encoded, err := contractevents.MarshalEnvelope(envelope)
	if err != nil {
		return ErrInvalidEvent
	}
	eventID, err := uuid.Parse(envelope.GetEventId())
	if err != nil || eventID == uuid.Nil || eventID.String() != envelope.GetEventId() {
		return ErrInvalidEvent
	}
	aggregateID, err := uuid.Parse(envelope.GetAggregateId())
	if err != nil || aggregateID == uuid.Nil || aggregateID.String() != envelope.GetAggregateId() || envelope.GetAggregateVersion() > math.MaxInt64 {
		return ErrInvalidEvent
	}
	occurredAt := envelope.GetOccurredAt().AsTime()
	params := store.InsertOutboxEventParams{
		EventID:          eventID,
		EventType:        envelope.GetEventType(),
		AggregateType:    envelope.GetAggregateType(),
		AggregateID:      aggregateID,
		AggregateVersion: int64(envelope.GetAggregateVersion()), // #nosec G115 -- MarshalEnvelope rejects values above MaxInt64.
		IdempotencyKey:   envelope.GetIdempotencyKey(),
		Payload:          append([]byte(nil), encoded...),
		OccurredAt:       occurredAt,
		AvailableAt:      occurredAt,
	}
	if err := repository.store.InsertOutboxEvent(ctx, params); err != nil {
		return ErrStore
	}
	return nil
}
