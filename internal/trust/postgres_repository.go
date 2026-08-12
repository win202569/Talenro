package trust

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"talenro.local/platform/internal/store"
)

const metadataRollbackTimeout = 2 * time.Second

// PGXMetadataDB is the narrow pgx pool boundary used for generated trust queries.
type PGXMetadataDB interface {
	store.DBTX
	Begin(context.Context) (pgx.Tx, error)
}

// PostgresMetadataRepository atomically persists root metadata and its complete key set.
type PostgresMetadataRepository struct {
	database     PGXMetadataDB
	trustedRoots map[string]ed25519.PublicKey
	now          func() time.Time
}

var _ MetadataRepository = (*PostgresMetadataRepository)(nil)

// NewPostgresMetadataRepository copies provisioned roots into a generated-query repository.
func NewPostgresMetadataRepository(database PGXMetadataDB, trustedRoots map[string]ed25519.PublicKey) (*PostgresMetadataRepository, error) {
	if isNilValue(database) || len(trustedRoots) == 0 || len(trustedRoots) > maximumSigningKeys {
		return nil, ErrInvalidArgument
	}
	ownedRoots := make(map[string]ed25519.PublicKey, len(trustedRoots))
	for keyID, publicKey := range trustedRoots {
		if !validKeyID(keyID) || len(publicKey) != ed25519.PublicKeySize || deriveKeyID(rootKeyIDDomain, publicKey) != keyID {
			return nil, ErrInvalidArgument
		}
		ownedRoots[keyID] = append(ed25519.PublicKey(nil), publicKey...)
	}
	return &PostgresMetadataRepository{database: database, trustedRoots: ownedRoots, now: time.Now}, nil
}

// List reconstructs and authenticates all complete stored metadata versions.
func (repository *PostgresMetadataRepository) List(ctx context.Context) (result []SignedRootMetadataV1, resultErr error) {
	defer func() {
		if recover() != nil {
			result = nil
			resultErr = ErrRepository
		}
	}()
	if repository == nil || isNilValue(repository.database) || isNilValue(ctx) || ctx.Err() != nil {
		return nil, ErrRepository
	}
	return listStoredMetadata(ctx, store.New(repository.database), repository.trustedRoots)
}

// Publish verifies and atomically inserts one root plus its complete key set.
func (repository *PostgresMetadataRepository) Publish(ctx context.Context, value SignedRootMetadataV1) (resultErr error) {
	if repository == nil || isNilValue(repository.database) || isNilValue(ctx) || ctx.Err() != nil {
		return ErrRepository
	}
	var transaction pgx.Tx
	defer func() {
		if recover() != nil {
			resultErr = ErrRepository
		}
		if !isNilValue(transaction) {
			rollbackMetadataTransaction(ctx, transaction)
		}
	}()
	var err error
	transaction, err = repository.database.Begin(ctx)
	if err != nil || isNilValue(transaction) {
		return ErrRepository
	}
	queries := store.New(transaction)
	existing, err := listStoredMetadata(ctx, queries, repository.trustedRoots)
	if err != nil {
		return ErrRepository
	}
	payload, version, err := verifyRootMetadataSignature(value, repository.trustedRoots)
	if err != nil {
		return err
	}
	var highest uint64
	for _, record := range existing {
		_, rowVersion, verifyErr := verifyRootMetadataSignature(record, repository.trustedRoots)
		if verifyErr != nil {
			return ErrRepository
		}
		if rowVersion > highest {
			highest = rowVersion
		}
	}
	if version <= highest {
		return ErrMetadataRollback
	}
	signature, err := base64.RawURLEncoding.DecodeString(value.Signature)
	if err != nil || len(signature) != ed25519.SignatureSize {
		clear(signature)
		return ErrInvalidSignature
	}
	validFrom, _ := parseProtocolTime(payload.ValidFrom)
	validUntil, _ := parseProtocolTime(payload.ValidUntil)
	// #nosec G115 -- metadata versions are validated against MaxInt64.
	versionInt64 := int64(version)
	if err := queries.InsertTrustRootMetadata(ctx, store.InsertTrustRootMetadataParams{
		Version: versionInt64, CanonicalPayload: []byte(value.PayloadJCS), Signature: bytes.Clone(signature),
		ValidFrom: validFrom, ValidUntil: validUntil, CreatedAt: repository.now().UTC(),
	}); err != nil {
		clear(signature)
		return ErrRepository
	}
	clear(signature)
	for _, key := range payload.SigningKeys {
		publicKey, decodeErr := base64.RawURLEncoding.DecodeString(key.PublicKey)
		notBefore, beforeOK := parseProtocolTime(key.NotBefore)
		notAfter, afterOK := parseProtocolTime(key.NotAfter)
		if decodeErr != nil || len(publicKey) != ed25519.PublicKeySize || !beforeOK || !afterOK {
			clear(publicKey)
			return ErrInvalidMetadata
		}
		err = queries.InsertSigningKeyMetadata(ctx, store.InsertSigningKeyMetadataParams{
			KeyID: key.KeyID, RootMetadataVersion: versionInt64, PublicKey: bytes.Clone(publicKey),
			State: key.State, NotBefore: notBefore, NotAfter: notAfter,
		})
		clear(publicKey)
		if err != nil {
			return ErrRepository
		}
	}
	if err := transaction.Commit(ctx); err != nil {
		return ErrRepository
	}
	return nil
}

func listStoredMetadata(
	ctx context.Context,
	queries *store.Queries,
	trustedRoots map[string]ed25519.PublicKey,
) ([]SignedRootMetadataV1, error) {
	if queries == nil {
		return nil, ErrRepository
	}
	roots, err := queries.ListTrustRootMetadata(ctx)
	if err != nil || len(roots) > maximumMetadataVersions {
		return nil, ErrRepository
	}
	result := make([]SignedRootMetadataV1, 0, len(roots))
	for _, root := range roots {
		if root.Version <= 0 || len(root.CanonicalPayload) < 2 || len(root.CanonicalPayload) > maximumCanonicalPayloadBytes ||
			len(root.Signature) != ed25519.SignatureSize || root.ValidFrom.IsZero() || root.ValidUntil.IsZero() || root.CreatedAt.IsZero() {
			return nil, ErrRepository
		}
		root.ValidFrom = root.ValidFrom.UTC()
		root.ValidUntil = root.ValidUntil.UTC()
		root.CreatedAt = root.CreatedAt.UTC()
		record := SignedRootMetadataV1{PayloadJCS: string(bytes.Clone(root.CanonicalPayload)), Signature: encodeBase64URL(root.Signature)}
		payload, verifiedVersion, verifyErr := verifyRootMetadataSignature(record, trustedRoots)
		version, versionErr := strconv.ParseInt(payload.Version, 10, 64)
		if verifyErr != nil || versionErr != nil || version != root.Version || verifiedVersion != uint64(root.Version) {
			return nil, ErrRepository
		}
		validFrom, _ := parseProtocolTime(payload.ValidFrom)
		validUntil, _ := parseProtocolTime(payload.ValidUntil)
		if !root.ValidFrom.Equal(validFrom) || !root.ValidUntil.Equal(validUntil) {
			return nil, ErrRepository
		}
		keys, keysErr := queries.ListSigningKeysForMetadata(ctx, root.Version)
		if keysErr != nil || !storedKeysMatch(root.Version, payload.SigningKeys, keys) {
			return nil, ErrRepository
		}
		result = append(result, record)
	}
	return result, nil
}

func storedKeysMatch(rootVersion int64, expected []SigningKeyMetadataV1, rows []store.TrustSigningKeyMetadatum) bool {
	if len(expected) != len(rows) {
		return false
	}
	for index, key := range expected {
		row := rows[index]
		if row.NotBefore.IsZero() || row.NotAfter.IsZero() {
			return false
		}
		row.NotBefore = row.NotBefore.UTC()
		row.NotAfter = row.NotAfter.UTC()
		publicKey, err := base64.RawURLEncoding.DecodeString(key.PublicKey)
		notBefore, beforeOK := parseProtocolTime(key.NotBefore)
		notAfter, afterOK := parseProtocolTime(key.NotAfter)
		matches := err == nil && beforeOK && afterOK && row.RootMetadataVersion == rootVersion &&
			row.KeyID == key.KeyID && row.Algorithm == key.Algorithm &&
			bytes.Equal(row.PublicKey, publicKey) && row.State == key.State && row.NotBefore.Equal(notBefore) && row.NotAfter.Equal(notAfter)
		clear(publicKey)
		if !matches {
			return false
		}
	}
	return true
}

func rollbackMetadataTransaction(operationContext context.Context, transaction pgx.Tx) {
	defer func() { _ = recover() }()
	base := context.WithoutCancel(operationContext)
	ctx, cancel := context.WithTimeout(base, metadataRollbackTimeout)
	defer cancel()
	_ = transaction.Rollback(ctx)
}
