// Package redis adapts a Redis client to the readiness probe contract.
package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

// Probe checks whether a Redis client is reachable.
type Probe struct {
	Client *goredis.Client
}

// Name returns the public readiness component name.
func (Probe) Name() string { return "redis" }

// Ping checks the client using the caller context.
func (p Probe) Ping(ctx context.Context) error { return p.Client.Ping(ctx).Err() }
