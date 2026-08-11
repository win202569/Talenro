package identity

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-webauthn/webauthn/protocol"
	wa "github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"

	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

// These protocol-minimal option assertions follow the WebAuthn Level 3 data model:
// https://www.w3.org/TR/webauthn-3/
func TestPasskeyRegistrationOptionsUseOpaquePrincipalAndExactPolicy(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	executor := &fakeChallengeExecutor{setAllowed: true}
	ceremonies, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	application := newTask11WebAuthnApplication(t, transaction, ceremonies)
	command := BeginPasskeyRegistrationCommand{
		PrincipalID: PrincipalID(transaction.account.ID.String()), DisplayName: "Primary account",
		Reauthentication: task11PasswordReauthentication(transaction.session.ID),
		IdempotencyKey:   "task11-passkey-register-01",
	}
	options, err := application.BeginPasskeyRegistration(context.Background(), command)
	if err != nil {
		t.Fatalf("begin passkey registration: %v", err)
	}
	var payload struct {
		CeremonyID string `json:"ceremony_id"`
		PublicKey  struct {
			RP struct {
				ID string `json:"id"`
			} `json:"rp"`
			User struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				DisplayName string `json:"displayName"`
			} `json:"user"`
			Timeout                int                             `json:"timeout"`
			Attestation            protocol.ConveyancePreference   `json:"attestation"`
			AuthenticatorSelection protocol.AuthenticatorSelection `json:"authenticatorSelection"`
		} `json:"publicKey"`
	}
	if err := json.Unmarshal(options, &payload); err != nil {
		t.Fatalf("decode registration options: %v", err)
	}
	userHandle, err := base64.RawURLEncoding.DecodeString(payload.PublicKey.User.ID)
	if err != nil {
		t.Fatalf("decode user handle: %v", err)
	}
	if !bytes.Equal(userHandle, transaction.account.ID[:]) || bytes.Contains(userHandle, []byte("@")) {
		t.Fatalf("user handle is not the opaque principal: %x", userHandle)
	}
	if payload.PublicKey.User.Name == "member@example.test" || payload.PublicKey.User.DisplayName != "Primary account" {
		t.Fatalf("unexpected user display fields: %+v", payload.PublicKey.User)
	}
	if payload.PublicKey.RP.ID != "login.example.test" || payload.PublicKey.Timeout != int((2*time.Minute).Milliseconds()) ||
		payload.PublicKey.Attestation != protocol.PreferNoAttestation ||
		payload.PublicKey.AuthenticatorSelection.ResidentKey != protocol.ResidentKeyRequirementPreferred ||
		payload.PublicKey.AuthenticatorSelection.UserVerification != protocol.VerificationRequired {
		t.Fatalf("unexpected registration policy: %+v", payload.PublicKey)
	}
	if _, parsed := parseCanonicalIdentityUUID(payload.CeremonyID); !parsed {
		t.Fatalf("ceremony id is not canonical random UUID: %q", payload.CeremonyID)
	}
	if executor.setKey != "talenro:identity:webauthn-registration:"+payload.CeremonyID || executor.setTTL != 2*time.Minute {
		t.Fatalf("registration ceremony key/ttl = %q/%s", executor.setKey, executor.setTTL)
	}
	record, err := decodeWebAuthnCeremony(payload.CeremonyID, WebAuthnRegistrationCeremony, executor.setValue)
	if err != nil {
		t.Fatalf("decode stored registration ceremony: %v", err)
	}
	if !bytes.Equal(record.Session.UserID, transaction.account.ID[:]) || record.Session.RelyingPartyID != "login.example.test" ||
		record.Session.UserVerification != protocol.VerificationRequired || !record.ExpiresAt.Equal(record.Session.Expires) {
		t.Fatalf("unexpected stored registration ceremony: %+v", record)
	}
	replayed, err := application.BeginPasskeyRegistration(context.Background(), command)
	if err != nil {
		t.Fatalf("replay begin passkey registration: %v", err)
	}
	if !bytes.Equal(replayed, options) || executor.setKey != "talenro:identity:webauthn-registration:"+payload.CeremonyID {
		t.Fatalf("begin replay created a second ceremony: key=%q", executor.setKey)
	}
}

func TestPasskeyRegistrationRejectsEleventhActiveCredential(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	transaction.passkeys = task11Passkeys(transaction.account.ID, 10)
	executor := &fakeChallengeExecutor{setAllowed: true}
	ceremonies, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	application := newTask11WebAuthnApplication(t, transaction, ceremonies)
	_, err = application.BeginPasskeyRegistration(context.Background(), BeginPasskeyRegistrationCommand{
		PrincipalID: PrincipalID(transaction.account.ID.String()), DisplayName: "Primary account",
		Reauthentication: task11PasswordReauthentication(transaction.session.ID),
		IdempotencyKey:   "task11-passkey-register-02",
	})
	if publicTask9Code(err) != apierrors.StateConflict {
		t.Fatalf("eleventh passkey error = %v", err)
	}
	if executor.setKey != "" {
		t.Fatal("rejected passkey registration created a Redis ceremony")
	}
}

func TestPasskeyAuthenticationOptionsUseDistinctRequiredUVCeremony(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	executor := &fakeChallengeExecutor{setAllowed: true}
	ceremonies, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	application := newTask11WebAuthnApplication(t, transaction, ceremonies)
	options, err := application.BeginPasskeyAuthentication(context.Background(), BeginPasskeyAuthenticationCommand{})
	if err != nil {
		t.Fatalf("begin passkey authentication: %v", err)
	}
	var payload struct {
		CeremonyID string `json:"ceremony_id"`
		PublicKey  struct {
			RPID             string                               `json:"rpId"`
			Timeout          int                                  `json:"timeout"`
			UserVerification protocol.UserVerificationRequirement `json:"userVerification"`
		} `json:"publicKey"`
	}
	if err := json.Unmarshal(options, &payload); err != nil {
		t.Fatalf("decode authentication options: %v", err)
	}
	if payload.PublicKey.RPID != "login.example.test" || payload.PublicKey.Timeout != int((2*time.Minute).Milliseconds()) ||
		payload.PublicKey.UserVerification != protocol.VerificationRequired {
		t.Fatalf("unexpected authentication policy: %+v", payload.PublicKey)
	}
	if executor.setKey != "talenro:identity:webauthn-authentication:"+payload.CeremonyID || executor.setTTL != 2*time.Minute {
		t.Fatalf("authentication ceremony key/ttl = %q/%s", executor.setKey, executor.setTTL)
	}
}

func TestPasskeyCounterSQLRejectsEqualNonzeroReplayButAllowsZeroCounter(t *testing.T) {
	t.Parallel()

	db := &task11RecordingDBTX{}
	queries := store.New(db)
	_, err := queries.UpdatePasskeyCounter(context.Background(), store.UpdatePasskeyCounterParams{
		CredentialID: []byte("0123456789abcdef"), PrincipalID: uuid.New(), SignCount: 7,
		ProtocolFlags: 5, UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(db.lastSQL, "sign_count < $3") || !strings.Contains(db.lastSQL, "sign_count = 0 AND $3 = 0") ||
		strings.Contains(db.lastSQL, "sign_count <= $3") {
		t.Fatalf("counter SQL does not enforce strict nonzero advance: %s", db.lastSQL)
	}
}

func TestPasskeyAuthoritySQLLocksRowsBeforeAccount(t *testing.T) {
	t.Parallel()

	db := &task11RecordingDBTX{}
	_, _ = store.New(db).ListActivePasskeys(context.Background(), uuid.New())
	if !strings.Contains(db.lastSQL, "ORDER BY created_at, credential_id") || !strings.Contains(db.lastSQL, "FOR UPDATE") {
		t.Fatalf("active passkey authority query is not stably ordered and locking: %s", db.lastSQL)
	}
}

// This fixture is constructed from the WebAuthn Level 3 registration data model:
// https://www.w3.org/TR/webauthn-3/
func TestPasskeyRegistrationPersistsOnlyAfterLibraryVerification(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	executor := &fakeChallengeExecutor{setAllowed: true}
	ceremonies, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	application := newTask11WebAuthnApplication(t, transaction, ceremonies)
	begin := BeginPasskeyRegistrationCommand{
		PrincipalID: PrincipalID(transaction.account.ID.String()), DisplayName: "Primary account",
		Reauthentication: task11PasswordReauthentication(transaction.session.ID),
		IdempotencyKey:   "task11-passkey-register-03",
	}
	options, err := application.BeginPasskeyRegistration(context.Background(), begin)
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		CeremonyID string `json:"ceremony_id"`
	}
	if err := json.Unmarshal(options, &envelope); err != nil {
		t.Fatal(err)
	}
	record, err := decodeWebAuthnCeremony(envelope.CeremonyID, WebAuthnRegistrationCeremony, executor.setValue)
	if err != nil {
		t.Fatal(err)
	}
	credentialID := bytes.Repeat([]byte{0x51}, 32)
	response := task11RegistrationResponse(t, record.Session.Challenge, "https://login.example.test", credentialID)
	executor.runValues = []any{bytes.Clone(executor.setValue)}
	transaction.operations = nil
	finish := FinishPasskeyRegistrationCommand{
		PrincipalID: PrincipalID(transaction.account.ID.String()), CeremonyID: envelope.CeremonyID,
		Response: response, Reauthentication: task11PasswordReauthentication(transaction.session.ID),
		IdempotencyKey: "task11-passkey-finish-0001",
	}
	err = application.FinishPasskeyRegistration(context.Background(), finish)
	if err != nil {
		t.Fatalf("finish passkey registration: %v", err)
	}
	created := transaction.createdPasskey
	if !bytes.Equal(created.CredentialID, credentialID) || created.PrincipalID != transaction.account.ID ||
		len(created.PublicKey) < 32 || created.AttestationFormat != "none" || created.SignCount != 0 ||
		created.ProtocolFlags&int16(protocol.FlagUserVerified) == 0 {
		t.Fatalf("unexpected persisted passkey: %+v", created)
	}
	if got := strings.Join(transaction.operations, ","); !strings.Contains(got, "create_passkey") ||
		strings.Index(got, "create_passkey") > strings.Index(got, "complete_idempotency") {
		t.Fatalf("passkey persistence order = %s", got)
	}
	transaction.passkeys = task11Passkeys(transaction.account.ID, maximumActivePasskeys)
	if err = application.FinishPasskeyRegistration(context.Background(), finish); err != nil {
		t.Fatalf("replay finished registration: %v", err)
	}
	if executor.runCalls != 1 || strings.Count(strings.Join(transaction.operations, ","), "create_passkey") != 1 {
		t.Fatalf("registration replay consumed ceremony or persisted twice: redis=%d operations=%v", executor.runCalls, transaction.operations)
	}
}

func TestPasskeyRegistrationBoundsResponseBeforeDependency(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	executor := &fakeChallengeExecutor{setAllowed: true}
	ceremonies, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	application := newTask11WebAuthnApplication(t, transaction, ceremonies)
	err = application.FinishPasskeyRegistration(context.Background(), FinishPasskeyRegistrationCommand{
		PrincipalID:      PrincipalID(transaction.account.ID.String()),
		CeremonyID:       "2dddc425-1bd2-4c33-9b0f-658082addbae",
		Response:         bytes.Repeat([]byte{'x'}, maxWebAuthnCeremonyBytes+1),
		Reauthentication: task11PasswordReauthentication(transaction.session.ID),
		IdempotencyKey:   "task11-passkey-finish-0002",
	})
	if publicTask9Code(err) != apierrors.MalformedRequest {
		t.Fatalf("oversized response error = %v", err)
	}
	if executor.runCalls != 0 || transaction.createdPasskey.PrincipalID != uuid.Nil {
		t.Fatal("oversized WebAuthn response reached Redis or persistence")
	}
}

// This fixture is constructed from the WebAuthn Level 3 assertion data model:
// https://www.w3.org/TR/webauthn-3/
func TestPasskeyAuthenticationVerifiesAssertionAdvancesCounterAndCreatesSession(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	credentialID := bytes.Repeat([]byte{0x61}, 32)
	privateKey, publicKey := task11PasskeyKey(t)
	transaction.passkeys = []store.IdentityPasskeyCredential{{
		CredentialID: credentialID, PrincipalID: transaction.account.ID, PublicKey: publicKey,
		AttestationFormat: "none", ProtocolFlags: int16(protocol.FlagUserPresent | protocol.FlagUserVerified),
		SignCount: 1, State: "active", CreatedAt: fixedTask10Time,
	}}
	executor := &fakeChallengeExecutor{setAllowed: true}
	ceremonies, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	application := newTask11WebAuthnApplication(t, transaction, ceremonies)
	options, err := application.BeginPasskeyAuthentication(context.Background(), BeginPasskeyAuthenticationCommand{})
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		CeremonyID string `json:"ceremony_id"`
	}
	if err := json.Unmarshal(options, &envelope); err != nil {
		t.Fatal(err)
	}
	record, err := decodeWebAuthnCeremony(envelope.CeremonyID, WebAuthnAuthenticationCeremony, executor.setValue)
	if err != nil {
		t.Fatal(err)
	}
	response := task11AuthenticationResponse(
		t, record.Session.Challenge, "https://login.example.test", transaction.account.ID, credentialID, privateKey, 2,
	)
	executor.runValues = []any{bytes.Clone(executor.setValue)}
	transaction.operations = nil
	clientKey := [32]byte{9, 8, 7, 6}
	command := CreateSessionCommand{
		Method: SessionPasskey, WebAuthnCeremonyID: envelope.CeremonyID, WebAuthnResponse: response,
		ClientSigningPublicKey: clientKey, IdempotencyKey: "task11-passkey-auth-0001",
	}
	result, err := application.CreateSession(context.Background(), command)
	if err != nil {
		t.Fatalf("finish passkey authentication: %v", err)
	}
	access := result.AccessToken.Copy()
	refresh := result.RefreshToken.Copy()
	defer clear(access)
	defer clear(refresh)
	if len(access) == 0 || len(refresh) == 0 {
		t.Fatal("passkey authentication did not create tokens")
	}
	if transaction.updatedPasskey.SignCount != 2 || !bytes.Equal(transaction.updatedPasskey.CredentialID, credentialID) ||
		transaction.updatedPasskey.ProtocolFlags&int16(protocol.FlagUserVerified) == 0 {
		t.Fatalf("unexpected passkey counter update: %+v", transaction.updatedPasskey)
	}
	if transaction.createdSession.PrincipalID != transaction.account.ID || !bytes.Equal(transaction.createdSession.ClientSigningPublicKey, clientKey[:]) {
		t.Fatalf("unexpected passkey session binding: %+v", transaction.createdSession)
	}
	operations := strings.Join(transaction.operations, ",")
	if !strings.Contains(operations, "update_passkey") || strings.Index(operations, "create_account_session") < strings.Index(operations, "update_passkey") ||
		strings.Index(operations, "complete_idempotency") < strings.Index(operations, "create_account_session") {
		t.Fatalf("passkey authentication order = %s", operations)
	}
	replayed, err := application.CreateSession(context.Background(), command)
	if err != nil {
		t.Fatalf("replay passkey authentication: %v", err)
	}
	replayedAccess := replayed.AccessToken.Copy()
	replayedRefresh := replayed.RefreshToken.Copy()
	defer clear(replayedAccess)
	defer clear(replayedRefresh)
	if !bytes.Equal(replayedAccess, access) || !bytes.Equal(replayedRefresh, refresh) || executor.runCalls != 1 ||
		strings.Count(strings.Join(transaction.operations, ","), "create_account_session") != 1 {
		t.Fatalf("authentication replay consumed ceremony or created a second session: redis=%d operations=%v", executor.runCalls, transaction.operations)
	}
}

// These negative fixtures mutate only WebAuthn Level 3 protocol fields:
// https://www.w3.org/TR/webauthn-3/
func TestPasskeyAuthenticationRejectsProtocolSubstitutionBeforeMutation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		challenge    func(string) string
		origin       string
		crossOrigin  bool
		userVerified bool
	}{
		{
			name: "wrong origin", challenge: func(value string) string { return value },
			origin: "https://substitute.example.test", userVerified: true,
		},
		{
			name: "wrong challenge", challenge: task11DifferentChallenge,
			origin: "https://login.example.test", userVerified: true,
		},
		{
			name: "cross origin", challenge: func(value string) string { return value },
			origin: "https://login.example.test", crossOrigin: true, userVerified: true,
		},
		{
			name: "missing user verification", challenge: func(value string) string { return value },
			origin: "https://login.example.test", userVerified: false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			transaction := activeTask11TOTPTransaction()
			credentialID := bytes.Repeat([]byte{0x71}, 32)
			privateKey, publicKey := task11PasskeyKey(t)
			transaction.passkeys = []store.IdentityPasskeyCredential{{
				CredentialID: credentialID, PrincipalID: transaction.account.ID, PublicKey: publicKey,
				AttestationFormat: "none", ProtocolFlags: int16(protocol.FlagUserPresent | protocol.FlagUserVerified),
				SignCount: 1, State: "active", CreatedAt: fixedTask10Time,
			}}
			executor := &fakeChallengeExecutor{setAllowed: true}
			ceremonies, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
			if err != nil {
				t.Fatal(err)
			}
			application := newTask11WebAuthnApplication(t, transaction, ceremonies)
			options, err := application.BeginPasskeyAuthentication(context.Background(), BeginPasskeyAuthenticationCommand{})
			if err != nil {
				t.Fatal(err)
			}
			var envelope struct {
				CeremonyID string `json:"ceremony_id"`
			}
			if err = json.Unmarshal(options, &envelope); err != nil {
				t.Fatal(err)
			}
			record, err := decodeWebAuthnCeremony(envelope.CeremonyID, WebAuthnAuthenticationCeremony, executor.setValue)
			if err != nil {
				t.Fatal(err)
			}
			response := task11AuthenticationResponseWithPolicy(
				t, test.challenge(record.Session.Challenge), test.origin, transaction.account.ID,
				credentialID, privateKey, 2, test.crossOrigin, test.userVerified,
			)
			executor.runValues = []any{bytes.Clone(executor.setValue)}
			transaction.operations = nil
			command := FinishPasskeyAuthenticationCommand{
				CeremonyID: envelope.CeremonyID, Response: response,
				ClientSigningPublicKey: [32]byte{4, 3, 2, 1}, IdempotencyKey: "task11-passkey-negative-01",
			}
			if _, err = application.FinishPasskeyAuthentication(context.Background(), command); publicTask9Code(err) != apierrors.AuthenticationFailed {
				t.Fatalf("protocol substitution error = %v", err)
			}
			if transaction.updatedPasskey.PrincipalID != uuid.Nil || transaction.createdSession.PrincipalID != uuid.Nil {
				t.Fatal("protocol substitution mutated a passkey or created a session")
			}
			if _, err = application.FinishPasskeyAuthentication(context.Background(), command); publicTask9Code(err) != apierrors.AuthenticationFailed {
				t.Fatalf("consumed ceremony replay error = %v", err)
			}
			if executor.runCalls != 2 || transaction.updatedPasskey.PrincipalID != uuid.Nil || transaction.createdSession.PrincipalID != uuid.Nil {
				t.Fatalf("protocol substitution ceremony was not single-use: redis=%d operations=%v", executor.runCalls, transaction.operations)
			}
		})
	}
}

func TestPasskeyRevocationIsIndependentAndRequiresReauthentication(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	credentialID := bytes.Repeat([]byte{0x41}, 32)
	transaction.passkeys = task11Passkeys(transaction.account.ID, 1)
	transaction.passkeys[0].CredentialID = credentialID
	transaction.totpFound = true
	transaction.totp = store.IdentityTotpCredential{PrincipalID: transaction.account.ID, State: "active"}
	transaction.recoveryFound = true
	transaction.recovery = store.IdentityRecoveryCodeSet{
		ID: uuid.New(), PrincipalID: transaction.account.ID, Generation: 1,
		CodeHashes: task11RecoveryHashes(), State: "active",
	}
	application, _ := newTask11TOTPApplication(t, transaction)
	command := RevokePasskeyCommand{
		PrincipalID: PrincipalID(transaction.account.ID.String()), CredentialID: credentialID,
		Reauthentication: task11PasswordReauthentication(transaction.session.ID),
		IdempotencyKey:   "task11-passkey-revoke-001",
	}
	err := application.RevokePasskey(context.Background(), command)
	if err != nil {
		t.Fatalf("revoke passkey: %v", err)
	}
	if !bytes.Equal(transaction.revokedPasskey.CredentialID, credentialID) || transaction.revokedPasskey.PrincipalID != transaction.account.ID {
		t.Fatalf("unexpected passkey revocation: %+v", transaction.revokedPasskey)
	}
	operations := strings.Join(transaction.operations, ",")
	if strings.Contains(operations, "revoke_totp") || strings.Contains(operations, "revoke_recovery") {
		t.Fatalf("passkey revocation touched another factor: %s", operations)
	}
	transaction.passkeys = nil
	if err = application.RevokePasskey(context.Background(), command); err != nil {
		t.Fatalf("replay passkey revocation: %v", err)
	}
	if strings.Count(strings.Join(transaction.operations, ","), "revoke_passkey") != 1 {
		t.Fatalf("passkey replay revoked twice: %v", transaction.operations)
	}
}

func TestPasskeyAuthenticationRejectsCounterRollbackBeforeSession(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	credentialID := bytes.Repeat([]byte{0x71}, 32)
	privateKey, publicKey := task11PasskeyKey(t)
	transaction.passkeys = []store.IdentityPasskeyCredential{{
		CredentialID: credentialID, PrincipalID: transaction.account.ID, PublicKey: publicKey,
		AttestationFormat: "none", ProtocolFlags: int16(protocol.FlagUserPresent | protocol.FlagUserVerified),
		SignCount: 5, State: "active", CreatedAt: fixedTask10Time,
	}}
	executor := &fakeChallengeExecutor{setAllowed: true}
	ceremonies, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	application := newTask11WebAuthnApplication(t, transaction, ceremonies)
	options, err := application.BeginPasskeyAuthentication(context.Background(), BeginPasskeyAuthenticationCommand{})
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		CeremonyID string `json:"ceremony_id"`
	}
	if err := json.Unmarshal(options, &envelope); err != nil {
		t.Fatal(err)
	}
	record, err := decodeWebAuthnCeremony(envelope.CeremonyID, WebAuthnAuthenticationCeremony, executor.setValue)
	if err != nil {
		t.Fatal(err)
	}
	executor.runValues = []any{bytes.Clone(executor.setValue)}
	_, err = application.FinishPasskeyAuthentication(context.Background(), FinishPasskeyAuthenticationCommand{
		CeremonyID: envelope.CeremonyID,
		Response: task11AuthenticationResponse(
			t, record.Session.Challenge, "https://login.example.test", transaction.account.ID, credentialID, privateKey, 4,
		),
		ClientSigningPublicKey: [32]byte{1}, IdempotencyKey: "task11-passkey-auth-0002",
	})
	if publicTask9Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("counter rollback error = %v", err)
	}
	if transaction.updatedPasskey.PrincipalID != uuid.Nil || transaction.createdSession.PrincipalID != uuid.Nil {
		t.Fatal("counter rollback updated credential or created session")
	}
}

func TestWebAuthnCeremonyStoreUsesSeparateBoundedSingleUseDomains(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	records := []struct {
		name      string
		operation WebAuthnCeremonyOperation
		ceremony  string
		wantKey   string
	}{
		{
			name: "registration", operation: WebAuthnRegistrationCeremony,
			ceremony: "2dddc425-1bd2-4c33-9b0f-658082addbae",
			wantKey:  "talenro:identity:webauthn-registration:2dddc425-1bd2-4c33-9b0f-658082addbae",
		},
		{
			name: "authentication", operation: WebAuthnAuthenticationCeremony,
			ceremony: "db4b2819-2a0f-4a44-8817-58ca60b24cb2",
			wantKey:  "talenro:identity:webauthn-authentication:db4b2819-2a0f-4a44-8817-58ca60b24cb2",
		},
	}

	for _, test := range records {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			record := WebAuthnCeremonyRecord{
				CeremonyID: test.ceremony,
				Operation:  test.operation,
				Session: wa.SessionData{
					Challenge:        "opaque-library-challenge",
					RelyingPartyID:   "example.com",
					UserID:           []byte{1, 2, 3, 4},
					Expires:          now.Add(2 * time.Minute),
					UserVerification: protocol.VerificationRequired,
				},
				ExpiresAt: now.Add(2 * time.Minute),
			}
			wire, err := encodeWebAuthnCeremony(record)
			if err != nil {
				t.Fatalf("encode ceremony: %v", err)
			}
			executor := &fakeChallengeExecutor{setAllowed: true, runValues: []any{wire, nil}}
			store, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
			if err != nil {
				t.Fatalf("new store: %v", err)
			}

			if err := store.CreateWebAuthnCeremony(context.Background(), record, 2*time.Minute); err != nil {
				t.Fatalf("create ceremony: %v", err)
			}
			if executor.setKey != test.wantKey {
				t.Fatalf("key = %q, want %q", executor.setKey, test.wantKey)
			}
			if executor.setTTL != 2*time.Minute {
				t.Fatalf("ttl = %s, want 2m", executor.setTTL)
			}
			if len(executor.setValue) > maxWebAuthnCeremonyBytes {
				t.Fatalf("serialized ceremony = %d bytes, want <= %d", len(executor.setValue), maxWebAuthnCeremonyBytes)
			}

			consumed, err := store.ConsumeWebAuthnCeremony(context.Background(), test.operation, test.ceremony)
			if err != nil {
				t.Fatalf("consume ceremony: %v", err)
			}
			if consumed.Session.Challenge != record.Session.Challenge {
				t.Fatalf("challenge = %q, want %q", consumed.Session.Challenge, record.Session.Challenge)
			}
			if executor.runScript != redisGetDeleteScript || executor.runCalls != 1 {
				t.Fatalf("consume used script %q in %d calls", executor.runScript, executor.runCalls)
			}

			_, err = store.ConsumeWebAuthnCeremony(context.Background(), test.operation, test.ceremony)
			if !errors.Is(err, ErrWebAuthnCeremonyNotFound) {
				t.Fatalf("second consume error = %v, want not found", err)
			}
		})
	}
}

func TestWebAuthnCeremonyStorePreservesTask10AccountRotationWire(t *testing.T) {
	t.Parallel()

	executor := &fakeChallengeExecutor{setAllowed: true}
	store, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	record := task10ChallengeRecord()
	wantWire, err := encodeChallengeRecord(record)
	if err != nil {
		t.Fatalf("encode Task 10 challenge: %v", err)
	}
	if err := store.Create(context.Background(), record, 2*time.Minute); err != nil {
		t.Fatalf("create Task 10 challenge: %v", err)
	}
	if executor.setKey != "talenro:identity:account-challenge:"+record.ChallengeID {
		t.Fatalf("Task 10 key changed: %q", executor.setKey)
	}
	if !bytes.Equal(executor.setValue, wantWire) {
		t.Fatal("Task 10 challenge wire changed")
	}
}

func TestWebAuthnCeremonyStoreSanitizesBackendErrors(t *testing.T) {
	t.Parallel()

	executor := &fakeChallengeExecutor{runErrors: []error{errors.New("CANARY backend detail")}}
	store, err := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	_, err = store.ConsumeWebAuthnCeremony(
		context.Background(),
		WebAuthnRegistrationCeremony,
		"2dddc425-1bd2-4c33-9b0f-658082addbae",
	)
	if !errors.Is(err, ErrWebAuthnCeremonyUnavailable) {
		t.Fatalf("consume error = %v, want unavailable", err)
	}
	if bytes.Contains([]byte(err.Error()), []byte("CANARY")) {
		t.Fatalf("backend detail leaked: %v", err)
	}
}

func TestWebAuthnProviderBoundarySanitizesPanics(t *testing.T) {
	t.Parallel()

	value, err := safeWebAuthnProviderCall(func() (int, error) {
		panic("TASK11_WEBAUTHN_PROVIDER_PANIC_CANARY")
	})
	if value != 0 || !errors.Is(err, errStrongAuthDependency) || strings.Contains(fmt.Sprint(err), "CANARY") {
		t.Fatalf("single-result provider panic escaped finite mapping: %d/%v", value, err)
	}
	first, second, err := safeWebAuthnProviderCall2(func() (int, string, error) {
		panic("TASK11_WEBAUTHN_PROVIDER_PANIC_CANARY")
	})
	if first != 0 || second != "" || !errors.Is(err, errStrongAuthDependency) || strings.Contains(fmt.Sprint(err), "CANARY") {
		t.Fatalf("two-result provider panic escaped finite mapping: %d/%q/%v", first, second, err)
	}
}

func TestWebAuthnCeremonyStoreSanitizesProviderPanics(t *testing.T) {
	t.Parallel()

	record := WebAuthnCeremonyRecord{
		CeremonyID: "2dddc425-1bd2-4c33-9b0f-658082addbae",
		Operation:  WebAuthnRegistrationCeremony,
		Session: wa.SessionData{
			Challenge: "opaque-provider-panic-challenge", RelyingPartyID: "example.com",
			UserID: []byte{1, 2, 3, 4}, Expires: time.Now().UTC().Add(webAuthnCeremonyTTL),
			UserVerification: protocol.VerificationRequired,
		},
	}
	record.ExpiresAt = record.Session.Expires
	for _, operation := range []string{"set", "run"} {
		operation := operation
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			store, err := newRedisChallengeStoreWithExecutor(
				panickingTask11ChallengeExecutor{operation: operation}, 250*time.Millisecond,
			)
			if err != nil {
				t.Fatalf("new store: %v", err)
			}
			var providerErr error
			panicked := captureTask11ProviderPanic(func() {
				if operation == "set" {
					providerErr = store.CreateWebAuthnCeremony(context.Background(), record, webAuthnCeremonyTTL)
					return
				}
				_, providerErr = store.ConsumeWebAuthnCeremony(context.Background(), record.Operation, record.CeremonyID)
			})
			if panicked || !errors.Is(providerErr, ErrWebAuthnCeremonyUnavailable) || strings.Contains(fmt.Sprint(providerErr), "CANARY") {
				t.Fatalf("Redis %s panic escaped finite mapping: panicked=%t err=%v", operation, panicked, providerErr)
			}
		})
	}
}

type panickingTask11ChallengeExecutor struct{ operation string }

func (executor panickingTask11ChallengeExecutor) SetNX(context.Context, string, []byte, time.Duration) (bool, error) {
	if executor.operation == "set" {
		panic("TASK11_REDIS_PROVIDER_PANIC_CANARY")
	}
	return true, nil
}

func (executor panickingTask11ChallengeExecutor) Run(context.Context, string, []string, ...any) (any, error) {
	if executor.operation == "run" {
		panic("TASK11_REDIS_PROVIDER_PANIC_CANARY")
	}
	return nil, nil
}

func captureTask11ProviderPanic(call func()) (panicked bool) {
	defer func() {
		panicked = recover() != nil
	}()
	call()
	return false
}

func TestStrongAuthValuesRedactPrivateMaterial(t *testing.T) {
	t.Parallel()

	const textCanary = "TASK11_PRIVATE_TEXT_CANARY"
	byteCanary := []byte("TASK11_PRIVATE_BYTES_CANARY")
	principalID := PrincipalID("28ceee8a-5f4f-4d3a-9e9f-d3ec23b815ac")
	sessionID := SessionID("ac9150c8-aaaf-468b-a9fe-605d8c831d1a")
	reauthentication := Reauthentication{
		SessionID: sessionID, Method: ReauthPassword, Proof: secret.NewBytes(byteCanary),
	}
	values := []any{
		BeginPasskeyRegistrationCommand{PrincipalID: principalID, DisplayName: textCanary, Reauthentication: reauthentication},
		FinishPasskeyRegistrationCommand{PrincipalID: principalID, CeremonyID: textCanary, Response: json.RawMessage(`{"canary":"TASK11_PRIVATE_TEXT_CANARY"}`), Reauthentication: reauthentication},
		FinishPasskeyAuthenticationCommand{CeremonyID: textCanary, Response: json.RawMessage(`{"canary":"TASK11_PRIVATE_TEXT_CANARY"}`)},
		RevokePasskeyCommand{PrincipalID: principalID, CredentialID: byteCanary, Reauthentication: reauthentication},
		BeginTOTPEnrollmentCommand{PrincipalID: principalID, Reauthentication: reauthentication},
		TOTPEnrollment{Secret: textCanary, URI: "otpauth://totp/" + textCanary},
		VerifyTOTPEnrollmentCommand{PrincipalID: principalID, Code: secret.NewBytes(byteCanary), Reauthentication: reauthentication},
		RevokeTOTPCommand{PrincipalID: principalID, Reauthentication: reauthentication},
		RotateRecoveryCodesCommand{PrincipalID: principalID, Reauthentication: reauthentication},
		RecoveryCodes{Codes: []string{textCanary}},
		ConsumeRecoveryCodeCommand{Email: textCanary, Code: secret.NewBytes(byteCanary)},
		WebAuthnCeremonyRecord{CeremonyID: textCanary, Session: wa.SessionData{Challenge: textCanary}},
	}
	for _, value := range values {
		rendered := fmt.Sprintf("%+v", value) + slog.Any("value", value).Value.String()
		if strings.Contains(rendered, textCanary) || strings.Contains(rendered, string(byteCanary)) {
			t.Fatalf("private material escaped redaction for %T", value)
		}
		body, err := json.Marshal(value)
		if err == nil {
			t.Fatalf("JSON serialization unexpectedly succeeded for %T: %s", value, body)
		}
		if bytes.Contains(body, []byte(textCanary)) || bytes.Contains(body, byteCanary) {
			t.Fatalf("private material escaped failed JSON serialization for %T", value)
		}
		clear(body)
	}
}

func (transaction *task11TOTPTransaction) ListActivePasskeys(
	context.Context,
	uuid.UUID,
) ([]store.IdentityPasskeyCredential, error) {
	if err := transaction.record("list_passkeys"); err != nil {
		return nil, err
	}
	return append([]store.IdentityPasskeyCredential(nil), transaction.passkeys...), nil
}

func (transaction *task11TOTPTransaction) CreatePasskeyCredential(
	_ context.Context,
	params store.CreatePasskeyCredentialParams,
) error {
	transaction.createdPasskey = params
	transaction.createdPasskey.CredentialID = append([]byte(nil), params.CredentialID...)
	transaction.createdPasskey.PublicKey = append([]byte(nil), params.PublicKey...)
	transaction.createdPasskey.Transports = append([]string(nil), params.Transports...)
	return transaction.record("create_passkey")
}

func (transaction *task11TOTPTransaction) UpdatePasskeyCounter(
	_ context.Context,
	params store.UpdatePasskeyCounterParams,
) (int64, error) {
	transaction.updatedPasskey = params
	return transaction.updatePasskeyRows, transaction.record("update_passkey")
}

func (transaction *task11TOTPTransaction) RevokePasskey(
	_ context.Context,
	params store.RevokePasskeyParams,
) (int64, error) {
	transaction.revokedPasskey = params
	return transaction.revokePasskeyRows, transaction.record("revoke_passkey")
}

func task11Passkeys(principalID uuid.UUID, count int) []store.IdentityPasskeyCredential {
	passkeys := make([]store.IdentityPasskeyCredential, count)
	for index := range passkeys {
		passkeys[index] = store.IdentityPasskeyCredential{
			CredentialID: bytes.Repeat([]byte{byte(index + 1)}, 16), PrincipalID: principalID,
			PublicKey: bytes.Repeat([]byte{byte(index + 11)}, 32), AttestationFormat: "none",
			ProtocolFlags: int16(protocol.FlagUserPresent | protocol.FlagUserVerified), State: "active",
			CreatedAt: fixedTask10Time.Add(time.Duration(index) * time.Second),
		}
	}
	return passkeys
}

func newTask11WebAuthnApplication(
	t *testing.T,
	transaction *task11TOTPTransaction,
	ceremonies ChallengeStore,
) *Service {
	t.Helper()
	protector, err := sensitive.NewLocal(
		secret.NewBytes(make([]byte, 32)), secret.NewBytes(make([]byte, 32)), 7,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = protector.Close() })
	bindTask11Idempotency(t, transaction, protector)
	application, err := newApplicationForTest(ApplicationDependencies{
		Repository: &task11TOTPRepository{tx: transaction}, Protector: protector,
		Random: &task9Random{}, Clock: task10Clock{}, Limiter: &task10Limiter{allowed: true},
		ChallengeStore: ceremonies, RateLimitKey: secret.NewBytes(make([]byte, 32)),
		Security: config.SecurityConfig{
			Profile: config.ProfileTest, EmailVerification: config.EmailDisabled,
			PublicBaseURL: "https://api.example.test", RequestDeadline: 5 * time.Second,
			RedisTimeout: 250 * time.Millisecond, WebAuthnRPID: "login.example.test",
			WebAuthnOrigins:    []string{"https://login.example.test"},
			LoginRateLimit:     config.RateLimitPolicy{Limit: 10, Window: 15 * time.Minute},
			DeliveryRateLimit:  config.RateLimitPolicy{Limit: 5, Window: time.Hour},
			ChallengeRateLimit: config.RateLimitPolicy{Limit: 20, Window: 5 * time.Minute},
		},
		DeviceAuthorizationParticipant: &task9Participant{owner: transaction.fakeIdentityTransaction},
	}, (&task10Deriver{}).Derive, uuid.New)
	if err != nil {
		t.Fatal(err)
	}
	return application
}

func task11RegistrationResponse(
	t *testing.T,
	challenge string,
	origin string,
	credentialID []byte,
) json.RawMessage {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	clientData, err := json.Marshal(map[string]any{
		"type": "webauthn.create", "challenge": challenge, "origin": origin, "crossOrigin": false,
	})
	if err != nil {
		t.Fatal(err)
	}
	publicKey := task11COSEPublicKey(t, &privateKey.PublicKey)
	rpIDHash := sha256.Sum256([]byte("login.example.test"))
	authenticatorData := make([]byte, 0, 32+1+4+16+2+len(credentialID)+len(publicKey))
	authenticatorData = append(authenticatorData, rpIDHash[:]...)
	authenticatorData = append(authenticatorData, byte(protocol.FlagUserPresent|protocol.FlagUserVerified|protocol.FlagAttestedCredentialData))
	authenticatorData = binary.BigEndian.AppendUint32(authenticatorData, 0)
	authenticatorData = append(authenticatorData, make([]byte, 16)...)
	// #nosec G115 -- fixture credential IDs are bounded by the production 1024-byte maximum.
	authenticatorData = binary.BigEndian.AppendUint16(authenticatorData, uint16(len(credentialID)))
	authenticatorData = append(authenticatorData, credentialID...)
	authenticatorData = append(authenticatorData, publicKey...)
	attestationObject, err := cbor.Marshal(map[string]any{
		"fmt": "none", "attStmt": map[string]any{}, "authData": authenticatorData,
	})
	if err != nil {
		t.Fatal(err)
	}
	encodedID := base64.RawURLEncoding.EncodeToString(credentialID)
	response, err := json.Marshal(map[string]any{
		"id": encodedID, "rawId": encodedID, "type": "public-key",
		"authenticatorAttachment": "platform", "clientExtensionResults": map[string]any{},
		"response": map[string]any{
			"clientDataJSON":    base64.RawURLEncoding.EncodeToString(clientData),
			"attestationObject": base64.RawURLEncoding.EncodeToString(attestationObject),
			"transports":        []string{"internal"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func task11PasskeyKey(t *testing.T) (*ecdsa.PrivateKey, []byte) {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	publicKey := task11COSEPublicKey(t, &privateKey.PublicKey)
	return privateKey, publicKey
}

func task11COSEPublicKey(t *testing.T, publicKey *ecdsa.PublicKey) []byte {
	t.Helper()
	encoded, err := publicKey.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) != 65 || encoded[0] != 4 {
		t.Fatalf("unexpected P-256 public key encoding length/prefix: %d/%d", len(encoded), encoded[0])
	}
	xCoordinate := bytes.Clone(encoded[1:33])
	yCoordinate := bytes.Clone(encoded[33:65])
	defer clear(xCoordinate)
	defer clear(yCoordinate)
	result, err := cbor.Marshal(map[int]any{
		1: 2, 3: -7, -1: 1,
		-2: xCoordinate,
		-3: yCoordinate,
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func task11AuthenticationResponse(
	t *testing.T,
	challenge string,
	origin string,
	principalID uuid.UUID,
	credentialID []byte,
	privateKey *ecdsa.PrivateKey,
	counter uint32,
) json.RawMessage {
	return task11AuthenticationResponseWithPolicy(
		t, challenge, origin, principalID, credentialID, privateKey, counter, false, true,
	)
}

func task11AuthenticationResponseWithPolicy(
	t *testing.T,
	challenge string,
	origin string,
	principalID uuid.UUID,
	credentialID []byte,
	privateKey *ecdsa.PrivateKey,
	counter uint32,
	crossOrigin bool,
	userVerified bool,
) json.RawMessage {
	t.Helper()
	clientData, err := json.Marshal(map[string]any{
		"type": "webauthn.get", "challenge": challenge, "origin": origin, "crossOrigin": crossOrigin,
	})
	if err != nil {
		t.Fatal(err)
	}
	rpIDHash := sha256.Sum256([]byte("login.example.test"))
	authenticatorData := make([]byte, 0, 37)
	authenticatorData = append(authenticatorData, rpIDHash[:]...)
	flags := protocol.FlagUserPresent
	if userVerified {
		flags |= protocol.FlagUserVerified
	}
	authenticatorData = append(authenticatorData, byte(flags))
	authenticatorData = binary.BigEndian.AppendUint32(authenticatorData, counter)
	clientDataHash := sha256.Sum256(clientData)
	signed := make([]byte, 0, len(authenticatorData)+len(clientDataHash))
	signed = append(signed, authenticatorData...)
	signed = append(signed, clientDataHash[:]...)
	signedHash := sha256.Sum256(signed)
	clear(signed)
	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, signedHash[:])
	clear(signedHash[:])
	if err != nil {
		t.Fatal(err)
	}
	encodedID := base64.RawURLEncoding.EncodeToString(credentialID)
	response, err := json.Marshal(map[string]any{
		"id": encodedID, "rawId": encodedID, "type": "public-key", "clientExtensionResults": map[string]any{},
		"response": map[string]any{
			"clientDataJSON":    base64.RawURLEncoding.EncodeToString(clientData),
			"authenticatorData": base64.RawURLEncoding.EncodeToString(authenticatorData),
			"signature":         base64.RawURLEncoding.EncodeToString(signature),
			"userHandle":        base64.RawURLEncoding.EncodeToString(principalID[:]),
		},
	})
	clear(signature)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func task11DifferentChallenge(value string) string {
	if value == "" {
		return "A"
	}
	replacement := byte('A')
	if value[0] == replacement {
		replacement = 'B'
	}
	return string(replacement) + value[1:]
}
