package authority

import (
	"context"
	"errors"
	"math/big"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/store"
)

// These doubles replace the external connection and readiness observations;
// the acquisition/order/comparison/error/release algorithm is real.
type boundTestConnection struct {
	store.DBTX
	token    *pgconn.PgConn
	closed   bool
	releases int
}

func (c *boundTestConnection) readIdentity() *pgconn.PgConn { return c.token }
func (c *boundTestConnection) usable() bool                 { return !c.closed }
func (c *boundTestConnection) Release()                     { c.releases++ }

func boundTestSnapshot() boundReadSnapshot {
	return boundReadSnapshot{
		readiness: Readiness{Ready: true, Reason: ReadinessReady,
			ProviderHead: Head{Epoch: 7}, DatabaseHead: DatabaseHead{Epoch: 7}},
		point: DatabasePoint{SystemID: 41, Timeline: 3, RequiredLSN: "0/30"},
	}
}

func TestBoundReadOrderAndErrors(t *testing.T) {
	private := errors.New("private connection details")
	for _, tc := range []struct {
		name              string
		queryErr, postErr error
		cancel, close     bool
		want              error
		wantPost          bool
	}{
		{"success", nil, nil, false, false, nil, true},
		{"not-found-postchecked", ErrNotFound, nil, false, false, ErrNotFound, true},
		{"corrupt-postchecked", ErrConflict, nil, false, false, ErrConflict, true},
		{"unknown-query-postchecked", private, nil, false, false, ErrAuthorityUnavailable, true},
		{"authority-over-domain", ErrNotFound, private, false, false, ErrAuthorityUnavailable, true},
		{"joined-authority-over-domain", errors.Join(ErrConflict, ErrAuthorityUnavailable), nil, false, false, ErrAuthorityUnavailable, true},
		{"source-canceled", ErrNotFound, context.Canceled, false, false, ErrCanceled, true},
		{"query-canceled", context.Canceled, nil, false, false, ErrCanceled, false},
		{"query-deadline", context.DeadlineExceeded, nil, false, false, ErrCanceled, false},
		{"caller-canceled", ErrConflict, private, true, false, ErrCanceled, false},
		{"connection-closed", ErrNotFound, nil, false, true, ErrAuthorityUnavailable, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			conn := &boundTestConnection{token: new(pgconn.PgConn)}
			var events []string
			checks := 0
			source := &boundAuthorityReadSource{
				acquire: func(context.Context) (boundReadConnection, error) {
					events = append(events, "acquire")
					return conn, nil
				},
				check: func(ctx context.Context, db store.DBTX) (boundReadSnapshot, error) {
					if db != conn {
						t.Fatal("readiness escaped acquired connection")
					}
					checks++
					if checks == 1 {
						events = append(events, "pre")
						return boundTestSnapshot(), nil
					}
					events = append(events, "post")
					return boundTestSnapshot(), tc.postErr
				},
			}
			err := source.WithConsistentReadyRead(ctx, func(actualCtx context.Context, db store.DBTX) error {
				if actualCtx != ctx || db != conn {
					t.Fatal("query escaped bound context/connection")
				}
				events = append(events, "query")
				if tc.cancel {
					cancel()
				}
				conn.closed = tc.close
				return tc.queryErr
			})
			wantEvents := []string{"acquire", "pre", "query"}
			if tc.wantPost {
				wantEvents = append(wantEvents, "post")
			}
			if err != tc.want || !reflect.DeepEqual(events, wantEvents) || conn.releases != 1 {
				t.Fatalf("error=%v events=%v releases=%d; want %v %v one release", err, events, conn.releases, tc.want, wantEvents)
			}
		})
	}
}

func TestBoundReadRejectsChangedSnapshot(t *testing.T) {
	for _, name := range []string{"not-ready", "provider-epoch", "database-epoch", "database-record-count", "database-system", "database-timeline", "connection-replaced", "closed-after-post", "malformed-point"} {
		t.Run(name, func(t *testing.T) {
			conn := &boundTestConnection{token: new(pgconn.PgConn)}
			checks := 0
			source := &boundAuthorityReadSource{
				acquire: func(context.Context) (boundReadConnection, error) { return conn, nil },
				check: func(context.Context, store.DBTX) (boundReadSnapshot, error) {
					checks++
					snapshot := boundTestSnapshot()
					if checks == 2 {
						switch name {
						case "not-ready":
							snapshot.readiness.Ready = false
						case "provider-epoch":
							snapshot.readiness.ProviderHead.Epoch++
						case "database-epoch":
							snapshot.readiness.DatabaseHead.Epoch++
						case "database-record-count":
							snapshot.readiness.DatabaseHead.RecordCount++
						case "database-system":
							snapshot.point.SystemID++
						case "database-timeline":
							snapshot.point.Timeline++
						case "closed-after-post":
							conn.closed = true
						case "malformed-point":
							snapshot.point.RequiredLSN = "invalid"
						}
					}
					return snapshot, nil
				},
			}
			err := source.WithConsistentReadyRead(t.Context(), func(context.Context, store.DBTX) error {
				if name == "connection-replaced" {
					conn.token = new(pgconn.PgConn)
				}
				return ErrNotFound
			})
			if err != ErrAuthorityUnavailable || conn.releases != 1 {
				t.Fatalf("error=%v releases=%d", err, conn.releases)
			}
		})
	}
}

func TestBoundReadPrecheckStopsQuery(t *testing.T) {
	for _, mode := range []string{"acquire-error", "acquire-cancel", "acquire-error-with-connection", "acquire-cancel-with-connection", "nil-connection", "nil-token", "closed", "precheck-error", "precheck-not-ready", "nil-query", "nil-context", "canceled-context"} {
		t.Run(mode, func(t *testing.T) {
			conn := &boundTestConnection{token: new(pgconn.PgConn)}
			ctx := t.Context()
			want := ErrAuthorityUnavailable
			source := &boundAuthorityReadSource{
				acquire: func(context.Context) (boundReadConnection, error) {
					switch mode {
					case "acquire-error":
						return nil, errors.New("private")
					case "acquire-cancel":
						return nil, context.Canceled
					case "acquire-error-with-connection":
						return conn, errors.New("private")
					case "acquire-cancel-with-connection":
						return conn, context.Canceled
					case "nil-connection":
						return nil, nil
					case "nil-token":
						conn.token = nil
					case "closed":
						conn.closed = true
					}
					return conn, nil
				},
				check: func(context.Context, store.DBTX) (boundReadSnapshot, error) {
					if mode == "precheck-error" {
						return boundReadSnapshot{}, ErrNotFound
					}
					s := boundTestSnapshot()
					if mode == "precheck-not-ready" {
						s.readiness.Ready = false
					}
					return s, nil
				},
			}
			query := func(context.Context, store.DBTX) error { t.Fatal("query ran before successful precheck"); return nil }
			if mode == "nil-query" {
				query = nil
			}
			if mode == "nil-context" {
				ctx = nil
			}
			if mode == "canceled-context" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
				want = ErrCanceled
			}
			if mode == "acquire-cancel" || mode == "acquire-cancel-with-connection" {
				want = ErrCanceled
			}
			if err := source.WithConsistentReadyRead(ctx, query); err != want {
				t.Fatalf("error=%v, want %v", err, want)
			}
			wantReleases := 1
			switch mode {
			case "acquire-error", "acquire-cancel", "nil-connection", "nil-query", "nil-context", "canceled-context":
				wantReleases = 0
			}
			if conn.releases != wantReleases {
				t.Fatalf("releases=%d, want %d", conn.releases, wantReleases)
			}
		})
	}
}

// Removing the deferred release must fail each case, without swallowing the
// original panic or treating an ordinary return as a successful test.
func TestBoundReadReleasesOnPanic(t *testing.T) {
	for _, stage := range []string{"precheck", "query", "postcheck"} {
		t.Run(stage, func(t *testing.T) {
			conn := &boundTestConnection{token: new(pgconn.PgConn)}
			marker := errors.New("owned panic marker")
			checks, queries := 0, 0
			source := &boundAuthorityReadSource{
				acquire: func(context.Context) (boundReadConnection, error) { return conn, nil },
				check: func(context.Context, store.DBTX) (boundReadSnapshot, error) {
					checks++
					if (stage == "precheck" && checks == 1) || (stage == "postcheck" && checks == 2) {
						panic(marker)
					}
					return boundTestSnapshot(), nil
				},
			}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				_ = source.WithConsistentReadyRead(t.Context(), func(context.Context, store.DBTX) error {
					queries++
					if stage == "query" {
						panic(marker)
					}
					return nil
				})
			}()
			wantChecks, wantQueries := 1, 1
			if stage == "precheck" {
				wantQueries = 0
			}
			if stage == "postcheck" {
				wantChecks = 2
			}
			if recovered != marker || conn.releases != 1 || checks != wantChecks || queries != wantQueries {
				t.Fatalf("panic=%v releases=%d checks=%d queries=%d", recovered, conn.releases, checks, queries)
			}
		})
	}
}

// Keep the database point independent, so removing only the provider-head copy
// is enough to make the in-place-change case wrongly succeed.
func TestBoundReadFreezesProviderHead(t *testing.T) {
	for _, changed := range []bool{false, true} {
		name := "unchanged"
		if changed {
			name = "in-place-change"
		}
		t.Run(name, func(t *testing.T) {
			conn := &boundTestConnection{token: new(pgconn.PgConn)}
			point := DatabasePoint{SystemID: 41, Timeline: 3, RequiredLSN: "0/10"}
			dbPoint := point
			snapshot := boundTestSnapshot()
			operation := uuid.MustParse("8f472e51-8135-4288-b7d1-4e95ba4bd210")
			snapshot.readiness.ProviderHead = Head{Epoch: 7, LatestReservedSequence: 1, LatestCommittedSequence: 1,
				LatestReservationDigest: contracts.Digest{1}, LatestCommittedOperationID: operation,
				LatestCommittedReceiptDigest: contracts.Digest{2}, LatestCommittedDatabasePoint: &point}
			snapshot.readiness.DatabaseHead = DatabaseHead{Epoch: 7, RecordCount: 1, LatestReservedSequence: 1, LatestCommittedSequence: 1,
				LatestReservationDigest: contracts.Digest{1}, LatestCommittedOperationID: operation,
				LatestCommittedReceiptDigest: contracts.Digest{2}, LatestCommittedDatabasePoint: &dbPoint}
			if snapshot.readiness.ProviderHead.Validate() != nil || !validDatabaseHead(snapshot.readiness.DatabaseHead) {
				t.Fatal("invalid positive fixture")
			}
			checks, queries := 0, 0
			source := &boundAuthorityReadSource{
				acquire: func(context.Context) (boundReadConnection, error) { return conn, nil },
				check: func(context.Context, store.DBTX) (boundReadSnapshot, error) {
					checks++
					if changed && checks == 2 {
						point.RequiredLSN = "0/20"
					}
					return snapshot, nil
				},
			}
			err := source.WithConsistentReadyRead(t.Context(), func(context.Context, store.DBTX) error { queries++; return nil })
			var want error
			if changed {
				want = ErrAuthorityUnavailable
			}
			if err != want || checks != 2 || queries != 1 || conn.releases != 1 {
				t.Fatalf("error=%v checks=%d queries=%d releases=%d, want %v", err, checks, queries, conn.releases, want)
			}
		})
	}
}

func TestBoundReadCoordinatorFactoryRejectsMismatch(t *testing.T) {
	coordinator, _, _, _ := newReadyCoordinatorFixture(t)
	if source, err := coordinator.newBoundReadSource(); err != ErrAuthorityUnavailable || source != nil {
		t.Fatal("accepted non-Postgres repository")
	}
	repository, err := NewPostgresRepository(&pgxpool.Pool{})
	if err != nil {
		t.Fatal(err)
	}
	coordinator.repository = repository
	if source, err := coordinator.newBoundReadSource(); err != ErrAuthorityUnavailable || source != nil {
		t.Fatal("accepted a different transaction repository")
	}
	coordinator.transaction = repository
	if source, err := coordinator.newBoundReadSource(); err != nil || source == nil {
		t.Fatalf("matching private factory: %v", err)
	}
	repository.database = coordinatorNoopDBTX{}
	if source, err := coordinator.newBoundReadSource(); err != ErrAuthorityUnavailable || source != nil {
		t.Fatal("accepted fake pool")
	}
}

func TestBoundReadReadinessUsesOnlyAcquiredDBTX(t *testing.T) {
	provider, err := NewDeterministicProvider(7)
	if err != nil {
		t.Fatal(err)
	}
	originalDB := &repositoryRecordingDBTX{forbidUse: true}
	repository, err := NewPostgresRepository(originalDB)
	if err != nil {
		t.Fatal(err)
	}
	coordinator := &Coordinator{provider: provider, repository: repository}
	pointRow := repositoryValuesRow{values: []any{pgtype.Numeric{Int: big.NewInt(41), Valid: true}, int64(3), "0/30"}}
	acquired := &repositoryRecordingDBTX{
		queryRows:   [][]any{},
		rowSequence: []pgx.Row{pointRow, pointRow, repositoryValuesRow{values: make([]any, 12)}},
	}
	snapshot, err := coordinator.checkReadyOnReadConnection(t.Context(), acquired)
	if err != nil || !snapshot.readiness.Ready || snapshot.point.SystemID != 41 {
		t.Fatalf("bound readiness = %#v, %v", snapshot, err)
	}
	if originalDB.calls != 0 || acquired.calls != 5 || len(acquired.rowSequence) != 0 {
		t.Fatalf("original DB calls=%d, acquired calls=%d, unused rows=%d", originalDB.calls, acquired.calls, len(acquired.rowSequence))
	}
	if acquired.beginCalls != 0 || acquired.commitCalls != 0 || acquired.rollbackCalls != 0 {
		t.Fatal("readiness opened an explicit transaction")
	}
}

func TestBoundReadWALProgressAndFrozenHeads(t *testing.T) {
	for _, mode := range []string{"advance", "regress", "aliased-head-change"} {
		t.Run(mode, func(t *testing.T) {
			conn := &boundTestConnection{token: new(pgconn.PgConn)}
			snapshot := boundTestSnapshot()
			// A mutable DB point is enough to exercise defensive snapshot copying;
			// the fake readiness engine is deliberately outside this unit's scope.
			snapshot.readiness.DatabaseHead.LatestCommittedDatabasePoint = &DatabasePoint{SystemID: 41, Timeline: 3, RequiredLSN: "0/10"}
			calls := 0
			source := &boundAuthorityReadSource{
				acquire: func(context.Context) (boundReadConnection, error) { return conn, nil },
				check: func(context.Context, store.DBTX) (boundReadSnapshot, error) {
					calls++
					if calls == 2 {
						switch mode {
						case "advance":
							snapshot.point.RequiredLSN = "0/40"
						case "regress":
							snapshot.point.RequiredLSN = "0/20"
						case "aliased-head-change":
							snapshot.readiness.DatabaseHead.LatestCommittedDatabasePoint.RequiredLSN = "0/20"
						}
					}
					return snapshot, nil
				},
			}
			err := source.WithConsistentReadyRead(t.Context(), func(context.Context, store.DBTX) error { return nil })
			var want error = ErrAuthorityUnavailable
			if mode == "advance" {
				want = nil
			}
			if err != want || calls != 2 || conn.releases != 1 {
				t.Fatalf("error=%v checks=%d releases=%d", err, calls, conn.releases)
			}
		})
	}
}
