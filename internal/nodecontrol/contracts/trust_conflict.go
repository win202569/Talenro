package contracts

import (
	"crypto/sha256"
	"encoding/binary"
	"math"

	"github.com/google/uuid"
)

const (
	trustConflictLocalFaultDomainV1 = "TALENRO-TRUST-CONFLICT-LOCAL-FAULT-ID-V1\x00"
	// UUIDv5(DNS, "talenro.local/nodecontrol/trust-conflict-local-fault-id/v1").
	trustConflictLocalFaultNamespaceV1 = "a9198a7e-0943-524d-b20d-1553bc9a036f"
)

type TrustConflictLocalFaultBindingV1 struct {
	NodeID            uuid.UUID
	IncidentID        uuid.UUID
	PollRequestDigest Digest
	CertificateDigest Digest
	IdentityEpoch     uint64
}

func DeriveTrustConflictLocalFaultIDV1(binding TrustConflictLocalFaultBindingV1) (uuid.UUID, error) {
	if !validTrustConflictUUID(binding.NodeID) || !validTrustConflictUUID(binding.IncidentID) ||
		binding.PollRequestDigest == (Digest{}) || binding.CertificateDigest == (Digest{}) ||
		binding.IdentityEpoch == 0 || binding.IdentityEpoch > math.MaxInt64 {
		return uuid.Nil, ErrInvalidAuthorityValue
	}

	var transcript [len(trustConflictLocalFaultDomainV1) + 16 + 16 + 32 + 32 + 8]byte
	offset := copy(transcript[:], trustConflictLocalFaultDomainV1)
	offset += copy(transcript[offset:], binding.NodeID[:])
	offset += copy(transcript[offset:], binding.IncidentID[:])
	offset += copy(transcript[offset:], binding.PollRequestDigest[:])
	offset += copy(transcript[offset:], binding.CertificateDigest[:])
	binary.BigEndian.PutUint64(transcript[offset:], binding.IdentityEpoch)

	transcriptDigest := sha256.Sum256(transcript[:])
	return uuid.NewSHA1(uuid.MustParse(trustConflictLocalFaultNamespaceV1), transcriptDigest[:]), nil
}

func validTrustConflictUUID(value uuid.UUID) bool {
	if value == uuid.Nil || value.Variant() != uuid.RFC4122 {
		return false
	}
	version := value.Version()
	return version >= 1 && version <= 5
}
