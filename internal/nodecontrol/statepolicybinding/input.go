package statepolicybinding

import (
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/statebinding"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
	"time"
)

type Input struct {
	Binding statebinding.Input
	Policy  statepolicy.ModelInput
}
type ResultStatus uint8

const (
	InvalidResult ResultStatus = iota
	ConditionalSelectedPolicyBound
)

type Obligation uint8

const (
	UnknownObligation Obligation = iota
	ExplicitObservationsAndBaseProvenance
	TimeAttestationAndFreshness
	OtherDomainPolicyCoverage
)

type SourceKind uint8

const (
	UnknownSource SourceKind = iota
	InitialAnchor
	RootRotationProof
)

type SourceIdentity struct {
	Kind              SourceKind
	Payload, Envelope contracts.Digest
}
type RootProjection struct {
	RootSet           statematerials.MaterialRef
	Version, Sequence uint64
	Source            SourceIdentity
}
type PolicyBase struct {
	StateID, CounterID                          contracts.Digest
	StateEpoch                                  uint64
	DeploymentSourceID, DeploymentObservationID contracts.Digest
}
type PolicyProjection struct {
	Root                   RootProjection
	Metadata, Desired      statepolicy.HighWater
	KeyCount, RevokedCount int
}
type TimeSnapshot struct {
	ProviderID, ObservationID    contracts.Digest
	Now, OldFloor, ObservedFloor time.Time
	MainFloor                    statecontract.OptionalTime
}
type Summary struct {
	Status                 ResultStatus
	Binding                statebinding.Summary
	Base                   PolicyBase
	Previous, Proposed     PolicyProjection
	HistoryCount           int
	Time                   TimeSnapshot
	ValidUntil             time.Time
	SelectedPolicyNoChange bool
	Obligations            [3]Obligation
}
