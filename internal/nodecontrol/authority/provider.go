package authority

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gowebpki/jcs"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

const (
	reservationTranscriptDomain = "TALENRO-CONTROL-PLANE-AUTHORITY-RESERVATION-V1\x00"
	receiptTranscriptDomain     = "TALENRO-CONTROL-PLANE-AUTHORITY-RECEIPT-V1\x00"
	globalNodeScopeTranscript   = "TALENRO-GLOBAL-NODE-TRUST-SCOPE-V1\x00global_node_trust"
	globalOperatorTranscript    = "TALENRO-GLOBAL-OPERATOR-TRUST-SCOPE-V1\x00global_operator_trust"
)

var (
	ErrInvalidArgument  = errors.New("nodecontrol authority: invalid argument")
	ErrConflict         = errors.New("nodecontrol authority: conflicting retry")
	ErrTerminalConflict = errors.New("nodecontrol authority: terminal conflict")
	ErrNotFound         = errors.New("nodecontrol authority: operation not found")
	ErrCanceled         = errors.New("nodecontrol authority: context canceled")
	ErrInjectedFailure  = errors.New("nodecontrol authority: injected failure")
	ErrResponseLost     = errors.New("nodecontrol authority: response lost after mutation")

	globalNodeScopeDigest     = sha256.Sum256([]byte(globalNodeScopeTranscript))
	globalOperatorScopeDigest = sha256.Sum256([]byte(globalOperatorTranscript))
)

type Provider interface {
	Reserve(context.Context, ReserveRequest) (Reservation, error)
	Finalize(context.Context, FinalizeRequest) (Receipt, error)
	Abort(context.Context, AbortRequest) (Receipt, error)
	Inspect(context.Context, uuid.UUID) (Record, error)
	Head(context.Context) (Head, error)
	CommittedNodeCheckpoint(context.Context, contracts.Digest) (NodeCheckpoint, error)
}

type EffectKind string

const (
	EffectTrustBundlePublish       EffectKind = "trust_bundle_publish"
	EffectRootPublish              EffectKind = "root_publish"
	EffectMetadataPublish          EffectKind = "metadata_publish"
	EffectGrantCreate              EffectKind = "grant_create"
	EffectGrantClaim               EffectKind = "grant_claim"
	EffectCertificateActivate      EffectKind = "certificate_activate"
	EffectCertificateRevoke        EffectKind = "certificate_revoke"
	EffectIdentityEpochAdvance     EffectKind = "identity_epoch_advance"
	EffectSecurityIncidentOpen     EffectKind = "security_incident_open"
	EffectSecurityIncidentResolve  EffectKind = "security_incident_resolve"
	EffectResourceEnvelopeActivate EffectKind = "resource_envelope_activate"
	EffectDesiredActivate          EffectKind = "desired_activate"
	EffectRecoveryActivate         EffectKind = "recovery_activate"
	EffectOperatorTransition       EffectKind = "operator_transition"
	EffectOperatorAuthorizerChange EffectKind = "operator_authorizer_change"
)

type ScopeKind string

const (
	ScopeNode                ScopeKind = "node"
	ScopeGlobalNodeTrust     ScopeKind = "global_node_trust"
	ScopeGlobalOperatorTrust ScopeKind = "global_operator_trust"
)

type WALPosition string

type AbortReason string

const (
	AbortValidationFailed          AbortReason = "validation_failed"
	AbortSuperseded                AbortReason = "superseded"
	AbortProviderDependencyFailed  AbortReason = "provider_dependency_failed"
	AbortActivationDeadlineExpired AbortReason = "activation_deadline_expired"
)

type ReceiptStatus string

const (
	StatusCommitted ReceiptStatus = "committed"
	StatusAborted   ReceiptStatus = "aborted"
)

type Operation string

const (
	OperationReserve  Operation = "reserve"
	OperationFinalize Operation = "finalize"
	OperationAbort    Operation = "abort"
	OperationInspect  Operation = "inspect"
	OperationHead     Operation = "head"
)

type Failure string

const (
	FailureBeforeMutation            Failure = "before_mutation"
	FailureAfterMutationResponseLost Failure = "after_mutation_response_lost"
)

type ReserveRequest struct {
	OperationID uuid.UUID
	Kind        EffectKind
	ScopeKind   ScopeKind
	ScopeDigest contracts.Digest
}

type FinalizeRequest struct {
	OperationID  uuid.UUID
	EffectDigest contracts.Digest
	DBSystemID   uint64
	DBTimeline   uint32
	RequiredLSN  WALPosition
}

type AbortRequest struct {
	OperationID uuid.UUID
	Reason      AbortReason
}

type Reservation struct {
	OperationID       uuid.UUID
	Kind              EffectKind
	ScopeKind         ScopeKind
	ScopeDigest       contracts.Digest
	Epoch             uint64
	Sequence          uint64
	ReservationDigest contracts.Digest
}

type DatabasePoint struct {
	SystemID    uint64
	Timeline    uint32
	RequiredLSN WALPosition
}

type Receipt struct {
	Reservation
	EffectDigest  *contracts.Digest
	DatabasePoint *DatabasePoint
	Status        ReceiptStatus
	AbortReason   *AbortReason
	ReceiptDigest contracts.Digest
}

type Record struct {
	Reservation
	BoundEffectDigest  *contracts.Digest
	BoundDatabasePoint *DatabasePoint
	TerminalReceipt    *Receipt
}

type Head struct {
	Epoch                        uint64
	LatestReservedSequence       uint64
	LatestCommittedSequence      uint64
	LatestReservationDigest      contracts.Digest
	LatestCommittedOperationID   uuid.UUID
	LatestCommittedReceiptDigest contracts.Digest
	LatestCommittedDatabasePoint *DatabasePoint
}

type NodeCheckpoint struct {
	AuthorityEpoch uint64
	Sequence       uint64
	ReceiptDigest  contracts.Digest
}

func (value EffectKind) Validate() error {
	switch value {
	case EffectTrustBundlePublish, EffectRootPublish, EffectMetadataPublish, EffectGrantCreate,
		EffectGrantClaim, EffectCertificateActivate, EffectCertificateRevoke, EffectIdentityEpochAdvance,
		EffectSecurityIncidentOpen, EffectSecurityIncidentResolve, EffectResourceEnvelopeActivate,
		EffectDesiredActivate, EffectRecoveryActivate, EffectOperatorTransition, EffectOperatorAuthorizerChange:
		return nil
	default:
		return ErrInvalidArgument
	}
}

func (value ScopeKind) Validate() error {
	switch value {
	case ScopeNode, ScopeGlobalNodeTrust, ScopeGlobalOperatorTrust:
		return nil
	default:
		return ErrInvalidArgument
	}
}

func (value AbortReason) Validate() error {
	switch value {
	case AbortValidationFailed, AbortSuperseded, AbortProviderDependencyFailed, AbortActivationDeadlineExpired:
		return nil
	default:
		return ErrInvalidArgument
	}
}

func (value ReceiptStatus) Validate() error {
	switch value {
	case StatusCommitted, StatusAborted:
		return nil
	default:
		return ErrInvalidArgument
	}
}

func (value Operation) Validate() error {
	switch value {
	case OperationReserve, OperationFinalize, OperationAbort, OperationInspect, OperationHead:
		return nil
	default:
		return ErrInvalidArgument
	}
}

func (value Failure) Validate() error {
	switch value {
	case FailureBeforeMutation, FailureAfterMutationResponseLost:
		return nil
	default:
		return ErrInvalidArgument
	}
}

func (request ReserveRequest) Validate() error {
	if !validOperationID(request.OperationID) || request.Kind.Validate() != nil || request.ScopeKind.Validate() != nil ||
		request.ScopeDigest == (contracts.Digest{}) || !validEffectScope(request.Kind, request.ScopeKind) ||
		!validScopeDigest(request.ScopeKind, request.ScopeDigest) {
		return ErrInvalidArgument
	}
	return nil
}

func (request FinalizeRequest) Validate() error {
	if !validOperationID(request.OperationID) || request.EffectDigest == (contracts.Digest{}) ||
		request.DBSystemID == 0 || request.DBTimeline == 0 {
		return ErrInvalidArgument
	}
	if _, err := canonicalWALPosition(request.RequiredLSN); err != nil {
		return ErrInvalidArgument
	}
	return nil
}

func (request AbortRequest) Validate() error {
	if !validOperationID(request.OperationID) || request.Reason.Validate() != nil {
		return ErrInvalidArgument
	}
	return nil
}

func (value Reservation) Validate() error {
	request := ReserveRequest{
		OperationID: value.OperationID,
		Kind:        value.Kind,
		ScopeKind:   value.ScopeKind,
		ScopeDigest: value.ScopeDigest,
	}
	if request.Validate() != nil || !validPositiveCoordinate(value.Epoch) || !validPositiveCoordinate(value.Sequence) {
		return ErrInvalidArgument
	}
	digest, err := reservationDigest(value)
	if err != nil || digest != value.ReservationDigest {
		return ErrInvalidArgument
	}
	return nil
}

func (value DatabasePoint) Validate() error {
	if value.SystemID == 0 || value.Timeline == 0 {
		return ErrInvalidArgument
	}
	canonical, err := canonicalWALPosition(value.RequiredLSN)
	if err != nil || canonical != value.RequiredLSN {
		return ErrInvalidArgument
	}
	return nil
}

func (value Receipt) Validate() error {
	if value.Reservation.Validate() != nil || value.Status.Validate() != nil || value.ReceiptDigest == (contracts.Digest{}) {
		return ErrInvalidArgument
	}
	if (value.EffectDigest == nil) != (value.DatabasePoint == nil) {
		return ErrInvalidArgument
	}
	if value.EffectDigest != nil && (*value.EffectDigest == (contracts.Digest{}) || value.DatabasePoint.Validate() != nil) {
		return ErrInvalidArgument
	}
	switch value.Status {
	case StatusCommitted:
		if value.EffectDigest == nil || value.DatabasePoint == nil || value.AbortReason != nil {
			return ErrInvalidArgument
		}
	case StatusAborted:
		if value.AbortReason == nil || value.AbortReason.Validate() != nil {
			return ErrInvalidArgument
		}
	}
	digest, err := receiptDigest(value)
	if err != nil || digest != value.ReceiptDigest {
		return ErrInvalidArgument
	}
	return nil
}

func (value Record) Validate() error {
	if value.Reservation.Validate() != nil || (value.BoundEffectDigest == nil) != (value.BoundDatabasePoint == nil) {
		return ErrInvalidArgument
	}
	if value.BoundEffectDigest != nil && (*value.BoundEffectDigest == (contracts.Digest{}) || value.BoundDatabasePoint.Validate() != nil) {
		return ErrInvalidArgument
	}
	if value.TerminalReceipt == nil {
		return nil
	}
	if value.TerminalReceipt.Validate() != nil || value.TerminalReceipt.Reservation != value.Reservation ||
		!equalOptionalDigest(value.BoundEffectDigest, value.TerminalReceipt.EffectDigest) ||
		!equalOptionalDatabasePoint(value.BoundDatabasePoint, value.TerminalReceipt.DatabasePoint) {
		return ErrInvalidArgument
	}
	return nil
}

func (value Head) Validate() error {
	if !validPositiveCoordinate(value.Epoch) || value.LatestReservedSequence > math.MaxInt64 ||
		value.LatestCommittedSequence > math.MaxInt64 || value.LatestCommittedSequence > value.LatestReservedSequence ||
		(value.LatestReservedSequence == 0) != (value.LatestReservationDigest == (contracts.Digest{})) {
		return ErrInvalidArgument
	}
	if value.LatestCommittedSequence == 0 {
		if value.LatestCommittedOperationID != uuid.Nil || value.LatestCommittedReceiptDigest != (contracts.Digest{}) ||
			value.LatestCommittedDatabasePoint != nil {
			return ErrInvalidArgument
		}
		return nil
	}
	if !validOperationID(value.LatestCommittedOperationID) || value.LatestCommittedReceiptDigest == (contracts.Digest{}) ||
		value.LatestCommittedDatabasePoint == nil || value.LatestCommittedDatabasePoint.Validate() != nil {
		return ErrInvalidArgument
	}
	return nil
}

func (value NodeCheckpoint) Validate() error {
	if !validPositiveCoordinate(value.AuthorityEpoch) || value.Sequence > math.MaxInt64 ||
		(value.Sequence == 0) != (value.ReceiptDigest == (contracts.Digest{})) {
		return ErrInvalidArgument
	}
	return nil
}

func validEffectScope(kind EffectKind, scope ScopeKind) bool {
	switch kind {
	case EffectTrustBundlePublish:
		return scope == ScopeGlobalNodeTrust || scope == ScopeGlobalOperatorTrust
	case EffectRootPublish, EffectMetadataPublish:
		return scope == ScopeGlobalNodeTrust
	case EffectOperatorAuthorizerChange:
		return scope == ScopeGlobalOperatorTrust
	case EffectGrantCreate, EffectGrantClaim, EffectCertificateActivate, EffectCertificateRevoke,
		EffectIdentityEpochAdvance, EffectSecurityIncidentOpen, EffectSecurityIncidentResolve,
		EffectResourceEnvelopeActivate, EffectDesiredActivate, EffectRecoveryActivate, EffectOperatorTransition:
		return scope == ScopeNode
	default:
		return false
	}
}

func validScopeDigest(scope ScopeKind, digest contracts.Digest) bool {
	switch scope {
	case ScopeNode:
		return digest != (contracts.Digest{}) && digest != globalNodeScopeDigest && digest != globalOperatorScopeDigest
	case ScopeGlobalNodeTrust:
		return digest == globalNodeScopeDigest
	case ScopeGlobalOperatorTrust:
		return digest == globalOperatorScopeDigest
	default:
		return false
	}
}

func validOperationID(value uuid.UUID) bool {
	return value != uuid.Nil && value.Variant() == uuid.RFC4122
}

func validPositiveCoordinate(value uint64) bool {
	return value > 0 && value <= math.MaxInt64
}

func canonicalWALPosition(value WALPosition) (WALPosition, error) {
	text := string(value)
	if strings.Count(text, "/") != 1 {
		return "", ErrInvalidArgument
	}
	parts := strings.SplitN(text, "/", 2)
	if !uppercaseHex(parts[0]) || !uppercaseHex(parts[1]) {
		return "", ErrInvalidArgument
	}
	high, highErr := strconv.ParseUint(parts[0], 16, 32)
	low, lowErr := strconv.ParseUint(parts[1], 16, 32)
	if highErr != nil || lowErr != nil {
		return "", ErrInvalidArgument
	}
	return WALPosition(strings.ToUpper(strconv.FormatUint(high, 16)) + "/" + strings.ToUpper(strconv.FormatUint(low, 16))), nil
}

func uppercaseHex(value string) bool {
	if value == "" {
		return false
	}
	for index := 0; index < len(value); index++ {
		if (value[index] < '0' || value[index] > '9') && (value[index] < 'A' || value[index] > 'F') {
			return false
		}
	}
	return true
}

type authorityReservationPayloadV1 struct {
	AuthorityEpoch    string     `json:"authority_epoch"`
	AuthoritySequence string     `json:"authority_sequence"`
	EffectKind        EffectKind `json:"effect_kind"`
	OperationID       string     `json:"operation_id"`
	ScopeDigest       string     `json:"scope_digest"`
	ScopeKind         ScopeKind  `json:"scope_kind"`
}

type authorityDatabasePointPayloadV1 struct {
	DBSystemID  string      `json:"db_system_id"`
	DBTimeline  string      `json:"db_timeline"`
	RequiredLSN WALPosition `json:"required_lsn"`
}

type authorityReceiptPayloadV1 struct {
	AbortReason       *AbortReason                     `json:"abort_reason"`
	AuthorityEpoch    string                           `json:"authority_epoch"`
	AuthoritySequence string                           `json:"authority_sequence"`
	DatabasePoint     *authorityDatabasePointPayloadV1 `json:"database_point"`
	EffectDigest      *string                          `json:"effect_digest"`
	EffectKind        EffectKind                       `json:"effect_kind"`
	OperationID       string                           `json:"operation_id"`
	ReservationDigest string                           `json:"reservation_digest"`
	ScopeDigest       string                           `json:"scope_digest"`
	ScopeKind         ScopeKind                        `json:"scope_kind"`
	Status            ReceiptStatus                    `json:"status"`
}

func reservationDigest(value Reservation) (contracts.Digest, error) {
	payload := authorityReservationPayloadV1{
		AuthorityEpoch:    strconv.FormatUint(value.Epoch, 10),
		AuthoritySequence: strconv.FormatUint(value.Sequence, 10),
		EffectKind:        value.Kind,
		OperationID:       value.OperationID.String(),
		ScopeDigest:       hex.EncodeToString(value.ScopeDigest[:]),
		ScopeKind:         value.ScopeKind,
	}
	return canonicalTranscriptDigest(reservationTranscriptDomain, payload)
}

func receiptDigest(value Receipt) (contracts.Digest, error) {
	payload := authorityReceiptPayloadV1{
		AbortReason:       value.AbortReason,
		AuthorityEpoch:    strconv.FormatUint(value.Epoch, 10),
		AuthoritySequence: strconv.FormatUint(value.Sequence, 10),
		EffectKind:        value.Kind,
		OperationID:       value.OperationID.String(),
		ReservationDigest: hex.EncodeToString(value.ReservationDigest[:]),
		ScopeDigest:       hex.EncodeToString(value.ScopeDigest[:]),
		ScopeKind:         value.ScopeKind,
		Status:            value.Status,
	}
	if value.EffectDigest != nil {
		digest := hex.EncodeToString(value.EffectDigest[:])
		payload.EffectDigest = &digest
	}
	if value.DatabasePoint != nil {
		payload.DatabasePoint = &authorityDatabasePointPayloadV1{
			DBSystemID:  strconv.FormatUint(value.DatabasePoint.SystemID, 10),
			DBTimeline:  strconv.FormatUint(uint64(value.DatabasePoint.Timeline), 10),
			RequiredLSN: value.DatabasePoint.RequiredLSN,
		}
	}
	return canonicalTranscriptDigest(receiptTranscriptDomain, payload)
}

func canonicalTranscriptDigest(domain string, payload any) (contracts.Digest, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return contracts.Digest{}, ErrInvalidArgument
	}
	canonical, err := jcs.Transform(body)
	if err != nil {
		return contracts.Digest{}, ErrInvalidArgument
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte(domain))
	_, _ = hash.Write(canonical)
	var digest contracts.Digest
	copy(digest[:], hash.Sum(nil))
	return digest, nil
}

func equalOptionalDigest(left, right *contracts.Digest) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func equalOptionalDatabasePoint(left, right *DatabasePoint) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
