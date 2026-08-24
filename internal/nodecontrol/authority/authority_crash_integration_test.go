//go:build integration

package authority

import (
	"context"
	"crypto/sha256"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
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
				fixture.repository.failTransaction(1, false, ErrInjectedFailure)
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
				fixture.repository.failTransaction(2, false, ErrInjectedFailure)
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
				fixture.repository.failTransaction(2, true, ErrResponseLost)
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
			if err != nil {
				t.Fatalf("Recover error = %v", err)
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
			if fixture.effects.commitCount(fixture.request.OperationID) != point.wantEffectCommits {
				t.Fatalf("domain effect commits = %d, want %d", fixture.effects.commitCount(fixture.request.OperationID), point.wantEffectCommits)
			}
			if fixture.guardedProvider.providerCallInTransaction() {
				t.Fatal("coordinator called provider from inside a database transaction")
			}
		})
	}
}

type postgresCrashFixture struct {
	pool            *pgxpool.Pool
	provider        *DeterministicProvider
	guardedProvider *postgresTransactionGuardProvider
	repository      *postgresCrashRepository
	rawRepository   *PostgresRepository
	effects         *coordinatorEffectResolver
	clock           coordinatorClock
	coordinator     *Coordinator
	request         ReserveRequest
	effectDigest    contracts.Digest
	effectCommitLSN WALPosition
}

func newPostgresCrashFixture(t *testing.T) *postgresCrashFixture {
	t.Helper()
	ctx, pool := openMigratedAuthorityDatabase(t)
	if _, err := pool.Exec(ctx, `CREATE TABLE nodecontrol.authority_task7_crash_effects(operation_id uuid PRIMARY KEY,effect_digest bytea NOT NULL CHECK(octet_length(effect_digest)=32))`); err != nil {
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
	guardedProvider := &postgresTransactionGuardProvider{Provider: provider, repository: repository}
	effects := newCoordinatorEffectResolver()
	clock := coordinatorClock{now: time.Date(2026, 8, 23, 13, 0, 0, 0, time.UTC)}
	coordinator := mustNewCoordinatorForTest(t, guardedProvider, repository, effects, clock)
	return &postgresCrashFixture{
		pool:            pool,
		provider:        provider,
		guardedProvider: guardedProvider,
		repository:      repository,
		rawRepository:   rawRepository,
		effects:         effects,
		clock:           clock,
		coordinator:     coordinator,
		request: ReserveRequest{
			OperationID: uuid.New(),
			Kind:        EffectCertificateRevoke,
			ScopeKind:   ScopeNode,
			ScopeDigest: sha256.Sum256([]byte("postgres-crash-node")),
		},
		effectDigest: sha256.Sum256([]byte("postgres-crash-effect")),
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
	if err := fixture.rawRepository.RecordPending(t.Context(), transaction, reservation, fixture.clock.Now().Add(-time.Second)); err != nil {
		_ = transaction.Rollback(t.Context())
		t.Fatal(err)
	}
	if commitEffect {
		if _, err := transaction.Exec(t.Context(), `INSERT INTO nodecontrol.authority_task7_crash_effects(operation_id,effect_digest) VALUES ($1,$2)`, fixture.request.OperationID, fixture.effectDigest[:]); err != nil {
			_ = transaction.Rollback(t.Context())
			t.Fatal(err)
		}
	}
	if err := transaction.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	if !commitEffect {
		return
	}
	point, err := fixture.rawRepository.CaptureDatabasePoint(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	fixture.effectCommitLSN = point.RequiredLSN
	fixture.effects.commit(t, fixture.request, fixture.effectDigest, fixture.effectCommitLSN)
}

func (fixture *postgresCrashFixture) expectFinalizeError(t *testing.T, want error) {
	t.Helper()
	if _, err := fixture.coordinator.Finalize(t.Context(), CoordinatorFinalizeRequest{OperationID: fixture.request.OperationID, EffectDigest: fixture.effectDigest}); err != want {
		t.Fatalf("Finalize error = %v, want %v", err, want)
	}
}

type postgresCrashTransactionFailure struct {
	call  int
	after bool
	err   error
}

type postgresCrashRepository struct {
	*PostgresRepository
	mu                sync.Mutex
	transactions      int
	captures          int
	transactionActive bool
	failure           *postgresCrashTransactionFailure
}

func (repository *postgresCrashRepository) CaptureDatabasePoint(ctx context.Context) (DatabasePoint, error) {
	repository.mu.Lock()
	repository.captures++
	repository.mu.Unlock()
	return repository.PostgresRepository.CaptureDatabasePoint(ctx)
}

func (repository *postgresCrashRepository) withAuthorityTransaction(ctx context.Context, operation func(store.DBTX) error) error {
	repository.mu.Lock()
	repository.transactions++
	call := repository.transactions
	failure := repository.failure
	if failure != nil && failure.call == call && !failure.after {
		repository.failure = nil
		repository.mu.Unlock()
		return failure.err
	}
	repository.mu.Unlock()

	err := repository.PostgresRepository.withAuthorityTransaction(ctx, func(dbtx store.DBTX) error {
		repository.mu.Lock()
		repository.transactionActive = true
		repository.mu.Unlock()
		defer func() {
			repository.mu.Lock()
			repository.transactionActive = false
			repository.mu.Unlock()
		}()
		return operation(dbtx)
	})
	if err != nil {
		return err
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if failure != nil && failure.call == call && failure.after {
		repository.failure = nil
		return failure.err
	}
	return nil
}

func (repository *postgresCrashRepository) failTransaction(call int, after bool, err error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.failure = &postgresCrashTransactionFailure{call: call, after: after, err: err}
}

func (repository *postgresCrashRepository) captureCount() int {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	return repository.captures
}

func (repository *postgresCrashRepository) inTransaction() bool {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	return repository.transactionActive
}

type postgresTransactionGuardProvider struct {
	Provider
	repository *postgresCrashRepository
	mu         sync.Mutex
	violation  bool
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

var _ Provider = (*postgresTransactionGuardProvider)(nil)
