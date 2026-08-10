//go:build integration

package identity

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/ratelimit"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
	"talenro.local/platform/internal/testinfra"
)

func TestConcurrentRegistrationAndVerificationReplayUseDatabaseLocks(t *testing.T) {
	pool := testinfra.OpenMigratedPostgres(t)
	protector, err := sensitive.NewLocal(
		secret.NewBytes(bytes.Repeat([]byte{0xb1}, 32)), secret.NewBytes(bytes.Repeat([]byte{0xb2}, 32)), 11,
	)
	if err != nil {
		t.Fatal("create integration protector")
	}
	t.Cleanup(func() { _ = protector.Close() })
	postgresRepository, err := NewPostgresRepository(pool, protector)
	if err != nil {
		t.Fatal("create integration repository")
	}

	lookupReady := make(chan struct{}, 2)
	lookupRelease := make(chan struct{})
	registerApplication := task9IntegrationApplication(t, &task9LookupBarrierRepository{
		inner: postgresRepository, ready: lookupReady, release: lookupRelease,
	}, protector)
	email := "Concurrent-" + uuid.NewString() + "@Example.test"
	canonicalEmail, err := CanonicalizeEmail(email)
	if err != nil {
		t.Fatal("canonicalize integration email")
	}
	emailBytes := canonicalEmail.Bytes()
	defer clear(emailBytes)
	lookupDigest := protector.LookupDigest(emailFieldDomain, emailBytes)
	if lookupDigest == [32]byte{} {
		t.Fatal("protect integration lookup")
	}

	runID := uuid.NewString()
	registrationKeys := []string{"registration-alpha-" + runID, "registration-bravo-" + runID}
	verificationKeys := []string{"verification-alpha-" + runID, "verification-bravo-" + runID}
	expiredVerificationKey := "verification-expired-" + runID
	unknownVerificationKey := "verification-unknown-" + runID
	t.Cleanup(func() {
		task9CleanupIntegrationIdentity(
			t, pool, lookupDigest, registrationKeys, verificationKeys, expiredVerificationKey, unknownVerificationKey,
		)
	})
	type registrationOutcome struct {
		result RegisterAccountResult
		err    error
	}
	registrationResults := make(chan registrationOutcome, 2)
	startRegistration := make(chan struct{})
	for _, key := range registrationKeys {
		key := key
		go func() {
			<-startRegistration
			result, registerErr := registerApplication.RegisterAccount(context.Background(), RegisterAccountCommand{
				Email: email, Password: secret.NewBytes([]byte("correct horse battery staple")), Locale: "en", IdempotencyKey: key,
			})
			registrationResults <- registrationOutcome{result: result, err: registerErr}
		}()
	}
	close(startRegistration)
	task9AwaitBarrier(t, lookupReady, 2)
	close(lookupRelease)
	for range 2 {
		select {
		case outcome := <-registrationResults:
			if outcome.err != nil || outcome.result != (RegisterAccountResult{Accepted: true}) {
				t.Fatal("concurrent registration did not return the generic accepted shape")
			}
		case <-time.After(10 * time.Second):
			t.Fatal("concurrent registration did not complete")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var principalID uuid.UUID
	var deliveryID uuid.UUID
	var deliveryCiphertext []byte
	var deliveryKeyVersion int32
	if scanErr := pool.QueryRow(ctx, `
SELECT principal_id, verification_delivery_id, verification_delivery_ciphertext, verification_delivery_key_version
FROM identity.email_identities WHERE lookup_digest = $1`, lookupDigest[:]).Scan(
		&principalID, &deliveryID, &deliveryCiphertext, &deliveryKeyVersion,
	); scanErr != nil {
		t.Fatal("read concurrent registration result")
	}
	for _, query := range []string{
		`SELECT count(*) FROM identity.accounts WHERE id = $1`,
		`SELECT count(*) FROM identity.email_identities WHERE principal_id = $1`,
		`SELECT count(*) FROM identity.password_credentials WHERE principal_id = $1`,
	} {
		var count int
		if countErr := pool.QueryRow(ctx, query, principalID).Scan(&count); countErr != nil || count != 1 {
			t.Fatal("concurrent registration did not create exactly one identity graph")
		}
	}

	opened, err := protector.Decrypt(emailVerificationDeliveryDomain, sensitive.EncryptedField{
		KeyVersion: uint32(deliveryKeyVersion), Ciphertext: bytes.Clone(deliveryCiphertext),
	})
	clear(deliveryCiphertext)
	if err != nil {
		t.Fatal("open integration verification delivery")
	}
	var delivery struct {
		Token string `json:"token"`
	}
	if jsonErr := json.Unmarshal(opened, &delivery); jsonErr != nil {
		clear(opened)
		t.Fatal("decode integration verification delivery")
	}
	clear(opened)
	verificationToken, err := securitykit.DecodeOpaqueToken(delivery.Token)
	delivery.Token = ""
	if err != nil {
		t.Fatal("decode integration verification token")
	}
	task9AssertLookupAndVerificationShareRowLock(t, pool, lookupDigest, securitykit.DigestToken(securitykit.EmailVerificationToken, verificationToken))

	verificationReady := make(chan struct{}, 2)
	verificationRelease := make(chan struct{})
	verificationRepository := &task9VerificationBarrierRepository{
		inner: postgresRepository, ready: verificationReady, release: verificationRelease,
	}
	verificationRepository.remaining.Store(2)
	verifyApplication := task9IntegrationApplication(t, verificationRepository, protector)
	verificationResults := make(chan struct {
		key string
		err error
	}, 2)
	startVerification := make(chan struct{})
	for _, key := range verificationKeys {
		key := key
		go func() {
			<-startVerification
			verifyErr := verifyApplication.VerifyEmail(context.Background(), VerifyEmailCommand{Token: verificationToken, IdempotencyKey: key})
			verificationResults <- struct {
				key string
				err error
			}{key: key, err: verifyErr}
		}()
	}
	close(startVerification)
	task9AwaitBarrier(t, verificationReady, 2)
	close(verificationRelease)
	successfulKey := ""
	authenticationFailures := 0
	for range 2 {
		select {
		case outcome := <-verificationResults:
			if outcome.err == nil {
				if successfulKey != "" {
					t.Fatal("verification token was consumed more than once")
				}
				successfulKey = outcome.key
			} else if publicTask9Code(outcome.err) == apierrors.AuthenticationFailed {
				authenticationFailures++
			} else {
				t.Fatal("concurrent verification returned an unexpected public error")
			}
		case <-time.After(10 * time.Second):
			t.Fatal("concurrent verification did not complete")
		}
	}
	if successfulKey == "" || authenticationFailures != 1 {
		t.Fatal("concurrent verification did not enforce single use")
	}
	if replayErr := verifyApplication.VerifyEmail(context.Background(), VerifyEmailCommand{Token: verificationToken, IdempotencyKey: successfulKey}); replayErr != nil {
		t.Fatal("exact verification request did not replay 204")
	}
	otherKey := verificationKeys[0]
	if otherKey == successfulKey {
		otherKey = verificationKeys[1]
	}
	if usedErr := verifyApplication.VerifyEmail(context.Background(), VerifyEmailCommand{Token: verificationToken, IdempotencyKey: otherKey}); publicTask9Code(usedErr) != apierrors.AuthenticationFailed {
		t.Fatal("consumed verification token with a different key was not rejected")
	}

	rotatedToken := secret.NewBytes(bytes.Repeat([]byte{0xc1}, 32))
	rotatedDigest := securitykit.DigestToken(securitykit.EmailVerificationToken, rotatedToken)
	if _, updateErr := pool.Exec(ctx, `UPDATE identity.email_identities
SET verification_token_hash = $2, verification_consumed_at = NULL, verification_expires_at = $3, updated_at = $3
WHERE principal_id = $1`, principalID, rotatedDigest[:], time.Now().UTC().Add(time.Hour)); updateErr != nil {
		t.Fatal("rotate integration verification token")
	}
	if conflictErr := verifyApplication.VerifyEmail(context.Background(), VerifyEmailCommand{Token: rotatedToken, IdempotencyKey: successfulKey}); publicTask9Code(conflictErr) != apierrors.IdempotencyConflict {
		t.Fatal("same verification key with a different request did not conflict")
	}

	expiredToken := secret.NewBytes(bytes.Repeat([]byte{0xc2}, 32))
	expiredDigest := securitykit.DigestToken(securitykit.EmailVerificationToken, expiredToken)
	if _, updateErr := pool.Exec(ctx, `UPDATE identity.email_identities
SET verification_token_hash = $2, verification_consumed_at = NULL, verification_expires_at = $3, updated_at = $4
WHERE principal_id = $1`, principalID, expiredDigest[:], time.Now().UTC().Add(-time.Minute), time.Now().UTC()); updateErr != nil {
		t.Fatal("expire integration verification token")
	}
	if expiredErr := verifyApplication.VerifyEmail(context.Background(), VerifyEmailCommand{Token: expiredToken, IdempotencyKey: expiredVerificationKey}); publicTask9Code(expiredErr) != apierrors.AuthenticationFailed {
		t.Fatal("expired verification token was not rejected")
	}
	unknownToken := secret.NewBytes(bytes.Repeat([]byte{0xc3}, 32))
	if unknownErr := verifyApplication.VerifyEmail(context.Background(), VerifyEmailCommand{Token: unknownToken, IdempotencyKey: unknownVerificationKey}); publicTask9Code(unknownErr) != apierrors.AuthenticationFailed {
		t.Fatal("unknown verification token was not rejected")
	}

}

type task9LookupBarrierRepository struct {
	inner   Repository
	ready   chan<- struct{}
	release <-chan struct{}
}

func (repository *task9LookupBarrierRepository) WithinTransaction(ctx context.Context, operation func(context.Context, Transaction) error) error {
	return repository.inner.WithinTransaction(ctx, func(transactionContext context.Context, transaction Transaction) error {
		return operation(transactionContext, &task9LookupBarrierTransaction{
			Transaction: transaction, ready: repository.ready, release: repository.release,
		})
	})
}

type task9LookupBarrierTransaction struct {
	Transaction
	ready   chan<- struct{}
	release <-chan struct{}
}

func (transaction *task9LookupBarrierTransaction) LockEmailLookupDigest(ctx context.Context, digest []byte) error {
	select {
	case transaction.ready <- struct{}{}:
	case <-ctx.Done():
		return ErrRepository
	}
	select {
	case <-transaction.release:
	case <-ctx.Done():
		return ErrRepository
	}
	return transaction.Transaction.LockEmailLookupDigest(ctx, digest)
}

type task9VerificationBarrierRepository struct {
	inner     Repository
	ready     chan<- struct{}
	release   <-chan struct{}
	remaining atomic.Int32
}

func (repository *task9VerificationBarrierRepository) WithinTransaction(ctx context.Context, operation func(context.Context, Transaction) error) error {
	return repository.inner.WithinTransaction(ctx, func(transactionContext context.Context, transaction Transaction) error {
		return operation(transactionContext, &task9VerificationBarrierTransaction{
			Transaction: transaction, ready: repository.ready, release: repository.release, remaining: &repository.remaining,
		})
	})
}

type task9VerificationBarrierTransaction struct {
	Transaction
	ready     chan<- struct{}
	release   <-chan struct{}
	remaining *atomic.Int32
}

func (transaction *task9VerificationBarrierTransaction) GetEmailVerificationForUpdate(ctx context.Context, digest []byte) (store.IdentityEmailIdentity, bool, error) {
	if transaction.remaining.Add(-1) < 0 {
		return transaction.Transaction.GetEmailVerificationForUpdate(ctx, digest)
	}
	select {
	case transaction.ready <- struct{}{}:
	case <-ctx.Done():
		return store.IdentityEmailIdentity{}, false, ErrRepository
	}
	select {
	case <-transaction.release:
	case <-ctx.Done():
		return store.IdentityEmailIdentity{}, false, ErrRepository
	}
	return transaction.Transaction.GetEmailVerificationForUpdate(ctx, digest)
}

func task9AwaitBarrier(t *testing.T, ready <-chan struct{}, count int) {
	t.Helper()
	deadline := time.After(10 * time.Second)
	for range count {
		select {
		case <-ready:
		case <-deadline:
			t.Fatal("integration transaction did not reach the database barrier")
		}
	}
}

type task9IntegrationClock struct{}

func (task9IntegrationClock) Now() time.Time { return time.Now().UTC() }

type task9IntegrationLimiter struct{}

func (task9IntegrationLimiter) Allow(context.Context, ratelimit.Operation, [32]byte, config.RateLimitPolicy) (bool, error) {
	return true, nil
}

type task9IntegrationParticipant struct{}

func (task9IntegrationParticipant) ActivateVerifiedPrincipal(context.Context, store.DBTX, PrincipalID, time.Time) error {
	return nil
}

func task9IntegrationApplication(t *testing.T, repository Repository, protector sensitive.Protector) *Service {
	t.Helper()
	application, err := newApplicationForTest(ApplicationDependencies{
		Repository: repository, Protector: protector, Random: rand.Reader, Clock: task9IntegrationClock{},
		Limiter: task9IntegrationLimiter{}, RateLimitKey: secret.NewBytes(bytes.Repeat([]byte{0xb3}, 32)),
		Security: config.SecurityConfig{
			Profile: config.ProfileTest, EmailVerification: config.EmailRequired, RequestDeadline: 8 * time.Second,
			RedisTimeout: 250 * time.Millisecond, DeliveryRateLimit: config.RateLimitPolicy{Limit: 5, Window: time.Hour},
		},
		DeviceAuthorizationParticipant: task9IntegrationParticipant{},
	}, func(_ []byte, _ []byte, policy PasswordPolicy) []byte {
		return bytes.Repeat([]byte{0xd1}, int(policy.TagBytes))
	}, uuid.New)
	if err != nil {
		t.Fatal("create integration application")
	}
	return application
}

func task9AssertLookupAndVerificationShareRowLock(t *testing.T, pool *pgxpool.Pool, lookupDigest, verificationDigest [32]byte) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	lookupTransaction, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal("begin lookup lock transaction")
	}
	lookupClosed := false
	defer func() {
		if !lookupClosed {
			_ = lookupTransaction.Rollback(context.Background())
		}
	}()
	locked, err := store.New(lookupTransaction).FindIdentityByLookupDigest(ctx, lookupDigest[:])
	if err != nil || locked.PrincipalID == uuid.Nil {
		t.Fatal("lookup path did not lock the email row")
	}

	verificationTransaction, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal("begin verification lock transaction")
	}
	verificationClosed := false
	defer func() {
		if !verificationClosed {
			_ = verificationTransaction.Rollback(context.Background())
		}
	}()
	type lockResult struct {
		identity store.IdentityEmailIdentity
		err      error
	}
	result := make(chan lockResult, 1)
	go func() {
		identity, queryErr := store.New(verificationTransaction).GetEmailVerificationForUpdate(ctx, verificationDigest[:])
		result <- lockResult{identity: identity, err: queryErr}
	}()
	select {
	case <-result:
		t.Fatal("verification path did not contend on the lookup-held email row")
	case <-time.After(150 * time.Millisecond):
	}
	if err = lookupTransaction.Rollback(ctx); err != nil {
		t.Fatal("release lookup email-row lock")
	}
	lookupClosed = true
	select {
	case outcome := <-result:
		if outcome.err != nil || outcome.identity.PrincipalID != locked.PrincipalID {
			t.Fatal("verification path did not acquire the released email-row lock")
		}
	case <-ctx.Done():
		t.Fatal("verification path remained blocked after lookup lock release")
	}
	if err = verificationTransaction.Rollback(ctx); err != nil {
		t.Fatal("release verification email-row lock")
	}
	verificationClosed = true
}

func task9CleanupIntegrationIdentity(
	t *testing.T,
	pool *pgxpool.Pool,
	lookupDigest [32]byte,
	registrationKeys, verificationKeys []string,
	expiredVerificationKey, unknownVerificationKey string,
) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cleanup := func(query string, arguments ...any) {
		if _, err := pool.Exec(ctx, query, arguments...); err != nil {
			t.Error("cleanup integration identity")
		}
	}
	allKeys := append([]string(nil), registrationKeys...)
	allKeys = append(allKeys, verificationKeys...)
	allKeys = append(allKeys, expiredVerificationKey, unknownVerificationKey)
	for _, key := range allKeys {
		cleanup(`DELETE FROM transactional_outbox WHERE idempotency_key = $1`, key)
	}
	for _, key := range registrationKeys {
		digest, err := idempotency.KeyDigest(key)
		if err != nil {
			t.Error("derive integration cleanup key")
			continue
		}
		cleanup(
			`DELETE FROM idempotency_records WHERE principal_scope = $1 AND operation = $2 AND idempotency_key_hash = $3`,
			idempotency.AnonymousRegistrationPrincipal, idempotency.AnonymousRegistrationOperation, digest[:],
		)
		clear(digest[:])
	}
	var principalID uuid.UUID
	queryErr := pool.QueryRow(ctx, `SELECT principal_id FROM identity.email_identities WHERE lookup_digest = $1`, lookupDigest[:]).Scan(&principalID)
	if errors.Is(queryErr, pgx.ErrNoRows) {
		return
	}
	if queryErr != nil {
		t.Error("locate integration identity for cleanup")
		return
	}
	authenticatedScope := "principal:" + principalID.String() + ":email"
	verificationRecords := append([]string(nil), verificationKeys...)
	verificationRecords = append(verificationRecords, expiredVerificationKey, unknownVerificationKey)
	for _, key := range verificationRecords {
		digest, err := idempotency.KeyDigest(key)
		if err != nil {
			t.Error("derive integration cleanup key")
			continue
		}
		cleanup(
			`DELETE FROM idempotency_records WHERE principal_scope = $1 AND operation = $2 AND idempotency_key_hash = $3`,
			authenticatedScope, "verify_email", digest[:],
		)
		clear(digest[:])
	}
	cleanup(`DELETE FROM identity.security_events WHERE principal_id = $1`, principalID)
	cleanup(`DELETE FROM identity.password_credentials WHERE principal_id = $1`, principalID)
	cleanup(`DELETE FROM identity.email_identities WHERE principal_id = $1`, principalID)
	cleanup(`DELETE FROM identity.accounts WHERE id = $1`, principalID)
}
