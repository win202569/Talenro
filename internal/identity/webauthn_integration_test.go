//go:build integration

package identity

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
	"talenro.local/platform/internal/testinfra"
)

func TestConcurrentPasskeyRegistrationEnforcesMaximumAtFinalAuthorityBoundary(t *testing.T) {
	pool := testinfra.OpenMigratedPostgres(t)
	protector, err := sensitive.NewLocal(
		secret.NewBytes(bytes.Repeat([]byte{0xe1}, 32)),
		secret.NewBytes(bytes.Repeat([]byte{0xe2}, 32)),
		13,
	)
	if err != nil {
		t.Fatal("create passkey integration protector")
	}
	t.Cleanup(func() { _ = protector.Close() })
	postgresRepository, err := NewPostgresRepository(pool, protector)
	if err != nil {
		t.Fatal("create passkey integration repository")
	}

	runID := uuid.NewString()
	principalID := uuid.New()
	sessionID := uuid.New()
	task11CleanupPasskeyIntegrationIdentity(t, pool, principalID)
	task11SeedPasskeyIntegrationIdentity(t, pool, runID, principalID, sessionID)

	commands := make([]FinishPasskeyRegistrationCommand, 2)
	executors := make([]*fakeChallengeExecutor, 2)
	for index := range commands {
		executor := &fakeChallengeExecutor{setAllowed: true}
		ceremonies, storeErr := newRedisChallengeStoreWithExecutor(executor, 250*time.Millisecond)
		if storeErr != nil {
			t.Fatal("create passkey integration ceremony store")
		}
		beginApplication := task11PasskeyIntegrationApplication(t, postgresRepository, protector, ceremonies)
		options, beginErr := beginApplication.BeginPasskeyRegistration(context.Background(), BeginPasskeyRegistrationCommand{
			PrincipalID:      PrincipalID(principalID.String()),
			DisplayName:      "Concurrent passkey",
			Reauthentication: task11PasswordReauthentication(sessionID),
			IdempotencyKey:   "passkey-integration-begin-" + runID + "-" + string(rune('a'+index)),
		})
		if beginErr != nil {
			t.Fatal("begin passkey integration registration")
		}
		var envelope passkeyRegistrationOptions
		if jsonErr := json.Unmarshal(options, &envelope); jsonErr != nil || envelope.CeremonyID == "" || len(envelope.PublicKey.Challenge) == 0 {
			t.Fatal("decode passkey integration options")
		}
		record, decodeErr := decodeWebAuthnCeremony(
			envelope.CeremonyID,
			WebAuthnRegistrationCeremony,
			executor.setValue,
		)
		if decodeErr != nil {
			t.Fatal("decode passkey integration ceremony")
		}
		executor.runValues = []any{bytes.Clone(executor.setValue)}
		credentialID := task11PasskeyIntegrationDigest(runID, "new-credential", index)
		commands[index] = FinishPasskeyRegistrationCommand{
			PrincipalID:      PrincipalID(principalID.String()),
			CeremonyID:       envelope.CeremonyID,
			Response:         task11RegistrationResponse(t, record.Session.Challenge, "https://login.example.test", credentialID),
			Reauthentication: task11PasswordReauthentication(sessionID),
			IdempotencyKey:   "passkey-integration-finish-" + runID + "-" + string(rune('a'+index)),
		}
		executors[index] = executor
	}
	if commands[0].CeremonyID == commands[1].CeremonyID || bytes.Equal(
		task11PasskeyIntegrationDigest(runID, "new-credential", 0),
		task11PasskeyIntegrationDigest(runID, "new-credential", 1),
	) || commands[0].IdempotencyKey == commands[1].IdempotencyKey {
		t.Fatal("concurrent passkey requests are not independent")
	}

	coordination := &task11PasskeyLockCoordination{
		leaderListReturned: make(chan struct{}),
		allowLeaderReturn:  make(chan struct{}),
		followerPID:        make(chan int32, 1),
		firstLists:         make(chan task11PasskeyFirstList, 2),
	}
	var allowLeaderOnce sync.Once
	operationContext, cancelOperations := context.WithCancel(context.Background())
	var workers sync.WaitGroup
	t.Cleanup(func() {
		cancelOperations()
		allowLeaderOnce.Do(func() { close(coordination.allowLeaderReturn) })
		workers.Wait()
	})
	repositories := make([]*task11PasskeyFinalBarrierRepository, 2)
	applications := make([]*Service, 2)
	for index := range applications {
		repositories[index] = &task11PasskeyFinalBarrierRepository{
			inner: postgresRepository, role: task11PasskeyBarrierRole(index), coordination: coordination,
		}
		ceremonies, storeErr := newRedisChallengeStoreWithExecutor(executors[index], 250*time.Millisecond)
		if storeErr != nil {
			t.Fatal("recreate passkey integration ceremony store")
		}
		applications[index] = task11PasskeyIntegrationApplication(t, repositories[index], protector, ceremonies)
	}

	type outcome struct {
		index int
		err   error
	}
	outcomes := make(chan outcome, 2)
	start := make(chan struct{})
	for index := range applications {
		index := index
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			finishErr := applications[index].FinishPasskeyRegistration(operationContext, commands[index])
			outcomes <- outcome{index: index, err: finishErr}
		}()
	}
	close(start)
	select {
	case <-coordination.leaderListReturned:
	case result := <-outcomes:
		t.Fatalf("concurrent passkey request %d completed before the leader acquired the initial active-passkey locks: %v", result.index, result.err)
	case <-time.After(12 * time.Second):
		t.Fatal("leader did not acquire the initial active-passkey locks")
	}
	var followerPID int32
	select {
	case followerPID = <-coordination.followerPID:
	case result := <-outcomes:
		t.Fatalf("concurrent passkey request %d completed before the follower entered its final authority transaction: %v", result.index, result.err)
	case <-time.After(12 * time.Second):
		t.Fatal("follower did not enter its final authority transaction")
	}
	// The leader has returned the real production ListActivePasskeys result and
	// still owns its password and passkey row locks. The follower has entered the
	// same production operation. Observing its own backend waiting on the earlier
	// password-row lock proves that final-authority serialization is enforced
	// before the follower can list active passkeys.
	task11AwaitPasskeyIntegrationLock(t, pool, followerPID)
	allowLeaderOnce.Do(func() { close(coordination.allowLeaderReturn) })

	successes := 0
	conflicts := 0
	for range 2 {
		select {
		case result := <-outcomes:
			if result.err == nil {
				successes++
				continue
			}
			if publicTask9Code(result.err) != apierrors.StateConflict {
				t.Fatalf("concurrent passkey request %d returned a non-conflict public code", result.index)
			}
			conflicts++
			lower := strings.ToLower(result.err.Error())
			for _, raw := range []string{"sqlstate", "duplicate key", "passkey_credentials", "pgx", "postgres"} {
				if strings.Contains(lower, raw) {
					t.Fatalf("concurrent passkey conflict exposed raw database text %q", raw)
				}
			}
		case <-time.After(12 * time.Second):
			t.Fatal("concurrent passkey registrations did not complete")
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent passkey outcomes = %d success / %d conflict", successes, conflicts)
	}
	for index, repository := range repositories {
		if repository.calls.Load() != 2 {
			t.Fatalf("passkey request %d used %d transactions, want preflight and final", index, repository.calls.Load())
		}
	}
	firstListCounts := make(map[task11PasskeyBarrierRole]int, 2)
	expectedFirstListCounts := map[task11PasskeyBarrierRole]int{
		task11PasskeyLeader:   9,
		task11PasskeyFollower: 10,
	}
	for range 2 {
		select {
		case firstList := <-coordination.firstLists:
			expectedCount, ok := expectedFirstListCounts[firstList.role]
			if !ok {
				t.Fatalf("unexpected final-authority passkey list role %d", firstList.role)
			}
			if firstList.count != expectedCount {
				t.Fatalf("%s initial final-authority passkey list length = %d, want %d", firstList.role, firstList.count, expectedCount)
			}
			firstListCounts[firstList.role]++
		case <-time.After(time.Second):
			t.Fatal("final-authority passkey list result was not recorded")
		}
	}
	if firstListCounts[task11PasskeyLeader] != 1 || firstListCounts[task11PasskeyFollower] != 1 {
		t.Fatal("leader and follower initial passkey lists were not both observed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var activeCount int
	if queryErr := pool.QueryRow(ctx, `SELECT count(*) FROM identity.passkey_credentials WHERE principal_id=$1 AND state='active'`, principalID).Scan(&activeCount); queryErr != nil {
		t.Fatal("count final active passkeys")
	}
	if activeCount != 10 {
		t.Fatalf("final active passkey count = %d, want 10", activeCount)
	}
	var newCount int
	if queryErr := pool.QueryRow(ctx, `
SELECT count(*) FROM identity.passkey_credentials
WHERE principal_id=$1 AND credential_id=ANY($2::bytea[])`, principalID, [][]byte{
		task11PasskeyIntegrationDigest(runID, "new-credential", 0),
		task11PasskeyIntegrationDigest(runID, "new-credential", 1),
	}).Scan(&newCount); queryErr != nil {
		t.Fatal("count newly registered passkeys")
	}
	if newCount != 1 {
		t.Fatalf("newly registered passkey count = %d, want 1", newCount)
	}
}

type task11PasskeyFinalBarrierRepository struct {
	inner        Repository
	role         task11PasskeyBarrierRole
	coordination *task11PasskeyLockCoordination
	calls        atomic.Int32
}

func (repository *task11PasskeyFinalBarrierRepository) WithinTransaction(
	ctx context.Context,
	operation func(context.Context, Transaction) error,
) error {
	call := repository.calls.Add(1)
	return repository.inner.WithinTransaction(ctx, func(transactionContext context.Context, transaction Transaction) error {
		if call != 2 {
			return operation(transactionContext, transaction)
		}
		passkeys, ok := transaction.(passkeyTransaction)
		if !ok || nilIdentityValue(passkeys) {
			return ErrRepository
		}
		if repository.role == task11PasskeyFollower {
			select {
			case <-repository.coordination.leaderListReturned:
			case <-transactionContext.Done():
				return ErrRepository
			}
			var backendPID int32
			if err := passkeys.DBTX().QueryRow(transactionContext, `SELECT pg_backend_pid()`).Scan(&backendPID); err != nil || backendPID <= 0 {
				return ErrRepository
			}
			select {
			case repository.coordination.followerPID <- backendPID:
			case <-transactionContext.Done():
				return ErrRepository
			}
		}
		return operation(transactionContext, &task11PasskeyFinalBarrierTransaction{
			passkeyTransaction: passkeys,
			role:               repository.role,
			coordination:       repository.coordination,
		})
	})
}

type task11PasskeyBarrierRole int

const (
	task11PasskeyLeader task11PasskeyBarrierRole = iota
	task11PasskeyFollower
)

func (role task11PasskeyBarrierRole) String() string {
	if role == task11PasskeyLeader {
		return "leader"
	}
	return "follower"
}

type task11PasskeyLockCoordination struct {
	leaderListReturned chan struct{}
	allowLeaderReturn  chan struct{}
	followerPID        chan int32
	firstLists         chan task11PasskeyFirstList
}

type task11PasskeyFirstList struct {
	role  task11PasskeyBarrierRole
	count int
}

type task11PasskeyFinalBarrierTransaction struct {
	passkeyTransaction
	role         task11PasskeyBarrierRole
	coordination *task11PasskeyLockCoordination
}

func (transaction *task11PasskeyFinalBarrierTransaction) ListActivePasskeys(
	ctx context.Context,
	principalID uuid.UUID,
) ([]store.IdentityPasskeyCredential, error) {
	rows, err := transaction.passkeyTransaction.ListActivePasskeys(ctx, principalID)
	if transaction.coordination != nil {
		transaction.coordination.firstLists <- task11PasskeyFirstList{role: transaction.role, count: len(rows)}
		coordination := transaction.coordination
		transaction.coordination = nil
		if transaction.role == task11PasskeyLeader {
			close(coordination.leaderListReturned)
			select {
			case <-coordination.allowLeaderReturn:
			case <-ctx.Done():
				return nil, ErrRepository
			}
		}
	}
	return rows, err
}

func task11PasskeyIntegrationApplication(
	t *testing.T,
	repository Repository,
	protector sensitive.Protector,
	ceremonies ChallengeStore,
) *Service {
	t.Helper()
	application, err := newApplicationForTest(ApplicationDependencies{
		Repository:     repository,
		Protector:      protector,
		Random:         rand.Reader,
		Clock:          task9IntegrationClock{},
		Limiter:        task9IntegrationLimiter{},
		ChallengeStore: ceremonies,
		RateLimitKey:   secret.NewBytes(bytes.Repeat([]byte{0xe3}, 32)),
		Security: config.SecurityConfig{
			Profile:            config.ProfileTest,
			EmailVerification:  config.EmailDisabled,
			PublicBaseURL:      "https://api.example.test",
			RequestDeadline:    10 * time.Second,
			RedisTimeout:       250 * time.Millisecond,
			WebAuthnRPID:       "login.example.test",
			WebAuthnOrigins:    []string{"https://login.example.test"},
			LoginRateLimit:     config.RateLimitPolicy{Limit: 10, Window: 15 * time.Minute},
			DeliveryRateLimit:  config.RateLimitPolicy{Limit: 5, Window: time.Hour},
			ChallengeRateLimit: config.RateLimitPolicy{Limit: 20, Window: 5 * time.Minute},
		},
		DeviceAuthorizationParticipant: task9IntegrationParticipant{},
	}, func(_ []byte, _ []byte, policy PasswordPolicy) []byte {
		return bytes.Repeat([]byte{0xd1}, int(policy.TagBytes))
	}, uuid.New)
	if err != nil {
		t.Fatal("create passkey integration application")
	}
	return application
}

func task11SeedPasskeyIntegrationIdentity(
	t *testing.T,
	pool *pgxpool.Pool,
	runID string,
	principalID uuid.UUID,
	sessionID uuid.UUID,
) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal("begin passkey integration seed")
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	now := time.Now().UTC().Add(-time.Minute)
	accessHash := task11PasskeyIntegrationDigest(runID, "access-token", 0)
	statements := []struct {
		query     string
		arguments []any
	}{
		{`INSERT INTO identity.accounts (id,state,state_version,locale,created_at,updated_at) VALUES ($1,'active',1,'en',$2,$2)`, []any{principalID, now}},
		{`INSERT INTO identity.password_credentials (principal_id,policy_version,memory_kib,time_cost,parallelism,salt,password_hash,updated_at) VALUES ($1,1,65536,3,4,$2,$3,$4)`, []any{principalID, bytes.Repeat([]byte{0xd0}, 16), bytes.Repeat([]byte{0xd1}, 32), now}},
		{`INSERT INTO identity.account_sessions (id,principal_id,state,state_version,client_signing_public_key,access_token_hash,access_expires_at,absolute_expires_at,created_at,updated_at) VALUES ($1,$2,'active',1,$3,$4,$5,$6,$7,$7)`, []any{sessionID, principalID, bytes.Repeat([]byte{0xd2}, 32), accessHash, now.Add(2 * time.Hour), now.Add(24 * time.Hour), now}},
	}
	for _, statement := range statements {
		if _, execErr := tx.Exec(ctx, statement.query, statement.arguments...); execErr != nil {
			t.Fatal("seed passkey integration identity graph")
		}
	}
	for index := range 9 {
		credentialID := task11PasskeyIntegrationDigest(runID, "seed-credential", index)
		publicKey := task11PasskeyIntegrationDigest(runID, "seed-public-key", index)
		createdAt := now.Add(time.Duration(index) * time.Nanosecond)
		if _, execErr := tx.Exec(ctx, `
INSERT INTO identity.passkey_credentials
  (credential_id,principal_id,public_key,attestation_format,transports,protocol_flags,sign_count,state,created_at,updated_at)
VALUES ($1,$2,$3,'none',$4,$5,0,'active',$6,$6)`,
			credentialID, principalID, publicKey, []string{"internal"},
			int16(protocol.FlagUserPresent|protocol.FlagUserVerified), createdAt,
		); execErr != nil {
			t.Fatal("seed active passkey")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal("commit passkey integration seed")
	}
}

func task11CleanupPasskeyIntegrationIdentity(t *testing.T, pool *pgxpool.Pool, principalID uuid.UUID) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		for _, cleanup := range []struct {
			query    string
			argument any
		}{
			{`DELETE FROM idempotency_records WHERE principal_scope=$1`, "principal:" + principalID.String() + ":passkey"},
			{`DELETE FROM identity.passkey_credentials WHERE principal_id=$1`, principalID},
			{`DELETE FROM identity.account_sessions WHERE principal_id=$1`, principalID},
			{`DELETE FROM identity.password_credentials WHERE principal_id=$1`, principalID},
			{`DELETE FROM identity.accounts WHERE id=$1`, principalID},
		} {
			if _, err := pool.Exec(ctx, cleanup.query, cleanup.argument); err != nil {
				t.Error("clean passkey integration identity")
			}
		}
	})
}

func task11PasskeyIntegrationDigest(runID, domain string, index int) []byte {
	digest := sha256.Sum256([]byte(runID + ":" + domain + ":" + strconv.Itoa(index)))
	return bytes.Clone(digest[:])
}

func task11AwaitPasskeyIntegrationLock(t *testing.T, pool *pgxpool.Pool, backendPID int32) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	for {
		var waitEventType string
		var waitEvent string
		if err := pool.QueryRow(ctx, `
SELECT COALESCE(wait_event_type,''), COALESCE(wait_event,'')
FROM pg_stat_activity WHERE pid=$1`, backendPID).Scan(&waitEventType, &waitEvent); err != nil {
			t.Fatal("observe follower PostgreSQL lock wait")
		}
		if waitEventType == "Lock" && waitEvent != "" {
			return
		}
		select {
		case <-ticker.C:
		case <-deadline.C:
			t.Fatal("follower PostgreSQL backend did not enter a lock wait")
		case <-ctx.Done():
			t.Fatal("follower PostgreSQL lock observation timed out")
		}
	}
}
