package main

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/trust"
)

// TestVerifyCLISeparateProcessAcceptsValidAndRejectsTamper catches a
// conformance command that bypasses the real file/store boundary, returns the
// wrong exit status, or discloses file contents and provider errors.
func TestVerifyCLISeparateProcessAcceptsValidAndRejectsTamper(t *testing.T) {
	fixture := writeConformanceFixture(t)
	binary := filepath.Join(t.TempDir(), "trust-conformance")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.CommandContext(t.Context(), "go", "build", "-buildvcs=false", "-o", binary, ".") //nolint:gosec // The test builds the fixed local package.
	build.Dir = "."
	build.Env = append(os.Environ(), "GOTOOLCHAIN=local")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v: %s", err, output)
	}

	stdout, stderr, code := runConformanceProcess(t, binary, fixture.envelopePath, fixture.metadataPath, fixture.privateKeyPath, t.TempDir())
	if code != 0 || stdout != "verified\n" || stderr != "" {
		t.Fatalf("valid CLI = code %d stdout %q stderr %q", code, stdout, stderr)
	}

	tampered := append([]byte(nil), fixture.envelope...)
	tampered[len(tampered)/2] ^= 1
	tamperedPath := filepath.Join(t.TempDir(), "tampered.json")
	if err := os.WriteFile(tamperedPath, tampered, 0o600); err != nil {
		t.Fatalf("write tamper: %v", err)
	}
	stdout, stderr, code = runConformanceProcess(t, binary, tamperedPath, fixture.metadataPath, fixture.privateKeyPath, t.TempDir())
	if code != 2 || stderr != "" || !finiteCLIOutput(stdout) || strings.Contains(stdout, string(tampered)) {
		t.Fatalf("tampered CLI = code %d stdout %q stderr %q", code, stdout, stderr)
	}

	canaryPath := filepath.Join(t.TempDir(), "private-canary")
	const canary = "TASK15-PRIVATE-KEY-CANARY"
	if err := os.WriteFile(canaryPath, []byte(canary), 0o600); err != nil {
		t.Fatalf("write canary: %v", err)
	}
	stdout, stderr, code = runConformanceProcess(t, binary, fixture.envelopePath, fixture.metadataPath, canaryPath, t.TempDir())
	if code != 2 || stderr != "" || !finiteCLIOutput(stdout) || strings.Contains(stdout, canary) {
		t.Fatalf("private-key failure = code %d stdout %q stderr %q", code, stdout, stderr)
	}
}

type conformanceFixture struct {
	envelope       []byte
	envelopePath   string
	metadataPath   string
	privateKeyPath string
}

func writeConformanceFixture(t *testing.T) conformanceFixture {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Second)
	config, err := trust.NewLocalConfigSigner(secret.NewBytes(bytes.Repeat([]byte{0x81}, ed25519.SeedSize)))
	if err != nil {
		t.Fatalf("config signer: %v", err)
	}
	root, err := trust.NewLocalRootSigner(secret.NewBytes(bytes.Repeat([]byte{0x82}, ed25519.SeedSize)))
	if err != nil {
		t.Fatalf("root signer: %v", err)
	}
	t.Cleanup(func() { _ = config.Close(); _ = root.Close() })
	boundedConfig, err := trust.NewTimeoutConfigSigner(config, 2*time.Second)
	if err != nil {
		t.Fatalf("bounded config: %v", err)
	}
	boundedRoot, err := trust.NewTimeoutConfigSigner(root, 2*time.Second)
	if err != nil {
		t.Fatalf("bounded root: %v", err)
	}
	t.Cleanup(func() { _ = boundedConfig.Close(); _ = boundedRoot.Close() })

	locator := [32]byte{0x91, 0x92, 0x93}
	const audience = "7fa85f64-5717-4562-b3fc-2c963f66afa6"
	payload := trust.PayloadV1{
		SchemaVersion: trust.BundleSchemaV1, BundleID: "d64cc450-b7eb-4575-9d8a-8a304096e719",
		BundleLocator: base64.RawURLEncoding.EncodeToString(locator[:]), BundleVersion: "1", Audience: audience,
		IssuedAt: now.Add(-time.Minute).Format(time.RFC3339), NotBefore: now.Format(time.RFC3339), ExpiresAt: now.Add(24 * time.Hour).Format(time.RFC3339),
		PolicySnapshot: json.RawMessage(`{"mode":"standard"}`), TestConfig: trust.TestConfigV1{Message: "hello", Sequence: "1"},
	}
	signed, err := trust.SignBundleV1(context.Background(), payload, boundedConfig)
	if err != nil {
		t.Fatalf("sign bundle: %v", err)
	}
	metadataPayload := trust.RootMetadataV1{
		SchemaVersion: trust.TrustMetadataSchemaV1, Version: "1", RootKeyID: root.KeyID(), RootAlgorithm: trust.SignatureAlgorithm,
		ValidFrom: now.Add(-time.Hour).Format(time.RFC3339), ValidUntil: now.Add(72 * time.Hour).Format(time.RFC3339),
		SigningKeys: []trust.SigningKeyMetadataV1{{
			KeyID: config.KeyID(), Algorithm: trust.SignatureAlgorithm, PublicKey: base64.RawURLEncoding.EncodeToString(config.PublicKey()), State: "active",
			NotBefore: now.Add(-time.Hour).Format(time.RFC3339), NotAfter: now.Add(48 * time.Hour).Format(time.RFC3339),
		}},
	}
	signedMetadata, err := trust.SignRootMetadataV1(context.Background(), metadataPayload, boundedRoot)
	if err != nil {
		t.Fatalf("sign metadata: %v", err)
	}
	privateKey, err := ecdh.X25519().NewPrivateKey(bytes.Repeat([]byte{0x83}, 32))
	if err != nil {
		t.Fatalf("private key: %v", err)
	}
	var recipient [32]byte
	copy(recipient[:], privateKey.PublicKey().Bytes())
	envelope, _, err := trust.Seal(bytes.NewReader(bytes.Repeat([]byte{0xa1}, 64<<10)), recipient, locator, signed)
	if err != nil {
		t.Fatalf("seal: %v", err)
	}

	metadataFile := struct {
		SignedMetadata        trust.SignedRootMetadataV1 `json:"signed_metadata"`
		TrustedRoots          []rootFile                 `json:"trusted_roots"`
		HighestTrusted        string                     `json:"highest_trusted_version"`
		ExpectedAudience      string                     `json:"expected_audience"`
		ExpectedBundleLocator string                     `json:"expected_bundle_locator"`
	}{
		SignedMetadata: signedMetadata,
		TrustedRoots:   []rootFile{{KeyID: root.KeyID(), PublicKey: base64.RawURLEncoding.EncodeToString(root.PublicKey())}},
		HighestTrusted: "0", ExpectedAudience: audience, ExpectedBundleLocator: base64.RawURLEncoding.EncodeToString(locator[:]),
	}
	metadataBytes, err := json.Marshal(metadataFile)
	if err != nil {
		t.Fatalf("marshal metadata file: %v", err)
	}
	directory := t.TempDir()
	envelopePath := filepath.Join(directory, "envelope.json")
	metadataPath := filepath.Join(directory, "metadata.json")
	privateKeyPath := filepath.Join(directory, "recipient.key")
	if err := os.WriteFile(envelopePath, envelope, 0o600); err != nil {
		t.Fatalf("write envelope: %v", err)
	}
	if err := os.WriteFile(metadataPath, metadataBytes, 0o600); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
	privateText := base64.RawURLEncoding.EncodeToString(privateKey.Bytes()) + "\n"
	if err := os.WriteFile(privateKeyPath, []byte(privateText), 0o600); err != nil {
		t.Fatalf("write private key: %v", err)
	}
	return conformanceFixture{envelope: envelope, envelopePath: envelopePath, metadataPath: metadataPath, privateKeyPath: privateKeyPath}
}

func runConformanceProcess(t *testing.T, binary, envelope, metadata, privateKey, state string) (string, string, int) {
	t.Helper()
	command := exec.CommandContext(t.Context(), binary, //nolint:gosec // The path is the just-built test binary in t.TempDir.
		"-envelope", envelope, "-metadata", metadata, "-private-key", privateKey, "-state-dir", state,
	)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if err == nil {
		return stdout.String(), stderr.String(), 0
	}
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		t.Fatalf("start CLI: %v", err)
	}
	return stdout.String(), stderr.String(), exitError.ExitCode()
}

func finiteCLIOutput(value string) bool {
	switch strings.TrimSpace(value) {
	case "arguments", "envelope", "metadata", "private_key", "state", "recipient", "signature", "payload", "rollback":
		return true
	default:
		return false
	}
}
