package trust

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// FuzzEnvelopeSeal keeps all attacker-controlled signed-frame fields bounded
// and asserts that failures remain finite and never echo fuzz bytes.
func FuzzEnvelopeSeal(f *testing.F) {
	const canary = "TASK15-FUZZ-SECRET-CANARY-7D3A9E"
	f.Add([]byte(`{}`), byte(0))
	f.Add([]byte(`{"unknown":true}`), byte(1))
	f.Add([]byte(canary), byte(0))
	privateKey := envelopeRecipientKey(f, 0x55)
	var recipient [32]byte
	copy(recipient[:], privateKey.PublicKey().Bytes())
	f.Fuzz(func(t *testing.T, payload []byte, selector byte) {
		if len(payload) > 70<<10 {
			return
		}
		signed := testEnvelopeSignedBundle()
		switch selector % 4 {
		case 0:
			signed.PayloadJCS = string(payload)
		case 1:
			signed.PayloadSHA256 = string(payload)
		case 2:
			signed.Signature = string(payload)
		case 3:
			signed.Padding = string(payload)
		}
		body, _, err := Seal(bytes.NewReader(bytes.Repeat([]byte{0x6a}, 64<<10)), recipient, [32]byte{1}, signed)
		if len(body) > 1<<20 {
			t.Fatalf("envelope length = %d", len(body))
		}
		if err == nil {
			return
		}
		if !errors.Is(err, ErrInvalidArgument) && !errors.Is(err, ErrInvalidPayload) &&
			!errors.Is(err, ErrRecipientKey) && !errors.Is(err, ErrEnvelopeSeal) &&
			!errors.Is(err, ErrEnvelopeTooLarge) && !errors.Is(err, ErrRandomSource) {
			t.Fatalf("non-finite error: %v", err)
		}
		if bytes.Contains(payload, []byte(canary)) && strings.Contains(err.Error(), canary) {
			t.Fatal("error echoed fuzz input")
		}
	})
}
