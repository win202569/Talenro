package statematerials

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sort"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
)

func replaceRoot(t *testing.T, in *Input, index int, mut func(map[string]any)) {
	w := obj(t, in.Materials[index].Raw)
	mut(w)
	raw := canon(t, w)
	r := in.Materials[index].Ref
	r.Payload = sha256.Sum256(raw)
	in.Materials[index] = Material{r, raw}
	in.Expected[index].Ref = r
}
func TestSubsetRelations(t *testing.T) {
	for _, tc := range []struct {
		name string
		mut  func(*Input)
		want error
	}{
		{"missing old", func(i *Input) { i.Materials = i.Materials[1:]; i.Expected = i.Expected[1:] }, ErrMissingMaterial},
		{"missing new", func(i *Input) {
			i.Materials = append(i.Materials[:1], i.Materials[2:]...)
			i.Expected = append(i.Expected[:1], i.Expected[2:]...)
		}, ErrMissingMaterial},
		{"metadata missing root", func(i *Input) { i.Expected[3].Metadata.RootPayload[0] ^= 1 }, ErrMissingMaterial},
		{"metadata root version", func(i *Input) { i.Expected[3].Metadata.RootPayload = i.Expected[0].Ref.Payload }, ErrReferenceMismatch},
		{"desired missing metadata", func(i *Input) { i.Expected[4].Desired.MetadataPayload[0] ^= 1 }, ErrMissingMaterial},
		{"desired root conflict", func(i *Input) { i.Expected[4].Desired.RootPayload = i.Expected[0].Ref.Payload }, ErrReferenceMismatch},
		{"desired scope", func(i *Input) { i.Scope.NodeID = "22345678-1234-4234-8234-123456789abc" }, ErrReferenceMismatch},
		{"embedded bytes", func(i *Input) {
			replaceRoot(t, i, 1, func(w map[string]any) { w["authority_sequence"] = "6" })
			i.Expected[1].Sequence = 6
			i.Expected[2].Rotation.NewRootPayload = i.Expected[1].Ref.Payload
			i.Expected[3].Metadata.RootPayload = i.Expected[1].Ref.Payload
			i.Expected[4].Desired.RootPayload = i.Expected[1].Ref.Payload
		}, ErrReferenceMismatch},
		{"old sequence", func(i *Input) {
			replaceRoot(t, i, 0, func(w map[string]any) { w["authority_sequence"] = "5" })
			i.Expected[0].Sequence = 5
			i.Expected[2].Rotation.OldRootPayload = i.Expected[0].Ref.Payload
		}, ErrReferenceMismatch},
		{"duplicate version", func(i *Input) {
			fresh := subsetFixture(t)
			replaceRoot(t, &fresh, 1, func(w map[string]any) { w["authority_sequence"] = "6" })
			fresh.Expected[1].Sequence = 6
			i.Materials = append(i.Materials, fresh.Materials[1])
			i.Expected = append(i.Expected, fresh.Expected[1])
		}, ErrReferenceMismatch},
	} {
		t.Run(tc.name, func(t *testing.T) { in := subsetFixture(t); tc.mut(&in); expectError(t, in, tc.want) })
	}
	in := subsetFixture(t)
	rewrapMaterial(t, &in, 4, func(w map[string]any) { w["node_id"] = "22345678-1234-4234-8234-123456789abc" })
	in.Expected[4].Desired.NodeID = "22345678-1234-4234-8234-123456789abc"
	expectError(t, in, ErrReferenceMismatch)
	// Duplicate edge with a different rotation payload must not become a second accepted edge.
	in = subsetFixture(t)
	other := subsetFixture(t)
	rewrapMaterial(t, &other, 2, func(w map[string]any) { w["valid_until"] = "2026-10-15T00:00:00Z" })
	other.Expected[2].Rotation.Window.Until -= 86400
	in.Materials = append(in.Materials, other.Materials[2])
	in.Expected = append(in.Expected, other.Expected[2])
	expectError(t, in, ErrReferenceMismatch)
	// Signing key in envelope can be valid format but absent from selected metadata.
	in = subsetFixture(t)
	w := obj(t, in.Materials[4].Raw)
	w["key_id"] = fmt.Sprintf("%064x", 1)
	in.Materials[4].Raw = canon(t, w)
	r := envelopeRef(t, Desired, in.Materials[4].Raw)
	in.Materials[4].Ref = r
	in.Expected[4].Ref = r
	in.Expected[4].Desired.SigningKeyID = digestText(t, w["key_id"].(string))
	expectError(t, in, ErrReferenceMismatch)
}
func TestSubsetHistory(t *testing.T) {
	in := subsetFixture(t)
	replaceRoot(t, &in, 0, func(w map[string]any) { w["root_version"] = "100" })
	in.Expected[0].Root.Version = 100
	standalone := Input{Scope: in.Scope, Materials: in.Materials[:1], Expected: in.Expected[:1]}
	m, e := Evaluate(testContext(), standalone)
	if e != nil {
		t.Fatal("high historical root", e)
	}
	m.Clear()
	full := subsetFixture(t)
	full.Materials = append(full.Materials, standalone.Materials[0])
	full.Expected = append(full.Expected, standalone.Expected[0])
	m, e = Evaluate(testContext(), full)
	if e != nil {
		t.Fatal("disconnected historical root", e)
	}
	m.Clear()
	// New metadata revokes the old key but retains an active replacement. Old desired remains bound to old metadata.
	newer := subsetFixture(t)
	rewrapMaterial(t, &newer, 3, func(w map[string]any) {
		w["metadata_version"] = "7"
		w["authority_sequence"] = "6"
		keys := w["keys"].([]any)
		old := keys[0].(map[string]any)
		old["status"] = "revoked"
		old["revoked_at"] = "2026-10-09T00:00:00Z"
		old["revocation_reason"] = "scheduled"
		var pub [32]byte
		pub[0] = 2
		id := fmt.Sprintf("%x", sha256.Sum256(pub[:]))
		next := obj(t, canon(t, old))
		next["key_id"] = id
		next["public_key"] = base64.RawURLEncoding.EncodeToString(pub[:])
		next["status"] = "active"
		next["revoked_at"] = nil
		next["revocation_reason"] = nil
		keys = append(keys, next)
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].(map[string]any)["key_id"].(string) < keys[j].(map[string]any)["key_id"].(string)
		})
		w["keys"] = keys
		w["revoked_key_ids"] = []string{old["key_id"].(string)}
	})
	outer := obj(t, newer.Materials[3].Raw)
	outer["authority_sequence"] = "6"
	newer.Materials[3].Raw = canon(t, outer)
	newer.Materials[3].Ref = envelopeRef(t, Metadata, newer.Materials[3].Raw)
	newer.Expected[3].Ref = newer.Materials[3].Ref
	newer.Expected[3].Sequence = 6
	newer.Expected[3].Metadata.Version = 7
	var pub [32]byte
	pub[0] = 2
	newer.Expected[3].Metadata.KeyIDs = append(newer.Expected[3].Metadata.KeyIDs, sha256.Sum256(pub[:]))
	sort.Slice(newer.Expected[3].Metadata.KeyIDs, func(i, j int) bool {
		return fmt.Sprintf("%x", newer.Expected[3].Metadata.KeyIDs[i]) < fmt.Sprintf("%x", newer.Expected[3].Metadata.KeyIDs[j])
	})
	newer.Expected[3].Metadata.RevokedIDs = full.Expected[3].Metadata.KeyIDs
	full.Materials = append(full.Materials, newer.Materials[3])
	full.Expected = append(full.Expected, newer.Expected[3])
	m, e = Evaluate(testContext(), full)
	if e != nil {
		t.Fatal("new revocation erased history", e)
	}
	m.Clear()
}
func TestSubsetReferenceIdentity(t *testing.T) {
	in := subsetFixture(t)
	in.Expected[2].Rotation.NewRootPayload = in.Expected[2].Ref.Payload
	expectError(t, in, ErrMissingMaterial)
	in = subsetFixture(t)
	in.Materials = append(in.Materials, in.Materials[4])
	in.Expected = append(in.Expected, in.Expected[4])
	in.Expected[5].Desired.RootPayload = in.Expected[0].Ref.Payload
	expectError(t, in, ErrDuplicateMaterial)
	in = subsetFixture(t)
	in.Expected[4].Desired.ResourcePayload[0] ^= 1
	expectError(t, in, ErrReferenceMismatch)
	in = subsetFixture(t)
	in.Expected[4].Desired.InventoryVersion++
	expectError(t, in, ErrReferenceMismatch)
}

// Protect the three coverage gaps identified by the one fresh source review.
func TestSubsetReviewCoverage(t *testing.T) {
	t.Run("same coordinates different embedded root", func(t *testing.T) {
		in := subsetFixture(t)
		var pub [32]byte
		pub[0] = 3
		id := sha256.Sum256(pub[:])
		replaceRoot(t, &in, 1, func(w map[string]any) {
			w["keys"] = []any{map[string]any{"algorithm": "ed25519", "key_id": fmt.Sprintf("%x", id), "public_key": base64.RawURLEncoding.EncodeToString(pub[:])}}
		})
		in.Expected[1].Root.KeyIDs[0] = id
		in.Expected[2].Rotation.NewRootPayload = in.Expected[1].Ref.Payload
		in.Expected[3].Metadata.RootPayload = in.Expected[1].Ref.Payload
		in.Expected[4].Desired.RootPayload = in.Expected[1].Ref.Payload
		expectError(t, in, ErrReferenceMismatch)
	})
	t.Run("expected slot mismatch", func(t *testing.T) {
		in := populatedFixture(t)
		in.Expected[4].Desired.SlotIDs[0] = "slot-b"
		expectError(t, in, ErrReferenceMismatch)
	})
	t.Run("expired disconnected rotation fragments", func(t *testing.T) {
		in := subsetFixture(t)
		week := Window{1577836800, 1578441600}
		day := Window{1577836800, 1577923200}
		rewrapMaterial(t, &in, 2, func(w map[string]any) {
			w["issued_at"] = "2020-01-01T00:00:00Z"
			w["valid_until"] = "2020-01-08T00:00:00Z"
		})
		in.Expected[2].Rotation.Window = week
		rewrapMaterial(t, &in, 3, func(w map[string]any) {
			w["issued_at"] = "2020-01-01T00:00:00Z"
			w["valid_until"] = "2020-01-08T00:00:00Z"
			w["next_refresh_at"] = "2020-01-06T00:00:00Z"
			for _, k := range w["keys"].([]any) {
				key := k.(map[string]any)
				key["not_before"] = "2020-01-01T00:00:00Z"
				key["not_after"] = "2020-01-08T00:00:00Z"
			}
		})
		in.Expected[3].Metadata.Window = week
		in.Expected[4].Desired.MetadataPayload = in.Expected[3].Ref.Payload
		rewrapMaterial(t, &in, 4, func(w map[string]any) {
			w["issued_at"] = "2020-01-01T00:00:00Z"
			w["valid_until"] = "2020-01-02T00:00:00Z"
		})
		in.Expected[4].Desired.Window = day
		keyID := in.Expected[0].Root.KeyIDs[0]
		var roots []MaterialRef
		for _, v := range []uint64{100, 101} {
			w := obj(t, rawFixture(t, "root-initial.payload"))
			w["root_version"] = fmt.Sprint(v)
			w["authority_sequence"] = fmt.Sprint(v)
			raw := canon(t, w)
			ref := MaterialRef{Kind: RootSet, Payload: sha256.Sum256(raw)}
			roots = append(roots, ref)
			in.Materials = append(in.Materials, Material{ref, raw})
			in.Expected = append(in.Expected, Expectation{Ref: ref, Epoch: 19, Sequence: v, Root: RootExpectation{v, 1, []contracts.Digest{keyID}}})
		}
		w := obj(t, rawFixture(t, "root-rotation.payload"))
		w["previous_root_version"] = "100"
		w["authority_sequence"] = "101"
		w["new_root_set"] = obj(t, in.Materials[6].Raw)
		w["issued_at"] = "2020-01-01T00:00:00Z"
		w["valid_until"] = "2020-01-08T00:00:00Z"
		payload := canon(t, w)
		outer := obj(t, rawFixture(t, "root-rotation.envelope"))
		outer["authority_sequence"] = "101"
		outer["root_version"] = "101"
		outer["previous_root_version"] = "100"
		outer["payload_b64url"] = base64.RawURLEncoding.EncodeToString(payload)
		outer["payload_sha256"] = fmt.Sprintf("%x", sha256.Sum256(payload))
		raw := canon(t, outer)
		ref := envelopeRef(t, RootRotation, raw)
		in.Materials = append(in.Materials, Material{ref, raw})
		in.Expected = append(in.Expected, Expectation{Ref: ref, Epoch: 19, Sequence: 101, Rotation: RotationExpectation{100, 101, roots[0].Payload, roots[1].Payload, week}})
		m, e := Evaluate(testContext(), in)
		if e != nil {
			t.Fatal("historical fragments require invented closure/freshness", e)
		}
		m.Clear()
	})
}
