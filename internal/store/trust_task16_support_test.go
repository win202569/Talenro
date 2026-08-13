package store_test

import (
	"bytes"
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

// TestTask16LatestBundleIssuanceUsesGeneratedAuthorizationOrdering catches a
// query that returns an arbitrary or oldest issuance for an authorization.
func TestTask16LatestBundleIssuanceUsesGeneratedAuthorizationOrdering(t *testing.T) {
	authorizationID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	bundleID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	now := time.Date(2026, time.August, 13, 8, 0, 0, 0, time.UTC)
	database := &task16LatestBundleDB{row: task16LatestBundleRow{values: []any{
		bundleID,
		authorizationID,
		int64(7),
		bytes.Repeat([]byte{1}, 32),
		bytes.Repeat([]byte{2}, 60),
		int32(3),
		[]byte(`{"envelope_version":"talenro-config-envelope/v1"}`),
		bytes.Repeat([]byte{4}, 32),
		"abcdefghijklmnop",
		now,
		now,
		now.Add(24 * time.Hour),
	}}}

	issuance, err := store.New(database).GetLatestBundleIssuance(context.Background(), authorizationID)
	if err != nil {
		t.Fatalf("GetLatestBundleIssuance: %v", err)
	}
	if database.authorizationID != authorizationID {
		t.Fatalf("authorization argument = %s, want %s", database.authorizationID, authorizationID)
	}
	normalized := strings.Join(strings.Fields(database.query), " ")
	if !strings.Contains(normalized, "WHERE authorization_id=$1 ORDER BY bundle_version DESC LIMIT 1") {
		t.Fatalf("generated query does not select latest authorization issuance: %q", normalized)
	}
	if issuance.ID != bundleID || issuance.AuthorizationID != authorizationID || issuance.BundleVersion != 7 ||
		!bytes.Equal(issuance.EnvelopeSha256, bytes.Repeat([]byte{4}, 32)) {
		t.Fatal("generated query did not scan the complete immutable issuance")
	}
}

type task16LatestBundleDB struct {
	query           string
	authorizationID uuid.UUID
	row             task16LatestBundleRow
}

func (*task16LatestBundleDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}

func (*task16LatestBundleDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}

func (database *task16LatestBundleDB) QueryRow(_ context.Context, query string, arguments ...any) pgx.Row {
	database.query = query
	if len(arguments) == 1 {
		database.authorizationID, _ = arguments[0].(uuid.UUID)
	}
	return database.row
}

type task16LatestBundleRow struct {
	values []any
	err    error
}

func (row task16LatestBundleRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != len(row.values) {
		return errors.New("unexpected scan destination count")
	}
	for index, value := range row.values {
		switch destination := destinations[index].(type) {
		case *uuid.UUID:
			*destination = value.(uuid.UUID)
		case *int64:
			*destination = value.(int64)
		case *[]byte:
			*destination = bytes.Clone(value.([]byte))
		case *int32:
			*destination = value.(int32)
		case *string:
			*destination = value.(string)
		case *time.Time:
			*destination = value.(time.Time)
		default:
			return errors.New("unexpected scan destination type")
		}
	}
	return nil
}
