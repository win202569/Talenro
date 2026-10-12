package statepolicy

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"talenro.local/platform/internal/nodecontrol/stateverify"
)

func readFact(ctx context.Context, dst *artifactFact, kind stateformat.Kind, raw []byte, id stateverify.ArtifactIdentity) error {
	if e := contextError(ctx); e != nil {
		return e
	}
	a, e := stateformat.ParseEnvelope(kind, raw)
	if e != nil {
		return fail(InvalidInput, Malformed)
	}
	defer a.Clear()
	if a.PayloadDigest() != id.PayloadDigest || a.EnvelopeDigest() != id.EnvelopeDigest {
		return fail(Rejected, SignatureOrBinding)
	}
	dst.bindings = a.Bindings()
	dst.id = id
	switch kind {
	case stateformat.InitialRoot, stateformat.RootRotation:
		dst.root, e = a.RootMaterial()
	case stateformat.Metadata:
		dst.metadata, e = a.MetadataMaterial()
	}
	if e != nil {
		return fail(InvalidInput, Malformed)
	}
	return contextError(ctx)
}
func loadFacts(ctx context.Context, b stateverify.Batch) (result *verifiedFacts, err error) {
	if e := contextError(ctx); e != nil {
		return nil, e
	}
	if len(b.Rotations) > 32 {
		return nil, fail(AssuranceMissing, UnsupportedRange)
	}
	evidence, e := stateverify.Verify(ctx, b)
	if e != nil {
		switch e {
		case stateverify.ErrCanceled:
			return nil, fail(Canceled, Cancellation)
		case stateverify.ErrInvalidInput:
			return nil, fail(InvalidInput, Malformed)
		default:
			return nil, fail(Rejected, SignatureOrBinding)
		}
	}
	defer evidence.Clear()
	f := &verifiedFacts{}
	done := false
	defer func() {
		if !done {
			f.Clear()
		}
	}()
	f.rotations = make([]artifactFact, len(b.Rotations))
	if e = readFact(ctx, &f.anchor, stateformat.InitialRoot, b.AnchorEnvelope, evidence.Anchor()); e != nil {
		return nil, e
	}
	ids := evidence.Rotations()
	defer clear(ids)
	for i := range b.Rotations {
		if e = readFact(ctx, &f.rotations[i], stateformat.RootRotation, b.Rotations[i], ids[i]); e != nil {
			return nil, e
		}
	}
	if e = readFact(ctx, &f.metadata, stateformat.Metadata, b.MetadataEnvelope, evidence.Metadata()); e != nil {
		return nil, e
	}
	if e = readFact(ctx, &f.desired, stateformat.Desired, b.DesiredEnvelope, evidence.Desired()); e != nil {
		return nil, e
	}
	c := evidence.Coordinates()
	d, m := f.desired.bindings, f.metadata.bindings
	if d.NodeID != c.ExpectedNodeID || d.Epoch != c.ExpectedEpoch || d.Generation != c.DesiredGeneration || d.Sequence != c.DesiredSequence || d.SigningKeyID != c.DesiredKeyID || m.MetadataVersion != c.MetadataVersion || m.RootVersion != c.FinalRootVersion {
		return nil, fail(Rejected, SignatureOrBinding)
	}
	if e = contextError(ctx); e != nil {
		return nil, e
	}
	done = true
	return f, nil
}
