package events

import (
	"bytes"
	"errors"
	"math"
	"strconv"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	deviceauthv1 "talenro.local/platform/gen/go/talenro/deviceauth/v1"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	identityv1 "talenro.local/platform/gen/go/talenro/identity/v1"
	trustv1 "talenro.local/platform/gen/go/talenro/trust/v1"
)

const (
	// AccountStateChangedType is the registered account state event type.
	AccountStateChangedType = "talenro.identity.account_state_changed.v1"
	// EmailDeliveryRequestedType is the registered protected-delivery event type.
	EmailDeliveryRequestedType = "talenro.identity.email_delivery_requested.v1"
	// DeviceAuthorizationChangedType is the registered device authorization event type.
	DeviceAuthorizationChangedType = "talenro.deviceauth.authorization_changed.v1"
	// DeviceTokenFamilyCompromisedType is the registered token-family event type.
	DeviceTokenFamilyCompromisedType = "talenro.deviceauth.token_family_compromised.v1" // #nosec G101 -- public event type, not a credential.
	// BundleIssuedType is the registered trust bundle issuance event type.
	BundleIssuedType = "talenro.trust.bundle_issued.v1"
	// BundleAcknowledgedType is the registered trust bundle acknowledgement event type.
	BundleAcknowledgedType = "talenro.trust.bundle_acknowledged.v1"

	maximumEnvelopeBytes = 264 * 1024
)

var (
	// ErrInvalidEnvelope reports malformed envelope metadata without retaining values.
	ErrInvalidEnvelope = errors.New("events: invalid envelope")
	// ErrInvalidPayload reports malformed or unregistered payloads without retaining values.
	ErrInvalidPayload = errors.New("events: invalid payload")
	// ErrUnsafePayload reports a payload descriptor capable of carrying forbidden material.
	ErrUnsafePayload = errors.New("events: unsafe payload")
)

type payloadSpec struct {
	newMessage     func() proto.Message
	messageName    protoreflect.FullName
	aggregateType  string
	aggregateID    func(proto.Message) string
	payloadVersion func(proto.Message) (uint64, bool)
	validate       func(proto.Message) bool
}

// MarshalPayload validates a finite event type and deterministically marshals its exact message.
func MarshalPayload(eventType string, payload proto.Message) ([]byte, error) {
	if payload == nil || !payload.ProtoReflect().IsValid() {
		return nil, ErrInvalidPayload
	}
	if err := ValidatePayloadDescriptor(payload.ProtoReflect().Descriptor()); err != nil {
		return nil, ErrUnsafePayload
	}
	spec, ok := registeredPayload(eventType)
	if !ok || payload.ProtoReflect().Descriptor().FullName() != spec.messageName || !spec.validate(payload) {
		return nil, ErrInvalidPayload
	}
	encoded, err := (proto.MarshalOptions{Deterministic: true}).Marshal(payload)
	if err != nil || len(encoded) == 0 || len(encoded) > maxPayloadSize {
		return nil, ErrInvalidPayload
	}
	return append([]byte(nil), encoded...), nil
}

// MarshalEnvelope validates metadata and payload, then returns canonical protobuf bytes.
func MarshalEnvelope(envelope *eventsv1.EventEnvelope) ([]byte, error) {
	if err := validateEnvelope(envelope); err != nil {
		return nil, err
	}
	encoded, err := (proto.MarshalOptions{Deterministic: true}).Marshal(envelope)
	if err != nil || len(encoded) == 0 || len(encoded) > maximumEnvelopeBytes {
		return nil, ErrInvalidEnvelope
	}
	return append([]byte(nil), encoded...), nil
}

// UnmarshalEnvelope rejects unknown or noncanonical wire data and returns an independent envelope.
func UnmarshalEnvelope(encoded []byte) (*eventsv1.EventEnvelope, error) {
	if len(encoded) == 0 || len(encoded) > maximumEnvelopeBytes {
		return nil, ErrInvalidEnvelope
	}
	copyOfEncoded := append([]byte(nil), encoded...)
	defer clear(copyOfEncoded)
	envelope := new(eventsv1.EventEnvelope)
	if err := (proto.UnmarshalOptions{DiscardUnknown: false}).Unmarshal(copyOfEncoded, envelope); err != nil ||
		len(envelope.ProtoReflect().GetUnknown()) != 0 {
		return nil, ErrInvalidEnvelope
	}
	canonical, err := MarshalEnvelope(envelope)
	if err != nil || !bytes.Equal(canonical, copyOfEncoded) {
		return nil, ErrInvalidEnvelope
	}
	return proto.Clone(envelope).(*eventsv1.EventEnvelope), nil
}

func validateEnvelope(envelope *eventsv1.EventEnvelope) error {
	if ValidateEnvelope(envelope) != nil || !canonicalUUID(envelope.GetEventId()) ||
		!safeName(envelope.GetProducer(), 64, true) || !safeName(envelope.GetAggregateType(), 64, false) ||
		!canonicalUUID(envelope.GetAggregateId()) || envelope.GetAggregateVersion() == 0 || envelope.GetAggregateVersion() > math.MaxInt64 ||
		!safeIdempotencyKey(envelope.GetIdempotencyKey()) || !validTraceParent(envelope.GetTraceParent()) {
		return ErrInvalidEnvelope
	}
	occurredAt := envelope.GetOccurredAt().AsTime()
	if occurredAt.Year() < 2020 || occurredAt.Year() > 2100 {
		return ErrInvalidEnvelope
	}
	spec, message, err := decodePayload(envelope.GetEventType(), envelope.GetPayload())
	if err != nil || envelope.GetAggregateType() != spec.aggregateType || envelope.GetAggregateId() != spec.aggregateID(message) {
		return ErrInvalidEnvelope
	}
	if version, exact := spec.payloadVersion(message); exact && version != envelope.GetAggregateVersion() {
		return ErrInvalidEnvelope
	}
	return nil
}

func decodePayload(eventType string, encoded []byte) (payloadSpec, proto.Message, error) {
	spec, ok := registeredPayload(eventType)
	if !ok || len(encoded) == 0 || len(encoded) > maxPayloadSize {
		return payloadSpec{}, nil, ErrInvalidPayload
	}
	message := spec.newMessage()
	if err := (proto.UnmarshalOptions{DiscardUnknown: false}).Unmarshal(encoded, message); err != nil ||
		len(message.ProtoReflect().GetUnknown()) != 0 || ValidatePayloadDescriptor(message.ProtoReflect().Descriptor()) != nil || !spec.validate(message) {
		return payloadSpec{}, nil, ErrInvalidPayload
	}
	canonical, err := (proto.MarshalOptions{Deterministic: true}).Marshal(message)
	if err != nil || !bytes.Equal(canonical, encoded) {
		return payloadSpec{}, nil, ErrInvalidPayload
	}
	return spec, message, nil
}

func registeredPayload(eventType string) (payloadSpec, bool) {
	switch eventType {
	case AccountStateChangedType:
		return payloadSpec{
			newMessage:  func() proto.Message { return new(identityv1.AccountStateChanged) },
			messageName: "talenro.identity.v1.AccountStateChanged", aggregateType: "account",
			aggregateID: func(message proto.Message) string { return message.(*identityv1.AccountStateChanged).GetPrincipalId() },
			payloadVersion: func(message proto.Message) (uint64, bool) {
				return message.(*identityv1.AccountStateChanged).GetVersion(), true
			},
			validate: func(message proto.Message) bool {
				payload := message.(*identityv1.AccountStateChanged)
				return canonicalUUID(payload.GetPrincipalId()) && safeName(payload.GetState(), 32, false) && payload.GetVersion() > 0
			},
		}, true
	case EmailDeliveryRequestedType:
		return payloadSpec{
			newMessage:  func() proto.Message { return new(identityv1.EmailDeliveryRequested) },
			messageName: "talenro.identity.v1.EmailDeliveryRequested", aggregateType: "account",
			aggregateID: func(message proto.Message) string {
				return message.(*identityv1.EmailDeliveryRequested).GetPrincipalId()
			},
			payloadVersion: func(proto.Message) (uint64, bool) { return 0, false },
			validate: func(message proto.Message) bool {
				payload := message.(*identityv1.EmailDeliveryRequested)
				return canonicalUUID(payload.GetDeliveryId()) && canonicalUUID(payload.GetPrincipalId()) &&
					safeName(payload.GetTemplateId(), 64, false) && validLocale(payload.GetLocale())
			},
		}, true
	case DeviceAuthorizationChangedType:
		return payloadSpec{
			newMessage:  func() proto.Message { return new(deviceauthv1.DeviceAuthorizationChanged) },
			messageName: "talenro.deviceauth.v1.DeviceAuthorizationChanged", aggregateType: "device_authorization",
			aggregateID: func(message proto.Message) string {
				return message.(*deviceauthv1.DeviceAuthorizationChanged).GetAuthorizationId()
			},
			payloadVersion: func(message proto.Message) (uint64, bool) {
				return message.(*deviceauthv1.DeviceAuthorizationChanged).GetVersion(), true
			},
			validate: func(message proto.Message) bool {
				payload := message.(*deviceauthv1.DeviceAuthorizationChanged)
				return canonicalUUID(payload.GetPrincipalId()) && canonicalUUID(payload.GetDeviceId()) && canonicalUUID(payload.GetAuthorizationId()) &&
					safeName(payload.GetState(), 32, false) && payload.GetVersion() > 0
			},
		}, true
	case DeviceTokenFamilyCompromisedType:
		return payloadSpec{
			newMessage:  func() proto.Message { return new(deviceauthv1.DeviceTokenFamilyCompromised) },
			messageName: "talenro.deviceauth.v1.DeviceTokenFamilyCompromised", aggregateType: "device_authorization",
			aggregateID: func(message proto.Message) string {
				return message.(*deviceauthv1.DeviceTokenFamilyCompromised).GetAuthorizationId()
			},
			payloadVersion: func(message proto.Message) (uint64, bool) {
				return message.(*deviceauthv1.DeviceTokenFamilyCompromised).GetVersion(), true
			},
			validate: func(message proto.Message) bool {
				payload := message.(*deviceauthv1.DeviceTokenFamilyCompromised)
				return canonicalUUID(payload.GetAuthorizationId()) && canonicalUUID(payload.GetFamilyId()) && payload.GetVersion() > 0
			},
		}, true
	case BundleIssuedType:
		return payloadSpec{
			newMessage:  func() proto.Message { return new(trustv1.BundleIssued) },
			messageName: "talenro.trust.v1.BundleIssued", aggregateType: "bundle",
			aggregateID:    func(message proto.Message) string { return message.(*trustv1.BundleIssued).GetBundleId() },
			payloadVersion: bundleIssuedVersion,
			validate: func(message proto.Message) bool {
				payload := message.(*trustv1.BundleIssued)
				_, validVersion := positiveDecimal(payload.GetBundleVersion())
				return canonicalUUID(payload.GetAuthorizationId()) && canonicalUUID(payload.GetBundleId()) && validVersion && lowerHex(payload.GetEnvelopeSha256(), 64)
			},
		}, true
	case BundleAcknowledgedType:
		return payloadSpec{
			newMessage:  func() proto.Message { return new(trustv1.BundleAcknowledged) },
			messageName: "talenro.trust.v1.BundleAcknowledged", aggregateType: "bundle",
			aggregateID:    func(message proto.Message) string { return message.(*trustv1.BundleAcknowledged).GetBundleId() },
			payloadVersion: bundleAcknowledgedVersion,
			validate: func(message proto.Message) bool {
				payload := message.(*trustv1.BundleAcknowledged)
				_, validVersion := positiveDecimal(payload.GetBundleVersion())
				return canonicalUUID(payload.GetAuthorizationId()) && canonicalUUID(payload.GetBundleId()) && validVersion
			},
		}, true
	default:
		return payloadSpec{}, false
	}
}

func bundleIssuedVersion(message proto.Message) (uint64, bool) {
	return positiveDecimal(message.(*trustv1.BundleIssued).GetBundleVersion())
}

func bundleAcknowledgedVersion(message proto.Message) (uint64, bool) {
	return positiveDecimal(message.(*trustv1.BundleAcknowledged).GetBundleVersion())
}

func positiveDecimal(value string) (uint64, bool) {
	if len(value) == 0 || len(value) > 19 || value[0] == '0' {
		return 0, false
	}
	parsed, err := strconv.ParseUint(value, 10, 63)
	return parsed, err == nil && parsed > 0
}

func canonicalUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}

func safeName(value string, maximum int, allowDot bool) bool {
	if len(value) == 0 || len(value) > maximum || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' || character == '_' || character == '-' || allowDot && character == '.' {
			continue
		}
		return false
	}
	return true
}

func safeIdempotencyKey(value string) bool {
	if len(value) == 0 || len(value) > 128 {
		return false
	}
	for index := range len(value) {
		character := value[index]
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' ||
			character >= '0' && character <= '9' || character == ':' || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
}

func validLocale(value string) bool {
	if len(value) < 2 || len(value) > 16 {
		return false
	}
	for index := range len(value) {
		character := value[index]
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '-' {
			continue
		}
		return false
	}
	return true
}

func validTraceParent(value string) bool {
	if value == "" {
		return true
	}
	if len(value) != 55 || value[2] != '-' || value[35] != '-' || value[52] != '-' {
		return false
	}
	for index := range len(value) {
		if index == 2 || index == 35 || index == 52 {
			continue
		}
		character := value[index]
		if character < '0' || character > '9' {
			if character < 'a' || character > 'f' {
				return false
			}
		}
	}
	return value[:2] != "ff" && value[3:35] != "00000000000000000000000000000000" && value[36:52] != "0000000000000000"
}

func lowerHex(value string, size int) bool {
	if len(value) != size {
		return false
	}
	for index := range len(value) {
		character := value[index]
		if character < '0' || character > '9' {
			if character < 'a' || character > 'f' {
				return false
			}
		}
	}
	return true
}
