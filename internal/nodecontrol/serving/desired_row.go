package serving

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/store"
)

// desiredProjectionFromRows converts private staging rows, not authenticated
// facts. SQL validity flags are not authority when supplied by arbitrary callers.
// A future bound adapter must supply the exact query and same-source epoch.
func desiredProjectionFromRows(nodeID uuid.UUID, expectedEpoch uint64, rows []store.ListExactDesiredProjectionRow) (desiredProjection, error) {
	if !factsUUID(nodeID) || !factsCoordinate(expectedEpoch) {
		return desiredProjection{}, authority.ErrInvalidArgument
	}
	if len(rows) == 0 {
		return desiredProjection{}, authority.ErrNotFound
	}
	if len(rows) > 32 {
		return desiredProjection{}, authority.ErrConflict
	}
	r := rows[0]
	if r.NodeID != nodeID || !r.PointerGroupsValid ||
		!desiredPointerPair(r.ActiveRootPublishID, r.ActiveRootVersion) ||
		!desiredPointerPair(r.ActiveMetadataPublishID, r.ActiveMetadataVersion) {
		return desiredProjection{}, authority.ErrConflict
	}
	if !r.HasDesired {
		if len(rows) != 1 || r.ActiveDesiredGeneration != 0 || r.CurrentValid || r.HasRotation || r.RotationValid || !emptyDesiredRotation(r) {
			return desiredProjection{}, authority.ErrConflict
		}
		return desiredProjection{}, authority.ErrNotFound
	}
	if !r.CurrentValid || r.ActiveDesiredGeneration <= 0 || r.ActiveRootVersion < 1 || r.ActiveRootVersion > 33 || r.ActiveMetadataVersion <= 0 {
		return desiredProjection{}, authority.ErrConflict
	}
	count := int(r.ActiveRootVersion - 1)
	expectedRows := count
	if expectedRows == 0 {
		expectedRows = 1
	}
	if len(rows) != expectedRows {
		return desiredProjection{}, authority.ErrConflict
	}
	for _, next := range rows[1:] {
		if !equalDesiredRowHeader(r, next) {
			return desiredProjection{}, authority.ErrConflict
		}
	}
	if r.DesiredGeneration <= 0 || r.DesiredInventoryVersion <= 0 || r.DesiredResourceEnvelopeVersion <= 0 ||
		r.DesiredRootVersion <= 0 || r.DesiredMetadataVersion <= 0 || r.DesiredAuthorityEpoch <= 0 || r.DesiredAuthoritySequence <= 0 ||
		len(r.DesiredResourceEnvelopeDigest) != 32 || len(r.DesiredSigningKeyID) != 32 || len(r.DesiredPayloadDigest) != 32 ||
		len(r.DesiredSignature) != 64 || !finiteDesiredTime(r.DesiredIssuedAt) || !finiteDesiredTime(r.DesiredEffectiveDeadline) || !finiteDesiredTime(r.DesiredValidUntil) {
		return desiredProjection{}, authority.ErrConflict
	}
	root, rootOK := desiredRootFromRow(r)
	meta, metaOK := desiredMetadataFromRow(r)
	if !rootOK || !metaOK {
		return desiredProjection{}, authority.ErrConflict
	}
	s := desiredProjectionSnapshot{
		NodeID: r.NodeID, ActiveDesiredGeneration: uint64(r.ActiveDesiredGeneration),
		ActiveRootPublishID: r.ActiveRootPublishID, ActiveRootVersion: uint64(r.ActiveRootVersion),
		ActiveMetadataPublishID: r.ActiveMetadataPublishID, ActiveMetadataVersion: uint64(r.ActiveMetadataVersion),
		ChainBaseRootVersion: 1, ActiveRoot: root, ActiveMetadata: meta,
		Desired: desiredArtifactProjection{
			NodeID:                  r.DesiredNodeID,
			SigningID:               r.DesiredSigningID,
			AuthorityOperationID:    r.DesiredAuthorityOperationID,
			RootPublishID:           r.DesiredRootPublishID,
			MetadataPublishID:       r.DesiredMetadataPublishID,
			Generation:              uint64(r.DesiredGeneration),
			InventoryVersion:        uint64(r.DesiredInventoryVersion),
			ResourceEnvelopeVersion: uint64(r.DesiredResourceEnvelopeVersion),
			RootVersion:             uint64(r.DesiredRootVersion),
			MetadataVersion:         uint64(r.DesiredMetadataVersion),
			ResourceEnvelopeDigest:  contracts.Digest(r.DesiredResourceEnvelopeDigest),
			SigningKeyID:            contracts.Digest(r.DesiredSigningKeyID),
			PayloadDigest:           contracts.Digest(r.DesiredPayloadDigest),
			Signature:               [64]byte(r.DesiredSignature),
			Payload:                 r.DesiredPayload,
			IssuedAt:                r.DesiredIssuedAt.Time,
			EffectiveDeadline:       r.DesiredEffectiveDeadline.Time,
			ValidUntil:              r.DesiredValidUntil.Time,
			Authority:               contracts.AuthorityVersion{Epoch: uint64(r.DesiredAuthorityEpoch), Sequence: uint64(r.DesiredAuthoritySequence)},
		},
	}
	if count == 0 {
		if r.HasRotation || r.RotationValid || !emptyDesiredRotation(r) {
			return desiredProjection{}, authority.ErrConflict
		}
	} else {
		s.RootRotations = make([]desiredPublishProjection, 0, count)
		for i, row := range rows {
			if !row.HasRotation || !row.RotationValid || row.RotationExpectedVersion != int64(i+2) || row.RotationVersion != row.RotationExpectedVersion {
				return desiredProjection{}, authority.ErrConflict
			}
			p, ok := desiredRotationFromRow(row)
			if !ok {
				return desiredProjection{}, authority.ErrConflict
			}
			s.RootRotations = append(s.RootRotations, p)
		}
	}
	// Only the constructor owns copies. All intermediate slices still belong to
	// the caller; neither successful nor failed conversion mutates those bytes.
	return newDesiredProjection(nodeID, expectedEpoch, s)
}
func desiredPointerPair(id uuid.UUID, v int64) bool {
	return (id == uuid.Nil && v == 0) || (factsUUID(id) && v > 0)
}
func finiteDesiredTime(t pgtype.Timestamptz) bool {
	return t.Valid && t.InfinityModifier == pgtype.Finite
}

func desiredRootFromRow(r store.ListExactDesiredProjectionRow) (desiredPublishProjection, bool) {
	if r.RootVersion <= 0 || r.RootBaseRootVersion < 0 || r.RootBaseMetadataVersion < 0 || r.RootAuthorityEpoch <= 0 || r.RootAuthoritySequence <= 0 || len(r.RootPayloadDigest) != 32 || len(r.RootEnvelopeDigest) != 32 {
		return desiredPublishProjection{}, false
	}
	return desiredPublishProjection{
		PublishID:            r.RootPublishID,
		AuthorityOperationID: r.RootAuthorityOperationID,
		Version:              uint64(r.RootVersion),
		BaseRootVersion:      uint64(r.RootBaseRootVersion),
		BaseMetadataVersion:  uint64(r.RootBaseMetadataVersion),
		Payload:              r.RootPayload,
		PublishedEnvelope:    r.RootPublishedEnvelope,
		PayloadDigest:        contracts.Digest(r.RootPayloadDigest),
		EnvelopeDigest:       contracts.Digest(r.RootEnvelopeDigest),
		Authority:            contracts.AuthorityVersion{Epoch: uint64(r.RootAuthorityEpoch), Sequence: uint64(r.RootAuthoritySequence)},
	}, true
}
func desiredMetadataFromRow(r store.ListExactDesiredProjectionRow) (desiredPublishProjection, bool) {
	if r.MetadataVersion <= 0 || r.MetadataBaseRootVersion < 0 || r.MetadataBaseMetadataVersion < 0 || r.MetadataAuthorityEpoch <= 0 || r.MetadataAuthoritySequence <= 0 || len(r.MetadataPayloadDigest) != 32 || len(r.MetadataEnvelopeDigest) != 32 {
		return desiredPublishProjection{}, false
	}
	return desiredPublishProjection{
		PublishID:            r.MetadataPublishID,
		AuthorityOperationID: r.MetadataAuthorityOperationID,
		Version:              uint64(r.MetadataVersion),
		BaseRootVersion:      uint64(r.MetadataBaseRootVersion),
		BaseMetadataVersion:  uint64(r.MetadataBaseMetadataVersion),
		Payload:              r.MetadataPayload,
		PublishedEnvelope:    r.MetadataPublishedEnvelope,
		PayloadDigest:        contracts.Digest(r.MetadataPayloadDigest),
		EnvelopeDigest:       contracts.Digest(r.MetadataEnvelopeDigest),
		Authority:            contracts.AuthorityVersion{Epoch: uint64(r.MetadataAuthorityEpoch), Sequence: uint64(r.MetadataAuthoritySequence)},
	}, true
}
func desiredRotationFromRow(r store.ListExactDesiredProjectionRow) (desiredPublishProjection, bool) {
	if r.RotationVersion <= 0 || r.RotationBaseRootVersion < 0 || r.RotationBaseMetadataVersion < 0 || r.RotationAuthorityEpoch <= 0 || r.RotationAuthoritySequence <= 0 || len(r.RotationPayloadDigest) != 32 || len(r.RotationEnvelopeDigest) != 32 {
		return desiredPublishProjection{}, false
	}
	return desiredPublishProjection{
		PublishID:            r.RotationPublishID,
		AuthorityOperationID: r.RotationAuthorityOperationID,
		Version:              uint64(r.RotationVersion),
		BaseRootVersion:      uint64(r.RotationBaseRootVersion),
		BaseMetadataVersion:  uint64(r.RotationBaseMetadataVersion),
		Payload:              r.RotationPayload,
		PublishedEnvelope:    r.RotationPublishedEnvelope,
		PayloadDigest:        contracts.Digest(r.RotationPayloadDigest),
		EnvelopeDigest:       contracts.Digest(r.RotationEnvelopeDigest),
		Authority:            contracts.AuthorityVersion{Epoch: uint64(r.RotationAuthorityEpoch), Sequence: uint64(r.RotationAuthoritySequence)},
	}, true
}
func emptyDesiredRotation(r store.ListExactDesiredProjectionRow) bool {
	return r.RotationExpectedVersion == 0 &&
		r.RotationPublishID == uuid.Nil &&
		r.RotationAuthorityOperationID == uuid.Nil &&
		r.RotationVersion == 0 &&
		r.RotationBaseRootVersion == 0 &&
		r.RotationBaseMetadataVersion == 0 &&
		r.RotationAuthorityEpoch == 0 &&
		r.RotationAuthoritySequence == 0 &&
		len(r.RotationPayload) == 0 &&
		len(r.RotationPublishedEnvelope) == 0 &&
		len(r.RotationPayloadDigest) == 0 &&
		len(r.RotationEnvelopeDigest) == 0
}
func equalDesiredRowHeader(a, b store.ListExactDesiredProjectionRow) bool {
	return a.NodeID == b.NodeID &&
		a.PointerGroupsValid == b.PointerGroupsValid &&
		a.HasDesired == b.HasDesired &&
		a.CurrentValid == b.CurrentValid &&
		a.ActiveDesiredGeneration == b.ActiveDesiredGeneration &&
		a.ActiveRootPublishID == b.ActiveRootPublishID &&
		a.ActiveRootVersion == b.ActiveRootVersion &&
		a.ActiveMetadataPublishID == b.ActiveMetadataPublishID &&
		a.ActiveMetadataVersion == b.ActiveMetadataVersion &&
		a.DesiredNodeID == b.DesiredNodeID &&
		a.DesiredSigningID == b.DesiredSigningID &&
		a.DesiredAuthorityOperationID == b.DesiredAuthorityOperationID &&
		a.DesiredRootPublishID == b.DesiredRootPublishID &&
		a.DesiredMetadataPublishID == b.DesiredMetadataPublishID &&
		a.DesiredGeneration == b.DesiredGeneration &&
		a.DesiredInventoryVersion == b.DesiredInventoryVersion &&
		a.DesiredResourceEnvelopeVersion == b.DesiredResourceEnvelopeVersion &&
		a.DesiredRootVersion == b.DesiredRootVersion &&
		a.DesiredMetadataVersion == b.DesiredMetadataVersion &&
		a.DesiredAuthorityEpoch == b.DesiredAuthorityEpoch &&
		a.DesiredAuthoritySequence == b.DesiredAuthoritySequence &&
		bytes.Equal(a.DesiredResourceEnvelopeDigest, b.DesiredResourceEnvelopeDigest) &&
		bytes.Equal(a.DesiredSigningKeyID, b.DesiredSigningKeyID) &&
		bytes.Equal(a.DesiredPayloadDigest, b.DesiredPayloadDigest) &&
		bytes.Equal(a.DesiredSignature, b.DesiredSignature) &&
		bytes.Equal(a.DesiredPayload, b.DesiredPayload) &&
		a.DesiredIssuedAt == b.DesiredIssuedAt &&
		a.DesiredEffectiveDeadline == b.DesiredEffectiveDeadline &&
		a.DesiredValidUntil == b.DesiredValidUntil &&
		a.RootPublishID == b.RootPublishID &&
		a.RootAuthorityOperationID == b.RootAuthorityOperationID &&
		a.RootVersion == b.RootVersion &&
		a.RootBaseRootVersion == b.RootBaseRootVersion &&
		a.RootBaseMetadataVersion == b.RootBaseMetadataVersion &&
		a.RootAuthorityEpoch == b.RootAuthorityEpoch &&
		a.RootAuthoritySequence == b.RootAuthoritySequence &&
		bytes.Equal(a.RootPayload, b.RootPayload) &&
		bytes.Equal(a.RootPublishedEnvelope, b.RootPublishedEnvelope) &&
		bytes.Equal(a.RootPayloadDigest, b.RootPayloadDigest) &&
		bytes.Equal(a.RootEnvelopeDigest, b.RootEnvelopeDigest) &&
		a.MetadataPublishID == b.MetadataPublishID &&
		a.MetadataAuthorityOperationID == b.MetadataAuthorityOperationID &&
		a.MetadataVersion == b.MetadataVersion &&
		a.MetadataBaseRootVersion == b.MetadataBaseRootVersion &&
		a.MetadataBaseMetadataVersion == b.MetadataBaseMetadataVersion &&
		a.MetadataAuthorityEpoch == b.MetadataAuthorityEpoch &&
		a.MetadataAuthoritySequence == b.MetadataAuthoritySequence &&
		bytes.Equal(a.MetadataPayload, b.MetadataPayload) &&
		bytes.Equal(a.MetadataPublishedEnvelope, b.MetadataPublishedEnvelope) &&
		bytes.Equal(a.MetadataPayloadDigest, b.MetadataPayloadDigest) &&
		bytes.Equal(a.MetadataEnvelopeDigest, b.MetadataEnvelopeDigest)
}
