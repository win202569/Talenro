package sensitive

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestObservedProtectorRecordsRealFiniteSuccesses(t *testing.T) {
	digest := [32]byte{1}
	base := &task18ObservedProtector{
		digest:    digest,
		encrypted: EncryptedField{KeyVersion: 1, Ciphertext: []byte{1, 2, 3}},
		plaintext: []byte("opened"),
	}
	observer := &task18SensitiveObserver{}
	observed, err := NewObservedProtector(base, observer)
	if err != nil {
		t.Fatal(err)
	}
	if got := observed.LookupDigest("identity.email", []byte("value")); got != digest {
		t.Fatalf("lookup digest = %x", got)
	}
	encrypted, err := observed.Encrypt("identity.email", []byte("value"))
	if err != nil || encrypted.KeyVersion != 1 || !bytes.Equal(encrypted.Ciphertext, []byte{1, 2, 3}) {
		t.Fatalf("encrypt = %#v, %v", encrypted, err)
	}
	opened, err := observed.Decrypt("identity.email", encrypted)
	if err != nil || string(opened) != "opened" {
		t.Fatalf("decrypt = %q, %v", opened, err)
	}
	want := []CryptoEvent{
		{Operation: CryptoOperationLookup, Result: CryptoResultSuccess, Reason: CryptoReasonNone},
		{Operation: CryptoOperationEncrypt, Result: CryptoResultSuccess, Reason: CryptoReasonNone},
		{Operation: CryptoOperationDecrypt, Result: CryptoResultSuccess, Reason: CryptoReasonNone},
	}
	if !sameTask18SensitiveEvents(observer.events, want) {
		t.Fatalf("events = %#v, want %#v", observer.events, want)
	}
}

func TestObservedProtectorCollapsesFailureAndPanicWithoutRawValues(t *testing.T) {
	const canary = "CANARY-sensitive-provider-secret"
	observer := &task18SensitiveObserver{}
	observed, err := NewObservedProtector(&task18ObservedProtector{
		digest: [32]byte{}, encryptErr: errors.New(canary), panicDecrypt: true,
	}, observer)
	if err != nil {
		t.Fatal(err)
	}
	if got := observed.LookupDigest("identity.email", []byte(canary)); got != [32]byte{} {
		t.Fatalf("failed lookup digest = %x", got)
	}
	if _, err = observed.Encrypt("identity.email", []byte(canary)); !errors.Is(err, ErrProtectionFailed) || strings.Contains(err.Error(), canary) {
		t.Fatalf("encrypt error = %v", err)
	}
	if _, err = observed.Decrypt("identity.email", EncryptedField{KeyVersion: 1, Ciphertext: bytes.Repeat([]byte{1}, 32)}); !errors.Is(err, ErrProtectionFailed) || strings.Contains(err.Error(), canary) {
		t.Fatalf("decrypt panic error = %v", err)
	}
	for _, event := range observer.events {
		if event.Result != CryptoResultFailure || event.Reason == CryptoReasonNone {
			t.Fatalf("failure event = %#v", event)
		}
	}
	if strings.Contains(strings.Join(observer.rendered, ","), canary) {
		t.Fatal("raw value reached the finite observer")
	}
}

func TestObservedProtectorClearsPartialProviderResultsAndClassifiesClosedKey(t *testing.T) {
	const canary = "CANARY-sensitive-provider-partial"
	observer := &task18SensitiveObserver{}
	observed, err := NewObservedProtector(&task18ObservedProtector{
		encrypted:  EncryptedField{KeyVersion: 7, Ciphertext: []byte(canary)},
		plaintext:  []byte(canary),
		encryptErr: ErrClosed,
		decryptErr: errors.New(canary),
	}, observer)
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := observed.Encrypt("identity.email", []byte(canary))
	if !errors.Is(err, ErrProtectionFailed) || encrypted.KeyVersion != 0 || len(encrypted.Ciphertext) != 0 {
		t.Fatalf("partial encrypt result = %#v, %v", encrypted, err)
	}
	plaintext, err := observed.Decrypt("identity.email", EncryptedField{KeyVersion: 1, Ciphertext: bytes.Repeat([]byte{1}, 32)})
	if !errors.Is(err, ErrProtectionFailed) || len(plaintext) != 0 {
		t.Fatalf("partial decrypt result = %q, %v", plaintext, err)
	}
	want := []CryptoEvent{
		{Operation: CryptoOperationEncrypt, Result: CryptoResultFailure, Reason: CryptoReasonKeyUnavailable},
		{Operation: CryptoOperationDecrypt, Result: CryptoResultFailure, Reason: CryptoReasonInvalid},
	}
	if !sameTask18SensitiveEvents(observer.events, want) {
		t.Fatalf("events = %#v, want %#v", observer.events, want)
	}
}

func TestObservedProtectorContainsObserverPanicAndClosesDelegate(t *testing.T) {
	base := &task18ObservedProtector{digest: [32]byte{1}}
	observed, err := NewObservedProtector(base, &task18SensitiveObserver{panicObserve: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := observed.LookupDigest("identity.email", []byte("value")); got != base.digest {
		t.Fatalf("observer panic changed lookup = %x", got)
	}
	if err := observed.Close(); err != nil || base.closeCalls != 1 {
		t.Fatalf("Close = %v, calls %d", err, base.closeCalls)
	}
}

func sameTask18SensitiveEvents(left, right []CryptoEvent) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

type task18ObservedProtector struct {
	digest       [32]byte
	encrypted    EncryptedField
	plaintext    []byte
	encryptErr   error
	decryptErr   error
	panicDecrypt bool
	closeCalls   int
}

func (fake *task18ObservedProtector) LookupDigest(string, []byte) [32]byte { return fake.digest }

func (fake *task18ObservedProtector) Encrypt(string, []byte) (EncryptedField, error) {
	return EncryptedField{KeyVersion: fake.encrypted.KeyVersion, Ciphertext: bytes.Clone(fake.encrypted.Ciphertext)}, fake.encryptErr
}

func (fake *task18ObservedProtector) Decrypt(string, EncryptedField) ([]byte, error) {
	if fake.panicDecrypt {
		panic("CANARY-sensitive-provider-panic")
	}
	return bytes.Clone(fake.plaintext), fake.decryptErr
}

func (fake *task18ObservedProtector) Close() error {
	fake.closeCalls++
	return nil
}

type task18SensitiveObserver struct {
	events       []CryptoEvent
	rendered     []string
	panicObserve bool
}

func (observer *task18SensitiveObserver) ObserveSensitiveCrypto(event CryptoEvent) {
	if observer.panicObserve {
		panic("CANARY-observer-panic")
	}
	observer.events = append(observer.events, event)
	observer.rendered = append(observer.rendered, event.String())
}
