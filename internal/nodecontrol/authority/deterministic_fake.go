package authority

import (
	"context"
	"math"
	"sort"
	"sync"

	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

type DeterministicProvider struct {
	mu           sync.Mutex
	epoch        uint64
	nextSequence uint64
	records      map[uuid.UUID]*deterministicRecord
	failures     map[Operation]Failure
}

type deterministicRecord struct {
	reserveRequest  ReserveRequest
	record          Record
	finalizeRequest *FinalizeRequest
	abortRequest    *AbortRequest
}

var _ Provider = (*DeterministicProvider)(nil)

func NewDeterministicProvider(epoch uint64) (*DeterministicProvider, error) {
	if !validPositiveCoordinate(epoch) {
		return nil, ErrInvalidArgument
	}
	return &DeterministicProvider{
		epoch:    epoch,
		records:  make(map[uuid.UUID]*deterministicRecord),
		failures: make(map[Operation]Failure),
	}, nil
}

func (provider *DeterministicProvider) Reserve(ctx context.Context, request ReserveRequest) (Reservation, error) {
	if err := contextError(ctx); err != nil {
		return Reservation{}, err
	}
	if provider == nil {
		return Reservation{}, ErrInvalidArgument
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	if err := contextError(ctx); err != nil {
		return Reservation{}, err
	}
	if provider.consumeBeforeMutationFailure(OperationReserve) {
		return Reservation{}, ErrInjectedFailure
	}
	if existing, ok := provider.records[request.OperationID]; ok {
		if existing.reserveRequest != request {
			return Reservation{}, ErrConflict
		}
		result := existing.record.Reservation
		return result, nil
	}
	if request.Validate() != nil || provider.nextSequence >= math.MaxInt64 {
		return Reservation{}, ErrInvalidArgument
	}
	reservation := Reservation{
		OperationID: request.OperationID,
		Kind:        request.Kind,
		ScopeKind:   request.ScopeKind,
		ScopeDigest: request.ScopeDigest,
		Epoch:       provider.epoch,
		Sequence:    provider.nextSequence + 1,
	}
	digest, err := reservationDigest(reservation)
	if err != nil {
		return Reservation{}, err
	}
	reservation.ReservationDigest = digest
	if reservation.Validate() != nil {
		return Reservation{}, ErrInvalidArgument
	}
	provider.nextSequence = reservation.Sequence
	provider.records[request.OperationID] = &deterministicRecord{
		reserveRequest: request,
		record:         Record{Reservation: reservation},
	}
	if provider.consumeAfterMutationFailure(OperationReserve) {
		return Reservation{}, ErrResponseLost
	}
	return reservation, nil
}

func (provider *DeterministicProvider) Finalize(ctx context.Context, request FinalizeRequest) (Receipt, error) {
	if err := contextError(ctx); err != nil {
		return Receipt{}, err
	}
	if provider == nil {
		return Receipt{}, ErrInvalidArgument
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	if err := contextError(ctx); err != nil {
		return Receipt{}, err
	}
	if provider.consumeBeforeMutationFailure(OperationFinalize) {
		return Receipt{}, ErrInjectedFailure
	}
	stored, ok := provider.records[request.OperationID]
	if !ok {
		if request.Validate() != nil {
			return Receipt{}, ErrInvalidArgument
		}
		return Receipt{}, ErrNotFound
	}
	if stored.record.TerminalReceipt != nil {
		if stored.record.TerminalReceipt.Status == StatusAborted {
			return Receipt{}, ErrTerminalConflict
		}
		normalized, normalizeErr := normalizeFinalizeRequest(request)
		if normalizeErr != nil || stored.finalizeRequest == nil || *stored.finalizeRequest != normalized {
			return Receipt{}, ErrConflict
		}
		result := cloneReceipt(*stored.record.TerminalReceipt)
		return result, nil
	}
	normalized, err := normalizeFinalizeRequest(request)
	if err != nil {
		return Receipt{}, err
	}
	effectDigest := normalized.EffectDigest
	databasePoint := DatabasePoint{
		SystemID:    normalized.DBSystemID,
		Timeline:    normalized.DBTimeline,
		RequiredLSN: normalized.RequiredLSN,
	}
	receipt := Receipt{
		Reservation:   stored.record.Reservation,
		EffectDigest:  &effectDigest,
		DatabasePoint: &databasePoint,
		Status:        StatusCommitted,
	}
	receiptDigestValue, err := receiptDigest(receipt)
	if err != nil {
		return Receipt{}, err
	}
	receipt.ReceiptDigest = receiptDigestValue
	if receipt.Validate() != nil {
		return Receipt{}, ErrInvalidArgument
	}
	stored.finalizeRequest = &normalized
	boundDigest := effectDigest
	boundPoint := databasePoint
	stored.record.BoundEffectDigest = &boundDigest
	stored.record.BoundDatabasePoint = &boundPoint
	terminal := cloneReceipt(receipt)
	stored.record.TerminalReceipt = &terminal
	if provider.consumeAfterMutationFailure(OperationFinalize) {
		return Receipt{}, ErrResponseLost
	}
	return cloneReceipt(receipt), nil
}

func (provider *DeterministicProvider) Abort(ctx context.Context, request AbortRequest) (Receipt, error) {
	if err := contextError(ctx); err != nil {
		return Receipt{}, err
	}
	if provider == nil {
		return Receipt{}, ErrInvalidArgument
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	if err := contextError(ctx); err != nil {
		return Receipt{}, err
	}
	if provider.consumeBeforeMutationFailure(OperationAbort) {
		return Receipt{}, ErrInjectedFailure
	}
	stored, ok := provider.records[request.OperationID]
	if !ok {
		if request.Validate() != nil {
			return Receipt{}, ErrInvalidArgument
		}
		return Receipt{}, ErrNotFound
	}
	if stored.record.TerminalReceipt != nil {
		if stored.record.TerminalReceipt.Status == StatusCommitted {
			return Receipt{}, ErrTerminalConflict
		}
		if stored.abortRequest == nil || *stored.abortRequest != request {
			return Receipt{}, ErrConflict
		}
		result := cloneReceipt(*stored.record.TerminalReceipt)
		return result, nil
	}
	if request.Validate() != nil {
		return Receipt{}, ErrInvalidArgument
	}
	reason := request.Reason
	receipt := Receipt{
		Reservation:   stored.record.Reservation,
		EffectDigest:  cloneDigestPointer(stored.record.BoundEffectDigest),
		DatabasePoint: cloneDatabasePointPointer(stored.record.BoundDatabasePoint),
		Status:        StatusAborted,
		AbortReason:   &reason,
	}
	receiptDigestValue, err := receiptDigest(receipt)
	if err != nil {
		return Receipt{}, err
	}
	receipt.ReceiptDigest = receiptDigestValue
	if receipt.Validate() != nil {
		return Receipt{}, ErrInvalidArgument
	}
	requestCopy := request
	stored.abortRequest = &requestCopy
	terminal := cloneReceipt(receipt)
	stored.record.TerminalReceipt = &terminal
	if provider.consumeAfterMutationFailure(OperationAbort) {
		return Receipt{}, ErrResponseLost
	}
	return cloneReceipt(receipt), nil
}

func (provider *DeterministicProvider) Inspect(ctx context.Context, operationID uuid.UUID) (Record, error) {
	if err := contextError(ctx); err != nil {
		return Record{}, err
	}
	if provider == nil {
		return Record{}, ErrInvalidArgument
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	if err := contextError(ctx); err != nil {
		return Record{}, err
	}
	if provider.consumeBeforeMutationFailure(OperationInspect) {
		return Record{}, ErrInjectedFailure
	}
	if !validOperationID(operationID) {
		return Record{}, ErrInvalidArgument
	}
	stored, ok := provider.records[operationID]
	if !ok {
		return Record{}, ErrNotFound
	}
	return cloneRecord(stored.record), nil
}

func (provider *DeterministicProvider) Head(ctx context.Context) (Head, error) {
	if err := contextError(ctx); err != nil {
		return Head{}, err
	}
	if provider == nil {
		return Head{}, ErrInvalidArgument
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	if err := contextError(ctx); err != nil {
		return Head{}, err
	}
	if provider.consumeBeforeMutationFailure(OperationHead) {
		return Head{}, ErrInjectedFailure
	}
	head := Head{Epoch: provider.epoch, LatestReservedSequence: provider.nextSequence}
	for _, stored := range provider.records {
		if stored.record.Sequence == provider.nextSequence {
			head.LatestReservationDigest = stored.record.ReservationDigest
		}
		if stored.record.TerminalReceipt == nil || stored.record.TerminalReceipt.Status != StatusCommitted ||
			stored.record.Sequence <= head.LatestCommittedSequence {
			continue
		}
		head.LatestCommittedSequence = stored.record.Sequence
		head.LatestCommittedOperationID = stored.record.OperationID
		head.LatestCommittedReceiptDigest = stored.record.TerminalReceipt.ReceiptDigest
		head.LatestCommittedDatabasePoint = cloneDatabasePointPointer(stored.record.TerminalReceipt.DatabasePoint)
	}
	if head.Validate() != nil {
		return Head{}, ErrInvalidArgument
	}
	return cloneHead(head), nil
}

func (provider *DeterministicProvider) CommittedNodeCheckpoint(ctx context.Context, nodeScopeDigest contracts.Digest) (NodeCheckpoint, error) {
	if err := contextError(ctx); err != nil {
		return NodeCheckpoint{}, err
	}
	if provider == nil {
		return NodeCheckpoint{}, ErrInvalidArgument
	}
	if nodeScopeDigest == (contracts.Digest{}) || nodeScopeDigest == globalNodeScopeDigest || nodeScopeDigest == globalOperatorScopeDigest {
		return NodeCheckpoint{}, ErrInvalidArgument
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	if err := contextError(ctx); err != nil {
		return NodeCheckpoint{}, err
	}
	checkpoint := NodeCheckpoint{AuthorityEpoch: provider.epoch}
	for _, stored := range provider.records {
		if stored.record.TerminalReceipt == nil || stored.record.TerminalReceipt.Status != StatusCommitted ||
			stored.record.Sequence <= checkpoint.Sequence {
			continue
		}
		matchesNode := stored.record.ScopeKind == ScopeNode && stored.record.ScopeDigest == nodeScopeDigest
		matchesGlobalNode := stored.record.ScopeKind == ScopeGlobalNodeTrust && stored.record.ScopeDigest == globalNodeScopeDigest
		if matchesNode || matchesGlobalNode {
			checkpoint.Sequence = stored.record.Sequence
			checkpoint.ReceiptDigest = stored.record.TerminalReceipt.ReceiptDigest
		}
	}
	if checkpoint.Validate() != nil {
		return NodeCheckpoint{}, ErrInvalidArgument
	}
	return checkpoint, nil
}

func (provider *DeterministicProvider) FailNext(operation Operation, failure Failure) error {
	if provider == nil || operation.Validate() != nil || failure.Validate() != nil ||
		(failure == FailureAfterMutationResponseLost && operation != OperationReserve && operation != OperationFinalize && operation != OperationAbort) {
		return ErrInvalidArgument
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	provider.failures[operation] = failure
	return nil
}

func (provider *DeterministicProvider) Snapshot() []Record {
	if provider == nil {
		return nil
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	result := make([]Record, 0, len(provider.records))
	for _, stored := range provider.records {
		result = append(result, cloneRecord(stored.record))
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].Sequence < result[right].Sequence
	})
	return result
}

func (provider *DeterministicProvider) consumeBeforeMutationFailure(operation Operation) bool {
	if provider.failures[operation] != FailureBeforeMutation {
		return false
	}
	delete(provider.failures, operation)
	return true
}

func (provider *DeterministicProvider) consumeAfterMutationFailure(operation Operation) bool {
	if provider.failures[operation] != FailureAfterMutationResponseLost {
		return false
	}
	delete(provider.failures, operation)
	return true
}

func normalizeFinalizeRequest(request FinalizeRequest) (FinalizeRequest, error) {
	if request.Validate() != nil {
		return FinalizeRequest{}, ErrInvalidArgument
	}
	canonical, err := canonicalWALPosition(request.RequiredLSN)
	if err != nil {
		return FinalizeRequest{}, ErrInvalidArgument
	}
	request.RequiredLSN = canonical
	return request, nil
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return ErrInvalidArgument
	}
	select {
	case <-ctx.Done():
		return ErrCanceled
	default:
		return nil
	}
}

func cloneDigestPointer(value *contracts.Digest) *contracts.Digest {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func cloneDatabasePointPointer(value *DatabasePoint) *DatabasePoint {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func cloneAbortReasonPointer(value *AbortReason) *AbortReason {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func cloneReceipt(value Receipt) Receipt {
	clone := value
	clone.EffectDigest = cloneDigestPointer(value.EffectDigest)
	clone.DatabasePoint = cloneDatabasePointPointer(value.DatabasePoint)
	clone.AbortReason = cloneAbortReasonPointer(value.AbortReason)
	return clone
}

func cloneRecord(value Record) Record {
	clone := value
	clone.BoundEffectDigest = cloneDigestPointer(value.BoundEffectDigest)
	clone.BoundDatabasePoint = cloneDatabasePointPointer(value.BoundDatabasePoint)
	if value.TerminalReceipt != nil {
		receipt := cloneReceipt(*value.TerminalReceipt)
		clone.TerminalReceipt = &receipt
	}
	return clone
}

func cloneHead(value Head) Head {
	clone := value
	clone.LatestCommittedDatabasePoint = cloneDatabasePointPointer(value.LatestCommittedDatabasePoint)
	return clone
}
