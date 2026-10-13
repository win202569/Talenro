package nodeissuerbinding

import "testing"

func TestIssuerBindingOwnedContent(t *testing.T) {
	in := ownedTestInput()
	in.Time.Policy.Binding.Main.Previous.Root.KeyIDs[0][0] = 2
	in.Time.Policy.Binding.Main.Previous.Root.KeyIDs[1][0] = 1
	in.Time.Policy.Binding.Main.Previous.Desired.SlotIDs = []string{"b", "a"}
	in.Time.Policy.Binding.Main.Previous.LKG.Entries[0].SlotID = "b"
	in.Time.Policy.Binding.Main.Previous.LKG.Entries[1].SlotID = "a"
	in.Time.Policy.Binding.Main.Previous.Recovery.Local[0].FaultID = "b"
	in.Time.Policy.Binding.Main.Previous.Recovery.Local[1].FaultID = "a"
	assertOwnedUnit(t, in.Time.Policy.Binding.Main.Previous, ownContent, clearContent)
}
