// Package postgres adapts a PostgreSQL pool to the readiness probe contract.
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Probe checks whether a PostgreSQL pool is reachable.
type Probe struct {
	Pool *pgxpool.Pool
}

// Name returns the public readiness component name.
func (Probe) Name() string { return "postgres" }

// Ping checks the pool using the caller context.
func (p Probe) Ping(ctx context.Context) error { return p.Pool.Ping(ctx) }
