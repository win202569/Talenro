package identity

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	identityv1 "talenro.local/platform/gen/go/talenro/identity/v1"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/config"
	contractevents "talenro.local/platform/internal/contracts/events"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/ratelimit"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

const (
	emailFieldDomain                = "identity/email/v1"
	emailVerificationDeliveryDomain = "identity/email-verification-delivery/v1"
	passwordResetDeliveryDomain     = "identity/password-reset-delivery/v1"
	passwordResetDigestDomain       = "TALENRO-PASSWORD-RESET-TOKEN-V1\x00"
	privateRequestDigestDomain      = "identity_private_request_v1"
	acceptedResponseBody            = `{"accepted":true}`
	ordinaryIdempotencyRetention    = 24 * time.Hour
	securityIdempotencyRetention    = 90 * 24 * time.Hour
	emailVerificationTTL            = 24 * time.Hour
	passwordResetTTL                = 30 * time.Minute
	enrollmentGrantTTL              = 10 * time.Minute
	provisionalAuthorizationTTL     = 24 * time.Hour
	accountAccessTTL                = 10 * time.Minute
	accountRefreshIdleTTL           = 30 * 24 * time.Hour
	accountRefreshAbsoluteTTL       = 90 * 24 * time.Hour
	maximumPendingDeliveryBytes     = 4096
)

// Service implements Task 9 account establishment and recovery operations.
type Service struct {
	repository   Repository
	protector    sensitive.Protector
	random       securitykit.RandomSource
	clock        securitykit.Clock
	rateLimiter  ratelimit.Limiter
	rateLimitKey secret.Bytes
	security     config.SecurityConfig
	participant  DeviceAuthorizationParticipant
	derive       PasswordDeriver
	newUUID      func() uuid.UUID
}

// NewApplication creates a production Task 9 service with real Argon2id and UUID generation.
func NewApplication(dependencies ApplicationDependencies) (*Service, error) {
	return newApplicationForTest(dependencies, deriveArgon2id, uuid.New)
}

func newApplicationForTest(dependencies ApplicationDependencies, derive PasswordDeriver, newUUID func() uuid.UUID) (*Service, error) {
	key := dependencies.RateLimitKey.Copy()
	defer clear(key)
	if nilIdentityValue(dependencies.Repository) || nilIdentityValue(dependencies.Protector) || nilIdentityValue(dependencies.Random) ||
		nilIdentityValue(dependencies.Clock) || nilIdentityValue(dependencies.Limiter) || nilIdentityValue(dependencies.DeviceAuthorizationParticipant) ||
		derive == nil || newUUID == nil || len(key) != sha256.Size || !validApplicationSecurity(dependencies.Security) {
		return nil, ErrInvalidApplication
	}
	return &Service{
		repository: dependencies.Repository, protector: dependencies.Protector, random: dependencies.Random, clock: dependencies.Clock,
		rateLimiter: dependencies.Limiter, rateLimitKey: secret.NewBytes(key), security: dependencies.Security,
		participant: dependencies.DeviceAuthorizationParticipant, derive: derive, newUUID: newUUID,
	}, nil
}

// RegisterAccount creates one private principal or returns the same generic duplicate result.
func (service *Service) RegisterAccount(ctx context.Context, command RegisterAccountCommand) (RegisterAccountResult, error) {
	if !validService(service) || nilIdentityValue(ctx) || !validIdentityLocale(command.Locale) || !validPasswordSecret(command.Password) || !validIdempotencyKeyForApplication(command.IdempotencyKey) {
		return RegisterAccountResult{}, malformedRequest()
	}
	canonicalEmail, canonicalErr := CanonicalizeEmail(command.Email)
	if canonicalErr != nil {
		return RegisterAccountResult{}, malformedRequest()
	}
	canonicalEmailBytes := canonicalEmail.Bytes()
	defer clear(canonicalEmailBytes)
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	requestPassword := command.Password.Copy()
	defer clear(requestPassword)
	canonicalRequest, requestErr := privateCanonicalRequest(service.protector, "register_account", canonicalEmailBytes, requestPassword, []byte(command.Locale))
	if requestErr != nil {
		return RegisterAccountResult{}, dependencyUnavailable()
	}
	defer clear(canonicalRequest)
	accepted := RegisterAccountResult{Accepted: true}
	err := service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, transaction Transaction) error {
		now, ok := service.now()
		if !ok {
			return dependencyUnavailable()
		}
		record, outcome, beginErr := transaction.BeginIdempotency(
			transactionContext, idempotency.AnonymousRegistrationScope(), command.IdempotencyKey, canonicalRequest, now, now.Add(ordinaryIdempotencyRetention),
		)
		if beginErr != nil {
			return dependencyUnavailable()
		}
		if replayErr := registrationIdempotencyOutcome(outcome, record); replayErr != nil || outcome != idempotency.Started {
			return replayErr
		}

		emailBytes := append([]byte(nil), canonicalEmailBytes...)
		defer clear(emailBytes)
		lookupDigest := service.protector.LookupDigest(emailFieldDomain, emailBytes)
		if lookupDigest == [32]byte{} {
			return dependencyUnavailable()
		}
		protectedEmail, protectErr := service.protector.Encrypt(emailFieldDomain, emailBytes)
		if protectErr != nil || !validProtectedField(protectedEmail, 2048) {
			clear(protectedEmail.Ciphertext)
			return dependencyUnavailable()
		}
		defer clear(protectedEmail.Ciphertext)
		if lockErr := transaction.LockEmailLookupDigest(transactionContext, lookupDigest[:]); lockErr != nil {
			return dependencyUnavailable()
		}
		_, exists, findErr := transaction.FindIdentityByLookupDigest(transactionContext, lookupDigest[:])
		if findErr != nil {
			return dependencyUnavailable()
		}
		if exists {
			if !service.performDummyPasswordWork() {
				return dependencyUnavailable()
			}
			if completeErr := transaction.CompleteIdempotency(transactionContext, record, 202, []byte(acceptedResponseBody)); completeErr != nil {
				return dependencyUnavailable()
			}
			return nil
		}

		password := command.Password.Copy()
		defer clear(password)
		credential, credentialErr := service.hashNewPassword(password)
		if credentialErr != nil {
			return dependencyUnavailable()
		}
		defer credential.clear()
		principalID := service.newUUID()
		emailIdentityID := service.newUUID()
		securityEventID := service.newUUID()
		if principalID == uuid.Nil || emailIdentityID == uuid.Nil || securityEventID == uuid.Nil {
			return dependencyUnavailable()
		}

		accountState := "pending_email"
		var verificationHash []byte
		var verificationExpires sql.NullTime
		var deliveryID uuid.NullUUID
		var deliveryCiphertext []byte
		var deliveryKeyVersion pgtype.Int4
		if service.security.EmailVerification == config.EmailDisabled {
			accountState = "active"
		} else {
			delivery, deliveryErr := service.newProtectedDelivery(canonicalEmail, VerifyEmailTemplate, command.Locale)
			if deliveryErr != nil {
				return dependencyUnavailable()
			}
			defer delivery.clear()
			verificationHash = append([]byte(nil), delivery.tokenDigest...)
			verificationExpires = sql.NullTime{Time: now.Add(emailVerificationTTL), Valid: true}
			deliveryID = uuid.NullUUID{UUID: delivery.id, Valid: true}
			deliveryCiphertext = append([]byte(nil), delivery.protected.Ciphertext...)
			storedKeyVersion, versionOK := protectedKeyVersionInt32(delivery.protected.KeyVersion)
			if !versionOK {
				return dependencyUnavailable()
			}
			deliveryKeyVersion = pgtype.Int4{Int32: storedKeyVersion, Valid: true}
		}
		defer clear(verificationHash)
		defer clear(deliveryCiphertext)

		if err := transaction.CreateAccount(transactionContext, store.CreateAccountParams{
			ID: principalID, State: accountState, StateVersion: 1, Locale: command.Locale, CreatedAt: now, UpdatedAt: now,
		}); err != nil {
			return dependencyUnavailable()
		}
		protectedEmailKeyVersion, versionOK := protectedKeyVersionInt32(protectedEmail.KeyVersion)
		if !versionOK {
			return dependencyUnavailable()
		}
		if err := transaction.CreateEmailIdentity(transactionContext, store.CreateEmailIdentityParams{
			ID: emailIdentityID, PrincipalID: principalID, LookupKeyVersion: protectedEmailKeyVersion, LookupDigest: lookupDigest[:],
			Ciphertext: append([]byte(nil), protectedEmail.Ciphertext...), EncryptionKeyVersion: protectedEmailKeyVersion,
			VerificationTokenHash: append([]byte(nil), verificationHash...), VerificationExpiresAt: verificationExpires,
			VerificationDeliveryID: deliveryID, VerificationDeliveryCiphertext: append([]byte(nil), deliveryCiphertext...),
			VerificationDeliveryKeyVersion: deliveryKeyVersion, CreatedAt: now,
		}); err != nil {
			return dependencyUnavailable()
		}
		if err := transaction.CreatePasswordCredential(transactionContext, credential.storeParams(principalID, now)); err != nil {
			return dependencyUnavailable()
		}
		if err := transaction.InsertSecurityEvent(transactionContext, store.InsertSecurityEventParams{
			ID: securityEventID, PrincipalID: uuid.NullUUID{UUID: principalID, Valid: true}, Category: "account_registered",
			Fingerprint: "identity.registration", AggregateVersion: 1, OccurredAt: now,
		}); err != nil {
			return dependencyUnavailable()
		}
		if deliveryID.Valid {
			envelope, eventErr := service.emailDeliveryEvent(deliveryID.UUID, VerifyEmailTemplate, command.Locale, command.IdempotencyKey, now)
			if eventErr != nil || transaction.AppendEvent(transactionContext, envelope) != nil {
				return dependencyUnavailable()
			}
		}
		if err := transaction.CompleteIdempotency(transactionContext, record, 202, []byte(acceptedResponseBody)); err != nil {
			return dependencyUnavailable()
		}
		return nil
	})
	if err != nil {
		return RegisterAccountResult{}, mapApplicationError(operationContext, err)
	}
	return accepted, nil
}

type passwordMaterial struct {
	policy PasswordPolicy
	salt   []byte
	hash   []byte
}

func (material passwordMaterial) storeParams(principalID uuid.UUID, now time.Time) store.CreatePasswordCredentialParams {
	policyVersion, memoryKiB, timeCost, parallelism, _ := passwordPolicyInt32(material.policy)
	return store.CreatePasswordCredentialParams{
		PrincipalID: principalID, PolicyVersion: policyVersion, MemoryKib: memoryKiB,
		TimeCost: timeCost, Parallelism: parallelism,
		Salt: append([]byte(nil), material.salt...), PasswordHash: append([]byte(nil), material.hash...), UpdatedAt: now,
	}
}

func (material *passwordMaterial) clear() {
	if material == nil {
		return
	}
	clear(material.salt)
	clear(material.hash)
}

func (service *Service) hashNewPassword(password []byte) (passwordMaterial, error) {
	if !validPassword(password) {
		return passwordMaterial{}, ErrInvalidPassword
	}
	policy := CurrentPasswordPolicy()
	salt := make([]byte, policy.SaltBytes)
	count, err := io.ReadFull(&boundedRandomSource{source: service.random}, salt)
	if err != nil || count != len(salt) {
		clear(salt)
		return passwordMaterial{}, ErrRandomSource
	}
	passwordCopy := append([]byte(nil), password...)
	defer clear(passwordCopy)
	derived := service.derive(passwordCopy, salt, policy)
	if len(derived) != int(policy.TagBytes) {
		clear(salt)
		clear(derived)
		return passwordMaterial{}, ErrInvalidPasswordCredential
	}
	hash := append([]byte(nil), derived...)
	clear(derived)
	return passwordMaterial{policy: policy, salt: salt, hash: hash}, nil
}

func (service *Service) performDummyPasswordWork() bool {
	policy := CurrentPasswordPolicy()
	password := append([]byte(nil), dummyPassword...)
	salt := append([]byte(nil), dummySalt[:]...)
	defer clear(password)
	defer clear(salt)
	derived := service.derive(password, salt, policy)
	valid := len(derived) == int(policy.TagBytes)
	clear(derived)
	return valid
}

type protectedDelivery struct {
	id          uuid.UUID
	tokenDigest []byte
	protected   sensitive.EncryptedField
}

func (delivery *protectedDelivery) clear() {
	if delivery == nil {
		return
	}
	clear(delivery.tokenDigest)
	clear(delivery.protected.Ciphertext)
}

func (service *Service) newProtectedDelivery(recipient CanonicalEmail, template TemplateID, locale string) (protectedDelivery, error) {
	if !validTemplateID(template) || !validIdentityLocale(locale) {
		return protectedDelivery{}, ErrEmailDelivery
	}
	token, err := securitykit.NewOpaqueToken(service.random)
	if err != nil {
		return protectedDelivery{}, ErrRandomSource
	}
	raw := token.Copy()
	defer clear(raw)
	var digest [32]byte
	domain := emailVerificationDeliveryDomain
	if template == VerifyEmailTemplate {
		digest = securitykit.DigestToken(securitykit.EmailVerificationToken, token)
	} else {
		domain = passwordResetDeliveryDomain
		digest = passwordResetTokenDigest(token)
	}
	if digest == [32]byte{} {
		return protectedDelivery{}, ErrRandomSource
	}
	recipientBytes := recipient.Bytes()
	defer clear(recipientBytes)
	plaintext := pendingDeliveryJSON(recipientBytes, raw, template, locale)
	if len(plaintext) == 0 || len(plaintext) > maximumPendingDeliveryBytes {
		clear(plaintext)
		return protectedDelivery{}, ErrEmailDelivery
	}
	protected, err := service.protector.Encrypt(domain, plaintext)
	clear(plaintext)
	if err != nil || !validProtectedField(protected, maximumPendingDeliveryBytes) {
		clear(protected.Ciphertext)
		return protectedDelivery{}, ErrEmailProvider
	}
	id := service.newUUID()
	if id == uuid.Nil {
		clear(protected.Ciphertext)
		return protectedDelivery{}, ErrRandomSource
	}
	return protectedDelivery{id: id, tokenDigest: append([]byte(nil), digest[:]...), protected: protected}, nil
}

func pendingDeliveryJSON(recipient, rawToken []byte, template TemplateID, locale string) []byte {
	encodedToken := make([]byte, base64.RawURLEncoding.EncodedLen(len(rawToken)))
	base64.RawURLEncoding.Encode(encodedToken, rawToken)
	defer clear(encodedToken)
	result := make([]byte, 0, len(recipient)+len(encodedToken)+len(template)+len(locale)+64)
	result = append(result, `{"recipient":"`...)
	result = append(result, recipient...)
	result = append(result, `","token":"`...)
	result = append(result, encodedToken...)
	result = append(result, `","template":"`...)
	result = append(result, template...)
	result = append(result, `","locale":"`...)
	result = append(result, locale...)
	result = append(result, `"}`...)
	return result
}

func (service *Service) emailDeliveryEvent(deliveryID uuid.UUID, template TemplateID, locale, idempotencyKey string, now time.Time) (*eventsv1.EventEnvelope, error) {
	payload, err := contractevents.MarshalPayload(contractevents.EmailDeliveryRequestedType, &identityv1.EmailDeliveryRequested{
		DeliveryId: deliveryID.String(), TemplateId: string(template), Locale: locale,
	})
	if err != nil {
		return nil, ErrRepository
	}
	eventID := service.newUUID()
	if eventID == uuid.Nil {
		return nil, ErrRepository
	}
	return &eventsv1.EventEnvelope{
		EventId: eventID.String(), EventType: contractevents.EmailDeliveryRequestedType, OccurredAt: timestamppb.New(now),
		Producer: "identity", AggregateType: "email_delivery", AggregateId: deliveryID.String(), AggregateVersion: 1,
		IdempotencyKey: idempotencyKey, Payload: payload,
	}, nil
}

func (service *Service) accountStateEvent(principalID uuid.UUID, state string, version int64, idempotencyKey string, now time.Time) (*eventsv1.EventEnvelope, error) {
	if version < 1 {
		return nil, ErrRepository
	}
	payload, err := contractevents.MarshalPayload(contractevents.AccountStateChangedType, &identityv1.AccountStateChanged{
		PrincipalId: principalID.String(), State: state, Version: uint64(version),
	})
	if err != nil {
		return nil, ErrRepository
	}
	eventID := service.newUUID()
	if eventID == uuid.Nil {
		return nil, ErrRepository
	}
	return &eventsv1.EventEnvelope{
		EventId: eventID.String(), EventType: contractevents.AccountStateChangedType, OccurredAt: timestamppb.New(now),
		Producer: "identity", AggregateType: "account", AggregateId: principalID.String(), AggregateVersion: uint64(version),
		IdempotencyKey: idempotencyKey, Payload: payload,
	}, nil
}

func privateCanonicalRequest(protector sensitive.Protector, operation string, parts ...[]byte) ([]byte, error) {
	if nilIdentityValue(protector) || !validPrivateRequestOperation(operation) || len(parts) == 0 || len(parts) > math.MaxUint32 || len(operation) > math.MaxUint32 {
		return nil, ErrInvalidApplication
	}
	materialLength := 8 + len(operation)
	for _, part := range parts {
		if len(part) > math.MaxUint32 || materialLength > math.MaxInt-4-len(part) {
			return nil, ErrInvalidApplication
		}
		materialLength += 4 + len(part)
	}
	material := make([]byte, 0, materialLength)
	var length [4]byte
	// #nosec G115 -- operation and field lengths are explicitly bounded above.
	binary.BigEndian.PutUint32(length[:], uint32(len(operation)))
	material = append(material, length[:]...)
	material = append(material, operation...)
	// #nosec G115 -- the field count is explicitly bounded above.
	binary.BigEndian.PutUint32(length[:], uint32(len(parts)))
	material = append(material, length[:]...)
	for _, part := range parts {
		// #nosec G115 -- field lengths are explicitly bounded above.
		binary.BigEndian.PutUint32(length[:], uint32(len(part)))
		material = append(material, length[:]...)
		material = append(material, part...)
	}
	clear(length[:])
	defer clear(material)
	digest := protector.LookupDigest(privateRequestDigestDomain, material)
	if digest == [32]byte{} {
		return nil, ErrInvalidApplication
	}
	canonical := append([]byte(nil), digest[:]...)
	clear(digest[:])
	return canonical, nil
}

func validPrivateRequestOperation(operation string) bool {
	switch operation {
	case "register_account", idempotency.CreateEmailVerificationDeliveryOperation, idempotency.CreatePasswordResetDeliveryOperation,
		"verify_email", "reset_password", "create_enrollment_grant":
		return true
	default:
		return false
	}
}

func passwordResetTokenDigest(token secret.Bytes) [32]byte {
	raw := token.Copy()
	defer clear(raw)
	if len(raw) != 32 {
		return [32]byte{}
	}
	material := make([]byte, 0, len(passwordResetDigestDomain)+len(raw))
	material = append(material, passwordResetDigestDomain...)
	material = append(material, raw...)
	digest := sha256.Sum256(material)
	clear(material)
	return digest
}

func registrationIdempotencyOutcome(outcome idempotency.Outcome, record idempotency.Record) error {
	switch outcome {
	case idempotency.Started:
		return nil
	case idempotency.Replay:
		body, owned := record.TakeResponseBody()
		defer clear(body)
		if owned && record.ResponseStatus() == 202 && string(body) == acceptedResponseBody {
			return nil
		}
		return dependencyUnavailable()
	case idempotency.Conflict:
		return apierrors.New(apierrors.IdempotencyConflict, apierrors.ContactSupport)
	case idempotency.InProgress:
		return apierrors.NewRetryAfter(apierrors.StateConflict, apierrors.Retry, time.Second)
	default:
		return dependencyUnavailable()
	}
}

func validService(service *Service) bool {
	return service != nil && !nilIdentityValue(service.repository) && !nilIdentityValue(service.protector) && !nilIdentityValue(service.random) &&
		!nilIdentityValue(service.clock) && !nilIdentityValue(service.rateLimiter) && !nilIdentityValue(service.participant) && service.derive != nil && service.newUUID != nil
}

func validApplicationSecurity(security config.SecurityConfig) bool {
	if security.Profile != config.ProfileLocal && security.Profile != config.ProfileTest && security.Profile != config.ProfileProduction {
		return false
	}
	if security.EmailVerification != config.EmailRequired && security.EmailVerification != config.EmailGrace && security.EmailVerification != config.EmailDisabled {
		return false
	}
	if security.Profile == config.ProfileProduction && security.EmailVerification == config.EmailDisabled {
		return false
	}
	return security.RequestDeadline >= 2*time.Second && security.RequestDeadline <= 10*time.Second &&
		security.DeliveryRateLimit.Limit >= 1 && security.DeliveryRateLimit.Limit <= 20 &&
		security.DeliveryRateLimit.Window >= 10*time.Minute && security.DeliveryRateLimit.Window <= 24*time.Hour
}

func validIdentityLocale(locale string) bool {
	switch locale {
	case "zh-Hans", "en", "ru", "fa", "ja":
		return true
	default:
		return false
	}
}

func validPasswordSecret(value secret.Bytes) bool {
	bytes := value.Copy()
	defer clear(bytes)
	return validPassword(bytes)
}

func validIdempotencyKeyForApplication(value string) bool {
	_, err := idempotency.KeyDigest(value)
	return err == nil
}

func validProtectedField(field sensitive.EncryptedField, maximum int) bool {
	return field.KeyVersion >= 1 && field.KeyVersion <= math.MaxInt32 && len(field.Ciphertext) >= 29 && len(field.Ciphertext) <= maximum
}

func (service *Service) now() (time.Time, bool) {
	now := service.clock.Now().UTC()
	return now, !now.IsZero() && now.Year() >= 2020 && now.Year() <= 2100
}

func malformedRequest() error {
	return apierrors.New(apierrors.MalformedRequest, apierrors.ContactSupport)
}
func authenticationFailed() error {
	return apierrors.New(apierrors.AuthenticationFailed, apierrors.Reauthenticate)
}
func actionNotAllowed() error {
	return apierrors.New(apierrors.ActionNotAllowed, apierrors.Reauthenticate)
}
func dependencyUnavailable() error {
	return apierrors.NewRetryAfter(apierrors.DependencyUnavailable, apierrors.Retry, time.Second)
}

func mapApplicationError(ctx context.Context, err error) error {
	var classified apierrors.Error
	if errors.As(err, &classified) {
		return classified
	}
	if ctx != nil && ctx.Err() != nil {
		return dependencyUnavailable()
	}
	return dependencyUnavailable()
}
