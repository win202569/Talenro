package statepolicy

import (
	"encoding/binary"
	"github.com/google/uuid"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
	"time"
)

func historyCase(t *testing.T) (*verifiedFacts, ModelInput) {
	in := modelInput(t, 1)
	in.Old = previousModel(t, in)
	return factsFor(t, 1), in
}
func historyResult(t *testing.T, f *verifiedFacts, in ModelInput, want Code) []KeyRecord {
	t.Helper()
	before := append([]KeyRecord(nil), in.Old.Keys...)
	keys, ids, e := checkHistory(f, in)
	wantCode(t, e, want)
	if e != nil && (keys != nil || ids != nil) {
		t.Fatal("partial history")
	}
	if !reflect.DeepEqual(before, in.Old.Keys) {
		t.Fatal("old history mutated")
	}
	return keys
}
func TestPolicyKeyHistory(t *testing.T) {
	for _, name := range []string{"same", "attributes", "reverse", "resurrect", "ledger removed", "missing key", "reappear changed", "duplicate old", "duplicate observations"} {
		t.Run(name, func(t *testing.T) {
			f, in := historyCase(t)
			want := Code(0)
			switch name {
			case "attributes":
				f.metadata.metadata.Keys[0].NotAfter = f.metadata.metadata.Keys[0].NotAfter.Add(time.Hour)
				want = SecurityFault
			case "reverse":
				in.Old.Keys[0].Material.Status = "retiring"
				want = SecurityFault
			case "resurrect":
				in.Old.RevokedIDs = []contracts.Digest{in.Old.Keys[0].Material.KeyID}
				want = SecurityFault
			case "ledger removed":
				id := contracts.Digest{1}
				in.Old.RevokedIDs = []contracts.Digest{id}
				want = SecurityFault
			case "missing key":
				f.metadata.metadata.Keys = nil
			case "reappear changed":
				in.Old.Keys[0].Material.NotBefore = in.Old.Keys[0].Material.NotBefore.Add(-time.Hour)
				want = SecurityFault
			case "duplicate old":
				in.Old.Keys = append(in.Old.Keys, in.Old.Keys[0])
				want = SecurityFault
			case "duplicate observations":
				h := deadlineObservation(f, in)
				in.History = []KeyHistoryObservation{h, h}
				want = InvalidInput
			}
			keys := historyResult(t, f, in, want)
			if name == "missing key" && len(keys) != 1 {
				t.Fatal("forgot absent key")
			}
			if e := in.Old; name == "same" {
				keys[0].Material.PublicKey[0] ^= 1
				if keys[0].Material.PublicKey == e.Keys[0].Material.PublicKey {
					t.Fatal("key output aliases")
				}
			}
		})
	}
}
func deadlineObservation(f *verifiedFacts, in ModelInput) KeyHistoryObservation {
	return KeyHistoryObservation{Status: Available, ObservationID: contracts.Digest{7}, BaseStateID: in.Old.StateID, MetadataPayload: f.metadata.id.PayloadDigest, KeyID: f.metadata.metadata.Keys[0].KeyID, RetiringKnown: true, DeadlinesKnown: true, LastSnapshotUntil: stamp("2019-12-30T00:00:00Z"), LastCompositeDeadline: stamp("2019-12-30T00:00:00Z")}
}
func revokeCase(t *testing.T) (*verifiedFacts, ModelInput) {
	f, in := historyCase(t)
	in.Old.Keys[0].Material.Status = "retiring"
	k := &f.metadata.metadata.Keys[0]
	k.Status = "revoked"
	k.HasRevocation = true
	k.RevokedAt = stamp("2019-12-31T00:00:00Z")
	k.RevocationReason = "scheduled"
	f.metadata.metadata.RevokedIDs = []contracts.Digest{k.KeyID}
	in.History = []KeyHistoryObservation{deadlineObservation(f, in)}
	return f, in
}
func TestPolicyRevocation(t *testing.T) {
	for _, name := range []string{"scheduled", "unknown deadline", "early", "jump known", "jump unknown", "wrong base", "wrong metadata", "wrong key", "bad deadline order", "emergency", "incident missing", "incident closed", "incident kind", "incident id"} {
		t.Run(name, func(t *testing.T) {
			f, in := revokeCase(t)
			want := Code(0)
			switch name {
			case "unknown deadline":
				in.History = nil
				want = AssuranceMissing
			case "early":
				in.History[0].LastCompositeDeadline = in.Time.Now
				want = Rejected
			case "jump known":
				in.Old.Keys[0].Material.Status = "active"
			case "jump unknown":
				in.Old.Keys[0].Material.Status = "active"
				in.History[0].RetiringKnown = false
				want = AssuranceMissing
			case "wrong base":
				in.History[0].BaseStateID[0] ^= 1
				want = AssuranceMissing
			case "wrong metadata":
				in.History[0].MetadataPayload[0] ^= 1
				want = AssuranceMissing
			case "wrong key":
				in.History[0].KeyID[0] ^= 1
				want = AssuranceMissing
			case "bad deadline order":
				in.History[0].LastSnapshotUntil = in.Time.Now
				want = InvalidInput
			default:
				if name != "scheduled" {
					k := &f.metadata.metadata.Keys[0]
					k.RevocationReason = "security_incident"
					k.HasIncident = true
					k.IncidentID = uuid.MustParse("23456789-1234-4234-8234-123456789abc")
					in.History[0].IncidentID = k.IncidentID
					in.History[0].IncidentKind = "online_signer_equivocation"
					in.History[0].IncidentOpen = true
					want = SecurityFault
					switch name {
					case "incident missing":
						in.History = nil
						want = AssuranceMissing
					case "incident closed":
						in.History[0].IncidentOpen = false
						want = AssuranceMissing
					case "incident kind":
						in.History[0].IncidentKind = "other"
						want = AssuranceMissing
					case "incident id":
						in.History[0].IncidentID[0] ^= 1
						want = AssuranceMissing
					}
				}
			}
			historyResult(t, f, in, want)
		})
	}
}
func TestPolicyRetention(t *testing.T) {
	for _, name := range []string{"29days", "30days", "snapshot equal", "deadline unknown"} {
		t.Run(name, func(t *testing.T) {
			f, in := historyCase(t)
			k := &in.Old.Keys[0]
			k.Material.Status = "revoked"
			k.Material.HasRevocation = true
			k.Material.RevocationReason = "scheduled"
			k.Material.RevokedAt = stamp("2019-12-01T00:00:00Z")
			k.FirstRevokedAt = k.Material.RevokedAt
			k.DeadlinesKnown = true
			k.LastSnapshotUntil = k.FirstRevokedAt
			k.LastCompositeDeadline = k.FirstRevokedAt
			in.Old.RevokedIDs = []contracts.Digest{k.Material.KeyID}
			f.metadata.metadata.RevokedIDs = append([]contracts.Digest(nil), in.Old.RevokedIDs...)
			f.metadata.metadata.Keys = nil
			in.Time.Now = k.FirstRevokedAt.Add(30 * 24 * time.Hour)
			want := Code(0)
			switch name {
			case "29days":
				in.Time.Now = in.Time.Now.Add(-time.Nanosecond)
				want = Rejected
			case "snapshot equal":
				k.LastSnapshotUntil = in.Time.Now
				k.LastCompositeDeadline = in.Time.Now
				want = Rejected
			case "deadline unknown":
				k.DeadlinesKnown = false
				want = AssuranceMissing
			}
			keys := historyResult(t, f, in, want)
			if want == 0 && (len(keys) != 1 || !keys[0].FirstRevokedAt.Equal(stamp("2019-12-01T00:00:00Z"))) {
				t.Fatal("tombstone history reset")
			}
		})
	}
	for _, n := range []int{4095, 4096, 4097} {
		f, in := historyCase(t)
		ids := make([]contracts.Digest, n)
		for i := range ids {
			binary.BigEndian.PutUint64(ids[i][24:], uint64(i+1))
		}
		in.Old.RevokedIDs = ids
		f.metadata.metadata.RevokedIDs = append([]contracts.Digest(nil), ids...)
		want := Code(0)
		if n >= 4096 {
			want = AssuranceMissing
		}
		historyResult(t, f, in, want)
	}
}
