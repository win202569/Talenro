package idempotency

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

func TestRequestDigestUsesExactDomainAndBoundedCanonicalBytes(t *testing.T) {
	t.Parallel()

	canonical := []byte(`{"email":"member@example.test"}`)
	want := sha256.Sum256(append([]byte("TALENRO-IDEMPOTENCY-REQUEST-V1\x00register_account\x00"), canonical...))
	got, err := RequestDigest("register_account", canonical)
	if err != nil {
		t.Fatalf("RequestDigest: %v", err)
	}
	if got != want {
		t.Fatalf("digest = %x, want %x", got, want)
	}
	canonical[0] = 'X'
	if got != want {
		t.Fatal("digest changed after caller mutated canonical bytes")
	}

	for _, invalid := range [][]byte{nil, make([]byte, 64*1024+1)} {
		if _, err := RequestDigest("register_account", invalid); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("canonical length %d error = %v", len(invalid), err)
		}
	}
}

func TestKeyDigestUsesDistinctLengthFramedDomainAndExactBounds(t *testing.T) {
	t.Parallel()

	key := "abcdefghijklmnopqrstuv"
	got, err := KeyDigest(key)
	if err != nil {
		t.Fatal(err)
	}
	wantMaterial := append([]byte("TALENRO-IDEMPOTENCY-KEY-V1\x00\x00\x16"), []byte(key)...)
	want := sha256.Sum256(wantMaterial)
	if got != want {
		t.Fatalf("digest = %x, want %x", got, want)
	}
	request, err := RequestDigest("register_account", []byte(key))
	if err != nil {
		t.Fatal(err)
	}
	if got == request {
		t.Fatal("key and request domains produced the same digest")
	}
	for _, length := range []int{22, 86} {
		if _, err := KeyDigest(string(bytes.Repeat([]byte{'a'}, length))); err != nil {
			t.Fatalf("valid length %d: %v", length, err)
		}
	}
	for _, invalid := range []string{
		string(bytes.Repeat([]byte{'a'}, 21)),
		string(bytes.Repeat([]byte{'a'}, 87)),
		"abcdefghijklmnopqrstu=",
	} {
		if _, err := KeyDigest(invalid); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("invalid length %d error = %v", len(invalid), err)
		}
	}
}

func TestBeginStartsWithDigestsOnlyAndDefensiveCopies(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	storeFake := &fakeIdempotencyStore{tryRows: 1}
	repository := newWithStore(storeFake, newProtector(t))
	canonical := []byte(`{"value":1}`)
	record, outcome, err := repository.Begin(context.Background(), AnonymousRegistrationScope(), "abcdefghijklmnopqrstuv", canonical, createdAt, createdAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if outcome != Started {
		t.Fatalf("outcome = %q, want %q", outcome, Started)
	}
	if storeFake.tryCalls != 1 || storeFake.getCalls != 0 {
		t.Fatalf("calls try=%d get=%d", storeFake.tryCalls, storeFake.getCalls)
	}
	if bytes.Contains(storeFake.tryParams.IdempotencyKeyHash, []byte("abcdefghijklmnopqrstuv")) ||
		bytes.Contains(storeFake.tryParams.RequestDigest, canonical) {
		t.Fatal("store parameters retained plaintext key or canonical request")
	}
	if len(storeFake.tryParams.IdempotencyKeyHash) != 32 || len(storeFake.tryParams.RequestDigest) != 32 {
		t.Fatalf("digest lengths = %d/%d", len(storeFake.tryParams.IdempotencyKeyHash), len(storeFake.tryParams.RequestDigest))
	}
	canonical[0] = 'X'
	storeFake.tryParams.RequestDigest[0] ^= 0xff
	recordDigest := record.RequestDigest()
	if bytes.Equal(recordDigest[:], storeFake.tryParams.RequestDigest) {
		t.Fatal("record aliases store request digest")
	}
	if rendered := fmt.Sprintf("%+v", record); bytes.Contains([]byte(rendered), storeFake.tryParams.IdempotencyKeyHash) {
		t.Fatal("record formatting exposed key digest")
	}
}

func TestOnlyFixedAnonymousDeliveryScopesAreAccepted(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	tests := []struct {
		name      string
		scope     Scope
		principal string
		operation string
	}{
		{
			name:      "email verification delivery",
			scope:     AnonymousEmailVerificationDeliveryScope(),
			principal: AnonymousEmailVerificationDeliveryPrincipal,
			operation: CreateEmailVerificationDeliveryOperation,
		},
		{
			name:      "password reset delivery",
			scope:     AnonymousPasswordResetDeliveryScope(),
			principal: AnonymousPasswordResetDeliveryPrincipal,
			operation: CreatePasswordResetDeliveryOperation,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.scope != (Scope{Principal: test.principal, Operation: test.operation}) {
				t.Fatalf("scope = %#v, want fixed principal and operation", test.scope)
			}
			storeFake := &fakeIdempotencyStore{tryRows: 1}
			repository := newWithStore(storeFake, newProtector(t))
			_, outcome, err := repository.Begin(
				context.Background(), test.scope, "abcdefghijklmnopqrstuv", []byte(`{"email":"opaque"}`), now, now.Add(time.Hour),
			)
			if err != nil || outcome != Started || storeFake.tryCalls != 1 {
				t.Fatalf("fixed scope begin = (%q, %v), store calls %d", outcome, err, storeFake.tryCalls)
			}
		})
	}

	for _, arbitrary := range []Scope{
		{Principal: "anonymous_arbitrary", Operation: CreateEmailVerificationDeliveryOperation},
		{Principal: AnonymousEmailVerificationDeliveryPrincipal, Operation: "arbitrary_operation"},
		{Principal: AnonymousPasswordResetDeliveryPrincipal, Operation: CreateEmailVerificationDeliveryOperation},
	} {
		storeFake := &fakeIdempotencyStore{tryRows: 1}
		repository := newWithStore(storeFake, newProtector(t))
		if _, _, err := repository.Begin(context.Background(), arbitrary, "abcdefghijklmnopqrstuv", []byte(`{}`), now, now.Add(time.Hour)); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("arbitrary anonymous scope %#v error = %v", arbitrary, err)
		}
		if storeFake.tryCalls != 0 {
			t.Fatalf("arbitrary anonymous scope reached store: %#v", arbitrary)
		}
	}
}

func TestBeginClassifiesExistingRecord(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	scope, err := AuthenticatedScope(uuid.MustParse("4b4d278b-9e7a-4ce0-865d-1dc14fcf96da"), "password", "rotate_session")
	if err != nil {
		t.Fatal(err)
	}
	key := "abcdefghijklmnopqrstuv"
	canonical := []byte(`{"session":"fixed"}`)
	digest, err := RequestDigest(scope.Operation, canonical)
	if err != nil {
		t.Fatal(err)
	}
	keyHash, err := KeyDigest(key)
	if err != nil {
		t.Fatal(err)
	}
	protector := newProtector(t)
	protected, err := protector.Encrypt(responseProtectionDomain, append([]byte{1}, []byte(`{"accepted":true}`)...))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		mutate       func(*store.IdempotencyRecord)
		wantOutcome  Outcome
		wantErr      error
		wantResponse []byte
		wantDecrypts int
	}{
		{name: "completed matching digest replays", wantOutcome: Replay, wantResponse: []byte(`{"accepted":true}`), wantDecrypts: 1},
		{name: "different request conflicts without decrypt", mutate: func(row *store.IdempotencyRecord) { row.RequestDigest[0] ^= 1 }, wantOutcome: Conflict},
		{name: "matching in progress does not poll", mutate: func(row *store.IdempotencyRecord) { row.State = "in_progress" }, wantOutcome: InProgress},
		{name: "failed record fails closed", mutate: func(row *store.IdempotencyRecord) { row.State = "failed" }, wantErr: ErrRecordUnavailable},
		{name: "expired record fails closed", mutate: func(row *store.IdempotencyRecord) { row.ExpiresAt = createdAt }, wantErr: ErrRecordUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			row := store.IdempotencyRecord{
				PrincipalScope: scope.Principal, Operation: scope.Operation,
				IdempotencyKeyHash: append([]byte(nil), keyHash[:]...), RequestDigest: append([]byte(nil), digest[:]...), State: "completed",
				ResponseStatus: pgtype.Int4{Int32: 202, Valid: true}, ResponseCiphertext: protected.Ciphertext,
				ResponseKeyVersion: pgtype.Int4{Int32: math.MaxInt32, Valid: true},
				CreatedAt:          createdAt.Add(-time.Minute), ExpiresAt: createdAt.Add(time.Hour),
			}
			if test.mutate != nil {
				test.mutate(&row)
			}
			counting := &countingProtector{Protector: protector}
			storeFake := &fakeIdempotencyStore{tryRows: 0, getRecord: row}
			repository := newWithStore(storeFake, counting)
			record, outcome, beginErr := repository.Begin(context.Background(), scope, key, canonical, createdAt, createdAt.Add(time.Hour))
			if !errors.Is(beginErr, test.wantErr) {
				t.Fatalf("error = %v, want %v", beginErr, test.wantErr)
			}
			if outcome != test.wantOutcome {
				t.Fatalf("outcome = %q, want %q", outcome, test.wantOutcome)
			}
			if counting.decrypts != test.wantDecrypts {
				t.Fatalf("decrypt calls = %d, want %d", counting.decrypts, test.wantDecrypts)
			}
			response, owned := record.TakeResponseBody()
			wantOwned := test.wantOutcome == Replay
			if owned != wantOwned || !bytes.Equal(response, test.wantResponse) {
				t.Fatalf("response = %q/%v, want %q/%v", response, owned, test.wantResponse, wantOwned)
			}
			clear(response)
			if second, secondOwned := record.TakeResponseBody(); secondOwned || second != nil {
				t.Fatalf("response transferred twice = %q/%v", second, secondOwned)
			}
			if storeFake.tryCalls != 1 || storeFake.getCalls != 1 {
				t.Fatalf("calls try=%d get=%d", storeFake.tryCalls, storeFake.getCalls)
			}
		})
	}
}

func TestOnlyNewlyStartedRecordCanComplete(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	scope, err := AuthenticatedScope(uuid.MustParse("4b4d278b-9e7a-4ce0-865d-1dc14fcf96da"), "password", "rotate_session")
	if err != nil {
		t.Fatal(err)
	}
	key := "abcdefghijklmnopqrstuv"
	canonical := []byte(`{"session":"fixed"}`)
	digest, _ := RequestDigest(scope.Operation, canonical)
	keyHash, _ := KeyDigest(key)
	baseProtector := newProtector(t)
	protected, err := baseProtector.Encrypt(responseProtectionDomain, append([]byte{1}, []byte(`{"accepted":true}`)...))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		mutate func(*store.IdempotencyRecord)
	}{
		{name: "replay"},
		{name: "conflict", mutate: func(row *store.IdempotencyRecord) { row.RequestDigest[0] ^= 1 }},
		{name: "in progress", mutate: func(row *store.IdempotencyRecord) { row.State = "in_progress" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			row := store.IdempotencyRecord{
				PrincipalScope: scope.Principal, Operation: scope.Operation,
				IdempotencyKeyHash: append([]byte(nil), keyHash[:]...), RequestDigest: append([]byte(nil), digest[:]...), State: "completed",
				ResponseStatus: pgtype.Int4{Int32: 200, Valid: true}, ResponseCiphertext: bytes.Clone(protected.Ciphertext),
				ResponseKeyVersion: pgtype.Int4{Int32: math.MaxInt32, Valid: true}, CreatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour),
			}
			if test.mutate != nil {
				test.mutate(&row)
			}
			storeFake := &fakeIdempotencyStore{getRecord: row}
			protector := &countingProtector{Protector: baseProtector}
			repository := newWithStore(storeFake, protector)
			record, _, beginErr := repository.Begin(context.Background(), scope, key, canonical, now, now.Add(time.Hour))
			if beginErr != nil {
				t.Fatal(beginErr)
			}
			encryptsBefore := protector.encrypts
			if _, completeErr := repository.Complete(context.Background(), record, 200, []byte("response")); !errors.Is(completeErr, ErrConflict) {
				t.Fatalf("Complete error = %v, want ErrConflict", completeErr)
			}
			if protector.encrypts != encryptsBefore || storeFake.completeCalls != 0 {
				t.Fatalf("non-owner completion encrypted/stored: encrypts=%d calls=%d", protector.encrypts-encryptsBefore, storeFake.completeCalls)
			}
		})
	}
}

func TestReplayResponseOwnershipTransfersExactlyOnceAcrossRecordCopies(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	scope := AnonymousRegistrationScope()
	key := "abcdefghijklmnopqrstuv"
	canonical := []byte(`{"request":"fixed"}`)
	digest, err := RequestDigest(scope.Operation, canonical)
	if err != nil {
		t.Fatal(err)
	}
	keyHash, err := KeyDigest(key)
	if err != nil {
		t.Fatal(err)
	}
	protector := newProtector(t)
	want := []byte("SECRET_RESPONSE_CANARY")
	protected, err := protector.Encrypt(responseProtectionDomain, append([]byte{responseFrameVersion}, want...))
	if err != nil {
		t.Fatal(err)
	}
	repository := newWithStore(&fakeIdempotencyStore{getRecord: store.IdempotencyRecord{
		PrincipalScope: scope.Principal, Operation: scope.Operation,
		IdempotencyKeyHash: bytes.Clone(keyHash[:]), RequestDigest: bytes.Clone(digest[:]), State: "completed",
		ResponseStatus: pgtype.Int4{Int32: 200, Valid: true}, ResponseCiphertext: bytes.Clone(protected.Ciphertext),
		ResponseKeyVersion: pgtype.Int4{Int32: math.MaxInt32, Valid: true}, CreatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour),
	}}, protector)
	record, outcome, err := repository.Begin(context.Background(), scope, key, canonical, now, now.Add(time.Hour))
	if err != nil || outcome != Replay {
		t.Fatalf("Begin = %q, %v", outcome, err)
	}
	copyOfRecord := record
	body, ok := record.TakeResponseBody()
	if !ok || !bytes.Equal(body, want) {
		t.Fatalf("first transfer = %q, %v", body, ok)
	}
	if second, secondOK := copyOfRecord.TakeResponseBody(); secondOK || second != nil {
		t.Fatalf("second transfer through copied record = %q, %v", second, secondOK)
	}
	clear(body)
	if third, thirdOK := record.TakeResponseBody(); thirdOK || third != nil {
		t.Fatalf("third transfer after caller clear = %q, %v", third, thirdOK)
	}
	if rendered := fmt.Sprintf("%+v", copyOfRecord); rendered != "idempotency.Record([REDACTED])" || bytes.Contains([]byte(rendered), want) {
		t.Fatalf("record formatting exposed response ownership: %q", rendered)
	}
}

func TestBeginRejectsInvalidKeyAndScopeBeforeStore(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	tests := []struct {
		name  string
		scope Scope
		key   string
	}{
		{name: "short key", scope: AnonymousRegistrationScope(), key: "short"},
		{name: "padded key", scope: AnonymousRegistrationScope(), key: "abcdefghijklmnopqrstu="},
		{name: "ambiguous anonymous operation", scope: Scope{Principal: AnonymousRegistrationPrincipal, Operation: "rotate_session"}, key: "abcdefghijklmnopqrstuv"},
		{name: "unstructured authenticated principal", scope: Scope{Principal: uuid.NewString(), Operation: "rotate_session"}, key: "abcdefghijklmnopqrstuv"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storeFake := &fakeIdempotencyStore{}
			repository := newWithStore(storeFake, newProtector(t))
			_, _, err := repository.Begin(context.Background(), test.scope, test.key, []byte(`{}`), now, now.Add(time.Hour))
			if !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("error = %v, want ErrInvalidArgument", err)
			}
			if storeFake.tryCalls != 0 || storeFake.getCalls != 0 {
				t.Fatal("invalid request reached store")
			}
		})
	}
}

func TestCompleteEncryptsBoundedResponseAndHandlesLostRace(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	storeFake := &fakeIdempotencyStore{tryRows: 1}
	repository := newWithStore(storeFake, newProtector(t))
	record, _, err := repository.Begin(context.Background(), AnonymousRegistrationScope(), "abcdefghijklmnopqrstuv", []byte(`{}`), now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"accepted":true}`)
	storeFake.completeRecord = store.IdempotencyRecord{
		PrincipalScope:     storeFake.tryParams.PrincipalScope,
		Operation:          storeFake.tryParams.Operation,
		IdempotencyKeyHash: bytes.Clone(storeFake.tryParams.IdempotencyKeyHash),
		RequestDigest:      bytes.Clone(storeFake.tryParams.RequestDigest),
		State:              "completed",
		CreatedAt:          storeFake.tryParams.CreatedAt,
		ExpiresAt:          storeFake.tryParams.ExpiresAt,
	}
	completed, err := repository.Complete(context.Background(), record, 202, body)
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	body[0] = 'X'
	completedBody, owned := completed.TakeResponseBody()
	defer clear(completedBody)
	if bytes.Equal(storeFake.completeParams.ResponseCiphertext, body) || !owned || !bytes.Equal(completedBody, []byte(`{"accepted":true}`)) {
		t.Fatal("completion retained plaintext or aliased caller response")
	}
	if !storeFake.completeParams.ResponseKeyVersion.Valid || storeFake.completeParams.ResponseKeyVersion.Int32 <= 0 || len(storeFake.completeParams.ResponseCiphertext) > 1_048_608 {
		t.Fatalf("invalid protected response metadata: %+v", storeFake.completeParams.ResponseKeyVersion)
	}

	for _, test := range []struct {
		name   string
		status int
		body   []byte
	}{
		{name: "negative status", status: -1, body: []byte("x")},
		{name: "non http status", status: 99, body: []byte("x")},
		{name: "oversize body", status: 200, body: make([]byte, 1<<20+1)},
	} {
		t.Run(test.name, func(t *testing.T) {
			before := storeFake.completeCalls
			if _, err := repository.Complete(context.Background(), record, test.status, test.body); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("error = %v", err)
			}
			if storeFake.completeCalls != before {
				t.Fatal("invalid completion reached store")
			}
		})
	}

	storeFake.completeErr = pgx.ErrNoRows
	if _, err := repository.Complete(context.Background(), record, 200, []byte("x")); !errors.Is(err, ErrConflict) {
		t.Fatalf("lost race error = %v, want ErrConflict", err)
	}
}

func TestCompleteAndReplayPreserveEmpty204BodyWithVersionedFrame(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	protector := newProtector(t)
	storeFake := &fakeIdempotencyStore{tryRows: 1}
	repository := newWithStore(storeFake, protector)
	record, _, err := repository.Begin(context.Background(), AnonymousRegistrationScope(), "abcdefghijklmnopqrstuv", []byte(`{}`), now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	storeFake.completeRecord = store.IdempotencyRecord{
		PrincipalScope: storeFake.tryParams.PrincipalScope, Operation: storeFake.tryParams.Operation,
		IdempotencyKeyHash: bytes.Clone(storeFake.tryParams.IdempotencyKeyHash), RequestDigest: bytes.Clone(storeFake.tryParams.RequestDigest),
		State: "completed", CreatedAt: now, ExpiresAt: now.Add(time.Hour),
	}
	completed, err := repository.Complete(context.Background(), record, 204, nil)
	if err != nil {
		t.Fatalf("Complete empty 204: %v", err)
	}
	completedBody, completedOwned := completed.TakeResponseBody()
	if completed.ResponseStatus() != 204 || !completedOwned || len(completedBody) != 0 {
		t.Fatalf("completion status/body = %d/%q/%v", completed.ResponseStatus(), completedBody, completedOwned)
	}
	opened, err := protector.Decrypt(responseProtectionDomain, sensitive.EncryptedField{
		KeyVersion: uint32(math.MaxInt32),
		Ciphertext: bytes.Clone(storeFake.completeParams.ResponseCiphertext),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(opened, []byte{1}) {
		t.Fatalf("encrypted frame = %v, want [1]", opened)
	}

	replayStore := &fakeIdempotencyStore{getRecord: store.IdempotencyRecord{
		PrincipalScope: storeFake.tryParams.PrincipalScope, Operation: storeFake.tryParams.Operation,
		IdempotencyKeyHash: bytes.Clone(storeFake.tryParams.IdempotencyKeyHash), RequestDigest: bytes.Clone(storeFake.tryParams.RequestDigest),
		State: "completed", ResponseStatus: pgtype.Int4{Int32: 204, Valid: true},
		ResponseCiphertext: bytes.Clone(storeFake.completeParams.ResponseCiphertext), ResponseKeyVersion: storeFake.completeParams.ResponseKeyVersion,
		CreatedAt: now, ExpiresAt: now.Add(time.Hour),
	}}
	replayed, outcome, err := newWithStore(replayStore, protector).Begin(context.Background(), AnonymousRegistrationScope(), "abcdefghijklmnopqrstuv", []byte(`{}`), now.Add(time.Minute), now.Add(time.Hour))
	replayedBody, replayedOwned := replayed.TakeResponseBody()
	if err != nil || outcome != Replay || replayed.ResponseStatus() != 204 || !replayedOwned || len(replayedBody) != 0 {
		t.Fatalf("replay = outcome %q status %d body %q/%v error %v", outcome, replayed.ResponseStatus(), replayedBody, replayedOwned, err)
	}
}

func TestCompleteAcceptsExactOneMiBBodyWithinDatabaseCiphertextBound(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	storeFake := &fakeIdempotencyStore{tryRows: 1}
	repository := newWithStore(storeFake, newProtector(t))
	record, _, err := repository.Begin(context.Background(), AnonymousRegistrationScope(), "abcdefghijklmnopqrstuv", []byte(`{}`), now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	storeFake.completeRecord = store.IdempotencyRecord{
		PrincipalScope: storeFake.tryParams.PrincipalScope, Operation: storeFake.tryParams.Operation,
		IdempotencyKeyHash: bytes.Clone(storeFake.tryParams.IdempotencyKeyHash), RequestDigest: bytes.Clone(storeFake.tryParams.RequestDigest),
		State: "completed", CreatedAt: now, ExpiresAt: now.Add(time.Hour),
	}
	body := bytes.Repeat([]byte{'x'}, 1<<20)
	completed, err := repository.Complete(context.Background(), record, 200, body)
	if err != nil {
		t.Fatalf("Complete exact 1 MiB: %v", err)
	}
	completedBody, owned := completed.TakeResponseBody()
	defer clear(completedBody)
	if !owned || len(completedBody) != len(body) || len(storeFake.completeParams.ResponseCiphertext) > 1_048_608 {
		t.Fatalf("body/ciphertext lengths = %d/%d", len(completedBody), len(storeFake.completeParams.ResponseCiphertext))
	}
}

func TestIdempotencyRejectsTypedNilDependenciesWithoutCallingThem(t *testing.T) {
	t.Parallel()

	if _, err := New(nil, nil); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("nil dependencies error = %v", err)
	}
	var nilDB *typedNilDBTX
	if _, err := New(nilDB, newProtector(t)); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("typed-nil DBTX error = %v", err)
	}
	var nilProtector *typedNilProtector
	if _, err := New(&typedNilDBTX{}, nilProtector); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("typed-nil protector error = %v", err)
	}
	var nilStore *typedNilIdempotencyStore
	now := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	_, _, err := newWithStore(nilStore, newProtector(t)).Begin(
		context.Background(), AnonymousRegistrationScope(), "abcdefghijklmnopqrstuv", []byte(`{}`), now, now.Add(time.Hour),
	)
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("typed-nil bound store error = %v", err)
	}
	var zero repository
	_, _, err = zero.Begin(context.Background(), AnonymousRegistrationScope(), "abcdefghijklmnopqrstuv", []byte(`{}`), now, now.Add(time.Hour))
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("zero repository error = %v", err)
	}
}

func TestCompleteCancellationAfterProtectionDoesNotReachStore(t *testing.T) {
	t.Parallel()

	operationContext, cancel := context.WithCancel(context.Background())
	protector := &cancelingEncryptProtector{Protector: newProtector(t), cancel: cancel}
	now := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	storeFake := &fakeIdempotencyStore{tryRows: 1}
	repository := newWithStore(storeFake, protector)
	record, _, err := repository.Begin(operationContext, AnonymousRegistrationScope(), "abcdefghijklmnopqrstuv", []byte(`{}`), now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	_, err = repository.Complete(operationContext, record, 200, []byte("response"))
	if !errors.Is(err, ErrCanceled) {
		t.Fatalf("Complete error = %v, want ErrCanceled", err)
	}
	if storeFake.completeCalls != 0 {
		t.Fatal("canceled completion reached store")
	}
}

func TestReplayRejectsMissingOrUnknownResponseFrame(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	digest, _ := RequestDigest(AnonymousRegistrationOperation, []byte(`{}`))
	keyHash, _ := KeyDigest("abcdefghijklmnopqrstuv")
	row := store.IdempotencyRecord{
		PrincipalScope: AnonymousRegistrationPrincipal, Operation: AnonymousRegistrationOperation,
		IdempotencyKeyHash: keyHash[:], RequestDigest: digest[:], State: "completed",
		ResponseStatus: pgtype.Int4{Int32: 200, Valid: true}, ResponseCiphertext: bytes.Repeat([]byte{1}, 29),
		ResponseKeyVersion: pgtype.Int4{Int32: 1, Valid: true}, CreatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour),
	}
	for _, frame := range [][]byte{nil, {2, 'x'}} {
		protector := &fixedDecryptProtector{plaintext: frame}
		_, _, err := newWithStore(&fakeIdempotencyStore{getRecord: row}, protector).Begin(
			context.Background(), AnonymousRegistrationScope(), "abcdefghijklmnopqrstuv", []byte(`{}`), now, now.Add(time.Hour),
		)
		if !errors.Is(err, ErrProtection) {
			t.Fatalf("frame %v error = %v, want ErrProtection", frame, err)
		}
	}
}

type fakeIdempotencyStore struct {
	tryRows        int64
	tryErr         error
	tryParams      store.TryBeginIdempotencyParams
	tryCalls       int
	getRecord      store.IdempotencyRecord
	getErr         error
	getParams      store.GetIdempotencyForUpdateParams
	getCalls       int
	completeRecord store.IdempotencyRecord
	completeErr    error
	completeParams store.CompleteIdempotencyParams
	completeCalls  int
}

func (fake *fakeIdempotencyStore) TryBeginIdempotency(_ context.Context, params store.TryBeginIdempotencyParams) (int64, error) {
	fake.tryCalls++
	fake.tryParams = cloneTryParams(params)
	return fake.tryRows, fake.tryErr
}

func (fake *fakeIdempotencyStore) GetIdempotencyForUpdate(_ context.Context, params store.GetIdempotencyForUpdateParams) (store.IdempotencyRecord, error) {
	fake.getCalls++
	fake.getParams = store.GetIdempotencyForUpdateParams{PrincipalScope: params.PrincipalScope, Operation: params.Operation, IdempotencyKeyHash: bytes.Clone(params.IdempotencyKeyHash)}
	return cloneStoreRecord(fake.getRecord), fake.getErr
}

func (fake *fakeIdempotencyStore) CompleteIdempotency(_ context.Context, params store.CompleteIdempotencyParams) (store.IdempotencyRecord, error) {
	fake.completeCalls++
	fake.completeParams = cloneCompleteParams(params)
	return cloneStoreRecord(fake.completeRecord), fake.completeErr
}

func cloneTryParams(params store.TryBeginIdempotencyParams) store.TryBeginIdempotencyParams {
	params.IdempotencyKeyHash = bytes.Clone(params.IdempotencyKeyHash)
	params.RequestDigest = bytes.Clone(params.RequestDigest)
	return params
}

func cloneCompleteParams(params store.CompleteIdempotencyParams) store.CompleteIdempotencyParams {
	params.IdempotencyKeyHash = bytes.Clone(params.IdempotencyKeyHash)
	params.RequestDigest = bytes.Clone(params.RequestDigest)
	params.ResponseCiphertext = bytes.Clone(params.ResponseCiphertext)
	return params
}

func cloneStoreRecord(record store.IdempotencyRecord) store.IdempotencyRecord {
	record.IdempotencyKeyHash = bytes.Clone(record.IdempotencyKeyHash)
	record.RequestDigest = bytes.Clone(record.RequestDigest)
	record.ResponseCiphertext = bytes.Clone(record.ResponseCiphertext)
	return record
}

func newProtector(t *testing.T) *sensitive.Local {
	t.Helper()
	protector, err := sensitive.NewLocal(
		secret.NewBytes(bytes.Repeat([]byte{0x11}, 32)),
		secret.NewBytes(bytes.Repeat([]byte{0x22}, 32)),
		math.MaxInt32,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = protector.Close() })
	return protector
}

type countingProtector struct {
	sensitive.Protector
	decrypts int
	encrypts int
}

type cancelingEncryptProtector struct {
	sensitive.Protector
	cancel context.CancelFunc
}

func (protector *cancelingEncryptProtector) Encrypt(domain string, plaintext []byte) (sensitive.EncryptedField, error) {
	protector.cancel()
	return protector.Protector.Encrypt(domain, plaintext)
}

type fixedDecryptProtector struct{ plaintext []byte }

type typedNilProtector struct{}

func (*typedNilProtector) LookupDigest(string, []byte) [32]byte { panic("typed nil protector called") }
func (*typedNilProtector) Encrypt(string, []byte) (sensitive.EncryptedField, error) {
	panic("typed nil protector called")
}
func (*typedNilProtector) Decrypt(string, sensitive.EncryptedField) ([]byte, error) {
	panic("typed nil protector called")
}

type typedNilDBTX struct{}

func (*typedNilDBTX) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	panic("typed nil DBTX called")
}
func (*typedNilDBTX) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	panic("typed nil DBTX called")
}
func (*typedNilDBTX) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	panic("typed nil DBTX called")
}

type typedNilIdempotencyStore struct{}

func (*typedNilIdempotencyStore) TryBeginIdempotency(context.Context, store.TryBeginIdempotencyParams) (int64, error) {
	panic("typed nil store called")
}
func (*typedNilIdempotencyStore) GetIdempotencyForUpdate(context.Context, store.GetIdempotencyForUpdateParams) (store.IdempotencyRecord, error) {
	panic("typed nil store called")
}
func (*typedNilIdempotencyStore) CompleteIdempotency(context.Context, store.CompleteIdempotencyParams) (store.IdempotencyRecord, error) {
	panic("typed nil store called")
}

func (*fixedDecryptProtector) LookupDigest(string, []byte) [32]byte { return [32]byte{} }

func (*fixedDecryptProtector) Encrypt(string, []byte) (sensitive.EncryptedField, error) {
	return sensitive.EncryptedField{}, errors.New("unexpected encrypt")
}

func (protector *fixedDecryptProtector) Decrypt(string, sensitive.EncryptedField) ([]byte, error) {
	return bytes.Clone(protector.plaintext), nil
}

func (protector *countingProtector) Encrypt(domain string, plaintext []byte) (sensitive.EncryptedField, error) {
	protector.encrypts++
	return protector.Protector.Encrypt(domain, plaintext)
}

func (protector *countingProtector) Decrypt(domain string, value sensitive.EncryptedField) ([]byte, error) {
	protector.decrypts++
	return protector.Protector.Decrypt(domain, value)
}
