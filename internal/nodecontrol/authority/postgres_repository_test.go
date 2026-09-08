package authority

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

func TestPostgresRepositoryAbortedOutcomeAbsence(t *testing.T) {
	committed, completeRow := persistedOutcomeRowFixture(t, "may_apply")
	reason := AbortValidationFailed
	aborted := Receipt{Reservation: committed.Reservation, Status: StatusAborted, AbortReason: &reason}
	aborted.ReceiptDigest, _ = receiptDigest(aborted)
	baseTime := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	for _, route := range []string{"get", "lock"} {
		for _, scenario := range []struct {
			name    string
			receipt Receipt
			row     pgx.Row
			want    error
		}{
			{"aborted absent", aborted, repositoryErrorRow{err: pgx.ErrNoRows}, nil},
			{"aborted all-null row", aborted, repositoryValuesRow{values: persistedOutcomeRowValues(persistedOutcomeDatabaseRow{})}, ErrInjectedFailure},
			{"aborted full row", aborted, repositoryValuesRow{values: persistedOutcomeRowValues(completeRow)}, ErrInjectedFailure},
			{"committed absent", *committed.TerminalReceipt, repositoryErrorRow{err: pgx.ErrNoRows}, ErrInjectedFailure},
			{"aborted SQL error", aborted, repositoryErrorRow{err: errors.New("private SQL failure")}, ErrInjectedFailure},
			{"aborted canceled", aborted, repositoryErrorRow{err: context.Canceled}, ErrCanceled},
		} {
			t.Run(route+"/"+scenario.name, func(t *testing.T) {
				receipt := scenario.receipt
				values := repositoryFenceRowValues(receipt.Reservation, receipt.EffectDigest, receipt.DatabasePoint,
					baseTime, &receipt, baseTime.Add(2*time.Second))
				claimTime := pgtype.Timestamptz{}
				if receipt.Status == StatusAborted {
					claimTime = requiredTimestamp(baseTime.Add(time.Second))
				}
				values = append(values, "claim_v1", claimTime, uuid.NullUUID{UUID: uuid.MustParse("71000000-0000-4000-8000-000000000001"), Valid: true})
				db := &repositoryRecordingDBTX{rowSequence: []pgx.Row{repositoryValuesRow{values: values}, scenario.row}}
				base := &repositoryRecordingDBTX{forbidUse: true}
				if route == "get" {
					base = db
				}
				repository, err := NewPostgresRepository(base)
				if err != nil {
					t.Fatal(err)
				}
				var stored StoredFence
				if route == "get" {
					stored, err = repository.GetStoredFence(t.Context(), receipt.OperationID)
				} else {
					stored, err = repository.Lock(t.Context(), db, receipt.OperationID)
				}
				if !errors.Is(err, scenario.want) {
					t.Fatalf("outcome error=%v, want %v", err, scenario.want)
				}
				if err == nil && (stored.AbortClaim == nil || stored.Record.TerminalReceipt == nil ||
					!receiptEquals(*stored.Record.TerminalReceipt, aborted) || stored.PersistedOutcome != nil) {
					t.Fatal("aborted projection changed or acquired domain proof")
				}
				if db.calls != 2 || (route == "lock" && base.calls != 0) {
					t.Fatal("lookup escaped caller DBTX or omitted domain absence query")
				}
			})
		}
	}
}

func persistedOutcomeRowValues(row persistedOutcomeDatabaseRow) []any {
	v := reflect.ValueOf(row)
	values := make([]any, v.NumField())
	for i := range values {
		values[i] = v.Field(i).Interface()
	}
	return values
}

func TestPersistedOutcomeTimeCheckpointBranches(t *testing.T) {
	for _, scenario := range []string{"may_apply", "deadline_expired", "higher_node", "final"} {
		t.Run(scenario, func(t *testing.T) {
			record, row := persistedOutcomeRowFixture(t, scenario)
			got, err := persistedOutcomeFromDatabaseRow(record, row)
			if err != nil || got == nil {
				t.Fatalf("legal %s outcome rejected: %v", scenario, err)
			}
			if !bytes.Equal(got.CheckpointAnchorJCS, row.CheckpointAnchorJcs) ||
				!got.AttestationExpiresAt.Equal(row.AttestationExpiresAt.Time) ||
				!got.ActivationDeadline.Equal(row.ActivationDeadline.Time) ||
				!receiptEquals(got.Receipt, *record.TerminalReceipt) ||
				!bytes.Equal(got.EvidenceJCS, row.ActivationEvidenceJcs) {
				t.Fatal("loader changed an exact persistence preimage")
			}
			wantIdentity := contracts.Digest{}
			copy(wantIdentity[:], row.ExpectedProviderIdentityDigest)
			if got.ExpectedProviderIdentityDigest != wantIdentity {
				t.Fatal("loader lost expected identity")
			}
		})
	}
	for _, mutation := range []struct {
		name   string
		change func(*persistedOutcomeDatabaseRow)
	}{
		{"missing attestation", func(r *persistedOutcomeDatabaseRow) { r.AttestationExpiresAt = pgtype.Timestamptz{} }},
		{"missing deadline", func(r *persistedOutcomeDatabaseRow) { r.ActivationDeadline = pgtype.Timestamptz{} }},
		{"missing identity", func(r *persistedOutcomeDatabaseRow) { r.ExpectedProviderIdentityDigest = nil }},
		{"zero identity", func(r *persistedOutcomeDatabaseRow) { r.ExpectedProviderIdentityDigest = make([]byte, 32) }},
		{"zero deadline", func(r *persistedOutcomeDatabaseRow) { r.ActivationDeadline.Time = time.Time{} }},
		{"mixed checkpoint and time", func(r *persistedOutcomeDatabaseRow) {
			_, higher := persistedOutcomeRowFixture(t, "higher_node")
			r.CheckpointAnchorJcs, r.CheckpointAnchorDigest = higher.CheckpointAnchorJcs, higher.CheckpointAnchorDigest
		}},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			record, row := persistedOutcomeRowFixture(t, "may_apply")
			mutation.change(&row)
			if got, err := persistedOutcomeFromDatabaseRow(record, row); err != ErrInjectedFailure || got != nil {
				t.Fatalf("invalid branch returned outcome=%t, error=%v; want false, ErrInjectedFailure", got != nil, err)
			}
		})
	}
	for _, field := range []string{"body", "digest"} {
		t.Run("partial checkpoint "+field, func(t *testing.T) {
			record, row := persistedOutcomeRowFixture(t, "higher_node")
			if field == "body" {
				row.CheckpointAnchorJcs = nil
			} else {
				row.CheckpointAnchorDigest = nil
			}
			if got, err := persistedOutcomeFromDatabaseRow(record, row); err != ErrInjectedFailure || got != nil {
				t.Fatalf("partial checkpoint = %v, %v", got, err)
			}
		})
	}
	for _, mutation := range []struct {
		name     string
		scenario string
		change   func(*persistedOutcomeDatabaseRow)
	}{
		{"checkpoint reason must be superseded", "higher_node", func(r *persistedOutcomeDatabaseRow) { r.EffectReason.String = string(EffectReasonFailed) }},
		{"time reason cannot be superseded", "may_apply", func(r *persistedOutcomeDatabaseRow) { r.EffectReason.String = string(EffectReasonSuperseded) }},
		{"empty groups need final reason", "final", func(r *persistedOutcomeDatabaseRow) { r.EffectReason.String = string(EffectReasonNone) }},
		{"expired deadline equals expiry", "deadline_expired", func(r *persistedOutcomeDatabaseRow) { r.ActivationDeadline.Time = r.AttestationExpiresAt.Time }},
		{"expired deadline exceeds expiry", "deadline_expired", func(r *persistedOutcomeDatabaseRow) {
			r.ActivationDeadline.Time = r.AttestationExpiresAt.Time.Add(time.Microsecond)
		}},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			record, row := persistedOutcomeRowFixture(t, mutation.scenario)
			mutation.change(&row)
			if got, err := persistedOutcomeFromDatabaseRow(record, row); err != ErrInjectedFailure || got != nil {
				t.Fatalf("invalid reason/boundary accepted=%t, error=%v", got != nil, err)
			}
		})
	}
}

func persistedOutcomeRowFixture(t *testing.T, scenario string) (Record, persistedOutcomeDatabaseRow) {
	t.Helper()
	input := evidenceInputFixture(t, scenario)
	evidence, _ := freshEvidenceProofFixture(t, scenario)
	resolutionScenario := scenario
	if scenario == "may_apply" {
		resolutionScenario = "applied"
	}
	resolution := resolutionFixture(t, resolutionScenario)
	row := persistedOutcomeDatabaseRow{
		CommitmentJcs: input.Material.Commitment.CanonicalJCS(), CommitmentDigest: digestBytes(input.Material.Commitment.Digest()),
		ProviderHeadJcs: input.ProviderHead.CanonicalJCS(), ProviderHeadDigest: digestBytes(input.ProviderHead.Digest()),
		EffectReason:          pgtype.Text{String: string(input.Material.Reason), Valid: true},
		ActivationEvidenceJcs: evidence.CanonicalJCS(), ActivationEvidenceDigest: digestBytes(evidence.Digest()),
		EffectResolutionJcs: resolution.CanonicalJCS(), EffectResolutionDigest: digestBytes(resolution.Digest()),
	}
	if input.Checkpoint != nil {
		row.CheckpointAnchorJcs, row.CheckpointAnchorDigest = input.Checkpoint.CanonicalJCS(), digestBytes(input.Checkpoint.Digest())
	}
	if input.Material.TrustedTimeKind == TrustedTimeRollbackResistant {
		row.AttestationExpiresAt = requiredTimestamp(input.Material.AttestationExpiresAt)
		row.ActivationDeadline = requiredTimestamp(input.Material.ActivationDeadline)
		row.ExpectedProviderIdentityDigest = digestBytes(input.Material.ExpectedProviderIdentityDigest)
	}
	receipt := cloneReceipt(input.Receipt)
	return Record{Reservation: receipt.Reservation, BoundEffectDigest: cloneDigestPointer(receipt.EffectDigest),
		BoundDatabasePoint: cloneDatabasePointPointer(receipt.DatabasePoint), TerminalReceipt: &receipt}, row
}

func TestPersistedAuthorityArtifactsRejectJCSAndDigestTampering(t *testing.T) {
	fixture := loadLiteralAuthorityEffectFixture(t)
	kinds := map[string]persistedAuthorityArtifactKind{
		"commitment":    persistedCommitmentArtifact,
		"provider_head": persistedProviderHeadArtifact,
		"checkpoint":    persistedCheckpointArtifact,
		"evidence":      persistedEvidenceArtifact,
		"resolution":    persistedResolutionArtifact,
	}
	seen := map[string]bool{}
	for _, vector := range fixture.Vectors {
		kind, ok := kinds[vector.ArtifactKind]
		if !ok || seen[vector.ArtifactKind] {
			continue
		}
		seen[vector.ArtifactKind] = true
		body := []byte(vector.CanonicalJCS)
		digest := mustAuthorityDigest(t, vector.DigestHex)
		t.Run(vector.ArtifactKind, func(t *testing.T) {
			if got, err := exactJCSAndDigest(body, digest[:], kind); err != nil || got != digest {
				t.Fatalf("exact artifact = %x, %v", got, err)
			}
			for _, mutation := range []struct {
				name   string
				body   []byte
				digest []byte
			}{
				{name: "digest", body: body, digest: append([]byte(nil), digest[:]...)},
				{name: "noncanonical_jcs", body: append([]byte(" "), body...), digest: digest[:]},
				{name: "schema", body: bytes.Replace(body, []byte(`"schema_version":"1"`), []byte(`"schema_version":"2"`), 1), digest: digest[:]},
			} {
				if mutation.name == "digest" {
					mutation.digest[0] ^= 1
				}
				if _, err := exactJCSAndDigest(mutation.body, mutation.digest, kind); !errors.Is(err, ErrInjectedFailure) {
					t.Errorf("%s error = %v, want ErrInjectedFailure", mutation.name, err)
				}
			}
		})
	}
	if len(seen) != len(kinds) {
		t.Fatalf("covered kinds = %v, want %v", seen, kinds)
	}
}

func TestPostgresRepositoryRecordsExactFenceStateWithCallerDBTX(t *testing.T) {
	base := &repositoryRecordingDBTX{forbidUse: true}
	tx := &repositoryRecordingDBTX{}
	repository, err := NewPostgresRepository(base)
	if err != nil {
		t.Fatal(err)
	}

	provider, err := NewDeterministicProvider(17)
	if err != nil {
		t.Fatal(err)
	}
	reservation, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		Kind:        EffectDesiredActivate,
		ScopeKind:   ScopeNode,
		ScopeDigest: sha256.Sum256([]byte("repository-node-scope")),
	})
	if err != nil {
		t.Fatal(err)
	}
	reservedAt := time.Date(2026, time.August, 24, 7, 8, 9, 123456000, time.FixedZone("repository-test", 4*60*60))

	if err := repository.RecordPending(t.Context(), tx, reservation, reservedAt); err != nil {
		t.Fatal(err)
	}

	if base.calls != 0 {
		t.Fatalf("constructor DBTX calls = %d, want 0", base.calls)
	}
	if tx.calls != 1 {
		t.Fatalf("supplied DBTX calls = %d, want 1", tx.calls)
	}
	wantArguments := []any{
		reservation.OperationID,
		string(reservation.Kind),
		string(reservation.ScopeKind),
		int64(reservation.Epoch),
		int64(reservation.Sequence),
		reservation.ScopeDigest[:],
		reservation.ReservationDigest[:],
		pgtype.Timestamptz{Time: reservedAt.UTC(), Valid: true},
	}
	if !reflect.DeepEqual(tx.arguments, wantArguments) {
		t.Fatalf("persisted arguments = %#v, want %#v", tx.arguments, wantArguments)
	}
	if tx.beginCalls != 0 || tx.commitCalls != 0 || tx.rollbackCalls != 0 {
		t.Fatalf("transaction lifecycle calls = begin:%d commit:%d rollback:%d, want all zero", tx.beginCalls, tx.commitCalls, tx.rollbackCalls)
	}
}

func TestPostgresRepositoryUsesCallerDBTXForEveryTransactionalWrite(t *testing.T) {
	base := &repositoryRecordingDBTX{forbidUse: true}
	repository, err := NewPostgresRepository(base)
	if err != nil {
		t.Fatal(err)
	}
	provider, err := NewDeterministicProvider(19)
	if err != nil {
		t.Fatal(err)
	}
	nodeDigest := sha256.Sum256([]byte("repository-transaction-node"))
	committedReservation, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		Kind:        EffectCertificateRevoke,
		ScopeKind:   ScopeNode,
		ScopeDigest: nodeDigest,
	})
	if err != nil {
		t.Fatal(err)
	}
	effectDigest := contracts.Digest(sha256.Sum256([]byte("repository-transaction-effect")))
	databasePoint := DatabasePoint{SystemID: 18446744073709551615, Timeline: 4294967295, RequiredLSN: "ABC/123"}
	committedReceipt, err := provider.Finalize(t.Context(), FinalizeRequest{
		OperationID:  committedReservation.OperationID,
		EffectDigest: effectDigest,
		DBSystemID:   databasePoint.SystemID,
		DBTimeline:   databasePoint.Timeline,
		RequiredLSN:  databasePoint.RequiredLSN,
	})
	if err != nil {
		t.Fatal(err)
	}
	abortedReservation, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		Kind:        EffectDesiredActivate,
		ScopeKind:   ScopeNode,
		ScopeDigest: nodeDigest,
	})
	if err != nil {
		t.Fatal(err)
	}
	abortedReceipt, err := provider.Abort(t.Context(), AbortRequest{
		OperationID: abortedReservation.OperationID,
		Reason:      AbortSuperseded,
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.August, 24, 9, 10, 11, 654321000, time.UTC)

	tests := []struct {
		name      string
		rowValues []any
		call      func(*repositoryRecordingDBTX) error
	}{
		{
			name:      "bind effect",
			rowValues: repositoryFenceRowValues(committedReservation, nil, nil, time.Time{}, nil, time.Time{}),
			call: func(tx *repositoryRecordingDBTX) error {
				return repository.BindEffect(t.Context(), tx, committedReservation.OperationID, effectDigest, databasePoint, now)
			},
		},
		{
			name:      "activate committed",
			rowValues: repositoryFenceRowValues(committedReservation, &effectDigest, &databasePoint, now.Add(-time.Minute), nil, time.Time{}),
			call: func(tx *repositoryRecordingDBTX) error {
				return repository.ActivateCommitted(t.Context(), tx, committedReceipt, now)
			},
		},
		{
			name:      "record aborted",
			rowValues: repositoryFenceRowValues(abortedReservation, nil, nil, time.Time{}, nil, time.Time{}),
			call: func(tx *repositoryRecordingDBTX) error {
				return repository.RecordAborted(t.Context(), tx, abortedReceipt, now)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := &repositoryRecordingDBTX{rowValues: test.rowValues}
			if err := test.call(tx); err != nil {
				t.Fatal(err)
			}
			if tx.calls != 2 {
				t.Fatalf("supplied DBTX calls = %d, want 2", tx.calls)
			}
			if tx.beginCalls != 0 || tx.commitCalls != 0 || tx.rollbackCalls != 0 {
				t.Fatalf("transaction lifecycle calls = begin:%d commit:%d rollback:%d, want all zero", tx.beginCalls, tx.commitCalls, tx.rollbackCalls)
			}
		})
	}
	if base.calls != 0 {
		t.Fatalf("constructor DBTX calls = %d, want 0", base.calls)
	}
}

func TestPostgresRepositoryRejectsInvalidBoundaryValuesWithoutDatabaseAccess(t *testing.T) {
	if _, err := NewPostgresRepository(nil); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("nil constructor DBTX error = %v, want ErrInvalidArgument", err)
	}
	var typedNil *repositoryRecordingDBTX
	if _, err := NewPostgresRepository(typedNil); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("typed-nil constructor DBTX error = %v, want ErrInvalidArgument", err)
	}

	base := &repositoryRecordingDBTX{forbidUse: true}
	repository, err := NewPostgresRepository(base)
	if err != nil {
		t.Fatal(err)
	}
	tx := &repositoryRecordingDBTX{forbidUse: true}
	now := time.Date(2026, time.August, 24, 12, 13, 14, 0, time.UTC)
	nodeDigest := sha256.Sum256([]byte("repository-invalid-node"))

	tests := []struct {
		name string
		call func() error
	}{
		{name: "pending nil context", call: func() error { return repository.RecordPending(nil, tx, Reservation{}, now) }},
		{name: "pending nil DBTX", call: func() error { return repository.RecordPending(t.Context(), nil, Reservation{}, now) }},
		{name: "pending invalid reservation", call: func() error { return repository.RecordPending(t.Context(), tx, Reservation{}, now) }},
		{name: "pending zero time", call: func() error { return repository.RecordPending(t.Context(), tx, Reservation{}, time.Time{}) }},
		{name: "bind nil DBTX", call: func() error {
			return repository.BindEffect(t.Context(), nil, uuid.Nil, contracts.Digest{}, DatabasePoint{}, now)
		}},
		{name: "bind invalid values", call: func() error {
			return repository.BindEffect(t.Context(), tx, uuid.Nil, contracts.Digest{}, DatabasePoint{}, now)
		}},
		{name: "activate invalid receipt", call: func() error { return repository.ActivateCommitted(t.Context(), tx, Receipt{}, now) }},
		{name: "abort invalid receipt", call: func() error { return repository.RecordAborted(t.Context(), tx, Receipt{}, now) }},
		{name: "get nil operation", call: func() error { _, err := repository.Get(t.Context(), uuid.Nil); return err }},
		{name: "head nil context", call: func() error { _, err := repository.Head(nil); return err }},
		{name: "pending zero epoch", call: func() error { _, err := repository.ListPending(t.Context(), 0); return err }},
		{name: "checkpoint zero epoch", call: func() error { _, err := repository.CommittedNodeCheckpoint(t.Context(), 0, nodeDigest); return err }},
		{name: "checkpoint zero digest", call: func() error {
			_, err := repository.CommittedNodeCheckpoint(t.Context(), 1, contracts.Digest{})
			return err
		}},
		{name: "capture nil context", call: func() error { _, err := repository.CaptureDatabasePoint(nil); return err }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("error = %v, want ErrInvalidArgument", err)
			}
		})
	}
	if base.calls != 0 || tx.calls != 0 {
		t.Fatalf("database calls = base:%d supplied:%d, want zero", base.calls, tx.calls)
	}
}

func TestPostgresRepositoryPendingAndTerminalRetriesAreExact(t *testing.T) {
	provider, err := NewDeterministicProvider(23)
	if err != nil {
		t.Fatal(err)
	}
	scopeDigest := contracts.Digest(sha256.Sum256([]byte("repository-retry-node")))
	reservation, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse("44444444-4444-4444-8444-444444444444"),
		Kind:        EffectSecurityIncidentOpen,
		ScopeKind:   ScopeNode,
		ScopeDigest: scopeDigest,
	})
	if err != nil {
		t.Fatal(err)
	}
	reservedAt := time.Date(2026, time.August, 24, 8, 0, 0, 0, time.UTC)

	t.Run("exact pending retry", func(t *testing.T) {
		database := &repositoryRecordingDBTX{
			execRowsSet: true,
			rowValues:   repositoryFenceRowValues(reservation, nil, nil, time.Time{}, nil, time.Time{}),
		}
		repository, err := NewPostgresRepository(database)
		if err != nil {
			t.Fatal(err)
		}
		if err := repository.RecordPending(t.Context(), database, reservation, reservedAt); err != nil {
			t.Fatal(err)
		}
		if database.calls != 2 {
			t.Fatalf("database calls = %d, want 2", database.calls)
		}
	})

	t.Run("changed pending retry", func(t *testing.T) {
		changed := reservation
		changed.Sequence++
		changed.ReservationDigest, err = reservationDigest(changed)
		if err != nil {
			t.Fatal(err)
		}
		database := &repositoryRecordingDBTX{
			execRowsSet: true,
			rowValues:   repositoryFenceRowValues(reservation, nil, nil, time.Time{}, nil, time.Time{}),
		}
		repository, err := NewPostgresRepository(database)
		if err != nil {
			t.Fatal(err)
		}
		if err := repository.RecordPending(t.Context(), database, changed, reservedAt); !errors.Is(err, ErrConflict) {
			t.Fatalf("changed retry error = %v, want ErrConflict", err)
		}
		if err := repository.RecordPending(t.Context(), database, reservation, reservedAt.Add(time.Second)); !errors.Is(err, ErrConflict) {
			t.Fatalf("changed time retry error = %v, want ErrConflict", err)
		}
	})

	effectDigest := contracts.Digest(sha256.Sum256([]byte("repository-retry-effect")))
	point := DatabasePoint{SystemID: 18446744073709551615, Timeline: 4294967295, RequiredLSN: "ABC/DEF"}
	receipt, err := provider.Finalize(t.Context(), FinalizeRequest{
		OperationID:  reservation.OperationID,
		EffectDigest: effectDigest,
		DBSystemID:   point.SystemID,
		DBTimeline:   point.Timeline,
		RequiredLSN:  point.RequiredLSN,
	})
	if err != nil {
		t.Fatal(err)
	}
	effectBoundAt := reservedAt.Add(time.Hour)
	terminalAt := effectBoundAt.Add(time.Minute)

	t.Run("pending retry after committed terminal", func(t *testing.T) {
		database := &repositoryRecordingDBTX{
			execRowsSet: true,
			rowValues:   repositoryFenceRowValues(reservation, &effectDigest, &point, effectBoundAt, &receipt, terminalAt),
		}
		repository, err := NewPostgresRepository(database)
		if err != nil {
			t.Fatal(err)
		}
		if err := repository.RecordPending(t.Context(), database, reservation, reservedAt); err != nil {
			t.Fatalf("terminal exact retry error = %v", err)
		}
	})

	t.Run("bind retry is exact", func(t *testing.T) {
		database := &repositoryRecordingDBTX{
			rowValues: repositoryFenceRowValues(reservation, &effectDigest, &point, effectBoundAt, nil, time.Time{}),
		}
		repository, err := NewPostgresRepository(database)
		if err != nil {
			t.Fatal(err)
		}
		if err := repository.BindEffect(t.Context(), database, reservation.OperationID, effectDigest, point, effectBoundAt); err != nil {
			t.Fatal(err)
		}
		changedPoint := point
		changedPoint.RequiredLSN = "ABC/DF0"
		if err := repository.BindEffect(t.Context(), database, reservation.OperationID, effectDigest, changedPoint, effectBoundAt); !errors.Is(err, ErrConflict) {
			t.Fatalf("changed bind error = %v, want ErrConflict", err)
		}
	})

	t.Run("committed retry and opposite terminal", func(t *testing.T) {
		committedDB := &repositoryRecordingDBTX{
			rowValues: repositoryFenceRowValues(reservation, &effectDigest, &point, effectBoundAt, &receipt, terminalAt),
		}
		repository, err := NewPostgresRepository(committedDB)
		if err != nil {
			t.Fatal(err)
		}
		if err := repository.ActivateCommitted(t.Context(), committedDB, receipt, terminalAt); err != nil {
			t.Fatal(err)
		}
		if err := repository.ActivateCommitted(t.Context(), committedDB, receipt, terminalAt.Add(time.Second)); !errors.Is(err, ErrConflict) {
			t.Fatalf("changed terminal time error = %v, want ErrConflict", err)
		}
		abortReason := AbortSuperseded
		abortedForCommitted := Receipt{
			Reservation: receipt.Reservation,
			Status:      StatusAborted,
			AbortReason: &abortReason,
		}
		abortedForCommitted.ReceiptDigest, err = receiptDigest(abortedForCommitted)
		if err != nil || abortedForCommitted.Validate() != nil {
			t.Fatalf("construct same-operation aborted receipt: %v", err)
		}
		callsBeforeOppositeTerminal := committedDB.calls
		if terminalErr := repository.RecordAborted(t.Context(), committedDB, abortedForCommitted, terminalAt); terminalErr != ErrTerminalConflict {
			t.Fatalf("abort after commit error = %v, want exact ErrTerminalConflict", terminalErr)
		}
		if committedDB.calls != callsBeforeOppositeTerminal+1 {
			t.Fatalf("abort after commit database calls = %d, want one locked-row read", committedDB.calls-callsBeforeOppositeTerminal)
		}

		abortProvider, err := NewDeterministicProvider(23)
		if err != nil {
			t.Fatal(err)
		}
		abortReservation, err := abortProvider.Reserve(t.Context(), ReserveRequest{
			OperationID: uuid.MustParse("55555555-5555-4555-8555-555555555555"),
			Kind:        EffectSecurityIncidentResolve,
			ScopeKind:   ScopeNode,
			ScopeDigest: scopeDigest,
		})
		if err != nil {
			t.Fatal(err)
		}
		abortReceipt, err := abortProvider.Abort(t.Context(), AbortRequest{OperationID: abortReservation.OperationID, Reason: AbortSuperseded})
		if err != nil {
			t.Fatal(err)
		}
		abortedDB := &repositoryRecordingDBTX{
			rowValues: repositoryFenceRowValues(abortReservation, nil, nil, time.Time{}, &abortReceipt, terminalAt),
		}
		abortedRepository, err := NewPostgresRepository(abortedDB)
		if err != nil {
			t.Fatal(err)
		}
		if err := abortedRepository.RecordAborted(t.Context(), abortedDB, abortReceipt, terminalAt); err != nil {
			t.Fatal(err)
		}
		oppositeEffect := contracts.Digest(sha256.Sum256([]byte("repository-opposite-terminal-effect")))
		oppositePoint := DatabasePoint{SystemID: 17, Timeline: 19, RequiredLSN: "A/B"}
		committedForAborted := Receipt{
			Reservation:   abortReservation,
			EffectDigest:  &oppositeEffect,
			DatabasePoint: &oppositePoint,
			Status:        StatusCommitted,
		}
		committedForAborted.ReceiptDigest, err = receiptDigest(committedForAborted)
		if err != nil || committedForAborted.Validate() != nil {
			t.Fatalf("construct same-operation committed receipt: %v", err)
		}
		callsBeforeOppositeTerminal = abortedDB.calls
		if terminalErr := abortedRepository.ActivateCommitted(t.Context(), abortedDB, committedForAborted, terminalAt); terminalErr != ErrTerminalConflict {
			t.Fatalf("commit after abort error = %v, want exact ErrTerminalConflict", terminalErr)
		}
		if abortedDB.calls != callsBeforeOppositeTerminal+1 {
			t.Fatalf("commit after abort database calls = %d, want one locked-row read", abortedDB.calls-callsBeforeOppositeTerminal)
		}
	})
}

func TestPostgresRepositoryCancellationWinsNoRowRaces(t *testing.T) {
	provider, err := NewDeterministicProvider(27)
	if err != nil {
		t.Fatal(err)
	}
	scopeDigest := contracts.Digest(sha256.Sum256([]byte("repository-cancellation-node")))
	committedReservation, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse("56565656-5656-4656-8656-565656565656"),
		Kind:        EffectDesiredActivate,
		ScopeKind:   ScopeNode,
		ScopeDigest: scopeDigest,
	})
	if err != nil {
		t.Fatal(err)
	}
	effectDigest := contracts.Digest(sha256.Sum256([]byte("repository-cancellation-effect")))
	point := DatabasePoint{SystemID: 23, Timeline: 29, RequiredLSN: "C/D"}
	committedReceipt, err := provider.Finalize(t.Context(), FinalizeRequest{
		OperationID: committedReservation.OperationID, EffectDigest: effectDigest,
		DBSystemID: point.SystemID, DBTimeline: point.Timeline, RequiredLSN: point.RequiredLSN,
	})
	if err != nil {
		t.Fatal(err)
	}
	abortedReservation, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse("57575757-5757-4757-8757-575757575757"),
		Kind:        EffectRecoveryActivate,
		ScopeKind:   ScopeNode,
		ScopeDigest: scopeDigest,
	})
	if err != nil {
		t.Fatal(err)
	}
	abortedReceipt, err := provider.Abort(t.Context(), AbortRequest{
		OperationID: abortedReservation.OperationID,
		Reason:      AbortProviderDependencyFailed,
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.August, 24, 15, 0, 0, 0, time.UTC)

	t.Run("pending insert no-row race", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		database := &repositoryRecordingDBTX{execRowsSet: true, rowErr: pgx.ErrNoRows, execHook: cancel}
		repository, constructErr := NewPostgresRepository(database)
		if constructErr != nil {
			t.Fatal(constructErr)
		}
		if recordErr := repository.RecordPending(ctx, database, committedReservation, now); recordErr != ErrCanceled {
			t.Fatalf("pending cancellation race error = %v, want exact ErrCanceled", recordErr)
		}
		if database.calls != 1 {
			t.Fatalf("pending cancellation race calls = %d, want insert only", database.calls)
		}
	})

	t.Run("pending lookup no-row race", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		database := &repositoryRecordingDBTX{execRowsSet: true, rowErr: pgx.ErrNoRows, rowScanHook: cancel}
		repository, constructErr := NewPostgresRepository(database)
		if constructErr != nil {
			t.Fatal(constructErr)
		}
		if recordErr := repository.RecordPending(ctx, database, committedReservation, now); recordErr != ErrCanceled {
			t.Fatalf("pending lookup cancellation race error = %v, want exact ErrCanceled", recordErr)
		}
		if database.calls != 2 {
			t.Fatalf("pending lookup cancellation race calls = %d, want insert and locked-row read", database.calls)
		}
	})

	lookupCalls := []struct {
		name string
		call func(context.Context, *PostgresRepository, *repositoryRecordingDBTX) error
	}{
		{name: "bind", call: func(ctx context.Context, repository *PostgresRepository, database *repositoryRecordingDBTX) error {
			return repository.BindEffect(ctx, database, committedReservation.OperationID, effectDigest, point, now)
		}},
		{name: "activate", call: func(ctx context.Context, repository *PostgresRepository, database *repositoryRecordingDBTX) error {
			return repository.ActivateCommitted(ctx, database, committedReceipt, now)
		}},
		{name: "abort", call: func(ctx context.Context, repository *PostgresRepository, database *repositoryRecordingDBTX) error {
			return repository.RecordAborted(ctx, database, abortedReceipt, now)
		}},
		{name: "get", call: func(ctx context.Context, repository *PostgresRepository, _ *repositoryRecordingDBTX) error {
			_, getErr := repository.Get(ctx, committedReservation.OperationID)
			return getErr
		}},
	}
	for _, test := range lookupCalls {
		t.Run(test.name+" lookup race", func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			database := &repositoryRecordingDBTX{rowErr: pgx.ErrNoRows, rowScanHook: cancel}
			repository, constructErr := NewPostgresRepository(database)
			if constructErr != nil {
				t.Fatal(constructErr)
			}
			if lookupErr := test.call(ctx, repository, database); lookupErr != ErrCanceled {
				t.Fatalf("lookup cancellation race error = %v, want exact ErrCanceled", lookupErr)
			}
			if database.calls != 1 {
				t.Fatalf("lookup cancellation race calls = %d, want locked-row read only", database.calls)
			}
		})
	}

	t.Run("checkpoint lookup race", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		database := &repositoryRecordingDBTX{rowErr: pgx.ErrNoRows, rowScanHook: cancel}
		repository, constructErr := NewPostgresRepository(database)
		if constructErr != nil {
			t.Fatal(constructErr)
		}
		checkpoint, checkpointErr := repository.CommittedNodeCheckpoint(ctx, 27, scopeDigest)
		if checkpointErr != ErrCanceled || checkpoint != (NodeCheckpoint{}) {
			t.Fatalf("checkpoint cancellation race = %#v, %v; want zero and exact ErrCanceled", checkpoint, checkpointErr)
		}
		if database.calls != 1 {
			t.Fatalf("checkpoint cancellation race calls = %d, want one", database.calls)
		}
	})

	t.Run("deadline exceeded lookup race", func(t *testing.T) {
		ctx := &repositoryMutableErrorContext{Context: t.Context()}
		database := &repositoryRecordingDBTX{
			rowErr: pgx.ErrNoRows,
			rowScanHook: func() {
				ctx.err = context.DeadlineExceeded
			},
		}
		repository, constructErr := NewPostgresRepository(database)
		if constructErr != nil {
			t.Fatal(constructErr)
		}
		if _, getErr := repository.Get(ctx, committedReservation.OperationID); getErr != ErrCanceled {
			t.Fatalf("deadline race error = %v, want exact ErrCanceled", getErr)
		}
	})

	t.Run("live no-row classifications remain stable", func(t *testing.T) {
		pendingDB := &repositoryRecordingDBTX{execRowsSet: true, rowErr: pgx.ErrNoRows}
		pendingRepository, constructErr := NewPostgresRepository(pendingDB)
		if constructErr != nil {
			t.Fatal(constructErr)
		}
		if pendingErr := pendingRepository.RecordPending(t.Context(), pendingDB, committedReservation, now); pendingErr != ErrConflict {
			t.Fatalf("live pending no-row error = %v, want exact ErrConflict", pendingErr)
		}
		for _, test := range lookupCalls {
			database := &repositoryRecordingDBTX{rowErr: pgx.ErrNoRows}
			repository, repositoryErr := NewPostgresRepository(database)
			if repositoryErr != nil {
				t.Fatal(repositoryErr)
			}
			if lookupErr := test.call(t.Context(), repository, database); lookupErr != ErrNotFound {
				t.Fatalf("live %s no-row error = %v, want exact ErrNotFound", test.name, lookupErr)
			}
		}
		checkpointDB := &repositoryRecordingDBTX{rowErr: pgx.ErrNoRows}
		checkpointRepository, repositoryErr := NewPostgresRepository(checkpointDB)
		if repositoryErr != nil {
			t.Fatal(repositoryErr)
		}
		checkpoint, checkpointErr := checkpointRepository.CommittedNodeCheckpoint(t.Context(), 27, scopeDigest)
		if checkpointErr != nil || checkpoint != (NodeCheckpoint{AuthorityEpoch: 27}) {
			t.Fatalf("live empty checkpoint = %#v, %v", checkpoint, checkpointErr)
		}
	})
}

func TestPostgresRepositoryCanonicalizesDatabaseTimestampsForExactRetries(t *testing.T) {
	provider, err := NewDeterministicProvider(31)
	if err != nil {
		t.Fatal(err)
	}
	scopeDigest := contracts.Digest(sha256.Sum256([]byte("repository-time-node")))
	reservation, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse("45454545-4545-4545-8545-454545454545"),
		Kind:        EffectDesiredActivate,
		ScopeKind:   ScopeNode,
		ScopeDigest: scopeDigest,
	})
	if err != nil {
		t.Fatal(err)
	}
	effectDigest := contracts.Digest(sha256.Sum256([]byte("repository-time-effect")))
	point := DatabasePoint{SystemID: 18446744073709551615, Timeline: 4294967295, RequiredLSN: "ABC/DEF"}
	receipt, err := provider.Finalize(t.Context(), FinalizeRequest{
		OperationID:  reservation.OperationID,
		EffectDigest: effectDigest,
		DBSystemID:   point.SystemID,
		DBTimeline:   point.Timeline,
		RequiredLSN:  point.RequiredLSN,
	})
	if err != nil {
		t.Fatal(err)
	}
	zone := time.FixedZone("repository-test", 4*60*60)
	reservedAt := time.Date(2026, time.August, 24, 12, 0, 0, 123456789, zone)
	effectBoundAt := reservedAt.Add(time.Hour)
	terminalAt := effectBoundAt.Add(time.Minute)
	canonicalReservedAt := reservedAt.UTC().Truncate(time.Microsecond)
	canonicalEffectBoundAt := effectBoundAt.UTC().Truncate(time.Microsecond)
	canonicalTerminalAt := terminalAt.UTC().Truncate(time.Microsecond)

	pendingRow := repositoryFenceRowValues(reservation, nil, nil, time.Time{}, nil, time.Time{})
	pendingRow[15] = pgtype.Timestamptz{Time: canonicalReservedAt, Valid: true}
	pendingDB := &repositoryRecordingDBTX{execRowsSet: true, rowValues: pendingRow}
	repository, err := NewPostgresRepository(pendingDB)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.RecordPending(t.Context(), pendingDB, reservation, reservedAt); err != nil {
		t.Fatalf("nanosecond pending retry = %v", err)
	}

	boundDB := &repositoryRecordingDBTX{
		rowValues: repositoryFenceRowValues(reservation, &effectDigest, &point, canonicalEffectBoundAt, nil, time.Time{}),
	}
	repository, err = NewPostgresRepository(boundDB)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.BindEffect(t.Context(), boundDB, reservation.OperationID, effectDigest, point, effectBoundAt); err != nil {
		t.Fatalf("nanosecond bind retry = %v", err)
	}

	committedDB := &repositoryRecordingDBTX{
		rowValues: repositoryFenceRowValues(reservation, &effectDigest, &point, canonicalEffectBoundAt, &receipt, canonicalTerminalAt),
	}
	repository, err = NewPostgresRepository(committedDB)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.ActivateCommitted(t.Context(), committedDB, receipt, terminalAt); err != nil {
		t.Fatalf("nanosecond commit retry = %v", err)
	}

	abortProvider, err := NewDeterministicProvider(31)
	if err != nil {
		t.Fatal(err)
	}
	abortReservation, err := abortProvider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse("46464646-4646-4646-8646-464646464646"),
		Kind:        EffectRecoveryActivate,
		ScopeKind:   ScopeNode,
		ScopeDigest: scopeDigest,
	})
	if err != nil {
		t.Fatal(err)
	}
	abortReceipt, err := abortProvider.Abort(t.Context(), AbortRequest{
		OperationID: abortReservation.OperationID,
		Reason:      AbortProviderDependencyFailed,
	})
	if err != nil {
		t.Fatal(err)
	}
	abortedDB := &repositoryRecordingDBTX{
		rowValues: repositoryFenceRowValues(abortReservation, nil, nil, time.Time{}, &abortReceipt, canonicalTerminalAt),
	}
	repository, err = NewPostgresRepository(abortedDB)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.RecordAborted(t.Context(), abortedDB, abortReceipt, terminalAt); err != nil {
		t.Fatalf("nanosecond abort retry = %v", err)
	}

	canonicalZero := time.Time{}.Add(999 * time.Nanosecond)
	invalidDB := &repositoryRecordingDBTX{forbidUse: true}
	repository, err = NewPostgresRepository(invalidDB)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		call func() error
	}{
		{name: "pending", call: func() error {
			return repository.RecordPending(t.Context(), invalidDB, reservation, canonicalZero)
		}},
		{name: "bind", call: func() error {
			return repository.BindEffect(t.Context(), invalidDB, reservation.OperationID, effectDigest, point, canonicalZero)
		}},
		{name: "commit", call: func() error {
			return repository.ActivateCommitted(t.Context(), invalidDB, receipt, canonicalZero)
		}},
		{name: "abort", call: func() error {
			return repository.RecordAborted(t.Context(), invalidDB, abortReceipt, canonicalZero)
		}},
	} {
		t.Run("canonical zero "+test.name, func(t *testing.T) {
			if callErr := test.call(); !errors.Is(callErr, ErrInvalidArgument) {
				t.Fatalf("error = %v, want ErrInvalidArgument", callErr)
			}
		})
	}
	if invalidDB.calls != 0 {
		t.Fatalf("canonical-zero database calls = %d, want zero", invalidDB.calls)
	}
}

func TestPostgresRepositorySupportsBoundAndUnboundAbortedReceipts(t *testing.T) {
	provider, err := NewDeterministicProvider(33)
	if err != nil {
		t.Fatal(err)
	}
	scopeDigest := contracts.Digest(sha256.Sum256([]byte("repository-bound-abort-node")))
	reservation, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse("47474747-4747-4747-8747-474747474747"),
		Kind:        EffectRecoveryActivate,
		ScopeKind:   ScopeNode,
		ScopeDigest: scopeDigest,
	})
	if err != nil {
		t.Fatal(err)
	}
	effectDigest := contracts.Digest(sha256.Sum256([]byte("repository-bound-abort-effect")))
	point := DatabasePoint{SystemID: 18446744073709551615, Timeline: 4294967295, RequiredLSN: "ABC/DEF"}
	reason := AbortProviderDependencyFailed
	boundReceipt := Receipt{
		Reservation:   reservation,
		EffectDigest:  &effectDigest,
		DatabasePoint: &point,
		Status:        StatusAborted,
		AbortReason:   &reason,
	}
	boundReceipt.ReceiptDigest, err = receiptDigest(boundReceipt)
	if err != nil || boundReceipt.Validate() != nil {
		t.Fatalf("construct bound abort receipt: %v", err)
	}
	effectBoundAt := time.Date(2026, time.August, 24, 14, 0, 0, 0, time.UTC)
	terminalAt := effectBoundAt.Add(time.Minute)

	t.Run("exact bound abort retry and get", func(t *testing.T) {
		pendingDatabase := &repositoryRecordingDBTX{
			rowValues: repositoryFenceRowValues(reservation, &effectDigest, &point, effectBoundAt, nil, time.Time{}),
		}
		pendingRepository, constructErr := NewPostgresRepository(pendingDatabase)
		if constructErr != nil {
			t.Fatal(constructErr)
		}
		if writeErr := pendingRepository.RecordAborted(t.Context(), pendingDatabase, boundReceipt, terminalAt); writeErr != nil {
			t.Fatalf("initial bound abort write = %v", writeErr)
		}

		database := &repositoryRecordingDBTX{
			rowValues: repositoryFenceRowValues(reservation, &effectDigest, &point, effectBoundAt, &boundReceipt, terminalAt),
		}
		repository, constructErr := NewPostgresRepository(database)
		if constructErr != nil {
			t.Fatal(constructErr)
		}
		if retryErr := repository.RecordAborted(t.Context(), database, boundReceipt, terminalAt); retryErr != nil {
			t.Fatalf("bound abort retry = %v", retryErr)
		}
		record, getErr := repository.Get(t.Context(), reservation.OperationID)
		if getErr != nil || record.TerminalReceipt == nil || !receiptsEqual(*record.TerminalReceipt, boundReceipt) {
			t.Fatalf("bound aborted record = %#v, %v", record, getErr)
		}
	})

	t.Run("mismatched bound abort is conflict", func(t *testing.T) {
		changedEffect := effectDigest
		changedEffect[0] ^= 0xff
		changedReceipt := boundReceipt
		changedReceipt.EffectDigest = &changedEffect
		changedReceipt.ReceiptDigest, err = receiptDigest(changedReceipt)
		if err != nil || changedReceipt.Validate() != nil {
			t.Fatalf("construct changed bound abort receipt: %v", err)
		}
		database := &repositoryRecordingDBTX{
			rowValues: repositoryFenceRowValues(reservation, &effectDigest, &point, effectBoundAt, nil, time.Time{}),
		}
		repository, constructErr := NewPostgresRepository(database)
		if constructErr != nil {
			t.Fatal(constructErr)
		}
		if abortErr := repository.RecordAborted(t.Context(), database, changedReceipt, terminalAt); !errors.Is(abortErr, ErrConflict) {
			t.Fatalf("mismatched bound abort = %v, want ErrConflict", abortErr)
		}
	})

	t.Run("unbound abort remains exact over a bound row", func(t *testing.T) {
		unboundReceipt, abortErr := provider.Abort(t.Context(), AbortRequest{
			OperationID: reservation.OperationID,
			Reason:      AbortSuperseded,
		})
		if abortErr != nil {
			t.Fatal(abortErr)
		}
		database := &repositoryRecordingDBTX{
			rowValues: repositoryFenceRowValues(reservation, &effectDigest, &point, effectBoundAt, &unboundReceipt, terminalAt),
		}
		repository, constructErr := NewPostgresRepository(database)
		if constructErr != nil {
			t.Fatal(constructErr)
		}
		if retryErr := repository.RecordAborted(t.Context(), database, unboundReceipt, terminalAt); retryErr != nil {
			t.Fatalf("unbound abort retry over bound row = %v", retryErr)
		}
		record, getErr := repository.Get(t.Context(), reservation.OperationID)
		if getErr != nil || record.TerminalReceipt == nil || !receiptsEqual(*record.TerminalReceipt, unboundReceipt) ||
			record.BoundEffectDigest == nil || *record.BoundEffectDigest != effectDigest ||
			record.BoundDatabasePoint == nil || *record.BoundDatabasePoint != point {
			t.Fatalf("unbound aborted record over bound row = %#v, %v", record, getErr)
		}
	})
}

func TestPostgresRepositoryRejectsEmptyNullableByteaValues(t *testing.T) {
	provider, err := NewDeterministicProvider(35)
	if err != nil {
		t.Fatal(err)
	}
	reservation, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse("48484848-4848-4848-8848-484848484848"),
		Kind:        EffectDesiredActivate,
		ScopeKind:   ScopeNode,
		ScopeDigest: contracts.Digest(sha256.Sum256([]byte("repository-empty-bytea-node"))),
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name   string
		mutate func([]any)
	}{
		{name: "empty effect digest", mutate: func(row []any) { row[7] = []byte{} }},
		{name: "empty reserved receipt digest", mutate: func(row []any) { row[9] = []byte{} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			row := repositoryFenceRowValues(reservation, nil, nil, time.Time{}, nil, time.Time{})
			test.mutate(row)
			database := &repositoryRecordingDBTX{rowValues: row}
			repository, constructErr := NewPostgresRepository(database)
			if constructErr != nil {
				t.Fatal(constructErr)
			}
			if _, getErr := repository.Get(t.Context(), reservation.OperationID); !errors.Is(getErr, ErrInjectedFailure) {
				t.Fatalf("empty nullable bytea error = %v, want ErrInjectedFailure", getErr)
			}
		})
	}

	emptyHeadDB := &repositoryRecordingDBTX{rowValues: []any{
		int64(0), int64(0), int64(0), int64(0), int64(0), []byte{}, uuid.NullUUID{}, []byte(nil),
		pgtype.Numeric{}, pgtype.Int8{}, nil, false,
	}}
	repository, err := NewPostgresRepository(emptyHeadDB)
	if err != nil {
		t.Fatal(err)
	}
	if _, headErr := repository.Head(t.Context()); !errors.Is(headErr, ErrInjectedFailure) {
		t.Fatalf("empty head with present empty digest error = %v, want ErrInjectedFailure", headErr)
	}
}

func TestPostgresRepositoryReadsExactHeadPendingCheckpointAndDatabasePoint(t *testing.T) {
	provider, err := NewDeterministicProvider(29)
	if err != nil {
		t.Fatal(err)
	}
	scopeDigest := contracts.Digest(sha256.Sum256([]byte("repository-read-node")))
	first, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse("66666666-6666-4666-8666-666666666666"),
		Kind:        EffectGrantCreate,
		ScopeKind:   ScopeNode,
		ScopeDigest: scopeDigest,
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse("77777777-7777-4777-8777-777777777777"),
		Kind:        EffectDesiredActivate,
		ScopeKind:   ScopeNode,
		ScopeDigest: scopeDigest,
	})
	if err != nil {
		t.Fatal(err)
	}
	effectDigest := contracts.Digest(sha256.Sum256([]byte("repository-read-effect")))
	point := DatabasePoint{SystemID: 18446744073709551615, Timeline: 4294967295, RequiredLSN: "1ABC/2DEF"}
	receipt, err := provider.Finalize(t.Context(), FinalizeRequest{
		OperationID: first.OperationID, EffectDigest: effectDigest, DBSystemID: point.SystemID,
		DBTimeline: point.Timeline, RequiredLSN: point.RequiredLSN,
	})
	if err != nil {
		t.Fatal(err)
	}
	reservedAt := time.Date(2026, time.August, 24, 8, 0, 0, 123456000, time.UTC)
	effectBoundAt := reservedAt.Add(time.Minute)
	terminalAt := effectBoundAt.Add(time.Minute)

	t.Run("empty head is exact zero", func(t *testing.T) {
		database := &repositoryRecordingDBTX{rowValues: []any{
			int64(0), int64(0), int64(0), int64(0), int64(0), []byte(nil), uuid.NullUUID{}, []byte(nil),
			pgtype.Numeric{}, pgtype.Int8{}, nil, false,
		}}
		repository, err := NewPostgresRepository(database)
		if err != nil {
			t.Fatal(err)
		}
		head, err := repository.Head(t.Context())
		if err != nil || !reflect.DeepEqual(head, DatabaseHead{}) {
			t.Fatalf("head = %#v, %v; want exact zero", head, err)
		}
	})

	t.Run("latest committed anchors share one row", func(t *testing.T) {
		headValues := []any{
			int64(29), int64(2), int64(2), int64(1), int64(1), append([]byte(nil), second.ReservationDigest[:]...),
			uuid.NullUUID{UUID: first.OperationID, Valid: true}, append([]byte(nil), receipt.ReceiptDigest[:]...),
			pgtype.Numeric{Int: new(big.Int).SetUint64(point.SystemID), Valid: true},
			pgtype.Int8{Int64: int64(point.Timeline), Valid: true}, string(point.RequiredLSN), false,
		}
		database := &repositoryRecordingDBTX{rowValues: headValues}
		repository, err := NewPostgresRepository(database)
		if err != nil {
			t.Fatal(err)
		}
		head, err := repository.Head(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		if head.Epoch != 29 || head.RecordCount != 2 || head.LatestReservedSequence != 2 ||
			head.LatestCommittedSequence != 1 || head.PendingCount != 1 || head.HasSequenceGap ||
			head.LatestReservationDigest != second.ReservationDigest || head.LatestCommittedOperationID != first.OperationID ||
			head.LatestCommittedReceiptDigest != receipt.ReceiptDigest || head.LatestCommittedDatabasePoint == nil ||
			*head.LatestCommittedDatabasePoint != point {
			t.Fatalf("head = %#v", head)
		}
		head.LatestCommittedDatabasePoint.RequiredLSN = "BAD/1"
		again, err := repository.Head(t.Context())
		if err != nil || again.LatestCommittedDatabasePoint == nil || *again.LatestCommittedDatabasePoint != point {
			t.Fatalf("defensive head = %#v, %v", again, err)
		}
	})

	t.Run("pending rows are ascending complete and defensive", func(t *testing.T) {
		pendingRows := [][]any{
			repositoryPendingRowValues(first, nil, nil, time.Time{}, reservedAt),
			repositoryPendingRowValues(second, &effectDigest, &point, effectBoundAt, reservedAt),
		}
		database := &repositoryRecordingDBTX{queryRows: pendingRows}
		repository, err := NewPostgresRepository(database)
		if err != nil {
			t.Fatal(err)
		}
		pending, err := repository.ListPending(t.Context(), 29)
		if err != nil {
			t.Fatal(err)
		}
		if len(pending) != 2 || pending[0].OperationID != first.OperationID || pending[1].OperationID != second.OperationID ||
			pending[0].BoundEffectDigest != nil || pending[0].BoundDatabasePoint != nil || pending[0].EffectBoundAt != nil ||
			pending[1].BoundEffectDigest == nil || *pending[1].BoundEffectDigest != effectDigest ||
			pending[1].BoundDatabasePoint == nil || *pending[1].BoundDatabasePoint != point ||
			pending[1].EffectBoundAt == nil || !pending[1].EffectBoundAt.Equal(effectBoundAt) {
			t.Fatalf("pending = %#v", pending)
		}
		pending[1].BoundEffectDigest[0] ^= 0xff
		pending[1].BoundDatabasePoint.RequiredLSN = "BAD/2"
		*pending[1].EffectBoundAt = pending[1].EffectBoundAt.Add(time.Hour)
		again, err := repository.ListPending(t.Context(), 29)
		if err != nil || again[1].BoundEffectDigest == nil || *again[1].BoundEffectDigest != effectDigest ||
			again[1].BoundDatabasePoint == nil || *again[1].BoundDatabasePoint != point ||
			again[1].EffectBoundAt == nil || !again[1].EffectBoundAt.Equal(effectBoundAt) {
			t.Fatalf("defensive pending = %#v, %v", again, err)
		}
	})

	t.Run("get checkpoint and capture preserve exact values", func(t *testing.T) {
		getDB := &repositoryRecordingDBTX{rowValues: repositoryFenceRowValues(first, &effectDigest, &point, effectBoundAt, &receipt, terminalAt)}
		getRepository, err := NewPostgresRepository(getDB)
		if err != nil {
			t.Fatal(err)
		}
		record, err := getRepository.Get(t.Context(), first.OperationID)
		if err != nil || record.TerminalReceipt == nil || record.BoundDatabasePoint == nil || *record.BoundDatabasePoint != point {
			t.Fatalf("record = %#v, %v", record, err)
		}

		checkpointDB := &repositoryRecordingDBTX{rowValues: []any{int64(29), int64(1), append([]byte(nil), receipt.ReceiptDigest[:]...)}}
		checkpointRepository, err := NewPostgresRepository(checkpointDB)
		if err != nil {
			t.Fatal(err)
		}
		checkpoint, err := checkpointRepository.CommittedNodeCheckpoint(t.Context(), 29, scopeDigest)
		if err != nil || checkpoint != (NodeCheckpoint{AuthorityEpoch: 29, Sequence: 1, ReceiptDigest: receipt.ReceiptDigest}) {
			t.Fatalf("checkpoint = %#v, %v", checkpoint, err)
		}
		emptyDB := &repositoryRecordingDBTX{rowErr: pgx.ErrNoRows}
		emptyRepository, err := NewPostgresRepository(emptyDB)
		if err != nil {
			t.Fatal(err)
		}
		empty, err := emptyRepository.CommittedNodeCheckpoint(t.Context(), 29, scopeDigest)
		if err != nil || empty != (NodeCheckpoint{AuthorityEpoch: 29}) {
			t.Fatalf("empty checkpoint = %#v, %v", empty, err)
		}

		identityDB := &repositoryRecordingDBTX{rowValues: []any{
			pgtype.Numeric{Int: new(big.Int).SetUint64(point.SystemID), Valid: true}, int64(point.Timeline), string(point.RequiredLSN),
		}}
		identityRepository, err := NewPostgresRepository(identityDB)
		if err != nil {
			t.Fatal(err)
		}
		captured, err := identityRepository.CaptureDatabasePoint(t.Context())
		if err != nil || captured != point {
			t.Fatalf("database point = %#v, %v", captured, err)
		}
	})
}

func TestPostgresRepositorySanitizesDatabaseAndMalformedRowErrors(t *testing.T) {
	operationID := uuid.MustParse("88888888-8888-4888-8888-888888888888")
	databaseFailure := errors.New("SELECT secret_provider_receipt FROM private_table")
	database := &repositoryRecordingDBTX{rowErr: databaseFailure}
	repository, err := NewPostgresRepository(database)
	if err != nil {
		t.Fatal(err)
	}
	_, err = repository.Get(t.Context(), operationID)
	if !errors.Is(err, ErrInjectedFailure) || err.Error() != ErrInjectedFailure.Error() {
		t.Fatalf("database error = %q, want exact sanitized sentinel", err)
	}

	malformed := &repositoryRecordingDBTX{rowValues: []any{
		int64(1), int64(1), int64(1), int64(0), int64(0), make([]byte, 31), uuid.NullUUID{}, []byte(nil),
		pgtype.Numeric{}, pgtype.Int8{}, nil, false,
	}}
	malformedRepository, err := NewPostgresRepository(malformed)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := malformedRepository.Head(t.Context()); !errors.Is(err, ErrInjectedFailure) {
		t.Fatalf("malformed head error = %v, want ErrInjectedFailure", err)
	}
}

type repositoryRecordingDBTX struct {
	rowSequence   []pgx.Row
	forbidUse     bool
	calls         int
	arguments     []any
	beginCalls    int
	commitCalls   int
	rollbackCalls int
	rowValues     []any
	rowErr        error
	queryRows     [][]any
	queryErr      error
	execRows      int64
	execRowsSet   bool
	execErr       error
	execHook      func()
	rowScanHook   func()
}

func (db *repositoryRecordingDBTX) Exec(_ context.Context, _ string, arguments ...any) (pgconn.CommandTag, error) {
	db.calls++
	if db.execHook != nil {
		db.execHook()
	}
	if db.forbidUse {
		return pgconn.CommandTag{}, errors.New("constructor DBTX used")
	}
	db.arguments = append([]any(nil), arguments...)
	if db.execErr != nil {
		return pgconn.CommandTag{}, db.execErr
	}
	if db.execRowsSet {
		return pgconn.NewCommandTag(fmt.Sprintf("UPDATE %d", db.execRows)), nil
	}
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

func (db *repositoryRecordingDBTX) Query(context.Context, string, ...any) (pgx.Rows, error) {
	db.calls++
	if db.queryErr != nil {
		return nil, db.queryErr
	}
	if db.queryRows != nil {
		return &repositoryRows{rows: db.queryRows}, nil
	}
	return nil, errors.New("unexpected Query call")
}

func (db *repositoryRecordingDBTX) QueryRow(context.Context, string, ...any) pgx.Row {
	db.calls++
	if len(db.rowSequence) > 0 {
		row := db.rowSequence[0]
		db.rowSequence = db.rowSequence[1:]
		return row
	}
	if db.rowErr != nil {
		return repositoryErrorRow{err: db.rowErr, beforeScan: db.rowScanHook}
	}
	if db.rowValues != nil {
		return repositoryValuesRow{values: db.rowValues}
	}
	return repositoryErrorRow{err: errors.New("unexpected QueryRow call")}
}

type repositoryRows struct {
	rows   [][]any
	index  int
	closed bool
	err    error
}

func (rows *repositoryRows) Close() {
	rows.closed = true
}

func (rows *repositoryRows) Err() error {
	return rows.err
}

func (rows *repositoryRows) CommandTag() pgconn.CommandTag {
	return pgconn.NewCommandTag(fmt.Sprintf("SELECT %d", len(rows.rows)))
}

func (*repositoryRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (rows *repositoryRows) Next() bool {
	return !rows.closed && rows.index < len(rows.rows)
}

func (rows *repositoryRows) Scan(destinations ...any) error {
	if rows.closed || rows.index >= len(rows.rows) {
		return errors.New("scan without current row")
	}
	err := (repositoryValuesRow{values: rows.rows[rows.index]}).Scan(destinations...)
	rows.index++
	return err
}

func (rows *repositoryRows) Values() ([]any, error) {
	if rows.closed || rows.index >= len(rows.rows) {
		return nil, errors.New("values without current row")
	}
	return append([]any(nil), rows.rows[rows.index]...), nil
}

func (*repositoryRows) RawValues() [][]byte {
	return nil
}

func (*repositoryRows) Conn() *pgx.Conn {
	return nil
}

func (db *repositoryRecordingDBTX) Begin(context.Context) error {
	db.beginCalls++
	return errors.New("unexpected Begin call")
}

func (db *repositoryRecordingDBTX) Commit(context.Context) error {
	db.commitCalls++
	return errors.New("unexpected Commit call")
}

func (db *repositoryRecordingDBTX) Rollback(context.Context) error {
	db.rollbackCalls++
	return errors.New("unexpected Rollback call")
}

type repositoryErrorRow struct {
	err        error
	beforeScan func()
}

func (row repositoryErrorRow) Scan(...any) error {
	if row.beforeScan != nil {
		row.beforeScan()
	}
	return row.err
}

type repositoryMutableErrorContext struct {
	context.Context
	err error
}

func (ctx *repositoryMutableErrorContext) Err() error {
	return ctx.err
}

type repositoryValuesRow struct {
	values []any
}

func (row repositoryValuesRow) Scan(destinations ...any) error {
	if len(destinations) != len(row.values) {
		return fmt.Errorf("scan destinations = %d, want %d", len(destinations), len(row.values))
	}
	for index, value := range row.values {
		if value == nil {
			continue
		}
		destination := reflect.ValueOf(destinations[index])
		if destination.Kind() != reflect.Pointer || destination.IsNil() {
			return fmt.Errorf("scan destination %d is not a non-nil pointer", index)
		}
		target := destination.Elem()
		source := reflect.ValueOf(value)
		if source.Type().AssignableTo(target.Type()) {
			target.Set(source)
			continue
		}
		if source.Type().ConvertibleTo(target.Type()) {
			target.Set(source.Convert(target.Type()))
			continue
		}
		if target.Kind() == reflect.Interface && source.Type().Implements(target.Type()) {
			target.Set(source)
			continue
		}
		return fmt.Errorf("cannot scan %T into %T", value, destinations[index])
	}
	return nil
}

func repositoryFenceRowValues(
	reservation Reservation,
	effectDigest *contracts.Digest,
	databasePoint *DatabasePoint,
	effectBoundAt time.Time,
	terminalReceipt *Receipt,
	terminalAt time.Time,
) []any {
	effectBytes := []byte(nil)
	systemID := pgtype.Numeric{}
	timeline := pgtype.Int8{}
	var requiredLSN any
	effectTime := pgtype.Timestamptz{}
	if effectDigest != nil && databasePoint != nil {
		effectBytes = append([]byte(nil), effectDigest[:]...)
		systemID = pgtype.Numeric{Int: new(big.Int).SetUint64(databasePoint.SystemID), Valid: true}
		timeline = pgtype.Int8{Int64: int64(databasePoint.Timeline), Valid: true}
		requiredLSN = string(databasePoint.RequiredLSN)
		effectTime = pgtype.Timestamptz{Time: effectBoundAt.UTC(), Valid: true}
	}
	providerStatus := "reserved"
	visibilityState := "fence_pending"
	receiptBytes := []byte(nil)
	abortReason := pgtype.Text{}
	terminalTime := pgtype.Timestamptz{}
	if terminalReceipt != nil {
		providerStatus = string(terminalReceipt.Status)
		receiptBytes = append([]byte(nil), terminalReceipt.ReceiptDigest[:]...)
		terminalTime = pgtype.Timestamptz{Time: terminalAt.UTC(), Valid: true}
		if terminalReceipt.Status == StatusCommitted {
			visibilityState = "active"
		} else {
			visibilityState = "aborted"
			abortReason = pgtype.Text{String: string(*terminalReceipt.AbortReason), Valid: true}
		}
	}
	reservedAt := effectBoundAt.Add(-time.Hour)
	if effectBoundAt.IsZero() {
		reservedAt = time.Date(2026, time.August, 24, 8, 0, 0, 0, time.UTC)
	}
	return []any{
		reservation.OperationID,
		string(reservation.Kind),
		string(reservation.ScopeKind),
		int64(reservation.Epoch),
		int64(reservation.Sequence),
		append([]byte(nil), reservation.ScopeDigest[:]...),
		append([]byte(nil), reservation.ReservationDigest[:]...),
		effectBytes,
		providerStatus,
		receiptBytes,
		systemID,
		timeline,
		requiredLSN,
		abortReason,
		visibilityState,
		pgtype.Timestamptz{Time: reservedAt.UTC(), Valid: true},
		effectTime,
		terminalTime,
	}
}

func repositoryPendingRowValues(
	reservation Reservation,
	effectDigest *contracts.Digest,
	databasePoint *DatabasePoint,
	effectBoundAt time.Time,
	reservedAt time.Time,
) []any {
	effectBytes := []byte(nil)
	systemID := pgtype.Numeric{}
	timeline := pgtype.Int8{}
	var requiredLSN any
	effectTime := pgtype.Timestamptz{}
	if effectDigest != nil && databasePoint != nil {
		effectBytes = append([]byte(nil), effectDigest[:]...)
		systemID = pgtype.Numeric{Int: new(big.Int).SetUint64(databasePoint.SystemID), Valid: true}
		timeline = pgtype.Int8{Int64: int64(databasePoint.Timeline), Valid: true}
		requiredLSN = string(databasePoint.RequiredLSN)
		effectTime = pgtype.Timestamptz{Time: effectBoundAt.UTC(), Valid: true}
	}
	return []any{
		reservation.OperationID,
		string(reservation.Kind),
		string(reservation.ScopeKind),
		append([]byte(nil), reservation.ScopeDigest[:]...),
		int64(reservation.Epoch),
		int64(reservation.Sequence),
		append([]byte(nil), reservation.ReservationDigest[:]...),
		effectBytes,
		systemID,
		timeline,
		requiredLSN,
		pgtype.Timestamptz{Time: reservedAt.UTC(), Valid: true},
		effectTime,
	}
}
