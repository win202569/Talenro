package authority

import (
	"bytes"
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
	repository  AuthorityV7Repository
	dispatcher  *effectDispatcher
	clock       securitykit.Clock
	transaction authorityTransactionRunner
}

type authorityTransactionRunner interface {
	beginAuthorityTransaction(context.Context) (authorityTransaction, error)
}

type authorityTransaction interface {
	store.DBTX
	Commit(context.Context) error
	Rollback(context.Context) error
}

type authorityTransactionBeginner interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}

func NewCoordinator(provider Provider, repository Repository, dispatcher EffectDispatcher) (*Coordinator, error) {
	// A promoted private marker does not seal an embedded wrapper. Check the
	// exact dynamic type before calling any supplied dependency.
	sealed, exact := dispatcher.(*effectDispatcher)
	if !exact || sealed == nil {
		return nil, ErrInvalidArgument
	}
	transaction, ok := repository.(authorityTransactionRunner)
	claimRepository, claimAware := repository.(AuthorityV7Repository)
	if nilAuthorityRepositoryValue(provider) || nilAuthorityRepositoryValue(repository) ||
		!claimAware || !ok || nilAuthorityRepositoryValue(transaction) {
		return nil, ErrInvalidArgument
	}
	return &Coordinator{
		provider:    provider,
		repository:  claimRepository,
		dispatcher:  sealed,
		clock:       coordinatorSystemClock{},
		transaction: transaction,
	}, nil
}

type coordinatorSystemClock struct{}

func (coordinatorSystemClock) Now() time.Time { return time.Now().UTC() }

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
	if err := coordinatorCallError(ctx, coordinator); err != nil || !validOperationID(request.OperationID) || request.EffectDigest == (contracts.Digest{}) {
		return Receipt{}, coordinatorInputError(ctx)
	}
	observed, err := coordinator.provider.Inspect(ctx, request.OperationID)
	if err != nil {
		return Receipt{}, coordinatorDependencyError(ctx, err)
	}
	if observed.Validate() != nil || observed.OperationID != request.OperationID {
		return Receipt{}, ErrInjectedFailure
	}
	if observed.TerminalReceipt != nil {
		if observed.TerminalReceipt.Status != StatusCommitted {
			return Receipt{}, ErrTerminalConflict
		}
		if observed.TerminalReceipt.EffectDigest == nil || *observed.TerminalReceipt.EffectDigest != request.EffectDigest {
			return Receipt{}, ErrConflict
		}
		return coordinator.completeCommitted(ctx, *observed.TerminalReceipt)
	}

	// Capture the post-effect WAL point outside the short binding transaction.
	stored, err := coordinator.repository.GetStoredFence(ctx, request.OperationID)
	if err != nil {
		return Receipt{}, coordinatorDependencyError(ctx, err)
	}
	if stored.Record.Reservation != observed.Reservation || stored.AbortClaim != nil || stored.Record.TerminalReceipt != nil {
		return Receipt{}, ErrConflict
	}
	var point DatabasePoint
	if stored.Record.BoundDatabasePoint != nil {
		if stored.Record.BoundEffectDigest == nil || *stored.Record.BoundEffectDigest != request.EffectDigest {
			return Receipt{}, ErrConflict
		}
		point = *stored.Record.BoundDatabasePoint
	} else {
		err = coordinator.transact(ctx, nil, func(tx store.DBTX) error {
			fence, effect, err := coordinator.lockResolved(ctx, tx, observed.Reservation)
			if err != nil {
				return err
			}
			if fence.AbortClaim != nil || fence.Record.TerminalReceipt != nil || effect.State != EffectCommitted ||
				effect.EffectDigest != request.EffectDigest {
				return ErrConflict
			}
			return nil
		})
		if err != nil {
			return Receipt{}, err
		}
		point, err = coordinator.repository.CaptureDatabasePoint(ctx)
		if err != nil {
			return Receipt{}, coordinatorDependencyError(ctx, err)
		}
		if point.Validate() != nil {
			return Receipt{}, ErrInjectedFailure
		}
	}
	err = coordinator.transact(ctx, nil, func(tx store.DBTX) error {
		fence, resolved, err := coordinator.lockResolved(ctx, tx, observed.Reservation)
		if err != nil {
			return err
		}
		if fence.AbortClaim != nil || fence.Record.TerminalReceipt != nil || resolved.State != EffectCommitted ||
			resolved.EffectDigest != request.EffectDigest {
			return ErrConflict
		}
		if fence.Record.BoundEffectDigest != nil {
			if *fence.Record.BoundEffectDigest != request.EffectDigest || fence.Record.BoundDatabasePoint == nil ||
				*fence.Record.BoundDatabasePoint != point {
				return ErrConflict
			}
			return nil
		}
		now, err := coordinator.now(ctx)
		if err != nil {
			return err
		}
		return coordinator.repository.BindEffect(ctx, tx, request.OperationID, request.EffectDigest, point, now)
	})
	if err != nil {
		return Receipt{}, err
	}
	receipt, err := coordinator.provider.Finalize(ctx, FinalizeRequest{OperationID: request.OperationID,
		EffectDigest: request.EffectDigest, DBSystemID: point.SystemID, DBTimeline: point.Timeline, RequiredLSN: point.RequiredLSN})
	if err != nil {
		return Receipt{}, coordinatorDependencyError(ctx, err)
	}
	if !validCommittedReceiptFor(receipt, observed.Reservation, request.EffectDigest, point) {
		return Receipt{}, ErrInjectedFailure
	}
	return coordinator.completeCommitted(ctx, receipt)
}

func (coordinator *Coordinator) Abort(ctx context.Context, request AbortRequest) (Receipt, error) {
	if err := coordinatorCallError(ctx, coordinator); err != nil || request.Validate() != nil {
		return Receipt{}, coordinatorInputError(ctx)
	}
	observed, err := coordinator.provider.Inspect(ctx, request.OperationID)
	if err != nil {
		return Receipt{}, coordinatorDependencyError(ctx, err)
	}
	if observed.Validate() != nil || observed.OperationID != request.OperationID {
		return Receipt{}, ErrInjectedFailure
	}
	if observed.BoundEffectDigest != nil || observed.BoundDatabasePoint != nil {
		return Receipt{}, ErrConflict
	}
	if observed.TerminalReceipt != nil && (observed.TerminalReceipt.Status != StatusAborted ||
		observed.TerminalReceipt.AbortReason == nil || *observed.TerminalReceipt.AbortReason != request.Reason) {
		return Receipt{}, ErrTerminalConflict
	}
	// The committed durable claim is the semantic lock. No provider call is
	// reachable until this READ COMMITTED transaction has released its row locks.
	var terminal *Receipt
	err = coordinator.transact(ctx, nil, func(tx store.DBTX) error {
		fence, resolved, err := coordinator.lockResolved(ctx, tx, observed.Reservation)
		if err != nil {
			return err
		}
		if resolved.State != EffectAbsent || fence.Record.BoundEffectDigest != nil || fence.Record.BoundDatabasePoint != nil {
			return ErrConflict
		}
		if fence.Record.TerminalReceipt != nil {
			if observed.TerminalReceipt == nil || !receiptEquals(*fence.Record.TerminalReceipt, *observed.TerminalReceipt) ||
				fence.AbortClaim == nil || fence.AbortClaim.Reason != request.Reason {
				return ErrConflict
			}
			terminal = cloneReceiptPointer(fence.Record.TerminalReceipt)
			return nil
		}
		if fence.AbortClaim != nil {
			if fence.AbortClaim.Reason != request.Reason {
				return ErrConflict
			}
			return nil
		}
		now, err := coordinator.now(ctx)
		if err != nil {
			return err
		}
		return coordinator.repository.ClaimAbort(ctx, tx, request.OperationID, request.Reason, now)
	})
	if err != nil {
		return Receipt{}, err
	}
	if terminal != nil {
		return cloneReceipt(*terminal), nil
	}
	var receipt Receipt
	if observed.TerminalReceipt != nil {
		receipt = cloneReceipt(*observed.TerminalReceipt)
	} else {
		receipt, err = coordinator.provider.Abort(ctx, request)
		if err != nil {
			return Receipt{}, coordinatorDependencyError(ctx, err)
		}
	}
	if !validAbortedReceiptFor(receipt, observed, request.Reason) {
		return Receipt{}, ErrInjectedFailure
	}
	err = coordinator.transact(ctx, nil, func(tx store.DBTX) error {
		fence, resolved, err := coordinator.lockResolved(ctx, tx, receipt.Reservation)
		if err != nil {
			return err
		}
		if resolved.State != EffectAbsent || fence.AbortClaim == nil || fence.AbortClaim.Reason != request.Reason ||
			fence.Record.BoundEffectDigest != nil || fence.Record.BoundDatabasePoint != nil {
			return ErrConflict
		}
		if fence.Record.TerminalReceipt != nil {
			if !receiptEquals(*fence.Record.TerminalReceipt, receipt) {
				return ErrConflict
			}
			return nil
		}
		now, err := coordinator.now(ctx)
		if err != nil {
			return err
		}
		return coordinator.repository.RecordAborted(ctx, tx, receipt, now)
	})
	if err != nil {
		return Receipt{}, err
	}
	return cloneReceipt(receipt), nil
}

func (coordinator *Coordinator) Recover(ctx context.Context, operationID uuid.UUID) (Receipt, error) {
	if err := coordinatorCallError(ctx, coordinator); err != nil || !validOperationID(operationID) {
		return Receipt{}, coordinatorInputError(ctx)
	}
	observed, err := coordinator.provider.Inspect(ctx, operationID)
	if err != nil {
		return Receipt{}, coordinatorDependencyError(ctx, err)
	}
	if observed.Validate() != nil || observed.OperationID != operationID {
		return Receipt{}, ErrInjectedFailure
	}
	if observed.TerminalReceipt != nil {
		if observed.TerminalReceipt.Status == StatusCommitted {
			return coordinator.completeCommitted(ctx, *observed.TerminalReceipt)
		}
		if observed.BoundEffectDigest != nil || observed.BoundDatabasePoint != nil || observed.TerminalReceipt.AbortReason == nil {
			return Receipt{}, ErrConflict
		}
		return coordinator.Abort(ctx, AbortRequest{OperationID: operationID, Reason: *observed.TerminalReceipt.AbortReason})
	}
	var stored StoredFence
	var effect ResolvedEffect
	err = coordinator.transact(ctx, nil, func(tx store.DBTX) error {
		var err error
		stored, effect, err = coordinator.lockResolved(ctx, tx, observed.Reservation)
		return err
	})
	if err != nil {
		return Receipt{}, err
	}
	if stored.Record.TerminalReceipt != nil || observed.BoundEffectDigest != nil || observed.BoundDatabasePoint != nil {
		return Receipt{}, ErrConflict
	}
	if stored.AbortClaim != nil {
		if effect.State != EffectAbsent {
			return Receipt{}, ErrConflict
		}
		return coordinator.Abort(ctx, AbortRequest{OperationID: operationID, Reason: stored.AbortClaim.Reason})
	}
	if effect.State == EffectCommitted {
		return coordinator.Finalize(ctx, CoordinatorFinalizeRequest{OperationID: operationID, EffectDigest: effect.EffectDigest})
	}
	// Provider reservations contain no Abort reason or claim-v1 activation ID.
	// Absent claims/fences cannot be invented from a current mutable DB pointer.
	return Receipt{}, ErrConflict
}

func (coordinator *Coordinator) lockResolved(ctx context.Context, tx store.DBTX, expected Reservation) (StoredFence, ResolvedEffect, error) {
	fence, err := coordinator.repository.Lock(ctx, tx, expected.OperationID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return StoredFence{}, ResolvedEffect{}, ErrConflict
		}
		return StoredFence{}, ResolvedEffect{}, coordinatorDependencyError(ctx, err)
	}
	if !validStoredFenceShape(fence) || fence.Record.Reservation != expected {
		return StoredFence{}, ResolvedEffect{}, ErrConflict
	}
	effect, err := coordinator.dispatcher.ResolveAuthorityEffectForUpdate(ctx, tx, fence.Record.Reservation)
	if err != nil {
		return StoredFence{}, ResolvedEffect{}, err
	}
	return fence, effect, nil
}

func validStoredFenceShape(fence StoredFence) bool {
	if fence.Record.Validate() != nil {
		return false
	}
	if claim := fence.AbortClaim; claim != nil {
		if claim.Reason.Validate() != nil || claim.ClaimedAt.IsZero() || claim.ClaimedAt.Location() != time.UTC ||
			fence.Record.BoundEffectDigest != nil || fence.Record.BoundDatabasePoint != nil {
			return false
		}
		if receipt := fence.Record.TerminalReceipt; receipt != nil &&
			(receipt.Status != StatusAborted || receipt.AbortReason == nil || *receipt.AbortReason != claim.Reason) {
			return false
		}
	}
	if receipt := fence.Record.TerminalReceipt; receipt != nil && receipt.Status == StatusAborted && fence.AbortClaim == nil {
		return false
	}
	return fence.Record.TerminalReceipt != nil || fence.PersistedOutcome == nil
}

func (coordinator *Coordinator) completeCommitted(ctx context.Context, receipt Receipt) (Receipt, error) {
	if receipt.Validate() != nil || receipt.Status != StatusCommitted {
		return Receipt{}, ErrInvalidArgument
	}
	// A locked read resolves any earlier COMMIT uncertainty before recapture.
	// Never replace persisted Head/evidence with current external observations.
	terminal := false
	err := coordinator.transact(ctx, nil, func(tx store.DBTX) error {
		fence, effect, err := coordinator.lockResolved(ctx, tx, receipt.Reservation)
		if err != nil {
			return err
		}
		if fence.AbortClaim != nil || fence.Record.BoundEffectDigest == nil || fence.Record.BoundDatabasePoint == nil ||
			!validCommittedReceiptFor(receipt, fence.Record.Reservation, *fence.Record.BoundEffectDigest, *fence.Record.BoundDatabasePoint) {
			return ErrConflict
		}
		if fence.Record.TerminalReceipt == nil {
			if effect.State != EffectCommitted || effect.EffectDigest != *receipt.EffectDigest || fence.PersistedOutcome != nil {
				return ErrConflict
			}
			return nil
		}
		if !receiptEquals(*fence.Record.TerminalReceipt, receipt) || effect.State != EffectTerminal {
			return ErrConflict
		}
		if err := coordinator.validatePersistedOutcome(ctx, tx, fence, receipt); err != nil {
			return err
		}
		terminal = true
		return nil
	})
	if err != nil {
		return Receipt{}, err
	}
	if terminal {
		return cloneReceipt(receipt), nil
	}

	proof, err := coordinator.captureActivation(ctx, receipt)
	if err != nil {
		return Receipt{}, err
	}
	err = coordinator.transact(ctx, &proof, func(tx store.DBTX) error {
		fence, effect, err := coordinator.lockResolved(ctx, tx, receipt.Reservation)
		if err != nil {
			return err
		}
		if fence.AbortClaim != nil || fence.Record.TerminalReceipt != nil || effect.State != EffectCommitted ||
			effect.EffectDigest != *receipt.EffectDigest || fence.Record.BoundEffectDigest == nil || fence.Record.BoundDatabasePoint == nil ||
			!validCommittedReceiptFor(receipt, fence.Record.Reservation, *fence.Record.BoundEffectDigest, *fence.Record.BoundDatabasePoint) {
			return ErrConflict
		}
		now, err := coordinator.now(ctx)
		if err != nil {
			return err
		}
		if err := coordinator.repository.ActivateCommitted(ctx, tx, receipt, now); err != nil {
			return err
		}
		if err := coordinator.dispatcher.ActivateAuthorityEffect(ctx, tx, receipt, proof); err != nil {
			return err
		}
		// Load the same domain projection after all handler writes. This catches
		// missing persistence before admission, without using a fresh proof as a
		// recovery credential.
		persisted, err := coordinator.repository.Lock(ctx, tx, receipt.OperationID)
		if err != nil {
			return err
		}
		if persisted.Record.TerminalReceipt == nil || !receiptEquals(*persisted.Record.TerminalReceipt, receipt) ||
			persisted.PersistedOutcome == nil {
			return ErrConflict
		}
		parsed, _, err := parsePersistedOutcome(persisted.PersistedOutcome, receipt)
		if err != nil {
			return err
		}
		if !bytes.Equal(parsed.Evidence().CanonicalJCS(), proof.Evidence().CanonicalJCS()) ||
			parsed.Input().Material.Commitment.Digest() != proof.Input().Material.Commitment.Digest() {
			return ErrConflict
		}
		return nil
	})
	if err != nil {
		return Receipt{}, err
	}
	return cloneReceipt(receipt), nil
}

func (coordinator *Coordinator) captureActivation(ctx context.Context, receipt Receipt) (ValidatedActivationDecisionEvidence, error) {
	capture := BeginActivationEvidenceCapture()
	material, err := coordinator.dispatcher.CaptureActivationDecisionMaterial(ctx, receipt)
	if err != nil {
		return ValidatedActivationDecisionEvidence{}, err
	}
	head, err := coordinator.provider.Head(ctx)
	if err != nil {
		return ValidatedActivationDecisionEvidence{}, coordinatorDependencyError(ctx, err)
	}
	snapshot, err := NewAuthorityProviderHeadSnapshot(head)
	if err != nil {
		return ValidatedActivationDecisionEvidence{}, ErrInjectedFailure
	}
	input := ActivationDecisionEvidenceInput{Material: material, Receipt: cloneReceipt(receipt), ProviderHead: snapshot}
	if material.CheckpointKind != CheckpointNone {
		var checkpoint NodeCheckpoint
		switch material.CheckpointKind {
		case CheckpointNode:
			checkpoint, err = coordinator.provider.CommittedNodeCheckpoint(ctx, material.CheckpointScopeDigest)
		case CheckpointGlobal:
			var observed Record
			observed, err = coordinator.provider.Inspect(ctx, head.LatestCommittedOperationID)
			if err == nil {
				if observed.Validate() != nil || observed.TerminalReceipt == nil || observed.TerminalReceipt.Status != StatusCommitted ||
					observed.OperationID != head.LatestCommittedOperationID || observed.Epoch != head.Epoch ||
					observed.Sequence != head.LatestCommittedSequence || observed.TerminalReceipt.ReceiptDigest != head.LatestCommittedReceiptDigest ||
					!equalOptionalDatabasePoint(observed.TerminalReceipt.DatabasePoint, head.LatestCommittedDatabasePoint) ||
					observed.ScopeKind != material.Commitment.Facts().ScopeKind || observed.ScopeDigest != material.CheckpointScopeDigest {
					return ValidatedActivationDecisionEvidence{}, ErrConflict
				}
				checkpoint = NodeCheckpoint{AuthorityEpoch: observed.Epoch, Sequence: observed.Sequence, ReceiptDigest: observed.TerminalReceipt.ReceiptDigest}
			}
		default:
			return ValidatedActivationDecisionEvidence{}, ErrConflict
		}
		if err != nil {
			return ValidatedActivationDecisionEvidence{}, coordinatorDependencyError(ctx, err)
		}
		anchor, err := NewAuthorityCheckpointAnchor(AuthorityCheckpointAnchorInput{Kind: material.CheckpointKind, ScopeDigest: material.CheckpointScopeDigest, Checkpoint: checkpoint})
		if err != nil {
			return ValidatedActivationDecisionEvidence{}, ErrConflict
		}
		input.Checkpoint = &anchor
	}
	var evidence ActivationDecisionEvidence
	if material.TrustedTimeKind == TrustedTimeRollbackResistant {
		evidence, err = capture.Complete(input)
	} else {
		evidence, err = NewActivationDecisionEvidence(input)
	}
	if err != nil {
		return ValidatedActivationDecisionEvidence{}, coordinatorDependencyError(ctx, err)
	}
	proof, err := ValidateActivationDecisionEvidence(evidence, input)
	return proof, coordinatorDependencyError(ctx, err)
}

func (coordinator *Coordinator) validatePersistedOutcome(ctx context.Context, tx store.DBTX, fence StoredFence, receipt Receipt) error {
	if fence.PersistedOutcome == nil {
		return ErrConflict
	}
	proof, resolution, err := parsePersistedOutcome(fence.PersistedOutcome, receipt)
	if err != nil {
		return err
	}
	return coordinator.dispatcher.ValidatePersistedAuthorityEffect(ctx, tx, receipt, proof, resolution)
}

func parsePersistedOutcome(outcome *PersistedAuthorityEffectOutcome, receipt Receipt) (ValidatedActivationDecisionEvidence, AuthorityEffectResolution, error) {
	bad := func() (ValidatedActivationDecisionEvidence, AuthorityEffectResolution, error) {
		return ValidatedActivationDecisionEvidence{}, AuthorityEffectResolution{}, ErrConflict
	}
	if outcome == nil || !receiptEquals(outcome.Receipt, receipt) {
		return bad()
	}
	commitment, err := ParseAuthorityEffectCommitment(outcome.CommitmentJCS)
	if err != nil || commitment.Digest() != outcome.CommitmentDigest || !bytes.Equal(commitment.CanonicalJCS(), outcome.CommitmentJCS) {
		return bad()
	}
	head, err := ParseAuthorityProviderHeadSnapshot(outcome.ProviderHeadJCS)
	if err != nil || head.Digest() != outcome.ProviderHeadDigest || !bytes.Equal(head.CanonicalJCS(), outcome.ProviderHeadJCS) {
		return bad()
	}
	evidence, err := ParseActivationDecisionEvidence(outcome.EvidenceJCS)
	if err != nil || evidence.Digest() != outcome.EvidenceDigest || !bytes.Equal(evidence.CanonicalJCS(), outcome.EvidenceJCS) {
		return bad()
	}
	facts := evidence.Facts()
	material := ActivationDecisionMaterial{Commitment: commitment, Reason: outcome.Reason,
		CheckpointKind: facts.CheckpointKind, CheckpointScopeDigest: facts.CheckpointScopeDigest,
		TrustedTimeKind: facts.TrustedTimeKind, TrustedInstant: facts.TrustedInstant, EvidenceValidUntil: facts.EvidenceValidUntil,
		AttestationExpiresAt: outcome.AttestationExpiresAt, ActivationDeadline: outcome.ActivationDeadline,
		ProviderIdentityDigest: facts.ProviderIdentityDigest, ExpectedProviderIdentityDigest: outcome.ExpectedProviderIdentityDigest,
		FloorAttestationDigest: facts.FloorAttestationDigest, Capability: facts.Capability}
	input := ActivationDecisionEvidenceInput{Material: material, Receipt: cloneReceipt(receipt), ProviderHead: head}
	if facts.CheckpointKind != CheckpointNone {
		anchor, err := ParseAuthorityCheckpointAnchor(outcome.CheckpointAnchorJCS)
		if err != nil || anchor.Digest() != outcome.CheckpointAnchorDigest || !bytes.Equal(anchor.CanonicalJCS(), outcome.CheckpointAnchorJCS) {
			return bad()
		}
		input.Checkpoint = &anchor
	} else if len(outcome.CheckpointAnchorJCS) != 0 || outcome.CheckpointAnchorDigest != (contracts.Digest{}) {
		return bad()
	}
	proof, err := ValidateActivationDecisionEvidence(evidence, input)
	if err != nil {
		return bad()
	}
	resolution, err := ParseAuthorityEffectResolution(outcome.ResolutionJCS)
	if err != nil || resolution.Digest() != outcome.ResolutionDigest || !bytes.Equal(resolution.CanonicalJCS(), outcome.ResolutionJCS) ||
		ValidateAuthorityEffectResolution(resolution, commitment, proof) != nil {
		return bad()
	}
	return proof, resolution, nil
}

func (coordinator *Coordinator) transact(ctx context.Context, proof *ValidatedActivationDecisionEvidence, operation func(store.DBTX) error) error {
	tx, err := coordinator.transaction.beginAuthorityTransaction(ctx)
	if err != nil {
		return coordinatorDependencyError(ctx, err)
	}
	if nilAuthorityRepositoryValue(tx) {
		return ErrInjectedFailure
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	if err := operation(tx); err != nil {
		return coordinatorDependencyError(ctx, err)
	}
	if proof != nil && proof.Evidence().Facts().TrustedTimeKind == TrustedTimeRollbackResistant {
		if err := consumeActivationAdmission(ctx, *proof); err != nil {
			return coordinatorDependencyError(ctx, err)
		}
		// No observation, write, context check or helper call between admission
		// and this single Commit attempt.
		return coordinatorDependencyError(ctx, tx.Commit(ctx))
	}
	return coordinatorDependencyError(ctx, tx.Commit(ctx))
}

func cloneReceiptPointer(receipt *Receipt) *Receipt {
	if receipt == nil {
		return nil
	}
	result := cloneReceipt(*receipt)
	return &result
}

type readinessObservations struct {
	head    Head
	records map[uuid.UUID]Record
	errors  map[uuid.UUID]error
}

type authoritySnapshotReader interface {
	readAuthoritySnapshot(context.Context, store.DBTX, uint64) (DatabasePoint, DatabaseHead, []PendingFence, error)
}

func (coordinator *Coordinator) collectReadinessObservations(ctx context.Context) (readinessObservations, ReadinessReason, error) {
	head, err := coordinator.provider.Head(ctx)
	if err != nil || head.Validate() != nil {
		if coordinatorDependencyError(ctx, err) == ErrCanceled {
			return readinessObservations{}, "", ErrCanceled
		}
		return readinessObservations{}, ReadinessProviderUnavailable, nil
	}
	observation := readinessObservations{head: cloneHead(head), records: make(map[uuid.UUID]Record), errors: make(map[uuid.UUID]error)}
	pending, err := coordinator.repository.ListPending(ctx, head.Epoch)
	if err != nil {
		if coordinatorDependencyError(ctx, err) == ErrCanceled {
			return observation, "", ErrCanceled
		}
		return observation, ReadinessDatabaseUnavailable, nil
	}
	ids := make(map[uuid.UUID]struct{}, len(pending)+1)
	for _, fence := range pending {
		ids[fence.OperationID] = struct{}{}
	}
	if head.LatestCommittedOperationID != uuid.Nil {
		ids[head.LatestCommittedOperationID] = struct{}{}
	}
	for id := range ids {
		record, err := coordinator.provider.Inspect(ctx, id)
		if coordinatorDependencyError(ctx, err) == ErrCanceled {
			return observation, "", ErrCanceled
		}
		observation.records[id], observation.errors[id] = cloneRecord(record), err
	}
	return observation, "", nil
}

func (coordinator *Coordinator) CheckReady(ctx context.Context) (Readiness, error) {
	if err := coordinatorCallError(ctx, coordinator); err != nil {
		return Readiness{}, coordinatorInputError(ctx)
	}
	observations, reason, err := coordinator.collectReadinessObservations(ctx)
	if err != nil {
		return Readiness{}, err
	}
	if reason != "" {
		return unavailableReadiness(observations.head, DatabaseHead{}, reason), nil
	}
	var result Readiness
	err = coordinator.transact(ctx, nil, func(tx store.DBTX) error {
		var err error
		result, err = coordinator.checkReadyWithDBTX(ctx, tx, observations)
		return err
	})
	if err == ErrCanceled {
		return Readiness{}, err
	}
	if err != nil {
		return unavailableReadiness(observations.head, DatabaseHead{}, ReadinessDatabaseUnavailable), nil
	}
	return result, nil
}

// Provider observations must be collected before the caller acquires SQL
// locks. This DBTX-bound phase never performs an external call.
func (coordinator *Coordinator) checkReadyWithDBTX(ctx context.Context, tx store.DBTX, observations readinessObservations) (Readiness, error) {
	reader, ok := coordinator.repository.(authoritySnapshotReader)
	if !ok || nilAuthorityRepositoryValue(tx) {
		return Readiness{}, ErrInvalidArgument
	}
	providerHead := observations.head
	currentPoint, databaseHead, pending, err := reader.readAuthoritySnapshot(ctx, tx, providerHead.Epoch)
	if err != nil || currentPoint.Validate() != nil || !validDatabaseHead(databaseHead) {
		if coordinatorDependencyError(ctx, err) == ErrCanceled {
			return Readiness{}, ErrCanceled
		}
		return unavailableReadiness(providerHead, DatabaseHead{}, ReadinessDatabaseUnavailable), nil
	}
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

	if uint64(len(pending)) != databaseHead.PendingCount {
		return unavailableReadiness(providerHead, databaseHead, ReadinessPendingCountMismatch), nil
	}
	if providerHead.LatestCommittedOperationID != uuid.Nil {
		id := providerHead.LatestCommittedOperationID
		record, present := observations.records[id]
		if observations.errors[id] != nil {
			return unavailableReadiness(providerHead, databaseHead, ReadinessProviderUnavailable), nil
		}
		if !present || record.Validate() != nil || record.TerminalReceipt == nil || record.TerminalReceipt.Status != StatusCommitted ||
			record.Epoch != providerHead.Epoch || record.Sequence != providerHead.LatestCommittedSequence ||
			record.TerminalReceipt.ReceiptDigest != providerHead.LatestCommittedReceiptDigest {
			return unavailableReadiness(providerHead, databaseHead, ReadinessCommittedMismatch), nil
		}
		fence, effect, err := coordinator.lockResolved(ctx, tx, record.Reservation)
		if err == ErrCanceled {
			return Readiness{}, err
		}
		if err != nil || fence.Record.TerminalReceipt == nil || !receiptEquals(*fence.Record.TerminalReceipt, *record.TerminalReceipt) ||
			fence.AbortClaim != nil || effect.State != EffectTerminal {
			return unavailableReadiness(providerHead, databaseHead, ReadinessCommittedMismatch), nil
		}
		if err := coordinator.validatePersistedOutcome(ctx, tx, fence, *record.TerminalReceipt); err != nil {
			if err == ErrCanceled {
				return Readiness{}, err
			}
			return unavailableReadiness(providerHead, databaseHead, ReadinessEffectUnavailable), nil
		}
	}
	if len(pending) == 0 {
		return Readiness{ProviderHead: providerHead, DatabaseHead: databaseHead, Ready: true, Reason: ReadinessReady}, nil
	}
	reason := ReadinessPendingUnresolved
	for _, row := range pending {
		providerRecord, present := observations.records[row.OperationID]
		providerErr := observations.errors[row.OperationID]
		if providerErr != nil {
			candidate := ReadinessProviderUnavailable
			if errors.Is(providerErr, ErrNotFound) || errors.Is(providerErr, ErrConflict) || errors.Is(providerErr, ErrTerminalConflict) {
				candidate = ReadinessPendingMismatch
			}
			reason = readinessReasonWithPriority(reason, candidate)
		}
		reservation := Reservation{OperationID: row.OperationID, Kind: row.Kind, ScopeKind: row.ScopeKind, ScopeDigest: row.ScopeDigest,
			Epoch: row.Epoch, Sequence: row.Sequence, ReservationDigest: row.ReservationDigest}
		fence, effect, resolveErr := coordinator.lockResolved(ctx, tx, reservation)
		if resolveErr == ErrCanceled {
			return Readiness{}, resolveErr
		}
		if resolveErr != nil {
			reason = readinessReasonWithPriority(reason, ReadinessEffectUnavailable)
		}
		if !validPendingFence(row, providerHead.Epoch) || !present {
			reason = readinessReasonWithPriority(reason, ReadinessPendingMismatch)
		}
		if !validPendingFence(row, providerHead.Epoch) || !present || providerErr != nil || resolveErr != nil {
			continue
		}
		if !pendingStateMatches(row, providerRecord, effect) || fence.Record.TerminalReceipt != nil ||
			(fence.AbortClaim == nil) != (row.AbortClaim == nil) ||
			(fence.AbortClaim != nil && *fence.AbortClaim != *row.AbortClaim) {
			reason = readinessReasonWithPriority(reason, ReadinessPendingMismatch)
		}
	}
	return unavailableReadiness(providerHead, databaseHead, reason), nil
}

func (repository *PostgresRepository) readAuthoritySnapshot(ctx context.Context, dbtx store.DBTX, epoch uint64) (DatabasePoint, DatabaseHead, []PendingFence, error) {
	bound, err := NewPostgresRepository(dbtx)
	if err != nil {
		return DatabasePoint{}, DatabaseHead{}, nil, err
	}
	point, err := bound.CaptureDatabasePoint(ctx)
	if err != nil {
		return DatabasePoint{}, DatabaseHead{}, nil, err
	}
	head, err := bound.Head(ctx)
	if err != nil {
		return DatabasePoint{}, DatabaseHead{}, nil, err
	}
	pending, err := bound.ListPending(ctx, epoch)
	return point, head, pending, err
}

func (coordinator *Coordinator) CommittedNodeCheckpoint(ctx context.Context, scope contracts.Digest) (NodeCheckpoint, error) {
	if err := coordinatorCallError(ctx, coordinator); err != nil || !validScopeDigest(ScopeNode, scope) {
		return NodeCheckpoint{}, coordinatorInputError(ctx)
	}
	ready, err := coordinator.CheckReady(ctx)
	if err != nil {
		return NodeCheckpoint{}, err
	}
	if !ready.Ready {
		return NodeCheckpoint{}, ErrAuthorityUnavailable
	}
	checkpoint, err := coordinator.provider.CommittedNodeCheckpoint(ctx, scope)
	if err != nil {
		return NodeCheckpoint{}, coordinatorDependencyError(ctx, err)
	}
	if checkpoint.Validate() != nil || checkpoint.AuthorityEpoch != ready.ProviderHead.Epoch {
		return NodeCheckpoint{}, ErrConflict
	}
	databaseCheckpoint, err := coordinator.repository.CommittedNodeCheckpoint(ctx, checkpoint.AuthorityEpoch, scope)
	if err != nil {
		return NodeCheckpoint{}, coordinatorDependencyError(ctx, err)
	}
	if databaseCheckpoint != checkpoint {
		return NodeCheckpoint{}, ErrConflict
	}
	return checkpoint, nil
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

func (coordinator *Coordinator) resolve(ctx context.Context, operationID uuid.UUID) (ResolvedEffect, error) {
	stored, err := coordinator.repository.GetStoredFence(ctx, operationID)
	if err != nil {
		return ResolvedEffect{}, coordinatorDependencyError(ctx, err)
	}
	var effect ResolvedEffect
	err = coordinator.transact(ctx, nil, func(tx store.DBTX) error {
		_, resolved, err := coordinator.lockResolved(ctx, tx, stored.Record.Reservation)
		effect = resolved
		return err
	})
	return effect, err
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
		nilAuthorityRepositoryValue(coordinator.repository) || coordinator.dispatcher == nil ||
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

func (repository *PostgresRepository) beginAuthorityTransaction(ctx context.Context) (authorityTransaction, error) {
	if repositoryCallError(ctx, repository) != nil {
		return nil, coordinatorInputError(ctx)
	}
	beginner, ok := repository.database.(authorityTransactionBeginner)
	if !ok || nilAuthorityRepositoryValue(beginner) {
		return nil, ErrInvalidArgument
	}
	tx, err := beginner.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	return tx, coordinatorDependencyError(ctx, err)
}

func (repository *PostgresRepository) withAuthorityTransaction(ctx context.Context, operation func(store.DBTX) error) error {
	if operation == nil {
		return ErrInvalidArgument
	}
	tx, err := repository.beginAuthorityTransaction(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	if err := operation(tx); err != nil {
		return coordinatorDependencyError(ctx, err)
	}
	return coordinatorDependencyError(ctx, tx.Commit(ctx))
}
