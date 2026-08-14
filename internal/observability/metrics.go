// Package observability provides bounded, privacy-safe service metrics.
package observability

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"talenro.local/platform/internal/buildinfo"
	"talenro.local/platform/internal/errorreport"
)

// MetricResult is the finite operation result label registry.
type MetricResult string

// SecurityOperation is the finite security-operation label registry.
type SecurityOperation string

// SecurityReason is the finite security-result reason label registry.
type SecurityReason string

// CryptoOperation is the finite cryptographic-operation label registry.
type CryptoOperation string

// CryptoReason is the finite cryptographic-result reason label registry.
type CryptoReason string

//nolint:revive // These closed metric-label registries are documented by their exported types and self-describing names.
const (
	// MetricResultSuccess and the following constants are finite metric label registries.
	MetricResultSuccess MetricResult = "success"
	// MetricResultFailure records a finite failed outcome.
	MetricResultFailure MetricResult = "failure"

	// SecurityOperationAccountRegister and the following constants are finite security operations.
	SecurityOperationAccountRegister SecurityOperation = "account_registration"
	SecurityOperationAccountAuth     SecurityOperation = "account_authentication"
	SecurityOperationAccountRecovery SecurityOperation = "account_recovery"
	SecurityOperationDeviceEnroll    SecurityOperation = "device_enrollment"
	SecurityOperationDeviceAuth      SecurityOperation = "device_authentication"
	SecurityOperationDeviceRevoke    SecurityOperation = "device_revocation"
	SecurityOperationBundleResolve   SecurityOperation = "bundle_resolution"
	SecurityOperationBundleAck       SecurityOperation = "bundle_acknowledgement"

	SecurityReasonNone              SecurityReason = "none"
	SecurityReasonInvalidCredential SecurityReason = "invalid_credential" // #nosec G101 -- a finite metric reason label, not a credential.
	SecurityReasonExpired           SecurityReason = "expired"
	SecurityReasonRevoked           SecurityReason = "revoked"
	SecurityReasonRateLimited       SecurityReason = "rate_limited"
	SecurityReasonDependency        SecurityReason = "dependency"
	SecurityReasonInternal          SecurityReason = "internal"
	SecurityReasonReplay            SecurityReason = "replay"

	CryptoOperationLookup         CryptoOperation = "lookup"
	CryptoOperationEncrypt        CryptoOperation = "encrypt"
	CryptoOperationDecrypt        CryptoOperation = "decrypt"
	CryptoOperationTokenVerify    CryptoOperation = "token_verify"
	CryptoOperationProofVerify    CryptoOperation = "proof_verify"
	CryptoOperationMetadataVerify CryptoOperation = "metadata_verify"
	CryptoOperationBundleVerify   CryptoOperation = "bundle_verify"
	CryptoOperationBundleDecrypt  CryptoOperation = "bundle_decrypt"
	CryptoOperationBundleSign     CryptoOperation = "bundle_sign"

	CryptoReasonNone           CryptoReason = "none"
	CryptoReasonMalformed      CryptoReason = "malformed"
	CryptoReasonInvalid        CryptoReason = "invalid"
	CryptoReasonExpired        CryptoReason = "expired"
	CryptoReasonRevoked        CryptoReason = "revoked"
	CryptoReasonRollback       CryptoReason = "rollback"
	CryptoReasonKeyUnavailable CryptoReason = "key_unavailable"
	CryptoReasonWrongDomain    CryptoReason = "wrong_domain"
)

// Registry owns the control API's bounded Prometheus collectors.
type Registry struct {
	prometheus.Registerer
	Gatherer       prometheus.Gatherer
	Requests       *prometheus.CounterVec
	Duration       *prometheus.HistogramVec
	BuildInfo      *prometheus.GaugeVec
	securityEvents *prometheus.CounterVec
	outboxBacklog  prometheus.Gauge
	outboxOldest   prometheus.Gauge
	errorReports   *prometheus.CounterVec
	cryptoChecks   *prometheus.CounterVec
}

var _ errorreport.Observer = (*Registry)(nil)

// NewRegistry constructs an isolated registry with control API collectors.
func NewRegistry() *Registry {
	registry := prometheus.NewRegistry()
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "talenro_control_http_requests_total",
		Help: "Completed control API HTTP requests.",
	}, []string{"method", "route", "status_class"})
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "talenro_control_http_request_duration_seconds",
		Help:    "Control API HTTP request duration.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route", "status_class"})
	build := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "talenro_control_build_info",
		Help: "Talenro control API build identity.",
	}, []string{"version", "commit"})
	securityEvents := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "talenro_security_events_total",
		Help: "Security outcomes by finite operation, result, and reason.",
	}, []string{"operation", "result", "reason"})
	outboxBacklog := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "talenro_outbox_backlog",
		Help: "Current number of unpublished transactional outbox events.",
	})
	outboxOldest := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "talenro_outbox_oldest_seconds",
		Help: "Current age in seconds of the oldest unpublished outbox event.",
	})
	errorReports := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "talenro_error_reports_total",
		Help: "Privacy-safe error report delivery outcomes by finite component and result.",
	}, []string{"component", "result"})
	cryptoChecks := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "talenro_crypto_validations_total",
		Help: "Cryptographic validation outcomes by finite operation, result, and reason.",
	}, []string{"operation", "result", "reason"})
	info := buildinfo.Current()
	build.WithLabelValues(info.Version, info.Commit).Set(1)
	registry.MustRegister(requests, duration, build, securityEvents, outboxBacklog, outboxOldest, errorReports, cryptoChecks)
	return &Registry{
		Registerer: registry, Gatherer: registry,
		Requests: requests, Duration: duration, BuildInfo: build,
		securityEvents: securityEvents, outboxBacklog: outboxBacklog,
		outboxOldest: outboxOldest, errorReports: errorReports, cryptoChecks: cryptoChecks,
	}
}

// RecordSecurity records one allowlisted security outcome.
func (r *Registry) RecordSecurity(operation SecurityOperation, result MetricResult, reason SecurityReason) bool {
	if r == nil || !validSecurityOperation(operation) || !validMetricResult(result) || !validSecurityReason(reason) {
		return false
	}
	r.securityEvents.WithLabelValues(string(operation), string(result), string(reason)).Inc()
	return true
}

// SetOutbox records a nonnegative point-in-time outbox health snapshot.
func (r *Registry) SetOutbox(backlog int64, oldest time.Duration) bool {
	if r == nil || backlog < 0 || oldest < 0 {
		return false
	}
	r.outboxBacklog.Set(float64(backlog))
	r.outboxOldest.Set(oldest.Seconds())
	return true
}

// Record implements errorreport.Observer while rejecting values outside the
// reporter's finite component/result registries.
func (r *Registry) Record(component errorreport.Component, result errorreport.DeliveryResult) {
	if r == nil || !validReportComponent(component) || !validReportResult(result) {
		return
	}
	r.errorReports.WithLabelValues(string(component), string(result)).Inc()
}

// RecordCrypto records one allowlisted cryptographic validation outcome.
func (r *Registry) RecordCrypto(operation CryptoOperation, result MetricResult, reason CryptoReason) bool {
	if r == nil || !validCryptoOperation(operation) || !validMetricResult(result) || !validCryptoReason(reason) {
		return false
	}
	r.cryptoChecks.WithLabelValues(string(operation), string(result), string(reason)).Inc()
	return true
}

// Handler serves metrics gathered from the registry.
func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.Gatherer, promhttp.HandlerOpts{})
}

func validMetricResult(value MetricResult) bool {
	return value == MetricResultSuccess || value == MetricResultFailure
}

func validSecurityOperation(value SecurityOperation) bool {
	switch value {
	case SecurityOperationAccountRegister, SecurityOperationAccountAuth, SecurityOperationAccountRecovery,
		SecurityOperationDeviceEnroll, SecurityOperationDeviceAuth, SecurityOperationDeviceRevoke,
		SecurityOperationBundleResolve, SecurityOperationBundleAck:
		return true
	default:
		return false
	}
}

func validSecurityReason(value SecurityReason) bool {
	switch value {
	case SecurityReasonNone, SecurityReasonInvalidCredential, SecurityReasonExpired, SecurityReasonRevoked,
		SecurityReasonRateLimited, SecurityReasonDependency, SecurityReasonInternal, SecurityReasonReplay:
		return true
	default:
		return false
	}
}

func validCryptoOperation(value CryptoOperation) bool {
	switch value {
	case CryptoOperationLookup, CryptoOperationEncrypt, CryptoOperationDecrypt, CryptoOperationTokenVerify,
		CryptoOperationProofVerify, CryptoOperationMetadataVerify, CryptoOperationBundleVerify, CryptoOperationBundleDecrypt,
		CryptoOperationBundleSign:
		return true
	default:
		return false
	}
}

func validCryptoReason(value CryptoReason) bool {
	switch value {
	case CryptoReasonNone, CryptoReasonMalformed, CryptoReasonInvalid, CryptoReasonExpired,
		CryptoReasonRevoked, CryptoReasonRollback, CryptoReasonKeyUnavailable, CryptoReasonWrongDomain:
		return true
	default:
		return false
	}
}

func validReportComponent(value errorreport.Component) bool {
	switch value {
	case errorreport.ComponentControlAPI, errorreport.ComponentIdentity, errorreport.ComponentDeviceAuth,
		errorreport.ComponentTrust, errorreport.ComponentOutbox, errorreport.ComponentEmail, errorreport.ComponentReporter:
		return true
	default:
		return false
	}
}

func validReportResult(value errorreport.DeliveryResult) bool {
	switch value {
	case errorreport.ResultSent, errorreport.ResultDropped, errorreport.ResultProviderFailed:
		return true
	default:
		return false
	}
}
