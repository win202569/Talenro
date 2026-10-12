package statebinding

import (
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

func checkBudget(in Input) error {
	if len(in.Main.RootTransition) > 32 || len(in.Proposed.Target.Links) > 32 || len(in.Proposed.Anchor.Raw) > 262144 {
		return ErrLimitExceeded
	}
	if !contentBudget(in.Main.Previous) || !contentBudget(in.Main.Proposed) {
		return ErrLimitExceeded
	}
	total := 0
	for _, table := range [2]statematerials.Input{in.Previous, in.Proposed.Materials} {
		if len(table.Materials) > 128 || len(table.Expected) > 128 {
			return ErrLimitExceeded
		}
		tableTotal, rotations := 0, 0
		for _, m := range table.Materials {
			limit := 262144
			if m.Ref.Kind == statematerials.RootSet {
				limit = 65536
			}
			if m.Ref.Kind == statematerials.RootRotation {
				rotations++
			}
			if len(m.Raw) > limit || len(m.Raw) > 16*1024*1024-tableTotal || len(m.Raw) > 32*1024*1024+262144-total {
				return ErrLimitExceeded
			}
			tableTotal += len(m.Raw)
			total += len(m.Raw)
		}
		if rotations > 32 {
			return ErrLimitExceeded
		}
		rotations = 0
		for _, e := range table.Expected {
			if e.Ref.Kind == statematerials.RootRotation {
				rotations++
			}
			if len(e.Root.KeyIDs) > 5 || len(e.Metadata.KeyIDs) > 64 || len(e.Metadata.RevokedIDs) > 4096 || len(e.Desired.SlotIDs) > 8 {
				return ErrLimitExceeded
			}
		}
		if rotations > 32 {
			return ErrLimitExceeded
		}
	}
	if len(in.Proposed.Anchor.Raw) > 32*1024*1024+262144-total {
		return ErrLimitExceeded
	}
	if len(in.Proposed.Anchor.Raw) == 0 {
		return ErrInvalidInput
	}
	return nil
}
func contentBudget(x statecontract.MainStateContentModel) bool {
	return len(x.Root.KeyIDs) <= 5 && len(x.Metadata.KeyIDs) <= 64 && len(x.Metadata.RevokedIDs) <= 4096 &&
		len(x.ServerTrust.Bootstrap.Deauthorized) <= 128 && len(x.ServerTrust.Agent.Deauthorized) <= 128 &&
		len(x.Desired.SlotIDs) <= 8 && len(x.LKG.Entries) <= 8 && len(x.Recovery.Local) <= 64 && len(x.Recovery.Supervisor) <= 16
}
