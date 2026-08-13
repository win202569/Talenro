package controlapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/deviceauth"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/readiness"
	"talenro.local/platform/internal/secret"
)

type accountAuthenticatorStub struct {
	authority identity.AccountAuthority
	err       error
	calls     int
}

func (stub *accountAuthenticatorStub) Authenticate(_ context.Context, token secret.Bytes) (identity.AccountAuthority, error) {
	stub.calls++
	copied := token.Copy()
	defer clear(copied)
	if len(copied) != 32 {
		return identity.AccountAuthority{}, errors.New("bad token length")
	}
	return stub.authority, stub.err
}

type deviceAuthenticatorStub struct {
	authority deviceauth.BundleAuthority
	err       error
	calls     int
}

func (stub *deviceAuthenticatorStub) Authenticate(_ context.Context, token secret.Bytes) (deviceauth.BundleAuthority, error) {
	stub.calls++
	copied := token.Copy()
	defer clear(copied)
	if len(copied) != 32 {
		return deviceauth.BundleAuthority{}, errors.New("bad token length")
	}
	return stub.authority, stub.err
}

func TestAuthAcceptsExactBearerAndPreservesTypedAuthority(t *testing.T) {
	principal := uuid.MustParse("28ceee8a-5f4f-4d3a-9e9f-d3ec23b815ac")
	session := uuid.MustParse("d12dca8a-ced3-471d-a6ad-55b228221f10")
	account := &accountAuthenticatorStub{authority: identity.AccountAuthority{PrincipalID: identity.PrincipalID(principal.String()), SessionID: identity.SessionID(session.String())}}
	device := &deviceAuthenticatorStub{authority: deviceauth.BundleAuthority{PrincipalID: principal, DeviceID: uuid.MustParse("1645ba9c-ad1d-4a06-b996-d9feb78ee88a")}}
	handler := &Handler{apps: Applications{AccountAuth: account, DeviceAuth: device}}
	token := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x31}, 32))

	accountRequest := httptest.NewRequestWithContext(t.Context(), "POST", "/", nil)
	accountRequest.Header.Set("Authorization", "Bearer "+token)
	accountRequest, err := handler.authenticateAccount(accountRequest)
	if err != nil || accountAuthority(accountRequest.Context()) != account.authority || account.calls != 1 || device.calls != 0 {
		t.Fatalf("account auth = request %v error %v calls %d/%d", accountRequest != nil, err, account.calls, device.calls)
	}

	deviceRequest := httptest.NewRequestWithContext(t.Context(), "POST", "/", nil)
	deviceRequest.Header.Set("Authorization", "Bearer "+token)
	deviceRequest, err = handler.authenticateDevice(deviceRequest)
	if err != nil || deviceAuthority(deviceRequest.Context()).DeviceID != device.authority.DeviceID || account.calls != 1 || device.calls != 1 {
		t.Fatalf("device auth = request %v error %v calls %d/%d", deviceRequest != nil, err, account.calls, device.calls)
	}
}

func TestAuthRejectsNonExactBearerWithoutCallingAuthenticator(t *testing.T) {
	stub := &accountAuthenticatorStub{}
	handler := &Handler{apps: Applications{AccountAuth: stub}}
	token := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x32}, 32))
	for _, header := range []string{"", "bearer " + token, "Bearer  " + token, "Bearer\t" + token, "Bearer " + token + " extra"} {
		request := httptest.NewRequestWithContext(t.Context(), "POST", "/?access_token="+token, nil)
		request.Header.Set("Authorization", header)
		request.AddCookie(&http.Cookie{Name: "access_token", Value: token, Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode})
		if _, err := handler.authenticateAccount(request); err == nil {
			t.Fatalf("header %q accepted", header)
		}
	}
	if stub.calls != 0 {
		t.Fatalf("authenticator calls = %d, want 0", stub.calls)
	}
}

func TestWrongBearerDomainUsesCorrectAuthenticatorAndSame401Shape(t *testing.T) {
	authenticationFailure := apierrors.New(apierrors.AuthenticationFailed, apierrors.Reauthenticate)
	token := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x34}, 32))
	account := &accountAuthenticatorStub{err: authenticationFailure}
	wrongDevice := &deviceAuthenticatorStub{}
	accountHandler := NewHandler(readiness.New(time.Millisecond), Applications{AccountAuth: account, DeviceAuth: wrongDevice}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x44}, 16)))
	accountRequest := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/password-changes", strings.NewReader(`{}`))
	accountRequest.Header.Set("Authorization", "Bearer "+token)
	accountRecorder := httptest.NewRecorder()
	accountHandler.ChangePassword(accountRecorder, accountRequest, controlapiv1.ChangePasswordParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})

	wrongAccount := &accountAuthenticatorStub{}
	device := &deviceAuthenticatorStub{err: authenticationFailure}
	deviceHandler := NewHandler(readiness.New(time.Millisecond), Applications{AccountAuth: wrongAccount, DeviceAuth: device}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x44}, 16)))
	deviceRequest := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/config-bundle-resolutions", strings.NewReader(`{}`))
	deviceRequest.Header.Set("Authorization", "Bearer "+token)
	deviceRecorder := httptest.NewRecorder()
	deviceHandler.ResolveConfigBundle(deviceRecorder, deviceRequest, controlapiv1.ResolveConfigBundleParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})

	if account.calls != 1 || wrongDevice.calls != 0 || wrongAccount.calls != 0 || device.calls != 1 {
		t.Fatalf("authenticator domain calls = account %d wrong-device %d wrong-account %d device %d", account.calls, wrongDevice.calls, wrongAccount.calls, device.calls)
	}
	if accountRecorder.Code != http.StatusUnauthorized || deviceRecorder.Code != http.StatusUnauthorized || accountRecorder.Body.String() != deviceRecorder.Body.String() {
		t.Fatalf("wrong-domain shapes = account %d %s device %d %s", accountRecorder.Code, accountRecorder.Body.String(), deviceRecorder.Code, deviceRecorder.Body.String())
	}
	if accountRecorder.Header().Get("WWW-Authenticate") != "Bearer" || deviceRecorder.Header().Get("WWW-Authenticate") != "Bearer" {
		t.Fatalf("wrong-domain challenges = account %q device %q", accountRecorder.Header().Get("WWW-Authenticate"), deviceRecorder.Header().Get("WWW-Authenticate"))
	}
}

func TestAuthTypedNilAuthenticatorsFailClosed(t *testing.T) {
	var account *panicAccountAuthenticator
	var device *panicDeviceAuthenticator
	handler := &Handler{apps: Applications{AccountAuth: account, DeviceAuth: device}}
	token := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x33}, 32))

	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	if _, err := handler.authenticateAccount(request); err == nil {
		t.Fatal("typed-nil account authenticator was accepted")
	}
	if _, err := handler.authenticateDevice(request); err == nil {
		t.Fatal("typed-nil device authenticator was accepted")
	}
}

func TestAuthRejectsMalformedBearerBeforeMissingAuthenticator(t *testing.T) {
	handler := &Handler{}
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", nil)
	request.Header.Set("Authorization", "Basic CANARY")
	_, accountErr := handler.authenticateAccount(request)
	var accountClassified apierrors.Error
	if !errors.As(accountErr, &accountClassified) || accountClassified.Public("trace-safe-000001").Code != controlapiv1.PublicErrorCodeAuthenticationFailed {
		t.Fatalf("account error = %v, want authentication_failed", accountErr)
	}
	_, deviceErr := handler.authenticateDevice(request)
	var deviceClassified apierrors.Error
	if !errors.As(deviceErr, &deviceClassified) || deviceClassified.Public("trace-safe-000001").Code != controlapiv1.PublicErrorCodeAuthenticationFailed {
		t.Fatalf("device error = %v, want authentication_failed", deviceErr)
	}
}

type panicAccountAuthenticator struct{}

func (*panicAccountAuthenticator) Authenticate(context.Context, secret.Bytes) (identity.AccountAuthority, error) {
	panic("typed nil account authenticator was invoked")
}

type panicDeviceAuthenticator struct{}

func (*panicDeviceAuthenticator) Authenticate(context.Context, secret.Bytes) (deviceauth.BundleAuthority, error) {
	panic("typed nil device authenticator was invoked")
}
