package nodeleafbinding

import "testing"

func TestLeafBindingOwnedCertificate(t *testing.T) {
	in := ownedTestInput()
	assertOwnedUnit(t, in.Certificate, ownCertificate, clearCertificate)
}
