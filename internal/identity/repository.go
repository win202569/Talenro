package identity

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
	"talenro.local/platform/internal/ratelimit"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

var (
	// ErrInvalidApplication reports missing or unsafe application dependencies.
	ErrInvalidApplication = errors.New("identity: invalid application configuration")
	// ErrInvalidRepository reports a nil or malformed PostgreSQL adapter dependency.
	ErrInvalidRepository = errors.New("identity: invalid repository")
	// ErrRepository reports a value-free transaction or persistence failure.
	ErrRepository = errors.New("identity: repository unavailable")
	// ErrEmailDelivery reports malformed provider input without retaining it.
	ErrEmailDelivery = errors.New("identity: invalid email delivery")
	// ErrEmailProvider reports a value-free email provider failure.
	ErrEmailProvider = errors.New("identity: email provider unavailable")
)

// PrincipalID is an opaque account UUID. It cannot be formatted or marshaled accidentally.
type PrincipalID string

// SessionID is an opaque account-session UUID. It cannot be formatted or marshaled accidentally.
type SessionID string

// ReauthMethod is the finite recent-authentication proof class.
type ReauthMethod string

const (
	// ReauthPassword verifies the current password in this Task 9 implementation.
	ReauthPassword ReauthMethod = "password"
	// ReauthPasskey is completed by Task 11.
	ReauthPasskey ReauthMethod = "passkey"
	// ReauthTOTP is completed by Task 11.
	ReauthTOTP ReauthMethod = "totp"
	// ReauthRecoveryCode is completed by Task 11.
	ReauthRecoveryCode ReauthMethod = "recovery_code"
)

// Reauthentication binds one finite proof to the account session it protects.
type Reauthentication struct {
	SessionID SessionID
	Method    ReauthMethod
	Proof     secret.Bytes
}

// RegisterAccountCommand contains the private account-registration input.
type RegisterAccountCommand struct {
	Email          string
	Password       secret.Bytes
	Locale         string
	IdempotencyKey string
}

// RegisterAccountResult is intentionally generic at the unauthenticated boundary.
type RegisterAccountResult struct{ Accepted bool }

// VerifyEmailCommand consumes one opaque email-verification token.
type VerifyEmailCommand struct {
	Token          secret.Bytes
	IdempotencyKey string
}

// CreateEmailVerificationDeliveryCommand rotates a protected verification delivery.
type CreateEmailVerificationDeliveryCommand struct {
	Email          string
	Locale         string
	IdempotencyKey string
}

// CreatePasswordResetDeliveryCommand requests one generic reset delivery.
type CreatePasswordResetDeliveryCommand struct {
	Email          string
	Locale         string
	IdempotencyKey string
}

// ResetPasswordCommand consumes one reset token and binds the replacement session key.
type ResetPasswordCommand struct {
	Email                  string
	Token                  secret.Bytes
	NewPassword            secret.Bytes
	ClientSigningPublicKey [32]byte
	IdempotencyKey         string
}

// EnrollmentGrant transfers one opaque, short-lived device enrollment capability.
type EnrollmentGrant struct {
	Token        secret.Bytes
	ExpiresAt    time.Time
	PolicyMarker string
}

// DeviceEnrollmentAuthority is the minimum redacted identity authority needed to enroll a device.
type DeviceEnrollmentAuthority struct {
	principalID      PrincipalID
	sessionID        SessionID
	policyMarker     string
	provisionalUntil time.Time
}

// NewDeviceEnrollmentAuthority constructs one finite enrollment authority without exposing storage details.
func NewDeviceEnrollmentAuthority(principalID PrincipalID, sessionID SessionID, policyMarker string, provisionalUntil time.Time) (DeviceEnrollmentAuthority, error) {
	if _, err := canonicalIdentityUUID(string(principalID)); err != nil {
		return DeviceEnrollmentAuthority{}, ErrRepository
	}
	if _, err := canonicalIdentityUUID(string(sessionID)); err != nil {
		return DeviceEnrollmentAuthority{}, ErrRepository
	}
	switch policyMarker {
	case "standard":
		if !provisionalUntil.IsZero() {
			return DeviceEnrollmentAuthority{}, ErrRepository
		}
	case "trial_restricted":
		if provisionalUntil.IsZero() {
			return DeviceEnrollmentAuthority{}, ErrRepository
		}
	default:
		return DeviceEnrollmentAuthority{}, ErrRepository
	}
	return DeviceEnrollmentAuthority{
		principalID: principalID, sessionID: sessionID, policyMarker: policyMarker,
		provisionalUntil: provisionalUntil,
	}, nil
}

// PrincipalID returns the redacted principal identifier by value.
func (authority DeviceEnrollmentAuthority) PrincipalID() PrincipalID { return authority.principalID }

// SessionID returns the redacted account-session identifier by value.
func (authority DeviceEnrollmentAuthority) SessionID() SessionID { return authority.sessionID }

// PolicyMarker returns the finite immutable enrollment policy marker.
func (authority DeviceEnrollmentAuthority) PolicyMarker() string { return authority.policyMarker }

// ProvisionalUntil returns the immutable trial boundary when the marker is trial_restricted.
func (authority DeviceEnrollmentAuthority) ProvisionalUntil() (time.Time, bool) {
	return authority.provisionalUntil, !authority.provisionalUntil.IsZero()
}

// CreateEnrollmentGrantCommand binds a grant to a principal, session, and recent proof.
type CreateEnrollmentGrantCommand struct {
	PrincipalID      PrincipalID
	SessionID        SessionID
	Reauthentication Reauthentication
	IdempotencyKey   string
}

// SessionMethod is the frozen account-session authentication method.
type SessionMethod string

const (
	// SessionPassword authenticates with email and password.
	SessionPassword SessionMethod = "password"
	// SessionPasskey authenticates with a discoverable passkey.
	SessionPasskey SessionMethod = "passkey"
)

// CreateSessionCommand is implemented in Task 10/11; its contract is frozen here.
type CreateSessionCommand struct {
	Method                 SessionMethod
	Email                  string
	Password               secret.Bytes
	WebAuthnCeremonyID     string
	WebAuthnResponse       json.RawMessage
	ClientSigningPublicKey [32]byte
	IdempotencyKey         string
}

// SessionTokens are isolated account-domain bearer material.
type SessionTokens struct {
	AccessToken              secret.Bytes
	AccessExpiresAt          time.Time
	RefreshToken             secret.Bytes
	RefreshIdleExpiresAt     time.Time
	RefreshAbsoluteExpiresAt time.Time
}

// RotateSessionCommand is implemented in Task 10; its contract is frozen here.
type RotateSessionCommand struct {
	RefreshToken   secret.Bytes
	ChallengeID    string
	RequestNonce   [32]byte
	Signature      [64]byte
	IdempotencyKey string
}

// CreateSessionChallengeCommand is implemented in Task 10.
type CreateSessionChallengeCommand struct {
	RefreshToken   secret.Bytes
	RequestNonce   [32]byte
	IdempotencyKey string
}

// SessionChallenge is the bounded PoP challenge returned by Task 10.
type SessionChallenge struct {
	ChallengeID string
	Challenge   [32]byte
	ExpiresAt   time.Time
}

// SessionRevokeScope is the finite Task 10 revocation target.
type SessionRevokeScope string

const (
	// RevokeCurrentSession revokes only the named session.
	RevokeCurrentSession SessionRevokeScope = "current"
	// RevokeOtherSessions revokes every other session.
	RevokeOtherSessions SessionRevokeScope = "others"
	// RevokeAllSessions revokes every account session.
	RevokeAllSessions SessionRevokeScope = "all"
)

// RevokeSessionsCommand is implemented in Task 10.
type RevokeSessionsCommand struct {
	PrincipalID    PrincipalID
	Scope          SessionRevokeScope
	SessionID      SessionID
	IdempotencyKey string
}

// ChangePasswordCommand is implemented in Task 10.
type ChangePasswordCommand struct {
	PrincipalID            PrincipalID
	CurrentPassword        secret.Bytes
	NewPassword            secret.Bytes
	Reauthentication       Reauthentication
	ClientSigningPublicKey [32]byte
	IdempotencyKey         string
}

// Application is the frozen account application surface. Tasks 10 and 11
// complete the methods whose command contracts are declared above.
type Application interface {
	RegisterAccount(context.Context, RegisterAccountCommand) (RegisterAccountResult, error)
	CreateEmailVerificationDelivery(context.Context, CreateEmailVerificationDeliveryCommand) (RegisterAccountResult, error)
	VerifyEmail(context.Context, VerifyEmailCommand) error
	CreatePasswordResetDelivery(context.Context, CreatePasswordResetDeliveryCommand) (RegisterAccountResult, error)
	ResetPassword(context.Context, ResetPasswordCommand) (SessionTokens, error)
	ChangePassword(context.Context, ChangePasswordCommand) (SessionTokens, error)
	CreateSession(context.Context, CreateSessionCommand) (SessionTokens, error)
	CreateSessionChallenge(context.Context, CreateSessionChallengeCommand) (SessionChallenge, error)
	RotateSession(context.Context, RotateSessionCommand) (SessionTokens, error)
	RevokeSessions(context.Context, RevokeSessionsCommand) error
	CreateEnrollmentGrant(context.Context, CreateEnrollmentGrantCommand) (EnrollmentGrant, error)
}

// DeviceTransactionParticipant is implemented by identity and used by deviceauth.
type DeviceTransactionParticipant interface {
	ValidateDeviceEnrollment(context.Context, store.DBTX, [32]byte, time.Time) (DeviceEnrollmentAuthority, bool, error)
	BindSessionToAuthorization(context.Context, store.DBTX, SessionID, uuid.UUID, time.Time) error
	RevokeAuthorizationSessions(context.Context, store.DBTX, uuid.UUID, time.Time) error
}

// DeviceAuthorizationParticipant upgrades provisional authorizations inside an identity transaction.
type DeviceAuthorizationParticipant interface {
	ActivateVerifiedPrincipal(context.Context, store.DBTX, PrincipalID, time.Time) error
}

// TemplateID is the finite email template registry.
type TemplateID string

const (
	// VerifyEmailTemplate renders one email-verification token.
	VerifyEmailTemplate TemplateID = "verify_email"
	// ResetPasswordTemplate renders one password-reset token.
	ResetPasswordTemplate TemplateID = "password_reset"
)

// Delivery is created only after a consumer authenticates and decrypts a pending record.
type Delivery struct {
	DeliveryID   uuid.UUID
	TemplateID   TemplateID
	Locale       string
	Recipient    string
	OneTimeToken secret.Bytes
}

// EmailSender accepts only the fixed private delivery schema.
type EmailSender interface {
	Send(context.Context, Delivery) error
}

// ApplicationDependencies are the production-safe boundaries needed by Task 9.
type ApplicationDependencies struct {
	Repository                     Repository
	Protector                      sensitive.Protector
	Random                         securitykit.RandomSource
	Clock                          securitykit.Clock
	Limiter                        ratelimit.Limiter
	ChallengeStore                 ChallengeStore
	RateLimitKey                   secret.Bytes
	Security                       config.SecurityConfig
	DeviceAuthorizationParticipant DeviceAuthorizationParticipant
}

// Repository owns transaction lifecycle but exposes no provider or raw SQL surface.
type Repository interface {
	WithinTransaction(context.Context, func(context.Context, Transaction) error) error
}

// Transaction is the minimal generated-store surface used by Task 9.
type Transaction interface {
	DBTX() store.DBTX
	BeginIdempotency(context.Context, idempotency.Scope, string, []byte, time.Time, time.Time) (idempotency.Record, idempotency.Outcome, error)
	CompleteIdempotency(context.Context, idempotency.Record, int, []byte) error
	LockEmailLookupDigest(context.Context, []byte) error
	FindIdentityByLookupDigest(context.Context, []byte) (store.IdentityEmailIdentity, bool, error)
	FindIdentityByLookupDigestRead(context.Context, []byte) (store.IdentityEmailIdentity, bool, error)
	GetEmailVerificationForUpdate(context.Context, []byte) (store.IdentityEmailIdentity, bool, error)
	GetAccountForUpdate(context.Context, uuid.UUID) (store.IdentityAccount, bool, error)
	GetAccountSessionForUpdate(context.Context, store.GetAccountSessionForUpdateParams) (store.IdentityAccountSession, bool, error)
	GetPasswordCredential(context.Context, uuid.UUID) (store.IdentityPasswordCredential, bool, error)
	GetPasswordResetForUpdate(context.Context, []byte) (store.IdentityPasswordCredential, bool, error)
	LockPrincipalRefreshTokens(context.Context, uuid.UUID) ([]store.IdentityAccountRefreshToken, error)
	ListPrincipalRefreshTokens(context.Context, uuid.UUID) ([]store.IdentityAccountRefreshToken, error)
	LockPrincipalAccountSessions(context.Context, uuid.UUID) ([]store.IdentityAccountSession, error)
	CreateAccount(context.Context, store.CreateAccountParams) error
	CreateEmailIdentity(context.Context, store.CreateEmailIdentityParams) error
	CreatePasswordCredential(context.Context, store.CreatePasswordCredentialParams) error
	InsertSecurityEvent(context.Context, store.InsertSecurityEventParams) error
	ResetEmailVerification(context.Context, store.ResetEmailVerificationParams) (bool, error)
	SetPasswordReset(context.Context, store.SetPasswordResetParams) (bool, error)
	ConsumeEmailVerification(context.Context, store.ConsumeEmailVerificationParams) (store.IdentityEmailIdentity, bool, error)
	ClearPendingEmailDelivery(context.Context, store.ClearPendingEmailDeliveryParams) (int64, error)
	ActivateVerifiedAccount(context.Context, store.ActivateVerifiedAccountParams) (store.IdentityAccount, bool, error)
	ConsumePasswordReset(context.Context, store.ConsumePasswordResetParams) (bool, error)
	MarkPrincipalSessionsReviewRequired(context.Context, store.MarkPrincipalSessionsReviewRequiredParams) (int64, error)
	CreateAccountSession(context.Context, store.CreateAccountSessionParams) error
	InsertAccountRefreshToken(context.Context, store.InsertAccountRefreshTokenParams) error
	CreateEnrollmentGrant(context.Context, store.CreateEnrollmentGrantParams) error
	AppendEvent(context.Context, *eventsv1.EventEnvelope) error
}

// Format redacts the principal identifier for every formatting verb.
func (PrincipalID) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("[REDACTED-PRINCIPAL]"))
}

// LogValue returns the fixed redacted structured-log value.
func (PrincipalID) LogValue() slog.Value { return slog.StringValue("[REDACTED-PRINCIPAL]") }

// MarshalJSON forbids direct principal serialization.
func (PrincipalID) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: principal serialization forbidden")
}

// Format redacts the session identifier for every formatting verb.
func (SessionID) Format(state fmt.State, _ rune) { _, _ = state.Write([]byte("[REDACTED-SESSION]")) }

// LogValue returns the fixed redacted structured-log value.
func (SessionID) LogValue() slog.Value { return slog.StringValue("[REDACTED-SESSION]") }

// MarshalJSON forbids direct session serialization.
func (SessionID) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: session serialization forbidden")
}

// Format redacts every reauthentication field.
func (Reauthentication) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.Reauthentication([REDACTED])"))
}

// LogValue returns the fixed redacted structured-log value.
func (Reauthentication) LogValue() slog.Value {
	return slog.StringValue("identity.Reauthentication([REDACTED])")
}

// MarshalJSON forbids direct reauthentication serialization.
func (Reauthentication) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: reauthentication serialization forbidden")
}

// Format redacts every private delivery field.
func (Delivery) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.Delivery([REDACTED])"))
}

// LogValue returns the fixed redacted structured-log value.
func (Delivery) LogValue() slog.Value { return slog.StringValue("identity.Delivery([REDACTED])") }

// MarshalJSON forbids direct delivery serialization.
func (Delivery) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: delivery serialization forbidden")
}

// Format redacts every enrollment-grant field.
func (EnrollmentGrant) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.EnrollmentGrant([REDACTED])"))
}

// LogValue returns the fixed redacted structured-log value.
func (EnrollmentGrant) LogValue() slog.Value {
	return slog.StringValue("identity.EnrollmentGrant([REDACTED])")
}

// MarshalJSON forbids direct enrollment-grant serialization.
func (EnrollmentGrant) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: grant serialization forbidden")
}

// Format prevents enrollment authority facts from entering diagnostic output.
func (DeviceEnrollmentAuthority) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.DeviceEnrollmentAuthority([REDACTED])"))
}

// LogValue prevents structured logging from reflecting enrollment authority facts.
func (DeviceEnrollmentAuthority) LogValue() slog.Value {
	return slog.StringValue("identity.DeviceEnrollmentAuthority([REDACTED])")
}

// MarshalJSON forbids direct enrollment authority serialization.
func (DeviceEnrollmentAuthority) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: device enrollment authority serialization forbidden")
}

// Format redacts every session-token field.
func (SessionTokens) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.SessionTokens([REDACTED])"))
}

// LogValue returns the fixed redacted structured-log value.
func (SessionTokens) LogValue() slog.Value {
	return slog.StringValue("identity.SessionTokens([REDACTED])")
}

// MarshalJSON forbids direct session-token serialization.
func (SessionTokens) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: session token serialization forbidden")
}

// Format redacts every account-session login field.
func (CreateSessionCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.CreateSessionCommand([REDACTED])"))
}

// LogValue returns a fixed value rather than reflecting private login fields.
func (CreateSessionCommand) LogValue() slog.Value {
	return slog.StringValue("identity.CreateSessionCommand([REDACTED])")
}

// MarshalJSON forbids direct login-command serialization.
func (CreateSessionCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: create session serialization forbidden")
}

// Format redacts every rotation field.
func (RotateSessionCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.RotateSessionCommand([REDACTED])"))
}

// LogValue returns a fixed value rather than reflecting rotation proof material.
func (RotateSessionCommand) LogValue() slog.Value {
	return slog.StringValue("identity.RotateSessionCommand([REDACTED])")
}

// MarshalJSON forbids direct rotation-command serialization.
func (RotateSessionCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: rotate session serialization forbidden")
}

// Format redacts every challenge command field.
func (CreateSessionChallengeCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.CreateSessionChallengeCommand([REDACTED])"))
}

// LogValue returns a fixed value rather than reflecting challenge context.
func (CreateSessionChallengeCommand) LogValue() slog.Value {
	return slog.StringValue("identity.CreateSessionChallengeCommand([REDACTED])")
}

// MarshalJSON forbids direct challenge-command serialization.
func (CreateSessionChallengeCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: session challenge command serialization forbidden")
}

// Format redacts every challenge response field.
func (SessionChallenge) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.SessionChallenge([REDACTED])"))
}

// LogValue returns a fixed value rather than reflecting the nonce.
func (SessionChallenge) LogValue() slog.Value {
	return slog.StringValue("identity.SessionChallenge([REDACTED])")
}

// MarshalJSON forbids direct domain-response serialization.
func (SessionChallenge) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: session challenge serialization forbidden")
}

// Format redacts every revocation field.
func (RevokeSessionsCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.RevokeSessionsCommand([REDACTED])"))
}

// LogValue returns a fixed value rather than reflecting identifiers.
func (RevokeSessionsCommand) LogValue() slog.Value {
	return slog.StringValue("identity.RevokeSessionsCommand([REDACTED])")
}

// MarshalJSON forbids direct revocation-command serialization.
func (RevokeSessionsCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: revoke sessions serialization forbidden")
}

// Format redacts every password-change field.
func (ChangePasswordCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.ChangePasswordCommand([REDACTED])"))
}

// LogValue returns a fixed value rather than reflecting password proofs.
func (ChangePasswordCommand) LogValue() slog.Value {
	return slog.StringValue("identity.ChangePasswordCommand([REDACTED])")
}

// MarshalJSON forbids direct password-change serialization.
func (ChangePasswordCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: change password serialization forbidden")
}

// Format redacts every registration command field.
func (RegisterAccountCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.RegisterAccountCommand([REDACTED])"))
}

// LogValue returns the fixed redacted structured-log value.
func (RegisterAccountCommand) LogValue() slog.Value {
	return slog.StringValue("identity.RegisterAccountCommand([REDACTED])")
}

// MarshalJSON forbids direct registration-command serialization.
func (RegisterAccountCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: registration command serialization forbidden")
}

// Format redacts every verification command field.
func (VerifyEmailCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.VerifyEmailCommand([REDACTED])"))
}

// LogValue returns the fixed redacted structured-log value.
func (VerifyEmailCommand) LogValue() slog.Value {
	return slog.StringValue("identity.VerifyEmailCommand([REDACTED])")
}

// MarshalJSON forbids direct verification-command serialization.
func (VerifyEmailCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: verification command serialization forbidden")
}

// Format redacts every verification-delivery command field.
func (CreateEmailVerificationDeliveryCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.CreateEmailVerificationDeliveryCommand([REDACTED])"))
}

// LogValue returns the fixed redacted structured-log value.
func (CreateEmailVerificationDeliveryCommand) LogValue() slog.Value {
	return slog.StringValue("identity.CreateEmailVerificationDeliveryCommand([REDACTED])")
}

// MarshalJSON forbids direct verification-delivery command serialization.
func (CreateEmailVerificationDeliveryCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: delivery command serialization forbidden")
}

// Format redacts every reset-delivery command field.
func (CreatePasswordResetDeliveryCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.CreatePasswordResetDeliveryCommand([REDACTED])"))
}

// LogValue returns the fixed redacted structured-log value.
func (CreatePasswordResetDeliveryCommand) LogValue() slog.Value {
	return slog.StringValue("identity.CreatePasswordResetDeliveryCommand([REDACTED])")
}

// MarshalJSON forbids direct reset-delivery command serialization.
func (CreatePasswordResetDeliveryCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: reset delivery command serialization forbidden")
}

// Format redacts every password-reset command field.
func (ResetPasswordCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.ResetPasswordCommand([REDACTED])"))
}

// LogValue returns the fixed redacted structured-log value.
func (ResetPasswordCommand) LogValue() slog.Value {
	return slog.StringValue("identity.ResetPasswordCommand([REDACTED])")
}

// MarshalJSON forbids direct password-reset command serialization.
func (ResetPasswordCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: password reset command serialization forbidden")
}

// Format redacts every enrollment-grant command field.
func (CreateEnrollmentGrantCommand) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.CreateEnrollmentGrantCommand([REDACTED])"))
}

// LogValue returns the fixed redacted structured-log value.
func (CreateEnrollmentGrantCommand) LogValue() slog.Value {
	return slog.StringValue("identity.CreateEnrollmentGrantCommand([REDACTED])")
}

// MarshalJSON forbids direct enrollment-grant command serialization.
func (CreateEnrollmentGrantCommand) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: enrollment command serialization forbidden")
}

func canonicalIdentityUUID(value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil || parsed.String() != value {
		return uuid.Nil, ErrRepository
	}
	return parsed, nil
}
