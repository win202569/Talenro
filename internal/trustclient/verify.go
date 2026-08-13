package trustclient

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/hpke"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math"
	"reflect"
	"strconv"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/gowebpki/jcs"
)

const (
	envelopeVersion          = "talenro-config-envelope/v1"
	kemName                  = "DHKEM-X25519-HKDF-SHA256"
	kdfName                  = "HKDF-SHA256"
	aeadName                 = "CHACHA20-POLY1305"
	hpkeInfo                 = "talenro-config-bundle/v1"
	bundleSchema             = "talenro-config-bundle/v1"
	metadataSchema           = "talenro-trust-metadata/v1"
	signatureAlgorithm       = "Ed25519"
	recipientSelectorDomain  = "TALENRO-RECIPIENT-SELECTOR-V1\x00"
	bundleSignatureDomain    = "TALENRO-CONFIG-BUNDLE-SIGNATURE-V1\x00"
	metadataSignatureDomain  = "TALENRO-TRUST-METADATA-V1\x00"
	configKeyIDDomain        = "TALENRO-CONFIG-SIGNING-KEY-ID-V1\x00"
	rootKeyIDDomain          = "TALENRO-ROOT-SIGNING-KEY-ID-V1\x00"
	maximumEnvelopeBytes     = 1 << 20
	maximumPlaintextBytes    = 64 << 10
	maximumPolicyBytes       = 4 << 10
	maximumMessageBytes      = 256
	maximumJSONDepth         = 16
	fixedBundleLifetime      = 24 * time.Hour
	fixedTrialGrace          = 24 * time.Hour
	minimumClockSkew         = 30 * time.Second
	maximumClockSkew         = 5 * time.Minute
	plaintextLengthPrefix    = 4
	chacha20Poly1305TagBytes = 16
)

var plaintextBucketSizes = [...]int{4096, 8192, 16384, 32768, 65536}

type outerEnvelopeWire struct {
	EnvelopeVersion string `json:"envelope_version"`
	KEM             string `json:"kem"`
	KDF             string `json:"kdf"`
	AEAD            string `json:"aead"`
	RecipientKeyID  string `json:"recipient_key_id"`
	BundleLocator   string `json:"bundle_locator"`
	Enc             string `json:"enc"`
	Ciphertext      string `json:"ciphertext"`
}

type outerHeaderWire struct {
	EnvelopeVersion string `json:"envelope_version"`
	KEM             string `json:"kem"`
	KDF             string `json:"kdf"`
	AEAD            string `json:"aead"`
	RecipientKeyID  string `json:"recipient_key_id"`
	BundleLocator   string `json:"bundle_locator"`
	Enc             string `json:"enc"`
}

type signedBundleWire struct {
	PayloadJCS    string `json:"payload_jcs"`
	PayloadSHA256 string `json:"payload_sha256"`
	SignerKeyID   string `json:"signer_key_id"`
	Algorithm     string `json:"algorithm"`
	Signature     string `json:"signature"`
	Padding       string `json:"padding"`
}

type payloadWire struct {
	SchemaVersion  string          `json:"schema_version"`
	BundleID       string          `json:"bundle_id"`
	BundleLocator  string          `json:"bundle_locator"`
	BundleVersion  string          `json:"bundle_version"`
	Audience       string          `json:"audience"`
	IssuedAt       string          `json:"issued_at"`
	NotBefore      string          `json:"not_before"`
	ExpiresAt      string          `json:"expires_at"`
	PolicySnapshot json.RawMessage `json:"policy_snapshot"`
	TestConfig     json.RawMessage `json:"test_config"`
}

type testConfigWire struct {
	Message  string `json:"message"`
	Sequence string `json:"sequence"`
}

type signedMetadataWire struct {
	PayloadJCS string `json:"payload_jcs"`
	Signature  string `json:"signature"`
}

type rootMetadataWire struct {
	SchemaVersion string                   `json:"schema_version"`
	Version       string                   `json:"version"`
	RootKeyID     string                   `json:"root_key_id"`
	RootAlgorithm string                   `json:"root_algorithm"`
	ValidFrom     string                   `json:"valid_from"`
	ValidUntil    string                   `json:"valid_until"`
	SigningKeys   []signingKeyMetadataWire `json:"signing_keys"`
}

type signingKeyMetadataWire struct {
	KeyID     string `json:"key_id"`
	Algorithm string `json:"algorithm"`
	PublicKey string `json:"public_key"`
	State     string `json:"state"`
	NotBefore string `json:"not_before"`
	NotAfter  string `json:"not_after"`
}

// VerifyAndStage independently authenticates, validates, durably stages, and
// activates one fixed-suite configuration envelope in the frozen order.
func VerifyAndStage(
	ctx context.Context,
	envelopeBytes []byte,
	privateKey ecdh.KeyExchanger,
	metadata TrustMetadata,
	expected Expected,
	store Store,
	now time.Time,
	clockSkew time.Duration,
) (VerifiedBundle, error) {
	if len(envelopeBytes) < 2 || len(envelopeBytes) > maximumEnvelopeBytes {
		return VerifiedBundle{}, ErrEnvelope
	}
	if isNil(ctx) || ctx.Err() != nil || isNil(privateKey) || isNil(store) ||
		len(metadata.signed) < 2 || len(metadata.signed) > maximumMetadataBytes || len(metadata.trustedRoots) == 0 ||
		expected.audience == "" || now.IsZero() || now.Location() != time.UTC ||
		clockSkew < minimumClockSkew || clockSkew > maximumClockSkew {
		return VerifiedBundle{}, ErrInvalidArgument
	}

	var envelope outerEnvelopeWire
	if _, err := decodeStrictObject(envelopeBytes, maximumEnvelopeBytes, &envelope, []string{
		"envelope_version", "kem", "kdf", "aead", "recipient_key_id", "bundle_locator", "enc", "ciphertext",
	}); err != nil {
		return VerifiedBundle{}, ErrEnvelope
	}
	canonicalEnvelope, err := canonicalJSON(envelopeBytes)
	if err != nil || !bytes.Equal(canonicalEnvelope, envelopeBytes) {
		clear(canonicalEnvelope)
		return VerifiedBundle{}, ErrEnvelope
	}
	clear(canonicalEnvelope)
	if envelope.EnvelopeVersion != envelopeVersion || envelope.KEM != kemName || envelope.KDF != kdfName || envelope.AEAD != aeadName {
		return VerifiedBundle{}, ErrEnvelope
	}
	enc, ok := decodeExactBase64URL(envelope.Enc, 32)
	if !ok {
		return VerifiedBundle{}, ErrEnvelope
	}
	ciphertext, ok := decodeCiphertext(envelope.Ciphertext)
	if !ok {
		clear(enc)
		return VerifiedBundle{}, ErrEnvelope
	}
	locator, ok := decodeExactBase64URL(envelope.BundleLocator, 32)
	if !ok || !bytes.Equal(locator, expected.locator[:]) {
		clear(enc)
		clear(ciphertext)
		clear(locator)
		return VerifiedBundle{}, ErrEnvelope
	}
	clear(locator)
	publicKey := privateKey.PublicKey()
	if publicKey == nil || privateKey.Curve() != ecdh.X25519() || publicKey.Curve() != ecdh.X25519() || len(publicKey.Bytes()) != 32 {
		clear(enc)
		clear(ciphertext)
		return VerifiedBundle{}, ErrRecipient
	}
	selectorMaterial := make([]byte, 0, len(recipientSelectorDomain)+32+32)
	selectorMaterial = append(selectorMaterial, recipientSelectorDomain...)
	selectorMaterial = append(selectorMaterial, publicKey.Bytes()...)
	selectorMaterial = append(selectorMaterial, expected.locator[:]...)
	selector := sha256.Sum256(selectorMaterial)
	clear(selectorMaterial)
	if envelope.RecipientKeyID != base64.RawURLEncoding.EncodeToString(selector[:16]) {
		clear(enc)
		clear(ciphertext)
		return VerifiedBundle{}, ErrEnvelope
	}
	header := outerHeaderWire{
		EnvelopeVersion: envelope.EnvelopeVersion, KEM: envelope.KEM, KDF: envelope.KDF, AEAD: envelope.AEAD,
		RecipientKeyID: envelope.RecipientKeyID, BundleLocator: envelope.BundleLocator, Enc: envelope.Enc,
	}
	headerBody, err := json.Marshal(header)
	if err != nil {
		clear(enc)
		clear(ciphertext)
		return VerifiedBundle{}, ErrEnvelope
	}
	aad, err := canonicalJSON(headerBody)
	clear(headerBody)
	if err != nil {
		clear(enc)
		clear(ciphertext)
		return VerifiedBundle{}, ErrEnvelope
	}
	hpkeKey, err := hpke.NewDHKEMPrivateKey(privateKey)
	if err != nil {
		clear(enc)
		clear(ciphertext)
		clear(aad)
		return VerifiedBundle{}, ErrRecipient
	}
	recipient, err := hpke.NewRecipient(enc, hpkeKey, hpke.HKDFSHA256(), hpke.ChaCha20Poly1305(), []byte(hpkeInfo))
	clear(enc)
	if err != nil {
		clear(ciphertext)
		clear(aad)
		return VerifiedBundle{}, ErrRecipient
	}
	plaintext, err := recipient.Open(aad, ciphertext)
	clear(aad)
	clear(ciphertext)
	if err != nil {
		clear(plaintext)
		return VerifiedBundle{}, ErrRecipient
	}
	signedBytes, ok := removeValidatedPadding(plaintext)
	clear(plaintext)
	if !ok {
		clear(signedBytes)
		return VerifiedBundle{}, ErrEnvelope
	}
	defer clear(signedBytes)

	var signed signedBundleWire
	if _, err := decodeStrictObject(signedBytes, maximumPlaintextBytes, &signed, []string{
		"payload_jcs", "payload_sha256", "signer_key_id", "algorithm", "signature", "padding",
	}); err != nil || !validSignedBundle(signed) {
		return VerifiedBundle{}, ErrSignature
	}
	canonicalSigned, err := canonicalJSON(signedBytes)
	if err != nil || !bytes.Equal(canonicalSigned, signedBytes) {
		clear(canonicalSigned)
		return VerifiedBundle{}, ErrSignature
	}
	clear(canonicalSigned)

	var payload payloadWire
	payloadBytes := []byte(signed.PayloadJCS)
	if _, err := decodeStrictObject(payloadBytes, maximumPlaintextBytes, &payload, []string{
		"schema_version", "bundle_id", "bundle_locator", "bundle_version", "audience", "issued_at", "not_before", "expires_at", "policy_snapshot", "test_config",
	}); err != nil {
		return VerifiedBundle{}, ErrPayload
	}
	canonicalPayload, err := canonicalJSON(payloadBytes)
	if err != nil || !bytes.Equal(canonicalPayload, payloadBytes) {
		clear(canonicalPayload)
		return VerifiedBundle{}, ErrPayload
	}
	defer clear(canonicalPayload)
	digest := sha256.Sum256(canonicalPayload)
	if signed.PayloadSHA256 != base64.RawURLEncoding.EncodeToString(digest[:]) {
		return VerifiedBundle{}, ErrSignature
	}

	rootMetadata, signingPublicKey, err := verifyTrustMetadata(metadata, signed.SignerKeyID, now)
	if err != nil {
		clear(signingPublicKey)
		return VerifiedBundle{}, err
	}
	defer clear(signingPublicKey)
	signature, ok := decodeExactBase64URL(signed.Signature, ed25519.SignatureSize)
	if !ok {
		return VerifiedBundle{}, ErrSignature
	}
	message := make([]byte, 0, len(bundleSignatureDomain)+len(canonicalPayload))
	message = append(message, bundleSignatureDomain...)
	message = append(message, canonicalPayload...)
	verifiedSignature := ed25519.Verify(signingPublicKey, message, signature)
	clear(message)
	clear(signature)
	if !verifiedSignature {
		return VerifiedBundle{}, ErrSignature
	}

	version, notBefore, expiresAt, err := validatePayload(payload, expected, now, clockSkew)
	if err != nil {
		return VerifiedBundle{}, err
	}
	if !metadataKeyAuthorizes(rootMetadata, signed.SignerKeyID, notBefore, expiresAt) {
		return VerifiedBundle{}, ErrSignature
	}
	highest, err := store.LoadHighest(ctx, expected.audience)
	if err != nil {
		return VerifiedBundle{}, ErrStore
	}
	if version <= highest {
		return VerifiedBundle{}, ErrRollback
	}
	if err := store.StageAndAdvance(ctx, expected.audience, version, canonicalPayload); err != nil {
		if errors.Is(err, ErrRollback) {
			return VerifiedBundle{}, ErrRollback
		}
		return VerifiedBundle{}, ErrStore
	}
	if err := store.Activate(ctx, expected.audience, version); err != nil {
		if errors.Is(err, ErrRollback) {
			return VerifiedBundle{}, ErrRollback
		}
		return VerifiedBundle{}, ErrStore
	}
	return VerifiedBundle{
		version: version, payloadJCS: append([]byte(nil), canonicalPayload...), envelopeDigest: sha256.Sum256(envelopeBytes),
	}, nil
}

func verifyTrustMetadata(metadata TrustMetadata, signerKeyID string, now time.Time) (rootMetadataWire, ed25519.PublicKey, error) {
	var signed signedMetadataWire
	if _, err := decodeStrictObject(metadata.signed, maximumMetadataBytes, &signed, []string{"payload_jcs", "signature"}); err != nil ||
		len(signed.PayloadJCS) < 2 || len(signed.PayloadJCS) > maximumMetadataBytes {
		return rootMetadataWire{}, nil, ErrMetadata
	}
	var payload rootMetadataWire
	payloadBytes := []byte(signed.PayloadJCS)
	if _, err := decodeStrictObject(payloadBytes, maximumMetadataBytes, &payload, []string{
		"schema_version", "version", "root_key_id", "root_algorithm", "valid_from", "valid_until", "signing_keys",
	}); err != nil {
		return rootMetadataWire{}, nil, ErrMetadata
	}
	canonical, err := canonicalJSON(payloadBytes)
	if err != nil || !bytes.Equal(canonical, payloadBytes) {
		clear(canonical)
		return rootMetadataWire{}, nil, ErrMetadata
	}
	defer clear(canonical)
	version, validFrom, validUntil, ok := validateMetadataPayload(payload, now)
	if !ok {
		return rootMetadataWire{}, nil, ErrMetadata
	}
	rootPublicKey, exists := metadata.trustedRoots[payload.RootKeyID]
	if !exists || len(rootPublicKey) != ed25519.PublicKeySize || deriveKeyID(rootKeyIDDomain, rootPublicKey) != payload.RootKeyID {
		return rootMetadataWire{}, nil, ErrMetadata
	}
	rootSignature, ok := decodeExactBase64URL(signed.Signature, ed25519.SignatureSize)
	if !ok {
		return rootMetadataWire{}, nil, ErrMetadata
	}
	message := make([]byte, 0, len(metadataSignatureDomain)+len(canonical))
	message = append(message, metadataSignatureDomain...)
	message = append(message, canonical...)
	verified := ed25519.Verify(rootPublicKey, message, rootSignature)
	clear(message)
	clear(rootSignature)
	if !verified {
		return rootMetadataWire{}, nil, ErrMetadata
	}
	if version <= metadata.highestTrusted {
		return rootMetadataWire{}, nil, ErrMetadataRollback
	}
	_ = validFrom
	_ = validUntil
	for _, key := range payload.SigningKeys {
		if key.KeyID != signerKeyID {
			continue
		}
		publicKey, ok := decodeExactBase64URL(key.PublicKey, ed25519.PublicKeySize)
		if !ok || deriveKeyID(configKeyIDDomain, publicKey) != key.KeyID || key.Algorithm != signatureAlgorithm ||
			(key.State != "active" && key.State != "retiring") {
			clear(publicKey)
			return rootMetadataWire{}, nil, ErrSignature
		}
		return payload, ed25519.PublicKey(publicKey), nil
	}
	return rootMetadataWire{}, nil, ErrSignature
}

func validateMetadataPayload(payload rootMetadataWire, now time.Time) (uint64, time.Time, time.Time, bool) {
	version, ok := parsePositiveUint(payload.Version)
	validFrom, fromOK := parseProtocolTime(payload.ValidFrom)
	validUntil, untilOK := parseProtocolTime(payload.ValidUntil)
	if payload.SchemaVersion != metadataSchema || payload.RootAlgorithm != signatureAlgorithm || !validKeyID(payload.RootKeyID) ||
		!ok || !fromOK || !untilOK || !validUntil.After(validFrom) || now.Before(validFrom) || !now.Before(validUntil) ||
		len(payload.SigningKeys) == 0 || len(payload.SigningKeys) > maximumTrustedRoots {
		return 0, time.Time{}, time.Time{}, false
	}
	previous := ""
	seen := make(map[string]struct{}, len(payload.SigningKeys))
	for _, key := range payload.SigningKeys {
		keyNotBefore, beforeOK := parseProtocolTime(key.NotBefore)
		keyNotAfter, afterOK := parseProtocolTime(key.NotAfter)
		publicKey, keyOK := decodeExactBase64URL(key.PublicKey, ed25519.PublicKeySize)
		validState := key.State == "future" || key.State == "active" || key.State == "retiring" || key.State == "revoked"
		_, duplicate := seen[key.KeyID]
		valid := validKeyID(key.KeyID) && key.Algorithm == signatureAlgorithm && keyOK &&
			deriveKeyID(configKeyIDDomain, publicKey) == key.KeyID && validState && beforeOK && afterOK &&
			keyNotAfter.After(keyNotBefore) && !keyNotBefore.Before(validFrom) && !keyNotAfter.After(validUntil) &&
			!duplicate && (previous == "" || previous < key.KeyID)
		clear(publicKey)
		if !valid {
			return 0, time.Time{}, time.Time{}, false
		}
		seen[key.KeyID] = struct{}{}
		previous = key.KeyID
	}
	return version, validFrom, validUntil, true
}

func validatePayload(payload payloadWire, expected Expected, now time.Time, skew time.Duration) (uint64, time.Time, time.Time, error) {
	version, versionOK := parsePositiveUint(payload.BundleVersion)
	issuedAt, issuedOK := parseProtocolTime(payload.IssuedAt)
	notBefore, beforeOK := parseProtocolTime(payload.NotBefore)
	expiresAt, expiresOK := parseProtocolTime(payload.ExpiresAt)
	locator, locatorOK := decodeExactBase64URL(payload.BundleLocator, 32)
	locatorMatches := locatorOK && bytes.Equal(locator, expected.locator[:])
	clear(locator)
	var testConfig testConfigWire
	_, testConfigErr := decodeStrictObject(payload.TestConfig, maximumPlaintextBytes, &testConfig, []string{"message", "sequence"})
	if payload.SchemaVersion != bundleSchema || !canonicalUUID(payload.BundleID) || !canonicalUUID(payload.Audience) ||
		payload.Audience != expected.audience || !locatorMatches || !versionOK || !issuedOK || !beforeOK || !expiresOK ||
		notBefore.Before(issuedAt) || !expiresAt.After(notBefore) || issuedAt.After(now.Add(skew)) ||
		notBefore.After(now.Add(skew)) || !expiresAt.After(now.Add(-skew)) ||
		testConfigErr != nil || len(testConfig.Message) > maximumMessageBytes || !utf8.ValidString(testConfig.Message) ||
		expiresAt.Sub(notBefore) != fixedBundleLifetime || !unsignedIntegerString(testConfig.Sequence) ||
		!validPolicy(payload.PolicySnapshot, issuedAt, expiresAt, now) {
		return 0, time.Time{}, time.Time{}, ErrPayload
	}
	return version, notBefore, expiresAt, nil
}

func metadataKeyAuthorizes(metadata rootMetadataWire, keyID string, notBefore, expiresAt time.Time) bool {
	for _, key := range metadata.SigningKeys {
		if key.KeyID != keyID || (key.State != "active" && key.State != "retiring") {
			continue
		}
		keyNotBefore, beforeOK := parseProtocolTime(key.NotBefore)
		keyNotAfter, afterOK := parseProtocolTime(key.NotAfter)
		return beforeOK && afterOK && !notBefore.Before(keyNotBefore) && !expiresAt.After(keyNotAfter)
	}
	return false
}

func validSignedBundle(value signedBundleWire) bool {
	_, digestOK := decodeExactBase64URL(value.PayloadSHA256, sha256.Size)
	_, signatureOK := decodeExactBase64URL(value.Signature, ed25519.SignatureSize)
	return len(value.PayloadJCS) >= 2 && len(value.PayloadJCS) <= maximumPlaintextBytes && digestOK &&
		validKeyID(value.SignerKeyID) && value.Algorithm == signatureAlgorithm && signatureOK && value.Padding == ""
}

func validPolicy(body []byte, issuedAt, bundleExpiresAt, now time.Time) bool {
	if len(body) < 2 || len(body) > maximumPolicyBytes {
		return false
	}
	type policyWire struct {
		Mode       string          `json:"mode"`
		ExpiresAt  json.RawMessage `json:"expires_at"`
		MaxDevices json.RawMessage `json:"max_devices"`
	}
	var policy policyWire
	keys, err := decodeStrictObject(body, maximumPolicyBytes, &policy, nil)
	if err != nil {
		return false
	}
	if policy.Mode == "standard" {
		return len(keys) == 1 && keys["mode"]
	}
	if policy.Mode != "trial_restricted" || len(keys) != 3 || !keys["mode"] || !keys["expires_at"] || !keys["max_devices"] {
		return false
	}
	var expiresAt string
	var maxDevices string
	if _, err := decodeStrictValue(policy.ExpiresAt, maximumPolicyBytes, &expiresAt); err != nil {
		return false
	}
	if _, err := decodeStrictValue(policy.MaxDevices, maximumPolicyBytes, &maxDevices); err != nil {
		return false
	}
	policyExpiresAt, validTime := parseProtocolTime(expiresAt)
	return validTime && maxDevices == "1" && policyExpiresAt.After(issuedAt) &&
		!policyExpiresAt.After(issuedAt.Add(fixedTrialGrace)) && !policyExpiresAt.After(bundleExpiresAt) && policyExpiresAt.After(now)
}

func removeValidatedPadding(plaintext []byte) ([]byte, bool) {
	validBucket := false
	for _, bucket := range plaintextBucketSizes {
		if len(plaintext) == bucket {
			validBucket = true
			break
		}
	}
	if !validBucket || len(plaintext) < plaintextLengthPrefix {
		return nil, false
	}
	length := int(uint32(plaintext[0])<<24 | uint32(plaintext[1])<<16 | uint32(plaintext[2])<<8 | uint32(plaintext[3]))
	if length < 2 || length > maximumPlaintextBytes || length > len(plaintext)-plaintextLengthPrefix {
		return nil, false
	}
	return append([]byte(nil), plaintext[plaintextLengthPrefix:plaintextLengthPrefix+length]...), true
}

func decodeCiphertext(value string) ([]byte, bool) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || base64.RawURLEncoding.EncodeToString(decoded) != value {
		clear(decoded)
		return nil, false
	}
	for _, bucket := range plaintextBucketSizes {
		if len(decoded) == bucket+chacha20Poly1305TagBytes {
			return decoded, true
		}
	}
	clear(decoded)
	return nil, false
}

func decodeExactBase64URL(value string, size int) ([]byte, bool) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) != size || base64.RawURLEncoding.EncodeToString(decoded) != value {
		clear(decoded)
		return nil, false
	}
	return decoded, true
}

func deriveKeyID(domain string, publicKey []byte) string {
	material := make([]byte, 0, len(domain)+len(publicKey))
	material = append(material, domain...)
	material = append(material, publicKey...)
	digest := sha256.Sum256(material)
	clear(material)
	return base64.RawURLEncoding.EncodeToString(digest[:16])
}

func parsePositiveUint(value string) (uint64, bool) {
	if !unsignedIntegerString(value) || value == "0" {
		return 0, false
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	return parsed, err == nil && parsed <= math.MaxInt64
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

func canonicalJSON(body []byte) ([]byte, error) {
	canonical, err := jcs.Transform(body)
	if err != nil {
		return nil, ErrEnvelope
	}
	return canonical, nil
}

func decodeStrictObject(body []byte, maximum int, target any, required []string) (map[string]bool, error) {
	keys, err := scanStrictJSON(body, maximum)
	if err != nil || keys == nil {
		return nil, ErrEnvelope
	}
	for _, name := range required {
		if !keys[name] {
			return nil, ErrEnvelope
		}
	}
	if len(required) > 0 && len(keys) != len(required) {
		return nil, ErrEnvelope
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return nil, ErrEnvelope
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, ErrEnvelope
	}
	return keys, nil
}

func decodeStrictValue(body []byte, maximum int, target any) (bool, error) {
	keys, err := scanStrictJSON(body, maximum)
	if err != nil || keys != nil {
		return false, ErrPayload
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return false, ErrPayload
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return false, ErrPayload
	}
	return true, nil
}

func scanStrictJSON(body []byte, maximum int) (map[string]bool, error) {
	if len(body) < 1 || len(body) > maximum || !utf8.Valid(body) || !validUnicodeEscapes(body) {
		return nil, ErrEnvelope
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	keys, err := scanJSONValue(decoder, 0, true)
	if err != nil {
		return nil, ErrEnvelope
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, ErrEnvelope
	}
	return keys, nil
}

func scanJSONValue(decoder *json.Decoder, depth int, top bool) (map[string]bool, error) {
	if depth > maximumJSONDepth {
		return nil, ErrEnvelope
	}
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	switch value := token.(type) {
	case json.Delim:
		switch value {
		case '{':
			seen := make(map[string]bool)
			for decoder.More() {
				nameToken, err := decoder.Token()
				name, ok := nameToken.(string)
				if err != nil || !ok || seen[name] {
					return nil, ErrEnvelope
				}
				seen[name] = true
				if _, err := scanJSONValue(decoder, depth+1, false); err != nil {
					return nil, err
				}
			}
			closing, err := decoder.Token()
			if err != nil || closing != json.Delim('}') {
				return nil, ErrEnvelope
			}
			if top {
				return seen, nil
			}
			return nil, nil
		case '[':
			for decoder.More() {
				if _, err := scanJSONValue(decoder, depth+1, false); err != nil {
					return nil, err
				}
			}
			closing, err := decoder.Token()
			if err != nil || closing != json.Delim(']') {
				return nil, ErrEnvelope
			}
			return nil, nil
		default:
			return nil, ErrEnvelope
		}
	case json.Number:
		parsed, err := strconv.ParseFloat(string(value), 64)
		if err != nil || math.IsInf(parsed, 0) || math.IsNaN(parsed) {
			return nil, ErrEnvelope
		}
	}
	return nil, nil
}

func validUnicodeEscapes(body []byte) bool {
	inString := false
	escaped := false
	for index := 0; index < len(body); index++ {
		character := body[index]
		if !inString {
			if character == '"' {
				inString = true
			}
			continue
		}
		if escaped {
			escaped = false
			if character != 'u' {
				continue
			}
			first, ok := decodeHexQuad(body, index+1)
			if !ok {
				return false
			}
			index += 4
			if first >= 0xdc00 && first <= 0xdfff {
				return false
			}
			if first < 0xd800 || first > 0xdbff {
				continue
			}
			if index+6 >= len(body) || body[index+1] != '\\' || body[index+2] != 'u' {
				return false
			}
			second, secondOK := decodeHexQuad(body, index+3)
			if !secondOK || second < 0xdc00 || second > 0xdfff || utf16.DecodeRune(rune(first), rune(second)) == utf8.RuneError {
				return false
			}
			index += 6
			continue
		}
		if character == '\\' {
			escaped = true
			continue
		}
		if character == '"' {
			inString = false
		}
	}
	return !inString && !escaped
}

func decodeHexQuad(body []byte, start int) (uint16, bool) {
	if start < 0 || start+4 > len(body) {
		return 0, false
	}
	var value uint16
	for _, character := range body[start : start+4] {
		value <<= 4
		switch {
		case character >= '0' && character <= '9':
			value += uint16(character - '0')
		case character >= 'a' && character <= 'f':
			value += uint16(character-'a') + 10
		case character >= 'A' && character <= 'F':
			value += uint16(character-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}

func isNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() { //nolint:exhaustive // Only nil-capable kinds matter.
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
