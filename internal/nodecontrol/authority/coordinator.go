package authority

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/store"
)

type EffectState string

const (
	EffectAbsent    EffectState = "absent"
	EffectPrepared  EffectState = "prepared"
	EffectCommitted EffectState = "committed"
	EffectTerminal  EffectState = "terminal"
)

type ResolvedEffect struct {
	Kind         EffectKind
	ScopeKind    ScopeKind
	ScopeDigest  contracts.Digest
	EffectDigest contracts.Digest
	State        EffectState
}

type EffectResolver interface {
	ResolveAuthorityEffect(context.Context, uuid.UUID) (ResolvedEffect, error)
}

type CoordinatorFinalizeRequest struct {
	OperationID  uuid.UUID
	EffectDigest contracts.Digest
}

var ErrAuthorityUnavailable = errors.New("nodecontrol authority: unavailable")

type ReadinessReason string

const (
	ReadinessReady                    ReadinessReason = "ready"
	ReadinessProviderUnavailable      ReadinessReason = "provider_unavailable"
	ReadinessDatabaseUnavailable      ReadinessReason = "database_unavailable"
	ReadinessEffectUnavailable        ReadinessReason = "effect_unavailable"
	ReadinessDatabaseIdentityMismatch ReadinessReason = "database_identity_mismatch"
	ReadinessDatabaseTimelineBehind   ReadinessReason = "database_timeline_behind"
	ReadinessDatabaseWALBehind        ReadinessReason = "database_wal_behind"
	ReadinessDatabaseBehindProvider   ReadinessReason = "database_behind_provider"
	ReadinessDatabaseAheadProvider    ReadinessReason = "database_ahead_provider"
	ReadinessDatabaseSequenceGap      ReadinessReason = "database_sequence_gap"
	ReadinessReservationMismatch      ReadinessReason = "reservation_mismatch"
	ReadinessCommittedMismatch        ReadinessReason = "committed_mismatch"
	ReadinessPendingCountMismatch     ReadinessReason = "pending_count_mismatch"
	ReadinessPendingMismatch          ReadinessReason = "pending_mismatch"
	ReadinessPendingUnresolved        ReadinessReason = "pending_unresolved"
)

type Readiness struct {
	ProviderHead Head
	DatabaseHead DatabaseHead
	Ready        bool
	Reason       ReadinessReason
}

type Coordinator struct {
	provider    Provider
	repository  Repository
	resolver    EffectResolver
	clock       securitykit.Clock
	transaction authorityTransactionRunner
}

type authorityTransactionRunner interface {
	withAuthorityTransaction(context.Context, func(store.DBTX) error) error
}

type authorityTransactionBeginner interface {
	Begin(context.Context) (pgx.Tx, error)
}

func NewCoordinator(provider Provider, repository Repository, resolver EffectResolver, clock securitykit.Clock) (*Coordinator, error) {
	transaction, ok := repository.(authorityTransactionRunner)
	if nilAuthorityRepositoryValue(provider) || nilAuthorityRepositoryValue(repository) ||
		nilAuthorityRepositoryValue(resolver) || nilAuthorityRepositoryValue(clock) || !ok || nilAuthorityRepositoryValue(transaction) {
		return nil, ErrInvalidArgument
	}
	return &Coordinator{
		provider:    provider,
		repository:  repository,
		resolver:    resolver,
		clock:       clock,
		transaction: transaction,
	}, nil
}

func (coordinator *Coordinator) Reserve(ctx context.Context, request ReserveRequest) (Reservation, error) {
	if err := coordinatorCallError(ctx, coordinator); err != nil || request.Validate() != nil {
		return Reservation{}, coordinatorInputError(ctx)
	}
	reservation, err := coordinator.provider.Reserve(ctx, request)
	if err != nil {
		return Reservation{}, coordinatorDependencyError(ctx, err)
	}
	if reservation.Validate() != nil || reservation.OperationID != request.OperationID || reservation.Kind != request.Kind ||
		reservation.ScopeKind != request.ScopeKind || reservation.ScopeDigest != request.ScopeDigest {
		return Reservation{}, ErrInjectedFailure
	}
	return reservation, nil
}

func (coordinator *Coordinator) Finalize(ctx context.Context, request CoordinatorFinalizeRequest) (Receipt, error) {
	if err := coordinatorCallError(ctx, coordinator); err != nil || !validOperationID(request.OperationID) ||
		request.EffectDigest == (contracts.Digest{}) {
		return Receipt{}, coordinatorInputError(ctx)
	}
	record, err := coordinator.repository.Get(ctx, request.OperationID)
	if err != nil {
		return Receipt{}, coordinatorDependencyError(ctx, err)
	}
	resolved, err := coordinator.resolve(ctx, request.OperationID)
	if err != nil {
		return Receipt{}, err
	}
	if resolved.State != EffectCommitted || !resolvedMatchesReservation(resolved, record.Reservation) ||
		resolved.EffectDigest != request.EffectDigest {
		return Receipt{}, ErrConflict
	}
	if record.TerminalReceipt != nil {
		if record.TerminalReceipt.Status != StatusCommitted {
			return Receipt{}, ErrTerminalConflict
		}
		if record.TerminalReceipt.EffectDigest == nil || *record.TerminalReceipt.EffectDigest != request.EffectDigest {
			return Receipt{}, ErrConflict
		}
		return cloneReceipt(*record.TerminalReceipt), nil
	}

	point, err := coordinator.ensureBound(ctx, record, request.EffectDigest)
	if err != nil {
		return Receipt{}, err
	}
	providerRequest := FinalizeRequest{
		OperationID:  request.OperationID,
		EffectDigest: request.EffectDigest,
		DBSystemID:   point.SystemID,
		DBTimeline:   point.Timeline,
		RequiredLSN:  point.RequiredLSN,
	}
	receipt, err := coordinator.provider.Finalize(ctx, providerRequest)
	if err != nil {
		return Receipt{}, coordinatorDependencyError(ctx, err)
	}
	if !validCommittedReceiptFor(receipt, record.Reservation, request.EffectDigest, point) {
		return Receipt{}, ErrInjectedFailure
	}
	if err := coordinator.activateCommitted(ctx, receipt); err != nil {
		return Receipt{}, err
	}
	return cloneReceipt(receipt), nil
}

func (coordinator *Coordinator) Abort(ctx context.Context, request AbortRequest) (Receipt, error) {
	if err := coordinatorCallError(ctx, coordinator); err != nil || request.Validate() != nil {
		return Receipt{}, coordinatorInputError(ctx)
	}
	record, err := coordinator.repository.Get(ctx, request.OperationID)
	if err != nil {
		return Receipt{}, coordinatorDependencyError(ctx, err)
	}
	providerRecord, err := coordinator.provider.Inspect(ctx, request.OperationID)
	if err != nil {
		return Receipt{}, coordinatorDependencyError(ctx, err)
	}
	if record.Validate() != nil || providerRecord.Validate() != nil || providerRecord.Reservation != record.Reservation ||
		!equalOptionalDigest(providerRecord.BoundEffectDigest, record.BoundEffectDigest) ||
		!equalOptionalDatabasePoint(providerRecord.BoundDatabasePoint, record.BoundDatabasePoint) {
		return Receipt{}, ErrConflict
	}
	if record.TerminalReceipt != nil {
		if providerRecord.TerminalReceipt == nil ||
			providerRecord.TerminalReceipt.ReceiptDigest != record.TerminalReceipt.ReceiptDigest {
			return Receipt{}, ErrConflict
		}
		if record.TerminalReceipt.Status != StatusAborted || providerRecord.TerminalReceipt.Status != StatusAborted {
			return Receipt{}, ErrTerminalConflict
		}
		if record.TerminalReceipt.AbortReason == nil || *record.TerminalReceipt.AbortReason != request.Reason ||
			providerRecord.TerminalReceipt.AbortReason == nil || *providerRecord.TerminalReceipt.AbortReason != request.Reason {
			return Receipt{}, ErrConflict
		}
		return cloneReceipt(*record.TerminalReceipt), nil
	}
	if providerRecord.TerminalReceipt != nil {
		if providerRecord.TerminalReceipt.Status != StatusAborted {
			return Receipt{}, ErrTerminalConflict
		}
		if providerRecord.TerminalReceipt.AbortReason == nil || *providerRecord.TerminalReceipt.AbortReason != request.Reason {
			return Receipt{}, ErrConflict
		}
		resolved, resolveErr := coordinator.resolve(ctx, request.OperationID)
		if resolveErr != nil {
			return Receipt{}, resolveErr
		}
		if resolved.State != EffectAbsent || record.BoundEffectDigest != nil || record.BoundDatabasePoint != nil ||
			providerRecord.BoundEffectDigest != nil || providerRecord.BoundDatabasePoint != nil {
			return Receipt{}, ErrConflict
		}
		if err := coordinator.recordAborted(ctx, *providerRecord.TerminalReceipt); err != nil {
			return Receipt{}, err
		}
		return cloneReceipt(*providerRecord.TerminalReceipt), nil
	}
	resolved, err := coordinator.resolve(ctx, request.OperationID)
	if err != nil {
		return Receipt{}, err
	}
	if resolved.State != EffectAbsent || record.BoundEffectDigest != nil || record.BoundDatabasePoint != nil ||
		providerRecord.BoundEffectDigest != nil || providerRecord.BoundDatabasePoint != nil {
		return Receipt{}, ErrConflict
	}
	receipt, err := coordinator.provider.Abort(ctx, request)
	if err != nil {
		return Receipt{}, coordinatorDependencyError(ctx, err)
	}
	if !validAbortedReceiptFor(receipt, providerRecord, request.Reason) {
		return Receipt{}, ErrInjectedFailure
	}
	if err := coordinator.recordAborted(ctx, receipt); err != nil {
		return Receipt{}, err
	}
	return cloneReceipt(receipt), nil
}

func (coordinator *Coordinator) Recover(ctx context.Context, operationID uuid.UUID) (Receipt, error) {
	if err := coordinatorCallError(ctx, coordinator); err != nil || !validOperationID(operationID) {
		return Receipt{}, coordinatorInputError(ctx)
	}
	providerRecord, err := coordinator.provider.Inspect(ctx, operationID)
	if err != nil {
		return Receipt{}, coordinatorDependencyError(ctx, err)
	}
	if providerRecord.Validate() != nil || providerRecord.OperationID != operationID {
		return Receipt{}, ErrInjectedFailure
	}
	resolved, err := coordinator.resolve(ctx, operationID)
	if err != nil {
		return Receipt{}, err
	}
	databaseRecord, err := coordinator.repository.Get(ctx, operationID)
	if err == ErrNotFound {
		if providerRecord.TerminalReceipt != nil || resolved.State != EffectAbsent ||
			providerRecord.BoundEffectDigest != nil || providerRecord.BoundDatabasePoint != nil {
			return Receipt{}, ErrConflict
		}
		if err := coordinator.recordPending(ctx, providerRecord.Reservation); err != nil {
			return Receipt{}, err
		}
		databaseRecord, err = coordinator.repository.Get(ctx, operationID)
	}
	if err != nil {
		return Receipt{}, coordinatorDependencyError(ctx, err)
	}
	if databaseRecord.Validate() != nil || databaseRecord.Reservation != providerRecord.Reservation {
		return Receipt{}, ErrConflict
	}
	if providerRecord.TerminalReceipt == nil {
		if providerRecord.BoundEffectDigest != nil || providerRecord.BoundDatabasePoint != nil {
			return Receipt{}, ErrConflict
		}
	} else if !equalOptionalDigest(databaseRecord.BoundEffectDigest, providerRecord.BoundEffectDigest) ||
		!equalOptionalDatabasePoint(databaseRecord.BoundDatabasePoint, providerRecord.BoundDatabasePoint) {
		return Receipt{}, ErrConflict
	}
	if databaseRecord.TerminalReceipt != nil {
		if providerRecord.TerminalReceipt == nil ||
			databaseRecord.TerminalReceipt.ReceiptDigest != providerRecord.TerminalReceipt.ReceiptDigest {
			return Receipt{}, ErrConflict
		}
		return cloneReceipt(*providerRecord.TerminalReceipt), nil
	}
	if providerRecord.TerminalReceipt != nil {
		return coordinator.completeProviderTerminal(ctx, databaseRecord, providerRecord, resolved)
	}

	switch resolved.State {
	case EffectAbsent:
		if databaseRecord.BoundEffectDigest != nil {
			return Receipt{}, ErrConflict
		}
		return coordinator.abortKnownRecord(ctx, databaseRecord, providerRecord, AbortValidationFailed)
	case EffectCommitted:
		if !resolvedMatchesReservation(resolved, databaseRecord.Reservation) {
			return Receipt{}, ErrConflict
		}
		return coordinator.Finalize(ctx, CoordinatorFinalizeRequest{OperationID: operationID, EffectDigest: resolved.EffectDigest})
	case EffectPrepared, EffectTerminal:
		return Receipt{}, ErrConflict
	default:
		return Receipt{}, ErrInjectedFailure
	}
}

func (coordinator *Coordinator) CheckReady(ctx context.Context) (Readiness, error) {
	if err := coordinatorCallError(ctx, coordinator); err != nil {
		return Readiness{}, coordinatorInputError(ctx)
	}
	providerHead, err := coordinator.provider.Head(ctx)
	if err != nil {
		if coordinatorDependencyError(ctx, err) == ErrCanceled {
			return Readiness{}, ErrCanceled
		}
		return unavailableReadiness(Head{}, DatabaseHead{}, ReadinessProviderUnavailable), nil
	}
	if providerHead.Validate() != nil {
		return unavailableReadiness(Head{}, DatabaseHead{}, ReadinessProviderUnavailable), nil
	}
	providerHead = cloneHead(providerHead)

	currentPoint, err := coordinator.repository.CaptureDatabasePoint(ctx)
	if err != nil {
		if coordinatorDependencyError(ctx, err) == ErrCanceled {
			return Readiness{}, ErrCanceled
		}
		return unavailableReadiness(providerHead, DatabaseHead{}, ReadinessDatabaseUnavailable), nil
	}
	if currentPoint.Validate() != nil {
		return unavailableReadiness(providerHead, DatabaseHead{}, ReadinessDatabaseUnavailable), nil
	}
	databaseHead, err := coordinator.repository.Head(ctx)
	if err != nil {
		if coordinatorDependencyError(ctx, err) == ErrCanceled {
			return Readiness{}, ErrCanceled
		}
		return unavailableReadiness(providerHead, DatabaseHead{}, ReadinessDatabaseUnavailable), nil
	}
	if !validDatabaseHead(databaseHead) {
		return unavailableReadiness(providerHead, DatabaseHead{}, ReadinessDatabaseUnavailable), nil
	}
	databaseHead = cloneDatabaseHead(databaseHead)

	if providerHead.LatestCommittedDatabasePoint != nil &&
		currentPoint.SystemID != providerHead.LatestCommittedDatabasePoint.SystemID {
		return unavailableReadiness(providerHead, databaseHead, ReadinessDatabaseIdentityMismatch), nil
	}

	providerEmpty := providerHead.LatestReservedSequence == 0
	databaseEmpty := databaseHead.Epoch == 0
	if providerEmpty {
		if !databaseEmpty {
			return unavailableReadiness(providerHead, databaseHead, ReadinessDatabaseAheadProvider), nil
		}
		databaseHead.Epoch = providerHead.Epoch
	} else {
		if databaseEmpty || databaseHead.Epoch < providerHead.Epoch {
			return unavailableReadiness(providerHead, databaseHead, ReadinessDatabaseBehindProvider), nil
		}
		if databaseHead.Epoch > providerHead.Epoch {
			return unavailableReadiness(providerHead, databaseHead, ReadinessDatabaseAheadProvider), nil
		}
		if databaseHead.RecordCount != databaseHead.LatestReservedSequence || databaseHead.HasSequenceGap {
			return unavailableReadiness(providerHead, databaseHead, ReadinessDatabaseSequenceGap), nil
		}
		if databaseHead.LatestReservedSequence < providerHead.LatestReservedSequence {
			return unavailableReadiness(providerHead, databaseHead, ReadinessDatabaseBehindProvider), nil
		}
		if databaseHead.LatestReservedSequence > providerHead.LatestReservedSequence {
			return unavailableReadiness(providerHead, databaseHead, ReadinessDatabaseAheadProvider), nil
		}
		if databaseHead.LatestReservationDigest != providerHead.LatestReservationDigest {
			return unavailableReadiness(providerHead, databaseHead, ReadinessReservationMismatch), nil
		}
		if databaseHead.LatestCommittedSequence < providerHead.LatestCommittedSequence {
			return unavailableReadiness(providerHead, databaseHead, ReadinessDatabaseBehindProvider), nil
		}
		if databaseHead.LatestCommittedSequence > providerHead.LatestCommittedSequence {
			return unavailableReadiness(providerHead, databaseHead, ReadinessDatabaseAheadProvider), nil
		}
		if databaseHead.LatestCommittedOperationID != providerHead.LatestCommittedOperationID ||
			databaseHead.LatestCommittedReceiptDigest != providerHead.LatestCommittedReceiptDigest ||
			!equalOptionalDatabasePoint(databaseHead.LatestCommittedDatabasePoint, providerHead.LatestCommittedDatabasePoint) {
			return unavailableReadiness(providerHead, databaseHead, ReadinessCommittedMismatch), nil
		}
	}

	if providerHead.LatestCommittedDatabasePoint != nil {
		if currentPoint.Timeline < providerHead.LatestCommittedDatabasePoint.Timeline {
			return unavailableReadiness(providerHead, databaseHead, ReadinessDatabaseTimelineBehind), nil
		}
		if walPositionLess(currentPoint.RequiredLSN, providerHead.LatestCommittedDatabasePoint.RequiredLSN) {
			return unavailableReadiness(providerHead, databaseHead, ReadinessDatabaseWALBehind), nil
		}
	}

	pending, err := coordinator.repository.ListPending(ctx, providerHead.Epoch)
	if err != nil {
		if coordinatorDependencyError(ctx, err) == ErrCanceled {
			return Readiness{}, ErrCanceled
		}
		return unavailableReadiness(providerHead, databaseHead, ReadinessDatabaseUnavailable), nil
	}
	if uint64(len(pending)) != databaseHead.PendingCount {
		return unavailableReadiness(providerHead, databaseHead, ReadinessPendingCountMismatch), nil
	}
	if len(pending) != 0 {
		reason, inspectErr := coordinator.inspectPending(ctx, providerHead.Epoch, pending)
		if inspectErr != nil {
			return Readiness{}, inspectErr
		}
		return unavailableReadiness(providerHead, databaseHead, reason), nil
	}
	return Readiness{ProviderHead: providerHead, DatabaseHead: databaseHead, Ready: true, Reason: ReadinessReady}, nil
}

func (coordinator *Coordinator) inspectPending(ctx context.Context, epoch uint64, pending []PendingFence) (ReadinessReason, error) {
	reason := ReadinessPendingUnresolved
	for _, fence := range pending {
		rowMismatch := !validPendingFence(fence, epoch)
		providerRecord, providerErr := coordinator.provider.Inspect(ctx, fence.OperationID)
		resolved, resolveErr := coordinator.resolver.ResolveAuthorityEffect(ctx, fence.OperationID)
		if ctx == nil || ctx.Err() != nil || errors.Is(providerErr, context.Canceled) || errors.Is(providerErr, context.DeadlineExceeded) ||
			errors.Is(resolveErr, context.Canceled) || errors.Is(resolveErr, context.DeadlineExceeded) ||
			errors.Is(providerErr, ErrCanceled) || errors.Is(resolveErr, ErrCanceled) {
			return "", ErrCanceled
		}
		if providerErr != nil {
			providerReason := ReadinessProviderUnavailable
			if errors.Is(providerErr, ErrNotFound) || errors.Is(providerErr, ErrConflict) || errors.Is(providerErr, ErrTerminalConflict) {
				providerReason = ReadinessPendingMismatch
			}
			reason = readinessReasonWithPriority(reason, providerReason)
		}
		if resolveErr != nil {
			reason = readinessReasonWithPriority(reason, ReadinessEffectUnavailable)
		}
		if rowMismatch || providerErr != nil || resolveErr != nil {
			continue
		}
		if resolved.Validate() != nil || !pendingStateMatches(fence, providerRecord, resolved) {
			reason = readinessReasonWithPriority(reason, ReadinessPendingMismatch)
		}
	}
	return reason, nil
}

func pendingStateMatches(fence PendingFence, providerRecord Record, resolved ResolvedEffect) bool {
	reservation := Reservation{
		OperationID:       fence.OperationID,
		Kind:              fence.Kind,
		ScopeKind:         fence.ScopeKind,
		ScopeDigest:       fence.ScopeDigest,
		Epoch:             fence.Epoch,
		Sequence:          fence.Sequence,
		ReservationDigest: fence.ReservationDigest,
	}
	if providerRecord.Validate() != nil || providerRecord.Reservation != reservation {
		return false
	}
	if providerRecord.TerminalReceipt == nil {
		if providerRecord.BoundEffectDigest != nil || providerRecord.BoundDatabasePoint != nil {
			return false
		}
	} else if !equalOptionalDigest(providerRecord.BoundEffectDigest, fence.BoundEffectDigest) ||
		!equalOptionalDatabasePoint(providerRecord.BoundDatabasePoint, fence.BoundDatabasePoint) {
		return false
	}

	switch resolved.State {
	case EffectAbsent:
		return fence.BoundEffectDigest == nil && providerRecord.TerminalReceipt == nil
	case EffectPrepared:
		return resolvedMatchesReservation(resolved, reservation) && fence.BoundEffectDigest == nil &&
			providerRecord.TerminalReceipt == nil
	case EffectCommitted:
		if !resolvedMatchesReservation(resolved, reservation) {
			return false
		}
		if fence.BoundEffectDigest != nil && *fence.BoundEffectDigest != resolved.EffectDigest {
			return false
		}
		if providerRecord.TerminalReceipt == nil {
			return true
		}
		return providerRecord.TerminalReceipt.Status == StatusCommitted &&
			providerRecord.TerminalReceipt.EffectDigest != nil &&
			*providerRecord.TerminalReceipt.EffectDigest == resolved.EffectDigest
	case EffectTerminal:
		return false
	default:
		return false
	}
}

func unavailableReadiness(provider Head, database DatabaseHead, reason ReadinessReason) Readiness {
	return Readiness{ProviderHead: cloneHead(provider), DatabaseHead: cloneDatabaseHead(database), Reason: reason}
}

func cloneDatabaseHead(value DatabaseHead) DatabaseHead {
	clone := value
	clone.LatestCommittedDatabasePoint = cloneDatabasePointPointer(value.LatestCommittedDatabasePoint)
	return clone
}

func validDatabaseHead(value DatabaseHead) bool {
	if value.Epoch == 0 {
		return value.RecordCount == 0 && value.LatestReservedSequence == 0 && value.LatestCommittedSequence == 0 &&
			value.PendingCount == 0 && !value.HasSequenceGap && value.LatestReservationDigest == (contracts.Digest{}) &&
			value.LatestCommittedOperationID == uuid.Nil && value.LatestCommittedReceiptDigest == (contracts.Digest{}) &&
			value.LatestCommittedDatabasePoint == nil
	}
	if !validPositiveCoordinate(value.Epoch) || value.RecordCount == 0 || value.RecordCount > math.MaxInt64 ||
		value.LatestReservedSequence == 0 || value.LatestReservedSequence > math.MaxInt64 ||
		value.LatestCommittedSequence > value.LatestReservedSequence || value.PendingCount > value.RecordCount ||
		value.HasSequenceGap != (value.RecordCount != value.LatestReservedSequence) ||
		value.LatestReservationDigest == (contracts.Digest{}) {
		return false
	}
	if value.LatestCommittedSequence == 0 {
		return value.LatestCommittedOperationID == uuid.Nil && value.LatestCommittedReceiptDigest == (contracts.Digest{}) &&
			value.LatestCommittedDatabasePoint == nil
	}
	return validOperationID(value.LatestCommittedOperationID) &&
		value.LatestCommittedReceiptDigest != (contracts.Digest{}) && value.LatestCommittedDatabasePoint != nil &&
		value.LatestCommittedDatabasePoint.Validate() == nil
}

func validPendingFence(value PendingFence, epoch uint64) bool {
	reservation := Reservation{
		OperationID:       value.OperationID,
		Kind:              value.Kind,
		ScopeKind:         value.ScopeKind,
		ScopeDigest:       value.ScopeDigest,
		Epoch:             value.Epoch,
		Sequence:          value.Sequence,
		ReservationDigest: value.ReservationDigest,
	}
	if value.Epoch != epoch || reservation.Validate() != nil || value.ReservedAt.IsZero() || value.ReservedAt.Location() != time.UTC ||
		(value.BoundEffectDigest == nil) != (value.BoundDatabasePoint == nil) ||
		(value.BoundEffectDigest == nil) != (value.EffectBoundAt == nil) {
		return false
	}
	if value.BoundEffectDigest == nil {
		return true
	}
	return *value.BoundEffectDigest != (contracts.Digest{}) && value.BoundDatabasePoint.Validate() == nil &&
		!value.EffectBoundAt.IsZero() && value.EffectBoundAt.Location() == time.UTC && !value.EffectBoundAt.Before(value.ReservedAt)
}

func readinessReasonWithPriority(current, candidate ReadinessReason) ReadinessReason {
	priority := func(reason ReadinessReason) int {
		switch reason {
		case ReadinessProviderUnavailable:
			return 4
		case ReadinessEffectUnavailable:
			return 3
		case ReadinessPendingMismatch:
			return 2
		case ReadinessPendingUnresolved:
			return 1
		default:
			return 0
		}
	}
	if priority(candidate) > priority(current) {
		return candidate
	}
	return current
}

func walPositionLess(left, right WALPosition) bool {
	leftHigh, leftLow, leftOK := walPositionParts(left)
	rightHigh, rightLow, rightOK := walPositionParts(right)
	if !leftOK || !rightOK {
		return true
	}
	return leftHigh < rightHigh || leftHigh == rightHigh && leftLow < rightLow
}

func walPositionParts(value WALPosition) (uint64, uint64, bool) {
	canonical, err := canonicalWALPosition(value)
	if err != nil || canonical != value {
		return 0, 0, false
	}
	parts := strings.SplitN(string(value), "/", 2)
	high, highErr := strconv.ParseUint(parts[0], 16, 32)
	low, lowErr := strconv.ParseUint(parts[1], 16, 32)
	return high, low, highErr == nil && lowErr == nil
}

func (coordinator *Coordinator) ensureBound(ctx context.Context, record Record, effectDigest contracts.Digest) (DatabasePoint, error) {
	if record.BoundEffectDigest != nil {
		if *record.BoundEffectDigest != effectDigest || record.BoundDatabasePoint == nil || record.BoundDatabasePoint.Validate() != nil {
			return DatabasePoint{}, ErrConflict
		}
		return *cloneDatabasePointPointer(record.BoundDatabasePoint), nil
	}
	if record.BoundDatabasePoint != nil {
		return DatabasePoint{}, ErrConflict
	}
	point, err := coordinator.repository.CaptureDatabasePoint(ctx)
	if err != nil {
		return DatabasePoint{}, coordinatorDependencyError(ctx, err)
	}
	if point.Validate() != nil {
		return DatabasePoint{}, ErrInjectedFailure
	}
	now, err := coordinator.now(ctx)
	if err != nil {
		return DatabasePoint{}, err
	}
	err = coordinator.transaction.withAuthorityTransaction(ctx, func(dbtx store.DBTX) error {
		return coordinator.repository.BindEffect(ctx, dbtx, record.OperationID, effectDigest, point, now)
	})
	if err != nil {
		return DatabasePoint{}, coordinatorDependencyError(ctx, err)
	}
	bound, err := coordinator.repository.Get(ctx, record.OperationID)
	if err != nil {
		return DatabasePoint{}, coordinatorDependencyError(ctx, err)
	}
	if bound.BoundEffectDigest == nil || *bound.BoundEffectDigest != effectDigest ||
		bound.BoundDatabasePoint == nil || *bound.BoundDatabasePoint != point {
		return DatabasePoint{}, ErrConflict
	}
	return point, nil
}

func (coordinator *Coordinator) completeProviderTerminal(
	ctx context.Context,
	databaseRecord Record,
	providerRecord Record,
	resolved ResolvedEffect,
) (Receipt, error) {
	receipt := *providerRecord.TerminalReceipt
	switch receipt.Status {
	case StatusCommitted:
		if resolved.State != EffectCommitted || !resolvedMatchesReservation(resolved, providerRecord.Reservation) ||
			receipt.EffectDigest == nil || resolved.EffectDigest != *receipt.EffectDigest ||
			databaseRecord.BoundEffectDigest == nil || !equalOptionalDigest(databaseRecord.BoundEffectDigest, receipt.EffectDigest) ||
			!equalOptionalDatabasePoint(databaseRecord.BoundDatabasePoint, receipt.DatabasePoint) {
			return Receipt{}, ErrConflict
		}
		if err := coordinator.activateCommitted(ctx, receipt); err != nil {
			return Receipt{}, err
		}
	case StatusAborted:
		if resolved.State != EffectAbsent || databaseRecord.BoundEffectDigest != nil || databaseRecord.BoundDatabasePoint != nil ||
			receipt.EffectDigest != nil || receipt.DatabasePoint != nil ||
			!equalOptionalDigest(databaseRecord.BoundEffectDigest, receipt.EffectDigest) ||
			!equalOptionalDatabasePoint(databaseRecord.BoundDatabasePoint, receipt.DatabasePoint) {
			return Receipt{}, ErrConflict
		}
		if err := coordinator.recordAborted(ctx, receipt); err != nil {
			return Receipt{}, err
		}
	default:
		return Receipt{}, ErrInjectedFailure
	}
	return cloneReceipt(receipt), nil
}

func (coordinator *Coordinator) abortKnownRecord(
	ctx context.Context,
	databaseRecord Record,
	providerRecord Record,
	reason AbortReason,
) (Receipt, error) {
	if databaseRecord.Reservation != providerRecord.Reservation {
		return Receipt{}, ErrConflict
	}
	receipt, err := coordinator.provider.Abort(ctx, AbortRequest{OperationID: providerRecord.OperationID, Reason: reason})
	if err != nil {
		return Receipt{}, coordinatorDependencyError(ctx, err)
	}
	if !validAbortedReceiptFor(receipt, providerRecord, reason) {
		return Receipt{}, ErrInjectedFailure
	}
	if err := coordinator.recordAborted(ctx, receipt); err != nil {
		return Receipt{}, err
	}
	return cloneReceipt(receipt), nil
}

func (coordinator *Coordinator) recordPending(ctx context.Context, reservation Reservation) error {
	now, err := coordinator.now(ctx)
	if err != nil {
		return err
	}
	err = coordinator.transaction.withAuthorityTransaction(ctx, func(dbtx store.DBTX) error {
		return coordinator.repository.RecordPending(ctx, dbtx, reservation, now)
	})
	return coordinatorDependencyError(ctx, err)
}

func (coordinator *Coordinator) activateCommitted(ctx context.Context, receipt Receipt) error {
	now, err := coordinator.now(ctx)
	if err != nil {
		return err
	}
	err = coordinator.transaction.withAuthorityTransaction(ctx, func(dbtx store.DBTX) error {
		return coordinator.repository.ActivateCommitted(ctx, dbtx, receipt, now)
	})
	return coordinatorDependencyError(ctx, err)
}

func (coordinator *Coordinator) recordAborted(ctx context.Context, receipt Receipt) error {
	now, err := coordinator.now(ctx)
	if err != nil {
		return err
	}
	err = coordinator.transaction.withAuthorityTransaction(ctx, func(dbtx store.DBTX) error {
		return coordinator.repository.RecordAborted(ctx, dbtx, receipt, now)
	})
	return coordinatorDependencyError(ctx, err)
}

func (coordinator *Coordinator) resolve(ctx context.Context, operationID uuid.UUID) (ResolvedEffect, error) {
	resolved, err := coordinator.resolver.ResolveAuthorityEffect(ctx, operationID)
	if err != nil {
		return ResolvedEffect{}, coordinatorDependencyError(ctx, err)
	}
	if resolved.Validate() != nil {
		return ResolvedEffect{}, ErrInjectedFailure
	}
	return resolved, nil
}

func (resolved ResolvedEffect) Validate() error {
	switch resolved.State {
	case EffectAbsent:
		if resolved.Kind != "" || resolved.ScopeKind != "" || resolved.ScopeDigest != (contracts.Digest{}) ||
			resolved.EffectDigest != (contracts.Digest{}) {
			return ErrInvalidArgument
		}
		return nil
	case EffectPrepared, EffectCommitted, EffectTerminal:
		if resolved.Kind.Validate() != nil || resolved.ScopeKind.Validate() != nil ||
			!validEffectScope(resolved.Kind, resolved.ScopeKind) || !validScopeDigest(resolved.ScopeKind, resolved.ScopeDigest) ||
			resolved.EffectDigest == (contracts.Digest{}) {
			return ErrInvalidArgument
		}
		return nil
	default:
		return ErrInvalidArgument
	}
}

func resolvedMatchesReservation(resolved ResolvedEffect, reservation Reservation) bool {
	return resolved.Kind == reservation.Kind && resolved.ScopeKind == reservation.ScopeKind &&
		resolved.ScopeDigest == reservation.ScopeDigest
}

func validCommittedReceiptFor(receipt Receipt, reservation Reservation, effectDigest contracts.Digest, point DatabasePoint) bool {
	return receipt.Validate() == nil && receipt.Status == StatusCommitted && receipt.Reservation == reservation &&
		receipt.EffectDigest != nil && *receipt.EffectDigest == effectDigest && receipt.DatabasePoint != nil &&
		*receipt.DatabasePoint == point
}

func validAbortedReceiptFor(receipt Receipt, record Record, reason AbortReason) bool {
	return receipt.Validate() == nil && receipt.Status == StatusAborted && receipt.Reservation == record.Reservation &&
		receipt.AbortReason != nil && *receipt.AbortReason == reason &&
		equalOptionalDigest(receipt.EffectDigest, record.BoundEffectDigest) &&
		equalOptionalDatabasePoint(receipt.DatabasePoint, record.BoundDatabasePoint)
}

func coordinatorCallError(ctx context.Context, coordinator *Coordinator) error {
	if coordinator == nil || nilAuthorityRepositoryValue(coordinator.provider) ||
		nilAuthorityRepositoryValue(coordinator.repository) || nilAuthorityRepositoryValue(coordinator.resolver) ||
		nilAuthorityRepositoryValue(coordinator.clock) || nilAuthorityRepositoryValue(coordinator.transaction) {
		return ErrInvalidArgument
	}
	return contextError(ctx)
}

func coordinatorInputError(ctx context.Context) error {
	if ctx != nil && ctx.Err() != nil {
		return ErrCanceled
	}
	return ErrInvalidArgument
}

func coordinatorDependencyError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) ||
		(ctx != nil && ctx.Err() != nil) || errors.Is(err, ErrCanceled) {
		return ErrCanceled
	}
	for _, sentinel := range []error{
		ErrInvalidArgument, ErrConflict, ErrTerminalConflict, ErrNotFound,
		ErrInjectedFailure, ErrResponseLost,
	} {
		if errors.Is(err, sentinel) {
			return sentinel
		}
	}
	return ErrInjectedFailure
}

func (coordinator *Coordinator) now(ctx context.Context) (time.Time, error) {
	if ctx == nil || ctx.Err() != nil {
		return time.Time{}, ErrCanceled
	}
	now := coordinator.clock.Now()
	if now.IsZero() || now.Location() != time.UTC || now.Year() < 2020 || now.Year() > 2100 {
		return time.Time{}, ErrInvalidArgument
	}
	return now, nil
}

func (repository *PostgresRepository) withAuthorityTransaction(ctx context.Context, operation func(store.DBTX) error) error {
	if repositoryCallError(ctx, repository) != nil || operation == nil {
		return coordinatorInputError(ctx)
	}
	beginner, ok := repository.database.(authorityTransactionBeginner)
	if !ok || nilAuthorityRepositoryValue(beginner) {
		return ErrInvalidArgument
	}
	transaction, err := beginner.Begin(ctx)
	if err != nil {
		return coordinatorDependencyError(ctx, err)
	}
	if err := operation(transaction); err != nil {
		_ = transaction.Rollback(context.WithoutCancel(ctx))
		return coordinatorDependencyError(ctx, err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return coordinatorDependencyError(ctx, err)
	}
	return nil
}
