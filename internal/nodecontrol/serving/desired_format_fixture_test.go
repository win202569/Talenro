package serving

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"os"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/store"
	"testing"
	"time"
)

func formatJSON(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if e := json.Unmarshal(b, &m); e != nil {
		t.Fatal(e)
	}
	return m
}
func formatBytes(t *testing.T, m any) []byte {
	t.Helper()
	b, e := json.Marshal(m)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func formatLiteral(t *testing.T, name string) []byte {
	t.Helper()
	b, e := os.ReadFile("../stateformat/testdata/" + name + ".json")
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func setFormatPublish(t *testing.T, p *desiredPublishProjection, name string, body map[string]any) {
	p.Payload = formatBytes(t, body)
	p.PayloadDigest = sha256.Sum256(p.Payload)
	e := formatJSON(t, formatLiteral(t, name+".envelope"))
	e["control_plane_authority_epoch"] = fmt.Sprint(p.Authority.Epoch)
	e["authority_sequence"] = fmt.Sprint(p.Authority.Sequence)
	root := p.Version
	if name == "metadata" {
		root = p.BaseRootVersion
	}
	e["root_version"] = fmt.Sprint(root)
	if name == "root-rotation" {
		e["previous_root_version"] = fmt.Sprint(p.BaseRootVersion)
	}
	e["payload_b64url"] = base64.RawURLEncoding.EncodeToString(p.Payload)
	e["payload_sha256"] = hex.EncodeToString(p.PayloadDigest[:])
	p.PublishedEnvelope = formatBytes(t, e)
	p.EnvelopeDigest = sha256.Sum256(p.PublishedEnvelope)
}

// Independent canonical fixture; the existing opaque fixtures remain unchanged.
// Distinct IDs/authority sequences separate node, signing, operation and publication.
func desiredFormatFixture(t *testing.T, n int) desiredProjection {
	t.Helper()
	s := desiredProjectionSnapshot{NodeID: desiredTestID(100), ActiveDesiredGeneration: 7, ActiveMetadataVersion: 6, ChainBaseRootVersion: 1}
	root := desiredPublishProjection{PublishID: desiredTestID(1), AuthorityOperationID: desiredTestID(1001), Version: 1, Authority: contracts.AuthorityVersion{Epoch: 19, Sequence: 3}}
	setFormatPublish(t, &root, "root-initial", formatJSON(t, formatLiteral(t, "root-initial.payload")))
	for i := 0; i < n; i++ {
		p := desiredPublishProjection{PublishID: desiredTestID(i + 2), AuthorityOperationID: desiredTestID(i + 1002), Version: uint64(i + 2), BaseRootVersion: uint64(i + 1), Authority: contracts.AuthorityVersion{Epoch: 19, Sequence: uint64(i + 4)}}
		body := formatJSON(t, formatLiteral(t, "root-rotation.payload"))
		body["authority_sequence"] = fmt.Sprint(p.Authority.Sequence)
		body["previous_root_version"] = fmt.Sprint(p.BaseRootVersion)
		inner := body["new_root_set"].(map[string]any)
		inner["authority_sequence"] = fmt.Sprint(p.Authority.Sequence)
		inner["root_version"] = fmt.Sprint(p.Version)
		setFormatPublish(t, &p, "root-rotation", body)
		s.RootRotations = append(s.RootRotations, p)
		root = cloneDesiredPublish(p)
	}
	s.ActiveRoot = root
	s.ActiveRootPublishID = root.PublishID
	s.ActiveRootVersion = root.Version
	meta := desiredPublishProjection{PublishID: desiredTestID(200), AuthorityOperationID: desiredTestID(1200), Version: 6, BaseRootVersion: root.Version, BaseMetadataVersion: 5, Authority: contracts.AuthorityVersion{Epoch: 19, Sequence: 80}}
	body := formatJSON(t, formatLiteral(t, "metadata.payload"))
	body["authority_sequence"] = "80"
	body["root_version"] = fmt.Sprint(root.Version)
	setFormatPublish(t, &meta, "metadata", body)
	s.ActiveMetadata = meta
	s.ActiveMetadataPublishID = meta.PublishID
	d := desiredArtifactProjection{NodeID: s.NodeID, SigningID: desiredTestID(300), AuthorityOperationID: desiredTestID(1300), RootPublishID: root.PublishID, MetadataPublishID: meta.PublishID, Generation: 7, InventoryVersion: 8, ResourceEnvelopeVersion: 9, RootVersion: root.Version, MetadataVersion: 6, Authority: contracts.AuthorityVersion{Epoch: 19, Sequence: 101}, Signature: [64]byte{2}, IssuedAt: time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)}
	d.ValidUntil = d.IssuedAt.Add(24 * time.Hour)
	d.EffectiveDeadline = d.IssuedAt.Add(time.Hour)
	for i := range d.ResourceEnvelopeDigest {
		d.ResourceEnvelopeDigest[i] = 0xab
	}
	key, _ := hex.DecodeString("01d0fabd251fcbbe2b93b4b927b26ad2a1a99077152e45ded1e678afa45dbec5")
	copy(d.SigningKeyID[:], key)
	body = formatJSON(t, formatLiteral(t, "desired.payload"))
	body["node_id"] = s.NodeID.String()
	body["authority_sequence"] = "101"
	d.Payload = formatBytes(t, body)
	d.PayloadDigest = sha256.Sum256(d.Payload)
	s.Desired = d
	return desiredProjection{snapshot: s}
}
func desiredFormatRows(p desiredProjection) []store.ListExactDesiredProjectionRow {
	s := p.snapshot
	d := s.Desired
	h := store.ListExactDesiredProjectionRow{NodeID: s.NodeID, PointerGroupsValid: true, HasDesired: true, CurrentValid: true,
		ActiveDesiredGeneration: int64(s.ActiveDesiredGeneration), ActiveRootPublishID: s.ActiveRootPublishID, ActiveRootVersion: int64(s.ActiveRootVersion), ActiveMetadataPublishID: s.ActiveMetadataPublishID, ActiveMetadataVersion: int64(s.ActiveMetadataVersion)}
	h.DesiredNodeID = d.NodeID
	h.DesiredSigningID = d.SigningID
	h.DesiredAuthorityOperationID = d.AuthorityOperationID
	h.DesiredRootPublishID = d.RootPublishID
	h.DesiredMetadataPublishID = d.MetadataPublishID
	h.DesiredGeneration = int64(d.Generation)
	h.DesiredInventoryVersion = int64(d.InventoryVersion)
	h.DesiredResourceEnvelopeVersion = int64(d.ResourceEnvelopeVersion)
	h.DesiredRootVersion = int64(d.RootVersion)
	h.DesiredMetadataVersion = int64(d.MetadataVersion)
	h.DesiredAuthorityEpoch = int64(d.Authority.Epoch)
	h.DesiredAuthoritySequence = int64(d.Authority.Sequence)
	h.DesiredResourceEnvelopeDigest = bytes.Clone(d.ResourceEnvelopeDigest[:])
	h.DesiredSigningKeyID = bytes.Clone(d.SigningKeyID[:])
	h.DesiredPayloadDigest = bytes.Clone(d.PayloadDigest[:])
	h.DesiredSignature = bytes.Clone(d.Signature[:])
	h.DesiredPayload = bytes.Clone(d.Payload)
	h.DesiredIssuedAt = pgtype.Timestamptz{Time: d.IssuedAt, Valid: true}
	h.DesiredEffectiveDeadline = pgtype.Timestamptz{Time: d.EffectiveDeadline, Valid: true}
	h.DesiredValidUntil = pgtype.Timestamptz{Time: d.ValidUntil, Valid: true}
	h.RootPublishID = s.ActiveRoot.PublishID
	h.RootAuthorityOperationID = s.ActiveRoot.AuthorityOperationID
	h.RootVersion = int64(s.ActiveRoot.Version)
	h.RootBaseRootVersion = int64(s.ActiveRoot.BaseRootVersion)
	h.RootBaseMetadataVersion = int64(s.ActiveRoot.BaseMetadataVersion)
	h.RootAuthorityEpoch = int64(s.ActiveRoot.Authority.Epoch)
	h.RootAuthoritySequence = int64(s.ActiveRoot.Authority.Sequence)
	h.RootPayload = bytes.Clone(s.ActiveRoot.Payload)
	h.RootPublishedEnvelope = bytes.Clone(s.ActiveRoot.PublishedEnvelope)
	h.RootPayloadDigest = bytes.Clone(s.ActiveRoot.PayloadDigest[:])
	h.RootEnvelopeDigest = bytes.Clone(s.ActiveRoot.EnvelopeDigest[:])
	h.MetadataPublishID = s.ActiveMetadata.PublishID
	h.MetadataAuthorityOperationID = s.ActiveMetadata.AuthorityOperationID
	h.MetadataVersion = int64(s.ActiveMetadata.Version)
	h.MetadataBaseRootVersion = int64(s.ActiveMetadata.BaseRootVersion)
	h.MetadataBaseMetadataVersion = int64(s.ActiveMetadata.BaseMetadataVersion)
	h.MetadataAuthorityEpoch = int64(s.ActiveMetadata.Authority.Epoch)
	h.MetadataAuthoritySequence = int64(s.ActiveMetadata.Authority.Sequence)
	h.MetadataPayload = bytes.Clone(s.ActiveMetadata.Payload)
	h.MetadataPublishedEnvelope = bytes.Clone(s.ActiveMetadata.PublishedEnvelope)
	h.MetadataPayloadDigest = bytes.Clone(s.ActiveMetadata.PayloadDigest[:])
	h.MetadataEnvelopeDigest = bytes.Clone(s.ActiveMetadata.EnvelopeDigest[:])
	if len(s.RootRotations) == 0 {
		return []store.ListExactDesiredProjectionRow{h}
	}
	rows := make([]store.ListExactDesiredProjectionRow, len(s.RootRotations))
	for i, p := range s.RootRotations {
		r := h
		r.HasRotation = true
		r.RotationValid = true
		r.RotationExpectedVersion = int64(i + 2)
		r.RotationPublishID = p.PublishID
		r.RotationAuthorityOperationID = p.AuthorityOperationID
		r.RotationVersion = int64(p.Version)
		r.RotationBaseRootVersion = int64(p.BaseRootVersion)
		r.RotationBaseMetadataVersion = int64(p.BaseMetadataVersion)
		r.RotationAuthorityEpoch = int64(p.Authority.Epoch)
		r.RotationAuthoritySequence = int64(p.Authority.Sequence)
		r.RotationPayload = bytes.Clone(p.Payload)
		r.RotationPublishedEnvelope = bytes.Clone(p.PublishedEnvelope)
		r.RotationPayloadDigest = bytes.Clone(p.PayloadDigest[:])
		r.RotationEnvelopeDigest = bytes.Clone(p.EnvelopeDigest[:])
		rows[i] = r
	}
	return rows
}
