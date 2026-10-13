package nodeissuermapping

import (
	"context"
	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

func completeHead(h SnapshotHead) bool {
	return h.Deployment != (contracts.Digest{}) && h.Version != 0 && h.Digest != (contracts.Digest{})
}
func validateDeclarations(ctx context.Context, c Claims) (RegistrationClaim, error) {
	if e := stageError(ctx, nil); e != nil {
		return RegistrationClaim{}, e
	}
	if !completeHead(c.ExpectedHead) || !completeHead(c.Catalog.Head) || !completeHead(c.Association.Head) {
		return RegistrationClaim{}, stageError(ctx, ErrEvidenceInsufficient)
	}
	if e := ctx.Err(); e != nil {
		return RegistrationClaim{}, e
	}
	if c.ExpectedHead != c.Catalog.Head || c.ExpectedHead != c.Association.Head {
		return RegistrationClaim{}, stageError(ctx, ErrClaimConflict)
	}
	if e := ctx.Err(); e != nil {
		return RegistrationClaim{}, e
	}
	if len(c.Catalog.Entries) == 0 {
		return RegistrationClaim{}, stageError(ctx, ErrEvidenceInsufficient)
	}
	for _, r := range c.Catalog.Entries {
		if e := ctx.Err(); e != nil {
			return RegistrationClaim{}, e
		}
		if r.IssuerID == "" || r.CAVersion == 0 || r.CADERDigest == (contracts.Digest{}) || r.State == UnknownRegistrationState {
			return RegistrationClaim{}, stageError(ctx, ErrEvidenceInsufficient)
		}
	}
	for i, r := range c.Catalog.Entries {
		if e := ctx.Err(); e != nil {
			return RegistrationClaim{}, e
		}
		for j := 0; j < i; j++ {
			other := c.Catalog.Entries[j]
			if r.IssuerID == other.IssuerID && r.CAVersion == other.CAVersion {
				return RegistrationClaim{}, stageError(ctx, ErrClaimConflict)
			}
		}
	}
	if e := ctx.Err(); e != nil {
		return RegistrationClaim{}, e
	}
	a := c.Association
	if a.CertificateID == uuid.Nil || a.IssuanceID == uuid.Nil || a.NodeID == "" || a.IssuerID == "" || a.CAVersion == 0 || a.LeafDER == (contracts.Digest{}) || a.PublicKey == (contracts.Digest{}) || len(a.SerialBytes) == 0 {
		return RegistrationClaim{}, stageError(ctx, ErrEvidenceInsufficient)
	}
	for _, r := range c.Catalog.Entries {
		if e := ctx.Err(); e != nil {
			return RegistrationClaim{}, e
		}
		if r.IssuerID == a.IssuerID && r.CAVersion == a.CAVersion {
			if r.State != DeclaredEligible {
				return RegistrationClaim{}, stageError(ctx, ErrClaimConflict)
			}
			if e := ctx.Err(); e != nil {
				return RegistrationClaim{}, e
			}
			return r, nil
		}
	}
	return RegistrationClaim{}, stageError(ctx, ErrEvidenceInsufficient)
}
