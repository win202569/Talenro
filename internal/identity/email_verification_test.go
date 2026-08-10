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

	tx = &fakeIdentityTransaction{}
	application, _, _, participant = newTask9Application(t, config.EmailRequired, tx)
	if err := application.VerifyEmail(context.Background(), VerifyEmailCommand{Token: token, IdempotencyKey: "abcdefghijklmnopqrstuv"}); publicTask9Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("used/unknown verification token error = %v", err)
	}
	if participant.calls != 0 {
		t.Fatal("unknown verification token reached participant")
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
	assertTask9Order(t, knownTx.operations, []string{
		"begin_idempotency", "find_identity", "get_account", "get_credential", "set_password_reset", "append_event", "complete_idempotency", "commit",
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
		"begin_idempotency", "find_identity", "get_account", "get_credential", "complete_idempotency", "commit",
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
	result, err := application.ResetPassword(context.Background(), ResetPasswordCommand{
		Email: "member@example.test", Token: token, NewPassword: secret.NewBytes([]byte("a newer correct horse password")),
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
		"find_identity", "get_account", "get_credential", "begin_idempotency", "consume_password_reset",
		"mark_sessions_review_required", "create_account_session", "insert_account_refresh", "insert_security_event",
		"complete_idempotency", "commit",
	})
	if !bytes.Equal(tx.createdSession.ClientSigningPublicKey, clientKey[:]) {
		t.Fatal("reset session did not bind the supplied client signing key")
	}

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
				accountFound: true, account: store.IdentityAccount{ID: principalID, State: test.state, StateVersion: 2, Locale: "en"},
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
			if tx.createdGrant.PolicyMarker != test.wantMarker || tx.createdGrant.ProvisionalUntil.Valid != test.provisional {
				t.Fatalf("stored grant policy = %#v", tx.createdGrant)
			}
			if test.provisional && tx.createdGrant.ProvisionalUntil.Time.Sub(fixedTask9Time) != 24*time.Hour {
				t.Fatalf("provisional boundary = %s", tx.createdGrant.ProvisionalUntil.Time.Sub(fixedTask9Time))
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
