package statetimebinding

import (
	"talenro.local/platform/internal/nodecontrol/statebinding"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

func bindingBudget(in statebinding.Input) (rawTotal int, err error) {
	if len(in.Main.RootTransition) > 32 || len(in.Proposed.Target.Links) > 32 || len(in.Proposed.Anchor.Raw) > 262144 {
		return 0, ErrLimitExceeded
	}
	if !contentBudget(in.Main.Previous) || !contentBudget(in.Main.Proposed) {
		return 0, ErrLimitExceeded
	}
	total := 0
	for _, table := range [2]statematerials.Input{in.Previous, in.Proposed.Materials} {
		if len(table.Materials) > 128 || len(table.Expected) > 128 {
			return 0, ErrLimitExceeded
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
				return 0, ErrLimitExceeded
			}
			tableTotal += len(m.Raw)
			total += len(m.Raw)
		}
		if rotations > 32 {
			return 0, ErrLimitExceeded
		}
		rotations = 0
		for _, e := range table.Expected {
			if e.Ref.Kind == statematerials.RootRotation {
				rotations++
			}
			if len(e.Root.KeyIDs) > 5 || len(e.Metadata.KeyIDs) > 64 || len(e.Metadata.RevokedIDs) > 4096 || len(e.Desired.SlotIDs) > 8 {
				return 0, ErrLimitExceeded
			}
		}
		if rotations > 32 {
			return 0, ErrLimitExceeded
		}
	}
	if len(in.Proposed.Anchor.Raw) > 32*1024*1024+262144-total {
		return 0, ErrLimitExceeded
	}
	if len(in.Proposed.Anchor.Raw) == 0 {
		return 0, ErrInvalidInput
	}
	return total + len(in.Proposed.Anchor.Raw), nil
}
func contentBudget(x statecontract.MainStateContentModel) bool {
	return len(x.Root.KeyIDs) <= 5 && len(x.Metadata.KeyIDs) <= 64 && len(x.Metadata.RevokedIDs) <= 4096 &&
		len(x.ServerTrust.Bootstrap.Deauthorized) <= 128 && len(x.ServerTrust.Agent.Deauthorized) <= 128 &&
		len(x.Desired.SlotIDs) <= 8 && len(x.LKG.Entries) <= 8 && len(x.Recovery.Local) <= 64 && len(x.Recovery.Supervisor) <= 16
}

func checkBudget(in Input) error {
	total, err := bindingBudget(in.Policy.Binding)
	if err != nil {
		return err
	}
	b := in.Policy.Policy.Batch
	if len(b.Rotations) > 32 {
		return ErrLimitExceeded
	}
	batchTotal := 0
	add := func(raw []byte) bool {
		if len(raw) > 262144 || len(raw) > 36*262144-batchTotal || len(raw) > 32*1024*1024+37*262144-total {
			return false
		}
		batchTotal += len(raw)
		total += len(raw)
		return true
	}
	if !add(b.AnchorEnvelope) {
		return ErrLimitExceeded
	}
	for _, raw := range b.Rotations {
		if !add(raw) {
			return ErrLimitExceeded
		}
	}
	if !add(b.ActiveRootEnvelope) || !add(b.MetadataEnvelope) || !add(b.DesiredEnvelope) {
		return ErrLimitExceeded
	}
	if in.Policy.Policy.Old != nil {
		old := in.Policy.Policy.Old
		if len(old.Roots) > 33 || len(old.Keys) > 4096 || len(old.RevokedIDs) > 4096 {
			return ErrLimitExceeded
		}
		for _, r := range old.Roots {
			if len(r.Root.Keys) > 5 {
				return ErrLimitExceeded
			}
		}
		for _, k := range old.Keys {
			if len(k.Material.Status) > 8 || len(k.Material.RevocationReason) > 32 {
				return ErrLimitExceeded
			}
		}
	}
	if len(in.Policy.Policy.History) > 4096 {
		return ErrLimitExceeded
	}
	for _, h := range in.Policy.Policy.History {
		if len(h.IncidentKind) > 64 {
			return ErrLimitExceeded
		}
	}

	timeTotal := 0
	for _, raw := range [2][]byte{in.Time.AttestationEnvelope, in.Time.MetadataEnvelope} {
		if len(raw) > 262144 || len(raw) > 524288-timeTotal || len(raw) > 32*1024*1024+39*262144-total {
			return ErrLimitExceeded
		}
		timeTotal += len(raw)
		total += len(raw)
	}
	return nil
}
