package statematerials

import (
	"bytes"
	"github.com/google/uuid"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

const maxCoordinate = uint64(9223372036854775807)

func positive(v uint64) bool { return v > 0 && v <= maxCoordinate }
func canonicalNode(s string) bool {
	if len(s) != 36 {
		return false
	}
	u, e := uuid.Parse(s)
	return e == nil && u != uuid.Nil && u.Variant() == uuid.RFC4122 && u.String() == s
}
func nonzero(d contracts.Digest) bool { return d != (contracts.Digest{}) }
func validRef(r MaterialRef) bool {
	return r.Kind >= RootSet && r.Kind <= Desired && nonzero(r.Payload) && ((r.Kind == RootSet && !nonzero(r.Envelope)) || (r.Kind != RootSet && nonzero(r.Envelope)))
}
func orderedDigests(ds []contracts.Digest) bool {
	for i, d := range ds {
		if !nonzero(d) || i > 0 && bytes.Compare(ds[i-1][:], d[:]) >= 0 {
			return false
		}
	}
	return true
}
func validSlots(ss []string) bool {
	for i, s := range ss {
		if len(s) == 0 || len(s) > 64 || i > 0 && ss[i-1] >= s {
			return false
		}
		for j, c := range []byte(s) {
			alnum := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
			if !alnum && (j == 0 || c != '.' && c != '_' && c != '-') {
				return false
			}
		}
	}
	return true
}
func validWindow(w Window, days int64) bool {
	return w.Issued >= -62135596799 && w.Until <= 253402300799 && w.Until > w.Issued && w.Until-w.Issued <= days*86400
}
func validateInput(in Input) error {
	if len(in.Materials) > 128 || len(in.Expected) > 128 {
		return ErrLimitExceeded
	}
	total, rotations := 0, 0
	for _, m := range in.Materials {
		limit := 262144
		if m.Ref.Kind == RootSet {
			limit = 65536
		}
		if len(m.Raw) > limit || len(m.Raw) > 16*1024*1024-total {
			return ErrLimitExceeded
		}
		total += len(m.Raw)
		if m.Ref.Kind == RootRotation {
			rotations++
		}
	}
	expectedRotations := 0
	for _, e := range in.Expected {
		if e.Ref.Kind == RootRotation {
			expectedRotations++
		}
		if len(e.Root.KeyIDs) > 5 || len(e.Metadata.KeyIDs) > 64 || len(e.Metadata.RevokedIDs) > 4096 || len(e.Desired.SlotIDs) > 8 {
			return ErrLimitExceeded
		}
	}
	if rotations > 32 || expectedRotations > 32 {
		return ErrLimitExceeded
	}
	if len(in.Materials) == 0 && len(in.Expected) == 0 || !canonicalNode(in.Scope.NodeID) || !positive(in.Scope.Epoch) {
		return ErrInvalidInput
	}
	materials := map[materialKey]MaterialRef{}
	for _, m := range in.Materials {
		if !validRef(m.Ref) || len(m.Raw) == 0 {
			return ErrInvalidInput
		}
		k := key(m.Ref)
		if _, ok := materials[k]; ok {
			return ErrDuplicateMaterial
		}
		materials[k] = m.Ref
	}
	expectations := map[materialKey]MaterialRef{}
	for _, e := range in.Expected {
		if !validRef(e.Ref) || !positive(e.Epoch) || e.Epoch != in.Scope.Epoch || !positive(e.Sequence) {
			return ErrInvalidInput
		}
		if !validExpectation(e) {
			return ErrInvalidInput
		}
		k := key(e.Ref)
		if _, ok := expectations[k]; ok {
			return ErrDuplicateMaterial
		}
		expectations[k] = e.Ref
	}
	for k, r := range materials {
		other, ok := expectations[k]
		if !ok {
			return ErrMissingMaterial
		}
		if r != other {
			return ErrReferenceMismatch
		}
	}
	for k := range expectations {
		if _, ok := materials[k]; !ok {
			return ErrMissingMaterial
		}
	}
	return nil
}
func validExpectation(e Expectation) bool {
	residual := e
	residual.Ref = MaterialRef{}
	residual.Epoch = 0
	residual.Sequence = 0
	switch e.Ref.Kind {
	case RootSet:
		r := e.Root
		if !positive(r.Version) || !positive(r.Threshold) || len(r.KeyIDs) == 0 || r.Threshold > uint64(len(r.KeyIDs)) || !orderedDigests(r.KeyIDs) {
			return false
		}
		residual.Root = RootExpectation{}
	case RootRotation:
		r := e.Rotation
		if !positive(r.PreviousVersion) || !positive(r.NewVersion) || !nonzero(r.OldRootPayload) || !nonzero(r.NewRootPayload) || !validWindow(r.Window, 7) {
			return false
		}
		residual.Rotation = RotationExpectation{}
	case Metadata:
		m := e.Metadata
		if !positive(m.Version) || !positive(m.RootVersion) || !nonzero(m.RootPayload) || len(m.KeyIDs) == 0 || !orderedDigests(m.KeyIDs) || !orderedDigests(m.RevokedIDs) || !validWindow(m.Window, 7) {
			return false
		}
		residual.Metadata = MetadataExpectation{}
	case Desired:
		d := e.Desired
		if !canonicalNode(d.NodeID) || !positive(d.Generation) || !positive(d.InventoryVersion) || !positive(d.ResourceVersion) || !nonzero(d.ResourcePayload) || !nonzero(d.RootPayload) || !nonzero(d.MetadataPayload) || !nonzero(d.SigningKeyID) || !validWindow(d.Window, 1) || !validSlots(d.SlotIDs) {
			return false
		}
		residual.Desired = DesiredExpectation{}
	default:
		return false
	}
	if len(residual.Root.KeyIDs) == 0 {
		residual.Root.KeyIDs = nil
	}
	if len(residual.Metadata.KeyIDs) == 0 {
		residual.Metadata.KeyIDs = nil
	}
	if len(residual.Metadata.RevokedIDs) == 0 {
		residual.Metadata.RevokedIDs = nil
	}
	if len(residual.Desired.SlotIDs) == 0 {
		residual.Desired.SlotIDs = nil
	}
	return reflect.DeepEqual(residual, Expectation{})
}
