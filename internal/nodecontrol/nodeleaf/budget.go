package nodeleaf

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
	if len(in.DER) > 65536 || len(in.ExpectedSPKI) > 512 {
		return ErrLimitExceeded
	}
	if len(in.DER) == 0 || len(in.ExpectedSPKI) == 0 || in.NodeID == uuid.Nil || in.NodeID.Variant() != uuid.RFC4122 || !validDomain(in.TrustDomain) {
		return ErrInvalidInput
	}
	return nil
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
func expectedURI(in Input) string {
	return "spiffe://" + in.TrustDomain + "/node/" + in.NodeID.String()
}
