package authority

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gowebpki/jcs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/store"
)

type PostgresRepository struct {
	database store.DBTX
}

var _ Repository = (*PostgresRepository)(nil)
var _ AuthorityV7Repository = (*PostgresRepository)(nil)
var _ FreshRestoreImportRepository = (*PostgresRepository)(nil)

func NewPostgresRepository(database store.DBTX) (*PostgresRepository, error) {
	if nilAuthorityRepositoryValue(database) {
		return nil, ErrInvalidArgument
	}
	return &PostgresRepository{database: database}, nil
}

func (repository *PostgresRepository) RecordPending(
	ctx context.Context,
	dbtx store.DBTX,
	reservation Reservation,
	reservedAt time.Time,
) error {
	reservedAt = canonicalDatabaseTimestamp(reservedAt)
	if repositoryCallError(ctx, repository) != nil || nilAuthorityRepositoryValue(dbtx) ||
		reservation.Validate() != nil || reservedAt.IsZero() {
		return repositoryInputError(ctx)
	}
	queries := store.New(dbtx)
	rows, err := queries.InsertAuthorityFencePending(ctx, store.InsertAuthorityFencePendingParams{
		OperationID:               reservation.OperationID,
		EffectKind:                string(reservation.Kind),
		ScopeKind:                 string(reservation.ScopeKind),
		AuthorityEpoch:            int64(reservation.Epoch),
		AuthoritySequence:         int64(reservation.Sequence),
		ScopeDigest:               digestBytes(reservation.ScopeDigest),
		ProviderReservationDigest: digestBytes(reservation.ReservationDigest),
		ReservedAt:                requiredTimestamp(reservedAt),
	})
	if err != nil {
		if cancellationErr := repositoryCancellationError(ctx, err); cancellationErr != nil {
			return cancellationErr
		}
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return ErrConflict
		}
		return ErrInjectedFailure
	}
	switch rows {
	case 1:
		return nil
	case 0:
		if cancellationErr := repositoryCancellationError(ctx, nil); cancellationErr != nil {
			return cancellationErr
		}
		currentRow, getErr := queries.GetAuthorityFenceForUpdate(ctx, reservation.OperationID)
		if getErr != nil {
			if cancellationErr := repositoryCancellationError(ctx, getErr); cancellationErr != nil {
				return cancellationErr
			}
			if errors.Is(getErr, pgx.ErrNoRows) {
				return ErrConflict
			}
			return ErrInjectedFailure
		}
		current := authorityFenceRowFromForUpdate(currentRow)
		if _, conversionErr := authorityRecordFromRow(current); conversionErr != nil {
			return conversionErr
		}
		if current.ReservedAt.Valid && current.ReservedAt.Time.Equal(reservedAt) && reservationMatchesRow(reservation, current) {
			return nil
		}
		return ErrConflict
	default:
		return ErrInjectedFailure
	}
}

func (repository *PostgresRepository) RecordPendingClaimV1(
	ctx context.Context,
	dbtx store.DBTX,
	claim ClaimV1Reservation,
	reservedAt time.Time,
) error {
	reservedAt = canonicalDatabaseTimestamp(reservedAt)
	if repositoryCallError(ctx, repository) != nil || nilAuthorityRepositoryValue(dbtx) ||
		claim.Reservation.Validate() != nil || !validOperationID(claim.ProtocolActivationID) || reservedAt.IsZero() {
		return repositoryInputError(ctx)
	}
	queries := store.New(dbtx)
	rows, err := queries.InsertClaimV1AuthorityFencePending(ctx, store.InsertClaimV1AuthorityFencePendingParams{
		OperationID:               claim.Reservation.OperationID,
		EffectKind:                string(claim.Reservation.Kind),
		ScopeKind:                 string(claim.Reservation.ScopeKind),
		AuthorityEpoch:            int64(claim.Reservation.Epoch),
		AuthoritySequence:         int64(claim.Reservation.Sequence),
		ScopeDigest:               digestBytes(claim.Reservation.ScopeDigest),
		ProviderReservationDigest: digestBytes(claim.Reservation.ReservationDigest),
		ReservedAt:                requiredTimestamp(reservedAt),
		ProtocolActivationID:      uuid.NullUUID{UUID: claim.ProtocolActivationID, Valid: true},
	})
	if err != nil {
		if cancellationErr := repositoryCancellationError(ctx, err); cancellationErr != nil {
			return cancellationErr
		}
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return ErrConflict
		}
		return ErrInjectedFailure
	}
	if rows == 1 {
		return nil
	}
	if rows != 0 {
		return ErrInjectedFailure
	}
	current, getErr := queries.LockAuthorityFence(ctx, claim.Reservation.OperationID)
	if getErr != nil {
		return repositoryLookupError(ctx, getErr)
	}
	stored, conversionErr := storedFenceFromV7Row(current)
	if conversionErr != nil {
		return conversionErr
	}
	if stored.AbortClaim == nil && current.ReservedAt.Valid && current.ReservedAt.Time.Equal(reservedAt) &&
		current.AuthorityProtocolProfile == string(contracts.AuthorityV7ProtocolProfileClaimV1) &&
		current.ProtocolActivationID.Valid && current.ProtocolActivationID.UUID == claim.ProtocolActivationID &&
		stored.Record.Reservation == claim.Reservation {
		return nil
	}
	return ErrConflict
}

func (repository *PostgresRepository) ClaimAbort(
	ctx context.Context,
	dbtx store.DBTX,
	operationID uuid.UUID,
	reason AbortReason,
	claimedAt time.Time,
) error {
	claimedAt = canonicalDatabaseTimestamp(claimedAt)
	if repositoryCallError(ctx, repository) != nil || nilAuthorityRepositoryValue(dbtx) ||
		!validOperationID(operationID) || reason.Validate() != nil || claimedAt.IsZero() {
		return repositoryInputError(ctx)
	}
	queries := store.New(dbtx)
	current, err := queries.LockAuthorityFence(ctx, operationID)
	if err != nil {
		return repositoryLookupError(ctx, err)
	}
	stored, conversionErr := storedFenceFromV7Row(current)
	if conversionErr != nil {
		return conversionErr
	}
	if current.AuthorityProtocolProfile != string(contracts.AuthorityV7ProtocolProfileClaimV1) ||
		!current.ProtocolActivationID.Valid {
		return ErrConflict
	}
	if stored.Record.TerminalReceipt != nil || stored.Record.BoundEffectDigest != nil {
		return ErrTerminalConflict
	}
	if stored.AbortClaim != nil {
		if stored.AbortClaim.Reason == reason && stored.AbortClaim.ClaimedAt.Equal(claimedAt) {
			return nil
		}
		return ErrConflict
	}
	rows, updateErr := queries.ClaimAuthorityAbort(ctx, store.ClaimAuthorityAbortParams{
		OperationID:    operationID,
		AbortReason:    pgtype.Text{String: string(reason), Valid: true},
		AbortClaimedAt: requiredTimestamp(claimedAt),
	})
	if updateErr != nil {
		return repositoryDependencyError(ctx, updateErr)
	}
	if rows != 1 {
		return ErrConflict
	}
	return nil
}

func (repository *PostgresRepository) GetStoredFence(ctx context.Context, operationID uuid.UUID) (StoredFence, error) {
	if repositoryCallError(ctx, repository) != nil || !validOperationID(operationID) {
		return StoredFence{}, repositoryInputError(ctx)
	}
	row, err := store.New(repository.database).GetStoredAuthorityFence(ctx, operationID)
	if err != nil {
		return StoredFence{}, fmt.Errorf("get stored authority fence: %v: %w", err, repositoryLookupError(ctx, err))
	}
	stored, conversionErr := storedFenceFromV7Row(row)
	if conversionErr != nil {
		return StoredFence{}, fmt.Errorf("decode stored authority fence: %w", conversionErr)
	}
	if row.AuthorityProtocolProfile == string(contracts.AuthorityV7ProtocolProfileClaimV1) {
		outcome, outcomeErr := repository.persistedOutcomeForUpdate(ctx, store.New(repository.database), stored.Record)
		if outcomeErr != nil {
			return StoredFence{}, outcomeErr
		}
		stored.PersistedOutcome = outcome
	}
	return cloneStoredFence(stored), nil
}

func (repository *PostgresRepository) Lock(ctx context.Context, dbtx store.DBTX, operationID uuid.UUID) (StoredFence, error) {
	if repositoryCallError(ctx, repository) != nil || nilAuthorityRepositoryValue(dbtx) || !validOperationID(operationID) {
		return StoredFence{}, repositoryInputError(ctx)
	}
	queries := store.New(dbtx)
	row, err := queries.LockAuthorityFence(ctx, operationID)
	if err != nil {
		return StoredFence{}, repositoryLookupError(ctx, err)
	}
	stored, conversionErr := storedFenceFromV7Row(row)
	if conversionErr != nil {
		return StoredFence{}, conversionErr
	}
	if row.AuthorityProtocolProfile == string(contracts.AuthorityV7ProtocolProfileClaimV1) {
		outcome, outcomeErr := repository.persistedOutcomeForUpdate(ctx, queries, stored.Record)
		if outcomeErr != nil {
			return StoredFence{}, outcomeErr
		}
		stored.PersistedOutcome = outcome
	}
	return cloneStoredFence(stored), nil
}

func (repository *PostgresRepository) ConsumeVerifiedFreshRestoreImport(
	ctx context.Context,
	admission VerifiedFreshRestoreImportAdmission,
	projection contracts.FreshImportTopologyProjectionV1,
) (contracts.FreshRestoreImportApplicationV1, error) {
	view, consumeErr := consumeVerifiedFreshRestoreImportAdmission(admission)
	if consumeErr != nil {
		return contracts.FreshRestoreImportApplicationV1{}, consumeErr
	}
	facts := view.Facts
	if repositoryCallError(ctx, repository) != nil || projection.Validate() != nil ||
		projection.TargetActivationID != facts.ManifestTopology.TargetActivationID ||
		projection.TargetDeploymentID != facts.ManifestTopology.TargetDeploymentID ||
		projection.TargetDatabaseIdentityDigest != facts.CurrentDatabaseIdentityDigest ||
		projection.NormalizedCatalogDigest != facts.NormalizedCatalogDigest ||
		projection.ObjectCount != facts.ManifestTopology.ObjectCount ||
		len(projection.Objects) != len(facts.ManifestTopology.Objects) {
		return contracts.FreshRestoreImportApplicationV1{}, repositoryInputError(ctx)
	}
	for index := range projection.Objects {
		if projection.Objects[index].ObjectType != facts.ManifestTopology.Objects[index].ObjectType ||
			projection.Objects[index].CanonicalKey != facts.ManifestTopology.Objects[index].CanonicalKey {
			return contracts.FreshRestoreImportApplicationV1{}, ErrInvalidArgument
		}
	}
	projectionJCS, projectionDigest, err := canonicalFreshImportTopologyProjection(projection)
	if err != nil || projectionDigest != facts.ManifestTopology.ExpectedPostImportInventoryDigest ||
		projectionDigest != facts.StagingImportCapability.ExpectedPostImportInventoryDigest {
		return contracts.FreshRestoreImportApplicationV1{}, ErrInvalidArgument
	}
	objectsJSON, err := freshImportProjectionObjectsJSON(projection.Objects)
	if err != nil {
		return contracts.FreshRestoreImportApplicationV1{}, ErrInvalidArgument
	}
	row, queryErr := store.New(repository.database).BeginStagingImport(ctx, store.BeginStagingImportParams{
		CapabilityDigest:      digestBytes(facts.StagingImportCapability.CapabilityDigest),
		ManifestDigest:        digestBytes(facts.ManifestTopology.ManifestDigest),
		ExclusionLeaseDigest:  digestBytes(facts.StagingExclusionLease.LeaseDigest),
		AcquisitionHeadDigest: digestBytes(facts.StagingExclusionLease.AcquisitionLockedProviderHeadDigest),
		RouteClosedDigest:     digestBytes(facts.DatabaseRouteClosedDigest),
		ProjectionBodyJcs:     projectionJCS,
		ProjectionDigest:      digestBytes(projectionDigest),
		ProjectionObjects:     objectsJSON,
	})
	if queryErr != nil {
		if cancellationErr := repositoryCancellationError(ctx, queryErr); cancellationErr != nil {
			return contracts.FreshRestoreImportApplicationV1{}, cancellationErr
		}
		if errors.Is(queryErr, pgx.ErrNoRows) {
			return contracts.FreshRestoreImportApplicationV1{}, ErrConflict
		}
		return contracts.FreshRestoreImportApplicationV1{}, ErrInjectedFailure
	}
	application, conversionErr := freshRestoreImportApplicationFromRow(row)
	if conversionErr != nil || application.SingleUseApplyID != facts.ManifestTopology.SingleUseApplyID ||
		application.StagingImportCapabilityDigest != facts.StagingImportCapability.CapabilityDigest ||
		application.ManifestDigest != facts.ManifestTopology.ManifestDigest ||
		application.TargetActivationID != facts.ManifestTopology.TargetActivationID ||
		application.CurrentDatabaseIdentityDigest != facts.CurrentDatabaseIdentityDigest ||
		application.DatabaseTimelineLineageChainDigest != facts.DatabaseTimelineLineageChainDigest ||
		application.TargetDatabaseIncarnationRegistrationDigest != facts.CurrentDatabaseIncarnationRegistrationDigest ||
		application.RuntimeRebindChainDigest != facts.RuntimeRebindChainDigest ||
		application.RuntimeInstanceBindingDigest != facts.RuntimeInstanceBindingDigest ||
		application.StagingExclusionLeaseDigest != facts.StagingExclusionLease.LeaseDigest ||
		application.AcquisitionLockedProviderHeadDigest != facts.CurrentProviderHeadDigest ||
		application.DatabaseRouteClosedDigest != facts.DatabaseRouteClosedDigest ||
		application.PreImportInventoryDigest != facts.PreImportInventoryDigest ||
		application.PostImportInventoryDigest != projectionDigest ||
		application.ImportedObjectCount != projection.ObjectCount ||
		application.CompleteNodeSetDigest != facts.ManifestTopology.CompleteNodeSetDigest {
		return contracts.FreshRestoreImportApplicationV1{}, ErrInjectedFailure
	}
	return application, nil
}

func (repository *PostgresRepository) CaptureDatabasePoint(ctx context.Context) (DatabasePoint, error) {
	if repositoryCallError(ctx, repository) != nil {
		return DatabasePoint{}, repositoryInputError(ctx)
	}
	row, err := store.New(repository.database).GetNodeControlDatabaseIdentity(ctx)
	if err != nil {
		return DatabasePoint{}, repositoryDependencyError(ctx, err)
	}
	point, conversionErr := databasePoint(row.SystemID, pgtype.Int8{Int64: row.Timeline, Valid: true}, row.RequiredLsn)
	if conversionErr != nil {
		return DatabasePoint{}, conversionErr
	}
	return point, nil
}

func (repository *PostgresRepository) BindEffect(
	ctx context.Context,
	dbtx store.DBTX,
	operationID uuid.UUID,
	effectDigest contracts.Digest,
	point DatabasePoint,
	effectBoundAt time.Time,
) error {
	effectBoundAt = canonicalDatabaseTimestamp(effectBoundAt)
	if repositoryCallError(ctx, repository) != nil || nilAuthorityRepositoryValue(dbtx) ||
		!validOperationID(operationID) || effectDigest == (contracts.Digest{}) ||
		point.Validate() != nil || effectBoundAt.IsZero() {
		return repositoryInputError(ctx)
	}
	queries := store.New(dbtx)
	currentRow, err := queries.GetAuthorityFenceForUpdate(ctx, operationID)
	if err != nil {
		return repositoryLookupError(ctx, err)
	}
	current := authorityFenceRowFromForUpdate(currentRow)
	if current.ProviderStatus == "reserved" && current.AbortReason.Valid {
		return ErrTerminalConflict
	}
	record, conversionErr := authorityRecordFromRow(current)
	if conversionErr != nil {
		return conversionErr
	}
	if record.TerminalReceipt != nil && record.TerminalReceipt.Status == StatusAborted {
		return ErrTerminalConflict
	}
	if record.BoundEffectDigest != nil {
		if *record.BoundEffectDigest == effectDigest && *record.BoundDatabasePoint == point &&
			current.EffectBoundAt.Valid && current.EffectBoundAt.Time.Equal(effectBoundAt) {
			return nil
		}
		return ErrConflict
	}
	if record.TerminalReceipt != nil || current.ProviderStatus != "reserved" || current.VisibilityState != "fence_pending" {
		return ErrTerminalConflict
	}
	rows, updateErr := queries.BindAuthorityFenceEffect(ctx, store.BindAuthorityFenceEffectParams{
		OperationID:   operationID,
		EffectDigest:  digestBytes(effectDigest),
		DbSystemID:    numericFromUint64(point.SystemID),
		DbTimeline:    pgtype.Int8{Int64: int64(point.Timeline), Valid: true},
		RequiredLsn:   string(point.RequiredLSN),
		EffectBoundAt: requiredTimestamp(effectBoundAt),
	})
	if updateErr != nil {
		return repositoryDependencyError(ctx, updateErr)
	}
	if rows != 1 {
		return ErrConflict
	}
	return nil
}

func (repository *PostgresRepository) ActivateCommitted(
	ctx context.Context,
	dbtx store.DBTX,
	receipt Receipt,
	terminalAt time.Time,
) error {
	terminalAt = canonicalDatabaseTimestamp(terminalAt)
	if repositoryCallError(ctx, repository) != nil || nilAuthorityRepositoryValue(dbtx) ||
		receipt.Validate() != nil || receipt.Status != StatusCommitted || terminalAt.IsZero() {
		return repositoryInputError(ctx)
	}
	queries := store.New(dbtx)
	currentRow, err := queries.GetAuthorityFenceForUpdate(ctx, receipt.OperationID)
	if err != nil {
		return repositoryLookupError(ctx, err)
	}
	current := authorityFenceRowFromForUpdate(currentRow)
	if current.ProviderStatus == "reserved" && current.AbortReason.Valid {
		return ErrTerminalConflict
	}
	record, conversionErr := authorityRecordFromRow(current)
	if conversionErr != nil {
		return conversionErr
	}
	if record.Reservation != receipt.Reservation {
		return ErrConflict
	}
	if record.TerminalReceipt != nil {
		if record.TerminalReceipt.Status == StatusAborted {
			return ErrTerminalConflict
		}
		if receiptsEqual(*record.TerminalReceipt, receipt) && current.TerminalAt.Valid && current.TerminalAt.Time.Equal(terminalAt) {
			return nil
		}
		return ErrConflict
	}
	if record.BoundEffectDigest == nil || *record.BoundEffectDigest != *receipt.EffectDigest ||
		*record.BoundDatabasePoint != *receipt.DatabasePoint {
		return ErrConflict
	}
	rows, updateErr := queries.ActivateCommittedAuthorityFence(ctx, store.ActivateCommittedAuthorityFenceParams{
		OperationID:           receipt.OperationID,
		ProviderReceiptDigest: digestBytes(receipt.ReceiptDigest),
		TerminalAt:            requiredTimestamp(terminalAt),
		EffectDigest:          digestBytes(*receipt.EffectDigest),
		DbSystemID:            numericFromUint64(receipt.DatabasePoint.SystemID),
		DbTimeline:            pgtype.Int8{Int64: int64(receipt.DatabasePoint.Timeline), Valid: true},
		RequiredLsn:           string(receipt.DatabasePoint.RequiredLSN),
	})
	if updateErr != nil {
		return repositoryDependencyError(ctx, updateErr)
	}
	if rows != 1 {
		return ErrConflict
	}
	return nil
}

func (repository *PostgresRepository) RecordAborted(
	ctx context.Context,
	dbtx store.DBTX,
	receipt Receipt,
	terminalAt time.Time,
) error {
	terminalAt = canonicalDatabaseTimestamp(terminalAt)
	if repositoryCallError(ctx, repository) != nil || nilAuthorityRepositoryValue(dbtx) ||
		receipt.Validate() != nil || receipt.Status != StatusAborted || terminalAt.IsZero() {
		return repositoryInputError(ctx)
	}
	queries := store.New(dbtx)
	currentRow, err := queries.GetAuthorityFenceForUpdate(ctx, receipt.OperationID)
	if err != nil {
		return repositoryLookupError(ctx, err)
	}
	current := authorityFenceRowFromForUpdate(currentRow)
	claimedReason := AbortReason("")
	if current.ProviderStatus == "reserved" && current.AbortReason.Valid {
		claimedReason = AbortReason(current.AbortReason.String)
		if claimedReason.Validate() != nil || *receipt.AbortReason != claimedReason {
			return ErrConflict
		}
		current.AbortReason = pgtype.Text{}
	}
	record, conversionErr := authorityRecordFromRow(current)
	if conversionErr != nil {
		return conversionErr
	}
	if record.Reservation != receipt.Reservation {
		return ErrConflict
	}
	if record.TerminalReceipt != nil {
		if record.TerminalReceipt.Status == StatusCommitted {
			return ErrTerminalConflict
		}
		if receiptsEqual(*record.TerminalReceipt, receipt) && current.TerminalAt.Valid && current.TerminalAt.Time.Equal(terminalAt) {
			return nil
		}
		return ErrConflict
	}
	if receipt.EffectDigest != nil && (record.BoundEffectDigest == nil ||
		*record.BoundEffectDigest != *receipt.EffectDigest ||
		*record.BoundDatabasePoint != *receipt.DatabasePoint) {
		return ErrConflict
	}
	rows, updateErr := queries.AbortAuthorityFence(ctx, store.AbortAuthorityFenceParams{
		OperationID:           receipt.OperationID,
		ProviderReceiptDigest: digestBytes(receipt.ReceiptDigest),
		AbortReason:           pgtype.Text{String: string(*receipt.AbortReason), Valid: true},
		TerminalAt:            requiredTimestamp(terminalAt),
	})
	if updateErr != nil {
		return repositoryDependencyError(ctx, updateErr)
	}
	if rows != 1 {
		return ErrConflict
	}
	return nil
}

func (repository *PostgresRepository) Get(ctx context.Context, operationID uuid.UUID) (Record, error) {
	if repositoryCallError(ctx, repository) != nil || !validOperationID(operationID) {
		return Record{}, repositoryInputError(ctx)
	}
	row, err := store.New(repository.database).GetAuthorityFence(ctx, operationID)
	if err != nil {
		return Record{}, repositoryLookupError(ctx, err)
	}
	return authorityRecordFromRow(authorityFenceRowFromGet(row))
}

func (repository *PostgresRepository) Head(ctx context.Context) (DatabaseHead, error) {
	if repositoryCallError(ctx, repository) != nil {
		return DatabaseHead{}, repositoryInputError(ctx)
	}
	row, err := store.New(repository.database).GetAuthorityFenceHead(ctx)
	if err != nil {
		return DatabaseHead{}, repositoryDependencyError(ctx, err)
	}
	return databaseHeadFromRow(row)
}

func (repository *PostgresRepository) ListPending(ctx context.Context, epoch uint64) ([]PendingFence, error) {
	if repositoryCallError(ctx, repository) != nil || !validPositiveCoordinate(epoch) {
		return nil, repositoryInputError(ctx)
	}
	rows, err := store.New(repository.database).ListPendingAuthorityFences(ctx, int64(epoch))
	if err != nil {
		return nil, repositoryDependencyError(ctx, err)
	}
	result := make([]PendingFence, 0, len(rows))
	var previousSequence uint64
	for _, row := range rows {
		pending, conversionErr := pendingFenceFromRow(row)
		if conversionErr != nil || pending.Epoch != epoch || pending.Sequence <= previousSequence {
			return nil, ErrInjectedFailure
		}
		previousSequence = pending.Sequence
		result = append(result, pending)
	}
	return result, nil
}

func (repository *PostgresRepository) CommittedNodeCheckpoint(
	ctx context.Context,
	epoch uint64,
	nodeScopeDigest contracts.Digest,
) (NodeCheckpoint, error) {
	if repositoryCallError(ctx, repository) != nil || !validPositiveCoordinate(epoch) ||
		!validScopeDigest(ScopeNode, nodeScopeDigest) {
		return NodeCheckpoint{}, repositoryInputError(ctx)
	}
	row, err := store.New(repository.database).GetCommittedNodeCheckpoint(ctx, store.GetCommittedNodeCheckpointParams{
		AuthorityEpoch:  int64(epoch),
		NodeScopeDigest: digestBytes(nodeScopeDigest),
	})
	if err != nil {
		if cancellationErr := repositoryCancellationError(ctx, err); cancellationErr != nil {
			return NodeCheckpoint{}, cancellationErr
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return NodeCheckpoint{AuthorityEpoch: epoch}, nil
		}
		return NodeCheckpoint{}, ErrInjectedFailure
	}
	if row.AuthorityEpoch != int64(epoch) || row.AuthoritySequence <= 0 {
		return NodeCheckpoint{}, ErrInjectedFailure
	}
	receiptDigest, digestErr := exactDigest(row.ProviderReceiptDigest)
	if digestErr != nil {
		return NodeCheckpoint{}, digestErr
	}
	checkpoint := NodeCheckpoint{
		AuthorityEpoch: uint64(row.AuthorityEpoch),
		Sequence:       uint64(row.AuthoritySequence),
		ReceiptDigest:  receiptDigest,
	}
	if checkpoint.Validate() != nil {
		return NodeCheckpoint{}, ErrInjectedFailure
	}
	return checkpoint, nil
}

type authorityFenceDatabaseRow struct {
	OperationID               uuid.UUID
	EffectKind                string
	ScopeKind                 string
	AuthorityEpoch            int64
	AuthoritySequence         int64
	ScopeDigest               []byte
	ProviderReservationDigest []byte
	EffectDigest              []byte
	ProviderStatus            string
	ProviderReceiptDigest     []byte
	DbSystemID                pgtype.Numeric
	DbTimeline                pgtype.Int8
	RequiredLsn               any
	AbortReason               pgtype.Text
	VisibilityState           string
	ReservedAt                pgtype.Timestamptz
	EffectBoundAt             pgtype.Timestamptz
	TerminalAt                pgtype.Timestamptz
}

func authorityFenceRowFromGet(row store.GetAuthorityFenceRow) authorityFenceDatabaseRow {
	return authorityFenceDatabaseRow{
		OperationID: row.OperationID, EffectKind: row.EffectKind, ScopeKind: row.ScopeKind,
		AuthorityEpoch: row.AuthorityEpoch, AuthoritySequence: row.AuthoritySequence,
		ScopeDigest: row.ScopeDigest, ProviderReservationDigest: row.ProviderReservationDigest,
		EffectDigest: row.EffectDigest, ProviderStatus: row.ProviderStatus,
		ProviderReceiptDigest: row.ProviderReceiptDigest, DbSystemID: row.DbSystemID,
		DbTimeline: row.DbTimeline, RequiredLsn: row.RequiredLsn, AbortReason: row.AbortReason,
		VisibilityState: row.VisibilityState, ReservedAt: row.ReservedAt,
		EffectBoundAt: row.EffectBoundAt, TerminalAt: row.TerminalAt,
	}
}

func authorityFenceRowFromForUpdate(row store.GetAuthorityFenceForUpdateRow) authorityFenceDatabaseRow {
	return authorityFenceDatabaseRow{
		OperationID: row.OperationID, EffectKind: row.EffectKind, ScopeKind: row.ScopeKind,
		AuthorityEpoch: row.AuthorityEpoch, AuthoritySequence: row.AuthoritySequence,
		ScopeDigest: row.ScopeDigest, ProviderReservationDigest: row.ProviderReservationDigest,
		EffectDigest: row.EffectDigest, ProviderStatus: row.ProviderStatus,
		ProviderReceiptDigest: row.ProviderReceiptDigest, DbSystemID: row.DbSystemID,
		DbTimeline: row.DbTimeline, RequiredLsn: row.RequiredLsn, AbortReason: row.AbortReason,
		VisibilityState: row.VisibilityState, ReservedAt: row.ReservedAt,
		EffectBoundAt: row.EffectBoundAt, TerminalAt: row.TerminalAt,
	}
}

func authorityFenceRowFromV7(row store.NodecontrolControlPlaneAuthorityFence) authorityFenceDatabaseRow {
	return authorityFenceDatabaseRow{
		OperationID: row.OperationID, EffectKind: row.EffectKind, ScopeKind: row.ScopeKind,
		AuthorityEpoch: row.AuthorityEpoch, AuthoritySequence: row.AuthoritySequence,
		ScopeDigest: row.ScopeDigest, ProviderReservationDigest: row.ProviderReservationDigest,
		EffectDigest: row.EffectDigest, ProviderStatus: row.ProviderStatus,
		ProviderReceiptDigest: row.ProviderReceiptDigest, DbSystemID: row.DbSystemID,
		DbTimeline: row.DbTimeline, RequiredLsn: row.RequiredLsn, AbortReason: row.AbortReason,
		VisibilityState: row.VisibilityState, ReservedAt: row.ReservedAt,
		EffectBoundAt: row.EffectBoundAt, TerminalAt: row.TerminalAt,
	}
}

type persistedOutcomeDatabaseRow struct {
	CommitmentJcs                  []byte             `json:"commitment_jcs"`
	CommitmentDigest               []byte             `json:"commitment_digest"`
	ProviderHeadJcs                []byte             `json:"provider_head_jcs"`
	ProviderHeadDigest             []byte             `json:"provider_head_digest"`
	CheckpointAnchorJcs            []byte             `json:"checkpoint_anchor_jcs"`
	CheckpointAnchorDigest         []byte             `json:"checkpoint_anchor_digest"`
	EffectReason                   pgtype.Text        `json:"effect_reason"`
	AttestationExpiresAt           pgtype.Timestamptz `json:"attestation_expires_at"`
	ActivationDeadline             pgtype.Timestamptz `json:"activation_deadline"`
	ExpectedProviderIdentityDigest []byte             `json:"expected_provider_identity_digest"`
	ActivationEvidenceJcs          []byte             `json:"activation_evidence_jcs"`
	ActivationEvidenceDigest       []byte             `json:"activation_evidence_digest"`
	EffectResolutionJcs            []byte             `json:"effect_resolution_jcs"`
	EffectResolutionDigest         []byte             `json:"effect_resolution_digest"`
}

func storedFenceFromV7Row(row store.NodecontrolControlPlaneAuthorityFence) (StoredFence, error) {
	if row.AuthorityProtocolProfile != string(contracts.AuthorityV7ProtocolProfileLegacyV6) &&
		row.AuthorityProtocolProfile != string(contracts.AuthorityV7ProtocolProfileClaimV1) {
		return StoredFence{}, ErrInjectedFailure
	}
	base := authorityFenceRowFromV7(row)
	var claim *AbortClaim
	if row.AuthorityProtocolProfile == string(contracts.AuthorityV7ProtocolProfileLegacyV6) {
		if row.AbortClaimedAt.Valid || row.ProtocolActivationID.Valid {
			return StoredFence{}, ErrInjectedFailure
		}
	} else {
		if !row.ProtocolActivationID.Valid || !validOperationID(row.ProtocolActivationID.UUID) ||
			(row.AbortClaimedAt.Valid != row.AbortReason.Valid) {
			return StoredFence{}, ErrInjectedFailure
		}
		if row.AbortClaimedAt.Valid {
			reason := AbortReason(row.AbortReason.String)
			if reason.Validate() != nil || row.AbortClaimedAt.Time.IsZero() || row.EffectDigest != nil {
				return StoredFence{}, ErrInjectedFailure
			}
			claim = &AbortClaim{Reason: reason, ClaimedAt: row.AbortClaimedAt.Time.UTC()}
			if row.ProviderStatus == "reserved" {
				base.AbortReason = pgtype.Text{}
			}
		}
	}
	record, err := authorityRecordFromRow(base)
	if err != nil {
		return StoredFence{}, fmt.Errorf("decode %s authority record: %w", row.AuthorityProtocolProfile, err)
	}
	return StoredFence{Record: record, AbortClaim: cloneAbortClaim(claim)}, nil
}

func (repository *PostgresRepository) persistedOutcomeForUpdate(
	ctx context.Context,
	queries *store.Queries,
	record Record,
) (*PersistedAuthorityEffectOutcome, error) {
	var row persistedOutcomeDatabaseRow
	var err error
	switch record.Kind {
	case EffectGrantCreate:
		value, queryErr := queries.LockEnrollmentGrantCreateOutcome(ctx, record.OperationID)
		row, err = persistedOutcomeDatabaseRow(value), queryErr
	case EffectGrantClaim:
		value, queryErr := queries.LockEnrollmentGrantClaimOutcome(ctx, uuid.NullUUID{UUID: record.OperationID, Valid: true})
		row, err = persistedOutcomeDatabaseRow(value), queryErr
	case EffectCertificateActivate:
		value, queryErr := queries.LockCertificateIssuanceActivationOutcome(ctx, record.OperationID)
		row, err = persistedOutcomeDatabaseRow(value), queryErr
	case EffectCertificateRevoke:
		value, queryErr := queries.LockCertificateRevocationOutcome(ctx, uuid.NullUUID{UUID: record.OperationID, Valid: true})
		row, err = persistedOutcomeDatabaseRow(value), queryErr
	case EffectIdentityEpochAdvance, EffectOperatorTransition:
		value, queryErr := queries.LockStateTransitionOutcome(ctx, uuid.NullUUID{UUID: record.OperationID, Valid: true})
		row, err = persistedOutcomeDatabaseRow(value), queryErr
	case EffectSecurityIncidentOpen:
		value, queryErr := queries.LockSecurityIncidentOpenOutcome(ctx, record.OperationID)
		row, err = persistedOutcomeDatabaseRow(value), queryErr
	case EffectSecurityIncidentResolve:
		value, queryErr := queries.LockSecurityIncidentResolveOutcome(ctx, uuid.NullUUID{UUID: record.OperationID, Valid: true})
		row, err = persistedOutcomeDatabaseRow(value), queryErr
	case EffectResourceEnvelopeActivate:
		value, queryErr := queries.LockResourceEnvelopeActivationOutcome(ctx, record.OperationID)
		row, err = persistedOutcomeDatabaseRow(value), queryErr
	case EffectDesiredActivate, EffectRecoveryActivate:
		value, queryErr := queries.LockStateSigningIntentActivationOutcome(ctx, record.OperationID)
		row, err = persistedOutcomeDatabaseRow(value), queryErr
	case EffectRootPublish, EffectMetadataPublish:
		value, queryErr := queries.LockRootMetadataPublishIntentActivationOutcome(ctx, record.OperationID)
		row, err = persistedOutcomeDatabaseRow(value), queryErr
	default:
		return nil, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		if record.TerminalReceipt != nil {
			return nil, ErrInjectedFailure
		}
		return nil, nil
	}
	if err != nil {
		return nil, repositoryDependencyError(ctx, err)
	}
	return persistedOutcomeFromDatabaseRow(record, row)
}

func persistedOutcomeFromDatabaseRow(record Record, row persistedOutcomeDatabaseRow) (*PersistedAuthorityEffectOutcome, error) {
	allNull := row.CommitmentJcs == nil && row.CommitmentDigest == nil && row.ProviderHeadJcs == nil &&
		row.ProviderHeadDigest == nil && row.CheckpointAnchorJcs == nil && row.CheckpointAnchorDigest == nil &&
		!row.EffectReason.Valid && !row.AttestationExpiresAt.Valid && !row.ActivationDeadline.Valid &&
		row.ExpectedProviderIdentityDigest == nil && row.ActivationEvidenceJcs == nil && row.ActivationEvidenceDigest == nil &&
		row.EffectResolutionJcs == nil && row.EffectResolutionDigest == nil
	if allNull {
		if record.TerminalReceipt != nil {
			return nil, ErrInjectedFailure
		}
		return nil, nil
	}
	commitmentDigest, err := exactJCSAndDigest(row.CommitmentJcs, row.CommitmentDigest, persistedCommitmentArtifact)
	if err != nil {
		return nil, err
	}
	terminalProofAbsent := row.ProviderHeadJcs == nil && row.ProviderHeadDigest == nil &&
		row.CheckpointAnchorJcs == nil && row.CheckpointAnchorDigest == nil && !row.EffectReason.Valid &&
		!row.AttestationExpiresAt.Valid && !row.ActivationDeadline.Valid && row.ExpectedProviderIdentityDigest == nil &&
		row.ActivationEvidenceJcs == nil && row.ActivationEvidenceDigest == nil && row.EffectResolutionJcs == nil && row.EffectResolutionDigest == nil
	if terminalProofAbsent {
		if record.TerminalReceipt != nil {
			return nil, ErrInjectedFailure
		}
		_ = commitmentDigest
		return nil, nil
	}
	if record.TerminalReceipt == nil || !row.EffectReason.Valid {
		return nil, ErrInjectedFailure
	}
	providerHeadDigest, err := exactJCSAndDigest(row.ProviderHeadJcs, row.ProviderHeadDigest, persistedProviderHeadArtifact)
	if err != nil {
		return nil, err
	}
	evidenceDigest, err := exactJCSAndDigest(row.ActivationEvidenceJcs, row.ActivationEvidenceDigest, persistedEvidenceArtifact)
	if err != nil {
		return nil, err
	}
	resolutionDigest, err := exactJCSAndDigest(row.EffectResolutionJcs, row.EffectResolutionDigest, persistedResolutionArtifact)
	if err != nil {
		return nil, err
	}
	reason := AuthorityEffectReason(row.EffectReason.String)
	if !reason.valid() {
		return nil, ErrInjectedFailure
	}
	var checkpointDigest contracts.Digest
	var attestationExpiresAt, activationDeadline time.Time
	var expectedProviderIdentity contracts.Digest
	hasCheckpoint := row.CheckpointAnchorJcs != nil || row.CheckpointAnchorDigest != nil
	hasTime := row.AttestationExpiresAt.Valid || row.ActivationDeadline.Valid || row.ExpectedProviderIdentityDigest != nil
	if hasCheckpoint {
		checkpointDigest, err = exactJCSAndDigest(row.CheckpointAnchorJcs, row.CheckpointAnchorDigest, persistedCheckpointArtifact)
		if err != nil || hasTime || reason != EffectReasonSuperseded {
			return nil, ErrInjectedFailure
		}
	}
	if hasTime {
		if reason == EffectReasonSuperseded || !row.AttestationExpiresAt.Valid || !row.ActivationDeadline.Valid ||
			row.AttestationExpiresAt.Time.IsZero() || row.ActivationDeadline.Time.IsZero() {
			return nil, ErrInjectedFailure
		}
		if reason == EffectReasonActivationDeadlineExpired && !row.ActivationDeadline.Time.Before(row.AttestationExpiresAt.Time) {
			return nil, ErrInjectedFailure
		}
		expectedProviderIdentity, err = exactDigest(row.ExpectedProviderIdentityDigest)
		if err != nil {
			return nil, err
		}
		attestationExpiresAt = row.AttestationExpiresAt.Time.UTC()
		activationDeadline = row.ActivationDeadline.Time.UTC()
	}
	if !hasCheckpoint && !hasTime && reason == EffectReasonNone {
		return nil, ErrInjectedFailure
	}
	return &PersistedAuthorityEffectOutcome{
		CommitmentJCS: cloneBytes(row.CommitmentJcs), CommitmentDigest: commitmentDigest,
		Receipt:         cloneReceipt(*record.TerminalReceipt),
		ProviderHeadJCS: cloneBytes(row.ProviderHeadJcs), ProviderHeadDigest: providerHeadDigest,
		CheckpointAnchorJCS: cloneBytes(row.CheckpointAnchorJcs), CheckpointAnchorDigest: checkpointDigest,
		Reason: reason, AttestationExpiresAt: attestationExpiresAt, ActivationDeadline: activationDeadline,
		ExpectedProviderIdentityDigest: expectedProviderIdentity,
		EvidenceJCS:                    cloneBytes(row.ActivationEvidenceJcs), EvidenceDigest: evidenceDigest,
		ResolutionJCS: cloneBytes(row.EffectResolutionJcs), ResolutionDigest: resolutionDigest,
	}, nil
}

type persistedAuthorityArtifactKind uint8

const (
	persistedCommitmentArtifact persistedAuthorityArtifactKind = iota + 1
	persistedProviderHeadArtifact
	persistedCheckpointArtifact
	persistedEvidenceArtifact
	persistedResolutionArtifact
)

func exactJCSAndDigest(jcsBytes, digestBytes []byte, kind persistedAuthorityArtifactKind) (contracts.Digest, error) {
	if len(jcsBytes) < 1 || len(jcsBytes) > authorityCanonicalMaximumBytes {
		return contracts.Digest{}, ErrInjectedFailure
	}
	stored, err := exactDigest(digestBytes)
	if err != nil {
		return contracts.Digest{}, err
	}
	var computed contracts.Digest
	switch kind {
	case persistedCommitmentArtifact:
		value, parseErr := ParseAuthorityEffectCommitment(jcsBytes)
		if parseErr != nil {
			return contracts.Digest{}, ErrInjectedFailure
		}
		computed = value.Digest()
	case persistedProviderHeadArtifact:
		value, parseErr := ParseAuthorityProviderHeadSnapshot(jcsBytes)
		if parseErr != nil {
			return contracts.Digest{}, ErrInjectedFailure
		}
		computed = value.Digest()
	case persistedCheckpointArtifact:
		value, parseErr := ParseAuthorityCheckpointAnchor(jcsBytes)
		if parseErr != nil {
			return contracts.Digest{}, ErrInjectedFailure
		}
		computed = value.Digest()
	case persistedEvidenceArtifact:
		value, parseErr := ParseActivationDecisionEvidence(jcsBytes)
		if parseErr != nil {
			return contracts.Digest{}, ErrInjectedFailure
		}
		computed = value.Digest()
	case persistedResolutionArtifact:
		value, parseErr := ParseAuthorityEffectResolution(jcsBytes)
		if parseErr != nil {
			return contracts.Digest{}, ErrInjectedFailure
		}
		computed = value.Digest()
	default:
		return contracts.Digest{}, ErrInjectedFailure
	}
	if computed != stored {
		return contracts.Digest{}, ErrInjectedFailure
	}
	return stored, nil
}

func cloneStoredFence(value StoredFence) StoredFence {
	result := value
	result.Record = cloneRecord(value.Record)
	result.AbortClaim = cloneAbortClaim(value.AbortClaim)
	if value.PersistedOutcome != nil {
		outcome := *value.PersistedOutcome
		outcome.CommitmentJCS = cloneBytes(value.PersistedOutcome.CommitmentJCS)
		outcome.Receipt = cloneReceipt(value.PersistedOutcome.Receipt)
		outcome.ProviderHeadJCS = cloneBytes(value.PersistedOutcome.ProviderHeadJCS)
		outcome.CheckpointAnchorJCS = cloneBytes(value.PersistedOutcome.CheckpointAnchorJCS)
		outcome.EvidenceJCS = cloneBytes(value.PersistedOutcome.EvidenceJCS)
		outcome.ResolutionJCS = cloneBytes(value.PersistedOutcome.ResolutionJCS)
		result.PersistedOutcome = &outcome
	}
	return result
}

func cloneAbortClaim(value *AbortClaim) *AbortClaim {
	if value == nil {
		return nil
	}
	result := *value
	return &result
}

func cloneBytes(value []byte) []byte {
	return append([]byte(nil), value...)
}

func canonicalFreshImportTopologyProjection(value contracts.FreshImportTopologyProjectionV1) ([]byte, contracts.Digest, error) {
	objects := make([]map[string]any, len(value.Objects))
	for index := range value.Objects {
		var payload any
		if err := json.Unmarshal(value.Objects[index].NormalizedPayload, &payload); err != nil {
			return nil, contracts.Digest{}, err
		}
		objects[index] = map[string]any{
			"object_type":        value.Objects[index].ObjectType,
			"canonical_key":      value.Objects[index].CanonicalKey,
			"normalized_payload": payload,
		}
	}
	body := map[string]any{
		"projection_version":              strconv.FormatUint(value.ProjectionVersion, 10),
		"target_activation_id":            value.TargetActivationID.String(),
		"target_deployment_id":            value.TargetDeploymentID.String(),
		"target_database_identity_digest": hex.EncodeToString(value.TargetDatabaseIdentityDigest[:]),
		"normalized_catalog_digest":       hex.EncodeToString(value.NormalizedCatalogDigest[:]),
		"object_count":                    strconv.FormatUint(value.ObjectCount, 10),
		"objects":                         objects,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, contracts.Digest{}, err
	}
	canonical, err := jcs.Transform(raw)
	if err != nil {
		return nil, contracts.Digest{}, err
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte("talenro.c12.fresh-import-topology-projection.v1"))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(canonical)
	var digest contracts.Digest
	copy(digest[:], hash.Sum(nil))
	return canonical, digest, nil
}

func freshImportProjectionObjectsJSON(values []contracts.FreshImportTopologyProjectionObjectV1) (json.RawMessage, error) {
	objects := make([]map[string]any, len(values))
	for index := range values {
		var body any
		if err := json.Unmarshal(values[index].NormalizedPayload, &body); err != nil {
			return nil, err
		}
		objects[index] = map[string]any{
			"object_type":   values[index].ObjectType,
			"canonical_key": values[index].CanonicalKey,
			"body":          body,
		}
	}
	return json.Marshal(objects)
}

func freshRestoreImportApplicationFromRow(row store.NodecontrolControlPlaneAuthorityFreshRestoreImportApplication) (contracts.FreshRestoreImportApplicationV1, error) {
	if row.StagingImportCapabilityRecoveryIntentDigestOrNull != nil ||
		row.StagingImportCapabilityRecoveryApplicationDigestOrNull != nil ||
		row.ImportedObjectCount <= 0 || !row.AppliedAt.Valid || row.AppliedAt.Time.IsZero() {
		return contracts.FreshRestoreImportApplicationV1{}, ErrInjectedFailure
	}
	txid, ok := uint64FromNumeric(row.DatabaseTransactionID)
	if !ok {
		return contracts.FreshRestoreImportApplicationV1{}, ErrInjectedFailure
	}
	capabilityDigest, err := exactDigest(row.StagingImportCapabilityDigest)
	if err != nil {
		return contracts.FreshRestoreImportApplicationV1{}, err
	}
	manifestDigest, err := exactDigest(row.ManifestDigest)
	if err != nil {
		return contracts.FreshRestoreImportApplicationV1{}, err
	}
	currentIdentity, err := exactDigest(row.CurrentDatabaseIdentityDigest)
	if err != nil {
		return contracts.FreshRestoreImportApplicationV1{}, err
	}
	lineage, err := exactDigest(row.DatabaseTimelineLineageChainDigest)
	if err != nil {
		return contracts.FreshRestoreImportApplicationV1{}, err
	}
	incarnation, err := exactDigest(row.TargetDatabaseIncarnationRegistrationDigest)
	if err != nil {
		return contracts.FreshRestoreImportApplicationV1{}, err
	}
	runtimeChain, err := exactDigest(row.RuntimeRebindChainDigest)
	if err != nil {
		return contracts.FreshRestoreImportApplicationV1{}, err
	}
	runtimeBinding, err := exactDigest(row.RuntimeInstanceBindingDigest)
	if err != nil {
		return contracts.FreshRestoreImportApplicationV1{}, err
	}
	exclusionLease, err := exactDigest(row.StagingExclusionLeaseDigest)
	if err != nil {
		return contracts.FreshRestoreImportApplicationV1{}, err
	}
	acquisitionHead, err := exactDigest(row.AcquisitionLockedProviderHeadDigest)
	if err != nil {
		return contracts.FreshRestoreImportApplicationV1{}, err
	}
	routeClosed, err := exactDigest(row.DatabaseRouteClosedDigest)
	if err != nil {
		return contracts.FreshRestoreImportApplicationV1{}, err
	}
	preInventory, err := exactDigest(row.PreImportInventoryDigest)
	if err != nil {
		return contracts.FreshRestoreImportApplicationV1{}, err
	}
	postInventory, err := exactDigest(row.PostImportInventoryDigest)
	if err != nil {
		return contracts.FreshRestoreImportApplicationV1{}, err
	}
	completeSet, err := exactDigest(row.CompleteNodeSetDigest)
	if err != nil {
		return contracts.FreshRestoreImportApplicationV1{}, err
	}
	forbiddenZero, err := exactDigest(row.ForbiddenStateZeroDigest)
	if err != nil {
		return contracts.FreshRestoreImportApplicationV1{}, err
	}
	snapshot, err := exactDigest(row.TransactionSnapshotDigest)
	if err != nil {
		return contracts.FreshRestoreImportApplicationV1{}, err
	}
	application := contracts.FreshRestoreImportApplicationV1{
		SingleUseApplyID: row.SingleUseApplyID, StagingImportCapabilityDigest: capabilityDigest,
		ManifestDigest: manifestDigest, TargetActivationID: row.TargetActivationID,
		CurrentDatabaseIdentityDigest: currentIdentity, DatabaseTimelineLineageChainDigest: lineage,
		TargetDatabaseIncarnationRegistrationDigest: incarnation, RuntimeRebindChainDigest: runtimeChain,
		RuntimeInstanceBindingDigest: runtimeBinding, StagingExclusionLeaseDigest: exclusionLease,
		AcquisitionLockedProviderHeadDigest: acquisitionHead, DatabaseRouteClosedDigest: routeClosed,
		PreImportInventoryDigest: preInventory, PostImportInventoryDigest: postInventory,
		ImportedObjectCount: uint64(row.ImportedObjectCount), CompleteNodeSetDigest: completeSet,
		ForbiddenStateZeroDigest: forbiddenZero, DatabaseTransactionID: txid,
		TransactionSnapshotDigest: snapshot, AppliedAt: row.AppliedAt.Time.UTC(),
	}
	if application.Validate() != nil {
		return contracts.FreshRestoreImportApplicationV1{}, ErrInjectedFailure
	}
	return application, nil
}

func authorityRecordFromRow(row authorityFenceDatabaseRow) (Record, error) {
	reservation, err := reservationFromRow(
		row.OperationID,
		row.EffectKind,
		row.ScopeKind,
		row.AuthorityEpoch,
		row.AuthoritySequence,
		row.ScopeDigest,
		row.ProviderReservationDigest,
	)
	if err != nil || !row.ReservedAt.Valid || row.ReservedAt.Time.IsZero() {
		return Record{}, ErrInjectedFailure
	}
	effectDigest, point, effectBoundAt, bindingErr := optionalBinding(
		row.EffectDigest,
		row.DbSystemID,
		row.DbTimeline,
		row.RequiredLsn,
		row.EffectBoundAt,
	)
	if bindingErr != nil || (effectBoundAt != nil && effectBoundAt.Before(row.ReservedAt.Time)) {
		return Record{}, ErrInjectedFailure
	}
	record := Record{
		Reservation:        reservation,
		BoundEffectDigest:  effectDigest,
		BoundDatabasePoint: point,
	}
	switch row.ProviderStatus {
	case "reserved":
		if row.VisibilityState != "fence_pending" || row.ProviderReceiptDigest != nil ||
			row.AbortReason.Valid || row.TerminalAt.Valid {
			return Record{}, ErrInjectedFailure
		}
	case "committed":
		if row.VisibilityState != "active" || effectDigest == nil || !row.TerminalAt.Valid ||
			row.TerminalAt.Time.IsZero() || row.AbortReason.Valid || row.TerminalAt.Time.Before(row.ReservedAt.Time) {
			return Record{}, ErrInjectedFailure
		}
		receiptDigest, digestErr := exactDigest(row.ProviderReceiptDigest)
		if digestErr != nil {
			return Record{}, digestErr
		}
		receipt := Receipt{
			Reservation:   reservation,
			EffectDigest:  cloneDigestPointer(effectDigest),
			DatabasePoint: cloneDatabasePointPointer(point),
			Status:        StatusCommitted,
			ReceiptDigest: receiptDigest,
		}
		if receipt.Validate() != nil {
			return Record{}, ErrInjectedFailure
		}
		record.TerminalReceipt = &receipt
	case "aborted":
		if row.VisibilityState != "aborted" || !row.TerminalAt.Valid || row.TerminalAt.Time.IsZero() ||
			row.TerminalAt.Time.Before(row.ReservedAt.Time) || !row.AbortReason.Valid {
			return Record{}, ErrInjectedFailure
		}
		reason := AbortReason(row.AbortReason.String)
		if reason.Validate() != nil {
			return Record{}, ErrInjectedFailure
		}
		receiptDigest, digestErr := exactDigest(row.ProviderReceiptDigest)
		if digestErr != nil {
			return Record{}, digestErr
		}
		receipt, receiptErr := abortedReceiptFromRow(reservation, effectDigest, point, reason, receiptDigest)
		if receiptErr != nil {
			return Record{}, receiptErr
		}
		record.TerminalReceipt = &receipt
	default:
		return Record{}, ErrInjectedFailure
	}
	return record, nil
}

func databaseHeadFromRow(row store.GetAuthorityFenceHeadRow) (DatabaseHead, error) {
	if row.AuthorityEpoch == 0 && row.RecordCount == 0 && row.LatestReservedSequence == 0 &&
		row.LatestCommittedSequence == 0 && row.PendingCount == 0 {
		if row.ProviderReservationDigest != nil || row.LatestCommittedOperationID.Valid ||
			row.ProviderReceiptDigest != nil || row.DbSystemID.Valid || row.DbTimeline.Valid ||
			row.RequiredLsn != nil || row.HasSequenceGap {
			return DatabaseHead{}, ErrInjectedFailure
		}
		return DatabaseHead{}, nil
	}
	if row.AuthorityEpoch <= 0 || row.RecordCount <= 0 || row.LatestReservedSequence <= 0 ||
		row.LatestCommittedSequence < 0 || row.PendingCount < 0 || row.PendingCount > row.RecordCount ||
		row.LatestCommittedSequence > row.LatestReservedSequence ||
		row.HasSequenceGap != (row.RecordCount != row.LatestReservedSequence) {
		return DatabaseHead{}, ErrInjectedFailure
	}
	reservationDigest, err := exactDigest(row.ProviderReservationDigest)
	if err != nil {
		return DatabaseHead{}, err
	}
	head := DatabaseHead{
		Epoch:                   uint64(row.AuthorityEpoch),
		RecordCount:             uint64(row.RecordCount),
		LatestReservedSequence:  uint64(row.LatestReservedSequence),
		LatestCommittedSequence: uint64(row.LatestCommittedSequence),
		PendingCount:            uint64(row.PendingCount),
		HasSequenceGap:          row.HasSequenceGap,
		LatestReservationDigest: reservationDigest,
	}
	if row.LatestCommittedSequence == 0 {
		if row.LatestCommittedOperationID.Valid || row.ProviderReceiptDigest != nil ||
			row.DbSystemID.Valid || row.DbTimeline.Valid || row.RequiredLsn != nil {
			return DatabaseHead{}, ErrInjectedFailure
		}
		return head, nil
	}
	if !row.LatestCommittedOperationID.Valid || !validOperationID(row.LatestCommittedOperationID.UUID) {
		return DatabaseHead{}, ErrInjectedFailure
	}
	receiptDigest, digestErr := exactDigest(row.ProviderReceiptDigest)
	if digestErr != nil {
		return DatabaseHead{}, digestErr
	}
	point, pointErr := databasePoint(row.DbSystemID, row.DbTimeline, row.RequiredLsn)
	if pointErr != nil {
		return DatabaseHead{}, pointErr
	}
	head.LatestCommittedOperationID = row.LatestCommittedOperationID.UUID
	head.LatestCommittedReceiptDigest = receiptDigest
	head.LatestCommittedDatabasePoint = &point
	return head, nil
}

func pendingFenceFromRow(row store.ListPendingAuthorityFencesRow) (PendingFence, error) {
	reservation, err := reservationFromRow(
		row.OperationID,
		row.EffectKind,
		row.ScopeKind,
		row.AuthorityEpoch,
		row.AuthoritySequence,
		row.ScopeDigest,
		row.ProviderReservationDigest,
	)
	if err != nil || !row.ReservedAt.Valid || row.ReservedAt.Time.IsZero() {
		return PendingFence{}, ErrInjectedFailure
	}
	effectDigest, point, effectBoundAt, bindingErr := optionalBinding(
		row.EffectDigest,
		row.DbSystemID,
		row.DbTimeline,
		row.RequiredLsn,
		row.EffectBoundAt,
	)
	if bindingErr != nil || (effectBoundAt != nil && effectBoundAt.Before(row.ReservedAt.Time)) {
		return PendingFence{}, ErrInjectedFailure
	}
	return PendingFence{
		OperationID:        reservation.OperationID,
		Kind:               reservation.Kind,
		ScopeKind:          reservation.ScopeKind,
		ScopeDigest:        reservation.ScopeDigest,
		Epoch:              reservation.Epoch,
		Sequence:           reservation.Sequence,
		ReservationDigest:  reservation.ReservationDigest,
		BoundEffectDigest:  cloneDigestPointer(effectDigest),
		BoundDatabasePoint: cloneDatabasePointPointer(point),
		ReservedAt:         row.ReservedAt.Time.UTC(),
		EffectBoundAt:      cloneTimePointer(effectBoundAt),
	}, nil
}

func reservationFromRow(
	operationID uuid.UUID,
	effectKind string,
	scopeKind string,
	epoch int64,
	sequence int64,
	scopeBytes []byte,
	reservationBytes []byte,
) (Reservation, error) {
	if epoch <= 0 || sequence <= 0 {
		return Reservation{}, ErrInjectedFailure
	}
	scopeDigest, err := exactDigest(scopeBytes)
	if err != nil {
		return Reservation{}, err
	}
	reservationDigest, err := exactDigest(reservationBytes)
	if err != nil {
		return Reservation{}, err
	}
	reservation := Reservation{
		OperationID:       operationID,
		Kind:              EffectKind(effectKind),
		ScopeKind:         ScopeKind(scopeKind),
		ScopeDigest:       scopeDigest,
		Epoch:             uint64(epoch),
		Sequence:          uint64(sequence),
		ReservationDigest: reservationDigest,
	}
	if reservation.Validate() != nil {
		return Reservation{}, ErrInjectedFailure
	}
	return reservation, nil
}

func optionalBinding(
	effectBytes []byte,
	systemID pgtype.Numeric,
	timeline pgtype.Int8,
	requiredLSN any,
	effectBoundAt pgtype.Timestamptz,
) (*contracts.Digest, *DatabasePoint, *time.Time, error) {
	present := []bool{effectBytes != nil, systemID.Valid, timeline.Valid, requiredLSN != nil, effectBoundAt.Valid}
	count := 0
	for _, value := range present {
		if value {
			count++
		}
	}
	if count == 0 {
		return nil, nil, nil, nil
	}
	if count != len(present) || effectBoundAt.Time.IsZero() {
		return nil, nil, nil, ErrInjectedFailure
	}
	effectDigest, err := exactDigest(effectBytes)
	if err != nil {
		return nil, nil, nil, err
	}
	point, err := databasePoint(systemID, timeline, requiredLSN)
	if err != nil {
		return nil, nil, nil, err
	}
	boundAt := effectBoundAt.Time.UTC()
	return &effectDigest, &point, &boundAt, nil
}

func abortedReceiptFromRow(
	reservation Reservation,
	effectDigest *contracts.Digest,
	point *DatabasePoint,
	reason AbortReason,
	receiptDigest contracts.Digest,
) (Receipt, error) {
	receipt := Receipt{
		Reservation:   reservation,
		Status:        StatusAborted,
		AbortReason:   &reason,
		ReceiptDigest: receiptDigest,
	}
	if receipt.Validate() == nil {
		return receipt, nil
	}
	if effectDigest == nil || point == nil {
		return Receipt{}, ErrInjectedFailure
	}
	receipt.EffectDigest = cloneDigestPointer(effectDigest)
	receipt.DatabasePoint = cloneDatabasePointPointer(point)
	if receipt.Validate() != nil {
		return Receipt{}, ErrInjectedFailure
	}
	return receipt, nil
}

func databasePoint(systemID pgtype.Numeric, timeline pgtype.Int8, requiredLSN any) (DatabasePoint, error) {
	convertedSystemID, ok := uint64FromNumeric(systemID)
	if !ok || convertedSystemID == 0 || !timeline.Valid || timeline.Int64 <= 0 || timeline.Int64 > math.MaxUint32 {
		return DatabasePoint{}, ErrInjectedFailure
	}
	lsn, ok := walPositionFromDatabase(requiredLSN)
	if !ok {
		return DatabasePoint{}, ErrInjectedFailure
	}
	point := DatabasePoint{SystemID: convertedSystemID, Timeline: uint32(timeline.Int64), RequiredLSN: lsn}
	if point.Validate() != nil {
		return DatabasePoint{}, ErrInjectedFailure
	}
	return point, nil
}

func walPositionFromDatabase(value any) (WALPosition, bool) {
	var position WALPosition
	switch typed := value.(type) {
	case string:
		position = WALPosition(typed)
	case []byte:
		position = WALPosition(string(typed))
	case pgtype.Text:
		if !typed.Valid {
			return "", false
		}
		position = WALPosition(typed.String)
	case WALPosition:
		position = typed
	default:
		return "", false
	}
	canonical, err := canonicalWALPosition(position)
	return canonical, err == nil && canonical == position
}

func uint64FromNumeric(value pgtype.Numeric) (uint64, bool) {
	if !value.Valid || value.Int == nil || value.NaN || value.InfinityModifier != pgtype.Finite ||
		value.Exp < -20 || value.Exp > 20 {
		return 0, false
	}
	integer := new(big.Int).Set(value.Int)
	if value.Exp > 0 {
		integer.Mul(integer, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(value.Exp)), nil))
	}
	if value.Exp < 0 {
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-value.Exp)), nil)
		remainder := new(big.Int)
		integer.QuoRem(integer, divisor, remainder)
		if remainder.Sign() != 0 {
			return 0, false
		}
	}
	if integer.Sign() <= 0 || integer.BitLen() > 64 {
		return 0, false
	}
	return integer.Uint64(), true
}

func numericFromUint64(value uint64) pgtype.Numeric {
	return pgtype.Numeric{Int: new(big.Int).SetUint64(value), Valid: true}
}

func exactDigest(value []byte) (contracts.Digest, error) {
	if len(value) != len(contracts.Digest{}) {
		return contracts.Digest{}, ErrInjectedFailure
	}
	var digest contracts.Digest
	copy(digest[:], value)
	if digest == (contracts.Digest{}) {
		return contracts.Digest{}, ErrInjectedFailure
	}
	return digest, nil
}

func digestBytes(value contracts.Digest) []byte {
	result := make([]byte, len(value))
	copy(result, value[:])
	return result
}

func requiredTimestamp(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: canonicalDatabaseTimestamp(value), Valid: true}
}

func canonicalDatabaseTimestamp(value time.Time) time.Time {
	return value.UTC().Truncate(time.Microsecond)
}

func reservationMatchesRow(reservation Reservation, row authorityFenceDatabaseRow) bool {
	return reservation.OperationID == row.OperationID && string(reservation.Kind) == row.EffectKind &&
		string(reservation.ScopeKind) == row.ScopeKind && int64(reservation.Epoch) == row.AuthorityEpoch &&
		int64(reservation.Sequence) == row.AuthoritySequence &&
		bytesEqualDigest(row.ScopeDigest, reservation.ScopeDigest) &&
		bytesEqualDigest(row.ProviderReservationDigest, reservation.ReservationDigest)
}

func bytesEqualDigest(value []byte, digest contracts.Digest) bool {
	if len(value) != len(digest) {
		return false
	}
	for index := range value {
		if value[index] != digest[index] {
			return false
		}
	}
	return true
}

func receiptsEqual(left Receipt, right Receipt) bool {
	return left.Reservation == right.Reservation && left.Status == right.Status &&
		left.ReceiptDigest == right.ReceiptDigest && equalOptionalDigest(left.EffectDigest, right.EffectDigest) &&
		equalOptionalDatabasePoint(left.DatabasePoint, right.DatabasePoint) &&
		equalOptionalAbortReason(left.AbortReason, right.AbortReason)
}

func equalOptionalAbortReason(left *AbortReason, right *AbortReason) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func cloneTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func repositoryCallError(ctx context.Context, repository *PostgresRepository) error {
	if repository == nil || nilAuthorityRepositoryValue(repository.database) {
		return ErrInvalidArgument
	}
	return contextError(ctx)
}

func repositoryInputError(ctx context.Context) error {
	if cancellationErr := repositoryCancellationError(ctx, nil); cancellationErr != nil {
		return cancellationErr
	}
	return ErrInvalidArgument
}

func repositoryLookupError(ctx context.Context, err error) error {
	if cancellationErr := repositoryCancellationError(ctx, err); cancellationErr != nil {
		return cancellationErr
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err == nil {
		return nil
	}
	return ErrInjectedFailure
}

func repositoryDependencyError(ctx context.Context, err error) error {
	if cancellationErr := repositoryCancellationError(ctx, err); cancellationErr != nil {
		return cancellationErr
	}
	if err == nil {
		return nil
	}
	return ErrInjectedFailure
}

func repositoryCancellationError(ctx context.Context, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) ||
		(ctx != nil && ctx.Err() != nil) {
		return ErrCanceled
	}
	return nil
}

func nilAuthorityRepositoryValue(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
