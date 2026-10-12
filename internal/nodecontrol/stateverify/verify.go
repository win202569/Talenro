package stateverify

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/stateformat"
)

// Verify returns conditional signature evidence only. It never chooses its own
// anchor, consults a clock, merges snapshots, retries a shorter chain, or installs
// any state. Inputs must not be modified concurrently with this call.
func Verify(ctx context.Context, batch Batch) (*SignatureEvidence, error) {
	if e := canceled(ctx); e != nil {
		return nil, e
	}
	if batch.ExpectedEpoch == 0 || batch.ExpectedEpoch > 9223372036854775807 || len(batch.Rotations) > 32 {
		return nil, ErrInvalidInput
	}
	node, e := uuid.Parse(batch.ExpectedNodeID)
	if e != nil || len(batch.ExpectedNodeID) != 36 || node == uuid.Nil || node.Variant() != uuid.RFC4122 || node.String() != batch.ExpectedNodeID {
		return nil, ErrInvalidInput
	}
	for _, raw := range [][]byte{batch.AnchorEnvelope, batch.ActiveRootEnvelope, batch.MetadataEnvelope, batch.DesiredEnvelope} {
		if len(raw) == 0 || len(raw) > 262144 {
			return nil, ErrInvalidInput
		}
	}
	for _, raw := range batch.Rotations {
		if len(raw) == 0 || len(raw) > 262144 {
			return nil, ErrInvalidInput
		}
	}
	o := &verificationOwner{}
	defer o.Clear()
	o.evidence = &SignatureEvidence{}
	anchor, e := o.parse(ctx, 0, stateformat.InitialRoot, batch.AnchorEnvelope)
	if e != nil {
		return nil, e
	}
	ab := anchor.Bindings()
	o.roots[0], e = anchor.RootMaterial()
	if e != nil {
		return nil, ErrInvalidInput
	}
	if ab.Epoch != batch.ExpectedEpoch || ab.RootVersion != 1 || o.roots[0].Threshold != 1 || len(o.roots[0].Keys) != 1 {
		return nil, ErrBindingMismatch
	}
	o.evidence.anchor = identity(anchor)
	o.evidence.rotations = make([]ArtifactIdentity, len(batch.Rotations))
	current := &o.roots[0]
	sequence := ab.Sequence
	last := batch.AnchorEnvelope
	for i, raw := range batch.Rotations {
		a, e := o.parse(ctx, i+1, stateformat.RootRotation, raw)
		if e != nil {
			return nil, e
		}
		b := a.Bindings()
		if b.Epoch != batch.ExpectedEpoch || b.PreviousRootVersion != current.Version || current.Version >= 9223372036854775807 || b.RootVersion != current.Version+1 || b.Sequence <= sequence {
			return nil, ErrBindingMismatch
		}
		o.roots[i+1], e = a.RootMaterial()
		if e != nil {
			return nil, ErrInvalidInput
		}
		o.shares[i], e = a.SignatureMaterial()
		if e != nil {
			return nil, ErrInvalidInput
		}
		o.transcripts[i], e = a.Transcript()
		if e != nil {
			return nil, ErrInvalidInput
		}
		if e = verifyThreshold(ctx, *current, o.shares[i].Current, o.transcripts[i]); e != nil {
			return nil, e
		}
		if e = verifyThreshold(ctx, o.roots[i+1], o.shares[i].New, o.transcripts[i]); e != nil {
			return nil, e
		}
		current = &o.roots[i+1]
		sequence = b.Sequence
		last = raw
		o.evidence.rotations[i] = identity(a)
		if e = canceled(ctx); e != nil {
			return nil, e
		}
	}
	if e = canceled(ctx); e != nil {
		return nil, e
	}
	// The exact matched bytes have already been parsed with their fixed Kind.
	if !bytes.Equal(batch.ActiveRootEnvelope, last) {
		// A different artifact must still be well-formed before it can be
		// classified as a relationship mismatch. Never use it as a fallback.
		kind := stateformat.InitialRoot
		if len(batch.Rotations) > 0 {
			kind = stateformat.RootRotation
		}
		if _, e := o.parse(ctx, 35, kind, batch.ActiveRootEnvelope); e != nil {
			return nil, e
		}
		return nil, ErrBindingMismatch
	}
	metadata, e := o.parse(ctx, 33, stateformat.Metadata, batch.MetadataEnvelope)
	if e != nil {
		return nil, e
	}
	mb := metadata.Bindings()
	if mb.Epoch != batch.ExpectedEpoch || mb.RootVersion != current.Version {
		return nil, ErrBindingMismatch
	}
	o.metadata, e = metadata.MetadataMaterial()
	if e != nil {
		return nil, ErrInvalidInput
	}
	o.shares[32], e = metadata.SignatureMaterial()
	if e != nil {
		return nil, ErrInvalidInput
	}
	o.transcripts[32], e = metadata.Transcript()
	if e != nil {
		return nil, ErrInvalidInput
	}
	if e = verifyThreshold(ctx, *current, o.shares[32].Metadata, o.transcripts[32]); e != nil {
		return nil, e
	}
	desired, e := o.parse(ctx, 34, stateformat.Desired, batch.DesiredEnvelope)
	if e != nil {
		return nil, e
	}
	db := desired.Bindings()
	if db.Epoch != batch.ExpectedEpoch || db.NodeID != node {
		return nil, ErrBindingMismatch
	}
	for _, id := range o.metadata.RevokedIDs {
		if id == db.SigningKeyID {
			return nil, ErrBindingMismatch
		}
	}
	var key *stateformat.MetadataKeyMaterial
	for i := range o.metadata.Keys {
		if o.metadata.Keys[i].KeyID == db.SigningKeyID {
			key = &o.metadata.Keys[i]
			break
		}
	}
	if key == nil || (key.Status != "active" && key.Status != "retiring") {
		return nil, ErrBindingMismatch
	}
	if db.IssuedAt.Before(key.NotBefore) || !db.IssuedAt.Before(key.NotAfter) || db.ValidUntil.After(key.NotAfter) || db.ValidUntil.After(mb.ValidUntil) {
		return nil, ErrBindingMismatch
	}
	o.transcripts[33], e = desired.Transcript()
	if e != nil {
		return nil, ErrInvalidInput
	}
	if e = canceled(ctx); e != nil {
		return nil, e
	}
	valid := ed25519.Verify(key.PublicKey[:], o.transcripts[33], db.Signature[:])
	if e = canceled(ctx); e != nil {
		return nil, e
	}
	if !valid {
		return nil, ErrSignatureMismatch
	}
	o.evidence.coordinates = EvidenceCoordinates{ExpectedNodeID: node, ExpectedEpoch: batch.ExpectedEpoch, FinalRootVersion: current.Version, MetadataVersion: mb.MetadataVersion, DesiredGeneration: db.Generation, DesiredSequence: db.Sequence, DesiredKeyID: db.SigningKeyID}
	o.evidence.metadata = identity(metadata)
	o.evidence.desired = identity(desired)
	if e = canceled(ctx); e != nil {
		return nil, e
	}
	result := o.evidence
	o.evidence = nil
	return result, nil
}
func identity(a *stateformat.ParsedArtifact) ArtifactIdentity {
	return ArtifactIdentity{PayloadDigest: a.PayloadDigest(), EnvelopeDigest: a.EnvelopeDigest()}
}
func (o *verificationOwner) parse(ctx context.Context, index int, kind stateformat.Kind, raw []byte) (*stateformat.ParsedArtifact, error) {
	if e := canceled(ctx); e != nil {
		return nil, e
	}
	a, e := stateformat.ParseEnvelope(kind, raw)
	o.artifacts[index] = a
	if ce := canceled(ctx); ce != nil {
		return nil, ce
	}
	if e != nil {
		return nil, ErrInvalidInput
	}
	return a, nil
}
