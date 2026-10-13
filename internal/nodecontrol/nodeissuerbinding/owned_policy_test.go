package nodeissuerbinding

import "testing"

func TestIssuerBindingOwnedPolicy(t *testing.T) {
	in := ownedTestInput()
	in.Time.Policy.Policy.Batch.Rotations = [][]byte{{2}, {1}}
	in.Time.Policy.Policy.Old.Roots[0].Root.Version = 2
	in.Time.Policy.Policy.Old.Roots[1].Root.Version = 1
	in.Time.Policy.Policy.Old.Keys[0].Material.KeyID[0] = 2
	in.Time.Policy.Policy.Old.Keys[1].Material.KeyID[0] = 1
	in.Time.Policy.Policy.History[0].ObservationID[0] = 2
	in.Time.Policy.Policy.History[1].ObservationID[0] = 1
	assertOwnedUnit(t, in.Time.Policy.Policy, ownPolicy, clearPolicy)
}
