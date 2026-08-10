package apierrors_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
	"talenro.local/platform/internal/apierrors"
)

func TestPublicErrorContainsOnlyFiniteFields(t *testing.T) {
	t.Parallel()

	err := apierrors.New(apierrors.AuthenticationFailed, apierrors.Reauthenticate)
	body, marshalErr := json.Marshal(err.Public("trace-safe-000001"))
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}
	for _, forbidden := range []string{"SECRET-CANARY", "cause", "message", "stack"} {
		if bytes.Contains(body, []byte(forbidden)) {
			t.Fatalf("disclosed %q", forbidden)
		}
	}
	const want = `{"action":"reauthenticate","code":"authentication_failed","trace_id":"trace-safe-000001"}`
	if string(body) != want {
		t.Fatalf("Public() JSON = %s, want %s", body, want)
	}
}

func TestCodeAndActionMatchOpenAPIEnumsExactly(t *testing.T) {
	t.Parallel()

	codes := []apierrors.Code{
		apierrors.MalformedRequest,
		apierrors.AuthenticationFailed,
		apierrors.ActionNotAllowed,
		apierrors.IdempotencyConflict,
		apierrors.StateConflict,
		apierrors.VersionRollback,
		apierrors.RequestTooLarge,
		apierrors.UnsupportedSchema,
		apierrors.RateLimited,
		apierrors.DependencyUnavailable,
		apierrors.SigningUnavailable,
	}
	wantCodes := []controlapiv1.PublicErrorCode{
		controlapiv1.PublicErrorCodeMalformedRequest,
		controlapiv1.PublicErrorCodeAuthenticationFailed,
		controlapiv1.PublicErrorCodeActionNotAllowed,
		controlapiv1.PublicErrorCodeIdempotencyConflict,
		controlapiv1.PublicErrorCodeStateConflict,
		controlapiv1.PublicErrorCodeVersionRollback,
		controlapiv1.PublicErrorCodeRequestTooLarge,
		controlapiv1.PublicErrorCodeUnsupportedSchema,
		controlapiv1.PublicErrorCodeRateLimited,
		controlapiv1.PublicErrorCodeDependencyUnavailable,
		controlapiv1.PublicErrorCodeSigningUnavailable,
	}
	if len(codes) != len(wantCodes) {
		t.Fatalf("code count = %d, want %d", len(codes), len(wantCodes))
	}
	for index := range codes {
		if controlapiv1.PublicErrorCode(codes[index]) != wantCodes[index] {
			t.Fatalf("code[%d] = %q, want %q", index, codes[index], wantCodes[index])
		}
	}

	actions := []apierrors.Action{
		apierrors.Retry,
		apierrors.Reauthenticate,
		apierrors.Reenroll,
		apierrors.UpgradeClient,
		apierrors.ContactSupport,
	}
	wantActions := []controlapiv1.PublicErrorAction{
		controlapiv1.Retry,
		controlapiv1.Reauthenticate,
		controlapiv1.Reenroll,
		controlapiv1.UpgradeClient,
		controlapiv1.ContactSupport,
	}
	if len(actions) != len(wantActions) {
		t.Fatalf("action count = %d, want %d", len(actions), len(wantActions))
	}
	for index := range actions {
		if controlapiv1.PublicErrorAction(actions[index]) != wantActions[index] {
			t.Fatalf("action[%d] = %q, want %q", index, actions[index], wantActions[index])
		}
	}
}

func TestUnknownEnumsAndUnsafeTraceFailClosed(t *testing.T) {
	t.Parallel()

	err := apierrors.New(apierrors.Code("SECRET-CANARY"), apierrors.Action("bad-action"))
	public := err.Public("../../SECRET-CANARY")
	if public.Code != controlapiv1.PublicErrorCodeDependencyUnavailable || public.Action != controlapiv1.ContactSupport {
		t.Fatalf("unknown enum projection = %#v", public)
	}
	if public.TraceId != "trace-unavailable" {
		t.Fatalf("unsafe trace projection = %q", public.TraceId)
	}
	if got := err.Error(); got != "apierrors: classified failure" {
		t.Fatalf("Error() = %q", got)
	}
	body, marshalErr := json.Marshal(public)
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}
	if bytes.Contains(body, []byte("SECRET-CANARY")) {
		t.Fatalf("fail-closed error disclosed canary: %s", body)
	}
}

func TestUnsafeTraceIsReplacedWithoutChangingValidClassification(t *testing.T) {
	t.Parallel()

	public := apierrors.New(apierrors.AuthenticationFailed, apierrors.Reauthenticate).Public("SECRET/CANARY")
	if public.Code != controlapiv1.PublicErrorCodeAuthenticationFailed || public.Action != controlapiv1.Reauthenticate {
		t.Fatalf("unsafe trace changed valid classification: %#v", public)
	}
	if public.TraceId != "trace-unavailable" {
		t.Fatalf("unsafe trace projection = %q", public.TraceId)
	}
}

func TestRetryAfterIsBoundedAndOptional(t *testing.T) {
	t.Parallel()

	retryable := apierrors.NewRetryAfter(apierrors.RateLimited, apierrors.Retry, 1500*time.Millisecond)
	public := retryable.Public("trace-safe-000002")
	if public.RetryAfterMs == nil || *public.RetryAfterMs != 1500 {
		t.Fatalf("retry_after_ms = %v, want 1500", public.RetryAfterMs)
	}

	for _, invalid := range []time.Duration{0, time.Microsecond, 10*time.Second + time.Millisecond} {
		invalidError := apierrors.NewRetryAfter(apierrors.RateLimited, apierrors.Retry, invalid).Public("trace-safe-000003")
		if invalidError.Code != controlapiv1.PublicErrorCodeDependencyUnavailable || invalidError.Action != controlapiv1.ContactSupport || invalidError.RetryAfterMs != nil {
			t.Fatalf("invalid retry %s did not fail closed: %#v", invalid, invalidError)
		}
	}

	wrongAction := apierrors.NewRetryAfter(apierrors.RateLimited, apierrors.ContactSupport, time.Second).Public("trace-safe-000004")
	if wrongAction.Code != controlapiv1.PublicErrorCodeDependencyUnavailable || wrongAction.Action != controlapiv1.ContactSupport || wrongAction.RetryAfterMs != nil {
		t.Fatalf("retry delay with non-retry action did not fail closed: %#v", wrongAction)
	}
}

func TestErrorCannotBeMarshaledDirectly(t *testing.T) {
	t.Parallel()

	body, err := json.Marshal(apierrors.New(apierrors.AuthenticationFailed, apierrors.Reauthenticate))
	if err == nil || body != nil {
		t.Fatalf("json.Marshal(Error) = %q, %v; want blocked", body, err)
	}
	if bytes.Contains([]byte(err.Error()), []byte("authentication_failed")) {
		t.Fatalf("marshal error disclosed classification: %q", err)
	}
}
