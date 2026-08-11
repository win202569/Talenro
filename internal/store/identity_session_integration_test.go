//go:build integration

package store_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
	refreshDigest := task10IntegrationDigest("authority-refresh", principalID, sessionID)
	createTask10SessionFixture(t, ctx, queries, principalID, sessionID, refreshDigest, 0x41, now)

	authority, err := task10LockRefreshAuthority(ctx, queries, refreshDigest)
	if err != nil || authority.account.State != "active" || authority.session.State != "active" {
		t.Fatal("refresh authority omitted active account/session state")
	}
	if _, err = tx.Exec(ctx, `UPDATE identity.accounts SET state='suspended', state_version=state_version+1, updated_at=$2 WHERE id=$1`, principalID, now.Add(time.Second)); err != nil {
		t.Fatal("suspend account fixture")
	}
	authority, err = task10LockRefreshAuthority(ctx, queries, refreshDigest)
	if err != nil || authority.account.State != "suspended" {
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
	currentRefresh := task10IntegrationDigest("current-refresh", principalID, currentSession)
	otherRefresh := task10IntegrationDigest("other-refresh", principalID, otherSession)
	createTask10SessionFixture(t, ctx, queries, principalID, currentSession, currentRefresh, 0x42, now)
	if err = queries.CreateAccountSession(ctx, store.CreateAccountSessionParams{
		ID: otherSession, PrincipalID: principalID, ClientSigningPublicKey: bytes.Repeat([]byte{0x43}, 32),
		AccessTokenHash: task10IntegrationDigest("other-access", principalID, otherSession), AccessExpiresAt: now.Add(10 * time.Minute),
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

	lockedRefresh, err := queries.LockPrincipalRefreshTokens(ctx, principalID)
	if err != nil {
		t.Fatal("lock principal refresh tokens")
	}
	lockedRefresh, err = queries.LockPrincipalRefreshTokens(ctx, principalID)
	if err != nil {
		t.Fatal("stabilize principal refresh tokens")
	}
	lockedSessions, err := queries.LockPrincipalAccountSessions(ctx, principalID)
	if err != nil {
		t.Fatal("lock principal sessions")
	}
	if _, err = queries.GetAccountForUpdate(ctx, principalID); err != nil {
		t.Fatal("lock principal account")
	}
	revoked, err := queries.RevokePrincipalAccountSessions(ctx, store.RevokePrincipalAccountSessionsParams{
		SessionIds: []uuid.UUID{otherSession}, TokenHashes: [][]byte{otherRefresh}, RevokedAt: now.Add(time.Second),
	})
	if err != nil || len(revoked) != 1 || revoked[0] != otherSession {
		t.Fatalf("revoked sessions = %v", revoked)
	}
	if len(lockedSessions) != 2 || len(lockedRefresh) != 2 {
		t.Fatal("principal authority collection was incomplete")
	}
	current, err := queries.ListSessionRefreshTokens(ctx, currentSession)
	if err != nil || len(current) != 1 || current[0].State != "active" {
		t.Fatal("current session was changed by others-scope revocation")
	}
	other, err := queries.ListSessionRefreshTokens(ctx, otherSession)
	if err != nil || len(other) != 1 || other[0].State != "revoked" {
		t.Fatal("other session and refresh were not revoked together")
	}
}

func TestAccountSessionLockHierarchyAvoidsCrossOperationDeadlocks(t *testing.T) {
	pool := testinfra.OpenMigratedPostgres(t)
	for _, operation := range []string{"enrollment", "change_password", "reset_password"} {
		t.Run("rotation_vs_"+operation, func(t *testing.T) {
			fixture := createTask10LockHierarchyFixture(t, pool, byte(len(operation)+0x71))
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			contender, err := pool.Begin(ctx)
			if err != nil {
				t.Fatal("begin contender transaction")
			}
			contenderClosed := false
			defer func() {
				if !contenderClosed {
					_ = contender.Rollback(context.Background())
				}
			}()
			contenderQueries := store.New(contender)
			if operation == "reset_password" {
				if _, err = contenderQueries.GetPasswordResetForUpdate(ctx, fixture.resetDigest); err != nil {
					t.Fatal("lock reset authority")
				}
			} else if _, err = contenderQueries.GetPasswordCredential(ctx, fixture.principalID); err != nil {
				t.Fatal("lock password authority")
			}
			if operation != "enrollment" {
				if _, err = contenderQueries.LockPrincipalRefreshTokens(ctx, fixture.principalID); err != nil {
					t.Fatal("lock principal refresh authority")
				}
				if _, err = contenderQueries.LockPrincipalRefreshTokens(ctx, fixture.principalID); err != nil {
					t.Fatal("stabilize principal refresh authority")
				}
			}
			locked, lockErr := contenderQueries.LockPrincipalAccountSessions(ctx, fixture.principalID)
			if lockErr != nil || len(locked) != 1 || locked[0].ID != fixture.sessionID {
				t.Fatal("lock stable principal sessions")
			}

			rotation := task10StartIntegrationRotation(t, ctx, pool, fixture.refreshDigest)
			select {
			case <-rotation:
				t.Fatal("rotation bypassed contender-held session lock")
			case <-time.After(150 * time.Millisecond):
			}
			if _, err = contenderQueries.GetAccountForUpdate(ctx, fixture.principalID); err != nil {
				t.Fatal("lock account after sessions")
			}
			now := fixture.now.Add(time.Second)
			switch operation {
			case "enrollment":
				err = contenderQueries.CreateEnrollmentGrant(ctx, store.CreateEnrollmentGrantParams{
					ID: uuid.New(), PrincipalID: fixture.principalID, AccountSessionID: fixture.sessionID,
					TokenHash: task10IntegrationDigest("enrollment-grant", fixture.principalID, fixture.sessionID), PolicyMarker: "standard", ExpiresAt: now.Add(10 * time.Minute), CreatedAt: now,
				})
			case "change_password":
				_, err = contenderQueries.UpdatePasswordCredential(ctx, task10UpdatedCredential(fixture.principalID, 0xd2, now))
				if err == nil {
					_, err = contenderQueries.MarkPrincipalSessionsReviewRequired(ctx, store.MarkPrincipalSessionsReviewRequiredParams{SessionIds: []uuid.UUID{fixture.sessionID}, UpdatedAt: now})
				}
			case "reset_password":
				_, err = contenderQueries.ConsumePasswordReset(ctx, store.ConsumePasswordResetParams{
					PrincipalID: fixture.principalID, ResetTokenHash: fixture.resetDigest, PolicyVersion: 1,
					MemoryKib: 65536, TimeCost: 3, Parallelism: 4, Salt: bytes.Repeat([]byte{0xd3}, 16),
					PasswordHash: bytes.Repeat([]byte{0xd4}, 32), ResetConsumedAt: sql.NullTime{Time: now, Valid: true},
				})
				if err == nil {
					_, err = contenderQueries.MarkPrincipalSessionsReviewRequired(ctx, store.MarkPrincipalSessionsReviewRequiredParams{SessionIds: []uuid.UUID{fixture.sessionID}, UpdatedAt: now})
				}
			}
			if err != nil {
				t.Fatal("apply contender mutation")
			}
			if err = contender.Commit(ctx); err != nil {
				t.Fatal("commit contender transaction")
			}
			contenderClosed = true
			outcome := task10AwaitIntegrationRotation(t, ctx, rotation)
			wantState := "active"
			if operation != "enrollment" {
				wantState = "review_required"
			}
			if outcome.err != nil || outcome.authority.session.State != wantState {
				t.Fatalf("rotation after %s = state %q / %v", operation, outcome.authority.session.State, outcome.err)
			}
		})
	}

	t.Run("rotation_vs_revoke", func(t *testing.T) {
		fixture := createTask10LockHierarchyFixture(t, pool, 0xe1)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		childDigest := task10IntegrationDigest("revoke-child", fixture.principalID, fixture.sessionID)
		if err := store.New(pool).InsertAccountRefreshToken(ctx, store.InsertAccountRefreshTokenParams{
			TokenHash: childDigest, SessionID: fixture.sessionID, PreviousTokenHash: fixture.refreshDigest,
			IssuedAt: fixture.now.Add(time.Second), IdleExpiresAt: fixture.now.Add(30 * 24 * time.Hour),
			AbsoluteExpiresAt: fixture.now.Add(90 * 24 * time.Hour),
		}); err != nil {
			t.Fatal("insert child refresh for principal revoke race")
		}
		rotationTx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal("begin rotation transaction")
		}
		rotationClosed := false
		defer func() {
			if !rotationClosed {
				_ = rotationTx.Rollback(context.Background())
			}
		}()
		if _, err = task10LockRefreshAuthority(ctx, store.New(rotationTx), childDigest); err != nil {
			t.Fatal("lock rotation authority hierarchy")
		}
		revocation := make(chan task10RevocationOutcome, 1)
		go func() {
			revokeTx, beginErr := pool.Begin(ctx)
			if beginErr != nil {
				revocation <- task10RevocationOutcome{err: beginErr}
				return
			}
			revokeQueries := store.New(revokeTx)
			lockedRefresh, revokeErr := revokeQueries.LockPrincipalRefreshTokens(ctx, fixture.principalID)
			if revokeErr == nil {
				lockedRefresh, revokeErr = revokeQueries.LockPrincipalRefreshTokens(ctx, fixture.principalID)
			}
			var lockedSessions []store.IdentityAccountSession
			if revokeErr == nil {
				lockedSessions, revokeErr = revokeQueries.LockPrincipalAccountSessions(ctx, fixture.principalID)
			}
			if revokeErr == nil {
				_, revokeErr = revokeQueries.GetAccountForUpdate(ctx, fixture.principalID)
			}
			var ids []uuid.UUID
			if revokeErr == nil {
				ids, revokeErr = revokeQueries.RevokePrincipalAccountSessions(ctx, store.RevokePrincipalAccountSessionsParams{
					SessionIds: task10SessionIDs(lockedSessions), TokenHashes: task10RefreshHashes(lockedRefresh), RevokedAt: fixture.now.Add(time.Second),
				})
			}
			if revokeErr == nil {
				revokeErr = revokeTx.Commit(ctx)
			} else {
				_ = revokeTx.Rollback(context.Background())
			}
			revocation <- task10RevocationOutcome{ids: ids, err: revokeErr}
		}()
		select {
		case <-revocation:
			t.Fatal("revocation bypassed rotation-held refresh hierarchy")
		case <-time.After(150 * time.Millisecond):
		}
		if err = rotationTx.Commit(ctx); err != nil {
			t.Fatal("commit rotation lock transaction")
		}
		rotationClosed = true
		select {
		case outcome := <-revocation:
			if outcome.err != nil || len(outcome.ids) != 1 || outcome.ids[0] != fixture.sessionID {
				t.Fatalf("revocation after rotation = %v / %v", outcome.ids, outcome.err)
			}
		case <-ctx.Done():
			t.Fatal("revocation deadlocked after rotation commit")
		}
	})
}

func TestUsedParentReplayAndActiveChildRotationShareFamilyPrelock(t *testing.T) {
	pool := testinfra.OpenMigratedPostgres(t)
	fixture := createTask10LockHierarchyFixture(t, pool, 0xf1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	childDigest := task10IntegrationDigest("used-parent-child", fixture.principalID, fixture.sessionID)
	queries := store.New(pool)
	if _, err := queries.MarkAccountRefreshUsed(ctx, store.MarkAccountRefreshUsedParams{
		TokenHash: fixture.refreshDigest, UsedAt: sql.NullTime{Time: fixture.now.Add(time.Second), Valid: true},
	}); err != nil {
		t.Fatal("mark parent refresh used")
	}
	if err := queries.InsertAccountRefreshToken(ctx, store.InsertAccountRefreshTokenParams{
		TokenHash: childDigest, SessionID: fixture.sessionID, PreviousTokenHash: fixture.refreshDigest,
		IssuedAt: fixture.now.Add(time.Second), IdleExpiresAt: fixture.now.Add(30 * 24 * time.Hour),
		AbsoluteExpiresAt: fixture.now.Add(90 * 24 * time.Hour),
	}); err != nil {
		t.Fatal("insert active refresh child")
	}

	replayTx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal("begin used-parent replay")
	}
	replayClosed := false
	defer func() {
		if !replayClosed {
			_ = replayTx.Rollback(context.Background())
		}
	}()
	replayQueries := store.New(replayTx)
	authority, err := task10LockRefreshAuthority(ctx, replayQueries, fixture.refreshDigest)
	if err != nil || len(authority.refreshes) != 2 {
		t.Fatal("used replay did not prelock the complete refresh family")
	}
	childRotation := task10StartIntegrationRotation(t, ctx, pool, childDigest)
	select {
	case outcome := <-childRotation:
		t.Fatalf("child rotation bypassed family prelock: %v", outcome.err)
	case <-time.After(150 * time.Millisecond):
	}
	if _, err = replayQueries.MarkAccountSessionCompromised(ctx, store.MarkAccountSessionCompromisedParams{
		ID: fixture.sessionID, UpdatedAt: fixture.now.Add(2 * time.Second),
	}); err != nil {
		t.Fatal("mark replayed session compromised")
	}
	if _, err = replayQueries.RevokeAccountRefreshTokens(ctx, store.RevokeAccountRefreshTokensParams{
		TokenHashes: task10RefreshHashes(authority.refreshes), RevokedAt: sql.NullTime{Time: fixture.now.Add(2 * time.Second), Valid: true},
	}); err != nil {
		t.Fatal("revoke prelocked refresh family")
	}
	if err = replayTx.Commit(ctx); err != nil {
		t.Fatal("commit used-parent compromise")
	}
	replayClosed = true
	outcome := task10AwaitIntegrationRotation(t, ctx, childRotation)
	if outcome.err != nil || outcome.authority.session.State != "compromised" || len(outcome.authority.refreshes) != 2 {
		t.Fatalf("child rotation after compromise = %#v / %v", outcome.authority, outcome.err)
	}
	for _, refresh := range outcome.authority.refreshes {
		if refresh.State == "active" {
			t.Fatal("active refresh child survived used-parent replay")
		}
	}
}

type task10LockHierarchyFixture struct {
	principalID, sessionID     uuid.UUID
	refreshDigest, resetDigest []byte
	now                        time.Time
}

type task10RotationOutcome struct {
	authority task10IntegrationRefreshAuthority
	err       error
}

type task10IntegrationRefreshAuthority struct {
	refreshes []store.IdentityAccountRefreshToken
	session   store.IdentityAccountSession
	account   store.IdentityAccount
}

type task10RevocationOutcome struct {
	ids []uuid.UUID
	err error
}

func createTask10LockHierarchyFixture(t *testing.T, pool interface {
	store.DBTX
}, discriminator byte) task10LockHierarchyFixture {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	principalID := uuid.New()
	sessionID := uuid.New()
	fixture := task10LockHierarchyFixture{
		principalID: principalID, sessionID: sessionID, refreshDigest: task10IntegrationDigest("fixture-refresh", principalID, sessionID),
		resetDigest: task10IntegrationDigest("fixture-reset", principalID), now: now,
	}
	queries := store.New(pool)
	createTask10SessionFixture(t, context.Background(), queries, fixture.principalID, fixture.sessionID, fixture.refreshDigest, discriminator+2, now)
	if err := queries.CreatePasswordCredential(context.Background(), store.CreatePasswordCredentialParams{
		PrincipalID: fixture.principalID, PolicyVersion: 1, MemoryKib: 65536, TimeCost: 3, Parallelism: 4,
		Salt: bytes.Repeat([]byte{discriminator + 3}, 16), PasswordHash: bytes.Repeat([]byte{discriminator + 4}, 32), UpdatedAt: now,
	}); err != nil {
		t.Fatal("create lock hierarchy password credential")
	}
	if _, err := queries.SetPasswordReset(context.Background(), store.SetPasswordResetParams{
		PrincipalID: fixture.principalID, ResetTokenHash: fixture.resetDigest,
		ResetExpiresAt: sql.NullTime{Time: now.Add(30 * time.Minute), Valid: true}, ResetDeliveryID: uuid.NullUUID{UUID: uuid.New(), Valid: true},
		ResetDeliveryCiphertext: bytes.Repeat([]byte{discriminator + 5}, 29), ResetDeliveryKeyVersion: pgtype.Int4{Int32: 1, Valid: true}, UpdatedAt: now,
	}); err != nil {
		t.Fatal("create lock hierarchy reset authority")
	}
	t.Cleanup(func() {
		cleanupContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pool.Exec(cleanupContext, `DELETE FROM deviceauth.enrollment_grants WHERE principal_id=$1`, fixture.principalID)
		_, _ = pool.Exec(cleanupContext, `DELETE FROM identity.account_refresh_tokens WHERE session_id=$1`, fixture.sessionID)
		_, _ = pool.Exec(cleanupContext, `DELETE FROM identity.account_sessions WHERE principal_id=$1`, fixture.principalID)
		_, _ = pool.Exec(cleanupContext, `DELETE FROM identity.password_credentials WHERE principal_id=$1`, fixture.principalID)
		_, _ = pool.Exec(cleanupContext, `DELETE FROM identity.accounts WHERE id=$1`, fixture.principalID)
	})
	return fixture
}

func task10StartIntegrationRotation(t *testing.T, ctx context.Context, pool interface {
	Begin(context.Context) (pgx.Tx, error)
}, digest []byte) <-chan task10RotationOutcome {
	t.Helper()
	result := make(chan task10RotationOutcome, 1)
	go func() {
		tx, err := pool.Begin(ctx)
		if err != nil {
			result <- task10RotationOutcome{err: err}
			return
		}
		authority, queryErr := task10LockRefreshAuthority(ctx, store.New(tx), digest)
		if queryErr == nil {
			queryErr = tx.Commit(ctx)
		} else {
			_ = tx.Rollback(context.Background())
		}
		result <- task10RotationOutcome{authority: authority, err: queryErr}
	}()
	return result
}

func task10LockRefreshAuthority(ctx context.Context, queries *store.Queries, digest []byte) (task10IntegrationRefreshAuthority, error) {
	discovered, err := queries.DiscoverRefreshToken(ctx, digest)
	if err != nil {
		return task10IntegrationRefreshAuthority{}, err
	}
	if _, err = queries.LockSessionRefreshTokens(ctx, discovered.SessionID); err != nil {
		return task10IntegrationRefreshAuthority{}, err
	}
	refreshes, err := queries.LockSessionRefreshTokens(ctx, discovered.SessionID)
	if err != nil {
		return task10IntegrationRefreshAuthority{}, err
	}
	sessions, err := queries.LockPrincipalAccountSessions(ctx, discovered.PrincipalID)
	if err != nil {
		return task10IntegrationRefreshAuthority{}, err
	}
	var session store.IdentityAccountSession
	for _, candidate := range sessions {
		if candidate.ID == discovered.SessionID {
			session = candidate
			break
		}
	}
	if session.ID == uuid.Nil {
		return task10IntegrationRefreshAuthority{}, pgx.ErrNoRows
	}
	account, err := queries.GetAccountForUpdate(ctx, discovered.PrincipalID)
	if err != nil {
		return task10IntegrationRefreshAuthority{}, err
	}
	current, err := queries.ListSessionRefreshTokens(ctx, discovered.SessionID)
	if err != nil || !task10SameRefreshSet(refreshes, current) {
		if err == nil {
			err = pgx.ErrNoRows
		}
		return task10IntegrationRefreshAuthority{}, err
	}
	return task10IntegrationRefreshAuthority{refreshes: refreshes, session: session, account: account}, nil
}

func task10SameRefreshSet(left, right []store.IdentityAccountRefreshToken) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].SessionID != right[index].SessionID || !bytes.Equal(left[index].TokenHash, right[index].TokenHash) {
			return false
		}
	}
	return true
}

func task10SessionIDs(rows []store.IdentityAccountSession) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.ID)
	}
	return result
}

func task10RefreshHashes(rows []store.IdentityAccountRefreshToken) [][]byte {
	result := make([][]byte, 0, len(rows))
	for _, row := range rows {
		result = append(result, bytes.Clone(row.TokenHash))
	}
	return result
}

func task10AwaitIntegrationRotation(t *testing.T, ctx context.Context, result <-chan task10RotationOutcome) task10RotationOutcome {
	t.Helper()
	select {
	case outcome := <-result:
		return outcome
	case <-ctx.Done():
		t.Fatal("rotation deadlocked after contender commit")
		return task10RotationOutcome{}
	}
}

func task10UpdatedCredential(principalID uuid.UUID, discriminator byte, now time.Time) store.UpdatePasswordCredentialParams {
	return store.UpdatePasswordCredentialParams{
		PrincipalID: principalID, PolicyVersion: 1, MemoryKib: 65536, TimeCost: 3, Parallelism: 4,
		Salt: bytes.Repeat([]byte{discriminator}, 16), PasswordHash: bytes.Repeat([]byte{discriminator + 1}, 32), UpdatedAt: now,
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
		AccessTokenHash: task10IntegrationDigest("fixture-access", principalID, sessionID), AccessExpiresAt: now.Add(10 * time.Minute),
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

func task10IntegrationDigest(label string, ids ...uuid.UUID) []byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte(label))
	for _, id := range ids {
		_, _ = hash.Write(id[:])
	}
	return hash.Sum(nil)
}
