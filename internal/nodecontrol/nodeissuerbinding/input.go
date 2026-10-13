package nodeissuerbinding

import (
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/nodeleafbinding"
	"time"
)

type Input = nodeleafbinding.Input
type Proof uint8

const (
	None Proof = iota
	SelectedIssuerSerialReferencesConditionalTimeBound
)
const selectedIssuerSerialProfile = "main-selected-issuer-serial.ca-der-bound.v1"

type Obligation uint8

const (
	UnknownObligation Obligation = iota
	StoreIssuerMappingAndDeploymentProvenance
	KeyHandleAndPrivateKeyPossession
	IdentityEpochLineageHistoryAndRowStatus
	MainStateAndObservationProvenance
	FullIdentityCompositionCommitAndRuntimeAuthorization
)

type Projection struct {
	Leaf     nodeleafbinding.Projection
	IssuerID contracts.Digest
	Serial   contracts.Digest
}
type Summary struct {
	Proof       Proof
	Profile     string
	Evidence    nodeleafbinding.Summary
	Proposed    Projection
	ObservedAt  time.Time
	ValidUntil  time.Time
	Obligations [5]Obligation
}
type Model struct{ owned *owner }
