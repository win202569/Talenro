package identity

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base32"
	"errors"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/ratelimit"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

const task11TOTPProtectionDomain = "identity/totp/v1"

func TestReauthRejectsExpiredOrUnsupportedProofBeforeFactorMutation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*task11TOTPTransaction, *BeginTOTPEnrollmentCommand)
		want      apierrors.Code
	}{
		{
			name: "expired session", want: apierrors.AuthenticationFailed,
			configure: func(transaction *task11TOTPTransaction, _ *BeginTOTPEnrollmentCommand) {
				transaction.session.AccessExpiresAt = fixedTask10Time
			},
		},
		{
			name: "unsupported TOTP reauth", want: apierrors.ActionNotAllowed,
			configure: func(_ *task11TOTPTransaction, command *BeginTOTPEnrollmentCommand) {
				command.Reauthentication.Method = ReauthTOTP
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			transaction := activeTask11TOTPTransaction()
			application, _ := newTask11TOTPApplication(t, transaction)
			command := BeginTOTPEnrollmentCommand{
				PrincipalID:      PrincipalID(transaction.account.ID.String()),
				Reauthentication: task11PasswordReauthentication(transaction.session.ID),
				IdempotencyKey:   "task11-reauth-negative-01",
			}
			test.configure(transaction, &command)
			_, err := application.BeginTOTPEnrollment(context.Background(), command)
			if publicTask9Code(err) != test.want {
				t.Fatalf("reauth error = %v, want %s", err, test.want)
			}
			if transaction.createdTOTP.PrincipalID != uuid.Nil {
				t.Fatal("failed reauthentication created a factor")
			}
		})
	}
}

func TestConcurrentTOTPVerificationAcceptsOneStep(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	application, protector := newTask11TOTPApplication(t, transaction)
	secretBytes := bytes.Repeat([]byte{0x5a}, totpSecretBytes)
	protected, err := protector.Encrypt(task11TOTPProtectionDomain, secretBytes)
	if err != nil {
		t.Fatal(err)
	}
	transaction.totpFound = true
	transaction.totp = store.IdentityTotpCredential{
		PrincipalID: transaction.account.ID, Ciphertext: protected.Ciphertext,
		// #nosec G115 -- the test protector uses the fixed key version 7.
		EncryptionKeyVersion: int32(protected.KeyVersion), State: "pending",
	}
	transaction.statefulTOTP = true
	code, err := totp.GenerateCodeCustom(task11Base32.EncodeToString(secretBytes), fixedTask10Time, totp.ValidateOpts{
		Period: totpPeriod, Skew: 0, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	idempotencyKeys := [...]string{"task11-totp-concurrent-01", "task11-totp-concurrent-02"}
	var workers sync.WaitGroup
	for index := range 2 {
		workers.Add(1)
		go func(worker int) {
			defer workers.Done()
			<-start
			results <- application.VerifyTOTPEnrollment(context.Background(), VerifyTOTPEnrollmentCommand{
				PrincipalID: PrincipalID(transaction.account.ID.String()), Code: secret.NewBytes([]byte(code)),
				Reauthentication: task11PasswordReauthentication(transaction.session.ID),
				IdempotencyKey:   idempotencyKeys[worker],
			})
		}(index)
	}
	close(start)
	workers.Wait()
	close(results)
	succeeded := 0
	failed := 0
	for result := range results {
		if result == nil {
			succeeded++
		} else {
			failed++
		}
	}
	if succeeded != 1 || failed != 1 {
		t.Fatalf("concurrent TOTP results = %d success/%d failure, want 1/1", succeeded, failed)
	}
	operations := strings.Join(transaction.operations, ",")
	if strings.Count(operations, "activate_totp") != 1 || strings.Count(operations, "accept_totp_step") != 1 {
		t.Fatalf("concurrent TOTP mutations = %s", operations)
	}
}

func TestBeginTOTPEnrollmentReauthenticatesAndPersistsOnlyEncryptedSecret(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	application, protector := newTask11TOTPApplication(t, transaction)
	principalID := transaction.account.ID
	command := BeginTOTPEnrollmentCommand{
		PrincipalID:      PrincipalID(principalID.String()),
		Reauthentication: task11PasswordReauthentication(transaction.session.ID),
		IdempotencyKey:   "task11-totp-enroll-0001",
	}
	result, err := application.BeginTOTPEnrollment(context.Background(), command)
	if err != nil {
		t.Fatalf("begin TOTP enrollment: %v", err)
	}
	if result.Secret == "" || result.URI == "" {
		t.Fatal("enrollment secret or URI is empty")
	}
	if transaction.createdTOTP.PrincipalID != principalID || transaction.createdTOTP.KeyVersion != 7 {
		t.Fatalf("unexpected persisted TOTP metadata: %+v", transaction.createdTOTP)
	}
	// #nosec G115 -- the assertion above proves the persisted test key version is exactly 7.
	persistedKeyVersion := uint32(transaction.createdTOTP.KeyVersion)
	plaintext, err := protector.Decrypt(task11TOTPProtectionDomain, sensitive.EncryptedField{
		KeyVersion: persistedKeyVersion, Ciphertext: transaction.createdTOTP.EncryptedSecret,
	})
	if err != nil {
		t.Fatalf("decrypt persisted TOTP secret: %v", err)
	}
	defer clear(plaintext)
	if len(plaintext) != totpSecretBytes {
		t.Fatalf("persisted secret bytes = %d, want %d", len(plaintext), totpSecretBytes)
	}
	decodedSecret, err := task11Base32.DecodeString(result.Secret)
	if err != nil || len(result.Secret) != 32 || !bytes.Equal(decodedSecret, plaintext) {
		clear(decodedSecret)
		t.Fatalf("returned base32 secret does not encode the exact generated %d bytes", totpSecretBytes)
	}
	defer clear(decodedSecret)
	if _, wrongDomainErr := protector.Decrypt("identity/totp/v0", sensitive.EncryptedField{
		KeyVersion: persistedKeyVersion, Ciphertext: transaction.createdTOTP.EncryptedSecret,
	}); wrongDomainErr == nil {
		t.Fatal("TOTP ciphertext decrypted under a different domain")
	}
	parsed, err := url.Parse(result.URI)
	if err != nil {
		t.Fatalf("parse enrollment URI: %v", err)
	}
	if got := parsed.Query().Get("secret"); got != result.Secret {
		t.Fatal("enrollment URI and explicit secret differ")
	}
	wantOperations := []string{
		"get_credential", "get_totp", "get_session", "get_account", "begin_idempotency",
		"create_totp", "complete_idempotency", "commit",
	}
	if got := strings.Join(transaction.operations, ","); got != strings.Join(wantOperations, ",") {
		t.Fatalf("operation order = %s, want %s", got, strings.Join(wantOperations, ","))
	}
	replayed, err := application.BeginTOTPEnrollment(context.Background(), command)
	if err != nil || replayed.Secret != result.Secret || replayed.URI != result.URI || strings.Count(strings.Join(transaction.operations, ","), "create_totp") != 1 {
		t.Fatalf("TOTP enrollment replay = %#v/%v, operations=%v", replayed, err, transaction.operations)
	}
}

func TestBeginTOTPEnrollmentRejectsPendingOrActiveFactor(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	transaction.createTOTPRows = 0
	application, _ := newTask11TOTPApplication(t, transaction)
	result, err := application.BeginTOTPEnrollment(context.Background(), BeginTOTPEnrollmentCommand{
		PrincipalID:      PrincipalID(transaction.account.ID.String()),
		Reauthentication: task11PasswordReauthentication(transaction.session.ID),
		IdempotencyKey:   "task11-totp-enroll-0002",
	})
	if publicTask9Code(err) != apierrors.StateConflict {
		t.Fatalf("pending-factor error = %v", err)
	}
	if result.URI != "" {
		t.Fatal("failed enrollment returned a provisioning URI")
	}
}

func TestTOTPRevocationDoesNotMutateOtherFactors(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	transaction.totpFound = true
	transaction.totp = store.IdentityTotpCredential{PrincipalID: transaction.account.ID, State: "active"}
	transaction.passkeys = []store.IdentityPasskeyCredential{{
		CredentialID: bytes.Repeat([]byte{0x42}, 32), PrincipalID: transaction.account.ID, State: "active",
	}}
	transaction.recoveryFound = true
	transaction.recovery = store.IdentityRecoveryCodeSet{
		ID: uuid.MustParse("ec27e38e-a794-42c1-bce8-1baa6a7caa39"), PrincipalID: transaction.account.ID,
		Generation: 1, State: "active", CodeHashes: task11RecoveryHashes(),
	}
	application, _ := newTask11TOTPApplication(t, transaction)
	if err := application.RevokeTOTP(context.Background(), RevokeTOTPCommand{
		PrincipalID:      PrincipalID(transaction.account.ID.String()),
		Reauthentication: task11PasswordReauthentication(transaction.session.ID),
		IdempotencyKey:   "task11-totp-revoke-0001",
	}); err != nil {
		t.Fatalf("revoke TOTP: %v", err)
	}
	if transaction.revokedTOTP.PrincipalID != transaction.account.ID || !transaction.revokedTOTP.RevokedAt.Valid {
		t.Fatalf("revoked TOTP params = %+v", transaction.revokedTOTP)
	}
	if transaction.revokedPasskey.PrincipalID != uuid.Nil || transaction.revokedRecovery.PrincipalID != uuid.Nil {
		t.Fatal("TOTP revocation mutated another factor")
	}
	wantOperations := "get_credential,get_totp,get_session,get_account,begin_idempotency,revoke_totp,complete_idempotency,commit"
	if got := strings.Join(transaction.operations, ","); got != wantOperations {
		t.Fatalf("operation order = %s, want %s", got, wantOperations)
	}
}

func TestVerifyTOTPEnrollmentActivatesThenAtomicallyAcceptsFirstStep(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	application, protector := newTask11TOTPApplication(t, transaction)
	secretBytes := make([]byte, totpSecretBytes)
	for index := range secretBytes {
		secretBytes[index] = byte(index + 1)
	}
	protected, err := protector.Encrypt(task11TOTPProtectionDomain, secretBytes)
	if err != nil {
		t.Fatal(err)
	}
	transaction.totpFound = true
	transaction.totp = store.IdentityTotpCredential{
		PrincipalID: transaction.account.ID, Ciphertext: protected.Ciphertext,
		// #nosec G115 -- the test protector uses the fixed key version 7.
		EncryptionKeyVersion: int32(protected.KeyVersion), State: "pending",
	}
	transaction.statefulTOTP = true
	code, err := totp.GenerateCodeCustom(task11Base32.EncodeToString(secretBytes), fixedTask10Time, totp.ValidateOpts{
		Period: totpPeriod, Skew: 0, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		t.Fatal(err)
	}
	transaction.operations = nil
	command := VerifyTOTPEnrollmentCommand{
		PrincipalID:      PrincipalID(transaction.account.ID.String()),
		Code:             secret.NewBytes([]byte(code)),
		Reauthentication: task11PasswordReauthentication(transaction.session.ID),
		IdempotencyKey:   "task11-totp-verify-0001",
	}
	err = application.VerifyTOTPEnrollment(context.Background(), command)
	if err != nil {
		t.Fatalf("verify TOTP enrollment: %v", err)
	}
	wantStep := fixedTask10Time.Unix() / totpPeriod
	if !transaction.acceptedStep.Valid || transaction.acceptedStep.Int64 != wantStep {
		t.Fatalf("accepted step = %+v, want %d", transaction.acceptedStep, wantStep)
	}
	wantOperations := []string{
		"get_credential", "get_totp", "get_session", "get_account", "begin_idempotency",
		"activate_totp", "accept_totp_step", "complete_idempotency", "commit",
	}
	if got := strings.Join(transaction.operations, ","); got != strings.Join(wantOperations, ",") {
		t.Fatalf("operation order = %s, want %s", got, strings.Join(wantOperations, ","))
	}
	if err = application.VerifyTOTPEnrollment(context.Background(), command); err != nil {
		t.Fatalf("replay successful TOTP verification: %v", err)
	}
	if strings.Count(strings.Join(transaction.operations, ","), "accept_totp_step") != 1 {
		t.Fatalf("successful TOTP replay accepted the step twice: %v", transaction.operations)
	}

	transaction.operations = nil
	transaction.acceptTOTPRows = 0
	err = application.VerifyTOTPEnrollment(context.Background(), VerifyTOTPEnrollmentCommand{
		PrincipalID:      PrincipalID(transaction.account.ID.String()),
		Code:             secret.NewBytes([]byte(code)),
		Reauthentication: task11PasswordReauthentication(transaction.session.ID),
		IdempotencyKey:   "task11-totp-verify-0002",
	})
	if publicTask9Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("replayed-step error = %v", err)
	}
	if got := transaction.operations[len(transaction.operations)-1]; got != "rollback" {
		t.Fatalf("replayed-step terminal operation = %s, want rollback", got)
	}
}

func TestTOTPUsesExactSecretAndRFC6238CompatibilityWindow(t *testing.T) {
	t.Parallel()

	secretBytes := make([]byte, 20)
	for index := range secretBytes {
		secretBytes[index] = byte(index + 1)
	}
	uri, err := buildTOTPEnrollmentURI(secretBytes)
	if err != nil {
		t.Fatalf("build enrollment URI: %v", err)
	}
	parsed, err := url.Parse(uri)
	if err != nil {
		t.Fatalf("parse enrollment URI: %v", err)
	}
	wantSecret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secretBytes)
	if parsed.Scheme != "otpauth" || parsed.Host != "totp" ||
		parsed.Query().Get("secret") != wantSecret || parsed.Query().Get("digits") != "6" ||
		parsed.Query().Get("algorithm") != "SHA1" || parsed.Query().Get("period") != "30" {
		t.Fatalf("unexpected enrollment URI contract: %s", uri)
	}

	now := time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)
	previousCode, err := totp.GenerateCodeCustom(wantSecret, now.Add(-30*time.Second), totp.ValidateOpts{
		Period: 30, Skew: 0, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		t.Fatalf("generate previous code: %v", err)
	}
	step, ok := verifyTOTPStep(secretBytes, previousCode, now)
	if !ok || step != now.Unix()/30-1 {
		t.Fatalf("previous-window result = %d/%v", step, ok)
	}
	if _, accepted := verifyTOTPStep(secretBytes, previousCode, now.Add(60*time.Second)); accepted {
		t.Fatal("code outside +/-1 step was accepted")
	}
}

func TestTOTPGeneratedQueriesEnforceRevokedOnlyEnrollmentAndSeparateFirstStep(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	principalID := uuid.New()
	now := time.Now().UTC()
	db := &task11RecordingDBTX{rowsAffected: 1}
	queries := store.New(db)

	rows, err := queries.CreateTOTPEnrollment(ctx, store.CreateTOTPEnrollmentParams{
		PrincipalID:     principalID,
		EncryptedSecret: []byte("ciphertext"),
		KeyVersion:      7,
		EnrolledAt:      now,
	})
	if err != nil {
		t.Fatalf("create TOTP enrollment: %v", err)
	}
	if rows != 1 {
		t.Fatalf("affected rows = %d, want 1", rows)
	}
	if !strings.Contains(db.lastSQL, "ON CONFLICT (principal_id)") ||
		!strings.Contains(db.lastSQL, "WHERE totp_factors.state = 'revoked'") {
		t.Fatalf("create TOTP SQL must permit only revoked-factor replacement: %s", db.lastSQL)
	}
	if got := len(db.lastArgs); got != 4 {
		t.Fatalf("create TOTP args = %d, want 4", got)
	}

	activated, err := queries.ActivateTOTP(ctx, store.ActivateTOTPParams{
		PrincipalID: principalID,
		VerifiedAt:  sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		t.Fatalf("activate TOTP: %v", err)
	}
	if activated != 1 {
		t.Fatalf("activated rows = %d, want 1", activated)
	}
	if strings.Contains(db.lastSQL, "last_accepted_step") {
		t.Fatalf("activation must not accept a step: %s", db.lastSQL)
	}
	if got := len(db.lastArgs); got != 2 {
		t.Fatalf("activate TOTP args = %d, want 2", got)
	}

	accepted, err := queries.AcceptTOTPStep(ctx, store.AcceptTOTPStepParams{
		PrincipalID:      principalID,
		LastAcceptedStep: pgtype.Int8{Int64: 42, Valid: true},
	})
	if err != nil {
		t.Fatalf("accept first TOTP step: %v", err)
	}
	if accepted != 1 {
		t.Fatalf("accepted rows = %d, want 1", accepted)
	}
	if !strings.Contains(db.lastSQL, "last_accepted_step") {
		t.Fatalf("step acceptance SQL must advance the replay barrier: %s", db.lastSQL)
	}
}

type task11RecordingDBTX struct {
	lastSQL      string
	lastArgs     []any
	rowsAffected int64
}

func (db *task11RecordingDBTX) Exec(
	_ context.Context,
	query string,
	args ...any,
) (pgconn.CommandTag, error) {
	db.lastSQL = query
	db.lastArgs = append([]any(nil), args...)
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func (db *task11RecordingDBTX) Query(
	_ context.Context,
	query string,
	_ ...any,
) (pgx.Rows, error) {
	db.lastSQL = query
	return nil, errors.New("unexpected Query call")
}

func (db *task11RecordingDBTX) QueryRow(
	context.Context,
	string,
	...any,
) pgx.Row {
	return task11ErrorRow{err: errors.New("unexpected QueryRow call")}
}

type task11ErrorRow struct {
	err error
}

func (row task11ErrorRow) Scan(...any) error {
	return row.err
}

var _ store.DBTX = (*task11RecordingDBTX)(nil)

type task11TOTPRepository struct {
	mu sync.Mutex
	tx *task11TOTPTransaction
}

type task11Limiter struct{}

func (task11Limiter) Allow(
	context.Context,
	ratelimit.Operation,
	[32]byte,
	config.RateLimitPolicy,
) (bool, error) {
	return true, nil
}

func (repository *task11TOTPRepository) WithinTransaction(
	ctx context.Context,
	operation func(context.Context, Transaction) error,
) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	recoveryBefore := repository.tx.recovery
	recoveryBefore.CodeHashes = cloneTask11Hashes(repository.tx.recovery.CodeHashes)
	totpBefore := repository.tx.totp
	totpBefore.Ciphertext = append([]byte(nil), repository.tx.totp.Ciphertext...)
	lastAcceptedStepBefore := repository.tx.lastAcceptedStep
	createdSessionBefore := repository.tx.createdSession
	createdRefreshBefore := repository.tx.createdRefresh
	var idempotencyBefore map[string]store.IdempotencyRecord
	if repository.tx.idempotencyDB != nil {
		idempotencyBefore = repository.tx.idempotencyDB.snapshot()
	}
	err := operation(ctx, repository.tx)
	if err != nil {
		repository.tx.recovery = recoveryBefore
		repository.tx.totp = totpBefore
		repository.tx.lastAcceptedStep = lastAcceptedStepBefore
		repository.tx.createdSession = createdSessionBefore
		repository.tx.createdRefresh = createdRefreshBefore
		if repository.tx.idempotencyDB != nil {
			repository.tx.idempotencyDB.restore(idempotencyBefore)
		}
		repository.tx.operations = append(repository.tx.operations, "rollback")
		return err
	}
	repository.tx.operations = append(repository.tx.operations, "commit")
	return nil
}

type task11TOTPTransaction struct {
	*fakeIdentityTransaction
	totp                   store.IdentityTotpCredential
	totpFound              bool
	createTOTPRows         int64
	activateTOTPRows       int64
	acceptTOTPRows         int64
	createdTOTP            store.CreateTOTPEnrollmentParams
	revokedTOTP            store.RevokeTOTPParams
	acceptedStep           pgtype.Int8
	recovery               store.IdentityRecoveryCodeSet
	recoveryFound          bool
	nextRecoveryGeneration int32
	createdRecovery        store.CreateRecoveryCodeSetParams
	revokedRecovery        store.RevokeRecoveryCodeSetsParams
	consumeRecoveryFound   bool
	consumedRecovery       store.ConsumeRecoveryCodeParams
	passkeys               []store.IdentityPasskeyCredential
	createdPasskey         store.CreatePasskeyCredentialParams
	updatedPasskey         store.UpdatePasskeyCounterParams
	revokedPasskey         store.RevokePasskeyParams
	updatePasskeyRows      int64
	revokePasskeyRows      int64
	statefulTOTP           bool
	lastAcceptedStep       int64
	statefulRecovery       bool
	idempotency            idempotency.Repository
	idempotencyDB          *task10IdempotencyDB
}

func (transaction *task11TOTPTransaction) BeginIdempotency(
	ctx context.Context,
	scope idempotency.Scope,
	key string,
	canonical []byte,
	createdAt time.Time,
	expiresAt time.Time,
) (idempotency.Record, idempotency.Outcome, error) {
	transaction.idempotencyCanonical = append(transaction.idempotencyCanonical, bytes.Clone(canonical))
	if err := transaction.record("begin_idempotency"); err != nil {
		return idempotency.Record{}, "", err
	}
	return transaction.idempotency.Begin(ctx, scope, key, canonical, createdAt, expiresAt)
}

func (transaction *task11TOTPTransaction) CompleteIdempotency(
	ctx context.Context,
	record idempotency.Record,
	status int,
	body []byte,
) error {
	if err := transaction.record("complete_idempotency"); err != nil {
		return err
	}
	completed, err := transaction.idempotency.Complete(ctx, record, status, body)
	if err != nil {
		return err
	}
	owned, available := completed.TakeResponseBody()
	defer clear(owned)
	if !available {
		return errors.New("idempotency completion did not transfer response ownership")
	}
	return nil
}

func (transaction *task11TOTPTransaction) GetTOTPForUpdate(
	context.Context,
	uuid.UUID,
) (store.IdentityTotpCredential, bool, error) {
	return transaction.totp, transaction.totpFound, transaction.record("get_totp")
}

func (transaction *task11TOTPTransaction) CreateTOTPEnrollment(
	_ context.Context,
	params store.CreateTOTPEnrollmentParams,
) (int64, error) {
	transaction.createdTOTP = params
	transaction.createdTOTP.EncryptedSecret = append([]byte(nil), params.EncryptedSecret...)
	return transaction.createTOTPRows, transaction.record("create_totp")
}

func (transaction *task11TOTPTransaction) ActivateTOTP(
	context.Context,
	store.ActivateTOTPParams,
) (int64, error) {
	if transaction.statefulTOTP {
		if transaction.totp.State != "pending" {
			return 0, transaction.record("activate_totp")
		}
		transaction.totp.State = "active"
	}
	return transaction.activateTOTPRows, transaction.record("activate_totp")
}

func (transaction *task11TOTPTransaction) AcceptTOTPStep(
	_ context.Context,
	params store.AcceptTOTPStepParams,
) (int64, error) {
	transaction.acceptedStep = params.LastAcceptedStep
	if transaction.statefulTOTP {
		if transaction.totp.State != "active" || !params.LastAcceptedStep.Valid ||
			(transaction.lastAcceptedStep != 0 && params.LastAcceptedStep.Int64 <= transaction.lastAcceptedStep) {
			return 0, transaction.record("accept_totp_step")
		}
		transaction.lastAcceptedStep = params.LastAcceptedStep.Int64
	}
	return transaction.acceptTOTPRows, transaction.record("accept_totp_step")
}

func (transaction *task11TOTPTransaction) RevokeTOTP(
	_ context.Context,
	params store.RevokeTOTPParams,
) (int64, error) {
	transaction.revokedTOTP = params
	return 1, transaction.record("revoke_totp")
}

func activeTask11TOTPTransaction() *task11TOTPTransaction {
	principalID := uuid.MustParse("28ceee8a-5f4f-4d3a-9e9f-d3ec23b815ac")
	sessionID := uuid.MustParse("ac9150c8-aaaf-468b-a9fe-605d8c831d1a")
	base := &fakeIdentityTransaction{
		accountFound: true,
		account:      store.IdentityAccount{ID: principalID, State: "active", StateVersion: 4},
		sessionFound: true,
		session: store.IdentityAccountSession{
			ID: sessionID, PrincipalID: principalID, State: "active",
			AccessExpiresAt: fixedTask10Time.Add(time.Hour), AbsoluteExpiresAt: fixedTask10Time.Add(24 * time.Hour),
		},
		credentialFound: true,
		credential: store.IdentityPasswordCredential{
			PrincipalID: principalID, PolicyVersion: 1, MemoryKib: 65536, TimeCost: 3, Parallelism: 4,
			Salt: make([]byte, 16), PasswordHash: make([]byte, 32), UpdatedAt: fixedTask10Time,
		},
	}
	for index := range base.credential.PasswordHash {
		base.credential.PasswordHash[index] = 0xa5
	}
	return &task11TOTPTransaction{
		fakeIdentityTransaction: base,
		createTOTPRows:          1, activateTOTPRows: 1, acceptTOTPRows: 1,
		updatePasskeyRows: 1, revokePasskeyRows: 1,
	}
}

func task11PasswordReauthentication(sessionID uuid.UUID) Reauthentication {
	return Reauthentication{
		SessionID: SessionID(sessionID.String()), Method: ReauthPassword,
		Proof: secret.NewBytes([]byte("correct horse battery staple")),
	}
}

func newTask11TOTPApplication(
	t *testing.T,
	transaction *task11TOTPTransaction,
) (*Service, *sensitive.Local) {
	t.Helper()
	protector, err := sensitive.NewLocal(
		secret.NewBytes(make([]byte, 32)), secret.NewBytes(make([]byte, 32)), 7,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = protector.Close() })
	bindTask11Idempotency(t, transaction, protector)
	application, err := newApplicationForTest(ApplicationDependencies{
		Repository: &task11TOTPRepository{tx: transaction}, Protector: protector,
		Random: &task9Random{}, Clock: task10Clock{}, Limiter: task11Limiter{},
		RateLimitKey: secret.NewBytes(make([]byte, 32)),
		Security: config.SecurityConfig{
			Profile: config.ProfileTest, EmailVerification: config.EmailDisabled,
			PublicBaseURL: "https://api.example.test", RequestDeadline: 5 * time.Second,
			RedisTimeout:       250 * time.Millisecond,
			LoginRateLimit:     config.RateLimitPolicy{Limit: 10, Window: 15 * time.Minute},
			DeliveryRateLimit:  config.RateLimitPolicy{Limit: 5, Window: time.Hour},
			ChallengeRateLimit: config.RateLimitPolicy{Limit: 20, Window: 5 * time.Minute},
		},
		DeviceAuthorizationParticipant: &task9Participant{owner: transaction.fakeIdentityTransaction},
	}, (&task10Deriver{}).Derive, uuid.New)
	if err != nil {
		t.Fatal(err)
	}
	return application, protector
}

func bindTask11Idempotency(t *testing.T, transaction *task11TOTPTransaction, protector sensitive.Protector) {
	t.Helper()
	transaction.idempotencyDB = newTask10IdempotencyDB()
	bound, err := idempotency.New(transaction.idempotencyDB, protector)
	if err != nil {
		t.Fatal(err)
	}
	transaction.idempotency = bound
}
