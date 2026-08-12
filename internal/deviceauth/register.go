package deviceauth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	"reflect"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	deviceauthv1 "talenro.local/platform/gen/go/talenro/deviceauth/v1"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/contracts/events"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/ratelimit"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

const (
	registrationContextDomain        = "TALENRO-DEVICE-REGISTRATION-CONTEXT-V1\x00"
	rotationContextDomain            = "TALENRO-DEVICE-ROTATION-CONTEXT-V1\x00"
	privateRegistrationBindingDomain = "deviceauth/register-request/v1"
	displayNameProtectionDomain      = "deviceauth/display-name/v1"
	deviceAccessTTL                  = 10 * time.Minute
	deviceRefreshIdleTTL             = 30 * 24 * time.Hour
	deviceRefreshAbsoluteTTL         = 90 * 24 * time.Hour
	registrationIdempotencyRetention = 24 * time.Hour
	maximumDisplayNameBytes          = 256
	maximumDisplayNameRunes          = 64
	maximumDisplayCiphertextBytes    = 1024
	maximumEmptyRandomReads          = 8
	registrationResponseStatus       = 201
)

var (
	// ErrInvalidApplication reports unsafe Task 12 dependencies without retaining them.
	ErrInvalidApplication = errors.New("deviceauth: invalid application configuration")
	// ErrRepository reports a fixed transaction or storage ambiguity.
	ErrRepository                    = errors.New("deviceauth: repository unavailable")
	errRegistrationPreflightRollback = errors.New("deviceauth: registration preflight rollback")
	errChallengeValidationRollback   = errors.New("deviceauth: challenge validation rollback")
	errRotationChallengeRollback     = errors.New("deviceauth: rotation challenge validation rollback")
)

// Service implements Task 12 registration while retaining the frozen Task 13 surface for later completion.
type Service struct {
	repository   Repository
	identity     identity.DeviceTransactionParticipant
	protector    sensitive.Protector
	random       securitykit.RandomSource
	clock        securitykit.Clock
	limiter      ratelimit.Limiter
	challenges   ChallengeStore
	rateLimitKey secret.Bytes
	security     config.SecurityConfig
}

// NewApplication validates and isolates all Task 12 dependencies.
func NewApplication(dependencies ApplicationDependencies) (*Service, error) {
	key := dependencies.RateLimitKey.Copy()
	defer clear(key)
	if nilDeviceauthValue(dependencies.Repository) || nilDeviceauthValue(dependencies.IdentityParticipant) ||
		nilDeviceauthValue(dependencies.Protector) || nilDeviceauthValue(dependencies.Random) || nilDeviceauthValue(dependencies.Clock) ||
		nilDeviceauthValue(dependencies.Limiter) || nilDeviceauthValue(dependencies.ChallengeStore) || len(key) != sha256.Size ||
		!validPublicOrigin(dependencies.Security.PublicBaseURL) || dependencies.Security.RequestDeadline < 2*time.Second ||
		dependencies.Security.RequestDeadline > 10*time.Second || dependencies.Security.ChallengeRateLimit.Limit == 0 ||
		dependencies.Security.ChallengeRateLimit.Window <= 0 {
		return nil, ErrInvalidApplication
	}
	return &Service{
		repository: dependencies.Repository, identity: dependencies.IdentityParticipant, protector: dependencies.Protector,
		random: dependencies.Random, clock: dependencies.Clock, limiter: dependencies.Limiter,
		challenges: dependencies.ChallengeStore, rateLimitKey: secret.NewBytes(key), security: dependencies.Security,
	}, nil
}

// CreateChallenge validates an enrollment grant before applying the limiter and creating Redis state.
func (service *Service) CreateChallenge(ctx context.Context, command CreateChallengeCommand) (result Challenge, resultErr error) {
	defer func() {
		if recover() != nil {
			result = Challenge{}
			resultErr = deviceDependencyUnavailable()
		}
	}()
	if !validDeviceauthService(service) || nilDeviceauthValue(ctx) {
		return Challenge{}, deviceDependencyUnavailable()
	}
	if command.Kind == ChallengeRotation {
		return service.createRotationChallenge(ctx, command)
	}
	if !validRegistrationChallengeCommand(command) {
		return Challenge{}, deviceAuthenticationFailed()
	}
	grantDigest := securitykit.DigestToken(securitykit.EnrollmentGrantToken, command.EnrollmentGrant)
	if grantDigest == [32]byte{} {
		return Challenge{}, deviceAuthenticationFailed()
	}
	defer clear(grantDigest[:])
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	now, ok := service.now()
	if !ok {
		return Challenge{}, deviceDependencyUnavailable()
	}
	validationErr := service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, transaction Transaction) error {
		if nilDeviceauthValue(transaction) || nilDeviceauthValue(transaction.DBTX()) {
			return deviceDependencyUnavailable()
		}
		_, found, err := service.identity.ValidateDeviceEnrollment(transactionContext, transaction.DBTX(), grantDigest, now)
		if err != nil {
			return deviceDependencyUnavailable()
		}
		if !found {
			return deviceAuthenticationFailed()
		}
		return errChallengeValidationRollback
	})
	if !errors.Is(validationErr, errChallengeValidationRollback) {
		if validationErr != nil {
			return Challenge{}, validationErr
		}
		return Challenge{}, deviceDependencyUnavailable()
	}
	subject, err := ratelimit.SubjectDigest(
		service.rateLimitKey, ratelimit.Challenge, now, service.security.ChallengeRateLimit, hex.EncodeToString(grantDigest[:]),
	)
	if err != nil || subject == [32]byte{} {
		return Challenge{}, deviceDependencyUnavailable()
	}
	defer clear(subject[:])
	allowed, err := safeDeviceLimiterAllow(operationContext, service.limiter, subject, service.security.ChallengeRateLimit)
	if err != nil || operationContext.Err() != nil {
		return Challenge{}, deviceDependencyUnavailable()
	}
	if !allowed {
		return Challenge{}, apierrors.NewRetryAfter(apierrors.RateLimited, apierrors.Retry, time.Second)
	}
	contextDigest, err := registrationContextDigest(command.RequestNonce, command.SigningPublicKey, command.HPKEPublicKey, service.security.PublicBaseURL)
	if err != nil {
		return Challenge{}, deviceDependencyUnavailable()
	}
	defer clear(contextDigest[:])
	challengeID, err := service.randomUUID()
	if err != nil {
		return Challenge{}, deviceDependencyUnavailable()
	}
	challengeBytes, err := service.randomNonzero32()
	if err != nil {
		return Challenge{}, deviceDependencyUnavailable()
	}
	expiresAt := now.Add(deviceChallengeTTL)
	record := ChallengeRecord{
		ChallengeID: challengeID.String(), Kind: ChallengeRegistration, ProtocolVersion: deviceProofProtocolVersion,
		Operation: registerDeviceOperation, Challenge: challengeBytes, GrantDigest: grantDigest,
		ContextDigest: contextDigest, ExpiresAt: expiresAt,
	}
	if err := safeDeviceChallengeCreate(operationContext, service.challenges, record); err != nil {
		return Challenge{}, deviceDependencyUnavailable()
	}
	return Challenge{ChallengeID: challengeID.String(), Challenge: challengeBytes, ExpiresAt: expiresAt}, nil
}

func (service *Service) createRotationChallenge(ctx context.Context, command CreateChallengeCommand) (Challenge, error) {
	if !validRotationChallengeCommand(command) {
		return Challenge{}, deviceAuthenticationFailed()
	}
	refreshDigest := securitykit.DigestToken(securitykit.DeviceRefreshToken, command.RefreshToken)
	if refreshDigest == [32]byte{} {
		return Challenge{}, deviceAuthenticationFailed()
	}
	defer clear(refreshDigest[:])
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	now, ok := service.now()
	if !ok {
		return Challenge{}, deviceDependencyUnavailable()
	}
	var familyID uuid.UUID
	validationErr := service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, transaction Transaction) error {
		if nilDeviceauthValue(transaction) || nilDeviceauthValue(transaction.DBTX()) {
			return deviceDependencyUnavailable()
		}
		queries := store.New(transaction.DBTX())
		digestCopy := append([]byte(nil), refreshDigest[:]...)
		defer clear(digestCopy)
		discovered, err := queries.DiscoverDeviceRefreshToken(transactionContext, digestCopy)
		defer clearDeviceRefreshDiscovery(&discovered)
		if errors.Is(err, pgx.ErrNoRows) {
			return deviceAuthenticationFailed()
		}
		if err != nil {
			return deviceDependencyUnavailable()
		}
		lockedRefresh, err := queries.LockDeviceFamilyRefreshTokens(transactionContext, discovered.FamilyID)
		defer clearDeviceRefreshRows(lockedRefresh)
		if err != nil {
			return deviceDependencyUnavailable()
		}
		family, err := queries.GetDeviceTokenFamilyForUpdate(transactionContext, discovered.FamilyID)
		defer clear(family.AccessTokenHash)
		if errors.Is(err, pgx.ErrNoRows) {
			return deviceAuthenticationFailed()
		}
		if err != nil {
			return deviceDependencyUnavailable()
		}
		authorization, err := queries.GetAuthorizationForUpdate(transactionContext, discovered.AuthorizationID)
		if errors.Is(err, pgx.ErrNoRows) {
			return deviceAuthenticationFailed()
		}
		if err != nil {
			return deviceDependencyUnavailable()
		}
		device, err := queries.GetDeviceForUpdate(transactionContext, discovered.DeviceID)
		defer clearDeviceRowSecrets(&device)
		if errors.Is(err, pgx.ErrNoRows) {
			return deviceAuthenticationFailed()
		}
		if err != nil {
			return deviceDependencyUnavailable()
		}
		accountActive, err := service.identity.ValidateDeviceAccountAuthority(
			transactionContext, transaction.DBTX(), identity.PrincipalID(discovered.PrincipalID.String()), now,
		)
		if err != nil {
			return deviceDependencyUnavailable()
		}
		freshRefresh, err := queries.ListDeviceFamilyRefreshTokens(transactionContext, discovered.FamilyID)
		defer clearDeviceRefreshRows(freshRefresh)
		if err != nil {
			return deviceDependencyUnavailable()
		}
		if !accountActive || !validRotationChallengeAuthority(discovered, family, authorization, device, lockedRefresh, freshRefresh, refreshDigest, now) {
			return deviceAuthenticationFailed()
		}
		familyID = discovered.FamilyID
		return errRotationChallengeRollback
	})
	if !errors.Is(validationErr, errRotationChallengeRollback) {
		if validationErr != nil {
			return Challenge{}, validationErr
		}
		return Challenge{}, deviceDependencyUnavailable()
	}
	subject, err := ratelimit.SubjectDigest(
		service.rateLimitKey, ratelimit.Challenge, now, service.security.ChallengeRateLimit, familyID.String(),
	)
	if err != nil || subject == [32]byte{} {
		return Challenge{}, deviceDependencyUnavailable()
	}
	defer clear(subject[:])
	allowed, err := safeDeviceLimiterAllow(operationContext, service.limiter, subject, service.security.ChallengeRateLimit)
	if err != nil || operationContext.Err() != nil {
		return Challenge{}, deviceDependencyUnavailable()
	}
	if !allowed {
		return Challenge{}, apierrors.NewRetryAfter(apierrors.RateLimited, apierrors.Retry, time.Second)
	}
	contextDigest, err := rotationContextDigest(familyID, command.RequestNonce, service.security.PublicBaseURL)
	if err != nil {
		return Challenge{}, deviceDependencyUnavailable()
	}
	defer clear(contextDigest[:])
	challengeID, err := service.randomUUID()
	if err != nil {
		return Challenge{}, deviceDependencyUnavailable()
	}
	challengeBytes, err := service.randomNonzero32()
	if err != nil {
		return Challenge{}, deviceDependencyUnavailable()
	}
	expiresAt := now.Add(deviceChallengeTTL)
	record := ChallengeRecord{
		ChallengeID: challengeID.String(), Kind: ChallengeRotation, ProtocolVersion: deviceRotationProtocolVersion,
		Operation: rotateDeviceTokenOperation, Challenge: challengeBytes, GrantDigest: refreshDigest,
		ContextDigest: contextDigest, ExpiresAt: expiresAt,
	}
	if err := safeDeviceChallengeCreate(operationContext, service.challenges, record); err != nil {
		return Challenge{}, deviceDependencyUnavailable()
	}
	return Challenge{ChallengeID: challengeID.String(), Challenge: challengeBytes, ExpiresAt: expiresAt}, nil
}

// RegisterDevice consumes a one-time challenge and creates the exact device graph in one final transaction.
func (service *Service) RegisterDevice(ctx context.Context, command RegisterDeviceCommand) (result DeviceTokens, resultErr error) {
	defer func() {
		if recover() != nil {
			result.AccessToken.Clear()
			result.RefreshToken.Clear()
			result = DeviceTokens{}
			resultErr = deviceDependencyUnavailable()
		}
	}()
	if !validDeviceauthService(service) || nilDeviceauthValue(ctx) || !validRegisterDeviceCommand(command) {
		return DeviceTokens{}, deviceMalformedRequest()
	}
	grantDigest := securitykit.DigestToken(securitykit.EnrollmentGrantToken, command.EnrollmentGrant)
	if grantDigest == [32]byte{} {
		return DeviceTokens{}, deviceAuthenticationFailed()
	}
	defer clear(grantDigest[:])
	privateBinding, err := privateRegistrationBinding(service.protector, grantDigest, command, service.security.PublicBaseURL)
	if err != nil {
		return DeviceTokens{}, deviceDependencyUnavailable()
	}
	defer clear(privateBinding)
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	now, ok := service.now()
	if !ok {
		return DeviceTokens{}, deviceDependencyUnavailable()
	}
	scope := idempotency.AnonymousDeviceRegistrationScope()
	var preflightAuthority identity.DeviceEnrollmentAuthority
	var replayTokens DeviceTokens
	defer replayTokens.AccessToken.Clear()
	defer replayTokens.RefreshToken.Clear()
	var replayed bool
	preflightErr := service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, transaction Transaction) error {
		record, outcome, err := transaction.BeginIdempotency(
			transactionContext, scope, command.IdempotencyKey, privateBinding, now, now.Add(registrationIdempotencyRetention),
		)
		if err != nil {
			return deviceDependencyUnavailable()
		}
		switch outcome {
		case idempotency.Replay:
			decoded, err := tokensFromReplayRecord(record)
			if err != nil {
				return deviceDependencyUnavailable()
			}
			replayTokens = takeDeviceTokens(&decoded)
			replayed = true
			return nil
		case idempotency.Conflict:
			return apierrors.New(apierrors.IdempotencyConflict, apierrors.ContactSupport)
		case idempotency.InProgress:
			return apierrors.NewRetryAfter(apierrors.StateConflict, apierrors.Retry, time.Second)
		case idempotency.Started:
		default:
			return deviceDependencyUnavailable()
		}
		authority, found, err := service.identity.ValidateDeviceEnrollment(transactionContext, transaction.DBTX(), grantDigest, now)
		if err != nil {
			return deviceDependencyUnavailable()
		}
		if !found {
			return deviceAuthenticationFailed()
		}
		preflightAuthority = authority
		return errRegistrationPreflightRollback
	})
	if replayed && preflightErr == nil {
		return takeDeviceTokens(&replayTokens), nil
	}
	if !errors.Is(preflightErr, errRegistrationPreflightRollback) {
		if preflightErr != nil {
			return DeviceTokens{}, preflightErr
		}
		return DeviceTokens{}, deviceDependencyUnavailable()
	}
	contextDigest, err := registrationContextDigest(command.RequestNonce, command.SigningPublicKey, command.HPKEPublicKey, service.security.PublicBaseURL)
	if err != nil {
		return DeviceTokens{}, deviceDependencyUnavailable()
	}
	defer clear(contextDigest[:])
	record, err := safeDeviceChallengeConsume(operationContext, service.challenges, command.ChallengeID, grantDigest, contextDigest)
	if err != nil {
		if errors.Is(err, ErrChallengeNotFound) || errors.Is(err, ErrInvalidChallenge) {
			return DeviceTokens{}, deviceAuthenticationFailed()
		}
		return DeviceTokens{}, deviceDependencyUnavailable()
	}
	if !record.ExpiresAt.After(now) || record.Kind != ChallengeRegistration || record.ProtocolVersion != deviceProofProtocolVersion ||
		record.Operation != registerDeviceOperation {
		return DeviceTokens{}, deviceAuthenticationFailed()
	}
	if err := verifyProof(ProofInput{
		ProtocolVersion: record.ProtocolVersion, Challenge: record.Challenge, GrantDigest: grantDigest,
		SigningPublicKey: command.SigningPublicKey, HPKEPublicKey: command.HPKEPublicKey,
		Operation: record.Operation, Audience: service.security.PublicBaseURL, RequestNonce: command.RequestNonce,
	}, command.Signature, service.security.PublicBaseURL); err != nil {
		return DeviceTokens{}, deviceAuthenticationFailed()
	}
	displayPlaintext := []byte(command.DisplayName)
	defer clear(displayPlaintext)
	displayField, err := service.protector.Encrypt(displayNameProtectionDomain, displayPlaintext)
	if err != nil || displayField.KeyVersion == 0 || displayField.KeyVersion > math.MaxInt32 ||
		len(displayField.Ciphertext) < 29 || len(displayField.Ciphertext) > maximumDisplayCiphertextBytes {
		clear(displayField.Ciphertext)
		return DeviceTokens{}, deviceDependencyUnavailable()
	}
	defer clear(displayField.Ciphertext)
	created, err := service.prepareRegistration(command, preflightAuthority, displayField, now)
	if err != nil {
		return DeviceTokens{}, err
	}
	defer created.clear()

	finalErr := service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, transaction Transaction) error {
		idempotencyRecord, outcome, err := transaction.BeginIdempotency(
			transactionContext, scope, command.IdempotencyKey, privateBinding, now, now.Add(registrationIdempotencyRetention),
		)
		if err != nil {
			return deviceDependencyUnavailable()
		}
		switch outcome {
		case idempotency.Replay:
			decoded, err := tokensFromReplayRecord(idempotencyRecord)
			if err != nil {
				return deviceDependencyUnavailable()
			}
			replayTokens = takeDeviceTokens(&decoded)
			replayed = true
			return nil
		case idempotency.Conflict:
			return apierrors.New(apierrors.IdempotencyConflict, apierrors.ContactSupport)
		case idempotency.InProgress:
			return apierrors.NewRetryAfter(apierrors.StateConflict, apierrors.Retry, time.Second)
		case idempotency.Started:
		default:
			return deviceDependencyUnavailable()
		}
		authority, found, err := service.identity.ValidateDeviceEnrollment(transactionContext, transaction.DBTX(), grantDigest, now)
		if err != nil {
			return deviceDependencyUnavailable()
		}
		if !found || !sameEnrollmentAuthority(authority, preflightAuthority) {
			return deviceAuthenticationFailed()
		}
		if err := transaction.CreateDevice(transactionContext, created.deviceParams); err != nil {
			return deviceDependencyUnavailable()
		}
		consumed, found, err := transaction.ConsumeEnrollmentGrant(transactionContext, grantDigest, created.deviceID, now)
		defer clear(consumed.TokenHash)
		if err != nil {
			return deviceDependencyUnavailable()
		}
		if !found || !consumedGrantMatchesAuthority(consumed, authority, grantDigest, created.deviceID) {
			return deviceAuthenticationFailed()
		}
		if err := transaction.CreateDeviceAuthorization(transactionContext, created.authorizationParams); err != nil {
			return deviceDependencyUnavailable()
		}
		if err := service.identity.BindSessionToAuthorization(
			transactionContext, transaction.DBTX(), authority.SessionID(), created.authorizationID, now,
		); err != nil {
			return deviceDependencyUnavailable()
		}
		if err := transaction.CreateDevicePolicySnapshot(transactionContext, created.policyParams); err != nil {
			return deviceDependencyUnavailable()
		}
		if err := transaction.CreateDeviceTokenFamily(transactionContext, created.familyParams); err != nil {
			return deviceDependencyUnavailable()
		}
		if err := transaction.InsertDeviceRefreshToken(transactionContext, created.refreshParams); err != nil {
			return deviceDependencyUnavailable()
		}
		if err := transaction.AppendEvent(transactionContext, created.event); err != nil {
			return deviceDependencyUnavailable()
		}
		if err := transaction.CompleteIdempotency(transactionContext, idempotencyRecord, registrationResponseStatus, created.replayBody); err != nil {
			return deviceDependencyUnavailable()
		}
		return nil
	})
	if finalErr != nil {
		return DeviceTokens{}, finalErr
	}
	if replayed {
		return takeDeviceTokens(&replayTokens), nil
	}
	return created.takeTokens(), nil
}

type preparedRegistration struct {
	deviceID, authorizationID uuid.UUID
	tokens                    DeviceTokens
	deviceParams              store.CreateDeviceParams
	authorizationParams       store.CreateDeviceAuthorizationParams
	policyParams              store.CreateDevicePolicySnapshotParams
	familyParams              store.CreateDeviceTokenFamilyParams
	refreshParams             store.InsertDeviceRefreshTokenParams
	event                     *eventsv1.EventEnvelope
	replayBody                []byte
}

func (service *Service) prepareRegistration(
	command RegisterDeviceCommand,
	authority identity.DeviceEnrollmentAuthority,
	displayField sensitive.EncryptedField,
	now time.Time,
) (preparedRegistration, error) {
	principalID, err := canonicalDeviceauthUUID(string(authority.PrincipalID()))
	if err != nil {
		return preparedRegistration{}, deviceDependencyUnavailable()
	}
	deviceID, err := service.randomUUID()
	if err != nil {
		return preparedRegistration{}, deviceDependencyUnavailable()
	}
	authorizationID, err := service.randomUUID()
	if err != nil {
		return preparedRegistration{}, deviceDependencyUnavailable()
	}
	familyID, err := service.randomUUID()
	if err != nil {
		return preparedRegistration{}, deviceDependencyUnavailable()
	}
	eventID, err := service.randomUUID()
	if err != nil {
		return preparedRegistration{}, deviceDependencyUnavailable()
	}
	access, err := securitykit.NewOpaqueToken(service.random)
	if err != nil {
		return preparedRegistration{}, deviceDependencyUnavailable()
	}
	defer access.Clear()
	refresh, err := securitykit.NewOpaqueToken(service.random)
	if err != nil {
		return preparedRegistration{}, deviceDependencyUnavailable()
	}
	defer refresh.Clear()
	accessDigest := securitykit.DigestToken(securitykit.DeviceAccessToken, access)
	refreshDigest := securitykit.DigestToken(securitykit.DeviceRefreshToken, refresh)
	if accessDigest == [32]byte{} || refreshDigest == [32]byte{} {
		clear(accessDigest[:])
		clear(refreshDigest[:])
		return preparedRegistration{}, deviceDependencyUnavailable()
	}
	defer clear(accessDigest[:])
	defer clear(refreshDigest[:])
	state, policy, provisional, err := policyForAuthority(authority, now)
	if err != nil {
		clear(policy)
		return preparedRegistration{}, err
	}
	defer clear(policy)
	tokens := DeviceTokens{
		DeviceID: deviceID, AuthorizationID: authorizationID, AccessToken: access, RefreshToken: refresh,
		AccessExpiresAt: now.Add(deviceAccessTTL), RefreshIdleExpiresAt: now.Add(deviceRefreshIdleTTL),
		RefreshAbsoluteExpiresAt: now.Add(deviceRefreshAbsoluteTTL),
	}
	replayBody, err := encodeDeviceTokens(tokens)
	if err != nil {
		return preparedRegistration{}, deviceDependencyUnavailable()
	}
	payload, err := events.MarshalPayload(events.DeviceAuthorizationChangedType, &deviceauthv1.DeviceAuthorizationChanged{
		PrincipalId: principalID.String(), DeviceId: deviceID.String(), AuthorizationId: authorizationID.String(), State: state, Version: 1,
	})
	if err != nil {
		clear(replayBody)
		return preparedRegistration{}, deviceDependencyUnavailable()
	}
	var displayNameKeyVersion pgtype.Int4
	if err := displayNameKeyVersion.Scan(int64(displayField.KeyVersion)); err != nil || !displayNameKeyVersion.Valid {
		clear(replayBody)
		return preparedRegistration{}, deviceDependencyUnavailable()
	}
	event := &eventsv1.EventEnvelope{
		EventId: eventID.String(), EventType: events.DeviceAuthorizationChangedType, OccurredAt: timestamppb.New(now),
		Producer: "deviceauth", AggregateType: "device_authorization", AggregateId: authorizationID.String(), AggregateVersion: 1,
		IdempotencyKey: command.IdempotencyKey, Payload: payload,
	}
	tokens.AccessToken = access.Take()
	tokens.RefreshToken = refresh.Take()
	return preparedRegistration{
		deviceID: deviceID, authorizationID: authorizationID, tokens: tokens,
		deviceParams: store.CreateDeviceParams{
			ID: deviceID, PrincipalID: principalID, DisplayNameCiphertext: bytes.Clone(displayField.Ciphertext),
			DisplayNameKeyVersion: displayNameKeyVersion,
			SigningPublicKey:      append([]byte(nil), command.SigningPublicKey[:]...), HpkePublicKey: append([]byte(nil), command.HPKEPublicKey[:]...),
			KeyVersion: 1, CreatedAt: now,
		},
		authorizationParams: store.CreateDeviceAuthorizationParams{
			ID: authorizationID, PrincipalID: principalID, DeviceID: deviceID, State: state, ProvisionalUntil: provisional, CreatedAt: now,
		},
		policyParams: store.CreateDevicePolicySnapshotParams{AuthorizationID: authorizationID, Policy: bytes.Clone(policy), CreatedAt: now},
		familyParams: store.CreateDeviceTokenFamilyParams{
			ID: familyID, AuthorizationID: authorizationID, AccessTokenHash: append([]byte(nil), accessDigest[:]...),
			AccessExpiresAt: tokens.AccessExpiresAt, IdleExpiresAt: tokens.RefreshIdleExpiresAt,
			AbsoluteExpiresAt: tokens.RefreshAbsoluteExpiresAt, CreatedAt: now,
		},
		refreshParams: store.InsertDeviceRefreshTokenParams{
			TokenHash: append([]byte(nil), refreshDigest[:]...), FamilyID: familyID, PreviousTokenHash: nil, IssuedAt: now,
		},
		event: event, replayBody: replayBody,
	}, nil
}

func (created *preparedRegistration) takeTokens() DeviceTokens {
	if created == nil {
		return DeviceTokens{}
	}
	return takeDeviceTokens(&created.tokens)
}

func (created *preparedRegistration) clear() {
	if created == nil {
		return
	}
	created.tokens.AccessToken.Clear()
	created.tokens.RefreshToken.Clear()
	clear(created.deviceParams.DisplayNameCiphertext)
	clear(created.deviceParams.SigningPublicKey)
	clear(created.deviceParams.HpkePublicKey)
	clear(created.policyParams.Policy)
	clear(created.familyParams.AccessTokenHash)
	clear(created.refreshParams.TokenHash)
	clear(created.refreshParams.PreviousTokenHash)
	clear(created.replayBody)
}

func takeDeviceTokens(source *DeviceTokens) DeviceTokens {
	if source == nil {
		return DeviceTokens{}
	}
	result := DeviceTokens{
		DeviceID: source.DeviceID, AuthorizationID: source.AuthorizationID,
		AccessExpiresAt: source.AccessExpiresAt, RefreshIdleExpiresAt: source.RefreshIdleExpiresAt,
		RefreshAbsoluteExpiresAt: source.RefreshAbsoluteExpiresAt,
	}
	result.AccessToken = source.AccessToken.Take()
	result.RefreshToken = source.RefreshToken.Take()
	*source = DeviceTokens{}
	return result
}

func registrationContextDigest(requestNonce, signingPublicKey, hpkePublicKey [32]byte, audience string) ([32]byte, error) {
	if requestNonce == [32]byte{} || signingPublicKey == [32]byte{} || hpkePublicKey == [32]byte{} || !validPublicOrigin(audience) {
		return [32]byte{}, ErrInvalidApplication
	}
	material := make([]byte, 0, len(registrationContextDomain)+12+len(ChallengeRegistration)+len(deviceProofProtocolVersion)+len(registerDeviceOperation)+len(audience)+96)
	material = append(material, registrationContextDomain...)
	material = appendRegistrationFrame(material, []byte(ChallengeRegistration))
	material = appendRegistrationFrame(material, []byte(deviceProofProtocolVersion))
	material = appendRegistrationFrame(material, []byte(registerDeviceOperation))
	material = appendRegistrationFrame(material, []byte(audience))
	material = append(material, requestNonce[:]...)
	material = append(material, signingPublicKey[:]...)
	material = append(material, hpkePublicKey[:]...)
	digest := sha256.Sum256(material)
	clear(material)
	return digest, nil
}

func rotationContextDigest(familyID uuid.UUID, requestNonce [32]byte, audience string) ([32]byte, error) {
	if familyID == uuid.Nil || requestNonce == [32]byte{} || !validPublicOrigin(audience) {
		return [32]byte{}, ErrInvalidApplication
	}
	material := make([]byte, 0, len(rotationContextDomain)+16+requestNonceSize+16+len(ChallengeRotation)+len(deviceRotationProtocolVersion)+len(rotateDeviceTokenOperation)+len(audience))
	material = append(material, rotationContextDomain...)
	material = appendRegistrationFrame(material, []byte(ChallengeRotation))
	material = appendRegistrationFrame(material, []byte(deviceRotationProtocolVersion))
	material = appendRegistrationFrame(material, []byte(rotateDeviceTokenOperation))
	material = appendRegistrationFrame(material, []byte(audience))
	material = append(material, familyID[:]...)
	material = append(material, requestNonce[:]...)
	digest := sha256.Sum256(material)
	clear(material)
	return digest, nil
}

const requestNonceSize = 32

func validRotationChallengeAuthority(
	discovered store.DiscoverDeviceRefreshTokenRow,
	family store.DeviceauthDeviceTokenFamily,
	authorization store.DeviceauthDeviceAuthorization,
	device store.DeviceauthDevice,
	lockedRefresh, freshRefresh []store.DeviceauthDeviceRefreshToken,
	digest [32]byte,
	now time.Time,
) bool {
	return len(discovered.TokenHash) == len(digest) && subtle.ConstantTimeCompare(discovered.TokenHash, digest[:]) == 1 &&
		discovered.FamilyID != uuid.Nil && discovered.AuthorizationID != uuid.Nil && discovered.PrincipalID != uuid.Nil && discovered.DeviceID != uuid.Nil &&
		discovered.RefreshState == "active" && discovered.FamilyState == "active" && discovered.AuthorizationState == "active" && discovered.DeviceState == "active" &&
		discovered.IdleExpiresAt.After(now) && discovered.AbsoluteExpiresAt.After(now) && len(discovered.SigningPublicKey) == 32 && discovered.KeyVersion > 0 &&
		family.ID == discovered.FamilyID && family.AuthorizationID == discovered.AuthorizationID && family.State == "active" &&
		family.IdleExpiresAt.After(now) && family.AbsoluteExpiresAt.After(now) &&
		authorization.ID == discovered.AuthorizationID && authorization.PrincipalID == discovered.PrincipalID &&
		authorization.DeviceID == discovered.DeviceID && authorization.State == "active" &&
		device.ID == discovered.DeviceID && device.PrincipalID == discovered.PrincipalID && device.State == "active" &&
		len(device.SigningPublicKey) == 32 && device.KeyVersion == discovered.KeyVersion &&
		subtle.ConstantTimeCompare(device.SigningPublicKey, discovered.SigningPublicKey) == 1 &&
		validStableDeviceRefreshSet(lockedRefresh, discovered.FamilyID, digest) && sameDeviceRefreshSet(lockedRefresh, freshRefresh)
}

func validStableDeviceRefreshSet(rows []store.DeviceauthDeviceRefreshToken, familyID uuid.UUID, activeDigest [32]byte) bool {
	if len(rows) == 0 || familyID == uuid.Nil || activeDigest == [32]byte{} {
		return false
	}
	activeFound := false
	for index := range rows {
		if rows[index].FamilyID != familyID || len(rows[index].TokenHash) != sha256.Size ||
			(index > 0 && bytes.Compare(rows[index-1].TokenHash, rows[index].TokenHash) >= 0) {
			return false
		}
		if subtle.ConstantTimeCompare(rows[index].TokenHash, activeDigest[:]) == 1 {
			if activeFound || rows[index].State != "active" {
				return false
			}
			activeFound = true
		}
	}
	return activeFound
}

func sameDeviceRefreshSet(left, right []store.DeviceauthDeviceRefreshToken) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].FamilyID != right[index].FamilyID || left[index].State != right[index].State ||
			len(left[index].TokenHash) != len(right[index].TokenHash) || subtle.ConstantTimeCompare(left[index].TokenHash, right[index].TokenHash) != 1 {
			return false
		}
	}
	return true
}

func clearDeviceRefreshDiscovery(row *store.DiscoverDeviceRefreshTokenRow) {
	if row == nil {
		return
	}
	clear(row.TokenHash)
	clear(row.PreviousTokenHash)
	clear(row.SigningPublicKey)
}

func clearDeviceRefreshRows(rows []store.DeviceauthDeviceRefreshToken) {
	for index := range rows {
		clear(rows[index].TokenHash)
		clear(rows[index].PreviousTokenHash)
	}
}

func clearDeviceRowSecrets(row *store.DeviceauthDevice) {
	if row == nil {
		return
	}
	clear(row.DisplayNameCiphertext)
	clear(row.SigningPublicKey)
	clear(row.HpkePublicKey)
}

func privateRegistrationBinding(
	protector sensitive.Protector,
	grantDigest [32]byte,
	command RegisterDeviceCommand,
	audience string,
) ([]byte, error) {
	if nilDeviceauthValue(protector) || grantDigest == [32]byte{} || !validRegisterDeviceCommand(command) || !validPublicOrigin(audience) {
		return nil, ErrInvalidApplication
	}
	return withOwnedDisplayName(command.DisplayName, func(displayBytes []byte) ([]byte, error) {
		contextDigest, err := registrationContextDigest(command.RequestNonce, command.SigningPublicKey, command.HPKEPublicKey, audience)
		if err != nil {
			return nil, ErrInvalidApplication
		}
		defer clear(contextDigest[:])
		challengeIDBytes := []byte(command.ChallengeID)
		defer clear(challengeIDBytes)
		kindBytes := []byte(ChallengeRegistration)
		defer clear(kindBytes)
		protocolBytes := []byte(deviceProofProtocolVersion)
		defer clear(protocolBytes)
		operationBytes := []byte(registerDeviceOperation)
		defer clear(operationBytes)
		audienceBytes := []byte(audience)
		defer clear(audienceBytes)
		parts := [][]byte{
			grantDigest[:], challengeIDBytes, kindBytes, protocolBytes, operationBytes, audienceBytes, command.RequestNonce[:], command.SigningPublicKey[:],
			command.HPKEPublicKey[:], displayBytes, command.Signature[:], contextDigest[:],
		}
		material := make([]byte, 0, 512+len(command.ChallengeID)+len(displayBytes)+len(audience))
		for _, part := range parts {
			material = appendRegistrationFrame(material, part)
		}
		defer clear(material)
		digest := protector.LookupDigest(privateRegistrationBindingDomain, material)
		if digest == [32]byte{} {
			return nil, ErrInvalidApplication
		}
		result := append([]byte(nil), digest[:]...)
		clear(digest[:])
		return result, nil
	})
}

func withOwnedDisplayName(value string, operation func([]byte) ([]byte, error)) ([]byte, error) {
	if operation == nil {
		return nil, ErrInvalidApplication
	}
	displayBytes := []byte(value)
	defer clear(displayBytes)
	return operation(displayBytes)
}

func appendRegistrationFrame(target, value []byte) []byte {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value))) // #nosec G115 -- every caller enforces small fixed bounds.
	target = append(target, length[:]...)
	clear(length[:])
	return append(target, value...)
}

func policyForAuthority(authority identity.DeviceEnrollmentAuthority, now time.Time) (string, []byte, sql.NullTime, error) {
	switch authority.PolicyMarker() {
	case "standard":
		if boundary, valid := authority.ProvisionalUntil(); valid || !boundary.IsZero() {
			return "", nil, sql.NullTime{}, deviceAuthenticationFailed()
		}
		return "active", []byte(`{"mode":"standard"}`), sql.NullTime{}, nil
	case "trial_restricted":
		boundary, valid := authority.ProvisionalUntil()
		if !valid || !boundary.After(now) || boundary.Location() != time.UTC || boundary.Nanosecond() != 0 {
			return "", nil, sql.NullTime{}, deviceAuthenticationFailed()
		}
		policy := []byte(`{"expires_at":"` + boundary.Format(time.RFC3339) + `","max_devices":"1","mode":"trial_restricted"}`)
		return "provisional", policy, sql.NullTime{Time: boundary, Valid: true}, nil
	default:
		return "", nil, sql.NullTime{}, deviceAuthenticationFailed()
	}
}

func sameEnrollmentAuthority(left, right identity.DeviceEnrollmentAuthority) bool {
	leftBoundary, leftValid := left.ProvisionalUntil()
	rightBoundary, rightValid := right.ProvisionalUntil()
	return left.PrincipalID() == right.PrincipalID() && left.SessionID() == right.SessionID() &&
		left.PolicyMarker() == right.PolicyMarker() && leftValid == rightValid && leftBoundary.Equal(rightBoundary)
}

func consumedGrantMatchesAuthority(
	grant store.DeviceauthEnrollmentGrant,
	authority identity.DeviceEnrollmentAuthority,
	digest [32]byte,
	deviceID uuid.UUID,
) bool {
	principalID, principalErr := canonicalDeviceauthUUID(string(authority.PrincipalID()))
	sessionID, sessionErr := canonicalDeviceauthUUID(string(authority.SessionID()))
	boundary, boundaryValid := authority.ProvisionalUntil()
	return principalErr == nil && sessionErr == nil && grant.PrincipalID == principalID && grant.AccountSessionID == sessionID &&
		grant.PolicyMarker == authority.PolicyMarker() && grant.ProvisionalUntil.Valid == boundaryValid && grant.ProvisionalUntil.Time.Equal(boundary) &&
		grant.State == "consumed" && grant.ConsumedDeviceID.Valid && grant.ConsumedDeviceID.UUID == deviceID &&
		len(grant.TokenHash) == len(digest) && subtle.ConstantTimeCompare(grant.TokenHash, digest[:]) == 1
}

func validRegistrationChallengeCommand(command CreateChallengeCommand) bool {
	refresh := command.RefreshToken.Copy()
	grant := command.EnrollmentGrant.Copy()
	defer clear(refresh)
	defer clear(grant)
	return command.Kind == ChallengeRegistration && len(grant) == 32 && len(refresh) == 0 && command.RequestNonce != [32]byte{} &&
		command.SigningPublicKey != [32]byte{} && command.HPKEPublicKey != [32]byte{} && validDeviceauthIdempotencyKey(command.IdempotencyKey)
}

func validRotationChallengeCommand(command CreateChallengeCommand) bool {
	refresh := command.RefreshToken.Copy()
	grant := command.EnrollmentGrant.Copy()
	defer clear(refresh)
	defer clear(grant)
	return command.Kind == ChallengeRotation && len(refresh) == 32 && len(grant) == 0 && command.RequestNonce != [32]byte{} &&
		command.SigningPublicKey == [32]byte{} && command.HPKEPublicKey == [32]byte{} && validDeviceauthIdempotencyKey(command.IdempotencyKey)
}

func validRegisterDeviceCommand(command RegisterDeviceCommand) bool {
	grant := command.EnrollmentGrant.Copy()
	defer clear(grant)
	return len(grant) == 32 && validChallengeID(command.ChallengeID) && command.RequestNonce != [32]byte{} &&
		command.SigningPublicKey != [32]byte{} && command.HPKEPublicKey != [32]byte{} && command.Signature != [64]byte{} &&
		validDisplayName(command.DisplayName) && validDeviceauthIdempotencyKey(command.IdempotencyKey)
}

func validDisplayName(value string) bool {
	return len(value) >= 1 && len(value) <= maximumDisplayNameBytes && utf8.ValidString(value) &&
		utf8.RuneCountInString(value) >= 1 && utf8.RuneCountInString(value) <= maximumDisplayNameRunes
}

func validDeviceauthIdempotencyKey(value string) bool {
	_, err := idempotency.KeyDigest(value)
	return err == nil
}

func validDeviceauthService(service *Service) bool {
	return service != nil && !nilDeviceauthValue(service.repository) && !nilDeviceauthValue(service.identity) &&
		!nilDeviceauthValue(service.protector) && !nilDeviceauthValue(service.random) && !nilDeviceauthValue(service.clock) &&
		!nilDeviceauthValue(service.limiter) && !nilDeviceauthValue(service.challenges) && validPublicOrigin(service.security.PublicBaseURL)
}

func (service *Service) now() (result time.Time, ok bool) {
	defer func() {
		if recover() != nil {
			result = time.Time{}
			ok = false
		}
	}()
	now := service.clock.Now().UTC()
	return now, !now.IsZero() && now.Year() >= 2020 && now.Year() <= 2100
}

func (service *Service) randomUUID() (uuid.UUID, error) {
	var raw [16]byte
	if err := readDeviceauthRandom(service.random, raw[:]); err != nil {
		clear(raw[:])
		return uuid.Nil, err
	}
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80
	result, err := uuid.FromBytes(raw[:])
	clear(raw[:])
	if err != nil || result == uuid.Nil {
		return uuid.Nil, ErrInvalidApplication
	}
	return result, nil
}

func (service *Service) randomNonzero32() ([32]byte, error) {
	for range maximumEmptyRandomReads {
		var result [32]byte
		if err := readDeviceauthRandom(service.random, result[:]); err != nil {
			clear(result[:])
			return [32]byte{}, err
		}
		if result != [32]byte{} {
			return result, nil
		}
		clear(result[:])
	}
	return [32]byte{}, ErrInvalidApplication
}

type boundedDeviceauthRandom struct {
	source     securitykit.RandomSource
	emptyReads int
}

func (random *boundedDeviceauthRandom) Read(target []byte) (count int, err error) {
	defer func() {
		if recover() != nil {
			count = 0
			err = io.ErrNoProgress
		}
	}()
	count, err = random.source.Read(target)
	if count < 0 || count > len(target) {
		return 0, io.ErrNoProgress
	}
	if count == 0 && err == nil {
		random.emptyReads++
		if random.emptyReads >= maximumEmptyRandomReads {
			return 0, io.ErrNoProgress
		}
	} else {
		random.emptyReads = 0
	}
	return count, err
}

func readDeviceauthRandom(source securitykit.RandomSource, target []byte) error {
	if nilDeviceauthValue(source) || len(target) == 0 {
		return ErrInvalidApplication
	}
	bounded := &boundedDeviceauthRandom{source: source}
	count, err := io.ReadFull(bounded, target)
	if err != nil || count != len(target) {
		clear(target)
		return ErrInvalidApplication
	}
	return nil
}

func safeDeviceLimiterAllow(ctx context.Context, limiter ratelimit.Limiter, subject [32]byte, policy config.RateLimitPolicy) (allowed bool, err error) {
	defer func() {
		if recover() != nil {
			allowed = false
			err = ErrInvalidApplication
		}
	}()
	return limiter.Allow(ctx, ratelimit.Challenge, subject, policy)
}

func safeDeviceChallengeCreate(ctx context.Context, challengeStore ChallengeStore, record ChallengeRecord) (err error) {
	defer func() {
		if recover() != nil {
			err = ErrChallengeUnavailable
		}
	}()
	return challengeStore.Create(ctx, record, deviceChallengeTTL)
}

func safeDeviceChallengeConsume(ctx context.Context, challengeStore ChallengeStore, challengeID string, grantDigest, contextDigest [32]byte) (record ChallengeRecord, err error) {
	defer func() {
		if recover() != nil {
			record = ChallengeRecord{}
			err = ErrChallengeUnavailable
		}
	}()
	return challengeStore.Consume(ctx, challengeID, grantDigest, contextDigest)
}

type deviceTokenReplay struct {
	DeviceID                 string `json:"device_id"`
	AuthorizationID          string `json:"authorization_id"`
	AccessToken              string `json:"access_token"`  // #nosec G117 -- immediately encrypted by idempotency.
	RefreshToken             string `json:"refresh_token"` // #nosec G117 -- immediately encrypted by idempotency.
	AccessExpiresAt          string `json:"access_expires_at"`
	RefreshIdleExpiresAt     string `json:"refresh_idle_expires_at"`
	RefreshAbsoluteExpiresAt string `json:"refresh_absolute_expires_at"`
}

func encodeDeviceTokens(tokens DeviceTokens) ([]byte, error) {
	payload := deviceTokenReplay{
		DeviceID: tokens.DeviceID.String(), AuthorizationID: tokens.AuthorizationID.String(),
		AccessToken: securitykit.EncodeOpaqueToken(tokens.AccessToken), RefreshToken: securitykit.EncodeOpaqueToken(tokens.RefreshToken),
		AccessExpiresAt:          tokens.AccessExpiresAt.UTC().Format(time.RFC3339Nano),
		RefreshIdleExpiresAt:     tokens.RefreshIdleExpiresAt.UTC().Format(time.RFC3339Nano),
		RefreshAbsoluteExpiresAt: tokens.RefreshAbsoluteExpiresAt.UTC().Format(time.RFC3339Nano),
	}
	if tokens.DeviceID == uuid.Nil || tokens.AuthorizationID == uuid.Nil || payload.AccessToken == "" || payload.RefreshToken == "" {
		return nil, ErrInvalidApplication
	}
	type trustedReplayWire deviceTokenReplay
	return json.Marshal(trustedReplayWire(payload)) // #nosec G117 -- the private replay payload is immediately encrypted by idempotency.
}

func tokensFromReplayRecord(record idempotency.Record) (DeviceTokens, error) {
	return tokensFromReplayRecordWithStatus(record, registrationResponseStatus)
}

func tokensFromReplayRecordWithStatus(record idempotency.Record, expectedStatus int) (DeviceTokens, error) {
	if record.ResponseStatus() != expectedStatus {
		return DeviceTokens{}, ErrRepository
	}
	body, owned := record.TakeResponseBody()
	if !owned {
		return DeviceTokens{}, ErrRepository
	}
	defer clear(body)
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	type trustedReplayWire deviceTokenReplay
	var wire trustedReplayWire
	if err := decoder.Decode(&wire); err != nil || decoder.More() {
		return DeviceTokens{}, ErrRepository
	}
	payload := deviceTokenReplay(wire)
	deviceID, err := canonicalDeviceauthUUID(payload.DeviceID)
	if err != nil {
		return DeviceTokens{}, ErrRepository
	}
	authorizationID, err := canonicalDeviceauthUUID(payload.AuthorizationID)
	if err != nil {
		return DeviceTokens{}, ErrRepository
	}
	access, err := securitykit.DecodeOpaqueToken(payload.AccessToken)
	if err != nil {
		return DeviceTokens{}, ErrRepository
	}
	defer access.Clear()
	refresh, err := securitykit.DecodeOpaqueToken(payload.RefreshToken)
	if err != nil {
		return DeviceTokens{}, ErrRepository
	}
	defer refresh.Clear()
	accessAt, err := time.Parse(time.RFC3339Nano, payload.AccessExpiresAt)
	if err != nil {
		return DeviceTokens{}, ErrRepository
	}
	idleAt, err := time.Parse(time.RFC3339Nano, payload.RefreshIdleExpiresAt)
	if err != nil {
		return DeviceTokens{}, ErrRepository
	}
	absoluteAt, err := time.Parse(time.RFC3339Nano, payload.RefreshAbsoluteExpiresAt)
	if err != nil || !accessAt.Before(idleAt) || !idleAt.Before(absoluteAt) {
		return DeviceTokens{}, ErrRepository
	}
	return DeviceTokens{
		DeviceID: deviceID, AuthorizationID: authorizationID, AccessToken: access.Take(), RefreshToken: refresh.Take(),
		AccessExpiresAt: accessAt, RefreshIdleExpiresAt: idleAt, RefreshAbsoluteExpiresAt: absoluteAt,
	}, nil
}

func canonicalDeviceauthUUID(value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil || parsed.String() != value {
		return uuid.Nil, ErrRepository
	}
	return parsed, nil
}

func deviceMalformedRequest() error {
	return apierrors.New(apierrors.MalformedRequest, apierrors.ContactSupport)
}
func deviceAuthenticationFailed() error {
	return apierrors.New(apierrors.AuthenticationFailed, apierrors.Reenroll)
}
func deviceDependencyUnavailable() error {
	return apierrors.New(apierrors.DependencyUnavailable, apierrors.ContactSupport)
}

func nilDeviceauthValue(value any) bool {
	if value == nil {
		return true
	}
	representation := reflect.ValueOf(value)
	switch representation.Kind() { //nolint:exhaustive // Only nil-capable interfaces matter.
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return representation.IsNil()
	default:
		return false
	}
}
