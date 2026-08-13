package trust

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/deviceauth"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/store"
)

func TestIssueOrdersReplayAuthorizationVersionCryptoAndPersistence(t *testing.T) {
	fixture := newTask16ApplicationFixture(t)
	command := fixture.issueCommand("task16-issue-key-0000000001", "1", "hello")

	issued, err := fixture.application.Issue(context.Background(), command)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	defer issued.Locator.Clear()
	issuedLocator := issued.Locator.Copy()
	defer clear(issuedLocator)
	if issued.BundleVersion != 1 || issued.BundleID == uuid.Nil || len(issuedLocator) != 32 || issued.EnvelopeSHA256 == [32]byte{} {
		t.Fatal("Issue returned an incomplete immutable-bundle handle")
	}
	wantOrder := []string{"begin:issue_bundle", "authorize", "version", "random:16", "random:16", "random:32", "sign", "random:padding", "insert", "event", "complete:200", "commit"}
	if got := fixture.trace.snapshot(); !reflect.DeepEqual(got, wantOrder) {
		t.Fatalf("operation order = %v, want %v", got, wantOrder)
	}
	if fixture.participant.dbtx != fixture.repository.transaction {
		t.Fatal("authorization participant did not receive the trust transaction DBTX")
	}
	if fixture.participant.lastAuthorization != command.AuthorizationID {
		t.Fatal("authorization result was not bound to the command authorization")
	}

	firstLocator := issued.Locator.Copy()
	fixture.trace.reset()
	fixture.participant.err = errors.New("AUTHORITY_SECRET_CANARY")
	_ = fixture.signer.Close()
	replayed, err := fixture.application.Issue(context.Background(), command)
	if err != nil {
		clear(firstLocator)
		t.Fatalf("Issue replay after authority/signer outage: %v", err)
	}
	defer replayed.Locator.Clear()
	replayedLocator := replayed.Locator.Copy()
	defer clear(replayedLocator)
	if replayed.BundleID != issued.BundleID || replayed.BundleVersion != issued.BundleVersion ||
		replayed.EnvelopeSHA256 != issued.EnvelopeSHA256 || !bytes.Equal(replayedLocator, firstLocator) {
		clear(firstLocator)
		t.Fatal("successful retry did not return the exact original response")
	}
	clear(firstLocator)
	if got := fixture.trace.snapshot(); !reflect.DeepEqual(got, []string{"begin:issue_bundle", "commit"}) {
		t.Fatalf("replay consulted mutable authority or crypto: %v", got)
	}
}

func TestIssueRejectsDifferentBodyAndRollsBackFailedNewVersion(t *testing.T) {
	fixture := newTask16ApplicationFixture(t)
	first := fixture.issueCommand("task16-conflict-key-0000001", "1", "first")
	issued, err := fixture.application.Issue(context.Background(), first)
	if err != nil {
		t.Fatalf("first Issue: %v", err)
	}
	issued.Locator.Clear()

	fixture.trace.reset()
	changed := first
	changed.TestConfig.Message = "different"
	if _, err := fixture.application.Issue(context.Background(), changed); err == nil {
		t.Fatal("same idempotency key with a different body was accepted")
	}
	if got := fixture.trace.snapshot(); !reflect.DeepEqual(got, []string{"begin:issue_bundle", "rollback"}) {
		t.Fatalf("conflict consulted authority or crypto: %v", got)
	}

	fixture.participant.err = nil
	fixture.trace.reset()
	_ = fixture.signer.Close()
	second := fixture.issueCommand("task16-new-version-key-0001", "2", "second")
	if _, err := fixture.application.Issue(context.Background(), second); err == nil {
		t.Fatal("new issuance succeeded after signer stopped")
	}
	if fixture.repository.transaction.highest != 1 || len(fixture.repository.transaction.bundles) != 1 {
		t.Fatal("failed issuance did not roll back its allocated version and bytes")
	}
	if got := fixture.trace.snapshot(); !containsTask16Sequence(got, []string{"begin:issue_bundle", "authorize", "version", "rollback"}) {
		t.Fatalf("failed issuance order = %v", got)
	}
}

func TestConcurrentIssueReplaysOneExactBundleWithoutDuplicateVersionOrStorage(t *testing.T) {
	fixture := newTask16ApplicationFixture(t)
	command := fixture.issueCommand("task16-concurrent-key-000001", "1", "concurrent")
	defer command.AccessToken.Clear()
	type outcome struct {
		bundle IssuedBundle
		err    error
	}
	start := make(chan struct{})
	results := make(chan outcome, 2)
	for range 2 {
		go func() {
			<-start
			bundle, err := fixture.application.Issue(context.Background(), command)
			results <- outcome{bundle: bundle, err: err}
		}()
	}
	close(start)
	first, second := <-results, <-results
	defer first.bundle.Locator.Clear()
	defer second.bundle.Locator.Clear()
	if first.err != nil || second.err != nil {
		t.Fatalf("concurrent Issue errors = (%v, %v)", first.err, second.err)
	}
	firstLocator, secondLocator := first.bundle.Locator.Copy(), second.bundle.Locator.Copy()
	defer clear(firstLocator)
	defer clear(secondLocator)
	if first.bundle.BundleID != second.bundle.BundleID || first.bundle.BundleVersion != second.bundle.BundleVersion ||
		first.bundle.EnvelopeSHA256 != second.bundle.EnvelopeSHA256 || !bytes.Equal(firstLocator, secondLocator) {
		t.Fatal("concurrent idempotent Issue returned different immutable handles")
	}
	if fixture.repository.transaction.highest != 1 || len(fixture.repository.transaction.bundles) != 1 {
		t.Fatalf("concurrent Issue stored highest=%d bundles=%d, want 1/1",
			fixture.repository.transaction.highest, len(fixture.repository.transaction.bundles))
	}
}

func TestIssueRejectsAuthorityAuthorizationMismatchWithoutSideEffects(t *testing.T) {
	fixture := newTask16ApplicationFixture(t)
	fixture.participant.authority.AuthorizationID = uuid.MustParse("ffffffff-ffff-4fff-8fff-ffffffffffff")
	command := fixture.issueCommand("task16-authority-mismatch-01", "1", "mismatch")
	defer command.AccessToken.Clear()
	if _, err := fixture.application.Issue(context.Background(), command); task16PublicCode(err) != apierrors.AuthenticationFailed {
		t.Fatalf("authority mismatch error = %v, want authentication_failed", err)
	}
	transaction := fixture.repository.transaction
	if transaction.highest != 0 || len(transaction.bundles) != 0 || len(transaction.events) != 0 || len(transaction.idempotency) != 0 {
		t.Fatal("authority mismatch mutated issuance, event, version, or idempotency state")
	}
	if got := fixture.trace.snapshot(); !reflect.DeepEqual(got, []string{"begin:issue_bundle", "authorize", "rollback"}) {
		t.Fatalf("authority mismatch order = %v", got)
	}
}

func TestResolveReusesCurrentBundleBeforeCurrentSequenceAndIssuesOnAdvance(t *testing.T) {
	fixture := newTask16ApplicationFixture(t)
	query := fixture.resolveQuery("task16-resolve-key-000001")

	first, err := fixture.application.Resolve(context.Background(), query)
	if err != nil {
		t.Fatalf("first Resolve: %v", err)
	}
	if first.Locator == "" || first.EnvelopeSHA256 == "" || first.CacheControl != ImmutableCacheControl {
		t.Fatal("first Resolve returned an incomplete resolution")
	}
	if fixture.repository.transaction.highest != 1 {
		t.Fatalf("first Resolve highest = %d, want 1", fixture.repository.transaction.highest)
	}

	fixture.trace.reset()
	fixture.sequence.set(TestConfigV1{Message: "must-not-be-read", Sequence: "2"})
	fixture.participant.err = errors.New("AUTHORITY_SECRET_CANARY")
	replayed, err := fixture.application.Resolve(context.Background(), query)
	if err != nil {
		t.Fatalf("Resolve replay after mutable-state outage: %v", err)
	}
	if replayed != first {
		t.Fatal("Resolve replay did not return byte-identical fields")
	}
	if got := fixture.trace.snapshot(); !reflect.DeepEqual(got, []string{"begin:resolve_bundle", "commit"}) {
		t.Fatalf("Resolve replay consulted current sequence or authority: %v", got)
	}

	fixture.participant.err = nil
	fixture.trace.reset()
	second, err := fixture.application.Resolve(context.Background(), fixture.resolveQuery("task16-resolve-key-000002"))
	if err != nil {
		t.Fatalf("Resolve advanced sequence: %v", err)
	}
	if second.Locator == first.Locator || second.EnvelopeSHA256 == first.EnvelopeSHA256 || fixture.repository.transaction.highest != 2 {
		t.Fatal("advanced server sequence did not publish a new immutable version")
	}

	fixture.trace.reset()
	fixture.sequence.set(TestConfigV1{Message: "same", Sequence: "2"})
	third, err := fixture.application.Resolve(context.Background(), fixture.resolveQuery("task16-resolve-key-000003"))
	if err != nil {
		t.Fatalf("Resolve current sequence: %v", err)
	}
	if third != second || fixture.repository.transaction.highest != 2 {
		t.Fatal("current unexpired bundle was not reused")
	}
}

func TestAcknowledgeIsIdempotentAndCannotLowerIssuedVersion(t *testing.T) {
	fixture := newTask16ApplicationFixture(t)
	first, err := fixture.application.Issue(context.Background(), fixture.issueCommand("task16-ack-issue-key-00001", "1", "hello"))
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	first.Locator.Clear()
	second, err := fixture.application.Issue(context.Background(), fixture.issueCommand("task16-ack-issue-key-00002", "2", "hello"))
	if err != nil {
		t.Fatalf("second Issue: %v", err)
	}
	second.Locator.Clear()

	command := AcknowledgeCommand{
		AuthorizationID: fixture.authorizationID,
		AccessToken:     secret.NewBytes([]byte("device-access-token-ack")),
		BundleID:        first.BundleID,
		BundleVersion:   first.BundleVersion,
		IdempotencyKey:  "task16-ack-key-000000000001",
	}
	if err := fixture.application.Acknowledge(context.Background(), command); err != nil {
		t.Fatalf("Acknowledge: %v", err)
	}
	fixture.trace.reset()
	fixture.participant.err = errors.New("ACK-AUTHORITY-OUTAGE-CANARY")
	if err := fixture.application.Acknowledge(context.Background(), command); err != nil {
		t.Fatalf("Acknowledge replay during authority outage: %v", err)
	}
	if got := fixture.trace.snapshot(); !reflect.DeepEqual(got, []string{"begin:acknowledge_bundle", "commit"}) {
		t.Fatalf("Acknowledge replay consulted authority: %v", got)
	}
	if fixture.repository.transaction.highest != 2 || len(fixture.repository.transaction.acknowledgements) != 1 {
		t.Fatal("acknowledgement lowered issuance state or duplicated its row")
	}
}

func TestTrustApplicationSecretBearingValuesAlwaysRedactAndRejectJSON(t *testing.T) {
	canary := "TASK16-LOCATOR-TOKEN-CANARY"
	issueCommand := &IssueCommand{AccessToken: secret.NewBytes([]byte(canary)), IdempotencyKey: canary}
	defer issueCommand.AccessToken.Clear()
	issued := &IssuedBundle{Locator: secret.NewBytes([]byte(canary))}
	defer issued.Locator.Clear()
	resolveQuery := &ResolveQuery{AccessToken: secret.NewBytes([]byte(canary)), IdempotencyKey: canary}
	defer resolveQuery.AccessToken.Clear()
	acknowledge := &AcknowledgeCommand{AccessToken: secret.NewBytes([]byte(canary)), IdempotencyKey: canary}
	defer acknowledge.AccessToken.Clear()
	issuance := &bundleIssuance{SignerKeyID: canary, Envelope: []byte(canary)}
	defer issuance.clear()
	stored := &storedBundle{Locator: secret.NewBytes([]byte(canary))}
	defer stored.Locator.Clear()
	idempotencyRecord := &transactionIdempotency{response: []byte(canary)}
	defer clear(idempotencyRecord.response)
	immutable := &ImmutableBundle{Envelope: []byte(canary)}
	defer immutable.Clear()
	values := []any{
		issueCommand,
		issued,
		resolveQuery,
		Resolution{Locator: canary, Locations: [3]string{canary}},
		acknowledge,
		&issuedReplayV1{Locator: canary, EnvelopeSHA256: canary},
		&resolutionReplayV1{Locator: canary, Locations: [3]string{canary}},
		issuance,
		stored,
		idempotencyRecord,
		immutable,
	}
	for _, value := range values {
		name := fmt.Sprintf("%T", value)
		if rendered := fmt.Sprintf("%+v", value); bytes.Contains([]byte(rendered), []byte(canary)) {
			t.Fatalf("%s fmt leaked canary", name)
		}
		if rendered := slog.Any("value", value).Value.Resolve().String(); bytes.Contains([]byte(rendered), []byte(canary)) {
			t.Fatalf("%s slog leaked canary", name)
		}
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%s generic JSON succeeded", name)
		}
	}
}

type task16ApplicationFixture struct {
	application     *Service
	repository      *task16Repository
	participant     *task16Participant
	signer          *TimeoutConfigSigner
	sequence        *task16Sequence
	trace           *task16Trace
	authorizationID uuid.UUID
}

func newTask16ApplicationFixture(t *testing.T) *task16ApplicationFixture {
	t.Helper()
	now := time.Date(2026, time.August, 13, 8, 0, 0, 0, time.UTC)
	trace := &task16Trace{}
	repository := &task16Repository{transaction: newTask16Transaction(trace)}
	authorizationID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	participant := &task16Participant{trace: trace, authority: task16Authority(t, authorizationID)}
	local, err := NewLocalConfigSigner(secret.NewBytes(bytes.Repeat([]byte{0x31}, 32)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = local.Close() })
	traced := &task16TracingSigner{inner: local, trace: trace}
	signer, err := NewTimeoutConfigSigner(traced, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = signer.Close() })
	sequence := &task16Sequence{current: TestConfigV1{Message: "hello", Sequence: "1"}, trace: trace}
	application, err := NewApplication(ApplicationDependencies{
		Repository:      repository,
		Device:          participant,
		Signer:          signer,
		Metadata:        task16Metadata(local, now),
		Random:          &task16Random{trace: trace, next: 1},
		Clock:           task16Clock{now: now},
		TestConfig:      sequence,
		BundleBaseURLs:  [3]string{"https://primary.example", "https://mirror-a.example", "https://mirror-b.example"},
		RequestDeadline: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}
	return &task16ApplicationFixture{
		application: application, repository: repository, participant: participant,
		signer: signer, sequence: sequence, trace: trace, authorizationID: authorizationID,
	}
}

func (fixture *task16ApplicationFixture) issueCommand(key, sequence, message string) IssueCommand {
	return IssueCommand{
		AuthorizationID: fixture.authorizationID,
		AccessToken:     secret.NewBytes([]byte("device-access-token-issue")),
		TestConfig:      TestConfigV1{Message: message, Sequence: sequence},
		IdempotencyKey:  key,
	}
}

func (fixture *task16ApplicationFixture) resolveQuery(key string) ResolveQuery {
	return ResolveQuery{
		AuthorizationID: fixture.authorizationID,
		AccessToken:     secret.NewBytes([]byte("device-access-token-resolve")),
		IdempotencyKey:  key,
	}
}

func task16Authority(t *testing.T, authorizationID uuid.UUID) deviceauth.BundleAuthority {
	t.Helper()
	private, err := ecdh.X25519().NewPrivateKey(bytes.Repeat([]byte{0x42}, 32))
	if err != nil {
		t.Fatal(err)
	}
	var hpke [32]byte
	copy(hpke[:], private.PublicKey().Bytes())
	return deviceauth.BundleAuthority{
		AuthorizationID:  authorizationID,
		PrincipalID:      uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"),
		DeviceID:         uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc"),
		HPKEPublicKey:    hpke,
		DeviceKeyVersion: 1,
		PolicySchema:     "device-policy-v1",
		Policy:           []byte(`{"mode":"standard"}`),
	}
}

func task16Metadata(local *LocalConfigSigner, now time.Time) RootMetadataV1 {
	return RootMetadataV1{
		SchemaVersion: TrustMetadataSchemaV1,
		Version:       "1",
		RootKeyID:     "abcdefghijklmnop",
		RootAlgorithm: SignatureAlgorithm,
		ValidFrom:     now.Add(-time.Hour).Format(time.RFC3339),
		ValidUntil:    now.Add(72 * time.Hour).Format(time.RFC3339),
		SigningKeys: []SigningKeyMetadataV1{{
			KeyID: local.KeyID(), Algorithm: SignatureAlgorithm,
			PublicKey: base64.RawURLEncoding.EncodeToString(local.PublicKey()), State: "active",
			NotBefore: now.Add(-time.Hour).Format(time.RFC3339), NotAfter: now.Add(48 * time.Hour).Format(time.RFC3339),
		}},
	}
}

type task16Trace struct {
	mu     sync.Mutex
	values []string
}

func (trace *task16Trace) add(value string) {
	trace.mu.Lock()
	defer trace.mu.Unlock()
	trace.values = append(trace.values, value)
}

func (trace *task16Trace) snapshot() []string {
	trace.mu.Lock()
	defer trace.mu.Unlock()
	return append([]string(nil), trace.values...)
}

func (trace *task16Trace) reset() {
	trace.mu.Lock()
	defer trace.mu.Unlock()
	trace.values = nil
}

type task16Random struct {
	trace *task16Trace
	next  byte
}

func (random *task16Random) Read(target []byte) (int, error) {
	label := fmt.Sprintf("random:%d", len(target))
	if len(target) != 16 && len(target) != 32 {
		label = "random:padding"
	}
	random.trace.add(label)
	for index := range target {
		target[index] = random.next
		random.next++
	}
	return len(target), nil
}

type task16Clock struct{ now time.Time }

func (clock task16Clock) Now() time.Time { return clock.now }

type task16Sequence struct {
	mu      sync.Mutex
	current TestConfigV1
	trace   *task16Trace
}

func (sequence *task16Sequence) Current(context.Context) (TestConfigV1, error) {
	sequence.mu.Lock()
	defer sequence.mu.Unlock()
	sequence.trace.add("sequence")
	return sequence.current, nil
}

func (sequence *task16Sequence) set(value TestConfigV1) {
	sequence.mu.Lock()
	defer sequence.mu.Unlock()
	sequence.current = value
}

type task16TracingSigner struct {
	inner *LocalConfigSigner
	trace *task16Trace
}

func (signer *task16TracingSigner) KeyID() string { return signer.inner.KeyID() }

func (signer *task16TracingSigner) Sign(ctx context.Context, message []byte) ([]byte, error) {
	signer.trace.add("sign")
	return signer.inner.Sign(ctx, message)
}

type task16Participant struct {
	trace             *task16Trace
	authority         deviceauth.BundleAuthority
	err               error
	dbtx              store.DBTX
	lastAuthorization uuid.UUID
}

func (participant *task16Participant) AuthorizeBundleInTransaction(
	_ context.Context,
	dbtx store.DBTX,
	_ deviceauth.AuthorizeBundleQuery,
) (deviceauth.BundleAuthority, error) {
	participant.trace.add("authorize")
	participant.dbtx = dbtx
	participant.lastAuthorization = participant.authority.AuthorizationID
	if participant.err != nil {
		return deviceauth.BundleAuthority{}, participant.err
	}
	authority := participant.authority
	authority.Policy = bytes.Clone(participant.authority.Policy)
	return authority, nil
}

type task16Repository struct {
	mu          sync.Mutex
	transaction *task16Transaction
}

func (repository *task16Repository) WithinTransaction(ctx context.Context, operation func(context.Context, Transaction) error) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	snapshot := repository.transaction.snapshot()
	err := operation(ctx, repository.transaction)
	if err != nil {
		repository.transaction.restore(snapshot)
		repository.transaction.trace.add("rollback")
		return err
	}
	repository.transaction.trace.add("commit")
	return nil
}

type task16Transaction struct {
	mu               sync.Mutex
	trace            *task16Trace
	highest          uint64
	bundles          []bundleIssuance
	idempotency      map[string]task16IdempotencyRecord
	acknowledgements map[uuid.UUID]struct{}
	events           []*eventsv1.EventEnvelope
}

type task16IdempotencyRecord struct {
	digest [32]byte
	status int
	body   []byte
}

type task16TransactionSnapshot struct {
	highest          uint64
	bundles          []bundleIssuance
	idempotency      map[string]task16IdempotencyRecord
	acknowledgements map[uuid.UUID]struct{}
	events           []*eventsv1.EventEnvelope
}

func newTask16Transaction(trace *task16Trace) *task16Transaction {
	return &task16Transaction{trace: trace, idempotency: make(map[string]task16IdempotencyRecord), acknowledgements: make(map[uuid.UUID]struct{})}
}

func (transaction *task16Transaction) DBTX() store.DBTX { return transaction }

func (transaction *task16Transaction) BeginIdempotency(
	_ context.Context, scope idempotency.Scope, key string, canonical []byte, _, _ time.Time,
) (transactionIdempotency, error) {
	transaction.trace.add("begin:" + scope.Operation)
	digest := sha256.Sum256(canonical)
	mapKey := scope.Operation + ":" + key
	record, exists := transaction.idempotency[mapKey]
	if !exists {
		return transactionIdempotency{outcome: idempotency.Started, key: mapKey, digest: digest}, nil
	}
	if record.digest != digest {
		return transactionIdempotency{outcome: idempotency.Conflict, key: mapKey, digest: digest}, nil
	}
	return transactionIdempotency{outcome: idempotency.Replay, key: mapKey, digest: digest, status: record.status, response: bytes.Clone(record.body)}, nil
}

func (transaction *task16Transaction) CompleteIdempotency(_ context.Context, value transactionIdempotency, status int, body []byte) error {
	transaction.trace.add(fmt.Sprintf("complete:%d", status))
	transaction.idempotency[value.key] = task16IdempotencyRecord{digest: value.digest, status: status, body: bytes.Clone(body)}
	return nil
}

func (transaction *task16Transaction) NextBundleVersion(context.Context, uuid.UUID, time.Time) (uint64, error) {
	transaction.trace.add("version")
	transaction.highest++
	return transaction.highest, nil
}

func (transaction *task16Transaction) LatestBundle(_ context.Context, authorizationID uuid.UUID) (storedBundle, bool, error) {
	transaction.trace.add("latest")
	for index := len(transaction.bundles) - 1; index >= 0; index-- {
		bundle := transaction.bundles[index]
		if bundle.AuthorizationID == authorizationID {
			return storedBundleFromIssuance(bundle), true, nil
		}
	}
	return storedBundle{}, false, nil
}

func (transaction *task16Transaction) BundleByID(_ context.Context, bundleID uuid.UUID) (storedBundle, bool, error) {
	for _, bundle := range transaction.bundles {
		if bundle.ID == bundleID {
			return storedBundleFromIssuance(bundle), true, nil
		}
	}
	return storedBundle{}, false, nil
}

func (transaction *task16Transaction) InsertBundle(_ context.Context, value bundleIssuance) error {
	transaction.trace.add("insert")
	copyValue := value.clone()
	transaction.bundles = append(transaction.bundles, copyValue)
	return nil
}

func (transaction *task16Transaction) InsertAcknowledgement(_ context.Context, bundle storedBundle, _ time.Time) error {
	transaction.acknowledgements[bundle.ID] = struct{}{}
	return nil
}

func (transaction *task16Transaction) AppendEvent(_ context.Context, event *eventsv1.EventEnvelope) error {
	transaction.trace.add("event")
	transaction.events = append(transaction.events, event)
	return nil
}

func (transaction *task16Transaction) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}

func (transaction *task16Transaction) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}

func (transaction *task16Transaction) QueryRow(context.Context, string, ...any) pgx.Row {
	return task16ErrorRow{}
}

type task16ErrorRow struct{}

func (task16ErrorRow) Scan(...any) error { return errors.New("unexpected QueryRow") }

func (transaction *task16Transaction) snapshot() task16TransactionSnapshot {
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	result := task16TransactionSnapshot{
		highest:          transaction.highest,
		idempotency:      make(map[string]task16IdempotencyRecord, len(transaction.idempotency)),
		acknowledgements: make(map[uuid.UUID]struct{}, len(transaction.acknowledgements)),
		events:           append([]*eventsv1.EventEnvelope(nil), transaction.events...),
	}
	for _, bundle := range transaction.bundles {
		result.bundles = append(result.bundles, bundle.clone())
	}
	for key, record := range transaction.idempotency {
		record.body = bytes.Clone(record.body)
		result.idempotency[key] = record
	}
	for key := range transaction.acknowledgements {
		result.acknowledgements[key] = struct{}{}
	}
	return result
}

func (transaction *task16Transaction) restore(snapshot task16TransactionSnapshot) {
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	transaction.highest = snapshot.highest
	transaction.bundles = snapshot.bundles
	transaction.idempotency = snapshot.idempotency
	transaction.acknowledgements = snapshot.acknowledgements
	transaction.events = snapshot.events
}

func storedBundleFromIssuance(value bundleIssuance) storedBundle {
	return storedBundle{
		ID: value.ID, AuthorizationID: value.AuthorizationID, BundleVersion: value.BundleVersion,
		Locator: secret.NewBytes(value.Locator[:]), EnvelopeSHA256: value.EnvelopeSHA256,
		IssuedAt: value.IssuedAt, NotBefore: value.NotBefore, ExpiresAt: value.ExpiresAt,
	}
}

func containsTask16Sequence(values, sequence []string) bool {
	position := 0
	for _, value := range values {
		if position < len(sequence) && value == sequence[position] {
			position++
		}
	}
	return position == len(sequence)
}

func task16PublicCode(err error) apierrors.Code {
	var classified apierrors.Error
	if !errors.As(err, &classified) {
		return ""
	}
	return apierrors.Code(classified.Public("trace-safe-task16").Code)
}
