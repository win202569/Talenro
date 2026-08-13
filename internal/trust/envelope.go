package trust

import (
	"crypto/ecdh"
	"crypto/hpke"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/gowebpki/jcs"
	"talenro.local/platform/internal/securitykit"
)

const (
	// EnvelopeVersionV1 is the only public HPKE envelope format in C1.1.
	EnvelopeVersionV1 = "talenro-config-envelope/v1"
	// KEMX25519HKDFSHA256 is the fixed RFC 9180 DHKEM identifier.
	KEMX25519HKDFSHA256 = "DHKEM-X25519-HKDF-SHA256"
	// KDFHKDFSHA256 is the fixed RFC 9180 key schedule identifier.
	KDFHKDFSHA256 = "HKDF-SHA256"
	// AEADChaCha20Poly1305 is the fixed RFC 9180 AEAD identifier.
	AEADChaCha20Poly1305 = "CHACHA20-POLY1305"

	envelopeInfo             = "talenro-config-bundle/v1"
	recipientSelectorDomain  = "TALENRO-RECIPIENT-SELECTOR-V1\x00"
	maximumEnvelopeBytes     = 1 << 20
	maximumEnvelopePlaintext = 64 << 10
	plaintextLengthBytes     = 4
)

var (
	// ErrRecipientKey reports a recipient key that cannot instantiate the
	// fixed X25519 KEM.
	ErrRecipientKey = errors.New("trust: invalid recipient key")
	// ErrEnvelopeSeal collapses all HPKE and canonical-envelope failures.
	ErrEnvelopeSeal = errors.New("trust: envelope seal failed")
	// ErrEnvelopeTooLarge reports signed or final bytes outside fixed limits.
	ErrEnvelopeTooLarge = errors.New("trust: envelope too large")
	// ErrRandomSource collapses all server padding-randomness failures.
	ErrRandomSource = errors.New("trust: random source failed")
)

var plaintextBuckets = [...]int{4096, 8192, 16384, 32768, 65536}

// OuterEnvelopeV1 is the exact public HPKE distribution envelope. Its fields
// include observable ciphertext and selectors and therefore reject generic
// serialization and use fixed redacted rendering. Seal uses a private wire
// alias for the one authorized serialization path.
type OuterEnvelopeV1 struct {
	EnvelopeVersion string `json:"envelope_version"`
	KEM             string `json:"kem"`
	KDF             string `json:"kdf"`
	AEAD            string `json:"aead"`
	RecipientKeyID  string `json:"recipient_key_id"`
	BundleLocator   string `json:"bundle_locator"`
	Enc             string `json:"enc"`
	Ciphertext      string `json:"ciphertext"`
}

// Format redacts every fmt rendering of an envelope.
func (OuterEnvelopeV1) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("trust.OuterEnvelopeV1([REDACTED])"))
}

// LogValue redacts structured logging of an envelope.
func (OuterEnvelopeV1) LogValue() slog.Value {
	return slog.StringValue("trust.OuterEnvelopeV1([REDACTED])")
}

// MarshalJSON rejects generic serialization of secret-bearing envelopes.
func (OuterEnvelopeV1) MarshalJSON() ([]byte, error) { return nil, ErrInvalidArgument }

// UnmarshalJSON rejects generic parsing outside the independent verifier.
func (*OuterEnvelopeV1) UnmarshalJSON([]byte) error { return ErrInvalidArgument }

type outerEnvelopeWire struct {
	EnvelopeVersion string `json:"envelope_version"`
	KEM             string `json:"kem"`
	KDF             string `json:"kdf"`
	AEAD            string `json:"aead"`
	RecipientKeyID  string `json:"recipient_key_id"`
	BundleLocator   string `json:"bundle_locator"`
	Enc             string `json:"enc"`
	Ciphertext      string `json:"ciphertext"`
}

type outerHeaderWire struct {
	EnvelopeVersion string `json:"envelope_version"`
	KEM             string `json:"kem"`
	KDF             string `json:"kdf"`
	AEAD            string `json:"aead"`
	RecipientKeyID  string `json:"recipient_key_id"`
	BundleLocator   string `json:"bundle_locator"`
	Enc             string `json:"enc"`
}

// Seal forms one fixed-suite RFC 9180 Base-mode envelope and returns the exact
// SHA-256 digest of its final JCS bytes.
func Seal(
	random securitykit.RandomSource,
	recipient [32]byte,
	locator [32]byte,
	signed SignedBundleV1,
) ([]byte, [32]byte, error) {
	if isNilValue(random) || signed.Padding != "" {
		return nil, [32]byte{}, ErrInvalidArgument
	}
	if err := ValidateSignedBundleV1(signed); err != nil {
		return nil, [32]byte{}, err
	}
	signedBody, err := json.Marshal(signedBundleWire(signed))
	if err != nil || len(signedBody) > maximumEnvelopePlaintext {
		return nil, [32]byte{}, ErrEnvelopeTooLarge
	}
	signedJCS, err := jcs.Transform(signedBody)
	clear(signedBody)
	if err != nil || len(signedJCS) < 2 || len(signedJCS) > maximumEnvelopePlaintext ||
		len(signedJCS)+plaintextLengthBytes > maximumEnvelopePlaintext {
		clear(signedJCS)
		return nil, [32]byte{}, ErrEnvelopeTooLarge
	}
	bucket := selectPlaintextBucket(len(signedJCS) + plaintextLengthBytes)
	if bucket == 0 {
		clear(signedJCS)
		return nil, [32]byte{}, ErrEnvelopeTooLarge
	}
	plaintext := make([]byte, bucket)
	binary.BigEndian.PutUint32(plaintext[:plaintextLengthBytes], uint32(len(signedJCS))) //nolint:gosec // Signed JCS is bounded to 64 KiB above.
	copy(plaintext[plaintextLengthBytes:], signedJCS)
	clear(signedJCS)
	if _, err := io.ReadFull(random, plaintext[plaintextLengthBytes+int(binary.BigEndian.Uint32(plaintext[:plaintextLengthBytes])):]); err != nil {
		clear(plaintext)
		return nil, [32]byte{}, ErrRandomSource
	}

	publicKey, err := ecdh.X25519().NewPublicKey(recipient[:])
	if err != nil {
		clear(plaintext)
		return nil, [32]byte{}, ErrRecipientKey
	}
	hpkeKey, err := hpke.NewDHKEMPublicKey(publicKey)
	if err != nil {
		clear(plaintext)
		return nil, [32]byte{}, ErrRecipientKey
	}
	enc, sender, err := hpke.NewSender(
		hpkeKey,
		hpke.HKDFSHA256(),
		hpke.ChaCha20Poly1305(),
		[]byte(envelopeInfo),
	)
	if err != nil {
		clear(plaintext)
		return nil, [32]byte{}, ErrEnvelopeSeal
	}

	selectorMaterial := make([]byte, 0, len(recipientSelectorDomain)+len(recipient)+len(locator))
	selectorMaterial = append(selectorMaterial, recipientSelectorDomain...)
	selectorMaterial = append(selectorMaterial, recipient[:]...)
	selectorMaterial = append(selectorMaterial, locator[:]...)
	selector := sha256.Sum256(selectorMaterial)
	clear(selectorMaterial)
	header := outerHeaderWire{
		EnvelopeVersion: EnvelopeVersionV1,
		KEM:             KEMX25519HKDFSHA256,
		KDF:             KDFHKDFSHA256,
		AEAD:            AEADChaCha20Poly1305,
		RecipientKeyID:  base64.RawURLEncoding.EncodeToString(selector[:16]),
		BundleLocator:   base64.RawURLEncoding.EncodeToString(locator[:]),
		Enc:             base64.RawURLEncoding.EncodeToString(enc),
	}
	aad, err := canonicalTrustedJSON(header)
	if err != nil {
		clear(plaintext)
		return nil, [32]byte{}, ErrEnvelopeSeal
	}
	ciphertext, err := sender.Seal(aad, plaintext)
	clear(plaintext)
	clear(aad)
	if err != nil {
		clear(ciphertext)
		return nil, [32]byte{}, ErrEnvelopeSeal
	}
	envelope := outerEnvelopeWire{
		EnvelopeVersion: header.EnvelopeVersion,
		KEM:             header.KEM,
		KDF:             header.KDF,
		AEAD:            header.AEAD,
		RecipientKeyID:  header.RecipientKeyID,
		BundleLocator:   header.BundleLocator,
		Enc:             header.Enc,
		Ciphertext:      base64.RawURLEncoding.EncodeToString(ciphertext),
	}
	clear(ciphertext)
	result, err := canonicalTrustedJSON(envelope)
	if err != nil {
		return nil, [32]byte{}, ErrEnvelopeSeal
	}
	if len(result) > maximumEnvelopeBytes {
		clear(result)
		return nil, [32]byte{}, ErrEnvelopeTooLarge
	}
	digest := sha256.Sum256(result)
	return result, digest, nil
}

type signedBundleWire SignedBundleV1

func selectPlaintextBucket(length int) int {
	for _, bucket := range plaintextBuckets {
		if length <= bucket {
			return bucket
		}
	}
	return 0
}

func canonicalTrustedJSON(value any) ([]byte, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	canonical, err := jcs.Transform(body)
	clear(body)
	return canonical, err
}
