package authority

import (
	"context"
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"talenro.local/platform/internal/store"
)

type epochTransport struct {
	store.DBTX
	check func(context.Context, string, []any)
	calls int
	err   error
	rows  pgx.Rows
	row   pgx.Row
}

func (d *epochTransport) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	d.calls++
	d.check(ctx, sql, args)
	return pgconn.NewCommandTag("UPDATE 3"), d.err
}
func (d *epochTransport) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	d.calls++
	d.check(ctx, sql, args)
	return d.rows, d.err
}
func (d *epochTransport) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	d.calls++
	d.check(ctx, sql, args)
	return d.row
}

type epochFake struct{ store.DBTX }

func (*epochFake) BoundReadEpoch() (uint64, error) { return 19, nil }

type epochTestRow struct{ pgx.Row }
type epochTestRows struct{ pgx.Rows }

func requireExpiredEpoch(t *testing.T, db store.DBTX) {
	t.Helper()
	if n, err := BoundReadEpoch(db); n != 0 || err != ErrAuthorityUnavailable {
		t.Fatalf("unbound accepted: %d %v", n, err)
	}
}

func TestBoundReadEpochBinding(t *testing.T) {
	ctx := t.Context()
	transport := &epochTransport{err: errors.New("transport error"), rows: &epochTestRows{}, row: &epochTestRow{}}
	transport.check = func(got context.Context, sql string, args []any) {
		if got != ctx || sql != "fixture" || !reflect.DeepEqual(args, []any{17, "x"}) {
			t.Fatal("forwarding changed context/query/args")
		}
	}
	conn := &boundTestConnection{DBTX: transport, token: new(pgconn.PgConn)}
	checks, acquires := 0, 0
	var captured store.DBTX
	source := &boundAuthorityReadSource{
		acquire: func(context.Context) (boundReadConnection, error) { acquires++; return conn, nil },
		check: func(_ context.Context, db store.DBTX) (boundReadSnapshot, error) {
			checks++
			if db != conn {
				t.Fatal("checks changed connection")
			}
			if checks == 2 {
				requireExpiredEpoch(t, captured)
			}
			s := boundTestSnapshot()
			s.readiness.ProviderHead.Epoch = 19
			s.readiness.DatabaseHead.Epoch = 19
			return s, nil
		},
	}
	err := source.WithConsistentReadyRead(ctx, func(actual context.Context, db store.DBTX) error {
		captured = db
		if n, e := BoundReadEpoch(db); n != 19 || e != nil || actual != ctx {
			t.Fatalf("missing bound epoch: %d %v", n, e)
		}
		tag, e := db.Exec(ctx, "fixture", 17, "x")
		if tag.String() != "UPDATE 3" || e != transport.err {
			t.Fatal("exec result changed")
		}
		rows, e := db.Query(ctx, "fixture", 17, "x")
		if rows != transport.rows || e != transport.err {
			t.Fatal("query result changed")
		}
		if db.QueryRow(ctx, "fixture", 17, "x") != transport.row {
			t.Fatal("row result changed")
		}
		return nil
	})
	if err != nil || acquires != 1 || checks != 2 || conn.releases != 1 || transport.calls != 3 {
		t.Fatalf("binding flow: %v %d %d %d %d", err, acquires, checks, conn.releases, transport.calls)
	}
	requireExpiredEpoch(t, captured)
}

func TestBoundReadEpochExactType(t *testing.T) {
	var absent *boundEpochDB
	var nilTransport *epochTransport
	valid := &boundEpochDB{db: &epochTransport{}}
	valid.epoch.Store(19)
	for _, db := range []store.DBTX{nil, absent, &epochTransport{}, &epochFake{}, &epochFake{DBTX: valid}, &struct{ store.DBTX }{valid}, &boundEpochDB{}, &boundEpochDB{db: nilTransport}} {
		requireExpiredEpoch(t, db)
	}
	for _, n := range []uint64{0, math.MaxInt64 + 1} {
		valid.epoch.Store(n)
		requireExpiredEpoch(t, valid)
	}
}

func TestBoundReadEpochRange(t *testing.T) {
	for _, tc := range []struct {
		provider, database uint64
		ok                 bool
	}{
		{0, 0, false}, {19, 0, false}, {19, 20, false}, {math.MaxInt64 + 1, math.MaxInt64 + 1, false},
		{1, 1, true}, {19, 19, true}, {math.MaxInt64, math.MaxInt64, true},
	} {
		conn := &boundTestConnection{token: new(pgconn.PgConn)}
		called := 0
		source := &boundAuthorityReadSource{
			acquire: func(context.Context) (boundReadConnection, error) { return conn, nil },
			check: func(context.Context, store.DBTX) (boundReadSnapshot, error) {
				s := boundTestSnapshot()
				s.readiness.ProviderHead.Epoch = tc.provider
				s.readiness.DatabaseHead.Epoch = tc.database
				return s, nil
			},
		}
		err := source.WithConsistentReadyRead(t.Context(), func(_ context.Context, db store.DBTX) error {
			called++
			n, e := BoundReadEpoch(db)
			if e != nil || n != tc.provider {
				t.Fatalf("epoch=%d err=%v", n, e)
			}
			return nil
		})
		if tc.ok {
			if err != nil || called != 1 {
				t.Fatalf("valid pair rejected: %+v %v", tc, err)
			}
		} else if err != ErrAuthorityUnavailable || called != 0 {
			t.Fatalf("invalid pair accepted: %+v %v", tc, err)
		}
		if conn.releases != 1 {
			t.Fatal("connection not released")
		}
	}
}

func TestBoundReadEpochExpiry(t *testing.T) {
	for _, name := range []string{"return", "panic", "cancel"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			conn := &boundTestConnection{token: new(pgconn.PgConn)}
			source := &boundAuthorityReadSource{acquire: func(context.Context) (boundReadConnection, error) { return conn, nil }, check: func(context.Context, store.DBTX) (boundReadSnapshot, error) { return boundTestSnapshot(), nil }}
			var captured store.DBTX
			func() {
				defer func() {
					p := recover()
					if name == "panic" && p != "epoch panic" {
						t.Fatal("panic lost")
					}
					if name != "panic" && p != nil {
						t.Fatal(p)
					}
				}()
				err := source.WithConsistentReadyRead(ctx, func(_ context.Context, db store.DBTX) error {
					captured = db
					if n, e := BoundReadEpoch(db); n != 7 || e != nil {
						t.Fatal("no active binding")
					}
					if name == "panic" {
						panic("epoch panic")
					}
					if name == "cancel" {
						cancel()
					}
					return nil
				})
				if name == "cancel" {
					if err != ErrCanceled {
						t.Fatal(err)
					}
				} else if err != nil {
					t.Fatal(err)
				}
			}()
			requireExpiredEpoch(t, captured)
			if conn.releases != 1 {
				t.Fatal("release lost")
			}
		})
	}
}

func TestBoundReadEpochConcurrentExpiry(t *testing.T) {
	db := &boundEpochDB{db: &epochTransport{}}
	db.epoch.Store(19)
	if n, e := BoundReadEpoch(db); n != 19 || e != nil {
		t.Fatal("inactive before expiry")
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 100; j++ {
				n, e := BoundReadEpoch(db)
				if !((n == 19 && e == nil) || (n == 0 && e == ErrAuthorityUnavailable)) {
					t.Error("invalid atomic state")
				}
			}
		}()
	}
	close(start)
	db.expire()
	wg.Wait()
	requireExpiredEpoch(t, db)
}
