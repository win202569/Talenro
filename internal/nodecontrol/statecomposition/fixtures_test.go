package statecomposition

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"sort"
	"strconv"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"testing"
	"time"
)

func frozen(t *testing.T, name string) []byte {
	t.Helper()
	raw, e := os.ReadFile("../stateverify/testdata/" + name)
	if e != nil {
		t.Fatal(e)
	}
	return raw
}
func object(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var o map[string]any
	if e := json.Unmarshal(raw, &o); e != nil {
		t.Fatal(e)
	}
	return o
}
func encode(t *testing.T, o any) []byte {
	t.Helper()
	raw, e := json.Marshal(o)
	if e != nil {
		t.Fatal(e)
	}
	return raw
}
func payload(t *testing.T, raw []byte) []byte {
	t.Helper()
	o := object(t, raw)
	p, e := base64.RawURLEncoding.DecodeString(o["payload_b64url"].(string))
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func digest(t *testing.T, s string) (d contracts.Digest) {
	t.Helper()
	b, e := hex.DecodeString(s)
	if e != nil || len(b) != 32 {
		t.Fatal("digest")
	}
	copy(d[:], b)
	return
}
func number(t *testing.T, o map[string]any, k string) uint64 {
	t.Helper()
	n, e := strconv.ParseUint(o[k].(string), 10, 64)
	if e != nil {
		t.Fatal(e)
	}
	return n
}
func window(t *testing.T, o map[string]any) statematerials.Window {
	t.Helper()
	a, e := time.Parse(time.RFC3339, o["issued_at"].(string))
	if e != nil {
		t.Fatal(e)
	}
	b, e := time.Parse(time.RFC3339, o["valid_until"].(string))
	if e != nil {
		t.Fatal(e)
	}
	return statematerials.Window{Issued: a.Unix(), Until: b.Unix()}
}
func private(n byte) ed25519.PrivateKey {
	seed := make([]byte, 32)
	seed[0] = n
	return ed25519.NewKeyFromSeed(seed)
}
func wireKey(n byte) map[string]any {
	p := private(n).Public().(ed25519.PublicKey)
	id := sha256.Sum256(p)
	return map[string]any{"algorithm": "ed25519", "key_id": hex.EncodeToString(id[:]), "public_key": base64.RawURLEncoding.EncodeToString(p)}
}
func mint(t *testing.T, name string, p map[string]any, key ed25519.PrivateKey) []byte {
	t.Helper()
	raw := encode(t, p)
	pd := sha256.Sum256(raw)
	e := object(t, frozen(t, name+".envelope.json"))
	e["payload_b64url"] = base64.RawURLEncoding.EncodeToString(raw)
	e["payload_sha256"] = hex.EncodeToString(pd[:])
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
		pub := key.Public().(ed25519.PublicKey)
		id := sha256.Sum256(pub)
		sig := ed25519.Sign(key, append([]byte(domain+"\x00"), raw...))
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
	return encode(t, e)
}
func addRoot(t *testing.T, in *Input, raw []byte) statematerials.MaterialRef {
	t.Helper()
	p := object(t, raw)
	ref := statematerials.MaterialRef{Kind: statematerials.RootSet, Payload: sha256.Sum256(raw)}
	ids := []contracts.Digest{}
	for _, k := range p["keys"].([]any) {
		ids = append(ids, digest(t, k.(map[string]any)["key_id"].(string)))
	}
	in.Materials.Materials = append(in.Materials.Materials, statematerials.Material{Ref: ref, Raw: raw})
	in.Materials.Expected = append(in.Materials.Expected, statematerials.Expectation{Ref: ref, Epoch: number(t, p, "control_plane_authority_epoch"), Sequence: number(t, p, "authority_sequence"), Root: statematerials.RootExpectation{Version: number(t, p, "root_version"), Threshold: number(t, p, "threshold"), KeyIDs: ids}})
	return ref
}
func addEnvelope(t *testing.T, in *Input, kind statematerials.MaterialKind, raw []byte, old, newRoot, meta statematerials.MaterialRef) statematerials.MaterialRef {
	t.Helper()
	pr := payload(t, raw)
	p := object(t, pr)
	env := object(t, raw)
	ref := statematerials.MaterialRef{Kind: kind, Payload: sha256.Sum256(pr), Envelope: sha256.Sum256(raw)}
	e := statematerials.Expectation{Ref: ref, Epoch: number(t, p, "control_plane_authority_epoch"), Sequence: number(t, p, "authority_sequence")}
	switch kind {
	case statematerials.RootRotation:
		e.Rotation = statematerials.RotationExpectation{PreviousVersion: number(t, p, "previous_root_version"), NewVersion: number(t, p["new_root_set"].(map[string]any), "root_version"), OldRootPayload: old.Payload, NewRootPayload: newRoot.Payload, Window: window(t, p)}
	case statematerials.Metadata:
		ids := []contracts.Digest{}
		rev := []contracts.Digest{}
		for _, k := range p["keys"].([]any) {
			ids = append(ids, digest(t, k.(map[string]any)["key_id"].(string)))
		}
		for _, k := range p["revoked_key_ids"].([]any) {
			rev = append(rev, digest(t, k.(string)))
		}
		e.Metadata = statematerials.MetadataExpectation{Version: number(t, p, "metadata_version"), RootVersion: number(t, p, "root_version"), RootPayload: newRoot.Payload, KeyIDs: ids, RevokedIDs: rev, Window: window(t, p)}
	case statematerials.Desired:
		slots := []string{}
		for _, proc := range p["processes"].([]any) {
			slots = append(slots, proc.(map[string]any)["slot_id"].(string))
		}
		e.Desired = statematerials.DesiredExpectation{NodeID: p["node_id"].(string), Generation: number(t, p, "generation"), InventoryVersion: number(t, p, "inventory_version"), ResourceVersion: number(t, p, "node_resource_envelope_version"), ResourcePayload: digest(t, p["node_resource_envelope_digest"].(string)), RootPayload: newRoot.Payload, MetadataPayload: meta.Payload, SigningKeyID: digest(t, env["key_id"].(string)), Window: window(t, p), SlotIDs: slots}
	}
	in.Materials.Materials = append(in.Materials.Materials, statematerials.Material{Ref: ref, Raw: raw})
	in.Materials.Expected = append(in.Materials.Expected, e)
	return ref
}
func embedded(t *testing.T, raw []byte) []byte {
	t.Helper()
	var p struct {
		Root json.RawMessage `json:"new_root_set"`
	}
	if e := json.Unmarshal(payload(t, raw), &p); e != nil {
		t.Fatal(e)
	}
	return p.Root
}
func fixture(t *testing.T, n int) Input {
	t.Helper()
	in := Input{Materials: statematerials.Input{Scope: statematerials.Scope{NodeID: "12345678-1234-4234-8234-123456789abc", Epoch: 19}}}
	a := frozen(t, "anchor.envelope.json")
	in.Anchor = AnchorInput{Scope: in.Materials.Scope, Raw: a, Payload: sha256.Sum256(frozen(t, "anchor.payload.json")), Envelope: sha256.Sum256(a)}
	in.Target.AnchorRoot = addRoot(t, &in, frozen(t, "anchor.payload.json"))
	tail := in.Target.AnchorRoot
	for i := 0; i < n; i++ {
		raw := frozen(t, "rotation.envelope.json")
		if n != 1 {
			p := object(t, frozen(t, "rotation.payload.json"))
			p["authority_sequence"] = strconv.Itoa(i + 4)
			p["previous_root_version"] = strconv.Itoa(i + 1)
			r := p["new_root_set"].(map[string]any)
			r["root_version"] = strconv.Itoa(i + 2)
			r["authority_sequence"] = strconv.Itoa(i + 4)
			raw = mint(t, "rotation", p, private(1))
		}
		nr := addRoot(t, &in, embedded(t, raw))
		rr := addEnvelope(t, &in, statematerials.RootRotation, raw, tail, nr, statematerials.MaterialRef{})
		in.Target.Links = append(in.Target.Links, RotationLink{Rotation: rr, NewRoot: nr})
		tail = nr
	}
	raw := frozen(t, "metadata.envelope.json")
	if n != 1 {
		p := object(t, frozen(t, "metadata.payload.json"))
		p["root_version"] = strconv.Itoa(n + 1)
		raw = mint(t, "metadata", p, private(1))
	}
	in.Target.Metadata = addEnvelope(t, &in, statematerials.Metadata, raw, statematerials.MaterialRef{}, tail, statematerials.MaterialRef{})
	in.Target.Desired = addEnvelope(t, &in, statematerials.Desired, frozen(t, "desired.envelope.json"), statematerials.MaterialRef{}, tail, in.Target.Metadata)
	if n == 1 {
		var vs []struct{ Name, PayloadSHA256, EnvelopeSHA256 string }
		if e := json.Unmarshal(frozen(t, "vectors.json"), &vs); e != nil {
			t.Fatal(e)
		}
		refs := map[string]statematerials.MaterialRef{"anchor": {Payload: in.Anchor.Payload, Envelope: in.Anchor.Envelope}, "rotation": in.Target.Links[0].Rotation, "metadata": in.Target.Metadata, "desired": in.Target.Desired}
		for _, v := range vs {
			if refs[v.Name].Payload != digest(t, v.PayloadSHA256) || refs[v.Name].Envelope != digest(t, v.EnvelopeSHA256) {
				t.Fatal("frozen vector divergence")
			}
		}
	}
	return in
}
func material(in *Input, ref statematerials.MaterialRef) *statematerials.Material {
	for i := range in.Materials.Materials {
		if in.Materials.Materials[i].Ref == ref {
			return &in.Materials.Materials[i]
		}
	}
	return nil
}
func expected(in *Input, ref statematerials.MaterialRef) *statematerials.Expectation {
	for i := range in.Materials.Expected {
		if in.Materials.Expected[i].Ref == ref {
			return &in.Materials.Expected[i]
		}
	}
	return nil
}
func tailRef(in Input) statematerials.MaterialRef {
	if len(in.Target.Links) == 0 {
		return in.Target.AnchorRoot
	}
	return in.Target.Links[len(in.Target.Links)-1].NewRoot
}
func compatibleMetadata(t *testing.T, in *Input, version int, revoked bool) statematerials.MaterialRef {
	p := object(t, payload(t, material(in, in.Target.Metadata).Raw))
	p["metadata_version"] = strconv.Itoa(version)
	if revoked {
		key := p["keys"].([]any)[0].(map[string]any)
		key["status"] = "revoked"
		key["revoked_at"] = "2020-01-01T00:00:00Z"
		key["revocation_reason"] = "scheduled"
		p["revoked_key_ids"] = []any{key["key_id"]}
		k := wireKey(3)
		k["status"] = "active"
		k["not_before"] = "2019-12-31T00:00:00Z"
		k["not_after"] = "2020-01-10T00:00:00Z"
		k["revoked_at"] = nil
		k["revocation_reason"] = nil
		k["incident_id"] = nil
		keys := append(p["keys"].([]any), k)
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].(map[string]any)["key_id"].(string) < keys[j].(map[string]any)["key_id"].(string)
		})
		p["keys"] = keys
	}
	return addEnvelope(t, in, statematerials.Metadata, mint(t, "metadata", p, private(1)), statematerials.MaterialRef{}, tailRef(*in), statematerials.MaterialRef{})
}
func extraShare(t *testing.T, raw []byte, field string) []byte {
	e := object(t, raw)
	s := map[string]any{"key_id": wireKey(3)["key_id"], "signature": base64.RawURLEncoding.EncodeToString(ed25519.Sign(private(3), []byte("unknown-share-test")))}
	shares := append(e[field].([]any), s)
	sort.Slice(shares, func(i, j int) bool {
		return shares[i].(map[string]any)["key_id"].(string) < shares[j].(map[string]any)["key_id"].(string)
	})
	e[field] = shares
	return encode(t, e)
}
func replaceEnvelope(t *testing.T, in *Input, ref statematerials.MaterialRef, raw []byte) statematerials.MaterialRef {
	m, e := material(in, ref), expected(in, ref)
	nr := ref
	nr.Payload = sha256.Sum256(payload(t, raw))
	nr.Envelope = sha256.Sum256(raw)
	m.Ref = nr
	m.Raw = raw
	e.Ref = nr
	return nr
}
func disconnected(t *testing.T, in *Input) RotationLink {
	p := object(t, frozen(t, "anchor.payload.json"))
	p["root_version"] = "100"
	p["authority_sequence"] = "100"
	old := addRoot(t, in, encode(t, p))
	r := object(t, frozen(t, "rotation.payload.json"))
	r["previous_root_version"] = "100"
	r["authority_sequence"] = "101"
	nr := r["new_root_set"].(map[string]any)
	nr["root_version"] = "101"
	nr["authority_sequence"] = "101"
	raw := mint(t, "rotation", r, private(1))
	next := addRoot(t, in, embedded(t, raw))
	rr := addEnvelope(t, in, statematerials.RootRotation, raw, old, next, statematerials.MaterialRef{})
	return RotationLink{Rotation: rr, NewRoot: next}
}
