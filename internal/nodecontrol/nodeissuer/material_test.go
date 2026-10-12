package nodeissuer

import (
	"talenro.local/platform/internal/nodecontrol/nodeca"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
	"talenro.local/platform/internal/nodecontrol/nodeleafkeys"
	"testing"
)

// Catches partial results, non-shared Clear and detached-summary mutation.
func TestIssuerMaterialLifecycle(t *testing.T) {
	var nilMaterial *Material
	nilMaterial.Clear()
	if nilMaterial.Summary() != (Summary{}) {
		t.Fatal("nil")
	}
	var zero Material
	zero.Clear()
	if zero.Summary() != (Summary{}) {
		t.Fatal("zero")
	}
	o := ownInput(contractInput())
	m := &Material{owned: o}
	o.summary.Proof = NodeLeafDirectIssuerSignatureBound
	if m.Summary() != (Summary{}) {
		t.Fatal("unvalidated owner leaked")
	}
	o.valid = true
	o.summary.Leaf.Proof = nodeleafkeys.NodeLeafDeclaredKeyIdentifiersBound
	o.summary.Leaf.Leaf.Proof = nodeleaf.NodeLeafMaterialShapeValid
	o.summary.CA.Proof = nodeca.NodeCAMaterialShapeAndDeclaredKeyBound
	o.summary.LeafTBSDERDigest[0] = 9
	copyM := *m
	detached := m.Summary()
	buffers := [][]byte{o.leafDER, o.expectedSPKI, o.caDER, o.declaredSPKI}
	copyM.Clear()
	m.Clear()
	copyM.Clear()
	if m.Summary() != (Summary{}) || copyM.Summary() != (Summary{}) || o.summary != (Summary{}) || o.valid {
		t.Fatal("shared lifecycle")
	}
	if detached.Proof != NodeLeafDirectIssuerSignatureBound || detached.Leaf.Proof != nodeleafkeys.NodeLeafDeclaredKeyIdentifiersBound || detached.CA.Proof != nodeca.NodeCAMaterialShapeAndDeclaredKeyBound || detached.LeafTBSDERDigest[0] != 9 {
		t.Fatal("detached summary changed")
	}
	for _, b := range buffers {
		for _, v := range b {
			if v != 0 {
				t.Fatal("owned bytes")
			}
		}
	}
}
