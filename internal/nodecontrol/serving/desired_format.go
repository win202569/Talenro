package serving

import (
	"bytes"
	"context"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/nodecontrol/stateformat"
)

type desiredFormat struct {
	desired, root, metadata *stateformat.ParsedArtifact
	rotations               []*stateformat.ParsedArtifact
}

// Private format staging only. No production reader consumes this result, and
// successful construction provides no signature, deployment or time authority.
func bindDesiredFormat(ctx context.Context, p desiredProjection) (desiredFormat, error) {
	if ctx != nil && ctx.Err() != nil {
		return desiredFormat{}, authority.ErrCanceled
	}
	if ctx == nil {
		return desiredFormat{}, authority.ErrAuthorityUnavailable
	}
	s := p.snapshot
	checked, err := newDesiredProjection(s.NodeID, s.Desired.Authority.Epoch, s)
	defer discardDesiredProjection(&checked)
	if ctx.Err() != nil {
		return desiredFormat{}, authority.ErrCanceled
	}
	if err != nil || s.ChainBaseRootVersion != 1 {
		return desiredFormat{}, authority.ErrConflict
	}
	life := desiredFormatLifetime{}
	defer life.cleanup()
	s = checked.snapshot
	life.staged.rotations = make([]*stateformat.ParsedArtifact, len(s.RootRotations))
	for i, root := range s.RootRotations {
		if err := life.bindPublish(ctx, root, stateformat.RootRotation, &life.staged.rotations[i]); err != nil {
			return desiredFormat{}, err
		}
	}
	kind := stateformat.RootRotation
	if s.ActiveRoot.Version == 1 {
		kind = stateformat.InitialRoot
	}
	if err := life.bindPublish(ctx, s.ActiveRoot, kind, &life.staged.root); err != nil {
		return desiredFormat{}, err
	}
	if err := life.bindPublish(ctx, s.ActiveMetadata, stateformat.Metadata, &life.staged.metadata); err != nil {
		return desiredFormat{}, err
	}
	if ctx.Err() != nil {
		return desiredFormat{}, authority.ErrCanceled
	}
	d := s.Desired
	life.staged.desired, err = stateformat.BuildDesiredEnvelope(d.Payload, d.PayloadDigest, d.SigningKeyID, d.Signature)
	if ctx.Err() != nil {
		return desiredFormat{}, authority.ErrCanceled
	}
	if err != nil {
		return desiredFormat{}, authority.ErrConflict
	}
	b := life.staged.desired.Bindings()
	if b.Kind != stateformat.Desired || b.NodeID != d.NodeID || b.Epoch != d.Authority.Epoch || b.Sequence != d.Authority.Sequence || b.Generation != d.Generation || b.InventoryVersion != d.InventoryVersion || b.ResourceEnvelopeVersion != d.ResourceEnvelopeVersion || b.ResourceEnvelopeDigest != d.ResourceEnvelopeDigest || !b.IssuedAt.Equal(d.IssuedAt) || !b.ValidUntil.Equal(d.ValidUntil) || b.SigningKeyID != d.SigningKeyID || b.Signature != d.Signature {
		return desiredFormat{}, formatFailure(ctx)
	}
	return life.finish(ctx)
}
func discardDesiredFormat(f *desiredFormat) {
	if f == nil {
		return
	}
	f.desired.Clear()
	f.root.Clear()
	f.metadata.Clear()
	for _, a := range f.rotations {
		a.Clear()
	}
	clear(f.rotations)
	*f = desiredFormat{}
}

type desiredFormatLifetime struct{ staged desiredFormat }

func (l *desiredFormatLifetime) cleanup() { discardDesiredFormat(&l.staged) }
func (l *desiredFormatLifetime) finish(ctx context.Context) (desiredFormat, error) {
	if ctx.Err() != nil {
		return desiredFormat{}, authority.ErrCanceled
	}
	out := l.staged
	l.staged = desiredFormat{}
	return out, nil
}
func formatFailure(ctx context.Context) error {
	if ctx.Err() != nil {
		return authority.ErrCanceled
	}
	return authority.ErrConflict
}
func (l *desiredFormatLifetime) bindPublish(ctx context.Context, p desiredPublishProjection, kind stateformat.Kind, destination **stateformat.ParsedArtifact) error {
	if ctx.Err() != nil {
		return authority.ErrCanceled
	}
	a, err := stateformat.ParseEnvelope(kind, p.PublishedEnvelope)
	// The local pointer owns a until the final transfer, including on panic.
	defer func() { a.Clear() }()
	if ctx.Err() != nil {
		return authority.ErrCanceled
	}
	if err != nil {
		return authority.ErrConflict
	}
	payload := a.Payload()
	defer clear(payload)
	if !bytes.Equal(payload, p.Payload) || a.PayloadDigest() != p.PayloadDigest || a.EnvelopeDigest() != p.EnvelopeDigest {
		return formatFailure(ctx)
	}
	b := a.Bindings()
	if b.Kind != kind || b.Epoch != p.Authority.Epoch || b.Sequence != p.Authority.Sequence {
		return formatFailure(ctx)
	}
	switch kind {
	case stateformat.InitialRoot:
		if b.RootVersion != p.Version || p.Version != 1 || p.BaseRootVersion != 0 {
			return formatFailure(ctx)
		}
	case stateformat.RootRotation:
		if b.RootVersion != p.Version || b.PreviousRootVersion != p.BaseRootVersion {
			return formatFailure(ctx)
		}
	case stateformat.Metadata:
		if b.RootVersion != p.BaseRootVersion || b.MetadataVersion != p.Version {
			return formatFailure(ctx)
		}
	default:
		return formatFailure(ctx)
	}
	if ctx.Err() != nil {
		return authority.ErrCanceled
	}
	(*destination).Clear()
	*destination = a
	a = nil
	return nil
}
