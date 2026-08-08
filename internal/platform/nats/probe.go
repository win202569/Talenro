// Package nats adapts a NATS connection to the readiness probe contract.
package nats

import (
	"context"

	natsgo "github.com/nats-io/nats.go"
)

// Probe checks whether a NATS connection can flush within a caller deadline.
type Probe struct {
	Conn *natsgo.Conn
}

// Name returns the public readiness component name.
func (Probe) Name() string { return "nats" }

// Ping flushes the connection using the caller context.
func (p Probe) Ping(ctx context.Context) error { return p.Conn.FlushWithContext(ctx) }
