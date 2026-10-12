package nodeleafkeys

import (
	"bytes"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
	"testing"
)

func TestKeysOwnedAliases(t *testing.T) {
	backing := []byte{1, 2, 3, 4, 5, 6}
	in := Input{Leaf: nodeleaf.Input{DER: backing[:4], ExpectedSPKI: backing[1:5]}, DeclaredIssuerSPKI: backing[2:]}
	o := ownInput(in)
	defer o.clear()
	if o.valid {
		t.Fatal("premature valid")
	}
	if !bytes.Equal(o.der, in.Leaf.DER) || !bytes.Equal(o.expectedSPKI, in.Leaf.ExpectedSPKI) || !bytes.Equal(o.declaredIssuerSPKI, in.DeclaredIssuerSPKI) {
		t.Fatal("clone content")
	}
	o.der[1] = 99
	if o.expectedSPKI[0] != 2 || o.declaredIssuerSPKI[0] != 3 || backing[1] != 2 {
		t.Fatal("independent clone")
	}
	clear(backing)
	if o.expectedSPKI[0] != 2 || o.declaredIssuerSPKI[0] != 3 {
		t.Fatal("caller alias")
	}
}
