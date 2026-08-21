package deviceauth

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/store"
)

func TestDeviceTokenDomainsNeverAuthenticateAsAccountTokens(t *testing.T) {
	t.Parallel()

	token := secret.NewBytes(bytes.Repeat([]byte{0x61}, 32))
	defer token.Clear()
	deviceAccess := securitykit.DigestToken(securitykit.DeviceAccessToken, token)
	accountAccess := securitykit.DigestToken(securitykit.AccountAccessToken, token)
	deviceRefresh := securitykit.DigestToken(securitykit.DeviceRefreshToken, token)
	accountRefresh := securitykit.DigestToken(securitykit.AccountRefreshToken, token)
	if deviceAccess == accountAccess || deviceRefresh == accountRefresh || deviceAccess == deviceRefresh {
		t.Fatal("opaque token domains collided")
	}
	var _ DeviceAuthenticator = (*Service)(nil)
}

func TestAuthorizeBundleRequiresExactActiveAuthorityAndImmutablePolicy(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 12, 1, 2, 3, 0, time.UTC)
	authorizationID := uuid.MustParse("fa01e838-cab9-414e-8f03-ed0418cd4f20")
	principalID := uuid.MustParse("a6493384-9407-4ad9-b220-7f3b49ef9054")
	deviceID := uuid.MustParse("f353613c-d08f-4141-b287-b37db9fb6f8e")
	row := store.DiscoverDeviceAccessTokenRow{
		FamilyID: uuid.MustParse("0ff820a5-5022-48e6-8867-77761f8e2f07"), AuthorizationID: authorizationID,
		FamilyState: "active", AccessExpiresAt: now.Add(10 * time.Minute), IdleExpiresAt: now.Add(30 * 24 * time.Hour),
		AbsoluteExpiresAt: now.Add(90 * 24 * time.Hour), PrincipalID: principalID, DeviceID: deviceID,
		AuthorizationState: "active", DeviceState: "active", HpkePublicKey: bytes.Repeat([]byte{0x72}, 32), KeyVersion: 4,
	}
	snapshot := store.DeviceauthDevicePolicySnapshot{
		AuthorizationID: authorizationID, SchemaVersion: "device-policy-v1", Policy: json.RawMessage(`{"mode": "standard"}`), CreatedAt: now.Add(-time.Hour),
	}
	authority, ok := bundleAuthorityFromRows(row, snapshot, true, now)
	if !ok || authority.AuthorizationID != authorizationID || authority.PrincipalID != principalID || authority.DeviceID != deviceID ||
		authority.HPKEPublicKey != [32]byte{0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72, 0x72} ||
		authority.DeviceKeyVersion != 4 || authority.PolicySchema != "device-policy-v1" || string(authority.Policy) != `{"mode":"standard"}` {
		t.Fatal("active device authority did not preserve the exact immutable bundle facts")
	}

	for _, mutate := range []func(*store.DiscoverDeviceAccessTokenRow, *store.DeviceauthDevicePolicySnapshot, *bool){
		func(value *store.DiscoverDeviceAccessTokenRow, _ *store.DeviceauthDevicePolicySnapshot, _ *bool) {
			value.FamilyState = "compromised"
		},
		func(value *store.DiscoverDeviceAccessTokenRow, _ *store.DeviceauthDevicePolicySnapshot, _ *bool) {
			value.AuthorizationState = "suspended"
		},
		func(value *store.DiscoverDeviceAccessTokenRow, _ *store.DeviceauthDevicePolicySnapshot, _ *bool) {
			value.DeviceState = "revoked"
		},
		func(value *store.DiscoverDeviceAccessTokenRow, _ *store.DeviceauthDevicePolicySnapshot, _ *bool) {
			value.AccessExpiresAt = now
		},
		func(_ *store.DiscoverDeviceAccessTokenRow, value *store.DeviceauthDevicePolicySnapshot, _ *bool) {
			value.AuthorizationID = uuid.New()
		},
		func(_ *store.DiscoverDeviceAccessTokenRow, value *store.DeviceauthDevicePolicySnapshot, _ *bool) {
			value.SchemaVersion = "other"
		},
		func(_ *store.DiscoverDeviceAccessTokenRow, value *store.DeviceauthDevicePolicySnapshot, _ *bool) {
			value.Policy = json.RawMessage(`{"mode":"other"}`)
		},
		func(_ *store.DiscoverDeviceAccessTokenRow, _ *store.DeviceauthDevicePolicySnapshot, active *bool) {
			*active = false
		},
	} {
		changedRow := row
		changedRow.HpkePublicKey = bytes.Clone(row.HpkePublicKey)
		changedSnapshot := snapshot
		changedSnapshot.Policy = bytes.Clone(snapshot.Policy)
		accountActive := true
		mutate(&changedRow, &changedSnapshot, &accountActive)
		if _, ok := bundleAuthorityFromRows(changedRow, changedSnapshot, accountActive, now); ok {
			t.Fatal("invalid authority state authorized bundle issuance")
		}
	}
}

func TestVerifiedAuthorizationReturnsEffectiveStandardForExactHistoricalTrialSnapshot(t *testing.T) {
	// Mutation caught: requiring a standard snapshot after verification strands
	// the immutable trial snapshot, while returning it leaks obsolete restricted
	// policy instead of the effective verified policy.
	now := time.Date(2026, 8, 12, 1, 2, 3, 0, time.UTC)
	authorizationID := uuid.MustParse("17f2a72b-7fb9-44ba-b20a-b2a631b7ce2c")
	row := store.DiscoverDeviceAccessTokenRow{
		AuthorizationID: authorizationID, PrincipalID: uuid.MustParse("f592632b-e3c9-43dd-8e31-7cc89dc59cf7"), DeviceID: uuid.MustParse("32c657a5-6e21-4f24-870d-c796f1be3c40"),
		FamilyID: uuid.MustParse("a2baf518-5a17-4e59-b53d-15d3fab33d85"), FamilyState: "active", AuthorizationState: "active", DeviceState: "active", KeyVersion: 1,
		HpkePublicKey: bytes.Repeat([]byte{0x72}, 32), AccessExpiresAt: now.Add(time.Minute), IdleExpiresAt: now.Add(time.Hour), AbsoluteExpiresAt: now.Add(2 * time.Hour),
	}
	snapshot := store.DeviceauthDevicePolicySnapshot{
		AuthorizationID: authorizationID, SchemaVersion: "device-policy-v1",
		Policy: json.RawMessage(`{"expires_at":"2026-08-13T01:02:03Z","max_devices":"1","mode":"trial_restricted"}`),
	}
	authority, ok := bundleAuthorityFromRows(row, snapshot, true, now)
	if !ok || string(authority.Policy) != `{"mode":"standard"}` {
		t.Fatalf("verified historical trial policy = %s, valid=%v", authority.Policy, ok)
	}
	row.ProvisionalUntil = sql.NullTime{Time: now.Add(time.Hour), Valid: true}
	if _, ok := bundleAuthorityFromRows(row, snapshot, true, now); ok {
		t.Fatal("active authorization with a non-null provisional marker was accepted")
	}
}

func TestAccessAuthorityRequiresMatchingProvisionalStateAndMarker(t *testing.T) {
	// Mutations caught: trusting only the discovered or only the locked
	// authorization lets hybrid state/version/marker rows authorize a bundle.
	now := time.Date(2026, 8, 12, 1, 2, 3, 0, time.UTC)
	digest := [32]byte{0x41}
	familyID := uuid.MustParse("a2baf518-5a17-4e59-b53d-15d3fab33d85")
	authorizationID := uuid.MustParse("17f2a72b-7fb9-44ba-b20a-b2a631b7ce2c")
	principalID := uuid.MustParse("f592632b-e3c9-43dd-8e31-7cc89dc59cf7")
	deviceID := uuid.MustParse("32c657a5-6e21-4f24-870d-c796f1be3c40")
	deadline := now.Add(time.Hour)
	discovered := store.DiscoverDeviceAccessTokenRow{
		FamilyID: familyID, AuthorizationID: authorizationID, PrincipalID: principalID, DeviceID: deviceID,
		FamilyState: "active", FamilyStateVersion: 1, AuthorizationState: "provisional", AuthorizationStateVersion: 2,
		ProvisionalUntil: sql.NullTime{Time: deadline, Valid: true}, DeviceState: "active", KeyVersion: 1,
		HpkePublicKey: bytes.Repeat([]byte{0x72}, 32), AccessExpiresAt: deadline, IdleExpiresAt: now.Add(30 * 24 * time.Hour), AbsoluteExpiresAt: now.Add(90 * 24 * time.Hour),
	}
	family := store.DeviceauthDeviceTokenFamily{
		ID: familyID, AuthorizationID: authorizationID, State: "active", StateVersion: 1, AccessTokenHash: bytes.Clone(digest[:]),
		AccessExpiresAt: discovered.AccessExpiresAt, IdleExpiresAt: discovered.IdleExpiresAt, AbsoluteExpiresAt: discovered.AbsoluteExpiresAt,
	}
	authorization := store.DeviceauthDeviceAuthorization{
		ID: authorizationID, PrincipalID: principalID, DeviceID: deviceID, State: "provisional", StateVersion: 2,
		ProvisionalUntil: sql.NullTime{Time: deadline, Valid: true},
	}
	device := store.DeviceauthDevice{ID: deviceID, PrincipalID: principalID, State: "active", KeyVersion: 1, HpkePublicKey: bytes.Repeat([]byte{0x72}, 32)}
	refresh := []store.DeviceauthDeviceRefreshToken{{TokenHash: bytes.Repeat([]byte{0x22}, 32), FamilyID: familyID, State: "active"}}
	if !validAccessAuthority(discovered, family, authorization, device, refresh, refresh, digest, now) {
		t.Fatal("exact provisional access authority rejected")
	}
	for _, mutate := range []func(*store.DiscoverDeviceAccessTokenRow, *store.DeviceauthDeviceAuthorization){
		func(row *store.DiscoverDeviceAccessTokenRow, _ *store.DeviceauthDeviceAuthorization) {
			row.ProvisionalUntil.Time = deadline.Add(time.Nanosecond)
		},
		func(_ *store.DiscoverDeviceAccessTokenRow, row *store.DeviceauthDeviceAuthorization) {
			row.ProvisionalUntil = sql.NullTime{}
		},
		func(row *store.DiscoverDeviceAccessTokenRow, _ *store.DeviceauthDeviceAuthorization) {
			row.AuthorizationStateVersion++
		},
		func(row *store.DiscoverDeviceAccessTokenRow, locked *store.DeviceauthDeviceAuthorization) {
			row.AuthorizationState = "active"
			locked.State = "active"
		},
	} {
		changedDiscovered := discovered
		changedAuthorization := authorization
		mutate(&changedDiscovered, &changedAuthorization)
		if validAccessAuthority(changedDiscovered, family, changedAuthorization, device, refresh, refresh, digest, now) {
			t.Fatal("hybrid provisional access authority accepted")
		}
	}
}

func TestAuthorizeBundleAcceptsJSONBNormalizedFinitePolicyAndRejectsExtraFields(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 12, 1, 2, 3, 0, time.UTC)
	authorizationID := uuid.MustParse("fa01e838-cab9-414e-8f03-ed0418cd4f20")
	row := store.DiscoverDeviceAccessTokenRow{AuthorizationID: authorizationID, PrincipalID: uuid.New(), DeviceID: uuid.New(),
		FamilyID: uuid.New(), FamilyState: "active", AuthorizationState: "active", DeviceState: "active", KeyVersion: 1,
		HpkePublicKey: bytes.Repeat([]byte{0x72}, 32), AccessExpiresAt: now.Add(time.Minute), IdleExpiresAt: now.Add(time.Hour), AbsoluteExpiresAt: now.Add(2 * time.Hour)}
	for _, policy := range []string{`{"mode": "standard"}`, `{ "mode" : "standard" }`} {
		snapshot := store.DeviceauthDevicePolicySnapshot{AuthorizationID: authorizationID, SchemaVersion: "device-policy-v1", Policy: json.RawMessage(policy)}
		if _, ok := bundleAuthorityFromRows(row, snapshot, true, now); !ok {
			t.Fatalf("JSONB-normalized finite policy rejected: %q", policy)
		}
	}
	for _, policy := range []string{
		`{"mode":"standard","extra":true}`,
		`{"mode":"standard","mode":"standard"}`,
		`{"mode":"other"}`,
		`{"mode":"standard","max_devices":0}`,
		`{"mode":"standard","max_devices":null}`,
		`{"mode":"standard","expires_at":""}`,
		`{"mode":"standard","expires_at":null}`,
	} {
		snapshot := store.DeviceauthDevicePolicySnapshot{AuthorizationID: authorizationID, SchemaVersion: "device-policy-v1", Policy: json.RawMessage(policy)}
		if _, ok := bundleAuthorityFromRows(row, snapshot, true, now); ok {
			t.Fatalf("non-finite policy accepted: %q", policy)
		}
	}

	provisionalUntil := now.Add(time.Hour)
	provisional := row
	provisional.AuthorizationState = "provisional"
	provisional.ProvisionalUntil = sql.NullTime{Time: provisionalUntil, Valid: true}
	validProvisional := `{"max_devices": "1", "expires_at": "` + provisionalUntil.Format(time.RFC3339) + `", "mode": "trial_restricted"}`
	snapshot := store.DeviceauthDevicePolicySnapshot{AuthorizationID: authorizationID, SchemaVersion: "device-policy-v1", Policy: json.RawMessage(validProvisional)}
	if _, ok := bundleAuthorityFromRows(provisional, snapshot, true, now); !ok {
		t.Fatalf("JSONB-normalized exact provisional policy rejected: %q", validProvisional)
	}
	for _, policy := range []string{
		`{"mode":"trial_restricted","expires_at":"` + provisionalUntil.Format(time.RFC3339) + `"}`,
		`{"mode":"trial_restricted","max_devices":1}`,
		`{"mode":"trial_restricted","expires_at":"` + provisionalUntil.Format(time.RFC3339) + `","max_devices":1}`,
		`{"mode":"trial_restricted","expires_at":null,"max_devices":"1"}`,
		`{"mode":"trial_restricted","expires_at":"` + provisionalUntil.Format(time.RFC3339) + `","max_devices":null}`,
		`{"mode":"trial_restricted","expires_at":"` + provisionalUntil.Format(time.RFC3339) + `","max_devices":"0"}`,
	} {
		snapshot := store.DeviceauthDevicePolicySnapshot{AuthorizationID: authorizationID, SchemaVersion: "device-policy-v1", Policy: json.RawMessage(policy)}
		if _, ok := bundleAuthorityFromRows(provisional, snapshot, true, now); ok {
			t.Fatalf("inexact provisional policy accepted: %q", policy)
		}
	}
}

func TestDeviceAuthenticatorAndBundleAuthorityRedactDiagnosticsAndRejectJSON(t *testing.T) {
	t.Parallel()

	query := AuthorizeBundleQuery{AccessToken: secret.NewBytes(bytes.Repeat([]byte("T"), 32))}
	defer query.AccessToken.Clear()
	authority := BundleAuthority{
		AuthorizationID: uuid.MustParse("fa01e838-cab9-414e-8f03-ed0418cd4f20"),
		PrincipalID:     uuid.MustParse("a6493384-9407-4ad9-b220-7f3b49ef9054"),
		DeviceID:        uuid.MustParse("f353613c-d08f-4141-b287-b37db9fb6f8e"),
		HPKEPublicKey:   [32]byte{0x73}, DeviceKeyVersion: 1, PolicySchema: "device-policy-v1",
		Policy: json.RawMessage(`{"mode":"standard"}`),
	}
	for _, value := range []any{query, authority} {
		rendered := fmt.Sprintf("%+v", value)
		for _, forbidden := range []string{"TTTT", "fa01e838", "a6493384", "f353613c", "device-policy-v1", "standard"} {
			if strings.Contains(rendered, forbidden) {
				t.Fatalf("diagnostics exposed %q in %q", forbidden, rendered)
			}
		}
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("generic JSON serialization succeeded for %T", value)
		}
	}
	var authenticator DeviceAuthenticator = (*Service)(nil)
	if _, err := authenticator.Authenticate(context.Background(), secret.Bytes{}); err == nil || errors.Is(err, context.Canceled) {
		t.Fatalf("empty device credential error = %v", err)
	}
}

func TestAuthorizeBundleUsesDeviceDigestAndRejectsMutableAuthorityStates(t *testing.T) {
	fixture := newTask13Fixture(t)
	authority, err := fixture.application.AuthorizeBundle(context.Background(), AuthorizeBundleQuery{AccessToken: fixture.access})
	if err != nil {
		t.Fatalf("valid device-domain access authorization: %v", err)
	}
	clear(authority.Policy)

	family := fixture.state.families[fixture.familyID]
	originalDigest := bytes.Clone(family.AccessTokenHash)
	accountDigest := securitykit.DigestToken(securitykit.AccountAccessToken, fixture.access)
	family.AccessTokenHash = bytes.Clone(accountDigest[:])
	fixture.state.families[fixture.familyID] = family
	if _, err := fixture.application.AuthorizeBundle(context.Background(), AuthorizeBundleQuery{AccessToken: fixture.access}); publicTask12Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("account-domain digest on device route = %v, want authentication failure", err)
	}
	family.AccessTokenHash = originalDigest
	fixture.state.families[fixture.familyID] = family

	tests := []struct {
		name   string
		mutate func(*task13Fixture)
	}{
		{name: "suspended device", mutate: func(value *task13Fixture) {
			row := value.state.devices[value.deviceID]
			row.State = "suspended"
			value.state.devices[value.deviceID] = row
		}},
		{name: "revoked authorization", mutate: func(value *task13Fixture) {
			row := value.state.authorizations[value.authorizationID]
			row.State = "revoked"
			value.state.authorizations[value.authorizationID] = row
		}},
		{name: "suspended account", mutate: func(value *task13Fixture) { value.identity.accountActive = false }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			denied := newTask13Fixture(t)
			test.mutate(denied)
			if _, err := denied.application.AuthorizeBundle(context.Background(), AuthorizeBundleQuery{AccessToken: denied.access}); publicTask12Code(err) != apierrors.AuthenticationFailed {
				t.Fatalf("invalid mutable authority = %v, want authentication failure", err)
			}
		})
	}
}

func TestProvisionalBundleAuthorityRequiresCurrentGraceProfile(t *testing.T) {
	// Mutation caught: a structurally valid provisional graph must close
	// immediately when the deployment switches from grace to required/disabled.
	for _, mode := range []config.EmailVerificationMode{config.EmailRequired, config.EmailDisabled} {
		t.Run(string(mode), func(t *testing.T) {
			fixture := newTask13Fixture(t)
			deadline := fixedTask12Time.Add(time.Hour)
			authorization := fixture.state.authorizations[fixture.authorizationID]
			authorization.State = "provisional"
			authorization.ProvisionalUntil = sql.NullTime{Time: deadline, Valid: true}
			fixture.state.authorizations[fixture.authorizationID] = authorization
			family := fixture.state.families[fixture.familyID]
			family.AccessExpiresAt = deadline
			fixture.state.families[fixture.familyID] = family
			fixture.state.policies[fixture.authorizationID] = store.DeviceauthDevicePolicySnapshot{
				AuthorizationID: fixture.authorizationID, SchemaVersion: "device-policy-v1",
				Policy: json.RawMessage(`{"expires_at":"2026-08-10T02:02:03Z","max_devices":"1","mode":"trial_restricted"}`),
			}
			fixture.application.security.EmailVerification = mode
			if _, err := fixture.application.AuthorizeBundle(context.Background(), AuthorizeBundleQuery{AccessToken: fixture.access}); publicTask12Code(err) != apierrors.AuthenticationFailed {
				t.Fatalf("closed provisional bundle error = %v", err)
			}
		})
	}
}
