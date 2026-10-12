package statepolicybinding

import (
	"context"
	"time"
)

func bindTime(ctx context.Context, in Input) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if in.Policy.Old == nil {
		return ErrBindingMismatch
	}
	a, b := in.Binding.Main.Previous.Scope, in.Binding.Main.Proposed.Scope
	if a.TimeProviderID != b.TimeProviderID || a.TimeProviderID != in.Policy.Old.TimeProviderID || a.TimeProviderID != in.Policy.Time.ProviderID {
		return ErrBindingMismatch
	}
	f := b.TimeFloor
	if f.Present {
		if f.UnixSeconds < -62135596800 || f.UnixSeconds > 253402300799 {
			return ErrInvalidInput
		}
		constraint := time.Unix(f.UnixSeconds, 0).UTC()
		if in.Policy.Old.TimeFloor.Before(constraint) {
			return ErrBindingMismatch
		}
	}
	// Observation validity, rollback and all half-open windows remain policy decisions.
	return ctx.Err()
}
