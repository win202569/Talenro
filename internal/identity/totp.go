package identity

import (
	"context"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

const (
	totpSecretBytes           = 20
	totpPeriod                = 30
	totpProtectionDomain      = "identity/totp/v1"
	maximumTOTPCiphertextSize = 128
	maximumTOTPReplaySize     = 1024
)

var task11Base32 = base32.StdEncoding.WithPadding(base32.NoPadding)

var errInvalidTOTPSecret = errors.New("identity: invalid TOTP secret")

// BeginTOTPEnrollmentCommand starts one independently revocable TOTP factor.
type BeginTOTPEnrollmentCommand struct {
	PrincipalID      PrincipalID
	Reauthentication Reauthentication
	IdempotencyKey   string
}

// TOTPEnrollment exposes the one-time base32 secret and matching provisioning URI.
type TOTPEnrollment struct {
	Secret string
	URI    string
}

// VerifyTOTPEnrollmentCommand verifies the pending factor's first accepted step.
type VerifyTOTPEnrollmentCommand struct {
	PrincipalID      PrincipalID
	Code             secret.Bytes
	Reauthentication Reauthentication
	IdempotencyKey   string
}

// RevokeTOTPCommand independently revokes the principal's TOTP factor.
type RevokeTOTPCommand struct {
	PrincipalID      PrincipalID
	Reauthentication Reauthentication
	IdempotencyKey   string
}

type totpTransaction interface {
	Transaction
	GetTOTPForUpdate(context.Context, uuid.UUID) (store.IdentityTotpCredential, bool, error)
	CreateTOTPEnrollment(context.Context, store.CreateTOTPEnrollmentParams) (int64, error)
	ActivateTOTP(context.Context, store.ActivateTOTPParams) (int64, error)
	AcceptTOTPStep(context.Context, store.AcceptTOTPStepParams) (int64, error)
	RevokeTOTP(context.Context, store.RevokeTOTPParams) (int64, error)
}

// BeginTOTPEnrollment returns one idempotently replayable provisioning URI.
func (service *Service) BeginTOTPEnrollment(
	ctx context.Context,
	command BeginTOTPEnrollmentCommand,
) (TOTPEnrollment, error) {
	principalID, principalOK := parseCanonicalIdentityUUID(string(command.PrincipalID))
	sessionID, sessionOK := parseCanonicalIdentityUUID(string(command.Reauthentication.SessionID))
	if !validService(service) || nilIdentityValue(ctx) || !principalOK || !sessionOK ||
		!validIdempotencyKeyForApplication(command.IdempotencyKey) {
		return TOTPEnrollment{}, malformedRequest()
	}
	if command.Reauthentication.Method != ReauthPassword {
		return TOTPEnrollment{}, actionNotAllowed()
	}
	if !validPasswordSecret(command.Reauthentication.Proof) {
		return TOTPEnrollment{}, malformedRequest()
	}
	proof := command.Reauthentication.Proof.Copy()
	defer clear(proof)
	canonical, err := strongAuthCanonicalRequest(
		service.protector, "begin_totp_enrollment", principalID[:], sessionID[:], proof,
	)
	if err != nil {
		return TOTPEnrollment{}, dependencyUnavailable()
	}
	defer clear(canonical)
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	var result TOTPEnrollment
	err = service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, base Transaction) error {
		transaction, ok := base.(totpTransaction)
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
		if _, _, factorErr := transaction.GetTOTPForUpdate(transactionContext, principalID); factorErr != nil {
			return dependencyUnavailable()
		}
		if _, reauthErr = service.verifyLockedPasswordReauthentication(
			transactionContext, transaction, principalID, command.Reauthentication, credential, now,
		); reauthErr != nil {
			return reauthErr
		}
		scope, scopeErr := idempotency.AuthenticatedScope(principalID, "totp", "begin_totp_enrollment")
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
				decoded, decodeErr := decodeTOTPEnrollment(body)
				if decodeErr != nil {
					return dependencyUnavailable()
				}
				result = decoded
				return nil
			}
		}
		secretBytes := make([]byte, totpSecretBytes)
		if randomErr := readTask11Random(service.random, secretBytes); randomErr != nil {
			clear(secretBytes)
			return dependencyUnavailable()
		}
		defer clear(secretBytes)
		uri, uriErr := buildTOTPEnrollmentURI(secretBytes)
		if uriErr != nil {
			return dependencyUnavailable()
		}
		protected, protectErr := service.protector.Encrypt(totpProtectionDomain, secretBytes)
		if protectErr != nil || !validProtectedField(protected, maximumTOTPCiphertextSize) {
			clear(protected.Ciphertext)
			return dependencyUnavailable()
		}
		defer clear(protected.Ciphertext)
		keyVersion, versionOK := protectedKeyVersionInt32(protected.KeyVersion)
		if !versionOK {
			return dependencyUnavailable()
		}
		persistedCiphertext := append([]byte(nil), protected.Ciphertext...)
		defer clear(persistedCiphertext)
		rows, createErr := transaction.CreateTOTPEnrollment(transactionContext, store.CreateTOTPEnrollmentParams{
			PrincipalID: principalID, EncryptedSecret: persistedCiphertext,
			KeyVersion: keyVersion, EnrolledAt: now,
		})
		if createErr != nil {
			return dependencyUnavailable()
		}
		if rows != 1 {
			return strongAuthStateConflict()
		}
		result = TOTPEnrollment{Secret: task11Base32.EncodeToString(secretBytes), URI: uri}
		body, encodeErr := encodeTOTPEnrollment(result)
		if encodeErr != nil {
			return dependencyUnavailable()
		}
		defer clear(body)
		if completeErr := transaction.CompleteIdempotency(transactionContext, record, 200, body); completeErr != nil {
			return dependencyUnavailable()
		}
		return nil
	})
	if err != nil {
		return TOTPEnrollment{}, mapApplicationError(operationContext, err)
	}
	return result, nil
}

// VerifyTOTPEnrollment activates the pending factor and advances its replay barrier in one transaction.
func (service *Service) VerifyTOTPEnrollment(ctx context.Context, command VerifyTOTPEnrollmentCommand) error {
	principalID, principalOK := parseCanonicalIdentityUUID(string(command.PrincipalID))
	sessionID, sessionOK := parseCanonicalIdentityUUID(string(command.Reauthentication.SessionID))
	code := command.Code.Copy()
	defer clear(code)
	if !validService(service) || nilIdentityValue(ctx) || !principalOK || !sessionOK ||
		!validTOTPCode(code) || !validIdempotencyKeyForApplication(command.IdempotencyKey) {
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
		service.protector, "verify_totp_enrollment", principalID[:], sessionID[:], code, proof,
	)
	if err != nil {
		return dependencyUnavailable()
	}
	defer clear(canonical)
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	err = service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, base Transaction) error {
		transaction, ok := base.(totpTransaction)
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
		factor, factorFound, factorErr := transaction.GetTOTPForUpdate(transactionContext, principalID)
		if factorErr != nil {
			return dependencyUnavailable()
		}
		if _, reauthErr = service.verifyLockedPasswordReauthentication(
			transactionContext, transaction, principalID, command.Reauthentication, credential, now,
		); reauthErr != nil {
			return reauthErr
		}
		scope, scopeErr := idempotency.AuthenticatedScope(principalID, "totp", "verify_totp_enrollment")
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
		if factor.EncryptionKeyVersion <= 0 {
			return authenticationFailed()
		}
		// #nosec G115 -- the persisted key version is positive and represented by int32.
		keyVersion := uint32(factor.EncryptionKeyVersion)
		if !factorFound || factor.PrincipalID != principalID || factor.State != "pending" ||
			!validProtectedField(sensitive.EncryptedField{KeyVersion: keyVersion, Ciphertext: factor.Ciphertext}, maximumTOTPCiphertextSize) {
			return authenticationFailed()
		}
		plaintext, decryptErr := service.protector.Decrypt(totpProtectionDomain, sensitive.EncryptedField{
			KeyVersion: keyVersion, Ciphertext: factor.Ciphertext,
		})
		if decryptErr != nil || len(plaintext) != totpSecretBytes {
			clear(plaintext)
			return dependencyUnavailable()
		}
		defer clear(plaintext)
		step, accepted := verifyTOTPStep(plaintext, string(code), now)
		if !accepted {
			return authenticationFailed()
		}
		activated, activateErr := transaction.ActivateTOTP(transactionContext, store.ActivateTOTPParams{
			PrincipalID: principalID, VerifiedAt: sql.NullTime{Time: now, Valid: true},
		})
		if activateErr != nil {
			return dependencyUnavailable()
		}
		if activated != 1 {
			return strongAuthStateConflict()
		}
		acceptedRows, acceptErr := transaction.AcceptTOTPStep(transactionContext, store.AcceptTOTPStepParams{
			PrincipalID: principalID, LastAcceptedStep: pgtype.Int8{Int64: step, Valid: true},
		})
		if acceptErr != nil {
			return dependencyUnavailable()
		}
		if acceptedRows != 1 {
			return authenticationFailed()
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

// RevokeTOTP independently revokes the active or pending factor after password reauthentication.
func (service *Service) RevokeTOTP(ctx context.Context, command RevokeTOTPCommand) error {
	principalID, principalOK := parseCanonicalIdentityUUID(string(command.PrincipalID))
	sessionID, sessionOK := parseCanonicalIdentityUUID(string(command.Reauthentication.SessionID))
	if !validService(service) || nilIdentityValue(ctx) || !principalOK || !sessionOK ||
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
	canonical, err := strongAuthCanonicalRequest(service.protector, "revoke_totp", principalID[:], sessionID[:], proof)
	if err != nil {
		return dependencyUnavailable()
	}
	defer clear(canonical)
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	err = service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, base Transaction) error {
		transaction, ok := base.(totpTransaction)
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
		if _, _, factorErr := transaction.GetTOTPForUpdate(transactionContext, principalID); factorErr != nil {
			return dependencyUnavailable()
		}
		if _, reauthErr = service.verifyLockedPasswordReauthentication(
			transactionContext, transaction, principalID, command.Reauthentication, credential, now,
		); reauthErr != nil {
			return reauthErr
		}
		scope, scopeErr := idempotency.AuthenticatedScope(principalID, "totp", "revoke_totp")
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
		rows, revokeErr := transaction.RevokeTOTP(transactionContext, store.RevokeTOTPParams{
			PrincipalID: principalID, RevokedAt: sql.NullTime{Time: now, Valid: true},
		})
		if revokeErr != nil {
			return dependencyUnavailable()
		}
		if rows != 1 {
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

func buildTOTPEnrollmentURI(secretBytes []byte) (string, error) {
	if len(secretBytes) != totpSecretBytes {
		return "", errInvalidTOTPSecret
	}
	owned := append([]byte(nil), secretBytes...)
	defer clear(owned)
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Talenro",
		AccountName: "account",
		Period:      totpPeriod,
		SecretSize:  totpSecretBytes,
		Secret:      owned,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil || key == nil {
		return "", ErrRepository
	}
	uri := key.URL()
	if uri == "" || len(uri) > 512 {
		return "", ErrRepository
	}
	return uri, nil
}

func validTOTPCode(code []byte) bool {
	if len(code) != 6 {
		return false
	}
	for _, character := range code {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

type totpEnrollmentReplay struct {
	Secret string `json:"secret"`
	URI    string `json:"uri"`
}

func encodeTOTPEnrollment(enrollment TOTPEnrollment) ([]byte, error) {
	if !validTOTPEnrollment(enrollment) {
		return nil, errInvalidTOTPSecret
	}
	// #nosec G117 -- this fixed private replay DTO is immediately encrypted by the idempotency repository.
	body, err := json.Marshal(totpEnrollmentReplay(enrollment))
	if err != nil || len(body) == 0 || len(body) > maximumTOTPReplaySize {
		clear(body)
		return nil, errInvalidTOTPSecret
	}
	return body, nil
}

func decodeTOTPEnrollment(body []byte) (TOTPEnrollment, error) {
	if len(body) == 0 || len(body) > maximumTOTPReplaySize {
		return TOTPEnrollment{}, errInvalidTOTPSecret
	}
	var replay totpEnrollmentReplay
	if err := json.Unmarshal(body, &replay); err != nil || !validTOTPEnrollment(TOTPEnrollment(replay)) {
		return TOTPEnrollment{}, errInvalidTOTPSecret
	}
	return TOTPEnrollment(replay), nil
}

func validTOTPEnrollment(enrollment TOTPEnrollment) bool {
	if len(enrollment.Secret) != 32 || !validTOTPEnrollmentURI(enrollment.URI) {
		return false
	}
	decoded, err := task11Base32.DecodeString(enrollment.Secret)
	if err != nil || len(decoded) != totpSecretBytes {
		clear(decoded)
		return false
	}
	clear(decoded)
	parsed, err := url.Parse(enrollment.URI)
	return err == nil && parsed.Query().Get("secret") == enrollment.Secret
}

func validTOTPEnrollmentURI(value string) bool {
	if value == "" || len(value) > 512 {
		return false
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "otpauth" || parsed.Host != "totp" || parsed.User != nil || parsed.Fragment != "" {
		return false
	}
	query := parsed.Query()
	return query.Get("issuer") == "Talenro" && query.Get("digits") == "6" && query.Get("algorithm") == "SHA1" &&
		query.Get("period") == "30" && len(query.Get("secret")) == 32 && strings.HasPrefix(parsed.Path, "/Talenro:")
}

// Format redacts TOTP enrollment input.
func (BeginTOTPEnrollmentCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.BeginTOTPEnrollmentCommand([REDACTED])"))
}

// LogValue redacts TOTP enrollment input from structured logs.
func (BeginTOTPEnrollmentCommand) LogValue() slog.Value {
	return slog.StringValue("identity.BeginTOTPEnrollmentCommand([REDACTED])")
}

// MarshalJSON forbids serialization of TOTP enrollment input.
func (BeginTOTPEnrollmentCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: TOTP enrollment command serialization forbidden")
}

// Format redacts the one-time TOTP secret and provisioning URI.
func (TOTPEnrollment) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.TOTPEnrollment([REDACTED])"))
}

// LogValue redacts the TOTP secret and provisioning URI from structured logs.
func (TOTPEnrollment) LogValue() slog.Value {
	return slog.StringValue("identity.TOTPEnrollment([REDACTED])")
}

// MarshalJSON forbids generic serialization of the TOTP secret and provisioning URI.
func (TOTPEnrollment) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: TOTP enrollment serialization forbidden")
}

// Format redacts TOTP verification input.
func (VerifyTOTPEnrollmentCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.VerifyTOTPEnrollmentCommand([REDACTED])"))
}

// LogValue redacts TOTP verification input from structured logs.
func (VerifyTOTPEnrollmentCommand) LogValue() slog.Value {
	return slog.StringValue("identity.VerifyTOTPEnrollmentCommand([REDACTED])")
}

// MarshalJSON forbids serialization of TOTP verification input.
func (VerifyTOTPEnrollmentCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: TOTP verification command serialization forbidden")
}

// Format redacts TOTP revocation input.
func (RevokeTOTPCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.RevokeTOTPCommand([REDACTED])"))
}

// LogValue redacts TOTP revocation input from structured logs.
func (RevokeTOTPCommand) LogValue() slog.Value {
	return slog.StringValue("identity.RevokeTOTPCommand([REDACTED])")
}

// MarshalJSON forbids serialization of TOTP revocation input.
func (RevokeTOTPCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: TOTP revocation command serialization forbidden")
}

func verifyTOTPStep(secretBytes []byte, code string, now time.Time) (int64, bool) {
	if len(secretBytes) != totpSecretBytes || len(code) != 6 || now.IsZero() {
		return 0, false
	}
	for index := range code {
		if code[index] < '0' || code[index] > '9' {
			return 0, false
		}
	}
	secret := task11Base32.EncodeToString(secretBytes)
	for _, offset := range [...]int64{0, -1, 1} {
		step := now.Unix()/totpPeriod + offset
		valid, err := totp.ValidateCustom(code, secret, time.Unix(step*totpPeriod, 0).UTC(), totp.ValidateOpts{
			Period: totpPeriod, Skew: 0, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1,
		})
		if err != nil {
			return 0, false
		}
		if valid {
			return step, true
		}
	}
	return 0, false
}
