//go:build integration

package store_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"talenro.local/platform/internal/store"
	"talenro.local/platform/internal/testinfra"
)

func TestAccountRefreshAuthorityIncludesCurrentAccountState(t *testing.T) {
	pool := testinfra.OpenMigratedPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal("begin account refresh authority transaction")
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	queries := store.New(tx)
	now := time.Now().UTC().Truncate(time.Microsecond)
	principalID := uuid.New()
	sessionID := uuid.New()
	refreshDigest := bytes.Repeat([]byte{0x51}, 32)
	createTask10SessionFixture(t, ctx, queries, principalID, sessionID, refreshDigest, 0x41, now)

	authority, err := queries.GetRefreshTokenForUpdate(ctx, refreshDigest)
	if err != nil || authority.AccountState != "active" || authority.SessionState != "active" {
		t.Fatal("refresh authority omitted active account/session state")
	}
	if _, err = tx.Exec(ctx, `UPDATE identity.accounts SET state='suspended', state_version=state_version+1, updated_at=$2 WHERE id=$1`, principalID, now.Add(time.Second)); err != nil {
		t.Fatal("suspend account fixture")
	}
	authority, err = queries.GetRefreshTokenForUpdate(ctx, refreshDigest)
	if err != nil || authority.AccountState != "suspended" {
		t.Fatal("refresh authority did not observe account suspension")
	}
}

func TestPrincipalSessionRevocationAtomicallyExcludesCurrentSession(t *testing.T) {
	pool := testinfra.OpenMigratedPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal("begin principal session revocation transaction")
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	queries := store.New(tx)
	now := time.Now().UTC().Truncate(time.Microsecond)
	principalID := uuid.New()
	currentSession := uuid.New()
	otherSession := uuid.New()
	currentRefresh := bytes.Repeat([]byte{0x61}, 32)
	otherRefresh := bytes.Repeat([]byte{0x62}, 32)
	createTask10SessionFixture(t, ctx, queries, principalID, currentSession, currentRefresh, 0x42, now)
	if err = queries.CreateAccountSession(ctx, store.CreateAccountSessionParams{
		ID: otherSession, PrincipalID: principalID, ClientSigningPublicKey: bytes.Repeat([]byte{0x43}, 32),
		AccessTokenHash: bytes.Repeat([]byte{0x44}, 32), AccessExpiresAt: now.Add(10 * time.Minute),
		AbsoluteExpiresAt: now.Add(90 * 24 * time.Hour), CreatedAt: now,
	}); err != nil {
		t.Fatal("create other account session")
	}
	if err = queries.InsertAccountRefreshToken(ctx, store.InsertAccountRefreshTokenParams{
		TokenHash: otherRefresh, SessionID: otherSession, IssuedAt: now,
		IdleExpiresAt: now.Add(30 * 24 * time.Hour), AbsoluteExpiresAt: now.Add(90 * 24 * time.Hour),
	}); err != nil {
		t.Fatal("create other account refresh")
	}

	revoked, err := queries.RevokePrincipalAccountSessions(ctx, store.RevokePrincipalAccountSessionsParams{
		PrincipalID: principalID, RevokeScope: "others", SessionID: currentSession, RevokedAt: now.Add(time.Second),
	})
	if err != nil || len(revoked) != 1 || revoked[0] != otherSession {
		t.Fatalf("revoked sessions = %v", revoked)
	}
	current, err := queries.GetRefreshTokenForUpdate(ctx, currentRefresh)
	if err != nil || current.RefreshState != "active" || current.SessionState != "active" {
		t.Fatal("current session was changed by others-scope revocation")
	}
	other, err := queries.GetRefreshTokenForUpdate(ctx, otherRefresh)
	if err != nil || other.RefreshState != "revoked" || other.SessionState != "revoked" {
		t.Fatal("other session and refresh were not revoked together")
	}
}

func createTask10SessionFixture(t *testing.T, ctx context.Context, queries *store.Queries, principalID, sessionID uuid.UUID, refreshDigest []byte, discriminator byte, now time.Time) {
	t.Helper()
	if err := queries.CreateAccount(ctx, store.CreateAccountParams{
		ID: principalID, State: "active", StateVersion: 1, Locale: "en", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal("create Task 10 account")
	}
	if err := queries.CreateAccountSession(ctx, store.CreateAccountSessionParams{
		ID: sessionID, PrincipalID: principalID, ClientSigningPublicKey: bytes.Repeat([]byte{discriminator}, 32),
		AccessTokenHash: bytes.Repeat([]byte{discriminator + 1}, 32), AccessExpiresAt: now.Add(10 * time.Minute),
		AbsoluteExpiresAt: now.Add(90 * 24 * time.Hour), CreatedAt: now,
	}); err != nil {
		t.Fatal("create Task 10 account session")
	}
	if err := queries.InsertAccountRefreshToken(ctx, store.InsertAccountRefreshTokenParams{
		TokenHash: refreshDigest, SessionID: sessionID, IssuedAt: now,
		IdleExpiresAt: now.Add(30 * 24 * time.Hour), AbsoluteExpiresAt: now.Add(90 * 24 * time.Hour),
	}); err != nil {
		t.Fatal("create Task 10 refresh")
	}
}
