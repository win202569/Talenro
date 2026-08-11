package deviceauth

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

	rotation := command
	rotation.Kind = ChallengeRotation
	rotation.RefreshToken = secret.NewBytes(bytes.Repeat([]byte{0x44}, 32))
	rotation.EnrollmentGrant = secret.Bytes{}
	if _, err := fixture.application.CreateChallenge(context.Background(), rotation); publicTask12Code(err) != apierrors.ActionNotAllowed {
		t.Fatalf("Task13 rotation branch error = %v", err)
	}
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
	mu            sync.Mutex
	database      *task12Database
	inTransaction atomic.Bool
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
	repository.database.record("commit")
	return nil
}

type task12Transaction struct{ database *task12Database }

func (transaction *task12Transaction) DBTX() store.DBTX { return transaction.database }
func (transaction *task12Transaction) BeginIdempotency(ctx context.Context, scope idempotency.Scope, key string, canonical []byte, createdAt, expiresAt time.Time) (idempotency.Record, idempotency.Outcome, error) {
	transaction.database.record("begin_idempotency")
	return transaction.database.idempotency.Begin(ctx, scope, key, canonical, createdAt, expiresAt)
}
func (transaction *task12Transaction) CompleteIdempotency(ctx context.Context, record idempotency.Record, status int, body []byte) error {
	transaction.database.record("complete_idempotency")
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
	transaction.database.familyCount++
	transaction.database.family = params
	transaction.database.family.AccessTokenHash = bytes.Clone(params.AccessTokenHash)
	return nil
}
func (transaction *task12Transaction) InsertDeviceRefreshToken(_ context.Context, params store.InsertDeviceRefreshTokenParams) error {
	transaction.database.record("insert_refresh")
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
	repository         *task12Repository
	idempotencyDB      *task12IdempotencyDB
	idempotency        idempotency.Repository
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
	operations         []string
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
	return database.idempotencyDB.Query(ctx, query, arguments...)
}
func (database *task12Database) QueryRow(ctx context.Context, query string, arguments ...any) pgx.Row {
	return database.idempotencyDB.QueryRow(ctx, query, arguments...)
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
