package deviceauth

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/subtle"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	deviceauthv1 "talenro.local/platform/gen/go/talenro/deviceauth/v1"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/contracts/events"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/store"
)

const deviceRotationProofPrefix = "TALENRO-DEVICE-ROTATION-V1\x00"

type deviceTokenTransaction interface {
	Transaction
	DiscoverDeviceRefreshToken(context.Context, []byte) (store.DiscoverDeviceRefreshTokenRow, bool, error)
	LockDeviceFamilyRefreshTokens(context.Context, uuid.UUID) ([]store.DeviceauthDeviceRefreshToken, error)
	ListDeviceFamilyRefreshTokens(context.Context, uuid.UUID) ([]store.DeviceauthDeviceRefreshToken, error)
	ListDeviceAuthorizationFamilies(context.Context, uuid.UUID) ([]store.DeviceauthDeviceTokenFamily, error)
	GetDeviceTokenFamilyForUpdate(context.Context, uuid.UUID) (store.DeviceauthDeviceTokenFamily, bool, error)
	GetDeviceAuthorizationForUpdate(context.Context, uuid.UUID) (store.DeviceauthDeviceAuthorization, bool, error)
	DiscoverDeviceAuthorization(context.Context, uuid.UUID) (store.DeviceauthDeviceAuthorization, bool, error)
	GetDeviceForUpdate(context.Context, uuid.UUID) (store.DeviceauthDevice, bool, error)
	GetDevicePolicySnapshot(context.Context, uuid.UUID) (store.DeviceauthDevicePolicySnapshot, bool, error)
	MarkDeviceRefreshUsed(context.Context, store.MarkDeviceRefreshUsedParams) (bool, error)
	RotateDeviceFamilyAccess(context.Context, store.RotateDeviceFamilyAccessParams) (store.DeviceauthDeviceTokenFamily, bool, error)
	CompromiseDeviceTokenFamily(context.Context, store.CompromiseDeviceTokenFamilyParams) (store.DeviceauthDeviceTokenFamily, bool, error)
	RevokeDeviceRefreshTokens(context.Context, store.RevokeDeviceRefreshTokensParams) (int64, error)
	RevokeDeviceTokenFamily(context.Context, store.RevokeDeviceTokenFamilyParams) (int64, error)
	RevokeDeviceAuthorization(context.Context, store.RevokeDeviceAuthorizationParams) (int64, error)
	RevokeDeviceRecord(context.Context, store.RevokeDeviceRecordParams) (int64, error)
}

type deviceRefreshAuthority struct {
	discovered    store.DiscoverDeviceRefreshTokenRow
	family        store.DeviceauthDeviceTokenFamily
	authorization store.DeviceauthDeviceAuthorization
	device        store.DeviceauthDevice
	refresh       []store.DeviceauthDeviceRefreshToken
}

func (authority *deviceRefreshAuthority) clear() {
	if authority == nil {
		return
	}
	clearDeviceRefreshDiscovery(&authority.discovered)
	clear(authority.family.AccessTokenHash)
	clearDeviceRowSecrets(&authority.device)
	clearDeviceRefreshRows(authority.refresh)
	*authority = deviceRefreshAuthority{}
}

type preparedDeviceRotation struct {
	tokens        DeviceTokens
	accessDigest  [32]byte
	refreshDigest [32]byte
}

func (prepared *preparedDeviceRotation) clear() {
	if prepared == nil {
		return
	}
	prepared.tokens.AccessToken.Clear()
	prepared.tokens.RefreshToken.Clear()
	clear(prepared.accessDigest[:])
	clear(prepared.refreshDigest[:])
	*prepared = preparedDeviceRotation{}
}

func (prepared *preparedDeviceRotation) takeTokens() DeviceTokens {
	if prepared == nil {
		return DeviceTokens{}
	}
	return takeDeviceTokens(&prepared.tokens)
}

func (deviceRefreshAuthority) Format(state fmt.State, _ rune) {
	redactDeviceauthValue(state, "deviceRefreshAuthority")
}

func (deviceRefreshAuthority) LogValue() slog.Value {
	return slog.StringValue("deviceauth.deviceRefreshAuthority([REDACTED])")
}

func (deviceRefreshAuthority) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: refresh authority serialization forbidden")
}

func (preparedDeviceRotation) Format(state fmt.State, _ rune) {
	redactDeviceauthValue(state, "preparedDeviceRotation")
}

func (preparedDeviceRotation) LogValue() slog.Value {
	return slog.StringValue("deviceauth.preparedDeviceRotation([REDACTED])")
}

func (preparedDeviceRotation) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: prepared rotation serialization forbidden")
}

const (
	privateRotationBindingDomain       = "deviceauth/rotate-request/v1"
	deviceRotationIdempotencyRetention = 90 * 24 * time.Hour
	deviceRotationResponseStatus       = 200
	deviceReplayResponseStatus         = 401
	deviceReplayResponseBody           = `{"authenticated":false}`
)

var errDeviceRotationPreflight = errors.New("deviceauth: rotation preflight rollback")

func deviceRotationScope(principalID uuid.UUID) (idempotency.Scope, error) {
	return idempotency.AuthenticatedScope(principalID, "device_refresh", rotateDeviceTokenOperation)
}

func usedAt(now time.Time) sql.NullTime { return sql.NullTime{Time: now, Valid: true} }

// RotateDeviceToken classifies PostgreSQL replay before Redis, then rotates one device family atomically.
func (service *Service) RotateDeviceToken(ctx context.Context, command RotateDeviceTokenCommand) (result DeviceTokens, resultErr error) {
	defer func() {
		if recover() != nil {
			result.AccessToken.Clear()
			result.RefreshToken.Clear()
			result = DeviceTokens{}
			resultErr = deviceDependencyUnavailable()
		}
	}()
	if !validDeviceauthService(service) || nilDeviceauthValue(ctx) || !validRotateDeviceTokenCommand(command) {
		return DeviceTokens{}, deviceMalformedRequest()
	}
	refreshDigest := securitykit.DigestToken(securitykit.DeviceRefreshToken, command.RefreshToken)
	if refreshDigest == [32]byte{} {
		return DeviceTokens{}, deviceAuthenticationFailed()
	}
	defer clear(refreshDigest[:])
	canonical, err := privateRotationBinding(service, refreshDigest, command)
	if err != nil {
		return DeviceTokens{}, deviceDependencyUnavailable()
	}
	defer clear(canonical)
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	now, ok := service.now()
	if !ok {
		return DeviceTokens{}, deviceDependencyUnavailable()
	}
	var authority deviceRefreshAuthority
	defer authority.clear()
	var replay DeviceTokens
	defer replay.AccessToken.Clear()
	defer replay.RefreshToken.Clear()
	preflightResolved := false
	postCommitAuthenticationFailure := false
	preflightErr := service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, base Transaction) error {
		transaction, typed := base.(deviceTokenTransaction)
		if !typed || nilDeviceauthValue(transaction) {
			return deviceDependencyUnavailable()
		}
		discovered, found, discoverErr := transaction.DiscoverDeviceRefreshToken(transactionContext, refreshDigest[:])
		defer clearDeviceRefreshDiscovery(&discovered)
		if discoverErr != nil {
			return deviceDependencyUnavailable()
		}
		if !found || discovered.PrincipalID == uuid.Nil {
			return deviceAuthenticationFailed()
		}
		scope, scopeErr := deviceRotationScope(discovered.PrincipalID)
		if scopeErr != nil {
			return deviceDependencyUnavailable()
		}
		record, outcome, beginErr := transaction.BeginIdempotency(
			transactionContext, scope, command.IdempotencyKey, canonical, now, now.Add(deviceRotationIdempotencyRetention),
		)
		if beginErr != nil {
			return deviceDependencyUnavailable()
		}
		resolved, authenticationFailure, outcomeErr := rotationIdempotencyOutcome(record, outcome, &replay)
		if outcomeErr != nil {
			return outcomeErr
		}
		if resolved {
			preflightResolved = true
			postCommitAuthenticationFailure = authenticationFailure
			return nil
		}
		locked, valid, lockErr := service.lockDeviceRefreshAuthority(transactionContext, transaction, refreshDigest, now, true)
		if lockErr != nil {
			return lockErr
		}
		if !valid || locked.discovered.PrincipalID != discovered.PrincipalID || locked.discovered.FamilyID != discovered.FamilyID {
			locked.clear()
			return deviceAuthenticationFailed()
		}
		if locked.discovered.RefreshState == "used" {
			defer locked.clear()
			if err := service.compromiseUsedDeviceRefresh(transactionContext, transaction, locked, record, now); err != nil {
				return err
			}
			preflightResolved = true
			postCommitAuthenticationFailure = true
			return nil
		}
		authority = locked
		return errDeviceRotationPreflight
	})
	if preflightErr == nil && preflightResolved {
		if postCommitAuthenticationFailure {
			return DeviceTokens{}, deviceAuthenticationFailed()
		}
		return takeDeviceTokens(&replay), nil
	}
	if !errors.Is(preflightErr, errDeviceRotationPreflight) {
		if preflightErr != nil {
			return DeviceTokens{}, preflightErr
		}
		return DeviceTokens{}, deviceDependencyUnavailable()
	}
	contextDigest, err := rotationContextDigest(authority.discovered.FamilyID, command.RequestNonce, service.security.PublicBaseURL)
	if err != nil {
		return DeviceTokens{}, deviceDependencyUnavailable()
	}
	defer clear(contextDigest[:])
	record, consumeErr := safeDeviceChallengeConsume(operationContext, service.challenges, command.ChallengeID, refreshDigest, contextDigest)
	if consumeErr != nil {
		ambiguous := !errors.Is(consumeErr, ErrChallengeNotFound) && !errors.Is(consumeErr, ErrInvalidChallenge)
		reclassified, resolved, authenticationFailure, reclassifyErr := service.reclassifyDeviceRotationAfterChallengeFailure(
			operationContext, command, canonical, refreshDigest, now, ambiguous,
		)
		if reclassifyErr != nil {
			return DeviceTokens{}, reclassifyErr
		}
		if authenticationFailure {
			return DeviceTokens{}, deviceAuthenticationFailed()
		}
		if resolved {
			return reclassified, nil
		}
		return DeviceTokens{}, deviceDependencyUnavailable()
	}
	if !record.ExpiresAt.After(now) || record.Kind != ChallengeRotation || record.ProtocolVersion != deviceRotationProtocolVersion ||
		record.Operation != rotateDeviceTokenOperation {
		return DeviceTokens{}, deviceAuthenticationFailed()
	}
	proof := DeviceRotationProofBytes(DeviceRotationProofInput{
		ProtocolVersion: record.ProtocolVersion, Challenge: record.Challenge, FamilyID: authority.discovered.FamilyID,
		Operation: record.Operation, Audience: service.security.PublicBaseURL, RequestNonce: command.RequestNonce,
	})
	validProof := len(authority.device.SigningPublicKey) == ed25519.PublicKeySize && len(proof) != 0 &&
		ed25519.Verify(ed25519.PublicKey(authority.device.SigningPublicKey), proof, command.Signature[:])
	observeDeviceProofVerification(operationContext, validProof)
	clear(proof)
	if !validProof {
		return DeviceTokens{}, deviceAuthenticationFailed()
	}
	var prepared preparedDeviceRotation
	defer prepared.clear()
	finalResolved := false
	postCommitAuthenticationFailure = false
	finalErr := service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, base Transaction) error {
		transaction, typed := base.(deviceTokenTransaction)
		if !typed || nilDeviceauthValue(transaction) {
			return deviceDependencyUnavailable()
		}
		discovered, found, discoverErr := transaction.DiscoverDeviceRefreshToken(transactionContext, refreshDigest[:])
		defer clearDeviceRefreshDiscovery(&discovered)
		if discoverErr != nil {
			return deviceDependencyUnavailable()
		}
		if !found || discovered.PrincipalID == uuid.Nil {
			return deviceAuthenticationFailed()
		}
		scope, scopeErr := deviceRotationScope(discovered.PrincipalID)
		if scopeErr != nil {
			return deviceDependencyUnavailable()
		}
		idempotencyRecord, outcome, beginErr := transaction.BeginIdempotency(
			transactionContext, scope, command.IdempotencyKey, canonical, now, now.Add(deviceRotationIdempotencyRetention),
		)
		if beginErr != nil {
			return deviceDependencyUnavailable()
		}
		resolved, authenticationFailure, outcomeErr := rotationIdempotencyOutcome(idempotencyRecord, outcome, &replay)
		if outcomeErr != nil {
			return outcomeErr
		}
		if resolved {
			finalResolved = true
			postCommitAuthenticationFailure = authenticationFailure
			return nil
		}
		locked, valid, lockErr := service.lockDeviceRefreshAuthority(transactionContext, transaction, refreshDigest, now, true)
		if lockErr != nil {
			return lockErr
		}
		defer locked.clear()
		if !valid || !sameDeviceRefreshAuthority(authority, locked) {
			return deviceAuthenticationFailed()
		}
		if locked.discovered.RefreshState == "used" {
			if err := service.compromiseUsedDeviceRefresh(transactionContext, transaction, locked, idempotencyRecord, now); err != nil {
				return err
			}
			postCommitAuthenticationFailure = true
			return nil
		}
		if locked.discovered.RefreshState != "active" {
			return deviceAuthenticationFailed()
		}
		created, createErr := service.prepareDeviceRotation(locked, now)
		if createErr != nil {
			return createErr
		}
		prepared = created
		oldDigest := append([]byte(nil), refreshDigest[:]...)
		accessDigest := append([]byte(nil), prepared.accessDigest[:]...)
		newRefreshDigest := append([]byte(nil), prepared.refreshDigest[:]...)
		previousDigest := append([]byte(nil), refreshDigest[:]...)
		defer clear(oldDigest)
		defer clear(accessDigest)
		defer clear(newRefreshDigest)
		defer clear(previousDigest)
		marked, markErr := transaction.MarkDeviceRefreshUsed(transactionContext, store.MarkDeviceRefreshUsedParams{
			TokenHash: oldDigest, UsedAt: usedAt(now),
		})
		if markErr != nil || !marked {
			return deviceDependencyUnavailable()
		}
		rotated, changed, rotateErr := transaction.RotateDeviceFamilyAccess(transactionContext, store.RotateDeviceFamilyAccessParams{
			ID: locked.family.ID, AccessTokenHash: accessDigest,
			AccessExpiresAt: prepared.tokens.AccessExpiresAt, IdleExpiresAt: prepared.tokens.RefreshIdleExpiresAt, UpdatedAt: now,
		})
		defer clear(rotated.AccessTokenHash)
		if rotateErr != nil || !changed || rotated.ID != locked.family.ID || rotated.AuthorizationID != locked.authorization.ID || rotated.State != "active" ||
			rotated.StateVersion != locked.family.StateVersion+1 || !rotated.AbsoluteExpiresAt.Equal(locked.family.AbsoluteExpiresAt) {
			return deviceDependencyUnavailable()
		}
		if err := transaction.InsertDeviceRefreshToken(transactionContext, store.InsertDeviceRefreshTokenParams{
			TokenHash: newRefreshDigest, FamilyID: locked.family.ID,
			PreviousTokenHash: previousDigest, IssuedAt: now,
		}); err != nil {
			return deviceDependencyUnavailable()
		}
		body, encodeErr := encodeDeviceTokens(prepared.tokens)
		if encodeErr != nil {
			return deviceDependencyUnavailable()
		}
		defer clear(body)
		if err := transaction.CompleteIdempotency(transactionContext, idempotencyRecord, deviceRotationResponseStatus, body); err != nil {
			return deviceDependencyUnavailable()
		}
		return nil
	})
	if finalErr != nil {
		return DeviceTokens{}, finalErr
	}
	if postCommitAuthenticationFailure {
		return DeviceTokens{}, deviceAuthenticationFailed()
	}
	if finalResolved {
		return takeDeviceTokens(&replay), nil
	}
	return prepared.takeTokens(), nil
}

func (service *Service) lockDeviceRefreshAuthority(
	ctx context.Context,
	transaction deviceTokenTransaction,
	digest [32]byte,
	now time.Time,
	allowUsed bool,
) (deviceRefreshAuthority, bool, error) {
	discovered, found, err := transaction.DiscoverDeviceRefreshToken(ctx, digest[:])
	if err != nil {
		clearDeviceRefreshDiscovery(&discovered)
		return deviceRefreshAuthority{}, false, deviceDependencyUnavailable()
	}
	if !found {
		clearDeviceRefreshDiscovery(&discovered)
		return deviceRefreshAuthority{}, false, nil
	}
	anchorFamilyID := discovered.FamilyID
	anchorAuthorizationID := discovered.AuthorizationID
	anchorPrincipalID := discovered.PrincipalID
	anchorDeviceID := discovered.DeviceID
	clearDeviceRefreshDiscovery(&discovered)
	if anchorFamilyID == uuid.Nil || anchorAuthorizationID == uuid.Nil || anchorPrincipalID == uuid.Nil || anchorDeviceID == uuid.Nil {
		return deviceRefreshAuthority{}, false, nil
	}

	var result deviceRefreshAuthority
	result.refresh, err = transaction.LockDeviceFamilyRefreshTokens(ctx, anchorFamilyID)
	if err != nil {
		result.clear()
		return deviceRefreshAuthority{}, false, deviceDependencyUnavailable()
	}
	result.family, found, err = transaction.GetDeviceTokenFamilyForUpdate(ctx, anchorFamilyID)
	if err != nil {
		result.clear()
		return deviceRefreshAuthority{}, false, deviceDependencyUnavailable()
	}
	if !found {
		result.clear()
		return deviceRefreshAuthority{}, false, nil
	}
	result.authorization, found, err = transaction.GetDeviceAuthorizationForUpdate(ctx, anchorAuthorizationID)
	if err != nil {
		result.clear()
		return deviceRefreshAuthority{}, false, deviceDependencyUnavailable()
	}
	if !found {
		result.clear()
		return deviceRefreshAuthority{}, false, nil
	}
	result.device, found, err = transaction.GetDeviceForUpdate(ctx, anchorDeviceID)
	if err != nil {
		result.clear()
		return deviceRefreshAuthority{}, false, deviceDependencyUnavailable()
	}
	if !found {
		result.clear()
		return deviceRefreshAuthority{}, false, nil
	}
	result.discovered, found, err = transaction.DiscoverDeviceRefreshToken(ctx, digest[:])
	if err != nil {
		result.clear()
		return deviceRefreshAuthority{}, false, deviceDependencyUnavailable()
	}
	if !found || result.discovered.FamilyID != anchorFamilyID || result.discovered.AuthorizationID != anchorAuthorizationID ||
		result.discovered.PrincipalID != anchorPrincipalID || result.discovered.DeviceID != anchorDeviceID {
		result.clear()
		return deviceRefreshAuthority{}, false, nil
	}
	accountActive, err := service.identity.ValidateDeviceAccountAuthority(ctx, transaction.DBTX(), identity.PrincipalID(result.discovered.PrincipalID.String()), now)
	if err != nil {
		result.clear()
		return deviceRefreshAuthority{}, false, deviceDependencyUnavailable()
	}
	freshRefresh, err := transaction.ListDeviceFamilyRefreshTokens(ctx, result.discovered.FamilyID)
	if err != nil {
		clearDeviceRefreshRows(freshRefresh)
		result.clear()
		return deviceRefreshAuthority{}, false, deviceDependencyUnavailable()
	}
	valid := accountActive && validDeviceRefreshAuthority(result.discovered, result.family, result.authorization, result.device, result.refresh, freshRefresh, digest, now, allowUsed) &&
		(result.authorization.State != "provisional" || service.security.EmailVerification == config.EmailGrace)
	clearDeviceRefreshRows(freshRefresh)
	if !valid {
		result.clear()
		return deviceRefreshAuthority{}, false, nil
	}
	return result, true, nil
}

func validDeviceRefreshAuthority(
	discovered store.DiscoverDeviceRefreshTokenRow,
	family store.DeviceauthDeviceTokenFamily,
	authorization store.DeviceauthDeviceAuthorization,
	device store.DeviceauthDevice,
	lockedRefresh, freshRefresh []store.DeviceauthDeviceRefreshToken,
	digest [32]byte,
	now time.Time,
	allowUsed bool,
) bool {
	refreshStateValid := discovered.RefreshState == "active" || (allowUsed && discovered.RefreshState == "used")
	expiryValid := family.AbsoluteExpiresAt.After(now) &&
		(discovered.RefreshState == "used" || family.IdleExpiresAt.After(now))
	if !refreshStateValid || len(discovered.TokenHash) != len(digest) || subtle.ConstantTimeCompare(discovered.TokenHash, digest[:]) != 1 ||
		discovered.FamilyID == uuid.Nil || discovered.AuthorizationID == uuid.Nil || discovered.PrincipalID == uuid.Nil || discovered.DeviceID == uuid.Nil ||
		family.ID != discovered.FamilyID || family.AuthorizationID != discovered.AuthorizationID || family.State != "active" ||
		family.State != discovered.FamilyState || family.StateVersion != discovered.FamilyStateVersion || !family.IdleExpiresAt.Equal(discovered.IdleExpiresAt) ||
		!family.AbsoluteExpiresAt.Equal(discovered.AbsoluteExpiresAt) || !expiryValid ||
		authorization.ID != discovered.AuthorizationID || authorization.PrincipalID != discovered.PrincipalID || authorization.DeviceID != discovered.DeviceID ||
		authorization.State != discovered.AuthorizationState || authorization.StateVersion != discovered.AuthorizationStateVersion ||
		!validAuthorizationMarker(discovered.AuthorizationState, discovered.ProvisionalUntil, authorization.ProvisionalUntil, now) ||
		device.ID != discovered.DeviceID || device.PrincipalID != discovered.PrincipalID || device.State != "active" || device.State != discovered.DeviceState ||
		device.KeyVersion != discovered.KeyVersion || len(device.SigningPublicKey) != ed25519.PublicKeySize || len(discovered.SigningPublicKey) != ed25519.PublicKeySize ||
		subtle.ConstantTimeCompare(device.SigningPublicKey, discovered.SigningPublicKey) != 1 || !sameDeviceRefreshSet(lockedRefresh, freshRefresh) {
		return false
	}
	found := false
	for index := range lockedRefresh {
		if lockedRefresh[index].FamilyID != discovered.FamilyID || len(lockedRefresh[index].TokenHash) != 32 ||
			(index > 0 && bytes.Compare(lockedRefresh[index-1].TokenHash, lockedRefresh[index].TokenHash) >= 0) {
			return false
		}
		if subtle.ConstantTimeCompare(lockedRefresh[index].TokenHash, digest[:]) == 1 {
			if found || lockedRefresh[index].State != discovered.RefreshState {
				return false
			}
			found = true
		}
	}
	return found
}

func sameDeviceRefreshAuthority(left, right deviceRefreshAuthority) bool {
	return left.discovered.FamilyID == right.discovered.FamilyID && left.discovered.AuthorizationID == right.discovered.AuthorizationID &&
		left.discovered.PrincipalID == right.discovered.PrincipalID && left.discovered.DeviceID == right.discovered.DeviceID &&
		left.family.AbsoluteExpiresAt.Equal(right.family.AbsoluteExpiresAt) && left.authorization.StateVersion == right.authorization.StateVersion &&
		left.device.KeyVersion == right.device.KeyVersion && len(left.device.SigningPublicKey) == ed25519.PublicKeySize &&
		len(right.device.SigningPublicKey) == ed25519.PublicKeySize && subtle.ConstantTimeCompare(left.device.SigningPublicKey, right.device.SigningPublicKey) == 1
}

func (service *Service) prepareDeviceRotation(authority deviceRefreshAuthority, now time.Time) (preparedDeviceRotation, error) {
	absolute := authority.family.AbsoluteExpiresAt
	accessExpiresAt := now.Add(deviceAccessTTL)
	if authority.authorization.ProvisionalUntil.Valid && accessExpiresAt.After(authority.authorization.ProvisionalUntil.Time) {
		accessExpiresAt = authority.authorization.ProvisionalUntil.Time
	}
	if !accessExpiresAt.Before(absolute) {
		return preparedDeviceRotation{}, deviceAuthenticationFailed()
	}
	idleExpiresAt := now.Add(deviceRefreshIdleTTL)
	if !idleExpiresAt.Before(absolute) {
		idleExpiresAt = absolute.Add(-time.Nanosecond)
	}
	if !accessExpiresAt.Before(idleExpiresAt) {
		return preparedDeviceRotation{}, deviceAuthenticationFailed()
	}
	access, err := securitykit.NewOpaqueToken(service.random)
	if err != nil {
		return preparedDeviceRotation{}, deviceDependencyUnavailable()
	}
	defer access.Clear()
	refresh, err := securitykit.NewOpaqueToken(service.random)
	if err != nil {
		return preparedDeviceRotation{}, deviceDependencyUnavailable()
	}
	defer refresh.Clear()
	accessDigest := securitykit.DigestToken(securitykit.DeviceAccessToken, access)
	refreshDigest := securitykit.DigestToken(securitykit.DeviceRefreshToken, refresh)
	if accessDigest == [32]byte{} || refreshDigest == [32]byte{} {
		clear(accessDigest[:])
		clear(refreshDigest[:])
		return preparedDeviceRotation{}, deviceDependencyUnavailable()
	}
	return preparedDeviceRotation{tokens: DeviceTokens{DeviceID: authority.device.ID, AuthorizationID: authority.authorization.ID, FamilyID: authority.family.ID,
		AccessToken: access.Take(), RefreshToken: refresh.Take(), AccessExpiresAt: accessExpiresAt,
		RefreshIdleExpiresAt: idleExpiresAt, RefreshAbsoluteExpiresAt: absolute}, accessDigest: accessDigest, refreshDigest: refreshDigest}, nil
}

func (service *Service) compromiseUsedDeviceRefresh(
	ctx context.Context,
	transaction deviceTokenTransaction,
	authority deviceRefreshAuthority,
	record idempotency.Record,
	now time.Time,
) error {
	compromised, changed, err := transaction.CompromiseDeviceTokenFamily(ctx, store.CompromiseDeviceTokenFamilyParams{ID: authority.family.ID, UpdatedAt: now})
	defer clear(compromised.AccessTokenHash)
	if err != nil || !changed || compromised.ID != authority.family.ID || compromised.AuthorizationID != authority.authorization.ID ||
		compromised.State != "compromised" || compromised.StateVersion != authority.family.StateVersion+1 {
		return deviceDependencyUnavailable()
	}
	if _, err := transaction.RevokeDeviceRefreshTokens(ctx, store.RevokeDeviceRefreshTokensParams{FamilyID: authority.family.ID, RevokedAt: usedAt(now)}); err != nil {
		return deviceDependencyUnavailable()
	}
	securityEventID, err := service.randomUUID()
	if err != nil {
		return deviceDependencyUnavailable()
	}
	securityRecord, err := identity.NewDeviceTokenReplaySecurityRecord(securityEventID, authority.authorization.ID)
	if err != nil || service.identity.RecordDeviceTokenReplay(ctx, transaction.DBTX(), securityRecord, now) != nil {
		return deviceDependencyUnavailable()
	}
	outboxEventID, err := service.randomUUID()
	if err != nil {
		return deviceDependencyUnavailable()
	}
	if compromised.StateVersion <= 0 {
		return deviceDependencyUnavailable()
	}
	eventVersion := uint64(compromised.StateVersion) // #nosec G115 -- strictly positive int64 is representable by uint64.
	payload, err := events.MarshalPayload(events.DeviceTokenFamilyCompromisedType, &deviceauthv1.DeviceTokenFamilyCompromised{
		AuthorizationId: authority.authorization.ID.String(), FamilyId: authority.family.ID.String(), Version: eventVersion,
	})
	if err != nil {
		return deviceDependencyUnavailable()
	}
	event := &eventsv1.EventEnvelope{EventId: outboxEventID.String(), EventType: events.DeviceTokenFamilyCompromisedType,
		OccurredAt: timestamppb.New(now), Producer: "deviceauth", AggregateType: "device_authorization",
		AggregateId: authority.authorization.ID.String(), AggregateVersion: eventVersion, IdempotencyKey: outboxEventID.String(), Payload: payload}
	if transaction.AppendEvent(ctx, event) != nil || transaction.CompleteIdempotency(ctx, record, deviceReplayResponseStatus, []byte(deviceReplayResponseBody)) != nil {
		return deviceDependencyUnavailable()
	}
	return nil
}

func (service *Service) reclassifyDeviceRotationAfterChallengeFailure(
	ctx context.Context,
	command RotateDeviceTokenCommand,
	canonical []byte,
	digest [32]byte,
	now time.Time,
	ambiguous bool,
) (result DeviceTokens, resolved, authenticationFailure bool, resultErr error) {
	err := service.repository.WithinTransaction(ctx, func(transactionContext context.Context, base Transaction) error {
		transaction, typed := base.(deviceTokenTransaction)
		if !typed || nilDeviceauthValue(transaction) {
			return deviceDependencyUnavailable()
		}
		discovered, found, discoverErr := transaction.DiscoverDeviceRefreshToken(transactionContext, digest[:])
		defer clearDeviceRefreshDiscovery(&discovered)
		if discoverErr != nil {
			return deviceDependencyUnavailable()
		}
		if !found || discovered.PrincipalID == uuid.Nil {
			if ambiguous {
				return deviceDependencyUnavailable()
			}
			return deviceAuthenticationFailed()
		}
		scope, scopeErr := deviceRotationScope(discovered.PrincipalID)
		if scopeErr != nil {
			return deviceDependencyUnavailable()
		}
		record, outcome, beginErr := transaction.BeginIdempotency(transactionContext, scope, command.IdempotencyKey, canonical, now, now.Add(deviceRotationIdempotencyRetention))
		if beginErr != nil {
			return deviceDependencyUnavailable()
		}
		outcomeResolved, outcomeAuthenticationFailure, outcomeErr := rotationIdempotencyOutcome(record, outcome, &result)
		if outcomeErr != nil {
			return outcomeErr
		}
		if outcomeResolved {
			resolved, authenticationFailure = true, outcomeAuthenticationFailure
			return nil
		}
		locked, valid, lockErr := service.lockDeviceRefreshAuthority(transactionContext, transaction, digest, now, true)
		if lockErr != nil {
			return lockErr
		}
		if !valid {
			if ambiguous {
				return deviceDependencyUnavailable()
			}
			return deviceAuthenticationFailed()
		}
		defer locked.clear()
		if locked.discovered.RefreshState == "used" {
			if err := service.compromiseUsedDeviceRefresh(transactionContext, transaction, locked, record, now); err != nil {
				return err
			}
			resolved, authenticationFailure = true, true
			return nil
		}
		if ambiguous {
			return deviceDependencyUnavailable()
		}
		return errDeviceRotationPreflight
	})
	if errors.Is(err, errDeviceRotationPreflight) {
		return DeviceTokens{}, false, true, nil
	}
	if err != nil {
		return DeviceTokens{}, false, false, err
	}
	return result, resolved, authenticationFailure, nil
}

func rotationIdempotencyOutcome(record idempotency.Record, outcome idempotency.Outcome, replay *DeviceTokens) (bool, bool, error) {
	switch outcome {
	case idempotency.Started:
		return false, false, nil
	case idempotency.Replay:
		if record.ResponseStatus() == deviceReplayResponseStatus {
			body, owned := record.TakeResponseBody()
			defer clear(body)
			if !owned || string(body) != deviceReplayResponseBody {
				return false, false, deviceDependencyUnavailable()
			}
			return true, true, nil
		}
		decoded, err := tokensFromReplayRecordWithStatus(record, deviceRotationResponseStatus)
		if err != nil {
			return false, false, deviceDependencyUnavailable()
		}
		if replay == nil {
			decoded.AccessToken.Clear()
			decoded.RefreshToken.Clear()
			return false, false, deviceDependencyUnavailable()
		}
		replay.AccessToken.Clear()
		replay.RefreshToken.Clear()
		*replay = takeDeviceTokens(&decoded)
		return true, false, nil
	case idempotency.Conflict:
		return false, false, apierrors.New(apierrors.IdempotencyConflict, apierrors.ContactSupport)
	case idempotency.InProgress:
		return false, false, apierrors.NewRetryAfter(apierrors.StateConflict, apierrors.Retry, time.Second)
	default:
		return false, false, deviceDependencyUnavailable()
	}
}

func validRotateDeviceTokenCommand(command RotateDeviceTokenCommand) bool {
	refresh := command.RefreshToken.Copy()
	defer clear(refresh)
	return len(refresh) == 32 && validChallengeID(command.ChallengeID) && command.RequestNonce != [32]byte{} &&
		command.Signature != [64]byte{} && validDeviceauthIdempotencyKey(command.IdempotencyKey)
}

func privateRotationBinding(service *Service, digest [32]byte, command RotateDeviceTokenCommand) ([]byte, error) {
	if service == nil || nilDeviceauthValue(service.protector) || digest == [32]byte{} || !validRotateDeviceTokenCommand(command) {
		return nil, ErrInvalidApplication
	}
	challengeID := []byte(command.ChallengeID)
	audience := []byte(service.security.PublicBaseURL)
	material := make([]byte, 0, 256+len(challengeID)+len(audience))
	material = appendRegistrationFrame(material, digest[:])
	material = appendRegistrationFrame(material, challengeID)
	material = appendRegistrationFrame(material, command.RequestNonce[:])
	material = appendRegistrationFrame(material, command.Signature[:])
	material = appendRegistrationFrame(material, audience)
	clear(challengeID)
	clear(audience)
	defer clear(material)
	bound := service.protector.LookupDigest(privateRotationBindingDomain, material)
	if bound == [32]byte{} {
		return nil, ErrInvalidApplication
	}
	result := append([]byte(nil), bound[:]...)
	clear(bound[:])
	return result, nil
}

// DeviceRotationProofInput is the exact device refresh proof transcript.
type DeviceRotationProofInput struct {
	ProtocolVersion string
	Challenge       [32]byte
	FamilyID        uuid.UUID
	Operation       string
	Audience        string
	RequestNonce    [32]byte
}

// DeviceRotationProofBytes returns the fixed unambiguous rotation transcript.
func DeviceRotationProofBytes(input DeviceRotationProofInput) []byte {
	if input.ProtocolVersion == "" || len(input.ProtocolVersion) > maximumProofNameBytes || input.Challenge == [32]byte{} ||
		input.FamilyID == uuid.Nil || input.Operation == "" || len(input.Operation) > maximumProofNameBytes ||
		!validProofString(input.Audience, maximumProofAudienceBytes) || input.RequestNonce == [32]byte{} {
		return nil
	}
	result := make([]byte, 0, len(deviceRotationProofPrefix)+12+len(input.ProtocolVersion)+len(input.Operation)+len(input.Audience)+32+16+32)
	result = append(result, deviceRotationProofPrefix...)
	result = appendDeviceRotationString(result, input.ProtocolVersion)
	result = append(result, input.Challenge[:]...)
	result = append(result, input.FamilyID[:]...)
	result = appendDeviceRotationString(result, input.Operation)
	result = appendDeviceRotationString(result, input.Audience)
	result = append(result, input.RequestNonce[:]...)
	return result
}

func appendDeviceRotationString(target []byte, value string) []byte {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value))) // #nosec G115 -- all callers enforce fixed small bounds.
	target = append(target, length[:]...)
	clear(length[:])
	return append(target, value...)
}

// Format redacts rotation proof input.
func (DeviceRotationProofInput) Format(state fmt.State, _ rune) {
	redactDeviceauthValue(state, "DeviceRotationProofInput")
}

// LogValue redacts rotation proof input.
func (DeviceRotationProofInput) LogValue() slog.Value {
	return slog.StringValue("deviceauth.DeviceRotationProofInput([REDACTED])")
}

// MarshalJSON forbids generic rotation proof serialization.
func (DeviceRotationProofInput) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: rotation proof serialization forbidden")
}
