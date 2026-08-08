// Package controlapi implements the generated control API server contract.
package controlapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
)

func TestEmbeddedContractHasHealthRoutes(t *testing.T) {
	t.Parallel()

	// Task 7 requires exercising the generated compatibility entry point.
	spec, err := controlapiv1.GetSwagger() //nolint:staticcheck // Deliberately verify the required legacy API.
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]struct {
		operationID string
		statuses    []string
	}{
		"/livez":  {operationID: "getLiveness", statuses: []string{"200"}},
		"/readyz": {operationID: "getReadiness", statuses: []string{"200", "503"}},
	}
	for path, test := range tests {
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

type contractServer struct{}

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
