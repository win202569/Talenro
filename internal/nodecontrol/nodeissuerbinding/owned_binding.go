package nodeissuerbinding

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/statebinding"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

func ownBinding(ctx context.Context, in statebinding.Input) (out statebinding.Input, err error) {
	out = in
	out.Main.Previous = statecontract.MainStateContentModel{}
	out.Main.Proposed = statecontract.MainStateContentModel{}
	out.Main.RootTransition = nil
	out.Previous = statematerials.Input{}
	out.Proposed.Materials = statematerials.Input{}
	out.Proposed.Target.Links = nil
	out.Proposed.Anchor.Raw = nil
	done := false
	defer func() {
		if !done {
			clearBinding(&out)
		}
	}()
	if ctx == nil {
		return out, ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return out, e
	}
	out.Main.Previous, err = ownContent(ctx, in.Main.Previous)
	if err != nil {
		return out, err
	}
	out.Main.Proposed, err = ownContent(ctx, in.Main.Proposed)
	if err != nil {
		return out, err
	}
	out.Main.RootTransition = copySlice(in.Main.RootTransition)
	out.Previous, err = ownTable(ctx, in.Previous)
	if err != nil {
		return out, err
	}
	out.Proposed.Materials, err = ownTable(ctx, in.Proposed.Materials)
	if err != nil {
		return out, err
	}
	out.Proposed.Target.Links = copySlice(in.Proposed.Target.Links)
	out.Proposed.Anchor.Raw = copyRaw(in.Proposed.Anchor.Raw)
	if e := ctx.Err(); e != nil {
		return out, e
	}
	done = true
	return out, nil
}
func clearBinding(in *statebinding.Input) {
	if in == nil {
		return
	}
	clearContent(&in.Main.Previous)
	clearContent(&in.Main.Proposed)
	clear(in.Main.RootTransition)
	clearTable(&in.Previous)
	clearTable(&in.Proposed.Materials)
	clear(in.Proposed.Target.Links)
	clear(in.Proposed.Anchor.Raw)
	*in = statebinding.Input{}
}
