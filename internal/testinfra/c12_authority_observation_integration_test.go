//go:build integration

package testinfra

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type c12ObservationTestRows struct {
	pgx.Rows
	rows   []c12AuthorityPITRDecodedRow
	at     int
	failAt int
	err    error
	closed bool
	next   func()
}

func (r *c12ObservationTestRows) Next() bool {
	if r.next != nil {
		r.next()
	}
	if r.failAt > 0 && r.at == r.failAt {
		r.err = errors.New("stream lost")
		return false
	}
	if r.at == len(r.rows) {
		return false
	}
	r.at++
	return true
}
func (r *c12ObservationTestRows) RawValues() [][]byte {
	v := r.rows[r.at-1]
	return [][]byte{[]byte(v.LSN), []byte(strconv.FormatUint(v.XID, 10)), []byte(v.Data)}
}
func (*c12ObservationTestRows) FieldDescriptions() []pgconn.FieldDescription {
	return []pgconn.FieldDescription{{DataTypeOID: pgtype.TextOID}, {DataTypeOID: pgtype.TextOID}, {DataTypeOID: pgtype.TextOID}}
}
func (r *c12ObservationTestRows) Err() error { return r.err }
func (r *c12ObservationTestRows) Close()     { r.closed = true }

type c12ObservationTestDriver struct {
	*c12AccessTestDriver
	marker    string
	committed []c12AuthorityPITRDecodedRow
	bindErr   error
}

func (d *c12ObservationTestDriver) BeginTx(ctx context.Context, opts pgx.TxOptions) (c12AccessDriver, error) {
	_, err := d.c12AccessTestDriver.BeginTx(ctx, opts)
	return d, err
}
func (d *c12ObservationTestDriver) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if sql == c12MarkerInsertSQL {
		d.fixture.event("bind")
		if len(args) != 1 {
			return pgconn.CommandTag{}, errors.New("marker arity")
		}
		marker, ok := args[0].(string)
		if !ok || len(marker) != 64 {
			return pgconn.CommandTag{}, errors.New("marker format")
		}
		d.marker = marker
		return pgconn.NewCommandTag("INSERT 0 1"), d.bindErr
	}
	return d.c12AccessTestDriver.Exec(ctx, sql, args...)
}
func (d *c12ObservationTestDriver) Commit(ctx context.Context) error {
	if d.marker != "" {
		d.committed = c12DecodedCommit(7, d.marker, "0/30")
	}
	return d.c12AccessTestDriver.Commit(ctx)
}

type c12WrappedObservationTx struct{ C12AuthorityTransaction }

type c12ObservationSQLConnector struct{ rows []c12AuthorityPITRDecodedRow }

func (c c12ObservationSQLConnector) Connect(context.Context) (driver.Conn, error) {
	return &c12ObservationSQLConn{rows: c.rows}, nil
}
func (c12ObservationSQLConnector) Driver() driver.Driver { return c12InitDriver{} }

type c12ObservationSQLConn struct{ rows []c12AuthorityPITRDecodedRow }

func (*c12ObservationSQLConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (*c12ObservationSQLConn) Close() error              { return nil }
func (*c12ObservationSQLConn) Begin() (driver.Tx, error) { return nil, errors.New("unexpected begin") }
func (c *c12ObservationSQLConn) QueryContext(_ context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
	if q != "SELECT lsn::text, xid::text, data FROM pg_catalog.pg_logical_slot_get_changes($1,NULL,65536,'include-xids','1')" || len(args) != 1 || args[0].Value != "slot" {
		return nil, errors.New("unexpected destructive query")
	}
	return &c12ObservationSQLRows{rows: c.rows}, nil
}

type c12ObservationSQLRows struct {
	rows []c12AuthorityPITRDecodedRow
	at   int
}

func (*c12ObservationSQLRows) Columns() []string { return []string{"lsn", "xid", "data"} }
func (*c12ObservationSQLRows) Close() error      { return nil }
func (r *c12ObservationSQLRows) Next(values []driver.Value) error {
	if r.at == len(r.rows) {
		return io.EOF
	}
	row := r.rows[r.at]
	r.at++
	copy(values, []driver.Value{row.LSN, strconv.FormatUint(row.XID, 10), row.Data})
	return nil
}

func newC12ObservationTest(t *testing.T) (C12AuthorityPITRController, *c12ObservationTestDriver, *int) {
	t.Helper()
	f := &c12AccessTestFixture{}
	s := newC12AccessTestState(f)
	d := &c12ObservationTestDriver{c12AccessTestDriver: &c12AccessTestDriver{fixture: f}}
	s.accessOpen = func(context.Context, *c12CandidateBinding) (c12AccessBackend, error) { f.event("open"); return d, nil }
	drains := new(int)
	s.observationQuery = func(context.Context) (pgx.Rows, error) {
		*drains++
		f.event("drain")
		return &c12ObservationTestRows{rows: d.committed}, nil
	}
	return C12AuthorityPITRController{state: s}, d, drains
}

func c12BoundObservation(t *testing.T, c C12AuthorityPITRController, commit bool) C12AuthorityPITRObservation {
	t.Helper()
	o, err := c.NewCommitObservation()
	if err != nil || o.state == nil {
		t.Fatalf("new observation: %v", err)
	}
	err = c.WithPrimaryAuthorityAccess(context.Background(), func(a C12AuthorityAccess) error {
		tx, e := a.Begin(context.Background())
		if e != nil {
			return e
		}
		if e = c.BindCommitObservation(context.Background(), o, tx); e != nil {
			return e
		}
		if commit {
			return tx.Commit(context.Background())
		}
		return tx.Rollback(context.Background())
	})
	if err != nil && !errors.Is(err, C12PITRIndeterminate) {
		t.Fatalf("bind/finish: %v", err)
	}
	return o
}

func c12DecodedCommit(xid uint64, marker, end string) []c12AuthorityPITRDecodedRow {
	return []c12AuthorityPITRDecodedRow{
		{LSN: "0/10", XID: xid, Data: fmt.Sprintf("BEGIN %d", xid)},
		{LSN: "0/18", XID: xid, Data: "table public.c12_authority_pitr_markers: INSERT: marker[text]:'" + marker + "'"},
		{LSN: end, XID: xid, Data: fmt.Sprintf("COMMIT %d", xid)},
	}
}

func TestC12AuthorityPITRCommitObservationStateMachine(t *testing.T) {
	t.Run("legacy_shared_bounded_reader", func(t *testing.T) {
		for _, count := range []int{3, 65537} {
			t.Run(strconv.Itoa(count), func(t *testing.T) {
				rows := make([]c12AuthorityPITRDecodedRow, count)
				for i := range rows {
					rows[i] = c12AuthorityPITRDecodedRow{LSN: "0/10", XID: 7, Data: "table public.other: INSERT: id[integer]:1"}
				}
				rows[count-1].Data = "COMMIT PREPARED 'test_gid'"
				db := sql.OpenDB(c12ObservationSQLConnector{rows: rows})
				defer db.Close()
				s := &c12AuthorityPITRState{database: db}
				s.descriptor.SlotName = "slot"
				got, err := c12ReadLegacyObservation(context.Background(), s)
				if count > 65536 {
					if !errors.Is(err, C12PITRIndeterminate) || !s.observationPoison {
						t.Fatalf("legacy unbounded: %d %v", len(got), err)
					}
					return
				}
				if err != nil || len(got) != 3 {
					t.Fatalf("legacy reader: %d %v", len(got), err)
				}
				facts, err := parseC12AuthorityPITRTerminal(got, C12AuthorityPITRCommitPrepared, 7, "test_gid")
				if err != nil || facts.GID != "test_gid" || facts.EndLSN != "0/10" {
					t.Fatalf("legacy parser contract: %#v %v", facts, err)
				}
			})
		}
	})
	t.Run("same_block_equal_lsn", func(t *testing.T) {
		a := &c12ObservationState{marker: strings.Repeat("a", 64), bound: true}
		rows := c12DecodedCommit(7, a.marker, "0/10")
		rows[1].LSN = "0/10"
		got, err := c12ParseObservedTransactions(rows, map[string]*c12ObservationState{a.marker: a})
		if err != nil || got[a].EndLSN != "0/10" {
			t.Fatalf("equal LSN rejected: %#v %v", got, err)
		}
	})
	t.Run("terminal_lsn_is_retained_without_rewriting", func(t *testing.T) {
		a := &c12ObservationState{marker: strings.Repeat("a", 64), bound: true}
		got, err := c12ParseObservedTransactions(c12DecodedCommit(7, a.marker, "0/2f"), map[string]*c12ObservationState{a.marker: a})
		if err != nil || got[a].EndLSN != "0/2f" {
			t.Fatalf("terminal boundary rewritten: %#v %v", got, err)
		}
	})
	t.Run("transaction_generation_cannot_transfer", func(t *testing.T) {
		c, d, _ := newC12ObservationTest(t)
		err := c.WithPrimaryAuthorityAccess(context.Background(), func(a C12AuthorityAccess) error {
			tx, err := a.Begin(context.Background())
			if err != nil {
				return err
			}
			c.state.mu.Lock()
			c.state.runGeneration++
			c.state.mu.Unlock()
			o, err := c.NewCommitObservation()
			if err != nil {
				return err
			}
			before := d.fixture.snapshot()
			if err = c.BindCommitObservation(context.Background(), o, tx); !errors.Is(err, C12PITRInvalidHandle) || !reflect.DeepEqual(before, d.fixture.snapshot()) {
				t.Fatalf("prior-generation tx transferred: %v", err)
			}
			return tx.Rollback(context.Background())
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("row_byte_and_single_row_limits", func(t *testing.T) {
		for _, tc := range []struct {
			name    string
			count   int
			payload int
			wantErr bool
		}{
			{"rows_exact", 65536, 1, false}, {"rows_over", 65537, 1, true},
			{"single_exact", 1, 524155, false}, {"single_over", 1, 524156, true},
			{"bytes_exact", 64, 524155, false}, {"bytes_over", 65, 524155, true},
		} {
			t.Run(tc.name, func(t *testing.T) {
				stream := &c12ObservationTestRows{rows: make([]c12AuthorityPITRDecodedRow, tc.count)}
				for i := range stream.rows {
					stream.rows[i] = c12AuthorityPITRDecodedRow{LSN: "0/10", XID: 7, Data: strings.Repeat("x", tc.payload)}
				}
				got, err := c12ReadObservedRows(context.Background(), stream)
				if tc.wantErr {
					if !errors.Is(err, C12PITRCapacityExceeded) {
						t.Fatalf("over budget accepted: %d %v", len(got), err)
					}
				} else if err != nil || len(got) != tc.count {
					t.Fatalf("exact budget rejected: %d %v", len(got), err)
				}
			})
		}
	})
	t.Run("capped_complete_response_not_slot_empty", func(t *testing.T) {
		c, d, drains := newC12ObservationTest(t)
		o := c12BoundObservation(t, c, true)
		rows := make([]c12AuthorityPITRDecodedRow, 65536)
		rows[0] = c12AuthorityPITRDecodedRow{LSN: "0/10", XID: 8, Data: "BEGIN 8"}
		for i := 1; i < len(rows)-1; i++ {
			rows[i] = c12AuthorityPITRDecodedRow{LSN: "0/18", XID: 8, Data: "table public.other: INSERT: id[integer]:1"}
		}
		rows[len(rows)-1] = c12AuthorityPITRDecodedRow{LSN: "0/20", XID: 8, Data: "COMMIT 8"}
		pending := d.committed
		d.committed = rows
		if _, _, err := c.ObserveCommit(context.Background(), o); !errors.Is(err, C12PITRNotObserved) || *drains != 1 || c.state.observationPoison {
			t.Fatalf("bounded complete response: %v", err)
		}
		d.committed = pending
		if _, facts, err := c.ObserveCommit(context.Background(), o); err != nil || facts.EndLSN != "0/30" || *drains != 2 {
			t.Fatalf("later response proof: %#v %v", facts, err)
		}
	})
	t.Run("over_cap_and_incomplete_response_poison", func(t *testing.T) {
		for _, count := range []int{65536, 65537} {
			t.Run(strconv.Itoa(count), func(t *testing.T) {
				c, d, _ := newC12ObservationTest(t)
				o := c12BoundObservation(t, c, true)
				d.committed = make([]c12AuthorityPITRDecodedRow, count)
				d.committed[0] = c12AuthorityPITRDecodedRow{LSN: "0/10", XID: 8, Data: "BEGIN 8"}
				for i := 1; i < count; i++ {
					d.committed[i] = c12AuthorityPITRDecodedRow{LSN: "0/18", XID: 8, Data: "table public.other: INSERT: id[integer]:1"}
				}
				if count > 65536 {
					d.committed[count-1].Data = "COMMIT 8"
				}
				if _, _, err := c.ObserveCommit(context.Background(), o); !errors.Is(err, C12PITRIndeterminate) || !c.state.observationPoison {
					t.Fatalf("unsafe capped response: %v", err)
				}
			})
		}
	})
	t.Run("all_results_installed_and_diagnostics_survive_poison", func(t *testing.T) {
		c, d, drains := newC12ObservationTest(t)
		a := c12BoundObservation(t, c, true)
		first := d.committed
		c.state.phase.Store(2)
		b := c12BoundObservation(t, c, true)
		second := c12DecodedCommit(8, b.state.marker, "0/40")
		d.committed = append(first, second...)
		_, aFacts, err := c.ObserveCommit(context.Background(), a)
		if err != nil {
			t.Fatal(err)
		}
		bHandle, bFacts, err := c.ObserveCommit(context.Background(), b)
		if err != nil || bHandle.state == nil || bFacts.SQLXID != 8 || *drains != 1 {
			t.Fatalf("discarded other result: %#v %v %d", bFacts, err, *drains)
		}
		pending := c12BoundObservation(t, c, true)
		d.committed = d.committed[:2]
		if _, _, err = c.ObserveCommit(context.Background(), pending); !errors.Is(err, C12PITRIndeterminate) {
			t.Fatal(err)
		}
		before := *drains
		if _, again, e := c.ObserveCommit(context.Background(), a); e != nil || again != aFacts || *drains != before {
			t.Fatalf("diagnostic lost: %#v %v", again, e)
		}
		unbound := C12AuthorityPITRObservation{state: &c12ObservationState{owner: c.state, generation: 1}}
		err = c.WithPrimaryAuthorityAccess(context.Background(), func(access C12AuthorityAccess) error {
			tx, e := access.Begin(context.Background())
			if e != nil {
				return e
			}
			c.state.mu.Lock()
			c.state.observations[3] = unbound.state
			c.state.mu.Unlock()
			events := d.fixture.snapshot()
			if e = c.BindCommitObservation(context.Background(), unbound, tx); !errors.Is(e, C12PITRIndeterminate) || !reflect.DeepEqual(events, d.fixture.snapshot()) {
				t.Fatalf("poison bind performed SQL: %v", e)
			}
			return tx.Rollback(context.Background())
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("drain_reservation_excludes_new_access_and_drain", func(t *testing.T) {
		c, _, _ := newC12ObservationTest(t)
		o := c12BoundObservation(t, c, true)
		started := make(chan struct{})
		release := make(chan struct{})
		done := make(chan error, 1)
		c.state.observationQuery = func(ctx context.Context) (pgx.Rows, error) {
			close(started)
			select {
			case <-release:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			return &c12ObservationTestRows{}, nil
		}
		go func() { _, _, err := c.ObserveCommit(context.Background(), o); done <- err }()
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("drain did not start")
		}
		if err := c.WithPrimaryAuthorityAccess(context.Background(), func(C12AuthorityAccess) error { t.Error("callback entered during drain"); return nil }); !errors.Is(err, C12PITRWrongPhase) {
			t.Fatalf("new access: %v", err)
		}
		if _, _, err := c.ObserveCommit(context.Background(), o); !errors.Is(err, C12PITRWrongPhase) {
			t.Fatalf("concurrent drain: %v", err)
		}
		if _, err := c.NewCommitObservation(); !errors.Is(err, C12PITRWrongPhase) {
			t.Fatalf("new during drain: %v", err)
		}
		close(release)
		if err := <-done; !errors.Is(err, C12PITRNotObserved) {
			t.Fatal(err)
		}
	})
	t.Run("legacy_family_is_one_way", func(t *testing.T) {
		c, _, _ := newC12ObservationTest(t)
		c.state.phase.Store(2)
		if err := c12BeginLegacyObservation(c.state); err != nil {
			t.Fatal(err)
		}
		c.state.endC12Transition()
		if _, err := c.NewCommitObservation(); !errors.Is(err, C12PITRWrongPhase) {
			t.Fatalf("new after legacy: %v", err)
		}
		if err := c12BeginLegacyObservation(c.state); err != nil {
			t.Fatal("legacy sequence broken", err)
		}
		c.state.endC12Transition()
	})
	t.Run("zero_foreign_and_wrapped_transaction", func(t *testing.T) {
		c, _, _ := newC12ObservationTest(t)
		other, _, _ := newC12ObservationTest(t)
		if _, err := (C12AuthorityPITRController{}).NewCommitObservation(); !errors.Is(err, C12PITRInvalidHandle) {
			t.Fatal("zero controller accepted")
		}
		o, err := c.NewCommitObservation()
		if err != nil || o.state == nil {
			t.Fatalf("new: %v", err)
		}
		err = c.WithPrimaryAuthorityAccess(context.Background(), func(a C12AuthorityAccess) error {
			tx, e := a.Begin(context.Background())
			if e != nil {
				return e
			}
			for _, bad := range []C12AuthorityTransaction{nil, (*c12AuthorityTx)(nil), c12WrappedObservationTx{tx}} {
				if e = c.BindCommitObservation(context.Background(), o, bad); !errors.Is(e, C12PITRInvalidHandle) {
					t.Fatalf("bad tx accepted: %T %v", bad, e)
				}
			}
			if e = other.BindCommitObservation(context.Background(), o, tx); !errors.Is(e, C12PITRInvalidHandle) {
				t.Fatalf("foreign: %v", e)
			}
			if e = c.BindCommitObservation(context.Background(), C12AuthorityPITRObservation{}, tx); !errors.Is(e, C12PITRInvalidHandle) {
				t.Fatalf("zero: %v", e)
			}
			return tx.Rollback(context.Background())
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("capacity_64_no_side_effect_65", func(t *testing.T) {
		c, d, _ := newC12ObservationTest(t)
		random := 0
		c.state.observationRandom = func(p []byte) (int, error) {
			random++
			for i := range p {
				p[i] = byte(random)
			}
			return len(p), nil
		}
		for i := 0; i < 64; i++ {
			o, e := c.NewCommitObservation()
			if e != nil || o.state == nil {
				t.Fatalf("capacity %d: %v", i, e)
			}
		}
		before := d.fixture.snapshot()
		if _, e := c.NewCommitObservation(); !errors.Is(e, C12PITRCapacityExceeded) {
			t.Fatalf("65: %v", e)
		}
		if random != 64 || !reflect.DeepEqual(before, d.fixture.snapshot()) {
			t.Fatal("capacity generated randomness or SQL")
		}
	})
	for _, name := range []string{"copy_cannot_rebind", "closed_transaction", "failed_binding_never_rebinds", "consume_then_one_driver_commit"} {
		t.Run(name, func(t *testing.T) {
			c, d, _ := newC12ObservationTest(t)
			o, e := c.NewCommitObservation()
			if e != nil || o.state == nil {
				t.Fatalf("new: %v", e)
			}
			e = c.WithPrimaryAuthorityAccess(context.Background(), func(a C12AuthorityAccess) error {
				tx, e := a.Begin(context.Background())
				if e != nil {
					return e
				}
				if name == "closed_transaction" {
					if e = tx.Rollback(context.Background()); e != nil {
						return e
					}
					if e = c.BindCommitObservation(context.Background(), o, tx); !errors.Is(e, C12PITRWrongPhase) {
						t.Fatalf("closed: %v", e)
					}
					return nil
				}
				if name == "failed_binding_never_rebinds" {
					d.bindErr = errors.New("lost bind response")
				}
				e = c.BindCommitObservation(context.Background(), o, tx)
				if name == "failed_binding_never_rebinds" {
					if !errors.Is(e, C12PITRDependencyFailure) {
						t.Fatalf("failed bind: %v", e)
					}
					d.bindErr = nil
				} else if e != nil {
					return e
				}
				if name == "consume_then_one_driver_commit" {
					rows, err := tx.Query(context.Background(), "SELECT fixture")
					if err != nil {
						return err
					}
					for rows.Next() {
					}
					rows.Close()
					d.fixture.event("consume")
					if e = tx.Commit(context.Background()); e != nil {
						return e
					}
					if e = tx.Commit(context.Background()); !errors.Is(e, C12PITRWrongPhase) {
						t.Fatalf("second commit: %v", e)
					}
					events := d.fixture.snapshot()
					want := []string{"open", "begin:read committed", "bind", "query", "cursor:next", "cursor:next", "cursor:close", "consume", "commit:start", "commit:end"}
					if !reflect.DeepEqual(events, want) {
						t.Fatalf("commit path: %v", events)
					}
					return nil
				}
				copied := o
				if e = c.BindCommitObservation(context.Background(), copied, tx); !errors.Is(e, C12PITRWrongPhase) {
					t.Fatalf("copy rebound: %v", e)
				}
				second, _ := c.NewCommitObservation()
				if e = c.BindCommitObservation(context.Background(), second, tx); !errors.Is(e, C12PITRWrongPhase) {
					t.Fatalf("tx rebound: %v", e)
				}
				if e = tx.Rollback(context.Background()); e != nil {
					return e
				}
				next, e := a.Begin(context.Background())
				if e != nil {
					return e
				}
				if e = c.BindCommitObservation(context.Background(), copied, next); !errors.Is(e, C12PITRWrongPhase) {
					t.Fatalf("other tx rebound: %v", e)
				}
				return next.Rollback(context.Background())
			})
			if e != nil {
				t.Fatal(e)
			}
		})
	}
	for _, name := range []string{"rollback_not_observed", "commit_response_lost", "observe_response_lost_same_facts", "completed_handle_after_generation_change"} {
		t.Run(name, func(t *testing.T) {
			c, d, drains := newC12ObservationTest(t)
			if name == "commit_response_lost" {
				d.fixture.commitErr = errors.New("commit response lost")
				c.state.fixture = &c12FixtureLedger{}
			}
			o := c12BoundObservation(t, c, name != "rollback_not_observed")
			if name == "observe_response_lost_same_facts" {
				c.state.observationResponse = func() error { return errors.New("response lost") }
			}
			handle, facts, err := c.ObserveCommit(context.Background(), o)
			if name == "rollback_not_observed" {
				if !errors.Is(err, C12PITRNotObserved) || handle.state != nil {
					t.Fatalf("rollback proved: %v %v", facts, err)
				}
				return
			}
			if name == "observe_response_lost_same_facts" {
				if !errors.Is(err, C12PITRDependencyFailure) {
					t.Fatalf("response seam: %v", err)
				}
				c.state.observationResponse = nil
				handle, facts, err = c.ObserveCommit(context.Background(), o)
			}
			if err != nil || handle.state == nil || facts != (C12AuthorityPITRCommit{Kind: C12AuthorityPITRCommitImmediate, SQLXID: 7, EndLSN: "0/30"}) {
				t.Fatalf("facts: %#v %v", facts, err)
			}
			if name == "commit_response_lost" {
				_, _, commits, _, _ := d.fixture.counts()
				if commits != 1 || !c.state.writesUncertain {
					t.Fatal("uncertain actual commit not retained")
				}
			}
			facts.EndLSN = "mutated"
			_, again, err := c.ObserveCommit(context.Background(), o)
			if err != nil || again.EndLSN != "0/30" || *drains != 1 {
				t.Fatalf("retry changed facts/io: %#v %v %d", again, err, *drains)
			}
			if name == "completed_handle_after_generation_change" {
				c.state.runGeneration++
				if _, _, err = c.ObserveCommit(context.Background(), o); !errors.Is(err, C12PITRInvalidHandle) {
					t.Fatalf("generation transferred: %v", err)
				}
			}
		})
	}
	for _, name := range []string{"partial_drain_poison", "cancel_before_query", "cancel_after_query_poison", "query_error_poison", "all_results_install_or_none", "active_callback_exclusion", "legacy_mixing_exclusion"} {
		t.Run(name, func(t *testing.T) {
			c, d, drains := newC12ObservationTest(t)
			o := c12BoundObservation(t, c, true)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			switch name {
			case "partial_drain_poison":
				c.state.observationQuery = func(context.Context) (pgx.Rows, error) {
					*drains++
					return &c12ObservationTestRows{rows: d.committed, failAt: 2}, nil
				}
			case "cancel_before_query":
				cancel()
			case "cancel_after_query_poison":
				c.state.observationQuery = func(context.Context) (pgx.Rows, error) {
					*drains++
					cancel()
					return &c12ObservationTestRows{rows: d.committed}, nil
				}
			case "query_error_poison":
				c.state.observationQuery = func(context.Context) (pgx.Rows, error) { *drains++; return nil, errors.New("query lost") }
			case "all_results_install_or_none":
				d.committed = append(d.committed, c12DecodedCommit(8, strings.Repeat("f", 64), "bad")...)
			case "active_callback_exclusion":
				err := c.WithPrimaryAuthorityAccess(ctx, func(C12AuthorityAccess) error {
					_, _, e := c.ObserveCommit(ctx, o)
					if !errors.Is(e, C12PITRWrongPhase) || *drains != 0 {
						t.Fatalf("active drain: %v %d", e, *drains)
					}
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
				return
			case "legacy_mixing_exclusion":
				c.state.phase.Store(2)
				// A missing guard yields an assertion failure, not a nil-DB panic.
				c.state.database, _ = sql.Open("pgx", "")
				_ = c.state.database.Close()
				if _, err := issueC12AuthorityPITRCommit(ctx, c, C12AuthorityPITRCommitPrepared); !errors.Is(err, C12PITRWrongPhase) {
					t.Fatalf("legacy commit mixed: %v", err)
				}
				if err := issueC12AuthorityPITRRollbackProbe(ctx, c); !errors.Is(err, C12PITRWrongPhase) {
					t.Fatalf("rollback mixed: %v", err)
				}
				return
			}
			_, _, err := c.ObserveCommit(ctx, o)
			if name == "cancel_before_query" {
				if !errors.Is(err, C12PITRCanceled) || *drains != 0 || c.state.observationPoison {
					t.Fatalf("pre-query cancellation: %v %d", err, *drains)
				}
				return
			}
			if !errors.Is(err, C12PITRIndeterminate) || !c.state.observationPoison || o.state.result != nil {
				t.Fatalf("partial response not poisoned: %v", err)
			}
			before := *drains
			if _, _, err = c.ObserveCommit(context.Background(), o); !errors.Is(err, C12PITRIndeterminate) || *drains != before {
				t.Fatalf("poison retried: %v", err)
			}
			if _, err = c.NewCommitObservation(); !errors.Is(err, C12PITRIndeterminate) {
				t.Fatalf("poison allocated: %v", err)
			}
			c.state.phase.Store(2)
			dockerBefore := c.state.dockerCalls.Load()
			if _, err = c.CrashPrimary(context.Background(), C12AuthorityPITRBaseBackup{}, C12AuthorityPITRCommit{Kind: C12AuthorityPITRCommitImmediate, SQLXID: 7, EndLSN: "0/30"}); !errors.Is(err, C12PITRIndeterminate) {
				t.Fatalf("poison crash: %v", err)
			}
			if c.state.phase.Load() != 2 || c.state.dockerCalls.Load() != dockerBefore {
				t.Fatal("poison crash had side effects")
			}
		})
	}
	t.Run("multi_result_xid_wrap", func(t *testing.T) {
		a := &c12ObservationState{marker: strings.Repeat("a", 64), bound: true}
		b := &c12ObservationState{marker: strings.Repeat("b", 64), bound: true}
		rows := append(c12DecodedCommit(4294967295, a.marker, "0/20"), c12DecodedCommit(3, b.marker, "0/40")...)
		got, err := c12ParseObservedTransactions(rows, map[string]*c12ObservationState{a.marker: a, b.marker: b})
		if err != nil || len(got) != 2 || got[a].EndLSN != "0/20" || got[b].EndLSN != "0/40" || got[a].SQLXID != 4294967295 || got[b].SQLXID != 3 {
			t.Fatalf("lost or mixed transaction: %#v %v", got, err)
		}
	})
	for _, name := range []string{"ambiguous_marker", "missing_terminal", "duplicate_terminal", "prepared_terminal_rejected", "cross_block_marker", "text_xid_mismatch", "update_marker", "delete_marker", "invalid_lsn", "backward_lsn", "marker_suffix"} {
		t.Run(name, func(t *testing.T) {
			a := &c12ObservationState{marker: strings.Repeat("a", 64), bound: true}
			rows := c12DecodedCommit(7, a.marker, "0/20")
			switch name {
			case "ambiguous_marker":
				rows = append(rows, c12DecodedCommit(8, a.marker, "0/40")...)
			case "missing_terminal":
				rows = rows[:2]
			case "duplicate_terminal":
				rows = append(rows, rows[2])
			case "prepared_terminal_rejected":
				rows[2].Data = "COMMIT PREPARED 'probe', txid 7"
			case "cross_block_marker":
				rows[1].XID = 8
			case "text_xid_mismatch":
				rows[2].Data = "COMMIT 8"
			case "update_marker":
				rows[1].Data = strings.Replace(rows[1].Data, "INSERT:", "UPDATE:", 1)
			case "delete_marker":
				rows[1].Data = strings.Replace(rows[1].Data, "INSERT:", "DELETE:", 1)
			case "invalid_lsn":
				rows[1].LSN = "bad"
			case "backward_lsn":
				rows[2].LSN = "0/17"
			case "marker_suffix":
				rows[1].Data += " extra[text]:'x'"
			}
			if _, err := c12ParseObservedTransactions(rows, map[string]*c12ObservationState{a.marker: a}); err == nil {
				t.Fatal("malformed block proved commit")
			}
		})
	}
	t.Run("unknown_complete_and_numeric_lsn", func(t *testing.T) {
		a := &c12ObservationState{marker: strings.Repeat("a", 64), bound: true}
		rows := c12DecodedCommit(7, strings.Repeat("f", 64), "0/20")
		block := c12DecodedCommit(8, a.marker, "1/0")
		block[1].LSN = "0/FFFFFFFE"
		got, err := c12ParseObservedTransactions(append(rows, block...), map[string]*c12ObservationState{a.marker: a})
		if err != nil || len(got) != 1 || got[a].EndLSN != "1/0" {
			t.Fatalf("unknown/numeric block: %#v %v", got, err)
		}
	})
}
