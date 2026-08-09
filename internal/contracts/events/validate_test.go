package events

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	deviceauthv1 "talenro.local/platform/gen/go/talenro/deviceauth/v1"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	identityv1 "talenro.local/platform/gen/go/talenro/identity/v1"
	trustv1 "talenro.local/platform/gen/go/talenro/trust/v1"
)

type expectedField struct {
	name        protoreflect.Name
	number      protoreflect.FieldNumber
	cardinality protoreflect.Cardinality
	kind        protoreflect.Kind
	messageType protoreflect.FullName
}

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

func TestEnvelopeSchemaMatchesApprovedPrivacySurface(t *testing.T) {
	t.Parallel()

	expected := []expectedField{
		{name: "event_id", number: 1, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
		{name: "event_type", number: 2, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
		{name: "occurred_at", number: 3, cardinality: protoreflect.Optional, kind: protoreflect.MessageKind, messageType: "google.protobuf.Timestamp"},
		{name: "producer", number: 4, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
		{name: "aggregate_type", number: 5, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
		{name: "aggregate_id", number: 6, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
		{name: "aggregate_version", number: 7, cardinality: protoreflect.Optional, kind: protoreflect.Uint64Kind},
		{name: "idempotency_key", number: 8, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
		{name: "trace_parent", number: 9, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
		{name: "payload", number: 10, cardinality: protoreflect.Optional, kind: protoreflect.BytesKind},
	}

	assertMessageFields(t, (&eventsv1.EventEnvelope{}).ProtoReflect().Descriptor(), expected)
}

func TestC11EventSchemasMatchApprovedPrivacySurface(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		message  proto.Message
		expected []expectedField
	}{
		{
			name:    "account state changed",
			message: &identityv1.AccountStateChanged{},
			expected: []expectedField{
				{name: "principal_id", number: 1, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
				{name: "state", number: 2, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
				{name: "version", number: 3, cardinality: protoreflect.Optional, kind: protoreflect.Uint64Kind},
			},
		},
		{
			name:    "email delivery requested",
			message: &identityv1.EmailDeliveryRequested{},
			expected: []expectedField{
				{name: "delivery_id", number: 1, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
				{name: "principal_id", number: 2, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
				{name: "template_id", number: 3, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
				{name: "locale", number: 4, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
			},
		},
		{
			name:    "device authorization changed",
			message: &deviceauthv1.DeviceAuthorizationChanged{},
			expected: []expectedField{
				{name: "principal_id", number: 1, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
				{name: "device_id", number: 2, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
				{name: "authorization_id", number: 3, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
				{name: "state", number: 4, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
				{name: "version", number: 5, cardinality: protoreflect.Optional, kind: protoreflect.Uint64Kind},
			},
		},
		{
			name:    "device token family compromised",
			message: &deviceauthv1.DeviceTokenFamilyCompromised{},
			expected: []expectedField{
				{name: "authorization_id", number: 1, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
				{name: "family_id", number: 2, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
				{name: "version", number: 3, cardinality: protoreflect.Optional, kind: protoreflect.Uint64Kind},
			},
		},
		{
			name:    "bundle issued",
			message: &trustv1.BundleIssued{},
			expected: []expectedField{
				{name: "authorization_id", number: 1, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
				{name: "bundle_id", number: 2, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
				{name: "bundle_version", number: 3, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
				{name: "envelope_sha256", number: 4, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
			},
		},
		{
			name:    "bundle acknowledged",
			message: &trustv1.BundleAcknowledged{},
			expected: []expectedField{
				{name: "authorization_id", number: 1, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
				{name: "bundle_id", number: 2, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
				{name: "bundle_version", number: 3, cardinality: protoreflect.Optional, kind: protoreflect.StringKind},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			descriptor := test.message.ProtoReflect().Descriptor()
			assertMessageFields(t, descriptor, test.expected)
			if err := ValidatePayloadDescriptor(descriptor); err != nil {
				t.Fatalf("privacy validation: %v", err)
			}
		})
	}
}

func TestValidatePayloadDescriptorRejectsSensitiveFields(t *testing.T) {
	t.Parallel()

	for _, fieldName := range []string{
		"email", "refresh_token", "request_nonce", "public_key", "private_key", "bundle_locator",
		"callback_url", "ciphertext", "provider_body", "raw_error", "emailaddress", "callbackuri", "rawerror",
	} {
		t.Run(fieldName, func(t *testing.T) {
			t.Parallel()
			descriptor := descriptorWithStringField(t, fieldName)
			if err := ValidatePayloadDescriptor(descriptor); err == nil {
				t.Fatalf("expected %q to be rejected", fieldName)
			}
		})
	}
}

func TestValidatePayloadDescriptorAllowsUnrelatedCompoundNames(t *testing.T) {
	t.Parallel()

	for _, fieldName := range []string{"monkey", "allocator", "somebody", "security", "premailer"} {
		t.Run(fieldName, func(t *testing.T) {
			t.Parallel()
			descriptor := descriptorWithStringField(t, fieldName)
			if err := ValidatePayloadDescriptor(descriptor); err != nil {
				t.Fatalf("expected unrelated field %q to be allowed: %v", fieldName, err)
			}
		})
	}
}

func assertMessageFields(t *testing.T, descriptor protoreflect.MessageDescriptor, expected []expectedField) {
	t.Helper()

	fields := descriptor.Fields()
	if fields.Len() != len(expected) {
		t.Fatalf("field count = %d, want %d approved fields", fields.Len(), len(expected))
	}

	for i, want := range expected {
		field := fields.Get(i)
		var messageType protoreflect.FullName
		if field.Message() != nil {
			messageType = field.Message().FullName()
		}

		if field.Name() != want.name ||
			field.Number() != want.number ||
			field.Cardinality() != want.cardinality ||
			field.Kind() != want.kind ||
			messageType != want.messageType {
			t.Errorf(
				"field %d = {name:%q number:%d cardinality:%s kind:%s message:%q}, want {name:%q number:%d cardinality:%s kind:%s message:%q}",
				i,
				field.Name(),
				field.Number(),
				field.Cardinality(),
				field.Kind(),
				messageType,
				want.name,
				want.number,
				want.cardinality,
				want.kind,
				want.messageType,
			)
		}
	}
}

func descriptorWithStringField(t *testing.T, fieldName string) protoreflect.MessageDescriptor {
	t.Helper()

	file, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Syntax:  proto.String("proto3"),
		Name:    proto.String("privacy_test.proto"),
		Package: proto.String("talenro.privacytest.v1"),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("SensitivePayload"),
			Field: []*descriptorpb.FieldDescriptorProto{{
				Name:   proto.String(fieldName),
				Number: proto.Int32(1),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
			}},
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return file.Messages().Get(0)
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
