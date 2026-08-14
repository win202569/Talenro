package deviceauth

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/store"
)

func TestObservedApplicationRecordsTokenButDoesNotInferProofFromDelegateResult(t *testing.T) {
	delegate := &task18ObservedDeviceApplication{}
	observer := &task18DeviceObserver{}
	observed, err := NewObservedApplication(delegate, observer)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := observed.Authenticate(context.Background(), secret.NewBytes([]byte("token"))); err != nil {
		t.Fatal(err)
	}
	if _, err := observed.RegisterDevice(context.Background(), RegisterDeviceCommand{}); err != nil {
		t.Fatal(err)
	}
	if _, err := observed.RotateDeviceToken(context.Background(), RotateDeviceTokenCommand{}); err != nil {
		t.Fatal(err)
	}
	want := []CryptoEvent{
		{Operation: CryptoOperationTokenVerify, Result: CryptoResultSuccess, Reason: CryptoReasonNone},
	}
	if !sameTask18DeviceEvents(observer.events, want) {
		t.Fatalf("events = %#v, want %#v", observer.events, want)
	}
}

func TestObservedApplicationInjectsFiniteProofObserverAndContainsObserverPanic(t *testing.T) {
	verified, rejected := true, false
	observer := &task18DeviceObserver{}
	observed, err := NewObservedApplication(&task18ObservedDeviceApplication{
		registerProofResult: &verified,
		rotateProofResult:   &rejected,
	}, observer)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := observed.RegisterDevice(context.Background(), RegisterDeviceCommand{}); err != nil {
		t.Fatal(err)
	}
	if _, err := observed.RotateDeviceToken(context.Background(), RotateDeviceTokenCommand{}); err != nil {
		t.Fatal(err)
	}
	want := []CryptoEvent{
		{Operation: CryptoOperationProofVerify, Result: CryptoResultSuccess, Reason: CryptoReasonNone},
		{Operation: CryptoOperationProofVerify, Result: CryptoResultFailure, Reason: CryptoReasonInvalid},
	}
	if !sameTask18DeviceEvents(observer.events, want) {
		t.Fatalf("context-injected proof events = %#v, want %#v", observer.events, want)
	}

	panicObserved, err := NewObservedApplication(
		&task18ObservedDeviceApplication{registerProofResult: &verified},
		&task18DeviceObserver{panicObserve: true},
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := panicObserved.RegisterDevice(context.Background(), RegisterDeviceCommand{}); err != nil {
		t.Fatalf("proof observer panic escaped application boundary: %v", err)
	}
}

func TestObservedApplicationCollapsesRawFailureAndPanic(t *testing.T) {
	const canary = "CANARY-device-provider-secret"
	observer := &task18DeviceObserver{}
	observed, err := NewObservedApplication(&task18ObservedDeviceApplication{
		authenticateErr: errors.New(canary), panicRegister: true,
	}, observer)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = observed.Authenticate(context.Background(), secret.NewBytes([]byte(canary))); !fixedTask18DeviceError(err, canary) {
		t.Fatalf("Authenticate error = %v", err)
	}
	if _, err = observed.RegisterDevice(context.Background(), RegisterDeviceCommand{}); !fixedTask18DeviceError(err, canary) {
		t.Fatalf("RegisterDevice panic error = %v", err)
	}
	for _, event := range observer.events {
		if event.Result != CryptoResultFailure || event.Reason == CryptoReasonNone {
			t.Fatalf("failure event = %#v", event)
		}
	}
}

func TestObservedApplicationPreservesFiniteDomainErrorAndContainsObserverPanic(t *testing.T) {
	domainErr := apierrors.New(apierrors.AuthenticationFailed, apierrors.Reauthenticate)
	observed, err := NewObservedApplication(
		&task18ObservedDeviceApplication{authenticateErr: domainErr},
		&task18DeviceObserver{panicObserve: true},
	)
	if err != nil {
		t.Fatal(err)
	}
	_, got := observed.Authenticate(context.Background(), secret.NewBytes([]byte("token")))
	var classified apierrors.Error
	if !errors.As(got, &classified) {
		t.Fatalf("finite domain error was not preserved: %v", got)
	}
}

func TestObservedApplicationClearsPartialSensitiveResultsOnFailure(t *testing.T) {
	const canary = "CANARY-device-provider-partial"
	delegate := &task18ObservedDeviceApplication{
		registerResult: DeviceTokens{
			DeviceID: uuid.New(), AccessToken: secret.NewBytes([]byte(canary)), RefreshToken: secret.NewBytes([]byte(canary)),
		},
		registerErr: errors.New(canary),
		authenticateResult: BundleAuthority{
			DeviceID: uuid.New(), Policy: []byte(canary),
		},
		authenticateErr: errors.New(canary),
	}
	observed, err := NewObservedApplication(delegate, &task18DeviceObserver{})
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := observed.RegisterDevice(context.Background(), RegisterDeviceCommand{})
	if !fixedTask18DeviceError(err, canary) || tokens.DeviceID != uuid.Nil ||
		len(tokens.AccessToken.Copy()) != 0 || len(tokens.RefreshToken.Copy()) != 0 {
		t.Fatalf("partial registration result was retained: %#v, %v", tokens, err)
	}
	authority, err := observed.Authenticate(context.Background(), secret.NewBytes([]byte(canary)))
	if !fixedTask18DeviceError(err, canary) || authority.DeviceID != uuid.Nil || len(authority.Policy) != 0 {
		t.Fatalf("partial authority result was retained: %#v, %v", authority, err)
	}
}

func fixedTask18DeviceError(err error, canary string) bool {
	var classified apierrors.Error
	return errors.As(err, &classified) && !strings.Contains(err.Error(), canary)
}

func sameTask18DeviceEvents(left, right []CryptoEvent) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		// #nosec G602 -- the equal-length guard above bounds this index for both slices.
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

type task18ObservedDeviceApplication struct {
	registerResult      DeviceTokens
	registerErr         error
	registerProofResult *bool
	rotateProofResult   *bool
	authenticateResult  BundleAuthority
	authenticateErr     error
	panicRegister       bool
}

func (*task18ObservedDeviceApplication) CreateChallenge(context.Context, CreateChallengeCommand) (Challenge, error) {
	return Challenge{}, nil
}

func (fake *task18ObservedDeviceApplication) RegisterDevice(ctx context.Context, _ RegisterDeviceCommand) (DeviceTokens, error) {
	if fake.panicRegister {
		panic("CANARY-device-provider-panic")
	}
	if fake.registerProofResult != nil {
		observeDeviceProofVerification(ctx, *fake.registerProofResult)
	}
	return fake.registerResult, fake.registerErr
}

func (fake *task18ObservedDeviceApplication) RotateDeviceToken(ctx context.Context, _ RotateDeviceTokenCommand) (DeviceTokens, error) {
	if fake.rotateProofResult != nil {
		observeDeviceProofVerification(ctx, *fake.rotateProofResult)
	}
	return DeviceTokens{}, nil
}

func (*task18ObservedDeviceApplication) RevokeDevice(context.Context, RevokeDeviceCommand) error {
	return nil
}

func (*task18ObservedDeviceApplication) AuthorizeBundle(context.Context, AuthorizeBundleQuery) (BundleAuthority, error) {
	return BundleAuthority{}, nil
}

func (fake *task18ObservedDeviceApplication) Authenticate(context.Context, secret.Bytes) (BundleAuthority, error) {
	return fake.authenticateResult, fake.authenticateErr
}

func (*task18ObservedDeviceApplication) AuthorizeBundleInTransaction(context.Context, store.DBTX, AuthorizeBundleQuery) (BundleAuthority, error) {
	return BundleAuthority{}, nil
}

type task18DeviceObserver struct {
	events       []CryptoEvent
	panicObserve bool
}

func (observer *task18DeviceObserver) ObserveDeviceCrypto(event CryptoEvent) {
	if observer.panicObserve {
		panic("CANARY-device-observer-panic")
	}
	observer.events = append(observer.events, event)
}
