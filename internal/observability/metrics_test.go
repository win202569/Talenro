package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiddlewareUsesOnlyBoundedRouteLabel(t *testing.T) {
	registry := NewRegistry()
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := registry.Middleware("/readyz", next)
	for _, target := range []string{
		"/readyz?email=user@example.invalid",
		"/random-device-123?target=198.51.100.1",
	} {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, target, nil))
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
		httptest.NewRequest(http.MethodGet, "/device/private-node-123?email=user@example.invalid", nil),
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
