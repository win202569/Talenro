// Package sensitive provides domain-separated lookup and field protection.
package sensitive

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"sync"

	"talenro.local/platform/internal/secret"
)

const (
	keyBytes               = 32
	maximumDomainBytes     = 128
	maximumProtectedBytes  = 1 << 20
	lookupPrefix           = "TALENRO-LOOKUP-V1\x00"
	fieldPrefix            = "TALENRO-FIELD-V1\x00"
	encodedKeyVersionBytes = 4
)

var (
	// ErrInvalidArgument reports malformed input without retaining its value.
	ErrInvalidArgument = errors.New("sensitive: invalid argument")
	// ErrProtectionFailed collapses random-source and cryptographic failures.
	ErrProtectionFailed = errors.New("sensitive: protection failed")
	// ErrClosed reports use after the local fixture has erased its keys.
	ErrClosed = errors.New("sensitive: protector closed")
)

// EncryptedField is the frozen key-versioned storage representation.
type EncryptedField struct {
	KeyVersion uint32
	Ciphertext []byte
}

// Protector is the frozen sensitive-field boundary.
type Protector interface {
	LookupDigest(domain string, canonical []byte) [32]byte
	Encrypt(domain string, plaintext []byte) (EncryptedField, error)
	Decrypt(domain string, value EncryptedField) ([]byte, error)
}

// Local is a local/test-only AES-256-GCM and HMAC-SHA256 protector.
type Local struct {
	mu            sync.RWMutex
	lookupKey     []byte
	encryptionKey []byte
	keyVersion    uint32
	closed        bool
}

var _ Protector = (*Local)(nil)

// NewLocal copies two exact 256-bit keys from redacted containers.
func NewLocal(lookupKey, encryptionKey secret.Bytes, keyVersion uint32) (*Local, error) {
	lookupCopy := lookupKey.Copy()
	encryptionCopy := encryptionKey.Copy()
	defer clear(lookupCopy)
	defer clear(encryptionCopy)
	if len(lookupCopy) != keyBytes || len(encryptionCopy) != keyBytes || keyVersion == 0 {
		return nil, ErrInvalidArgument
	}
	return &Local{
		lookupKey:     append([]byte(nil), lookupCopy...),
		encryptionKey: append([]byte(nil), encryptionCopy...),
		keyVersion:    keyVersion,
	}, nil
}

// LookupDigest computes HMAC-SHA256 over an unambiguous domain and value.
// Invalid input and use after Close fail closed to the all-zero digest.
func (local *Local) LookupDigest(domain string, canonical []byte) [32]byte {
	if local == nil || !validDomain(domain) || len(canonical) == 0 || len(canonical) > maximumProtectedBytes {
		return [32]byte{}
	}
	canonicalCopy := append([]byte(nil), canonical...)
	defer clear(canonicalCopy)
	local.mu.RLock()
	defer local.mu.RUnlock()
	if local.closed {
		return [32]byte{}
	}
	mac := hmac.New(sha256.New, local.lookupKey)
	_, _ = mac.Write([]byte(lookupPrefix))
	_, _ = mac.Write([]byte(domain))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write(canonicalCopy)
	material := mac.Sum(nil)
	defer clear(material)
	var digest [32]byte
	copy(digest[:], material)
	return digest
}

// Encrypt returns nonce || AES-256-GCM sealed plaintext with exact versioned AAD.
func (local *Local) Encrypt(domain string, plaintext []byte) (EncryptedField, error) {
	if local == nil || !validDomain(domain) || len(plaintext) == 0 || len(plaintext) > maximumProtectedBytes {
		return EncryptedField{}, ErrInvalidArgument
	}
	plaintextCopy := append([]byte(nil), plaintext...)
	defer clear(plaintextCopy)
	local.mu.RLock()
	defer local.mu.RUnlock()
	if local.closed {
		return EncryptedField{}, ErrClosed
	}
	aead, err := newGCM(local.encryptionKey)
	if err != nil {
		return EncryptedField{}, ErrProtectionFailed
	}
	nonce := make([]byte, aead.NonceSize())
	if count, randomErr := io.ReadFull(rand.Reader, nonce); randomErr != nil || count != len(nonce) {
		clear(nonce)
		return EncryptedField{}, ErrProtectionFailed
	}
	defer clear(nonce)
	aad := fieldAAD(domain, local.keyVersion)
	defer clear(aad)
	ciphertext := make([]byte, len(nonce), len(nonce)+len(plaintextCopy)+aead.Overhead())
	copy(ciphertext, nonce)
	ciphertext = aead.Seal(ciphertext, nonce, plaintextCopy, aad)
	return EncryptedField{KeyVersion: local.keyVersion, Ciphertext: ciphertext}, nil
}

// Decrypt authenticates domain, fixed-width big-endian key version, and value.
func (local *Local) Decrypt(domain string, value EncryptedField) ([]byte, error) {
	if local == nil || !validDomain(domain) || value.KeyVersion == 0 ||
		len(value.Ciphertext) < aesGCMNonceBytes()+aesGCMOverheadBytes() ||
		len(value.Ciphertext) > aesGCMNonceBytes()+aesGCMOverheadBytes()+maximumProtectedBytes {
		return nil, ErrInvalidArgument
	}
	ciphertextCopy := append([]byte(nil), value.Ciphertext...)
	defer clear(ciphertextCopy)
	local.mu.RLock()
	defer local.mu.RUnlock()
	if local.closed {
		return nil, ErrClosed
	}
	if value.KeyVersion != local.keyVersion {
		return nil, ErrProtectionFailed
	}
	aead, err := newGCM(local.encryptionKey)
	if err != nil {
		return nil, ErrProtectionFailed
	}
	nonceBytes := aead.NonceSize()
	aad := fieldAAD(domain, value.KeyVersion)
	defer clear(aad)
	opened, err := aead.Open(nil, ciphertextCopy[:nonceBytes], ciphertextCopy[nonceBytes:], aad)
	if err != nil {
		clear(opened)
		return nil, ErrProtectionFailed
	}
	plaintext := append([]byte(nil), opened...)
	clear(opened)
	return plaintext, nil
}

// Close atomically prevents new operations and erases locally held key bytes.
// It is idempotent and is intentionally additional to the frozen Protector.
func (local *Local) Close() error {
	if local == nil {
		return nil
	}
	local.mu.Lock()
	defer local.mu.Unlock()
	if local.closed {
		return nil
	}
	clear(local.lookupKey)
	clear(local.encryptionKey)
	local.lookupKey = nil
	local.encryptionKey = nil
	local.keyVersion = 0
	local.closed = true
	return nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func fieldAAD(domain string, keyVersion uint32) []byte {
	aad := make([]byte, 0, len(fieldPrefix)+len(domain)+1+encodedKeyVersionBytes)
	aad = append(aad, fieldPrefix...)
	aad = append(aad, domain...)
	aad = append(aad, 0)
	versionStart := len(aad)
	aad = append(aad, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(aad[versionStart:], keyVersion)
	return aad
}

func validDomain(domain string) bool {
	if len(domain) == 0 || len(domain) > maximumDomainBytes {
		return false
	}
	for index := range len(domain) {
		character := domain[index]
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' {
			continue
		}
		switch character {
		case '/', '.', ':', '_', '-':
			continue
		default:
			return false
		}
	}
	return true
}

func aesGCMNonceBytes() int { return 12 }

func aesGCMOverheadBytes() int { return 16 }
