package nodeleaf

import (
	"bytes"
	"testing"
)

func TestLeafOwnedAliases(t *testing.T) {
	source := []byte{1, 2, 3, 4}
	o := ownInput(Input{DER: source[:3], ExpectedSPKI: source[1:]})
	defer o.clear()
	if o.valid || !bytes.Equal(o.der, []byte{1, 2, 3}) || !bytes.Equal(o.expectedSPKI, []byte{2, 3, 4}) {
		t.Fatal("initial owner")
	}
	source[1] = 9
	if o.der[1] != 2 || o.expectedSPKI[0] != 2 {
		t.Fatal("caller alias")
	}
	o.der[1] = 8
	if o.expectedSPKI[0] != 2 {
		t.Fatal("cross-owned alias")
	}
	der, spki := o.der, o.expectedSPKI
	o.clear()
	o.clear()
	if !bytes.Equal(der, []byte{0, 0, 0}) || !bytes.Equal(spki, []byte{0, 0, 0}) || o.summary != (Summary{}) || o.valid {
		t.Fatal("clear")
	}
	if source[1] != 9 {
		t.Fatal("cleared caller")
	}
}
