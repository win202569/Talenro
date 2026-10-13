package nodeissuerbinding

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"testing"
)

func TestIssuerBindingOwnedTable(t *testing.T) {
	in := ownedTestInput()
	assertOwnedUnit(t, in.Time.Policy.Binding.Previous, ownTable, clearTable)
}

func TestIssuerBindingOwnedTableOrdering(t *testing.T) {
	for _, sameKind := range []bool{false, true} {
		in := ownedTestInput().Time.Policy.Binding.Previous
		in.Materials[0].Ref.Kind = statematerials.Metadata
		in.Materials[1].Ref.Kind = statematerials.RootSet
		in.Expected[0].Ref.Kind = statematerials.Metadata
		in.Expected[1].Ref.Kind = statematerials.RootSet
		in.Materials[0].Ref.Payload[0] = 2
		in.Materials[1].Ref.Payload[0] = 1
		in.Expected[0].Ref.Payload[0] = 2
		in.Expected[1].Ref.Payload[0] = 1
		if sameKind {
			in.Materials[1].Ref.Kind = statematerials.Metadata
			in.Expected[1].Ref.Kind = statematerials.Metadata
		}
		in.Materials[0].Raw = []byte{20}
		in.Materials[1].Raw = []byte{10}
		out, e := ownTable(context.Background(), in)
		if e != nil {
			t.Fatal(e)
		}
		if out.Materials[0].Raw[0] != 10 || out.Materials[1].Raw[0] != 20 || out.Expected[0].Ref.Payload[0] != 1 || out.Expected[1].Ref.Payload[0] != 2 {
			t.Error("Kind/P sorting", sameKind)
		}
		clearTable(&out)
		if in.Materials[0].Raw[0] != 20 || in.Materials[1].Raw[0] != 10 {
			t.Fatal("caller order changed")
		}
	}
}
