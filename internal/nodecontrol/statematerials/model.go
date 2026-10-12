package statematerials

import (
	"bytes"
	"sort"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

type ResultStatus uint8

const (
	InvalidResult ResultStatus = iota
	SubsetConsistent
)

type CoverageState uint8

const (
	NotProvided CoverageState = iota
	Checked
)

type Coverage struct {
	Kind  MaterialKind
	State CoverageState
	Count int
}
type Obligation uint8

const (
	UnknownObligation Obligation = iota
	OtherMaterialsAndHistory
	SignaturesAndAnchor
	DomainSemantics
	CompositionRecheck
	PhysicalCommit
	RuntimeAuthorization
)

type Summary struct {
	Status        ResultStatus
	Scope         Scope
	MaterialCount int
	Coverage      [4]Coverage
	Obligations   [6]Obligation
}

func (m *MaterialSubsetModel) Summary() Summary {
	if m == nil || !m.valid {
		return Summary{}
	}
	s := Summary{Status: SubsetConsistent, Scope: m.input.Scope, MaterialCount: len(m.input.Materials), Obligations: [6]Obligation{OtherMaterialsAndHistory, SignaturesAndAnchor, DomainSemantics, CompositionRecheck, PhysicalCommit, RuntimeAuthorization}}
	for j := range s.Coverage {
		s.Coverage[j].Kind = MaterialKind(j + 1)
	}
	for _, material := range m.input.Materials {
		c := &s.Coverage[int(material.Ref.Kind)-1]
		c.Count++
		c.State = Checked
	}
	return s
}

type MaterialSubsetModel struct {
	valid bool
	input Input
}

func cloneInput(in Input) Input {
	out := in
	out.Materials = append([]Material(nil), in.Materials...)
	for i := range out.Materials {
		out.Materials[i].Raw = bytes.Clone(in.Materials[i].Raw)
	}
	out.Expected = append([]Expectation(nil), in.Expected...)
	for i := range out.Expected {
		e := &out.Expected[i]
		e.Root.KeyIDs = append([]contracts.Digest(nil), e.Root.KeyIDs...)
		e.Metadata.KeyIDs = append([]contracts.Digest(nil), e.Metadata.KeyIDs...)
		e.Metadata.RevokedIDs = append([]contracts.Digest(nil), e.Metadata.RevokedIDs...)
		e.Desired.SlotIDs = append([]string(nil), e.Desired.SlotIDs...)
	}
	sort.Slice(out.Materials, func(i, j int) bool { return refLess(out.Materials[i].Ref, out.Materials[j].Ref) })
	sort.Slice(out.Expected, func(i, j int) bool { return refLess(out.Expected[i].Ref, out.Expected[j].Ref) })
	return out
}
func refLess(a, b MaterialRef) bool {
	if a.Kind != b.Kind {
		return a.Kind < b.Kind
	}
	return bytes.Compare(a.Payload[:], b.Payload[:]) < 0
}
func clearInput(in *Input) {
	if in == nil {
		return
	}
	for i := range in.Materials {
		clear(in.Materials[i].Raw)
	}
	clear(in.Materials)
	for i := range in.Expected {
		e := &in.Expected[i]
		clear(e.Root.KeyIDs)
		clear(e.Metadata.KeyIDs)
		clear(e.Metadata.RevokedIDs)
		clear(e.Desired.SlotIDs)
	}
	clear(in.Expected)
	*in = Input{}
}
func (m *MaterialSubsetModel) Clear() {
	if m != nil {
		clearInput(&m.input)
		*m = MaterialSubsetModel{}
	}
}
