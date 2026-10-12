package statecomposition

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/stateverify"
)

func identity(ref statematerials.MaterialRef) stateverify.ArtifactIdentity {
	return stateverify.ArtifactIdentity{PayloadDigest: ref.Payload, EnvelopeDigest: ref.Envelope}
}
func checkEvidence(ctx context.Context, in Input, s *selection, evidence *stateverify.SignatureEvidence) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if s == nil || evidence == nil {
		return ErrBindingMismatch
	}
	if evidence.Anchor() != (stateverify.ArtifactIdentity{PayloadDigest: in.Anchor.Payload, EnvelopeDigest: in.Anchor.Envelope}) || evidence.Metadata() != identity(in.Target.Metadata) || evidence.Desired() != identity(in.Target.Desired) {
		return ErrBindingMismatch
	}
	rotations := evidence.Rotations()
	defer clear(rotations)
	if len(rotations) != len(in.Target.Links) {
		return ErrBindingMismatch
	}
	for i, l := range in.Target.Links {
		if e := ctx.Err(); e != nil {
			return e
		}
		if rotations[i] != identity(l.Rotation) {
			return ErrBindingMismatch
		}
	}
	c := evidence.Coordinates()
	if c.ExpectedNodeID.String() != in.Anchor.Scope.NodeID || c.ExpectedEpoch != in.Anchor.Scope.Epoch || c.FinalRootVersion != s.rootVersion || c.MetadataVersion != s.metadataVersion || c.DesiredGeneration != s.desiredGeneration || c.DesiredSequence != s.desiredSequence || c.DesiredKeyID != s.desiredKeyID {
		return ErrBindingMismatch
	}
	return ctx.Err()
}
