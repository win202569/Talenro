package statetime

import "context"

// EvaluateModel verifies signatures against caller-designated metadata only.
// It does not authenticate metadata roots, TLS, clocks or committed state, and
// never consumes a nonce. Repeating an unconsumed model may succeed repeatedly.
func EvaluateModel(ctx context.Context, in ModelInput) (*EvidenceModel, error) {
	f, e := verifyModel(ctx, in)
	if e != nil {
		return nil, e
	}
	defer f.Clear()
	s, e := evaluateFacts(in, f)
	if e != nil {
		return nil, e
	}
	if e = canceled(ctx); e != nil {
		return nil, e
	}
	return &EvidenceModel{ready: true, summary: s, metadata: in.Metadata, clock: in.Time, request: in.Request, checkpoint: in.Checkpoint}, nil
}

// RecheckModel re-verifies exact inputs and compares the complete observation
// snapshot, allowing only forward time from the same provider. It is not a
// commit, TLS check or nonce consumption, and does not mutate evidence.
func RecheckModel(ctx context.Context, evidence *EvidenceModel, in ModelInput) error {
	if e := canceled(ctx); e != nil {
		return e
	}
	if evidence == nil || !evidence.ready {
		return ErrStaleEvidence
	}
	fresh, e := EvaluateModel(ctx, in)
	if e != nil {
		return e
	}
	defer fresh.Clear()
	if fresh.summary != evidence.summary || fresh.metadata != evidence.metadata || fresh.request != evidence.request || fresh.checkpoint != evidence.checkpoint || fresh.clock.ProviderID != evidence.clock.ProviderID || fresh.clock.Now.Before(evidence.clock.Now) || fresh.clock.Floor.Before(evidence.clock.Floor) {
		return ErrStaleEvidence
	}
	return canceled(ctx)
}
