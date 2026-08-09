// Package events validates shared event contracts before publication or consumption.
package events

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
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

// ValidatePayloadDescriptor rejects event payload schemas that can carry secret or routing material.
func ValidatePayloadDescriptor(descriptor protoreflect.MessageDescriptor) error {
	return validatePayloadDescriptor(descriptor, make(map[protoreflect.FullName]struct{}))
}

func validatePayloadDescriptor(descriptor protoreflect.MessageDescriptor, seen map[protoreflect.FullName]struct{}) error {
	if descriptor == nil {
		return fmt.Errorf("payload descriptor is nil")
	}
	if _, exists := seen[descriptor.FullName()]; exists {
		return nil
	}
	seen[descriptor.FullName()] = struct{}{}

	fields := descriptor.Fields()
	for index := range fields.Len() {
		field := fields.Get(index)
		for _, term := range strings.Split(strings.ToLower(string(field.Name())), "_") {
			if isForbiddenPayloadFieldTerm(term) {
				return fmt.Errorf("payload field %s contains forbidden term %s", field.FullName(), term)
			}
		}
		if field.Message() != nil {
			if err := validatePayloadDescriptor(field.Message(), seen); err != nil {
				return err
			}
		}
	}
	return nil
}

func isForbiddenPayloadFieldTerm(term string) bool {
	switch term {
	case "body", "ciphertext", "email", "emails", "error", "errors", "key", "keys", "locator", "nonce", "provider", "token", "uri", "url":
		return true
	default:
		return false
	}
}
