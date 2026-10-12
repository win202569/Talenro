package statematerials

import (
	"bytes"
	"context"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

func validateRelations(ctx context.Context, in Input, p *parsedBatch) error {
	versions := map[uint64]contracts.Digest{}
	for _, e := range in.Expected {
		if err := ctx.Err(); err != nil {
			return err
		}
		if e.Ref.Kind == RootSet {
			if _, ok := versions[e.Root.Version]; ok {
				return ErrReferenceMismatch
			}
			versions[e.Root.Version] = e.Ref.Payload
		}
	}
	edges := map[contracts.Digest]bool{}
	for _, e := range in.Expected {
		if err := ctx.Err(); err != nil {
			return err
		}
		switch e.Ref.Kind {
		case RootRotation:
			r := e.Rotation
			old := p.items[materialKey{RootSet, r.OldRootPayload}]
			next := p.items[materialKey{RootSet, r.NewRootPayload}]
			if old == nil || next == nil {
				return ErrMissingMaterial
			}
			if r.PreviousVersion == maxCoordinate || r.NewVersion != r.PreviousVersion+1 || old.expected.Root.Version != r.PreviousVersion || next.expected.Root.Version != r.NewVersion || old.expected.Sequence >= next.expected.Sequence || next.expected.Sequence != e.Sequence || edges[r.OldRootPayload] {
				return ErrReferenceMismatch
			}
			edges[r.OldRootPayload] = true
			embedded, err := p.items[key(e.Ref)].artifact.EmbeddedRootSet()
			if err != nil {
				return ErrInvalidMaterial
			}
			left, right := embedded.Payload(), next.root.Payload()
			same := embedded.PayloadDigest() == r.NewRootPayload && bytes.Equal(left, right)
			clear(left)
			clear(right)
			embedded.Clear()
			if !same {
				return ErrReferenceMismatch
			}
		case Metadata:
			m := e.Metadata
			root := p.items[materialKey{RootSet, m.RootPayload}]
			if root == nil {
				return ErrMissingMaterial
			}
			if root.expected.Root.Version != m.RootVersion {
				return ErrReferenceMismatch
			}
		case Desired:
			d := e.Desired
			metadata := p.items[materialKey{Metadata, d.MetadataPayload}]
			root := p.items[materialKey{RootSet, d.RootPayload}]
			if metadata == nil || root == nil {
				return ErrMissingMaterial
			}
			if d.NodeID != in.Scope.NodeID || metadata.expected.Metadata.RootPayload != d.RootPayload {
				return ErrReferenceMismatch
			}
			found := false
			for _, id := range metadata.expected.Metadata.KeyIDs {
				if err := ctx.Err(); err != nil {
					return err
				}
				if id == d.SigningKeyID {
					found = true
				}
			}
			if !found {
				return ErrReferenceMismatch
			}
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return nil
}
