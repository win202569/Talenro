package nodeleafbinding

import "testing"

func TestLeafBindingOwnedTime(t *testing.T) {
	in := ownedTestInput()
	assertOwnedUnit(t, in.Time, ownTimeBinding, clearTimeBinding)
}
