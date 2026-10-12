package nodeleafkeys

import (
	"bytes"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
	"testing"
)

func TestKeysMaterialLifecycle(t *testing.T) {
	in := keysBudgetInput()
	original := bytes.Clone(in.Leaf.DER)
	o := ownInput(in)
	a, b, c := o.der, o.expectedSPKI, o.declaredIssuerSPKI
	o.summary = Summary{Proof: NodeLeafDeclaredKeyIdentifiersBound, Leaf: nodeleaf.Summary{Proof: nodeleaf.NodeLeafMaterialShapeValid}, SubjectKeyIdentifier: KeyIdentifier{1}}
	m := &Material{owned: o}
	if m.Summary() != (Summary{}) {
		t.Fatal("invalid owner visible")
	}
	o.valid = true
	copyM := *m
	s := m.Summary()
	s.SubjectKeyIdentifier[0] = 9
	if m.Summary().SubjectKeyIdentifier[0] != 1 {
		t.Fatal("detached")
	}
	detached := m.Summary()
	m.Clear()
	m.Clear()
	copyM.Clear()
	if copyM.Summary() != (Summary{}) || m.Summary() != (Summary{}) {
		t.Fatal("stale summary")
	}
	for _, buf := range [][]byte{a, b, c} {
		if !bytes.Equal(buf, make([]byte, len(buf))) {
			t.Fatal("not zeroed")
		}
	}
	if o.der != nil || o.expectedSPKI != nil || o.declaredIssuerSPKI != nil || o.valid || o.summary != (Summary{}) {
		t.Fatal("retained owner")
	}
	if detached.Proof != NodeLeafDeclaredKeyIdentifiersBound || detached.Leaf.Proof != nodeleaf.NodeLeafMaterialShapeValid || !bytes.Equal(in.Leaf.DER, original) {
		t.Fatal("caller/detached changed")
	}
}
func TestKeysNilMaterial(t *testing.T) {
	var m *Material
	m.Clear()
	if m.Summary() != (Summary{}) {
		t.Fatal("nil")
	}
	m = &Material{}
	m.Clear()
	m.Clear()
	if m.Summary() != (Summary{}) {
		t.Fatal("zero")
	}
	var o *owner
	o.clear()
}
