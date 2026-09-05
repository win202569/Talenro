package contracts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AuthorityV7InstallationKindV1 string
type AuthorityV7ProtocolProfileV1 string
type AuthorityV7MigrationDownStateV1 string
type AuthorityV7UpgradeClassificationStateV1 string
type AuthorityV7DowngradeAuthorizationScopeV1 string
type AuthorityV7DowngradeStageV1 string
type FreshRestoreImportObjectTypeV1 string
type FreshRestoreProviderPhaseV1 string
type FreshRestoreStagingExclusionStateV1 string

const (
	AuthorityV7InstallationKindProduction        AuthorityV7InstallationKindV1 = "production"
	AuthorityV7InstallationKindDisposableFixture AuthorityV7InstallationKindV1 = "disposable_fixture"

	AuthorityV7ProtocolProfileLegacyV6 AuthorityV7ProtocolProfileV1 = "legacy_v6"
	AuthorityV7ProtocolProfileClaimV1  AuthorityV7ProtocolProfileV1 = "claim_v1"

	AuthorityV7MigrationDownStateLocked AuthorityV7MigrationDownStateV1 = "locked"

	AuthorityV7UpgradeClassificationPending AuthorityV7UpgradeClassificationStateV1 = "pending"

	AuthorityV7DowngradeAuthorizationScopeDown00007Only AuthorityV7DowngradeAuthorizationScopeV1 = "down_00007_only"
	AuthorityV7DowngradeStagePreAuthorization           AuthorityV7DowngradeStageV1              = "pre_authorization"

	FreshRestoreImportObjectPOP                       FreshRestoreImportObjectTypeV1 = "pop"
	FreshRestoreImportObjectFailureDomainDefinition   FreshRestoreImportObjectTypeV1 = "failure_domain_definition"
	FreshRestoreImportObjectCapacityProfileDefinition FreshRestoreImportObjectTypeV1 = "capacity_profile_definition"
	FreshRestoreImportObjectNodeReconstructionSeed    FreshRestoreImportObjectTypeV1 = "node_reconstruction_seed"

	FreshRestoreProviderPhaseStagingClosed FreshRestoreProviderPhaseV1         = "fresh_v7_staging_closed"
	FreshRestoreStagingExclusionHeld       FreshRestoreStagingExclusionStateV1 = "held"
)

type AuthorityV7MigrationLatchFactsV1 struct {
	InstallationID         uuid.UUID                       `json:"installation_id"`
	InstallationKind       AuthorityV7InstallationKindV1   `json:"installation_kind"`
	MigrationVersion       uint64                          `json:"migration_version"`
	DatabaseIdentityDigest Digest                          `json:"database_identity_digest"`
	UpCatalogDigest        Digest                          `json:"up_catalog_digest"`
	DownState              AuthorityV7MigrationDownStateV1 `json:"down_state"`
	InstalledAt            time.Time                       `json:"installed_at"`
}

type AuthorityV7UpgradeIntentFactsV1 struct {
	IntentID                    uuid.UUID                               `json:"intent_id"`
	InstallationID              uuid.UUID                               `json:"installation_id"`
	InstallationKind            AuthorityV7InstallationKindV1           `json:"installation_kind"`
	ActivationID                uuid.UUID                               `json:"activation_id"`
	RequestNonce                Digest                                  `json:"request_nonce"`
	CredentialPolicyUpdateID    *uuid.UUID                              `json:"credential_policy_update_id"`
	IncarnationRegistrationID   uuid.UUID                               `json:"incarnation_registration_id"`
	ProviderAbsenceProofID      uuid.UUID                               `json:"provider_absence_proof_id"`
	ObservedDeploymentID        uuid.UUID                               `json:"observed_deployment_id"`
	DatabaseIdentityDigest      Digest                                  `json:"database_identity_digest"`
	LocalRuntimeIsolationDigest Digest                                  `json:"local_runtime_isolation_digest"`
	ClassificationState         AuthorityV7UpgradeClassificationStateV1 `json:"classification_state"`
	CreatedAt                   time.Time                               `json:"created_at"`
}

type AuthorityV7StableTableInventoryItemV1 struct {
	TableName      string `json:"table_name"`
	Classification string `json:"classification"`
	RowCount       uint64 `json:"row_count"`
	ContentDigest  Digest `json:"content_digest"`
}

type AuthorityV7ProviderRetirementFactsV1 struct {
	ProviderIdentityDigest                    Digest `json:"provider_identity_digest"`
	ProviderEndpointIdentityDigest            Digest `json:"provider_endpoint_identity_digest"`
	ProviderProtocolDowngradeRetirementDigest Digest `json:"provider_protocol_downgrade_retirement_digest"`
}

type AuthorityV7RetiredEnvironmentMemberFactsV1 struct {
	EnvironmentRecordDigest        Digest    `json:"environment_record_digest"`
	EnvironmentAttestationDigest   Digest    `json:"environment_attestation_digest"`
	DeploymentID                   uuid.UUID `json:"deployment_id"`
	PostgresSystemID               uint64    `json:"postgres_system_id"`
	Timeline                       uint64    `json:"timeline"`
	DatabaseOID                    uint64    `json:"database_oid"`
	DatabaseName                   string    `json:"database_name"`
	DatabaseIdentityDigest         Digest    `json:"database_identity_digest"`
	EnvironmentInstanceGeneration  uint64    `json:"environment_instance_generation"`
	ProviderIdentityDigest         Digest    `json:"provider_identity_digest"`
	ProviderEndpointIdentityDigest Digest    `json:"provider_endpoint_identity_digest"`
	ProviderNamespace              string    `json:"provider_namespace"`
	ProviderProfile                string    `json:"provider_profile"`
}

type AuthorityV7ProviderRetirementSetFactsV1 struct {
	InstallationID                                 uuid.UUID                                    `json:"installation_id"`
	ReleaseScope                                   string                                       `json:"release_scope"`
	EnvironmentInventoryDigest                     Digest                                       `json:"environment_inventory_digest"`
	EnvironmentInventoryAnchorSetDigest            Digest                                       `json:"environment_inventory_anchor_set_digest"`
	EnvironmentInventoryMembershipRetirementDigest Digest                                       `json:"environment_inventory_membership_retirement_digest"`
	ProviderCount                                  uint64                                       `json:"provider_count"`
	Retirements                                    []AuthorityV7ProviderRetirementFactsV1       `json:"retirements"`
	RetiredMemberCount                             uint64                                       `json:"retired_member_count"`
	RetiredMembers                                 []AuthorityV7RetiredEnvironmentMemberFactsV1 `json:"retired_members"`
	CreatedAt                                      time.Time                                    `json:"created_at"`
}

type AuthorityV7PristineDowngradeInventoryFactsV1 struct {
	InstallationID              uuid.UUID                               `json:"installation_id"`
	InstallationKind            AuthorityV7InstallationKindV1           `json:"installation_kind"`
	DatabaseIdentityDigest      Digest                                  `json:"database_identity_digest"`
	MigrationLatchDigest        Digest                                  `json:"migration_latch_digest"`
	CurrentCatalogDigest        Digest                                  `json:"current_catalog_digest"`
	ManifestID                  uuid.UUID                               `json:"manifest_id"`
	ManifestDigest              Digest                                  `json:"manifest_digest"`
	StableTableCount            uint64                                  `json:"stable_table_count"`
	StableTableInventory        []AuthorityV7StableTableInventoryItemV1 `json:"stable_table_inventory"`
	ControlTableCount           uint64                                  `json:"control_table_count"`
	NonControlProtocolRowCount  uint64                                  `json:"non_control_protocol_row_count"`
	MigrationLatchCount         uint64                                  `json:"migration_latch_count"`
	DowngradeAuthorizationCount uint64                                  `json:"downgrade_authorization_count"`
	DatabaseTransactionID       uint64                                  `json:"database_transaction_id"`
	TransactionNonce            Digest                                  `json:"transaction_nonce"`
	Stage                       AuthorityV7DowngradeStageV1             `json:"stage"`
	ObservedAt                  time.Time                               `json:"observed_at"`
}

type AuthorityV7ProtocolDowngradeAuthorizationFactsV1 struct {
	AuthorizationID                              uuid.UUID                                `json:"authorization_id"`
	InstallationID                               uuid.UUID                                `json:"installation_id"`
	MigrationLatchDigest                         Digest                                   `json:"migration_latch_digest"`
	DatabaseIdentityDigest                       Digest                                   `json:"database_identity_digest"`
	MigrationVersion                             uint64                                   `json:"migration_version"`
	CurrentCatalogDigest                         Digest                                   `json:"current_catalog_digest"`
	PristineDowngradeInventoryDigest             Digest                                   `json:"pristine_downgrade_inventory_digest"`
	ProviderProtocolDowngradeRetirementSetDigest Digest                                   `json:"provider_protocol_downgrade_retirement_set_digest"`
	EnvironmentInventoryAnchorSetDigest          Digest                                   `json:"environment_inventory_anchor_set_digest"`
	DatabaseTransactionID                        uint64                                   `json:"database_transaction_id"`
	TransactionNonce                             Digest                                   `json:"transaction_nonce"`
	AuthorizationScope                           AuthorityV7DowngradeAuthorizationScopeV1 `json:"authorization_scope"`
	IssuedAt                                     time.Time                                `json:"issued_at"`
	ExpiresAt                                    time.Time                                `json:"expires_at"`
}

type FreshRestoreImportManifestObjectV1 struct {
	ObjectType    FreshRestoreImportObjectTypeV1 `json:"object_type"`
	CanonicalKey  string                         `json:"canonical_key"`
	Payload       json.RawMessage                `json:"payload"`
	PayloadDigest Digest                         `json:"payload_digest"`
}

type FreshImportTopologyProjectionObjectV1 struct {
	ObjectType        FreshRestoreImportObjectTypeV1 `json:"object_type"`
	CanonicalKey      string                         `json:"canonical_key"`
	NormalizedPayload json.RawMessage                `json:"normalized_payload"`
}

type FreshRestoreImportManifestTopologyFactsV1 struct {
	ManifestID                                  uuid.UUID                            `json:"manifest_id"`
	SingleUseApplyID                            uuid.UUID                            `json:"single_use_apply_id"`
	ManifestDigest                              Digest                               `json:"manifest_digest"`
	TargetActivationID                          uuid.UUID                            `json:"target_activation_id"`
	TargetDeploymentID                          uuid.UUID                            `json:"target_deployment_id"`
	TargetDatabaseIncarnationRegistrationDigest Digest                               `json:"target_database_incarnation_registration_digest"`
	TargetEpochEvidenceDigest                   Digest                               `json:"target_epoch_evidence_digest"`
	IssuedAt                                    time.Time                            `json:"issued_at"`
	ExpiresAt                                   time.Time                            `json:"expires_at"`
	ObjectCount                                 uint64                               `json:"object_count"`
	Objects                                     []FreshRestoreImportManifestObjectV1 `json:"objects"`
	CompleteNodeSetDigest                       Digest                               `json:"complete_node_set_digest"`
	ForbiddenObjectClassSetDigest               Digest                               `json:"forbidden_object_class_set_digest"`
	ExpectedPostImportInventoryDigest           Digest                               `json:"expected_post_import_inventory_digest"`
}

type FreshRestoreStagingImportCapabilityFactsV1 struct {
	CapabilityDigest                            Digest                      `json:"capability_digest"`
	CapabilityID                                uuid.UUID                   `json:"capability_id"`
	SingleUseApplyID                            uuid.UUID                   `json:"single_use_apply_id"`
	ManifestDigest                              Digest                      `json:"manifest_digest"`
	TargetActivationID                          uuid.UUID                   `json:"target_activation_id"`
	TargetDatabaseIdentityDigest                Digest                      `json:"target_database_identity_digest"`
	DatabaseTimelineLineageChainDigest          Digest                      `json:"database_timeline_lineage_chain_digest"`
	TargetDatabaseIncarnationRegistrationDigest Digest                      `json:"target_database_incarnation_registration_digest"`
	RuntimeRebindChainDigest                    Digest                      `json:"runtime_rebind_chain_digest"`
	RuntimeInstanceBindingDigest                Digest                      `json:"runtime_instance_binding_digest"`
	PlannedStagingExclusionID                   uuid.UUID                   `json:"planned_staging_exclusion_id"`
	ExpectedPreAcquireProviderHeadDigest        Digest                      `json:"expected_pre_acquire_provider_head_digest"`
	PreAcquireDatabaseIncarnationProofDigest    Digest                      `json:"pre_acquire_database_incarnation_proof_digest"`
	ProviderPhase                               FreshRestoreProviderPhaseV1 `json:"provider_phase"`
	ProviderServingLeaseAbsentDigest            Digest                      `json:"provider_serving_lease_absent_digest"`
	DatabaseRouteClosedDigest                   Digest                      `json:"database_route_closed_digest"`
	PreImportInventoryDigest                    Digest                      `json:"pre_import_inventory_digest"`
	AllowedObjectSetDigest                      Digest                      `json:"allowed_object_set_digest"`
	ExpectedPostImportInventoryDigest           Digest                      `json:"expected_post_import_inventory_digest"`
	TransactionNonce                            Digest                      `json:"transaction_nonce"`
	IssuedAt                                    time.Time                   `json:"issued_at"`
	ExpiresAt                                   time.Time                   `json:"expires_at"`
}

type FreshV7StagingExclusionLeaseFactsV1 struct {
	LeaseDigest                                   Digest                              `json:"lease_digest"`
	ExclusionID                                   uuid.UUID                           `json:"exclusion_id"`
	StagingImportCapabilityDigest                 Digest                              `json:"staging_import_capability_digest"`
	CapabilityRegistrationCommitChallengeDigest   Digest                              `json:"capability_registration_commit_challenge_digest"`
	CapabilityRegistrationCommitAttestationDigest Digest                              `json:"capability_registration_commit_attestation_digest"`
	TargetActivationID                            uuid.UUID                           `json:"target_activation_id"`
	TargetDatabaseIdentityDigest                  Digest                              `json:"target_database_identity_digest"`
	DatabaseTimelineLineageChainDigest            Digest                              `json:"database_timeline_lineage_chain_digest"`
	ExpectedProviderHeadDigest                    Digest                              `json:"expected_provider_head_digest"`
	PreAcquireDatabaseIncarnationProofDigest      Digest                              `json:"pre_acquire_database_incarnation_proof_digest"`
	TargetDatabaseIncarnationRegistrationDigest   Digest                              `json:"target_database_incarnation_registration_digest"`
	RuntimeRebindChainDigest                      Digest                              `json:"runtime_rebind_chain_digest"`
	RuntimeInstanceBindingDigest                  Digest                              `json:"runtime_instance_binding_digest"`
	RequestedAdmissionExpiresAt                   time.Time                           `json:"requested_admission_expires_at"`
	RequestNonce                                  Digest                              `json:"request_nonce"`
	RequestDigest                                 Digest                              `json:"request_digest"`
	ExclusionLeaseID                              uuid.UUID                           `json:"exclusion_lease_id"`
	AcquisitionLockedProviderHeadDigest           Digest                              `json:"acquisition_locked_provider_head_digest"`
	ProviderServingLeaseAbsentDigest              Digest                              `json:"provider_serving_lease_absent_digest"`
	ProviderControlSequence                       uint64                              `json:"provider_control_sequence"`
	ExclusionState                                FreshRestoreStagingExclusionStateV1 `json:"exclusion_state"`
	IssuedAt                                      time.Time                           `json:"issued_at"`
	AdmissionExpiresAt                            time.Time                           `json:"admission_expires_at"`
}

type FreshRestorePOPImportPayloadV1 struct {
	POPCode    string `json:"pop_code"`
	ISOCountry string `json:"iso_country"`
	Region     string `json:"region"`
}

type FreshRestoreFailureDomainImportPayloadV1 struct {
	FailureDomainID uuid.UUID `json:"failure_domain_id"`
	DomainType      string    `json:"domain_type"`
	StableID        string    `json:"stable_id"`
}

type FreshRestoreCapacityProfileImportPayloadV1 struct {
	ProfileID                  string   `json:"profile_id"`
	Version                    uint64   `json:"version"`
	Adapter                    string   `json:"adapter"`
	EgressLimitBPS             uint64   `json:"egress_limit_bps"`
	ConnectionLimit            uint64   `json:"connection_limit"`
	HandshakeLimitPerSecond    uint64   `json:"handshake_limit_per_second"`
	CPUQuotaMillicores         uint64   `json:"cpu_quota_millicores"`
	CPULimitBasisPoints        uint64   `json:"cpu_limit_basis_points"`
	MemoryLimitBytes           uint64   `json:"memory_limit_bytes"`
	TaskLimit                  uint64   `json:"task_limit"`
	FileDescriptorLimit        uint64   `json:"file_descriptor_limit"`
	QueueLimit                 uint64   `json:"queue_limit"`
	PacketLossLimitBasisPoints uint64   `json:"packet_loss_limit_basis_points"`
	RequiredMetrics            []string `json:"required_metrics"`
}

type FreshRestoreNodeReconstructionSeedPayloadV1 struct {
	NodeID  uuid.UUID `json:"node_id"`
	POPCode string    `json:"pop_code"`
}

type FreshRestoreNormalizedPOPPayloadV1 struct {
	POPCode       string `json:"pop_code"`
	ISOCountry    string `json:"iso_country"`
	Region        string `json:"region"`
	OperatorState string `json:"operator_state"`
	Version       string `json:"version"`
}

type FreshRestoreNormalizedFailureDomainPayloadV1 struct {
	FailureDomainID uuid.UUID `json:"failure_domain_id"`
	DomainType      string    `json:"domain_type"`
	StableID        string    `json:"stable_id"`
	Version         string    `json:"version"`
}

type FreshRestoreNormalizedNodePayloadV1 struct {
	NodeID                          uuid.UUID         `json:"node_id"`
	POPCode                         string            `json:"pop_code"`
	OperatorState                   string            `json:"operator_state"`
	SecurityState                   string            `json:"security_state"`
	IdentityState                   string            `json:"identity_state"`
	HealthState                     string            `json:"health_state"`
	IdentityEpoch                   string            `json:"identity_epoch"`
	InventoryVersion                string            `json:"inventory_version"`
	SecurityVersion                 string            `json:"security_version"`
	NextDesiredGeneration           string            `json:"next_desired_generation"`
	NextRecoveryGeneration          string            `json:"next_recovery_generation"`
	ResumeOperatorStateOrNull       *string           `json:"resume_operator_state_or_null"`
	PendingOperatorTransitionOrNull *string           `json:"pending_operator_transition_or_null"`
	ActivePointerSet                []json.RawMessage `json:"active_pointer_set"`
	AuthorityAnchorSet              []json.RawMessage `json:"authority_anchor_set"`
}

type AuthorityV7UpMigrationFactsV1 struct {
	MigrationLatch              AuthorityV7MigrationLatchFactsV1 `json:"migration_latch"`
	MigrationLatchDigest        Digest                           `json:"migration_latch_digest"`
	UpgradeIntentOrNull         *AuthorityV7UpgradeIntentFactsV1 `json:"upgrade_intent_or_null"`
	UpgradeIntentDigestOrNull   *Digest                          `json:"upgrade_intent_digest_or_null"`
	AuthorityProtocolProfile    AuthorityV7ProtocolProfileV1     `json:"authority_protocol_profile"`
	LocalRuntimeIsolationDigest Digest                           `json:"local_runtime_isolation_digest"`
	TransactionNonce            Digest                           `json:"transaction_nonce"`
	ExpiresAt                   time.Time                        `json:"expires_at"`
}

type AuthorityV7DownMigrationFactsV1 struct {
	MigrationLatch              AuthorityV7MigrationLatchFactsV1         `json:"migration_latch"`
	MigrationLatchDigest        Digest                                   `json:"migration_latch_digest"`
	AuthorityProtocolProfile    AuthorityV7ProtocolProfileV1             `json:"authority_protocol_profile"`
	CurrentCatalogDigest        Digest                                   `json:"current_catalog_digest"`
	ManifestID                  uuid.UUID                                `json:"manifest_id"`
	ManifestDigest              Digest                                   `json:"manifest_digest"`
	StableTableCount            uint64                                   `json:"stable_table_count"`
	StableTableInventory        []AuthorityV7StableTableInventoryItemV1  `json:"stable_table_inventory"`
	ProviderRetirementSet       AuthorityV7ProviderRetirementSetFactsV1  `json:"provider_retirement_set"`
	ProviderRetirementSetDigest Digest                                   `json:"provider_retirement_set_digest"`
	EnvironmentAnchorSetDigest  Digest                                   `json:"environment_anchor_set_digest"`
	AuthorizationID             uuid.UUID                                `json:"authorization_id"`
	AuthorizationScope          AuthorityV7DowngradeAuthorizationScopeV1 `json:"authorization_scope"`
	TransactionNonce            Digest                                   `json:"transaction_nonce"`
	ExpiresAt                   time.Time                                `json:"expires_at"`
}

type AuthorityV7DownAuthorizationRequestV1 struct {
	AuthorizationID             uuid.UUID                                    `json:"authorization_id"`
	MigrationLatch              AuthorityV7MigrationLatchFactsV1             `json:"migration_latch"`
	MigrationLatchDigest        Digest                                       `json:"migration_latch_digest"`
	CurrentCatalogDigest        Digest                                       `json:"current_catalog_digest"`
	PristineInventory           AuthorityV7PristineDowngradeInventoryFactsV1 `json:"pristine_inventory"`
	PristineInventoryDigest     Digest                                       `json:"pristine_inventory_digest"`
	ProviderRetirementSet       AuthorityV7ProviderRetirementSetFactsV1      `json:"provider_retirement_set"`
	ProviderRetirementSetDigest Digest                                       `json:"provider_retirement_set_digest"`
	EnvironmentAnchorSetDigest  Digest                                       `json:"environment_anchor_set_digest"`
	ActualDatabaseTransactionID uint64                                       `json:"actual_database_transaction_id"`
	TransactionNonce            Digest                                       `json:"transaction_nonce"`
	AuthorizationScope          AuthorityV7DowngradeAuthorizationScopeV1     `json:"authorization_scope"`
}

type AuthorityV7DownAuthorizationFactsV1 struct {
	Request                  AuthorityV7DownAuthorizationRequestV1            `json:"request"`
	Authorization            AuthorityV7ProtocolDowngradeAuthorizationFactsV1 `json:"authorization"`
	AuthorizationDigest      Digest                                           `json:"authorization_digest"`
	AuthorizationEnvelopeJCS []byte                                           `json:"authorization_envelope_jcs"`
	SignerRole               string                                           `json:"signer_role"`
	SignerKeyID              string                                           `json:"signer_key_id"`
	SignaturePolicyVersion   uint64                                           `json:"signature_policy_version"`
	TrustRootDigest          Digest                                           `json:"trust_root_digest"`
	SignatureAlgorithm       string                                           `json:"signature_algorithm"`
}

type FreshRestoreImportProjectionInputV1 struct {
	ManifestTopology                                FreshRestoreImportManifestTopologyFactsV1  `json:"manifest_topology"`
	StagingImportCapability                         FreshRestoreStagingImportCapabilityFactsV1 `json:"staging_import_capability"`
	StagingExclusionLease                           FreshV7StagingExclusionLeaseFactsV1        `json:"staging_exclusion_lease"`
	CurrentProviderHeadDigest                       Digest                                     `json:"current_provider_head_digest"`
	ProviderPhase                                   FreshRestoreProviderPhaseV1                `json:"provider_phase"`
	ServingLeaseAbsentDigest                        Digest                                     `json:"serving_lease_absent_digest"`
	StagingExclusionState                           FreshRestoreStagingExclusionStateV1        `json:"staging_exclusion_state"`
	DatabaseRouteClosedDigest                       Digest                                     `json:"database_route_closed_digest"`
	CurrentDatabaseIdentityDigest                   Digest                                     `json:"current_database_identity_digest"`
	DatabaseTimelineLineageChainDigest              Digest                                     `json:"database_timeline_lineage_chain_digest"`
	CurrentDatabaseIncarnationRegistrationDigest    Digest                                     `json:"current_database_incarnation_registration_digest"`
	LatestDatabaseAuthorityRebindResultDigestOrNull *Digest                                    `json:"latest_database_authority_rebind_result_digest_or_null"`
	RuntimeRebindChainDigest                        Digest                                     `json:"runtime_rebind_chain_digest"`
	RuntimeInstanceBindingDigest                    Digest                                     `json:"runtime_instance_binding_digest"`
	PreImportInventoryDigest                        Digest                                     `json:"pre_import_inventory_digest"`
	NormalizedCatalogDigest                         Digest                                     `json:"normalized_catalog_digest"`
	AdmissionCheckedAt                              time.Time                                  `json:"admission_checked_at"`
}

type FreshImportTopologyProjectionV1 struct {
	ProjectionVersion            uint64                                  `json:"projection_version"`
	TargetActivationID           uuid.UUID                               `json:"target_activation_id"`
	TargetDeploymentID           uuid.UUID                               `json:"target_deployment_id"`
	TargetDatabaseIdentityDigest Digest                                  `json:"target_database_identity_digest"`
	NormalizedCatalogDigest      Digest                                  `json:"normalized_catalog_digest"`
	ObjectCount                  uint64                                  `json:"object_count"`
	Objects                      []FreshImportTopologyProjectionObjectV1 `json:"objects"`
}

type FreshRestoreImportApplicationV1 struct {
	SingleUseApplyID                                       uuid.UUID `json:"single_use_apply_id"`
	StagingImportCapabilityDigest                          Digest    `json:"staging_import_capability_digest"`
	StagingImportCapabilityRecoveryIntentDigestOrNull      *Digest   `json:"staging_import_capability_recovery_intent_digest_or_null"`
	StagingImportCapabilityRecoveryApplicationDigestOrNull *Digest   `json:"staging_import_capability_recovery_application_digest_or_null"`
	ManifestDigest                                         Digest    `json:"manifest_digest"`
	TargetActivationID                                     uuid.UUID `json:"target_activation_id"`
	CurrentDatabaseIdentityDigest                          Digest    `json:"current_database_identity_digest"`
	DatabaseTimelineLineageChainDigest                     Digest    `json:"database_timeline_lineage_chain_digest"`
	TargetDatabaseIncarnationRegistrationDigest            Digest    `json:"target_database_incarnation_registration_digest"`
	RuntimeRebindChainDigest                               Digest    `json:"runtime_rebind_chain_digest"`
	RuntimeInstanceBindingDigest                           Digest    `json:"runtime_instance_binding_digest"`
	StagingExclusionLeaseDigest                            Digest    `json:"staging_exclusion_lease_digest"`
	AcquisitionLockedProviderHeadDigest                    Digest    `json:"acquisition_locked_provider_head_digest"`
	DatabaseRouteClosedDigest                              Digest    `json:"database_route_closed_digest"`
	PreImportInventoryDigest                               Digest    `json:"pre_import_inventory_digest"`
	PostImportInventoryDigest                              Digest    `json:"post_import_inventory_digest"`
	ImportedObjectCount                                    uint64    `json:"imported_object_count"`
	CompleteNodeSetDigest                                  Digest    `json:"complete_node_set_digest"`
	ForbiddenStateZeroDigest                               Digest    `json:"forbidden_state_zero_digest"`
	DatabaseTransactionID                                  uint64    `json:"database_transaction_id"`
	TransactionSnapshotDigest                              Digest    `json:"transaction_snapshot_digest"`
	AppliedAt                                              time.Time `json:"applied_at"`
}

var (
	freshPOPCodePattern   = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?$`)
	freshRegionPattern    = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,62}[a-z0-9])?$`)
	freshProfileIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
	freshStableIDPattern  = regexp.MustCompile(`^[!-~]{1,128}$`)
	freshNamePattern      = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/-]{0,252}$`)
)

func (value AuthorityV7MigrationLatchFactsV1) Validate() error {
	if value.InstallationID == uuid.Nil || !validInstallationKind(value.InstallationKind) || value.MigrationVersion != 7 ||
		zeroDigest(value.DatabaseIdentityDigest) || zeroDigest(value.UpCatalogDigest) || value.DownState != AuthorityV7MigrationDownStateLocked || value.InstalledAt.IsZero() {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func (value AuthorityV7UpgradeIntentFactsV1) Validate() error {
	if value.IntentID == uuid.Nil || value.InstallationID == uuid.Nil || value.InstallationKind != AuthorityV7InstallationKindProduction ||
		value.ActivationID == uuid.Nil || zeroDigest(value.RequestNonce) || value.IncarnationRegistrationID == uuid.Nil ||
		value.ProviderAbsenceProofID == uuid.Nil || value.ObservedDeploymentID == uuid.Nil || zeroDigest(value.DatabaseIdentityDigest) ||
		zeroDigest(value.LocalRuntimeIsolationDigest) || value.ClassificationState != AuthorityV7UpgradeClassificationPending || value.CreatedAt.IsZero() {
		return ErrInvalidAuthorityValue
	}
	if value.CredentialPolicyUpdateID != nil && *value.CredentialPolicyUpdateID == uuid.Nil {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func (value AuthorityV7StableTableInventoryItemV1) Validate() error {
	if len(value.TableName) < 3 || len(value.TableName) > 253 || strings.Count(value.TableName, ".") != 1 ||
		!freshNamePattern.MatchString(value.TableName) || len(value.Classification) == 0 || len(value.Classification) > 64 ||
		value.RowCount > math.MaxInt64 || zeroDigest(value.ContentDigest) {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func (value AuthorityV7ProviderRetirementFactsV1) Validate() error {
	if zeroDigest(value.ProviderIdentityDigest) || zeroDigest(value.ProviderEndpointIdentityDigest) || zeroDigest(value.ProviderProtocolDowngradeRetirementDigest) {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func (value AuthorityV7RetiredEnvironmentMemberFactsV1) Validate() error {
	if zeroDigest(value.EnvironmentRecordDigest) || zeroDigest(value.EnvironmentAttestationDigest) || value.DeploymentID == uuid.Nil ||
		value.PostgresSystemID == 0 || value.Timeline == 0 || value.Timeline > math.MaxUint32 || value.DatabaseOID == 0 || value.DatabaseOID > math.MaxUint32 ||
		!boundedASCII(value.DatabaseName, 1, 63) || zeroDigest(value.DatabaseIdentityDigest) || value.EnvironmentInstanceGeneration == 0 || value.EnvironmentInstanceGeneration > math.MaxInt64 ||
		zeroDigest(value.ProviderIdentityDigest) || zeroDigest(value.ProviderEndpointIdentityDigest) || !boundedASCII(value.ProviderNamespace, 1, 128) || !boundedASCII(value.ProviderProfile, 1, 128) {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func (value AuthorityV7ProviderRetirementSetFactsV1) Validate() error {
	if value.InstallationID == uuid.Nil || !boundedASCII(value.ReleaseScope, 1, 128) || zeroDigest(value.EnvironmentInventoryDigest) ||
		zeroDigest(value.EnvironmentInventoryAnchorSetDigest) || zeroDigest(value.EnvironmentInventoryMembershipRetirementDigest) ||
		value.ProviderCount == 0 || value.ProviderCount > math.MaxInt64 || value.Retirements == nil || value.ProviderCount != uint64(len(value.Retirements)) ||
		value.RetiredMemberCount == 0 || value.RetiredMemberCount > math.MaxInt64 || value.RetiredMembers == nil || value.RetiredMemberCount != uint64(len(value.RetiredMembers)) || value.CreatedAt.IsZero() {
		return ErrInvalidAuthorityValue
	}
	for index := range value.Retirements {
		if value.Retirements[index].Validate() != nil || index > 0 && compareRetirements(value.Retirements[index-1], value.Retirements[index]) >= 0 {
			return ErrInvalidAuthorityValue
		}
	}
	for index := range value.RetiredMembers {
		if value.RetiredMembers[index].Validate() != nil || index > 0 && compareRetiredMembers(value.RetiredMembers[index-1], value.RetiredMembers[index]) >= 0 {
			return ErrInvalidAuthorityValue
		}
	}
	type providerPair struct {
		identity Digest
		endpoint Digest
	}
	retirementPairs := make(map[providerPair]struct{}, len(value.Retirements))
	for _, retirement := range value.Retirements {
		retirementPairs[providerPair{identity: retirement.ProviderIdentityDigest, endpoint: retirement.ProviderEndpointIdentityDigest}] = struct{}{}
	}
	memberPairs := make(map[providerPair]struct{}, len(value.RetiredMembers))
	for _, member := range value.RetiredMembers {
		pair := providerPair{identity: member.ProviderIdentityDigest, endpoint: member.ProviderEndpointIdentityDigest}
		if _, present := retirementPairs[pair]; !present {
			return ErrInvalidAuthorityValue
		}
		memberPairs[pair] = struct{}{}
	}
	if len(memberPairs) != len(retirementPairs) {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func (value AuthorityV7PristineDowngradeInventoryFactsV1) Validate() error {
	if value.InstallationID == uuid.Nil || value.InstallationKind != AuthorityV7InstallationKindDisposableFixture ||
		zeroDigest(value.DatabaseIdentityDigest) || zeroDigest(value.MigrationLatchDigest) || zeroDigest(value.CurrentCatalogDigest) ||
		value.ManifestID == uuid.Nil || zeroDigest(value.ManifestDigest) || value.StableTableCount > math.MaxInt64 || value.StableTableInventory == nil ||
		value.StableTableCount != uint64(len(value.StableTableInventory)) || value.ControlTableCount != 2 || value.NonControlProtocolRowCount != 0 ||
		value.MigrationLatchCount != 1 || value.DowngradeAuthorizationCount != 0 || value.DatabaseTransactionID == 0 || zeroDigest(value.TransactionNonce) ||
		value.Stage != AuthorityV7DowngradeStagePreAuthorization || value.ObservedAt.IsZero() {
		return ErrInvalidAuthorityValue
	}
	for index := range value.StableTableInventory {
		if value.StableTableInventory[index].Validate() != nil || index > 0 && value.StableTableInventory[index-1].TableName >= value.StableTableInventory[index].TableName {
			return ErrInvalidAuthorityValue
		}
	}
	return nil
}

func (value AuthorityV7ProtocolDowngradeAuthorizationFactsV1) Validate() error {
	if value.AuthorizationID == uuid.Nil || value.InstallationID == uuid.Nil || zeroDigest(value.MigrationLatchDigest) || zeroDigest(value.DatabaseIdentityDigest) ||
		value.MigrationVersion != 7 || zeroDigest(value.CurrentCatalogDigest) || zeroDigest(value.PristineDowngradeInventoryDigest) ||
		zeroDigest(value.ProviderProtocolDowngradeRetirementSetDigest) || zeroDigest(value.EnvironmentInventoryAnchorSetDigest) ||
		value.DatabaseTransactionID == 0 || zeroDigest(value.TransactionNonce) || value.AuthorizationScope != AuthorityV7DowngradeAuthorizationScopeDown00007Only ||
		value.IssuedAt.IsZero() || value.ExpiresAt.IsZero() || !value.ExpiresAt.After(value.IssuedAt) || value.ExpiresAt.Sub(value.IssuedAt) > 2*time.Minute {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func (value FreshRestoreImportManifestObjectV1) Validate() error {
	if !validImportObjectType(value.ObjectType) || !boundedASCII(value.CanonicalKey, 1, 256) || len(value.Payload) == 0 || len(value.Payload) > 1<<20 || zeroDigest(value.PayloadDigest) {
		return ErrInvalidAuthorityValue
	}
	key, err := validateFreshPayload(value.ObjectType, value.Payload, false)
	if err != nil || key != value.CanonicalKey {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func (value FreshImportTopologyProjectionObjectV1) Validate() error {
	if !validImportObjectType(value.ObjectType) || !boundedASCII(value.CanonicalKey, 1, 256) || len(value.NormalizedPayload) == 0 || len(value.NormalizedPayload) > 1<<20 {
		return ErrInvalidAuthorityValue
	}
	key, err := validateFreshPayload(value.ObjectType, value.NormalizedPayload, true)
	if err != nil || key != value.CanonicalKey {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func (value FreshRestoreImportManifestTopologyFactsV1) Validate() error {
	if value.ManifestID == uuid.Nil || value.SingleUseApplyID == uuid.Nil || zeroDigest(value.ManifestDigest) || value.TargetActivationID == uuid.Nil ||
		value.TargetDeploymentID == uuid.Nil || zeroDigest(value.TargetDatabaseIncarnationRegistrationDigest) || zeroDigest(value.TargetEpochEvidenceDigest) ||
		value.IssuedAt.IsZero() || value.ExpiresAt.IsZero() || !value.ExpiresAt.After(value.IssuedAt) || value.ExpiresAt.Sub(value.IssuedAt) > 15*time.Minute ||
		value.ObjectCount == 0 || value.ObjectCount > math.MaxInt64 || value.Objects == nil || value.ObjectCount != uint64(len(value.Objects)) ||
		zeroDigest(value.CompleteNodeSetDigest) || zeroDigest(value.ForbiddenObjectClassSetDigest) || zeroDigest(value.ExpectedPostImportInventoryDigest) {
		return ErrInvalidAuthorityValue
	}
	for index := range value.Objects {
		if value.Objects[index].Validate() != nil || index > 0 && compareManifestObjects(value.Objects[index-1], value.Objects[index]) >= 0 {
			return ErrInvalidAuthorityValue
		}
	}
	return nil
}

func (value FreshRestoreStagingImportCapabilityFactsV1) Validate() error {
	if zeroDigest(value.CapabilityDigest) || value.CapabilityID == uuid.Nil || value.SingleUseApplyID == uuid.Nil || zeroDigest(value.ManifestDigest) ||
		value.TargetActivationID == uuid.Nil || zeroDigest(value.TargetDatabaseIdentityDigest) || zeroDigest(value.DatabaseTimelineLineageChainDigest) ||
		zeroDigest(value.TargetDatabaseIncarnationRegistrationDigest) || zeroDigest(value.RuntimeRebindChainDigest) || zeroDigest(value.RuntimeInstanceBindingDigest) ||
		value.PlannedStagingExclusionID == uuid.Nil || zeroDigest(value.ExpectedPreAcquireProviderHeadDigest) || zeroDigest(value.PreAcquireDatabaseIncarnationProofDigest) ||
		value.ProviderPhase != FreshRestoreProviderPhaseStagingClosed || zeroDigest(value.ProviderServingLeaseAbsentDigest) || zeroDigest(value.DatabaseRouteClosedDigest) ||
		zeroDigest(value.PreImportInventoryDigest) || zeroDigest(value.AllowedObjectSetDigest) || zeroDigest(value.ExpectedPostImportInventoryDigest) || zeroDigest(value.TransactionNonce) ||
		value.IssuedAt.IsZero() || value.ExpiresAt.IsZero() || !value.ExpiresAt.After(value.IssuedAt) || value.ExpiresAt.Sub(value.IssuedAt) > 5*time.Minute {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func (value FreshV7StagingExclusionLeaseFactsV1) Validate() error {
	if zeroDigest(value.LeaseDigest) || value.ExclusionID == uuid.Nil || zeroDigest(value.StagingImportCapabilityDigest) ||
		zeroDigest(value.CapabilityRegistrationCommitChallengeDigest) || zeroDigest(value.CapabilityRegistrationCommitAttestationDigest) || value.TargetActivationID == uuid.Nil ||
		zeroDigest(value.TargetDatabaseIdentityDigest) || zeroDigest(value.DatabaseTimelineLineageChainDigest) || zeroDigest(value.ExpectedProviderHeadDigest) ||
		zeroDigest(value.PreAcquireDatabaseIncarnationProofDigest) || zeroDigest(value.TargetDatabaseIncarnationRegistrationDigest) || zeroDigest(value.RuntimeRebindChainDigest) ||
		zeroDigest(value.RuntimeInstanceBindingDigest) || value.RequestedAdmissionExpiresAt.IsZero() || zeroDigest(value.RequestNonce) || zeroDigest(value.RequestDigest) ||
		value.ExclusionLeaseID == uuid.Nil || value.ExclusionLeaseID != value.ExclusionID || zeroDigest(value.AcquisitionLockedProviderHeadDigest) || zeroDigest(value.ProviderServingLeaseAbsentDigest) ||
		value.ProviderControlSequence == 0 || value.ProviderControlSequence > math.MaxInt64 || value.ExclusionState != FreshRestoreStagingExclusionHeld || value.IssuedAt.IsZero() ||
		value.AdmissionExpiresAt.IsZero() || !value.AdmissionExpiresAt.After(value.IssuedAt) || value.AdmissionExpiresAt.Sub(value.IssuedAt) > 5*time.Minute ||
		!value.AdmissionExpiresAt.Equal(value.RequestedAdmissionExpiresAt) {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func (value AuthorityV7UpMigrationFactsV1) Validate() error {
	if value.MigrationLatch.Validate() != nil || zeroDigest(value.MigrationLatchDigest) || zeroDigest(value.LocalRuntimeIsolationDigest) || zeroDigest(value.TransactionNonce) || value.ExpiresAt.IsZero() ||
		!value.ExpiresAt.After(value.MigrationLatch.InstalledAt) {
		return ErrInvalidAuthorityValue
	}
	switch value.MigrationLatch.InstallationKind {
	case AuthorityV7InstallationKindProduction:
		if value.AuthorityProtocolProfile != AuthorityV7ProtocolProfileClaimV1 || value.UpgradeIntentOrNull == nil || value.UpgradeIntentDigestOrNull == nil || zeroDigest(*value.UpgradeIntentDigestOrNull) ||
			value.UpgradeIntentOrNull.Validate() != nil || value.UpgradeIntentOrNull.InstallationID != value.MigrationLatch.InstallationID ||
			value.UpgradeIntentOrNull.DatabaseIdentityDigest != value.MigrationLatch.DatabaseIdentityDigest || value.UpgradeIntentOrNull.LocalRuntimeIsolationDigest != value.LocalRuntimeIsolationDigest ||
			value.UpgradeIntentOrNull.RequestNonce != value.TransactionNonce {
			return ErrInvalidAuthorityValue
		}
	case AuthorityV7InstallationKindDisposableFixture:
		if value.AuthorityProtocolProfile != AuthorityV7ProtocolProfileLegacyV6 || value.UpgradeIntentOrNull != nil || value.UpgradeIntentDigestOrNull != nil {
			return ErrInvalidAuthorityValue
		}
	default:
		return ErrInvalidAuthorityValue
	}
	return nil
}

func (value AuthorityV7DownMigrationFactsV1) Validate() error {
	if value.MigrationLatch.Validate() != nil || value.MigrationLatch.InstallationKind != AuthorityV7InstallationKindDisposableFixture || zeroDigest(value.MigrationLatchDigest) ||
		value.AuthorityProtocolProfile != AuthorityV7ProtocolProfileLegacyV6 || zeroDigest(value.CurrentCatalogDigest) || value.ManifestID == uuid.Nil || zeroDigest(value.ManifestDigest) ||
		value.StableTableCount > math.MaxInt64 || value.StableTableInventory == nil || value.StableTableCount != uint64(len(value.StableTableInventory)) ||
		value.ProviderRetirementSet.Validate() != nil || value.ProviderRetirementSet.InstallationID != value.MigrationLatch.InstallationID || zeroDigest(value.ProviderRetirementSetDigest) ||
		zeroDigest(value.EnvironmentAnchorSetDigest) || value.EnvironmentAnchorSetDigest != value.ProviderRetirementSet.EnvironmentInventoryAnchorSetDigest ||
		value.AuthorizationID == uuid.Nil || value.AuthorizationScope != AuthorityV7DowngradeAuthorizationScopeDown00007Only || zeroDigest(value.TransactionNonce) || value.ExpiresAt.IsZero() ||
		!value.ExpiresAt.After(value.MigrationLatch.InstalledAt) {
		return ErrInvalidAuthorityValue
	}
	for index := range value.StableTableInventory {
		if value.StableTableInventory[index].Validate() != nil || index > 0 && value.StableTableInventory[index-1].TableName >= value.StableTableInventory[index].TableName {
			return ErrInvalidAuthorityValue
		}
	}
	return nil
}

func (value AuthorityV7DownAuthorizationRequestV1) Validate() error {
	if value.AuthorizationID == uuid.Nil || value.MigrationLatch.Validate() != nil || value.MigrationLatch.InstallationKind != AuthorityV7InstallationKindDisposableFixture ||
		zeroDigest(value.MigrationLatchDigest) || zeroDigest(value.CurrentCatalogDigest) || value.PristineInventory.Validate() != nil || zeroDigest(value.PristineInventoryDigest) ||
		value.ProviderRetirementSet.Validate() != nil || zeroDigest(value.ProviderRetirementSetDigest) || zeroDigest(value.EnvironmentAnchorSetDigest) || value.ActualDatabaseTransactionID == 0 ||
		zeroDigest(value.TransactionNonce) || value.AuthorizationScope != AuthorityV7DowngradeAuthorizationScopeDown00007Only ||
		value.PristineInventory.InstallationID != value.MigrationLatch.InstallationID || value.PristineInventory.DatabaseIdentityDigest != value.MigrationLatch.DatabaseIdentityDigest ||
		value.PristineInventory.MigrationLatchDigest != value.MigrationLatchDigest || value.PristineInventory.CurrentCatalogDigest != value.CurrentCatalogDigest ||
		value.PristineInventory.DatabaseTransactionID != value.ActualDatabaseTransactionID || value.PristineInventory.TransactionNonce != value.TransactionNonce ||
		value.ProviderRetirementSet.InstallationID != value.MigrationLatch.InstallationID || value.ProviderRetirementSet.EnvironmentInventoryAnchorSetDigest != value.EnvironmentAnchorSetDigest {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func (value AuthorityV7DownAuthorizationFactsV1) Validate() error {
	if value.Request.Validate() != nil || value.Authorization.Validate() != nil || zeroDigest(value.AuthorizationDigest) || len(value.AuthorizationEnvelopeJCS) == 0 || len(value.AuthorizationEnvelopeJCS) > 1<<20 ||
		value.SignerRole != "authority_protocol_downgrade_authorizer" || !boundedASCII(value.SignerKeyID, 1, 256) || value.SignaturePolicyVersion == 0 || value.SignaturePolicyVersion > math.MaxInt64 ||
		zeroDigest(value.TrustRootDigest) || !boundedASCII(value.SignatureAlgorithm, 1, 64) {
		return ErrInvalidAuthorityValue
	}
	authorization := value.Authorization
	request := value.Request
	if authorization.AuthorizationID != request.AuthorizationID || authorization.InstallationID != request.MigrationLatch.InstallationID ||
		authorization.MigrationLatchDigest != request.MigrationLatchDigest || authorization.DatabaseIdentityDigest != request.MigrationLatch.DatabaseIdentityDigest ||
		authorization.CurrentCatalogDigest != request.CurrentCatalogDigest || authorization.PristineDowngradeInventoryDigest != request.PristineInventoryDigest ||
		authorization.ProviderProtocolDowngradeRetirementSetDigest != request.ProviderRetirementSetDigest || authorization.EnvironmentInventoryAnchorSetDigest != request.EnvironmentAnchorSetDigest ||
		authorization.DatabaseTransactionID != request.ActualDatabaseTransactionID || authorization.TransactionNonce != request.TransactionNonce || authorization.AuthorizationScope != request.AuthorizationScope {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func (value FreshRestoreImportProjectionInputV1) Validate() error {
	if value.ManifestTopology.Validate() != nil || value.StagingImportCapability.Validate() != nil || value.StagingExclusionLease.Validate() != nil ||
		zeroDigest(value.CurrentProviderHeadDigest) || value.ProviderPhase != FreshRestoreProviderPhaseStagingClosed || zeroDigest(value.ServingLeaseAbsentDigest) ||
		value.StagingExclusionState != FreshRestoreStagingExclusionHeld || zeroDigest(value.DatabaseRouteClosedDigest) || zeroDigest(value.CurrentDatabaseIdentityDigest) ||
		zeroDigest(value.DatabaseTimelineLineageChainDigest) || zeroDigest(value.CurrentDatabaseIncarnationRegistrationDigest) ||
		value.LatestDatabaseAuthorityRebindResultDigestOrNull != nil && zeroDigest(*value.LatestDatabaseAuthorityRebindResultDigestOrNull) ||
		zeroDigest(value.RuntimeRebindChainDigest) || zeroDigest(value.RuntimeInstanceBindingDigest) || zeroDigest(value.PreImportInventoryDigest) || zeroDigest(value.NormalizedCatalogDigest) || value.AdmissionCheckedAt.IsZero() {
		return ErrInvalidAuthorityValue
	}
	manifest, capability, lease := value.ManifestTopology, value.StagingImportCapability, value.StagingExclusionLease
	if manifest.SingleUseApplyID != capability.SingleUseApplyID || manifest.ManifestDigest != capability.ManifestDigest || manifest.TargetActivationID != capability.TargetActivationID ||
		manifest.ExpectedPostImportInventoryDigest != capability.ExpectedPostImportInventoryDigest || manifest.TargetDatabaseIncarnationRegistrationDigest != capability.TargetDatabaseIncarnationRegistrationDigest ||
		capability.CapabilityDigest != lease.StagingImportCapabilityDigest || capability.TargetActivationID != lease.TargetActivationID || capability.TargetDatabaseIdentityDigest != lease.TargetDatabaseIdentityDigest ||
		capability.DatabaseTimelineLineageChainDigest != lease.DatabaseTimelineLineageChainDigest || capability.TargetDatabaseIncarnationRegistrationDigest != lease.TargetDatabaseIncarnationRegistrationDigest ||
		capability.RuntimeRebindChainDigest != lease.RuntimeRebindChainDigest || capability.RuntimeInstanceBindingDigest != lease.RuntimeInstanceBindingDigest ||
		capability.PlannedStagingExclusionID != lease.ExclusionID || capability.ExpectedPreAcquireProviderHeadDigest != lease.ExpectedProviderHeadDigest ||
		capability.PreAcquireDatabaseIncarnationProofDigest != lease.PreAcquireDatabaseIncarnationProofDigest ||
		capability.ExpectedPreAcquireProviderHeadDigest == value.CurrentProviderHeadDigest || lease.AcquisitionLockedProviderHeadDigest != value.CurrentProviderHeadDigest ||
		capability.ProviderPhase != value.ProviderPhase || lease.ProviderServingLeaseAbsentDigest != value.ServingLeaseAbsentDigest ||
		lease.ExclusionState != value.StagingExclusionState || capability.DatabaseRouteClosedDigest != value.DatabaseRouteClosedDigest || capability.TargetDatabaseIdentityDigest != value.CurrentDatabaseIdentityDigest ||
		capability.DatabaseTimelineLineageChainDigest != value.DatabaseTimelineLineageChainDigest || capability.TargetDatabaseIncarnationRegistrationDigest != value.CurrentDatabaseIncarnationRegistrationDigest ||
		capability.RuntimeRebindChainDigest != value.RuntimeRebindChainDigest || capability.RuntimeInstanceBindingDigest != value.RuntimeInstanceBindingDigest ||
		capability.PreImportInventoryDigest != value.PreImportInventoryDigest || !value.AdmissionCheckedAt.Before(lease.AdmissionExpiresAt) {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func (value FreshImportTopologyProjectionV1) Validate() error {
	if value.ProjectionVersion != 1 || value.TargetActivationID == uuid.Nil || value.TargetDeploymentID == uuid.Nil || zeroDigest(value.TargetDatabaseIdentityDigest) ||
		zeroDigest(value.NormalizedCatalogDigest) || value.ObjectCount > math.MaxInt64 || value.Objects == nil || value.ObjectCount != uint64(len(value.Objects)) {
		return ErrInvalidAuthorityValue
	}
	for index := range value.Objects {
		if value.Objects[index].Validate() != nil || index > 0 && compareProjectionObjects(value.Objects[index-1], value.Objects[index]) >= 0 {
			return ErrInvalidAuthorityValue
		}
	}
	return nil
}

func (value FreshRestoreImportApplicationV1) Validate() error {
	if value.SingleUseApplyID == uuid.Nil || zeroDigest(value.StagingImportCapabilityDigest) || value.StagingImportCapabilityRecoveryIntentDigestOrNull != nil ||
		value.StagingImportCapabilityRecoveryApplicationDigestOrNull != nil || zeroDigest(value.ManifestDigest) || value.TargetActivationID == uuid.Nil ||
		zeroDigest(value.CurrentDatabaseIdentityDigest) || zeroDigest(value.DatabaseTimelineLineageChainDigest) || zeroDigest(value.TargetDatabaseIncarnationRegistrationDigest) ||
		zeroDigest(value.RuntimeRebindChainDigest) || zeroDigest(value.RuntimeInstanceBindingDigest) || zeroDigest(value.StagingExclusionLeaseDigest) ||
		zeroDigest(value.AcquisitionLockedProviderHeadDigest) || zeroDigest(value.DatabaseRouteClosedDigest) || zeroDigest(value.PreImportInventoryDigest) ||
		zeroDigest(value.PostImportInventoryDigest) || value.ImportedObjectCount == 0 || value.ImportedObjectCount > math.MaxInt64 || zeroDigest(value.CompleteNodeSetDigest) ||
		zeroDigest(value.ForbiddenStateZeroDigest) || value.DatabaseTransactionID == 0 || zeroDigest(value.TransactionSnapshotDigest) || value.AppliedAt.IsZero() {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func validInstallationKind(value AuthorityV7InstallationKindV1) bool {
	return value == AuthorityV7InstallationKindProduction || value == AuthorityV7InstallationKindDisposableFixture
}

func validImportObjectType(value FreshRestoreImportObjectTypeV1) bool {
	return value == FreshRestoreImportObjectPOP || value == FreshRestoreImportObjectFailureDomainDefinition ||
		value == FreshRestoreImportObjectCapacityProfileDefinition || value == FreshRestoreImportObjectNodeReconstructionSeed
}

func zeroDigest(value Digest) bool { return value == (Digest{}) }

func boundedASCII(value string, minimum, maximum int) bool {
	if len(value) < minimum || len(value) > maximum {
		return false
	}
	for index := range len(value) {
		if value[index] < 0x20 || value[index] > 0x7e {
			return false
		}
	}
	return true
}

func compareRetirements(left, right AuthorityV7ProviderRetirementFactsV1) int {
	if comparison := bytes.Compare(left.ProviderIdentityDigest[:], right.ProviderIdentityDigest[:]); comparison != 0 {
		return comparison
	}
	return bytes.Compare(left.ProviderEndpointIdentityDigest[:], right.ProviderEndpointIdentityDigest[:])
}

func compareRetiredMembers(left, right AuthorityV7RetiredEnvironmentMemberFactsV1) int {
	leftKey := fmt.Sprintf("%s\x00%020d\x00%010d\x00%010d\x00%s\x00%x\x00%s", left.DeploymentID, left.PostgresSystemID, left.Timeline, left.DatabaseOID, left.DatabaseName, left.ProviderIdentityDigest, left.ProviderNamespace)
	rightKey := fmt.Sprintf("%s\x00%020d\x00%010d\x00%010d\x00%s\x00%x\x00%s", right.DeploymentID, right.PostgresSystemID, right.Timeline, right.DatabaseOID, right.DatabaseName, right.ProviderIdentityDigest, right.ProviderNamespace)
	return strings.Compare(leftKey, rightKey)
}

func compareManifestObjects(left, right FreshRestoreImportManifestObjectV1) int {
	if comparison := strings.Compare(string(left.ObjectType), string(right.ObjectType)); comparison != 0 {
		return comparison
	}
	return strings.Compare(left.CanonicalKey, right.CanonicalKey)
}

func compareProjectionObjects(left, right FreshImportTopologyProjectionObjectV1) int {
	if comparison := strings.Compare(string(left.ObjectType), string(right.ObjectType)); comparison != 0 {
		return comparison
	}
	return strings.Compare(left.CanonicalKey, right.CanonicalKey)
}

func validateFreshPayload(objectType FreshRestoreImportObjectTypeV1, raw json.RawMessage, normalized bool) (string, error) {
	if err := rejectDuplicateJSONKeys(raw); err != nil {
		return "", err
	}
	switch objectType {
	case FreshRestoreImportObjectPOP:
		if normalized {
			var value FreshRestoreNormalizedPOPPayloadV1
			if strictDecodeJSON(raw, &value) != nil || !validPOP(value.POPCode, value.ISOCountry, value.Region) || value.OperatorState != "disabled" || value.Version != "1" {
				return "", ErrInvalidAuthorityValue
			}
			return value.POPCode, nil
		}
		var value FreshRestorePOPImportPayloadV1
		if strictDecodeJSON(raw, &value) != nil || !validPOP(value.POPCode, value.ISOCountry, value.Region) {
			return "", ErrInvalidAuthorityValue
		}
		return value.POPCode, nil
	case FreshRestoreImportObjectFailureDomainDefinition:
		if normalized {
			var value FreshRestoreNormalizedFailureDomainPayloadV1
			if strictDecodeJSON(raw, &value) != nil || !validFailureDomain(value.FailureDomainID, value.DomainType, value.StableID) || value.Version != "1" {
				return "", ErrInvalidAuthorityValue
			}
			return value.FailureDomainID.String(), nil
		}
		var value FreshRestoreFailureDomainImportPayloadV1
		if strictDecodeJSON(raw, &value) != nil || !validFailureDomain(value.FailureDomainID, value.DomainType, value.StableID) {
			return "", ErrInvalidAuthorityValue
		}
		return value.FailureDomainID.String(), nil
	case FreshRestoreImportObjectCapacityProfileDefinition:
		var value FreshRestoreCapacityProfileImportPayloadV1
		if strictDecodeJSON(raw, &value) != nil || !validCapacityProfile(value) {
			return "", ErrInvalidAuthorityValue
		}
		return value.ProfileID + "/" + strconv.FormatUint(value.Version, 10), nil
	case FreshRestoreImportObjectNodeReconstructionSeed:
		if normalized {
			var value FreshRestoreNormalizedNodePayloadV1
			if strictDecodeJSON(raw, &value) != nil || value.NodeID == uuid.Nil || !freshPOPCodePattern.MatchString(value.POPCode) ||
				value.OperatorState != "disabled" || value.SecurityState != "quarantined" || value.IdentityState != "unauthorized" || value.HealthState != "unknown" ||
				value.IdentityEpoch != "0" || value.InventoryVersion != "1" || value.SecurityVersion != "1" || value.NextDesiredGeneration != "1" || value.NextRecoveryGeneration != "1" ||
				value.ResumeOperatorStateOrNull != nil || value.PendingOperatorTransitionOrNull != nil || value.ActivePointerSet == nil || len(value.ActivePointerSet) != 0 ||
				value.AuthorityAnchorSet == nil || len(value.AuthorityAnchorSet) != 0 {
				return "", ErrInvalidAuthorityValue
			}
			return value.NodeID.String(), nil
		}
		var value FreshRestoreNodeReconstructionSeedPayloadV1
		if strictDecodeJSON(raw, &value) != nil || value.NodeID == uuid.Nil || !freshPOPCodePattern.MatchString(value.POPCode) {
			return "", ErrInvalidAuthorityValue
		}
		return value.NodeID.String(), nil
	default:
		return "", ErrInvalidAuthorityValue
	}
}

func validPOP(popCode, isoCountry, region string) bool {
	return freshPOPCodePattern.MatchString(popCode) && len(isoCountry) == 2 && isoCountry[0] >= 'A' && isoCountry[0] <= 'Z' && isoCountry[1] >= 'A' && isoCountry[1] <= 'Z' && freshRegionPattern.MatchString(region)
}

func validFailureDomain(id uuid.UUID, domainType, stableID string) bool {
	return id != uuid.Nil && (domainType == "facility" || domainType == "compute" || domainType == "upstream") && freshStableIDPattern.MatchString(stableID)
}

func validCapacityProfile(value FreshRestoreCapacityProfileImportPayloadV1) bool {
	if len(value.ProfileID) > 128 || !freshProfileIDPattern.MatchString(value.ProfileID) || value.Version == 0 || value.Version > math.MaxInt64 ||
		(value.Adapter != "fixture" && value.Adapter != "xray" && value.Adapter != "sing_box") || value.EgressLimitBPS < 1_000_000 || value.EgressLimitBPS > 1_000_000_000_000 ||
		value.ConnectionLimit < 1 || value.ConnectionLimit > 10_000_000 || value.HandshakeLimitPerSecond < 1 || value.HandshakeLimitPerSecond > 1_000_000 ||
		value.CPUQuotaMillicores < 100 || value.CPUQuotaMillicores > 64_000 || value.CPULimitBasisPoints < 1 || value.CPULimitBasisPoints > 10_000 ||
		value.MemoryLimitBytes < 67_108_864 || value.MemoryLimitBytes > 1_099_511_627_776 || value.TaskLimit < 32 || value.TaskLimit > 4_096 ||
		value.FileDescriptorLimit < 64 || value.FileDescriptorLimit > 1_000_000 || value.QueueLimit < 1 || value.QueueLimit > 1_000_000 ||
		value.PacketLossLimitBasisPoints < 1 || value.PacketLossLimitBasisPoints > 10_000 || value.RequiredMetrics == nil || len(value.RequiredMetrics) < 1 || len(value.RequiredMetrics) > 5 {
		return false
	}
	allowed := map[string]bool{"cpu_basis_points": true, "egress_bps": true, "memory_bytes": true, "open_file_descriptors": true, "task_count": true}
	for index, metric := range value.RequiredMetrics {
		if !allowed[metric] || index > 0 && value.RequiredMetrics[index-1] >= metric {
			return false
		}
	}
	return true
}

func strictDecodeJSON(raw []byte, destination any) error {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return ErrInvalidAuthorityValue
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return ErrInvalidAuthorityValue
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errorsIsEOF(err) {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func rejectDuplicateJSONKeys(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := inspectJSONValue(decoder); err != nil {
		return ErrInvalidAuthorityValue
	}
	if _, err := decoder.Token(); err != io.EOF {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func inspectJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]bool)
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok || seen[key] {
				return ErrInvalidAuthorityValue
			}
			seen[key] = true
			if err := inspectJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return ErrInvalidAuthorityValue
		}
	case '[':
		for decoder.More() {
			if err := inspectJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return ErrInvalidAuthorityValue
		}
	default:
		return ErrInvalidAuthorityValue
	}
	return nil
}

func errorsIsEOF(err error) bool { return err == io.EOF }
