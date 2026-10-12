package nodeleafkeys

import "bytes"

type owner struct {
	der, expectedSPKI, declaredIssuerSPKI []byte
	summary                               Summary
	valid                                 bool
}

func ownInput(in Input) *owner {
	o := &owner{}
	transferred := false
	defer func() {
		if !transferred {
			o.clear()
		}
	}()
	o.der = bytes.Clone(in.Leaf.DER)
	o.expectedSPKI = bytes.Clone(in.Leaf.ExpectedSPKI)
	o.declaredIssuerSPKI = bytes.Clone(in.DeclaredIssuerSPKI)
	transferred = true
	return o
}
func (o *owner) clear() {
	if o == nil {
		return
	}
	clear(o.der)
	clear(o.expectedSPKI)
	clear(o.declaredIssuerSPKI)
	o.der = nil
	o.expectedSPKI = nil
	o.declaredIssuerSPKI = nil
	o.summary = Summary{}
	o.valid = false
}
