package trust

import (
	"bytes"
	"crypto/ecdh"
	"crypto/hpke"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/gowebpki/jcs"
)

type rfc9180Vector struct {
	Mode        int    `json:"mode"`
	KEMID       int    `json:"kem_id"`
	KDFID       int    `json:"kdf_id"`
	AEADID      int    `json:"aead_id"`
	Info        string `json:"info"`
	RecipientSK string `json:"skRm"`
	Enc         string `json:"enc"`
	Encryptions []struct {
		AAD        string `json:"aad"`
		Ciphertext string `json:"ct"`
		Plaintext  string `json:"pt"`
	} `json:"encryptions"`
}

// TestHPKERFC9180BaseModeRecipientOpen catches suite drift, hex/base64
// confusion, or a client that cannot consume the official RFC 9180 result.
func TestHPKERFC9180BaseModeRecipientOpen(t *testing.T) {
	body, err := os.ReadFile("../../testdata/crypto/rfc9180/base-x25519-hkdf-sha256-chacha20poly1305.json")
	if err != nil {
		t.Fatalf("read vector: %v", err)
	}
	if got := sha256.Sum256(body); fmt.Sprintf("%x", got) != "6d0810aeccf32aa5084c90093776da3c34bcf12fc25e98ae19e76d31baa27547" {
		t.Fatalf("vector SHA-256 = %x", got)
	}
	var vector rfc9180Vector
	if err := json.Unmarshal(body, &vector); err != nil {
		t.Fatalf("decode vector: %v", err)
	}
	if vector.Mode != 0 || vector.KEMID != 32 || vector.KDFID != 1 || vector.AEADID != 3 || len(vector.Encryptions) != 1 {
		t.Fatal("unexpected vector suite or case count")
	}
	privateBytes := decodeHexFixture(t, vector.RecipientSK)
	privateKey, err := ecdh.X25519().NewPrivateKey(privateBytes)
	clear(privateBytes)
	if err != nil {
		t.Fatalf("parse recipient private key: %v", err)
	}
	hpkeKey, err := hpke.NewDHKEMPrivateKey(privateKey)
	if err != nil {
		t.Fatalf("adapt recipient private key: %v", err)
	}
	recipient, err := hpke.NewRecipient(
		decodeHexFixture(t, vector.Enc), hpkeKey, hpke.HKDFSHA256(), hpke.ChaCha20Poly1305(), decodeHexFixture(t, vector.Info),
	)
	if err != nil {
		t.Fatalf("new recipient: %v", err)
	}
	opened, err := recipient.Open(decodeHexFixture(t, vector.Encryptions[0].AAD), decodeHexFixture(t, vector.Encryptions[0].Ciphertext))
	if err != nil {
		t.Fatalf("open vector: %v", err)
	}
	want := decodeHexFixture(t, vector.Encryptions[0].Plaintext)
	if !bytes.Equal(opened, want) {
		t.Fatalf("plaintext = %x, want %x", opened, want)
	}
	clear(opened)
	clear(want)
}

// TestEnvelopeSealUsesExactSuiteAADSelectorAndBucket catches using the wrong
// HPKE suite, authenticating ciphertext instead of the header, stable device
// IDs in the selector, or variable-size plaintext.
func TestEnvelopeSealUsesExactSuiteAADSelectorAndBucket(t *testing.T) {
	privateKey := envelopeRecipientKey(t, 0x41)
	var recipient [32]byte
	copy(recipient[:], privateKey.PublicKey().Bytes())
	locator := [32]byte{0x71, 0x72, 0x73}
	signed := testEnvelopeSignedBundle()

	envelopeBytes, digest, err := Seal(bytes.NewReader(bytes.Repeat([]byte{0xa5}, 64<<10)), recipient, locator, signed)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	if got := sha256.Sum256(envelopeBytes); got != digest {
		t.Fatalf("distribution digest = %x, want %x", digest, got)
	}
	if len(envelopeBytes) > 1<<20 {
		t.Fatalf("envelope length = %d", len(envelopeBytes))
	}

	var envelope envelopeTestWire
	if err := json.Unmarshal(envelopeBytes, &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	canonical, err := jcs.Transform(envelopeBytes)
	if err != nil || !bytes.Equal(canonical, envelopeBytes) {
		t.Fatal("envelope is not exact JCS")
	}
	if envelope.EnvelopeVersion != EnvelopeVersionV1 || envelope.KEM != KEMX25519HKDFSHA256 ||
		envelope.KDF != KDFHKDFSHA256 || envelope.AEAD != AEADChaCha20Poly1305 {
		t.Fatalf("unexpected public suite: %#v", envelope)
	}
	selectorInput := append([]byte("TALENRO-RECIPIENT-SELECTOR-V1\x00"), recipient[:]...)
	selectorInput = append(selectorInput, locator[:]...)
	wantSelector := sha256.Sum256(selectorInput)
	clear(selectorInput)
	if envelope.RecipientKeyID != base64.RawURLEncoding.EncodeToString(wantSelector[:16]) {
		t.Fatal("recipient selector is not the exact domain-separated 128-bit value")
	}
	if envelope.BundleLocator != base64.RawURLEncoding.EncodeToString(locator[:]) {
		t.Fatal("locator encoding drifted")
	}

	enc, err := base64.RawURLEncoding.DecodeString(envelope.Enc)
	if err != nil || len(enc) != 32 {
		t.Fatalf("enc length/encoding invalid")
	}
	ciphertext, err := base64.RawURLEncoding.DecodeString(envelope.Ciphertext)
	if err != nil || len(ciphertext) != 4096+16 {
		t.Fatalf("ciphertext length = %d, want 4112", len(ciphertext))
	}
	headerBody, err := json.Marshal(envelope.header())
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}
	aad, err := jcs.Transform(headerBody)
	if err != nil {
		t.Fatalf("canonicalize header: %v", err)
	}
	hpkeKey, err := hpke.NewDHKEMPrivateKey(privateKey)
	if err != nil {
		t.Fatalf("adapt private key: %v", err)
	}
	opener, err := hpke.NewRecipient(enc, hpkeKey, hpke.HKDFSHA256(), hpke.ChaCha20Poly1305(), []byte("talenro-config-bundle/v1"))
	if err != nil {
		t.Fatalf("new recipient: %v", err)
	}
	plaintext, err := opener.Open(aad, ciphertext)
	if err != nil {
		t.Fatalf("open envelope: %v", err)
	}
	defer clear(plaintext)
	if len(plaintext) != 4096 {
		t.Fatalf("plaintext length = %d, want fixed bucket", len(plaintext))
	}
	signedLength := int(binary.BigEndian.Uint32(plaintext[:4]))
	if signedLength < 2 || signedLength > len(plaintext)-4 {
		t.Fatalf("signed frame length = %d", signedLength)
	}
	wantJSON, err := json.Marshal(signed)
	if err != nil {
		t.Fatalf("marshal signed bundle: %v", err)
	}
	wantJSON, err = jcs.Transform(wantJSON)
	if err != nil {
		t.Fatalf("canonicalize signed bundle: %v", err)
	}
	if !bytes.Equal(plaintext[4:4+signedLength], wantJSON) {
		t.Fatal("framed signed bundle differs")
	}
}

// TestEnvelopeRejectsOversizeBeforeRandomnessOrHPKE catches allocating a
// maximum-bucket plaintext or invoking cryptography for an already-invalid
// signed input.
func TestEnvelopeRejectsOversizeBeforeRandomnessOrHPKE(t *testing.T) {
	privateKey := envelopeRecipientKey(t, 0x42)
	var recipient [32]byte
	copy(recipient[:], privateKey.PublicKey().Bytes())
	signed := testEnvelopeSignedBundle()
	signed.PayloadJCS = strings.Repeat("x", 64<<10)
	random := &countingRandomSource{}
	if _, _, err := Seal(random, recipient, [32]byte{1}, signed); !errors.Is(err, ErrEnvelopeTooLarge) {
		t.Fatalf("Seal oversize error = %v", err)
	}
	if random.calls != 0 {
		t.Fatalf("random calls = %d, want zero", random.calls)
	}
}

// TestEnvelopeSensitiveSurfaceIsRedacted catches accidental ciphertext,
// selector, or locator disclosure through fmt, slog, or generic JSON.
func TestEnvelopeSensitiveSurfaceIsRedacted(t *testing.T) {
	const canary = "TASK15-ENVELOPE-CANARY"
	value := OuterEnvelopeV1{RecipientKeyID: canary, BundleLocator: canary, Ciphertext: canary}
	formatted := fmt.Sprintf("%v %#v %+v", value, value, value)
	logged := value.LogValue().String()
	if strings.Contains(formatted, canary) || strings.Contains(logged, canary) {
		t.Fatal("secret-bearing envelope rendered a canary")
	}
	if _, err := json.Marshal(value); err == nil {
		t.Fatal("generic JSON serialization accepted an envelope")
	}
	_ = slog.Any("envelope", value)
}

func decodeHexFixture(t *testing.T, value string) []byte {
	t.Helper()
	decoded, err := hex.DecodeString(value)
	if err != nil {
		t.Fatalf("decode fixture hex: %v", err)
	}
	return decoded
}

type envelopeTestingTB interface {
	Helper()
	Fatalf(string, ...any)
}

func envelopeRecipientKey(t envelopeTestingTB, fill byte) *ecdh.PrivateKey {
	t.Helper()
	privateKey, err := ecdh.X25519().NewPrivateKey(bytes.Repeat([]byte{fill}, 32))
	if err != nil {
		t.Fatalf("new recipient key: %v", err)
	}
	return privateKey
}

func testEnvelopeSignedBundle() SignedBundleV1 {
	return SignedBundleV1{
		PayloadJCS:    `{}`,
		PayloadSHA256: base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x12}, 32)),
		SignerKeyID:   base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x13}, 16)),
		Algorithm:     SignatureAlgorithm,
		Signature:     base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x14}, 64)),
		Padding:       "",
	}
}

type envelopeTestWire struct {
	EnvelopeVersion string `json:"envelope_version"`
	KEM             string `json:"kem"`
	KDF             string `json:"kdf"`
	AEAD            string `json:"aead"`
	RecipientKeyID  string `json:"recipient_key_id"`
	BundleLocator   string `json:"bundle_locator"`
	Enc             string `json:"enc"`
	Ciphertext      string `json:"ciphertext"`
}

type envelopeHeaderTestWire struct {
	EnvelopeVersion string `json:"envelope_version"`
	KEM             string `json:"kem"`
	KDF             string `json:"kdf"`
	AEAD            string `json:"aead"`
	RecipientKeyID  string `json:"recipient_key_id"`
	BundleLocator   string `json:"bundle_locator"`
	Enc             string `json:"enc"`
}

func (value envelopeTestWire) header() envelopeHeaderTestWire {
	return envelopeHeaderTestWire{
		EnvelopeVersion: value.EnvelopeVersion, KEM: value.KEM, KDF: value.KDF, AEAD: value.AEAD,
		RecipientKeyID: value.RecipientKeyID, BundleLocator: value.BundleLocator, Enc: value.Enc,
	}
}

type countingRandomSource struct{ calls int }

func (source *countingRandomSource) Read(value []byte) (int, error) {
	source.calls++
	clear(value)
	return len(value), nil
}
