package serving

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"

	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/store"
)

type readerFakeEpoch struct{ store.DBTX }

func (*readerFakeEpoch) BoundReadEpoch() (uint64, error) { return 4, nil }

func TestCertificateReaderRequiresBinding(t *testing.T) {
	l, _, _ := certificateRowFixture(t, true)
	db := &adapterDB{}
	var absent *adapterDB
	for _, candidate := range []store.DBTX{db, absent, &readerFakeEpoch{DBTX: db}} {
		source := boundReadFunc(func(ctx context.Context, q func(context.Context, store.DBTX) error) error { return q(ctx, candidate) })
		got, err := readCertificateFacts(t.Context(), source, l)
		if err != authority.ErrAuthorityUnavailable || db.calls != 0 || !reflect.DeepEqual(got, CertificateFacts{}) {
			t.Fatal("unbound data accepted")
		}
	}
}

// The resolver double models the unavailable cross-package private source.
// Real guard, adapter, generated query and validation run; this is NOT a live
// Coordinator/DB test or proof of production readiness.
func TestCertificateReaderBoundEpoch(t *testing.T) {
	for _, agent := range []bool{false, true} {
		l, row, want := certificateRowFixture(t, agent)
		row.AuthorityEpoch = 19
		want.Authority.Epoch = 19
		if agent {
			row.TrustAuthorityEpoch = 19
			want.ServerTrust.AuthorityEpoch = 19
		}
		r := &adapterRows{values: []store.ListExactCertificateProjectionRow{row}}
		db := &adapterDB{rows: r, check: func(_ context.Context, args []any) {
			if args[9] != int64(19) {
				t.Fatal("wrong bound epoch")
			}
		}}
		resolves := 0
		source := boundReadFunc(func(ctx context.Context, q func(context.Context, store.DBTX) error) error { return q(ctx, db) })
		got, err := readCertificateFactsWithEpoch(t.Context(), source, l, func(actual store.DBTX) (uint64, error) {
			resolves++
			if actual != db {
				t.Fatal("resolver got another DBTX")
			}
			return 19, nil
		})
		if err != nil || resolves != 1 || db.calls != 1 || !reflect.DeepEqual(got.Snapshot(), want) {
			t.Fatalf("bound read failed: %v", err)
		}
		assertAdapterClean(t, r)
	}
}

func TestCertificateReaderErrors(t *testing.T) {
	l, _, _ := certificateRowFixture(t, true)
	for _, canceled := range []bool{false, true} {
		for _, name := range []string{"nil", "zero", "overflow", "private", "canceled"} {
			ctx, cancel := context.WithCancel(t.Context())
			if canceled {
				cancel()
			}
			db := &adapterDB{}
			source := boundReadFunc(func(ctx context.Context, q func(context.Context, store.DBTX) error) error { return q(ctx, db) })
			var resolve func(store.DBTX) (uint64, error)
			switch name {
			case "zero":
				resolve = func(store.DBTX) (uint64, error) { return 0, nil }
			case "overflow":
				resolve = func(store.DBTX) (uint64, error) { return math.MaxInt64 + 1, nil }
			case "private":
				resolve = func(store.DBTX) (uint64, error) { return 4, errors.New("private resolver details") }
			case "canceled":
				resolve = func(store.DBTX) (uint64, error) { return 4, context.Canceled }
			}
			got, err := readCertificateFactsWithEpoch(ctx, source, l, resolve)
			want := authority.ErrAuthorityUnavailable
			if canceled || name == "canceled" {
				want = authority.ErrCanceled
			}
			if err != want || db.calls != 0 || !reflect.DeepEqual(got, CertificateFacts{}) {
				t.Fatalf("%s cancel=%v: %v", name, canceled, err)
			}
			cancel()
		}
	}
}

func TestCertificateReaderPostcheck(t *testing.T) {
	for _, name := range []string{"success", "reject", "cancel", "missing", "duplicate", "domain", "domain rejected"} {
		t.Run(name, func(t *testing.T) {
			l, row, want := certificateRowFixture(t, true)
			r := &adapterRows{values: []store.ListExactCertificateProjectionRow{row}}
			if name == "domain" || name == "domain rejected" {
				r.values = nil
			}
			db := &adapterDB{rows: r}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			source := boundReadFunc(func(ctx context.Context, q func(context.Context, store.DBTX) error) error {
				if name == "missing" {
					return nil
				}
				if e := q(ctx, db); e != nil {
					t.Fatal("query bypassed postcheck")
				}
				switch name {
				case "reject", "domain rejected":
					return authority.ErrAuthorityUnavailable
				case "cancel":
					cancel()
				case "duplicate":
					_ = q(ctx, db)
				}
				return nil
			})
			got, err := readCertificateFactsWithEpoch(ctx, source, l, func(store.DBTX) (uint64, error) { return 4, nil })
			expected := authority.ErrAuthorityUnavailable
			switch name {
			case "success":
				expected = nil
			case "cancel":
				expected = authority.ErrCanceled
			case "domain":
				expected = authority.ErrNotFound
			}
			if err != expected {
				t.Fatalf("got %v want %v", err, expected)
			}
			if name == "success" {
				if !reflect.DeepEqual(got.Snapshot(), want) {
					t.Fatal("success lost facts")
				}
			} else if !reflect.DeepEqual(got, CertificateFacts{}) {
				t.Fatal("failed read published facts")
			}
			if name == "missing" {
				if db.calls != 0 {
					t.Fatal("query without callback")
				}
			} else {
				if db.calls != 1 {
					t.Fatal("duplicate query")
				}
				assertAdapterClean(t, r)
			}
		})
	}
}

func TestCertificateReaderPanic(t *testing.T) {
	l, row, _ := certificateRowFixture(t, true)
	r := &adapterRows{values: []store.ListExactCertificateProjectionRow{row}}
	db := &adapterDB{rows: r}
	source := boundReadFunc(func(ctx context.Context, q func(context.Context, store.DBTX) error) error {
		if e := q(ctx, db); e != nil {
			t.Fatal(e)
		}
		panic("postcheck panic")
	})
	defer func() {
		if recover() != "postcheck panic" {
			t.Fatal("panic changed")
		}
		if db.calls != 1 {
			t.Fatal("query did not run")
		}
		assertAdapterClean(t, r)
	}()
	_, _ = readCertificateFactsWithEpoch(t.Context(), source, l, func(store.DBTX) (uint64, error) { return 4, nil })
}

// Observe owned facts through the real guard/adapter composition, without
// introducing a production hook that leaks the reader's private staged facts.
func TestCertificateReaderPanicDiscardComposition(t *testing.T) {
	l, row, _ := certificateRowFixture(t, true)
	r := &adapterRows{values: []store.ListExactCertificateProjectionRow{row}}
	db := &adapterDB{rows: r}
	var serial, der []byte
	var trust *ServerTrustFacts
	source := boundReadFunc(func(ctx context.Context, q func(context.Context, store.DBTX) error) error {
		if e := q(ctx, db); e != nil {
			t.Fatal(e)
		}
		panic("staged panic")
	})
	defer func() {
		if recover() != "staged panic" {
			t.Fatal("panic changed")
		}
		if len(serial) == 0 || len(der) == 0 || trust == nil {
			t.Fatal("facts never staged")
		}
		for _, b := range [][]byte{serial, der} {
			for _, v := range b {
				if v != 0 {
					t.Fatal("staged bytes survived panic")
				}
			}
		}
		if !reflect.DeepEqual(*trust, ServerTrustFacts{}) {
			t.Fatal("trust survived panic")
		}
		assertAdapterClean(t, r)
	}()
	_, _ = guardedRead(t.Context(), source, func(ctx context.Context, db store.DBTX) (CertificateFacts, error) {
		facts, e := queryCertificateFacts(ctx, db, l, 4)
		serial, der, trust = facts.snapshot.Identity.SerialBytes, facts.snapshot.LeafDER, facts.snapshot.ServerTrust
		return facts, e
	}, discardCertificateFacts)
}
