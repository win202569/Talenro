package serving

import (
	"bytes"
	"crypto/sha256"
	"math"
	"time"

	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

// ServerTrustFacts carries the exact agent trust group, not proof of its fence.
type ServerTrustFacts struct {
	Selector                AgentTrustSelector
	AuthorityEpoch          uint64
	AuthorityOperationID    uuid.UUID
	ServerCABundleHighWater contracts.VersionedDigest
}

// CertificateSnapshot is a detached certificate projection. Inventory epoch and
// lineage are retained separately so a mismatched join cannot be hidden here.
// This is not a complete B03 endpoint authorization or inventory-state snapshot.
type CertificateSnapshot struct {
	Identity                              CertificateLookupInput
	CertificateID, LineageID              uuid.UUID
	IdentityEpoch, InventoryIdentityEpoch uint64
	InventoryLineageID                    uuid.UUID
	InventoryVersion                      uint64
	AuthorityOperationID                  uuid.UUID
	Authority                             contracts.AuthorityVersion
	Status                                string
	ValidFrom, ValidUntil                 time.Time
	LeafDER                               []byte
	ServerTrust                           *ServerTrustFacts
}

// CertificateFacts owns its material. There is intentionally no public factory:
// pure validation cannot prove ready authority or committed-active fences.
// A future guarded reader must stage these facts until its post-check passes.
type CertificateFacts struct{ snapshot CertificateSnapshot }

// newCertificateFacts validates a private staged projection. It does not parse
// or authenticate X.509, evaluate trusted current time, verify fence provenance,
// or grant endpoint access. Those checks must not be inferred from success.
func newCertificateFacts(lookup CertificateLookup, s CertificateSnapshot) (CertificateFacts, error) {
	if lookup.Validate() != nil {
		return CertificateFacts{}, authority.ErrInvalidArgument
	}
	i := lookup.input
	if s.Identity.NodeID != i.NodeID || s.Identity.IssuerID != i.IssuerID ||
		!bytes.Equal(s.Identity.SerialBytes, i.SerialBytes) ||
		s.Identity.LeafDERDigest != i.LeafDERDigest || s.Identity.PublicKeyDigest != i.PublicKeyDigest ||
		!factsUUID(s.CertificateID) || !factsUUID(s.LineageID) ||
		!factsCoordinate(s.IdentityEpoch) || s.IdentityEpoch != s.InventoryIdentityEpoch ||
		s.LineageID != s.InventoryLineageID || !factsCoordinate(s.InventoryVersion) ||
		!factsUUID(s.AuthorityOperationID) || s.Authority.Validate() != nil ||
		s.ValidFrom.IsZero() || s.ValidUntil.IsZero() || !s.ValidUntil.After(s.ValidFrom) ||
		len(s.LeafDER) < 1 || len(s.LeafDER) > 65536 || sha256.Sum256(s.LeafDER) != s.Identity.LeafDERDigest {
		return CertificateFacts{}, authority.ErrConflict
	}
	switch s.Status {
	case "active", "recovery_pending", "recovery_limited":
	default:
		return CertificateFacts{}, authority.ErrConflict
	}
	selector, agent := lookup.AgentTrustSelector()
	if agent {
		t := s.ServerTrust
		if t == nil || t.Selector != selector || t.AuthorityEpoch != s.Authority.Epoch ||
			!factsUUID(t.AuthorityOperationID) || !factsCoordinate(t.ServerCABundleHighWater.Version) ||
			!factsCoordinate(t.ServerCABundleHighWater.AuthoritySequence) || t.ServerCABundleHighWater.Digest == (contracts.Digest{}) {
			return CertificateFacts{}, authority.ErrConflict
		}
	} else if s.ServerTrust != nil {
		return CertificateFacts{}, authority.ErrConflict
	}
	return CertificateFacts{snapshot: cloneCertificateSnapshot(s)}, nil
}

// Snapshot returns caller-owned serial, DER and trust material. A zero value
// returns an empty snapshot; neither value is itself authorization evidence.
func (f CertificateFacts) Snapshot() CertificateSnapshot { return cloneCertificateSnapshot(f.snapshot) }

func cloneCertificateSnapshot(s CertificateSnapshot) CertificateSnapshot {
	s.Identity.SerialBytes = bytes.Clone(s.Identity.SerialBytes)
	s.LeafDER = bytes.Clone(s.LeafDER)
	if s.ServerTrust != nil {
		trust := *s.ServerTrust
		s.ServerTrust = &trust
	}
	return s
}

func factsUUID(id uuid.UUID) bool   { return id != uuid.Nil && id.Variant() == uuid.RFC4122 }
func factsCoordinate(v uint64) bool { return v > 0 && v <= math.MaxInt64 }
