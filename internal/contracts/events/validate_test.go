package events

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
)

func TestValidateEnvelopeRejectsInvalidEnvelopes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*eventsv1.EventEnvelope) *eventsv1.EventEnvelope
	}{
		{
			name: "nil envelope",
			mutate: func(*eventsv1.EventEnvelope) *eventsv1.EventEnvelope {
				return nil
			},
		},
		{
			name: "missing event id",
			mutate: func(e *eventsv1.EventEnvelope) *eventsv1.EventEnvelope {
				e.EventId = ""
				return e
			},
		},
		{
			name: "event type outside namespace",
			mutate: func(e *eventsv1.EventEnvelope) *eventsv1.EventEnvelope {
				e.EventType = "system.ready.v1"
				return e
			},
		},
		{
			name: "missing occurrence time",
			mutate: func(e *eventsv1.EventEnvelope) *eventsv1.EventEnvelope {
				e.OccurredAt = nil
				return e
			},
		},
		{
			name: "invalid occurrence time",
			mutate: func(e *eventsv1.EventEnvelope) *eventsv1.EventEnvelope {
				e.OccurredAt = &timestamppb.Timestamp{Seconds: 253402300800}
				return e
			},
		},
		{
			name: "missing producer",
			mutate: func(e *eventsv1.EventEnvelope) *eventsv1.EventEnvelope {
				e.Producer = ""
				return e
			},
		},
		{
			name: "missing aggregate type",
			mutate: func(e *eventsv1.EventEnvelope) *eventsv1.EventEnvelope {
				e.AggregateType = ""
				return e
			},
		},
		{
			name: "missing aggregate id",
			mutate: func(e *eventsv1.EventEnvelope) *eventsv1.EventEnvelope {
				e.AggregateId = ""
				return e
			},
		},
		{
			name: "zero aggregate version",
			mutate: func(e *eventsv1.EventEnvelope) *eventsv1.EventEnvelope {
				e.AggregateVersion = 0
				return e
			},
		},
		{
			name: "missing idempotency key",
			mutate: func(e *eventsv1.EventEnvelope) *eventsv1.EventEnvelope {
				e.IdempotencyKey = ""
				return e
			},
		},
		{
			name: "missing payload",
			mutate: func(e *eventsv1.EventEnvelope) *eventsv1.EventEnvelope {
				e.Payload = nil
				return e
			},
		},
		{
			name: "payload over 256 KiB",
			mutate: func(e *eventsv1.EventEnvelope) *eventsv1.EventEnvelope {
				e.Payload = make([]byte, 256*1024+1)
				return e
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := ValidateEnvelope(tt.mutate(validEnvelope())); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestEnvelopeRoundTrip(t *testing.T) {
	t.Parallel()

	want := validEnvelope()
	encoded, err := proto.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}

	got := new(eventsv1.EventEnvelope)
	if err := proto.Unmarshal(encoded, got); err != nil {
		t.Fatal(err)
	}
	if err := ValidateEnvelope(got); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(want, got) {
		t.Fatalf("round trip mismatch: got %v want %v", got, want)
	}
}

func TestEnvelopeSchemaExcludesSensitiveFields(t *testing.T) {
	t.Parallel()

	forbidden := []string{
		"credential",
		"destination",
		"dns",
		"domain",
		"email",
		"error",
		"ip_address",
		"password",
		"traffic",
	}
	fields := (&eventsv1.EventEnvelope{}).ProtoReflect().Descriptor().Fields()
	for i := range fields.Len() {
		name := string(fields.Get(i).Name())
		for _, fragment := range forbidden {
			if strings.Contains(name, fragment) {
				t.Errorf("sensitive field %q contains forbidden fragment %q", name, fragment)
			}
		}
	}
}

func validEnvelope() *eventsv1.EventEnvelope {
	return &eventsv1.EventEnvelope{
		EventId:          "01JTESTEVENT000000000000002",
		EventType:        "talenro.system.ready.v1",
		OccurredAt:       timestamppb.Now(),
		Producer:         "control-api",
		AggregateType:    "system",
		AggregateId:      "foundation",
		AggregateVersion: 1,
		IdempotencyKey:   "system:foundation:1",
		TraceParent:      "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		Payload:          []byte(`{"ready":true}`),
	}
}
