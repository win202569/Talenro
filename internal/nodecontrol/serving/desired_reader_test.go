package serving

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/store"
)

func TestDesiredReaderRequiresBinding(t *testing.T) {
	node, _, _ := desiredRowsFixture(t, 0)
	db := &desiredAdapterDB{}
	var absent *desiredAdapterDB
	for _, candidate := range []store.DBTX{db, absent, &readerFakeEpoch{DBTX: db}, &desiredScanDB{DBTX: db}} {
		source := boundReadFunc(func(ctx context.Context, q func(context.Context, store.DBTX) error) error { return q(ctx, candidate) })
		got, err := readDesiredProjection(t.Context(), source, node)
		if err != authority.ErrAuthorityUnavailable || db.calls != 0 || !reflect.DeepEqual(got, desiredProjection{}) {
			t.Fatal("unbound DBTX accepted")
		}
	}
}

func TestDesiredReaderPreflight(t *testing.T) {
	node, _, _ := desiredRowsFixture(t, 0)
	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	calls := 0
	source := boundReadFunc(func(context.Context, func(context.Context, store.DBTX) error) error { calls++; return nil })
	var absent boundReadFunc
	for _, tc := range []struct {
		name   string
		ctx    context.Context
		source BoundAuthorityReadSource
		node   uuid.UUID
		want   error
	}{
		{"canceled", canceled, source, uuid.Nil, authority.ErrCanceled},
		{"nil context", nil, source, node, authority.ErrAuthorityUnavailable},
		{"node", t.Context(), source, uuid.Nil, authority.ErrInvalidArgument},
		{"nil source", t.Context(), nil, node, authority.ErrAuthorityUnavailable},
		{"typed nil source", t.Context(), absent, node, authority.ErrAuthorityUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := readDesiredProjection(tc.ctx, tc.source, tc.node)
			if err != tc.want || calls != 0 || !reflect.DeepEqual(got, desiredProjection{}) {
				t.Fatalf("preflight=%v", err)
			}
		})
	}
}

// A resolver seam models the cross-package private capability. Success here
// proves composition, not a live Coordinator, physical connection or database.
func TestDesiredReaderBindingOrder(t *testing.T) {
	node, rows, want := desiredRowsFixture(t, 3)
	r := &desiredAdapterRows{values: rows}
	scoped, cancel := context.WithCancel(t.Context())
	defer cancel()
	var events []string
	db := &desiredAdapterDB{rows: r, check: func(ctx context.Context, args []any) {
		events = append(events, "query")
		if ctx != scoped || args[1] != int64(19) {
			t.Fatal("wrong context or epoch")
		}
	}}
	source := boundReadFunc(func(_ context.Context, q func(context.Context, store.DBTX) error) error { return q(scoped, db) })
	got, err := readDesiredProjectionWithEpoch(t.Context(), source, node, func(actual store.DBTX) (uint64, error) {
		events = append(events, "resolve")
		if actual != db {
			t.Fatal("DBTX decorated before epoch resolution")
		}
		return 19, nil
	})
	if err != nil || db.calls != 1 || !reflect.DeepEqual(events, []string{"resolve", "query"}) || !reflect.DeepEqual(got.Snapshot(), want) {
		t.Fatalf("composition=%v events=%v", err, events)
	}
	assertDesiredAdapterClean(t, r, db)
}

func TestDesiredReaderResolverErrors(t *testing.T) {
	node, _, _ := desiredRowsFixture(t, 0)
	for _, tc := range []struct {
		name      string
		epoch     uint64
		err, want error
	}{
		{"nil", 0, nil, authority.ErrAuthorityUnavailable},
		{"zero", 0, nil, authority.ErrAuthorityUnavailable},
		{"overflow", math.MaxInt64 + 1, nil, authority.ErrAuthorityUnavailable},
		{"private", 19, errors.New("private resolver details"), authority.ErrAuthorityUnavailable},
		{"domain", 19, authority.ErrConflict, authority.ErrAuthorityUnavailable},
		{"canceled", 19, context.Canceled, authority.ErrCanceled},
		{"deadline", 19, context.DeadlineExceeded, authority.ErrCanceled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := &desiredAdapterDB{}
			source := boundReadFunc(func(ctx context.Context, q func(context.Context, store.DBTX) error) error { return q(ctx, db) })
			var resolve func(store.DBTX) (uint64, error)
			if tc.name != "nil" {
				resolve = func(store.DBTX) (uint64, error) { return tc.epoch, tc.err }
			}
			got, err := readDesiredProjectionWithEpoch(t.Context(), source, node, resolve)
			if err != tc.want || db.calls != 0 || !reflect.DeepEqual(got, desiredProjection{}) {
				t.Fatalf("resolver=%v", err)
			}
		})
	}
}

func TestDesiredReaderPostcheck(t *testing.T) {
	for _, domain := range []string{"success", "not found", "conflict"} {
		for _, post := range []string{"accept", "reject", "cancel and reject", "source deadline", "missing", "duplicate"} {
			t.Run(domain+"/"+post, func(t *testing.T) {
				node, rows, want := desiredRowsFixture(t, 3)
				r := &desiredAdapterRows{values: rows}
				domainErr := error(nil)
				if domain == "not found" {
					r.values = nil
					domainErr = authority.ErrNotFound
				}
				if domain == "conflict" {
					r.values[0].CurrentValid = false
					domainErr = authority.ErrConflict
				}
				db := &desiredAdapterDB{rows: r}
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				source := boundReadFunc(func(boundCtx context.Context, q func(context.Context, store.DBTX) error) error {
					if post == "missing" {
						return nil
					}
					if err := q(boundCtx, db); err != nil {
						t.Fatal("domain error bypassed postcheck", err)
					}
					switch post {
					case "reject":
						return errors.New("private postcheck details")
					case "cancel and reject":
						cancel()
						return authority.ErrAuthorityUnavailable
					case "source deadline":
						return context.DeadlineExceeded
					case "duplicate":
						if err := q(boundCtx, db); err != authority.ErrAuthorityUnavailable {
							t.Fatal("duplicate callback accepted")
						}
					}
					return nil
				})
				got, err := readDesiredProjectionWithEpoch(ctx, source, node, func(store.DBTX) (uint64, error) { return 19, nil })
				expected := authority.ErrAuthorityUnavailable
				if post == "accept" {
					expected = domainErr
				}
				if post == "cancel and reject" || post == "source deadline" {
					expected = authority.ErrCanceled
				}
				if err != expected {
					t.Fatalf("got %v want %v", err, expected)
				}
				if expected == nil {
					if !reflect.DeepEqual(got.Snapshot(), want) {
						t.Fatal("success lost data")
					}
				} else if !reflect.DeepEqual(got, desiredProjection{}) {
					t.Fatal("failure published data")
				}
				if post == "missing" {
					if db.calls != 0 {
						t.Fatal("query without callback")
					}
				} else {
					if db.calls != 1 {
						t.Fatal("query repeated")
					}
					assertDesiredAdapterClean(t, r, db)
				}
			})
		}
	}
}

func TestDesiredReaderScopedContext(t *testing.T) {
	node, rows, want := desiredRowsFixture(t, 3)
	r := &desiredAdapterRows{values: rows}
	db := &desiredAdapterDB{rows: r}
	source := boundReadFunc(func(ctx context.Context, q func(context.Context, store.DBTX) error) error {
		scoped, cancel := context.WithCancel(ctx)
		defer cancel()
		return q(scoped, db)
	})
	got, err := readDesiredProjectionWithEpoch(t.Context(), source, node, func(store.DBTX) (uint64, error) { return 19, nil })
	if err != nil || db.calls != 1 || !reflect.DeepEqual(got.Snapshot(), want) {
		t.Fatalf("scoped cleanup canceled request: %v", err)
	}
	assertDesiredAdapterClean(t, r, db)
}

func TestDesiredReaderPanic(t *testing.T) {
	node, rows, _ := desiredRowsFixture(t, 3)
	r := &desiredAdapterRows{values: rows}
	db := &desiredAdapterDB{rows: r}
	source := boundReadFunc(func(ctx context.Context, q func(context.Context, store.DBTX) error) error {
		if err := q(ctx, db); err != nil {
			t.Fatal(err)
		}
		panic("desired postcheck panic")
	})
	defer func() {
		if recover() != "desired postcheck panic" {
			t.Fatal("panic changed")
		}
		if db.calls != 1 {
			t.Fatal("query missing")
		}
		assertDesiredAdapterClean(t, r, db)
	}()
	_, _ = readDesiredProjectionWithEpoch(t.Context(), source, node, func(store.DBTX) (uint64, error) { return 19, nil })
}

// Observe real guard/adapter staging without a production hook. The discard
// wrapper captures the actual guard-owned pointer, including its fixed signature.
func TestDesiredReaderDiscardComposition(t *testing.T) {
	for _, mode := range []string{"success", "reject", "cancel", "panic"} {
		t.Run(mode, func(t *testing.T) {
			node, rows, want := desiredRowsFixture(t, 3)
			r := &desiredAdapterRows{values: rows}
			db := &desiredAdapterDB{rows: r}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var buffers [][]byte
			var chain []desiredPublishProjection
			var discarded *desiredProjection
			discards := 0
			source := boundReadFunc(func(boundCtx context.Context, q func(context.Context, store.DBTX) error) error {
				if err := q(boundCtx, db); err != nil {
					t.Fatal(err)
				}
				switch mode {
				case "reject":
					return authority.ErrAuthorityUnavailable
				case "cancel":
					cancel()
				case "panic":
					panic("desired staged panic")
				}
				return nil
			})
			var got desiredProjection
			var err error
			var panicValue any
			func() {
				defer func() { panicValue = recover() }()
				got, err = guardedRead(ctx, source, func(boundCtx context.Context, actual store.DBTX) (desiredProjection, error) {
					p, e := queryDesiredProjection(boundCtx, actual, node, 19)
					buffers = desiredTestBuffers(&p.snapshot)
					chain = p.snapshot.RootRotations
					return p, e
				}, func(p *desiredProjection) {
					discards++
					discarded = p
					discardDesiredProjection(p)
				})
			}()
			if db.calls != 1 || len(buffers) == 0 || len(chain) != 3 {
				t.Fatal("no staged material observed")
			}
			assertDesiredAdapterClean(t, r, db)
			if mode == "success" {
				if panicValue != nil || err != nil || discards != 0 || !reflect.DeepEqual(got.Snapshot(), want) {
					t.Fatal("success discarded")
				}
				return
			}
			if mode == "panic" {
				if panicValue != "desired staged panic" {
					t.Fatal("panic changed")
				}
			} else {
				expected := authority.ErrAuthorityUnavailable
				if mode == "cancel" {
					expected = authority.ErrCanceled
				}
				if panicValue != nil || err != expected {
					t.Fatalf("failure=%v", err)
				}
			}
			if !reflect.DeepEqual(got, desiredProjection{}) || discards != 1 || discarded == nil || !reflect.DeepEqual(*discarded, desiredProjection{}) {
				t.Fatal("staged projection retained")
			}
			for _, b := range buffers {
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
