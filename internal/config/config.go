// Package config loads process configuration from an environment lookup function.
package config

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"talenro.local/platform/internal/secret"
)

// Profile determines which runtime safety constraints apply.
type Profile string

// EmailVerificationMode determines the email-verification policy.
type EmailVerificationMode string

// Provider identifies a configured adapter class.
type Provider string

//nolint:revive // These closed string registries are documented by their exported types and self-describing names.
const (
	// ProfileLocal and the following constants are the supported runtime profiles.
	ProfileLocal Profile = "local"
	// ProfileTest selects deterministic integration fixtures.
	ProfileTest Profile = "test"
	// ProfileProduction selects external security adapters and HTTPS-only origins.
	ProfileProduction Profile = "production"

	EmailRequired EmailVerificationMode = "required"
	EmailGrace    EmailVerificationMode = "grace"
	EmailDisabled EmailVerificationMode = "disabled"

	ProviderLocal    Provider = "local"
	ProviderExternal Provider = "external"
	ProviderDiscard  Provider = "discard"
)

const (
	defaultLookupKeyB64         = "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXowMTIzNDU"
	defaultEncryptionKeyB64     = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY"
	defaultRootSigningSeedB64   = "cm9vdC1zaWduaW5nLXNlZWQtZm9yLWxvY2FsLXRlc3Q"
	defaultConfigSigningSeedB64 = "Y29uZmlnLXNpZ25pbmctc2VlZC1sb2NhbC10ZXN0LTE"
)

// RateLimitPolicy fixes a server-side rate limit and its window.
type RateLimitPolicy struct {
	Limit  uint32
	Window time.Duration
}

// SecurityConfig contains configuration that affects identity and trust safety.
type SecurityConfig struct {
	Profile                Profile
	EmailVerification      EmailVerificationMode
	PublicBaseURL          string
	BundleBaseURLs         [3]string
	WebAuthnRPID           string
	WebAuthnOrigins        []string
	SignerProvider         Provider
	FieldProtectorProvider Provider
	EmailProvider          Provider
	ErrorReporterProvider  Provider
	RequestDeadline        time.Duration
	RedisTimeout           time.Duration
	SignerTimeout          time.Duration
	ErrorReportTimeout     time.Duration
	ClockSkew              time.Duration
	LoginRateLimit         RateLimitPolicy
	DeliveryRateLimit      RateLimitPolicy
	ChallengeRateLimit     RateLimitPolicy
	SensitiveLookupKey     secret.Bytes
	SensitiveEncryptionKey secret.Bytes
	LocalRootSigningSeed   secret.Bytes
	LocalConfigSigningSeed secret.Bytes
}

// Config contains validated control API process settings.
type Config struct {
	HTTPAddress                string
	MetricsAddress             string
	AllowPublicMetrics         bool
	DatabaseURL                string
	RedisAddress               string
	NATSURL                    string
	DependencyTimeout          time.Duration
	ShutdownTimeout            time.Duration
	RedisDownAfterFailures     int
	RedisRecoverAfterSuccesses int
	OutboxDegradedBacklog      int64
	OutboxDownBacklog          int64
	OutboxDegradedAge          time.Duration
	OutboxDownAge              time.Duration
	ErrorReportQueue           int
	ErrorReportBatch           int
	Security                   SecurityConfig
}

// Lookup retrieves a configuration value by environment-style key.
type Lookup func(string) (string, bool)

// Load reads and validates control API configuration through lookup.
func Load(lookup Lookup) (Config, error) {
	cfg := Config{
		HTTPAddress:                "127.0.0.1:8080",
		MetricsAddress:             "127.0.0.1:9090",
		RedisAddress:               "127.0.0.1:6379",
		NATSURL:                    "nats://127.0.0.1:4222",
		DependencyTimeout:          2 * time.Second,
		ShutdownTimeout:            10 * time.Second,
		RedisDownAfterFailures:     3,
		RedisRecoverAfterSuccesses: 2,
		OutboxDegradedBacklog:      1000,
		OutboxDownBacklog:          10000,
		OutboxDegradedAge:          time.Minute,
		OutboxDownAge:              5 * time.Minute,
		ErrorReportQueue:           100,
		ErrorReportBatch:           20,
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

	allowPublic, err := parseBool(lookup, "TALENRO_ALLOW_PUBLIC_HTTP", false)
	if err != nil {
		return Config{}, err
	}
	cfg.AllowPublicMetrics, err = parseBool(lookup, "TALENRO_ALLOW_PUBLIC_METRICS", false)
	if err != nil {
		return Config{}, err
	}

	host, _, err := net.SplitHostPort(cfg.HTTPAddress)
	if err != nil {
		return Config{}, fmt.Errorf("parse TALENRO_HTTP_ADDRESS: invalid host:port")
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

	if cfg.RedisDownAfterFailures, err = parseBoundedInt(lookup, "TALENRO_REDIS_DOWN_AFTER_FAILURES", cfg.RedisDownAfterFailures, 1, 10); err != nil {
		return Config{}, err
	}
	if cfg.RedisRecoverAfterSuccesses, err = parseBoundedInt(lookup, "TALENRO_REDIS_RECOVER_AFTER_SUCCESSES", cfg.RedisRecoverAfterSuccesses, 1, 10); err != nil {
		return Config{}, err
	}
	if cfg.OutboxDegradedBacklog, err = parseBoundedInt64(lookup, "TALENRO_OUTBOX_DEGRADED_BACKLOG", cfg.OutboxDegradedBacklog, 100, 10000); err != nil {
		return Config{}, err
	}
	if cfg.OutboxDownBacklog, err = parseBoundedInt64(lookup, "TALENRO_OUTBOX_DOWN_BACKLOG", cfg.OutboxDownBacklog, cfg.OutboxDegradedBacklog+1, 100000); err != nil {
		return Config{}, err
	}
	if cfg.OutboxDegradedAge, err = parseDuration(lookup, "TALENRO_OUTBOX_DEGRADED_AGE", cfg.OutboxDegradedAge, 10*time.Second, 10*time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.OutboxDownAge, err = parseDuration(lookup, "TALENRO_OUTBOX_DOWN_AGE", cfg.OutboxDownAge, cfg.OutboxDegradedAge+time.Second, time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.ErrorReportQueue, err = parseBoundedInt(lookup, "TALENRO_ERROR_REPORT_QUEUE", cfg.ErrorReportQueue, 10, 1000); err != nil {
		return Config{}, err
	}
	if cfg.ErrorReportBatch, err = parseBoundedInt(lookup, "TALENRO_ERROR_REPORT_BATCH", cfg.ErrorReportBatch, 1, 100); err != nil {
		return Config{}, err
	}
	if cfg.ErrorReportBatch > cfg.ErrorReportQueue {
		return Config{}, fmt.Errorf("parse TALENRO_ERROR_REPORT_BATCH: invalid batch size")
	}

	cfg.Security, err = loadSecurity(lookup)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func loadSecurity(lookup Lookup) (SecurityConfig, error) {
	profile, err := parseProfile(lookup)
	if err != nil {
		return SecurityConfig{}, err
	}
	security := defaultSecurity(profile)
	if security.EmailVerification, err = parseEmailMode(lookup, security.EmailVerification); err != nil {
		return SecurityConfig{}, err
	}
	if security.PublicBaseURL, err = parseBaseURL(lookup, "TALENRO_PUBLIC_BASE_URL", security.PublicBaseURL, profile == ProfileProduction); err != nil {
		return SecurityConfig{}, err
	}
	bundleKeys := [3]string{"TALENRO_PRIMARY_BUNDLE_BASE_URL", "TALENRO_MIRROR_A_BASE_URL", "TALENRO_MIRROR_B_BASE_URL"}
	for index, key := range bundleKeys {
		if security.BundleBaseURLs[index], err = parseBaseURL(lookup, key, security.BundleBaseURLs[index], profile == ProfileProduction); err != nil {
			return SecurityConfig{}, err
		}
	}
	if security.BundleBaseURLs[0] == security.BundleBaseURLs[1] || security.BundleBaseURLs[0] == security.BundleBaseURLs[2] || security.BundleBaseURLs[1] == security.BundleBaseURLs[2] {
		return SecurityConfig{}, fmt.Errorf("bundle base URLs must be distinct")
	}
	if security.WebAuthnRPID, err = parseRequiredString(lookup, "TALENRO_WEBAUTHN_RP_ID", security.WebAuthnRPID); err != nil {
		return SecurityConfig{}, err
	}
	if security.WebAuthnOrigins, err = parseOrigins(lookup, profile, security.WebAuthnOrigins); err != nil {
		return SecurityConfig{}, err
	}
	if security.SignerProvider, err = parseProvider(lookup, "TALENRO_SIGNER_PROVIDER", security.SignerProvider); err != nil {
		return SecurityConfig{}, err
	}
	if security.FieldProtectorProvider, err = parseProvider(lookup, "TALENRO_FIELD_PROTECTOR_PROVIDER", security.FieldProtectorProvider); err != nil {
		return SecurityConfig{}, err
	}
	if security.EmailProvider, err = parseProvider(lookup, "TALENRO_EMAIL_PROVIDER", security.EmailProvider); err != nil {
		return SecurityConfig{}, err
	}
	if security.ErrorReporterProvider, err = parseProvider(lookup, "TALENRO_ERROR_REPORTER_PROVIDER", security.ErrorReporterProvider); err != nil {
		return SecurityConfig{}, err
	}
	if security.RequestDeadline, err = parseDuration(lookup, "TALENRO_REQUEST_DEADLINE", security.RequestDeadline, 2*time.Second, 10*time.Second); err != nil {
		return SecurityConfig{}, err
	}
	if security.RedisTimeout, err = parseDuration(lookup, "TALENRO_REDIS_TIMEOUT", security.RedisTimeout, 100*time.Millisecond, time.Second); err != nil {
		return SecurityConfig{}, err
	}
	if security.SignerTimeout, err = parseDuration(lookup, "TALENRO_SIGNER_TIMEOUT", security.SignerTimeout, 500*time.Millisecond, 5*time.Second); err != nil {
		return SecurityConfig{}, err
	}
	if security.ErrorReportTimeout, err = parseDuration(lookup, "TALENRO_ERROR_REPORT_TIMEOUT", security.ErrorReportTimeout, 100*time.Millisecond, 2*time.Second); err != nil {
		return SecurityConfig{}, err
	}
	if security.ClockSkew, err = parseDuration(lookup, "TALENRO_CLOCK_SKEW", security.ClockSkew, 30*time.Second, 300*time.Second); err != nil {
		return SecurityConfig{}, err
	}
	if security.LoginRateLimit, err = parseRateLimit(lookup, "TALENRO_LOGIN", security.LoginRateLimit, 5, 50, 5*time.Minute, time.Hour); err != nil {
		return SecurityConfig{}, err
	}
	if security.DeliveryRateLimit, err = parseRateLimit(lookup, "TALENRO_DELIVERY", security.DeliveryRateLimit, 1, 20, 10*time.Minute, 24*time.Hour); err != nil {
		return SecurityConfig{}, err
	}
	if security.ChallengeRateLimit, err = parseRateLimit(lookup, "TALENRO_CHALLENGE", security.ChallengeRateLimit, 5, 100, time.Minute, 30*time.Minute); err != nil {
		return SecurityConfig{}, err
	}

	if profile == ProfileProduction {
		if security.EmailVerification == EmailDisabled {
			return SecurityConfig{}, fmt.Errorf("TALENRO_EMAIL_VERIFICATION_MODE is unsafe for production")
		}
		if security.SignerProvider != ProviderExternal || security.FieldProtectorProvider != ProviderExternal || security.EmailProvider != ProviderExternal || security.ErrorReporterProvider == ProviderLocal {
			return SecurityConfig{}, fmt.Errorf("production requires external security providers")
		}
		for _, key := range localKeyNames() {
			if value, exists := lookup(key); exists && value != "" {
				return SecurityConfig{}, fmt.Errorf("production forbids local key material")
			}
		}
		return security, nil
	}

	if security.SensitiveLookupKey, err = parseKey(lookup, "TALENRO_SENSITIVE_LOOKUP_KEY_B64", defaultLookupKeyB64); err != nil {
		return SecurityConfig{}, err
	}
	if security.SensitiveEncryptionKey, err = parseKey(lookup, "TALENRO_SENSITIVE_ENCRYPTION_KEY_B64", defaultEncryptionKeyB64); err != nil {
		return SecurityConfig{}, err
	}
	if security.LocalRootSigningSeed, err = parseKey(lookup, "TALENRO_LOCAL_ROOT_SIGNING_SEED_B64", defaultRootSigningSeedB64); err != nil {
		return SecurityConfig{}, err
	}
	if security.LocalConfigSigningSeed, err = parseKey(lookup, "TALENRO_LOCAL_CONFIG_SIGNING_SEED_B64", defaultConfigSigningSeedB64); err != nil {
		return SecurityConfig{}, err
	}
	return security, nil
}

func defaultSecurity(profile Profile) SecurityConfig {
	return SecurityConfig{
		Profile:                profile,
		EmailVerification:      EmailDisabled,
		PublicBaseURL:          "http://localhost:8080",
		BundleBaseURLs:         [3]string{"http://localhost:8080", "http://localhost:8081", "http://localhost:8082"},
		WebAuthnRPID:           "localhost",
		WebAuthnOrigins:        []string{"http://localhost:8080"},
		SignerProvider:         ProviderLocal,
		FieldProtectorProvider: ProviderLocal,
		EmailProvider:          ProviderLocal,
		ErrorReporterProvider:  ProviderDiscard,
		RequestDeadline:        5 * time.Second,
		RedisTimeout:           250 * time.Millisecond,
		SignerTimeout:          2 * time.Second,
		ErrorReportTimeout:     time.Second,
		ClockSkew:              120 * time.Second,
		LoginRateLimit:         RateLimitPolicy{Limit: 10, Window: 15 * time.Minute},
		DeliveryRateLimit:      RateLimitPolicy{Limit: 5, Window: time.Hour},
		ChallengeRateLimit:     RateLimitPolicy{Limit: 20, Window: 5 * time.Minute},
	}
}

func parseBool(lookup Lookup, key string, fallback bool) (bool, error) {
	value, exists := lookup(key)
	if !exists || value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("parse %s: invalid boolean", key)
	}
	return parsed, nil
}

func parseBoundedInt(lookup Lookup, key string, fallback, minimum, maximum int) (int, error) {
	parsed, err := parseBoundedInt64(lookup, key, int64(fallback), int64(minimum), int64(maximum))
	if err != nil {
		return 0, err
	}
	return int(parsed), nil
}

func parseBoundedInt64(lookup Lookup, key string, fallback, minimum, maximum int64) (int64, error) {
	value, exists := lookup(key)
	if !exists || value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < minimum || parsed > maximum {
		return 0, fmt.Errorf("parse %s: invalid integer", key)
	}
	return parsed, nil
}

func parseProfile(lookup Lookup) (Profile, error) {
	value, exists := lookup("TALENRO_PROFILE")
	if !exists || value == "" {
		return ProfileLocal, nil
	}
	profile := Profile(value)
	if profile != ProfileLocal && profile != ProfileTest && profile != ProfileProduction {
		return "", fmt.Errorf("parse TALENRO_PROFILE: invalid profile")
	}
	return profile, nil
}

func parseEmailMode(lookup Lookup, fallback EmailVerificationMode) (EmailVerificationMode, error) {
	value, exists := lookup("TALENRO_EMAIL_VERIFICATION_MODE")
	if !exists || value == "" {
		return fallback, nil
	}
	mode := EmailVerificationMode(value)
	if mode != EmailRequired && mode != EmailGrace && mode != EmailDisabled {
		return "", fmt.Errorf("parse TALENRO_EMAIL_VERIFICATION_MODE: invalid mode")
	}
	return mode, nil
}

func parseProvider(lookup Lookup, key string, fallback Provider) (Provider, error) {
	value, exists := lookup(key)
	if !exists || value == "" {
		return fallback, nil
	}
	provider := Provider(value)
	if provider != ProviderLocal && provider != ProviderExternal && provider != ProviderDiscard {
		return "", fmt.Errorf("parse %s: invalid provider", key)
	}
	return provider, nil
}

func parseRequiredString(lookup Lookup, key, fallback string) (string, error) {
	value, exists := lookup(key)
	if !exists || value == "" {
		value = fallback
	}
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func parseBaseURL(lookup Lookup, key, fallback string, requireHTTPS bool) (string, error) {
	value, exists := lookup(key)
	if !exists || value == "" {
		value = fallback
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" || parsed.Opaque != "" || (parsed.Path != "" && parsed.Path != "/") || (parsed.Scheme != "http" && parsed.Scheme != "https") || (requireHTTPS && parsed.Scheme != "https") {
		return "", fmt.Errorf("parse %s: invalid base URL", key)
	}
	parsed.Path = ""
	return parsed.String(), nil
}

func parseOrigins(lookup Lookup, profile Profile, fallback []string) ([]string, error) {
	value, exists := lookup("TALENRO_WEBAUTHN_ORIGINS")
	if !exists || value == "" {
		return append([]string(nil), fallback...), nil
	}
	origins := strings.Split(value, ",")
	for index, origin := range origins {
		parsed, err := url.ParseRequestURI(origin)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Opaque != "" || parsed.Path != "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || (profile == ProfileProduction && parsed.Scheme != "https") {
			return nil, fmt.Errorf("parse TALENRO_WEBAUTHN_ORIGINS: invalid origin")
		}
		origins[index] = parsed.String()
	}
	return origins, nil
}

func parseDuration(lookup Lookup, key string, fallback, minimum, maximum time.Duration) (time.Duration, error) {
	value, exists := lookup(key)
	if !exists || value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed < minimum || parsed > maximum {
		return 0, fmt.Errorf("parse %s: invalid duration", key)
	}
	return parsed, nil
}

func parseRateLimit(lookup Lookup, prefix string, fallback RateLimitPolicy, minLimit, maxLimit uint32, minWindow, maxWindow time.Duration) (RateLimitPolicy, error) {
	limitKey := prefix + "_RATE_LIMIT"
	windowKey := prefix + "_RATE_WINDOW"
	limitValue, limitExists := lookup(limitKey)
	windowValue, windowExists := lookup(windowKey)
	if !limitExists || limitValue == "" {
		limitValue = strconv.FormatUint(uint64(fallback.Limit), 10)
	}
	if !windowExists || windowValue == "" {
		windowValue = fallback.Window.String()
	}
	limit, err := strconv.ParseUint(limitValue, 10, 32)
	if err != nil || limit < uint64(minLimit) || limit > uint64(maxLimit) {
		return RateLimitPolicy{}, fmt.Errorf("parse %s: invalid rate limit", limitKey)
	}
	window, err := time.ParseDuration(windowValue)
	if err != nil || window < minWindow || window > maxWindow {
		return RateLimitPolicy{}, fmt.Errorf("parse %s: invalid rate window", windowKey)
	}
	return RateLimitPolicy{Limit: uint32(limit), Window: window}, nil
}

func parseKey(lookup Lookup, key, fallback string) (secret.Bytes, error) {
	value, exists := lookup(key)
	if !exists || value == "" {
		value = fallback
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) != 32 {
		return secret.Bytes{}, fmt.Errorf("parse %s: invalid key material", key)
	}
	return secret.NewBytes(decoded), nil
}

func localKeyNames() [4]string {
	return [4]string{
		"TALENRO_SENSITIVE_LOOKUP_KEY_B64",
		"TALENRO_SENSITIVE_ENCRYPTION_KEY_B64",
		"TALENRO_LOCAL_ROOT_SIGNING_SEED_B64",
		"TALENRO_LOCAL_CONFIG_SIGNING_SEED_B64",
	}
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
