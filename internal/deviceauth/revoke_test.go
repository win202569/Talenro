package deviceauth

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/store"
)

func TestRevokeDeviceSelectsOnlyOneAuthorizationGraph(t *testing.T) {
	t.Parallel()

	principalID := uuid.MustParse("a6493384-9407-4ad9-b220-7f3b49ef9054")
	deviceID := uuid.MustParse("f353613c-d08f-4141-b287-b37db9fb6f8e")
	authorizationID := uuid.MustParse("fa01e838-cab9-414e-8f03-ed0418cd4f20")
	familyID := uuid.MustParse("0ff820a5-5022-48e6-8867-77761f8e2f07")
	authorization := store.DeviceauthDeviceAuthorization{
		ID: authorizationID, PrincipalID: principalID, DeviceID: deviceID, State: "active",
	}
	device := store.DeviceauthDevice{ID: deviceID, PrincipalID: principalID, State: "active"}
	families := []store.DeviceauthDeviceTokenFamily{{ID: familyID, AuthorizationID: authorizationID, State: "active"}}
	refresh := []store.DeviceauthDeviceRefreshToken{{TokenHash: make([]byte, 32), FamilyID: familyID, State: "active"}}
	if !validDeviceRevocationGraph(principalID, deviceID, authorization, device, families, refresh) {
		t.Fatal("exact device authorization graph was rejected")
	}

	otherDevice := device
	otherDevice.ID = uuid.New()
	if validDeviceRevocationGraph(principalID, deviceID, authorization, otherDevice, families, refresh) {
		t.Fatal("revocation accepted a different device")
	}
	otherFamily := families[0]
	otherFamily.AuthorizationID = uuid.New()
	if validDeviceRevocationGraph(principalID, deviceID, authorization, device, []store.DeviceauthDeviceTokenFamily{otherFamily}, refresh) {
		t.Fatal("revocation accepted another authorization's token family")
	}
	otherRefresh := refresh[0]
	otherRefresh.FamilyID = uuid.New()
	if validDeviceRevocationGraph(principalID, deviceID, authorization, device, families, []store.DeviceauthDeviceRefreshToken{otherRefresh}) {
		t.Fatal("revocation accepted another device's refresh token")
	}
}

func TestRevokeDeviceAtomicallySelectsOneDeviceAndReplaysBeforeCurrentAuthority(t *testing.T) {
	fixture := newTask13Fixture(t)
	otherDeviceID := uuid.MustParse("a2afc4a7-9c55-4c5c-a6f8-2ced6f9f9e11")
	otherAuthorizationID := uuid.MustParse("29be9502-d6d1-45b6-a2d1-316aeac6db91")
	otherFamilyID := uuid.MustParse("4fc5e6cf-ece7-46c9-ad57-614f6bd83bf0")
	fixture.state.devices[otherDeviceID] = store.DeviceauthDevice{ID: otherDeviceID, PrincipalID: fixture.principalID, State: "active", KeyVersion: 1}
	fixture.state.authorizations[otherAuthorizationID] = store.DeviceauthDeviceAuthorization{ID: otherAuthorizationID, PrincipalID: fixture.principalID,
		DeviceID: otherDeviceID, State: "active", StateVersion: 1}
	fixture.state.families[otherFamilyID] = store.DeviceauthDeviceTokenFamily{ID: otherFamilyID, AuthorizationID: otherAuthorizationID, State: "active", StateVersion: 1}
	otherDigest := bytes.Repeat([]byte{0x91}, 32)
	fixture.state.refresh[string(otherDigest)] = store.DeviceauthDeviceRefreshToken{TokenHash: bytes.Clone(otherDigest), FamilyID: otherFamilyID, State: "active"}
	command := fixture.revocationCommand("task13-device-revocation-0001")
	defer command.Reauthentication.Proof.Clear()

	if err := fixture.application.RevokeDevice(context.Background(), command); err != nil {
		t.Fatalf("revoke device: %v", err)
	}
	if fixture.state.devices[fixture.deviceID].State != "revoked" || fixture.state.authorizations[fixture.authorizationID].State != "revoked" ||
		fixture.state.families[fixture.familyID].State != "revoked" || fixture.identity.revokedAuthorizations[len(fixture.identity.revokedAuthorizations)-1] != fixture.authorizationID {
		t.Fatal("target device authority graph or same-DBTX identity sessions were not revoked")
	}
	if fixture.state.devices[otherDeviceID].State != "active" || fixture.state.authorizations[otherAuthorizationID].State != "active" ||
		fixture.state.families[otherFamilyID].State != "active" || fixture.state.refresh[string(otherDigest)].State != "active" {
		t.Fatal("device revocation changed another device's authority graph")
	}
	if len(fixture.state.events) != 1 || fixture.state.events[0].GetEventType() != "talenro.deviceauth.authorization_changed.v1" {
		t.Fatal("device revocation did not append the bounded authorization event")
	}
	if fixture.state.events[0].GetIdempotencyKey() == command.IdempotencyKey {
		t.Fatal("revocation event exposed the caller idempotency key")
	}
	if parsed, err := uuid.Parse(fixture.state.events[0].GetIdempotencyKey()); err != nil || parsed == uuid.Nil || parsed.String() != fixture.state.events[0].GetIdempotencyKey() {
		t.Fatal("revocation event did not use an approved opaque synthetic deduplication key")
	}
	validations := fixture.identity.revocationValidations
	fixture.identity.revocationAllowed = false
	if err := fixture.application.RevokeDevice(context.Background(), command); err != nil {
		t.Fatalf("completed revocation replay depended on current authority: %v", err)
	}
	if fixture.identity.revocationValidations != validations {
		t.Fatal("completed revocation replay reached identity authority validation")
	}
}

func TestRevokeDeviceRejectsIdentityAuthorityWithoutMutation(t *testing.T) {
	fixture := newTask13Fixture(t)
	fixture.identity.revocationAllowed = false
	command := fixture.revocationCommand("task13-device-revocation-denied-0001")
	defer command.Reauthentication.Proof.Clear()

	err := fixture.application.RevokeDevice(context.Background(), command)
	if publicTask12Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("revocation authority failure = %v, want finite authentication failure", err)
	}
	if fixture.state.devices[fixture.deviceID].State != "active" || fixture.state.authorizations[fixture.authorizationID].State != "active" ||
		fixture.state.families[fixture.familyID].State != "active" || len(fixture.state.events) != 0 {
		t.Fatal("failed revocation mutated device authority")
	}
}

func TestDeviceRevocationRaceBlocksOnTheAccessResolversRefreshPrelock(t *testing.T) {
	fixture := newTask13Fixture(t)
	barrier := newTask13BarrierRepository(fixture.state)
	fixture.application.repository = barrier
	t.Cleanup(barrier.release)
	command := fixture.revocationCommand("task13-device-revocation-race-0001")
	defer command.Reauthentication.Proof.Clear()
	authorizeResult := make(chan error, 1)
	revokeResult := make(chan error, 1)
	go func() {
		authority, err := fixture.application.AuthorizeBundle(context.Background(), AuthorizeBundleQuery{AccessToken: fixture.access})
		clear(authority.Policy)
		authorizeResult <- err
	}()
	select {
	case <-barrier.firstAuthorityLock:
	case <-time.After(time.Second):
		t.Fatal("access resolution never acquired the shared refresh-token prelock")
	}
	go func() {
		revokeResult <- fixture.application.RevokeDevice(context.Background(), command)
	}()
	select {
	case revokeErr := <-revokeResult:
		barrier.release()
		<-authorizeResult
		t.Fatalf("revocation bypassed the access resolver's refresh-token lock: %v", revokeErr)
	case <-time.After(50 * time.Millisecond):
	}
	barrier.release()
	authorizeErr, revokeErr := <-authorizeResult, <-revokeResult
	if authorizeErr != nil {
		t.Fatalf("prelocked access resolution: %v", authorizeErr)
	}
	if revokeErr != nil {
		t.Fatalf("concurrent revocation: %v", revokeErr)
	}
	if _, err := fixture.application.AuthorizeBundle(context.Background(), AuthorizeBundleQuery{AccessToken: fixture.access}); publicTask12Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("post-revocation resolve = %v, want authentication failure", err)
	}
}

// task13BarrierRepository models the shared PostgreSQL refresh-row lock without
// serializing transaction entry. The test therefore fails if either production
// path stops acquiring the stable refresh-token prelock.
type task13BarrierRepository struct {
	state              *task13State
	authorityLock      sync.Mutex
	pauseOnce          sync.Once
	releaseOnce        sync.Once
	firstAuthorityLock chan struct{}
	releaseAuthority   chan struct{}
}

func newTask13BarrierRepository(state *task13State) *task13BarrierRepository {
	return &task13BarrierRepository{state: state, firstAuthorityLock: make(chan struct{}), releaseAuthority: make(chan struct{})}
}

func (repository *task13BarrierRepository) WithinTransaction(ctx context.Context, operation func(context.Context, Transaction) error) error {
	transaction := &task13BarrierTransaction{task13Transaction: &task13Transaction{
		task12Transaction: &task12Transaction{database: repository.state.base}, state: repository.state,
	}, repository: repository}
	transaction.dbtx = &task13BarrierDBTX{transaction: transaction}
	defer transaction.unlock()
	return operation(ctx, transaction)
}

func (repository *task13BarrierRepository) release() {
	repository.releaseOnce.Do(func() { close(repository.releaseAuthority) })
}

type task13BarrierTransaction struct {
	*task13Transaction
	repository *task13BarrierRepository
	dbtx       *task13BarrierDBTX
	locked     bool
}

func (transaction *task13BarrierTransaction) DBTX() store.DBTX { return transaction.dbtx }

func (transaction *task13BarrierTransaction) lockAuthority() {
	if transaction.locked {
		return
	}
	transaction.repository.authorityLock.Lock()
	transaction.locked = true
	transaction.repository.pauseOnce.Do(func() {
		close(transaction.repository.firstAuthorityLock)
		<-transaction.repository.releaseAuthority
	})
}

func (transaction *task13BarrierTransaction) unlock() {
	if transaction.locked {
		transaction.repository.authorityLock.Unlock()
		transaction.locked = false
	}
}

func (transaction *task13BarrierTransaction) LockDeviceFamilyRefreshTokens(ctx context.Context, familyID uuid.UUID) ([]store.DeviceauthDeviceRefreshToken, error) {
	transaction.lockAuthority()
	return transaction.task13Transaction.LockDeviceFamilyRefreshTokens(ctx, familyID)
}

type task13BarrierDBTX struct{ transaction *task13BarrierTransaction }

func (database *task13BarrierDBTX) Exec(ctx context.Context, query string, arguments ...any) (pgconn.CommandTag, error) {
	return database.transaction.state.Exec(ctx, query, arguments...)
}

func (database *task13BarrierDBTX) Query(ctx context.Context, query string, arguments ...any) (pgx.Rows, error) {
	if strings.Contains(query, "FROM deviceauth.device_refresh_tokens") && strings.Contains(query, "FOR UPDATE") {
		database.transaction.lockAuthority()
	}
	return database.transaction.state.Query(ctx, query, arguments...)
}

func (database *task13BarrierDBTX) QueryRow(ctx context.Context, query string, arguments ...any) pgx.Row {
	return database.transaction.state.QueryRow(ctx, query, arguments...)
}

func (fixture *task13Fixture) revocationCommand(key string) RevokeDeviceCommand {
	return RevokeDeviceCommand{AccountPrincipal: identity.PrincipalID(fixture.principalID.String()), DeviceID: fixture.deviceID,
		Reauthentication: identity.Reauthentication{SessionID: identity.SessionID("46434e83-f7e2-43c1-b7c5-80a3b49eec14"), Method: identity.ReauthPassword,
			Proof: secret.NewBytes([]byte("task13-current-password-proof"))}, IdempotencyKey: key}
}
