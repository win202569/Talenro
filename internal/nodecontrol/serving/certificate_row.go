package serving

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/store"
)

// certificateFactsFromRows converts driver values into private staged facts.
// It neither executes SQL nor proves ready authority/fence provenance. The
// expected epoch must ultimately be captured by the same bound read source.
// Input rows remain caller-owned; no driver buffer is retained in the result.
func certificateFactsFromRows(lookup CertificateLookup, expectedEpoch uint64, rows []store.ListExactCertificateProjectionRow) (CertificateFacts, error) {
	if lookup.Validate() != nil || !factsCoordinate(expectedEpoch) {
		return CertificateFacts{}, authority.ErrInvalidArgument
	}
	if len(rows) == 0 {
		return CertificateFacts{}, authority.ErrNotFound
	}
	if len(rows) != 1 {
		return CertificateFacts{}, authority.ErrConflict
	}
	r := rows[0]
	if r.AuthorityEpoch <= 0 || uint64(r.AuthorityEpoch) != expectedEpoch || r.AuthoritySequence <= 0 ||
		r.IdentityEpoch <= 0 || r.InventoryIdentityEpoch <= 0 || r.InventoryVersion <= 0 ||
		!r.InventoryLineageID.Valid || !r.ValidFrom.Valid || !r.ValidUntil.Valid ||
		r.ValidFrom.InfinityModifier != pgtype.Finite || r.ValidUntil.InfinityModifier != pgtype.Finite ||
		len(r.LeafDerSha256) != 32 || len(r.PublicKeySha256) != 32 {
		return CertificateFacts{}, authority.ErrConflict
	}
	_, agent := lookup.AgentTrustSelector()
	if r.HasServerTrust != agent {
		return CertificateFacts{}, authority.ErrConflict
	}
	if agent {
		if r.TrustAuthorityEpoch <= 0 || uint64(r.TrustAuthorityEpoch) != expectedEpoch ||
			r.TrustAuthoritySequence <= 0 || r.TrustBundleVersion <= 0 || len(r.TrustBundleDigest) != 32 {
			return CertificateFacts{}, authority.ErrConflict
		}
	} else if r.TrustPurpose != "" || r.TrustListenerKind != "" || r.TrustDomain != "" ||
		r.TrustOperationID != uuid.Nil || r.TrustAuthorityEpoch != 0 || r.TrustAuthoritySequence != 0 ||
		r.TrustBundleVersion != 0 || len(r.TrustBundleDigest) != 0 {
		return CertificateFacts{}, authority.ErrConflict
	}
	s := CertificateSnapshot{
		Identity: CertificateLookupInput{NodeID: r.NodeID, IssuerID: r.IssuerID, SerialBytes: r.SerialBytes,
			LeafDERDigest: contracts.Digest(r.LeafDerSha256), PublicKeyDigest: contracts.Digest(r.PublicKeySha256)},
		CertificateID: r.CertificateID, LineageID: r.LineageID, IdentityEpoch: uint64(r.IdentityEpoch),
		InventoryIdentityEpoch: uint64(r.InventoryIdentityEpoch), InventoryLineageID: r.InventoryLineageID.UUID,
		InventoryVersion: uint64(r.InventoryVersion), AuthorityOperationID: r.AuthorityOperationID,
		Authority: contracts.AuthorityVersion{Epoch: uint64(r.AuthorityEpoch), Sequence: uint64(r.AuthoritySequence)},
		Status:    r.Status, ValidFrom: r.ValidFrom.Time, ValidUntil: r.ValidUntil.Time, LeafDER: r.LeafDer,
	}
	if agent {
		s.ServerTrust = &ServerTrustFacts{Selector: AgentTrustSelector{Purpose: r.TrustPurpose, ListenerKind: r.TrustListenerKind, TrustDomain: r.TrustDomain},
			AuthorityEpoch: uint64(r.TrustAuthorityEpoch), AuthorityOperationID: r.TrustOperationID,
			ServerCABundleHighWater: contracts.VersionedDigest{Version: uint64(r.TrustBundleVersion), AuthoritySequence: uint64(r.TrustAuthoritySequence), Digest: contracts.Digest(r.TrustBundleDigest)}}
	}
	// Constructor rechecks exact identity/domain/status/material and owns copies.
	return newCertificateFacts(lookup, s)
}
