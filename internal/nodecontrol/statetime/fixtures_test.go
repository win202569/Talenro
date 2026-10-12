package statetime

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gowebpki/jcs"
	"sort"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"testing"
	"time"
)

func canon(t *testing.T, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	b, e = jcs.Transform(b)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func digest(n byte) contracts.Digest { return sha256.Sum256([]byte{n}) }
func decodeObject(t *testing.T, b []byte, m *map[string]any) {
	t.Helper()
	if e := json.Unmarshal(b, m); e != nil {
		t.Fatal(e)
	}
}
func makeNonzero(n int) []byte        { return bytes.Repeat([]byte{1}, n) }
func hexID(d contracts.Digest) string { return fmt.Sprintf("%x", d) }
func instant(s string) time.Time {
	v, e := time.Parse(time.RFC3339, s)
	if e != nil {
		panic(e)
	}
	return v
}
func private(n byte) ed25519.PrivateKey { return ed25519.NewKeyFromSeed(bytes.Repeat([]byte{n}, 32)) }
func keyID(n byte) contracts.Digest     { return sha256.Sum256(private(n).Public().(ed25519.PublicKey)) }
func b64(b []byte) string               { return base64.RawURLEncoding.EncodeToString(b) }
func payload(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if e := json.Unmarshal(raw, &m); e != nil {
		t.Fatal(e)
	}
	p, e := base64.RawURLEncoding.DecodeString(m["payload_b64url"].(string))
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(p, &m); e != nil {
		t.Fatal(e)
	}
	delete(m, "payload_b64url")
	delete(m, "payload_sha256")
	delete(m, "key_id")
	delete(m, "signature")
	delete(m, "signatures")
	return m
}
func attestation(t *testing.T, p map[string]any, n byte, domain string) []byte {
	b := canon(t, p)
	s := ed25519.Sign(private(n), append([]byte(domain+"\x00"), b...))
	return canon(t, map[string]any{"schema_version": "talenro-node-time-attestation/v1", "payload_b64url": b64(b), "payload_sha256": fmt.Sprintf("%x", sha256.Sum256(b)), "key_id": fmt.Sprintf("%x", keyID(n)), "signature": b64(s)})
}
func setMetadata(t *testing.T, in *ModelInput, p map[string]any) {
	b := canon(t, p)
	in.MetadataEnvelope = canon(t, map[string]any{"schema_version": "talenro-node-state-metadata/v1", "payload_b64url": b64(b), "payload_sha256": fmt.Sprintf("%x", sha256.Sum256(b)), "control_plane_authority_epoch": p["control_plane_authority_epoch"], "authority_sequence": p["authority_sequence"], "root_version": p["root_version"], "signatures": []any{}})
	a, e := stateformat.ParseEnvelope(stateformat.Metadata, in.MetadataEnvelope)
	if e != nil {
		t.Fatal("fixture metadata", e)
	}
	defer a.Clear()
	in.Metadata.Payload = a.PayloadDigest()
	in.Metadata.Envelope = a.EnvelopeDigest()
}
func base(t *testing.T) ModelInput {
	t.Helper()
	node := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	in := ModelInput{
		Metadata:   CurrentMetadataObservation{Status: Available, SourceID: digest(1), ObservationID: digest(2), StateID: digest(3), NodeID: node, Epoch: 19, Version: 6},
		Time:       TimeObservation{Status: Available, ProviderID: digest(4), ObservationID: digest(5), Now: instant("2026-10-09T00:01:00Z"), Floor: instant("2026-10-09T00:00:00Z")},
		Request:    RequestObservation{Status: Available, RequestID: digest(6), ObservationID: digest(7), NodeID: node, Epoch: 19},
		Checkpoint: CheckpointObservation{Status: Available, SourceID: digest(8), ObservationID: digest(9), StateID: digest(3), NodeID: node, Epoch: 19, Sequence: 7}}
	keys := []any{}
	for _, n := range []byte{1, 2} {
		keys = append(keys, map[string]any{"key_id": fmt.Sprintf("%x", keyID(n)), "algorithm": "ed25519", "public_key": b64(private(n).Public().(ed25519.PublicKey)), "status": "active", "not_before": "2026-10-08T00:00:00Z", "not_after": "2026-10-17T00:00:00Z", "revoked_at": nil, "revocation_reason": nil, "incident_id": nil})
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].(map[string]any)["key_id"].(string) < keys[j].(map[string]any)["key_id"].(string)
	})
	setMetadata(t, &in, map[string]any{"schema_version": "node-state-trust-metadata.v1", "control_plane_authority_epoch": "19", "authority_sequence": "900", "root_version": "2", "metadata_version": "6", "issued_at": "2026-10-09T00:00:00Z", "valid_until": "2026-10-16T00:00:00Z", "next_refresh_at": "2026-10-14T00:00:00Z", "role": "node_state", "keys": keys, "revoked_key_ids": []any{}})
	in.AttestationEnvelope = attestation(t, map[string]any{"schema_version": "node-time-attestation.v1", "control_plane_authority_epoch": "19", "node_authority_checkpoint_sequence": "7", "node_id": node.String(), "request_nonce": b64(make([]byte, 32)), "issued_at": "2026-10-09T00:00:00Z", "valid_until": "2026-10-09T00:05:00Z"}, 1, "TALENRO-NODE-TIME-ATTESTATION-V1")
	return in
}
func changeAtt(t *testing.T, in *ModelInput, k string, v any) {
	p := payload(t, in.AttestationEnvelope)
	p[k] = v
	in.AttestationEnvelope = attestation(t, p, 1, "TALENRO-NODE-TIME-ATTESTATION-V1")
}
func signingKey(p map[string]any) map[string]any {
	for _, k := range p["keys"].([]any) {
		m := k.(map[string]any)
		if m["key_id"] == fmt.Sprintf("%x", keyID(1)) {
			return m
		}
	}
	panic("key absent")
}
