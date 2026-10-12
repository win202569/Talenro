package nodeca

import (
	"bytes"
	"testing"
)

func TestCAOwnedInput(t *testing.T) {
	backing := []byte{1, 2, 3, 4, 5}
	in := Input{DER: backing[:3], DeclaredSPKI: backing[1:4]}
	o := ownInput(in)
	if o.valid {
		t.Fatal("premature valid")
	}
	der, spki := o.der, o.declaredSPKI
	backing[1] = 99
	if !bytes.Equal(der, []byte{1, 2, 3}) || !bytes.Equal(spki, []byte{2, 3, 4}) {
		t.Fatal("caller alias")
	}
	der[1] = 88
	if spki[0] != 2 {
		t.Fatal("clones alias")
	}
	o.clear()
	o.clear()
	if !bytes.Equal(der, []byte{0, 0, 0}) || !bytes.Equal(spki, []byte{0, 0, 0}) {
		t.Fatal("buffers not wiped")
	}
	if backing[1] != 99 || o.der != nil || o.declaredSPKI != nil || o.valid || o.summary != (Summary{}) {
		t.Fatal("clear contract")
	}
	var zero *owner
	zero.clear()
}
