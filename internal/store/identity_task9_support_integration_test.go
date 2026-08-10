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
	"github.com/jackc/pgx/v5/pgtype"
	"talenro.local/platform/internal/store"
	"talenro.local/platform/internal/testinfra"
)

func TestIdentityTask9SupportIndexesAndLockQueries(t *testing.T) {
	pool := testinfra.OpenMigratedPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal("begin Task 9 support transaction")
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	queries := store.New(tx)

	now := time.Now().UTC().Truncate(time.Microsecond)
	firstPrincipal := uuid.New()
	secondPrincipal := uuid.New()
	createTask9SupportAccount(t, ctx, queries, firstPrincipal, now)
	createTask9SupportAccount(t, ctx, queries, secondPrincipal, now)

	verificationHash := bytes.Repeat([]byte{0x31}, 32)
	if err = queries.CreateEmailIdentity(ctx, task9EmailIdentity(firstPrincipal, verificationHash, 0x41, now)); err != nil {
		t.Fatal("create first Task 9 email identity")
	}
	locked, err := queries.GetEmailVerificationForUpdate(ctx, store.GetEmailVerificationForUpdateParams{
		VerificationTokenHash: verificationHash,
		VerificationExpiresAt: sql.NullTime{Time: now, Valid: true},
	})
	if err != nil || locked.PrincipalID != firstPrincipal {
		t.Fatal("verification-token lock query did not return the owning principal")
	}
	if _, err = queries.GetEmailVerificationForUpdate(ctx, store.GetEmailVerificationForUpdateParams{
		VerificationTokenHash: verificationHash,
		VerificationExpiresAt: sql.NullTime{Time: now.Add(25 * time.Hour), Valid: true},
	}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatal("expired verification token remained eligible")
	}
	requireTask9UniqueViolation(t, ctx, tx, "verification token hash", func(nested *store.Queries) error {
		return nested.CreateEmailIdentity(ctx, task9EmailIdentity(secondPrincipal, verificationHash, 0x42, now))
	})

	sessionID := uuid.New()
	if err = queries.CreateAccountSession(ctx, store.CreateAccountSessionParams{
		ID: sessionID, PrincipalID: firstPrincipal,
		ClientSigningPublicKey: bytes.Repeat([]byte{0x51}, 32),
		AccessTokenHash:        bytes.Repeat([]byte{0x52}, 32),
		AccessExpiresAt:        now.Add(10 * time.Minute),
		AbsoluteExpiresAt:      now.Add(90 * 24 * time.Hour),
		CreatedAt:              now,
	}); err != nil {
		t.Fatal("create Task 9 account session")
	}
	if _, err = queries.GetAccountSessionForUpdate(ctx, store.GetAccountSessionForUpdateParams{ID: sessionID, PrincipalID: firstPrincipal}); err != nil {
		t.Fatal("owned account session was not lockable")
	}
	if _, err = queries.GetAccountSessionForUpdate(ctx, store.GetAccountSessionForUpdateParams{ID: sessionID, PrincipalID: secondPrincipal}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatal("account session was visible to a different principal")
	}

	createTask9PasswordCredential(t, ctx, queries, firstPrincipal, 0x61, now)
	createTask9PasswordCredential(t, ctx, queries, secondPrincipal, 0x62, now)
	resetHash := bytes.Repeat([]byte{0x63}, 32)
	if _, err = queries.SetPasswordReset(ctx, task9PasswordReset(firstPrincipal, resetHash, 0x64, now)); err != nil {
		t.Fatal("set first Task 9 password reset")
	}
	requireTask9UniqueViolation(t, ctx, tx, "password reset token hash", func(nested *store.Queries) error {
		_, setErr := nested.SetPasswordReset(ctx, task9PasswordReset(secondPrincipal, resetHash, 0x65, now))
		return setErr
	})
}

func requireTask9UniqueViolation(t *testing.T, ctx context.Context, parent pgx.Tx, field string, attempt func(*store.Queries) error) {
	t.Helper()
	nested, err := parent.Begin(ctx)
	if err != nil {
		t.Fatalf("begin %s uniqueness savepoint", field)
	}
	if attempt(store.New(nested)) == nil {
		_ = nested.Rollback(ctx)
		t.Fatalf("duplicate non-null %s was accepted", field)
	}
	if err = nested.Rollback(ctx); err != nil {
		t.Fatalf("rollback %s uniqueness savepoint", field)
	}
}

func createTask9SupportAccount(t *testing.T, ctx context.Context, queries *store.Queries, principalID uuid.UUID, now time.Time) {
	t.Helper()
	if err := queries.CreateAccount(ctx, store.CreateAccountParams{
		ID: principalID, State: "pending_email", StateVersion: 1, Locale: "en", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal("create Task 9 support account")
	}
}

func task9EmailIdentity(principalID uuid.UUID, verificationHash []byte, discriminator byte, now time.Time) store.CreateEmailIdentityParams {
	return store.CreateEmailIdentityParams{
		ID: uuid.New(), PrincipalID: principalID, LookupKeyVersion: 1,
		LookupDigest: bytes.Repeat([]byte{discriminator}, 32),
		Ciphertext:   bytes.Repeat([]byte{discriminator}, 29), EncryptionKeyVersion: 1,
		VerificationTokenHash:          verificationHash,
		VerificationExpiresAt:          sql.NullTime{Time: now.Add(24 * time.Hour), Valid: true},
		VerificationDeliveryID:         uuid.NullUUID{UUID: uuid.New(), Valid: true},
		VerificationDeliveryCiphertext: bytes.Repeat([]byte{discriminator}, 29),
		VerificationDeliveryKeyVersion: pgtype.Int4{Int32: 1, Valid: true},
		CreatedAt:                      now,
	}
}

func createTask9PasswordCredential(t *testing.T, ctx context.Context, queries *store.Queries, principalID uuid.UUID, discriminator byte, now time.Time) {
	t.Helper()
	if err := queries.CreatePasswordCredential(ctx, store.CreatePasswordCredentialParams{
		PrincipalID: principalID, PolicyVersion: 1, MemoryKib: 65536, TimeCost: 3, Parallelism: 4,
		Salt: bytes.Repeat([]byte{discriminator}, 16), PasswordHash: bytes.Repeat([]byte{discriminator}, 32), UpdatedAt: now,
	}); err != nil {
		t.Fatal("create Task 9 password credential")
	}
}

func task9PasswordReset(principalID uuid.UUID, resetHash []byte, discriminator byte, now time.Time) store.SetPasswordResetParams {
	return store.SetPasswordResetParams{
		PrincipalID: principalID, ResetTokenHash: resetHash,
		ResetExpiresAt:          sql.NullTime{Time: now.Add(30 * time.Minute), Valid: true},
		ResetDeliveryID:         uuid.NullUUID{UUID: uuid.New(), Valid: true},
		ResetDeliveryCiphertext: bytes.Repeat([]byte{discriminator}, 29),
		ResetDeliveryKeyVersion: pgtype.Int4{Int32: 1, Valid: true}, UpdatedAt: now,
	}
}
