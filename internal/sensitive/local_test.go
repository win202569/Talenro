package sensitive

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"sync"
	"testing"

	"talenro.local/platform/internal/secret"
)

func newTestLocal(t *testing.T, keyVersion uint32) *Local {
	t.Helper()
	local, err := NewLocal(
		secret.NewBytes(bytes.Repeat([]byte{0x11}, 32)),
		secret.NewBytes(bytes.Repeat([]byte{0x22}, 32)),
		keyVersion,
	)
	if err != nil {
		t.Fatalf("NewLocal: %v", err)
	}
	t.Cleanup(func() { _ = local.Close() })
	return local
}

func TestLocalImplementsFrozenProtector(t *testing.T) {
	var _ Protector = newTestLocal(t, 7)
}

func TestLocalRejectsInvalidKeysAndVersion(t *testing.T) {
	tests := []struct {
		name       string
		lookupLen  int
		encryptLen int
		version    uint32
	}{
		{name: "short lookup", lookupLen: 31, encryptLen: 32, version: 1},
		{name: "long lookup", lookupLen: 33, encryptLen: 32, version: 1},
		{name: "short encryption", lookupLen: 32, encryptLen: 31, version: 1},
		{name: "long encryption", lookupLen: 32, encryptLen: 33, version: 1},
		{name: "zero version", lookupLen: 32, encryptLen: 32, version: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewLocal(
				secret.NewBytes(bytes.Repeat([]byte{1}, test.lookupLen)),
				secret.NewBytes(bytes.Repeat([]byte{2}, test.encryptLen)),
				test.version,
			)
			if !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("error = %v, want ErrInvalidArgument", err)
			}
		})
	}
}

func TestLocalKeyVersionMatchesPostgresIntegerRange(t *testing.T) {
	lookup := secret.NewBytes(bytes.Repeat([]byte{0x11}, 32))
	encryption := secret.NewBytes(bytes.Repeat([]byte{0x22}, 32))
	for _, version := range []uint32{1, math.MaxInt32} {
		local, err := NewLocal(lookup, encryption, version)
		if err != nil {
			t.Fatalf("NewLocal version %d: %v", version, err)
		}
		value, err := local.Encrypt("identity/email/v1", []byte("value"))
		if err != nil {
			t.Fatalf("Encrypt version %d: %v", version, err)
		}
		opened, err := local.Decrypt("identity/email/v1", value)
		if err != nil || string(opened) != "value" {
			t.Fatalf("roundtrip version %d = %q, %v", version, opened, err)
		}
		if version == math.MaxInt32 {
			block, cipherErr := aes.NewCipher(bytes.Repeat([]byte{0x22}, 32))
			if cipherErr != nil {
				t.Fatalf("aes.NewCipher: %v", cipherErr)
			}
			aead, gcmErr := cipher.NewGCM(block)
			if gcmErr != nil {
				t.Fatalf("cipher.NewGCM: %v", gcmErr)
			}
			aad := append([]byte("TALENRO-FIELD-V1\x00identity/email/v1\x00"), 0x7f, 0xff, 0xff, 0xff)
			independent, openErr := aead.Open(nil, value.Ciphertext[:12], value.Ciphertext[12:], aad)
			if openErr != nil || string(independent) != "value" {
				t.Fatalf("independent max-version open = %q, %v", independent, openErr)
			}
		}
		_ = local.Close()
	}
	for _, version := range []uint32{math.MaxInt32 + 1, math.MaxUint32} {
		local, err := NewLocal(lookup, encryption, version)
		if !errors.Is(err, ErrInvalidArgument) || local != nil {
			t.Fatalf("NewLocal version %d returnedLocal=%v error=%v, want false ErrInvalidArgument", version, local != nil, err)
		}
	}
}

func TestLocalRedactsFormattingAndStructuredLogging(t *testing.T) {
	lookupCanary := []byte("LOOKUP_SECRET_CANARY_0123456789A")
	encryptionCanary := []byte("ENCRYPT_SECRET_CANARY_0123456789")
	const domainCanary = "identity/domain-canary/v1"
	const plaintextCanary = "PLAINTEXT_SECRET_CANARY"
	local, err := NewLocal(secret.NewBytes(lookupCanary), secret.NewBytes(encryptionCanary), 1)
	if err != nil {
		t.Fatalf("NewLocal: %v", err)
	}
	t.Cleanup(func() { _ = local.Close() })
	if _, err := local.Encrypt(domainCanary, []byte(plaintextCanary)); err != nil {
		t.Fatalf("Encrypt canary: %v", err)
	}

	localValue := *local
	var nilLocal *Local
	subjects := []struct {
		name  string
		value any
	}{
		{name: "pointer", value: local},
		{name: "value", value: localValue},
		{name: "nil pointer", value: nilLocal},
		{name: "zero value", value: Local{}},
	}
	formats := []string{"%v", "%+v", "%#v", "%s", "%q", "%x", "%X"}
	forbidden := sensitiveRepresentations(lookupCanary, encryptionCanary)
	forbidden = append(forbidden, domainCanary, plaintextCanary)
	for _, subject := range subjects {
		for _, format := range formats {
			rendered := fmt.Sprintf(format, subject.value)
			assertNoSensitiveRepresentation(t, subject.name+" "+format, rendered, forbidden)
		}
	}

	handlers := []struct {
		name string
		new  func(*bytes.Buffer) slog.Handler
	}{
		{name: "text", new: func(buffer *bytes.Buffer) slog.Handler {
			return slog.NewTextHandler(buffer, nil)
		}},
		{name: "json", new: func(buffer *bytes.Buffer) slog.Handler {
			return slog.NewJSONHandler(buffer, nil)
		}},
	}
	for _, handler := range handlers {
		var output bytes.Buffer
		logger := slog.New(handler.new(&output))
		logger.Info("protector state",
			"pointer", local,
			"value", localValue,
			"nil", nilLocal,
			"zero", Local{},
		)
		assertNoSensitiveRepresentation(t, handler.name+" slog handler", output.String(), forbidden)
	}
}

func TestLocalValueCopiesShareCloseStateAndZeroValueFailsClosed(t *testing.T) {
	local := newTestLocal(t, 1)
	protected, err := local.Encrypt("identity/email/v1", []byte("before-close"))
	if err != nil {
		t.Fatalf("Encrypt before copy: %v", err)
	}
	localValue := *local
	if err := localValue.Close(); err != nil {
		t.Fatalf("Close copied value: %v", err)
	}
	if _, err := local.Encrypt("identity/email/v1", []byte("after-copy-close")); !errors.Is(err, ErrClosed) {
		t.Fatalf("original Encrypt after copied Close error = %v, want ErrClosed", err)
	}
	if err := local.Close(); err != nil {
		t.Fatalf("idempotent original Close: %v", err)
	}

	var zero Local
	if digest := zero.LookupDigest("identity/email/v1", []byte("value")); digest != [32]byte{} {
		t.Fatalf("zero-value digest = %x", digest)
	}
	if _, err := zero.Encrypt("identity/email/v1", []byte("value")); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero-value Encrypt error = %v, want ErrClosed", err)
	}
	if _, err := zero.Decrypt("identity/email/v1", protected); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero-value Decrypt error = %v, want ErrClosed", err)
	}
	if err := zero.Close(); err != nil {
		t.Fatalf("zero-value Close: %v", err)
	}
}

func sensitiveRepresentations(values ...[]byte) []string {
	formats := []string{"%v", "%+v", "%#v", "%s", "%q", "%x", "%X"}
	representations := make([]string, 0, len(values)*len(formats))
	for _, value := range values {
		for _, format := range formats {
			representations = append(representations, fmt.Sprintf(format, value))
		}
	}
	return representations
}

func assertNoSensitiveRepresentation(t *testing.T, name, rendered string, forbidden []string) {
	t.Helper()
	for _, representation := range forbidden {
		if representation != "" && strings.Contains(rendered, representation) {
			t.Fatalf("%s exposed sensitive representation", name)
		}
	}
}

func TestLocalLookupDigestUsesExactDomainSeparation(t *testing.T) {
	local := newTestLocal(t, 7)
	canonical := []byte("Mi.Xed+tag@example.com")
	got := local.LookupDigest("identity/email/v1", canonical)

	mac := hmac.New(sha256.New, bytes.Repeat([]byte{0x11}, 32))
	_, _ = mac.Write([]byte("TALENRO-LOOKUP-V1\x00identity/email/v1\x00"))
	_, _ = mac.Write(canonical)
	var want [32]byte
	copy(want[:], mac.Sum(nil))
	if got != want {
		t.Fatalf("digest = %x, want %x", got, want)
	}
	if got == local.LookupDigest("identity/password-reset/v1", canonical) {
		t.Fatal("distinct lookup domains produced the same digest")
	}
	if zero := local.LookupDigest("identity/email/v1\x00other", canonical); zero != [32]byte{} {
		t.Fatalf("invalid domain digest = %x, want zero", zero)
	}
}

func TestLocalEncryptDecryptUsesExactAADAndDefensiveCopies(t *testing.T) {
	local := newTestLocal(t, 0x01020304)
	plaintext := []byte("sensitive-value")
	value, err := local.Encrypt("identity/email/v1", plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	plaintext[0] = 'X'
	if value.KeyVersion != 0x01020304 {
		t.Fatalf("key version = %d", value.KeyVersion)
	}
	if len(value.Ciphertext) != 12+len("sensitive-value")+16 {
		t.Fatalf("ciphertext length = %d", len(value.Ciphertext))
	}

	block, err := aes.NewCipher(bytes.Repeat([]byte{0x22}, 32))
	if err != nil {
		t.Fatalf("aes.NewCipher: %v", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("cipher.NewGCM: %v", err)
	}
	version := make([]byte, 4)
	binary.BigEndian.PutUint32(version, 0x01020304)
	aad := append([]byte("TALENRO-FIELD-V1\x00identity/email/v1\x00"), version...)
	opened, err := aead.Open(nil, value.Ciphertext[:12], value.Ciphertext[12:], aad)
	if err != nil {
		t.Fatalf("independent AEAD open: %v", err)
	}
	if string(opened) != "sensitive-value" {
		t.Fatalf("opened plaintext = %q", opened)
	}

	copyForDecrypt := EncryptedField{KeyVersion: value.KeyVersion, Ciphertext: append([]byte(nil), value.Ciphertext...)}
	decrypted, err := local.Decrypt("identity/email/v1", copyForDecrypt)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	decrypted[0] = 'X'
	decryptedAgain, err := local.Decrypt("identity/email/v1", copyForDecrypt)
	if err != nil {
		t.Fatalf("Decrypt again: %v", err)
	}
	if string(decryptedAgain) != "sensitive-value" {
		t.Fatalf("second plaintext = %q", decryptedAgain)
	}
}

func TestLocalRejectsSubstitutionTamperAndBounds(t *testing.T) {
	local := newTestLocal(t, 7)
	value, err := local.Encrypt("identity/email/v1", []byte("value"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	tampered := EncryptedField{KeyVersion: value.KeyVersion, Ciphertext: append([]byte(nil), value.Ciphertext...)}
	tampered.Ciphertext[len(tampered.Ciphertext)-1] ^= 1
	tests := []struct {
		name   string
		domain string
		value  EncryptedField
	}{
		{name: "domain substitution", domain: "identity/totp/v1", value: value},
		{name: "key version substitution", domain: "identity/email/v1", value: EncryptedField{KeyVersion: 8, Ciphertext: value.Ciphertext}},
		{name: "tamper", domain: "identity/email/v1", value: tampered},
		{name: "short ciphertext", domain: "identity/email/v1", value: EncryptedField{KeyVersion: 7, Ciphertext: make([]byte, 28)}},
		{name: "oversize ciphertext", domain: "identity/email/v1", value: EncryptedField{KeyVersion: 7, Ciphertext: make([]byte, 12+16+1048578)}},
		{name: "empty domain", domain: "", value: value},
		{name: "oversize domain", domain: string(bytes.Repeat([]byte{'a'}, 129)), value: value},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := local.Decrypt(test.domain, test.value); !errors.Is(err, ErrProtectionFailed) && !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("error = %v", err)
			}
		})
	}
	for _, plaintext := range [][]byte{nil, {}, make([]byte, 1048578)} {
		if _, err := local.Encrypt("identity/email/v1", plaintext); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("Encrypt length %d error = %v", len(plaintext), err)
		}
	}
	if digest := local.LookupDigest("identity/email/v1", make([]byte, 1048577)); digest != [32]byte{} {
		t.Fatalf("oversize canonical digest = %x", digest)
	}
}

func TestLocalAllowsMaximumIdempotencyResponseFrameOnly(t *testing.T) {
	t.Parallel()

	local := newTestLocal(t, 7)
	maximumFrame := bytes.Repeat([]byte{'x'}, (1<<20)+1)
	protected, err := local.Encrypt("idempotency/response/v1", maximumFrame)
	if err != nil {
		t.Fatalf("Encrypt maximum idempotency frame: %v", err)
	}
	opened, err := local.Decrypt("idempotency/response/v1", protected)
	if err != nil {
		t.Fatalf("Decrypt maximum idempotency frame: %v", err)
	}
	if !bytes.Equal(opened, maximumFrame) {
		t.Fatal("maximum idempotency frame did not round trip")
	}
	if _, err := local.Encrypt("idempotency/response/v1", make([]byte, (1<<20)+2)); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("Encrypt oversized frame error = %v, want ErrInvalidArgument", err)
	}
}

func TestLocalDecryptRejectsAuthenticatedEmptyPlaintext(t *testing.T) {
	local := newTestLocal(t, 7)
	block, err := aes.NewCipher(bytes.Repeat([]byte{0x22}, 32))
	if err != nil {
		t.Fatalf("aes.NewCipher: %v", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("cipher.NewGCM: %v", err)
	}
	nonce := make([]byte, 12)
	aad := append([]byte("TALENRO-FIELD-V1\x00identity/email/v1\x00"), 0, 0, 0, 7)
	ciphertext := append(append([]byte(nil), nonce...), aead.Seal(nil, nonce, nil, aad)...)
	if len(ciphertext) != 28 {
		t.Fatalf("fixture ciphertext length = %d", len(ciphertext))
	}
	_, err = local.Decrypt("identity/email/v1", EncryptedField{KeyVersion: 7, Ciphertext: ciphertext})
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("Decrypt authenticated empty plaintext error = %v, want ErrInvalidArgument", err)
	}
}

func TestLocalUsesFreshNonce(t *testing.T) {
	local := newTestLocal(t, 1)
	first, err := local.Encrypt("identity/email/v1", []byte("same"))
	if err != nil {
		t.Fatalf("first Encrypt: %v", err)
	}
	second, err := local.Encrypt("identity/email/v1", []byte("same"))
	if err != nil {
		t.Fatalf("second Encrypt: %v", err)
	}
	if bytes.Equal(first.Ciphertext, second.Ciphertext) {
		t.Fatal("encryption reused nonce")
	}
}

func TestLocalCloseIsConcurrentAndFailsClosed(t *testing.T) {
	local := newTestLocal(t, 1)
	protected, err := local.Encrypt("identity/email/v1", []byte("before-close"))
	if err != nil {
		t.Fatalf("Encrypt before close: %v", err)
	}

	const workerCount = 8
	startWorkers := make(chan struct{})
	firstSuccess := make(chan struct{})
	workerErrors := make(chan error, workerCount)
	var firstSuccessOnce sync.Once
	var workersReady sync.WaitGroup
	var workersDone sync.WaitGroup
	workersReady.Add(workerCount)
	workersDone.Add(workerCount)
	largePlaintext := bytes.Repeat([]byte{'x'}, 1<<20)
	for range workerCount {
		go func() {
			defer workersDone.Done()
			workersReady.Done()
			<-startWorkers
			for {
				_, encryptErr := local.Encrypt("identity/email/v1", largePlaintext)
				if errors.Is(encryptErr, ErrClosed) {
					return
				}
				if encryptErr != nil {
					workerErrors <- encryptErr
					return
				}
				firstSuccessOnce.Do(func() { close(firstSuccess) })
				_ = local.LookupDigest("identity/email/v1", largePlaintext)
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
			if closeErr := local.Close(); closeErr != nil {
				closerErrors <- closeErr
			}
		}()
	}
	closersReady.Wait()
	close(startClosers)
	closersDone.Wait()
	workersDone.Wait()
	close(closerErrors)
	close(workerErrors)
	for closeErr := range closerErrors {
		t.Fatalf("concurrent Close: %v", closeErr)
	}
	for workerErr := range workerErrors {
		t.Fatalf("concurrent operation: %v", workerErr)
	}
	if err := local.Close(); err != nil {
		t.Fatalf("idempotent Close: %v", err)
	}
	if _, err := local.Encrypt("identity/email/v1", []byte("after-close")); !errors.Is(err, ErrClosed) {
		t.Fatalf("Encrypt after close error = %v", err)
	}
	if _, err := local.Decrypt("identity/email/v1", protected); !errors.Is(err, ErrClosed) {
		t.Fatalf("Decrypt after close error = %v", err)
	}
	if digest := local.LookupDigest("identity/email/v1", []byte("after-close")); digest != [32]byte{} {
		t.Fatalf("digest after close = %x", digest)
	}
}
