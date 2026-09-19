//go:build integration

package authority

import (
	"context"
	"crypto/sha256"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/store"
	"talenro.local/platform/internal/testinfra"
)

type pitrAuthorityRepository struct {
	*PostgresRepository
	access testinfra.C12AuthorityAccess
}

func (r *pitrAuthorityRepository) beginAuthorityTransaction(ctx context.Context) (authorityTransaction, error) {
	return r.access.Begin(ctx)
}

type pitrOperationFixture struct {
	request               ReserveRequest
	nodeID, certificateID uuid.UUID
	sequence              uint64
}

func pitrOperations() [3]pitrOperationFixture {
	namespace := uuid.MustParse("87bd4647-c131-4d50-96d4-f690ab276760")
	var operations [3]pitrOperationFixture
	for i, label := range []string{"prefix", "cut", "revoke"} {
		id := uuid.NewSHA1(namespace, []byte(label))
		nodeID := uuid.NewSHA1(namespace, []byte("node:"+label))
		scope := sha256.Sum256(append([]byte("TALENRO-NODE-AUTHORITY-SCOPE-V1\x00"), nodeID[:]...))
		operations[i] = pitrOperationFixture{request: ReserveRequest{OperationID: id, Kind: EffectCertificateRevoke, ScopeKind: ScopeNode, ScopeDigest: scope}, nodeID: nodeID, certificateID: uuid.NewSHA1(id, []byte("certificate")), sequence: uint64(i + 1)}
	}
	return operations
}

const pitrCertificateSnapshotSQL = `SELECT status,revoke_authority_operation_id,revoke_authority_effect_commitment_jcs,revoke_authority_activation_evidence_jcs,revoke_authority_effect_resolution_jcs FROM nodecontrol.node_certificates WHERE certificate_id=$1`
const pitrOperationSnapshotSQL = `SELECT provider_status,visibility_state,authority_protocol_profile FROM nodecontrol.control_plane_authority_fences WHERE operation_id=$1`
const pitrAuxCountSQL = `SELECT count(*) FROM nodecontrol.authority_task7_crash_effects WHERE operation_id=$1`

type pitrCertificateSnapshot struct {
	status                           string
	operationID                      uuid.NullUUID
	commitment, evidence, resolution []byte
}

func pitrReadCertificateSnapshot(ctx context.Context, db store.DBTX, id uuid.UUID) (pitrCertificateSnapshot, error) {
	var result pitrCertificateSnapshot
	err := db.QueryRow(ctx, pitrCertificateSnapshotSQL, id).Scan(&result.status, &result.operationID, &result.commitment, &result.evidence, &result.resolution)
	return result, err
}

type pitrOperationSnapshot struct{ status, visibility, profile string }

func pitrReadOperationSnapshot(ctx context.Context, db store.DBTX, id uuid.UUID) (pitrOperationSnapshot, error) {
	var result pitrOperationSnapshot
	err := db.QueryRow(ctx, pitrOperationSnapshotSQL, id).Scan(&result.status, &result.visibility, &result.profile)
	return result, err
}
func pitrReadAuxCount(ctx context.Context, db store.DBTX, id uuid.UUID) (int64, error) {
	var count int64
	err := db.QueryRow(ctx, pitrAuxCountSQL, id).Scan(&count)
	return count, err
}

type pitrObservedResolver struct {
	*postgresCrashEffectResolver
	controller  testinfra.C12AuthorityPITRController
	observation testinfra.C12AuthorityPITRObservation
}

func (r *pitrObservedResolver) ActivateAuthorityEffect(ctx context.Context, db store.DBTX, receipt Receipt, proof ValidatedActivationDecisionEvidence) error {
	if err := r.postgresCrashEffectResolver.ActivateAuthorityEffect(ctx, db, receipt, proof); err != nil {
		return err
	}
	tx, ok := db.(testinfra.C12AuthorityTransaction)
	if !ok {
		return ErrInvalidArgument
	}
	return r.controller.BindCommitObservation(ctx, r.observation, tx)
}

func pitrFinalizeObserved(t *testing.T, c testinfra.C12AuthorityPITRController, provider *DeterministicProvider, op pitrOperationFixture, activationID uuid.UUID, clock coordinatorClock) (Receipt, testinfra.C12AuthorityPITRObservedCommit, testinfra.C12AuthorityPITRCommit) {
	t.Helper()
	observation, err := c.NewCommitObservation()
	if err != nil {
		t.Fatal(err)
	}
	var receipt Receipt
	err = c.WithPrimaryAuthorityAccess(t.Context(), func(access testinfra.C12AuthorityAccess) error {
		ctx := t.Context()
		raw, err := NewPostgresRepository(access)
		if err != nil {
			return err
		}
		repository := &pitrAuthorityRepository{PostgresRepository: raw, access: access}
		resolver := &pitrObservedResolver{postgresCrashEffectResolver: newPostgresCrashEffectResolver(access), controller: c, observation: observation}
		coordinator := mustNewCoordinatorForTest(t, provider, repository, resolver, clock)
		reservation, err := coordinator.Reserve(ctx, op.request)
		if err != nil {
			return err
		}
		if reservation.Sequence != op.sequence {
			return ErrConflict
		}
		tx, err := access.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		if err = raw.RecordPendingClaimV1(ctx, tx, ClaimV1Reservation{Reservation: reservation, ProtocolActivationID: activationID}, clock.Now().Add(-time.Second)); err != nil {
			return err
		}
		digest, err := pitrCommitDomain(ctx, tx, raw, reservation, op.nodeID)
		if err != nil {
			return err
		}
		if err = tx.Commit(ctx); err != nil {
			return err
		}
		receipt, err = coordinator.Finalize(ctx, CoordinatorFinalizeRequest{OperationID: op.request.OperationID, EffectDigest: digest})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	observed, facts, err := c.ObserveCommit(t.Context(), observation)
	if err != nil {
		t.Fatal(err)
	}
	return receipt, observed, facts
}

func pitrCopyCommittedProvider(ctx context.Context, src *DeterministicProvider) (*DeterministicProvider, error) {
	head, err := src.Head(ctx)
	if err != nil {
		return nil, err
	}
	dst, err := NewDeterministicProvider(head.Epoch)
	if err != nil {
		return nil, err
	}
	for _, record := range src.Snapshot() {
		receipt := record.TerminalReceipt
		if receipt == nil || receipt.Status != StatusCommitted || receipt.DatabasePoint == nil || receipt.EffectDigest == nil {
			return nil, ErrInvalidArgument
		}
		reservation, err := dst.Reserve(ctx, ReserveRequest{OperationID: record.OperationID, Kind: record.Kind, ScopeKind: record.ScopeKind, ScopeDigest: record.ScopeDigest})
		if err != nil || reservation != record.Reservation {
			return nil, ErrConflict
		}
		point := receipt.DatabasePoint
		copied, err := dst.Finalize(ctx, FinalizeRequest{OperationID: record.OperationID, EffectDigest: *receipt.EffectDigest, DBSystemID: point.SystemID, DBTimeline: point.Timeline, RequiredLSN: point.RequiredLSN})
		if err != nil || !receiptEquals(copied, *receipt) {
			return nil, ErrConflict
		}
	}
	return dst, nil
}

// Catches shallow aliasing, unsorted replay and acceptance of noncommitted records.
func TestAuthorityPITRProviderReplayIsIndependent(t *testing.T) {
	provider, err := NewDeterministicProvider(31)
	if err != nil {
		t.Fatal(err)
	}
	ops := pitrOperations()
	finalize := func(op pitrOperationFixture) Receipt {
		t.Helper()
		reservation, err := provider.Reserve(t.Context(), op.request)
		if err != nil || reservation.Sequence != op.sequence {
			t.Fatalf("reserve %v %v", reservation, err)
		}
		receipt, err := provider.Finalize(t.Context(), FinalizeRequest{OperationID: op.request.OperationID, EffectDigest: sha256.Sum256([]byte(op.request.OperationID.String())), DBSystemID: 41, DBTimeline: 3, RequiredLSN: "0/40"})
		if err != nil {
			t.Fatal(err)
		}
		return receipt
	}
	finalize(ops[0])
	receiptA := finalize(ops[1])
	snapshot := provider.Snapshot()
	if len(snapshot) != 2 || snapshot[0].Sequence != 1 || snapshot[1].Sequence != 2 {
		t.Fatalf("Snapshot replay order: %#v", snapshot)
	}
	copyA, err := pitrCopyCommittedProvider(t.Context(), provider)
	if err != nil {
		t.Fatal("committed replay failed:", err)
	}
	if !reflect.DeepEqual(copyA.Snapshot(), snapshot) {
		t.Fatal("replayed reservations/receipts changed")
	}
	// Snapshot and replay must not alias either provider's nested receipt fields.
	snapshot[1].TerminalReceipt.DatabasePoint.Timeline = 999
	*snapshot[1].TerminalReceipt.EffectDigest = contracts.Digest{}
	finalize(ops[2])
	headA, err := copyA.Head(t.Context())
	if err != nil || headA.LatestReservedSequence != 2 || headA.LatestCommittedSequence != 2 || headA.LatestCommittedReceiptDigest != receiptA.ReceiptDigest || !reflect.DeepEqual(headA.LatestCommittedDatabasePoint, receiptA.DatabasePoint) {
		t.Fatalf("A copy changed: %#v %v", headA, err)
	}
	recordA, err := copyA.Inspect(t.Context(), ops[1].request.OperationID)
	if err != nil || !receiptEquals(*recordA.TerminalReceipt, receiptA) {
		t.Fatal("A copy receipt aliased")
	}
	for _, terminal := range []bool{false, true} {
		t.Run(map[bool]string{false: "pending", true: "aborted"}[terminal], func(t *testing.T) {
			other, _ := NewDeterministicProvider(31)
			if _, err := other.Reserve(t.Context(), ops[0].request); err != nil {
				t.Fatal(err)
			}
			if terminal {
				if _, err := other.Abort(t.Context(), AbortRequest{OperationID: ops[0].request.OperationID, Reason: AbortValidationFailed}); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := pitrCopyCommittedProvider(t.Context(), other); err != ErrInvalidArgument {
				t.Fatalf("noncommitted replay = %v", err)
			}
		})
	}
}

// A deliberately narrow transport double: it neither embeds nor exposes pgx.Tx.
type pitrBridgeFakeAccess struct {
	tx           *pitrBridgeFakeTx
	active       *pitrBridgeFakeTx
	Calls        int
	memory       *coordinatorMemoryRepository
	transactions []*pitrBridgeFakeTx
	queryRow     func(string, ...any) pgx.Row
}

type pitrBridgeFakeTx struct {
	access                    *pitrBridgeFakeAccess
	Commits, Rollbacks, Calls int
	Events                    []string
	done                      bool
	memory                    *coordinatorMemoryTransaction
	queryRow                  func(string, ...any) pgx.Row
	exec                      func(string, ...any) (pgconn.CommandTag, error)
}

func (a *pitrBridgeFakeAccess) Begin(ctx context.Context) (testinfra.C12AuthorityTransaction, error) {
	if a.active != nil {
		return nil, ErrConflict
	}
	if a.memory != nil {
		inner, err := a.memory.beginAuthorityTransaction(ctx)
		if err != nil {
			return nil, err
		}
		a.tx = &pitrBridgeFakeTx{memory: inner.(*coordinatorMemoryTransaction)}
	}
	a.active = a.tx
	a.tx.access = a
	a.tx.Events = append(a.tx.Events, "begin")
	a.transactions = append(a.transactions, a.tx)
	return a.tx, nil
}
func (a *pitrBridgeFakeAccess) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	a.Calls++
	return pgconn.CommandTag{}, ErrInjectedFailure
}
func (a *pitrBridgeFakeAccess) Query(context.Context, string, ...any) (pgx.Rows, error) {
	a.Calls++
	return nil, ErrInjectedFailure
}
func (a *pitrBridgeFakeAccess) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	a.Calls++
	if a.queryRow != nil {
		return a.queryRow(sql, args...)
	}
	return pitrBridgeErrorRow{}
}
func (tx *pitrBridgeFakeTx) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	tx.Calls++
	if tx.exec != nil {
		return tx.exec(sql, args...)
	}
	return pgconn.CommandTag{}, ErrInjectedFailure
}
func (tx *pitrBridgeFakeTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	tx.Calls++
	return nil, ErrInjectedFailure
}
func (tx *pitrBridgeFakeTx) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	tx.Calls++
	if tx.queryRow != nil {
		return tx.queryRow(sql, args...)
	}
	return pitrBridgeErrorRow{}
}
func (tx *pitrBridgeFakeTx) Commit(ctx context.Context) error {
	tx.Commits++
	tx.Events = append(tx.Events, "commit")
	if tx.done {
		return ErrConflict
	}
	if tx.memory != nil {
		if err := tx.memory.Commit(ctx); err != nil {
			return err
		}
	}
	tx.done = true
	if tx.access.active == tx {
		tx.access.active = nil
	}
	return nil
}
func (tx *pitrBridgeFakeTx) Rollback(ctx context.Context) error {
	tx.Rollbacks++
	tx.Events = append(tx.Events, "rollback")
	if tx.memory != nil {
		if err := tx.memory.Rollback(ctx); err != nil {
			return err
		}
	}
	if !tx.done && tx.access.active == tx {
		tx.access.active = nil
	}
	tx.done = true
	return nil
}

type pitrBridgeErrorRow struct{}

func (pitrBridgeErrorRow) Scan(...any) error { return ErrInjectedFailure }

// R22: this adapter only translates the scheduling double's transaction into
// the existing memory fixture. No algorithm or production implementation is copied.
// It is never used by the physical P/A/B regression.
type pitrBridgeMemoryRepository struct {
	*coordinatorMemoryRepository
	bridge *pitrAuthorityRepository
}

func (r *pitrBridgeMemoryRepository) beginAuthorityTransaction(ctx context.Context) (authorityTransaction, error) {
	return r.bridge.beginAuthorityTransaction(ctx)
}
func (r *pitrBridgeMemoryRepository) Lock(ctx context.Context, db store.DBTX, id uuid.UUID) (StoredFence, error) {
	return r.coordinatorMemoryRepository.Lock(ctx, db.(*pitrBridgeFakeTx).memory, id)
}
func (r *pitrBridgeMemoryRepository) BindEffect(ctx context.Context, db store.DBTX, id uuid.UUID, digest contracts.Digest, point DatabasePoint, at time.Time) error {
	return r.coordinatorMemoryRepository.BindEffect(ctx, db.(*pitrBridgeFakeTx).memory, id, digest, point, at)
}
func (r *pitrBridgeMemoryRepository) ActivateCommitted(ctx context.Context, db store.DBTX, receipt Receipt, at time.Time) error {
	return r.coordinatorMemoryRepository.ActivateCommitted(ctx, db.(*pitrBridgeFakeTx).memory, receipt, at)
}
func (r *pitrBridgeMemoryRepository) readAuthoritySnapshot(ctx context.Context, db store.DBTX, epoch uint64) (DatabasePoint, DatabaseHead, []PendingFence, error) {
	return r.coordinatorMemoryRepository.readAuthoritySnapshot(ctx, db.(*pitrBridgeFakeTx).memory, epoch)
}

type pitrBridgeMemoryHandler struct{ inner *coordinatorTestHandler }

func (h *pitrBridgeMemoryHandler) ResolveAuthorityEffect(ctx context.Context, id uuid.UUID) (ResolvedEffect, error) {
	return h.inner.source.ResolveAuthorityEffect(ctx, id)
}
func (h *pitrBridgeMemoryHandler) ResolveRegisteredAuthorityEffectForUpdate(ctx context.Context, db store.DBTX, q TransactionalEffectQuery) (TransactionalResolvedEffect, error) {
	return h.inner.ResolveRegisteredAuthorityEffectForUpdate(ctx, db.(*pitrBridgeFakeTx).memory, q)
}
func (h *pitrBridgeMemoryHandler) CaptureActivationDecisionMaterial(ctx context.Context, r Receipt) (ActivationDecisionMaterial, error) {
	return h.inner.CaptureActivationDecisionMaterial(ctx, r)
}
func (h *pitrBridgeMemoryHandler) ActivateAuthorityEffect(ctx context.Context, db store.DBTX, r Receipt, p ValidatedActivationDecisionEvidence) error {
	return h.inner.ActivateAuthorityEffect(ctx, db.(*pitrBridgeFakeTx).memory, r, p)
}
func (h *pitrBridgeMemoryHandler) ValidatePersistedAuthorityEffect(ctx context.Context, db store.DBTX, r Receipt, p ValidatedActivationDecisionEvidence, resolution AuthorityEffectResolution) error {
	return h.inner.ValidatePersistedAuthorityEffect(ctx, db.(*pitrBridgeFakeTx).memory, r, p, resolution)
}

type pitrBridgeGuardProvider struct {
	Provider
	access     *pitrBridgeFakeAccess
	calls      map[string]int
	violations int
}

func (p *pitrBridgeGuardProvider) check(name string) error {
	p.calls[name]++
	if p.access.active != nil {
		p.violations++
		return ErrConflict
	}
	return nil
}
func (p *pitrBridgeGuardProvider) Reserve(ctx context.Context, r ReserveRequest) (Reservation, error) {
	if err := p.check("reserve"); err != nil {
		return Reservation{}, err
	}
	return p.Provider.Reserve(ctx, r)
}
func (p *pitrBridgeGuardProvider) Finalize(ctx context.Context, r FinalizeRequest) (Receipt, error) {
	if err := p.check("finalize"); err != nil {
		return Receipt{}, err
	}
	return p.Provider.Finalize(ctx, r)
}
func (p *pitrBridgeGuardProvider) Abort(ctx context.Context, r AbortRequest) (Receipt, error) {
	if err := p.check("abort"); err != nil {
		return Receipt{}, err
	}
	return p.Provider.Abort(ctx, r)
}
func (p *pitrBridgeGuardProvider) Inspect(ctx context.Context, id uuid.UUID) (Record, error) {
	if err := p.check("inspect"); err != nil {
		return Record{}, err
	}
	return p.Provider.Inspect(ctx, id)
}
func (p *pitrBridgeGuardProvider) Head(ctx context.Context) (Head, error) {
	if err := p.check("head"); err != nil {
		return Head{}, err
	}
	return p.Provider.Head(ctx)
}
func (p *pitrBridgeGuardProvider) CommittedNodeCheckpoint(ctx context.Context, digest contracts.Digest) (NodeCheckpoint, error) {
	if err := p.check("checkpoint"); err != nil {
		return NodeCheckpoint{}, err
	}
	return p.Provider.CommittedNodeCheckpoint(ctx, digest)
}

func TestAuthorityPITRBridgeProviderCallsOutsideTransaction(t *testing.T) {
	memory := newCoordinatorMemoryRepository([]DatabasePoint{{SystemID: 41, Timeline: 3, RequiredLSN: "0/40"}})
	access := &pitrBridgeFakeAccess{memory: memory}
	repository := &pitrBridgeMemoryRepository{coordinatorMemoryRepository: memory, bridge: &pitrAuthorityRepository{access: access}}
	provider, err := NewDeterministicProvider(31)
	if err != nil {
		t.Fatal(err)
	}
	guard := &pitrBridgeGuardProvider{Provider: provider, access: access, calls: map[string]int{}}
	effects := newCoordinatorEffectResolver()
	effects.repository, effects.conditional = memory, true
	handler := &pitrBridgeMemoryHandler{inner: &coordinatorTestHandler{source: effects, repository: memory}}
	clock := coordinatorClock{now: time.Date(2026, 8, 23, 13, 0, 0, 0, time.UTC)}
	coordinator := mustNewCoordinatorForTest(t, guard, repository, handler, clock)
	op := pitrOperations()[0]
	coordinatorReserveAndRecord(t, coordinator, memory, op.request)
	digest := effects.commit(t, op.request, sha256.Sum256([]byte("pitr-guard-domain")), "0/20")
	receipt, err := coordinator.Finalize(t.Context(), CoordinatorFinalizeRequest{OperationID: op.request.OperationID, EffectDigest: digest})
	if err != nil || receipt.Status != StatusCommitted {
		t.Fatalf("real Coordinator Finalize: %#v %v", receipt, err)
	}
	ready, err := coordinator.CheckReady(t.Context())
	if err != nil || !ready.Ready || ready.Reason != ReadinessReady {
		t.Fatalf("real Coordinator readiness: %#v %v", ready, err)
	}
	if guard.violations != 0 || access.active != nil {
		t.Fatalf("provider called under lock: %#v", guard)
	}
	for _, method := range []string{"reserve", "inspect", "finalize", "head"} {
		if guard.calls[method] == 0 {
			t.Fatalf("provider method %s not exercised", method)
		}
	}
	if len(access.transactions) < 5 {
		t.Fatalf("fresh Finalize/readiness did not exercise bridge: %d", len(access.transactions))
	}
	for _, tx := range access.transactions {
		if tx.Commits != 1 || tx.Rollbacks != 1 || !tx.done {
			t.Fatalf("Coordinator transaction accounting: %#v", tx)
		}
	}
	// Negative control proves the guard actually detects active fake transactions.
	tx, err := repository.beginAuthorityTransaction(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := guard.Head(t.Context()); err != ErrConflict || guard.violations != 1 {
		t.Fatal("active-transaction guard did not reject provider call")
	}
	if err := tx.Rollback(t.Context()); err != nil {
		t.Fatal(err)
	}
}

type pitrBridgeScanRow func(...any) error

func (row pitrBridgeScanRow) Scan(targets ...any) error { return row(targets...) }

// The original registered handler runs here; only its SQL transport is a double.
// This is execution-order evidence, not PostgreSQL atomicity/physical acceptance.
func TestAuthorityPITRBridgeRealHandlerTransactionOrder(t *testing.T) {
	for _, mode := range []string{"success", "audit_failure", "bind_after_real_writes"} {
		t.Run(mode, func(t *testing.T) {
			op := pitrOperations()[0]
			provider, _ := NewDeterministicProvider(31)
			reservation, err := provider.Reserve(t.Context(), op.request)
			if err != nil {
				t.Fatal(err)
			}
			commitment, err := NewAuthorityEffectCommitment(AuthorityEffectCommitmentInput{OperationID: reservation.OperationID, Kind: reservation.Kind, ScopeKind: reservation.ScopeKind, ScopeDigest: reservation.ScopeDigest, Epoch: reservation.Epoch, Sequence: reservation.Sequence, BaseEffectDigest: sha256.Sum256([]byte("task9-revoke:" + reservation.OperationID.String())), Mode: CommitmentConditionalApply, Reason: EffectReasonNone, ActivationPolicyVersion: 1, ActivationInputsDigest: task9CertificateInputDigest(reservation.OperationID, op.nodeID, op.certificateID, reservation.ScopeDigest[:])})
			if err != nil {
				t.Fatal(err)
			}
			receipt, err := provider.Finalize(t.Context(), FinalizeRequest{OperationID: reservation.OperationID, EffectDigest: commitment.Digest(), DBSystemID: 41, DBTimeline: 3, RequiredLSN: "0/40"})
			if err != nil {
				t.Fatal(err)
			}
			access := &pitrBridgeFakeAccess{tx: &pitrBridgeFakeTx{}}
			access.queryRow = func(sql string, args ...any) pgx.Row {
				if sql != "SELECT revoke_authority_effect_commitment_jcs FROM nodecontrol.node_certificates WHERE revoke_authority_operation_id=$1" || len(args) != 1 || args[0] != receipt.OperationID || access.active != nil {
					t.Fatal("material capture must be outside transaction with exact operation")
				}
				return pitrBridgeScanRow(func(dst ...any) error {
					if len(dst) != 1 {
						return ErrConflict
					}
					*dst[0].(*[]byte) = commitment.CanonicalJCS()
					return nil
				})
			}
			access.tx.queryRow = func(sql string, args ...any) pgx.Row {
				if sql != "SELECT node_id,certificate_id,leaf_der_sha256,revoke_authority_effect_commitment_jcs FROM nodecontrol.node_certificates WHERE revoke_authority_operation_id=$1 FOR UPDATE" || len(args) != 1 || args[0] != receipt.OperationID || access.active != access.tx || access.tx.Commits != 0 {
					t.Fatal("handler input read escaped exact live transaction")
				}
				access.tx.Events = append(access.tx.Events, "inputs")
				return pitrBridgeScanRow(func(dst ...any) error {
					if len(dst) != 4 {
						return ErrConflict
					}
					*dst[0].(*uuid.UUID) = op.nodeID
					*dst[1].(*uuid.UUID) = op.certificateID
					*dst[2].(*[]byte) = append([]byte(nil), reservation.ScopeDigest[:]...)
					*dst[3].(*[]byte) = commitment.CanonicalJCS()
					return nil
				})
			}
			access.tx.exec = func(sql string, args ...any) (pgconn.CommandTag, error) {
				if access.active != access.tx || access.tx.Commits != 0 || len(args) == 0 || args[0] != receipt.OperationID {
					t.Fatal("handler wrote after commit or with wrong operation")
				}
				var event string
				switch {
				case strings.HasPrefix(sql, "UPDATE nodecontrol.node_certificates SET\n"):
					event = "certificate"
					if len(args) != 12 {
						t.Fatal("certificate argument shape")
					}
				case strings.HasPrefix(sql, "INSERT INTO nodecontrol.node_operator_audit("):
					event = "audit"
					if len(args) != 7 {
						t.Fatal("audit argument shape")
					}
				case strings.HasPrefix(sql, "INSERT INTO public.transactional_outbox("):
					event = "outbox"
					if len(args) != 6 {
						t.Fatal("outbox argument shape")
					}
				default:
					t.Fatal("unexpected handler SQL")
				}
				access.tx.Events = append(access.tx.Events, event)
				return pgconn.NewCommandTag("UPDATE 1"), nil
			}
			raw, err := NewPostgresRepository(access)
			if err != nil {
				t.Fatal(err)
			}
			repository := &pitrAuthorityRepository{PostgresRepository: raw, access: access}
			resolver := newPostgresCrashEffectResolver(access)
			coordinator := mustNewCoordinatorForTest(t, provider, repository, resolver, coordinatorClock{now: time.Date(2026, 8, 23, 13, 0, 0, 0, time.UTC)})
			proof, err := coordinator.captureActivation(t.Context(), receipt)
			if err != nil {
				t.Fatal(err)
			}
			if mode == "audit_failure" {
				resolver.activationFailure = "audit"
			}
			err = coordinator.transact(t.Context(), &proof, func(db store.DBTX) error {
				if mode == "audit_failure" {
					observed := &pitrObservedResolver{postgresCrashEffectResolver: resolver}
					activationErr := observed.ActivateAuthorityEffect(t.Context(), db, receipt, proof)
					if activationErr != ErrInjectedFailure {
						t.Fatalf("original handler failure was replaced by binding: %v", activationErr)
					}
					return activationErr
				}
				if mode == "bind_after_real_writes" {
					// A zero controller cannot bind, but rejection must happen only
					// after the original handler completed all three writes.
					observed := &pitrObservedResolver{postgresCrashEffectResolver: resolver}
					bindErr := observed.ActivateAuthorityEffect(t.Context(), db, receipt, proof)
					if bindErr != testinfra.C12PITRInvalidHandle {
						t.Fatalf("bind boundary = %v", bindErr)
					}
					access.tx.Events = append(access.tx.Events, "bind_rejected")
					return bindErr
				}
				return resolver.ActivateAuthorityEffect(t.Context(), db, receipt, proof)
			})
			want := []string{"begin", "inputs", "certificate", "audit", "outbox", "commit", "rollback"}
			commits := 1
			if mode == "audit_failure" {
				want = []string{"begin", "inputs", "certificate", "audit", "rollback"}
				commits = 0
			}
			if mode == "bind_after_real_writes" {
				want = []string{"begin", "inputs", "certificate", "audit", "outbox", "bind_rejected", "rollback"}
				commits = 0
			}
			if (mode == "success" && err != nil) || (mode != "success" && err == nil) || access.tx.Commits != commits || access.tx.Rollbacks != 1 || !reflect.DeepEqual(access.tx.Events, want) {
				t.Fatalf("real handler commit order: err=%v events=%v commits=%d rollbacks=%d", err, access.tx.Events, access.tx.Commits, access.tx.Rollbacks)
			}
		})
	}
}

// Catches replacement of the controlled object and premature/duplicate commit.
func TestAuthorityPITRBridgeUsesControlledTransaction(t *testing.T) {
	t.Run("exact_object", func(t *testing.T) {
		access := &pitrBridgeFakeAccess{tx: &pitrBridgeFakeTx{}}
		repository := &pitrAuthorityRepository{access: access}
		tx, err := repository.beginAuthorityTransaction(t.Context())
		if err != nil || tx != access.tx {
			t.Fatal("bridge replaced exact controlled transaction")
		}
		if _, ok := any(tx).(pgx.Tx); ok {
			t.Fatal("bridge exposed full pgx.Tx")
		}
		if err := tx.Rollback(t.Context()); err != nil {
			t.Fatal(err)
		}
	})
	for _, fail := range []bool{false, true} {
		name := "success"
		if fail {
			name = "operation_failure"
		}
		t.Run(name, func(t *testing.T) {
			access := &pitrBridgeFakeAccess{tx: &pitrBridgeFakeTx{}}
			coordinator := &Coordinator{transaction: &pitrAuthorityRepository{access: access}}
			err := coordinator.transact(t.Context(), nil, func(db store.DBTX) error {
				if db != access.tx || access.tx.Commits != 0 || access.active != access.tx {
					t.Fatal("operation did not retain exact live transaction")
				}
				access.tx.Events = append(access.tx.Events, "operation")
				if fail {
					return ErrConflict
				}
				return nil
			})
			wantEvents := []string{"begin", "operation", "commit", "rollback"}
			wantCommits := 1
			if fail {
				wantCommits = 0
				wantEvents = []string{"begin", "operation", "rollback"}
				if err != ErrConflict {
					t.Fatalf("operation error = %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if access.tx.Commits != wantCommits || access.tx.Rollbacks != 1 || access.active != nil || !reflect.DeepEqual(access.tx.Events, wantEvents) {
				t.Fatalf("transaction trace: %#v", access.tx)
			}
			// Deferred rollback of a finished transaction must not release a later one.
			old := access.tx
			access.tx = &pitrBridgeFakeTx{}
			if _, err := access.Begin(t.Context()); err != nil {
				t.Fatal(err)
			}
			if err := old.Rollback(t.Context()); err != nil || access.active != access.tx {
				t.Fatal("old rollback cleared later active transaction")
			}
		})
	}
	t.Run("registered_handlers_do_not_own_transactions", func(t *testing.T) {
		wanted := map[string]bool{"task9ReadCertificateInput": false, "pitrObservedResolver.ActivateAuthorityEffect": false}
		for _, receiver := range []string{"postgresCrashEffectResolver", "coordinatorTestHandler"} {
			for _, method := range []string{"ResolveRegisteredAuthorityEffectForUpdate", "CaptureActivationDecisionMaterial", "ActivateAuthorityEffect", "ValidatePersistedAuthorityEffect"} {
				wanted[receiver+"."+method] = false
			}
		}
		for _, path := range []string{"authority_crash_integration_test.go", "authority_pitr_bridge_integration_test.go", "coordinator_test.go"} {
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				name := fn.Name.Name
				if fn.Recv != nil && len(fn.Recv.List) == 1 {
					if pointer, ok := fn.Recv.List[0].Type.(*ast.StarExpr); ok {
						if receiver, ok := pointer.X.(*ast.Ident); ok {
							name = receiver.Name + "." + name
						}
					}
				}
				if _, ok := wanted[name]; !ok {
					continue
				}
				wanted[name] = true
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					if call, ok := n.(*ast.CallExpr); ok {
						if s, ok := call.Fun.(*ast.SelectorExpr); ok {
							switch s.Sel.Name {
							case "Commit", "Rollback", "Begin", "BeginTx":
								t.Errorf("handler %s owns transaction via %s", fn.Name.Name, s.Sel.Name)
							}
						}
					}
					if assertion, ok := n.(*ast.TypeAssertExpr); ok {
						if s, ok := assertion.Type.(*ast.SelectorExpr); ok {
							if p, ok := s.X.(*ast.Ident); ok && p.Name == "pgx" && s.Sel.Name == "Tx" {
								t.Errorf("handler %s unwraps pgx.Tx", fn.Name.Name)
							}
						}
					}
					return true
				})
			}
		}
		for name, seen := range wanted {
			if !seen {
				t.Errorf("missing handler %s", name)
			}
		}
	})
}
