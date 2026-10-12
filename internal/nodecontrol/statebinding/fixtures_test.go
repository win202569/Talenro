package statebinding

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"sort"
	"strconv"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/statecomposition"
	"talenro.local/platform/internal/nodecontrol/statecontract"
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
func compatibleMetadata(t *testing.T, in *statecomposition.Input, version int, revoked bool) statematerials.MaterialRef {
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
func replaceEnvelope(t *testing.T, in *statecomposition.Input, ref statematerials.MaterialRef, raw []byte) statematerials.MaterialRef {
	m, e := material(in, ref), expected(in, ref)
	nr := ref
	nr.Payload = sha256.Sum256(payload(t, raw))
	nr.Envelope = sha256.Sum256(raw)
	m.Ref = nr
	m.Raw = raw
	e.Ref = nr
	return nr
}
func disconnected(t *testing.T, in *statecomposition.Input) statecomposition.RotationLink {
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
	return statecomposition.RotationLink{Rotation: rr, NewRoot: next}
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
func bindingFixture(t *testing.T, rotations int) Input {
	t.Helper()
	comp := compositionFixture(t, rotations)
	x := declarationFor(t, comp)
	return copyBindingInput(t, Input{Main: statecontract.ModelInput{Previous: x, Proposed: x}, Previous: comp.Materials, Proposed: comp})
}
func copyBindingInput(t *testing.T, in Input) Input {
	t.Helper()
	var out Input
	if err := json.Unmarshal(jsonInput(t, in), &out); err != nil {
		t.Fatal(err)
	}
	return out
}
func checkBindingLayers(t *testing.T, in Input) statecomposition.Summary {
	t.Helper()
	l0, err := statecontract.EvaluateModel(context.Background(), in.Main)
	if err != nil {
		t.Fatal("L0 fixture", err)
	}
	t.Cleanup(l0.Clear)
	old, err := statematerials.Evaluate(context.Background(), in.Previous)
	if err != nil {
		t.Fatal("old fixture", err)
	}
	t.Cleanup(old.Clear)
	proposed, err := statecomposition.Evaluate(context.Background(), in.Proposed)
	if err != nil {
		t.Fatal("proposed fixture", err)
	}
	t.Cleanup(proposed.Clear)
	return proposed.Summary()
}
func separatedFixture(t *testing.T) Input {
	in := bindingFixture(t, 1)
	oldRoot := tailRef(in.Proposed)
	r := object(t, frozen(t, "rotation.payload.json"))
	r["previous_root_version"] = "2"
	r["authority_sequence"] = "5"
	root := r["new_root_set"].(map[string]any)
	root["root_version"] = "3"
	root["authority_sequence"] = "5"
	raw := mint(t, "rotation", r, private(1))
	next := addRoot(t, &in.Proposed, embedded(t, raw))
	rr := addEnvelope(t, &in.Proposed, statematerials.RootRotation, raw, oldRoot, next, statematerials.MaterialRef{})
	p := object(t, frozen(t, "metadata.payload.json"))
	p["root_version"] = "3"
	p["metadata_version"] = "8"
	p["authority_sequence"] = "10"
	mr := addEnvelope(t, &in.Proposed, statematerials.Metadata, mint(t, "metadata", p, private(1)), statematerials.MaterialRef{}, next, statematerials.MaterialRef{})
	re := expected(&in.Proposed, next)
	me := expected(&in.Proposed, mr)
	in.Main.Proposed.Root.Mark = statecontract.StreamMark{Version: 3, Sequence: re.Sequence, Material: artifact(next)}
	in.Main.Proposed.Root.KeyIDs = append([]contracts.Digest(nil), re.Root.KeyIDs...)
	in.Main.Proposed.Metadata.Mark = statecontract.StreamMark{Version: 8, Sequence: me.Sequence, Material: artifact(mr)}
	in.Main.Proposed.Metadata.RootVersion = 3
	in.Main.Proposed.Metadata.RootPayload = next.Payload
	in.Main.RootTransition = []statecontract.RotationRef{{PreviousVersion: 2, NewVersion: 3, Sequence: 5, OldRootPayload: oldRoot.Payload, NewRootPayload: next.Payload, Material: artifact(rr)}}
	return copyBindingInput(t, in)
}

// The selected historical v1 authorizer is complete, while declared current
// roots live in separate tables and the new table omits the transition start.
// No material relation points at the undeclared transition, so all lower layers
// remain valid; the binding layer must report the missing projected object.
func missingTransitionStartFixture(t *testing.T) Input {
	t.Helper()
	in := bindingFixture(t, 0)
	current := separatedFixture(t)
	oldArtifact := current.Main.Previous.Root.Mark.Material
	newArtifact := current.Main.Proposed.Root.Mark.Material
	oldRef := statematerials.MaterialRef{Kind: statematerials.RootSet, Payload: oldArtifact.Payload}
	newRef := statematerials.MaterialRef{Kind: statematerials.RootSet, Payload: newArtifact.Payload}
	old := statecomposition.Input{Materials: in.Previous}
	addRoot(t, &old, material(&current.Proposed, oldRef).Raw)
	in.Previous = old.Materials
	addRoot(t, &in.Proposed, material(&current.Proposed, newRef).Raw)
	in.Main.Previous.Root = current.Main.Previous.Root
	in.Main.Proposed.Root = current.Main.Proposed.Root
	in.Main.RootTransition = current.Main.RootTransition
	return copyBindingInput(t, in)
}

func currentMetadataFixture(t *testing.T, revoked bool) Input {
	in := bindingFixture(t, 1)
	if revoked {
		// Frozen desired has no processes. Add one valid, independently signed
		// process so same-count slot mismatches exercise a real lower-valid fact.
		p := object(t, frozen(t, "desired.payload.json"))
		process := object(t, []byte(`{"adapter":"fixture","capacity_limits":{"connection_limit":"1","cpu_limit_basis_points":"1","cpu_quota_millicores":"100","egress_limit_bps":"1000000","file_descriptor_limit":"64","handshake_limit_per_second":"1","memory_limit_bytes":"67108864","packet_loss_limit_basis_points":"1","queue_limit":"1","task_limit":"32"},"capacity_profile_id":"capacity.v1","capacity_profile_version":"1","lifecycle":"running","probe_profile":"probe.v1","release_id":"release.v1","required":false,"required_metrics":["bytes"],"restart_profile":"restart.v1","slot_id":"slot-a","test_profile":"test.v1"}`))
		p["processes"] = []any{process}
		in.Proposed.Target.Desired = replaceEnvelope(t, &in.Proposed, in.Proposed.Target.Desired, mint(t, "desired", p, private(2)))
		expected(&in.Proposed, in.Proposed.Target.Desired).Desired.SlotIDs = []string{"slot-a"}
		x := declarationFor(t, in.Proposed)
		in.Main.Previous = x
		in.Main.Proposed = x
	}
	r := compatibleMetadata(t, &in.Proposed, 8, revoked)
	e := expected(&in.Proposed, r)
	for _, x := range []*statecontract.MainStateContentModel{&in.Main.Previous, &in.Main.Proposed} {
		x.Metadata.Mark = statecontract.StreamMark{Version: e.Metadata.Version, Sequence: e.Sequence, Material: artifact(r)}
		x.Metadata.KeyIDs = append([]contracts.Digest(nil), e.Metadata.KeyIDs...)
		x.Metadata.RevokedIDs = append([]contracts.Digest(nil), e.Metadata.RevokedIDs...)
	}
	in.Previous = in.Proposed.Materials
	return copyBindingInput(t, in)
}

func neverPrevious(in *Input) {
	x := &in.Main.Previous
	x.Desired = statecontract.DesiredSection{Header: statecontract.SectionHeader{Domain: statecontract.DomainDesired, Coverage: statecontract.NeverEstablished}}
	x.Supervisor.Verified = statecontract.AppliedRecord{Coverage: statecontract.NeverEstablished}
	x.Supervisor.Applied = statecontract.AppliedRecord{Coverage: statecontract.NeverEstablished}
	x.LKG = statecontract.LKGSection{Header: statecontract.SectionHeader{Domain: statecontract.DomainLKG, Coverage: statecontract.NeverEstablished}}
}

// All nineteen L0 collections are nonempty and independently lower-valid.
func richBindingFixture(t *testing.T) Input {
	in := currentMetadataFixture(t, true)
	next := separatedFixture(t)
	rotation := next.Main.RootTransition[0].Material
	rr := statematerials.MaterialRef{Kind: statematerials.RootRotation, Payload: rotation.Payload, Envelope: rotation.Envelope}
	root := statematerials.MaterialRef{Kind: statematerials.RootSet, Payload: next.Main.Proposed.Root.Mark.Material.Payload}
	for _, r := range []statematerials.MaterialRef{root, rr} {
		in.Proposed.Materials.Materials = append(in.Proposed.Materials.Materials, *material(&next.Proposed, r))
		in.Proposed.Materials.Expected = append(in.Proposed.Materials.Expected, *expected(&next.Proposed, r))
	}
	in.Main.Proposed.Root = next.Main.Proposed.Root
	in.Main.RootTransition = append([]statecontract.RotationRef(nil), next.Main.RootTransition...)
	for _, x := range []*statecontract.MainStateContentModel{&in.Main.Previous, &in.Main.Proposed} {
		x.ServerTrust.Bootstrap.Deauthorized = []contracts.Digest{d(70)}
		x.ServerTrust.Agent.Deauthorized = []contracts.Digest{d(71)}
		x.Recovery = statecontract.RecoverySection{Header: statecontract.SectionHeader{Domain: statecontract.DomainRecovery, Coverage: statecontract.Present}, Mark: statecontract.StreamMark{Version: 1, Sequence: 1, Material: declRef(statecontract.RecoverySnapshot, 60)}, IdentityEpoch: x.Identity.IdentityEpoch, RecoveryID: x.Scope.NodeID, SessionVersion: 1, Session: declRef(statecontract.RecoverySession, 62), RootPayload: x.Desired.RootPayload, MetadataPayload: x.Desired.MetadataPayload, SigningKeyID: x.Desired.SigningKeyID, Window: x.Desired.Window, Local: []statecontract.LocalFaultBinding{{FaultID: x.Scope.NodeID, IncidentID: x.Scope.NodeID}}, Supervisor: []statecontract.SupervisorFaultBinding{{FaultID: x.Scope.NodeID, IncidentID: x.Scope.NodeID, Evidence: d(72)}}}
	}
	return copyBindingInput(t, in)
}
