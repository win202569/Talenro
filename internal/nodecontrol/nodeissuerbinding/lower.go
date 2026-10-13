package nodeissuerbinding

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/nodeleafbinding"
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
func evaluateLower(ctx context.Context, o *owner) error {
	if e := stageError(ctx, nil); e != nil {
		return e
	}
	if o == nil {
		return ErrInvalidInput
	}
	lower, e := nodeleafbinding.Evaluate(ctx, o.input)
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
	return stageError(ctx, validateEvidence(o.input, o.summary.Evidence))
}
func validateEvidence(in Input, s nodeleafbinding.Summary) error {
	if s.Proof != nodeleafbinding.SelectedLeafProjectionConditionalTimeBound ||
		s.ObservedAt.IsZero() || s.ValidUntil.IsZero() ||
		s.ObservedAt.Year() < 1 || s.ObservedAt.Year() > 9999 ||
		s.ValidUntil.Year() < 1 || s.ValidUntil.Year() > 9999 ||
		!s.ObservedAt.Before(s.ValidUntil) ||
		s.ObservedAt != normalizeTime(s.ObservedAt) || s.ValidUntil != normalizeTime(s.ValidUntil) ||
		s.ObservedAt != s.Evidence.ObservedAt || s.ObservedAt != s.Evidence.Time.Policy.Time.Now ||
		s.ValidUntil != s.Evidence.ValidUntil {
		return ErrInvalidEvidence
	}
	main := in.Time.Policy.Binding.Main.Proposed
	expected := nodeleafbinding.Projection{NodeID: main.Scope.NodeID, Leaf: main.Identity.Leaf.Payload, PublicKey: main.Identity.PublicKey, Window: main.Identity.Window}
	if s.Proposed != expected {
		return ErrInvalidEvidence
	}
	return nil
}
