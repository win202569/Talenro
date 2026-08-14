package store_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"talenro.local/platform/internal/store"
)

func TestTask18GeneratedConsumedEventPresenceUsesCompositeKey(t *testing.T) {
	eventID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	database := &task18DeliveryDB{row: task18DeliveryRow{exists: true}}

	exists, err := store.New(database).HasConsumedEvent(context.Background(), store.HasConsumedEventParams{
		Consumer: "identity.email-delivery.v1",
		EventID:  eventID,
	})
	if err != nil {
		t.Fatalf("HasConsumedEvent: %v", err)
	}
	if !exists {
		t.Fatal("stored consumer/event pair was not reported present")
	}
	normalized := strings.Join(strings.Fields(database.query), " ")
	if !strings.Contains(normalized, "SELECT EXISTS ( SELECT 1 FROM consumed_event_ids WHERE consumer=$1 AND event_id=$2 )") {
		t.Fatalf("generated query does not use the consumer/event composite key: %q", normalized)
	}
	if len(database.arguments) != 2 || database.arguments[0] != "identity.email-delivery.v1" || database.arguments[1] != eventID {
		t.Fatalf("generated presence arguments = %#v", database.arguments)
	}
}

func TestTask18GeneratedPasswordResetClearUsesExactDeliveryID(t *testing.T) {
	deliveryID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	now := time.Date(2026, time.August, 13, 12, 0, 0, 0, time.UTC)
	database := &task18DeliveryDB{commandTag: pgconn.NewCommandTag("UPDATE 1")}

	rows, err := store.New(database).ClearPendingPasswordResetDelivery(context.Background(), store.ClearPendingPasswordResetDeliveryParams{
		ResetDeliveryID: uuid.NullUUID{UUID: deliveryID, Valid: true},
		UpdatedAt:       now,
	})
	if err != nil {
		t.Fatalf("ClearPendingPasswordResetDelivery: %v", err)
	}
	if rows != 1 {
		t.Fatalf("cleared rows = %d, want 1", rows)
	}
	normalized := strings.Join(strings.Fields(database.query), " ")
	for _, fragment := range []string{
		"reset_delivery_id=NULL",
		"reset_delivery_ciphertext=NULL",
		"reset_delivery_key_version=NULL",
		"updated_at=$2 WHERE reset_delivery_id=$1",
	} {
		if !strings.Contains(normalized, fragment) {
			t.Fatalf("generated clear query missing %q: %q", fragment, normalized)
		}
	}
	if len(database.arguments) != 2 || database.arguments[0] != (uuid.NullUUID{UUID: deliveryID, Valid: true}) || database.arguments[1] != now {
		t.Fatalf("generated clear arguments = %#v", database.arguments)
	}
}

type task18DeliveryDB struct {
	query      string
	arguments  []any
	row        task18DeliveryRow
	commandTag pgconn.CommandTag
}

func (database *task18DeliveryDB) Exec(_ context.Context, query string, arguments ...any) (pgconn.CommandTag, error) {
	database.query = query
	database.arguments = append([]any(nil), arguments...)
	return database.commandTag, nil
}

func (*task18DeliveryDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}

func (database *task18DeliveryDB) QueryRow(_ context.Context, query string, arguments ...any) pgx.Row {
	database.query = query
	database.arguments = append([]any(nil), arguments...)
	return database.row
}

type task18DeliveryRow struct {
	exists bool
	err    error
}

func (row task18DeliveryRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != 1 {
		return errors.New("unexpected scan destination count")
	}
	destination, ok := destinations[0].(*bool)
	if !ok {
		return errors.New("unexpected scan destination type")
	}
	*destination = row.exists
	return nil
}
