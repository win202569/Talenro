package nats

import (
	"context"

	natsgo "github.com/nats-io/nats.go"
)

type Probe struct {
	Conn *natsgo.Conn
}

func (Probe) Name() string { return "nats" }

func (p Probe) Ping(ctx context.Context) error { return p.Conn.FlushWithContext(ctx) }
