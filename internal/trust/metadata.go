package trust

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"time"
)

// ValidateRootMetadataV1 enforces the complete bounded v1 root/key set.
func ValidateRootMetadataV1(value RootMetadataV1) error {
	validFrom, validFromOK := parseProtocolTime(value.ValidFrom)
	validUntil, validUntilOK := parseProtocolTime(value.ValidUntil)
	if value.SchemaVersion != TrustMetadataSchemaV1 || !positiveIntegerString(value.Version) ||
		!validKeyID(value.RootKeyID) || value.RootAlgorithm != SignatureAlgorithm ||
		!validFromOK || !validUntilOK || !validUntil.After(validFrom) ||
		len(value.SigningKeys) == 0 || len(value.SigningKeys) > maximumSigningKeys {
		return ErrInvalidMetadata
	}
	seen := make(map[string]struct{}, len(value.SigningKeys))
	previous := ""
	for _, key := range value.SigningKeys {
		keyNotBefore, beforeOK := parseProtocolTime(key.NotBefore)
		keyNotAfter, afterOK := parseProtocolTime(key.NotAfter)
		publicKey, decodeErr := base64.RawURLEncoding.DecodeString(key.PublicKey)
		_, duplicate := seen[key.KeyID]
		validState := key.State == "future" || key.State == "active" || key.State == "retiring" || key.State == "revoked"
		if !validKeyID(key.KeyID) || key.Algorithm != SignatureAlgorithm || decodeErr != nil ||
			len(publicKey) != ed25519.PublicKeySize || encodeBase64URL(publicKey) != key.PublicKey ||
			key.KeyID != deriveKeyID(configKeyIDDomain, publicKey) || !validState || !beforeOK || !afterOK ||
			!keyNotAfter.After(keyNotBefore) || keyNotBefore.Before(validFrom) || keyNotAfter.After(validUntil) ||
			duplicate || previous >= key.KeyID && previous != "" {
			clear(publicKey)
			return ErrInvalidMetadata
		}
		clear(publicKey)
		seen[key.KeyID] = struct{}{}
		previous = key.KeyID
	}
	return nil
}

// CanonicalizeRootMetadataV1 validates and RFC 8785-canonicalizes metadata.
func CanonicalizeRootMetadataV1(value RootMetadataV1) ([]byte, error) {
	if err := ValidateRootMetadataV1(value); err != nil {
		return nil, err
	}
	body, err := json.Marshal(value)
	if err != nil {
		return nil, ErrInvalidMetadata
	}
	var decoded RootMetadataV1
	canonical, err := CanonicalizeJSON(body, &decoded)
	if err != nil || ValidateRootMetadataV1(decoded) != nil {
		return nil, ErrInvalidMetadata
	}
	return canonical, nil
}

// SignRootMetadataV1 signs exact canonical metadata with its declared root.
func SignRootMetadataV1(ctx context.Context, payload RootMetadataV1, signer ConfigSigner) (SignedRootMetadataV1, error) {
	boundedSigner, bounded := signer.(*TimeoutConfigSigner)
	if isNilValue(ctx) || !bounded || boundedSigner == nil || ctx.Err() != nil || payload.RootKeyID != boundedSigner.KeyID() ||
		payload.RootAlgorithm != SignatureAlgorithm {
		return SignedRootMetadataV1{}, ErrInvalidArgument
	}
	canonical, err := CanonicalizeRootMetadataV1(payload)
	if err != nil {
		return SignedRootMetadataV1{}, err
	}
	message := make([]byte, 0, len(metadataSignatureDomain)+len(canonical))
	message = append(message, metadataSignatureDomain...)
	message = append(message, canonical...)
	signature, err := boundedSigner.Sign(ctx, message)
	clear(message)
	if err != nil {
		return SignedRootMetadataV1{}, err
	}
	if len(signature) != ed25519.SignatureSize {
		clear(signature)
		return SignedRootMetadataV1{}, ErrInvalidSignature
	}
	result := SignedRootMetadataV1{PayloadJCS: string(canonical), Signature: encodeBase64URL(signature)}
	clear(signature)
	return result, nil
}

// VerifyRootMetadataV1 authenticates metadata and rejects rollback or invalid time.
func VerifyRootMetadataV1(
	value SignedRootMetadataV1,
	trustedRoots map[string]ed25519.PublicKey,
	highestTrusted uint64,
	now time.Time,
) (RootMetadataV1, error) {
	payload, version, err := verifyRootMetadataSignature(value, trustedRoots)
	if err != nil {
		return RootMetadataV1{}, err
	}
	if version <= highestTrusted {
		return RootMetadataV1{}, ErrMetadataRollback
	}
	validFrom, _ := parseProtocolTime(payload.ValidFrom)
	validUntil, _ := parseProtocolTime(payload.ValidUntil)
	if now.IsZero() || now.Location() != time.UTC || now.Before(validFrom) || !now.Before(validUntil) {
		return RootMetadataV1{}, ErrInvalidMetadata
	}
	return payload, nil
}

func verifyRootMetadataSignature(value SignedRootMetadataV1, trustedRoots map[string]ed25519.PublicKey) (RootMetadataV1, uint64, error) {
	if len(value.PayloadJCS) < 2 || len(value.PayloadJCS) > maximumCanonicalPayloadBytes || !exactBase64URL(value.Signature, ed25519.SignatureSize) {
		return RootMetadataV1{}, 0, ErrInvalidMetadata
	}
	var payload RootMetadataV1
	canonical, err := CanonicalizeJSON([]byte(value.PayloadJCS), &payload)
	if err != nil || !bytes.Equal(canonical, []byte(value.PayloadJCS)) || ValidateRootMetadataV1(payload) != nil {
		return RootMetadataV1{}, 0, ErrInvalidMetadata
	}
	publicKey, exists := trustedRoots[payload.RootKeyID]
	if !exists || len(publicKey) != ed25519.PublicKeySize || deriveKeyID(rootKeyIDDomain, publicKey) != payload.RootKeyID {
		return RootMetadataV1{}, 0, ErrUnknownRoot
	}
	signature, err := base64.RawURLEncoding.DecodeString(value.Signature)
	if err != nil {
		return RootMetadataV1{}, 0, ErrInvalidSignature
	}
	message := make([]byte, 0, len(metadataSignatureDomain)+len(canonical))
	message = append(message, metadataSignatureDomain...)
	message = append(message, canonical...)
	verified := ed25519.Verify(publicKey, message, signature)
	clear(message)
	clear(signature)
	if !verified {
		return RootMetadataV1{}, 0, ErrInvalidSignature
	}
	version, err := strconv.ParseUint(payload.Version, 10, 64)
	if err != nil {
		return RootMetadataV1{}, 0, ErrInvalidMetadata
	}
	return payload, version, nil
}

// SelectSigningKeyForIssue returns only an active key covering the full interval.
func SelectSigningKeyForIssue(payload RootMetadataV1, keyID string, notBefore, expiresAt time.Time) (ed25519.PublicKey, error) {
	return selectSigningKey(payload, keyID, notBefore, expiresAt, false)
}

// SelectSigningKeyForVerification permits active or retiring keys inside validity.
func SelectSigningKeyForVerification(payload RootMetadataV1, keyID string, notBefore, expiresAt time.Time) (ed25519.PublicKey, error) {
	return selectSigningKey(payload, keyID, notBefore, expiresAt, true)
}

func selectSigningKey(payload RootMetadataV1, keyID string, notBefore, expiresAt time.Time, allowRetiring bool) (ed25519.PublicKey, error) {
	if ValidateRootMetadataV1(payload) != nil || !validKeyID(keyID) || notBefore.IsZero() || expiresAt.IsZero() ||
		notBefore.Location() != time.UTC || expiresAt.Location() != time.UTC || !expiresAt.After(notBefore) {
		return nil, ErrKeyUnavailable
	}
	for _, key := range payload.SigningKeys {
		if key.KeyID != keyID {
			continue
		}
		if key.State != "active" && (!allowRetiring || key.State != "retiring") {
			return nil, ErrKeyUnavailable
		}
		keyNotBefore, _ := parseProtocolTime(key.NotBefore)
		keyNotAfter, _ := parseProtocolTime(key.NotAfter)
		if notBefore.Before(keyNotBefore) || expiresAt.After(keyNotAfter) {
			return nil, ErrKeyUnavailable
		}
		publicKey, err := base64.RawURLEncoding.DecodeString(key.PublicKey)
		if err != nil || len(publicKey) != ed25519.PublicKeySize {
			clear(publicKey)
			return nil, ErrKeyUnavailable
		}
		return append(ed25519.PublicKey(nil), publicKey...), nil
	}
	return nil, ErrKeyUnavailable
}

// PublishMetadata validates strict monotonicity before complete-set publication.
func PublishMetadata(
	ctx context.Context,
	repository MetadataRepository,
	trustedRoots map[string]ed25519.PublicKey,
	value SignedRootMetadataV1,
	now time.Time,
) error {
	if isNilValue(ctx) || isNilValue(repository) || ctx.Err() != nil {
		return ErrInvalidArgument
	}
	records, err := repository.List(ctx)
	if err != nil || len(records) > maximumMetadataVersions {
		return ErrRepository
	}
	var highest uint64
	for _, record := range records {
		_, version, verifyErr := verifyRootMetadataSignature(record, trustedRoots)
		if verifyErr != nil {
			return ErrRepository
		}
		if version > highest {
			highest = version
		}
	}
	if _, err := VerifyRootMetadataV1(value, trustedRoots, highest, now); err != nil {
		return err
	}
	if err := repository.Publish(ctx, value); err != nil {
		return ErrRepository
	}
	return nil
}

// EnsureLocalMetadata idempotently publishes deterministic v1 local/test
// metadata only from explicitly supplied fixture signers.
func EnsureLocalMetadata(
	ctx context.Context,
	repository MetadataRepository,
	root *LocalRootSigner,
	config *LocalConfigSigner,
	validFrom, validUntil time.Time,
) error {
	if isNilValue(ctx) || isNilValue(repository) || root == nil || config == nil || ctx.Err() != nil ||
		validFrom.Location() != time.UTC || validUntil.Location() != time.UTC || validFrom.Nanosecond() != 0 ||
		validUntil.Nanosecond() != 0 || !validUntil.After(validFrom) {
		return ErrInvalidArgument
	}
	records, err := repository.List(ctx)
	if err != nil {
		return ErrRepository
	}
	if len(records) != 0 {
		return nil
	}
	keys := []SigningKeyMetadataV1{{
		KeyID: config.KeyID(), Algorithm: SignatureAlgorithm, PublicKey: encodeBase64URL(config.PublicKey()),
		State: "active", NotBefore: validFrom.Format(time.RFC3339), NotAfter: validUntil.Format(time.RFC3339),
	}}
	sort.Slice(keys, func(left, right int) bool { return keys[left].KeyID < keys[right].KeyID })
	payload := RootMetadataV1{
		SchemaVersion: TrustMetadataSchemaV1, Version: "1", RootKeyID: root.KeyID(), RootAlgorithm: SignatureAlgorithm,
		ValidFrom: validFrom.Format(time.RFC3339), ValidUntil: validUntil.Format(time.RFC3339), SigningKeys: keys,
	}
	boundedRoot, err := NewTimeoutConfigSigner(root, localSignerTimeout)
	if err != nil {
		return ErrInvalidArgument
	}
	defer func() { _ = boundedRoot.Close() }()
	signed, err := SignRootMetadataV1(ctx, payload, boundedRoot)
	if err != nil {
		return err
	}
	publishErr := PublishMetadata(ctx, repository, trustedTestRoots(root), signed, validFrom)
	if publishErr == nil {
		return publishErr
	}
	if !errors.Is(publishErr, ErrRepository) && !errors.Is(publishErr, ErrMetadataRollback) {
		return publishErr
	}
	records, listErr := repository.List(ctx)
	if listErr == nil && len(records) == 1 && records[0] == signed {
		return nil
	}
	return ErrRepository
}

func trustedTestRoots(root *LocalRootSigner) map[string]ed25519.PublicKey {
	return map[string]ed25519.PublicKey{root.KeyID(): root.PublicKey()}
}
