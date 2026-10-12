package statepolicybinding

import (
	"context"
	"reflect"
)

// Model owns a supplied snapshot and a conditional result. Do not copy it or
// use it concurrently with Clear. It confers no commit or runtime authority.
type Model struct {
	valid   bool
	input   Input
	summary Summary
}

func clearSummary(s *Summary) {
	if s == nil {
		return
	}
	clear(s.Binding.Proposed.Target.Links)
	clear(s.Binding.Proposed.Coverage)
	*s = Summary{}
}
func (model *Model) Clear() {
	if model == nil {
		return
	}
	clearInput(&model.input)
	clearSummary(&model.summary)
	*model = Model{}
}
func cloneSummary(in Summary) (out Summary) {
	out = in
	out.Binding.Proposed.Target.Links = nil
	out.Binding.Proposed.Coverage = nil
	done := false
	defer func() {
		if !done {
			clearSummary(&out)
		}
	}()
	out.Binding.Proposed.Target.Links = copySlice(in.Binding.Proposed.Target.Links)
	out.Binding.Proposed.Coverage = copySlice(in.Binding.Proposed.Coverage)
	done = true
	return out
}
func (model *Model) Summary() Summary {
	if model == nil || !model.valid {
		return Summary{}
	}
	return cloneSummary(model.summary)
}

// Recheck compares the complete normalized snapshot, including opaque and
// unprojected inputs. Even forward time changes require a new Evaluate.
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
