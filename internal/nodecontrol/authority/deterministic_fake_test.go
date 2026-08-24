package authority

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

func TestDeterministicProviderAllocatesOneTotalOrderUnderConcurrency(t *testing.T) {
	provider, err := NewDeterministicProvider(17)
	if err != nil {
		t.Fatal(err)
	}
	nodeScope := validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))
	const count = 1000
	results := make(chan Reservation, count)
	errorsFound := make(chan error, count)
	var wait sync.WaitGroup
	wait.Add(count)
	for index := 0; index < count; index++ {
		index := index
		go func() {
			defer wait.Done()
			operationID := uuid.MustParse(fmt.Sprintf("10000000-0000-4000-8000-%012x", index+1))
			reservation, reserveErr := provider.Reserve(t.Context(), ReserveRequest{
				OperationID: operationID, Kind: EffectDesiredActivate, ScopeKind: ScopeNode, ScopeDigest: nodeScope,
			})
			if reserveErr != nil {
				errorsFound <- reserveErr
				return
			}
			results <- reservation
		}()
	}
	wait.Wait()
	close(results)
	close(errorsFound)
	for reserveErr := range errorsFound {
		t.Fatalf("concurrent reserve: %v", reserveErr)
	}
	seen := make([]bool, count+1)
	for reservation := range results {
		if reservation.Epoch != 17 || reservation.Sequence == 0 || reservation.Sequence > count || seen[reservation.Sequence] {
			t.Fatalf("duplicate/out-of-range reservation: %#v", reservation)
		}
		seen[reservation.Sequence] = true
	}
	for sequence := 1; sequence <= count; sequence++ {
		if !seen[sequence] {
			t.Fatalf("missing sequence %d", sequence)
		}
	}
	head, err := provider.Head(t.Context())
	if err != nil || head.LatestReservedSequence != count || head.LatestCommittedSequence != 0 {
		t.Fatalf("head after concurrency = %#v, %v", head, err)
	}
}

func TestDeterministicProviderCommittedNodeCheckpointIsScoped(t *testing.T) {
	provider, err := NewDeterministicProvider(23)
	if err != nil {
		t.Fatal(err)
	}
	nodeA := validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))
	nodeB := validNodeScopeDigest(t, uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"))
	globalNode := sha256Digest("TALENRO-GLOBAL-NODE-TRUST-SCOPE-V1\x00global_node_trust")
	globalOperator := sha256Digest("TALENRO-GLOBAL-OPERATOR-TRUST-SCOPE-V1\x00global_operator_trust")

	commit := func(operationID string, kind EffectKind, scope ScopeKind, digest contracts.Digest, effect byte) Receipt {
		t.Helper()
		reservation, reserveErr := provider.Reserve(t.Context(), ReserveRequest{
			OperationID: uuid.MustParse(operationID), Kind: kind, ScopeKind: scope, ScopeDigest: digest,
		})
		if reserveErr != nil {
			t.Fatal(reserveErr)
		}
		receipt, finalizeErr := provider.Finalize(t.Context(), FinalizeRequest{
			OperationID: reservation.OperationID, EffectDigest: contracts.Digest{effect}, DBSystemID: 1,
			DBTimeline: 1, RequiredLSN: WALPosition(fmt.Sprintf("0/%X", effect)),
		})
		if finalizeErr != nil {
			t.Fatal(finalizeErr)
		}
		return receipt
	}

	receiptA := commit("10000000-0000-4000-8000-000000000001", EffectDesiredActivate, ScopeNode, nodeA, 1)
	checkpointA := mustCheckpoint(t, provider, nodeA)
	checkpointB := mustCheckpoint(t, provider, nodeB)
	if checkpointA.Sequence != receiptA.Sequence || checkpointA.ReceiptDigest != receiptA.ReceiptDigest || checkpointB != (NodeCheckpoint{AuthorityEpoch: 23}) {
		t.Fatalf("initial checkpoints A=%#v B=%#v", checkpointA, checkpointB)
	}

	receiptB := commit("10000000-0000-4000-8000-000000000002", EffectDesiredActivate, ScopeNode, nodeB, 2)
	if got := mustCheckpoint(t, provider, nodeA); got.Sequence != receiptA.Sequence {
		t.Fatalf("node B advanced A: %#v", got)
	}
	if got := mustCheckpoint(t, provider, nodeB); got.Sequence != receiptB.Sequence || got.ReceiptDigest != receiptB.ReceiptDigest {
		t.Fatalf("node B checkpoint = %#v", got)
	}

	pending, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse("10000000-0000-4000-8000-000000000003"), Kind: EffectRecoveryActivate,
		ScopeKind: ScopeNode, ScopeDigest: nodeA,
	})
	if err != nil {
		t.Fatal(err)
	}
	aborted, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse("10000000-0000-4000-8000-000000000004"), Kind: EffectOperatorTransition,
		ScopeKind: ScopeNode, ScopeDigest: nodeA,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Abort(t.Context(), AbortRequest{OperationID: aborted.OperationID, Reason: AbortSuperseded}); err != nil {
		t.Fatal(err)
	}
	operatorReceipt := commit("10000000-0000-4000-8000-000000000005", EffectOperatorAuthorizerChange, ScopeGlobalOperatorTrust, globalOperator, 5)
	if got := mustCheckpoint(t, provider, nodeA); got.Sequence != receiptA.Sequence {
		t.Fatalf("pending/aborted/operator-global advanced A: %#v; pending=%#v operator=%#v", got, pending, operatorReceipt)
	}

	globalReceipt := commit("10000000-0000-4000-8000-000000000006", EffectRootPublish, ScopeGlobalNodeTrust, globalNode, 6)
	for name, digest := range map[string]contracts.Digest{"A": nodeA, "B": nodeB} {
		got := mustCheckpoint(t, provider, digest)
		if got.Sequence != globalReceipt.Sequence || got.ReceiptDigest != globalReceipt.ReceiptDigest {
			t.Fatalf("global node trust did not advance %s: %#v", name, got)
		}
	}
}

func TestDeterministicProviderOutOfOrderFinalizationRetainsHighestReservationAnchors(t *testing.T) {
	provider, err := NewDeterministicProvider(27)
	if err != nil {
		t.Fatal(err)
	}
	nodeScope := validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))
	lower := mustReserve(t, provider, "17000000-0000-4000-8000-000000000001", EffectDesiredActivate, ScopeNode, nodeScope)
	higher := mustReserve(t, provider, "17000000-0000-4000-8000-000000000002", EffectRecoveryActivate, ScopeNode, nodeScope)

	higherReceipt, err := provider.Finalize(t.Context(), FinalizeRequest{
		OperationID: higher.OperationID, EffectDigest: contracts.Digest{2}, DBSystemID: 22,
		DBTimeline: 2, RequiredLSN: "0/2",
	})
	if err != nil {
		t.Fatal(err)
	}
	lowerReceipt, err := provider.Finalize(t.Context(), FinalizeRequest{
		OperationID: lower.OperationID, EffectDigest: contracts.Digest{1}, DBSystemID: 11,
		DBTimeline: 1, RequiredLSN: "0/1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if lowerReceipt.Sequence >= higherReceipt.Sequence {
		t.Fatalf("test setup did not finalize a lower reservation last: lower=%#v higher=%#v", lowerReceipt, higherReceipt)
	}

	t.Run("head", func(t *testing.T) {
		head, err := provider.Head(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		if head.LatestReservedSequence != higherReceipt.Sequence ||
			head.LatestReservationDigest != higherReceipt.ReservationDigest ||
			head.LatestCommittedSequence != higherReceipt.Sequence ||
			head.LatestCommittedOperationID != higherReceipt.OperationID ||
			head.LatestCommittedReceiptDigest != higherReceipt.ReceiptDigest ||
			head.LatestCommittedDatabasePoint == nil || higherReceipt.DatabasePoint == nil ||
			*head.LatestCommittedDatabasePoint != *higherReceipt.DatabasePoint {
			t.Fatalf("late lower finalization replaced maximum committed reservation anchors: head=%#v higher=%#v", head, higherReceipt)
		}
	})
	t.Run("node checkpoint", func(t *testing.T) {
		checkpoint := mustCheckpoint(t, provider, nodeScope)
		if checkpoint.AuthorityEpoch != higherReceipt.Epoch || checkpoint.Sequence != higherReceipt.Sequence ||
			checkpoint.ReceiptDigest != higherReceipt.ReceiptDigest {
			t.Fatalf("late lower finalization replaced maximum node checkpoint: %#v, want anchors from %#v", checkpoint, higherReceipt)
		}
	})
}

func TestDeterministicProviderCancellationNeverMutates(t *testing.T) {
	provider, err := NewDeterministicProvider(29)
	if err != nil {
		t.Fatal(err)
	}
	if err := provider.FailNext(OperationReserve, FailureBeforeMutation); err != nil {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	request := ReserveRequest{
		OperationID: uuid.MustParse("20000000-0000-4000-8000-000000000001"), Kind: EffectDesiredActivate,
		ScopeKind: ScopeNode, ScopeDigest: validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")),
	}
	if _, err := provider.Reserve(canceled, request); !errors.Is(err, ErrCanceled) {
		t.Fatalf("canceled reserve error = %v", err)
	}
	if records := provider.Snapshot(); len(records) != 0 {
		t.Fatalf("canceled reserve mutated state: %#v", records)
	}
	if _, err := provider.Reserve(t.Context(), request); !errors.Is(err, ErrInjectedFailure) {
		t.Fatalf("cancellation consumed fault or failed to preserve state: %v", err)
	}
	reservation, err := provider.Reserve(t.Context(), request)
	if err != nil || reservation.Sequence != 1 {
		t.Fatalf("reserve after cancellation/fault = %#v, %v", reservation, err)
	}
	if _, err := provider.Finalize(canceled, FinalizeRequest{OperationID: reservation.OperationID, EffectDigest: contracts.Digest{1}, DBSystemID: 1, DBTimeline: 1, RequiredLSN: "0/1"}); !errors.Is(err, ErrCanceled) {
		t.Fatalf("canceled finalize error = %v", err)
	}
	if _, err := provider.Abort(canceled, AbortRequest{OperationID: reservation.OperationID, Reason: AbortSuperseded}); !errors.Is(err, ErrCanceled) {
		t.Fatalf("canceled abort error = %v", err)
	}
	record, err := provider.Inspect(t.Context(), reservation.OperationID)
	if err != nil || record.TerminalReceipt != nil || record.BoundEffectDigest != nil || record.BoundDatabasePoint != nil {
		t.Fatalf("canceled terminal calls mutated record: %#v, %v", record, err)
	}
	if _, err := provider.Inspect(canceled, reservation.OperationID); !errors.Is(err, ErrCanceled) {
		t.Fatalf("canceled inspect error = %v", err)
	}
	if _, err := provider.Head(canceled); !errors.Is(err, ErrCanceled) {
		t.Fatalf("canceled head error = %v", err)
	}
	if _, err := provider.CommittedNodeCheckpoint(canceled, request.ScopeDigest); !errors.Is(err, ErrCanceled) {
		t.Fatalf("canceled checkpoint error = %v", err)
	}
}

func TestDeterministicProviderTerminalRetriesAndFieldConflicts(t *testing.T) {
	provider, err := NewDeterministicProvider(31)
	if err != nil {
		t.Fatal(err)
	}
	nodeScope := validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))
	committedReservation := mustReserve(t, provider, "30000000-0000-4000-8000-000000000001", EffectDesiredActivate, ScopeNode, nodeScope)
	finalize := FinalizeRequest{OperationID: committedReservation.OperationID, EffectDigest: contracts.Digest{1}, DBSystemID: 1, DBTimeline: 1, RequiredLSN: "00/0A"}
	firstCommit, err := provider.Finalize(t.Context(), finalize)
	if err != nil {
		t.Fatal(err)
	}
	finalize.RequiredLSN = "0/A"
	retryCommit, err := provider.Finalize(t.Context(), finalize)
	if err != nil || !reflect.DeepEqual(firstCommit, retryCommit) {
		t.Fatalf("canonical finalize retry = %#v, %v; want %#v", retryCommit, err, firstCommit)
	}
	finalize.EffectDigest = contracts.Digest{2}
	if _, err := provider.Finalize(t.Context(), finalize); !errors.Is(err, ErrConflict) {
		t.Fatalf("changed finalize error = %v", err)
	}

	abortedReservation := mustReserve(t, provider, "30000000-0000-4000-8000-000000000002", EffectRecoveryActivate, ScopeNode, nodeScope)
	firstAbort, err := provider.Abort(t.Context(), AbortRequest{OperationID: abortedReservation.OperationID, Reason: AbortValidationFailed})
	if err != nil {
		t.Fatal(err)
	}
	retryAbort, err := provider.Abort(t.Context(), AbortRequest{OperationID: abortedReservation.OperationID, Reason: AbortValidationFailed})
	if err != nil || !reflect.DeepEqual(firstAbort, retryAbort) {
		t.Fatalf("abort retry = %#v, %v; want %#v", retryAbort, err, firstAbort)
	}
	if _, err := provider.Abort(t.Context(), AbortRequest{OperationID: abortedReservation.OperationID, Reason: AbortSuperseded}); !errors.Is(err, ErrConflict) {
		t.Fatalf("changed abort error = %v", err)
	}
	if _, err := provider.Finalize(t.Context(), FinalizeRequest{OperationID: abortedReservation.OperationID, EffectDigest: contracts.Digest{2}, DBSystemID: 1, DBTimeline: 1, RequiredLSN: "0/2"}); !errors.Is(err, ErrTerminalConflict) {
		t.Fatalf("finalize after abort error = %v", err)
	}
	unknown := uuid.MustParse("30000000-0000-4000-8000-000000000003")
	if _, err := provider.Inspect(t.Context(), unknown); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown inspect error = %v", err)
	}
}

func TestDeterministicProviderFaultInjectionAndSnapshotIsolation(t *testing.T) {
	provider, err := NewDeterministicProvider(37)
	if err != nil {
		t.Fatal(err)
	}
	if err := provider.FailNext(Operation("checkpoint"), FailureBeforeMutation); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("unknown fault operation error = %v", err)
	}
	if err := provider.FailNext(OperationHead, FailureAfterMutationResponseLost); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("head after-mutation fault error = %v", err)
	}
	if err := provider.FailNext(OperationInspect, Failure("other")); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("unknown failure error = %v", err)
	}

	nodeScope := validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))
	if err := provider.FailNext(OperationReserve, FailureAfterMutationResponseLost); err != nil {
		t.Fatal(err)
	}
	lostRequest := ReserveRequest{
		OperationID: uuid.MustParse("40000000-0000-4000-8000-000000000001"), Kind: EffectDesiredActivate,
		ScopeKind: ScopeNode, ScopeDigest: nodeScope,
	}
	if _, err := provider.Reserve(t.Context(), lostRequest); !errors.Is(err, ErrResponseLost) {
		t.Fatalf("lost reserve error = %v", err)
	}
	lostReservation, err := provider.Reserve(t.Context(), lostRequest)
	if err != nil || lostReservation.Sequence != 1 {
		t.Fatalf("reserve recovery = %#v, %v", lostReservation, err)
	}

	if err := provider.FailNext(OperationFinalize, FailureAfterMutationResponseLost); err != nil {
		t.Fatal(err)
	}
	finalize := FinalizeRequest{OperationID: lostReservation.OperationID, EffectDigest: contracts.Digest{1}, DBSystemID: 1, DBTimeline: 1, RequiredLSN: "0/1"}
	if _, err := provider.Finalize(t.Context(), finalize); !errors.Is(err, ErrResponseLost) {
		t.Fatalf("lost finalize error = %v", err)
	}
	committed, err := provider.Finalize(t.Context(), finalize)
	if err != nil || committed.Status != StatusCommitted {
		t.Fatalf("finalize recovery = %#v, %v", committed, err)
	}

	abortReservation := mustReserve(t, provider, "40000000-0000-4000-8000-000000000002", EffectRecoveryActivate, ScopeNode, nodeScope)
	if err := provider.FailNext(OperationAbort, FailureAfterMutationResponseLost); err != nil {
		t.Fatal(err)
	}
	abortRequest := AbortRequest{OperationID: abortReservation.OperationID, Reason: AbortProviderDependencyFailed}
	if _, err := provider.Abort(t.Context(), abortRequest); !errors.Is(err, ErrResponseLost) {
		t.Fatalf("lost abort error = %v", err)
	}
	aborted, err := provider.Abort(t.Context(), abortRequest)
	if err != nil || aborted.Status != StatusAborted {
		t.Fatalf("abort recovery = %#v, %v", aborted, err)
	}

	if err := provider.FailNext(OperationInspect, FailureBeforeMutation); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Inspect(t.Context(), lostReservation.OperationID); !errors.Is(err, ErrInjectedFailure) {
		t.Fatalf("inspect fault error = %v", err)
	}
	if err := provider.FailNext(OperationHead, FailureBeforeMutation); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Head(t.Context()); !errors.Is(err, ErrInjectedFailure) {
		t.Fatalf("head fault error = %v", err)
	}

	snapshot := provider.Snapshot()
	if len(snapshot) != 2 || !sort.SliceIsSorted(snapshot, func(i, j int) bool { return snapshot[i].Sequence < snapshot[j].Sequence }) {
		t.Fatalf("snapshot order = %#v", snapshot)
	}
	wantFirst, err := provider.Inspect(t.Context(), lostReservation.OperationID)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot[0].BoundEffectDigest == nil || snapshot[0].BoundDatabasePoint == nil || snapshot[0].TerminalReceipt == nil ||
		snapshot[0].TerminalReceipt.EffectDigest == nil || snapshot[0].TerminalReceipt.DatabasePoint == nil {
		t.Fatalf("snapshot missing committed pointers: %#v", snapshot[0])
	}
	*snapshot[0].BoundEffectDigest = contracts.Digest{9}
	snapshot[0].BoundDatabasePoint.RequiredLSN = "BAD"
	*snapshot[0].TerminalReceipt.EffectDigest = contracts.Digest{8}
	snapshot[0].TerminalReceipt.DatabasePoint.RequiredLSN = "WORSE"
	if snapshot[1].TerminalReceipt == nil || snapshot[1].TerminalReceipt.AbortReason == nil {
		t.Fatalf("snapshot missing abort pointer: %#v", snapshot[1])
	}
	*snapshot[1].TerminalReceipt.AbortReason = AbortValidationFailed
	gotFirst, err := provider.Inspect(t.Context(), lostReservation.OperationID)
	if err != nil || !reflect.DeepEqual(gotFirst, wantFirst) {
		t.Fatalf("snapshot mutation escaped: %#v, %v; want %#v", gotFirst, err, wantFirst)
	}
	head, err := provider.Head(t.Context())
	if err != nil || head.LatestCommittedDatabasePoint == nil {
		t.Fatalf("head = %#v, %v", head, err)
	}
	head.LatestCommittedDatabasePoint.RequiredLSN = "BAD"
	retryHead, err := provider.Head(t.Context())
	if err != nil || retryHead.LatestCommittedDatabasePoint == nil || retryHead.LatestCommittedDatabasePoint.RequiredLSN != "0/1" {
		t.Fatalf("head pointer mutation escaped: %#v, %v", retryHead, err)
	}
}

func TestDeterministicProviderAfterMutationFaultWaitsForMutation(t *testing.T) {
	provider, err := NewDeterministicProvider(41)
	if err != nil {
		t.Fatal(err)
	}
	nodeScope := validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))
	firstRequest := ReserveRequest{
		OperationID: uuid.MustParse("50000000-0000-4000-8000-000000000001"), Kind: EffectDesiredActivate,
		ScopeKind: ScopeNode, ScopeDigest: nodeScope,
	}
	first, err := provider.Reserve(t.Context(), firstRequest)
	if err != nil {
		t.Fatal(err)
	}
	if err := provider.FailNext(OperationReserve, FailureAfterMutationResponseLost); err != nil {
		t.Fatal(err)
	}
	if retry, err := provider.Reserve(t.Context(), firstRequest); err != nil || retry != first {
		t.Fatalf("non-mutating reserve retry consumed after-mutation fault: %#v, %v", retry, err)
	}
	secondRequest := firstRequest
	secondRequest.OperationID = uuid.MustParse("50000000-0000-4000-8000-000000000002")
	if _, err := provider.Reserve(t.Context(), secondRequest); !errors.Is(err, ErrResponseLost) {
		t.Fatalf("next mutating reserve did not receive response-loss fault: %v", err)
	}
	second, err := provider.Reserve(t.Context(), secondRequest)
	if err != nil || second.Sequence != 2 {
		t.Fatalf("lost reserve recovery = %#v, %v", second, err)
	}

	if err := provider.FailNext(OperationFinalize, FailureAfterMutationResponseLost); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Finalize(t.Context(), FinalizeRequest{OperationID: first.OperationID, DBSystemID: 1, DBTimeline: 1, RequiredLSN: "0/1"}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("invalid finalize error = %v", err)
	}
	finalize := FinalizeRequest{OperationID: first.OperationID, EffectDigest: contracts.Digest{1}, DBSystemID: 1, DBTimeline: 1, RequiredLSN: "0/1"}
	if _, err := provider.Finalize(t.Context(), finalize); !errors.Is(err, ErrResponseLost) {
		t.Fatalf("next mutating finalize did not receive response-loss fault: %v", err)
	}
	if receipt, err := provider.Finalize(t.Context(), finalize); err != nil || receipt.Status != StatusCommitted {
		t.Fatalf("lost finalize recovery = %#v, %v", receipt, err)
	}

	third := mustReserve(t, provider, "50000000-0000-4000-8000-000000000003", EffectRecoveryActivate, ScopeNode, nodeScope)
	abort := AbortRequest{OperationID: third.OperationID, Reason: AbortSuperseded}
	firstAbort, err := provider.Abort(t.Context(), abort)
	if err != nil {
		t.Fatal(err)
	}
	if err := provider.FailNext(OperationAbort, FailureAfterMutationResponseLost); err != nil {
		t.Fatal(err)
	}
	if retry, err := provider.Abort(t.Context(), abort); err != nil || !reflect.DeepEqual(retry, firstAbort) {
		t.Fatalf("non-mutating abort retry consumed after-mutation fault: %#v, %v", retry, err)
	}
	fourth := mustReserve(t, provider, "50000000-0000-4000-8000-000000000004", EffectRecoveryActivate, ScopeNode, nodeScope)
	fourthAbort := AbortRequest{OperationID: fourth.OperationID, Reason: AbortSuperseded}
	if _, err := provider.Abort(t.Context(), fourthAbort); !errors.Is(err, ErrResponseLost) {
		t.Fatalf("next mutating abort did not receive response-loss fault: %v", err)
	}
	if receipt, err := provider.Abort(t.Context(), fourthAbort); err != nil || receipt.Status != StatusAborted {
		t.Fatalf("lost abort recovery = %#v, %v", receipt, err)
	}
}

func TestDeterministicProviderValidatesMalformedUnknownRequests(t *testing.T) {
	provider, err := NewDeterministicProvider(43)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Finalize(t.Context(), FinalizeRequest{}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("malformed unknown finalize error = %v", err)
	}
	if _, err := provider.Abort(t.Context(), AbortRequest{}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("malformed unknown abort error = %v", err)
	}
}

func mustReserve(t *testing.T, provider *DeterministicProvider, operationID string, kind EffectKind, scope ScopeKind, digest contracts.Digest) Reservation {
	t.Helper()
	reservation, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse(operationID), Kind: kind, ScopeKind: scope, ScopeDigest: digest,
	})
	if err != nil {
		t.Fatal(err)
	}
	return reservation
}

func mustCheckpoint(t *testing.T, provider *DeterministicProvider, digest contracts.Digest) NodeCheckpoint {
	t.Helper()
	checkpoint, err := provider.CommittedNodeCheckpoint(t.Context(), digest)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkpoint.Validate(); err != nil {
		t.Fatalf("checkpoint validation: %v", err)
	}
	return checkpoint
}

func sha256Digest(value string) contracts.Digest {
	return sha256.Sum256([]byte(value))
}
