package stateverify

import (
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
)

func TestEvidenceOwnership(t *testing.T) {
	id := ArtifactIdentity{PayloadDigest: contracts.Digest{1}, EnvelopeDigest: contracts.Digest{2}}
	e := &SignatureEvidence{coordinates: EvidenceCoordinates{ExpectedEpoch: 19, FinalRootVersion: 2, MetadataVersion: 6, DesiredGeneration: 7, DesiredSequence: 9, DesiredKeyID: contracts.Digest{3}}, anchor: id, rotations: []ArtifactIdentity{id}, metadata: id, desired: id}
	if e.Coordinates().ExpectedEpoch != 19 || e.Anchor() != id || e.Metadata() != id || e.Desired() != id || len(e.Rotations()) != 1 {
		t.Fatal("lost evidence")
	}
	c := e.Rotations()
	c[0].PayloadDigest[0] ^= 1
	if e.Rotations()[0] != id {
		t.Fatal("rotation alias")
	}
	backing := e.rotations
	e.Clear()
	if backing[0] != (ArtifactIdentity{}) {
		t.Fatal("owned evidence not cleared")
	}
	e.Clear()
	for _, z := range []*SignatureEvidence{nil, {}, e} {
		if z.Coordinates() != (EvidenceCoordinates{}) || z.Anchor() != (ArtifactIdentity{}) || z.Metadata() != (ArtifactIdentity{}) || z.Desired() != (ArtifactIdentity{}) || len(z.Rotations()) != 0 {
			t.Fatal("zero evidence")
		}
		z.Clear()
	}
}
