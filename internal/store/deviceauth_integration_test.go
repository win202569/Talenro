//go:build integration

package store_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"talenro.local/platform/internal/store"
	"talenro.local/platform/internal/testinfra"
)

type grantConsumptionResult struct {
	deviceID uuid.UUID
	err      error
}

func TestDeviceEnrollmentGrantIsConsumedExactlyOnce(t *testing.T) {
	pool := testinfra.OpenMigratedPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	queries := store.New(pool)
	principalID := uuid.New()
	sessionID := uuid.New()
	grantID := uuid.New()
	tokenHash := bytes.Repeat(grantID[:], 2)
	now := time.Now().UTC().Truncate(time.Microsecond)

	if err := queries.CreateAccount(ctx, store.CreateAccountParams{
		ID:           principalID,
		State:        "active",
		StateVersion: 1,
		Locale:       "en",
		CreatedAt:    now,
		UpdatedAt:    now,
	}); err != nil {
		t.Fatal("create account fixture failed")
	}
	if err := queries.CreateAccountSession(ctx, store.CreateAccountSessionParams{
		ID:                     sessionID,
		PrincipalID:            principalID,
		ClientSigningPublicKey: bytes.Repeat([]byte{4}, 32),
		AccessTokenHash:        bytes.Repeat(sessionID[:], 2),
		AccessExpiresAt:        now.Add(10 * time.Minute),
		AbsoluteExpiresAt:      now.Add(90 * 24 * time.Hour),
		CreatedAt:              now,
	}); err != nil {
		t.Fatal("create account session fixture failed")
	}
	if err := queries.CreateEnrollmentGrant(ctx, store.CreateEnrollmentGrantParams{
		ID:               grantID,
		PrincipalID:      principalID,
		AccountSessionID: sessionID,
		TokenHash:        tokenHash,
		PolicyMarker:     "standard",
		ExpiresAt:        now.Add(10 * time.Minute),
		CreatedAt:        now,
	}); err != nil {
		t.Fatal("create enrollment grant fixture failed")
	}

	ready := make(chan struct{}, 2)
	start := make(chan struct{})
	results := make(chan grantConsumptionResult, 2)
	for range 2 {
		deviceID := uuid.New()
		go func() {
			err := consumeGrant(ctx, pool, principalID, tokenHash, deviceID, now, ready, start)
			results <- grantConsumptionResult{deviceID: deviceID, err: err}
		}()
	}

	for range 2 {
		select {
		case <-ready:
		case <-ctx.Done():
			t.Fatal("concurrent device transactions did not reach consume barrier")
		}
	}
	close(start)

	got := [2]grantConsumptionResult{<-results, <-results}
	var winner uuid.UUID
	var loser uuid.UUID
	successes := 0
	noRows := 0
	for _, result := range got {
		switch {
		case result.err == nil:
			successes++
			winner = result.deviceID
		case errors.Is(result.err, pgx.ErrNoRows):
			noRows++
			loser = result.deviceID
		default:
			t.Fatal("unexpected concurrent grant consumption failure")
		}
	}
	if successes != 1 || noRows != 1 {
		t.Fatalf("grant outcomes: successes=%d no_rows=%d", successes, noRows)
	}

	var deviceCount int
	var consumedDeviceID uuid.UUID
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM deviceauth.devices WHERE principal_id = $1`,
		principalID,
	).Scan(&deviceCount); err != nil {
		t.Fatal("count devices failed")
	}
	if deviceCount != 1 {
		t.Fatalf("persisted device count = %d, want 1", deviceCount)
	}
	if err := pool.QueryRow(ctx,
		`SELECT consumed_device_id FROM deviceauth.enrollment_grants WHERE id = $1`,
		grantID,
	).Scan(&consumedDeviceID); err != nil {
		t.Fatal("read consumed grant failed")
	}
	if consumedDeviceID != winner {
		t.Fatal("grant was not bound to the winning device")
	}
	var loserCount int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM deviceauth.devices WHERE id = $1`,
		loser,
	).Scan(&loserCount); err != nil {
		t.Fatal("count losing device failed")
	}
	if loserCount != 0 {
		t.Fatal("losing device insert was not rolled back")
	}
}

func consumeGrant(
	ctx context.Context,
	pool *pgxpool.Pool,
	principalID uuid.UUID,
	tokenHash []byte,
	deviceID uuid.UUID,
	now time.Time,
	ready chan<- struct{},
	start <-chan struct{},
) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := store.New(tx)
	if err = queries.CreateDevice(ctx, store.CreateDeviceParams{
		ID:               deviceID,
		PrincipalID:      principalID,
		SigningPublicKey: bytes.Repeat(deviceID[:], 2),
		HpkePublicKey:    bytes.Repeat(deviceID[:], 2),
		KeyVersion:       1,
		CreatedAt:        now,
	}); err != nil {
		return err
	}

	ready <- struct{}{}
	select {
	case <-start:
	case <-ctx.Done():
		return ctx.Err()
	}

	if _, err = queries.ConsumeEnrollmentGrant(ctx, store.ConsumeEnrollmentGrantParams{
		TokenHash:        tokenHash,
		ConsumedDeviceID: uuid.NullUUID{UUID: deviceID, Valid: true},
		ConsumedAt:       sql.NullTime{Time: now, Valid: true},
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
