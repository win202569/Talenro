package nodecerttime

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/nodeissuer"
	"talenro.local/platform/internal/nodecontrol/statetimebinding"
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
	certificate, e := nodeissuer.Evaluate(ctx, o.input.Certificate)
	if certificate != nil {
		defer certificate.Clear()
	}
	if e = stageError(ctx, e); e != nil {
		return e
	}
	if certificate == nil {
		return ErrInvalidEvidence
	}
	o.summary.Certificate = certificate.Summary()
	if o.summary.Certificate.Proof != nodeissuer.NodeLeafDirectIssuerSignatureBound {
		return stageError(ctx, ErrInvalidEvidence)
	}
	timeModel, e := statetimebinding.Evaluate(ctx, o.input.Time)
	if timeModel != nil {
		defer timeModel.Clear()
	}
	if e = stageError(ctx, e); e != nil {
		return e
	}
	if timeModel == nil {
		return ErrInvalidEvidence
	}
	o.summary.Time = timeModel.Summary()
	return stageError(ctx, checkLowerProofs(o.summary))
}
func checkLowerProofs(s Summary) error {
	if s.Certificate.Proof != nodeissuer.NodeLeafDirectIssuerSignatureBound || s.Time.Status != statetimebinding.ConditionalTimeEvidenceBound {
		return ErrInvalidEvidence
	}
	return nil
}
