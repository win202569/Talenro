package statepolicybinding

import (
	"bytes"
	"context"
	"sort"
	"talenro.local/platform/internal/nodecontrol/statebinding"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
	"time"
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

func ownBinding(ctx context.Context, in statebinding.Input) (out statebinding.Input, err error) {
	if ctx == nil {
		return statebinding.Input{}, ErrInvalidInput
	}
	if err = ctx.Err(); err != nil {
		return statebinding.Input{}, err
	}
	if _, err = bindingBudget(in); err != nil {
		return statebinding.Input{}, err
	}
	done := false
	defer func() {
		if !done {
			clearBinding(&out)
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
func clearBinding(in *statebinding.Input) {
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
	*in = statebinding.Input{}
}

func normalizeTime(v time.Time) time.Time {
	if v.IsZero() {
		return time.Time{}
	}
	return v.Round(0).UTC()
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
	if out.Binding, err = ownBinding(ctx, in.Binding); err != nil {
		return out, err
	}
	if out.Policy, err = ownPolicy(ctx, in.Policy); err != nil {
		return out, err
	}
	if err = ctx.Err(); err != nil {
		return out, err
	}
	done = true
	return out, nil
}
func ownPolicy(ctx context.Context, in statepolicy.ModelInput) (out statepolicy.ModelInput, err error) {
	out = in
	out.Batch.AnchorEnvelope = nil
	out.Batch.Rotations = nil
	out.Batch.ActiveRootEnvelope = nil
	out.Batch.MetadataEnvelope = nil
	out.Batch.DesiredEnvelope = nil
	out.Old = nil
	out.History = nil
	done := false
	defer func() {
		if !done {
			clearPolicy(&out)
		}
	}()
	if err = ctx.Err(); err != nil {
		return out, err
	}
	out.Batch.AnchorEnvelope = copyRaw(in.Batch.AnchorEnvelope)
	if len(in.Batch.Rotations) > 0 {
		out.Batch.Rotations = make([][]byte, len(in.Batch.Rotations))
	}
	for i := range in.Batch.Rotations {
		if err = ctx.Err(); err != nil {
			return out, err
		}
		out.Batch.Rotations[i] = copyRaw(in.Batch.Rotations[i])
	}
	out.Batch.ActiveRootEnvelope = copyRaw(in.Batch.ActiveRootEnvelope)
	out.Batch.MetadataEnvelope = copyRaw(in.Batch.MetadataEnvelope)
	out.Batch.DesiredEnvelope = copyRaw(in.Batch.DesiredEnvelope)
	out.Time.Now = normalizeTime(in.Time.Now)
	out.Time.Floor = normalizeTime(in.Time.Floor)
	if in.Old != nil {
		out.Old = &statepolicy.StateModel{}
		*out.Old = *in.Old
		out.Old.Roots = nil
		out.Old.Keys = nil
		out.Old.RevokedIDs = nil
		out.Old.TimeFloor = normalizeTime(in.Old.TimeFloor)
		if len(in.Old.Roots) > 0 {
			out.Old.Roots = make([]statepolicy.RootRecord, len(in.Old.Roots))
		}
		for i := range in.Old.Roots {
			if err = ctx.Err(); err != nil {
				return out, err
			}
			out.Old.Roots[i].Mark = in.Old.Roots[i].Mark
			out.Old.Roots[i].Root.Version = in.Old.Roots[i].Root.Version
			out.Old.Roots[i].Root.Threshold = in.Old.Roots[i].Root.Threshold
			out.Old.Roots[i].Root.Keys = copySlice(in.Old.Roots[i].Root.Keys)
		}
		out.Old.Keys = copySlice(in.Old.Keys)
		for i := range out.Old.Keys {
			if err = ctx.Err(); err != nil {
				return out, err
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
		if err = ctx.Err(); err != nil {
			return out, err
		}
		out.History[i].LastSnapshotUntil = normalizeTime(out.History[i].LastSnapshotUntil)
		out.History[i].LastCompositeDeadline = normalizeTime(out.History[i].LastCompositeDeadline)
	}
	if err = ctx.Err(); err != nil {
		return out, err
	}
	done = true
	return out, nil
}
func clearPolicy(in *statepolicy.ModelInput) {
	if in == nil {
		return
	}
	clear(in.Batch.AnchorEnvelope)
	for i := range in.Batch.Rotations {
		clear(in.Batch.Rotations[i])
	}
	clear(in.Batch.Rotations)
	clear(in.Batch.ActiveRootEnvelope)
	clear(in.Batch.MetadataEnvelope)
	clear(in.Batch.DesiredEnvelope)
	in.Old.Clear()
	clear(in.History)
	*in = statepolicy.ModelInput{}
}
func clearInput(in *Input) {
	if in == nil {
		return
	}
	clearBinding(&in.Binding)
	clearPolicy(&in.Policy)
	*in = Input{}
}
