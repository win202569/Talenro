package stateverify

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"sort"
	"testing"
)

func groupFor(t *testing.T, p map[string]any, domain string, seeds ...byte) []any {
	msg := append([]byte(domain+"\x00"), jsonBytes(t, p)...)
	s := []any{}
	for _, seed := range seeds {
		k := testPrivate(seed)
		id := sha256.Sum256(k.Public().(ed25519.PublicKey))
		s = append(s, map[string]any{"key_id": hex.EncodeToString(id[:]), "signature": base64.RawURLEncoding.EncodeToString(ed25519.Sign(k, msg))})
	}
	sort.Slice(s, func(i, j int) bool {
		return s[i].(map[string]any)["key_id"].(string) < s[j].(map[string]any)["key_id"].(string)
	})
	return s
}
func TestVerifyRotationDistinctRoots(t *testing.T) {
	b := signedBatch(t, 1)
	p := envelopePayload(t, b.Rotations[0])
	r := p["new_root_set"].(map[string]any)
	keys := []any{testKeyWire(4), testKeyWire(5), testKeyWire(6)}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].(map[string]any)["key_id"].(string) < keys[j].(map[string]any)["key_id"].(string)
	})
	r["keys"] = keys
	r["threshold"] = "2"
	e := jsonObject(t, mint(t, "rotation", p, testPrivate(1)))
	e["new_signatures"] = groupFor(t, p, "TALENRO-NODE-STATE-ROOT-ROTATION-V1", 4, 5)
	b.Rotations[0] = jsonBytes(t, e)
	b.ActiveRootEnvelope = b.Rotations[0]
	m := envelopePayload(t, b.MetadataEnvelope)
	me := jsonObject(t, mint(t, "metadata", m, testPrivate(4)))
	me["signatures"] = groupFor(t, m, "TALENRO-NODE-STATE-METADATA-V1", 4, 5, 6)
	b.MetadataEnvelope = jsonBytes(t, me)
	assertBatch(t, b, nil)
	original := b.Rotations[0]
	e["current_signatures"], e["new_signatures"] = e["new_signatures"], e["current_signatures"]
	b.Rotations[0] = jsonBytes(t, e)
	b.ActiveRootEnvelope = b.Rotations[0]
	assertBatch(t, b, ErrSignatureMismatch)
	b.Rotations[0] = original
	b.ActiveRootEnvelope = original
	me["signatures"] = groupFor(t, m, "TALENRO-NODE-STATE-METADATA-V1", 4)
	b.MetadataEnvelope = jsonBytes(t, me)
	assertBatch(t, b, ErrSignatureMismatch)
}
