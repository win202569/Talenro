package nodeleaf

import (
	"bytes"
	"testing"
)

func TestLeafMaterialLifecycle(t *testing.T) {
	source := []byte{1, 2, 3}
	o := ownInput(Input{DER: source, ExpectedSPKI: []byte{4, 5}})
	o.valid = true
	o.summary = Summary{Proof: NodeLeafMaterialShapeValid, TrustDomain: "node.test", Serial: [20]byte{0x80}, SerialLength: 1, Pending: [6]Obligation{LeafSignatureAndIssuerTrust, KeyIdentifierDerivationAndIssuerBinding, TrustedCurrentTime, IdentityAndAuthorityProvenance, PrivateKeyPossession, MainStateAndRuntimeBinding}}
	m := &Material{owned: o}
	copied := *m
	detached := m.DER()
	s := m.Summary()
	s.Serial[0] = 7
	s.Pending[0] = UnknownObligation
	if m.Summary().Serial[0] != 0x80 || m.Summary().Pending[0] != LeafSignatureAndIssuerTrust {
		t.Fatal("summary alias")
	}
	detached[0] = 9
	if m.DER()[0] != 1 {
		t.Fatal("DER alias")
	}
	backing, spki := o.der, o.expectedSPKI
	m.Clear()
	copied.Clear()
	m.Clear()
	if copied.Summary() != (Summary{}) || copied.DER() != nil || m.Summary() != (Summary{}) {
		t.Fatal("copied wrapper stale")
	}
	if !bytes.Equal(backing, []byte{0, 0, 0}) || !bytes.Equal(spki, []byte{0, 0}) {
		t.Fatal("backing not cleared")
	}
	if detached[0] != 9 || !bytes.Equal(source, []byte{1, 2, 3}) {
		t.Fatal("external data changed")
	}
}
func TestLeafNilMaterial(t *testing.T) {
	for _, m := range []*Material{nil, {}, {owned: &owner{}}} {
		if m.Summary() != (Summary{}) || m.DER() != nil {
			t.Fatal("invalid material became evidence")
		}
		m.Clear()
		m.Clear()
	}
}
