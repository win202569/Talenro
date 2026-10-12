package nodeissuer

import (
	"bytes"
	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/nodeca"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
)

type owner struct {
	leafDER, expectedSPKI, caDER, declaredSPKI []byte
	summary                                    Summary
	valid                                      bool
}

func ownInput(in Input) *owner {
	o := &owner{}
	transferred := false
	defer func() {
		if !transferred {
			o.clear()
		}
	}()
	o.leafDER = bytes.Clone(in.Leaf.DER)
	o.expectedSPKI = bytes.Clone(in.Leaf.ExpectedSPKI)
	o.caDER = bytes.Clone(in.CA.DER)
	o.declaredSPKI = bytes.Clone(in.CA.DeclaredSPKI)
	transferred = true
	return o
}
func (o *owner) input(nodeID uuid.UUID, trustDomain string) Input {
	return Input{Leaf: nodeleaf.Input{DER: o.leafDER, ExpectedSPKI: o.expectedSPKI, NodeID: nodeID, TrustDomain: trustDomain}, CA: nodeca.Input{DER: o.caDER, DeclaredSPKI: o.declaredSPKI}}
}
func (o *owner) clear() {
	if o == nil {
		return
	}
	clear(o.leafDER)
	clear(o.expectedSPKI)
	clear(o.caDER)
	clear(o.declaredSPKI)
	o.leafDER = nil
	o.expectedSPKI = nil
	o.caDER = nil
	o.declaredSPKI = nil
	o.summary = Summary{}
	o.valid = false
}
