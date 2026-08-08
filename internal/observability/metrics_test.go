package observability

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMiddlewareUsesOnlyBoundedRouteLabel(t *testing.T) {
	registry := NewRegistry()
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := registry.Middleware("/readyz", next)
	for _, target := range []string{
		"/readyz?email=user@example.invalid",
		"/random-device-123?target=198.51.100.1",
	} {
		handler.ServeHTTP(httptest.NewRecorder(), newServerRequest(t, target))
	}
	families, err := registry.Gatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range families {
		encoded := family.String()
		for _, forbidden := range []string{"example.invalid", "198.51.100.1", "random-device-123"} {
			if strings.Contains(encoded, forbidden) {
				t.Fatalf("unbounded label leaked: %s", forbidden)
			}
		}
		if strings.Contains(family.GetName(), "http_requests") &&
			(!strings.Contains(encoded, `name:"route"`) || !strings.Contains(encoded, `value:"/readyz"`)) {
			t.Fatalf("expected normalized route label: %s", encoded)
		}
	}
}

func TestMiddlewareCollapsesUnknownRouteToUnmatched(t *testing.T) {
	registry := NewRegistry()
	handler := registry.Middleware("/device/private-node-123", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	handler.ServeHTTP(
		httptest.NewRecorder(),
		newServerRequest(t, "/device/private-node-123?email=user@example.invalid"),
	)

	families, err := registry.Gatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	foundRequests := false
	for _, family := range families {
		if family.GetName() != "talenro_control_http_requests_total" {
			continue
		}
		foundRequests = true
		encoded := family.String()
		labels := make(map[string]string)
		for _, label := range family.GetMetric()[0].GetLabel() {
			labels[label.GetName()] = label.GetValue()
		}
		wantLabels := map[string]string{
			"method":       http.MethodGet,
			"route":        unmatchedRoute,
			"status_class": "4xx",
		}
		if len(labels) != len(wantLabels) {
			t.Fatalf("request labels = %v, want exactly %v", labels, wantLabels)
		}
		for name, want := range wantLabels {
			if got := labels[name]; got != want {
				t.Fatalf("label %s = %q, want %q", name, got, want)
			}
		}
		for _, forbidden := range []string{"private-node-123", "example.invalid"} {
			if strings.Contains(encoded, forbidden) {
				t.Fatalf("private route data leaked: %s", forbidden)
			}
		}
	}
	if !foundRequests {
		t.Fatal("request counter metric family was not gathered")
	}
}

func TestMiddlewareRecordsFirstActualStatus(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantStatus int
	}{
		{
			name: "implicit 200 ignores later 500",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte("ok"))
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "explicit 204 ignores later 500",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantStatus: http.StatusNoContent,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry := NewRegistry()
			recorder := httptest.NewRecorder()
			registry.Middleware("/livez", test.handler).ServeHTTP(
				recorder,
				newServerRequest(t, "/livez"),
			)
			if recorder.Code != test.wantStatus {
				t.Fatalf("response status = %d, want %d", recorder.Code, test.wantStatus)
			}
			if got := gatheredRequestLabel(t, registry, "status_class"); got != "2xx" {
				t.Fatalf("status class = %q, want %q", got, "2xx")
			}
		})
	}
}

func TestMiddlewareRecordsFlushCommittedStatus(t *testing.T) {
	registry := NewRegistry()
	handler := registry.Middleware("/livez", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		flusher.Flush()
		w.WriteHeader(http.StatusInternalServerError)
	}))
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	response, err := server.Client().Do(newClientRequest(t, server.URL+"/livez"))
	if err != nil {
		t.Fatal(err)
	}
	_, copyErr := io.Copy(io.Discard, response.Body)
	closeErr := response.Body.Close()
	if copyErr != nil || closeErr != nil {
		t.Fatalf("consume response: copy=%v close=%v", copyErr, closeErr)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("client status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if got := gatheredRequestLabel(t, registry, "status_class"); got != "2xx" {
		t.Fatalf("status class = %q, want %q", got, "2xx")
	}
}

func TestMiddlewareRecordsFlushErrorCommittedStatus(t *testing.T) {
	registry := NewRegistry()
	flushErrCh := make(chan error, 1)
	handler := registry.Middleware("/livez", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		flusher, ok := w.(interface{ FlushError() error })
		if !ok {
			flushErrCh <- errors.New("wrapped server writer does not implement FlushError")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		flushErrCh <- flusher.FlushError()
		w.WriteHeader(http.StatusInternalServerError)
	}))
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	response, err := server.Client().Do(newClientRequest(t, server.URL+"/livez"))
	if err != nil {
		t.Fatal(err)
	}
	_, copyErr := io.Copy(io.Discard, response.Body)
	closeErr := response.Body.Close()
	if copyErr != nil || closeErr != nil {
		t.Fatalf("consume response: copy=%v close=%v", copyErr, closeErr)
	}
	if flushErr := <-flushErrCh; flushErr != nil {
		t.Fatal(flushErr)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("client status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if got := gatheredRequestLabel(t, registry, "status_class"); got != "2xx" {
		t.Fatalf("status class = %q, want %q", got, "2xx")
	}
}

func TestMiddlewarePreservesOnlyUnderlyingResponseWriterCapabilities(t *testing.T) {
	t.Run("capable writer", func(t *testing.T) {
		registry := NewRegistry()
		underlying := newCapableResponseWriter()
		handler := registry.Middleware("/livez", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			flusher, ok := w.(http.Flusher)
			if !ok {
				t.Error("wrapped writer does not implement http.Flusher")
				return
			}
			flusher.Flush()

			hijacker, ok := w.(http.Hijacker)
			if !ok {
				t.Error("wrapped writer does not implement http.Hijacker")
				return
			}
			_, _, _ = hijacker.Hijack()

			pusher, ok := w.(http.Pusher)
			if !ok {
				t.Error("wrapped writer does not implement http.Pusher")
				return
			}
			_ = pusher.Push("/asset", nil)
		}))
		handler.ServeHTTP(underlying, newServerRequest(t, "/livez"))
		if underlying.flushCalls != 1 || underlying.hijackCalls != 1 || underlying.pushCalls != 1 {
			t.Fatalf(
				"capability calls = flush:%d hijack:%d push:%d, want each once",
				underlying.flushCalls,
				underlying.hijackCalls,
				underlying.pushCalls,
			)
		}
	})

	t.Run("minimal writer", func(t *testing.T) {
		registry := NewRegistry()
		underlying := newMinimalResponseWriter()
		handler := registry.Middleware("/livez", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if _, ok := w.(http.Flusher); ok {
				t.Error("wrapped writer falsely implements http.Flusher")
			}
			if _, ok := w.(http.Hijacker); ok {
				t.Error("wrapped writer falsely implements http.Hijacker")
			}
			if _, ok := w.(http.Pusher); ok {
				t.Error("wrapped writer falsely implements http.Pusher")
			}
			unwrapper, ok := w.(interface{ Unwrap() http.ResponseWriter })
			if !ok || unwrapper.Unwrap() != underlying {
				t.Error("wrapped writer does not unwrap to the underlying writer")
			}
		}))
		handler.ServeHTTP(underlying, newServerRequest(t, "/livez"))
	})
}

func TestMiddlewareAllowsResponseControllerTraversal(t *testing.T) {
	registry := NewRegistry()
	underlying := &writeDeadlineResponseWriter{minimalResponseWriter: *newMinimalResponseWriter()}
	wantDeadline := time.Unix(123, 456)
	handler := registry.Middleware("/livez", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := http.NewResponseController(w).SetWriteDeadline(wantDeadline); err != nil {
			t.Errorf("set write deadline through wrapped writer: %v", err)
		}
	}))
	handler.ServeHTTP(underlying, newServerRequest(t, "/livez"))
	if !underlying.writeDeadline.Equal(wantDeadline) {
		t.Fatalf("write deadline = %s, want %s", underlying.writeDeadline, wantDeadline)
	}
}

func gatheredRequestLabel(t *testing.T, registry *Registry, name string) string {
	t.Helper()
	families, err := registry.Gatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range families {
		if family.GetName() != "talenro_control_http_requests_total" {
			continue
		}
		for _, label := range family.GetMetric()[0].GetLabel() {
			if label.GetName() == name {
				return label.GetValue()
			}
		}
	}
	t.Fatalf("request label %q was not gathered", name)
	return ""
}

func newServerRequest(t *testing.T, target string) *http.Request {
	t.Helper()
	return httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
}

func newClientRequest(t *testing.T, target string) *http.Request {
	t.Helper()
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

type minimalResponseWriter struct {
	header http.Header
	status int
}

func newMinimalResponseWriter() *minimalResponseWriter {
	return &minimalResponseWriter{header: make(http.Header)}
}

func (w *minimalResponseWriter) Header() http.Header {
	return w.header
}

func (w *minimalResponseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return len(body), nil
}

func (w *minimalResponseWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
}

type capableResponseWriter struct {
	minimalResponseWriter
	flushCalls  int
	hijackCalls int
	pushCalls   int
}

func newCapableResponseWriter() *capableResponseWriter {
	return &capableResponseWriter{minimalResponseWriter: *newMinimalResponseWriter()}
}

func (w *capableResponseWriter) Flush() {
	w.flushCalls++
}

func (w *capableResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijackCalls++
	return nil, nil, nil
}

func (w *capableResponseWriter) Push(string, *http.PushOptions) error {
	w.pushCalls++
	return nil
}

type writeDeadlineResponseWriter struct {
	minimalResponseWriter
	writeDeadline time.Time
}

func (w *writeDeadlineResponseWriter) SetWriteDeadline(deadline time.Time) error {
	w.writeDeadline = deadline
	return nil
}
