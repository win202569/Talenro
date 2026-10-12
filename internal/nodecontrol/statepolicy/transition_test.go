package statepolicy

import (
	"context"
	"sort"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"testing"
)

func modelInput(t *testing.T, n int) ModelInput {
	b := signedBatch(t, n)
	f, e := loadFacts(context.Background(), b)
	wantCode(t, e, 0)
	defer f.Clear()
	return ModelInput{Batch: b, Time: modelTime(), Deployment: DeploymentObservation{Status: Available, SourceID: contracts.Digest{3}, ObservationID: contracts.Digest{4}, AnchorPayload: f.anchor.id.PayloadDigest, NodeID: f.desired.bindings.NodeID, Epoch: 19, FreshInstall: true}}
}
func previousModel(t *testing.T, in ModelInput) *StateModel {
	t.Helper()
	f, e := loadFacts(context.Background(), in.Batch)
	wantCode(t, e, 0)
	defer f.Clear()
	s := &StateModel{Status: Available, StateID: contracts.Digest{5}, CounterID: contracts.Digest{6}, DeploymentSourceID: in.Deployment.SourceID, TimeProviderID: in.Time.ProviderID, AnchorPayload: in.Deployment.AnchorPayload, NodeID: in.Deployment.NodeID, Epoch: 19, StateEpoch: 1, TimeFloor: in.Time.Floor,
		Metadata: HighWater{6, 5, f.metadata.id.PayloadDigest}, Desired: HighWater{7, 9, f.desired.id.PayloadDigest}}
	for _, a := range append([]artifactFact{f.anchor}, f.rotations...) {
		r := a.root
		r.Keys = append([]stateformat.PublicKeyMaterial(nil), r.Keys...)
		s.Roots = append(s.Roots, RootRecord{HighWater{r.Version, a.bindings.Sequence, a.id.PayloadDigest}, r})
	}
	for _, k := range f.metadata.metadata.Keys {
		s.Keys = append(s.Keys, KeyRecord{Material: k})
	}
	s.RevokedIDs = append([]contracts.Digest(nil), f.metadata.metadata.RevokedIDs...)
	t.Cleanup(s.Clear)
	return s
}
func transitionFor(t *testing.T, in ModelInput) (int, bool, error) {
	t.Helper()
	f, e := loadFacts(context.Background(), in.Batch)
	if e != nil {
		return 0, false, e
	}
	defer f.Clear()
	return checkTransition(f, in)
}
func TestPolicyDeployment(t *testing.T) {
	for _, name := range []string{"fresh", "missing", "not fresh", "wrong anchor", "wrong epoch", "wrong node", "old unavailable", "old rollback", "reset existing"} {
		t.Run(name, func(t *testing.T) {
			in := modelInput(t, 1)
			want := Code(0)
			switch name {
			case "missing":
				in.Deployment.Status = Unknown
				want = AssuranceMissing
			case "not fresh":
				in.Deployment.FreshInstall = false
				want = AssuranceMissing
			case "wrong anchor":
				in.Deployment.AnchorPayload[0] ^= 1
				want = Rejected
			case "wrong epoch":
				in.Deployment.Epoch++
				want = Rejected
			case "wrong node":
				in.Deployment.NodeID[0] ^= 1
				want = Rejected
			case "old unavailable":
				in.Old = previousModel(t, in)
				in.Old.Status = Unavailable
				want = AssuranceMissing
			case "old rollback":
				in.Old = previousModel(t, in)
				in.Old.Status = Rollback
				want = SecurityFault
			case "reset existing":
				in.Old = previousModel(t, in)
				in.Old.Epoch++
				want = Rejected
			}
			_, _, e := transitionFor(t, in)
			wantCode(t, e, want)
		})
	}
}
func TestPolicyRootPrefix(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		in := modelInput(t, n)
		in.Old = previousModel(t, in)
		a, changed, e := transitionFor(t, in)
		wantCode(t, e, 0)
		if a != n+1 || changed {
			t.Fatal("prefix/no-change")
		}
	}
	for _, name := range []string{"payload", "key", "empty", "gap", "truncated"} {
		t.Run(name, func(t *testing.T) {
			in := modelInput(t, 1)
			in.Old = previousModel(t, in)
			switch name {
			case "payload":
				in.Old.Roots[1].Mark.Payload[0] ^= 1
			case "key":
				in.Old.Roots[1].Root.Keys[0].PublicKey[0] ^= 1
			case "empty":
				in.Old.Roots = nil
			case "gap":
				in.Old.Roots[1].Mark.Version = 3
			case "truncated":
				in.Batch = signedBatch(t, 0)
			}
			_, _, e := transitionFor(t, in)
			if e == nil {
				t.Fatal("invalid prefix passed")
			}
		})
	}
	in := modelInput(t, 1)
	oldInput := modelInput(t, 0)
	in.Old = previousModel(t, oldInput)
	in.Batch.MetadataEnvelope = changePayload(t, in.Batch.MetadataEnvelope, "metadata", func(p map[string]any) { p["metadata_version"] = "7"; p["authority_sequence"] = "6" })
	a, changed, e := transitionFor(t, in)
	wantCode(t, e, 0)
	if a != 1 || !changed {
		t.Fatal("new suffix not recognized")
	}
}
func TestPolicyHighWater(t *testing.T) {
	for _, object := range []string{"metadata", "desired"} {
		for _, name := range []string{"higher", "lower", "same version", "same sequence", "payload", "cross stream"} {
			t.Run(object+"/"+name, func(t *testing.T) {
				in := modelInput(t, 1)
				in.Old = previousModel(t, in)
				m := &in.Old.Metadata
				if object == "desired" {
					m = &in.Old.Desired
				}
				want := Code(0)
				switch name {
				case "higher":
					m.Version--
					m.Sequence--
					m.Payload[0] ^= 1
				case "lower":
					m.Version++
					m.Sequence++
					m.Payload[0] ^= 1
					want = Rejected
				case "same version":
					m.Sequence--
					m.Payload[0] ^= 1
					want = SecurityFault
				case "same sequence":
					m.Version--
					m.Payload[0] ^= 1
					want = SecurityFault
				case "payload":
					m.Payload[0] ^= 1
					want = SecurityFault
				case "cross stream":
					if object == "metadata" {
						in.Old.Desired.Version--
						in.Old.Desired.Sequence = 1
						in.Old.Desired.Payload[0] ^= 1
					} else {
						in.Old.Metadata.Version--
						in.Old.Metadata.Sequence = 1
						in.Old.Metadata.Payload[0] ^= 1
					}
				}
				_, _, e := transitionFor(t, in)
				wantCode(t, e, want)
			})
		}
	}
}
func alternateMetadata(t *testing.T, raw []byte) []byte {
	env := jsonObject(t, raw)
	p := envelopePayload(t, raw)
	other := jsonObject(t, mint(t, "metadata", p, testPrivate(3)))
	shares := append(env["signatures"].([]any), other["signatures"].([]any)...)
	sort.Slice(shares, func(i, j int) bool {
		return shares[i].(map[string]any)["key_id"].(string) < shares[j].(map[string]any)["key_id"].(string)
	})
	env["signatures"] = shares
	return jsonBytes(t, env)
}
func TestPolicyEnvelopeIdentity(t *testing.T) {
	in := modelInput(t, 1)
	in.Old = previousModel(t, in)
	in.Batch.MetadataEnvelope = alternateMetadata(t, in.Batch.MetadataEnvelope)
	_, changed, e := transitionFor(t, in)
	wantCode(t, e, 0)
	if changed {
		t.Fatal("envelope-only change")
	}
	in.Batch.MetadataEnvelope = changePayload(t, in.Batch.MetadataEnvelope, "metadata", func(p map[string]any) { p["next_refresh_at"] = "2020-01-05T00:00:00Z" })
	_, _, e = transitionFor(t, in)
	wantCode(t, e, SecurityFault)
}
