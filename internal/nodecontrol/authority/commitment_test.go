package authority

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/strictjson"
)

type literalAuthorityEffectFixture struct {
	SchemaVersion    string                           `json:"schema_version"`
	Vectors          []literalAuthorityEffectVector   `json:"vectors"`
	MutationManifest []literalAuthorityEffectMutation `json:"mutation_manifest"`
}

type literalAuthorityEffectVector struct {
	Name         string          `json:"name"`
	ArtifactKind string          `json:"artifact_kind"`
	Input        json.RawMessage `json:"input"`
	CanonicalJCS string          `json:"canonical_jcs"`
	DigestHex    string          `json:"digest_hex"`
}

type literalAuthorityEffectMutation struct {
	Name            string `json:"name"`
	ArtifactKind    string `json:"artifact_kind"`
	BaseVector      string `json:"base_vector,omitempty"`
	Field           string `json:"field"`
	Mutation        string `json:"mutation"`
	Classification  string `json:"classification"`
	Sentinel        string `json:"sentinel,omitempty"`
	AlternateVector string `json:"alternate_vector,omitempty"`
}

type literalCommitmentInput struct {
	OperationID             string                `json:"operation_id"`
	Kind                    EffectKind            `json:"kind"`
	ScopeKind               ScopeKind             `json:"scope_kind"`
	ScopeDigest             string                `json:"scope_digest"`
	Epoch                   uint64                `json:"epoch"`
	Sequence                uint64                `json:"sequence"`
	BaseEffectDigest        string                `json:"base_effect_digest"`
	Mode                    CommitmentMode        `json:"mode"`
	Reason                  AuthorityEffectReason `json:"reason"`
	ActivationPolicyVersion uint64                `json:"activation_policy_version"`
	ActivationInputsDigest  string                `json:"activation_inputs_digest"`
}

type literalProviderHeadInput struct {
	Epoch                        uint64 `json:"epoch"`
	LatestReservedSequence       uint64 `json:"latest_reserved_sequence"`
	LatestCommittedSequence      uint64 `json:"latest_committed_sequence"`
	LatestReservationDigest      string `json:"latest_reservation_digest"`
	LatestCommittedOperationID   string `json:"latest_committed_operation_id"`
	LatestCommittedReceiptDigest string `json:"latest_committed_receipt_digest"`
	DBSystemID                   uint64 `json:"db_system_id"`
	DBTimeline                   uint32 `json:"db_timeline"`
	RequiredLSN                  string `json:"required_lsn"`
}

type literalCheckpointInput struct {
	Kind          CheckpointKind `json:"kind"`
	ScopeDigest   string         `json:"scope_digest"`
	Epoch         uint64         `json:"epoch"`
	Sequence      uint64         `json:"sequence"`
	ReceiptDigest string         `json:"receipt_digest"`
}

type literalReceiptInput struct {
	OperationID       string        `json:"operation_id"`
	Kind              EffectKind    `json:"kind"`
	ScopeKind         ScopeKind     `json:"scope_kind"`
	ScopeDigest       string        `json:"scope_digest"`
	Epoch             uint64        `json:"epoch"`
	Sequence          uint64        `json:"sequence"`
	ReservationDigest string        `json:"reservation_digest"`
	EffectDigest      string        `json:"effect_digest"`
	DBSystemID        uint64        `json:"db_system_id"`
	DBTimeline        uint32        `json:"db_timeline"`
	RequiredLSN       string        `json:"required_lsn"`
	Status            ReceiptStatus `json:"status"`
	ReceiptDigest     string        `json:"receipt_digest"`
}

type literalEvidenceMaterialInput struct {
	Commitment                     literalCommitmentInput `json:"commitment"`
	Reason                         AuthorityEffectReason  `json:"reason"`
	CheckpointKind                 CheckpointKind         `json:"checkpoint_kind"`
	CheckpointScopeDigest          string                 `json:"checkpoint_scope_digest"`
	TrustedTimeKind                TrustedTimeKind        `json:"trusted_time_kind"`
	TrustedInstant                 string                 `json:"trusted_instant"`
	EvidenceValidUntil             string                 `json:"evidence_valid_until"`
	AttestationExpiresAt           string                 `json:"attestation_expires_at"`
	ActivationDeadline             string                 `json:"activation_deadline"`
	ProviderIdentityDigest         string                 `json:"provider_identity_digest"`
	ExpectedProviderIdentityDigest string                 `json:"expected_provider_identity_digest"`
	FloorAttestationDigest         string                 `json:"floor_attestation_digest"`
	Capability                     DecisionCapability     `json:"capability"`
}

type literalEvidenceInput struct {
	Material     literalEvidenceMaterialInput `json:"material"`
	Receipt      literalReceiptInput          `json:"receipt"`
	ProviderHead literalProviderHeadInput     `json:"provider_head"`
	Checkpoint   *literalCheckpointInput      `json:"checkpoint"`
}

type literalResolutionInput struct {
	Commitment   literalCommitmentInput `json:"commitment"`
	Evidence     literalEvidenceInput   `json:"evidence"`
	Disposition  EffectDisposition      `json:"disposition"`
	Reason       AuthorityEffectReason  `json:"reason"`
	AnchorKind   DecisionAnchorKind     `json:"anchor_kind"`
	AnchorDigest string                 `json:"anchor_digest"`
}

func TestAuthorityEffectLiteralInputsAreSemantic(t *testing.T) {
	fixture := loadLiteralAuthorityEffectFixture(t)
	required := map[string][]string{
		"commitment":    {"operation_id", "kind", "scope_kind", "scope_digest", "epoch", "sequence", "base_effect_digest", "mode", "reason", "activation_policy_version", "activation_inputs_digest"},
		"provider_head": {"epoch", "latest_reserved_sequence", "latest_committed_sequence", "latest_reservation_digest", "latest_committed_operation_id", "latest_committed_receipt_digest", "db_system_id", "db_timeline", "required_lsn"},
		"checkpoint":    {"kind", "scope_digest", "epoch", "sequence", "receipt_digest"},
		"evidence":      {"material", "receipt", "provider_head", "checkpoint"},
		"resolution":    {"commitment", "evidence", "disposition", "reason", "anchor_kind", "anchor_digest"},
	}
	for _, vector := range fixture.Vectors {
		var input map[string]json.RawMessage
		if err := strictjson.Decode(bytes.NewReader(vector.Input), 4096, &input); err != nil {
			t.Fatalf("%s input: %v", vector.Name, err)
		}
		for _, field := range required[vector.ArtifactKind] {
			if _, exists := input[field]; !exists {
				t.Errorf("%s semantic input missing %s", vector.Name, field)
			}
		}
	}
}

func TestAuthorityEffectCompositeLiteralInputsAreSelfContained(t *testing.T) {
	fixture := loadLiteralAuthorityEffectFixture(t)
	for _, vector := range fixture.Vectors {
		var input map[string]json.RawMessage
		if err := strictjson.Decode(bytes.NewReader(vector.Input), 4096, &input); err != nil {
			t.Fatalf("%s input: %v", vector.Name, err)
		}
		switch vector.ArtifactKind {
		case "evidence":
			for _, field := range []string{"material", "receipt", "provider_head", "checkpoint"} {
				if !literalCompositeField(input[field], field == "checkpoint") {
					t.Errorf("%s %s is not a nested literal preimage", vector.Name, field)
				}
			}
		case "resolution":
			for _, field := range []string{"commitment", "evidence"} {
				if !literalCompositeField(input[field], false) {
					t.Errorf("%s %s is not a nested literal preimage", vector.Name, field)
				}
			}
		}
	}
}

func literalCompositeField(value json.RawMessage, nullable bool) bool {
	value = bytes.TrimSpace(value)
	return len(value) > 1 && (value[0] == '{' || nullable && bytes.Equal(value, []byte("null")))
}

func TestAuthorityEffectCanonicalVectors(t *testing.T) {
	fixture := loadLiteralAuthorityEffectFixture(t)
	if got := hex.EncodeToString(AuthorityEffectActivationInputsEmptyV1[:]); got != "eafc8658c2b1ebcfce129176ac4e10ce698e5dc6fc012012184d604e6e9b69aa" {
		t.Fatalf("empty activation-input digest = %s", got)
	}

	for _, vector := range fixture.Vectors {
		vector := vector
		t.Run(vector.Name, func(t *testing.T) {
			gotJCS, gotDigest := constructLiteralAuthorityArtifact(t, vector)
			if !bytes.Equal(gotJCS, []byte(vector.CanonicalJCS)) {
				t.Fatalf("canonical JCS = %s, want literal %s", gotJCS, vector.CanonicalJCS)
			}
			if got := hex.EncodeToString(gotDigest[:]); got != vector.DigestHex {
				t.Fatalf("digest = %s, want literal %s", got, vector.DigestHex)
			}
			parsedJCS, parsedDigest, err := parseLiteralAuthorityArtifact(vector.ArtifactKind, gotJCS)
			if err != nil {
				t.Fatalf("parse exact JCS: %v", err)
			}
			if !bytes.Equal(parsedJCS, gotJCS) || parsedDigest != gotDigest {
				t.Fatal("parse round trip changed canonical value")
			}
			if len(gotJCS) > 0 {
				gotJCS[0] ^= 1
				again, _, err := parseLiteralAuthorityArtifact(vector.ArtifactKind, []byte(vector.CanonicalJCS))
				if err != nil || !bytes.Equal(again, []byte(vector.CanonicalJCS)) {
					t.Fatal("canonical accessor aliases caller bytes")
				}
			}
		})
	}
}

func TestAuthorityCanonicalMutationManifest(t *testing.T) {
	fixture := loadLiteralAuthorityEffectFixture(t)
	vectors := make(map[string]literalAuthorityEffectVector, len(fixture.Vectors))
	firstByKind := make(map[string]literalAuthorityEffectVector)
	for _, vector := range fixture.Vectors {
		vectors[vector.Name] = vector
		if _, exists := firstByKind[vector.ArtifactKind]; !exists {
			firstByKind[vector.ArtifactKind] = vector
		}
	}

	for _, mutation := range fixture.MutationManifest {
		mutation := mutation
		t.Run(mutation.Name, func(t *testing.T) {
			base, exists := firstByKind[mutation.ArtifactKind]
			if mutation.BaseVector != "" {
				base, exists = vectors[mutation.BaseVector]
			}
			if !exists {
				t.Fatalf("unknown artifact kind %q", mutation.ArtifactKind)
			}
			if mutation.Classification == "valid_alternate" {
				alternate, exists := vectors[mutation.AlternateVector]
				if !exists || alternate.ArtifactKind != mutation.ArtifactKind || alternate.CanonicalJCS == base.CanonicalJCS || alternate.DigestHex == base.DigestHex {
					t.Fatal("valid alternate lacks a literal, different JCS/digest")
				}
				if _, _, err := parseLiteralAuthorityArtifact(alternate.ArtifactKind, []byte(alternate.CanonicalJCS)); err != nil {
					t.Fatalf("valid alternate rejected: %v", err)
				}
				return
			}
			if mutation.Classification != "reject" || mutation.Sentinel != "invalid_argument" {
				t.Fatal("mutation is not classified with an exact finite sentinel")
			}
			var body map[string]string
			if err := json.Unmarshal([]byte(base.CanonicalJCS), &body); err != nil {
				t.Fatal(err)
			}
			body[mutation.Field] = mutation.Mutation
			mutated, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := parseLiteralAuthorityArtifact(mutation.ArtifactKind, mutated); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("mutation error = %v, want ErrInvalidArgument", err)
			}
		})
	}
}

func TestAuthorityEffectParsersRejectAmbiguousBodies(t *testing.T) {
	fixture := loadLiteralAuthorityEffectFixture(t)
	firstByKind := make(map[string]literalAuthorityEffectVector)
	for _, vector := range fixture.Vectors {
		if _, exists := firstByKind[vector.ArtifactKind]; !exists {
			firstByKind[vector.ArtifactKind] = vector
		}
	}
	for kind, vector := range firstByKind {
		kind, vector := kind, vector
		t.Run(kind, func(t *testing.T) {
			base := vector.CanonicalJCS
			closing := strings.LastIndexByte(base, '}')
			firstFieldEnd := strings.IndexByte(base[2:], '"') + 2
			firstField := base[2:firstFieldEnd]
			valueStart := firstFieldEnd + len(`":"`)
			valueEnd := strings.IndexByte(base[valueStart:], '"') + valueStart
			firstValue := base[valueStart:valueEnd]
			cases := [][]byte{
				nil,
				bytes.Repeat([]byte{'x'}, 4097),
				[]byte(" " + base),
				[]byte(base[:closing] + `,"extra":"x"}`),
				[]byte(base[:closing] + `,"` + firstField + `":"` + firstValue + `"}`),
				[]byte(strings.Replace(base, `"`+firstField+`":"`+firstValue+`"`, `"`+firstField+`":null`, 1)),
				[]byte(strings.Replace(base, `"`+firstField+`":"`+firstValue+`"`, `"`+firstField+`":1`, 1)),
			}
			for index, body := range cases {
				if _, _, err := parseLiteralAuthorityArtifact(kind, body); !errors.Is(err, ErrInvalidArgument) {
					t.Fatalf("case %d error = %v, want ErrInvalidArgument", index, err)
				}
			}
		})
	}
}

func TestAuthorityEffectCommitmentRejectsFixedUnsupportedKinds(t *testing.T) {
	base := commitmentFixture(t, "conditional").Facts()
	for _, kind := range []EffectKind{EffectTrustBundlePublish, EffectOperatorAuthorizerChange} {
		input := AuthorityEffectCommitmentInput{
			OperationID: base.OperationID, Kind: kind, ScopeKind: base.ScopeKind, ScopeDigest: base.ScopeDigest,
			Epoch: base.Epoch, Sequence: base.Sequence, BaseEffectDigest: base.BaseEffectDigest,
			Mode: base.Mode, Reason: base.Reason, ActivationPolicyVersion: base.ActivationPolicyVersion,
			ActivationInputsDigest: base.ActivationInputsDigest,
		}
		if kind == EffectTrustBundlePublish {
			input.ScopeKind = ScopeGlobalNodeTrust
			input.ScopeDigest = globalNodeScopeDigest
		} else {
			input.ScopeKind = ScopeGlobalOperatorTrust
			input.ScopeDigest = globalOperatorScopeDigest
		}
		if _, err := NewAuthorityEffectCommitment(input); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("kind %s error = %v, want ErrInvalidArgument", kind, err)
		}
	}
}

func FuzzAuthorityEffectCanonical(f *testing.F) {
	fixture := loadLiteralAuthorityEffectFixture(f)
	for _, vector := range fixture.Vectors {
		f.Add(vector.ArtifactKind, []byte(vector.CanonicalJCS))
	}
	f.Fuzz(func(t *testing.T, kind string, body []byte) {
		jcs, digest, err := parseLiteralAuthorityArtifact(kind, body)
		if err != nil {
			return
		}
		if len(jcs) > 0 {
			jcs[0] ^= 1
		}
		again, againDigest, err := parseLiteralAuthorityArtifact(kind, body)
		if err != nil || againDigest != digest || bytes.Equal(jcs, again) {
			t.Fatal("canonical parse aliases returned bytes")
		}
	})
}

func FuzzAuthorityEffectParsers(f *testing.F) {
	fixture := loadLiteralAuthorityEffectFixture(f)
	for _, vector := range fixture.Vectors {
		f.Add([]byte(vector.CanonicalJCS))
	}
	f.Add([]byte(`{"schema_version":"1"}`))
	f.Fuzz(func(t *testing.T, body []byte) {
		_, _ = ParseAuthorityEffectCommitment(body)
		_, _ = ParseAuthorityProviderHeadSnapshot(body)
		_, _ = ParseAuthorityCheckpointAnchor(body)
		_, _ = ParseActivationDecisionEvidence(body)
		_, _ = ParseAuthorityEffectResolution(body)
	})
}

func loadLiteralAuthorityEffectFixture(tb testing.TB) literalAuthorityEffectFixture {
	tb.Helper()
	body, err := os.ReadFile("testdata/authority-effect-vectors.json")
	if err != nil {
		tb.Fatal(err)
	}
	var fixture literalAuthorityEffectFixture
	if err := strictjson.Decode(bytes.NewReader(body), 64<<10, &fixture); err != nil {
		tb.Fatalf("strict vector fixture: %v", err)
	}
	if fixture.SchemaVersion != "1" || len(fixture.Vectors) == 0 || len(fixture.MutationManifest) == 0 {
		tb.Fatal("incomplete vector fixture")
	}
	return fixture
}

func constructLiteralAuthorityArtifact(t *testing.T, vector literalAuthorityEffectVector) ([]byte, contracts.Digest) {
	t.Helper()
	switch vector.ArtifactKind {
	case "commitment":
		var input literalCommitmentInput
		decodeLiteralInput(t, vector.Input, &input)
		value := commitmentFromLiteralInput(t, input)
		return value.CanonicalJCS(), value.Digest()
	case "provider_head":
		var input literalProviderHeadInput
		decodeLiteralInput(t, vector.Input, &input)
		value := providerHeadFromLiteralInput(t, input)
		return value.CanonicalJCS(), value.Digest()
	case "checkpoint":
		var input literalCheckpointInput
		decodeLiteralInput(t, vector.Input, &input)
		value := checkpointFromLiteralInput(t, input)
		return value.CanonicalJCS(), value.Digest()
	case "evidence":
		var input literalEvidenceInput
		decodeLiteralInput(t, vector.Input, &input)
		value, _ := freshEvidenceProofFromLiteralInput(t, input)
		return value.CanonicalJCS(), value.Digest()
	case "resolution":
		var input literalResolutionInput
		decodeLiteralInput(t, vector.Input, &input)
		value := resolutionFromLiteralInput(t, input)
		return value.CanonicalJCS(), value.Digest()
	default:
		t.Fatalf("unknown artifact kind %q", vector.ArtifactKind)
		return nil, contracts.Digest{}
	}
}

func decodeLiteralInput(t *testing.T, body []byte, destination any) {
	t.Helper()
	if err := strictjson.Decode(bytes.NewReader(body), 4096, destination); err != nil {
		t.Fatalf("vector input: %v", err)
	}
}

func mustOptionalAuthorityDigest(t *testing.T, value string) contracts.Digest {
	t.Helper()
	if value == "" {
		return contracts.Digest{}
	}
	return mustAuthorityDigest(t, value)
}

func parseLiteralInstant(t *testing.T, value string) time.Time {
	t.Helper()
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || parsed.Format(time.RFC3339Nano) != value {
		t.Fatalf("literal instant %q: %v", value, err)
	}
	return parsed
}

func commitmentFromLiteralInput(t *testing.T, input literalCommitmentInput) AuthorityEffectCommitment {
	t.Helper()
	value, err := NewAuthorityEffectCommitment(AuthorityEffectCommitmentInput{
		OperationID: uuid.MustParse(input.OperationID), Kind: input.Kind, ScopeKind: input.ScopeKind,
		ScopeDigest: mustOptionalAuthorityDigest(t, input.ScopeDigest), Epoch: input.Epoch, Sequence: input.Sequence,
		BaseEffectDigest: mustOptionalAuthorityDigest(t, input.BaseEffectDigest), Mode: input.Mode, Reason: input.Reason,
		ActivationPolicyVersion: input.ActivationPolicyVersion,
		ActivationInputsDigest:  mustOptionalAuthorityDigest(t, input.ActivationInputsDigest),
	})
	if err != nil {
		t.Fatalf("commitment: %v", err)
	}
	return value
}

func providerHeadFromLiteralInput(t *testing.T, input literalProviderHeadInput) AuthorityProviderHeadSnapshot {
	t.Helper()
	value, err := NewAuthorityProviderHeadSnapshot(Head{
		Epoch: input.Epoch, LatestReservedSequence: input.LatestReservedSequence,
		LatestCommittedSequence:      input.LatestCommittedSequence,
		LatestReservationDigest:      mustAuthorityDigest(t, input.LatestReservationDigest),
		LatestCommittedOperationID:   uuid.MustParse(input.LatestCommittedOperationID),
		LatestCommittedReceiptDigest: mustAuthorityDigest(t, input.LatestCommittedReceiptDigest),
		LatestCommittedDatabasePoint: &DatabasePoint{SystemID: input.DBSystemID, Timeline: input.DBTimeline, RequiredLSN: WALPosition(input.RequiredLSN)},
	})
	if err != nil {
		t.Fatalf("provider head: %v", err)
	}
	return value
}

func checkpointFromLiteralInput(t *testing.T, input literalCheckpointInput) AuthorityCheckpointAnchor {
	t.Helper()
	value, err := NewAuthorityCheckpointAnchor(AuthorityCheckpointAnchorInput{
		Kind: input.Kind, ScopeDigest: mustAuthorityDigest(t, input.ScopeDigest),
		Checkpoint: NodeCheckpoint{AuthorityEpoch: input.Epoch, Sequence: input.Sequence, ReceiptDigest: mustAuthorityDigest(t, input.ReceiptDigest)},
	})
	if err != nil {
		t.Fatalf("checkpoint: %v", err)
	}
	return value
}

func receiptFromLiteralInput(t *testing.T, input literalReceiptInput) Receipt {
	t.Helper()
	effectDigest := mustAuthorityDigest(t, input.EffectDigest)
	point := DatabasePoint{SystemID: input.DBSystemID, Timeline: input.DBTimeline, RequiredLSN: WALPosition(input.RequiredLSN)}
	value := Receipt{
		Reservation: Reservation{
			OperationID: uuid.MustParse(input.OperationID), Kind: input.Kind, ScopeKind: input.ScopeKind,
			ScopeDigest: mustAuthorityDigest(t, input.ScopeDigest), Epoch: input.Epoch, Sequence: input.Sequence,
			ReservationDigest: mustAuthorityDigest(t, input.ReservationDigest),
		},
		EffectDigest: &effectDigest, DatabasePoint: &point, Status: input.Status,
		ReceiptDigest: mustAuthorityDigest(t, input.ReceiptDigest),
	}
	if err := value.Validate(); err != nil {
		t.Fatalf("receipt: %v", err)
	}
	return value
}

func freshEvidenceProofFromLiteralInput(t *testing.T, input literalEvidenceInput) (ActivationDecisionEvidence, ValidatedActivationDecisionEvidence) {
	t.Helper()
	materialInput := input.Material
	material := ActivationDecisionMaterial{
		Commitment: commitmentFromLiteralInput(t, materialInput.Commitment), Reason: materialInput.Reason,
		CheckpointKind: materialInput.CheckpointKind, CheckpointScopeDigest: mustOptionalAuthorityDigest(t, materialInput.CheckpointScopeDigest),
		TrustedTimeKind: materialInput.TrustedTimeKind, TrustedInstant: parseLiteralInstant(t, materialInput.TrustedInstant),
		EvidenceValidUntil:             parseLiteralInstant(t, materialInput.EvidenceValidUntil),
		AttestationExpiresAt:           parseLiteralInstant(t, materialInput.AttestationExpiresAt),
		ActivationDeadline:             parseLiteralInstant(t, materialInput.ActivationDeadline),
		ProviderIdentityDigest:         mustOptionalAuthorityDigest(t, materialInput.ProviderIdentityDigest),
		ExpectedProviderIdentityDigest: mustOptionalAuthorityDigest(t, materialInput.ExpectedProviderIdentityDigest),
		FloorAttestationDigest:         mustOptionalAuthorityDigest(t, materialInput.FloorAttestationDigest), Capability: materialInput.Capability,
	}
	var checkpoint *AuthorityCheckpointAnchor
	if input.Checkpoint != nil {
		anchor := checkpointFromLiteralInput(t, *input.Checkpoint)
		checkpoint = &anchor
	}
	evidenceInput := ActivationDecisionEvidenceInput{
		Material: material, Receipt: receiptFromLiteralInput(t, input.Receipt),
		ProviderHead: providerHeadFromLiteralInput(t, input.ProviderHead), Checkpoint: checkpoint,
	}
	var evidence ActivationDecisionEvidence
	var err error
	if materialInput.TrustedTimeKind == TrustedTimeRollbackResistant {
		start := time.Date(2026, 8, 28, 11, 59, 59, 0, time.UTC)
		calls := 0
		capture := beginActivationEvidenceCaptureForTest(func() time.Time {
			calls++
			return start.Add(time.Duration(calls-1) * time.Second)
		})
		evidence, err = capture.Complete(evidenceInput)
	} else {
		evidence, err = NewActivationDecisionEvidence(evidenceInput)
	}
	if err != nil {
		t.Fatalf("evidence: %v", err)
	}
	proof, err := ValidateActivationDecisionEvidence(evidence, evidenceInput)
	if err != nil {
		t.Fatalf("validate evidence: %v", err)
	}
	return evidence, proof
}

func resolutionFromLiteralInput(t *testing.T, input literalResolutionInput) AuthorityEffectResolution {
	t.Helper()
	_, proof := freshEvidenceProofFromLiteralInput(t, input.Evidence)
	commitment := commitmentFromLiteralInput(t, input.Commitment)
	if proof.Input().Material.Commitment.Digest() != commitment.Digest() {
		t.Fatal("resolution semantic input names mismatched preimages")
	}
	value, err := NewAuthorityEffectResolution(AuthorityEffectResolutionInput{
		Commitment: commitment, Disposition: input.Disposition, Reason: input.Reason, AnchorKind: input.AnchorKind,
		AnchorDigest: mustAuthorityDigest(t, input.AnchorDigest), Evidence: proof,
	})
	if err != nil {
		t.Fatalf("resolution: %v", err)
	}
	return value
}

func parseLiteralAuthorityArtifact(kind string, body []byte) ([]byte, contracts.Digest, error) {
	switch kind {
	case "commitment":
		value, err := ParseAuthorityEffectCommitment(body)
		return value.CanonicalJCS(), value.Digest(), err
	case "provider_head":
		value, err := ParseAuthorityProviderHeadSnapshot(body)
		return value.CanonicalJCS(), value.Digest(), err
	case "checkpoint":
		value, err := ParseAuthorityCheckpointAnchor(body)
		return value.CanonicalJCS(), value.Digest(), err
	case "evidence":
		value, err := ParseActivationDecisionEvidence(body)
		return value.CanonicalJCS(), value.Digest(), err
	case "resolution":
		value, err := ParseAuthorityEffectResolution(body)
		return value.CanonicalJCS(), value.Digest(), err
	default:
		return nil, contracts.Digest{}, ErrInvalidArgument
	}
}

func commitmentFixture(t *testing.T, scenario string) AuthorityEffectCommitment {
	t.Helper()
	input := AuthorityEffectCommitmentInput{
		OperationID:             uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		Kind:                    EffectCertificateActivate,
		ScopeKind:               ScopeNode,
		ScopeDigest:             mustAuthorityDigest(t, strings.Repeat("11", 32)),
		Epoch:                   7,
		Sequence:                11,
		BaseEffectDigest:        mustAuthorityDigest(t, strings.Repeat("22", 32)),
		Mode:                    CommitmentConditionalApply,
		Reason:                  EffectReasonNone,
		ActivationPolicyVersion: 3,
		ActivationInputsDigest:  mustAuthorityDigest(t, strings.Repeat("33", 32)),
	}
	if scenario == "final" {
		input.OperationID = uuid.MustParse("22222222-2222-4222-8222-222222222222")
		input.Sequence = 12
		input.BaseEffectDigest = mustAuthorityDigest(t, strings.Repeat("44", 32))
		input.Mode = CommitmentFinalNotApplied
		input.Reason = EffectReasonFailed
		input.ActivationPolicyVersion = 0
		input.ActivationInputsDigest = contracts.Digest{}
	} else if scenario == "global" {
		input.OperationID = uuid.MustParse("44444444-4444-4444-8444-444444444444")
		input.Kind = EffectRootPublish
		input.ScopeKind = ScopeGlobalNodeTrust
		input.ScopeDigest = globalNodeScopeDigest
		input.Sequence = 21
		input.BaseEffectDigest = mustAuthorityDigest(t, strings.Repeat("aa", 32))
		input.ActivationPolicyVersion = 4
		input.ActivationInputsDigest = mustAuthorityDigest(t, strings.Repeat("bb", 32))
	} else if scenario != "conditional" && scenario != "may_apply" && scenario != "exact_capture" && scenario != "deadline_expired" && scenario != "higher_node" && scenario != "applied" {
		t.Fatalf("unknown commitment scenario %q", scenario)
	}
	value, err := NewAuthorityEffectCommitment(input)
	if err != nil {
		t.Fatalf("commitment: %v", err)
	}
	return value
}

func committedReceiptFixture(t *testing.T, scenario string) Receipt {
	t.Helper()
	commitmentScenario := "conditional"
	operationID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	sequence := uint64(11)
	reservationHex := "6a413ac602211f3e65c1812720220aaa03be6b8b9b4d782c7bc7f523bb4dbf13"
	receiptHex := "c4dfaaf6fd0e5e3b57a8c57670469c2f116d0ed0ba6b7930a6909df2088d0e89"
	if scenario == "final" {
		commitmentScenario = "final"
		operationID = uuid.MustParse("22222222-2222-4222-8222-222222222222")
		sequence = 12
		reservationHex = "ca9232b8b91bdc79d21cee28ce4e0727d0df704343faa1fa5140479fa9ee602e"
		receiptHex = "7022a94c3543624469b7926d6919f93518270d2cab025be8827e39e615a4da90"
	}
	effectDigest := commitmentFixture(t, commitmentScenario).Digest()
	point := DatabasePoint{SystemID: 123456789, Timeline: 2, RequiredLSN: "1/2"}
	receipt := Receipt{
		Reservation: Reservation{
			OperationID: operationID, Kind: EffectCertificateActivate, ScopeKind: ScopeNode,
			ScopeDigest: mustAuthorityDigest(t, strings.Repeat("11", 32)), Epoch: 7, Sequence: sequence,
			ReservationDigest: mustAuthorityDigest(t, reservationHex),
		},
		EffectDigest: &effectDigest, DatabasePoint: &point, Status: StatusCommitted,
		ReceiptDigest: mustAuthorityDigest(t, receiptHex),
	}
	if err := receipt.Validate(); err != nil {
		t.Fatalf("receipt fixture: %v", err)
	}
	return receipt
}

func headFixture(t *testing.T, scenario string) Head {
	t.Helper()
	switch scenario {
	case "exact":
		receipt := committedReceiptFixture(t, "conditional")
		return Head{
			Epoch: 7, LatestReservedSequence: 11, LatestCommittedSequence: 11,
			LatestReservationDigest: receipt.ReservationDigest, LatestCommittedOperationID: receipt.OperationID,
			LatestCommittedReceiptDigest: receipt.ReceiptDigest,
			LatestCommittedDatabasePoint: &DatabasePoint{SystemID: 123456789, Timeline: 2, RequiredLSN: "1/2"},
		}
	case "final":
		receipt := committedReceiptFixture(t, "final")
		return Head{
			Epoch: 7, LatestReservedSequence: 12, LatestCommittedSequence: 12,
			LatestReservationDigest: receipt.ReservationDigest, LatestCommittedOperationID: receipt.OperationID,
			LatestCommittedReceiptDigest: receipt.ReceiptDigest,
			LatestCommittedDatabasePoint: &DatabasePoint{SystemID: 123456789, Timeline: 2, RequiredLSN: "1/2"},
		}
	case "later":
		return Head{
			Epoch: 7, LatestReservedSequence: 13, LatestCommittedSequence: 13,
			LatestReservationDigest:      mustAuthorityDigest(t, strings.Repeat("88", 32)),
			LatestCommittedOperationID:   uuid.MustParse("33333333-3333-4333-8333-333333333333"),
			LatestCommittedReceiptDigest: mustAuthorityDigest(t, strings.Repeat("77", 32)),
			LatestCommittedDatabasePoint: &DatabasePoint{SystemID: 123456790, Timeline: 2, RequiredLSN: "1/3"},
		}
	case "global_later":
		return Head{
			Epoch: 7, LatestReservedSequence: 22, LatestCommittedSequence: 22,
			LatestReservationDigest:      mustAuthorityDigest(t, strings.Repeat("dd", 32)),
			LatestCommittedOperationID:   uuid.MustParse("55555555-5555-4555-8555-555555555555"),
			LatestCommittedReceiptDigest: mustAuthorityDigest(t, strings.Repeat("cc", 32)),
			LatestCommittedDatabasePoint: &DatabasePoint{SystemID: math.MaxUint64, Timeline: math.MaxUint32, RequiredLSN: "2/B"},
		}
	default:
		t.Fatalf("unknown head scenario %q", scenario)
		return Head{}
	}
}

func checkpointFixture(t *testing.T, scenario string) AuthorityCheckpointAnchor {
	t.Helper()
	if scenario == "global" {
		value, err := NewAuthorityCheckpointAnchor(AuthorityCheckpointAnchorInput{
			Kind: CheckpointGlobal, ScopeDigest: globalNodeScopeDigest,
			Checkpoint: NodeCheckpoint{AuthorityEpoch: 7, Sequence: 22, ReceiptDigest: mustAuthorityDigest(t, strings.Repeat("cc", 32))},
		})
		if err != nil {
			t.Fatalf("checkpoint: %v", err)
		}
		return value
	}
	if scenario != "node" && scenario != "higher_node" {
		t.Fatalf("unknown checkpoint scenario %q", scenario)
	}
	value, err := NewAuthorityCheckpointAnchor(AuthorityCheckpointAnchorInput{
		Kind: CheckpointNode, ScopeDigest: mustAuthorityDigest(t, strings.Repeat("11", 32)),
		Checkpoint: NodeCheckpoint{AuthorityEpoch: 7, Sequence: 13, ReceiptDigest: mustAuthorityDigest(t, strings.Repeat("77", 32))},
	})
	if err != nil {
		t.Fatalf("checkpoint: %v", err)
	}
	return value
}

func evidenceInputFixture(t *testing.T, scenario string) ActivationDecisionEvidenceInput {
	t.Helper()
	if scenario == "higher_global" {
		return globalHigherAuthorityInput(t)
	}
	commitmentScenario := "conditional"
	headScenario := "exact"
	receiptScenario := "conditional"
	material := ActivationDecisionMaterial{
		Reason: EffectReasonNone, CheckpointKind: CheckpointNone,
		TrustedTimeKind:                TrustedTimeRollbackResistant,
		TrustedInstant:                 time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC),
		EvidenceValidUntil:             time.Date(2026, 8, 28, 12, 0, 4, 0, time.UTC),
		AttestationExpiresAt:           time.Date(2026, 8, 28, 12, 0, 10, 0, time.UTC),
		ActivationDeadline:             time.Date(2026, 8, 28, 12, 0, 30, 0, time.UTC),
		ProviderIdentityDigest:         mustAuthorityDigest(t, strings.Repeat("55", 32)),
		ExpectedProviderIdentityDigest: mustAuthorityDigest(t, strings.Repeat("55", 32)),
		FloorAttestationDigest:         mustAuthorityDigest(t, strings.Repeat("66", 32)),
		Capability:                     DecisionCapabilityMayApply,
	}
	var checkpoint *AuthorityCheckpointAnchor
	switch scenario {
	case "may_apply", "applied":
	case "exact_capture":
		material.Reason = EffectReasonFailed
		material.Capability = DecisionCapabilityNotAppliedOnly
	case "deadline_expired":
		material.Reason = EffectReasonActivationDeadlineExpired
		material.TrustedInstant = time.Date(2026, 8, 28, 12, 0, 30, 0, time.UTC)
		material.EvidenceValidUntil = time.Date(2026, 8, 28, 12, 0, 34, 0, time.UTC)
		material.AttestationExpiresAt = time.Date(2026, 8, 28, 12, 0, 40, 0, time.UTC)
		material.Capability = DecisionCapabilityNotAppliedOnly
	case "final":
		commitmentScenario, headScenario, receiptScenario = "final", "final", "final"
		material.Reason = EffectReasonFailed
		material.TrustedTimeKind = TrustedTimeNone
		material.TrustedInstant = time.Time{}
		material.EvidenceValidUntil = time.Time{}
		material.AttestationExpiresAt = time.Time{}
		material.ActivationDeadline = time.Time{}
		material.ProviderIdentityDigest = contracts.Digest{}
		material.ExpectedProviderIdentityDigest = contracts.Digest{}
		material.FloorAttestationDigest = contracts.Digest{}
		material.Capability = DecisionCapabilityNotAppliedOnly
	case "higher_node":
		headScenario = "later"
		material.Reason = EffectReasonSuperseded
		material.CheckpointKind = CheckpointNode
		material.CheckpointScopeDigest = mustAuthorityDigest(t, strings.Repeat("11", 32))
		material.TrustedTimeKind = TrustedTimeNone
		material.TrustedInstant = time.Time{}
		material.EvidenceValidUntil = time.Time{}
		material.AttestationExpiresAt = time.Time{}
		material.ActivationDeadline = time.Time{}
		material.ProviderIdentityDigest = contracts.Digest{}
		material.ExpectedProviderIdentityDigest = contracts.Digest{}
		material.FloorAttestationDigest = contracts.Digest{}
		material.Capability = DecisionCapabilityNotAppliedOnly
		anchor := checkpointFixture(t, "node")
		checkpoint = &anchor
	default:
		t.Fatalf("unknown evidence scenario %q", scenario)
	}
	material.Commitment = commitmentFixture(t, commitmentScenario)
	head, err := NewAuthorityProviderHeadSnapshot(headFixture(t, headScenario))
	if err != nil {
		t.Fatalf("head snapshot: %v", err)
	}
	return ActivationDecisionEvidenceInput{
		Material: material, Receipt: committedReceiptFixture(t, receiptScenario), ProviderHead: head, Checkpoint: checkpoint,
	}
}

func freshEvidenceProofFixture(t *testing.T, scenario string) (ActivationDecisionEvidence, ValidatedActivationDecisionEvidence) {
	t.Helper()
	input := evidenceInputFixture(t, scenario)
	var evidence ActivationDecisionEvidence
	var err error
	if input.Material.TrustedTimeKind == TrustedTimeNone {
		evidence, err = NewActivationDecisionEvidence(input)
	} else {
		start := time.Date(2026, 8, 28, 11, 59, 59, 0, time.UTC)
		calls := 0
		capture := beginActivationEvidenceCaptureForTest(func() time.Time {
			calls++
			return start.Add(time.Duration(calls-1) * time.Second)
		})
		evidence, err = capture.Complete(input)
	}
	if err != nil {
		t.Fatalf("evidence: %v", err)
	}
	proof, err := ValidateActivationDecisionEvidence(evidence, input)
	if err != nil {
		t.Fatalf("validate evidence: %v", err)
	}
	return evidence, proof
}

func resolutionFixture(t *testing.T, scenario string) AuthorityEffectResolution {
	t.Helper()
	evidenceScenario := scenario
	if scenario == "applied" {
		evidenceScenario = "may_apply"
	}
	evidence, proof := freshEvidenceProofFixture(t, evidenceScenario)
	commitment := proof.Input().Material.Commitment
	input := AuthorityEffectResolutionInput{Commitment: commitment, Evidence: proof}
	switch scenario {
	case "applied":
		input.Disposition = DispositionApplied
		input.Reason = EffectReasonNone
		input.AnchorKind = DecisionAnchorTrustedTime
		input.AnchorDigest = evidence.Digest()
	case "exact_capture":
		input.Disposition = DispositionNotApplied
		input.Reason = EffectReasonFailed
		input.AnchorKind = DecisionAnchorExactCapture
		input.AnchorDigest = commitment.Facts().ActivationInputsDigest
	case "deadline_expired":
		input.Disposition = DispositionNotApplied
		input.Reason = EffectReasonActivationDeadlineExpired
		input.AnchorKind = DecisionAnchorTrustedTime
		input.AnchorDigest = evidence.Digest()
	case "final":
		input.Disposition = DispositionNotApplied
		input.Reason = EffectReasonFailed
		input.AnchorKind = DecisionAnchorFinalCommitment
		input.AnchorDigest = commitment.Digest()
	case "higher_node", "higher_global":
		input.Disposition = DispositionNotApplied
		input.Reason = EffectReasonSuperseded
		input.AnchorKind = DecisionAnchorHigherAuthority
		input.AnchorDigest = proof.Input().Checkpoint.Digest()
	default:
		t.Fatalf("unknown resolution scenario %q", scenario)
	}
	value, err := NewAuthorityEffectResolution(input)
	if err != nil {
		t.Fatalf("resolution: %v", err)
	}
	return value
}

func mustAuthorityDigest(tb testing.TB, value string) contracts.Digest {
	tb.Helper()
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 32 {
		tb.Fatalf("digest fixture %q", value)
	}
	var digest contracts.Digest
	copy(digest[:], decoded)
	return digest
}
