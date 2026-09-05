package authority

import (
	"context"
	"time"

	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/store"
)

// Repository persists the database-visible mirror of provider authority fences.
// Transactional methods use the caller-owned DBTX exactly as supplied.
type Repository interface {
	RecordPending(context.Context, store.DBTX, Reservation, time.Time) error
	CaptureDatabasePoint(context.Context) (DatabasePoint, error)
	BindEffect(context.Context, store.DBTX, uuid.UUID, contracts.Digest, DatabasePoint, time.Time) error
	ActivateCommitted(context.Context, store.DBTX, Receipt, time.Time) error
	RecordAborted(context.Context, store.DBTX, Receipt, time.Time) error
	Get(context.Context, uuid.UUID) (Record, error)
	Head(context.Context) (DatabaseHead, error)
	ListPending(context.Context, uint64) ([]PendingFence, error)
	CommittedNodeCheckpoint(context.Context, uint64, contracts.Digest) (NodeCheckpoint, error)
}

type DatabaseHead struct {
	Epoch                        uint64
	RecordCount                  uint64
	LatestReservedSequence       uint64
	LatestCommittedSequence      uint64
	PendingCount                 uint64
	HasSequenceGap               bool
	LatestReservationDigest      contracts.Digest
	LatestCommittedOperationID   uuid.UUID
	LatestCommittedReceiptDigest contracts.Digest
	LatestCommittedDatabasePoint *DatabasePoint
}

type PendingFence struct {
	OperationID        uuid.UUID
	Kind               EffectKind
	ScopeKind          ScopeKind
	ScopeDigest        contracts.Digest
	Epoch              uint64
	Sequence           uint64
	ReservationDigest  contracts.Digest
	BoundEffectDigest  *contracts.Digest
	BoundDatabasePoint *DatabasePoint
	ReservedAt         time.Time
	EffectBoundAt      *time.Time
	AbortClaim         *AbortClaim
}

type AbortClaim struct {
	Reason    AbortReason
	ClaimedAt time.Time
}

type PersistedAuthorityEffectOutcome struct {
	CommitmentJCS                  []byte
	CommitmentDigest               contracts.Digest
	Receipt                        Receipt
	ProviderHeadJCS                []byte
	ProviderHeadDigest             contracts.Digest
	CheckpointAnchorJCS            []byte
	CheckpointAnchorDigest         contracts.Digest
	Reason                         AuthorityEffectReason
	AttestationExpiresAt           time.Time
	ActivationDeadline             time.Time
	ExpectedProviderIdentityDigest contracts.Digest
	EvidenceJCS                    []byte
	EvidenceDigest                 contracts.Digest
	ResolutionJCS                  []byte
	ResolutionDigest               contracts.Digest
}

type StoredFence struct {
	Record           Record
	AbortClaim       *AbortClaim
	PersistedOutcome *PersistedAuthorityEffectOutcome
}

type ClaimV1Reservation struct {
	Reservation          Reservation
	ProtocolActivationID uuid.UUID
}

type AuthorityV7Repository interface {
	Repository
	RecordPendingClaimV1(context.Context, store.DBTX, ClaimV1Reservation, time.Time) error
	ClaimAbort(context.Context, store.DBTX, uuid.UUID, AbortReason, time.Time) error
	GetStoredFence(context.Context, uuid.UUID) (StoredFence, error)
	Lock(context.Context, store.DBTX, uuid.UUID) (StoredFence, error)
}
