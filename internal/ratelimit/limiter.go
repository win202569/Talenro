// Package ratelimit defines privacy-safe rate-limit identities and backend contracts.
package ratelimit

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"time"

	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/secret"
)

const subjectDigestPrefix = "TALENRO-RATE-LIMIT-V1\x00"

// Operation is the finite rate-limited operation class.
type Operation string

const (
	// Login identifies account-login limits.
	Login Operation = "login"
	// Delivery identifies email and notification delivery limits.
	Delivery Operation = "delivery"
	// Challenge identifies security challenge and rotation limits.
	Challenge Operation = "challenge"
)

// Limiter decides whether one digested subject is allowed under a server policy.
type Limiter interface {
	Allow(context.Context, Operation, [32]byte, config.RateLimitPolicy) (bool, error)
}

var (
	// ErrInvalidInput reports invalid digest parameters without echoing them.
	ErrInvalidInput = errors.New("ratelimit: invalid input")
	// ErrBackendUnavailable reports a sanitized rate-limit backend failure.
	ErrBackendUnavailable = errors.New("ratelimit: backend unavailable")
)

// SubjectDigest returns the window-scoped HMAC lookup key sent to a backend.
func SubjectDigest(key secret.Bytes, operation Operation, at time.Time, policy config.RateLimitPolicy, canonicalSubject string) ([32]byte, error) {
	keyBytes := key.Copy()
	defer clear(keyBytes)
	if len(keyBytes) != sha256.Size || !validOperation(operation) || policy.Limit == 0 || policy.Window <= 0 || canonicalSubject == "" {
		return [32]byte{}, ErrInvalidInput
	}

	windowStart := at.UTC().Truncate(policy.Window).Format(time.RFC3339Nano)
	material := make([]byte, 0, len(subjectDigestPrefix)+len(operation)+len(windowStart)+len(canonicalSubject))
	material = append(material, subjectDigestPrefix...)
	material = append(material, operation...)
	material = append(material, windowStart...)
	material = append(material, canonicalSubject...)
	defer clear(material)

	mac := hmac.New(sha256.New, keyBytes)
	_, _ = mac.Write(material)
	sum := mac.Sum(nil)
	defer clear(sum)
	var digest [32]byte
	copy(digest[:], sum)
	return digest, nil
}

// NewBackendError discards a backend cause and returns a fixed sentinel.
func NewBackendError(_ error) error { return ErrBackendUnavailable }

func validOperation(operation Operation) bool {
	switch operation {
	case Login, Delivery, Challenge:
		return true
	default:
		return false
	}
}
