package sensitive

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
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
		{name: "oversize ciphertext", domain: "identity/email/v1", value: EncryptedField{KeyVersion: 7, Ciphertext: make([]byte, 12+16+1048577)}},
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
	for _, plaintext := range [][]byte{nil, {}, make([]byte, 1048577)} {
		if _, err := local.Encrypt("identity/email/v1", plaintext); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("Encrypt length %d error = %v", len(plaintext), err)
		}
	}
	if digest := local.LookupDigest("identity/email/v1", make([]byte, 1048577)); digest != [32]byte{} {
		t.Fatalf("oversize canonical digest = %x", digest)
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
	var wait sync.WaitGroup
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 50 {
				_, _ = local.Encrypt("identity/email/v1", []byte("concurrent"))
				_ = local.LookupDigest("identity/email/v1", []byte("concurrent"))
			}
		}()
	}
	if err := local.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	wait.Wait()
	if _, err := local.Encrypt("identity/email/v1", []byte("after-close")); !errors.Is(err, ErrClosed) {
		t.Fatalf("Encrypt after close error = %v", err)
	}
	if digest := local.LookupDigest("identity/email/v1", []byte("after-close")); digest != [32]byte{} {
		t.Fatalf("digest after close = %x", digest)
	}
}
