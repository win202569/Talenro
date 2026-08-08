// Package platform owns control API infrastructure lifecycles and safe error categories.
package platform

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"talenro.local/platform/internal/config"
)

type dependencyOperations struct {
	newPostgres   func(context.Context, string) (*pgxpool.Pool, error)
	pingPostgres  func(context.Context, *pgxpool.Pool) error
	closePostgres func(*pgxpool.Pool)
	newRedis      func(*redis.Options) *redis.Client
	pingRedis     func(context.Context, *redis.Client) error
	closeRedis    func(*redis.Client) error
	connectNATS   func(string, ...nats.Option) (*nats.Conn, error)
	drainNATS     func(*nats.Conn) error
	closeNATS     func(*nats.Conn)
}

// Dependencies contains the live PostgreSQL, Redis, and NATS clients.
type Dependencies struct {
	Postgres *pgxpool.Pool
	Redis    *redis.Client
	NATS     *nats.Conn

	closeOnce     sync.Once
	closePostgres func()
	closeRedis    func()
	closeNATS     func()
}

// Open connects and checks all required infrastructure dependencies.
func Open(ctx context.Context, cfg config.Config) (_ *Dependencies, err error) {
	return openWithOperations(ctx, cfg, dependencyOperations{
		newPostgres:   pgxpool.New,
		pingPostgres:  func(ctx context.Context, pool *pgxpool.Pool) error { return pool.Ping(ctx) },
		closePostgres: func(pool *pgxpool.Pool) { pool.Close() },
		newRedis:      redis.NewClient,
		pingRedis:     func(ctx context.Context, client *redis.Client) error { return client.Ping(ctx).Err() },
		closeRedis:    func(client *redis.Client) error { return client.Close() },
		connectNATS:   nats.Connect,
		drainNATS:     func(conn *nats.Conn) error { return conn.Drain() },
		closeNATS:     func(conn *nats.Conn) { conn.Close() },
	})
}

func openWithOperations(
	ctx context.Context,
	cfg config.Config,
	operations dependencyOperations,
) (_ *Dependencies, err error) {
	deps := new(Dependencies)
	defer func() {
		if err != nil {
			deps.Close()
		}
	}()

	checkCtx, cancel := context.WithTimeout(ctx, cfg.DependencyTimeout)
	defer cancel()

	deps.Postgres, err = operations.newPostgres(checkCtx, cfg.DatabaseURL)
	if err != nil {
		return nil, NewCategorizedError(CategoryDependencies, err)
	}
	deps.closePostgres = func() { operations.closePostgres(deps.Postgres) }
	if err = operations.pingPostgres(checkCtx, deps.Postgres); err != nil {
		return nil, NewCategorizedError(CategoryDependencies, err)
	}

	deps.Redis = operations.newRedis(&redis.Options{
		Addr:         cfg.RedisAddress,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})
	deps.closeRedis = func() { _ = operations.closeRedis(deps.Redis) }
	if err = operations.pingRedis(checkCtx, deps.Redis); err != nil {
		return nil, NewCategorizedError(CategoryDependencies, err)
	}

	deps.NATS, err = operations.connectNATS(cfg.NATSURL,
		nats.Name("talenro-control-api"),
		nats.Timeout(2*time.Second),
		nats.ReconnectWait(500*time.Millisecond),
		nats.MaxReconnects(10),
		nats.DrainTimeout(5*time.Second),
	)
	if err != nil {
		return nil, NewCategorizedError(CategoryDependencies, err)
	}
	deps.closeNATS = func() {
		_ = operations.drainNATS(deps.NATS)
		operations.closeNATS(deps.NATS)
	}

	return deps, nil
}

// Close releases dependencies once in reverse acquisition order.
func (d *Dependencies) Close() {
	if d == nil {
		return
	}

	d.closeOnce.Do(func() {
		if d.closeNATS != nil {
			d.closeNATS()
		} else if d.NATS != nil {
			_ = d.NATS.Drain()
			d.NATS.Close()
		}
		if d.closeRedis != nil {
			d.closeRedis()
		} else if d.Redis != nil {
			_ = d.Redis.Close()
		}
		if d.closePostgres != nil {
			d.closePostgres()
		} else if d.Postgres != nil {
			d.Postgres.Close()
		}
	})
}
