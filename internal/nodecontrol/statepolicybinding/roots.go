package statepolicybinding

import (
	"context"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
)

type rootFact struct {
	projection RootProjection
	material   stateformat.RootMaterial
}
type rootPath struct{ items []rootFact }

func (path *rootPath) Clear() {
	if path == nil {
		return
	}
	for i := range path.items {
		path.items[i].material.Clear()
	}
	clear(path.items)
	*path = rootPath{}
}
func loadRootFact(ctx context.Context, table statematerials.Input, ref statematerials.MaterialRef, source SourceIdentity) (out rootFact, err error) {
	done := false
	defer func() {
		if !done {
			out.material.Clear()
			out = rootFact{}
		}
	}()
	if ref.Kind != statematerials.RootSet {
		return out, ErrMissingMaterial
	}
	m, e := findMaterial(ctx, table, ref)
	if e != nil {
		return out, e
	}
	parsed, e := stateformat.ParseRootSet(m.Raw)
	defer parsed.Clear()
	if e != nil {
		return out, ErrInvalidMaterial
	}
	out.material, e = parsed.RootMaterial()
	if e != nil {
		return out, ErrInvalidMaterial
	}
	b := parsed.Bindings()
	if parsed.PayloadDigest() != ref.Payload || b.Epoch != table.Scope.Epoch {
		return out, ErrBindingMismatch
	}
	out.projection = RootProjection{RootSet: ref, Version: b.Version, Sequence: b.Sequence, Source: source}
	if err = ctx.Err(); err != nil {
		return out, err
	}
	done = true
	return out, nil
}
func buildRootPath(ctx context.Context, in Input) (result *rootPath, err error) {
	if ctx == nil {
		return nil, ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	t := in.Binding.Proposed.Target
	if len(t.Links) > 32 {
		return nil, ErrLimitExceeded
	}
	path := &rootPath{items: make([]rootFact, len(t.Links)+1)}
	defer func() {
		if result == nil {
			path.Clear()
		}
	}()
	a := in.Binding.Proposed.Anchor
	path.items[0], err = loadRootFact(ctx, in.Binding.Proposed.Materials, t.AnchorRoot, SourceIdentity{Kind: InitialAnchor, Payload: a.Payload, Envelope: a.Envelope})
	if err != nil {
		return nil, err
	}
	for i, l := range t.Links {
		path.items[i+1], err = loadRootFact(ctx, in.Binding.Proposed.Materials, l.NewRoot, SourceIdentity{Kind: RootRotationProof, Payload: l.Rotation.Payload, Envelope: l.Rotation.Envelope})
		if err != nil {
			return nil, err
		}
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	return path, nil
}
func bindRoots(ctx context.Context, in Input, path *rootPath, state *statepolicy.StateModel, proposed bool) (RootProjection, error) {
	if ctx == nil {
		return RootProjection{}, ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return RootProjection{}, e
	}
	if path == nil || state == nil || len(state.Roots) == 0 || len(state.Roots) > len(path.items) || (proposed && len(state.Roots) != len(path.items)) {
		return RootProjection{}, ErrBindingMismatch
	}
	for i, r := range state.Roots {
		if e := ctx.Err(); e != nil {
			return RootProjection{}, e
		}
		f := path.items[i]
		if r.Mark.Version != f.projection.Version || r.Mark.Sequence != f.projection.Sequence || r.Mark.Payload != f.projection.Source.Payload || !reflect.DeepEqual(r.Root, f.material) {
			return RootProjection{}, ErrBindingMismatch
		}
	}
	last := path.items[len(state.Roots)-1].projection
	main := in.Binding.Main.Previous
	if proposed {
		main = in.Binding.Main.Proposed
	}
	ref, e := projectRef(main.Root.Mark.Material)
	if e != nil {
		return RootProjection{}, e
	}
	if ref != last.RootSet || main.Root.Mark.Version != last.Version || main.Root.Mark.Sequence != last.Sequence {
		return RootProjection{}, ErrBindingMismatch
	}
	if e := ctx.Err(); e != nil {
		return RootProjection{}, e
	}
	return last, nil
}
func bindRootTransition(ctx context.Context, in Input, path *rootPath) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if path == nil || in.Policy.Old == nil || len(in.Policy.Old.Roots) == 0 || len(in.Policy.Old.Roots) > len(path.items) {
		return ErrBindingMismatch
	}
	start := len(in.Policy.Old.Roots) - 1
	links := in.Binding.Proposed.Target.Links
	if len(links)+1 != len(path.items) || len(in.Binding.Main.RootTransition) != len(links)-start {
		return ErrBindingMismatch
	}
	for j, r := range in.Binding.Main.RootTransition {
		i := start + j
		l := links[i]
		a := path.items[i].projection
		b := path.items[i+1].projection
		ref, e := projectRef(r.Material)
		if e != nil {
			return e
		}
		ex, e := findExpected(ctx, in.Binding.Proposed.Materials, l.Rotation)
		if e != nil {
			return e
		}
		if ref != l.Rotation || r.OldRootPayload != a.RootSet.Payload || r.NewRootPayload != b.RootSet.Payload || r.PreviousVersion != a.Version || r.NewVersion != b.Version || r.Sequence != b.Sequence || ex.Rotation.PreviousVersion != r.PreviousVersion || ex.Rotation.NewVersion != r.NewVersion || ex.Rotation.OldRootPayload != r.OldRootPayload || ex.Rotation.NewRootPayload != r.NewRootPayload || ex.Sequence != r.Sequence {
			return ErrBindingMismatch
		}
	}
	return ctx.Err()
}
