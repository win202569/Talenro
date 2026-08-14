package observability

import (
	"io"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/felixge/httpsnoop"
)

const (
	unmatchedRoute = "unmatched"
	otherMethod    = "other"
)

// Middleware records bounded request metrics around next.
func (r *Registry) Middleware(route string, next http.Handler) http.Handler {
	dynamicRoute := route == ""
	route = boundedRoute(route)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		started := time.Now()
		var status atomic.Int64
		commitOK := func() {
			status.CompareAndSwap(0, http.StatusOK)
		}
		wrapped := httpsnoop.Wrap(w, httpsnoop.Hooks{
			WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
				return func(code int) {
					next(code)
					if code < 100 || code > 199 {
						status.CompareAndSwap(0, int64(code))
					}
				}
			},
			Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
				return func(body []byte) (int, error) {
					written, err := next(body)
					commitOK()
					return written, err
				}
			},
			WriteString: func(next httpsnoop.WriteStringFunc) httpsnoop.WriteStringFunc {
				return func(body string) (int, error) {
					written, err := next(body)
					commitOK()
					return written, err
				}
			},
			ReadFrom: func(next httpsnoop.ReadFromFunc) httpsnoop.ReadFromFunc {
				return func(source io.Reader) (int64, error) {
					written, err := next(source)
					commitOK()
					return written, err
				}
			},
			Flush: func(next httpsnoop.FlushFunc) httpsnoop.FlushFunc {
				return func() {
					next()
					commitOK()
				}
			},
			FlushError: func(next httpsnoop.FlushErrorFunc) httpsnoop.FlushErrorFunc {
				return func() error {
					err := next()
					commitOK()
					return err
				}
			},
		})
		next.ServeHTTP(wrapped, req)
		statusCode := status.Load()
		if statusCode == 0 {
			statusCode = http.StatusOK
		}
		r.recordSecurityOutcome(req.Pattern, int(statusCode))
		statusClass := strconv.Itoa(int(statusCode)/100) + "xx"
		requestRoute := route
		if dynamicRoute {
			requestRoute = boundedRoute(req.Pattern)
		}
		labels := []string{boundedMethod(req.Method), requestRoute, statusClass}
		r.Requests.WithLabelValues(labels...).Inc()
		r.Duration.WithLabelValues(labels...).Observe(time.Since(started).Seconds())
	})
}

func (r *Registry) recordSecurityOutcome(pattern string, status int) {
	operation, ok := securityOperationForPattern(pattern)
	if !ok {
		return
	}
	result, reason := securityResultForStatus(status)
	r.RecordSecurity(operation, result, reason)
}

func securityOperationForPattern(pattern string) (SecurityOperation, bool) {
	switch pattern {
	case "POST /v1/accounts", "POST /v1/email-verification-deliveries", "POST /v1/email-verifications":
		return SecurityOperationAccountRegister, true
	case "POST /v1/password-reset-deliveries", "POST /v1/password-resets", "POST /v1/recovery-code-consumptions":
		return SecurityOperationAccountRecovery, true
	case "POST /v1/password-changes", "POST /v1/account-sessions", "POST /v1/account-auth-challenges",
		"POST /v1/account-token-rotations", "POST /v1/account-session-revocations",
		"POST /v1/passkey-registration-options", "POST /v1/passkey-credentials",
		"POST /v1/passkey-authentication-options", "POST /v1/passkey-revocations",
		"POST /v1/totp-enrollments", "POST /v1/totp-verifications", "POST /v1/totp-revocations",
		"POST /v1/recovery-code-rotations":
		return SecurityOperationAccountAuth, true
	case "POST /v1/device-enrollment-grants", "POST /v1/device-auth-challenges", "POST /v1/devices":
		return SecurityOperationDeviceEnroll, true
	case "POST /v1/device-token-rotations":
		return SecurityOperationDeviceAuth, true
	case "POST /v1/device-revocations":
		return SecurityOperationDeviceRevoke, true
	case "POST /v1/config-bundle-resolutions", "GET /b/{bundle_locator}":
		return SecurityOperationBundleResolve, true
	case "POST /v1/config-bundle-acknowledgements":
		return SecurityOperationBundleAck, true
	default:
		return "", false
	}
}

func securityResultForStatus(status int) (MetricResult, SecurityReason) {
	if status >= http.StatusOK && status < http.StatusMultipleChoices {
		return MetricResultSuccess, SecurityReasonNone
	}
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return MetricResultFailure, SecurityReasonInvalidCredential
	case http.StatusConflict:
		return MetricResultFailure, SecurityReasonReplay
	case http.StatusTooManyRequests:
		return MetricResultFailure, SecurityReasonRateLimited
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return MetricResultFailure, SecurityReasonDependency
	default:
		return MetricResultFailure, SecurityReasonInternal
	}
}

func boundedMethod(method string) string {
	if method == http.MethodGet {
		return method
	}
	return otherMethod
}

func boundedRoute(route string) string {
	switch route {
	case "/livez", "GET /livez":
		return "/livez"
	case "/readyz", "GET /readyz":
		return "/readyz"
	default:
		return unmatchedRoute
	}
}
