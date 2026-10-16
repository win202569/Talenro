// Package contracts exposes wire contracts without granting authority.
package contracts

import "errors"

// AuthorityV7SchemaName is a schema/object label only. Membership is not
// authentication, authorization, evidence eligibility, or proof of a Go codec.
type AuthorityV7SchemaName struct {
	Schema     string
	ObjectName string
}

// ErrUnknownAuthorityV7SchemaName reports an exact schema absent from this
// name-only directory. Its message never includes caller input.
var ErrUnknownAuthorityV7SchemaName = errors.New("contracts: unknown authority v7 schema name")

var authorityV7SchemaNameTable = [...]AuthorityV7SchemaName{
	{Schema: "abort-fresh-v7-staging-exclusion-request.v1", ObjectName: "AbortFreshV7StagingExclusionRequestV1"},
	{Schema: "acquire-database-commit-challenge-request.v1", ObjectName: "AcquireDatabaseCommitChallengeRequestV1"},
	{Schema: "acquire-database-incarnation-proof-request.v1", ObjectName: "AcquireDatabaseIncarnationProofRequestV1"},
	{Schema: "acquire-database-incarnation-registration-challenge-request.v1", ObjectName: "AcquireDatabaseIncarnationRegistrationChallengeRequestV1"},
	{Schema: "acquire-fresh-v7-staging-exclusion-request.v1", ObjectName: "AcquireFreshV7StagingExclusionRequestV1"},
	{Schema: "acquire-staging-database-commit-challenge-request.v1", ObjectName: "AcquireStagingDatabaseCommitChallengeRequestV1"},
	{Schema: "authority-epoch-transition-application.v1", ObjectName: "AuthorityEpochTransitionApplicationV1"},
	{Schema: "authority-epoch-transition-cancellation.v1", ObjectName: "AuthorityEpochTransitionCancellationV1"},
	{Schema: "authority-epoch-transition-intent.v1", ObjectName: "AuthorityEpochTransitionIntentV1"},
	{Schema: "authority-epoch-transition-recovery-application.v1", ObjectName: "AuthorityEpochTransitionRecoveryApplicationV1"},
	{Schema: "authority-epoch-transition-recovery-intent.v1", ObjectName: "AuthorityEpochTransitionRecoveryIntentV1"},
	{Schema: "authority-epoch-transition-recovery-prefix-decision.v1", ObjectName: "AuthorityEpochTransitionRecoveryPrefixDecisionV1"},
	{Schema: "authority-epoch-transition-recovery.v1", ObjectName: "AuthorityEpochTransitionRecoveryV1"},
	{Schema: "authority-epoch-transition-resolution-receipt.v1", ObjectName: "AuthorityEpochTransitionResolutionReceiptV1"},
	{Schema: "authority-epoch-transition-resolution.v1", ObjectName: "AuthorityEpochTransitionResolutionV1"},
	{Schema: "authority-epoch-transition-terminal-application.v1", ObjectName: "AuthorityEpochTransitionTerminalApplicationV1"},
	{Schema: "authority-epoch-transition.v1", ObjectName: "AuthorityEpochTransitionV1"},
	{Schema: "authority-genesis-epoch-evidence.v1", ObjectName: "AuthorityGenesisEpochEvidenceV1"},
	{Schema: "authority-protocol-activation-completion.v1", ObjectName: "AuthorityProtocolActivationCompletionV1"},
	{Schema: "authority-protocol-activation-release.v1", ObjectName: "AuthorityProtocolActivationReleaseV1"},
	{Schema: "authority-protocol-activation.v1", ObjectName: "AuthorityProtocolActivationV1"},
	{Schema: "authority-protocol-context.v1", ObjectName: "AuthorityProtocolContextV1"},
	{Schema: "authority-protocol-downgrade-authorization.v1", ObjectName: "AuthorityProtocolDowngradeAuthorizationV1"},
	{Schema: "authority-protocol-migration-latch.v1", ObjectName: "AuthorityProtocolMigrationLatchV1"},
	{Schema: "authority-protocol-upgrade-attempt.v1", ObjectName: "AuthorityProtocolUpgradeAttemptV1"},
	{Schema: "authority-protocol-upgrade-intent.v1", ObjectName: "AuthorityProtocolUpgradeIntentV1"},
	{Schema: "authority-runtime-rebind-result.v1", ObjectName: "AuthorityRuntimeRebindResultV1"},
	{Schema: "authority-runtime-registration-result.v1", ObjectName: "AuthorityRuntimeRegistrationResultV1"},
	{Schema: "canonical-evidence-bundle.v1", ObjectName: "CanonicalEvidenceBundleV1"},
	{Schema: "claim-v1-abort-request.v1", ObjectName: "ClaimV1AbortRequestV1"},
	{Schema: "claim-v1-credential-policy.v1", ObjectName: "ClaimV1CredentialPolicyV1"},
	{Schema: "claim-v1-finalize-request.v1", ObjectName: "ClaimV1FinalizeRequestV1"},
	{Schema: "claim-v1-genesis-completion.v1", ObjectName: "ClaimV1GenesisCompletionV1"},
	{Schema: "claim-v1-genesis-open.v1", ObjectName: "ClaimV1GenesisOpenV1"},
	{Schema: "claim-v1-genesis-preparation.v1", ObjectName: "ClaimV1GenesisPreparationV1"},
	{Schema: "claim-v1-genesis-release-preparation.v1", ObjectName: "ClaimV1GenesisReleasePreparationV1"},
	{Schema: "claim-v1-provider-head.v1", ObjectName: "ClaimV1ProviderHeadV1"},
	{Schema: "claim-v1-receipt.v1", ObjectName: "ClaimV1ReceiptV1"},
	{Schema: "claim-v1-record.v1", ObjectName: "ClaimV1RecordV1"},
	{Schema: "claim-v1-reservation.v1", ObjectName: "ClaimV1ReservationV1"},
	{Schema: "claim-v1-reserve-request.v1", ObjectName: "ClaimV1ReserveRequestV1"},
	{Schema: "claim-v1-runtime-instance-rebind.v1", ObjectName: "ClaimV1RuntimeInstanceRebindV1"},
	{Schema: "claim-v1-serving-lease-revocation.v1", ObjectName: "ClaimV1ServingLeaseRevocationV1"},
	{Schema: "claim-v1-serving-lease.v1", ObjectName: "ClaimV1ServingLeaseV1"},
	{Schema: "complete-claim-v1-genesis-request.v1", ObjectName: "CompleteClaimV1GenesisRequestV1"},
	{Schema: "database-authority-head-attestation.v1", ObjectName: "DatabaseAuthorityHeadAttestationV1"},
	{Schema: "database-authority-head.v1", ObjectName: "DatabaseAuthorityHeadV1"},
	{Schema: "database-authority-post-recovery-rebind-attestation.v1", ObjectName: "DatabaseAuthorityPostRecoveryRebindAttestationV1"},
	{Schema: "database-authority-rebind-gap-attestation.v1", ObjectName: "DatabaseAuthorityRebindGapAttestationV1"},
	{Schema: "database-commit-archive-fence-semantic-id-journal-record.v1", ObjectName: "FenceDatabaseCommitAttestationArchiveSemanticIDJournalRecordV1"},
	{Schema: "database-commit-attestation.v1", ObjectName: "DatabaseCommitAttestationV1"},
	{Schema: "database-commit-challenge.v1", ObjectName: "DatabaseCommitChallengeV1"},
	{Schema: "database-incarnation-attestation.v1", ObjectName: "DatabaseIncarnationAttestationV1"},
	{Schema: "database-incarnation-proof.v1", ObjectName: "DatabaseIncarnationProofV1"},
	{Schema: "database-incarnation-rebinding-registration.v1", ObjectName: "DatabaseIncarnationRebindingRegistrationV1"},
	{Schema: "database-incarnation-registration-challenge.v1", ObjectName: "DatabaseIncarnationRegistrationChallengeV1"},
	{Schema: "database-incarnation-registration.v1", ObjectName: "DatabaseIncarnationRegistrationV1"},
	{Schema: "database-legacy-absence-observation.v1", ObjectName: "DatabaseLegacyAbsenceObservationV1"},
	{Schema: "database-legacy-absence-projection.v1", ObjectName: "DatabaseLegacyAbsenceProjectionV1"},
	{Schema: "database-runtime-instance-binding.v1", ObjectName: "DatabaseRuntimeInstanceBindingV1"},
	{Schema: "database-runtime-instance-lease-descendant-chain.v1", ObjectName: "DatabaseRuntimeInstanceLeaseDescendantChainV1"},
	{Schema: "database-runtime-instance-lease-termination.v1", ObjectName: "DatabaseRuntimeInstanceLeaseTerminationV1"},
	{Schema: "database-runtime-instance-lease.v1", ObjectName: "DatabaseRuntimeInstanceLeaseV1"},
	{Schema: "database-runtime-rebind-authorization.v1", ObjectName: "DatabaseRuntimeRebindAuthorizationV1"},
	{Schema: "database-timeline-lineage-attestation.v1", ObjectName: "DatabaseTimelineLineageAttestationV1"},
	{Schema: "deployment-environment-attestation.v1", ObjectName: "DeploymentEnvironmentAttestationV1"},
	{Schema: "empty-database-inventory.v1", ObjectName: "EmptyDatabaseInventoryV1"},
	{Schema: "environment-inventory-anchor-set.v1", ObjectName: "EnvironmentInventoryAnchorSetV1"},
	{Schema: "environment-inventory-anchor.v1", ObjectName: "EnvironmentInventoryAnchorV1"},
	{Schema: "environment-inventory-membership-retirement.v1", ObjectName: "EnvironmentInventoryMembershipRetirementV1"},
	{Schema: "environment-inventory-record.v1", ObjectName: "EnvironmentInventoryRecordV1"},
	{Schema: "environment-inventory.v1", ObjectName: "EnvironmentInventoryV1"},
	{Schema: "fence-database-commit-attestation-archive-request.v1", ObjectName: "FenceDatabaseCommitAttestationArchiveRequestV1"},
	{Schema: "fence-database-commit-attestation-archive-response.v1", ObjectName: "FenceDatabaseCommitAttestationArchiveResponseV1"},
	{Schema: "fresh-import-topology-projection.v1", ObjectName: "FreshImportTopologyProjectionV1"},
	{Schema: "fresh-restore-import-application.v1", ObjectName: "FreshRestoreImportApplicationV1"},
	{Schema: "fresh-restore-import-manifest.v1", ObjectName: "FreshRestoreImportManifestV1"},
	{Schema: "fresh-restore-requirement.v1", ObjectName: "FreshRestoreRequirementV1"},
	{Schema: "fresh-target-inventory.v1", ObjectName: "FreshTargetInventoryV1"},
	{Schema: "fresh-v7-staging-evidence.v1", ObjectName: "FreshV7StagingEvidenceV1"},
	{Schema: "fresh-v7-staging-exclusion-abort.v1", ObjectName: "FreshV7StagingExclusionAbortV1"},
	{Schema: "fresh-v7-staging-exclusion-lease.v1", ObjectName: "FreshV7StagingExclusionLeaseV1"},
	{Schema: "fresh-v7-staging-exclusion-release.v1", ObjectName: "FreshV7StagingExclusionReleaseV1"},
	{Schema: "indeterminate-source-seal.v1", ObjectName: "IndeterminateSourceSealV1"},
	{Schema: "inspect-authority-epoch-transition-request.v1", ObjectName: "InspectAuthorityEpochTransitionRequestV1"},
	{Schema: "inspect-authority-epoch-transition-response.v1", ObjectName: "InspectAuthorityEpochTransitionResponseV1"},
	{Schema: "inspect-claim-v1-genesis-request.v1", ObjectName: "InspectClaimV1GenesisRequestV1"},
	{Schema: "inspect-claim-v1-genesis-response.v1", ObjectName: "InspectClaimV1GenesisResponseV1"},
	{Schema: "inspect-database-commit-attestation-archive-request.v1", ObjectName: "InspectDatabaseCommitAttestationArchiveRequestV1"},
	{Schema: "inspect-database-commit-attestation-archive-response.v1", ObjectName: "InspectDatabaseCommitAttestationArchiveResponseV1"},
	{Schema: "legacy-database-source-retirement-commit-attestation.v1", ObjectName: "LegacyDatabaseSourceRetirementCommitAttestationV1"},
	{Schema: "legacy-database-source-retirement.v1", ObjectName: "LegacyDatabaseSourceRetirementV1"},
	{Schema: "legacy-deployment-epoch-manifest.v1", ObjectName: "LegacyDeploymentEpochManifestV1"},
	{Schema: "legacy-environment-retirement.v1", ObjectName: "LegacyEnvironmentRetirementV1"},
	{Schema: "legacy-epoch-maximum-evidence.v1", ObjectName: "LegacyEpochMaximumEvidenceV1"},
	{Schema: "legacy-epoch-source-observation.v1", ObjectName: "LegacyEpochSourceObservationV1"},
	{Schema: "legacy-epoch-source-set.v1", ObjectName: "LegacyEpochSourceSetV1"},
	{Schema: "legacy-provider-epoch-transcript.v1", ObjectName: "LegacyProviderEpochTranscriptV1"},
	{Schema: "legacy-provider-namespace-retirement.v1", ObjectName: "LegacyProviderNamespaceRetirementV1"},
	{Schema: "legacy-runtime-shutdown-set.v1", ObjectName: "LegacyRuntimeShutdownSetV1"},
	{Schema: "legacy-runtime-shutdown.v1", ObjectName: "LegacyRuntimeShutdownV1"},
	{Schema: "legacy-source-archive-evidence.v1", ObjectName: "LegacySourceArchiveEvidenceV1"},
	{Schema: "legacy-source-inventory.v1", ObjectName: "LegacySourceInventoryV1"},
	{Schema: "legacy-source-membership-retirement.v1", ObjectName: "LegacySourceMembershipRetirementV1"},
	{Schema: "legacy-source-post-seal-exact-cover.v1", ObjectName: "LegacySourcePostSealExactCoverV1"},
	{Schema: "legacy-source-post-seal-observation.v1", ObjectName: "LegacySourcePostSealObservationV1"},
	{Schema: "legacy-source-retirement-authorization.v1", ObjectName: "LegacySourceRetirementAuthorizationV1"},
	{Schema: "legacy-source-retirement-set.v1", ObjectName: "LegacySourceRetirementSetV1"},
	{Schema: "legacy-source-seal-commit-attestation.v1", ObjectName: "LegacySourceSealCommitAttestationV1"},
	{Schema: "legacy-source-seal.v1", ObjectName: "LegacySourceSealV1"},
	{Schema: "local-runtime-isolation.v1", ObjectName: "LocalRuntimeIsolationV1"},
	{Schema: "open-claim-v1-genesis-request.v1", ObjectName: "OpenClaimV1GenesisRequestV1"},
	{Schema: "prepare-claim-v1-genesis-release-request.v1", ObjectName: "PrepareClaimV1GenesisReleaseRequestV1"},
	{Schema: "prepare-claim-v1-genesis-request.v1", ObjectName: "PrepareClaimV1GenesisRequestV1"},
	{Schema: "pristine-downgrade-authorized-state.v1", ObjectName: "PristineDowngradeAuthorizedStateV1"},
	{Schema: "pristine-downgrade-consume-receipt.v1", ObjectName: "PristineDowngradeConsumeReceiptV1"},
	{Schema: "pristine-downgrade-inventory.v1", ObjectName: "PristineDowngradeInventoryV1"},
	{Schema: "provider-namespace-absence.v1", ObjectName: "ProviderNamespaceAbsenceV1"},
	{Schema: "provider-protocol-downgrade-retirement-authorization.v1", ObjectName: "ProviderProtocolDowngradeRetirementAuthorizationV1"},
	{Schema: "provider-protocol-downgrade-retirement-set.v1", ObjectName: "ProviderProtocolDowngradeRetirementSetV1"},
	{Schema: "provider-protocol-downgrade-retirement.v1", ObjectName: "ProviderProtocolDowngradeRetirementV1"},
	{Schema: "provider-protocol-history-zero-projection.v1", ObjectName: "ProviderProtocolHistoryZeroProjectionV1"},
	{Schema: "publish-claim-v1-credential-policy-request.v1", ObjectName: "PublishClaimV1CredentialPolicyRequestV1"},
	{Schema: "rebind-claim-v1-runtime-instance-request.v1", ObjectName: "RebindClaimV1RuntimeInstanceRequestV1"},
	{Schema: "recover-authority-epoch-transition-request.v1", ObjectName: "RecoverAuthorityEpochTransitionRequestV1"},
	{Schema: "recover-staging-import-capability-request.v1", ObjectName: "RecoverStagingImportCapabilityRequestV1"},
	{Schema: "register-database-incarnation-request.v1", ObjectName: "RegisterDatabaseIncarnationRequestV1"},
	{Schema: "release-fresh-v7-staging-exclusion-request.v1", ObjectName: "ReleaseFreshV7StagingExclusionRequestV1"},
	{Schema: "renew-claim-v1-serving-lease-request.v1", ObjectName: "RenewClaimV1ServingLeaseRequestV1"},
	{Schema: "renew-database-runtime-instance-lease-request.v1", ObjectName: "RenewDatabaseRuntimeInstanceLeaseRequestV1"},
	{Schema: "request-authority-epoch-transition.v1", ObjectName: "RequestAuthorityEpochTransitionV1"},
	{Schema: "resolve-authority-epoch-transition-request.v1", ObjectName: "ResolveAuthorityEpochTransitionRequestV1"},
	{Schema: "retire-environment-inventory-membership-request.v1", ObjectName: "RetireEnvironmentInventoryMembershipRequestV1"},
	{Schema: "retire-legacy-environment-request.v1", ObjectName: "RetireLegacyEnvironmentRequestV1"},
	{Schema: "retire-legacy-provider-namespace-request.v1", ObjectName: "RetireLegacyProviderNamespaceRequestV1"},
	{Schema: "retire-legacy-source-membership-request.v1", ObjectName: "RetireLegacySourceMembershipRequestV1"},
	{Schema: "retire-provider-protocol-for-downgrade-request.v1", ObjectName: "RetireProviderProtocolForDowngradeRequestV1"},
	{Schema: "revoke-claim-v1-serving-lease-request.v1", ObjectName: "RevokeClaimV1ServingLeaseRequestV1"},
	{Schema: "staging-database-commit-attestation.v1", ObjectName: "StagingDatabaseCommitAttestationV1"},
	{Schema: "staging-database-commit-challenge.v1", ObjectName: "StagingDatabaseCommitChallengeV1"},
	{Schema: "staging-import-capability-recovery-application.v1", ObjectName: "StagingImportCapabilityRecoveryApplicationV1"},
	{Schema: "staging-import-capability-recovery-intent.v1", ObjectName: "StagingImportCapabilityRecoveryIntentV1"},
	{Schema: "staging-import-capability-recovery.v1", ObjectName: "StagingImportCapabilityRecoveryV1"},
	{Schema: "staging-import-capability-revocation-application.v1", ObjectName: "StagingImportCapabilityRevocationApplicationV1"},
	{Schema: "staging-import-capability-revocation.v1", ObjectName: "StagingImportCapabilityRevocationV1"},
	{Schema: "staging-import-capability.v1", ObjectName: "StagingImportCapabilityV1"},
}

// LookupAuthorityV7SchemaName returns an exact case-sensitive name pair.
// Success conveys no authentication, authorization, or evidence eligibility.
func LookupAuthorityV7SchemaName(schema string) (AuthorityV7SchemaName, error) {
	for _, pair := range authorityV7SchemaNameTable {
		if pair.Schema == schema {
			return pair, nil
		}
	}
	return AuthorityV7SchemaName{}, ErrUnknownAuthorityV7SchemaName
}

// AuthorityV7SchemaNames returns independent value labels in bytewise schema
// order. Membership conveys no authentication, authorization, or evidence
// eligibility; modifying the returned slice cannot modify the directory.
func AuthorityV7SchemaNames() []AuthorityV7SchemaName {
	names := make([]AuthorityV7SchemaName, len(authorityV7SchemaNameTable))
	copy(names, authorityV7SchemaNameTable[:])
	return names
}
