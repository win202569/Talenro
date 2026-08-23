package contracts

import (
	"bytes"
	"errors"
	"math"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	nodecontrolv1 "talenro.local/platform/gen/go/talenro/nodecontrol/v1"
	sharedevents "talenro.local/platform/internal/contracts/events"
)

const (
	nodeInventoryChangedType         = "node_inventory_changed.v1"
	nodeDesiredStatePublishedType    = "node_desired_state_published.v1"
	nodeAvailabilityChangedType      = "node_availability_changed.v1"
	nodeSecurityStateChangedType     = "node_security_state_changed.v1"
	nodeCertificateStatusChangedType = "node_certificate_status_changed.v1"
	nodeOperatorActionRecordedType   = "node_operator_action_recorded.v1"
	nodeControlSubjectPrefix         = "talenro.nodecontrol.v1."
)

var (
	errInvalidNodeControlEvent   = errors.New("nodecontrol contracts: invalid node control event")
	errInvalidNodeControlSubject = errors.New("nodecontrol contracts: invalid node control subject")
)

// MarshalNodeControlEvent validates a registered node-control event and returns canonical protobuf bytes.
func MarshalNodeControlEvent(event *nodecontrolv1.NodeControlEventV1) ([]byte, error) {
	if event == nil || nodeControlHasUnknownFields(event.ProtoReflect()) || validateNodeControlEvent(event) != nil {
		return nil, errInvalidNodeControlEvent
	}
	encoded, err := (proto.MarshalOptions{Deterministic: true}).Marshal(event)
	if err != nil || len(encoded) == 0 {
		return nil, errInvalidNodeControlEvent
	}
	return append([]byte(nil), encoded...), nil
}

// UnmarshalNodeControlEvent rejects unknown and noncanonical protobuf bytes before returning an independent event.
func UnmarshalNodeControlEvent(encoded []byte) (*nodecontrolv1.NodeControlEventV1, error) {
	if len(encoded) == 0 {
		return nil, errInvalidNodeControlEvent
	}
	copyOfEncoded := append([]byte(nil), encoded...)
	defer clear(copyOfEncoded)

	event := new(nodecontrolv1.NodeControlEventV1)
	if err := (proto.UnmarshalOptions{DiscardUnknown: false}).Unmarshal(copyOfEncoded, event); err != nil ||
		nodeControlHasUnknownFields(event.ProtoReflect()) {
		return nil, errInvalidNodeControlEvent
	}
	canonical, err := MarshalNodeControlEvent(event)
	if err != nil || !bytes.Equal(canonical, copyOfEncoded) {
		return nil, errInvalidNodeControlEvent
	}
	return proto.Clone(event).(*nodecontrolv1.NodeControlEventV1), nil
}

// NodeControlSubject returns the one permitted subject for a registered node-control event type.
func NodeControlSubject(eventType string) (string, error) {
	if !registeredNodeControlEventType(eventType) {
		return "", errInvalidNodeControlSubject
	}
	return nodeControlSubjectPrefix + eventType, nil
}

func validateNodeControlEvent(event *nodecontrolv1.NodeControlEventV1) error {
	if !canonicalNodeControlUUID(event.GetEventId()) || event.GetOccurredAt() == nil || !event.GetOccurredAt().IsValid() ||
		!safeNodeControlName(event.GetAggregateType(), 64) || !canonicalNodeControlUUID(event.GetAggregateId()) ||
		event.GetAggregateVersion() == 0 || event.GetAggregateVersion() > math.MaxInt64 {
		return errInvalidNodeControlEvent
	}

	switch event.GetEventType() {
	case nodeInventoryChangedType:
		payload, ok := event.GetPayload().(*nodecontrolv1.NodeControlEventV1_NodeInventoryChanged)
		if !ok || !validNodeInventoryChanged(payload.NodeInventoryChanged) ||
			event.GetAggregateType() != "node" || event.GetAggregateId() != payload.NodeInventoryChanged.GetNodeId() ||
			event.GetAggregateVersion() != payload.NodeInventoryChanged.GetInventoryVersion() {
			return errInvalidNodeControlEvent
		}
		return validateNodeControlPayloadDescriptor(payload.NodeInventoryChanged)
	case nodeDesiredStatePublishedType:
		payload, ok := event.GetPayload().(*nodecontrolv1.NodeControlEventV1_NodeDesiredStatePublished)
		if !ok || !validNodeDesiredStatePublished(payload.NodeDesiredStatePublished) ||
			event.GetAggregateType() != "node" || event.GetAggregateId() != payload.NodeDesiredStatePublished.GetNodeId() ||
			event.GetAggregateVersion() != payload.NodeDesiredStatePublished.GetGeneration() {
			return errInvalidNodeControlEvent
		}
		return validateNodeControlPayloadDescriptor(payload.NodeDesiredStatePublished)
	case nodeAvailabilityChangedType:
		payload, ok := event.GetPayload().(*nodecontrolv1.NodeControlEventV1_NodeAvailabilityChanged)
		if !ok || !validNodeAvailabilityChanged(payload.NodeAvailabilityChanged) ||
			event.GetAggregateType() != "node" || event.GetAggregateId() != payload.NodeAvailabilityChanged.GetNodeId() ||
			event.GetAggregateVersion() != payload.NodeAvailabilityChanged.GetSequence() {
			return errInvalidNodeControlEvent
		}
		return validateNodeControlPayloadDescriptor(payload.NodeAvailabilityChanged)
	case nodeSecurityStateChangedType:
		payload, ok := event.GetPayload().(*nodecontrolv1.NodeControlEventV1_NodeSecurityStateChanged)
		if !ok || !validNodeSecurityStateChanged(payload.NodeSecurityStateChanged) ||
			event.GetAggregateType() != "node" || event.GetAggregateId() != payload.NodeSecurityStateChanged.GetNodeId() {
			return errInvalidNodeControlEvent
		}
		return validateNodeControlPayloadDescriptor(payload.NodeSecurityStateChanged)
	case nodeCertificateStatusChangedType:
		payload, ok := event.GetPayload().(*nodecontrolv1.NodeControlEventV1_NodeCertificateStatusChanged)
		if !ok || !validNodeCertificateStatusChanged(payload.NodeCertificateStatusChanged) ||
			event.GetAggregateType() != "node" || event.GetAggregateId() != payload.NodeCertificateStatusChanged.GetNodeId() {
			return errInvalidNodeControlEvent
		}
		return validateNodeControlPayloadDescriptor(payload.NodeCertificateStatusChanged)
	case nodeOperatorActionRecordedType:
		payload, ok := event.GetPayload().(*nodecontrolv1.NodeControlEventV1_NodeOperatorActionRecorded)
		if !ok || !validNodeOperatorActionRecorded(payload.NodeOperatorActionRecorded) ||
			event.GetAggregateType() != "operator_action" || event.GetAggregateId() != payload.NodeOperatorActionRecorded.GetAuditId() {
			return errInvalidNodeControlEvent
		}
		return validateNodeControlPayloadDescriptor(payload.NodeOperatorActionRecorded)
	default:
		return errInvalidNodeControlEvent
	}
}

func validNodeInventoryChanged(payload *nodecontrolv1.NodeInventoryChangedV1) bool {
	return payload != nil && canonicalNodeControlUUID(payload.GetNodeId()) &&
		payload.GetInventoryVersion() > 0 && payload.GetInventoryVersion() <= math.MaxInt64 &&
		safeNodeControlName(payload.GetPopCode(), 64) && safeNodeControlName(payload.GetOperatorState(), 64)
}

func validNodeDesiredStatePublished(payload *nodecontrolv1.NodeDesiredStatePublishedV1) bool {
	return payload != nil && canonicalNodeControlUUID(payload.GetNodeId()) &&
		payload.GetGeneration() > 0 && payload.GetGeneration() <= math.MaxInt64 && len(payload.GetContentDigest()) == 32 &&
		payload.GetValidUntil() != nil && payload.GetValidUntil().IsValid() && safeNodeControlName(payload.GetReason(), 128)
}

func validNodeAvailabilityChanged(payload *nodecontrolv1.NodeAvailabilityChangedV1) bool {
	return payload != nil && canonicalNodeControlUUID(payload.GetNodeId()) && canonicalNodeControlUUID(payload.GetBootId()) &&
		payload.GetSequence() > 0 && payload.GetSequence() <= math.MaxInt64 && safeNodeControlName(payload.GetHealth(), 64) &&
		payload.AcceptingNew != nil && safeNodeControlName(payload.GetReason(), 128)
}

func validNodeSecurityStateChanged(payload *nodecontrolv1.NodeSecurityStateChangedV1) bool {
	return payload != nil && canonicalNodeControlUUID(payload.GetNodeId()) && canonicalNodeControlUUID(payload.GetIncidentId()) &&
		safeNodeControlName(payload.GetFaultSubtype(), 64) && safeNodeControlName(payload.GetSecurityState(), 64) &&
		safeNodeControlName(payload.GetReason(), 128)
}

func validNodeCertificateStatusChanged(payload *nodecontrolv1.NodeCertificateStatusChangedV1) bool {
	return payload != nil && canonicalNodeControlUUID(payload.GetNodeId()) && canonicalNodeControlUUID(payload.GetCertificateRecordId()) &&
		safeNodeControlName(payload.GetStatus(), 64)
}

func validNodeOperatorActionRecorded(payload *nodecontrolv1.NodeOperatorActionRecordedV1) bool {
	return payload != nil && canonicalNodeControlUUID(payload.GetAuditId()) && canonicalNodeControlUUID(payload.GetOperatorId()) &&
		safeNodeControlName(payload.GetAction(), 64) && safeNodeControlName(payload.GetTarget(), 64) &&
		safeNodeControlName(payload.GetResult(), 64) && safeNodeControlName(payload.GetReason(), 128)
}

func validateNodeControlPayloadDescriptor(payload proto.Message) error {
	if payload == nil || sharedevents.ValidatePayloadDescriptor(payload.ProtoReflect().Descriptor()) != nil {
		return errInvalidNodeControlEvent
	}
	return nil
}

func registeredNodeControlEventType(eventType string) bool {
	switch eventType {
	case nodeInventoryChangedType, nodeDesiredStatePublishedType, nodeAvailabilityChangedType,
		nodeSecurityStateChangedType, nodeCertificateStatusChangedType, nodeOperatorActionRecordedType:
		return true
	default:
		return false
	}
}

func canonicalNodeControlUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}

func safeNodeControlName(value string, maximum int) bool {
	if len(value) == 0 || len(value) > maximum || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
}

func nodeControlHasUnknownFields(message protoreflect.Message) bool {
	if !message.IsValid() || len(message.GetUnknown()) != 0 {
		return true
	}
	hasUnknown := false
	message.Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		switch {
		case field.IsMap():
			if field.MapValue().Message() == nil {
				return true
			}
			value.Map().Range(func(_ protoreflect.MapKey, entry protoreflect.Value) bool {
				hasUnknown = nodeControlHasUnknownFields(entry.Message())
				return !hasUnknown
			})
		case field.IsList():
			if field.Message() == nil {
				return true
			}
			for index := 0; index < value.List().Len(); index++ {
				if nodeControlHasUnknownFields(value.List().Get(index).Message()) {
					hasUnknown = true
					return false
				}
			}
		case field.Message() != nil:
			hasUnknown = nodeControlHasUnknownFields(value.Message())
		}
		return !hasUnknown
	})
	return hasUnknown
}
