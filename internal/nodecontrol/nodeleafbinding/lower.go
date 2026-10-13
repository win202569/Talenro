package nodeleafbinding

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/nodecerttime"
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
	lower, e := nodecerttime.Evaluate(ctx, o.input)
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
func validateEvidence(s nodecerttime.Summary) error {
	if s.Proof != nodecerttime.NodeCertificateConditionalTimeBound ||
		s.ObservedAt.IsZero() || s.ValidUntil.IsZero() ||
		s.ObservedAt.UTC().Year() < 1 || s.ObservedAt.UTC().Year() > 9999 ||
		s.ValidUntil.UTC().Year() < 1 || s.ValidUntil.UTC().Year() > 9999 ||
		!s.ObservedAt.Before(s.ValidUntil) ||
		s.ObservedAt != s.Time.Policy.Time.Now {
		return ErrInvalidEvidence
	}
	return nil
}
