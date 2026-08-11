package identity

import (
	"bytes"
	"context"
	"encoding/base32"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/store"
)

func TestConcurrentRecoveryConsumeCreatesExactlyOneSession(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	codes, hashes, err := generateRecoveryCodeSet(&task11RecoveryRandom{})
	if err != nil {
		t.Fatal(err)
	}
	defer clearRecoveryHashes(hashes)
	transaction.identityFound = true
	transaction.identity = store.IdentityEmailIdentity{PrincipalID: transaction.account.ID}
	transaction.recoveryFound = true
	transaction.recovery = store.IdentityRecoveryCodeSet{
		ID: uuid.MustParse("581c1f7a-6d76-44da-aa95-461fa42c929e"), PrincipalID: transaction.account.ID,
		Generation: 2, CodeHashes: cloneTask11Hashes(hashes), State: "active",
	}
	transaction.consumeRecoveryFound = true
	transaction.statefulRecovery = true
	application, _ := newTask11TOTPApplication(t, transaction)
	start := make(chan struct{})
	results := make(chan error, 2)
	clientKeys := [...][32]byte{{1}, {2}}
	idempotencyKeys := [...]string{"task11-recovery-race-01", "task11-recovery-race-02"}
	var workers sync.WaitGroup
	for index := range 2 {
		workers.Add(1)
		go func(worker int) {
			defer workers.Done()
			<-start
			_, consumeErr := application.ConsumeRecoveryCode(context.Background(), ConsumeRecoveryCodeCommand{
				Email: "member@example.test", Code: secret.NewBytes([]byte(codes[0])),
				ClientSigningPublicKey: clientKeys[worker],
				IdempotencyKey:         idempotencyKeys[worker],
			})
			results <- consumeErr
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
		t.Fatalf("concurrent recovery results = %d success/%d failure, want 1/1", succeeded, failed)
	}
	operations := strings.Join(transaction.operations, ",")
	if strings.Count(operations, "create_account_session") != 1 || strings.Count(operations, "insert_account_refresh") != 1 {
		t.Fatalf("concurrent recovery session mutations = %s", operations)
	}
}

func TestRecoveryConsumeRemovesOneDigestAndCreatesSessionInSameTransaction(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	codes, hashes, err := generateRecoveryCodeSet(&task11RecoveryRandom{})
	if err != nil {
		t.Fatal(err)
	}
	defer clearRecoveryHashes(hashes)
	transaction.identityFound = true
	transaction.identity = store.IdentityEmailIdentity{PrincipalID: transaction.account.ID}
	transaction.recoveryFound = true
	transaction.recovery = store.IdentityRecoveryCodeSet{
		ID: uuid.MustParse("d3f51fb9-26a3-46cf-a5df-b5cd5edfc508"), PrincipalID: transaction.account.ID,
		Generation: 8, CodeHashes: cloneTask11Hashes(hashes), State: "active",
	}
	transaction.consumeRecoveryFound = true
	application, _ := newTask11TOTPApplication(t, transaction)
	clientKey := [32]byte{1, 2, 3, 4}
	result, err := application.ConsumeRecoveryCode(context.Background(), ConsumeRecoveryCodeCommand{
		Email: "member@example.test", Code: secret.NewBytes([]byte(codes[0])),
		ClientSigningPublicKey: clientKey, IdempotencyKey: "task11-recovery-consume-1",
	})
	if err != nil {
		t.Fatalf("consume recovery code: %v", err)
	}
	accessToken := result.AccessToken.Copy()
	refreshToken := result.RefreshToken.Copy()
	defer clear(accessToken)
	defer clear(refreshToken)
	if len(accessToken) == 0 || len(refreshToken) == 0 {
		t.Fatal("recovery consumption did not return a session")
	}
	wantDigest, valid := recoveryCodeDigest(codes[0])
	if !valid || transaction.consumedRecovery.ID != transaction.recovery.ID ||
		!bytes.Equal(transaction.consumedRecovery.Column2, wantDigest[:]) {
		t.Fatalf("consumed digest = %x, want target digest", transaction.consumedRecovery.Column2)
	}
	clear(wantDigest[:])
	if transaction.createdSession.PrincipalID != transaction.account.ID || !bytes.Equal(transaction.createdSession.ClientSigningPublicKey, clientKey[:]) {
		t.Fatalf("created session binding = %+v", transaction.createdSession)
	}
	if transaction.createdRefresh.SessionID != transaction.createdSession.ID {
		t.Fatal("refresh token was not bound to the created session")
	}
	wantOperations := []string{
		"find_identity", "get_recovery", "get_account", "begin_idempotency", "consume_recovery",
		"create_account_session", "insert_account_refresh", "complete_idempotency", "commit",
	}
	if got := strings.Join(transaction.operations, ","); got != strings.Join(wantOperations, ",") {
		t.Fatalf("operation order = %s, want %s", got, strings.Join(wantOperations, ","))
	}

	transaction.operations = nil
	transaction.consumeRecoveryFound = false
	_, err = application.ConsumeRecoveryCode(context.Background(), ConsumeRecoveryCodeCommand{
		Email: "member@example.test", Code: secret.NewBytes([]byte(codes[0])),
		ClientSigningPublicKey: clientKey, IdempotencyKey: "task11-recovery-consume-2",
	})
	if publicTask9Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("reused recovery code error = %v", err)
	}
	if strings.Contains(strings.Join(transaction.operations, ","), "create_account_session") {
		t.Fatal("failed recovery consumption created a session")
	}
}

func TestRecoveryConsumeRollsBackDigestWhenSessionIssuanceFails(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	codes, hashes, err := generateRecoveryCodeSet(&task11RecoveryRandom{})
	if err != nil {
		t.Fatal(err)
	}
	defer clearRecoveryHashes(hashes)
	transaction.identityFound = true
	transaction.identity = store.IdentityEmailIdentity{PrincipalID: transaction.account.ID}
	transaction.recoveryFound = true
	transaction.recovery = store.IdentityRecoveryCodeSet{
		ID: uuid.MustParse("0baa07f7-3505-438f-bf49-d042e8e8eb8d"), PrincipalID: transaction.account.ID,
		Generation: 3, CodeHashes: cloneTask11Hashes(hashes), State: "active",
	}
	transaction.consumeRecoveryFound = true
	transaction.statefulRecovery = true
	transaction.failOperation = "insert_account_refresh"
	application, _ := newTask11TOTPApplication(t, transaction)
	command := ConsumeRecoveryCodeCommand{
		Email: "member@example.test", Code: secret.NewBytes([]byte(codes[0])),
		ClientSigningPublicKey: [32]byte{4, 5, 6}, IdempotencyKey: "task11-recovery-rollback-1",
	}
	_, err = application.ConsumeRecoveryCode(context.Background(), command)
	if publicTask9Code(err) != apierrors.DependencyUnavailable {
		t.Fatalf("failed session issuance error = %v", err)
	}
	digest, valid := recoveryCodeDigest(codes[0])
	if !valid || !containsRecoveryHash(transaction.recovery.CodeHashes, digest[:]) {
		t.Fatal("failed session issuance consumed the recovery digest")
	}
	clear(digest[:])
	if transaction.createdSession.ID != uuid.Nil || transaction.createdRefresh.SessionID != uuid.Nil {
		t.Fatal("failed transaction retained session mutations")
	}
	if len(transaction.operations) == 0 || transaction.operations[len(transaction.operations)-1] != "rollback" {
		t.Fatalf("failed transaction operations = %v", transaction.operations)
	}

	transaction.failOperation = ""
	command.IdempotencyKey = "task11-recovery-rollback-2"
	if _, err = application.ConsumeRecoveryCode(context.Background(), command); err != nil {
		t.Fatalf("consume recovery code after rollback: %v", err)
	}
}

func TestRecoveryLastCodeConsumptionReplaysAfterSetExhaustion(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	codes, hashes, err := generateRecoveryCodeSet(&task11RecoveryRandom{})
	if err != nil {
		t.Fatal(err)
	}
	defer clearRecoveryHashes(hashes)
	transaction.identityFound = true
	transaction.identity = store.IdentityEmailIdentity{PrincipalID: transaction.account.ID}
	transaction.recoveryFound = true
	transaction.recovery = store.IdentityRecoveryCodeSet{
		ID: uuid.MustParse("4953473f-3c2c-4033-862f-122dad2ca10f"), PrincipalID: transaction.account.ID,
		Generation: 7, CodeHashes: cloneTask11Hashes(hashes[:1]), State: "active",
	}
	transaction.consumeRecoveryFound = true
	transaction.statefulRecovery = true
	application, _ := newTask11TOTPApplication(t, transaction)
	command := ConsumeRecoveryCodeCommand{
		Email: "member@example.test", Code: secret.NewBytes([]byte(codes[0])),
		ClientSigningPublicKey: [32]byte{7, 8, 9}, IdempotencyKey: "task11-recovery-last-code-1",
	}
	first, err := application.ConsumeRecoveryCode(context.Background(), command)
	if err != nil {
		t.Fatalf("consume last recovery code: %v", err)
	}
	firstAccess := first.AccessToken.Copy()
	firstRefresh := first.RefreshToken.Copy()
	defer clear(firstAccess)
	defer clear(firstRefresh)
	transaction.recoveryFound = false
	replayed, err := application.ConsumeRecoveryCode(context.Background(), command)
	if err != nil {
		t.Fatalf("replay last recovery consumption: %v", err)
	}
	replayedAccess := replayed.AccessToken.Copy()
	replayedRefresh := replayed.RefreshToken.Copy()
	defer clear(replayedAccess)
	defer clear(replayedRefresh)
	if !bytes.Equal(replayedAccess, firstAccess) || !bytes.Equal(replayedRefresh, firstRefresh) ||
		strings.Count(strings.Join(transaction.operations, ","), "create_account_session") != 1 {
		t.Fatalf("last-code replay created a second session: %v", transaction.operations)
	}
}

func TestRecoveryRotationSupersedesOldSetAndPersistsOnlyDigests(t *testing.T) {
	t.Parallel()

	transaction := activeTask11TOTPTransaction()
	transaction.recoveryFound = true
	transaction.recovery = store.IdentityRecoveryCodeSet{
		ID:          uuid.MustParse("2758e4b3-3525-4a30-8d3d-b70d9beae277"),
		PrincipalID: transaction.account.ID, Generation: 3, State: "active", CodeHashes: task11RecoveryHashes(),
	}
	transaction.nextRecoveryGeneration = 4
	application, _ := newTask11TOTPApplication(t, transaction)
	command := RotateRecoveryCodesCommand{
		PrincipalID:      PrincipalID(transaction.account.ID.String()),
		Reauthentication: task11PasswordReauthentication(transaction.session.ID),
		IdempotencyKey:   "task11-recovery-rotate-01",
	}
	result, err := application.RotateRecoveryCodes(context.Background(), command)
	if err != nil {
		t.Fatalf("rotate recovery codes: %v", err)
	}
	if len(result.Codes) != recoveryCodeCount || len(transaction.createdRecovery.CodeHashes) != recoveryCodeCount {
		t.Fatalf("returned/persisted recovery values = %d/%d, want %d/%d", len(result.Codes), len(transaction.createdRecovery.CodeHashes), recoveryCodeCount, recoveryCodeCount)
	}
	if transaction.createdRecovery.PrincipalID != transaction.account.ID || transaction.createdRecovery.Generation != 4 || transaction.createdRecovery.ID == uuid.Nil {
		t.Fatalf("unexpected recovery set metadata: %+v", transaction.createdRecovery)
	}
	if transaction.revokedRecovery.State != "superseded" || transaction.revokedRecovery.PrincipalID != transaction.account.ID {
		t.Fatalf("old recovery set was not superseded: %+v", transaction.revokedRecovery)
	}
	for index, code := range result.Codes {
		digest, valid := recoveryCodeDigest(code)
		if !valid || len(transaction.createdRecovery.CodeHashes[index]) != 32 || !bytes.Equal(transaction.createdRecovery.CodeHashes[index], digest[:]) {
			t.Fatalf("persisted recovery digest %d does not match returned code", index)
		}
		for _, persisted := range transaction.createdRecovery.CodeHashes {
			if bytes.Contains(persisted, []byte(code)) {
				t.Fatalf("plaintext recovery code %d reached persistence", index)
			}
		}
		clear(digest[:])
	}
	wantOperations := []string{
		"get_credential", "get_recovery", "get_session", "get_account", "begin_idempotency",
		"get_next_recovery_generation", "revoke_recovery", "create_recovery", "complete_idempotency", "commit",
	}
	if got := strings.Join(transaction.operations, ","); got != strings.Join(wantOperations, ",") {
		t.Fatalf("operation order = %s, want %s", got, strings.Join(wantOperations, ","))
	}
	replayed, err := application.RotateRecoveryCodes(context.Background(), command)
	if err != nil || strings.Join(replayed.Codes, ",") != strings.Join(result.Codes, ",") ||
		strings.Count(strings.Join(transaction.operations, ","), "create_recovery") != 1 {
		t.Fatalf("recovery rotation replay = %#v/%v, operations=%v", replayed, err, transaction.operations)
	}
}

func TestRecoveryGenerationCreatesTenTwentyByteGroupedCodesAndOnlyDigests(t *testing.T) {
	t.Parallel()

	random := &task11RecoveryRandom{}
	codes, hashes, err := generateRecoveryCodeSet(random)
	if err != nil {
		t.Fatalf("generate recovery codes: %v", err)
	}
	if len(codes) != 10 || len(hashes) != 10 {
		t.Fatalf("codes/hashes = %d/%d, want 10/10", len(codes), len(hashes))
	}
	seen := make(map[string]struct{}, len(codes))
	for index, code := range codes {
		parts := strings.Split(code, "-")
		if len(parts) != 8 {
			t.Fatalf("code %d groups = %d, want 8", index, len(parts))
		}
		for _, part := range parts {
			if len(part) != 4 {
				t.Fatalf("code %d group length = %d, want 4", index, len(part))
			}
		}
		compact := strings.ReplaceAll(code, "-", "")
		raw, decodeErr := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(compact)
		if decodeErr != nil || len(raw) != 20 {
			t.Fatalf("code %d decoded length/error = %d/%v", index, len(raw), decodeErr)
		}
		wantDigest, valid := recoveryCodeDigest(code)
		if !valid || !bytes.Equal(hashes[index], wantDigest[:]) || bytes.Contains(hashes[index], raw) {
			t.Fatalf("code %d digest contract failed", index)
		}
		if _, duplicate := seen[code]; duplicate {
			t.Fatalf("duplicate code %d", index)
		}
		seen[code] = struct{}{}
		clear(raw)
	}
}

type task11RecoveryRandom struct {
	next byte
}

func (random *task11RecoveryRandom) Read(target []byte) (int, error) {
	for index := range target {
		random.next++
		target[index] = random.next
	}
	return len(target), nil
}

func TestRecoveryGeneratedQueryGetsNextGenerationInsideCallerTransaction(t *testing.T) {
	t.Parallel()

	principalID := uuid.New()
	db := &task11GenerationDBTX{generation: 4}
	queries := store.New(db)

	generation, err := queries.GetNextRecoveryCodeGeneration(context.Background(), principalID)
	if err != nil {
		t.Fatalf("get next recovery generation: %v", err)
	}
	if generation != 4 {
		t.Fatalf("generation = %d, want 4", generation)
	}
	if !strings.Contains(db.lastSQL, "MAX(generation)") {
		t.Fatalf("generation query must derive the next generation: %s", db.lastSQL)
	}
	if len(db.lastArgs) != 1 || db.lastArgs[0] != principalID {
		t.Fatalf("generation args = %#v, want principal id", db.lastArgs)
	}
}

type task11GenerationDBTX struct {
	lastSQL    string
	lastArgs   []any
	generation int32
}

func (*task11GenerationDBTX) Exec(
	context.Context,
	string,
	...any,
) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec call")
}

func (*task11GenerationDBTX) Query(
	context.Context,
	string,
	...any,
) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query call")
}

func (db *task11GenerationDBTX) QueryRow(
	_ context.Context,
	query string,
	args ...any,
) pgx.Row {
	db.lastSQL = query
	db.lastArgs = append([]any(nil), args...)
	return task11GenerationRow{generation: db.generation}
}

type task11GenerationRow struct {
	generation int32
}

func (row task11GenerationRow) Scan(destinations ...any) error {
	if len(destinations) != 1 {
		return errors.New("unexpected destination count")
	}
	destination, ok := destinations[0].(*int32)
	if !ok {
		return errors.New("unexpected destination type")
	}
	*destination = row.generation
	return nil
}

var _ store.DBTX = (*task11GenerationDBTX)(nil)

func (transaction *task11TOTPTransaction) GetActiveRecoveryCodeSetForUpdate(
	context.Context,
	uuid.UUID,
) (store.IdentityRecoveryCodeSet, bool, error) {
	return transaction.recovery, transaction.recoveryFound, transaction.record("get_recovery")
}

func (transaction *task11TOTPTransaction) GetNextRecoveryCodeGeneration(
	context.Context,
	uuid.UUID,
) (int32, error) {
	return transaction.nextRecoveryGeneration, transaction.record("get_next_recovery_generation")
}

func (transaction *task11TOTPTransaction) RevokeRecoveryCodeSets(
	_ context.Context,
	params store.RevokeRecoveryCodeSetsParams,
) (int64, error) {
	transaction.revokedRecovery = params
	return 1, transaction.record("revoke_recovery")
}

func (transaction *task11TOTPTransaction) CreateRecoveryCodeSet(
	_ context.Context,
	params store.CreateRecoveryCodeSetParams,
) error {
	transaction.createdRecovery = params
	transaction.createdRecovery.CodeHashes = cloneTask11Hashes(params.CodeHashes)
	return transaction.record("create_recovery")
}

func (transaction *task11TOTPTransaction) ConsumeRecoveryCode(
	_ context.Context,
	params store.ConsumeRecoveryCodeParams,
) (store.IdentityRecoveryCodeSet, bool, error) {
	transaction.consumedRecovery = params
	transaction.consumedRecovery.Column2 = append([]byte(nil), params.Column2...)
	if err := transaction.record("consume_recovery"); err != nil {
		return store.IdentityRecoveryCodeSet{}, false, err
	}
	if !transaction.consumeRecoveryFound {
		return store.IdentityRecoveryCodeSet{}, false, nil
	}
	if transaction.statefulRecovery && !containsRecoveryHash(transaction.recovery.CodeHashes, params.Column2) {
		return store.IdentityRecoveryCodeSet{}, false, nil
	}
	updated := transaction.recovery
	updated.CodeHashes = make([][]byte, 0, len(transaction.recovery.CodeHashes)-1)
	for _, hash := range transaction.recovery.CodeHashes {
		if !bytes.Equal(hash, params.Column2) {
			updated.CodeHashes = append(updated.CodeHashes, append([]byte(nil), hash...))
		}
	}
	if transaction.statefulRecovery {
		if len(updated.CodeHashes) == 0 {
			updated.State = "exhausted"
		}
		transaction.recovery = updated
	}
	return updated, true, nil
}

func cloneTask11Hashes(source [][]byte) [][]byte {
	result := make([][]byte, len(source))
	for index := range source {
		result[index] = append([]byte(nil), source[index]...)
	}
	return result
}

func task11RecoveryHashes() [][]byte {
	hashes := make([][]byte, recoveryCodeCount)
	for index := range hashes {
		hashes[index] = bytes.Repeat([]byte{byte(index + 1)}, 32)
	}
	return hashes
}
