package trust

import (
	"context"
	"reflect"
)

// CryptoOperation is the finite trust cryptographic operation registry.
type CryptoOperation string

// CryptoResult is the finite trust cryptographic result registry.
type CryptoResult string

// CryptoReason is the finite trust cryptographic reason registry.
type CryptoReason string

//nolint:revive // These closed observation registries are documented by their exported types and finite values.
const (
	// CryptoOperationBundleSign and the following constants are finite observation values.
	CryptoOperationBundleSign CryptoOperation = "bundle_sign"

	CryptoResultSuccess CryptoResult = "success"
	CryptoResultFailure CryptoResult = "failure"

	CryptoReasonNone           CryptoReason = "none"
	CryptoReasonKeyUnavailable CryptoReason = "key_unavailable"
)

// CryptoEvent is the complete finite trust cryptographic observation.
type CryptoEvent struct {
	Operation CryptoOperation
	Result    CryptoResult
	Reason    CryptoReason
}

// CryptoObserver receives only finite trust outcomes.
type CryptoObserver interface {
	ObserveTrustCrypto(CryptoEvent)
}

// ObservedConfigSigner contains raw provider failures and records real signing outcomes.
type ObservedConfigSigner struct {
	signer   ConfigSigner
	observer CryptoObserver
}

var _ ConfigSigner = (*ObservedConfigSigner)(nil)

// NewObservedConfigSigner binds a real signer to a finite observer.
func NewObservedConfigSigner(signer ConfigSigner, observer CryptoObserver) (*ObservedConfigSigner, error) {
	if observedTrustNil(signer) || observedTrustNil(observer) {
		return nil, ErrInvalidArgument
	}
	return &ObservedConfigSigner{signer: signer, observer: observer}, nil
}

// KeyID returns the delegate's identifier while containing provider panics.
func (observed *ObservedConfigSigner) KeyID() (keyID string) {
	if observed == nil || observedTrustNil(observed.signer) {
		return ""
	}
	defer func() {
		if recover() != nil {
			keyID = ""
		}
	}()
	return observed.signer.KeyID()
}

// Sign delegates one real signing call and collapses provider detail.
func (observed *ObservedConfigSigner) Sign(ctx context.Context, message []byte) (signature []byte, resultErr error) {
	if observed == nil || observedTrustNil(observed.signer) {
		return nil, ErrSignerFailure
	}
	defer func() {
		if recover() != nil {
			clear(signature)
			signature = nil
			resultErr = ErrSignerFailure
		}
		if resultErr != nil {
			clear(signature)
			signature = nil
			resultErr = ErrSignerFailure
		}
		observed.observe(resultErr)
	}()
	signature, resultErr = observed.signer.Sign(ctx, message)
	return signature, resultErr
}

func (observed *ObservedConfigSigner) observe(operationErr error) {
	if observed == nil || observedTrustNil(observed.observer) {
		return
	}
	event := CryptoEvent{Operation: CryptoOperationBundleSign, Result: CryptoResultSuccess, Reason: CryptoReasonNone}
	if operationErr != nil {
		event.Result = CryptoResultFailure
		event.Reason = CryptoReasonKeyUnavailable
	}
	defer func() { _ = recover() }()
	observed.observer.ObserveTrustCrypto(event)
}

func observedTrustNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	kind := reflected.Kind()
	return (kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface || kind == reflect.Map ||
		kind == reflect.Pointer || kind == reflect.Slice) && reflected.IsNil()
}
