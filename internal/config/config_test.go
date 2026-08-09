package config

import (
	"maps"
	"strings"
	"testing"
	"time"
)

func lookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	_, err := Load(lookup(map[string]string{}))
	if err == nil || !strings.Contains(err.Error(), "TALENRO_DATABASE_URL") {
		t.Fatalf("expected database error, got %v", err)
	}
}

func TestLoadAppliesSafeDefaults(t *testing.T) {
	got, err := Load(lookup(map[string]string{
		"TALENRO_DATABASE_URL": "database-fixture",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if got.HTTPAddress != "127.0.0.1:8080" || got.RedisAddress != "127.0.0.1:6379" || got.NATSURL != "nats://127.0.0.1:4222" {
		t.Fatalf("unexpected defaults: %+v", got)
	}
	if got.MetricsAddress != "127.0.0.1:9090" {
		t.Fatalf("unexpected metrics address: %s", got.MetricsAddress)
	}
	if got.AllowPublicMetrics {
		t.Fatal("public metrics must be disabled by default")
	}
	if got.DependencyTimeout != 2*time.Second || got.ShutdownTimeout != 10*time.Second {
		t.Fatalf("unexpected timeouts: %+v", got)
	}
}

func TestLoadRejectsPublicHTTPWithoutExplicitOptIn(t *testing.T) {
	_, err := Load(lookup(map[string]string{
		"TALENRO_DATABASE_URL": "postgres://local",
		"TALENRO_HTTP_ADDRESS": "0.0.0.0:8080",
	}))
	if err == nil {
		t.Fatal("expected public bind rejection")
	}
}

func TestLoadReadsMetricsAddress(t *testing.T) {
	got, err := Load(lookup(map[string]string{
		"TALENRO_DATABASE_URL":    "postgres://local",
		"TALENRO_METRICS_ADDRESS": "127.0.0.1:19090",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if got.MetricsAddress != "127.0.0.1:19090" {
		t.Fatalf("metrics address = %q, want %q", got.MetricsAddress, "127.0.0.1:19090")
	}
}

func TestLoadRejectsUnsafeMetricsAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
	}{
		{name: "public", address: "0.0.0.0:9090"},
		{name: "malformed", address: "127.0.0.1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Load(lookup(map[string]string{
				"TALENRO_DATABASE_URL":    "postgres://local",
				"TALENRO_METRICS_ADDRESS": test.address,
			}))
			if err == nil || !strings.Contains(err.Error(), "METRICS") {
				t.Fatalf("expected metrics bind rejection, got %v", err)
			}
		})
	}
}

func TestLoadAcceptsLoopbackMetricsAddresses(t *testing.T) {
	for _, address := range []string{
		"127.0.0.1:9090",
		"[::1]:9090",
		"LOCALHOST:9090",
	} {
		t.Run(address, func(t *testing.T) {
			got, err := Load(lookup(map[string]string{
				"TALENRO_DATABASE_URL":    "postgres://local",
				"TALENRO_METRICS_ADDRESS": address,
			}))
			if err != nil {
				t.Fatalf("loopback metrics address rejected: %v", err)
			}
			if got.MetricsAddress != address {
				t.Fatalf("metrics address = %q, want %q", got.MetricsAddress, address)
			}
		})
	}
}

func TestLoadRejectsNonLoopbackMetricsAddressesByDefault(t *testing.T) {
	for _, address := range []string{
		"192.0.2.10:9090",
		"metrics.internal:9090",
		"0.0.0.0:9090",
		"[::]:9090",
	} {
		t.Run(address, func(t *testing.T) {
			_, err := Load(lookup(map[string]string{
				"TALENRO_DATABASE_URL":    "postgres://local",
				"TALENRO_METRICS_ADDRESS": address,
			}))
			if err == nil || !strings.Contains(err.Error(), "TALENRO_ALLOW_PUBLIC_METRICS") {
				t.Fatalf("expected private metrics bind rejection, got %v", err)
			}
			if strings.Contains(err.Error(), address) {
				t.Fatalf("metrics address leaked in error: %v", err)
			}
		})
	}
}

func TestLoadRequiresIndependentPublicMetricsOptIn(t *testing.T) {
	_, err := Load(lookup(map[string]string{
		"TALENRO_DATABASE_URL":      "postgres://local",
		"TALENRO_METRICS_ADDRESS":   "192.0.2.10:9090",
		"TALENRO_ALLOW_PUBLIC_HTTP": "true",
	}))
	if err == nil {
		t.Fatal("public HTTP opt-in unexpectedly authorized public metrics")
	}

	for _, address := range []string{
		"192.0.2.10:9090",
		"metrics.internal:9090",
		"0.0.0.0:9090",
		"[::]:9090",
	} {
		got, err := Load(lookup(map[string]string{
			"TALENRO_DATABASE_URL":         "postgres://local",
			"TALENRO_METRICS_ADDRESS":      address,
			"TALENRO_ALLOW_PUBLIC_METRICS": "true",
		}))
		if err != nil {
			t.Fatalf("independent public metrics opt-in rejected %q: %v", address, err)
		}
		if !got.AllowPublicMetrics {
			t.Fatal("public metrics opt-in was not retained")
		}
	}
}

func TestLoadRejectsUnsafeProductionSecurityProviders(t *testing.T) {
	base := map[string]string{
		"TALENRO_DATABASE_URL":             "database-fixture",
		"TALENRO_PROFILE":                  "production",
		"TALENRO_PUBLIC_BASE_URL":          "https://api.example.invalid",
		"TALENRO_PRIMARY_BUNDLE_BASE_URL":  "https://api.example.invalid",
		"TALENRO_MIRROR_A_BASE_URL":        "https://mirror-a.example.invalid",
		"TALENRO_MIRROR_B_BASE_URL":        "https://mirror-b.example.invalid",
		"TALENRO_WEBAUTHN_RP_ID":           "example.invalid",
		"TALENRO_WEBAUTHN_ORIGINS":         "https://app.example.invalid",
		"TALENRO_EMAIL_VERIFICATION_MODE":  "required",
		"TALENRO_SIGNER_PROVIDER":          "external",
		"TALENRO_FIELD_PROTECTOR_PROVIDER": "external",
		"TALENRO_EMAIL_PROVIDER":           "external",
		"TALENRO_ERROR_REPORTER_PROVIDER":  "external",
	}

	tests := []struct{ key, value string }{
		{key: "TALENRO_SIGNER_PROVIDER", value: "local"},
		{key: "TALENRO_FIELD_PROTECTOR_PROVIDER", value: "local"},
		{key: "TALENRO_EMAIL_PROVIDER", value: "local"},
		{key: "TALENRO_PUBLIC_BASE_URL", value: "http://example.invalid"},
		{key: "TALENRO_PUBLIC_BASE_URL", value: "https://api.example.invalid?"},
		{key: "TALENRO_EMAIL_VERIFICATION_MODE", value: "disabled"},
	}
	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			values := maps.Clone(base)
			values[test.key] = test.value
			_, err := Load(lookup(values))
			if err == nil || strings.Contains(err.Error(), test.value) {
				t.Fatalf("expected sanitized rejection, got %v", err)
			}
		})
	}
}

func TestLoadBuildsValidatedLocalSecurityProfile(t *testing.T) {
	got, err := Load(lookup(map[string]string{
		"TALENRO_DATABASE_URL":            "database-fixture",
		"TALENRO_PRIMARY_BUNDLE_BASE_URL": "http://localhost:8080/",
		"TALENRO_MIRROR_A_BASE_URL":       "http://localhost:8081/",
		"TALENRO_MIRROR_B_BASE_URL":       "http://localhost:8082/",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if got.Security.Profile != ProfileLocal || got.Security.EmailVerification != EmailDisabled {
		t.Fatal("local security defaults were not applied")
	}
	if got.Security.BundleBaseURLs != [3]string{
		"http://localhost:8080", "http://localhost:8081", "http://localhost:8082",
	} {
		t.Fatal("bundle base URLs were not normalized")
	}
	if got.Security.RequestDeadline != 5*time.Second || got.Security.RedisTimeout != 250*time.Millisecond ||
		got.Security.SignerTimeout != 2*time.Second || got.Security.ErrorReportTimeout != time.Second ||
		got.Security.ClockSkew != 120*time.Second {
		t.Fatal("security timeout defaults were not applied")
	}
	if got.Security.LoginRateLimit != (RateLimitPolicy{Limit: 10, Window: 15 * time.Minute}) ||
		got.Security.DeliveryRateLimit != (RateLimitPolicy{Limit: 5, Window: time.Hour}) ||
		got.Security.ChallengeRateLimit != (RateLimitPolicy{Limit: 20, Window: 5 * time.Minute}) {
		t.Fatal("security rate limits were not applied")
	}
	if len(got.Security.SensitiveLookupKey.Copy()) != 32 || len(got.Security.LocalRootSigningSeed.Copy()) != 32 {
		t.Fatal("local security keys were not decoded")
	}
}

func TestLoadRejectsErrorReporterTimeoutOverTwoSeconds(t *testing.T) {
	_, err := Load(lookup(map[string]string{
		"TALENRO_DATABASE_URL":         "database-fixture",
		"TALENRO_ERROR_REPORT_TIMEOUT": "3s",
	}))
	if err == nil || strings.Contains(err.Error(), "3s") {
		t.Fatalf("expected sanitized timeout rejection, got %v", err)
	}
}
