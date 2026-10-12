package serving

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/store"
)

type desiredAdapterDB struct {
	store.DBTX
	rows  pgx.Rows
	err   error
	calls int
	args  []any
	check func(context.Context, []any)
}

func (d *desiredAdapterDB) Query(ctx context.Context, _ string, args ...any) (pgx.Rows, error) {
	d.calls++
	d.args = args
	if d.check != nil {
		d.check(ctx, args)
	}
	return d.rows, d.err
}

type desiredAdapterRows struct {
	pgx.Rows
	values                       []store.ListExactDesiredProjectionRow
	index, scans, closes         int
	buffers                      [][]byte
	targets                      []any
	scanErr, finalErr            error
	failRow, panicRow, panicNext int
	panicClose, endless          bool
	afterScan, onErr             func()
}

func (r *desiredAdapterRows) Next() bool {
	if r.panicNext > 0 && r.index == r.panicNext {
		panic("desired next panic")
	}
	if !r.endless && r.index == len(r.values) {
		return false
	}
	r.index++
	return true
}
func (r *desiredAdapterRows) Err() error {
	if r.onErr != nil {
		r.onErr()
	}
	return r.finalErr
}
func (r *desiredAdapterRows) Close() {
	r.closes++
	if r.panicClose {
		panic("desired close panic")
	}
}
func (r *desiredAdapterRows) Scan(dest ...any) error {
	r.scans++
	v := r.values[(r.index-1)%len(r.values)]
	values := []any{
		v.NodeID, v.PointerGroupsValid, v.HasDesired, v.CurrentValid, v.ActiveDesiredGeneration, v.ActiveRootPublishID, v.ActiveRootVersion, v.ActiveMetadataPublishID, v.ActiveMetadataVersion, v.HasRotation, v.RotationExpectedVersion, v.RotationValid,
		v.DesiredNodeID, v.DesiredSigningID, v.DesiredAuthorityOperationID, v.DesiredRootPublishID, v.DesiredMetadataPublishID, v.DesiredGeneration, v.DesiredInventoryVersion, v.DesiredResourceEnvelopeVersion, v.DesiredRootVersion, v.DesiredMetadataVersion, v.DesiredAuthorityEpoch, v.DesiredAuthoritySequence, v.DesiredResourceEnvelopeDigest, v.DesiredSigningKeyID, v.DesiredPayloadDigest, v.DesiredSignature, v.DesiredPayload, v.DesiredIssuedAt, v.DesiredEffectiveDeadline, v.DesiredValidUntil,
		v.RootPublishID, v.RootAuthorityOperationID, v.RootVersion, v.RootBaseRootVersion, v.RootBaseMetadataVersion, v.RootAuthorityEpoch, v.RootAuthoritySequence, v.RootPayload, v.RootPublishedEnvelope, v.RootPayloadDigest, v.RootEnvelopeDigest,
		v.MetadataPublishID, v.MetadataAuthorityOperationID, v.MetadataVersion, v.MetadataBaseRootVersion, v.MetadataBaseMetadataVersion, v.MetadataAuthorityEpoch, v.MetadataAuthoritySequence, v.MetadataPayload, v.MetadataPublishedEnvelope, v.MetadataPayloadDigest, v.MetadataEnvelopeDigest,
		v.RotationPublishID, v.RotationAuthorityOperationID, v.RotationVersion, v.RotationBaseRootVersion, v.RotationBaseMetadataVersion, v.RotationAuthorityEpoch, v.RotationAuthoritySequence, v.RotationPayload, v.RotationPublishedEnvelope, v.RotationPayloadDigest, v.RotationEnvelopeDigest,
	}
	if len(dest) != len(values) {
		return fmt.Errorf("wrong scan width")
	}
	r.targets = append(r.targets, dest...)
	for i, value := range values {
		if b, ok := value.([]byte); ok {
			b = bytes.Clone(b)
			r.buffers = append(r.buffers, b)
			value = b
		}
		reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(value))
		if i == 39 && r.index == r.panicRow {
			panic("desired scan panic")
		}
		if i == 39 && r.index == r.failRow {
			return r.scanErr
		}
	}
	if r.afterScan != nil {
		r.afterScan()
	}
	return nil
}
func assertDesiredAdapterClean(t *testing.T, r *desiredAdapterRows, db *desiredAdapterDB) {
	t.Helper()
	if r != nil {
		if r.closes != 1 {
			t.Fatalf("closes=%d", r.closes)
		}
		for _, b := range r.buffers {
			for _, v := range b {
				if v != 0 {
					t.Fatal("decoded material retained")
				}
			}
		}
		for _, target := range r.targets {
			if !reflect.ValueOf(target).Elem().IsZero() {
				t.Fatal("scan target retained")
			}
		}
	}
	if db != nil && len(db.args) == 4 {
		for _, i := range []int{2, 3} {
			for _, v := range db.args[i].([]byte) {
				if v != 0 {
					t.Fatal("argument digest retained")
				}
			}
		}
	}
}
func TestDesiredAdapterTransport(t *testing.T) {
	for _, where := range []string{"query", "rows+error", "scan first", "scan middle", "iteration"} {
		for _, fault := range []error{errors.New("private details"), authority.ErrNotFound, authority.ErrConflict, context.Canceled, context.DeadlineExceeded} {
			t.Run(where+"/"+fault.Error(), func(t *testing.T) {
				node, rows, _ := desiredRowsFixture(t, 3)
				r := &desiredAdapterRows{values: rows}
				db := &desiredAdapterDB{rows: r}
				switch where {
				case "query":
					db.rows = nil
					db.err = fault
				case "rows+error":
					db.err = fault
				case "scan first":
					r.failRow = 1
					r.scanErr = fault
				case "scan middle":
					r.failRow = 2
					r.scanErr = fault
				case "iteration":
					r.finalErr = fault
				}
				got, err := queryDesiredProjection(t.Context(), db, node, 19)
				want := authority.ErrAuthorityUnavailable
				if readCanceled(fault) {
					want = authority.ErrCanceled
				}
				if err != want || !reflect.DeepEqual(got, desiredProjection{}) {
					t.Fatalf("error=%v want %v", err, want)
				}
				if where == "query" {
					assertDesiredAdapterClean(t, nil, db)
				} else {
					assertDesiredAdapterClean(t, r, db)
				}
			})
		}
	}
	node, _, _ := desiredRowsFixture(t, 0)
	var typed *desiredAdapterRows
	for _, r := range []pgx.Rows{nil, typed} {
		db := &desiredAdapterDB{rows: r}
		got, err := queryDesiredProjection(t.Context(), db, node, 19)
		if err != authority.ErrAuthorityUnavailable || !reflect.DeepEqual(got, desiredProjection{}) {
			t.Fatal("nil rows accepted")
		}
		assertDesiredAdapterClean(t, nil, db)
	}
}
func TestDesiredAdapterCleanup(t *testing.T) {
	node, rows, want := desiredRowsFixture(t, 3)
	_, before, _ := desiredRowsFixture(t, 3)
	r := &desiredAdapterRows{values: rows}
	db := &desiredAdapterDB{rows: r}
	p, err := queryDesiredProjection(t.Context(), db, node, 19)
	if err != nil {
		t.Fatal(err)
	}
	assertDesiredAdapterClean(t, r, db)
	if !reflect.DeepEqual(rows, before) || !reflect.DeepEqual(p.Snapshot(), want) {
		t.Fatal("cleanup changed owned output or fixture")
	}
	s := p.Snapshot()
	for _, b := range desiredTestBuffers(&s) {
		old := b[0]
		b[0] ^= 127
		if !reflect.DeepEqual(p.Snapshot(), want) {
			t.Fatal("snapshot alias")
		}
		b[0] = old
	}
}
func TestDesiredAdapterPanic(t *testing.T) {
	for _, at := range []string{"scan first", "scan middle", "next", "close"} {
		t.Run(at, func(t *testing.T) {
			node, rows, _ := desiredRowsFixture(t, 3)
			r := &desiredAdapterRows{values: rows}
			db := &desiredAdapterDB{rows: r}
			want := "desired scan panic"
			switch at {
			case "scan first":
				r.panicRow = 1
			case "scan middle":
				r.panicRow = 2
			case "next":
				r.panicNext = 1
				want = "desired next panic"
			case "close":
				r.panicClose = true
				want = "desired close panic"
			}
			defer func() {
				if recover() != want {
					t.Fatal("panic lost")
				}
				assertDesiredAdapterClean(t, r, db)
			}()
			_, _ = queryDesiredProjection(t.Context(), db, node, 19)
		})
	}
}
func TestDesiredAdapterRowLimit(t *testing.T) {
	for _, n := range []int{33, 34, -1} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			node, rows, _ := desiredRowsFixture(t, 1)
			r := &desiredAdapterRows{values: rows}
			if n < 0 {
				r.endless = true
			} else {
				r.values = make([]store.ListExactDesiredProjectionRow, n)
				for i := range r.values {
					r.values[i] = rows[0]
				}
			}
			db := &desiredAdapterDB{rows: r}
			got, err := queryDesiredProjection(t.Context(), db, node, 19)
			want := authority.ErrAuthorityUnavailable
			if n == 33 {
				want = authority.ErrConflict
			}
			if err != want || !reflect.DeepEqual(got, desiredProjection{}) || r.scans != 33 || db.calls != 1 {
				t.Fatalf("limit err=%v scans=%d", err, r.scans)
			}
			assertDesiredAdapterClean(t, r, db)
		})
	}
}

// A synchronous context script observes the final post-conversion check without
// sleeps or a mutable production hook. Ordinary cancellation cases use real contexts.
type desiredLateContext struct {
	context.Context
	checks int
}

func (c *desiredLateContext) Err() error {
	c.checks++
	if c.checks >= 3 {
		return context.Canceled
	}
	return nil
}
func TestDesiredAdapterCancellation(t *testing.T) {
	for _, at := range []string{"query", "scan", "iteration", "late", "late domain"} {
		t.Run(at, func(t *testing.T) {
			node, rows, _ := desiredRowsFixture(t, 3)
			r := &desiredAdapterRows{values: rows}
			db := &desiredAdapterDB{rows: r}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var use context.Context = ctx
			switch at {
			case "query":
				db.check = func(context.Context, []any) { cancel() }
			case "scan":
				r.afterScan = cancel
			case "iteration":
				r.onErr = cancel
			case "late", "late domain":
				use = &desiredLateContext{Context: ctx}
				if at == "late domain" {
					r.values = nil
				}
			}
			got, err := queryDesiredProjection(use, db, node, 19)
			if err != authority.ErrCanceled || !reflect.DeepEqual(got, desiredProjection{}) {
				t.Fatal("cancellation published")
			}
			assertDesiredAdapterClean(t, r, db)
		})
	}
}

func TestDesiredAdapterQuery(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			node, rows, want := desiredRowsFixture(t, n)
			db := &desiredAdapterDB{rows: &desiredAdapterRows{values: rows}, check: func(ctx context.Context, args []any) {
				if args[0] != node || args[1] != int64(19) {
					t.Fatal("wrong node/epoch")
				}
				if hex.EncodeToString(args[2].([]byte)) != "7b0f433d1b28ac2429e776263de6da84eb1e7e5969333b9bf4a8a6035f460549" ||
					hex.EncodeToString(args[3].([]byte)) != "f9911591db5e19d9c4eff72a30b4fd0325416e63d8fb00f8fbd81f0860f6b0a1" {
					t.Fatal("wrong scopes")
				}
			}}
			got, err := queryDesiredProjection(t.Context(), db, node, 19)
			if err != nil || db.calls != 1 || !reflect.DeepEqual(got.Snapshot(), want) {
				t.Fatalf("projection: %v", err)
			}
		})
	}
}
func TestDesiredAdapterPreflight(t *testing.T) {
	node, _, _ := desiredRowsFixture(t, 0)
	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	var absent *desiredAdapterDB
	for _, tc := range []struct {
		name  string
		ctx   context.Context
		db    store.DBTX
		node  uuid.UUID
		epoch uint64
		want  error
	}{
		{"canceled", canceled, nil, uuid.Nil, 0, authority.ErrCanceled},
		{"nil context", nil, &desiredAdapterDB{}, node, 19, authority.ErrAuthorityUnavailable},
		{"nil db", t.Context(), nil, node, 19, authority.ErrAuthorityUnavailable},
		{"typed nil db", t.Context(), absent, node, 19, authority.ErrAuthorityUnavailable},
		{"node", t.Context(), &desiredAdapterDB{}, uuid.Nil, 19, authority.ErrInvalidArgument},
		{"epoch zero", t.Context(), &desiredAdapterDB{}, node, 0, authority.ErrInvalidArgument},
		{"epoch overflow", t.Context(), &desiredAdapterDB{}, node, math.MaxInt64 + 1, authority.ErrInvalidArgument},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := queryDesiredProjection(tc.ctx, tc.db, tc.node, tc.epoch)
			if err != tc.want || !reflect.DeepEqual(got, desiredProjection{}) {
				t.Fatalf("error %v want %v", err, tc.want)
			}
			if d, ok := tc.db.(*desiredAdapterDB); ok && d != nil && d.calls != 0 {
				t.Fatal("query during preflight")
			}
		})
	}
}

// Exercise the private ownership method directly: the generated query normally
// closes before conversion, so a post-conversion Close panic is defensive only.
func TestDesiredAdapterLifetime(t *testing.T) {
	for _, mode := range []string{"success", "cancel", "close panic"} {
		t.Run(mode, func(t *testing.T) {
			node, rows, want := desiredRowsFixture(t, 3)
			p, err := desiredProjectionFromRows(node, 19, rows)
			if err != nil {
				t.Fatal(err)
			}
			r := &desiredAdapterRows{panicClose: mode == "close panic"}
			life := desiredQueryLifetime{db: desiredScanDB{rows: &desiredScanRows{Rows: r}}, staged: p}
			aliases := desiredTestBuffers(&life.staged.snapshot)
			chain := life.staged.snapshot.RootRotations
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if mode == "cancel" {
				cancel()
			}
			var out desiredProjection
			var finishErr error
			var panicValue any
			func() {
				defer func() { panicValue = recover() }()
				defer life.cleanup()
				out, finishErr = life.finish(ctx)
			}()
			life.cleanup()
			if r.closes != 1 || !reflect.DeepEqual(life.staged, desiredProjection{}) {
				t.Fatal("lifetime cleanup failed")
			}
			if mode == "success" {
				if panicValue != nil || finishErr != nil || !reflect.DeepEqual(out.Snapshot(), want) {
					t.Fatal("success lost material")
				}
				return
			}
			if mode == "close panic" && panicValue != "desired close panic" {
				t.Fatal("panic changed")
			}
			if mode == "cancel" && (panicValue != nil || finishErr != authority.ErrCanceled) {
				t.Fatal("cancellation lost")
			}
			if !reflect.DeepEqual(out, desiredProjection{}) {
				t.Fatal("failed finish published")
			}
			for _, b := range aliases {
				for _, v := range b {
					if v != 0 {
						t.Fatal("staged bytes retained")
					}
				}
			}
			for _, v := range chain {
				if !reflect.DeepEqual(v, desiredPublishProjection{}) {
					t.Fatal("staged chain retained")
				}
			}
		})
	}
}
