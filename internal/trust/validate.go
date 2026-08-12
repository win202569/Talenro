package trust

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"math"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"talenro.local/platform/internal/strictjson"
)

const (
	maximumCanonicalPayloadBytes = 64 << 10
	maximumPolicyBytes           = 4 << 10
	maximumMessageBytes          = 256
	maximumPaddingEncodedBytes   = 64 << 10
	minimumKeyIDBytes            = 16
	maximumKeyIDBytes            = 64
	maximumSigningKeys           = 64
	maximumMetadataVersions      = 1024
)

// DecodePayloadV1 strictly decodes and validates one bounded v1 payload.
func DecodePayloadV1(body []byte) (PayloadV1, error) {
	var value PayloadV1
	if len(body) < 2 || len(body) > maximumCanonicalPayloadBytes || !validIJSON(body) {
		return PayloadV1{}, ErrInvalidJSON
	}
	members, ok := exactJSONMembers(body, maximumCanonicalPayloadBytes, []string{
		"schema_version", "bundle_id", "bundle_locator", "bundle_version", "audience",
		"issued_at", "not_before", "expires_at", "policy_snapshot", "test_config",
	})
	if !ok {
		return PayloadV1{}, ErrInvalidJSON
	}
	if _, ok := exactJSONMembers(members["test_config"], maximumCanonicalPayloadBytes, []string{"message", "sequence"}); !ok {
		return PayloadV1{}, ErrInvalidJSON
	}
	if err := strictjson.Decode(bytes.NewReader(body), maximumCanonicalPayloadBytes, &value); err != nil {
		return PayloadV1{}, ErrInvalidJSON
	}
	if err := ValidatePayloadV1(value); err != nil {
		return PayloadV1{}, err
	}
	value.PolicySnapshot = bytes.Clone(value.PolicySnapshot)
	return value, nil
}

// ValidatePayloadV1 enforces all finite v1 field, time, and policy rules.
func ValidatePayloadV1(value PayloadV1) error {
	issuedAt, issuedOK := parseProtocolTime(value.IssuedAt)
	notBefore, notBeforeOK := parseProtocolTime(value.NotBefore)
	expiresAt, expiresOK := parseProtocolTime(value.ExpiresAt)
	if value.SchemaVersion != BundleSchemaV1 || !canonicalUUID(value.BundleID) || !canonicalUUID(value.Audience) ||
		!exactBase64URL(value.BundleLocator, 32) || !positiveIntegerString(value.BundleVersion) ||
		!issuedOK || !notBeforeOK || !expiresOK || notBefore.Before(issuedAt) || !expiresAt.After(notBefore) ||
		len(value.TestConfig.Message) > maximumMessageBytes || !utf8.ValidString(value.TestConfig.Message) ||
		!unsignedIntegerString(value.TestConfig.Sequence) || !validDevicePolicyV1(value.PolicySnapshot) {
		return ErrInvalidPayload
	}
	return nil
}

// DecodeSignedBundleV1 strictly decodes and validates one bounded signed frame.
func DecodeSignedBundleV1(body []byte) (SignedBundleV1, error) {
	var value SignedBundleV1
	if len(body) < 2 || len(body) > maximumCanonicalPayloadBytes || !validIJSON(body) {
		return SignedBundleV1{}, ErrInvalidJSON
	}
	if _, ok := exactJSONMembers(body, maximumCanonicalPayloadBytes, []string{
		"payload_jcs", "payload_sha256", "signer_key_id", "algorithm", "signature", "padding",
	}); !ok {
		return SignedBundleV1{}, ErrInvalidJSON
	}
	if err := strictjson.Decode(bytes.NewReader(body), maximumCanonicalPayloadBytes, &value); err != nil {
		return SignedBundleV1{}, ErrInvalidJSON
	}
	if err := ValidateSignedBundleV1(value); err != nil {
		return SignedBundleV1{}, err
	}
	return value, nil
}

func exactJSONMembers(body []byte, maximumBytes int64, required []string) (map[string]json.RawMessage, bool) {
	members := make(map[string]json.RawMessage, len(required))
	if strictjson.Decode(bytes.NewReader(body), maximumBytes, &members) != nil || len(members) != len(required) {
		return nil, false
	}
	for _, name := range required {
		if _, exists := members[name]; !exists {
			return nil, false
		}
	}
	return members, true
}

// ValidateSignedBundleV1 enforces the fixed signed-frame encodings and bounds.
func ValidateSignedBundleV1(value SignedBundleV1) error {
	if len(value.PayloadJCS) < 2 || len(value.PayloadJCS) > maximumCanonicalPayloadBytes ||
		!exactBase64URL(value.PayloadSHA256, 32) || !validKeyID(value.SignerKeyID) ||
		value.Algorithm != SignatureAlgorithm || !exactBase64URL(value.Signature, 64) ||
		len(value.Padding) > maximumPaddingEncodedBytes || !validOptionalBase64URL(value.Padding) {
		return ErrInvalidPayload
	}
	return nil
}

func validDevicePolicyV1(policy json.RawMessage) bool {
	if len(policy) < 2 || len(policy) > maximumPolicyBytes {
		return false
	}
	type finitePolicy struct {
		Mode       string          `json:"mode"`
		ExpiresAt  json.RawMessage `json:"expires_at"`
		MaxDevices json.RawMessage `json:"max_devices"`
	}
	var decoded finitePolicy
	if err := strictjson.Decode(bytes.NewReader(policy), maximumPolicyBytes, &decoded); err != nil {
		return false
	}
	if decoded.Mode == "standard" {
		return len(decoded.ExpiresAt) == 0 && len(decoded.MaxDevices) == 0
	}
	if decoded.Mode != "trial_restricted" || len(decoded.ExpiresAt) == 0 || len(decoded.MaxDevices) == 0 {
		return false
	}
	var expiresAt string
	var maxDevices string
	if strictjson.Decode(bytes.NewReader(decoded.ExpiresAt), int64(len(decoded.ExpiresAt)), &expiresAt) != nil ||
		strictjson.Decode(bytes.NewReader(decoded.MaxDevices), int64(len(decoded.MaxDevices)), &maxDevices) != nil {
		return false
	}
	_, validTime := parseProtocolTime(expiresAt)
	return validTime && maxDevices == "1"
}

func positiveIntegerString(value string) bool {
	if !unsignedIntegerString(value) || value == "0" {
		return false
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	return err == nil && parsed <= math.MaxInt64
}

func unsignedIntegerString(value string) bool {
	if len(value) == 0 || len(value) > 20 || len(value) > 1 && value[0] == '0' {
		return false
	}
	for index := range len(value) {
		if value[index] < '0' || value[index] > '9' {
			return false
		}
	}
	_, err := strconv.ParseUint(value, 10, 64)
	return err == nil
}

func parseProtocolTime(value string) (time.Time, bool) {
	if len(value) != len("2006-01-02T15:04:05Z") {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, value)
	return parsed, err == nil && parsed.Location() == time.UTC && parsed.Format(time.RFC3339) == value
}

func canonicalUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}

func exactBase64URL(value string, size int) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && len(decoded) == size && base64.RawURLEncoding.EncodeToString(decoded) == value
}

func validOptionalBase64URL(value string) bool {
	if value == "" {
		return true
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && base64.RawURLEncoding.EncodeToString(decoded) == value
}

func validKeyID(value string) bool {
	if len(value) < minimumKeyIDBytes || len(value) > maximumKeyIDBytes {
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
