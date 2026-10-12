package nodecerttime

import "testing"

func TestCertTimeOwnedTime(t *testing.T) {
	in := ownedTestInput()
	assertOwnedUnit(t, in.Time, ownTimeBinding, clearTimeBinding)
}
