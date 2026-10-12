package stateverify

import (
	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

type ArtifactIdentity struct{ PayloadDigest, EnvelopeDigest contracts.Digest }
type EvidenceCoordinates struct {
	ExpectedNodeID                                   uuid.UUID
	ExpectedEpoch, FinalRootVersion, MetadataVersion uint64
	DesiredGeneration, DesiredSequence               uint64
	DesiredKeyID                                     contracts.Digest
}

// SignatureEvidence owns conditional identities, not trust or authorization.
// Do not copy the owner or use accessors concurrently with Clear.
type SignatureEvidence struct {
	coordinates               EvidenceCoordinates
	anchor, metadata, desired ArtifactIdentity
	rotations                 []ArtifactIdentity
}

func (e *SignatureEvidence) Coordinates() EvidenceCoordinates {
	if e == nil {
		return EvidenceCoordinates{}
	}
	return e.coordinates
}
func (e *SignatureEvidence) Anchor() ArtifactIdentity {
	if e == nil {
		return ArtifactIdentity{}
	}
	return e.anchor
}
func (e *SignatureEvidence) Metadata() ArtifactIdentity {
	if e == nil {
		return ArtifactIdentity{}
	}
	return e.metadata
}
func (e *SignatureEvidence) Desired() ArtifactIdentity {
	if e == nil {
		return ArtifactIdentity{}
	}
	return e.desired
}
func (e *SignatureEvidence) Rotations() []ArtifactIdentity {
	if e == nil {
		return nil
	}
	return append([]ArtifactIdentity(nil), e.rotations...)
}
func (e *SignatureEvidence) Clear() {
	if e != nil {
		clear(e.rotations)
		*e = SignatureEvidence{}
	}
}
