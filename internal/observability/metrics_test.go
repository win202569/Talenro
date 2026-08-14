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

	"talenro.local/platform/internal/errorreport"
)

func TestTask18MetricsExposeOnlyExactFixedLabelSets(t *testing.T) {
	registry := NewRegistry()
	if !registry.RecordSecurity(SecurityOperationAccountAuth, MetricResultSuccess, SecurityReasonNone) {
		t.Fatal("valid security metric was rejected")
	}
	if !registry.SetOutbox(1234, 61*time.Second) {
		t.Fatal("valid outbox health was rejected")
	}
	registry.Record(errorreport.ComponentOutbox, errorreport.ResultSent)
	if !registry.RecordCrypto(CryptoOperationBundleVerify, MetricResultFailure, CryptoReasonInvalid) {
		t.Fatal("valid crypto metric was rejected")
	}

	wantLabels := map[string][]string{
		"talenro_security_events_total":    {"operation", "reason", "result"},
		"talenro_outbox_backlog":           {},
		"talenro_outbox_oldest_seconds":    {},
		"talenro_error_reports_total":      {"component", "result"},
		"talenro_crypto_validations_total": {"operation", "reason", "result"},
	}
	families, err := registry.Gatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range families {
		want, tracked := wantLabels[family.GetName()]
		if !tracked {
			continue
		}
		if len(family.GetMetric()) != 1 {
			t.Fatalf("%s series count = %d, want 1", family.GetName(), len(family.GetMetric()))
		}
		labels := family.GetMetric()[0].GetLabel()
		if len(labels) != len(want) {
			t.Fatalf("%s label count = %d, want %d", family.GetName(), len(labels), len(want))
		}
		for index, label := range labels {
			if label.GetName() != want[index] {
				t.Fatalf("%s label %d = %q, want %q", family.GetName(), index, label.GetName(), want[index])
			}
		}
		delete(wantLabels, family.GetName())
	}
	if len(wantLabels) != 0 {
		t.Fatalf("missing Task 18 metric families: %v", wantLabels)
	}
}

func TestTask18MetricsRejectAttackerControlledLabelsWithoutCreatingSeries(t *testing.T) {
	registry := NewRegistry()
	if !registry.RecordSecurity(SecurityOperationAccountAuth, MetricResultSuccess, SecurityReasonNone) ||
		!registry.RecordCrypto(CryptoOperationBundleVerify, MetricResultFailure, CryptoReasonInvalid) {
		t.Fatal("valid baseline metric was rejected")
	}
	registry.Record(errorreport.ComponentOutbox, errorreport.ResultSent)
	before := task18MetricSeriesCounts(t, registry)

	const canary = "CANARY_user@example.invalid_device-123"
	if registry.RecordSecurity(SecurityOperation(canary), MetricResult(canary), SecurityReason(canary)) {
		t.Fatal("attacker-controlled security labels were accepted")
	}
	if registry.RecordCrypto(CryptoOperation(canary), MetricResult(canary), CryptoReason(canary)) {
		t.Fatal("attacker-controlled crypto labels were accepted")
	}
	registry.Record(errorreport.Component(canary), errorreport.DeliveryResult(canary))
	if registry.SetOutbox(-1, -time.Second) {
		t.Fatal("invalid outbox health was accepted")
	}

	after := task18MetricSeriesCounts(t, registry)
	for name, want := range before {
		if got := after[name]; got != want {
			t.Fatalf("%s series count = %d after invalid labels, want %d", name, got, want)
		}
	}
	families, err := registry.Gatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range families {
		if strings.Contains(family.String(), canary) {
			t.Fatalf("attacker canary reached metric descriptor or labels: %s", family.String())
		}
	}
}

func TestTask18MetricsAllowProductionBundleSigningOperation(t *testing.T) {
	registry := NewRegistry()
	if !registry.RecordCrypto(CryptoOperation("bundle_sign"), MetricResultSuccess, CryptoReasonNone) {
		t.Fatal("production bundle signing operation was rejected")
	}
	series := task18MetricSeriesCounts(t, registry)
	if got := series["talenro_crypto_validations_total"]; got != 1 {
		t.Fatalf("bundle signing series = %d, want 1", got)
	}
}

func TestMiddlewareRecordsSecurityOutcomesFromRealPublicRoutes(t *testing.T) {
	registry := NewRegistry()
	mux := http.NewServeMux()
	tests := []struct {
		pattern   string
		status    int
		operation SecurityOperation
		result    MetricResult
		reason    SecurityReason
	}{
		{"POST /v1/accounts", http.StatusAccepted, SecurityOperationAccountRegister, MetricResultSuccess, SecurityReasonNone},
		{"POST /v1/account-sessions", http.StatusUnauthorized, SecurityOperationAccountAuth, MetricResultFailure, SecurityReasonInvalidCredential},
		{"POST /v1/password-resets", http.StatusTooManyRequests, SecurityOperationAccountRecovery, MetricResultFailure, SecurityReasonRateLimited},
		{"POST /v1/devices", http.StatusConflict, SecurityOperationDeviceEnroll, MetricResultFailure, SecurityReasonReplay},
		{"POST /v1/device-token-rotations", http.StatusServiceUnavailable, SecurityOperationDeviceAuth, MetricResultFailure, SecurityReasonDependency},
		{"POST /v1/device-revocations", http.StatusInternalServerError, SecurityOperationDeviceRevoke, MetricResultFailure, SecurityReasonInternal},
		{"POST /v1/config-bundle-resolutions", http.StatusOK, SecurityOperationBundleResolve, MetricResultSuccess, SecurityReasonNone},
		{"POST /v1/config-bundle-acknowledgements", http.StatusForbidden, SecurityOperationBundleAck, MetricResultFailure, SecurityReasonInvalidCredential},
	}
	for _, test := range tests {
		mux.HandleFunc(test.pattern, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(test.status)
		})
	}
	handler := registry.Middleware("", mux)
	for _, test := range tests {
		parts := strings.SplitN(test.pattern, " ", 2)
		request := httptest.NewRequestWithContext(t.Context(), parts[0], parts[1], nil)
		handler.ServeHTTP(httptest.NewRecorder(), request)
	}

	const canary = "CANARY_user@example.invalid_device-123"
	handler.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/"+canary, nil),
	)

	want := make(map[string]float64, len(tests))
	for _, test := range tests {
		want[task18SecurityMetricKey(test.operation, test.result, test.reason)]++
	}
	got := task18GatheredSecuritySeries(t, registry)
	if len(got) != len(want) {
		t.Fatalf("security series = %v, want exactly %v", got, want)
	}
	for labels, count := range want {
		if got[labels] != count {
			t.Errorf("security series %q = %v, want %v", labels, got[labels], count)
		}
	}
	families, err := registry.Gatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range families {
		if strings.Contains(family.String(), canary) {
			t.Fatalf("attacker-controlled route reached a metric: %s", family.String())
		}
	}
}

func TestMiddlewareMapsExactPublicSecuritySurface(t *testing.T) {
	tests := []struct {
		pattern   string
		target    string
		operation SecurityOperation
	}{
		{"POST /v1/accounts", "/v1/accounts", SecurityOperationAccountRegister},
		{"POST /v1/email-verification-deliveries", "/v1/email-verification-deliveries", SecurityOperationAccountRegister},
		{"POST /v1/email-verifications", "/v1/email-verifications", SecurityOperationAccountRegister},
		{"POST /v1/password-reset-deliveries", "/v1/password-reset-deliveries", SecurityOperationAccountRecovery},
		{"POST /v1/password-resets", "/v1/password-resets", SecurityOperationAccountRecovery},
		{"POST /v1/recovery-code-consumptions", "/v1/recovery-code-consumptions", SecurityOperationAccountRecovery},
		{"POST /v1/password-changes", "/v1/password-changes", SecurityOperationAccountAuth},
		{"POST /v1/account-sessions", "/v1/account-sessions", SecurityOperationAccountAuth},
		{"POST /v1/account-auth-challenges", "/v1/account-auth-challenges", SecurityOperationAccountAuth},
		{"POST /v1/account-token-rotations", "/v1/account-token-rotations", SecurityOperationAccountAuth},
		{"POST /v1/account-session-revocations", "/v1/account-session-revocations", SecurityOperationAccountAuth},
		{"POST /v1/passkey-registration-options", "/v1/passkey-registration-options", SecurityOperationAccountAuth},
		{"POST /v1/passkey-credentials", "/v1/passkey-credentials", SecurityOperationAccountAuth},
		{"POST /v1/passkey-authentication-options", "/v1/passkey-authentication-options", SecurityOperationAccountAuth},
		{"POST /v1/passkey-revocations", "/v1/passkey-revocations", SecurityOperationAccountAuth},
		{"POST /v1/totp-enrollments", "/v1/totp-enrollments", SecurityOperationAccountAuth},
		{"POST /v1/totp-verifications", "/v1/totp-verifications", SecurityOperationAccountAuth},
		{"POST /v1/totp-revocations", "/v1/totp-revocations", SecurityOperationAccountAuth},
		{"POST /v1/recovery-code-rotations", "/v1/recovery-code-rotations", SecurityOperationAccountAuth},
		{"POST /v1/device-enrollment-grants", "/v1/device-enrollment-grants", SecurityOperationDeviceEnroll},
		{"POST /v1/device-auth-challenges", "/v1/device-auth-challenges", SecurityOperationDeviceEnroll},
		{"POST /v1/devices", "/v1/devices", SecurityOperationDeviceEnroll},
		{"POST /v1/device-token-rotations", "/v1/device-token-rotations", SecurityOperationDeviceAuth},
		{"POST /v1/device-revocations", "/v1/device-revocations", SecurityOperationDeviceRevoke},
		{"POST /v1/config-bundle-resolutions", "/v1/config-bundle-resolutions", SecurityOperationBundleResolve},
		{"GET /b/{bundle_locator}", "/b/abcdefghijklmnopqrstuv", SecurityOperationBundleResolve},
		{"POST /v1/config-bundle-acknowledgements", "/v1/config-bundle-acknowledgements", SecurityOperationBundleAck},
	}
	for _, test := range tests {
		t.Run(test.pattern, func(t *testing.T) {
			registry := NewRegistry()
			mux := http.NewServeMux()
			mux.HandleFunc(test.pattern, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})
			method := strings.SplitN(test.pattern, " ", 2)[0]
			request := httptest.NewRequestWithContext(t.Context(), method, test.target, nil)
			registry.Middleware("", mux).ServeHTTP(httptest.NewRecorder(), request)
			want := map[string]float64{
				task18SecurityMetricKey(test.operation, MetricResultSuccess, SecurityReasonNone): 1,
			}
			if got := task18GatheredSecuritySeries(t, registry); len(got) != 1 {
				t.Fatalf("security series = %v, want %v", got, want)
			} else if got[task18SecurityMetricKey(test.operation, MetricResultSuccess, SecurityReasonNone)] != 1 {
				t.Fatalf("security series = %v, want %v", got, want)
			}
		})
	}
}

func task18GatheredSecuritySeries(t *testing.T, registry *Registry) map[string]float64 {
	t.Helper()
	families, err := registry.Gatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	series := make(map[string]float64)
	for _, family := range families {
		if family.GetName() != "talenro_security_events_total" {
			continue
		}
		for _, metric := range family.GetMetric() {
			labels := make(map[string]string)
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			key := strings.Join([]string{labels["operation"], labels["result"], labels["reason"]}, "\x00")
			series[key] = metric.GetCounter().GetValue()
		}
	}
	return series
}

func task18SecurityMetricKey(operation SecurityOperation, result MetricResult, reason SecurityReason) string {
	return strings.Join([]string{string(operation), string(result), string(reason)}, "\x00")
}

func task18MetricSeriesCounts(t *testing.T, registry *Registry) map[string]int {
	t.Helper()
	families, err := registry.Gatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	counts := make(map[string]int)
	for _, family := range families {
		switch family.GetName() {
		case "talenro_security_events_total", "talenro_error_reports_total", "talenro_crypto_validations_total":
			counts[family.GetName()] = len(family.GetMetric())
		}
	}
	return counts
}

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

func TestMiddlewareBoundsMethodLabelsForAllHTTPMetricFamilies(t *testing.T) {
	registry := NewRegistry()
	handler := registry.Middleware("/livez", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	methods := []string{
		"POST",
		"PUT",
		"PATCH",
		"DELETE",
		"OPTIONS",
		"HEAD",
		"CONNECT",
		"TRACE",
		"DEVICE-private-node-123",
		"TENANT-customer-8472",
		"USER-session-token-991",
		"REQUEST-correlation-id-550e8400",
	}
	handler.ServeHTTP(httptest.NewRecorder(), newServerRequest(t, "/livez"))
	for _, method := range methods {
		request := httptest.NewRequestWithContext(t.Context(), method, "/livez", nil)
		handler.ServeHTTP(httptest.NewRecorder(), request)
	}

	families, err := registry.Gatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	wantFamilies := map[string]bool{
		"talenro_control_http_requests_total":           false,
		"talenro_control_http_request_duration_seconds": false,
	}
	for _, family := range families {
		if _, ok := wantFamilies[family.GetName()]; !ok {
			continue
		}
		wantFamilies[family.GetName()] = true
		if got := len(family.GetMetric()); got != 2 {
			t.Fatalf("%s series count = %d, want 2", family.GetName(), got)
		}

		observationsByMethod := make(map[string]float64)
		for _, metric := range family.GetMetric() {
			labels := make(map[string]string)
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
				for _, forbidden := range methods {
					if strings.Contains(label.GetValue(), forbidden) {
						t.Fatalf("%s label %s leaked raw method %q", family.GetName(), label.GetName(), forbidden)
					}
				}
			}
			if labels["route"] != "/livez" || labels["status_class"] != "2xx" {
				t.Fatalf("%s labels = %v, want route /livez and status class 2xx", family.GetName(), labels)
			}
			method := labels["method"]
			if method != http.MethodGet && method != "other" {
				t.Fatalf("%s method label = %q, want GET or other", family.GetName(), method)
			}
			if _, exists := observationsByMethod[method]; exists {
				t.Fatalf("%s has multiple %q method series", family.GetName(), method)
			}
			switch family.GetName() {
			case "talenro_control_http_requests_total":
				observationsByMethod[method] = metric.GetCounter().GetValue()
			case "talenro_control_http_request_duration_seconds":
				observationsByMethod[method] = float64(metric.GetHistogram().GetSampleCount())
			}
		}
		if got := observationsByMethod[http.MethodGet]; got != 1 {
			t.Errorf("%s GET observations = %v, want 1", family.GetName(), got)
		}
		if got := observationsByMethod["other"]; got != 12 {
			t.Errorf("%s other observations = %v, want 12", family.GetName(), got)
		}
	}
	for name, found := range wantFamilies {
		if !found {
			t.Errorf("metric family %s was not gathered", name)
		}
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
