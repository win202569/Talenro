package staterecovery

import (
	"context"
	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

func checkContext(ctx context.Context) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	return ctx.Err()
}
func nonzero(d contracts.Digest) bool { return d != (contracts.Digest{}) }
func validScope(s Scope) bool {
	id, e := uuid.Parse(s.NodeID)
	return e == nil && id != uuid.Nil && id.Variant() == uuid.RFC4122 && id.String() == s.NodeID && s.AuthorityEpoch > 0 && s.AuthorityEpoch <= 1<<63-1 && nonzero(s.GuardID) && nonzero(s.SealKeyID) && nonzero(s.DeploymentID)
}
func slotAt(in RecoveryInput, id SlotID) SlotObservation {
	if id == SlotA {
		return in.A
	}
	return in.B
}
func otherSlot(id SlotID) SlotID {
	if id == SlotA {
		return SlotB
	}
	return SlotA
}
func validateSlot(s SlotObservation, scope Scope) error {
	switch s.Status {
	case SlotUnknownStatus, SlotUnavailable:
		return ErrAssuranceMissing
	case SlotCorrupt:
		return security(CorruptSlot)
	case SlotUnsupported:
		return security(UnsupportedVersion)
	case SlotAbsent:
		if s.Scope != (Scope{}) || s.StateEpoch != 0 || nonzero(s.ContentDigest) || nonzero(s.BlobDigest) || s.ModelVersion != 0 {
			return security(CorruptSlot)
		}
	case SlotIntact:
		if s.ModelVersion != 1 {
			return security(UnsupportedVersion)
		}
		if s.Scope != scope {
			return security(ScopeMismatch)
		}
		if s.StateEpoch == 0 || !nonzero(s.ContentDigest) || !nonzero(s.BlobDigest) {
			return security(CorruptSlot)
		}
	default:
		return security(CorruptSlot)
	}
	return nil
}

// RecoverModel validates both slots before selecting the unique counter match.
// DiscardFuture is diagnostic only; no state is deleted or repaired.
func RecoverModel(ctx context.Context, in RecoveryInput) (*RecoveryModel, error) {
	if e := checkContext(ctx); e != nil {
		return nil, e
	}
	if !validScope(in.ExpectedScope) {
		return nil, ErrInvalidInput
	}
	c := in.Counter
	if !nonzero(c.SourceID) || !nonzero(c.ObservationID) || !nonzero(c.ScanID) {
		return nil, ErrInvalidInput
	}
	for _, s := range [2]SlotObservation{in.A, in.B} {
		if !nonzero(s.SourceID) || !nonzero(s.ObservationID) || s.ScanID != c.ScanID {
			return nil, ErrInvalidInput
		}
	}
	switch c.Status {
	case CounterUnknown, CounterUnavailable:
		return nil, ErrAssuranceMissing
	case CounterRollback:
		return nil, security(CounterMismatch)
	case CounterAvailable:
	default:
		return nil, ErrInvalidInput
	}
	if c.Value == 0 || c.Maximum == 0 || c.Value > c.Maximum {
		return nil, ErrInvalidInput
	}
	if c.Scope != in.ExpectedScope {
		return nil, security(ScopeMismatch)
	}
	for _, s := range [2]SlotObservation{in.A, in.B} {
		if e := validateSlot(s, in.ExpectedScope); e != nil {
			return nil, e
		}
	}
	if in.A.Status == SlotIntact && in.B.Status == SlotIntact && in.A.StateEpoch == in.B.StateEpoch {
		if in.A.BlobDigest != in.B.BlobDigest || in.A.ContentDigest != in.B.ContentDigest {
			return nil, security(Fork)
		}
		return nil, security(DuplicateEpoch)
	}
	// BlobDigest names the complete sealed blob, including StateEpoch. The
	// same identity cannot consistently describe distinct epochs, regardless
	// of which one matches the counter. This checks observations, not a seal.
	if in.A.Status == SlotIntact && in.B.Status == SlotIntact && in.A.BlobDigest == in.B.BlobDigest {
		return nil, security(CorruptSlot)
	}
	selected := SlotUnknown
	if in.A.Status == SlotIntact && in.A.StateEpoch == c.Value {
		selected = SlotA
	}
	if in.B.Status == SlotIntact && in.B.StateEpoch == c.Value {
		selected = SlotB
	}
	if selected == SlotUnknown {
		return nil, security(CounterMismatch)
	}
	s := slotAt(in, selected)
	other := slotAt(in, otherSlot(selected))
	summary := RecoverySummary{Scope: in.ExpectedScope, SelectedSlot: selected, StateEpoch: s.StateEpoch, ContentDigest: s.ContentDigest, BlobDigest: s.BlobDigest}
	if other.Status == SlotIntact && other.StateEpoch > c.Value {
		if c.Value == c.Maximum || other.StateEpoch > c.Maximum || other.StateEpoch != c.Value+1 {
			return nil, security(UnexpectedFuture)
		}
		summary.DiscardFuture = true
		summary.FutureSlot = otherSlot(selected)
	}
	if e := checkContext(ctx); e != nil {
		return nil, e
	}
	in.HasPointerHint = false
	in.PointerHint = 0
	return &RecoveryModel{valid: true, snapshot: in, summary: summary}, nil
}
