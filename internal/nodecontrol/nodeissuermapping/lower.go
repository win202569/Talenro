package nodeissuermapping

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/nodeissuerbinding"
)

func stageError(ctx context.Context, err error) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	return err
}
func evaluateLower(ctx context.Context, o *owner, material nodeissuerbinding.Input) error {
	if e := stageError(ctx, nil); e != nil {
		return e
	}
	if o == nil {
		return ErrInvalidInput
	}
	lower, e := nodeissuerbinding.Evaluate(ctx, material)
	if lower != nil {
		defer lower.Clear()
	}
	if e = stageError(ctx, e); e != nil {
		return e
	}
	if lower == nil {
		return ErrInvalidEvidence
	}
	o.summary.Evidence = lower.Summary()
	return stageError(ctx, validateEvidence(o.summary.Evidence))
}
func validateEvidence(s nodeissuerbinding.Summary) error {
	if s.Proof != nodeissuerbinding.SelectedIssuerSerialReferencesConditionalTimeBound || s.Profile != materialReferenceProfile ||
		s.ObservedAt.IsZero() || s.ValidUntil.IsZero() || s.ObservedAt.Year() < 1 || s.ObservedAt.Year() > 9999 || s.ValidUntil.Year() < 1 || s.ValidUntil.Year() > 9999 ||
		!s.ObservedAt.Before(s.ValidUntil) || s.ObservedAt != normalizeTime(s.ObservedAt) || s.ValidUntil != normalizeTime(s.ValidUntil) ||
		s.ObservedAt != s.Evidence.ObservedAt || s.ValidUntil != s.Evidence.ValidUntil ||
		s.ObservedAt != s.Evidence.Evidence.ObservedAt || s.ValidUntil != s.Evidence.Evidence.ValidUntil ||
		s.ObservedAt != s.Evidence.Evidence.Time.Policy.Time.Now || s.Proposed.Leaf != s.Evidence.Proposed {
		return ErrInvalidEvidence
	}
	if s.Evidence.Evidence.Certificate.CA.CADERDigest == (contracts.Digest{}) {
		return ErrInvalidEvidence
	}
	leaf := s.Evidence.Evidence.Certificate.Leaf.Leaf
	if leaf.SerialLength < 1 || leaf.SerialLength > 20 || leaf.Serial[0] == 0 {
		return ErrInvalidEvidence
	}
	for i := int(leaf.SerialLength); i < len(leaf.Serial); i++ {
		if leaf.Serial[i] != 0 {
			return ErrInvalidEvidence
		}
	}
	return nil
}
