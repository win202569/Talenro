package nodeissuerbinding

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/statecontract"
)

func ownContent(ctx context.Context, in statecontract.MainStateContentModel) (out statecontract.MainStateContentModel, err error) {
	out = in
	out.Root.KeyIDs = nil
	out.Metadata.KeyIDs = nil
	out.Metadata.RevokedIDs = nil
	out.ServerTrust.Bootstrap.Deauthorized = nil
	out.ServerTrust.Agent.Deauthorized = nil
	out.Desired.SlotIDs = nil
	out.Recovery.Local = nil
	out.Recovery.Supervisor = nil
	out.LKG.Entries = nil
	done := false
	defer func() {
		if !done {
			clearContent(&out)
		}
	}()
	if ctx == nil {
		return out, ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return out, e
	}
	out.Root.KeyIDs = copySlice(in.Root.KeyIDs)
	out.Metadata.KeyIDs = copySlice(in.Metadata.KeyIDs)
	out.Metadata.RevokedIDs = copySlice(in.Metadata.RevokedIDs)
	out.ServerTrust.Bootstrap.Deauthorized = copySlice(in.ServerTrust.Bootstrap.Deauthorized)
	out.ServerTrust.Agent.Deauthorized = copySlice(in.ServerTrust.Agent.Deauthorized)
	out.Desired.SlotIDs = copySlice(in.Desired.SlotIDs)
	out.Recovery.Local = copySlice(in.Recovery.Local)
	out.Recovery.Supervisor = copySlice(in.Recovery.Supervisor)
	out.LKG.Entries = copySlice(in.LKG.Entries)
	if e := ctx.Err(); e != nil {
		return out, e
	}
	done = true
	return out, nil
}
func clearContent(in *statecontract.MainStateContentModel) {
	if in == nil {
		return
	}
	clear(in.Root.KeyIDs)
	clear(in.Metadata.KeyIDs)
	clear(in.Metadata.RevokedIDs)
	clear(in.ServerTrust.Bootstrap.Deauthorized)
	clear(in.ServerTrust.Agent.Deauthorized)
	clear(in.Desired.SlotIDs)
	clear(in.Recovery.Local)
	clear(in.Recovery.Supervisor)
	clear(in.LKG.Entries)
	*in = statecontract.MainStateContentModel{}
}
