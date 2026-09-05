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
			if repository.transactionCount() != transactionsBefore {
				t.Fatalf("transactions = %d, want unchanged %d", repository.transactionCount(), transactionsBefore)
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
	if repository.transactionCount() != 0 {
		t.Fatalf("transactions after response loss = %d, want 0", repository.transactionCount())
	}

	restarted := mustNewCoordinatorForTest(t, countingProvider, repository, effects, coordinatorClock{now: time.Date(2026, 8, 24, 11, 1, 0, 0, time.UTC)})
	receipt, err := restarted.Abort(t.Context(), abortRequest)
	if err != nil || receipt.Status != StatusAborted {
		t.Fatalf("response-loss retry = %#v, %v", receipt, err)
	}
	if countingProvider.abortCount(request.OperationID) != 1 || repository.transactionCount() != 1 {
		t.Fatalf("after recovery Abort=%d transactions=%d, want 1/1", countingProvider.abortCount(request.OperationID), repository.transactionCount())
	}
	transactionsBefore := repository.transactionCount()
	retry := mustNewCoordinatorForTest(t, countingProvider, repository, effects, coordinatorClock{now: time.Date(2026, 8, 24, 11, 2, 0, 0, time.UTC)})
	retried, err := retry.Abort(t.Context(), abortRequest)
	if err != nil || retried.ReceiptDigest != receipt.ReceiptDigest || retried.Validate() != nil {
		t.Fatalf("exact terminal retry = %#v, %v; want %#v", retried, err, receipt)
	}
	if countingProvider.abortCount(request.OperationID) != 1 || repository.transactionCount() != transactionsBefore {
		t.Fatalf("terminal retry Abort=%d transactions=%d, want 1/%d", countingProvider.abortCount(request.OperationID), repository.transactionCount(), transactionsBefore)
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
	effects.commit(t, request, effectDigest, "0/20")
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

	if _, abortErr := coordinator.Abort(t.Context(), AbortRequest{OperationID: request.OperationID, Reason: AbortSuperseded}); abortErr != ErrConflict {
		close(release)
		t.Fatalf("concurrent Abort error = %v, want ErrConflict", abortErr)
	}
	if countingProvider.abortCount(request.OperationID) != 0 {
		close(release)
		t.Fatalf("provider Abort calls = %d, want 0", countingProvider.abortCount(request.OperationID))
	}
	close(release)
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
			if repository.transactionCount() != transactionsBefore {
				t.Fatalf("transactions = %d, want unchanged %d", repository.transactionCount(), transactionsBefore)
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
	errors   map[uuid.UUID]error
	commits  map[uuid.UUID]int
	lsns     map[uuid.UUID]WALPosition
	resolves map[uuid.UUID]int
}

func newCoordinatorEffectResolver() *coordinatorEffectResolver {
	return &coordinatorEffectResolver{
		effects:  make(map[uuid.UUID]ResolvedEffect),
		errors:   make(map[uuid.UUID]error),
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
