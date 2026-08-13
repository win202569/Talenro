package controlapi

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
	"talenro.local/platform/internal/readiness"
)

type testProbe struct {
	name string
	err  error
}

func (p testProbe) Name() string               { return p.name }
func (p testProbe) Ping(context.Context) error { return p.err }

func TestHealthRoutes(t *testing.T) {
	checker := readiness.New(50*time.Millisecond,
		testProbe{name: "postgres"},
		testProbe{name: "redis", err: errors.New("redis://user:secret@example")},
	)
	handler := NewHandler(checker, Applications{}, 5*time.Second, rand.Reader)
	router := controlapiv1.HandlerWithOptions(handler, controlapiv1.StdHTTPServerOptions{
		BaseRouter: http.NewServeMux(), ErrorHandlerFunc: handler.GeneratedParameterError,
	})

	tests := []struct {
		path   string
		status int
	}{
		{path: "/livez", status: http.StatusOK},
		{path: "/readyz", status: http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.path, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != tt.status {
			t.Fatalf("%s got %d want %d", tt.path, rec.Code, tt.status)
		}
		body, err := io.ReadAll(rec.Result().Body)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(body), "secret") || strings.Contains(string(body), "example") {
			t.Fatalf("sensitive dependency error leaked: %s", body)
		}
	}
}

func TestNilReadinessCheckerFailsClosed(t *testing.T) {
	handler := NewHandler(nil, Applications{}, time.Second, rand.Reader)
	recorder := httptest.NewRecorder()
	handler.GetReadiness(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/readyz", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503; body=%s", recorder.Code, recorder.Body.String())
	}
}
