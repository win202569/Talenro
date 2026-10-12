package statepolicy

import (
	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"time"
)

type HighWater struct {
	Version, Sequence uint64
	Payload           contracts.Digest
}
type RootRecord struct {
	Mark HighWater
	Root stateformat.RootMaterial
}
type KeyRecord struct {
	Material                                                 stateformat.MetadataKeyMaterial
	FirstRevokedAt, LastSnapshotUntil, LastCompositeDeadline time.Time
	DeadlinesKnown                                           bool
}

// StateModel owns its slices. Identity fields describe a supplied snapshot, not
// a hardware commit produced by this package. Do not copy an owner by value.
type StateModel struct {
	Status                                                                ObservationStatus
	StateID, CounterID, DeploymentSourceID, TimeProviderID, AnchorPayload contracts.Digest
	NodeID                                                                uuid.UUID
	Epoch, StateEpoch                                                     uint64
	TimeFloor                                                             time.Time
	Roots                                                                 []RootRecord
	Metadata, Desired                                                     HighWater
	Keys                                                                  []KeyRecord
	RevokedIDs                                                            []contracts.Digest
}

func (s *StateModel) Clear() {
	if s == nil {
		return
	}
	for i := range s.Roots {
		s.Roots[i].Root.Clear()
	}
	clear(s.Roots)
	clear(s.Keys)
	clear(s.RevokedIDs)
	*s = StateModel{}
}
