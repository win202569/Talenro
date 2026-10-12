package serving

import (
	"bytes"
	"crypto/sha256"
	"github.com/google/uuid"
	"math"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"time"
)

type desiredPublishProjection struct {
	PublishID, AuthorityOperationID               uuid.UUID
	Version, BaseRootVersion, BaseMetadataVersion uint64
	Authority                                     contracts.AuthorityVersion
	Payload, PublishedEnvelope                    []byte
	PayloadDigest, EnvelopeDigest                 contracts.Digest
}
type desiredArtifactProjection struct {
	NodeID, SigningID, AuthorityOperationID, RootPublishID, MetadataPublishID           uuid.UUID
	Generation, InventoryVersion, ResourceEnvelopeVersion, RootVersion, MetadataVersion uint64
	Authority                                                                           contracts.AuthorityVersion
	ResourceEnvelopeDigest, SigningKeyID, PayloadDigest                                 contracts.Digest
	Payload                                                                             []byte
	Signature                                                                           [64]byte
	IssuedAt, EffectiveDeadline, ValidUntil                                             time.Time
}
type desiredProjectionSnapshot struct {
	NodeID, ActiveRootPublishID, ActiveMetadataPublishID                                    uuid.UUID
	ActiveDesiredGeneration, ActiveRootVersion, ActiveMetadataVersion, ChainBaseRootVersion uint64
	RootRotations                                                                           []desiredPublishProjection
	ActiveRoot, ActiveMetadata                                                              desiredPublishProjection
	Desired                                                                                 desiredArtifactProjection
}

// desiredProjection is owned structural staging only. It does not establish
// canonical encoding, signatures, active intent/fence provenance, trusted time,
// or a trusted chain starting point. No serving entry accepts this as authority.
type desiredProjection struct{ snapshot desiredProjectionSnapshot }

func newDesiredProjection(nodeID uuid.UUID, epoch uint64, s desiredProjectionSnapshot) (desiredProjection, error) {
	if !factsUUID(nodeID) || !factsCoordinate(epoch) {
		return desiredProjection{}, authority.ErrInvalidArgument
	}
	d := s.Desired
	if s.NodeID != nodeID || d.NodeID != nodeID ||
		!factsCoordinate(s.ActiveDesiredGeneration) || s.ActiveDesiredGeneration != d.Generation ||
		!factsCoordinate(s.ActiveRootVersion) || !factsCoordinate(s.ActiveMetadataVersion) ||
		s.ActiveRootPublishID != s.ActiveRoot.PublishID || s.ActiveRootVersion != s.ActiveRoot.Version ||
		s.ActiveMetadataPublishID != s.ActiveMetadata.PublishID || s.ActiveMetadataVersion != s.ActiveMetadata.Version ||
		d.RootPublishID != s.ActiveRootPublishID || d.RootVersion != s.ActiveRootVersion ||
		d.MetadataPublishID != s.ActiveMetadataPublishID || d.MetadataVersion != s.ActiveMetadataVersion ||
		!validDesiredPublish(s.ActiveRoot, epoch) || s.ActiveRoot.BaseRootVersion != s.ActiveRoot.Version-1 ||
		!validDesiredPublish(s.ActiveMetadata, epoch) || s.ActiveMetadata.BaseMetadataVersion != s.ActiveMetadata.Version-1 ||
		s.ActiveMetadata.BaseRootVersion != s.ActiveRootVersion ||
		!factsUUID(d.SigningID) || !factsUUID(d.AuthorityOperationID) || d.Authority.Validate() != nil || d.Authority.Epoch != epoch ||
		!factsCoordinate(d.InventoryVersion) || !factsCoordinate(d.ResourceEnvelopeVersion) ||
		d.ResourceEnvelopeDigest == (contracts.Digest{}) || d.SigningKeyID == (contracts.Digest{}) || d.Signature == ([64]byte{}) ||
		!validDesiredBytes(d.Payload, d.PayloadDigest, 65536) ||
		d.IssuedAt.IsZero() || d.EffectiveDeadline.IsZero() || d.ValidUntil.IsZero() ||
		!d.EffectiveDeadline.After(d.IssuedAt) || d.EffectiveDeadline.After(d.ValidUntil) || d.ValidUntil.After(d.IssuedAt.Add(24*time.Hour)) ||
		!factsCoordinate(s.ChainBaseRootVersion) || len(s.RootRotations) > 32 {
		return desiredProjection{}, authority.ErrConflict
	}
	version := s.ChainBaseRootVersion
	var sequence uint64
	seen := make(map[uuid.UUID]struct{}, len(s.RootRotations))
	for _, root := range s.RootRotations {
		if !validDesiredPublish(root, epoch) || version == math.MaxInt64 || root.BaseRootVersion != version || root.Version != version+1 || root.Authority.Sequence <= sequence {
			return desiredProjection{}, authority.ErrConflict
		}
		if _, duplicate := seen[root.PublishID]; duplicate {
			return desiredProjection{}, authority.ErrConflict
		}
		seen[root.PublishID] = struct{}{}
		version = root.Version
		sequence = root.Authority.Sequence
	}
	if version != s.ActiveRootVersion {
		return desiredProjection{}, authority.ErrConflict
	}
	if len(s.RootRotations) > 0 && !equalDesiredPublish(s.RootRotations[len(s.RootRotations)-1], s.ActiveRoot) {
		return desiredProjection{}, authority.ErrConflict
	}
	return desiredProjection{snapshot: cloneDesiredProjectionSnapshot(s)}, nil
}
func validDesiredBytes(b []byte, d contracts.Digest, limit int) bool {
	return len(b) > 0 && len(b) <= limit && d != (contracts.Digest{}) && sha256.Sum256(b) == d
}
func validDesiredPublish(p desiredPublishProjection, epoch uint64) bool {
	return factsUUID(p.PublishID) && factsUUID(p.AuthorityOperationID) && factsCoordinate(p.Version) && p.BaseRootVersion <= math.MaxInt64 && p.BaseMetadataVersion <= math.MaxInt64 && p.Authority.Validate() == nil && p.Authority.Epoch == epoch && validDesiredBytes(p.Payload, p.PayloadDigest, 65536) && validDesiredBytes(p.PublishedEnvelope, p.EnvelopeDigest, 262144)
}
func equalDesiredPublish(a, b desiredPublishProjection) bool {
	return a.PublishID == b.PublishID && a.AuthorityOperationID == b.AuthorityOperationID && a.Version == b.Version && a.BaseRootVersion == b.BaseRootVersion && a.BaseMetadataVersion == b.BaseMetadataVersion && a.Authority == b.Authority && a.PayloadDigest == b.PayloadDigest && a.EnvelopeDigest == b.EnvelopeDigest && bytes.Equal(a.Payload, b.Payload) && bytes.Equal(a.PublishedEnvelope, b.PublishedEnvelope)
}
func (p desiredProjection) Snapshot() desiredProjectionSnapshot {
	return cloneDesiredProjectionSnapshot(p.snapshot)
}
func cloneDesiredPublish(p desiredPublishProjection) desiredPublishProjection {
	p.Payload = bytes.Clone(p.Payload)
	p.PublishedEnvelope = bytes.Clone(p.PublishedEnvelope)
	return p
}
func cloneDesiredProjectionSnapshot(s desiredProjectionSnapshot) desiredProjectionSnapshot {
	if s.RootRotations != nil {
		roots := make([]desiredPublishProjection, len(s.RootRotations))
		for i, p := range s.RootRotations {
			roots[i] = cloneDesiredPublish(p)
		}
		s.RootRotations = roots
	}
	s.ActiveRoot = cloneDesiredPublish(s.ActiveRoot)
	s.ActiveMetadata = cloneDesiredPublish(s.ActiveMetadata)
	s.Desired.Payload = bytes.Clone(s.Desired.Payload)
	return s
}

// Clear only exclusively owned staging material. Caller input and detached
// snapshots remain untouched; this does not promise erasure of runtime copies.
func discardDesiredProjection(p *desiredProjection) {
	for i := range p.snapshot.RootRotations {
		clear(p.snapshot.RootRotations[i].Payload)
		clear(p.snapshot.RootRotations[i].PublishedEnvelope)
	}
	clear(p.snapshot.RootRotations)
	clear(p.snapshot.ActiveRoot.Payload)
	clear(p.snapshot.ActiveRoot.PublishedEnvelope)
	clear(p.snapshot.ActiveMetadata.Payload)
	clear(p.snapshot.ActiveMetadata.PublishedEnvelope)
	clear(p.snapshot.Desired.Payload)
	*p = desiredProjection{}
}
