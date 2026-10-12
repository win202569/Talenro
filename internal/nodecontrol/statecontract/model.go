package statecontract

type ContractStatus uint8

const (
	ContractUnknown ContractStatus = iota
	ContractShapeValid
)

type Obligation uint8

const (
	ObligationUnknown Obligation = iota
	MaterialClosure
	DomainSemantics
	CompositionRecheck
	PhysicalCommit
	RuntimeAuthorization
)

type ContractSummary struct {
	Status      ContractStatus
	Scope       MainStateScope
	Coverage    [12]Coverage
	Outstanding [5]Obligation
}

// ContractModel owns untrusted declarations, never evidence of authorization.
// It is not safe for concurrent mutation or Clear.
type ContractModel struct {
	valid bool
	input ModelInput
}

func (m *ContractModel) Summary() ContractSummary {
	if m == nil || !m.valid {
		return ContractSummary{}
	}
	s := ContractSummary{Status: ContractShapeValid, Scope: m.input.Proposed.Scope, Outstanding: [5]Obligation{MaterialClosure, DomainSemantics, CompositionRecheck, PhysicalCommit, RuntimeAuthorization}}
	for i, h := range headers(m.input.Proposed) {
		s.Coverage[i] = h.Coverage
	}
	return s
}
func owned[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	return append([]T(nil), s...)
}
func cloneContent(x MainStateContentModel) MainStateContentModel {
	x.Root.KeyIDs = owned(x.Root.KeyIDs)
	x.Metadata.KeyIDs = owned(x.Metadata.KeyIDs)
	x.Metadata.RevokedIDs = owned(x.Metadata.RevokedIDs)
	x.ServerTrust.Agent.Deauthorized = owned(x.ServerTrust.Agent.Deauthorized)
	x.ServerTrust.Bootstrap.Deauthorized = owned(x.ServerTrust.Bootstrap.Deauthorized)
	x.Desired.SlotIDs = owned(x.Desired.SlotIDs)
	x.Recovery.Local = owned(x.Recovery.Local)
	x.Recovery.Supervisor = owned(x.Recovery.Supervisor)
	x.LKG.Entries = owned(x.LKG.Entries)
	return x
}
func cloneInput(in ModelInput) ModelInput {
	return ModelInput{cloneContent(in.Previous), cloneContent(in.Proposed), owned(in.RootTransition)}
}
func clearContent(x *MainStateContentModel) {
	clear(x.Root.KeyIDs)
	clear(x.Metadata.KeyIDs)
	clear(x.Metadata.RevokedIDs)
	clear(x.ServerTrust.Agent.Deauthorized)
	clear(x.ServerTrust.Bootstrap.Deauthorized)
	clear(x.Desired.SlotIDs)
	clear(x.Recovery.Local)
	clear(x.Recovery.Supervisor)
	clear(x.LKG.Entries)
	*x = MainStateContentModel{}
}
func (m *ContractModel) Clear() {
	if m == nil {
		return
	}
	clearContent(&m.input.Previous)
	clearContent(&m.input.Proposed)
	clear(m.input.RootTransition)
	*m = ContractModel{}
}
