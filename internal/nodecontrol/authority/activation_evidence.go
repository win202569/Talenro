package authority

import (
	"bytes"
	"context"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

const activationDecisionEvidenceDomain = "talenro.nodecontrol.activation-decision-evidence.v1\x00"

type TrustedTimeKind string

const (
	TrustedTimeNone              TrustedTimeKind = "none"
	TrustedTimeRollbackResistant TrustedTimeKind = "rollback_resistant"
)

type DecisionCapability string

const (
	DecisionCapabilityMayApply       DecisionCapability = "may_apply"
	DecisionCapabilityNotAppliedOnly DecisionCapability = "not_applied_only"
)

type ActivationDecisionMaterial struct {
	Commitment                     AuthorityEffectCommitment
	Reason                         AuthorityEffectReason
	CheckpointKind                 CheckpointKind
	CheckpointScopeDigest          contracts.Digest
	TrustedTimeKind                TrustedTimeKind
	TrustedInstant                 time.Time
	EvidenceValidUntil             time.Time
	AttestationExpiresAt           time.Time
	ActivationDeadline             time.Time
	ProviderIdentityDigest         contracts.Digest
	ExpectedProviderIdentityDigest contracts.Digest
	FloorAttestationDigest         contracts.Digest
	Capability                     DecisionCapability
}

type ActivationDecisionEvidenceInput struct {
	Material     ActivationDecisionMaterial
	Receipt      Receipt
	ProviderHead AuthorityProviderHeadSnapshot
	Checkpoint   *AuthorityCheckpointAnchor
}

type ActivationDecisionEvidenceFacts struct {
	OperationID            uuid.UUID
	CommitmentDigest       contracts.Digest
	ProviderHeadDigest     contracts.Digest
	CheckpointKind         CheckpointKind
	CheckpointScopeDigest  contracts.Digest
	CheckpointDigest       contracts.Digest
	TrustedTimeKind        TrustedTimeKind
	TrustedInstant         time.Time
	EvidenceValidUntil     time.Time
	ProviderIdentityDigest contracts.Digest
	FloorAttestationDigest contracts.Digest
	Capability             DecisionCapability
}

type activationEvidenceOrigin uint8

const (
	activationEvidenceOriginNone activationEvidenceOrigin = iota
	activationEvidenceOriginFresh
	activationEvidenceOriginParsed
)

type activationAdmission struct {
	state            atomic.Uint32
	operationID      uuid.UUID
	commitmentDigest contracts.Digest
	evidenceDigest   contracts.Digest
	deadline         time.Time
	now              func() time.Time
}

type ActivationDecisionEvidence struct {
	facts     ActivationDecisionEvidenceFacts
	canonical []byte
	digest    contracts.Digest
	origin    activationEvidenceOrigin
	admission *activationAdmission
}

type ValidatedActivationDecisionEvidence struct {
	evidence  ActivationDecisionEvidence
	input     ActivationDecisionEvidenceInput
	origin    activationEvidenceOrigin
	admission *activationAdmission
	valid     bool
}

type activationDecisionEvidenceV1 struct {
	CheckpointDigest       string             `json:"checkpoint_digest"`
	CheckpointKind         CheckpointKind     `json:"checkpoint_kind"`
	CheckpointScopeDigest  string             `json:"checkpoint_scope_digest"`
	CommitmentDigest       string             `json:"commitment_digest"`
	DecisionCapability     DecisionCapability `json:"decision_capability"`
	EvidenceValidUntil     string             `json:"evidence_valid_until"`
	FloorAttestationDigest string             `json:"floor_attestation_digest"`
	OperationID            string             `json:"operation_id"`
	ProviderHeadDigest     string             `json:"provider_head_digest"`
	ProviderIdentityDigest string             `json:"provider_identity_digest"`
	SchemaVersion          string             `json:"schema_version"`
	TrustedInstant         string             `json:"trusted_instant"`
	TrustedTimeKind        TrustedTimeKind    `json:"trusted_time_kind"`
}

type ActivationEvidenceCapture struct {
	started time.Time
	now     func() time.Time
	used    atomic.Bool
}

func BeginActivationEvidenceCapture() *ActivationEvidenceCapture {
	return beginActivationEvidenceCaptureForTest(time.Now)
}

func beginActivationEvidenceCaptureForTest(now func() time.Time) *ActivationEvidenceCapture {
	if now == nil {
		return &ActivationEvidenceCapture{}
	}
	return &ActivationEvidenceCapture{started: now(), now: now}
}

func (capture *ActivationEvidenceCapture) Complete(input ActivationDecisionEvidenceInput) (ActivationDecisionEvidence, error) {
	if capture == nil || capture.now == nil || capture.started.IsZero() {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	if !capture.used.CompareAndSwap(false, true) {
		return ActivationDecisionEvidence{}, ErrConflict
	}
	if input.Material.TrustedTimeKind != TrustedTimeRollbackResistant {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	evidence, err := buildActivationDecisionEvidence(input, activationEvidenceOriginFresh, nil)
	if err != nil {
		return ActivationDecisionEvidence{}, err
	}
	budget := input.Material.EvidenceValidUntil.Sub(input.Material.TrustedInstant)
	if budget <= 0 {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	if budget > 5*time.Second {
		budget = 5 * time.Second
	}
	deadline, ok := addCaptureDuration(capture.started, budget)
	if !ok {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	if !capture.now().Before(deadline) {
		return ActivationDecisionEvidence{}, ErrConflict
	}
	evidence.admission = &activationAdmission{
		operationID: evidence.facts.OperationID, commitmentDigest: evidence.facts.CommitmentDigest,
		evidenceDigest: evidence.digest, deadline: deadline, now: capture.now,
	}
	return evidence, nil
}

func NewActivationDecisionEvidence(input ActivationDecisionEvidenceInput) (ActivationDecisionEvidence, error) {
	if input.Material.TrustedTimeKind != TrustedTimeNone {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	return buildActivationDecisionEvidence(input, activationEvidenceOriginFresh, nil)
}

func buildActivationDecisionEvidence(input ActivationDecisionEvidenceInput, origin activationEvidenceOrigin, admission *activationAdmission) (ActivationDecisionEvidence, error) {
	if origin != activationEvidenceOriginFresh && origin != activationEvidenceOriginParsed {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	material := input.Material
	if !validAuthorityEffectCommitment(material.Commitment) || !material.Reason.valid() ||
		input.Receipt.Validate() != nil || input.Receipt.Status != StatusCommitted || input.Receipt.EffectDigest == nil ||
		!validAuthorityProviderHead(input.ProviderHead) {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	commitment := material.Commitment.Facts()
	if input.Receipt.OperationID != commitment.OperationID || input.Receipt.Kind != commitment.Kind ||
		input.Receipt.ScopeKind != commitment.ScopeKind || input.Receipt.ScopeDigest != commitment.ScopeDigest ||
		input.Receipt.Epoch != commitment.Epoch || input.Receipt.Sequence != commitment.Sequence ||
		*input.Receipt.EffectDigest != material.Commitment.Digest() {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	head := input.ProviderHead.Facts()
	if head.Epoch != input.Receipt.Epoch || head.LatestCommittedSequence < input.Receipt.Sequence {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	if head.LatestCommittedSequence == input.Receipt.Sequence &&
		(head.LatestCommittedOperationID != input.Receipt.OperationID ||
			head.LatestCommittedReceiptDigest != input.Receipt.ReceiptDigest ||
			!equalOptionalDatabasePoint(head.LatestCommittedDatabasePoint, input.Receipt.DatabasePoint)) {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}

	facts := ActivationDecisionEvidenceFacts{
		OperationID: commitment.OperationID, CommitmentDigest: material.Commitment.Digest(),
		ProviderHeadDigest: input.ProviderHead.Digest(), CheckpointKind: material.CheckpointKind,
		CheckpointScopeDigest: material.CheckpointScopeDigest, TrustedTimeKind: material.TrustedTimeKind,
		TrustedInstant: canonicalTimeValue(material.TrustedInstant), EvidenceValidUntil: canonicalTimeValue(material.EvidenceValidUntil),
		ProviderIdentityDigest: material.ProviderIdentityDigest, FloorAttestationDigest: material.FloorAttestationDigest,
		Capability: material.Capability,
	}

	if err := validateEvidenceCheckpoint(input, commitment, head, &facts); err != nil {
		return ActivationDecisionEvidence{}, err
	}
	if err := validateEvidenceTimeAndMatrix(material, commitment); err != nil {
		return ActivationDecisionEvidence{}, err
	}
	wire := activationEvidenceWire(facts)
	canonical, digest, err := canonicalAuthorityArtifact(activationDecisionEvidenceDomain, wire)
	if err != nil {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	return ActivationDecisionEvidence{facts: facts, canonical: canonical, digest: digest, origin: origin, admission: admission}, nil
}

func validateEvidenceCheckpoint(input ActivationDecisionEvidenceInput, commitment AuthorityEffectCommitmentFacts, head Head, facts *ActivationDecisionEvidenceFacts) error {
	switch input.Material.CheckpointKind {
	case CheckpointNone:
		if input.Material.CheckpointScopeDigest != (contracts.Digest{}) || input.Checkpoint != nil {
			return ErrInvalidArgument
		}
		facts.CheckpointScopeDigest = contracts.Digest{}
		facts.CheckpointDigest = contracts.Digest{}
	case CheckpointNode, CheckpointGlobal:
		if input.Checkpoint == nil || !validAuthorityCheckpointAnchor(*input.Checkpoint) {
			return ErrInvalidArgument
		}
		anchor := input.Checkpoint.Facts()
		if anchor.Kind != input.Material.CheckpointKind || anchor.ScopeDigest != input.Material.CheckpointScopeDigest ||
			anchor.Epoch != commitment.Epoch || anchor.Sequence <= commitment.Sequence || head.LatestCommittedSequence < anchor.Sequence {
			return ErrInvalidArgument
		}
		if head.LatestCommittedSequence == anchor.Sequence && head.LatestCommittedReceiptDigest != anchor.ReceiptDigest {
			return ErrInvalidArgument
		}
		if anchor.Kind == CheckpointNode {
			if commitment.ScopeKind != ScopeNode || anchor.ScopeDigest != commitment.ScopeDigest {
				return ErrInvalidArgument
			}
		} else if (commitment.ScopeKind != ScopeGlobalNodeTrust && commitment.ScopeKind != ScopeGlobalOperatorTrust) ||
			anchor.ScopeDigest != commitment.ScopeDigest {
			return ErrInvalidArgument
		}
		facts.CheckpointDigest = input.Checkpoint.Digest()
	default:
		return ErrInvalidArgument
	}
	return nil
}

func validateEvidenceTimeAndMatrix(material ActivationDecisionMaterial, commitment AuthorityEffectCommitmentFacts) error {
	if material.Capability != DecisionCapabilityMayApply && material.Capability != DecisionCapabilityNotAppliedOnly {
		return ErrInvalidArgument
	}
	switch material.TrustedTimeKind {
	case TrustedTimeNone:
		if !material.TrustedInstant.IsZero() || !material.EvidenceValidUntil.IsZero() || !material.AttestationExpiresAt.IsZero() ||
			!material.ActivationDeadline.IsZero() || material.ProviderIdentityDigest != (contracts.Digest{}) ||
			material.ExpectedProviderIdentityDigest != (contracts.Digest{}) || material.FloorAttestationDigest != (contracts.Digest{}) ||
			material.Capability != DecisionCapabilityNotAppliedOnly {
			return ErrInvalidArgument
		}
	case TrustedTimeRollbackResistant:
		if !canonicalInstant(material.TrustedInstant) || !canonicalInstant(material.EvidenceValidUntil) ||
			!canonicalInstant(material.AttestationExpiresAt) || !canonicalInstant(material.ActivationDeadline) ||
			material.ProviderIdentityDigest == (contracts.Digest{}) || material.ExpectedProviderIdentityDigest == (contracts.Digest{}) ||
			material.ProviderIdentityDigest != material.ExpectedProviderIdentityDigest || material.FloorAttestationDigest == (contracts.Digest{}) ||
			!material.TrustedInstant.Before(material.EvidenceValidUntil) || !material.TrustedInstant.Before(material.AttestationExpiresAt) {
			return ErrInvalidArgument
		}
		fiveSecondBound, ok := addCanonicalDuration(material.TrustedInstant, 5*time.Second)
		if !ok {
			return ErrInvalidArgument
		}
		upper := earlierInstant(material.AttestationExpiresAt, fiveSecondBound)
		if material.Reason != EffectReasonActivationDeadlineExpired {
			if !material.TrustedInstant.Before(material.ActivationDeadline) {
				return ErrInvalidArgument
			}
			upper = earlierInstant(upper, material.ActivationDeadline)
		} else if material.TrustedInstant.Before(material.ActivationDeadline) {
			return ErrInvalidArgument
		}
		if material.EvidenceValidUntil.After(upper) {
			return ErrInvalidArgument
		}
	default:
		return ErrInvalidArgument
	}

	switch commitment.Mode {
	case CommitmentFinalNotApplied:
		if material.Reason != commitment.Reason || material.CheckpointKind != CheckpointNone ||
			material.TrustedTimeKind != TrustedTimeNone || material.Capability != DecisionCapabilityNotAppliedOnly {
			return ErrInvalidArgument
		}
	case CommitmentConditionalApply:
		switch material.Reason {
		case EffectReasonNone:
			if material.CheckpointKind != CheckpointNone || material.TrustedTimeKind != TrustedTimeRollbackResistant || material.Capability != DecisionCapabilityMayApply {
				return ErrInvalidArgument
			}
		case EffectReasonFailed, EffectReasonValidationRejected:
			if material.CheckpointKind != CheckpointNone || material.TrustedTimeKind != TrustedTimeRollbackResistant || material.Capability != DecisionCapabilityNotAppliedOnly {
				return ErrInvalidArgument
			}
		case EffectReasonActivationDeadlineExpired:
			if material.CheckpointKind != CheckpointNone || material.TrustedTimeKind != TrustedTimeRollbackResistant || material.Capability != DecisionCapabilityNotAppliedOnly {
				return ErrInvalidArgument
			}
		case EffectReasonSuperseded:
			if (material.CheckpointKind != CheckpointNode && material.CheckpointKind != CheckpointGlobal) ||
				material.TrustedTimeKind != TrustedTimeNone || material.Capability != DecisionCapabilityNotAppliedOnly {
				return ErrInvalidArgument
			}
		default:
			return ErrInvalidArgument
		}
	default:
		return ErrInvalidArgument
	}
	return nil
}

func ParseActivationDecisionEvidence(body []byte) (ActivationDecisionEvidence, error) {
	var wire activationDecisionEvidenceV1
	if err := decodeCanonicalAuthorityArtifact(body, &wire); err != nil || wire.SchemaVersion != "1" {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	operationID, ok := parseCanonicalUUID(wire.OperationID)
	if !ok {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	commitmentDigest, ok := parseDigestString(wire.CommitmentDigest, false)
	if !ok {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	headDigest, ok := parseDigestString(wire.ProviderHeadDigest, false)
	if !ok {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	checkpointScope, ok := parseDigestString(wire.CheckpointScopeDigest, true)
	if !ok {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	checkpointDigest, ok := parseDigestString(wire.CheckpointDigest, true)
	if !ok {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	providerIdentity, ok := parseDigestString(wire.ProviderIdentityDigest, true)
	if !ok {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	floorDigest, ok := parseDigestString(wire.FloorAttestationDigest, true)
	if !ok {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	facts := ActivationDecisionEvidenceFacts{
		OperationID: operationID, CommitmentDigest: commitmentDigest, ProviderHeadDigest: headDigest,
		CheckpointKind: wire.CheckpointKind, CheckpointScopeDigest: checkpointScope, CheckpointDigest: checkpointDigest,
		TrustedTimeKind: wire.TrustedTimeKind, ProviderIdentityDigest: providerIdentity,
		FloorAttestationDigest: floorDigest, Capability: wire.DecisionCapability,
	}
	switch wire.CheckpointKind {
	case CheckpointNone:
		if checkpointScope != (contracts.Digest{}) || checkpointDigest != (contracts.Digest{}) {
			return ActivationDecisionEvidence{}, ErrInvalidArgument
		}
	case CheckpointNode:
		if checkpointScope == (contracts.Digest{}) || checkpointDigest == (contracts.Digest{}) {
			return ActivationDecisionEvidence{}, ErrInvalidArgument
		}
		if !validScopeDigest(ScopeNode, checkpointScope) {
			return ActivationDecisionEvidence{}, ErrInvalidArgument
		}
	case CheckpointGlobal:
		if checkpointDigest == (contracts.Digest{}) ||
			(checkpointScope != globalNodeScopeDigest && checkpointScope != globalOperatorScopeDigest) {
			return ActivationDecisionEvidence{}, ErrInvalidArgument
		}
	default:
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	switch wire.TrustedTimeKind {
	case TrustedTimeNone:
		if wire.TrustedInstant != "" || wire.EvidenceValidUntil != "" || providerIdentity != (contracts.Digest{}) ||
			floorDigest != (contracts.Digest{}) || facts.Capability != DecisionCapabilityNotAppliedOnly {
			return ActivationDecisionEvidence{}, ErrInvalidArgument
		}
	case TrustedTimeRollbackResistant:
		trusted, valid := parseCanonicalInstant(wire.TrustedInstant)
		if !valid {
			return ActivationDecisionEvidence{}, ErrInvalidArgument
		}
		validUntil, valid := parseCanonicalInstant(wire.EvidenceValidUntil)
		if !valid || facts.CheckpointKind != CheckpointNone || !trusted.Before(validUntil) ||
			providerIdentity == (contracts.Digest{}) || floorDigest == (contracts.Digest{}) ||
			(facts.Capability != DecisionCapabilityMayApply && facts.Capability != DecisionCapabilityNotAppliedOnly) {
			return ActivationDecisionEvidence{}, ErrInvalidArgument
		}
		fiveSecondBound, valid := addCanonicalDuration(trusted, 5*time.Second)
		if !valid || validUntil.After(fiveSecondBound) {
			return ActivationDecisionEvidence{}, ErrInvalidArgument
		}
		facts.TrustedInstant, facts.EvidenceValidUntil = trusted, validUntil
	default:
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	canonical, digest, err := canonicalAuthorityArtifact(activationDecisionEvidenceDomain, wire)
	if err != nil || !bytes.Equal(canonical, body) {
		return ActivationDecisionEvidence{}, ErrInvalidArgument
	}
	return ActivationDecisionEvidence{facts: facts, canonical: canonical, digest: digest, origin: activationEvidenceOriginParsed}, nil
}

func ValidateActivationDecisionEvidence(evidence ActivationDecisionEvidence, input ActivationDecisionEvidenceInput) (ValidatedActivationDecisionEvidence, error) {
	if evidence.origin != activationEvidenceOriginFresh && evidence.origin != activationEvidenceOriginParsed {
		return ValidatedActivationDecisionEvidence{}, ErrInvalidArgument
	}
	expected, err := buildActivationDecisionEvidence(input, evidence.origin, evidence.admission)
	if err != nil {
		return ValidatedActivationDecisionEvidence{}, err
	}
	if expected.digest != evidence.digest || !bytes.Equal(expected.canonical, evidence.canonical) || expected.facts != evidence.facts {
		return ValidatedActivationDecisionEvidence{}, ErrConflict
	}
	admission := evidence.admission
	if evidence.origin == activationEvidenceOriginParsed || evidence.facts.TrustedTimeKind == TrustedTimeNone {
		admission = nil
	}
	return ValidatedActivationDecisionEvidence{
		evidence: cloneActivationEvidence(evidence), input: cloneActivationEvidenceInput(input),
		origin: evidence.origin, admission: admission, valid: true,
	}, nil
}

func (value ActivationDecisionEvidence) Facts() ActivationDecisionEvidenceFacts { return value.facts }
func (value ActivationDecisionEvidence) CanonicalJCS() []byte                   { return bytes.Clone(value.canonical) }
func (value ActivationDecisionEvidence) Digest() contracts.Digest               { return value.digest }

func (value ValidatedActivationDecisionEvidence) Evidence() ActivationDecisionEvidence {
	return cloneActivationEvidence(value.evidence)
}

func (value ValidatedActivationDecisionEvidence) Input() ActivationDecisionEvidenceInput {
	return cloneActivationEvidenceInput(value.input)
}

func consumeActivationAdmission(ctx context.Context, proof ValidatedActivationDecisionEvidence) error {
	if !proof.valid || proof.admission == nil {
		return ErrInvalidArgument
	}
	if !proof.admission.state.CompareAndSwap(0, 1) {
		return ErrConflict
	}
	if ctx == nil {
		return ErrInvalidArgument
	}
	if err := ctx.Err(); err != nil {
		return ErrCanceled
	}
	if proof.origin != activationEvidenceOriginFresh || proof.evidence.facts.TrustedTimeKind != TrustedTimeRollbackResistant ||
		proof.admission.operationID != proof.evidence.facts.OperationID ||
		proof.admission.commitmentDigest != proof.evidence.facts.CommitmentDigest ||
		proof.admission.evidenceDigest != proof.evidence.digest || proof.admission.now == nil {
		return ErrConflict
	}
	if !proof.admission.now().Before(proof.admission.deadline) {
		return ErrConflict
	}
	return nil
}

func activationEvidenceWire(facts ActivationDecisionEvidenceFacts) activationDecisionEvidenceV1 {
	wire := activationDecisionEvidenceV1{
		CheckpointDigest: digestString(facts.CheckpointDigest), CheckpointKind: facts.CheckpointKind,
		CheckpointScopeDigest: digestString(facts.CheckpointScopeDigest), CommitmentDigest: digestString(facts.CommitmentDigest),
		DecisionCapability: facts.Capability, FloorAttestationDigest: digestString(facts.FloorAttestationDigest),
		OperationID: facts.OperationID.String(), ProviderHeadDigest: digestString(facts.ProviderHeadDigest),
		ProviderIdentityDigest: digestString(facts.ProviderIdentityDigest), SchemaVersion: "1", TrustedTimeKind: facts.TrustedTimeKind,
	}
	if facts.TrustedTimeKind == TrustedTimeRollbackResistant {
		wire.TrustedInstant = facts.TrustedInstant.Format(time.RFC3339Nano)
		wire.EvidenceValidUntil = facts.EvidenceValidUntil.Format(time.RFC3339Nano)
	}
	return wire
}

func cloneActivationEvidence(value ActivationDecisionEvidence) ActivationDecisionEvidence {
	value.canonical = bytes.Clone(value.canonical)
	return value
}

func cloneActivationEvidenceInput(value ActivationDecisionEvidenceInput) ActivationDecisionEvidenceInput {
	value.Material.Commitment.canonical = bytes.Clone(value.Material.Commitment.canonical)
	value.Receipt = cloneReceiptValue(value.Receipt)
	value.ProviderHead.canonical = bytes.Clone(value.ProviderHead.canonical)
	value.ProviderHead.facts = cloneHeadValue(value.ProviderHead.facts)
	if value.Checkpoint != nil {
		checkpoint := *value.Checkpoint
		checkpoint.canonical = bytes.Clone(checkpoint.canonical)
		value.Checkpoint = &checkpoint
	}
	return value
}

func cloneReceiptValue(value Receipt) Receipt {
	if value.EffectDigest != nil {
		digest := *value.EffectDigest
		value.EffectDigest = &digest
	}
	if value.DatabasePoint != nil {
		point := *value.DatabasePoint
		value.DatabasePoint = &point
	}
	if value.AbortReason != nil {
		reason := *value.AbortReason
		value.AbortReason = &reason
	}
	return value
}

func canonicalInstant(value time.Time) bool {
	if value.IsZero() || value.Location() != time.UTC || value.Year() < 1 || value.Year() > 9999 {
		return false
	}
	text := value.Format(time.RFC3339Nano)
	parsed, err := time.Parse(time.RFC3339Nano, text)
	return err == nil && len(text) > 0 && text[len(text)-1] == 'Z' && parsed.Equal(value)
}

func canonicalTimeValue(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	return value.UTC().Round(0)
}

func parseCanonicalInstant(text string) (time.Time, bool) {
	if text == "" || text[len(text)-1] != 'Z' {
		return time.Time{}, false
	}
	value, err := time.Parse(time.RFC3339Nano, text)
	return value, err == nil && value.Location() == time.UTC && value.Format(time.RFC3339Nano) == text && canonicalInstant(value)
}

func addCanonicalDuration(value time.Time, duration time.Duration) (time.Time, bool) {
	if duration <= 0 || !canonicalInstant(value) {
		return time.Time{}, false
	}
	result := value.Add(duration)
	return result, result.After(value) && canonicalInstant(result)
}

func addCaptureDuration(value time.Time, duration time.Duration) (time.Time, bool) {
	if value.IsZero() || duration <= 0 {
		return time.Time{}, false
	}
	result := value.Add(duration)
	return result, result.After(value) && result.Year() >= 1 && result.Year() <= 9999
}

func earlierInstant(left, right time.Time) time.Time {
	if left.Before(right) {
		return left
	}
	return right
}

func validValidatedEvidence(proof ValidatedActivationDecisionEvidence) bool {
	if !proof.valid || (proof.origin != activationEvidenceOriginFresh && proof.origin != activationEvidenceOriginParsed) {
		return false
	}
	expected, err := buildActivationDecisionEvidence(proof.input, proof.origin, proof.admission)
	return err == nil && expected.digest == proof.evidence.digest && bytes.Equal(expected.canonical, proof.evidence.canonical) && expected.facts == proof.evidence.facts
}
