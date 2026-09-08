//go:build integration

package authority

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/store"
)

func TestCoordinatorPostgresCrashRecoveryMatrix(t *testing.T) {
	type crashPoint struct {
		name              string
		prepare           func(*testing.T, *postgresCrashFixture)
		wantStatus        ReceiptStatus
		wantEffectCommits int
		boundBeforeCrash  bool
	}

	points := []crashPoint{
		{
			name: "provider reserve committed before response",
			prepare: func(t *testing.T, fixture *postgresCrashFixture) {
				t.Helper()
				if err := fixture.provider.FailNext(OperationReserve, FailureAfterMutationResponseLost); err != nil {
					t.Fatal(err)
				}
				if _, err := fixture.coordinator.Reserve(t.Context(), fixture.request); err != ErrResponseLost {
					t.Fatalf("Reserve error = %v, want ErrResponseLost", err)
				}
			},
			wantStatus: StatusAborted,
		},
		{
			name: "DB pending commit before caller response",
			prepare: func(t *testing.T, fixture *postgresCrashFixture) {
				t.Helper()
				fixture.reserveAndCommitDomain(t, false)
			},
			wantStatus: StatusAborted,
		},
		{
			name: "domain effect transaction committed before database-point capture",
			prepare: func(t *testing.T, fixture *postgresCrashFixture) {
				t.Helper()
				fixture.reserveAndCommitDomain(t, true)
			},
			wantStatus:        StatusCommitted,
			wantEffectCommits: 1,
		},
		{
			name: "database point captured before bind",
			prepare: func(t *testing.T, fixture *postgresCrashFixture) {
				t.Helper()
				fixture.reserveAndCommitDomain(t, true)
				fixture.repository.failTransaction(2, false, ErrInjectedFailure)
				fixture.expectFinalizeError(t, ErrInjectedFailure)
			},
			wantStatus:        StatusCommitted,
			wantEffectCommits: 1,
		},
		{
			name: "DB effect/database bind before provider finalize",
			prepare: func(t *testing.T, fixture *postgresCrashFixture) {
				t.Helper()
				fixture.reserveAndCommitDomain(t, true)
				if err := fixture.provider.FailNext(OperationFinalize, FailureBeforeMutation); err != nil {
					t.Fatal(err)
				}
				fixture.expectFinalizeError(t, ErrInjectedFailure)
			},
			wantStatus:        StatusCommitted,
			wantEffectCommits: 1,
			boundBeforeCrash:  true,
		},
		{
			name: "provider finalize committed before response",
			prepare: func(t *testing.T, fixture *postgresCrashFixture) {
				t.Helper()
				fixture.reserveAndCommitDomain(t, true)
				if err := fixture.provider.FailNext(OperationFinalize, FailureAfterMutationResponseLost); err != nil {
					t.Fatal(err)
				}
				fixture.expectFinalizeError(t, ErrResponseLost)
			},
			wantStatus:        StatusCommitted,
			wantEffectCommits: 1,
			boundBeforeCrash:  true,
		},
		{
			name: "provider finalize response before visibility activation",
			prepare: func(t *testing.T, fixture *postgresCrashFixture) {
				t.Helper()
				fixture.reserveAndCommitDomain(t, true)
				fixture.repository.failTransaction(4, false, ErrInjectedFailure)
				fixture.expectFinalizeError(t, ErrInjectedFailure)
			},
			wantStatus:        StatusCommitted,
			wantEffectCommits: 1,
			boundBeforeCrash:  true,
		},
		{
			name: "visibility activation commit before response",
			prepare: func(t *testing.T, fixture *postgresCrashFixture) {
				t.Helper()
				fixture.reserveAndCommitDomain(t, true)
				fixture.repository.failTransaction(4, true, ErrResponseLost)
				fixture.expectFinalizeError(t, ErrResponseLost)
			},
			wantStatus:        StatusCommitted,
			wantEffectCommits: 1,
			boundBeforeCrash:  true,
		},
	}

	for _, point := range points {
		point := point
		t.Run(point.name, func(t *testing.T) {
			fixture := newPostgresCrashFixture(t)
			point.prepare(t, fixture)
			capturesBeforeRecovery := fixture.repository.captureCount()

			restarted := mustNewCoordinatorForTest(t, fixture.guardedProvider, fixture.repository, fixture.effects, fixture.clock)
			receipt, err := restarted.Recover(t.Context(), fixture.request.OperationID)
			if point.name == "provider reserve committed before response" {
				if err != ErrConflict {
					t.Fatalf("missing whole fence Recover = %v, want fail closed", err)
				}
				if _, getErr := fixture.repository.Get(t.Context(), fixture.request.OperationID); getErr != ErrNotFound {
					t.Fatalf("missing activation provenance manufactured a fence: %v", getErr)
				}
				if fixture.provider.Snapshot()[0].TerminalReceipt != nil {
					t.Fatal("missing fence recovery mutated provider")
				}
				return
			}
			if point.wantStatus == StatusAborted {
				if err != ErrConflict {
					t.Fatalf("reserved absent recovery guessed a reason: %v", err)
				}
				receipt, err = restarted.Abort(t.Context(), AbortRequest{OperationID: fixture.request.OperationID, Reason: AbortValidationFailed})
			}
			if err != nil {
				t.Fatalf("Recover/explicit Abort error = %v", err)
			}
			if point.wantStatus == StatusCommitted {
				fixture.assertAtomicOutcome(t, true)
			}
			if receipt.Validate() != nil || receipt.Sequence != 1 || receipt.Status != point.wantStatus {
				t.Fatalf("terminal receipt = %#v, want valid sequence 1 status %s", receipt, point.wantStatus)
			}
			if point.wantStatus == StatusCommitted && (receipt.DatabasePoint == nil || walPositionLess(receipt.DatabasePoint.RequiredLSN, fixture.effectCommitLSN)) {
				t.Fatalf("required LSN = %#v, want not earlier than domain commit %s", receipt.DatabasePoint, fixture.effectCommitLSN)
			}
			if point.boundBeforeCrash && fixture.repository.captureCount() != capturesBeforeRecovery {
				t.Fatalf("recovery replaced bound database point: captures before=%d after=%d", capturesBeforeRecovery, fixture.repository.captureCount())
			}
			providerRecords := fixture.provider.Snapshot()
			if len(providerRecords) != 1 || providerRecords[0].Sequence != 1 || providerRecords[0].TerminalReceipt == nil ||
				providerRecords[0].TerminalReceipt.ReceiptDigest != receipt.ReceiptDigest {
				t.Fatalf("provider records = %#v, want one exact terminal sequence", providerRecords)
			}
			databaseRecord, getErr := fixture.repository.Get(t.Context(), fixture.request.OperationID)
			if getErr != nil || databaseRecord.TerminalReceipt == nil || databaseRecord.TerminalReceipt.ReceiptDigest != receipt.ReceiptDigest {
				t.Fatalf("database record = %#v, %v; want exact terminal receipt", databaseRecord, getErr)
			}
			var effectRows int
			if err := fixture.pool.QueryRow(t.Context(), `SELECT count(*) FROM nodecontrol.authority_task7_crash_effects WHERE operation_id=$1`, fixture.request.OperationID).Scan(&effectRows); err != nil {
				t.Fatal(err)
			}
			if effectRows != point.wantEffectCommits {
				t.Fatalf("SQL effect row count = %d, want %d", effectRows, point.wantEffectCommits)
			}
			if fixture.guardedProvider.providerCallInTransaction() {
				t.Fatal("coordinator called provider from inside a database transaction")
			}
		})
	}
}

func TestCoordinatorPostgresAbortDomainSafetyAndIdempotence(t *testing.T) {
	t.Run("prepared committed terminal and mismatched domain effects reject without mutation", func(t *testing.T) {
		tests := []struct {
			name   string
			mutate func(*testing.T, *postgresCrashFixture)
		}{
			{
				name: "prepared exact tuple",
				mutate: func(t *testing.T, fixture *postgresCrashFixture) {
					t.Helper()
					if _, err := fixture.pool.Exec(t.Context(), `UPDATE nodecontrol.authority_task7_crash_effects SET effect_state=$2 WHERE operation_id=$1`, fixture.request.OperationID, EffectPrepared); err != nil {
						t.Fatal(err)
					}
				},
			},
			{name: "committed exact tuple"},
			{
				name: "terminal exact tuple",
				mutate: func(t *testing.T, fixture *postgresCrashFixture) {
					t.Helper()
					if _, err := fixture.pool.Exec(t.Context(), `UPDATE nodecontrol.authority_task7_crash_effects SET effect_state=$2 WHERE operation_id=$1`, fixture.request.OperationID, EffectTerminal); err != nil {
						t.Fatal(err)
					}
				},
			},
			{
				name: "committed mismatched tuple",
				mutate: func(t *testing.T, fixture *postgresCrashFixture) {
					t.Helper()
					other := sha256.Sum256([]byte("postgres-abort-other-effect"))
					if _, err := fixture.pool.Exec(t.Context(), `UPDATE nodecontrol.authority_task7_crash_effects SET effect_digest=$2 WHERE operation_id=$1`, fixture.request.OperationID, other[:]); err != nil {
						t.Fatal(err)
					}
				},
			},
		}

		for _, test := range tests {
			test := test
			t.Run(test.name, func(t *testing.T) {
				fixture := newPostgresCrashFixture(t)
				fixture.reserveAndCommitDomain(t, true)
				if test.mutate != nil {
					test.mutate(t, fixture)
				}
				transactionsBefore := fixture.repository.transactionCount()
				abortsBefore := fixture.guardedProvider.abortCount(fixture.request.OperationID)
				if _, err := fixture.coordinator.Abort(t.Context(), AbortRequest{OperationID: fixture.request.OperationID, Reason: AbortSuperseded}); err != ErrConflict {
					t.Fatalf("Abort error = %v, want ErrConflict", err)
				}
				if fixture.repository.transactionCount() != transactionsBefore+1 || fixture.guardedProvider.abortCount(fixture.request.OperationID) != abortsBefore {
					t.Fatalf("transactions/Abort calls = %d/%d, want unchanged %d/%d", fixture.repository.transactionCount(), fixture.guardedProvider.abortCount(fixture.request.OperationID), transactionsBefore, abortsBefore)
				}
				databaseRecord, err := fixture.repository.Get(t.Context(), fixture.request.OperationID)
				if err != nil || databaseRecord.TerminalReceipt != nil {
					t.Fatalf("database record = %#v, %v; want pending", databaseRecord, err)
				}
				providerRecords := fixture.provider.Snapshot()
				if len(providerRecords) != 1 || providerRecords[0].TerminalReceipt != nil {
					t.Fatalf("provider records = %#v, want one pending reservation", providerRecords)
				}
			})
		}
	})

	t.Run("response loss recovers once and exact stored retry is zero write", func(t *testing.T) {
		fixture := newPostgresCrashFixture(t)
		fixture.reserveAndCommitDomain(t, false)
		request := AbortRequest{OperationID: fixture.request.OperationID, Reason: AbortProviderDependencyFailed}
		if err := fixture.provider.FailNext(OperationAbort, FailureAfterMutationResponseLost); err != nil {
			t.Fatal(err)
		}
		if _, err := fixture.coordinator.Abort(t.Context(), request); err != ErrResponseLost {
			t.Fatalf("first Abort error = %v, want ErrResponseLost", err)
		}
		if fixture.repository.transactionCount() != 1 || fixture.guardedProvider.abortCount(fixture.request.OperationID) != 1 {
			t.Fatalf("after response loss transactions/Abort calls = %d/%d, want 1/1 (durable claim before provider)", fixture.repository.transactionCount(), fixture.guardedProvider.abortCount(fixture.request.OperationID))
		}

		restarted := mustNewCoordinatorForTest(t, fixture.guardedProvider, fixture.repository, fixture.effects, coordinatorClock{now: fixture.clock.Now().Add(time.Minute)})
		receipt, err := restarted.Abort(t.Context(), request)
		if err != nil || receipt.Validate() != nil || receipt.Status != StatusAborted {
			t.Fatalf("response-loss retry = %#v, %v", receipt, err)
		}
		var terminalAt time.Time
		if err := fixture.pool.QueryRow(t.Context(), `SELECT terminal_at FROM nodecontrol.control_plane_authority_fences WHERE operation_id=$1`, fixture.request.OperationID).Scan(&terminalAt); err != nil {
			t.Fatal(err)
		}
		transactionsBefore := fixture.repository.transactionCount()
		retry := mustNewCoordinatorForTest(t, fixture.guardedProvider, fixture.repository, fixture.effects, coordinatorClock{now: fixture.clock.Now().Add(2 * time.Minute)})
		retried, err := retry.Abort(t.Context(), request)
		if err != nil || retried.Validate() != nil || retried.ReceiptDigest != receipt.ReceiptDigest {
			t.Fatalf("exact terminal retry = %#v, %v; want receipt %x", retried, err, receipt.ReceiptDigest)
		}
		var terminalAtAfter time.Time
		if err := fixture.pool.QueryRow(t.Context(), `SELECT terminal_at FROM nodecontrol.control_plane_authority_fences WHERE operation_id=$1`, fixture.request.OperationID).Scan(&terminalAtAfter); err != nil {
			t.Fatal(err)
		}
		if fixture.repository.transactionCount() != transactionsBefore+1 || fixture.guardedProvider.abortCount(fixture.request.OperationID) != 1 || !terminalAtAfter.Equal(terminalAt) {
			t.Fatalf("terminal retry transactions=%d want=%d Abort=%d terminal_at=%s want=%s", fixture.repository.transactionCount(), transactionsBefore, fixture.guardedProvider.abortCount(fixture.request.OperationID), terminalAtAfter, terminalAt)
		}
	})
}

func TestCoordinatorPostgresCrashResolverFailsClosedOnSQLDeletionOrCorruption(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, *postgresCrashFixture)
	}{
		{
			name: "deleted committed effect row",
			mutate: func(t *testing.T, fixture *postgresCrashFixture) {
				t.Helper()
				if _, err := fixture.pool.Exec(t.Context(), `DELETE FROM nodecontrol.authority_task7_crash_effects WHERE operation_id=$1`, fixture.request.OperationID); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "corrupt committed effect digest",
			mutate: func(t *testing.T, fixture *postgresCrashFixture) {
				t.Helper()
				corrupt := sha256.Sum256([]byte("postgres-crash-corrupt-effect"))
				if _, err := fixture.pool.Exec(t.Context(), `UPDATE nodecontrol.authority_task7_crash_effects SET effect_digest=$2 WHERE operation_id=$1`, fixture.request.OperationID, corrupt[:]); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "corrupt committed effect kind",
			mutate: func(t *testing.T, fixture *postgresCrashFixture) {
				t.Helper()
				if _, err := fixture.pool.Exec(t.Context(), `UPDATE nodecontrol.authority_task7_crash_effects SET effect_kind=$2 WHERE operation_id=$1`, fixture.request.OperationID, EffectDesiredActivate); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "corrupt committed effect scope tuple",
			mutate: func(t *testing.T, fixture *postgresCrashFixture) {
				t.Helper()
				if _, err := fixture.pool.Exec(t.Context(), `UPDATE nodecontrol.authority_task7_crash_effects SET effect_kind=$2,scope_kind=$3,scope_digest=$4 WHERE operation_id=$1`, fixture.request.OperationID, EffectRootPublish, ScopeGlobalNodeTrust, globalNodeScopeDigest[:]); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "corrupt committed effect scope digest",
			mutate: func(t *testing.T, fixture *postgresCrashFixture) {
				t.Helper()
				corrupt := sha256.Sum256([]byte("postgres-crash-corrupt-scope"))
				if _, err := fixture.pool.Exec(t.Context(), `UPDATE nodecontrol.authority_task7_crash_effects SET scope_digest=$2 WHERE operation_id=$1`, fixture.request.OperationID, corrupt[:]); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "corrupt committed effect state",
			mutate: func(t *testing.T, fixture *postgresCrashFixture) {
				t.Helper()
				if _, err := fixture.pool.Exec(t.Context(), `UPDATE nodecontrol.authority_task7_crash_effects SET effect_state=$2 WHERE operation_id=$1`, fixture.request.OperationID, EffectPrepared); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "duplicate committed effect rows",
			mutate: func(t *testing.T, fixture *postgresCrashFixture) {
				t.Helper()
				if _, err := fixture.pool.Exec(t.Context(), `INSERT INTO nodecontrol.authority_task7_crash_effects SELECT * FROM nodecontrol.authority_task7_crash_effects WHERE operation_id=$1`, fixture.request.OperationID); err != nil {
					t.Fatal(err)
				}
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			fixture := newPostgresCrashFixture(t)
			fixture.reserveAndCommitDomain(t, true)
			if err := fixture.provider.FailNext(OperationFinalize, FailureBeforeMutation); err != nil {
				t.Fatal(err)
			}
			fixture.expectFinalizeError(t, ErrInjectedFailure)
			test.mutate(t, fixture)

			restarted := mustNewCoordinatorForTest(t, fixture.guardedProvider, fixture.repository, fixture.effects, fixture.clock)
			if _, err := restarted.Recover(t.Context(), fixture.request.OperationID); err != ErrConflict {
				t.Fatalf("Recover error = %v, want ErrConflict", err)
			}
			databaseRecord, err := fixture.repository.Get(t.Context(), fixture.request.OperationID)
			if err != nil || databaseRecord.TerminalReceipt != nil {
				t.Fatalf("database record = %#v, %v; want pending after resolver mismatch", databaseRecord, err)
			}
			providerRecords := fixture.provider.Snapshot()
			if len(providerRecords) != 1 || providerRecords[0].TerminalReceipt != nil {
				t.Fatalf("provider records = %#v, want one pending reservation", providerRecords)
			}
			if fixture.guardedProvider.providerCallInTransaction() {
				t.Fatal("coordinator called provider from inside a database transaction")
			}
		})
	}
}

// postgresCrashFixture uses the real v7 fence and certificate proof owner. The
// auxiliary resolver table preserves Task 7 SQL-corruption/cardinality cases;
// it is not an outcome cache and cannot substitute for the real domain proof.
type postgresCrashFixture struct {
	pool            *pgxpool.Pool
	provider        *DeterministicProvider
	guardedProvider *postgresTransactionGuardProvider
	repository      *postgresCrashRepository
	rawRepository   *PostgresRepository
	effects         *postgresCrashEffectResolver
	clock           coordinatorClock
	coordinator     *Coordinator
	request         ReserveRequest
	nodeID          uuid.UUID
	activationID    uuid.UUID
	effectDigest    contracts.Digest
	effectCommitLSN WALPosition
}

func newPostgresCrashFixture(t *testing.T) *postgresCrashFixture {
	t.Helper()
	ctx, pool := openAuthorityV7RepositoryDatabase(t, "coordinator-crash")
	task9InstallCrashAuxiliaryTables(t, ctx, pool)
	activationID := uuid.MustParse("79000000-0000-4000-8000-000000000001")
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	task9SeedProofActivation(t, tx, activationID, 79, time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC))
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	provider, err := NewDeterministicProvider(31)
	if err != nil {
		t.Fatal(err)
	}
	rawRepository, err := NewPostgresRepository(pool)
	if err != nil {
		t.Fatal(err)
	}
	repository := &postgresCrashRepository{PostgresRepository: rawRepository}
	guardedProvider := &postgresTransactionGuardProvider{Provider: provider, repository: repository, aborts: make(map[uuid.UUID]int)}
	effects := newPostgresCrashEffectResolver(pool)
	effects.repository = repository
	clock := coordinatorClock{now: time.Date(2026, 8, 23, 13, 0, 0, 0, time.UTC)}
	nodeID := uuid.New()
	scope := sha256.Sum256(append([]byte("TALENRO-NODE-AUTHORITY-SCOPE-V1\x00"), nodeID[:]...))
	fixture := &postgresCrashFixture{pool: pool, provider: provider, guardedProvider: guardedProvider, repository: repository, rawRepository: rawRepository, effects: effects, clock: clock, nodeID: nodeID, activationID: activationID,
		request: ReserveRequest{OperationID: uuid.New(), Kind: EffectCertificateRevoke, ScopeKind: ScopeNode, ScopeDigest: scope}}
	fixture.coordinator = mustNewCoordinatorForTest(t, guardedProvider, repository, effects, clock)
	return fixture
}

func task9InstallCrashAuxiliaryTables(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	if _, err := pool.Exec(ctx, `CREATE TABLE nodecontrol.authority_task7_crash_effects(
 operation_id uuid NOT NULL,effect_kind text NOT NULL,scope_kind text NOT NULL,
 scope_digest bytea NOT NULL CHECK(octet_length(scope_digest)=32),
 effect_digest bytea NOT NULL CHECK(octet_length(effect_digest)=32),effect_state text NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	// Same outbox schema as 00004_trust_delivery.sql. The isolated authority fixture
	// installs 00006 + v7 only, so this prerequisite is explicit and test-local.
	if _, err := pool.Exec(ctx, `CREATE TABLE public.transactional_outbox (
 event_id uuid PRIMARY KEY,event_type text NOT NULL CHECK(event_type ~ '^talenro[.][a-z0-9_.-]{1,127}$'),
 aggregate_type text NOT NULL CHECK(aggregate_type ~ '^[a-z][a-z0-9_]{0,63}$'),aggregate_id uuid NOT NULL,
 aggregate_version bigint NOT NULL CHECK(aggregate_version>0),
 idempotency_key text NOT NULL CHECK(idempotency_key ~ '^[A-Za-z0-9:_-]{1,128}$'),
 payload bytea NOT NULL CHECK(octet_length(payload) BETWEEN 1 AND 262144),occurred_at timestamptz NOT NULL,
 available_at timestamptz NOT NULL,claimed_until timestamptz,attempts integer NOT NULL DEFAULT 0 CHECK(attempts>=0),published_at timestamptz)`); err != nil {
		t.Fatal(err)
	}
}

func (fixture *postgresCrashFixture) reserveAndCommitDomain(t *testing.T, commitEffect bool) {
	t.Helper()
	reservation, err := fixture.coordinator.Reserve(t.Context(), fixture.request)
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := fixture.pool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback(t.Context())
	if err := fixture.rawRepository.RecordPendingClaimV1(t.Context(), transaction, ClaimV1Reservation{Reservation: reservation, ProtocolActivationID: fixture.activationID}, fixture.clock.Now().Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	if commitEffect {
		fixture.seedCertificate(t, transaction, reservation)
		fixture.effectDigest = fixture.commitDomain(t, transaction, reservation)
	}
	if err := transaction.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	if commitEffect {
		point, err := fixture.rawRepository.CaptureDatabasePoint(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		fixture.effectCommitLSN = point.RequiredLSN
	}
}

func (fixture *postgresCrashFixture) seedCertificate(t *testing.T, tx pgx.Tx, reservation Reservation) {
	t.Helper()
	ctx, at := t.Context(), fixture.clock.Now().Add(-time.Minute)
	id := func(label string) uuid.UUID { return uuid.NewSHA1(reservation.OperationID, []byte(label)) }
	// Historical issuance/node dependencies are inert setup, not the operation under
	// test. This matches task8SeedPreparedProofOwner's replica-only dependencies.
	if _, err := tx.Exec(ctx, `SET LOCAL session_replication_role=replica`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO nodecontrol.node_pops(pop_code,iso_country,region,operator_state,created_at,updated_at)
 VALUES('coordinator-crash','US','fixture','enabled',$1,$1) ON CONFLICT DO NOTHING`, at); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO nodecontrol.node_inventory(node_id,pop_code,operator_state,security_state,identity_state,identity_epoch,lineage_id,created_at,updated_at)
 VALUES($1,'coordinator-crash','enabled','normal','active',1,$2,$3,$3)`, fixture.nodeID, id("lineage"), at); err != nil {
		t.Fatal(err)
	}
	point, err := fixture.rawRepository.CaptureDatabasePoint(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// The historical issuance belongs to an earlier epoch and is never a current
	// provider anchor. All values are deterministic fixture data, not signed evidence.
	if _, err := tx.Exec(ctx, `INSERT INTO nodecontrol.control_plane_authority_fences(operation_id,effect_kind,scope_kind,authority_epoch,authority_sequence,scope_digest,provider_reservation_digest,effect_digest,provider_status,provider_receipt_digest,db_system_id,db_timeline,required_lsn,visibility_state,reserved_at,effect_bound_at,terminal_at,authority_protocol_profile)
 VALUES($1,'certificate_activate','node',1,$2,$3,$3,$3,'committed',$3,$4,$5,$6::pg_lsn,'active',$7,$7,$7,'legacy_v6')`, id("issuance-operation"), int64(reservation.Sequence), reservation.ScopeDigest[:], fmt.Sprint(point.SystemID), int64(point.Timeline), string(point.RequiredLSN), at); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO nodecontrol.node_certificate_issuances(
 issuance_id,authority_operation_id,authority_epoch,authority_sequence,node_id,attempt_id,issuance_kind,identity_epoch,lineage_id,
 issuer_id,csr_sha256,public_key_sha256,template_sha256,request_digest,status,serial_bytes,leaf_der,leaf_der_sha256,
 chain_der,chain_der_sha256,not_before,not_after,created_at,updated_at,terminal_at,retention_until)
 VALUES($1,$2,1,$3,$4,$5,'initial',1,$6,'fixture-ca',$7,$7,$7,$7,'active',decode('01','hex'),decode('02','hex'),$7,
 decode('03','hex'),$7,$8,$9,$8,$8,$8,$10)`, id("issuance"), id("issuance-operation"), int64(reservation.Sequence), fixture.nodeID, id("attempt"), id("lineage"), reservation.ScopeDigest[:], at, at.Add(time.Hour), time.Now().UTC().Add(365*24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `SET LOCAL session_replication_role=origin`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO nodecontrol.node_certificates(
 certificate_id,issuance_id,authority_operation_id,authority_epoch,authority_sequence,node_id,identity_epoch,lineage_id,
 issuer_id,serial_bytes,leaf_der,leaf_der_sha256,public_key_sha256,chain_der_sha256,valid_from,valid_until,status,created_at,updated_at,retention_until)
 VALUES($1,$2,$3,1,$4,$5,1,$6,'fixture-ca',decode('01','hex'),decode('02','hex'),$7,$7,$7,$8,$9,'active',$8,$8,$10)`,
		id("certificate"), id("issuance"), id("issuance-operation"), int64(reservation.Sequence), fixture.nodeID, id("lineage"), reservation.ScopeDigest[:], at, at.Add(time.Hour), time.Now().UTC().Add(365*24*time.Hour)); err != nil {
		t.Fatal(err)
	}
}

func task9CertificateInputDigest(operationID, nodeID, certificateID uuid.UUID, leafDigest []byte) contracts.Digest {
	body := append([]byte("TASK9-CERTIFICATE-REVOKE-INPUT-V1\x00"), operationID[:]...)
	body = append(body, nodeID[:]...)
	body = append(body, certificateID[:]...)
	body = append(body, leafDigest...)
	return sha256.Sum256(body)
}

func (fixture *postgresCrashFixture) commitDomain(t *testing.T, tx pgx.Tx, reservation Reservation) contracts.Digest {
	t.Helper()
	// The first-writer contract is explicit: lock the fence before any domain row.
	stored, err := fixture.rawRepository.Lock(t.Context(), tx, reservation.OperationID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.AbortClaim != nil || stored.Record.TerminalReceipt != nil {
		t.Fatal("effect writer lost fence to Abort")
	}
	certificateID := uuid.NewSHA1(reservation.OperationID, []byte("certificate"))
	inputs := task9CertificateInputDigest(reservation.OperationID, fixture.nodeID, certificateID, reservation.ScopeDigest[:])
	commitment, err := NewAuthorityEffectCommitment(AuthorityEffectCommitmentInput{OperationID: reservation.OperationID, Kind: reservation.Kind, ScopeKind: reservation.ScopeKind, ScopeDigest: reservation.ScopeDigest, Epoch: reservation.Epoch, Sequence: reservation.Sequence,
		BaseEffectDigest: sha256.Sum256([]byte("task9-revoke:" + reservation.OperationID.String())), Mode: CommitmentConditionalApply, Reason: EffectReasonNone, ActivationPolicyVersion: 1, ActivationInputsDigest: inputs})
	if err != nil {
		t.Fatal(err)
	}
	digest := commitment.Digest()
	if _, err := tx.Exec(t.Context(), `UPDATE nodecontrol.node_certificates SET revoke_authority_operation_id=$2,revoke_authority_epoch=$3,revoke_authority_sequence=$4,revoke_authority_effect_commitment_jcs=$5,revoke_authority_effect_commitment_digest=$6 WHERE certificate_id=$1`,
		certificateID, reservation.OperationID, int64(reservation.Epoch), int64(reservation.Sequence), commitment.CanonicalJCS(), digest[:]); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(t.Context(), `INSERT INTO nodecontrol.authority_task7_crash_effects(operation_id,effect_kind,scope_kind,scope_digest,effect_digest,effect_state) VALUES($1,$2,$3,$4,$5,$6)`, reservation.OperationID, reservation.Kind, reservation.ScopeKind, reservation.ScopeDigest[:], digest[:], EffectCommitted); err != nil {
		t.Fatal(err)
	}
	return digest
}

type postgresCrashEffectResolver struct {
	pool              *pgxpool.Pool
	repository        *postgresCrashRepository
	mu                sync.Mutex
	captures          int
	validations       int
	activationFailure string
	captureErr        error
	onResolve         func(context.Context, store.DBTX) error
}

func newPostgresCrashEffectResolver(pool *pgxpool.Pool) *postgresCrashEffectResolver {
	return &postgresCrashEffectResolver{pool: pool}
}
func (resolver *postgresCrashEffectResolver) ResolveAuthorityEffect(ctx context.Context, id uuid.UUID) (ResolvedEffect, error) {
	// The obsolete unbound resolver seam must never be reached by Coordinator.
	return ResolvedEffect{}, ErrConflict
}
func (resolver *postgresCrashEffectResolver) ResolveRegisteredAuthorityEffectForUpdate(ctx context.Context, dbtx store.DBTX, query TransactionalEffectQuery) (TransactionalResolvedEffect, error) {
	expected := query.Expected()
	result := TransactionalResolvedEffect{OperationID: expected.OperationID, Epoch: expected.Epoch, Sequence: expected.Sequence, Effect: ResolvedEffect{State: EffectAbsent}}
	if query.RegisteredKind() != expected.Kind {
		return result, nil
	}
	if resolver.onResolve != nil {
		if err := resolver.onResolve(ctx, dbtx); err != nil {
			return result, err
		}
	}
	rows, err := dbtx.Query(ctx, `SELECT effect_kind,scope_kind,scope_digest,effect_digest,effect_state FROM nodecontrol.authority_task7_crash_effects WHERE operation_id=$1 FOR UPDATE`, expected.OperationID)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	if !rows.Next() {
		return result, rows.Err()
	}
	var kind, scope, state string
	var scopeBytes, digestBytes []byte
	if err := rows.Scan(&kind, &scope, &scopeBytes, &digestBytes, &state); err != nil {
		return result, err
	}
	if rows.Next() {
		return result, ErrConflict
	}
	if err := rows.Err(); err != nil {
		return result, err
	}
	rows.Close()
	if len(scopeBytes) != 32 || len(digestBytes) != 32 {
		return result, ErrConflict
	}
	var digest, scopeDigest contracts.Digest
	copy(digest[:], digestBytes)
	copy(scopeDigest[:], scopeBytes)
	result.Effect = ResolvedEffect{Kind: EffectKind(kind), ScopeKind: ScopeKind(scope), ScopeDigest: scopeDigest, EffectDigest: digest, State: EffectState(state)}
	if result.Effect.Validate() != nil {
		return result, ErrConflict
	}
	var terminal bool
	if err := dbtx.QueryRow(ctx, `SELECT revoke_authority_effect_resolution_jcs IS NOT NULL FROM nodecontrol.node_certificates WHERE revoke_authority_operation_id=$1 FOR UPDATE`, expected.OperationID).Scan(&terminal); err != nil {
		return result, ErrConflict
	}
	if terminal {
		result.Effect.State = EffectTerminal
	}
	return result, nil
}

func (resolver *postgresCrashEffectResolver) CaptureActivationDecisionMaterial(ctx context.Context, receipt Receipt) (ActivationDecisionMaterial, error) {
	resolver.mu.Lock()
	resolver.captures++
	captureErr := resolver.captureErr
	resolver.mu.Unlock()
	if captureErr != nil {
		return ActivationDecisionMaterial{}, captureErr
	}
	if resolver.repository != nil && resolver.repository.inTransaction() {
		return ActivationDecisionMaterial{}, ErrConflict
	}
	var body []byte
	if err := resolver.pool.QueryRow(ctx, `SELECT revoke_authority_effect_commitment_jcs FROM nodecontrol.node_certificates WHERE revoke_authority_operation_id=$1`, receipt.OperationID).Scan(&body); err != nil {
		return ActivationDecisionMaterial{}, err
	}
	commitment, err := ParseAuthorityEffectCommitment(body)
	if err != nil {
		return ActivationDecisionMaterial{}, err
	}
	// Choose PostgreSQL-exact fixture input before evidence construction; never
	// truncate an existing proof or a signed/committed observation after capture.
	now := time.Now().UTC().Truncate(time.Microsecond)
	identity := sha256.Sum256([]byte("task9-fixture-trusted-time-provider"))
	return ActivationDecisionMaterial{Commitment: commitment, Reason: EffectReasonNone, CheckpointKind: CheckpointNone, TrustedTimeKind: TrustedTimeRollbackResistant, Capability: DecisionCapabilityMayApply,
		TrustedInstant: now, EvidenceValidUntil: now.Add(4 * time.Second), AttestationExpiresAt: now.Add(10 * time.Second), ActivationDeadline: now.Add(10 * time.Second),
		ProviderIdentityDigest: identity, ExpectedProviderIdentityDigest: identity, FloorAttestationDigest: sha256.Sum256([]byte("task9-fixture-floor"))}, nil
}
func task9ReadCertificateInput(ctx context.Context, dbtx store.DBTX, receipt Receipt) (uuid.UUID, AuthorityEffectCommitment, error) {
	var nodeID, certificateID uuid.UUID
	var body, leaf []byte
	if err := dbtx.QueryRow(ctx, `SELECT node_id,certificate_id,leaf_der_sha256,revoke_authority_effect_commitment_jcs FROM nodecontrol.node_certificates WHERE revoke_authority_operation_id=$1 FOR UPDATE`, receipt.OperationID).Scan(&nodeID, &certificateID, &leaf, &body); err != nil {
		return uuid.Nil, AuthorityEffectCommitment{}, err
	}
	commitment, err := ParseAuthorityEffectCommitment(body)
	if err != nil {
		return uuid.Nil, AuthorityEffectCommitment{}, err
	}
	if commitment.Facts().ActivationInputsDigest != task9CertificateInputDigest(receipt.OperationID, nodeID, certificateID, leaf) {
		return uuid.Nil, AuthorityEffectCommitment{}, ErrConflict
	}
	return nodeID, commitment, nil
}
func (resolver *postgresCrashEffectResolver) ActivateAuthorityEffect(ctx context.Context, dbtx store.DBTX, receipt Receipt, proof ValidatedActivationDecisionEvidence) error {
	if proof.origin != activationEvidenceOriginFresh {
		return ErrConflict
	}
	nodeID, commitment, err := task9ReadCertificateInput(ctx, dbtx, receipt)
	if err != nil {
		return err
	}
	if !bytes.Equal(commitment.CanonicalJCS(), proof.Input().Material.Commitment.CanonicalJCS()) {
		return ErrConflict
	}
	resolution, err := coordinatorTestResolution(proof)
	if err != nil {
		return err
	}
	outcome := coordinatorTestOutcome(proof, resolution)
	at := proof.Input().Material.TrustedInstant
	if _, err := dbtx.Exec(ctx, `UPDATE nodecontrol.node_certificates SET
 status='revoked',revoked_at=$2,revoke_reason='scheduled',updated_at=$2,
 revoke_authority_provider_head_jcs=$3,revoke_authority_provider_head_digest=$4,
 revoke_authority_effect_reason=$5,revoke_authority_attestation_expires_at=$6,revoke_authority_activation_deadline=$7,
 revoke_authority_expected_provider_identity_digest=$8,revoke_authority_activation_evidence_jcs=$9,revoke_authority_activation_evidence_digest=$10,
 revoke_authority_effect_resolution_jcs=$11,revoke_authority_effect_resolution_digest=$12
 WHERE revoke_authority_operation_id=$1`, receipt.OperationID, at, outcome.ProviderHeadJCS, outcome.ProviderHeadDigest[:], outcome.Reason, outcome.AttestationExpiresAt, outcome.ActivationDeadline, outcome.ExpectedProviderIdentityDigest[:], outcome.EvidenceJCS, outcome.EvidenceDigest[:], outcome.ResolutionJCS, outcome.ResolutionDigest[:]); err != nil {
		return err
	}
	if resolver.activationFailure == "domain" {
		return ErrInjectedFailure
	}
	if _, err := dbtx.Exec(ctx, `INSERT INTO nodecontrol.node_operator_audit(audit_id,command_id,authority_operation_id,authority_epoch,authority_sequence,operator_id,credential_digest,role,action,target_kind,target_id,reason,result,occurred_at,retention_until)
 VALUES($1,$1,$1,$2,$3,'task9-fixture',$4,'node_security_admin','disable_node','node',$5,'administrative_disable','accepted',$6,$7)`, receipt.OperationID, int64(receipt.Epoch), int64(receipt.Sequence), receipt.ReceiptDigest[:], nodeID.String(), at, at.Add(181*24*time.Hour)); err != nil {
		return err
	}
	if resolver.activationFailure == "audit" {
		return ErrInjectedFailure
	}
	// Persist the original full context independently of the certificate proof
	// columns. Recovery must not validate a projection against itself.
	originalContext, err := json.Marshal(outcome)
	if err != nil {
		return err
	}
	if _, err := dbtx.Exec(ctx, `INSERT INTO public.transactional_outbox(event_id,event_type,aggregate_type,aggregate_id,aggregate_version,idempotency_key,payload,occurred_at,available_at)
 VALUES($1,'talenro.node.certificate_revoked','node',$2,$3,$4,$5,$6,$6)`, receipt.OperationID, nodeID, int64(receipt.Sequence), receipt.OperationID.String(), originalContext, at); err != nil {
		return err
	}
	if resolver.activationFailure == "outbox" {
		return ErrInjectedFailure
	}
	return nil
}
func (resolver *postgresCrashEffectResolver) ValidatePersistedAuthorityEffect(ctx context.Context, dbtx store.DBTX, receipt Receipt, proof ValidatedActivationDecisionEvidence, resolution AuthorityEffectResolution) error {
	if proof.origin != activationEvidenceOriginParsed {
		return ErrConflict
	}
	resolver.mu.Lock()
	resolver.validations++
	resolver.mu.Unlock()
	nodeID, commitment, err := task9ReadCertificateInput(ctx, dbtx, receipt)
	if err != nil {
		return err
	}
	if !bytes.Equal(commitment.CanonicalJCS(), proof.Input().Material.Commitment.CanonicalJCS()) {
		return ErrConflict
	}
	var status string
	var evidence, resolved []byte
	var audit, outbox int
	if err := dbtx.QueryRow(ctx, `SELECT status,revoke_authority_activation_evidence_jcs,revoke_authority_effect_resolution_jcs FROM nodecontrol.node_certificates WHERE revoke_authority_operation_id=$1 FOR UPDATE`, receipt.OperationID).Scan(&status, &evidence, &resolved); err != nil {
		return err
	}
	if err := dbtx.QueryRow(ctx, `SELECT (SELECT count(*) FROM nodecontrol.node_operator_audit WHERE authority_operation_id=$1),(SELECT count(*) FROM public.transactional_outbox WHERE event_id=$1)`, receipt.OperationID).Scan(&audit, &outbox); err != nil {
		return err
	}
	var originalContext []byte
	if err := dbtx.QueryRow(ctx, `SELECT payload FROM public.transactional_outbox WHERE event_id=$1 AND aggregate_id=$2 AND aggregate_version=$3`, receipt.OperationID, nodeID, int64(receipt.Sequence)).Scan(&originalContext); err != nil {
		return err
	}
	reconstructedContext, err := json.Marshal(coordinatorTestOutcome(proof, resolution))
	if err != nil || !bytes.Equal(originalContext, reconstructedContext) {
		return ErrConflict
	}
	if status != "revoked" || audit != 1 || outbox != 1 || !bytes.Equal(evidence, proof.Evidence().CanonicalJCS()) || !bytes.Equal(resolved, resolution.CanonicalJCS()) {
		return ErrConflict
	}
	return nil
}

func (fixture *postgresCrashFixture) assertAtomicOutcome(t *testing.T, terminal bool) []byte {
	t.Helper()
	var fence, domain, proof, audit, outbox bool
	var snapshot []byte
	err := fixture.pool.QueryRow(t.Context(), `SELECT
 f.provider_status='committed' AND f.visibility_state='active',
 c.status='revoked',
 c.revoke_authority_effect_commitment_jcs IS NOT NULL AND c.revoke_authority_provider_head_jcs IS NOT NULL
 AND c.revoke_authority_activation_evidence_jcs IS NOT NULL AND c.revoke_authority_effect_resolution_jcs IS NOT NULL,
 EXISTS(SELECT 1 FROM nodecontrol.node_operator_audit a WHERE a.authority_operation_id=f.operation_id),
 EXISTS(SELECT 1 FROM public.transactional_outbox o WHERE o.event_id=f.operation_id),
 convert_to(jsonb_build_array(to_jsonb(f),to_jsonb(c),
 (SELECT to_jsonb(a) FROM nodecontrol.node_operator_audit a WHERE a.authority_operation_id=f.operation_id),
 (SELECT to_jsonb(o) FROM public.transactional_outbox o WHERE o.event_id=f.operation_id))::text,'UTF8')
 FROM nodecontrol.control_plane_authority_fences f JOIN nodecontrol.node_certificates c ON c.revoke_authority_operation_id=f.operation_id
 WHERE f.operation_id=$1`, fixture.request.OperationID).Scan(&fence, &domain, &proof, &audit, &outbox, &snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if fence != terminal || domain != terminal || proof != terminal || audit != terminal || outbox != terminal {
		t.Fatalf("atomic terminal components fence/domain/all-proof/audit/outbox=%t/%t/%t/%t/%t want=%t", fence, domain, proof, audit, outbox, terminal)
	}
	if terminal {
		stored, err := fixture.rawRepository.GetStoredFence(t.Context(), fixture.request.OperationID)
		if err != nil || stored.PersistedOutcome == nil {
			t.Fatalf("real persisted outcome loader: %#v %v", stored, err)
		}
	}
	return snapshot
}
func (fixture *postgresCrashFixture) expectFinalizeError(t *testing.T, want error) {
	t.Helper()
	if _, err := fixture.coordinator.Finalize(t.Context(), CoordinatorFinalizeRequest{OperationID: fixture.request.OperationID, EffectDigest: fixture.effectDigest}); err != want {
		t.Fatalf("Finalize error=%v want=%v", err, want)
	}
}

type postgresCrashTransactionFailure struct {
	call  int
	after bool
	err   error
}
type postgresCrashRepository struct {
	*PostgresRepository
	mu                     sync.Mutex
	transactions, captures int
	transactionActive      bool
	failure                *postgresCrashTransactionFailure
	began                  chan<- int32
	beforeCommit           func(context.Context, store.DBTX) error
	beforeCapture          func()
}

func (repository *postgresCrashRepository) CaptureDatabasePoint(ctx context.Context) (DatabasePoint, error) {
	if repository.beforeCapture != nil {
		repository.beforeCapture()
	}
	repository.mu.Lock()
	repository.captures++
	repository.mu.Unlock()
	return repository.PostgresRepository.CaptureDatabasePoint(ctx)
}
func (repository *postgresCrashRepository) beginAuthorityTransaction(ctx context.Context) (authorityTransaction, error) {
	repository.mu.Lock()
	repository.transactions++
	call := repository.transactions
	failure := repository.failure
	if failure != nil && failure.call == call {
		repository.failure = nil
	} else {
		failure = nil
	}
	repository.mu.Unlock()
	if failure != nil && !failure.after {
		return nil, failure.err
	}
	tx, err := repository.PostgresRepository.beginAuthorityTransaction(ctx)
	if err != nil {
		return nil, err
	}
	repository.mu.Lock()
	repository.transactionActive = true
	repository.mu.Unlock()
	if repository.began != nil {
		var pid int32
		if err := tx.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&pid); err != nil {
			_ = tx.Rollback(ctx)
			return nil, err
		}
		repository.began <- pid
	}
	return &postgresCrashTransaction{authorityTransaction: tx, repository: repository, failure: failure}, nil
}

type postgresCrashTransaction struct {
	authorityTransaction
	repository *postgresCrashRepository
	failure    *postgresCrashTransactionFailure
	done       bool
}

func (tx *postgresCrashTransaction) Commit(ctx context.Context) error {
	if tx.repository.beforeCommit != nil {
		if err := tx.repository.beforeCommit(ctx, tx); err != nil {
			return err
		}
	}
	err := tx.authorityTransaction.Commit(ctx)
	tx.release()
	if err == nil && tx.failure != nil && tx.failure.after {
		return tx.failure.err
	}
	return err
}
func (tx *postgresCrashTransaction) Rollback(ctx context.Context) error {
	err := tx.authorityTransaction.Rollback(ctx)
	tx.release()
	return err
}
func (tx *postgresCrashTransaction) release() {
	if !tx.done {
		tx.done = true
		tx.repository.mu.Lock()
		tx.repository.transactionActive = false
		tx.repository.mu.Unlock()
	}
}
func (repository *postgresCrashRepository) failTransaction(call int, after bool, err error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.failure = &postgresCrashTransactionFailure{call, after, err}
}
func (repository *postgresCrashRepository) captureCount() int {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	return repository.captures
}
func (repository *postgresCrashRepository) transactionCount() int {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	return repository.transactions
}
func (repository *postgresCrashRepository) inTransaction() bool {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	return repository.transactionActive
}

// Test-only inert closure setup copied from internal/store/
// nodecontrol_v7_acl_integration_test.go:task8SeedProofActivation. Replica mode
// applies only to prerequisite rows; target claim/fence/domain/proof mutations
// run with origin triggers. These CHECK-valid fixture bodies are not signed
// protocol evidence and this does not validate production activation.
func task9SeedProofActivation(t *testing.T, tx pgx.Tx, activationID uuid.UUID, discriminator int, now time.Time) {
	t.Helper()
	var installationID uuid.UUID
	if err := tx.QueryRow(t.Context(), `SELECT installation_id FROM nodecontrol.control_plane_authority_protocol_migration_latches WHERE singleton_key ORDER BY installation_id LIMIT 1`).Scan(&installationID); err != nil {
		t.Fatal("load exact installed migration latch for claim-v1 closure:", err)
	}
	if _, err := tx.Exec(t.Context(), `SET LOCAL session_replication_role=replica`); err != nil {
		t.Fatal(err)
	}
	deploymentID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-deployment:%d", discriminator)))
	intentID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-intent:%d", discriminator)))
	registrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-registration:%d", discriminator)))
	attemptID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-attempt:%d", discriminator)))
	preparationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-preparation:%d", discriminator)))
	completionID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-completion:%d", discriminator)))
	releasePreparationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-release:%d", discriminator)))
	openID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-open:%d", discriminator)))
	if _, err := tx.Exec(t.Context(), `
INSERT INTO nodecontrol.control_plane_authority_protocol_upgrade_intents(
 intent_id,installation_id,installation_kind,activation_id,request_nonce,incarnation_registration_id,
 provider_absence_proof_id,observed_deployment_id,database_identity_digest,local_runtime_isolation_digest,
 classification_state,created_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest)
VALUES($1,$2,'production',$3,decode(repeat('91',32),'hex'),$4,$5,$6,decode(repeat('92',32),'hex'),
 decode(repeat('93',32),'hex'),'pending',$7,decode('7b7d','hex'),decode('7b7d','hex'),decode(repeat('a1',32),'hex'))`,
		intentID, installationID, activationID,
		uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-incarnation:%d", discriminator))),
		uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-absence:%d", discriminator))),
		deploymentID, now); err != nil {
		t.Fatal("seed legal claim-v1 upgrade intent:", err)
	}
	if _, err := tx.Exec(t.Context(), `
INSERT INTO nodecontrol.control_plane_authority_runtime_registration_results(
 registration_id,upgrade_intent_digest,activation_id,provider_identity_digest,provider_endpoint_identity_digest,
 namespace,credential_policy_digest,database_incarnation_attestation_digest,provider_registration_digest,
 genesis_database_identity_digest,database_timeline_lineage_chain_digest,runtime_instance_binding_digest,
 runtime_instance_id,runtime_instance_generation,attestor_runtime_lease_digest,runtime_rebind_chain_digest,
 provider_head_digest,provider_phase,provider_control_sequence,database_point,recorded_at,
 canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest)
VALUES($1,decode(repeat('a1',32),'hex'),$2,decode(repeat('94',32),'hex'),decode(repeat('95',32),'hex'),
 'task8-proof',decode(repeat('a6',32),'hex'),decode(repeat('ab',32),'hex'),decode(repeat('98',32),'hex'),
 decode(repeat('92',32),'hex'),decode(repeat('9a',32),'hex'),decode(repeat('ae',32),'hex'),$3,1,
 decode(repeat('9c',32),'hex'),decode(repeat('ad',32),'hex'),decode(repeat('9e',32),'hex'),
 'registered_pending_genesis',0,decode('01','hex'),$4,decode('7b7d','hex'),decode('7b7d','hex'),decode(repeat('a2',32),'hex'))`,
		registrationID, activationID,
		uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("task8-proof-runtime:%d", discriminator))), now); err != nil {
		t.Fatal("seed legal claim-v1 runtime registration:", err)
	}
	if _, err := tx.Exec(t.Context(), `
INSERT INTO nodecontrol.control_plane_authority_protocol_upgrade_attempts(
 attempt_id,upgrade_intent_digest,activation_id,preparation_id,completion_id,release_preparation_id,open_id,
 mode,deployment_id,request_nonce,environment_inventory_digest,environment_inventory_anchor_set_digest,
 local_runtime_isolation_digest,legacy_runtime_shutdown_digest,legacy_runtime_shutdown_set_digest,
 credential_policy_digest,database_legacy_absence_projection_digest,attempt_database_observation_digest,
 provider_namespace_absence_digest,database_inventory_digest,database_incarnation_attestation_digest,
 database_incarnation_registration_digest,runtime_registration_result_digest,runtime_rebind_chain_digest,
 runtime_instance_binding_digest,epoch_evidence_digest,genesis_epoch_transition_root_digest,selected_genesis_epoch,
 created_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest)
VALUES($1,decode(repeat('a1',32),'hex'),$2,$3,$4,$5,$6,'empty_in_place',$7,decode(repeat('91',32),'hex'),
 decode(repeat('a1',32),'hex'),decode(repeat('a2',32),'hex'),decode(repeat('93',32),'hex'),decode(repeat('a4',32),'hex'),
 decode(repeat('a5',32),'hex'),decode(repeat('a6',32),'hex'),decode(repeat('a7',32),'hex'),decode(repeat('a8',32),'hex'),
 decode(repeat('a9',32),'hex'),decode(repeat('aa',32),'hex'),decode(repeat('ab',32),'hex'),decode(repeat('ac',32),'hex'),
 decode(repeat('a2',32),'hex'),decode(repeat('ad',32),'hex'),decode(repeat('ae',32),'hex'),decode(repeat('af',32),'hex'),
 decode(repeat('b0',32),'hex'),1,$8,decode('7b7d','hex'),decode('7b7d','hex'),decode(repeat('a3',32),'hex'))`,
		attemptID, activationID, preparationID, completionID, releasePreparationID, openID, deploymentID, now); err != nil {
		t.Fatal("seed legal claim-v1 upgrade attempt:", err)
	}
	_, err := tx.Exec(t.Context(), `
INSERT INTO nodecontrol.control_plane_authority_protocol_activations(
 activation_id,mode,attempt_digest,deployment_id,database_identity_digest,database_incarnation_attestation_digest,
 genesis_database_incarnation_registration_digest,runtime_registration_result_digest,runtime_rebind_chain_digest,
 activation_runtime_instance_binding_digest,database_legacy_absence_projection_digest,attempt_database_inventory_digest,
 activation_database_observation_digest,provider_namespace_absence_digest,environment_inventory_digest,
 environment_inventory_anchor_set_digest,local_runtime_isolation_digest,legacy_runtime_shutdown_digest,
 legacy_runtime_shutdown_set_digest,credential_policy_digest,epoch_evidence_digest,genesis_epoch_transition_root_digest,
 selected_genesis_epoch,provider_identity_digest,provider_endpoint_identity_digest,namespace,protocol_profile,
 preparation_digest,activated_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest)
VALUES($1,'empty_in_place',decode(repeat('a3',32),'hex'),$2,decode(repeat('92',32),'hex'),decode(repeat('ab',32),'hex'),
	 decode(repeat('ac',32),'hex'),decode(repeat('a2',32),'hex'),decode(repeat('ad',32),'hex'),decode(repeat('ae',32),'hex'),
	 decode(repeat('a7',32),'hex'),decode(repeat('aa',32),'hex'),decode(repeat('5a',32),'hex'),decode(repeat('a9',32),'hex'),
	 decode(repeat('a1',32),'hex'),decode(repeat('a2',32),'hex'),decode(repeat('93',32),'hex'),decode(repeat('a4',32),'hex'),
	 decode(repeat('a5',32),'hex'),decode(repeat('a6',32),'hex'),decode(repeat('af',32),'hex'),decode(repeat('b0',32),'hex'),
	 1,decode(repeat('94',32),'hex'),decode(repeat('95',32),'hex'),'task8-proof','claim_v1',decode(repeat('66',32),'hex'),
	 $3,decode('7b7d','hex'),decode('7b7d','hex'),decode(repeat('a4',32),'hex'))`, activationID, deploymentID, now)
	if err != nil {
		t.Fatal("seed CHECK-valid claim-v1 activation:", err)
	}
	if _, err := tx.Exec(t.Context(), `
INSERT INTO nodecontrol.control_plane_authority_protocol_activation_completions(
 completion_id,activation_id,activation_digest,preparation_digest,provider_completion_digest,
 provider_completion_phase,database_activation_attestation_digest,current_database_incarnation_registration_digest,
 latest_runtime_rebind_result_digest_or_null,runtime_rebind_chain_digest,current_runtime_instance_binding_digest,
 credential_policy_digest,epoch_evidence_digest,genesis_epoch_transition_root_digest,selected_genesis_epoch,
 completed_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest)
VALUES($1,$2,decode(repeat('a4',32),'hex'),decode(repeat('66',32),'hex'),decode(repeat('b1',32),'hex'),
	 'genesis_completed_pending_release',decode(repeat('b2',32),'hex'),decode(repeat('ac',32),'hex'),NULL,
	 decode(repeat('ad',32),'hex'),decode(repeat('ae',32),'hex'),decode(repeat('a6',32),'hex'),
	 decode(repeat('af',32),'hex'),decode(repeat('b0',32),'hex'),1,$3,decode('7b7d','hex'),decode('7b7d','hex'),
 decode(repeat('a5',32),'hex'))`, completionID, activationID, now); err != nil {
		t.Fatal("seed CHECK-valid claim-v1 activation completion:", err)
	}
	if _, err := tx.Exec(t.Context(), `
INSERT INTO nodecontrol.control_plane_authority_protocol_activation_releases(
 release_preparation_id,open_id,activation_id,activation_digest,completion_digest,
 provider_release_preparation_digest,provider_release_phase,database_completion_attestation_digest,open_nonce,
 current_database_incarnation_registration_digest,latest_runtime_rebind_result_digest_or_null,runtime_rebind_chain_digest,
 current_runtime_instance_binding_digest,credential_policy_digest,epoch_evidence_digest,genesis_epoch_transition_root_digest,
 selected_genesis_epoch,released_at,canonical_evidence_bundle_jcs,canonical_body_jcs,body_digest)
VALUES($1,$2,$3,decode(repeat('a4',32),'hex'),decode(repeat('a5',32),'hex'),decode(repeat('b3',32),'hex'),
	 'genesis_release_prepared',decode(repeat('b4',32),'hex'),decode(repeat('b5',32),'hex'),decode(repeat('ac',32),'hex'),
	 NULL,decode(repeat('ad',32),'hex'),decode(repeat('ae',32),'hex'),decode(repeat('a6',32),'hex'),
	 decode(repeat('af',32),'hex'),decode(repeat('b0',32),'hex'),1,$4,decode('7b7d','hex'),decode('7b7d','hex'),
 decode(repeat('a6',32),'hex'))`, releasePreparationID, openID, activationID, now); err != nil {
		t.Fatal("seed CHECK-valid claim-v1 activation release:", err)
	}
	if _, err := tx.Exec(t.Context(), `SET LOCAL session_replication_role=origin`); err != nil {
		t.Fatal(err)
	}
}

type postgresTransactionGuardProvider struct {
	Provider
	repository *postgresCrashRepository
	mu         sync.Mutex
	violation  bool
	aborts     map[uuid.UUID]int
}

func (provider *postgresTransactionGuardProvider) Reserve(ctx context.Context, request ReserveRequest) (Reservation, error) {
	if provider.checkTransaction() {
		return Reservation{}, ErrInjectedFailure
	}
	return provider.Provider.Reserve(ctx, request)
}

func (provider *postgresTransactionGuardProvider) Finalize(ctx context.Context, request FinalizeRequest) (Receipt, error) {
	if provider.checkTransaction() {
		return Receipt{}, ErrInjectedFailure
	}
	return provider.Provider.Finalize(ctx, request)
}

func (provider *postgresTransactionGuardProvider) Abort(ctx context.Context, request AbortRequest) (Receipt, error) {
	if provider.checkTransaction() {
		return Receipt{}, ErrInjectedFailure
	}
	provider.mu.Lock()
	provider.aborts[request.OperationID]++
	provider.mu.Unlock()
	return provider.Provider.Abort(ctx, request)
}

func (provider *postgresTransactionGuardProvider) Inspect(ctx context.Context, operationID uuid.UUID) (Record, error) {
	if provider.checkTransaction() {
		return Record{}, ErrInjectedFailure
	}
	return provider.Provider.Inspect(ctx, operationID)
}

func (provider *postgresTransactionGuardProvider) Head(ctx context.Context) (Head, error) {
	if provider.checkTransaction() {
		return Head{}, ErrInjectedFailure
	}
	return provider.Provider.Head(ctx)
}

func (provider *postgresTransactionGuardProvider) CommittedNodeCheckpoint(ctx context.Context, digest contracts.Digest) (NodeCheckpoint, error) {
	if provider.checkTransaction() {
		return NodeCheckpoint{}, ErrInjectedFailure
	}
	return provider.Provider.CommittedNodeCheckpoint(ctx, digest)
}

func (provider *postgresTransactionGuardProvider) checkTransaction() bool {
	if !provider.repository.inTransaction() {
		return false
	}
	provider.mu.Lock()
	provider.violation = true
	provider.mu.Unlock()
	return true
}

func (provider *postgresTransactionGuardProvider) providerCallInTransaction() bool {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	return provider.violation
}

func (provider *postgresTransactionGuardProvider) abortCount(operationID uuid.UUID) int {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	return provider.aborts[operationID]
}

var _ Provider = (*postgresTransactionGuardProvider)(nil)

func TestCoordinatorPostgresAtomicActivationCrashMatrix(t *testing.T) {
	for _, cut := range []string{"material", "domain", "audit", "outbox", "commit_not_invoked", "commit_response_lost"} {
		t.Run(cut, func(t *testing.T) {
			fixture := newPostgresCrashFixture(t)
			fixture.reserveAndCommitDomain(t, true)
			switch cut {
			case "material":
				fixture.effects.captureErr = ErrInjectedFailure
			case "domain", "audit", "outbox":
				fixture.effects.activationFailure = cut
			case "commit_not_invoked":
				fixture.repository.beforeCommit = func(ctx context.Context, dbtx store.DBTX) error {
					var terminal bool
					if err := dbtx.QueryRow(ctx, `SELECT provider_status='committed' FROM nodecontrol.control_plane_authority_fences WHERE operation_id=$1`, fixture.request.OperationID).Scan(&terminal); err != nil {
						return err
					}
					if terminal {
						return ErrInjectedFailure
					}
					return nil
				}
			case "commit_response_lost":
				fixture.repository.failTransaction(4, true, ErrResponseLost)
			}
			want := error(ErrInjectedFailure)
			if cut == "commit_response_lost" {
				want = ErrResponseLost
			}
			fixture.expectFinalizeError(t, want)
			committed := cut == "commit_response_lost"
			before := fixture.assertAtomicOutcome(t, committed)
			fixture.effects.mu.Lock()
			captures := fixture.effects.captures
			fixture.effects.captureErr = nil
			fixture.effects.mu.Unlock()
			fixture.effects.activationFailure = ""
			fixture.repository.beforeCommit = nil
			restarted := mustNewCoordinatorForTest(t, fixture.guardedProvider, fixture.repository, fixture.effects, fixture.clock)
			receipt, err := restarted.Recover(t.Context(), fixture.request.OperationID)
			if err != nil || receipt.Status != StatusCommitted {
				t.Fatalf("recover %s: %#v %v", cut, receipt, err)
			}
			after := fixture.assertAtomicOutcome(t, true)
			if committed {
				if !bytes.Equal(before, after) {
					t.Fatal("uncertain committed recovery changed original fence/domain/proof/audit/outbox bytes")
				}
				fixture.effects.mu.Lock()
				gotCaptures, validations := fixture.effects.captures, fixture.effects.validations
				fixture.effects.mu.Unlock()
				if gotCaptures != captures || validations == 0 {
					t.Fatalf("committed recovery recaptured or omitted parsed-only validator: captures=%d/%d validations=%d", gotCaptures, captures, validations)
				}
			}
			if fixture.guardedProvider.providerCallInTransaction() {
				t.Fatal("Provider call while SQL transaction active")
			}
		})
	}
}

func TestCoordinatorPostgresFenceFirstAbortRace(t *testing.T) {
	t.Run("effect commits before Coordinator Abort", func(t *testing.T) {
		fixture := newPostgresCrashFixture(t)
		fixture.reserveAndCommitDomain(t, false)
		reservation := fixture.provider.Snapshot()[0].Reservation
		setup, err := fixture.pool.Begin(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		fixture.seedCertificate(t, setup, reservation)
		if err := setup.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		writer, err := fixture.pool.BeginTx(t.Context(), pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
		if err != nil {
			t.Fatal(err)
		}
		defer writer.Rollback(t.Context())
		fixture.effectDigest = fixture.commitDomain(t, writer, reservation)
		var writerPID int32
		if err := writer.QueryRow(t.Context(), "SELECT pg_backend_pid()").Scan(&writerPID); err != nil {
			t.Fatal(err)
		}
		began := make(chan int32, 4)
		fixture.repository.began = began
		result := make(chan error, 1)
		go func() {
			_, err := fixture.coordinator.Abort(t.Context(), AbortRequest{OperationID: reservation.OperationID, Reason: AbortSuperseded})
			result <- err
		}()
		waiterPID := task9WaitPID(t, began)
		task8WaitForRepositoryLockBlock(t, t.Context(), fixture.pool, writerPID, waiterPID)
		if err := writer.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		if err := task9WaitError(t, result); err != ErrConflict {
			t.Fatalf("effect-losing Abort=%v", err)
		}
		stored, err := fixture.rawRepository.GetStoredFence(t.Context(), reservation.OperationID)
		if err != nil || stored.AbortClaim != nil || fixture.guardedProvider.abortCount(reservation.OperationID) != 0 {
			t.Fatalf("Abort loser mutated claim/provider: %#v %v", stored, err)
		}
		if _, err := fixture.coordinator.Finalize(t.Context(), CoordinatorFinalizeRequest{OperationID: reservation.OperationID, EffectDigest: fixture.effectDigest}); err != nil {
			t.Fatal(err)
		}
		fixture.assertAtomicOutcome(t, true)
	})
	t.Run("Coordinator claim commits before effect writer", func(t *testing.T) {
		fixture := newPostgresCrashFixture(t)
		fixture.reserveAndCommitDomain(t, false)
		reservation := fixture.provider.Snapshot()[0].Reservation
		setup, err := fixture.pool.Begin(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		fixture.seedCertificate(t, setup, reservation)
		if err := setup.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		entered := make(chan int32, 1)
		release := make(chan struct{})
		var once sync.Once
		fixture.effects.onResolve = func(ctx context.Context, dbtx store.DBTX) error {
			var wait bool
			once.Do(func() { wait = true })
			if wait {
				var pid int32
				if err := dbtx.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&pid); err != nil {
					return err
				}
				entered <- pid
				select {
				case <-release:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		}
		aborted := make(chan error, 1)
		go func() {
			_, err := fixture.coordinator.Abort(t.Context(), AbortRequest{OperationID: reservation.OperationID, Reason: AbortSuperseded})
			aborted <- err
		}()
		firstPID := task9WaitPID(t, entered)
		writer, err := fixture.pool.BeginTx(t.Context(), pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
		if err != nil {
			t.Fatal(err)
		}
		defer writer.Rollback(t.Context())
		var writerPID int32
		if err := writer.QueryRow(t.Context(), "SELECT pg_backend_pid()").Scan(&writerPID); err != nil {
			t.Fatal(err)
		}
		loser := make(chan error, 1)
		go func() {
			stored, err := fixture.rawRepository.Lock(t.Context(), writer, reservation.OperationID)
			if err == nil && (stored.AbortClaim != nil || stored.Record.TerminalReceipt != nil) {
				err = ErrConflict
			}
			_ = writer.Rollback(t.Context())
			loser <- err
		}()
		task8WaitForRepositoryLockBlock(t, t.Context(), fixture.pool, firstPID, writerPID)
		close(release)
		if err := task9WaitError(t, loser); err != ErrConflict {
			t.Fatalf("claim-losing effect=%v", err)
		}
		if err := task9WaitError(t, aborted); err != nil {
			t.Fatal(err)
		}
		stored, err := fixture.rawRepository.GetStoredFence(t.Context(), reservation.OperationID)
		if err != nil || stored.AbortClaim == nil || stored.AbortClaim.Reason != AbortSuperseded || stored.Record.TerminalReceipt == nil || stored.Record.TerminalReceipt.Status != StatusAborted {
			t.Fatalf("claim winner state=%#v error=%v", stored, err)
		}
		var effects, proofs int
		if err := fixture.pool.QueryRow(t.Context(), `SELECT (SELECT count(*) FROM nodecontrol.authority_task7_crash_effects WHERE operation_id=$1),(SELECT count(*) FROM nodecontrol.node_certificates WHERE revoke_authority_operation_id=$1)`, reservation.OperationID).Scan(&effects, &proofs); err != nil {
			t.Fatal(err)
		}
		if effects != 0 || proofs != 0 || fixture.guardedProvider.abortCount(reservation.OperationID) != 1 {
			t.Fatalf("semantic winners effect/proof/provider-abort=%d/%d/%d", effects, proofs, fixture.guardedProvider.abortCount(reservation.OperationID))
		}
	})
}

func TestCoordinatorPostgresFinalizeCapturesAfterEffectWriterCommit(t *testing.T) {
	fixture := newPostgresCrashFixture(t)
	fixture.reserveAndCommitDomain(t, false)
	reservation := fixture.provider.Snapshot()[0].Reservation
	setup, err := fixture.pool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	fixture.seedCertificate(t, setup, reservation)
	if err := setup.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	writer, err := fixture.pool.BeginTx(t.Context(), pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Rollback(t.Context())
	fixture.effectDigest = fixture.commitDomain(t, writer, reservation)
	var writerPID int32
	if err := writer.QueryRow(t.Context(), "SELECT pg_backend_pid()").Scan(&writerPID); err != nil {
		t.Fatal(err)
	}
	began := make(chan int32, 8)
	fixture.repository.began = began
	captureEntered := make(chan struct{})
	captureRelease := make(chan struct{})
	fixture.repository.beforeCapture = func() { close(captureEntered); <-captureRelease }
	receiptCh := make(chan Receipt, 1)
	result := make(chan error, 1)
	go func() {
		receipt, err := fixture.coordinator.Finalize(t.Context(), CoordinatorFinalizeRequest{OperationID: reservation.OperationID, EffectDigest: fixture.effectDigest})
		receiptCh <- receipt
		result <- err
	}()
	waiterPID := task9WaitPID(t, began)
	task8WaitForRepositoryLockBlock(t, t.Context(), fixture.pool, writerPID, waiterPID)
	if fixture.repository.captureCount() != 0 {
		t.Fatal("captured WAL before locked EffectCommitted proof")
	}
	if err := writer.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-captureEntered:
	case <-time.After(10 * time.Second):
		t.Fatal("post-effect capture not reached")
	}
	postCommitPoint, err := fixture.rawRepository.CaptureDatabasePoint(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	close(captureRelease)
	if err := task9WaitError(t, result); err != nil {
		t.Fatal(err)
	}
	receipt := <-receiptCh
	if receipt.DatabasePoint == nil || walPositionLess(receipt.DatabasePoint.RequiredLSN, postCommitPoint.RequiredLSN) {
		t.Fatalf("receipt WAL point %#v precedes released effect commit %s", receipt.DatabasePoint, postCommitPoint.RequiredLSN)
	}
	fixture.assertAtomicOutcome(t, true)
}
func task9WaitPID(t *testing.T, ch <-chan int32) int32 {
	t.Helper()
	select {
	case pid := <-ch:
		return pid
	case <-time.After(10 * time.Second):
		t.Fatal("transaction barrier not reached")
		return 0
	}
}
func task9WaitError(t *testing.T, ch <-chan error) error {
	t.Helper()
	select {
	case err := <-ch:
		return err
	case <-time.After(10 * time.Second):
		t.Fatal("transaction did not finish after lock release")
		return ErrInjectedFailure
	}
}
