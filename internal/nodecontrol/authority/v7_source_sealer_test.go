package authority

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"talenro.local/platform/internal/store"
)

func TestAuthorityV7SourceSealerOrdersAcquireBeforeWorkAndCommit(t *testing.T) {
	tx := &task8SourceSealFakeTx{}
	sealer := &AuthorityV7SourceSealer{
		begin: func(_ context.Context, options pgx.TxOptions) (authorityV7SourceSealTx, error) {
			tx.events = append(tx.events, "begin")
			tx.options = options
			return tx, nil
		},
	}
	workRan := false
	err := sealer.Run(context.Background(), func(_ context.Context, dbtx store.DBTX, queries *store.Queries) error {
		workRan = true
		if dbtx != tx || queries == nil {
			t.Fatalf("source-seal callback received dbtx=%T queries=%v", dbtx, queries != nil)
		}
		tx.events = append(tx.events, "work")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !workRan {
		t.Fatal("source-seal callback did not run")
	}
	if tx.options.IsoLevel != pgx.ReadCommitted || tx.options.AccessMode != pgx.ReadWrite {
		t.Fatalf("source-seal options=%+v, want read-committed/read-write", tx.options)
	}
	if want := []string{"begin", "acquire", "work", "commit"}; !reflect.DeepEqual(tx.events, want) {
		t.Fatalf("source-seal events=%v, want %v", tx.events, want)
	}
}

func TestAuthorityV7SourceSealerRollsBackBeforeCallingWorkWhenAcquireFails(t *testing.T) {
	acquireErr := errors.New("acquire failed")
	tx := &task8SourceSealFakeTx{execErr: acquireErr}
	sealer := &AuthorityV7SourceSealer{
		begin: func(_ context.Context, options pgx.TxOptions) (authorityV7SourceSealTx, error) {
			tx.events = append(tx.events, "begin")
			tx.options = options
			return tx, nil
		},
	}
	workRan := false
	err := sealer.Run(context.Background(), func(context.Context, store.DBTX, *store.Queries) error {
		workRan = true
		return nil
	})
	if !errors.Is(err, acquireErr) {
		t.Fatalf("source-seal acquire error=%v, want %v", err, acquireErr)
	}
	if workRan {
		t.Fatal("source-seal callback ran before a successful acquire")
	}
	if want := []string{"begin", "acquire", "rollback"}; !reflect.DeepEqual(tx.events, want) {
		t.Fatalf("source-seal events=%v, want %v", tx.events, want)
	}
}

func TestAuthorityV7SourceSealerRollsBackCallbackAndCommitFailures(t *testing.T) {
	workErr := errors.New("work failed")
	commitErr := errors.New("commit failed")
	for _, testCase := range []struct {
		name      string
		workErr   error
		commitErr error
		want      []string
	}{
		{name: "callback", workErr: workErr, want: []string{"begin", "acquire", "work", "rollback"}},
		{name: "commit", commitErr: commitErr, want: []string{"begin", "acquire", "work", "commit", "rollback"}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			tx := &task8SourceSealFakeTx{commitErr: testCase.commitErr}
			sealer := &AuthorityV7SourceSealer{
				begin: func(_ context.Context, _ pgx.TxOptions) (authorityV7SourceSealTx, error) {
					tx.events = append(tx.events, "begin")
					return tx, nil
				},
			}
			err := sealer.Run(context.Background(), func(context.Context, store.DBTX, *store.Queries) error {
				tx.events = append(tx.events, "work")
				return testCase.workErr
			})
			wantErr := testCase.workErr
			if wantErr == nil {
				wantErr = testCase.commitErr
			}
			if !errors.Is(err, wantErr) {
				t.Fatalf("source-seal error=%v, want %v", err, wantErr)
			}
			if !reflect.DeepEqual(tx.events, testCase.want) {
				t.Fatalf("source-seal events=%v, want %v", tx.events, testCase.want)
			}
		})
	}
}

type task8SourceSealFakeTx struct {
	events    []string
	options   pgx.TxOptions
	execErr   error
	commitErr error
	closed    bool
}

func (tx *task8SourceSealFakeTx) Exec(_ context.Context, sql string, _ ...interface{}) (pgconn.CommandTag, error) {
	const want = "-- name: acquireauthorityv7sourcefreezeforseal :exec\nselect nodecontrol.v7_acquire_source_freeze_for_seal()"
	if strings.TrimSpace(strings.ToLower(strings.ReplaceAll(sql, "\r\n", "\n"))) != want {
		return pgconn.CommandTag{}, errors.New("unexpected source-seal SQL")
	}
	tx.events = append(tx.events, "acquire")
	return pgconn.CommandTag{}, tx.execErr
}

func (*task8SourceSealFakeTx) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}

func (*task8SourceSealFakeTx) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	return nil
}

func (tx *task8SourceSealFakeTx) Commit(context.Context) error {
	tx.events = append(tx.events, "commit")
	if tx.commitErr == nil {
		tx.closed = true
	}
	return tx.commitErr
}

func (tx *task8SourceSealFakeTx) Rollback(context.Context) error {
	if tx.closed {
		return pgx.ErrTxClosed
	}
	tx.events = append(tx.events, "rollback")
	tx.closed = true
	return nil
}
