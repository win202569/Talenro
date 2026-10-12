package nodecerttime

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/nodeissuer"
	"talenro.local/platform/internal/nodecontrol/statebinding"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/statetimebinding"
)

func addRaw(total, size, limit int) (int, bool) {
	if total < 0 || size < 0 || limit < 0 || total > limit || size > limit-total {
		return 0, false
	}
	return total + size, true
}
func certificateBudget(in nodeissuer.Input) (int, error) {
	if len(in.Leaf.TrustDomain) > 253 {
		return 0, ErrLimitExceeded
	}
	total := 0
	for i, raw := range [4][]byte{in.Leaf.DER, in.Leaf.ExpectedSPKI, in.CA.DER, in.CA.DeclaredSPKI} {
		if len(raw) > [4]int{65536, 512, 65536, 8192}[i] {
			return 0, ErrLimitExceeded
		}
		next, ok := addRaw(total, len(raw), 139776)
		if !ok {
			return 0, ErrLimitExceeded
		}
		total = next
	}
	return total, nil
}
func contentBudget(x statecontract.MainStateContentModel) bool {
	return len(x.Root.KeyIDs) <= 5 && len(x.Metadata.KeyIDs) <= 64 && len(x.Metadata.RevokedIDs) <= 4096 &&
		len(x.ServerTrust.Bootstrap.Deauthorized) <= 128 && len(x.ServerTrust.Agent.Deauthorized) <= 128 &&
		len(x.Desired.SlotIDs) <= 8 && len(x.LKG.Entries) <= 8 && len(x.Recovery.Local) <= 64 && len(x.Recovery.Supervisor) <= 16
}
func bindingBudget(in statebinding.Input) (int, error) {
	if len(in.Main.RootTransition) > 32 || len(in.Proposed.Target.Links) > 32 || len(in.Proposed.Anchor.Raw) > 262144 || !contentBudget(in.Main.Previous) || !contentBudget(in.Main.Proposed) {
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
			if len(m.Raw) > limit {
				return 0, ErrLimitExceeded
			}
			next, ok := addRaw(tableTotal, len(m.Raw), 16*1024*1024)
			if !ok {
				return 0, ErrLimitExceeded
			}
			tableTotal = next
			next, ok = addRaw(total, len(m.Raw), 32*1024*1024+262144)
			if !ok {
				return 0, ErrLimitExceeded
			}
			total = next
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
	total, ok := addRaw(total, len(in.Proposed.Anchor.Raw), 32*1024*1024+262144)
	if !ok {
		return 0, ErrLimitExceeded
	}
	return total, nil
}
func timeBudget(in statetimebinding.Input) (int, error) {
	total, e := bindingBudget(in.Policy.Binding)
	if e != nil {
		return 0, e
	}
	b := in.Policy.Policy.Batch
	if len(b.Rotations) > 32 {
		return 0, ErrLimitExceeded
	}
	batchTotal := 0
	add := func(raw []byte) bool {
		if len(raw) > 262144 {
			return false
		}
		next, ok := addRaw(batchTotal, len(raw), 36*262144)
		if !ok {
			return false
		}
		batchTotal = next
		next, ok = addRaw(total, len(raw), 32*1024*1024+37*262144)
		if !ok {
			return false
		}
		total = next
		return true
	}
	for _, raw := range [4][]byte{b.AnchorEnvelope, b.ActiveRootEnvelope, b.MetadataEnvelope, b.DesiredEnvelope} {
		if !add(raw) {
			return 0, ErrLimitExceeded
		}
	}
	for _, raw := range b.Rotations {
		if !add(raw) {
			return 0, ErrLimitExceeded
		}
	}
	if old := in.Policy.Policy.Old; old != nil {
		if len(old.Roots) > 33 || len(old.Keys) > 4096 || len(old.RevokedIDs) > 4096 {
			return 0, ErrLimitExceeded
		}
		for _, r := range old.Roots {
			if len(r.Root.Keys) > 5 {
				return 0, ErrLimitExceeded
			}
		}
		for _, k := range old.Keys {
			if len(k.Material.Status) > 8 || len(k.Material.RevocationReason) > 32 {
				return 0, ErrLimitExceeded
			}
		}
	}
	if len(in.Policy.Policy.History) > 4096 {
		return 0, ErrLimitExceeded
	}
	for _, h := range in.Policy.Policy.History {
		if len(h.IncidentKind) > 64 {
			return 0, ErrLimitExceeded
		}
	}
	timeTotal := 0
	for _, raw := range [2][]byte{in.Time.AttestationEnvelope, in.Time.MetadataEnvelope} {
		if len(raw) > 262144 {
			return 0, ErrLimitExceeded
		}
		next, ok := addRaw(timeTotal, len(raw), 524288)
		if !ok {
			return 0, ErrLimitExceeded
		}
		timeTotal = next
		next, ok = addRaw(total, len(raw), 43778048)
		if !ok {
			return 0, ErrLimitExceeded
		}
		total = next
	}
	return total, nil
}
func preflight(ctx context.Context, in Input) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	certTotal, e := certificateBudget(in.Certificate)
	if ce := ctx.Err(); ce != nil {
		return ce
	}
	if e != nil {
		return e
	}
	timeTotal, e := timeBudget(in.Time)
	if ce := ctx.Err(); ce != nil {
		return ce
	}
	if e != nil {
		return e
	}
	if _, ok := addRaw(certTotal, timeTotal, 43917824); !ok {
		return ErrLimitExceeded
	}
	if len(in.Certificate.Leaf.DER) == 0 || len(in.Certificate.Leaf.ExpectedSPKI) == 0 || len(in.Certificate.CA.DER) == 0 || len(in.Certificate.CA.DeclaredSPKI) == 0 || len(in.Time.Policy.Binding.Proposed.Anchor.Raw) == 0 {
		return ErrInvalidInput
	}
	return ctx.Err()
}
