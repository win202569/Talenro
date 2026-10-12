package statetime

import "time"

func finiteTime(t time.Time) bool { return !t.IsZero() && t.Year() >= 1 && t.Year() <= 9999 }
func evaluateFacts(in ModelInput, f *verifiedFacts) (EvidenceSummary, error) {
	now, floor := in.Time.Now, in.Time.Floor
	if !finiteTime(now) || !finiteTime(floor) || now.Before(floor) {
		return EvidenceSummary{}, ErrInvalidTime
	}
	a, m := f.att.Bindings(), f.meta.Bindings()
	k := f.key
	if a.IssuedAt.Before(k.NotBefore) || !a.IssuedAt.Before(k.NotAfter) || a.ValidUntil.After(k.NotAfter) || a.ValidUntil.After(m.ValidUntil) {
		return EvidenceSummary{}, ErrBindingMismatch
	}
	if now.Before(m.IssuedAt) || !now.Before(m.ValidUntil) || now.Before(k.NotBefore) || !now.Before(k.NotAfter) || now.Before(a.IssuedAt) || !now.Before(a.ValidUntil) {
		return EvidenceSummary{}, ErrExpired
	}
	// Sub saturates for distant dates. Compare both directions without negating MinInt64.
	if now.Sub(a.IssuedAt) > 5*time.Minute || a.IssuedAt.Sub(now) > 5*time.Minute {
		return EvidenceSummary{}, ErrExpired
	}
	if a.CheckpointSequence < in.Checkpoint.Sequence {
		return EvidenceSummary{}, ErrCheckpointRollback
	}
	return EvidenceSummary{AttestationPayload: f.att.PayloadDigest(), AttestationEnvelope: f.att.EnvelopeDigest(), MetadataPayload: f.meta.PayloadDigest(), MetadataEnvelope: f.meta.EnvelopeDigest(), KeyID: a.SigningKeyID, RequestID: in.Request.RequestID, NodeID: a.NodeID, Epoch: a.Epoch, Nonce: a.RequestNonce, OldCheckpoint: in.Checkpoint.Sequence, ProposedCheckpoint: a.CheckpointSequence, AdvanceRequired: a.CheckpointSequence > in.Checkpoint.Sequence, ValidUntil: a.ValidUntil}, nil
}
