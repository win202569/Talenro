//go:build integration

package testinfra

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var c12AccessCases = []string{
	"eager_rows_and_expired_views", "queryrow_is_not_lazy", "raw_values_are_copied",
	"callback_error", "callback_panic", "parent_cancel", "lease_deadline",
	"one_active_transaction", "one_active_callback", "double_commit",
	"unconsumed_views_need_no_commit_io", "foreign_candidate", "no_rows_is_preserved",
}

type c12AccessTestFixture struct {
	mu           sync.Mutex
	events       []string
	opens        int
	closes       int
	commits      int
	rollbacks    int
	cursorCloses int
	rows         func() pgx.Rows
	queryErr     error
	commitErr    error
	commitCtxErr error
	commitWait   chan struct{}
	commitStart  chan struct{}
	closeOnce    sync.Once
	startOnce    sync.Once
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
	fixture *c12AccessTestFixture
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
	d.fixture.mu.Lock()
	d.fixture.rollbacks++
	d.fixture.events = append(d.fixture.events, "rollback")
	d.fixture.mu.Unlock()
	return nil
}

func (d *c12AccessTestDriver) Close(context.Context) error {
	d.fixture.mu.Lock()
	d.fixture.closes++
	d.fixture.events = append(d.fixture.events, "close")
	d.fixture.mu.Unlock()
	if d.fixture.commitWait != nil {
		d.fixture.closeOnce.Do(func() { close(d.fixture.commitWait) })
	}
	return nil
}

func (*c12AccessTestDriver) TypeMap() *pgtype.Map { return pgtype.NewMap() }

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
	fixture *c12AccessTestFixture
	values  [][][]byte
	index   int
	errAt   int
	err     error
	closed  bool
	started bool
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
		closeIndex := slices.Index(events, "close")
		commitEnd := slices.Index(events, "commit:end")
		if commitStart < 0 || closeIndex < commitStart || commitEnd < closeIndex {
			t.Fatalf("cancel event order = %v", events)
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
