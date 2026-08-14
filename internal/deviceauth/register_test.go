package deviceauth

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/ratelimit"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

var fixedTask12Time = time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)

func TestRegisterValidProofCreatesExactStandardGraphAndReplaysBeforeRedis(t *testing.T) {
	t.Parallel()

	fixture := newTask12Fixture(t, "standard", time.Time{})
	command := fixture.registrationCommand(t, "workstation", "task12-register-standard-01")
	tokens, err := fixture.application.RegisterDevice(context.Background(), command)
	if err != nil {
		t.Fatalf("register device: %v", err)
	}
	assertTask12Tokens(t, tokens)
	if fixture.database.deviceCount != 1 || fixture.database.authorizationCount != 1 || fixture.database.familyCount != 1 || fixture.database.refreshCount != 1 {
		t.Fatalf("created graph counts = device %d authorization %d family %d refresh %d", fixture.database.deviceCount, fixture.database.authorizationCount, fixture.database.familyCount, fixture.database.refreshCount)
	}
	if fixture.database.authorization.State != "active" || fixture.database.authorization.ProvisionalUntil.Valid {
		t.Fatalf("standard authorization = %#v", fixture.database.authorization)
	}
	if string(fixture.database.policy.Policy) != `{"mode":"standard"}` {
		t.Fatalf("standard policy = %s", fixture.database.policy.Policy)
	}
	assertTask12FinalOrder(t, fixture.database.operations)
	if fixture.challenges.remoteInsideTransaction.Load() || fixture.limiter.remoteInsideTransaction.Load() {
		t.Fatal("Redis or limiter call occurred while a PostgreSQL transaction was open")
	}

	consumeCalls := fixture.challenges.consumeCalls
	replayed, err := fixture.application.RegisterDevice(context.Background(), command)
	if err != nil {
		t.Fatalf("replay device registration: %v", err)
	}
	if fixture.challenges.consumeCalls != consumeCalls {
		t.Fatal("completed replay touched Redis")
	}
	if !sameTask12Tokens(tokens, replayed) || fixture.database.deviceCount != 1 || fixture.database.authorizationCount != 1 || fixture.database.familyCount != 1 {
		t.Fatal("completed replay did not return the original token response without mutation")
	}
}

func TestRegisterProofObservationOccursOnlyAtEd25519Boundary(t *testing.T) {
	t.Run("success once and completed replay zero", func(t *testing.T) {
		fixture := newTask12Fixture(t, "standard", time.Time{})
		observer := &task18DeviceObserver{}
		observed, err := NewObservedApplication(fixture.application, observer)
		if err != nil {
			t.Fatal(err)
		}
		command := fixture.registrationCommand(t, "observed-register", "task18-register-observed-success")
		tokens, err := observed.RegisterDevice(context.Background(), command)
		if err != nil {
			t.Fatal(err)
		}
		tokens.AccessToken.Clear()
		tokens.RefreshToken.Clear()
		want := []CryptoEvent{{
			Operation: CryptoOperationProofVerify,
			Result:    CryptoResultSuccess,
			Reason:    CryptoReasonNone,
		}}
		if !sameTask18DeviceEvents(observer.events, want) {
			t.Fatalf("registration proof events = %#v, want %#v", observer.events, want)
		}

		replayed, err := observed.RegisterDevice(context.Background(), command)
		if err != nil {
			t.Fatal(err)
		}
		replayed.AccessToken.Clear()
		replayed.RefreshToken.Clear()
		if !sameTask18DeviceEvents(observer.events, want) {
			t.Fatalf("completed replay changed proof events: %#v", observer.events)
		}
	})

	t.Run("malformed request exits before verification", func(t *testing.T) {
		fixture := newTask12Fixture(t, "standard", time.Time{})
		observer := &task18DeviceObserver{}
		observed, err := NewObservedApplication(fixture.application, observer)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := observed.RegisterDevice(context.Background(), RegisterDeviceCommand{}); err == nil {
			t.Fatal("malformed registration unexpectedly succeeded")
		}
		if len(observer.events) != 0 {
			t.Fatalf("pre-verification registration events = %#v, want none", observer.events)
		}
	})

	t.Run("fresh invalid signature records one failure", func(t *testing.T) {
		fixture := newTask12Fixture(t, "standard", time.Time{})
		observer := &task18DeviceObserver{}
		observed, err := NewObservedApplication(fixture.application, observer)
		if err != nil {
			t.Fatal(err)
		}
		command := fixture.registrationCommand(t, "observed-register-invalid", "task18-register-observed-invalid")
		command.Signature[0] ^= 0xff
		if _, err := observed.RegisterDevice(context.Background(), command); publicTask12Code(err) != apierrors.AuthenticationFailed {
			t.Fatalf("invalid registration proof error = %v", err)
		}
		want := []CryptoEvent{{
			Operation: CryptoOperationProofVerify,
			Result:    CryptoResultFailure,
			Reason:    CryptoReasonInvalid,
		}}
		if !sameTask18DeviceEvents(observer.events, want) {
			t.Fatalf("invalid registration proof events = %#v, want %#v", observer.events, want)
		}
	})

	t.Run("post-verification transaction failure retains success", func(t *testing.T) {
		fixture := newTask12Fixture(t, "standard", time.Time{})
		observer := &task18DeviceObserver{}
		observed, err := NewObservedApplication(fixture.application, observer)
		if err != nil {
			t.Fatal(err)
		}
		command := fixture.registrationCommand(t, "observed-register-post", "task18-register-observed-post")
		fixture.repository.failFinalCommit = true
		if _, err := observed.RegisterDevice(context.Background(), command); err == nil {
			t.Fatal("post-verification transaction failure unexpectedly succeeded")
		}
		want := []CryptoEvent{{
			Operation: CryptoOperationProofVerify,
			Result:    CryptoResultSuccess,
			Reason:    CryptoReasonNone,
		}}
		if !sameTask18DeviceEvents(observer.events, want) {
			t.Fatalf("post-verification registration events = %#v, want %#v", observer.events, want)
		}
	})
}

func TestRegisterFinalTransactionReplayTransfersIntactTokensAfterCommit(t *testing.T) {
	fixture := newTask12Fixture(t, "standard", time.Time{})
	expected := DeviceTokens{
		DeviceID: uuid.MustParse("f8412733-a3c2-45fd-a48e-f09c61a8f569"), AuthorizationID: uuid.MustParse("5dff9970-d54f-4273-839f-f560de4371b8"),
		AccessToken: secret.NewBytes(bytes.Repeat([]byte{0xe1}, 32)), RefreshToken: secret.NewBytes(bytes.Repeat([]byte{0xf2}, 32)),
		AccessExpiresAt: fixedTask12Time.Add(10 * time.Minute), RefreshIdleExpiresAt: fixedTask12Time.Add(30 * 24 * time.Hour),
		RefreshAbsoluteExpiresAt: fixedTask12Time.Add(90 * 24 * time.Hour),
	}
	defer expected.AccessToken.Clear()
	defer expected.RefreshToken.Clear()
	replayBody, err := encodeDeviceTokens(expected)
	if err != nil {
		t.Fatal(err)
	}
	defer clear(replayBody)
	fixture.database.finalReplayBody = replayBody
	command := fixture.registrationCommand(t, "final-replay", "task12-register-final-replay")

	result, err := fixture.application.RegisterDevice(context.Background(), command)
	if err != nil {
		t.Fatal(err)
	}
	defer result.AccessToken.Clear()
	defer result.RefreshToken.Clear()
	if !sameTask12Tokens(result, expected) {
		t.Fatal("final-transaction replay token ownership was erased after commit")
	}
	if fixture.database.deviceCount != 0 || fixture.database.authorizationCount != 0 || fixture.database.familyCount != 0 || fixture.database.refreshCount != 0 {
		t.Fatal("final-transaction replay created a second device graph")
	}
}

func TestRegisterGraceCreatesExactImmutableTrialPolicy(t *testing.T) {
	t.Parallel()

	boundary := time.Date(2026, 8, 11, 1, 2, 3, 0, time.UTC)
	fixture := newTask12Fixture(t, "trial_restricted", boundary)
	command := fixture.registrationCommand(t, "tablet", "task12-register-grace-0001")
	if _, err := fixture.application.RegisterDevice(context.Background(), command); err != nil {
		t.Fatalf("register grace device: %v", err)
	}
	if fixture.database.authorization.State != "provisional" || !fixture.database.authorization.ProvisionalUntil.Valid ||
		!fixture.database.authorization.ProvisionalUntil.Time.Equal(boundary) {
		t.Fatalf("trial authorization = %#v", fixture.database.authorization)
	}
	want := `{"expires_at":"2026-08-11T01:02:03Z","max_devices":"1","mode":"trial_restricted"}`
	if string(fixture.database.policy.Policy) != want {
		t.Fatalf("trial policy = %s, want %s", fixture.database.policy.Policy, want)
	}
	for _, forbidden := range []string{"plan", "quota", "node", "node_group", "tablet"} {
		if bytes.Contains(fixture.database.policy.Policy, []byte(forbidden)) {
			t.Fatalf("trial policy contains forbidden field/value %q", forbidden)
		}
	}
}

func TestChallengeAnonymousAndRequiredUnverifiedAccountsFailBeforeRedis(t *testing.T) {
	t.Parallel()

	fixture := newTask12Fixture(t, "standard", time.Time{})
	command := fixture.challengeCommand("task12-anonymous-challenge")
	command.EnrollmentGrant = secret.Bytes{}
	if _, err := fixture.application.CreateChallenge(context.Background(), command); publicTask12Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("anonymous challenge error = %v", err)
	}
	if fixture.challenges.createCalls != 0 || fixture.limiter.calls != 0 {
		t.Fatal("anonymous challenge reached limiter or Redis")
	}

	fixture.participant.found = false
	command = fixture.challengeCommand("task12-required-unverified")
	if _, err := fixture.application.CreateChallenge(context.Background(), command); publicTask12Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("required-unverified challenge error = %v", err)
	}
	if fixture.challenges.createCalls != 0 || fixture.limiter.calls != 0 {
		t.Fatal("unverified challenge reached limiter or Redis")
	}
}

func TestChallengeRegistrationBindsExactPublicContextAndRejectsRotationFields(t *testing.T) {
	t.Parallel()

	fixture := newTask12Fixture(t, "standard", time.Time{})
	command := fixture.challengeCommand("task12-public-context-001")
	challenge, err := fixture.application.CreateChallenge(context.Background(), command)
	if err != nil {
		t.Fatal(err)
	}
	record := fixture.challenges.records[challenge.ChallengeID]
	wantContext, err := registrationContextDigest(command.RequestNonce, command.SigningPublicKey, command.HPKEPublicKey, "https://api.example.test")
	if err != nil || record.ContextDigest != wantContext {
		t.Fatalf("public context digest = %x / %v", record.ContextDigest, err)
	}
	if fixture.limiter.operation != ratelimit.Challenge || fixture.limiter.policy != (config.RateLimitPolicy{Limit: 20, Window: 5 * time.Minute}) {
		t.Fatal("challenge limiter did not receive the fixed operation and policy")
	}

	refresh := secret.NewBytes(bytes.Repeat([]byte{0x44}, 32))
	defer refresh.Clear()
	refreshDigest := securitykit.DigestToken(securitykit.DeviceRefreshToken, refresh)
	familyID := uuid.MustParse("0ff820a5-5022-48e6-8867-77761f8e2f07")
	authorizationID := uuid.MustParse("fa01e838-cab9-414e-8f03-ed0418cd4f20")
	deviceID := uuid.MustParse("f353613c-d08f-4141-b287-b37db9fb6f8e")
	fixture.database.rotationRefresh = store.DiscoverDeviceRefreshTokenRow{
		TokenHash: bytes.Clone(refreshDigest[:]), FamilyID: familyID, RefreshState: "active",
		AuthorizationID: authorizationID, FamilyState: "active", AccessExpiresAt: fixedTask12Time.Add(time.Minute),
		IdleExpiresAt: fixedTask12Time.Add(time.Hour), AbsoluteExpiresAt: fixedTask12Time.Add(24 * time.Hour),
		PrincipalID: uuid.MustParse("a6493384-9407-4ad9-b220-7f3b49ef9054"), DeviceID: deviceID,
		AuthorizationState: "active", DeviceState: "active", SigningPublicKey: bytes.Repeat([]byte{0x71}, 32), KeyVersion: 1,
	}
	fixture.database.rotationFamily = store.DeviceauthDeviceTokenFamily{
		ID: familyID, AuthorizationID: authorizationID, State: "active", AccessExpiresAt: fixedTask12Time.Add(time.Minute),
		IdleExpiresAt: fixedTask12Time.Add(time.Hour), AbsoluteExpiresAt: fixedTask12Time.Add(24 * time.Hour),
	}
	fixture.database.rotationAuthorization = store.DeviceauthDeviceAuthorization{
		ID: authorizationID, PrincipalID: fixture.database.rotationRefresh.PrincipalID, DeviceID: deviceID, State: "active",
	}
	fixture.database.rotationDevice = store.DeviceauthDevice{
		ID: deviceID, PrincipalID: fixture.database.rotationRefresh.PrincipalID, State: "active",
		SigningPublicKey: bytes.Clone(fixture.database.rotationRefresh.SigningPublicKey), KeyVersion: 1,
	}
	fixture.database.rotationLockedRefresh = []store.DeviceauthDeviceRefreshToken{{
		TokenHash: bytes.Clone(refreshDigest[:]), FamilyID: familyID, State: "active", IssuedAt: fixedTask12Time.Add(-time.Minute),
	}}
	rotation := CreateChallengeCommand{
		Kind: ChallengeRotation, RefreshToken: refresh, RequestNonce: [32]byte{0x91}, IdempotencyKey: "task13-rotation-challenge-0001",
	}
	rotationChallenge, err := fixture.application.CreateChallenge(context.Background(), rotation)
	if err != nil {
		t.Fatalf("create rotation challenge: %v", err)
	}
	rotationRecord := fixture.challenges.records[rotationChallenge.ChallengeID]
	wantRotationContext := task13RotationContextDigest(familyID, rotation.RequestNonce, "https://api.example.test")
	if rotationRecord.Kind != ChallengeRotation || rotationRecord.ProtocolVersion != "device-token-rotation-v1" ||
		rotationRecord.Operation != "rotate_device_token" || rotationRecord.GrantDigest != refreshDigest ||
		rotationRecord.ContextDigest != wantRotationContext {
		t.Fatal("rotation challenge did not bind the authoritative family and exact public request context")
	}
	if got := strings.Join(fixture.database.rotationOperations, ","); got != "discover_refresh,lock_refresh,family,authorization,device,account,list_refresh" {
		t.Fatalf("rotation authority order = %q", got)
	}
	if fixture.limiter.remoteInsideTransaction.Load() || fixture.challenges.remoteInsideTransaction.Load() {
		t.Fatal("rotation limiter or Redis ran inside the PostgreSQL transaction")
	}

	malformedRotation := rotation
	malformedRotation.SigningPublicKey = [32]byte{1}
	if _, err := fixture.application.CreateChallenge(context.Background(), malformedRotation); publicTask12Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("rotation with registration key error = %v", err)
	}
}

func task13RotationContextDigest(familyID uuid.UUID, requestNonce [32]byte, audience string) [32]byte {
	material := []byte("TALENRO-DEVICE-ROTATION-CONTEXT-V1\x00")
	for _, part := range [][]byte{[]byte("rotation"), []byte("device-token-rotation-v1"), []byte("rotate_device_token"), []byte(audience)} {
		var length [4]byte
		binary.BigEndian.PutUint32(length[:], uint32(len(part))) // #nosec G115 -- literal test fixtures are bounded.
		material = append(material, length[:]...)
		material = append(material, part...)
	}
	material = append(material, familyID[:]...)
	material = append(material, requestNonce[:]...)
	digest := sha256.Sum256(material)
	clear(material)
	return digest
}

func TestConcurrentRegistrationUsingOneGrantCreatesOneGraph(t *testing.T) {
	fixture := newTask12Fixture(t, "standard", time.Time{})
	commands := [2]RegisterDeviceCommand{
		fixture.registrationCommand(t, "desktop-a", "task12-concurrent-register-a"),
		fixture.registrationCommand(t, "desktop-b", "task12-concurrent-register-b"),
	}

	start := make(chan struct{})
	errorsSeen := make(chan error, len(commands))
	for index := range commands {
		go func(command RegisterDeviceCommand) {
			<-start
			_, err := fixture.application.RegisterDevice(context.Background(), command)
			errorsSeen <- err
		}(commands[index])
	}
	close(start)
	successes := 0
	for range commands {
		if err := <-errorsSeen; err == nil {
			successes++
		}
	}
	if successes != 1 || fixture.database.deviceCount != 1 || fixture.database.authorizationCount != 1 || fixture.database.familyCount != 1 || fixture.database.refreshCount != 1 {
		t.Fatalf("race result successes=%d graph=%d/%d/%d/%d", successes, fixture.database.deviceCount, fixture.database.authorizationCount, fixture.database.familyCount, fixture.database.refreshCount)
	}
}

func TestRegisterPrivateBindingFramesEveryCallerControlledField(t *testing.T) {
	t.Parallel()

	fixture := newTask12Fixture(t, "standard", time.Time{})
	command := fixture.registrationCommand(t, "private-display-name", "task12-private-binding-001")
	grantDigest := securitykit.DigestToken(securitykit.EnrollmentGrantToken, command.EnrollmentGrant)
	canonical, err := privateRegistrationBinding(fixture.protector, grantDigest, command, "https://api.example.test")
	if err != nil {
		t.Fatal(err)
	}
	defer clear(canonical)
	if len(canonical) != 32 {
		t.Fatalf("private binding length = %d, want keyed digest only", len(canonical))
	}
	for _, forbidden := range [][]byte{
		command.EnrollmentGrant.Copy(), command.Signature[:], []byte(command.DisplayName), command.SigningPublicKey[:], command.HPKEPublicKey[:],
	} {
		if bytes.Contains(canonical, forbidden) {
			t.Fatal("database idempotency canonical retained raw private or caller-controlled material")
		}
		clear(forbidden)
	}
}

func TestWithOwnedDisplayNameClearsCallbackBacking(t *testing.T) {
	var retained []byte
	result, err := withOwnedDisplayName("BINDING-DISPLAY-CANARY", func(displayBytes []byte) ([]byte, error) {
		retained = displayBytes
		return bytes.Clone(displayBytes), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer clear(result)
	if string(result) != "BINDING-DISPLAY-CANARY" {
		t.Fatalf("callback result = %q", result)
	}
	if len(retained) == 0 || !allTask12Zero(retained) {
		t.Fatalf("owned display backing survived callback: %q", retained)
	}
}

func TestRegisterPrivateKeyBytesNeverCrossApplicationBoundaries(t *testing.T) {
	t.Parallel()

	fixture := newTask12Fixture(t, "standard", time.Time{})
	command := fixture.registrationCommand(t, "privacy-device", "task12-register-privacy-01")
	privateSigning := bytes.Clone(fixture.signingPrivate)
	privateHPKE := bytes.Clone(fixture.hpkePrivate)
	defer clear(privateSigning)
	defer clear(privateHPKE)
	if _, err := fixture.application.RegisterDevice(context.Background(), command); err != nil {
		t.Fatal(err)
	}
	serialized := fixture.database.capturedBytes()
	defer clear(serialized)
	if bytes.Contains(serialized, privateSigning) || bytes.Contains(serialized, privateHPKE) {
		t.Fatal("repository or event capture contains a client private key")
	}
	for _, value := range []any{command, fixture.database.lastEvent} {
		rendered := fmt.Sprintf("%+v", value)
		if strings.Contains(rendered, string(privateSigning)) || strings.Contains(rendered, string(privateHPKE)) {
			t.Fatal("diagnostic formatting contains a client private key")
		}
	}
	if _, err := json.Marshal(command); err == nil {
		t.Fatal("registration command JSON serialization succeeded")
	}
}

func TestDeviceauthSensitiveValuesRedactAcrossFmtSlogAndJSON(t *testing.T) {
	const (
		grantCanary       = "TOKEN-GRANT-PRIVATE-CANARY"
		displayCanary     = "DISPLAY-NAME-PRIVATE-CANARY"
		idempotencyCanary = "IDEMPOTENCY-PRIVATE-CANARY"
		policyCanary      = `{"policy":"POLICY-PRIVATE-CANARY"}`
		replayCanary      = "REPLAY-TOKEN-PRIVATE-CANARY"
		providerCanary    = "REDIS-RAW-ERROR-CANARY"
	)
	principalID := uuid.MustParse("56975242-a514-457f-b329-6b0c94fefad6")
	deviceID := uuid.MustParse("55f28bc1-efaf-4ee2-a8e6-421b4303a275")
	authorizationID := uuid.MustParse("0e50bdfd-39ee-49b9-b103-9a9f3bb37463")
	publicSigning := [32]byte{}
	publicHPKE := [32]byte{}
	copy(publicSigning[:], bytes.Repeat([]byte{'S'}, 32))
	copy(publicHPKE[:], bytes.Repeat([]byte{'H'}, 32))
	proofInput := ProofInput{
		ProtocolVersion: deviceProofProtocolVersion, Challenge: [32]byte{1}, GrantDigest: [32]byte{2}, SigningPublicKey: publicSigning,
		HPKEPublicKey: publicHPKE, Operation: registerDeviceOperation, Audience: "https://redaction.example.test", RequestNonce: [32]byte{3},
	}
	challengeCommand := CreateChallengeCommand{
		Kind: ChallengeRegistration, EnrollmentGrant: secret.NewBytes([]byte(grantCanary)), RequestNonce: [32]byte{4},
		SigningPublicKey: publicSigning, HPKEPublicKey: publicHPKE, IdempotencyKey: idempotencyCanary,
	}
	defer challengeCommand.EnrollmentGrant.Clear()
	challenge := Challenge{ChallengeID: task12ChallengeID, Challenge: [32]byte{5}, ExpiresAt: fixedTask12Time.Add(time.Minute)}
	registerCommand := RegisterDeviceCommand{
		EnrollmentGrant: secret.NewBytes([]byte(grantCanary)), ChallengeID: task12ChallengeID, RequestNonce: [32]byte{6},
		SigningPublicKey: publicSigning, HPKEPublicKey: publicHPKE, DisplayName: displayCanary, Signature: [64]byte{7},
		IdempotencyKey: idempotencyCanary,
	}
	defer registerCommand.EnrollmentGrant.Clear()
	deviceTokens := DeviceTokens{
		DeviceID: deviceID, AuthorizationID: authorizationID, AccessToken: secret.NewBytes([]byte(grantCanary)),
		RefreshToken: secret.NewBytes([]byte(replayCanary)), AccessExpiresAt: fixedTask12Time.Add(time.Minute),
		RefreshIdleExpiresAt: fixedTask12Time.Add(time.Hour), RefreshAbsoluteExpiresAt: fixedTask12Time.Add(2 * time.Hour),
	}
	defer deviceTokens.AccessToken.Clear()
	defer deviceTokens.RefreshToken.Clear()
	rotateCommand := RotateDeviceTokenCommand{
		RefreshToken: secret.NewBytes([]byte(replayCanary)), ChallengeID: task12ChallengeID, RequestNonce: [32]byte{8},
		Signature: [64]byte{9}, IdempotencyKey: idempotencyCanary,
	}
	defer rotateCommand.RefreshToken.Clear()
	revokeCommand := RevokeDeviceCommand{
		AccountPrincipal: identity.PrincipalID(principalID.String()), DeviceID: deviceID,
		Reauthentication: identity.Reauthentication{
			SessionID: identity.SessionID("a7841655-f6e0-436c-86d6-24349ebda8f2"), Method: identity.ReauthPassword,
			Proof: secret.NewBytes([]byte(grantCanary)),
		},
		IdempotencyKey: idempotencyCanary,
	}
	defer revokeCommand.Reauthentication.Proof.Clear()
	authorizeQuery := AuthorizeBundleQuery{AccessToken: secret.NewBytes([]byte(grantCanary))}
	defer authorizeQuery.AccessToken.Clear()
	bundleAuthority := BundleAuthority{
		AuthorizationID: authorizationID, PrincipalID: principalID, DeviceID: deviceID, HPKEPublicKey: publicHPKE,
		DeviceKeyVersion: 1, PolicySchema: "POLICY-SCHEMA-PRIVATE-CANARY", Policy: json.RawMessage(policyCanary),
	}
	dependencies := ApplicationDependencies{
		RateLimitKey: secret.NewBytes([]byte(grantCanary)), Security: config.SecurityConfig{PublicBaseURL: "https://provider-private.example.test"},
	}
	defer dependencies.RateLimitKey.Clear()
	service := Service{rateLimitKey: secret.NewBytes([]byte(grantCanary)), security: config.SecurityConfig{PublicBaseURL: "https://service-private.example.test"}}
	defer service.rateLimitKey.Clear()
	challengeRecord := task12ChallengeRecord()
	redisStore := RedisChallengeStore{
		executor: &fakeDeviceChallengeExecutor{setError: errors.New(providerCanary), setValue: []byte(grantCanary)}, timeout: 250 * time.Millisecond,
	}
	prepared := preparedRegistration{
		deviceID: deviceID, authorizationID: authorizationID,
		tokens: DeviceTokens{
			DeviceID: deviceID, AuthorizationID: authorizationID, AccessToken: secret.NewBytes([]byte(grantCanary)), RefreshToken: secret.NewBytes([]byte(replayCanary)),
		},
		deviceParams: store.CreateDeviceParams{
			DisplayNameCiphertext: []byte(displayCanary), SigningPublicKey: publicSigning[:], HpkePublicKey: publicHPKE[:],
		},
		policyParams: store.CreateDevicePolicySnapshotParams{Policy: []byte(policyCanary)}, replayBody: []byte(replayCanary),
	}
	defer prepared.clear()
	replay := deviceTokenReplay{
		DeviceID: deviceID.String(), AuthorizationID: authorizationID.String(), AccessToken: grantCanary, RefreshToken: replayCanary,
		AccessExpiresAt: fixedTask12Time.String(), RefreshIdleExpiresAt: fixedTask12Time.Add(time.Hour).String(),
		RefreshAbsoluteExpiresAt: fixedTask12Time.Add(2 * time.Hour).String(),
	}

	subjects := []task12RedactionSubject{
		newTask12RedactionSubject("ProofInput", proofInput, ProofInput{}),
		newTask12RedactionSubject("CreateChallengeCommand", challengeCommand, CreateChallengeCommand{}),
		newTask12RedactionSubject("Challenge", challenge, Challenge{}),
		newTask12RedactionSubject("RegisterDeviceCommand", registerCommand, RegisterDeviceCommand{}),
		newTask12RedactionSubject("DeviceTokens", deviceTokens, DeviceTokens{}),
		newTask12RedactionSubject("RotateDeviceTokenCommand", rotateCommand, RotateDeviceTokenCommand{}),
		newTask12RedactionSubject("RevokeDeviceCommand", revokeCommand, RevokeDeviceCommand{}),
		newTask12RedactionSubject("AuthorizeBundleQuery", authorizeQuery, AuthorizeBundleQuery{}),
		newTask12RedactionSubject("BundleAuthority", bundleAuthority, BundleAuthority{}),
		newTask12RedactionSubject("ApplicationDependencies", dependencies, ApplicationDependencies{}),
		newTask12RedactionSubject("Service", service, Service{}),
		newTask12RedactionSubject("ChallengeRecord", challengeRecord, ChallengeRecord{}),
		newTask12RedactionSubject("RedisChallengeStore", redisStore, RedisChallengeStore{}),
		newTask12RedactionSubject("preparedRegistration", prepared, preparedRegistration{}),
		newTask12RedactionSubject("deviceTokenReplay", replay, deviceTokenReplay{}),
	}
	forbidden := []string{
		grantCanary, displayCanary, idempotencyCanary, policyCanary, replayCanary, providerCanary,
		principalID.String(), deviceID.String(), authorizationID.String(), strings.Repeat("S", 32), strings.Repeat("H", 32),
		"provider-private.example.test", "service-private.example.test",
	}
	assertTask12RedactionSubjects(t, subjects, forbidden)
}

func TestRegisterClearsOwnedDisplayPlaintextAndCiphertextCopies(t *testing.T) {
	t.Parallel()

	fixture := newTask12Fixture(t, "standard", time.Time{})
	tracker := &task12TrackingProtector{delegate: fixture.protector}
	fixture.application.protector = tracker
	command := fixture.registrationCommand(t, "erase-this-display", "task12-register-clear-display")
	if _, err := fixture.application.RegisterDevice(context.Background(), command); err != nil {
		t.Fatal(err)
	}
	if len(tracker.displayPlaintext) == 0 || !allTask12Zero(tracker.displayPlaintext) {
		t.Fatal("owned display-name plaintext copy was not cleared")
	}
	if len(tracker.displayCiphertext) == 0 || !allTask12Zero(tracker.displayCiphertext) {
		t.Fatal("returned display-name ciphertext copy was not cleared after persistence")
	}
}

func TestPreparedRegistrationClearZeroizesFailedCommitTokenOwners(t *testing.T) {
	prepared := preparedRegistration{tokens: DeviceTokens{
		AccessToken:  secret.NewBytes(bytes.Repeat([]byte{0xa7}, 32)),
		RefreshToken: secret.NewBytes(bytes.Repeat([]byte{0xb8}, 32)),
	}}
	accessBackingCanary := prepared.tokens.AccessToken
	refreshBackingCanary := prepared.tokens.RefreshToken
	defer accessBackingCanary.Clear()
	defer refreshBackingCanary.Clear()

	prepared.clear()

	accessOwner := prepared.tokens.AccessToken.Copy()
	refreshOwner := prepared.tokens.RefreshToken.Copy()
	defer clear(accessOwner)
	defer clear(refreshOwner)
	if len(accessOwner) != 0 || len(refreshOwner) != 0 {
		t.Fatal("failed-commit cleanup retained a local token owner")
	}
	accessBacking := accessBackingCanary.Copy()
	refreshBacking := refreshBackingCanary.Copy()
	defer clear(accessBacking)
	defer clear(refreshBacking)
	if len(accessBacking) != 32 || !allTask12Zero(accessBacking) || len(refreshBacking) != 32 || !allTask12Zero(refreshBacking) {
		t.Fatal("failed-commit cleanup did not zero token backing storage")
	}
}

func TestPreparedRegistrationTakeTransfersSuccessfulTokenOwnershipPastDeferredClear(t *testing.T) {
	prepared := preparedRegistration{tokens: DeviceTokens{
		DeviceID: uuid.MustParse("35bb5d71-bb9b-4724-a20c-fae3280b84ae"), AuthorizationID: uuid.MustParse("b7d9a44c-4a7a-45eb-8d89-3c8da16258ec"),
		AccessToken: secret.NewBytes(bytes.Repeat([]byte{0xc9}, 32)), RefreshToken: secret.NewBytes(bytes.Repeat([]byte{0xda}, 32)),
		AccessExpiresAt: fixedTask12Time.Add(10 * time.Minute), RefreshIdleExpiresAt: fixedTask12Time.Add(30 * 24 * time.Hour),
		RefreshAbsoluteExpiresAt: fixedTask12Time.Add(90 * 24 * time.Hour),
	}}

	result := prepared.takeTokens()
	prepared.clear()
	defer result.AccessToken.Clear()
	defer result.RefreshToken.Clear()

	if len(prepared.tokens.AccessToken.Copy()) != 0 || len(prepared.tokens.RefreshToken.Copy()) != 0 {
		t.Fatal("successful transfer retained a prepared token owner")
	}
	access := result.AccessToken.Copy()
	refresh := result.RefreshToken.Copy()
	defer clear(access)
	defer clear(refresh)
	if !bytes.Equal(access, bytes.Repeat([]byte{0xc9}, 32)) || !bytes.Equal(refresh, bytes.Repeat([]byte{0xda}, 32)) {
		t.Fatal("deferred prepared cleanup erased successfully transferred tokens")
	}
	if result.DeviceID != uuid.MustParse("35bb5d71-bb9b-4724-a20c-fae3280b84ae") ||
		result.AuthorizationID != uuid.MustParse("b7d9a44c-4a7a-45eb-8d89-3c8da16258ec") ||
		!result.AccessExpiresAt.Equal(fixedTask12Time.Add(10*time.Minute)) ||
		!result.RefreshIdleExpiresAt.Equal(fixedTask12Time.Add(30*24*time.Hour)) ||
		!result.RefreshAbsoluteExpiresAt.Equal(fixedTask12Time.Add(90*24*time.Hour)) {
		t.Fatal("successful transfer changed public token metadata")
	}
}

func TestRegisterFinalTransactionFailureRollsBackAndClearsOwnedMaterial(t *testing.T) {
	tests := []struct {
		name        string
		panicCommit bool
	}{
		{name: "commit error"},
		{name: "commit panic", panicCommit: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newTask12Fixture(t, "standard", time.Time{})
			fixture.repository.failFinalCommit = true
			fixture.repository.panicFinalCommit = test.panicCommit
			command := fixture.registrationCommand(t, "rollback-owner", "task12-register-rollback-owner")

			tokens, err := fixture.application.RegisterDevice(context.Background(), command)

			if err == nil || len(tokens.AccessToken.Copy()) != 0 || len(tokens.RefreshToken.Copy()) != 0 {
				t.Fatalf("failed final transaction returned tokens or no error: %#v / %v", tokens, err)
			}
			if fixture.database.deviceCount != 0 || fixture.database.authorizationCount != 0 || fixture.database.familyCount != 0 ||
				fixture.database.refreshCount != 0 || fixture.database.grant.State != "unused" {
				t.Fatal("failed final transaction did not restore the database snapshot")
			}
			for name, owned := range map[string][]byte{
				"access digest":  fixture.database.accessDigestOwner,
				"refresh digest": fixture.database.refreshDigestOwner,
				"private replay": fixture.database.replayBodyOwner,
			} {
				if len(owned) == 0 || !allTask12Zero(owned) {
					t.Fatalf("%s survived failed final transaction cleanup", name)
				}
			}
			if fixture.database.operations[len(fixture.database.operations)-1] != "rollback" {
				t.Fatalf("failed final transaction operations = %v", fixture.database.operations)
			}
		})
	}
}

func TestRegisterPostgresTransactionUsesOnlyGeneratedTask4EnrollmentQueries(t *testing.T) {
	t.Parallel()

	principalID := uuid.MustParse("a6493384-9407-4ad9-b220-7f3b49ef9054")
	sessionID := uuid.MustParse("46434e83-f7e2-43c1-b7c5-80a3b49eec14")
	deviceID := uuid.MustParse("83890bd3-f31f-4b88-9f4a-aad77921ee8f")
	authorizationID := uuid.MustParse("b350891f-68af-4e0f-a8b7-b07938354c18")
	familyID := uuid.MustParse("5e51110b-d4f1-4ff5-b4e4-5aa823420883")
	digest := [32]byte{1, 2, 3, 4}
	database := &task12GeneratedDBTX{grant: store.DeviceauthEnrollmentGrant{
		ID: uuid.MustParse("29250900-c6a2-4262-b393-28eb72197248"), PrincipalID: principalID, AccountSessionID: sessionID,
		TokenHash: bytes.Clone(digest[:]), PolicyMarker: "standard", State: "consumed", ExpiresAt: fixedTask12Time.Add(time.Minute),
		ConsumedAt: sql.NullTime{Time: fixedTask12Time, Valid: true}, ConsumedDeviceID: uuid.NullUUID{UUID: deviceID, Valid: true},
	}}
	transaction := &postgresTransaction{queries: store.New(database)}

	if err := transaction.CreateDevice(context.Background(), store.CreateDeviceParams{ID: deviceID, PrincipalID: principalID, DisplayNameCiphertext: bytes.Repeat([]byte{1}, 29), DisplayNameKeyVersion: pgtype.Int4{Int32: 1, Valid: true}, SigningPublicKey: bytes.Repeat([]byte{2}, 32), HpkePublicKey: bytes.Repeat([]byte{3}, 32), KeyVersion: 1, CreatedAt: fixedTask12Time}); err != nil {
		t.Fatal(err)
	}
	grant, found, err := transaction.ConsumeEnrollmentGrant(context.Background(), digest, deviceID, fixedTask12Time)
	if err != nil || !found || grant.ConsumedDeviceID.UUID != deviceID {
		t.Fatalf("consume enrollment grant = (%#v, %v, %v)", grant, found, err)
	}
	if err := transaction.CreateDeviceAuthorization(context.Background(), store.CreateDeviceAuthorizationParams{ID: authorizationID, PrincipalID: principalID, DeviceID: deviceID, State: "active", CreatedAt: fixedTask12Time}); err != nil {
		t.Fatal(err)
	}
	if err := transaction.CreateDevicePolicySnapshot(context.Background(), store.CreateDevicePolicySnapshotParams{AuthorizationID: authorizationID, Policy: json.RawMessage(`{"mode":"standard"}`), CreatedAt: fixedTask12Time}); err != nil {
		t.Fatal(err)
	}
	if err := transaction.CreateDeviceTokenFamily(context.Background(), store.CreateDeviceTokenFamilyParams{ID: familyID, AuthorizationID: authorizationID, AccessTokenHash: bytes.Repeat([]byte{4}, 32), AccessExpiresAt: fixedTask12Time.Add(10 * time.Minute), IdleExpiresAt: fixedTask12Time.Add(30 * 24 * time.Hour), AbsoluteExpiresAt: fixedTask12Time.Add(90 * 24 * time.Hour), CreatedAt: fixedTask12Time}); err != nil {
		t.Fatal(err)
	}
	if err := transaction.InsertDeviceRefreshToken(context.Background(), store.InsertDeviceRefreshTokenParams{TokenHash: bytes.Repeat([]byte{5}, 32), FamilyID: familyID, IssuedAt: fixedTask12Time}); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(database.operations, ","); got != "device,consume,authorization,policy,family,refresh" {
		t.Fatalf("generated operation order = %q", got)
	}
}

func TestRegisterPostgresRepositoryActivatesVerifiedPrincipalThroughCallerDBTX(t *testing.T) {
	t.Parallel()

	principal := identity.PrincipalID("a6493384-9407-4ad9-b220-7f3b49ef9054")
	database := &task12GeneratedDBTX{activateRows: 0}
	var participant identity.DeviceAuthorizationParticipant = &PostgresRepository{}
	if err := participant.ActivateVerifiedPrincipal(context.Background(), database, principal, fixedTask12Time); err != nil {
		t.Fatalf("zero-row activation no-op: %v", err)
	}
	if got := strings.Join(database.operations, ","); got != "activate" {
		t.Fatalf("activation operations = %q", got)
	}

	var typedNil *task12GeneratedDBTX
	if err := participant.ActivateVerifiedPrincipal(context.Background(), typedNil, principal, fixedTask12Time); !errors.Is(err, ErrRepository) {
		t.Fatalf("typed-nil activation error = %v", err)
	}
	if err := participant.ActivateVerifiedPrincipal(context.Background(), panicTask12DBTX{}, principal, fixedTask12Time); !errors.Is(err, ErrRepository) || strings.Contains(err.Error(), "CANARY") {
		t.Fatalf("panic activation error = %v", err)
	}
}

type task12Fixture struct {
	application    *Service
	repository     *task12Repository
	database       *task12Database
	participant    *task12IdentityParticipant
	challenges     *task12ChallengeStore
	limiter        *task12Limiter
	protector      sensitive.Protector
	grant          secret.Bytes
	signingPrivate ed25519.PrivateKey
	hpkePrivate    []byte
	signingPublic  [32]byte
	hpkePublic     [32]byte
}

func newTask12Fixture(t *testing.T, marker string, provisionalUntil time.Time) *task12Fixture {
	t.Helper()
	protector, err := sensitive.NewLocal(
		secret.NewBytes(bytes.Repeat([]byte{0x31}, 32)), secret.NewBytes(bytes.Repeat([]byte{0x32}, 32)), 7,
	)
	if err != nil {
		t.Fatal(err)
	}
	grant := secret.NewBytes(bytes.Repeat([]byte{0x41}, 32))
	grantDigest := securitykit.DigestToken(securitykit.EnrollmentGrantToken, grant)
	principalID := uuid.MustParse("a6493384-9407-4ad9-b220-7f3b49ef9054")
	sessionID := uuid.MustParse("46434e83-f7e2-43c1-b7c5-80a3b49eec14")
	authority, err := identity.NewDeviceEnrollmentAuthority(
		identity.PrincipalID(principalID.String()), identity.SessionID(sessionID.String()), marker, provisionalUntil,
	)
	if err != nil {
		t.Fatal(err)
	}
	database := newTask12Database(protector)
	database.grant = store.DeviceauthEnrollmentGrant{
		ID: uuid.MustParse("29250900-c6a2-4262-b393-28eb72197248"), PrincipalID: principalID, AccountSessionID: sessionID,
		TokenHash: bytes.Clone(grantDigest[:]), PolicyMarker: marker, State: "unused", ExpiresAt: fixedTask12Time.Add(10 * time.Minute),
	}
	if !provisionalUntil.IsZero() {
		database.grant.ProvisionalUntil = sql.NullTime{Time: provisionalUntil, Valid: true}
	}
	repository := &task12Repository{database: database}
	database.repository = repository
	participant := &task12IdentityParticipant{database: database, authority: authority, found: true}
	challenges := &task12ChallengeStore{records: make(map[string]ChallengeRecord), repository: repository}
	limiter := &task12Limiter{allowed: true, repository: repository}
	application, err := NewApplication(ApplicationDependencies{
		Repository: repository, IdentityParticipant: participant, Protector: protector, Random: rand.Reader,
		Clock: task12Clock{}, Limiter: limiter, ChallengeStore: challenges,
		RateLimitKey: secret.NewBytes(bytes.Repeat([]byte{0x33}, 32)),
		Security: config.SecurityConfig{
			Profile: config.ProfileTest, PublicBaseURL: "https://api.example.test", RequestDeadline: 2 * time.Second,
			ChallengeRateLimit: config.RateLimitPolicy{Limit: 20, Window: 5 * time.Minute},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	publicSigning, privateSigning, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	hpkePrivate, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	fixture := &task12Fixture{
		application: application, repository: repository, database: database, participant: participant,
		challenges: challenges, limiter: limiter, protector: protector, grant: grant,
		signingPrivate: privateSigning, hpkePrivate: bytes.Clone(hpkePrivate.Bytes()),
	}
	copy(fixture.signingPublic[:], publicSigning)
	copy(fixture.hpkePublic[:], hpkePrivate.PublicKey().Bytes())
	t.Cleanup(func() {
		clear(fixture.signingPrivate)
		clear(fixture.hpkePrivate)
		clear(database.grant.TokenHash)
		_ = protector.Close()
	})
	return fixture
}

func (fixture *task12Fixture) challengeCommand(key string) CreateChallengeCommand {
	return CreateChallengeCommand{
		Kind: ChallengeRegistration, EnrollmentGrant: fixture.grant, RequestNonce: [32]byte{7, 8, 9},
		SigningPublicKey: fixture.signingPublic, HPKEPublicKey: fixture.hpkePublic, IdempotencyKey: key,
	}
}

func (fixture *task12Fixture) registrationCommand(t *testing.T, displayName, key string) RegisterDeviceCommand {
	t.Helper()
	challenge, err := fixture.application.CreateChallenge(context.Background(), fixture.challengeCommand(key+"-challenge"))
	if err != nil {
		t.Fatal(err)
	}
	grantDigest := securitykit.DigestToken(securitykit.EnrollmentGrantToken, fixture.grant)
	proof := ProofBytes(ProofInput{
		ProtocolVersion: deviceProofProtocolVersion, Challenge: challenge.Challenge, GrantDigest: grantDigest,
		SigningPublicKey: fixture.signingPublic, HPKEPublicKey: fixture.hpkePublic, Operation: registerDeviceOperation,
		Audience: "https://api.example.test", RequestNonce: [32]byte{7, 8, 9},
	})
	if proof == nil {
		t.Fatal("ProofBytes rejected valid registration input")
	}
	signed := ed25519.Sign(fixture.signingPrivate, proof)
	clear(proof)
	var signature [64]byte
	copy(signature[:], signed)
	clear(signed)
	return RegisterDeviceCommand{
		EnrollmentGrant: fixture.grant, ChallengeID: challenge.ChallengeID, RequestNonce: [32]byte{7, 8, 9},
		SigningPublicKey: fixture.signingPublic, HPKEPublicKey: fixture.hpkePublic, DisplayName: displayName,
		Signature: signature, IdempotencyKey: key,
	}
}

type task12Clock struct{}

func (task12Clock) Now() time.Time { return fixedTask12Time }

type task12TrackingProtector struct {
	delegate          sensitive.Protector
	displayPlaintext  []byte
	displayCiphertext []byte
}

func (protector *task12TrackingProtector) LookupDigest(domain string, canonical []byte) [32]byte {
	return protector.delegate.LookupDigest(domain, canonical)
}
func (protector *task12TrackingProtector) Encrypt(domain string, plaintext []byte) (sensitive.EncryptedField, error) {
	field, err := protector.delegate.Encrypt(domain, plaintext)
	if domain == displayNameProtectionDomain {
		protector.displayPlaintext = plaintext
		protector.displayCiphertext = field.Ciphertext
	}
	return field, err
}
func (protector *task12TrackingProtector) Decrypt(domain string, field sensitive.EncryptedField) ([]byte, error) {
	return protector.delegate.Decrypt(domain, field)
}

func allTask12Zero(value []byte) bool {
	for _, element := range value {
		if element != 0 {
			return false
		}
	}
	return true
}

type task12RedactionSubject struct {
	name       string
	value      any
	pointer    any
	zero       any
	nilPointer any
}

func newTask12RedactionSubject[T any](name string, value, zero T) task12RedactionSubject {
	var nilPointer *T
	return task12RedactionSubject{name: name, value: value, pointer: &value, zero: zero, nilPointer: nilPointer}
}

func assertTask12RedactionSubjects(t *testing.T, subjects []task12RedactionSubject, forbidden []string) {
	t.Helper()
	for _, subject := range subjects {
		marker := "deviceauth." + subject.name + "([REDACTED])"
		values := []struct {
			name        string
			value       any
			requireMark bool
		}{
			{name: "value", value: subject.value, requireMark: true},
			{name: "pointer", value: subject.pointer, requireMark: true},
			{name: "zero", value: subject.zero, requireMark: true},
			{name: "nil", value: subject.nilPointer},
		}
		for _, value := range values {
			for _, format := range []string{"%v", "%+v", "%#v", "%s", "%q", "%x", "%X"} {
				rendered := fmt.Sprintf(format, value.value)
				assertTask12NoCanary(t, subject.name+" "+value.name+" "+format, rendered, forbidden)
				if value.requireMark && !strings.Contains(rendered, marker) {
					t.Fatalf("%s %s %s = %q, want fixed marker %q", subject.name, value.name, format, rendered, marker)
				}
			}
			for _, handler := range []struct {
				name string
				new  func(*bytes.Buffer) slog.Handler
			}{
				{name: "text", new: func(buffer *bytes.Buffer) slog.Handler { return slog.NewTextHandler(buffer, nil) }},
				{name: "json", new: func(buffer *bytes.Buffer) slog.Handler { return slog.NewJSONHandler(buffer, nil) }},
			} {
				var output bytes.Buffer
				slog.New(handler.new(&output)).Info("sensitive state", "subject", value.value)
				rendered := output.String()
				assertTask12NoCanary(t, subject.name+" "+value.name+" slog "+handler.name, rendered, forbidden)
				if strings.Contains(rendered, "!PANIC") || strings.Contains(rendered, "!ERROR") {
					t.Fatalf("%s %s slog %s used an error fallback: %q", subject.name, value.name, handler.name, rendered)
				}
				if value.requireMark && !strings.Contains(rendered, marker) {
					t.Fatalf("%s %s slog %s = %q, want fixed marker %q", subject.name, value.name, handler.name, rendered, marker)
				}
			}
		}

		for _, value := range []struct {
			name       string
			value      any
			mustReject bool
		}{
			{name: "value", value: subject.value, mustReject: true},
			{name: "pointer", value: subject.pointer, mustReject: true},
			{name: "zero", value: subject.zero, mustReject: true},
			{name: "nil", value: subject.nilPointer},
		} {
			encoded, err := json.Marshal(value.value)
			assertTask12NoCanary(t, subject.name+" "+value.name+" json", string(encoded), forbidden)
			if value.mustReject && err == nil {
				t.Fatalf("json.Marshal(%s %s) succeeded: %q", subject.name, value.name, encoded)
			}
			if !value.mustReject && (err != nil || string(encoded) != "null") {
				t.Fatalf("json.Marshal(%s nil) = %q, %v, want safe null", subject.name, encoded, err)
			}
		}
	}
}

func assertTask12NoCanary(t *testing.T, context, rendered string, forbidden []string) {
	t.Helper()
	for _, canary := range forbidden {
		if canary != "" && strings.Contains(rendered, canary) {
			t.Fatalf("%s exposed %q in %q", context, canary, rendered)
		}
	}
}

type task12Limiter struct {
	allowed                 bool
	err                     error
	calls                   int
	operation               ratelimit.Operation
	policy                  config.RateLimitPolicy
	repository              *task12Repository
	remoteInsideTransaction atomic.Bool
}

func (limiter *task12Limiter) Allow(_ context.Context, operation ratelimit.Operation, _ [32]byte, policy config.RateLimitPolicy) (bool, error) {
	if limiter.repository.inTransaction.Load() {
		limiter.remoteInsideTransaction.Store(true)
	}
	limiter.calls++
	limiter.operation = operation
	limiter.policy = policy
	return limiter.allowed, limiter.err
}

type task12ChallengeStore struct {
	mu                      sync.Mutex
	records                 map[string]ChallengeRecord
	repository              *task12Repository
	createCalls             int
	consumeCalls            int
	remoteInsideTransaction atomic.Bool
}

func (challengeStore *task12ChallengeStore) Create(_ context.Context, record ChallengeRecord, ttl time.Duration) error {
	challengeStore.mu.Lock()
	defer challengeStore.mu.Unlock()
	if challengeStore.repository.inTransaction.Load() {
		challengeStore.remoteInsideTransaction.Store(true)
	}
	challengeStore.createCalls++
	if ttl != 2*time.Minute {
		return ErrInvalidChallenge
	}
	if _, exists := challengeStore.records[record.ChallengeID]; exists {
		return ErrChallengeUnavailable
	}
	challengeStore.records[record.ChallengeID] = record
	return nil
}

func (challengeStore *task12ChallengeStore) Consume(_ context.Context, challengeID string, grantDigest, contextDigest [32]byte) (ChallengeRecord, error) {
	challengeStore.mu.Lock()
	defer challengeStore.mu.Unlock()
	if challengeStore.repository.inTransaction.Load() {
		challengeStore.remoteInsideTransaction.Store(true)
	}
	challengeStore.consumeCalls++
	record, exists := challengeStore.records[challengeID]
	if !exists {
		return ChallengeRecord{}, ErrChallengeNotFound
	}
	delete(challengeStore.records, challengeID)
	if !challengeDigestsMatch(record, grantDigest, contextDigest) || !record.ExpiresAt.After(fixedTask12Time) {
		return ChallengeRecord{}, ErrChallengeNotFound
	}
	return record, nil
}

type task12IdentityParticipant struct {
	database  *task12Database
	authority identity.DeviceEnrollmentAuthority
	found     bool
	err       error
}

func (participant *task12IdentityParticipant) ValidateDeviceEnrollment(_ context.Context, dbtx store.DBTX, digest [32]byte, _ time.Time) (identity.DeviceEnrollmentAuthority, bool, error) {
	if dbtx != participant.database {
		return identity.DeviceEnrollmentAuthority{}, false, errors.New("wrong caller DBTX")
	}
	participant.database.record("validate_enrollment")
	if participant.err != nil {
		return identity.DeviceEnrollmentAuthority{}, false, participant.err
	}
	if !bytes.Equal(participant.database.grant.TokenHash, digest[:]) || participant.database.grant.State != "unused" {
		return identity.DeviceEnrollmentAuthority{}, false, nil
	}
	return participant.authority, participant.found, nil
}

func (participant *task12IdentityParticipant) ValidateDeviceAccountAuthority(_ context.Context, dbtx store.DBTX, _ identity.PrincipalID, _ time.Time) (bool, error) {
	if dbtx != participant.database {
		return false, errors.New("wrong caller DBTX")
	}
	participant.database.rotationOperations = append(participant.database.rotationOperations, "account")
	return true, nil
}

func (*task12IdentityParticipant) ValidateDeviceRevocation(context.Context, store.DBTX, identity.DeviceRevocationRequest, time.Time) (identity.DeviceRevocationAuthority, bool, error) {
	return identity.DeviceRevocationAuthority{}, false, errors.New("Task13 unavailable")
}

func (*task12IdentityParticipant) RecordDeviceTokenReplay(context.Context, store.DBTX, identity.DeviceTokenReplaySecurityRecord, time.Time) error {
	return errors.New("Task13 unavailable")
}

func (participant *task12IdentityParticipant) BindSessionToAuthorization(_ context.Context, dbtx store.DBTX, _ identity.SessionID, _ uuid.UUID, _ time.Time) error {
	if dbtx != participant.database {
		return errors.New("wrong caller DBTX")
	}
	participant.database.record("bind_session")
	return nil
}

func (*task12IdentityParticipant) RevokeAuthorizationSessions(context.Context, store.DBTX, uuid.UUID, time.Time) error {
	return errors.New("Task13 unavailable")
}

type task12Repository struct {
	mu               sync.Mutex
	database         *task12Database
	inTransaction    atomic.Bool
	failFinalCommit  bool
	panicFinalCommit bool
}

func (repository *task12Repository) WithinTransaction(ctx context.Context, operation func(context.Context, Transaction) error) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	snapshot := repository.database.snapshot()
	repository.inTransaction.Store(true)
	transaction := &task12Transaction{database: repository.database}
	err := operation(ctx, transaction)
	repository.inTransaction.Store(false)
	if err != nil {
		repository.database.restore(snapshot)
		repository.database.record("rollback")
		return err
	}
	if repository.failFinalCommit {
		repository.database.restore(snapshot)
		repository.database.record("rollback")
		if repository.panicFinalCommit {
			panic("FINAL-COMMIT-PANIC-CANARY")
		}
		return ErrRepository
	}
	repository.database.record("commit")
	return nil
}

type task12Transaction struct{ database *task12Database }

func (transaction *task12Transaction) DBTX() store.DBTX { return transaction.database }
func (transaction *task12Transaction) BeginIdempotency(ctx context.Context, scope idempotency.Scope, key string, canonical []byte, createdAt, expiresAt time.Time) (idempotency.Record, idempotency.Outcome, error) {
	transaction.database.record("begin_idempotency")
	record, outcome, err := transaction.database.idempotency.Begin(ctx, scope, key, canonical, createdAt, expiresAt)
	transaction.database.beginIdempotencyCalls++
	if err != nil || transaction.database.finalReplayBody == nil || transaction.database.beginIdempotencyCalls != 2 || outcome != idempotency.Started {
		return record, outcome, err
	}
	completed, err := transaction.database.idempotency.Complete(ctx, record, registrationResponseStatus, transaction.database.finalReplayBody)
	if err != nil {
		return idempotency.Record{}, idempotency.Outcome(""), err
	}
	owned, ok := completed.TakeResponseBody()
	clear(owned)
	if !ok {
		return idempotency.Record{}, idempotency.Outcome(""), errors.New("final replay completion ownership unavailable")
	}
	return transaction.database.idempotency.Begin(ctx, scope, key, canonical, createdAt, expiresAt)
}
func (transaction *task12Transaction) CompleteIdempotency(ctx context.Context, record idempotency.Record, status int, body []byte) error {
	transaction.database.record("complete_idempotency")
	transaction.database.replayBodyOwner = body
	completed, err := transaction.database.idempotency.Complete(ctx, record, status, body)
	if err != nil {
		return err
	}
	owned, ok := completed.TakeResponseBody()
	clear(owned)
	if !ok {
		return errors.New("completion ownership unavailable")
	}
	return nil
}
func (transaction *task12Transaction) ConsumeEnrollmentGrant(_ context.Context, digest [32]byte, deviceID uuid.UUID, now time.Time) (store.DeviceauthEnrollmentGrant, bool, error) {
	transaction.database.record("consume_grant")
	if transaction.database.grant.State != "unused" || !bytes.Equal(transaction.database.grant.TokenHash, digest[:]) || transaction.database.grant.ExpiresAt.Before(now) {
		return store.DeviceauthEnrollmentGrant{}, false, nil
	}
	transaction.database.grant.State = "consumed"
	transaction.database.grant.ConsumedAt = sql.NullTime{Time: now, Valid: true}
	transaction.database.grant.ConsumedDeviceID = uuid.NullUUID{UUID: deviceID, Valid: true}
	return transaction.database.grant, true, nil
}
func (transaction *task12Transaction) CreateDevice(_ context.Context, params store.CreateDeviceParams) error {
	transaction.database.record("create_device")
	transaction.database.deviceCount++
	transaction.database.device = cloneTask12DeviceParams(params)
	return nil
}
func (transaction *task12Transaction) CreateDeviceAuthorization(_ context.Context, params store.CreateDeviceAuthorizationParams) error {
	transaction.database.record("create_authorization")
	transaction.database.authorizationCount++
	transaction.database.authorization = params
	return nil
}
func (transaction *task12Transaction) CreateDevicePolicySnapshot(_ context.Context, params store.CreateDevicePolicySnapshotParams) error {
	transaction.database.record("create_policy")
	transaction.database.policy = params
	transaction.database.policy.Policy = bytes.Clone(params.Policy)
	return nil
}
func (transaction *task12Transaction) CreateDeviceTokenFamily(_ context.Context, params store.CreateDeviceTokenFamilyParams) error {
	transaction.database.record("create_family")
	transaction.database.accessDigestOwner = params.AccessTokenHash
	transaction.database.familyCount++
	transaction.database.family = params
	transaction.database.family.AccessTokenHash = bytes.Clone(params.AccessTokenHash)
	return nil
}
func (transaction *task12Transaction) InsertDeviceRefreshToken(_ context.Context, params store.InsertDeviceRefreshTokenParams) error {
	transaction.database.record("insert_refresh")
	transaction.database.refreshDigestOwner = params.TokenHash
	transaction.database.refreshCount++
	transaction.database.refresh = params
	transaction.database.refresh.TokenHash = bytes.Clone(params.TokenHash)
	return nil
}
func (transaction *task12Transaction) AppendEvent(_ context.Context, event *eventsv1.EventEnvelope) error {
	transaction.database.record("append_event")
	transaction.database.lastEvent = event
	return nil
}

type task12Database struct {
	repository            *task12Repository
	idempotencyDB         *task12IdempotencyDB
	idempotency           idempotency.Repository
	grant                 store.DeviceauthEnrollmentGrant
	deviceCount           int
	authorizationCount    int
	familyCount           int
	refreshCount          int
	device                store.CreateDeviceParams
	authorization         store.CreateDeviceAuthorizationParams
	policy                store.CreateDevicePolicySnapshotParams
	family                store.CreateDeviceTokenFamilyParams
	refresh               store.InsertDeviceRefreshTokenParams
	lastEvent             *eventsv1.EventEnvelope
	operations            []string
	accessDigestOwner     []byte
	refreshDigestOwner    []byte
	replayBodyOwner       []byte
	finalReplayBody       []byte
	beginIdempotencyCalls int
	rotationRefresh       store.DiscoverDeviceRefreshTokenRow
	rotationFamily        store.DeviceauthDeviceTokenFamily
	rotationAuthorization store.DeviceauthDeviceAuthorization
	rotationDevice        store.DeviceauthDevice
	rotationLockedRefresh []store.DeviceauthDeviceRefreshToken
	rotationOperations    []string
}

func newTask12Database(protector sensitive.Protector) *task12Database {
	idempotencyDB := newTask12IdempotencyDB()
	bound, err := idempotency.New(idempotencyDB, protector)
	if err != nil {
		panic(err)
	}
	return &task12Database{idempotencyDB: idempotencyDB, idempotency: bound}
}

func (database *task12Database) record(operation string) {
	database.operations = append(database.operations, operation)
}

func (database *task12Database) Exec(ctx context.Context, query string, arguments ...any) (pgconn.CommandTag, error) {
	return database.idempotencyDB.Exec(ctx, query, arguments...)
}
func (database *task12Database) Query(ctx context.Context, query string, arguments ...any) (pgx.Rows, error) {
	if strings.Contains(query, "FROM deviceauth.device_refresh_tokens") && strings.Contains(query, "ORDER BY token_hash") {
		operation := "list_refresh"
		if strings.Contains(query, "FOR UPDATE") {
			operation = "lock_refresh"
		}
		database.rotationOperations = append(database.rotationOperations, operation)
		return &task13RotationRows{refresh: database.rotationLockedRefresh}, nil
	}
	return database.idempotencyDB.Query(ctx, query, arguments...)
}
func (database *task12Database) QueryRow(ctx context.Context, query string, arguments ...any) pgx.Row {
	switch {
	case strings.Contains(query, "FROM deviceauth.device_refresh_tokens r") && !strings.Contains(query, "FOR UPDATE"):
		database.rotationOperations = append(database.rotationOperations, "discover_refresh")
		row := database.rotationRefresh
		return task13RotationRow{values: []any{
			row.TokenHash, row.FamilyID, row.PreviousTokenHash, row.RefreshState, row.IssuedAt, row.UsedAt, row.RevokedAt,
			row.AuthorizationID, row.FamilyState, row.FamilyStateVersion, row.AccessExpiresAt, row.IdleExpiresAt, row.AbsoluteExpiresAt,
			row.PrincipalID, row.DeviceID, row.AuthorizationState, row.AuthorizationStateVersion, row.ProvisionalUntil,
			row.DeviceState, row.SigningPublicKey, row.KeyVersion,
		}}
	case strings.Contains(query, "FROM deviceauth.device_token_families") && strings.Contains(query, "FOR UPDATE"):
		database.rotationOperations = append(database.rotationOperations, "family")
		row := database.rotationFamily
		return task13RotationRow{values: []any{
			row.ID, row.AuthorizationID, row.State, row.StateVersion, row.AccessTokenHash, row.AccessExpiresAt,
			row.IdleExpiresAt, row.AbsoluteExpiresAt, row.CreatedAt, row.UpdatedAt,
		}}
	case strings.Contains(query, "FROM deviceauth.device_authorizations") && strings.Contains(query, "FOR UPDATE"):
		database.rotationOperations = append(database.rotationOperations, "authorization")
		row := database.rotationAuthorization
		return task13RotationRow{values: []any{
			row.ID, row.PrincipalID, row.DeviceID, row.State, row.StateVersion, row.ProvisionalUntil, row.CreatedAt, row.UpdatedAt,
		}}
	case strings.Contains(query, "FROM deviceauth.devices") && strings.Contains(query, "FOR UPDATE"):
		database.rotationOperations = append(database.rotationOperations, "device")
		row := database.rotationDevice
		return task13RotationRow{values: []any{
			row.ID, row.PrincipalID, row.DisplayNameCiphertext, row.DisplayNameKeyVersion, row.SigningPublicKey,
			row.HpkePublicKey, row.KeyVersion, row.State, row.CreatedAt, row.UpdatedAt,
		}}
	}
	return database.idempotencyDB.QueryRow(ctx, query, arguments...)
}

type task13RotationRows struct {
	refresh []store.DeviceauthDeviceRefreshToken
	index   int
}

func (*task13RotationRows) Close()                                       {}
func (*task13RotationRows) Err() error                                   { return nil }
func (*task13RotationRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (*task13RotationRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (rows *task13RotationRows) Next() bool {
	if rows.index >= len(rows.refresh) {
		return false
	}
	rows.index++
	return true
}
func (rows *task13RotationRows) Scan(destinations ...any) error {
	if rows.index == 0 || rows.index > len(rows.refresh) {
		return errors.New("rotation Scan without current row")
	}
	row := rows.refresh[rows.index-1]
	return task13AssignRotationValues(destinations, []any{
		row.TokenHash, row.FamilyID, row.PreviousTokenHash, row.State, row.IssuedAt, row.UsedAt, row.RevokedAt,
	})
}
func (*task13RotationRows) Values() ([]any, error) { return nil, errors.New("unused") }
func (*task13RotationRows) RawValues() [][]byte    { return nil }
func (*task13RotationRows) Conn() *pgx.Conn        { return nil }

type task13RotationRow struct{ values []any }

func (row task13RotationRow) Scan(destinations ...any) error {
	return task13AssignRotationValues(destinations, row.values)
}

func task13AssignRotationValues(destinations, values []any) error {
	if len(destinations) != len(values) {
		return errors.New("unexpected rotation destination count")
	}
	for index := range destinations {
		switch destination := destinations[index].(type) {
		case *uuid.UUID:
			*destination = values[index].(uuid.UUID)
		case *[]byte:
			*destination = bytes.Clone(values[index].([]byte))
		case *json.RawMessage:
			switch value := values[index].(type) {
			case []byte:
				*destination = bytes.Clone(value)
			case json.RawMessage:
				*destination = bytes.Clone(value)
			default:
				return fmt.Errorf("unexpected JSON rotation value %T", values[index])
			}
		case *string:
			*destination = values[index].(string)
		case *int64:
			*destination = values[index].(int64)
		case *int32:
			*destination = values[index].(int32)
		case *time.Time:
			*destination = values[index].(time.Time)
		case *sql.NullTime:
			*destination = values[index].(sql.NullTime)
		case *pgtype.Int4:
			*destination = values[index].(pgtype.Int4)
		default:
			return fmt.Errorf("unexpected rotation destination %T", destinations[index])
		}
	}
	return nil
}

type task12DatabaseSnapshot struct {
	idempotency        map[string]store.IdempotencyRecord
	grant              store.DeviceauthEnrollmentGrant
	deviceCount        int
	authorizationCount int
	familyCount        int
	refreshCount       int
	device             store.CreateDeviceParams
	authorization      store.CreateDeviceAuthorizationParams
	policy             store.CreateDevicePolicySnapshotParams
	family             store.CreateDeviceTokenFamilyParams
	refresh            store.InsertDeviceRefreshTokenParams
	lastEvent          *eventsv1.EventEnvelope
}

func (database *task12Database) snapshot() task12DatabaseSnapshot {
	return task12DatabaseSnapshot{
		idempotency: database.idempotencyDB.snapshot(), grant: cloneTask12Grant(database.grant),
		deviceCount: database.deviceCount, authorizationCount: database.authorizationCount, familyCount: database.familyCount, refreshCount: database.refreshCount,
		device: cloneTask12DeviceParams(database.device), authorization: database.authorization,
		policy: database.policy, family: database.family, refresh: database.refresh, lastEvent: database.lastEvent,
	}
}

func (database *task12Database) restore(snapshot task12DatabaseSnapshot) {
	database.idempotencyDB.restore(snapshot.idempotency)
	database.grant = cloneTask12Grant(snapshot.grant)
	database.deviceCount, database.authorizationCount = snapshot.deviceCount, snapshot.authorizationCount
	database.familyCount, database.refreshCount = snapshot.familyCount, snapshot.refreshCount
	database.device, database.authorization = cloneTask12DeviceParams(snapshot.device), snapshot.authorization
	database.policy, database.family, database.refresh, database.lastEvent = snapshot.policy, snapshot.family, snapshot.refresh, snapshot.lastEvent
}

func (database *task12Database) capturedBytes() []byte {
	result := append([]byte(nil), database.device.DisplayNameCiphertext...)
	result = append(result, database.device.SigningPublicKey...)
	result = append(result, database.device.HpkePublicKey...)
	result = append(result, database.family.AccessTokenHash...)
	result = append(result, database.refresh.TokenHash...)
	result = append(result, database.policy.Policy...)
	if database.lastEvent != nil {
		result = append(result, database.lastEvent.Payload...)
	}
	return result
}

func cloneTask12Grant(grant store.DeviceauthEnrollmentGrant) store.DeviceauthEnrollmentGrant {
	grant.TokenHash = bytes.Clone(grant.TokenHash)
	return grant
}
func cloneTask12DeviceParams(params store.CreateDeviceParams) store.CreateDeviceParams {
	params.DisplayNameCiphertext = bytes.Clone(params.DisplayNameCiphertext)
	params.SigningPublicKey = bytes.Clone(params.SigningPublicKey)
	params.HpkePublicKey = bytes.Clone(params.HpkePublicKey)
	return params
}

type task12IdempotencyDB struct {
	mu      sync.Mutex
	records map[string]store.IdempotencyRecord
}

func newTask12IdempotencyDB() *task12IdempotencyDB {
	return &task12IdempotencyDB{records: make(map[string]store.IdempotencyRecord)}
}

func (database *task12IdempotencyDB) Exec(_ context.Context, query string, arguments ...any) (pgconn.CommandTag, error) {
	if !strings.Contains(query, "INSERT INTO idempotency_records") || len(arguments) != 6 {
		return pgconn.CommandTag{}, errors.New("unexpected idempotency exec")
	}
	database.mu.Lock()
	defer database.mu.Unlock()
	key := task12IdempotencyKey(arguments[0].(string), arguments[1].(string), arguments[2].([]byte))
	if _, exists := database.records[key]; exists {
		return pgconn.NewCommandTag("INSERT 0 0"), nil
	}
	database.records[key] = store.IdempotencyRecord{
		PrincipalScope: arguments[0].(string), Operation: arguments[1].(string), IdempotencyKeyHash: bytes.Clone(arguments[2].([]byte)),
		RequestDigest: bytes.Clone(arguments[3].([]byte)), State: "in_progress", CreatedAt: arguments[4].(time.Time), ExpiresAt: arguments[5].(time.Time),
	}
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

func (*task12IdempotencyDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected idempotency query")
}
func (database *task12IdempotencyDB) QueryRow(_ context.Context, query string, arguments ...any) pgx.Row {
	database.mu.Lock()
	defer database.mu.Unlock()
	if strings.Contains(query, "UPDATE idempotency_records") && len(arguments) == 7 {
		key := task12IdempotencyKey(arguments[0].(string), arguments[1].(string), arguments[2].([]byte))
		record, exists := database.records[key]
		if !exists || record.State != "in_progress" || !bytes.Equal(record.RequestDigest, arguments[3].([]byte)) {
			return task12IdempotencyRow{err: pgx.ErrNoRows}
		}
		record.State = "completed"
		record.ResponseStatus = arguments[4].(pgtype.Int4)
		record.ResponseCiphertext = bytes.Clone(arguments[5].([]byte))
		record.ResponseKeyVersion = arguments[6].(pgtype.Int4)
		database.records[key] = cloneTask12IdempotencyRecord(record)
		return task12IdempotencyRow{record: record}
	}
	if strings.Contains(query, "FROM idempotency_records") && len(arguments) == 3 {
		key := task12IdempotencyKey(arguments[0].(string), arguments[1].(string), arguments[2].([]byte))
		record, exists := database.records[key]
		if !exists {
			return task12IdempotencyRow{err: pgx.ErrNoRows}
		}
		return task12IdempotencyRow{record: cloneTask12IdempotencyRecord(record)}
	}
	return task12IdempotencyRow{err: errors.New("unexpected idempotency query row")}
}
func (database *task12IdempotencyDB) snapshot() map[string]store.IdempotencyRecord {
	database.mu.Lock()
	defer database.mu.Unlock()
	result := make(map[string]store.IdempotencyRecord, len(database.records))
	for key, record := range database.records {
		result[key] = cloneTask12IdempotencyRecord(record)
	}
	return result
}
func (database *task12IdempotencyDB) restore(snapshot map[string]store.IdempotencyRecord) {
	database.mu.Lock()
	defer database.mu.Unlock()
	database.records = make(map[string]store.IdempotencyRecord, len(snapshot))
	for key, record := range snapshot {
		database.records[key] = cloneTask12IdempotencyRecord(record)
	}
}

func task12IdempotencyKey(principal, operation string, digest []byte) string {
	return principal + "\x00" + operation + "\x00" + string(digest)
}
func cloneTask12IdempotencyRecord(record store.IdempotencyRecord) store.IdempotencyRecord {
	record.IdempotencyKeyHash = bytes.Clone(record.IdempotencyKeyHash)
	record.RequestDigest = bytes.Clone(record.RequestDigest)
	record.ResponseCiphertext = bytes.Clone(record.ResponseCiphertext)
	return record
}

type task12IdempotencyRow struct {
	record store.IdempotencyRecord
	err    error
}

func (row task12IdempotencyRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != 10 {
		return errors.New("unexpected idempotency destination count")
	}
	*destinations[0].(*string) = row.record.PrincipalScope
	*destinations[1].(*string) = row.record.Operation
	*destinations[2].(*[]byte) = bytes.Clone(row.record.IdempotencyKeyHash)
	*destinations[3].(*[]byte) = bytes.Clone(row.record.RequestDigest)
	*destinations[4].(*string) = row.record.State
	*destinations[5].(*pgtype.Int4) = row.record.ResponseStatus
	*destinations[6].(*[]byte) = bytes.Clone(row.record.ResponseCiphertext)
	*destinations[7].(*pgtype.Int4) = row.record.ResponseKeyVersion
	*destinations[8].(*time.Time) = row.record.CreatedAt
	*destinations[9].(*time.Time) = row.record.ExpiresAt
	return nil
}

func assertTask12Tokens(t *testing.T, tokens DeviceTokens) {
	t.Helper()
	if tokens.DeviceID == uuid.Nil || tokens.AuthorizationID == uuid.Nil || len(tokens.AccessToken.Copy()) != 32 || len(tokens.RefreshToken.Copy()) != 32 ||
		tokens.AccessExpiresAt != fixedTask12Time.Add(10*time.Minute) || tokens.RefreshIdleExpiresAt != fixedTask12Time.Add(30*24*time.Hour) ||
		tokens.RefreshAbsoluteExpiresAt != fixedTask12Time.Add(90*24*time.Hour) {
		t.Fatalf("device token contract mismatch: %+v", tokens)
	}
}

func sameTask12Tokens(left, right DeviceTokens) bool {
	leftAccess, rightAccess := left.AccessToken.Copy(), right.AccessToken.Copy()
	leftRefresh, rightRefresh := left.RefreshToken.Copy(), right.RefreshToken.Copy()
	defer clear(leftAccess)
	defer clear(rightAccess)
	defer clear(leftRefresh)
	defer clear(rightRefresh)
	return left.DeviceID == right.DeviceID && left.AuthorizationID == right.AuthorizationID && bytes.Equal(leftAccess, rightAccess) &&
		bytes.Equal(leftRefresh, rightRefresh) && left.AccessExpiresAt.Equal(right.AccessExpiresAt) &&
		left.RefreshIdleExpiresAt.Equal(right.RefreshIdleExpiresAt) && left.RefreshAbsoluteExpiresAt.Equal(right.RefreshAbsoluteExpiresAt)
}

func assertTask12FinalOrder(t *testing.T, operations []string) {
	t.Helper()
	want := []string{
		"begin_idempotency", "validate_enrollment", "create_device", "consume_grant", "create_authorization", "bind_session",
		"create_policy", "create_family", "insert_refresh", "append_event", "complete_idempotency", "commit",
	}
	position := 0
	for _, operation := range operations {
		if position < len(want) && operation == want[position] {
			position++
		}
	}
	if position != len(want) {
		t.Fatalf("operations = %s, missing exact final order suffix %s", strings.Join(operations, ","), strings.Join(want[position:], ","))
	}
}

func publicTask12Code(err error) apierrors.Code {
	var classified apierrors.Error
	if !errors.As(err, &classified) {
		return ""
	}
	return apierrors.Code(classified.Public("trace-safe-000012").Code)
}

var _ securitykit.Clock = task12Clock{}
var _ ratelimit.Limiter = (*task12Limiter)(nil)
var _ ChallengeStore = (*task12ChallengeStore)(nil)
var _ identity.DeviceTransactionParticipant = (*task12IdentityParticipant)(nil)
var _ Repository = (*task12Repository)(nil)
var _ Transaction = (*task12Transaction)(nil)
var _ store.DBTX = (*task12IdempotencyDB)(nil)

type task12GeneratedDBTX struct {
	operations   []string
	grant        store.DeviceauthEnrollmentGrant
	activateRows int64
}

func (database *task12GeneratedDBTX) Exec(_ context.Context, query string, _ ...any) (pgconn.CommandTag, error) {
	switch {
	case strings.Contains(query, "INSERT INTO deviceauth.devices"):
		database.operations = append(database.operations, "device")
		return pgconn.NewCommandTag("INSERT 0 1"), nil
	case strings.Contains(query, "INSERT INTO deviceauth.device_authorizations"):
		database.operations = append(database.operations, "authorization")
		return pgconn.NewCommandTag("INSERT 0 1"), nil
	case strings.Contains(query, "INSERT INTO deviceauth.device_policy_snapshots"):
		database.operations = append(database.operations, "policy")
		return pgconn.NewCommandTag("INSERT 0 1"), nil
	case strings.Contains(query, "INSERT INTO deviceauth.device_token_families"):
		database.operations = append(database.operations, "family")
		return pgconn.NewCommandTag("INSERT 0 1"), nil
	case strings.Contains(query, "INSERT INTO deviceauth.device_refresh_tokens"):
		database.operations = append(database.operations, "refresh")
		return pgconn.NewCommandTag("INSERT 0 1"), nil
	case strings.Contains(query, "SET state='active'") && strings.Contains(query, "deviceauth.device_authorizations"):
		database.operations = append(database.operations, "activate")
		return pgconn.NewCommandTag(fmt.Sprintf("UPDATE %d", database.activateRows)), nil
	default:
		return pgconn.CommandTag{}, errors.New("unexpected generated Exec")
	}
}
func (*task12GeneratedDBTX) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected generated Query")
}
func (database *task12GeneratedDBTX) QueryRow(_ context.Context, query string, _ ...any) pgx.Row {
	if strings.Contains(query, "UPDATE deviceauth.enrollment_grants") {
		database.operations = append(database.operations, "consume")
		return task12GeneratedGrantRow{grant: database.grant}
	}
	return task12GeneratedGrantRow{err: errors.New("unexpected generated QueryRow")}
}

type task12GeneratedGrantRow struct {
	grant store.DeviceauthEnrollmentGrant
	err   error
}

func (row task12GeneratedGrantRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != 11 {
		return errors.New("unexpected grant destination count")
	}
	*destinations[0].(*uuid.UUID) = row.grant.ID
	*destinations[1].(*uuid.UUID) = row.grant.PrincipalID
	*destinations[2].(*uuid.UUID) = row.grant.AccountSessionID
	*destinations[3].(*[]byte) = bytes.Clone(row.grant.TokenHash)
	*destinations[4].(*string) = row.grant.PolicyMarker
	*destinations[5].(*sql.NullTime) = row.grant.ProvisionalUntil
	*destinations[6].(*string) = row.grant.State
	*destinations[7].(*time.Time) = row.grant.ExpiresAt
	*destinations[8].(*time.Time) = row.grant.CreatedAt
	*destinations[9].(*sql.NullTime) = row.grant.ConsumedAt
	*destinations[10].(*uuid.NullUUID) = row.grant.ConsumedDeviceID
	return nil
}

type panicTask12DBTX struct{}

func (panicTask12DBTX) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	panic("EXEC-CANARY")
}
func (panicTask12DBTX) Query(context.Context, string, ...any) (pgx.Rows, error) {
	panic("QUERY-CANARY")
}
func (panicTask12DBTX) QueryRow(context.Context, string, ...any) pgx.Row { panic("ROW-CANARY") }

var _ store.DBTX = (*task12GeneratedDBTX)(nil)
var _ store.DBTX = panicTask12DBTX{}
