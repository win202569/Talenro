package statecomposition

import (
	"bytes"
	"context"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/stateverify"
)

type selection struct {
	batch                                                            stateverify.Batch
	finalRoot                                                        statematerials.MaterialRef
	rootVersion, metadataVersion, desiredGeneration, desiredSequence uint64
	desiredKeyID                                                     contracts.Digest
}

func (s *selection) Clear() {
	if s != nil {
		clear(s.batch.Rotations)
		*s = selection{}
	}
}

type materialKey struct {
	kind    statematerials.MaterialKind
	payload contracts.Digest
}

func key(r statematerials.MaterialRef) materialKey { return materialKey{r.Kind, r.Payload} }
func validTargetRef(r statematerials.MaterialRef, kind statematerials.MaterialKind) bool {
	return r.Kind == kind && r.Payload != (contracts.Digest{}) && ((kind == statematerials.RootSet && r.Envelope == (contracts.Digest{})) || (kind != statematerials.RootSet && r.Envelope != (contracts.Digest{})))
}

// in must be owned and must have just passed the complete material evaluation.
// The selected batch borrows raw bytes from in; Clear only erases outer refs.
func selectBatch(ctx context.Context, in Input) (*selection, error) {
	if ctx == nil {
		return nil, ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if len(in.Target.Links) > 32 {
		return nil, ErrLimitExceeded
	}
	if in.Anchor.Payload == (contracts.Digest{}) || in.Anchor.Envelope == (contracts.Digest{}) || len(in.Anchor.Raw) == 0 || !validTargetRef(in.Target.AnchorRoot, statematerials.RootSet) || !validTargetRef(in.Target.Metadata, statematerials.Metadata) || !validTargetRef(in.Target.Desired, statematerials.Desired) {
		return nil, ErrInvalidInput
	}
	for _, l := range in.Target.Links {
		if e := ctx.Err(); e != nil {
			return nil, e
		}
		if !validTargetRef(l.Rotation, statematerials.RootRotation) || !validTargetRef(l.NewRoot, statematerials.RootSet) {
			return nil, ErrInvalidInput
		}
	}
	if in.Anchor.Scope != in.Materials.Scope {
		return nil, ErrBindingMismatch
	}
	materials := map[materialKey]statematerials.Material{}
	expectations := map[materialKey]statematerials.Expectation{}
	for _, m := range in.Materials.Materials {
		if e := ctx.Err(); e != nil {
			return nil, e
		}
		materials[key(m.Ref)] = m
	}
	for _, e := range in.Materials.Expected {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		expectations[key(e.Ref)] = e
	}
	find := func(ref statematerials.MaterialRef) (statematerials.Material, statematerials.Expectation, error) {
		if e := ctx.Err(); e != nil {
			return statematerials.Material{}, statematerials.Expectation{}, e
		}
		m, ok := materials[key(ref)]
		e, found := expectations[key(ref)]
		if !ok || !found {
			return m, e, ErrMissingMaterial
		}
		if m.Ref != ref || e.Ref != ref {
			return m, e, ErrBindingMismatch
		}
		return m, e, nil
	}
	anchor, e := stateformat.ParseEnvelope(stateformat.InitialRoot, in.Anchor.Raw)
	if e != nil {
		return nil, ErrInvalidMaterial
	}
	defer anchor.Clear()
	root, e := anchor.RootMaterial()
	if e != nil {
		return nil, ErrInvalidMaterial
	}
	defer root.Clear()
	b := anchor.Bindings()
	if b.RootVersion != 1 || root.Version != 1 || root.Threshold != 1 || len(root.Keys) != 1 {
		return nil, ErrInvalidMaterial
	}
	if b.Epoch != in.Anchor.Scope.Epoch || anchor.PayloadDigest() != in.Anchor.Payload || anchor.EnvelopeDigest() != in.Anchor.Envelope {
		return nil, ErrBindingMismatch
	}
	ar, ae, e := find(in.Target.AnchorRoot)
	if e != nil {
		return nil, e
	}
	raw := anchor.Payload()
	defer clear(raw)
	if in.Target.AnchorRoot.Payload != in.Anchor.Payload || ae.Root.Version != 1 || !bytes.Equal(raw, ar.Raw) {
		return nil, ErrBindingMismatch
	}
	s := &selection{}
	done := false
	defer func() {
		if !done {
			s.Clear()
		}
	}()
	s.batch = stateverify.Batch{ExpectedNodeID: in.Anchor.Scope.NodeID, ExpectedEpoch: in.Anchor.Scope.Epoch, AnchorEnvelope: in.Anchor.Raw, ActiveRootEnvelope: in.Anchor.Raw}
	s.finalRoot = in.Target.AnchorRoot
	s.rootVersion = 1
	if len(in.Target.Links) > 0 {
		s.batch.Rotations = make([][]byte, len(in.Target.Links))
	}
	for i, l := range in.Target.Links {
		rotation, re, e := find(l.Rotation)
		if e != nil {
			return nil, e
		}
		_, nr, e := find(l.NewRoot)
		if e != nil {
			return nil, e
		}
		if re.Rotation.OldRootPayload != s.finalRoot.Payload || re.Rotation.NewRootPayload != l.NewRoot.Payload || re.Rotation.PreviousVersion != s.rootVersion || nr.Root.Version != s.rootVersion+1 {
			return nil, ErrBindingMismatch
		}
		s.batch.Rotations[i] = rotation.Raw
		s.batch.ActiveRootEnvelope = rotation.Raw
		s.finalRoot = l.NewRoot
		s.rootVersion = nr.Root.Version
	}
	metadata, me, e := find(in.Target.Metadata)
	if e != nil {
		return nil, e
	}
	desired, de, e := find(in.Target.Desired)
	if e != nil {
		return nil, e
	}
	if me.Metadata.RootPayload != s.finalRoot.Payload || me.Metadata.RootVersion != s.rootVersion || de.Desired.RootPayload != s.finalRoot.Payload || de.Desired.MetadataPayload != in.Target.Metadata.Payload {
		return nil, ErrBindingMismatch
	}
	s.batch.MetadataEnvelope = metadata.Raw
	s.batch.DesiredEnvelope = desired.Raw
	s.metadataVersion = me.Metadata.Version
	s.desiredGeneration = de.Desired.Generation
	s.desiredSequence = de.Sequence
	s.desiredKeyID = de.Desired.SigningKeyID
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	done = true
	return s, nil
}
