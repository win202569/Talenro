package identity

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/ratelimit"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/store"
)

// AccountAuthority is the PostgreSQL-backed identity established by an account access token.
type AccountAuthority struct {
	PrincipalID PrincipalID
	SessionID   SessionID
}

// AccountAuthenticator validates only account-domain access tokens.
type AccountAuthenticator interface {
	Authenticate(context.Context, secret.Bytes) (AccountAuthority, error)
}

// Format redacts both opaque authority identifiers.
func (AccountAuthority) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.AccountAuthority([REDACTED])"))
}

// LogValue prevents structured logging from reflecting authority identifiers.
func (AccountAuthority) LogValue() slog.Value {
	return slog.StringValue("identity.AccountAuthority([REDACTED])")
}

// MarshalJSON forbids direct authority serialization.
func (AccountAuthority) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: account authority serialization forbidden")
}

type sessionTransaction interface {
	Transaction
	FindAccountAccessToken(context.Context, []byte) (store.FindAccountAccessTokenRow, bool, error)
	DiscoverRefreshToken(context.Context, []byte) (store.DiscoverRefreshTokenRow, bool, error)
	LockSessionRefreshTokens(context.Context, uuid.UUID) ([]store.IdentityAccountRefreshToken, error)
	ListSessionRefreshTokens(context.Context, uuid.UUID) ([]store.IdentityAccountRefreshToken, error)
	MarkAccountRefreshUsed(context.Context, store.MarkAccountRefreshUsedParams) (bool, error)
	RotateAccountSessionAccess(context.Context, store.RotateAccountSessionAccessParams) (store.IdentityAccountSession, bool, error)
	RevokeAccountRefreshTokens(context.Context, store.RevokeAccountRefreshTokensParams) (int64, error)
	MarkAccountSessionCompromised(context.Context, store.MarkAccountSessionCompromisedParams) (int64, error)
	RevokePrincipalAccountSessions(context.Context, store.RevokePrincipalAccountSessionsParams) ([]uuid.UUID, error)
	UpdatePasswordCredential(context.Context, store.UpdatePasswordCredentialParams) (int64, error)
}

type accountRefreshAuthority struct {
	TokenHash               []byte
	SessionID               uuid.UUID
	PreviousTokenHash       []byte
	RefreshState            string
	IssuedAt                time.Time
	IdleExpiresAt           time.Time
	AbsoluteExpiresAt       time.Time
	UsedAt                  sql.NullTime
	RevokedAt               sql.NullTime
	SessionState            string
	SessionStateVersion     int64
	ClientSigningPublicKey  []byte
	PrincipalID             uuid.UUID
	AccountState            string
	LockedFamilyTokenHashes [][]byte
}

const (
	accountRotationProofPrefix       = "TALENRO-ACCOUNT-ROTATION-V1\x00"
	accountChallengeContextPrefix    = "TALENRO-ACCOUNT-CHALLENGE-CONTEXT-V1\x00"
	accountRefreshReplayResponseBody = `{"authenticated":false}`
)

var errSessionIdempotencyPreflight = errors.New("identity: session idempotency preflight")

// AccountRotationProofInput contains every field bound by the client session key.
type AccountRotationProofInput struct {
	ProtocolVersion string
	Challenge       [32]byte
	SessionID       uuid.UUID
	Operation       string
	Audience        string
	RequestNonce    [32]byte
}

// Format redacts the proof transcript fields.
func (AccountRotationProofInput) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.AccountRotationProofInput([REDACTED])"))
}

// LogValue returns a fixed value rather than reflecting proof material.
func (AccountRotationProofInput) LogValue() slog.Value {
	return slog.StringValue("identity.AccountRotationProofInput([REDACTED])")
}

// MarshalJSON forbids direct transcript serialization.
func (AccountRotationProofInput) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: rotation proof serialization forbidden")
}

// AccountRotationProofBytes returns the fixed, unambiguous account rotation transcript.
func AccountRotationProofBytes(input AccountRotationProofInput) []byte {
	if input.ProtocolVersion == "" || len(input.ProtocolVersion) > 64 || input.Challenge == [32]byte{} || input.SessionID == uuid.Nil ||
		input.Operation == "" || len(input.Operation) > 64 || input.Audience == "" || len(input.Audience) > 2048 || input.RequestNonce == [32]byte{} {
		return nil
	}
	result := make([]byte, 0, len(accountRotationProofPrefix)+4*3+len(input.ProtocolVersion)+len(input.Operation)+len(input.Audience)+32+16+32)
	result = append(result, accountRotationProofPrefix...)
	result = appendRotationString(result, input.ProtocolVersion)
	result = append(result, input.Challenge[:]...)
	result = append(result, input.SessionID[:]...)
	result = appendRotationString(result, input.Operation)
	result = appendRotationString(result, input.Audience)
	result = append(result, input.RequestNonce[:]...)
	return result
}

func appendRotationString(target []byte, value string) []byte {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value))) // #nosec G115 -- every string is explicitly bounded by AccountRotationProofBytes.
	target = append(target, length[:]...)
	clear(length[:])
	return append(target, value...)
}

var _ AccountAuthenticator = (*Service)(nil)
var _ Application = (*Service)(nil)

// CreateSession authenticates the Task 10 password proof and creates an isolated account session.
func (service *Service) CreateSession(ctx context.Context, command CreateSessionCommand) (SessionTokens, error) {
	if !validTask10Service(service) || nilIdentityValue(ctx) || command.Method != SessionPassword || command.Email == "" ||
		!validPasswordSecret(command.Password) || command.WebAuthnCeremonyID != "" || len(command.WebAuthnResponse) != 0 ||
		command.ClientSigningPublicKey == [32]byte{} || !validIdempotencyKeyForApplication(command.IdempotencyKey) {
		if validService(service) && nilIdentityValue(service.challenges) {
			return SessionTokens{}, dependencyUnavailable()
		}
		return SessionTokens{}, malformedRequest()
	}
	canonicalEmail, emailErr := CanonicalizeEmail(command.Email)
	if emailErr != nil {
		return SessionTokens{}, malformedRequest()
	}
	emailBytes := canonicalEmail.Bytes()
	defer clear(emailBytes)
	password := command.Password.Copy()
	defer clear(password)
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	now, ok := service.now()
	if !ok {
		return SessionTokens{}, dependencyUnavailable()
	}
	subject, digestErr := ratelimit.SubjectDigest(service.rateLimitKey, ratelimit.Login, now, service.security.LoginRateLimit, string(emailBytes))
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
	canonicalRequest, requestErr := privateCanonicalRequest(service.protector, "create_account_session", emailBytes, password, command.ClientSigningPublicKey[:])
	if requestErr != nil {
		return SessionTokens{}, dependencyUnavailable()
	}
	defer clear(canonicalRequest)
	var result SessionTokens
	err := service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, base Transaction) error {
		transaction, typed := base.(sessionTransaction)
		if !typed || nilIdentityValue(transaction) {
			return dependencyUnavailable()
		}
		lookupDigest := service.protector.LookupDigest(emailFieldDomain, emailBytes)
		if lookupDigest == [32]byte{} {
			return dependencyUnavailable()
		}
		identity, identityFound, findErr := transaction.FindIdentityByLookupDigestRead(transactionContext, lookupDigest[:])
		if findErr != nil {
			return dependencyUnavailable()
		}
		principalID := uuid.Nil
		if identityFound {
			principalID = identity.PrincipalID
		}
		credentialRow, credentialFound, credentialErr := transaction.GetPasswordCredential(transactionContext, principalID)
		if credentialErr != nil {
			return dependencyUnavailable()
		}
		_, _, account, accountFound, accountErr := lockPrincipalRefreshAuthority(transactionContext, transaction, principalID)
		if accountErr != nil {
			return dependencyUnavailable()
		}
		credential := DummyCredential()
		storedCredentialValid := false
		if identityFound && accountFound && credentialFound && identity.PrincipalID == account.ID && credentialRow.PrincipalID == account.ID {
			parsed, parseErr := passwordCredentialFromStore(credentialRow)
			if parseErr != nil {
				if !service.performDummyPasswordWork() {
					return dependencyUnavailable()
				}
				return dependencyUnavailable()
			}
			credential = parsed
			storedCredentialValid = true
		}
		match, needsUpgrade := VerifyPasswordWithDeriver(password, credential, CurrentPasswordPolicy(), service.derive)
		if !storedCredentialValid || account.State != "active" || !match {
			return authenticationFailed()
		}
		scope, scopeErr := idempotency.AuthenticatedScope(account.ID, "account_session", "create_account_session")
		if scopeErr != nil {
			return dependencyUnavailable()
		}
		record, outcome, beginErr := transaction.BeginIdempotency(transactionContext, scope, command.IdempotencyKey, canonicalRequest, now, now.Add(ordinaryIdempotencyRetention))
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
		if needsUpgrade {
			material, hashErr := service.hashNewPassword(password)
			if hashErr != nil {
				return dependencyUnavailable()
			}
			defer material.clear()
			policyVersion, memoryKiB, timeCost, parallelism, policyOK := passwordPolicyInt32(material.policy)
			if !policyOK {
				return dependencyUnavailable()
			}
			updated, updateErr := transaction.UpdatePasswordCredential(transactionContext, store.UpdatePasswordCredentialParams{
				PrincipalID: account.ID, PolicyVersion: policyVersion, MemoryKib: memoryKiB,
				TimeCost: timeCost, Parallelism: parallelism,
				Salt: append([]byte(nil), material.salt...), PasswordHash: append([]byte(nil), material.hash...), UpdatedAt: now,
			})
			if updateErr != nil || updated != 1 {
				return dependencyUnavailable()
			}
		}
		created, createErr := service.newSessionTokens(now)
		if createErr != nil {
			return dependencyUnavailable()
		}
		result = created.tokens
		if createErr = transaction.CreateAccountSession(transactionContext, store.CreateAccountSessionParams{
			ID: created.sessionID, PrincipalID: account.ID, ClientSigningPublicKey: append([]byte(nil), command.ClientSigningPublicKey[:]...),
			AccessTokenHash: append([]byte(nil), created.accessDigest[:]...), AccessExpiresAt: result.AccessExpiresAt,
			AbsoluteExpiresAt: result.RefreshAbsoluteExpiresAt, CreatedAt: now,
		}); createErr != nil {
			return dependencyUnavailable()
		}
		if createErr = transaction.InsertAccountRefreshToken(transactionContext, store.InsertAccountRefreshTokenParams{
			TokenHash: append([]byte(nil), created.refreshDigest[:]...), SessionID: created.sessionID, IssuedAt: now,
			IdleExpiresAt: result.RefreshIdleExpiresAt, AbsoluteExpiresAt: result.RefreshAbsoluteExpiresAt,
		}); createErr != nil {
			return dependencyUnavailable()
		}
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

// Authenticate resolves a non-expired account-domain access token through PostgreSQL only.
func (service *Service) Authenticate(ctx context.Context, accessToken secret.Bytes) (AccountAuthority, error) {
	if !validAccessAuthenticator(service) || nilIdentityValue(ctx) {
		return AccountAuthority{}, dependencyUnavailable()
	}
	digest := securitykit.DigestToken(securitykit.AccountAccessToken, accessToken)
	if digest == [32]byte{} {
		return AccountAuthority{}, authenticationFailed()
	}
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	var authority AccountAuthority
	err := service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, base Transaction) error {
		transaction, typed := base.(sessionTransaction)
		if !typed || nilIdentityValue(transaction) {
			return dependencyUnavailable()
		}
		row, found, findErr := transaction.FindAccountAccessToken(transactionContext, digest[:])
		if findErr != nil {
			return dependencyUnavailable()
		}
		now, ok := service.now()
		if !ok {
			return dependencyUnavailable()
		}
		if !found || row.ID == uuid.Nil || row.PrincipalID == uuid.Nil || row.AccountState != "active" ||
			(row.State != "active" && row.State != "review_required") || !row.AccessExpiresAt.After(now) || !row.AbsoluteExpiresAt.After(now) {
			return authenticationFailed()
		}
		authority = AccountAuthority{PrincipalID: PrincipalID(row.PrincipalID.String()), SessionID: SessionID(row.ID.String())}
		return nil
	})
	if err != nil {
		return AccountAuthority{}, mapApplicationError(operationContext, err)
	}
	return authority, nil
}

// CreateSessionChallenge authenticates PostgreSQL refresh authority before creating Redis state.
func (service *Service) CreateSessionChallenge(ctx context.Context, command CreateSessionChallengeCommand) (SessionChallenge, error) {
	if !validTask10Service(service) || nilIdentityValue(ctx) || command.RequestNonce == [32]byte{} || !validIdempotencyKeyForApplication(command.IdempotencyKey) {
		if validService(service) && nilIdentityValue(service.challenges) {
			return SessionChallenge{}, dependencyUnavailable()
		}
		return SessionChallenge{}, malformedRequest()
	}
	refreshDigest := securitykit.DigestToken(securitykit.AccountRefreshToken, command.RefreshToken)
	if refreshDigest == [32]byte{} {
		return SessionChallenge{}, authenticationFailed()
	}
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	now, ok := service.now()
	if !ok {
		return SessionChallenge{}, dependencyUnavailable()
	}
	canonicalRequest, requestErr := privateCanonicalRequest(
		service.protector, "create_account_session_challenge", refreshDigest[:], command.RequestNonce[:],
	)
	if requestErr != nil {
		return SessionChallenge{}, dependencyUnavailable()
	}
	defer clear(canonicalRequest)
	var result SessionChallenge
	preflightReplay := false
	err := service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, base Transaction) error {
		transaction, typed := base.(sessionTransaction)
		if !typed || nilIdentityValue(transaction) {
			return dependencyUnavailable()
		}
		row, found, findErr := lockSessionRefreshAuthority(transactionContext, transaction, refreshDigest[:])
		if findErr != nil {
			return dependencyUnavailable()
		}
		if !found || !validRefreshAuthority(row, now, false) {
			return authenticationFailed()
		}
		scope, scopeErr := idempotency.AuthenticatedScope(row.PrincipalID, "account_refresh", "create_account_session_challenge")
		if scopeErr != nil {
			return dependencyUnavailable()
		}
		idempotencyRecord, outcome, beginErr := transaction.BeginIdempotency(
			transactionContext, scope, command.IdempotencyKey, canonicalRequest, now, now.Add(securityIdempotencyRetention),
		)
		if beginErr != nil {
			return dependencyUnavailable()
		}
		if outcome == idempotency.Started {
			return errSessionIdempotencyPreflight
		}
		body, replayed, replayErr := privateIdempotencyOutcome(outcome, idempotencyRecord, 201)
		if replayErr != nil {
			return replayErr
		}
		if !replayed {
			return dependencyUnavailable()
		}
		defer clear(body)
		decoded, decodeErr := decodeSessionChallenge(body)
		if decodeErr != nil {
			return dependencyUnavailable()
		}
		result = decoded
		preflightReplay = true
		return nil
	})
	if err == nil && preflightReplay {
		return result, nil
	}
	if !errors.Is(err, errSessionIdempotencyPreflight) {
		return SessionChallenge{}, mapApplicationError(operationContext, err)
	}
	subject, subjectErr := ratelimit.SubjectDigest(service.rateLimitKey, ratelimit.Challenge, now, service.security.ChallengeRateLimit, hex.EncodeToString(refreshDigest[:]))
	if subjectErr != nil {
		return SessionChallenge{}, dependencyUnavailable()
	}
	allowed, limitErr := service.rateLimiter.Allow(operationContext, ratelimit.Challenge, subject, service.security.ChallengeRateLimit)
	if limitErr != nil {
		return SessionChallenge{}, dependencyUnavailable()
	}
	if !allowed {
		return SessionChallenge{}, apierrors.NewRetryAfter(apierrors.RateLimited, apierrors.Retry, time.Second)
	}
	challengeID := service.newUUID()
	if challengeID == uuid.Nil {
		return SessionChallenge{}, dependencyUnavailable()
	}
	var challenge [32]byte
	count, randomErr := io.ReadFull(&boundedRandomSource{source: service.random}, challenge[:])
	if randomErr != nil || count != len(challenge) || challenge == [32]byte{} {
		clear(challenge[:])
		return SessionChallenge{}, dependencyUnavailable()
	}
	expiresAt := now.Add(accountChallengeTTL)
	contextDigest := accountChallengeContextDigest(refreshDigest, command.RequestNonce)
	if contextDigest == [32]byte{} {
		clear(challenge[:])
		return SessionChallenge{}, dependencyUnavailable()
	}
	record := ChallengeRecord{
		ChallengeID: challengeID.String(), ProtocolVersion: accountRotationProtocolVersion, Operation: accountRotationOperation,
		Challenge: challenge, ContextDigest: contextDigest, ExpiresAt: expiresAt,
	}
	if createErr := service.challenges.Create(operationContext, record, accountChallengeTTL); createErr != nil {
		clear(challenge[:])
		return SessionChallenge{}, dependencyUnavailable()
	}
	result = SessionChallenge{ChallengeID: challengeID.String(), Challenge: challenge, ExpiresAt: expiresAt}
	body, encodeErr := encodeSessionChallenge(result)
	if encodeErr != nil {
		clear(result.Challenge[:])
		return SessionChallenge{}, dependencyUnavailable()
	}
	defer clear(body)
	err = service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, base Transaction) error {
		transaction, typed := base.(sessionTransaction)
		if !typed || nilIdentityValue(transaction) {
			return dependencyUnavailable()
		}
		row, found, findErr := lockSessionRefreshAuthority(transactionContext, transaction, refreshDigest[:])
		if findErr != nil {
			return dependencyUnavailable()
		}
		if !found || !validRefreshAuthority(row, now, false) {
			return authenticationFailed()
		}
		scope, scopeErr := idempotency.AuthenticatedScope(row.PrincipalID, "account_refresh", "create_account_session_challenge")
		if scopeErr != nil {
			return dependencyUnavailable()
		}
		idempotencyRecord, outcome, beginErr := transaction.BeginIdempotency(
			transactionContext, scope, command.IdempotencyKey, canonicalRequest, now, now.Add(securityIdempotencyRetention),
		)
		if beginErr != nil {
			return dependencyUnavailable()
		}
		if outcome != idempotency.Started {
			replayBody, replayed, replayErr := privateIdempotencyOutcome(outcome, idempotencyRecord, 201)
			if replayErr != nil {
				return replayErr
			}
			if !replayed {
				return dependencyUnavailable()
			}
			defer clear(replayBody)
			decoded, decodeErr := decodeSessionChallenge(replayBody)
			if decodeErr != nil {
				return dependencyUnavailable()
			}
			clear(result.Challenge[:])
			result = decoded
			return nil
		}
		if completeErr := transaction.CompleteIdempotency(transactionContext, idempotencyRecord, 201, body); completeErr != nil {
			return dependencyUnavailable()
		}
		return nil
	})
	if err != nil {
		clear(result.Challenge[:])
		return SessionChallenge{}, mapApplicationError(operationContext, err)
	}
	return result, nil
}

// RotateSession classifies replay before consuming Redis, then performs one PostgreSQL refresh transition.
func (service *Service) RotateSession(ctx context.Context, command RotateSessionCommand) (SessionTokens, error) {
	if !validTask10Service(service) || nilIdentityValue(ctx) || !validChallengeID(command.ChallengeID) || command.RequestNonce == [32]byte{} ||
		command.Signature == [64]byte{} || !validIdempotencyKeyForApplication(command.IdempotencyKey) {
		if validService(service) && nilIdentityValue(service.challenges) {
			return SessionTokens{}, dependencyUnavailable()
		}
		return SessionTokens{}, malformedRequest()
	}
	refreshDigest := securitykit.DigestToken(securitykit.AccountRefreshToken, command.RefreshToken)
	if refreshDigest == [32]byte{} {
		return SessionTokens{}, authenticationFailed()
	}
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	now, ok := service.now()
	if !ok {
		return SessionTokens{}, dependencyUnavailable()
	}
	canonicalRequest, requestErr := privateCanonicalRequest(
		service.protector, "rotate_account_token", refreshDigest[:], []byte(command.ChallengeID), command.RequestNonce[:], command.Signature[:],
	)
	if requestErr != nil {
		return SessionTokens{}, dependencyUnavailable()
	}
	defer clear(canonicalRequest)
	var authority accountRefreshAuthority
	var result SessionTokens
	postCommitAuthenticationFailure := false
	preflightReplay := false
	err := service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, base Transaction) error {
		transaction, typed := base.(sessionTransaction)
		if !typed || nilIdentityValue(transaction) {
			return dependencyUnavailable()
		}
		row, found, findErr := lockSessionRefreshAuthority(transactionContext, transaction, refreshDigest[:])
		if findErr != nil {
			return dependencyUnavailable()
		}
		if !found || !validRefreshAuthority(row, now, true) {
			return authenticationFailed()
		}
		scope, scopeErr := idempotency.AuthenticatedScope(row.PrincipalID, "account_refresh", "rotate_account_token")
		if scopeErr != nil {
			return dependencyUnavailable()
		}
		idempotencyRecord, outcome, beginErr := transaction.BeginIdempotency(
			transactionContext, scope, command.IdempotencyKey, canonicalRequest, now, now.Add(securityIdempotencyRetention),
		)
		if beginErr != nil {
			return dependencyUnavailable()
		}
		if outcome == idempotency.Started {
			if row.RefreshState == "used" {
				if compromiseErr := service.compromiseUsedRefresh(transactionContext, transaction, row, idempotencyRecord, now); compromiseErr != nil {
					return compromiseErr
				}
				postCommitAuthenticationFailure = true
				preflightReplay = true
				return nil
			}
			authority = cloneRefreshAuthority(row)
			return errSessionIdempotencyPreflight
		}
		if outcome == idempotency.Replay && idempotencyRecord.ResponseStatus() == 401 {
			body, owned := idempotencyRecord.TakeResponseBody()
			defer clear(body)
			if !owned || string(body) != accountRefreshReplayResponseBody {
				return dependencyUnavailable()
			}
			postCommitAuthenticationFailure = true
			preflightReplay = true
			return nil
		}
		body, replayed, replayErr := privateIdempotencyOutcome(outcome, idempotencyRecord, 200)
		if replayErr != nil {
			return replayErr
		}
		if !replayed {
			return dependencyUnavailable()
		}
		defer clear(body)
		decoded, decodeErr := decodeSessionTokens(body)
		if decodeErr != nil {
			return dependencyUnavailable()
		}
		result = decoded
		preflightReplay = true
		return nil
	})
	if err == nil && preflightReplay {
		if postCommitAuthenticationFailure {
			return SessionTokens{}, authenticationFailed()
		}
		return result, nil
	}
	if !errors.Is(err, errSessionIdempotencyPreflight) {
		return SessionTokens{}, mapApplicationError(operationContext, err)
	}
	defer clear(authority.ClientSigningPublicKey)
	contextDigest := accountChallengeContextDigest(refreshDigest, command.RequestNonce)
	record, consumeErr := service.challenges.Consume(operationContext, command.ChallengeID, contextDigest)
	if consumeErr != nil {
		ambiguous := !errors.Is(consumeErr, ErrChallengeNotFound) && !errors.Is(consumeErr, ErrInvalidChallenge)
		reclassified, resolved, authenticationFailure, reclassifyErr := service.reclassifyRotationAfterChallengeFailure(
			operationContext, command, canonicalRequest, refreshDigest, now, ambiguous,
		)
		if reclassifyErr != nil {
			return SessionTokens{}, reclassifyErr
		}
		if authenticationFailure {
			return SessionTokens{}, authenticationFailed()
		}
		if resolved {
			return reclassified, nil
		}
		return SessionTokens{}, dependencyUnavailable()
	}
	if !record.ExpiresAt.After(now) || record.ProtocolVersion != accountRotationProtocolVersion || record.Operation != accountRotationOperation {
		return SessionTokens{}, authenticationFailed()
	}
	proof := AccountRotationProofBytes(AccountRotationProofInput{
		ProtocolVersion: accountRotationProtocolVersion, Challenge: record.Challenge, SessionID: authority.SessionID,
		Operation: accountRotationOperation, Audience: service.security.PublicBaseURL, RequestNonce: command.RequestNonce,
	})
	if len(authority.ClientSigningPublicKey) != ed25519.PublicKeySize || len(proof) == 0 || !ed25519.Verify(ed25519.PublicKey(authority.ClientSigningPublicKey), proof, command.Signature[:]) {
		clear(proof)
		return SessionTokens{}, authenticationFailed()
	}
	clear(proof)
	err = service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, base Transaction) error {
		transaction, typed := base.(sessionTransaction)
		if !typed || nilIdentityValue(transaction) {
			return dependencyUnavailable()
		}
		row, found, findErr := lockSessionRefreshAuthority(transactionContext, transaction, refreshDigest[:])
		if findErr != nil {
			return dependencyUnavailable()
		}
		if !found || !sameRefreshAuthority(authority, row) || !validRefreshAuthority(row, now, true) {
			return authenticationFailed()
		}
		scope, scopeErr := idempotency.AuthenticatedScope(row.PrincipalID, "account_refresh", "rotate_account_token")
		if scopeErr != nil {
			return dependencyUnavailable()
		}
		idempotencyRecord, outcome, beginErr := transaction.BeginIdempotency(transactionContext, scope, command.IdempotencyKey, canonicalRequest, now, now.Add(securityIdempotencyRetention))
		if beginErr != nil {
			return dependencyUnavailable()
		}
		if outcome != idempotency.Started {
			if outcome == idempotency.Replay && idempotencyRecord.ResponseStatus() == 401 {
				body, owned := idempotencyRecord.TakeResponseBody()
				defer clear(body)
				if owned && string(body) == accountRefreshReplayResponseBody {
					postCommitAuthenticationFailure = true
					return nil
				}
				return dependencyUnavailable()
			}
			body, replayed, replayErr := privateIdempotencyOutcome(outcome, idempotencyRecord, 200)
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
		switch row.RefreshState {
		case "active":
			if row.SessionState != "active" || !row.IdleExpiresAt.After(now) || !row.AbsoluteExpiresAt.After(now) {
				return authenticationFailed()
			}
			marked, markErr := transaction.MarkAccountRefreshUsed(transactionContext, store.MarkAccountRefreshUsedParams{
				TokenHash: append([]byte(nil), refreshDigest[:]...), UsedAt: sql.NullTime{Time: now, Valid: true},
			})
			if markErr != nil {
				return dependencyUnavailable()
			}
			if !marked {
				return authenticationFailed()
			}
			created, createErr := service.newRotatedSessionTokens(now, row.AbsoluteExpiresAt)
			if createErr != nil {
				return dependencyUnavailable()
			}
			result = created.tokens
			_, rotated, rotateErr := transaction.RotateAccountSessionAccess(transactionContext, store.RotateAccountSessionAccessParams{
				ID: row.SessionID, AccessTokenHash: append([]byte(nil), created.accessDigest[:]...), AccessExpiresAt: result.AccessExpiresAt, UpdatedAt: now,
			})
			if rotateErr != nil || !rotated {
				return dependencyUnavailable()
			}
			if insertErr := transaction.InsertAccountRefreshToken(transactionContext, store.InsertAccountRefreshTokenParams{
				TokenHash: append([]byte(nil), created.refreshDigest[:]...), SessionID: row.SessionID,
				PreviousTokenHash: append([]byte(nil), refreshDigest[:]...), IssuedAt: now,
				IdleExpiresAt: result.RefreshIdleExpiresAt, AbsoluteExpiresAt: result.RefreshAbsoluteExpiresAt,
			}); insertErr != nil {
				return dependencyUnavailable()
			}
			body, encodeErr := encodeSessionTokens(result)
			if encodeErr != nil {
				return dependencyUnavailable()
			}
			defer clear(body)
			if completeErr := transaction.CompleteIdempotency(transactionContext, idempotencyRecord, 200, body); completeErr != nil {
				return dependencyUnavailable()
			}
			return nil
		case "used":
			if compromiseErr := service.compromiseUsedRefresh(transactionContext, transaction, row, idempotencyRecord, now); compromiseErr != nil {
				return compromiseErr
			}
			postCommitAuthenticationFailure = true
			return nil
		default:
			return authenticationFailed()
		}
	})
	if err != nil {
		return SessionTokens{}, mapApplicationError(operationContext, err)
	}
	if postCommitAuthenticationFailure {
		return SessionTokens{}, authenticationFailed()
	}
	return result, nil
}

func (service *Service) compromiseUsedRefresh(ctx context.Context, transaction sessionTransaction, row accountRefreshAuthority, record idempotency.Record, now time.Time) error {
	changed, compromiseErr := transaction.MarkAccountSessionCompromised(ctx, store.MarkAccountSessionCompromisedParams{ID: row.SessionID, UpdatedAt: now})
	if compromiseErr != nil || changed != 1 {
		return dependencyUnavailable()
	}
	if _, revokeErr := transaction.RevokeAccountRefreshTokens(ctx, store.RevokeAccountRefreshTokensParams{
		TokenHashes: cloneTokenHashes(row.LockedFamilyTokenHashes), RevokedAt: sql.NullTime{Time: now, Valid: true},
	}); revokeErr != nil {
		return dependencyUnavailable()
	}
	eventID := service.newUUID()
	if eventID == uuid.Nil || transaction.InsertSecurityEvent(ctx, store.InsertSecurityEventParams{
		ID: eventID, PrincipalID: uuid.NullUUID{UUID: row.PrincipalID, Valid: true}, Category: "account_refresh_replay",
		Fingerprint: "identity.account_refresh_replay", AggregateVersion: row.SessionStateVersion + 1, OccurredAt: now,
	}) != nil {
		return dependencyUnavailable()
	}
	if completeErr := transaction.CompleteIdempotency(ctx, record, 401, []byte(accountRefreshReplayResponseBody)); completeErr != nil {
		return dependencyUnavailable()
	}
	return nil
}

func (service *Service) reclassifyRotationAfterChallengeFailure(
	ctx context.Context,
	command RotateSessionCommand,
	canonicalRequest []byte,
	refreshDigest [32]byte,
	now time.Time,
	ambiguous bool,
) (result SessionTokens, resolved, authenticationFailure bool, resultErr error) {
	err := service.repository.WithinTransaction(ctx, func(transactionContext context.Context, base Transaction) error {
		transaction, typed := base.(sessionTransaction)
		if !typed || nilIdentityValue(transaction) {
			return dependencyUnavailable()
		}
		row, found, findErr := lockSessionRefreshAuthority(transactionContext, transaction, refreshDigest[:])
		if findErr != nil {
			return dependencyUnavailable()
		}
		if !found || !validRefreshAuthority(row, now, true) {
			if ambiguous {
				return dependencyUnavailable()
			}
			return authenticationFailed()
		}
		scope, scopeErr := idempotency.AuthenticatedScope(row.PrincipalID, "account_refresh", "rotate_account_token")
		if scopeErr != nil {
			return dependencyUnavailable()
		}
		idempotencyRecord, outcome, beginErr := transaction.BeginIdempotency(
			transactionContext, scope, command.IdempotencyKey, canonicalRequest, now, now.Add(securityIdempotencyRetention),
		)
		if beginErr != nil {
			return dependencyUnavailable()
		}
		if outcome == idempotency.Started {
			if row.RefreshState == "used" {
				if compromiseErr := service.compromiseUsedRefresh(transactionContext, transaction, row, idempotencyRecord, now); compromiseErr != nil {
					return compromiseErr
				}
				resolved = true
				authenticationFailure = true
				return nil
			}
			return errSessionIdempotencyPreflight
		}
		if ambiguous && outcome != idempotency.Replay {
			return dependencyUnavailable()
		}
		if outcome == idempotency.Replay && idempotencyRecord.ResponseStatus() == 401 {
			body, owned := idempotencyRecord.TakeResponseBody()
			defer clear(body)
			if !owned || string(body) != accountRefreshReplayResponseBody {
				return dependencyUnavailable()
			}
			resolved = true
			authenticationFailure = true
			return nil
		}
		body, replayed, replayErr := privateIdempotencyOutcome(outcome, idempotencyRecord, 200)
		if replayErr != nil {
			return replayErr
		}
		if !replayed {
			return dependencyUnavailable()
		}
		defer clear(body)
		decoded, decodeErr := decodeSessionTokens(body)
		if decodeErr != nil {
			return dependencyUnavailable()
		}
		result = decoded
		resolved = true
		return nil
	})
	if errors.Is(err, errSessionIdempotencyPreflight) {
		if ambiguous {
			return SessionTokens{}, false, false, dependencyUnavailable()
		}
		return SessionTokens{}, false, true, nil
	}
	if err != nil {
		return SessionTokens{}, false, false, mapApplicationError(ctx, err)
	}
	return result, resolved, authenticationFailure, nil
}

// RevokeSessions applies one generated principal/scope transition and its tombstone atomically.
func (service *Service) RevokeSessions(ctx context.Context, command RevokeSessionsCommand) error {
	if !validTask10Service(service) || nilIdentityValue(ctx) || !validIdempotencyKeyForApplication(command.IdempotencyKey) {
		if validService(service) && nilIdentityValue(service.challenges) {
			return dependencyUnavailable()
		}
		return malformedRequest()
	}
	principalID, principalOK := parseCanonicalIdentityUUID(string(command.PrincipalID))
	if !principalOK {
		return malformedRequest()
	}
	sessionID := uuid.Nil
	switch command.Scope {
	case RevokeAllSessions:
		if command.SessionID != "" {
			return malformedRequest()
		}
	case RevokeCurrentSession, RevokeOtherSessions:
		var sessionOK bool
		sessionID, sessionOK = parseCanonicalIdentityUUID(string(command.SessionID))
		if !sessionOK {
			return malformedRequest()
		}
	default:
		return malformedRequest()
	}
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	canonicalRequest, requestErr := privateCanonicalRequest(service.protector, "revoke_account_sessions", []byte(command.Scope), sessionID[:])
	if requestErr != nil {
		return dependencyUnavailable()
	}
	defer clear(canonicalRequest)
	err := service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, base Transaction) error {
		transaction, typed := base.(sessionTransaction)
		if !typed || nilIdentityValue(transaction) {
			return dependencyUnavailable()
		}
		now, ok := service.now()
		if !ok {
			return dependencyUnavailable()
		}
		scope, scopeErr := idempotency.AuthenticatedScope(principalID, "account_session", "revoke_account_sessions")
		if scopeErr != nil {
			return dependencyUnavailable()
		}
		record, outcome, beginErr := transaction.BeginIdempotency(transactionContext, scope, command.IdempotencyKey, canonicalRequest, now, now.Add(securityIdempotencyRetention))
		if beginErr != nil {
			return dependencyUnavailable()
		}
		if outcome != idempotency.Started {
			replayed, replayErr := genericIdempotencyOutcome(outcome, record, 204)
			if replayErr != nil {
				return replayErr
			}
			if replayed {
				return nil
			}
		}
		lockedRefresh, lockedSessions, account, accountFound, lockErr := lockPrincipalRefreshAuthority(transactionContext, transaction, principalID)
		if lockErr != nil {
			return dependencyUnavailable()
		}
		if !accountFound || account.ID != principalID {
			return authenticationFailed()
		}
		targetSessions, selectionOK := selectRevokedSessions(lockedSessions, principalID, command.Scope, sessionID)
		if !selectionOK {
			return authenticationFailed()
		}
		targetIDs := lockedSessionIDs(targetSessions)
		_, revokeErr := transaction.RevokePrincipalAccountSessions(transactionContext, store.RevokePrincipalAccountSessionsParams{
			RevokedAt: now, SessionIds: targetIDs, TokenHashes: refreshTokenHashesForSessions(lockedRefresh, targetIDs),
		})
		if revokeErr != nil {
			return dependencyUnavailable()
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

// ChangePassword verifies one inline password reauthentication and creates a new bound session.
func (service *Service) ChangePassword(ctx context.Context, command ChangePasswordCommand) (SessionTokens, error) {
	if !validTask10Service(service) || nilIdentityValue(ctx) || !validPasswordSecret(command.CurrentPassword) ||
		!validPasswordSecret(command.NewPassword) || command.ClientSigningPublicKey == [32]byte{} ||
		!validIdempotencyKeyForApplication(command.IdempotencyKey) {
		if validService(service) && nilIdentityValue(service.challenges) {
			return SessionTokens{}, dependencyUnavailable()
		}
		return SessionTokens{}, malformedRequest()
	}
	principalID, principalOK := parseCanonicalIdentityUUID(string(command.PrincipalID))
	sessionID, sessionOK := parseCanonicalIdentityUUID(string(command.Reauthentication.SessionID))
	if !principalOK || !sessionOK {
		return SessionTokens{}, malformedRequest()
	}
	if command.Reauthentication.Method != ReauthPassword {
		return SessionTokens{}, actionNotAllowed()
	}
	if !validPasswordSecret(command.Reauthentication.Proof) {
		return SessionTokens{}, malformedRequest()
	}
	currentPassword := command.CurrentPassword.Copy()
	newPassword := command.NewPassword.Copy()
	reauthenticationProof := command.Reauthentication.Proof.Copy()
	defer clear(currentPassword)
	defer clear(newPassword)
	defer clear(reauthenticationProof)
	proofMatches := len(currentPassword) == len(reauthenticationProof) && subtle.ConstantTimeCompare(currentPassword, reauthenticationProof) == 1
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	canonicalRequest, requestErr := privateCanonicalRequest(
		service.protector, "change_password", sessionID[:], currentPassword, newPassword, reauthenticationProof, command.ClientSigningPublicKey[:],
	)
	if requestErr != nil {
		return SessionTokens{}, dependencyUnavailable()
	}
	defer clear(canonicalRequest)
	var result SessionTokens
	err := service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, base Transaction) error {
		transaction, typed := base.(sessionTransaction)
		if !typed || nilIdentityValue(transaction) {
			return dependencyUnavailable()
		}
		now, ok := service.now()
		if !ok {
			return dependencyUnavailable()
		}
		credentialRow, credentialFound, credentialErr := transaction.GetPasswordCredential(transactionContext, principalID)
		if credentialErr != nil {
			return dependencyUnavailable()
		}
		_, lockedSessions, account, accountFound, lockErr := lockPrincipalRefreshAuthority(transactionContext, transaction, principalID)
		if lockErr != nil {
			return dependencyUnavailable()
		}
		session, sessionFound := findLockedAccountSession(lockedSessions, sessionID, principalID)
		if !accountFound || !sessionFound || account.ID != principalID || session.PrincipalID != principalID || account.State != "active" ||
			(session.State != "active" && session.State != "review_required") || !session.AbsoluteExpiresAt.After(now) {
			return authenticationFailed()
		}
		scope, scopeErr := idempotency.AuthenticatedScope(principalID, "password", "change_password")
		if scopeErr != nil {
			return dependencyUnavailable()
		}
		record, outcome, beginErr := transaction.BeginIdempotency(transactionContext, scope, command.IdempotencyKey, canonicalRequest, now, now.Add(securityIdempotencyRetention))
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
		if !credentialFound || credentialRow.PrincipalID != principalID {
			return authenticationFailed()
		}
		credential, parseErr := passwordCredentialFromStore(credentialRow)
		if parseErr != nil {
			if !service.performDummyPasswordWork() {
				return dependencyUnavailable()
			}
			return dependencyUnavailable()
		}
		match, _ := VerifyPasswordWithDeriver(currentPassword, credential, CurrentPasswordPolicy(), service.derive)
		if !match || !proofMatches {
			return authenticationFailed()
		}
		material, hashErr := service.hashNewPassword(newPassword)
		if hashErr != nil {
			return dependencyUnavailable()
		}
		defer material.clear()
		policyVersion, memoryKiB, timeCost, parallelism, policyOK := passwordPolicyInt32(material.policy)
		if !policyOK {
			return dependencyUnavailable()
		}
		updated, updateErr := transaction.UpdatePasswordCredential(transactionContext, store.UpdatePasswordCredentialParams{
			PrincipalID: principalID, PolicyVersion: policyVersion, MemoryKib: memoryKiB, TimeCost: timeCost, Parallelism: parallelism,
			Salt: append([]byte(nil), material.salt...), PasswordHash: append([]byte(nil), material.hash...), UpdatedAt: now,
		})
		if updateErr != nil || updated != 1 {
			return dependencyUnavailable()
		}
		if _, markErr := transaction.MarkPrincipalSessionsReviewRequired(transactionContext, store.MarkPrincipalSessionsReviewRequiredParams{SessionIds: lockedSessionIDs(lockedSessions), UpdatedAt: now}); markErr != nil {
			return dependencyUnavailable()
		}
		created, createErr := service.newSessionTokens(now)
		if createErr != nil {
			return dependencyUnavailable()
		}
		result = created.tokens
		if createErr = transaction.CreateAccountSession(transactionContext, store.CreateAccountSessionParams{
			ID: created.sessionID, PrincipalID: principalID, ClientSigningPublicKey: append([]byte(nil), command.ClientSigningPublicKey[:]...),
			AccessTokenHash: append([]byte(nil), created.accessDigest[:]...), AccessExpiresAt: result.AccessExpiresAt,
			AbsoluteExpiresAt: result.RefreshAbsoluteExpiresAt, CreatedAt: now,
		}); createErr != nil {
			return dependencyUnavailable()
		}
		if createErr = transaction.InsertAccountRefreshToken(transactionContext, store.InsertAccountRefreshTokenParams{
			TokenHash: append([]byte(nil), created.refreshDigest[:]...), SessionID: created.sessionID, IssuedAt: now,
			IdleExpiresAt: result.RefreshIdleExpiresAt, AbsoluteExpiresAt: result.RefreshAbsoluteExpiresAt,
		}); createErr != nil {
			return dependencyUnavailable()
		}
		eventID := service.newUUID()
		if eventID == uuid.Nil || transaction.InsertSecurityEvent(transactionContext, store.InsertSecurityEventParams{
			ID: eventID, PrincipalID: uuid.NullUUID{UUID: principalID, Valid: true}, Category: "password_changed",
			Fingerprint: "identity.password_changed", AggregateVersion: account.StateVersion, OccurredAt: now,
		}) != nil {
			return dependencyUnavailable()
		}
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

type rotatedSessionTokens struct {
	tokens        SessionTokens
	accessDigest  [32]byte
	refreshDigest [32]byte
}

func (service *Service) newRotatedSessionTokens(now, absoluteExpiresAt time.Time) (rotatedSessionTokens, error) {
	if !absoluteExpiresAt.After(now) {
		return rotatedSessionTokens{}, ErrRandomSource
	}
	idleExpiresAt := now.Add(accountRefreshIdleTTL)
	if !idleExpiresAt.Before(absoluteExpiresAt) {
		idleExpiresAt = absoluteExpiresAt.Add(-time.Nanosecond)
	}
	accessExpiresAt := now.Add(accountAccessTTL)
	if !accessExpiresAt.Before(idleExpiresAt) {
		accessExpiresAt = idleExpiresAt.Add(-time.Nanosecond)
	}
	if !accessExpiresAt.After(now) || !idleExpiresAt.After(accessExpiresAt) {
		return rotatedSessionTokens{}, ErrRandomSource
	}
	access, err := securitykit.NewOpaqueToken(service.random)
	if err != nil {
		return rotatedSessionTokens{}, ErrRandomSource
	}
	refresh, err := securitykit.NewOpaqueToken(service.random)
	if err != nil {
		return rotatedSessionTokens{}, ErrRandomSource
	}
	accessDigest := securitykit.DigestToken(securitykit.AccountAccessToken, access)
	refreshDigest := securitykit.DigestToken(securitykit.AccountRefreshToken, refresh)
	if accessDigest == [32]byte{} || refreshDigest == [32]byte{} {
		return rotatedSessionTokens{}, ErrRandomSource
	}
	return rotatedSessionTokens{
		tokens: SessionTokens{
			AccessToken: access, AccessExpiresAt: accessExpiresAt, RefreshToken: refresh,
			RefreshIdleExpiresAt: idleExpiresAt, RefreshAbsoluteExpiresAt: absoluteExpiresAt,
		}, accessDigest: accessDigest, refreshDigest: refreshDigest,
	}, nil
}

func accountChallengeContextDigest(refreshDigest, requestNonce [32]byte) [32]byte {
	if refreshDigest == [32]byte{} || requestNonce == [32]byte{} {
		return [32]byte{}
	}
	material := make([]byte, 0, len(accountChallengeContextPrefix)+64)
	material = append(material, accountChallengeContextPrefix...)
	material = append(material, refreshDigest[:]...)
	material = append(material, requestNonce[:]...)
	digest := sha256.Sum256(material)
	clear(material)
	return digest
}

func encodeSessionChallenge(challenge SessionChallenge) ([]byte, error) {
	id, err := uuid.Parse(challenge.ChallengeID)
	if err != nil || id == uuid.Nil || id.String() != challenge.ChallengeID || challenge.Challenge == [32]byte{} || challenge.ExpiresAt.IsZero() {
		return nil, errors.New("identity: invalid session challenge response")
	}
	body := make([]byte, 16+32+8)
	copy(body[:16], id[:])
	copy(body[16:48], challenge.Challenge[:])
	binary.BigEndian.PutUint64(body[48:], uint64(challenge.ExpiresAt.UnixNano())) // #nosec G115 -- validated application timestamps are positive and bounded.
	return body, nil
}

func decodeSessionChallenge(body []byte) (SessionChallenge, error) {
	if len(body) != 16+32+8 {
		return SessionChallenge{}, errors.New("identity: invalid session challenge replay")
	}
	id, err := uuid.FromBytes(body[:16])
	if err != nil || id == uuid.Nil {
		return SessionChallenge{}, errors.New("identity: invalid session challenge replay")
	}
	var challenge [32]byte
	copy(challenge[:], body[16:48])
	nanoseconds := binary.BigEndian.Uint64(body[48:])
	if challenge == [32]byte{} || nanoseconds == 0 || nanoseconds > uint64(^uint64(0)>>1) {
		clear(challenge[:])
		return SessionChallenge{}, errors.New("identity: invalid session challenge replay")
	}
	return SessionChallenge{
		ChallengeID: id.String(), Challenge: challenge, ExpiresAt: time.Unix(0, int64(nanoseconds)).UTC(), // #nosec G115 -- bounded immediately above.
	}, nil
}

func lockSessionRefreshAuthority(ctx context.Context, transaction sessionTransaction, digest []byte) (accountRefreshAuthority, bool, error) {
	discovered, found, err := transaction.DiscoverRefreshToken(ctx, digest)
	if err != nil || !found {
		return accountRefreshAuthority{}, false, err
	}
	if discovered.SessionID == uuid.Nil || discovered.PrincipalID == uuid.Nil || !bytes.Equal(discovered.TokenHash, digest) {
		return accountRefreshAuthority{}, false, nil
	}
	if _, err = transaction.LockSessionRefreshTokens(ctx, discovered.SessionID); err != nil {
		return accountRefreshAuthority{}, false, err
	}
	lockedRefresh, err := transaction.LockSessionRefreshTokens(ctx, discovered.SessionID)
	if err != nil {
		return accountRefreshAuthority{}, false, err
	}
	refresh, refreshFound := findLockedRefreshToken(lockedRefresh, digest, discovered.SessionID)
	if !refreshFound {
		return accountRefreshAuthority{}, false, nil
	}
	lockedSessions, err := transaction.LockPrincipalAccountSessions(ctx, discovered.PrincipalID)
	if err != nil {
		return accountRefreshAuthority{}, false, err
	}
	session, sessionFound := findLockedAccountSession(lockedSessions, discovered.SessionID, discovered.PrincipalID)
	if !sessionFound {
		return accountRefreshAuthority{}, false, nil
	}
	account, accountFound, err := transaction.GetAccountForUpdate(ctx, discovered.PrincipalID)
	if err != nil {
		return accountRefreshAuthority{}, false, err
	}
	if !accountFound || account.ID != discovered.PrincipalID {
		return accountRefreshAuthority{}, false, nil
	}
	currentRefresh, err := transaction.ListSessionRefreshTokens(ctx, discovered.SessionID)
	if err != nil {
		return accountRefreshAuthority{}, false, err
	}
	if !sameRefreshCollection(lockedRefresh, currentRefresh) {
		return accountRefreshAuthority{}, false, ErrRepository
	}
	return accountRefreshAuthority{
		TokenHash: append([]byte(nil), refresh.TokenHash...), SessionID: refresh.SessionID,
		PreviousTokenHash: append([]byte(nil), refresh.PreviousTokenHash...), RefreshState: refresh.State,
		IssuedAt: refresh.IssuedAt, IdleExpiresAt: refresh.IdleExpiresAt, AbsoluteExpiresAt: refresh.AbsoluteExpiresAt,
		UsedAt: refresh.UsedAt, RevokedAt: refresh.RevokedAt, SessionState: session.State,
		SessionStateVersion: session.StateVersion, ClientSigningPublicKey: append([]byte(nil), session.ClientSigningPublicKey...),
		PrincipalID: session.PrincipalID, AccountState: account.State, LockedFamilyTokenHashes: refreshTokenHashes(lockedRefresh),
	}, true, nil
}

func lockPrincipalRefreshAuthority(ctx context.Context, transaction Transaction, principalID uuid.UUID) ([]store.IdentityAccountRefreshToken, []store.IdentityAccountSession, store.IdentityAccount, bool, error) {
	if _, err := transaction.LockPrincipalRefreshTokens(ctx, principalID); err != nil {
		return nil, nil, store.IdentityAccount{}, false, err
	}
	lockedRefresh, err := transaction.LockPrincipalRefreshTokens(ctx, principalID)
	if err != nil {
		return nil, nil, store.IdentityAccount{}, false, err
	}
	lockedSessions, err := transaction.LockPrincipalAccountSessions(ctx, principalID)
	if err != nil {
		return nil, nil, store.IdentityAccount{}, false, err
	}
	account, accountFound, err := transaction.GetAccountForUpdate(ctx, principalID)
	if err != nil {
		return nil, nil, store.IdentityAccount{}, false, err
	}
	currentRefresh, err := transaction.ListPrincipalRefreshTokens(ctx, principalID)
	if err != nil {
		return nil, nil, store.IdentityAccount{}, false, err
	}
	if !sameRefreshCollection(lockedRefresh, currentRefresh) {
		return nil, nil, store.IdentityAccount{}, false, ErrRepository
	}
	return lockedRefresh, lockedSessions, account, accountFound, nil
}

func findLockedRefreshToken(rows []store.IdentityAccountRefreshToken, digest []byte, sessionID uuid.UUID) (store.IdentityAccountRefreshToken, bool) {
	for _, row := range rows {
		if row.SessionID == sessionID && bytes.Equal(row.TokenHash, digest) {
			return row, true
		}
	}
	return store.IdentityAccountRefreshToken{}, false
}

func sameRefreshCollection(left, right []store.IdentityAccountRefreshToken) bool {
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

func refreshTokenHashes(rows []store.IdentityAccountRefreshToken) [][]byte {
	result := make([][]byte, 0, len(rows))
	for _, row := range rows {
		result = append(result, append([]byte(nil), row.TokenHash...))
	}
	return result
}

func lockedSessionIDs(rows []store.IdentityAccountSession) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.ID)
	}
	return result
}

func selectRevokedSessions(rows []store.IdentityAccountSession, principalID uuid.UUID, scope SessionRevokeScope, current uuid.UUID) ([]store.IdentityAccountSession, bool) {
	if scope != RevokeAllSessions {
		if _, found := findLockedAccountSession(rows, current, principalID); !found {
			return nil, false
		}
	}
	result := make([]store.IdentityAccountSession, 0, len(rows))
	for _, row := range rows {
		switch scope {
		case RevokeAllSessions:
			result = append(result, row)
		case RevokeCurrentSession:
			if row.ID == current {
				result = append(result, row)
			}
		case RevokeOtherSessions:
			if row.ID != current {
				result = append(result, row)
			}
		default:
			return nil, false
		}
	}
	return result, true
}

func refreshTokenHashesForSessions(rows []store.IdentityAccountRefreshToken, sessionIDs []uuid.UUID) [][]byte {
	selected := make(map[uuid.UUID]struct{}, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		selected[sessionID] = struct{}{}
	}
	result := make([][]byte, 0, len(rows))
	for _, row := range rows {
		if _, ok := selected[row.SessionID]; ok {
			result = append(result, append([]byte(nil), row.TokenHash...))
		}
	}
	return result
}

func validRefreshAuthority(row accountRefreshAuthority, now time.Time, allowUsed bool) bool {
	if row.SessionID == uuid.Nil || row.PrincipalID == uuid.Nil || row.AccountState != "active" ||
		(row.SessionState != "active" && (!allowUsed || row.SessionState != "review_required")) || len(row.ClientSigningPublicKey) != ed25519.PublicKeySize ||
		!row.AbsoluteExpiresAt.After(now) || row.IdleExpiresAt.After(row.AbsoluteExpiresAt) {
		return false
	}
	switch row.RefreshState {
	case "active":
		return row.IdleExpiresAt.After(now)
	case "used":
		// A rotated child can outlive the used parent's idle deadline. Keep
		// replay detection active until the family-wide absolute deadline.
		return allowUsed
	default:
		return false
	}
}

func findLockedAccountSession(sessions []store.IdentityAccountSession, sessionID, principalID uuid.UUID) (store.IdentityAccountSession, bool) {
	for _, session := range sessions {
		if session.ID == sessionID && session.PrincipalID == principalID {
			return session, true
		}
	}
	return store.IdentityAccountSession{}, false
}

func cloneRefreshAuthority(row accountRefreshAuthority) accountRefreshAuthority {
	row.TokenHash = append([]byte(nil), row.TokenHash...)
	row.PreviousTokenHash = append([]byte(nil), row.PreviousTokenHash...)
	row.ClientSigningPublicKey = append([]byte(nil), row.ClientSigningPublicKey...)
	row.LockedFamilyTokenHashes = cloneTokenHashes(row.LockedFamilyTokenHashes)
	return row
}

func cloneTokenHashes(rows [][]byte) [][]byte {
	result := make([][]byte, 0, len(rows))
	for _, row := range rows {
		result = append(result, append([]byte(nil), row...))
	}
	return result
}

func sameRefreshAuthority(left, right accountRefreshAuthority) bool {
	return left.SessionID == right.SessionID && left.PrincipalID == right.PrincipalID &&
		len(left.ClientSigningPublicKey) == ed25519.PublicKeySize && len(right.ClientSigningPublicKey) == ed25519.PublicKeySize &&
		subtle.ConstantTimeCompare(left.ClientSigningPublicKey, right.ClientSigningPublicKey) == 1
}

func validTask10Service(service *Service) bool {
	return validAccessAuthenticator(service) && !nilIdentityValue(service.protector) && !nilIdentityValue(service.random) &&
		!nilIdentityValue(service.rateLimiter) && !nilIdentityValue(service.challenges) && service.derive != nil && service.newUUID != nil &&
		service.security.PublicBaseURL != "" && service.security.LoginRateLimit.Limit >= 1 && service.security.LoginRateLimit.Window > 0
}

func validAccessAuthenticator(service *Service) bool {
	return service != nil && !nilIdentityValue(service.repository) && !nilIdentityValue(service.clock) &&
		service.security.RequestDeadline >= 2*time.Second && service.security.RequestDeadline <= 10*time.Second
}
