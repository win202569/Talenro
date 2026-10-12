package statetimebinding

import (
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/statepolicybinding"
	"talenro.local/platform/internal/nodecontrol/statetime"
	"time"
)

type Input struct {
	Policy statepolicybinding.Input
	Time   statetime.ModelInput
}
type ResultStatus uint8

const (
	InvalidResult ResultStatus = iota
	ConditionalTimeEvidenceBound
)

type Obligation uint8

const (
	UnknownObligation Obligation = iota
	ObservationAndBaseProvenance
	TrustedTimeAndTLS
	AtomicRequestConsumption
	DurableCheckpointCommit
	FullMainStateAndRuntimeAuthorization
)

type ObservationBinding struct {
	BaseStateID, MetadataSourceID, MetadataObservationID              contracts.Digest
	RequestObservationID, CheckpointSourceID, CheckpointObservationID contracts.Digest
}
type Summary struct {
	Status       ResultStatus
	Policy       statepolicybinding.Summary
	Time         statetime.EvidenceSummary
	Observations ObservationBinding
	ValidUntil   time.Time
	Obligations  [5]Obligation
}
