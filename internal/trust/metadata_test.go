package trust

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"talenro.local/platform/internal/secret"
)

func testMetadata(t *testing.T, version string, state string) (SignedRootMetadataV1, *LocalRootSigner, *LocalConfigSigner) {
	t.Helper()
	root, err := NewLocalRootSigner(secret.NewBytes(bytes.Repeat([]byte{0x55}, ed25519.SeedSize)))
	if err != nil {
		t.Fatalf("NewLocalRootSigner: %v", err)
	}
	config := newTestConfigSigner(t, 0x66)
	t.Cleanup(func() { _ = root.Close() })
	payload := RootMetadataV1{
		SchemaVersion: "talenro-trust-metadata/v1", Version: version,
		RootKeyID: root.KeyID(), RootAlgorithm: "Ed25519",
		ValidFrom: "2026-08-09T00:00:00Z", ValidUntil: "2026-09-09T00:00:00Z",
		SigningKeys: []SigningKeyMetadataV1{{
			KeyID: config.KeyID(), Algorithm: "Ed25519", PublicKey: encodeBase64URL(config.PublicKey()), State: state,
			NotBefore: "2026-08-09T00:00:00Z", NotAfter: "2026-08-20T00:00:00Z",
		}},
	}
	signed, err := SignRootMetadataV1(context.Background(), payload, newTestBoundedSigner(t, root))
	if err != nil {
		t.Fatalf("SignRootMetadataV1: %v", err)
	}
	return signed, root, config
}

func trustedTestRoot(root *LocalRootSigner) map[string]ed25519.PublicKey {
	return map[string]ed25519.PublicKey{root.KeyID(): root.PublicKey()}
}

// TestMetadataVerificationRequiresMonotonicTrustedRootAndAlgorithm catches
// rollback, unknown-root, algorithm substitution, validity, and root-signature bypasses.
func TestMetadataVerificationRequiresMonotonicTrustedRootAndAlgorithm(t *testing.T) {
	signed, root, _ := testMetadata(t, "2", "active")
	now := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	payload, err := VerifyRootMetadataV1(signed, trustedTestRoot(root), 1, now)
	if err != nil || payload.Version != "2" {
		t.Fatalf("VerifyRootMetadataV1 = %#v, %v", payload, err)
	}
	if _, err := VerifyRootMetadataV1(signed, trustedTestRoot(root), 2, now); !errors.Is(err, ErrMetadataRollback) {
		t.Fatalf("same version error = %v, want ErrMetadataRollback", err)
	}
	if _, err := VerifyRootMetadataV1(signed, map[string]ed25519.PublicKey{}, 1, now); !errors.Is(err, ErrUnknownRoot) {
		t.Fatalf("unknown root error = %v, want ErrUnknownRoot", err)
	}
	signature, decodeErr := base64.RawURLEncoding.DecodeString(signed.Signature)
	if decodeErr != nil || !ed25519.Verify(root.PublicKey(), append([]byte("TALENRO-TRUST-METADATA-V1\x00"), []byte(signed.PayloadJCS)...), signature) {
		t.Fatal("independent root metadata verification failed")
	}
	tampered := signed
	tampered.Signature = tamperFirstCharacter(tampered.Signature)
	if _, err := VerifyRootMetadataV1(tampered, trustedTestRoot(root), 1, now); err == nil {
		t.Fatal("accepted root-signature tamper")
	}
	var decoded RootMetadataV1
	if _, err := CanonicalizeJSON([]byte(signed.PayloadJCS), &decoded); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	unknownAlgorithmJCS := []byte(strings.Replace(signed.PayloadJCS, `"root_algorithm":"Ed25519"`, `"root_algorithm":"unknown"`, 1))
	unknownSignature, signErr := root.Sign(context.Background(), append([]byte("TALENRO-TRUST-METADATA-V1\x00"), unknownAlgorithmJCS...))
	if signErr != nil {
		t.Fatalf("independently sign unknown algorithm: %v", signErr)
	}
	unknownAlgorithm := SignedRootMetadataV1{PayloadJCS: string(unknownAlgorithmJCS), Signature: base64.RawURLEncoding.EncodeToString(unknownSignature)}
	if _, err := VerifyRootMetadataV1(unknownAlgorithm, trustedTestRoot(root), 1, now); err == nil {
		t.Fatal("accepted root-signed unknown algorithm")
	}
	payloadTamper := signed
	payloadTamper.PayloadJCS = strings.Replace(payloadTamper.PayloadJCS, `"version":"2"`, `"version":"3"`, 1)
	if _, err := VerifyRootMetadataV1(payloadTamper, trustedTestRoot(root), 1, now); err == nil {
		t.Fatal("accepted metadata payload tamper")
	}
	if _, err := VerifyRootMetadataV1(signed, trustedTestRoot(root), 1, time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("accepted expired metadata")
	}
}

// TestMetadataSigningCollapsesProviderErrors catches disclosing raw root signer
// identifiers or locators from the metadata signing boundary.
func TestMetadataSigningCollapsesProviderErrors(t *testing.T) {
	signed, _, _ := testMetadata(t, "1", "active")
	var payload RootMetadataV1
	if _, err := CanonicalizeJSON([]byte(signed.PayloadJCS), &payload); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	payload.RootKeyID = strings.Repeat("r", 22)
	bounded := newTestBoundedSigner(t, failingConfigSigner{keyID: payload.RootKeyID})
	_, err := SignRootMetadataV1(context.Background(), payload, bounded)
	if !errors.Is(err, ErrSignerFailure) || strings.Contains(fmt.Sprint(err), "SIGNER_PROVIDER_SECRET_CANARY") {
		t.Fatalf("SignRootMetadataV1 error = %v, want fixed ErrSignerFailure", err)
	}
}

// TestMetadataSigningRequiresBoundedProvider catches invoking a raw root signer
// or bypassing the configured deadline at the exported metadata boundary.
func TestMetadataSigningRequiresBoundedProvider(t *testing.T) {
	signed, root, _ := testMetadata(t, "1", "active")
	var payload RootMetadataV1
	if _, err := CanonicalizeJSON([]byte(signed.PayloadJCS), &payload); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	if _, err := SignRootMetadataV1(context.Background(), payload, root); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("raw-provider error = %v, want ErrInvalidArgument", err)
	}
	payload.RootKeyID = strings.Repeat("r", 22)
	bounded, err := NewTimeoutConfigSigner(blockingConfigSigner{keyID: payload.RootKeyID}, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("NewTimeoutConfigSigner: %v", err)
	}
	defer func() { _ = bounded.Close() }()
	if _, err := SignRootMetadataV1(context.Background(), payload, bounded); !errors.Is(err, ErrSignerTimeout) {
		t.Fatalf("bounded-provider error = %v, want ErrSignerTimeout", err)
	}
}

// TestMetadataValidationEnforcesCompleteBoundedKeySet catches duplicate,
// unordered, malformed, empty, oversized, unknown-state, and invalid-window keys.
func TestMetadataValidationEnforcesCompleteBoundedKeySet(t *testing.T) {
	signed, _, _ := testMetadata(t, "1", "active")
	var base RootMetadataV1
	if _, err := CanonicalizeJSON([]byte(signed.PayloadJCS), &base); err != nil {
		t.Fatalf("decode base metadata: %v", err)
	}
	tests := []struct {
		name   string
		mutate func(*RootMetadataV1)
	}{
		{name: "zero version", mutate: func(v *RootMetadataV1) { v.Version = "0" }},
		{name: "empty set", mutate: func(v *RootMetadataV1) { v.SigningKeys = nil }},
		{name: "too many keys", mutate: func(v *RootMetadataV1) { v.SigningKeys = validTestSigningKeys(65) }},
		{name: "duplicate key", mutate: func(v *RootMetadataV1) { v.SigningKeys = append(v.SigningKeys, v.SigningKeys[0]) }},
		{name: "unordered keys", mutate: func(v *RootMetadataV1) {
			v.SigningKeys = validTestSigningKeys(2)
			v.SigningKeys[0], v.SigningKeys[1] = v.SigningKeys[1], v.SigningKeys[0]
		}},
		{name: "short root key ID", mutate: func(v *RootMetadataV1) { v.RootKeyID = "short" }},
		{name: "short key ID", mutate: func(v *RootMetadataV1) { v.SigningKeys[0].KeyID = "short" }},
		{name: "unknown algorithm", mutate: func(v *RootMetadataV1) { v.SigningKeys[0].Algorithm = "unknown" }},
		{name: "bad public key", mutate: func(v *RootMetadataV1) { v.SigningKeys[0].PublicKey = "AA" }},
		{name: "unknown state", mutate: func(v *RootMetadataV1) { v.SigningKeys[0].State = "disabled" }},
		{name: "invalid key window", mutate: func(v *RootMetadataV1) { v.SigningKeys[0].NotAfter = v.SigningKeys[0].NotBefore }},
		{name: "key outside metadata", mutate: func(v *RootMetadataV1) { v.SigningKeys[0].NotBefore = "2026-08-08T23:59:59Z" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := base
			value.SigningKeys = append([]SigningKeyMetadataV1(nil), base.SigningKeys...)
			test.mutate(&value)
			if err := ValidateRootMetadataV1(value); err == nil {
				t.Fatal("ValidateRootMetadataV1 accepted invalid metadata")
			}
		})
	}
	unknown := []byte(strings.Replace(signed.PayloadJCS, `"root_algorithm":`, `"unknown":true,"root_algorithm":`, 1))
	if _, err := CanonicalizeJSON(unknown, new(RootMetadataV1)); err == nil {
		t.Fatal("accepted unknown metadata member")
	}
	duplicate := []byte(strings.Replace(signed.PayloadJCS, `"root_algorithm":`, `"schema_version":"other","root_algorithm":`, 1))
	if _, err := CanonicalizeJSON(duplicate, new(RootMetadataV1)); err == nil {
		t.Fatal("accepted duplicate metadata member")
	}
}

func validTestSigningKeys(count int) []SigningKeyMetadataV1 {
	keys := make([]SigningKeyMetadataV1, 0, count)
	for index := range count {
		seed := sha256.Sum256([]byte(fmt.Sprintf("task14-signing-key-%d", index)))
		privateKey := ed25519.NewKeyFromSeed(seed[:])
		publicKey := privateKey.Public().(ed25519.PublicKey)
		keys = append(keys, SigningKeyMetadataV1{
			KeyID: deriveKeyID(configKeyIDDomain, publicKey), Algorithm: "Ed25519", PublicKey: encodeBase64URL(publicKey), State: "future",
			NotBefore: "2026-08-09T00:00:00Z", NotAfter: "2026-08-20T00:00:00Z",
		})
		clear(privateKey)
	}
	sort.Slice(keys, func(left, right int) bool { return keys[left].KeyID < keys[right].KeyID })
	return keys
}

// TestMetadataKeySelectionRequiresPublishedActiveFullInterval catches signing
// with a future, retiring, revoked, absent, unpublished, or partially valid key.
func TestMetadataKeySelectionRequiresPublishedActiveFullInterval(t *testing.T) {
	signed, root, config := testMetadata(t, "1", "active")
	payload, err := VerifyRootMetadataV1(signed, trustedTestRoot(root), 0, time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("VerifyRootMetadataV1: %v", err)
	}
	from := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)
	if _, err := SelectSigningKeyForIssue(payload, config.KeyID(), from, to); err != nil {
		t.Fatalf("SelectSigningKeyForIssue active: %v", err)
	}
	states := []string{"future", "retiring", "revoked"}
	for _, state := range states {
		changed := payload
		changed.SigningKeys = append([]SigningKeyMetadataV1(nil), payload.SigningKeys...)
		changed.SigningKeys[0].State = state
		if _, err := SelectSigningKeyForIssue(changed, config.KeyID(), from, to); err == nil {
			t.Fatalf("selected %s key for issue", state)
		}
	}
	if _, err := SelectSigningKeyForIssue(payload, "unknown_key_id_000000", from, to); err == nil {
		t.Fatal("selected absent key")
	}
	changed := payload
	changed.SigningKeys = nil
	if _, err := SelectSigningKeyForIssue(changed, config.KeyID(), from, to); err == nil {
		t.Fatal("selected not-yet-published key")
	}
	changed = payload
	changed.SigningKeys = append([]SigningKeyMetadataV1(nil), payload.SigningKeys...)
	changed.SigningKeys[0].NotAfter = from.Add(12 * time.Hour).Format(time.RFC3339)
	if _, err := SelectSigningKeyForIssue(changed, config.KeyID(), from, to); err == nil {
		t.Fatal("selected key that does not cover full interval")
	}
}

// TestMetadataRetiringKeyVerifiesOnlyInsideGrace catches premature rejection of
// immutable bundles and accidental extension beyond the retiring not-after grace.
func TestMetadataRetiringKeyVerifiesOnlyInsideGrace(t *testing.T) {
	signed, root, config := testMetadata(t, "1", "retiring")
	payload, err := VerifyRootMetadataV1(signed, trustedTestRoot(root), 0, time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("VerifyRootMetadataV1: %v", err)
	}
	inside := time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC)
	if _, err := SelectSigningKeyForVerification(payload, config.KeyID(), inside.Add(-24*time.Hour), inside); err != nil {
		t.Fatalf("retiring key inside grace: %v", err)
	}
	outside := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	if _, err := SelectSigningKeyForVerification(payload, config.KeyID(), outside.Add(-24*time.Hour), outside); err == nil {
		t.Fatal("retiring key accepted outside grace")
	}
}

type memoryMetadataRepository struct {
	records []SignedRootMetadataV1
	writes  int
}

type conflictingMetadataRepository struct {
	mu          sync.Mutex
	records     []SignedRootMetadataV1
	emptyLists  int
	firstReady  chan struct{}
	secondReady chan struct{}
}

type rollbackMetadataRepository struct {
	mu         sync.Mutex
	listCalls  int
	records    []SignedRootMetadataV1
	firstReady chan struct{}
	published  chan struct{}
}

func (repository *rollbackMetadataRepository) List(context.Context) ([]SignedRootMetadataV1, error) {
	repository.mu.Lock()
	repository.listCalls++
	call := repository.listCalls
	if call == 2 {
		close(repository.firstReady)
	}
	repository.mu.Unlock()
	if call <= 2 {
		<-repository.firstReady
		return nil, nil
	}
	if call == 3 {
		return nil, nil
	}
	if call == 4 {
		<-repository.published
	}
	repository.mu.Lock()
	records := append([]SignedRootMetadataV1(nil), repository.records...)
	repository.mu.Unlock()
	return records, nil
}

func (repository *rollbackMetadataRepository) Publish(_ context.Context, value SignedRootMetadataV1) error {
	repository.mu.Lock()
	repository.records = []SignedRootMetadataV1{value}
	repository.mu.Unlock()
	close(repository.published)
	return nil
}

func (repository *conflictingMetadataRepository) List(context.Context) ([]SignedRootMetadataV1, error) {
	repository.mu.Lock()
	if repository.emptyLists < 4 {
		repository.emptyLists++
		listNumber := repository.emptyLists
		if listNumber == 2 {
			close(repository.firstReady)
		}
		if listNumber == 4 {
			close(repository.secondReady)
		}
		repository.mu.Unlock()
		if listNumber <= 2 {
			<-repository.firstReady
		} else {
			<-repository.secondReady
		}
		return nil, nil
	}
	records := append([]SignedRootMetadataV1(nil), repository.records...)
	repository.mu.Unlock()
	return records, nil
}

func (repository *conflictingMetadataRepository) Publish(_ context.Context, value SignedRootMetadataV1) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if len(repository.records) != 0 {
		return errors.New("injected publication conflict")
	}
	repository.records = []SignedRootMetadataV1{value}
	return nil
}

func (repository *memoryMetadataRepository) List(context.Context) ([]SignedRootMetadataV1, error) {
	return append([]SignedRootMetadataV1(nil), repository.records...), nil
}
func (repository *memoryMetadataRepository) Publish(_ context.Context, value SignedRootMetadataV1) error {
	repository.records = append(repository.records, value)
	repository.writes++
	return nil
}

// TestMetadataPublicationRejectsRollbackAndIsLocalIdempotent catches replacing
// a newer complete set and generating/publishing fixture keys more than once.
func TestMetadataPublicationRejectsRollbackAndIsLocalIdempotent(t *testing.T) {
	version2, root, _ := testMetadata(t, "2", "active")
	repository := &memoryMetadataRepository{records: []SignedRootMetadataV1{version2}}
	version1, _, _ := testMetadata(t, "1", "active")
	now := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	if err := PublishMetadata(context.Background(), repository, trustedTestRoot(root), version1, now); !errors.Is(err, ErrMetadataRollback) {
		t.Fatalf("rollback publish error = %v, want ErrMetadataRollback", err)
	}
	if repository.writes != 0 {
		t.Fatal("rollback mutated repository")
	}

	empty := &memoryMetadataRepository{}
	localRoot, err := NewLocalRootSigner(secret.NewBytes(bytes.Repeat([]byte{0x77}, ed25519.SeedSize)))
	if err != nil {
		t.Fatalf("NewLocalRootSigner: %v", err)
	}
	localConfig := newTestConfigSigner(t, 0x88)
	t.Cleanup(func() { _ = localRoot.Close() })
	validFrom := time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC)
	validUntil := validFrom.Add(30 * 24 * time.Hour)
	if err := EnsureLocalMetadata(context.Background(), empty, localRoot, localConfig, validFrom, validUntil); err != nil {
		t.Fatalf("EnsureLocalMetadata first: %v", err)
	}
	first := empty.records[0]
	const wantPayloadJCS = `{"root_algorithm":"Ed25519","root_key_id":"-zSxe3tI8vlRfEuEgJHzTA","schema_version":"talenro-trust-metadata/v1","signing_keys":[{"algorithm":"Ed25519","key_id":"-yW43uUMh7yR0mn70Lbqmw","not_after":"2026-09-08T00:00:00Z","not_before":"2026-08-09T00:00:00Z","public_key":"skkdlQKuKGMKK6yy4MdFEP_N0yjDNP8-E5PnWy0x59w","state":"active"}],"valid_from":"2026-08-09T00:00:00Z","valid_until":"2026-09-08T00:00:00Z","version":"1"}`
	const wantSignature = `N90e05HBRFMXZBv6xA8Rzhi5m71QG2wqXYocOVU6HWjY5waX06s4U2dOLuopLhpNfURZsSbxWpF96dszRAzeAw`
	if first.PayloadJCS != wantPayloadJCS || first.Signature != wantSignature {
		t.Fatalf("deterministic metadata = %#v", first)
	}
	if err := EnsureLocalMetadata(context.Background(), empty, localRoot, localConfig, validFrom, validUntil); err != nil {
		t.Fatalf("EnsureLocalMetadata second: %v", err)
	}
	if empty.writes != 1 || fmt.Sprint(empty.records) != fmt.Sprint([]SignedRootMetadataV1{first}) {
		t.Fatalf("idempotent publication writes=%d records=%v", empty.writes, empty.records)
	}
}

// TestMetadataPublicationAuthenticatesHistoryBeforeMonotonicity catches a
// corrupt higher-version record becoming the authoritative rollback base.
func TestMetadataPublicationAuthenticatesHistoryBeforeMonotonicity(t *testing.T) {
	existing, root, _ := testMetadata(t, "2", "active")
	existing.Signature = tamperFirstCharacter(existing.Signature)
	repository := &memoryMetadataRepository{records: []SignedRootMetadataV1{existing}}
	candidate, err := signTestMetadataVersion(t, root, "1")
	if err != nil {
		t.Fatalf("sign candidate: %v", err)
	}
	now := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	if err := PublishMetadata(context.Background(), repository, trustedTestRoot(root), candidate, now); !errors.Is(err, ErrRepository) {
		t.Fatalf("PublishMetadata corrupt history error = %v, want ErrRepository", err)
	}
	if repository.writes != 0 {
		t.Fatal("corrupt history allowed publication")
	}
}

// TestMetadataPublicationBoundsHistory catches an unbounded repository result
// forcing signature work over more than the finite metadata-version limit.
func TestMetadataPublicationBoundsHistory(t *testing.T) {
	existing, root, _ := testMetadata(t, "2", "active")
	repository := &memoryMetadataRepository{records: make([]SignedRootMetadataV1, maximumMetadataVersions+1)}
	for index := range repository.records {
		repository.records[index] = existing
	}
	candidate, err := signTestMetadataVersion(t, root, "3")
	if err != nil {
		t.Fatalf("sign candidate: %v", err)
	}
	now := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	if err := PublishMetadata(context.Background(), repository, trustedTestRoot(root), candidate, now); !errors.Is(err, ErrRepository) {
		t.Fatalf("PublishMetadata oversized history error = %v, want ErrRepository", err)
	}
}

// TestMetadataEnsureLocalIsConcurrentConflictIdempotent catches returning an
// error to the loser after an identical deterministic v1 wins publication.
func TestMetadataEnsureLocalIsConcurrentConflictIdempotent(t *testing.T) {
	repository := &conflictingMetadataRepository{firstReady: make(chan struct{}), secondReady: make(chan struct{})}
	root, err := NewLocalRootSigner(secret.NewBytes(bytes.Repeat([]byte{0x77}, ed25519.SeedSize)))
	if err != nil {
		t.Fatalf("NewLocalRootSigner: %v", err)
	}
	config := newTestConfigSigner(t, 0x88)
	t.Cleanup(func() { _ = root.Close() })
	validFrom := time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC)
	validUntil := validFrom.Add(30 * 24 * time.Hour)
	errorsByCall := make(chan error, 2)
	var callers sync.WaitGroup
	callers.Add(2)
	for range 2 {
		go func() {
			defer callers.Done()
			errorsByCall <- EnsureLocalMetadata(context.Background(), repository, root, config, validFrom, validUntil)
		}()
	}
	callers.Wait()
	close(errorsByCall)
	for err := range errorsByCall {
		if err != nil {
			t.Fatalf("concurrent EnsureLocalMetadata: %v", err)
		}
	}
	if len(repository.records) != 1 {
		t.Fatalf("published records = %d, want 1", len(repository.records))
	}
}

// TestMetadataEnsureLocalIsConcurrentRollbackIdempotent catches returning a
// rollback error when the winning deterministic v1 appears before revalidation.
func TestMetadataEnsureLocalIsConcurrentRollbackIdempotent(t *testing.T) {
	repository := &rollbackMetadataRepository{firstReady: make(chan struct{}), published: make(chan struct{})}
	root, err := NewLocalRootSigner(secret.NewBytes(bytes.Repeat([]byte{0x77}, ed25519.SeedSize)))
	if err != nil {
		t.Fatalf("NewLocalRootSigner: %v", err)
	}
	config := newTestConfigSigner(t, 0x88)
	t.Cleanup(func() { _ = root.Close() })
	validFrom := time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC)
	validUntil := validFrom.Add(30 * 24 * time.Hour)
	errorsByCall := make(chan error, 2)
	var callers sync.WaitGroup
	callers.Add(2)
	for range 2 {
		go func() {
			defer callers.Done()
			errorsByCall <- EnsureLocalMetadata(context.Background(), repository, root, config, validFrom, validUntil)
		}()
	}
	callers.Wait()
	close(errorsByCall)
	for err := range errorsByCall {
		if err != nil {
			t.Fatalf("concurrent EnsureLocalMetadata: %v", err)
		}
	}
}

func signTestMetadataVersion(t *testing.T, root *LocalRootSigner, version string) (SignedRootMetadataV1, error) {
	t.Helper()
	config := newTestConfigSigner(t, 0x6a)
	payload := RootMetadataV1{
		SchemaVersion: TrustMetadataSchemaV1, Version: version, RootKeyID: root.KeyID(), RootAlgorithm: SignatureAlgorithm,
		ValidFrom: "2026-08-09T00:00:00Z", ValidUntil: "2026-09-09T00:00:00Z",
		SigningKeys: []SigningKeyMetadataV1{{
			KeyID: config.KeyID(), Algorithm: SignatureAlgorithm, PublicKey: encodeBase64URL(config.PublicKey()), State: "active",
			NotBefore: "2026-08-09T00:00:00Z", NotAfter: "2026-08-20T00:00:00Z",
		}},
	}
	return SignRootMetadataV1(context.Background(), payload, newTestBoundedSigner(t, root))
}

// TestMetadataPostgresPublicationIsOneCompleteGeneratedQueryTransaction catches
// partial publication, signing keys inserted before their root, missing rollback,
// accepting a bad root signature, or committing an incomplete key set.
func TestMetadataPostgresPublicationIsOneCompleteGeneratedQueryTransaction(t *testing.T) {
	signed, root, _ := testMetadata(t, "1", "active")
	secondConfig := newTestConfigSigner(t, 0x99)
	var payload RootMetadataV1
	if _, err := CanonicalizeJSON([]byte(signed.PayloadJCS), &payload); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	payload.SigningKeys = append(payload.SigningKeys, SigningKeyMetadataV1{
		KeyID: secondConfig.KeyID(), Algorithm: "Ed25519", PublicKey: encodeBase64URL(secondConfig.PublicKey()), State: "future",
		NotBefore: "2026-08-11T00:00:00Z", NotAfter: "2026-08-20T00:00:00Z",
	})
	if payload.SigningKeys[1].KeyID < payload.SigningKeys[0].KeyID {
		payload.SigningKeys[0], payload.SigningKeys[1] = payload.SigningKeys[1], payload.SigningKeys[0]
	}
	signed, err := SignRootMetadataV1(context.Background(), payload, newTestBoundedSigner(t, root))
	if err != nil {
		t.Fatalf("SignRootMetadataV1 two keys: %v", err)
	}
	tx := &metadataPGXTx{}
	database := &metadataPGXDatabase{tx: tx}
	repository, err := NewPostgresMetadataRepository(database, trustedTestRoot(root))
	if err != nil {
		t.Fatalf("NewPostgresMetadataRepository: %v", err)
	}
	createdAt := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	repository.now = func() time.Time { return createdAt }
	if err := repository.Publish(context.Background(), signed); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if tx.commitCalls != 1 || tx.rollbackCalls != 1 {
		t.Fatalf("commit/rollback calls = %d/%d, want 1/1", tx.commitCalls, tx.rollbackCalls)
	}
	if len(tx.execSQL) != 3 || !strings.Contains(tx.execSQL[0], "trust.trust_root_metadata") ||
		!strings.Contains(tx.execSQL[1], "trust.signing_key_metadata") || !strings.Contains(tx.execSQL[2], "trust.signing_key_metadata") {
		t.Fatalf("generated publication order = %v", tx.execSQL)
	}
	signature, decodeErr := base64.RawURLEncoding.DecodeString(signed.Signature)
	if decodeErr != nil {
		t.Fatalf("decode signature: %v", decodeErr)
	}
	wantRootArgs := []any{int64(1), []byte(signed.PayloadJCS), signature,
		time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC), createdAt}
	if len(tx.execArgs) != 3 || !reflect.DeepEqual(tx.execArgs[0], wantRootArgs) {
		t.Fatalf("root insert args = %#v", tx.execArgs)
	}
	for index, key := range payload.SigningKeys {
		publicKey, keyDecodeErr := base64.RawURLEncoding.DecodeString(key.PublicKey)
		if keyDecodeErr != nil {
			t.Fatalf("decode key %d: %v", index, keyDecodeErr)
		}
		wantKeyArgs := []any{key.KeyID, int64(1), publicKey, key.State,
			mustTestTime(t, key.NotBefore), mustTestTime(t, key.NotAfter)}
		if !reflect.DeepEqual(tx.execArgs[index+1], wantKeyArgs) {
			t.Fatalf("key %d insert args = %#v, want %#v", index, tx.execArgs[index+1], wantKeyArgs)
		}
	}

	bad := signed
	bad.Signature = tamperFirstCharacter(bad.Signature)
	badTx := &metadataPGXTx{}
	badRepository, err := NewPostgresMetadataRepository(&metadataPGXDatabase{tx: badTx}, trustedTestRoot(root))
	if err != nil {
		t.Fatalf("NewPostgresMetadataRepository bad: %v", err)
	}
	if err := badRepository.Publish(context.Background(), bad); err == nil {
		t.Fatal("Publish accepted invalid root signature")
	}
	if badTx.commitCalls != 0 || len(badTx.execSQL) != 0 {
		t.Fatal("invalid signature mutated transaction")
	}

	failingTx := &metadataPGXTx{failAt: 3}
	failingRepository, err := NewPostgresMetadataRepository(&metadataPGXDatabase{tx: failingTx}, trustedTestRoot(root))
	if err != nil {
		t.Fatalf("NewPostgresMetadataRepository failing: %v", err)
	}
	if err := failingRepository.Publish(context.Background(), signed); !errors.Is(err, ErrRepository) {
		t.Fatalf("incomplete Publish error = %v, want ErrRepository", err)
	}
	if failingTx.commitCalls != 0 || failingTx.rollbackCalls != 1 || len(failingTx.execSQL) != 3 {
		t.Fatalf("incomplete publication commit/rollback/exec = %d/%d/%d", failingTx.commitCalls, failingTx.rollbackCalls, len(failingTx.execSQL))
	}
}

// TestMetadataPostgresListReconstructsAndRejectsEveryStoredFieldMutation catches
// ignoring any root/key column or returning an unauthenticated partial row set.
func TestMetadataPostgresListReconstructsAndRejectsEveryStoredFieldMutation(t *testing.T) {
	signed, root, _ := testMetadata(t, "1", "active")
	rootRows, keyRows := storedMetadataTestRows(t, signed)
	repository, err := NewPostgresMetadataRepository(&metadataPGXDatabase{rootRows: rootRows, keyRows: keyRows}, trustedTestRoot(root))
	if err != nil {
		t.Fatalf("NewPostgresMetadataRepository: %v", err)
	}
	got, err := repository.List(context.Background())
	if err != nil || !reflect.DeepEqual(got, []SignedRootMetadataV1{signed}) {
		t.Fatalf("List = %#v, %v", got, err)
	}

	tests := []struct {
		name   string
		mutate func([][]any, [][]any)
	}{
		{name: "root version", mutate: func(roots, _ [][]any) { roots[0][0] = int64(2) }},
		{name: "root payload", mutate: func(roots, _ [][]any) { roots[0][1] = []byte(`{}`) }},
		{name: "root signature", mutate: func(roots, _ [][]any) { roots[0][2].([]byte)[0] ^= 1 }},
		{name: "root valid from", mutate: func(roots, _ [][]any) { roots[0][3] = roots[0][3].(time.Time).Add(time.Second) }},
		{name: "root valid until", mutate: func(roots, _ [][]any) { roots[0][4] = roots[0][4].(time.Time).Add(time.Second) }},
		{name: "root created at", mutate: func(roots, _ [][]any) { roots[0][5] = time.Time{} }},
		{name: "key ID", mutate: func(_, keys [][]any) { keys[0][0] = strings.Repeat("x", 22) }},
		{name: "key root version", mutate: func(_, keys [][]any) { keys[0][1] = int64(2) }},
		{name: "key algorithm", mutate: func(_, keys [][]any) { keys[0][2] = "unknown" }},
		{name: "key public bytes", mutate: func(_, keys [][]any) { keys[0][3].([]byte)[0] ^= 1 }},
		{name: "key state", mutate: func(_, keys [][]any) { keys[0][4] = "revoked" }},
		{name: "key not before", mutate: func(_, keys [][]any) { keys[0][5] = keys[0][5].(time.Time).Add(time.Second) }},
		{name: "key not after", mutate: func(_, keys [][]any) { keys[0][6] = keys[0][6].(time.Time).Add(time.Second) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			roots := cloneTestRows(rootRows)
			keys := cloneTestRows(keyRows)
			test.mutate(roots, keys)
			repository, newErr := NewPostgresMetadataRepository(&metadataPGXDatabase{rootRows: roots, keyRows: keys}, trustedTestRoot(root))
			if newErr != nil {
				t.Fatalf("NewPostgresMetadataRepository: %v", newErr)
			}
			if _, listErr := repository.List(context.Background()); !errors.Is(listErr, ErrRepository) {
				t.Fatalf("List mutation error = %v, want ErrRepository", listErr)
			}
		})
	}
}

// TestMetadataPostgresAcceptsEquivalentTimestamptzLocations catches rejecting
// valid pgx timestamptz scans whose locations are local or fixed-offset zones.
func TestMetadataPostgresAcceptsEquivalentTimestamptzLocations(t *testing.T) {
	existing, root, _ := testMetadata(t, "1", "active")
	rootRows, keyRows := storedMetadataTestRows(t, existing)
	location := time.FixedZone("pgx-scan", 4*60*60)
	for _, field := range []int{3, 4, 5} {
		rootRows[0][field] = rootRows[0][field].(time.Time).In(location)
	}
	for _, field := range []int{5, 6} {
		keyRows[0][field] = keyRows[0][field].(time.Time).In(location)
	}
	database := &metadataPGXDatabase{rootRows: rootRows, keyRows: keyRows}
	repository, err := NewPostgresMetadataRepository(database, trustedTestRoot(root))
	if err != nil {
		t.Fatalf("NewPostgresMetadataRepository List: %v", err)
	}
	got, err := repository.List(context.Background())
	if err != nil || !reflect.DeepEqual(got, []SignedRootMetadataV1{existing}) {
		t.Fatalf("List fixed-offset timestamptz = %#v, %v", got, err)
	}

	candidate, err := signTestMetadataVersion(t, root, "2")
	if err != nil {
		t.Fatalf("sign candidate: %v", err)
	}
	tx := &metadataPGXTx{rootRows: rootRows, keyRows: keyRows}
	repository, err = NewPostgresMetadataRepository(&metadataPGXDatabase{tx: tx}, trustedTestRoot(root))
	if err != nil {
		t.Fatalf("NewPostgresMetadataRepository Publish: %v", err)
	}
	if err := repository.Publish(context.Background(), candidate); err != nil {
		t.Fatalf("Publish with fixed-offset authenticated history: %v", err)
	}
	if tx.commitCalls != 1 {
		t.Fatalf("commit calls = %d, want 1", tx.commitCalls)
	}
}

// TestMetadataPostgresPublishAuthenticatesStoredRowsInTransaction catches
// deriving rollback state from a corrupt root row without its complete keys.
func TestMetadataPostgresPublishAuthenticatesStoredRowsInTransaction(t *testing.T) {
	existing, root, _ := testMetadata(t, "2", "active")
	rootRows, keyRows := storedMetadataTestRows(t, existing)
	rootRows[0][2].([]byte)[0] ^= 1
	candidate, err := signTestMetadataVersion(t, root, "1")
	if err != nil {
		t.Fatalf("sign candidate: %v", err)
	}
	tx := &metadataPGXTx{rootRows: rootRows, keyRows: keyRows}
	repository, err := NewPostgresMetadataRepository(&metadataPGXDatabase{tx: tx}, trustedTestRoot(root))
	if err != nil {
		t.Fatalf("NewPostgresMetadataRepository: %v", err)
	}
	if err := repository.Publish(context.Background(), candidate); !errors.Is(err, ErrRepository) {
		t.Fatalf("Publish corrupt stored history error = %v, want ErrRepository", err)
	}
	if tx.commitCalls != 0 || len(tx.execSQL) != 0 {
		t.Fatal("corrupt stored history allowed mutation")
	}
}

func storedMetadataTestRows(t *testing.T, signed SignedRootMetadataV1) ([][]any, [][]any) {
	t.Helper()
	var payload RootMetadataV1
	if _, err := CanonicalizeJSON([]byte(signed.PayloadJCS), &payload); err != nil {
		t.Fatalf("decode metadata payload: %v", err)
	}
	version, err := strconv.ParseInt(payload.Version, 10, 64)
	if err != nil {
		t.Fatalf("parse version: %v", err)
	}
	signature, err := base64.RawURLEncoding.DecodeString(signed.Signature)
	if err != nil {
		t.Fatalf("decode signature: %v", err)
	}
	roots := [][]any{{version, []byte(signed.PayloadJCS), signature, mustTestTime(t, payload.ValidFrom),
		mustTestTime(t, payload.ValidUntil), time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)}}
	keys := make([][]any, 0, len(payload.SigningKeys))
	for _, key := range payload.SigningKeys {
		publicKey, decodeErr := base64.RawURLEncoding.DecodeString(key.PublicKey)
		if decodeErr != nil {
			t.Fatalf("decode public key: %v", decodeErr)
		}
		keys = append(keys, []any{key.KeyID, version, key.Algorithm, publicKey, key.State,
			mustTestTime(t, key.NotBefore), mustTestTime(t, key.NotAfter)})
	}
	return roots, keys
}

func cloneTestRows(rows [][]any) [][]any {
	clone := make([][]any, len(rows))
	for index, row := range rows {
		clone[index] = append([]any(nil), row...)
		for field, value := range clone[index] {
			if bytesValue, ok := value.([]byte); ok {
				clone[index][field] = bytes.Clone(bytesValue)
			}
		}
	}
	return clone
}

func mustTestTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}
	return parsed
}

type metadataPGXDatabase struct {
	tx       pgx.Tx
	rootRows [][]any
	keyRows  [][]any
}

func (database *metadataPGXDatabase) Begin(context.Context) (pgx.Tx, error) { return database.tx, nil }
func (*metadataPGXDatabase) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("direct Exec not expected")
}
func (database *metadataPGXDatabase) Query(_ context.Context, sql string, _ ...interface{}) (pgx.Rows, error) {
	return metadataQueryRows(sql, database.rootRows, database.keyRows), nil
}
func (*metadataPGXDatabase) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	return metadataRow{}
}

type metadataPGXTx struct {
	pgx.Tx
	execSQL       []string
	execArgs      [][]any
	failAt        int
	commitCalls   int
	rollbackCalls int
	rootRows      [][]any
	keyRows       [][]any
}

func (tx *metadataPGXTx) Exec(_ context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	tx.execSQL = append(tx.execSQL, sql)
	tx.execArgs = append(tx.execArgs, append([]any(nil), arguments...))
	if tx.failAt == len(tx.execSQL) {
		return pgconn.CommandTag{}, errors.New("injected metadata insert failure")
	}
	return pgconn.CommandTag{}, nil
}
func (tx *metadataPGXTx) Query(_ context.Context, sql string, _ ...interface{}) (pgx.Rows, error) {
	return metadataQueryRows(sql, tx.rootRows, tx.keyRows), nil
}
func (*metadataPGXTx) QueryRow(context.Context, string, ...interface{}) pgx.Row { return metadataRow{} }
func (tx *metadataPGXTx) Commit(context.Context) error {
	tx.commitCalls++
	return nil
}
func (tx *metadataPGXTx) Rollback(context.Context) error {
	tx.rollbackCalls++
	return nil
}

type metadataRows struct {
	pgx.Rows
	rows  [][]any
	index int
}

func metadataQueryRows(sql string, rootRows, keyRows [][]any) *metadataRows {
	if strings.Contains(sql, "trust_root_metadata") {
		return &metadataRows{rows: cloneTestRows(rootRows)}
	}
	if strings.Contains(sql, "signing_key_metadata") {
		return &metadataRows{rows: cloneTestRows(keyRows)}
	}
	return &metadataRows{}
}

func (*metadataRows) Close()          {}
func (*metadataRows) Err() error      { return nil }
func (rows *metadataRows) Next() bool { return rows.index < len(rows.rows) }
func (rows *metadataRows) Scan(destinations ...any) error {
	if rows.index >= len(rows.rows) || len(destinations) != len(rows.rows[rows.index]) {
		return errors.New("invalid scan")
	}
	for index, destination := range destinations {
		reflect.ValueOf(destination).Elem().Set(reflect.ValueOf(rows.rows[rows.index][index]))
	}
	rows.index++
	return nil
}

type metadataRow struct{}

func (metadataRow) Scan(...interface{}) error { return errors.New("row not expected") }
