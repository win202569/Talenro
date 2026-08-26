package contracts

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
	nodecontrolv1 "talenro.local/platform/gen/go/talenro/nodecontrol/v1"
	sharedevents "talenro.local/platform/internal/contracts/events"
)

const nodeControlGoldenPath = "testdata/node_control_event_v1.bin"

func TestNodeControlEventGoldenAndPrivate(t *testing.T) {
	event := validNodeControlEvents()[0]
	encoded, err := MarshalNodeControlEvent(event)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(nodeControlGoldenPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(want) != 179 {
		t.Fatalf("golden length = %d; want 179", len(want))
	}
	digest := sha256.Sum256(want)
	if got := hex.EncodeToString(digest[:]); got != "49a23a7d0398764cf88989aeb508662d6398deecfd6fc6fdf84c9ea98a69e9d6" {
		t.Fatalf("golden SHA-256 = %s", got)
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

func TestNodeControlGoldenVectorHasNoUpdateFlag(t *testing.T) {
	if flag.Lookup("update") != nil {
		t.Fatal("reviewed nodecontrol golden still has a mutable -update path")
	}
}

func TestCanonicalNodeControlUUIDVersionAndVariant(t *testing.T) {
	for _, value := range []string{
		"11111111-1111-1111-8111-111111111111",
		"11111111-1111-2111-8111-111111111111",
		"11111111-1111-3111-8111-111111111111",
		"11111111-1111-4111-8111-111111111111",
		"11111111-1111-5111-8111-111111111111",
	} {
		if !canonicalNodeControlUUID(value) {
			t.Fatalf("canonicalNodeControlUUID(%q) = false; want true", value)
		}
	}

	for _, value := range []string{
		"00000000-0000-0000-0000-000000000000",
		"11111111-1111-0111-8111-111111111111",
		"11111111-1111-6111-8111-111111111111",
		"11111111-1111-4111-7111-111111111111",
		"11111111-1111-4111-c111-111111111111",
		"11111111-1111-4111-8111-11111111111A",
		"11111111111141118111111111111111",
	} {
		if canonicalNodeControlUUID(value) {
			t.Fatalf("canonicalNodeControlUUID(%q) = true; want false", value)
		}
	}
}

func TestMarshalNodeControlEventRejectsTypedNilPayloadsWithoutPanic(t *testing.T) {
	cases := []struct {
		name       string
		eventIndex int
		mutate     func(*nodecontrolv1.NodeControlEventV1)
	}{
		{name: "inventory", eventIndex: 0, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.Payload = (*nodecontrolv1.NodeControlEventV1_NodeInventoryChanged)(nil)
		}},
		{name: "desired", eventIndex: 1, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.Payload = (*nodecontrolv1.NodeControlEventV1_NodeDesiredStatePublished)(nil)
		}},
		{name: "availability", eventIndex: 2, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.Payload = (*nodecontrolv1.NodeControlEventV1_NodeAvailabilityChanged)(nil)
		}},
		{name: "security", eventIndex: 3, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.Payload = (*nodecontrolv1.NodeControlEventV1_NodeSecurityStateChanged)(nil)
		}},
		{name: "certificate", eventIndex: 4, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.Payload = (*nodecontrolv1.NodeControlEventV1_NodeCertificateStatusChanged)(nil)
		}},
		{name: "operator", eventIndex: 5, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.Payload = (*nodecontrolv1.NodeControlEventV1_NodeOperatorActionRecorded)(nil)
		}},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			event := clonedNodeControlEvent(validNodeControlEvents()[test.eventIndex])
			test.mutate(event)
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("MarshalNodeControlEvent panicked: %v", recovered)
				}
			}()
			encoded, err := MarshalNodeControlEvent(event)
			if err == nil || encoded != nil {
				t.Fatalf("MarshalNodeControlEvent = %x, %v; want nil bytes and error", encoded, err)
			}
		})
	}
}

func TestUnmarshalNodeControlEventRejectsOversizedWire(t *testing.T) {
	const maximumNodeControlEventWireBytes = 264 * 1024
	if maxNodeControlEventWireBytes != maximumNodeControlEventWireBytes {
		t.Fatalf("maxNodeControlEventWireBytes = %d; want %d", maxNodeControlEventWireBytes, maximumNodeControlEventWireBytes)
	}
	oversized := make([]byte, maximumNodeControlEventWireBytes+1)
	copy(oversized, []byte{0xf8, 0x07, 0x01})
	if event, err := UnmarshalNodeControlEvent(oversized); err == nil || event != nil {
		t.Fatalf("UnmarshalNodeControlEvent oversized = %v, %v; want nil event and error", event, err)
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

func TestMarshalNodeControlEventRejectsNoncanonicalUUIDInEveryRole(t *testing.T) {
	const invalidUUID = "aaaaaaaa-aaaa-6aaa-8aaa-aaaaaaaaaaaa"
	cases := []struct {
		name       string
		eventIndex int
		mutate     func(*nodecontrolv1.NodeControlEventV1)
	}{
		{name: "event id", eventIndex: 0, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.EventId = invalidUUID
		}},
		{name: "inventory node and aggregate id", eventIndex: 0, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.AggregateId = invalidUUID
			event.GetNodeInventoryChanged().NodeId = invalidUUID
		}},
		{name: "desired node and aggregate id", eventIndex: 1, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.AggregateId = invalidUUID
			event.GetNodeDesiredStatePublished().NodeId = invalidUUID
		}},
		{name: "availability node and aggregate id", eventIndex: 2, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.AggregateId = invalidUUID
			event.GetNodeAvailabilityChanged().NodeId = invalidUUID
		}},
		{name: "availability boot id", eventIndex: 2, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeAvailabilityChanged().BootId = invalidUUID
		}},
		{name: "security node and aggregate id", eventIndex: 3, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.AggregateId = invalidUUID
			event.GetNodeSecurityStateChanged().NodeId = invalidUUID
		}},
		{name: "security incident id", eventIndex: 3, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeSecurityStateChanged().IncidentId = invalidUUID
		}},
		{name: "certificate node and aggregate id", eventIndex: 4, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.AggregateId = invalidUUID
			event.GetNodeCertificateStatusChanged().NodeId = invalidUUID
		}},
		{name: "certificate record id", eventIndex: 4, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeCertificateStatusChanged().CertificateRecordId = invalidUUID
		}},
		{name: "operator audit and aggregate id", eventIndex: 5, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.AggregateId = invalidUUID
			event.GetNodeOperatorActionRecorded().AuditId = invalidUUID
		}},
		{name: "operator id", eventIndex: 5, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeOperatorActionRecorded().OperatorId = invalidUUID
		}},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			event := clonedNodeControlEvent(validNodeControlEvents()[test.eventIndex])
			test.mutate(event)
			encoded, err := MarshalNodeControlEvent(event)
			if err == nil || encoded != nil {
				t.Fatalf("MarshalNodeControlEvent = %x, %v; want nil bytes and error", encoded, err)
			}
			if strings.Contains(err.Error(), invalidUUID) {
				t.Fatal("validation error disclosed the rejected UUID")
			}
		})
	}
}

func TestMarshalNodeControlEventRejectsEveryPayloadBindingMismatch(t *testing.T) {
	valid := validNodeControlEvents()
	for index := range valid {
		t.Run(valid[index].GetEventType(), func(t *testing.T) {
			wrongEventType := clonedNodeControlEvent(valid[index])
			wrongEventType.EventType = valid[(index+1)%len(valid)].GetEventType()
			assertNodeControlMarshalRejected(t, wrongEventType)

			wrongAggregateType := clonedNodeControlEvent(valid[index])
			if wrongAggregateType.GetAggregateType() == "node" {
				wrongAggregateType.AggregateType = "operator_action"
			} else {
				wrongAggregateType.AggregateType = "node"
			}
			assertNodeControlMarshalRejected(t, wrongAggregateType)

			wrongAggregateID := clonedNodeControlEvent(valid[index])
			wrongAggregateID.AggregateId = "88888888-8888-4888-8888-888888888888"
			assertNodeControlMarshalRejected(t, wrongAggregateID)

			if index < 3 {
				wrongAggregateVersion := clonedNodeControlEvent(valid[index])
				wrongAggregateVersion.AggregateVersion++
				assertNodeControlMarshalRejected(t, wrongAggregateVersion)
			}
		})
	}
}

func TestNodeControlPayloadVersionDigestAndPresenceBoundaries(t *testing.T) {
	type versionContract struct {
		name       string
		eventIndex int
		set        func(*nodecontrolv1.NodeControlEventV1, uint64)
	}
	versions := []versionContract{
		{name: "inventory version", eventIndex: 0, set: func(event *nodecontrolv1.NodeControlEventV1, value uint64) {
			event.AggregateVersion = value
			event.GetNodeInventoryChanged().InventoryVersion = value
		}},
		{name: "desired generation", eventIndex: 1, set: func(event *nodecontrolv1.NodeControlEventV1, value uint64) {
			event.AggregateVersion = value
			event.GetNodeDesiredStatePublished().Generation = value
		}},
		{name: "availability sequence", eventIndex: 2, set: func(event *nodecontrolv1.NodeControlEventV1, value uint64) {
			event.AggregateVersion = value
			event.GetNodeAvailabilityChanged().Sequence = value
		}},
	}
	for _, contract := range versions {
		t.Run(contract.name, func(t *testing.T) {
			for _, invalid := range []uint64{0, uint64(math.MaxInt64) + 1} {
				event := clonedNodeControlEvent(validNodeControlEvents()[contract.eventIndex])
				contract.set(event, invalid)
				assertNodeControlMarshalRejected(t, event)
			}
			event := clonedNodeControlEvent(validNodeControlEvents()[contract.eventIndex])
			contract.set(event, math.MaxInt64)
			assertNodeControlMarshalAccepted(t, event)
		})
	}

	for _, test := range []struct {
		name   string
		length int
	}{{name: "zero", length: 0}, {name: "short", length: 31}, {name: "long", length: 33}} {
		t.Run("digest length "+test.name, func(t *testing.T) {
			event := clonedNodeControlEvent(validNodeControlEvents()[1])
			event.GetNodeDesiredStatePublished().ContentDigest = make([]byte, test.length)
			assertNodeControlMarshalRejected(t, event)
		})
	}
	for _, value := range []bool{false, true} {
		event := clonedNodeControlEvent(validNodeControlEvents()[2])
		event.GetNodeAvailabilityChanged().AcceptingNew = boolPointer(value)
		assertNodeControlMarshalAccepted(t, event)
	}
	absent := clonedNodeControlEvent(validNodeControlEvents()[2])
	absent.GetNodeAvailabilityChanged().AcceptingNew = nil
	assertNodeControlMarshalRejected(t, absent)
}

func TestNodeControlEventRejectsUnknownFieldsAtEveryNestedLevel(t *testing.T) {
	for index, source := range validNodeControlEvents() {
		t.Run(source.GetEventType(), func(t *testing.T) {
			event := clonedNodeControlEvent(validNodeControlEvents()[index])
			payload := nodeControlPayloadMessage(t, event)
			payload.SetUnknown([]byte{0xf8, 0x07, 0x01})
			assertNodeControlMarshalRejected(t, event)

			raw, err := (proto.MarshalOptions{Deterministic: true}).Marshal(event)
			if err != nil {
				t.Fatal(err)
			}
			if decoded, err := UnmarshalNodeControlEvent(raw); err == nil || decoded != nil {
				t.Fatalf("UnmarshalNodeControlEvent nested unknown = %v, %v; want nil event and error", decoded, err)
			}
		})
	}

	event := clonedNodeControlEvent(validNodeControlEvents()[0])
	event.GetOccurredAt().ProtoReflect().SetUnknown([]byte{0xf8, 0x07, 0x01})
	assertNodeControlMarshalRejected(t, event)

	event = clonedNodeControlEvent(validNodeControlEvents()[1])
	event.GetNodeDesiredStatePublished().GetValidUntil().ProtoReflect().SetUnknown([]byte{0xf8, 0x07, 0x01})
	assertNodeControlMarshalRejected(t, event)
}

func TestMarshalNodeControlEventRejectsSemanticCovertChannels(t *testing.T) {
	valid := validNodeControlEvents()
	cases := []struct {
		name       string
		eventIndex int
		mutate     func(*nodecontrolv1.NodeControlEventV1)
	}{
		{name: "inventory pop code violates canonical grammar", eventIndex: 0, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeInventoryChanged().PopCode = "endpoint-canary-"
		}},
		{name: "inventory operator state is unregistered", eventIndex: 0, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeInventoryChanged().OperatorState = "serial-canary"
		}},
		{name: "desired reason is unregistered", eventIndex: 1, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeDesiredStatePublished().Reason = "csr-canary"
		}},
		{name: "availability health is unregistered", eventIndex: 2, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeAvailabilityChanged().Health = "endpoint-canary"
		}},
		{name: "availability reason is unregistered", eventIndex: 2, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeAvailabilityChanged().Reason = "serial-canary"
		}},
		{name: "security fault subtype is unregistered", eventIndex: 3, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeSecurityStateChanged().FaultSubtype = "csr-canary"
		}},
		{name: "security state is unregistered", eventIndex: 3, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeSecurityStateChanged().SecurityState = "endpoint-canary"
		}},
		{name: "security reason is unregistered", eventIndex: 3, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeSecurityStateChanged().Reason = "serial-canary"
		}},
		{name: "certificate status is unregistered", eventIndex: 4, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeCertificateStatusChanged().Status = "csr-canary"
		}},
		{name: "operator action is unregistered", eventIndex: 5, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeOperatorActionRecorded().Action = "endpoint-canary"
		}},
		{name: "operator target is unregistered", eventIndex: 5, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeOperatorActionRecorded().Target = "serial-canary"
		}},
		{name: "operator result is unregistered", eventIndex: 5, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeOperatorActionRecorded().Result = "csr-canary"
		}},
		{name: "operator reason is unregistered", eventIndex: 5, mutate: func(event *nodecontrolv1.NodeControlEventV1) {
			event.GetNodeOperatorActionRecorded().Reason = "endpoint-canary"
		}},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			event := clonedNodeControlEvent(valid[test.eventIndex])
			test.mutate(event)
			encoded, err := MarshalNodeControlEvent(event)
			if err == nil || encoded != nil {
				t.Fatal("MarshalNodeControlEvent accepted a semantic covert channel")
			}
			for _, canary := range []string{"endpoint-canary", "serial-canary", "csr-canary"} {
				if strings.Contains(err.Error(), canary) {
					t.Fatal("validation error disclosed a private canary")
				}
			}
		})
	}
}

func TestNodeControlValueRegistriesAreExact(t *testing.T) {
	for _, contract := range nodeControlRegistryContracts() {
		t.Run(contract.name, func(t *testing.T) {
			if len(contract.registry) != len(contract.values) {
				t.Fatalf("registry size = %d; want %d", len(contract.registry), len(contract.values))
			}
			want := make(map[string]struct{}, len(contract.values))
			for _, value := range contract.values {
				if _, duplicate := want[value]; duplicate {
					t.Fatalf("duplicate expected value %q", value)
				}
				want[value] = struct{}{}
				if !contract.registry.has(value) {
					t.Fatalf("missing registered value %q", value)
				}
			}
			for value := range contract.registry {
				if _, ok := want[value]; !ok {
					t.Fatalf("unexpected registered value %q", value)
				}
			}
		})
	}
}

func TestNodeControlValueRegistriesAcceptEveryValue(t *testing.T) {
	for _, contract := range nodeControlRegistryContracts() {
		t.Run(contract.name, func(t *testing.T) {
			for _, value := range contract.values {
				t.Run(value, func(t *testing.T) {
					event := clonedNodeControlEvent(validNodeControlEvents()[contract.eventIndex])
					contract.set(event, value)
					assertNodeControlMarshalAccepted(t, event)
				})
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

func TestUnmarshalNodeControlEventRejectsNoncanonicalWireForms(t *testing.T) {
	canonical, err := MarshalNodeControlEvent(validNodeControlEvents()[0])
	if err != nil {
		t.Fatal(err)
	}
	eventIDField := nodeControlWireField(t, canonical, 1)
	if !bytes.HasPrefix(canonical, eventIDField) {
		t.Fatal("canonical event_id is not the first wire field")
	}
	reordered := append(append([]byte(nil), canonical[len(eventIDField):]...), eventIDField...)

	versionField := nodeControlWireField(t, canonical, 5)
	if !bytes.Equal(versionField, []byte{0x28, 0x07}) {
		t.Fatalf("aggregate_version wire = %x; want 2807", versionField)
	}
	overlongVarint := replaceNodeControlWireField(t, canonical, versionField, []byte{0x28, 0x87, 0x00})

	inventoryOneof := nodeControlWireField(t, canonical, 10)
	sameOneofDuplicate := append(append([]byte(nil), canonical...), inventoryOneof...)
	desiredCanonical, err := MarshalNodeControlEvent(validNodeControlEvents()[1])
	if err != nil {
		t.Fatal(err)
	}
	differentOneofDuplicate := append(append([]byte(nil), canonical...), nodeControlWireField(t, desiredCanonical, 11)...)

	_, wireType, tagLength := protowire.ConsumeTag(inventoryOneof)
	if tagLength < 0 || wireType != protowire.BytesType {
		t.Fatalf("inventory oneof tag is malformed: %x", inventoryOneof)
	}
	inventoryPayload, payloadLength := protowire.ConsumeBytes(inventoryOneof[tagLength:])
	if payloadLength < 0 {
		t.Fatalf("inventory payload is malformed: %x", inventoryOneof)
	}
	nestedDuplicate := append(append([]byte(nil), inventoryPayload...), nodeControlWireField(t, inventoryPayload, 1)...)
	nestedReplacement := protowire.AppendTag(nil, 10, protowire.BytesType)
	nestedReplacement = protowire.AppendBytes(nestedReplacement, nestedDuplicate)
	nestedFieldDuplicate := replaceNodeControlWireField(t, canonical, inventoryOneof, nestedReplacement)

	invalidUTF8 := append([]byte(nil), canonical...)
	aggregateType := []byte{0x1a, 0x04, 'n', 'o', 'd', 'e'}
	position := bytes.Index(invalidUTF8, aggregateType)
	if position < 0 {
		t.Fatal("aggregate_type wire field not found")
	}
	invalidUTF8[position+2] = 0xff

	wrongWireType := append([]byte(nil), canonical...)
	wrongWireType[0] = 0x0d

	cases := []struct {
		name string
		wire []byte
	}{
		{name: "out of order", wire: reordered},
		{name: "overlong varint", wire: overlongVarint},
		{name: "same oneof duplicate", wire: sameOneofDuplicate},
		{name: "different oneof duplicate", wire: differentOneofDuplicate},
		{name: "nested duplicate", wire: nestedFieldDuplicate},
		{name: "invalid utf8", wire: invalidUTF8},
		{name: "wrong wire type", wire: wrongWireType},
		{name: "truncated", wire: append([]byte(nil), canonical[:len(canonical)-1]...)},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			decoded, err := UnmarshalNodeControlEvent(test.wire)
			if err == nil || decoded != nil {
				t.Fatalf("UnmarshalNodeControlEvent = %v, %v; want nil event and error", decoded, err)
			}
		})
	}
}

func FuzzUnmarshalNodeControlEventCanonical(f *testing.F) {
	for _, event := range validNodeControlEvents() {
		encoded, err := MarshalNodeControlEvent(event)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(encoded)
	}
	f.Add([]byte(nil))
	f.Add([]byte{0xf8, 0x07, 0x01})
	f.Fuzz(func(t *testing.T, encoded []byte) {
		decoded, err := UnmarshalNodeControlEvent(encoded)
		if err != nil {
			return
		}
		canonical, err := MarshalNodeControlEvent(decoded)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(canonical, encoded) {
			t.Fatal("accepted nodecontrol wire did not round-trip byte-identically")
		}
	})
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

func TestNodeControlEventFileDescriptorIsExact(t *testing.T) {
	file := nodecontrolv1.File_talenro_nodecontrol_v1_events_proto
	if file.Path() != "talenro/nodecontrol/v1/events.proto" || file.Package() != "talenro.nodecontrol.v1" || file.Syntax() != protoreflect.Proto3 {
		t.Fatalf("file identity = %q, %q, %v", file.Path(), file.Package(), file.Syntax())
	}
	if file.Imports().Len() != 1 || file.Imports().Get(0).Path() != "google/protobuf/timestamp.proto" {
		t.Fatalf("file imports = %v; want only google/protobuf/timestamp.proto", file.Imports())
	}
	if file.Enums().Len() != 0 || file.Extensions().Len() != 0 || file.Services().Len() != 0 {
		t.Fatal("nodecontrol event file gained an enum, extension, or service")
	}

	wantMessages := nodeControlMessageContracts()
	if file.Messages().Len() != len(wantMessages) {
		t.Fatalf("message count = %d; want %d", file.Messages().Len(), len(wantMessages))
	}
	for index, want := range wantMessages {
		message := file.Messages().Get(index)
		if message.Name() != want.name || message.FullName() != protoreflect.FullName("talenro.nodecontrol.v1."+string(want.name)) {
			t.Fatalf("message[%d] = %q", index, message.FullName())
		}
		if message.Fields().Len() != len(want.fields) {
			t.Fatalf("message %s field count = %d; want %d", message.FullName(), message.Fields().Len(), len(want.fields))
		}
		if message.Oneofs().Len() != len(want.oneofs) {
			t.Fatalf("message %s oneof count = %d; want %d", message.FullName(), message.Oneofs().Len(), len(want.oneofs))
		}
		for oneofIndex, wantOneof := range want.oneofs {
			oneof := message.Oneofs().Get(oneofIndex)
			if oneof.Name() != wantOneof.name || oneof.IsSynthetic() != wantOneof.synthetic {
				t.Fatalf("message %s oneof[%d] = %q synthetic=%v", message.FullName(), oneofIndex, oneof.Name(), oneof.IsSynthetic())
			}
		}
		for fieldIndex, wantField := range want.fields {
			field := message.Fields().Get(fieldIndex)
			if field.Name() != wantField.name || field.Number() != wantField.number || field.Kind() != wantField.kind ||
				field.Cardinality() != protoreflect.Optional || field.HasPresence() != wantField.presence ||
				field.HasOptionalKeyword() != wantField.optionalKeyword {
				t.Fatalf("field %s = number %d kind %v cardinality %v presence=%v optional=%v", field.FullName(), field.Number(), field.Kind(), field.Cardinality(), field.HasPresence(), field.HasOptionalKeyword())
			}
			oneofName := protoreflect.Name("")
			if field.ContainingOneof() != nil {
				oneofName = field.ContainingOneof().Name()
			}
			if oneofName != wantField.oneof {
				t.Fatalf("field %s oneof = %q; want %q", field.FullName(), oneofName, wantField.oneof)
			}
			messageName := protoreflect.FullName("")
			if field.Message() != nil {
				messageName = field.Message().FullName()
			}
			if messageName != wantField.message {
				t.Fatalf("field %s message = %q; want %q", field.FullName(), messageName, wantField.message)
			}
			if field.IsMap() || field.IsList() {
				t.Fatalf("field %s unexpectedly became a map or list", field.FullName())
			}
		}
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
			Payload: &nodecontrolv1.NodeControlEventV1_NodeDesiredStatePublished{NodeDesiredStatePublished: &nodecontrolv1.NodeDesiredStatePublishedV1{NodeId: nodeID, Generation: 8, ContentDigest: bytes.Repeat([]byte{0x01}, 32), ValidUntil: validUntil, Reason: "operator_update"}},
		},
		{
			EventId: "11111111-1111-4111-8111-111111111113", OccurredAt: occurredAt, AggregateType: "node", AggregateId: nodeID, AggregateVersion: 9, EventType: "node_availability_changed.v1",
			Payload: &nodecontrolv1.NodeControlEventV1_NodeAvailabilityChanged{NodeAvailabilityChanged: &nodecontrolv1.NodeAvailabilityChangedV1{NodeId: nodeID, BootId: "33333333-3333-4333-8333-333333333333", Sequence: 9, Health: "healthy", AcceptingNew: boolPointer(false), Reason: "observation"}},
		},
		{
			EventId: "11111111-1111-4111-8111-111111111114", OccurredAt: occurredAt, AggregateType: "node", AggregateId: nodeID, AggregateVersion: 10, EventType: "node_security_state_changed.v1",
			Payload: &nodecontrolv1.NodeControlEventV1_NodeSecurityStateChanged{NodeSecurityStateChanged: &nodecontrolv1.NodeSecurityStateChangedV1{NodeId: nodeID, IncidentId: "44444444-4444-4444-8444-444444444444", FaultSubtype: "identity_compromise", SecurityState: "quarantined", Reason: "fault_confirmed"}},
		},
		{
			EventId: "11111111-1111-4111-8111-111111111115", OccurredAt: occurredAt, AggregateType: "node", AggregateId: nodeID, AggregateVersion: 11, EventType: "node_certificate_status_changed.v1",
			Payload: &nodecontrolv1.NodeControlEventV1_NodeCertificateStatusChanged{NodeCertificateStatusChanged: &nodecontrolv1.NodeCertificateStatusChangedV1{NodeId: nodeID, CertificateRecordId: "55555555-5555-4555-8555-555555555555", Status: "active"}},
		},
		{
			EventId: "11111111-1111-4111-8111-111111111116", OccurredAt: occurredAt, AggregateType: "operator_action", AggregateId: "66666666-6666-4666-8666-666666666666", AggregateVersion: 12, EventType: "node_operator_action_recorded.v1",
			Payload: &nodecontrolv1.NodeControlEventV1_NodeOperatorActionRecorded{NodeOperatorActionRecorded: &nodecontrolv1.NodeOperatorActionRecordedV1{AuditId: "66666666-6666-4666-8666-666666666666", OperatorId: "77777777-7777-4777-8777-777777777777", Action: "put_node_desired_state", Target: "node", Result: "accepted", Reason: "operator_update"}},
		},
	}
}

func boolPointer(value bool) *bool {
	return &value
}

func clonedNodeControlEvent(event *nodecontrolv1.NodeControlEventV1) *nodecontrolv1.NodeControlEventV1 {
	return proto.Clone(event).(*nodecontrolv1.NodeControlEventV1)
}

func assertNodeControlMarshalRejected(t *testing.T, event *nodecontrolv1.NodeControlEventV1) {
	t.Helper()
	encoded, err := MarshalNodeControlEvent(event)
	if err == nil || encoded != nil {
		t.Fatalf("MarshalNodeControlEvent = %x, %v; want nil bytes and error", encoded, err)
	}
}

func assertNodeControlMarshalAccepted(t *testing.T, event *nodecontrolv1.NodeControlEventV1) {
	t.Helper()
	encoded, err := MarshalNodeControlEvent(event)
	if err != nil || len(encoded) == 0 {
		t.Fatalf("MarshalNodeControlEvent = %x, %v; want canonical bytes", encoded, err)
	}
}

func nodeControlPayloadMessage(t *testing.T, event *nodecontrolv1.NodeControlEventV1) protoreflect.Message {
	t.Helper()
	message := event.ProtoReflect()
	oneof := message.Descriptor().Oneofs().ByName("payload")
	field := message.WhichOneof(oneof)
	if field == nil {
		t.Fatal("event payload is absent")
	}
	return message.Get(field).Message()
}

func nodeControlWireField(t *testing.T, encoded []byte, wanted protowire.Number) []byte {
	t.Helper()
	remaining := encoded
	for len(remaining) > 0 {
		number, wireType, tagLength := protowire.ConsumeTag(remaining)
		if tagLength < 0 {
			t.Fatalf("malformed protobuf tag: %x", remaining)
		}
		valueLength := protowire.ConsumeFieldValue(number, wireType, remaining[tagLength:])
		if valueLength < 0 {
			t.Fatalf("malformed protobuf field %d: %x", number, remaining)
		}
		fieldLength := tagLength + valueLength
		if number == wanted {
			return append([]byte(nil), remaining[:fieldLength]...)
		}
		remaining = remaining[fieldLength:]
	}
	t.Fatalf("protobuf field %d not found", wanted)
	return nil
}

func replaceNodeControlWireField(t *testing.T, encoded, oldField, newField []byte) []byte {
	t.Helper()
	position := bytes.Index(encoded, oldField)
	if position < 0 {
		t.Fatalf("protobuf field %x not found", oldField)
	}
	replaced := make([]byte, 0, len(encoded)-len(oldField)+len(newField))
	replaced = append(replaced, encoded[:position]...)
	replaced = append(replaced, newField...)
	replaced = append(replaced, encoded[position+len(oldField):]...)
	return replaced
}

type nodeControlRegistryContract struct {
	name       string
	registry   nodeControlValueRegistry
	values     []string
	eventIndex int
	set        func(*nodecontrolv1.NodeControlEventV1, string)
}

func nodeControlRegistryContracts() []nodeControlRegistryContract {
	return []nodeControlRegistryContract{
		{
			name: "operator states", registry: nodeControlOperatorStates,
			values: []string{"provisioning", "enabled", "draining", "disabled"}, eventIndex: 0,
			set: func(event *nodecontrolv1.NodeControlEventV1, value string) {
				event.GetNodeInventoryChanged().OperatorState = value
			},
		},
		{
			name: "desired reasons", registry: nodeControlDesiredReasons,
			values: []string{"initial", "operator_update", "drain", "resume", "lease_refresh", "clear_slot_quarantine", "restore_reauthorize"}, eventIndex: 1,
			set: func(event *nodecontrolv1.NodeControlEventV1, value string) {
				event.GetNodeDesiredStatePublished().Reason = value
			},
		},
		{
			name: "health states", registry: nodeControlHealthStates,
			values: []string{"unknown", "healthy", "degraded", "offline", "quarantined"}, eventIndex: 2,
			set: func(event *nodecontrolv1.NodeControlEventV1, value string) {
				event.GetNodeAvailabilityChanged().Health = value
			},
		},
		{
			name: "availability reasons", registry: nodeControlAvailabilityReasons,
			values: []string{"observation", "observation_timeout", "boot_changed", "capacity_blocked", "profile_mismatch", "agent_reducer_mismatch"}, eventIndex: 2,
			set: func(event *nodecontrolv1.NodeControlEventV1, value string) {
				event.GetNodeAvailabilityChanged().Reason = value
			},
		},
		{
			name: "fault subtypes", registry: nodeControlFaultSubtypes,
			values: []string{
				"identity_compromise", "online_signer_equivocation", "metadata_rollback", "root_rollback", "root_equivocation",
				"unverified_client_highwater_conflict", "client_highwater_ahead", "server_trust_bundle_conflict",
				"trusted_time_rollback_or_unavailable", "local_state_corruption_or_rollback", "release_or_process_integrity",
				"profile_binding_mismatch", "incident_overflow",
			}, eventIndex: 3,
			set: func(event *nodecontrolv1.NodeControlEventV1, value string) {
				event.GetNodeSecurityStateChanged().FaultSubtype = value
			},
		},
		{
			name: "security states", registry: nodeControlSecurityStates,
			values: []string{"normal", "quarantined"}, eventIndex: 3,
			set: func(event *nodecontrolv1.NodeControlEventV1, value string) {
				event.GetNodeSecurityStateChanged().SecurityState = value
			},
		},
		{
			name: "security reasons", registry: nodeControlSecurityReasons,
			values: []string{"fault_confirmed", "incident_capacity_exceeded", "host_remediation", "incident_resolved"}, eventIndex: 3,
			set: func(event *nodecontrolv1.NodeControlEventV1, value string) {
				event.GetNodeSecurityStateChanged().Reason = value
			},
		},
		{
			name: "certificate statuses", registry: nodeControlCertificateStatuses,
			values: []string{"active", "recovery_pending", "recovery_limited", "revoked"}, eventIndex: 4,
			set: func(event *nodecontrolv1.NodeControlEventV1, value string) {
				event.GetNodeCertificateStatusChanged().Status = value
			},
		},
		{
			name: "operator actions", registry: nodeControlOperatorActions,
			values: []string{
				"create_node_pop", "update_node_pop", "create_node_failure_domain", "update_node_failure_domain", "create_node",
				"update_node", "create_node_endpoint", "update_node_endpoint", "create_node_process_slot", "update_node_process_slot",
				"create_node_enrollment_grant", "put_node_desired_state", "drain_node", "disable_node", "reenroll_node",
				"complete_node_reenrollment", "register_node_host_security_incident", "register_node_resource_envelope",
				"clear_node_security_quarantine", "resume_node_after_security", "reauthorize_node_after_restore",
			}, eventIndex: 5,
			set: func(event *nodecontrolv1.NodeControlEventV1, value string) {
				event.GetNodeOperatorActionRecorded().Action = value
			},
		},
		{
			name: "operator targets", registry: nodeControlOperatorTargets,
			values: []string{
				"node_pop", "node_failure_domain", "node", "node_endpoint", "node_process_slot", "node_enrollment_grant",
				"node_desired_state", "node_security_incident", "node_resource_envelope", "node_recovery_session",
			}, eventIndex: 5,
			set: func(event *nodecontrolv1.NodeControlEventV1, value string) {
				event.GetNodeOperatorActionRecorded().Target = value
			},
		},
		{
			name: "operator results", registry: nodeControlOperatorResults,
			values: []string{"pending", "active", "accepted", "completed", "superseded", "rejected"}, eventIndex: 5,
			set: func(event *nodecontrolv1.NodeControlEventV1, value string) {
				event.GetNodeOperatorActionRecorded().Result = value
			},
		},
		{
			name: "operator reasons", registry: nodeControlOperatorReasons,
			values: []string{
				"provision", "inventory_update", "capacity_change", "drain_maintenance", "administrative_disable", "retire",
				"identity_compromise", "host_remediation", "security_recovery", "authority_restore", "release_update", "initial",
				"operator_update", "drain", "resume", "lease_refresh", "clear_slot_quarantine", "restore_reauthorize",
			}, eventIndex: 5,
			set: func(event *nodecontrolv1.NodeControlEventV1, value string) {
				event.GetNodeOperatorActionRecorded().Reason = value
			},
		},
	}
}

type nodeControlOneofContract struct {
	name      protoreflect.Name
	synthetic bool
}

type nodeControlFieldContract struct {
	name            protoreflect.Name
	number          protoreflect.FieldNumber
	kind            protoreflect.Kind
	presence        bool
	optionalKeyword bool
	oneof           protoreflect.Name
	message         protoreflect.FullName
}

type nodeControlMessageContract struct {
	name   protoreflect.Name
	oneofs []nodeControlOneofContract
	fields []nodeControlFieldContract
}

func nodeControlMessageContracts() []nodeControlMessageContract {
	field := func(name string, number protoreflect.FieldNumber, kind protoreflect.Kind, presence bool, optionalKeyword bool, oneof string, message string) nodeControlFieldContract {
		return nodeControlFieldContract{
			name: protoreflect.Name(name), number: number, kind: kind, presence: presence, optionalKeyword: optionalKeyword,
			oneof: protoreflect.Name(oneof), message: protoreflect.FullName(message),
		}
	}
	return []nodeControlMessageContract{
		{
			name: "NodeControlEventV1", oneofs: []nodeControlOneofContract{{name: "payload"}},
			fields: []nodeControlFieldContract{
				field("event_id", 1, protoreflect.StringKind, false, false, "", ""),
				field("occurred_at", 2, protoreflect.MessageKind, true, false, "", "google.protobuf.Timestamp"),
				field("aggregate_type", 3, protoreflect.StringKind, false, false, "", ""),
				field("aggregate_id", 4, protoreflect.StringKind, false, false, "", ""),
				field("aggregate_version", 5, protoreflect.Uint64Kind, false, false, "", ""),
				field("event_type", 6, protoreflect.StringKind, false, false, "", ""),
				field("node_inventory_changed", 10, protoreflect.MessageKind, true, false, "payload", "talenro.nodecontrol.v1.NodeInventoryChangedV1"),
				field("node_desired_state_published", 11, protoreflect.MessageKind, true, false, "payload", "talenro.nodecontrol.v1.NodeDesiredStatePublishedV1"),
				field("node_availability_changed", 12, protoreflect.MessageKind, true, false, "payload", "talenro.nodecontrol.v1.NodeAvailabilityChangedV1"),
				field("node_security_state_changed", 13, protoreflect.MessageKind, true, false, "payload", "talenro.nodecontrol.v1.NodeSecurityStateChangedV1"),
				field("node_certificate_status_changed", 14, protoreflect.MessageKind, true, false, "payload", "talenro.nodecontrol.v1.NodeCertificateStatusChangedV1"),
				field("node_operator_action_recorded", 15, protoreflect.MessageKind, true, false, "payload", "talenro.nodecontrol.v1.NodeOperatorActionRecordedV1"),
			},
		},
		{
			name: "NodeInventoryChangedV1",
			fields: []nodeControlFieldContract{
				field("node_id", 1, protoreflect.StringKind, false, false, "", ""),
				field("inventory_version", 2, protoreflect.Uint64Kind, false, false, "", ""),
				field("pop_code", 3, protoreflect.StringKind, false, false, "", ""),
				field("operator_state", 4, protoreflect.StringKind, false, false, "", ""),
			},
		},
		{
			name: "NodeDesiredStatePublishedV1",
			fields: []nodeControlFieldContract{
				field("node_id", 1, protoreflect.StringKind, false, false, "", ""),
				field("generation", 2, protoreflect.Uint64Kind, false, false, "", ""),
				field("content_digest", 3, protoreflect.BytesKind, false, false, "", ""),
				field("valid_until", 4, protoreflect.MessageKind, true, false, "", "google.protobuf.Timestamp"),
				field("reason", 5, protoreflect.StringKind, false, false, "", ""),
			},
		},
		{
			name: "NodeAvailabilityChangedV1", oneofs: []nodeControlOneofContract{{name: "_accepting_new", synthetic: true}},
			fields: []nodeControlFieldContract{
				field("node_id", 1, protoreflect.StringKind, false, false, "", ""),
				field("boot_id", 2, protoreflect.StringKind, false, false, "", ""),
				field("sequence", 3, protoreflect.Uint64Kind, false, false, "", ""),
				field("health", 4, protoreflect.StringKind, false, false, "", ""),
				field("accepting_new", 5, protoreflect.BoolKind, true, true, "_accepting_new", ""),
				field("reason", 6, protoreflect.StringKind, false, false, "", ""),
			},
		},
		{
			name: "NodeSecurityStateChangedV1",
			fields: []nodeControlFieldContract{
				field("node_id", 1, protoreflect.StringKind, false, false, "", ""),
				field("incident_id", 2, protoreflect.StringKind, false, false, "", ""),
				field("fault_subtype", 3, protoreflect.StringKind, false, false, "", ""),
				field("security_state", 4, protoreflect.StringKind, false, false, "", ""),
				field("reason", 5, protoreflect.StringKind, false, false, "", ""),
			},
		},
		{
			name: "NodeCertificateStatusChangedV1",
			fields: []nodeControlFieldContract{
				field("node_id", 1, protoreflect.StringKind, false, false, "", ""),
				field("certificate_record_id", 2, protoreflect.StringKind, false, false, "", ""),
				field("status", 3, protoreflect.StringKind, false, false, "", ""),
			},
		},
		{
			name: "NodeOperatorActionRecordedV1",
			fields: []nodeControlFieldContract{
				field("audit_id", 1, protoreflect.StringKind, false, false, "", ""),
				field("operator_id", 2, protoreflect.StringKind, false, false, "", ""),
				field("action", 3, protoreflect.StringKind, false, false, "", ""),
				field("target", 4, protoreflect.StringKind, false, false, "", ""),
				field("result", 5, protoreflect.StringKind, false, false, "", ""),
				field("reason", 6, protoreflect.StringKind, false, false, "", ""),
			},
		},
	}
}
