//go:build integration

package store

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ func(*Queries, context.Context, string) (json.RawMessage, error) = (*Queries).GetSystemMetadata

func TestSystemMetadataRoundTrip(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("TALENRO_DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	queries := New(pool)
	want := json.RawMessage(`{"version":1}`)
	if err := queries.PutSystemMetadata(ctx, PutSystemMetadataParams{Key: "contracts.version", Value: want}); err != nil {
		t.Fatal(err)
	}
	got, err := queries.GetSystemMetadata(ctx, "contracts.version")
	if err != nil {
		t.Fatal(err)
	}
	equal, err := jsonEquivalent(got, want)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatalf("got JSON %s want JSON %s", got, want)
	}
}
