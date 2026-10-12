package statepolicy

import (
	"talenro.local/platform/internal/nodecontrol/contracts"
	"time"
)

func validInstant(v time.Time) bool {
	return !v.IsZero() && v.UTC().Year() >= 1 && v.UTC().Year() <= 9999
}
func checkedAdd(v time.Time, d time.Duration) (time.Time, bool) {
	if !validInstant(v) {
		return time.Time{}, false
	}
	end := v.Add(d)
	return end, validInstant(end)
}
func inWindow(now, start, end time.Time) bool {
	return validInstant(start) && validInstant(end) && end.After(start) && !now.Before(start) && now.Before(end)
}
func observationError(s ObservationStatus) error {
	switch s {
	case Available:
		return nil
	case Unknown, Unavailable:
		return fail(AssuranceMissing, MissingObservation)
	case Rollback:
		return fail(SecurityFault, StateRollback)
	default:
		return fail(InvalidInput, Malformed)
	}
}
func checkTime(f *verifiedFacts, o TimeObservation, old *StateModel, acceptedRoots int) (time.Time, error) {
	if f == nil || acceptedRoots < 0 || acceptedRoots > len(f.rotations)+1 {
		return time.Time{}, fail(InvalidInput, Malformed)
	}
	if e := observationError(o.Status); e != nil {
		return time.Time{}, e
	}
	if o.ProviderID == (contracts.Digest{}) || o.ObservationID == (contracts.Digest{}) {
		return time.Time{}, fail(AssuranceMissing, MissingObservation)
	}
	if !validInstant(o.Now) || !validInstant(o.Floor) {
		return time.Time{}, fail(InvalidInput, InvalidTime)
	}
	if o.Now.Before(o.Floor) {
		return time.Time{}, fail(SecurityFault, StateRollback)
	}
	if old != nil {
		if old.TimeProviderID != o.ProviderID {
			return time.Time{}, fail(AssuranceMissing, IdentityMismatch)
		}
		if !validInstant(old.TimeFloor) || o.Floor.Before(old.TimeFloor) {
			return time.Time{}, fail(SecurityFault, StateRollback)
		}
	}
	future, ok := checkedAdd(o.Now, 5*time.Minute)
	if !ok {
		return time.Time{}, fail(Rejected, InvalidTime)
	}
	for i := range f.rotations {
		if i+1 >= acceptedRoots {
			b := f.rotations[i].bindings
			if !inWindow(o.Now, b.IssuedAt, b.ValidUntil) {
				return time.Time{}, fail(Rejected, Expired)
			}
		}
	}
	m, d := f.metadata.bindings, f.desired.bindings
	if !inWindow(o.Now, m.IssuedAt, m.ValidUntil) {
		return time.Time{}, fail(Rejected, Expired)
	}
	limit, ok := checkedAdd(d.IssuedAt, 24*time.Hour)
	if !ok || !validInstant(d.ValidUntil) || !d.ValidUntil.After(d.IssuedAt) || !o.Now.Before(d.ValidUntil) || d.IssuedAt.After(future) || d.ValidUntil.After(limit) || d.ValidUntil.After(m.ValidUntil) {
		return time.Time{}, fail(Rejected, InvalidTime)
	}
	for _, k := range f.metadata.metadata.Keys {
		if k.KeyID != d.SigningKeyID {
			continue
		}
		if (k.Status != "active" && k.Status != "retiring") || !inWindow(d.IssuedAt, k.NotBefore, k.NotAfter) || d.ValidUntil.After(k.NotAfter) {
			return time.Time{}, fail(Rejected, InvalidTime)
		}
		return d.ValidUntil, nil
	}
	return time.Time{}, fail(Rejected, SignatureOrBinding)
}
