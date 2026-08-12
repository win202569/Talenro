package deviceauth

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/ratelimit"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

func TestRotateProofBindsChallengeFamilyOperationAudienceAndNonce(t *testing.T) {
	t.Parallel()

	input := DeviceRotationProofInput{
		ProtocolVersion: "device-token-rotation-v1", Challenge: [32]byte{1, 2, 3},
		FamilyID: uuid.MustParse("0ff820a5-5022-48e6-8867-77761f8e2f07"), Operation: "rotate_device_token",
		Audience: "https://api.example.test", RequestNonce: [32]byte{4, 5, 6},
	}
	want := []byte("TALENRO-DEVICE-ROTATION-V1\x00")
	for _, part := range [][]byte{[]byte(input.ProtocolVersion), input.Challenge[:], input.FamilyID[:], []byte(input.Operation), []byte(input.Audience), input.RequestNonce[:]} {
		if len(part) == 32 || len(part) == 16 {
			want = append(want, part...)
			continue
		}
		var length [4]byte
		binary.BigEndian.PutUint32(length[:], uint32(len(part))) // #nosec G115 -- literal test fixture is bounded.
		want = append(want, length[:]...)
		want = append(want, part...)
	}
	got := DeviceRotationProofBytes(input)
	if !bytes.Equal(got, want) {
		t.Fatalf("rotation transcript mismatch\n got %x\nwant %x", got, want)
	}

	mutations := []func(*DeviceRotationProofInput){
		func(value *DeviceRotationProofInput) { value.ProtocolVersion = "device-token-rotation-v2" },
		func(value *DeviceRotationProofInput) { value.Challenge[0] ^= 0xff },
		func(value *DeviceRotationProofInput) { value.FamilyID = uuid.New() },
		func(value *DeviceRotationProofInput) { value.Operation = "register_device" },
		func(value *DeviceRotationProofInput) { value.Audience = "https://other.example.test" },
		func(value *DeviceRotationProofInput) { value.RequestNonce[0] ^= 0xff },
	}
	for _, mutate := range mutations {
		changed := input
		mutate(&changed)
		if bytes.Equal(DeviceRotationProofBytes(changed), got) {
			t.Fatal("rotation proof field substitution retained the original transcript")
		}
	}
}

func TestRotateDeviceTokenUsesExactFixedTTLs(t *testing.T) {
	t.Parallel()

	if deviceAccessTTL.String() != "10m0s" || deviceRefreshIdleTTL.Hours() != 30*24 || deviceRefreshAbsoluteTTL.Hours() != 90*24 {
		t.Fatal("device token TTL contract changed")
	}
}

func TestDeviceTokenInternalSecretDTOsRedactDiagnosticsAndRejectJSON(t *testing.T) {
	t.Parallel()
	canary := bytes.Repeat([]byte("S"), 32)
	authority := deviceRefreshAuthority{discovered: store.DiscoverDeviceRefreshTokenRow{TokenHash: bytes.Clone(canary), SigningPublicKey: bytes.Clone(canary)},
		family: store.DeviceauthDeviceTokenFamily{AccessTokenHash: bytes.Clone(canary)}, device: store.DeviceauthDevice{SigningPublicKey: bytes.Clone(canary)},
		refresh: []store.DeviceauthDeviceRefreshToken{{TokenHash: bytes.Clone(canary)}}}
	prepared := preparedDeviceRotation{tokens: DeviceTokens{AccessToken: secret.NewBytes(canary), RefreshToken: secret.NewBytes(canary)}, accessDigest: [32]byte{0x53}, refreshDigest: [32]byte{0x53}}
	locked := lockedDeviceRevocation{device: store.DeviceauthDevice{SigningPublicKey: bytes.Clone(canary)},
		families: []store.DeviceauthDeviceTokenFamily{{AccessTokenHash: bytes.Clone(canary)}}, refresh: []store.DeviceauthDeviceRefreshToken{{TokenHash: bytes.Clone(canary)}}}
	defer authority.clear()
	defer prepared.clear()
	defer locked.clear()
	assertTask12RedactionSubjects(t, []task12RedactionSubject{
		newTask12RedactionSubject("deviceRefreshAuthority", authority, deviceRefreshAuthority{}),
		newTask12RedactionSubject("preparedDeviceRotation", prepared, preparedDeviceRotation{}),
		newTask12RedactionSubject("lockedDeviceRevocation", locked, lockedDeviceRevocation{}),
	}, []string{string(canary)})
}

func TestRotateDeviceTokenConcurrentIdempotentCallsHaveOneMutationWinner(t *testing.T) {
	fixture := newTask13Fixture(t)
	command := fixture.rotationCommand(t, "task13-concurrent-rotation-0001", [32]byte{0x41})

	start := make(chan struct{})
	results := make(chan DeviceTokens, 2)
	errorsFound := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			result, err := fixture.application.RotateDeviceToken(context.Background(), command)
			results <- result
			errorsFound <- err
		}()
	}
	close(start)
	first, second := <-results, <-results
	firstErr, secondErr := <-errorsFound, <-errorsFound
	defer first.AccessToken.Clear()
	defer first.RefreshToken.Clear()
	defer second.AccessToken.Clear()
	defer second.RefreshToken.Clear()
	successes := 0
	if firstErr == nil {
		successes++
	}
	if secondErr == nil {
		successes++
	}
	if successes == 0 || (successes == 2 && !sameTask12Tokens(first, second)) ||
		(firstErr != nil && publicTask12Code(firstErr) != apierrors.AuthenticationFailed) ||
		(secondErr != nil && publicTask12Code(secondErr) != apierrors.AuthenticationFailed) {
		t.Fatalf("concurrent idempotent rotations = (%v, %v) / (%v, %v), want one mutation winner and only replay/auth-failure follower", first, second, firstErr, secondErr)
	}
	if fixture.state.rotationCount != 1 || fixture.challenges.consumeCalls < 1 || fixture.challenges.consumeCalls > 2 || len(fixture.state.refresh) != 2 {
		t.Fatalf("mutation/Redis-attempt/refresh counts = %d/%d/%d, want 1/[1,2]/2", fixture.state.rotationCount, fixture.challenges.consumeCalls, len(fixture.state.refresh))
	}
}

func TestUsedDeviceRefreshReplayCommitsCompromiseBeforeAuthenticationFailure(t *testing.T) {
	fixture := newTask13Fixture(t)
	first := fixture.rotationCommand(t, "task13-first-rotation-0001", [32]byte{0x51})
	second := fixture.rotationCommand(t, "task13-replay-rotation-0001", [32]byte{0x52})

	tokens, err := fixture.application.RotateDeviceToken(context.Background(), first)
	if err != nil {
		t.Fatalf("first rotation: %v", err)
	}
	tokens.AccessToken.Clear()
	tokens.RefreshToken.Clear()
	consumes := fixture.challenges.consumeCalls
	if _, err := fixture.application.RotateDeviceToken(context.Background(), second); publicTask12Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("used replay error = %v, want finite authentication failure", err)
	}
	if fixture.challenges.consumeCalls != consumes {
		t.Fatal("used-token replay reached Redis before PostgreSQL replay classification")
	}
	family := fixture.state.families[fixture.familyID]
	if family.State != "compromised" || fixture.identity.replayRecords != 1 || fixture.state.compromiseEvents != 1 {
		t.Fatalf("committed replay effects = family %q, security %d, outbox %d", family.State, fixture.identity.replayRecords, fixture.state.compromiseEvents)
	}
	compromiseEvent := fixture.state.events[len(fixture.state.events)-1]
	if compromiseEvent.GetIdempotencyKey() == second.IdempotencyKey {
		t.Fatal("compromise event exposed the caller idempotency key")
	}
	if parsed, err := uuid.Parse(compromiseEvent.GetIdempotencyKey()); err != nil || parsed == uuid.Nil || parsed.String() != compromiseEvent.GetIdempotencyKey() {
		t.Fatal("compromise event did not use an approved opaque synthetic deduplication key")
	}
	for _, refresh := range fixture.state.refresh {
		if refresh.State == "active" {
			t.Fatal("compromise left an active refresh token in the replayed family")
		}
	}
}

func TestUsedDeviceRefreshReplayAfterIdleBeforeAbsoluteStillCommitsCompromise(t *testing.T) {
	fixture := newTask13Fixture(t)
	winner := fixture.rotationCommand(t, "task13-post-idle-winner-0001", [32]byte{0x53})
	replay := fixture.rotationCommand(t, "task13-post-idle-replay-0001", [32]byte{0x54})

	tokens, err := fixture.application.RotateDeviceToken(context.Background(), winner)
	if err != nil {
		t.Fatalf("winner rotation: %v", err)
	}
	tokens.AccessToken.Clear()
	tokens.RefreshToken.Clear()
	fixture.application.clock = task13FixedClock{now: fixedTask12Time.Add(deviceRefreshIdleTTL + time.Hour)}
	consumes := fixture.challenges.consumeCalls
	if _, err := fixture.application.RotateDeviceToken(context.Background(), replay); publicTask12Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("used replay after idle and before absolute = %v, want authentication failure", err)
	}
	if fixture.challenges.consumeCalls != consumes {
		t.Fatal("post-idle used-token replay reached Redis before PostgreSQL replay classification")
	}
	if fixture.state.families[fixture.familyID].State != "compromised" || fixture.identity.replayRecords != 1 || fixture.state.compromiseEvents != 1 {
		t.Fatalf("post-idle replay effects = family %q, security %d, outbox %d",
			fixture.state.families[fixture.familyID].State, fixture.identity.replayRecords, fixture.state.compromiseEvents)
	}
	for _, refresh := range fixture.state.refresh {
		if refresh.State == "active" {
			t.Fatal("post-idle used-token replay left the winner's successor active")
		}
	}
	if _, err := fixture.application.RotateDeviceToken(context.Background(), replay); publicTask12Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("post-idle replay tombstone = %v, want authentication failure", err)
	}
	if fixture.challenges.consumeCalls != consumes || fixture.identity.replayRecords != 1 || fixture.state.compromiseEvents != 1 {
		t.Fatal("post-idle replay tombstone reached Redis or duplicated compromise effects")
	}
}

func TestConcurrentDeviceRefreshHasOneSuccessAndCompromisesTheReplay(t *testing.T) {
	fixture := newTask13Fixture(t)
	first := fixture.rotationCommand(t, "task13-concurrent-winner-0001", [32]byte{0x55})
	second := fixture.rotationCommand(t, "task13-concurrent-loser-0001", [32]byte{0x56})
	start := make(chan struct{})
	results := make(chan DeviceTokens, 2)
	errorsFound := make(chan error, 2)
	for _, command := range []RotateDeviceTokenCommand{first, second} {
		command := command
		go func() {
			<-start
			result, err := fixture.application.RotateDeviceToken(context.Background(), command)
			results <- result
			errorsFound <- err
		}()
	}
	close(start)
	firstResult, secondResult := <-results, <-results
	firstErr, secondErr := <-errorsFound, <-errorsFound
	defer firstResult.AccessToken.Clear()
	defer firstResult.RefreshToken.Clear()
	defer secondResult.AccessToken.Clear()
	defer secondResult.RefreshToken.Clear()
	successes := 0
	for _, err := range []error{firstErr, secondErr} {
		if err == nil {
			successes++
		} else if publicTask12Code(err) != apierrors.AuthenticationFailed {
			t.Fatalf("concurrent rotation error = %v, want finite authentication failure", err)
		}
	}
	if successes != 1 || fixture.state.rotationCount != 1 || fixture.state.families[fixture.familyID].State != "compromised" ||
		fixture.identity.replayRecords != 1 || fixture.state.compromiseEvents != 1 {
		t.Fatalf("concurrent outcome = successes %d, mutations %d, family %q, security/outbox %d/%d",
			successes, fixture.state.rotationCount, fixture.state.families[fixture.familyID].State, fixture.identity.replayRecords, fixture.state.compromiseEvents)
	}
}

func TestAmbiguousChallengeAfterConcurrentWinnerCompromisesUsedRefresh(t *testing.T) {
	fixture := newTask13Fixture(t)
	loser := fixture.rotationCommand(t, "task13-ambiguous-loser-0001", [32]byte{0x57})
	winner := fixture.rotationCommand(t, "task13-ambiguous-winner-0001", [32]byte{0x58})
	challenges := &task13WinnerThenAmbiguousChallengeStore{
		delegate: fixture.challenges, service: fixture.application, targetChallengeID: loser.ChallengeID, winner: winner,
	}
	fixture.application.challenges = challenges

	if _, err := fixture.application.RotateDeviceToken(context.Background(), loser); publicTask12Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("ambiguous challenge after concurrent winner = %v, want authentication failure", err)
	}
	if challenges.winnerErr != nil {
		t.Fatalf("concurrent winner: %v", challenges.winnerErr)
	}
	if fixture.state.families[fixture.familyID].State != "compromised" || fixture.identity.replayRecords != 1 || fixture.state.compromiseEvents != 1 {
		t.Fatalf("ambiguous replay effects = family %q, security %d, outbox %d",
			fixture.state.families[fixture.familyID].State, fixture.identity.replayRecords, fixture.state.compromiseEvents)
	}
	for _, refresh := range fixture.state.refresh {
		if refresh.State == "active" {
			t.Fatal("ambiguous used-token replay left the winner's successor active")
		}
	}
	consumeCalls := challenges.targetCalls
	if _, err := fixture.application.RotateDeviceToken(context.Background(), loser); publicTask12Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("ambiguous replay tombstone = %v, want authentication failure", err)
	}
	if challenges.targetCalls != consumeCalls {
		t.Fatal("completed ambiguous replay tombstone reached Redis")
	}
}

func TestRotateDeviceTokenReplaysCompletedResultBeforeRedisAndMutableAuthority(t *testing.T) {
	fixture := newTask13Fixture(t)
	command := fixture.rotationCommand(t, "task13-replay-success-0001", [32]byte{0x61})
	first, err := fixture.application.RotateDeviceToken(context.Background(), command)
	if err != nil {
		t.Fatal(err)
	}
	defer first.AccessToken.Clear()
	defer first.RefreshToken.Clear()
	family := fixture.state.families[fixture.familyID]
	family.State = "revoked"
	fixture.state.families[fixture.familyID] = family
	consumes := fixture.challenges.consumeCalls
	second, err := fixture.application.RotateDeviceToken(context.Background(), command)
	if err != nil {
		t.Fatalf("completed replay after authority change: %v", err)
	}
	defer second.AccessToken.Clear()
	defer second.RefreshToken.Clear()
	if !sameTask12Tokens(first, second) || fixture.challenges.consumeCalls != consumes {
		t.Fatal("completed replay depended on Redis or current mutable device authority")
	}
}

func TestRotateDeviceTokenReturnsExactTTLBoundariesAndRejectsSuspendedAuthority(t *testing.T) {
	fixture := newTask13Fixture(t)
	command := fixture.rotationCommand(t, "task13-rotation-ttl-0001", [32]byte{0x62})
	tokens, err := fixture.application.RotateDeviceToken(context.Background(), command)
	if err != nil {
		t.Fatal(err)
	}
	defer tokens.AccessToken.Clear()
	defer tokens.RefreshToken.Clear()
	if !tokens.AccessExpiresAt.Equal(fixedTask12Time.Add(deviceAccessTTL)) ||
		!tokens.RefreshIdleExpiresAt.Equal(fixedTask12Time.Add(deviceRefreshIdleTTL)) ||
		!tokens.RefreshAbsoluteExpiresAt.Equal(fixedTask12Time.Add(deviceRefreshAbsoluteTTL)) {
		t.Fatalf("rotated TTLs = %v/%v/%v", tokens.AccessExpiresAt, tokens.RefreshIdleExpiresAt, tokens.RefreshAbsoluteExpiresAt)
	}

	for _, mutate := range []func(*task13Fixture){
		func(value *task13Fixture) {
			row := value.state.devices[value.deviceID]
			row.State = "suspended"
			value.state.devices[value.deviceID] = row
		},
		func(value *task13Fixture) {
			row := value.state.authorizations[value.authorizationID]
			row.State = "revoked"
			value.state.authorizations[value.authorizationID] = row
		},
		func(value *task13Fixture) { value.identity.accountActive = false },
	} {
		denied := newTask13Fixture(t)
		mutate(denied)
		deniedCommand := denied.rotationCommand(t, "task13-rotation-denied-0001-"+uuid.NewString(), [32]byte{0x63})
		if _, err := denied.application.RotateDeviceToken(context.Background(), deniedCommand); publicTask12Code(err) != apierrors.AuthenticationFailed {
			t.Fatalf("suspended/revoked rotation = %v, want authentication failure", err)
		}
	}
}

type task13Fixture struct {
	application     *Service
	repository      *task13Repository
	state           *task13State
	identity        *task13IdentityParticipant
	challenges      *task13ChallengeStore
	protector       sensitive.Protector
	refresh         secret.Bytes
	access          secret.Bytes
	signingPrivate  ed25519.PrivateKey
	principalID     uuid.UUID
	deviceID        uuid.UUID
	authorizationID uuid.UUID
	familyID        uuid.UUID
}

type task13FixedClock struct{ now time.Time }

func (clock task13FixedClock) Now() time.Time { return clock.now }

func newTask13Fixture(t *testing.T) *task13Fixture {
	t.Helper()
	protector, err := sensitive.NewLocal(secret.NewBytes(bytes.Repeat([]byte{0x71}, 32)), secret.NewBytes(bytes.Repeat([]byte{0x72}, 32)), 9)
	if err != nil {
		t.Fatal(err)
	}
	principalID := uuid.MustParse("a6493384-9407-4ad9-b220-7f3b49ef9054")
	deviceID := uuid.MustParse("f353613c-d08f-4141-b287-b37db9fb6f8e")
	authorizationID := uuid.MustParse("fa01e838-cab9-414e-8f03-ed0418cd4f20")
	familyID := uuid.MustParse("0ff820a5-5022-48e6-8867-77761f8e2f07")
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	refresh := secret.NewBytes(bytes.Repeat([]byte{0x73}, 32))
	refreshDigest := securitykit.DigestToken(securitykit.DeviceRefreshToken, refresh)
	access := secret.NewBytes(bytes.Repeat([]byte{0x74}, 32))
	accessDigest := securitykit.DigestToken(securitykit.DeviceAccessToken, access)
	state := &task13State{
		base: newTask12Database(protector),
		refresh: map[string]store.DeviceauthDeviceRefreshToken{string(refreshDigest[:]): {
			TokenHash: bytes.Clone(refreshDigest[:]), FamilyID: familyID, State: "active", IssuedAt: fixedTask12Time,
		}},
		families: map[uuid.UUID]store.DeviceauthDeviceTokenFamily{familyID: {
			ID: familyID, AuthorizationID: authorizationID, State: "active", StateVersion: 1,
			AccessTokenHash: bytes.Clone(accessDigest[:]), AccessExpiresAt: fixedTask12Time.Add(deviceAccessTTL),
			IdleExpiresAt: fixedTask12Time.Add(deviceRefreshIdleTTL), AbsoluteExpiresAt: fixedTask12Time.Add(deviceRefreshAbsoluteTTL), CreatedAt: fixedTask12Time,
		}},
		authorizations: map[uuid.UUID]store.DeviceauthDeviceAuthorization{authorizationID: {
			ID: authorizationID, PrincipalID: principalID, DeviceID: deviceID, State: "active", StateVersion: 1, CreatedAt: fixedTask12Time,
		}},
		devices: map[uuid.UUID]store.DeviceauthDevice{deviceID: {
			ID: deviceID, PrincipalID: principalID, SigningPublicKey: bytes.Clone(publicKey), HpkePublicKey: bytes.Repeat([]byte{0x75}, 32),
			KeyVersion: 1, State: "active", CreatedAt: fixedTask12Time,
		}},
		policies: map[uuid.UUID]store.DeviceauthDevicePolicySnapshot{authorizationID: {
			AuthorizationID: authorizationID, SchemaVersion: "device-policy-v1", Policy: []byte(`{"mode":"standard"}`), CreatedAt: fixedTask12Time,
		}},
	}
	repository := &task13Repository{state: state}
	participant := &task13IdentityParticipant{state: state, accountActive: true, revocationAllowed: true}
	challenges := &task13ChallengeStore{records: make(map[string]ChallengeRecord), repository: repository}
	application, err := NewApplication(ApplicationDependencies{
		Repository: repository, IdentityParticipant: participant, Protector: protector, Random: rand.Reader, Clock: task12Clock{},
		Limiter: task13Limiter{}, ChallengeStore: challenges, RateLimitKey: secret.NewBytes(bytes.Repeat([]byte{0x76}, 32)),
		Security: config.SecurityConfig{Profile: config.ProfileTest, PublicBaseURL: "https://api.example.test", RequestDeadline: 2 * time.Second,
			ChallengeRateLimit: config.RateLimitPolicy{Limit: 20, Window: 5 * time.Minute}},
	})
	if err != nil {
		t.Fatal(err)
	}
	fixture := &task13Fixture{application: application, repository: repository, state: state, identity: participant, challenges: challenges,
		protector: protector, refresh: refresh, access: access, signingPrivate: privateKey, principalID: principalID, deviceID: deviceID,
		authorizationID: authorizationID, familyID: familyID}
	t.Cleanup(func() {
		fixture.refresh.Clear()
		fixture.access.Clear()
		clear(fixture.signingPrivate)
		_ = protector.Close()
	})
	return fixture
}

func (fixture *task13Fixture) rotationCommand(t *testing.T, key string, nonce [32]byte) RotateDeviceTokenCommand {
	t.Helper()
	challengeID := uuid.New().String()
	refreshDigest := securitykit.DigestToken(securitykit.DeviceRefreshToken, fixture.refresh)
	contextDigest, err := rotationContextDigest(fixture.familyID, nonce, "https://api.example.test")
	if err != nil {
		t.Fatal(err)
	}
	record := ChallengeRecord{ChallengeID: challengeID, Kind: ChallengeRotation, ProtocolVersion: deviceRotationProtocolVersion,
		Operation: rotateDeviceTokenOperation, Challenge: [32]byte{0x77}, GrantDigest: refreshDigest,
		ContextDigest: contextDigest, ExpiresAt: fixedTask12Time.Add(deviceChallengeTTL)}
	fixture.challenges.records[challengeID] = record
	proof := DeviceRotationProofBytes(DeviceRotationProofInput{ProtocolVersion: record.ProtocolVersion, Challenge: record.Challenge,
		FamilyID: fixture.familyID, Operation: record.Operation, Audience: "https://api.example.test", RequestNonce: nonce})
	signed := ed25519.Sign(fixture.signingPrivate, proof)
	clear(proof)
	var signature [64]byte
	copy(signature[:], signed)
	clear(signed)
	return RotateDeviceTokenCommand{RefreshToken: fixture.refresh, ChallengeID: challengeID, RequestNonce: nonce, Signature: signature, IdempotencyKey: key}
}

type task13State struct {
	base             *task12Database
	refresh          map[string]store.DeviceauthDeviceRefreshToken
	families         map[uuid.UUID]store.DeviceauthDeviceTokenFamily
	authorizations   map[uuid.UUID]store.DeviceauthDeviceAuthorization
	devices          map[uuid.UUID]store.DeviceauthDevice
	policies         map[uuid.UUID]store.DeviceauthDevicePolicySnapshot
	events           []*eventsv1.EventEnvelope
	rotationCount    int
	compromiseEvents int
}

type task13StateSnapshot struct {
	idempotency                     map[string]store.IdempotencyRecord
	refresh                         map[string]store.DeviceauthDeviceRefreshToken
	families                        map[uuid.UUID]store.DeviceauthDeviceTokenFamily
	authorizations                  map[uuid.UUID]store.DeviceauthDeviceAuthorization
	devices                         map[uuid.UUID]store.DeviceauthDevice
	policies                        map[uuid.UUID]store.DeviceauthDevicePolicySnapshot
	events                          []*eventsv1.EventEnvelope
	rotationCount, compromiseEvents int
}

func (state *task13State) snapshot() task13StateSnapshot {
	result := task13StateSnapshot{idempotency: state.base.idempotencyDB.snapshot(), refresh: cloneTask13Refresh(state.refresh),
		families: cloneTask13Families(state.families), authorizations: cloneTask13Authorizations(state.authorizations), devices: cloneTask13Devices(state.devices),
		policies: cloneTask13Policies(state.policies),
		events:   append([]*eventsv1.EventEnvelope(nil), state.events...), rotationCount: state.rotationCount, compromiseEvents: state.compromiseEvents}
	return result
}

func (state *task13State) restore(snapshot task13StateSnapshot) {
	state.base.idempotencyDB.restore(snapshot.idempotency)
	state.refresh, state.families, state.authorizations, state.devices, state.policies = snapshot.refresh, snapshot.families, snapshot.authorizations, snapshot.devices, snapshot.policies
	state.events, state.rotationCount, state.compromiseEvents = snapshot.events, snapshot.rotationCount, snapshot.compromiseEvents
}

func (state *task13State) Exec(ctx context.Context, query string, arguments ...any) (pgconn.CommandTag, error) {
	return state.base.Exec(ctx, query, arguments...)
}

func (state *task13State) Query(ctx context.Context, query string, arguments ...any) (pgx.Rows, error) {
	if strings.Contains(query, "FROM deviceauth.device_refresh_tokens") && strings.Contains(query, "ORDER BY token_hash") && len(arguments) == 1 {
		familyID, ok := arguments[0].(uuid.UUID)
		if !ok {
			return nil, errors.New("unexpected refresh family argument")
		}
		rows := make([]store.DeviceauthDeviceRefreshToken, 0)
		for _, row := range state.refresh {
			if row.FamilyID == familyID {
				row.TokenHash = bytes.Clone(row.TokenHash)
				row.PreviousTokenHash = bytes.Clone(row.PreviousTokenHash)
				rows = append(rows, row)
			}
		}
		sort.Slice(rows, func(i, j int) bool { return bytes.Compare(rows[i].TokenHash, rows[j].TokenHash) < 0 })
		return &task13RotationRows{refresh: rows}, nil
	}
	return state.base.Query(ctx, query, arguments...)
}

func (state *task13State) QueryRow(ctx context.Context, query string, arguments ...any) pgx.Row {
	switch {
	case strings.Contains(query, "WHERE f.access_token_hash") && len(arguments) == 1:
		digest, ok := arguments[0].([]byte)
		if !ok {
			return task13StateRow{err: errors.New("unexpected access digest")}
		}
		for _, family := range state.families {
			if !bytes.Equal(family.AccessTokenHash, digest) {
				continue
			}
			authorization, authorizationFound := state.authorizations[family.AuthorizationID]
			device, deviceFound := state.devices[authorization.DeviceID]
			if !authorizationFound || !deviceFound {
				break
			}
			return task13StateRow{values: []any{family.ID, family.AuthorizationID, family.State, family.StateVersion,
				family.AccessExpiresAt, family.IdleExpiresAt, family.AbsoluteExpiresAt, authorization.PrincipalID, device.ID,
				authorization.State, authorization.StateVersion, authorization.ProvisionalUntil, device.State, device.HpkePublicKey, device.KeyVersion}}
		}
		return task13StateRow{err: pgx.ErrNoRows}
	case strings.Contains(query, "FROM deviceauth.device_token_families") && strings.Contains(query, "FOR UPDATE") && len(arguments) == 1:
		row, ok := state.families[arguments[0].(uuid.UUID)]
		if !ok {
			return task13StateRow{err: pgx.ErrNoRows}
		}
		return task13StateRow{values: []any{row.ID, row.AuthorizationID, row.State, row.StateVersion, row.AccessTokenHash,
			row.AccessExpiresAt, row.IdleExpiresAt, row.AbsoluteExpiresAt, row.CreatedAt, row.UpdatedAt}}
	case strings.Contains(query, "FROM deviceauth.device_authorizations") && strings.Contains(query, "FOR UPDATE") && len(arguments) == 1:
		row, ok := state.authorizations[arguments[0].(uuid.UUID)]
		if !ok {
			return task13StateRow{err: pgx.ErrNoRows}
		}
		return task13StateRow{values: []any{row.ID, row.PrincipalID, row.DeviceID, row.State, row.StateVersion, row.ProvisionalUntil, row.CreatedAt, row.UpdatedAt}}
	case strings.Contains(query, "FROM deviceauth.devices") && strings.Contains(query, "FOR UPDATE") && len(arguments) == 1:
		row, ok := state.devices[arguments[0].(uuid.UUID)]
		if !ok {
			return task13StateRow{err: pgx.ErrNoRows}
		}
		return task13StateRow{values: []any{row.ID, row.PrincipalID, row.DisplayNameCiphertext, row.DisplayNameKeyVersion,
			row.SigningPublicKey, row.HpkePublicKey, row.KeyVersion, row.State, row.CreatedAt, row.UpdatedAt}}
	case strings.Contains(query, "FROM deviceauth.device_policy_snapshots") && len(arguments) == 1:
		row, ok := state.policies[arguments[0].(uuid.UUID)]
		if !ok {
			return task13StateRow{err: pgx.ErrNoRows}
		}
		return task13StateRow{values: []any{row.AuthorizationID, row.SchemaVersion, row.Policy, row.CreatedAt}}
	default:
		return state.base.QueryRow(ctx, query, arguments...)
	}
}

type task13StateRow struct {
	values []any
	err    error
}

func (row task13StateRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	return task13AssignRotationValues(destinations, row.values)
}

type task13Repository struct {
	mu            sync.Mutex
	state         *task13State
	inTransaction bool
}

func (repository *task13Repository) WithinTransaction(ctx context.Context, operation func(context.Context, Transaction) error) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	snapshot := repository.state.snapshot()
	repository.inTransaction = true
	err := operation(ctx, &task13Transaction{task12Transaction: &task12Transaction{database: repository.state.base}, state: repository.state})
	repository.inTransaction = false
	if err != nil {
		repository.state.restore(snapshot)
	}
	return err
}

type task13Transaction struct {
	*task12Transaction
	state *task13State
}

func (transaction *task13Transaction) DBTX() store.DBTX { return transaction.state }

func (transaction *task13Transaction) DiscoverDeviceRefreshToken(_ context.Context, digest []byte) (store.DiscoverDeviceRefreshTokenRow, bool, error) {
	refresh, ok := transaction.state.refresh[string(digest)]
	if !ok {
		return store.DiscoverDeviceRefreshTokenRow{}, false, nil
	}
	family, ok := transaction.state.families[refresh.FamilyID]
	if !ok {
		return store.DiscoverDeviceRefreshTokenRow{}, false, nil
	}
	authorization, ok := transaction.state.authorizations[family.AuthorizationID]
	if !ok {
		return store.DiscoverDeviceRefreshTokenRow{}, false, nil
	}
	device, ok := transaction.state.devices[authorization.DeviceID]
	if !ok {
		return store.DiscoverDeviceRefreshTokenRow{}, false, nil
	}
	return store.DiscoverDeviceRefreshTokenRow{TokenHash: bytes.Clone(refresh.TokenHash), FamilyID: refresh.FamilyID,
		PreviousTokenHash: bytes.Clone(refresh.PreviousTokenHash), RefreshState: refresh.State, IssuedAt: refresh.IssuedAt, UsedAt: refresh.UsedAt, RevokedAt: refresh.RevokedAt,
		AuthorizationID: authorization.ID, FamilyState: family.State, FamilyStateVersion: family.StateVersion, AccessExpiresAt: family.AccessExpiresAt,
		IdleExpiresAt: family.IdleExpiresAt, AbsoluteExpiresAt: family.AbsoluteExpiresAt, PrincipalID: authorization.PrincipalID, DeviceID: device.ID,
		AuthorizationState: authorization.State, AuthorizationStateVersion: authorization.StateVersion, ProvisionalUntil: authorization.ProvisionalUntil,
		DeviceState: device.State, SigningPublicKey: bytes.Clone(device.SigningPublicKey), KeyVersion: device.KeyVersion}, true, nil
}

func (transaction *task13Transaction) LockDeviceFamilyRefreshTokens(ctx context.Context, familyID uuid.UUID) ([]store.DeviceauthDeviceRefreshToken, error) {
	return transaction.ListDeviceFamilyRefreshTokens(ctx, familyID)
}
func (transaction *task13Transaction) ListDeviceFamilyRefreshTokens(_ context.Context, familyID uuid.UUID) ([]store.DeviceauthDeviceRefreshToken, error) {
	rows := make([]store.DeviceauthDeviceRefreshToken, 0)
	for _, row := range transaction.state.refresh {
		if row.FamilyID == familyID {
			row.TokenHash = bytes.Clone(row.TokenHash)
			row.PreviousTokenHash = bytes.Clone(row.PreviousTokenHash)
			rows = append(rows, row)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return bytes.Compare(rows[i].TokenHash, rows[j].TokenHash) < 0 })
	return rows, nil
}
func (transaction *task13Transaction) ListDeviceAuthorizationFamilies(_ context.Context, authorizationID uuid.UUID) ([]store.DeviceauthDeviceTokenFamily, error) {
	rows := make([]store.DeviceauthDeviceTokenFamily, 0)
	for _, row := range transaction.state.families {
		if row.AuthorizationID == authorizationID {
			row.AccessTokenHash = bytes.Clone(row.AccessTokenHash)
			rows = append(rows, row)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return bytes.Compare(rows[i].ID[:], rows[j].ID[:]) < 0 })
	return rows, nil
}
func (transaction *task13Transaction) GetDeviceTokenFamilyForUpdate(_ context.Context, id uuid.UUID) (store.DeviceauthDeviceTokenFamily, bool, error) {
	row, ok := transaction.state.families[id]
	row.AccessTokenHash = bytes.Clone(row.AccessTokenHash)
	return row, ok, nil
}
func (transaction *task13Transaction) GetDeviceAuthorizationForUpdate(_ context.Context, id uuid.UUID) (store.DeviceauthDeviceAuthorization, bool, error) {
	row, ok := transaction.state.authorizations[id]
	return row, ok, nil
}
func (transaction *task13Transaction) DiscoverDeviceAuthorization(_ context.Context, deviceID uuid.UUID) (store.DeviceauthDeviceAuthorization, bool, error) {
	for _, row := range transaction.state.authorizations {
		if row.DeviceID == deviceID {
			return row, true, nil
		}
	}
	return store.DeviceauthDeviceAuthorization{}, false, nil
}
func (transaction *task13Transaction) GetDeviceForUpdate(_ context.Context, id uuid.UUID) (store.DeviceauthDevice, bool, error) {
	row, ok := transaction.state.devices[id]
	row.SigningPublicKey = bytes.Clone(row.SigningPublicKey)
	row.HpkePublicKey = bytes.Clone(row.HpkePublicKey)
	return row, ok, nil
}
func (*task13Transaction) GetDevicePolicySnapshot(context.Context, uuid.UUID) (store.DeviceauthDevicePolicySnapshot, bool, error) {
	return store.DeviceauthDevicePolicySnapshot{}, false, nil
}
func (transaction *task13Transaction) MarkDeviceRefreshUsed(_ context.Context, params store.MarkDeviceRefreshUsedParams) (bool, error) {
	key := string(params.TokenHash)
	row, ok := transaction.state.refresh[key]
	if !ok || row.State != "active" {
		return false, nil
	}
	row.State, row.UsedAt = "used", params.UsedAt
	transaction.state.refresh[key] = row
	return true, nil
}
func (transaction *task13Transaction) RotateDeviceFamilyAccess(_ context.Context, params store.RotateDeviceFamilyAccessParams) (store.DeviceauthDeviceTokenFamily, bool, error) {
	row, ok := transaction.state.families[params.ID]
	if !ok || row.State != "active" {
		return store.DeviceauthDeviceTokenFamily{}, false, nil
	}
	row.AccessTokenHash, row.AccessExpiresAt, row.IdleExpiresAt, row.UpdatedAt = bytes.Clone(params.AccessTokenHash), params.AccessExpiresAt, params.IdleExpiresAt, params.UpdatedAt
	row.StateVersion++
	transaction.state.families[params.ID] = row
	transaction.state.rotationCount++
	return row, true, nil
}
func (transaction *task13Transaction) InsertDeviceRefreshToken(_ context.Context, params store.InsertDeviceRefreshTokenParams) error {
	transaction.state.refresh[string(params.TokenHash)] = store.DeviceauthDeviceRefreshToken{TokenHash: bytes.Clone(params.TokenHash), FamilyID: params.FamilyID,
		PreviousTokenHash: bytes.Clone(params.PreviousTokenHash), State: "active", IssuedAt: params.IssuedAt}
	return nil
}
func (transaction *task13Transaction) CompromiseDeviceTokenFamily(_ context.Context, params store.CompromiseDeviceTokenFamilyParams) (store.DeviceauthDeviceTokenFamily, bool, error) {
	row, ok := transaction.state.families[params.ID]
	if !ok || row.State != "active" {
		return store.DeviceauthDeviceTokenFamily{}, false, nil
	}
	row.State, row.UpdatedAt = "compromised", params.UpdatedAt
	row.StateVersion++
	transaction.state.families[params.ID] = row
	return row, true, nil
}
func (transaction *task13Transaction) RevokeDeviceRefreshTokens(_ context.Context, params store.RevokeDeviceRefreshTokensParams) (int64, error) {
	var count int64
	for key, row := range transaction.state.refresh {
		if row.FamilyID == params.FamilyID && row.State == "active" {
			row.State, row.RevokedAt = "revoked", params.RevokedAt
			transaction.state.refresh[key] = row
			count++
		}
	}
	return count, nil
}
func (transaction *task13Transaction) RevokeDeviceTokenFamily(_ context.Context, params store.RevokeDeviceTokenFamilyParams) (int64, error) {
	row, ok := transaction.state.families[params.ID]
	if !ok || (row.State != "active" && row.State != "compromised") {
		return 0, nil
	}
	row.State = "revoked"
	row.StateVersion++
	row.UpdatedAt = params.UpdatedAt
	transaction.state.families[params.ID] = row
	return 1, nil
}
func (transaction *task13Transaction) RevokeDeviceAuthorization(_ context.Context, params store.RevokeDeviceAuthorizationParams) (int64, error) {
	row, ok := transaction.state.authorizations[params.ID]
	if !ok || row.State == "revoked" {
		return 0, nil
	}
	row.State = "revoked"
	row.StateVersion++
	row.UpdatedAt = params.UpdatedAt
	transaction.state.authorizations[params.ID] = row
	return 1, nil
}
func (transaction *task13Transaction) RevokeDeviceRecord(_ context.Context, params store.RevokeDeviceRecordParams) (int64, error) {
	row, ok := transaction.state.devices[params.ID]
	if !ok || row.State == "revoked" {
		return 0, nil
	}
	row.State, row.UpdatedAt = "revoked", params.UpdatedAt
	transaction.state.devices[params.ID] = row
	return 1, nil
}
func (transaction *task13Transaction) AppendEvent(_ context.Context, event *eventsv1.EventEnvelope) error {
	transaction.state.events = append(transaction.state.events, event)
	if event.GetEventType() == "talenro.deviceauth.token_family_compromised.v1" {
		transaction.state.compromiseEvents++
	}
	return nil
}

type task13IdentityParticipant struct {
	state                            *task13State
	accountActive, revocationAllowed bool
	replayRecords                    int
	revocationValidations            int
	revokedAuthorizations            []uuid.UUID
}

func (*task13IdentityParticipant) ValidateDeviceEnrollment(context.Context, store.DBTX, [32]byte, time.Time) (identity.DeviceEnrollmentAuthority, bool, error) {
	return identity.DeviceEnrollmentAuthority{}, false, errors.New("unused")
}
func (participant *task13IdentityParticipant) ValidateDeviceAccountAuthority(_ context.Context, dbtx store.DBTX, _ identity.PrincipalID, _ time.Time) (bool, error) {
	if !participant.ownsDBTX(dbtx) {
		return false, errors.New("wrong DBTX")
	}
	return participant.accountActive, nil
}
func (participant *task13IdentityParticipant) ValidateDeviceRevocation(_ context.Context, dbtx store.DBTX, request identity.DeviceRevocationRequest, _ time.Time) (identity.DeviceRevocationAuthority, bool, error) {
	participant.revocationValidations++
	if !participant.ownsDBTX(dbtx) || !participant.revocationAllowed {
		return identity.DeviceRevocationAuthority{}, false, nil
	}
	authority, err := identity.NewDeviceRevocationAuthority(request.PrincipalID, request.Reauthentication.SessionID)
	return authority, err == nil, err
}
func (participant *task13IdentityParticipant) RecordDeviceTokenReplay(_ context.Context, dbtx store.DBTX, _ identity.DeviceTokenReplaySecurityRecord, _ time.Time) error {
	if !participant.ownsDBTX(dbtx) {
		return errors.New("wrong DBTX")
	}
	participant.replayRecords++
	return nil
}
func (*task13IdentityParticipant) BindSessionToAuthorization(context.Context, store.DBTX, identity.SessionID, uuid.UUID, time.Time) error {
	return errors.New("unused")
}
func (participant *task13IdentityParticipant) RevokeAuthorizationSessions(_ context.Context, dbtx store.DBTX, authorizationID uuid.UUID, _ time.Time) error {
	if !participant.ownsDBTX(dbtx) {
		return errors.New("wrong DBTX")
	}
	participant.revokedAuthorizations = append(participant.revokedAuthorizations, authorizationID)
	return nil
}

func (participant *task13IdentityParticipant) ownsDBTX(dbtx store.DBTX) bool {
	if dbtx == participant.state {
		return true
	}
	barrier, ok := dbtx.(*task13BarrierDBTX)
	return ok && barrier.transaction != nil && barrier.transaction.state == participant.state
}

type task13ChallengeStore struct {
	mu           sync.Mutex
	records      map[string]ChallengeRecord
	repository   *task13Repository
	consumeCalls int
}

type task13WinnerThenAmbiguousChallengeStore struct {
	delegate          ChallengeStore
	service           *Service
	targetChallengeID string
	winner            RotateDeviceTokenCommand
	targetCalls       int
	winnerErr         error
}

func (store *task13WinnerThenAmbiguousChallengeStore) Create(ctx context.Context, record ChallengeRecord, ttl time.Duration) error {
	return store.delegate.Create(ctx, record, ttl)
}

func (store *task13WinnerThenAmbiguousChallengeStore) Consume(
	ctx context.Context,
	id string,
	grantDigest, contextDigest [32]byte,
) (ChallengeRecord, error) {
	if id != store.targetChallengeID {
		return store.delegate.Consume(ctx, id, grantDigest, contextDigest)
	}
	store.targetCalls++
	if store.targetCalls == 1 {
		tokens, err := store.service.RotateDeviceToken(ctx, store.winner)
		tokens.AccessToken.Clear()
		tokens.RefreshToken.Clear()
		store.winnerErr = err
	}
	return ChallengeRecord{}, ErrChallengeUnavailable
}

func (store *task13ChallengeStore) Create(_ context.Context, record ChallengeRecord, _ time.Duration) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.records[record.ChallengeID] = record
	return nil
}
func (store *task13ChallengeStore) Consume(_ context.Context, id string, grantDigest, contextDigest [32]byte) (ChallengeRecord, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.consumeCalls++
	record, ok := store.records[id]
	if !ok {
		return ChallengeRecord{}, ErrChallengeNotFound
	}
	delete(store.records, id)
	if !challengeDigestsMatch(record, grantDigest, contextDigest) {
		return ChallengeRecord{}, ErrChallengeNotFound
	}
	return record, nil
}

type task13Limiter struct{}

func (task13Limiter) Allow(context.Context, ratelimit.Operation, [32]byte, config.RateLimitPolicy) (bool, error) {
	return true, nil
}

func cloneTask13Refresh(source map[string]store.DeviceauthDeviceRefreshToken) map[string]store.DeviceauthDeviceRefreshToken {
	result := make(map[string]store.DeviceauthDeviceRefreshToken, len(source))
	for key, row := range source {
		row.TokenHash = bytes.Clone(row.TokenHash)
		row.PreviousTokenHash = bytes.Clone(row.PreviousTokenHash)
		result[key] = row
	}
	return result
}
func cloneTask13Families(source map[uuid.UUID]store.DeviceauthDeviceTokenFamily) map[uuid.UUID]store.DeviceauthDeviceTokenFamily {
	result := make(map[uuid.UUID]store.DeviceauthDeviceTokenFamily, len(source))
	for key, row := range source {
		row.AccessTokenHash = bytes.Clone(row.AccessTokenHash)
		result[key] = row
	}
	return result
}
func cloneTask13Authorizations(source map[uuid.UUID]store.DeviceauthDeviceAuthorization) map[uuid.UUID]store.DeviceauthDeviceAuthorization {
	result := make(map[uuid.UUID]store.DeviceauthDeviceAuthorization, len(source))
	for key, row := range source {
		result[key] = row
	}
	return result
}
func cloneTask13Devices(source map[uuid.UUID]store.DeviceauthDevice) map[uuid.UUID]store.DeviceauthDevice {
	result := make(map[uuid.UUID]store.DeviceauthDevice, len(source))
	for key, row := range source {
		row.SigningPublicKey = bytes.Clone(row.SigningPublicKey)
		row.HpkePublicKey = bytes.Clone(row.HpkePublicKey)
		result[key] = row
	}
	return result
}
func cloneTask13Policies(source map[uuid.UUID]store.DeviceauthDevicePolicySnapshot) map[uuid.UUID]store.DeviceauthDevicePolicySnapshot {
	result := make(map[uuid.UUID]store.DeviceauthDevicePolicySnapshot, len(source))
	for key, row := range source {
		row.Policy = bytes.Clone(row.Policy)
		result[key] = row
	}
	return result
}

var _ Repository = (*task13Repository)(nil)
var _ deviceTokenTransaction = (*task13Transaction)(nil)
var _ identity.DeviceTransactionParticipant = (*task13IdentityParticipant)(nil)
var _ ChallengeStore = (*task13ChallengeStore)(nil)
var _ ratelimit.Limiter = task13Limiter{}
var _ store.DBTX = (*task13State)(nil)
