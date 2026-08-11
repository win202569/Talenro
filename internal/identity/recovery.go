package identity

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/ratelimit"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/store"
)

const (
	recoveryCodeCount        = 10
	recoveryCodeBytes        = 20
	recoveryCodeDigestPrefix = "TALENRO-RECOVERY-CODE-V1\x00"
	maximumRandomEmptyReads  = 3
	maximumRecoveryReplay    = 2048
)

var errStrongAuthDependency = errors.New("identity: strong authentication dependency unavailable")

// RotateRecoveryCodesCommand replaces any active recovery set after password reauthentication.
type RotateRecoveryCodesCommand struct {
	PrincipalID      PrincipalID
	Reauthentication Reauthentication
	IdempotencyKey   string
}

// RecoveryCodes contains the ten one-time codes shown exactly once to the account holder.
type RecoveryCodes struct {
	Codes []string
}

// ConsumeRecoveryCodeCommand consumes one recovery code and creates a bound account session.
type ConsumeRecoveryCodeCommand struct {
	Email                  string
	Code                   secret.Bytes
	ClientSigningPublicKey [32]byte
	IdempotencyKey         string
}

type recoveryTransaction interface {
	Transaction
	GetActiveRecoveryCodeSetForUpdate(context.Context, uuid.UUID) (store.IdentityRecoveryCodeSet, bool, error)
	GetNextRecoveryCodeGeneration(context.Context, uuid.UUID) (int32, error)
	CreateRecoveryCodeSet(context.Context, store.CreateRecoveryCodeSetParams) error
	ConsumeRecoveryCode(context.Context, store.ConsumeRecoveryCodeParams) (store.IdentityRecoveryCodeSet, bool, error)
	RevokeRecoveryCodeSets(context.Context, store.RevokeRecoveryCodeSetsParams) (int64, error)
}

// RotateRecoveryCodes supersedes the active set and returns ten codes through an encrypted idempotency response.
func (service *Service) RotateRecoveryCodes(
	ctx context.Context,
	command RotateRecoveryCodesCommand,
) (RecoveryCodes, error) {
	principalID, principalOK := parseCanonicalIdentityUUID(string(command.PrincipalID))
	sessionID, sessionOK := parseCanonicalIdentityUUID(string(command.Reauthentication.SessionID))
	if !validService(service) || nilIdentityValue(ctx) || !principalOK || !sessionOK ||
		!validIdempotencyKeyForApplication(command.IdempotencyKey) {
		return RecoveryCodes{}, malformedRequest()
	}
	if command.Reauthentication.Method != ReauthPassword {
		return RecoveryCodes{}, actionNotAllowed()
	}
	if !validPasswordSecret(command.Reauthentication.Proof) {
		return RecoveryCodes{}, malformedRequest()
	}
	proof := command.Reauthentication.Proof.Copy()
	defer clear(proof)
	canonical, err := strongAuthCanonicalRequest(
		service.protector, "rotate_recovery_codes", principalID[:], sessionID[:], proof,
	)
	if err != nil {
		return RecoveryCodes{}, dependencyUnavailable()
	}
	defer clear(canonical)
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	var result RecoveryCodes
	err = service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, base Transaction) error {
		transaction, ok := base.(recoveryTransaction)
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
		activeSet, activeFound, findErr := transaction.GetActiveRecoveryCodeSetForUpdate(transactionContext, principalID)
		if findErr != nil {
			return dependencyUnavailable()
		}
		if activeFound && !validRecoveryCodeSet(activeSet, principalID) {
			return dependencyUnavailable()
		}
		if _, reauthErr = service.verifyLockedPasswordReauthentication(
			transactionContext, transaction, principalID, command.Reauthentication, credential, now,
		); reauthErr != nil {
			return reauthErr
		}
		scope, scopeErr := idempotency.AuthenticatedScope(principalID, "recovery", "rotate_recovery_codes")
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
				decoded, decodeErr := decodeRecoveryCodes(body)
				if decodeErr != nil {
					return dependencyUnavailable()
				}
				result = decoded
				return nil
			}
		}
		generation, generationErr := transaction.GetNextRecoveryCodeGeneration(transactionContext, principalID)
		if generationErr != nil || generation < 1 || (activeFound && generation <= activeSet.Generation) {
			return dependencyUnavailable()
		}
		codes, hashes, generateErr := generateRecoveryCodeSet(service.random)
		if generateErr != nil || len(codes) != recoveryCodeCount || !validRecoveryHashes(hashes) {
			clearRecoveryHashes(hashes)
			return dependencyUnavailable()
		}
		defer clearRecoveryHashes(hashes)
		revoked, revokeErr := transaction.RevokeRecoveryCodeSets(transactionContext, store.RevokeRecoveryCodeSetsParams{
			PrincipalID: principalID, State: "superseded", UpdatedAt: now,
		})
		if revokeErr != nil || revoked < 0 || revoked > 1 || (activeFound && revoked != 1) || (!activeFound && revoked != 0) {
			return dependencyUnavailable()
		}
		setID := service.newUUID()
		if setID == uuid.Nil {
			return dependencyUnavailable()
		}
		persistedHashes := cloneRecoveryHashes(hashes)
		defer clearRecoveryHashes(persistedHashes)
		if createErr := transaction.CreateRecoveryCodeSet(transactionContext, store.CreateRecoveryCodeSetParams{
			ID: setID, PrincipalID: principalID, Generation: generation,
			CodeHashes: persistedHashes, CreatedAt: now,
		}); createErr != nil {
			return dependencyUnavailable()
		}
		result = RecoveryCodes{Codes: append([]string(nil), codes...)}
		body, encodeErr := encodeRecoveryCodes(result)
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
		return RecoveryCodes{}, mapApplicationError(operationContext, err)
	}
	return result, nil
}

// ConsumeRecoveryCode removes one digest and issues a session atomically in PostgreSQL.
func (service *Service) ConsumeRecoveryCode(
	ctx context.Context,
	command ConsumeRecoveryCodeCommand,
) (SessionTokens, error) {
	code := command.Code.Copy()
	defer clear(code)
	if !validService(service) || nilIdentityValue(ctx) || command.Email == "" ||
		command.ClientSigningPublicKey == [32]byte{} || !validIdempotencyKeyForApplication(command.IdempotencyKey) {
		return SessionTokens{}, malformedRequest()
	}
	codeDigest, codeOK := recoveryCodeDigest(string(code))
	if !codeOK {
		return SessionTokens{}, malformedRequest()
	}
	defer clear(codeDigest[:])
	canonicalEmail, emailErr := CanonicalizeEmail(command.Email)
	if emailErr != nil {
		return SessionTokens{}, malformedRequest()
	}
	emailBytes := canonicalEmail.Bytes()
	defer clear(emailBytes)
	canonical, err := strongAuthCanonicalRequest(
		service.protector, "consume_recovery_code", emailBytes, code, command.ClientSigningPublicKey[:],
	)
	if err != nil {
		return SessionTokens{}, dependencyUnavailable()
	}
	defer clear(canonical)
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	now, ok := service.now()
	if !ok {
		return SessionTokens{}, dependencyUnavailable()
	}
	subject, digestErr := ratelimit.SubjectDigest(
		service.rateLimitKey, ratelimit.Login, now, service.security.LoginRateLimit, string(emailBytes),
	)
	if digestErr != nil {
		return SessionTokens{}, dependencyUnavailable()
	}
	allowed, limitErr := service.rateLimiter.Allow(operationContext, ratelimit.Login, subject, service.security.LoginRateLimit)
	if limitErr != nil {
		return SessionTokens{}, dependencyUnavailable()
	}
	if !allowed {
		return SessionTokens{}, apierrors.NewRetryAfter(apierrors.RateLimited, apierrors.Retry, time.Second)
	}
	var result SessionTokens
	err = service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, base Transaction) error {
		transaction, typed := base.(recoveryTransaction)
		if !typed || nilIdentityValue(transaction) {
			return dependencyUnavailable()
		}
		lookupDigest := service.protector.LookupDigest(emailFieldDomain, emailBytes)
		if lookupDigest == [sha256.Size]byte{} {
			return dependencyUnavailable()
		}
		identity, identityFound, findErr := transaction.FindIdentityByLookupDigestRead(transactionContext, lookupDigest[:])
		clear(lookupDigest[:])
		if findErr != nil {
			return dependencyUnavailable()
		}
		principalID := uuid.Nil
		if identityFound {
			principalID = identity.PrincipalID
		}
		activeSet, activeFound, findErr := transaction.GetActiveRecoveryCodeSetForUpdate(transactionContext, principalID)
		if findErr != nil {
			return dependencyUnavailable()
		}
		account, accountFound, accountErr := transaction.GetAccountForUpdate(transactionContext, principalID)
		if accountErr != nil {
			return dependencyUnavailable()
		}
		if !identityFound || !accountFound || identity.PrincipalID != principalID || account.ID != principalID || account.State != "active" {
			return authenticationFailed()
		}
		scope, scopeErr := idempotency.AuthenticatedScope(principalID, "recovery", "consume_recovery_code")
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
		if !activeFound || !validRecoveryCodeSet(activeSet, principalID) {
			return authenticationFailed()
		}
		persistedDigest := append([]byte(nil), codeDigest[:]...)
		defer clear(persistedDigest)
		updatedSet, consumed, consumeErr := transaction.ConsumeRecoveryCode(transactionContext, store.ConsumeRecoveryCodeParams{
			ID: activeSet.ID, Column2: persistedDigest, UpdatedAt: now,
		})
		if consumeErr != nil {
			return dependencyUnavailable()
		}
		if !consumed {
			return authenticationFailed()
		}
		expectedCount := len(activeSet.CodeHashes) - 1
		expectedState := "active"
		if expectedCount == 0 {
			expectedState = "exhausted"
		}
		if updatedSet.ID != activeSet.ID || updatedSet.PrincipalID != principalID || updatedSet.Generation != activeSet.Generation || updatedSet.State != expectedState ||
			!validRecoveryHashCount(updatedSet.CodeHashes, expectedCount) || containsRecoveryHash(updatedSet.CodeHashes, codeDigest[:]) {
			return dependencyUnavailable()
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
	if err != nil {
		return SessionTokens{}, mapApplicationError(operationContext, err)
	}
	return result, nil
}

func generateRecoveryCodeSet(random securitykit.RandomSource) ([]string, [][]byte, error) {
	if nilIdentityValue(random) {
		return nil, nil, errStrongAuthDependency
	}
	codes := make([]string, 0, recoveryCodeCount)
	hashes := make([][]byte, 0, recoveryCodeCount)
	seen := make(map[[sha256.Size]byte]struct{}, recoveryCodeCount)
	for range recoveryCodeCount {
		raw := make([]byte, recoveryCodeBytes)
		if err := readTask11Random(random, raw); err != nil {
			clear(raw)
			clearRecoveryHashes(hashes)
			return nil, nil, errStrongAuthDependency
		}
		compact := task11Base32.EncodeToString(raw)
		code := groupRecoveryCode(compact)
		digest, valid := recoveryCodeDigest(code)
		clear(raw)
		if !valid {
			clearRecoveryHashes(hashes)
			return nil, nil, errStrongAuthDependency
		}
		if _, duplicate := seen[digest]; duplicate {
			clearRecoveryHashes(hashes)
			return nil, nil, errStrongAuthDependency
		}
		seen[digest] = struct{}{}
		codes = append(codes, code)
		hashes = append(hashes, append([]byte(nil), digest[:]...))
		clear(digest[:])
	}
	return codes, hashes, nil
}

func recoveryCodeDigest(code string) ([sha256.Size]byte, bool) {
	if len(code) != 39 {
		return [sha256.Size]byte{}, false
	}
	parts := strings.Split(code, "-")
	if len(parts) != 8 {
		return [sha256.Size]byte{}, false
	}
	for _, part := range parts {
		if len(part) != 4 {
			return [sha256.Size]byte{}, false
		}
	}
	compact := strings.Join(parts, "")
	raw, err := task11Base32.DecodeString(compact)
	if err != nil || len(raw) != recoveryCodeBytes || groupRecoveryCode(task11Base32.EncodeToString(raw)) != code {
		clear(raw)
		return [sha256.Size]byte{}, false
	}
	material := make([]byte, 0, len(recoveryCodeDigestPrefix)+len(raw))
	material = append(material, recoveryCodeDigestPrefix...)
	material = append(material, raw...)
	digest := sha256.Sum256(material)
	clear(material)
	clear(raw)
	return digest, true
}

func groupRecoveryCode(compact string) string {
	if len(compact) != 32 {
		return ""
	}
	var builder strings.Builder
	builder.Grow(39)
	for index := 0; index < len(compact); index += 4 {
		if index != 0 {
			builder.WriteByte('-')
		}
		builder.WriteString(compact[index : index+4])
	}
	return builder.String()
}

func readTask11Random(random securitykit.RandomSource, target []byte) error {
	if nilIdentityValue(random) || len(target) == 0 {
		return errStrongAuthDependency
	}
	emptyReads := 0
	for offset := 0; offset < len(target); {
		count, err := random.Read(target[offset:])
		if count < 0 || count > len(target)-offset {
			clear(target)
			return errStrongAuthDependency
		}
		offset += count
		if err != nil {
			clear(target)
			return errStrongAuthDependency
		}
		if count == 0 {
			emptyReads++
			if emptyReads >= maximumRandomEmptyReads {
				clear(target)
				return io.ErrNoProgress
			}
		} else {
			emptyReads = 0
		}
	}
	return nil
}

func clearRecoveryHashes(hashes [][]byte) {
	for _, hash := range hashes {
		clear(hash)
	}
}

func validRecoveryCodeSet(set store.IdentityRecoveryCodeSet, principalID uuid.UUID) bool {
	return set.ID != uuid.Nil && set.PrincipalID == principalID && set.Generation >= 1 && set.State == "active" &&
		len(set.CodeHashes) >= 1 && len(set.CodeHashes) <= recoveryCodeCount && validRecoveryHashCount(set.CodeHashes, len(set.CodeHashes))
}

func validRecoveryHashes(hashes [][]byte) bool {
	return validRecoveryHashCount(hashes, recoveryCodeCount)
}

func validRecoveryHashCount(hashes [][]byte, count int) bool {
	if count < 0 || count > recoveryCodeCount || len(hashes) != count {
		return false
	}
	seen := make(map[[sha256.Size]byte]struct{}, count)
	for _, hash := range hashes {
		if len(hash) != sha256.Size {
			return false
		}
		var value [sha256.Size]byte
		copy(value[:], hash)
		if value == [sha256.Size]byte{} {
			return false
		}
		if _, duplicate := seen[value]; duplicate {
			clear(value[:])
			return false
		}
		seen[value] = struct{}{}
		clear(value[:])
	}
	return true
}

func containsRecoveryHash(hashes [][]byte, target []byte) bool {
	if len(target) != sha256.Size {
		return false
	}
	for _, hash := range hashes {
		if len(hash) != sha256.Size {
			return false
		}
		var different byte
		for index := range hash {
			different |= hash[index] ^ target[index]
		}
		if different == 0 {
			return true
		}
	}
	return false
}

func cloneRecoveryHashes(hashes [][]byte) [][]byte {
	cloned := make([][]byte, len(hashes))
	for index := range hashes {
		cloned[index] = append([]byte(nil), hashes[index]...)
	}
	return cloned
}

type recoveryCodesReplay struct {
	Codes []string `json:"codes"`
}

func encodeRecoveryCodes(codes RecoveryCodes) ([]byte, error) {
	if !validRecoveryCodes(codes.Codes) {
		return nil, errStrongAuthDependency
	}
	body, err := json.Marshal(recoveryCodesReplay(codes))
	if err != nil || len(body) == 0 || len(body) > maximumRecoveryReplay {
		clear(body)
		return nil, errStrongAuthDependency
	}
	return body, nil
}

func decodeRecoveryCodes(body []byte) (RecoveryCodes, error) {
	if len(body) == 0 || len(body) > maximumRecoveryReplay {
		return RecoveryCodes{}, errStrongAuthDependency
	}
	var replay recoveryCodesReplay
	if err := json.Unmarshal(body, &replay); err != nil || !validRecoveryCodes(replay.Codes) {
		return RecoveryCodes{}, errStrongAuthDependency
	}
	return RecoveryCodes{Codes: append([]string(nil), replay.Codes...)}, nil
}

func validRecoveryCodes(codes []string) bool {
	if len(codes) != recoveryCodeCount {
		return false
	}
	seen := make(map[[sha256.Size]byte]struct{}, recoveryCodeCount)
	for _, code := range codes {
		digest, valid := recoveryCodeDigest(code)
		if !valid {
			return false
		}
		if _, duplicate := seen[digest]; duplicate {
			clear(digest[:])
			return false
		}
		seen[digest] = struct{}{}
		clear(digest[:])
	}
	return true
}

// Format redacts recovery-code rotation input.
func (RotateRecoveryCodesCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.RotateRecoveryCodesCommand([REDACTED])"))
}

// LogValue redacts recovery-code rotation input from structured logs.
func (RotateRecoveryCodesCommand) LogValue() slog.Value {
	return slog.StringValue("identity.RotateRecoveryCodesCommand([REDACTED])")
}

// MarshalJSON forbids serialization of recovery-code rotation input.
func (RotateRecoveryCodesCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: recovery rotation command serialization forbidden")
}

// Format redacts plaintext recovery codes.
func (RecoveryCodes) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.RecoveryCodes([REDACTED])"))
}

// LogValue redacts plaintext recovery codes from structured logs.
func (RecoveryCodes) LogValue() slog.Value {
	return slog.StringValue("identity.RecoveryCodes([REDACTED])")
}

// MarshalJSON forbids generic serialization of plaintext recovery codes.
func (RecoveryCodes) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: recovery codes serialization forbidden")
}

// Format redacts recovery-code consumption input.
func (ConsumeRecoveryCodeCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.ConsumeRecoveryCodeCommand([REDACTED])"))
}

// LogValue redacts recovery-code consumption input from structured logs.
func (ConsumeRecoveryCodeCommand) LogValue() slog.Value {
	return slog.StringValue("identity.ConsumeRecoveryCodeCommand([REDACTED])")
}

// MarshalJSON forbids serialization of recovery-code consumption input.
func (ConsumeRecoveryCodeCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: recovery consumption command serialization forbidden")
}
