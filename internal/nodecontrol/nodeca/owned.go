package nodeca

import "bytes"

type owner struct {
	der, declaredSPKI []byte
	summary           Summary
	valid             bool
}

func ownInput(in Input) *owner {
	o := &owner{}
	transferred := false
	defer func() {
		if !transferred {
			o.clear()
		}
	}()
	o.der = bytes.Clone(in.DER)
	o.declaredSPKI = bytes.Clone(in.DeclaredSPKI)
	transferred = true
	return o
}
func (o *owner) clear() {
	if o == nil {
		return
	}
	clear(o.der)
	clear(o.declaredSPKI)
	o.der = nil
	o.declaredSPKI = nil
	o.summary = Summary{}
	o.valid = false
}
