package serving

import (
	"context"
	"encoding/hex"
	"errors"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/store"
)

// Only the database transport is doubled. The real generated query, Scan
// destinations, adapter and facts validator are exercised together.
type adapterDB struct {
	store.DBTX
	rows  pgx.Rows
	err   error
	calls int
	check func(context.Context, []any)
}

func (d *adapterDB) Query(ctx context.Context, _ string, args ...any) (pgx.Rows, error) {
	d.calls++
	if d.check != nil {
		d.check(ctx, args)
	}
	return d.rows, d.err
}

type adapterRows struct {
	pgx.Rows
	values             []store.ListExactCertificateProjectionRow
	index, closes      int
	scanErr, finalErr  error
	partial, panicScan bool
	afterScan          func()
	buffers            [][]byte
	targets            []any
}

func (r *adapterRows) Next() bool {
	if r.index == len(r.values) {
		return false
	}
	r.index++
	return true
}
func (r *adapterRows) Close()     { r.closes++ }
func (r *adapterRows) Err() error { return r.finalErr }
func (r *adapterRows) Scan(dest ...any) error {
	v := r.values[r.index-1]
	values := []any{v.NodeID, v.IssuerID, v.SerialBytes, v.LeafDerSha256, v.PublicKeySha256, v.CertificateID, v.LineageID, v.IdentityEpoch, v.InventoryIdentityEpoch, v.InventoryLineageID, v.InventoryVersion, v.AuthorityOperationID, v.AuthorityEpoch, v.AuthoritySequence, v.Status, v.ValidFrom, v.ValidUntil, v.LeafDer, v.HasServerTrust, v.TrustPurpose, v.TrustListenerKind, v.TrustDomain, v.TrustOperationID, v.TrustAuthorityEpoch, v.TrustAuthoritySequence, v.TrustBundleVersion, v.TrustBundleDigest}
	if len(dest) != len(values) {
		return errors.New("unexpected projection width")
	}
	r.targets = append(r.targets, dest...)
	for i, value := range values {
		if b, ok := value.([]byte); ok {
			b = append([]byte(nil), b...)
			r.buffers = append(r.buffers, b)
			value = b
		}
		reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(value))
		if r.partial && i == 4 {
			break
		}
	}
	if r.afterScan != nil {
		r.afterScan()
	}
	if r.panicScan {
		panic("scan panic")
	}
	return r.scanErr
}
func assertAdapterClean(t *testing.T, r *adapterRows) {
	t.Helper()
	if r.closes != 1 {
		t.Fatalf("close count=%d", r.closes)
	}
	for _, b := range r.buffers {
		for _, v := range b {
			if v != 0 {
				t.Fatal("decoded bytes retained")
			}
		}
	}
	for _, d := range r.targets {
		if !reflect.ValueOf(d).Elem().IsZero() {
			t.Fatal("scan destination retained material")
		}
	}
}

func TestCertificateAdapterQueryAndCleanup(t *testing.T) {
	for _, agent := range []bool{false, true} {
		lookup, row, want := certificateRowFixture(t, agent)
		r := &adapterRows{values: []store.ListExactCertificateProjectionRow{row}}
		ctx := t.Context()
		db := &adapterDB{rows: r, check: func(actual context.Context, args []any) {
			if actual != ctx || len(args) != 10 {
				t.Fatal("lost bound context/parameters")
			}
			if hex.EncodeToString(args[0].([]byte)) != "18823b1e06b1065ff0fb1fc8d6a7b87fa2aa01ea5c7337e75ef3b72ad0d2adee" || hex.EncodeToString(args[3].([]byte)) != "f9911591db5e19d9c4eff72a30b4fd0325416e63d8fb00f8fbd81f0860f6b0a1" {
				t.Fatal("wrong scope parameters")
			}
			domain := ""
			if agent {
				domain = "nodes.example"
			}
			if args[1] != agent || args[2] != domain || args[4] != row.NodeID || args[5] != row.IssuerID || args[9] != int64(4) {
				t.Fatal("wrong scalar binding")
			}
			if !reflect.DeepEqual(args[6], row.SerialBytes) || !reflect.DeepEqual(args[7], row.LeafDerSha256) || !reflect.DeepEqual(args[8], row.PublicKeySha256) {
				t.Fatal("wrong byte binding")
			}
		}}
		got, err := queryCertificateFacts(ctx, db, lookup, 4)
		if err != nil || !reflect.DeepEqual(got.Snapshot(), want) || db.calls != 1 {
			t.Fatalf("query failed: %v calls=%d", err, db.calls)
		}
		assertAdapterClean(t, r)
		if !reflect.DeepEqual(lookup.Input(), want.Identity) {
			t.Fatal("cleanup changed lookup")
		}
	}
}

func TestCertificateAdapterErrors(t *testing.T) {
	for _, name := range []string{"empty", "duplicate", "corrupt", "query", "query with rows", "scan", "partial scan", "iteration", "cancel query", "cancel scan", "cancel after scan", "nil rows", "typed nil rows"} {
		t.Run(name, func(t *testing.T) {
			l, row, _ := certificateRowFixture(t, true)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			r := &adapterRows{values: []store.ListExactCertificateProjectionRow{row}}
			db := &adapterDB{rows: r}
			want := authority.ErrAuthorityUnavailable
			closeExpected := true
			switch name {
			case "empty":
				r.values = nil
				want = authority.ErrNotFound
			case "duplicate":
				r.values = append(r.values, row)
				want = authority.ErrConflict
			case "corrupt":
				r.values[0].TrustDomain = "wrong.example"
				want = authority.ErrConflict
			case "query":
				db.rows = nil
				db.err = errors.New("secret database details")
				closeExpected = false
			case "query with rows":
				db.err = errors.New("secret database details")
			case "scan":
				r.scanErr = errors.New("secret scan details")
			case "partial scan":
				r.scanErr = errors.New("secret partial details")
				r.partial = true
			case "iteration":
				r.finalErr = errors.New("secret iteration details")
			case "cancel query":
				db.err = context.Canceled
				want = authority.ErrCanceled
			case "cancel scan":
				r.scanErr = context.DeadlineExceeded
				want = authority.ErrCanceled
			case "cancel after scan":
				r.afterScan = cancel
				want = authority.ErrCanceled
			case "nil rows":
				db.rows = nil
				closeExpected = false
			case "typed nil rows":
				var absent *adapterRows
				db.rows = absent
				closeExpected = false
			}
			got, err := queryCertificateFacts(ctx, db, l, 4)
			if err != want || !reflect.DeepEqual(got, CertificateFacts{}) {
				t.Fatalf("got %v want %v or leaked facts", err, want)
			}
			if closeExpected {
				assertAdapterClean(t, r)
			}
		})
	}
}

func TestCertificateAdapterPreflight(t *testing.T) {
	l, _, _ := certificateRowFixture(t, true)
	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	for _, name := range []string{"nil context", "canceled", "nil db", "typed nil db", "invalid lookup", "invalid epoch"} {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			db := &adapterDB{}
			var inputDB store.DBTX = db
			lookup := l
			epoch := uint64(4)
			want := authority.ErrInvalidArgument
			switch name {
			case "nil context":
				ctx = nil
				want = authority.ErrAuthorityUnavailable
			case "canceled":
				ctx = canceled
				want = authority.ErrCanceled
			case "nil db":
				inputDB = nil
				want = authority.ErrAuthorityUnavailable
			case "typed nil db":
				var absent *adapterDB
				inputDB = absent
				want = authority.ErrAuthorityUnavailable
			case "invalid lookup":
				lookup = CertificateLookup{}
			case "invalid epoch":
				epoch = 0
			}
			got, err := queryCertificateFacts(ctx, inputDB, lookup, epoch)
			if err != want || db.calls != 0 || !reflect.DeepEqual(got, CertificateFacts{}) {
				t.Fatalf("preflight %v calls=%d", err, db.calls)
			}
		})
	}
}

func TestCertificateAdapterPanicCleanup(t *testing.T) {
	l, row, _ := certificateRowFixture(t, true)
	r := &adapterRows{values: []store.ListExactCertificateProjectionRow{row}, partial: true, panicScan: true}
	defer func() {
		if recover() != "scan panic" {
			t.Fatal("panic swallowed/changed")
		}
		assertAdapterClean(t, r)
	}()
	_, _ = queryCertificateFacts(t.Context(), &adapterDB{rows: r}, l, 4)
}

func TestCertificateAdapterPostcheckCleanup(t *testing.T) {
	for _, reject := range []bool{false, true} {
		lookup, row, want := certificateRowFixture(t, true)
		r := &adapterRows{values: []store.ListExactCertificateProjectionRow{row}}
		db := &adapterDB{rows: r}
		var serial, der []byte
		var trust *ServerTrustFacts
		source := boundReadFunc(func(ctx context.Context, query func(context.Context, store.DBTX) error) error {
			if err := query(ctx, db); err != nil {
				t.Fatalf("query failed before postcheck: %v", err)
			}
			if reject {
				return authority.ErrAuthorityUnavailable
			}
			return nil
		})
		got, err := guardedRead(t.Context(), source, func(ctx context.Context, boundDB store.DBTX) (CertificateFacts, error) {
			facts, queryErr := queryCertificateFacts(ctx, boundDB, lookup, 4)
			serial, der, trust = facts.snapshot.Identity.SerialBytes, facts.snapshot.LeafDER, facts.snapshot.ServerTrust
			return facts, queryErr
		}, discardCertificateFacts)
		assertAdapterClean(t, r)
		if len(serial) == 0 || len(der) == 0 || trust == nil {
			t.Fatal("test did not stage owned facts")
		}
		if reject {
			if err != authority.ErrAuthorityUnavailable || !reflect.DeepEqual(got, CertificateFacts{}) {
				t.Fatal("postcheck rejection published facts")
			}
			for _, b := range append(append([]byte(nil), serial...), der...) {
				if b != 0 {
					t.Fatal("rejected staged bytes retained")
				}
			}
			if !reflect.DeepEqual(*trust, ServerTrustFacts{}) {
				t.Fatal("rejected trust retained")
			}
		} else if err != nil || !reflect.DeepEqual(got.Snapshot(), want) {
			t.Fatal("successful postcheck discarded published facts")
		}
	}
}
