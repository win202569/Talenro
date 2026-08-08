// Package config loads process configuration from an environment lookup function.
package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddress        string
	MetricsAddress     string
	AllowPublicMetrics bool
	DatabaseURL        string
	RedisAddress       string
	NATSURL            string
	DependencyTimeout  time.Duration
	ShutdownTimeout    time.Duration
}

type Lookup func(string) (string, bool)

func Load(lookup Lookup) (Config, error) {
	cfg := Config{
		HTTPAddress:       "127.0.0.1:8080",
		MetricsAddress:    "127.0.0.1:9090",
		RedisAddress:      "127.0.0.1:6379",
		NATSURL:           "nats://127.0.0.1:4222",
		DependencyTimeout: 2 * time.Second,
		ShutdownTimeout:   10 * time.Second,
	}

	var ok bool
	if cfg.DatabaseURL, ok = lookup("TALENRO_DATABASE_URL"); !ok || cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("TALENRO_DATABASE_URL is required")
	}
	if value, exists := lookup("TALENRO_HTTP_ADDRESS"); exists && value != "" {
		cfg.HTTPAddress = value
	}
	if value, exists := lookup("TALENRO_METRICS_ADDRESS"); exists && value != "" {
		cfg.MetricsAddress = value
	}
	if value, exists := lookup("TALENRO_REDIS_ADDRESS"); exists && value != "" {
		cfg.RedisAddress = value
	}
	if value, exists := lookup("TALENRO_NATS_URL"); exists && value != "" {
		cfg.NATSURL = value
	}

	allowPublic := false
	if value, exists := lookup("TALENRO_ALLOW_PUBLIC_HTTP"); exists {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return Config{}, fmt.Errorf("parse TALENRO_ALLOW_PUBLIC_HTTP: %w", err)
		}
		allowPublic = parsed
	}
	if value, exists := lookup("TALENRO_ALLOW_PUBLIC_METRICS"); exists {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return Config{}, fmt.Errorf("parse TALENRO_ALLOW_PUBLIC_METRICS: invalid boolean")
		}
		cfg.AllowPublicMetrics = parsed
	}

	host, _, err := net.SplitHostPort(cfg.HTTPAddress)
	if err != nil {
		return Config{}, fmt.Errorf("parse TALENRO_HTTP_ADDRESS: %w", err)
	}
	ip := net.ParseIP(host)
	publicBind := host == "" || (ip != nil && ip.IsUnspecified())
	if publicBind && !allowPublic {
		return Config{}, fmt.Errorf("public HTTP bind requires TALENRO_ALLOW_PUBLIC_HTTP=true")
	}

	metricsHost, _, err := net.SplitHostPort(cfg.MetricsAddress)
	if err != nil {
		return Config{}, fmt.Errorf("parse TALENRO_METRICS_ADDRESS: invalid host:port")
	}
	if !cfg.AllowPublicMetrics && !isLoopbackHost(metricsHost) {
		return Config{}, fmt.Errorf("non-loopback metrics bind requires TALENRO_ALLOW_PUBLIC_METRICS=true")
	}

	return cfg, nil
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
