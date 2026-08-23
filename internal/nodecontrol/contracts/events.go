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

	nodeControlOperatorStates = nodeControlRegistry("provisioning", "enabled", "draining", "disabled")
	nodeControlDesiredReasons = nodeControlRegistry(
		"initial", "operator_update", "drain", "resume", "lease_refresh", "clear_slot_quarantine", "restore_reauthorize",
	)
	nodeControlHealthStates        = nodeControlRegistry("unknown", "healthy", "degraded", "offline", "quarantined")
	nodeControlAvailabilityReasons = nodeControlRegistry(
		"observation", "observation_timeout", "boot_changed", "capacity_blocked", "profile_mismatch", "agent_reducer_mismatch",
	)
	nodeControlFaultSubtypes = nodeControlRegistry(
		"identity_compromise", "online_signer_equivocation", "metadata_rollback", "root_rollback", "root_equivocation",
		"unverified_client_highwater_conflict", "client_highwater_ahead", "server_trust_bundle_conflict",
		"trusted_time_rollback_or_unavailable", "local_state_corruption_or_rollback", "release_or_process_integrity",
		"profile_binding_mismatch", "incident_overflow",
	)
	nodeControlSecurityStates      = nodeControlRegistry("normal", "quarantined")
	nodeControlSecurityReasons     = nodeControlRegistry("fault_confirmed", "incident_capacity_exceeded", "host_remediation", "incident_resolved")
	nodeControlCertificateStatuses = nodeControlRegistry(
		"active", "recovery_pending", "recovery_limited", "revoked",
	)
	nodeControlOperatorActions = nodeControlRegistry(
		"create_node_pop", "update_node_pop", "create_node_failure_domain", "update_node_failure_domain", "create_node",
		"update_node", "create_node_endpoint", "update_node_endpoint", "create_node_process_slot", "update_node_process_slot",
		"create_node_enrollment_grant", "put_node_desired_state", "drain_node", "disable_node", "reenroll_node",
		"complete_node_reenrollment", "register_node_host_security_incident", "register_node_resource_envelope",
		"clear_node_security_quarantine", "resume_node_after_security", "reauthorize_node_after_restore",
	)
	nodeControlOperatorTargets = nodeControlRegistry(
		"node_pop", "node_failure_domain", "node", "node_endpoint", "node_process_slot", "node_enrollment_grant",
		"node_desired_state", "node_security_incident", "node_resource_envelope", "node_recovery_session",
	)
	nodeControlOperatorResults = nodeControlRegistry("pending", "active", "accepted", "completed", "superseded", "rejected")
	nodeControlOperatorReasons = nodeControlRegistry(
		"provision", "inventory_update", "capacity_change", "drain_maintenance", "administrative_disable", "retire",
		"identity_compromise", "host_remediation", "security_recovery", "authority_restore", "release_update", "initial",
		"operator_update", "drain", "resume", "lease_refresh", "clear_slot_quarantine", "restore_reauthorize",
	)
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
		!canonicalNodeControlUUID(event.GetAggregateId()) || event.GetAggregateVersion() == 0 ||
		event.GetAggregateVersion() > math.MaxInt64 {
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
		canonicalNodeControlPOPCode(payload.GetPopCode()) && nodeControlOperatorStates.has(payload.GetOperatorState())
}

func validNodeDesiredStatePublished(payload *nodecontrolv1.NodeDesiredStatePublishedV1) bool {
	return payload != nil && canonicalNodeControlUUID(payload.GetNodeId()) &&
		payload.GetGeneration() > 0 && payload.GetGeneration() <= math.MaxInt64 && len(payload.GetContentDigest()) == 32 &&
		payload.GetValidUntil() != nil && payload.GetValidUntil().IsValid() && nodeControlDesiredReasons.has(payload.GetReason())
}

func validNodeAvailabilityChanged(payload *nodecontrolv1.NodeAvailabilityChangedV1) bool {
	return payload != nil && canonicalNodeControlUUID(payload.GetNodeId()) && canonicalNodeControlUUID(payload.GetBootId()) &&
		payload.GetSequence() > 0 && payload.GetSequence() <= math.MaxInt64 && nodeControlHealthStates.has(payload.GetHealth()) &&
		payload.AcceptingNew != nil && nodeControlAvailabilityReasons.has(payload.GetReason())
}

func validNodeSecurityStateChanged(payload *nodecontrolv1.NodeSecurityStateChangedV1) bool {
	return payload != nil && canonicalNodeControlUUID(payload.GetNodeId()) && canonicalNodeControlUUID(payload.GetIncidentId()) &&
		nodeControlFaultSubtypes.has(payload.GetFaultSubtype()) && nodeControlSecurityStates.has(payload.GetSecurityState()) &&
		nodeControlSecurityReasons.has(payload.GetReason())
}

func validNodeCertificateStatusChanged(payload *nodecontrolv1.NodeCertificateStatusChangedV1) bool {
	return payload != nil && canonicalNodeControlUUID(payload.GetNodeId()) && canonicalNodeControlUUID(payload.GetCertificateRecordId()) &&
		nodeControlCertificateStatuses.has(payload.GetStatus())
}

func validNodeOperatorActionRecorded(payload *nodecontrolv1.NodeOperatorActionRecordedV1) bool {
	return payload != nil && canonicalNodeControlUUID(payload.GetAuditId()) && canonicalNodeControlUUID(payload.GetOperatorId()) &&
		nodeControlOperatorActions.has(payload.GetAction()) && nodeControlOperatorTargets.has(payload.GetTarget()) &&
		nodeControlOperatorResults.has(payload.GetResult()) && nodeControlOperatorReasons.has(payload.GetReason())
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

type nodeControlValueRegistry map[string]struct{}

func nodeControlRegistry(values ...string) nodeControlValueRegistry {
	registry := make(nodeControlValueRegistry, len(values))
	for _, value := range values {
		registry[value] = struct{}{}
	}
	return registry
}

func (registry nodeControlValueRegistry) has(value string) bool {
	_, ok := registry[value]
	return ok
}

func canonicalNodeControlPOPCode(value string) bool {
	if len(value) < 2 || len(value) > 32 || !lowercaseAlphaNumeric(value[0]) || !lowercaseAlphaNumeric(value[len(value)-1]) {
		return false
	}
	for index := 1; index < len(value)-1; index++ {
		character := value[index]
		if lowercaseAlphaNumeric(character) || character == '-' {
			continue
		}
		return false
	}
	return true
}

func lowercaseAlphaNumeric(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= '0' && value <= '9'
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
