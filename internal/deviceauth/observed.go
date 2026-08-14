package deviceauth

import (
	"context"
	"errors"

	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/store"
)

// CryptoOperation is the finite device cryptographic operation registry.
type CryptoOperation string

// CryptoResult is the finite device cryptographic result registry.
type CryptoResult string

// CryptoReason is the finite device cryptographic reason registry.
type CryptoReason string

//nolint:revive // These closed observation registries are documented by their exported types and finite values.
const (
	// CryptoOperationTokenVerify and the following constants are finite observation values.
	CryptoOperationTokenVerify CryptoOperation = "token_verify"
	CryptoOperationProofVerify CryptoOperation = "proof_verify"

	CryptoResultSuccess CryptoResult = "success"
	CryptoResultFailure CryptoResult = "failure"

	CryptoReasonNone           CryptoReason = "none"
	CryptoReasonInvalid        CryptoReason = "invalid"
	CryptoReasonKeyUnavailable CryptoReason = "key_unavailable"
)

// CryptoEvent is the complete finite device cryptographic observation.
type CryptoEvent struct {
	Operation CryptoOperation
	Result    CryptoResult
	Reason    CryptoReason
}

// CryptoObserver receives only finite device outcomes.
type CryptoObserver interface {
	ObserveDeviceCrypto(CryptoEvent)
}

type deviceCryptoObserverContextKey struct{}

type observedDeviceDelegate interface {
	Application
	DeviceAuthenticator
	TrustTransactionParticipant
}

// ObservedApplication decorates the frozen device surfaces with finite crypto observations.
type ObservedApplication struct {
	delegate observedDeviceDelegate
	observer CryptoObserver
}

var (
	_ Application                 = (*ObservedApplication)(nil)
	_ DeviceAuthenticator         = (*ObservedApplication)(nil)
	_ TrustTransactionParticipant = (*ObservedApplication)(nil)
)

// NewObservedApplication binds a real device application to a finite observer.
func NewObservedApplication(delegate observedDeviceDelegate, observer CryptoObserver) (*ObservedApplication, error) {
	if nilDeviceauthValue(delegate) || nilDeviceauthValue(observer) {
		return nil, ErrInvalidApplication
	}
	return &ObservedApplication{delegate: delegate, observer: observer}, nil
}

// CreateChallenge delegates challenge construction while containing raw failures.
func (observed *ObservedApplication) CreateChallenge(ctx context.Context, command CreateChallengeCommand) (result Challenge, resultErr error) {
	if observed == nil || nilDeviceauthValue(observed.delegate) {
		return Challenge{}, deviceDependencyUnavailable()
	}
	defer func() {
		if recover() != nil {
			resultErr = deviceDependencyUnavailable()
		}
		resultErr = sanitizeObservedDeviceError(resultErr)
		if resultErr != nil {
			result = Challenge{}
		}
	}()
	result, resultErr = observed.delegate.CreateChallenge(ctx, command)
	return result, sanitizeObservedDeviceError(resultErr)
}

// RegisterDevice delegates registration and observes its proof verification outcome.
func (observed *ObservedApplication) RegisterDevice(ctx context.Context, command RegisterDeviceCommand) (result DeviceTokens, resultErr error) {
	if observed == nil || nilDeviceauthValue(observed.delegate) {
		return DeviceTokens{}, deviceDependencyUnavailable()
	}
	defer func() {
		if recover() != nil {
			result = DeviceTokens{}
			resultErr = deviceDependencyUnavailable()
		}
		resultErr = sanitizeObservedDeviceError(resultErr)
		if resultErr != nil {
			clearObservedDeviceTokens(&result)
		}
	}()
	return observed.delegate.RegisterDevice(withDeviceCryptoObserver(ctx, observed.observer), command)
}

// RotateDeviceToken delegates rotation and observes its proof verification outcome.
func (observed *ObservedApplication) RotateDeviceToken(ctx context.Context, command RotateDeviceTokenCommand) (result DeviceTokens, resultErr error) {
	if observed == nil || nilDeviceauthValue(observed.delegate) {
		return DeviceTokens{}, deviceDependencyUnavailable()
	}
	defer func() {
		if recover() != nil {
			result = DeviceTokens{}
			resultErr = deviceDependencyUnavailable()
		}
		resultErr = sanitizeObservedDeviceError(resultErr)
		if resultErr != nil {
			clearObservedDeviceTokens(&result)
		}
	}()
	return observed.delegate.RotateDeviceToken(withDeviceCryptoObserver(ctx, observed.observer), command)
}

// RevokeDevice delegates revocation while containing raw failures.
func (observed *ObservedApplication) RevokeDevice(ctx context.Context, command RevokeDeviceCommand) (resultErr error) {
	if observed == nil || nilDeviceauthValue(observed.delegate) {
		return deviceDependencyUnavailable()
	}
	defer observedDevicePanic(&resultErr)
	resultErr = observed.delegate.RevokeDevice(ctx, command)
	return sanitizeObservedDeviceError(resultErr)
}

// AuthorizeBundle delegates bundle authorization and observes token verification.
func (observed *ObservedApplication) AuthorizeBundle(ctx context.Context, query AuthorizeBundleQuery) (result BundleAuthority, resultErr error) {
	if observed == nil || nilDeviceauthValue(observed.delegate) {
		return BundleAuthority{}, deviceDependencyUnavailable()
	}
	defer func() {
		if recover() != nil {
			clear(result.Policy)
			result = BundleAuthority{}
			resultErr = deviceDependencyUnavailable()
		}
		resultErr = sanitizeObservedDeviceError(resultErr)
		if resultErr != nil {
			clearObservedBundleAuthority(&result)
		}
		observed.observe(CryptoOperationTokenVerify, resultErr)
	}()
	return observed.delegate.AuthorizeBundle(ctx, query)
}

// Authenticate delegates device authentication and observes token verification.
func (observed *ObservedApplication) Authenticate(ctx context.Context, token secret.Bytes) (result BundleAuthority, resultErr error) {
	if observed == nil || nilDeviceauthValue(observed.delegate) {
		return BundleAuthority{}, deviceDependencyUnavailable()
	}
	defer func() {
		if recover() != nil {
			clear(result.Policy)
			result = BundleAuthority{}
			resultErr = deviceDependencyUnavailable()
		}
		resultErr = sanitizeObservedDeviceError(resultErr)
		if resultErr != nil {
			clearObservedBundleAuthority(&result)
		}
		observed.observe(CryptoOperationTokenVerify, resultErr)
	}()
	return observed.delegate.Authenticate(ctx, token)
}

// AuthorizeBundleInTransaction delegates trust-owned transactional authorization and observes token verification.
func (observed *ObservedApplication) AuthorizeBundleInTransaction(
	ctx context.Context,
	database store.DBTX,
	query AuthorizeBundleQuery,
) (result BundleAuthority, resultErr error) {
	if observed == nil || nilDeviceauthValue(observed.delegate) {
		return BundleAuthority{}, deviceDependencyUnavailable()
	}
	defer func() {
		if recover() != nil {
			clear(result.Policy)
			result = BundleAuthority{}
			resultErr = deviceDependencyUnavailable()
		}
		resultErr = sanitizeObservedDeviceError(resultErr)
		if resultErr != nil {
			clearObservedBundleAuthority(&result)
		}
		observed.observe(CryptoOperationTokenVerify, resultErr)
	}()
	return observed.delegate.AuthorizeBundleInTransaction(ctx, database, query)
}

func (observed *ObservedApplication) observe(operation CryptoOperation, operationErr error) {
	if observed == nil || nilDeviceauthValue(observed.observer) {
		return
	}
	event := CryptoEvent{Operation: operation, Result: CryptoResultSuccess, Reason: CryptoReasonNone}
	if operationErr != nil {
		event.Result = CryptoResultFailure
		event.Reason = observedDeviceReason(operationErr)
	}
	defer func() { _ = recover() }()
	observed.observer.ObserveDeviceCrypto(event)
}

func withDeviceCryptoObserver(ctx context.Context, observer CryptoObserver) context.Context {
	if ctx == nil || nilDeviceauthValue(observer) {
		return ctx
	}
	return context.WithValue(ctx, deviceCryptoObserverContextKey{}, observer)
}

func observeDeviceProofVerification(ctx context.Context, verified bool) {
	if ctx == nil {
		return
	}
	observer, ok := ctx.Value(deviceCryptoObserverContextKey{}).(CryptoObserver)
	if !ok || nilDeviceauthValue(observer) {
		return
	}
	event := CryptoEvent{
		Operation: CryptoOperationProofVerify,
		Result:    CryptoResultSuccess,
		Reason:    CryptoReasonNone,
	}
	if !verified {
		event.Result = CryptoResultFailure
		event.Reason = CryptoReasonInvalid
	}
	defer func() { _ = recover() }()
	observer.ObserveDeviceCrypto(event)
}

func observedDevicePanic(resultErr *error) {
	if recover() != nil {
		*resultErr = deviceDependencyUnavailable()
	}
	*resultErr = sanitizeObservedDeviceError(*resultErr)
}

func clearObservedDeviceTokens(tokens *DeviceTokens) {
	if tokens == nil {
		return
	}
	tokens.AccessToken.Clear()
	tokens.RefreshToken.Clear()
	*tokens = DeviceTokens{}
}

func clearObservedBundleAuthority(authority *BundleAuthority) {
	if authority == nil {
		return
	}
	clear(authority.Policy)
	*authority = BundleAuthority{}
}

func sanitizeObservedDeviceError(err error) error {
	if err == nil {
		return nil
	}
	var classified apierrors.Error
	if errors.As(err, &classified) {
		return classified
	}
	return deviceDependencyUnavailable()
}

func observedDeviceReason(err error) CryptoReason {
	var classified apierrors.Error
	if !errors.As(err, &classified) {
		return CryptoReasonKeyUnavailable
	}
	public := classified.Public("trace-unavailable")
	if public.Code == "authentication_failed" || public.Code == "malformed_request" || public.Code == "action_not_allowed" {
		return CryptoReasonInvalid
	}
	return CryptoReasonKeyUnavailable
}
