package config

import (
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
		"TALENRO_DATABASE_URL": "postgres://talenro:talenro_dev@localhost:5432/talenro?sslmode=disable",
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
