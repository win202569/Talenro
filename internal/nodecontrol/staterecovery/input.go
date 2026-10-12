package staterecovery

import "talenro.local/platform/internal/nodecontrol/contracts"

type Scope struct {
	NodeID                           string
	AuthorityEpoch                   uint64
	GuardID, SealKeyID, DeploymentID contracts.Digest
}
type CounterStatus uint8

const (
	CounterUnknown CounterStatus = iota
	CounterAvailable
	CounterUnavailable
	CounterRollback
)

type CounterObservation struct {
	Status                          CounterStatus
	SourceID, ObservationID, ScanID contracts.Digest
	Scope                           Scope
	Value, Maximum                  uint64
}
type SlotID uint8

const (
	SlotUnknown SlotID = iota
	SlotA
	SlotB
)

type SlotStatus uint8

const (
	SlotUnknownStatus SlotStatus = iota
	SlotAbsent
	SlotIntact
	SlotCorrupt
	SlotUnsupported
	SlotUnavailable
)

type SlotObservation struct {
	Status                          SlotStatus
	SourceID, ObservationID, ScanID contracts.Digest
	Scope                           Scope
	StateEpoch                      uint64
	ContentDigest, BlobDigest       contracts.Digest
	ModelVersion                    uint32
}
type RecoveryInput struct {
	ExpectedScope  Scope
	Counter        CounterObservation
	A, B           SlotObservation
	HasPointerHint bool
	PointerHint    int
}
type RecoverySummary struct {
	Scope                     Scope
	SelectedSlot              SlotID
	StateEpoch                uint64
	ContentDigest, BlobDigest contracts.Digest
	DiscardFuture             bool
	FutureSlot                SlotID
}

type ProposalObservation struct {
	ProposalID, AttemptID, BaseBlobDigest, ProposedContentDigest contracts.Digest
	ExpectedScope                                                Scope
	BaseStateEpoch                                               uint64
}
type CommitPhase uint8

const (
	PhaseUnknown CommitPhase = iota
	NoChange
	Prepared
	Written
	FileDurable
	CandidateDurable
	ReadyToAdvance
	NeedsRecovery
)

type CommitSummary struct {
	Phase                                                      CommitPhase
	Scope                                                      Scope
	ProposalID, AttemptID                                      contracts.Digest
	BaseSlot, TargetSlot                                       SlotID
	BaseStateEpoch, TargetStateEpoch                           uint64
	BaseBlobDigest, ProposedContentDigest, CandidateBlobDigest contracts.Digest
}
type EventKind uint8

const (
	EventUnknown EventKind = iota
	CandidateWritten
	FileSynced
	DirectorySynced
	BaseRechecked
	CounterAdvanced
	Failed
	OutcomeUnknown
)

type Event struct {
	Kind                         EventKind
	Phase                        CommitPhase
	AttemptID, EventID, SourceID contracts.Digest
	Scope                        Scope
	TargetSlot                   SlotID
	TargetStateEpoch             uint64
	ContentDigest, BlobDigest    contracts.Digest
	Scan                         RecoveryInput
	CounterBefore, CounterAfter  uint64
}
type ConfirmationInput struct {
	AttemptID contracts.Digest
	Scan      RecoveryInput
}
type ConfirmationOutcome uint8

const (
	ConfirmationUnknown ConfirmationOutcome = iota
	NotCommitted
	CommittedModel
)

type ConfirmationSummary struct {
	Outcome   ConfirmationOutcome
	AttemptID contracts.Digest
	Recovery  RecoverySummary
}
