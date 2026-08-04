package events

import (
	"fmt"
	"strings"

	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
)

const maxPayloadSize = 256 * 1024

// ValidateEnvelope checks the metadata required for safe at-least-once delivery.
// Consumers remain responsible for deduplicating by event ID or idempotency key.
func ValidateEnvelope(e *eventsv1.EventEnvelope) error {
	if e == nil {
		return fmt.Errorf("event envelope is nil")
	}
	if e.EventId == "" {
		return fmt.Errorf("event_id is required")
	}
	if !strings.HasPrefix(e.EventType, "talenro.") {
		return fmt.Errorf("event_type must use talenro namespace")
	}
	if e.OccurredAt == nil || !e.OccurredAt.IsValid() {
		return fmt.Errorf("occurred_at is invalid")
	}
	if e.Producer == "" {
		return fmt.Errorf("producer is required")
	}
	if e.AggregateType == "" || e.AggregateId == "" || e.AggregateVersion == 0 {
		return fmt.Errorf("aggregate identity is incomplete")
	}
	if e.IdempotencyKey == "" {
		return fmt.Errorf("idempotency_key is required")
	}
	if len(e.Payload) == 0 {
		return fmt.Errorf("payload is required")
	}
	if len(e.Payload) > maxPayloadSize {
		return fmt.Errorf("payload exceeds 256 KiB")
	}
	return nil
}
