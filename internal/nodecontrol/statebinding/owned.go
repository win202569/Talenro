package statebinding

import (
	"bytes"
	"context"
	"sort"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

func copySlice[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	return append([]T(nil), s...)
}
func copyRaw(s []byte) []byte {
	if len(s) == 0 {
		return nil
	}
	return bytes.Clone(s)
}

func ownInput(ctx context.Context, in Input) (out Input, err error) {
	if ctx == nil {
		return Input{}, ErrInvalidInput
	}
	if err = ctx.Err(); err != nil {
		return Input{}, err
	}
	if err = checkBudget(in); err != nil {
		return Input{}, err
	}
	done := false
	defer func() {
		if !done {
			clearInput(&out)
		}
	}()
	if err = ctx.Err(); err != nil {
		return out, err
	}
	if out.Main.Previous, err = copyContent(ctx, in.Main.Previous); err != nil {
		return out, err
	}
	if out.Main.Proposed, err = copyContent(ctx, in.Main.Proposed); err != nil {
		return out, err
	}
	out.Main.RootTransition = copySlice(in.Main.RootTransition)
	if out.Previous, err = copyTable(ctx, in.Previous); err != nil {
		return out, err
	}
	if out.Proposed.Materials, err = copyTable(ctx, in.Proposed.Materials); err != nil {
		return out, err
	}
	out.Proposed.Anchor = in.Proposed.Anchor
	out.Proposed.Anchor.Raw = nil
	out.Proposed.Anchor.Raw = copyRaw(in.Proposed.Anchor.Raw)
	out.Proposed.Target = in.Proposed.Target
	out.Proposed.Target.Links = nil
	out.Proposed.Target.Links = copySlice(in.Proposed.Target.Links)
	if err = ctx.Err(); err != nil {
		return out, err
	}
	done = true
	return out, nil
}
func copyContent(ctx context.Context, x statecontract.MainStateContentModel) (out statecontract.MainStateContentModel, err error) {
	out = x
	out.Root.KeyIDs = nil
	out.Metadata.KeyIDs = nil
	out.Metadata.RevokedIDs = nil
	out.ServerTrust.Bootstrap.Deauthorized = nil
	out.ServerTrust.Agent.Deauthorized = nil
	out.Desired.SlotIDs = nil
	out.LKG.Entries = nil
	out.Recovery.Local = nil
	out.Recovery.Supervisor = nil
	done := false
	defer func() {
		if !done {
			clearContent(&out)
		}
	}()
	if err = ctx.Err(); err != nil {
		return statecontract.MainStateContentModel{}, err
	}
	out.Root.KeyIDs = copySlice(x.Root.KeyIDs)
	out.Metadata.KeyIDs = copySlice(x.Metadata.KeyIDs)
	out.Metadata.RevokedIDs = copySlice(x.Metadata.RevokedIDs)
	out.ServerTrust.Bootstrap.Deauthorized = copySlice(x.ServerTrust.Bootstrap.Deauthorized)
	out.ServerTrust.Agent.Deauthorized = copySlice(x.ServerTrust.Agent.Deauthorized)
	out.Desired.SlotIDs = copySlice(x.Desired.SlotIDs)
	out.LKG.Entries = copySlice(x.LKG.Entries)
	out.Recovery.Local = copySlice(x.Recovery.Local)
	out.Recovery.Supervisor = copySlice(x.Recovery.Supervisor)
	if err = ctx.Err(); err != nil {
		return out, err
	}
	done = true
	return out, nil
}
func copyTable(ctx context.Context, in statematerials.Input) (out statematerials.Input, err error) {
	done := false
	defer func() {
		if !done {
			clearTable(&out)
		}
	}()
	out.Scope = in.Scope
	if len(in.Materials) > 0 {
		out.Materials = make([]statematerials.Material, len(in.Materials))
	}
	for i, m := range in.Materials {
		if err = ctx.Err(); err != nil {
			return out, err
		}
		out.Materials[i] = statematerials.Material{Ref: m.Ref, Raw: copyRaw(m.Raw)}
	}
	if len(in.Expected) > 0 {
		out.Expected = make([]statematerials.Expectation, len(in.Expected))
	}
	for i, e := range in.Expected {
		if err = ctx.Err(); err != nil {
			return out, err
		}
		dst := &out.Expected[i]
		*dst = e
		dst.Root.KeyIDs = nil
		dst.Metadata.KeyIDs = nil
		dst.Metadata.RevokedIDs = nil
		dst.Desired.SlotIDs = nil
		dst.Root.KeyIDs = copySlice(e.Root.KeyIDs)
		dst.Metadata.KeyIDs = copySlice(e.Metadata.KeyIDs)
		dst.Metadata.RevokedIDs = copySlice(e.Metadata.RevokedIDs)
		dst.Desired.SlotIDs = copySlice(e.Desired.SlotIDs)
	}
	sort.Slice(out.Materials, func(i, j int) bool { return refLess(out.Materials[i].Ref, out.Materials[j].Ref) })
	sort.Slice(out.Expected, func(i, j int) bool { return refLess(out.Expected[i].Ref, out.Expected[j].Ref) })
	if err = ctx.Err(); err != nil {
		return out, err
	}
	done = true
	return out, nil
}
func refLess(a, b statematerials.MaterialRef) bool {
	if a.Kind != b.Kind {
		return a.Kind < b.Kind
	}
	return bytes.Compare(a.Payload[:], b.Payload[:]) < 0
}
func clearContent(x *statecontract.MainStateContentModel) {
	clear(x.Root.KeyIDs)
	clear(x.Metadata.KeyIDs)
	clear(x.Metadata.RevokedIDs)
	clear(x.ServerTrust.Bootstrap.Deauthorized)
	clear(x.ServerTrust.Agent.Deauthorized)
	clear(x.Desired.SlotIDs)
	clear(x.LKG.Entries)
	clear(x.Recovery.Local)
	clear(x.Recovery.Supervisor)
	*x = statecontract.MainStateContentModel{}
}
func clearTable(in *statematerials.Input) {
	for i := range in.Materials {
		clear(in.Materials[i].Raw)
	}
	for i := range in.Expected {
		e := &in.Expected[i]
		clear(e.Root.KeyIDs)
		clear(e.Metadata.KeyIDs)
		clear(e.Metadata.RevokedIDs)
		clear(e.Desired.SlotIDs)
	}
	clear(in.Materials)
	clear(in.Expected)
	*in = statematerials.Input{}
}
func clearInput(in *Input) {
	if in == nil {
		return
	}
	clearContent(&in.Main.Previous)
	clearContent(&in.Main.Proposed)
	clear(in.Main.RootTransition)
	clearTable(&in.Previous)
	clearTable(&in.Proposed.Materials)
	clear(in.Proposed.Anchor.Raw)
	clear(in.Proposed.Target.Links)
	*in = Input{}
}
