package nodeleaf

import (
	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"time"
)

type Input struct {
	DER          []byte
	NodeID       uuid.UUID
	TrustDomain  string
	ExpectedSPKI []byte
}
type Proof uint8

const (
	None Proof = iota
	NodeLeafMaterialShapeValid
)

type Obligation uint8

const (
	UnknownObligation Obligation = iota
	LeafSignatureAndIssuerTrust
	KeyIdentifierDerivationAndIssuerBinding
	TrustedCurrentTime
	IdentityAndAuthorityProvenance
	PrivateKeyPossession
	MainStateAndRuntimeBinding
)

type Summary struct {
	Proof                        Proof
	Pending                      [6]Obligation
	NodeID                       uuid.UUID
	TrustDomain                  string
	LeafDERDigest, SPKIDERDigest contracts.Digest
	Serial                       [20]byte
	SerialLength                 uint8
	NotBefore, NotAfter          time.Time
}
