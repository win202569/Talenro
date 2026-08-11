package identity

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/outbox"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

const identityRollbackTimeout = 2 * time.Second

// PGXBeginner is the narrow production transaction source used by identity.
type PGXBeginner interface {
	Begin(context.Context) (pgx.Tx, error)
}

// PostgresRepository binds identity operations to generated sqlc methods.
type PostgresRepository struct {
	beginner  PGXBeginner
	protector sensitive.Protector
}

var _ Repository = (*PostgresRepository)(nil)

// NewPostgresRepository creates a generated-query-only identity repository.
func NewPostgresRepository(beginner PGXBeginner, protector sensitive.Protector) (*PostgresRepository, error) {
	if nilIdentityValue(beginner) || nilIdentityValue(protector) {
		return nil, ErrInvalidRepository
	}
	return &PostgresRepository{beginner: beginner, protector: protector}, nil
}

// WithinTransaction runs one callback and owns commit/independent rollback.
func (repository *PostgresRepository) WithinTransaction(ctx context.Context, operation func(context.Context, Transaction) error) (result error) {
	if nilIdentityValue(ctx) || repository == nil || nilIdentityValue(repository.beginner) || nilIdentityValue(repository.protector) || operation == nil {
		return ErrInvalidRepository
	}
	if ctx.Err() != nil {
		return ErrRepository
	}
	var tx pgx.Tx
	defer func() {
		if recover() != nil {
			result = ErrRepository
		}
		if !nilIdentityValue(tx) {
			rollbackIdentityTransaction(ctx, tx)
		}
	}()
	var err error
	tx, err = repository.beginner.Begin(ctx)
	if err != nil || nilIdentityValue(tx) {
		return ErrRepository
	}
	bound, err := newPostgresTransaction(tx, repository.protector)
	if err != nil {
		return ErrRepository
	}
	if err = operation(ctx, bound); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
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

func newPostgresTransaction(tx pgx.Tx, protector sensitive.Protector) (*postgresTransaction, error) {
	if nilIdentityValue(tx) || nilIdentityValue(protector) {
		return nil, ErrInvalidRepository
	}
	idempotencyRepository, err := idempotency.New(tx, protector)
	if err != nil {
		return nil, ErrRepository
	}
	outboxRepository, err := outbox.NewRepository(tx)
	if err != nil {
		return nil, ErrRepository
	}
	return &postgresTransaction{tx: tx, queries: store.New(tx), idempotency: idempotencyRepository, outbox: outboxRepository}, nil
}

func (transaction *postgresTransaction) DBTX() store.DBTX { return transaction.tx }

func (transaction *postgresTransaction) BeginIdempotency(ctx context.Context, scope idempotency.Scope, key string, canonical []byte, createdAt, expiresAt time.Time) (idempotency.Record, idempotency.Outcome, error) {
	if transaction == nil || nilIdentityValue(transaction.idempotency) {
		return idempotency.Record{}, "", ErrRepository
	}
	record, outcome, err := transaction.idempotency.Begin(ctx, scope, key, canonical, createdAt, expiresAt)
	if err != nil {
		return idempotency.Record{}, "", ErrRepository
	}
	return record, outcome, nil
}

func (transaction *postgresTransaction) CompleteIdempotency(ctx context.Context, record idempotency.Record, status int, body []byte) error {
	if transaction == nil || nilIdentityValue(transaction.idempotency) {
		return ErrRepository
	}
	completed, err := transaction.idempotency.Complete(ctx, record, status, body)
	if err != nil {
		return ErrRepository
	}
	ownedBody, owned := completed.TakeResponseBody()
	defer clear(ownedBody)
	if !owned {
		return ErrRepository
	}
	return nil
}

func (transaction *postgresTransaction) FindIdentityByLookupDigest(ctx context.Context, digest []byte) (store.IdentityEmailIdentity, bool, error) {
	row, err := transaction.queries.FindIdentityByLookupDigest(ctx, digest)
	return identityEmailResult(row, err)
}

func (transaction *postgresTransaction) FindIdentityByLookupDigestRead(ctx context.Context, digest []byte) (store.IdentityEmailIdentity, bool, error) {
	row, err := transaction.queries.FindIdentityByLookupDigestRead(ctx, digest)
	return identityEmailResult(row, err)
}

func (transaction *postgresTransaction) LockEmailLookupDigest(ctx context.Context, digest []byte) error {
	return mapIdentityStoreError(transaction.queries.LockEmailLookupDigest(ctx, digest))
}

func (transaction *postgresTransaction) GetEmailVerificationForUpdate(ctx context.Context, digest []byte) (store.IdentityEmailIdentity, bool, error) {
	row, err := transaction.queries.GetEmailVerificationForUpdate(ctx, digest)
	return identityEmailResult(row, err)
}

func (transaction *postgresTransaction) GetAccountForUpdate(ctx context.Context, principalID uuid.UUID) (store.IdentityAccount, bool, error) {
	row, err := transaction.queries.GetAccountForUpdate(ctx, principalID)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.IdentityAccount{}, false, nil
	}
	if err != nil {
		return store.IdentityAccount{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) GetAccountSessionForUpdate(ctx context.Context, params store.GetAccountSessionForUpdateParams) (store.IdentityAccountSession, bool, error) {
	row, err := transaction.queries.GetAccountSessionForUpdate(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.IdentityAccountSession{}, false, nil
	}
	if err != nil {
		return store.IdentityAccountSession{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) GetPasswordCredential(ctx context.Context, principalID uuid.UUID) (store.IdentityPasswordCredential, bool, error) {
	row, err := transaction.queries.GetPasswordCredential(ctx, principalID)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.IdentityPasswordCredential{}, false, nil
	}
	if err != nil {
		return store.IdentityPasswordCredential{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) GetPasswordResetForUpdate(ctx context.Context, digest []byte) (store.IdentityPasswordCredential, bool, error) {
	row, err := transaction.queries.GetPasswordResetForUpdate(ctx, digest)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.IdentityPasswordCredential{}, false, nil
	}
	if err != nil {
		return store.IdentityPasswordCredential{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) LockPrincipalAccountSessions(ctx context.Context, principalID uuid.UUID) ([]store.IdentityAccountSession, error) {
	rows, err := transaction.queries.LockPrincipalAccountSessions(ctx, principalID)
	if err != nil {
		return nil, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) LockPrincipalRefreshTokens(ctx context.Context, principalID uuid.UUID) ([]store.IdentityAccountRefreshToken, error) {
	rows, err := transaction.queries.LockPrincipalRefreshTokens(ctx, principalID)
	if err != nil {
		return nil, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) ListPrincipalRefreshTokens(ctx context.Context, principalID uuid.UUID) ([]store.IdentityAccountRefreshToken, error) {
	rows, err := transaction.queries.ListPrincipalRefreshTokens(ctx, principalID)
	if err != nil {
		return nil, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) CreateAccount(ctx context.Context, params store.CreateAccountParams) error {
	return mapIdentityStoreError(transaction.queries.CreateAccount(ctx, params))
}

func (transaction *postgresTransaction) CreateEmailIdentity(ctx context.Context, params store.CreateEmailIdentityParams) error {
	return mapIdentityStoreError(transaction.queries.CreateEmailIdentity(ctx, params))
}

func (transaction *postgresTransaction) CreatePasswordCredential(ctx context.Context, params store.CreatePasswordCredentialParams) error {
	return mapIdentityStoreError(transaction.queries.CreatePasswordCredential(ctx, params))
}

func (transaction *postgresTransaction) InsertSecurityEvent(ctx context.Context, params store.InsertSecurityEventParams) error {
	return mapIdentityStoreError(transaction.queries.InsertSecurityEvent(ctx, params))
}

func (transaction *postgresTransaction) ResetEmailVerification(ctx context.Context, params store.ResetEmailVerificationParams) (bool, error) {
	_, err := transaction.queries.ResetEmailVerification(ctx, params)
	return oneRowResult(err)
}

func (transaction *postgresTransaction) SetPasswordReset(ctx context.Context, params store.SetPasswordResetParams) (bool, error) {
	_, err := transaction.queries.SetPasswordReset(ctx, params)
	return oneRowResult(err)
}

func (transaction *postgresTransaction) ConsumeEmailVerification(ctx context.Context, params store.ConsumeEmailVerificationParams) (store.IdentityEmailIdentity, bool, error) {
	row, err := transaction.queries.ConsumeEmailVerification(ctx, params)
	return identityEmailResult(row, err)
}

func (transaction *postgresTransaction) ClearPendingEmailDelivery(ctx context.Context, params store.ClearPendingEmailDeliveryParams) (int64, error) {
	rows, err := transaction.queries.ClearPendingEmailDelivery(ctx, params)
	if err != nil || rows < 0 || rows > 1 {
		return 0, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) ActivateVerifiedAccount(ctx context.Context, params store.ActivateVerifiedAccountParams) (store.IdentityAccount, bool, error) {
	row, err := transaction.queries.ActivateVerifiedAccount(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.IdentityAccount{}, false, nil
	}
	if err != nil {
		return store.IdentityAccount{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) ConsumePasswordReset(ctx context.Context, params store.ConsumePasswordResetParams) (bool, error) {
	_, err := transaction.queries.ConsumePasswordReset(ctx, params)
	return oneRowResult(err)
}

func (transaction *postgresTransaction) MarkPrincipalSessionsReviewRequired(ctx context.Context, params store.MarkPrincipalSessionsReviewRequiredParams) (int64, error) {
	rows, err := transaction.queries.MarkPrincipalSessionsReviewRequired(ctx, params)
	if err != nil || rows < 0 {
		return 0, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) CreateAccountSession(ctx context.Context, params store.CreateAccountSessionParams) error {
	return mapIdentityStoreError(transaction.queries.CreateAccountSession(ctx, params))
}

func (transaction *postgresTransaction) InsertAccountRefreshToken(ctx context.Context, params store.InsertAccountRefreshTokenParams) error {
	return mapIdentityStoreError(transaction.queries.InsertAccountRefreshToken(ctx, params))
}

func (transaction *postgresTransaction) FindAccountAccessToken(ctx context.Context, digest []byte) (store.FindAccountAccessTokenRow, bool, error) {
	row, err := transaction.queries.FindAccountAccessToken(ctx, digest)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.FindAccountAccessTokenRow{}, false, nil
	}
	if err != nil {
		return store.FindAccountAccessTokenRow{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) DiscoverRefreshToken(ctx context.Context, digest []byte) (store.DiscoverRefreshTokenRow, bool, error) {
	row, err := transaction.queries.DiscoverRefreshToken(ctx, digest)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.DiscoverRefreshTokenRow{}, false, nil
	}
	if err != nil {
		return store.DiscoverRefreshTokenRow{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) LockSessionRefreshTokens(ctx context.Context, sessionID uuid.UUID) ([]store.IdentityAccountRefreshToken, error) {
	rows, err := transaction.queries.LockSessionRefreshTokens(ctx, sessionID)
	if err != nil {
		return nil, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) ListSessionRefreshTokens(ctx context.Context, sessionID uuid.UUID) ([]store.IdentityAccountRefreshToken, error) {
	rows, err := transaction.queries.ListSessionRefreshTokens(ctx, sessionID)
	if err != nil {
		return nil, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) MarkAccountRefreshUsed(ctx context.Context, params store.MarkAccountRefreshUsedParams) (bool, error) {
	_, err := transaction.queries.MarkAccountRefreshUsed(ctx, params)
	return oneRowResult(err)
}

func (transaction *postgresTransaction) RotateAccountSessionAccess(ctx context.Context, params store.RotateAccountSessionAccessParams) (store.IdentityAccountSession, bool, error) {
	row, err := transaction.queries.RotateAccountSessionAccess(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.IdentityAccountSession{}, false, nil
	}
	if err != nil {
		return store.IdentityAccountSession{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) RevokeAccountRefreshTokens(ctx context.Context, params store.RevokeAccountRefreshTokensParams) (int64, error) {
	rows, err := transaction.queries.RevokeAccountRefreshTokens(ctx, params)
	if err != nil || rows < 0 {
		return 0, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) MarkAccountSessionCompromised(ctx context.Context, params store.MarkAccountSessionCompromisedParams) (int64, error) {
	rows, err := transaction.queries.MarkAccountSessionCompromised(ctx, params)
	if err != nil || rows < 0 || rows > 1 {
		return 0, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) RevokePrincipalAccountSessions(ctx context.Context, params store.RevokePrincipalAccountSessionsParams) ([]uuid.UUID, error) {
	ids, err := transaction.queries.RevokePrincipalAccountSessions(ctx, params)
	if err != nil {
		return nil, ErrRepository
	}
	for _, id := range ids {
		if id == uuid.Nil {
			return nil, ErrRepository
		}
	}
	return ids, nil
}

func (transaction *postgresTransaction) UpdatePasswordCredential(ctx context.Context, params store.UpdatePasswordCredentialParams) (int64, error) {
	rows, err := transaction.queries.UpdatePasswordCredential(ctx, params)
	if err != nil || rows < 0 || rows > 1 {
		return 0, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) CreateEnrollmentGrant(ctx context.Context, params store.CreateEnrollmentGrantParams) error {
	return mapIdentityStoreError(transaction.queries.CreateEnrollmentGrant(ctx, params))
}

func (transaction *postgresTransaction) GetTOTPForUpdate(
	ctx context.Context,
	principalID uuid.UUID,
) (store.IdentityTotpCredential, bool, error) {
	row, err := transaction.queries.GetTOTPForUpdate(ctx, principalID)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.IdentityTotpCredential{}, false, nil
	}
	if err != nil {
		return store.IdentityTotpCredential{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) CreateTOTPEnrollment(
	ctx context.Context,
	params store.CreateTOTPEnrollmentParams,
) (int64, error) {
	rows, err := transaction.queries.CreateTOTPEnrollment(ctx, params)
	if err != nil || rows < 0 || rows > 1 {
		return 0, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) ActivateTOTP(
	ctx context.Context,
	params store.ActivateTOTPParams,
) (int64, error) {
	rows, err := transaction.queries.ActivateTOTP(ctx, params)
	if err != nil || rows < 0 || rows > 1 {
		return 0, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) AcceptTOTPStep(
	ctx context.Context,
	params store.AcceptTOTPStepParams,
) (int64, error) {
	rows, err := transaction.queries.AcceptTOTPStep(ctx, params)
	if err != nil || rows < 0 || rows > 1 {
		return 0, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) RevokeTOTP(
	ctx context.Context,
	params store.RevokeTOTPParams,
) (int64, error) {
	rows, err := transaction.queries.RevokeTOTP(ctx, params)
	if err != nil || rows < 0 || rows > 1 {
		return 0, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) GetActiveRecoveryCodeSetForUpdate(
	ctx context.Context,
	principalID uuid.UUID,
) (store.IdentityRecoveryCodeSet, bool, error) {
	row, err := transaction.queries.GetActiveRecoveryCodeSetForUpdate(ctx, principalID)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.IdentityRecoveryCodeSet{}, false, nil
	}
	if err != nil {
		return store.IdentityRecoveryCodeSet{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) GetNextRecoveryCodeGeneration(
	ctx context.Context,
	principalID uuid.UUID,
) (int32, error) {
	generation, err := transaction.queries.GetNextRecoveryCodeGeneration(ctx, principalID)
	if err != nil || generation < 1 {
		return 0, ErrRepository
	}
	return generation, nil
}

func (transaction *postgresTransaction) CreateRecoveryCodeSet(
	ctx context.Context,
	params store.CreateRecoveryCodeSetParams,
) error {
	return mapIdentityStoreError(transaction.queries.CreateRecoveryCodeSet(ctx, params))
}

func (transaction *postgresTransaction) ConsumeRecoveryCode(
	ctx context.Context,
	params store.ConsumeRecoveryCodeParams,
) (store.IdentityRecoveryCodeSet, bool, error) {
	row, err := transaction.queries.ConsumeRecoveryCode(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.IdentityRecoveryCodeSet{}, false, nil
	}
	if err != nil {
		return store.IdentityRecoveryCodeSet{}, false, ErrRepository
	}
	return row, true, nil
}

func (transaction *postgresTransaction) RevokeRecoveryCodeSets(
	ctx context.Context,
	params store.RevokeRecoveryCodeSetsParams,
) (int64, error) {
	rows, err := transaction.queries.RevokeRecoveryCodeSets(ctx, params)
	if err != nil || rows < 0 || rows > 1 {
		return 0, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) ListActivePasskeys(
	ctx context.Context,
	principalID uuid.UUID,
) ([]store.IdentityPasskeyCredential, error) {
	rows, err := transaction.queries.ListActivePasskeys(ctx, principalID)
	if err != nil || len(rows) > maximumActivePasskeys {
		return nil, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) CreatePasskeyCredential(
	ctx context.Context,
	params store.CreatePasskeyCredentialParams,
) error {
	return mapIdentityStoreError(transaction.queries.CreatePasskeyCredential(ctx, params))
}

func (transaction *postgresTransaction) UpdatePasskeyCounter(
	ctx context.Context,
	params store.UpdatePasskeyCounterParams,
) (int64, error) {
	rows, err := transaction.queries.UpdatePasskeyCounter(ctx, params)
	if err != nil || rows < 0 || rows > 1 {
		return 0, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) RevokePasskey(
	ctx context.Context,
	params store.RevokePasskeyParams,
) (int64, error) {
	rows, err := transaction.queries.RevokePasskey(ctx, params)
	if err != nil || rows < 0 || rows > 1 {
		return 0, ErrRepository
	}
	return rows, nil
}

func (transaction *postgresTransaction) AppendEvent(ctx context.Context, envelope *eventsv1.EventEnvelope) error {
	if err := transaction.outbox.Append(ctx, envelope); err != nil {
		return ErrRepository
	}
	return nil
}

func identityEmailResult(row store.IdentityEmailIdentity, err error) (store.IdentityEmailIdentity, bool, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return store.IdentityEmailIdentity{}, false, nil
	}
	if err != nil {
		return store.IdentityEmailIdentity{}, false, ErrRepository
	}
	return row, true, nil
}

func oneRowResult(err error) (bool, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, ErrRepository
	}
	return true, nil
}

func mapIdentityStoreError(err error) error {
	if err != nil {
		return ErrRepository
	}
	return nil
}

func rollbackIdentityTransaction(operationContext context.Context, tx pgx.Tx) {
	defer func() { _ = recover() }()
	base := context.WithoutCancel(operationContext)
	ctx, cancel := context.WithTimeout(base, identityRollbackTimeout)
	defer cancel()
	_ = tx.Rollback(ctx)
}

func nilIdentityValue(value any) bool {
	if value == nil {
		return true
	}
	representation := reflect.ValueOf(value)
	switch representation.Kind() { //nolint:exhaustive // Only nil-capable interface representations matter.
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return representation.IsNil()
	default:
		return false
	}
}
