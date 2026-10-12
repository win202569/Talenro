package nodeleafkeys

import (
	"context"
	"github.com/google/uuid"
)

func preflight(ctx context.Context, in Input) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(in.Leaf.DER) > 65536 || len(in.Leaf.ExpectedSPKI) > 512 || len(in.DeclaredIssuerSPKI) > 8192 {
		return ErrLimitExceeded
	}
	if len(in.Leaf.DER) == 0 || len(in.Leaf.ExpectedSPKI) == 0 || len(in.DeclaredIssuerSPKI) == 0 || in.Leaf.NodeID == uuid.Nil || in.Leaf.NodeID.Variant() != uuid.RFC4122 || !validTrustDomain(in.Leaf.TrustDomain) {
		return ErrInvalidInput
	}
	return nil
}
func validTrustDomain(value string) bool {
	if len(value) < 1 || len(value) > 253 {
		return false
	}
	alnum := func(b byte) bool { return b >= 'a' && b <= 'z' || b >= '0' && b <= '9' }
	start := 0
	for i := 0; i <= len(value); i++ {
		if i == len(value) || value[i] == '.' {
			n := i - start
			if n < 1 || n > 63 || !alnum(value[start]) || !alnum(value[i-1]) {
				return false
			}
			start = i + 1
		} else if !alnum(value[i]) && value[i] != '-' {
			return false
		}
	}
	return true
}
