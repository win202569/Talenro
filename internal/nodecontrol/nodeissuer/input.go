package nodeissuer

import (
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/nodeca"
	"talenro.local/platform/internal/nodecontrol/nodeleaf"
	"talenro.local/platform/internal/nodecontrol/nodeleafkeys"
)

const issuerRelationProfile = "node-leaf-direct-issuer.p256-sha256.v1"

type Input struct {
	Leaf nodeleaf.Input
	CA   nodeca.Input
}
type Proof uint8

const (
	None Proof = iota
	NodeLeafDirectIssuerSignatureBound
)

type Summary struct {
	Proof                                 Proof
	Leaf                                  nodeleafkeys.Summary
	CA                                    nodeca.Summary
	LeafTBSDERDigest, LeafIssuerDERDigest contracts.Digest
}
