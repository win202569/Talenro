package authority

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/store"
)

// coordinatorTestHandler is a deterministic domain fixture. Its persistence is
// part of the caller's transaction, never an out-of-band outcome cache.
type coordinatorTestHandler struct {
	source     EffectResolver
	repository Repository
}

func (handler *coordinatorTestHandler) memory() *coordinatorEffectResolver {
	switch source := handler.source.(type) {
	case *coordinatorEffectResolver:
		return source
	case *coordinatorFirstResolveBlockingResolver:
		memory, _ := source.EffectResolver.(*coordinatorEffectResolver)
		return memory
	default:
		return nil
	}
}

func (handler *coordinatorTestHandler) ResolveRegisteredAuthorityEffectForUpdate(ctx context.Context, dbtx store.DBTX, query TransactionalEffectQuery) (TransactionalResolvedEffect, error) {
	expected := query.Expected()
	result := TransactionalResolvedEffect{OperationID: expected.OperationID, Epoch: expected.Epoch, Sequence: expected.Sequence, Effect: ResolvedEffect{State: EffectAbsent}}
	if query.RegisteredKind() != expected.Kind {
		return result, nil
	}
	tx, ok := dbtx.(*coordinatorMemoryTransaction)
	if !ok {
		if registered, ok := handler.source.(RegisteredEffectResolver); ok {
			return registered.ResolveRegisteredAuthorityEffectForUpdate(ctx, dbtx, query)
		}
		return result, ErrConflict
	}
	if tx.owner != handler.repository {
		return result, ErrConflict
	}
	resolved, err := handler.source.ResolveAuthorityEffect(ctx, expected.OperationID)
	if err != nil {
		return result, err
	}
	if tx.shadow.outcomes[expected.OperationID] != nil {
		resolved.State = EffectTerminal
	}
	result.Effect = resolved
	return result, nil
}

func (handler *coordinatorTestHandler) CaptureActivationDecisionMaterial(ctx context.Context, receipt Receipt) (ActivationDecisionMaterial, error) {
	if registered, ok := handler.source.(RegisteredEffectActivator); ok {
		return registered.CaptureActivationDecisionMaterial(ctx, receipt)
	}
	memory := handler.memory()
	if memory == nil {
		return ActivationDecisionMaterial{}, ErrConflict
	}
	repository, ok := handler.repository.(*coordinatorMemoryRepository)
	if !ok {
		return ActivationDecisionMaterial{}, ErrConflict
	}
	repository.mu.Lock()
	if repository.active != nil {
		repository.mu.Unlock()
		return ActivationDecisionMaterial{}, ErrConflict
	}
	repository.events = append(repository.events, "material")
	repository.mu.Unlock()
	memory.mu.Lock()
	defer memory.mu.Unlock()
	memory.captures++
	if memory.captureErr != nil {
		return ActivationDecisionMaterial{}, memory.captureErr
	}
	material, ok := memory.materials[receipt.OperationID]
	if !ok {
		return ActivationDecisionMaterial{}, ErrConflict
	}
	return cloneActivationDecisionMaterial(material), nil
}

func coordinatorTestResolution(proof ValidatedActivationDecisionEvidence) (AuthorityEffectResolution, error) {
	input := proof.Input()
	resolution := AuthorityEffectResolutionInput{Commitment: input.Material.Commitment, Evidence: proof,
		Disposition: DispositionNotApplied, Reason: input.Material.Reason}
	switch {
	case input.Material.Commitment.Facts().Mode == CommitmentFinalNotApplied:
		resolution.AnchorKind, resolution.AnchorDigest = DecisionAnchorFinalCommitment, input.Material.Commitment.Digest()
	case input.Material.CheckpointKind != CheckpointNone:
		resolution.AnchorKind, resolution.AnchorDigest = DecisionAnchorHigherAuthority, input.Checkpoint.Digest()
	case input.Material.Capability == DecisionCapabilityMayApply:
		resolution.Disposition, resolution.Reason = DispositionApplied, EffectReasonNone
		resolution.AnchorKind, resolution.AnchorDigest = DecisionAnchorTrustedTime, proof.Evidence().Digest()
	case input.Material.Reason == EffectReasonActivationDeadlineExpired:
		resolution.AnchorKind, resolution.AnchorDigest = DecisionAnchorTrustedTime, proof.Evidence().Digest()
	default:
		resolution.AnchorKind, resolution.AnchorDigest = DecisionAnchorExactCapture, input.Material.Commitment.Facts().ActivationInputsDigest
	}
	return NewAuthorityEffectResolution(resolution)
}

func coordinatorTestOutcome(proof ValidatedActivationDecisionEvidence, resolution AuthorityEffectResolution) *PersistedAuthorityEffectOutcome {
	input, evidence := proof.Input(), proof.Evidence()
	outcome := &PersistedAuthorityEffectOutcome{CommitmentJCS: input.Material.Commitment.CanonicalJCS(), CommitmentDigest: input.Material.Commitment.Digest(),
		Receipt: cloneReceipt(input.Receipt), ProviderHeadJCS: input.ProviderHead.CanonicalJCS(), ProviderHeadDigest: input.ProviderHead.Digest(),
		Reason: input.Material.Reason, AttestationExpiresAt: input.Material.AttestationExpiresAt, ActivationDeadline: input.Material.ActivationDeadline,
		ExpectedProviderIdentityDigest: input.Material.ExpectedProviderIdentityDigest, EvidenceJCS: evidence.CanonicalJCS(), EvidenceDigest: evidence.Digest(),
		ResolutionJCS: resolution.CanonicalJCS(), ResolutionDigest: resolution.Digest()}
	if input.Checkpoint != nil {
		outcome.CheckpointAnchorJCS, outcome.CheckpointAnchorDigest = input.Checkpoint.CanonicalJCS(), input.Checkpoint.Digest()
	}
	return outcome
}

func (handler *coordinatorTestHandler) ActivateAuthorityEffect(ctx context.Context, dbtx store.DBTX, receipt Receipt, proof ValidatedActivationDecisionEvidence) error {
	if registered, ok := handler.source.(RegisteredEffectActivator); ok {
		return registered.ActivateAuthorityEffect(ctx, dbtx, receipt, proof)
	}
	tx, ok := dbtx.(*coordinatorMemoryTransaction)
	memory := handler.memory()
	if !ok || memory == nil || tx.owner != handler.repository || proof.origin != activationEvidenceOriginFresh {
		return ErrConflict
	}
	input := proof.Input()
	memory.mu.Lock()
	material := memory.materials[receipt.OperationID]
	memory.activations++
	hook, activationErr := memory.activationHook, memory.activationErr
	memory.mu.Unlock()
	if !bytes.Equal(material.Commitment.CanonicalJCS(), input.Material.Commitment.CanonicalJCS()) {
		return ErrConflict
	}
	if material.Commitment.Facts().Mode == CommitmentConditionalApply {
		want := sha256.Sum256([]byte("task9-typed-activation-input:" + receipt.OperationID.String()))
		if material.Commitment.Facts().ActivationInputsDigest != want {
			return ErrConflict
		}
	}
	record := tx.shadow.records[receipt.OperationID]
	if record.TerminalReceipt == nil || !receiptEquals(*record.TerminalReceipt, receipt) {
		return ErrConflict
	}
	resolution, err := coordinatorTestResolution(proof)
	if err != nil {
		return err
	}
	tx.shadow.outcomes[receipt.OperationID] = coordinatorTestOutcome(proof, resolution)
	tx.writes = append(tx.writes, "domain", "preimages", "audit", "outbox")
	if hook != nil {
		hook(proof)
	}
	return activationErr
}

func (handler *coordinatorTestHandler) ValidatePersistedAuthorityEffect(ctx context.Context, dbtx store.DBTX, receipt Receipt, proof ValidatedActivationDecisionEvidence, resolution AuthorityEffectResolution) error {
	if registered, ok := handler.source.(RegisteredEffectActivator); ok {
		return registered.ValidatePersistedAuthorityEffect(ctx, dbtx, receipt, proof, resolution)
	}
	tx, ok := dbtx.(*coordinatorMemoryTransaction)
	memory := handler.memory()
	if !ok || memory == nil || tx.owner != handler.repository || proof.origin != activationEvidenceOriginParsed {
		return ErrConflict
	}
	memory.mu.Lock()
	memory.validations++
	material := memory.materials[receipt.OperationID]
	err := memory.validationErr
	memory.mu.Unlock()
	if err != nil {
		return err
	}
	persisted := tx.shadow.outcomes[receipt.OperationID]
	if persisted == nil || !bytes.Equal(persisted.CommitmentJCS, material.Commitment.CanonicalJCS()) ||
		persisted.ResolutionDigest != resolution.Digest() || persisted.EvidenceDigest != proof.Evidence().Digest() {
		return ErrConflict
	}
	if material.Commitment.Facts().Mode == CommitmentConditionalApply &&
		material.Commitment.Facts().ActivationInputsDigest != sha256.Sum256([]byte("task9-typed-activation-input:"+receipt.OperationID.String())) {
		return ErrConflict
	}
	return nil
}

type coordinatorMemoryTransaction struct {
	coordinatorNoopDBTX
	owner   *coordinatorMemoryRepository
	shadow  *coordinatorMemoryRepository
	failure *coordinatorTransactionFailure
	done    bool
	writes  []string
	commits int
}

func (repository *coordinatorMemoryRepository) beginAuthorityTransaction(ctx context.Context) (authorityTransaction, error) {
	if ctx == nil || ctx.Err() != nil {
		return nil, ErrCanceled
	}
	repository.txMu.Lock()
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.transactions++
	failure := repository.transactionFail
	if failure != nil && failure.call == repository.transactions {
		repository.transactionFail = nil
		if !failure.after {
			repository.txMu.Unlock()
			return nil, failure.err
		}
	} else {
		failure = nil
	}
	shadow := newCoordinatorMemoryRepository(repository.points)
	for id, record := range repository.records {
		shadow.records[id] = cloneRecord(record)
	}
	for id, value := range repository.reservedAt {
		shadow.reservedAt[id] = value
	}
	for id, value := range repository.boundAt {
		shadow.boundAt[id] = value
	}
	for id, value := range repository.terminalAt {
		shadow.terminalAt[id] = value
	}
	for id, value := range repository.claims {
		shadow.claims[id] = cloneAbortClaim(value)
	}
	for id, value := range repository.outcomes {
		shadow.outcomes[id] = cloneStoredFence(StoredFence{PersistedOutcome: value}).PersistedOutcome
	}
	shadow.headErr, shadow.headOverride = repository.headErr, repository.headOverride
	shadow.captures = repository.captures
	tx := &coordinatorMemoryTransaction{owner: repository, shadow: shadow, failure: failure}
	repository.active = tx
	repository.events = append(repository.events, "begin")
	return tx, nil
}

func (tx *coordinatorMemoryTransaction) Commit(ctx context.Context) error {
	if tx.done {
		return ErrConflict
	}
	tx.commits++
	if ctx.Err() != nil {
		_ = tx.Rollback(context.WithoutCancel(ctx))
		return ErrCanceled
	}
	repository := tx.owner
	repository.mu.Lock()
	repository.records, repository.reservedAt, repository.boundAt, repository.terminalAt = tx.shadow.records, tx.shadow.reservedAt, tx.shadow.boundAt, tx.shadow.terminalAt
	repository.claims, repository.outcomes = tx.shadow.claims, tx.shadow.outcomes
	repository.events = append(repository.events, tx.writes...)
	repository.events = append(repository.events, "commit")
	repository.active = nil
	tx.done = true
	repository.mu.Unlock()
	repository.txMu.Unlock()
	if tx.failure != nil {
		return tx.failure.err
	}
	return nil
}

func (tx *coordinatorMemoryTransaction) Rollback(context.Context) error {
	if tx.done {
		return nil
	}
	tx.owner.mu.Lock()
	tx.owner.active = nil
	tx.owner.events = append(tx.owner.events, "rollback")
	tx.done = true
	tx.owner.mu.Unlock()
	tx.owner.txMu.Unlock()
	return nil
}

func (repository *coordinatorMemoryRepository) GetStoredFence(ctx context.Context, id uuid.UUID) (StoredFence, error) {
	record, err := repository.Get(ctx, id)
	if err != nil {
		return StoredFence{}, err
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	return cloneStoredFence(StoredFence{Record: record, AbortClaim: repository.claims[id], PersistedOutcome: repository.outcomes[id]}), nil
}

func (repository *coordinatorMemoryRepository) Lock(ctx context.Context, dbtx store.DBTX, id uuid.UUID) (StoredFence, error) {
	tx, ok := dbtx.(*coordinatorMemoryTransaction)
	if !ok || tx.owner != repository || tx.done {
		return StoredFence{}, ErrConflict
	}
	return tx.shadow.GetStoredFence(ctx, id)
}

func (repository *coordinatorMemoryRepository) ClaimAbort(ctx context.Context, dbtx store.DBTX, id uuid.UUID, reason AbortReason, at time.Time) error {
	tx, ok := dbtx.(*coordinatorMemoryTransaction)
	if !ok || tx.owner != repository || tx.done || ctx.Err() != nil {
		return ErrConflict
	}
	record, present := tx.shadow.records[id]
	if !present || record.TerminalReceipt != nil || record.BoundEffectDigest != nil || record.BoundDatabasePoint != nil {
		return ErrConflict
	}
	if old := tx.shadow.claims[id]; old != nil {
		if old.Reason != reason || !old.ClaimedAt.Equal(at) {
			return ErrConflict
		}
		return nil
	}
	tx.shadow.claims[id] = &AbortClaim{Reason: reason, ClaimedAt: at}
	tx.writes = append(tx.writes, "claim")
	return nil
}

func (repository *coordinatorMemoryRepository) RecordPendingClaimV1(ctx context.Context, dbtx store.DBTX, value ClaimV1Reservation, at time.Time) error {
	if value.ProtocolActivationID == uuid.Nil {
		return ErrInvalidArgument
	}
	return repository.RecordPending(ctx, dbtx, value.Reservation, at)
}

func (repository *coordinatorMemoryRepository) readAuthoritySnapshot(ctx context.Context, dbtx store.DBTX, epoch uint64) (DatabasePoint, DatabaseHead, []PendingFence, error) {
	tx, ok := dbtx.(*coordinatorMemoryTransaction)
	if !ok || tx.owner != repository {
		return DatabasePoint{}, DatabaseHead{}, nil, ErrConflict
	}
	point, err := tx.shadow.CaptureDatabasePoint(ctx)
	if err != nil {
		return DatabasePoint{}, DatabaseHead{}, nil, err
	}
	head, err := tx.shadow.Head(ctx)
	if err != nil {
		return DatabasePoint{}, DatabaseHead{}, nil, err
	}
	pending, err := tx.shadow.ListPending(ctx, epoch)
	return point, head, pending, err
}

type task9ObservedProvider struct {
	Provider
	repository       *coordinatorMemoryRepository
	headFailure      error
	finalizeMismatch bool
}

func (provider *task9ObservedProvider) observe(name string) error {
	provider.repository.mu.Lock()
	defer provider.repository.mu.Unlock()
	if provider.repository.active != nil {
		return ErrConflict
	}
	provider.repository.events = append(provider.repository.events, name)
	return nil
}
func (provider *task9ObservedProvider) Reserve(ctx context.Context, request ReserveRequest) (Reservation, error) {
	if err := provider.observe("reserve"); err != nil {
		return Reservation{}, err
	}
	return provider.Provider.Reserve(ctx, request)
}
func (provider *task9ObservedProvider) Inspect(ctx context.Context, id uuid.UUID) (Record, error) {
	if err := provider.observe("inspect"); err != nil {
		return Record{}, err
	}
	return provider.Provider.Inspect(ctx, id)
}
func (provider *task9ObservedProvider) Abort(ctx context.Context, request AbortRequest) (Receipt, error) {
	if err := provider.observe("abort"); err != nil {
		return Receipt{}, err
	}
	return provider.Provider.Abort(ctx, request)
}
func (provider *task9ObservedProvider) Finalize(ctx context.Context, request FinalizeRequest) (Receipt, error) {
	if err := provider.observe("provider-finalize"); err != nil {
		return Receipt{}, err
	}
	receipt, err := provider.Provider.Finalize(ctx, request)
	if err == nil && provider.finalizeMismatch {
		receipt.DatabasePoint.Timeline++
		receipt.ReceiptDigest, _ = receiptDigest(receipt)
	}
	return receipt, err
}
func (provider *task9ObservedProvider) Head(ctx context.Context) (Head, error) {
	if err := provider.observe("head"); err != nil {
		return Head{}, err
	}
	if provider.headFailure != nil {
		return Head{}, provider.headFailure
	}
	return provider.Provider.Head(ctx)
}
func (provider *task9ObservedProvider) CommittedNodeCheckpoint(ctx context.Context, scope contracts.Digest) (NodeCheckpoint, error) {
	if err := provider.observe("checkpoint"); err != nil {
		return NodeCheckpoint{}, err
	}
	return provider.Provider.CommittedNodeCheckpoint(ctx, scope)
}

type task9Fixture struct {
	coordinator *Coordinator
	provider    *task9ObservedProvider
	raw         *DeterministicProvider
	repository  *coordinatorMemoryRepository
	effects     *coordinatorEffectResolver
	request     ReserveRequest
	digest      contracts.Digest
}

func newTask9Fixture(t *testing.T, mode string) *task9Fixture {
	t.Helper()
	raw, err := NewDeterministicProvider(7)
	if err != nil {
		t.Fatal(err)
	}
	repository := newCoordinatorMemoryRepository([]DatabasePoint{{SystemID: 41, Timeline: 3, RequiredLSN: "0/30"}})
	provider := &task9ObservedProvider{Provider: raw, repository: repository}
	effects := newCoordinatorEffectResolver()
	effects.conditional = mode == "conditional"
	coordinator := mustNewCoordinatorForTest(t, provider, repository, effects, coordinatorClock{now: time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)})
	request := ReserveRequest{OperationID: uuid.New(), Kind: EffectCertificateRevoke, ScopeKind: ScopeNode, ScopeDigest: sha256.Sum256([]byte("task9-node"))}
	coordinatorReserveAndRecord(t, coordinator, repository, request)
	var digest contracts.Digest
	if mode != "absent" {
		digest = effects.commit(t, request, sha256.Sum256([]byte("task9-domain-base")), "0/20")
	}
	return &task9Fixture{coordinator, provider, raw, repository, effects, request, digest}
}

type task9EmbeddedDispatcher struct{ EffectDispatcher }

func TestCoordinatorAbortLinearization(t *testing.T) {
	for _, scenario := range []string{"provider response loss", "claim commit response loss", "lost claim", "reserved absent", "resolver failure"} {
		t.Run(scenario, func(t *testing.T) {
			f := newTask9Fixture(t, "absent")
			request := AbortRequest{OperationID: f.request.OperationID, Reason: AbortProviderDependencyFailed}
			switch scenario {
			case "provider response loss":
				_ = f.raw.FailNext(OperationAbort, FailureAfterMutationResponseLost)
				if _, err := f.coordinator.Abort(t.Context(), request); err != ErrResponseLost {
					t.Fatalf("Abort: %v", err)
				}
			case "claim commit response loss":
				f.repository.failTransaction(1, true, ErrResponseLost)
				if _, err := f.coordinator.Abort(t.Context(), request); err != ErrResponseLost {
					t.Fatalf("claim commit: %v", err)
				}
				record, _ := f.raw.Inspect(t.Context(), request.OperationID)
				if record.TerminalReceipt != nil {
					t.Fatal("provider called before confirmed claim commit")
				}
			case "lost claim":
				if _, err := f.raw.Abort(t.Context(), request); err != nil {
					t.Fatal(err)
				}
			case "reserved absent":
				if _, err := f.coordinator.Recover(t.Context(), request.OperationID); err != ErrConflict {
					t.Fatalf("guessed absent reason: %v", err)
				}
				if f.repository.claims[request.OperationID] != nil {
					t.Fatal("invented Abort reason")
				}
				return
			case "resolver failure":
				f.effects.setError(request.OperationID, errors.New("private SQL"))
				if _, err := f.coordinator.Abort(t.Context(), request); err != ErrInjectedFailure {
					t.Fatalf("resolve: %v", err)
				}
				if f.repository.claims[request.OperationID] != nil || f.raw.Snapshot()[0].TerminalReceipt != nil {
					t.Fatal("failed resolver retained claim or provider mutation")
				}
				return
			}
			receipt, err := f.coordinator.Recover(t.Context(), request.OperationID)
			if err != nil || receipt.AbortReason == nil || *receipt.AbortReason != request.Reason {
				t.Fatalf("exact recovery: %v", err)
			}
			claim := *f.repository.claims[request.OperationID]
			if _, err := f.coordinator.Recover(t.Context(), request.OperationID); err != nil {
				t.Fatal(err)
			}
			if *f.repository.claims[request.OperationID] != claim {
				t.Fatal("terminal replay rewrote claim")
			}
			firstCommit, providerAbort := -1, -1
			for i, event := range f.repository.events {
				if event == "commit" && firstCommit < 0 {
					firstCommit = i
				}
				if event == "abort" && providerAbort < 0 {
					providerAbort = i
				}
			}
			if providerAbort >= 0 && (firstCommit < 0 || firstCommit >= providerAbort) {
				t.Fatal("provider Abort preceded released claim transaction")
			}
		})
	}
}

func TestCoordinatorEvidenceCaptureOrder(t *testing.T) {
	for _, mode := range []string{"final", "conditional"} {
		t.Run(mode, func(t *testing.T) {
			f := newTask9Fixture(t, mode)
			var captured ValidatedActivationDecisionEvidence
			f.effects.activationHook = func(proof ValidatedActivationDecisionEvidence) {
				captured = proof
				if proof.origin != activationEvidenceOriginFresh {
					t.Error("activation received parsed proof")
				}
				if mode == "conditional" && (proof.admission == nil || proof.admission.state.Load() != 0) {
					t.Error("admission consumed before writes completed")
				}
				if mode == "final" && proof.admission != nil {
					t.Error("none-time proof acquired token")
				}
			}
			if _, err := f.coordinator.Finalize(t.Context(), CoordinatorFinalizeRequest{f.request.OperationID, f.digest}); err != nil {
				t.Fatal(err)
			}
			if mode == "conditional" && captured.admission.state.Load() != 1 {
				t.Fatal("activation did not consume admission")
			}
			last := -1
			for _, want := range []string{"provider-finalize", "material", "head", "domain", "preimages", "audit", "outbox", "commit"} {
				found := -1
				for i := last + 1; i < len(f.repository.events); i++ {
					if f.repository.events[i] == want {
						found = i
						break
					}
				}
				if found < 0 {
					t.Fatalf("missing ordered %s in %v", want, f.repository.events)
				}
				last = found
			}
			if f.effects.captures != 1 || f.effects.activations != 1 {
				t.Fatal("duplicate capture or activation")
			}
		})
	}
}

func TestCoordinatorAtomicActivationCrashMatrix(t *testing.T) {
	for _, scenario := range []string{"capture failure", "Head failure", "activation write failure", "expired admission", "mismatched admission", "spent admission", "Commit response loss", "Receipt database point"} {
		t.Run(scenario, func(t *testing.T) {
			f := newTask9Fixture(t, "conditional")
			var spent *activationAdmission
			want := ErrInjectedFailure
			switch scenario {
			case "capture failure":
				f.effects.captureErr = errors.New("private capture")
			case "Head failure":
				f.provider.headFailure = errors.New("private provider")
			case "activation write failure":
				f.effects.activationErr = errors.New("private SQL after writes")
			case "expired admission":
				want = ErrConflict
				f.effects.activationHook = func(p ValidatedActivationDecisionEvidence) {
					spent = p.admission
					p.admission.now = func() time.Time { return p.admission.deadline }
				}
			case "mismatched admission":
				want = ErrConflict
				f.effects.activationHook = func(p ValidatedActivationDecisionEvidence) { spent = p.admission; p.admission.operationID = uuid.New() }
			case "spent admission":
				want = ErrConflict
				f.effects.activationHook = func(p ValidatedActivationDecisionEvidence) {
					spent = p.admission
					if err := consumeActivationAdmission(t.Context(), p); err != nil {
						t.Error(err)
					}
				}
			case "Commit response loss":
				f.repository.failTransaction(4, true, ErrResponseLost)
				want = ErrResponseLost
			case "Receipt database point":
				f.provider.finalizeMismatch = true
			}
			_, err := f.coordinator.Finalize(t.Context(), CoordinatorFinalizeRequest{f.request.OperationID, f.digest})
			if err != want {
				t.Fatalf("error=%v,want=%v", err, want)
			}
			if spent != nil && spent.state.Load() != 1 {
				t.Fatal("failed consume did not burn shared admission")
			}
			stored, readErr := f.repository.GetStoredFence(t.Context(), f.request.OperationID)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if scenario == "Commit response loss" {
				if stored.Record.TerminalReceipt == nil || stored.PersistedOutcome == nil {
					t.Fatal("lost response lost committed transaction")
				}
				before := cloneStoredFence(stored)
				captures := f.effects.captures
				f.effects.captureErr = errors.New("recapture forbidden")
				if _, err := f.coordinator.Recover(t.Context(), f.request.OperationID); err != nil {
					t.Fatal(err)
				}
				after, _ := f.repository.GetStoredFence(t.Context(), f.request.OperationID)
				if !reflect.DeepEqual(before, after) || f.effects.captures != captures || f.effects.validations == 0 {
					t.Fatal("uncertain Commit did not use unchanged parsed outcome")
				}
			} else if stored.Record.TerminalReceipt != nil || stored.PersistedOutcome != nil {
				t.Fatal("fence/domain/preimages split across transaction")
			}
		})
	}
}

func TestCoordinatorPersistedOutcomeRecovery(t *testing.T) {
	mutations := []struct {
		name   string
		change func(*PersistedAuthorityEffectOutcome)
	}{
		{"commitment missing", func(o *PersistedAuthorityEffectOutcome) { o.CommitmentJCS = nil }},
		{"commitment digest", func(o *PersistedAuthorityEffectOutcome) { o.CommitmentDigest[0] ^= 1 }},
		{"Head missing", func(o *PersistedAuthorityEffectOutcome) { o.ProviderHeadJCS = nil }},
		{"Head digest", func(o *PersistedAuthorityEffectOutcome) { o.ProviderHeadDigest[0] ^= 1 }},
		{"unexpected checkpoint", func(o *PersistedAuthorityEffectOutcome) { o.CheckpointAnchorJCS = []byte("{}") }},
		{"evidence missing", func(o *PersistedAuthorityEffectOutcome) { o.EvidenceJCS = nil }},
		{"evidence digest", func(o *PersistedAuthorityEffectOutcome) { o.EvidenceDigest[0] ^= 1 }},
		{"resolution missing", func(o *PersistedAuthorityEffectOutcome) { o.ResolutionJCS = nil }},
		{"resolution digest", func(o *PersistedAuthorityEffectOutcome) { o.ResolutionDigest[0] ^= 1 }},
		{"reason", func(o *PersistedAuthorityEffectOutcome) { o.Reason = EffectReasonFailed }},
		{"expected identity", func(o *PersistedAuthorityEffectOutcome) { o.ExpectedProviderIdentityDigest[0] ^= 1 }},
		{"deadline", func(o *PersistedAuthorityEffectOutcome) { o.ActivationDeadline = time.Time{} }},
		{"expiry", func(o *PersistedAuthorityEffectOutcome) { o.AttestationExpiresAt = time.Time{} }},
		{"Receipt database point", func(o *PersistedAuthorityEffectOutcome) { o.Receipt.DatabasePoint.Timeline++ }},
		{"noncanonical Head", func(o *PersistedAuthorityEffectOutcome) {
			o.ProviderHeadJCS = append([]byte(" "), o.ProviderHeadJCS...)
		}},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			f := newTask9Fixture(t, "conditional")
			if _, err := f.coordinator.Finalize(t.Context(), CoordinatorFinalizeRequest{f.request.OperationID, f.digest}); err != nil {
				t.Fatal(err)
			}
			captures, activations := f.effects.captures, f.effects.activations
			mutation.change(f.repository.outcomes[f.request.OperationID])
			if _, err := f.coordinator.Recover(t.Context(), f.request.OperationID); err != ErrConflict {
				t.Fatalf("corrupt persisted outcome error=%v", err)
			}
			if f.effects.captures != captures || f.effects.activations != activations {
				t.Fatal("corruption triggered recapture or activation")
			}
			ready, err := f.coordinator.CheckReady(t.Context())
			if err != nil || ready.Ready {
				t.Fatalf("corrupt outcome ready=%t,error=%v", ready.Ready, err)
			}
		})
	}
}

func TestCoordinatorDispatcherConstruction(t *testing.T) {
	factory := reflect.TypeOf(NewCoordinator)
	if factory.NumIn() != 3 || factory.In(2) != reflect.TypeFor[EffectDispatcher]() {
		t.Fatalf("constructor must accept exactly Provider, Repository, EffectDispatcher; got %v", factory)
	}
	f := newTask9Fixture(t, "absent")
	var typedNil *effectDispatcher
	for _, bad := range []EffectDispatcher{nil, typedNil, &task9EmbeddedDispatcher{f.coordinator.dispatcher}, task9EmbeddedDispatcher{f.coordinator.dispatcher}} {
		before := len(f.repository.events)
		if coordinator, err := NewCoordinator(f.provider, f.repository, bad); err != ErrInvalidArgument || coordinator != nil {
			t.Fatalf("nonexact %T accepted: %v", bad, err)
		}
		if len(f.repository.events) != before {
			t.Fatal("constructor invoked dependency before type rejection")
		}
	}
}

func TestCoordinatorCrashRecoveryMatrix(t *testing.T) {
	t.Parallel()

	type crashPoint struct {
		name              string
		prepare           func(*testing.T, *Coordinator, *DeterministicProvider, *coordinatorMemoryRepository, *coordinatorEffectResolver, ReserveRequest, contracts.Digest)
		wantStatus        ReceiptStatus
		wantEffectCommits int
		boundBeforeCrash  bool
	}

	points := []crashPoint{
		{
			name: "provider reserve committed before response",
			prepare: func(t *testing.T, coordinator *Coordinator, provider *DeterministicProvider, _ *coordinatorMemoryRepository, _ *coordinatorEffectResolver, request ReserveRequest, _ contracts.Digest) {
				t.Helper()
				if err := provider.FailNext(OperationReserve, FailureAfterMutationResponseLost); err != nil {
					t.Fatal(err)
				}
				if _, err := coordinator.Reserve(t.Context(), request); err != ErrResponseLost {
					t.Fatalf("Reserve error = %v, want ErrResponseLost", err)
				}
			},
			wantStatus: StatusAborted,
		},
		{
			name: "DB pending commit before caller response",
			prepare: func(t *testing.T, coordinator *Coordinator, _ *DeterministicProvider, repository *coordinatorMemoryRepository, _ *coordinatorEffectResolver, request ReserveRequest, _ contracts.Digest) {
				t.Helper()
				reservation := coordinatorReserveAndRecord(t, coordinator, repository, request)
				if reservation.Sequence != 1 {
					t.Fatalf("reservation sequence = %d, want 1", reservation.Sequence)
				}
			},
			wantStatus: StatusAborted,
		},
		{
			name: "domain effect transaction committed before database-point capture",
			prepare: func(t *testing.T, coordinator *Coordinator, _ *DeterministicProvider, repository *coordinatorMemoryRepository, effects *coordinatorEffectResolver, request ReserveRequest, effectDigest contracts.Digest) {
				t.Helper()
				coordinatorReserveAndRecord(t, coordinator, repository, request)
				effectDigest = effects.commit(t, request, effectDigest, "0/20")
			},
			wantStatus:        StatusCommitted,
			wantEffectCommits: 1,
		},
		{
			name: "database point captured before bind",
			prepare: func(t *testing.T, coordinator *Coordinator, _ *DeterministicProvider, repository *coordinatorMemoryRepository, effects *coordinatorEffectResolver, request ReserveRequest, effectDigest contracts.Digest) {
				t.Helper()
				coordinatorReserveAndRecord(t, coordinator, repository, request)
				effectDigest = effects.commit(t, request, effectDigest, "0/20")
				repository.failTransaction(2, false, ErrInjectedFailure)
				if _, err := coordinator.Finalize(t.Context(), CoordinatorFinalizeRequest{OperationID: request.OperationID, EffectDigest: effectDigest}); err != ErrInjectedFailure {
					t.Fatalf("Finalize error = %v, want ErrInjectedFailure", err)
				}
				if repository.captureCount() != 1 {
					t.Fatalf("capture count = %d, want 1", repository.captureCount())
				}
			},
			wantStatus:        StatusCommitted,
			wantEffectCommits: 1,
		},
		{
			name: "DB effect/database bind before provider finalize",
			prepare: func(t *testing.T, coordinator *Coordinator, provider *DeterministicProvider, repository *coordinatorMemoryRepository, effects *coordinatorEffectResolver, request ReserveRequest, effectDigest contracts.Digest) {
				t.Helper()
				coordinatorReserveAndRecord(t, coordinator, repository, request)
				effectDigest = effects.commit(t, request, effectDigest, "0/20")
				if err := provider.FailNext(OperationFinalize, FailureBeforeMutation); err != nil {
					t.Fatal(err)
				}
				if _, err := coordinator.Finalize(t.Context(), CoordinatorFinalizeRequest{OperationID: request.OperationID, EffectDigest: effectDigest}); err != ErrInjectedFailure {
					t.Fatalf("Finalize error = %v, want ErrInjectedFailure", err)
				}
			},
			wantStatus:        StatusCommitted,
			wantEffectCommits: 1,
			boundBeforeCrash:  true,
		},
		{
			name: "provider finalize committed before response",
			prepare: func(t *testing.T, coordinator *Coordinator, provider *DeterministicProvider, repository *coordinatorMemoryRepository, effects *coordinatorEffectResolver, request ReserveRequest, effectDigest contracts.Digest) {
				t.Helper()
				coordinatorReserveAndRecord(t, coordinator, repository, request)
				effectDigest = effects.commit(t, request, effectDigest, "0/20")
				if err := provider.FailNext(OperationFinalize, FailureAfterMutationResponseLost); err != nil {
					t.Fatal(err)
				}
				if _, err := coordinator.Finalize(t.Context(), CoordinatorFinalizeRequest{OperationID: request.OperationID, EffectDigest: effectDigest}); err != ErrResponseLost {
					t.Fatalf("Finalize error = %v, want ErrResponseLost", err)
				}
			},
			wantStatus:        StatusCommitted,
			wantEffectCommits: 1,
			boundBeforeCrash:  true,
		},
		{
			name: "provider finalize response before visibility activation",
			prepare: func(t *testing.T, coordinator *Coordinator, _ *DeterministicProvider, repository *coordinatorMemoryRepository, effects *coordinatorEffectResolver, request ReserveRequest, effectDigest contracts.Digest) {
				t.Helper()
				coordinatorReserveAndRecord(t, coordinator, repository, request)
				effectDigest = effects.commit(t, request, effectDigest, "0/20")
				repository.failTransaction(4, false, ErrInjectedFailure)
				if _, err := coordinator.Finalize(t.Context(), CoordinatorFinalizeRequest{OperationID: request.OperationID, EffectDigest: effectDigest}); err != ErrInjectedFailure {
					t.Fatalf("Finalize error = %v, want ErrInjectedFailure", err)
				}
			},
			wantStatus:        StatusCommitted,
			wantEffectCommits: 1,
			boundBeforeCrash:  true,
		},
		{
			name: "visibility activation commit before response",
			prepare: func(t *testing.T, coordinator *Coordinator, _ *DeterministicProvider, repository *coordinatorMemoryRepository, effects *coordinatorEffectResolver, request ReserveRequest, effectDigest contracts.Digest) {
				t.Helper()
				coordinatorReserveAndRecord(t, coordinator, repository, request)
				effectDigest = effects.commit(t, request, effectDigest, "0/20")
				repository.failTransaction(4, true, ErrResponseLost)
				if _, err := coordinator.Finalize(t.Context(), CoordinatorFinalizeRequest{OperationID: request.OperationID, EffectDigest: effectDigest}); err != ErrResponseLost {
					t.Fatalf("Finalize error = %v, want ErrResponseLost", err)
				}
			},
			wantStatus:        StatusCommitted,
			wantEffectCommits: 1,
			boundBeforeCrash:  true,
		},
	}

	for _, point := range points {
		point := point
		t.Run(point.name, func(t *testing.T) {
			t.Parallel()
			provider, err := NewDeterministicProvider(7)
			if err != nil {
				t.Fatal(err)
			}
			repository := newCoordinatorMemoryRepository([]DatabasePoint{
				{SystemID: 41, Timeline: 3, RequiredLSN: "0/30"},
				{SystemID: 41, Timeline: 3, RequiredLSN: "0/40"},
			})
			effects := newCoordinatorEffectResolver()
			clock := coordinatorClock{now: time.Date(2026, time.August, 23, 10, 0, 0, 0, time.UTC)}
			coordinator := mustNewCoordinatorForTest(t, provider, repository, effects, clock)
			operationID := uuid.MustParse("16d13f5b-ea19-4d83-a7a8-ad70ca725f00")
			request := ReserveRequest{
				OperationID: operationID,
				Kind:        EffectCertificateRevoke,
				ScopeKind:   ScopeNode,
				ScopeDigest: sha256.Sum256([]byte("coordinator-crash-node")),
			}
			effectDigest := sha256.Sum256([]byte("committed-domain-effect"))

			point.prepare(t, coordinator, provider, repository, effects, request, effectDigest)
			capturesBeforeRecovery := repository.captureCount()

			restarted := mustNewCoordinatorForTest(t, provider, repository, effects, clock)
			receipt, recoverErr := restarted.Recover(t.Context(), operationID)
			if point.wantStatus == StatusAborted {
				if recoverErr != ErrConflict {
					t.Fatalf("absent reserved recovery guessed reason: %v", recoverErr)
				}
				if point.name == "provider reserve committed before response" {
					if _, err := repository.Get(t.Context(), operationID); err != ErrNotFound {
						t.Fatal("missing claim-v1 fence was invented")
					}
					if provider.Snapshot()[0].TerminalReceipt != nil {
						t.Fatal("provider reservation was mutated")
					}
					return
				}
				receipt, recoverErr = restarted.Abort(t.Context(), AbortRequest{OperationID: operationID, Reason: AbortValidationFailed})
			}
			if recoverErr != nil {
				t.Fatalf("Recover error = %v", recoverErr)
			}
			if receipt.Validate() != nil || receipt.Status != point.wantStatus || receipt.Sequence != 1 {
				t.Fatalf("terminal receipt = %#v, want valid status %q at sequence 1", receipt, point.wantStatus)
			}
			if point.wantStatus == StatusCommitted {
				if receipt.DatabasePoint == nil || walPositionOrdinal(t, receipt.DatabasePoint.RequiredLSN) < walPositionOrdinal(t, "0/20") {
					t.Fatalf("required LSN = %#v, want not earlier than effect commit 0/20", receipt.DatabasePoint)
				}
			}
			if point.boundBeforeCrash && repository.captureCount() != capturesBeforeRecovery {
				t.Fatalf("recovery captured a later database point: before=%d after=%d", capturesBeforeRecovery, repository.captureCount())
			}

			records := provider.Snapshot()
			if len(records) != 1 || records[0].Sequence != 1 || records[0].TerminalReceipt == nil || records[0].TerminalReceipt.ReceiptDigest != receipt.ReceiptDigest {
				t.Fatalf("provider records = %#v, want one terminal sequence", records)
			}
			databaseRecord, getErr := repository.Get(t.Context(), operationID)
			if getErr != nil || databaseRecord.TerminalReceipt == nil || databaseRecord.TerminalReceipt.ReceiptDigest != receipt.ReceiptDigest {
				t.Fatalf("database record = %#v, %v; want exact terminal receipt", databaseRecord, getErr)
			}
			if effects.commitCount(operationID) != point.wantEffectCommits {
				t.Fatalf("effect commit count = %d, want %d", effects.commitCount(operationID), point.wantEffectCommits)
			}
		})
	}
}

func TestCoordinatorAbortRejectsNonAbsentEffectsWithoutMutation(t *testing.T) {
	t.Parallel()

	baseRequest := ReserveRequest{
		OperationID: uuid.MustParse("5f8821be-b9f1-4cef-ae96-a7cbba920101"),
		Kind:        EffectCertificateRevoke,
		ScopeKind:   ScopeNode,
		ScopeDigest: sha256.Sum256([]byte("abort-domain-safety-node")),
	}
	effectDigest := sha256.Sum256([]byte("abort-domain-safety-effect"))
	tests := []struct {
		name     string
		resolved ResolvedEffect
	}{
		{
			name: "prepared exact tuple",
			resolved: ResolvedEffect{
				Kind:         baseRequest.Kind,
				ScopeKind:    baseRequest.ScopeKind,
				ScopeDigest:  baseRequest.ScopeDigest,
				EffectDigest: effectDigest,
				State:        EffectPrepared,
			},
		},
		{
			name: "committed exact tuple",
			resolved: ResolvedEffect{
				Kind:         baseRequest.Kind,
				ScopeKind:    baseRequest.ScopeKind,
				ScopeDigest:  baseRequest.ScopeDigest,
				EffectDigest: effectDigest,
				State:        EffectCommitted,
			},
		},
		{
			name: "terminal exact tuple",
			resolved: ResolvedEffect{
				Kind:         baseRequest.Kind,
				ScopeKind:    baseRequest.ScopeKind,
				ScopeDigest:  baseRequest.ScopeDigest,
				EffectDigest: effectDigest,
				State:        EffectTerminal,
			},
		},
		{
			name: "prepared cross kind",
			resolved: ResolvedEffect{
				Kind:         EffectDesiredActivate,
				ScopeKind:    ScopeNode,
				ScopeDigest:  baseRequest.ScopeDigest,
				EffectDigest: effectDigest,
				State:        EffectPrepared,
			},
		},
		{
			name: "committed cross scope",
			resolved: ResolvedEffect{
				Kind:         EffectRootPublish,
				ScopeKind:    ScopeGlobalNodeTrust,
				ScopeDigest:  globalNodeScopeDigest,
				EffectDigest: effectDigest,
				State:        EffectCommitted,
			},
		},
		{
			name: "terminal cross scope digest",
			resolved: ResolvedEffect{
				Kind:         baseRequest.Kind,
				ScopeKind:    baseRequest.ScopeKind,
				ScopeDigest:  sha256.Sum256([]byte("abort-domain-safety-other-node")),
				EffectDigest: effectDigest,
				State:        EffectTerminal,
			},
		},
		{
			name: "committed cross effect digest",
			resolved: ResolvedEffect{
				Kind:         baseRequest.Kind,
				ScopeKind:    baseRequest.ScopeKind,
				ScopeDigest:  baseRequest.ScopeDigest,
				EffectDigest: sha256.Sum256([]byte("abort-domain-safety-other-effect")),
				State:        EffectCommitted,
			},
		},
	}

	for index, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			provider, err := NewDeterministicProvider(17)
			if err != nil {
				t.Fatal(err)
			}
			countingProvider := newCoordinatorCountingProvider(provider)
			repository := newCoordinatorMemoryRepository([]DatabasePoint{{SystemID: 41, Timeline: 3, RequiredLSN: "0/30"}})
			effects := newCoordinatorEffectResolver()
			request := baseRequest
			request.OperationID = uuid.MustParse(fmt.Sprintf("5f8821be-b9f1-4cef-ae96-a7cbba9201%02d", index+1))
			coordinator := mustNewCoordinatorForTest(t, countingProvider, repository, effects, coordinatorClock{now: time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC)})
			coordinatorReserveAndRecord(t, coordinator, repository, request)
			effects.set(request.OperationID, test.resolved)
			transactionsBefore := repository.transactionCount()

			if _, abortErr := coordinator.Abort(t.Context(), AbortRequest{OperationID: request.OperationID, Reason: AbortSuperseded}); abortErr != ErrConflict {
				t.Fatalf("Abort error = %v, want ErrConflict", abortErr)
			}
			if countingProvider.abortCount(request.OperationID) != 0 {
				t.Fatalf("provider Abort calls = %d, want 0", countingProvider.abortCount(request.OperationID))
			}
			if repository.transactionCount() != transactionsBefore+1 || repository.active != nil || repository.claims[request.OperationID] != nil {
				t.Fatal("rejected effect must roll back the locked inspection without retaining a claim")
			}
			record, getErr := repository.Get(t.Context(), request.OperationID)
			if getErr != nil || record.TerminalReceipt != nil {
				t.Fatalf("database record = %#v, %v; want pending", record, getErr)
			}
			providerRecord, inspectErr := provider.Inspect(t.Context(), request.OperationID)
			if inspectErr != nil || providerRecord.TerminalReceipt != nil {
				t.Fatalf("provider record = %#v, %v; want pending", providerRecord, inspectErr)
			}
		})
	}
}

func TestCoordinatorAbortResponseLossAndExactTerminalRetryAreIdempotent(t *testing.T) {
	t.Parallel()

	provider, err := NewDeterministicProvider(19)
	if err != nil {
		t.Fatal(err)
	}
	countingProvider := newCoordinatorCountingProvider(provider)
	repository := newCoordinatorMemoryRepository([]DatabasePoint{{SystemID: 41, Timeline: 3, RequiredLSN: "0/30"}})
	effects := newCoordinatorEffectResolver()
	request := ReserveRequest{
		OperationID: uuid.MustParse("619632c5-2004-4c5a-8426-0e6bf379c701"),
		Kind:        EffectCertificateRevoke,
		ScopeKind:   ScopeNode,
		ScopeDigest: sha256.Sum256([]byte("abort-response-loss-node")),
	}
	abortRequest := AbortRequest{OperationID: request.OperationID, Reason: AbortProviderDependencyFailed}
	first := mustNewCoordinatorForTest(t, countingProvider, repository, effects, coordinatorClock{now: time.Date(2026, 8, 24, 11, 0, 0, 0, time.UTC)})
	coordinatorReserveAndRecord(t, first, repository, request)
	if err := provider.FailNext(OperationAbort, FailureAfterMutationResponseLost); err != nil {
		t.Fatal(err)
	}
	if _, err := first.Abort(t.Context(), abortRequest); err != ErrResponseLost {
		t.Fatalf("first Abort error = %v, want ErrResponseLost", err)
	}
	if repository.transactionCount() != 1 || repository.claims[request.OperationID] == nil {
		t.Fatalf("claim must commit before provider response loss; transactions=%d", repository.transactionCount())
	}

	restarted := mustNewCoordinatorForTest(t, countingProvider, repository, effects, coordinatorClock{now: time.Date(2026, 8, 24, 11, 1, 0, 0, time.UTC)})
	receipt, err := restarted.Abort(t.Context(), abortRequest)
	if err != nil || receipt.Status != StatusAborted {
		t.Fatalf("response-loss retry = %#v, %v", receipt, err)
	}
	if countingProvider.abortCount(request.OperationID) != 1 || repository.transactionCount() != 3 {
		t.Fatalf("after recovery Abort=%d transactions=%d, want 1/3", countingProvider.abortCount(request.OperationID), repository.transactionCount())
	}
	transactionsBefore := repository.transactionCount()
	retry := mustNewCoordinatorForTest(t, countingProvider, repository, effects, coordinatorClock{now: time.Date(2026, 8, 24, 11, 2, 0, 0, time.UTC)})
	retried, err := retry.Abort(t.Context(), abortRequest)
	if err != nil || retried.ReceiptDigest != receipt.ReceiptDigest || retried.Validate() != nil {
		t.Fatalf("exact terminal retry = %#v, %v; want %#v", retried, err, receipt)
	}
	if countingProvider.abortCount(request.OperationID) != 1 || repository.transactionCount() != transactionsBefore+1 {
		t.Fatalf("terminal retry must do one locked read and zero provider mutations: Abort=%d transactions=%d", countingProvider.abortCount(request.OperationID), repository.transactionCount())
	}
}

func TestCoordinatorFinalizeAbortRaceRejectsAbortAfterCommittedEffect(t *testing.T) {
	t.Parallel()

	provider, err := NewDeterministicProvider(23)
	if err != nil {
		t.Fatal(err)
	}
	countingProvider := newCoordinatorCountingProvider(provider)
	repository := newCoordinatorMemoryRepository([]DatabasePoint{{SystemID: 41, Timeline: 3, RequiredLSN: "0/40"}})
	effects := newCoordinatorEffectResolver()
	request := ReserveRequest{
		OperationID: uuid.MustParse("4edbc7ca-8918-46ca-9e89-207a638a7d01"),
		Kind:        EffectCertificateRevoke,
		ScopeKind:   ScopeNode,
		ScopeDigest: sha256.Sum256([]byte("finalize-abort-race-node")),
	}
	effectDigest := sha256.Sum256([]byte("finalize-abort-race-effect"))
	coordinatorReserveAndRecord(t, mustNewCoordinatorForTest(t, countingProvider, repository, effects, coordinatorClock{now: time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)}), repository, request)
	effectDigest = effects.commit(t, request, effectDigest, "0/20")
	entered := make(chan struct{})
	release := make(chan struct{})
	blockingResolver := &coordinatorFirstResolveBlockingResolver{EffectResolver: effects, entered: entered, release: release}
	coordinator := mustNewCoordinatorForTest(t, countingProvider, repository, blockingResolver, coordinatorClock{now: time.Date(2026, 8, 24, 12, 1, 0, 0, time.UTC)})
	type finalizeResult struct {
		receipt Receipt
		err     error
	}
	finalized := make(chan finalizeResult, 1)
	go func() {
		receipt, finalizeErr := coordinator.Finalize(context.Background(), CoordinatorFinalizeRequest{OperationID: request.OperationID, EffectDigest: effectDigest})
		finalized <- finalizeResult{receipt: receipt, err: finalizeErr}
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("Finalize did not reach the bounded resolver barrier")
	}

	aborted := make(chan error, 1)
	go func() {
		_, err := coordinator.Abort(context.Background(), AbortRequest{OperationID: request.OperationID, Reason: AbortSuperseded})
		aborted <- err
	}()
	close(release)
	if abortErr := <-aborted; abortErr != ErrConflict {
		t.Fatalf("concurrent Abort error = %v, want ErrConflict", abortErr)
	}
	if countingProvider.abortCount(request.OperationID) != 0 {
		t.Fatal("Abort mutated provider after committed effect")
	}
	select {
	case result := <-finalized:
		if result.err != nil || result.receipt.Status != StatusCommitted {
			t.Fatalf("Finalize result = %#v, %v; want committed", result.receipt, result.err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Finalize did not complete after releasing resolver barrier")
	}
}

func TestCoordinatorRecoverRejectsUnexplainedAbortedTerminalEffects(t *testing.T) {
	t.Parallel()

	baseRequest := ReserveRequest{
		Kind:        EffectCertificateRevoke,
		ScopeKind:   ScopeNode,
		ScopeDigest: sha256.Sum256([]byte("recover-terminal-node")),
	}
	effectDigest := sha256.Sum256([]byte("recover-terminal-effect"))
	tests := []struct {
		name     string
		resolved ResolvedEffect
	}{
		{name: "exact terminal tuple", resolved: ResolvedEffect{Kind: baseRequest.Kind, ScopeKind: baseRequest.ScopeKind, ScopeDigest: baseRequest.ScopeDigest, EffectDigest: effectDigest, State: EffectTerminal}},
		{name: "cross kind", resolved: ResolvedEffect{Kind: EffectDesiredActivate, ScopeKind: ScopeNode, ScopeDigest: baseRequest.ScopeDigest, EffectDigest: effectDigest, State: EffectTerminal}},
		{name: "cross scope", resolved: ResolvedEffect{Kind: EffectRootPublish, ScopeKind: ScopeGlobalNodeTrust, ScopeDigest: globalNodeScopeDigest, EffectDigest: effectDigest, State: EffectTerminal}},
		{name: "cross scope digest", resolved: ResolvedEffect{Kind: baseRequest.Kind, ScopeKind: baseRequest.ScopeKind, ScopeDigest: sha256.Sum256([]byte("recover-terminal-other-node")), EffectDigest: effectDigest, State: EffectTerminal}},
		{name: "cross effect digest", resolved: ResolvedEffect{Kind: baseRequest.Kind, ScopeKind: baseRequest.ScopeKind, ScopeDigest: baseRequest.ScopeDigest, EffectDigest: sha256.Sum256([]byte("recover-terminal-other-effect")), State: EffectTerminal}},
	}

	for index, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			provider, err := NewDeterministicProvider(29)
			if err != nil {
				t.Fatal(err)
			}
			repository := newCoordinatorMemoryRepository([]DatabasePoint{{SystemID: 41, Timeline: 3, RequiredLSN: "0/30"}})
			effects := newCoordinatorEffectResolver()
			request := baseRequest
			request.OperationID = uuid.MustParse(fmt.Sprintf("9962a8ce-e1bd-4c5f-a477-d92a67f872%02d", index+1))
			coordinator := mustNewCoordinatorForTest(t, provider, repository, effects, coordinatorClock{now: time.Date(2026, 8, 24, 13, 0, 0, 0, time.UTC)})
			coordinatorReserveAndRecord(t, coordinator, repository, request)
			if _, err := provider.Abort(t.Context(), AbortRequest{OperationID: request.OperationID, Reason: AbortValidationFailed}); err != nil {
				t.Fatal(err)
			}
			effects.set(request.OperationID, test.resolved)
			transactionsBefore := repository.transactionCount()

			if _, err := coordinator.Recover(t.Context(), request.OperationID); err != ErrConflict {
				t.Fatalf("Recover error = %v, want ErrConflict", err)
			}
			if repository.transactionCount() != transactionsBefore+1 || repository.active != nil || repository.claims[request.OperationID] != nil {
				t.Fatal("rejected effect must roll back the locked inspection without retaining a claim")
			}
			record, getErr := repository.Get(t.Context(), request.OperationID)
			if getErr != nil || record.TerminalReceipt != nil {
				t.Fatalf("database record = %#v, %v; unexplained terminal escaped pending visibility", record, getErr)
			}
			head, headErr := repository.Head(t.Context())
			if headErr != nil || head.PendingCount != 1 {
				t.Fatalf("database head = %#v, %v; want one pending", head, headErr)
			}
		})
	}
}

func TestCoordinatorRecoverCopiesProviderAbortForExactAbsentUnboundEffect(t *testing.T) {
	t.Parallel()

	provider, err := NewDeterministicProvider(31)
	if err != nil {
		t.Fatal(err)
	}
	repository := newCoordinatorMemoryRepository([]DatabasePoint{{SystemID: 41, Timeline: 3, RequiredLSN: "0/30"}})
	effects := newCoordinatorEffectResolver()
	request := ReserveRequest{
		OperationID: uuid.MustParse("e98616b0-401a-4583-94ba-462891b39a01"),
		Kind:        EffectCertificateRevoke,
		ScopeKind:   ScopeNode,
		ScopeDigest: sha256.Sum256([]byte("recover-absent-abort-node")),
	}
	coordinator := mustNewCoordinatorForTest(t, provider, repository, effects, coordinatorClock{now: time.Date(2026, 8, 24, 14, 0, 0, 0, time.UTC)})
	coordinatorReserveAndRecord(t, coordinator, repository, request)
	providerReceipt, err := provider.Abort(t.Context(), AbortRequest{OperationID: request.OperationID, Reason: AbortValidationFailed})
	if err != nil {
		t.Fatal(err)
	}
	recovered, err := coordinator.Recover(t.Context(), request.OperationID)
	if err != nil || recovered.ReceiptDigest != providerReceipt.ReceiptDigest || recovered.Validate() != nil {
		t.Fatalf("Recover = %#v, %v; want exact provider receipt %#v", recovered, err, providerReceipt)
	}
}

func TestAuthorityReadiness(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mutate     func(*testing.T, *DeterministicProvider, *coordinatorMemoryRepository, *coordinatorEffectResolver)
		wantReady  bool
		wantReason string
	}{
		{name: "exact anchors", wantReady: true, wantReason: "ready"},
		{
			name: "provider unavailable",
			mutate: func(t *testing.T, provider *DeterministicProvider, _ *coordinatorMemoryRepository, _ *coordinatorEffectResolver) {
				if err := provider.FailNext(OperationHead, FailureBeforeMutation); err != nil {
					t.Fatal(err)
				}
			},
			wantReason: "provider_unavailable",
		},
		{
			name: "database unavailable",
			mutate: func(_ *testing.T, _ *DeterministicProvider, repository *coordinatorMemoryRepository, _ *coordinatorEffectResolver) {
				repository.setHeadError(ErrInjectedFailure)
			},
			wantReason: "database_unavailable",
		},
		{
			name: "same database identity restored behind provider",
			mutate: func(_ *testing.T, _ *DeterministicProvider, repository *coordinatorMemoryRepository, _ *coordinatorEffectResolver) {
				repository.clearRecords()
				repository.setCurrentPoint(DatabasePoint{SystemID: 41, Timeline: 3, RequiredLSN: "0/20"})
			},
			wantReason: "database_behind_provider",
		},
		{
			name: "different database identity wins precedence",
			mutate: func(_ *testing.T, _ *DeterministicProvider, repository *coordinatorMemoryRepository, _ *coordinatorEffectResolver) {
				repository.clearRecords()
				repository.setCurrentPoint(DatabasePoint{SystemID: 99, Timeline: 1, RequiredLSN: "0/1"})
			},
			wantReason: "database_identity_mismatch",
		},
		{
			name: "lower timeline",
			mutate: func(_ *testing.T, _ *DeterministicProvider, repository *coordinatorMemoryRepository, _ *coordinatorEffectResolver) {
				repository.setCurrentPoint(DatabasePoint{SystemID: 41, Timeline: 2, RequiredLSN: "0/40"})
			},
			wantReason: "database_timeline_behind",
		},
		{
			name: "WAL older than committed receipt",
			mutate: func(_ *testing.T, _ *DeterministicProvider, repository *coordinatorMemoryRepository, _ *coordinatorEffectResolver) {
				repository.setCurrentPoint(DatabasePoint{SystemID: 41, Timeline: 3, RequiredLSN: "0/20"})
			},
			wantReason: "database_wal_behind",
		},
		{
			name: "same coordinate reservation fork",
			mutate: func(_ *testing.T, _ *DeterministicProvider, repository *coordinatorMemoryRepository, _ *coordinatorEffectResolver) {
				head, err := repository.Head(context.Background())
				if err != nil {
					panic(err)
				}
				head.LatestReservationDigest = sha256.Sum256([]byte("different reservation"))
				repository.setHeadOverride(head)
			},
			wantReason: "reservation_mismatch",
		},
		{
			name: "same committed coordinate receipt fork",
			mutate: func(_ *testing.T, _ *DeterministicProvider, repository *coordinatorMemoryRepository, _ *coordinatorEffectResolver) {
				head, err := repository.Head(context.Background())
				if err != nil {
					panic(err)
				}
				head.LatestCommittedReceiptDigest = sha256.Sum256([]byte("different receipt"))
				repository.setHeadOverride(head)
			},
			wantReason: "committed_mismatch",
		},
		{
			name: "record count gap",
			mutate: func(_ *testing.T, _ *DeterministicProvider, repository *coordinatorMemoryRepository, _ *coordinatorEffectResolver) {
				head, err := repository.Head(context.Background())
				if err != nil {
					panic(err)
				}
				head.RecordCount = 1
				head.LatestReservedSequence = 2
				head.HasSequenceGap = true
				repository.setHeadOverride(head)
			},
			wantReason: "database_sequence_gap",
		},
		{
			name: "database higher unexplained sequence",
			mutate: func(_ *testing.T, _ *DeterministicProvider, repository *coordinatorMemoryRepository, _ *coordinatorEffectResolver) {
				head, err := repository.Head(context.Background())
				if err != nil {
					panic(err)
				}
				head.RecordCount = 2
				head.LatestReservedSequence = 2
				head.LatestReservationDigest = sha256.Sum256([]byte("unexplained later reservation"))
				repository.setHeadOverride(head)
			},
			wantReason: "database_ahead_provider",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			coordinator, provider, repository, effects := newReadyCoordinatorFixture(t)
			if test.mutate != nil {
				test.mutate(t, provider, repository, effects)
			}
			readiness, err := coordinator.CheckReady(t.Context())
			if err != nil {
				t.Fatalf("CheckReady error = %v", err)
			}
			if readiness.Ready != test.wantReady || string(readiness.Reason) != test.wantReason {
				t.Fatalf("readiness = %#v, want ready=%v reason=%q", readiness, test.wantReady, test.wantReason)
			}
		})
	}
}

func TestAuthorityReadinessInspectsEveryPendingReservation(t *testing.T) {
	t.Parallel()
	provider, err := NewDeterministicProvider(7)
	if err != nil {
		t.Fatal(err)
	}
	countingProvider := &coordinatorCountingProvider{Provider: provider, inspections: make(map[uuid.UUID]int)}
	repository := newCoordinatorMemoryRepository([]DatabasePoint{{SystemID: 41, Timeline: 3, RequiredLSN: "0/30"}})
	effects := newCoordinatorEffectResolver()
	coordinator := mustNewCoordinatorForTest(t, countingProvider, repository, effects, coordinatorClock{now: time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC)})

	operationIDs := []uuid.UUID{
		uuid.MustParse("b19c8f55-3432-44a6-802d-f68001217f01"),
		uuid.MustParse("b19c8f55-3432-44a6-802d-f68001217f02"),
	}
	for index, operationID := range operationIDs {
		request := ReserveRequest{
			OperationID: operationID,
			Kind:        EffectCertificateRevoke,
			ScopeKind:   ScopeNode,
			ScopeDigest: sha256.Sum256([]byte(fmt.Sprintf("pending-node-%d", index))),
		}
		reservation, reserveErr := coordinator.Reserve(t.Context(), request)
		if reserveErr != nil {
			t.Fatal(reserveErr)
		}
		if recordErr := repository.RecordPending(t.Context(), coordinatorNoopDBTX{}, reservation, time.Date(2026, 8, 23, 10, 0, index, 0, time.UTC)); recordErr != nil {
			t.Fatal(recordErr)
		}
		effects.prepare(request, sha256.Sum256([]byte(fmt.Sprintf("pending-effect-%d", index))))
	}

	readiness, err := coordinator.CheckReady(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if readiness.Ready || string(readiness.Reason) != "pending_unresolved" {
		t.Fatalf("readiness = %#v, want pending_unresolved", readiness)
	}
	for _, operationID := range operationIDs {
		if countingProvider.inspectCount(operationID) != 1 || effects.resolveCount(operationID) != 1 {
			t.Fatalf("operation %s calls: Inspect=%d Resolve=%d, want 1 each", operationID, countingProvider.inspectCount(operationID), effects.resolveCount(operationID))
		}
	}
}

func TestAuthorityReadinessNormalizesOnlyExactEmptyHeads(t *testing.T) {
	t.Parallel()

	t.Run("exact empty pair copies the positive provider epoch into the returned database head", func(t *testing.T) {
		t.Parallel()
		provider, err := NewDeterministicProvider(37)
		if err != nil {
			t.Fatal(err)
		}
		repository := newCoordinatorMemoryRepository([]DatabasePoint{{SystemID: 41, Timeline: 3, RequiredLSN: "0/30"}})
		coordinator := mustNewCoordinatorForTest(t, provider, repository, newCoordinatorEffectResolver(), coordinatorClock{now: time.Date(2026, 8, 24, 15, 0, 0, 0, time.UTC)})

		readiness, err := coordinator.CheckReady(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		if !readiness.Ready || readiness.Reason != ReadinessReady || readiness.ProviderHead.Epoch != 37 || readiness.DatabaseHead.Epoch != 37 {
			t.Fatalf("readiness = %#v, want exact normalized empty epoch 37", readiness)
		}
		if readiness.DatabaseHead.RecordCount != 0 || readiness.DatabaseHead.LatestReservedSequence != 0 ||
			readiness.DatabaseHead.LatestCommittedSequence != 0 || readiness.DatabaseHead.PendingCount != 0 ||
			readiness.DatabaseHead.LatestReservationDigest != (contracts.Digest{}) ||
			readiness.DatabaseHead.LatestCommittedOperationID != uuid.Nil ||
			readiness.DatabaseHead.LatestCommittedReceiptDigest != (contracts.Digest{}) ||
			readiness.DatabaseHead.LatestCommittedDatabasePoint != nil {
			t.Fatalf("normalized database head fabricated non-epoch data: %#v", readiness.DatabaseHead)
		}

		readiness.DatabaseHead.Epoch = 0
		again, err := coordinator.CheckReady(t.Context())
		if err != nil || again.DatabaseHead.Epoch != 37 {
			t.Fatalf("caller mutation changed subsequent normalized head: %#v, %v", again, err)
		}
	})

	t.Run("partial provider empty is rejected before normalization", func(t *testing.T) {
		t.Parallel()
		provider, err := NewDeterministicProvider(37)
		if err != nil {
			t.Fatal(err)
		}
		wrapped := newCoordinatorCountingProvider(provider)
		wrapped.setHeadOverride(Head{Epoch: 37, LatestReservationDigest: sha256.Sum256([]byte("partial-provider-empty"))})
		repository := newCoordinatorMemoryRepository([]DatabasePoint{{SystemID: 41, Timeline: 3, RequiredLSN: "0/30"}})
		coordinator := mustNewCoordinatorForTest(t, wrapped, repository, newCoordinatorEffectResolver(), coordinatorClock{now: time.Date(2026, 8, 24, 15, 0, 0, 0, time.UTC)})

		readiness, err := coordinator.CheckReady(t.Context())
		if err != nil || readiness.Ready || readiness.Reason != ReadinessProviderUnavailable || readiness.DatabaseHead.Epoch != 0 {
			t.Fatalf("readiness = %#v, %v; want provider_unavailable without normalization", readiness, err)
		}
	})

	t.Run("partial database empty is rejected before normalization", func(t *testing.T) {
		t.Parallel()
		provider, err := NewDeterministicProvider(37)
		if err != nil {
			t.Fatal(err)
		}
		repository := newCoordinatorMemoryRepository([]DatabasePoint{{SystemID: 41, Timeline: 3, RequiredLSN: "0/30"}})
		repository.setHeadOverride(DatabaseHead{RecordCount: 1})
		coordinator := mustNewCoordinatorForTest(t, provider, repository, newCoordinatorEffectResolver(), coordinatorClock{now: time.Date(2026, 8, 24, 15, 0, 0, 0, time.UTC)})

		readiness, err := coordinator.CheckReady(t.Context())
		if err != nil || readiness.Ready || readiness.Reason != ReadinessDatabaseUnavailable || readiness.DatabaseHead.Epoch != 0 {
			t.Fatalf("readiness = %#v, %v; want database_unavailable without normalization", readiness, err)
		}
	})
}

func TestAuthorityReadinessClassifiesPendingInspectErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		inspectErr error
		wantReason ReadinessReason
		wantError  error
	}{
		{name: "not found is a mismatch", inspectErr: ErrNotFound, wantReason: ReadinessPendingMismatch},
		{name: "conflict is a mismatch", inspectErr: ErrConflict, wantReason: ReadinessPendingMismatch},
		{name: "terminal conflict is a mismatch", inspectErr: ErrTerminalConflict, wantReason: ReadinessPendingMismatch},
		{name: "dependency failure is unavailable", inspectErr: ErrInjectedFailure, wantReason: ReadinessProviderUnavailable},
		{name: "context cancellation is exact", inspectErr: context.Canceled, wantError: ErrCanceled},
		{name: "authority cancellation is exact", inspectErr: ErrCanceled, wantError: ErrCanceled},
	}

	for index, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			provider, err := NewDeterministicProvider(41)
			if err != nil {
				t.Fatal(err)
			}
			wrapped := newCoordinatorCountingProvider(provider)
			repository := newCoordinatorMemoryRepository([]DatabasePoint{{SystemID: 41, Timeline: 3, RequiredLSN: "0/30"}})
			effects := newCoordinatorEffectResolver()
			request := ReserveRequest{
				OperationID: uuid.MustParse(fmt.Sprintf("b4bef512-ff22-47da-b49e-733cc8d4a1%02d", index+1)),
				Kind:        EffectCertificateRevoke,
				ScopeKind:   ScopeNode,
				ScopeDigest: sha256.Sum256([]byte(fmt.Sprintf("pending-inspect-node-%d", index))),
			}
			coordinator := mustNewCoordinatorForTest(t, wrapped, repository, effects, coordinatorClock{now: time.Date(2026, 8, 24, 16, 0, 0, 0, time.UTC)})
			coordinatorReserveAndRecord(t, coordinator, repository, request)
			effects.prepare(request, sha256.Sum256([]byte(fmt.Sprintf("pending-inspect-effect-%d", index))))
			wrapped.setInspectError(request.OperationID, test.inspectErr)

			readiness, checkErr := coordinator.CheckReady(t.Context())
			if test.wantError != nil {
				if checkErr != test.wantError || readiness != (Readiness{}) {
					t.Fatalf("CheckReady = %#v, %v; want zero/%v", readiness, checkErr, test.wantError)
				}
				return
			}
			if checkErr != nil || readiness.Ready || readiness.Reason != test.wantReason {
				t.Fatalf("CheckReady = %#v, %v; want %q", readiness, checkErr, test.wantReason)
			}
		})
	}
}

func TestAuthorityReadinessPendingReasonPriorityAcrossAllRows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		includeUnavailable bool
		wantReason         ReadinessReason
	}{
		{name: "effect unavailable beats mismatch and unresolved", wantReason: ReadinessEffectUnavailable},
		{name: "provider unavailable beats every other pending reason", includeUnavailable: true, wantReason: ReadinessProviderUnavailable},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			provider, err := NewDeterministicProvider(43)
			if err != nil {
				t.Fatal(err)
			}
			wrapped := newCoordinatorCountingProvider(provider)
			repository := newCoordinatorMemoryRepository([]DatabasePoint{{SystemID: 41, Timeline: 3, RequiredLSN: "0/30"}})
			effects := newCoordinatorEffectResolver()
			coordinator := mustNewCoordinatorForTest(t, wrapped, repository, effects, coordinatorClock{now: time.Date(2026, 8, 24, 17, 0, 0, 0, time.UTC)})
			operationIDs := []uuid.UUID{
				uuid.MustParse("45ba0535-6209-4e90-a4a2-17f8bc7f4401"),
				uuid.MustParse("45ba0535-6209-4e90-a4a2-17f8bc7f4402"),
				uuid.MustParse("45ba0535-6209-4e90-a4a2-17f8bc7f4403"),
				uuid.MustParse("45ba0535-6209-4e90-a4a2-17f8bc7f4404"),
			}
			for index, operationID := range operationIDs {
				request := ReserveRequest{
					OperationID: operationID,
					Kind:        EffectCertificateRevoke,
					ScopeKind:   ScopeNode,
					ScopeDigest: sha256.Sum256([]byte(fmt.Sprintf("pending-priority-node-%d", index))),
				}
				coordinatorReserveAndRecord(t, coordinator, repository, request)
				effects.prepare(request, sha256.Sum256([]byte(fmt.Sprintf("pending-priority-effect-%d", index))))
			}
			wrapped.setInspectError(operationIDs[0], ErrNotFound)
			effects.setError(operationIDs[1], ErrInjectedFailure)
			if test.includeUnavailable {
				wrapped.setInspectError(operationIDs[2], ErrInjectedFailure)
			}

			readiness, checkErr := coordinator.CheckReady(t.Context())
			if checkErr != nil || readiness.Ready || readiness.Reason != test.wantReason {
				t.Fatalf("CheckReady = %#v, %v; want %q", readiness, checkErr, test.wantReason)
			}
			for _, operationID := range operationIDs {
				if wrapped.inspectCount(operationID) != 1 || effects.resolveCount(operationID) != 1 {
					t.Fatalf("operation %s calls Inspect=%d Resolve=%d, want 1/1", operationID, wrapped.inspectCount(operationID), effects.resolveCount(operationID))
				}
			}
		})
	}
}

func newReadyCoordinatorFixture(t *testing.T) (*Coordinator, *DeterministicProvider, *coordinatorMemoryRepository, *coordinatorEffectResolver) {
	t.Helper()
	provider, err := NewDeterministicProvider(7)
	if err != nil {
		t.Fatal(err)
	}
	repository := newCoordinatorMemoryRepository([]DatabasePoint{
		{SystemID: 41, Timeline: 3, RequiredLSN: "0/30"},
		{SystemID: 41, Timeline: 3, RequiredLSN: "0/40"},
	})
	effects := newCoordinatorEffectResolver()
	coordinator := mustNewCoordinatorForTest(t, provider, repository, effects, coordinatorClock{now: time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC)})
	request := ReserveRequest{
		OperationID: uuid.MustParse("8f472e51-8135-4288-b7d1-4e95ba4bd210"),
		Kind:        EffectCertificateRevoke,
		ScopeKind:   ScopeNode,
		ScopeDigest: sha256.Sum256([]byte("ready-node")),
	}
	coordinatorReserveAndRecord(t, coordinator, repository, request)
	effectDigest := sha256.Sum256([]byte("ready-effect"))
	effectDigest = effects.commit(t, request, effectDigest, "0/20")
	if _, err := coordinator.Finalize(t.Context(), CoordinatorFinalizeRequest{OperationID: request.OperationID, EffectDigest: effectDigest}); err != nil {
		t.Fatal(err)
	}
	return coordinator, provider, repository, effects
}

func coordinatorReserveAndRecord(t *testing.T, coordinator *Coordinator, repository *coordinatorMemoryRepository, request ReserveRequest) Reservation {
	t.Helper()
	reservation, err := coordinator.Reserve(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.RecordPending(t.Context(), coordinatorNoopDBTX{}, reservation, time.Date(2026, time.August, 23, 10, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	return reservation
}

func mustNewCoordinatorForTest(t *testing.T, provider Provider, repository Repository, effects EffectResolver, clock securitykit.Clock) *Coordinator {
	t.Helper()
	handler := &coordinatorTestHandler{source: effects, repository: repository}
	if memory, ok := effects.(*coordinatorEffectResolver); ok {
		memory.repository = repository
	}
	registrations := make([]EffectRegistration, 0, 15)
	for _, kind := range dispatcherTestSupportedKinds {
		registrations = append(registrations, EffectRegistration{Kind: kind, Resolver: handler, Activator: handler})
	}
	registrations = append(registrations, EffectRegistration{Kind: EffectTrustBundlePublish}, EffectRegistration{Kind: EffectOperatorAuthorizerChange})
	dispatcher, err := NewEffectDispatcher(registrations)
	if err != nil {
		t.Fatal(err)
	}
	coordinator, err := NewCoordinator(provider, repository, dispatcher)
	if err != nil {
		t.Fatal(err)
	}
	coordinator.clock = clock
	return coordinator
}

type coordinatorClock struct{ now time.Time }

func (clock coordinatorClock) Now() time.Time { return clock.now }

type coordinatorEffectResolver struct {
	mu             sync.Mutex
	effects        map[uuid.UUID]ResolvedEffect
	errors         map[uuid.UUID]error
	commits        map[uuid.UUID]int
	lsns           map[uuid.UUID]WALPosition
	resolves       map[uuid.UUID]int
	repository     Repository
	materials      map[uuid.UUID]ActivationDecisionMaterial
	captures       int
	activations    int
	validations    int
	captureErr     error
	activationErr  error
	validationErr  error
	conditional    bool
	activationHook func(ValidatedActivationDecisionEvidence)
}

func newCoordinatorEffectResolver() *coordinatorEffectResolver {
	return &coordinatorEffectResolver{
		effects:   make(map[uuid.UUID]ResolvedEffect),
		errors:    make(map[uuid.UUID]error),
		commits:   make(map[uuid.UUID]int),
		lsns:      make(map[uuid.UUID]WALPosition),
		resolves:  make(map[uuid.UUID]int),
		materials: make(map[uuid.UUID]ActivationDecisionMaterial),
	}
}

func (resolver *coordinatorEffectResolver) ResolveAuthorityEffect(ctx context.Context, operationID uuid.UUID) (ResolvedEffect, error) {
	if ctx == nil || ctx.Err() != nil {
		return ResolvedEffect{}, ErrCanceled
	}
	resolver.mu.Lock()
	defer resolver.mu.Unlock()
	resolver.resolves[operationID]++
	if err := resolver.errors[operationID]; err != nil {
		return ResolvedEffect{}, err
	}
	resolved, ok := resolver.effects[operationID]
	if !ok {
		return ResolvedEffect{State: EffectAbsent}, nil
	}
	return resolved, nil
}

func (resolver *coordinatorEffectResolver) set(operationID uuid.UUID, resolved ResolvedEffect) {
	resolver.mu.Lock()
	defer resolver.mu.Unlock()
	resolver.effects[operationID] = resolved
}

func (resolver *coordinatorEffectResolver) setError(operationID uuid.UUID, err error) {
	resolver.mu.Lock()
	defer resolver.mu.Unlock()
	resolver.errors[operationID] = err
}

func (resolver *coordinatorEffectResolver) prepare(request ReserveRequest, digest contracts.Digest) {
	resolver.mu.Lock()
	defer resolver.mu.Unlock()
	resolver.effects[request.OperationID] = ResolvedEffect{
		Kind:         request.Kind,
		ScopeKind:    request.ScopeKind,
		ScopeDigest:  request.ScopeDigest,
		EffectDigest: digest,
		State:        EffectPrepared,
	}
}

func (resolver *coordinatorEffectResolver) commit(t *testing.T, request ReserveRequest, digest contracts.Digest, lsn WALPosition) contracts.Digest {
	t.Helper()
	if walPositionOrdinal(t, lsn) == 0 {
		t.Fatal("effect commit LSN must be positive")
	}
	record, err := resolver.repository.Get(t.Context(), request.OperationID)
	if err != nil {
		t.Fatal(err)
	}
	input := AuthorityEffectCommitmentInput{OperationID: request.OperationID, Kind: request.Kind, ScopeKind: request.ScopeKind,
		ScopeDigest: request.ScopeDigest, Epoch: record.Epoch, Sequence: record.Sequence, BaseEffectDigest: digest,
		Mode: CommitmentFinalNotApplied, Reason: EffectReasonFailed}
	if resolver.conditional {
		input.Mode, input.Reason, input.ActivationPolicyVersion = CommitmentConditionalApply, EffectReasonNone, 1
		input.ActivationInputsDigest = sha256.Sum256([]byte("task9-typed-activation-input:" + request.OperationID.String()))
	}
	commitment, err := NewAuthorityEffectCommitment(input)
	if err != nil {
		t.Fatal(err)
	}
	digest = commitment.Digest()
	resolver.mu.Lock()
	defer resolver.mu.Unlock()
	resolver.commits[request.OperationID]++
	resolver.effects[request.OperationID] = ResolvedEffect{
		Kind:         request.Kind,
		ScopeKind:    request.ScopeKind,
		ScopeDigest:  request.ScopeDigest,
		EffectDigest: digest,
		State:        EffectCommitted,
	}
	resolver.lsns[request.OperationID] = lsn
	material := ActivationDecisionMaterial{Commitment: commitment, Reason: EffectReasonFailed, CheckpointKind: CheckpointNone,
		TrustedTimeKind: TrustedTimeNone, Capability: DecisionCapabilityNotAppliedOnly}
	if resolver.conditional {
		now := time.Now().UTC()
		identity := sha256.Sum256([]byte("task9-authenticated-time-provider"))
		material.Reason, material.TrustedTimeKind, material.Capability = EffectReasonNone, TrustedTimeRollbackResistant, DecisionCapabilityMayApply
		material.TrustedInstant, material.EvidenceValidUntil = now, now.Add(4*time.Second)
		material.AttestationExpiresAt, material.ActivationDeadline = now.Add(10*time.Second), now.Add(10*time.Second)
		material.ProviderIdentityDigest, material.ExpectedProviderIdentityDigest = identity, identity
		material.FloorAttestationDigest = sha256.Sum256([]byte("task9-authenticated-time-floor"))
	}
	resolver.materials[request.OperationID] = material
	return digest
}

func (resolver *coordinatorEffectResolver) commitCount(operationID uuid.UUID) int {
	resolver.mu.Lock()
	defer resolver.mu.Unlock()
	return resolver.commits[operationID]
}

func (resolver *coordinatorEffectResolver) resolveCount(operationID uuid.UUID) int {
	resolver.mu.Lock()
	defer resolver.mu.Unlock()
	return resolver.resolves[operationID]
}

type coordinatorCountingProvider struct {
	Provider
	mu            sync.Mutex
	inspections   map[uuid.UUID]int
	aborts        map[uuid.UUID]int
	inspectErrors map[uuid.UUID]error
	headOverride  *Head
}

func newCoordinatorCountingProvider(provider Provider) *coordinatorCountingProvider {
	return &coordinatorCountingProvider{
		Provider:      provider,
		inspections:   make(map[uuid.UUID]int),
		aborts:        make(map[uuid.UUID]int),
		inspectErrors: make(map[uuid.UUID]error),
	}
}

func (provider *coordinatorCountingProvider) Inspect(ctx context.Context, operationID uuid.UUID) (Record, error) {
	provider.mu.Lock()
	provider.inspections[operationID]++
	err := provider.inspectErrors[operationID]
	provider.mu.Unlock()
	if err != nil {
		return Record{}, err
	}
	return provider.Provider.Inspect(ctx, operationID)
}

func (provider *coordinatorCountingProvider) Abort(ctx context.Context, request AbortRequest) (Receipt, error) {
	provider.mu.Lock()
	provider.aborts[request.OperationID]++
	provider.mu.Unlock()
	return provider.Provider.Abort(ctx, request)
}

func (provider *coordinatorCountingProvider) Head(ctx context.Context) (Head, error) {
	provider.mu.Lock()
	override := provider.headOverride
	provider.mu.Unlock()
	if override != nil {
		return cloneHead(*override), nil
	}
	return provider.Provider.Head(ctx)
}

func (provider *coordinatorCountingProvider) inspectCount(operationID uuid.UUID) int {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	return provider.inspections[operationID]
}

func (provider *coordinatorCountingProvider) abortCount(operationID uuid.UUID) int {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	return provider.aborts[operationID]
}

func (provider *coordinatorCountingProvider) setInspectError(operationID uuid.UUID, err error) {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	provider.inspectErrors[operationID] = err
}

func (provider *coordinatorCountingProvider) setHeadOverride(head Head) {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	cloned := cloneHead(head)
	provider.headOverride = &cloned
}

type coordinatorFirstResolveBlockingResolver struct {
	EffectResolver
	once    sync.Once
	entered chan<- struct{}
	release <-chan struct{}
}

func (resolver *coordinatorFirstResolveBlockingResolver) ResolveAuthorityEffect(ctx context.Context, operationID uuid.UUID) (ResolvedEffect, error) {
	blocked := false
	resolver.once.Do(func() {
		blocked = true
		close(resolver.entered)
	})
	if blocked {
		select {
		case <-ctx.Done():
			return ResolvedEffect{}, ErrCanceled
		case <-resolver.release:
		}
	}
	return resolver.EffectResolver.ResolveAuthorityEffect(ctx, operationID)
}

type coordinatorTransactionFailure struct {
	call  int
	after bool
	err   error
}

type coordinatorMemoryRepository struct {
	txMu            sync.Mutex
	mu              sync.Mutex
	records         map[uuid.UUID]Record
	reservedAt      map[uuid.UUID]time.Time
	boundAt         map[uuid.UUID]time.Time
	terminalAt      map[uuid.UUID]time.Time
	points          []DatabasePoint
	captures        int
	transactions    int
	transactionFail *coordinatorTransactionFailure
	headOverride    *DatabaseHead
	headErr         error
	claims          map[uuid.UUID]*AbortClaim
	outcomes        map[uuid.UUID]*PersistedAuthorityEffectOutcome
	active          *coordinatorMemoryTransaction
	events          []string
}

func newCoordinatorMemoryRepository(points []DatabasePoint) *coordinatorMemoryRepository {
	cloned := append([]DatabasePoint(nil), points...)
	return &coordinatorMemoryRepository{
		records:    make(map[uuid.UUID]Record),
		reservedAt: make(map[uuid.UUID]time.Time),
		boundAt:    make(map[uuid.UUID]time.Time),
		terminalAt: make(map[uuid.UUID]time.Time),
		points:     cloned,
		claims:     make(map[uuid.UUID]*AbortClaim),
		outcomes:   make(map[uuid.UUID]*PersistedAuthorityEffectOutcome),
	}
}

func (repository *coordinatorMemoryRepository) failTransaction(call int, after bool, err error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.transactionFail = &coordinatorTransactionFailure{call: call, after: after, err: err}
}

func (repository *coordinatorMemoryRepository) withAuthorityTransaction(ctx context.Context, operation func(store.DBTX) error) error {
	if ctx == nil || ctx.Err() != nil {
		return ErrCanceled
	}
	repository.mu.Lock()
	repository.transactions++
	call := repository.transactions
	failure := repository.transactionFail
	if failure != nil && failure.call == call && !failure.after {
		repository.transactionFail = nil
		repository.mu.Unlock()
		return failure.err
	}
	repository.mu.Unlock()

	err := operation(coordinatorNoopDBTX{})
	if err != nil {
		return err
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if failure != nil && failure.call == call && failure.after {
		repository.transactionFail = nil
		return failure.err
	}
	return nil
}

func (repository *coordinatorMemoryRepository) RecordPending(_ context.Context, _ store.DBTX, reservation Reservation, reservedAt time.Time) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if reservation.Validate() != nil || reservedAt.IsZero() {
		return ErrInvalidArgument
	}
	if current, ok := repository.records[reservation.OperationID]; ok {
		if current.Reservation == reservation && repository.reservedAt[reservation.OperationID].Equal(reservedAt) {
			return nil
		}
		return ErrConflict
	}
	repository.records[reservation.OperationID] = Record{Reservation: reservation}
	repository.reservedAt[reservation.OperationID] = reservedAt
	return nil
}

func (repository *coordinatorMemoryRepository) CaptureDatabasePoint(ctx context.Context) (DatabasePoint, error) {
	if ctx == nil || ctx.Err() != nil {
		return DatabasePoint{}, ErrCanceled
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if len(repository.points) == 0 {
		return DatabasePoint{}, ErrInjectedFailure
	}
	index := repository.captures
	if index >= len(repository.points) {
		index = len(repository.points) - 1
	}
	repository.captures++
	return repository.points[index], nil
}

func (repository *coordinatorMemoryRepository) BindEffect(ctx context.Context, dbtx store.DBTX, operationID uuid.UUID, effectDigest contracts.Digest, point DatabasePoint, boundAt time.Time) error {
	if tx, ok := dbtx.(*coordinatorMemoryTransaction); ok {
		return tx.shadow.BindEffect(ctx, coordinatorNoopDBTX{}, operationID, effectDigest, point, boundAt)
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	record, ok := repository.records[operationID]
	if !ok {
		return ErrNotFound
	}
	if record.TerminalReceipt != nil {
		return ErrTerminalConflict
	}
	if repository.claims[operationID] != nil {
		return ErrTerminalConflict
	}
	if record.BoundEffectDigest != nil {
		if *record.BoundEffectDigest == effectDigest && *record.BoundDatabasePoint == point && repository.boundAt[operationID].Equal(boundAt) {
			return nil
		}
		return ErrConflict
	}
	record.BoundEffectDigest = cloneDigestPointer(&effectDigest)
	record.BoundDatabasePoint = cloneDatabasePointPointer(&point)
	repository.records[operationID] = record
	repository.boundAt[operationID] = boundAt
	return nil
}

func (repository *coordinatorMemoryRepository) ActivateCommitted(ctx context.Context, dbtx store.DBTX, receipt Receipt, terminalAt time.Time) error {
	if tx, ok := dbtx.(*coordinatorMemoryTransaction); ok {
		return tx.shadow.ActivateCommitted(ctx, coordinatorNoopDBTX{}, receipt, terminalAt)
	}
	if repository.claims[receipt.OperationID] != nil {
		return ErrConflict
	}
	return repository.recordTerminal(receipt, terminalAt)
}

func (repository *coordinatorMemoryRepository) RecordAborted(ctx context.Context, dbtx store.DBTX, receipt Receipt, terminalAt time.Time) error {
	if tx, ok := dbtx.(*coordinatorMemoryTransaction); ok {
		return tx.shadow.RecordAborted(ctx, coordinatorNoopDBTX{}, receipt, terminalAt)
	}
	claim := repository.claims[receipt.OperationID]
	if claim == nil || receipt.AbortReason == nil || claim.Reason != *receipt.AbortReason {
		return ErrConflict
	}
	return repository.recordTerminal(receipt, terminalAt)
}

func (repository *coordinatorMemoryRepository) recordTerminal(receipt Receipt, terminalAt time.Time) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	record, ok := repository.records[receipt.OperationID]
	if !ok {
		return ErrNotFound
	}
	if record.Reservation != receipt.Reservation {
		return ErrConflict
	}
	if record.TerminalReceipt != nil {
		if record.TerminalReceipt.ReceiptDigest == receipt.ReceiptDigest && repository.terminalAt[receipt.OperationID].Equal(terminalAt) {
			return nil
		}
		return ErrTerminalConflict
	}
	if receipt.Status == StatusCommitted && (record.BoundEffectDigest == nil || !equalOptionalDigest(record.BoundEffectDigest, receipt.EffectDigest) || !equalOptionalDatabasePoint(record.BoundDatabasePoint, receipt.DatabasePoint)) {
		return ErrConflict
	}
	record.TerminalReceipt = func() *Receipt { cloned := cloneReceipt(receipt); return &cloned }()
	repository.records[receipt.OperationID] = record
	repository.terminalAt[receipt.OperationID] = terminalAt
	return nil
}

func (repository *coordinatorMemoryRepository) Get(ctx context.Context, operationID uuid.UUID) (Record, error) {
	if ctx == nil || ctx.Err() != nil {
		return Record{}, ErrCanceled
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	record, ok := repository.records[operationID]
	if !ok {
		return Record{}, ErrNotFound
	}
	return cloneRecord(record), nil
}

func (repository *coordinatorMemoryRepository) Head(ctx context.Context) (DatabaseHead, error) {
	if ctx == nil || ctx.Err() != nil {
		return DatabaseHead{}, ErrCanceled
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.headErr != nil {
		return DatabaseHead{}, repository.headErr
	}
	if repository.headOverride != nil {
		clone := *repository.headOverride
		clone.LatestCommittedDatabasePoint = cloneDatabasePointPointer(repository.headOverride.LatestCommittedDatabasePoint)
		return clone, nil
	}
	var head DatabaseHead
	for _, record := range repository.records {
		if record.Epoch > head.Epoch {
			head = DatabaseHead{Epoch: record.Epoch}
		}
	}
	for _, record := range repository.records {
		if record.Epoch != head.Epoch {
			continue
		}
		head.RecordCount++
		if record.Sequence > head.LatestReservedSequence {
			head.LatestReservedSequence = record.Sequence
			head.LatestReservationDigest = record.ReservationDigest
		}
		if record.TerminalReceipt == nil {
			head.PendingCount++
		}
		if record.TerminalReceipt != nil && record.TerminalReceipt.Status == StatusCommitted && record.Sequence > head.LatestCommittedSequence {
			head.LatestCommittedSequence = record.Sequence
			head.LatestCommittedOperationID = record.OperationID
			head.LatestCommittedReceiptDigest = record.TerminalReceipt.ReceiptDigest
			head.LatestCommittedDatabasePoint = cloneDatabasePointPointer(record.TerminalReceipt.DatabasePoint)
		}
	}
	head.HasSequenceGap = head.RecordCount != head.LatestReservedSequence
	return head, nil
}

func (repository *coordinatorMemoryRepository) setHeadError(err error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.headErr = err
}

func (repository *coordinatorMemoryRepository) setHeadOverride(head DatabaseHead) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	clone := head
	clone.LatestCommittedDatabasePoint = cloneDatabasePointPointer(head.LatestCommittedDatabasePoint)
	repository.headOverride = &clone
}

func (repository *coordinatorMemoryRepository) setCurrentPoint(point DatabasePoint) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.points = []DatabasePoint{point}
	repository.captures = 0
}

func (repository *coordinatorMemoryRepository) clearRecords() {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.records = make(map[uuid.UUID]Record)
	repository.reservedAt = make(map[uuid.UUID]time.Time)
	repository.boundAt = make(map[uuid.UUID]time.Time)
	repository.terminalAt = make(map[uuid.UUID]time.Time)
	repository.headOverride = nil
}

func (repository *coordinatorMemoryRepository) ListPending(ctx context.Context, epoch uint64) ([]PendingFence, error) {
	if ctx == nil || ctx.Err() != nil {
		return nil, ErrCanceled
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	result := make([]PendingFence, 0)
	for operationID, record := range repository.records {
		if record.Epoch != epoch || record.TerminalReceipt != nil {
			continue
		}
		result = append(result, PendingFence{
			OperationID:        operationID,
			Kind:               record.Kind,
			ScopeKind:          record.ScopeKind,
			ScopeDigest:        record.ScopeDigest,
			Epoch:              record.Epoch,
			Sequence:           record.Sequence,
			ReservationDigest:  record.ReservationDigest,
			BoundEffectDigest:  cloneDigestPointer(record.BoundEffectDigest),
			BoundDatabasePoint: cloneDatabasePointPointer(record.BoundDatabasePoint),
			ReservedAt:         repository.reservedAt[operationID],
			EffectBoundAt: func() *time.Time {
				if at, ok := repository.boundAt[operationID]; ok {
					return &at
				}
				return nil
			}(),
			AbortClaim: cloneAbortClaim(repository.claims[operationID]),
		})
	}
	return result, nil
}

func (repository *coordinatorMemoryRepository) CommittedNodeCheckpoint(context.Context, uint64, contracts.Digest) (NodeCheckpoint, error) {
	return NodeCheckpoint{}, ErrNotFound
}

func (repository *coordinatorMemoryRepository) captureCount() int {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	return repository.captures
}

func (repository *coordinatorMemoryRepository) transactionCount() int {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	return repository.transactions
}

type coordinatorNoopDBTX struct{}

func (coordinatorNoopDBTX) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected SQL from coordinator memory repository")
}

func (coordinatorNoopDBTX) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, errors.New("unexpected SQL from coordinator memory repository")
}

func (coordinatorNoopDBTX) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	return coordinatorNoopRow{}
}

type coordinatorNoopRow struct{}

func (coordinatorNoopRow) Scan(...interface{}) error {
	return errors.New("unexpected SQL from coordinator memory repository")
}

func walPositionOrdinal(t *testing.T, value WALPosition) uint64 {
	t.Helper()
	canonical, err := canonicalWALPosition(value)
	if err != nil || canonical != value {
		t.Fatalf("noncanonical WAL position %q", value)
	}
	var high, low uint64
	if _, err := fmt.Sscanf(string(value), "%X/%X", &high, &low); err != nil {
		t.Fatalf("parse WAL position %q: %v", value, err)
	}
	return high<<32 | low
}
