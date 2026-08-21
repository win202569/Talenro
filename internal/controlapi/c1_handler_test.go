package controlapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
	"talenro.local/platform/internal/deviceauth"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/readiness"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/trust"
)

var _ controlapiv1.ServerInterface = (*Handler)(nil)

func TestC1NilApplicationsFailClosedWithoutPlaceholderTrace(t *testing.T) {
	handler := NewHandler(readiness.New(time.Millisecond), Applications{}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x22}, 64)))
	router := controlapiv1.HandlerWithOptions(handler, controlapiv1.StdHTTPServerOptions{
		BaseRouter:       http.NewServeMux(),
		ErrorHandlerFunc: handler.GeneratedParameterError,
	})
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/accounts", strings.NewReader(`{"email":"person@example.test","password":"correct horse battery staple","locale":"en"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "abcdefghijklmnopqrstuv")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503; body=%s", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "c1-not-wired") || strings.Contains(recorder.Body.String(), "person@example.test") {
		t.Fatalf("placeholder or request secret leaked: %s", recorder.Body.String())
	}
}

func TestBodyGuardsRejectOversizeAndUnknownBeforeApplication(t *testing.T) {
	handler := NewHandler(readiness.New(time.Millisecond), Applications{}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x23}, 64)))
	router := controlapiv1.HandlerWithOptions(handler, controlapiv1.StdHTTPServerOptions{
		BaseRouter:       http.NewServeMux(),
		ErrorHandlerFunc: handler.GeneratedParameterError,
	})
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "unknown", body: `{"email":"person@example.test","password":"correct horse battery staple","locale":"en","extra":true}`, wantStatus: http.StatusBadRequest},
		{name: "duplicate", body: `{"email":"person@example.test","email":"other@example.test","password":"correct horse battery staple","locale":"en"}`, wantStatus: http.StatusBadRequest},
		{name: "trailing", body: `{"email":"person@example.test","password":"correct horse battery staple","locale":"en"}{}`, wantStatus: http.StatusBadRequest},
		{name: "oversize", body: strings.Repeat("x", (64<<10)+1), wantStatus: http.StatusRequestEntityTooLarge},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/accounts", strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Idempotency-Key", "abcdefghijklmnopqrstuv")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, test.wantStatus, recorder.Body.String())
			}
		})
	}
}

func TestBodyGuardAcceptsExactLimitAndRejectsNextByte(t *testing.T) {
	capture := &identityApplicationCapture{}
	handler := NewHandler(readiness.New(time.Millisecond), Applications{Identity: capture}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x23}, 64)))
	base := `{"email":"person@example.test","password":"correct horse battery staple","locale":"en"}`
	exact := base + strings.Repeat(" ", (64<<10)-len(base))

	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/accounts", strings.NewReader(exact))
	recorder := httptest.NewRecorder()
	handler.CreateAccount(recorder, request, controlapiv1.CreateAccountParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
	if recorder.Code != http.StatusAccepted || capture.registerCalls != 1 {
		t.Fatalf("exact limit = status %d, calls %d; body=%s", recorder.Code, capture.registerCalls, recorder.Body.String())
	}

	request = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/accounts", strings.NewReader(exact+" "))
	recorder = httptest.NewRecorder()
	handler.CreateAccount(recorder, request, controlapiv1.CreateAccountParams{IdempotencyKey: "bcdefghijklmnopqrstuvw"})
	if recorder.Code != http.StatusRequestEntityTooLarge || capture.registerCalls != 1 {
		t.Fatalf("next byte = status %d, calls %d; body=%s", recorder.Code, capture.registerCalls, recorder.Body.String())
	}
}

func TestC1GeneratedParameterErrorsAreSanitized(t *testing.T) {
	handler := NewHandler(readiness.New(time.Millisecond), Applications{}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x24}, 64)))
	router := controlapiv1.HandlerWithOptions(handler, controlapiv1.StdHTTPServerOptions{
		BaseRouter:       http.NewServeMux(),
		ErrorHandlerFunc: handler.GeneratedParameterError,
	})
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/accounts", strings.NewReader(`{}`))
	request.Header.Add("Idempotency-Key", "abcdefghijklmnopqrstuv")
	request.Header.Add("Idempotency-Key", "CANARY-duplicate-header")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "malformed_request") {
		t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "CANARY") || strings.Contains(recorder.Body.String(), "duplicate") {
		t.Fatalf("raw generated error leaked: %s", recorder.Body.String())
	}

	request = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/accounts", strings.NewReader(`{}`))
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "malformed_request") || strings.Contains(recorder.Body.String(), "Idempotency-Key") {
		t.Fatalf("missing generated parameter response = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestC1AllGeneratedOperationsReachConcreteHandler(t *testing.T) {
	handler := NewHandler(readiness.New(time.Millisecond), Applications{}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x27}, 16*64)))
	router := controlapiv1.HandlerWithOptions(handler, controlapiv1.StdHTTPServerOptions{
		BaseRouter: http.NewServeMux(), ErrorHandlerFunc: handler.GeneratedParameterError,
	})
	paths := []string{
		"/v1/account-auth-challenges", "/v1/account-session-revocations", "/v1/account-sessions",
		"/v1/account-token-rotations", "/v1/accounts", "/v1/config-bundle-acknowledgements",
		"/v1/config-bundle-resolutions", "/v1/device-auth-challenges", "/v1/device-enrollment-grants",
		"/v1/device-revocations", "/v1/device-token-rotations", "/v1/devices",
		"/v1/email-verification-deliveries", "/v1/email-verifications", "/v1/passkey-authentication-options",
		"/v1/passkey-credentials", "/v1/passkey-registration-options", "/v1/passkey-revocations",
		"/v1/password-changes", "/v1/password-reset-deliveries", "/v1/password-resets",
		"/v1/recovery-code-consumptions", "/v1/recovery-code-rotations", "/v1/totp-enrollments",
		"/v1/totp-revocations", "/v1/totp-verifications",
	}
	for _, path := range paths {
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, strings.NewReader(`{}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Idempotency-Key", "abcdefghijklmnopqrstuv")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code == http.StatusNotFound || recorder.Code == http.StatusMethodNotAllowed {
			t.Errorf("POST %s did not reach a concrete operation: %d", path, recorder.Code)
		}
		if strings.Contains(recorder.Body.String(), "c1-not-wired") {
			t.Errorf("POST %s returned the deleted placeholder", path)
		}
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/b/abcdefghijklmnopqrstuv", nil))
	if recorder.Code == http.StatusNotFound || strings.Contains(recorder.Body.String(), "c1-not-wired") {
		t.Fatalf("immutable operation response = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestC1OperationBehaviorMatrix(t *testing.T) {
	type operationCase struct {
		OperationID         string
		Method              string
		Path                string
		AuthDomain          string
		RequiresIdempotency bool
		WantStatus          int
		Body                string
	}
	encoded32 := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x41}, 32))
	encoded64 := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x42}, 64))
	credentialID := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x43}, 16))
	webAuthnData := base64.RawURLEncoding.EncodeToString([]byte{1})
	reauth := `{"reauthentication":{"method":"password","password":"correct horse battery staple"}}`
	registrationResponse := `{"id":"` + credentialID + `","rawId":"` + credentialID + `","type":"public-key","response":{"clientDataJSON":"` + webAuthnData + `","attestationObject":"` + webAuthnData + `"}}`
	cases := []operationCase{
		{OperationID: "createAccount", Method: http.MethodPost, Path: "/v1/accounts", RequiresIdempotency: true, WantStatus: 202, Body: `{"email":"a@example.com","password":"correct horse battery staple","locale":"en"}`},
		{OperationID: "createEmailVerificationDelivery", Method: http.MethodPost, Path: "/v1/email-verification-deliveries", RequiresIdempotency: true, WantStatus: 202, Body: `{"email":"a@example.com","locale":"en"}`},
		{OperationID: "verifyEmail", Method: http.MethodPost, Path: "/v1/email-verifications", RequiresIdempotency: true, WantStatus: 204, Body: `{"token":"` + encoded32 + `"}`},
		{OperationID: "createPasswordResetDelivery", Method: http.MethodPost, Path: "/v1/password-reset-deliveries", RequiresIdempotency: true, WantStatus: 202, Body: `{"email":"a@example.com","locale":"en"}`},
		{OperationID: "resetPassword", Method: http.MethodPost, Path: "/v1/password-resets", RequiresIdempotency: true, WantStatus: 200, Body: `{"email":"a@example.com","token":"` + encoded32 + `","new_password":"correct horse battery staple","client_signing_public_key":"` + encoded32 + `"}`},
		{OperationID: "changePassword", Method: http.MethodPost, Path: "/v1/password-changes", AuthDomain: "account", RequiresIdempotency: true, WantStatus: 200, Body: `{"current_password":"correct horse battery staple","new_password":"another correct horse battery staple","client_signing_public_key":"` + encoded32 + `",` + strings.TrimPrefix(reauth, "{")},
		{OperationID: "createAccountSession", Method: http.MethodPost, Path: "/v1/account-sessions", RequiresIdempotency: true, WantStatus: 200, Body: `{"method":"password","email":"a@example.com","password":"correct horse battery staple","client_signing_public_key":"` + encoded32 + `"}`},
		{OperationID: "createAccountAuthChallenge", Method: http.MethodPost, Path: "/v1/account-auth-challenges", RequiresIdempotency: true, WantStatus: 201, Body: `{"refresh_token":"` + encoded32 + `","request_nonce":"` + encoded32 + `"}`},
		{OperationID: "rotateAccountToken", Method: http.MethodPost, Path: "/v1/account-token-rotations", RequiresIdempotency: true, WantStatus: 200, Body: `{"challenge_id":"d12dca8a-ced3-471d-a6ad-55b228221f10","refresh_token":"` + encoded32 + `","request_nonce":"` + encoded32 + `","signature":"` + encoded64 + `"}`},
		{OperationID: "revokeAccountSessions", Method: http.MethodPost, Path: "/v1/account-session-revocations", AuthDomain: "account", RequiresIdempotency: true, WantStatus: 204, Body: `{"scope":"all",` + strings.TrimPrefix(reauth, "{")},
		{OperationID: "createPasskeyRegistrationOptions", Method: http.MethodPost, Path: "/v1/passkey-registration-options", AuthDomain: "account", RequiresIdempotency: true, WantStatus: 200, Body: reauth},
		{OperationID: "createPasskeyCredential", Method: http.MethodPost, Path: "/v1/passkey-credentials", AuthDomain: "account", RequiresIdempotency: true, WantStatus: 204, Body: `{"ceremony_id":"1645ba9c-ad1d-4a06-b996-d9feb78ee88a","response":` + registrationResponse + `,` + strings.TrimPrefix(reauth, "{")},
		{OperationID: "createPasskeyAuthenticationOptions", Method: http.MethodPost, Path: "/v1/passkey-authentication-options", RequiresIdempotency: true, WantStatus: 200, Body: `{}`},
		{OperationID: "revokePasskey", Method: http.MethodPost, Path: "/v1/passkey-revocations", AuthDomain: "account", RequiresIdempotency: true, WantStatus: 204, Body: `{"credential_id":"` + credentialID + `",` + strings.TrimPrefix(reauth, "{")},
		{OperationID: "createTOTPEnrollment", Method: http.MethodPost, Path: "/v1/totp-enrollments", AuthDomain: "account", RequiresIdempotency: true, WantStatus: 201, Body: reauth},
		{OperationID: "verifyTOTPEnrollment", Method: http.MethodPost, Path: "/v1/totp-verifications", AuthDomain: "account", RequiresIdempotency: true, WantStatus: 204, Body: `{"code":"123456",` + strings.TrimPrefix(reauth, "{")},
		{OperationID: "revokeTOTP", Method: http.MethodPost, Path: "/v1/totp-revocations", AuthDomain: "account", RequiresIdempotency: true, WantStatus: 204, Body: reauth},
		{OperationID: "rotateRecoveryCodes", Method: http.MethodPost, Path: "/v1/recovery-code-rotations", AuthDomain: "account", RequiresIdempotency: true, WantStatus: 201, Body: reauth},
		{OperationID: "consumeRecoveryCode", Method: http.MethodPost, Path: "/v1/recovery-code-consumptions", RequiresIdempotency: true, WantStatus: 200, Body: `{"email":"a@example.com","code":"ABCD-EFGH-IJKL-MNOP-QRST-UVWX-YZ23-4567","client_signing_public_key":"` + encoded32 + `"}`},
		{OperationID: "createDeviceEnrollmentGrant", Method: http.MethodPost, Path: "/v1/device-enrollment-grants", AuthDomain: "account", RequiresIdempotency: true, WantStatus: 201, Body: reauth},
		{OperationID: "createDeviceAuthChallenge", Method: http.MethodPost, Path: "/v1/device-auth-challenges", RequiresIdempotency: true, WantStatus: 201, Body: `{"enrollment_grant":"` + encoded32 + `","request_nonce":"` + encoded32 + `","signing_public_key":"` + encoded32 + `","hpke_public_key":"` + encoded32 + `"}`},
		{OperationID: "registerDevice", Method: http.MethodPost, Path: "/v1/devices", RequiresIdempotency: true, WantStatus: 201, Body: `{"challenge_id":"d12dca8a-ced3-471d-a6ad-55b228221f10","display_name":"device","enrollment_grant":"` + encoded32 + `","request_nonce":"` + encoded32 + `","signing_public_key":"` + encoded32 + `","hpke_public_key":"` + encoded32 + `","signature":"` + encoded64 + `"}`},
		{OperationID: "rotateDeviceToken", Method: http.MethodPost, Path: "/v1/device-token-rotations", RequiresIdempotency: true, WantStatus: 200, Body: `{"challenge_id":"d12dca8a-ced3-471d-a6ad-55b228221f10","refresh_token":"` + encoded32 + `","request_nonce":"` + encoded32 + `","signature":"` + encoded64 + `"}`},
		{OperationID: "revokeDevice", Method: http.MethodPost, Path: "/v1/device-revocations", AuthDomain: "account", RequiresIdempotency: true, WantStatus: 204, Body: `{"device_id":"1645ba9c-ad1d-4a06-b996-d9feb78ee88a",` + strings.TrimPrefix(reauth, "{")},
		{OperationID: "resolveConfigBundle", Method: http.MethodPost, Path: "/v1/config-bundle-resolutions", AuthDomain: "device", RequiresIdempotency: true, WantStatus: 200, Body: `{}`},
		{OperationID: "acknowledgeConfigBundle", Method: http.MethodPost, Path: "/v1/config-bundle-acknowledgements", AuthDomain: "device", RequiresIdempotency: true, WantStatus: 204, Body: `{"bundle_id":"1645ba9c-ad1d-4a06-b996-d9feb78ee88a","bundle_version":"1"}`},
		{OperationID: "getImmutableBundle", Method: http.MethodGet, Path: "/b/abcdefghijklmnopqrstuv", WantStatus: 200},
	}
	if len(cases) != len(c11Operations) {
		t.Fatalf("operation matrix contains %d cases, want %d", len(cases), len(c11Operations))
	}
	principal := uuid.MustParse("28ceee8a-5f4f-4d3a-9e9f-d3ec23b815ac")
	session := identity.SessionID("d12dca8a-ced3-471d-a6ad-55b228221f10")
	token := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x44}, 32))
	for _, test := range cases {
		t.Run(test.OperationID, func(t *testing.T) {
			capture := &operationCapture{}
			accountAuth := &accountAuthenticatorStub{authority: identity.AccountAuthority{PrincipalID: identity.PrincipalID(principal.String()), SessionID: session}}
			deviceAuth := &deviceAuthenticatorStub{authority: deviceauth.BundleAuthority{PrincipalID: principal, AuthorizationID: principal}}
			immutable := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				capture.record("getImmutableBundle")
				w.WriteHeader(http.StatusOK)
			})
			handler := NewHandler(readiness.New(time.Millisecond), Applications{
				Identity: operationIdentityApplication{capture}, StrongAuth: operationStrongAuthApplication{capture}, Device: operationDeviceApplication{capture}, Trust: operationTrustApplication{capture},
				AccountAuth: accountAuth, DeviceAuth: deviceAuth, ImmutableBundle: immutable,
			}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x45}, 16*64)))
			router := controlapiv1.HandlerWithOptions(handler, controlapiv1.StdHTTPServerOptions{BaseRouter: http.NewServeMux(), ErrorHandlerFunc: handler.GeneratedParameterError})
			request := httptest.NewRequestWithContext(t.Context(), test.Method, test.Path, strings.NewReader(test.Body))
			if test.Method == http.MethodPost {
				request.Header.Set("Content-Type", "application/json")
			}
			if test.AuthDomain != "" {
				request.Header.Set("Authorization", "Bearer "+token)
			}
			if test.RequiresIdempotency {
				request.Header.Set("Idempotency-Key", "abcdefghijklmnopqrstuv")
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != test.WantStatus || capture.operation != test.OperationID || test.RequiresIdempotency && capture.idempotency != "abcdefghijklmnopqrstuv" {
				t.Fatalf("%s %s = %d operation %q idempotency %q, want %d %q exact key; body=%s", test.Method, test.Path, recorder.Code, capture.operation, capture.idempotency, test.WantStatus, test.OperationID, recorder.Body.String())
			}
			if test.RequiresIdempotency {
				capture.operation, capture.idempotency = "", ""
				missing := httptest.NewRequestWithContext(t.Context(), test.Method, test.Path, strings.NewReader(test.Body))
				missing.Header.Set("Content-Type", "application/json")
				if test.AuthDomain != "" {
					missing.Header.Set("Authorization", "Bearer "+token)
				}
				missingRecorder := httptest.NewRecorder()
				router.ServeHTTP(missingRecorder, missing)
				if missingRecorder.Code != http.StatusBadRequest || capture.operation != "" {
					t.Fatalf("missing idempotency %s = %d operation %q; body=%s", test.Path, missingRecorder.Code, capture.operation, missingRecorder.Body.String())
				}
			}
			nilHandler := NewHandler(readiness.New(time.Millisecond), Applications{AccountAuth: accountAuth, DeviceAuth: deviceAuth}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x46}, 16*64)))
			nilRouter := controlapiv1.HandlerWithOptions(nilHandler, controlapiv1.StdHTTPServerOptions{BaseRouter: http.NewServeMux(), ErrorHandlerFunc: nilHandler.GeneratedParameterError})
			nilRequest := httptest.NewRequestWithContext(t.Context(), test.Method, test.Path, strings.NewReader(test.Body))
			if test.Method == http.MethodPost {
				nilRequest.Header.Set("Content-Type", "application/json")
			}
			if test.AuthDomain != "" {
				nilRequest.Header.Set("Authorization", "Bearer "+token)
			}
			if test.RequiresIdempotency {
				nilRequest.Header.Set("Idempotency-Key", "abcdefghijklmnopqrstuv")
			}
			nilRecorder := httptest.NewRecorder()
			nilRouter.ServeHTTP(nilRecorder, nilRequest)
			if nilRecorder.Code != http.StatusServiceUnavailable || !strings.Contains(nilRecorder.Body.String(), "dependency_unavailable") {
				t.Fatalf("valid nil dependency %s = %d %s, want sanitized 503", test.Path, nilRecorder.Code, nilRecorder.Body.String())
			}
			if test.AuthDomain == "" {
				return
			}
			capture.operation, capture.idempotency = "", ""
			if test.AuthDomain == "account" {
				accountAuth.err = authenticationFailedError()
			} else {
				deviceAuth.err = authenticationFailedError()
			}
			wrong := httptest.NewRequestWithContext(t.Context(), test.Method, test.Path, strings.NewReader(test.Body))
			wrong.Header.Set("Content-Type", "application/json")
			wrong.Header.Set("Authorization", "Bearer "+token)
			wrong.Header.Set("Idempotency-Key", "abcdefghijklmnopqrstuv")
			wrongRecorder := httptest.NewRecorder()
			router.ServeHTTP(wrongRecorder, wrong)
			if wrongRecorder.Code != http.StatusUnauthorized || capture.operation != "" || wrongRecorder.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Fatalf("wrong-domain %s = %d operation %q challenge %q; body=%s", test.Path, wrongRecorder.Code, capture.operation, wrongRecorder.Header().Get("WWW-Authenticate"), wrongRecorder.Body.String())
			}
		})
	}
}

func TestC1EveryWriteRouteRejectsMalformedIdempotencyBeforeTargetDependency(t *testing.T) {
	principal := uuid.MustParse("28ceee8a-5f4f-4d3a-9e9f-d3ec23b815ac")
	session := uuid.MustParse("d12dca8a-ced3-471d-a6ad-55b228221f10")
	handler := NewHandler(readiness.New(time.Millisecond), Applications{
		AccountAuth: &accountAuthenticatorStub{authority: identity.AccountAuthority{
			PrincipalID: identity.PrincipalID(principal.String()), SessionID: identity.SessionID(session.String()),
		}},
		DeviceAuth: &deviceAuthenticatorStub{authority: deviceauth.BundleAuthority{PrincipalID: principal}},
	}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x2a}, 16*64)))
	router := controlapiv1.HandlerWithOptions(handler, controlapiv1.StdHTTPServerOptions{
		BaseRouter: http.NewServeMux(), ErrorHandlerFunc: handler.GeneratedParameterError,
	})
	paths := []string{
		"/v1/account-auth-challenges", "/v1/account-session-revocations", "/v1/account-sessions",
		"/v1/account-token-rotations", "/v1/accounts", "/v1/config-bundle-acknowledgements",
		"/v1/config-bundle-resolutions", "/v1/device-auth-challenges", "/v1/device-enrollment-grants",
		"/v1/device-revocations", "/v1/device-token-rotations", "/v1/devices",
		"/v1/email-verification-deliveries", "/v1/email-verifications", "/v1/passkey-authentication-options",
		"/v1/passkey-credentials", "/v1/passkey-registration-options", "/v1/passkey-revocations",
		"/v1/password-changes", "/v1/password-reset-deliveries", "/v1/password-resets",
		"/v1/recovery-code-consumptions", "/v1/recovery-code-rotations", "/v1/totp-enrollments",
		"/v1/totp-revocations", "/v1/totp-verifications",
	}
	token := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x52}, 32))
	for _, path := range paths {
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, strings.NewReader(`{}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+token)
		request.Header.Set("Idempotency-Key", "CANARY/invalid")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "malformed_request") {
			t.Errorf("POST %s = %d %s, want sanitized 400", path, recorder.Code, recorder.Body.String())
		}
		if strings.Contains(recorder.Body.String(), "CANARY") {
			t.Errorf("POST %s leaked idempotency key: %s", path, recorder.Body.String())
		}
	}
}

func TestEveryPOSTRejectsMalformedBodyBeforeUnavailableTarget(t *testing.T) {
	principal := uuid.MustParse("28ceee8a-5f4f-4d3a-9e9f-d3ec23b815ac")
	handler := NewHandler(readiness.New(time.Millisecond), Applications{
		AccountAuth: &accountAuthenticatorStub{authority: identity.AccountAuthority{
			PrincipalID: identity.PrincipalID(principal.String()), SessionID: identity.SessionID("d12dca8a-ced3-471d-a6ad-55b228221f10"),
		}},
		DeviceAuth: &deviceAuthenticatorStub{authority: deviceauth.BundleAuthority{PrincipalID: principal, AuthorizationID: principal}},
	}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x26}, 16*64)))
	router := controlapiv1.HandlerWithOptions(handler, controlapiv1.StdHTTPServerOptions{BaseRouter: http.NewServeMux(), ErrorHandlerFunc: handler.GeneratedParameterError})
	paths := []string{
		"/v1/account-auth-challenges", "/v1/account-session-revocations", "/v1/account-sessions", "/v1/account-token-rotations", "/v1/accounts",
		"/v1/config-bundle-acknowledgements", "/v1/config-bundle-resolutions", "/v1/device-auth-challenges", "/v1/device-enrollment-grants",
		"/v1/device-revocations", "/v1/device-token-rotations", "/v1/devices", "/v1/email-verification-deliveries", "/v1/email-verifications",
		"/v1/passkey-authentication-options", "/v1/passkey-credentials", "/v1/passkey-registration-options", "/v1/passkey-revocations",
		"/v1/password-changes", "/v1/password-reset-deliveries", "/v1/password-resets", "/v1/recovery-code-consumptions",
		"/v1/recovery-code-rotations", "/v1/totp-enrollments", "/v1/totp-revocations", "/v1/totp-verifications",
	}
	token := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x25}, 32))
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, strings.NewReader(`{"unknown":true}`))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Authorization", "Bearer "+token)
			request.Header.Set("Idempotency-Key", "abcdefghijklmnopqrstuv")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "malformed_request") {
				t.Fatalf("POST %s = %d %s, want sanitized 400", path, recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestC1ImmutableBundleDelegatesPublicRequest(t *testing.T) {
	called := false
	immutable := http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		called = true
		if request.URL.Path != "/b/task17-locator" || request.Header.Get("Authorization") != "" {
			t.Fatalf("immutable request = %s authorization=%q", request.URL.Path, request.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
	})
	handler := NewHandler(readiness.New(time.Millisecond), Applications{ImmutableBundle: immutable}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x25}, 64)))
	router := controlapiv1.HandlerWithOptions(handler, controlapiv1.StdHTTPServerOptions{BaseRouter: http.NewServeMux(), ErrorHandlerFunc: handler.GeneratedParameterError})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/b/task17-locator", nil))
	if recorder.Code != http.StatusOK || !called {
		t.Fatalf("immutable delegation = status %d, called %v", recorder.Code, called)
	}
}

func TestDeadlineCancellationReachesApplicationBoundary(t *testing.T) {
	capture := &identityApplicationCapture{waitForDeadline: true}
	handler := NewHandler(readiness.New(time.Millisecond), Applications{Identity: capture}, time.Millisecond, bytes.NewReader(bytes.Repeat([]byte{0x26}, 64)))
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/accounts", strings.NewReader(`{"email":"person@example.test","password":"correct horse battery staple","locale":"en"}`))
	request.Header.Set("Idempotency-Key", "abcdefghijklmnopqrstuv")
	recorder := httptest.NewRecorder()
	handler.CreateAccount(recorder, request, controlapiv1.CreateAccountParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
	if recorder.Code != http.StatusServiceUnavailable || !capture.sawDeadline || capture.registerCalls != 1 {
		t.Fatalf("deadline boundary status = %d, calls %d, deadline %v; body=%s", recorder.Code, capture.registerCalls, capture.sawDeadline, recorder.Body.String())
	}
}

type identityApplicationCapture struct {
	identity.Application
	revoke          identity.RevokeSessionsCommand
	revokeCalls     int
	registerCalls   int
	rotateCalls     int
	waitForDeadline bool
	sawDeadline     bool
}

func (capture *identityApplicationCapture) RotateSession(_ context.Context, _ identity.RotateSessionCommand) (identity.SessionTokens, error) {
	capture.rotateCalls++
	return identity.SessionTokens{}, nil
}

func (*identityApplicationCapture) CreateSessionChallenge(_ context.Context, _ identity.CreateSessionChallengeCommand) (identity.SessionChallenge, error) {
	return identity.SessionChallenge{
		ChallengeID: "d12dca8a-ced3-471d-a6ad-55b228221f10", Challenge: [32]byte{1}, ExpiresAt: time.Now().Add(time.Minute),
	}, nil
}

func (*identityApplicationCapture) CreateEnrollmentGrant(_ context.Context, _ identity.CreateEnrollmentGrantCommand) (identity.EnrollmentGrant, error) {
	return identity.EnrollmentGrant{Token: secret.NewBytes(bytes.Repeat([]byte{1}, 32)), ExpiresAt: time.Now().Add(time.Minute)}, nil
}

func (capture *identityApplicationCapture) RegisterAccount(ctx context.Context, _ identity.RegisterAccountCommand) (identity.RegisterAccountResult, error) {
	capture.registerCalls++
	_, capture.sawDeadline = ctx.Deadline()
	if capture.waitForDeadline {
		<-ctx.Done()
		return identity.RegisterAccountResult{}, ctx.Err()
	}
	return identity.RegisterAccountResult{Accepted: true}, nil
}

func (capture *identityApplicationCapture) RevokeSessions(_ context.Context, command identity.RevokeSessionsCommand) error {
	capture.revoke = command
	capture.revokeCalls++
	return nil
}

type deviceApplicationCapture struct {
	deviceauth.Application
	registerCalls int
	rotateCalls   int
}

func (capture *deviceApplicationCapture) RotateDeviceToken(_ context.Context, _ deviceauth.RotateDeviceTokenCommand) (deviceauth.DeviceTokens, error) {
	capture.rotateCalls++
	return deviceauth.DeviceTokens{}, nil
}

func (*deviceApplicationCapture) CreateChallenge(_ context.Context, _ deviceauth.CreateChallengeCommand) (deviceauth.Challenge, error) {
	return deviceauth.Challenge{
		ChallengeID: "d12dca8a-ced3-471d-a6ad-55b228221f10", Challenge: [32]byte{1}, ExpiresAt: time.Now().Add(time.Minute),
	}, nil
}

func (capture *deviceApplicationCapture) RegisterDevice(_ context.Context, _ deviceauth.RegisterDeviceCommand) (deviceauth.DeviceTokens, error) {
	capture.registerCalls++
	return deviceauth.DeviceTokens{
		DeviceID: uuid.MustParse("1645ba9c-ad1d-4a06-b996-d9feb78ee88a"), AuthorizationID: uuid.MustParse("28ceee8a-5f4f-4d3a-9e9f-d3ec23b815ac"),
		FamilyID:    uuid.MustParse("0ff820a5-5022-48e6-8867-77761f8e2f07"),
		AccessToken: secret.NewBytes(bytes.Repeat([]byte{2}, 32)), RefreshToken: secret.NewBytes(bytes.Repeat([]byte{3}, 32)), AccessExpiresAt: time.Now().Add(time.Minute),
	}, nil
}

func TestDeviceTokensBodyPublishesAndRequiresFamilyID(t *testing.T) {
	t.Parallel()

	tokens := testDeviceTokens()
	defer tokens.AccessToken.Clear()
	defer tokens.RefreshToken.Clear()
	body, err := deviceTokensBody(tokens)
	if err != nil {
		t.Fatal(err)
	}
	if body.FamilyId != tokens.FamilyID {
		t.Fatalf("family_id = %s, want %s", body.FamilyId, tokens.FamilyID)
	}

	missing := testDeviceTokens()
	defer missing.AccessToken.Clear()
	defer missing.RefreshToken.Clear()
	missing.FamilyID = uuid.Nil
	if _, err := deviceTokensBody(missing); err == nil {
		t.Fatal("deviceTokensBody accepted a missing family ID")
	}
}

func TestRegisterDeviceRejectsMoreThanSixtyFourUnicodeScalars(t *testing.T) {
	capture := &deviceApplicationCapture{}
	handler := NewHandler(readiness.New(time.Millisecond), Applications{Device: capture}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x29}, 64)))
	encoded32 := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x31}, 32))
	encoded64 := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x32}, 64))
	body, err := json.Marshal(map[string]any{
		"challenge_id": "d12dca8a-ced3-471d-a6ad-55b228221f10", "display_name": strings.Repeat("é", 65),
		"enrollment_grant": encoded32, "request_nonce": encoded32, "signing_public_key": encoded32,
		"hpke_public_key": encoded32, "signature": encoded64,
	})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/devices", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	handler.RegisterDevice(recorder, request, controlapiv1.RegisterDeviceParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
	if recorder.Code != http.StatusBadRequest || capture.registerCalls != 0 {
		t.Fatalf("unicode scalar limit = status %d, calls %d; body=%s", recorder.Code, capture.registerCalls, recorder.Body.String())
	}
}

type strongAuthCapture struct {
	identity.StrongAuthApplication
	registration              identity.BeginPasskeyRegistrationCommand
	authentication            identity.BeginPasskeyAuthenticationCommand
	verifyTOTPCalls           int
	finishRegistrationCalls   int
	finishAuthenticationCalls int
	beginTOTPCalls            int
}

func (capture *strongAuthCapture) FinishPasskeyRegistration(_ context.Context, _ identity.FinishPasskeyRegistrationCommand) error {
	capture.finishRegistrationCalls++
	return nil
}

func (capture *strongAuthCapture) FinishPasskeyAuthentication(_ context.Context, _ identity.FinishPasskeyAuthenticationCommand) (identity.SessionTokens, error) {
	capture.finishAuthenticationCalls++
	return identity.SessionTokens{}, nil
}

func (capture *strongAuthCapture) BeginPasskeyRegistration(_ context.Context, command identity.BeginPasskeyRegistrationCommand) (json.RawMessage, error) {
	capture.registration = command
	return json.RawMessage(`{"ceremony_id":"d12dca8a-ced3-471d-a6ad-55b228221f10","publicKey":{}}`), nil
}

func (capture *strongAuthCapture) BeginPasskeyAuthentication(_ context.Context, command identity.BeginPasskeyAuthenticationCommand) (json.RawMessage, error) {
	capture.authentication = command
	return json.RawMessage(`{"ceremony_id":"d12dca8a-ced3-471d-a6ad-55b228221f10","publicKey":{}}`), nil
}

func (capture *strongAuthCapture) BeginTOTPEnrollment(_ context.Context, _ identity.BeginTOTPEnrollmentCommand) (identity.TOTPEnrollment, error) {
	capture.beginTOTPCalls++
	const fixture = "JBSWY3DPEHPK3PXP" // #nosec G101 -- deterministic non-production test fixture.
	return identity.TOTPEnrollment{Secret: fixture, URI: "otpauth://totp/Talenro:account?secret=" + fixture}, nil
}

func (*strongAuthCapture) RotateRecoveryCodes(_ context.Context, _ identity.RotateRecoveryCodesCommand) (identity.RecoveryCodes, error) {
	return identity.RecoveryCodes{Codes: []string{"ABCD-EFGH-IJKL-MNOP-QRST-UVWX-YZ23-4567"}}, nil
}

func (capture *strongAuthCapture) VerifyTOTPEnrollment(_ context.Context, _ identity.VerifyTOTPEnrollmentCommand) error {
	capture.verifyTOTPCalls++
	return nil
}

type trustCapture struct {
	trust.Application
	query            trust.ResolveQuery
	acknowledgeCalls int
}

type operationCapture struct {
	operation   string
	idempotency string
}

func (capture *operationCapture) record(operation string) { capture.operation = operation }

func (capture *operationCapture) recordCommand(operation string, command any) {
	capture.operation = operation
	value := reflect.ValueOf(command)
	field := value.FieldByName("IdempotencyKey")
	if field.IsValid() && field.Kind() == reflect.String {
		capture.idempotency = field.String()
	}
}

type operationIdentityApplication struct{ capture *operationCapture }

func testAccountTokens() identity.SessionTokens {
	return identity.SessionTokens{AccessToken: secret.NewBytes(bytes.Repeat([]byte{0x71}, 32)), RefreshToken: secret.NewBytes(bytes.Repeat([]byte{0x72}, 32)), AccessExpiresAt: time.Now().Add(time.Minute)}
}

func (application operationIdentityApplication) RegisterAccount(_ context.Context, command identity.RegisterAccountCommand) (identity.RegisterAccountResult, error) {
	application.capture.recordCommand("createAccount", command)
	return identity.RegisterAccountResult{Accepted: true}, nil
}
func (application operationIdentityApplication) CreateEmailVerificationDelivery(_ context.Context, command identity.CreateEmailVerificationDeliveryCommand) (identity.RegisterAccountResult, error) {
	application.capture.recordCommand("createEmailVerificationDelivery", command)
	return identity.RegisterAccountResult{Accepted: true}, nil
}
func (application operationIdentityApplication) VerifyEmail(_ context.Context, command identity.VerifyEmailCommand) error {
	application.capture.recordCommand("verifyEmail", command)
	return nil
}
func (application operationIdentityApplication) CreatePasswordResetDelivery(_ context.Context, command identity.CreatePasswordResetDeliveryCommand) (identity.RegisterAccountResult, error) {
	application.capture.recordCommand("createPasswordResetDelivery", command)
	return identity.RegisterAccountResult{Accepted: true}, nil
}
func (application operationIdentityApplication) ResetPassword(_ context.Context, command identity.ResetPasswordCommand) (identity.SessionTokens, error) {
	application.capture.recordCommand("resetPassword", command)
	return testAccountTokens(), nil
}
func (application operationIdentityApplication) ChangePassword(_ context.Context, command identity.ChangePasswordCommand) (identity.SessionTokens, error) {
	application.capture.recordCommand("changePassword", command)
	return testAccountTokens(), nil
}
func (application operationIdentityApplication) CreateSession(_ context.Context, command identity.CreateSessionCommand) (identity.SessionTokens, error) {
	application.capture.recordCommand("createAccountSession", command)
	return testAccountTokens(), nil
}
func (application operationIdentityApplication) CreateSessionChallenge(_ context.Context, command identity.CreateSessionChallengeCommand) (identity.SessionChallenge, error) {
	application.capture.recordCommand("createAccountAuthChallenge", command)
	return identity.SessionChallenge{ChallengeID: "d12dca8a-ced3-471d-a6ad-55b228221f10", Challenge: [32]byte{1}, ExpiresAt: time.Now().Add(time.Minute)}, nil
}
func (application operationIdentityApplication) RotateSession(_ context.Context, command identity.RotateSessionCommand) (identity.SessionTokens, error) {
	application.capture.recordCommand("rotateAccountToken", command)
	return testAccountTokens(), nil
}
func (application operationIdentityApplication) RevokeSessions(_ context.Context, command identity.RevokeSessionsCommand) error {
	application.capture.recordCommand("revokeAccountSessions", command)
	return nil
}
func (application operationIdentityApplication) CreateEnrollmentGrant(_ context.Context, command identity.CreateEnrollmentGrantCommand) (identity.EnrollmentGrant, error) {
	application.capture.recordCommand("createDeviceEnrollmentGrant", command)
	return identity.EnrollmentGrant{Token: secret.NewBytes(bytes.Repeat([]byte{0x73}, 32)), ExpiresAt: time.Now().Add(time.Minute)}, nil
}

type operationStrongAuthApplication struct{ capture *operationCapture }

func (application operationStrongAuthApplication) BeginPasskeyRegistration(_ context.Context, command identity.BeginPasskeyRegistrationCommand) (json.RawMessage, error) {
	application.capture.recordCommand("createPasskeyRegistrationOptions", command)
	return json.RawMessage(`{"ceremony_id":"d12dca8a-ced3-471d-a6ad-55b228221f10","publicKey":{}}`), nil
}
func (application operationStrongAuthApplication) FinishPasskeyRegistration(_ context.Context, command identity.FinishPasskeyRegistrationCommand) error {
	application.capture.recordCommand("createPasskeyCredential", command)
	return nil
}
func (application operationStrongAuthApplication) BeginPasskeyAuthentication(_ context.Context, command identity.BeginPasskeyAuthenticationCommand) (json.RawMessage, error) {
	application.capture.recordCommand("createPasskeyAuthenticationOptions", command)
	return json.RawMessage(`{"ceremony_id":"d12dca8a-ced3-471d-a6ad-55b228221f10","publicKey":{}}`), nil
}
func (application operationStrongAuthApplication) FinishPasskeyAuthentication(_ context.Context, command identity.FinishPasskeyAuthenticationCommand) (identity.SessionTokens, error) {
	application.capture.recordCommand("createAccountSession", command)
	return testAccountTokens(), nil
}
func (application operationStrongAuthApplication) RevokePasskey(_ context.Context, command identity.RevokePasskeyCommand) error {
	application.capture.recordCommand("revokePasskey", command)
	return nil
}
func (application operationStrongAuthApplication) BeginTOTPEnrollment(_ context.Context, command identity.BeginTOTPEnrollmentCommand) (identity.TOTPEnrollment, error) {
	application.capture.recordCommand("createTOTPEnrollment", command)
	return identity.TOTPEnrollment{Secret: "JBSWY3DPEHPK3PXP", URI: "otpauth://totp/Talenro:account?secret=JBSWY3DPEHPK3PXP"}, nil // #nosec G101 -- deterministic test fixture.
}
func (application operationStrongAuthApplication) VerifyTOTPEnrollment(_ context.Context, command identity.VerifyTOTPEnrollmentCommand) error {
	application.capture.recordCommand("verifyTOTPEnrollment", command)
	return nil
}
func (application operationStrongAuthApplication) RevokeTOTP(_ context.Context, command identity.RevokeTOTPCommand) error {
	application.capture.recordCommand("revokeTOTP", command)
	return nil
}
func (application operationStrongAuthApplication) RotateRecoveryCodes(_ context.Context, command identity.RotateRecoveryCodesCommand) (identity.RecoveryCodes, error) {
	application.capture.recordCommand("rotateRecoveryCodes", command)
	return identity.RecoveryCodes{Codes: []string{"ABCD-EFGH-IJKL-MNOP-QRST-UVWX-YZ23-4567"}}, nil
}
func (application operationStrongAuthApplication) ConsumeRecoveryCode(_ context.Context, command identity.ConsumeRecoveryCodeCommand) (identity.SessionTokens, error) {
	application.capture.recordCommand("consumeRecoveryCode", command)
	return testAccountTokens(), nil
}

type operationDeviceApplication struct{ capture *operationCapture }

func testDeviceTokens() deviceauth.DeviceTokens {
	return deviceauth.DeviceTokens{DeviceID: uuid.MustParse("1645ba9c-ad1d-4a06-b996-d9feb78ee88a"), AuthorizationID: uuid.MustParse("28ceee8a-5f4f-4d3a-9e9f-d3ec23b815ac"), FamilyID: uuid.MustParse("0ff820a5-5022-48e6-8867-77761f8e2f07"), AccessToken: secret.NewBytes(bytes.Repeat([]byte{0x74}, 32)), RefreshToken: secret.NewBytes(bytes.Repeat([]byte{0x75}, 32)), AccessExpiresAt: time.Now().Add(time.Minute)}
}
func (application operationDeviceApplication) CreateChallenge(_ context.Context, command deviceauth.CreateChallengeCommand) (deviceauth.Challenge, error) {
	application.capture.recordCommand("createDeviceAuthChallenge", command)
	return deviceauth.Challenge{ChallengeID: "d12dca8a-ced3-471d-a6ad-55b228221f10", Challenge: [32]byte{1}, ExpiresAt: time.Now().Add(time.Minute)}, nil
}
func (application operationDeviceApplication) RegisterDevice(_ context.Context, command deviceauth.RegisterDeviceCommand) (deviceauth.DeviceTokens, error) {
	application.capture.recordCommand("registerDevice", command)
	return testDeviceTokens(), nil
}
func (application operationDeviceApplication) RotateDeviceToken(_ context.Context, command deviceauth.RotateDeviceTokenCommand) (deviceauth.DeviceTokens, error) {
	application.capture.recordCommand("rotateDeviceToken", command)
	return testDeviceTokens(), nil
}
func (application operationDeviceApplication) RevokeDevice(_ context.Context, command deviceauth.RevokeDeviceCommand) error {
	application.capture.recordCommand("revokeDevice", command)
	return nil
}
func (application operationDeviceApplication) AuthorizeBundle(context.Context, deviceauth.AuthorizeBundleQuery) (deviceauth.BundleAuthority, error) {
	return deviceauth.BundleAuthority{}, nil
}

type operationTrustApplication struct{ capture *operationCapture }

func (application operationTrustApplication) Issue(context.Context, trust.IssueCommand) (trust.IssuedBundle, error) {
	return trust.IssuedBundle{}, nil
}
func (application operationTrustApplication) Resolve(_ context.Context, command trust.ResolveQuery) (trust.Resolution, error) {
	application.capture.recordCommand("resolveConfigBundle", command)
	return trust.Resolution{Locator: "abcdefghijklmnopqrstuv", EnvelopeSHA256: strings.Repeat("a", 64), Locations: [3]string{"https://one.example/b", "https://two.example/b", "https://three.example/b"}, CacheControl: "public,max-age=31536000,immutable"}, nil
}
func (application operationTrustApplication) Acknowledge(_ context.Context, command trust.AcknowledgeCommand) error {
	application.capture.recordCommand("acknowledgeConfigBundle", command)
	return nil
}

func (capture *trustCapture) Acknowledge(_ context.Context, _ trust.AcknowledgeCommand) error {
	capture.acknowledgeCalls++
	return nil
}

func (capture *trustCapture) Resolve(_ context.Context, query trust.ResolveQuery) (trust.Resolution, error) {
	capture.query = query
	return trust.Resolution{
		Locator: "abcdefghijklmnopqrstuv", EnvelopeSHA256: strings.Repeat("a", 64),
		Locations: [3]string{"https://one.example/b", "https://two.example/b", "https://three.example/b"}, CacheControl: "public,max-age=31536000,immutable",
	}, nil
}

func TestC1KeyApplicationMappingsPreserveFrozenContracts(t *testing.T) {
	principal := uuid.MustParse("28ceee8a-5f4f-4d3a-9e9f-d3ec23b815ac")
	session := uuid.MustParse("d12dca8a-ced3-471d-a6ad-55b228221f10")
	authorization := uuid.MustParse("1645ba9c-ad1d-4a06-b996-d9feb78ee88a")
	accountAuth := &accountAuthenticatorStub{authority: identity.AccountAuthority{
		PrincipalID: identity.PrincipalID(principal.String()), SessionID: identity.SessionID(session.String()),
	}}
	deviceAuth := &deviceAuthenticatorStub{authority: deviceauth.BundleAuthority{AuthorizationID: authorization, PrincipalID: principal}}
	identityCapture := &identityApplicationCapture{}
	strongCapture := &strongAuthCapture{}
	trustApplication := &trustCapture{}
	handler := NewHandler(readiness.New(time.Millisecond), Applications{
		Identity: identityCapture, StrongAuth: strongCapture, AccountAuth: accountAuth, DeviceAuth: deviceAuth, Trust: trustApplication,
	}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x28}, 64)))
	token := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x51}, 32))

	passkeyRequest := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/passkey-registration-options", strings.NewReader(`{"reauthentication":{"method":"password","password":"correct horse battery staple"}}`))
	passkeyRequest.Header.Set("Authorization", "Bearer "+token)
	passkeyRecorder := httptest.NewRecorder()
	handler.CreatePasskeyRegistrationOptions(passkeyRecorder, passkeyRequest, controlapiv1.CreatePasskeyRegistrationOptionsParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
	if passkeyRecorder.Code != http.StatusOK || strongCapture.registration.DisplayName != "account" || strongCapture.registration.PrincipalID != identity.PrincipalID(principal.String()) {
		t.Fatalf("passkey mapping = status %d command %+v", passkeyRecorder.Code, strongCapture.registration)
	}
	authenticationRequest := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/passkey-authentication-options", strings.NewReader(`{}`))
	authenticationRecorder := httptest.NewRecorder()
	handler.CreatePasskeyAuthenticationOptions(authenticationRecorder, authenticationRequest, controlapiv1.CreatePasskeyAuthenticationOptionsParams{IdempotencyKey: "bcdefghijklmnopqrstuvw"})
	if authenticationRecorder.Code != http.StatusOK || strongCapture.authentication.IdempotencyKey != "bcdefghijklmnopqrstuvw" {
		t.Fatalf("passkey authentication mapping = status %d command %+v", authenticationRecorder.Code, strongCapture.authentication)
	}

	revokeRequest := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/account-session-revocations", strings.NewReader(`{"scope":"all","reauthentication":{"method":"password","password":"correct horse battery staple"}}`))
	revokeRequest.Header.Set("Authorization", "Bearer "+token)
	revokeRecorder := httptest.NewRecorder()
	handler.RevokeAccountSessions(revokeRecorder, revokeRequest, controlapiv1.RevokeAccountSessionsParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
	if revokeRecorder.Code != http.StatusNoContent || identityCapture.revoke.Reauthentication.SessionID != identity.SessionID(session.String()) ||
		identityCapture.revoke.Reauthentication.Method != identity.ReauthPassword || identityCapture.revoke.Scope != identity.RevokeAllSessions {
		t.Fatalf("session revoke mapping = status %d command %+v", revokeRecorder.Code, identityCapture.revoke)
	}

	resolveRequest := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/config-bundle-resolutions", strings.NewReader(`{}`))
	resolveRequest.Header.Set("Authorization", "Bearer "+token)
	resolveRecorder := httptest.NewRecorder()
	handler.ResolveConfigBundle(resolveRecorder, resolveRequest, controlapiv1.ResolveConfigBundleParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
	if resolveRecorder.Code != http.StatusOK || trustApplication.query.AuthorizationID != authorization {
		t.Fatalf("resolve mapping = status %d query %+v body=%s", resolveRecorder.Code, trustApplication.query, resolveRecorder.Body.String())
	}
	var resolved map[string]any
	if err := json.Unmarshal(resolveRecorder.Body.Bytes(), &resolved); err != nil {
		t.Fatal(err)
	}
	if len(resolved) != 5 || resolved["status"] != "available" || len(resolved["locations"].([]any)) != 3 {
		t.Fatalf("resolution shape = %#v", resolved)
	}
}

func TestRevokeSessionScopeMappingMatchesHTTPContract(t *testing.T) {
	principal := uuid.MustParse("28ceee8a-5f4f-4d3a-9e9f-d3ec23b815ac")
	authenticatedSession := uuid.MustParse("d12dca8a-ced3-471d-a6ad-55b228221f10")
	targetSession := uuid.MustParse("1645ba9c-ad1d-4a06-b996-d9feb78ee88a")
	capture := &identityApplicationCapture{}
	handler := NewHandler(readiness.New(time.Millisecond), Applications{
		Identity: capture, AccountAuth: &accountAuthenticatorStub{authority: identity.AccountAuthority{
			PrincipalID: identity.PrincipalID(principal.String()), SessionID: identity.SessionID(authenticatedSession.String()),
		}},
	}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x2b}, 64)))
	tests := []struct {
		name        string
		body        string
		wantScope   identity.SessionRevokeScope
		wantSession identity.SessionID
	}{
		{name: "all", body: `{"scope":"all","reauthentication":{"method":"password","password":"correct horse battery staple"}}`, wantScope: identity.RevokeAllSessions},
		{name: "others", body: `{"scope":"others","reauthentication":{"method":"password","password":"correct horse battery staple"}}`, wantScope: identity.RevokeOtherSessions, wantSession: identity.SessionID(authenticatedSession.String())},
		{name: "one", body: `{"scope":"one","session_id":"` + targetSession.String() + `","reauthentication":{"method":"password","password":"correct horse battery staple"}}`, wantScope: identity.RevokeCurrentSession, wantSession: identity.SessionID(targetSession.String())},
	}
	token := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x53}, 32))
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/account-session-revocations", strings.NewReader(test.body))
			request.Header.Set("Authorization", "Bearer "+token)
			recorder := httptest.NewRecorder()
			handler.RevokeAccountSessions(recorder, request, controlapiv1.RevokeAccountSessionsParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
			if recorder.Code != http.StatusNoContent || capture.revoke.Scope != test.wantScope || capture.revoke.SessionID != test.wantSession {
				t.Fatalf("mapping = status %d command %+v", recorder.Code, capture.revoke)
			}
		})
	}
}

func TestCreatedOperationsReturnExact201Statuses(t *testing.T) {
	principal := uuid.MustParse("28ceee8a-5f4f-4d3a-9e9f-d3ec23b815ac")
	session := uuid.MustParse("d12dca8a-ced3-471d-a6ad-55b228221f10")
	identityCapture := &identityApplicationCapture{}
	strongCapture := &strongAuthCapture{}
	deviceCapture := &deviceApplicationCapture{}
	handler := NewHandler(readiness.New(time.Millisecond), Applications{
		Identity: identityCapture, StrongAuth: strongCapture, Device: deviceCapture,
		AccountAuth: &accountAuthenticatorStub{authority: identity.AccountAuthority{
			PrincipalID: identity.PrincipalID(principal.String()), SessionID: identity.SessionID(session.String()),
		}},
	}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x2c}, 512)))
	token := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x54}, 32))
	opaque := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x55}, 32))
	encoded32 := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x56}, 32))
	encoded64 := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x57}, 64))
	reauth := `{"reauthentication":{"method":"password","password":"correct horse battery staple"}}`
	tests := []struct {
		name string
		body string
		call func(http.ResponseWriter, *http.Request)
	}{
		{name: "account challenge", body: `{"refresh_token":"` + opaque + `","request_nonce":"` + encoded32 + `"}`, call: func(w http.ResponseWriter, r *http.Request) {
			//nolint:contextcheck // The request already carries the subtest context; the adapter derives a bounded child.
			handler.CreateAccountAuthChallenge(w, r, controlapiv1.CreateAccountAuthChallengeParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
		}},
		{name: "totp enrollment", body: reauth, call: func(w http.ResponseWriter, r *http.Request) {
			//nolint:contextcheck // The request already carries the subtest context; the adapter derives a bounded child.
			handler.CreateTOTPEnrollment(w, r, controlapiv1.CreateTOTPEnrollmentParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
		}},
		{name: "recovery rotation", body: reauth, call: func(w http.ResponseWriter, r *http.Request) {
			//nolint:contextcheck // The request already carries the subtest context; the adapter derives a bounded child.
			handler.RotateRecoveryCodes(w, r, controlapiv1.RotateRecoveryCodesParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
		}},
		{name: "enrollment grant", body: reauth, call: func(w http.ResponseWriter, r *http.Request) {
			//nolint:contextcheck // The request already carries the subtest context; the adapter derives a bounded child.
			handler.CreateDeviceEnrollmentGrant(w, r, controlapiv1.CreateDeviceEnrollmentGrantParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
		}},
		{name: "device challenge", body: `{"enrollment_grant":"` + opaque + `","request_nonce":"` + encoded32 + `","signing_public_key":"` + encoded32 + `","hpke_public_key":"` + encoded32 + `"}`, call: func(w http.ResponseWriter, r *http.Request) {
			//nolint:contextcheck // The request already carries the subtest context; the adapter derives a bounded child.
			handler.CreateDeviceAuthChallenge(w, r, controlapiv1.CreateDeviceAuthChallengeParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
		}},
		{name: "device registration", body: `{"challenge_id":"d12dca8a-ced3-471d-a6ad-55b228221f10","display_name":"device","enrollment_grant":"` + opaque + `","request_nonce":"` + encoded32 + `","signing_public_key":"` + encoded32 + `","hpke_public_key":"` + encoded32 + `","signature":"` + encoded64 + `"}`, call: func(w http.ResponseWriter, r *http.Request) {
			//nolint:contextcheck // The request already carries the subtest context; the adapter derives a bounded child.
			handler.RegisterDevice(w, r, controlapiv1.RegisterDeviceParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(test.body))
			request.Header.Set("Authorization", "Bearer "+token)
			recorder := httptest.NewRecorder()
			test.call(recorder, request)
			if recorder.Code != http.StatusCreated {
				t.Fatalf("status = %d, want 201; body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestTransportScalarValidatorsRejectBeforeApplication(t *testing.T) {
	principal := uuid.MustParse("28ceee8a-5f4f-4d3a-9e9f-d3ec23b815ac")
	session := uuid.MustParse("d12dca8a-ced3-471d-a6ad-55b228221f10")
	target := uuid.MustParse("1645ba9c-ad1d-4a06-b996-d9feb78ee88a")
	identityCapture := &identityApplicationCapture{}
	strongCapture := &strongAuthCapture{}
	handler := NewHandler(readiness.New(time.Millisecond), Applications{
		Identity: identityCapture, StrongAuth: strongCapture,
		AccountAuth: &accountAuthenticatorStub{authority: identity.AccountAuthority{
			PrincipalID: identity.PrincipalID(principal.String()), SessionID: identity.SessionID(session.String()),
		}},
	}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x2d}, 256)))
	token := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x58}, 32))

	for _, test := range []struct {
		name string
		body string
	}{
		{name: "email", body: `{"email":"a@","password":"correct horse battery staple","locale":"en"}`},
		{name: "locale", body: `{"email":"person@example.test","password":"correct horse battery staple","locale":"e"}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			before := identityCapture.registerCalls
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/accounts", strings.NewReader(test.body))
			recorder := httptest.NewRecorder()
			handler.CreateAccount(recorder, request, controlapiv1.CreateAccountParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
			if recorder.Code != http.StatusBadRequest || identityCapture.registerCalls != before {
				t.Fatalf("response = %d calls=%d; body=%s", recorder.Code, identityCapture.registerCalls-before, recorder.Body.String())
			}
		})
	}

	totpRequest := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/totp-verifications", strings.NewReader(`{"code":"12ab56","reauthentication":{"method":"password","password":"correct horse battery staple"}}`))
	totpRequest.Header.Set("Authorization", "Bearer "+token)
	totpRecorder := httptest.NewRecorder()
	handler.VerifyTOTPEnrollment(totpRecorder, totpRequest, controlapiv1.VerifyTOTPEnrollmentParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
	if totpRecorder.Code != http.StatusBadRequest || strongCapture.verifyTOTPCalls != 0 {
		t.Fatalf("TOTP response = %d calls=%d; body=%s", totpRecorder.Code, strongCapture.verifyTOTPCalls, totpRecorder.Body.String())
	}

	recoveryRequest := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/device-enrollment-grants", strings.NewReader(`{"reauthentication":{"method":"recovery_code","code":"not-a-recovery-code"}}`))
	recoveryRequest.Header.Set("Authorization", "Bearer "+token)
	recoveryRecorder := httptest.NewRecorder()
	handler.CreateDeviceEnrollmentGrant(recoveryRecorder, recoveryRequest, controlapiv1.CreateDeviceEnrollmentGrantParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
	if recoveryRecorder.Code != http.StatusBadRequest {
		t.Fatalf("recovery response = %d; body=%s", recoveryRecorder.Code, recoveryRecorder.Body.String())
	}

	canonicalRequest := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/account-session-revocations", strings.NewReader(`{"scope":"one","session_id":"`+strings.ToUpper(target.String())+`","reauthentication":{"method":"password","password":"correct horse battery staple"}}`))
	canonicalRequest.Header.Set("Authorization", "Bearer "+token)
	canonicalRecorder := httptest.NewRecorder()
	before := identityCapture.revokeCalls
	handler.RevokeAccountSessions(canonicalRecorder, canonicalRequest, controlapiv1.RevokeAccountSessionsParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
	if canonicalRecorder.Code != http.StatusBadRequest || identityCapture.revokeCalls != before {
		t.Fatalf("canonical UUID response = %d command=%+v; body=%s", canonicalRecorder.Code, identityCapture.revoke, canonicalRecorder.Body.String())
	}
}

func TestRequiredAndCanonicalUUIDsRejectBeforeApplication(t *testing.T) {
	encoded32 := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x61}, 32))
	encoded64 := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x62}, 64))
	principal := uuid.MustParse("28ceee8a-5f4f-4d3a-9e9f-d3ec23b815ac")
	session := uuid.MustParse("d12dca8a-ced3-471d-a6ad-55b228221f10")
	token := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x63}, 32))
	accountAuth := &accountAuthenticatorStub{authority: identity.AccountAuthority{PrincipalID: identity.PrincipalID(principal.String()), SessionID: identity.SessionID(session.String())}}
	deviceAuth := &deviceAuthenticatorStub{authority: deviceauth.BundleAuthority{AuthorizationID: principal, PrincipalID: principal}}

	for _, malformedUUID := range []string{"", "00000000-0000-0000-8000-000000000001", "00000000-0000-4000-7000-000000000001"} {
		t.Run("account rotation "+malformedUUID, func(t *testing.T) {
			capture := &identityApplicationCapture{}
			handler := NewHandler(readiness.New(time.Millisecond), Applications{Identity: capture}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x64}, 64)))
			body := `{"refresh_token":"` + encoded32 + `","request_nonce":"` + encoded32 + `","signature":"` + encoded64 + `"}`
			if malformedUUID != "" {
				body = `{"challenge_id":"` + malformedUUID + `","refresh_token":"` + encoded32 + `","request_nonce":"` + encoded32 + `","signature":"` + encoded64 + `"}`
			}
			recorder := httptest.NewRecorder()
			handler.RotateAccountToken(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(body)), controlapiv1.RotateAccountTokenParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
			if recorder.Code != http.StatusBadRequest || capture.rotateCalls != 0 {
				t.Fatalf("status=%d calls=%d body=%s", recorder.Code, capture.rotateCalls, recorder.Body.String())
			}
		})
	}

	deviceCases := []struct {
		name string
		call func(*Handler, http.ResponseWriter, *http.Request)
		body string
	}{
		{name: "registration", body: `{"display_name":"device","enrollment_grant":"` + encoded32 + `","request_nonce":"` + encoded32 + `","signing_public_key":"` + encoded32 + `","hpke_public_key":"` + encoded32 + `","signature":"` + encoded64 + `"}`, call: func(handler *Handler, w http.ResponseWriter, request *http.Request) {
			handler.RegisterDevice(w, request, controlapiv1.RegisterDeviceParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
		}},
		{name: "rotation", body: `{"refresh_token":"` + encoded32 + `","request_nonce":"` + encoded32 + `","signature":"` + encoded64 + `"}`, call: func(handler *Handler, w http.ResponseWriter, request *http.Request) {
			handler.RotateDeviceToken(w, request, controlapiv1.RotateDeviceTokenParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
		}},
	}
	for _, test := range deviceCases {
		t.Run(test.name, func(t *testing.T) {
			capture := &deviceApplicationCapture{}
			handler := NewHandler(readiness.New(time.Millisecond), Applications{Device: capture}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x65}, 64)))
			recorder := httptest.NewRecorder()
			test.call(handler, recorder, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(test.body)))
			if recorder.Code != http.StatusBadRequest || capture.registerCalls+capture.rotateCalls != 0 {
				t.Fatalf("status=%d calls=%d body=%s", recorder.Code, capture.registerCalls+capture.rotateCalls, recorder.Body.String())
			}
		})
	}

	t.Run("acknowledgement", func(t *testing.T) {
		capture := &trustCapture{}
		handler := NewHandler(readiness.New(time.Millisecond), Applications{Trust: capture, DeviceAuth: deviceAuth}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x66}, 64)))
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"bundle_version":"1"}`))
		request.Header.Set("Authorization", "Bearer "+token)
		recorder := httptest.NewRecorder()
		handler.AcknowledgeConfigBundle(recorder, request, controlapiv1.AcknowledgeConfigBundleParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
		if recorder.Code != http.StatusBadRequest || capture.acknowledgeCalls != 0 {
			t.Fatalf("status=%d calls=%d body=%s", recorder.Code, capture.acknowledgeCalls, recorder.Body.String())
		}
	})
	_ = accountAuth
}

func TestWebAuthnScalarsRejectBeforeApplication(t *testing.T) {
	principal := uuid.MustParse("28ceee8a-5f4f-4d3a-9e9f-d3ec23b815ac")
	session := uuid.MustParse("d12dca8a-ced3-471d-a6ad-55b228221f10")
	token := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x67}, 32))
	encoded32 := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x68}, 32))
	invalidResponse := `{"id":"!","rawId":"!","type":"not-public-key","response":{"clientDataJSON":"!","authenticatorData":"!","signature":"!"}}`
	authenticator := &accountAuthenticatorStub{authority: identity.AccountAuthority{PrincipalID: identity.PrincipalID(principal.String()), SessionID: identity.SessionID(session.String())}}

	t.Run("registration", func(t *testing.T) {
		capture := &strongAuthCapture{}
		handler := NewHandler(readiness.New(time.Millisecond), Applications{StrongAuth: capture, AccountAuth: authenticator}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x69}, 64)))
		body := `{"ceremony_id":"1645ba9c-ad1d-4a06-b996-d9feb78ee88a","response":{"id":"!","rawId":"!","type":"not-public-key","response":{"clientDataJSON":"!","attestationObject":"!"}},"reauthentication":{"method":"password","password":"correct horse battery staple"}}`
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer "+token)
		recorder := httptest.NewRecorder()
		handler.CreatePasskeyCredential(recorder, request, controlapiv1.CreatePasskeyCredentialParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
		if recorder.Code != http.StatusBadRequest || capture.finishRegistrationCalls != 0 {
			t.Fatalf("status=%d calls=%d body=%s", recorder.Code, capture.finishRegistrationCalls, recorder.Body.String())
		}
	})

	t.Run("account session", func(t *testing.T) {
		capture := &strongAuthCapture{}
		handler := NewHandler(readiness.New(time.Millisecond), Applications{StrongAuth: capture}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x6a}, 64)))
		body := `{"method":"passkey","ceremony_id":"1645ba9c-ad1d-4a06-b996-d9feb78ee88a","response":` + invalidResponse + `,"client_signing_public_key":"` + encoded32 + `"}`
		recorder := httptest.NewRecorder()
		handler.CreateAccountSession(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(body)), controlapiv1.CreateAccountSessionParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
		if recorder.Code != http.StatusBadRequest || capture.finishAuthenticationCalls != 0 {
			t.Fatalf("status=%d calls=%d body=%s", recorder.Code, capture.finishAuthenticationCalls, recorder.Body.String())
		}
	})

	t.Run("reauthentication", func(t *testing.T) {
		capture := &strongAuthCapture{}
		handler := NewHandler(readiness.New(time.Millisecond), Applications{StrongAuth: capture, AccountAuth: authenticator}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x6b}, 64)))
		body := `{"reauthentication":{"method":"passkey","ceremony_id":"1645ba9c-ad1d-4a06-b996-d9feb78ee88a","response":` + invalidResponse + `}}`
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer "+token)
		recorder := httptest.NewRecorder()
		handler.CreateTOTPEnrollment(recorder, request, controlapiv1.CreateTOTPEnrollmentParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
		if recorder.Code != http.StatusBadRequest || capture.beginTOTPCalls != 0 {
			t.Fatalf("status=%d calls=%d body=%s", recorder.Code, capture.beginTOTPCalls, recorder.Body.String())
		}
	})
}

func TestValidAccountSessionWithUnavailableTargetFails503(t *testing.T) {
	handler := NewHandler(readiness.New(time.Millisecond), Applications{}, time.Second, bytes.NewReader(bytes.Repeat([]byte{0x2e}, 64)))
	encoded32 := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x59}, 32))
	credentialID := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x5a}, 16))
	tests := []struct {
		name string
		body string
	}{
		{name: "password", body: `{"method":"password","email":"person@example.test","password":"correct horse battery staple","client_signing_public_key":"` + encoded32 + `"}`},
		{name: "passkey", body: `{"method":"passkey","ceremony_id":"d12dca8a-ced3-471d-a6ad-55b228221f10","response":{"id":"` + credentialID + `","rawId":"` + credentialID + `","response":{"authenticatorData":"AQ","clientDataJSON":"AQ","signature":"AQ"},"type":"public-key"},"client_signing_public_key":"` + encoded32 + `"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/account-sessions", strings.NewReader(test.body))
			recorder := httptest.NewRecorder()
			handler.CreateAccountSession(recorder, request, controlapiv1.CreateAccountSessionParams{IdempotencyKey: "abcdefghijklmnopqrstuv"})
			if recorder.Code != http.StatusServiceUnavailable || !strings.Contains(recorder.Body.String(), "dependency_unavailable") {
				t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestImmutableDelegationUsesConfiguredDeadlineAndCancels(t *testing.T) {
	release := make(chan struct{})
	contextResult := make(chan error, 1)
	immutable := http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		select {
		case <-request.Context().Done():
			contextResult <- request.Context().Err()
		case <-release:
			contextResult <- nil
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	handler := NewHandler(readiness.New(time.Millisecond), Applications{ImmutableBundle: immutable}, 5*time.Millisecond, bytes.NewReader(bytes.Repeat([]byte{0x2f}, 64)))
	done := make(chan struct{})
	go func() {
		recorder := httptest.NewRecorder()
		handler.GetImmutableBundle(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/b/task17-locator", nil), "task17-locator")
		close(done)
	}()
	select {
	case <-done:
		if err := <-contextResult; err == nil {
			t.Fatal("immutable request returned without deadline cancellation")
		}
	case <-time.After(100 * time.Millisecond):
		close(release)
		<-done
		t.Fatal("immutable request did not receive configured deadline")
	}
}
