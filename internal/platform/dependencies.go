package platform

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"talenro.local/platform/internal/config"
)

type Dependencies struct {
	Postgres *pgxpool.Pool
	Redis    *redis.Client
	NATS     *nats.Conn

	closeOnce     sync.Once
	closePostgres func()
	closeRedis    func()
	closeNATS     func()
}

func Open(ctx context.Context, cfg config.Config) (_ *Dependencies, err error) {
	deps := new(Dependencies)
	defer func() {
		if err != nil {
			deps.Close()
		}
	}()

	checkCtx, cancel := context.WithTimeout(ctx, cfg.DependencyTimeout)
	defer cancel()

	deps.Postgres, err = pgxpool.New(checkCtx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	deps.closePostgres = deps.Postgres.Close
	if err = deps.Postgres.Ping(checkCtx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	deps.Redis = redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddress,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})
	deps.closeRedis = func() { _ = deps.Redis.Close() }
	if err = deps.Redis.Ping(checkCtx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	deps.NATS, err = nats.Connect(cfg.NATSURL,
		nats.Name("talenro-control-api"),
		nats.Timeout(2*time.Second),
		nats.ReconnectWait(500*time.Millisecond),
		nats.MaxReconnects(10),
		nats.DrainTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}
	deps.closeNATS = func() {
		_ = deps.NATS.Drain()
		deps.NATS.Close()
	}

	return deps, nil
}

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
