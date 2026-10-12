package statebinding

import (
	"context"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/statecomposition"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

// BindingModel owns untrusted declarations and a conditional projection result.
// It must not be copied or used concurrently with Clear.
type BindingModel struct {
	valid           bool
	input           Input
	contractSummary statecontract.ContractSummary
	previousSummary statematerials.Summary
	proposedSummary statecomposition.Summary
}

func clearCompositionSummary(s *statecomposition.Summary) {
	if s == nil {
		return
	}
	clear(s.Target.Links)
	clear(s.Coverage)
	*s = statecomposition.Summary{}
}
func (model *BindingModel) Clear() {
	if model == nil {
		return
	}
	clearInput(&model.input)
	clearCompositionSummary(&model.proposedSummary)
	*model = BindingModel{}
}

type ResultStatus uint8

const (
	InvalidResult ResultStatus = iota
	ConditionalProjectionBound
)

type Obligation uint8

const (
	UnknownObligation Obligation = iota
	ExternalAnchorAndPlatformProvenance
	PreviousSignaturesAndUnselectedHistory
	UnverifiedHistoryAndSetDigests
	DomainAndCurrentTime
	FullStateMaterialClosure
	FullStateComposition
	PhysicalCommit
	RuntimeAuthorization
)

type Summary struct {
	Status      ResultStatus
	Contract    statecontract.ContractSummary
	Previous    statematerials.Summary
	Proposed    statecomposition.Summary
	Obligations [8]Obligation
}

func cloneCompositionSummary(s statecomposition.Summary) statecomposition.Summary {
	s.Target.Links = copySlice(s.Target.Links)
	s.Coverage = copySlice(s.Coverage)
	return s
}
func (model *BindingModel) Summary() Summary {
	if model == nil || !model.valid {
		return Summary{}
	}
	return Summary{Status: ConditionalProjectionBound, Contract: model.contractSummary, Previous: model.previousSummary, Proposed: cloneCompositionSummary(model.proposedSummary), Obligations: [8]Obligation{ExternalAnchorAndPlatformProvenance, PreviousSignaturesAndUnselectedHistory, UnverifiedHistoryAndSetDigests, DomainAndCurrentTime, FullStateMaterialClosure, FullStateComposition, PhysicalCommit, RuntimeAuthorization}}
}

// Recheck re-evaluates the whole input, including unprojected/unselected facts.
// A newly valid conditional proof is not the identity of this model's input.
func Recheck(ctx context.Context, model *BindingModel, in Input) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if model == nil || !model.valid {
		return ErrClearedModel
	}
	fresh, e := Evaluate(ctx, in)
	defer fresh.Clear()
	if e != nil {
		return e
	}
	same := reflect.DeepEqual(model.input, fresh.input)
	if e := ctx.Err(); e != nil {
		return e
	}
	if !same {
		return ErrBindingMismatch
	}
	return nil
}
