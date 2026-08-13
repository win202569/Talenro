package trust

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/outbox"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

const (
	metadataRollbackTimeout = 2 * time.Second
	locatorProtectionDomain = "trust/bundle-locator/v1"
)

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

// PostgresRepository is the generated-query-only Task 16 transaction and
// immutable byte-store adapter.
type PostgresRepository struct {
	database  PGXMetadataDB
	protector sensitive.Protector
}

var _ Repository = (*PostgresRepository)(nil)

// PostgresByteStore is the protector-free immutable lookup adapter used by
// primary and mirror distribution processes.
type PostgresByteStore struct{ database store.DBTX }

var _ ByteStore = (*PostgresByteStore)(nil)

type postgresTrustTransaction struct {
	dbtx        pgx.Tx
	queries     *store.Queries
	idempotency idempotency.Repository
	outbox      outbox.Repository
	protector   sensitive.Protector
}

// NewPostgresRepository binds Task 16 to generated queries and one protected
// locator/idempotency provider.
func NewPostgresRepository(database PGXMetadataDB, protector sensitive.Protector) (*PostgresRepository, error) {
	if isNilValue(database) || isNilValue(protector) {
		return nil, ErrInvalidArgument
	}
	return &PostgresRepository{database: database, protector: protector}, nil
}

// NewPostgresByteStore binds immutable lookup to only the generated DBTX
// boundary; it intentionally accepts no locator protector or signer.
func NewPostgresByteStore(database store.DBTX) (*PostgresByteStore, error) {
	if isNilValue(database) {
		return nil, ErrInvalidArgument
	}
	return &PostgresByteStore{database: database}, nil
}

// WithinTransaction owns begin, commit, panic collapse, and bounded rollback.
func (repository *PostgresRepository) WithinTransaction(
	ctx context.Context,
	operation func(context.Context, Transaction) error,
) (resultErr error) {
	if repository == nil || isNilValue(repository.database) || isNilValue(repository.protector) ||
		isNilValue(ctx) || ctx.Err() != nil || operation == nil {
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
	bound, err := newPostgresTrustTransaction(transaction, repository.protector)
	if err != nil {
		return ErrRepository
	}
	if err := operation(ctx, bound); err != nil {
		return err
	}
	if err := transaction.Commit(ctx); err != nil {
		return ErrRepository
	}
	transaction = nil
	return nil
}

func newPostgresTrustTransaction(transaction pgx.Tx, protector sensitive.Protector) (*postgresTrustTransaction, error) {
	if isNilValue(transaction) || isNilValue(protector) {
		return nil, ErrRepository
	}
	idempotencyRepository, err := idempotency.New(transaction, protector)
	if err != nil {
		return nil, ErrRepository
	}
	outboxRepository, err := outbox.NewRepository(transaction)
	if err != nil {
		return nil, ErrRepository
	}
	return &postgresTrustTransaction{
		dbtx: transaction, queries: store.New(transaction), idempotency: idempotencyRepository,
		outbox: outboxRepository, protector: protector,
	}, nil
}

func (transaction *postgresTrustTransaction) DBTX() store.DBTX {
	if transaction == nil {
		return nil
	}
	return transaction.dbtx
}

func (transaction *postgresTrustTransaction) BeginIdempotency(
	ctx context.Context,
	scope idempotency.Scope,
	key string,
	canonical []byte,
	createdAt, expiresAt time.Time,
) (transactionIdempotency, error) {
	if transaction == nil || isNilValue(transaction.idempotency) {
		return transactionIdempotency{}, ErrRepository
	}
	record, outcome, err := transaction.idempotency.Begin(ctx, scope, key, canonical, createdAt, expiresAt)
	if err != nil {
		return transactionIdempotency{}, ErrRepository
	}
	value := transactionIdempotency{record: record, outcome: outcome, status: record.ResponseStatus()}
	if outcome == idempotency.Replay {
		body, ok := record.TakeResponseBody()
		if !ok {
			return transactionIdempotency{}, ErrRepository
		}
		value.response = body
	}
	return value, nil
}

func (transaction *postgresTrustTransaction) CompleteIdempotency(
	ctx context.Context,
	value transactionIdempotency,
	status int,
	body []byte,
) error {
	if transaction == nil || isNilValue(transaction.idempotency) {
		return ErrRepository
	}
	completed, err := transaction.idempotency.Complete(ctx, value.record, status, body)
	if err != nil {
		return ErrRepository
	}
	owned, ok := completed.TakeResponseBody()
	clear(owned)
	if !ok {
		return ErrRepository
	}
	return nil
}

func (transaction *postgresTrustTransaction) NextBundleVersion(
	ctx context.Context,
	authorizationID uuid.UUID,
	updatedAt time.Time,
) (uint64, error) {
	if transaction == nil || transaction.queries == nil || authorizationID == uuid.Nil || updatedAt.IsZero() {
		return 0, ErrRepository
	}
	version, err := transaction.queries.NextBundleVersion(ctx, store.NextBundleVersionParams{
		AuthorizationID: authorizationID, UpdatedAt: updatedAt,
	})
	if err != nil || version <= 0 {
		return 0, ErrRepository
	}
	return uint64(version), nil
}

func (transaction *postgresTrustTransaction) LatestBundle(
	ctx context.Context,
	authorizationID uuid.UUID,
) (storedBundle, bool, error) {
	if transaction == nil || transaction.queries == nil || authorizationID == uuid.Nil {
		return storedBundle{}, false, ErrRepository
	}
	row, err := transaction.queries.GetLatestBundleIssuance(ctx, authorizationID)
	defer clearStoredBundleRow(&row)
	if errors.Is(err, pgx.ErrNoRows) {
		return storedBundle{}, false, nil
	}
	if err != nil {
		return storedBundle{}, false, ErrRepository
	}
	bundle, err := storedBundleFromRow(row, transaction.protector)
	if err != nil {
		return storedBundle{}, false, err
	}
	return bundle, true, nil
}

func (transaction *postgresTrustTransaction) BundleByID(
	ctx context.Context,
	bundleID uuid.UUID,
) (storedBundle, bool, error) {
	if transaction == nil || transaction.queries == nil || bundleID == uuid.Nil {
		return storedBundle{}, false, ErrRepository
	}
	row, err := transaction.queries.GetBundleIssuance(ctx, bundleID)
	defer clearStoredBundleRow(&row)
	if errors.Is(err, pgx.ErrNoRows) {
		return storedBundle{}, false, nil
	}
	if err != nil {
		return storedBundle{}, false, ErrRepository
	}
	bundle, err := storedBundleFromRow(row, transaction.protector)
	if err != nil {
		return storedBundle{}, false, err
	}
	return bundle, true, nil
}

func (transaction *postgresTrustTransaction) InsertBundle(ctx context.Context, value bundleIssuance) error {
	if transaction == nil || transaction.queries == nil || isNilValue(transaction.protector) || !validBundleIssuance(value) {
		return ErrRepository
	}
	digest := LocatorDigest(value.Locator[:])
	if digest == [32]byte{} {
		return ErrRepository
	}
	protected, err := transaction.protector.Encrypt(locatorProtectionDomain, value.Locator[:])
	if err != nil || protected.KeyVersion == 0 || protected.KeyVersion > math.MaxInt32 ||
		len(protected.Ciphertext) < 29 || len(protected.Ciphertext) > 512 {
		clear(protected.Ciphertext)
		return ErrRepository
	}
	defer clear(protected.Ciphertext)
	bundleVersion := int64(value.BundleVersion) // #nosec G115 -- validBundleIssuance proves value is at most MaxInt64.
	params := store.InsertBundleIssuanceParams{
		ID: value.ID, AuthorizationID: value.AuthorizationID, BundleVersion: bundleVersion,
		LocatorHash: bytes.Clone(digest[:]), LocatorCiphertext: bytes.Clone(protected.Ciphertext),
		LocatorKeyVersion: int32(protected.KeyVersion), Envelope: bytes.Clone(value.Envelope),
		EnvelopeSha256: bytes.Clone(value.EnvelopeSHA256[:]), SignerKeyID: value.SignerKeyID,
		IssuedAt: value.IssuedAt, NotBefore: value.NotBefore, ExpiresAt: value.ExpiresAt,
	}
	defer clearInsertBundleParams(&params)
	if transaction.queries.InsertBundleIssuance(ctx, params) != nil {
		return ErrRepository
	}
	return nil
}

func (transaction *postgresTrustTransaction) InsertAcknowledgement(
	ctx context.Context,
	bundle storedBundle,
	acknowledgedAt time.Time,
) error {
	if transaction == nil || transaction.queries == nil || bundle.ID == uuid.Nil || bundle.AuthorizationID == uuid.Nil ||
		bundle.BundleVersion == 0 || bundle.BundleVersion > math.MaxInt64 || acknowledgedAt.IsZero() {
		return ErrRepository
	}
	if transaction.queries.InsertBundleAcknowledgement(ctx, store.InsertBundleAcknowledgementParams{
		BundleID: bundle.ID, AuthorizationID: bundle.AuthorizationID, BundleVersion: int64(bundle.BundleVersion),
		AcknowledgedAt: acknowledgedAt,
	}) != nil {
		return ErrRepository
	}
	return nil
}

func (transaction *postgresTrustTransaction) AppendEvent(ctx context.Context, envelope *eventsv1.EventEnvelope) error {
	if transaction == nil || isNilValue(transaction.outbox) || transaction.outbox.Append(ctx, envelope) != nil {
		return ErrRepository
	}
	return nil
}

// GetImmutableBundle reads exact stored bytes by only the public locator digest.
func (repository *PostgresByteStore) GetImmutableBundle(
	ctx context.Context,
	digest [32]byte,
	now time.Time,
) (ImmutableBundle, error) {
	if repository == nil || isNilValue(repository.database) || isNilValue(ctx) || ctx.Err() != nil ||
		digest == [32]byte{} || now.IsZero() {
		return ImmutableBundle{}, ErrRepository
	}
	params := store.GetBundleByLocatorHashParams{LocatorHash: bytes.Clone(digest[:]), ExpiresAt: now}
	defer clear(params.LocatorHash)
	row, err := store.New(repository.database).GetBundleByLocatorHash(ctx, params)
	defer clearStoredBundleRow(&row)
	if errors.Is(err, pgx.ErrNoRows) {
		return ImmutableBundle{}, ErrBundleNotFound
	}
	if err != nil || !validStoredRow(row) || !bytes.Equal(row.LocatorHash, digest[:]) {
		return ImmutableBundle{}, ErrRepository
	}
	if !row.ExpiresAt.After(now) {
		return ImmutableBundle{}, ErrBundleNotFound
	}
	var envelopeDigest [32]byte
	copy(envelopeDigest[:], row.EnvelopeSha256)
	return ImmutableBundle{Envelope: bytes.Clone(row.Envelope), EnvelopeSHA256: envelopeDigest}, nil
}

func storedBundleFromRow(row store.TrustBundleIssuance, protector sensitive.Protector) (storedBundle, error) {
	if isNilValue(protector) || !validStoredRow(row) {
		return storedBundle{}, ErrRepository
	}
	protected := sensitive.EncryptedField{
		KeyVersion: uint32(row.LocatorKeyVersion), // #nosec G115 -- validStoredRow proves the signed key version is positive.
		Ciphertext: bytes.Clone(row.LocatorCiphertext),
	}
	defer clear(protected.Ciphertext)
	plaintext, err := protector.Decrypt(locatorProtectionDomain, protected)
	if err != nil || len(plaintext) != 32 {
		clear(plaintext)
		return storedBundle{}, ErrRepository
	}
	digest := LocatorDigest(plaintext)
	if !bytes.Equal(digest[:], row.LocatorHash) {
		clear(plaintext)
		return storedBundle{}, ErrRepository
	}
	var envelopeDigest [32]byte
	copy(envelopeDigest[:], row.EnvelopeSha256)
	result := storedBundle{
		ID: row.ID, AuthorizationID: row.AuthorizationID,
		BundleVersion: uint64(row.BundleVersion), // #nosec G115 -- validStoredRow proves the signed version is positive.
		Locator:       secret.NewBytes(plaintext), EnvelopeSHA256: envelopeDigest,
		IssuedAt: row.IssuedAt.UTC(), NotBefore: row.NotBefore.UTC(), ExpiresAt: row.ExpiresAt.UTC(),
	}
	clear(plaintext)
	return result, nil
}

func validBundleIssuance(value bundleIssuance) bool {
	return value.ID != uuid.Nil && value.AuthorizationID != uuid.Nil && value.BundleVersion > 0 && value.BundleVersion <= math.MaxInt64 &&
		value.Locator != [32]byte{} && len(value.Envelope) >= 64 && len(value.Envelope) <= maximumEnvelopeBytes &&
		sha256.Sum256(value.Envelope) == value.EnvelopeSHA256 && validKeyID(value.SignerKeyID) &&
		!value.IssuedAt.IsZero() && !value.NotBefore.IsZero() && !value.ExpiresAt.IsZero() &&
		!value.NotBefore.Before(value.IssuedAt) && value.ExpiresAt.After(value.NotBefore)
}

func validStoredRow(row store.TrustBundleIssuance) bool {
	envelopeDigest := sha256.Sum256(row.Envelope)
	return row.ID != uuid.Nil && row.AuthorizationID != uuid.Nil && row.BundleVersion > 0 &&
		len(row.LocatorHash) == 32 && len(row.LocatorCiphertext) >= 29 && len(row.LocatorCiphertext) <= 512 &&
		row.LocatorKeyVersion > 0 && len(row.Envelope) >= 64 && len(row.Envelope) <= maximumEnvelopeBytes &&
		len(row.EnvelopeSha256) == 32 && bytes.Equal(row.EnvelopeSha256, envelopeDigest[:]) &&
		validKeyID(row.SignerKeyID) && !row.IssuedAt.IsZero() && !row.NotBefore.IsZero() && !row.ExpiresAt.IsZero() &&
		!row.NotBefore.Before(row.IssuedAt) && row.ExpiresAt.After(row.NotBefore)
}

func clearInsertBundleParams(params *store.InsertBundleIssuanceParams) {
	if params == nil {
		return
	}
	clear(params.LocatorHash)
	clear(params.LocatorCiphertext)
	clear(params.Envelope)
	clear(params.EnvelopeSha256)
}

func clearStoredBundleRow(row *store.TrustBundleIssuance) {
	if row == nil {
		return
	}
	clear(row.LocatorHash)
	clear(row.LocatorCiphertext)
	clear(row.Envelope)
	clear(row.EnvelopeSha256)
	row.LocatorHash = nil
	row.LocatorCiphertext = nil
	row.Envelope = nil
	row.EnvelopeSha256 = nil
}

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
