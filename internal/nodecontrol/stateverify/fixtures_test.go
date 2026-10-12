package stateverify

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"sort"
	"strconv"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"testing"
)

func frozenFile(t *testing.T, name string) []byte {
	t.Helper()
	b, e := os.ReadFile("testdata/" + name)
	if e != nil {
		t.Fatal(e)
	}
	return b
}

func jsonObject(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var p map[string]any
	if e := json.Unmarshal(b, &p); e != nil {
		t.Fatal(e)
	}
	return p
}
func jsonBytes(t *testing.T, p any) []byte {
	t.Helper()
	b, e := json.Marshal(p)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func envelopePayload(t *testing.T, b []byte) map[string]any {
	t.Helper()
	e := jsonObject(t, b)
	p, err := base64.RawURLEncoding.DecodeString(e["payload_b64url"].(string))
	if err != nil {
		t.Fatal(err)
	}
	return jsonObject(t, p)
}
func testKeyWire(n byte) map[string]any {
	k := testPrivate(n)
	pub := k.Public().(ed25519.PublicKey)
	id := sha256.Sum256(pub)
	return map[string]any{"algorithm": "ed25519", "key_id": hex.EncodeToString(id[:]), "public_key": base64.RawURLEncoding.EncodeToString(pub)}
}
func mint(t *testing.T, name string, p map[string]any, signer ed25519.PrivateKey) []byte {
	t.Helper()
	raw := jsonBytes(t, p)
	hash := sha256.Sum256(raw)
	e := jsonObject(t, frozenFile(t, name+".envelope.json"))
	e["payload_b64url"] = base64.RawURLEncoding.EncodeToString(raw)
	e["payload_sha256"] = hex.EncodeToString(hash[:])
	if name != "desired" {
		e["control_plane_authority_epoch"] = p["control_plane_authority_epoch"]
		e["authority_sequence"] = p["authority_sequence"]
		e["root_version"] = p["root_version"]
	}
	if name == "rotation" {
		e["previous_root_version"] = p["previous_root_version"]
		e["root_version"] = p["new_root_set"].(map[string]any)["root_version"]
	}
	if name != "anchor" {
		domain := map[string]string{"rotation": "TALENRO-NODE-STATE-ROOT-ROTATION-V1", "metadata": "TALENRO-NODE-STATE-METADATA-V1", "desired": "TALENRO-NODE-DESIRED-STATE-V1"}[name]
		pub := signer.Public().(ed25519.PublicKey)
		id := sha256.Sum256(pub)
		sig := ed25519.Sign(signer, append([]byte(domain+"\x00"), raw...))
		share := map[string]any{"key_id": hex.EncodeToString(id[:]), "signature": base64.RawURLEncoding.EncodeToString(sig)}
		switch name {
		case "rotation":
			e["current_signatures"] = []any{share}
			e["new_signatures"] = []any{share}
		case "metadata":
			e["signatures"] = []any{share}
		case "desired":
			e["key_id"] = share["key_id"]
			e["signature"] = share["signature"]
		}
	}
	return jsonBytes(t, e)
}
func signedBatch(t *testing.T, n int) Batch {
	t.Helper()
	b := Batch{ExpectedNodeID: "12345678-1234-4234-8234-123456789abc", ExpectedEpoch: 19, AnchorEnvelope: frozenFile(t, "anchor.envelope.json"), DesiredEnvelope: frozenFile(t, "desired.envelope.json")}
	for i := 0; i < n; i++ {
		p := jsonObject(t, frozenFile(t, "rotation.payload.json"))
		p["authority_sequence"] = strconv.Itoa(i + 4)
		p["previous_root_version"] = strconv.Itoa(i + 1)
		r := p["new_root_set"].(map[string]any)
		r["authority_sequence"] = strconv.Itoa(i + 4)
		r["root_version"] = strconv.Itoa(i + 2)
		b.Rotations = append(b.Rotations, mint(t, "rotation", p, testPrivate(1)))
	}
	b.ActiveRootEnvelope = b.AnchorEnvelope
	if n > 0 {
		b.ActiveRootEnvelope = b.Rotations[n-1]
	}
	m := jsonObject(t, frozenFile(t, "metadata.payload.json"))
	m["root_version"] = strconv.Itoa(n + 1)
	b.MetadataEnvelope = mint(t, "metadata", m, testPrivate(1))
	return b
}
func changePayload(t *testing.T, raw []byte, name string, change func(map[string]any)) []byte {
	p := envelopePayload(t, raw)
	change(p)
	k := testPrivate(1)
	if name == "desired" {
		k = testPrivate(2)
	}
	return mint(t, name, p, k)
}
func extraActive(t *testing.T, p map[string]any) {
	k := testKeyWire(3)
	k["status"] = "active"
	k["not_before"] = "2019-12-01T00:00:00Z"
	k["not_after"] = "2020-02-01T00:00:00Z"
	k["revoked_at"] = nil
	k["revocation_reason"] = nil
	k["incident_id"] = nil
	keys := append(p["keys"].([]any), k)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].(map[string]any)["key_id"].(string) < keys[j].(map[string]any)["key_id"].(string)
	})
	p["keys"] = keys
}
func TestFrozenSignatureVectors(t *testing.T) {
	var vectors []struct{ Name, PayloadSHA256, EnvelopeSHA256, TranscriptHex, PublicKey, Signature string }
	if e := json.Unmarshal(frozenFile(t, "vectors.json"), &vectors); e != nil {
		t.Fatal(e)
	}
	kinds := map[string]stateformat.Kind{"anchor": stateformat.InitialRoot, "rotation": stateformat.RootRotation, "metadata": stateformat.Metadata, "desired": stateformat.Desired}
	for _, v := range vectors {
		t.Run(v.Name, func(t *testing.T) {
			raw := frozenFile(t, v.Name+".envelope.json")
			p := frozenFile(t, v.Name+".payload.json")
			pd, ed := sha256.Sum256(p), sha256.Sum256(raw)
			if hex.EncodeToString(pd[:]) != v.PayloadSHA256 || hex.EncodeToString(ed[:]) != v.EnvelopeSHA256 {
				t.Fatal("frozen digest")
			}
			a, e := stateformat.ParseEnvelope(kinds[v.Name], raw)
			if e != nil {
				t.Fatal(e)
			}
			defer a.Clear()
			if !bytes.Equal(p, a.Payload()) {
				t.Fatal("payload")
			}
			if v.Name == "anchor" {
				return
			}
			msg, e := hex.DecodeString(v.TranscriptHex)
			if e != nil {
				t.Fatal(e)
			}
			fileMsg, e := hex.DecodeString(string(frozenFile(t, v.Name+".transcript.hex")))
			if e != nil || !bytes.Equal(msg, fileMsg) {
				t.Fatal("frozen transcript")
			}
			actual, e := a.Transcript()
			if e != nil || !bytes.Equal(actual, msg) {
				t.Fatal("product transcript diverges")
			}
			pub, _ := base64.RawURLEncoding.DecodeString(v.PublicKey)
			sig, _ := base64.RawURLEncoding.DecodeString(v.Signature)
			if !ed25519.Verify(pub, msg, sig) {
				t.Fatal("frozen signature")
			}
			msg[0] ^= 1
			if ed25519.Verify(pub, msg, sig) {
				t.Fatal("domain tamper passed")
			}
		})
	}
}
