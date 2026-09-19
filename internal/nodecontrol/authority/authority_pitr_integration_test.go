//go:build integration

package authority

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"talenro.local/platform/internal/store"
	"talenro.local/platform/internal/testinfra"
)

// The runner owns the cluster, backup, credentials, recovery and cleanup. This
// consumer performs only real claim-v1 certificate operations through scoped access.
func TestPITRBeforeRevocationFailsClosed(t *testing.T) {
	controller, err := testinfra.OpenC12AuthorityPITR()
	if err != nil {
		t.Fatal(err)
	}
	provider, err := NewDeterministicProvider(31)
	if err != nil {
		t.Fatal(err)
	}
	operations := pitrOperations()
	activationID := uuid.MustParse("79000000-0000-4000-8000-000000000001")
	clock := coordinatorClock{now: time.Date(2026, 8, 23, 13, 0, 0, 0, time.UTC)}
	var primaryPoint DatabasePoint
	err = controller.WithPrimaryAuthorityAccess(t.Context(), func(access testinfra.C12AuthorityAccess) error {
		ctx := t.Context()
		raw, err := NewPostgresRepository(access)
		if err != nil {
			return err
		}
		primaryPoint, err = raw.CaptureDatabasePoint(ctx)
		if err != nil {
			return err
		}
		tx, err := access.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		if err := pitrSeedClosure(ctx, tx, activationID, 79, time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)); err != nil {
			return err
		}
		for _, op := range operations {
			if err := pitrSeedCertificate(ctx, tx, primaryPoint, op, clock.Now().Add(-time.Minute)); err != nil {
				return err
			}
		}
		return tx.Commit(ctx)
	})
	if err != nil {
		t.Fatal(err)
	}
	// P establishes a real current-epoch prefix; historical fences alone must
	// never be described as Ready against an empty provider.
	receiptP, _, _ := pitrFinalizeObserved(t, controller, provider, operations[0], activationID, clock)
	pitrAssertPrimaryReady(t, controller, provider, operations[0], receiptP, clock)
	backup, err := controller.CreateBaseBackup(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	receiptA, observedA, factsA := pitrFinalizeObserved(t, controller, provider, operations[1], activationID, clock)
	pitrAssertPrimaryReady(t, controller, provider, operations[1], receiptA, clock)
	selection, err := controller.SelectRecoveryCut(backup, observedA)
	if err != nil {
		t.Fatal(err)
	}
	providerA, err := pitrCopyCommittedProvider(t.Context(), provider)
	if err != nil {
		t.Fatal(err)
	}
	headA, err := providerA.Head(t.Context())
	if err != nil || headA.LatestReservedSequence != 2 || headA.LatestCommittedSequence != receiptA.Sequence || headA.LatestCommittedReceiptDigest != receiptA.ReceiptDigest || !reflect.DeepEqual(headA.LatestCommittedDatabasePoint, receiptA.DatabasePoint) {
		t.Fatalf("independent provider A: %#v %v", headA, err)
	}
	receiptB, observedB, factsB := pitrFinalizeObserved(t, controller, provider, operations[2], activationID, clock)
	pitrAssertPrimaryReady(t, controller, provider, operations[2], receiptB, clock)
	if factsA.EndLSN == factsB.EndLSN {
		t.Fatal("B must be later than A")
	}
	headB, err := provider.Head(t.Context())
	if err != nil || headB.LatestReservedSequence != 3 || headB.LatestCommittedSequence != 3 || headB.LatestCommittedReceiptDigest != receiptB.ReceiptDigest {
		t.Fatalf("provider B: %#v %v", headB, err)
	}
	recordsB := provider.Snapshot()
	cut, err := controller.CrashPrimaryAtCut(t.Context(), selection, observedB)
	if err != nil {
		t.Fatal(err)
	}
	if cut.RecoveryTargetLSN != factsA.EndLSN || cut.TerminalCommit != factsB {
		t.Fatalf("B replaced A recovery target: %#v", cut)
	}
	candidate, err := controller.RestoreAtCut(t.Context(), cut)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err = controller.PromoteCandidate(t.Context(), candidate)
	if err != nil {
		t.Fatal(err)
	}
	timeline, err := controller.InspectTimeline(t.Context(), candidate)
	if err != nil || !timeline.Promoted || timeline.TimelineID <= uint64(primaryPoint.Timeline) {
		t.Fatalf("promotion: %#v %v", timeline, err)
	}
	err = controller.WithCandidateAuthorityAccess(t.Context(), candidate, func(access testinfra.C12AuthorityAccess) error {
		ctx := t.Context()
		raw, err := NewPostgresRepository(access)
		if err != nil {
			return err
		}
		repository := &pitrAuthorityRepository{PostgresRepository: raw, access: access}
		resolver := newPostgresCrashEffectResolver(access)
		coordinatorB := mustNewCoordinatorForTest(t, provider, repository, resolver, clock)
		coordinatorA := mustNewCoordinatorForTest(t, providerA, repository, resolver, clock)
		point, err := raw.CaptureDatabasePoint(ctx)
		if err != nil {
			return err
		}
		if point.SystemID != primaryPoint.SystemID || point.Timeline <= primaryPoint.Timeline || uint64(point.Timeline) != timeline.TimelineID {
			t.Fatalf("candidate physical identity: %#v versus %#v", point, primaryPoint)
		}
		// State proof precedes readiness, so a wrapper/SQL permission failure
		// cannot accidentally count as the required business rejection.
		pitrAssertCommittedState(t, access, coordinatorA, operations[1], receiptA)
		beforeB := pitrAssertAbsentState(t, access, operations[2])
		beforeA := pitrMustCertificateSnapshot(t, access, operations[1])
		behind, err := coordinatorB.CheckReady(ctx)
		if err != nil || behind.Ready || behind.Reason != ReadinessDatabaseBehindProvider {
			t.Fatalf("provider B must be exactly behind: %#v %v", behind, err)
		}
		pitrAssertProviderUnchanged(t, provider, headB, recordsB)
		ready, err := coordinatorA.CheckReady(ctx)
		if err != nil || !ready.Ready || ready.Reason != ReadinessReady || ready.DatabaseHead.PendingCount != 0 {
			t.Fatalf("same candidate provider A must really be ready: %#v %v", ready, err)
		}
		pitrAssertProviderUnchanged(t, provider, headB, recordsB)
		pitrAssertProviderUnchanged(t, providerA, headA, recordsB[:2])
		// Use a valid fixed domain UPDATE and arguments. The candidate wrapper,
		// not PostgreSQL read-only mode, must reject before execution.
		tx, err := access.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		_, writeErr := tx.Exec(ctx, pitrCertificateCommitmentUpdateSQL, operations[1].certificateID, receiptA.OperationID, int64(receiptA.Epoch), int64(receiptA.Sequence), beforeA.commitment, (*receiptA.EffectDigest)[:])
		if writeErr != testinfra.C12PITRInvalidHandle {
			t.Fatalf("candidate domain write not rejected by wrapper: %v", writeErr)
		}
		if err := tx.Rollback(ctx); err != nil {
			return err
		}
		afterB := pitrAssertAbsentState(t, access, operations[2])
		afterA := pitrMustCertificateSnapshot(t, access, operations[1])
		if !reflect.DeepEqual(beforeB, afterB) || !reflect.DeepEqual(beforeA, afterA) {
			t.Fatal("rejected DML changed certificate snapshots")
		}
		pitrAssertCommittedState(t, access, coordinatorA, operations[1], receiptA)
		// Test-local gate only; this is not Task10 same-connection serving.
		data, err := pitrReadReadyCertificate(ctx, coordinatorB, access, operations[2].certificateID)
		if err != ErrAuthorityUnavailable || !reflect.DeepEqual(data, pitrCertificateSnapshot{}) {
			t.Fatalf("readiness gate returned data: %#v %v", data, err)
		}
		pitrAssertProviderUnchanged(t, provider, headB, recordsB)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func pitrAssertPrimaryReady(t *testing.T, controller testinfra.C12AuthorityPITRController, provider *DeterministicProvider, op pitrOperationFixture, receipt Receipt, clock coordinatorClock) {
	t.Helper()
	err := controller.WithPrimaryAuthorityAccess(t.Context(), func(access testinfra.C12AuthorityAccess) error {
		raw, err := NewPostgresRepository(access)
		if err != nil {
			return err
		}
		repository := &pitrAuthorityRepository{PostgresRepository: raw, access: access}
		coordinator := mustNewCoordinatorForTest(t, provider, repository, newPostgresCrashEffectResolver(access), clock)
		pitrAssertCommittedState(t, access, coordinator, op, receipt)
		ready, err := coordinator.CheckReady(t.Context())
		if err != nil {
			return err
		}
		if !ready.Ready || ready.Reason != ReadinessReady || ready.DatabaseHead.PendingCount != 0 {
			t.Fatalf("primary not ready: %#v", ready)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func pitrAssertCommittedState(t *testing.T, access testinfra.C12AuthorityAccess, coordinator *Coordinator, op pitrOperationFixture, receipt Receipt) {
	t.Helper()
	ctx := t.Context()
	fence, err := pitrReadOperationSnapshot(ctx, access, op.request.OperationID)
	if err != nil || fence != (pitrOperationSnapshot{"committed", "active", "claim_v1"}) {
		t.Fatalf("committed fence: %#v %v", fence, err)
	}
	certificate := pitrMustCertificateSnapshot(t, access, op)
	if certificate.status != "revoked" || !certificate.operationID.Valid || certificate.operationID.UUID != receipt.OperationID || len(certificate.commitment) == 0 || len(certificate.evidence) == 0 || len(certificate.resolution) == 0 {
		t.Fatalf("incomplete revoked certificate: %#v", certificate)
	}
	commitment, err := ParseAuthorityEffectCommitment(certificate.commitment)
	if err != nil || receipt.EffectDigest == nil || commitment.Digest() != *receipt.EffectDigest {
		t.Fatalf("commitment differs from provider receipt: %v", err)
	}
	aux, err := pitrReadAuxCount(ctx, access, receipt.OperationID)
	if err != nil || aux != 1 {
		t.Fatalf("committed auxiliary count: %d %v", aux, err)
	}
	var audit, outbox int64
	if err := access.QueryRow(ctx, pitrAuditOutboxCountsSQL, receipt.OperationID).Scan(&audit, &outbox); err != nil || audit != 1 || outbox != 1 {
		t.Fatalf("committed audit/outbox: %d/%d %v", audit, outbox, err)
	}
	tx, err := access.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	stored, err := coordinator.repository.Lock(ctx, tx, receipt.OperationID)
	if err != nil || stored.Record.TerminalReceipt == nil || !receiptEquals(*stored.Record.TerminalReceipt, receipt) {
		t.Fatalf("persisted receipt: %#v %v", stored, err)
	}
	if err := coordinator.validatePersistedOutcome(ctx, tx, stored, receipt); err != nil {
		t.Fatal("real persisted validator:", err)
	}
	if !bytes.Equal(stored.PersistedOutcome.CommitmentJCS, certificate.commitment) || !bytes.Equal(stored.PersistedOutcome.EvidenceJCS, certificate.evidence) || !bytes.Equal(stored.PersistedOutcome.ResolutionJCS, certificate.resolution) {
		t.Fatal("snapshot differs from validated original outcome")
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

func pitrMustCertificateSnapshot(t *testing.T, db store.DBTX, op pitrOperationFixture) pitrCertificateSnapshot {
	t.Helper()
	value, err := pitrReadCertificateSnapshot(t.Context(), db, op.certificateID)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func pitrAssertAbsentState(t *testing.T, db store.DBTX, op pitrOperationFixture) pitrCertificateSnapshot {
	t.Helper()
	if _, err := pitrReadOperationSnapshot(t.Context(), db, op.request.OperationID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("B fence must be absent: %v", err)
	}
	certificate := pitrMustCertificateSnapshot(t, db, op)
	if certificate.status != "active" || certificate.operationID.Valid || certificate.commitment != nil || certificate.evidence != nil || certificate.resolution != nil {
		t.Fatalf("B certificate must have SQL NULL revoke fields: %#v", certificate)
	}
	aux, err := pitrReadAuxCount(t.Context(), db, op.request.OperationID)
	if err != nil || aux != 0 {
		t.Fatalf("B auxiliary rows: %d %v", aux, err)
	}
	var audit, outbox int64
	if err := db.QueryRow(t.Context(), pitrAuditOutboxCountsSQL, op.request.OperationID).Scan(&audit, &outbox); err != nil || audit != 0 || outbox != 0 {
		t.Fatalf("B audit/outbox: %d/%d %v", audit, outbox, err)
	}
	return certificate
}

func pitrAssertProviderUnchanged(t *testing.T, provider *DeterministicProvider, expected Head, records []Record) {
	t.Helper()
	head, err := provider.Head(t.Context())
	if err != nil || !reflect.DeepEqual(head, expected) || !reflect.DeepEqual(provider.Snapshot(), records) {
		t.Fatalf("external provider changed: %#v %v", head, err)
	}
}

func pitrReadReadyCertificate(ctx context.Context, coordinator *Coordinator, db store.DBTX, id uuid.UUID) (pitrCertificateSnapshot, error) {
	readiness, err := coordinator.CheckReady(ctx)
	if err != nil {
		return pitrCertificateSnapshot{}, err
	}
	if !readiness.Ready {
		return pitrCertificateSnapshot{}, ErrAuthorityUnavailable
	}
	return pitrReadCertificateSnapshot(ctx, db, id)
}
