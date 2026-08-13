package trust

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

func TestPostgresRepositoryPersistsOnlyProtectedLocatorAndImmutableEnvelope(t *testing.T) {
	protector, err := sensitive.NewLocal(
		secret.NewBytes(bytes.Repeat([]byte{0x81}, 32)),
		secret.NewBytes(bytes.Repeat([]byte{0x82}, 32)),
		7,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = protector.Close() })
	database := &task16RepositoryDB{}
	repository, err := NewPostgresRepository(database, protector)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.August, 13, 8, 0, 0, 0, time.UTC)
	locator := [32]byte{1, 2, 3, 4, 5, 6, 7, 8}
	envelope := bytes.Repeat([]byte{0x45}, 96)
	envelopeDigest := sha256.Sum256(envelope)
	issuance := bundleIssuance{
		ID:              uuid.MustParse("dddddddd-dddd-4ddd-8ddd-dddddddddddd"),
		AuthorizationID: uuid.MustParse("eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee"),
		BundleVersion:   1,
		Locator:         locator,
		Envelope:        bytes.Clone(envelope),
		EnvelopeSHA256:  envelopeDigest,
		SignerKeyID:     "abcdefghijklmnop",
		IssuedAt:        now, NotBefore: now, ExpiresAt: now.Add(24 * time.Hour),
	}

	err = repository.WithinTransaction(context.Background(), func(ctx context.Context, transaction Transaction) error {
		return transaction.InsertBundle(ctx, issuance)
	})
	if err != nil {
		t.Fatalf("WithinTransaction: %v", err)
	}
	params := database.transaction.insert
	if bytes.Contains(params.LocatorCiphertext, locator[:]) || bytes.Equal(params.LocatorHash, locator[:]) {
		t.Fatal("raw locator was persisted")
	}
	wantDigest := LocatorDigest(locator[:])
	if !bytes.Equal(params.LocatorHash, wantDigest[:]) || params.LocatorKeyVersion != 7 {
		t.Fatal("locator digest or protection metadata is incorrect")
	}
	if !bytes.Equal(params.Envelope, envelope) || !bytes.Equal(params.EnvelopeSha256, envelopeDigest[:]) {
		t.Fatal("immutable envelope bytes or digest changed before persistence")
	}
	plaintext, err := protector.Decrypt(locatorProtectionDomain, sensitive.EncryptedField{
		KeyVersion: uint32(params.LocatorKeyVersion), // #nosec G115 -- the assertion above proves the fixture version is exactly 7.
		Ciphertext: bytes.Clone(params.LocatorCiphertext),
	})
	if err != nil || !bytes.Equal(plaintext, locator[:]) {
		clear(plaintext)
		t.Fatal("protected locator is not recoverable under the locator domain")
	}
	clear(plaintext)
}

func TestPostgresRepositoryCollapsesExactlyExpiredImmutableBundleToNotFound(t *testing.T) {
	now := time.Date(2026, time.August, 13, 8, 0, 0, 0, time.UTC)
	locator := bytes.Repeat([]byte{0x93}, 32)
	digest := LocatorDigest(locator)
	envelope := bytes.Repeat([]byte{0x94}, 96)
	envelopeDigest := sha256.Sum256(envelope)
	database := &task16RepositoryDB{immutable: store.TrustBundleIssuance{
		ID:              uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		AuthorizationID: uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		BundleVersion:   1, LocatorHash: bytes.Clone(digest[:]), LocatorCiphertext: bytes.Repeat([]byte{0x95}, 32),
		LocatorKeyVersion: 8, Envelope: envelope, EnvelopeSha256: bytes.Clone(envelopeDigest[:]),
		SignerKeyID: "abcdefghijklmnop", IssuedAt: now.Add(-24 * time.Hour), NotBefore: now.Add(-24 * time.Hour), ExpiresAt: now,
	}}
	byteStore, err := NewPostgresByteStore(database)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := byteStore.GetImmutableBundle(context.Background(), digest, now); !errors.Is(err, ErrBundleNotFound) {
		t.Fatalf("exactly expired bundle error = %v, want ErrBundleNotFound", err)
	}
}

func TestPostgresByteStoreRequiresOnlyDatabaseDependency(t *testing.T) {
	now := time.Date(2026, time.August, 13, 8, 0, 0, 0, time.UTC)
	locator := bytes.Repeat([]byte{0xa1}, 32)
	digest := LocatorDigest(locator)
	envelope := bytes.Repeat([]byte{0xa2}, 96)
	envelopeDigest := sha256.Sum256(envelope)
	database := &task16RepositoryDB{immutable: store.TrustBundleIssuance{
		ID:              uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		AuthorizationID: uuid.MustParse("44444444-4444-4444-8444-444444444444"),
		BundleVersion:   1, LocatorHash: bytes.Clone(digest[:]), LocatorCiphertext: bytes.Repeat([]byte{0xa3}, 32),
		LocatorKeyVersion: 9, Envelope: bytes.Clone(envelope), EnvelopeSha256: bytes.Clone(envelopeDigest[:]),
		SignerKeyID: "abcdefghijklmnop", IssuedAt: now, NotBefore: now, ExpiresAt: now.Add(24 * time.Hour),
	}}
	byteStore, err := NewPostgresByteStore(database)
	if err != nil {
		t.Fatalf("NewPostgresByteStore without protector: %v", err)
	}
	var dependency ByteStore = byteStore
	_ = dependency
	if reflect.TypeOf((*PostgresRepository)(nil)).Implements(reflect.TypeFor[ByteStore]()) {
		t.Fatal("transaction repository still transports protector into the mirror ByteStore boundary")
	}
	bundle, err := byteStore.GetImmutableBundle(context.Background(), digest, now)
	if err != nil {
		t.Fatalf("GetImmutableBundle without protector: %v", err)
	}
	defer bundle.Clear()
	if !bytes.Equal(bundle.Envelope, envelope) || bundle.EnvelopeSHA256 != envelopeDigest {
		t.Fatal("database-only immutable byte store changed published bytes")
	}
}

type task16RepositoryDB struct {
	transaction task16RepositoryTx
	immutable   store.TrustBundleIssuance
}

func (database *task16RepositoryDB) Begin(context.Context) (pgx.Tx, error) {
	return &database.transaction, nil
}

func (*task16RepositoryDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected pool Exec")
}

func (*task16RepositoryDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected pool Query")
}

func (database *task16RepositoryDB) QueryRow(_ context.Context, query string, _ ...any) pgx.Row {
	if bytes.Contains([]byte(query), []byte("WHERE locator_hash = $1")) {
		return task16BundleRow{value: database.immutable}
	}
	return task16ErrorRow{}
}

type task16BundleRow struct{ value store.TrustBundleIssuance }

func (row task16BundleRow) Scan(destinations ...any) error {
	if len(destinations) != 12 {
		return errors.New("unexpected bundle scan")
	}
	*destinations[0].(*uuid.UUID) = row.value.ID
	*destinations[1].(*uuid.UUID) = row.value.AuthorizationID
	*destinations[2].(*int64) = row.value.BundleVersion
	*destinations[3].(*[]byte) = bytes.Clone(row.value.LocatorHash)
	*destinations[4].(*[]byte) = bytes.Clone(row.value.LocatorCiphertext)
	*destinations[5].(*int32) = row.value.LocatorKeyVersion
	*destinations[6].(*[]byte) = bytes.Clone(row.value.Envelope)
	*destinations[7].(*[]byte) = bytes.Clone(row.value.EnvelopeSha256)
	*destinations[8].(*string) = row.value.SignerKeyID
	*destinations[9].(*time.Time) = row.value.IssuedAt
	*destinations[10].(*time.Time) = row.value.NotBefore
	*destinations[11].(*time.Time) = row.value.ExpiresAt
	return nil
}

type task16RepositoryTx struct {
	insert     store.InsertBundleIssuanceParams
	committed  bool
	rolledBack bool
}

func (transaction *task16RepositoryTx) Begin(context.Context) (pgx.Tx, error) {
	return nil, errors.New("unexpected nested Begin")
}
func (*task16RepositoryTx) Config() *pgx.ConnConfig { return nil }
func (*task16RepositoryTx) Conn() *pgx.Conn         { return nil }
func (*task16RepositoryTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, errors.New("unexpected CopyFrom")
}
func (*task16RepositoryTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }
func (*task16RepositoryTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, errors.New("unexpected Prepare")
}
func (*task16RepositoryTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (*task16RepositoryTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}
func (*task16RepositoryTx) QueryRow(context.Context, string, ...any) pgx.Row { return task16ErrorRow{} }

func (transaction *task16RepositoryTx) Exec(_ context.Context, query string, arguments ...any) (pgconn.CommandTag, error) {
	if !bytes.Contains([]byte(query), []byte("INSERT INTO trust.bundle_issuances")) || len(arguments) != 12 {
		return pgconn.CommandTag{}, errors.New("unexpected Exec")
	}
	transaction.insert = store.InsertBundleIssuanceParams{
		ID: arguments[0].(uuid.UUID), AuthorizationID: arguments[1].(uuid.UUID), BundleVersion: arguments[2].(int64),
		LocatorHash: bytes.Clone(arguments[3].([]byte)), LocatorCiphertext: bytes.Clone(arguments[4].([]byte)),
		LocatorKeyVersion: arguments[5].(int32), Envelope: bytes.Clone(arguments[6].([]byte)),
		EnvelopeSha256: bytes.Clone(arguments[7].([]byte)), SignerKeyID: arguments[8].(string),
		IssuedAt: arguments[9].(time.Time), NotBefore: arguments[10].(time.Time), ExpiresAt: arguments[11].(time.Time),
	}
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

func (transaction *task16RepositoryTx) Commit(context.Context) error {
	transaction.committed = true
	return nil
}
func (transaction *task16RepositoryTx) Rollback(context.Context) error {
	transaction.rolledBack = true
	return nil
}

var _ pgx.Tx = (*task16RepositoryTx)(nil)
