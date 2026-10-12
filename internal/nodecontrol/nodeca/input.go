package nodeca

import (
	"talenro.local/platform/internal/nodecontrol/contracts"
	"time"
)

type Input struct{ DER, DeclaredSPKI []byte }
type Proof uint8

const (
	None Proof = iota
	NodeCAMaterialShapeAndDeclaredKeyBound
)

type Obligation uint8

const (
	UnknownObligation Obligation = iota
	CertificateSignatureAndIssuerChain
	AuthorityIdentifierAndParentBinding
	TrustedCurrentTime
	DeploymentAuthorizationAndPurpose
	BundleStatusAndRollback
	LeafIssuerRelationAndRuntimeUse
)

type KeyIdentifier [32]byte
type Summary struct {
	Proof                                                         Proof
	Pending                                                       [6]Obligation
	CADERDigest, SPKIDERDigest, SubjectDERDigest, IssuerDERDigest contracts.Digest
	SubjectKeyIdentifier, AuthorityKeyIdentifier                  KeyIdentifier
	Serial                                                        [20]byte
	SerialLength                                                  uint8
	NotBefore, NotAfter                                           time.Time
	HasPathLen                                                    bool
	PathLen                                                       uint8
	CRLSignAllowed                                                bool
}
