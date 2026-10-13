package nodeissuermapping

import (
	"context"
	"github.com/google/uuid"
)

func addRaw(total, size, limit int) (int, bool) {
	if total < 0 || size < 0 || limit < 0 || total > limit || size > limit-total {
		return 0, false
	}
	return total + size, true
}
func claimBudget(in Claims) error {
	if len(in.Catalog.Entries) > maxRegistrationClaims || len(in.Association.IssuerID) > 128 || len(in.Association.NodeID) > 36 || len(in.Association.SerialBytes) > 20 {
		return ErrLimitExceeded
	}
	total := 0
	for _, r := range in.Catalog.Entries {
		if len(r.IssuerID) > 128 {
			return ErrLimitExceeded
		}
		next, ok := addRaw(total, len(r.IssuerID), maxDeclaredRawBytes)
		if !ok {
			return ErrLimitExceeded
		}
		total = next
	}
	for _, size := range [3]int{len(in.Association.IssuerID), len(in.Association.NodeID), len(in.Association.SerialBytes)} {
		next, ok := addRaw(total, size, maxDeclaredRawBytes)
		if !ok {
			return ErrLimitExceeded
		}
		total = next
	}
	return nil
}
func validIssuer(in string) bool {
	for i := 0; i < len(in); i++ {
		b := in[i]
		if !((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == ':' || b == '_' || b == '-') {
			return false
		}
	}
	return true
}
func validateClaimShape(in Claims) error {
	if in.ExpectedHead.Version > maxDeclaredVersion || in.Catalog.Head.Version > maxDeclaredVersion || in.Association.Head.Version > maxDeclaredVersion || in.Association.CAVersion > maxDeclaredVersion {
		return ErrInvalidInput
	}
	for _, r := range in.Catalog.Entries {
		if r.CAVersion > maxDeclaredVersion || r.State > DeclaredWithdrawn || !validIssuer(r.IssuerID) {
			return ErrInvalidInput
		}
	}
	if !validIssuer(in.Association.IssuerID) || (len(in.Association.NodeID) != 0 && len(in.Association.NodeID) != 36) {
		return ErrInvalidInput
	}
	for _, id := range [2]uuid.UUID{in.Association.CertificateID, in.Association.IssuanceID} {
		if id != uuid.Nil && id.Variant() != uuid.RFC4122 {
			return ErrInvalidInput
		}
	}
	return nil
}
func preflight(ctx context.Context, in Input) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	e := claimBudget(in.Claims)
	if ce := ctx.Err(); ce != nil {
		return ce
	}
	if e != nil {
		return e
	}
	e = validateClaimShape(in.Claims)
	if ce := ctx.Err(); ce != nil {
		return ce
	}
	return e
}
