package outbox

import (
	"context"
	"reflect"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	// EventStreamName is the exact finite JetStream that persists outbox events.
	EventStreamName       = "TALENRO_EVENTS"
	minimumPublishTimeout = 100 * time.Millisecond
	maximumPublishTimeout = 10 * time.Second
)

// JetStreamPublisher is the narrow synchronous persistence-acknowledgement
// surface used by the broker adapter.
type JetStreamPublisher interface {
	PublishMsg(*nats.Msg, ...nats.PubOpt) (*nats.PubAck, error)
}

// NATSBroker publishes validated envelopes with stable NATS duplicate IDs.
type NATSBroker struct {
	publisher JetStreamPublisher
	timeout   time.Duration
}

var _ Broker = (*NATSBroker)(nil)

// NewNATSBroker binds a live JetStream publisher without exposing provider errors.
func NewNATSBroker(publisher JetStreamPublisher, timeout time.Duration) (*NATSBroker, error) {
	if publisher == nil || timeout < minimumPublishTimeout || timeout > maximumPublishTimeout {
		return nil, ErrPublisherConfiguration
	}
	value := reflect.ValueOf(publisher)
	if value.Kind() == reflect.Pointer && value.IsNil() {
		return nil, ErrPublisherConfiguration
	}
	return &NATSBroker{publisher: publisher, timeout: timeout}, nil
}

// Publish sends one immutable envelope and requires a bounded durable PubAck.
func (broker *NATSBroker) Publish(ctx context.Context, subject, messageID string, payload []byte) error {
	if broker == nil || broker.publisher == nil || ctx == nil || ctx.Err() != nil || messageID == "" || len(payload) == 0 {
		return ErrPublish
	}
	if registered, ok := SubjectForEventType(subject); !ok || registered != subject {
		return ErrPublish
	}
	message := nats.NewMsg(subject)
	message.Header.Set(nats.MsgIdHdr, messageID)
	message.Data = append([]byte(nil), payload...)
	publishCtx, cancel := context.WithTimeout(ctx, broker.timeout)
	defer cancel()
	ack, failed := natsProviderCall(func() (*nats.PubAck, error) {
		return broker.publisher.PublishMsg(message, nats.Context(publishCtx))
	})
	if failed || ack == nil || ack.Stream != EventStreamName || ack.Sequence == 0 {
		return ErrPublish
	}
	return nil
}

func natsProviderCall(call func() (*nats.PubAck, error)) (ack *nats.PubAck, failed bool) {
	defer func() {
		if recover() != nil {
			ack = nil
			failed = true
		}
	}()
	ack, err := call()
	return ack, err != nil
}
