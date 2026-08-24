//go:build integration

package authority

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

func TestPostgresRepositoryRecordsExactFenceState(t *testing.T) {
	ctx, pool := openMigratedAuthorityDatabase(t)
	repository, err := NewPostgresRepository(pool)
	if err != nil {
		t.Fatal(err)
	}
	provider, err := NewDeterministicProvider(37)
	if err != nil {
		t.Fatal(err)
	}
	nodeScope := contracts.Digest(sha256.Sum256([]byte("task-6-integration-node")))
	baseTime := time.Date(2026, time.August, 24, 10, 0, 0, 123456000, time.UTC)

	reservedOnly := reserveAuthorityIntegration(t, provider, "10000000-0000-4000-8000-000000000001", EffectDesiredActivate, ScopeNode, nodeScope)
	recordPendingAuthorityIntegration(t, ctx, pool, repository, reservedOnly, baseTime)
	recordPendingAuthorityIntegration(t, ctx, pool, repository, reservedOnly, baseTime)

	changed := reservedOnly
	changed.Sequence = 9
	changed.ReservationDigest, err = reservationDigest(changed)
	if err != nil {
		t.Fatal(err)
	}
	changedErr := inAuthorityTransaction(ctx, pool, func(tx pgx.Tx) error {
		return repository.RecordPending(ctx, tx, changed, baseTime)
	})
	if !errors.Is(changedErr, ErrConflict) {
		t.Fatalf("changed pending retry error = %v, want ErrConflict", changedErr)
	}

	effectBound := reserveAuthorityIntegration(t, provider, "10000000-0000-4000-8000-000000000002", EffectRecoveryActivate, ScopeNode, nodeScope)
	recordPendingAuthorityIntegration(t, ctx, pool, repository, effectBound, baseTime.Add(time.Second))
	effectDigest := contracts.Digest(sha256.Sum256([]byte("task-6-bound-effect")))
	point, err := repository.CaptureDatabasePoint(ctx)
	if err != nil {
		t.Fatal(err)
	}
	boundAt := baseTime.Add(2 * time.Second)
	bindAuthorityIntegration(t, ctx, pool, repository, effectBound.OperationID, effectDigest, point, boundAt)
	bindAuthorityIntegration(t, ctx, pool, repository, effectBound.OperationID, effectDigest, point, boundAt)

	committed := reserveAuthorityIntegration(t, provider, "10000000-0000-4000-8000-000000000003", EffectCertificateRevoke, ScopeNode, nodeScope)
	recordPendingAuthorityIntegration(t, ctx, pool, repository, committed, baseTime.Add(3*time.Second))
	committedEffect := contracts.Digest(sha256.Sum256([]byte("task-6-committed-effect")))
	committedPoint, err := repository.CaptureDatabasePoint(ctx)
	if err != nil {
		t.Fatal(err)
	}
	committedBoundAt := baseTime.Add(4 * time.Second)
	bindAuthorityIntegration(t, ctx, pool, repository, committed.OperationID, committedEffect, committedPoint, committedBoundAt)

	mismatchedPoint := committedPoint
	if mismatchedPoint.Timeline == ^uint32(0) {
		mismatchedPoint.Timeline--
	} else {
		mismatchedPoint.Timeline++
	}
	mismatchReceipt := Receipt{
		Reservation:   committed,
		EffectDigest:  &committedEffect,
		DatabasePoint: &mismatchedPoint,
		Status:        StatusCommitted,
	}
	mismatchReceipt.ReceiptDigest, err = receiptDigest(mismatchReceipt)
	if err != nil || mismatchReceipt.Validate() != nil {
		t.Fatalf("construct timeline mismatch receipt: %v", err)
	}
	mismatchErr := inAuthorityTransaction(ctx, pool, func(tx pgx.Tx) error {
		return repository.ActivateCommitted(ctx, tx, mismatchReceipt, baseTime.Add(5*time.Second))
	})
	if !errors.Is(mismatchErr, ErrConflict) {
		t.Fatalf("timeline mismatch error = %v, want ErrConflict", mismatchErr)
	}

	committedReceipt, err := provider.Finalize(ctx, FinalizeRequest{
		OperationID: committed.OperationID, EffectDigest: committedEffect, DBSystemID: committedPoint.SystemID,
		DBTimeline: committedPoint.Timeline, RequiredLSN: committedPoint.RequiredLSN,
	})
	if err != nil {
		t.Fatal(err)
	}
	terminalAt := baseTime.Add(5 * time.Second)
	activateAuthorityIntegration(t, ctx, pool, repository, committedReceipt, terminalAt)
	activateAuthorityIntegration(t, ctx, pool, repository, committedReceipt, terminalAt)
	recordPendingAuthorityIntegration(t, ctx, pool, repository, committed, baseTime.Add(3*time.Second))

	aborted := reserveAuthorityIntegration(t, provider, "10000000-0000-4000-8000-000000000004", EffectOperatorTransition, ScopeNode, nodeScope)
	recordPendingAuthorityIntegration(t, ctx, pool, repository, aborted, baseTime.Add(6*time.Second))
	abortedReceipt, err := provider.Abort(ctx, AbortRequest{OperationID: aborted.OperationID, Reason: AbortSuperseded})
	if err != nil {
		t.Fatal(err)
	}
	abortedAt := baseTime.Add(7 * time.Second)
	abortAuthorityIntegration(t, ctx, pool, repository, abortedReceipt, abortedAt)
	abortAuthorityIntegration(t, ctx, pool, repository, abortedReceipt, abortedAt)

	head, err := repository.Head(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if head.Epoch != 37 || head.RecordCount != 4 || head.LatestReservedSequence != 4 ||
		head.LatestCommittedSequence != 3 || head.PendingCount != 2 || head.HasSequenceGap ||
		head.LatestReservationDigest != aborted.ReservationDigest || head.LatestCommittedOperationID != committed.OperationID ||
		head.LatestCommittedReceiptDigest != committedReceipt.ReceiptDigest || head.LatestCommittedDatabasePoint == nil ||
		*head.LatestCommittedDatabasePoint != committedPoint {
		t.Fatalf("head = %#v", head)
	}
	pending, err := repository.ListPending(ctx, 37)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 2 || pending[0].OperationID != reservedOnly.OperationID || pending[1].OperationID != effectBound.OperationID ||
		pending[0].BoundEffectDigest != nil || pending[0].BoundDatabasePoint != nil || pending[0].EffectBoundAt != nil ||
		pending[1].BoundEffectDigest == nil || *pending[1].BoundEffectDigest != effectDigest ||
		pending[1].BoundDatabasePoint == nil || *pending[1].BoundDatabasePoint != point ||
		pending[1].EffectBoundAt == nil || !pending[1].EffectBoundAt.Equal(boundAt) {
		t.Fatalf("pending = %#v", pending)
	}
	for _, expected := range []struct {
		reservation Reservation
		status      *ReceiptStatus
	}{
		{reservation: reservedOnly},
		{reservation: effectBound},
		{reservation: committed, status: receiptStatusPointer(StatusCommitted)},
		{reservation: aborted, status: receiptStatusPointer(StatusAborted)},
	} {
		record, getErr := repository.Get(ctx, expected.reservation.OperationID)
		if getErr != nil || record.Reservation != expected.reservation {
			t.Fatalf("record %s = %#v, %v", expected.reservation.OperationID, record, getErr)
		}
		if expected.status == nil && record.TerminalReceipt != nil {
			t.Fatalf("pending record unexpectedly terminal: %#v", record)
		}
		if expected.status != nil && (record.TerminalReceipt == nil || record.TerminalReceipt.Status != *expected.status) {
			t.Fatalf("terminal record status = %#v, want %s", record.TerminalReceipt, *expected.status)
		}
	}

	assertConcurrentAuthoritySequenceForkRejected(t, ctx, pool, repository, nodeScope, baseTime.Add(8*time.Second))

	rolloverOne := manualAuthorityReservation(t, "20000000-0000-4000-8000-000000000001", EffectGrantCreate, ScopeNode, nodeScope, 38, 1)
	rolloverThree := manualAuthorityReservation(t, "20000000-0000-4000-8000-000000000003", EffectGrantClaim, ScopeNode, nodeScope, 38, 3)
	recordPendingAuthorityIntegration(t, ctx, pool, repository, rolloverOne, baseTime.Add(9*time.Second))
	recordPendingAuthorityIntegration(t, ctx, pool, repository, rolloverThree, baseTime.Add(10*time.Second))
	rolloverHead, err := repository.Head(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if rolloverHead.Epoch != 38 || rolloverHead.RecordCount != 2 || rolloverHead.LatestReservedSequence != 3 ||
		rolloverHead.LatestCommittedSequence != 0 || rolloverHead.PendingCount != 2 || !rolloverHead.HasSequenceGap ||
		rolloverHead.LatestReservationDigest != rolloverThree.ReservationDigest || rolloverHead.LatestCommittedOperationID != uuid.Nil ||
		rolloverHead.LatestCommittedReceiptDigest != (contracts.Digest{}) || rolloverHead.LatestCommittedDatabasePoint != nil {
		t.Fatalf("rollover head = %#v", rolloverHead)
	}
	rolloverPending, err := repository.ListPending(ctx, 38)
	if err != nil || len(rolloverPending) != 2 || rolloverPending[0].Sequence != 1 || rolloverPending[1].Sequence != 3 {
		t.Fatalf("rollover pending = %#v, %v", rolloverPending, err)
	}
}

func TestAuthorityHeadCommittedNodeCheckpointScope(t *testing.T) {
	ctx, pool := openMigratedAuthorityDatabase(t)
	repository, err := NewPostgresRepository(pool)
	if err != nil {
		t.Fatal(err)
	}
	provider, err := NewDeterministicProvider(43)
	if err != nil {
		t.Fatal(err)
	}
	nodeA := contracts.Digest(sha256.Sum256([]byte("task-6-checkpoint-node-a")))
	nodeB := contracts.Digest(sha256.Sum256([]byte("task-6-checkpoint-node-b")))
	globalNode := contracts.Digest(sha256.Sum256([]byte(globalNodeScopeTranscript)))
	globalOperator := contracts.Digest(sha256.Sum256([]byte(globalOperatorTranscript)))
	now := time.Date(2026, time.August, 24, 11, 0, 0, 0, time.UTC)

	commit := func(operationID string, kind EffectKind, scope ScopeKind, digest contracts.Digest, marker byte) Receipt {
		t.Helper()
		reservation := reserveAuthorityIntegration(t, provider, operationID, kind, scope, digest)
		recordPendingAuthorityIntegration(t, ctx, pool, repository, reservation, now.Add(time.Duration(reservation.Sequence)*time.Second))
		effect := contracts.Digest{marker}
		point, captureErr := repository.CaptureDatabasePoint(ctx)
		if captureErr != nil {
			t.Fatal(captureErr)
		}
		boundAt := now.Add((time.Duration(reservation.Sequence) + 10) * time.Second)
		bindAuthorityIntegration(t, ctx, pool, repository, reservation.OperationID, effect, point, boundAt)
		receipt, finalizeErr := provider.Finalize(ctx, FinalizeRequest{
			OperationID: reservation.OperationID, EffectDigest: effect, DBSystemID: point.SystemID,
			DBTimeline: point.Timeline, RequiredLSN: point.RequiredLSN,
		})
		if finalizeErr != nil {
			t.Fatal(finalizeErr)
		}
		activateAuthorityIntegration(t, ctx, pool, repository, receipt, boundAt.Add(time.Second))
		return receipt
	}

	receiptA := commit("30000000-0000-4000-8000-000000000001", EffectDesiredActivate, ScopeNode, nodeA, 1)
	assertAuthorityCheckpoint(t, ctx, repository, 43, nodeA, receiptA.Sequence, receiptA.ReceiptDigest)
	assertAuthorityCheckpoint(t, ctx, repository, 43, nodeB, 0, contracts.Digest{})

	pendingA := reserveAuthorityIntegration(t, provider, "31000000-0000-4000-8000-000000000001", EffectRecoveryActivate, ScopeNode, nodeA)
	recordPendingAuthorityIntegration(t, ctx, pool, repository, pendingA, now.Add(2*time.Minute))
	assertAuthorityCheckpoint(t, ctx, repository, 43, nodeA, receiptA.Sequence, receiptA.ReceiptDigest)
	pendingEffect := contracts.Digest(sha256.Sum256([]byte("task-6-checkpoint-pending-effect")))
	pendingPoint, err := repository.CaptureDatabasePoint(ctx)
	if err != nil {
		t.Fatal(err)
	}
	bindAuthorityIntegration(t, ctx, pool, repository, pendingA.OperationID, pendingEffect, pendingPoint, now.Add(3*time.Minute))
	assertAuthorityCheckpoint(t, ctx, repository, 43, nodeA, receiptA.Sequence, receiptA.ReceiptDigest)

	abortedA := reserveAuthorityIntegration(t, provider, "31000000-0000-4000-8000-000000000002", EffectOperatorTransition, ScopeNode, nodeA)
	recordPendingAuthorityIntegration(t, ctx, pool, repository, abortedA, now.Add(4*time.Minute))
	abortedReceipt, err := provider.Abort(ctx, AbortRequest{OperationID: abortedA.OperationID, Reason: AbortSuperseded})
	if err != nil {
		t.Fatal(err)
	}
	abortAuthorityIntegration(t, ctx, pool, repository, abortedReceipt, now.Add(5*time.Minute))
	assertAuthorityCheckpoint(t, ctx, repository, 43, nodeA, receiptA.Sequence, receiptA.ReceiptDigest)

	receiptB := commit("30000000-0000-4000-8000-000000000002", EffectRecoveryActivate, ScopeNode, nodeB, 2)
	assertAuthorityCheckpoint(t, ctx, repository, 43, nodeA, receiptA.Sequence, receiptA.ReceiptDigest)
	assertAuthorityCheckpoint(t, ctx, repository, 43, nodeB, receiptB.Sequence, receiptB.ReceiptDigest)

	_ = commit("30000000-0000-4000-8000-000000000003", EffectOperatorAuthorizerChange, ScopeGlobalOperatorTrust, globalOperator, 3)
	assertAuthorityCheckpoint(t, ctx, repository, 43, nodeA, receiptA.Sequence, receiptA.ReceiptDigest)
	assertAuthorityCheckpoint(t, ctx, repository, 43, nodeB, receiptB.Sequence, receiptB.ReceiptDigest)

	globalReceipt := commit("30000000-0000-4000-8000-000000000004", EffectRootPublish, ScopeGlobalNodeTrust, globalNode, 4)
	assertAuthorityCheckpoint(t, ctx, repository, 43, nodeA, globalReceipt.Sequence, globalReceipt.ReceiptDigest)
	assertAuthorityCheckpoint(t, ctx, repository, 43, nodeB, globalReceipt.Sequence, globalReceipt.ReceiptDigest)
	assertAuthorityCheckpoint(t, ctx, repository, 44, nodeA, 0, contracts.Digest{})
}

func TestPostgresRepositoryCanonicalizesPostgresTimestampRetries(t *testing.T) {
	ctx, pool := openMigratedAuthorityDatabase(t)
	repository, err := NewPostgresRepository(pool)
	if err != nil {
		t.Fatal(err)
	}
	provider, err := NewDeterministicProvider(49)
	if err != nil {
		t.Fatal(err)
	}
	nodeScope := contracts.Digest(sha256.Sum256([]byte("task-6-timestamp-node")))
	reservedAt := time.Date(2026, time.August, 24, 12, 0, 0, 123456789, time.FixedZone("task-6", 4*60*60))
	boundAt := reservedAt.Add(time.Hour + 111*time.Nanosecond)
	terminalAt := boundAt.Add(time.Minute + 222*time.Nanosecond)

	aborted := reserveAuthorityIntegration(t, provider, "50000000-0000-4000-8000-000000000001", EffectRecoveryActivate, ScopeNode, nodeScope)
	recordPendingAuthorityIntegration(t, ctx, pool, repository, aborted, reservedAt)
	recordPendingAuthorityIntegration(t, ctx, pool, repository, aborted, reservedAt)
	effect := contracts.Digest(sha256.Sum256([]byte("task-6-timestamp-abort-effect")))
	point, err := repository.CaptureDatabasePoint(ctx)
	if err != nil {
		t.Fatal(err)
	}
	bindAuthorityIntegration(t, ctx, pool, repository, aborted.OperationID, effect, point, boundAt)
	bindAuthorityIntegration(t, ctx, pool, repository, aborted.OperationID, effect, point, boundAt)
	pending, err := repository.ListPending(ctx, 49)
	if err != nil || len(pending) != 1 || !pending[0].ReservedAt.Equal(reservedAt.UTC().Truncate(time.Microsecond)) ||
		pending[0].EffectBoundAt == nil || !pending[0].EffectBoundAt.Equal(boundAt.UTC().Truncate(time.Microsecond)) {
		t.Fatalf("canonical pending timestamp row = %#v, %v", pending, err)
	}
	abortedReceipt, err := provider.Abort(ctx, AbortRequest{OperationID: aborted.OperationID, Reason: AbortProviderDependencyFailed})
	if err != nil {
		t.Fatal(err)
	}
	abortAuthorityIntegration(t, ctx, pool, repository, abortedReceipt, terminalAt)
	abortAuthorityIntegration(t, ctx, pool, repository, abortedReceipt, terminalAt)

	committed := reserveAuthorityIntegration(t, provider, "50000000-0000-4000-8000-000000000002", EffectDesiredActivate, ScopeNode, nodeScope)
	recordPendingAuthorityIntegration(t, ctx, pool, repository, committed, reservedAt.Add(2*time.Hour))
	committedEffect := contracts.Digest(sha256.Sum256([]byte("task-6-timestamp-commit-effect")))
	committedPoint, err := repository.CaptureDatabasePoint(ctx)
	if err != nil {
		t.Fatal(err)
	}
	committedBoundAt := boundAt.Add(2 * time.Hour)
	bindAuthorityIntegration(t, ctx, pool, repository, committed.OperationID, committedEffect, committedPoint, committedBoundAt)
	committedReceipt, err := provider.Finalize(ctx, FinalizeRequest{
		OperationID: committed.OperationID, EffectDigest: committedEffect, DBSystemID: committedPoint.SystemID,
		DBTimeline: committedPoint.Timeline, RequiredLSN: committedPoint.RequiredLSN,
	})
	if err != nil {
		t.Fatal(err)
	}
	committedTerminalAt := terminalAt.Add(2 * time.Hour)
	activateAuthorityIntegration(t, ctx, pool, repository, committedReceipt, committedTerminalAt)
	activateAuthorityIntegration(t, ctx, pool, repository, committedReceipt, committedTerminalAt)
}

func assertConcurrentAuthoritySequenceForkRejected(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	repository *PostgresRepository,
	scopeDigest contracts.Digest,
	reservedAt time.Time,
) {
	t.Helper()
	left := manualAuthorityReservation(t, "40000000-0000-4000-8000-000000000001", EffectDesiredActivate, ScopeNode, scopeDigest, 37, 5)
	right := manualAuthorityReservation(t, "40000000-0000-4000-8000-000000000002", EffectRecoveryActivate, ScopeNode, scopeDigest, 37, 5)
	start := make(chan struct{})
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for _, reservation := range []Reservation{left, right} {
		reservation := reservation
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			results <- inAuthorityTransaction(ctx, pool, func(tx pgx.Tx) error {
				return repository.RecordPending(ctx, tx, reservation, reservedAt)
			})
		}()
	}
	close(start)
	wait.Wait()
	close(results)
	successes := 0
	conflicts := 0
	for result := range results {
		switch {
		case result == nil:
			successes++
		case errors.Is(result, ErrConflict):
			conflicts++
		default:
			t.Fatalf("same-sequence fork error = %v, want ErrConflict", result)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("same-sequence fork results = success:%d conflict:%d, want 1/1", successes, conflicts)
	}
}

func reserveAuthorityIntegration(
	t *testing.T,
	provider *DeterministicProvider,
	operationID string,
	kind EffectKind,
	scope ScopeKind,
	scopeDigest contracts.Digest,
) Reservation {
	t.Helper()
	reservation, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse(operationID), Kind: kind, ScopeKind: scope, ScopeDigest: scopeDigest,
	})
	if err != nil {
		t.Fatal(err)
	}
	return reservation
}

func manualAuthorityReservation(
	t *testing.T,
	operationID string,
	kind EffectKind,
	scope ScopeKind,
	scopeDigest contracts.Digest,
	epoch uint64,
	sequence uint64,
) Reservation {
	t.Helper()
	reservation := Reservation{
		OperationID: uuid.MustParse(operationID), Kind: kind, ScopeKind: scope, ScopeDigest: scopeDigest,
		Epoch: epoch, Sequence: sequence,
	}
	var err error
	reservation.ReservationDigest, err = reservationDigest(reservation)
	if err != nil || reservation.Validate() != nil {
		t.Fatalf("construct manual reservation: %v", err)
	}
	return reservation
}

func recordPendingAuthorityIntegration(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	repository *PostgresRepository,
	reservation Reservation,
	reservedAt time.Time,
) {
	t.Helper()
	if err := inAuthorityTransaction(ctx, pool, func(tx pgx.Tx) error {
		return repository.RecordPending(ctx, tx, reservation, reservedAt)
	}); err != nil {
		t.Fatal(err)
	}
}

func bindAuthorityIntegration(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	repository *PostgresRepository,
	operationID uuid.UUID,
	effectDigest contracts.Digest,
	point DatabasePoint,
	effectBoundAt time.Time,
) {
	t.Helper()
	if err := inAuthorityTransaction(ctx, pool, func(tx pgx.Tx) error {
		return repository.BindEffect(ctx, tx, operationID, effectDigest, point, effectBoundAt)
	}); err != nil {
		t.Fatal(err)
	}
}

func activateAuthorityIntegration(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	repository *PostgresRepository,
	receipt Receipt,
	terminalAt time.Time,
) {
	t.Helper()
	if err := inAuthorityTransaction(ctx, pool, func(tx pgx.Tx) error {
		return repository.ActivateCommitted(ctx, tx, receipt, terminalAt)
	}); err != nil {
		t.Fatal(err)
	}
}

func abortAuthorityIntegration(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	repository *PostgresRepository,
	receipt Receipt,
	terminalAt time.Time,
) {
	t.Helper()
	if err := inAuthorityTransaction(ctx, pool, func(tx pgx.Tx) error {
		return repository.RecordAborted(ctx, tx, receipt, terminalAt)
	}); err != nil {
		t.Fatal(err)
	}
}

func inAuthorityTransaction(ctx context.Context, pool *pgxpool.Pool, operation func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	if operationErr := operation(tx); operationErr != nil {
		_ = tx.Rollback(ctx)
		return operationErr
	}
	return tx.Commit(ctx)
}

func assertAuthorityCheckpoint(
	t *testing.T,
	ctx context.Context,
	repository *PostgresRepository,
	epoch uint64,
	nodeScopeDigest contracts.Digest,
	wantSequence uint64,
	wantDigest contracts.Digest,
) {
	t.Helper()
	checkpoint, err := repository.CommittedNodeCheckpoint(ctx, epoch, nodeScopeDigest)
	if err != nil || checkpoint.AuthorityEpoch != epoch || checkpoint.Sequence != wantSequence || checkpoint.ReceiptDigest != wantDigest {
		t.Fatalf("checkpoint = %#v, %v; want epoch=%d sequence=%d digest=%x", checkpoint, err, epoch, wantSequence, wantDigest)
	}
}

func receiptStatusPointer(value ReceiptStatus) *ReceiptStatus {
	return &value
}

func openMigratedAuthorityDatabase(t *testing.T) (context.Context, *pgxpool.Pool) {
	t.Helper()
	pool := openOwnedAuthorityDatabase(t)
	raw, err := os.ReadFile("../../../db/migrations/00006_nodecontrol.sql")
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	upMarker := strings.Index(body, "-- +goose Up")
	downMarker := strings.Index(body, "-- +goose Down")
	if upMarker < 0 || downMarker <= upMarker {
		t.Fatal("nodecontrol migration lacks ordered goose sections")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	if _, err := pool.Exec(ctx, body[upMarker+len("-- +goose Up"):downMarker]); err != nil {
		t.Fatal("nodecontrol migration Up:", err)
	}
	var version string
	if err := pool.QueryRow(ctx, `SHOW server_version`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != "18.4" {
		t.Fatalf("PostgreSQL version = %q, want 18.4", version)
	}
	return ctx, pool
}

func openOwnedAuthorityDatabase(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("TALENRO_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TALENRO_DATABASE_URL is required for integration tests")
	}
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	adminConfig, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	admin, err := pgx.ConnectConfig(context.Background(), adminConfig)
	if err != nil {
		t.Fatal(err)
	}
	databaseName := "nodecontrol_task6_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	databaseIdentifier := pgx.Identifier{databaseName}.Sanitize()
	createCtx, createCancel := context.WithTimeout(context.Background(), 20*time.Second)
	if _, err := admin.Exec(createCtx, "CREATE DATABASE "+databaseIdentifier); err != nil {
		createCancel()
		admin.Close(context.Background())
		t.Fatal(err)
	}
	createCancel()
	admin.Close(context.Background())

	parsed.Path = "/" + databaseName
	pool, err := pgxpool.New(context.Background(), parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		cleanup, cleanupErr := pgx.ConnectConfig(context.Background(), adminConfig)
		if cleanupErr != nil {
			t.Error(cleanupErr)
			return
		}
		defer cleanup.Close(context.Background())
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cleanupCancel()
		if _, cleanupErr := cleanup.Exec(cleanupCtx, "DROP DATABASE "+databaseIdentifier); cleanupErr != nil {
			t.Error(cleanupErr)
		}
	})
	return pool
}
