package statetimebinding

import (
	"context"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/statepolicybinding"
)

// Model owns complete input and summary. Do not copy or use concurrently with Clear.
type Model struct {
	valid   bool
	input   Input
	summary Summary
}

func clearPolicySummary(in *statepolicybinding.Summary) {
	if in == nil {
		return
	}
	clear(in.Binding.Proposed.Target.Links)
	clear(in.Binding.Proposed.Coverage)
	*in = statepolicybinding.Summary{}
}
func clearSummary(in *Summary) {
	if in == nil {
		return
	}
	clearPolicySummary(&in.Policy)
	*in = Summary{}
}
func clonePolicySummary(in statepolicybinding.Summary) (out statepolicybinding.Summary) {
	out = in
	out.Binding.Proposed.Target.Links = nil
	out.Binding.Proposed.Coverage = nil
	done := false
	defer func() {
		if !done {
			clearPolicySummary(&out)
		}
	}()
	out.Binding.Proposed.Target.Links = copySlice(in.Binding.Proposed.Target.Links)
	out.Binding.Proposed.Coverage = copySlice(in.Binding.Proposed.Coverage)
	done = true
	return out
}
func cloneSummary(in Summary) (out Summary) {
	out = in
	out.Policy = statepolicybinding.Summary{}
	done := false
	defer func() {
		if !done {
			clearSummary(&out)
		}
	}()
	out.Policy = clonePolicySummary(in.Policy)
	done = true
	return out
}
func (model *Model) Summary() Summary {
	if model == nil || !model.valid {
		return Summary{}
	}
	return cloneSummary(model.summary)
}
func (model *Model) Clear() {
	if model == nil {
		return
	}
	clearInput(&model.input)
	clearSummary(&model.summary)
	*model = Model{}
}

// Recheck compares all normalized inputs after complete fresh evaluation.
// Forward time and changed observations require new Evaluate, never an old owner.
func Recheck(ctx context.Context, model *Model, in Input) error {
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
	if e = ctx.Err(); e != nil {
		return e
	}
	if !same {
		return ErrBindingMismatch
	}
	return nil
}
