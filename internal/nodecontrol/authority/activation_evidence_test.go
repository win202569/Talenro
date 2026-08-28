package authority

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

func TestActivationDecisionEvidenceClosedMatrix(t *testing.T) {
	valid := []string{"may_apply", "exact_capture", "deadline_expired", "final", "higher_node"}
	for _, scenario := range valid {
		scenario := scenario
		t.Run(scenario, func(t *testing.T) {
			evidence, proof := freshEvidenceProofFixture(t, scenario)
			if evidence.Digest() == (contracts.Digest{}) || proof.Evidence().Digest() != evidence.Digest() {
				t.Fatal("valid row did not produce an exact branded proof")
			}
		})
	}

	validationRejected := evidenceInputFixture(t, "exact_capture")
	validationRejected.Material.Reason = EffectReasonValidationRejected
	if _, _, err := completeAndValidateEvidence(t, validationRejected, 0); err != nil {
		t.Fatalf("validation-rejected exact-capture row: %v", err)
	}
	if _, _, err := completeAndValidateEvidence(t, globalHigherAuthorityInput(t), 0); err != nil {
		t.Fatalf("global higher-authority row: %v", err)
	}

	tests := []struct {
		name   string
		base   string
		mutate func(*ActivationDecisionEvidenceInput)
	}{
		{name: "final with rollback time", base: "final", mutate: func(v *ActivationDecisionEvidenceInput) {
			v.Material.TrustedTimeKind = TrustedTimeRollbackResistant
			v.Material.TrustedInstant = time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
			v.Material.EvidenceValidUntil = time.Date(2026, 8, 28, 12, 0, 1, 0, time.UTC)
			v.Material.AttestationExpiresAt = time.Date(2026, 8, 28, 12, 0, 2, 0, time.UTC)
			v.Material.ActivationDeadline = time.Date(2026, 8, 28, 12, 0, 3, 0, time.UTC)
			v.Material.ProviderIdentityDigest = mustAuthorityDigest(t, "55"+string(bytes.Repeat([]byte("55"), 31)))
		}},
		{name: "conditional may-apply without time", base: "may_apply", mutate: clearEvidenceTime},
		{name: "conditional none not-applied capability", base: "may_apply", mutate: func(v *ActivationDecisionEvidenceInput) { v.Material.Capability = DecisionCapabilityNotAppliedOnly }},
		{name: "failed may-apply capability", base: "exact_capture", mutate: func(v *ActivationDecisionEvidenceInput) { v.Material.Capability = DecisionCapabilityMayApply }},
		{name: "deadline before activation deadline", base: "deadline_expired", mutate: func(v *ActivationDecisionEvidenceInput) {
			v.Material.TrustedInstant = v.Material.ActivationDeadline.Add(-time.Nanosecond)
		}},
		{name: "superseded with rollback time", base: "higher_node", mutate: func(v *ActivationDecisionEvidenceInput) {
			v.Material.TrustedTimeKind = TrustedTimeRollbackResistant
			v.Material.TrustedInstant = time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
			v.Material.EvidenceValidUntil = time.Date(2026, 8, 28, 12, 0, 1, 0, time.UTC)
			v.Material.AttestationExpiresAt = time.Date(2026, 8, 28, 12, 0, 2, 0, time.UTC)
			v.Material.ActivationDeadline = time.Date(2026, 8, 28, 12, 0, 3, 0, time.UTC)
			v.Material.ProviderIdentityDigest = mustAuthorityDigest(t, "55"+string(bytes.Repeat([]byte("55"), 31)))
			v.Material.ExpectedProviderIdentityDigest = v.Material.ProviderIdentityDigest
			v.Material.FloorAttestationDigest = mustAuthorityDigest(t, "66"+string(bytes.Repeat([]byte("66"), 31)))
		}},
		{name: "superseded without checkpoint", base: "higher_node", mutate: func(v *ActivationDecisionEvidenceInput) { v.Checkpoint = nil }},
		{name: "present checkpoint with may apply", base: "higher_node", mutate: func(v *ActivationDecisionEvidenceInput) { v.Material.Capability = DecisionCapabilityMayApply }},
		{name: "exact capture with checkpoint", base: "exact_capture", mutate: func(v *ActivationDecisionEvidenceInput) {
			anchor := checkpointFixture(t, "node")
			v.Material.CheckpointKind = CheckpointNode
			v.Material.CheckpointScopeDigest = anchor.Facts().ScopeDigest
			v.Checkpoint = &anchor
		}},
	}
	for _, test := range tests {
		test := test
		t.Run("reject_"+test.name, func(t *testing.T) {
			input := evidenceInputFixture(t, test.base)
			test.mutate(&input)
			if _, _, err := completeAndValidateEvidence(t, input, 0); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("error = %v, want ErrInvalidArgument", err)
			}
		})
	}
}

func TestActivationDecisionEvidenceContextRejectsForks(t *testing.T) {
	tests := []struct {
		name   string
		base   string
		mutate func(*ActivationDecisionEvidenceInput)
	}{
		{name: "receipt operation", base: "may_apply", mutate: func(v *ActivationDecisionEvidenceInput) {
			v.Receipt.OperationID = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
		}},
		{name: "receipt database point", base: "may_apply", mutate: func(v *ActivationDecisionEvidenceInput) { v.Receipt.DatabasePoint.RequiredLSN = "1/4" }},
		{name: "same sequence head operation", base: "may_apply", mutate: func(v *ActivationDecisionEvidenceInput) {
			v.ProviderHead = snapshotFromHead(t, mutateHead(headFixture(t, "exact"), func(h *Head) { h.LatestCommittedOperationID = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa") }))
		}},
		{name: "same sequence head receipt", base: "may_apply", mutate: func(v *ActivationDecisionEvidenceInput) {
			v.ProviderHead = snapshotFromHead(t, mutateHead(headFixture(t, "exact"), func(h *Head) {
				h.LatestCommittedReceiptDigest = mustAuthorityDigest(t, "aa"+string(bytes.Repeat([]byte("aa"), 31)))
			}))
		}},
		{name: "same sequence head database point", base: "may_apply", mutate: func(v *ActivationDecisionEvidenceInput) {
			v.ProviderHead = snapshotFromHead(t, mutateHead(headFixture(t, "exact"), func(h *Head) { h.LatestCommittedDatabasePoint.RequiredLSN = "1/4" }))
		}},
		{name: "lower head", base: "higher_node", mutate: func(v *ActivationDecisionEvidenceInput) {
			v.ProviderHead = snapshotFromHead(t, headFixture(t, "exact"))
		}},
		{name: "different head epoch", base: "higher_node", mutate: func(v *ActivationDecisionEvidenceInput) {
			v.ProviderHead = snapshotFromHead(t, mutateHead(headFixture(t, "later"), func(h *Head) { h.Epoch = 8 }))
		}},
		{name: "checkpoint equal sequence", base: "higher_node", mutate: func(v *ActivationDecisionEvidenceInput) {
			replaceCheckpoint(t, v, 7, 11, v.Checkpoint.Facts().ScopeDigest)
		}},
		{name: "checkpoint different epoch", base: "higher_node", mutate: func(v *ActivationDecisionEvidenceInput) {
			replaceCheckpoint(t, v, 8, 13, v.Checkpoint.Facts().ScopeDigest)
		}},
		{name: "checkpoint wrong scope", base: "higher_node", mutate: func(v *ActivationDecisionEvidenceInput) {
			replaceCheckpoint(t, v, 7, 13, mustAuthorityDigest(t, "aa"+string(bytes.Repeat([]byte("aa"), 31))))
		}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			input := evidenceInputFixture(t, test.base)
			test.mutate(&input)
			if _, _, err := completeAndValidateEvidence(t, input, 0); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("error = %v, want ErrInvalidArgument", err)
			}
		})
	}
}

func TestActivationDecisionEvidenceParseCannotMintAdmission(t *testing.T) {
	evidence, freshProof := freshEvidenceProofFixture(t, "may_apply")
	parsed, err := ParseActivationDecisionEvidence(evidence.CanonicalJCS())
	if err != nil {
		t.Fatal(err)
	}
	parsedProof, err := ValidateActivationDecisionEvidence(parsed, evidenceInputFixture(t, "may_apply"))
	if err != nil {
		t.Fatal(err)
	}
	if err := consumeActivationAdmission(context.Background(), parsedProof); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("parsed consume error = %v, want ErrInvalidArgument", err)
	}
	if _, err := NewAuthorityEffectResolution(AuthorityEffectResolutionInput{
		Commitment: parsedProof.Input().Material.Commitment, Disposition: DispositionApplied,
		Reason: EffectReasonNone, AnchorKind: DecisionAnchorTrustedTime, AnchorDigest: parsed.Digest(), Evidence: parsedProof,
	}); !errors.Is(err, ErrConflict) {
		t.Fatalf("parsed resolution constructor error = %v, want ErrConflict", err)
	}

	resolution := resolutionFixture(t, "applied")
	parsedResolution, err := ParseAuthorityEffectResolution(resolution.CanonicalJCS())
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateAuthorityEffectResolution(parsedResolution, parsedProof.Input().Material.Commitment, parsedProof); err != nil {
		t.Fatalf("persisted contextual validation: %v", err)
	}
	if freshProof.Evidence().Digest() != parsedProof.Evidence().Digest() {
		t.Fatal("parsed proof changed persisted evidence")
	}
}

func TestActivationDecisionEvidenceParseRejectsSelfContainedInvalidBranches(t *testing.T) {
	fixture := loadLiteralAuthorityEffectFixture(t)
	vectors := make(map[string]literalAuthorityEffectVector, len(fixture.Vectors))
	for _, vector := range fixture.Vectors {
		vectors[vector.Name] = vector
	}
	for _, test := range []struct {
		name   string
		vector string
		field  string
		value  string
	}{
		{name: "rollback window over five seconds", vector: "evidence-conditional-may-apply", field: "evidence_valid_until", value: "2026-08-28T12:00:06Z"},
		{name: "global kind with node scope", vector: "evidence-higher-authority-global", field: "checkpoint_scope_digest", value: "1111111111111111111111111111111111111111111111111111111111111111"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var body map[string]string
			if err := json.Unmarshal([]byte(vectors[test.vector].CanonicalJCS), &body); err != nil {
				t.Fatal(err)
			}
			body[test.field] = test.value
			canonical, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ParseActivationDecisionEvidence(canonical); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("error = %v, want ErrInvalidArgument", err)
			}
		})
	}
}

func TestActivationDecisionEvidenceParseRejectsCheckpointWithRollbackTime(t *testing.T) {
	fixture := loadLiteralAuthorityEffectFixture(t)
	var source literalAuthorityEffectVector
	for _, vector := range fixture.Vectors {
		if vector.Name == "evidence-conditional-may-apply" {
			source = vector
			break
		}
	}
	var body map[string]string
	if err := json.Unmarshal([]byte(source.CanonicalJCS), &body); err != nil {
		t.Fatal(err)
	}
	body["checkpoint_kind"] = "node"
	body["checkpoint_scope_digest"] = strings.Repeat("11", 32)
	body["checkpoint_digest"] = "fb1df6952b612b1a4b45bf4e122a1450e26c38130927467244799e9b10e544e3"
	canonical, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseActivationDecisionEvidence(canonical); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("error = %v, want ErrInvalidArgument", err)
	}
}

func TestActivationDecisionEvidenceDefensiveViews(t *testing.T) {
	evidence, proof := freshEvidenceProofFixture(t, "higher_node")
	first := evidence.CanonicalJCS()
	first[0] ^= 1
	if bytes.Equal(first, evidence.CanonicalJCS()) {
		t.Fatal("evidence JCS aliases internal bytes")
	}
	input := proof.Input()
	input.Receipt.DatabasePoint.RequiredLSN = "F/F"
	input.ProviderHead.Facts().LatestCommittedDatabasePoint.RequiredLSN = "F/F"
	input.Checkpoint.canonical[0] ^= 1
	again := proof.Input()
	if again.Receipt.DatabasePoint.RequiredLSN != "1/2" || again.ProviderHead.Facts().LatestCommittedDatabasePoint.RequiredLSN != "1/3" || !bytes.Equal(again.Checkpoint.CanonicalJCS(), evidenceInputFixture(t, "higher_node").Checkpoint.CanonicalJCS()) {
		t.Fatal("proof Input aliases pointer or canonical state")
	}
}

func TestActivationAdmissionCaptureBoundsAndOneShot(t *testing.T) {
	input := evidenceInputFixture(t, "may_apply")
	start := time.Date(2026, 8, 28, 11, 59, 59, 0, time.UTC)
	for _, test := range []struct {
		name  string
		delay time.Duration
		want  error
	}{
		{name: "just under", delay: 4*time.Second - time.Nanosecond},
		{name: "equal", delay: 4 * time.Second, want: ErrConflict},
		{name: "over", delay: 4*time.Second + time.Nanosecond, want: ErrConflict},
	} {
		t.Run(test.name, func(t *testing.T) {
			capture := captureWithReadings(start, start.Add(test.delay))
			_, err := capture.Complete(input)
			if !errors.Is(err, test.want) || (test.want == nil && err != nil) {
				t.Fatalf("Complete error = %v, want %v", err, test.want)
			}
			if _, retryErr := capture.Complete(input); !errors.Is(retryErr, ErrConflict) {
				t.Fatalf("retry error = %v, want ErrConflict", retryErr)
			}
		})
	}

	invalid := evidenceInputFixture(t, "may_apply")
	invalid.Material.ProviderIdentityDigest = contracts.Digest{}
	capture := captureWithReadings(start, start.Add(time.Second))
	if _, err := capture.Complete(invalid); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("first invalid completion = %v", err)
	}
	if _, err := capture.Complete(input); !errors.Is(err, ErrConflict) {
		t.Fatalf("failed completion did not burn capture: %v", err)
	}
}

func TestActivationAdmissionProductionClockRetainsMonotonicBudget(t *testing.T) {
	input := evidenceInputFixture(t, "may_apply")
	capture := BeginActivationEvidenceCapture()
	if _, err := capture.Complete(input); err != nil {
		t.Fatalf("production clock completion = %v", err)
	}
}

func TestActivationAdmissionFreshnessMinBoundsAndOverflow(t *testing.T) {
	base := evidenceInputFixture(t, "may_apply")
	for _, test := range []struct {
		name   string
		mutate func(*ActivationDecisionEvidenceInput)
		valid  bool
	}{
		{name: "attestation equal", valid: true, mutate: func(v *ActivationDecisionEvidenceInput) {
			v.Material.AttestationExpiresAt = v.Material.EvidenceValidUntil
		}},
		{name: "attestation over", mutate: func(v *ActivationDecisionEvidenceInput) {
			v.Material.AttestationExpiresAt = v.Material.TrustedInstant.Add(4 * time.Second)
			v.Material.EvidenceValidUntil = v.Material.AttestationExpiresAt.Add(time.Nanosecond)
		}},
		{name: "deadline equal", valid: true, mutate: func(v *ActivationDecisionEvidenceInput) {
			v.Material.ActivationDeadline = v.Material.EvidenceValidUntil
		}},
		{name: "deadline over", mutate: func(v *ActivationDecisionEvidenceInput) {
			v.Material.ActivationDeadline = v.Material.TrustedInstant.Add(4 * time.Second)
			v.Material.EvidenceValidUntil = v.Material.ActivationDeadline.Add(time.Nanosecond)
		}},
		{name: "five seconds equal", valid: true, mutate: func(v *ActivationDecisionEvidenceInput) {
			v.Material.AttestationExpiresAt = v.Material.TrustedInstant.Add(10 * time.Second)
			v.Material.ActivationDeadline = v.Material.TrustedInstant.Add(10 * time.Second)
			v.Material.EvidenceValidUntil = v.Material.TrustedInstant.Add(5 * time.Second)
		}},
		{name: "five seconds over", mutate: func(v *ActivationDecisionEvidenceInput) {
			v.Material.AttestationExpiresAt = v.Material.TrustedInstant.Add(10 * time.Second)
			v.Material.ActivationDeadline = v.Material.TrustedInstant.Add(10 * time.Second)
			v.Material.EvidenceValidUntil = v.Material.TrustedInstant.Add(5*time.Second + time.Nanosecond)
		}},
		{name: "trusted equals valid", mutate: func(v *ActivationDecisionEvidenceInput) { v.Material.EvidenceValidUntil = v.Material.TrustedInstant }},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := base
			test.mutate(&input)
			_, _, err := completeAndValidateEvidence(t, input, 0)
			if test.valid && err != nil {
				t.Fatalf("valid bound: %v", err)
			}
			if !test.valid && !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("invalid bound error = %v", err)
			}
		})
	}

	overflow := base
	overflow.Material.TrustedInstant = time.Date(9999, 12, 31, 23, 59, 58, 0, time.UTC)
	overflow.Material.EvidenceValidUntil = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	overflow.Material.AttestationExpiresAt = overflow.Material.EvidenceValidUntil
	overflow.Material.ActivationDeadline = overflow.Material.EvidenceValidUntil
	if _, _, err := completeAndValidateEvidence(t, overflow, 0); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("overflow error = %v, want ErrInvalidArgument", err)
	}
}

func TestActivationAdmissionCopyConcurrentAndFailedConsumeBurns(t *testing.T) {
	t.Run("copy and concurrent", func(t *testing.T) {
		_, proof := freshEvidenceProofFixture(t, "may_apply")
		const contenders = 32
		var successes atomic.Int32
		var conflicts atomic.Int32
		var wait sync.WaitGroup
		wait.Add(contenders)
		for range contenders {
			copyProof := proof
			go func() {
				defer wait.Done()
				err := consumeActivationAdmission(context.Background(), copyProof)
				switch {
				case err == nil:
					successes.Add(1)
				case errors.Is(err, ErrConflict):
					conflicts.Add(1)
				default:
					t.Errorf("consume error = %v", err)
				}
			}()
		}
		wait.Wait()
		if successes.Load() != 1 || conflicts.Load() != contenders-1 {
			t.Fatalf("successes=%d conflicts=%d", successes.Load(), conflicts.Load())
		}
	})

	t.Run("cancellation burns", func(t *testing.T) {
		_, proof := freshEvidenceProofFixture(t, "may_apply")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := consumeActivationAdmission(ctx, proof); !errors.Is(err, ErrCanceled) {
			t.Fatalf("canceled consume = %v", err)
		}
		if err := consumeActivationAdmission(context.Background(), proof); !errors.Is(err, ErrConflict) {
			t.Fatalf("canceled retry = %v", err)
		}
	})

	t.Run("expiry burns", func(t *testing.T) {
		input := evidenceInputFixture(t, "may_apply")
		start := time.Date(2026, 8, 28, 11, 59, 59, 0, time.UTC)
		capture := captureWithReadings(start, start.Add(time.Second), start.Add(4*time.Second))
		evidence, err := capture.Complete(input)
		if err != nil {
			t.Fatal(err)
		}
		proof, err := ValidateActivationDecisionEvidence(evidence, input)
		if err != nil {
			t.Fatal(err)
		}
		if err := consumeActivationAdmission(context.Background(), proof); !errors.Is(err, ErrConflict) {
			t.Fatalf("expired consume = %v", err)
		}
		if err := consumeActivationAdmission(context.Background(), proof); !errors.Is(err, ErrConflict) {
			t.Fatalf("expired retry = %v", err)
		}
	})

	t.Run("binding mismatch burns", func(t *testing.T) {
		_, proof := freshEvidenceProofFixture(t, "may_apply")
		proof.admission.operationID = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
		if err := consumeActivationAdmission(context.Background(), proof); !errors.Is(err, ErrConflict) {
			t.Fatalf("binding consume = %v", err)
		}
		if err := consumeActivationAdmission(context.Background(), proof); !errors.Is(err, ErrConflict) {
			t.Fatalf("binding retry = %v", err)
		}
	})
}

func TestActivationEvidenceCapturePreUseCopiesShareOneClaim(t *testing.T) {
	input := evidenceInputFixture(t, "may_apply")
	start := time.Date(2026, 8, 28, 11, 59, 59, 0, time.UTC)

	t.Run("sequential copy", func(t *testing.T) {
		capture := captureWithReadings(start, start.Add(time.Second))
		copyCapture := *capture
		if _, err := capture.Complete(input); err != nil {
			t.Fatalf("first completion = %v", err)
		}
		if _, err := copyCapture.Complete(input); !errors.Is(err, ErrConflict) {
			t.Fatalf("copied completion = %v, want ErrConflict", err)
		}
	})

	t.Run("concurrent copies", func(t *testing.T) {
		capture := captureWithReadings(start, start.Add(time.Second))
		const contenders = 32
		copies := make([]ActivationEvidenceCapture, contenders)
		for index := range copies {
			copies[index] = *capture
		}
		var successes atomic.Int32
		var conflicts atomic.Int32
		var wait sync.WaitGroup
		wait.Add(contenders)
		for index := range copies {
			go func(candidate *ActivationEvidenceCapture) {
				defer wait.Done()
				_, err := candidate.Complete(input)
				switch {
				case err == nil:
					successes.Add(1)
				case errors.Is(err, ErrConflict):
					conflicts.Add(1)
				default:
					t.Errorf("completion = %v", err)
				}
			}(&copies[index])
		}
		wait.Wait()
		if successes.Load() != 1 || conflicts.Load() != contenders-1 {
			t.Fatalf("successes=%d conflicts=%d", successes.Load(), conflicts.Load())
		}
	})

	t.Run("invalid completion burns every copy", func(t *testing.T) {
		capture := captureWithReadings(start, start.Add(time.Second))
		copyCapture := *capture
		invalid := input
		invalid.Material.ProviderIdentityDigest = contracts.Digest{}
		if _, err := copyCapture.Complete(invalid); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("invalid completion = %v, want ErrInvalidArgument", err)
		}
		if _, err := capture.Complete(input); !errors.Is(err, ErrConflict) {
			t.Fatalf("retry through original = %v, want ErrConflict", err)
		}
	})
}

func TestAuthorityEffectResolutionClosedMatrix(t *testing.T) {
	for _, scenario := range []string{"applied", "exact_capture", "deadline_expired", "final", "higher_node"} {
		scenario := scenario
		t.Run(scenario, func(t *testing.T) {
			resolution := resolutionFixture(t, scenario)
			parsed, err := ParseAuthorityEffectResolution(resolution.CanonicalJCS())
			if err != nil {
				t.Fatal(err)
			}
			evidenceScenario := scenario
			if scenario == "applied" {
				evidenceScenario = "may_apply"
			}
			evidence, _ := freshEvidenceProofFixture(t, evidenceScenario)
			parsedEvidence, err := ParseActivationDecisionEvidence(evidence.CanonicalJCS())
			if err != nil {
				t.Fatal(err)
			}
			input := evidenceInputFixture(t, evidenceScenario)
			proof, err := ValidateActivationDecisionEvidence(parsedEvidence, input)
			if err != nil {
				t.Fatal(err)
			}
			if err := ValidateAuthorityEffectResolution(parsed, input.Material.Commitment, proof); err != nil {
				t.Fatalf("context validation: %v", err)
			}
		})
	}

	resolution := resolutionFixture(t, "applied")
	evidence, _ := freshEvidenceProofFixture(t, "may_apply")
	parsedEvidence, _ := ParseActivationDecisionEvidence(evidence.CanonicalJCS())
	wrongInput := evidenceInputFixture(t, "final")
	wrongProof, err := ValidateActivationDecisionEvidence(parsedEvidence, evidenceInputFixture(t, "may_apply"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateAuthorityEffectResolution(resolution, wrongInput.Material.Commitment, wrongProof); !errors.Is(err, ErrConflict) {
		t.Fatalf("cross commitment error = %v, want ErrConflict", err)
	}
}

func TestAuthorityEffectResolutionParseRejectsImpossiblePairs(t *testing.T) {
	fixture := loadLiteralAuthorityEffectFixture(t)
	vectors := make(map[string]literalAuthorityEffectVector, len(fixture.Vectors))
	for _, vector := range fixture.Vectors {
		vectors[vector.Name] = vector
	}
	for _, test := range []struct {
		name   string
		vector string
		field  string
		value  string
	}{
		{name: "failed with trusted time", vector: "resolution-final-not-applied", field: "decision_anchor_kind", value: "trusted_time"},
		{name: "superseded with exact capture", vector: "resolution-exact-capture", field: "reason_code", value: "superseded"},
		{name: "failed with higher authority", vector: "resolution-higher-authority", field: "reason_code", value: "failed"},
		{name: "superseded with trusted time", vector: "resolution-deadline-expired", field: "reason_code", value: "superseded"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var body map[string]string
			if err := json.Unmarshal([]byte(vectors[test.vector].CanonicalJCS), &body); err != nil {
				t.Fatal(err)
			}
			body[test.field] = test.value
			canonical, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ParseAuthorityEffectResolution(canonical); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("error = %v, want ErrInvalidArgument", err)
			}
		})
	}
}

func completeAndValidateEvidence(t *testing.T, input ActivationDecisionEvidenceInput, completeDelay time.Duration) (ActivationDecisionEvidence, ValidatedActivationDecisionEvidence, error) {
	t.Helper()
	var evidence ActivationDecisionEvidence
	var err error
	if input.Material.TrustedTimeKind == TrustedTimeNone {
		evidence, err = NewActivationDecisionEvidence(input)
	} else {
		start := time.Date(2026, 8, 28, 11, 59, 59, 0, time.UTC)
		capture := captureWithReadings(start, start.Add(time.Second+completeDelay))
		evidence, err = capture.Complete(input)
	}
	if err != nil {
		return ActivationDecisionEvidence{}, ValidatedActivationDecisionEvidence{}, err
	}
	proof, err := ValidateActivationDecisionEvidence(evidence, input)
	return evidence, proof, err
}

func captureWithReadings(readings ...time.Time) *ActivationEvidenceCapture {
	var index atomic.Int32
	return beginActivationEvidenceCaptureForTest(func() time.Time {
		current := int(index.Add(1)) - 1
		if current >= len(readings) {
			return readings[len(readings)-1]
		}
		return readings[current]
	})
}

func clearEvidenceTime(input *ActivationDecisionEvidenceInput) {
	input.Material.TrustedTimeKind = TrustedTimeNone
	input.Material.TrustedInstant = time.Time{}
	input.Material.EvidenceValidUntil = time.Time{}
	input.Material.AttestationExpiresAt = time.Time{}
	input.Material.ActivationDeadline = time.Time{}
	input.Material.ProviderIdentityDigest = contracts.Digest{}
	input.Material.ExpectedProviderIdentityDigest = contracts.Digest{}
	input.Material.FloorAttestationDigest = contracts.Digest{}
}

func snapshotFromHead(t *testing.T, head Head) AuthorityProviderHeadSnapshot {
	t.Helper()
	snapshot, err := NewAuthorityProviderHeadSnapshot(head)
	if err != nil {
		t.Fatalf("head snapshot: %v", err)
	}
	return snapshot
}

func mutateHead(head Head, mutate func(*Head)) Head {
	point := *head.LatestCommittedDatabasePoint
	head.LatestCommittedDatabasePoint = &point
	mutate(&head)
	return head
}

func replaceCheckpoint(t *testing.T, input *ActivationDecisionEvidenceInput, epoch, sequence uint64, scope contracts.Digest) {
	t.Helper()
	anchor, err := NewAuthorityCheckpointAnchor(AuthorityCheckpointAnchorInput{
		Kind: CheckpointNode, ScopeDigest: scope,
		Checkpoint: NodeCheckpoint{AuthorityEpoch: epoch, Sequence: sequence, ReceiptDigest: mustAuthorityDigest(t, "77"+string(bytes.Repeat([]byte("77"), 31)))},
	})
	if err != nil {
		input.Checkpoint = &AuthorityCheckpointAnchor{}
		return
	}
	input.Material.CheckpointScopeDigest = scope
	input.Checkpoint = &anchor
}

func globalHigherAuthorityInput(t *testing.T) ActivationDecisionEvidenceInput {
	t.Helper()
	commitment, err := NewAuthorityEffectCommitment(AuthorityEffectCommitmentInput{
		OperationID: uuid.MustParse("44444444-4444-4444-8444-444444444444"), Kind: EffectRootPublish,
		ScopeKind: ScopeGlobalNodeTrust, ScopeDigest: globalNodeScopeDigest, Epoch: 7, Sequence: 21,
		BaseEffectDigest: mustAuthorityDigest(t, "aa"+string(bytes.Repeat([]byte("aa"), 31))),
		Mode:             CommitmentConditionalApply, Reason: EffectReasonNone, ActivationPolicyVersion: 4,
		ActivationInputsDigest: mustAuthorityDigest(t, "bb"+string(bytes.Repeat([]byte("bb"), 31))),
	})
	if err != nil {
		t.Fatal(err)
	}
	reservation := Reservation{
		OperationID: commitment.Facts().OperationID, Kind: EffectRootPublish, ScopeKind: ScopeGlobalNodeTrust,
		ScopeDigest: globalNodeScopeDigest, Epoch: 7, Sequence: 21,
	}
	reservation.ReservationDigest, _ = reservationDigest(reservation)
	point := DatabasePoint{SystemID: math.MaxUint64, Timeline: math.MaxUint32, RequiredLSN: "2/A"}
	effectDigest := commitment.Digest()
	receipt := Receipt{Reservation: reservation, EffectDigest: &effectDigest, DatabasePoint: &point, Status: StatusCommitted}
	receipt.ReceiptDigest, _ = receiptDigest(receipt)
	checkpoint, err := NewAuthorityCheckpointAnchor(AuthorityCheckpointAnchorInput{
		Kind: CheckpointGlobal, ScopeDigest: globalNodeScopeDigest,
		Checkpoint: NodeCheckpoint{AuthorityEpoch: 7, Sequence: 22, ReceiptDigest: mustAuthorityDigest(t, "cc"+string(bytes.Repeat([]byte("cc"), 31)))},
	})
	if err != nil {
		t.Fatal(err)
	}
	head := Head{
		Epoch: 7, LatestReservedSequence: 22, LatestCommittedSequence: 22,
		LatestReservationDigest:      mustAuthorityDigest(t, "dd"+string(bytes.Repeat([]byte("dd"), 31))),
		LatestCommittedOperationID:   uuid.MustParse("55555555-5555-4555-8555-555555555555"),
		LatestCommittedReceiptDigest: checkpoint.Facts().ReceiptDigest,
		LatestCommittedDatabasePoint: &DatabasePoint{SystemID: math.MaxUint64, Timeline: math.MaxUint32, RequiredLSN: "2/B"},
	}
	return ActivationDecisionEvidenceInput{
		Material: ActivationDecisionMaterial{
			Commitment: commitment, Reason: EffectReasonSuperseded, CheckpointKind: CheckpointGlobal,
			CheckpointScopeDigest: globalNodeScopeDigest, TrustedTimeKind: TrustedTimeNone,
			Capability: DecisionCapabilityNotAppliedOnly,
		},
		Receipt: receipt, ProviderHead: snapshotFromHead(t, head), Checkpoint: &checkpoint,
	}
}
