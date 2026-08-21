package deviceauth

import (
	"bytes"
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/store"
	"talenro.local/platform/internal/strictjson"
)

// DeviceAuthenticator resolves only device-domain access tokens for device-only routes.
type DeviceAuthenticator interface {
	Authenticate(context.Context, secret.Bytes) (BundleAuthority, error)
}

var _ DeviceAuthenticator = (*Service)(nil)
var _ TrustTransactionParticipant = (*Service)(nil)

// Authenticate resolves one isolated device access token.
func (service *Service) Authenticate(ctx context.Context, accessToken secret.Bytes) (BundleAuthority, error) {
	return service.AuthorizeBundle(ctx, AuthorizeBundleQuery{AccessToken: accessToken})
}

// AuthorizeBundle validates the full device/account authority graph in a deviceauth-owned transaction.
func (service *Service) AuthorizeBundle(ctx context.Context, query AuthorizeBundleQuery) (BundleAuthority, error) {
	if !validDeviceauthService(service) || nilDeviceauthValue(ctx) {
		return BundleAuthority{}, deviceDependencyUnavailable()
	}
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	var authority BundleAuthority
	err := service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, transaction Transaction) error {
		if nilDeviceauthValue(transaction) || nilDeviceauthValue(transaction.DBTX()) {
			return deviceDependencyUnavailable()
		}
		resolved, resolveErr := service.authorizeBundleInDBTX(transactionContext, transaction.DBTX(), query)
		if resolveErr != nil {
			return resolveErr
		}
		authority = resolved
		return nil
	})
	if err != nil {
		clear(authority.Policy)
		return BundleAuthority{}, err
	}
	return authority, nil
}

// AuthorizeBundleInTransaction validates authority in a trust-owned caller transaction.
func (service *Service) AuthorizeBundleInTransaction(
	ctx context.Context,
	dbtx store.DBTX,
	query AuthorizeBundleQuery,
) (BundleAuthority, error) {
	if !validDeviceauthService(service) || nilDeviceauthValue(ctx) || nilDeviceauthValue(dbtx) || ctx.Err() != nil {
		return BundleAuthority{}, deviceDependencyUnavailable()
	}
	return service.authorizeBundleInDBTX(ctx, dbtx, query)
}

func (service *Service) authorizeBundleInDBTX(ctx context.Context, dbtx store.DBTX, query AuthorizeBundleQuery) (BundleAuthority, error) {
	digest := securitykit.DigestToken(securitykit.DeviceAccessToken, query.AccessToken)
	if digest == [32]byte{} {
		return BundleAuthority{}, deviceAuthenticationFailed()
	}
	defer clear(digest[:])
	queries := store.New(dbtx)
	digestCopy := append([]byte(nil), digest[:]...)
	defer clear(digestCopy)
	discovered, err := queries.DiscoverDeviceAccessToken(ctx, digestCopy)
	defer clear(discovered.HpkePublicKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return BundleAuthority{}, deviceAuthenticationFailed()
	}
	if err != nil {
		return BundleAuthority{}, deviceDependencyUnavailable()
	}
	lockedRefresh, err := queries.LockDeviceFamilyRefreshTokens(ctx, discovered.FamilyID)
	defer clearDeviceRefreshRows(lockedRefresh)
	if err != nil {
		return BundleAuthority{}, deviceDependencyUnavailable()
	}
	family, err := queries.GetDeviceTokenFamilyForUpdate(ctx, discovered.FamilyID)
	defer clear(family.AccessTokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return BundleAuthority{}, deviceAuthenticationFailed()
	}
	if err != nil {
		return BundleAuthority{}, deviceDependencyUnavailable()
	}
	authorization, err := queries.GetAuthorizationForUpdate(ctx, discovered.AuthorizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return BundleAuthority{}, deviceAuthenticationFailed()
	}
	if err != nil {
		return BundleAuthority{}, deviceDependencyUnavailable()
	}
	device, err := queries.GetDeviceForUpdate(ctx, discovered.DeviceID)
	defer clearDeviceRowSecrets(&device)
	if errors.Is(err, pgx.ErrNoRows) {
		return BundleAuthority{}, deviceAuthenticationFailed()
	}
	if err != nil {
		return BundleAuthority{}, deviceDependencyUnavailable()
	}
	now, ok := service.now()
	if !ok {
		return BundleAuthority{}, deviceDependencyUnavailable()
	}
	accountActive, err := service.identity.ValidateDeviceAccountAuthority(
		ctx, dbtx, identity.PrincipalID(discovered.PrincipalID.String()), now,
	)
	if err != nil {
		return BundleAuthority{}, deviceDependencyUnavailable()
	}
	freshRefresh, err := queries.ListDeviceFamilyRefreshTokens(ctx, discovered.FamilyID)
	defer clearDeviceRefreshRows(freshRefresh)
	if err != nil {
		return BundleAuthority{}, deviceDependencyUnavailable()
	}
	snapshot, err := queries.GetDevicePolicySnapshot(ctx, discovered.AuthorizationID)
	defer clear(snapshot.Policy)
	if errors.Is(err, pgx.ErrNoRows) {
		return BundleAuthority{}, deviceAuthenticationFailed()
	}
	if err != nil {
		return BundleAuthority{}, deviceDependencyUnavailable()
	}
	if !validAccessAuthority(discovered, family, authorization, device, lockedRefresh, freshRefresh, digest, now) {
		return BundleAuthority{}, deviceAuthenticationFailed()
	}
	if authorization.State == "provisional" && service.security.EmailVerification != config.EmailGrace {
		return BundleAuthority{}, deviceAuthenticationFailed()
	}
	authority, valid := bundleAuthorityFromRows(discovered, snapshot, accountActive, now)
	if !valid {
		return BundleAuthority{}, deviceAuthenticationFailed()
	}
	return authority, nil
}

func validAccessAuthority(
	discovered store.DiscoverDeviceAccessTokenRow,
	family store.DeviceauthDeviceTokenFamily,
	authorization store.DeviceauthDeviceAuthorization,
	device store.DeviceauthDevice,
	lockedRefresh, freshRefresh []store.DeviceauthDeviceRefreshToken,
	digest [32]byte,
	now time.Time,
) bool {
	if len(family.AccessTokenHash) != len(digest) || subtle.ConstantTimeCompare(family.AccessTokenHash, digest[:]) != 1 ||
		discovered.FamilyID == uuid.Nil || discovered.AuthorizationID == uuid.Nil || discovered.PrincipalID == uuid.Nil || discovered.DeviceID == uuid.Nil ||
		family.ID != discovered.FamilyID || family.AuthorizationID != discovered.AuthorizationID || family.State != discovered.FamilyState ||
		family.AccessExpiresAt != discovered.AccessExpiresAt || family.IdleExpiresAt != discovered.IdleExpiresAt || family.AbsoluteExpiresAt != discovered.AbsoluteExpiresAt ||
		authorization.ID != discovered.AuthorizationID || authorization.PrincipalID != discovered.PrincipalID || authorization.DeviceID != discovered.DeviceID ||
		authorization.State != discovered.AuthorizationState || authorization.StateVersion != discovered.AuthorizationStateVersion ||
		!validAuthorizationMarker(discovered.AuthorizationState, discovered.ProvisionalUntil, authorization.ProvisionalUntil, now) ||
		device.ID != discovered.DeviceID || device.PrincipalID != discovered.PrincipalID ||
		device.State != discovered.DeviceState || device.KeyVersion != discovered.KeyVersion || len(device.HpkePublicKey) != 32 ||
		len(discovered.HpkePublicKey) != 32 || subtle.ConstantTimeCompare(device.HpkePublicKey, discovered.HpkePublicKey) != 1 ||
		!sameDeviceRefreshSet(lockedRefresh, freshRefresh) {
		return false
	}
	return family.State == "active" && family.AccessExpiresAt.After(now) && family.IdleExpiresAt.After(now) && family.AbsoluteExpiresAt.After(now) &&
		device.State == "active"
}

func bundleAuthorityFromRows(
	row store.DiscoverDeviceAccessTokenRow,
	snapshot store.DeviceauthDevicePolicySnapshot,
	accountActive bool,
	now time.Time,
) (BundleAuthority, bool) {
	if !accountActive || row.AuthorizationID == uuid.Nil || row.PrincipalID == uuid.Nil || row.DeviceID == uuid.Nil ||
		row.FamilyState != "active" || row.DeviceState != "active" || !row.AccessExpiresAt.After(now) ||
		!row.IdleExpiresAt.After(now) || !row.AbsoluteExpiresAt.After(now) || len(row.HpkePublicKey) != 32 || row.KeyVersion <= 0 ||
		snapshot.AuthorizationID != row.AuthorizationID || snapshot.SchemaVersion != "device-policy-v1" || !validDevicePolicy(row, snapshot.Policy, now) {
		return BundleAuthority{}, false
	}
	if !validAuthorizationMarker(row.AuthorizationState, row.ProvisionalUntil, row.ProvisionalUntil, now) {
		return BundleAuthority{}, false
	}
	var hpke [32]byte
	copy(hpke[:], row.HpkePublicKey)
	return BundleAuthority{
		AuthorizationID: row.AuthorizationID, PrincipalID: row.PrincipalID, DeviceID: row.DeviceID,
		HPKEPublicKey: hpke, DeviceKeyVersion: uint32(row.KeyVersion), PolicySchema: snapshot.SchemaVersion,
		Policy: effectiveDevicePolicy(row, snapshot.Policy),
	}, true
}

func validAuthorizationMarker(state string, discovered, locked sql.NullTime, now time.Time) bool {
	if discovered.Valid != locked.Valid || (discovered.Valid && !discovered.Time.Equal(locked.Time)) {
		return false
	}
	switch state {
	case "active":
		return !discovered.Valid
	case "provisional":
		return discovered.Valid && !discovered.Time.IsZero() && discovered.Time.After(now)
	default:
		return false
	}
}

func effectiveDevicePolicy(row store.DiscoverDeviceAccessTokenRow, policy json.RawMessage) json.RawMessage {
	if row.AuthorizationState == "active" {
		return json.RawMessage(`{"mode":"standard"}`)
	}
	return bytes.Clone(policy)
}

func validDevicePolicy(row store.DiscoverDeviceAccessTokenRow, policy json.RawMessage, now time.Time) bool {
	if !json.Valid(policy) || len(policy) == 0 || len(policy) > 4096 {
		return false
	}
	type finitePolicy struct {
		Mode       string          `json:"mode"`
		ExpiresAt  json.RawMessage `json:"expires_at"`
		MaxDevices json.RawMessage `json:"max_devices"`
	}
	var decoded finitePolicy
	if err := strictjson.Decode(bytes.NewReader(policy), 4096, &decoded); err != nil {
		return false
	}
	if row.AuthorizationState == "active" && decoded.Mode == "standard" {
		return len(decoded.ExpiresAt) == 0 && len(decoded.MaxDevices) == 0
	}
	if row.AuthorizationState != "provisional" && row.AuthorizationState != "active" ||
		len(decoded.ExpiresAt) == 0 || len(decoded.MaxDevices) == 0 {
		return false
	}
	var expiresAt string
	var maxDevices string
	if json.Unmarshal(decoded.ExpiresAt, &expiresAt) != nil || json.Unmarshal(decoded.MaxDevices, &maxDevices) != nil {
		return false
	}
	if decoded.Mode != "trial_restricted" || maxDevices != "1" {
		return false
	}
	parsed, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil || parsed.IsZero() || expiresAt != parsed.UTC().Format(time.RFC3339) {
		return false
	}
	if row.AuthorizationState == "active" {
		return true
	}
	return row.ProvisionalUntil.Valid && row.ProvisionalUntil.Time.After(now) && expiresAt == row.ProvisionalUntil.Time.UTC().Format(time.RFC3339)
}
