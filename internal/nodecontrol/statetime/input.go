package statetime

import (
	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"time"
)

type ObservationStatus uint8

const (
	Unknown ObservationStatus = iota
	Available
	Unavailable
	Rollback
)

type CurrentMetadataObservation struct {
	Status                                              ObservationStatus
	SourceID, ObservationID, StateID, Payload, Envelope contracts.Digest
	NodeID                                              uuid.UUID
	Epoch, Version                                      uint64
}
type TimeObservation struct {
	Status                    ObservationStatus
	ProviderID, ObservationID contracts.Digest
	Now, Floor                time.Time
}
type RequestObservation struct {
	Status                   ObservationStatus
	RequestID, ObservationID contracts.Digest
	NodeID                   uuid.UUID
	Epoch                    uint64
	ExpectedNonce            [32]byte
}
type CheckpointObservation struct {
	Status                           ObservationStatus
	SourceID, ObservationID, StateID contracts.Digest
	NodeID                           uuid.UUID
	Epoch, Sequence                  uint64
}
type ModelInput struct {
	AttestationEnvelope, MetadataEnvelope []byte
	Metadata                              CurrentMetadataObservation
	Time                                  TimeObservation
	Request                               RequestObservation
	Checkpoint                            CheckpointObservation
}
