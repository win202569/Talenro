package stateformat

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

const timePayload = `{"control_plane_authority_epoch":"19","issued_at":"2026-10-09T00:00:00Z","node_authority_checkpoint_sequence":"7","node_id":"11111111-1111-4111-8111-111111111111","request_nonce":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","schema_version":"node-time-attestation.v1","valid_until":"2026-10-09T00:05:00Z"}`

func timeWrap(t *testing.T, p []byte) []byte {
	return encoded(t, map[string]any{"schema_version": "talenro-node-time-attestation/v1", "payload_b64url": base64.RawURLEncoding.EncodeToString(p), "payload_sha256": fmt.Sprintf("%x", sha256.Sum256(p)), "key_id": strings.Repeat("ab", 32), "signature": base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{1}, 64))})
}

// A missing Kind branch, relaxed schema, or shared creation-sequence field breaks these tests.
func TestTimeAttestationFormat(t *testing.T) {
	a, e := ParseEnvelope(Kind(5), timeWrap(t, []byte(timePayload)))
	if e != nil {
		t.Fatal(e)
	}
	defer a.Clear()
	if a.Bindings().Kind != Kind(5) || a.Bindings().Sequence != 0 {
		t.Fatal("wrong kind/sequence")
	}
	parse := func(p []byte) (artifactBindings, error) {
		a, e := ParseEnvelope(Kind(5), timeWrap(t, p))
		if e != nil {
			return artifactBindings{}, e
		}
		defer a.Clear()
		return a.Bindings(), nil
	}
	assertClosedObjects(t, object(t, []byte(timePayload)), parse)
	for _, p := range []string{
		strings.Replace(timePayload, `"7"`, `7`, 1), strings.Replace(timePayload, `"7"`, `"0"`, 1), strings.Replace(timePayload, `"7"`, `"9223372036854775808"`, 1),
		strings.Replace(timePayload, "00:05:00Z", "00:05:01Z", 1), strings.Replace(timePayload, "00:05:00Z", "00:00:00Z", 1),
		strings.Replace(timePayload, "node-time-attestation.v1", "node-desired-state.v1", 1),
		strings.Replace(timePayload, `"issued_at":`, `"issued_at":"2026-10-09T00:00:00Z","issued_at":`, 1),
		strings.Replace(timePayload, `"issued_at":`, `"\u0069ssued_at":"2026-10-09T00:00:00Z","issued_at":`, 1), timePayload + " {}", " " + timePayload,
		strings.Replace(timePayload, "00:00:00Z", "00:00:00.0Z", 1), string([]byte{0xff}),
	} {
		if a, e := ParseEnvelope(Kind(5), timeWrap(t, []byte(p))); e == nil || a != nil {
			t.Fatal("accepted malformed", p)
		}
	}
	raw := timeWrap(t, []byte(timePayload))
	for _, k := range []Kind{Invalid, Desired, InitialRoot, RootRotation, Metadata} {
		if a, e := ParseEnvelope(k, raw); e == nil || a != nil {
			t.Fatal("mixed kind")
		}
	}
	env := object(t, raw)
	env["payload_sha256"] = strings.Repeat("cd", 32)
	if a, e := ParseEnvelope(Kind(5), encoded(t, env)); e == nil || a != nil {
		t.Fatal("digest")
	}
	for _, n := range []int{65537, 262145} {
		p := []byte(`"` + strings.Repeat("a", n-2) + `"`)
		var e error
		if n == 65537 {
			_, e = ParseEnvelope(Kind(5), timeWrap(t, p))
		} else {
			_, e = ParseEnvelope(Kind(5), p)
		}
		if e == nil {
			t.Fatal("size")
		}
	}
}
func TestTimeAttestationNonce(t *testing.T) {
	for _, n := range []string{strings.Repeat("A", 43), base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{255}, 32)), strings.Repeat("A", 42), strings.Repeat("A", 44), strings.Repeat("A", 43) + "=", strings.Repeat("A", 42) + "B"} {
		p := strings.Replace(timePayload, strings.Repeat("A", 43), n, 1)
		a, e := ParseEnvelope(Kind(5), timeWrap(t, []byte(p)))
		want := n == strings.Repeat("A", 43) || n == base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{255}, 32))
		if (e == nil) != want {
			t.Fatalf("nonce %s %v", n, e)
		}
		a.Clear()
	}
}
func TestTimeAttestationVector(t *testing.T) {
	a, e := ParseEnvelope(Kind(5), timeWrap(t, []byte(timePayload)))
	if e != nil {
		t.Fatal(e)
	}
	defer a.Clear()
	msg, e := a.Transcript()
	if e != nil || !bytes.Equal(msg, append([]byte("TALENRO-NODE-TIME-ATTESTATION-V1\x00"), []byte(timePayload)...)) {
		t.Fatal("domain transcript")
	}
	if a.PayloadDigest() != sha256.Sum256([]byte(timePayload)) {
		t.Fatal("payload digest")
	}
	var v struct {
		Payload      string
		Envelope     string
		EnvelopeHash string `json:"envelope_sha256"`
		Hash         string `json:"payload_sha256"`
		Transcript   string `json:"transcript_hex"`
	}
	if e := json.Unmarshal(fixture(t, "time-attestation-v1"), &v); e != nil {
		t.Fatal(e)
	}
	d := a.PayloadDigest()
	ed := a.EnvelopeDigest()
	if v.Envelope != string(a.Envelope()) || v.EnvelopeHash != hex.EncodeToString(ed[:]) {
		t.Fatal("frozen envelope")
	}
	if v.Payload != timePayload || v.Hash != hex.EncodeToString(d[:]) || v.Transcript != hex.EncodeToString(msg) || a.Bindings().CheckpointSequence != 7 || a.Bindings().RequestNonce != ([32]byte{}) {
		t.Fatal("frozen vector")
	}
	clear(msg)
}
func TestTimeAttestationOwnership(t *testing.T) {
	raw := timeWrap(t, []byte(timePayload))
	a, e := ParseEnvelope(Kind(5), raw)
	if e != nil {
		t.Fatal(e)
	}
	raw[0] ^= 1
	p := a.Payload()
	p[0] ^= 1
	if string(a.Payload()) != timePayload {
		t.Fatal("alias")
	}
	a.Clear()
	a.Clear()
	if a.Bindings() != (Bindings{}) || len(a.Payload()) != 0 {
		t.Fatal("clear")
	}
}
