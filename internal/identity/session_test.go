package identity

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/ratelimit"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

var fixedTask10Time = time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)

func TestSessionPasswordLoginCreatesIsolatedBoundTokens(t *testing.T) {
	transaction := activeTask10Transaction()
	application, deriver, limiter := newTask10Application(t, transaction, &task10ChallengeStore{})
	publicKey := task10PublicKey()
	tokens, err := application.CreateSession(context.Background(), CreateSessionCommand{
		Method: SessionPassword, Email: "member@EXAMPLE.test", Password: secret.NewBytes([]byte("correct horse battery staple")),
		ClientSigningPublicKey: publicKey, IdempotencyKey: "abcdefghijklmnopqrstuv",
	})
	if err != nil {
		t.Fatal(err)
	}
	if deriver.calls != 1 || limiter.calls != 1 || limiter.operation != ratelimit.Login {
		t.Fatalf("password/limiter calls = %d/%d (%q)", deriver.calls, limiter.calls, limiter.operation)
	}
	if tokens.AccessExpiresAt != fixedTask10Time.Add(10*time.Minute) ||
		tokens.RefreshIdleExpiresAt != fixedTask10Time.Add(30*24*time.Hour) ||
		tokens.RefreshAbsoluteExpiresAt != fixedTask10Time.Add(90*24*time.Hour) {
		t.Fatalf("session expiry contract = %#v", tokens)
	}
	if !bytes.Equal(transaction.createdSession.ClientSigningPublicKey, publicKey[:]) {
		t.Fatal("session did not persist the client Ed25519 public key")
	}
	accessDigest := securitykit.DigestToken(securitykit.AccountAccessToken, tokens.AccessToken)
	deviceDigest := securitykit.DigestToken(securitykit.DeviceAccessToken, tokens.AccessToken)
	if accessDigest == [32]byte{} || accessDigest == deviceDigest || !bytes.Equal(transaction.createdSession.AccessTokenHash, accessDigest[:]) {
		t.Fatal("account access token was not isolated in its own digest domain")
	}
	refreshDigest := securitykit.DigestToken(securitykit.AccountRefreshToken, tokens.RefreshToken)
	if !bytes.Equal(transaction.createdRefresh.TokenHash, refreshDigest[:]) {
		t.Fatal("account refresh token was not persisted in its isolated domain")
	}
	if got := transaction.operations; !containsTask10Order(got, []string{"find_identity", "get_account", "get_credential", "begin_idempotency", "create_account_session", "insert_account_refresh", "complete_idempotency", "commit"}) {
		t.Fatalf("login operation order = %v", got)
	}
	var authenticator AccountAuthenticator = application
	authority, err := authenticator.Authenticate(context.Background(), tokens.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if authority.PrincipalID != PrincipalID(transaction.account.ID.String()) || authority.SessionID != SessionID(transaction.createdSession.ID.String()) {
		t.Fatalf("account authority = %#v", authority)
	}
	if _, err = authenticator.Authenticate(context.Background(), secret.NewBytes(bytes.Repeat([]byte{0x9d}, 32))); publicTask10Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("unknown access error = %v", err)
	}
	deviceRaw := secret.NewBytes(bytes.Repeat([]byte{0x7c}, 32))
	transaction.accessFound = false
	if _, err = authenticator.Authenticate(context.Background(), deviceRaw); publicTask10Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("device-domain token error = %v", err)
	}
}

func TestSessionPasswordFailuresShareAuthenticationErrorAndConstantWork(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*task10Transaction)
	}{
		{name: "unknown email", mutate: func(tx *task10Transaction) {
			tx.identityFound = false
			tx.accountFound = false
			tx.credentialFound = false
		}},
		{name: "wrong password", mutate: func(tx *task10Transaction) { tx.credential.PasswordHash = bytes.Repeat([]byte{0x44}, 32) }},
		{name: "suspended account", mutate: func(tx *task10Transaction) { tx.account.State = "suspended" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transaction := activeTask10Transaction()
			test.mutate(transaction)
			application, deriver, limiter := newTask10Application(t, transaction, &task10ChallengeStore{})
			_, err := application.CreateSession(context.Background(), task10CreateSessionCommand())
			if publicTask10Code(err) != apierrors.AuthenticationFailed {
				t.Fatalf("public error = %v", err)
			}
			if limiter.calls != 1 || deriver.calls != 1 || transaction.createdSession.ID != uuid.Nil {
				t.Fatalf("limiter/Argon/session = %d/%d/%v", limiter.calls, deriver.calls, transaction.createdSession.ID)
			}
		})
	}
}

func TestSessionLoginLimiterFailsBeforeArgonWithFiniteErrors(t *testing.T) {
	tests := []struct {
		name     string
		allowed  bool
		limitErr error
		want     apierrors.Code
	}{
		{name: "denied", allowed: false, want: apierrors.RateLimited},
		{name: "ambiguous", limitErr: errors.New("REDIS-CANARY"), want: apierrors.DependencyUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transaction := activeTask10Transaction()
			application, deriver, limiter := newTask10Application(t, transaction, &task10ChallengeStore{})
			limiter.allowed = test.allowed
			limiter.err = test.limitErr
			_, err := application.CreateSession(context.Background(), task10CreateSessionCommand())
			if publicTask10Code(err) != test.want {
				t.Fatalf("public error = %v", err)
			}
			if limiter.calls != 1 || deriver.calls != 0 || transaction.createdSession.ID != uuid.Nil {
				t.Fatal("rate-limit decision did not precede password work")
			}
		})
	}
}

func TestRedisOutageClosesLoginAndRotationButExistingAccessStillAuthenticates(t *testing.T) {
	transaction := activeTask10Transaction()
	refresh := secret.NewBytes(bytes.Repeat([]byte{0x51}, 32))
	publicKey := task10PublicKey()
	configureTask10Refresh(transaction, refresh, publicKey, "active")
	challenges := &task10ChallengeStore{}
	application, _, limiter := newTask10Application(t, transaction, challenges)

	tokens, err := application.CreateSession(context.Background(), task10CreateSessionCommand())
	if err != nil {
		t.Fatal(err)
	}
	rotate := task10RotateCommand(t, application, refresh, transaction.refresh.SessionID, "bcdefghijklmnopqrstuvw")
	limiter.err = errors.New("REDIS-LIMITER-CANARY")
	challenges.consumeErr = errors.New("REDIS-CHALLENGE-CANARY")

	if _, err = application.CreateSession(context.Background(), task10CreateSessionCommand()); publicTask10Code(err) != apierrors.DependencyUnavailable {
		t.Fatalf("login during Redis outage error = %v", err)
	}
	if _, err = application.RotateSession(context.Background(), rotate); publicTask10Code(err) != apierrors.DependencyUnavailable {
		t.Fatalf("rotation during Redis outage error = %v", err)
	}
	authority, err := application.Authenticate(context.Background(), tokens.AccessToken)
	if err != nil || authority.PrincipalID != PrincipalID(transaction.account.ID.String()) {
		t.Fatalf("existing access during Redis outage = %#v / %v", authority, err)
	}
}

func TestTask9MethodsRemainIndependentOfOptionalChallengeStore(t *testing.T) {
	transaction := &fakeIdentityTransaction{}
	application, _, _, _ := newTask9Application(t, config.EmailDisabled, transaction)
	_, err := application.CreateSession(context.Background(), task10CreateSessionCommand())
	if publicTask10Code(err) != apierrors.DependencyUnavailable {
		t.Fatalf("session without challenge store error = %v", err)
	}
	transaction = &fakeIdentityTransaction{}
	application, _, _, _ = newTask9Application(t, config.EmailDisabled, transaction)
	_, err = application.RegisterAccount(context.Background(), task9RegisterCommand())
	if err != nil {
		t.Fatalf("Task9 registration gained Redis dependency: %v", err)
	}
}

func TestSessionChallengeAuthenticatesAuthorityBeforeRedis(t *testing.T) {
	transaction := activeTask10Transaction()
	refresh := secret.NewBytes(bytes.Repeat([]byte{0x51}, 32))
	configureTask10Refresh(transaction, refresh, task10PublicKey(), "active")
	challenges := &task10ChallengeStore{}
	application, _, limiter := newTask10Application(t, transaction, challenges)
	challenge, err := application.CreateSessionChallenge(context.Background(), CreateSessionChallengeCommand{
		RefreshToken: refresh, RequestNonce: [32]byte{9}, IdempotencyKey: "abcdefghijklmnopqrstuv",
	})
	if err != nil {
		t.Fatal(err)
	}
	if challenge.ChallengeID == "" || challenge.Challenge == [32]byte{} || challenge.ExpiresAt != fixedTask10Time.Add(2*time.Minute) {
		t.Fatalf("challenge = %#v", challenge)
	}
	if limiter.calls != 1 || limiter.operation != ratelimit.Challenge || challenges.createCalls != 1 {
		t.Fatal("authority challenge did not apply the bounded Redis dependencies")
	}
	if !containsTask10Order(transaction.operations, []string{"get_refresh", "begin_idempotency", "rollback", "get_refresh", "begin_idempotency", "complete_idempotency", "commit"}) || challenges.createdAtOperationCount < 3 {
		t.Fatalf("authority/store order = %v / %d", transaction.operations, challenges.createdAtOperationCount)
	}

	transaction = activeTask10Transaction()
	transaction.refreshFound = false
	challenges = &task10ChallengeStore{}
	application, _, _ = newTask10Application(t, transaction, challenges)
	_, err = application.CreateSessionChallenge(context.Background(), CreateSessionChallengeCommand{
		RefreshToken: refresh, RequestNonce: [32]byte{9}, IdempotencyKey: "abcdefghijklmnopqrstuv",
	})
	if publicTask10Code(err) != apierrors.AuthenticationFailed || challenges.createCalls != 0 {
		t.Fatalf("unknown refresh challenge error/calls = %v/%d", err, challenges.createCalls)
	}
}

func TestSessionRotationProofBindsEveryField(t *testing.T) {
	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{0x73}, ed25519.SeedSize))
	defer clear(private)
	input := AccountRotationProofInput{
		ProtocolVersion: "account-rotation-v1", Challenge: [32]byte{1, 2, 3},
		SessionID: uuid.MustParse("d12dca8a-ced3-471d-a6ad-55b228221f10"), Operation: "rotate_account_token",
		Audience: "https://api.example.test", RequestNonce: [32]byte{4, 5, 6},
	}
	proof := AccountRotationProofBytes(input)
	if !bytes.HasPrefix(proof, []byte("TALENRO-ACCOUNT-ROTATION-V1\x00")) {
		t.Fatal("rotation proof omitted the exact domain prefix")
	}
	signature := ed25519.Sign(private, proof)
	mutations := []AccountRotationProofInput{
		{ProtocolVersion: "account-rotation-v2", Challenge: input.Challenge, SessionID: input.SessionID, Operation: input.Operation, Audience: input.Audience, RequestNonce: input.RequestNonce},
		{ProtocolVersion: input.ProtocolVersion, Challenge: [32]byte{9}, SessionID: input.SessionID, Operation: input.Operation, Audience: input.Audience, RequestNonce: input.RequestNonce},
		{ProtocolVersion: input.ProtocolVersion, Challenge: input.Challenge, SessionID: uuid.MustParse("6bb12bea-a5e6-49f6-a225-3550b0b4f6c5"), Operation: input.Operation, Audience: input.Audience, RequestNonce: input.RequestNonce},
		{ProtocolVersion: input.ProtocolVersion, Challenge: input.Challenge, SessionID: input.SessionID, Operation: "other_operation", Audience: input.Audience, RequestNonce: input.RequestNonce},
		{ProtocolVersion: input.ProtocolVersion, Challenge: input.Challenge, SessionID: input.SessionID, Operation: input.Operation, Audience: "https://other.example.test", RequestNonce: input.RequestNonce},
		{ProtocolVersion: input.ProtocolVersion, Challenge: input.Challenge, SessionID: input.SessionID, Operation: input.Operation, Audience: input.Audience, RequestNonce: [32]byte{8}},
	}
	for index, mutation := range mutations {
		if ed25519.Verify(private.Public().(ed25519.PublicKey), AccountRotationProofBytes(mutation), signature) {
			t.Fatalf("field mutation %d retained a valid signature", index)
		}
	}
}

func TestSessionRotationConsumesChallengeAndRefresh(t *testing.T) {
	transaction := activeTask10Transaction()
	refresh := secret.NewBytes(bytes.Repeat([]byte{0x51}, 32))
	publicKey := task10PublicKey()
	configureTask10Refresh(transaction, refresh, publicKey, "active")
	challenges := &task10ChallengeStore{}
	application, _, _ := newTask10Application(t, transaction, challenges)
	command := task10RotateCommand(t, application, refresh, transaction.refresh.SessionID, "abcdefghijklmnopqrstuv")
	tokens, err := application.RotateSession(context.Background(), command)
	if err != nil {
		t.Fatal(err)
	}
	if challenges.consumeCalls != 1 || transaction.refresh.RefreshState != "used" || tokens.AccessExpiresAt != fixedTask10Time.Add(10*time.Minute) {
		t.Fatal("rotation did not consume both one-time authorities")
	}
	if !containsTask10Order(transaction.operations, []string{"get_refresh", "commit", "get_refresh", "begin_idempotency", "rollback", "get_refresh", "begin_idempotency", "mark_refresh_used", "rotate_session_access", "insert_account_refresh", "complete_idempotency", "commit"}) {
		t.Fatalf("rotation operation order = %v", transaction.operations)
	}
}

func TestSessionRotationReplayAcceptsCappedExpiryRelationships(t *testing.T) {
	application, _, _ := newTask10Application(t, activeTask10Transaction(), &task10ChallengeStore{})
	remaining := []time.Duration{24 * time.Hour, 5 * time.Minute}
	for _, lifetime := range remaining {
		t.Run(lifetime.String(), func(t *testing.T) {
			created, err := application.newRotatedSessionTokens(fixedTask10Time, fixedTask10Time.Add(lifetime))
			if err != nil {
				t.Fatal(err)
			}
			if created.tokens.AccessExpiresAt.After(created.tokens.RefreshAbsoluteExpiresAt) ||
				created.tokens.RefreshIdleExpiresAt.After(created.tokens.RefreshAbsoluteExpiresAt) {
				t.Fatalf("rotation expiries exceed absolute deadline: %#v", created.tokens)
			}
			body, err := encodeSessionTokens(created.tokens)
			if err != nil {
				t.Fatal(err)
			}
			defer clear(body)
			replayed, err := decodeSessionTokens(body)
			if err != nil {
				t.Fatalf("committed rotation could not replay: %v", err)
			}
			if replayed.AccessExpiresAt != created.tokens.AccessExpiresAt || replayed.RefreshIdleExpiresAt != created.tokens.RefreshIdleExpiresAt ||
				replayed.RefreshAbsoluteExpiresAt != created.tokens.RefreshAbsoluteExpiresAt {
				t.Fatalf("replayed expiries differ: %#v / %#v", replayed, created.tokens)
			}
		})
	}
}

func TestUsedRefreshReplayCommitsCompromiseBeforePublicFailure(t *testing.T) {
	transaction := activeTask10Transaction()
	refresh := secret.NewBytes(bytes.Repeat([]byte{0x51}, 32))
	publicKey := task10PublicKey()
	configureTask10Refresh(transaction, refresh, publicKey, "active")
	challenges := &task10ChallengeStore{}
	application, _, _ := newTask10Application(t, transaction, challenges)
	command := task10RotateCommand(t, application, refresh, transaction.refresh.SessionID, "abcdefghijklmnopqrstuv")
	transaction.refresh.RefreshState = "used"
	_, err := application.RotateSession(context.Background(), command)
	if publicTask10Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("used refresh public error = %v", err)
	}
	if !containsTask10Order(transaction.operations, []string{"mark_session_compromised", "revoke_refresh", "insert_security_event", "complete_idempotency", "commit"}) {
		t.Fatalf("used refresh operation order = %v", transaction.operations)
	}
}

func TestExpiredUsedRefreshReplayStillCompromisesLiveFamily(t *testing.T) {
	transaction := activeTask10Transaction()
	refresh := secret.NewBytes(bytes.Repeat([]byte{0x51}, 32))
	publicKey := task10PublicKey()
	configureTask10Refresh(transaction, refresh, publicKey, "active")
	challenges := &task10ChallengeStore{}
	application, _, _ := newTask10Application(t, transaction, challenges)
	command := task10RotateCommand(t, application, refresh, transaction.refresh.SessionID, "abcdefghijklmnopqrstuv")
	transaction.refresh.RefreshState = "used"
	transaction.refresh.IdleExpiresAt = fixedTask10Time.Add(-time.Nanosecond)
	transaction.refresh.AbsoluteExpiresAt = fixedTask10Time.Add(60 * 24 * time.Hour)
	_, err := application.RotateSession(context.Background(), command)
	if publicTask10Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("expired used refresh public error = %v", err)
	}
	if !containsTask10Order(transaction.operations, []string{"mark_session_compromised", "revoke_refresh", "insert_security_event", "commit"}) {
		t.Fatalf("expired used replay escaped family compromise: %v", transaction.operations)
	}
}

func TestConcurrentAccountRefreshHasExactlyOneSuccess(t *testing.T) {
	transaction := activeTask10Transaction()
	refresh := secret.NewBytes(bytes.Repeat([]byte{0x51}, 32))
	publicKey := task10PublicKey()
	configureTask10Refresh(transaction, refresh, publicKey, "active")
	challenges := &task10ChallengeStore{}
	application, _, _ := newTask10Application(t, transaction, challenges)
	commands := []RotateSessionCommand{
		task10RotateCommand(t, application, refresh, transaction.refresh.SessionID, "abcdefghijklmnopqrstuv"),
		task10RotateCommand(t, application, refresh, transaction.refresh.SessionID, "bcdefghijklmnopqrstuvw"),
	}
	var successes atomic.Int32
	var wait sync.WaitGroup
	for _, command := range commands {
		command := command
		wait.Add(1)
		go func() {
			defer wait.Done()
			if _, err := application.RotateSession(context.Background(), command); err == nil {
				successes.Add(1)
			}
		}()
	}
	wait.Wait()
	if successes.Load() != 1 {
		t.Fatalf("rotation successes = %d", successes.Load())
	}
}

func TestRevokeSessionsUsesOneGeneratedPrincipalScopeTransition(t *testing.T) {
	transaction := activeTask10Transaction()
	currentSession := uuid.MustParse("d12dca8a-ced3-471d-a6ad-55b228221f10")
	application, _, _ := newTask10Application(t, transaction, &task10ChallengeStore{})
	err := application.RevokeSessions(context.Background(), RevokeSessionsCommand{
		PrincipalID: PrincipalID(transaction.account.ID.String()), Scope: RevokeOtherSessions,
		SessionID: SessionID(currentSession.String()), IdempotencyKey: "abcdefghijklmnopqrstuv",
	})
	if err != nil {
		t.Fatal(err)
	}
	if transaction.revokedParams.PrincipalID != transaction.account.ID || transaction.revokedParams.RevokeScope != "others" || transaction.revokedParams.SessionID != currentSession {
		t.Fatalf("generated revoke params = %#v", transaction.revokedParams)
	}
	if !containsTask10Order(transaction.operations, []string{"begin_idempotency", "revoke_sessions", "complete_idempotency", "commit"}) {
		t.Fatalf("revoke operation order = %v", transaction.operations)
	}
}

func TestChangePasswordVerifiesInlineReauthenticationAndReviewsOldSessions(t *testing.T) {
	transaction := activeTask10Transaction()
	currentSession := uuid.MustParse("d12dca8a-ced3-471d-a6ad-55b228221f10")
	currentPublicKey := task10PublicKey()
	transaction.sessionFound = true
	transaction.session = store.IdentityAccountSession{
		ID: currentSession, PrincipalID: transaction.account.ID, State: "active", StateVersion: 2,
		ClientSigningPublicKey: currentPublicKey[:], AccessTokenHash: bytes.Repeat([]byte{0x41}, 32),
		AccessExpiresAt: fixedTask10Time.Add(5 * time.Minute), AbsoluteExpiresAt: fixedTask10Time.Add(24 * time.Hour),
		CreatedAt: fixedTask10Time.Add(-time.Hour), UpdatedAt: fixedTask10Time.Add(-time.Hour),
	}
	application, deriver, _ := newTask10Application(t, transaction, &task10ChallengeStore{})
	currentPassword := []byte("correct horse battery staple")
	tokens, err := application.ChangePassword(context.Background(), ChangePasswordCommand{
		PrincipalID: PrincipalID(transaction.account.ID.String()), CurrentPassword: secret.NewBytes(currentPassword),
		NewPassword:            secret.NewBytes([]byte("new correct horse battery staple")),
		Reauthentication:       Reauthentication{SessionID: SessionID(currentSession.String()), Method: ReauthPassword, Proof: secret.NewBytes(currentPassword)},
		ClientSigningPublicKey: task10PublicKey(), IdempotencyKey: "abcdefghijklmnopqrstuv",
	})
	if err != nil {
		t.Fatal(err)
	}
	if deriver.calls != 2 || transaction.createdSession.ID == uuid.Nil || transaction.createdSession.ID == currentSession || tokens.AccessExpiresAt != fixedTask10Time.Add(10*time.Minute) {
		t.Fatal("password change did not verify once and create one replacement session")
	}
	if !containsTask10Order(transaction.operations, []string{
		"get_session", "get_account", "begin_idempotency", "get_credential", "update_password", "mark_sessions_review_required",
		"create_account_session", "insert_account_refresh", "insert_security_event", "complete_idempotency", "commit",
	}) {
		t.Fatalf("password change order = %v", transaction.operations)
	}
}

func TestChangePasswordRejectsMismatchedOrUnsupportedReauthentication(t *testing.T) {
	tests := []struct {
		name   string
		method ReauthMethod
		proof  string
		want   apierrors.Code
		calls  int
	}{
		{name: "mismatched password", method: ReauthPassword, proof: "different horse battery staple", want: apierrors.AuthenticationFailed, calls: 1},
		{name: "passkey deferred", method: ReauthPasskey, proof: "correct horse battery staple", want: apierrors.ActionNotAllowed, calls: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transaction := activeTask10Transaction()
			currentSession := uuid.MustParse("d12dca8a-ced3-471d-a6ad-55b228221f10")
			transaction.sessionFound = true
			transaction.session = store.IdentityAccountSession{
				ID: currentSession, PrincipalID: transaction.account.ID, State: "active", StateVersion: 2,
				ClientSigningPublicKey: bytes.Repeat([]byte{0x31}, 32), AccessTokenHash: bytes.Repeat([]byte{0x41}, 32),
				AccessExpiresAt: fixedTask10Time.Add(5 * time.Minute), AbsoluteExpiresAt: fixedTask10Time.Add(24 * time.Hour),
				CreatedAt: fixedTask10Time.Add(-time.Hour), UpdatedAt: fixedTask10Time.Add(-time.Hour),
			}
			application, deriver, _ := newTask10Application(t, transaction, &task10ChallengeStore{})
			_, err := application.ChangePassword(context.Background(), ChangePasswordCommand{
				PrincipalID: PrincipalID(transaction.account.ID.String()), CurrentPassword: secret.NewBytes([]byte("correct horse battery staple")),
				NewPassword:            secret.NewBytes([]byte("new correct horse battery staple")),
				Reauthentication:       Reauthentication{SessionID: SessionID(currentSession.String()), Method: test.method, Proof: secret.NewBytes([]byte(test.proof))},
				ClientSigningPublicKey: task10PublicKey(), IdempotencyKey: "abcdefghijklmnopqrstuv",
			})
			if publicTask10Code(err) != test.want || deriver.calls != test.calls || transaction.createdSession.ID != uuid.Nil {
				t.Fatalf("error/deriver/session = %v/%d/%v", err, deriver.calls, transaction.createdSession.ID)
			}
		})
	}
}

func TestSessionPrivateTypesRedactFormattingAndForbidDirectJSON(t *testing.T) {
	secretCanary := []byte("SESSION-SECRET-CANARY-0123456789")
	principal := PrincipalID("28ceee8a-5f4f-4d3a-9e9f-d3ec23b815ac")
	session := SessionID("d12dca8a-ced3-471d-a6ad-55b228221f10")
	challenge := [32]byte{}
	copy(challenge[:], secretCanary)
	values := []any{
		CreateSessionCommand{Method: SessionPassword, Email: "canary@example.test", Password: secret.NewBytes(secretCanary), WebAuthnResponse: json.RawMessage(`{"canary":"SESSION-SECRET-CANARY"}`)},
		RotateSessionCommand{RefreshToken: secret.NewBytes(secretCanary), ChallengeID: "d6df0ff2-b723-4b62-bc43-f63cc1bf012a", RequestNonce: challenge},
		CreateSessionChallengeCommand{RefreshToken: secret.NewBytes(secretCanary), RequestNonce: challenge},
		ChangePasswordCommand{PrincipalID: principal, CurrentPassword: secret.NewBytes(secretCanary), NewPassword: secret.NewBytes(secretCanary), Reauthentication: Reauthentication{SessionID: session, Proof: secret.NewBytes(secretCanary)}},
		RevokeSessionsCommand{PrincipalID: principal, SessionID: session},
		SessionChallenge{ChallengeID: "d6df0ff2-b723-4b62-bc43-f63cc1bf012a", Challenge: challenge},
		AccountAuthority{PrincipalID: principal, SessionID: session},
		AccountRotationProofInput{Challenge: challenge, SessionID: uuid.MustParse(string(session)), Audience: "https://target-canary.example"},
		task10ChallengeRecord(),
	}
	for _, value := range values {
		rendered := fmt.Sprintf("%+v", value) + slog.Any("session", value).Value.String()
		for _, forbidden := range []string{"SESSION-SECRET-CANARY", "canary@example", string(principal), string(session), "target-canary"} {
			if strings.Contains(rendered, forbidden) {
				t.Fatalf("private session type exposed input: %s", rendered)
			}
		}
		if encoded, err := json.Marshal(value); err == nil || encoded != nil {
			t.Fatalf("private session type unexpectedly marshaled: %s", encoded)
		}
	}
}

func task10RotateCommand(t *testing.T, application *Service, refresh secret.Bytes, sessionID uuid.UUID, idempotencyKey string) RotateSessionCommand {
	t.Helper()
	nonce := [32]byte{7, 8, 9}
	challenge, err := application.CreateSessionChallenge(context.Background(), CreateSessionChallengeCommand{
		RefreshToken: refresh, RequestNonce: nonce, IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		t.Fatal(err)
	}
	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{0x73}, ed25519.SeedSize))
	defer clear(private)
	proof := AccountRotationProofBytes(AccountRotationProofInput{
		ProtocolVersion: "account-rotation-v1", Challenge: challenge.Challenge, SessionID: sessionID,
		Operation: "rotate_account_token", Audience: "https://api.example.test", RequestNonce: nonce,
	})
	signed := ed25519.Sign(private, proof)
	var signature [64]byte
	copy(signature[:], signed)
	clear(signed)
	return RotateSessionCommand{
		RefreshToken: refresh, ChallengeID: challenge.ChallengeID, RequestNonce: nonce, Signature: signature, IdempotencyKey: idempotencyKey,
	}
}

func configureTask10Refresh(transaction *task10Transaction, token secret.Bytes, publicKey [32]byte, state string) {
	digest := securitykit.DigestToken(securitykit.AccountRefreshToken, token)
	transaction.refreshDigest = append([]byte(nil), digest[:]...)
	transaction.refreshFound = true
	transaction.refresh = store.GetRefreshTokenForUpdateRow{
		TokenHash: append([]byte(nil), digest[:]...), SessionID: uuid.MustParse("d12dca8a-ced3-471d-a6ad-55b228221f10"),
		RefreshState: state, IssuedAt: fixedTask10Time.Add(-time.Hour), IdleExpiresAt: fixedTask10Time.Add(time.Hour),
		AbsoluteExpiresAt: fixedTask10Time.Add(24 * time.Hour), SessionState: "active", SessionStateVersion: 3,
		ClientSigningPublicKey: append([]byte(nil), publicKey[:]...), PrincipalID: transaction.account.ID, AccountState: "active",
	}
}

func task10CreateSessionCommand() CreateSessionCommand {
	return CreateSessionCommand{
		Method: SessionPassword, Email: "member@example.test", Password: secret.NewBytes([]byte("correct horse battery staple")),
		ClientSigningPublicKey: task10PublicKey(), IdempotencyKey: "abcdefghijklmnopqrstuv",
	}
}

func task10PublicKey() [32]byte {
	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{0x73}, ed25519.SeedSize))
	var public [32]byte
	copy(public[:], private.Public().(ed25519.PublicKey))
	clear(private)
	return public
}

func newTask10Application(t *testing.T, transaction *task10Transaction, challenges ChallengeStore) (*Service, *task10Deriver, *task10Limiter) {
	t.Helper()
	protector, err := sensitive.NewLocal(secret.NewBytes(bytes.Repeat([]byte{0x11}, 32)), secret.NewBytes(bytes.Repeat([]byte{0x22}, 32)), 7)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = protector.Close() })
	deriver := &task10Deriver{}
	limiter := &task10Limiter{allowed: true}
	repository := &task10Repository{tx: transaction}
	if concrete, ok := challenges.(*task10ChallengeStore); ok {
		concrete.repository = repository
	}
	application, err := newApplicationForTest(ApplicationDependencies{
		Repository: repository, Protector: protector, Random: &task9Random{}, Clock: task10Clock{},
		Limiter: limiter, RateLimitKey: secret.NewBytes(bytes.Repeat([]byte{0x33}, 32)), ChallengeStore: challenges,
		Security: config.SecurityConfig{
			Profile: config.ProfileTest, EmailVerification: config.EmailDisabled, PublicBaseURL: "https://api.example.test",
			RequestDeadline: 5 * time.Second, RedisTimeout: 250 * time.Millisecond,
			LoginRateLimit:     config.RateLimitPolicy{Limit: 10, Window: 15 * time.Minute},
			DeliveryRateLimit:  config.RateLimitPolicy{Limit: 5, Window: time.Hour},
			ChallengeRateLimit: config.RateLimitPolicy{Limit: 20, Window: 5 * time.Minute},
		}, DeviceAuthorizationParticipant: &task9Participant{owner: transaction.fakeIdentityTransaction},
	}, deriver.Derive, uuid.New)
	if err != nil {
		t.Fatal(err)
	}
	return application, deriver, limiter
}

func activeTask10Transaction() *task10Transaction {
	principalID := uuid.MustParse("28ceee8a-5f4f-4d3a-9e9f-d3ec23b815ac")
	base := &fakeIdentityTransaction{
		identityFound: true, identity: store.IdentityEmailIdentity{PrincipalID: principalID},
		accountFound: true, account: store.IdentityAccount{ID: principalID, State: "active", StateVersion: 4},
		credentialFound: true, credential: store.IdentityPasswordCredential{
			PrincipalID: principalID, PolicyVersion: 1, MemoryKib: 65536, TimeCost: 3, Parallelism: 4,
			Salt: bytes.Repeat([]byte{0x31}, 16), PasswordHash: bytes.Repeat([]byte{0xa5}, 32), UpdatedAt: fixedTask10Time,
		},
	}
	return &task10Transaction{fakeIdentityTransaction: base}
}

type task10Clock struct{}

func (task10Clock) Now() time.Time { return fixedTask10Time }

type task10Deriver struct{ calls int }

func (deriver *task10Deriver) Derive(_ []byte, _ []byte, policy PasswordPolicy) []byte {
	deriver.calls++
	return bytes.Repeat([]byte{0xa5}, int(policy.TagBytes))
}

type task10Limiter struct {
	calls     int
	operation ratelimit.Operation
	allowed   bool
	err       error
}

func (limiter *task10Limiter) Allow(_ context.Context, operation ratelimit.Operation, _ [32]byte, _ config.RateLimitPolicy) (bool, error) {
	limiter.calls++
	limiter.operation = operation
	return limiter.allowed, limiter.err
}

type task10ChallengeStore struct {
	mu                      sync.Mutex
	records                 map[string]ChallengeRecord
	createCalls             int
	consumeCalls            int
	createdAtOperationCount int
	repository              *task10Repository
	createErr               error
	consumeErr              error
}

func (store *task10ChallengeStore) Create(_ context.Context, record ChallengeRecord, _ time.Duration) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.records == nil {
		store.records = make(map[string]ChallengeRecord)
	}
	store.createCalls++
	if store.createErr != nil {
		return store.createErr
	}
	if store.repository != nil {
		if store.repository.inTransaction.Load() {
			return errors.New("challenge store called inside database transaction")
		}
		store.createdAtOperationCount = len(store.repository.tx.operations)
	}
	store.records[record.ChallengeID] = record
	return nil
}

func (store *task10ChallengeStore) Consume(_ context.Context, challengeID string, digest [32]byte) (ChallengeRecord, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.consumeCalls++
	if store.consumeErr != nil {
		return ChallengeRecord{}, store.consumeErr
	}
	if store.repository != nil && store.repository.inTransaction.Load() {
		return ChallengeRecord{}, errors.New("challenge store called inside database transaction")
	}
	record, found := store.records[challengeID]
	if !found {
		return ChallengeRecord{}, ErrChallengeNotFound
	}
	delete(store.records, challengeID)
	if !challengeContextMatches(record.ContextDigest, digest) {
		return ChallengeRecord{}, ErrChallengeNotFound
	}
	return record, nil
}

type task10Repository struct {
	mu            sync.Mutex
	tx            *task10Transaction
	inTransaction atomic.Bool
}

func (repository *task10Repository) WithinTransaction(ctx context.Context, operation func(context.Context, Transaction) error) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.inTransaction.Store(true)
	err := operation(ctx, repository.tx)
	repository.inTransaction.Store(false)
	if err != nil {
		repository.tx.operations = append(repository.tx.operations, "rollback")
		return err
	}
	repository.tx.operations = append(repository.tx.operations, "commit")
	return nil
}

type task10Transaction struct {
	*fakeIdentityTransaction
	accessFound   bool
	access        store.FindAccountAccessTokenRow
	accessDigest  []byte
	refreshFound  bool
	refresh       store.GetRefreshTokenForUpdateRow
	refreshDigest []byte
	revokedParams store.RevokePrincipalAccountSessionsParams
}

func (transaction *task10Transaction) CreateAccountSession(ctx context.Context, params store.CreateAccountSessionParams) error {
	if err := transaction.fakeIdentityTransaction.CreateAccountSession(ctx, params); err != nil {
		return err
	}
	transaction.accessFound = true
	transaction.accessDigest = bytes.Clone(params.AccessTokenHash)
	transaction.access = store.FindAccountAccessTokenRow{
		ID: params.ID, PrincipalID: params.PrincipalID, State: "active", StateVersion: 1,
		AccessExpiresAt: params.AccessExpiresAt, AbsoluteExpiresAt: params.AbsoluteExpiresAt, AccountState: "active",
	}
	return nil
}

func (transaction *task10Transaction) FindAccountAccessToken(_ context.Context, digest []byte) (store.FindAccountAccessTokenRow, bool, error) {
	err := transaction.record("find_access")
	return transaction.access, transaction.accessFound && bytes.Equal(digest, transaction.accessDigest), err
}

func (transaction *task10Transaction) GetRefreshTokenForUpdate(_ context.Context, digest []byte) (store.GetRefreshTokenForUpdateRow, bool, error) {
	err := transaction.record("get_refresh")
	return transaction.refresh, transaction.refreshFound && bytes.Equal(digest, transaction.refreshDigest), err
}

func (transaction *task10Transaction) MarkAccountRefreshUsed(context.Context, store.MarkAccountRefreshUsedParams) (bool, error) {
	if err := transaction.record("mark_refresh_used"); err != nil {
		return false, err
	}
	if transaction.refresh.RefreshState != "active" {
		return false, nil
	}
	transaction.refresh.RefreshState = "used"
	transaction.refresh.UsedAt = sql.NullTime{Time: fixedTask10Time, Valid: true}
	return true, nil
}

func (transaction *task10Transaction) RotateAccountSessionAccess(context.Context, store.RotateAccountSessionAccessParams) (store.IdentityAccountSession, bool, error) {
	return store.IdentityAccountSession{}, true, transaction.record("rotate_session_access")
}

func (transaction *task10Transaction) RevokeAccountRefreshTokens(context.Context, store.RevokeAccountRefreshTokensParams) (int64, error) {
	return 1, transaction.record("revoke_refresh")
}

func (transaction *task10Transaction) MarkAccountSessionCompromised(context.Context, store.MarkAccountSessionCompromisedParams) (int64, error) {
	return 1, transaction.record("mark_session_compromised")
}

func (transaction *task10Transaction) RevokePrincipalAccountSessions(_ context.Context, params store.RevokePrincipalAccountSessionsParams) ([]uuid.UUID, error) {
	transaction.revokedParams = params
	return []uuid.UUID{transaction.account.ID}, transaction.record("revoke_sessions")
}

func (transaction *task10Transaction) UpdatePasswordCredential(context.Context, store.UpdatePasswordCredentialParams) (int64, error) {
	return 1, transaction.record("update_password")
}

func publicTask10Code(err error) apierrors.Code {
	var classified apierrors.Error
	if !errors.As(err, &classified) {
		return ""
	}
	return apierrors.Code(classified.Public("trace-safe-000010").Code)
}

func containsTask10Order(got, want []string) bool {
	position := 0
	for _, operation := range got {
		if position < len(want) && operation == want[position] {
			position++
		}
	}
	return position == len(want)
}

var _ securitykit.Clock = task10Clock{}
var _ Repository = (*task10Repository)(nil)
var _ sessionTransaction = (*postgresTransaction)(nil)
var _ = sql.NullTime{}
var _ = idempotency.Started
