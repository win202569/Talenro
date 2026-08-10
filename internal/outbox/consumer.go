package outbox

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	contractevents "talenro.local/platform/internal/contracts/events"
	"talenro.local/platform/internal/store"
)

const (
	maximumConsumerRetention = 365 * 24 * time.Hour
	rollbackTimeout          = 2 * time.Second
)

var (
	// ErrInvalidConfiguration reports invalid consumer construction parameters.
	ErrInvalidConfiguration = errors.New("outbox: invalid configuration")
	// ErrHandler reports a database-only handler failure without wrapping details.
	ErrHandler = errors.New("outbox: handler failed")
	// ErrCommit reports an ambiguous or failed transaction commit.
	ErrCommit = errors.New("outbox: commit failed")
	// ErrCanceled reports cancellation without wrapping caller details.
	ErrCanceled = errors.New("outbox: canceled")
)

// Handler performs database-only side effects through the supplied transaction.
// External provider calls are intentionally outside this contract.
type Handler interface {
	HandleInTransaction(context.Context, store.DBTX, *eventsv1.EventEnvelope) error
}

// HandlerFunc adapts a function to Handler.
type HandlerFunc func(context.Context, store.DBTX, *eventsv1.EventEnvelope) error

// HandleInTransaction invokes the adapted database-only handler.
func (function HandlerFunc) HandleInTransaction(ctx context.Context, tx store.DBTX, envelope *eventsv1.EventEnvelope) error {
	return function(ctx, tx, envelope)
}

// PGXBeginner is the narrow production transaction source accepted by NewConsumer.
type PGXBeginner interface {
	Begin(context.Context) (pgx.Tx, error)
}

type transaction interface {
	store.DBTX
	Commit(context.Context) error
	Rollback(context.Context) error
}

type transactionBeginner interface {
	Begin(context.Context) (transaction, error)
}

type consumerStore interface {
	RecordConsumedEvent(context.Context, store.RecordConsumedEventParams) (int64, error)
}

type consumer struct {
	beginner     transactionBeginner
	storeFactory func(store.DBTX) consumerStore
	name         string
	retention    time.Duration
	handler      Handler
}

type pgxBeginnerAdapter struct{ beginner PGXBeginner }

func (adapter pgxBeginnerAdapter) Begin(ctx context.Context) (transaction, error) {
	tx, err := adapter.beginner.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// Consumer validates delivery bytes and commits deduplicated database side effects.
type Consumer struct{ implementation *consumer }

// NewConsumer creates a DB-only consumer over a pgx-compatible transaction source.
func NewConsumer(beginner PGXBeginner, name string, retention time.Duration, handler Handler) (*Consumer, error) {
	if beginner == nil {
		return nil, ErrInvalidConfiguration
	}
	implementation, err := newConsumerWithBeginner(pgxBeginnerAdapter{beginner: beginner}, func(tx store.DBTX) consumerStore {
		return store.New(tx)
	}, name, retention, handler)
	if err != nil {
		return nil, err
	}
	return &Consumer{implementation: implementation}, nil
}

func newConsumerWithBeginner(beginner transactionBeginner, storeFactory func(store.DBTX) consumerStore, name string, retention time.Duration, handler Handler) (*consumer, error) {
	if beginner == nil || storeFactory == nil || handler == nil || !validConsumerName(name) || retention <= 0 || retention > maximumConsumerRetention {
		return nil, ErrInvalidConfiguration
	}
	return &consumer{beginner: beginner, storeFactory: storeFactory, name: name, retention: retention, handler: handler}, nil
}

// Consume validates and decodes before beginning a transaction. It never acknowledges a broker delivery.
func (public *Consumer) Consume(ctx context.Context, encoded []byte, consumedAt time.Time) error {
	if public == nil || public.implementation == nil {
		return ErrInvalidConfiguration
	}
	return public.implementation.Consume(ctx, encoded, consumedAt)
}

func (consumer *consumer) Consume(ctx context.Context, encoded []byte, consumedAt time.Time) error {
	if ctx == nil || consumedAt.IsZero() || consumedAt.Year() < 2020 || consumedAt.Year() > 2100 {
		return ErrInvalidEvent
	}
	if ctx.Err() != nil {
		return ErrCanceled
	}
	envelope, err := contractevents.UnmarshalEnvelope(encoded)
	if err != nil {
		return ErrInvalidEvent
	}
	eventID, err := uuid.Parse(envelope.GetEventId())
	if err != nil || eventID == uuid.Nil || eventID.String() != envelope.GetEventId() {
		return ErrInvalidEvent
	}
	expiresAt := consumedAt.Add(consumer.retention)
	if !expiresAt.After(consumedAt) || expiresAt.Year() > 2101 {
		return ErrInvalidEvent
	}
	tx, err := consumer.beginner.Begin(ctx)
	if err != nil {
		return mapConsumerContextOrStore(ctx)
	}
	boundStore := consumer.storeFactory(tx)
	if boundStore == nil {
		rollback(ctx, tx)
		return ErrStore
	}
	rows, err := boundStore.RecordConsumedEvent(ctx, store.RecordConsumedEventParams{
		Consumer:   consumer.name,
		EventID:    eventID,
		ConsumedAt: consumedAt,
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		rollback(ctx, tx)
		return mapConsumerContextOrStore(ctx)
	}
	if rows != 0 && rows != 1 {
		rollback(ctx, tx)
		return ErrStore
	}
	if rows == 1 {
		if err := consumer.handler.HandleInTransaction(ctx, tx, envelope); err != nil {
			rollback(ctx, tx)
			if ctx.Err() != nil {
				return ErrCanceled
			}
			return ErrHandler
		}
	}
	if err := tx.Commit(ctx); err != nil {
		if ctx.Err() != nil {
			return ErrCanceled
		}
		return ErrCommit
	}
	return nil
}

func rollback(operationContext context.Context, tx transaction) {
	base := context.WithoutCancel(operationContext)
	ctx, cancel := context.WithTimeout(base, rollbackTimeout)
	defer cancel()
	_ = tx.Rollback(ctx)
}

func mapConsumerContextOrStore(ctx context.Context) error {
	if ctx != nil && ctx.Err() != nil {
		return ErrCanceled
	}
	return ErrStore
}

func validConsumerName(value string) bool {
	if len(value) == 0 || len(value) > 128 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' || character == '.' || character == '-' {
			continue
		}
		return false
	}
	return true
}
