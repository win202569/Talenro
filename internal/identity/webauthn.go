package identity

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"
	"unicode/utf8"

	"github.com/go-webauthn/webauthn/protocol"
	wa "github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"

	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/store"
)

const (
	// WebAuthnRegistrationCeremony stores registration-only protocol state.
	WebAuthnRegistrationCeremony WebAuthnCeremonyOperation = "registration"
	// WebAuthnAuthenticationCeremony stores authentication-only protocol state.
	WebAuthnAuthenticationCeremony WebAuthnCeremonyOperation = "authentication"

	webAuthnCeremonyTTL          = 2 * time.Minute
	maxWebAuthnCeremonyBytes     = 64 << 10
	webAuthnCeremonyRecordPrefix = "TALENRO-WEBAUTHN-CEREMONY-V1\x00"
	webAuthnDisplayNameBytes     = 256
	webAuthnDisplayNameRunes     = 64
	maximumActivePasskeys        = 10
)

var (
	// ErrInvalidWebAuthnCeremony classifies malformed ceremony state or configuration.
	ErrInvalidWebAuthnCeremony = errors.New("identity: invalid WebAuthn ceremony")
	// ErrWebAuthnCeremonyNotFound classifies absent, expired, or consumed ceremony state.
	ErrWebAuthnCeremonyNotFound = errors.New("identity: WebAuthn ceremony unavailable")
	// ErrWebAuthnCeremonyUnavailable classifies an ambiguous ceremony-store failure.
	ErrWebAuthnCeremonyUnavailable = errors.New("identity: WebAuthn ceremony store unavailable")
)

func safeWebAuthnProviderCall[Result any](call func() (Result, error)) (result Result, err error) {
	if call == nil {
		return result, errStrongAuthDependency
	}
	defer func() {
		if recover() != nil {
			var zero Result
			result = zero
			err = errStrongAuthDependency
		}
	}()
	return call()
}

func safeWebAuthnProviderCall2[First, Second any](call func() (First, Second, error)) (first First, second Second, err error) {
	if call == nil {
		return first, second, errStrongAuthDependency
	}
	defer func() {
		if recover() != nil {
			var zeroFirst First
			var zeroSecond Second
			first = zeroFirst
			second = zeroSecond
			err = errStrongAuthDependency
		}
	}()
	return call()
}

// WebAuthnCeremonyOperation identifies one isolated ceremony Redis domain.
type WebAuthnCeremonyOperation string

// WebAuthnCeremonyRecord contains bounded go-webauthn protocol state for one use.
type WebAuthnCeremonyRecord struct {
	CeremonyID string
	Operation  WebAuthnCeremonyOperation
	Session    wa.SessionData
	ExpiresAt  time.Time
}

// WebAuthnCeremonyStore creates and atomically consumes bounded ceremony state.
type WebAuthnCeremonyStore interface {
	CreateWebAuthnCeremony(context.Context, WebAuthnCeremonyRecord, time.Duration) error
	ConsumeWebAuthnCeremony(context.Context, WebAuthnCeremonyOperation, string) (WebAuthnCeremonyRecord, error)
}

// StrongAuthApplication is the exact account-factor surface consumed by the control API.
type StrongAuthApplication interface {
	BeginPasskeyRegistration(context.Context, BeginPasskeyRegistrationCommand) (json.RawMessage, error)
	FinishPasskeyRegistration(context.Context, FinishPasskeyRegistrationCommand) error
	BeginPasskeyAuthentication(context.Context, BeginPasskeyAuthenticationCommand) (json.RawMessage, error)
	FinishPasskeyAuthentication(context.Context, FinishPasskeyAuthenticationCommand) (SessionTokens, error)
	RevokePasskey(context.Context, RevokePasskeyCommand) error
	BeginTOTPEnrollment(context.Context, BeginTOTPEnrollmentCommand) (TOTPEnrollment, error)
	VerifyTOTPEnrollment(context.Context, VerifyTOTPEnrollmentCommand) error
	RevokeTOTP(context.Context, RevokeTOTPCommand) error
	RotateRecoveryCodes(context.Context, RotateRecoveryCodesCommand) (RecoveryCodes, error)
	ConsumeRecoveryCode(context.Context, ConsumeRecoveryCodeCommand) (SessionTokens, error)
}

var _ StrongAuthApplication = (*Service)(nil)

// BeginPasskeyRegistrationCommand starts an authenticated discoverable-credential ceremony.
type BeginPasskeyRegistrationCommand struct {
	PrincipalID      PrincipalID
	DisplayName      string
	Reauthentication Reauthentication
	IdempotencyKey   string
}

// FinishPasskeyRegistrationCommand finishes and persists one library-verified credential.
type FinishPasskeyRegistrationCommand struct {
	PrincipalID      PrincipalID
	CeremonyID       string
	Response         json.RawMessage
	Reauthentication Reauthentication
	IdempotencyKey   string
}

// BeginPasskeyAuthenticationCommand has no caller-controlled fields.
type BeginPasskeyAuthenticationCommand struct{}

// FinishPasskeyAuthenticationCommand finishes a discoverable assertion and creates a bound session.
type FinishPasskeyAuthenticationCommand struct {
	CeremonyID             string
	Response               json.RawMessage
	ClientSigningPublicKey [32]byte
	IdempotencyKey         string
}

// RevokePasskeyCommand independently revokes one credential.
type RevokePasskeyCommand struct {
	PrincipalID      PrincipalID
	CredentialID     []byte
	Reauthentication Reauthentication
	IdempotencyKey   string
}

type passkeyTransaction interface {
	Transaction
	ListActivePasskeys(context.Context, uuid.UUID) ([]store.IdentityPasskeyCredential, error)
	CreatePasskeyCredential(context.Context, store.CreatePasskeyCredentialParams) error
	UpdatePasskeyCounter(context.Context, store.UpdatePasskeyCounterParams) (int64, error)
	RevokePasskey(context.Context, store.RevokePasskeyParams) (int64, error)
}

type passkeyUser struct {
	id          []byte
	name        string
	displayName string
	credentials []wa.Credential
}

func (user passkeyUser) WebAuthnID() []byte          { return append([]byte(nil), user.id...) }
func (user passkeyUser) WebAuthnName() string        { return user.name }
func (user passkeyUser) WebAuthnDisplayName() string { return user.displayName }
func (user passkeyUser) WebAuthnCredentials() []wa.Credential {
	return append([]wa.Credential(nil), user.credentials...)
}

// BeginPasskeyRegistration creates a two-minute single-use ceremony outside every database transaction.
func (service *Service) BeginPasskeyRegistration(
	ctx context.Context,
	command BeginPasskeyRegistrationCommand,
) (json.RawMessage, error) {
	principalID, principalOK := parseCanonicalIdentityUUID(string(command.PrincipalID))
	sessionID, sessionOK := parseCanonicalIdentityUUID(string(command.Reauthentication.SessionID))
	ceremonies, ceremonyOK := webAuthnCeremonies(service)
	if !validService(service) || nilIdentityValue(ctx) || !principalOK || !sessionOK || !ceremonyOK ||
		!validWebAuthnDisplayName(command.DisplayName) || !validIdempotencyKeyForApplication(command.IdempotencyKey) {
		return nil, malformedRequest()
	}
	if command.Reauthentication.Method != ReauthPassword {
		return nil, actionNotAllowed()
	}
	if !validPasswordSecret(command.Reauthentication.Proof) {
		return nil, malformedRequest()
	}
	proof := command.Reauthentication.Proof.Copy()
	defer clear(proof)
	canonical, err := strongAuthCanonicalRequest(
		service.protector, "begin_passkey_registration", principalID[:], sessionID[:], []byte(command.DisplayName), proof,
	)
	if err != nil {
		return nil, dependencyUnavailable()
	}
	defer clear(canonical)
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	user, replayedOptions, replayed, err := service.preflightPasskeyRegistrationBegin(
		operationContext, command, principalID, canonical,
	)
	if err != nil {
		return nil, mapApplicationError(operationContext, err)
	}
	if replayed {
		return replayedOptions, nil
	}
	user.displayName = command.DisplayName
	webAuthn, err := service.newWebAuthn()
	if err != nil {
		return nil, dependencyUnavailable()
	}
	creation, session, err := safeWebAuthnProviderCall2(func() (*protocol.CredentialCreation, *wa.SessionData, error) {
		return webAuthn.BeginRegistration(
			user,
			wa.WithConveyancePreference(protocol.PreferNoAttestation),
			wa.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
				ResidentKey: protocol.ResidentKeyRequirementPreferred, UserVerification: protocol.VerificationRequired,
			}),
		)
	})
	if err != nil || creation == nil || session == nil || session.Expires.IsZero() {
		return nil, dependencyUnavailable()
	}
	ceremonyID := service.newUUID()
	if ceremonyID == uuid.Nil {
		return nil, dependencyUnavailable()
	}
	record := WebAuthnCeremonyRecord{
		CeremonyID: ceremonyID.String(), Operation: WebAuthnRegistrationCeremony,
		Session: *session, ExpiresAt: session.Expires,
	}
	if err = ceremonies.CreateWebAuthnCeremony(operationContext, record, webAuthnCeremonyTTL); err != nil {
		return nil, dependencyUnavailable()
	}
	options, err := encodePasskeyRegistrationOptions(record.CeremonyID, creation)
	if err != nil {
		return nil, dependencyUnavailable()
	}
	result, err := service.completePasskeyRegistrationOptions(
		operationContext, command, principalID, canonical, options,
	)
	clear(options)
	if err != nil {
		return nil, mapApplicationError(operationContext, err)
	}
	return result, nil
}

// BeginPasskeyAuthentication creates a public discoverable-login ceremony in the authentication Redis domain.
func (service *Service) BeginPasskeyAuthentication(
	ctx context.Context,
	_ BeginPasskeyAuthenticationCommand,
) (json.RawMessage, error) {
	ceremonies, ceremonyOK := webAuthnCeremonies(service)
	if !validService(service) || nilIdentityValue(ctx) || !ceremonyOK {
		return nil, malformedRequest()
	}
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	webAuthn, err := service.newWebAuthn()
	if err != nil {
		return nil, dependencyUnavailable()
	}
	assertion, session, err := safeWebAuthnProviderCall2(func() (*protocol.CredentialAssertion, *wa.SessionData, error) {
		return webAuthn.BeginDiscoverableLogin(wa.WithUserVerification(protocol.VerificationRequired))
	})
	if err != nil || assertion == nil || session == nil || session.Expires.IsZero() {
		return nil, dependencyUnavailable()
	}
	ceremonyID := service.newUUID()
	if ceremonyID == uuid.Nil {
		return nil, dependencyUnavailable()
	}
	record := WebAuthnCeremonyRecord{
		CeremonyID: ceremonyID.String(), Operation: WebAuthnAuthenticationCeremony,
		Session: *session, ExpiresAt: session.Expires,
	}
	if err = ceremonies.CreateWebAuthnCeremony(operationContext, record, webAuthnCeremonyTTL); err != nil {
		return nil, dependencyUnavailable()
	}
	options, err := encodePasskeyAuthenticationOptions(record.CeremonyID, assertion)
	if err != nil {
		return nil, dependencyUnavailable()
	}
	return options, nil
}

// FinishPasskeyRegistration consumes one ceremony, verifies it outside PostgreSQL, then persists the credential.
func (service *Service) FinishPasskeyRegistration(
	ctx context.Context,
	command FinishPasskeyRegistrationCommand,
) error {
	principalID, principalOK := parseCanonicalIdentityUUID(string(command.PrincipalID))
	sessionID, sessionOK := parseCanonicalIdentityUUID(string(command.Reauthentication.SessionID))
	_, ceremonyIDOK := parseCanonicalIdentityUUID(command.CeremonyID)
	response := append([]byte(nil), command.Response...)
	defer clear(response)
	ceremonies, ceremonyOK := webAuthnCeremonies(service)
	if !validService(service) || nilIdentityValue(ctx) || !principalOK || !sessionOK || !ceremonyIDOK || !ceremonyOK ||
		len(response) == 0 || len(response) > maxWebAuthnCeremonyBytes || !json.Valid(response) ||
		!validIdempotencyKeyForApplication(command.IdempotencyKey) {
		return malformedRequest()
	}
	if command.Reauthentication.Method != ReauthPassword {
		return actionNotAllowed()
	}
	if !validPasswordSecret(command.Reauthentication.Proof) {
		return malformedRequest()
	}
	proof := command.Reauthentication.Proof.Copy()
	defer clear(proof)
	canonical, err := strongAuthCanonicalRequest(
		service.protector, "finish_passkey_registration", principalID[:], sessionID[:], []byte(command.CeremonyID), response, proof,
	)
	if err != nil {
		return dependencyUnavailable()
	}
	defer clear(canonical)
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	user, replayed, err := service.preflightPasskeyRegistrationFinish(
		operationContext, command, principalID, canonical,
	)
	if err != nil {
		return mapApplicationError(operationContext, err)
	}
	if replayed {
		return nil
	}
	record, err := ceremonies.ConsumeWebAuthnCeremony(
		operationContext, WebAuthnRegistrationCeremony, command.CeremonyID,
	)
	if err != nil {
		if errors.Is(err, ErrWebAuthnCeremonyNotFound) {
			return authenticationFailed()
		}
		return dependencyUnavailable()
	}
	if !bytes.Equal(record.Session.UserID, principalID[:]) || record.Session.RelyingPartyID != service.security.WebAuthnRPID ||
		record.Session.UserVerification != protocol.VerificationRequired {
		return authenticationFailed()
	}
	parsed, err := safeWebAuthnProviderCall(func() (*protocol.ParsedCredentialCreationData, error) {
		return protocol.ParseCredentialCreationResponseBytes(response)
	})
	if err != nil || parsed == nil {
		return authenticationFailed()
	}
	webAuthn, err := service.newWebAuthn()
	if err != nil {
		return dependencyUnavailable()
	}
	credential, err := safeWebAuthnProviderCall(func() (*wa.Credential, error) {
		return webAuthn.CreateCredential(user, record.Session, parsed)
	})
	if err != nil || !validVerifiedPasskey(credential) {
		return authenticationFailed()
	}
	if err = service.persistPasskeyRegistration(
		operationContext, command, principalID, canonical, credential,
	); err != nil {
		return mapApplicationError(operationContext, err)
	}
	return nil
}

// FinishPasskeyAuthentication verifies one discoverable assertion outside PostgreSQL and issues a session atomically.
func (service *Service) FinishPasskeyAuthentication(
	ctx context.Context,
	command FinishPasskeyAuthenticationCommand,
) (SessionTokens, error) {
	_, ceremonyIDOK := parseCanonicalIdentityUUID(command.CeremonyID)
	response := append([]byte(nil), command.Response...)
	defer clear(response)
	ceremonies, ceremonyOK := webAuthnCeremonies(service)
	if !validService(service) || nilIdentityValue(ctx) || !ceremonyIDOK || !ceremonyOK ||
		len(response) == 0 || len(response) > maxWebAuthnCeremonyBytes || !json.Valid(response) ||
		command.ClientSigningPublicKey == [32]byte{} || !validIdempotencyKeyForApplication(command.IdempotencyKey) {
		return SessionTokens{}, malformedRequest()
	}
	parsed, err := safeWebAuthnProviderCall(func() (*protocol.ParsedCredentialAssertionData, error) {
		return protocol.ParseCredentialRequestResponseBytes(response)
	})
	if err != nil || parsed == nil || len(parsed.Response.UserHandle) != len(uuid.UUID{}) ||
		len(parsed.RawID) < 16 || len(parsed.RawID) > 1024 {
		return SessionTokens{}, authenticationFailed()
	}
	principalID, err := uuid.FromBytes(parsed.Response.UserHandle)
	if err != nil || principalID == uuid.Nil {
		return SessionTokens{}, authenticationFailed()
	}
	canonical, err := strongAuthCanonicalRequest(
		service.protector, "finish_passkey_authentication", principalID[:], []byte(command.CeremonyID), response,
		command.ClientSigningPublicKey[:],
	)
	if err != nil {
		return SessionTokens{}, dependencyUnavailable()
	}
	defer clear(canonical)
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	user, replayedTokens, replayed, preloadErr := service.preflightPasskeyAuthentication(
		operationContext, command, principalID, parsed.RawID, canonical,
	)
	if preloadErr != nil {
		return SessionTokens{}, mapApplicationError(operationContext, preloadErr)
	}
	if replayed {
		return replayedTokens, nil
	}
	record, err := ceremonies.ConsumeWebAuthnCeremony(
		operationContext, WebAuthnAuthenticationCeremony, command.CeremonyID,
	)
	if err != nil {
		if errors.Is(err, ErrWebAuthnCeremonyNotFound) {
			return SessionTokens{}, authenticationFailed()
		}
		return SessionTokens{}, dependencyUnavailable()
	}
	if len(record.Session.UserID) != 0 || record.Session.RelyingPartyID != service.security.WebAuthnRPID ||
		record.Session.UserVerification != protocol.VerificationRequired {
		return SessionTokens{}, authenticationFailed()
	}
	webAuthn, err := service.newWebAuthn()
	if err != nil {
		return SessionTokens{}, dependencyUnavailable()
	}
	handler := func(rawID, userHandle []byte) (wa.User, error) {
		if !bytes.Equal(rawID, parsed.RawID) || !bytes.Equal(userHandle, principalID[:]) {
			return nil, errStrongAuthDependency
		}
		return user, nil
	}
	verifiedUser, verifiedCredential, err := safeWebAuthnProviderCall2(func() (wa.User, *wa.Credential, error) {
		return webAuthn.ValidatePasskeyLogin(handler, record.Session, parsed)
	})
	if err != nil || verifiedUser == nil || verifiedCredential == nil || !bytes.Equal(verifiedUser.WebAuthnID(), principalID[:]) ||
		!bytes.Equal(verifiedCredential.ID, parsed.RawID) || !validVerifiedPasskey(verifiedCredential) {
		return SessionTokens{}, authenticationFailed()
	}
	result, err := service.completePasskeyAuthentication(
		operationContext, command, principalID, canonical, verifiedCredential,
	)
	if err != nil {
		return SessionTokens{}, mapApplicationError(operationContext, err)
	}
	return result, nil
}

// RevokePasskey independently revokes one active credential after password reauthentication.
func (service *Service) RevokePasskey(ctx context.Context, command RevokePasskeyCommand) error {
	principalID, principalOK := parseCanonicalIdentityUUID(string(command.PrincipalID))
	sessionID, sessionOK := parseCanonicalIdentityUUID(string(command.Reauthentication.SessionID))
	credentialID := append([]byte(nil), command.CredentialID...)
	defer clear(credentialID)
	if !validService(service) || nilIdentityValue(ctx) || !principalOK || !sessionOK ||
		len(credentialID) < 16 || len(credentialID) > 1024 || !validIdempotencyKeyForApplication(command.IdempotencyKey) {
		return malformedRequest()
	}
	if command.Reauthentication.Method != ReauthPassword {
		return actionNotAllowed()
	}
	if !validPasswordSecret(command.Reauthentication.Proof) {
		return malformedRequest()
	}
	proof := command.Reauthentication.Proof.Copy()
	defer clear(proof)
	canonical, err := strongAuthCanonicalRequest(
		service.protector, "revoke_passkey", principalID[:], sessionID[:], credentialID, proof,
	)
	if err != nil {
		return dependencyUnavailable()
	}
	defer clear(canonical)
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	err = service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, base Transaction) error {
		transaction, ok := base.(passkeyTransaction)
		if !ok || nilIdentityValue(transaction) {
			return dependencyUnavailable()
		}
		now, ok := service.now()
		if !ok {
			return dependencyUnavailable()
		}
		passwordCredential, reauthErr := service.loadPasswordReauthenticationCredential(transactionContext, transaction, principalID)
		if reauthErr != nil {
			return reauthErr
		}
		rows, listErr := transaction.ListActivePasskeys(transactionContext, principalID)
		if listErr != nil {
			return dependencyUnavailable()
		}
		if _, reauthErr = service.verifyLockedPasswordReauthentication(
			transactionContext, transaction, principalID, command.Reauthentication, passwordCredential, now,
		); reauthErr != nil {
			return reauthErr
		}
		scope, scopeErr := idempotency.AuthenticatedScope(principalID, "passkey", "revoke_passkey")
		if scopeErr != nil {
			return dependencyUnavailable()
		}
		record, outcome, beginErr := transaction.BeginIdempotency(
			transactionContext, scope, command.IdempotencyKey, canonical, now, now.Add(securityIdempotencyRetention),
		)
		if beginErr != nil {
			return dependencyUnavailable()
		}
		if outcome != idempotency.Started {
			replayed, replayErr := genericIdempotencyOutcome(outcome, record, 204)
			if replayErr != nil || replayed {
				return replayErr
			}
		}
		if !containsPasskeyCredential(rows, credentialID) {
			return strongAuthStateConflict()
		}
		rowsAffected, revokeErr := transaction.RevokePasskey(transactionContext, store.RevokePasskeyParams{
			CredentialID: append([]byte(nil), credentialID...), PrincipalID: principalID,
			RevokedAt: sql.NullTime{Time: now, Valid: true},
		})
		if revokeErr != nil {
			return dependencyUnavailable()
		}
		if rowsAffected != 1 {
			return strongAuthStateConflict()
		}
		if completeErr := transaction.CompleteIdempotency(transactionContext, record, 204, nil); completeErr != nil {
			return dependencyUnavailable()
		}
		return nil
	})
	if err != nil {
		return mapApplicationError(operationContext, err)
	}
	return nil
}

func (service *Service) preflightPasskeyAuthentication(
	ctx context.Context,
	command FinishPasskeyAuthenticationCommand,
	principalID uuid.UUID,
	credentialID []byte,
	canonical []byte,
) (passkeyUser, SessionTokens, bool, error) {
	var user passkeyUser
	var result SessionTokens
	var replayed bool
	err := service.repository.WithinTransaction(ctx, func(transactionContext context.Context, base Transaction) error {
		transaction, ok := base.(passkeyTransaction)
		if !ok || nilIdentityValue(transaction) {
			return dependencyUnavailable()
		}
		rows, listErr := transaction.ListActivePasskeys(transactionContext, principalID)
		if listErr != nil {
			return dependencyUnavailable()
		}
		account, accountFound, accountErr := transaction.GetAccountForUpdate(transactionContext, principalID)
		if accountErr != nil {
			return dependencyUnavailable()
		}
		if !accountFound || account.ID != principalID || account.State != "active" || !containsPasskeyCredential(rows, credentialID) {
			return authenticationFailed()
		}
		loaded, loadErr := passkeyUserFromStore(principalID, rows)
		if loadErr != nil {
			return dependencyUnavailable()
		}
		user = loaded
		now, timeOK := service.now()
		if !timeOK {
			return dependencyUnavailable()
		}
		scope, scopeErr := idempotency.AuthenticatedScope(principalID, "passkey", "finish_passkey_authentication")
		if scopeErr != nil {
			return dependencyUnavailable()
		}
		record, outcome, beginErr := transaction.BeginIdempotency(
			transactionContext, scope, command.IdempotencyKey, canonical, now, now.Add(securityIdempotencyRetention),
		)
		if beginErr != nil {
			return dependencyUnavailable()
		}
		if outcome == idempotency.Started {
			return errSessionIdempotencyPreflight
		}
		body, wasReplayed, replayErr := privateIdempotencyOutcome(outcome, record, 200)
		if replayErr != nil {
			return replayErr
		}
		if !wasReplayed {
			return dependencyUnavailable()
		}
		defer clear(body)
		decoded, decodeErr := decodeSessionTokens(body)
		if decodeErr != nil {
			return dependencyUnavailable()
		}
		result = decoded
		replayed = true
		return nil
	})
	if errors.Is(err, errSessionIdempotencyPreflight) {
		return user, SessionTokens{}, false, nil
	}
	if err != nil {
		return passkeyUser{}, SessionTokens{}, false, err
	}
	return user, result, replayed, nil
}

func (service *Service) preflightPasskeyRegistrationFinish(
	ctx context.Context,
	command FinishPasskeyRegistrationCommand,
	principalID uuid.UUID,
	canonical []byte,
) (passkeyUser, bool, error) {
	var user passkeyUser
	var replayed bool
	err := service.repository.WithinTransaction(ctx, func(transactionContext context.Context, base Transaction) error {
		transaction, ok := base.(passkeyTransaction)
		if !ok || nilIdentityValue(transaction) {
			return dependencyUnavailable()
		}
		now, timeOK := service.now()
		if !timeOK {
			return dependencyUnavailable()
		}
		credential, reauthErr := service.loadPasswordReauthenticationCredential(transactionContext, transaction, principalID)
		if reauthErr != nil {
			return reauthErr
		}
		rows, listErr := transaction.ListActivePasskeys(transactionContext, principalID)
		if listErr != nil {
			return dependencyUnavailable()
		}
		if _, reauthErr = service.verifyLockedPasswordReauthentication(
			transactionContext, transaction, principalID, command.Reauthentication, credential, now,
		); reauthErr != nil {
			return reauthErr
		}
		loaded, loadErr := passkeyUserFromStore(principalID, rows)
		if loadErr != nil {
			return dependencyUnavailable()
		}
		user = loaded
		scope, scopeErr := idempotency.AuthenticatedScope(principalID, "passkey", "finish_passkey_registration")
		if scopeErr != nil {
			return dependencyUnavailable()
		}
		record, outcome, beginErr := transaction.BeginIdempotency(
			transactionContext, scope, command.IdempotencyKey, canonical, now, now.Add(securityIdempotencyRetention),
		)
		if beginErr != nil {
			return dependencyUnavailable()
		}
		if outcome == idempotency.Started {
			if len(rows) >= maximumActivePasskeys {
				return strongAuthStateConflict()
			}
			return errSessionIdempotencyPreflight
		}
		wasReplayed, replayErr := genericIdempotencyOutcome(outcome, record, 204)
		if replayErr != nil {
			return replayErr
		}
		if !wasReplayed {
			return dependencyUnavailable()
		}
		replayed = true
		return nil
	})
	if errors.Is(err, errSessionIdempotencyPreflight) {
		return user, false, nil
	}
	if err != nil {
		return passkeyUser{}, false, err
	}
	return user, replayed, nil
}

func (service *Service) completePasskeyAuthentication(
	ctx context.Context,
	command FinishPasskeyAuthenticationCommand,
	principalID uuid.UUID,
	canonical []byte,
	verified *wa.Credential,
) (SessionTokens, error) {
	var result SessionTokens
	err := service.repository.WithinTransaction(ctx, func(transactionContext context.Context, base Transaction) error {
		transaction, ok := base.(passkeyTransaction)
		if !ok || nilIdentityValue(transaction) {
			return dependencyUnavailable()
		}
		now, ok := service.now()
		if !ok {
			return dependencyUnavailable()
		}
		rows, listErr := transaction.ListActivePasskeys(transactionContext, principalID)
		if listErr != nil {
			return dependencyUnavailable()
		}
		stored, found := findPasskeyCredential(rows, verified.ID)
		account, accountFound, accountErr := transaction.GetAccountForUpdate(transactionContext, principalID)
		if accountErr != nil {
			return dependencyUnavailable()
		}
		if !found || !accountFound || account.ID != principalID || account.State != "active" {
			return authenticationFailed()
		}
		newCounter := int64(verified.Authenticator.SignCount)
		if stored.SignCount > 0 && newCounter <= stored.SignCount {
			return authenticationFailed()
		}
		scope, scopeErr := idempotency.AuthenticatedScope(principalID, "passkey", "finish_passkey_authentication")
		if scopeErr != nil {
			return dependencyUnavailable()
		}
		record, outcome, beginErr := transaction.BeginIdempotency(
			transactionContext, scope, command.IdempotencyKey, canonical, now, now.Add(securityIdempotencyRetention),
		)
		if beginErr != nil {
			return dependencyUnavailable()
		}
		if outcome != idempotency.Started {
			body, replayed, replayErr := privateIdempotencyOutcome(outcome, record, 200)
			if replayErr != nil {
				return replayErr
			}
			if replayed {
				defer clear(body)
				decoded, decodeErr := decodeSessionTokens(body)
				if decodeErr != nil {
					return dependencyUnavailable()
				}
				result = decoded
				return nil
			}
		}
		updated, updateErr := transaction.UpdatePasskeyCounter(transactionContext, store.UpdatePasskeyCounterParams{
			CredentialID: append([]byte(nil), verified.ID...), PrincipalID: principalID,
			SignCount: newCounter, ProtocolFlags: int16(verified.Flags.ProtocolValue()), UpdatedAt: now,
		})
		if updateErr != nil {
			return dependencyUnavailable()
		}
		if updated != 1 {
			return authenticationFailed()
		}
		issued, issueErr := service.issueAccountSession(
			transactionContext, transaction, principalID, command.ClientSigningPublicKey, now,
		)
		if issueErr != nil {
			return dependencyUnavailable()
		}
		result = issued
		body, encodeErr := encodeSessionTokens(result)
		if encodeErr != nil {
			return dependencyUnavailable()
		}
		defer clear(body)
		if completeErr := transaction.CompleteIdempotency(transactionContext, record, 200, body); completeErr != nil {
			return dependencyUnavailable()
		}
		return nil
	})
	return result, err
}

func findPasskeyCredential(
	rows []store.IdentityPasskeyCredential,
	credentialID []byte,
) (store.IdentityPasskeyCredential, bool) {
	for _, row := range rows {
		if bytes.Equal(row.CredentialID, credentialID) {
			return row, true
		}
	}
	return store.IdentityPasskeyCredential{}, false
}

func (service *Service) persistPasskeyRegistration(
	ctx context.Context,
	command FinishPasskeyRegistrationCommand,
	principalID uuid.UUID,
	canonical []byte,
	credential *wa.Credential,
) error {
	return service.repository.WithinTransaction(ctx, func(transactionContext context.Context, base Transaction) error {
		transaction, ok := base.(passkeyTransaction)
		if !ok || nilIdentityValue(transaction) {
			return dependencyUnavailable()
		}
		now, ok := service.now()
		if !ok {
			return dependencyUnavailable()
		}
		passwordCredential, reauthErr := service.loadPasswordReauthenticationCredential(transactionContext, transaction, principalID)
		if reauthErr != nil {
			return reauthErr
		}
		rows, listErr := transaction.ListActivePasskeys(transactionContext, principalID)
		if listErr != nil {
			return dependencyUnavailable()
		}
		if _, reauthErr = service.verifyLockedPasswordReauthentication(
			transactionContext, transaction, principalID, command.Reauthentication, passwordCredential, now,
		); reauthErr != nil {
			return reauthErr
		}
		if len(rows) >= maximumActivePasskeys || containsPasskeyCredential(rows, credential.ID) {
			return strongAuthStateConflict()
		}
		scope, scopeErr := idempotency.AuthenticatedScope(principalID, "passkey", "finish_passkey_registration")
		if scopeErr != nil {
			return dependencyUnavailable()
		}
		record, outcome, beginErr := transaction.BeginIdempotency(
			transactionContext, scope, command.IdempotencyKey, canonical, now, now.Add(securityIdempotencyRetention),
		)
		if beginErr != nil {
			return dependencyUnavailable()
		}
		if outcome != idempotency.Started {
			replayed, replayErr := genericIdempotencyOutcome(outcome, record, 204)
			if replayErr != nil || replayed {
				return replayErr
			}
		}
		transports := make([]string, len(credential.Transport))
		for index, transport := range credential.Transport {
			if !validWebAuthnTransport(transport) {
				return dependencyUnavailable()
			}
			transports[index] = string(transport)
		}
		flags := credential.Flags.ProtocolValue()
		if createErr := transaction.CreatePasskeyCredential(transactionContext, store.CreatePasskeyCredentialParams{
			CredentialID: append([]byte(nil), credential.ID...), PrincipalID: principalID,
			PublicKey: append([]byte(nil), credential.PublicKey...), AttestationFormat: credential.AttestationFormat,
			Transports: transports, ProtocolFlags: int16(flags), SignCount: int64(credential.Authenticator.SignCount), CreatedAt: now,
		}); createErr != nil {
			return dependencyUnavailable()
		}
		if completeErr := transaction.CompleteIdempotency(transactionContext, record, 204, nil); completeErr != nil {
			return dependencyUnavailable()
		}
		return nil
	})
}

func validVerifiedPasskey(credential *wa.Credential) bool {
	if credential == nil || len(credential.ID) < 16 || len(credential.ID) > 1024 || len(credential.PublicKey) < 32 ||
		len(credential.PublicKey) > 4096 || credential.AttestationFormat != "none" || credential.Authenticator.CloneWarning {
		return false
	}
	flags := credential.Flags.ProtocolValue()
	if !flags.HasUserPresent() || !flags.HasUserVerified() || (!flags.HasBackupEligible() && flags.HasBackupState()) {
		return false
	}
	for _, transport := range credential.Transport {
		if !validWebAuthnTransport(transport) {
			return false
		}
	}
	return true
}

func containsPasskeyCredential(rows []store.IdentityPasskeyCredential, credentialID []byte) bool {
	for _, row := range rows {
		if bytes.Equal(row.CredentialID, credentialID) {
			return true
		}
	}
	return false
}

func (service *Service) preflightPasskeyRegistrationBegin(
	ctx context.Context,
	command BeginPasskeyRegistrationCommand,
	principalID uuid.UUID,
	canonical []byte,
) (passkeyUser, json.RawMessage, bool, error) {
	var user passkeyUser
	var result json.RawMessage
	var replayed bool
	err := service.repository.WithinTransaction(ctx, func(transactionContext context.Context, base Transaction) error {
		transaction, ok := base.(passkeyTransaction)
		if !ok || nilIdentityValue(transaction) {
			return dependencyUnavailable()
		}
		now, ok := service.now()
		if !ok {
			return dependencyUnavailable()
		}
		credential, reauthErr := service.loadPasswordReauthenticationCredential(transactionContext, transaction, principalID)
		if reauthErr != nil {
			return reauthErr
		}
		rows, listErr := transaction.ListActivePasskeys(transactionContext, principalID)
		if listErr != nil {
			return dependencyUnavailable()
		}
		if _, reauthErr = service.verifyLockedPasswordReauthentication(
			transactionContext, transaction, principalID, command.Reauthentication, credential, now,
		); reauthErr != nil {
			return reauthErr
		}
		loaded, loadErr := passkeyUserFromStore(principalID, rows)
		if loadErr != nil {
			return dependencyUnavailable()
		}
		user = loaded
		scope, scopeErr := idempotency.AuthenticatedScope(principalID, "passkey", "begin_passkey_registration")
		if scopeErr != nil {
			return dependencyUnavailable()
		}
		record, outcome, beginErr := transaction.BeginIdempotency(
			transactionContext, scope, command.IdempotencyKey, canonical, now, now.Add(securityIdempotencyRetention),
		)
		if beginErr != nil {
			return dependencyUnavailable()
		}
		if outcome == idempotency.Started {
			if len(rows) >= maximumActivePasskeys {
				return strongAuthStateConflict()
			}
			return errSessionIdempotencyPreflight
		}
		body, wasReplayed, replayErr := privateIdempotencyOutcome(outcome, record, 200)
		if replayErr != nil {
			return replayErr
		}
		if !wasReplayed {
			return dependencyUnavailable()
		}
		defer clear(body)
		if !validPasskeyOptions(body) {
			return dependencyUnavailable()
		}
		result = append(json.RawMessage(nil), body...)
		replayed = true
		return nil
	})
	if errors.Is(err, errSessionIdempotencyPreflight) {
		return user, nil, false, nil
	}
	if err != nil {
		return passkeyUser{}, nil, false, err
	}
	return user, result, replayed, nil
}

func (service *Service) completePasskeyRegistrationOptions(
	ctx context.Context,
	command BeginPasskeyRegistrationCommand,
	principalID uuid.UUID,
	canonical []byte,
	options []byte,
) (json.RawMessage, error) {
	var result json.RawMessage
	err := service.repository.WithinTransaction(ctx, func(transactionContext context.Context, base Transaction) error {
		transaction, ok := base.(passkeyTransaction)
		if !ok || nilIdentityValue(transaction) {
			return dependencyUnavailable()
		}
		now, ok := service.now()
		if !ok {
			return dependencyUnavailable()
		}
		credential, reauthErr := service.loadPasswordReauthenticationCredential(transactionContext, transaction, principalID)
		if reauthErr != nil {
			return reauthErr
		}
		rows, listErr := transaction.ListActivePasskeys(transactionContext, principalID)
		if listErr != nil {
			return dependencyUnavailable()
		}
		if _, reauthErr = service.verifyLockedPasswordReauthentication(
			transactionContext, transaction, principalID, command.Reauthentication, credential, now,
		); reauthErr != nil {
			return reauthErr
		}
		if len(rows) >= maximumActivePasskeys {
			return strongAuthStateConflict()
		}
		scope, scopeErr := idempotency.AuthenticatedScope(principalID, "passkey", "begin_passkey_registration")
		if scopeErr != nil {
			return dependencyUnavailable()
		}
		record, outcome, beginErr := transaction.BeginIdempotency(
			transactionContext, scope, command.IdempotencyKey, canonical, now, now.Add(securityIdempotencyRetention),
		)
		if beginErr != nil {
			return dependencyUnavailable()
		}
		if outcome != idempotency.Started {
			body, replayed, replayErr := privateIdempotencyOutcome(outcome, record, 200)
			if replayErr != nil {
				return replayErr
			}
			if replayed {
				defer clear(body)
				if !validPasskeyOptions(body) {
					return dependencyUnavailable()
				}
				result = append(json.RawMessage(nil), body...)
				return nil
			}
		}
		if !validPasskeyOptions(options) {
			return dependencyUnavailable()
		}
		if completeErr := transaction.CompleteIdempotency(transactionContext, record, 200, options); completeErr != nil {
			return dependencyUnavailable()
		}
		result = append(json.RawMessage(nil), options...)
		return nil
	})
	return result, err
}

func (service *Service) newWebAuthn() (*wa.WebAuthn, error) {
	if service == nil || service.security.WebAuthnRPID == "" || len(service.security.WebAuthnOrigins) == 0 {
		return nil, ErrInvalidWebAuthnCeremony
	}
	timeout := wa.TimeoutConfig{Enforce: true, Timeout: webAuthnCeremonyTTL, TimeoutUVD: webAuthnCeremonyTTL}
	configured, err := safeWebAuthnProviderCall(func() (*wa.WebAuthn, error) {
		return wa.New(&wa.Config{
			RPID: service.security.WebAuthnRPID, RPDisplayName: "Talenro",
			RPOrigins:             append([]string(nil), service.security.WebAuthnOrigins...),
			AttestationPreference: protocol.PreferNoAttestation,
			AuthenticatorSelection: protocol.AuthenticatorSelection{
				ResidentKey: protocol.ResidentKeyRequirementPreferred, UserVerification: protocol.VerificationRequired,
			},
			Timeouts: wa.TimeoutsConfig{Login: timeout, Registration: timeout},
		})
	})
	if err != nil || configured == nil {
		return nil, ErrInvalidWebAuthnCeremony
	}
	return configured, nil
}

func webAuthnCeremonies(service *Service) (WebAuthnCeremonyStore, bool) {
	if service == nil || nilIdentityValue(service.challenges) {
		return nil, false
	}
	ceremonies, ok := service.challenges.(WebAuthnCeremonyStore)
	return ceremonies, ok && !nilIdentityValue(ceremonies)
}

func passkeyUserFromStore(principalID uuid.UUID, rows []store.IdentityPasskeyCredential) (passkeyUser, error) {
	if principalID == uuid.Nil || len(rows) > maximumActivePasskeys {
		return passkeyUser{}, ErrInvalidWebAuthnCeremony
	}
	credentials := make([]wa.Credential, len(rows))
	for index, row := range rows {
		if row.PrincipalID != principalID || len(row.CredentialID) < 16 || len(row.CredentialID) > 1024 ||
			len(row.PublicKey) < 32 || len(row.PublicKey) > 4096 || row.AttestationFormat == "" || len(row.AttestationFormat) > 64 ||
			row.ProtocolFlags < 0 || row.ProtocolFlags > math.MaxUint8 || row.SignCount < 0 || row.SignCount > math.MaxUint32 || row.State != "active" {
			return passkeyUser{}, ErrInvalidWebAuthnCeremony
		}
		flags := protocol.AuthenticatorFlags(row.ProtocolFlags)
		if !flags.HasUserPresent() || !flags.HasUserVerified() || (!flags.HasBackupEligible() && flags.HasBackupState()) {
			return passkeyUser{}, ErrInvalidWebAuthnCeremony
		}
		transports := make([]protocol.AuthenticatorTransport, len(row.Transports))
		for transportIndex, value := range row.Transports {
			transport := protocol.AuthenticatorTransport(value)
			if !validWebAuthnTransport(transport) {
				return passkeyUser{}, ErrInvalidWebAuthnCeremony
			}
			transports[transportIndex] = transport
		}
		credentials[index] = wa.Credential{
			ID: append([]byte(nil), row.CredentialID...), PublicKey: append([]byte(nil), row.PublicKey...),
			AttestationFormat: row.AttestationFormat, Transport: transports,
			Flags: wa.NewCredentialFlags(flags), Authenticator: wa.Authenticator{SignCount: uint32(row.SignCount)}, // #nosec G115 -- bounded above.
		}
	}
	return passkeyUser{
		id: append([]byte(nil), principalID[:]...), name: "account", displayName: "account", credentials: credentials,
	}, nil
}

func validWebAuthnTransport(value protocol.AuthenticatorTransport) bool {
	switch value {
	case protocol.USB, protocol.NFC, protocol.BLE, protocol.SmartCard, protocol.Hybrid, protocol.Internal:
		return true
	default:
		return false
	}
}

func validWebAuthnDisplayName(value string) bool {
	return value != "" && utf8.ValidString(value) && len(value) <= webAuthnDisplayNameBytes && utf8.RuneCountInString(value) <= webAuthnDisplayNameRunes
}

type passkeyRegistrationOptions struct {
	CeremonyID string                                      `json:"ceremony_id"`
	PublicKey  protocol.PublicKeyCredentialCreationOptions `json:"publicKey"`
	Mediation  protocol.CredentialMediationRequirement     `json:"mediation,omitempty"`
}

type passkeyAuthenticationOptions struct {
	CeremonyID string                                     `json:"ceremony_id"`
	PublicKey  protocol.PublicKeyCredentialRequestOptions `json:"publicKey"`
	Mediation  protocol.CredentialMediationRequirement    `json:"mediation,omitempty"`
}

func encodePasskeyRegistrationOptions(ceremonyID string, creation *protocol.CredentialCreation) ([]byte, error) {
	if _, ok := parseCanonicalIdentityUUID(ceremonyID); !ok || creation == nil {
		return nil, ErrInvalidWebAuthnCeremony
	}
	body, err := json.Marshal(passkeyRegistrationOptions{
		CeremonyID: ceremonyID, PublicKey: creation.Response, Mediation: creation.Mediation,
	})
	if err != nil || !validPasskeyOptions(body) {
		clear(body)
		return nil, ErrInvalidWebAuthnCeremony
	}
	return body, nil
}

func encodePasskeyAuthenticationOptions(ceremonyID string, assertion *protocol.CredentialAssertion) ([]byte, error) {
	if _, ok := parseCanonicalIdentityUUID(ceremonyID); !ok || assertion == nil {
		return nil, ErrInvalidWebAuthnCeremony
	}
	body, err := json.Marshal(passkeyAuthenticationOptions{
		CeremonyID: ceremonyID, PublicKey: assertion.Response, Mediation: assertion.Mediation,
	})
	if err != nil || !validPasskeyOptions(body) {
		clear(body)
		return nil, ErrInvalidWebAuthnCeremony
	}
	return body, nil
}

func validPasskeyOptions(body []byte) bool {
	return len(body) > 0 && len(body) <= maxWebAuthnCeremonyBytes && json.Valid(body)
}

// Format redacts passkey registration input.
func (BeginPasskeyRegistrationCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.BeginPasskeyRegistrationCommand([REDACTED])"))
}

// LogValue redacts passkey registration input from structured logs.
func (BeginPasskeyRegistrationCommand) LogValue() slog.Value {
	return slog.StringValue("identity.BeginPasskeyRegistrationCommand([REDACTED])")
}

// MarshalJSON forbids serialization of passkey registration input.
func (BeginPasskeyRegistrationCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: passkey registration command serialization forbidden")
}

// Format redacts passkey registration responses.
func (FinishPasskeyRegistrationCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.FinishPasskeyRegistrationCommand([REDACTED])"))
}

// LogValue redacts passkey registration responses from structured logs.
func (FinishPasskeyRegistrationCommand) LogValue() slog.Value {
	return slog.StringValue("identity.FinishPasskeyRegistrationCommand([REDACTED])")
}

// MarshalJSON forbids serialization of passkey registration responses.
func (FinishPasskeyRegistrationCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: passkey registration response serialization forbidden")
}

// Format redacts passkey authentication responses.
func (FinishPasskeyAuthenticationCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.FinishPasskeyAuthenticationCommand([REDACTED])"))
}

// LogValue redacts passkey authentication responses from structured logs.
func (FinishPasskeyAuthenticationCommand) LogValue() slog.Value {
	return slog.StringValue("identity.FinishPasskeyAuthenticationCommand([REDACTED])")
}

// MarshalJSON forbids serialization of passkey authentication responses.
func (FinishPasskeyAuthenticationCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: passkey authentication response serialization forbidden")
}

// Format redacts passkey revocation input.
func (RevokePasskeyCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.RevokePasskeyCommand([REDACTED])"))
}

// LogValue redacts passkey revocation input from structured logs.
func (RevokePasskeyCommand) LogValue() slog.Value {
	return slog.StringValue("identity.RevokePasskeyCommand([REDACTED])")
}

// MarshalJSON forbids serialization of passkey revocation input.
func (RevokePasskeyCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: passkey revocation command serialization forbidden")
}

// Format redacts ceremony protocol state.
func (WebAuthnCeremonyRecord) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.WebAuthnCeremonyRecord([REDACTED])"))
}

// LogValue redacts ceremony protocol state from structured logs.
func (WebAuthnCeremonyRecord) LogValue() slog.Value {
	return slog.StringValue("identity.WebAuthnCeremonyRecord([REDACTED])")
}

// MarshalJSON forbids generic serialization of ceremony protocol state.
func (WebAuthnCeremonyRecord) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: WebAuthn ceremony serialization forbidden")
}

func encodeWebAuthnCeremony(record WebAuthnCeremonyRecord) ([]byte, error) {
	if !validWebAuthnCeremonyRecord(record) {
		return nil, ErrInvalidWebAuthnCeremony
	}
	headerSize := len(webAuthnCeremonyRecordPrefix) + 1 + 8
	if record.Session.Msgsize() > maxWebAuthnCeremonyBytes-headerSize {
		return nil, ErrInvalidWebAuthnCeremony
	}
	session, err := record.Session.MarshalMsg(nil)
	if err != nil || len(session) > maxWebAuthnCeremonyBytes-headerSize {
		clear(session)
		return nil, ErrInvalidWebAuthnCeremony
	}
	defer clear(session)
	wire := make([]byte, 0, headerSize+len(session))
	wire = append(wire, webAuthnCeremonyRecordPrefix...)
	wire = append(wire, webAuthnCeremonyCode(record.Operation))
	var expiry [8]byte
	binary.BigEndian.PutUint64(expiry[:], uint64(record.ExpiresAt.UTC().UnixNano())) // #nosec G115 -- preserves same-width signed bits.
	wire = append(wire, expiry[:]...)
	wire = append(wire, session...)
	clear(expiry[:])
	return wire, nil
}

func decodeWebAuthnCeremony(
	ceremonyID string,
	operation WebAuthnCeremonyOperation,
	wire []byte,
) (WebAuthnCeremonyRecord, error) {
	headerSize := len(webAuthnCeremonyRecordPrefix) + 1 + 8
	if len(wire) <= headerSize || len(wire) > maxWebAuthnCeremonyBytes ||
		!bytes.Equal(wire[:min(len(wire), len(webAuthnCeremonyRecordPrefix))], []byte(webAuthnCeremonyRecordPrefix)) {
		return WebAuthnCeremonyRecord{}, ErrWebAuthnCeremonyUnavailable
	}
	offset := len(webAuthnCeremonyRecordPrefix)
	if wire[offset] != webAuthnCeremonyCode(operation) {
		return WebAuthnCeremonyRecord{}, ErrWebAuthnCeremonyUnavailable
	}
	offset++
	// #nosec G115 -- restores the same-width signed two's-complement bits.
	expiresUnixNanos := int64(binary.BigEndian.Uint64(wire[offset : offset+8]))
	offset += 8
	record := WebAuthnCeremonyRecord{
		CeremonyID: ceremonyID,
		Operation:  operation,
		ExpiresAt:  time.Unix(0, expiresUnixNanos).UTC(),
	}
	rest, err := record.Session.UnmarshalMsg(wire[offset:])
	if err != nil || len(rest) != 0 || !validWebAuthnCeremonyRecord(record) {
		return WebAuthnCeremonyRecord{}, ErrWebAuthnCeremonyUnavailable
	}
	return record, nil
}

func validWebAuthnCeremonyRecord(record WebAuthnCeremonyRecord) bool {
	parsedID, err := uuid.Parse(record.CeremonyID)
	if err != nil || parsedID == uuid.Nil || parsedID.String() != record.CeremonyID ||
		webAuthnCeremonyCode(record.Operation) == 0 || record.Session.Challenge == "" ||
		record.Session.RelyingPartyID == "" || len(record.Session.RelyingPartyID) > 253 ||
		len(record.Session.UserID) > 64 || len(record.Session.AllowedCredentialIDs) > 10 ||
		len(record.Session.Extensions) != 0 || len(record.Session.CredParams) > 16 ||
		record.Session.UserVerification != protocol.VerificationRequired ||
		record.ExpiresAt.IsZero() || !record.Session.Expires.Equal(record.ExpiresAt) ||
		record.ExpiresAt.Year() < 2020 || record.ExpiresAt.Year() > 2100 {
		return false
	}
	for _, credentialID := range record.Session.AllowedCredentialIDs {
		if len(credentialID) == 0 || len(credentialID) > 1024 {
			return false
		}
	}
	return true
}

func webAuthnCeremonyCode(operation WebAuthnCeremonyOperation) byte {
	switch operation {
	case WebAuthnRegistrationCeremony:
		return 1
	case WebAuthnAuthenticationCeremony:
		return 2
	default:
		return 0
	}
}

func webAuthnCeremonyRedisKey(operation WebAuthnCeremonyOperation, ceremonyID string) string {
	switch operation {
	case WebAuthnRegistrationCeremony:
		return "talenro:identity:webauthn-registration:" + ceremonyID
	case WebAuthnAuthenticationCeremony:
		return "talenro:identity:webauthn-authentication:" + ceremonyID
	default:
		return ""
	}
}
