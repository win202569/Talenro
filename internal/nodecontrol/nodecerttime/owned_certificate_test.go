package nodecerttime

import "testing"

func TestCertTimeOwnedCertificate(t *testing.T) {
	in := ownedTestInput()
	assertOwnedUnit(t, in.Certificate, ownCertificate, clearCertificate)
}
