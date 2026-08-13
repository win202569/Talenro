package trustclient_test

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/hpke"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/gowebpki/jcs"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/trust"
	"talenro.local/platform/internal/trustclient"
)

const (
	testAudience = "7fa85f64-5717-4562-b3fc-2c963f66afa6"
	testBundleID = "d64cc450-b7eb-4575-9d8a-8a304096e719"
)

// TestVerifyCrossPackageSealOpenStagesAndActivates catches client reuse of
// server verification flow, wrong HPKE info/AAD, or returning before durable
// stage and activation.
func TestVerifyCrossPackageSealOpenStagesAndActivates(t *testing.T) {
	fixture := newVerifyFixture(t, "1")
	store, err := trustclient.NewDirectoryStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewDirectoryStore: %v", err)
	}
	verified, err := trustclient.VerifyAndStage(
		context.Background(), fixture.envelope, fixture.recipient, fixture.metadata,
		fixture.expected, store, fixture.now, 120*time.Second,
	)
	if err != nil {
		t.Fatalf("VerifyAndStage: %v", err)
	}
	if verified.Version() != 1 || !bytes.Equal(verified.PayloadJCS(), []byte(fixture.signed.PayloadJCS)) {
		t.Fatal("verified result does not contain the exact trusted payload")
	}
	highest, err := store.LoadHighest(context.Background(), testAudience)
	if err != nil || highest != 1 {
		t.Fatalf("highest = %d, err %v", highest, err)
	}
	activeVersion, active, err := store.LoadActive(context.Background(), testAudience)
	if err != nil || activeVersion != 1 || !bytes.Equal(active, []byte(fixture.signed.PayloadJCS)) {
		t.Fatalf("active version/bytes = %d/%q, err %v", activeVersion, active, err)
	}
}

// TestVerifyRejectsWrongRecipientAndOuterTamperDimensions catches parsing or
// opening an envelope without authenticating every public header member.
func TestVerifyRejectsWrongRecipientAndOuterTamperDimensions(t *testing.T) {
	fixture := newVerifyFixture(t, "1")
	wrong, err := ecdh.X25519().NewPrivateKey(bytes.Repeat([]byte{0xe1}, 32))
	if err != nil {
		t.Fatalf("wrong key: %v", err)
	}
	if _, err := runVerifyFixture(t, fixture, fixture.envelope, wrong, trustclient.NewMemoryStore()); err == nil {
		t.Fatal("accepted wrong recipient key")
	}

	mutations := []struct {
		name  string
		field string
		value string
	}{
		{name: "envelope version", field: "envelope_version", value: "future"},
		{name: "KEM", field: "kem", value: "DHKEM-P256-HKDF-SHA256"},
		{name: "KDF", field: "kdf", value: "HKDF-SHA512"},
		{name: "AEAD", field: "aead", value: "AES-128-GCM"},
		{name: "recipient selector", field: "recipient_key_id", value: base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{9}, 16))},
		{name: "locator", field: "bundle_locator", value: base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{8}, 32))},
		{name: "enc", field: "enc", value: base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))},
		{name: "ciphertext", field: "ciphertext", value: base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{6}, 32))},
		{name: "plaintext over bucket limit", field: "ciphertext", value: base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{5}, (64<<10)+17))},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			changed := mutateOuterFixture(t, fixture.envelope, mutation.field, mutation.value)
			if _, err := runVerifyFixture(t, fixture, changed, fixture.recipient, trustclient.NewMemoryStore()); err == nil {
				t.Fatal("accepted outer-envelope tamper")
			}
		})
	}
	var outer map[string]any
	if err := json.Unmarshal(fixture.envelope, &outer); err != nil {
		t.Fatalf("decode same-length ciphertext fixture: %v", err)
	}
	ciphertext, err := base64.RawURLEncoding.DecodeString(outer["ciphertext"].(string))
	if err != nil {
		t.Fatalf("decode same-length ciphertext: %v", err)
	}
	ciphertext[len(ciphertext)/2] ^= 1
	changed := mutateOuterFixture(t, fixture.envelope, "ciphertext", base64.RawURLEncoding.EncodeToString(ciphertext))
	clear(ciphertext)
	if _, err := runVerifyFixture(t, fixture, changed, fixture.recipient, trustclient.NewMemoryStore()); err == nil {
		t.Fatal("accepted same-length authenticated ciphertext tamper")
	}
}

// TestVerifyRejectsStrictSignedPayloadAndPolicyJSON catches allowing missing,
// duplicate, unknown, trailing, or non-I-JSON members after successful HPKE.
func TestVerifyRejectsStrictSignedPayloadAndPolicyJSON(t *testing.T) {
	fixture := newVerifyFixture(t, "1")
	validSigned, err := json.Marshal(fixture.signed)
	if err != nil {
		t.Fatalf("marshal signed fixture: %v", err)
	}
	validSigned, err = jcs.Transform(validSigned)
	if err != nil {
		t.Fatalf("canonical signed fixture: %v", err)
	}
	closing := bytes.LastIndexByte(validSigned, '}')
	signedCases := map[string][]byte{
		"signed duplicate": append(append([]byte(nil), validSigned[:closing]...), []byte(`,"padding":""}`)...),
		"signed unknown":   append(append([]byte(nil), validSigned[:closing]...), []byte(`,"unknown":true}`)...),
		"signed trailing":  append(append([]byte(nil), validSigned...), []byte(` {}`)...),
		"signed missing":   bytes.Replace(validSigned, []byte(`,"padding":""`), nil, 1),
		"signed surrogate": bytes.Replace(validSigned, []byte(`"padding":""`), []byte(`"padding":"\ud800"`), 1),
	}
	for name, signedBytes := range signedCases {
		t.Run(name, func(t *testing.T) {
			envelope := sealRawSignedFixture(t, fixture, signedBytes)
			if _, err := runVerifyFixture(t, fixture, envelope, fixture.recipient, trustclient.NewMemoryStore()); err == nil {
				t.Fatal("accepted invalid signed JSON")
			}
		})
	}

	payloadCases := map[string]string{}
	payloadClosing := strings.LastIndex(fixture.signed.PayloadJCS, "}")
	payloadCases["payload duplicate"] = fixture.signed.PayloadJCS[:payloadClosing] + `,"audience":"` + testAudience + `"}`
	payloadCases["payload unknown"] = fixture.signed.PayloadJCS[:payloadClosing] + `,"unknown":true}`
	payloadCases["payload trailing"] = fixture.signed.PayloadJCS + ` {}`
	payloadCases["test config duplicate"] = strings.Replace(fixture.signed.PayloadJCS, `"sequence":"1"`, `"sequence":"1","sequence":"1"`, 1)
	payloadCases["policy duplicate"] = strings.Replace(fixture.signed.PayloadJCS, `"mode":"standard"`, `"mode":"standard","mode":"standard"`, 1)
	for name, payloadJCS := range payloadCases {
		t.Run(name, func(t *testing.T) {
			signed := fixture.signed
			signed.PayloadJCS = payloadJCS
			signedBody, marshalErr := json.Marshal(signed)
			if marshalErr != nil {
				t.Fatalf("marshal signed: %v", marshalErr)
			}
			signedBody, marshalErr = jcs.Transform(signedBody)
			if marshalErr != nil {
				t.Fatalf("canonical signed: %v", marshalErr)
			}
			envelope := sealRawSignedFixture(t, fixture, signedBody)
			if _, err := runVerifyFixture(t, fixture, envelope, fixture.recipient, trustclient.NewMemoryStore()); err == nil {
				t.Fatal("accepted invalid payload JSON")
			}
		})
	}
}

// TestVerifyRejectsStrictJSONAndBoundViolations catches duplicate/unknown/
// trailing/I-JSON acceptance or unbounded input before allocation and crypto.
func TestVerifyRejectsStrictJSONAndBoundViolations(t *testing.T) {
	fixture := newVerifyFixture(t, "1")
	closing := bytes.LastIndexByte(fixture.envelope, '}')
	unknown := append(append([]byte(nil), fixture.envelope[:closing]...), []byte(`,"unknown":true}`)...)
	duplicate := append(append([]byte(nil), fixture.envelope[:closing]...), []byte(`,"kem":"DHKEM-X25519-HKDF-SHA256"}`)...)
	cases := map[string][]byte{
		"trailing":       append(append([]byte(nil), fixture.envelope...), []byte(` {}`)...),
		"unknown":        unknown,
		"duplicate":      duplicate,
		"bad UTF-8":      append([]byte{'{'}, 0xff, '}'),
		"lone surrogate": []byte(strings.Replace(string(fixture.envelope), `"envelope_version":"`, `"envelope_version":"\ud800`, 1)),
		"over one MiB":   bytes.Repeat([]byte{'x'}, (1<<20)+1),
		"truncated":      fixture.envelope[:len(fixture.envelope)-1],
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := runVerifyFixture(t, fixture, body, fixture.recipient, trustclient.NewMemoryStore()); err == nil {
				t.Fatal("accepted invalid outer JSON")
			}
		})
	}
}

// TestVerifyRejectsEverySignedBindingAndSemanticDimension catches trusting a
// mutable digest, signature, key ID, audience, locator, schema, policy, or any
// time/version field after successful HPKE authentication.
func TestVerifyRejectsEverySignedBindingAndSemanticDimension(t *testing.T) {
	fixture := newVerifyFixture(t, "2")
	tests := []struct {
		name   string
		mutate func(map[string]any, *trust.SignedBundleV1)
	}{
		{name: "payload byte", mutate: func(_ map[string]any, signed *trust.SignedBundleV1) {
			signed.PayloadJCS = strings.Replace(signed.PayloadJCS, "hello", "jello", 1)
		}},
		{name: "digest", mutate: func(_ map[string]any, signed *trust.SignedBundleV1) {
			signed.PayloadSHA256 = tamperText(signed.PayloadSHA256)
		}},
		{name: "signature", mutate: func(_ map[string]any, signed *trust.SignedBundleV1) { signed.Signature = tamperText(signed.Signature) }},
		{name: "key ID", mutate: func(_ map[string]any, signed *trust.SignedBundleV1) {
			signed.SignerKeyID = tamperText(signed.SignerKeyID)
		}},
		{name: "audience", mutate: func(payload map[string]any, _ *trust.SignedBundleV1) {
			payload["audience"] = "38c0c6ef-3bc8-4df7-a69c-e8370ee45593"
		}},
		{name: "locator", mutate: func(payload map[string]any, _ *trust.SignedBundleV1) {
			payload["bundle_locator"] = base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x91}, 32))
		}},
		{name: "schema", mutate: func(payload map[string]any, _ *trust.SignedBundleV1) { payload["schema_version"] = "future" }},
		{name: "issued future", mutate: func(payload map[string]any, _ *trust.SignedBundleV1) {
			payload["issued_at"] = fixture.now.Add(121 * time.Second).Format(time.RFC3339)
			payload["not_before"] = fixture.now.Add(121 * time.Second).Format(time.RFC3339)
			payload["expires_at"] = fixture.now.Add(time.Hour).Format(time.RFC3339)
		}},
		{name: "not before future", mutate: func(payload map[string]any, _ *trust.SignedBundleV1) {
			payload["not_before"] = fixture.now.Add(121 * time.Second).Format(time.RFC3339)
			payload["expires_at"] = fixture.now.Add(time.Hour).Format(time.RFC3339)
		}},
		{name: "expired", mutate: func(payload map[string]any, _ *trust.SignedBundleV1) {
			payload["issued_at"] = fixture.now.Add(-time.Hour).Format(time.RFC3339)
			payload["not_before"] = fixture.now.Add(-time.Hour).Format(time.RFC3339)
			payload["expires_at"] = fixture.now.Add(-121 * time.Second).Format(time.RFC3339)
		}},
		{name: "zero bundle version", mutate: func(payload map[string]any, _ *trust.SignedBundleV1) { payload["bundle_version"] = "0" }},
		{name: "fractional bundle version", mutate: func(payload map[string]any, _ *trust.SignedBundleV1) { payload["bundle_version"] = "2.0" }},
		{name: "policy unknown", mutate: func(payload map[string]any, _ *trust.SignedBundleV1) {
			payload["policy_snapshot"] = map[string]any{"mode": "standard", "extra": true}
		}},
		{name: "sequence", mutate: func(payload map[string]any, _ *trust.SignedBundleV1) {
			payload["test_config"].(map[string]any)["sequence"] = "01"
		}},
		{name: "missing message member", mutate: func(payload map[string]any, _ *trust.SignedBundleV1) {
			delete(payload["test_config"].(map[string]any), "message")
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := fixture.payloadMap(t)
			signed := signPayloadFixture(t, payload, fixture.configSeed)
			test.mutate(payload, &signed)
			if test.name != "payload byte" && test.name != "digest" && test.name != "signature" && test.name != "key ID" {
				signed = signPayloadFixture(t, payload, fixture.configSeed)
			}
			envelope := fixture.seal(t, signed)
			if _, err := runVerifyFixture(t, fixture, envelope, fixture.recipient, trustclient.NewMemoryStore()); err == nil {
				t.Fatal("accepted signed or semantic tamper")
			}
		})
	}
}

// TestVerifyEnforcesFixedBundleLifetimeAndTrialGrace catches accepting a
// non-24-hour test bundle or a provisional policy extending beyond its fixed
// 24-hour grace boundary.
func TestVerifyEnforcesFixedBundleLifetimeAndTrialGrace(t *testing.T) {
	fixture := newVerifyFixture(t, "1")
	issuedAt := fixture.now.Add(-time.Minute)
	invalidLifetime := fixture.withPayloadMutation(t, func(payload map[string]any) {
		payload["expires_at"] = fixture.now.Add(time.Hour).Format(time.RFC3339)
	})
	if _, err := runVerifyFixture(t, invalidLifetime, invalidLifetime.envelope, invalidLifetime.recipient, trustclient.NewMemoryStore()); !errors.Is(err, trustclient.ErrPayload) {
		t.Fatalf("one-hour lifetime error = %v, want payload", err)
	}
	invalidGrace := fixture.withPayloadMutation(t, func(payload map[string]any) {
		payload["policy_snapshot"] = map[string]any{
			"mode": "trial_restricted", "max_devices": "1",
			"expires_at": issuedAt.Add(24*time.Hour + time.Second).Format(time.RFC3339),
		}
	})
	if _, err := runVerifyFixture(t, invalidGrace, invalidGrace.envelope, invalidGrace.recipient, trustclient.NewMemoryStore()); !errors.Is(err, trustclient.ErrPayload) {
		t.Fatalf("overlong trial grace error = %v, want payload", err)
	}
	validGrace := fixture.withPayloadMutation(t, func(payload map[string]any) {
		payload["policy_snapshot"] = map[string]any{
			"mode": "trial_restricted", "max_devices": "1",
			"expires_at": issuedAt.Add(23 * time.Hour).Format(time.RFC3339),
		}
	})
	if _, err := runVerifyFixture(t, validGrace, validGrace.envelope, validGrace.recipient, trustclient.NewMemoryStore()); err != nil {
		t.Fatalf("valid bounded trial grace: %v", err)
	}
}

// TestRollbackRejectsLowerEqualAndTrustMetadataVersions catches accepting a
// lower/equal conflicting config or a non-increasing root metadata record.
func TestRollbackRejectsLowerEqualAndTrustMetadataVersions(t *testing.T) {
	fixtureV2 := newVerifyFixture(t, "2")
	store, err := trustclient.NewDirectoryStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewDirectoryStore: %v", err)
	}
	if _, err := runVerifyFixture(t, fixtureV2, fixtureV2.envelope, fixtureV2.recipient, store); err != nil {
		t.Fatalf("verify version 2: %v", err)
	}
	for _, version := range []string{"1", "2"} {
		fixture := fixtureV2.withVersion(t, version)
		if _, err := runVerifyFixture(t, fixture, fixture.envelope, fixture.recipient, store); !errors.Is(err, trustclient.ErrRollback) {
			t.Fatalf("version %s error = %v, want rollback", version, err)
		}
	}
	rolledBackMetadata, err := trustclient.NewTrustMetadata(fixtureV2.metadataBytes, fixtureV2.roots, 1)
	if err != nil {
		t.Fatalf("NewTrustMetadata rollback fixture: %v", err)
	}
	fixtureV2.metadata = rolledBackMetadata
	if _, err := runVerifyFixture(t, fixtureV2, fixtureV2.envelope, fixtureV2.recipient, trustclient.NewMemoryStore()); !errors.Is(err, trustclient.ErrMetadataRollback) {
		t.Fatalf("metadata rollback error = %v", err)
	}
}

// TestVerifyRejectsMetadataEveryTimeVersionAndKeyState catches an independent
// client that checks only the root signature but skips metadata/key temporal
// semantics or permits a future/revoked signer.
func TestVerifyRejectsMetadataEveryTimeVersionAndKeyState(t *testing.T) {
	fixture := newVerifyFixture(t, "1")
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{name: "zero metadata version", mutate: func(payload map[string]any) { payload["version"] = "0" }},
		{name: "metadata valid from future", mutate: func(payload map[string]any) {
			payload["valid_from"] = fixture.now.Add(time.Second).Format(time.RFC3339)
		}},
		{name: "metadata valid until past", mutate: func(payload map[string]any) {
			payload["valid_until"] = fixture.now.Add(-time.Second).Format(time.RFC3339)
		}},
		{name: "key not before bundle", mutate: func(payload map[string]any) {
			payload["signing_keys"].([]any)[0].(map[string]any)["not_before"] = fixture.now.Add(time.Second).Format(time.RFC3339)
		}},
		{name: "key expires before bundle", mutate: func(payload map[string]any) {
			payload["signing_keys"].([]any)[0].(map[string]any)["not_after"] = fixture.now.Add(30 * time.Minute).Format(time.RFC3339)
		}},
		{name: "future key", mutate: func(payload map[string]any) { payload["signing_keys"].([]any)[0].(map[string]any)["state"] = "future" }},
		{name: "revoked key", mutate: func(payload map[string]any) { payload["signing_keys"].([]any)[0].(map[string]any)["state"] = "revoked" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := fixture.metadataMap(t)
			test.mutate(payload)
			metadataBytes := signMetadataFixture(t, payload, fixture.rootSeed)
			metadata, metadataErr := trustclient.NewTrustMetadata(metadataBytes, fixture.roots, 0)
			if metadataErr != nil {
				t.Fatalf("NewTrustMetadata: %v", metadataErr)
			}
			changed := fixture
			changed.metadata = metadata
			if _, err := runVerifyFixture(t, changed, changed.envelope, changed.recipient, trustclient.NewMemoryStore()); err == nil {
				t.Fatal("accepted invalid metadata time/version/state")
			}
		})
	}
}

// TestCrashAfterDurableStageRetainsHighestAndNeverActivatesLower simulates the
// exact stage/activate crash boundary and a fresh client process on restart.
func TestCrashAfterDurableStageRetainsHighestAndNeverActivatesLower(t *testing.T) {
	directory := t.TempDir()
	fixtureV2 := newVerifyFixture(t, "2")
	disk, err := trustclient.NewDirectoryStore(directory)
	if err != nil {
		t.Fatalf("NewDirectoryStore: %v", err)
	}
	crashing := &activateFailureStore{Store: disk}
	if _, err := runVerifyFixture(t, fixtureV2, fixtureV2.envelope, fixtureV2.recipient, crashing); !errors.Is(err, trustclient.ErrStore) {
		t.Fatalf("crash boundary error = %v", err)
	}

	restarted, err := trustclient.NewDirectoryStore(directory)
	if err != nil {
		t.Fatalf("restart store: %v", err)
	}
	if highest, err := restarted.LoadHighest(context.Background(), testAudience); err != nil || highest != 2 {
		t.Fatalf("restarted highest = %d, err %v", highest, err)
	}
	if _, _, err := restarted.LoadActive(context.Background(), testAudience); !errors.Is(err, trustclient.ErrStateNotFound) {
		t.Fatalf("active after crash error = %v", err)
	}
	fixtureV1 := fixtureV2.withVersion(t, "1")
	if _, err := runVerifyFixture(t, fixtureV1, fixtureV1.envelope, fixtureV1.recipient, restarted); !errors.Is(err, trustclient.ErrRollback) {
		t.Fatalf("lower after restart error = %v", err)
	}
	if _, _, err := restarted.LoadActive(context.Background(), testAudience); !errors.Is(err, trustclient.ErrStateNotFound) {
		t.Fatalf("lower bundle became active: %v", err)
	}
}

// TestVerifySensitiveTypesHaveFixedRedaction catches accidental metadata,
// locator, audience, policy, or plaintext exposure through generic surfaces.
func TestVerifySensitiveTypesHaveFixedRedaction(t *testing.T) {
	fixture := newVerifyFixture(t, "1")
	verified, err := runVerifyFixture(t, fixture, fixture.envelope, fixture.recipient, trustclient.NewMemoryStore())
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	const canary = "TASK15-CLIENT-CANARY"
	values := []any{fixture.metadata, fixture.expected, verified}
	for _, value := range values {
		formatted := fmt.Sprintf("%v %#v %+v", value, value, value)
		logged := slog.Any("value", value).Value.Resolve().String()
		if strings.Contains(formatted, canary) || strings.Contains(formatted, testAudience) || strings.Contains(logged, testAudience) {
			t.Fatalf("sensitive value rendered: %T", value)
		}
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("generic JSON accepted %T", value)
		}
	}
	_ = canary
}

type verifyFixture struct {
	now           time.Time
	configSeed    []byte
	rootSeed      []byte
	recipient     *ecdh.PrivateKey
	locator       [32]byte
	signed        trust.SignedBundleV1
	envelope      []byte
	metadata      trustclient.TrustMetadata
	metadataBytes []byte
	roots         map[string]ed25519.PublicKey
	expected      trustclient.Expected
}

type verifyTestingTB interface {
	Helper()
	Fatalf(string, ...any)
	Cleanup(func())
}

func newVerifyFixture(t verifyTestingTB, version string) verifyFixture {
	t.Helper()
	now := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)
	configSeed := bytes.Repeat([]byte{0x21}, ed25519.SeedSize)
	rootSeed := bytes.Repeat([]byte{0x31}, ed25519.SeedSize)
	configSigner, err := trust.NewLocalConfigSigner(secret.NewBytes(configSeed))
	if err != nil {
		t.Fatalf("new config signer: %v", err)
	}
	rootSigner, err := trust.NewLocalRootSigner(secret.NewBytes(rootSeed))
	if err != nil {
		t.Fatalf("new root signer: %v", err)
	}
	t.Cleanup(func() { _ = configSigner.Close(); _ = rootSigner.Close() })
	boundedConfig, err := trust.NewTimeoutConfigSigner(configSigner, 2*time.Second)
	if err != nil {
		t.Fatalf("bound config signer: %v", err)
	}
	boundedRoot, err := trust.NewTimeoutConfigSigner(rootSigner, 2*time.Second)
	if err != nil {
		t.Fatalf("bound root signer: %v", err)
	}
	t.Cleanup(func() { _ = boundedConfig.Close(); _ = boundedRoot.Close() })
	locator := [32]byte{0x51, 0x52, 0x53}
	payload := trust.PayloadV1{
		SchemaVersion: trust.BundleSchemaV1, BundleID: testBundleID,
		BundleLocator: base64.RawURLEncoding.EncodeToString(locator[:]), BundleVersion: version, Audience: testAudience,
		IssuedAt: now.Add(-time.Minute).Format(time.RFC3339), NotBefore: now.Format(time.RFC3339), ExpiresAt: now.Add(24 * time.Hour).Format(time.RFC3339),
		PolicySnapshot: json.RawMessage(`{"mode":"standard"}`), TestConfig: trust.TestConfigV1{Message: "hello", Sequence: "1"},
	}
	signed, err := trust.SignBundleV1(context.Background(), payload, boundedConfig)
	if err != nil {
		t.Fatalf("SignBundleV1: %v", err)
	}
	metadataPayload := trust.RootMetadataV1{
		SchemaVersion: trust.TrustMetadataSchemaV1, Version: "1", RootKeyID: rootSigner.KeyID(), RootAlgorithm: trust.SignatureAlgorithm,
		ValidFrom: now.Add(-24 * time.Hour).Format(time.RFC3339), ValidUntil: now.Add(30 * 24 * time.Hour).Format(time.RFC3339),
		SigningKeys: []trust.SigningKeyMetadataV1{{
			KeyID: configSigner.KeyID(), Algorithm: trust.SignatureAlgorithm,
			PublicKey: base64.RawURLEncoding.EncodeToString(configSigner.PublicKey()), State: "active",
			NotBefore: now.Add(-24 * time.Hour).Format(time.RFC3339), NotAfter: now.Add(48 * time.Hour).Format(time.RFC3339),
		}},
	}
	signedMetadata, err := trust.SignRootMetadataV1(context.Background(), metadataPayload, boundedRoot)
	if err != nil {
		t.Fatalf("SignRootMetadataV1: %v", err)
	}
	metadataBytes, err := json.Marshal(signedMetadata)
	if err != nil {
		t.Fatalf("marshal metadata: %v", err)
	}
	roots := map[string]ed25519.PublicKey{rootSigner.KeyID(): rootSigner.PublicKey()}
	metadata, err := trustclient.NewTrustMetadata(metadataBytes, roots, 0)
	if err != nil {
		t.Fatalf("NewTrustMetadata: %v", err)
	}
	expected, err := trustclient.NewExpected(testAudience, locator)
	if err != nil {
		t.Fatalf("NewExpected: %v", err)
	}
	recipient, err := ecdh.X25519().NewPrivateKey(bytes.Repeat([]byte{0x41}, 32))
	if err != nil {
		t.Fatalf("new recipient: %v", err)
	}
	fixture := verifyFixture{
		now: now, configSeed: append([]byte(nil), configSeed...), rootSeed: append([]byte(nil), rootSeed...), recipient: recipient, locator: locator,
		signed: signed, metadata: metadata, metadataBytes: metadataBytes, roots: roots, expected: expected,
	}
	fixture.envelope = fixture.seal(t, signed)
	return fixture
}

func (fixture verifyFixture) seal(t verifyTestingTB, signed trust.SignedBundleV1) []byte {
	t.Helper()
	var recipient [32]byte
	copy(recipient[:], fixture.recipient.PublicKey().Bytes())
	envelope, _, err := trust.Seal(bytes.NewReader(bytes.Repeat([]byte{0xc1}, 64<<10)), recipient, fixture.locator, signed)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	return envelope
}

func (fixture verifyFixture) payloadMap(t *testing.T) map[string]any {
	t.Helper()
	var payload map[string]any
	decoder := json.NewDecoder(strings.NewReader(fixture.signed.PayloadJCS))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		t.Fatalf("decode payload map: %v", err)
	}
	return payload
}

func (fixture verifyFixture) metadataMap(t *testing.T) map[string]any {
	t.Helper()
	var signed struct {
		PayloadJCS string `json:"payload_jcs"`
	}
	if err := json.Unmarshal(fixture.metadataBytes, &signed); err != nil {
		t.Fatalf("decode signed metadata: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(signed.PayloadJCS), &payload); err != nil {
		t.Fatalf("decode metadata payload: %v", err)
	}
	return payload
}

func (fixture verifyFixture) withVersion(t *testing.T, version string) verifyFixture {
	t.Helper()
	payload := fixture.payloadMap(t)
	payload["bundle_version"] = version
	fixture.signed = signPayloadFixture(t, payload, fixture.configSeed)
	fixture.envelope = fixture.seal(t, fixture.signed)
	return fixture
}

func (fixture verifyFixture) withPayloadMutation(t *testing.T, mutate func(map[string]any)) verifyFixture {
	t.Helper()
	payload := fixture.payloadMap(t)
	mutate(payload)
	fixture.signed = signPayloadFixture(t, payload, fixture.configSeed)
	fixture.envelope = fixture.seal(t, fixture.signed)
	return fixture
}

func signPayloadFixture(t *testing.T, payload map[string]any, seed []byte) trust.SignedBundleV1 {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload fixture: %v", err)
	}
	canonical, err := jcs.Transform(body)
	if err != nil {
		t.Fatalf("canonicalize payload fixture: %v", err)
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	message := append([]byte("TALENRO-CONFIG-BUNDLE-SIGNATURE-V1\x00"), canonical...)
	signature := ed25519.Sign(privateKey, message)
	digest := sha256.Sum256(canonical)
	idMaterial := append([]byte("TALENRO-CONFIG-SIGNING-KEY-ID-V1\x00"), publicKey...)
	idDigest := sha256.Sum256(idMaterial)
	clear(privateKey)
	clear(message)
	clear(idMaterial)
	return trust.SignedBundleV1{
		PayloadJCS: string(canonical), PayloadSHA256: base64.RawURLEncoding.EncodeToString(digest[:]),
		SignerKeyID: base64.RawURLEncoding.EncodeToString(idDigest[:16]), Algorithm: trust.SignatureAlgorithm,
		Signature: base64.RawURLEncoding.EncodeToString(signature), Padding: "",
	}
}

func signMetadataFixture(t *testing.T, payload map[string]any, seed []byte) []byte {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal metadata payload: %v", err)
	}
	canonical, err := jcs.Transform(body)
	if err != nil {
		t.Fatalf("canonicalize metadata payload: %v", err)
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	message := append([]byte("TALENRO-TRUST-METADATA-V1\x00"), canonical...)
	signature := ed25519.Sign(privateKey, message)
	clear(privateKey)
	clear(message)
	record, err := json.Marshal(struct {
		PayloadJCS string `json:"payload_jcs"`
		Signature  string `json:"signature"`
	}{PayloadJCS: string(canonical), Signature: base64.RawURLEncoding.EncodeToString(signature)})
	clear(signature)
	if err != nil {
		t.Fatalf("marshal metadata record: %v", err)
	}
	return record
}

func sealRawSignedFixture(t *testing.T, fixture verifyFixture, signedBytes []byte) []byte {
	t.Helper()
	plaintext := make([]byte, 4096)
	binary.BigEndian.PutUint32(plaintext[:4], uint32(len(signedBytes))) //nolint:gosec // Test frames are bounded by the 4096-byte fixture below.
	copy(plaintext[4:], signedBytes)
	publicKey, err := hpke.NewDHKEMPublicKey(fixture.recipient.PublicKey())
	if err != nil {
		t.Fatalf("adapt public key: %v", err)
	}
	enc, sender, err := hpke.NewSender(publicKey, hpke.HKDFSHA256(), hpke.ChaCha20Poly1305(), []byte("talenro-config-bundle/v1"))
	if err != nil {
		t.Fatalf("new sender: %v", err)
	}
	selectorMaterial := append([]byte("TALENRO-RECIPIENT-SELECTOR-V1\x00"), fixture.recipient.PublicKey().Bytes()...)
	selectorMaterial = append(selectorMaterial, fixture.locator[:]...)
	selector := sha256.Sum256(selectorMaterial)
	clear(selectorMaterial)
	header := map[string]any{
		"envelope_version": "talenro-config-envelope/v1", "kem": "DHKEM-X25519-HKDF-SHA256",
		"kdf": "HKDF-SHA256", "aead": "CHACHA20-POLY1305",
		"recipient_key_id": base64.RawURLEncoding.EncodeToString(selector[:16]),
		"bundle_locator":   base64.RawURLEncoding.EncodeToString(fixture.locator[:]), "enc": base64.RawURLEncoding.EncodeToString(enc),
	}
	headerBody, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("marshal raw header: %v", err)
	}
	aad, err := jcs.Transform(headerBody)
	if err != nil {
		t.Fatalf("canonical raw header: %v", err)
	}
	ciphertext, err := sender.Seal(aad, plaintext)
	clear(plaintext)
	clear(aad)
	if err != nil {
		t.Fatalf("seal raw signed frame: %v", err)
	}
	header["ciphertext"] = base64.RawURLEncoding.EncodeToString(ciphertext)
	clear(ciphertext)
	envelopeBody, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("marshal raw envelope: %v", err)
	}
	envelopeBody, err = jcs.Transform(envelopeBody)
	if err != nil {
		t.Fatalf("canonical raw envelope: %v", err)
	}
	return envelopeBody
}

func mutateOuterFixture(t *testing.T, body []byte, field, value string) []byte {
	t.Helper()
	var envelope map[string]any
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("decode outer fixture: %v", err)
	}
	envelope[field] = value
	encoded, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("encode outer fixture: %v", err)
	}
	canonical, err := jcs.Transform(encoded)
	if err != nil {
		t.Fatalf("canonicalize outer fixture: %v", err)
	}
	return canonical
}

func tamperText(value string) string {
	if strings.HasPrefix(value, "A") {
		return "B" + value[1:]
	}
	return "A" + value[1:]
}

func runVerifyFixture(t *testing.T, fixture verifyFixture, envelope []byte, key ecdh.KeyExchanger, store trustclient.Store) (trustclient.VerifiedBundle, error) {
	t.Helper()
	return trustclient.VerifyAndStage(context.Background(), envelope, key, fixture.metadata, fixture.expected, store, fixture.now, 120*time.Second)
}

type activateFailureStore struct{ trustclient.Store }

func (*activateFailureStore) Activate(context.Context, string, uint64) error {
	return errors.New("TASK15-RAW-ACTIVATE-CANARY")
}
