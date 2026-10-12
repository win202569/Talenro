package nodecerttime

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/nodeissuer"
)

func ownCertificate(ctx context.Context, in nodeissuer.Input) (out nodeissuer.Input, err error) {
	out = in
	out.Leaf.DER = nil
	out.Leaf.ExpectedSPKI = nil
	out.CA.DER = nil
	out.CA.DeclaredSPKI = nil
	done := false
	defer func() {
		if !done {
			clearCertificate(&out)
		}
	}()
	if ctx == nil {
		return out, ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return out, e
	}
	out.Leaf.DER = copyRaw(in.Leaf.DER)
	out.Leaf.ExpectedSPKI = copyRaw(in.Leaf.ExpectedSPKI)
	out.CA.DER = copyRaw(in.CA.DER)
	out.CA.DeclaredSPKI = copyRaw(in.CA.DeclaredSPKI)
	if e := ctx.Err(); e != nil {
		return out, e
	}
	done = true
	return out, nil
}
func clearCertificate(in *nodeissuer.Input) {
	if in == nil {
		return
	}
	clear(in.Leaf.DER)
	clear(in.Leaf.ExpectedSPKI)
	clear(in.CA.DER)
	clear(in.CA.DeclaredSPKI)
	*in = nodeissuer.Input{}
}
