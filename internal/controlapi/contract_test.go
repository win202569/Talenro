// Package controlapi implements the generated control API server contract.
package controlapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
)

var c11Operations = map[string]string{
	"POST /v1/accounts":                       "createAccount",
	"POST /v1/email-verification-deliveries":  "createEmailVerificationDelivery",
	"POST /v1/email-verifications":            "verifyEmail",
	"POST /v1/password-reset-deliveries":      "createPasswordResetDelivery",
	"POST /v1/password-resets":                "resetPassword",
	"POST /v1/password-changes":               "changePassword",
	"POST /v1/account-sessions":               "createAccountSession",
	"POST /v1/account-auth-challenges":        "createAccountAuthChallenge",
	"POST /v1/account-token-rotations":        "rotateAccountToken",
	"POST /v1/account-session-revocations":    "revokeAccountSessions",
	"POST /v1/passkey-registration-options":   "createPasskeyRegistrationOptions",
	"POST /v1/passkey-credentials":            "createPasskeyCredential",
	"POST /v1/passkey-authentication-options": "createPasskeyAuthenticationOptions",
	"POST /v1/passkey-revocations":            "revokePasskey",
	"POST /v1/totp-enrollments":               "createTOTPEnrollment",
	"POST /v1/totp-verifications":             "verifyTOTPEnrollment",
	"POST /v1/totp-revocations":               "revokeTOTP",
	"POST /v1/recovery-code-rotations":        "rotateRecoveryCodes",
	"POST /v1/recovery-code-consumptions":     "consumeRecoveryCode",
	"POST /v1/device-enrollment-grants":       "createDeviceEnrollmentGrant",
	"POST /v1/device-auth-challenges":         "createDeviceAuthChallenge",
	"POST /v1/devices":                        "registerDevice",
	"POST /v1/device-token-rotations":         "rotateDeviceToken",
	"POST /v1/device-revocations":             "revokeDevice",
	"POST /v1/config-bundle-resolutions":      "resolveConfigBundle",
	"POST /v1/config-bundle-acknowledgements": "acknowledgeConfigBundle",
	"GET /b/{bundle_locator}":                 "getImmutableBundle",
}

var c11SuccessStatuses = map[string]string{
	"createAccount":                      "202",
	"createEmailVerificationDelivery":    "202",
	"verifyEmail":                        "204",
	"createPasswordResetDelivery":        "202",
	"resetPassword":                      "200",
	"changePassword":                     "200",
	"createAccountSession":               "200",
	"createAccountAuthChallenge":         "201",
	"rotateAccountToken":                 "200",
	"revokeAccountSessions":              "204",
	"createPasskeyRegistrationOptions":   "200",
	"createPasskeyCredential":            "204",
	"createPasskeyAuthenticationOptions": "200",
	"revokePasskey":                      "204",
	"createTOTPEnrollment":               "201",
	"verifyTOTPEnrollment":               "204",
	"revokeTOTP":                         "204",
	"rotateRecoveryCodes":                "201",
	"consumeRecoveryCode":                "200",
	"createDeviceEnrollmentGrant":        "201",
	"createDeviceAuthChallenge":          "201",
	"registerDevice":                     "201",
	"rotateDeviceToken":                  "200",
	"revokeDevice":                       "204",
	"resolveConfigBundle":                "200",
	"acknowledgeConfigBundle":            "204",
	"getImmutableBundle":                 "200",
}

var c11OperationSecurity = map[string]string{
	"changePassword":                   "AccountOpaqueToken",
	"revokeAccountSessions":            "AccountOpaqueToken",
	"createPasskeyRegistrationOptions": "AccountOpaqueToken",
	"createPasskeyCredential":          "AccountOpaqueToken",
	"revokePasskey":                    "AccountOpaqueToken",
	"createTOTPEnrollment":             "AccountOpaqueToken",
	"verifyTOTPEnrollment":             "AccountOpaqueToken",
	"revokeTOTP":                       "AccountOpaqueToken",
	"rotateRecoveryCodes":              "AccountOpaqueToken",
	"createDeviceEnrollmentGrant":      "AccountOpaqueToken",
	"revokeDevice":                     "AccountOpaqueToken",
	"resolveConfigBundle":              "DeviceOpaqueToken",
	"acknowledgeConfigBundle":          "DeviceOpaqueToken",
}

func TestOpenAPIContract(t *testing.T) {
	t.Parallel()

	// Task 7 requires exercising the generated compatibility entry point.
	spec, err := controlapiv1.GetSwagger() //nolint:staticcheck // Deliberately verify the required legacy API.
	if err != nil {
		t.Fatal(err)
	}

	healthOperations := map[string]struct {
		operationID string
		statuses    []string
	}{
		"/livez":  {operationID: "getLiveness", statuses: []string{"200"}},
		"/readyz": {operationID: "getReadiness", statuses: []string{"200", "503"}},
	}
	for path, test := range healthOperations {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			item := spec.Paths.Find(path)
			if item == nil || item.Get == nil {
				t.Fatalf("missing GET %s", path)
			}
			if item.Get.OperationID != test.operationID {
				t.Errorf("GET %s operation ID = %q, want %q", path, item.Get.OperationID, test.operationID)
			}
			if item.Get.Responses.Len() != len(test.statuses) {
				t.Errorf("GET %s response count = %d, want %d", path, item.Get.Responses.Len(), len(test.statuses))
			}
			for _, status := range test.statuses {
				if item.Get.Responses.Value(status) == nil {
					t.Errorf("GET %s missing response status %s", path, status)
				}
			}
		})
	}

	for route, operationID := range c11Operations {
		method, path, ok := strings.Cut(route, " ")
		if !ok {
			t.Fatalf("invalid operation fixture %q", route)
		}
		item := spec.Paths.Find(path)
		if item == nil {
			t.Fatalf("missing C1.1 path %s", path)
		}
		operation := operationForMethod(item, method)
		if operation == nil {
			t.Fatalf("missing C1.1 operation %s", route)
		}
		if operation.OperationID != operationID {
			t.Errorf("%s operation ID = %q, want %q", route, operation.OperationID, operationID)
		}
		assertOperationSecurity(t, route, operation, c11OperationSecurity[operationID])
		if operation.Responses.Value(c11SuccessStatuses[operationID]) == nil {
			t.Errorf("%s missing success response %s", route, c11SuccessStatuses[operationID])
		}
		for _, status := range []string{"400", "401", "403", "409", "413", "415", "429", "503"} {
			if operation.Responses.Value(status) == nil {
				t.Errorf("%s missing stable error response %s", route, status)
			}
		}
		if method == http.MethodPost {
			assertRequiredParameter(t, operation, "header", "Idempotency-Key")
			if operation.RequestBody == nil || operation.RequestBody.Value == nil {
				t.Errorf("%s missing request body", route)
			} else if got := fmt.Sprint(operation.RequestBody.Value.Extensions["x-talenro-max-bytes"]); got != "65536" {
				t.Errorf("%s x-talenro-max-bytes = %q, want 65536", route, got)
			}
		} else {
			assertRequiredParameter(t, operation, "path", "bundle_locator")
		}
	}

	for _, name := range []string{"AccountOpaqueToken", "DeviceOpaqueToken"} {
		if spec.Components == nil || spec.Components.SecuritySchemes[name] == nil || spec.Components.SecuritySchemes[name].Value == nil {
			t.Errorf("missing security scheme %s", name)
			continue
		}
		scheme := spec.Components.SecuritySchemes[name].Value
		if scheme.Type != "http" || scheme.Scheme != "bearer" {
			t.Errorf("security scheme %s = type %q scheme %q, want http bearer", name, scheme.Type, scheme.Scheme)
		}
	}

	bundleOperation := spec.Paths.Find("/b/{bundle_locator}").Get
	bundleResponse := bundleOperation.Responses.Value("200")
	if bundleResponse == nil || bundleResponse.Value == nil || bundleResponse.Value.Content["application/vnd.talenro.bundle+json"] == nil {
		t.Error("immutable bundle response missing application/vnd.talenro.bundle+json media type")
	}

	for name, schemaRef := range spec.Components.Schemas {
		assertObjectsClosed(t, "#/components/schemas/"+name, schemaRef, map[*openapi3.Schema]bool{})
	}

	locale := spec.Components.Schemas["Locale"].Value
	if locale.MinLength != 2 || locale.MaxLength == nil || *locale.MaxLength != 35 || locale.Pattern != `^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$` {
		t.Errorf("Locale bounds = min %d max %v pattern %q, want 2..35 and approved language-tag pattern", locale.MinLength, locale.MaxLength, locale.Pattern)
	}
	password := spec.Components.Schemas["Password"].Value
	if got := fmt.Sprint(password.Extensions["x-talenro-max-utf8-bytes"]); got != "1024" {
		t.Errorf("Password x-talenro-max-utf8-bytes = %q, want 1024", got)
	}
}

func TestWebAuthnRegistrationContractAcceptsGoWebAuthnFixtures(t *testing.T) {
	t.Parallel()

	spec, err := controlapiv1.GetSwagger() //nolint:staticcheck // Exercise the embedded generated contract.
	if err != nil {
		t.Fatal(err)
	}

	options := map[string]any{
		"ceremony_id": "018f8d68-3d4b-7f42-8c6a-4ec9370c7e45",
		"publicKey": map[string]any{
			"rp": map[string]any{
				"id":   "example.com",
				"name": "Talenro",
			},
			"user": map[string]any{
				"id":          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
				"name":        "principal-018f8d68",
				"displayName": "Device owner",
			},
			"challenge": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
			"pubKeyCredParams": []any{
				map[string]any{"type": "public-key", "alg": float64(-7)},
			},
			"timeout": float64(60000),
			"authenticatorSelection": map[string]any{
				"authenticatorAttachment": "platform",
				"requireResidentKey":      true,
				"residentKey":             "required",
				"userVerification":        "required",
			},
			"attestation": "none",
			"excludeCredentials": []any{
				map[string]any{
					"type":       "public-key",
					"id":         "AQIDBAUGBwgJCgsMDQ4PEA",
					"transports": []any{"internal"},
				},
			},
		},
	}
	optionsSchema := spec.Paths.Find("/v1/passkey-registration-options").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value
	if err := optionsSchema.VisitJSON(options); err != nil {
		t.Fatalf("registration options fixture rejected: %v", err)
	}

	credential := map[string]any{
		"ceremony_id": "018f8d68-3d4b-7f42-8c6a-4ec9370c7e45",
		"response": map[string]any{
			"id":    "AQIDBAUGBwgJCgsMDQ4PEA",
			"rawId": "AQIDBAUGBwgJCgsMDQ4PEA",
			"type":  "public-key",
			"response": map[string]any{
				"clientDataJSON":     "AQID",
				"attestationObject":  "BAUG",
				"authenticatorData":  "BwgJ",
				"publicKey":          "CgsM",
				"publicKeyAlgorithm": float64(-7),
				"transports":         []any{"internal"},
			},
		},
		"reauthentication": map[string]any{
			"method":   "password",
			"password": "correct horse battery staple",
		},
	}
	credentialSchema := spec.Paths.Find("/v1/passkey-credentials").Post.RequestBody.Value.Content["application/json"].Schema.Value
	if err := credentialSchema.VisitJSON(credential); err != nil {
		t.Fatalf("registration response fixture rejected: %v", err)
	}
}

func TestWebAuthnAuthenticationContractAcceptsGoWebAuthnFixtures(t *testing.T) {
	t.Parallel()

	spec, err := controlapiv1.GetSwagger() //nolint:staticcheck // Exercise the embedded generated contract.
	if err != nil {
		t.Fatal(err)
	}

	options := map[string]any{
		"ceremony_id": "018f8d68-3d4b-7f42-8c6a-4ec9370c7e45",
		"publicKey": map[string]any{
			"challenge":        "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
			"timeout":          float64(60000),
			"rpId":             "example.com",
			"userVerification": "required",
			"allowCredentials": []any{
				map[string]any{
					"type":       "public-key",
					"id":         "AQIDBAUGBwgJCgsMDQ4PEA",
					"transports": []any{"internal"},
				},
			},
		},
	}
	optionsSchema := spec.Paths.Find("/v1/passkey-authentication-options").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value
	if err := optionsSchema.VisitJSON(options); err != nil {
		t.Fatalf("authentication options fixture rejected: %v", err)
	}

	session := map[string]any{
		"method":                    "passkey",
		"ceremony_id":               "018f8d68-3d4b-7f42-8c6a-4ec9370c7e45",
		"client_signing_public_key": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"response": map[string]any{
			"id":    "AQIDBAUGBwgJCgsMDQ4PEA",
			"rawId": "AQIDBAUGBwgJCgsMDQ4PEA",
			"type":  "public-key",
			"response": map[string]any{
				"clientDataJSON":    "AQID",
				"authenticatorData": "BAUG",
				"signature":         "BwgJ",
				"userHandle":        "CgsM",
			},
		},
	}
	sessionSchema := spec.Paths.Find("/v1/account-sessions").Post.RequestBody.Value.Content["application/json"].Schema.Value
	if err := sessionSchema.VisitJSON(session); err != nil {
		t.Fatalf("authentication response fixture rejected: %v", err)
	}
}

func assertOperationSecurity(t *testing.T, route string, operation *openapi3.Operation, scheme string) {
	t.Helper()
	if operation.Security == nil {
		t.Errorf("%s must declare operation-level security", route)
		return
	}
	requirements := *operation.Security
	if scheme == "" {
		if len(requirements) != 0 {
			t.Errorf("%s security = %#v, want public", route, requirements)
		}
		return
	}
	if len(requirements) != 1 || len(requirements[0]) != 1 {
		t.Errorf("%s security = %#v, want only %s", route, requirements, scheme)
		return
	}
	if _, exists := requirements[0][scheme]; !exists {
		t.Errorf("%s security = %#v, want only %s", route, requirements, scheme)
	}
}

func operationForMethod(item *openapi3.PathItem, method string) *openapi3.Operation {
	switch method {
	case http.MethodGet:
		return item.Get
	case http.MethodPost:
		return item.Post
	default:
		return nil
	}
}

func assertRequiredParameter(t *testing.T, operation *openapi3.Operation, location, name string) {
	t.Helper()
	for _, parameterRef := range operation.Parameters {
		if parameterRef.Value != nil && parameterRef.Value.In == location && parameterRef.Value.Name == name {
			if !parameterRef.Value.Required {
				t.Errorf("%s parameter %s must be required", location, name)
			}
			return
		}
	}
	t.Errorf("missing required %s parameter %s", location, name)
}

func assertObjectsClosed(t *testing.T, location string, schemaRef *openapi3.SchemaRef, seen map[*openapi3.Schema]bool) {
	t.Helper()
	if schemaRef == nil || schemaRef.Value == nil || seen[schemaRef.Value] {
		return
	}
	schema := schemaRef.Value
	seen[schema] = true
	if schema.Type != nil && schema.Type.Is(openapi3.TypeObject) {
		if schema.AdditionalProperties.Schema == nil && (schema.AdditionalProperties.Has == nil || *schema.AdditionalProperties.Has) {
			t.Errorf("%s must set additionalProperties: false", location)
		}
	}
	for name, property := range schema.Properties {
		assertObjectsClosed(t, location+"/properties/"+name, property, seen)
	}
	for index, variant := range schema.OneOf {
		assertObjectsClosed(t, fmt.Sprintf("%s/oneOf/%d", location, index), variant, seen)
	}
	if schema.Items != nil {
		assertObjectsClosed(t, location+"/items", schema.Items, seen)
	}
}

func TestGeneratedHandlerRoutesHealthResponses(t *testing.T) {
	t.Parallel()

	handler := controlapiv1.HandlerFromMux(contractServer{}, http.NewServeMux())
	tests := map[string]struct {
		statusCode int
		response   controlapiv1.HealthResponse
	}{
		"/livez": {
			statusCode: http.StatusOK,
			response: controlapiv1.HealthResponse{
				Status: controlapiv1.Ok,
				Checks: map[string]string{"process": "ok"},
			},
		},
		"/readyz": {
			statusCode: http.StatusServiceUnavailable,
			response: controlapiv1.HealthResponse{
				Status: controlapiv1.Unavailable,
				Checks: map[string]string{"database": "unavailable"},
			},
		},
	}
	for path, test := range tests {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.statusCode {
				t.Fatalf("GET %s status = %d, want %d", path, recorder.Code, test.statusCode)
			}
			if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("GET %s Content-Type = %q, want application/json", path, contentType)
			}
			var response controlapiv1.HealthResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode GET %s response: %v", path, err)
			}
			if response.Status != test.response.Status {
				t.Errorf("GET %s response status = %q, want %q", path, response.Status, test.response.Status)
			}
			if len(response.Checks) != 1 {
				t.Fatalf("GET %s checks = %#v, want one check", path, response.Checks)
			}
			for component, status := range test.response.Checks {
				if response.Checks[component] != status {
					t.Errorf("GET %s check %q = %q, want %q", path, component, response.Checks[component], status)
				}
			}

			var fields map[string]json.RawMessage
			if err := json.Unmarshal(recorder.Body.Bytes(), &fields); err != nil {
				t.Fatalf("decode GET %s response fields: %v", path, err)
			}
			if len(fields) != 2 {
				t.Fatalf("GET %s response fields = %#v, want only status and checks", path, fields)
			}
			for _, field := range []string{"status", "checks"} {
				if _, exists := fields[field]; !exists {
					t.Errorf("GET %s response missing %q field", path, field)
				}
			}
		})
	}
}

type contractServer struct{ C1Unavailable }

var _ controlapiv1.ServerInterface = contractServer{}

func (contractServer) GetLiveness(responseWriter http.ResponseWriter, _ *http.Request) {
	writeHealthResponse(responseWriter, http.StatusOK, controlapiv1.HealthResponse{
		Status: controlapiv1.Ok,
		Checks: map[string]string{"process": "ok"},
	})
}

func (contractServer) GetReadiness(responseWriter http.ResponseWriter, _ *http.Request) {
	writeHealthResponse(responseWriter, http.StatusServiceUnavailable, controlapiv1.HealthResponse{
		Status: controlapiv1.Unavailable,
		Checks: map[string]string{"database": "unavailable"},
	})
}

func writeHealthResponse(responseWriter http.ResponseWriter, statusCode int, response controlapiv1.HealthResponse) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)
	if err := json.NewEncoder(responseWriter).Encode(response); err != nil {
		return
	}
}
