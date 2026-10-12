package statepolicybinding

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/google/uuid"
	"os"
	"sort"
	"strconv"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/statebinding"
	"talenro.local/platform/internal/nodecontrol/statecomposition"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
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
func addRoot(t *testing.T, in *statecomposition.Input, raw []byte) statematerials.MaterialRef {
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
func addEnvelope(t *testing.T, in *statecomposition.Input, kind statematerials.MaterialKind, raw []byte, old, newRoot, meta statematerials.MaterialRef) statematerials.MaterialRef {
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
func compositionFixture(t *testing.T, n int) statecomposition.Input {
	t.Helper()
	in := statecomposition.Input{Materials: statematerials.Input{Scope: statematerials.Scope{NodeID: "12345678-1234-4234-8234-123456789abc", Epoch: 19}}}
	a := frozen(t, "anchor.envelope.json")
	in.Anchor = statecomposition.AnchorInput{Scope: in.Materials.Scope, Raw: a, Payload: sha256.Sum256(frozen(t, "anchor.payload.json")), Envelope: sha256.Sum256(a)}
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
		in.Target.Links = append(in.Target.Links, statecomposition.RotationLink{Rotation: rr, NewRoot: nr})
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
func material(in *statecomposition.Input, ref statematerials.MaterialRef) *statematerials.Material {
	for i := range in.Materials.Materials {
		if in.Materials.Materials[i].Ref == ref {
			return &in.Materials.Materials[i]
		}
	}
	return nil
}
func expected(in *statecomposition.Input, ref statematerials.MaterialRef) *statematerials.Expectation {
	for i := range in.Materials.Expected {
		if in.Materials.Expected[i].Ref == ref {
			return &in.Materials.Expected[i]
		}
	}
	return nil
}
func tailRef(in statecomposition.Input) statematerials.MaterialRef {
	if len(in.Target.Links) == 0 {
		return in.Target.AnchorRoot
	}
	return in.Target.Links[len(in.Target.Links)-1].NewRoot
}
func d(n byte) contracts.Digest { return contracts.Digest{n} }
func declRef(k statecontract.ArtifactKind, n byte) statecontract.ArtifactRef {
	r := statecontract.ArtifactRef{Kind: k, Payload: d(n)}
	switch k {
	case statecontract.RootRotation, statecontract.Metadata, statecontract.TrustBundle, statecontract.ResourceEnvelope, statecontract.HostMemoryPolicy, statecontract.DesiredSnapshot, statecontract.RecoverySnapshot:
		r.Envelope = d(n + 1)
	}
	return r
}
func declarationFixture() statecontract.MainStateContentModel {
	h := func(d statecontract.Domain) statecontract.SectionHeader {
		return statecontract.SectionHeader{d, statecontract.Present}
	}
	mark := func(k statecontract.ArtifactKind, n byte) statecontract.StreamMark {
		return statecontract.StreamMark{1, 1, declRef(k, n)}
	}
	scope := statecontract.MainStateScope{NodeID: "12345678-1234-4234-8234-123456789abc", AuthorityEpoch: 1, DeploymentID: d(1), TrustAnchorID: d(2), DeploymentKeySetID: d(3), TimeProviderID: d(4), Main: statecontract.GuardBinding{d(5), d(6)}, Supervisor: statecontract.GuardBinding{d(7), d(8)}, Latch: statecontract.GuardBinding{d(9), d(10)}}
	desired := statecontract.DesiredBinding{1, 1, d(40)}
	bundle := func(p statecontract.BundlePurpose, n byte) statecontract.BundleRef {
		return statecontract.BundleRef{Purpose: p, TrustDomain: "example.test", AuthorityEpoch: 1, Mark: mark(statecontract.TrustBundle, n), DeploymentKeyID: d(11), DeauthorizedSetDigest: d(12)}
	}
	return statecontract.MainStateContentModel{
		ModelVersion: 1, Scope: scope,
		Checkpoint:  statecontract.CheckpointSection{h(statecontract.DomainCheckpoint), scope.NodeID, 1, 100},
		Root:        statecontract.RootSection{h(statecontract.DomainRoot), mark(statecontract.RootPayload, 20), 1, []contracts.Digest{d(1)}},
		Metadata:    statecontract.MetadataSection{h(statecontract.DomainMetadata), mark(statecontract.Metadata, 22), 1, d(20), []contracts.Digest{d(2)}, nil, d(23), declRef(statecontract.MetadataHistory, 24), statecontract.TimeWindow{0, 100}},
		Identity:    statecontract.IdentitySection{h(statecontract.DomainIdentity), 1, scope.NodeID, 200, d(1), d(2), d(3), d(4), declRef(statecontract.CertificateDER, 25), declRef(statecontract.IdentityHistory, 26), statecontract.TimeWindow{0, 100}},
		ServerTrust: statecontract.ServerTrustSection{h(statecontract.DomainServerTrust), bundle(statecontract.BootstrapServer, 27), bundle(statecontract.AgentServer, 29)},
		Releases:    statecontract.ReleaseSection{h(statecontract.DomainReleases), d(30), statecontract.VersionMark{1, declRef(statecontract.ApprovedManifest, 31)}, statecontract.VersionMark{1, declRef(statecontract.InstalledMap, 32)}},
		Resources:   statecontract.ResourceSection{h(statecontract.DomainResources), mark(statecontract.ResourceEnvelope, 33), scope.NodeID, 1, d(11)},
		HostPolicy:  statecontract.HostPolicySection{h(statecontract.DomainHostPolicy), mark(statecontract.HostMemoryPolicy, 35), scope.NodeID, 1, d(11)},
		Desired:     statecontract.DesiredSection{h(statecontract.DomainDesired), mark(statecontract.DesiredSnapshot, 40), 1, 1, d(33), d(20), d(22), d(2), statecontract.TimeWindow{0, 100}, []string{"slot"}},
		Recovery:    statecontract.RecoverySection{Header: statecontract.SectionHeader{statecontract.DomainRecovery, statecontract.NeverEstablished}},
		Supervisor:  statecontract.SupervisorSection{h(statecontract.DomainSupervisor), d(45), scope.Supervisor, declRef(statecontract.SupervisorState, 46), statecontract.AppliedRecord{statecontract.Present, desired}, statecontract.AppliedRecord{statecontract.Present, desired}, statecontract.TransitionRef{Coverage: statecontract.NeverEstablished}},
		LKG:         statecontract.LKGSection{h(statecontract.DomainLKG), desired, declRef(statecontract.LKGHistory, 47), []statecontract.LKGEntry{{"slot", desired, declRef(statecontract.ProcessSpec, 48), declRef(statecontract.ConfigMaterial, 49), declRef(statecontract.ReleaseMaterial, 50), d(20), d(22), d(2), 100}}},
	}
}

func artifact(ref statematerials.MaterialRef) statecontract.ArtifactRef {
	kind := statecontract.ArtifactUnknown
	switch ref.Kind {
	case statematerials.RootSet:
		kind = statecontract.RootPayload
	case statematerials.RootRotation:
		kind = statecontract.RootRotation
	case statematerials.Metadata:
		kind = statecontract.Metadata
	case statematerials.Desired:
		kind = statecontract.DesiredSnapshot
	}
	return statecontract.ArtifactRef{Kind: kind, Payload: ref.Payload, Envelope: ref.Envelope}
}
func declarationFor(t *testing.T, comp statecomposition.Input) statecontract.MainStateContentModel {
	x := declarationFixture()
	root := expected(&comp, tailRef(comp))
	meta := expected(&comp, comp.Target.Metadata)
	des := expected(&comp, comp.Target.Desired)
	x.Scope.NodeID = comp.Materials.Scope.NodeID
	x.Scope.AuthorityEpoch = comp.Materials.Scope.Epoch
	x.Checkpoint.NodeID = x.Scope.NodeID
	x.Checkpoint.AuthorityEpoch = x.Scope.AuthorityEpoch
	x.Resources.NodeID = x.Scope.NodeID
	x.Resources.AuthorityEpoch = x.Scope.AuthorityEpoch
	x.HostPolicy.NodeID = x.Scope.NodeID
	x.HostPolicy.AuthorityEpoch = x.Scope.AuthorityEpoch
	x.ServerTrust.Bootstrap.AuthorityEpoch = x.Scope.AuthorityEpoch
	x.ServerTrust.Agent.AuthorityEpoch = x.Scope.AuthorityEpoch
	x.Root.Mark = statecontract.StreamMark{Version: root.Root.Version, Sequence: root.Sequence, Material: artifact(root.Ref)}
	x.Root.Threshold = uint8(root.Root.Threshold)
	x.Root.KeyIDs = append([]contracts.Digest(nil), root.Root.KeyIDs...)
	x.Metadata.Mark = statecontract.StreamMark{Version: meta.Metadata.Version, Sequence: meta.Sequence, Material: artifact(meta.Ref)}
	x.Metadata.RootVersion = meta.Metadata.RootVersion
	x.Metadata.RootPayload = meta.Metadata.RootPayload
	x.Metadata.KeyIDs = append([]contracts.Digest(nil), meta.Metadata.KeyIDs...)
	x.Metadata.RevokedIDs = append([]contracts.Digest(nil), meta.Metadata.RevokedIDs...)
	x.Metadata.Window = statecontract.TimeWindow{Issued: meta.Metadata.Window.Issued, Until: meta.Metadata.Window.Until}
	x.Desired.Mark = statecontract.StreamMark{Version: des.Desired.Generation, Sequence: des.Sequence, Material: artifact(des.Ref)}
	x.Desired.InventoryVersion = des.Desired.InventoryVersion
	x.Desired.ResourceVersion = des.Desired.ResourceVersion
	x.Desired.ResourcePayload = des.Desired.ResourcePayload
	x.Desired.RootPayload = des.Desired.RootPayload
	x.Desired.MetadataPayload = des.Desired.MetadataPayload
	x.Desired.SigningKeyID = des.Desired.SigningKeyID
	x.Desired.Window = statecontract.TimeWindow{Issued: des.Desired.Window.Issued, Until: des.Desired.Window.Until}
	x.Desired.SlotIDs = append([]string(nil), des.Desired.SlotIDs...)
	x.Resources.Mark.Version = des.Desired.ResourceVersion
	x.Resources.Mark.Material.Payload = des.Desired.ResourcePayload
	accepted := statecontract.DesiredBinding{Generation: des.Desired.Generation, Sequence: des.Sequence, Payload: des.Ref.Payload}
	x.Supervisor.Verified.Desired = accepted
	x.Supervisor.Applied.Desired = accepted
	x.LKG.HighestAccepted = accepted
	x.LKG.Entries[0].Desired = accepted
	x.LKG.Entries[0].RootPayload = des.Desired.RootPayload
	x.LKG.Entries[0].MetadataPayload = des.Desired.MetadataPayload
	x.LKG.Entries[0].SigningKeyID = des.Desired.SigningKeyID
	return x
}

func multiRootPolicyBindingFixture(t *testing.T) Input {
	t.Helper()
	original := compositionFixture(t, 1)
	next := statecomposition.Input{Materials: statematerials.Input{Scope: original.Materials.Scope}, Anchor: original.Anchor}
	next.Target.AnchorRoot = addRoot(t, &next, frozen(t, "anchor.payload.json"))
	r := object(t, frozen(t, "rotation.payload.json"))
	root := r["new_root_set"].(map[string]any)
	keys := append(root["keys"].([]any), wireKey(3))
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].(map[string]any)["key_id"].(string) < keys[j].(map[string]any)["key_id"].(string)
	})
	root["keys"] = keys
	raw := mint(t, "rotation", r, private(1))
	nr := addRoot(t, &next, embedded(t, raw))
	rr := addEnvelope(t, &next, statematerials.RootRotation, raw, next.Target.AnchorRoot, nr, statematerials.MaterialRef{})
	next.Target.Links = []statecomposition.RotationLink{{Rotation: rr, NewRoot: nr}}
	next.Target.Metadata = addEnvelope(t, &next, statematerials.Metadata, frozen(t, "metadata.envelope.json"), statematerials.MaterialRef{}, nr, statematerials.MaterialRef{})
	next.Target.Desired = addEnvelope(t, &next, statematerials.Desired, frozen(t, "desired.envelope.json"), statematerials.MaterialRef{}, nr, next.Target.Metadata)
	in := policyBindingFixture(t, 1, 1)
	x := declarationFor(t, next)
	in.Binding = statebinding.Input{Main: statecontract.ModelInput{Previous: x, Proposed: x}, Previous: next.Materials, Proposed: next}
	in.Policy.Old.Roots = []statepolicy.RootRecord{fixtureRoot(t, frozen(t, "anchor.payload.json"), next.Anchor.Payload), fixtureRoot(t, embedded(t, raw), rr.Payload)}
	in.Policy.Batch.Rotations = [][]byte{raw}
	in.Policy.Batch.ActiveRootEnvelope = raw
	in.Policy.Batch.MetadataEnvelope = material(&next, next.Target.Metadata).Raw
	in.Policy.Batch.DesiredEnvelope = material(&next, next.Target.Desired).Raw
	return copyPolicyBindingInput(t, in)
}
func fixtureTime(t *testing.T, s string) time.Time {
	t.Helper()
	v, e := time.Parse(time.RFC3339, s)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func fixturePublic(t *testing.T, k map[string]any) stateformat.PublicKeyMaterial {
	t.Helper()
	r := stateformat.PublicKeyMaterial{KeyID: digest(t, k["key_id"].(string))}
	raw, e := base64.RawURLEncoding.DecodeString(k["public_key"].(string))
	if e != nil || len(raw) != 32 {
		t.Fatal("fixture key")
	}
	copy(r.PublicKey[:], raw)
	return r
}
func fixtureRoot(t *testing.T, raw []byte, source contracts.Digest) statepolicy.RootRecord {
	t.Helper()
	p := object(t, raw)
	r := statepolicy.RootRecord{Mark: statepolicy.HighWater{Version: number(t, p, "root_version"), Sequence: number(t, p, "authority_sequence"), Payload: source}}
	r.Root.Version = r.Mark.Version
	r.Root.Threshold = number(t, p, "threshold")
	for _, k := range p["keys"].([]any) {
		r.Root.Keys = append(r.Root.Keys, fixturePublic(t, k.(map[string]any)))
	}
	return r
}
func fixtureMetadataKeys(t *testing.T, raw []byte) []statepolicy.KeyRecord {
	t.Helper()
	p := object(t, payload(t, raw))
	var keys []statepolicy.KeyRecord
	for _, item := range p["keys"].([]any) {
		k := item.(map[string]any)
		m := stateformat.MetadataKeyMaterial{PublicKeyMaterial: fixturePublic(t, k), Status: k["status"].(string), NotBefore: fixtureTime(t, k["not_before"].(string)), NotAfter: fixtureTime(t, k["not_after"].(string))}
		r := statepolicy.KeyRecord{Material: m}
		if k["revoked_at"] != nil {
			r.Material.HasRevocation = true
			r.Material.RevokedAt = fixtureTime(t, k["revoked_at"].(string))
			r.Material.RevocationReason = k["revocation_reason"].(string)
			r.FirstRevokedAt = r.Material.RevokedAt
		}
		if k["incident_id"] != nil {
			r.Material.HasIncident = true
			id, e := uuid.Parse(k["incident_id"].(string))
			if e != nil {
				t.Fatal(e)
			}
			r.Material.IncidentID = id
		}
		keys = append(keys, r)
	}
	return keys
}
func policyBindingFixture(t *testing.T, rotations, previousRotations int) Input {
	t.Helper()
	if previousRotations < 0 || rotations < previousRotations || rotations > 32 {
		t.Fatal("fixture rotations")
	}
	previous := compositionFixture(t, previousRotations)
	next := compositionFixture(t, rotations)
	if rotations > previousRotations {
		m := object(t, payload(t, material(&next, next.Target.Metadata).Raw))
		m["metadata_version"] = "7"
		m["authority_sequence"] = "6"
		next.Target.Metadata = addEnvelope(t, &next, statematerials.Metadata, mint(t, "metadata", m, private(1)), statematerials.MaterialRef{}, tailRef(next), statematerials.MaterialRef{})
		des := object(t, frozen(t, "desired.payload.json"))
		des["generation"] = "8"
		des["authority_sequence"] = "10"
		next.Target.Desired = addEnvelope(t, &next, statematerials.Desired, mint(t, "desired", des, private(2)), statematerials.MaterialRef{}, tailRef(next), next.Target.Metadata)
	}
	a := declarationFor(t, previous)
	b := declarationFor(t, next)
	in := Input{Binding: statebinding.Input{Main: statecontract.ModelInput{Previous: a, Proposed: b}, Previous: previous.Materials, Proposed: next}}
	for i := previousRotations; i < len(next.Target.Links); i++ {
		l := next.Target.Links[i]
		e := expected(&next, l.Rotation)
		in.Binding.Main.RootTransition = append(in.Binding.Main.RootTransition, statecontract.RotationRef{PreviousVersion: e.Rotation.PreviousVersion, NewVersion: e.Rotation.NewVersion, Sequence: e.Sequence, OldRootPayload: e.Rotation.OldRootPayload, NewRootPayload: e.Rotation.NewRootPayload, Material: artifact(l.Rotation)})
	}
	node, e := uuid.Parse(a.Scope.NodeID)
	if e != nil {
		t.Fatal(e)
	}
	old := &statepolicy.StateModel{Status: statepolicy.Available, StateID: d(80), CounterID: d(81), StateEpoch: 82, NodeID: node, Epoch: 19, AnchorPayload: next.Anchor.Payload, DeploymentSourceID: d(83), TimeProviderID: a.Scope.TimeProviderID, TimeFloor: fixtureTime(t, "2019-12-31T00:00:00Z")}
	old.Roots = append(old.Roots, fixtureRoot(t, frozen(t, "anchor.payload.json"), next.Anchor.Payload))
	for _, l := range previous.Target.Links {
		old.Roots = append(old.Roots, fixtureRoot(t, material(&previous, l.NewRoot).Raw, l.Rotation.Payload))
	}
	old.Metadata = statepolicy.HighWater{Version: a.Metadata.Mark.Version, Sequence: a.Metadata.Mark.Sequence, Payload: a.Metadata.Mark.Material.Payload}
	old.Desired = statepolicy.HighWater{Version: a.Desired.Mark.Version, Sequence: a.Desired.Mark.Sequence, Payload: a.Desired.Mark.Material.Payload}
	old.Keys = fixtureMetadataKeys(t, material(&previous, previous.Target.Metadata).Raw)
	old.RevokedIDs = append([]contracts.Digest(nil), a.Metadata.RevokedIDs...)
	in.Policy.Old = old
	in.Policy.Deployment = statepolicy.DeploymentObservation{Status: statepolicy.Available, SourceID: old.DeploymentSourceID, ObservationID: d(84), NodeID: node, Epoch: 19, AnchorPayload: next.Anchor.Payload}
	in.Policy.Time = statepolicy.TimeObservation{Status: statepolicy.Available, ProviderID: old.TimeProviderID, ObservationID: d(85), Now: fixtureTime(t, "2020-01-01T01:00:00Z"), Floor: old.TimeFloor}
	in.Policy.Batch.ExpectedNodeID = next.Materials.Scope.NodeID
	in.Policy.Batch.ExpectedEpoch = 19
	in.Policy.Batch.AnchorEnvelope = next.Anchor.Raw
	for _, l := range next.Target.Links {
		in.Policy.Batch.Rotations = append(in.Policy.Batch.Rotations, material(&next, l.Rotation).Raw)
	}
	in.Policy.Batch.ActiveRootEnvelope = next.Anchor.Raw
	if len(in.Policy.Batch.Rotations) > 0 {
		in.Policy.Batch.ActiveRootEnvelope = in.Policy.Batch.Rotations[len(in.Policy.Batch.Rotations)-1]
	}
	in.Policy.Batch.MetadataEnvelope = material(&next, next.Target.Metadata).Raw
	in.Policy.Batch.DesiredEnvelope = material(&next, next.Target.Desired).Raw
	return copyPolicyBindingInput(t, in)
}
func copyPolicyBindingInput(t *testing.T, in Input) Input {
	t.Helper()
	raw, e := json.Marshal(in)
	if e != nil {
		t.Fatal(e)
	}
	var out Input
	if e = json.Unmarshal(raw, &out); e != nil {
		t.Fatal(e)
	}
	return out
}
func checkBindingInput(t *testing.T, in Input) statebinding.Summary {
	t.Helper()
	m, e := statebinding.Evaluate(context.Background(), in.Binding)
	if e != nil {
		t.Fatal("binding fixture", e)
	}
	defer m.Clear()
	return m.Summary()
}
func checkPolicyInput(t *testing.T, in Input) {
	t.Helper()
	c, e := statepolicy.EvaluateModel(context.Background(), in.Policy)
	if e != nil {
		t.Fatal("policy fixture", e)
	}
	defer c.Clear()
}
func historicalPolicyBindingFixture(t *testing.T) Input {
	in := policyBindingFixture(t, 1, 1)
	next := &in.Binding.Proposed
	oldRoot := tailRef(*next)
	r := object(t, frozen(t, "rotation.payload.json"))
	r["previous_root_version"] = "2"
	r["authority_sequence"] = "5"
	root := r["new_root_set"].(map[string]any)
	root["root_version"] = "3"
	root["authority_sequence"] = "5"
	raw := mint(t, "rotation", r, private(1))
	nr := addRoot(t, next, embedded(t, raw))
	rr := addEnvelope(t, next, statematerials.RootRotation, raw, oldRoot, nr, statematerials.MaterialRef{})
	m := object(t, frozen(t, "metadata.payload.json"))
	m["root_version"] = "3"
	m["metadata_version"] = "8"
	m["authority_sequence"] = "10"
	mr := addEnvelope(t, next, statematerials.Metadata, mint(t, "metadata", m, private(1)), statematerials.MaterialRef{}, nr, statematerials.MaterialRef{})
	re := expected(next, nr)
	me := expected(next, mr)
	in.Binding.Main.Proposed.Root.Mark = statecontract.StreamMark{Version: 3, Sequence: re.Sequence, Material: artifact(nr)}
	in.Binding.Main.Proposed.Metadata.Mark = statecontract.StreamMark{Version: 8, Sequence: me.Sequence, Material: artifact(mr)}
	in.Binding.Main.Proposed.Metadata.RootVersion = 3
	in.Binding.Main.Proposed.Metadata.RootPayload = nr.Payload
	in.Binding.Main.RootTransition = []statecontract.RotationRef{{PreviousVersion: 2, NewVersion: 3, Sequence: 5, OldRootPayload: oldRoot.Payload, NewRootPayload: nr.Payload, Material: artifact(rr)}}
	return copyPolicyBindingInput(t, in)
}
