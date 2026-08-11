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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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
	if got := transaction.operations; !containsTask10Order(got, []string{"find_identity", "get_credential", "lock_principal_refresh", "lock_sessions", "get_account", "begin_idempotency", "create_account_session", "insert_account_refresh", "complete_idempotency", "commit"}) {
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
	if !containsTask10Order(transaction.operations, []string{"discover_refresh", "lock_session_refresh", "lock_sessions", "get_account", "begin_idempotency", "rollback", "discover_refresh", "lock_session_refresh", "lock_sessions", "get_account", "begin_idempotency", "complete_idempotency", "commit"}) || challenges.createdAtOperationCount < 3 {
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
	if !containsTask10Order(transaction.operations, []string{"discover_refresh", "lock_session_refresh", "lock_sessions", "get_account", "commit", "discover_refresh", "lock_session_refresh", "lock_sessions", "get_account", "begin_idempotency", "rollback", "discover_refresh", "lock_session_refresh", "lock_sessions", "get_account", "begin_idempotency", "mark_refresh_used", "rotate_session_access", "insert_account_refresh", "complete_idempotency", "commit"}) {
		t.Fatalf("rotation operation order = %v", transaction.operations)
	}
}

func TestUsedRefreshSequenceReplaysExactRequestAndCompromisesDifferentKeyWithoutRedis(t *testing.T) {
	transaction := activeTask10Transaction()
	refresh := secret.NewBytes(bytes.Repeat([]byte{0x51}, 32))
	publicKey := task10PublicKey()
	configureTask10Refresh(transaction, refresh, publicKey, "active")
	executor := &task10StatefulRedisExecutor{values: make(map[string][]byte)}
	challenges, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	application, _, _ := newTask10Application(t, transaction, challenges)
	executor.repository = application.repository.(*task10Repository)
	command := task10RotateCommand(t, application, refresh, transaction.refresh.SessionID, "abcdefghijklmnopqrstuv")

	first, err := application.RotateSession(context.Background(), command)
	if err != nil {
		t.Fatal(err)
	}
	if executor.runCalls != 1 || transaction.refresh.RefreshState != "used" || transaction.activeRefreshChildren != 1 {
		t.Fatalf("first rotation sequence = GETDEL:%d state:%q children:%d", executor.runCalls, transaction.refresh.RefreshState, transaction.activeRefreshChildren)
	}
	firstAccess := securitykit.DigestToken(securitykit.AccountAccessToken, first.AccessToken)
	firstRefresh := securitykit.DigestToken(securitykit.AccountRefreshToken, first.RefreshToken)

	replayed, err := application.RotateSession(context.Background(), command)
	if err != nil {
		t.Fatalf("exact committed replay: %v", err)
	}
	if securitykit.DigestToken(securitykit.AccountAccessToken, replayed.AccessToken) != firstAccess ||
		securitykit.DigestToken(securitykit.AccountRefreshToken, replayed.RefreshToken) != firstRefresh || executor.runCalls != 1 || transaction.sessionCompromised {
		t.Fatal("exact replay changed tokens, re-consumed Redis, or compromised the session")
	}

	differentKey := command
	differentKey.IdempotencyKey = "bcdefghijklmnopqrstuvw"
	_, err = application.RotateSession(context.Background(), differentKey)
	if publicTask10Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("different-key used refresh error = %v", err)
	}
	if executor.runCalls != 1 || !transaction.sessionCompromised || transaction.activeRefreshChildren != 0 ||
		len(transaction.revokedTokenHashes) != 2 || len(transaction.securityEvents) != 1 || transaction.completed401 != 1 {
		t.Fatalf("used replay did not commit compromise without Redis: GETDEL:%d compromised:%v children:%d events:%d tombstones:%d operations:%v",
			executor.runCalls, transaction.sessionCompromised, transaction.activeRefreshChildren, len(transaction.securityEvents), transaction.completed401, transaction.operations)
	}
}

func TestRotationRedisMissReclassifiesUsedAuthorityAndCompromisesFamily(t *testing.T) {
	transaction := activeTask10Transaction()
	refresh := secret.NewBytes(bytes.Repeat([]byte{0x51}, 32))
	publicKey := task10PublicKey()
	configureTask10Refresh(transaction, refresh, publicKey, "active")
	executor := &task10StatefulRedisExecutor{values: make(map[string][]byte)}
	challenges, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	application, _, _ := newTask10Application(t, transaction, challenges)
	executor.repository = application.repository.(*task10Repository)
	command := task10RotateCommand(t, application, refresh, transaction.refresh.SessionID, "abcdefghijklmnopqrstuv")
	executor.discard(accountChallengeRedisPrefix + command.ChallengeID)
	executor.beforeRun = func() {
		transaction.refresh.RefreshState = "used"
		transaction.activeRefreshChildren = 1
	}

	_, err = application.RotateSession(context.Background(), command)
	if publicTask10Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("used authority after Redis miss error = %v", err)
	}
	if !transaction.sessionCompromised || transaction.activeRefreshChildren != 0 || len(transaction.securityEvents) != 1 || transaction.completed401 != 1 {
		t.Fatalf("Redis miss bypassed authoritative compromise: compromised:%v children:%d events:%d tombstones:%d operations:%v",
			transaction.sessionCompromised, transaction.activeRefreshChildren, len(transaction.securityEvents), transaction.completed401, transaction.operations)
	}
}

func TestRotationRedisMissRequiresSecondAuthoritativeClassification(t *testing.T) {
	transaction := activeTask10Transaction()
	refresh := secret.NewBytes(bytes.Repeat([]byte{0x51}, 32))
	publicKey := task10PublicKey()
	configureTask10Refresh(transaction, refresh, publicKey, "active")
	executor := &task10StatefulRedisExecutor{values: make(map[string][]byte)}
	challenges, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	application, _, _ := newTask10Application(t, transaction, challenges)
	executor.repository = application.repository.(*task10Repository)
	command := task10RotateCommand(t, application, refresh, transaction.refresh.SessionID, "abcdefghijklmnopqrstuv")
	executor.discard(accountChallengeRedisPrefix + command.ChallengeID)
	transaction.getRefreshErrorAt = transaction.getRefreshCalls + 2

	_, err = application.RotateSession(context.Background(), command)
	if publicTask10Code(err) != apierrors.DependencyUnavailable {
		t.Fatalf("second authority lookup failure = %v", err)
	}
}

func TestRotationRedisAmbiguityReclassifiesProvablyUsedAuthority(t *testing.T) {
	transaction := activeTask10Transaction()
	refresh := secret.NewBytes(bytes.Repeat([]byte{0x51}, 32))
	publicKey := task10PublicKey()
	configureTask10Refresh(transaction, refresh, publicKey, "active")
	executor := &task10StatefulRedisExecutor{values: make(map[string][]byte), runError: errors.New("REDIS-BACKEND-CANARY")}
	challenges, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	application, _, _ := newTask10Application(t, transaction, challenges)
	executor.repository = application.repository.(*task10Repository)
	command := task10RotateCommand(t, application, refresh, transaction.refresh.SessionID, "abcdefghijklmnopqrstuv")
	executor.beforeRun = func() {
		transaction.refresh.RefreshState = "used"
		transaction.activeRefreshChildren = 1
	}

	_, err = application.RotateSession(context.Background(), command)
	if publicTask10Code(err) != apierrors.AuthenticationFailed || !transaction.sessionCompromised || transaction.completed401 != 1 {
		t.Fatalf("Redis ambiguity failed to honor proved used authority: error:%v compromised:%v tombstones:%d", err, transaction.sessionCompromised, transaction.completed401)
	}
}

func TestConcurrentIdenticalRotationBeforeWinnerCommitReturnsFiniteAuthenticationFailure(t *testing.T) {
	transaction := activeTask10Transaction()
	refresh := secret.NewBytes(bytes.Repeat([]byte{0x51}, 32))
	publicKey := task10PublicKey()
	configureTask10Refresh(transaction, refresh, publicKey, "active")
	executor := &task10StatefulRedisExecutor{values: make(map[string][]byte)}
	challenges, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	application, _, _ := newTask10Application(t, transaction, challenges)
	executor.repository = application.repository.(*task10Repository)
	command := task10RotateCommand(t, application, refresh, transaction.refresh.SessionID, "abcdefghijklmnopqrstuv")
	hitSeen := make(chan struct{}, 1)
	hitContext, releaseHit := context.WithCancel(context.Background())
	t.Cleanup(releaseHit)
	missSeen := make(chan struct{}, 1)
	executor.hitSeen, executor.hitRelease, executor.missSeen = hitSeen, hitContext.Done(), missSeen

	winnerDone := make(chan task10RotationResult, 1)
	go func() {
		tokens, rotateErr := application.RotateSession(context.Background(), command)
		winnerDone <- task10RotationResult{tokens: tokens, err: rotateErr}
	}()
	waitTask10Signal(t, hitSeen, "winner Redis consume")
	loserDone := make(chan task10RotationResult, 1)
	go func() {
		tokens, rotateErr := application.RotateSession(context.Background(), command)
		loserDone <- task10RotationResult{tokens: tokens, err: rotateErr}
	}()
	waitTask10Signal(t, missSeen, "loser Redis miss")
	loser := waitTask10Rotation(t, loserDone, "loser pre-commit result")
	releaseHit()
	winner := waitTask10Rotation(t, winnerDone, "winner result")
	if winner.err != nil || publicTask10Code(loser.err) != apierrors.AuthenticationFailed || transaction.sessionCompromised {
		t.Fatalf("pre-commit identical race = winner:%v loser:%v compromised:%v", winner.err, loser.err, transaction.sessionCompromised)
	}
}

func TestConcurrentIdenticalRotationAfterWinnerCommitReplaysOriginalTokens(t *testing.T) {
	transaction := activeTask10Transaction()
	transaction.rotationCompleted = make(chan struct{})
	t.Cleanup(func() {
		transaction.rotationCompletedOnce.Do(func() { close(transaction.rotationCompleted) })
	})
	refresh := secret.NewBytes(bytes.Repeat([]byte{0x51}, 32))
	publicKey := task10PublicKey()
	configureTask10Refresh(transaction, refresh, publicKey, "active")
	executor := &task10StatefulRedisExecutor{values: make(map[string][]byte)}
	challenges, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	application, _, _ := newTask10Application(t, transaction, challenges)
	executor.repository = application.repository.(*task10Repository)
	command := task10RotateCommand(t, application, refresh, transaction.refresh.SessionID, "abcdefghijklmnopqrstuv")
	hitSeen := make(chan struct{}, 1)
	hitContext, releaseHit := context.WithCancel(context.Background())
	t.Cleanup(releaseHit)
	missSeen := make(chan struct{}, 1)
	executor.hitSeen, executor.hitRelease = hitSeen, hitContext.Done()
	executor.missSeen, executor.missRelease = missSeen, transaction.rotationCompleted

	winnerDone := make(chan task10RotationResult, 1)
	go func() {
		tokens, rotateErr := application.RotateSession(context.Background(), command)
		winnerDone <- task10RotationResult{tokens: tokens, err: rotateErr}
	}()
	waitTask10Signal(t, hitSeen, "winner Redis consume")
	loserDone := make(chan task10RotationResult, 1)
	go func() {
		tokens, rotateErr := application.RotateSession(context.Background(), command)
		loserDone <- task10RotationResult{tokens: tokens, err: rotateErr}
	}()
	waitTask10Signal(t, missSeen, "loser Redis miss")
	releaseHit()
	winner := waitTask10Rotation(t, winnerDone, "winner result")
	loser := waitTask10Rotation(t, loserDone, "loser replay result")
	if winner.err != nil || loser.err != nil || transaction.sessionCompromised ||
		securitykit.DigestToken(securitykit.AccountAccessToken, winner.tokens.AccessToken) != securitykit.DigestToken(securitykit.AccountAccessToken, loser.tokens.AccessToken) ||
		securitykit.DigestToken(securitykit.AccountRefreshToken, winner.tokens.RefreshToken) != securitykit.DigestToken(securitykit.AccountRefreshToken, loser.tokens.RefreshToken) {
		t.Fatalf("post-commit identical race = winner:%v loser:%v compromised:%v", winner.err, loser.err, transaction.sessionCompromised)
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
	transaction.sessionFound = true
	transaction.session = store.IdentityAccountSession{ID: currentSession, PrincipalID: transaction.account.ID, State: "active", AbsoluteExpiresAt: fixedTask10Time.Add(time.Hour)}
	application, _, _ := newTask10Application(t, transaction, &task10ChallengeStore{})
	err := application.RevokeSessions(context.Background(), RevokeSessionsCommand{
		PrincipalID: PrincipalID(transaction.account.ID.String()), Scope: RevokeOtherSessions,
		SessionID: SessionID(currentSession.String()), IdempotencyKey: "abcdefghijklmnopqrstuv",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(transaction.revokedParams.SessionIds) != 0 || len(transaction.revokedParams.TokenHashes) != 0 {
		t.Fatalf("generated revoke params = %#v", transaction.revokedParams)
	}
	if !containsTask10Order(transaction.operations, []string{"begin_idempotency", "lock_principal_refresh", "lock_sessions", "get_account", "revoke_sessions", "complete_idempotency", "commit"}) {
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
		"get_credential", "lock_principal_refresh", "lock_sessions", "get_account", "begin_idempotency", "update_password", "mark_sessions_review_required",
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

type task10RotationResult struct {
	tokens SessionTokens
	err    error
}

func waitTask10Signal(t *testing.T, signal <-chan struct{}, label string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s", label)
	}
}

func waitTask10Rotation(t *testing.T, result <-chan task10RotationResult, label string) task10RotationResult {
	t.Helper()
	select {
	case received := <-result:
		return received
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s", label)
		return task10RotationResult{}
	}
}

func configureTask10Refresh(transaction *task10Transaction, token secret.Bytes, publicKey [32]byte, state string) {
	digest := securitykit.DigestToken(securitykit.AccountRefreshToken, token)
	transaction.refreshDigest = append([]byte(nil), digest[:]...)
	transaction.refreshFound = true
	transaction.refresh = accountRefreshAuthority{
		TokenHash: append([]byte(nil), digest[:]...), SessionID: uuid.MustParse("d12dca8a-ced3-471d-a6ad-55b228221f10"),
		RefreshState: state, IssuedAt: fixedTask10Time.Add(-time.Hour), IdleExpiresAt: fixedTask10Time.Add(time.Hour),
		AbsoluteExpiresAt: fixedTask10Time.Add(24 * time.Hour), SessionState: "active", SessionStateVersion: 3,
		ClientSigningPublicKey: append([]byte(nil), publicKey[:]...), PrincipalID: transaction.account.ID, AccountState: "active",
		LockedFamilyTokenHashes: [][]byte{append([]byte(nil), digest[:]...)},
	}
	transaction.sessionFound = true
	transaction.session = store.IdentityAccountSession{
		ID: transaction.refresh.SessionID, PrincipalID: transaction.refresh.PrincipalID, State: transaction.refresh.SessionState,
		StateVersion: transaction.refresh.SessionStateVersion, ClientSigningPublicKey: append([]byte(nil), publicKey[:]...),
		AbsoluteExpiresAt: transaction.refresh.AbsoluteExpiresAt,
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
	transaction.idempotencyDB = newTask10IdempotencyDB()
	transaction.idempotency, err = idempotency.New(transaction.idempotencyDB, protector)
	if err != nil {
		t.Fatal(err)
	}
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

type task10StatefulRedisExecutor struct {
	mu          sync.Mutex
	values      map[string][]byte
	runCalls    int
	repository  *task10Repository
	beforeRun   func()
	runError    error
	hitSeen     chan<- struct{}
	hitRelease  <-chan struct{}
	missSeen    chan<- struct{}
	missRelease <-chan struct{}
}

func (executor *task10StatefulRedisExecutor) SetNX(_ context.Context, key string, value []byte, _ time.Duration) (bool, error) {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	if executor.repository != nil && executor.repository.inTransaction.Load() {
		return false, errors.New("Redis SETNX called inside database transaction")
	}
	if _, exists := executor.values[key]; exists {
		return false, nil
	}
	executor.values[key] = bytes.Clone(value)
	return true, nil
}

func (executor *task10StatefulRedisExecutor) Run(ctx context.Context, script string, keys []string, _ ...any) (any, error) {
	executor.mu.Lock()
	if executor.repository != nil && executor.repository.inTransaction.Load() {
		executor.mu.Unlock()
		return nil, errors.New("Redis GETDEL called inside database transaction")
	}
	if script != accountChallengeGetDEL || len(keys) != 1 {
		executor.mu.Unlock()
		return nil, errors.New("unexpected Redis challenge operation")
	}
	executor.runCalls++
	if executor.beforeRun != nil {
		executor.beforeRun()
		executor.beforeRun = nil
	}
	if executor.runError != nil {
		err := executor.runError
		executor.mu.Unlock()
		return nil, err
	}
	value, exists := executor.values[keys[0]]
	if !exists {
		missSeen, missRelease := executor.missSeen, executor.missRelease
		executor.mu.Unlock()
		if missSeen != nil {
			select {
			case missSeen <- struct{}{}:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		if missRelease != nil {
			select {
			case <-missRelease:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return nil, nil
	}
	delete(executor.values, keys[0])
	hitSeen, hitRelease := executor.hitSeen, executor.hitRelease
	executor.mu.Unlock()
	if hitSeen != nil {
		select {
		case hitSeen <- struct{}{}:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if hitRelease != nil {
		select {
		case <-hitRelease:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return bytes.Clone(value), nil
}

func (executor *task10StatefulRedisExecutor) discard(key string) {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	delete(executor.values, key)
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
	var idempotencySnapshot map[string]store.IdempotencyRecord
	if repository.tx.idempotencyDB != nil {
		idempotencySnapshot = repository.tx.idempotencyDB.snapshot()
	}
	repository.inTransaction.Store(true)
	err := operation(ctx, repository.tx)
	repository.inTransaction.Store(false)
	if err != nil {
		if repository.tx.idempotencyDB != nil {
			repository.tx.idempotencyDB.restore(idempotencySnapshot)
		}
		repository.tx.operations = append(repository.tx.operations, "rollback")
		return err
	}
	repository.tx.operations = append(repository.tx.operations, "commit")
	return nil
}

type task10Transaction struct {
	*fakeIdentityTransaction
	accessFound           bool
	access                store.FindAccountAccessTokenRow
	accessDigest          []byte
	refreshFound          bool
	refresh               accountRefreshAuthority
	refreshDigest         []byte
	revokedParams         store.RevokePrincipalAccountSessionsParams
	idempotency           idempotency.Repository
	idempotencyDB         *task10IdempotencyDB
	activeRefreshChildren int
	revokedTokenHashes    [][]byte
	sessionCompromised    bool
	completed401          int
	getRefreshCalls       int
	getRefreshErrorAt     int
	rotationCompleted     chan struct{}
	rotationCompletedOnce sync.Once
}

func (transaction *task10Transaction) BeginIdempotency(ctx context.Context, scope idempotency.Scope, key string, canonical []byte, createdAt, expiresAt time.Time) (idempotency.Record, idempotency.Outcome, error) {
	transaction.idempotencyCanonical = append(transaction.idempotencyCanonical, bytes.Clone(canonical))
	if err := transaction.record("begin_idempotency"); err != nil {
		return idempotency.Record{}, "", err
	}
	return transaction.idempotency.Begin(ctx, scope, key, canonical, createdAt, expiresAt)
}

func (transaction *task10Transaction) CompleteIdempotency(ctx context.Context, record idempotency.Record, status int, body []byte) error {
	if err := transaction.record("complete_idempotency"); err != nil {
		return err
	}
	completed, err := transaction.idempotency.Complete(ctx, record, status, body)
	if err != nil {
		return err
	}
	owned, available := completed.TakeResponseBody()
	defer clear(owned)
	if !available {
		return errors.New("idempotency completion did not transfer response ownership")
	}
	if status == 401 {
		transaction.completed401++
	}
	if status == 200 && transaction.rotationCompleted != nil {
		transaction.rotationCompletedOnce.Do(func() { close(transaction.rotationCompleted) })
	}
	return nil
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

func (transaction *task10Transaction) InsertAccountRefreshToken(ctx context.Context, params store.InsertAccountRefreshTokenParams) error {
	if err := transaction.fakeIdentityTransaction.InsertAccountRefreshToken(ctx, params); err != nil {
		return err
	}
	if len(params.PreviousTokenHash) == 32 {
		transaction.activeRefreshChildren++
	}
	return nil
}

func (transaction *task10Transaction) FindAccountAccessToken(_ context.Context, digest []byte) (store.FindAccountAccessTokenRow, bool, error) {
	err := transaction.record("find_access")
	return transaction.access, transaction.accessFound && bytes.Equal(digest, transaction.accessDigest), err
}

func (transaction *task10Transaction) DiscoverRefreshToken(_ context.Context, digest []byte) (store.DiscoverRefreshTokenRow, bool, error) {
	transaction.getRefreshCalls++
	err := transaction.record("discover_refresh")
	if transaction.getRefreshErrorAt == transaction.getRefreshCalls {
		return store.DiscoverRefreshTokenRow{}, false, errors.New("authority lookup unavailable")
	}
	return store.DiscoverRefreshTokenRow{
		TokenHash: bytes.Clone(transaction.refresh.TokenHash), SessionID: transaction.refresh.SessionID, PrincipalID: transaction.refresh.PrincipalID,
	}, transaction.refreshFound && bytes.Equal(digest, transaction.refreshDigest), err
}

func (transaction *task10Transaction) LockSessionRefreshTokens(context.Context, uuid.UUID) ([]store.IdentityAccountRefreshToken, error) {
	if err := transaction.record("lock_session_refresh"); err != nil {
		return nil, err
	}
	if !transaction.refreshFound {
		return nil, nil
	}
	return transaction.task10RefreshRows(), nil
}

func (transaction *task10Transaction) ListSessionRefreshTokens(context.Context, uuid.UUID) ([]store.IdentityAccountRefreshToken, error) {
	if err := transaction.record("list_session_refresh"); err != nil {
		return nil, err
	}
	if !transaction.refreshFound {
		return nil, nil
	}
	return transaction.task10RefreshRows(), nil
}

func (transaction *task10Transaction) task10RefreshRows() []store.IdentityAccountRefreshToken {
	rows := []store.IdentityAccountRefreshToken{task10RefreshModel(transaction.refresh)}
	for index := 0; index < transaction.activeRefreshChildren; index++ {
		tokenHash := bytes.Repeat([]byte{byte(0xc0 + index)}, 32)
		rows = append(rows, store.IdentityAccountRefreshToken{
			TokenHash: tokenHash, SessionID: transaction.refresh.SessionID, PreviousTokenHash: bytes.Clone(transaction.refresh.TokenHash),
			State: "active", IssuedAt: fixedTask10Time, IdleExpiresAt: fixedTask10Time.Add(time.Hour), AbsoluteExpiresAt: transaction.refresh.AbsoluteExpiresAt,
		})
	}
	return rows
}

func task10RefreshModel(row accountRefreshAuthority) store.IdentityAccountRefreshToken {
	return store.IdentityAccountRefreshToken{
		TokenHash: bytes.Clone(row.TokenHash), SessionID: row.SessionID, PreviousTokenHash: bytes.Clone(row.PreviousTokenHash), State: row.RefreshState,
		IssuedAt: row.IssuedAt, IdleExpiresAt: row.IdleExpiresAt, AbsoluteExpiresAt: row.AbsoluteExpiresAt, UsedAt: row.UsedAt, RevokedAt: row.RevokedAt,
	}
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

func (transaction *task10Transaction) RevokeAccountRefreshTokens(_ context.Context, params store.RevokeAccountRefreshTokensParams) (int64, error) {
	if err := transaction.record("revoke_refresh"); err != nil {
		return 0, err
	}
	transaction.revokedTokenHashes = cloneTokenHashes(params.TokenHashes)
	revoked := transaction.activeRefreshChildren
	transaction.activeRefreshChildren = 0
	return int64(revoked), nil
}

func (transaction *task10Transaction) MarkAccountSessionCompromised(context.Context, store.MarkAccountSessionCompromisedParams) (int64, error) {
	if err := transaction.record("mark_session_compromised"); err != nil {
		return 0, err
	}
	transaction.sessionCompromised = true
	transaction.refresh.SessionState = "compromised"
	return 1, nil
}

func (transaction *task10Transaction) RevokePrincipalAccountSessions(_ context.Context, params store.RevokePrincipalAccountSessionsParams) ([]uuid.UUID, error) {
	transaction.revokedParams = params
	return []uuid.UUID{transaction.account.ID}, transaction.record("revoke_sessions")
}

func (transaction *task10Transaction) UpdatePasswordCredential(context.Context, store.UpdatePasswordCredentialParams) (int64, error) {
	return 1, transaction.record("update_password")
}

type task10IdempotencyDB struct {
	mu      sync.Mutex
	records map[string]store.IdempotencyRecord
}

func newTask10IdempotencyDB() *task10IdempotencyDB {
	return &task10IdempotencyDB{records: make(map[string]store.IdempotencyRecord)}
}

func (database *task10IdempotencyDB) Exec(_ context.Context, query string, arguments ...interface{}) (pgconn.CommandTag, error) {
	if !strings.Contains(query, "INSERT INTO idempotency_records") || len(arguments) != 6 {
		return pgconn.CommandTag{}, errors.New("unexpected idempotency exec")
	}
	database.mu.Lock()
	defer database.mu.Unlock()
	key := task10IdempotencyKey(arguments[0].(string), arguments[1].(string), arguments[2].([]byte))
	if _, exists := database.records[key]; exists {
		return pgconn.NewCommandTag("INSERT 0 0"), nil
	}
	database.records[key] = store.IdempotencyRecord{
		PrincipalScope: arguments[0].(string), Operation: arguments[1].(string),
		IdempotencyKeyHash: bytes.Clone(arguments[2].([]byte)), RequestDigest: bytes.Clone(arguments[3].([]byte)),
		State: "in_progress", CreatedAt: arguments[4].(time.Time), ExpiresAt: arguments[5].(time.Time),
	}
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

func (*task10IdempotencyDB) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, errors.New("unexpected idempotency query")
}

func (database *task10IdempotencyDB) QueryRow(_ context.Context, query string, arguments ...interface{}) pgx.Row {
	database.mu.Lock()
	defer database.mu.Unlock()
	if strings.Contains(query, "UPDATE idempotency_records") && len(arguments) == 7 {
		key := task10IdempotencyKey(arguments[0].(string), arguments[1].(string), arguments[2].([]byte))
		record, exists := database.records[key]
		if !exists || record.State != "in_progress" || !bytes.Equal(record.RequestDigest, arguments[3].([]byte)) {
			return task10IdempotencyRow{err: pgx.ErrNoRows}
		}
		record.State = "completed"
		record.ResponseStatus = arguments[4].(pgtype.Int4)
		record.ResponseCiphertext = bytes.Clone(arguments[5].([]byte))
		record.ResponseKeyVersion = arguments[6].(pgtype.Int4)
		database.records[key] = task10CloneIdempotencyRecord(record)
		return task10IdempotencyRow{record: record}
	}
	if strings.Contains(query, "FROM idempotency_records") && len(arguments) == 3 {
		key := task10IdempotencyKey(arguments[0].(string), arguments[1].(string), arguments[2].([]byte))
		record, exists := database.records[key]
		if !exists {
			return task10IdempotencyRow{err: pgx.ErrNoRows}
		}
		return task10IdempotencyRow{record: task10CloneIdempotencyRecord(record)}
	}
	return task10IdempotencyRow{err: errors.New("unexpected idempotency query row")}
}

func (database *task10IdempotencyDB) snapshot() map[string]store.IdempotencyRecord {
	database.mu.Lock()
	defer database.mu.Unlock()
	result := make(map[string]store.IdempotencyRecord, len(database.records))
	for key, record := range database.records {
		result[key] = task10CloneIdempotencyRecord(record)
	}
	return result
}

func (database *task10IdempotencyDB) restore(snapshot map[string]store.IdempotencyRecord) {
	database.mu.Lock()
	defer database.mu.Unlock()
	database.records = make(map[string]store.IdempotencyRecord, len(snapshot))
	for key, record := range snapshot {
		database.records[key] = task10CloneIdempotencyRecord(record)
	}
}

func task10IdempotencyKey(principal, operation string, digest []byte) string {
	return principal + "\x00" + operation + "\x00" + string(digest)
}

func task10CloneIdempotencyRecord(record store.IdempotencyRecord) store.IdempotencyRecord {
	record.IdempotencyKeyHash = bytes.Clone(record.IdempotencyKeyHash)
	record.RequestDigest = bytes.Clone(record.RequestDigest)
	record.ResponseCiphertext = bytes.Clone(record.ResponseCiphertext)
	return record
}

type task10IdempotencyRow struct {
	record store.IdempotencyRecord
	err    error
}

func (row task10IdempotencyRow) Scan(destinations ...interface{}) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != 10 {
		return errors.New("unexpected idempotency scan")
	}
	*destinations[0].(*string) = row.record.PrincipalScope
	*destinations[1].(*string) = row.record.Operation
	*destinations[2].(*[]byte) = bytes.Clone(row.record.IdempotencyKeyHash)
	*destinations[3].(*[]byte) = bytes.Clone(row.record.RequestDigest)
	*destinations[4].(*string) = row.record.State
	*destinations[5].(*pgtype.Int4) = row.record.ResponseStatus
	*destinations[6].(*[]byte) = bytes.Clone(row.record.ResponseCiphertext)
	*destinations[7].(*pgtype.Int4) = row.record.ResponseKeyVersion
	*destinations[8].(*time.Time) = row.record.CreatedAt
	*destinations[9].(*time.Time) = row.record.ExpiresAt
	return nil
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
