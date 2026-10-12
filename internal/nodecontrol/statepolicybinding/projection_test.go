package statepolicybinding

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"sort"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/statecomposition"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
	"testing"
	"time"
)

func TestPolicyBindingProjection(t *testing.T) {
	for _, proposed := range []bool{false, true} {
		in := policyBindingFixture(t, 1, 1)
		table := in.Binding.Previous
		main := in.Binding.Main.Previous
		if proposed {
			table = in.Binding.Proposed.Materials
			main = in.Binding.Main.Proposed
		}
		if e := bindStateProjection(context.Background(), table, main, in.Policy.Old); e != nil {
			t.Fatal(e)
		}
		for _, change := range []func(*statepolicy.StateModel){
			func(s *statepolicy.StateModel) { s.Metadata.Version++ }, func(s *statepolicy.StateModel) { s.Metadata.Sequence++ }, func(s *statepolicy.StateModel) { s.Metadata.Payload[0] ^= 1 },
			func(s *statepolicy.StateModel) { s.Desired.Version++ }, func(s *statepolicy.StateModel) { s.Desired.Sequence++ }, func(s *statepolicy.StateModel) { s.Desired.Payload[0] ^= 1 },
			func(s *statepolicy.StateModel) { s.Keys = nil }, func(s *statepolicy.StateModel) { s.Keys[0].Material.KeyID[0] ^= 1 },
			func(s *statepolicy.StateModel) { s.Keys[0].Material.PublicKey[0] ^= 1 }, func(s *statepolicy.StateModel) { s.Keys[0].Material.Status = "retiring" },
			func(s *statepolicy.StateModel) {
				s.Keys[0].Material.NotAfter = s.Keys[0].Material.NotAfter.Add(time.Nanosecond)
			},
			func(s *statepolicy.StateModel) { s.Keys[0].Material.HasRevocation = true },
			func(s *statepolicy.StateModel) { s.Keys[0].Material.RevokedAt = s.TimeFloor },
			func(s *statepolicy.StateModel) { s.Keys[0].Material.RevocationReason = "scheduled" },
			func(s *statepolicy.StateModel) { s.Keys[0].Material.HasIncident = true },
			func(s *statepolicy.StateModel) { s.Keys[0].Material.IncidentID[0] ^= 1 },
			func(s *statepolicy.StateModel) { s.Keys = append(s.Keys, s.Keys[0]) },
			func(s *statepolicy.StateModel) { s.RevokedIDs = []contracts.Digest{d(90)} },
		} {
			bad := copyPolicyBindingInput(t, in)
			change(bad.Policy.Old)
			if e := bindStateProjection(context.Background(), table, main, bad.Policy.Old); !errors.Is(e, ErrBindingMismatch) {
				t.Fatal("projection mismatch", e)
			}
		}
	}
}
func extraHistoricalKey(t *testing.T, in *Input, n byte, revoked bool) {
	k := in.Policy.Old.Keys[0]
	k.Material.PublicKeyMaterial = fixturePublic(t, wireKey(n))
	if revoked {
		k.Material.Status = "revoked"
		k.Material.HasRevocation = true
		k.Material.RevokedAt = fixtureTime(t, "2019-12-01T00:00:00Z")
		k.Material.RevocationReason = "scheduled"
		k.FirstRevokedAt = k.Material.RevokedAt
		k.DeadlinesKnown = true
		k.LastSnapshotUntil = k.FirstRevokedAt
		k.LastCompositeDeadline = k.FirstRevokedAt
	}
	in.Policy.Old.Keys = append(in.Policy.Old.Keys, k)
	sort.Slice(in.Policy.Old.Keys, func(i, j int) bool {
		return bytes.Compare(in.Policy.Old.Keys[i].Material.KeyID[:], in.Policy.Old.Keys[j].Material.KeyID[:]) < 0
	})
}
func TestPolicyBindingHistoricalKeys(t *testing.T) {
	in := policyBindingFixture(t, 1, 1)
	extraHistoricalKey(t, &in, 3, false)
	checkBindingInput(t, in)
	checkPolicyInput(t, in)
	if e := bindStateProjection(context.Background(), in.Binding.Previous, in.Binding.Main.Previous, in.Policy.Old); e != nil {
		t.Fatal(e)
	}
	m, e := Evaluate(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer m.Clear()
	if m.summary.Previous.KeyCount != 2 || m.summary.Proposed.KeyCount != 2 {
		t.Fatal("historical record dropped")
	}
	tomb := policyBindingFixture(t, 1, 0)
	extraHistoricalKey(t, &tomb, 4, true)
	id := fixturePublic(t, wireKey(4)).KeyID
	tomb.Policy.Old.RevokedIDs = []contracts.Digest{id}
	// Both metadata ledgers include the retained tombstone, while neither current key list does.
	for _, old := range []bool{true, false} {
		var comp statecomposition.Input
		if old {
			comp = tomb.Binding.Proposed
			comp.Materials = tomb.Binding.Previous
			comp.Target.Metadata = statematerials.MaterialRef{Kind: statematerials.Metadata, Payload: tomb.Binding.Main.Previous.Metadata.Mark.Material.Payload, Envelope: tomb.Binding.Main.Previous.Metadata.Mark.Material.Envelope}
			comp.Target.Desired = statematerials.MaterialRef{Kind: statematerials.Desired, Payload: tomb.Binding.Main.Previous.Desired.Mark.Material.Payload, Envelope: tomb.Binding.Main.Previous.Desired.Mark.Material.Envelope}
			comp.Target.Links = nil
		}
		if !old {
			comp = tomb.Binding.Proposed
		}
		mp := object(t, payload(t, material(&comp, comp.Target.Metadata).Raw))
		mp["revoked_key_ids"] = []any{hexDigest(id)}
		dp := object(t, payload(t, material(&comp, comp.Target.Desired).Raw))
		rewriteSelected(t, &comp, mp, dp, private(2))
		if old {
			tomb.Binding.Previous = comp.Materials
			x := declarationFor(t, comp)
			tomb.Binding.Main.Previous.Metadata = x.Metadata
			tomb.Binding.Main.Previous.Desired = x.Desired
			tomb.Binding.Main.Previous.Supervisor = x.Supervisor
			tomb.Binding.Main.Previous.LKG = x.LKG
			tomb.Policy.Old.Metadata.Payload = comp.Target.Metadata.Payload
			tomb.Policy.Old.Desired.Payload = comp.Target.Desired.Payload
		} else {
			tomb.Binding.Proposed = comp
			tomb.Binding.Main.Proposed = declarationFor(t, comp)
			tomb.Policy.Batch.MetadataEnvelope = material(&comp, comp.Target.Metadata).Raw
			tomb.Policy.Batch.DesiredEnvelope = material(&comp, comp.Target.Desired).Raw
		}
	}
	checkBindingInput(t, tomb)
	checkPolicyInput(t, tomb)
	mt, e := Evaluate(context.Background(), tomb)
	if e != nil {
		t.Fatal(e)
	}
	defer mt.Clear()
	if mt.summary.Proposed.KeyCount != 2 || mt.summary.Proposed.RevokedCount != 1 {
		t.Fatal("tombstone lost")
	}
}
func hexDigest(id contracts.Digest) string { return fmtDigest(id) }

// Keep complete envelope identity in both table occurrences and expectations.
func replaceRaw(table *statematerials.Input, old statematerials.MaterialRef, raw []byte) statematerials.MaterialRef {
	ref := old
	ref.Envelope = shaDigest(raw)
	for i := range table.Materials {
		if table.Materials[i].Ref == old {
			table.Materials[i].Ref = ref
			table.Materials[i].Raw = raw
		}
	}
	for i := range table.Expected {
		if table.Expected[i].Ref == old {
			table.Expected[i].Ref = ref
		}
	}
	return ref
}
func removeSelected(comp *statecomposition.Input, ref statematerials.MaterialRef) {
	ms := comp.Materials.Materials[:0]
	for _, m := range comp.Materials.Materials {
		if m.Ref.Kind != ref.Kind || m.Ref.Payload != ref.Payload {
			ms = append(ms, m)
		}
	}
	comp.Materials.Materials = ms
	es := comp.Materials.Expected[:0]
	for _, e := range comp.Materials.Expected {
		if e.Ref.Kind != ref.Kind || e.Ref.Payload != ref.Payload {
			es = append(es, e)
		}
	}
	comp.Materials.Expected = es
}
func rewriteSelected(t *testing.T, comp *statecomposition.Input, mp, dp map[string]any, key ed25519.PrivateKey) {
	removeSelected(comp, comp.Target.Metadata)
	removeSelected(comp, comp.Target.Desired)
	comp.Target.Metadata = addEnvelope(t, comp, statematerials.Metadata, mint(t, "metadata", mp, private(1)), statematerials.MaterialRef{}, tailRef(*comp), statematerials.MaterialRef{})
	comp.Target.Desired = addEnvelope(t, comp, statematerials.Desired, mint(t, "desired", dp, key), statematerials.MaterialRef{}, tailRef(*comp), comp.Target.Metadata)
}
func historyBindingFixture(t *testing.T, incident bool, removed bool) Input {
	in := policyBindingFixture(t, 1, 0)
	comp := &in.Binding.Proposed
	mp := object(t, payload(t, material(comp, comp.Target.Metadata).Raw))
	k := mp["keys"].([]any)[0].(map[string]any)
	k["status"] = "revoked"
	k["revoked_at"] = "2019-12-31T00:00:00Z"
	k["revocation_reason"] = "scheduled"
	if incident {
		k["revocation_reason"] = "security_incident"
		k["incident_id"] = "23456789-1234-4234-8234-123456789abc"
	}
	active := wireKey(3)
	oldCurrent := object(t, frozen(t, "metadata.payload.json"))["keys"].([]any)[0].(map[string]any)
	active["status"] = "active"
	active["not_before"] = oldCurrent["not_before"]
	active["not_after"] = oldCurrent["not_after"]
	active["revoked_at"] = nil
	active["revocation_reason"] = nil
	active["incident_id"] = nil
	keys := []any{k, active}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].(map[string]any)["key_id"].(string) < keys[j].(map[string]any)["key_id"].(string)
	})
	mp["keys"] = keys
	mp["revoked_key_ids"] = []any{k["key_id"]}
	dp := object(t, payload(t, material(comp, comp.Target.Desired).Raw))
	rewriteSelected(t, comp, mp, dp, private(3))
	in.Binding.Main.Proposed = declarationFor(t, *comp)
	in.Policy.Batch.MetadataEnvelope = material(comp, comp.Target.Metadata).Raw
	in.Policy.Batch.DesiredEnvelope = material(comp, comp.Target.Desired).Raw
	in.Policy.History = []statepolicy.KeyHistoryObservation{{Status: statepolicy.Available, ObservationID: d(86), BaseStateID: in.Policy.Old.StateID, MetadataPayload: comp.Target.Metadata.Payload, KeyID: digest(t, k["key_id"].(string)), RetiringKnown: true, DeadlinesKnown: true, LastSnapshotUntil: fixtureTime(t, "2019-12-30T00:00:00Z"), LastCompositeDeadline: fixtureTime(t, "2019-12-30T00:00:00Z")}}
	if incident {
		in.Policy.History[0].IncidentID = uuidFixture(t, k["incident_id"].(string))
		in.Policy.History[0].IncidentKind = "online_signer_equivocation"
		in.Policy.History[0].IncidentOpen = true
	}
	if removed {
		// Static old metadata has the revoked key and the replacement signer; old deadlines are independently supplied.
		oldComp := *comp
		oldComp.Materials = in.Binding.Previous
		oldComp.Target.Metadata = statematerials.MaterialRef{Kind: statematerials.Metadata, Payload: in.Binding.Main.Previous.Metadata.Mark.Material.Payload, Envelope: in.Binding.Main.Previous.Metadata.Mark.Material.Envelope}
		oldComp.Target.Desired = statematerials.MaterialRef{Kind: statematerials.Desired, Payload: in.Binding.Main.Previous.Desired.Mark.Material.Payload, Envelope: in.Binding.Main.Previous.Desired.Mark.Material.Envelope}
		oldComp.Target.Links = nil
		oldMP := object(t, encode(t, mp))
		oldMP["root_version"] = "1"
		oldMP["metadata_version"] = "6"
		oldMP["authority_sequence"] = "5"
		oldDP := object(t, payload(t, material(&oldComp, oldComp.Target.Desired).Raw))
		rewriteSelected(t, &oldComp, oldMP, oldDP, private(3))
		x := declarationFor(t, oldComp)
		in.Binding.Previous = oldComp.Materials
		in.Binding.Main.Previous = x
		in.Policy.Old.Metadata = statepolicy.HighWater{Version: x.Metadata.Mark.Version, Sequence: x.Metadata.Mark.Sequence, Payload: x.Metadata.Mark.Material.Payload}
		in.Policy.Old.Desired = statepolicy.HighWater{Version: x.Desired.Mark.Version, Sequence: x.Desired.Mark.Sequence, Payload: x.Desired.Mark.Material.Payload}
		in.Policy.Old.Keys = fixtureMetadataKeys(t, material(&oldComp, oldComp.Target.Metadata).Raw)
		in.Policy.Old.RevokedIDs = append([]contracts.Digest(nil), x.Metadata.RevokedIDs...)
		for i := range in.Policy.Old.Keys {
			if in.Policy.Old.Keys[i].Material.Status == "revoked" {
				r := &in.Policy.Old.Keys[i]
				r.DeadlinesKnown = true
				r.LastSnapshotUntil = in.Policy.History[0].LastSnapshotUntil
				r.LastCompositeDeadline = in.Policy.History[0].LastCompositeDeadline
			}
		}
		mp["keys"] = []any{active}
		rewriteSelected(t, comp, mp, dp, private(3))
		in.Binding.Main.Proposed = declarationFor(t, *comp)
		in.Policy.Batch.MetadataEnvelope = material(comp, comp.Target.Metadata).Raw
		in.Policy.Batch.DesiredEnvelope = material(comp, comp.Target.Desired).Raw
		in.Policy.History = nil
	}
	return copyPolicyBindingInput(t, in)
}
func corruptSignature(t *testing.T, raw []byte) []byte {
	o := object(t, raw)
	share := o["signatures"].([]any)[0].(map[string]any)
	sig, e := base64.RawURLEncoding.DecodeString(share["signature"].(string))
	if e != nil {
		t.Fatal(e)
	}
	sig[0] ^= 1
	share["signature"] = base64.RawURLEncoding.EncodeToString(sig)
	return encode(t, o)
}
