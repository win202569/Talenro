package nodeissuerbinding

import (
	"bytes"
	"context"
	"time"
)

type owner struct {
	input   Input
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
func ownInput(ctx context.Context, in Input) (result *owner, err error) {
	if e := preflight(ctx, in); e != nil {
		return nil, e
	}
	o := &owner{}
	defer func() {
		if result == nil {
			o.clear()
		}
	}()
	o.input.Certificate, err = ownCertificate(ctx, in.Certificate)
	if err != nil {
		return nil, err
	}
	o.input.Time, err = ownTimeBinding(ctx, in.Time)
	if err != nil {
		return nil, err
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	return o, nil
}
func (o *owner) clear() {
	if o == nil {
		return
	}
	clearCertificate(&o.input.Certificate)
	clearTimeBinding(&o.input.Time)
	clearEvidence(&o.summary.Evidence)
	*o = owner{}
}
