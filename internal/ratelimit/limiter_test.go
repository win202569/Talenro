package ratelimit_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/ratelimit"
	"talenro.local/platform/internal/secret"
)

type fixedLimiter struct {
	allowed bool
	err     error
}

func (limiter fixedLimiter) Allow(context.Context, ratelimit.Operation, [32]byte, config.RateLimitPolicy) (bool, error) {
	return limiter.allowed, limiter.err
}

func TestLimiterContractReturnsAllowedOrDenied(t *testing.T) {
	t.Parallel()

	var limiter ratelimit.Limiter = fixedLimiter{allowed: true}
	allowed, err := limiter.Allow(context.Background(), ratelimit.Login, [32]byte{}, config.RateLimitPolicy{Limit: 10, Window: 15 * time.Minute})
	if err != nil || !allowed {
		t.Fatalf("allowed result = %v, %v", allowed, err)
	}
	limiter = fixedLimiter{allowed: false}
	allowed, err = limiter.Allow(context.Background(), ratelimit.Login, [32]byte{}, config.RateLimitPolicy{Limit: 10, Window: 15 * time.Minute})
	if err != nil || allowed {
		t.Fatalf("denied result = %v, %v", allowed, err)
	}
}

func TestSubjectDigestUsesFiniteOperationAndRotatesAtUTCWindow(t *testing.T) {
	t.Parallel()

	keyBytes := make([]byte, 32)
	for index := range keyBytes {
		keyBytes[index] = byte(index)
	}
	key := secret.NewBytes(keyBytes)
	policy := config.RateLimitPolicy{Limit: 10, Window: 15 * time.Minute}
	at := time.Date(2026, time.August, 9, 12, 34, 56, 0, time.UTC)
	digest, err := ratelimit.SubjectDigest(key, ratelimit.Login, at, policy, "user@example.com")
	if err != nil {
		t.Fatalf("SubjectDigest() error = %v", err)
	}
	const wantHex = "f78e2d9046bfb1fc224d5709daa597b7eee641489749313c11eed92d77258503"
	if got := hex.EncodeToString(digest[:]); got != wantHex {
		t.Fatalf("SubjectDigest() = %s, want %s", got, wantHex)
	}

	sameInstantOtherZone := at.In(time.FixedZone("offset", 4*60*60))
	sameDigest, err := ratelimit.SubjectDigest(key, ratelimit.Login, sameInstantOtherZone, policy, "user@example.com")
	if err != nil || sameDigest != digest {
		t.Fatalf("same instant digest = %x, %v; want %x", sameDigest, err, digest)
	}

	nextWindow, err := ratelimit.SubjectDigest(key, ratelimit.Login, at.Add(15*time.Minute), policy, "user@example.com")
	if err != nil || nextWindow == digest {
		t.Fatalf("next-window digest = %x, %v; want distinct", nextWindow, err)
	}
	otherOperation, err := ratelimit.SubjectDigest(key, ratelimit.Delivery, at, policy, "user@example.com")
	if err != nil || otherOperation == digest {
		t.Fatalf("other-operation digest = %x, %v; want distinct", otherOperation, err)
	}
	keyCopy := key.Copy()
	defer clear(keyCopy)
	if !bytes.Equal(keyCopy, keyBytes) {
		t.Fatal("SubjectDigest mutated caller-owned key")
	}
}

func TestSubjectDigestRejectsInvalidInputsWithoutDisclosure(t *testing.T) {
	t.Parallel()

	validKey := secret.NewBytes(bytes.Repeat([]byte{0x42}, 32))
	validPolicy := config.RateLimitPolicy{Limit: 1, Window: time.Minute}
	tests := []struct {
		name      string
		key       secret.Bytes
		operation ratelimit.Operation
		policy    config.RateLimitPolicy
		subject   string
	}{
		{name: "unknown operation", key: validKey, operation: ratelimit.Operation("SECRET-CANARY"), policy: validPolicy, subject: "subject-canary"},
		{name: "zero limit", key: validKey, operation: ratelimit.Login, policy: config.RateLimitPolicy{Window: time.Minute}, subject: "subject-canary"},
		{name: "zero window", key: validKey, operation: ratelimit.Login, policy: config.RateLimitPolicy{Limit: 1}, subject: "subject-canary"},
		{name: "empty subject", key: validKey, operation: ratelimit.Login, policy: validPolicy, subject: ""},
		{name: "empty key", key: secret.Bytes{}, operation: ratelimit.Login, policy: validPolicy, subject: "subject-canary"},
		{name: "wrong key length", key: secret.NewBytes(bytes.Repeat([]byte{0x42}, 31)), operation: ratelimit.Login, policy: validPolicy, subject: "subject-canary"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			digest, err := ratelimit.SubjectDigest(test.key, test.operation, time.Unix(0, 0).UTC(), test.policy, test.subject)
			if !errors.Is(err, ratelimit.ErrInvalidInput) {
				t.Fatalf("SubjectDigest() error = %v, want ErrInvalidInput", err)
			}
			if digest != ([32]byte{}) {
				t.Fatalf("SubjectDigest() = %x after invalid input", digest)
			}
			for _, canary := range []string{"SECRET-CANARY", "subject-canary", "42424242"} {
				if bytes.Contains([]byte(err.Error()), []byte(canary)) {
					t.Fatalf("SubjectDigest() error disclosed %q", canary)
				}
			}
		})
	}
}

func TestBackendErrorDropsCauseAndSensitiveValues(t *testing.T) {
	t.Parallel()

	cause := errors.New("redis key=SECRET-CANARY value=subject-canary")
	err := ratelimit.NewBackendError(cause)
	if !errors.Is(err, ratelimit.ErrBackendUnavailable) {
		t.Fatalf("NewBackendError() = %v, want ErrBackendUnavailable", err)
	}
	if errors.Is(err, cause) {
		t.Fatal("NewBackendError() wrapped backend cause")
	}
	for _, canary := range []string{"SECRET-CANARY", "subject-canary", "redis key"} {
		if bytes.Contains([]byte(err.Error()), []byte(canary)) {
			t.Fatalf("NewBackendError() disclosed %q", canary)
		}
	}
}

func TestOperationEnumContainsOnlyThreeValues(t *testing.T) {
	t.Parallel()

	want := []string{"login", "delivery", "challenge"}
	got := []ratelimit.Operation{ratelimit.Login, ratelimit.Delivery, ratelimit.Challenge}
	for index := range want {
		if string(got[index]) != want[index] {
			t.Fatalf("operation[%d] = %q, want %q", index, got[index], want[index])
		}
	}
}
