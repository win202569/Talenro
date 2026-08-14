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

func TestLoadAppliesTask18RuntimePolicyDefaults(t *testing.T) {
	got, err := Load(lookup(map[string]string{
		"TALENRO_DATABASE_URL": "database-fixture",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if got.RedisDownAfterFailures != 3 || got.RedisRecoverAfterSuccesses != 2 {
		t.Fatalf("redis policy defaults = down %d recover %d, want 3 and 2", got.RedisDownAfterFailures, got.RedisRecoverAfterSuccesses)
	}
	if got.OutboxDegradedBacklog != 1000 || got.OutboxDownBacklog != 10000 {
		t.Fatalf("outbox backlog defaults = degraded %d down %d, want 1000 and 10000", got.OutboxDegradedBacklog, got.OutboxDownBacklog)
	}
	if got.OutboxDegradedAge != time.Minute || got.OutboxDownAge != 5*time.Minute {
		t.Fatalf("outbox age defaults = degraded %s down %s, want 1m and 5m", got.OutboxDegradedAge, got.OutboxDownAge)
	}
	if got.ErrorReportQueue != 100 || got.ErrorReportBatch != 20 {
		t.Fatalf("reporter defaults = queue %d batch %d, want 100 and 20", got.ErrorReportQueue, got.ErrorReportBatch)
	}
}

func TestLoadAcceptsInclusiveTask18RuntimePolicyBounds(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]string
		want   [8]int64
	}{
		{
			name: "minimums",
			values: map[string]string{
				"TALENRO_REDIS_DOWN_AFTER_FAILURES":     "1",
				"TALENRO_REDIS_RECOVER_AFTER_SUCCESSES": "1",
				"TALENRO_OUTBOX_DEGRADED_BACKLOG":       "100",
				"TALENRO_OUTBOX_DOWN_BACKLOG":           "101",
				"TALENRO_OUTBOX_DEGRADED_AGE":           "10s",
				"TALENRO_OUTBOX_DOWN_AGE":               "11s",
				"TALENRO_ERROR_REPORT_QUEUE":            "10",
				"TALENRO_ERROR_REPORT_BATCH":            "10",
			},
			want: [8]int64{1, 1, 100, 101, int64(10 * time.Second), int64(11 * time.Second), 10, 10},
		},
		{
			name: "maximums",
			values: map[string]string{
				"TALENRO_REDIS_DOWN_AFTER_FAILURES":     "10",
				"TALENRO_REDIS_RECOVER_AFTER_SUCCESSES": "10",
				"TALENRO_OUTBOX_DEGRADED_BACKLOG":       "10000",
				"TALENRO_OUTBOX_DOWN_BACKLOG":           "100000",
				"TALENRO_OUTBOX_DEGRADED_AGE":           "10m",
				"TALENRO_OUTBOX_DOWN_AGE":               "1h",
				"TALENRO_ERROR_REPORT_QUEUE":            "1000",
				"TALENRO_ERROR_REPORT_BATCH":            "100",
			},
			want: [8]int64{10, 10, 10000, 100000, int64(10 * time.Minute), int64(time.Hour), 1000, 100},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			values := maps.Clone(test.values)
			values["TALENRO_DATABASE_URL"] = "database-fixture"
			got, err := Load(lookup(values))
			if err != nil {
				t.Fatal(err)
			}
			actual := [8]int64{
				int64(got.RedisDownAfterFailures), int64(got.RedisRecoverAfterSuccesses),
				got.OutboxDegradedBacklog, got.OutboxDownBacklog,
				int64(got.OutboxDegradedAge), int64(got.OutboxDownAge),
				int64(got.ErrorReportQueue), int64(got.ErrorReportBatch),
			}
			if actual != test.want {
				t.Fatalf("runtime policy = %v, want %v", actual, test.want)
			}
		})
	}
}

func TestLoadRejectsInvalidTask18RuntimePolicy(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]string
		key    string
	}{
		{name: "redis down syntax", values: map[string]string{"TALENRO_REDIS_DOWN_AFTER_FAILURES": "three"}, key: "TALENRO_REDIS_DOWN_AFTER_FAILURES"},
		{name: "redis down below", values: map[string]string{"TALENRO_REDIS_DOWN_AFTER_FAILURES": "0"}, key: "TALENRO_REDIS_DOWN_AFTER_FAILURES"},
		{name: "redis recovery above", values: map[string]string{"TALENRO_REDIS_RECOVER_AFTER_SUCCESSES": "11"}, key: "TALENRO_REDIS_RECOVER_AFTER_SUCCESSES"},
		{name: "degraded backlog below", values: map[string]string{"TALENRO_OUTBOX_DEGRADED_BACKLOG": "99"}, key: "TALENRO_OUTBOX_DEGRADED_BACKLOG"},
		{name: "degraded backlog above", values: map[string]string{"TALENRO_OUTBOX_DEGRADED_BACKLOG": "10001"}, key: "TALENRO_OUTBOX_DEGRADED_BACKLOG"},
		{name: "down backlog equal", values: map[string]string{"TALENRO_OUTBOX_DEGRADED_BACKLOG": "1000", "TALENRO_OUTBOX_DOWN_BACKLOG": "1000"}, key: "TALENRO_OUTBOX_DOWN_BACKLOG"},
		{name: "down backlog above", values: map[string]string{"TALENRO_OUTBOX_DOWN_BACKLOG": "100001"}, key: "TALENRO_OUTBOX_DOWN_BACKLOG"},
		{name: "degraded age below", values: map[string]string{"TALENRO_OUTBOX_DEGRADED_AGE": "9s"}, key: "TALENRO_OUTBOX_DEGRADED_AGE"},
		{name: "degraded age above", values: map[string]string{"TALENRO_OUTBOX_DEGRADED_AGE": "10m1s"}, key: "TALENRO_OUTBOX_DEGRADED_AGE"},
		{name: "down age less than one second higher", values: map[string]string{"TALENRO_OUTBOX_DEGRADED_AGE": "60s", "TALENRO_OUTBOX_DOWN_AGE": "60.999s"}, key: "TALENRO_OUTBOX_DOWN_AGE"},
		{name: "down age above", values: map[string]string{"TALENRO_OUTBOX_DOWN_AGE": "1h1s"}, key: "TALENRO_OUTBOX_DOWN_AGE"},
		{name: "report queue below", values: map[string]string{"TALENRO_ERROR_REPORT_QUEUE": "9"}, key: "TALENRO_ERROR_REPORT_QUEUE"},
		{name: "report queue above", values: map[string]string{"TALENRO_ERROR_REPORT_QUEUE": "1001"}, key: "TALENRO_ERROR_REPORT_QUEUE"},
		{name: "report batch syntax", values: map[string]string{"TALENRO_ERROR_REPORT_BATCH": "many"}, key: "TALENRO_ERROR_REPORT_BATCH"},
		{name: "report batch above", values: map[string]string{"TALENRO_ERROR_REPORT_BATCH": "101"}, key: "TALENRO_ERROR_REPORT_BATCH"},
		{name: "report batch exceeds queue", values: map[string]string{"TALENRO_ERROR_REPORT_QUEUE": "10", "TALENRO_ERROR_REPORT_BATCH": "11"}, key: "TALENRO_ERROR_REPORT_BATCH"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			values := maps.Clone(test.values)
			values["TALENRO_DATABASE_URL"] = "database-fixture"
			_, err := Load(lookup(values))
			if err == nil || !strings.Contains(err.Error(), test.key) {
				t.Fatalf("expected sanitized %s rejection, got %v", test.key, err)
			}
			for _, private := range test.values {
				if len(private) > 3 && strings.Contains(err.Error(), private) {
					t.Fatalf("invalid value leaked in error: %v", err)
				}
			}
		})
	}
}
