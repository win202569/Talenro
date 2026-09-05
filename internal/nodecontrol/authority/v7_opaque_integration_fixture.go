//go:build integration

package authority

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gowebpki/jcs"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

const (
	authorityV7IntegrationSignerKeyID = "56475aa75463474c0285df5dbf2bcab73da651358839e9b77481b2eab107708c"
	authorityV7IntegrationTrustRoot   = "a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"
)

var authorityV7IntegrationPublicKey = ed25519.PublicKey{
	0x03, 0xa1, 0x07, 0xbf, 0xf3, 0xce, 0x10, 0xbe,
	0x1d, 0x70, 0xdd, 0x18, 0xe7, 0x4b, 0xc0, 0x99,
	0x67, 0xe4, 0xd6, 0x30, 0x9b, 0xa5, 0x0d, 0x5f,
	0x1d, 0xdc, 0x86, 0x64, 0x12, 0x55, 0x31, 0xb8,
}

var authorityV7IntegrationFactoryAfterSnapshot = func(string) {}

func cloneAuthorityV7DownAuthorizationIntegrationInput(
	facts contracts.AuthorityV7DownAuthorizationFactsV1,
	authorizationBodyJCS []byte,
) authorityV7DownAuthorizationPersistenceView {
	return cloneAuthorityV7DownAuthorizationPersistenceView(authorityV7DownAuthorizationPersistenceView{
		Facts:                    facts,
		AuthorizationBodyJCS:     authorizationBodyJCS,
		AuthorizationEnvelopeJCS: facts.AuthorizationEnvelopeJCS,
	})
}

func NewDisposableAuthorityV7UpGrantForIntegration(facts contracts.AuthorityV7UpMigrationFactsV1) (VerifiedAuthorityV7UpGrant, error) {
	if facts.MigrationLatch.InstallationKind != contracts.AuthorityV7InstallationKindDisposableFixture {
		return VerifiedAuthorityV7UpGrant{}, ErrInvalidArgument
	}
	return newVerifiedAuthorityV7UpGrant(facts)
}

func NewDisposableAuthorityV7DownGrantForIntegration(
	facts contracts.AuthorityV7DownMigrationFactsV1,
	providerRetirementSetBodyJCS []byte,
	providerRetirementEvidenceBundleJCS []byte,
) (VerifiedAuthorityV7DownGrant, error) {
	snapshot := cloneAuthorityV7DownPersistenceView(authorityV7DownPersistenceView{
		Facts:                               facts,
		ProviderRetirementSetBodyJCS:        providerRetirementSetBodyJCS,
		ProviderRetirementEvidenceBundleJCS: providerRetirementEvidenceBundleJCS,
	})
	authorityV7IntegrationFactoryAfterSnapshot("down")
	now := time.Now()
	if snapshot.Facts.MigrationLatch.InstallationKind != contracts.AuthorityV7InstallationKindDisposableFixture || snapshot.Facts.Validate() != nil {
		return VerifiedAuthorityV7DownGrant{}, ErrInvalidArgument
	}
	expectedBody, err := authorityV7IntegrationCanonicalFacts(snapshot.Facts.ProviderRetirementSet)
	if err != nil {
		return VerifiedAuthorityV7DownGrant{}, ErrInvalidArgument
	}
	bodyExpectation := authorityV7CanonicalBodyExpectation{
		Schema:          "provider-protocol-downgrade-retirement-set.v1",
		Fields:          authorityV7RetirementSetFields,
		ExpectedBodyJCS: expectedBody,
		ExpectedDigest:  snapshot.Facts.ProviderRetirementSetDigest,
	}
	var earliestEvidenceExpiry time.Time
	if _, err := verifyAuthorityV7CanonicalBody(snapshot.ProviderRetirementSetBodyJCS, bodyExpectation); err != nil ||
		verifyAuthorityV7DownEvidenceBundle(snapshot.ProviderRetirementEvidenceBundleJCS, snapshot.Facts, now, &earliestEvidenceExpiry) != nil ||
		!now.Before(snapshot.Facts.ExpiresAt) || snapshot.Facts.ExpiresAt.After(earliestEvidenceExpiry) {
		return VerifiedAuthorityV7DownGrant{}, ErrInvalidArgument
	}
	return newVerifiedAuthorityV7DownGrant(snapshot)
}

func NewAuthorityV7DownAuthorizationForIntegration(
	facts contracts.AuthorityV7DownAuthorizationFactsV1,
	authorizationBodyJCS []byte,
) (VerifiedAuthorityV7DownAuthorization, error) {
	snapshot := cloneAuthorityV7DownAuthorizationIntegrationInput(facts, authorizationBodyJCS)
	authorityV7IntegrationFactoryAfterSnapshot("authorization")
	now := time.Now()
	if snapshot.Facts.Validate() != nil || validateAuthorityV7EffectiveWindowAt(snapshot.Facts.Authorization.IssuedAt, snapshot.Facts.Authorization.ExpiresAt, now) != nil ||
		!authorityV7IntegrationAuthorizationPolicyMatchesFacts(snapshot.Facts) {
		return VerifiedAuthorityV7DownAuthorization{}, ErrInvalidArgument
	}
	expectedBody, err := authorityV7IntegrationCanonicalFacts(snapshot.Facts.Authorization)
	if err != nil {
		return VerifiedAuthorityV7DownAuthorization{}, ErrInvalidArgument
	}
	expectation := authorityV7EnvelopeExpectation{
		Body: authorityV7CanonicalBodyExpectation{
			Schema:          "authority-protocol-downgrade-authorization.v1",
			Fields:          authorityV7DownAuthorizationFields,
			ExpectedBodyJCS: expectedBody,
			ExpectedDigest:  snapshot.Facts.AuthorizationDigest,
		},
		EnvelopeJCS: snapshot.AuthorizationEnvelopeJCS,
		Policy:      authorityV7IntegrationPolicy("authority-protocol-downgrade-authorization.v1"),
	}
	if verifyAuthorityV7SignedEnvelope(snapshot.AuthorizationBodyJCS, snapshot.AuthorizationEnvelopeJCS, expectation) != nil {
		return VerifiedAuthorityV7DownAuthorization{}, ErrInvalidArgument
	}
	return newVerifiedAuthorityV7DownAuthorization(snapshot)
}

func NewFreshRestoreImportAdmissionForIntegration(
	facts contracts.FreshRestoreImportProjectionInputV1,
	stagingImportCapabilityBodyJCS []byte,
	stagingImportCapabilityEnvelopeJCS []byte,
	stagingImportCapabilityEvidenceBundleJCS []byte,
	manifestBodyJCS []byte,
	manifestEnvelopeJCS []byte,
	manifestEvidenceBundleJCS []byte,
	stagingExclusionLeaseBodyJCS []byte,
	stagingExclusionLeaseEnvelopeJCS []byte,
	stagingExclusionLeaseEvidenceBundleJCS []byte,
	acquisitionLockedProviderHeadBodyJCS []byte,
	acquisitionLockedProviderHeadEnvelopeJCS []byte,
	currentDatabaseIncarnationProofBodyJCS []byte,
	currentDatabaseIncarnationProofEnvelopeJCS []byte,
	currentDatabaseIncarnationProofEvidenceBundleJCS []byte,
) (VerifiedFreshRestoreImportAdmission, error) {
	snapshot := cloneFreshRestoreImportAdmissionPersistenceView(freshRestoreImportAdmissionPersistenceView{
		Facts:                                            facts,
		StagingImportCapabilityBodyJCS:                   stagingImportCapabilityBodyJCS,
		StagingImportCapabilityEnvelopeJCS:               stagingImportCapabilityEnvelopeJCS,
		StagingImportCapabilityEvidenceBundleJCS:         stagingImportCapabilityEvidenceBundleJCS,
		ManifestBodyJCS:                                  manifestBodyJCS,
		ManifestEnvelopeJCS:                              manifestEnvelopeJCS,
		ManifestEvidenceBundleJCS:                        manifestEvidenceBundleJCS,
		StagingExclusionLeaseBodyJCS:                     stagingExclusionLeaseBodyJCS,
		StagingExclusionLeaseEnvelopeJCS:                 stagingExclusionLeaseEnvelopeJCS,
		StagingExclusionLeaseEvidenceBundleJCS:           stagingExclusionLeaseEvidenceBundleJCS,
		AcquisitionLockedProviderHeadBodyJCS:             acquisitionLockedProviderHeadBodyJCS,
		AcquisitionLockedProviderHeadEnvelopeJCS:         acquisitionLockedProviderHeadEnvelopeJCS,
		CurrentDatabaseIncarnationProofBodyJCS:           currentDatabaseIncarnationProofBodyJCS,
		CurrentDatabaseIncarnationProofEnvelopeJCS:       currentDatabaseIncarnationProofEnvelopeJCS,
		CurrentDatabaseIncarnationProofEvidenceBundleJCS: currentDatabaseIncarnationProofEvidenceBundleJCS,
	})
	authorityV7IntegrationFactoryAfterSnapshot("staging")
	if snapshot.Facts.Validate() != nil {
		return VerifiedFreshRestoreImportAdmission{}, ErrInvalidArgument
	}
	if verifyAuthorityV7StagingImportView(snapshot) != nil {
		return VerifiedFreshRestoreImportAdmission{}, ErrInvalidArgument
	}
	return newVerifiedFreshRestoreImportAdmission(snapshot)
}

var (
	authorityV7RetirementSetFields = []string{
		"installation_id", "release_scope", "environment_inventory_digest", "environment_inventory_anchor_set_digest",
		"environment_inventory_membership_retirement_digest", "provider_count", "retirements", "retired_member_count", "retired_members", "created_at",
	}
	authorityV7DownAuthorizationFields = []string{
		"authorization_id", "installation_id", "migration_latch_digest", "database_identity_digest", "migration_version", "current_catalog_digest",
		"pristine_downgrade_inventory_digest", "provider_protocol_downgrade_retirement_set_digest", "environment_inventory_anchor_set_digest",
		"database_transaction_id", "transaction_nonce", "authorization_scope", "issued_at", "expires_at",
	}
	authorityV7StagingCapabilityFields = []string{
		"capability_id", "single_use_apply_id", "manifest_digest", "target_activation_id", "target_database_identity_digest",
		"database_timeline_lineage_chain_digest", "target_database_incarnation_registration_digest", "runtime_rebind_chain_digest",
		"runtime_instance_binding_digest", "planned_staging_exclusion_id", "expected_pre_acquire_provider_head_digest",
		"pre_acquire_database_incarnation_proof_digest", "provider_phase", "provider_serving_lease_absent_digest",
		"database_route_closed_digest", "pre_import_inventory_digest", "allowed_object_set_digest", "expected_post_import_inventory_digest",
		"transaction_nonce", "issued_at", "expires_at",
	}
	authorityV7ManifestFields = []string{
		"manifest_id", "single_use_apply_id", "fresh_restore_requirement_digest", "source_membership_retirement_digest",
		"source_legacy_retirement_set_digest", "source_post_seal_exact_cover_digest", "source_archive_evidence_digest",
		"source_legacy_epoch_source_set_digest", "source_legacy_epoch_maximum_evidence_digest", "source_snapshot_id", "source_archive_point",
		"target_activation_id", "target_deployment_id", "target_database_incarnation_registration_digest", "target_epoch_evidence_digest",
		"issued_at", "expires_at", "object_count", "objects", "complete_node_set_digest", "forbidden_object_class_set_digest",
		"expected_post_import_inventory_digest",
	}
	authorityV7StagingLeaseFields = []string{
		"exclusion_id", "staging_import_capability_digest", "capability_registration_commit_challenge_digest",
		"capability_registration_commit_attestation_digest", "target_activation_id", "target_database_identity_digest",
		"database_timeline_lineage_chain_digest", "expected_provider_head_digest", "pre_acquire_database_incarnation_proof_digest",
		"target_database_incarnation_registration_digest", "runtime_rebind_chain_digest", "runtime_instance_binding_digest",
		"requested_admission_expires_at", "request_nonce", "request_digest", "exclusion_lease_id",
		"acquisition_locked_provider_head_digest", "provider_serving_lease_absent_digest", "provider_control_sequence",
		"exclusion_state", "issued_at", "admission_expires_at",
	}
	authorityV7ProviderHeadFields = []string{
		"provider_identity_digest", "provider_endpoint_identity_digest", "namespace", "protocol_profile", "genesis_credential_policy_digest",
		"current_credential_policy_digest", "credential_policy_chain_digest", "activation_id", "mode_or_null", "preparation_digest_or_null",
		"activation_digest_or_null", "provider_completion_digest_or_null", "database_completion_digest_or_null",
		"provider_release_preparation_digest_or_null", "database_release_digest_or_null", "open_digest_or_null", "phase",
		"genesis_database_identity_digest", "current_database_identity_digest", "database_timeline_lineage_chain_digest",
		"latest_database_timeline_lineage_attestation_digest_or_null", "genesis_database_incarnation_registration_digest",
		"current_database_incarnation_registration_digest", "runtime_rebind_chain_digest", "incarnation_id", "incarnation_public_key_digest",
		"runtime_instance_binding_digest", "runtime_instance_id", "runtime_instance_generation", "serving_lease_id_or_null",
		"serving_lease_generation_or_null", "serving_lease_digest_or_null", "database_release_attestation_digest_or_null",
		"epoch_evidence_digest_or_null", "genesis_epoch_transition_root_digest_or_null", "selected_genesis_epoch_or_null",
		"current_epoch_or_null", "epoch_transition_chain_digest_or_null", "epoch_transition_terminal_chain_digest_or_null",
		"pending_epoch_transition_digest_or_null", "latest_epoch_transition_resolution_digest_or_null", "epoch_transition_recovery_state_or_null",
		"epoch_transition_recovery_request_digest_or_null", "staging_exclusion_id_or_null", "staging_exclusion_acquire_request_digest_or_null",
		"staging_exclusion_recovery_request_digest_or_null", "staging_exclusion_state_or_null", "latest_reserved_sequence",
		"latest_committed_sequence", "latest_reservation_digest_or_null", "latest_committed_operation_id_or_null",
		"latest_committed_receipt_digest_or_null", "latest_committed_database_point_or_null", "provider_control_sequence", "provider_history_high_water",
	}
	authorityV7ProviderHeadNullable   = authorityV7NullableFieldMap(authorityV7ProviderHeadFields)
	authorityV7IncarnationProofFields = []string{
		"proof_id", "request_digest", "caller_challenge_nonce", "purpose", "provider_challenge_nonce", "provider_head_digest",
		"provider_identity_digest", "provider_endpoint_identity_digest", "namespace", "current_database_identity_digest",
		"database_timeline_lineage_chain_digest", "current_database_incarnation_registration_digest", "runtime_rebind_chain_digest",
		"incarnation_id", "runtime_instance_binding_digest", "runtime_instance_id", "runtime_instance_generation",
		"attestor_runtime_lease_digest", "attestor_runtime_lease_sequence", "attestor_runtime_lease_expires_at",
		"attestor_runtime_lease_descendant_chain_digest_or_null", "serving_lease_id_or_null", "serving_lease_generation_or_null",
		"serving_lease_digest_or_null", "serving_lease_expires_at_or_null", "issued_at", "expires_at",
	}
	authorityV7IncarnationProofNullable = map[string]bool{
		"attestor_runtime_lease_descendant_chain_digest_or_null": true,
		"serving_lease_id_or_null":                               true,
		"serving_lease_generation_or_null":                       true,
		"serving_lease_digest_or_null":                           true,
		"serving_lease_expires_at_or_null":                       true,
	}
	authorityV7TopologyProjectionFields = []string{
		"projection_version", "target_activation_id", "target_deployment_id", "target_database_identity_digest",
		"normalized_catalog_digest", "object_count", "objects",
	}
	authorityV7MembershipRetirementFields = []string{
		"membership_retirement_id", "installation_id", "installation_kind", "migration_latch_digest", "database_identity_digest",
		"release_scope", "final_environment_inventory_digest", "final_inventory_sequence", "final_environment_inventory_anchor_set_digest",
		"environment_member_set_digest", "environment_count", "authorization_nonce", "request_nonce", "issued_at", "expires_at",
		"request_digest", "membership_retirement_tombstone_id", "phase", "retired_at",
	}
	authorityV7ProviderRetirementAuthorizationFields = []string{
		"authorization_id", "retirement_id", "installation_id", "installation_kind", "migration_latch_digest", "database_identity_digest",
		"release_scope", "environment_inventory_digest", "environment_inventory_anchor_set_digest",
		"environment_inventory_membership_retirement_digest", "provider_identity_digest", "provider_endpoint_identity_digest",
		"retirement_member_set_digest", "member_count", "authorization_scope", "authorization_nonce", "issued_at", "expires_at",
	}
	authorityV7HistoryZeroFields = []string{
		"provider_identity_digest", "provider_endpoint_identity_digest", "namespace", "observed_profiles", "unknown_profile_count",
		"unknown_record_count", "legacy_reservation_count", "legacy_terminal_count", "legacy_epoch_transition_count", "legacy_cutover_count",
		"claim_v1_credential_policy_count", "claim_v1_accepted_key_mutation_count", "incarnation_registration_count",
		"genesis_preparation_count", "genesis_completion_count", "genesis_release_preparation_count", "genesis_open_count",
		"claim_v1_reservation_count", "claim_v1_terminal_count", "claim_v1_epoch_transition_count", "claim_v1_epoch_recovery_count",
		"serving_lease_event_count", "runtime_rebind_count", "timeline_lineage_event_count", "staging_exclusion_count",
		"staging_recovery_count", "source_retirement_count", "genesis_authorizing_control_count", "other_mutation_count",
		"provider_history_high_water", "observed_at", "expires_at",
	}
	authorityV7ProviderRetirementFields = []string{
		"retirement_id", "retirement_nonce", "provider_downgrade_retirement_authorization_digest",
		"environment_inventory_membership_retirement_digest", "installation_id", "release_scope", "environment_inventory_digest",
		"environment_inventory_anchor_set_digest", "provider_identity_digest", "provider_endpoint_identity_digest", "member_count",
		"members", "retirement_member_set_digest", "request_nonce", "request_digest", "provider_control_sequence", "namespace_states",
		"phase", "retired_at",
	}
)

func authorityV7NullableFieldMap(fields []string) map[string]bool {
	result := make(map[string]bool)
	for _, field := range fields {
		if strings.HasSuffix(field, "_or_null") {
			result[field] = true
		}
	}
	return result
}

func authorityV7IntegrationPolicy(schema string) authorityV7SignerPolicy {
	return authorityV7SignerPolicy{
		Schema:                 schema,
		SignerRole:             authorityV7FixtureSignerRole(schema),
		SignerKeyID:            authorityV7IntegrationSignerKeyID,
		SignaturePolicyVersion: "1",
		TrustRootDigest:        authorityV7IntegrationTrustRoot,
		SignatureAlgorithm:     "ed25519",
		PublicKey:              cloneAuthorityV7Bytes(authorityV7IntegrationPublicKey),
	}
}

func authorityV7IntegrationCanonicalFacts(value any, omitted ...string) ([]byte, error) {
	wireValue, err := authorityV7IntegrationWireValue(reflect.ValueOf(value))
	if err != nil {
		return nil, ErrInvalidArgument
	}
	object, objectOK := wireValue.(map[string]any)
	if !objectOK {
		return nil, ErrInvalidArgument
	}
	for _, field := range omitted {
		delete(object, field)
	}
	raw, err := json.Marshal(object)
	if err != nil {
		return nil, ErrInvalidArgument
	}
	canonical, err := jcs.Transform(raw)
	if err != nil {
		return nil, ErrInvalidArgument
	}
	return canonical, nil
}

func authorityV7IntegrationWireValue(value reflect.Value) (any, error) {
	if !value.IsValid() {
		return nil, ErrInvalidArgument
	}
	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			return nil, nil
		}
		return authorityV7IntegrationWireValue(value.Elem())
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, nil
		}
		return authorityV7IntegrationWireValue(value.Elem())
	}
	if value.CanInterface() {
		switch typed := value.Interface().(type) {
		case time.Time:
			if typed.IsZero() {
				return nil, ErrInvalidArgument
			}
			return typed.UTC().Format(time.RFC3339Nano), nil
		case uuid.UUID:
			if typed == uuid.Nil {
				return nil, ErrInvalidArgument
			}
			return typed.String(), nil
		case contracts.Digest:
			return hex.EncodeToString(typed[:]), nil
		case json.RawMessage:
			decoder := json.NewDecoder(bytes.NewReader(typed))
			decoder.UseNumber()
			var decoded any
			if err := decoder.Decode(&decoded); err != nil {
				return nil, ErrInvalidArgument
			}
			var trailing json.RawMessage
			if err := decoder.Decode(&trailing); err != io.EOF {
				return nil, ErrInvalidArgument
			}
			return decoded, nil
		}
	}
	switch value.Kind() {
	case reflect.String:
		return value.String(), nil
	case reflect.Bool:
		return value.Bool(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(value.Uint(), 10), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Int() < 0 {
			return nil, ErrInvalidArgument
		}
		return strconv.FormatInt(value.Int(), 10), nil
	case reflect.Slice, reflect.Array:
		result := make([]any, value.Len())
		for index := 0; index < value.Len(); index++ {
			item, err := authorityV7IntegrationWireValue(value.Index(index))
			if err != nil {
				return nil, ErrInvalidArgument
			}
			result[index] = item
		}
		return result, nil
	case reflect.Struct:
		result := make(map[string]any)
		typeOfValue := value.Type()
		for index := 0; index < value.NumField(); index++ {
			fieldType := typeOfValue.Field(index)
			jsonName := strings.Split(fieldType.Tag.Get("json"), ",")[0]
			if jsonName == "" || jsonName == "-" {
				continue
			}
			fieldValue, err := authorityV7IntegrationWireValue(value.Field(index))
			if err != nil {
				return nil, ErrInvalidArgument
			}
			result[jsonName] = fieldValue
		}
		return result, nil
	default:
		return nil, ErrInvalidArgument
	}
}

func authorityV7IntegrationRequireSubset(bodyJCS []byte, expectedFacts any, omitted ...string) error {
	expectedJCS, err := authorityV7IntegrationCanonicalFacts(expectedFacts, omitted...)
	if err != nil {
		return ErrInvalidArgument
	}
	var expected map[string]json.RawMessage
	var body map[string]json.RawMessage
	if json.Unmarshal(expectedJCS, &expected) != nil || json.Unmarshal(bodyJCS, &body) != nil {
		return ErrInvalidArgument
	}
	for field, expectedValue := range expected {
		if actual, present := body[field]; !present || !bytes.Equal(actual, expectedValue) {
			return ErrInvalidArgument
		}
	}
	return nil
}

func authorityV7IntegrationRequireRawFields(bodyJCS []byte, expected map[string]json.RawMessage) error {
	var body map[string]json.RawMessage
	if json.Unmarshal(bodyJCS, &body) != nil {
		return ErrInvalidArgument
	}
	for field, expectedValue := range expected {
		actual, present := body[field]
		if !present || !bytes.Equal(actual, expectedValue) {
			return ErrInvalidArgument
		}
	}
	return nil
}

func authorityV7IntegrationJSONString(value string) json.RawMessage {
	raw, _ := json.Marshal(value)
	return raw
}

func authorityV7IntegrationJSONDigest(value contracts.Digest) json.RawMessage {
	return authorityV7IntegrationJSONString(hex.EncodeToString(value[:]))
}

func authorityV7IntegrationJSONUUID(value uuid.UUID) json.RawMessage {
	return authorityV7IntegrationJSONString(value.String())
}

func authorityV7IntegrationJSONDecimal(value uint64) json.RawMessage {
	return authorityV7IntegrationJSONString(strconv.FormatUint(value, 10))
}

func authorityV7IntegrationAuthorizationPolicyMatchesFacts(facts contracts.AuthorityV7DownAuthorizationFactsV1) bool {
	root, err := hex.DecodeString(authorityV7IntegrationTrustRoot)
	if err != nil || len(root) != len(facts.TrustRootDigest) {
		return false
	}
	var rootDigest contracts.Digest
	copy(rootDigest[:], root)
	return facts.SignerRole == "authority_protocol_downgrade_authorizer" &&
		facts.SignerKeyID == authorityV7IntegrationSignerKeyID && facts.SignaturePolicyVersion == 1 &&
		facts.TrustRootDigest == rootDigest && facts.SignatureAlgorithm == "ed25519"
}

func verifyAuthorityV7StagingImportView(view freshRestoreImportAdmissionPersistenceView) error {
	facts := view.Facts
	capabilityBody, err := authorityV7IntegrationCanonicalFacts(facts.StagingImportCapability, "capability_digest")
	if err != nil {
		return ErrInvalidArgument
	}
	leaseBody, err := authorityV7IntegrationCanonicalFacts(facts.StagingExclusionLease, "lease_digest")
	if err != nil {
		return ErrInvalidArgument
	}

	capability := authorityV7EnvelopeExpectation{
		Body: authorityV7CanonicalBodyExpectation{
			Schema:          "staging-import-capability.v1",
			Fields:          authorityV7StagingCapabilityFields,
			ExpectedBodyJCS: capabilityBody,
			ExpectedDigest:  facts.StagingImportCapability.CapabilityDigest,
		},
		EnvelopeJCS: view.StagingImportCapabilityEnvelopeJCS,
		Policy:      authorityV7IntegrationPolicy("staging-import-capability.v1"),
	}
	manifest := authorityV7EnvelopeExpectation{
		Body: authorityV7CanonicalBodyExpectation{
			Schema:         "fresh-restore-import-manifest.v1",
			Fields:         authorityV7ManifestFields,
			ExpectedDigest: facts.ManifestTopology.ManifestDigest,
		},
		EnvelopeJCS: view.ManifestEnvelopeJCS,
		Policy:      authorityV7IntegrationPolicy("fresh-restore-import-manifest.v1"),
	}
	lease := authorityV7EnvelopeExpectation{
		Body: authorityV7CanonicalBodyExpectation{
			Schema:          "fresh-v7-staging-exclusion-lease.v1",
			Fields:          authorityV7StagingLeaseFields,
			ExpectedBodyJCS: leaseBody,
			ExpectedDigest:  facts.StagingExclusionLease.LeaseDigest,
		},
		EnvelopeJCS: view.StagingExclusionLeaseEnvelopeJCS,
		Policy:      authorityV7IntegrationPolicy("fresh-v7-staging-exclusion-lease.v1"),
	}
	head := authorityV7EnvelopeExpectation{
		Body: authorityV7CanonicalBodyExpectation{
			Schema:         "claim-v1-provider-head.v1",
			Fields:         authorityV7ProviderHeadFields,
			NullablePaths:  authorityV7ProviderHeadNullable,
			ExpectedDigest: facts.CurrentProviderHeadDigest,
		},
		EnvelopeJCS: view.AcquisitionLockedProviderHeadEnvelopeJCS,
		Policy:      authorityV7IntegrationPolicy("claim-v1-provider-head.v1"),
	}
	currentProofDigest := authorityV7DomainDigest("database-incarnation-proof.v1", view.CurrentDatabaseIncarnationProofBodyJCS)
	currentProof := authorityV7EnvelopeExpectation{
		Body: authorityV7CanonicalBodyExpectation{
			Schema:         "database-incarnation-proof.v1",
			Fields:         authorityV7IncarnationProofFields,
			NullablePaths:  authorityV7IncarnationProofNullable,
			ExpectedDigest: currentProofDigest,
		},
		EnvelopeJCS: view.CurrentDatabaseIncarnationProofEnvelopeJCS,
		Policy:      authorityV7IntegrationPolicy("database-incarnation-proof.v1"),
	}
	preAcquireProofBodyJCS, preAcquireProofEnvelopeJCS, err := authorityV7IntegrationEvidenceEnvelope(
		view.StagingImportCapabilityEvidenceBundleJCS,
		"database-incarnation-proof.v1",
		facts.StagingImportCapability.PreAcquireDatabaseIncarnationProofDigest,
	)
	if err != nil {
		return ErrInvalidArgument
	}
	preAcquireProof := authorityV7EnvelopeExpectation{
		Body: authorityV7CanonicalBodyExpectation{
			Schema:          "database-incarnation-proof.v1",
			Fields:          authorityV7IncarnationProofFields,
			NullablePaths:   authorityV7IncarnationProofNullable,
			ExpectedBodyJCS: preAcquireProofBodyJCS,
			ExpectedDigest:  facts.StagingImportCapability.PreAcquireDatabaseIncarnationProofDigest,
		},
		EnvelopeJCS: preAcquireProofEnvelopeJCS,
		Policy:      authorityV7IntegrationPolicy("database-incarnation-proof.v1"),
	}

	for _, item := range []struct {
		body, envelope []byte
		expectation    authorityV7EnvelopeExpectation
	}{
		{view.StagingImportCapabilityBodyJCS, view.StagingImportCapabilityEnvelopeJCS, capability},
		{view.ManifestBodyJCS, view.ManifestEnvelopeJCS, manifest},
		{view.StagingExclusionLeaseBodyJCS, view.StagingExclusionLeaseEnvelopeJCS, lease},
		{view.AcquisitionLockedProviderHeadBodyJCS, view.AcquisitionLockedProviderHeadEnvelopeJCS, head},
		{view.CurrentDatabaseIncarnationProofBodyJCS, view.CurrentDatabaseIncarnationProofEnvelopeJCS, currentProof},
	} {
		if verifyAuthorityV7SignedEnvelope(item.body, item.envelope, item.expectation) != nil {
			return ErrInvalidArgument
		}
	}
	preAcquireProofFields, err := authorityV7IntegrationBindStagingProof(
		preAcquireProofBodyJCS,
		facts,
		facts.StagingImportCapability.ExpectedPreAcquireProviderHeadDigest,
		facts.StagingImportCapability.IssuedAt,
	)
	if err != nil {
		return ErrInvalidArgument
	}
	headFields, err := authorityV7IntegrationBindStagingHeadAndProof(view)
	if err != nil || authorityV7IntegrationMatchStagingHeadAndProof(headFields, preAcquireProofFields) != nil {
		return ErrInvalidArgument
	}
	if authorityV7IntegrationRequireSubset(view.ManifestBodyJCS, facts.ManifestTopology, "manifest_digest") != nil ||
		authorityV7IntegrationValidateManifestSourceFields(view.ManifestBodyJCS) != nil ||
		authorityV7IntegrationVerifyManifestDerivations(facts) != nil {
		return ErrInvalidArgument
	}

	capabilityMembers := authorityV7IntegrationSortedMembers([]authorityV7EvidenceMemberExpectation{
		{EvidenceKind: "external_signed_envelope", Envelope: &preAcquireProof},
		{EvidenceKind: "external_signed_envelope", Envelope: &manifest},
	})
	manifestProjection := authorityV7CanonicalBodyExpectation{
		Schema: "fresh-import-topology-projection.v1",
		Fields: authorityV7TopologyProjectionFields,
		NullablePaths: map[string]bool{
			"objects[].normalized_payload.resume_operator_state_or_null":       true,
			"objects[].normalized_payload.pending_operator_transition_or_null": true,
		},
		ExpectedDigest: facts.ManifestTopology.ExpectedPostImportInventoryDigest,
	}
	manifestMembers := []authorityV7EvidenceMemberExpectation{{EvidenceKind: "database_immutable_body", Body: &manifestProjection}}
	leaseMembers := authorityV7IntegrationSortedMembers([]authorityV7EvidenceMemberExpectation{
		{EvidenceKind: "external_signed_envelope", Envelope: &capability},
		{EvidenceKind: "external_signed_envelope", Envelope: &head},
	})
	currentProofMembers := []authorityV7EvidenceMemberExpectation{{EvidenceKind: "external_signed_envelope", Envelope: &head}}

	if verifyAuthorityV7EvidenceBundle(view.StagingImportCapabilityEvidenceBundleJCS, capability.Body.Schema, capability.Body.ExpectedDigest, capabilityMembers) != nil ||
		verifyAuthorityV7EvidenceBundle(view.ManifestEvidenceBundleJCS, manifest.Body.Schema, manifest.Body.ExpectedDigest, manifestMembers) != nil ||
		verifyAuthorityV7EvidenceBundle(view.StagingExclusionLeaseEvidenceBundleJCS, lease.Body.Schema, lease.Body.ExpectedDigest, leaseMembers) != nil ||
		verifyAuthorityV7EvidenceBundle(view.CurrentDatabaseIncarnationProofEvidenceBundleJCS, currentProof.Body.Schema, currentProof.Body.ExpectedDigest, currentProofMembers) != nil {
		return ErrInvalidArgument
	}
	projectionBody, err := authorityV7IntegrationEvidenceBody(
		view.ManifestEvidenceBundleJCS,
		"fresh-import-topology-projection.v1",
		facts.ManifestTopology.ExpectedPostImportInventoryDigest,
	)
	if err != nil || authorityV7IntegrationBindTopologyProjection(projectionBody, facts) != nil {
		return ErrInvalidArgument
	}
	return nil
}

func authorityV7IntegrationVerifyManifestDerivations(facts contracts.FreshRestoreImportProjectionInputV1) error {
	if validateFreshRestoreManifestPayloadDigests(facts.ManifestTopology) != nil {
		return ErrInvalidArgument
	}
	allowed := make([]map[string]string, len(facts.ManifestTopology.Objects))
	nodeIDs := make([]string, 0)
	for index, object := range facts.ManifestTopology.Objects {
		allowed[index] = map[string]string{
			"object_type": string(object.ObjectType), "canonical_key": object.CanonicalKey,
			"payload_digest": hex.EncodeToString(object.PayloadDigest[:]),
		}
		if object.ObjectType == contracts.FreshRestoreImportObjectNodeReconstructionSeed {
			var payload contracts.FreshRestoreNodeReconstructionSeedPayloadV1
			if json.Unmarshal(object.Payload, &payload) != nil || payload.NodeID == uuid.Nil {
				return ErrInvalidArgument
			}
			nodeIDs = append(nodeIDs, payload.NodeID.String())
		}
	}
	allowedRaw, err := json.Marshal(allowed)
	if err != nil {
		return ErrInvalidArgument
	}
	allowedJCS, err := jcs.Transform(allowedRaw)
	if err != nil || authorityV7DomainDigest("allowed-import-object-set.v1", allowedJCS) != facts.StagingImportCapability.AllowedObjectSetDigest {
		return ErrInvalidArgument
	}
	sort.Strings(nodeIDs)
	nodesRaw, err := json.Marshal(nodeIDs)
	if err != nil {
		return ErrInvalidArgument
	}
	nodesJCS, err := jcs.Transform(nodesRaw)
	if err != nil || authorityV7DomainDigest("complete-node-set.v1", nodesJCS) != facts.ManifestTopology.CompleteNodeSetDigest {
		return ErrInvalidArgument
	}
	emptyProjection := contracts.FreshImportTopologyProjectionV1{
		ProjectionVersion:            1,
		TargetActivationID:           facts.ManifestTopology.TargetActivationID,
		TargetDeploymentID:           facts.ManifestTopology.TargetDeploymentID,
		TargetDatabaseIdentityDigest: facts.CurrentDatabaseIdentityDigest,
		NormalizedCatalogDigest:      facts.NormalizedCatalogDigest,
		Objects:                      []contracts.FreshImportTopologyProjectionObjectV1{},
	}
	emptyBody, err := authorityV7IntegrationCanonicalFacts(emptyProjection)
	if err != nil || authorityV7DomainDigest("fresh-import-topology-projection.v1", emptyBody) != facts.PreImportInventoryDigest {
		return ErrInvalidArgument
	}
	return nil
}

func authorityV7IntegrationSortedMembers(values []authorityV7EvidenceMemberExpectation) []authorityV7EvidenceMemberExpectation {
	result := append([]authorityV7EvidenceMemberExpectation(nil), values...)
	sort.Slice(result, func(left, right int) bool {
		leftSchema, leftDigest := authorityV7IntegrationMemberKey(result[left])
		rightSchema, rightDigest := authorityV7IntegrationMemberKey(result[right])
		return authorityV7CompareEvidenceKeys(leftDigest, leftSchema, rightDigest, rightSchema) < 0
	})
	return result
}

func authorityV7IntegrationMemberKey(value authorityV7EvidenceMemberExpectation) (string, contracts.Digest) {
	if value.Body != nil {
		return value.Body.Schema, value.Body.ExpectedDigest
	}
	if value.Envelope != nil {
		return value.Envelope.Body.Schema, value.Envelope.Body.ExpectedDigest
	}
	return "", contracts.Digest{}
}

func authorityV7IntegrationValidateManifestSourceFields(bodyJCS []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(bodyJCS, &fields) != nil {
		return ErrInvalidArgument
	}
	for _, field := range []string{
		"fresh_restore_requirement_digest", "source_membership_retirement_digest", "source_legacy_retirement_set_digest",
		"source_post_seal_exact_cover_digest", "source_archive_evidence_digest", "source_legacy_epoch_source_set_digest",
		"source_legacy_epoch_maximum_evidence_digest",
	} {
		if _, err := authorityV7JSONDigest(fields[field]); err != nil {
			return ErrInvalidArgument
		}
	}
	if authorityV7IntegrationRequireUUID(fields["source_snapshot_id"]) != nil || authorityV7IntegrationRequirePrintableString(fields["source_archive_point"], 1, 128) != nil {
		return ErrInvalidArgument
	}
	return nil
}

func authorityV7IntegrationBindStagingHeadAndProof(view freshRestoreImportAdmissionPersistenceView) (map[string]json.RawMessage, error) {
	facts := view.Facts
	var headFields map[string]json.RawMessage
	if json.Unmarshal(view.AcquisitionLockedProviderHeadBodyJCS, &headFields) != nil ||
		authorityV7IntegrationValidateKnownScalars(headFields, authorityV7ProviderHeadNullable) != nil {
		return nil, ErrInvalidArgument
	}
	null := json.RawMessage("null")
	headExpected := map[string]json.RawMessage{
		"protocol_profile":                       authorityV7IntegrationJSONString("claim_v1"),
		"activation_id":                          authorityV7IntegrationJSONUUID(facts.ManifestTopology.TargetActivationID),
		"phase":                                  authorityV7IntegrationJSONString("fresh_v7_staging_closed"),
		"current_database_identity_digest":       authorityV7IntegrationJSONDigest(facts.CurrentDatabaseIdentityDigest),
		"database_timeline_lineage_chain_digest": authorityV7IntegrationJSONDigest(facts.DatabaseTimelineLineageChainDigest),
		"current_database_incarnation_registration_digest":  authorityV7IntegrationJSONDigest(facts.CurrentDatabaseIncarnationRegistrationDigest),
		"runtime_rebind_chain_digest":                       authorityV7IntegrationJSONDigest(facts.RuntimeRebindChainDigest),
		"runtime_instance_binding_digest":                   authorityV7IntegrationJSONDigest(facts.RuntimeInstanceBindingDigest),
		"serving_lease_id_or_null":                          null,
		"serving_lease_generation_or_null":                  null,
		"serving_lease_digest_or_null":                      null,
		"pending_epoch_transition_digest_or_null":           null,
		"epoch_transition_recovery_state_or_null":           null,
		"epoch_transition_recovery_request_digest_or_null":  null,
		"staging_exclusion_id_or_null":                      authorityV7IntegrationJSONUUID(facts.StagingExclusionLease.ExclusionID),
		"staging_exclusion_acquire_request_digest_or_null":  authorityV7IntegrationJSONDigest(facts.StagingExclusionLease.RequestDigest),
		"staging_exclusion_recovery_request_digest_or_null": null,
		"staging_exclusion_state_or_null":                   authorityV7IntegrationJSONString("held"),
		"latest_reserved_sequence":                          authorityV7IntegrationJSONDecimal(0),
		"latest_committed_sequence":                         authorityV7IntegrationJSONDecimal(0),
		"latest_reservation_digest_or_null":                 null,
		"latest_committed_operation_id_or_null":             null,
		"latest_committed_receipt_digest_or_null":           null,
		"latest_committed_database_point_or_null":           null,
		"provider_control_sequence":                         authorityV7IntegrationJSONDecimal(facts.StagingExclusionLease.ProviderControlSequence),
	}
	if authorityV7IntegrationRequireRawFields(view.AcquisitionLockedProviderHeadBodyJCS, headExpected) != nil {
		return nil, ErrInvalidArgument
	}
	proofFields, err := authorityV7IntegrationBindStagingProof(
		view.CurrentDatabaseIncarnationProofBodyJCS,
		facts,
		facts.CurrentProviderHeadDigest,
		time.Now(),
	)
	if err != nil {
		return nil, ErrInvalidArgument
	}
	if authorityV7IntegrationMatchStagingHeadAndProof(headFields, proofFields) != nil {
		return nil, ErrInvalidArgument
	}
	return headFields, nil
}

func authorityV7IntegrationMatchStagingHeadAndProof(headFields, proofFields map[string]json.RawMessage) error {
	for _, field := range []string{
		"provider_identity_digest", "provider_endpoint_identity_digest", "namespace", "current_database_identity_digest",
		"database_timeline_lineage_chain_digest", "current_database_incarnation_registration_digest", "runtime_rebind_chain_digest",
		"incarnation_id", "runtime_instance_binding_digest", "runtime_instance_id", "runtime_instance_generation",
	} {
		if !bytes.Equal(headFields[field], proofFields[field]) {
			return ErrInvalidArgument
		}
	}
	return nil
}

func authorityV7IntegrationBindStagingProof(
	bodyJCS []byte,
	facts contracts.FreshRestoreImportProjectionInputV1,
	expectedProviderHeadDigest contracts.Digest,
	checkedAt time.Time,
) (map[string]json.RawMessage, error) {
	var proofFields map[string]json.RawMessage
	if json.Unmarshal(bodyJCS, &proofFields) != nil ||
		authorityV7IntegrationValidateKnownScalars(proofFields, authorityV7IncarnationProofNullable) != nil {
		return nil, ErrInvalidArgument
	}
	null := json.RawMessage("null")
	proofExpected := map[string]json.RawMessage{
		"purpose":                                                authorityV7IntegrationJSONString("staging_import"),
		"provider_head_digest":                                   authorityV7IntegrationJSONDigest(expectedProviderHeadDigest),
		"current_database_identity_digest":                       authorityV7IntegrationJSONDigest(facts.CurrentDatabaseIdentityDigest),
		"database_timeline_lineage_chain_digest":                 authorityV7IntegrationJSONDigest(facts.DatabaseTimelineLineageChainDigest),
		"current_database_incarnation_registration_digest":       authorityV7IntegrationJSONDigest(facts.CurrentDatabaseIncarnationRegistrationDigest),
		"runtime_rebind_chain_digest":                            authorityV7IntegrationJSONDigest(facts.RuntimeRebindChainDigest),
		"runtime_instance_binding_digest":                        authorityV7IntegrationJSONDigest(facts.RuntimeInstanceBindingDigest),
		"attestor_runtime_lease_descendant_chain_digest_or_null": null,
		"serving_lease_id_or_null":                               null,
		"serving_lease_generation_or_null":                       null,
		"serving_lease_digest_or_null":                           null,
		"serving_lease_expires_at_or_null":                       null,
	}
	if authorityV7IntegrationRequireRawFields(bodyJCS, proofExpected) != nil {
		return nil, ErrInvalidArgument
	}
	issuedAt, err := authorityV7IntegrationJSONTime(proofFields["issued_at"])
	if err != nil {
		return nil, ErrInvalidArgument
	}
	expiresAt, err := authorityV7IntegrationJSONTime(proofFields["expires_at"])
	if err != nil {
		return nil, ErrInvalidArgument
	}
	attestorExpiresAt, err := authorityV7IntegrationJSONTime(proofFields["attestor_runtime_lease_expires_at"])
	if err != nil || validateFreshRestoreProofWindow(issuedAt, expiresAt, attestorExpiresAt, checkedAt) != nil {
		return nil, ErrInvalidArgument
	}
	return proofFields, nil
}

func authorityV7IntegrationEvidenceEnvelope(bundleJCS []byte, schema string, digest contracts.Digest) ([]byte, []byte, error) {
	items, err := authorityV7IntegrationParseEvidenceItems(bundleJCS)
	if err != nil {
		return nil, nil, ErrInvalidArgument
	}
	for _, item := range items {
		if item.Kind == "external_signed_envelope" && item.Schema == schema && item.Digest == digest {
			return cloneAuthorityV7Bytes(item.BodyJCS), cloneAuthorityV7Bytes(item.EnvelopeJCS), nil
		}
	}
	return nil, nil, ErrInvalidArgument
}

func authorityV7IntegrationValidateKnownScalars(fields map[string]json.RawMessage, nullable map[string]bool) error {
	for field, raw := range fields {
		if bytes.Equal(raw, []byte("null")) {
			if !nullable[field] {
				return ErrInvalidArgument
			}
			continue
		}
		switch {
		case strings.HasSuffix(field, "_digest"), strings.HasSuffix(field, "_digest_or_null"), strings.HasSuffix(field, "_nonce"):
			if _, err := authorityV7JSONDigest(raw); err != nil {
				return ErrInvalidArgument
			}
		case strings.HasSuffix(field, "_id"), strings.HasSuffix(field, "_id_or_null"):
			if authorityV7IntegrationRequireUUID(raw) != nil {
				return ErrInvalidArgument
			}
		case strings.HasSuffix(field, "_at"), strings.HasSuffix(field, "_at_or_null"):
			if _, err := authorityV7IntegrationJSONTime(raw); err != nil {
				return ErrInvalidArgument
			}
		case strings.HasSuffix(field, "_count"), strings.HasSuffix(field, "_sequence"), strings.HasSuffix(field, "_generation"),
			strings.HasSuffix(field, "_high_water"), strings.HasSuffix(field, "_epoch"), strings.HasSuffix(field, "_epoch_or_null"):
			text, err := authorityV7JSONString(raw)
			if err != nil || !authorityV7CanonicalNonnegativeDecimal(text) {
				return ErrInvalidArgument
			}
		}
	}
	return nil
}

func authorityV7IntegrationRequireUUID(raw json.RawMessage) error {
	text, err := authorityV7JSONString(raw)
	if err != nil {
		return ErrInvalidArgument
	}
	value, err := uuid.Parse(text)
	if err != nil || value == uuid.Nil || value.String() != text {
		return ErrInvalidArgument
	}
	return nil
}

func authorityV7IntegrationRequirePrintableString(raw json.RawMessage, minimum, maximum int) error {
	text, err := authorityV7JSONString(raw)
	if err != nil || len(text) < minimum || len(text) > maximum {
		return ErrInvalidArgument
	}
	for index := range text {
		if text[index] < 0x20 || text[index] > 0x7e {
			return ErrInvalidArgument
		}
	}
	return nil
}

func authorityV7IntegrationJSONTime(raw json.RawMessage) (time.Time, error) {
	return authorityV7CanonicalTime(raw)
}

func authorityV7IntegrationEvidenceBody(bundleJCS []byte, schema string, digest contracts.Digest) ([]byte, error) {
	var bundle struct {
		Evidence []json.RawMessage `json:"evidence"`
	}
	if json.Unmarshal(bundleJCS, &bundle) != nil {
		return nil, ErrInvalidArgument
	}
	for _, itemJCS := range bundle.Evidence {
		var item struct {
			EvidenceKind        string          `json:"evidence_kind"`
			Schema              string          `json:"schema"`
			BodyDigest          string          `json:"body_digest"`
			CanonicalBodyOrNull json.RawMessage `json:"canonical_body_or_null"`
		}
		if json.Unmarshal(itemJCS, &item) != nil || item.EvidenceKind != "database_immutable_body" || item.Schema != schema ||
			item.BodyDigest != hex.EncodeToString(digest[:]) {
			continue
		}
		if len(item.CanonicalBodyOrNull) == 0 || bytes.Equal(item.CanonicalBodyOrNull, []byte("null")) {
			return nil, ErrInvalidArgument
		}
		return cloneAuthorityV7Bytes(item.CanonicalBodyOrNull), nil
	}
	return nil, ErrInvalidArgument
}

func authorityV7IntegrationBindTopologyProjection(bodyJCS []byte, facts contracts.FreshRestoreImportProjectionInputV1) error {
	expectedBody, err := authorityV7IntegrationExpectedTopologyProjection(facts)
	if err != nil || !bytes.Equal(bodyJCS, expectedBody) {
		return ErrInvalidArgument
	}
	fields, err := validateAuthorityV7StrictJSON(bodyJCS, authorityV7StrictJSONSpec{
		Fields: authorityV7TopologyProjectionFields,
		NullablePaths: map[string]bool{
			"objects[].normalized_payload.resume_operator_state_or_null":       true,
			"objects[].normalized_payload.pending_operator_transition_or_null": true,
		},
		MaximumBytes: authorityV7CanonicalArtifactMaxBytes,
	})
	if err != nil {
		return ErrInvalidArgument
	}
	version, err := authorityV7JSONString(fields["projection_version"])
	if err != nil || version != "1" {
		return ErrInvalidArgument
	}
	activationID, err := authorityV7IntegrationParseUUID(fields["target_activation_id"])
	if err != nil {
		return ErrInvalidArgument
	}
	deploymentID, err := authorityV7IntegrationParseUUID(fields["target_deployment_id"])
	if err != nil {
		return ErrInvalidArgument
	}
	databaseIdentity, err := authorityV7JSONDigest(fields["target_database_identity_digest"])
	if err != nil {
		return ErrInvalidArgument
	}
	normalizedCatalog, err := authorityV7JSONDigest(fields["normalized_catalog_digest"])
	if err != nil {
		return ErrInvalidArgument
	}
	countText, err := authorityV7JSONString(fields["object_count"])
	if err != nil || !authorityV7CanonicalNonnegativeDecimal(countText) {
		return ErrInvalidArgument
	}
	count, err := strconv.ParseUint(countText, 10, 64)
	if err != nil {
		return ErrInvalidArgument
	}
	var rawObjects []json.RawMessage
	if json.Unmarshal(fields["objects"], &rawObjects) != nil || rawObjects == nil || uint64(len(rawObjects)) != count {
		return ErrInvalidArgument
	}
	objects := make([]contracts.FreshImportTopologyProjectionObjectV1, len(rawObjects))
	for index, objectJCS := range rawObjects {
		objectFields, err := validateAuthorityV7StrictJSON(objectJCS, authorityV7StrictJSONSpec{
			Fields: []string{"object_type", "canonical_key", "normalized_payload"},
			NullablePaths: map[string]bool{
				"normalized_payload.resume_operator_state_or_null":       true,
				"normalized_payload.pending_operator_transition_or_null": true,
			},
			MaximumBytes: authorityV7CanonicalArtifactMaxBytes,
		})
		if err != nil {
			return ErrInvalidArgument
		}
		objectType, err := authorityV7JSONString(objectFields["object_type"])
		if err != nil {
			return ErrInvalidArgument
		}
		canonicalKey, err := authorityV7JSONString(objectFields["canonical_key"])
		if err != nil {
			return ErrInvalidArgument
		}
		objects[index] = contracts.FreshImportTopologyProjectionObjectV1{
			ObjectType:        contracts.FreshRestoreImportObjectTypeV1(objectType),
			CanonicalKey:      canonicalKey,
			NormalizedPayload: cloneAuthorityV7Bytes(objectFields["normalized_payload"]),
		}
	}
	projection := contracts.FreshImportTopologyProjectionV1{
		ProjectionVersion:            1,
		TargetActivationID:           activationID,
		TargetDeploymentID:           deploymentID,
		TargetDatabaseIdentityDigest: databaseIdentity,
		NormalizedCatalogDigest:      normalizedCatalog,
		ObjectCount:                  count,
		Objects:                      objects,
	}
	if projection.Validate() != nil || activationID != facts.ManifestTopology.TargetActivationID || deploymentID != facts.ManifestTopology.TargetDeploymentID ||
		databaseIdentity != facts.CurrentDatabaseIdentityDigest || normalizedCatalog != facts.NormalizedCatalogDigest || count != facts.ManifestTopology.ObjectCount {
		return ErrInvalidArgument
	}
	return nil
}

func authorityV7IntegrationExpectedTopologyProjection(facts contracts.FreshRestoreImportProjectionInputV1) ([]byte, error) {
	objects := make([]contracts.FreshImportTopologyProjectionObjectV1, len(facts.ManifestTopology.Objects))
	for index, source := range facts.ManifestTopology.Objects {
		normalized, err := authorityV7IntegrationNormalizedPayload(source)
		if err != nil {
			return nil, ErrInvalidArgument
		}
		objects[index] = contracts.FreshImportTopologyProjectionObjectV1{
			ObjectType:        source.ObjectType,
			CanonicalKey:      source.CanonicalKey,
			NormalizedPayload: normalized,
		}
	}
	projection := contracts.FreshImportTopologyProjectionV1{
		ProjectionVersion:            1,
		TargetActivationID:           facts.ManifestTopology.TargetActivationID,
		TargetDeploymentID:           facts.ManifestTopology.TargetDeploymentID,
		TargetDatabaseIdentityDigest: facts.CurrentDatabaseIdentityDigest,
		NormalizedCatalogDigest:      facts.NormalizedCatalogDigest,
		ObjectCount:                  uint64(len(objects)),
		Objects:                      objects,
	}
	if projection.Validate() != nil {
		return nil, ErrInvalidArgument
	}
	return authorityV7IntegrationCanonicalFacts(projection)
}

func authorityV7IntegrationNormalizedPayload(source contracts.FreshRestoreImportManifestObjectV1) ([]byte, error) {
	var value any
	switch source.ObjectType {
	case contracts.FreshRestoreImportObjectPOP:
		var input contracts.FreshRestorePOPImportPayloadV1
		if json.Unmarshal(source.Payload, &input) != nil {
			return nil, ErrInvalidArgument
		}
		value = contracts.FreshRestoreNormalizedPOPPayloadV1{
			POPCode: input.POPCode, ISOCountry: input.ISOCountry, Region: input.Region,
			OperatorState: "disabled", Version: "1",
		}
	case contracts.FreshRestoreImportObjectFailureDomainDefinition:
		var input contracts.FreshRestoreFailureDomainImportPayloadV1
		if json.Unmarshal(source.Payload, &input) != nil {
			return nil, ErrInvalidArgument
		}
		value = contracts.FreshRestoreNormalizedFailureDomainPayloadV1{
			FailureDomainID: input.FailureDomainID, DomainType: input.DomainType, StableID: input.StableID, Version: "1",
		}
	case contracts.FreshRestoreImportObjectCapacityProfileDefinition:
		canonical, err := jcs.Transform(source.Payload)
		if err != nil || !bytes.Equal(canonical, source.Payload) {
			return nil, ErrInvalidArgument
		}
		return cloneAuthorityV7Bytes(source.Payload), nil
	case contracts.FreshRestoreImportObjectNodeReconstructionSeed:
		var input contracts.FreshRestoreNodeReconstructionSeedPayloadV1
		if json.Unmarshal(source.Payload, &input) != nil {
			return nil, ErrInvalidArgument
		}
		value = contracts.FreshRestoreNormalizedNodePayloadV1{
			NodeID: input.NodeID, POPCode: input.POPCode, OperatorState: "disabled", SecurityState: "quarantined",
			IdentityState: "unauthorized", HealthState: "unknown", IdentityEpoch: "0", InventoryVersion: "1",
			SecurityVersion: "1", NextDesiredGeneration: "1", NextRecoveryGeneration: "1",
			ActivePointerSet: []json.RawMessage{}, AuthorityAnchorSet: []json.RawMessage{},
		}
	default:
		return nil, ErrInvalidArgument
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, ErrInvalidArgument
	}
	canonical, err := jcs.Transform(raw)
	if err != nil {
		return nil, ErrInvalidArgument
	}
	return canonical, nil
}

func authorityV7IntegrationParseUUID(raw json.RawMessage) (uuid.UUID, error) {
	text, err := authorityV7JSONString(raw)
	if err != nil {
		return uuid.Nil, ErrInvalidArgument
	}
	value, err := uuid.Parse(text)
	if err != nil || value == uuid.Nil || value.String() != text {
		return uuid.Nil, ErrInvalidArgument
	}
	return value, nil
}

type authorityV7IntegrationEvidenceItem struct {
	Kind        string
	Schema      string
	Digest      contracts.Digest
	BodyJCS     []byte
	EnvelopeJCS []byte
	BodyFields  map[string]json.RawMessage
}

func verifyAuthorityV7DownEvidenceBundle(
	bundleJCS []byte,
	facts contracts.AuthorityV7DownMigrationFactsV1,
	now time.Time,
	earliestExpiry *time.Time,
) error {
	if facts.ProviderRetirementSet.Validate() != nil || earliestExpiry == nil {
		return ErrInvalidArgument
	}
	items, err := authorityV7IntegrationParseEvidenceItems(bundleJCS)
	wantCount := 1 + 3*len(facts.ProviderRetirementSet.Retirements)
	if err != nil || len(items) != wantCount || uint64(len(facts.ProviderRetirementSet.Retirements)) != facts.ProviderRetirementSet.ProviderCount {
		return ErrInvalidArgument
	}

	retirements := make(map[string]contracts.AuthorityV7ProviderRetirementFactsV1, len(facts.ProviderRetirementSet.Retirements))
	for _, retirement := range facts.ProviderRetirementSet.Retirements {
		retirements[authorityV7IntegrationProviderPair(retirement.ProviderIdentityDigest, retirement.ProviderEndpointIdentityDigest)] = retirement
	}
	adminDigests := make(map[string]contracts.Digest)
	adminBodies := make(map[string]map[string]json.RawMessage)
	historyDigests := make(map[string]contracts.Digest)
	historyBodies := make(map[string]map[string]json.RawMessage)
	retirementBodies := make(map[string]map[string]json.RawMessage)
	membershipCount := 0
	expectations := make([]authorityV7EvidenceMemberExpectation, 0, len(items))
	var earliest time.Time
	recordExpiry := func(value time.Time) {
		if earliest.IsZero() || value.Before(earliest) {
			earliest = value
		}
	}

	for _, item := range items {
		if item.Kind != "external_signed_envelope" {
			return ErrInvalidArgument
		}
		fields := authorityV7DownEvidenceSchemaFields(item.Schema)
		policy := authorityV7IntegrationPolicy(item.Schema)
		if len(fields) == 0 || policy.SignerRole == "" || authorityV7IntegrationValidateKnownScalars(item.BodyFields, nil) != nil {
			return ErrInvalidArgument
		}
		bodyExpectation := authorityV7CanonicalBodyExpectation{
			Schema:          item.Schema,
			Fields:          fields,
			ExpectedBodyJCS: item.BodyJCS,
			ExpectedDigest:  item.Digest,
		}
		envelopeExpectation := authorityV7EnvelopeExpectation{Body: bodyExpectation, EnvelopeJCS: item.EnvelopeJCS, Policy: policy}
		expectations = append(expectations, authorityV7EvidenceMemberExpectation{EvidenceKind: item.Kind, Envelope: &envelopeExpectation})

		switch item.Schema {
		case "environment-inventory-membership-retirement.v1":
			membershipCount++
			expiresAt, windowErr := authorityV7IntegrationBindMembershipRetirement(item.BodyJCS, facts, now)
			if membershipCount != 1 || item.Digest != facts.ProviderRetirementSet.EnvironmentInventoryMembershipRetirementDigest ||
				windowErr != nil {
				return ErrInvalidArgument
			}
			recordExpiry(expiresAt)
		case "provider-protocol-downgrade-retirement-authorization.v1":
			pair, err := authorityV7IntegrationBodyProviderPair(item.BodyFields)
			expiresAt, windowErr := authorityV7IntegrationBindProviderRetirementAuthorization(item.BodyJCS, facts, pair, now)
			if err != nil || retirements[pair].ProviderIdentityDigest == (contracts.Digest{}) || adminDigests[pair] != (contracts.Digest{}) ||
				windowErr != nil {
				return ErrInvalidArgument
			}
			recordExpiry(expiresAt)
			adminDigests[pair] = item.Digest
			adminBodies[pair] = item.BodyFields
		case "provider-protocol-history-zero-projection.v1":
			pair, err := authorityV7IntegrationBodyProviderPair(item.BodyFields)
			expiresAt, windowErr := authorityV7IntegrationBindHistoryZero(item.BodyJCS, facts, pair, now)
			if err != nil || retirements[pair].ProviderIdentityDigest == (contracts.Digest{}) || historyDigests[pair] != (contracts.Digest{}) ||
				windowErr != nil {
				return ErrInvalidArgument
			}
			recordExpiry(expiresAt)
			historyDigests[pair] = item.Digest
			historyBodies[pair] = item.BodyFields
		case "provider-protocol-downgrade-retirement.v1":
			pair, err := authorityV7IntegrationBodyProviderPair(item.BodyFields)
			retirement, present := retirements[pair]
			if err != nil || !present || retirement.ProviderProtocolDowngradeRetirementDigest != item.Digest || retirementBodies[pair] != nil ||
				authorityV7IntegrationBindProviderRetirement(item.BodyJCS, facts, pair) != nil {
				return ErrInvalidArgument
			}
			retirementBodies[pair] = item.BodyFields
		default:
			return ErrInvalidArgument
		}
	}
	if membershipCount != 1 || len(adminDigests) != len(retirements) || len(historyDigests) != len(retirements) || len(retirementBodies) != len(retirements) || earliest.IsZero() {
		return ErrInvalidArgument
	}
	for pair := range retirements {
		body := retirementBodies[pair]
		authorizationDigest, err := authorityV7JSONDigest(body["provider_downgrade_retirement_authorization_digest"])
		if err != nil || authorizationDigest != adminDigests[pair] || !bytes.Equal(body["retirement_id"], adminBodies[pair]["retirement_id"]) ||
			authorityV7IntegrationRetirementHistoryDigest(body, historyBodies[pair], historyDigests[pair], facts.ProviderRetirementSet.RetiredMembers, pair) != nil {
			return ErrInvalidArgument
		}
	}
	expectations = authorityV7IntegrationSortedMembers(expectations)
	if verifyAuthorityV7EvidenceBundle(
		bundleJCS,
		"provider-protocol-downgrade-retirement-set.v1",
		facts.ProviderRetirementSetDigest,
		expectations,
	) != nil {
		return ErrInvalidArgument
	}
	*earliestExpiry = earliest
	return nil
}

func authorityV7IntegrationParseEvidenceItems(bundleJCS []byte) ([]authorityV7IntegrationEvidenceItem, error) {
	fields, err := validateAuthorityV7StrictJSON(bundleJCS, authorityV7StrictJSONSpec{
		Fields: []string{"message_schema", "message_body_digest", "evidence_count", "evidence"},
		NullablePaths: map[string]bool{
			"evidence[].canonical_body_or_null":     true,
			"evidence[].canonical_envelope_or_null": true,
		},
		NullablePrefixes: []string{"evidence[].canonical_body_or_null.", "evidence[].canonical_envelope_or_null.body."},
		MaximumBytes:     authorityV7EvidenceBundleMaxBytes,
	})
	if err != nil {
		return nil, ErrInvalidArgument
	}
	var rawItems []json.RawMessage
	if json.Unmarshal(fields["evidence"], &rawItems) != nil || rawItems == nil {
		return nil, ErrInvalidArgument
	}
	items := make([]authorityV7IntegrationEvidenceItem, len(rawItems))
	for index, rawItem := range rawItems {
		itemFields, err := validateAuthorityV7StrictJSON(rawItem, authorityV7StrictJSONSpec{
			Fields: []string{"evidence_kind", "schema", "body_digest", "canonical_body_or_null", "canonical_envelope_or_null"},
			NullablePaths: map[string]bool{
				"canonical_body_or_null":     true,
				"canonical_envelope_or_null": true,
			},
			NullablePrefixes: []string{"canonical_body_or_null.", "canonical_envelope_or_null.body."},
			MaximumBytes:     authorityV7CanonicalArtifactMaxBytes,
		})
		if err != nil {
			return nil, ErrInvalidArgument
		}
		kind, err := authorityV7JSONString(itemFields["evidence_kind"])
		if err != nil {
			return nil, ErrInvalidArgument
		}
		schema, err := authorityV7JSONString(itemFields["schema"])
		if err != nil {
			return nil, ErrInvalidArgument
		}
		digest, err := authorityV7JSONDigest(itemFields["body_digest"])
		if err != nil {
			return nil, ErrInvalidArgument
		}
		if kind != "external_signed_envelope" || !bytes.Equal(itemFields["canonical_body_or_null"], []byte("null")) || bytes.Equal(itemFields["canonical_envelope_or_null"], []byte("null")) {
			return nil, ErrInvalidArgument
		}
		var envelope map[string]json.RawMessage
		if json.Unmarshal(itemFields["canonical_envelope_or_null"], &envelope) != nil || len(envelope["body"]) == 0 {
			return nil, ErrInvalidArgument
		}
		var bodyFields map[string]json.RawMessage
		if json.Unmarshal(envelope["body"], &bodyFields) != nil || bodyFields == nil {
			return nil, ErrInvalidArgument
		}
		items[index] = authorityV7IntegrationEvidenceItem{
			Kind:        kind,
			Schema:      schema,
			Digest:      digest,
			BodyJCS:     cloneAuthorityV7Bytes(envelope["body"]),
			EnvelopeJCS: cloneAuthorityV7Bytes(itemFields["canonical_envelope_or_null"]),
			BodyFields:  bodyFields,
		}
	}
	return items, nil
}

func authorityV7DownEvidenceSchemaFields(schema string) []string {
	switch schema {
	case "environment-inventory-membership-retirement.v1":
		return authorityV7MembershipRetirementFields
	case "provider-protocol-downgrade-retirement-authorization.v1":
		return authorityV7ProviderRetirementAuthorizationFields
	case "provider-protocol-history-zero-projection.v1":
		return authorityV7HistoryZeroFields
	case "provider-protocol-downgrade-retirement.v1":
		return authorityV7ProviderRetirementFields
	default:
		return nil
	}
}

func authorityV7IntegrationProviderPair(identity, endpoint contracts.Digest) string {
	return hex.EncodeToString(identity[:]) + "\x00" + hex.EncodeToString(endpoint[:])
}

func authorityV7IntegrationBodyProviderPair(fields map[string]json.RawMessage) (string, error) {
	identity, err := authorityV7JSONDigest(fields["provider_identity_digest"])
	if err != nil {
		return "", ErrInvalidArgument
	}
	endpoint, err := authorityV7JSONDigest(fields["provider_endpoint_identity_digest"])
	if err != nil {
		return "", ErrInvalidArgument
	}
	return authorityV7IntegrationProviderPair(identity, endpoint), nil
}

func authorityV7IntegrationBindMembershipRetirement(bodyJCS []byte, facts contracts.AuthorityV7DownMigrationFactsV1, now time.Time) (time.Time, error) {
	set := facts.ProviderRetirementSet
	memberSetBody, err := authorityV7IntegrationCanonicalJSONValue(authorityV7IntegrationEnvironmentRecordDigests(set.RetiredMembers))
	if err != nil {
		return time.Time{}, ErrInvalidArgument
	}
	memberSetDigest := authorityV7DomainDigest("environment-inventory-member-set.v1", memberSetBody)
	expected := map[string]json.RawMessage{
		"installation_id":                               authorityV7IntegrationJSONUUID(set.InstallationID),
		"installation_kind":                             authorityV7IntegrationJSONString("disposable_fixture"),
		"migration_latch_digest":                        authorityV7IntegrationJSONDigest(facts.MigrationLatchDigest),
		"database_identity_digest":                      authorityV7IntegrationJSONDigest(facts.MigrationLatch.DatabaseIdentityDigest),
		"release_scope":                                 authorityV7IntegrationJSONString(set.ReleaseScope),
		"final_environment_inventory_digest":            authorityV7IntegrationJSONDigest(set.EnvironmentInventoryDigest),
		"final_environment_inventory_anchor_set_digest": authorityV7IntegrationJSONDigest(set.EnvironmentInventoryAnchorSetDigest),
		"environment_member_set_digest":                 authorityV7IntegrationJSONDigest(memberSetDigest),
		"environment_count":                             authorityV7IntegrationJSONDecimal(set.RetiredMemberCount),
		"phase":                                         authorityV7IntegrationJSONString("inventory_membership_retired"),
	}
	if authorityV7IntegrationRequireRawFields(bodyJCS, expected) != nil {
		return time.Time{}, ErrInvalidArgument
	}
	expiresAt, err := validateAuthorityV7CanonicalWindowAt(bodyJCS, "issued_at", 5*time.Minute, now)
	if err != nil {
		return time.Time{}, ErrInvalidArgument
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(bodyJCS, &fields) != nil {
		return time.Time{}, ErrInvalidArgument
	}
	sequence, err := authorityV7JSONString(fields["final_inventory_sequence"])
	if err != nil || !authorityV7CanonicalPositiveDecimal(sequence) {
		return time.Time{}, ErrInvalidArgument
	}
	return expiresAt, nil
}

func authorityV7IntegrationBindProviderRetirementAuthorization(bodyJCS []byte, facts contracts.AuthorityV7DownMigrationFactsV1, pair string, now time.Time) (time.Time, error) {
	set := facts.ProviderRetirementSet
	retirement, present := authorityV7IntegrationRetirementForPair(set.Retirements, pair)
	if !present {
		return time.Time{}, ErrInvalidArgument
	}
	members := authorityV7IntegrationMembersForPair(set.RetiredMembers, pair)
	memberSetBody, err := authorityV7IntegrationCanonicalJSONValue(members)
	if err != nil {
		return time.Time{}, ErrInvalidArgument
	}
	memberSetDigest := authorityV7DomainDigest("provider-protocol-retirement-member-set.v1", memberSetBody)
	expected := map[string]json.RawMessage{
		"installation_id":                                    authorityV7IntegrationJSONUUID(set.InstallationID),
		"installation_kind":                                  authorityV7IntegrationJSONString("disposable_fixture"),
		"migration_latch_digest":                             authorityV7IntegrationJSONDigest(facts.MigrationLatchDigest),
		"database_identity_digest":                           authorityV7IntegrationJSONDigest(facts.MigrationLatch.DatabaseIdentityDigest),
		"release_scope":                                      authorityV7IntegrationJSONString(set.ReleaseScope),
		"environment_inventory_digest":                       authorityV7IntegrationJSONDigest(set.EnvironmentInventoryDigest),
		"environment_inventory_anchor_set_digest":            authorityV7IntegrationJSONDigest(set.EnvironmentInventoryAnchorSetDigest),
		"environment_inventory_membership_retirement_digest": authorityV7IntegrationJSONDigest(set.EnvironmentInventoryMembershipRetirementDigest),
		"provider_identity_digest":                           authorityV7IntegrationJSONDigest(retirement.ProviderIdentityDigest),
		"provider_endpoint_identity_digest":                  authorityV7IntegrationJSONDigest(retirement.ProviderEndpointIdentityDigest),
		"retirement_member_set_digest":                       authorityV7IntegrationJSONDigest(memberSetDigest),
		"member_count":                                       authorityV7IntegrationJSONDecimal(uint64(len(members))),
		"authorization_scope":                                authorityV7IntegrationJSONString("permanent_disposable_namespace_retirement"),
	}
	if authorityV7IntegrationRequireRawFields(bodyJCS, expected) != nil {
		return time.Time{}, ErrInvalidArgument
	}
	return validateAuthorityV7CanonicalWindowAt(bodyJCS, "issued_at", 5*time.Minute, now)
}

func authorityV7IntegrationBindHistoryZero(bodyJCS []byte, facts contracts.AuthorityV7DownMigrationFactsV1, pair string, now time.Time) (time.Time, error) {
	members := authorityV7IntegrationMembersForPair(facts.ProviderRetirementSet.RetiredMembers, pair)
	if len(members) == 0 {
		return time.Time{}, ErrInvalidArgument
	}
	namespace := members[0].ProviderNamespace
	for _, member := range members[1:] {
		if member.ProviderNamespace != namespace {
			return time.Time{}, ErrInvalidArgument
		}
	}
	retirement, present := authorityV7IntegrationRetirementForPair(facts.ProviderRetirementSet.Retirements, pair)
	if !present {
		return time.Time{}, ErrInvalidArgument
	}
	expected := map[string]json.RawMessage{
		"provider_identity_digest":          authorityV7IntegrationJSONDigest(retirement.ProviderIdentityDigest),
		"provider_endpoint_identity_digest": authorityV7IntegrationJSONDigest(retirement.ProviderEndpointIdentityDigest),
		"namespace":                         authorityV7IntegrationJSONString(namespace),
		"observed_profiles":                 json.RawMessage(`["claim_v1","legacy_v6"]`),
	}
	for _, field := range authorityV7HistoryZeroFields {
		if strings.HasSuffix(field, "_count") || field == "provider_history_high_water" {
			expected[field] = authorityV7IntegrationJSONDecimal(0)
		}
	}
	if authorityV7IntegrationRequireRawFields(bodyJCS, expected) != nil {
		return time.Time{}, ErrInvalidArgument
	}
	return validateAuthorityV7CanonicalWindowAt(bodyJCS, "observed_at", 2*time.Minute, now)
}

func authorityV7IntegrationBindProviderRetirement(bodyJCS []byte, facts contracts.AuthorityV7DownMigrationFactsV1, pair string) error {
	set := facts.ProviderRetirementSet
	retirement, present := authorityV7IntegrationRetirementForPair(set.Retirements, pair)
	if !present {
		return ErrInvalidArgument
	}
	members := authorityV7IntegrationMembersForPair(set.RetiredMembers, pair)
	membersJCS, err := authorityV7IntegrationCanonicalJSONValue(members)
	if err != nil {
		return ErrInvalidArgument
	}
	memberSetDigest := authorityV7DomainDigest("provider-protocol-retirement-member-set.v1", membersJCS)
	expected := map[string]json.RawMessage{
		"environment_inventory_membership_retirement_digest": authorityV7IntegrationJSONDigest(set.EnvironmentInventoryMembershipRetirementDigest),
		"installation_id":                         authorityV7IntegrationJSONUUID(set.InstallationID),
		"release_scope":                           authorityV7IntegrationJSONString(set.ReleaseScope),
		"environment_inventory_digest":            authorityV7IntegrationJSONDigest(set.EnvironmentInventoryDigest),
		"environment_inventory_anchor_set_digest": authorityV7IntegrationJSONDigest(set.EnvironmentInventoryAnchorSetDigest),
		"provider_identity_digest":                authorityV7IntegrationJSONDigest(retirement.ProviderIdentityDigest),
		"provider_endpoint_identity_digest":       authorityV7IntegrationJSONDigest(retirement.ProviderEndpointIdentityDigest),
		"member_count":                            authorityV7IntegrationJSONDecimal(uint64(len(members))),
		"members":                                 membersJCS,
		"retirement_member_set_digest":            authorityV7IntegrationJSONDigest(memberSetDigest),
		"phase":                                   authorityV7IntegrationJSONString("down_retired"),
	}
	if authorityV7IntegrationRequireRawFields(bodyJCS, expected) != nil {
		return ErrInvalidArgument
	}
	return nil
}

func authorityV7IntegrationRetirementHistoryDigest(
	fields map[string]json.RawMessage,
	historyFields map[string]json.RawMessage,
	expected contracts.Digest,
	allMembers []contracts.AuthorityV7RetiredEnvironmentMemberFactsV1,
	pair string,
) error {
	var states []map[string]json.RawMessage
	if json.Unmarshal(fields["namespace_states"], &states) != nil || len(states) != 1 {
		return ErrInvalidArgument
	}
	state := states[0]
	if len(state) != 7 {
		return ErrInvalidArgument
	}
	for _, field := range []string{
		"namespace", "environment_record_digests", "database_identity_digests", "history_zero_projection_digest",
		"pre_retirement_control_sequence", "provider_history_high_water", "retirement_tombstone_id",
	} {
		if len(state[field]) == 0 {
			return ErrInvalidArgument
		}
	}
	digest, err := authorityV7JSONDigest(state["history_zero_projection_digest"])
	if err != nil || digest != expected || !bytes.Equal(state["namespace"], historyFields["namespace"]) {
		return ErrInvalidArgument
	}
	members := authorityV7IntegrationMembersForPair(allMembers, pair)
	recordDigests := make([]string, len(members))
	databaseDigests := make([]string, len(members))
	for index, member := range members {
		recordDigests[index] = hex.EncodeToString(member.EnvironmentRecordDigest[:])
		databaseDigests[index] = hex.EncodeToString(member.DatabaseIdentityDigest[:])
	}
	recordsJCS, err := authorityV7IntegrationCanonicalJSONValue(recordDigests)
	if err != nil || !bytes.Equal(recordsJCS, state["environment_record_digests"]) {
		return ErrInvalidArgument
	}
	databasesJCS, err := authorityV7IntegrationCanonicalJSONValue(databaseDigests)
	if err != nil || !bytes.Equal(databasesJCS, state["database_identity_digests"]) {
		return ErrInvalidArgument
	}
	highWater, err := authorityV7JSONString(state["provider_history_high_water"])
	if err != nil || highWater != "0" {
		return ErrInvalidArgument
	}
	sequence, err := authorityV7JSONString(state["pre_retirement_control_sequence"])
	if err != nil || !authorityV7CanonicalNonnegativeDecimal(sequence) || authorityV7IntegrationRequireUUID(state["retirement_tombstone_id"]) != nil {
		return ErrInvalidArgument
	}
	return nil
}

func authorityV7IntegrationRetirementForPair(values []contracts.AuthorityV7ProviderRetirementFactsV1, pair string) (contracts.AuthorityV7ProviderRetirementFactsV1, bool) {
	for _, value := range values {
		if authorityV7IntegrationProviderPair(value.ProviderIdentityDigest, value.ProviderEndpointIdentityDigest) == pair {
			return value, true
		}
	}
	return contracts.AuthorityV7ProviderRetirementFactsV1{}, false
}

func authorityV7IntegrationMembersForPair(values []contracts.AuthorityV7RetiredEnvironmentMemberFactsV1, pair string) []contracts.AuthorityV7RetiredEnvironmentMemberFactsV1 {
	result := make([]contracts.AuthorityV7RetiredEnvironmentMemberFactsV1, 0)
	for _, value := range values {
		if authorityV7IntegrationProviderPair(value.ProviderIdentityDigest, value.ProviderEndpointIdentityDigest) == pair {
			result = append(result, value)
		}
	}
	return result
}

func authorityV7IntegrationEnvironmentRecordDigests(values []contracts.AuthorityV7RetiredEnvironmentMemberFactsV1) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = hex.EncodeToString(value.EnvironmentRecordDigest[:])
	}
	return result
}

func authorityV7IntegrationCanonicalJSONValue(value any) ([]byte, error) {
	wire, err := authorityV7IntegrationWireValue(reflect.ValueOf(value))
	if err != nil {
		return nil, ErrInvalidArgument
	}
	raw, err := json.Marshal(wire)
	if err != nil {
		return nil, ErrInvalidArgument
	}
	canonical, err := jcs.Transform(raw)
	if err != nil {
		return nil, ErrInvalidArgument
	}
	return canonical, nil
}
