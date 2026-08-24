package authority

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
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
				effects.commit(t, request, effectDigest, "0/20")
			},
			wantStatus:        StatusCommitted,
			wantEffectCommits: 1,
		},
		{
			name: "database point captured before bind",
			prepare: func(t *testing.T, coordinator *Coordinator, _ *DeterministicProvider, repository *coordinatorMemoryRepository, effects *coordinatorEffectResolver, request ReserveRequest, effectDigest contracts.Digest) {
				t.Helper()
				coordinatorReserveAndRecord(t, coordinator, repository, request)
				effects.commit(t, request, effectDigest, "0/20")
				repository.failTransaction(1, false, ErrInjectedFailure)
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
				effects.commit(t, request, effectDigest, "0/20")
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
				effects.commit(t, request, effectDigest, "0/20")
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
				effects.commit(t, request, effectDigest, "0/20")
				repository.failTransaction(2, false, ErrInjectedFailure)
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
				effects.commit(t, request, effectDigest, "0/20")
				repository.failTransaction(2, true, ErrResponseLost)
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
	effects.commit(t, request, effectDigest, "0/20")
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
	coordinator, err := NewCoordinator(provider, repository, effects, clock)
	if err != nil {
		t.Fatal(err)
	}
	return coordinator
}

type coordinatorClock struct{ now time.Time }

func (clock coordinatorClock) Now() time.Time { return clock.now }

type coordinatorEffectResolver struct {
	mu       sync.Mutex
	effects  map[uuid.UUID]ResolvedEffect
	commits  map[uuid.UUID]int
	lsns     map[uuid.UUID]WALPosition
	resolves map[uuid.UUID]int
}

func newCoordinatorEffectResolver() *coordinatorEffectResolver {
	return &coordinatorEffectResolver{
		effects:  make(map[uuid.UUID]ResolvedEffect),
		commits:  make(map[uuid.UUID]int),
		lsns:     make(map[uuid.UUID]WALPosition),
		resolves: make(map[uuid.UUID]int),
	}
}

func (resolver *coordinatorEffectResolver) ResolveAuthorityEffect(ctx context.Context, operationID uuid.UUID) (ResolvedEffect, error) {
	if ctx == nil || ctx.Err() != nil {
		return ResolvedEffect{}, ErrCanceled
	}
	resolver.mu.Lock()
	defer resolver.mu.Unlock()
	resolver.resolves[operationID]++
	resolved, ok := resolver.effects[operationID]
	if !ok {
		return ResolvedEffect{State: EffectAbsent}, nil
	}
	return resolved, nil
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

func (resolver *coordinatorEffectResolver) commit(t *testing.T, request ReserveRequest, digest contracts.Digest, lsn WALPosition) {
	t.Helper()
	if walPositionOrdinal(t, lsn) == 0 {
		t.Fatal("effect commit LSN must be positive")
	}
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
	mu          sync.Mutex
	inspections map[uuid.UUID]int
}

func (provider *coordinatorCountingProvider) Inspect(ctx context.Context, operationID uuid.UUID) (Record, error) {
	provider.mu.Lock()
	provider.inspections[operationID]++
	provider.mu.Unlock()
	return provider.Provider.Inspect(ctx, operationID)
}

func (provider *coordinatorCountingProvider) inspectCount(operationID uuid.UUID) int {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	return provider.inspections[operationID]
}

type coordinatorTransactionFailure struct {
	call  int
	after bool
	err   error
}

type coordinatorMemoryRepository struct {
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
}

func newCoordinatorMemoryRepository(points []DatabasePoint) *coordinatorMemoryRepository {
	cloned := append([]DatabasePoint(nil), points...)
	return &coordinatorMemoryRepository{
		records:    make(map[uuid.UUID]Record),
		reservedAt: make(map[uuid.UUID]time.Time),
		boundAt:    make(map[uuid.UUID]time.Time),
		terminalAt: make(map[uuid.UUID]time.Time),
		points:     cloned,
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

func (repository *coordinatorMemoryRepository) BindEffect(_ context.Context, _ store.DBTX, operationID uuid.UUID, effectDigest contracts.Digest, point DatabasePoint, boundAt time.Time) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	record, ok := repository.records[operationID]
	if !ok {
		return ErrNotFound
	}
	if record.TerminalReceipt != nil {
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

func (repository *coordinatorMemoryRepository) ActivateCommitted(_ context.Context, _ store.DBTX, receipt Receipt, terminalAt time.Time) error {
	return repository.recordTerminal(receipt, terminalAt)
}

func (repository *coordinatorMemoryRepository) RecordAborted(_ context.Context, _ store.DBTX, receipt Receipt, terminalAt time.Time) error {
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
