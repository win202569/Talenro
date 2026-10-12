package nodeissuer

import (
	"bytes"
	"talenro.local/platform/internal/nodecontrol/nodeca"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
	"testing"
)

// Catches borrowing or sharing caller/other owned backing storage.
func TestIssuerOwnedInput(t *testing.T) {
	backing := []byte{1, 2, 3, 4, 5, 6}
	in := Input{Leaf: nodeleaf.Input{DER: backing[:4], ExpectedSPKI: backing[1:3]}, CA: nodeca.Input{DER: backing[2:5], DeclaredSPKI: backing[3:]}}
	o := ownInput(in)
	defer o.clear()
	buffers := [][]byte{o.leafDER, o.expectedSPKI, o.caDER, o.declaredSPKI}
	expected := [][]byte{{1, 2, 3, 4}, {2, 3}, {3, 4, 5}, {4, 5, 6}}
	for i, b := range buffers {
		if !bytes.Equal(b, expected[i]) {
			t.Fatal("copy", i)
		}
	}
	for i, b := range buffers {
		v := b[0]
		b[0] ^= 0xff
		for j, other := range buffers {
			if j != i && !bytes.Equal(other, expected[j]) {
				t.Fatal("owned alias", i, j)
			}
		}
		b[0] = v
	}
	clear(backing)
	for i, b := range buffers {
		if !bytes.Equal(b, expected[i]) {
			t.Fatal("caller borrowed", i)
		}
	}
	plain := contractInput().Leaf
	snapshot := o.input(plain.NodeID, plain.TrustDomain)
	if snapshot.Leaf.NodeID != plain.NodeID || snapshot.Leaf.TrustDomain != plain.TrustDomain || !bytes.Equal(snapshot.Leaf.DER, expected[0]) || !bytes.Equal(snapshot.CA.DeclaredSPKI, expected[3]) {
		t.Fatal("owned input")
	}
	if &snapshot.Leaf.DER[0] != &o.leafDER[0] || &snapshot.Leaf.ExpectedSPKI[0] != &o.expectedSPKI[0] || &snapshot.CA.DER[0] != &o.caDER[0] || &snapshot.CA.DeclaredSPKI[0] != &o.declaredSPKI[0] {
		t.Fatal("internal snapshot allocated another buffer")
	}
	o.clear()
	o.clear()
	if o.leafDER != nil || o.expectedSPKI != nil || o.caDER != nil || o.declaredSPKI != nil || o.summary != (Summary{}) || o.valid {
		t.Fatal("cleared owner")
	}
	for _, b := range buffers {
		for _, v := range b {
			if v != 0 {
				t.Fatal("reachable byte not cleared")
			}
		}
	}
	var nilOwner *owner
	nilOwner.clear()
}
