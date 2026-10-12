package statetimebinding

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/statetime"
)

func bindOldCheckpoint(ctx context.Context, in Input) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	old := in.Policy.Binding.Main.Previous.Checkpoint
	next := in.Policy.Binding.Main.Proposed.Checkpoint
	obs := in.Time.Checkpoint
	if obs.Sequence != old.Sequence || obs.NodeID.String() != old.NodeID || obs.Epoch != old.AuthorityEpoch || obs.NodeID.String() != next.NodeID || obs.Epoch != next.AuthorityEpoch {
		return ErrBindingMismatch
	}
	return ctx.Err()
}
func bindEvidence(ctx context.Context, in Input, evidence statetime.EvidenceSummary) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	metadata := in.Time.Metadata
	request := in.Time.Request
	old := in.Time.Checkpoint.Sequence
	next := in.Policy.Binding.Main.Proposed.Checkpoint.Sequence
	if evidence.MetadataPayload != metadata.Payload || evidence.MetadataEnvelope != metadata.Envelope || evidence.NodeID != metadata.NodeID || evidence.Epoch != metadata.Epoch || evidence.RequestID != request.RequestID || evidence.Nonce != request.ExpectedNonce || evidence.OldCheckpoint != old || evidence.ProposedCheckpoint != next || evidence.AdvanceRequired != (next > old) {
		return ErrBindingMismatch
	}
	if evidence.ValidUntil.IsZero() {
		return ErrInvalidTime
	}
	return ctx.Err()
}
