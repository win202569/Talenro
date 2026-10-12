package nodeleaf

import "bytes"

type owner struct {
	der, expectedSPKI []byte
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
	o.expectedSPKI = bytes.Clone(in.ExpectedSPKI)
	transferred = true
	return o
}
func (o *owner) clear() {
	if o == nil {
		return
	}
	clear(o.der)
	clear(o.expectedSPKI)
	o.der = nil
	o.expectedSPKI = nil
	o.summary = Summary{}
	o.valid = false
}
