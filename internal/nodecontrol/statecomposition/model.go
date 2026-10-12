package statecomposition

import (
	"context"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/stateverify"
)

// CompositionModel owns one conditional target proof and the entire input.
// Do not copy it or access it concurrently with Clear.
type CompositionModel struct {
	valid       bool
	input       Input
	subset      statematerials.Summary
	coordinates stateverify.EvidenceCoordinates
	finalRoot   statematerials.MaterialRef
}

func (m *CompositionModel) Clear() {
	if m != nil {
		clearInput(&m.input)
		*m = CompositionModel{}
	}
}

type ResultStatus uint8

const (
	InvalidResult ResultStatus = iota
	ConditionalTargetVerified
)

type ProofStatus uint8

const (
	InvalidProof ProofStatus = iota
	MaterialOnly
	AnchorAssumed
	ChainDerived
	SignatureChecked
)

type Coverage struct {
	Ref    statematerials.MaterialRef
	Status ProofStatus
}
type Obligation uint8

const (
	UnknownObligation Obligation = iota
	ExternalAnchorProvenance
	UnselectedSignaturesAndHistory
	DomainAndCurrentTime
	FullStateMaterialClosure
	FullStateComposition
	PhysicalCommit
	RuntimeAuthorization
)

type Summary struct {
	Status      ResultStatus
	Scope       statematerials.Scope
	Subset      statematerials.Summary
	Anchor      stateverify.ArtifactIdentity
	Target      Target
	FinalRoot   statematerials.MaterialRef
	Coordinates stateverify.EvidenceCoordinates
	Coverage    []Coverage
	Obligations [7]Obligation
}

// Summary reports conditional coverage; unselected signatures are not checked.
func (m *CompositionModel) Summary() Summary {
	if m == nil || !m.valid {
		return Summary{}
	}
	target := m.input.Target
	target.Links = append([]RotationLink(nil), target.Links...)
	s := Summary{Status: ConditionalTargetVerified, Scope: m.input.Materials.Scope, Subset: m.subset, Anchor: stateverify.ArtifactIdentity{PayloadDigest: m.input.Anchor.Payload, EnvelopeDigest: m.input.Anchor.Envelope}, Target: target, FinalRoot: m.finalRoot, Coordinates: m.coordinates, Obligations: [7]Obligation{ExternalAnchorProvenance, UnselectedSignaturesAndHistory, DomainAndCurrentTime, FullStateMaterialClosure, FullStateComposition, PhysicalCommit, RuntimeAuthorization}}
	s.Coverage = make([]Coverage, len(m.input.Materials.Materials))
	selected := map[statematerials.MaterialRef]ProofStatus{target.AnchorRoot: AnchorAssumed, target.Metadata: SignatureChecked, target.Desired: SignatureChecked}
	for _, l := range target.Links {
		selected[l.Rotation] = SignatureChecked
		selected[l.NewRoot] = ChainDerived
	}
	for i, material := range m.input.Materials.Materials {
		status, ok := selected[material.Ref]
		if !ok {
			status = MaterialOnly
		}
		s.Coverage[i] = Coverage{Ref: material.Ref, Status: status}
	}
	return s
}

// Recheck reruns every stage, including unselected materials, then compares the
// entire normalized input. A new valid evaluation is not the old model's proof.
func Recheck(ctx context.Context, m *CompositionModel, in Input) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if m == nil || !m.valid {
		return ErrClearedModel
	}
	fresh, e := Evaluate(ctx, in)
	defer fresh.Clear()
	if e != nil {
		return e
	}
	same := reflect.DeepEqual(m.input, fresh.input)
	if e := ctx.Err(); e != nil {
		return e
	}
	if !same {
		return ErrBindingMismatch
	}
	return nil
}
