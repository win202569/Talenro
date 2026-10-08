package authority

import (
	"context"
	"errors"
	"reflect"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"talenro.local/platform/internal/store"
)

// This private mechanism is not a production serving capability. Its ordinary
// readiness projection must be replaced by Task 15's v7 proof/lease gate before
// any exported factory or reader may use it.
type boundAuthorityReadSource struct {
	acquire func(context.Context) (boundReadConnection, error)
	check   func(context.Context, store.DBTX) (boundReadSnapshot, error)
}

type boundReadConnection interface {
	store.DBTX
	readIdentity() *pgconn.PgConn
	usable() bool
	Release()
}

type boundReadSnapshot struct {
	readiness Readiness
	point     DatabasePoint
}

func (coordinator *Coordinator) newBoundReadSource() (*boundAuthorityReadSource, error) {
	if coordinatorCallError(context.Background(), coordinator) != nil {
		return nil, ErrAuthorityUnavailable
	}
	repository, exact := coordinator.repository.(*PostgresRepository)
	owner, sameType := coordinator.transaction.(*PostgresRepository)
	if !exact || repository == nil || !sameType || owner != repository {
		return nil, ErrAuthorityUnavailable
	}
	pool, exact := repository.database.(*pgxpool.Pool)
	if !exact || pool == nil {
		return nil, ErrAuthorityUnavailable
	}
	// Capture the Coordinator's dependencies, not a separately supplied pool or
	// caller-owned readiness callback. No connection is acquired by this factory.
	bound := *coordinator
	return &boundAuthorityReadSource{
		acquire: func(ctx context.Context) (boundReadConnection, error) {
			conn, err := pool.Acquire(ctx)
			if err != nil {
				return nil, err
			}
			return &pooledBoundReadConnection{Conn: conn}, nil
		},
		check: bound.checkReadyOnReadConnection,
	}, nil
}

func (source *boundAuthorityReadSource) WithConsistentReadyRead(ctx context.Context, query func(context.Context, store.DBTX) error) error {
	if boundReadCanceled(ctx, nil) {
		return ErrCanceled
	}
	if ctx == nil || source == nil || source.acquire == nil || source.check == nil || query == nil {
		return ErrAuthorityUnavailable
	}
	conn, err := source.acquire(ctx)
	if !nilAuthorityRepositoryValue(conn) {
		defer conn.Release()
	}
	if boundReadCanceled(ctx, err) {
		return ErrCanceled
	}
	if err != nil || nilAuthorityRepositoryValue(conn) {
		return ErrAuthorityUnavailable
	}
	identity := conn.readIdentity()
	intact := func() bool { return identity != nil && conn.usable() && conn.readIdentity() == identity }
	if !intact() {
		return ErrAuthorityUnavailable
	}
	before, err := source.check(ctx, conn)
	if boundReadCanceled(ctx, err) {
		return ErrCanceled
	}
	if err != nil || !before.ready() || !intact() {
		return ErrAuthorityUnavailable
	}
	// A provider or fixture may reuse pointed-to database-point storage. Freeze
	// the pre-check snapshot so an in-place mutation cannot erase the evidence.
	before.readiness.ProviderHead = cloneHead(before.readiness.ProviderHead)
	before.readiness.DatabaseHead = cloneDatabaseHead(before.readiness.DatabaseHead)
	queryErr := query(ctx, conn)
	if boundReadCanceled(ctx, queryErr) {
		return ErrCanceled
	}
	if !intact() {
		return ErrAuthorityUnavailable
	}
	// Even a no-row or malformed domain result must be post-checked.
	after, err := source.check(ctx, conn)
	if boundReadCanceled(ctx, err) {
		return ErrCanceled
	}
	if err != nil || !after.ready() || !intact() || !reflect.DeepEqual(before.readiness, after.readiness) ||
		before.point.SystemID != after.point.SystemID || before.point.Timeline != after.point.Timeline {
		return ErrAuthorityUnavailable
	}
	// WAL can advance without an authority change. It must never regress.
	if walPositionLess(after.point.RequiredLSN, before.point.RequiredLSN) {
		return ErrAuthorityUnavailable
	}
	switch {
	case queryErr == nil:
		return nil
	case errors.Is(queryErr, ErrAuthorityUnavailable):
		return ErrAuthorityUnavailable
	case errors.Is(queryErr, ErrNotFound):
		return ErrNotFound
	case errors.Is(queryErr, ErrConflict):
		return ErrConflict
	default:
		return ErrAuthorityUnavailable
	}
}

func (snapshot boundReadSnapshot) ready() bool {
	return snapshot.readiness.Ready && snapshot.readiness.Reason == ReadinessReady &&
		snapshot.readiness.ProviderHead.Validate() == nil && snapshot.point.Validate() == nil
}

func boundReadCanceled(ctx context.Context, err error) bool {
	return (ctx != nil && ctx.Err() != nil) || errors.Is(err, ErrCanceled) ||
		errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

type pooledBoundReadConnection struct{ *pgxpool.Conn }

func (conn *pooledBoundReadConnection) readIdentity() *pgconn.PgConn {
	return conn.Conn.Conn().PgConn()
}
func (conn *pooledBoundReadConnection) usable() bool {
	physical := conn.Conn.Conn()
	return !physical.IsClosed() && !physical.PgConn().IsBusy() && physical.PgConn().TxStatus() == 'I'
}

// Ordinary readiness only. Keep private until the v7 proof/lease gate exists.
// Every SQL read, including pending-fence discovery, uses the acquired DBTX.
// No explicit transaction spans collectReadinessObservations' provider calls.
func (coordinator *Coordinator) checkReadyOnReadConnection(ctx context.Context, db store.DBTX) (boundReadSnapshot, error) {
	repository, err := NewPostgresRepository(db)
	if err != nil {
		return boundReadSnapshot{}, err
	}
	bound := *coordinator
	bound.repository = repository
	observations, reason, err := bound.collectReadinessObservations(ctx)
	if err != nil {
		return boundReadSnapshot{}, err
	}
	if reason != "" {
		return boundReadSnapshot{}, ErrAuthorityUnavailable
	}
	point, err := repository.CaptureDatabasePoint(ctx)
	if err != nil {
		return boundReadSnapshot{}, err
	}
	ready, err := bound.checkReadyWithDBTX(ctx, db, observations)
	return boundReadSnapshot{readiness: ready, point: point}, err
}
