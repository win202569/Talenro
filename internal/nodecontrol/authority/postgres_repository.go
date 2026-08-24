package authority

import (
	"context"
	"errors"
	"math"
	"math/big"
	"reflect"
	"time"

	"github.com/google/uuid"
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
		var postgresError *pgconn.PgError
		if (ctx == nil || ctx.Err() == nil) && errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return ErrConflict
		}
		return repositoryDependencyError(ctx, err)
	}
	switch rows {
	case 1:
		return nil
	case 0:
		current, getErr := queries.GetAuthorityFenceForUpdate(ctx, reservation.OperationID)
		if getErr != nil {
			if errors.Is(getErr, pgx.ErrNoRows) {
				return ErrConflict
			}
			return repositoryDependencyError(ctx, getErr)
		}
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
	current, err := queries.GetAuthorityFenceForUpdate(ctx, operationID)
	if err != nil {
		return repositoryLookupError(ctx, err)
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
	current, err := queries.GetAuthorityFenceForUpdate(ctx, receipt.OperationID)
	if err != nil {
		return repositoryLookupError(ctx, err)
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
	current, err := queries.GetAuthorityFenceForUpdate(ctx, receipt.OperationID)
	if err != nil {
		return repositoryLookupError(ctx, err)
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
	return authorityRecordFromRow(row)
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
		if errors.Is(err, pgx.ErrNoRows) {
			return NodeCheckpoint{AuthorityEpoch: epoch}, nil
		}
		return NodeCheckpoint{}, repositoryDependencyError(ctx, err)
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

func authorityRecordFromRow(row store.NodecontrolControlPlaneAuthorityFence) (Record, error) {
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

func reservationMatchesRow(reservation Reservation, row store.NodecontrolControlPlaneAuthorityFence) bool {
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
	if ctx != nil && ctx.Err() != nil {
		return ErrCanceled
	}
	return ErrInvalidArgument
}

func repositoryLookupError(ctx context.Context, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return repositoryDependencyError(ctx, err)
}

func repositoryDependencyError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) ||
		(ctx != nil && ctx.Err() != nil) {
		return ErrCanceled
	}
	return ErrInjectedFailure
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
