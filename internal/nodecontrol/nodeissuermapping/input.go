package nodeissuermapping

import (
	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/nodeissuerbinding"
	"time"
)

type Proof uint8

const (
	None Proof = iota
	DeclaredIssuerSerialCorrespondenceConditionalTimeBound
)

type RegistrationState uint8

const (
	UnknownRegistrationState RegistrationState = iota
	DeclaredEligible
	DeclaredWithdrawn
)

type Obligation uint8

const (
	UnknownObligation Obligation = iota
	AuthenticatedRegistryPublicationAndRollbackProtection
	AuthenticatedIssuanceLinkAndRowStatus
	GuardedV7SnapshotAndLease
	CrossRenewalDatabaseUniqueness
	FullIdentityCompositionCommitAndRuntimeAuthorization
)
const declaredMappingProfile = "declared-issuer-serial.registry-version-link.v1"
const materialReferenceProfile = "main-selected-issuer-serial.ca-der-bound.v1"
const maxRegistrationClaims = 128
const maxDeclaredRawBytes = 16568
const maxDeclaredVersion uint64 = 9223372036854775807

type SnapshotHead struct {
	Deployment contracts.Digest
	Version    uint64
	Digest     contracts.Digest
}
type RegistrationClaim struct {
	IssuerID    string
	CAVersion   uint64
	CADERDigest contracts.Digest
	State       RegistrationState
}
type CatalogClaim struct {
	Head    SnapshotHead
	Entries []RegistrationClaim
}
type AssociationClaim struct {
	Head                      SnapshotHead
	CertificateID, IssuanceID uuid.UUID
	NodeID, IssuerID          string
	CAVersion                 uint64
	LeafDER, PublicKey        contracts.Digest
	SerialBytes               []byte
}
type Claims struct {
	ExpectedHead SnapshotHead
	Catalog      CatalogClaim
	Association  AssociationClaim
}
type Input struct {
	Material nodeissuerbinding.Input
	Claims   Claims
}
type Projection struct {
	Head                      SnapshotHead
	IssuerID                  string
	CAVersion                 uint64
	CertificateID, IssuanceID uuid.UUID
	Local                     nodeissuerbinding.Projection
}
type Summary struct {
	Proof                  Proof
	Profile                string
	Evidence               nodeissuerbinding.Summary
	Proposed               Projection
	ObservedAt, ValidUntil time.Time
	Obligations            [5]Obligation
}
type Model struct{ owned *owner }
