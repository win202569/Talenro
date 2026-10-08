package serving

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/store"
)

// Only the external bound capability is replaced. These tests exercise the
// reader's result/error publication boundary, not PostgreSQL readiness itself.
type boundReadFunc func(context.Context, func(context.Context, store.DBTX) error) error

func (f boundReadFunc) WithConsistentReadyRead(ctx context.Context, query func(context.Context, store.DBTX) error) error {
	return f(ctx, query)
}

type servingTestDB struct{ store.DBTX }
type servingTestFacts struct{ material []byte }

func discardServingTestFacts(f *servingTestFacts) {
	clear(f.material)
	*f = servingTestFacts{}
}

func TestGuardedReadOrder(t *testing.T) {
	for _, domainErr := range []error{nil, authority.ErrNotFound, authority.ErrConflict} {
		t.Run(fmt.Sprint(domainErr), func(t *testing.T) {
			var order []string
			db := &servingTestDB{}
			material := []byte("private fixture")
			source := boundReadFunc(func(ctx context.Context, query func(context.Context, store.DBTX) error) error {
				order = append(order, "pre")
				// The consumer must hold domain errors privately rather than
				// letting a source short-circuit before its post-check.
				if err := query(ctx, db); err != nil {
					t.Fatalf("domain error escaped before post-check: %v", err)
				}
				order = append(order, "post")
				return nil
			})
			got, err := guardedRead(t.Context(), source, func(ctx context.Context, actual store.DBTX) (servingTestFacts, error) {
				if actual != db || ctx != t.Context() {
					t.Fatal("query did not use the bound context/database")
				}
				order = append(order, "query")
				return servingTestFacts{material: material}, domainErr
			}, discardServingTestFacts)
			order = append(order, "return")
			if !reflect.DeepEqual(order, []string{"pre", "query", "post", "return"}) {
				t.Fatalf("order = %v", order)
			}
			if err != domainErr {
				t.Fatalf("error = %v, want %v", err, domainErr)
			}
			if domainErr == nil {
				if string(got.material) != "private fixture" {
					t.Fatalf("facts = %q", got.material)
				}
			} else if got.material != nil || string(material) != string(make([]byte, len(material))) {
				t.Fatal("failed read retained private bytes")
			}
		})
	}
}

func TestGuardedReadErrorPriority(t *testing.T) {
	private := errors.New("private database details must not escape")
	for _, tc := range []struct {
		name                string
		queryErr, sourceErr error
		cancel              bool
		want                error
	}{
		{"not-found", authority.ErrNotFound, nil, false, authority.ErrNotFound},
		{"wrapped-not-found", fmt.Errorf("private: %w", authority.ErrNotFound), nil, false, authority.ErrNotFound},
		{"corrupt", authority.ErrConflict, nil, false, authority.ErrConflict},
		{"unknown-query-error", private, nil, false, authority.ErrAuthorityUnavailable},
		{"authority-over-joined-not-found", errors.Join(authority.ErrAuthorityUnavailable, authority.ErrNotFound), nil, false, authority.ErrAuthorityUnavailable},
		{"authority-over-joined-conflict", errors.Join(authority.ErrConflict, authority.ErrAuthorityUnavailable), nil, false, authority.ErrAuthorityUnavailable},
		{"cancellation-over-joined-authority", errors.Join(authority.ErrAuthorityUnavailable, authority.ErrCanceled, authority.ErrNotFound), nil, false, authority.ErrCanceled},
		{"postcheck-over-not-found", authority.ErrNotFound, private, false, authority.ErrAuthorityUnavailable},
		{"postcheck-over-corrupt", authority.ErrConflict, authority.ErrAuthorityUnavailable, false, authority.ErrAuthorityUnavailable},
		{"postcheck-after-success", nil, private, false, authority.ErrAuthorityUnavailable},
		{"source-domain-is-not-query-domain", nil, authority.ErrNotFound, false, authority.ErrAuthorityUnavailable},
		{"query-cancellation", context.Canceled, private, false, authority.ErrCanceled},
		{"query-deadline", context.DeadlineExceeded, private, false, authority.ErrCanceled},
		{"source-cancellation", authority.ErrNotFound, context.Canceled, false, authority.ErrCanceled},
		{"source-deadline", nil, context.DeadlineExceeded, false, authority.ErrCanceled},
		{"context-cancellation-wins", authority.ErrConflict, private, true, authority.ErrCanceled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			material := []byte("private fixture")
			post := false
			source := boundReadFunc(func(ctx context.Context, query func(context.Context, store.DBTX) error) error {
				_ = query(ctx, &servingTestDB{})
				post = true
				if tc.cancel {
					cancel()
				}
				return tc.sourceErr
			})
			got, err := guardedRead(ctx, source, func(context.Context, store.DBTX) (servingTestFacts, error) {
				return servingTestFacts{material}, tc.queryErr
			}, discardServingTestFacts)
			if !post || err != tc.want || got.material != nil {
				t.Fatalf("post=%v error=%v facts=%q; want %v and zero facts", post, err, got.material, tc.want)
			}
			for _, b := range material {
				if b != 0 {
					t.Fatal("failed read did not wipe temporary bytes")
				}
			}
		})
	}
}

func TestGuardedReadInvalidCapability(t *testing.T) {
	for _, mode := range []string{"nil-source", "typed-nil-source", "no-callback", "duplicate-callback", "nil-db", "typed-nil-db", "nil-callback-context", "canceled-callback-context", "precheck-failure", "nil-context", "already-canceled", "nil-query", "nil-discard"} {
		t.Run(mode, func(t *testing.T) {
			ctx := t.Context()
			want := authority.ErrAuthorityUnavailable
			var material []byte
			calls := 0
			var source BoundAuthorityReadSource = boundReadFunc(func(ctx context.Context, query func(context.Context, store.DBTX) error) error {
				switch mode {
				case "no-callback":
					return nil
				case "precheck-failure":
					return errors.New("private head mismatch")
				case "nil-db":
					return query(ctx, nil)
				case "typed-nil-db":
					var db *servingTestDB
					return query(ctx, db)
				case "nil-callback-context":
					return query(nil, &servingTestDB{})
				case "canceled-callback-context":
					child, cancel := context.WithCancel(ctx)
					cancel()
					return query(child, &servingTestDB{})
				}
				if err := query(ctx, &servingTestDB{}); err != nil {
					return err
				}
				if mode == "duplicate-callback" {
					_ = query(ctx, &servingTestDB{})
				}
				return nil
			})
			query := func(context.Context, store.DBTX) (servingTestFacts, error) {
				calls++
				material = []byte("private fixture")
				return servingTestFacts{material}, nil
			}
			discard := discardServingTestFacts
			switch mode {
			case "nil-source":
				source = nil
			case "typed-nil-source":
				source = boundReadFunc(nil)
			case "nil-query":
				query = nil
			case "nil-discard":
				discard = nil
			case "nil-context":
				ctx = nil
			case "already-canceled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
				want = authority.ErrCanceled
			case "canceled-callback-context":
				want = authority.ErrCanceled
			}
			got, err := guardedRead(ctx, source, query, discard)
			wantCalls := 0
			if mode == "duplicate-callback" {
				wantCalls = 1
			}
			if err != want || got.material != nil || calls != wantCalls {
				t.Fatalf("error=%v facts=%q calls=%d; want %v, zero facts, %d calls", err, got.material, calls, want, wantCalls)
			}
			for _, b := range material {
				if b != 0 {
					t.Fatal("invalid capability retained private bytes")
				}
			}
		})
	}
}
