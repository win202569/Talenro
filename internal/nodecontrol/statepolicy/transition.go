package statepolicy

import (
	"math"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

func validMark(m HighWater) bool {
	return m.Version > 0 && m.Version <= math.MaxInt64 && m.Sequence > 0 && m.Sequence <= math.MaxInt64 && m.Payload != (contracts.Digest{})
}
func compareMark(old, next HighWater) (bool, error) {
	if !validMark(old) || !validMark(next) {
		return false, fail(SecurityFault, Malformed)
	}
	if old == next {
		return false, nil
	}
	if old.Version == next.Version || old.Sequence == next.Sequence || old.Payload == next.Payload {
		return false, fail(SecurityFault, ContentConflict)
	}
	if next.Version < old.Version || next.Sequence < old.Sequence {
		return false, fail(Rejected, StateRollback)
	}
	return true, nil
}
func checkTransition(f *verifiedFacts, in ModelInput) (int, bool, error) {
	if f == nil {
		return 0, false, fail(InvalidInput, Malformed)
	}
	d := in.Deployment
	if e := observationError(d.Status); e != nil {
		return 0, false, e
	}
	if d.SourceID == (contracts.Digest{}) || d.ObservationID == (contracts.Digest{}) {
		return 0, false, fail(AssuranceMissing, MissingObservation)
	}
	if d.AnchorPayload != f.anchor.id.PayloadDigest || d.NodeID != f.desired.bindings.NodeID || d.Epoch != f.anchor.bindings.Epoch {
		return 0, false, fail(Rejected, IdentityMismatch)
	}
	old := in.Old
	if old == nil {
		if !d.FreshInstall {
			return 0, false, fail(AssuranceMissing, MissingObservation)
		}
		return 0, true, nil
	}
	if e := observationError(old.Status); e != nil {
		return 0, false, e
	}
	if old.NodeID != d.NodeID || old.Epoch != d.Epoch || old.DeploymentSourceID != d.SourceID || old.AnchorPayload != d.AnchorPayload {
		return 0, false, fail(Rejected, IdentityMismatch)
	}
	if len(old.Roots) > 33 {
		return 0, false, fail(AssuranceMissing, UnsupportedRange)
	}
	if old.StateID == (contracts.Digest{}) || old.CounterID == (contracts.Digest{}) || old.StateEpoch == 0 || old.StateEpoch > math.MaxInt64 || len(old.Roots) == 0 {
		return 0, false, fail(SecurityFault, Malformed)
	}
	if len(old.Roots) > len(f.rotations)+1 {
		return 0, false, fail(Rejected, StateRollback)
	}
	for i, r := range old.Roots {
		a := &f.anchor
		if i > 0 {
			a = &f.rotations[i-1]
		}
		if !validMark(r.Mark) || r.Mark.Version != uint64(i+1) || r.Root.Version != r.Mark.Version {
			return 0, false, fail(SecurityFault, Malformed)
		}
		if r.Mark.Sequence != a.bindings.Sequence || r.Mark.Payload != a.id.PayloadDigest || !reflect.DeepEqual(r.Root, a.root) {
			return 0, false, fail(SecurityFault, ContentConflict)
		}
	}
	mc, e := compareMark(old.Metadata, HighWater{f.metadata.bindings.MetadataVersion, f.metadata.bindings.Sequence, f.metadata.id.PayloadDigest})
	if e != nil {
		return 0, false, e
	}
	dc, e := compareMark(old.Desired, HighWater{f.desired.bindings.Generation, f.desired.bindings.Sequence, f.desired.id.PayloadDigest})
	if e != nil {
		return 0, false, e
	}
	return len(old.Roots), mc || dc || len(old.Roots) < len(f.rotations)+1, nil
}
