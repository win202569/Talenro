# Task11 final local metadata reconciliation — 2026-10-17

Status: LOCAL DESIGN RECONCILIATION COMPLETE,146 classified/0 unresolved cells.
The user confirmed the entire written16-item appendix after0ec812cd.
This is an executor's source-backed reconciliation,not independent certification,
runtime implementation,canonical spec-set registration or production acceptance.

## Sources and precedence

The complete original identity set is [v7 §4.1](../superpowers/specs/2026-08-24-nodecontrol-authority-v7-upgrade-design.md).
Its146 exact object/schema pairs match the existing name-only production table
and both tracked name fixtures;no canonical identity is added,removed or renamed.
The original audit and all earlier counts remain historical evidence.

| Code | Classification source |
| --- | --- |
| B93 | [Original93 source-backed classifications](2026-10-16-task11-registry-metadata-audit.md),including approved[C1 request exception](../superpowers/specs/2026-10-16-authority-v7-request-auth-clarification.md) for#7. |
| N9 | [Approved nine-item clarification](../superpowers/specs/2026-10-17-authority-v7-metadata-nine-item-clarification-design.md). |
| S93 | Source-backed#93,LegacySourceInventoryV1:original L795/797/803/809/811;see[current reconciliation](2026-10-17-task11-metadata-current-reconciliation.md). Not a new independent review. |
| R27:profile | [Approved27-request matrix](../superpowers/specs/2026-10-17-authority-v7-request-auth-matrix-design.md),including11 explicit new roles and existing-channel obligations. |
| D16 | [Approved16-item data-path appendix](../superpowers/specs/2026-10-17-authority-v7-data-path-clarification-design.md):5 DB preimages,3 provider envelopes,8 local-only forms. New storage obligations are not existing SQL facts. |

R27 profile codes are U=upgrade,E=epoch,L=serving lease,B=rebind,O=ordinary,
R=read,C=commit challenge,F=archive fence,I=archive inspect,S=staging,D=provider
downgrade retirement. Each request retains independent receiver/scope/key/policy
and trusted-channel requirements;all27 have EvidenceKinds=[].
The [approved admission method](../superpowers/specs/2026-10-16-authority-v7-metadata-admission-clarification.md)
keeps request authentication,common-envelope role and bundle eligibility separate.

## Reconciliation checks and scope

Exact identity comparisons:original table146 = original audit146 = production
schema_names.go146 = schema-name-registry.v1.json146. The independent name golden
also contains146 exact identities. No duplicate schema/object,no missing/extra
identity,no unresolved role/form cell and no conflicting role/form combination
were found in this local consolidation. The count alone is not the check.

Final forms:71 external_signed_envelope with non-null role;27
database_immutable_body with null role;48 explicit empty sets,of which32 have
non-null role and16 null. There are37 distinct non-null roles.
Origins:93 B93 +9 N9 +1 S93 +27 R27 +16 D16 =146.
Null is explicit absence of common-envelope role,not an anonymous trust grant.
Empty evidence sets are explicit non-eligibility,not missing metadata.

Exclude AuthorityEffectCommitmentV1,AuthorityEffectResolutionV1 and
ActivationDecisionEvidenceV1 from this new directory;preserve existing historical
protocol support. No wire field,body digest,signature domain,key/trust policy,
SQL/manifest or approved canonical-set digest changed.

The following table is a portable literal planning oracle in original table order.
EvidenceKinds bracket notation is exact;runtime lists will sort by schema.
Original L references identify source context,not claims that later amendments
were already present there. B93#7 additionally requires C1;all deltas require their
approved written source. Implementers must review those sources,not infer policy
from neighboring rows or schema prefixes.

## Complete146-row checklist

| ID | ObjectName | Schema | SignerRoleOrNull | EvidenceKinds | Source | Original L |
| --- | --- | --- | --- | --- | --- | --- |
| 1 | CanonicalEvidenceBundleV1 | canonical-evidence-bundle.v1 | null | [] | B93 | 333,367,375,381,732,734,803,871,999,1003 |
| 2 | EnvironmentInventoryV1 | environment-inventory.v1 | release_deployment_operator | [external_signed_envelope] | B93 | 282,284,309,319 |
| 3 | EnvironmentInventoryRecordV1 | environment-inventory-record.v1 | null | [] | B93 | 291,295 |
| 4 | DeploymentEnvironmentAttestationV1 | deployment-environment-attestation.v1 | deployment_environment_attestor | [external_signed_envelope] | B93 | 293,295 |
| 5 | EnvironmentInventoryAnchorV1 | environment-inventory-anchor.v1 | claim_v1_inventory_anchor | [external_signed_envelope] | B93 | 305 |
| 6 | EnvironmentInventoryAnchorSetV1 | environment-inventory-anchor-set.v1 | null | [] | B93 | 307 |
| 7 | PublishClaimV1CredentialPolicyRequestV1 | publish-claim-v1-credential-policy-request.v1 | claim_v1_security_admin_request | [] | B93 | 325 |
| 8 | ClaimV1CredentialPolicyV1 | claim-v1-credential-policy.v1 | claim_v1_security_policy | [external_signed_envelope] | B93 | 321,325,734 |
| 9 | LocalRuntimeIsolationV1 | local-runtime-isolation.v1 | deployment_isolation_operator | [external_signed_envelope] | B93 | 315,317,331,351,782 |
| 10 | LegacyRuntimeShutdownV1 | legacy-runtime-shutdown.v1 | release_deployment_operator | [external_signed_envelope] | B93 | 322,331 |
| 11 | LegacyRuntimeShutdownSetV1 | legacy-runtime-shutdown-set.v1 | null | [] | B93 | 333 |
| 12 | AcquireDatabaseIncarnationRegistrationChallengeRequestV1 | acquire-database-incarnation-registration-challenge-request.v1 | claim_v1_upgrade_request | [] | R27:U | 355 |
| 13 | DatabaseIncarnationRegistrationChallengeV1 | database-incarnation-registration-challenge.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 355 |
| 14 | DatabaseIncarnationAttestationV1 | database-incarnation-attestation.v1 | database_incarnation_attestor | [external_signed_envelope] | B93 | 353,357 |
| 15 | DatabaseRuntimeInstanceBindingV1 | database-runtime-instance-binding.v1 | database_runtime_attestor | [external_signed_envelope] | B93 | 359 |
| 16 | RenewDatabaseRuntimeInstanceLeaseRequestV1 | renew-database-runtime-instance-lease-request.v1 | null | [] | B93 | 361 |
| 17 | DatabaseRuntimeInstanceLeaseV1 | database-runtime-instance-lease.v1 | database_runtime_attestor | [external_signed_envelope] | B93 | 363 |
| 18 | DatabaseRuntimeInstanceLeaseTerminationV1 | database-runtime-instance-lease-termination.v1 | database_runtime_attestor | [external_signed_envelope] | B93 | 365 |
| 19 | DatabaseTimelineLineageAttestationV1 | database-timeline-lineage-attestation.v1 | database_timeline_lineage_attestor | [external_signed_envelope] | B93 | 297,393,537 |
| 20 | RegisterDatabaseIncarnationRequestV1 | register-database-incarnation-request.v1 | claim_v1_upgrade_request | [] | R27:U | 367 |
| 21 | DatabaseIncarnationRegistrationV1 | database-incarnation-registration.v1 | claim_v1_incarnation_registry | [external_signed_envelope] | B93 | 367 |
| 22 | DatabaseIncarnationRebindingRegistrationV1 | database-incarnation-rebinding-registration.v1 | claim_v1_incarnation_registry | [external_signed_envelope] | B93 | 395,752,754,1111 |
| 23 | AuthorityRuntimeRegistrationResultV1 | authority-runtime-registration-result.v1 | null | [database_immutable_body] | B93 | 369,596 |
| 24 | AuthorityRuntimeRebindResultV1 | authority-runtime-rebind-result.v1 | null | [database_immutable_body] | B93 | 301,393,444,454,460,596,736,740,754 |
| 25 | DatabaseAuthorityHeadV1 | database-authority-head.v1 | null | [] | B93 | 303,371,375 |
| 26 | DatabaseAuthorityHeadAttestationV1 | database-authority-head-attestation.v1 | database_incarnation | [external_signed_envelope] | B93 | 301,375,454,458,474,734,1116 |
| 27 | DatabaseAuthorityRebindGapAttestationV1 | database-authority-rebind-gap-attestation.v1 | database_runtime_attestor | [external_signed_envelope] | B93 | 377,450,454,480,537,608,711,1015 |
| 28 | DatabaseAuthorityPostRecoveryRebindAttestationV1 | database-authority-post-recovery-rebind-attestation.v1 | database_runtime_attestor | [external_signed_envelope] | B93 | 301,399,537,711,1015 |
| 29 | DatabaseLegacyAbsenceProjectionV1 | database-legacy-absence-projection.v1 | null | [database_immutable_body] | D16 | 409,599 |
| 30 | DatabaseLegacyAbsenceObservationV1 | database-legacy-absence-observation.v1 | null | [database_immutable_body] | D16 | 413 |
| 31 | ProviderNamespaceAbsenceV1 | provider-namespace-absence.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 415,498,1079 |
| 32 | AuthorityGenesisEpochEvidenceV1 | authority-genesis-epoch-evidence.v1 | null | [database_immutable_body] | D16 | 419,421,728,825,918 |
| 33 | AuthorityEpochTransitionIntentV1 | authority-epoch-transition-intent.v1 | null | [database_immutable_body] | B93 | 432 |
| 34 | RequestAuthorityEpochTransitionV1 | request-authority-epoch-transition.v1 | claim_v1_epoch_request | [] | R27:E | 434 |
| 35 | AuthorityEpochTransitionV1 | authority-epoch-transition.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 434,448 |
| 36 | AuthorityEpochTransitionApplicationV1 | authority-epoch-transition-application.v1 | null | [database_immutable_body] | B93 | 436 |
| 37 | AuthorityEpochTransitionResolutionV1 | authority-epoch-transition-resolution.v1 | null | [database_immutable_body] | B93 | 436 |
| 38 | ResolveAuthorityEpochTransitionRequestV1 | resolve-authority-epoch-transition-request.v1 | claim_v1_epoch_request | [] | R27:E | 438 |
| 39 | AuthorityEpochTransitionResolutionReceiptV1 | authority-epoch-transition-resolution-receipt.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 438,448 |
| 40 | AuthorityEpochTransitionCancellationV1 | authority-epoch-transition-cancellation.v1 | null | [database_immutable_body] | N9 | 444,460,740,756 |
| 41 | AuthorityEpochTransitionTerminalApplicationV1 | authority-epoch-transition-terminal-application.v1 | null | [database_immutable_body] | B93 | 371,432,442,740,756 |
| 42 | InspectAuthorityEpochTransitionRequestV1 | inspect-authority-epoch-transition-request.v1 | claim_v1_read_request | [] | R27:R | 448 |
| 43 | InspectAuthorityEpochTransitionResponseV1 | inspect-authority-epoch-transition-response.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 383,448 |
| 44 | AuthorityEpochTransitionRecoveryIntentV1 | authority-epoch-transition-recovery-intent.v1 | null | [database_immutable_body] | N9 | 452 |
| 45 | RecoverAuthorityEpochTransitionRequestV1 | recover-authority-epoch-transition-request.v1 | claim_v1_epoch_request | [] | R27:E | 379,454,610,683,736 |
| 46 | AuthorityEpochTransitionRecoveryV1 | authority-epoch-transition-recovery.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 379,456 |
| 47 | AuthorityEpochTransitionRecoveryPrefixDecisionV1 | authority-epoch-transition-recovery-prefix-decision.v1 | null | [database_immutable_body] | N9 | 476,608,637 |
| 48 | AuthorityEpochTransitionRecoveryApplicationV1 | authority-epoch-transition-recovery-application.v1 | null | [database_immutable_body] | B93 | 371,464,637 |
| 49 | EmptyDatabaseInventoryV1 | empty-database-inventory.v1 | null | [database_immutable_body] | D16 | 486,490 |
| 50 | AuthorityProtocolUpgradeIntentV1 | authority-protocol-upgrade-intent.v1 | null | [database_immutable_body] | B93 | 488 |
| 51 | AuthorityProtocolUpgradeAttemptV1 | authority-protocol-upgrade-attempt.v1 | null | [database_immutable_body] | B93 | 492,597 |
| 52 | PrepareClaimV1GenesisRequestV1 | prepare-claim-v1-genesis-request.v1 | claim_v1_upgrade_request | [] | R27:U | 496 |
| 53 | ClaimV1GenesisPreparationV1 | claim-v1-genesis-preparation.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 494,496 |
| 54 | AuthorityProtocolActivationV1 | authority-protocol-activation.v1 | null | [database_immutable_body] | B93 | 553,599 |
| 55 | AuthorityProtocolActivationCompletionV1 | authority-protocol-activation-completion.v1 | null | [database_immutable_body] | B93 | 555,600 |
| 56 | AuthorityProtocolActivationReleaseV1 | authority-protocol-activation-release.v1 | null | [database_immutable_body] | B93 | 557,601 |
| 57 | AcquireDatabaseCommitChallengeRequestV1 | acquire-database-commit-challenge-request.v1 | claim_v1_commit_challenge_request | [] | R27:C | 604 |
| 58 | DatabaseCommitChallengeV1 | database-commit-challenge.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 604,839 |
| 59 | DatabaseCommitAttestationV1 | database-commit-attestation.v1 | database_incarnation | [external_signed_envelope] | B93 | 301,438,454,458,466,599,601,629,839,855,875 |
| 60 | FenceDatabaseCommitAttestationArchiveSemanticIDJournalRecordV1 | database-commit-archive-fence-semantic-id-journal-record.v1 | null | [] | N9 | 641 |
| 61 | FenceDatabaseCommitAttestationArchiveRequestV1 | fence-database-commit-attestation-archive-request.v1 | claim_v1_archive_fence_request | [] | R27:F | 637 |
| 62 | FenceDatabaseCommitAttestationArchiveResponseV1 | fence-database-commit-attestation-archive-response.v1 | database_incarnation | [external_signed_envelope] | B93 | 606,657 |
| 63 | InspectDatabaseCommitAttestationArchiveRequestV1 | inspect-database-commit-attestation-archive-request.v1 | claim_v1_archive_inspect_request | [] | R27:I | 659 |
| 64 | InspectDatabaseCommitAttestationArchiveResponseV1 | inspect-database-commit-attestation-archive-response.v1 | database_incarnation | [external_signed_envelope] | B93 | 606,661 |
| 65 | ClaimV1GenesisCompletionV1 | claim-v1-genesis-completion.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 675 |
| 66 | CompleteClaimV1GenesisRequestV1 | complete-claim-v1-genesis-request.v1 | claim_v1_upgrade_request | [] | R27:U | 675 |
| 67 | ClaimV1GenesisReleasePreparationV1 | claim-v1-genesis-release-preparation.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 677 |
| 68 | PrepareClaimV1GenesisReleaseRequestV1 | prepare-claim-v1-genesis-release-request.v1 | claim_v1_upgrade_request | [] | R27:U | 677 |
| 69 | OpenClaimV1GenesisRequestV1 | open-claim-v1-genesis-request.v1 | claim_v1_upgrade_request | [] | R27:U | 679 |
| 70 | ClaimV1GenesisOpenV1 | claim-v1-genesis-open.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 679,689,769,967 |
| 71 | InspectClaimV1GenesisRequestV1 | inspect-claim-v1-genesis-request.v1 | claim_v1_read_request | [] | R27:R | 681 |
| 72 | InspectClaimV1GenesisResponseV1 | inspect-claim-v1-genesis-response.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 681 |
| 73 | ClaimV1ProviderHeadV1 | claim-v1-provider-head.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 639,687,689,922 |
| 74 | AcquireDatabaseIncarnationProofRequestV1 | acquire-database-incarnation-proof-request.v1 | claim_v1_read_request | [] | R27:R | 709 |
| 75 | DatabaseIncarnationProofV1 | database-incarnation-proof.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 397,454,709,724,915,937 |
| 76 | DatabaseRuntimeInstanceLeaseDescendantChainV1 | database-runtime-instance-lease-descendant-chain.v1 | database_runtime_attestor | [external_signed_envelope] | B93 | 707 |
| 77 | AuthorityProtocolContextV1 | authority-protocol-context.v1 | null | [] | B93 | 713,719 |
| 78 | RenewClaimV1ServingLeaseRequestV1 | renew-claim-v1-serving-lease-request.v1 | claim_v1_serving_lease_request | [] | R27:L | 730 |
| 79 | ClaimV1ServingLeaseV1 | claim-v1-serving-lease.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 728,734,742 |
| 80 | RevokeClaimV1ServingLeaseRequestV1 | revoke-claim-v1-serving-lease-request.v1 | claim_v1_serving_lease_request | [] | R27:L | 744 |
| 81 | ClaimV1ServingLeaseRevocationV1 | claim-v1-serving-lease-revocation.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 744 |
| 82 | DatabaseRuntimeRebindAuthorizationV1 | database-runtime-rebind-authorization.v1 | runtime_rebind_operator | [external_signed_envelope] | B93 | 381 |
| 83 | RebindClaimV1RuntimeInstanceRequestV1 | rebind-claim-v1-runtime-instance-request.v1 | claim_v1_rebind_request | [] | R27:B | 746,752 |
| 84 | ClaimV1RuntimeInstanceRebindV1 | claim-v1-runtime-instance-rebind.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 395,448,691,746 |
| 85 | ClaimV1ReserveRequestV1 | claim-v1-reserve-request.v1 | claim_v1_ordinary_request | [] | R27:O | 715 |
| 86 | ClaimV1ReservationV1 | claim-v1-reservation.v1 | claim_v1_provider | [external_signed_envelope] | N9 | 716,719,732 |
| 87 | ClaimV1FinalizeRequestV1 | claim-v1-finalize-request.v1 | claim_v1_ordinary_request | [] | R27:O | 717 |
| 88 | ClaimV1AbortRequestV1 | claim-v1-abort-request.v1 | claim_v1_ordinary_request | [] | R27:O | 718 |
| 89 | ClaimV1ReceiptV1 | claim-v1-receipt.v1 | claim_v1_provider | [external_signed_envelope] | D16 | 719,720 |
| 90 | ClaimV1RecordV1 | claim-v1-record.v1 | claim_v1_provider | [external_signed_envelope] | D16 | 720,732 |
| 91 | IndeterminateSourceSealV1 | indeterminate-source-seal.v1 | null | [] | D16 | 586,782,784,1139 |
| 92 | LegacySourceSealV1 | legacy-source-seal.v1 | null | [database_immutable_body] | B93 | 586,786 |
| 93 | LegacySourceInventoryV1 | legacy-source-inventory.v1 | null | [database_immutable_body] | S93 | 795 |
| 94 | LegacyEpochSourceObservationV1 | legacy-epoch-source-observation.v1 | legacy_epoch_auditor | [external_signed_envelope] | B93 | 803,805 |
| 95 | LegacyProviderEpochTranscriptV1 | legacy-provider-epoch-transcript.v1 | legacy_epoch_provider | [external_signed_envelope] | B93 | 799 |
| 96 | LegacyDeploymentEpochManifestV1 | legacy-deployment-epoch-manifest.v1 | release_deployment_operator | [external_signed_envelope] | B93 | 801 |
| 97 | LegacyEpochSourceSetV1 | legacy-epoch-source-set.v1 | legacy_epoch_auditor | [external_signed_envelope] | B93 | 428,805 |
| 98 | LegacyEpochMaximumEvidenceV1 | legacy-epoch-maximum-evidence.v1 | legacy_epoch_auditor | [external_signed_envelope] | B93 | 807 |
| 99 | LegacySourceRetirementAuthorizationV1 | legacy-source-retirement-authorization.v1 | legacy_source_retirement_admin | [external_signed_envelope] | B93 | 335 |
| 100 | RetireLegacySourceMembershipRequestV1 | retire-legacy-source-membership-request.v1 | legacy_source_retirement_admin | [] | N9 | 337 |
| 101 | LegacySourceMembershipRetirementV1 | legacy-source-membership-retirement.v1 | release_deployment_operator | [external_signed_envelope] | B93 | 337,428,786 |
| 102 | LegacyDatabaseSourceRetirementV1 | legacy-database-source-retirement.v1 | null | [database_immutable_body] | B93 | 341,786 |
| 103 | LegacyDatabaseSourceRetirementCommitAttestationV1 | legacy-database-source-retirement-commit-attestation.v1 | legacy_source_archive_attestor | [external_signed_envelope] | B93 | 341 |
| 104 | RetireLegacyEnvironmentRequestV1 | retire-legacy-environment-request.v1 | legacy_source_retirement_admin | [] | N9 | 343 |
| 105 | LegacyEnvironmentRetirementV1 | legacy-environment-retirement.v1 | deployment_environment_retirement_authority | [external_signed_envelope] | B93 | 343 |
| 106 | RetireLegacyProviderNamespaceRequestV1 | retire-legacy-provider-namespace-request.v1 | legacy_source_retirement_admin | [] | N9 | 345 |
| 107 | LegacyProviderNamespaceRetirementV1 | legacy-provider-namespace-retirement.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 345 |
| 108 | LegacySourceRetirementSetV1 | legacy-source-retirement-set.v1 | null | [database_immutable_body] | D16 | 347,428,786 |
| 109 | LegacySourcePostSealObservationV1 | legacy-source-post-seal-observation.v1 | legacy_source_archive_attestor | [external_signed_envelope] | B93 | 811 |
| 110 | LegacySourcePostSealExactCoverV1 | legacy-source-post-seal-exact-cover.v1 | release_deployment_operator | [external_signed_envelope] | B93 | 428,813 |
| 111 | LegacySourceSealCommitAttestationV1 | legacy-source-seal-commit-attestation.v1 | legacy_source_archive_attestor | [external_signed_envelope] | B93 | 809 |
| 112 | LegacySourceArchiveEvidenceV1 | legacy-source-archive-evidence.v1 | release_deployment_operator | [external_signed_envelope] | B93 | 815 |
| 113 | FreshRestoreRequirementV1 | fresh-restore-requirement.v1 | null | [] | D16 | 586,780,817,1139 |
| 114 | FreshRestoreImportManifestV1 | fresh-restore-import-manifest.v1 | fresh_restore_export_operator | [external_signed_envelope] | B93 | 833,1026,1138 |
| 115 | FreshImportTopologyProjectionV1 | fresh-import-topology-projection.v1 | null | [] | D16 | 831,837,1150 |
| 116 | StagingImportCapabilityV1 | staging-import-capability.v1 | fresh_restore_staging_import_authorizer | [external_signed_envelope] | B93 | 588,839 |
| 117 | StagingImportCapabilityRevocationV1 | staging-import-capability-revocation.v1 | fresh_restore_staging_import_revoker | [external_signed_envelope] | B93 | 869 |
| 118 | StagingImportCapabilityRevocationApplicationV1 | staging-import-capability-revocation-application.v1 | null | [database_immutable_body] | B93 | 637,685,859,869,871 |
| 119 | StagingImportCapabilityRecoveryIntentV1 | staging-import-capability-recovery-intent.v1 | null | [database_immutable_body] | B93 | 847 |
| 120 | RecoverStagingImportCapabilityRequestV1 | recover-staging-import-capability-request.v1 | claim_v1_staging_request | [] | R27:S | 851 |
| 121 | StagingImportCapabilityRecoveryV1 | staging-import-capability-recovery.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 853 |
| 122 | StagingImportCapabilityRecoveryApplicationV1 | staging-import-capability-recovery-application.v1 | null | [database_immutable_body] | B93 | 637,855,859 |
| 123 | AcquireFreshV7StagingExclusionRequestV1 | acquire-fresh-v7-staging-exclusion-request.v1 | claim_v1_staging_request | [] | R27:S | 841 |
| 124 | FreshV7StagingExclusionLeaseV1 | fresh-v7-staging-exclusion-lease.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 841 |
| 125 | FreshRestoreImportApplicationV1 | fresh-restore-import-application.v1 | null | [database_immutable_body] | B93 | 637,653,685,845,871,1105,1145 |
| 126 | AcquireStagingDatabaseCommitChallengeRequestV1 | acquire-staging-database-commit-challenge-request.v1 | claim_v1_commit_challenge_request | [] | R27:C | 871,873 |
| 127 | StagingDatabaseCommitChallengeV1 | staging-database-commit-challenge.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 875 |
| 128 | StagingDatabaseCommitAttestationV1 | staging-database-commit-attestation.v1 | database_incarnation | [external_signed_envelope] | B93 | 301,871,875 |
| 129 | ReleaseFreshV7StagingExclusionRequestV1 | release-fresh-v7-staging-exclusion-request.v1 | claim_v1_staging_request | [] | R27:S | 877 |
| 130 | FreshV7StagingExclusionReleaseV1 | fresh-v7-staging-exclusion-release.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 877 |
| 131 | AbortFreshV7StagingExclusionRequestV1 | abort-fresh-v7-staging-exclusion-request.v1 | claim_v1_staging_request | [] | R27:S | 879 |
| 132 | FreshV7StagingExclusionAbortV1 | fresh-v7-staging-exclusion-abort.v1 | claim_v1_provider | [external_signed_envelope] | D16 | 879 |
| 133 | FreshTargetInventoryV1 | fresh-target-inventory.v1 | fresh_target_inventory_operator | [external_signed_envelope] | B93 | 881,883 |
| 134 | FreshV7StagingEvidenceV1 | fresh-v7-staging-evidence.v1 | fresh_restore_evidence_operator | [external_signed_envelope] | B93 | 889 |
| 135 | RetireEnvironmentInventoryMembershipRequestV1 | retire-environment-inventory-membership-request.v1 | provider_downgrade_retirement_admin | [] | N9 | 991 |
| 136 | EnvironmentInventoryMembershipRetirementV1 | environment-inventory-membership-retirement.v1 | release_deployment_operator | [external_signed_envelope] | B93 | 991 |
| 137 | ProviderProtocolDowngradeRetirementAuthorizationV1 | provider-protocol-downgrade-retirement-authorization.v1 | provider_downgrade_retirement_admin | [external_signed_envelope] | B93 | 993 |
| 138 | ProviderProtocolHistoryZeroProjectionV1 | provider-protocol-history-zero-projection.v1 | claim_v1_provider_history_auditor | [external_signed_envelope] | B93 | 999 |
| 139 | RetireProviderProtocolForDowngradeRequestV1 | retire-provider-protocol-for-downgrade-request.v1 | provider_downgrade_retirement_request | [] | R27:D | 995 |
| 140 | ProviderProtocolDowngradeRetirementV1 | provider-protocol-downgrade-retirement.v1 | claim_v1_provider | [external_signed_envelope] | B93 | 1001 |
| 141 | ProviderProtocolDowngradeRetirementSetV1 | provider-protocol-downgrade-retirement-set.v1 | null | [] | D16 | 1003 |
| 142 | PristineDowngradeInventoryV1 | pristine-downgrade-inventory.v1 | null | [] | D16 | 1005 |
| 143 | PristineDowngradeAuthorizedStateV1 | pristine-downgrade-authorized-state.v1 | null | [] | D16 | 1007 |
| 144 | PristineDowngradeConsumeReceiptV1 | pristine-downgrade-consume-receipt.v1 | null | [] | D16 | 1009 |
| 145 | AuthorityProtocolMigrationLatchV1 | authority-protocol-migration-latch.v1 | null | [] | D16 | 537 |
| 146 | AuthorityProtocolDowngradeAuthorizationV1 | authority-protocol-downgrade-authorization.v1 | authority_protocol_downgrade_authorizer | [external_signed_envelope] | B93 | 1007 |

## Handoff

The [metadata-only implementation plan](../superpowers/plans/2026-10-17-authority-v7-metadata-registry.md)
is proposed for review and execution-method selection,not executed.
It may expose these literal protocol forms only,not Verified/authorized values.
A-storage/consumer integration,real authentication/codecs,missing-row recovery,
canonical registration and physical DB/PITR acceptance require separate plans.

This turn performed document/source comparisons and independent standard SHA256
framing calculations only. No Go build/test,native workload,Docker,DB or live
signature validation ran. Existing consumed attempt identities remain consumed.
Necessary documents travel with Git;ignored raw evidence and tool caches do not.
