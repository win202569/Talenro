package nodeissuerbinding

import "testing"

func TestIssuerBindingOwnedCertificate(t *testing.T) {
	in := ownedTestInput()
	assertOwnedUnit(t, in.Certificate, ownCertificate, clearCertificate)
}
