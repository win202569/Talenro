// Package trustclient independently verifies and durably activates Talenro
// configuration envelopes. It intentionally does not call trust's parsing or
// verification helpers.
package trustclient

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

const (
	maximumMetadataBytes = 64 << 10
	maximumStateKeyBytes = 256
	maximumTrustedRoots  = 64
)

var (
	// ErrInvalidArgument reports an unusable dependency or trusted input.
	ErrInvalidArgument = errors.New("trustclient: invalid argument")
	// ErrEnvelope reports a malformed, unsupported, or unauthenticated outer
	// envelope and padded plaintext frame.
	ErrEnvelope = errors.New("trustclient: invalid envelope")
	// ErrRecipient reports a non-X25519 recipient key or HPKE open failure.
	ErrRecipient = errors.New("trustclient: recipient failure")
	// ErrMetadata reports malformed, expired, untrusted, or semantically
	// invalid trust metadata.
	ErrMetadata = errors.New("trustclient: invalid metadata")
	// ErrMetadataRollback reports a non-increasing signed metadata version.
	ErrMetadataRollback = errors.New("trustclient: metadata rollback")
	// ErrSignature reports an invalid bundle signature, digest, or key binding.
	ErrSignature = errors.New("trustclient: invalid signature")
	// ErrPayload reports a malformed or semantically invalid configuration.
	ErrPayload = errors.New("trustclient: invalid payload")
	// ErrRollback reports a bundle version at or below durable highest.
	ErrRollback = errors.New("trustclient: bundle rollback")
	// ErrStore collapses all filesystem, cancellation, and activation failures.
	ErrStore = errors.New("trustclient: store failure")
	// ErrStateNotFound reports a bounded state key with no durable record.
	ErrStateNotFound = errors.New("trustclient: state not found")
)

// Store is the frozen durable monotonic activation boundary.
type Store interface {
	LoadHighest(context.Context, string) (uint64, error)
	StageAndAdvance(context.Context, string, uint64, []byte) error
	Activate(context.Context, string, uint64) error
}

// TrustMetadata is immutable root-signed metadata plus its provisioned roots
// and the caller's already-durable metadata floor.
type TrustMetadata struct {
	signed         []byte
	trustedRoots   map[string]ed25519.PublicKey
	highestTrusted uint64
}

// NewTrustMetadata copies bounded metadata and provisioned public roots.
func NewTrustMetadata(signed []byte, trustedRoots map[string]ed25519.PublicKey, highestTrusted uint64) (TrustMetadata, error) {
	if len(signed) < 2 || len(signed) > maximumMetadataBytes || len(trustedRoots) == 0 || len(trustedRoots) > maximumTrustedRoots {
		return TrustMetadata{}, ErrInvalidArgument
	}
	roots := make(map[string]ed25519.PublicKey, len(trustedRoots))
	for keyID, publicKey := range trustedRoots {
		if !validKeyID(keyID) || len(publicKey) != ed25519.PublicKeySize {
			return TrustMetadata{}, ErrInvalidArgument
		}
		roots[keyID] = append(ed25519.PublicKey(nil), publicKey...)
	}
	return TrustMetadata{signed: append([]byte(nil), signed...), trustedRoots: roots, highestTrusted: highestTrusted}, nil
}

// Format redacts every fmt rendering of client trust metadata.
func (TrustMetadata) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("trustclient.TrustMetadata([REDACTED])"))
}

// LogValue redacts structured logging of client trust metadata.
func (TrustMetadata) LogValue() slog.Value {
	return slog.StringValue("trustclient.TrustMetadata([REDACTED])")
}

// MarshalJSON rejects generic serialization of client trust metadata.
func (TrustMetadata) MarshalJSON() ([]byte, error) { return nil, ErrInvalidArgument }

// UnmarshalJSON rejects untrusted construction that could bypass copying.
func (*TrustMetadata) UnmarshalJSON([]byte) error { return ErrInvalidArgument }

// Expected is the locally provisioned audience and per-bundle locator.
type Expected struct {
	audience string
	locator  [32]byte
}

// NewExpected validates and copies the local audience/locator binding.
func NewExpected(audience string, locator [32]byte) (Expected, error) {
	parsed, err := uuid.Parse(audience)
	if err != nil || parsed == uuid.Nil || parsed.String() != audience {
		return Expected{}, ErrInvalidArgument
	}
	return Expected{audience: audience, locator: locator}, nil
}

// Format redacts every fmt rendering of local expectations.
func (Expected) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("trustclient.Expected([REDACTED])"))
}

// LogValue redacts structured logging of local expectations.
func (Expected) LogValue() slog.Value {
	return slog.StringValue("trustclient.Expected([REDACTED])")
}

// MarshalJSON rejects generic serialization of local expectations.
func (Expected) MarshalJSON() ([]byte, error) { return nil, ErrInvalidArgument }

// UnmarshalJSON rejects untrusted construction that could bypass validation.
func (*Expected) UnmarshalJSON([]byte) error { return ErrInvalidArgument }

// VerifiedBundle is the trusted plaintext and distribution digest returned
// only after durable stage and activation.
type VerifiedBundle struct {
	version        uint64
	payloadJCS     []byte
	envelopeDigest [32]byte
}

// Version returns the durable bundle version.
func (value VerifiedBundle) Version() uint64 { return value.version }

// PayloadJCS returns a copy of the exact verified payload.
func (value VerifiedBundle) PayloadJCS() []byte { return append([]byte(nil), value.payloadJCS...) }

// EnvelopeSHA256 returns the digest of the exact canonical distribution bytes.
func (value VerifiedBundle) EnvelopeSHA256() [32]byte { return value.envelopeDigest }

// Format redacts every fmt rendering of verified plaintext.
func (VerifiedBundle) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("trustclient.VerifiedBundle([REDACTED])"))
}

// LogValue redacts structured logging of verified plaintext.
func (VerifiedBundle) LogValue() slog.Value {
	return slog.StringValue("trustclient.VerifiedBundle([REDACTED])")
}

// MarshalJSON rejects generic serialization of verified plaintext.
func (VerifiedBundle) MarshalJSON() ([]byte, error) { return nil, ErrInvalidArgument }

// UnmarshalJSON rejects untrusted construction of a verified result.
func (*VerifiedBundle) UnmarshalJSON([]byte) error { return ErrInvalidArgument }

func validKeyID(value string) bool {
	if len(value) < 16 || len(value) > 64 {
		return false
	}
	for index := range len(value) {
		character := value[index]
		if character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z' ||
			character >= '0' && character <= '9' || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
}

var _ json.Marshaler = TrustMetadata{}
var _ json.Marshaler = Expected{}
var _ json.Marshaler = VerifiedBundle{}
