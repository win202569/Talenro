package nodeissuerbinding

import "testing"

func TestIssuerBindingOwnedTime(t *testing.T) {
	in := ownedTestInput()
	assertOwnedUnit(t, in.Time, ownTimeBinding, clearTimeBinding)
}
