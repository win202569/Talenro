package nodeleafbinding

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
	"talenro.local/platform/internal/nodecontrol/statepolicybinding"
)

func ownPolicy(ctx context.Context, in statepolicy.ModelInput) (out statepolicy.ModelInput, err error) {
	out = in
	out.Batch.AnchorEnvelope = nil
	out.Batch.ActiveRootEnvelope = nil
	out.Batch.MetadataEnvelope = nil
	out.Batch.DesiredEnvelope = nil
	out.Batch.Rotations = nil
	out.Old = nil
	out.History = nil
	done := false
	defer func() {
		if !done {
			clearPolicy(&out)
		}
	}()
	if ctx == nil {
		return out, ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return out, e
	}
	out.Time.Now = normalizeTime(in.Time.Now)
	out.Time.Floor = normalizeTime(in.Time.Floor)
	out.Batch.AnchorEnvelope = copyRaw(in.Batch.AnchorEnvelope)
	out.Batch.ActiveRootEnvelope = copyRaw(in.Batch.ActiveRootEnvelope)
	out.Batch.MetadataEnvelope = copyRaw(in.Batch.MetadataEnvelope)
	out.Batch.DesiredEnvelope = copyRaw(in.Batch.DesiredEnvelope)
	out.Batch.Rotations = copySlice(in.Batch.Rotations)
	for i := range out.Batch.Rotations {
		out.Batch.Rotations[i] = nil
	}
	for i, raw := range in.Batch.Rotations {
		if e := ctx.Err(); e != nil {
			return out, e
		}
		out.Batch.Rotations[i] = copyRaw(raw)
	}
	if in.Old != nil {
		out.Old = new(statepolicy.StateModel)
		*out.Old = *in.Old
		out.Old.Roots = nil
		out.Old.Keys = nil
		out.Old.RevokedIDs = nil
		out.Old.TimeFloor = normalizeTime(in.Old.TimeFloor)
		out.Old.Roots = copySlice(in.Old.Roots)
		for i := range out.Old.Roots {
			out.Old.Roots[i].Root.Keys = nil
		}
		for i, r := range in.Old.Roots {
			if e := ctx.Err(); e != nil {
				return out, e
			}
			out.Old.Roots[i].Root.Keys = copySlice(r.Root.Keys)
		}
		out.Old.Keys = copySlice(in.Old.Keys)
		for i := range out.Old.Keys {
			if e := ctx.Err(); e != nil {
				return out, e
			}
			k := &out.Old.Keys[i]
			k.Material.NotBefore = normalizeTime(k.Material.NotBefore)
			k.Material.NotAfter = normalizeTime(k.Material.NotAfter)
			k.Material.RevokedAt = normalizeTime(k.Material.RevokedAt)
			k.FirstRevokedAt = normalizeTime(k.FirstRevokedAt)
			k.LastSnapshotUntil = normalizeTime(k.LastSnapshotUntil)
			k.LastCompositeDeadline = normalizeTime(k.LastCompositeDeadline)
		}
		out.Old.RevokedIDs = copySlice(in.Old.RevokedIDs)
	}
	out.History = copySlice(in.History)
	for i := range out.History {
		if e := ctx.Err(); e != nil {
			return out, e
		}
		out.History[i].LastSnapshotUntil = normalizeTime(out.History[i].LastSnapshotUntil)
		out.History[i].LastCompositeDeadline = normalizeTime(out.History[i].LastCompositeDeadline)
	}
	if e := ctx.Err(); e != nil {
		return out, e
	}
	done = true
	return out, nil
}
func clearPolicy(in *statepolicy.ModelInput) {
	if in == nil {
		return
	}
	clear(in.Batch.AnchorEnvelope)
	clear(in.Batch.ActiveRootEnvelope)
	clear(in.Batch.MetadataEnvelope)
	clear(in.Batch.DesiredEnvelope)
	for i := range in.Batch.Rotations {
		clear(in.Batch.Rotations[i])
	}
	clear(in.Batch.Rotations)
	in.Old.Clear()
	clear(in.History)
	*in = statepolicy.ModelInput{}
}
func ownPolicyBinding(ctx context.Context, in statepolicybinding.Input) (out statepolicybinding.Input, err error) {
	done := false
	defer func() {
		if !done {
			clearPolicyBinding(&out)
		}
	}()
	if ctx == nil {
		return out, ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return out, e
	}
	out.Binding, err = ownBinding(ctx, in.Binding)
	if err != nil {
		return out, err
	}
	out.Policy, err = ownPolicy(ctx, in.Policy)
	if err != nil {
		return out, err
	}
	if e := ctx.Err(); e != nil {
		return out, e
	}
	done = true
	return out, nil
}
func clearPolicyBinding(in *statepolicybinding.Input) {
	if in == nil {
		return
	}
	clearBinding(&in.Binding)
	clearPolicy(&in.Policy)
	*in = statepolicybinding.Input{}
}
