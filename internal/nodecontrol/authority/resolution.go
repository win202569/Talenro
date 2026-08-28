package authority

import (
	"bytes"

	"talenro.local/platform/internal/nodecontrol/contracts"
)

const authorityEffectResolutionDomain = "talenro.nodecontrol.authority-effect-resolution.v1\x00"

type EffectDisposition string

const (
	DispositionApplied    EffectDisposition = "applied"
	DispositionNotApplied EffectDisposition = "not_applied"
)

type DecisionAnchorKind string

const (
	DecisionAnchorFinalCommitment DecisionAnchorKind = "final_commitment"
	DecisionAnchorExactCapture    DecisionAnchorKind = "exact_capture"
	DecisionAnchorHigherAuthority DecisionAnchorKind = "higher_authority"
	DecisionAnchorTrustedTime     DecisionAnchorKind = "trusted_time"
)

type AuthorityEffectResolutionInput struct {
	Commitment   AuthorityEffectCommitment
	Disposition  EffectDisposition
	Reason       AuthorityEffectReason
	AnchorKind   DecisionAnchorKind
	AnchorDigest contracts.Digest
	Evidence     ValidatedActivationDecisionEvidence
}

type AuthorityEffectResolutionFacts struct {
	CommitmentDigest contracts.Digest
	Disposition      EffectDisposition
	Reason           AuthorityEffectReason
	AnchorKind       DecisionAnchorKind
	AnchorDigest     contracts.Digest
}

type AuthorityEffectResolution struct {
	facts     AuthorityEffectResolutionFacts
	canonical []byte
	digest    contracts.Digest
}

type authorityEffectResolutionV1 struct {
	CommitmentDigest     string                `json:"commitment_digest"`
	DecisionAnchorDigest string                `json:"decision_anchor_digest"`
	DecisionAnchorKind   DecisionAnchorKind    `json:"decision_anchor_kind"`
	Disposition          EffectDisposition     `json:"disposition"`
	ReasonCode           AuthorityEffectReason `json:"reason_code"`
	SchemaVersion        string                `json:"schema_version"`
}

func NewAuthorityEffectResolution(input AuthorityEffectResolutionInput) (AuthorityEffectResolution, error) {
	if !validAuthorityEffectCommitment(input.Commitment) || !validValidatedEvidence(input.Evidence) ||
		input.AnchorDigest == (contracts.Digest{}) || !resolutionEnumsValid(input.Disposition, input.Reason, input.AnchorKind) {
		return AuthorityEffectResolution{}, ErrInvalidArgument
	}
	if input.Evidence.origin != activationEvidenceOriginFresh {
		return AuthorityEffectResolution{}, ErrConflict
	}
	if input.Evidence.input.Material.Commitment.Digest() != input.Commitment.Digest() ||
		!bytes.Equal(input.Evidence.input.Material.Commitment.CanonicalJCS(), input.Commitment.CanonicalJCS()) {
		return AuthorityEffectResolution{}, ErrConflict
	}
	if err := validateResolutionMatrix(input.Disposition, input.Reason, input.AnchorKind, input.AnchorDigest, input.Commitment, input.Evidence, true); err != nil {
		return AuthorityEffectResolution{}, err
	}
	wire := authorityEffectResolutionV1{
		CommitmentDigest: digestString(input.Commitment.Digest()), DecisionAnchorDigest: digestString(input.AnchorDigest),
		DecisionAnchorKind: input.AnchorKind, Disposition: input.Disposition, ReasonCode: input.Reason, SchemaVersion: "1",
	}
	canonical, digest, err := canonicalAuthorityArtifact(authorityEffectResolutionDomain, wire)
	if err != nil {
		return AuthorityEffectResolution{}, ErrInvalidArgument
	}
	return AuthorityEffectResolution{facts: AuthorityEffectResolutionFacts{
		CommitmentDigest: input.Commitment.Digest(), Disposition: input.Disposition, Reason: input.Reason,
		AnchorKind: input.AnchorKind, AnchorDigest: input.AnchorDigest,
	}, canonical: canonical, digest: digest}, nil
}

func ParseAuthorityEffectResolution(body []byte) (AuthorityEffectResolution, error) {
	var wire authorityEffectResolutionV1
	if err := decodeCanonicalAuthorityArtifact(body, &wire); err != nil || wire.SchemaVersion != "1" ||
		!resolutionEnumsValid(wire.Disposition, wire.ReasonCode, wire.DecisionAnchorKind) {
		return AuthorityEffectResolution{}, ErrInvalidArgument
	}
	commitmentDigest, ok := parseDigestString(wire.CommitmentDigest, false)
	if !ok {
		return AuthorityEffectResolution{}, ErrInvalidArgument
	}
	anchorDigest, ok := parseDigestString(wire.DecisionAnchorDigest, false)
	if !ok {
		return AuthorityEffectResolution{}, ErrInvalidArgument
	}
	if wire.Disposition == DispositionApplied && (wire.ReasonCode != EffectReasonNone || wire.DecisionAnchorKind != DecisionAnchorTrustedTime) {
		return AuthorityEffectResolution{}, ErrInvalidArgument
	}
	if wire.Disposition == DispositionNotApplied && !wire.ReasonCode.finalReason() {
		return AuthorityEffectResolution{}, ErrInvalidArgument
	}
	canonical, digest, err := canonicalAuthorityArtifact(authorityEffectResolutionDomain, wire)
	if err != nil || !bytes.Equal(canonical, body) {
		return AuthorityEffectResolution{}, ErrInvalidArgument
	}
	return AuthorityEffectResolution{facts: AuthorityEffectResolutionFacts{
		CommitmentDigest: commitmentDigest, Disposition: wire.Disposition, Reason: wire.ReasonCode,
		AnchorKind: wire.DecisionAnchorKind, AnchorDigest: anchorDigest,
	}, canonical: canonical, digest: digest}, nil
}

func ValidateAuthorityEffectResolution(resolution AuthorityEffectResolution, commitment AuthorityEffectCommitment, proof ValidatedActivationDecisionEvidence) error {
	if !validAuthorityEffectResolution(resolution) || !validAuthorityEffectCommitment(commitment) || !validValidatedEvidence(proof) {
		return ErrInvalidArgument
	}
	if resolution.facts.CommitmentDigest != commitment.Digest() || proof.input.Material.Commitment.Digest() != commitment.Digest() {
		return ErrConflict
	}
	return validateResolutionMatrix(
		resolution.facts.Disposition, resolution.facts.Reason, resolution.facts.AnchorKind,
		resolution.facts.AnchorDigest, commitment, proof, false,
	)
}

func (value AuthorityEffectResolution) Facts() AuthorityEffectResolutionFacts { return value.facts }
func (value AuthorityEffectResolution) CanonicalJCS() []byte                  { return bytes.Clone(value.canonical) }
func (value AuthorityEffectResolution) Digest() contracts.Digest              { return value.digest }

func validateResolutionMatrix(disposition EffectDisposition, reason AuthorityEffectReason, anchorKind DecisionAnchorKind,
	anchorDigest contracts.Digest, commitment AuthorityEffectCommitment, proof ValidatedActivationDecisionEvidence, requireFresh bool,
) error {
	commitmentFacts := commitment.Facts()
	evidence := proof.evidence
	material := proof.input.Material
	if proof.evidence.facts.CommitmentDigest != commitment.Digest() || material.Commitment.Digest() != commitment.Digest() {
		return ErrConflict
	}
	rollbackResistant := evidence.facts.TrustedTimeKind == TrustedTimeRollbackResistant
	if rollbackResistant {
		if requireFresh && (proof.admission == nil || proof.admission.state.Load() != 0) {
			return ErrConflict
		}
	} else if proof.admission != nil {
		return ErrConflict
	}

	switch {
	case commitmentFacts.Mode == CommitmentFinalNotApplied:
		if disposition != DispositionNotApplied || reason != commitmentFacts.Reason || anchorKind != DecisionAnchorFinalCommitment ||
			anchorDigest != commitment.Digest() || rollbackResistant || evidence.facts.CheckpointKind != CheckpointNone ||
			evidence.facts.Capability != DecisionCapabilityNotAppliedOnly {
			return ErrConflict
		}
	case commitmentFacts.Mode == CommitmentConditionalApply && disposition == DispositionApplied:
		if reason != EffectReasonNone || anchorKind != DecisionAnchorTrustedTime || anchorDigest != evidence.Digest() ||
			!rollbackResistant || evidence.facts.Capability != DecisionCapabilityMayApply || evidence.facts.CheckpointKind != CheckpointNone ||
			material.Reason != EffectReasonNone {
			return ErrConflict
		}
	case commitmentFacts.Mode == CommitmentConditionalApply && disposition == DispositionNotApplied &&
		(reason == EffectReasonFailed || reason == EffectReasonValidationRejected):
		if material.Reason != reason || anchorKind != DecisionAnchorExactCapture ||
			anchorDigest != commitmentFacts.ActivationInputsDigest || !rollbackResistant ||
			evidence.facts.Capability != DecisionCapabilityNotAppliedOnly || evidence.facts.CheckpointKind != CheckpointNone {
			return ErrConflict
		}
	case commitmentFacts.Mode == CommitmentConditionalApply && disposition == DispositionNotApplied && reason == EffectReasonSuperseded:
		if material.Reason != reason || anchorKind != DecisionAnchorHigherAuthority || proof.input.Checkpoint == nil ||
			anchorDigest != proof.input.Checkpoint.Digest() || rollbackResistant ||
			(evidence.facts.CheckpointKind != CheckpointNode && evidence.facts.CheckpointKind != CheckpointGlobal) ||
			evidence.facts.Capability != DecisionCapabilityNotAppliedOnly {
			return ErrConflict
		}
	case commitmentFacts.Mode == CommitmentConditionalApply && disposition == DispositionNotApplied && reason == EffectReasonActivationDeadlineExpired:
		if material.Reason != reason || anchorKind != DecisionAnchorTrustedTime || anchorDigest != evidence.Digest() ||
			!rollbackResistant || evidence.facts.Capability != DecisionCapabilityNotAppliedOnly || evidence.facts.CheckpointKind != CheckpointNone {
			return ErrConflict
		}
	default:
		return ErrConflict
	}
	return nil
}

func resolutionEnumsValid(disposition EffectDisposition, reason AuthorityEffectReason, anchor DecisionAnchorKind) bool {
	if disposition != DispositionApplied && disposition != DispositionNotApplied {
		return false
	}
	if !reason.valid() {
		return false
	}
	switch anchor {
	case DecisionAnchorFinalCommitment, DecisionAnchorExactCapture, DecisionAnchorHigherAuthority, DecisionAnchorTrustedTime:
		return true
	default:
		return false
	}
}

func validAuthorityEffectResolution(value AuthorityEffectResolution) bool {
	if value.digest == (contracts.Digest{}) || len(value.canonical) == 0 ||
		value.facts.CommitmentDigest == (contracts.Digest{}) || value.facts.AnchorDigest == (contracts.Digest{}) ||
		!resolutionEnumsValid(value.facts.Disposition, value.facts.Reason, value.facts.AnchorKind) {
		return false
	}
	wire := authorityEffectResolutionV1{
		CommitmentDigest: digestString(value.facts.CommitmentDigest), DecisionAnchorDigest: digestString(value.facts.AnchorDigest),
		DecisionAnchorKind: value.facts.AnchorKind, Disposition: value.facts.Disposition, ReasonCode: value.facts.Reason, SchemaVersion: "1",
	}
	canonical, digest, err := canonicalAuthorityArtifact(authorityEffectResolutionDomain, wire)
	return err == nil && digest == value.digest && bytes.Equal(canonical, value.canonical)
}
