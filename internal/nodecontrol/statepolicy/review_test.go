package statepolicy

import (
	"context"
	"fmt"
	"sort"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
)

func TestReviewEmptyRevokedList(t *testing.T) {
	for _, check := range []string{"no change", "immediate recheck"} {
		t.Run(check, func(t *testing.T) {
			in := modelInput(t, 1)
			in.Old = previousModel(t, in)
			in.Old.RevokedIDs = make([]contracts.Digest, 0)
			c, e := EvaluateModel(context.Background(), in)
			wantCode(t, e, 0)
			defer c.Clear()
			if check == "no change" {
				if !c.NoChange() {
					t.Fatal("empty representation caused state change")
				}
			} else {
				wantCode(t, RecheckModel(context.Background(), c, in), 0)
			}
		})
	}
}

func revokedInput(t *testing.T, present bool) (ModelInput, contracts.Digest) {
	t.Helper()
	in := modelInput(t, 1)
	p := envelopePayload(t, in.Batch.MetadataEnvelope)
	extraActive(t, p)
	var kid string
	for _, v := range p["keys"].([]any) {
		k := v.(map[string]any)
		if k["key_id"] == testKeyWire(3)["key_id"] {
			kid = k["key_id"].(string)
			k["status"] = "revoked"
			k["revoked_at"] = "2019-12-01T00:00:00Z"
			k["revocation_reason"] = "scheduled"
		}
	}
	p["revoked_key_ids"] = []any{kid}
	in.Batch.MetadataEnvelope = mint(t, "metadata", p, testPrivate(1))
	old := previousModel(t, in)
	var key KeyRecord
	for i := range old.Keys {
		if old.Keys[i].Material.Status == "revoked" {
			k := &old.Keys[i]
			k.FirstRevokedAt = k.Material.RevokedAt
			k.DeadlinesKnown = true
			k.LastSnapshotUntil = stamp("2019-12-01T00:00:00Z")
			k.LastCompositeDeadline = k.LastSnapshotUntil
			key = *k
		}
	}
	if !present {
		var current []any
		for _, v := range p["keys"].([]any) {
			if v.(map[string]any)["key_id"] != kid {
				current = append(current, v)
			}
		}
		p["keys"] = current
		in.Batch.MetadataEnvelope = mint(t, "metadata", p, testPrivate(1))
	}
	in.Old = previousModel(t, in)
	if !present {
		in.Old.Keys = append(in.Old.Keys, key)
		sort.Slice(in.Old.Keys, func(i, j int) bool {
			return string(in.Old.Keys[i].Material.KeyID[:]) < string(in.Old.Keys[j].Material.KeyID[:])
		})
	} else {
		for i := range in.Old.Keys {
			if in.Old.Keys[i].Material.KeyID == key.Material.KeyID {
				in.Old.Keys[i] = key
			}
		}
	}
	return in, key.Material.KeyID
}
func TestReviewRevokedDeadlines(t *testing.T) {
	for _, present := range []bool{false, true} {
		for _, changed := range []bool{false, true} {
			for _, known := range []bool{false, true} {
				t.Run(fmt.Sprintf("present=%v/changed=%v/known=%v", present, changed, known), func(t *testing.T) {
					in, id := revokedInput(t, present)
					for i := range in.Old.Keys {
						if in.Old.Keys[i].Material.KeyID == id {
							in.Old.Keys[i].DeadlinesKnown = known
						}
					}
					f, e := loadFacts(context.Background(), in.Batch)
					wantCode(t, e, 0)
					h := KeyHistoryObservation{Status: Available, ObservationID: contracts.Digest{9}, BaseStateID: in.Old.StateID, MetadataPayload: f.metadata.id.PayloadDigest, KeyID: id, DeadlinesKnown: true, LastSnapshotUntil: stamp("2019-12-01T00:00:00Z"), LastCompositeDeadline: stamp("2019-12-01T00:00:00Z")}
					f.Clear()
					if changed {
						h.LastSnapshotUntil = stamp("2020-01-03T00:00:00Z")
						h.LastCompositeDeadline = h.LastSnapshotUntil
					}
					in.History = []KeyHistoryObservation{h}
					c, e := EvaluateModel(context.Background(), in)
					if changed {
						if c != nil {
							c.Clear()
							t.Errorf("normal candidate with contradictory deadlines, present=%v", present)
						}
						wantCode(t, e, SecurityFault)
					} else {
						wantCode(t, e, 0)
						c.Clear()
					}
				})
			}
		}
	}
}
