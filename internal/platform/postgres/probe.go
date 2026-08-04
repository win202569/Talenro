package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Probe struct {
	Pool *pgxpool.Pool
}

func (Probe) Name() string { return "postgres" }

func (p Probe) Ping(ctx context.Context) error { return p.Pool.Ping(ctx) }
