// Package trust provides canonical configuration payloads and signing metadata.
package trust

import (
	"context"
	"encoding/json"
	"errors"
)

const (
	// BundleSchemaV1 is the only configuration payload schema accepted by C1.1.
	BundleSchemaV1 = "talenro-config-bundle/v1"
	// TrustMetadataSchemaV1 is the only trust metadata schema accepted by C1.1.
	TrustMetadataSchemaV1 = "talenro-trust-metadata/v1"
	// SignatureAlgorithm fixes all Task 14 signatures to Ed25519.
	SignatureAlgorithm      = "Ed25519"
	bundleSignatureDomain   = "TALENRO-CONFIG-BUNDLE-SIGNATURE-V1\x00"
	metadataSignatureDomain = "TALENRO-TRUST-METADATA-V1\x00"
)

var (
	// ErrInvalidArgument reports an unusable caller-supplied dependency or value.
	ErrInvalidArgument = errors.New("trust: invalid argument")
	// ErrInvalidJSON reports a structurally unsafe JSON encoding.
	ErrInvalidJSON = errors.New("trust: invalid JSON")
	// ErrInvalidPayload reports a payload outside the exact v1 schema.
	ErrInvalidPayload = errors.New("trust: invalid payload")
	// ErrInvalidSignature reports a signature or signed binding that cannot be trusted.
	ErrInvalidSignature = errors.New("trust: invalid signature")
	// ErrClosed reports use after a local fixture erased its seed.
	ErrClosed = errors.New("trust: signer closed")
	// ErrSignerTimeout reports a signer call that exceeded its configured budget.
	ErrSignerTimeout = errors.New("trust: signer timeout")
	// ErrSignerFailure collapses all external signer/provider failures.
	ErrSignerFailure = errors.New("trust: signer failure")
	// ErrInvalidMetadata reports malformed or temporally invalid trust metadata.
	ErrInvalidMetadata = errors.New("trust: invalid metadata")
	// ErrMetadataRollback reports a non-increasing metadata version.
	ErrMetadataRollback = errors.New("trust: metadata rollback")
	// ErrUnknownRoot reports metadata whose root key is not provisioned.
	ErrUnknownRoot = errors.New("trust: unknown root")
	// ErrKeyUnavailable reports a key that cannot authorize the requested interval.
	ErrKeyUnavailable = errors.New("trust: signing key unavailable")
	// ErrRepository collapses storage and transaction failures.
	ErrRepository = errors.New("trust: repository failure")
)

// ConfigSigner is the frozen provider-neutral configuration signer boundary.
type ConfigSigner interface {
	KeyID() string
	Sign(context.Context, []byte) ([]byte, error)
}

// PayloadV1 is the exact canonical configuration payload schema.
type PayloadV1 struct {
	SchemaVersion  string          `json:"schema_version"`
	BundleID       string          `json:"bundle_id"`
	BundleLocator  string          `json:"bundle_locator"`
	BundleVersion  string          `json:"bundle_version"`
	Audience       string          `json:"audience"`
	IssuedAt       string          `json:"issued_at"`
	NotBefore      string          `json:"not_before"`
	ExpiresAt      string          `json:"expires_at"`
	PolicySnapshot json.RawMessage `json:"policy_snapshot"`
	TestConfig     TestConfigV1    `json:"test_config"`
}

// TestConfigV1 is the bounded C1.1 test configuration.
type TestConfigV1 struct {
	Message  string `json:"message"`
	Sequence string `json:"sequence"`
}

// SignedBundleV1 binds exact payload JCS to its digest and Ed25519 signature.
type SignedBundleV1 struct {
	PayloadJCS    string `json:"payload_jcs"`
	PayloadSHA256 string `json:"payload_sha256"`
	SignerKeyID   string `json:"signer_key_id"`
	Algorithm     string `json:"algorithm"`
	Signature     string `json:"signature"`
	Padding       string `json:"padding"`
}

// SigningKeyMetadataV1 is one complete public configuration-signing key entry.
type SigningKeyMetadataV1 struct {
	KeyID     string `json:"key_id"`
	Algorithm string `json:"algorithm"`
	PublicKey string `json:"public_key"`
	State     string `json:"state"`
	NotBefore string `json:"not_before"`
	NotAfter  string `json:"not_after"`
}

// RootMetadataV1 is the exact root-signed trust payload.
type RootMetadataV1 struct {
	SchemaVersion string                 `json:"schema_version"`
	Version       string                 `json:"version"`
	RootKeyID     string                 `json:"root_key_id"`
	RootAlgorithm string                 `json:"root_algorithm"`
	ValidFrom     string                 `json:"valid_from"`
	ValidUntil    string                 `json:"valid_until"`
	SigningKeys   []SigningKeyMetadataV1 `json:"signing_keys"`
}

// SignedRootMetadataV1 is the storage/wire record. Root identity and algorithm
// remain inside the signed canonical payload so neither can be substituted.
type SignedRootMetadataV1 struct {
	PayloadJCS string `json:"payload_jcs"`
	Signature  string `json:"signature"`
}

// MetadataRepository is the atomic complete-set publication boundary.
type MetadataRepository interface {
	List(context.Context) ([]SignedRootMetadataV1, error)
	Publish(context.Context, SignedRootMetadataV1) error
}
