package authority

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"sync/atomic"
	"talenro.local/platform/internal/store"
)

// boundEpochDB is issued only for a synchronous source callback. Its epoch is
// metadata, not SQL permission or v7 authorization. Never copy this object.
type boundEpochDB struct {
	db    store.DBTX
	epoch atomic.Uint64
}

// BoundReadEpoch returns the frozen pre-check epoch only for the exact private
// callback wrapper while active. It does not authenticate an arbitrary source,
// authorize serving, or unwrap caller-provided decorators. Retaining the DBTX
// beyond the callback is prohibited; expiry does not revoke its SQL methods.
func BoundReadEpoch(db store.DBTX) (uint64, error) {
	d, ok := db.(*boundEpochDB)
	if !ok || d == nil || nilAuthorityRepositoryValue(d.db) {
		return 0, ErrAuthorityUnavailable
	}
	n := d.epoch.Load()
	if !validPositiveCoordinate(n) {
		return 0, ErrAuthorityUnavailable
	}
	return n, nil
}
func (d *boundEpochDB) expire() { d.epoch.Store(0) }
func (d *boundEpochDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return d.db.Exec(ctx, sql, args...)
}
func (d *boundEpochDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return d.db.Query(ctx, sql, args...)
}
func (d *boundEpochDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return d.db.QueryRow(ctx, sql, args...)
}
