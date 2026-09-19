//go:build integration

package testinfra

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"talenro.local/platform/internal/store"
)

type C12AuthorityPITRError string

func (e C12AuthorityPITRError) Error() string { return string(e) }

const (
	C12PITRInvalidHandle     C12AuthorityPITRError = "pitr invalid handle"
	C12PITRWrongPhase        C12AuthorityPITRError = "pitr wrong phase"
	C12PITRNotObserved       C12AuthorityPITRError = "pitr not observed"
	C12PITRCapacityExceeded  C12AuthorityPITRError = "pitr capacity exceeded"
	C12PITRIndeterminate     C12AuthorityPITRError = "pitr indeterminate"
	C12PITRCanceled          C12AuthorityPITRError = "pitr canceled"
	C12PITRDependencyFailure C12AuthorityPITRError = "pitr dependency failure"
)

type C12AuthorityAccess interface {
	store.DBTX
	Begin(context.Context) (C12AuthorityTransaction, error)
}

type C12AuthorityTransaction interface {
	store.DBTX
	Commit(context.Context) error
	Rollback(context.Context) error
}

type c12CandidateBinding struct {
	owner         *c12AuthorityPITRState
	runGeneration uint64
	index         uint8
}

type c12ReadBudget struct {
	rows  int64
	bytes int64
}

func (b *c12ReadBudget) take(n int64) error {
	if n < 0 || n > 1<<20 || b.rows >= 65536 || b.bytes > (64<<20)-n {
		return C12PITRCapacityExceeded
	}
	b.rows++
	b.bytes += n
	return nil
}

func (b *c12ReadBudget) takeBytes(n int64) error {
	if n < 0 || b.bytes > (64<<20)-n {
		return C12PITRCapacityExceeded
	}
	b.bytes += n
	return nil
}

type c12AccessOperation uint8

const (
	c12AccessOperationExec c12AccessOperation = iota + 1
	c12AccessOperationQuery
	c12AccessOperationBegin
)

type c12AccessPolicy func(*c12CandidateBinding, bool, c12AccessOperation, string) bool

type c12AccessDriver interface {
	store.DBTX
	Commit(context.Context) error
	Rollback(context.Context) error
}

type c12AccessBackend interface {
	store.DBTX
	BeginTx(context.Context, pgx.TxOptions) (c12AccessDriver, error)
	Acquire(context.Context) error
	Release()
	Interrupt(context.Context) error
	Close(context.Context) error
	TypeMap() *pgtype.Map
}

type c12AccessOpener func(context.Context, *c12CandidateBinding) (c12AccessBackend, error)

type c12AccessLease struct {
	owner      *c12AuthorityPITRState
	binding    *c12CandidateBinding
	generation uint64
	ctx        context.Context
	cancel     context.CancelFunc
	live       atomic.Bool
	reason     atomic.Uint32
	budget     c12ReadBudget
	budgetMu   sync.Mutex
	revokeMu   sync.Mutex

	mu         sync.Mutex
	direct     c12AccessBackend
	txBackend  c12AccessBackend
	activeTx   *c12AuthorityTx
	txStarting bool
	cleanup    sync.Once
}

type c12AuthorityAccess struct {
	lease *c12AccessLease
}

type c12AuthorityTx struct {
	driver      c12AccessDriver
	backend     c12AccessBackend
	lease       *c12AccessLease
	generation  uint64
	terminal    atomic.Uint32
	operation   chan struct{}
	ownsBackend atomic.Bool
}

const (
	c12TxIdle uint32 = iota
	c12TxCommitting
	c12TxRollingBack
	c12TxOperation
	c12TxRevokedOperation
	c12TxRevoked
)

type c12PGXAccessBackend struct {
	conn      *pgx.Conn
	transport net.Conn
	ownership c12BackendOwnership
}

type c12BackendOwnership struct {
	once   sync.Once
	permit chan struct{}
	active atomic.Bool
}

func (ownership *c12BackendOwnership) acquire(ctx context.Context) error {
	ownership.once.Do(func() {
		ownership.permit = make(chan struct{}, 1)
		ownership.permit <- struct{}{}
	})
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ownership.permit:
	}
	if err := ctx.Err(); err != nil {
		ownership.permit <- struct{}{}
		return err
	}
	ownership.active.Store(true)
	return nil
}

func (ownership *c12BackendOwnership) release() {
	ownership.active.Store(false)
	ownership.permit <- struct{}{}
}

func (b *c12PGXAccessBackend) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return b.conn.Exec(ctx, sql, args...)
}

func (b *c12PGXAccessBackend) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return b.conn.Query(ctx, sql, args...)
}

func (b *c12PGXAccessBackend) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return b.conn.QueryRow(ctx, sql, args...)
}

func (b *c12PGXAccessBackend) BeginTx(ctx context.Context, options pgx.TxOptions) (c12AccessDriver, error) {
	return b.conn.BeginTx(ctx, options)
}

func (b *c12PGXAccessBackend) Acquire(ctx context.Context) error { return b.ownership.acquire(ctx) }
func (b *c12PGXAccessBackend) Release()                          { b.ownership.release() }
func (b *c12PGXAccessBackend) Interrupt(ctx context.Context) error {
	if !b.ownership.active.Load() {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return b.transport.Close()
}
func (b *c12PGXAccessBackend) Close(ctx context.Context) error { return b.conn.Close(ctx) }
func (b *c12PGXAccessBackend) TypeMap() *pgtype.Map            { return b.conn.TypeMap() }

func (state *c12AuthorityPITRState) openC12AccessBackend(ctx context.Context, binding *c12CandidateBinding) (c12AccessBackend, error) {
	if state.databaseURL == nil {
		return nil, C12PITRDependencyFailure
	}
	connectionURL := *state.databaseURL
	if binding != nil {
		state.mu.Lock()
		port := state.candidates[binding.index].port
		state.mu.Unlock()
		if port < 1 || port > 65535 {
			return nil, C12PITRDependencyFailure
		}
		connectionURL.Host = "127.0.0.1:" + strconv.Itoa(port)
	}
	config, err := pgx.ParseConfig(connectionURL.String())
	if err != nil {
		return nil, C12PITRDependencyFailure
	}
	connection, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, c12AccessError(err, false)
	}
	return &c12PGXAccessBackend{conn: connection, transport: connection.PgConn().Conn()}, nil
}

func (state *c12AuthorityPITRState) c12AccessOpener() c12AccessOpener {
	if state.accessOpen != nil {
		return state.accessOpen
	}
	return state.openC12AccessBackend
}

func (controller C12AuthorityPITRController) WithPrimaryAuthorityAccess(ctx context.Context, callback func(C12AuthorityAccess) error) error {
	if controller.state == nil {
		return C12PITRInvalidHandle
	}
	return c12WithAccess(ctx, controller.state, nil, callback)
}

func (controller C12AuthorityPITRController) WithCandidateAuthorityAccess(ctx context.Context, candidate C12AuthorityPITRCandidate, callback func(C12AuthorityAccess) error) error {
	if controller.state == nil {
		return C12PITRInvalidHandle
	}
	return c12WithAccess(ctx, controller.state, &candidate.binding, callback)
}

func c12WithAccess(ctx context.Context, state *c12AuthorityPITRState, binding *c12CandidateBinding, callback func(C12AuthorityAccess) error) (result error) {
	if ctx == nil || state == nil || callback == nil || state.runGeneration == 0 {
		return C12PITRInvalidHandle
	}
	state.accessGate.Lock()
	if state.accessActive || state.transitioning || !state.c12AccessPhaseValid(binding) {
		state.accessGate.Unlock()
		return C12PITRWrongPhase
	}
	if binding != nil && (binding.owner != state || binding.runGeneration != state.runGeneration || binding.index >= 8) {
		state.accessGate.Unlock()
		return C12PITRInvalidHandle
	}
	state.accessActive = true
	state.accessGate.Unlock()

	leaseContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	lease := &c12AccessLease{owner: state, binding: binding, generation: state.accessGeneration.Add(1), ctx: leaseContext, cancel: cancel}
	lease.live.Store(true)
	guardianDone := make(chan struct{})
	go func() {
		<-leaseContext.Done()
		lease.revoke(C12PITRCanceled)
		close(guardianDone)
	}()

	defer func() {
		wasCanceled := leaseContext.Err() != nil
		reason := C12PITRWrongPhase
		if wasCanceled {
			reason = C12PITRCanceled
		}
		if recovered := recover(); recovered != nil {
			result = C12PITRDependencyFailure
		}
		lease.revoke(reason)
		<-guardianDone
		state.accessGate.Lock()
		state.accessActive = false
		state.accessGate.Unlock()
		if wasCanceled && (result == nil || errors.Is(result, context.Canceled) || errors.Is(result, context.DeadlineExceeded) || errors.Is(result, C12PITRCanceled)) {
			result = C12PITRCanceled
		} else {
			result = c12AccessError(result, false)
		}
	}()

	result = callback(&c12AuthorityAccess{lease: lease})
	return result
}

func (state *c12AuthorityPITRState) c12AccessPhaseValid(binding *c12CandidateBinding) bool {
	if binding == nil {
		phase := state.phase.Load()
		return phase == 1 || phase == 2
	}
	if binding.owner != state || binding.runGeneration != state.runGeneration || binding.index >= 8 {
		return true
	}
	phase := state.candidatePhase[binding.index].Load()
	return phase >= 3 && phase <= 5
}

func (lease *c12AccessLease) revoke(reason C12AuthorityPITRError) {
	lease.revokeMu.Lock()
	if !lease.live.Load() {
		lease.revokeMu.Unlock()
		return
	}
	lease.reason.Store(c12AccessReasonCode(reason))
	lease.live.Store(false)
	lease.revokeMu.Unlock()
	lease.cancel()
	lease.cleanup.Do(func() {
		cleanupContext, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		lease.mu.Lock()
		activeTx, direct, txBackend := lease.activeTx, lease.direct, lease.txBackend
		lease.mu.Unlock()
		if activeTx != nil {
			activeTx.markRevoked()
		}
		if direct != nil {
			_ = direct.Interrupt(cleanupContext)
		}
		if txBackend != nil && txBackend != direct {
			_ = txBackend.Interrupt(cleanupContext)
		}
		if txBackend != nil && txBackend != direct {
			if activeTx != nil && activeTx.backend == txBackend {
				activeTx.cleanupRollback(cleanupContext)
			}
			lease.closeBackend(cleanupContext, txBackend)
		}
		if direct != nil {
			if activeTx != nil && activeTx.backend == direct {
				activeTx.cleanupRollback(cleanupContext)
			}
			lease.closeBackend(cleanupContext, direct)
		}
	})
}

func (lease *c12AccessLease) closeBackend(ctx context.Context, backend c12AccessBackend) {
	if err := backend.Acquire(ctx); err != nil {
		return
	}
	defer backend.Release()
	_ = backend.Close(ctx)
}

func (lease *c12AccessLease) status() error {
	if lease == nil || lease.owner == nil || lease.generation == 0 {
		return C12PITRInvalidHandle
	}
	if !lease.live.Load() {
		return c12AccessReason(lease.reason.Load())
	}
	if lease.ctx.Err() != nil {
		lease.revoke(C12PITRCanceled)
		return C12PITRCanceled
	}
	return nil
}

func (lease *c12AccessLease) authorize(inTransaction bool, operation c12AccessOperation, sql string) error {
	if err := lease.status(); err != nil {
		return err
	}
	if lease.owner.accessPolicy == nil || !lease.owner.accessPolicy(lease.binding, inTransaction, operation, sql) {
		return C12PITRDependencyFailure
	}
	return nil
}

func (lease *c12AccessLease) operationContext(caller context.Context) (context.Context, context.CancelFunc, error) {
	if caller == nil {
		return nil, nil, C12PITRInvalidHandle
	}
	if err := lease.status(); err != nil {
		return nil, nil, err
	}
	if caller.Err() != nil {
		return nil, nil, C12PITRCanceled
	}
	operationContext, cancel := context.WithCancel(lease.ctx)
	stop := context.AfterFunc(caller, cancel)
	return operationContext, func() { stop(); cancel() }, nil
}

func (lease *c12AccessLease) backend(ctx context.Context, transaction bool) (c12AccessBackend, error) {
	if err := lease.status(); err != nil {
		return nil, err
	}
	lease.mu.Lock()
	defer lease.mu.Unlock()
	if !lease.live.Load() {
		return nil, c12AccessReason(lease.reason.Load())
	}
	current := lease.direct
	if transaction {
		current = lease.txBackend
	}
	if current != nil {
		return current, nil
	}
	opened, err := lease.owner.c12AccessOpener()(ctx, lease.binding)
	if err != nil {
		return nil, c12AccessError(err, false)
	}
	if !lease.live.Load() || lease.ctx.Err() != nil {
		cleanupContext, cancel := context.WithTimeout(context.Background(), time.Second)
		_ = opened.Close(cleanupContext)
		cancel()
		return nil, C12PITRCanceled
	}
	if transaction {
		lease.txBackend = opened
	} else {
		lease.direct = opened
	}
	return opened, nil
}

func (lease *c12AccessLease) takeRow(bytes int64) error {
	lease.budgetMu.Lock()
	defer lease.budgetMu.Unlock()
	return lease.budget.take(bytes)
}

func (lease *c12AccessLease) takeBytes(bytes int64) error {
	lease.budgetMu.Lock()
	defer lease.budgetMu.Unlock()
	return lease.budget.takeBytes(bytes)
}

func (lease *c12AccessLease) finishTransaction(transaction *c12AuthorityTx) {
	lease.mu.Lock()
	if lease.activeTx == transaction {
		lease.activeTx = nil
	}
	lease.mu.Unlock()
}

func c12RejectPGXControlArguments(args []any) error {
	for _, arg := range args {
		switch arg.(type) {
		case pgx.QueryRewriter, pgx.QueryExecMode, pgx.QueryResultFormats, pgx.QueryResultFormatsByOID:
			return C12PITRDependencyFailure
		}
	}
	return nil
}

func (access *c12AuthorityAccess) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if access == nil || access.lease == nil {
		return pgconn.CommandTag{}, C12PITRInvalidHandle
	}
	if err := c12RejectPGXControlArguments(args); err != nil {
		return pgconn.CommandTag{}, err
	}
	if err := access.lease.authorize(false, c12AccessOperationExec, sql); err != nil {
		return pgconn.CommandTag{}, err
	}
	operationContext, cancel, err := access.lease.operationContext(ctx)
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	defer cancel()
	backend, err := access.lease.backend(operationContext, false)
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	if err := backend.Acquire(operationContext); err != nil {
		return pgconn.CommandTag{}, c12AccessError(err, false)
	}
	defer backend.Release()
	tag, err := backend.Exec(operationContext, sql, args...)
	return tag, c12AccessError(err, false)
}

func (access *c12AuthorityAccess) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if access == nil || access.lease == nil {
		return nil, C12PITRInvalidHandle
	}
	if err := c12RejectPGXControlArguments(args); err != nil {
		return nil, err
	}
	if err := access.lease.authorize(false, c12AccessOperationQuery, sql); err != nil {
		return nil, err
	}
	operationContext, cancel, err := access.lease.operationContext(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	backend, err := access.lease.backend(operationContext, false)
	if err != nil {
		return nil, err
	}
	if err := backend.Acquire(operationContext); err != nil {
		return nil, c12AccessError(err, false)
	}
	defer backend.Release()
	driverRows, err := backend.Query(operationContext, sql, args...)
	if err != nil {
		return nil, c12AccessError(err, false)
	}
	materialized, err := c12MaterializeRows(operationContext, access.lease, driverRows)
	if err != nil {
		return nil, err
	}
	materialized.(*c12MaterializedRows).typeMap = backend.TypeMap()
	return materialized, nil
}

func (access *c12AuthorityAccess) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	rows, err := access.Query(ctx, sql, args...)
	if err != nil {
		return c12ErrorRow{err: err}
	}
	return &c12MaterializedRow{rows: rows.(*c12MaterializedRows)}
}

func (access *c12AuthorityAccess) Begin(ctx context.Context) (C12AuthorityTransaction, error) {
	if access == nil || access.lease == nil {
		return nil, C12PITRInvalidHandle
	}
	if err := access.lease.authorize(false, c12AccessOperationBegin, ""); err != nil {
		return nil, err
	}
	operationContext, cancel, err := access.lease.operationContext(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	access.lease.mu.Lock()
	if access.lease.activeTx != nil || access.lease.txStarting {
		access.lease.mu.Unlock()
		return nil, C12PITRWrongPhase
	}
	access.lease.txStarting = true
	access.lease.mu.Unlock()
	defer func() {
		access.lease.mu.Lock()
		access.lease.txStarting = false
		access.lease.mu.Unlock()
	}()
	backend, err := access.lease.backend(operationContext, true)
	if err != nil {
		return nil, err
	}
	if err := backend.Acquire(operationContext); err != nil {
		return nil, c12AccessError(err, false)
	}
	driver, err := backend.BeginTx(operationContext, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		backend.Release()
		return nil, c12AccessError(err, false)
	}
	transaction := newC12AuthorityTx(driver, backend, access.lease, access.lease.owner.transactionGeneration.Add(1))
	access.lease.mu.Lock()
	if access.lease.activeTx != nil || !access.lease.live.Load() {
		access.lease.mu.Unlock()
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), time.Second)
		_ = driver.Rollback(cleanupContext)
		cleanupCancel()
		transaction.releaseBackend()
		return nil, C12PITRWrongPhase
	}
	access.lease.activeTx = transaction
	access.lease.mu.Unlock()
	return transaction, nil
}

func (transaction *c12AuthorityTx) status() error {
	if transaction == nil || transaction.driver == nil || transaction.lease == nil || transaction.generation == 0 {
		return C12PITRInvalidHandle
	}
	if err := transaction.lease.status(); err != nil {
		return err
	}
	if transaction.terminal.Load() != 0 {
		return C12PITRWrongPhase
	}
	return nil
}

func newC12AuthorityTx(driver c12AccessDriver, backend c12AccessBackend, lease *c12AccessLease, generation uint64) *c12AuthorityTx {
	transaction := &c12AuthorityTx{
		driver: driver, backend: backend, lease: lease, generation: generation,
		operation: make(chan struct{}, 1),
	}
	transaction.operation <- struct{}{}
	transaction.ownsBackend.Store(true)
	return transaction
}

func (transaction *c12AuthorityTx) acquireOperation(ctx context.Context, terminal uint32) error {
	if transaction == nil || transaction.operation == nil || ctx == nil {
		return C12PITRInvalidHandle
	}
	select {
	case <-ctx.Done():
		return C12PITRCanceled
	case <-transaction.operation:
	}
	if err := ctx.Err(); err != nil {
		transaction.operation <- struct{}{}
		return C12PITRCanceled
	}
	if !transaction.terminal.CompareAndSwap(c12TxIdle, terminal) {
		transaction.operation <- struct{}{}
		return C12PITRWrongPhase
	}
	return nil
}

func (transaction *c12AuthorityTx) releaseOperation(active bool) {
	if active {
		if !transaction.terminal.CompareAndSwap(c12TxOperation, c12TxIdle) {
			transaction.terminal.CompareAndSwap(c12TxRevokedOperation, c12TxRevoked)
		}
	}
	transaction.operation <- struct{}{}
}

func (transaction *c12AuthorityTx) markRevoked() {
	for {
		switch state := transaction.terminal.Load(); state {
		case c12TxIdle:
			if transaction.terminal.CompareAndSwap(state, c12TxRevoked) {
				return
			}
		case c12TxOperation:
			if transaction.terminal.CompareAndSwap(state, c12TxRevokedOperation) {
				return
			}
		default:
			return
		}
	}
}

func (transaction *c12AuthorityTx) releaseBackend() {
	if transaction.ownsBackend.CompareAndSwap(true, false) {
		transaction.backend.Release()
	}
}

func (transaction *c12AuthorityTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if err := transaction.status(); err != nil {
		return pgconn.CommandTag{}, err
	}
	if err := c12RejectPGXControlArguments(args); err != nil {
		return pgconn.CommandTag{}, err
	}
	if err := transaction.lease.authorize(true, c12AccessOperationExec, sql); err != nil {
		return pgconn.CommandTag{}, err
	}
	operationContext, cancel, err := transaction.lease.operationContext(ctx)
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	defer cancel()
	if err := transaction.acquireOperation(operationContext, c12TxOperation); err != nil {
		return pgconn.CommandTag{}, err
	}
	defer transaction.releaseOperation(true)
	tag, err := transaction.driver.Exec(operationContext, sql, args...)
	return tag, c12AccessError(err, false)
}

func (transaction *c12AuthorityTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if err := transaction.status(); err != nil {
		return nil, err
	}
	if err := c12RejectPGXControlArguments(args); err != nil {
		return nil, err
	}
	if err := transaction.lease.authorize(true, c12AccessOperationQuery, sql); err != nil {
		return nil, err
	}
	operationContext, cancel, err := transaction.lease.operationContext(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	if err := transaction.acquireOperation(operationContext, c12TxOperation); err != nil {
		return nil, err
	}
	defer transaction.releaseOperation(true)
	driverRows, err := transaction.driver.Query(operationContext, sql, args...)
	if err != nil {
		return nil, c12AccessError(err, false)
	}
	materialized, err := c12MaterializeRows(operationContext, transaction.lease, driverRows)
	if err != nil {
		return nil, err
	}
	rows := materialized.(*c12MaterializedRows)
	rows.typeMap = transaction.backend.TypeMap()
	rows.transaction = transaction
	return rows, nil
}

func (transaction *c12AuthorityTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	rows, err := transaction.Query(ctx, sql, args...)
	if err != nil {
		return c12ErrorRow{err: err}
	}
	return &c12MaterializedRow{rows: rows.(*c12MaterializedRows)}
}

func (transaction *c12AuthorityTx) Commit(ctx context.Context) error {
	if !transaction.terminal.CompareAndSwap(c12TxIdle, c12TxCommitting) {
		return C12PITRWrongPhase
	}
	err := transaction.driver.Commit(ctx)
	transaction.finishAfterDriver(err)
	return transaction.commitResult(err)
}

func (transaction *c12AuthorityTx) Rollback(ctx context.Context) error {
	if ctx == nil {
		return C12PITRInvalidHandle
	}
	if err := transaction.status(); err != nil {
		return err
	}
	operationContext, cancel, err := transaction.lease.operationContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	if err := transaction.acquireOperation(operationContext, c12TxRollingBack); err != nil {
		return err
	}
	defer transaction.releaseOperation(false)
	err = transaction.driver.Rollback(operationContext)
	transaction.finishAfterDriver(err)
	return c12AccessError(err, false)
}

func (transaction *c12AuthorityTx) cleanupRollback(ctx context.Context) {
	if transaction == nil || transaction.operation == nil {
		return
	}
	select {
	case <-ctx.Done():
		return
	case <-transaction.operation:
	}
	if err := ctx.Err(); err != nil {
		transaction.operation <- struct{}{}
		return
	}
	state := transaction.terminal.Load()
	if (state != c12TxRevoked && state != c12TxIdle) || !transaction.terminal.CompareAndSwap(state, c12TxRollingBack) {
		transaction.operation <- struct{}{}
		return
	}
	defer transaction.releaseOperation(false)
	err := transaction.driver.Rollback(ctx)
	transaction.finishAfterDriver(err)
}

func (transaction *c12AuthorityTx) finishAfterDriver(error) {
	transaction.releaseBackend()
	transaction.lease.finishTransaction(transaction)
}

func (transaction *c12AuthorityTx) commitResult(err error) error {
	if err != nil {
		return c12AccessError(err, true)
	}
	if transaction.lease.ctx.Err() != nil {
		return C12PITRCanceled
	}
	return nil
}

type c12ErrorRow struct{ err error }

func (row c12ErrorRow) Scan(...any) error { return row.err }

type c12NoRows struct{}

func (c12NoRows) Error() string { return C12PITRDependencyFailure.Error() }
func (c12NoRows) Is(target error) bool {
	return target == pgx.ErrNoRows || target == C12PITRDependencyFailure
}

func c12AccessError(err error, commit bool) error {
	if err == nil {
		return nil
	}
	for _, finite := range []C12AuthorityPITRError{
		C12PITRInvalidHandle, C12PITRWrongPhase, C12PITRNotObserved, C12PITRCapacityExceeded,
		C12PITRIndeterminate, C12PITRCanceled, C12PITRDependencyFailure,
	} {
		if errors.Is(err, finite) {
			return finite
		}
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return C12PITRCanceled
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return c12NoRows{}
	}
	if commit {
		return C12PITRIndeterminate
	}
	return C12PITRDependencyFailure
}

func c12AccessReasonCode(reason C12AuthorityPITRError) uint32 {
	if reason == C12PITRCanceled {
		return 2
	}
	return 1
}

func c12AccessReason(code uint32) error {
	if code == 2 {
		return C12PITRCanceled
	}
	return C12PITRWrongPhase
}
