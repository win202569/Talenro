//go:build integration

package store_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
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

func TestTwoTrialGrantsAcrossIndependentSessionsCommitOnlyOneProvisionalDevice(t *testing.T) {
	pool := testinfra.OpenMigratedPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	queries := store.New(pool)
	principalID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	deadline := now.Add(time.Hour)
	if err := queries.CreateAccount(ctx, store.CreateAccountParams{
		ID: principalID, State: "pending_email", StateVersion: 1, Locale: "en", CreatedAt: now.Add(-23 * time.Hour), UpdatedAt: now,
	}); err != nil {
		t.Fatal("create pending account fixture failed")
	}
	type participant struct {
		sessionID, grantID, deviceID, authorizationID, familyID uuid.UUID
		grantHash                                               []byte
	}
	participants := [2]participant{}
	for index := range participants {
		participants[index] = participant{
			sessionID: uuid.New(), grantID: uuid.New(), deviceID: uuid.New(), authorizationID: uuid.New(), familyID: uuid.New(),
		}
		participants[index].grantHash = bytes.Repeat(participants[index].grantID[:], 2)
		if err := queries.CreateAccountSession(ctx, store.CreateAccountSessionParams{
			ID: participants[index].sessionID, PrincipalID: principalID, ClientSigningPublicKey: bytes.Repeat([]byte{byte(0x31 + index)}, 32),
			AccessTokenHash: bytes.Repeat(participants[index].sessionID[:], 2), AccessExpiresAt: deadline, AbsoluteExpiresAt: deadline, CreatedAt: now,
		}); err != nil {
			t.Fatal("create independent session fixture failed")
		}
		if err := queries.CreateEnrollmentGrant(ctx, store.CreateEnrollmentGrantParams{
			ID: participants[index].grantID, PrincipalID: principalID, AccountSessionID: participants[index].sessionID,
			TokenHash: participants[index].grantHash, PolicyMarker: "trial_restricted",
			ProvisionalUntil: sql.NullTime{Time: deadline, Valid: true}, ExpiresAt: now.Add(10 * time.Minute), CreatedAt: now,
		}); err != nil {
			t.Fatal("create independent trial grant fixture failed")
		}
	}
	ready := make(chan struct{}, 2)
	start := make(chan struct{})
	results := make(chan grantConsumptionResult, 2)
	for index := range participants {
		participant := participants[index]
		go func() {
			err := enrollProvisionalGraph(ctx, pool, principalID, participant.sessionID, participant.grantHash,
				participant.deviceID, participant.authorizationID, participant.familyID, deadline, now, ready, start)
			results <- grantConsumptionResult{deviceID: participant.deviceID, err: err}
		}()
	}
	for range 2 {
		select {
		case <-ready:
		case <-ctx.Done():
			t.Fatal("two-grant transactions did not reach authorization barrier")
		}
	}
	close(start)
	got := [2]grantConsumptionResult{<-results, <-results}
	successes, conflicts := 0, 0
	loser := uuid.Nil
	for _, result := range got {
		if result.err == nil {
			successes++
		} else {
			conflicts++
			loser = result.deviceID
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("two-grant outcomes: successes=%d conflicts=%d", successes, conflicts)
	}
	for _, check := range []struct {
		name  string
		query string
		want  int
	}{
		{name: "devices", query: `SELECT count(*) FROM deviceauth.devices WHERE principal_id=$1`, want: 1},
		{name: "authorizations", query: `SELECT count(*) FROM deviceauth.device_authorizations WHERE principal_id=$1`, want: 1},
		{name: "families", query: `SELECT count(*) FROM deviceauth.device_token_families f JOIN deviceauth.device_authorizations a ON a.id=f.authorization_id WHERE a.principal_id=$1`, want: 1},
		{name: "snapshots", query: `SELECT count(*) FROM deviceauth.device_policy_snapshots p JOIN deviceauth.device_authorizations a ON a.id=p.authorization_id WHERE a.principal_id=$1`, want: 1},
		{name: "consumed grants", query: `SELECT count(*) FROM deviceauth.enrollment_grants WHERE principal_id=$1 AND state='consumed'`, want: 1},
		{name: "bound sessions", query: `SELECT count(*) FROM identity.account_sessions WHERE principal_id=$1 AND device_authorization_id IS NOT NULL`, want: 1},
	} {
		var count int
		if err := pool.QueryRow(ctx, check.query, principalID).Scan(&count); err != nil || count != check.want {
			t.Fatalf("%s count = %d, %v; want %d", check.name, count, err, check.want)
		}
	}
	var loserResidue int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM deviceauth.devices WHERE id=$1`, loser).Scan(&loserResidue); err != nil || loserResidue != 0 {
		t.Fatalf("losing graph residue = %d, %v", loserResidue, err)
	}
}

func enrollProvisionalGraph(
	ctx context.Context,
	pool *pgxpool.Pool,
	principalID, sessionID uuid.UUID,
	grantHash []byte,
	deviceID, authorizationID, familyID uuid.UUID,
	deadline, now time.Time,
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
		ID: deviceID, PrincipalID: principalID, SigningPublicKey: bytes.Repeat(deviceID[:], 2),
		HpkePublicKey: bytes.Repeat(familyID[:], 2), KeyVersion: 1, CreatedAt: now,
	}); err != nil {
		return err
	}
	if _, err = queries.ConsumeEnrollmentGrant(ctx, store.ConsumeEnrollmentGrantParams{
		TokenHash: grantHash, ConsumedDeviceID: uuid.NullUUID{UUID: deviceID, Valid: true}, ConsumedAt: sql.NullTime{Time: now, Valid: true},
	}); err != nil {
		return err
	}
	ready <- struct{}{}
	select {
	case <-start:
	case <-ctx.Done():
		return ctx.Err()
	}
	if err = queries.CreateDeviceAuthorization(ctx, store.CreateDeviceAuthorizationParams{
		ID: authorizationID, PrincipalID: principalID, DeviceID: deviceID, State: "provisional",
		ProvisionalUntil: sql.NullTime{Time: deadline, Valid: true}, CreatedAt: now,
	}); err != nil {
		return err
	}
	if rows, bindErr := queries.BindAccountSessionToAuthorization(ctx, store.BindAccountSessionToAuthorizationParams{
		ID: sessionID, DeviceAuthorizationID: uuid.NullUUID{UUID: authorizationID, Valid: true}, UpdatedAt: now,
	}); bindErr != nil || rows != 1 {
		return errors.New("session bind failed")
	}
	policy := json.RawMessage(`{"expires_at":"` + deadline.UTC().Format(time.RFC3339) + `","max_devices":"1","mode":"trial_restricted"}`)
	if err = queries.CreateDevicePolicySnapshot(ctx, store.CreateDevicePolicySnapshotParams{AuthorizationID: authorizationID, Policy: policy, CreatedAt: now}); err != nil {
		return err
	}
	if err = queries.CreateDeviceTokenFamily(ctx, store.CreateDeviceTokenFamilyParams{
		ID: familyID, AuthorizationID: authorizationID, AccessTokenHash: bytes.Repeat(familyID[:], 2),
		AccessExpiresAt: deadline, IdleExpiresAt: now.Add(30 * 24 * time.Hour), AbsoluteExpiresAt: now.Add(90 * 24 * time.Hour), CreatedAt: now,
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
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
