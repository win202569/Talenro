package nodeissuerbinding

import (
	"bytes"
	"context"
	"sort"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

func ownTable(ctx context.Context, in statematerials.Input) (out statematerials.Input, err error) {
	out = in
	out.Materials = nil
	out.Expected = nil
	done := false
	defer func() {
		if !done {
			clearTable(&out)
		}
	}()
	if ctx == nil {
		return out, ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return out, e
	}
	out.Materials = copySlice(in.Materials)
	for i := range out.Materials {
		out.Materials[i].Raw = nil
	}
	out.Expected = copySlice(in.Expected)
	for i := range out.Expected {
		out.Expected[i].Root.KeyIDs = nil
		out.Expected[i].Metadata.KeyIDs = nil
		out.Expected[i].Metadata.RevokedIDs = nil
		out.Expected[i].Desired.SlotIDs = nil
	}
	for i, m := range in.Materials {
		if e := ctx.Err(); e != nil {
			return out, e
		}
		out.Materials[i].Raw = copyRaw(m.Raw)
	}
	for i, e := range in.Expected {
		if ce := ctx.Err(); ce != nil {
			return out, ce
		}
		out.Expected[i].Root.KeyIDs = copySlice(e.Root.KeyIDs)
		out.Expected[i].Metadata.KeyIDs = copySlice(e.Metadata.KeyIDs)
		out.Expected[i].Metadata.RevokedIDs = copySlice(e.Metadata.RevokedIDs)
		out.Expected[i].Desired.SlotIDs = copySlice(e.Desired.SlotIDs)
	}
	sort.Slice(out.Materials, func(i, j int) bool {
		a, b := out.Materials[i].Ref, out.Materials[j].Ref
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return bytes.Compare(a.Payload[:], b.Payload[:]) < 0
	})
	sort.Slice(out.Expected, func(i, j int) bool {
		a, b := out.Expected[i].Ref, out.Expected[j].Ref
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return bytes.Compare(a.Payload[:], b.Payload[:]) < 0
	})
	if e := ctx.Err(); e != nil {
		return out, e
	}
	done = true
	return out, nil
}
func clearTable(in *statematerials.Input) {
	if in == nil {
		return
	}
	for i := range in.Materials {
		clear(in.Materials[i].Raw)
	}
	for i := range in.Expected {
		clear(in.Expected[i].Root.KeyIDs)
		clear(in.Expected[i].Metadata.KeyIDs)
		clear(in.Expected[i].Metadata.RevokedIDs)
		clear(in.Expected[i].Desired.SlotIDs)
	}
	clear(in.Materials)
	clear(in.Expected)
	*in = statematerials.Input{}
}
