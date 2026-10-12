package statepolicy

import (
	"context"
	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/stateverify"
	"time"
)

type ObservationStatus uint8

const (
	Unknown ObservationStatus = iota
	Available
	Unavailable
	Rollback
)

type DeploymentObservation struct {
	Status                                 ObservationStatus
	SourceID, ObservationID, AnchorPayload contracts.Digest
	NodeID                                 uuid.UUID
	Epoch                                  uint64
	FreshInstall                           bool
}
type TimeObservation struct {
	Status                    ObservationStatus
	ProviderID, ObservationID contracts.Digest
	Now, Floor                time.Time
}
type KeyHistoryObservation struct {
	Status                                             ObservationStatus
	ObservationID, BaseStateID, MetadataPayload, KeyID contracts.Digest
	RetiringKnown, DeadlinesKnown                      bool
	LastSnapshotUntil, LastCompositeDeadline           time.Time
	IncidentID                                         uuid.UUID
	IncidentKind                                       string
	IncidentOpen                                       bool
}
type ModelInput struct {
	Batch      stateverify.Batch
	Deployment DeploymentObservation
	Time       TimeObservation
	Old        *StateModel
	History    []KeyHistoryObservation
}

// These provider-neutral boundaries are intentionally not called by the core.
type DeploymentObserver interface {
	ObserveDeployment(context.Context) (DeploymentObservation, error)
}
type TimeObserver interface {
	ObserveTime(context.Context) (TimeObservation, error)
}
type StateObserver interface {
	ObserveState(context.Context) (*StateModel, error)
}
