package trust

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestObservedConfigSignerRecordsRealFiniteSuccess(t *testing.T) {
	delegate := &task18ObservedSigner{keyID: strings.Repeat("k", 22), signature: bytes.Repeat([]byte{1}, 64)}
	observer := &task18TrustObserver{}
	observed, err := NewObservedConfigSigner(delegate, observer)
	if err != nil {
		t.Fatal(err)
	}
	if observed.KeyID() != delegate.keyID {
		t.Fatalf("KeyID = %q", observed.KeyID())
	}
	signature, err := observed.Sign(context.Background(), []byte("message"))
	if err != nil || !bytes.Equal(signature, delegate.signature) {
		t.Fatalf("Sign = %x, %v", signature, err)
	}
	want := CryptoEvent{Operation: CryptoOperationBundleSign, Result: CryptoResultSuccess, Reason: CryptoReasonNone}
	if len(observer.events) != 1 || observer.events[0] != want {
		t.Fatalf("events = %#v, want %#v", observer.events, want)
	}
}

func TestObservedConfigSignerCollapsesRawFailureAndPanic(t *testing.T) {
	const canary = "CANARY-trust-signer-secret"
	for _, delegate := range []*task18ObservedSigner{
		{keyID: strings.Repeat("k", 22), err: errors.New(canary)},
		{keyID: strings.Repeat("k", 22), panicSign: true},
	} {
		observer := &task18TrustObserver{}
		observed, err := NewObservedConfigSigner(delegate, observer)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = observed.Sign(context.Background(), []byte(canary)); !errors.Is(err, ErrSignerFailure) || strings.Contains(err.Error(), canary) {
			t.Fatalf("Sign error = %v", err)
		}
		want := CryptoEvent{Operation: CryptoOperationBundleSign, Result: CryptoResultFailure, Reason: CryptoReasonKeyUnavailable}
		if len(observer.events) != 1 || observer.events[0] != want {
			t.Fatalf("events = %#v, want %#v", observer.events, want)
		}
	}
}

func TestObservedConfigSignerContainsObserverAndKeyIDPanic(t *testing.T) {
	delegate := &task18ObservedSigner{keyID: strings.Repeat("k", 22), signature: bytes.Repeat([]byte{1}, 64)}
	observed, err := NewObservedConfigSigner(delegate, &task18TrustObserver{panicObserve: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := observed.Sign(context.Background(), []byte("message")); err != nil {
		t.Fatalf("observer panic changed Sign: %v", err)
	}
	panicking, err := NewObservedConfigSigner(&task18ObservedSigner{panicKeyID: true}, &task18TrustObserver{})
	if err != nil {
		t.Fatal(err)
	}
	if got := panicking.KeyID(); got != "" {
		t.Fatalf("panicking KeyID = %q", got)
	}
}

type task18ObservedSigner struct {
	keyID      string
	signature  []byte
	err        error
	panicSign  bool
	panicKeyID bool
}

func (fake *task18ObservedSigner) KeyID() string {
	if fake.panicKeyID {
		panic("CANARY-key-id-panic")
	}
	return fake.keyID
}

func (fake *task18ObservedSigner) Sign(context.Context, []byte) ([]byte, error) {
	if fake.panicSign {
		panic("CANARY-sign-panic")
	}
	return bytes.Clone(fake.signature), fake.err
}

type task18TrustObserver struct {
	events       []CryptoEvent
	panicObserve bool
}

func (observer *task18TrustObserver) ObserveTrustCrypto(event CryptoEvent) {
	if observer.panicObserve {
		panic("CANARY-trust-observer-panic")
	}
	observer.events = append(observer.events, event)
}
