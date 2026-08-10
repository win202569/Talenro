// Package apierrors defines finite internal errors and their sole public projection.
package apierrors

import (
	"errors"
	"time"

	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
)

// Code is the finite public error code contract.
type Code string

// Action is the finite client action contract.
type Action string

const (
	// MalformedRequest reports syntactically or structurally invalid input.
	MalformedRequest Code = "malformed_request"
	// AuthenticationFailed reports invalid or expired authentication material.
	AuthenticationFailed Code = "authentication_failed"
	// ActionNotAllowed reports a valid identity lacking permission for an action.
	ActionNotAllowed Code = "action_not_allowed"
	// IdempotencyConflict reports reuse of a key for a different request.
	IdempotencyConflict Code = "idempotency_conflict"
	// StateConflict reports an incompatible current resource state.
	StateConflict Code = "state_conflict"
	// VersionRollback reports an attempted monotonic-version decrease.
	VersionRollback Code = "version_rollback"
	// RequestTooLarge reports input beyond a fixed request bound.
	RequestTooLarge Code = "request_too_large"
	// UnsupportedSchema reports a client schema the service cannot consume.
	UnsupportedSchema Code = "unsupported_schema"
	// RateLimited reports exhaustion of a server-side operation budget.
	RateLimited Code = "rate_limited"
	// DependencyUnavailable reports a required dependency outage.
	DependencyUnavailable Code = "dependency_unavailable"
	// SigningUnavailable reports an unavailable signing capability.
	SigningUnavailable Code = "signing_unavailable"
)

const (
	// Retry asks the client to retry when the response permits it.
	Retry Action = "retry"
	// Reauthenticate asks the client to establish recent authentication.
	Reauthenticate Action = "reauthenticate"
	// Reenroll asks the client to enroll its device again.
	Reenroll Action = "reenroll"
	// UpgradeClient asks the user to install a supported client.
	UpgradeClient Action = "upgrade_client"
	// ContactSupport asks the user to seek support rather than retry.
	ContactSupport Action = "contact_support"
)

type retryClass uint8

const (
	retryNever retryClass = iota
	retryAfter
)

type retryMetadata struct {
	class   retryClass
	afterMS int64
}

type internalCategory uint8

const (
	categoryInput internalCategory = iota + 1
	categoryAuthentication
	categoryAuthorization
	categoryConflict
	categoryRateLimit
	categoryDependency
	categorySigning
)

// Error contains only finite classifications and never retains an underlying cause.
type Error struct {
	code     Code
	action   Action
	retry    retryMetadata
	category internalCategory
}

var errDirectSerialization = errors.New("apierrors: direct serialization forbidden")

// New builds a non-retryable classified error and fails closed on unknown enums.
func New(code Code, action Action) Error {
	if !validCode(code) || !validAction(action) {
		return fallbackError()
	}
	return Error{
		code:     code,
		action:   action,
		retry:    retryMetadata{class: retryNever},
		category: categoryFor(code),
	}
}

// NewRetryAfter builds a classified error with an OpenAPI-bounded millisecond delay.
func NewRetryAfter(code Code, action Action, after time.Duration) Error {
	if !validCode(code) || action != Retry || after < time.Millisecond || after > 10*time.Second || after%time.Millisecond != 0 {
		return fallbackError()
	}
	return Error{
		code:     code,
		action:   action,
		retry:    retryMetadata{class: retryAfter, afterMS: after.Milliseconds()},
		category: categoryFor(code),
	}
}

// Error returns a fixed, value-free description.
func (Error) Error() string { return "apierrors: classified failure" }

// MarshalJSON blocks accidental serialization; Public is the only JSON projection.
func (Error) MarshalJSON() ([]byte, error) { return nil, errDirectSerialization }

// Public returns the generated OpenAPI shape with a validated trace identifier.
func (err Error) Public(traceID string) controlapiv1.PublicError {
	if !validCode(err.code) || !validAction(err.action) {
		fallback := fallbackError()
		err.code = fallback.code
		err.action = fallback.action
	}
	if !validTraceID(traceID) {
		traceID = "trace-unavailable"
	}
	public := controlapiv1.PublicError{
		Code:    controlapiv1.PublicErrorCode(err.code),
		Action:  controlapiv1.PublicErrorAction(err.action),
		TraceId: traceID,
	}
	if err.retry.class == retryAfter && err.retry.afterMS >= 1 && err.retry.afterMS <= 10000 {
		retryAfterMS := err.retry.afterMS
		public.RetryAfterMs = &retryAfterMS
	}
	return public
}

func fallbackError() Error {
	return Error{
		code:     DependencyUnavailable,
		action:   ContactSupport,
		retry:    retryMetadata{class: retryNever},
		category: categoryDependency,
	}
}

func validCode(code Code) bool {
	switch code {
	case MalformedRequest, AuthenticationFailed, ActionNotAllowed, IdempotencyConflict, StateConflict,
		VersionRollback, RequestTooLarge, UnsupportedSchema, RateLimited, DependencyUnavailable, SigningUnavailable:
		return true
	default:
		return false
	}
}

func validAction(action Action) bool {
	switch action {
	case Retry, Reauthenticate, Reenroll, UpgradeClient, ContactSupport:
		return true
	default:
		return false
	}
}

func categoryFor(code Code) internalCategory {
	switch code {
	case MalformedRequest, RequestTooLarge, UnsupportedSchema:
		return categoryInput
	case AuthenticationFailed:
		return categoryAuthentication
	case ActionNotAllowed:
		return categoryAuthorization
	case IdempotencyConflict, StateConflict, VersionRollback:
		return categoryConflict
	case RateLimited:
		return categoryRateLimit
	case SigningUnavailable:
		return categorySigning
	case DependencyUnavailable:
		return categoryDependency
	default:
		return categoryDependency
	}
}

func validTraceID(value string) bool {
	if len(value) < 16 || len(value) > 64 {
		return false
	}
	for index := range len(value) {
		character := value[index]
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
}
