package identity

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/store"
)

func TestLocalEmailSenderOwnsPrivateDeliveryAndDeduplicatesProviderID(t *testing.T) {
	t.Parallel()

	sender := NewLocalEmailSender()
	delivery := Delivery{
		DeliveryID: uuid.New(), TemplateID: VerifyEmailTemplate, Locale: "en", Recipient: "member@example.test",
		OneTimeToken: secret.NewBytes(bytes.Repeat([]byte{0x44}, 32)),
	}
	if err := sender.Send(context.Background(), delivery); err != nil {
		t.Fatal(err)
	}
	delivery.Recipient = "mutated@example.test"
	copyToken := delivery.OneTimeToken.Copy()
	copyToken[0] ^= 0xff
	duplicate := delivery
	duplicate.TemplateID = TemplateID("provider-mutated-after-acceptance")
	duplicate.OneTimeToken = secret.Bytes{}
	if err := sender.Send(context.Background(), duplicate); err != nil {
		t.Fatal("already accepted delivery must be idempotent")
	}
	captured := sender.Captured()
	if len(captured) != 1 || captured[0].Recipient != "member@example.test" || captured[0].OneTimeToken.Copy()[0] != 0x44 {
		t.Fatalf("captured delivery was not isolated: %v", captured)
	}
	captured[0].Recipient = "caller-mutated@example.test"
	if sender.Captured()[0].Recipient != "member@example.test" {
		t.Fatal("Captured returned aliased delivery")
	}

	rendered := fmt.Sprintf("%+v", captured[0]) + slog.Any("delivery", captured[0]).Value.String()
	if strings.Contains(rendered, "member@example") || strings.Contains(rendered, string(bytes.Repeat([]byte{0x44}, 8))) {
		t.Fatal("delivery formatting exposed private material")
	}
	if _, err := json.Marshal(captured[0]); err == nil {
		t.Fatal("Delivery unexpectedly marshaled")
	}
}

func TestVerifyEmailConsumesOnceActivatesParticipantInSameTransaction(t *testing.T) {
	t.Parallel()

	principalID := uuid.MustParse("4b4d278b-9e7a-4ce0-865d-1dc14fcf96da")
	deliveryID := uuid.MustParse("f4232063-70a4-4ef5-8627-5cd9f2a8ab9d")
	token := secret.NewBytes(bytes.Repeat([]byte{0x71}, 32))
	verificationDigest := securitykit.DigestToken(securitykit.EmailVerificationToken, token)
	tx := &fakeIdentityTransaction{
		verificationFound: true, consumeEmailOK: true,
		verification: store.IdentityEmailIdentity{
			PrincipalID: principalID, VerificationTokenHash: verificationDigest[:],
			VerificationDeliveryID: uuid.NullUUID{UUID: deliveryID, Valid: true},
			VerificationExpiresAt:  sql.NullTime{Time: fixedTask9Time.Add(emailVerificationTTL), Valid: true},
		},
		accountFound: true, account: store.IdentityAccount{ID: principalID, State: "pending_email", StateVersion: 1, Locale: "en"},
	}
	application, _, _, participant := newTask9Application(t, config.EmailRequired, tx)
	if err := application.VerifyEmail(context.Background(), VerifyEmailCommand{Token: token, IdempotencyKey: "abcdefghijklmnopqrstuv"}); err != nil {
		t.Fatal(err)
	}
	assertTask9Order(t, tx.operations, []string{
		"get_verification", "begin_idempotency", "consume_email", "clear_delivery", "activate_account", "participant",
		"append_event", "complete_idempotency", "commit",
	})
	if participant.calls != 1 || participant.tx != tx.DBTX() {
		t.Fatal("device authorization activation did not use the identity transaction")
	}
	if len(tx.events) != 1 || bytes.Contains(tx.events[0].GetPayload(), token.Copy()) {
		t.Fatal("account state event missing or exposed verification token")
	}
	assertTask9PrivateBinding(t, tx.idempotencyCanonical[0], "verify_email", verificationDigest[:])

	for _, test := range []struct {
		name      string
		consumed  bool
		expiresAt time.Time
	}{
		{name: "consumed", consumed: true, expiresAt: fixedTask9Time.Add(time.Hour)},
		{name: "expired", expiresAt: fixedTask9Time.Add(-time.Nanosecond)},
	} {
		t.Run(test.name, func(t *testing.T) {
			locked := store.IdentityEmailIdentity{
				PrincipalID: principalID, VerificationTokenHash: verificationDigest[:],
				VerificationExpiresAt: sql.NullTime{Time: test.expiresAt, Valid: true},
			}
			if test.consumed {
				locked.VerificationConsumedAt = sql.NullTime{Time: fixedTask9Time, Valid: true}
			}
			usedTx := &fakeIdentityTransaction{verificationFound: true, verification: locked}
			usedApplication, _, _, usedParticipant := newTask9Application(t, config.EmailRequired, usedTx)
			if verifyErr := usedApplication.VerifyEmail(context.Background(), VerifyEmailCommand{Token: token, IdempotencyKey: "abcdefghijklmnopqrstuv"}); publicTask9Code(verifyErr) != apierrors.AuthenticationFailed {
				t.Fatalf("verification state error = %v", verifyErr)
			}
			assertTask9Order(t, usedTx.operations, []string{"get_verification", "begin_idempotency", "rollback"})
			if usedParticipant.calls != 0 {
				t.Fatal("invalid verification state reached participant")
			}
		})
	}

	tx = &fakeIdentityTransaction{}
	application, _, _, participant = newTask9Application(t, config.EmailRequired, tx)
	if err := application.VerifyEmail(context.Background(), VerifyEmailCommand{Token: token, IdempotencyKey: "abcdefghijklmnopqrstuv"}); publicTask9Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("used/unknown verification token error = %v", err)
	}
	if participant.calls != 0 {
		t.Fatal("unknown verification token reached participant")
	}
}

func TestVerifyEmailAcceptsConsumerClearedDeliveryMetadata(t *testing.T) {
	t.Parallel()

	principalID := uuid.MustParse("75d3f431-8ab6-4c3a-9ce2-dd72c3555035")
	token := secret.NewBytes(bytes.Repeat([]byte{0x72}, 32))
	verificationDigest := securitykit.DigestToken(securitykit.EmailVerificationToken, token)
	tx := &fakeIdentityTransaction{
		failOperation:     "clear_delivery",
		verificationFound: true,
		consumeEmailOK:    true,
		verification: store.IdentityEmailIdentity{
			PrincipalID: principalID, VerificationTokenHash: verificationDigest[:],
			VerificationDeliveryID: uuid.NullUUID{},
			VerificationExpiresAt:  sql.NullTime{Time: fixedTask9Time.Add(emailVerificationTTL), Valid: true},
		},
		accountFound: true,
		account:      store.IdentityAccount{ID: principalID, State: "pending_email", StateVersion: 1, Locale: "en"},
	}
	application, _, _, participant := newTask9Application(t, config.EmailRequired, tx)

	if err := application.VerifyEmail(context.Background(), VerifyEmailCommand{Token: token, IdempotencyKey: "abcdefghijklmnopqrstuv"}); err != nil {
		t.Fatalf("VerifyEmail with consumer-cleared delivery metadata: %v; operations=%v", err, tx.operations)
	}
	assertTask9Order(t, tx.operations, []string{
		"get_verification", "begin_idempotency", "consume_email", "activate_account", "participant",
		"append_event", "complete_idempotency", "commit",
	})
	if participant.calls != 1 || participant.tx != tx.DBTX() {
		t.Fatal("consumer-cleared verification did not activate the participant in the identity transaction")
	}
	if len(tx.events) != 1 {
		t.Fatal("consumer-cleared verification did not append the account activation event")
	}
}

func TestPasswordResetDeliveryIsGenericDistinctAndExpiresAtThirtyMinutes(t *testing.T) {
	t.Parallel()

	principalID := uuid.New()
	knownTx := &fakeIdentityTransaction{
		identityFound: true, identity: store.IdentityEmailIdentity{PrincipalID: principalID},
		accountFound: true, account: store.IdentityAccount{ID: principalID, State: "active", StateVersion: 2, Locale: "en"},
		credentialFound: true, credential: task9StoredCredential(principalID),
	}
	knownApplication, knownDeriver, knownLimiter, _ := newTask9Application(t, config.EmailRequired, knownTx)
	knownResult, err := knownApplication.CreatePasswordResetDelivery(context.Background(), CreatePasswordResetDeliveryCommand{
		Email: "member@example.test", Locale: "en", IdempotencyKey: "abcdefghijklmnopqrstuv",
	})
	if err != nil || !knownResult.Accepted || knownDeriver.calls != 1 || knownLimiter.calls != 1 {
		t.Fatalf("known reset delivery = %#v, %v; derivations/limits=%d/%d", knownResult, err, knownDeriver.calls, knownLimiter.calls)
	}
	if knownTx.passwordReset.ResetExpiresAt.Time.Sub(fixedTask9Time) != 30*time.Minute {
		t.Fatalf("password reset TTL = %s", knownTx.passwordReset.ResetExpiresAt.Time.Sub(fixedTask9Time))
	}
	assertProtectedDelivery(t, knownApplication.protector, passwordResetDeliveryDomain, knownTx.passwordReset.ResetDeliveryCiphertext, knownTx.passwordReset.ResetDeliveryKeyVersion.Int32, "member@example.test", string(ResetPasswordTemplate), "en")
	if len(knownTx.events) != 1 {
		t.Fatal("eligible password reset did not append one delivery event")
	}
	assertTask9PrivateBinding(t, knownTx.idempotencyCanonical[0], idempotency.CreatePasswordResetDeliveryOperation, []byte("member@example.test"), []byte("en"))
	assertTask9Order(t, knownTx.operations, []string{
		"begin_idempotency", "find_identity", "get_credential", "get_account", "set_password_reset", "append_event", "complete_idempotency", "commit",
	})

	unknownTx := &fakeIdentityTransaction{}
	unknownApplication, unknownDeriver, unknownLimiter, _ := newTask9Application(t, config.EmailRequired, unknownTx)
	unknownResult, err := unknownApplication.CreatePasswordResetDelivery(context.Background(), CreatePasswordResetDeliveryCommand{
		Email: "unknown@example.test", Locale: "en", IdempotencyKey: "abcdefghijklmnopqrstuv",
	})
	if err != nil || unknownResult != knownResult || unknownDeriver.calls != 1 || unknownLimiter.calls != 1 {
		t.Fatalf("unknown reset delivery = %#v, %v; derivations/limits=%d/%d", unknownResult, err, unknownDeriver.calls, unknownLimiter.calls)
	}
	if unknownTx.mutationCount() != 0 || len(unknownTx.events) != 0 {
		t.Fatal("unknown reset delivery mutated identity state")
	}
	assertTask9Order(t, unknownTx.operations, []string{
		"begin_idempotency", "find_identity", "get_credential", "get_account", "complete_idempotency", "commit",
	})

	verificationTx := &fakeIdentityTransaction{
		identityFound: true, identity: store.IdentityEmailIdentity{PrincipalID: principalID},
		accountFound: true, account: store.IdentityAccount{ID: principalID, State: "pending_email", StateVersion: 1, Locale: "en"},
	}
	verificationApplication, _, _, _ := newTask9Application(t, config.EmailRequired, verificationTx)
	if _, err = verificationApplication.CreateEmailVerificationDelivery(context.Background(), CreateEmailVerificationDeliveryCommand{
		Email: "member@example.test", Locale: "en", IdempotencyKey: "abcdefghijklmnopqrstuv",
	}); err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(verificationTx.resetEmail.VerificationTokenHash, knownTx.passwordReset.ResetTokenHash) {
		t.Fatal("verification and password-reset token domains collided")
	}
	assertTask9PrivateBinding(t, verificationTx.idempotencyCanonical[0], idempotency.CreateEmailVerificationDeliveryOperation, []byte("member@example.test"), []byte("en"))
}

func TestPasswordResetConsumesOnceMarksSessionsBeforeCreatingBoundSession(t *testing.T) {
	t.Parallel()

	principalID := uuid.New()
	tx := &fakeIdentityTransaction{
		identityFound: true, identity: store.IdentityEmailIdentity{PrincipalID: principalID},
		accountFound: true, account: store.IdentityAccount{ID: principalID, State: "active", StateVersion: 4, Locale: "en"},
		credentialFound: true, credential: task9StoredCredential(principalID), consumeResetOK: true,
	}
	application, deriver, _, _ := newTask9Application(t, config.EmailRequired, tx)
	token := secret.NewBytes(bytes.Repeat([]byte{0x81}, 32))
	clientKey := [32]byte{0x91}
	newPassword := secret.NewBytes([]byte("a newer correct horse password"))
	result, err := application.ResetPassword(context.Background(), ResetPasswordCommand{
		Email: "member@example.test", Token: token, NewPassword: newPassword,
		ClientSigningPublicKey: clientKey, IdempotencyKey: "abcdefghijklmnopqrstuv",
	})
	if err != nil || deriver.calls != 1 {
		t.Fatalf("ResetPassword = %#v, %v; derivations=%d", result, err, deriver.calls)
	}
	if len(result.AccessToken.Copy()) != 32 || len(result.RefreshToken.Copy()) != 32 ||
		result.AccessExpiresAt.Sub(fixedTask9Time) != 10*time.Minute || result.RefreshIdleExpiresAt.Sub(fixedTask9Time) != 30*24*time.Hour ||
		result.RefreshAbsoluteExpiresAt.Sub(fixedTask9Time) != 90*24*time.Hour {
		t.Fatalf("reset session token contract = %#v", result)
	}
	assertTask9Order(t, tx.operations, []string{
		"get_reset_credential", "find_identity", "lock_principal_refresh", "lock_principal_refresh", "lock_sessions", "get_account", "list_principal_refresh", "begin_idempotency", "consume_password_reset",
		"mark_sessions_review_required", "create_account_session", "insert_account_refresh", "insert_security_event",
		"complete_idempotency", "commit",
	})
	if !bytes.Equal(tx.createdSession.ClientSigningPublicKey, clientKey[:]) {
		t.Fatal("reset session did not bind the supplied client signing key")
	}
	newPasswordCopy := newPassword.Copy()
	defer clear(newPasswordCopy)
	resetDigest := passwordResetTokenDigest(token)
	assertTask9PrivateBinding(t, tx.idempotencyCanonical[0], "reset_password", []byte("member@example.test"), resetDigest[:], newPasswordCopy, clientKey[:])

	failedTx := &fakeIdentityTransaction{
		identityFound: true, identity: store.IdentityEmailIdentity{PrincipalID: principalID},
		accountFound: true, account: tx.account, credentialFound: true, credential: tx.credential,
	}
	failedApplication, failedDeriver, _, _ := newTask9Application(t, config.EmailRequired, failedTx)
	if _, err = failedApplication.ResetPassword(context.Background(), ResetPasswordCommand{
		Email: "member@example.test", Token: token, NewPassword: secret.NewBytes([]byte("a newer correct horse password")),
		ClientSigningPublicKey: clientKey, IdempotencyKey: "abcdefghijklmnopqrstuv",
	}); publicTask9Code(err) != apierrors.AuthenticationFailed || failedDeriver.calls != 1 {
		t.Fatalf("used reset token error/derivations = %v/%d", err, failedDeriver.calls)
	}
}

func TestGracePasswordResetClampsReplacementSessionToFixedAccountDeadline(t *testing.T) {
	// Mutation caught: allowing pending reset under a non-grace deployment, or
	// deriving replacement-session authority from reset time instead of the
	// immutable account creation timestamp.
	principalID := uuid.MustParse("7e1a6a5a-b88e-41e9-9027-27d40cd8e66b")
	tx := &fakeIdentityTransaction{
		identityFound: true, identity: store.IdentityEmailIdentity{PrincipalID: principalID},
		accountFound: true, account: store.IdentityAccount{
			ID: principalID, State: "pending_email", StateVersion: 1, Locale: "en",
			CreatedAt: fixedTask9Time.Add(-23*time.Hour - 55*time.Minute),
		},
		credentialFound: true, credential: task9StoredCredential(principalID), consumeResetOK: true,
	}
	application, _, _, _ := newTask9Application(t, config.EmailGrace, tx)
	result, err := application.ResetPassword(context.Background(), ResetPasswordCommand{
		Email: "member@example.test", Token: secret.NewBytes(bytes.Repeat([]byte{0x81}, 32)),
		NewPassword: secret.NewBytes([]byte("a newer correct horse password")), ClientSigningPublicKey: [32]byte{0x91},
		IdempotencyKey: "abcdefghijklmnopqrstuv",
	})
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Date(2026, 8, 10, 1, 7, 3, 0, time.UTC)
	if result.AccessExpiresAt != deadline || result.RefreshIdleExpiresAt != deadline || result.RefreshAbsoluteExpiresAt != deadline ||
		tx.createdSession.AccessExpiresAt != deadline || tx.createdSession.AbsoluteExpiresAt != deadline ||
		tx.createdRefresh.IdleExpiresAt != deadline || tx.createdRefresh.AbsoluteExpiresAt != deadline {
		t.Fatalf("grace reset deadlines = result:%#v session:%#v refresh:%#v, want %v", result, tx.createdSession, tx.createdRefresh, deadline)
	}
}

func TestCompletedGracePasswordResetReplaySurvivesDeadlineAndRequiredSwitchWithoutMutation(t *testing.T) {
	// Mutations caught: evaluating mutable EmailGrace eligibility before a
	// completed reset replay, or re-consuming the reset credential and minting a
	// new password/session/token/event/idempotency record on replay or denial.
	for _, test := range []struct {
		name       string
		transition func(*Service, *task10MutableClock)
	}{
		{
			name: "deadline equality",
			transition: func(_ *Service, clock *task10MutableClock) {
				clock.now = time.Date(2026, 8, 10, 12, 1, 0, 0, time.UTC)
			},
		},
		{
			name: "switched to required",
			transition: func(application *Service, _ *task10MutableClock) {
				application.security.EmailVerification = config.EmailRequired
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			transaction := activeTask10Transaction()
			transaction.account.State = "pending_email"
			transaction.account.CreatedAt = fixedTask10Time.Add(-23*time.Hour - 59*time.Minute)
			transaction.consumeResetOK = true
			application, deriver, limiter := newTask10Application(t, transaction, &task10ChallengeStore{})
			application.security.EmailVerification = config.EmailGrace
			clock := &task10MutableClock{now: fixedTask10Time}
			application.clock = clock
			ids := &task9UUIDs{}
			uuidCalls := 0
			application.newUUID = func() uuid.UUID {
				uuidCalls++
				return ids.Next()
			}
			command := ResetPasswordCommand{
				Email: "member@example.test", Token: secret.NewBytes(bytes.Repeat([]byte{0x81}, 32)),
				NewPassword: secret.NewBytes([]byte("a newer correct horse password")), ClientSigningPublicKey: [32]byte{0x91},
				IdempotencyKey: "abcdefghijklmnopqrstuv",
			}
			created, err := application.ResetPassword(context.Background(), command)
			if err != nil {
				t.Fatal(err)
			}
			createdAccess := created.AccessToken.Copy()
			createdRefresh := created.RefreshToken.Copy()
			defer clear(createdAccess)
			defer clear(createdRefresh)
			random := application.random.(*task9Random)
			random.mu.Lock()
			randomBytes := random.counter
			random.mu.Unlock()
			createdSessionID := transaction.createdSession.ID
			createdRefreshHash := bytes.Clone(transaction.createdRefresh.TokenHash)
			defer clear(createdRefreshHash)
			createdSessionOps := countTask10Operation(transaction.operations, "create_account_session")
			createdRefreshOps := countTask10Operation(transaction.operations, "insert_account_refresh")
			consumeOps := countTask10Operation(transaction.operations, "consume_password_reset")
			reviewOps := countTask10Operation(transaction.operations, "mark_sessions_review_required")
			eventCount := len(transaction.securityEvents)
			if len(transaction.idempotencyDB.snapshot()) != 1 || deriver.calls != 1 || limiter.calls != 0 || uuidCalls != 2 {
				t.Fatalf("initial reset evidence = records:%d derive:%d limit:%d uuid:%d", len(transaction.idempotencyDB.snapshot()), deriver.calls, limiter.calls, uuidCalls)
			}

			test.transition(application, clock)
			replayed, err := application.ResetPassword(context.Background(), command)
			if err != nil {
				t.Fatalf("completed replay = %v", err)
			}
			replayedAccess := replayed.AccessToken.Copy()
			replayedRefresh := replayed.RefreshToken.Copy()
			defer clear(replayedAccess)
			defer clear(replayedRefresh)
			if !bytes.Equal(replayedAccess, createdAccess) || !bytes.Equal(replayedRefresh, createdRefresh) ||
				replayed.AccessExpiresAt != created.AccessExpiresAt || replayed.RefreshIdleExpiresAt != created.RefreshIdleExpiresAt ||
				replayed.RefreshAbsoluteExpiresAt != created.RefreshAbsoluteExpiresAt {
				t.Fatalf("completed replay = %#v, want byte-identical %#v", replayed, created)
			}
			random.mu.Lock()
			replayRandomBytes := random.counter
			random.mu.Unlock()
			if replayRandomBytes != randomBytes || uuidCalls != 2 || deriver.calls != 1 || limiter.calls != 0 ||
				transaction.createdSession.ID != createdSessionID || !bytes.Equal(transaction.createdRefresh.TokenHash, createdRefreshHash) ||
				countTask10Operation(transaction.operations, "consume_password_reset") != consumeOps ||
				countTask10Operation(transaction.operations, "mark_sessions_review_required") != reviewOps ||
				countTask10Operation(transaction.operations, "create_account_session") != createdSessionOps ||
				countTask10Operation(transaction.operations, "insert_account_refresh") != createdRefreshOps ||
				len(transaction.securityEvents) != eventCount || len(transaction.idempotencyDB.snapshot()) != 1 {
				t.Fatalf("replay side effects = random:%d/%d uuid:%d derive:%d limit:%d consume:%d/%d review:%d/%d session:%d/%d refresh:%d/%d events:%d/%d records:%d",
					replayRandomBytes, randomBytes, uuidCalls, deriver.calls, limiter.calls,
					countTask10Operation(transaction.operations, "consume_password_reset"), consumeOps,
					countTask10Operation(transaction.operations, "mark_sessions_review_required"), reviewOps,
					countTask10Operation(transaction.operations, "create_account_session"), createdSessionOps,
					countTask10Operation(transaction.operations, "insert_account_refresh"), createdRefreshOps,
					len(transaction.securityEvents), eventCount, len(transaction.idempotencyDB.snapshot()))
			}

			newCommand := command
			newCommand.IdempotencyKey = "bcdefghijklmnopqrstuvw"
			if _, err = application.ResetPassword(context.Background(), newCommand); publicTask9Code(err) != apierrors.AuthenticationFailed {
				t.Fatalf("closed grace new-key error = %v", err)
			}
			random.mu.Lock()
			denialRandomBytes := random.counter
			random.mu.Unlock()
			if denialRandomBytes != randomBytes || uuidCalls != 2 || deriver.calls != 2 || limiter.calls != 0 ||
				countTask10Operation(transaction.operations, "consume_password_reset") != consumeOps ||
				countTask10Operation(transaction.operations, "mark_sessions_review_required") != reviewOps ||
				countTask10Operation(transaction.operations, "create_account_session") != createdSessionOps ||
				countTask10Operation(transaction.operations, "insert_account_refresh") != createdRefreshOps ||
				len(transaction.securityEvents) != eventCount || len(transaction.idempotencyDB.snapshot()) != 1 ||
				len(transaction.operations) == 0 || transaction.operations[len(transaction.operations)-1] != "rollback" {
				t.Fatalf("closed grace new-key side effects = random:%d/%d uuid:%d derive:%d consume:%d review:%d session:%d refresh:%d events:%d records:%d operations:%v",
					denialRandomBytes, randomBytes, uuidCalls, deriver.calls,
					countTask10Operation(transaction.operations, "consume_password_reset"),
					countTask10Operation(transaction.operations, "mark_sessions_review_required"),
					countTask10Operation(transaction.operations, "create_account_session"),
					countTask10Operation(transaction.operations, "insert_account_refresh"),
					len(transaction.securityEvents), len(transaction.idempotencyDB.snapshot()), transaction.operations)
			}
		})
	}
}

func TestPendingPasswordResetRequiresCurrentGraceAfterIdempotencyClassificationBeforeConsumption(t *testing.T) {
	// Mutation caught: the former `(active || pending_email)` check admitted a
	// pending account under required/disabled and consumed its reset token. A
	// newly Started record must be rolled back after current eligibility fails.
	for _, mode := range []config.EmailVerificationMode{config.EmailRequired, config.EmailDisabled} {
		t.Run(string(mode), func(t *testing.T) {
			principalID := uuid.MustParse("7e1a6a5a-b88e-41e9-9027-27d40cd8e66b")
			tx := &fakeIdentityTransaction{
				identityFound: true, identity: store.IdentityEmailIdentity{PrincipalID: principalID},
				accountFound: true, account: store.IdentityAccount{ID: principalID, State: "pending_email", CreatedAt: fixedTask9Time.Add(-time.Hour)},
				credentialFound: true, credential: task9StoredCredential(principalID), consumeResetOK: true,
			}
			application, deriver, _, _ := newTask9Application(t, mode, tx)
			_, err := application.ResetPassword(context.Background(), ResetPasswordCommand{
				Email: "member@example.test", Token: secret.NewBytes(bytes.Repeat([]byte{0x81}, 32)),
				NewPassword: secret.NewBytes([]byte("a newer correct horse password")), ClientSigningPublicKey: [32]byte{0x91},
				IdempotencyKey: "abcdefghijklmnopqrstuv",
			})
			if publicTask9Code(err) != apierrors.AuthenticationFailed || deriver.calls != 1 || tx.createdSession.ID != uuid.Nil ||
				tx.consumedReset.PrincipalID != uuid.Nil || len(tx.idempotencyCanonical) != 1 || len(tx.operations) == 0 || tx.operations[len(tx.operations)-1] != "rollback" {
				t.Fatalf("pending reset denial = err:%v derive:%d session:%v consume:%v idempotency:%d operations:%v", err, deriver.calls, tx.createdSession.ID, tx.consumedReset.PrincipalID, len(tx.idempotencyCanonical), tx.operations)
			}
		})
	}
}

func TestEnrollmentPolicyRequiredGraceAndDisabled(t *testing.T) {
	t.Parallel()

	principalID := uuid.New()
	sessionID := uuid.New()
	proof := secret.NewBytes([]byte("correct horse battery staple"))
	for _, test := range []struct {
		name        string
		mode        config.EmailVerificationMode
		state       string
		wantCode    apierrors.Code
		wantMarker  string
		provisional bool
	}{
		{name: "required pending rejected", mode: config.EmailRequired, state: "pending_email", wantCode: apierrors.ActionNotAllowed},
		{name: "required verified standard", mode: config.EmailRequired, state: "active", wantMarker: "standard"},
		{name: "grace pending restricted", mode: config.EmailGrace, state: "pending_email", wantMarker: "trial_restricted", provisional: true},
		{name: "grace verified standard", mode: config.EmailGrace, state: "active", wantMarker: "standard"},
		{name: "disabled local standard", mode: config.EmailDisabled, state: "active", wantMarker: "standard"},
	} {
		t.Run(test.name, func(t *testing.T) {
			tx := &fakeIdentityTransaction{
				accountFound: true, account: store.IdentityAccount{ID: principalID, State: test.state, StateVersion: 2, Locale: "en", CreatedAt: fixedTask9Time.Add(-23 * time.Hour)},
				sessionFound: true, session: store.IdentityAccountSession{ID: sessionID, PrincipalID: principalID, State: "active", AccessExpiresAt: fixedTask9Time.Add(30 * time.Minute), AbsoluteExpiresAt: fixedTask9Time.Add(time.Hour)},
				credentialFound: true, credential: task9StoredCredential(principalID),
			}
			application, deriver, _, _ := newTask9Application(t, test.mode, tx)
			grant, err := application.CreateEnrollmentGrant(context.Background(), CreateEnrollmentGrantCommand{
				PrincipalID: PrincipalID(principalID.String()), SessionID: SessionID(sessionID.String()),
				Reauthentication: Reauthentication{SessionID: SessionID(sessionID.String()), Method: ReauthPassword, Proof: proof},
				IdempotencyKey:   "abcdefghijklmnopqrstuv",
			})
			if test.wantCode != "" {
				if publicTask9Code(err) != test.wantCode || tx.createdGrant.ID != uuid.Nil {
					t.Fatalf("grant rejection = %#v, %v", grant, err)
				}
				return
			}
			if err != nil || grant.PolicyMarker != test.wantMarker || len(grant.Token.Copy()) != 32 || grant.ExpiresAt.Sub(fixedTask9Time) != 10*time.Minute || deriver.calls != 1 {
				t.Fatalf("grant = %#v, %v; derivations=%d", grant, err, deriver.calls)
			}
			assertTask9Order(t, tx.operations, []string{
				"get_credential", "lock_sessions", "get_account", "begin_idempotency", "create_enrollment_grant", "complete_idempotency", "commit",
			})
			if tx.createdGrant.PolicyMarker != test.wantMarker || tx.createdGrant.ProvisionalUntil.Valid != test.provisional {
				t.Fatalf("stored grant policy = %#v", tx.createdGrant)
			}
			if test.provisional && tx.createdGrant.ProvisionalUntil.Time != time.Date(2026, 8, 10, 2, 2, 3, 0, time.UTC) {
				t.Fatalf("provisional boundary = %s", tx.createdGrant.ProvisionalUntil.Time)
			}
			if len(tx.idempotencyCanonical) > 0 {
				proofCopy := proof.Copy()
				defer clear(proofCopy)
				assertTask9PrivateBinding(t, tx.idempotencyCanonical[0], "create_enrollment_grant", principalID[:], sessionID[:], proofCopy)
			}
		})
	}
}

func TestGraceEnrollmentGrantExpiryClampsToFixedDeadline(t *testing.T) {
	// Mutation caught: a late grant using now+10m or now+24h extends the
	// principal's provisional authority past account.created_at+24h.
	principalID := uuid.MustParse("24ee2c85-b4b0-49f1-8b72-acde8cb0a934")
	sessionID := uuid.MustParse("a8e0fb47-6631-46f5-91fc-38370d870e1e")
	tx := &fakeIdentityTransaction{
		accountFound: true, account: store.IdentityAccount{
			ID: principalID, State: "pending_email", StateVersion: 1, Locale: "en",
			CreatedAt: fixedTask9Time.Add(-23*time.Hour - 55*time.Minute),
		},
		sessionFound: true, session: store.IdentityAccountSession{
			ID: sessionID, PrincipalID: principalID, State: "active",
			AccessExpiresAt: fixedTask9Time.Add(5 * time.Minute), AbsoluteExpiresAt: fixedTask9Time.Add(5 * time.Minute),
		},
		credentialFound: true, credential: task9StoredCredential(principalID),
	}
	application, _, _, _ := newTask9Application(t, config.EmailGrace, tx)
	grant, err := application.CreateEnrollmentGrant(context.Background(), CreateEnrollmentGrantCommand{
		PrincipalID: PrincipalID(principalID.String()), SessionID: SessionID(sessionID.String()),
		Reauthentication: Reauthentication{SessionID: SessionID(sessionID.String()), Method: ReauthPassword, Proof: secret.NewBytes([]byte("correct horse battery staple"))},
		IdempotencyKey:   "abcdefghijklmnopqrstuv",
	})
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Date(2026, 8, 10, 1, 7, 3, 0, time.UTC)
	if grant.PolicyMarker != "trial_restricted" || grant.ExpiresAt != deadline || !tx.createdGrant.ProvisionalUntil.Valid ||
		tx.createdGrant.ProvisionalUntil.Time != deadline || tx.createdGrant.ExpiresAt != deadline {
		t.Fatalf("late grace grant = %#v stored:%#v, want %v", grant, tx.createdGrant, deadline)
	}
}

func TestCompletedGraceEnrollmentGrantReplaySurvivesDeadlineAndRequiredSwitchWithoutMutation(t *testing.T) {
	// Mutations caught: current account/session grace authority before replay,
	// or a replay/new-key denial generating another opaque token, UUID, grant,
	// random byte, or durable idempotency record after password proof.
	for _, test := range []struct {
		name       string
		wantNewKey apierrors.Code
		transition func(*Service, *task10MutableClock)
	}{
		{
			name:       "deadline equality",
			wantNewKey: apierrors.AuthenticationFailed,
			transition: func(_ *Service, clock *task10MutableClock) {
				clock.now = time.Date(2026, 8, 10, 12, 1, 0, 0, time.UTC)
			},
		},
		{
			name:       "switched to required",
			wantNewKey: apierrors.ActionNotAllowed,
			transition: func(application *Service, _ *task10MutableClock) {
				application.security.EmailVerification = config.EmailRequired
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			transaction := activeTask10Transaction()
			transaction.account.State = "pending_email"
			transaction.account.CreatedAt = fixedTask10Time.Add(-23*time.Hour - 59*time.Minute)
			deadline := time.Date(2026, 8, 10, 12, 1, 0, 0, time.UTC)
			transaction.sessionFound = true
			transaction.session = store.IdentityAccountSession{
				ID: uuid.MustParse("a8e0fb47-6631-46f5-91fc-38370d870e1e"), PrincipalID: transaction.account.ID, State: "active",
				AccessExpiresAt: deadline, AbsoluteExpiresAt: deadline,
			}
			application, deriver, limiter := newTask10Application(t, transaction, &task10ChallengeStore{})
			application.security.EmailVerification = config.EmailGrace
			clock := &task10MutableClock{now: fixedTask10Time}
			application.clock = clock
			ids := &task9UUIDs{}
			uuidCalls := 0
			application.newUUID = func() uuid.UUID {
				uuidCalls++
				return ids.Next()
			}
			command := CreateEnrollmentGrantCommand{
				PrincipalID: PrincipalID(transaction.account.ID.String()), SessionID: SessionID(transaction.session.ID.String()),
				Reauthentication: Reauthentication{
					SessionID: SessionID(transaction.session.ID.String()), Method: ReauthPassword,
					Proof: secret.NewBytes([]byte("correct horse battery staple")),
				},
				IdempotencyKey: "abcdefghijklmnopqrstuv",
			}
			created, err := application.CreateEnrollmentGrant(context.Background(), command)
			if err != nil {
				t.Fatal(err)
			}
			createdToken := created.Token.Copy()
			defer clear(createdToken)
			random := application.random.(*task9Random)
			random.mu.Lock()
			randomBytes := random.counter
			random.mu.Unlock()
			createdGrantID := transaction.createdGrant.ID
			createdTokenHash := bytes.Clone(transaction.createdGrant.TokenHash)
			defer clear(createdTokenHash)
			grantOps := countTask10Operation(transaction.operations, "create_enrollment_grant")
			if len(transaction.idempotencyDB.snapshot()) != 1 || deriver.calls != 1 || limiter.calls != 0 || uuidCalls != 1 {
				t.Fatalf("initial grant evidence = records:%d derive:%d limit:%d uuid:%d", len(transaction.idempotencyDB.snapshot()), deriver.calls, limiter.calls, uuidCalls)
			}

			test.transition(application, clock)
			replayed, err := application.CreateEnrollmentGrant(context.Background(), command)
			if err != nil {
				t.Fatalf("completed replay = %v", err)
			}
			replayedToken := replayed.Token.Copy()
			defer clear(replayedToken)
			if !bytes.Equal(replayedToken, createdToken) || replayed.ExpiresAt != created.ExpiresAt || replayed.PolicyMarker != created.PolicyMarker {
				t.Fatalf("completed replay = %#v, want byte-identical %#v", replayed, created)
			}
			random.mu.Lock()
			replayRandomBytes := random.counter
			random.mu.Unlock()
			if replayRandomBytes != randomBytes || uuidCalls != 1 || deriver.calls != 2 || limiter.calls != 0 ||
				transaction.createdGrant.ID != createdGrantID || !bytes.Equal(transaction.createdGrant.TokenHash, createdTokenHash) ||
				countTask10Operation(transaction.operations, "create_enrollment_grant") != grantOps || len(transaction.idempotencyDB.snapshot()) != 1 {
				t.Fatalf("replay side effects = random:%d/%d uuid:%d derive:%d limit:%d grant:%d/%d records:%d",
					replayRandomBytes, randomBytes, uuidCalls, deriver.calls, limiter.calls,
					countTask10Operation(transaction.operations, "create_enrollment_grant"), grantOps, len(transaction.idempotencyDB.snapshot()))
			}

			newCommand := command
			newCommand.IdempotencyKey = "bcdefghijklmnopqrstuvw"
			if _, err = application.CreateEnrollmentGrant(context.Background(), newCommand); publicTask9Code(err) != test.wantNewKey {
				t.Fatalf("closed grace new-key error = %v, want %s", err, test.wantNewKey)
			}
			random.mu.Lock()
			denialRandomBytes := random.counter
			random.mu.Unlock()
			if denialRandomBytes != randomBytes || uuidCalls != 1 || deriver.calls != 3 || limiter.calls != 0 ||
				countTask10Operation(transaction.operations, "create_enrollment_grant") != grantOps || len(transaction.idempotencyDB.snapshot()) != 1 ||
				len(transaction.operations) == 0 || transaction.operations[len(transaction.operations)-1] != "rollback" {
				t.Fatalf("closed grace new-key side effects = random:%d/%d uuid:%d derive:%d limit:%d grant:%d records:%d operations:%v",
					denialRandomBytes, randomBytes, uuidCalls, deriver.calls, limiter.calls,
					countTask10Operation(transaction.operations, "create_enrollment_grant"), len(transaction.idempotencyDB.snapshot()), transaction.operations)
			}
		})
	}
}

func TestEnrollmentGrantRejectsExpiredOrReviewSessionAfterOneConstantCostProof(t *testing.T) {
	t.Parallel()

	principalID := uuid.New()
	sessionID := uuid.New()
	for _, test := range []struct {
		name     string
		state    string
		accessAt time.Time
		absolute time.Time
	}{
		{name: "access expired", state: "active", accessAt: fixedTask9Time.Add(-time.Second), absolute: fixedTask9Time.Add(time.Hour)},
		{name: "absolute expired", state: "active", accessAt: fixedTask9Time.Add(time.Hour), absolute: fixedTask9Time.Add(-time.Second)},
		{name: "review required", state: "review_required", accessAt: fixedTask9Time.Add(time.Hour), absolute: fixedTask9Time.Add(2 * time.Hour)},
	} {
		t.Run(test.name, func(t *testing.T) {
			tx := &fakeIdentityTransaction{
				accountFound: true, account: store.IdentityAccount{ID: principalID, State: "active", StateVersion: 2, Locale: "en"},
				sessionFound: true, session: store.IdentityAccountSession{
					ID: sessionID, PrincipalID: principalID, State: test.state,
					AccessExpiresAt: test.accessAt, AbsoluteExpiresAt: test.absolute,
				},
				credentialFound: true, credential: task9StoredCredential(principalID),
			}
			application, deriver, _, _ := newTask9Application(t, config.EmailRequired, tx)
			_, err := application.CreateEnrollmentGrant(context.Background(), CreateEnrollmentGrantCommand{
				PrincipalID: PrincipalID(principalID.String()), SessionID: SessionID(sessionID.String()),
				Reauthentication: Reauthentication{
					SessionID: SessionID(sessionID.String()), Method: ReauthPassword,
					Proof: secret.NewBytes([]byte("correct horse battery staple")),
				},
				IdempotencyKey: "abcdefghijklmnopqrstuv",
			})
			if publicTask9Code(err) != apierrors.AuthenticationFailed || deriver.calls != 1 || tx.createdGrant.ID != uuid.Nil {
				t.Fatalf("expired/review grant error=%v derivations=%d grant=%v", err, deriver.calls, tx.createdGrant.ID)
			}
		})
	}
}

func TestTask9PrivateCommandsCannotFormatLogOrMarshal(t *testing.T) {
	t.Parallel()

	emailCanary := "private-canary@example.test"
	tokenCanary := bytes.Repeat([]byte("TOKEN-CANARY"), 3)[:32]
	passwordCanary := []byte("PASSWORD-CANARY-PRIVATE")
	principalID := PrincipalID(uuid.New().String())
	sessionID := SessionID(uuid.New().String())
	values := []any{
		RegisterAccountCommand{Email: emailCanary, Password: secret.NewBytes(passwordCanary), Locale: "en", IdempotencyKey: "abcdefghijklmnopqrstuv"},
		CreateEmailVerificationDeliveryCommand{Email: emailCanary, Locale: "en", IdempotencyKey: "abcdefghijklmnopqrstuv"},
		CreatePasswordResetDeliveryCommand{Email: emailCanary, Locale: "en", IdempotencyKey: "abcdefghijklmnopqrstuv"},
		VerifyEmailCommand{Token: secret.NewBytes(tokenCanary), IdempotencyKey: "abcdefghijklmnopqrstuv"},
		ResetPasswordCommand{Email: emailCanary, Token: secret.NewBytes(tokenCanary), NewPassword: secret.NewBytes(passwordCanary), IdempotencyKey: "abcdefghijklmnopqrstuv"},
		CreateEnrollmentGrantCommand{
			PrincipalID: principalID, SessionID: sessionID,
			Reauthentication: Reauthentication{SessionID: sessionID, Method: ReauthPassword, Proof: secret.NewBytes(passwordCanary)},
			IdempotencyKey:   "abcdefghijklmnopqrstuv",
		},
	}
	for _, value := range values {
		rendered := fmt.Sprintf("%+v", value) + slog.Any("command", value).Value.String()
		for _, forbidden := range []string{emailCanary, string(tokenCanary), string(passwordCanary), string(principalID), string(sessionID)} {
			if strings.Contains(rendered, forbidden) {
				t.Fatalf("private command formatting exposed input: %s", rendered)
			}
		}
		if encoded, err := json.Marshal(value); err == nil || encoded != nil {
			t.Fatalf("private command unexpectedly marshaled: %s", encoded)
		}
	}
}

func task9StoredCredential(principalID uuid.UUID) store.IdentityPasswordCredential {
	return store.IdentityPasswordCredential{
		PrincipalID: principalID, PolicyVersion: 1, MemoryKib: 65536, TimeCost: 3, Parallelism: 4,
		Salt: bytes.Repeat([]byte{0x12}, 16), PasswordHash: bytes.Repeat([]byte{0xa5}, 32),
		ResetExpiresAt:          sql.NullTime{Time: fixedTask9Time.Add(30 * time.Minute), Valid: true},
		ResetDeliveryKeyVersion: pgtype.Int4{Int32: 7, Valid: true},
	}
}
