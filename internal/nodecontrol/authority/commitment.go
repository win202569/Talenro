package authority

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gowebpki/jcs"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/strictjson"
)

const (
	authorityCanonicalMaximumBytes = 4096
	authorityAbsentDigestText      = "0000000000000000000000000000000000000000000000000000000000000000"

	authorityEffectCommitmentDomain = "talenro.nodecontrol.authority-effect-commitment.v1\x00"
	authorityProviderHeadDomain     = "talenro.nodecontrol.authority-provider-head.v1\x00"
	authorityCheckpointAnchorDomain = "talenro.nodecontrol.authority-checkpoint-anchor.v1\x00"
	authorityEmptyInputsDomain      = "talenro.nodecontrol.authority-effect-activation-inputs.empty.v1\x00"
)

var (
	authorityEffectActivationInputsEmpty = contracts.Digest(sha256.Sum256([]byte(authorityEmptyInputsDomain)))
	// AuthorityEffectActivationInputsEmptyV1 is the frozen semantic digest for
	// the final-not-applied branch. Internal validation uses an immutable copy.
	AuthorityEffectActivationInputsEmptyV1 = authorityEffectActivationInputsEmpty
)

type CommitmentMode string

const (
	CommitmentConditionalApply CommitmentMode = "conditional_apply"
	CommitmentFinalNotApplied  CommitmentMode = "final_not_applied"
)

type AuthorityEffectReason string

const (
	EffectReasonNone                      AuthorityEffectReason = "none"
	EffectReasonFailed                    AuthorityEffectReason = "failed"
	EffectReasonSuperseded                AuthorityEffectReason = "superseded"
	EffectReasonActivationDeadlineExpired AuthorityEffectReason = "activation_deadline_expired"
	EffectReasonValidationRejected        AuthorityEffectReason = "validation_rejected"
)

func (value CommitmentMode) valid() bool {
	return value == CommitmentConditionalApply || value == CommitmentFinalNotApplied
}

func (value AuthorityEffectReason) valid() bool {
	switch value {
	case EffectReasonNone, EffectReasonFailed, EffectReasonSuperseded,
		EffectReasonActivationDeadlineExpired, EffectReasonValidationRejected:
		return true
	default:
		return false
	}
}

func (value AuthorityEffectReason) finalReason() bool {
	switch value {
	case EffectReasonFailed, EffectReasonSuperseded, EffectReasonActivationDeadlineExpired, EffectReasonValidationRejected:
		return true
	default:
		return false
	}
}

type AuthorityEffectCommitmentInput struct {
	OperationID             uuid.UUID
	Kind                    EffectKind
	ScopeKind               ScopeKind
	ScopeDigest             contracts.Digest
	Epoch                   uint64
	Sequence                uint64
	BaseEffectDigest        contracts.Digest
	Mode                    CommitmentMode
	Reason                  AuthorityEffectReason
	ActivationPolicyVersion uint64
	ActivationInputsDigest  contracts.Digest
}

type AuthorityEffectCommitmentFacts struct {
	OperationID             uuid.UUID
	Kind                    EffectKind
	ScopeKind               ScopeKind
	ScopeDigest             contracts.Digest
	Epoch                   uint64
	Sequence                uint64
	BaseEffectDigest        contracts.Digest
	Mode                    CommitmentMode
	Reason                  AuthorityEffectReason
	ActivationPolicyVersion uint64
	ActivationInputsDigest  contracts.Digest
}

type AuthorityEffectCommitment struct {
	facts     AuthorityEffectCommitmentFacts
	canonical []byte
	digest    contracts.Digest
}

type authorityEffectCommitmentV1 struct {
	ActivationInputsDigest  string                `json:"activation_inputs_digest"`
	ActivationPolicyVersion string                `json:"activation_policy_version"`
	AuthorityEpoch          string                `json:"authority_epoch"`
	AuthoritySequence       string                `json:"authority_sequence"`
	BaseEffectDigest        string                `json:"base_effect_digest"`
	CommitmentMode          CommitmentMode        `json:"commitment_mode"`
	EffectKind              EffectKind            `json:"effect_kind"`
	OperationID             string                `json:"operation_id"`
	ReasonCode              AuthorityEffectReason `json:"reason_code"`
	SchemaVersion           string                `json:"schema_version"`
	ScopeDigest             string                `json:"scope_digest"`
	ScopeKind               ScopeKind             `json:"scope_kind"`
}

func NewAuthorityEffectCommitment(input AuthorityEffectCommitmentInput) (AuthorityEffectCommitment, error) {
	if !validOperationID(input.OperationID) || input.Kind.Validate() != nil || input.ScopeKind.Validate() != nil ||
		input.Kind == EffectTrustBundlePublish || input.Kind == EffectOperatorAuthorizerChange ||
		!validEffectScope(input.Kind, input.ScopeKind) || !validScopeDigest(input.ScopeKind, input.ScopeDigest) ||
		!validPositiveCoordinate(input.Epoch) || !validPositiveCoordinate(input.Sequence) ||
		input.BaseEffectDigest == (contracts.Digest{}) || !input.Mode.valid() || !input.Reason.valid() {
		return AuthorityEffectCommitment{}, ErrInvalidArgument
	}
	activationInputs := input.ActivationInputsDigest
	policy := ""
	switch input.Mode {
	case CommitmentConditionalApply:
		if input.Reason != EffectReasonNone || !validPositiveCoordinate(input.ActivationPolicyVersion) ||
			activationInputs == (contracts.Digest{}) || activationInputs == authorityEffectActivationInputsEmpty {
			return AuthorityEffectCommitment{}, ErrInvalidArgument
		}
		policy = strconv.FormatUint(input.ActivationPolicyVersion, 10)
	case CommitmentFinalNotApplied:
		if !input.Reason.finalReason() || input.ActivationPolicyVersion != 0 || activationInputs != (contracts.Digest{}) {
			return AuthorityEffectCommitment{}, ErrInvalidArgument
		}
		policy = "none"
		activationInputs = authorityEffectActivationInputsEmpty
	}
	wire := authorityEffectCommitmentV1{
		ActivationInputsDigest: digestString(activationInputs), ActivationPolicyVersion: policy,
		AuthorityEpoch: strconv.FormatUint(input.Epoch, 10), AuthoritySequence: strconv.FormatUint(input.Sequence, 10),
		BaseEffectDigest: digestString(input.BaseEffectDigest), CommitmentMode: input.Mode, EffectKind: input.Kind,
		OperationID: input.OperationID.String(), ReasonCode: input.Reason, SchemaVersion: "1",
		ScopeDigest: digestString(input.ScopeDigest), ScopeKind: input.ScopeKind,
	}
	canonical, digest, err := canonicalAuthorityArtifact(authorityEffectCommitmentDomain, wire)
	if err != nil {
		return AuthorityEffectCommitment{}, ErrInvalidArgument
	}
	return AuthorityEffectCommitment{
		facts: AuthorityEffectCommitmentFacts{
			OperationID: input.OperationID, Kind: input.Kind, ScopeKind: input.ScopeKind, ScopeDigest: input.ScopeDigest,
			Epoch: input.Epoch, Sequence: input.Sequence, BaseEffectDigest: input.BaseEffectDigest, Mode: input.Mode,
			Reason: input.Reason, ActivationPolicyVersion: input.ActivationPolicyVersion, ActivationInputsDigest: activationInputs,
		},
		canonical: canonical, digest: digest,
	}, nil
}

func ParseAuthorityEffectCommitment(body []byte) (AuthorityEffectCommitment, error) {
	var wire authorityEffectCommitmentV1
	if err := decodeCanonicalAuthorityArtifact(body, &wire); err != nil || wire.SchemaVersion != "1" {
		return AuthorityEffectCommitment{}, ErrInvalidArgument
	}
	operationID, ok := parseCanonicalUUID(wire.OperationID)
	if !ok {
		return AuthorityEffectCommitment{}, ErrInvalidArgument
	}
	epoch, ok := parsePositiveDecimal(wire.AuthorityEpoch, math.MaxInt64)
	if !ok {
		return AuthorityEffectCommitment{}, ErrInvalidArgument
	}
	sequence, ok := parsePositiveDecimal(wire.AuthoritySequence, math.MaxInt64)
	if !ok {
		return AuthorityEffectCommitment{}, ErrInvalidArgument
	}
	scopeDigest, ok := parseDigestString(wire.ScopeDigest, false)
	if !ok {
		return AuthorityEffectCommitment{}, ErrInvalidArgument
	}
	baseDigest, ok := parseDigestString(wire.BaseEffectDigest, false)
	if !ok {
		return AuthorityEffectCommitment{}, ErrInvalidArgument
	}
	activationInputs, ok := parseDigestString(wire.ActivationInputsDigest, false)
	if !ok {
		return AuthorityEffectCommitment{}, ErrInvalidArgument
	}
	input := AuthorityEffectCommitmentInput{
		OperationID: operationID, Kind: wire.EffectKind, ScopeKind: wire.ScopeKind, ScopeDigest: scopeDigest,
		Epoch: epoch, Sequence: sequence, BaseEffectDigest: baseDigest, Mode: wire.CommitmentMode,
		Reason: wire.ReasonCode, ActivationInputsDigest: activationInputs,
	}
	if wire.CommitmentMode == CommitmentConditionalApply {
		policy, valid := parsePositiveDecimal(wire.ActivationPolicyVersion, math.MaxInt64)
		if !valid {
			return AuthorityEffectCommitment{}, ErrInvalidArgument
		}
		input.ActivationPolicyVersion = policy
	} else if wire.CommitmentMode == CommitmentFinalNotApplied {
		if wire.ActivationPolicyVersion != "none" || activationInputs != authorityEffectActivationInputsEmpty {
			return AuthorityEffectCommitment{}, ErrInvalidArgument
		}
		input.ActivationInputsDigest = contracts.Digest{}
	}
	value, err := NewAuthorityEffectCommitment(input)
	if err != nil || !bytes.Equal(value.canonical, body) {
		return AuthorityEffectCommitment{}, ErrInvalidArgument
	}
	return value, nil
}

func (value AuthorityEffectCommitment) Facts() AuthorityEffectCommitmentFacts { return value.facts }
func (value AuthorityEffectCommitment) CanonicalJCS() []byte                  { return bytes.Clone(value.canonical) }
func (value AuthorityEffectCommitment) Digest() contracts.Digest              { return value.digest }

func validAuthorityEffectCommitment(value AuthorityEffectCommitment) bool {
	if len(value.canonical) == 0 || value.digest == (contracts.Digest{}) {
		return false
	}
	input := AuthorityEffectCommitmentInput{
		OperationID: value.facts.OperationID, Kind: value.facts.Kind, ScopeKind: value.facts.ScopeKind,
		ScopeDigest: value.facts.ScopeDigest, Epoch: value.facts.Epoch, Sequence: value.facts.Sequence,
		BaseEffectDigest: value.facts.BaseEffectDigest, Mode: value.facts.Mode, Reason: value.facts.Reason,
		ActivationPolicyVersion: value.facts.ActivationPolicyVersion, ActivationInputsDigest: value.facts.ActivationInputsDigest,
	}
	if value.facts.Mode == CommitmentFinalNotApplied {
		input.ActivationInputsDigest = contracts.Digest{}
	}
	rebuilt, err := NewAuthorityEffectCommitment(input)
	return err == nil && rebuilt.digest == value.digest && bytes.Equal(rebuilt.canonical, value.canonical)
}

type AuthorityProviderHeadSnapshot struct {
	facts     Head
	canonical []byte
	digest    contracts.Digest
}

type authorityProviderHeadV1 struct {
	AuthorityEpoch               string      `json:"authority_epoch"`
	DBSystemID                   string      `json:"db_system_id"`
	DBTimeline                   string      `json:"db_timeline"`
	LatestCommittedOperationID   string      `json:"latest_committed_operation_id"`
	LatestCommittedReceiptDigest string      `json:"latest_committed_receipt_digest"`
	LatestCommittedSequence      string      `json:"latest_committed_sequence"`
	LatestReservationDigest      string      `json:"latest_reservation_digest"`
	LatestReservedSequence       string      `json:"latest_reserved_sequence"`
	RequiredLSN                  WALPosition `json:"required_lsn"`
	SchemaVersion                string      `json:"schema_version"`
}

func NewAuthorityProviderHeadSnapshot(head Head) (AuthorityProviderHeadSnapshot, error) {
	if head.Validate() != nil || head.LatestReservedSequence == 0 || head.LatestCommittedSequence == 0 ||
		head.LatestReservationDigest == (contracts.Digest{}) || head.LatestCommittedDatabasePoint == nil {
		return AuthorityProviderHeadSnapshot{}, ErrInvalidArgument
	}
	point := head.LatestCommittedDatabasePoint
	wire := authorityProviderHeadV1{
		AuthorityEpoch: strconv.FormatUint(head.Epoch, 10), DBSystemID: strconv.FormatUint(point.SystemID, 10),
		DBTimeline: strconv.FormatUint(uint64(point.Timeline), 10), LatestCommittedOperationID: head.LatestCommittedOperationID.String(),
		LatestCommittedReceiptDigest: digestString(head.LatestCommittedReceiptDigest),
		LatestCommittedSequence:      strconv.FormatUint(head.LatestCommittedSequence, 10),
		LatestReservationDigest:      digestString(head.LatestReservationDigest),
		LatestReservedSequence:       strconv.FormatUint(head.LatestReservedSequence, 10), RequiredLSN: point.RequiredLSN, SchemaVersion: "1",
	}
	canonical, digest, err := canonicalAuthorityArtifact(authorityProviderHeadDomain, wire)
	if err != nil {
		return AuthorityProviderHeadSnapshot{}, ErrInvalidArgument
	}
	return AuthorityProviderHeadSnapshot{facts: cloneHeadValue(head), canonical: canonical, digest: digest}, nil
}

func ParseAuthorityProviderHeadSnapshot(body []byte) (AuthorityProviderHeadSnapshot, error) {
	var wire authorityProviderHeadV1
	if err := decodeCanonicalAuthorityArtifact(body, &wire); err != nil || wire.SchemaVersion != "1" {
		return AuthorityProviderHeadSnapshot{}, ErrInvalidArgument
	}
	epoch, ok := parsePositiveDecimal(wire.AuthorityEpoch, math.MaxInt64)
	if !ok {
		return AuthorityProviderHeadSnapshot{}, ErrInvalidArgument
	}
	reserved, ok := parsePositiveDecimal(wire.LatestReservedSequence, math.MaxInt64)
	if !ok {
		return AuthorityProviderHeadSnapshot{}, ErrInvalidArgument
	}
	committed, ok := parsePositiveDecimal(wire.LatestCommittedSequence, math.MaxInt64)
	if !ok {
		return AuthorityProviderHeadSnapshot{}, ErrInvalidArgument
	}
	systemID, ok := parsePositiveDecimal(wire.DBSystemID, math.MaxUint64)
	if !ok {
		return AuthorityProviderHeadSnapshot{}, ErrInvalidArgument
	}
	timeline, ok := parsePositiveDecimal(wire.DBTimeline, math.MaxUint32)
	if !ok {
		return AuthorityProviderHeadSnapshot{}, ErrInvalidArgument
	}
	operationID, ok := parseCanonicalUUID(wire.LatestCommittedOperationID)
	if !ok {
		return AuthorityProviderHeadSnapshot{}, ErrInvalidArgument
	}
	reservationDigest, ok := parseDigestString(wire.LatestReservationDigest, false)
	if !ok {
		return AuthorityProviderHeadSnapshot{}, ErrInvalidArgument
	}
	receiptDigest, ok := parseDigestString(wire.LatestCommittedReceiptDigest, false)
	if !ok {
		return AuthorityProviderHeadSnapshot{}, ErrInvalidArgument
	}
	head := Head{
		Epoch: epoch, LatestReservedSequence: reserved, LatestCommittedSequence: committed,
		LatestReservationDigest: reservationDigest, LatestCommittedOperationID: operationID,
		LatestCommittedReceiptDigest: receiptDigest,
		LatestCommittedDatabasePoint: &DatabasePoint{SystemID: systemID, Timeline: uint32(timeline), RequiredLSN: wire.RequiredLSN},
	}
	value, err := NewAuthorityProviderHeadSnapshot(head)
	if err != nil || !bytes.Equal(value.canonical, body) {
		return AuthorityProviderHeadSnapshot{}, ErrInvalidArgument
	}
	return value, nil
}

func (value AuthorityProviderHeadSnapshot) Facts() Head              { return cloneHeadValue(value.facts) }
func (value AuthorityProviderHeadSnapshot) CanonicalJCS() []byte     { return bytes.Clone(value.canonical) }
func (value AuthorityProviderHeadSnapshot) Digest() contracts.Digest { return value.digest }
func AuthorityProviderHeadDigest(head Head) (contracts.Digest, error) {
	value, err := NewAuthorityProviderHeadSnapshot(head)
	return value.Digest(), err
}

func validAuthorityProviderHead(value AuthorityProviderHeadSnapshot) bool {
	rebuilt, err := NewAuthorityProviderHeadSnapshot(value.facts)
	return err == nil && rebuilt.digest == value.digest && bytes.Equal(rebuilt.canonical, value.canonical)
}

type CheckpointKind string

const (
	CheckpointNone   CheckpointKind = "none"
	CheckpointNode   CheckpointKind = "node"
	CheckpointGlobal CheckpointKind = "global"
)

type AuthorityCheckpointAnchorFacts struct {
	Kind          CheckpointKind
	ScopeDigest   contracts.Digest
	Epoch         uint64
	Sequence      uint64
	ReceiptDigest contracts.Digest
}

type AuthorityCheckpointAnchorInput struct {
	Kind        CheckpointKind
	ScopeDigest contracts.Digest
	Checkpoint  NodeCheckpoint
}

type AuthorityCheckpointAnchor struct {
	facts     AuthorityCheckpointAnchorFacts
	canonical []byte
	digest    contracts.Digest
}

type authorityCheckpointAnchorV1 struct {
	AuthorityEpoch        string         `json:"authority_epoch"`
	AuthoritySequence     string         `json:"authority_sequence"`
	CheckpointKind        CheckpointKind `json:"checkpoint_kind"`
	CheckpointScopeDigest string         `json:"checkpoint_scope_digest"`
	ReceiptDigest         string         `json:"receipt_digest"`
	SchemaVersion         string         `json:"schema_version"`
}

func NewAuthorityCheckpointAnchor(input AuthorityCheckpointAnchorInput) (AuthorityCheckpointAnchor, error) {
	if input.Checkpoint.Validate() != nil || input.Checkpoint.Sequence == 0 || input.Checkpoint.ReceiptDigest == (contracts.Digest{}) {
		return AuthorityCheckpointAnchor{}, ErrInvalidArgument
	}
	switch input.Kind {
	case CheckpointNode:
		if !validScopeDigest(ScopeNode, input.ScopeDigest) {
			return AuthorityCheckpointAnchor{}, ErrInvalidArgument
		}
	case CheckpointGlobal:
		if input.ScopeDigest != globalNodeScopeDigest && input.ScopeDigest != globalOperatorScopeDigest {
			return AuthorityCheckpointAnchor{}, ErrInvalidArgument
		}
	default:
		return AuthorityCheckpointAnchor{}, ErrInvalidArgument
	}
	wire := authorityCheckpointAnchorV1{
		AuthorityEpoch:    strconv.FormatUint(input.Checkpoint.AuthorityEpoch, 10),
		AuthoritySequence: strconv.FormatUint(input.Checkpoint.Sequence, 10), CheckpointKind: input.Kind,
		CheckpointScopeDigest: digestString(input.ScopeDigest), ReceiptDigest: digestString(input.Checkpoint.ReceiptDigest), SchemaVersion: "1",
	}
	canonical, digest, err := canonicalAuthorityArtifact(authorityCheckpointAnchorDomain, wire)
	if err != nil {
		return AuthorityCheckpointAnchor{}, ErrInvalidArgument
	}
	return AuthorityCheckpointAnchor{facts: AuthorityCheckpointAnchorFacts{
		Kind: input.Kind, ScopeDigest: input.ScopeDigest, Epoch: input.Checkpoint.AuthorityEpoch,
		Sequence: input.Checkpoint.Sequence, ReceiptDigest: input.Checkpoint.ReceiptDigest,
	}, canonical: canonical, digest: digest}, nil
}

func ParseAuthorityCheckpointAnchor(body []byte) (AuthorityCheckpointAnchor, error) {
	var wire authorityCheckpointAnchorV1
	if err := decodeCanonicalAuthorityArtifact(body, &wire); err != nil || wire.SchemaVersion != "1" {
		return AuthorityCheckpointAnchor{}, ErrInvalidArgument
	}
	epoch, ok := parsePositiveDecimal(wire.AuthorityEpoch, math.MaxInt64)
	if !ok {
		return AuthorityCheckpointAnchor{}, ErrInvalidArgument
	}
	sequence, ok := parsePositiveDecimal(wire.AuthoritySequence, math.MaxInt64)
	if !ok {
		return AuthorityCheckpointAnchor{}, ErrInvalidArgument
	}
	scopeDigest, ok := parseDigestString(wire.CheckpointScopeDigest, false)
	if !ok {
		return AuthorityCheckpointAnchor{}, ErrInvalidArgument
	}
	receiptDigest, ok := parseDigestString(wire.ReceiptDigest, false)
	if !ok {
		return AuthorityCheckpointAnchor{}, ErrInvalidArgument
	}
	value, err := NewAuthorityCheckpointAnchor(AuthorityCheckpointAnchorInput{
		Kind: wire.CheckpointKind, ScopeDigest: scopeDigest,
		Checkpoint: NodeCheckpoint{AuthorityEpoch: epoch, Sequence: sequence, ReceiptDigest: receiptDigest},
	})
	if err != nil || !bytes.Equal(value.canonical, body) {
		return AuthorityCheckpointAnchor{}, ErrInvalidArgument
	}
	return value, nil
}

func (value AuthorityCheckpointAnchor) Facts() AuthorityCheckpointAnchorFacts { return value.facts }
func (value AuthorityCheckpointAnchor) CanonicalJCS() []byte                  { return bytes.Clone(value.canonical) }
func (value AuthorityCheckpointAnchor) Digest() contracts.Digest              { return value.digest }

func validAuthorityCheckpointAnchor(value AuthorityCheckpointAnchor) bool {
	rebuilt, err := NewAuthorityCheckpointAnchor(AuthorityCheckpointAnchorInput{
		Kind: value.facts.Kind, ScopeDigest: value.facts.ScopeDigest,
		Checkpoint: NodeCheckpoint{AuthorityEpoch: value.facts.Epoch, Sequence: value.facts.Sequence, ReceiptDigest: value.facts.ReceiptDigest},
	})
	return err == nil && rebuilt.digest == value.digest && bytes.Equal(rebuilt.canonical, value.canonical)
}

func canonicalAuthorityArtifact(domain string, wire any) ([]byte, contracts.Digest, error) {
	body, err := json.Marshal(wire)
	if err != nil {
		return nil, contracts.Digest{}, err
	}
	canonical, err := jcs.Transform(body)
	if err != nil || len(canonical) == 0 || len(canonical) > authorityCanonicalMaximumBytes {
		return nil, contracts.Digest{}, ErrInvalidArgument
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte(domain))
	_, _ = hash.Write(canonical)
	var digest contracts.Digest
	copy(digest[:], hash.Sum(nil))
	return bytes.Clone(canonical), digest, nil
}

func decodeCanonicalAuthorityArtifact(body []byte, wire any) error {
	if len(body) == 0 || len(body) > authorityCanonicalMaximumBytes {
		return ErrInvalidArgument
	}
	if err := strictjson.Decode(bytes.NewReader(body), authorityCanonicalMaximumBytes, wire); err != nil {
		return ErrInvalidArgument
	}
	encoded, err := json.Marshal(wire)
	if err != nil {
		return ErrInvalidArgument
	}
	canonical, err := jcs.Transform(encoded)
	if err != nil || !bytes.Equal(canonical, body) {
		return ErrInvalidArgument
	}
	return nil
}

func digestString(value contracts.Digest) string { return hex.EncodeToString(value[:]) }

func parseDigestString(text string, allowAbsent bool) (contracts.Digest, bool) {
	if len(text) != 64 || strings.ToLower(text) != text {
		return contracts.Digest{}, false
	}
	decoded, err := hex.DecodeString(text)
	if err != nil || len(decoded) != 32 {
		return contracts.Digest{}, false
	}
	var value contracts.Digest
	copy(value[:], decoded)
	if value == (contracts.Digest{}) && !allowAbsent {
		return contracts.Digest{}, false
	}
	return value, true
}

func parsePositiveDecimal(text string, maximum uint64) (uint64, bool) {
	if text == "" || (len(text) > 1 && text[0] == '0') {
		return 0, false
	}
	value, err := strconv.ParseUint(text, 10, 64)
	return value, err == nil && value > 0 && value <= maximum && strconv.FormatUint(value, 10) == text
}

func parseCanonicalUUID(text string) (uuid.UUID, bool) {
	value, err := uuid.Parse(text)
	return value, err == nil && validOperationID(value) && value.String() == text
}

func cloneHeadValue(value Head) Head {
	if value.LatestCommittedDatabasePoint != nil {
		point := *value.LatestCommittedDatabasePoint
		value.LatestCommittedDatabasePoint = &point
	}
	return value
}
