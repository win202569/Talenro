package serving

import (
	"context"
	"github.com/jackc/pgx/v5"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/store"
)

type desiredScanDB struct {
	store.DBTX
	rows *desiredScanRows
}

func (d *desiredScanDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	r, err := d.DBTX.Query(ctx, sql, args...)
	if !nilCapability(r) {
		d.rows = &desiredScanRows{Rows: r}
	}
	if err != nil {
		return nil, err
	}
	if d.rows == nil {
		return nil, authority.ErrAuthorityUnavailable
	}
	return d.rows, nil
}
func (d *desiredScanDB) cleanup() {
	if d.rows == nil {
		return
	}
	defer func() {
		for _, target := range d.rows.targets {
			if b, ok := target.(*[]byte); ok {
				clear(*b)
			}
			reflect.ValueOf(target).Elem().SetZero()
		}
		clear(d.rows.targets)
		d.rows.targets = nil
	}()
	d.rows.Close()
}

type desiredScanRows struct {
	pgx.Rows
	targets  []any
	closed   bool
	count    int
	limitErr error
}

func (r *desiredScanRows) Next() bool {
	if r.closed || r.limitErr != nil {
		return false
	}
	if !r.Rows.Next() {
		return false
	}
	if r.count == 33 {
		r.limitErr = authority.ErrAuthorityUnavailable
		return false
	}
	r.count++
	return true
}
func (r *desiredScanRows) Scan(dest ...any) error {
	r.targets = append(r.targets, dest...)
	return r.Rows.Scan(dest...)
}
func (r *desiredScanRows) Err() error {
	if r.limitErr != nil {
		return r.limitErr
	}
	return r.Rows.Err()
}
func (r *desiredScanRows) Close() {
	if !r.closed {
		r.closed = true
		r.Rows.Close()
	}
}
