package statecomposition

import (
	"bytes"
	"context"
	"sort"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

func checkBudget(in Input) error {
	if len(in.Materials.Materials) > 128 || len(in.Materials.Expected) > 128 || len(in.Target.Links) > 32 || len(in.Anchor.Raw) > 262144 {
		return ErrLimitExceeded
	}
	total, rotations := 0, 0
	for _, m := range in.Materials.Materials {
		limit := 262144
		if m.Ref.Kind == statematerials.RootSet {
			limit = 65536
		}
		if m.Ref.Kind == statematerials.RootRotation {
			rotations++
		}
		if len(m.Raw) > limit || len(m.Raw) > 16*1024*1024-total {
			return ErrLimitExceeded
		}
		total += len(m.Raw)
	}
	if rotations > 32 {
		return ErrLimitExceeded
	}
	rotations = 0
	for _, e := range in.Materials.Expected {
		if e.Ref.Kind == statematerials.RootRotation {
			rotations++
		}
		if len(e.Root.KeyIDs) > 5 || len(e.Metadata.KeyIDs) > 64 || len(e.Metadata.RevokedIDs) > 4096 || len(e.Desired.SlotIDs) > 8 {
			return ErrLimitExceeded
		}
	}
	if rotations > 32 {
		return ErrLimitExceeded
	}
	if len(in.Anchor.Raw) == 0 {
		return ErrInvalidInput
	}
	return nil
}

func ownInput(ctx context.Context, in Input) (Input, error) {
	if ctx == nil {
		return Input{}, ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return Input{}, e
	}
	if e := checkBudget(in); e != nil {
		return Input{}, e
	}
	out := Input{}
	done := false
	defer func() {
		if !done {
			clearInput(&out)
		}
	}()
	if e := ctx.Err(); e != nil {
		return Input{}, e
	}
	out.Anchor = in.Anchor
	out.Anchor.Raw = bytes.Clone(in.Anchor.Raw)
	out.Materials.Scope = in.Materials.Scope
	out.Target = in.Target
	out.Target.Links = nil
	if len(in.Target.Links) > 0 {
		out.Target.Links = append([]RotationLink(nil), in.Target.Links...)
	}
	if len(in.Materials.Materials) > 0 {
		out.Materials.Materials = make([]statematerials.Material, len(in.Materials.Materials))
	}
	for i, m := range in.Materials.Materials {
		if e := ctx.Err(); e != nil {
			return Input{}, e
		}
		out.Materials.Materials[i] = statematerials.Material{Ref: m.Ref, Raw: bytes.Clone(m.Raw)}
	}
	if len(in.Materials.Expected) > 0 {
		out.Materials.Expected = make([]statematerials.Expectation, len(in.Materials.Expected))
	}
	for i, e := range in.Materials.Expected {
		if err := ctx.Err(); err != nil {
			return Input{}, err
		}
		dst := &out.Materials.Expected[i]
		*dst = e
		dst.Root.KeyIDs = nil
		dst.Metadata.KeyIDs = nil
		dst.Metadata.RevokedIDs = nil
		dst.Desired.SlotIDs = nil
		dst.Root.KeyIDs = append(dst.Root.KeyIDs, e.Root.KeyIDs...)
		dst.Metadata.KeyIDs = append(dst.Metadata.KeyIDs, e.Metadata.KeyIDs...)
		dst.Metadata.RevokedIDs = append(dst.Metadata.RevokedIDs, e.Metadata.RevokedIDs...)
		dst.Desired.SlotIDs = append(dst.Desired.SlotIDs, e.Desired.SlotIDs...)
	}
	sort.Slice(out.Materials.Materials, func(i, j int) bool { return refLess(out.Materials.Materials[i].Ref, out.Materials.Materials[j].Ref) })
	sort.Slice(out.Materials.Expected, func(i, j int) bool { return refLess(out.Materials.Expected[i].Ref, out.Materials.Expected[j].Ref) })
	if e := ctx.Err(); e != nil {
		return Input{}, e
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
func clearInput(in *Input) {
	if in == nil {
		return
	}
	clear(in.Anchor.Raw)
	for i := range in.Materials.Materials {
		clear(in.Materials.Materials[i].Raw)
	}
	for i := range in.Materials.Expected {
		e := &in.Materials.Expected[i]
		clear(e.Root.KeyIDs)
		clear(e.Metadata.KeyIDs)
		clear(e.Metadata.RevokedIDs)
		clear(e.Desired.SlotIDs)
	}
	clear(in.Materials.Materials)
	clear(in.Materials.Expected)
	clear(in.Target.Links)
	*in = Input{}
}
