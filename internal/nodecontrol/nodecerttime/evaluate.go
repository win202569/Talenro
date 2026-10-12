package nodecerttime

import "context"

func Evaluate(ctx context.Context, in Input) (result *Model, err error) {
	o, e := ownInput(ctx, in)
	if o != nil {
		defer func() {
			if result == nil {
				o.clear()
			}
		}()
	}
	if e = stageError(ctx, e); e != nil {
		return nil, e
	}
	if o == nil {
		return nil, ErrInvalidEvidence
	}
	if e = stageError(ctx, evaluateLower(ctx, o)); e != nil {
		return nil, e
	}
	if e = stageError(ctx, bindSnapshots(ctx, o.input, o.summary)); e != nil {
		return nil, e
	}
	until, e := checkWindows(o.summary)
	if e = stageError(ctx, e); e != nil {
		return nil, e
	}
	o.summary.ObservedAt = o.summary.Time.Policy.Time.Now
	o.summary.ValidUntil = until
	if e = stageError(ctx, nil); e != nil {
		return nil, e
	}
	o.summary.Proof = NodeCertificateConditionalTimeBound
	o.valid = true
	return &Model{owned: o}, nil
}
