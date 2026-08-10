package outbox

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	identityv1 "talenro.local/platform/gen/go/talenro/identity/v1"
	trustv1 "talenro.local/platform/gen/go/talenro/trust/v1"
	contractevents "talenro.local/platform/internal/contracts/events"
	"talenro.local/platform/internal/store"
)

func TestRepositoryAppendsDeterministicValidatedEnvelope(t *testing.T) {
	t.Parallel()

	envelope := accountStateEnvelope(t)
	storeFake := &fakeOutboxStore{}
	repository := newRepositoryWithStore(storeFake)
	if err := repository.Append(context.Background(), envelope); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if storeFake.calls != 1 {
		t.Fatalf("insert calls = %d, want 1", storeFake.calls)
	}
	params := storeFake.params
	storedSnapshot := bytes.Clone(params.Payload)
	if params.EventID.String() != envelope.EventId || params.EventType != envelope.EventType ||
		params.AggregateType != envelope.AggregateType || params.AggregateID.String() != envelope.AggregateId ||
		params.AggregateVersion != 3 || params.IdempotencyKey != envelope.IdempotencyKey ||
		!params.OccurredAt.Equal(envelope.OccurredAt.AsTime()) || !params.AvailableAt.Equal(envelope.OccurredAt.AsTime()) {
		t.Fatalf("insert mapping mismatch: %+v", params)
	}
	want, err := contractevents.MarshalEnvelope(envelope)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(params.Payload, want) {
		t.Fatal("stored payload is not deterministic envelope encoding")
	}
	envelope.Payload[0] ^= 1
	envelope.EventId = uuid.NewString()
	if !bytes.Equal(params.Payload, storedSnapshot) {
		t.Fatal("stored payload changed after caller mutated envelope")
	}
	decoded, err := contractevents.UnmarshalEnvelope(params.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.EventId != params.EventID.String() {
		t.Fatal("stored envelope did not round trip")
	}
}

func TestOutboxPayloadRejectsPIIFields(t *testing.T) {
	t.Parallel()

	file, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Syntax: proto.String("proto3"), Name: proto.String("pii.proto"), Package: proto.String("talenro.test.v1"),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("Unsafe"),
			Field: []*descriptorpb.FieldDescriptorProto{{
				Name: proto.String("refresh_token"), Number: proto.Int32(1),
				Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
			}},
		}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	unsafe := dynamicpb.NewMessage(file.Messages().Get(0))
	unsafe.Set(file.Messages().Get(0).Fields().Get(0), protoreflect.ValueOfString("secret"))
	if _, err := contractevents.MarshalPayload(contractevents.AccountStateChangedType, unsafe); !errors.Is(err, contractevents.ErrUnsafePayload) {
		t.Fatalf("error = %v, want ErrUnsafePayload", err)
	}
}

func TestMarshalRejectsUnknownPayloadEnvelopeAndNestedFields(t *testing.T) {
	t.Parallel()

	unknown := protowire.AppendTag(nil, 99, protowire.VarintType)
	unknown = protowire.AppendVarint(unknown, 1)
	payload := &identityv1.AccountStateChanged{
		PrincipalId: "4b4d278b-9e7a-4ce0-865d-1dc14fcf96da", State: "active", Version: 3,
	}
	payload.ProtoReflect().SetUnknown(bytes.Clone(unknown))
	if _, err := contractevents.MarshalPayload(contractevents.AccountStateChangedType, payload); !errors.Is(err, contractevents.ErrInvalidPayload) {
		t.Fatalf("payload unknown error = %v, want ErrInvalidPayload", err)
	}

	for _, mutate := range []func(*eventsv1.EventEnvelope){
		func(envelope *eventsv1.EventEnvelope) { envelope.ProtoReflect().SetUnknown(bytes.Clone(unknown)) },
		func(envelope *eventsv1.EventEnvelope) {
			envelope.OccurredAt.ProtoReflect().SetUnknown(bytes.Clone(unknown))
		},
	} {
		envelope := accountStateEnvelope(t)
		mutate(envelope)
		if _, err := contractevents.MarshalEnvelope(envelope); !errors.Is(err, contractevents.ErrInvalidEnvelope) {
			t.Fatalf("envelope unknown error = %v, want ErrInvalidEnvelope", err)
		}
	}
}

func TestMarshalEnvelopeRoundTripsThroughStrictUnmarshal(t *testing.T) {
	t.Parallel()

	envelope := accountStateEnvelope(t)
	encoded, err := contractevents.MarshalEnvelope(envelope)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := contractevents.UnmarshalEnvelope(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(envelope, decoded) {
		t.Fatal("marshal output did not close over strict unmarshal")
	}
}

func TestEmailDeliveryLocaleMatchesFrozenContract(t *testing.T) {
	t.Parallel()

	base := &identityv1.EmailDeliveryRequested{
		DeliveryId:  "f4232063-70a4-4ef5-8627-5cd9f2a8ab9d",
		PrincipalId: "4b4d278b-9e7a-4ce0-865d-1dc14fcf96da",
		TemplateId:  "verify_email",
	}
	for _, locale := range []string{"en", "fa-IR", "zh-Hans-CN-variant8", "abc-12345678-12345678-12345678"} {
		payload := proto.Clone(base).(*identityv1.EmailDeliveryRequested)
		payload.Locale = locale
		if _, err := contractevents.MarshalPayload(contractevents.EmailDeliveryRequestedType, payload); err != nil {
			t.Fatalf("valid locale %q: %v", locale, err)
		}
	}
	for _, locale := range []string{"12", "en-", "en--US", "e", "en-123456789", "abc-12345678-12345678-12345678-12345678"} {
		payload := proto.Clone(base).(*identityv1.EmailDeliveryRequested)
		payload.Locale = locale
		if _, err := contractevents.MarshalPayload(contractevents.EmailDeliveryRequestedType, payload); !errors.Is(err, contractevents.ErrInvalidPayload) {
			t.Fatalf("invalid locale %q error = %v", locale, err)
		}
	}
}

func TestRepositoryRejectsUnknownMutationAndMismatchedPayloadBeforeInsert(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*eventsv1.EventEnvelope)
	}{
		{name: "unknown event type", mutate: func(e *eventsv1.EventEnvelope) { e.EventType = "talenro.identity.unknown.v1" }},
		{name: "registered payload type mismatch", mutate: func(e *eventsv1.EventEnvelope) {
			mismatched, err := contractevents.MarshalPayload(contractevents.BundleIssuedType, &trustv1.BundleIssued{
				AuthorizationId: uuid.NewString(), BundleId: uuid.NewString(), BundleVersion: "3",
				EnvelopeSha256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			})
			if err != nil {
				t.Fatal(err)
			}
			e.Payload = mismatched
		}},
		{name: "unknown payload field", mutate: func(e *eventsv1.EventEnvelope) {
			e.Payload = protowire.AppendTag(e.Payload, 99, protowire.VarintType)
			e.Payload = protowire.AppendVarint(e.Payload, 1)
		}},
		{name: "aggregate id mismatch", mutate: func(e *eventsv1.EventEnvelope) { e.AggregateId = uuid.NewString() }},
		{name: "non canonical event uuid", mutate: func(e *eventsv1.EventEnvelope) { e.EventId = "{" + e.EventId + "}" }},
		{name: "invalid idempotency key", mutate: func(e *eventsv1.EventEnvelope) { e.IdempotencyKey = "contains/slash" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			envelope := accountStateEnvelope(t)
			test.mutate(envelope)
			storeFake := &fakeOutboxStore{}
			if err := newRepositoryWithStore(storeFake).Append(context.Background(), envelope); !errors.Is(err, ErrInvalidEvent) {
				t.Fatalf("error = %v, want ErrInvalidEvent", err)
			}
			if storeFake.calls != 0 {
				t.Fatal("invalid event reached store")
			}
		})
	}
}

func TestRepositoryRejectsTypedNilAndCanceledDependenciesFailClosed(t *testing.T) {
	t.Parallel()

	if _, err := NewRepository(nil); !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("nil DBTX error = %v", err)
	}
	var nilDB *typedNilOutboxDBTX
	if _, err := NewRepository(nilDB); !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("typed-nil DBTX error = %v", err)
	}
	var nilStore *typedNilOutboxStore
	if err := newRepositoryWithStore(nilStore).Append(context.Background(), accountStateEnvelope(t)); !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("typed-nil store error = %v", err)
	}
	var zero repository
	if err := zero.Append(context.Background(), accountStateEnvelope(t)); !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("zero repository error = %v", err)
	}

	preCanceled, cancel := context.WithCancel(context.Background())
	cancel()
	storeFake := &fakeOutboxStore{}
	if err := newRepositoryWithStore(storeFake).Append(preCanceled, accountStateEnvelope(t)); !errors.Is(err, ErrCanceled) {
		t.Fatalf("pre-canceled error = %v, want ErrCanceled", err)
	}
	if storeFake.calls != 0 {
		t.Fatal("pre-canceled append reached store")
	}

	operationContext, cancelDuringStore := context.WithCancel(context.Background())
	storeFake = &fakeOutboxStore{err: errors.New("SECRET_STORE_CANARY"), cancelOnInsert: cancelDuringStore}
	if err := newRepositoryWithStore(storeFake).Append(operationContext, accountStateEnvelope(t)); !errors.Is(err, ErrCanceled) {
		t.Fatalf("store cancellation error = %v, want ErrCanceled", err)
	}
}

func accountStateEnvelope(t *testing.T) *eventsv1.EventEnvelope {
	t.Helper()
	principalID := uuid.MustParse("4b4d278b-9e7a-4ce0-865d-1dc14fcf96da")
	payload, err := contractevents.MarshalPayload(contractevents.AccountStateChangedType, &identityv1.AccountStateChanged{
		PrincipalId: principalID.String(), State: "active", Version: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	return &eventsv1.EventEnvelope{
		EventId:    uuid.MustParse("f4232063-70a4-4ef5-8627-5cd9f2a8ab9d").String(),
		EventType:  contractevents.AccountStateChangedType,
		OccurredAt: timestamppb.New(time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)),
		Producer:   "identity", AggregateType: "account", AggregateId: principalID.String(), AggregateVersion: 3,
		IdempotencyKey: "account:4b4d278b:3", Payload: payload,
	}
}

type fakeOutboxStore struct {
	params         store.InsertOutboxEventParams
	calls          int
	err            error
	cancelOnInsert context.CancelFunc
}

func (fake *fakeOutboxStore) InsertOutboxEvent(_ context.Context, params store.InsertOutboxEventParams) error {
	fake.calls++
	if fake.cancelOnInsert != nil {
		fake.cancelOnInsert()
	}
	params.Payload = bytes.Clone(params.Payload)
	fake.params = params
	return fake.err
}

type typedNilOutboxDBTX struct{}

func (*typedNilOutboxDBTX) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	panic("typed nil DBTX called")
}
func (*typedNilOutboxDBTX) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	panic("typed nil DBTX called")
}
func (*typedNilOutboxDBTX) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	panic("typed nil DBTX called")
}

type typedNilOutboxStore struct{}

func (*typedNilOutboxStore) InsertOutboxEvent(context.Context, store.InsertOutboxEventParams) error {
	panic("typed nil outbox store called")
}
