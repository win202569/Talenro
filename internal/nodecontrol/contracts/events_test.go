package contracts

import (
	"bytes"
	"flag"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
	nodecontrolv1 "talenro.local/platform/gen/go/talenro/nodecontrol/v1"
	sharedevents "talenro.local/platform/internal/contracts/events"
)

const nodeControlGoldenPath = "testdata/node_control_event_v1.bin"

var update = flag.Bool("update", false, "update nodecontrol testdata")

func TestNodeControlEventGoldenAndPrivate(t *testing.T) {
	event := validNodeControlEvents()[0]
	encoded, err := MarshalNodeControlEvent(event)
	if err != nil {
		t.Fatal(err)
	}
	if *update {
		if err := os.MkdirAll(filepath.Dir(nodeControlGoldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(nodeControlGoldenPath, encoded, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(nodeControlGoldenPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, want) {
		t.Fatal("nodecontrol event wire drift")
	}
	for _, canary := range [][]byte{[]byte("serial-canary"), []byte("csr-canary"), []byte("endpoint-canary")} {
		if bytes.Contains(encoded, canary) {
			t.Fatal("private canary reached event bytes")
		}
	}
}

func TestNodeControlEventRoundTripSupportsAllPayloads(t *testing.T) {
	for _, event := range validNodeControlEvents() {
		t.Run(event.GetEventType(), func(t *testing.T) {
			encoded, err := MarshalNodeControlEvent(event)
			if err != nil {
				t.Fatal(err)
			}
			again, err := MarshalNodeControlEvent(event)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(encoded, again) {
				t.Fatal("deterministic marshal changed bytes")
			}
			decoded, err := UnmarshalNodeControlEvent(encoded)
			if err != nil {
				t.Fatal(err)
			}
			if !proto.Equal(event, decoded) {
				t.Fatalf("round trip = %v; want %v", decoded, event)
			}
			roundTrip, err := MarshalNodeControlEvent(decoded)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(encoded, roundTrip) {
				t.Fatal("marshal after unmarshal changed bytes")
			}
		})
	}
}

func TestMarshalNodeControlEventRejectsInvalidInput(t *testing.T) {
	valid := validNodeControlEvents()
	cases := []struct {
		name   string
		mutate func(*nodecontrolv1.NodeControlEventV1)
	}{
		{name: "nil", mutate: func(*nodecontrolv1.NodeControlEventV1) {}},
		{name: "unregistered event type", mutate: func(event *nodecontrolv1.NodeControlEventV1) { event.EventType = "node_unregistered.v1" }},
		{name: "missing payload", mutate: func(event *nodecontrolv1.NodeControlEventV1) { event.Payload = nil }},
		{name: "event payload mismatch", mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.Payload = clonedNodeControlEvent(valid[1]).Payload
		}},
		{name: "aggregate type mismatch", mutate: func(event *nodecontrolv1.NodeControlEventV1) { event.AggregateType = "certificate" }},
		{name: "aggregate id mismatch", mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.AggregateId = "33333333-3333-4333-8333-333333333333"
		}},
		{name: "noncanonical event id", mutate: func(event *nodecontrolv1.NodeControlEventV1) { event.EventId = "11111111-1111-4111-8111-11111111111A" }},
		{name: "noncanonical aggregate id", mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.AggregateId = "22222222-2222-4222-8222-22222222222A"
		}},
		{name: "invalid occurred at", mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.OccurredAt = &timestamppb.Timestamp{Nanos: 1_000_000_000}
		}},
		{name: "zero aggregate version", mutate: func(event *nodecontrolv1.NodeControlEventV1) { event.AggregateVersion = 0 }},
		{name: "aggregate version above postgres bigint", mutate: func(event *nodecontrolv1.NodeControlEventV1) { event.AggregateVersion = math.MaxInt64 + 1 }},
		{name: "zero inventory version", mutate: func(event *nodecontrolv1.NodeControlEventV1) { event.GetNodeInventoryChanged().InventoryVersion = 0 }},
		{name: "inventory node id is not canonical", mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeInventoryChanged().NodeId = "22222222-2222-4222-8222-22222222222A"
		}},
		{name: "unsafe inventory state", mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeInventoryChanged().OperatorState = "enabled now"
		}},
		{name: "desired digest has wrong length", mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.Payload = clonedNodeControlEvent(valid[1]).Payload
			event.EventType = valid[1].EventType
			event.AggregateVersion = valid[1].AggregateVersion
			event.GetNodeDesiredStatePublished().ContentDigest = make([]byte, 31)
		}},
		{name: "desired valid until is invalid", mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.Payload = clonedNodeControlEvent(valid[1]).Payload
			event.EventType = valid[1].EventType
			event.AggregateVersion = valid[1].AggregateVersion
			event.GetNodeDesiredStatePublished().ValidUntil = &timestamppb.Timestamp{Nanos: 1_000_000_000}
		}},
		{name: "availability accepting new is absent", mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.Payload = clonedNodeControlEvent(valid[2]).Payload
			event.EventType = valid[2].EventType
			event.AggregateVersion = valid[2].AggregateVersion
			event.GetNodeAvailabilityChanged().AcceptingNew = nil
		}},
		{name: "security incident id is not canonical", mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.Payload = clonedNodeControlEvent(valid[3]).Payload
			event.EventType = valid[3].EventType
			event.AggregateVersion = valid[3].AggregateVersion
			event.GetNodeSecurityStateChanged().IncidentId = "not-a-uuid"
		}},
		{name: "certificate record id is not canonical", mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.Payload = clonedNodeControlEvent(valid[4]).Payload
			event.EventType = valid[4].EventType
			event.AggregateVersion = valid[4].AggregateVersion
			event.GetNodeCertificateStatusChanged().CertificateRecordId = "not-a-uuid"
		}},
		{name: "operator id is not canonical", mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.Payload = clonedNodeControlEvent(valid[5]).Payload
			event.EventType = valid[5].EventType
			event.AggregateType = valid[5].AggregateType
			event.AggregateId = valid[5].AggregateId
			event.AggregateVersion = valid[5].AggregateVersion
			event.GetNodeOperatorActionRecorded().OperatorId = "not-a-uuid"
		}},
		{name: "unknown nested field", mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeInventoryChanged().ProtoReflect().SetUnknown([]byte{0xf8, 0x07, 0x01})
		}},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if test.name == "nil" {
				if _, err := MarshalNodeControlEvent(nil); err == nil {
					t.Fatal("MarshalNodeControlEvent(nil) succeeded")
				}
				return
			}
			event := clonedNodeControlEvent(valid[0])
			test.mutate(event)
			if _, err := MarshalNodeControlEvent(event); err == nil {
				t.Fatal("MarshalNodeControlEvent succeeded")
			}
		})
	}
}

func TestUnmarshalNodeControlEventRejectsUnknownAndNoncanonicalBytes(t *testing.T) {
	encoded, err := MarshalNodeControlEvent(validNodeControlEvents()[0])
	if err != nil {
		t.Fatal(err)
	}

	unknown := append(append([]byte(nil), encoded...), 0xf8, 0x07, 0x01)
	if _, err := UnmarshalNodeControlEvent(unknown); err == nil {
		t.Fatal("UnmarshalNodeControlEvent accepted an unknown envelope field")
	}

	noncanonical := append([]byte{0x0a, 0x24}, []byte("11111111-1111-4111-8111-111111111111")...)
	noncanonical = append(noncanonical, encoded...)
	if _, err := UnmarshalNodeControlEvent(noncanonical); err == nil {
		t.Fatal("UnmarshalNodeControlEvent accepted duplicate noncanonical fields")
	}

	if _, err := UnmarshalNodeControlEvent(nil); err == nil {
		t.Fatal("UnmarshalNodeControlEvent accepted empty bytes")
	}
}

func TestNodeControlSubjectAllowlist(t *testing.T) {
	cases := []struct {
		eventType string
		want      string
	}{
		{eventType: "node_inventory_changed.v1", want: "talenro.nodecontrol.v1.node_inventory_changed.v1"},
		{eventType: "node_desired_state_published.v1", want: "talenro.nodecontrol.v1.node_desired_state_published.v1"},
		{eventType: "node_availability_changed.v1", want: "talenro.nodecontrol.v1.node_availability_changed.v1"},
		{eventType: "node_security_state_changed.v1", want: "talenro.nodecontrol.v1.node_security_state_changed.v1"},
		{eventType: "node_certificate_status_changed.v1", want: "talenro.nodecontrol.v1.node_certificate_status_changed.v1"},
		{eventType: "node_operator_action_recorded.v1", want: "talenro.nodecontrol.v1.node_operator_action_recorded.v1"},
	}
	for _, test := range cases {
		t.Run(test.eventType, func(t *testing.T) {
			got, err := NodeControlSubject(test.eventType)
			if err != nil || got != test.want {
				t.Fatalf("NodeControlSubject(%q) = %q, %v; want %q, nil", test.eventType, got, err, test.want)
			}
		})
	}
	for _, eventType := range []string{"", "node_inventory_changed.v2", "talenro.nodecontrol.v1.node_inventory_changed.v1"} {
		if _, err := NodeControlSubject(eventType); err == nil {
			t.Fatalf("NodeControlSubject(%q) succeeded", eventType)
		}
	}
}

func TestNodeControlEventPayloadDescriptorsAreBoundedAndPrivate(t *testing.T) {
	descriptor := (&nodecontrolv1.NodeControlEventV1{}).ProtoReflect().Descriptor()
	payload := descriptor.Oneofs().ByName("payload")
	if payload == nil {
		t.Fatal("payload oneof is missing")
	}
	if payload.Fields().Len() != 6 {
		t.Fatalf("payload oneof has %d fields; want 6", payload.Fields().Len())
	}
	want := map[protoreflect.Name]protoreflect.FieldNumber{
		"node_inventory_changed":          10,
		"node_desired_state_published":    11,
		"node_availability_changed":       12,
		"node_security_state_changed":     13,
		"node_certificate_status_changed": 14,
		"node_operator_action_recorded":   15,
	}
	seenNumbers := make(map[protoreflect.FieldNumber]struct{}, payload.Fields().Len())
	for index := range payload.Fields().Len() {
		field := payload.Fields().Get(index)
		if number, ok := want[field.Name()]; !ok || field.Number() != number {
			t.Fatalf("payload field %s = %d; want registered field number", field.FullName(), field.Number())
		}
		if _, exists := seenNumbers[field.Number()]; exists {
			t.Fatalf("duplicate payload field number %d", field.Number())
		}
		seenNumbers[field.Number()] = struct{}{}
		if err := sharedevents.ValidatePayloadDescriptor(field.Message()); err != nil {
			t.Fatalf("payload %s is not safe: %v", field.FullName(), err)
		}
		assertNoPrivateDescriptorFields(t, field.Message())
	}
}

func assertNoPrivateDescriptorFields(t *testing.T, descriptor protoreflect.MessageDescriptor) {
	t.Helper()
	for index := range descriptor.Fields().Len() {
		field := descriptor.Fields().Get(index)
		name := strings.ToLower(string(field.Name()))
		for _, forbidden := range []string{"endpoint", "grant", "serial", "der", "csr", "material", "credential", "signed", "output", "capacity", "label"} {
			if strings.Contains(name, forbidden) {
				t.Fatalf("payload field %s contains private term %q", field.FullName(), forbidden)
			}
		}
		if field.Message() != nil {
			assertNoPrivateDescriptorFields(t, field.Message())
		}
	}
}

func validNodeControlEvents() []*nodecontrolv1.NodeControlEventV1 {
	occurredAt := timestamppb.New(time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC))
	validUntil := timestamppb.New(time.Date(2026, 8, 23, 13, 0, 0, 0, time.UTC))
	nodeID := "22222222-2222-4222-8222-222222222222"
	return []*nodecontrolv1.NodeControlEventV1{
		{
			EventId: "11111111-1111-4111-8111-111111111111", OccurredAt: occurredAt, AggregateType: "node", AggregateId: nodeID, AggregateVersion: 7, EventType: "node_inventory_changed.v1",
			Payload: &nodecontrolv1.NodeControlEventV1_NodeInventoryChanged{NodeInventoryChanged: &nodecontrolv1.NodeInventoryChangedV1{NodeId: nodeID, InventoryVersion: 7, PopCode: "us-east", OperatorState: "enabled"}},
		},
		{
			EventId: "11111111-1111-4111-8111-111111111112", OccurredAt: occurredAt, AggregateType: "node", AggregateId: nodeID, AggregateVersion: 8, EventType: "node_desired_state_published.v1",
			Payload: &nodecontrolv1.NodeControlEventV1_NodeDesiredStatePublished{NodeDesiredStatePublished: &nodecontrolv1.NodeDesiredStatePublishedV1{NodeId: nodeID, Generation: 8, ContentDigest: bytes.Repeat([]byte{0x01}, 32), ValidUntil: validUntil, Reason: "operator_request"}},
		},
		{
			EventId: "11111111-1111-4111-8111-111111111113", OccurredAt: occurredAt, AggregateType: "node", AggregateId: nodeID, AggregateVersion: 9, EventType: "node_availability_changed.v1",
			Payload: &nodecontrolv1.NodeControlEventV1_NodeAvailabilityChanged{NodeAvailabilityChanged: &nodecontrolv1.NodeAvailabilityChangedV1{NodeId: nodeID, BootId: "33333333-3333-4333-8333-333333333333", Sequence: 9, Health: "healthy", AcceptingNew: boolPointer(false), Reason: "observed"}},
		},
		{
			EventId: "11111111-1111-4111-8111-111111111114", OccurredAt: occurredAt, AggregateType: "node", AggregateId: nodeID, AggregateVersion: 10, EventType: "node_security_state_changed.v1",
			Payload: &nodecontrolv1.NodeControlEventV1_NodeSecurityStateChanged{NodeSecurityStateChanged: &nodecontrolv1.NodeSecurityStateChangedV1{NodeId: nodeID, IncidentId: "44444444-4444-4444-8444-444444444444", FaultSubtype: "attestation_failed", SecurityState: "stopped", Reason: "fault_confirmed"}},
		},
		{
			EventId: "11111111-1111-4111-8111-111111111115", OccurredAt: occurredAt, AggregateType: "node", AggregateId: nodeID, AggregateVersion: 11, EventType: "node_certificate_status_changed.v1",
			Payload: &nodecontrolv1.NodeControlEventV1_NodeCertificateStatusChanged{NodeCertificateStatusChanged: &nodecontrolv1.NodeCertificateStatusChangedV1{NodeId: nodeID, CertificateRecordId: "55555555-5555-4555-8555-555555555555", Status: "active"}},
		},
		{
			EventId: "11111111-1111-4111-8111-111111111116", OccurredAt: occurredAt, AggregateType: "operator_action", AggregateId: "66666666-6666-4666-8666-666666666666", AggregateVersion: 12, EventType: "node_operator_action_recorded.v1",
			Payload: &nodecontrolv1.NodeControlEventV1_NodeOperatorActionRecorded{NodeOperatorActionRecorded: &nodecontrolv1.NodeOperatorActionRecordedV1{AuditId: "66666666-6666-4666-8666-666666666666", OperatorId: "77777777-7777-4777-8777-777777777777", Action: "publish_desired_state", Target: "node", Result: "accepted", Reason: "operator_request"}},
		},
	}
}

func boolPointer(value bool) *bool {
	return &value
}

func clonedNodeControlEvent(event *nodecontrolv1.NodeControlEventV1) *nodecontrolv1.NodeControlEventV1 {
	return proto.Clone(event).(*nodecontrolv1.NodeControlEventV1)
}
