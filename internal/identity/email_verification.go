package identity

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/ratelimit"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/store"
	"talenro.local/platform/internal/strictjson"
)

const maximumPrivateReplayBytes int64 = 2048

// CreateEmailVerificationDelivery rotates a verification token without exposing account existence.
func (service *Service) CreateEmailVerificationDelivery(ctx context.Context, command CreateEmailVerificationDeliveryCommand) (RegisterAccountResult, error) {
	return service.createDelivery(ctx, command.Email, command.Locale, command.IdempotencyKey, VerifyEmailTemplate)
}

// CreatePasswordResetDelivery rotates a password-reset token without exposing account existence.
func (service *Service) CreatePasswordResetDelivery(ctx context.Context, command CreatePasswordResetDeliveryCommand) (RegisterAccountResult, error) {
	return service.createDelivery(ctx, command.Email, command.Locale, command.IdempotencyKey, ResetPasswordTemplate)
}

func (service *Service) createDelivery(ctx context.Context, email, locale, idempotencyKey string, template TemplateID) (RegisterAccountResult, error) {
	if !validService(service) || nilIdentityValue(ctx) || !validIdentityLocale(locale) || !validTemplateID(template) || !validIdempotencyKeyForApplication(idempotencyKey) {
		return RegisterAccountResult{}, malformedRequest()
	}
	canonicalEmail, err := CanonicalizeEmail(email)
	if err != nil {
		return RegisterAccountResult{}, malformedRequest()
	}
	emailBytes := canonicalEmail.Bytes()
	defer clear(emailBytes)
	now, ok := service.now()
	if !ok {
		return RegisterAccountResult{}, dependencyUnavailable()
	}
	subject, err := ratelimit.SubjectDigest(service.rateLimitKey, ratelimit.Delivery, now, service.security.DeliveryRateLimit, string(emailBytes))
	if err != nil {
		return RegisterAccountResult{}, dependencyUnavailable()
	}
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	allowed, err := service.rateLimiter.Allow(operationContext, ratelimit.Delivery, subject, service.security.DeliveryRateLimit)
	if err != nil {
		return RegisterAccountResult{}, dependencyUnavailable()
	}
	if !allowed {
		return RegisterAccountResult{}, apierrors.NewRetryAfter(apierrors.RateLimited, apierrors.Retry, time.Second)
	}

	operation := idempotency.CreateEmailVerificationDeliveryOperation
	scope := idempotency.AnonymousEmailVerificationDeliveryScope()
	if template == ResetPasswordTemplate {
		operation = idempotency.CreatePasswordResetDeliveryOperation
		scope = idempotency.AnonymousPasswordResetDeliveryScope()
	}
	canonicalRequest, requestErr := privateCanonicalRequest(service.protector, operation, emailBytes, []byte(locale))
	if requestErr != nil {
		return RegisterAccountResult{}, dependencyUnavailable()
	}
	defer clear(canonicalRequest)
	accepted := RegisterAccountResult{Accepted: true}
	err = service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, transaction Transaction) error {
		record, outcome, beginErr := transaction.BeginIdempotency(transactionContext, scope, idempotencyKey, canonicalRequest, now, now.Add(ordinaryIdempotencyRetention))
		if beginErr != nil {
			return dependencyUnavailable()
		}
		replayed, outcomeErr := genericIdempotencyOutcome(outcome, record, 202)
		if outcomeErr != nil || replayed {
			return outcomeErr
		}

		lookupDigest := service.protector.LookupDigest(emailFieldDomain, emailBytes)
		if lookupDigest == [32]byte{} {
			return dependencyUnavailable()
		}
		identity, found, findErr := transaction.FindIdentityByLookupDigest(transactionContext, lookupDigest[:])
		if findErr != nil {
			return dependencyUnavailable()
		}
		if !service.performDummyPasswordWork() {
			return dependencyUnavailable()
		}

		lookupPrincipalID := uuid.Nil
		if found {
			lookupPrincipalID = identity.PrincipalID
		}
		account, accountFound, accountErr := transaction.GetAccountForUpdate(transactionContext, lookupPrincipalID)
		if accountErr != nil {
			return dependencyUnavailable()
		}
		if found && (!accountFound || account.ID != identity.PrincipalID) {
			return dependencyUnavailable()
		}
		eligible := false
		if template == VerifyEmailTemplate {
			eligible = found && service.security.EmailVerification != config.EmailDisabled && account.State == "pending_email" && !identity.VerifiedAt.Valid
		} else {
			credential, credentialFound, credentialErr := transaction.GetPasswordCredential(transactionContext, lookupPrincipalID)
			if credentialErr != nil {
				return dependencyUnavailable()
			}
			if found && credentialFound && credential.PrincipalID != identity.PrincipalID {
				return dependencyUnavailable()
			}
			eligible = found && credentialFound && (account.State == "active" || account.State == "pending_email")
		}

		if eligible {
			delivery, deliveryErr := service.newProtectedDelivery(canonicalEmail, template, locale)
			if deliveryErr != nil {
				return dependencyUnavailable()
			}
			defer delivery.clear()
			if template == VerifyEmailTemplate {
				deliveryKeyVersion, versionOK := protectedKeyVersionInt32(delivery.protected.KeyVersion)
				if !versionOK {
					return dependencyUnavailable()
				}
				updated, updateErr := transaction.ResetEmailVerification(transactionContext, store.ResetEmailVerificationParams{
					PrincipalID: identity.PrincipalID, VerificationTokenHash: append([]byte(nil), delivery.tokenDigest...),
					VerificationExpiresAt:          sql.NullTime{Time: now.Add(emailVerificationTTL), Valid: true},
					VerificationDeliveryID:         uuid.NullUUID{UUID: delivery.id, Valid: true},
					VerificationDeliveryCiphertext: append([]byte(nil), delivery.protected.Ciphertext...),
					VerificationDeliveryKeyVersion: pgtype.Int4{Int32: deliveryKeyVersion, Valid: true}, UpdatedAt: now,
				})
				if updateErr != nil || !updated {
					return dependencyUnavailable()
				}
			} else {
				deliveryKeyVersion, versionOK := protectedKeyVersionInt32(delivery.protected.KeyVersion)
				if !versionOK {
					return dependencyUnavailable()
				}
				updated, updateErr := transaction.SetPasswordReset(transactionContext, store.SetPasswordResetParams{
					PrincipalID: identity.PrincipalID, ResetTokenHash: append([]byte(nil), delivery.tokenDigest...),
					ResetExpiresAt:          sql.NullTime{Time: now.Add(passwordResetTTL), Valid: true},
					ResetDeliveryID:         uuid.NullUUID{UUID: delivery.id, Valid: true},
					ResetDeliveryCiphertext: append([]byte(nil), delivery.protected.Ciphertext...),
					ResetDeliveryKeyVersion: pgtype.Int4{Int32: deliveryKeyVersion, Valid: true}, UpdatedAt: now,
				})
				if updateErr != nil || !updated {
					return dependencyUnavailable()
				}
			}
			envelope, eventErr := service.emailDeliveryEvent(delivery.id, template, locale, idempotencyKey, now)
			if eventErr != nil || transaction.AppendEvent(transactionContext, envelope) != nil {
				return dependencyUnavailable()
			}
		}
		if completeErr := transaction.CompleteIdempotency(transactionContext, record, 202, []byte(acceptedResponseBody)); completeErr != nil {
			return dependencyUnavailable()
		}
		return nil
	})
	if err != nil {
		return RegisterAccountResult{}, mapApplicationError(operationContext, err)
	}
	return accepted, nil
}

// VerifyEmail consumes a single-use token and activates provisional authorizations in the same transaction.
func (service *Service) VerifyEmail(ctx context.Context, command VerifyEmailCommand) error {
	if !validService(service) || nilIdentityValue(ctx) || !validIdempotencyKeyForApplication(command.IdempotencyKey) {
		return malformedRequest()
	}
	digest := securitykit.DigestToken(securitykit.EmailVerificationToken, command.Token)
	if digest == [32]byte{} {
		return authenticationFailed()
	}
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	canonicalRequest, requestErr := privateCanonicalRequest(service.protector, "verify_email", digest[:])
	if requestErr != nil {
		return dependencyUnavailable()
	}
	defer clear(canonicalRequest)
	err := service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, transaction Transaction) error {
		now, ok := service.now()
		if !ok {
			return dependencyUnavailable()
		}
		identity, found, findErr := transaction.GetEmailVerificationForUpdate(transactionContext, digest[:])
		if findErr != nil {
			return dependencyUnavailable()
		}
		if !found || identity.PrincipalID == uuid.Nil {
			return authenticationFailed()
		}
		scope, scopeErr := idempotency.AuthenticatedScope(identity.PrincipalID, "email", "verify_email")
		if scopeErr != nil {
			return dependencyUnavailable()
		}
		record, outcome, beginErr := transaction.BeginIdempotency(transactionContext, scope, command.IdempotencyKey, canonicalRequest, now, now.Add(securityIdempotencyRetention))
		if beginErr != nil {
			return dependencyUnavailable()
		}
		replayed, outcomeErr := genericIdempotencyOutcome(outcome, record, 204)
		if outcomeErr != nil || replayed {
			return outcomeErr
		}
		if identity.VerificationConsumedAt.Valid || !identity.VerificationExpiresAt.Valid || identity.VerificationExpiresAt.Time.Before(now) {
			return authenticationFailed()
		}
		consumed, consumedOK, consumeErr := transaction.ConsumeEmailVerification(transactionContext, store.ConsumeEmailVerificationParams{
			PrincipalID: identity.PrincipalID, VerificationTokenHash: digest[:], VerificationConsumedAt: sql.NullTime{Time: now, Valid: true},
		})
		if consumeErr != nil {
			return dependencyUnavailable()
		}
		if !consumedOK || consumed.PrincipalID != identity.PrincipalID {
			return authenticationFailed()
		}
		rows, clearErr := transaction.ClearPendingEmailDelivery(transactionContext, store.ClearPendingEmailDeliveryParams{
			VerificationDeliveryID: identity.VerificationDeliveryID, UpdatedAt: now,
		})
		if clearErr != nil || rows != 1 {
			return dependencyUnavailable()
		}
		account, activated, activateErr := transaction.ActivateVerifiedAccount(transactionContext, store.ActivateVerifiedAccountParams{ID: identity.PrincipalID, UpdatedAt: now})
		if activateErr != nil {
			return dependencyUnavailable()
		}
		if !activated || account.ID != identity.PrincipalID || account.State != "active" || account.StateVersion < 1 {
			return authenticationFailed()
		}
		if participantErr := service.participant.ActivateVerifiedPrincipal(transactionContext, transaction.DBTX(), PrincipalID(identity.PrincipalID.String()), now); participantErr != nil {
			return dependencyUnavailable()
		}
		envelope, eventErr := service.accountStateEvent(identity.PrincipalID, account.State, account.StateVersion, command.IdempotencyKey, now)
		if eventErr != nil || transaction.AppendEvent(transactionContext, envelope) != nil {
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

// ResetPassword consumes one reset token and creates a replacement account session.
func (service *Service) ResetPassword(ctx context.Context, command ResetPasswordCommand) (SessionTokens, error) {
	if !validService(service) || nilIdentityValue(ctx) || !validPasswordSecret(command.NewPassword) || command.ClientSigningPublicKey == [32]byte{} || !validIdempotencyKeyForApplication(command.IdempotencyKey) {
		return SessionTokens{}, malformedRequest()
	}
	canonicalEmail, emailErr := CanonicalizeEmail(command.Email)
	if emailErr != nil {
		return SessionTokens{}, malformedRequest()
	}
	emailBytes := canonicalEmail.Bytes()
	defer clear(emailBytes)
	resetDigest := passwordResetTokenDigest(command.Token)
	if resetDigest == [32]byte{} {
		if service.performDummyPasswordWork() {
			return SessionTokens{}, authenticationFailed()
		}
		return SessionTokens{}, dependencyUnavailable()
	}
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	newPassword := command.NewPassword.Copy()
	defer clear(newPassword)
	canonicalRequest, requestErr := privateCanonicalRequest(service.protector, "reset_password", emailBytes, resetDigest[:], newPassword, command.ClientSigningPublicKey[:])
	if requestErr != nil {
		return SessionTokens{}, dependencyUnavailable()
	}
	defer clear(canonicalRequest)
	var result SessionTokens
	err := service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, transaction Transaction) error {
		now, ok := service.now()
		if !ok {
			return dependencyUnavailable()
		}
		lookupDigest := service.protector.LookupDigest(emailFieldDomain, emailBytes)
		if lookupDigest == [32]byte{} {
			return dependencyUnavailable()
		}
		identity, identityFound, findErr := transaction.FindIdentityByLookupDigest(transactionContext, lookupDigest[:])
		if findErr != nil {
			return dependencyUnavailable()
		}
		lookupPrincipalID := uuid.Nil
		if identityFound {
			lookupPrincipalID = identity.PrincipalID
		}
		account, accountFound, findErr := transaction.GetAccountForUpdate(transactionContext, lookupPrincipalID)
		if findErr != nil {
			return dependencyUnavailable()
		}
		credential, credentialFound, findErr := transaction.GetPasswordCredential(transactionContext, lookupPrincipalID)
		if findErr != nil {
			return dependencyUnavailable()
		}
		if !identityFound || !accountFound || !credentialFound || account.ID != identity.PrincipalID || credential.PrincipalID != identity.PrincipalID || (account.State != "active" && account.State != "pending_email") {
			if !service.performDummyPasswordWork() {
				return dependencyUnavailable()
			}
			return authenticationFailed()
		}
		scope, scopeErr := idempotency.AuthenticatedScope(identity.PrincipalID, "password", "reset_password")
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
		password := command.NewPassword.Copy()
		defer clear(password)
		material, hashErr := service.hashNewPassword(password)
		if hashErr != nil {
			return dependencyUnavailable()
		}
		defer material.clear()
		policyVersion, memoryKiB, timeCost, parallelism, policyOK := passwordPolicyInt32(material.policy)
		if !policyOK {
			return dependencyUnavailable()
		}
		consumed, consumeErr := transaction.ConsumePasswordReset(transactionContext, store.ConsumePasswordResetParams{
			PrincipalID: identity.PrincipalID, ResetTokenHash: resetDigest[:], PolicyVersion: policyVersion,
			MemoryKib: memoryKiB, TimeCost: timeCost, Parallelism: parallelism,
			Salt: append([]byte(nil), material.salt...), PasswordHash: append([]byte(nil), material.hash...), ResetConsumedAt: sql.NullTime{Time: now, Valid: true},
		})
		if consumeErr != nil {
			return dependencyUnavailable()
		}
		if !consumed {
			return authenticationFailed()
		}
		created, createErr := service.newSessionTokens(now)
		if createErr != nil {
			return dependencyUnavailable()
		}
		result = created.tokens
		defer created.clear()
		if _, markErr := transaction.MarkPrincipalSessionsReviewRequired(transactionContext, store.MarkPrincipalSessionsReviewRequiredParams{PrincipalID: identity.PrincipalID, UpdatedAt: now}); markErr != nil {
			return dependencyUnavailable()
		}
		if createErr = transaction.CreateAccountSession(transactionContext, store.CreateAccountSessionParams{
			ID: created.sessionID, PrincipalID: identity.PrincipalID, ClientSigningPublicKey: command.ClientSigningPublicKey[:],
			AccessTokenHash: created.accessDigest[:], AccessExpiresAt: result.AccessExpiresAt,
			AbsoluteExpiresAt: result.RefreshAbsoluteExpiresAt, CreatedAt: now,
		}); createErr != nil {
			return dependencyUnavailable()
		}
		if createErr = transaction.InsertAccountRefreshToken(transactionContext, store.InsertAccountRefreshTokenParams{
			TokenHash: created.refreshDigest[:], SessionID: created.sessionID, IssuedAt: now,
			IdleExpiresAt: result.RefreshIdleExpiresAt, AbsoluteExpiresAt: result.RefreshAbsoluteExpiresAt,
		}); createErr != nil {
			return dependencyUnavailable()
		}
		eventID := service.newUUID()
		if eventID == uuid.Nil || transaction.InsertSecurityEvent(transactionContext, store.InsertSecurityEventParams{
			ID: eventID, PrincipalID: uuid.NullUUID{UUID: identity.PrincipalID, Valid: true}, Category: "password_reset",
			Fingerprint: "identity.password_reset", AggregateVersion: account.StateVersion, OccurredAt: now,
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

// CreateEnrollmentGrant creates one short-lived device enrollment capability after recent password proof.
func (service *Service) CreateEnrollmentGrant(ctx context.Context, command CreateEnrollmentGrantCommand) (EnrollmentGrant, error) {
	principalID, principalOK := parseCanonicalIdentityUUID(string(command.PrincipalID))
	sessionID, sessionOK := parseCanonicalIdentityUUID(string(command.SessionID))
	proofSessionID, proofSessionOK := parseCanonicalIdentityUUID(string(command.Reauthentication.SessionID))
	if !validService(service) || nilIdentityValue(ctx) || !principalOK || !sessionOK || !proofSessionOK || proofSessionID != sessionID ||
		command.Reauthentication.Method != ReauthPassword || !validPasswordSecret(command.Reauthentication.Proof) || !validIdempotencyKeyForApplication(command.IdempotencyKey) {
		return EnrollmentGrant{}, malformedRequest()
	}
	proof := command.Reauthentication.Proof.Copy()
	defer clear(proof)
	canonicalRequest, requestErr := privateCanonicalRequest(service.protector, "create_enrollment_grant", principalID[:], sessionID[:], proof)
	if requestErr != nil {
		return EnrollmentGrant{}, dependencyUnavailable()
	}
	defer clear(canonicalRequest)
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	var result EnrollmentGrant
	err := service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, transaction Transaction) error {
		now, ok := service.now()
		if !ok {
			return dependencyUnavailable()
		}
		account, accountFound, findErr := transaction.GetAccountForUpdate(transactionContext, principalID)
		if findErr != nil {
			return dependencyUnavailable()
		}
		session, sessionFound, findErr := transaction.GetAccountSessionForUpdate(transactionContext, store.GetAccountSessionForUpdateParams{ID: sessionID, PrincipalID: principalID})
		if findErr != nil {
			return dependencyUnavailable()
		}
		storedCredential, credentialFound, findErr := transaction.GetPasswordCredential(transactionContext, principalID)
		if findErr != nil {
			return dependencyUnavailable()
		}
		if !accountFound || !sessionFound || !credentialFound || account.ID != principalID || session.ID != sessionID || session.PrincipalID != principalID ||
			session.State != "active" || !session.AccessExpiresAt.After(now) || !session.AbsoluteExpiresAt.After(now) {
			if !service.performDummyPasswordWork() {
				return dependencyUnavailable()
			}
			return authenticationFailed()
		}
		credential, credentialErr := passwordCredentialFromStore(storedCredential)
		if credentialErr != nil {
			if !service.performDummyPasswordWork() {
				return dependencyUnavailable()
			}
			return authenticationFailed()
		}
		match, _ := VerifyPasswordWithDeriver(proof, credential, CurrentPasswordPolicy(), service.derive)
		if !match {
			return authenticationFailed()
		}
		scope, scopeErr := idempotency.AuthenticatedScope(principalID, "password", "create_enrollment_grant")
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
				decoded, decodeErr := decodeEnrollmentGrant(body)
				if decodeErr != nil {
					return dependencyUnavailable()
				}
				result = decoded
				return nil
			}
		}
		marker := "standard"
		var provisionalUntil sql.NullTime
		switch account.State {
		case "active":
		case "pending_email":
			if service.security.EmailVerification == config.EmailRequired {
				return actionNotAllowed()
			}
			if service.security.EmailVerification != config.EmailGrace {
				return actionNotAllowed()
			}
			marker = "trial_restricted"
			provisionalUntil = sql.NullTime{Time: now.Add(provisionalAuthorizationTTL), Valid: true}
		default:
			return actionNotAllowed()
		}
		token, tokenErr := securitykit.NewOpaqueToken(service.random)
		if tokenErr != nil {
			return dependencyUnavailable()
		}
		digest := securitykit.DigestToken(securitykit.EnrollmentGrantToken, token)
		grantID := service.newUUID()
		if digest == [32]byte{} || grantID == uuid.Nil {
			return dependencyUnavailable()
		}
		result = EnrollmentGrant{Token: token, ExpiresAt: now.Add(enrollmentGrantTTL), PolicyMarker: marker}
		if createErr := transaction.CreateEnrollmentGrant(transactionContext, store.CreateEnrollmentGrantParams{
			ID: grantID, PrincipalID: principalID, AccountSessionID: sessionID, TokenHash: digest[:], PolicyMarker: marker,
			ProvisionalUntil: provisionalUntil, ExpiresAt: result.ExpiresAt, CreatedAt: now,
		}); createErr != nil {
			return dependencyUnavailable()
		}
		body, encodeErr := encodeEnrollmentGrant(result)
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
		return EnrollmentGrant{}, mapApplicationError(operationContext, err)
	}
	return result, nil
}

type generatedSessionTokens struct {
	tokens        SessionTokens
	sessionID     uuid.UUID
	accessDigest  [32]byte
	refreshDigest [32]byte
}

func (generated *generatedSessionTokens) clear() {
	if generated == nil {
		return
	}
	access := generated.tokens.AccessToken.Copy()
	refresh := generated.tokens.RefreshToken.Copy()
	clear(access)
	clear(refresh)
}

func (service *Service) newSessionTokens(now time.Time) (generatedSessionTokens, error) {
	access, err := securitykit.NewOpaqueToken(service.random)
	if err != nil {
		return generatedSessionTokens{}, ErrRandomSource
	}
	refresh, err := securitykit.NewOpaqueToken(service.random)
	if err != nil {
		return generatedSessionTokens{}, ErrRandomSource
	}
	accessDigest := securitykit.DigestToken(securitykit.AccountAccessToken, access)
	refreshDigest := securitykit.DigestToken(securitykit.AccountRefreshToken, refresh)
	sessionID := service.newUUID()
	if accessDigest == [32]byte{} || refreshDigest == [32]byte{} || sessionID == uuid.Nil {
		return generatedSessionTokens{}, ErrRandomSource
	}
	return generatedSessionTokens{
		tokens: SessionTokens{
			AccessToken: access, AccessExpiresAt: now.Add(accountAccessTTL), RefreshToken: refresh,
			RefreshIdleExpiresAt: now.Add(accountRefreshIdleTTL), RefreshAbsoluteExpiresAt: now.Add(accountRefreshAbsoluteTTL),
		},
		sessionID: sessionID, accessDigest: accessDigest, refreshDigest: refreshDigest,
	}, nil
}

func passwordCredentialFromStore(stored store.IdentityPasswordCredential) (PasswordCredential, error) {
	if stored.PolicyVersion < 1 || stored.MemoryKib < 1 || stored.TimeCost < 1 || stored.Parallelism < 1 || stored.Parallelism > 255 ||
		uint64(len(stored.Salt)) > uint64(^uint32(0)) || uint64(len(stored.PasswordHash)) > uint64(^uint32(0)) {
		return PasswordCredential{}, ErrInvalidPasswordCredential
	}
	// #nosec G115 -- every signed field is positive and bounded above; slice lengths are bounded above explicitly.
	return NewPasswordCredential(PasswordPolicy{
		Version: uint32(stored.PolicyVersion), MemoryKiB: uint32(stored.MemoryKib), Time: uint32(stored.TimeCost),
		Parallelism: uint8(stored.Parallelism), SaltBytes: uint32(len(stored.Salt)), TagBytes: uint32(len(stored.PasswordHash)),
	}, stored.Salt, stored.PasswordHash)
}

func genericIdempotencyOutcome(outcome idempotency.Outcome, record idempotency.Record, status int) (bool, error) {
	switch outcome {
	case idempotency.Started:
		return false, nil
	case idempotency.Replay:
		body, owned := record.TakeResponseBody()
		defer clear(body)
		if owned && record.ResponseStatus() == status && ((status == 202 && string(body) == acceptedResponseBody) || (status == 204 && len(body) == 0)) {
			return true, nil
		}
		return false, dependencyUnavailable()
	case idempotency.Conflict:
		return false, apierrors.New(apierrors.IdempotencyConflict, apierrors.ContactSupport)
	case idempotency.InProgress:
		return false, apierrors.NewRetryAfter(apierrors.StateConflict, apierrors.Retry, time.Second)
	default:
		return false, dependencyUnavailable()
	}
}

func privateIdempotencyOutcome(outcome idempotency.Outcome, record idempotency.Record, status int) ([]byte, bool, error) {
	switch outcome {
	case idempotency.Started:
		return nil, false, dependencyUnavailable()
	case idempotency.Replay:
		body, owned := record.TakeResponseBody()
		if owned && record.ResponseStatus() == status && len(body) > 0 {
			return body, true, nil
		}
		clear(body)
		return nil, false, dependencyUnavailable()
	case idempotency.Conflict:
		return nil, false, apierrors.New(apierrors.IdempotencyConflict, apierrors.ContactSupport)
	case idempotency.InProgress:
		return nil, false, apierrors.NewRetryAfter(apierrors.StateConflict, apierrors.Retry, time.Second)
	default:
		return nil, false, dependencyUnavailable()
	}
}

type sessionTokensReplay struct {
	AccessToken              string `json:"access_token"`
	AccessExpiresAt          string `json:"access_expires_at"`
	RefreshToken             string `json:"refresh_token"`
	RefreshIdleExpiresAt     string `json:"refresh_idle_expires_at"`
	RefreshAbsoluteExpiresAt string `json:"refresh_absolute_expires_at"`
}

func encodeSessionTokens(tokens SessionTokens) ([]byte, error) {
	payload := sessionTokensReplay{
		AccessToken: securitykit.EncodeOpaqueToken(tokens.AccessToken), AccessExpiresAt: tokens.AccessExpiresAt.UTC().Format(time.RFC3339Nano),
		RefreshToken: securitykit.EncodeOpaqueToken(tokens.RefreshToken), RefreshIdleExpiresAt: tokens.RefreshIdleExpiresAt.UTC().Format(time.RFC3339Nano),
		RefreshAbsoluteExpiresAt: tokens.RefreshAbsoluteExpiresAt.UTC().Format(time.RFC3339Nano),
	}
	if payload.AccessToken == "" || payload.RefreshToken == "" {
		return nil, ErrRepository
	}
	// #nosec G117 -- this fixed private replay DTO is immediately encrypted by the idempotency repository.
	body, err := json.Marshal(payload)
	if err != nil || len(body) == 0 || int64(len(body)) > maximumPrivateReplayBytes {
		clear(body)
		return nil, ErrRepository
	}
	return body, nil
}

func decodeSessionTokens(body []byte) (SessionTokens, error) {
	var payload sessionTokensReplay
	if err := strictjson.Decode(bytes.NewReader(body), maximumPrivateReplayBytes, &payload); err != nil {
		return SessionTokens{}, ErrRepository
	}
	access, err := securitykit.DecodeOpaqueToken(payload.AccessToken)
	if err != nil {
		return SessionTokens{}, ErrRepository
	}
	refresh, err := securitykit.DecodeOpaqueToken(payload.RefreshToken)
	if err != nil {
		return SessionTokens{}, ErrRepository
	}
	accessExpiresAt, err := time.Parse(time.RFC3339Nano, payload.AccessExpiresAt)
	if err != nil {
		return SessionTokens{}, ErrRepository
	}
	refreshIdleExpiresAt, err := time.Parse(time.RFC3339Nano, payload.RefreshIdleExpiresAt)
	if err != nil {
		return SessionTokens{}, ErrRepository
	}
	refreshAbsoluteExpiresAt, err := time.Parse(time.RFC3339Nano, payload.RefreshAbsoluteExpiresAt)
	if err != nil || !accessExpiresAt.Before(refreshIdleExpiresAt) || !refreshIdleExpiresAt.Before(refreshAbsoluteExpiresAt) {
		return SessionTokens{}, ErrRepository
	}
	return SessionTokens{AccessToken: access, AccessExpiresAt: accessExpiresAt, RefreshToken: refresh, RefreshIdleExpiresAt: refreshIdleExpiresAt, RefreshAbsoluteExpiresAt: refreshAbsoluteExpiresAt}, nil
}

type enrollmentGrantReplay struct {
	Token        string `json:"token"`
	ExpiresAt    string `json:"expires_at"`
	PolicyMarker string `json:"policy_marker"`
}

func encodeEnrollmentGrant(grant EnrollmentGrant) ([]byte, error) {
	payload := enrollmentGrantReplay{Token: securitykit.EncodeOpaqueToken(grant.Token), ExpiresAt: grant.ExpiresAt.UTC().Format(time.RFC3339Nano), PolicyMarker: grant.PolicyMarker}
	if payload.Token == "" || (payload.PolicyMarker != "standard" && payload.PolicyMarker != "trial_restricted") {
		return nil, ErrRepository
	}
	body, err := json.Marshal(payload)
	if err != nil || len(body) == 0 || int64(len(body)) > maximumPrivateReplayBytes {
		clear(body)
		return nil, ErrRepository
	}
	return body, nil
}

func decodeEnrollmentGrant(body []byte) (EnrollmentGrant, error) {
	var payload enrollmentGrantReplay
	if err := strictjson.Decode(bytes.NewReader(body), maximumPrivateReplayBytes, &payload); err != nil {
		return EnrollmentGrant{}, ErrRepository
	}
	token, err := securitykit.DecodeOpaqueToken(payload.Token)
	if err != nil || (payload.PolicyMarker != "standard" && payload.PolicyMarker != "trial_restricted") {
		return EnrollmentGrant{}, ErrRepository
	}
	expiresAt, err := time.Parse(time.RFC3339Nano, payload.ExpiresAt)
	if err != nil {
		return EnrollmentGrant{}, ErrRepository
	}
	return EnrollmentGrant{Token: token, ExpiresAt: expiresAt, PolicyMarker: payload.PolicyMarker}, nil
}

func parseCanonicalIdentityUUID(value string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(value)
	return parsed, err == nil && parsed != uuid.Nil && parsed.String() == value
}

func protectedKeyVersionInt32(version uint32) (int32, bool) {
	if version < 1 || version > uint32(^uint32(0)>>1) {
		return 0, false
	}
	// #nosec G115 -- the explicit maximum above is exactly MaxInt32.
	return int32(version), true
}

func passwordPolicyInt32(policy PasswordPolicy) (version, memoryKiB, timeCost, parallelism int32, ok bool) {
	if policy != CurrentPasswordPolicy() {
		return 0, 0, 0, 0, false
	}
	// #nosec G115 -- the only accepted policy is the frozen, small v1 policy.
	return int32(policy.Version), int32(policy.MemoryKiB), int32(policy.Time), int32(policy.Parallelism), true
}
