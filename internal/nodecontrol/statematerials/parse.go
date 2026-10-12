package statematerials

import (
	"context"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/stateformat"
)

type parsedItem struct {
	root     *stateformat.ParsedRootSet
	artifact *stateformat.ParsedArtifact
	expected Expectation
}
type parsedBatch struct{ items map[materialKey]*parsedItem }

func (p *parsedBatch) Clear() {
	if p != nil {
		for _, item := range p.items {
			item.root.Clear()
			item.artifact.Clear()
			*item = parsedItem{}
		}
		p.items = nil
	}
}
func parseBatch(ctx context.Context, in Input) (result *parsedBatch, err error) {
	p := &parsedBatch{items: map[materialKey]*parsedItem{}}
	defer func() {
		if result == nil {
			p.Clear()
		}
	}()
	expected := map[materialKey]Expectation{}
	for _, e := range in.Expected {
		expected[key(e.Ref)] = e
	}
	for _, m := range in.Materials {
		if e := ctx.Err(); e != nil {
			return nil, e
		}
		item := &parsedItem{expected: expected[key(m.Ref)]}
		p.items[key(m.Ref)] = item
		if m.Ref.Kind == RootSet {
			item.root, err = stateformat.ParseRootSet(m.Raw)
		} else {
			kind := map[MaterialKind]stateformat.Kind{RootRotation: stateformat.RootRotation, Metadata: stateformat.Metadata, Desired: stateformat.Desired}[m.Ref.Kind]
			item.artifact, err = stateformat.ParseEnvelope(kind, m.Raw)
		}
		if e := ctx.Err(); e != nil {
			return nil, e
		}
		if err != nil {
			return nil, ErrInvalidMaterial
		}
		if err = matchParsed(item, m.Ref); err != nil {
			return nil, err
		}
		if e := ctx.Err(); e != nil {
			return nil, e
		}
	}
	return p, nil
}
func keyIDs(keys []stateformat.PublicKeyMaterial) []contracts.Digest {
	var ids []contracts.Digest
	for _, k := range keys {
		ids = append(ids, k.KeyID)
	}
	return ids
}
func matchParsed(item *parsedItem, ref MaterialRef) error {
	e := item.expected
	if ref.Kind == RootSet {
		r := item.root
		b := r.Bindings()
		m, err := r.RootMaterial()
		if err != nil {
			return ErrInvalidMaterial
		}
		defer m.Clear()
		if r.PayloadDigest() != ref.Payload || b.Epoch != e.Epoch || b.Sequence != e.Sequence || b.Version != e.Root.Version || m.Threshold != e.Root.Threshold || !reflect.DeepEqual(keyIDs(m.Keys), e.Root.KeyIDs) {
			return ErrReferenceMismatch
		}
		return nil
	}
	a := item.artifact
	b := a.Bindings()
	if a.PayloadDigest() != ref.Payload || a.EnvelopeDigest() != ref.Envelope || b.Epoch != e.Epoch || b.Sequence != e.Sequence {
		return ErrReferenceMismatch
	}
	w := Window{b.IssuedAt.Unix(), b.ValidUntil.Unix()}
	switch ref.Kind {
	case RootRotation:
		r := e.Rotation
		if b.PreviousRootVersion != r.PreviousVersion || b.RootVersion != r.NewVersion || w != r.Window {
			return ErrReferenceMismatch
		}
	case Metadata:
		m, err := a.MetadataMaterial()
		if err != nil {
			return ErrInvalidMaterial
		}
		defer m.Clear()
		var ids []contracts.Digest
		for _, k := range m.Keys {
			ids = append(ids, k.KeyID)
		}
		want := e.Metadata
		if b.MetadataVersion != want.Version || b.RootVersion != want.RootVersion || w != want.Window || !reflect.DeepEqual(ids, want.KeyIDs) || !equalDigests(m.RevokedIDs, want.RevokedIDs) {
			return ErrReferenceMismatch
		}
	case Desired:
		d := e.Desired
		slots, err := a.DesiredSlotIDs()
		if err != nil {
			return ErrInvalidMaterial
		}
		defer clear(slots)
		if b.NodeID.String() != d.NodeID || b.Generation != d.Generation || b.InventoryVersion != d.InventoryVersion || b.ResourceEnvelopeVersion != d.ResourceVersion || b.ResourceEnvelopeDigest != d.ResourcePayload || b.SigningKeyID != d.SigningKeyID || w != d.Window || !equalStrings(slots, d.SlotIDs) {
			return ErrReferenceMismatch
		}
	}
	return nil
}
func equalDigests(a, b []contracts.Digest) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
