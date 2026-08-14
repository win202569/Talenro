package trust

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	trustv1 "talenro.local/platform/gen/go/talenro/trust/v1"
	"talenro.local/platform/internal/apierrors"
	contractevents "talenro.local/platform/internal/contracts/events"
	"talenro.local/platform/internal/deviceauth"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/store"
	"talenro.local/platform/internal/strictjson"
)

const (
	bundleLifetime             = 24 * time.Hour
	bundleIdempotencyRetention = 24 * time.Hour
	minimumRequestDeadline     = 2 * time.Second
	maximumRequestDeadline     = 10 * time.Second
	issueBundleOperation       = "issue_bundle"
	resolveBundleOperation     = "resolve_bundle"
	acknowledgeBundleOperation = "acknowledge_bundle"
	deviceCredentialDomain     = "device"
	issueResponseStatus        = 200
	resolveResponseStatus      = 200
	acknowledgeResponseStatus  = 204
	maximumReplayBytes         = 64 << 10
)

var (
	// ErrBundleNotFound is the finite immutable-store miss classification.
	ErrBundleNotFound = errors.New("trust: bundle not found")
	// ErrApplication is the fixed trust application failure classification.
	ErrApplication        = errors.New("trust: application failure")
	errTrustSerialization = errors.New("trust: direct serialization forbidden")
)

// IssueCommand requests one exact authorization-bound test bundle.
type IssueCommand struct {
	AuthorizationID uuid.UUID
	AccessToken     secret.Bytes
	TestConfig      TestConfigV1
	IdempotencyKey  string
}

// IssuedBundle is the redacted immutable issuance handle.
type IssuedBundle struct {
	BundleID       uuid.UUID
	BundleVersion  uint64
	Locator        secret.Bytes
	EnvelopeSHA256 [32]byte
}

// ResolveQuery requests the current server-selected test bundle.
type ResolveQuery struct {
	AuthorizationID uuid.UUID
	AccessToken     secret.Bytes
	IdempotencyKey  string
}

// Resolution lists the three immutable sources for one locator.
type Resolution struct {
	Locator        string
	EnvelopeSHA256 string
	Locations      [3]string
	CacheControl   string
}

// AcknowledgeCommand records durable client acceptance without changing the
// highest issued version.
type AcknowledgeCommand struct {
	AuthorizationID uuid.UUID
	AccessToken     secret.Bytes
	BundleID        uuid.UUID
	BundleVersion   uint64
	IdempotencyKey  string
}

// Application is the frozen trust application surface.
type Application interface {
	Issue(context.Context, IssueCommand) (IssuedBundle, error)
	Resolve(context.Context, ResolveQuery) (Resolution, error)
	Acknowledge(context.Context, AcknowledgeCommand) error
}

// TestConfigSource supplies only the fixed server-owned C1.1 test sequence.
type TestConfigSource interface {
	Current(context.Context) (TestConfigV1, error)
}

// Repository owns the trust PostgreSQL transaction lifecycle.
type Repository interface {
	WithinTransaction(context.Context, func(context.Context, Transaction) error) error
}

// Transaction is the generated-query-backed surface used by trust issuance.
type Transaction interface {
	DBTX() store.DBTX
	BeginIdempotency(context.Context, idempotency.Scope, string, []byte, time.Time, time.Time) (transactionIdempotency, error)
	CompleteIdempotency(context.Context, transactionIdempotency, int, []byte) error
	NextBundleVersion(context.Context, uuid.UUID, time.Time) (uint64, error)
	LatestBundle(context.Context, uuid.UUID) (storedBundle, bool, error)
	BundleByID(context.Context, uuid.UUID) (storedBundle, bool, error)
	InsertBundle(context.Context, bundleIssuance) error
	InsertAcknowledgement(context.Context, storedBundle, time.Time) error
	AppendEvent(context.Context, *eventsv1.EventEnvelope) error
}

// ApplicationDependencies are the bounded collaborators for Task 16.
type ApplicationDependencies struct {
	Repository      Repository
	Device          deviceauth.TrustTransactionParticipant
	Signer          *TimeoutConfigSigner
	Metadata        RootMetadataV1
	Random          securitykit.RandomSource
	Clock           securitykit.Clock
	TestConfig      TestConfigSource
	BundleBaseURLs  [3]string
	RequestDeadline time.Duration
}

// Service implements immutable issuance, current resolution, and ack.
type Service struct {
	repository      Repository
	device          deviceauth.TrustTransactionParticipant
	signer          *TimeoutConfigSigner
	metadata        RootMetadataV1
	random          securitykit.RandomSource
	clock           securitykit.Clock
	testConfig      TestConfigSource
	bundleBaseURLs  [3]string
	requestDeadline time.Duration
}

var _ Application = (*Service)(nil)

type transactionIdempotency struct {
	record   idempotency.Record
	outcome  idempotency.Outcome
	key      string
	digest   [32]byte
	status   int
	response []byte
}

type bundleIssuance struct {
	ID              uuid.UUID
	AuthorizationID uuid.UUID
	BundleVersion   uint64
	Locator         [32]byte
	Envelope        []byte
	EnvelopeSHA256  [32]byte
	SignerKeyID     string
	IssuedAt        time.Time
	NotBefore       time.Time
	ExpiresAt       time.Time
}

type storedBundle struct {
	ID              uuid.UUID
	AuthorizationID uuid.UUID
	BundleVersion   uint64
	Locator         secret.Bytes
	EnvelopeSHA256  [32]byte
	IssuedAt        time.Time
	NotBefore       time.Time
	ExpiresAt       time.Time
}

type issuedReplayV1 struct {
	BundleID       string `json:"bundle_id"`
	BundleVersion  string `json:"bundle_version"`
	Locator        string `json:"locator"`
	EnvelopeSHA256 string `json:"envelope_sha256"`
}

type resolutionReplayV1 struct {
	Locator        string    `json:"locator"`
	EnvelopeSHA256 string    `json:"envelope_sha256"`
	Locations      [3]string `json:"locations"`
	CacheControl   string    `json:"cache_control"`
}

type issuedReplayWire issuedReplayV1
type resolutionReplayWire resolutionReplayV1

type issueRequestV1 struct {
	AuthorizationID string       `json:"authorization_id"`
	TestConfig      TestConfigV1 `json:"test_config"`
}

type authorizationRequestV1 struct {
	AuthorizationID string `json:"authorization_id"`
}

type acknowledgeRequestV1 struct {
	AuthorizationID string `json:"authorization_id"`
	BundleID        string `json:"bundle_id"`
	BundleVersion   string `json:"bundle_version"`
}

// NewApplication validates and copies the immutable application dependencies.
func NewApplication(dependencies ApplicationDependencies) (*Service, error) {
	if nilTrustValue(dependencies.Repository) || nilTrustValue(dependencies.Device) || dependencies.Signer == nil ||
		nilTrustValue(dependencies.Random) || nilTrustValue(dependencies.Clock) || nilTrustValue(dependencies.TestConfig) ||
		dependencies.RequestDeadline < minimumRequestDeadline || dependencies.RequestDeadline > maximumRequestDeadline ||
		ValidateRootMetadataV1(dependencies.Metadata) != nil || !validBundleBaseURLs(dependencies.BundleBaseURLs) {
		return nil, ErrInvalidArgument
	}
	return &Service{
		repository:      dependencies.Repository,
		device:          dependencies.Device,
		signer:          dependencies.Signer,
		metadata:        cloneRootMetadata(dependencies.Metadata),
		random:          dependencies.Random,
		clock:           dependencies.Clock,
		testConfig:      dependencies.TestConfig,
		bundleBaseURLs:  dependencies.BundleBaseURLs,
		requestDeadline: dependencies.RequestDeadline,
	}, nil
}

// Issue atomically authorizes, allocates, signs, seals, persists, and records
// one immutable bundle.
func (service *Service) Issue(ctx context.Context, command IssueCommand) (IssuedBundle, error) {
	if !validService(service) || ctx == nil || ctx.Err() != nil || !validIssueCommand(command) {
		return IssuedBundle{}, malformedTrustRequest()
	}
	canonical, err := canonicalTrustedJSON(issueRequestV1{
		AuthorizationID: command.AuthorizationID.String(), TestConfig: command.TestConfig,
	})
	if err != nil {
		return IssuedBundle{}, malformedTrustRequest()
	}
	defer clear(canonical)
	operationContext, cancel := context.WithTimeout(ctx, service.requestDeadline)
	defer cancel()
	var result IssuedBundle
	err = service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, transaction Transaction) error {
		if nilTrustValue(transaction) || nilTrustValue(transaction.DBTX()) {
			return ErrRepository
		}
		now, nowErr := service.now()
		if nowErr != nil {
			return nowErr
		}
		entry, beginErr := beginTrustIdempotency(transactionContext, transaction, command.AuthorizationID, issueBundleOperation,
			command.IdempotencyKey, canonical, now)
		if beginErr != nil {
			return beginErr
		}
		defer clear(entry.response)
		switch entry.outcome {
		case idempotency.Replay:
			replayed, replayErr := decodeIssuedReplay(entry.response)
			if replayErr != nil || entry.status != issueResponseStatus {
				return ErrRepository
			}
			result = replayed
			return nil
		case idempotency.Conflict:
			return idempotencyConflict()
		case idempotency.InProgress:
			return stateConflict()
		case idempotency.Started:
		default:
			return ErrRepository
		}
		authority, authorizeErr := service.authorize(transactionContext, transaction, command.AuthorizationID, command.AccessToken)
		if authorizeErr != nil {
			return authorizeErr
		}
		issued, event, issueErr := service.issueAuthorized(transactionContext, transaction, authority, command.TestConfig, now)
		clear(authority.Policy)
		if issueErr != nil {
			return issueErr
		}
		defer issued.Locator.Clear()
		body, encodeErr := encodeIssuedReplay(issued)
		if encodeErr != nil {
			return ErrApplication
		}
		defer clear(body)
		if transaction.AppendEvent(transactionContext, event) != nil ||
			transaction.CompleteIdempotency(transactionContext, entry, issueResponseStatus, body) != nil {
			return ErrRepository
		}
		result = IssuedBundle{
			BundleID: issued.BundleID, BundleVersion: issued.BundleVersion,
			Locator: issued.Locator.Take(), EnvelopeSHA256: issued.EnvelopeSHA256,
		}
		return nil
	})
	if err != nil {
		result.Locator.Clear()
		return IssuedBundle{}, collapseTrustError(err)
	}
	return result, nil
}

// Resolve authenticates the device, reuses a current unexpired sequence, or
// atomically issues exactly the next server-owned sequence.
func (service *Service) Resolve(ctx context.Context, query ResolveQuery) (Resolution, error) {
	if !validService(service) || ctx == nil || ctx.Err() != nil || !validResolveQuery(query) {
		return Resolution{}, malformedTrustRequest()
	}
	canonical, err := canonicalTrustedJSON(authorizationRequestV1{AuthorizationID: query.AuthorizationID.String()})
	if err != nil {
		return Resolution{}, malformedTrustRequest()
	}
	defer clear(canonical)
	operationContext, cancel := context.WithTimeout(ctx, service.requestDeadline)
	defer cancel()
	var result Resolution
	err = service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, transaction Transaction) error {
		if nilTrustValue(transaction) || nilTrustValue(transaction.DBTX()) {
			return ErrRepository
		}
		now, nowErr := service.now()
		if nowErr != nil {
			return nowErr
		}
		entry, beginErr := beginTrustIdempotency(transactionContext, transaction, query.AuthorizationID, resolveBundleOperation,
			query.IdempotencyKey, canonical, now)
		if beginErr != nil {
			return beginErr
		}
		defer clear(entry.response)
		switch entry.outcome {
		case idempotency.Replay:
			replayed, replayErr := decodeResolutionReplay(entry.response)
			if replayErr != nil || entry.status != resolveResponseStatus {
				return ErrRepository
			}
			result = replayed
			return nil
		case idempotency.Conflict:
			return idempotencyConflict()
		case idempotency.InProgress:
			return stateConflict()
		case idempotency.Started:
		default:
			return ErrRepository
		}
		authority, authorizeErr := service.authorize(transactionContext, transaction, query.AuthorizationID, query.AccessToken)
		if authorizeErr != nil {
			return authorizeErr
		}
		defer clear(authority.Policy)
		current, currentErr := service.testConfig.Current(transactionContext)
		if currentErr != nil || !validTestConfig(current) {
			return ErrApplication
		}
		target, parseErr := strconv.ParseUint(current.Sequence, 10, 64)
		if parseErr != nil || target == 0 || target > math.MaxInt64 {
			return ErrApplication
		}
		latest, found, latestErr := transaction.LatestBundle(transactionContext, query.AuthorizationID)
		if latestErr != nil {
			return ErrRepository
		}
		if found {
			defer latest.Locator.Clear()
		}
		var selected storedBundle
		if found && latest.BundleVersion >= target && !now.Before(latest.NotBefore) && now.Before(latest.ExpiresAt) {
			selected = latest.clone()
		} else {
			if found && target != latest.BundleVersion+1 || !found && target != 1 {
				return stateConflict()
			}
			issued, event, issueErr := service.issueAuthorized(transactionContext, transaction, authority, current, now)
			if issueErr != nil {
				return issueErr
			}
			defer issued.Locator.Clear()
			if transaction.AppendEvent(transactionContext, event) != nil {
				return ErrRepository
			}
			selected = storedBundle{
				ID: issued.BundleID, AuthorizationID: query.AuthorizationID, BundleVersion: issued.BundleVersion,
				Locator: issued.Locator.Take(), EnvelopeSHA256: issued.EnvelopeSHA256,
				IssuedAt: now, NotBefore: now, ExpiresAt: now.Add(bundleLifetime),
			}
		}
		defer selected.Locator.Clear()
		resolution, resolveErr := service.resolution(selected)
		if resolveErr != nil {
			return resolveErr
		}
		body, encodeErr := encodeResolutionReplay(resolution)
		if encodeErr != nil {
			return ErrApplication
		}
		defer clear(body)
		if transaction.CompleteIdempotency(transactionContext, entry, resolveResponseStatus, body) != nil {
			return ErrRepository
		}
		result = resolution
		return nil
	})
	if err != nil {
		return Resolution{}, collapseTrustError(err)
	}
	return result, nil
}

// Acknowledge authenticates and records an existing exact version without
// mutating highest-issued state.
func (service *Service) Acknowledge(ctx context.Context, command AcknowledgeCommand) error {
	if !validService(service) || ctx == nil || ctx.Err() != nil || !validAcknowledgeCommand(command) {
		return malformedTrustRequest()
	}
	canonical, err := canonicalTrustedJSON(acknowledgeRequestV1{
		AuthorizationID: command.AuthorizationID.String(), BundleID: command.BundleID.String(),
		BundleVersion: strconv.FormatUint(command.BundleVersion, 10),
	})
	if err != nil {
		return malformedTrustRequest()
	}
	defer clear(canonical)
	operationContext, cancel := context.WithTimeout(ctx, service.requestDeadline)
	defer cancel()
	err = service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, transaction Transaction) error {
		if nilTrustValue(transaction) || nilTrustValue(transaction.DBTX()) {
			return ErrRepository
		}
		now, nowErr := service.now()
		if nowErr != nil {
			return nowErr
		}
		entry, beginErr := beginTrustIdempotency(transactionContext, transaction, command.AuthorizationID, acknowledgeBundleOperation,
			command.IdempotencyKey, canonical, now)
		if beginErr != nil {
			return beginErr
		}
		defer clear(entry.response)
		switch entry.outcome {
		case idempotency.Replay:
			if entry.status != acknowledgeResponseStatus || len(entry.response) != 0 {
				return ErrRepository
			}
			return nil
		case idempotency.Conflict:
			return idempotencyConflict()
		case idempotency.InProgress:
			return stateConflict()
		case idempotency.Started:
		default:
			return ErrRepository
		}
		authority, authorizeErr := service.authorize(transactionContext, transaction, command.AuthorizationID, command.AccessToken)
		if authorizeErr != nil {
			return authorizeErr
		}
		clear(authority.Policy)
		bundle, found, bundleErr := transaction.BundleByID(transactionContext, command.BundleID)
		if bundleErr != nil {
			return ErrRepository
		}
		defer bundle.Locator.Clear()
		if !found || bundle.AuthorizationID != command.AuthorizationID || bundle.BundleVersion != command.BundleVersion {
			return stateConflict()
		}
		if transaction.InsertAcknowledgement(transactionContext, bundle, now) != nil {
			return ErrRepository
		}
		event, eventErr := service.acknowledgementEvent(bundle, now)
		if eventErr != nil || transaction.AppendEvent(transactionContext, event) != nil ||
			transaction.CompleteIdempotency(transactionContext, entry, acknowledgeResponseStatus, nil) != nil {
			return ErrRepository
		}
		return nil
	})
	return collapseTrustError(err)
}

func (service *Service) issueAuthorized(
	ctx context.Context,
	transaction Transaction,
	authority deviceauth.BundleAuthority,
	testConfig TestConfigV1,
	now time.Time,
) (IssuedBundle, *eventsv1.EventEnvelope, error) {
	version, err := transaction.NextBundleVersion(ctx, authority.AuthorizationID, now)
	if err != nil || version == 0 || version > math.MaxInt64 || testConfig.Sequence != strconv.FormatUint(version, 10) {
		return IssuedBundle{}, nil, stateConflict()
	}
	bundleID, err := randomUUID(service.random)
	if err != nil {
		return IssuedBundle{}, nil, ErrApplication
	}
	eventID, err := randomUUID(service.random)
	if err != nil {
		return IssuedBundle{}, nil, ErrApplication
	}
	var locator [32]byte
	if _, err := io.ReadFull(service.random, locator[:]); err != nil {
		clear(locator[:])
		return IssuedBundle{}, nil, ErrApplication
	}
	defer clear(locator[:])
	notBefore := now
	expiresAt := now.Add(bundleLifetime)
	publicKey, err := SelectSigningKeyForIssue(service.metadata, service.signer.KeyID(), notBefore, expiresAt)
	clear(publicKey)
	if err != nil {
		return IssuedBundle{}, nil, err
	}
	payload := PayloadV1{
		SchemaVersion:  BundleSchemaV1,
		BundleID:       bundleID.String(),
		BundleLocator:  base64.RawURLEncoding.EncodeToString(locator[:]),
		BundleVersion:  strconv.FormatUint(version, 10),
		Audience:       authority.AuthorizationID.String(),
		IssuedAt:       now.Format(time.RFC3339),
		NotBefore:      notBefore.Format(time.RFC3339),
		ExpiresAt:      expiresAt.Format(time.RFC3339),
		PolicySnapshot: bytes.Clone(authority.Policy),
		TestConfig:     testConfig,
	}
	signed, err := SignBundleV1(ctx, payload, service.signer)
	clear(payload.PolicySnapshot)
	if err != nil {
		return IssuedBundle{}, nil, err
	}
	envelope, envelopeDigest, err := Seal(service.random, authority.HPKEPublicKey, locator, signed)
	if err != nil {
		clear(envelope)
		return IssuedBundle{}, nil, err
	}
	defer clear(envelope)
	issuance := bundleIssuance{
		ID: bundleID, AuthorizationID: authority.AuthorizationID, BundleVersion: version,
		Locator: locator, Envelope: bytes.Clone(envelope), EnvelopeSHA256: envelopeDigest,
		SignerKeyID: service.signer.KeyID(), IssuedAt: now, NotBefore: notBefore, ExpiresAt: expiresAt,
	}
	defer issuance.clear()
	if transaction.InsertBundle(ctx, issuance) != nil {
		return IssuedBundle{}, nil, ErrRepository
	}
	payloadBytes, err := contractevents.MarshalPayload(contractevents.BundleIssuedType, &trustv1.BundleIssued{
		AuthorizationId: authority.AuthorizationID.String(), BundleId: bundleID.String(),
		BundleVersion: strconv.FormatUint(version, 10), EnvelopeSha256: hex.EncodeToString(envelopeDigest[:]),
	})
	if err != nil {
		return IssuedBundle{}, nil, ErrApplication
	}
	event := &eventsv1.EventEnvelope{
		EventId: eventID.String(), EventType: contractevents.BundleIssuedType, OccurredAt: timestamppb.New(now),
		Producer: "trust", AggregateType: "bundle", AggregateId: bundleID.String(), AggregateVersion: version,
		IdempotencyKey: eventID.String(), Payload: payloadBytes,
	}
	return IssuedBundle{
		BundleID: bundleID, BundleVersion: version, Locator: secret.NewBytes(locator[:]), EnvelopeSHA256: envelopeDigest,
	}, event, nil
}

func (service *Service) authorize(
	ctx context.Context,
	transaction Transaction,
	authorizationID uuid.UUID,
	accessToken secret.Bytes,
) (deviceauth.BundleAuthority, error) {
	authority, err := service.device.AuthorizeBundleInTransaction(ctx, transaction.DBTX(), deviceauth.AuthorizeBundleQuery{AccessToken: accessToken})
	if err != nil {
		return deviceauth.BundleAuthority{}, collapseTrustError(err)
	}
	if authority.AuthorizationID != authorizationID || authority.PrincipalID == uuid.Nil || authority.DeviceID == uuid.Nil ||
		authority.HPKEPublicKey == [32]byte{} || authority.DeviceKeyVersion == 0 || authority.PolicySchema != "device-policy-v1" ||
		!validDevicePolicyV1(authority.Policy) {
		clear(authority.Policy)
		return deviceauth.BundleAuthority{}, authenticationFailed()
	}
	policy := bytes.Clone(authority.Policy)
	clear(authority.Policy)
	authority.Policy = policy
	return authority, nil
}

func (service *Service) resolution(bundle storedBundle) (Resolution, error) {
	locatorBytes := bundle.Locator.Copy()
	defer clear(locatorBytes)
	if len(locatorBytes) != 32 || bundle.EnvelopeSHA256 == [32]byte{} {
		return Resolution{}, ErrRepository
	}
	locator := base64.RawURLEncoding.EncodeToString(locatorBytes)
	result := Resolution{
		Locator:        locator,
		EnvelopeSHA256: hex.EncodeToString(bundle.EnvelopeSHA256[:]),
		CacheControl:   ImmutableCacheControl,
	}
	for index, baseURL := range service.bundleBaseURLs {
		result.Locations[index] = baseURL + "/b/" + locator
	}
	return result, nil
}

func (service *Service) acknowledgementEvent(bundle storedBundle, now time.Time) (*eventsv1.EventEnvelope, error) {
	eventID, err := randomUUID(service.random)
	if err != nil {
		return nil, ErrApplication
	}
	payload, err := contractevents.MarshalPayload(contractevents.BundleAcknowledgedType, &trustv1.BundleAcknowledged{
		AuthorizationId: bundle.AuthorizationID.String(), BundleId: bundle.ID.String(),
		BundleVersion: strconv.FormatUint(bundle.BundleVersion, 10),
	})
	if err != nil {
		return nil, ErrApplication
	}
	return &eventsv1.EventEnvelope{
		EventId: eventID.String(), EventType: contractevents.BundleAcknowledgedType, OccurredAt: timestamppb.New(now),
		Producer: "trust", AggregateType: "bundle", AggregateId: bundle.ID.String(), AggregateVersion: bundle.BundleVersion,
		IdempotencyKey: eventID.String(), Payload: payload,
	}, nil
}

func beginTrustIdempotency(
	ctx context.Context,
	transaction Transaction,
	authorizationID uuid.UUID,
	operation, key string,
	canonical []byte,
	now time.Time,
) (transactionIdempotency, error) {
	scope, err := idempotency.AuthenticatedScope(authorizationID, deviceCredentialDomain, operation)
	if err != nil {
		return transactionIdempotency{}, malformedTrustRequest()
	}
	return transaction.BeginIdempotency(ctx, scope, key, canonical, now, now.Add(bundleIdempotencyRetention))
}

func encodeIssuedReplay(value IssuedBundle) ([]byte, error) {
	locator := value.Locator.Copy()
	defer clear(locator)
	if value.BundleID == uuid.Nil || value.BundleVersion == 0 || len(locator) != 32 || value.EnvelopeSHA256 == [32]byte{} {
		return nil, ErrApplication
	}
	return canonicalTrustedJSON(issuedReplayWire(issuedReplayV1{
		BundleID: value.BundleID.String(), BundleVersion: strconv.FormatUint(value.BundleVersion, 10),
		Locator: base64.RawURLEncoding.EncodeToString(locator), EnvelopeSHA256: hex.EncodeToString(value.EnvelopeSHA256[:]),
	}))
}

func decodeIssuedReplay(body []byte) (IssuedBundle, error) {
	var wire issuedReplayWire
	if len(body) == 0 || len(body) > maximumReplayBytes || strictjson.Decode(bytes.NewReader(body), maximumReplayBytes, &wire) != nil {
		return IssuedBundle{}, ErrRepository
	}
	value := issuedReplayV1(wire)
	bundleID, err := uuid.Parse(value.BundleID)
	version, versionErr := strconv.ParseUint(value.BundleVersion, 10, 64)
	locator, locatorErr := base64.RawURLEncoding.DecodeString(value.Locator)
	digest, digestErr := hex.DecodeString(value.EnvelopeSHA256)
	if err != nil || bundleID == uuid.Nil || bundleID.String() != value.BundleID || versionErr != nil || version == 0 ||
		locatorErr != nil || len(locator) != 32 || digestErr != nil || len(digest) != 32 {
		clear(locator)
		clear(digest)
		return IssuedBundle{}, ErrRepository
	}
	var envelopeDigest [32]byte
	copy(envelopeDigest[:], digest)
	clear(digest)
	result := IssuedBundle{BundleID: bundleID, BundleVersion: version, Locator: secret.NewBytes(locator), EnvelopeSHA256: envelopeDigest}
	clear(locator)
	return result, nil
}

func encodeResolutionReplay(value Resolution) ([]byte, error) {
	if !validResolution(value) {
		return nil, ErrApplication
	}
	return canonicalTrustedJSON(resolutionReplayWire(value))
}

func decodeResolutionReplay(body []byte) (Resolution, error) {
	var wire resolutionReplayWire
	if len(body) == 0 || len(body) > maximumReplayBytes || strictjson.Decode(bytes.NewReader(body), maximumReplayBytes, &wire) != nil {
		return Resolution{}, ErrRepository
	}
	value := resolutionReplayV1(wire)
	result := Resolution(value)
	if !validResolution(result) {
		return Resolution{}, ErrRepository
	}
	return result, nil
}

func validResolution(value Resolution) bool {
	locator, err := base64.RawURLEncoding.DecodeString(value.Locator)
	defer clear(locator)
	digest, digestErr := hex.DecodeString(value.EnvelopeSHA256)
	defer clear(digest)
	if err != nil || len(locator) != 32 || digestErr != nil || len(digest) != 32 || value.CacheControl != ImmutableCacheControl {
		return false
	}
	for _, location := range value.Locations {
		if !strings.HasSuffix(location, "/b/"+value.Locator) {
			return false
		}
	}
	return true
}

func validIssueCommand(command IssueCommand) bool {
	token := command.AccessToken.Copy()
	defer clear(token)
	_, keyErr := idempotency.KeyDigest(command.IdempotencyKey)
	return command.AuthorizationID != uuid.Nil && len(token) > 0 && keyErr == nil && validTestConfig(command.TestConfig)
}

func validResolveQuery(query ResolveQuery) bool {
	token := query.AccessToken.Copy()
	defer clear(token)
	_, keyErr := idempotency.KeyDigest(query.IdempotencyKey)
	return query.AuthorizationID != uuid.Nil && len(token) > 0 && keyErr == nil
}

func validAcknowledgeCommand(command AcknowledgeCommand) bool {
	token := command.AccessToken.Copy()
	defer clear(token)
	_, keyErr := idempotency.KeyDigest(command.IdempotencyKey)
	return command.AuthorizationID != uuid.Nil && command.BundleID != uuid.Nil && command.BundleVersion > 0 && len(token) > 0 && keyErr == nil
}

func validTestConfig(value TestConfigV1) bool {
	return len(value.Message) <= maximumMessageBytes && utf8.ValidString(value.Message) && positiveIntegerString(value.Sequence)
}

func validBundleBaseURLs(values [3]string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !validBundleBaseURL(value) {
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

func validBundleBaseURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" || parsed.Hostname() == "" || parsed.User != nil || parsed.Opaque != "" ||
		parsed.Path != "" || parsed.RawPath != "" || parsed.RawQuery != "" || parsed.ForceQuery ||
		parsed.Fragment != "" || parsed.RawFragment != "" || strings.HasSuffix(value, "/") || !validBundleURLPort(parsed.Host) {
		return false
	}
	if parsed.Scheme == "https" {
		return true
	}
	if parsed.Scheme != "http" {
		return false
	}
	hostname := parsed.Hostname()
	if strings.EqualFold(hostname, "localhost") {
		return true
	}
	address := net.ParseIP(hostname)
	return address != nil && address.IsLoopback()
}

func validBundleURLPort(host string) bool {
	port := ""
	present := false
	if strings.HasPrefix(host, "[") {
		closing := strings.LastIndexByte(host, ']')
		if closing < 0 {
			return false
		}
		remainder := host[closing+1:]
		if remainder != "" {
			if !strings.HasPrefix(remainder, ":") {
				return false
			}
			present = true
			port = strings.TrimPrefix(remainder, ":")
		}
	} else if colon := strings.LastIndexByte(host, ':'); colon >= 0 {
		present = true
		port = host[colon+1:]
	}
	if !present {
		return true
	}
	number, err := strconv.ParseUint(port, 10, 16)
	return err == nil && number >= 1 && strconv.FormatUint(number, 10) == port
}

func validService(service *Service) bool {
	return service != nil && !nilTrustValue(service.repository) && !nilTrustValue(service.device) && service.signer != nil &&
		!nilTrustValue(service.random) && !nilTrustValue(service.clock) && !nilTrustValue(service.testConfig)
}

func (service *Service) now() (time.Time, error) {
	now := service.clock.Now()
	if now.IsZero() {
		return time.Time{}, ErrApplication
	}
	now = now.UTC().Truncate(time.Second)
	return now, nil
}

func randomUUID(random securitykit.RandomSource) (uuid.UUID, error) {
	if nilTrustValue(random) {
		return uuid.Nil, ErrApplication
	}
	var value [16]byte
	if _, err := io.ReadFull(random, value[:]); err != nil {
		clear(value[:])
		return uuid.Nil, ErrApplication
	}
	value[6] = value[6]&0x0f | 0x40
	value[8] = value[8]&0x3f | 0x80
	result, err := uuid.FromBytes(value[:])
	clear(value[:])
	if err != nil || result == uuid.Nil {
		return uuid.Nil, ErrApplication
	}
	return result, nil
}

func cloneRootMetadata(value RootMetadataV1) RootMetadataV1 {
	value.SigningKeys = append([]SigningKeyMetadataV1(nil), value.SigningKeys...)
	return value
}

func (value bundleIssuance) clone() bundleIssuance {
	value.Envelope = bytes.Clone(value.Envelope)
	return value
}

func (value *bundleIssuance) clear() {
	if value == nil {
		return
	}
	clear(value.Locator[:])
	clear(value.Envelope)
	value.Envelope = nil
}

func (value storedBundle) clone() storedBundle {
	locator := value.Locator.Copy()
	value.Locator = secret.NewBytes(locator)
	clear(locator)
	return value
}

func collapseTrustError(err error) error {
	if err == nil {
		return nil
	}
	var classified apierrors.Error
	if errors.As(err, &classified) {
		return classified
	}
	if errors.Is(err, ErrSignerTimeout) || errors.Is(err, ErrSignerFailure) || errors.Is(err, ErrClosed) ||
		errors.Is(err, ErrKeyUnavailable) {
		return apierrors.New(apierrors.SigningUnavailable, apierrors.Retry)
	}
	if errors.Is(err, ErrInvalidArgument) || errors.Is(err, ErrInvalidPayload) {
		return malformedTrustRequest()
	}
	return apierrors.New(apierrors.DependencyUnavailable, apierrors.Retry)
}

func malformedTrustRequest() error {
	return apierrors.New(apierrors.MalformedRequest, apierrors.ContactSupport)
}
func authenticationFailed() error {
	return apierrors.New(apierrors.AuthenticationFailed, apierrors.Reauthenticate)
}
func idempotencyConflict() error {
	return apierrors.New(apierrors.IdempotencyConflict, apierrors.ContactSupport)
}
func stateConflict() error { return apierrors.New(apierrors.StateConflict, apierrors.ContactSupport) }

func nilTrustValue(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	if !reflected.IsValid() {
		return true
	}
	switch reflected.Kind() { //nolint:exhaustive // The default explicitly rejects every non-nil-capable representation.
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func redactTrustValue(state fmt.State, name string) {
	_, _ = state.Write([]byte("trust." + name + "([REDACTED])"))
}

// Format redacts IssueCommand diagnostics.
func (IssueCommand) Format(state fmt.State, _ rune) { redactTrustValue(state, "IssueCommand") }

// LogValue redacts IssueCommand structured logging.
func (IssueCommand) LogValue() slog.Value { return slog.StringValue("trust.IssueCommand([REDACTED])") }

// MarshalJSON forbids generic IssueCommand serialization.
func (IssueCommand) MarshalJSON() ([]byte, error) { return nil, errTrustSerialization }

// Format redacts IssuedBundle diagnostics.
func (IssuedBundle) Format(state fmt.State, _ rune) { redactTrustValue(state, "IssuedBundle") }

// LogValue redacts IssuedBundle structured logging.
func (IssuedBundle) LogValue() slog.Value { return slog.StringValue("trust.IssuedBundle([REDACTED])") }

// MarshalJSON forbids generic IssuedBundle serialization.
func (IssuedBundle) MarshalJSON() ([]byte, error) { return nil, errTrustSerialization }

// Format redacts ResolveQuery diagnostics.
func (ResolveQuery) Format(state fmt.State, _ rune) { redactTrustValue(state, "ResolveQuery") }

// LogValue redacts ResolveQuery structured logging.
func (ResolveQuery) LogValue() slog.Value { return slog.StringValue("trust.ResolveQuery([REDACTED])") }

// MarshalJSON forbids generic ResolveQuery serialization.
func (ResolveQuery) MarshalJSON() ([]byte, error) { return nil, errTrustSerialization }

// Format redacts Resolution diagnostics.
func (Resolution) Format(state fmt.State, _ rune) { redactTrustValue(state, "Resolution") }

// LogValue redacts Resolution structured logging.
func (Resolution) LogValue() slog.Value { return slog.StringValue("trust.Resolution([REDACTED])") }

// MarshalJSON forbids generic Resolution serialization.
func (Resolution) MarshalJSON() ([]byte, error) { return nil, errTrustSerialization }

// Format redacts AcknowledgeCommand diagnostics.
func (AcknowledgeCommand) Format(state fmt.State, _ rune) {
	redactTrustValue(state, "AcknowledgeCommand")
}

// LogValue redacts AcknowledgeCommand structured logging.
func (AcknowledgeCommand) LogValue() slog.Value {
	return slog.StringValue("trust.AcknowledgeCommand([REDACTED])")
}

// MarshalJSON forbids generic AcknowledgeCommand serialization.
func (AcknowledgeCommand) MarshalJSON() ([]byte, error) { return nil, errTrustSerialization }

// Format redacts ApplicationDependencies diagnostics.
func (ApplicationDependencies) Format(state fmt.State, _ rune) {
	redactTrustValue(state, "ApplicationDependencies")
}

// LogValue redacts ApplicationDependencies structured logging.
func (ApplicationDependencies) LogValue() slog.Value {
	return slog.StringValue("trust.ApplicationDependencies([REDACTED])")
}

// MarshalJSON forbids generic ApplicationDependencies serialization.
func (ApplicationDependencies) MarshalJSON() ([]byte, error) { return nil, errTrustSerialization }

// Format redacts Service diagnostics.
func (*Service) Format(state fmt.State, _ rune) { redactTrustValue(state, "Service") }

// LogValue redacts Service structured logging.
func (*Service) LogValue() slog.Value { return slog.StringValue("trust.Service([REDACTED])") }

// MarshalJSON forbids generic Service serialization.
func (*Service) MarshalJSON() ([]byte, error)         { return nil, errTrustSerialization }
func (issuedReplayV1) Format(state fmt.State, _ rune) { redactTrustValue(state, "issuedReplayV1") }
func (issuedReplayV1) LogValue() slog.Value {
	return slog.StringValue("trust.issuedReplayV1([REDACTED])")
}
func (issuedReplayV1) MarshalJSON() ([]byte, error) { return nil, errTrustSerialization }
func (resolutionReplayV1) Format(state fmt.State, _ rune) {
	redactTrustValue(state, "resolutionReplayV1")
}
func (resolutionReplayV1) LogValue() slog.Value {
	return slog.StringValue("trust.resolutionReplayV1([REDACTED])")
}
func (resolutionReplayV1) MarshalJSON() ([]byte, error) { return nil, errTrustSerialization }
func (bundleIssuance) Format(state fmt.State, _ rune)   { redactTrustValue(state, "bundleIssuance") }
func (bundleIssuance) LogValue() slog.Value {
	return slog.StringValue("trust.bundleIssuance([REDACTED])")
}
func (bundleIssuance) MarshalJSON() ([]byte, error) { return nil, errTrustSerialization }
func (storedBundle) Format(state fmt.State, _ rune) { redactTrustValue(state, "storedBundle") }
func (storedBundle) LogValue() slog.Value           { return slog.StringValue("trust.storedBundle([REDACTED])") }
func (storedBundle) MarshalJSON() ([]byte, error)   { return nil, errTrustSerialization }
func (transactionIdempotency) Format(state fmt.State, _ rune) {
	redactTrustValue(state, "transactionIdempotency")
}
func (transactionIdempotency) LogValue() slog.Value {
	return slog.StringValue("trust.transactionIdempotency([REDACTED])")
}
func (transactionIdempotency) MarshalJSON() ([]byte, error) { return nil, errTrustSerialization }
