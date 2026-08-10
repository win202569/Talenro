package securitykit

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"reflect"

	"talenro.local/platform/internal/secret"
)

const (
	opaqueTokenBytes      = 32
	opaqueTokenCharacters = 43
	// #nosec G101 -- this is a public domain-separation label, not a credential.
	tokenDigestPrefix = "TALENRO-TOKEN-DIGEST-V1\x00"
)

// TokenDomain separates hashes belonging to distinct token classes.
type TokenDomain string

const (
	// AccountAccessToken separates account access-token digests.
	// #nosec G101 -- this is a public domain label, not a credential.
	AccountAccessToken TokenDomain = "talenro/account-access/v1"
	// AccountRefreshToken separates account refresh-token digests.
	// #nosec G101 -- this is a public domain label, not a credential.
	AccountRefreshToken TokenDomain = "talenro/account-refresh/v1"
	// DeviceAccessToken separates device access-token digests.
	// #nosec G101 -- this is a public domain label, not a credential.
	DeviceAccessToken TokenDomain = "talenro/device-access/v1"
	// DeviceRefreshToken separates device refresh-token digests.
	// #nosec G101 -- this is a public domain label, not a credential.
	DeviceRefreshToken TokenDomain = "talenro/device-refresh/v1"
	// EnrollmentGrantToken separates enrollment-grant digests.
	// #nosec G101 -- this is a public domain label, not a credential.
	EnrollmentGrantToken TokenDomain = "talenro/enrollment-grant/v1"
	// EmailVerificationToken separates email-verification digests.
	// #nosec G101 -- this is a public domain label, not a credential.
	EmailVerificationToken TokenDomain = "talenro/email-verification/v1"
	// RecoveryCodeToken separates recovery-code digests.
	// #nosec G101 -- this is a public domain label, not a credential.
	RecoveryCodeToken TokenDomain = "talenro/recovery-code/v1"
)

var (
	// ErrInvalidArgument reports an unusable random source.
	ErrInvalidArgument = errors.New("securitykit: invalid argument")
	// ErrRandomSource reports a sanitized entropy-source failure.
	ErrRandomSource = errors.New("securitykit: random source failure")
	// ErrInvalidToken reports a malformed or noncanonical opaque token.
	ErrInvalidToken = errors.New("securitykit: invalid opaque token")
)

// NewOpaqueToken obtains exactly 32 random bytes and transfers an isolated copy.
func NewOpaqueToken(random RandomSource) (secret.Bytes, error) {
	if nilRandomSource(random) {
		return secret.Bytes{}, ErrInvalidArgument
	}
	buffer := make([]byte, opaqueTokenBytes)
	defer clear(buffer)
	count, err := io.ReadFull(random, buffer)
	if err != nil || count != opaqueTokenBytes {
		return secret.Bytes{}, ErrRandomSource
	}
	return secret.NewBytes(buffer), nil
}

// EncodeOpaqueToken returns the canonical unpadded base64url representation.
func EncodeOpaqueToken(raw secret.Bytes) string {
	bytes := raw.Copy()
	defer clear(bytes)
	if len(bytes) != opaqueTokenBytes {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

// DecodeOpaqueToken validates and owns a canonical 32-byte opaque token.
func DecodeOpaqueToken(value string) (secret.Bytes, error) {
	if len(value) != opaqueTokenCharacters {
		return secret.Bytes{}, ErrInvalidToken
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil {
		clear(decoded)
		return secret.Bytes{}, ErrInvalidToken
	}
	defer clear(decoded)
	if len(decoded) != opaqueTokenBytes {
		return secret.Bytes{}, ErrInvalidToken
	}
	canonical := make([]byte, base64.RawURLEncoding.EncodedLen(len(decoded)))
	defer clear(canonical)
	base64.RawURLEncoding.Encode(canonical, decoded)
	if string(canonical) != value {
		return secret.Bytes{}, ErrInvalidToken
	}
	return secret.NewBytes(decoded), nil
}

// DigestToken computes an unambiguous domain-separated SHA-256 lookup digest.
func DigestToken(domain TokenDomain, raw secret.Bytes) [32]byte {
	if !validTokenDomain(domain) {
		return [32]byte{}
	}
	rawBytes := raw.Copy()
	defer clear(rawBytes)
	if len(rawBytes) != opaqueTokenBytes {
		return [32]byte{}
	}

	domainBytes := []byte(domain)
	material := make([]byte, 0, len(tokenDigestPrefix)+len(domainBytes)+1+len(rawBytes))
	material = append(material, tokenDigestPrefix...)
	material = append(material, domainBytes...)
	material = append(material, 0)
	material = append(material, rawBytes...)
	digest := sha256.Sum256(material)
	clear(material)
	return digest
}

func validTokenDomain(domain TokenDomain) bool {
	switch domain {
	case AccountAccessToken, AccountRefreshToken, DeviceAccessToken, DeviceRefreshToken,
		EnrollmentGrantToken, EmailVerificationToken, RecoveryCodeToken:
		return true
	default:
		return false
	}
}

func nilRandomSource(random RandomSource) bool {
	if random == nil {
		return true
	}
	value := reflect.ValueOf(random)
	// Only nil-capable kinds can represent a typed-nil interface value.
	//nolint:exhaustive
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
