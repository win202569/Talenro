// Package deviceauth implements proof-bound device enrollment and device authorization.
package deviceauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/ratelimit"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

// ChallengeKind is the closed device challenge purpose.
type ChallengeKind string

const (
	// ChallengeRegistration creates a one-time enrollment proof challenge.
	ChallengeRegistration ChallengeKind = "registration"
	// ChallengeRotation is reserved for Task 13 and fails closed in Task 12.
	ChallengeRotation ChallengeKind = "rotation"
)

// ProofInput is the fixed device proof-of-possession transcript input.
type ProofInput struct {
	ProtocolVersion  string
	Challenge        [32]byte
	GrantDigest      [32]byte
	SigningPublicKey [32]byte
	HPKEPublicKey    [32]byte
	Operation        string
	Audience         string
	RequestNonce     [32]byte
}

// CreateChallengeCommand contains registration inputs or the reserved Task 13 rotation credential.
type CreateChallengeCommand struct {
	Kind             ChallengeKind
	EnrollmentGrant  secret.Bytes
	RefreshToken     secret.Bytes
	RequestNonce     [32]byte
	SigningPublicKey [32]byte
	HPKEPublicKey    [32]byte
	IdempotencyKey   string
}

// Challenge is the public one-time PoP challenge.
type Challenge struct {
	ChallengeID string
	Challenge   [32]byte
	ExpiresAt   time.Time
}

// RegisterDeviceCommand consumes one enrollment grant and one proof challenge.
type RegisterDeviceCommand struct {
	EnrollmentGrant  secret.Bytes
	ChallengeID      string
	RequestNonce     [32]byte
	SigningPublicKey [32]byte
	HPKEPublicKey    [32]byte
	DisplayName      string
	Signature        [64]byte
	IdempotencyKey   string
}

// DeviceTokens contains isolated device-domain identifiers and bearer material.
type DeviceTokens struct {
	DeviceID                 uuid.UUID
	AuthorizationID          uuid.UUID
	AccessToken              secret.Bytes
	RefreshToken             secret.Bytes
	AccessExpiresAt          time.Time
	RefreshIdleExpiresAt     time.Time
	RefreshAbsoluteExpiresAt time.Time
}

// RotateDeviceTokenCommand is reserved for Task 13.
type RotateDeviceTokenCommand struct {
	RefreshToken   secret.Bytes
	ChallengeID    string
	RequestNonce   [32]byte
	Signature      [64]byte
	IdempotencyKey string
}

// RevokeDeviceCommand is reserved for Task 13.
type RevokeDeviceCommand struct {
	AccountPrincipal identity.PrincipalID
	DeviceID         uuid.UUID
	Reauthentication identity.Reauthentication
	IdempotencyKey   string
}

// AuthorizeBundleQuery is reserved for Task 13.
type AuthorizeBundleQuery struct{ AccessToken secret.Bytes }

// BundleAuthority is the frozen device-owned trust authority returned by Task 13.
type BundleAuthority struct {
	AuthorizationID  uuid.UUID
	PrincipalID      uuid.UUID
	DeviceID         uuid.UUID
	HPKEPublicKey    [32]byte
	DeviceKeyVersion uint32
	PolicySchema     string
	Policy           json.RawMessage
}

// Application is the frozen device authorization application surface.
type Application interface {
	CreateChallenge(context.Context, CreateChallengeCommand) (Challenge, error)
	RegisterDevice(context.Context, RegisterDeviceCommand) (DeviceTokens, error)
	RotateDeviceToken(context.Context, RotateDeviceTokenCommand) (DeviceTokens, error)
	RevokeDevice(context.Context, RevokeDeviceCommand) error
	AuthorizeBundle(context.Context, AuthorizeBundleQuery) (BundleAuthority, error)
}

// TrustTransactionParticipant is the frozen Task 13 transaction participant surface.
type TrustTransactionParticipant interface {
	AuthorizeBundleInTransaction(context.Context, store.DBTX, AuthorizeBundleQuery) (BundleAuthority, error)
}

// ApplicationDependencies are the bounded collaborators needed for Task 12 enrollment.
type ApplicationDependencies struct {
	Repository          Repository
	IdentityParticipant identity.DeviceTransactionParticipant
	Protector           sensitive.Protector
	Random              securitykit.RandomSource
	Clock               securitykit.Clock
	Limiter             ratelimit.Limiter
	ChallengeStore      ChallengeStore
	RateLimitKey        secret.Bytes
	Security            config.SecurityConfig
}

// Repository owns PostgreSQL transaction lifecycle.
type Repository interface {
	WithinTransaction(context.Context, func(context.Context, Transaction) error) error
}

// Transaction is the exact generated-query-backed surface used by Task 12.
type Transaction interface {
	DBTX() store.DBTX
	BeginIdempotency(context.Context, idempotency.Scope, string, []byte, time.Time, time.Time) (idempotency.Record, idempotency.Outcome, error)
	CompleteIdempotency(context.Context, idempotency.Record, int, []byte) error
	ConsumeEnrollmentGrant(context.Context, [32]byte, uuid.UUID, time.Time) (store.DeviceauthEnrollmentGrant, bool, error)
	CreateDevice(context.Context, store.CreateDeviceParams) error
	CreateDeviceAuthorization(context.Context, store.CreateDeviceAuthorizationParams) error
	CreateDevicePolicySnapshot(context.Context, store.CreateDevicePolicySnapshotParams) error
	CreateDeviceTokenFamily(context.Context, store.CreateDeviceTokenFamilyParams) error
	InsertDeviceRefreshToken(context.Context, store.InsertDeviceRefreshTokenParams) error
	AppendEvent(context.Context, *eventsv1.EventEnvelope) error
}

func redactDeviceauthValue(state fmt.State, name string) {
	_, _ = state.Write([]byte("deviceauth." + name + "([REDACTED])"))
}

// Format redacts proof input from diagnostic formatting.
func (ProofInput) Format(state fmt.State, _ rune) { redactDeviceauthValue(state, "ProofInput") }

// LogValue redacts proof input from structured logs.
func (ProofInput) LogValue() slog.Value { return slog.StringValue("deviceauth.ProofInput([REDACTED])") }

// MarshalJSON forbids direct proof input serialization.
func (ProofInput) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: proof serialization forbidden")
}

// Format redacts challenge commands from diagnostic formatting.
func (CreateChallengeCommand) Format(state fmt.State, _ rune) {
	redactDeviceauthValue(state, "CreateChallengeCommand")
}

// LogValue redacts challenge commands from structured logs.
func (CreateChallengeCommand) LogValue() slog.Value {
	return slog.StringValue("deviceauth.CreateChallengeCommand([REDACTED])")
}

// MarshalJSON forbids direct challenge command serialization.
func (CreateChallengeCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: challenge command serialization forbidden")
}

// Format redacts challenges from diagnostic formatting.
func (Challenge) Format(state fmt.State, _ rune) { redactDeviceauthValue(state, "Challenge") }

// LogValue redacts challenges from structured logs.
func (Challenge) LogValue() slog.Value { return slog.StringValue("deviceauth.Challenge([REDACTED])") }

// MarshalJSON forbids direct challenge serialization.
func (Challenge) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: challenge serialization forbidden")
}

// Format redacts registration commands from diagnostic formatting.
func (RegisterDeviceCommand) Format(state fmt.State, _ rune) {
	redactDeviceauthValue(state, "RegisterDeviceCommand")
}

// LogValue redacts registration commands from structured logs.
func (RegisterDeviceCommand) LogValue() slog.Value {
	return slog.StringValue("deviceauth.RegisterDeviceCommand([REDACTED])")
}

// MarshalJSON forbids direct registration command serialization.
func (RegisterDeviceCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: registration command serialization forbidden")
}

// Format redacts device tokens from diagnostic formatting.
func (DeviceTokens) Format(state fmt.State, _ rune) { redactDeviceauthValue(state, "DeviceTokens") }

// LogValue redacts device tokens from structured logs.
func (DeviceTokens) LogValue() slog.Value {
	return slog.StringValue("deviceauth.DeviceTokens([REDACTED])")
}

// MarshalJSON forbids direct device token serialization.
func (DeviceTokens) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: device token serialization forbidden")
}

// Format redacts reserved rotation commands from diagnostic formatting.
func (RotateDeviceTokenCommand) Format(state fmt.State, _ rune) {
	redactDeviceauthValue(state, "RotateDeviceTokenCommand")
}

// LogValue redacts reserved rotation commands from structured logs.
func (RotateDeviceTokenCommand) LogValue() slog.Value {
	return slog.StringValue("deviceauth.RotateDeviceTokenCommand([REDACTED])")
}

// MarshalJSON forbids direct reserved rotation command serialization.
func (RotateDeviceTokenCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: rotation command serialization forbidden")
}

// Format redacts reserved revocation commands from diagnostic formatting.
func (RevokeDeviceCommand) Format(state fmt.State, _ rune) {
	redactDeviceauthValue(state, "RevokeDeviceCommand")
}

// LogValue redacts reserved revocation commands from structured logs.
func (RevokeDeviceCommand) LogValue() slog.Value {
	return slog.StringValue("deviceauth.RevokeDeviceCommand([REDACTED])")
}

// MarshalJSON forbids direct reserved revocation command serialization.
func (RevokeDeviceCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: revocation command serialization forbidden")
}

// Format redacts reserved authorization queries from diagnostic formatting.
func (AuthorizeBundleQuery) Format(state fmt.State, _ rune) {
	redactDeviceauthValue(state, "AuthorizeBundleQuery")
}

// LogValue redacts reserved authorization queries from structured logs.
func (AuthorizeBundleQuery) LogValue() slog.Value {
	return slog.StringValue("deviceauth.AuthorizeBundleQuery([REDACTED])")
}

// MarshalJSON forbids direct reserved authorization query serialization.
func (AuthorizeBundleQuery) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: authorization query serialization forbidden")
}

// Format redacts bundle authority from diagnostic formatting.
func (BundleAuthority) Format(state fmt.State, _ rune) {
	redactDeviceauthValue(state, "BundleAuthority")
}

// LogValue redacts bundle authority from structured logs.
func (BundleAuthority) LogValue() slog.Value {
	return slog.StringValue("deviceauth.BundleAuthority([REDACTED])")
}

// MarshalJSON forbids direct bundle authority serialization.
func (BundleAuthority) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: bundle authority serialization forbidden")
}

// Format redacts application dependencies from diagnostic formatting.
func (ApplicationDependencies) Format(state fmt.State, _ rune) {
	redactDeviceauthValue(state, "ApplicationDependencies")
}

// LogValue redacts application dependencies from structured logs.
func (ApplicationDependencies) LogValue() slog.Value {
	return slog.StringValue("deviceauth.ApplicationDependencies([REDACTED])")
}

// MarshalJSON forbids direct application dependency serialization.
func (ApplicationDependencies) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: application dependencies serialization forbidden")
}

// Format redacts application state from diagnostic formatting.
func (Service) Format(state fmt.State, _ rune) { redactDeviceauthValue(state, "Service") }

// LogValue redacts application state from structured logs.
func (Service) LogValue() slog.Value { return slog.StringValue("deviceauth.Service([REDACTED])") }

// MarshalJSON forbids direct application state serialization.
func (Service) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: service serialization forbidden")
}

// Format redacts Redis challenge store state from diagnostic formatting.
func (RedisChallengeStore) Format(state fmt.State, _ rune) {
	redactDeviceauthValue(state, "RedisChallengeStore")
}

// LogValue redacts Redis challenge store state from structured logs.
func (RedisChallengeStore) LogValue() slog.Value {
	return slog.StringValue("deviceauth.RedisChallengeStore([REDACTED])")
}

// MarshalJSON forbids direct Redis challenge store serialization.
func (RedisChallengeStore) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: Redis challenge store serialization forbidden")
}

func (preparedRegistration) Format(state fmt.State, _ rune) {
	redactDeviceauthValue(state, "preparedRegistration")
}

func (preparedRegistration) LogValue() slog.Value {
	return slog.StringValue("deviceauth.preparedRegistration([REDACTED])")
}

func (preparedRegistration) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: prepared registration serialization forbidden")
}

func (deviceTokenReplay) Format(state fmt.State, _ rune) {
	redactDeviceauthValue(state, "deviceTokenReplay")
}

func (deviceTokenReplay) LogValue() slog.Value {
	return slog.StringValue("deviceauth.deviceTokenReplay([REDACTED])")
}

func (deviceTokenReplay) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: replay serialization forbidden")
}
