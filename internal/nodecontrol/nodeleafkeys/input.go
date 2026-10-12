package nodeleafkeys

import (
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
)

const keyIdentifierProfile = "node-key-id.spki-sha256.v1"

type Input struct {
	Leaf               nodeleaf.Input
	DeclaredIssuerSPKI []byte
}
type KeyIdentifier [32]byte
type Proof uint8

const (
	None Proof = iota
	NodeLeafDeclaredKeyIdentifiersBound
)

type Summary struct {
	Proof                                        Proof
	Leaf                                         nodeleaf.Summary
	SubjectKeyIdentifier, AuthorityKeyIdentifier KeyIdentifier
	DeclaredIssuerSPKIDERDigest                  contracts.Digest
}
