package trust

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"talenro.local/platform/internal/secret"
)

func newTestConfigSigner(t *testing.T, fill byte) *LocalConfigSigner {
	t.Helper()
	signer, err := NewLocalConfigSigner(secret.NewBytes(bytes.Repeat([]byte{fill}, ed25519.SeedSize)))
	if err != nil {
		t.Fatalf("NewLocalConfigSigner: %v", err)
	}
	t.Cleanup(func() { _ = signer.Close() })
	return signer
}

func newTestBoundedSigner(t *testing.T, signer ConfigSigner) *TimeoutConfigSigner {
	t.Helper()
	bounded, err := NewTimeoutConfigSigner(signer, localSignerTimeout)
	if err != nil {
		t.Fatalf("NewTimeoutConfigSigner: %v", err)
	}
	t.Cleanup(func() { _ = bounded.Close() })
	return bounded
}

// TestSignUsesExactEd25519DomainDigestAndKeyID catches signing a digest instead
// of exact payload JCS, omitting domain separation, or sharing root/config IDs.
func TestSignUsesExactEd25519DomainDigestAndKeyID(t *testing.T) {
	signer := newTestConfigSigner(t, 0x11)
	payload, err := DecodePayloadV1([]byte(validPayloadJSON))
	if err != nil {
		t.Fatalf("DecodePayloadV1: %v", err)
	}
	bundle, err := SignBundleV1(context.Background(), payload, newTestBoundedSigner(t, signer))
	if err != nil {
		t.Fatalf("SignBundleV1: %v", err)
	}
	digest := sha256.Sum256([]byte(bundle.PayloadJCS))
	if bundle.PayloadSHA256 != base64.RawURLEncoding.EncodeToString(digest[:]) {
		t.Fatalf("payload digest = %q", bundle.PayloadSHA256)
	}
	publicKey := signer.PublicKey()
	message := append([]byte("TALENRO-CONFIG-BUNDLE-SIGNATURE-V1\x00"), []byte(bundle.PayloadJCS)...)
	signature, decodeErr := base64.RawURLEncoding.DecodeString(bundle.Signature)
	if decodeErr != nil || !ed25519.Verify(publicKey, message, signature) {
		t.Fatal("independent Ed25519 verification failed")
	}
	idMaterial := append([]byte("TALENRO-CONFIG-SIGNING-KEY-ID-V1\x00"), publicKey...)
	idDigest := sha256.Sum256(idMaterial)
	if want := base64.RawURLEncoding.EncodeToString(idDigest[:16]); bundle.SignerKeyID != want {
		t.Fatalf("key ID = %q, want %q", bundle.SignerKeyID, want)
	}
	root, rootErr := NewLocalRootSigner(secret.NewBytes(bytes.Repeat([]byte{0x11}, ed25519.SeedSize)))
	if rootErr != nil {
		t.Fatalf("NewLocalRootSigner: %v", rootErr)
	}
	t.Cleanup(func() { _ = root.Close() })
	if root.KeyID() == signer.KeyID() {
		t.Fatal("root and config key-ID domains collided")
	}
}

// TestSignVerificationRejectsEveryTamperDimension catches verification that
// trusts a mutable digest/header instead of binding exact payload, ID, algorithm,
// digest, and signature to the selected public key.
func TestSignVerificationRejectsEveryTamperDimension(t *testing.T) {
	signer := newTestConfigSigner(t, 0x22)
	payload, err := DecodePayloadV1([]byte(validPayloadJSON))
	if err != nil {
		t.Fatalf("DecodePayloadV1: %v", err)
	}
	bundle, err := SignBundleV1(context.Background(), payload, newTestBoundedSigner(t, signer))
	if err != nil {
		t.Fatalf("SignBundleV1: %v", err)
	}
	if _, err := VerifySignedBundleV1(bundle, signer.PublicKey()); err != nil {
		t.Fatalf("VerifySignedBundleV1: %v", err)
	}
	tests := []struct {
		name   string
		mutate func(*SignedBundleV1)
	}{
		{name: "payload byte", mutate: func(v *SignedBundleV1) { v.PayloadJCS = strings.Replace(v.PayloadJCS, `"hello"`, `"jello"`, 1) }},
		{name: "payload digest", mutate: func(v *SignedBundleV1) { v.PayloadSHA256 = tamperFirstCharacter(v.PayloadSHA256) }},
		{name: "key ID", mutate: func(v *SignedBundleV1) { v.SignerKeyID = tamperFirstCharacter(v.SignerKeyID) }},
		{name: "algorithm", mutate: func(v *SignedBundleV1) { v.Algorithm = "unknown" }},
		{name: "signature", mutate: func(v *SignedBundleV1) { v.Signature = tamperFirstCharacter(v.Signature) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changed := bundle
			test.mutate(&changed)
			if _, err := VerifySignedBundleV1(changed, signer.PublicKey()); err == nil {
				t.Fatal("VerifySignedBundleV1 accepted tamper")
			}
		})
	}
}

func tamperFirstCharacter(value string) string {
	if value[0] == 'A' {
		return "B" + value[1:]
	}
	return "A" + value[1:]
}

type blockingConfigSigner struct{ keyID string }

func (signer blockingConfigSigner) KeyID() string { return signer.keyID }
func (blockingConfigSigner) Sign(ctx context.Context, _ []byte) ([]byte, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

type nonCooperativeConfigSigner struct {
	keyID   string
	starts  atomic.Int32
	release <-chan struct{}
}

type failingConfigSigner struct{ keyID string }

func (signer failingConfigSigner) KeyID() string { return signer.keyID }
func (failingConfigSigner) Sign(context.Context, []byte) ([]byte, error) {
	return nil, errors.New("SIGNER_PROVIDER_SECRET_CANARY")
}

type secretBearingConfigSigner struct {
	keyID  string
	secret string
}

type changingKeyIDSigner struct{ calls atomic.Int32 }

func (signer *changingKeyIDSigner) KeyID() string {
	if signer.calls.Add(1) == 1 {
		return strings.Repeat("k", 22)
	}
	return "changed"
}
func (*changingKeyIDSigner) Sign(context.Context, []byte) ([]byte, error) {
	return make([]byte, ed25519.SignatureSize), nil
}

func (signer secretBearingConfigSigner) KeyID() string { return signer.keyID }
func (secretBearingConfigSigner) Sign(context.Context, []byte) ([]byte, error) {
	return nil, errors.New("provider failure")
}

func (signer *nonCooperativeConfigSigner) KeyID() string { return signer.keyID }
func (signer *nonCooperativeConfigSigner) Sign(context.Context, []byte) ([]byte, error) {
	signer.starts.Add(1)
	<-signer.release
	return nil, errors.New("NONCOOPERATIVE_PROVIDER_CANARY")
}

// TestSignTimeoutIsBoundedAndValidated catches unbounded external signer calls
// and unsafe timeout configuration outside the frozen 500ms to 5s contract.
func TestSignTimeoutIsBoundedAndValidated(t *testing.T) {
	for _, timeout := range []time.Duration{499 * time.Millisecond, 5*time.Second + time.Nanosecond} {
		if _, err := NewTimeoutConfigSigner(blockingConfigSigner{keyID: strings.Repeat("k", 22)}, timeout); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("timeout %s error = %v, want ErrInvalidArgument", timeout, err)
		}
	}
	wrapped, err := NewTimeoutConfigSigner(blockingConfigSigner{keyID: strings.Repeat("k", 22)}, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("NewTimeoutConfigSigner: %v", err)
	}
	defer func() { _ = wrapped.Close() }()
	started := time.Now()
	_, err = wrapped.Sign(context.Background(), []byte("message"))
	elapsed := time.Since(started)
	if !errors.Is(err, ErrSignerTimeout) {
		t.Fatalf("Sign error = %v, want ErrSignerTimeout", err)
	}
	if elapsed < 450*time.Millisecond || elapsed > 2*time.Second {
		t.Fatalf("bounded signer elapsed = %s", elapsed)
	}
}

// TestSignTimeoutUsesOneOwnedWorker catches leaking one detached goroutine for
// every timed-out call when a provider violates the context-aware contract.
func TestSignTimeoutUsesOneOwnedWorker(t *testing.T) {
	release := make(chan struct{})
	provider := &nonCooperativeConfigSigner{keyID: strings.Repeat("k", 22), release: release}
	wrapped, err := NewTimeoutConfigSigner(provider, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("NewTimeoutConfigSigner: %v", err)
	}
	defer func() { _ = wrapped.Close() }()
	defer close(release)

	for call := range 3 {
		started := time.Now()
		if _, err := wrapped.Sign(context.Background(), []byte("message")); !errors.Is(err, ErrSignerTimeout) {
			t.Fatalf("call %d error = %v, want ErrSignerTimeout", call, err)
		}
		if elapsed := time.Since(started); elapsed < 450*time.Millisecond || elapsed > 2*time.Second {
			t.Fatalf("call %d elapsed = %s", call, elapsed)
		}
	}
	if got := provider.starts.Load(); got != 1 {
		t.Fatalf("provider starts = %d, want one owned in-flight worker", got)
	}
	if err := wrapped.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := wrapped.Sign(context.Background(), []byte("message")); !errors.Is(err, ErrClosed) {
		t.Fatalf("Sign after Close error = %v, want ErrClosed", err)
	}
}

// TestSignBundleRejectsRawProvider catches bypassing the mandatory validated
// timeout adapter at the exported configuration-signing boundary.
func TestSignBundleRejectsRawProvider(t *testing.T) {
	payload, err := DecodePayloadV1([]byte(validPayloadJSON))
	if err != nil {
		t.Fatalf("DecodePayloadV1: %v", err)
	}
	if _, err := SignBundleV1(context.Background(), payload, newTestConfigSigner(t, 0xbb)); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("raw-provider SignBundleV1 error = %v, want ErrInvalidArgument", err)
	}
}

// TestSignBundleCollapsesProviderErrors catches returning attacker- or
// provider-controlled signer details through the trust boundary.
func TestSignBundleCollapsesProviderErrors(t *testing.T) {
	payload, err := DecodePayloadV1([]byte(validPayloadJSON))
	if err != nil {
		t.Fatalf("DecodePayloadV1: %v", err)
	}
	bounded := newTestBoundedSigner(t, failingConfigSigner{keyID: strings.Repeat("k", 22)})
	_, err = SignBundleV1(context.Background(), payload, bounded)
	if !errors.Is(err, ErrSignerFailure) || strings.Contains(fmt.Sprint(err), "SIGNER_PROVIDER_SECRET_CANARY") {
		t.Fatalf("SignBundleV1 error = %v, want fixed ErrSignerFailure", err)
	}
}

// TestSignBundleEnforcesConfiguredTimeout catches the exported signing boundary
// invoking an external provider outside its validated signer budget.
func TestSignBundleEnforcesConfiguredTimeout(t *testing.T) {
	payload, err := DecodePayloadV1([]byte(validPayloadJSON))
	if err != nil {
		t.Fatalf("DecodePayloadV1: %v", err)
	}
	bounded, err := NewTimeoutConfigSigner(blockingConfigSigner{keyID: strings.Repeat("k", 22)}, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("NewTimeoutConfigSigner: %v", err)
	}
	defer func() { _ = bounded.Close() }()
	if _, err := SignBundleV1(context.Background(), payload, bounded); !errors.Is(err, ErrSignerTimeout) {
		t.Fatalf("SignBundleV1 error = %v, want ErrSignerTimeout", err)
	}
}

// TestSignTimeoutAdapterRedactsProvider catches fmt or structured logging of
// the lifecycle wrapper disclosing fields held by an external signer.
func TestSignTimeoutAdapterRedactsProvider(t *testing.T) {
	const canary = "TIMEOUT_ADAPTER_PROVIDER_SECRET_CANARY"
	bounded := newTestBoundedSigner(t, secretBearingConfigSigner{keyID: strings.Repeat("k", 22), secret: canary})
	for _, format := range []string{"%v", "%+v", "%#v", "%s", "%q", "%x"} {
		if rendered := fmt.Sprintf(format, bounded); strings.Contains(rendered, canary) {
			t.Fatalf("format %s exposed provider", format)
		}
	}
	var output bytes.Buffer
	slog.New(slog.NewJSONHandler(&output, nil)).Info("signer", "value", bounded)
	if strings.Contains(output.String(), canary) {
		t.Fatal("slog exposed timeout-adapter provider")
	}
}

// TestSignTimeoutSnapshotsValidatedKeyIDOnce catches a stateful provider
// changing its identifier between constructor validation and adapter storage.
func TestSignTimeoutSnapshotsValidatedKeyIDOnce(t *testing.T) {
	provider := &changingKeyIDSigner{}
	bounded, err := NewTimeoutConfigSigner(provider, localSignerTimeout)
	if err != nil {
		t.Fatalf("NewTimeoutConfigSigner: %v", err)
	}
	defer func() { _ = bounded.Close() }()
	if got := bounded.KeyID(); got != strings.Repeat("k", 22) {
		t.Fatalf("KeyID = %q, want validated snapshot", got)
	}
}

// TestSignLocalLifecycleCopiesShareCloseAndRedact catches secret ownership
// aliasing, use-after-close, copied-value resurrection, and fmt/slog disclosure.
func TestSignLocalLifecycleCopiesShareCloseAndRedact(t *testing.T) {
	seedCanary := []byte("CONFIG_SIGNER_SEED_CANARY_123456")
	signer, err := NewLocalConfigSigner(secret.NewBytes(seedCanary))
	if err != nil {
		t.Fatalf("NewLocalConfigSigner: %v", err)
	}
	copyValue := *signer
	if err := copyValue.Close(); err != nil {
		t.Fatalf("Close copy: %v", err)
	}
	if _, err := signer.Sign(context.Background(), []byte("message")); !errors.Is(err, ErrClosed) {
		t.Fatalf("Sign after copied Close error = %v, want ErrClosed", err)
	}
	if err := signer.Close(); err != nil {
		t.Fatalf("idempotent Close: %v", err)
	}
	var zero LocalConfigSigner
	if _, err := zero.Sign(context.Background(), []byte("message")); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero Sign error = %v, want ErrClosed", err)
	}
	for _, subject := range []any{signer, copyValue, zero} {
		for _, format := range []string{"%v", "%+v", "%#v", "%s", "%q", "%x"} {
			if rendered := fmt.Sprintf(format, subject); strings.Contains(rendered, string(seedCanary)) || strings.Contains(rendered, fmt.Sprintf("%x", seedCanary)) {
				t.Fatalf("format %s exposed seed", format)
			}
		}
	}
	var logOutput bytes.Buffer
	slog.New(slog.NewJSONHandler(&logOutput, nil)).Info("signer", "value", signer, "copy", copyValue)
	if strings.Contains(logOutput.String(), string(seedCanary)) || strings.Contains(logOutput.String(), fmt.Sprintf("%x", seedCanary)) {
		t.Fatal("slog exposed signer seed")
	}
}

// TestSignConstructorCopiesSeed catches retaining caller-owned mutable seed
// storage instead of deriving from a private exact 32-byte copy.
func TestSignConstructorCopiesSeed(t *testing.T) {
	seed := bytes.Repeat([]byte{0x44}, ed25519.SeedSize)
	signer, err := NewLocalConfigSigner(secret.NewBytes(seed))
	if err != nil {
		t.Fatalf("NewLocalConfigSigner: %v", err)
	}
	t.Cleanup(func() { _ = signer.Close() })
	wantKeyID := signer.KeyID()
	clear(seed)
	if signer.KeyID() != wantKeyID {
		t.Fatal("caller mutation changed key ID")
	}
	if _, err := signer.Sign(context.Background(), []byte("message")); err != nil {
		t.Fatalf("Sign after caller clear: %v", err)
	}
	for _, length := range []int{0, 31, 33} {
		if got, err := NewLocalConfigSigner(secret.NewBytes(make([]byte, length))); !errors.Is(err, ErrInvalidArgument) || got != nil {
			t.Fatalf("seed length %d returned signer=%v error=%v", length, got != nil, err)
		}
	}
}

// TestSignLocalCloseIsConcurrentAndFailsClosed catches unsynchronized seed
// zeroization, copied-state races, and any operation succeeding after Close wins.
func TestSignLocalCloseIsConcurrentAndFailsClosed(t *testing.T) {
	signer := newTestConfigSigner(t, 0xaa)
	const workerCount = 8
	startWorkers := make(chan struct{})
	firstSuccess := make(chan struct{})
	workerErrors := make(chan error, workerCount)
	var firstSuccessOnce sync.Once
	var workersReady sync.WaitGroup
	var workersDone sync.WaitGroup
	workersReady.Add(workerCount)
	workersDone.Add(workerCount)
	message := bytes.Repeat([]byte{'x'}, maximumCanonicalPayloadBytes)
	for range workerCount {
		go func() {
			defer workersDone.Done()
			workersReady.Done()
			<-startWorkers
			for {
				_, err := signer.Sign(context.Background(), message)
				if errors.Is(err, ErrClosed) {
					return
				}
				if err != nil {
					workerErrors <- err
					return
				}
				firstSuccessOnce.Do(func() { close(firstSuccess) })
			}
		}()
	}
	workersReady.Wait()
	close(startWorkers)
	<-firstSuccess

	const closerCount = 8
	startClosers := make(chan struct{})
	closerErrors := make(chan error, closerCount)
	var closersReady sync.WaitGroup
	var closersDone sync.WaitGroup
	closersReady.Add(closerCount)
	closersDone.Add(closerCount)
	for range closerCount {
		go func() {
			defer closersDone.Done()
			closersReady.Done()
			<-startClosers
			if err := signer.Close(); err != nil {
				closerErrors <- err
			}
		}()
	}
	closersReady.Wait()
	close(startClosers)
	closersDone.Wait()
	workersDone.Wait()
	close(closerErrors)
	close(workerErrors)
	for err := range closerErrors {
		t.Fatalf("concurrent Close: %v", err)
	}
	for err := range workerErrors {
		t.Fatalf("concurrent Sign: %v", err)
	}
	if _, err := signer.Sign(context.Background(), []byte("after-close")); !errors.Is(err, ErrClosed) {
		t.Fatalf("Sign after Close error = %v, want ErrClosed", err)
	}
}
