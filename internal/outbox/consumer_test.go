package outbox

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	contractevents "talenro.local/platform/internal/contracts/events"
	"talenro.local/platform/internal/store"
)

func TestConsumerRunsSideEffectOnceForDuplicateEventID(t *testing.T) {
	t.Parallel()

	envelope := accountStateEnvelope(t)
	encoded, err := contractevents.MarshalEnvelope(envelope)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	state := &consumerFakeState{}
	handlerCalls := 0
	var handlerDBTX store.DBTX
	handler := HandlerFunc(func(_ context.Context, tx store.DBTX, got *eventsv1.EventEnvelope) error {
		state.order = append(state.order, "handle")
		handlerCalls++
		handlerDBTX = tx
		got.Payload[0] ^= 1
		return nil
	})
	consumer, err := newConsumerWithBeginner(&fakeBeginner{state: state}, func(tx store.DBTX) consumerStore {
		return &fakeConsumerStore{state: state, tx: tx}
	}, "identity.projector", 30*24*time.Hour, handler)
	if err != nil {
		t.Fatal(err)
	}
	for delivery := 0; delivery < 2; delivery++ {
		if err := consumer.Consume(context.Background(), encoded, now.Add(time.Duration(delivery)*time.Second)); err != nil {
			t.Fatalf("delivery %d: %v", delivery, err)
		}
	}
	if handlerCalls != 1 {
		t.Fatalf("handler calls = %d, want 1", handlerCalls)
	}
	if state.transactions[0] != handlerDBTX {
		t.Fatal("handler did not receive transaction used by RecordConsumedEvent")
	}
	if len(state.recordTxs) != 2 || state.transactions[0] != state.recordTxs[0] || state.transactions[1] != state.recordTxs[1] {
		t.Fatal("RecordConsumedEvent was not bound to the handler transaction")
	}
	wantOrder := []string{"begin", "record", "handle", "commit", "begin", "record", "commit"}
	if !equalStrings(state.order, wantOrder) {
		t.Fatalf("order = %v, want %v", state.order, wantOrder)
	}
	if !bytes.Equal(encoded, mustMarshalEnvelope(t, envelope)) {
		t.Fatal("consumer or handler mutated caller payload")
	}
}

func TestConsumerRollsBackOnRecordAndHandlerErrorsAndReturnsFixedErrors(t *testing.T) {
	t.Parallel()

	encoded := mustMarshalEnvelope(t, accountStateEnvelope(t))
	now := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	tests := []struct {
		name       string
		recordErr  error
		handlerErr error
		wantErr    error
		wantOrder  []string
	}{
		{name: "record error", recordErr: errors.New("SECRET_RECORD_CANARY"), wantErr: ErrStore, wantOrder: []string{"begin", "record", "rollback"}},
		{name: "handler error", handlerErr: errors.New("SECRET_HANDLER_CANARY"), wantErr: ErrHandler, wantOrder: []string{"begin", "record", "handle", "rollback"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := &consumerFakeState{recordErr: test.recordErr}
			consumer, err := newConsumerWithBeginner(&fakeBeginner{state: state}, func(tx store.DBTX) consumerStore {
				return &fakeConsumerStore{state: state, tx: tx}
			}, "identity.projector", time.Hour, HandlerFunc(func(context.Context, store.DBTX, *eventsv1.EventEnvelope) error {
				state.order = append(state.order, "handle")
				return test.handlerErr
			}))
			if err != nil {
				t.Fatal(err)
			}
			consumeErr := consumer.Consume(context.Background(), encoded, now)
			if !errors.Is(consumeErr, test.wantErr) || bytes.Contains([]byte(consumeErr.Error()), []byte("CANARY")) {
				t.Fatalf("error = %v, want fixed %v", consumeErr, test.wantErr)
			}
			if !equalStrings(state.order, test.wantOrder) {
				t.Fatalf("order = %v, want %v", state.order, test.wantOrder)
			}
		})
	}
}

func TestConsumerReturnsFixedCommitFailureWithoutSideEffectReplayWithinDelivery(t *testing.T) {
	t.Parallel()

	state := &consumerFakeState{commitErr: errors.New("SECRET_COMMIT_CANARY")}
	consumer, err := newConsumerWithBeginner(&fakeBeginner{state: state}, func(tx store.DBTX) consumerStore {
		return &fakeConsumerStore{state: state, tx: tx}
	}, "identity.projector", time.Hour, HandlerFunc(func(context.Context, store.DBTX, *eventsv1.EventEnvelope) error {
		state.order = append(state.order, "handle")
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	err = consumer.Consume(context.Background(), mustMarshalEnvelope(t, accountStateEnvelope(t)), time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC))
	if !errors.Is(err, ErrCommit) || bytes.Contains([]byte(err.Error()), []byte("CANARY")) {
		t.Fatalf("error = %v, want fixed ErrCommit", err)
	}
	if !equalStrings(state.order, []string{"begin", "record", "handle", "commit"}) {
		t.Fatalf("order = %v", state.order)
	}
}

func TestConsumerCancellationRollsBackWithIndependentBoundedContext(t *testing.T) {
	t.Parallel()

	operationContext, cancel := context.WithCancel(context.Background())
	state := &consumerFakeState{recordErr: errors.New("SECRET_CANCELED_RECORD"), cancelOnRecord: cancel}
	consumer, err := newConsumerWithBeginner(&fakeBeginner{state: state}, func(tx store.DBTX) consumerStore {
		return &fakeConsumerStore{state: state, tx: tx}
	}, "identity.projector", time.Hour, HandlerFunc(func(context.Context, store.DBTX, *eventsv1.EventEnvelope) error {
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	err = consumer.Consume(operationContext, mustMarshalEnvelope(t, accountStateEnvelope(t)), time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC))
	if !errors.Is(err, ErrCanceled) {
		t.Fatalf("error = %v, want ErrCanceled", err)
	}
	if !equalStrings(state.order, []string{"begin", "record", "rollback"}) {
		t.Fatalf("order = %v", state.order)
	}
	if state.rollbackContextErr != nil {
		t.Fatalf("rollback inherited operation cancellation: %v", state.rollbackContextErr)
	}
}

func TestConsumerRejectsInvalidInputBeforeTransaction(t *testing.T) {
	t.Parallel()

	valid := accountStateEnvelope(t)
	validEncoded := mustMarshalEnvelope(t, valid)
	tests := []struct {
		name      string
		consumer  string
		retention time.Duration
		payload   []byte
		consumed  time.Time
	}{
		{name: "invalid consumer", consumer: "UPPER", retention: time.Hour, payload: validEncoded, consumed: time.Now()},
		{name: "consumer outside database regex", consumer: "identity_projector", retention: time.Hour, payload: validEncoded, consumed: time.Now()},
		{name: "zero retention", consumer: "identity.projector", payload: validEncoded, consumed: time.Now()},
		{name: "unbounded retention", consumer: "identity.projector", retention: 366 * 24 * time.Hour, payload: validEncoded, consumed: time.Now()},
		{name: "malformed payload", consumer: "identity.projector", retention: time.Hour, payload: []byte{0xff}, consumed: time.Now()},
		{name: "zero consumed time", consumer: "identity.projector", retention: time.Hour, payload: validEncoded},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := &consumerFakeState{}
			consumer, err := newConsumerWithBeginner(&fakeBeginner{state: state}, func(tx store.DBTX) consumerStore {
				return &fakeConsumerStore{state: state, tx: tx}
			}, test.consumer, test.retention, HandlerFunc(func(context.Context, store.DBTX, *eventsv1.EventEnvelope) error { return nil }))
			if err == nil {
				err = consumer.Consume(context.Background(), test.payload, test.consumed)
			}
			if !errors.Is(err, ErrInvalidEvent) && !errors.Is(err, ErrInvalidConfiguration) {
				t.Fatalf("error = %v", err)
			}
			if len(state.order) != 0 {
				t.Fatalf("invalid input began transaction: %v", state.order)
			}
		})
	}
}

type consumerFakeState struct {
	order              []string
	transactions       []*fakeTx
	recorded           bool
	recordTxs          []store.DBTX
	recordErr          error
	commitErr          error
	cancelOnRecord     context.CancelFunc
	rollbackContextErr error
}

type fakeBeginner struct{ state *consumerFakeState }

func (beginner *fakeBeginner) Begin(context.Context) (transaction, error) {
	tx := &fakeTx{state: beginner.state}
	beginner.state.transactions = append(beginner.state.transactions, tx)
	beginner.state.order = append(beginner.state.order, "begin")
	return tx, nil
}

type fakeTx struct{ state *consumerFakeState }

func (*fakeTx) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	panic("unexpected Exec")
}
func (*fakeTx) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	panic("unexpected Query")
}
func (*fakeTx) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	panic("unexpected QueryRow")
}
func (tx *fakeTx) Commit(context.Context) error {
	tx.state.order = append(tx.state.order, "commit")
	return tx.state.commitErr
}

func (tx *fakeTx) Rollback(ctx context.Context) error {
	tx.state.order = append(tx.state.order, "rollback")
	tx.state.rollbackContextErr = ctx.Err()
	return nil
}

type fakeConsumerStore struct {
	state *consumerFakeState
	tx    store.DBTX
}

func (fake *fakeConsumerStore) RecordConsumedEvent(_ context.Context, _ store.RecordConsumedEventParams) (int64, error) {
	fake.state.order = append(fake.state.order, "record")
	fake.state.recordTxs = append(fake.state.recordTxs, fake.tx)
	if fake.state.cancelOnRecord != nil {
		fake.state.cancelOnRecord()
	}
	if fake.state.recordErr != nil {
		return 0, fake.state.recordErr
	}
	if fake.state.recorded {
		return 0, nil
	}
	fake.state.recorded = true
	return 1, nil
}

func mustMarshalEnvelope(t *testing.T, envelope *eventsv1.EventEnvelope) []byte {
	t.Helper()
	encoded, err := contractevents.MarshalEnvelope(envelope)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}
