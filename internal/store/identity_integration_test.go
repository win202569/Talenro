//go:build integration

package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"talenro.local/platform/internal/store"
	"talenro.local/platform/internal/testinfra"
)

func TestIdentityAccountAndSessionRoundTrip(t *testing.T) {
	pool := testinfra.OpenMigratedPostgres(t)
	queries := store.New(pool)
	principalID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	err := queries.CreateAccount(context.Background(), store.CreateAccountParams{
		ID:           principalID,
		State:        "pending_email",
		StateVersion: 1,
		Locale:       "en",
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := queries.GetAccountForUpdate(context.Background(), principalID)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != "pending_email" || got.StateVersion != 1 {
		t.Fatalf("unexpected account state: %q version: %d", got.State, got.StateVersion)
	}
}
