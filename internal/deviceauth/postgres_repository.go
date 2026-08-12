package deviceauth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/outbox"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

const deviceauthRollbackTimeout = 2 * time.Second

// PGXBeginner is the narrow production transaction source used by deviceauth.
type PGXBeginner interface {
	Begin(context.Context) (pgx.Tx, error)
}

// PostgresRepository binds Task 12 operations to generated queries.
type PostgresRepository struct {
	beginner  PGXBeginner
	protector sensitive.Protector
}

var _ Repository = (*PostgresRepository)(nil)
var _ identity.DeviceAuthorizationParticipant = (*PostgresRepository)(nil)

// NewPostgresRepository creates a generated-query-only deviceauth repository.
func NewPostgresRepository(beginner PGXBeginner, protector sensitive.Protector) (*PostgresRepository, error) {
	if nilDeviceauthValue(beginner) || nilDeviceauthValue(protector) {
		return nil, ErrRepository
	}
	return &PostgresRepository{beginner: beginner, protector: protector}, nil
}

// WithinTransaction owns begin, commit, panic recovery, and independent bounded rollback.
func (repository *PostgresRepository) WithinTransaction(
	ctx context.Context,
	operation func(context.Context, Transaction) error,
) (resultErr error) {
	if repository == nil || nilDeviceauthValue(ctx) || nilDeviceauthValue(repository.beginner) ||
		nilDeviceauthValue(repository.protector) || operation == nil || ctx.Err() != nil {
		return ErrRepository
	}
	var transaction pgx.Tx
	defer func() {
		if recover() != nil {
			resultErr = ErrRepository
		}
		if !nilDeviceauthValue(transaction) {
			rollbackDeviceauthTransaction(ctx, transaction)
		}
	}()
	var err error
	transaction, err = repository.beginner.Begin(ctx)
	if err != nil || nilDeviceauthValue(transaction) {
		return ErrRepository
	}
	bound, err := newPostgresTransaction(transaction, repository.protector)
	if err != nil {
		return ErrRepository
	}
	if err := operation(ctx, bound); err != nil {
		return err
	}
	if err := transaction.Commit(ctx); err != nil {
		return ErrRepository
	}
	return nil
}

// ActivateVerifiedPrincipal performs only the generated provisional-authorization transition.
func (repository *PostgresRepository) ActivateVerifiedPrincipal(
	ctx context.Context,
	dbtx store.DBTX,
	principalID identity.PrincipalID,
	now time.Time,
) (resultErr error) {
	defer func() {
		if recover() != nil {
			resultErr = ErrRepository
		}
	}()
	parsed, err := uuid.Parse(string(principalID))
	if repository == nil || nilDeviceauthValue(ctx) || nilDeviceauthValue(dbtx) || err != nil || parsed == uuid.Nil ||
		parsed.String() != string(principalID) || now.IsZero() || ctx.Err() != nil {
		return ErrRepository
	}
	rows, err := store.New(dbtx).ActivateProvisionalAuthorization(ctx, store.ActivateProvisionalAuthorizationParams{
		PrincipalID: parsed, UpdatedAt: now,
	})
	if err != nil || rows < 0 {
		return ErrRepository
	}
	return nil
}

type postgresTransaction struct {
	tx          pgx.Tx
	queries     *store.Queries
	idempotency idempotency.Repository
	outbox      outbox.Repository
}

var _ Transaction = (*postgresTransaction)(nil)
var _ deviceTokenTransaction = (*postgresTransaction)(nil)

func newPostgresTransaction(transaction pgx.Tx, protector sensitive.Protector) (*postgresTransaction, error) {
	if nilDeviceauthValue(transaction) || nilDeviceauthValue(protector) {
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
	return &postgresTransaction{
		tx: transaction, queries: store.New(transaction), idempotency: idempotencyRepository, outbox: outboxRepository,
	}, nil
}

func (transaction *postgresTransaction) DBTX() store.DBTX { return transaction.tx }

func (transaction *postgresTransaction) BeginIdempotency(
	ctx context.Context,
	scope idempotency.Scope,
	key string,
	canonical []byte,
	createdAt, expiresAt time.Time,
) (idempotency.Record, idempotency.Outcome, error) {
	if transaction == nil || nilDeviceauthValue(transaction.idempotency) {
		return idempotency.Record{}, "", ErrRepository
	}
	record, outcome, err := transaction.idempotency.Begin(ctx, scope, key, canonical, createdAt, expiresAt)
	if err != nil {
		return idempotency.Record{}, "", ErrRepository
	}
	return record, outcome, nil
}

func (transaction *postgresTransaction) CompleteIdempotency(ctx context.Context, record idempotency.Record, status int, body []byte) error {
	if transaction == nil || nilDeviceauthValue(transaction.idempotency) {
		return ErrRepository
	}
	completed, err := transaction.idempotency.Complete(ctx, record, status, body)
	if err != nil {
		return ErrRepository
	}
	owned, ok := completed.TakeResponseBody()
	defer clear(owned)
	if !ok {
		return ErrRepository
	}
	return nil
}

func (transaction *postgresTransaction) ConsumeEnrollmentGrant(
	ctx context.Context,
	digest [32]byte,
	deviceID uuid.UUID,
	now time.Time,
) (store.DeviceauthEnrollmentGrant, bool, error) {
	if transaction == nil || transaction.queries == nil || digest == [32]byte{} || deviceID == uuid.Nil || now.IsZero() {
		return store.DeviceauthEnrollmentGrant{}, false, ErrRepository
	}
	digestCopy := append([]byte(nil), digest[:]...)
	defer clear(digestCopy)
	row, err := transaction.queries.ConsumeEnrollmentGrant(ctx, store.ConsumeEnrollmentGrantParams{
		TokenHash: digestCopy, ConsumedDeviceID: uuid.NullUUID{UUID: deviceID, Valid: true},
		ConsumedAt: sql.NullTime{Time: now, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return store.DeviceauthEnrollmentGrant{}, false, nil
	}
	if err != nil {
		return store.DeviceauthEnrollmentGrant{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) CreateDevice(ctx context.Context, params store.CreateDeviceParams) error {
	if transaction == nil || transaction.queries == nil || transaction.queries.CreateDevice(ctx, params) != nil {
		return ErrRepository
	}
	return nil
}

func (transaction *postgresTransaction) CreateDeviceAuthorization(ctx context.Context, params store.CreateDeviceAuthorizationParams) error {
	if transaction == nil || transaction.queries == nil || transaction.queries.CreateDeviceAuthorization(ctx, params) != nil {
		return ErrRepository
	}
	return nil
}

func (transaction *postgresTransaction) CreateDevicePolicySnapshot(ctx context.Context, params store.CreateDevicePolicySnapshotParams) error {
	if transaction == nil || transaction.queries == nil || transaction.queries.CreateDevicePolicySnapshot(ctx, params) != nil {
		return ErrRepository
	}
	return nil
}

func (transaction *postgresTransaction) CreateDeviceTokenFamily(ctx context.Context, params store.CreateDeviceTokenFamilyParams) error {
	if transaction == nil || transaction.queries == nil || transaction.queries.CreateDeviceTokenFamily(ctx, params) != nil {
		return ErrRepository
	}
	return nil
}

func (transaction *postgresTransaction) InsertDeviceRefreshToken(ctx context.Context, params store.InsertDeviceRefreshTokenParams) error {
	if transaction == nil || transaction.queries == nil || transaction.queries.InsertDeviceRefreshToken(ctx, params) != nil {
		return ErrRepository
	}
	return nil
}

func (transaction *postgresTransaction) AppendEvent(ctx context.Context, envelope *eventsv1.EventEnvelope) error {
	if transaction == nil || nilDeviceauthValue(transaction.outbox) || transaction.outbox.Append(ctx, envelope) != nil {
		return ErrRepository
	}
	return nil
}

func (transaction *postgresTransaction) DiscoverDeviceRefreshToken(ctx context.Context, digest []byte) (store.DiscoverDeviceRefreshTokenRow, bool, error) {
	if transaction == nil || transaction.queries == nil || len(digest) != 32 {
		return store.DiscoverDeviceRefreshTokenRow{}, false, ErrRepository
	}
	owned := append([]byte(nil), digest...)
	defer clear(owned)
	row, err := transaction.queries.DiscoverDeviceRefreshToken(ctx, owned)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.DiscoverDeviceRefreshTokenRow{}, false, nil
	}
	if err != nil {
		return store.DiscoverDeviceRefreshTokenRow{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) LockDeviceFamilyRefreshTokens(ctx context.Context, familyID uuid.UUID) ([]store.DeviceauthDeviceRefreshToken, error) {
	if transaction == nil || transaction.queries == nil || familyID == uuid.Nil {
		return nil, ErrRepository
	}
	rows, err := transaction.queries.LockDeviceFamilyRefreshTokens(ctx, familyID)
	if err != nil {
		return nil, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) ListDeviceFamilyRefreshTokens(ctx context.Context, familyID uuid.UUID) ([]store.DeviceauthDeviceRefreshToken, error) {
	if transaction == nil || transaction.queries == nil || familyID == uuid.Nil {
		return nil, ErrRepository
	}
	rows, err := transaction.queries.ListDeviceFamilyRefreshTokens(ctx, familyID)
	if err != nil {
		return nil, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) ListDeviceAuthorizationFamilies(ctx context.Context, authorizationID uuid.UUID) ([]store.DeviceauthDeviceTokenFamily, error) {
	if transaction == nil || transaction.queries == nil || authorizationID == uuid.Nil {
		return nil, ErrRepository
	}
	rows, err := transaction.queries.ListDeviceAuthorizationFamilies(ctx, authorizationID)
	if err != nil {
		return nil, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) GetDeviceTokenFamilyForUpdate(ctx context.Context, familyID uuid.UUID) (store.DeviceauthDeviceTokenFamily, bool, error) {
	if transaction == nil || transaction.queries == nil || familyID == uuid.Nil {
		return store.DeviceauthDeviceTokenFamily{}, false, ErrRepository
	}
	row, err := transaction.queries.GetDeviceTokenFamilyForUpdate(ctx, familyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.DeviceauthDeviceTokenFamily{}, false, nil
	}
	if err != nil {
		return store.DeviceauthDeviceTokenFamily{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) GetDeviceAuthorizationForUpdate(ctx context.Context, authorizationID uuid.UUID) (store.DeviceauthDeviceAuthorization, bool, error) {
	if transaction == nil || transaction.queries == nil || authorizationID == uuid.Nil {
		return store.DeviceauthDeviceAuthorization{}, false, ErrRepository
	}
	row, err := transaction.queries.GetAuthorizationForUpdate(ctx, authorizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.DeviceauthDeviceAuthorization{}, false, nil
	}
	if err != nil {
		return store.DeviceauthDeviceAuthorization{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) DiscoverDeviceAuthorization(ctx context.Context, deviceID uuid.UUID) (store.DeviceauthDeviceAuthorization, bool, error) {
	if transaction == nil || transaction.queries == nil || deviceID == uuid.Nil {
		return store.DeviceauthDeviceAuthorization{}, false, ErrRepository
	}
	row, err := transaction.queries.DiscoverDeviceAuthorization(ctx, deviceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.DeviceauthDeviceAuthorization{}, false, nil
	}
	if err != nil {
		return store.DeviceauthDeviceAuthorization{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) GetDeviceForUpdate(ctx context.Context, deviceID uuid.UUID) (store.DeviceauthDevice, bool, error) {
	if transaction == nil || transaction.queries == nil || deviceID == uuid.Nil {
		return store.DeviceauthDevice{}, false, ErrRepository
	}
	row, err := transaction.queries.GetDeviceForUpdate(ctx, deviceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.DeviceauthDevice{}, false, nil
	}
	if err != nil {
		return store.DeviceauthDevice{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) GetDevicePolicySnapshot(ctx context.Context, authorizationID uuid.UUID) (store.DeviceauthDevicePolicySnapshot, bool, error) {
	if transaction == nil || transaction.queries == nil || authorizationID == uuid.Nil {
		return store.DeviceauthDevicePolicySnapshot{}, false, ErrRepository
	}
	row, err := transaction.queries.GetDevicePolicySnapshot(ctx, authorizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.DeviceauthDevicePolicySnapshot{}, false, nil
	}
	if err != nil {
		return store.DeviceauthDevicePolicySnapshot{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) MarkDeviceRefreshUsed(ctx context.Context, params store.MarkDeviceRefreshUsedParams) (bool, error) {
	if transaction == nil || transaction.queries == nil || len(params.TokenHash) != 32 || !params.UsedAt.Valid {
		return false, ErrRepository
	}
	row, err := transaction.queries.MarkDeviceRefreshUsed(ctx, params)
	clear(row.TokenHash)
	clear(row.PreviousTokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, ErrRepository
	}
	return true, nil
}

func (transaction *postgresTransaction) RotateDeviceFamilyAccess(ctx context.Context, params store.RotateDeviceFamilyAccessParams) (store.DeviceauthDeviceTokenFamily, bool, error) {
	if transaction == nil || transaction.queries == nil || params.ID == uuid.Nil || len(params.AccessTokenHash) != 32 {
		return store.DeviceauthDeviceTokenFamily{}, false, ErrRepository
	}
	row, err := transaction.queries.RotateDeviceFamilyAccess(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.DeviceauthDeviceTokenFamily{}, false, nil
	}
	if err != nil {
		return store.DeviceauthDeviceTokenFamily{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) CompromiseDeviceTokenFamily(ctx context.Context, params store.CompromiseDeviceTokenFamilyParams) (store.DeviceauthDeviceTokenFamily, bool, error) {
	if transaction == nil || transaction.queries == nil || params.ID == uuid.Nil {
		return store.DeviceauthDeviceTokenFamily{}, false, ErrRepository
	}
	row, err := transaction.queries.CompromiseDeviceTokenFamily(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.DeviceauthDeviceTokenFamily{}, false, nil
	}
	if err != nil {
		return store.DeviceauthDeviceTokenFamily{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) RevokeDeviceRefreshTokens(ctx context.Context, params store.RevokeDeviceRefreshTokensParams) (int64, error) {
	if transaction == nil || transaction.queries == nil || params.FamilyID == uuid.Nil || !params.RevokedAt.Valid {
		return 0, ErrRepository
	}
	rows, err := transaction.queries.RevokeDeviceRefreshTokens(ctx, params)
	if err != nil || rows < 0 {
		return 0, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) RevokeDeviceTokenFamily(ctx context.Context, params store.RevokeDeviceTokenFamilyParams) (int64, error) {
	if transaction == nil || transaction.queries == nil || params.ID == uuid.Nil {
		return 0, ErrRepository
	}
	rows, err := transaction.queries.RevokeDeviceTokenFamily(ctx, params)
	if err != nil || rows < 0 {
		return 0, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) RevokeDeviceAuthorization(ctx context.Context, params store.RevokeDeviceAuthorizationParams) (int64, error) {
	if transaction == nil || transaction.queries == nil || params.ID == uuid.Nil {
		return 0, ErrRepository
	}
	rows, err := transaction.queries.RevokeDeviceAuthorization(ctx, params)
	if err != nil || rows < 0 {
		return 0, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) RevokeDeviceRecord(ctx context.Context, params store.RevokeDeviceRecordParams) (int64, error) {
	if transaction == nil || transaction.queries == nil || params.ID == uuid.Nil {
		return 0, ErrRepository
	}
	rows, err := transaction.queries.RevokeDeviceRecord(ctx, params)
	if err != nil || rows < 0 {
		return 0, ErrRepository
	}
	return rows, nil
}

func rollbackDeviceauthTransaction(operationContext context.Context, transaction pgx.Tx) {
	defer func() { _ = recover() }()
	base := context.WithoutCancel(operationContext)
	ctx, cancel := context.WithTimeout(base, deviceauthRollbackTimeout)
	defer cancel()
	_ = transaction.Rollback(ctx)
}
