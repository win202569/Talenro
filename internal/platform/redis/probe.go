package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

type Probe struct {
	Client *goredis.Client
}

func (Probe) Name() string { return "redis" }

func (p Probe) Ping(ctx context.Context) error { return p.Client.Ping(ctx).Err() }
