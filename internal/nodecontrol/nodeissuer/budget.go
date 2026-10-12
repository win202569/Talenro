package nodeissuer

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
	var err error
	if len(in.Leaf.DER) > 65536 || len(in.Leaf.ExpectedSPKI) > 512 || len(in.CA.DER) > 65536 || len(in.CA.DeclaredSPKI) > 8192 {
		err = ErrLimitExceeded
	} else if len(in.Leaf.DER) == 0 || len(in.Leaf.ExpectedSPKI) == 0 || len(in.CA.DER) == 0 || len(in.CA.DeclaredSPKI) == 0 || in.Leaf.NodeID == uuid.Nil || in.Leaf.NodeID.Variant() != uuid.RFC4122 || !validDomain(in.Leaf.TrustDomain) {
		err = ErrInvalidInput
	}
	if canceled := ctx.Err(); canceled != nil {
		return canceled
	}
	return err
}
func validDomain(d string) bool {
	if len(d) < 1 || len(d) > 253 {
		return false
	}
	start := 0
	for i := 0; i <= len(d); i++ {
		if i == len(d) || d[i] == '.' {
			n := i - start
			if n < 1 || n > 63 || !lowerDigit(d[start]) || !lowerDigit(d[i-1]) {
				return false
			}
			start = i + 1
		} else if !lowerDigit(d[i]) && d[i] != '-' {
			return false
		}
	}
	return true
}
func lowerDigit(b byte) bool { return b >= 'a' && b <= 'z' || b >= '0' && b <= '9' }
