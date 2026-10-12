package nodecerttime

import "testing"

func TestCertTimeOwnedBinding(t *testing.T) {
	in := ownedTestInput()
	assertOwnedUnit(t, in.Time.Policy.Binding, ownBinding, clearBinding)
}
