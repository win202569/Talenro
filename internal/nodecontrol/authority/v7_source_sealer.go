package authority

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"talenro.local/platform/internal/store"
)

const authorityV7SourceSealRollbackTimeout = 5 * time.Second

// AuthorityV7SourceSealWork runs only after the source-freeze exclusive
// transaction lock has been acquired. Implementations must derive every local
// source projection through dbtx/queries and must not use observations captured
// before Run.
type AuthorityV7SourceSealWork func(
	context.Context,
	store.DBTX,
	*store.Queries,
) error

type authorityV7SourceSealTx interface {
	store.DBTX
	Commit(context.Context) error
	Rollback(context.Context) error
}

type authorityV7SourceSealBegin func(
	context.Context,
	pgx.TxOptions,
) (authorityV7SourceSealTx, error)

// AuthorityV7SourceSealer owns the transaction boundary for a source seal.
// The generated acquire query is always the first post-BEGIN statement before
// trusted source-scanning work is invoked.
type AuthorityV7SourceSealer struct {
	begin authorityV7SourceSealBegin
}

func NewAuthorityV7SourceSealer(database *pgxpool.Pool) (*AuthorityV7SourceSealer, error) {
	if database == nil {
		return nil, ErrInvalidArgument
	}
	return &AuthorityV7SourceSealer{
		begin: func(ctx context.Context, options pgx.TxOptions) (authorityV7SourceSealTx, error) {
			return database.BeginTx(ctx, options)
		},
	}, nil
}

func (sealer *AuthorityV7SourceSealer) Run(
	ctx context.Context,
	work AuthorityV7SourceSealWork,
) (resultErr error) {
	if ctx == nil || sealer == nil || sealer.begin == nil || work == nil {
		return ErrInvalidArgument
	}

	tx, err := sealer.begin(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return fmt.Errorf("begin authority v7 source seal: %w", err)
	}

	committed := false
	defer func() {
		if committed {
			return
		}
		rollbackContext, cancel := context.WithTimeout(
			context.WithoutCancel(ctx),
			authorityV7SourceSealRollbackTimeout,
		)
		defer cancel()
		rollbackErr := tx.Rollback(rollbackContext)
		if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			resultErr = errors.Join(
				resultErr,
				fmt.Errorf("rollback authority v7 source seal: %w", rollbackErr),
			)
		}
	}()

	queries := store.New(tx)
	if err := queries.AcquireAuthorityV7SourceFreezeForSeal(ctx); err != nil {
		return fmt.Errorf("acquire authority v7 source freeze: %w", err)
	}
	if err := work(ctx, tx, queries); err != nil {
		return fmt.Errorf("run authority v7 source seal: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit authority v7 source seal: %w", err)
	}
	committed = true
	return nil
}
