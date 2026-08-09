//go:build integration

// Package testinfra provides bounded integration-test dependencies.
package testinfra

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// OpenMigratedPostgres opens the externally migrated integration database.
func OpenMigratedPostgres(t testing.TB) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TALENRO_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("testinfra: postgres unavailable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal("testinfra: postgres unavailable")
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatal("testinfra: postgres unavailable")
	}

	return pool
}
