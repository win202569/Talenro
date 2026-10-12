package statetime

import (
	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"time"
)

type EvidenceSummary struct {
	AttestationPayload, AttestationEnvelope, MetadataPayload, MetadataEnvelope, KeyID, RequestID contracts.Digest
	NodeID                                                                                       uuid.UUID
	Epoch                                                                                        uint64
	Nonce                                                                                        [32]byte
	OldCheckpoint, ProposedCheckpoint                                                            uint64
	AdvanceRequired                                                                              bool
	ValidUntil                                                                                   time.Time
}

// EvidenceModel records conditional assumptions, not provenance or authorization.
// Do not copy owners or use methods concurrently with Clear.
type EvidenceModel struct {
	ready      bool
	summary    EvidenceSummary
	metadata   CurrentMetadataObservation
	clock      TimeObservation
	request    RequestObservation
	checkpoint CheckpointObservation
}

func (e *EvidenceModel) Summary() EvidenceSummary {
	if e == nil || !e.ready {
		return EvidenceSummary{}
	}
	return e.summary
}
func (e *EvidenceModel) Clear() {
	if e != nil {
		*e = EvidenceModel{}
	}
}
