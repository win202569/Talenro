package controlapi

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"talenro.local/platform/internal/apierrors"
)

func TestErrorFiniteStatusTableAndPrivacy(t *testing.T) {
	tests := []struct {
		err    error
		status int
		code   string
	}{
		{err: apierrors.New(apierrors.MalformedRequest, apierrors.ContactSupport), status: http.StatusBadRequest, code: "malformed_request"},
		{err: apierrors.New(apierrors.AuthenticationFailed, apierrors.Reauthenticate), status: http.StatusUnauthorized, code: "authentication_failed"},
		{err: apierrors.New(apierrors.ActionNotAllowed, apierrors.ContactSupport), status: http.StatusForbidden, code: "action_not_allowed"},
		{err: apierrors.New(apierrors.IdempotencyConflict, apierrors.ContactSupport), status: http.StatusConflict, code: "idempotency_conflict"},
		{err: apierrors.New(apierrors.RequestTooLarge, apierrors.ContactSupport), status: http.StatusRequestEntityTooLarge, code: "request_too_large"},
		{err: apierrors.New(apierrors.UnsupportedSchema, apierrors.UpgradeClient), status: http.StatusUnsupportedMediaType, code: "unsupported_schema"},
		{err: apierrors.New(apierrors.RateLimited, apierrors.Retry), status: http.StatusTooManyRequests, code: "rate_limited"},
		{err: errors.New("CANARY raw infrastructure error"), status: http.StatusServiceUnavailable, code: "dependency_unavailable"},
	}
	for _, test := range tests {
		recorder := httptest.NewRecorder()
		handler := &Handler{random: bytes.NewReader(bytes.Repeat([]byte{0x41}, 16))}
		handler.writeError(recorder, test.err)
		if recorder.Code != test.status || !strings.Contains(recorder.Body.String(), test.code) {
			t.Fatalf("error response = %d %s, want %d/%s", recorder.Code, recorder.Body.String(), test.status, test.code)
		}
		if strings.Contains(recorder.Body.String(), "CANARY") || strings.Contains(recorder.Body.String(), "infrastructure") {
			t.Fatalf("raw error leaked: %s", recorder.Body.String())
		}
		if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
			t.Fatalf("Cache-Control = %q", got)
		}
	}
}

func TestErrorRandomFailureUsesFixedFallbackTrace(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler := &Handler{random: errorReader{}}
	handler.writeError(recorder, apierrors.New(apierrors.MalformedRequest, apierrors.ContactSupport))
	if !strings.Contains(recorder.Body.String(), `"trace_id":"trace-unavailable"`) {
		t.Fatalf("fallback trace response = %s", recorder.Body.String())
	}
}

func TestAuthenticationFailureEmitsExactBearerChallenge(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler := &Handler{random: bytes.NewReader(bytes.Repeat([]byte{0x42}, 16))}
	handler.writeError(recorder, apierrors.New(apierrors.AuthenticationFailed, apierrors.Reauthenticate))
	if recorder.Code != http.StatusUnauthorized || recorder.Header().Get("WWW-Authenticate") != "Bearer" {
		t.Fatalf("authentication response = %d challenge %q", recorder.Code, recorder.Header().Get("WWW-Authenticate"))
	}

	recorder = httptest.NewRecorder()
	handler.writeError(recorder, apierrors.New(apierrors.MalformedRequest, apierrors.ContactSupport))
	if challenge := recorder.Header().Get("WWW-Authenticate"); challenge != "" {
		t.Fatalf("non-authentication challenge = %q", challenge)
	}
}

func TestErrorTypedNilRandomUsesFixedFallbackTrace(t *testing.T) {
	var random *panicReader
	recorder := httptest.NewRecorder()
	handler := &Handler{random: random}
	handler.writeError(recorder, apierrors.New(apierrors.MalformedRequest, apierrors.ContactSupport))
	if !strings.Contains(recorder.Body.String(), `"trace_id":"trace-unavailable"`) {
		t.Fatalf("typed-nil fallback trace response = %s", recorder.Body.String())
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("CANARY random failure") }

type panicReader struct{}

func (*panicReader) Read([]byte) (int, error) { panic("typed nil random was invoked") }
