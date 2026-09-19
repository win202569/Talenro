//go:build integration

package testinfra

import (
	"context"
	"database/sql/driver"
	"errors"
	"math/big"
	"reflect"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"talenro.local/platform/internal/store"
)

var c12AccessCases = []string{
	"eager_rows_and_expired_views", "queryrow_is_not_lazy", "raw_values_are_copied",
	"callback_error", "callback_panic", "parent_cancel", "lease_deadline",
	"one_active_transaction", "one_active_callback", "double_commit",
	"unconsumed_views_need_no_commit_io", "foreign_candidate", "no_rows_is_preserved",
}

type c12AccessTestFixture struct {
	mu             sync.Mutex
	events         []string
	opens          int
	closes         int
	commits        int
	rollbacks      int
	cursorCloses   int
	rows           func() pgx.Rows
	queryErr       error
	commitErr      error
	commitCtxErr   error
	commitWait     chan struct{}
	commitStart    chan struct{}
	queryWait      chan struct{}
	queryStart     chan struct{}
	interruptErr   error
	interruptStuck bool
	closeOnce      sync.Once
	startOnce      sync.Once
	protocolLive   atomic.Int32
	cleanupRace    atomic.Bool
	interrupts     atomic.Int32
}

func (f *c12AccessTestFixture) event(event string) {
	f.mu.Lock()
	f.events = append(f.events, event)
	f.mu.Unlock()
}

func (f *c12AccessTestFixture) snapshot() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.events...)
}

func (f *c12AccessTestFixture) counts() (opens, closes, commits, rollbacks, cursorCloses int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.opens, f.closes, f.commits, f.rollbacks, f.cursorCloses
}

type c12AccessTestDriver struct {
	pgx.Tx
	fixture     *c12AccessTestFixture
	ownership   c12BackendOwnership
	typeMapOnce sync.Once
	typeMap     *pgtype.Map
}

type c12AccessTestRewriter struct {
	called   bool
	retained *pgx.Conn
}

type c12AccessUnboundedCodec struct{ decoded atomic.Bool }

func (*c12AccessUnboundedCodec) FormatSupported(int16) bool { return true }
func (*c12AccessUnboundedCodec) PreferredFormat() int16     { return pgx.BinaryFormatCode }
func (*c12AccessUnboundedCodec) PlanEncode(*pgtype.Map, uint32, int16, any) pgtype.EncodePlan {
	return nil
}
func (*c12AccessUnboundedCodec) PlanScan(*pgtype.Map, uint32, int16, any) pgtype.ScanPlan {
	return nil
}
func (*c12AccessUnboundedCodec) DecodeDatabaseSQLValue(*pgtype.Map, uint32, int16, []byte) (driver.Value, error) {
	return nil, nil
}
func (codec *c12AccessUnboundedCodec) DecodeValue(*pgtype.Map, uint32, int16, []byte) (any, error) {
	codec.decoded.Store(true)
	return map[string][]byte{"mutable": make([]byte, 1<<20)}, nil
}

type c12AccessExpandingCodec struct {
	planned atomic.Bool
	scanned atomic.Bool
}

func (*c12AccessExpandingCodec) FormatSupported(int16) bool { return true }
func (*c12AccessExpandingCodec) PreferredFormat() int16     { return pgx.BinaryFormatCode }
func (*c12AccessExpandingCodec) PlanEncode(*pgtype.Map, uint32, int16, any) pgtype.EncodePlan {
	return nil
}
func (codec *c12AccessExpandingCodec) PlanScan(*pgtype.Map, uint32, int16, any) pgtype.ScanPlan {
	codec.planned.Store(true)
	return c12AccessExpandingScanPlan{codec: codec}
}
func (*c12AccessExpandingCodec) DecodeDatabaseSQLValue(*pgtype.Map, uint32, int16, []byte) (driver.Value, error) {
	return nil, nil
}
func (*c12AccessExpandingCodec) DecodeValue(*pgtype.Map, uint32, int16, []byte) (any, error) {
	return nil, nil
}

type c12AccessExpandingScanPlan struct{ codec *c12AccessExpandingCodec }

func (plan c12AccessExpandingScanPlan) Scan(_ []byte, target any) error {
	plan.codec.scanned.Store(true)
	*(target.(*[]byte)) = make([]byte, 1<<20)
	return nil
}

func (r *c12AccessTestRewriter) RewriteQuery(_ context.Context, conn *pgx.Conn, sql string, args []any) (string, []any, error) {
	r.called = true
	r.retained = conn
	return sql, args, nil
}

func (d *c12AccessTestDriver) Exec(ctx context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
	d.fixture.event("exec")
	select {
	case <-ctx.Done():
		return pgconn.CommandTag{}, ctx.Err()
	default:
		return pgconn.NewCommandTag("UPDATE 1"), nil
	}
}

func (d *c12AccessTestDriver) Query(ctx context.Context, _ string, _ ...any) (pgx.Rows, error) {
	if d.fixture.queryWait != nil {
		d.fixture.protocolLive.Add(1)
		defer d.fixture.protocolLive.Add(-1)
		d.fixture.event("query:start")
		d.fixture.startOnce.Do(func() { close(d.fixture.queryStart) })
		<-d.fixture.queryWait
		d.fixture.event("query:end")
		return nil, ctx.Err()
	}
	d.fixture.event("query")
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if d.fixture.queryErr != nil {
		return nil, d.fixture.queryErr
	}
	if d.fixture.rows == nil {
		return &c12AccessTestRows{fixture: d.fixture, values: [][][]byte{{{1, 2}}, {{3, 4}}}}, nil
	}
	return d.fixture.rows(), nil
}

func (d *c12AccessTestDriver) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	rows, err := d.Query(ctx, sql, args...)
	return c12AccessTestDriverRow{rows: rows, err: err}
}

func (d *c12AccessTestDriver) BeginTx(ctx context.Context, options pgx.TxOptions) (c12AccessDriver, error) {
	d.fixture.event("begin:" + string(options.IsoLevel))
	if options != (pgx.TxOptions{IsoLevel: pgx.ReadCommitted}) {
		return nil, errors.New("unexpected transaction options")
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return d, nil
	}
}

func (d *c12AccessTestDriver) Commit(ctx context.Context) error {
	d.fixture.protocolLive.Add(1)
	defer d.fixture.protocolLive.Add(-1)
	d.fixture.mu.Lock()
	d.fixture.commits++
	d.fixture.events = append(d.fixture.events, "commit:start")
	d.fixture.commitCtxErr = ctx.Err()
	wait, result := d.fixture.commitWait, d.fixture.commitErr
	d.fixture.mu.Unlock()
	if d.fixture.commitStart != nil {
		d.fixture.startOnce.Do(func() { close(d.fixture.commitStart) })
	}
	if wait != nil {
		select {
		case <-wait:
		case <-ctx.Done():
			d.fixture.event("commit:end")
			return ctx.Err()
		}
	}
	if ctx.Err() != nil {
		d.fixture.event("commit:end")
		return ctx.Err()
	}
	d.fixture.event("commit:end")
	return result
}

func (d *c12AccessTestDriver) Rollback(context.Context) error {
	if d.fixture.protocolLive.Load() != 0 {
		d.fixture.cleanupRace.Store(true)
	}
	d.fixture.mu.Lock()
	d.fixture.rollbacks++
	d.fixture.events = append(d.fixture.events, "rollback")
	d.fixture.mu.Unlock()
	return nil
}

func (d *c12AccessTestDriver) Close(context.Context) error {
	if d.fixture.protocolLive.Load() != 0 {
		d.fixture.cleanupRace.Store(true)
	}
	d.fixture.mu.Lock()
	d.fixture.closes++
	d.fixture.events = append(d.fixture.events, "close")
	d.fixture.mu.Unlock()
	if d.fixture.commitWait != nil {
		d.fixture.closeOnce.Do(func() { close(d.fixture.commitWait) })
	}
	if d.fixture.queryWait != nil {
		d.fixture.closeOnce.Do(func() { close(d.fixture.queryWait) })
	}
	return nil
}

func (d *c12AccessTestDriver) Acquire(ctx context.Context) error { return d.ownership.acquire(ctx) }
func (d *c12AccessTestDriver) Release()                          { d.ownership.release() }
func (d *c12AccessTestDriver) Interrupt(context.Context) error {
	if !d.ownership.active.Load() {
		return nil
	}
	d.fixture.interrupts.Add(1)
	d.fixture.event("interrupt")
	if !d.fixture.interruptStuck {
		if d.fixture.commitWait != nil {
			d.fixture.closeOnce.Do(func() { close(d.fixture.commitWait) })
		}
		if d.fixture.queryWait != nil {
			d.fixture.closeOnce.Do(func() { close(d.fixture.queryWait) })
		}
	}
	return d.fixture.interruptErr
}

func (d *c12AccessTestDriver) TypeMap() *pgtype.Map {
	d.typeMapOnce.Do(func() { d.typeMap = pgtype.NewMap() })
	return d.typeMap
}

type c12AccessTestDriverRow struct {
	rows pgx.Rows
	err  error
}

func (r c12AccessTestDriverRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	defer r.rows.Close()
	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return err
		}
		return pgx.ErrNoRows
	}
	return r.rows.Scan(dest...)
}

type c12AccessTestRows struct {
	fixture  *c12AccessTestFixture
	values   [][][]byte
	fields   []pgconn.FieldDescription
	nextHook func()
	hookOnce sync.Once
	index    int
	errAt    int
	err      error
	closed   bool
	started  bool
}

func (r *c12AccessTestRows) Close() {
	if r.closed {
		return
	}
	r.closed = true
	r.fixture.mu.Lock()
	r.fixture.cursorCloses++
	r.fixture.events = append(r.fixture.events, "cursor:close")
	r.fixture.mu.Unlock()
}

func (r *c12AccessTestRows) Err() error { return r.err }
func (r *c12AccessTestRows) CommandTag() pgconn.CommandTag {
	return pgconn.NewCommandTag("SELECT " + string(rune('0'+len(r.values))))
}
func (r *c12AccessTestRows) FieldDescriptions() []pgconn.FieldDescription {
	if r.fields != nil {
		return r.fields
	}
	return []pgconn.FieldDescription{{Name: "payload", DataTypeOID: pgtype.ByteaOID, Format: pgx.BinaryFormatCode}}
}
func (r *c12AccessTestRows) Next() bool {
	if r.closed {
		return false
	}
	next := 0
	if r.started {
		next = r.index + 1
	}
	if r.errAt > 0 && next == r.errAt {
		r.err = errors.New("raw cursor failure")
		r.Close()
		return false
	}
	if next >= len(r.values) {
		r.Close()
		return false
	}
	r.index = next
	r.started = true
	if r.nextHook != nil {
		r.hookOnce.Do(r.nextHook)
	}
	r.fixture.event("cursor:next")
	return true
}
func (r *c12AccessTestRows) Scan(dest ...any) error {
	if len(dest) != 1 || r.index < 0 || r.index >= len(r.values) {
		return errors.New("raw scan misuse")
	}
	value := append([]byte(nil), r.values[r.index][0]...)
	switch target := dest[0].(type) {
	case *[]byte:
		*target = value
		return nil
	default:
		return errors.New("raw scan target")
	}
}
func (r *c12AccessTestRows) Values() ([]any, error) {
	if r.index < 0 || r.index >= len(r.values) {
		return nil, errors.New("raw values misuse")
	}
	return []any{append([]byte(nil), r.values[r.index][0]...)}, nil
}
func (r *c12AccessTestRows) RawValues() [][]byte {
	if r.index < 0 || r.index >= len(r.values) {
		return nil
	}
	return r.values[r.index]
}
func (*c12AccessTestRows) Conn() *pgx.Conn { return &pgx.Conn{} }

func newC12AccessTestState(fixture *c12AccessTestFixture) *c12AuthorityPITRState {
	state := &c12AuthorityPITRState{runGeneration: 1}
	state.phase.Store(1)
	state.accessPolicy = func(*c12CandidateBinding, bool, c12AccessOperation, string) bool { return true }
	state.accessOpen = func(context.Context, *c12CandidateBinding) (c12AccessBackend, error) {
		fixture.mu.Lock()
		fixture.opens++
		fixture.events = append(fixture.events, "open")
		fixture.mu.Unlock()
		return &c12AccessTestDriver{fixture: fixture}, nil
	}
	return state
}

func TestC12AuthorityPITRScopedAccessPolicy(t *testing.T) {
	t.Run("independent_result_decoders", func(t *testing.T) {
		for _, transaction := range []bool{false, true} {
			name := "access"
			if transaction {
				name = "transaction"
			}
			t.Run(name, func(t *testing.T) {
				fixture := &c12AccessTestFixture{}
				backend := &c12AccessTestDriver{fixture: fixture}
				driverMap := backend.TypeMap()
				if backend.TypeMap() != driverMap {
					t.Fatal("backend must retain one stable driver map")
				}
				wireMap := pgtype.NewMap() // Encoding fixtures must not warm the driver's first scan.
				fields := []pgconn.FieldDescription{
					{DataTypeOID: pgtype.Int8OID, Format: pgx.BinaryFormatCode},
					{DataTypeOID: pgtype.Int8OID, Format: pgx.BinaryFormatCode},
					{DataTypeOID: pgtype.Int8OID, Format: pgx.BinaryFormatCode},
					{DataTypeOID: pgtype.Int8OID, Format: pgx.BinaryFormatCode},
					{DataTypeOID: pgtype.Int8OID, Format: pgx.BinaryFormatCode},
					{DataTypeOID: pgtype.ByteaOID, Format: pgx.BinaryFormatCode},
					{DataTypeOID: pgtype.UUIDOID, Format: pgx.BinaryFormatCode},
					{DataTypeOID: pgtype.ByteaOID, Format: pgx.BinaryFormatCode},
					{DataTypeOID: pgtype.NumericOID, Format: pgx.BinaryFormatCode},
					{DataTypeOID: pgtype.Int8OID, Format: pgx.BinaryFormatCode},
					{DataTypeOID: c12PGTypeLSNOID, Format: pgx.TextFormatCode},
					{DataTypeOID: pgtype.BoolOID, Format: pgx.BinaryFormatCode},
				}
				identifier := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
				queries := 0
				fixture.rows = func() pgx.Rows {
					lsn := []string{"0/16B6C50", "0/16B6D80"}[queries]
					queries++
					values := []any{int64(31), int64(1), int64(1), int64(1), int64(0), []byte{1}, identifier, []byte{2}, pgtype.Numeric{Int: big.NewInt(123), Valid: true}, int64(1), lsn, false}
					raw := make([][]byte, len(fields))
					for i, field := range fields {
						var err error
						raw[i], err = wireMap.Encode(field.DataTypeOID, field.Format, values[i], nil)
						if err != nil {
							t.Fatal(err)
						}
					}
					return &c12AccessTestRows{fixture: fixture, fields: fields, values: [][][]byte{raw}}
				}
				state := newC12AccessTestState(fixture)
				state.accessPolicy = nil // Exercise the real, fixed GetAuthorityFenceHead allowance.
				state.fixture = c12PolicyFixture()
				state.accessOpen = func(context.Context, *c12CandidateBinding) (c12AccessBackend, error) { return backend, nil }
				err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
					var reader store.DBTX = access
					if transaction {
						tx, err := access.Begin(t.Context())
						if err != nil {
							return err
						}
						defer tx.Rollback(t.Context())
						reader = tx
					}
					first, firstOK := reader.QueryRow(t.Context(), c12GetAuthorityFenceHeadSQL).(*c12MaterializedRow)
					second, secondOK := reader.QueryRow(t.Context(), c12GetAuthorityFenceHeadSQL).(*c12MaterializedRow)
					if !firstOK || !secondOK {
						t.Fatal("legal independent head reads were not materialized")
					}
					_, _, _, _, closes := fixture.counts()
					if queries != 2 || closes != 2 {
						t.Fatalf("eager head reads: queries=%d closed=%d", queries, closes)
					}
					if first.rows.typeMap == driverMap || second.rows.typeMap == driverMap || first.rows.typeMap == second.rows.typeMap {
						t.Fatal("independent results retain shared mutable driver decoding state")
					}
					start := make(chan struct{})
					ready := make(chan struct{}, 2)
					done := make(chan error, 2)
					var heads [2]store.GetAuthorityFenceHeadRow
					for i, row := range []*c12MaterializedRow{first, second} {
						go func() {
							ready <- struct{}{}
							<-start
							h := &heads[i]
							done <- row.Scan(&h.AuthorityEpoch, &h.RecordCount, &h.LatestReservedSequence, &h.LatestCommittedSequence, &h.PendingCount, &h.ProviderReservationDigest, &h.LatestCommittedOperationID, &h.ProviderReceiptDigest, &h.DbSystemID, &h.DbTimeline, &h.RequiredLsn, &h.HasSequenceGap)
						}()
					}
					<-ready
					<-ready
					close(start)
					for range heads {
						if err := <-done; err != nil {
							t.Fatal(err)
						}
					}
					for i, want := range []string{"0/16B6C50", "0/16B6D80"} {
						h := heads[i]
						if h.RequiredLsn != want || h.AuthorityEpoch != 31 || !h.LatestCommittedOperationID.Valid || h.LatestCommittedOperationID.UUID != identifier || !h.DbSystemID.Valid || h.DbSystemID.Int.Int64() != 123 || h.DbTimeline.Int64 != 1 || h.HasSequenceGap {
							t.Fatalf("head%d decoded incorrectly: %+v", i, h)
						}
					}
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
			})
		}
	})
	t.Run("decoder_state_budget_boundary", func(t *testing.T) {
		for _, remaining := range []int64{4095, 4096} {
			fixture := &c12AccessTestFixture{}
			state := newC12AccessTestState(fixture)
			ctx, cancel := context.WithCancel(t.Context())
			lease := &c12AccessLease{owner: state, generation: 1, ctx: ctx, cancel: cancel}
			lease.live.Store(true)
			lease.budget.bytes = (64 << 20) - remaining
			// No columns or rows: only the retained decoder can consume the budget.
			driverRows := &c12AccessTestRows{fixture: fixture, fields: []pgconn.FieldDescription{}}
			rows, err := c12MaterializeRows(ctx, lease, driverRows)
			cancel()
			if remaining == 4095 {
				if rows != nil || !errors.Is(err, C12PITRCapacityExceeded) {
					t.Fatalf("decoder allocated without complete reservation: rows=%T err=%v", rows, err)
				}
			} else if err != nil || rows == nil || lease.budget.bytes != 64<<20 || lease.budget.rows != 0 {
				t.Fatalf("exact decoder budget: rows=%T err=%v budget=%+v", rows, err, lease.budget)
			}
			if !driverRows.closed {
				t.Fatal("decoder budget path retained the cursor")
			}
		}
	})
	t.Run("bounded_materialization", func(t *testing.T) {
		b := c12ReadBudget{}
		for i := 0; i < 65536; i++ {
			if err := b.take(0); err != nil {
				t.Fatal(err)
			}
		}
		if !errors.Is(b.take(0), C12PITRCapacityExceeded) {
			t.Fatal("row cap")
		}
		if !errors.Is((&c12ReadBudget{}).take(1048577), C12PITRCapacityExceeded) {
			t.Fatal("single row cap")
		}
		b = c12ReadBudget{bytes: 67108864}
		if !errors.Is(b.take(1), C12PITRCapacityExceeded) {
			t.Fatal("byte cap")
		}
	})
	t.Run("copy_budget_is_reserved_before_decode", func(t *testing.T) {
		for _, test := range []struct {
			name      string
			remaining int64
			invoke    func(*testing.T, *c12MaterializedRows) error
		}{
			{
				name: "scan", remaining: 24 + 24 + 64,
				invoke: func(t *testing.T, rows *c12MaterializedRows) error {
					destination := []byte{99}
					err := rows.Scan(&destination)
					if !slices.Equal(destination, []byte{99}) {
						t.Fatalf("Scan mutated destination before capacity rejection: %v", destination)
					}
					return err
				},
			},
			{
				name: "values", remaining: 24 + 16 + 64,
				invoke: func(_ *testing.T, rows *c12MaterializedRows) error {
					_, err := rows.Values()
					return err
				},
			},
		} {
			t.Run(test.name, func(t *testing.T) {
				state := newC12AccessTestState(&c12AccessTestFixture{})
				leaseContext, cancel := context.WithCancel(t.Context())
				defer cancel()
				lease := &c12AccessLease{owner: state, generation: 1, ctx: leaseContext, cancel: cancel}
				lease.live.Store(true)
				lease.budget.bytes = (64 << 20) - test.remaining
				rows := &c12MaterializedRows{
					lease: lease, typeMap: pgtype.NewMap(), index: 0, current: true,
					fields: []pgconn.FieldDescription{{Name: "payload", DataTypeOID: pgtype.ByteaOID, Format: pgx.BinaryFormatCode}},
					values: [][][]byte{{make([]byte, 64)}},
				}
				if err := test.invoke(t, rows); !errors.Is(err, C12PITRCapacityExceeded) {
					t.Fatalf("copy reservation = %v", err)
				}
			})
		}
	})
	t.Run("unsupported_values_codec_is_not_decoded", func(t *testing.T) {
		const unsafeOID = 91001
		codec := &c12AccessUnboundedCodec{}
		typeMap := pgtype.NewMap()
		typeMap.RegisterType(&pgtype.Type{Name: "c12_unsafe", OID: unsafeOID, Codec: codec})
		state := newC12AccessTestState(&c12AccessTestFixture{})
		leaseContext, cancel := context.WithCancel(t.Context())
		defer cancel()
		lease := &c12AccessLease{owner: state, generation: 1, ctx: leaseContext, cancel: cancel}
		lease.live.Store(true)
		rows := &c12MaterializedRows{
			lease: lease, typeMap: typeMap, index: 0, current: true,
			fields: []pgconn.FieldDescription{{Name: "unsafe", DataTypeOID: unsafeOID, Format: pgx.BinaryFormatCode}},
			values: [][][]byte{{{1}}},
		}
		if _, err := rows.Values(); !errors.Is(err, C12PITRDependencyFailure) {
			t.Fatalf("unsupported codec = %v", err)
		}
		if codec.decoded.Load() {
			t.Fatal("unsupported mutable codec decoded before rejection")
		}
	})
	t.Run("decoded_copy_budget_uses_codec_and_target_bounds", func(t *testing.T) {
		t.Run("date_values_fixed_size", func(t *testing.T) {
			state := newC12AccessTestState(&c12AccessTestFixture{})
			leaseContext, cancel := context.WithCancel(t.Context())
			defer cancel()
			lease := &c12AccessLease{owner: state, generation: 1, ctx: leaseContext, cancel: cancel}
			lease.live.Store(true)
			lease.budget.bytes = (64 << 20) - 63
			rows := &c12MaterializedRows{
				lease: lease, typeMap: pgtype.NewMap(), index: 0, current: true,
				fields: []pgconn.FieldDescription{{Name: "date", DataTypeOID: pgtype.DateOID, Format: pgx.BinaryFormatCode}},
				values: [][][]byte{{{0, 0, 0, 0}}},
			}
			if _, err := rows.Values(); !errors.Is(err, C12PITRCapacityExceeded) {
				t.Fatalf("Date decoded-size reservation = %v", err)
			}
		})
		t.Run("unsupported_expanding_scan", func(t *testing.T) {
			const expandingOID = 91002
			codec := &c12AccessExpandingCodec{}
			typeMap := pgtype.NewMap()
			typeMap.RegisterType(&pgtype.Type{Name: "c12_expanding", OID: expandingOID, Codec: codec})
			state := newC12AccessTestState(&c12AccessTestFixture{})
			leaseContext, cancel := context.WithCancel(t.Context())
			defer cancel()
			lease := &c12AccessLease{owner: state, generation: 1, ctx: leaseContext, cancel: cancel}
			lease.live.Store(true)
			rows := &c12MaterializedRows{
				lease: lease, typeMap: typeMap, index: 0, current: true,
				fields: []pgconn.FieldDescription{{Name: "expanding", DataTypeOID: expandingOID, Format: pgx.BinaryFormatCode}},
				values: [][][]byte{{{1}}},
			}
			destination := []byte{99}
			if err := rows.Scan(&destination); !errors.Is(err, C12PITRDependencyFailure) {
				t.Fatalf("expanding Scan = %v", err)
			}
			if !slices.Equal(destination, []byte{99}) || codec.planned.Load() || codec.scanned.Load() {
				t.Fatalf("expanding codec invoked: destination=%d planned=%t scanned=%t", len(destination), codec.planned.Load(), codec.scanned.Load())
			}
		})
		for _, test := range []struct {
			name   string
			raw    []byte
			target func() any
		}{
			{
				name: "numeric_weight_expansion_to_string",
				raw:  []byte{0, 1, 0x03, 0xe8, 0, 0, 0, 0, 0, 1},
				target: func() any {
					value := "unchanged"
					return &value
				},
			},
			{
				name: "numeric_dscale_expansion_to_numeric",
				raw:  []byte{0, 1, 0, 0, 0, 0, 0x10, 0, 0, 1},
				target: func() any {
					return &pgtype.Numeric{}
				},
			},
		} {
			t.Run(test.name, func(t *testing.T) {
				state := newC12AccessTestState(&c12AccessTestFixture{})
				leaseContext, cancel := context.WithCancel(t.Context())
				defer cancel()
				lease := &c12AccessLease{owner: state, generation: 1, ctx: leaseContext, cancel: cancel}
				lease.live.Store(true)
				rows := &c12MaterializedRows{
					lease: lease, typeMap: pgtype.NewMap(), index: 0, current: true,
					fields: []pgconn.FieldDescription{{Name: "system_id", DataTypeOID: pgtype.NumericOID, Format: pgx.BinaryFormatCode}},
					values: [][][]byte{{test.raw}},
				}
				if err := rows.Scan(test.target()); !errors.Is(err, C12PITRDependencyFailure) {
					t.Fatalf("unbounded numeric Scan = %v", err)
				}
			})
		}
	})
	t.Run("fixed_query_scan_shapes_remain_supported", func(t *testing.T) {
		typeMap := pgtype.NewMap()
		encode := func(oid uint32, value any) []byte {
			raw, err := typeMap.Encode(oid, pgx.BinaryFormatCode, value, nil)
			if err != nil {
				t.Fatal(err)
			}
			return raw
		}
		identifier := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
		numericInt, ok := new(big.Int).SetString("18446744073709551615", 10)
		if !ok {
			t.Fatal("numeric fixture")
		}
		numericRaw := encode(pgtype.NumericOID, pgtype.Numeric{Int: numericInt, Valid: true})
		timestamp := time.Date(2026, time.September, 19, 12, 0, 0, 0, time.UTC)
		var (
			plainUUID    uuid.UUID
			nullUUID     uuid.NullUUID
			numeric      pgtype.Numeric
			numericNull  pgtype.Numeric
			pgInt        pgtype.Int8
			plainInt     int64
			fixtureCount int
			pgTimestamp  pgtype.Timestamptz
			pgText       pgtype.Text
			plainText    string
			bytesValue   []byte
			booleanValue bool
			unknownLSN   any
		)
		cases := []struct {
			name   string
			field  pgconn.FieldDescription
			raw    []byte
			target any
		}{
			{"uuid", pgconn.FieldDescription{DataTypeOID: pgtype.UUIDOID, Format: pgx.BinaryFormatCode}, identifier[:], &plainUUID},
			{"nullable_uuid", pgconn.FieldDescription{DataTypeOID: pgtype.UUIDOID, Format: pgx.BinaryFormatCode}, identifier[:], &nullUUID},
			{"numeric", pgconn.FieldDescription{DataTypeOID: pgtype.NumericOID, Format: pgx.BinaryFormatCode}, numericRaw, &numeric},
			{"numeric_null", pgconn.FieldDescription{DataTypeOID: pgtype.NumericOID, Format: pgx.BinaryFormatCode}, nil, &numericNull},
			{"pg_int8", pgconn.FieldDescription{DataTypeOID: pgtype.Int8OID, Format: pgx.BinaryFormatCode}, encode(pgtype.Int8OID, int64(7)), &pgInt},
			{"int64", pgconn.FieldDescription{DataTypeOID: pgtype.Int8OID, Format: pgx.BinaryFormatCode}, encode(pgtype.Int8OID, int64(7)), &plainInt},
			{"int", pgconn.FieldDescription{DataTypeOID: pgtype.Int8OID, Format: pgx.BinaryFormatCode}, encode(pgtype.Int8OID, int64(7)), &fixtureCount},
			{"timestamptz", pgconn.FieldDescription{DataTypeOID: pgtype.TimestamptzOID, Format: pgx.BinaryFormatCode}, encode(pgtype.TimestamptzOID, timestamp), &pgTimestamp},
			{"pg_text", pgconn.FieldDescription{DataTypeOID: pgtype.TextOID, Format: pgx.TextFormatCode}, []byte("text"), &pgText},
			{"string", pgconn.FieldDescription{DataTypeOID: pgtype.TextOID, Format: pgx.TextFormatCode}, []byte("text"), &plainText},
			{"bytea", pgconn.FieldDescription{DataTypeOID: pgtype.ByteaOID, Format: pgx.BinaryFormatCode}, []byte{1, 2, 3}, &bytesValue},
			{"bool", pgconn.FieldDescription{DataTypeOID: pgtype.BoolOID, Format: pgx.BinaryFormatCode}, []byte{1}, &booleanValue},
			{"unknown_lsn", pgconn.FieldDescription{DataTypeOID: c12PGTypeLSNOID, Format: pgx.TextFormatCode}, []byte("0/16B6C50"), &unknownLSN},
		}
		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				state := newC12AccessTestState(&c12AccessTestFixture{})
				leaseContext, cancel := context.WithCancel(t.Context())
				defer cancel()
				lease := &c12AccessLease{owner: state, generation: 1, ctx: leaseContext, cancel: cancel}
				lease.live.Store(true)
				rows := &c12MaterializedRows{
					lease: lease, typeMap: typeMap, index: 0, current: true,
					fields: []pgconn.FieldDescription{test.field}, values: [][][]byte{{test.raw}},
				}
				if err := rows.Scan(test.target); err != nil {
					t.Fatalf("fixed Scan shape = %v", err)
				}
			})
		}
	})
	t.Run("empty_bytea_remains_distinct_from_null", func(t *testing.T) {
		for _, view := range []string{"raw", "scan", "values"} {
			t.Run(view, func(t *testing.T) {
				fixture := &c12AccessTestFixture{}
				fixture.rows = func() pgx.Rows {
					return &c12AccessTestRows{fixture: fixture, values: [][][]byte{{make([]byte, 0)}, {nil}}}
				}
				state := newC12AccessTestState(fixture)
				err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
					rows, err := access.Query(t.Context(), "empty-and-null")
					if err != nil {
						return err
					}
					for index := 0; index < 2; index++ {
						if !rows.Next() {
							t.Fatalf("row %d missing: %v", index, rows.Err())
						}
						var value []byte
						switch view {
						case "raw":
							value = rows.RawValues()[0]
						case "scan":
							value = []byte{99}
							if err := rows.Scan(&value); err != nil {
								return err
							}
						case "values":
							decoded, err := rows.Values()
							if err != nil {
								return err
							}
							if decoded[0] != nil {
								value = decoded[0].([]byte)
							}
						}
						if index == 0 && (value == nil || len(value) != 0) {
							t.Fatalf("empty bytea became NULL: %#v", value)
						}
						if index == 1 && value != nil {
							t.Fatalf("NULL became nonnil: %#v", value)
						}
					}
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
			})
		}
	})
	t.Run("queryrow_success_is_never_replayed", func(t *testing.T) {
		t.Run("consumed", func(t *testing.T) {
			fixture := &c12AccessTestFixture{}
			state := newC12AccessTestState(fixture)
			err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
				row := access.QueryRow(t.Context(), "one")
				var first []byte
				if err := row.Scan(&first); err != nil {
					return err
				}
				second := []byte{99}
				if err := row.Scan(&second); !errors.Is(err, C12PITRWrongPhase) {
					t.Fatalf("second Scan = %v", err)
				}
				if !slices.Equal(second, []byte{99}) {
					t.Fatalf("second Scan mutated destination: %v", second)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Run("revoked", func(t *testing.T) {
			fixture := &c12AccessTestFixture{}
			state := newC12AccessTestState(fixture)
			var saved pgx.Row
			err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
				saved = access.QueryRow(t.Context(), "one")
				var first []byte
				return saved.Scan(&first)
			})
			if err != nil {
				t.Fatal(err)
			}
			destination := []byte{99}
			if err := saved.Scan(&destination); !errors.Is(err, C12PITRWrongPhase) {
				t.Fatalf("revoked Scan = %v", err)
			}
			if !slices.Equal(destination, []byte{99}) {
				t.Fatalf("revoked Scan mutated destination: %v", destination)
			}
		})
	})
	t.Run("zero_controller", func(t *testing.T) {
		called := false
		err := (C12AuthorityPITRController{}).WithPrimaryAuthorityAccess(
			t.Context(), func(C12AuthorityAccess) error { called = true; return nil })
		if !errors.Is(err, C12PITRInvalidHandle) || called {
			t.Fatal("zero authorized")
		}
	})
	t.Run("no_raw_method_set", func(t *testing.T) {
		for _, typ := range []reflect.Type{
			reflect.TypeFor[*c12AuthorityAccess](), reflect.TypeFor[*c12AuthorityTx](),
		} {
			for _, name := range []string{"Conn", "Config", "BeginTx", "Prepare", "SendBatch", "CopyFrom", "LargeObjects"} {
				if _, ok := typ.MethodByName(name); ok {
					t.Fatalf("escape: %s", name)
				}
			}
		}
		if reflect.TypeFor[*c12AuthorityTx]().Implements(reflect.TypeFor[pgx.Tx]()) {
			t.Fatal("transaction exposes pgx.Tx")
		}
	})

	t.Run("rejects_pgx_control_arguments", func(t *testing.T) {
		fixture := &c12AccessTestFixture{}
		state := newC12AccessTestState(fixture)
		err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			tx, err := access.Begin(t.Context())
			if err != nil {
				return err
			}
			rewriter := &c12AccessTestRewriter{}
			controls := []any{
				rewriter,
				pgx.QueryExecModeSimpleProtocol,
				pgx.QueryResultFormats{pgx.BinaryFormatCode},
				pgx.QueryResultFormatsByOID{pgtype.ByteaOID: pgx.BinaryFormatCode},
			}
			for _, db := range []interface {
				Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
				Query(context.Context, string, ...any) (pgx.Rows, error)
				QueryRow(context.Context, string, ...any) pgx.Row
			}{access, tx} {
				for _, control := range controls {
					before := len(fixture.snapshot())
					if _, err := db.Exec(t.Context(), "authorized", control); !errors.Is(err, C12PITRDependencyFailure) {
						t.Fatalf("Exec accepted %T: %v", control, err)
					}
					if _, err := db.Query(t.Context(), "authorized", control); !errors.Is(err, C12PITRDependencyFailure) {
						t.Fatalf("Query accepted %T: %v", control, err)
					}
					var payload []byte
					if err := db.QueryRow(t.Context(), "authorized", control).Scan(&payload); !errors.Is(err, C12PITRDependencyFailure) {
						t.Fatalf("QueryRow accepted %T: %v", control, err)
					}
					if got := len(fixture.snapshot()); got != before {
						t.Fatalf("%T reached backend: events %d -> %d", control, before, got)
					}
				}
			}
			if rewriter.called || rewriter.retained != nil {
				t.Fatal("QueryRewriter received raw connection")
			}
			return tx.Rollback(t.Context())
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("transport_interrupt_serializes_query_cleanup", func(t *testing.T) {
		fixture := &c12AccessTestFixture{
			queryWait: make(chan struct{}), queryStart: make(chan struct{}),
			interruptErr: errors.New("transport close status"),
		}
		state := newC12AccessTestState(fixture)
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		go func() {
			done <- c12WithAccess(ctx, state, nil, func(access C12AuthorityAccess) error {
				_, err := access.Query(context.Background(), "authorized")
				return err
			})
		}()
		<-fixture.queryStart
		cancel()
		if err := <-done; !errors.Is(err, C12PITRCanceled) {
			t.Fatalf("cancel result = %v", err)
		}
		events := fixture.snapshot()
		start := slices.Index(events, "query:start")
		interrupt := slices.Index(events, "interrupt")
		end := slices.Index(events, "query:end")
		closeIndex := slices.Index(events, "close")
		if start < 0 || interrupt < start || end < interrupt || closeIndex < end {
			t.Fatalf("query cancellation order = %v", events)
		}
		if fixture.interrupts.Load() != 1 || fixture.cleanupRace.Load() {
			t.Fatalf("transport interruption = interrupts %d cleanup overlap %t", fixture.interrupts.Load(), fixture.cleanupRace.Load())
		}
	})

	t.Run("unresponsive_transport_never_forces_protocol_cleanup_overlap", func(t *testing.T) {
		fixture := &c12AccessTestFixture{
			commitWait: make(chan struct{}), commitStart: make(chan struct{}),
			interruptErr: errors.New("transport did not close"), interruptStuck: true,
		}
		state := newC12AccessTestState(fixture)
		leaseContext, cancel := context.WithCancel(t.Context())
		lease := &c12AccessLease{owner: state, generation: 1, ctx: leaseContext, cancel: cancel}
		lease.live.Store(true)
		backend := &c12AccessTestDriver{fixture: fixture}
		if err := backend.Acquire(t.Context()); err != nil {
			t.Fatal(err)
		}
		transaction := newC12AuthorityTx(backend, backend, lease, 1)
		lease.txBackend = backend
		lease.activeTx = transaction
		commitDone := make(chan error, 1)
		go func() { commitDone <- transaction.Commit(context.Background()) }()
		<-fixture.commitStart
		started := time.Now()
		revokeDone := make(chan struct{})
		go func() {
			lease.revoke(C12PITRCanceled)
			close(revokeDone)
		}()
		select {
		case <-revokeDone:
		case <-time.After(2 * time.Second):
			t.Fatal("bounded cleanup did not return")
		}
		if elapsed := time.Since(started); elapsed < 500*time.Millisecond {
			t.Fatalf("cleanup did not wait for ownership: %v", elapsed)
		}
		_, closes, _, rollbacks, _ := fixture.counts()
		if fixture.interrupts.Load() != 1 || fixture.cleanupRace.Load() || closes != 0 || rollbacks != 0 {
			t.Fatalf("failed interrupt cleanup = interrupts %d overlap %t closes %d rollbacks %d", fixture.interrupts.Load(), fixture.cleanupRace.Load(), closes, rollbacks)
		}
		close(fixture.commitWait)
		if err := <-commitDone; !errors.Is(err, C12PITRCanceled) {
			t.Fatalf("released commit = %v", err)
		}
	})

	t.Run("query_self_revocation_survives_late_operation_unwind", func(t *testing.T) {
		fixture := &c12AccessTestFixture{
			interruptErr: errors.New("transport did not close"), interruptStuck: true,
		}
		state := newC12AccessTestState(fixture)
		leaseContext, cancel := context.WithCancel(t.Context())
		lease := &c12AccessLease{owner: state, generation: 1, ctx: leaseContext, cancel: cancel}
		lease.live.Store(true)
		fixture.rows = func() pgx.Rows {
			return &c12AccessTestRows{
				fixture: fixture, values: [][][]byte{{{1, 2}}}, nextHook: cancel,
			}
		}
		backend := &c12AccessTestDriver{fixture: fixture}
		if err := backend.Acquire(t.Context()); err != nil {
			t.Fatal(err)
		}
		transaction := newC12AuthorityTx(backend, backend, lease, 1)
		defer transaction.releaseBackend()
		lease.txBackend = backend
		lease.activeTx = transaction
		started := time.Now()
		if _, err := transaction.Query(context.Background(), "self-revoke"); !errors.Is(err, C12PITRCanceled) {
			t.Fatalf("self-revoking Query = %v", err)
		}
		if elapsed := time.Since(started); elapsed < 500*time.Millisecond || elapsed > 2*time.Second {
			t.Fatalf("cleanup timeout path elapsed %v", elapsed)
		}
		commitErr := transaction.Commit(context.Background())
		_, closes, commits, rollbacks, _ := fixture.counts()
		if !errors.Is(commitErr, C12PITRWrongPhase) || fixture.interrupts.Load() != 1 || commits != 0 || rollbacks != 0 || closes != 0 {
			t.Fatalf("late unwind state = commit %v interrupts %d commits %d rollbacks %d closes %d", commitErr, fixture.interrupts.Load(), commits, rollbacks, closes)
		}
	})

	t.Run("canceled_rollback_does_not_consume_transaction", func(t *testing.T) {
		fixture := &c12AccessTestFixture{}
		state := newC12AccessTestState(fixture)
		err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			tx, err := access.Begin(t.Context())
			if err != nil {
				return err
			}
			canceled, cancel := context.WithCancel(t.Context())
			cancel()
			if err := tx.Rollback(canceled); !errors.Is(err, C12PITRCanceled) {
				t.Fatalf("canceled rollback = %v", err)
			}
			if _, err := access.Begin(t.Context()); !errors.Is(err, C12PITRWrongPhase) {
				t.Fatalf("canceled rollback released active transaction: %v", err)
			}
			if err := tx.Rollback(t.Context()); err != nil {
				t.Fatalf("rollback retry = %v", err)
			}
			next, err := access.Begin(t.Context())
			if err != nil {
				return err
			}
			return next.Rollback(t.Context())
		})
		_, _, _, rollbacks, _ := fixture.counts()
		if err != nil || rollbacks != 2 {
			t.Fatalf("transaction cleanup = %v, rollbacks %d", err, rollbacks)
		}
	})

	t.Run("commit_after_admission_has_only_transition_and_driver", func(t *testing.T) {
		t.Run("lease_canceled_after_admission", func(t *testing.T) {
			fixture := &c12AccessTestFixture{}
			state := newC12AccessTestState(fixture)
			leaseContext, cancel := context.WithCancel(t.Context())
			cancel()
			lease := &c12AccessLease{owner: state, generation: 1, ctx: leaseContext, cancel: func() {}}
			lease.live.Store(true)
			backend := &c12AccessTestDriver{fixture: fixture}
			transaction := &c12AuthorityTx{driver: backend, backend: backend, lease: lease, generation: 1}
			lease.txBackend = backend
			lease.activeTx = transaction
			if err := transaction.Commit(context.Background()); !errors.Is(err, C12PITRCanceled) {
				t.Fatalf("commit result = %v", err)
			}
			_, closes, commits, rollbacks, _ := fixture.counts()
			if terminal := transaction.terminal.Load(); terminal != 1 || commits != 1 || rollbacks != 0 || closes != 0 {
				t.Fatalf("post-admission path = terminal %d commits %d rollbacks %d closes %d", terminal, commits, rollbacks, closes)
			}
		})
		t.Run("prepared_backend_ownership", func(t *testing.T) {
			fixture := &c12AccessTestFixture{commitStart: make(chan struct{})}
			state := newC12AccessTestState(fixture)
			leaseContext, cancel := context.WithCancel(t.Context())
			defer cancel()
			lease := &c12AccessLease{owner: state, generation: 1, ctx: leaseContext, cancel: cancel}
			lease.live.Store(true)
			backend := &c12AccessTestDriver{fixture: fixture}
			if err := backend.Acquire(t.Context()); err != nil {
				t.Fatal(err)
			}
			transaction := newC12AuthorityTx(backend, backend, lease, 1)
			done := make(chan error, 1)
			go func() { done <- transaction.Commit(context.Background()) }()
			select {
			case <-fixture.commitStart:
				if err := <-done; err != nil {
					t.Fatalf("commit result = %v", err)
				}
			case <-time.After(250 * time.Millisecond):
				transaction.ownsBackend.Store(false)
				backend.Release()
				<-done
				t.Fatal("Commit waited for backend ownership after admission")
			}
		})
	})

	for _, name := range c12AccessCases {
		t.Run(name, func(t *testing.T) {
			testC12AccessLifecycleCase(t, name)
		})
	}

	t.Run("deny_all_default", func(t *testing.T) {
		fixture := &c12AccessTestFixture{}
		state := newC12AccessTestState(fixture)
		state.accessPolicy = nil
		called := false
		err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			called = true
			if _, err := access.Exec(t.Context(), "SELECT forbidden"); !errors.Is(err, C12PITRDependencyFailure) {
				t.Fatalf("Exec error = %v", err)
			}
			if _, err := access.Query(t.Context(), "SELECT forbidden"); !errors.Is(err, C12PITRDependencyFailure) {
				t.Fatalf("Query error = %v", err)
			}
			if _, err := access.Begin(t.Context()); !errors.Is(err, C12PITRDependencyFailure) {
				t.Fatalf("Begin error = %v", err)
			}
			return nil
		})
		if err != nil || !called {
			t.Fatalf("deny callback = (%t,%v)", called, err)
		}
		if opens, _, _, _, _ := fixture.counts(); opens != 0 {
			t.Fatalf("deny opened %d connections", opens)
		}
	})

	t.Run("transition_exclusion", func(t *testing.T) {
		fixture := &c12AccessTestFixture{}
		state := newC12AccessTestState(fixture)
		entered, release := make(chan struct{}), make(chan struct{})
		done := make(chan error, 1)
		go func() {
			done <- c12WithAccess(t.Context(), state, nil, func(C12AuthorityAccess) error {
				close(entered)
				<-release
				return nil
			})
		}()
		select {
		case <-entered:
		case err := <-done:
			t.Fatalf("access did not enter: %v", err)
		}
		if state.beginC12Transition(func() bool { return state.phase.CompareAndSwap(1, 10) }) {
			t.Fatal("transition overlapped access")
		}
		close(release)
		if err := <-done; err != nil {
			t.Fatal(err)
		}
		if !state.beginC12Transition(func() bool { return state.phase.CompareAndSwap(1, 10) }) {
			t.Fatal("transition not admitted after access")
		}
		if err := c12WithAccess(t.Context(), state, nil, func(C12AuthorityAccess) error { return nil }); !errors.Is(err, C12PITRWrongPhase) {
			t.Fatalf("access overlapped transition: %v", err)
		}
		state.endC12Transition()
	})

	t.Run("materialization_failures_are_eager_and_finite", func(t *testing.T) {
		fixture := &c12AccessTestFixture{}
		fixture.rows = func() pgx.Rows {
			return &c12AccessTestRows{fixture: fixture, values: [][][]byte{{{1}}, {{2}}}, errAt: 2}
		}
		state := newC12AccessTestState(fixture)
		called := false
		err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			called = true
			rows, err := access.Query(t.Context(), "broken rows")
			if rows != nil || !errors.Is(err, C12PITRDependencyFailure) || strings.Contains(err.Error(), "cursor") {
				t.Fatalf("partial rows published: rows %v err %v", rows, err)
			}
			return nil
		})
		_, _, _, _, cursorCloses := fixture.counts()
		if err != nil || !called || cursorCloses != 1 {
			t.Fatalf("eager failure = called %t err %v closes %d", called, err, cursorCloses)
		}

		fixture = &c12AccessTestFixture{queryErr: errors.New("raw query detail")}
		state = newC12AccessTestState(fixture)
		err = c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			_, err := access.Query(t.Context(), "query error")
			if !errors.Is(err, C12PITRDependencyFailure) || strings.Contains(err.Error(), "raw") {
				t.Fatalf("query error leaked: %v", err)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("materialization_metadata_and_closed_views_are_bounded", func(t *testing.T) {
		fixture := &c12AccessTestFixture{}
		fixture.rows = func() pgx.Rows { return &c12AccessTestRows{fixture: fixture, values: [][][]byte{}} }
		state := newC12AccessTestState(fixture)
		err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			rows, err := access.Query(t.Context(), "empty")
			if err != nil {
				return err
			}
			lease := access.(*c12AuthorityAccess).lease
			lease.budgetMu.Lock()
			charged := lease.budget.bytes
			lease.budgetMu.Unlock()
			if charged == 0 {
				t.Fatal("zero-row metadata was free")
			}
			if rows.Conn() != nil {
				t.Fatal("raw connection escaped")
			}
			rows.Close()
			if rows.Next() || rows.RawValues() != nil || rows.FieldDescriptions() != nil {
				t.Fatal("closed rows exposed data")
			}
			if _, err := rows.Values(); !errors.Is(err, C12PITRWrongPhase) {
				t.Fatalf("closed Values = %v", err)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("transaction_rows_are_eager", func(t *testing.T) {
		fixture := &c12AccessTestFixture{}
		state := newC12AccessTestState(fixture)
		err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			tx, err := access.Begin(t.Context())
			if err != nil {
				return err
			}
			rows, err := tx.Query(t.Context(), "tx rows")
			if err != nil {
				return err
			}
			_, _, _, _, cursorCloses := fixture.counts()
			if cursorCloses != 1 || !rows.Next() {
				t.Fatalf("transaction rows remained lazy: closes %d", cursorCloses)
			}
			var payload []byte
			if err := rows.Scan(&payload); err != nil || !slices.Equal(payload, []byte{1, 2}) {
				t.Fatalf("transaction Scan = %v, %v", payload, err)
			}
			return tx.Rollback(t.Context())
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("commit_context_and_indeterminate_result", func(t *testing.T) {
		fixture := &c12AccessTestFixture{}
		state := newC12AccessTestState(fixture)
		err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			tx, err := access.Begin(t.Context())
			if err != nil {
				return err
			}
			canceled, cancel := context.WithCancel(context.Background())
			cancel()
			if err := tx.Commit(canceled); !errors.Is(err, C12PITRCanceled) {
				t.Fatalf("caller cancellation = %v", err)
			}
			return nil
		})
		fixture.mu.Lock()
		commitCtxErr := fixture.commitCtxErr
		fixture.mu.Unlock()
		if err != nil || !errors.Is(commitCtxErr, context.Canceled) {
			t.Fatalf("driver commit context = %v, callback %v", commitCtxErr, err)
		}

		fixture = &c12AccessTestFixture{commitErr: errors.New("raw commit detail")}
		state = newC12AccessTestState(fixture)
		err = c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			tx, err := access.Begin(t.Context())
			if err != nil {
				return err
			}
			err = tx.Commit(t.Context())
			if !errors.Is(err, C12PITRIndeterminate) || strings.Contains(err.Error(), "raw") {
				t.Fatalf("commit error = %v", err)
			}
			return nil
		})
		if _, _, commits, rollbacks, _ := fixture.counts(); err != nil || commits != 1 || rollbacks != 0 {
			t.Fatalf("indeterminate commit = callback %v commits %d rollback %d", err, commits, rollbacks)
		}
	})
}

func testC12AccessLifecycleCase(t *testing.T, name string) {
	t.Helper()
	fixture := &c12AccessTestFixture{}
	state := newC12AccessTestState(fixture)

	switch name {
	case "eager_rows_and_expired_views":
		var savedRows pgx.Rows
		var savedTx C12AuthorityTransaction
		var savedAccess C12AuthorityAccess
		called := false
		err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			called = true
			savedAccess = access
			rows, err := access.Query(context.Background(), "rows")
			if err != nil {
				return err
			}
			savedRows = rows
			_, _, _, _, cursorCloses := fixture.counts()
			if cursorCloses != 1 {
				t.Fatalf("cursor closes at Query return = %d", cursorCloses)
			}
			tx, err := access.Begin(t.Context())
			if err != nil {
				return err
			}
			savedTx = tx
			return nil
		})
		if err != nil || !called {
			t.Fatalf("callback = (%t,%v)", called, err)
		}
		if opens, _, _, _, _ := fixture.counts(); opens != 2 {
			t.Fatalf("access plus transaction connections = %d", opens)
		}
		before := len(fixture.snapshot())
		if savedRows.Next() {
			t.Fatal("expired rows advanced")
		}
		if _, err := savedRows.Values(); !errors.Is(err, C12PITRWrongPhase) {
			t.Fatalf("expired Values = %v", err)
		}
		if _, err := savedTx.Exec(t.Context(), "late"); !errors.Is(err, C12PITRWrongPhase) {
			t.Fatalf("expired tx = %v", err)
		}
		if _, err := savedAccess.Exec(t.Context(), "late"); !errors.Is(err, C12PITRWrongPhase) {
			t.Fatalf("expired access = %v", err)
		}
		if got := len(fixture.snapshot()); got != before {
			t.Fatalf("expired views did I/O: %d -> %d", before, got)
		}

	case "queryrow_is_not_lazy":
		called := false
		err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			called = true
			row := access.QueryRow(t.Context(), "row")
			_, _, _, _, cursorCloses := fixture.counts()
			if cursorCloses != 1 {
				t.Fatalf("cursor closes before Scan = %d", cursorCloses)
			}
			var payload []byte
			if err := row.Scan(&payload); err != nil || !slices.Equal(payload, []byte{1, 2}) {
				t.Fatalf("Scan = %v, %v", payload, err)
			}
			return nil
		})
		if err != nil || !called {
			t.Fatalf("callback = (%t,%v)", called, err)
		}

	case "raw_values_are_copied":
		called := false
		err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			called = true
			rows, err := access.Query(t.Context(), "rows")
			if err != nil || !rows.Next() {
				t.Fatalf("first row = %v", err)
			}
			raw := rows.RawValues()
			values, err := rows.Values()
			fields := rows.FieldDescriptions()
			if err != nil {
				t.Fatal(err)
			}
			raw[0][0] = 9
			values[0].([]byte)[0] = 8
			fields[0].Name = "mutated"
			if got := rows.RawValues()[0]; !slices.Equal(got, []byte{1, 2}) {
				t.Fatalf("raw alias = %v", got)
			}
			values, err = rows.Values()
			if err != nil || !slices.Equal(values[0].([]byte), []byte{1, 2}) {
				t.Fatalf("value alias = %v, %v", values, err)
			}
			if got := rows.FieldDescriptions()[0].Name; got != "payload" {
				t.Fatalf("field alias = %q", got)
			}
			return nil
		})
		if err != nil || !called {
			t.Fatalf("callback = (%t,%v)", called, err)
		}

	case "callback_error":
		called := false
		err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			called = true
			if _, beginErr := access.Begin(t.Context()); beginErr != nil {
				return beginErr
			}
			return errors.New("private callback detail")
		})
		_, closes, _, rollbacks, _ := fixture.counts()
		if !called || !errors.Is(err, C12PITRDependencyFailure) || strings.Contains(err.Error(), "private") || rollbacks != 1 || closes != 1 {
			t.Fatalf("callback cleanup = called %t err %v rollback %d close %d", called, err, rollbacks, closes)
		}

	case "callback_panic":
		called := false
		err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			called = true
			if _, beginErr := access.Begin(t.Context()); beginErr != nil {
				return beginErr
			}
			panic("private panic detail")
		})
		_, closes, _, rollbacks, _ := fixture.counts()
		if !called || !errors.Is(err, C12PITRDependencyFailure) || strings.Contains(err.Error(), "panic detail") || rollbacks != 1 || closes != 1 {
			t.Fatalf("panic cleanup = called %t err %v rollback %d close %d", called, err, rollbacks, closes)
		}

	case "parent_cancel":
		fixture.commitWait = make(chan struct{})
		fixture.commitStart = make(chan struct{})
		ctx, cancel := context.WithCancel(t.Context())
		entered := make(chan struct{})
		done := make(chan error, 1)
		go func() {
			done <- c12WithAccess(ctx, state, nil, func(access C12AuthorityAccess) error {
				tx, err := access.Begin(t.Context())
				if err != nil {
					return err
				}
				close(entered)
				return tx.Commit(context.Background())
			})
		}()
		select {
		case <-entered:
		case err := <-done:
			t.Fatalf("commit did not enter: %v", err)
		}
		<-fixture.commitStart
		cancel()
		if err := <-done; !errors.Is(err, C12PITRCanceled) {
			t.Fatalf("cancel result = %v", err)
		}
		events := fixture.snapshot()
		commitStart := slices.Index(events, "commit:start")
		interruptIndex := slices.Index(events, "interrupt")
		closeIndex := slices.Index(events, "close")
		commitEnd := slices.Index(events, "commit:end")
		if commitStart < 0 || interruptIndex < commitStart || commitEnd < interruptIndex || closeIndex < commitEnd {
			t.Fatalf("cancel event order = %v", events)
		}
		if fixture.interrupts.Load() != 1 || fixture.cleanupRace.Load() {
			t.Fatalf("transport interruption = interrupts %d cleanup overlap %t", fixture.interrupts.Load(), fixture.cleanupRace.Load())
		}
		if _, _, commits, rollbacks, _ := fixture.counts(); commits != 1 || rollbacks != 0 {
			t.Fatalf("cancel commit/rollback = %d/%d", commits, rollbacks)
		}

	case "lease_deadline":
		ctx, cancel := context.WithTimeout(t.Context(), 25*time.Millisecond)
		defer cancel()
		called := false
		started := time.Now()
		err := c12WithAccess(ctx, state, nil, func(access C12AuthorityAccess) error {
			called = true
			<-ctx.Done()
			_, execErr := access.Exec(context.Background(), "after deadline")
			return execErr
		})
		if !called || !errors.Is(err, C12PITRCanceled) || time.Since(started) > time.Second {
			t.Fatalf("deadline = called %t err %v elapsed %v", called, err, time.Since(started))
		}

	case "one_active_transaction":
		called := false
		err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			called = true
			first, err := access.Begin(t.Context())
			if err != nil {
				return err
			}
			if _, err := access.Begin(t.Context()); !errors.Is(err, C12PITRWrongPhase) {
				t.Fatalf("second Begin = %v", err)
			}
			firstGeneration := first.(*c12AuthorityTx).generation
			if err := first.Rollback(t.Context()); err != nil {
				return err
			}
			second, err := access.Begin(t.Context())
			if err != nil {
				return err
			}
			if second.(*c12AuthorityTx).generation == firstGeneration {
				t.Fatal("transaction generation reused")
			}
			return second.Rollback(t.Context())
		})
		opens, _, _, rollbacks, _ := fixture.counts()
		if err != nil || !called || opens != 1 || rollbacks != 2 {
			t.Fatalf("transactions = called %t err %v opens %d rollbacks %d", called, err, opens, rollbacks)
		}

	case "one_active_callback":
		entered, release := make(chan struct{}), make(chan struct{})
		firstDone := make(chan error, 1)
		go func() {
			firstDone <- c12WithAccess(t.Context(), state, nil, func(C12AuthorityAccess) error {
				close(entered)
				<-release
				return nil
			})
		}()
		select {
		case <-entered:
		case err := <-firstDone:
			t.Fatalf("first callback did not enter: %v", err)
		}
		secondCalled := false
		if err := c12WithAccess(t.Context(), state, nil, func(C12AuthorityAccess) error { secondCalled = true; return nil }); !errors.Is(err, C12PITRWrongPhase) || secondCalled {
			t.Fatalf("second callback = called %t err %v", secondCalled, err)
		}
		close(release)
		if err := <-firstDone; err != nil {
			t.Fatal(err)
		}

	case "double_commit":
		called := false
		err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			called = true
			tx, err := access.Begin(t.Context())
			if err != nil {
				return err
			}
			if err := tx.Commit(t.Context()); err != nil {
				return err
			}
			if err := tx.Commit(t.Context()); !errors.Is(err, C12PITRWrongPhase) {
				t.Fatalf("second Commit = %v", err)
			}
			return nil
		})
		if _, _, commits, rollbacks, _ := fixture.counts(); err != nil || !called || commits != 1 || rollbacks != 0 {
			t.Fatalf("double commit = called %t err %v commits %d rollbacks %d", called, err, commits, rollbacks)
		}

	case "unconsumed_views_need_no_commit_io":
		var rows pgx.Rows
		var row pgx.Row
		var tx C12AuthorityTransaction
		called := false
		err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			called = true
			var err error
			rows, err = access.Query(t.Context(), "rows")
			if err != nil {
				return err
			}
			row = access.QueryRow(t.Context(), "row")
			tx, err = access.Begin(t.Context())
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil || !called {
			t.Fatalf("callback = (%t,%v)", called, err)
		}
		before := len(fixture.snapshot())
		var payload []byte
		if rows.Next() || !errors.Is(row.Scan(&payload), C12PITRWrongPhase) || !errors.Is(tx.Commit(t.Context()), C12PITRWrongPhase) {
			t.Fatal("expired unconsumed view remained live")
		}
		if got := len(fixture.snapshot()); got != before {
			t.Fatalf("unconsumed view did I/O: %d -> %d", before, got)
		}

	case "foreign_candidate":
		foreign := newC12AccessTestState(&c12AccessTestFixture{})
		called := false
		err := c12WithAccess(t.Context(), state, &c12CandidateBinding{owner: foreign, runGeneration: foreign.runGeneration, index: 0}, func(C12AuthorityAccess) error {
			called = true
			return nil
		})
		if !errors.Is(err, C12PITRInvalidHandle) || called {
			t.Fatalf("foreign binding = called %t err %v", called, err)
		}

	case "no_rows_is_preserved":
		fixture.rows = func() pgx.Rows { return &c12AccessTestRows{fixture: fixture, values: [][][]byte{}} }
		called := false
		err := c12WithAccess(t.Context(), state, nil, func(access C12AuthorityAccess) error {
			called = true
			var payload []byte
			err := access.QueryRow(t.Context(), "none").Scan(&payload)
			if !errors.Is(err, pgx.ErrNoRows) || !errors.Is(err, C12PITRDependencyFailure) || err.Error() != C12PITRDependencyFailure.Error() {
				t.Fatalf("no rows = %q", err)
			}
			return nil
		})
		if err != nil || !called {
			t.Fatalf("callback = (%t,%v)", called, err)
		}

	default:
		t.Fatalf("unhandled case %q", name)
	}
}
