//go:build integration

package authority

import (
	"context"
	"crypto/sha256"
	"errors"
	"math"
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

func TestPostgresRepositoryAbortClaim(t *testing.T) {
	ctx, pool := openAuthorityV7RepositoryDatabase(t, "abort-claim")
	if _, err := pool.Exec(ctx, `ALTER TABLE nodecontrol.control_plane_authority_fences DROP CONSTRAINT ncv7_fence_activation_fk`); err != nil {
		t.Fatal("remove unrelated activation fixture dependency:", err)
	}
	repository, err := NewPostgresRepository(pool)
	if err != nil {
		t.Fatal(err)
	}
	baseTime := time.Date(2026, time.August, 29, 14, 0, 0, 123456000, time.UTC)
	scopeDigest := contracts.Digest(sha256.Sum256([]byte("task8-abort-claim-scope")))
	activationID := uuid.MustParse("72000000-0000-4000-8000-000000000001")

	newClaimFence := func(operationID string, sequence uint64) (Reservation, time.Time) {
		t.Helper()
		reservation := manualAuthorityReservation(t, operationID, EffectDesiredActivate, ScopeNode, scopeDigest, 72, sequence)
		reservedAt := baseTime.Add(time.Duration(sequence) * time.Minute)
		if err := inAuthorityTransaction(ctx, pool, func(tx pgx.Tx) error {
			return repository.RecordPendingClaimV1(ctx, tx, ClaimV1Reservation{Reservation: reservation, ProtocolActivationID: activationID}, reservedAt)
		}); err != nil {
			t.Fatal("record claim-v1 pending fence:", err)
		}
		return reservation, reservedAt
	}
	claim := func(operationID uuid.UUID, reason AbortReason, claimedAt time.Time) error {
		return inAuthorityTransaction(ctx, pool, func(tx pgx.Tx) error {
			return repository.ClaimAbort(ctx, tx, operationID, reason, claimedAt)
		})
	}

	t.Run("first claim exact retry list lock and defensive copies", func(t *testing.T) {
		reservation, reservedAt := newClaimFence("72000000-0000-4000-8000-000000000011", 11)
		claimedAt := reservedAt.Add(time.Second)
		if err := claim(reservation.OperationID, AbortProviderDependencyFailed, claimedAt); err != nil {
			t.Fatal("first durable claim:", err)
		}
		if err := claim(reservation.OperationID, AbortProviderDependencyFailed, claimedAt); err != nil {
			t.Fatal("exact durable claim retry:", err)
		}
		for _, changed := range []struct {
			label     string
			reason    AbortReason
			claimedAt time.Time
		}{
			{"reason", AbortSuperseded, claimedAt},
			{"time", AbortProviderDependencyFailed, claimedAt.Add(time.Microsecond)},
		} {
			if err := claim(reservation.OperationID, changed.reason, changed.claimedAt); !errors.Is(err, ErrConflict) {
				t.Errorf("changed %s claim error = %v, want ErrConflict", changed.label, err)
			}
		}
		stored, err := repository.GetStoredFence(ctx, reservation.OperationID)
		if err != nil || stored.AbortClaim == nil || stored.AbortClaim.Reason != AbortProviderDependencyFailed || !stored.AbortClaim.ClaimedAt.Equal(claimedAt) {
			t.Fatalf("GetStoredFence abort claim = %#v, %v; want exact reason/time", stored.AbortClaim, err)
		}
		stored.AbortClaim.Reason = AbortSuperseded
		stored.AbortClaim.ClaimedAt = claimedAt.Add(24 * time.Hour)
		again, err := repository.GetStoredFence(ctx, reservation.OperationID)
		if err != nil || again.AbortClaim == nil || again.AbortClaim.Reason != AbortProviderDependencyFailed || !again.AbortClaim.ClaimedAt.Equal(claimedAt) {
			t.Fatalf("GetStoredFence aliases caller mutation: %#v, %v", again.AbortClaim, err)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		locked, lockErr := repository.Lock(ctx, tx, reservation.OperationID)
		_ = tx.Rollback(ctx)
		if lockErr != nil || locked.AbortClaim == nil || locked.AbortClaim.Reason != AbortProviderDependencyFailed || !locked.AbortClaim.ClaimedAt.Equal(claimedAt) {
			t.Fatalf("caller-DBTX Lock abort claim = %#v, %v", locked.AbortClaim, lockErr)
		}
		locked.AbortClaim.Reason = AbortSuperseded
		locked.AbortClaim.ClaimedAt = claimedAt.Add(48 * time.Hour)
		recheckTx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		lockedAgain, lockAgainErr := repository.Lock(ctx, recheckTx, reservation.OperationID)
		_ = recheckTx.Rollback(ctx)
		if lockAgainErr != nil || lockedAgain.AbortClaim == nil || lockedAgain.AbortClaim.Reason != AbortProviderDependencyFailed || !lockedAgain.AbortClaim.ClaimedAt.Equal(claimedAt) {
			t.Fatalf("caller-DBTX Lock aliases caller mutation: claim=%#v error=%v", lockedAgain.AbortClaim, lockAgainErr)
		}
		pending, err := repository.ListPending(ctx, 72)
		if err != nil {
			t.Errorf("ListPending after durable claim: %v", err)
		} else {
			var matching *PendingFence
			for index := range pending {
				if pending[index].OperationID == reservation.OperationID {
					matching = &pending[index]
				}
			}
			if matching == nil || matching.AbortClaim == nil || matching.AbortClaim.Reason != AbortProviderDependencyFailed || !matching.AbortClaim.ClaimedAt.Equal(claimedAt) {
				t.Errorf("ListPending claim projection = %#v, want exact unique claim reason/time", matching)
			} else {
				matching.AbortClaim.Reason = AbortSuperseded
				pendingAgain, retryErr := repository.ListPending(ctx, 72)
				if retryErr != nil {
					t.Error(retryErr)
				} else {
					for index := range pendingAgain {
						if pendingAgain[index].OperationID == reservation.OperationID && (pendingAgain[index].AbortClaim == nil || pendingAgain[index].AbortClaim.Reason != AbortProviderDependencyFailed) {
							t.Errorf("ListPending AbortClaim aliases caller mutation: %#v", pendingAgain[index].AbortClaim)
						}
					}
				}
			}
		}
	})

	t.Run("claim time precedes reservation", func(t *testing.T) {
		reservation, reservedAt := newClaimFence("72000000-0000-4000-8000-000000000012", 12)
		err := claim(reservation.OperationID, AbortValidationFailed, reservedAt.Add(-time.Microsecond))
		task8RequireFiniteRepositoryReject(t, "claim before reserved_at", err)
		stored, readErr := repository.GetStoredFence(ctx, reservation.OperationID)
		if readErr != nil || stored.AbortClaim != nil {
			t.Fatalf("rejected early claim changed row: claim=%#v error=%v", stored.AbortClaim, readErr)
		}
	})

	t.Run("legacy bound and terminal fences reject claim", func(t *testing.T) {
		legacy := manualAuthorityReservation(t, "72000000-0000-4000-8000-000000000013", EffectDesiredActivate, ScopeNode, scopeDigest, 72, 13)
		if _, err := pool.Exec(ctx, `
INSERT INTO nodecontrol.control_plane_authority_fences(
 operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,provider_reservation_digest,
 provider_status,visibility_state,reserved_at,authority_protocol_profile)
VALUES($1,$2,$3,$4,$5,$6,$7,'reserved','fence_pending',$8,'legacy_v6')`,
			legacy.OperationID, string(legacy.Kind), string(legacy.ScopeKind), int64(legacy.Epoch), int64(legacy.Sequence),
			legacy.ScopeDigest[:], legacy.ReservationDigest[:], baseTime); err != nil {
			t.Fatal("seed legacy v6 fence:", err)
		}
		if err := claim(legacy.OperationID, AbortValidationFailed, baseTime.Add(time.Second)); !errors.Is(err, ErrConflict) {
			t.Errorf("legacy fence claim error = %v, want ErrConflict", err)
		}

		bound, reservedAt := newClaimFence("72000000-0000-4000-8000-000000000014", 14)
		point, err := repository.CaptureDatabasePoint(ctx)
		if err != nil {
			t.Fatal(err)
		}
		effectDigest := contracts.Digest(sha256.Sum256([]byte("task8-bound-effect")))
		if err := inAuthorityTransaction(ctx, pool, func(tx pgx.Tx) error {
			return repository.BindEffect(ctx, tx, bound.OperationID, effectDigest, point, reservedAt.Add(time.Second))
		}); err != nil {
			t.Fatal(err)
		}
		if err := claim(bound.OperationID, AbortValidationFailed, reservedAt.Add(2*time.Second)); !errors.Is(err, ErrTerminalConflict) {
			t.Errorf("effect-bound claim error = %v, want ErrTerminalConflict", err)
		}
		boundStored, err := repository.GetStoredFence(ctx, bound.OperationID)
		if err != nil || boundStored.Record.BoundEffectDigest == nil || boundStored.Record.BoundDatabasePoint == nil {
			t.Fatalf("read effect-bound stored fence: %#v error=%v", boundStored, err)
		}
		boundStored.Record.BoundEffectDigest[0] ^= 0xff
		boundStored.Record.BoundDatabasePoint.RequiredLSN = "f/ffffffff"
		boundAgain, err := repository.GetStoredFence(ctx, bound.OperationID)
		if err != nil || boundAgain.Record.BoundEffectDigest == nil || *boundAgain.Record.BoundEffectDigest != effectDigest ||
			boundAgain.Record.BoundDatabasePoint == nil || *boundAgain.Record.BoundDatabasePoint != point {
			t.Fatalf("GetStoredFence aliases bound pointers: %#v error=%v", boundAgain.Record, err)
		}

		terminal, terminalReservedAt := newClaimFence("72000000-0000-4000-8000-000000000015", 15)
		claimedAt := terminalReservedAt.Add(2 * time.Second)
		if err := claim(terminal.OperationID, AbortSuperseded, claimedAt); err != nil {
			t.Fatal(err)
		}
		reason := AbortSuperseded
		receipt := Receipt{Reservation: terminal, Status: StatusAborted, AbortReason: &reason}
		receipt.ReceiptDigest, err = receiptDigest(receipt)
		if err != nil || receipt.Validate() != nil {
			t.Fatalf("construct terminal abort receipt: %v", err)
		}
		if err := inAuthorityTransaction(ctx, pool, func(tx pgx.Tx) error {
			return repository.RecordAborted(ctx, tx, receipt, claimedAt.Add(time.Second))
		}); err != nil {
			t.Fatal("record claimed terminal abort:", err)
		}
		if err := claim(terminal.OperationID, AbortSuperseded, claimedAt); !errors.Is(err, ErrTerminalConflict) {
			t.Errorf("terminal fence claim error = %v, want ErrTerminalConflict", err)
		}
	})

	t.Run("terminal time precedes durable claim", func(t *testing.T) {
		reservation, reservedAt := newClaimFence("72000000-0000-4000-8000-000000000016", 16)
		claimedAt := reservedAt.Add(10 * time.Second)
		if err := claim(reservation.OperationID, AbortActivationDeadlineExpired, claimedAt); err != nil {
			t.Fatal(err)
		}
		reason := AbortActivationDeadlineExpired
		receipt := Receipt{Reservation: reservation, Status: StatusAborted, AbortReason: &reason}
		receipt.ReceiptDigest, err = receiptDigest(receipt)
		if err != nil || receipt.Validate() != nil {
			t.Fatal(err)
		}
		err = inAuthorityTransaction(ctx, pool, func(tx pgx.Tx) error {
			return repository.RecordAborted(ctx, tx, receipt, claimedAt.Add(-time.Microsecond))
		})
		task8RequireFiniteRepositoryReject(t, "terminal_at before abort_claimed_at", err)
		stored, readErr := repository.GetStoredFence(ctx, reservation.OperationID)
		if readErr != nil || stored.Record.TerminalReceipt != nil || stored.AbortClaim == nil || !stored.AbortClaim.ClaimedAt.Equal(claimedAt) {
			t.Fatalf("rejected early terminal changed durable claim: %#v error=%v", stored, readErr)
		}
	})
}

func TestFenceFirstWriterAbortRace(t *testing.T) {
	ctx, pool := openAuthorityV7RepositoryDatabase(t, "first-writer-race")
	if _, err := pool.Exec(ctx, `ALTER TABLE nodecontrol.control_plane_authority_fences DROP CONSTRAINT ncv7_fence_activation_fk`); err != nil {
		t.Fatal(err)
	}
	repository, err := NewPostgresRepository(pool)
	if err != nil {
		t.Fatal(err)
	}
	baseTime := time.Date(2026, time.August, 29, 15, 0, 0, 0, time.UTC)
	scopeDigest := contracts.Digest(sha256.Sum256([]byte("task8-first-writer-scope")))
	activationID := uuid.MustParse("73000000-0000-4000-8000-000000000001")
	point, err := repository.CaptureDatabasePoint(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, winner := range []string{"effect", "claim"} {
		winner := winner
		t.Run(winner+" commits before loser release", func(t *testing.T) {
			sequence := uint64(1)
			operationID := "73000000-0000-4000-8000-000000000011"
			if winner == "claim" {
				sequence = 2
				operationID = "73000000-0000-4000-8000-000000000012"
			}
			reservation := manualAuthorityReservation(t, operationID, EffectDesiredActivate, ScopeNode, scopeDigest, 73, sequence)
			if err := inAuthorityTransaction(ctx, pool, func(tx pgx.Tx) error {
				return repository.RecordPendingClaimV1(ctx, tx, ClaimV1Reservation{Reservation: reservation, ProtocolActivationID: activationID}, baseTime)
			}); err != nil {
				t.Fatal(err)
			}
			first, err := pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer first.Rollback(ctx)
			second, err := pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer second.Rollback(ctx)
			var firstPID, secondPID int32
			if err := first.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&firstPID); err != nil {
				t.Fatal(err)
			}
			if err := second.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&secondPID); err != nil {
				t.Fatal(err)
			}
			effectDigest := contracts.Digest(sha256.Sum256([]byte("task8-first-writer-" + winner)))
			claimedAt := baseTime.Add(2 * time.Second)
			if winner == "effect" {
				if err := repository.BindEffect(ctx, first, reservation.OperationID, effectDigest, point, baseTime.Add(time.Second)); err != nil {
					t.Fatal("first effect writer:", err)
				}
			} else if err := repository.ClaimAbort(ctx, first, reservation.OperationID, AbortValidationFailed, claimedAt); err != nil {
				t.Fatal("first claim writer:", err)
			}
			loserResult := make(chan error, 1)
			go func() {
				if winner == "effect" {
					loserResult <- repository.ClaimAbort(ctx, second, reservation.OperationID, AbortValidationFailed, claimedAt)
					return
				}
				loserResult <- repository.BindEffect(ctx, second, reservation.OperationID, effectDigest, point, baseTime.Add(time.Second))
			}()
			task8WaitForRepositoryLockBlock(t, ctx, pool, firstPID, secondPID)
			if err := first.Commit(ctx); err != nil {
				t.Fatal("commit first writer before releasing loser:", err)
			}
			select {
			case loserErr := <-loserResult:
				if !errors.Is(loserErr, ErrTerminalConflict) {
					t.Fatalf("%s-losing writer error = %v, want ErrTerminalConflict", winner, loserErr)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("losing writer remained blocked after winner commit")
			}
			_ = second.Rollback(ctx)
			stored, err := repository.GetStoredFence(ctx, reservation.OperationID)
			if err != nil {
				t.Fatal(err)
			}
			if winner == "effect" {
				if stored.Record.BoundEffectDigest == nil || *stored.Record.BoundEffectDigest != effectDigest || stored.AbortClaim != nil {
					t.Fatalf("effect winner stored state = %#v", stored)
				}
			} else if stored.AbortClaim == nil || stored.AbortClaim.Reason != AbortValidationFailed || stored.Record.BoundEffectDigest != nil {
				t.Fatalf("claim winner stored state = %#v", stored)
			}
		})
	}
}

func task8RequireFiniteRepositoryReject(t *testing.T, label string, err error) {
	t.Helper()
	if !errors.Is(err, ErrInvalidArgument) && !errors.Is(err, ErrConflict) && !errors.Is(err, ErrTerminalConflict) {
		t.Errorf("%s error = %v, want finite ErrInvalidArgument/ErrConflict/ErrTerminalConflict", label, err)
	}
}

func task8WaitForRepositoryLockBlock(t *testing.T, ctx context.Context, pool *pgxpool.Pool, blockerPID, waiterPID int32) {
	t.Helper()
	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var blocked bool
		if err := pool.QueryRow(ctx, `
SELECT wait_event_type='Lock' AND $2=ANY(pg_blocking_pids(pid))
FROM pg_catalog.pg_stat_activity WHERE pid=$1`, waiterPID, blockerPID).Scan(&blocked); err != nil {
			t.Fatal("observe repository race lock barrier:", err)
		}
		if blocked {
			return
		}
		select {
		case <-ticker.C:
		case <-deadline.C:
			t.Fatalf("backend %d did not block on first-writer backend %d", waiterPID, blockerPID)
		}
	}
}

func openAuthorityV7RepositoryDatabase(t *testing.T, label string) (context.Context, *pgxpool.Pool) {
	t.Helper()
	ctx, pool := openMigratedAuthorityDatabase(t)
	asset, err := os.ReadFile("../../../db/migrations/assets/nodecontrol_authority_v7_up.sql")
	if err != nil {
		t.Fatal(err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for index, statement := range strings.Split(string(asset), "-- talenro:statement") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := tx.Exec(ctx, statement); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("apply authority-v7 asset statement %d for %s: %v", index, label, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal("commit authority-v7 test asset:", err)
	}
	return ctx, pool
}

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

	abortReason := AbortProviderDependencyFailed
	abortedForCommitted := Receipt{
		Reservation: committed,
		Status:      StatusAborted,
		AbortReason: &abortReason,
	}
	abortedForCommitted.ReceiptDigest, err = receiptDigest(abortedForCommitted)
	if err != nil || abortedForCommitted.Validate() != nil {
		t.Fatalf("construct abort-after-commit receipt: %v", err)
	}
	abortAfterCommitErr := inAuthorityTransaction(ctx, pool, func(tx pgx.Tx) error {
		return repository.RecordAborted(ctx, tx, abortedForCommitted, abortedAt.Add(time.Second))
	})
	if abortAfterCommitErr != ErrTerminalConflict {
		t.Fatalf("abort after commit error = %v, want exact ErrTerminalConflict", abortAfterCommitErr)
	}
	committedAfterConflict, err := repository.Get(ctx, committed.OperationID)
	if err != nil || committedAfterConflict.TerminalReceipt == nil ||
		!receiptsEqual(*committedAfterConflict.TerminalReceipt, committedReceipt) {
		t.Fatalf("committed row after opposite terminal = %#v, %v", committedAfterConflict, err)
	}

	committedForAborted := Receipt{
		Reservation:   aborted,
		EffectDigest:  &committedEffect,
		DatabasePoint: &committedPoint,
		Status:        StatusCommitted,
	}
	committedForAborted.ReceiptDigest, err = receiptDigest(committedForAborted)
	if err != nil || committedForAborted.Validate() != nil {
		t.Fatalf("construct commit-after-abort receipt: %v", err)
	}
	commitAfterAbortErr := inAuthorityTransaction(ctx, pool, func(tx pgx.Tx) error {
		return repository.ActivateCommitted(ctx, tx, committedForAborted, abortedAt.Add(time.Second))
	})
	if commitAfterAbortErr != ErrTerminalConflict {
		t.Fatalf("commit after abort error = %v, want exact ErrTerminalConflict", commitAfterAbortErr)
	}
	abortedAfterConflict, err := repository.Get(ctx, aborted.OperationID)
	if err != nil || abortedAfterConflict.TerminalReceipt == nil ||
		!receiptsEqual(*abortedAfterConflict.TerminalReceipt, abortedReceipt) {
		t.Fatalf("aborted row after opposite terminal = %#v, %v", abortedAfterConflict, err)
	}

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

func TestPostgresRepositoryNormalizesUnsignedControlFileIdentity(t *testing.T) {
	ctx, pool := openMigratedAuthorityDatabase(t)
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})

	schemaName := "task6_unsigned_identity_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	schemaIdentifier := pgx.Identifier{schemaName}.Sanitize()
	statements := []string{
		"CREATE SCHEMA " + schemaIdentifier,
		"CREATE FUNCTION " + schemaIdentifier + ".pg_control_system() RETURNS TABLE(system_identifier bigint) LANGUAGE sql AS $$ SELECT -1::bigint $$",
		"CREATE FUNCTION " + schemaIdentifier + ".pg_control_checkpoint() RETURNS TABLE(timeline_id integer) LANGUAGE sql AS $$ SELECT -1::integer $$",
		"SET LOCAL search_path TO " + schemaIdentifier + ", pg_catalog",
	}
	for _, statement := range statements {
		if _, err := tx.Exec(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}

	var systemIdentifier int64
	var timelineID int32
	var systemType string
	var timelineType string
	if err := tx.QueryRow(ctx, `
		SELECT system_identifier, timeline_id,
		       pg_typeof(system_identifier)::text, pg_typeof(timeline_id)::text
		FROM pg_control_system(), pg_control_checkpoint()
	`).Scan(&systemIdentifier, &timelineID, &systemType, &timelineType); err != nil {
		t.Fatal(err)
	}
	if systemIdentifier != -1 || timelineID != -1 || systemType != "bigint" || timelineType != "integer" {
		t.Fatalf("shadow control identity = system:%d (%s) timeline:%d (%s)", systemIdentifier, systemType, timelineID, timelineType)
	}

	repository, err := NewPostgresRepository(tx)
	if err != nil {
		t.Fatal(err)
	}
	point, err := repository.CaptureDatabasePoint(ctx)
	if err != nil {
		t.Fatalf("capture unsigned control identity: %v", err)
	}
	canonicalLSN, canonicalErr := canonicalWALPosition(point.RequiredLSN)
	if point.SystemID != math.MaxUint64 || point.Timeline != math.MaxUint32 || canonicalErr != nil || canonicalLSN != point.RequiredLSN {
		t.Fatalf("unsigned control identity point = %#v, canonical LSN error = %v", point, canonicalErr)
	}
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
	const fixtureLock int64 = 0x54414c454e524f37
	if _, err := admin.Exec(context.Background(), `SELECT pg_advisory_lock($1)`, fixtureLock); err != nil {
		admin.Close(context.Background())
		t.Fatal("acquire cluster-wide authority-v7 fixture lock:", err)
	}
	var preexistingRoles int
	if err := admin.QueryRow(context.Background(), `SELECT count(*) FROM pg_catalog.pg_roles WHERE rolname=ANY($1::text[])`, []string{
		"nodecontrol_upgrade_executor", "nodecontrol_migration_downgrader", "nodecontrol_staging_importer",
	}).Scan(&preexistingRoles); err != nil || (preexistingRoles != 0 && preexistingRoles != 3) {
		_, _ = admin.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, fixtureLock)
		admin.Close(context.Background())
		t.Fatalf("authority-v7 fixture started with partial cluster capability roles count=%d error=%v", preexistingRoles, err)
	}
	fixtureOwnsRoles := preexistingRoles == 0
	databaseName := "nodecontrol_task6_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	databaseIdentifier := pgx.Identifier{databaseName}.Sanitize()
	createCtx, createCancel := context.WithTimeout(context.Background(), 20*time.Second)
	if _, err := admin.Exec(createCtx, "CREATE DATABASE "+databaseIdentifier); err != nil {
		createCancel()
		_, _ = admin.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, fixtureLock)
		admin.Close(context.Background())
		t.Fatal(err)
	}
	createCancel()

	parsed.Path = "/" + databaseName
	pool, err := pgxpool.New(context.Background(), parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cleanupCancel()
		if _, cleanupErr := admin.Exec(cleanupCtx, "DROP DATABASE "+databaseIdentifier+" WITH (FORCE)"); cleanupErr != nil {
			t.Error(cleanupErr)
		}
		if fixtureOwnsRoles {
			if _, cleanupErr := admin.Exec(cleanupCtx, `DROP ROLE IF EXISTS nodecontrol_staging_importer,nodecontrol_migration_downgrader,nodecontrol_upgrade_executor`); cleanupErr != nil {
				t.Error("drop exact authority-v7 cluster capability roles:", cleanupErr)
			}
			var remainingRoles int
			if cleanupErr := admin.QueryRow(cleanupCtx, `SELECT count(*) FROM pg_catalog.pg_roles WHERE rolname=ANY($1::text[])`, []string{
				"nodecontrol_upgrade_executor", "nodecontrol_migration_downgrader", "nodecontrol_staging_importer",
			}).Scan(&remainingRoles); cleanupErr != nil || remainingRoles != 0 {
				t.Errorf("authority-v7 fixture capability-role residue count=%d error=%v", remainingRoles, cleanupErr)
			}
		}
		if _, cleanupErr := admin.Exec(cleanupCtx, `SELECT pg_advisory_unlock($1)`, fixtureLock); cleanupErr != nil {
			t.Error("release cluster-wide authority-v7 fixture lock:", cleanupErr)
		}
		_ = admin.Close(context.Background())
	})
	return pool
}
