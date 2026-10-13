package nodeissuermapping

import (
	"bytes"
	"time"
)

type owner struct {
	claims  Claims
	summary Summary
	valid   bool
}

func copySlice[T any](in []T) []T {
	if len(in) == 0 {
		return nil
	}
	out := make([]T, len(in))
	copy(out, in)
	return out
}
func copyRaw(in []byte) []byte {
	if len(in) == 0 {
		return nil
	}
	return bytes.Clone(in)
}
func normalizeTime(in time.Time) time.Time {
	if in.IsZero() {
		return time.Time{}
	}
	return in.Round(0).UTC()
}
func (o *owner) clear() {
	if o == nil {
		return
	}
	clearClaims(&o.claims)
	clearEvidence(&o.summary.Evidence)
	*o = owner{}
}
