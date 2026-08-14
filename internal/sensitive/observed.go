package sensitive

import (
	"errors"
	"reflect"
)

// CryptoOperation is the finite sensitive-field operation registry.
type CryptoOperation string

// CryptoResult is the finite sensitive-field result registry.
type CryptoResult string

// CryptoReason is the finite sensitive-field reason registry.
type CryptoReason string

//nolint:revive // These closed observation registries are documented by their exported types and finite values.
const (
	// CryptoOperationLookup and the following constants are finite observation values.
	CryptoOperationLookup  CryptoOperation = "lookup"
	CryptoOperationEncrypt CryptoOperation = "encrypt"
	CryptoOperationDecrypt CryptoOperation = "decrypt"

	CryptoResultSuccess CryptoResult = "success"
	CryptoResultFailure CryptoResult = "failure"

	CryptoReasonNone           CryptoReason = "none"
	CryptoReasonInvalid        CryptoReason = "invalid"
	CryptoReasonKeyUnavailable CryptoReason = "key_unavailable"
)

// CryptoEvent is the complete finite sensitive-field observation.
type CryptoEvent struct {
	Operation CryptoOperation
	Result    CryptoResult
	Reason    CryptoReason
}

// String renders only finite registry values.
func (event CryptoEvent) String() string {
	return string(event.Operation) + ":" + string(event.Result) + ":" + string(event.Reason)
}

// CryptoObserver receives only finite sensitive-field outcomes.
type CryptoObserver interface {
	ObserveSensitiveCrypto(CryptoEvent)
}

// ObservedProtector contains provider panics/errors and emits finite outcomes.
type ObservedProtector struct {
	protector Protector
	observer  CryptoObserver
}

var _ Protector = (*ObservedProtector)(nil)

// NewObservedProtector binds a real protector to a finite observer.
func NewObservedProtector(protector Protector, observer CryptoObserver) (*ObservedProtector, error) {
	if observedNil(protector) || observedNil(observer) {
		return nil, ErrInvalidArgument
	}
	return &ObservedProtector{protector: protector, observer: observer}, nil
}

// LookupDigest delegates the real lookup and records whether it produced a digest.
func (observed *ObservedProtector) LookupDigest(domain string, canonical []byte) (digest [32]byte) {
	if observed == nil || observedNil(observed.protector) {
		return digest
	}
	failed := false
	func() {
		defer func() {
			if recover() != nil {
				digest = [32]byte{}
				failed = true
			}
		}()
		digest = observed.protector.LookupDigest(domain, canonical)
	}()
	failed = failed || digest == [32]byte{}
	observed.observe(CryptoOperationLookup, failed, CryptoReasonInvalid)
	return digest
}

// Encrypt delegates one real protected-field encryption.
func (observed *ObservedProtector) Encrypt(domain string, plaintext []byte) (encrypted EncryptedField, resultErr error) {
	if observed == nil || observedNil(observed.protector) {
		return EncryptedField{}, ErrProtectionFailed
	}
	defer func() {
		if recover() != nil {
			clear(encrypted.Ciphertext)
			encrypted = EncryptedField{}
			resultErr = ErrProtectionFailed
		}
		failed := resultErr != nil
		reason := observedReason(resultErr)
		if failed {
			clear(encrypted.Ciphertext)
			encrypted = EncryptedField{}
			resultErr = ErrProtectionFailed
		}
		observed.observe(CryptoOperationEncrypt, failed, reason)
	}()
	encrypted, resultErr = observed.protector.Encrypt(domain, plaintext)
	return encrypted, resultErr
}

// Decrypt delegates one real protected-field authentication/decryption.
func (observed *ObservedProtector) Decrypt(domain string, encrypted EncryptedField) (plaintext []byte, resultErr error) {
	if observed == nil || observedNil(observed.protector) {
		return nil, ErrProtectionFailed
	}
	defer func() {
		if recover() != nil {
			clear(plaintext)
			plaintext = nil
			resultErr = ErrProtectionFailed
		}
		failed := resultErr != nil
		reason := observedReason(resultErr)
		if failed {
			clear(plaintext)
			plaintext = nil
			resultErr = ErrProtectionFailed
		}
		observed.observe(CryptoOperationDecrypt, failed, reason)
	}()
	plaintext, resultErr = observed.protector.Decrypt(domain, encrypted)
	return plaintext, resultErr
}

// Close closes an underlying local/external resource when it owns one.
func (observed *ObservedProtector) Close() (resultErr error) {
	if observed == nil || observed.protector == nil {
		return nil
	}
	closer, ok := observed.protector.(interface{ Close() error })
	if !ok || closer == nil {
		return nil
	}
	defer func() {
		if recover() != nil {
			resultErr = ErrClosed
		}
	}()
	return closer.Close()
}

func (observed *ObservedProtector) observe(operation CryptoOperation, failed bool, reason CryptoReason) {
	if observed == nil || observedNil(observed.observer) {
		return
	}
	event := CryptoEvent{Operation: operation, Result: CryptoResultSuccess, Reason: CryptoReasonNone}
	if failed {
		event.Result = CryptoResultFailure
		event.Reason = reason
	}
	defer func() { _ = recover() }()
	observed.observer.ObserveSensitiveCrypto(event)
}

func observedReason(err error) CryptoReason {
	if errors.Is(err, ErrClosed) {
		return CryptoReasonKeyUnavailable
	}
	return CryptoReasonInvalid
}

func observedNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	kind := reflected.Kind()
	return (kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface || kind == reflect.Map ||
		kind == reflect.Pointer || kind == reflect.Slice) && reflected.IsNil()
}
