package nodeissuermapping

import (
	"bytes"
	"context"
	"talenro.local/platform/internal/nodecontrol/nodeissuerbinding"
)

func bindProjection(ctx context.Context, c Claims, s nodeissuerbinding.Summary) (Projection, error) {
	selected, e := validateDeclarations(ctx, c)
	if e != nil {
		return Projection{}, e
	}
	if e := ctx.Err(); e != nil {
		return Projection{}, e
	}
	if e := validateEvidence(s); e != nil {
		return Projection{}, stageError(ctx, e)
	}
	a := c.Association
	leaf := s.Evidence.Evidence.Certificate.Leaf.Leaf
	if selected.CADERDigest != s.Evidence.Evidence.Certificate.CA.CADERDigest || a.NodeID != s.Proposed.Leaf.NodeID || a.LeafDER != s.Proposed.Leaf.Leaf || a.PublicKey != s.Proposed.Leaf.PublicKey ||
		len(a.SerialBytes) != int(leaf.SerialLength) || !bytes.Equal(a.SerialBytes, leaf.Serial[:leaf.SerialLength]) {
		return Projection{}, stageError(ctx, ErrClaimConflict)
	}
	if e := ctx.Err(); e != nil {
		return Projection{}, e
	}
	return Projection{Head: c.ExpectedHead, IssuerID: a.IssuerID, CAVersion: a.CAVersion, CertificateID: a.CertificateID, IssuanceID: a.IssuanceID, Local: s.Proposed}, nil
}
