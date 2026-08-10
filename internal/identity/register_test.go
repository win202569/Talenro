package identity

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/proto"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	identityv1 "talenro.local/platform/gen/go/talenro/identity/v1"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/config"
	contractevents "talenro.local/platform/internal/contracts/events"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/ratelimit"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

func TestRegisterNewAndDuplicateAreIndistinguishableAndOrdered(t *testing.T) {
	t.Parallel()

	command := RegisterAccountCommand{
		Email: "Member@Example.test", Password: secret.NewBytes([]byte("correct horse battery staple")), Locale: "en",
		IdempotencyKey: "abcdefghijklmnopqrstuv",
	}
	newTx := &fakeIdentityTransaction{}
	newApplication, newDeriver, _, _ := newTask9Application(t, config.EmailRequired, newTx)
	newResult, err := newApplication.RegisterAccount(context.Background(), command)
	if err != nil || newResult != (RegisterAccountResult{Accepted: true}) {
		t.Fatalf("new registration = %#v, %v", newResult, err)
	}
	if newDeriver.calls != 1 {
		t.Fatalf("new registration derivations = %d, want 1", newDeriver.calls)
	}
	wantNewOrder := []string{
		"begin_idempotency", "find_identity", "create_account", "create_email_identity", "create_password_credential",
		"insert_security_event", "append_event", "complete_idempotency", "commit",
	}
	assertTask9Order(t, newTx.operations, wantNewOrder)
	if newTx.createdAccount.State != "pending_email" || newTx.createdEmail.PrincipalID != newTx.createdAccount.ID {
		t.Fatal("new registration did not preserve account/email ownership")
	}
	if newTx.createdEmail.VerificationExpiresAt.Time.Sub(fixedTask9Time) != 24*time.Hour {
		t.Fatalf("verification TTL = %s", newTx.createdEmail.VerificationExpiresAt.Time.Sub(fixedTask9Time))
	}
	assertProtectedDelivery(t, newApplication.protector, emailVerificationDeliveryDomain, newTx.createdEmail.VerificationDeliveryCiphertext, newTx.createdEmail.VerificationDeliveryKeyVersion.Int32, "Member@example.test", string(VerifyEmailTemplate), "en")
	assertPrivateEmailDeliveryEvent(t, newTx.events[0], newTx.createdEmail.VerificationDeliveryID.UUID, newTx.createdAccount.ID, VerifyEmailTemplate, "en", []byte("Member@example.test"), command.Password.Copy())

	existingTx := &fakeIdentityTransaction{identityFound: true, identity: store.IdentityEmailIdentity{PrincipalID: uuid.New()}}
	existingApplication, existingDeriver, _, _ := newTask9Application(t, config.EmailRequired, existingTx)
	existingResult, err := existingApplication.RegisterAccount(context.Background(), command)
	if err != nil || existingResult != newResult {
		t.Fatalf("duplicate registration = %#v, %v; new = %#v", existingResult, err, newResult)
	}
	if existingDeriver.calls != 1 {
		t.Fatalf("duplicate registration derivations = %d, want 1", existingDeriver.calls)
	}
	assertTask9Order(t, existingTx.operations, []string{"begin_idempotency", "find_identity", "complete_idempotency", "commit"})
	if existingTx.mutationCount() != 0 {
		t.Fatalf("duplicate registration mutations = %d", existingTx.mutationCount())
	}
}

func TestRegisterPoliciesAndFiftyRunEnumerationBudget(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name           string
		mode           config.EmailVerificationMode
		wantState      string
		wantDeliveries int
	}{
		{name: "required", mode: config.EmailRequired, wantState: "pending_email", wantDeliveries: 1},
		{name: "grace", mode: config.EmailGrace, wantState: "pending_email", wantDeliveries: 1},
		{name: "disabled", mode: config.EmailDisabled, wantState: "active", wantDeliveries: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			tx := &fakeIdentityTransaction{}
			application, _, _, _ := newTask9Application(t, test.mode, tx)
			_, err := application.RegisterAccount(context.Background(), task9RegisterCommand())
			if err != nil {
				t.Fatal(err)
			}
			if tx.createdAccount.State != test.wantState || len(tx.events) != test.wantDeliveries {
				t.Fatalf("state/events = %q/%d", tx.createdAccount.State, len(tx.events))
			}
		})
	}

	for iteration := range 50 {
		for _, existing := range []bool{false, true} {
			tx := &fakeIdentityTransaction{identityFound: existing, identity: store.IdentityEmailIdentity{PrincipalID: uuid.New()}}
			application, deriver, _, _ := newTask9Application(t, config.EmailRequired, tx)
			result, err := application.RegisterAccount(context.Background(), task9RegisterCommand())
			if err != nil || !result.Accepted || deriver.calls != 1 {
				t.Fatalf("iteration %d existing=%v result=%#v error=%v derivations=%d", iteration, existing, result, err, deriver.calls)
			}
			if tx.deadlineClass < 4*time.Second || tx.deadlineClass > 5*time.Second {
				t.Fatalf("iteration %d existing=%v deadline class=%s", iteration, existing, tx.deadlineClass)
			}
		}
	}
}

func TestRegisterTransactionFailureRollsBackWithoutRemoteEmailCall(t *testing.T) {
	t.Parallel()

	tx := &fakeIdentityTransaction{failOperation: "append_event"}
	application, _, _, _ := newTask9Application(t, config.EmailRequired, tx)
	result, err := application.RegisterAccount(context.Background(), task9RegisterCommand())
	if result.Accepted || publicTask9Code(err) != apierrors.DependencyUnavailable {
		t.Fatalf("registration failure = %#v, %v", result, err)
	}
	if tx.operations[len(tx.operations)-1] != "rollback" {
		t.Fatalf("failure order = %v", tx.operations)
	}
	// EmailSender is intentionally absent from application dependencies. The
	// only delivery effect in this transaction is the protected outbox event.
	if len(tx.events) != 0 {
		t.Fatal("failed outbox append retained an externally visible event")
	}
}

func TestPostgresRepositoryContainsPanicsAndUsesIndependentBoundedRollback(t *testing.T) {
	t.Parallel()

	protector, err := sensitive.NewLocal(
		secret.NewBytes(bytes.Repeat([]byte{0x61}, 32)), secret.NewBytes(bytes.Repeat([]byte{0x62}, 32)), 3,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = protector.Close() })

	var typedNilBeginner *task9PGXBeginner
	if repository, createErr := NewPostgresRepository(typedNilBeginner, protector); !errors.Is(createErr, ErrInvalidRepository) || repository != nil {
		t.Fatalf("typed-nil repository = %#v, %v", repository, createErr)
	}
	beginPanic := &task9PGXBeginner{panicOnBegin: true}
	repository, err := NewPostgresRepository(beginPanic, protector)
	if err != nil {
		t.Fatal(err)
	}
	if runErr := repository.WithinTransaction(context.Background(), func(context.Context, Transaction) error { return nil }); !errors.Is(runErr, ErrRepository) {
		t.Fatalf("begin panic error = %v", runErr)
	}

	tx := &task9PGXTx{panicOnRollback: true}
	repository, err = NewPostgresRepository(&task9PGXBeginner{tx: tx}, protector)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	runErr := repository.WithinTransaction(ctx, func(_ context.Context, transaction Transaction) error {
		if transaction.DBTX() != tx {
			t.Fatal("repository did not preserve the caller-owned transaction")
		}
		cancel()
		panic("SECRET_TRANSACTION_CANARY")
	})
	if !errors.Is(runErr, ErrRepository) || strings.Contains(runErr.Error(), "SECRET_TRANSACTION_CANARY") {
		t.Fatalf("panic projection = %v", runErr)
	}
	if tx.commitCalls != 0 || tx.rollbackCalls != 1 || tx.rollbackContextErr != nil || !tx.rollbackDeadline {
		t.Fatalf("rollback boundary = commit:%d rollback:%d context:%v deadline:%v", tx.commitCalls, tx.rollbackCalls, tx.rollbackContextErr, tx.rollbackDeadline)
	}
}

func TestTask9BoundariesRejectTypedNilContext(t *testing.T) {
	t.Parallel()

	application, _, _, _ := newTask9Application(t, config.EmailRequired, &fakeIdentityTransaction{})
	var ctx *task9TypedNilContext
	if _, err := application.RegisterAccount(ctx, task9RegisterCommand()); publicTask9Code(err) != apierrors.MalformedRequest {
		t.Fatalf("typed-nil application context error = %v", err)
	}

	protector, err := sensitive.NewLocal(
		secret.NewBytes(bytes.Repeat([]byte{0x71}, 32)), secret.NewBytes(bytes.Repeat([]byte{0x72}, 32)), 2,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = protector.Close() })
	repository, err := NewPostgresRepository(&task9PGXBeginner{tx: &task9PGXTx{}}, protector)
	if err != nil {
		t.Fatal(err)
	}
	if runErr := repository.WithinTransaction(ctx, func(context.Context, Transaction) error { return nil }); !errors.Is(runErr, ErrInvalidRepository) {
		t.Fatalf("typed-nil repository context error = %v", runErr)
	}
}

var fixedTask9Time = time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)

func task9RegisterCommand() RegisterAccountCommand {
	return RegisterAccountCommand{
		Email: "member@example.test", Password: secret.NewBytes([]byte("correct horse battery staple")), Locale: "en",
		IdempotencyKey: "abcdefghijklmnopqrstuv",
	}
}

type task9Deriver struct{ calls int }

func (deriver *task9Deriver) Derive(_ []byte, _ []byte, policy PasswordPolicy) []byte {
	deriver.calls++
	return bytes.Repeat([]byte{0xa5}, int(policy.TagBytes))
}

type task9Clock struct{}

func (task9Clock) Now() time.Time { return fixedTask9Time }

type task9Random struct {
	mu      sync.Mutex
	counter byte
}

func (random *task9Random) Read(target []byte) (int, error) {
	random.mu.Lock()
	defer random.mu.Unlock()
	for index := range target {
		random.counter++
		target[index] = random.counter
	}
	return len(target), nil
}

type task9Limiter struct {
	calls   int
	allowed bool
	err     error
}

func (limiter *task9Limiter) Allow(_ context.Context, _ ratelimit.Operation, _ [32]byte, _ config.RateLimitPolicy) (bool, error) {
	limiter.calls++
	return limiter.allowed, limiter.err
}

type task9Participant struct {
	calls int
	tx    store.DBTX
	err   error
	owner *fakeIdentityTransaction
}

func (participant *task9Participant) ActivateVerifiedPrincipal(_ context.Context, tx store.DBTX, _ PrincipalID, _ time.Time) error {
	participant.calls++
	participant.tx = tx
	if participant.owner != nil {
		participant.owner.operations = append(participant.owner.operations, "participant")
	}
	return participant.err
}

func newTask9Application(t *testing.T, mode config.EmailVerificationMode, tx *fakeIdentityTransaction) (*Service, *task9Deriver, *task9Limiter, *task9Participant) {
	t.Helper()
	protector, err := sensitive.NewLocal(
		secret.NewBytes(bytes.Repeat([]byte{0x11}, 32)), secret.NewBytes(bytes.Repeat([]byte{0x22}, 32)), 7,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = protector.Close() })
	deriver := &task9Deriver{}
	limiter := &task9Limiter{allowed: true}
	participant := &task9Participant{owner: tx}
	ids := &task9UUIDs{}
	application, err := newApplicationForTest(ApplicationDependencies{
		Repository: &fakeIdentityRepository{tx: tx}, Protector: protector, Random: &task9Random{}, Clock: task9Clock{},
		Limiter: limiter, RateLimitKey: secret.NewBytes(bytes.Repeat([]byte{0x33}, 32)),
		Security: config.SecurityConfig{
			Profile: config.ProfileTest, EmailVerification: mode, RequestDeadline: 5 * time.Second, RedisTimeout: 250 * time.Millisecond,
			DeliveryRateLimit: config.RateLimitPolicy{Limit: 5, Window: time.Hour},
		},
		DeviceAuthorizationParticipant: participant,
	}, deriver.Derive, ids.Next)
	if err != nil {
		t.Fatal(err)
	}
	return application, deriver, limiter, participant
}

type task9UUIDs struct{ next uint64 }

func (ids *task9UUIDs) Next() uuid.UUID {
	ids.next++
	var value uuid.UUID
	binary.BigEndian.PutUint64(value[8:], ids.next)
	value[6] = 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return value
}

type fakeIdentityRepository struct{ tx *fakeIdentityTransaction }

func (repository *fakeIdentityRepository) WithinTransaction(ctx context.Context, operation func(context.Context, Transaction) error) error {
	if deadline, ok := ctx.Deadline(); ok {
		repository.tx.deadlineClass = time.Until(deadline)
	}
	err := operation(ctx, repository.tx)
	if err != nil {
		repository.tx.operations = append(repository.tx.operations, "rollback")
		return err
	}
	repository.tx.operations = append(repository.tx.operations, "commit")
	return nil
}

type fakeIdentityTransaction struct {
	operations    []string
	deadlineClass time.Duration
	failOperation string

	identityFound     bool
	identity          store.IdentityEmailIdentity
	verificationFound bool
	verification      store.IdentityEmailIdentity
	accountFound      bool
	account           store.IdentityAccount
	sessionFound      bool
	session           store.IdentityAccountSession
	credentialFound   bool
	credential        store.IdentityPasswordCredential
	consumeEmailOK    bool
	consumeResetOK    bool

	createdAccount    store.CreateAccountParams
	createdEmail      store.CreateEmailIdentityParams
	createdCredential store.CreatePasswordCredentialParams
	resetEmail        store.ResetEmailVerificationParams
	passwordReset     store.SetPasswordResetParams
	consumedReset     store.ConsumePasswordResetParams
	createdSession    store.CreateAccountSessionParams
	createdRefresh    store.InsertAccountRefreshTokenParams
	createdGrant      store.CreateEnrollmentGrantParams
	securityEvents    []store.InsertSecurityEventParams
	events            []*eventsv1.EventEnvelope
	dbtx              fakeTask9DBTX
}

func (tx *fakeIdentityTransaction) record(operation string) error {
	tx.operations = append(tx.operations, operation)
	if tx.failOperation == operation {
		return errors.New("SECRET_STORE_CANARY")
	}
	return nil
}

func (tx *fakeIdentityTransaction) DBTX() store.DBTX { return &tx.dbtx }
func (tx *fakeIdentityTransaction) BeginIdempotency(_ context.Context, _ idempotency.Scope, _ string, _ []byte, _, _ time.Time) (idempotency.Record, idempotency.Outcome, error) {
	return idempotency.Record{}, idempotency.Started, tx.record("begin_idempotency")
}
func (tx *fakeIdentityTransaction) CompleteIdempotency(_ context.Context, _ idempotency.Record, _ int, _ []byte) error {
	return tx.record("complete_idempotency")
}
func (tx *fakeIdentityTransaction) FindIdentityByLookupDigest(context.Context, []byte) (store.IdentityEmailIdentity, bool, error) {
	err := tx.record("find_identity")
	return tx.identity, tx.identityFound, err
}
func (tx *fakeIdentityTransaction) GetEmailVerificationForUpdate(context.Context, store.GetEmailVerificationForUpdateParams) (store.IdentityEmailIdentity, bool, error) {
	err := tx.record("get_verification")
	return tx.verification, tx.verificationFound, err
}
func (tx *fakeIdentityTransaction) GetAccountForUpdate(context.Context, uuid.UUID) (store.IdentityAccount, bool, error) {
	err := tx.record("get_account")
	return tx.account, tx.accountFound, err
}
func (tx *fakeIdentityTransaction) GetAccountSessionForUpdate(context.Context, store.GetAccountSessionForUpdateParams) (store.IdentityAccountSession, bool, error) {
	err := tx.record("get_session")
	return tx.session, tx.sessionFound, err
}
func (tx *fakeIdentityTransaction) GetPasswordCredential(context.Context, uuid.UUID) (store.IdentityPasswordCredential, bool, error) {
	err := tx.record("get_credential")
	return tx.credential, tx.credentialFound, err
}
func (tx *fakeIdentityTransaction) CreateAccount(_ context.Context, params store.CreateAccountParams) error {
	tx.createdAccount = params
	return tx.record("create_account")
}
func (tx *fakeIdentityTransaction) CreateEmailIdentity(_ context.Context, params store.CreateEmailIdentityParams) error {
	tx.createdEmail = params
	return tx.record("create_email_identity")
}
func (tx *fakeIdentityTransaction) CreatePasswordCredential(_ context.Context, params store.CreatePasswordCredentialParams) error {
	tx.createdCredential = params
	return tx.record("create_password_credential")
}
func (tx *fakeIdentityTransaction) InsertSecurityEvent(_ context.Context, params store.InsertSecurityEventParams) error {
	if err := tx.record("insert_security_event"); err != nil {
		return err
	}
	tx.securityEvents = append(tx.securityEvents, params)
	return nil
}
func (tx *fakeIdentityTransaction) ResetEmailVerification(_ context.Context, params store.ResetEmailVerificationParams) (bool, error) {
	tx.resetEmail = params
	return true, tx.record("reset_email_verification")
}
func (tx *fakeIdentityTransaction) SetPasswordReset(_ context.Context, params store.SetPasswordResetParams) (bool, error) {
	tx.passwordReset = params
	return true, tx.record("set_password_reset")
}
func (tx *fakeIdentityTransaction) ConsumeEmailVerification(context.Context, store.ConsumeEmailVerificationParams) (store.IdentityEmailIdentity, bool, error) {
	err := tx.record("consume_email")
	return tx.verification, tx.consumeEmailOK, err
}
func (tx *fakeIdentityTransaction) ClearPendingEmailDelivery(context.Context, store.ClearPendingEmailDeliveryParams) (int64, error) {
	return 1, tx.record("clear_delivery")
}
func (tx *fakeIdentityTransaction) ActivateVerifiedAccount(_ context.Context, params store.ActivateVerifiedAccountParams) (store.IdentityAccount, bool, error) {
	if err := tx.record("activate_account"); err != nil {
		return store.IdentityAccount{}, false, err
	}
	account := tx.account
	account.ID = params.ID
	account.State = "active"
	account.StateVersion++
	return account, true, nil
}
func (tx *fakeIdentityTransaction) ConsumePasswordReset(_ context.Context, params store.ConsumePasswordResetParams) (bool, error) {
	tx.consumedReset = params
	return tx.consumeResetOK, tx.record("consume_password_reset")
}
func (tx *fakeIdentityTransaction) MarkPrincipalSessionsReviewRequired(context.Context, store.MarkPrincipalSessionsReviewRequiredParams) (int64, error) {
	return 2, tx.record("mark_sessions_review_required")
}
func (tx *fakeIdentityTransaction) CreateAccountSession(_ context.Context, params store.CreateAccountSessionParams) error {
	tx.createdSession = params
	return tx.record("create_account_session")
}
func (tx *fakeIdentityTransaction) InsertAccountRefreshToken(_ context.Context, params store.InsertAccountRefreshTokenParams) error {
	tx.createdRefresh = params
	return tx.record("insert_account_refresh")
}
func (tx *fakeIdentityTransaction) CreateEnrollmentGrant(_ context.Context, params store.CreateEnrollmentGrantParams) error {
	tx.createdGrant = params
	return tx.record("create_enrollment_grant")
}
func (tx *fakeIdentityTransaction) AppendEvent(_ context.Context, envelope *eventsv1.EventEnvelope) error {
	if err := tx.record("append_event"); err != nil {
		return err
	}
	tx.events = append(tx.events, proto.Clone(envelope).(*eventsv1.EventEnvelope))
	return nil
}

func (tx *fakeIdentityTransaction) mutationCount() int {
	return len(tx.securityEvents) + len(tx.events) + boolInt(tx.createdAccount.ID != uuid.Nil) + boolInt(tx.createdEmail.ID != uuid.Nil) + boolInt(tx.createdCredential.PrincipalID != uuid.Nil)
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

type fakeTask9DBTX struct{}

func (*fakeTask9DBTX) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (*fakeTask9DBTX) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, errors.New("not used")
}
func (*fakeTask9DBTX) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	return fakeTask9Row{}
}

type fakeTask9Row struct{}

func (fakeTask9Row) Scan(...interface{}) error { return errors.New("not used") }

type task9PGXBeginner struct {
	tx           pgx.Tx
	panicOnBegin bool
}

func (beginner *task9PGXBeginner) Begin(context.Context) (pgx.Tx, error) {
	if beginner.panicOnBegin {
		panic("SECRET_BEGIN_CANARY")
	}
	return beginner.tx, nil
}

type task9PGXTx struct {
	pgx.Tx
	commitCalls        int
	rollbackCalls      int
	panicOnRollback    bool
	rollbackContextErr error
	rollbackDeadline   bool
}

type task9TypedNilContext struct{}

func (*task9TypedNilContext) Deadline() (time.Time, bool) { panic("typed nil context used") }
func (*task9TypedNilContext) Done() <-chan struct{}       { panic("typed nil context used") }
func (*task9TypedNilContext) Err() error                  { panic("typed nil context used") }
func (*task9TypedNilContext) Value(any) any               { panic("typed nil context used") }

func (tx *task9PGXTx) Commit(context.Context) error {
	tx.commitCalls++
	return nil
}

func (tx *task9PGXTx) Rollback(ctx context.Context) error {
	tx.rollbackCalls++
	tx.rollbackContextErr = ctx.Err()
	_, tx.rollbackDeadline = ctx.Deadline()
	if tx.panicOnRollback {
		panic("SECRET_ROLLBACK_CANARY")
	}
	return nil
}

func assertTask9Order(t *testing.T, got, want []string) {
	t.Helper()
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("operation order = %v, want %v", got, want)
	}
}

func assertProtectedDelivery(t *testing.T, protector sensitive.Protector, domain string, ciphertext []byte, keyVersion int32, recipient, template, locale string) {
	t.Helper()
	if keyVersion < 1 {
		t.Fatal("invalid protected-delivery key version")
	}
	// #nosec G115 -- positivity is checked immediately above.
	opened, err := protector.Decrypt(domain, sensitive.EncryptedField{KeyVersion: uint32(keyVersion), Ciphertext: ciphertext})
	if err != nil {
		t.Fatal("decrypt protected delivery")
	}
	defer clear(opened)
	if len(opened) > 4096 {
		t.Fatalf("pending delivery length = %d", len(opened))
	}
	var payload map[string]string
	if err = json.Unmarshal(opened, &payload); err != nil {
		t.Fatal("decode protected delivery")
	}
	if len(payload) != 4 || payload["recipient"] != recipient || payload["template"] != template || payload["locale"] != locale || len(payload["token"]) != 43 {
		t.Fatalf("protected delivery schema = %#v", payload)
	}
}

func assertPrivateEmailDeliveryEvent(t *testing.T, envelope *eventsv1.EventEnvelope, deliveryID, principalID uuid.UUID, template TemplateID, locale string, forbidden ...[]byte) {
	t.Helper()
	if envelope.GetEventType() != contractevents.EmailDeliveryRequestedType {
		t.Fatalf("event type = %q", envelope.GetEventType())
	}
	for _, value := range forbidden {
		if bytes.Contains(envelope.GetPayload(), value) {
			t.Fatal("delivery event exposed forbidden request material")
		}
	}
	payload := new(identityv1.EmailDeliveryRequested)
	if err := proto.Unmarshal(envelope.GetPayload(), payload); err != nil {
		t.Fatal("decode delivery event")
	}
	if payload.GetDeliveryId() != deliveryID.String() || payload.GetPrincipalId() != principalID.String() || payload.GetTemplateId() != string(template) || payload.GetLocale() != locale {
		t.Fatalf("delivery event = %#v", payload)
	}
}

func publicTask9Code(err error) apierrors.Code {
	var classified apierrors.Error
	if !errors.As(err, &classified) {
		return ""
	}
	return apierrors.Code(classified.Public("trace-safe-000009").Code)
}

var _ = math.MaxInt32
var _ = pgtype.Int4{}
var _ = sql.NullTime{}
var _ securitykit.Clock = task9Clock{}
