package nodeleafbinding

import "testing"

func TestLeafBindingOwnedBinding(t *testing.T) {
	in := ownedTestInput()
	assertOwnedUnit(t, in.Time.Policy.Binding, ownBinding, clearBinding)
}
