package nodeissuerbinding

import "testing"

func TestIssuerBindingOwnedBinding(t *testing.T) {
	in := ownedTestInput()
	assertOwnedUnit(t, in.Time.Policy.Binding, ownBinding, clearBinding)
}
