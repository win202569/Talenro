# Task11 registry reconciliation — corrected 146-row source index

Follow-up:[consolidated C2 admission proposal](../superpowers/specs/2026-10-16-authority-v7-metadata-admission-clarification.md)
is written and awaiting approval. No row is certified merely by that proposal;
the146-row metadata reconciliation and executable plan remain incomplete.

Baseline:247f1f41c8bc1e8a0d8600133dae556e1489b38e. User approved the
[C1 local clarification](../superpowers/specs/2026-10-16-authority-v7-request-auth-clarification.md)
and further reconciliation/planning. No implementation,test,Go/native,SDK audit,
Docker/DB or provider operation. This is a source/contract report,not an executable
implementation plan or product registry.

## Count correction — executor error,not protocol change

The earlier147 count incorrectly included the Markdown header object/schema.
The exact §4.1 table contains146 V1 object/schema pairs,146 unique object names
and146 unique schemas. Filtering object labels to their literal V1 suffix
excludes the header. The canonical specification/table is unchanged;local design,
clarification and recovery counts are corrected with this explicit notice.
This also invalidates the earlier claim that147 policy rows must be implemented.
No implementation/fixture was produced from the wrong count.

## C1 locally resolved;independent gates remain

PublishClaimV1CredentialPolicyRequestV1 now has local approved common-envelope
role claim_v1_security_admin_request and EvidenceKinds=[];response retains
claim_v1_security_policy/external_signed_envelope. SignerRoleOrNull denotes the
common-envelope role independently of bundle eligibility. Mandatory request
authentication,independently provisioned administrator trust and request/response
key separation remain. No canonical-set registration,deployment compatibility,
key configuration or implementation was performed. This decides only its row.

## C2 — grouped unresolved descriptor-policy boundaries

The approved registry design requires full per-row metadata before coding.
The following recurring questions must be addressed together instead of copying
a neighboring response role or guessing from storage/location/name.

| Class | Concrete source examples | What is not certified |
| --- | --- | --- |
| Request bodies without an explicit request-envelope role | AcquireDatabaseIncarnationRegistrationChallengeRequestV1 §4.4 line355;RegisterDatabaseIncarnationRequestV1 line367;AcquireDatabaseIncarnationProofRequestV1 §6.2 line709;ClaimV1Reserve/Finalize/AbortRequestV1 lines715/717/718 | Whether common-envelope role is null or another explicit request-auth contract;C1 role must not be generalized |
| Embedded bodies/aggregates | EnvironmentInventoryRecordV1 §4.2 line291;EnvironmentInventoryAnchorSetV1 line307;AuthorityProtocolContextV1 §6.2 line713 | Independent evidence eligibility does not follow from a digest,embedding or signed parent;no implicit database_immutable_body or inherited signer |
| Temporary/transaction-local projections | DatabaseAuthorityHeadV1 §4.4 lines371/375;PristineDowngradeInventoryV1 §9.3 line1005;PristineDowngradeAuthorizedStateV1 line1007 | Head explicitly must not be treated as DB-immutable evidence;that example alone does not classify other projections |

These are gaps in assigning the proposed static descriptor,not a claim that
the old protocol permits unauthenticated calls or a verified product defect.
One paragraph may mention both a request and signed response;its role word
cannot be attached to every object by text scraping. Existing protocol/state/
preimage constraints remain mandatory. The index below assigns no inferred roles.

## Recommended next decision

Prepare one consolidated metadata-admission clarification covering the three
classes:separate non-envelope request authentication from envelope role,define
standalone evidence eligibility for embedded/container/transient forms,and
specify how genuinely undecided cases remain unpublishable. Each rule needs
explicit scope and compatibility limits. No new allow/deny default is adopted
by this report. Then reconcile all146 rows and list remaining exceptions together.

Alternative:explicitly approve a narrower schema/object-only first slice,
postponing policy metadata. That changes the approved design and cannot be
implemented silently. C1 approval alone does not close the remaining map.
No implementation-ready plan is claimed while C2 remains unresolved.

## Complete source index — references,not certified metadata

Source:docs/superpowers/specs/2026-08-24-nodecontrol-authority-v7-upgrade-design.md
at the baseline above. Pair set comes only from §4.1 literal V1 rows. Reference
lines are case-sensitive whole-object-name occurrences from line282 onward,
including headings/cross-references,not necessarily defining paragraphs.
There are146 rows,no header pseudo-row. C1 is marked locally approved;other
rows are indexed,not automatically certified as unresolved protocol defects.

| # | Object | Schema | Source lines | Reconciliation |
| ---: | --- | --- | --- | --- |
| 1 | CanonicalEvidenceBundleV1 | canonical-evidence-bundle.v1 | 333,367,375,381,732,734,803,871,999,1003 | Indexed;full metadata not certified |
| 2 | EnvironmentInventoryV1 | environment-inventory.v1 | 282,284,309,319 | Indexed;full metadata not certified |
| 3 | EnvironmentInventoryRecordV1 | environment-inventory-record.v1 | 291,295 | Indexed;full metadata not certified |
| 4 | DeploymentEnvironmentAttestationV1 | deployment-environment-attestation.v1 | 293,295 | Indexed;full metadata not certified |
| 5 | EnvironmentInventoryAnchorV1 | environment-inventory-anchor.v1 | 305 | Indexed;full metadata not certified |
| 6 | EnvironmentInventoryAnchorSetV1 | environment-inventory-anchor-set.v1 | 307 | Indexed;full metadata not certified |
| 7 | PublishClaimV1CredentialPolicyRequestV1 | publish-claim-v1-credential-policy-request.v1 | 325 | C1 locally approved;canonical/deployment gates open |
| 8 | ClaimV1CredentialPolicyV1 | claim-v1-credential-policy.v1 | 321,325,734 | Indexed;full metadata not certified |
| 9 | LocalRuntimeIsolationV1 | local-runtime-isolation.v1 | 315,317,331,351,782 | Indexed;full metadata not certified |
| 10 | LegacyRuntimeShutdownV1 | legacy-runtime-shutdown.v1 | 322,331 | Indexed;full metadata not certified |
| 11 | LegacyRuntimeShutdownSetV1 | legacy-runtime-shutdown-set.v1 | 333 | Indexed;full metadata not certified |
| 12 | AcquireDatabaseIncarnationRegistrationChallengeRequestV1 | acquire-database-incarnation-registration-challenge-request.v1 | 355 | Indexed;full metadata not certified |
| 13 | DatabaseIncarnationRegistrationChallengeV1 | database-incarnation-registration-challenge.v1 | 355 | Indexed;full metadata not certified |
| 14 | DatabaseIncarnationAttestationV1 | database-incarnation-attestation.v1 | 353,357 | Indexed;full metadata not certified |
| 15 | DatabaseRuntimeInstanceBindingV1 | database-runtime-instance-binding.v1 | 359 | Indexed;full metadata not certified |
| 16 | RenewDatabaseRuntimeInstanceLeaseRequestV1 | renew-database-runtime-instance-lease-request.v1 | 361 | Indexed;full metadata not certified |
| 17 | DatabaseRuntimeInstanceLeaseV1 | database-runtime-instance-lease.v1 | 363 | Indexed;full metadata not certified |
| 18 | DatabaseRuntimeInstanceLeaseTerminationV1 | database-runtime-instance-lease-termination.v1 | 365 | Indexed;full metadata not certified |
| 19 | DatabaseTimelineLineageAttestationV1 | database-timeline-lineage-attestation.v1 | 297,393,537 | Indexed;full metadata not certified |
| 20 | RegisterDatabaseIncarnationRequestV1 | register-database-incarnation-request.v1 | 367 | Indexed;full metadata not certified |
| 21 | DatabaseIncarnationRegistrationV1 | database-incarnation-registration.v1 | 367 | Indexed;full metadata not certified |
| 22 | DatabaseIncarnationRebindingRegistrationV1 | database-incarnation-rebinding-registration.v1 | 395,752,754,1111 | Indexed;full metadata not certified |
| 23 | AuthorityRuntimeRegistrationResultV1 | authority-runtime-registration-result.v1 | 369,596 | Indexed;full metadata not certified |
| 24 | AuthorityRuntimeRebindResultV1 | authority-runtime-rebind-result.v1 | 301,393,444,454,460,596,736,740,754 | Indexed;full metadata not certified |
| 25 | DatabaseAuthorityHeadV1 | database-authority-head.v1 | 303,371,375 | Indexed;full metadata not certified |
| 26 | DatabaseAuthorityHeadAttestationV1 | database-authority-head-attestation.v1 | 301,375,454,458,474,734,1116 | Indexed;full metadata not certified |
| 27 | DatabaseAuthorityRebindGapAttestationV1 | database-authority-rebind-gap-attestation.v1 | 377,450,454,480,537,608,711,1015 | Indexed;full metadata not certified |
| 28 | DatabaseAuthorityPostRecoveryRebindAttestationV1 | database-authority-post-recovery-rebind-attestation.v1 | 301,399,537,711,1015 | Indexed;full metadata not certified |
| 29 | DatabaseLegacyAbsenceProjectionV1 | database-legacy-absence-projection.v1 | 409,599 | Indexed;full metadata not certified |
| 30 | DatabaseLegacyAbsenceObservationV1 | database-legacy-absence-observation.v1 | 413 | Indexed;full metadata not certified |
| 31 | ProviderNamespaceAbsenceV1 | provider-namespace-absence.v1 | 415,498,1079 | Indexed;full metadata not certified |
| 32 | AuthorityGenesisEpochEvidenceV1 | authority-genesis-epoch-evidence.v1 | 419,421,728,825,918 | Indexed;full metadata not certified |
| 33 | AuthorityEpochTransitionIntentV1 | authority-epoch-transition-intent.v1 | 432 | Indexed;full metadata not certified |
| 34 | RequestAuthorityEpochTransitionV1 | request-authority-epoch-transition.v1 | 434 | Indexed;full metadata not certified |
| 35 | AuthorityEpochTransitionV1 | authority-epoch-transition.v1 | 434,448 | Indexed;full metadata not certified |
| 36 | AuthorityEpochTransitionApplicationV1 | authority-epoch-transition-application.v1 | 436 | Indexed;full metadata not certified |
| 37 | AuthorityEpochTransitionResolutionV1 | authority-epoch-transition-resolution.v1 | 436 | Indexed;full metadata not certified |
| 38 | ResolveAuthorityEpochTransitionRequestV1 | resolve-authority-epoch-transition-request.v1 | 438 | Indexed;full metadata not certified |
| 39 | AuthorityEpochTransitionResolutionReceiptV1 | authority-epoch-transition-resolution-receipt.v1 | 438,448 | Indexed;full metadata not certified |
| 40 | AuthorityEpochTransitionCancellationV1 | authority-epoch-transition-cancellation.v1 | 444,460,740,756 | Indexed;full metadata not certified |
| 41 | AuthorityEpochTransitionTerminalApplicationV1 | authority-epoch-transition-terminal-application.v1 | 371,432,442,740,756 | Indexed;full metadata not certified |
| 42 | InspectAuthorityEpochTransitionRequestV1 | inspect-authority-epoch-transition-request.v1 | 448 | Indexed;full metadata not certified |
| 43 | InspectAuthorityEpochTransitionResponseV1 | inspect-authority-epoch-transition-response.v1 | 383,448 | Indexed;full metadata not certified |
| 44 | AuthorityEpochTransitionRecoveryIntentV1 | authority-epoch-transition-recovery-intent.v1 | 452 | Indexed;full metadata not certified |
| 45 | RecoverAuthorityEpochTransitionRequestV1 | recover-authority-epoch-transition-request.v1 | 379,454,610,683,736 | Indexed;full metadata not certified |
| 46 | AuthorityEpochTransitionRecoveryV1 | authority-epoch-transition-recovery.v1 | 379,456 | Indexed;full metadata not certified |
| 47 | AuthorityEpochTransitionRecoveryPrefixDecisionV1 | authority-epoch-transition-recovery-prefix-decision.v1 | 476,608,637 | Indexed;full metadata not certified |
| 48 | AuthorityEpochTransitionRecoveryApplicationV1 | authority-epoch-transition-recovery-application.v1 | 371,464,637 | Indexed;full metadata not certified |
| 49 | EmptyDatabaseInventoryV1 | empty-database-inventory.v1 | 486,490 | Indexed;full metadata not certified |
| 50 | AuthorityProtocolUpgradeIntentV1 | authority-protocol-upgrade-intent.v1 | 488 | Indexed;full metadata not certified |
| 51 | AuthorityProtocolUpgradeAttemptV1 | authority-protocol-upgrade-attempt.v1 | 492,597 | Indexed;full metadata not certified |
| 52 | PrepareClaimV1GenesisRequestV1 | prepare-claim-v1-genesis-request.v1 | 496 | Indexed;full metadata not certified |
| 53 | ClaimV1GenesisPreparationV1 | claim-v1-genesis-preparation.v1 | 494,496 | Indexed;full metadata not certified |
| 54 | AuthorityProtocolActivationV1 | authority-protocol-activation.v1 | 553,599 | Indexed;full metadata not certified |
| 55 | AuthorityProtocolActivationCompletionV1 | authority-protocol-activation-completion.v1 | 555,600 | Indexed;full metadata not certified |
| 56 | AuthorityProtocolActivationReleaseV1 | authority-protocol-activation-release.v1 | 557,601 | Indexed;full metadata not certified |
| 57 | AcquireDatabaseCommitChallengeRequestV1 | acquire-database-commit-challenge-request.v1 | 604 | Indexed;full metadata not certified |
| 58 | DatabaseCommitChallengeV1 | database-commit-challenge.v1 | 604,839 | Indexed;full metadata not certified |
| 59 | DatabaseCommitAttestationV1 | database-commit-attestation.v1 | 301,438,454,458,466,599,601,629,839,855,875 | Indexed;full metadata not certified |
| 60 | FenceDatabaseCommitAttestationArchiveSemanticIDJournalRecordV1 | database-commit-archive-fence-semantic-id-journal-record.v1 | 641 | Indexed;full metadata not certified |
| 61 | FenceDatabaseCommitAttestationArchiveRequestV1 | fence-database-commit-attestation-archive-request.v1 | 637 | Indexed;full metadata not certified |
| 62 | FenceDatabaseCommitAttestationArchiveResponseV1 | fence-database-commit-attestation-archive-response.v1 | 606,657 | Indexed;full metadata not certified |
| 63 | InspectDatabaseCommitAttestationArchiveRequestV1 | inspect-database-commit-attestation-archive-request.v1 | 659 | Indexed;full metadata not certified |
| 64 | InspectDatabaseCommitAttestationArchiveResponseV1 | inspect-database-commit-attestation-archive-response.v1 | 606,661 | Indexed;full metadata not certified |
| 65 | ClaimV1GenesisCompletionV1 | claim-v1-genesis-completion.v1 | 675 | Indexed;full metadata not certified |
| 66 | CompleteClaimV1GenesisRequestV1 | complete-claim-v1-genesis-request.v1 | 675 | Indexed;full metadata not certified |
| 67 | ClaimV1GenesisReleasePreparationV1 | claim-v1-genesis-release-preparation.v1 | 677 | Indexed;full metadata not certified |
| 68 | PrepareClaimV1GenesisReleaseRequestV1 | prepare-claim-v1-genesis-release-request.v1 | 677 | Indexed;full metadata not certified |
| 69 | OpenClaimV1GenesisRequestV1 | open-claim-v1-genesis-request.v1 | 679 | Indexed;full metadata not certified |
| 70 | ClaimV1GenesisOpenV1 | claim-v1-genesis-open.v1 | 679,689,769,967 | Indexed;full metadata not certified |
| 71 | InspectClaimV1GenesisRequestV1 | inspect-claim-v1-genesis-request.v1 | 681 | Indexed;full metadata not certified |
| 72 | InspectClaimV1GenesisResponseV1 | inspect-claim-v1-genesis-response.v1 | 681 | Indexed;full metadata not certified |
| 73 | ClaimV1ProviderHeadV1 | claim-v1-provider-head.v1 | 639,687,689,922 | Indexed;full metadata not certified |
| 74 | AcquireDatabaseIncarnationProofRequestV1 | acquire-database-incarnation-proof-request.v1 | 709 | Indexed;full metadata not certified |
| 75 | DatabaseIncarnationProofV1 | database-incarnation-proof.v1 | 397,454,709,724,915,937 | Indexed;full metadata not certified |
| 76 | DatabaseRuntimeInstanceLeaseDescendantChainV1 | database-runtime-instance-lease-descendant-chain.v1 | 707 | Indexed;full metadata not certified |
| 77 | AuthorityProtocolContextV1 | authority-protocol-context.v1 | 713,719 | Indexed;full metadata not certified |
| 78 | RenewClaimV1ServingLeaseRequestV1 | renew-claim-v1-serving-lease-request.v1 | 730 | Indexed;full metadata not certified |
| 79 | ClaimV1ServingLeaseV1 | claim-v1-serving-lease.v1 | 728,734,742 | Indexed;full metadata not certified |
| 80 | RevokeClaimV1ServingLeaseRequestV1 | revoke-claim-v1-serving-lease-request.v1 | 744 | Indexed;full metadata not certified |
| 81 | ClaimV1ServingLeaseRevocationV1 | claim-v1-serving-lease-revocation.v1 | 744 | Indexed;full metadata not certified |
| 82 | DatabaseRuntimeRebindAuthorizationV1 | database-runtime-rebind-authorization.v1 | 381 | Indexed;full metadata not certified |
| 83 | RebindClaimV1RuntimeInstanceRequestV1 | rebind-claim-v1-runtime-instance-request.v1 | 746,752 | Indexed;full metadata not certified |
| 84 | ClaimV1RuntimeInstanceRebindV1 | claim-v1-runtime-instance-rebind.v1 | 395,448,691,746 | Indexed;full metadata not certified |
| 85 | ClaimV1ReserveRequestV1 | claim-v1-reserve-request.v1 | 715 | Indexed;full metadata not certified |
| 86 | ClaimV1ReservationV1 | claim-v1-reservation.v1 | 716,719,732 | Indexed;full metadata not certified |
| 87 | ClaimV1FinalizeRequestV1 | claim-v1-finalize-request.v1 | 717 | Indexed;full metadata not certified |
| 88 | ClaimV1AbortRequestV1 | claim-v1-abort-request.v1 | 718 | Indexed;full metadata not certified |
| 89 | ClaimV1ReceiptV1 | claim-v1-receipt.v1 | 719,720 | Indexed;full metadata not certified |
| 90 | ClaimV1RecordV1 | claim-v1-record.v1 | 720,732 | Indexed;full metadata not certified |
| 91 | IndeterminateSourceSealV1 | indeterminate-source-seal.v1 | 586,782,784,1139 | Indexed;full metadata not certified |
| 92 | LegacySourceSealV1 | legacy-source-seal.v1 | 586,786 | Indexed;full metadata not certified |
| 93 | LegacySourceInventoryV1 | legacy-source-inventory.v1 | 795 | Indexed;full metadata not certified |
| 94 | LegacyEpochSourceObservationV1 | legacy-epoch-source-observation.v1 | 803,805 | Indexed;full metadata not certified |
| 95 | LegacyProviderEpochTranscriptV1 | legacy-provider-epoch-transcript.v1 | 799 | Indexed;full metadata not certified |
| 96 | LegacyDeploymentEpochManifestV1 | legacy-deployment-epoch-manifest.v1 | 801 | Indexed;full metadata not certified |
| 97 | LegacyEpochSourceSetV1 | legacy-epoch-source-set.v1 | 428,805 | Indexed;full metadata not certified |
| 98 | LegacyEpochMaximumEvidenceV1 | legacy-epoch-maximum-evidence.v1 | 807 | Indexed;full metadata not certified |
| 99 | LegacySourceRetirementAuthorizationV1 | legacy-source-retirement-authorization.v1 | 335 | Indexed;full metadata not certified |
| 100 | RetireLegacySourceMembershipRequestV1 | retire-legacy-source-membership-request.v1 | 337 | Indexed;full metadata not certified |
| 101 | LegacySourceMembershipRetirementV1 | legacy-source-membership-retirement.v1 | 337,428,786 | Indexed;full metadata not certified |
| 102 | LegacyDatabaseSourceRetirementV1 | legacy-database-source-retirement.v1 | 341,786 | Indexed;full metadata not certified |
| 103 | LegacyDatabaseSourceRetirementCommitAttestationV1 | legacy-database-source-retirement-commit-attestation.v1 | 341 | Indexed;full metadata not certified |
| 104 | RetireLegacyEnvironmentRequestV1 | retire-legacy-environment-request.v1 | 343 | Indexed;full metadata not certified |
| 105 | LegacyEnvironmentRetirementV1 | legacy-environment-retirement.v1 | 343 | Indexed;full metadata not certified |
| 106 | RetireLegacyProviderNamespaceRequestV1 | retire-legacy-provider-namespace-request.v1 | 345 | Indexed;full metadata not certified |
| 107 | LegacyProviderNamespaceRetirementV1 | legacy-provider-namespace-retirement.v1 | 345 | Indexed;full metadata not certified |
| 108 | LegacySourceRetirementSetV1 | legacy-source-retirement-set.v1 | 347,428,786 | Indexed;full metadata not certified |
| 109 | LegacySourcePostSealObservationV1 | legacy-source-post-seal-observation.v1 | 811 | Indexed;full metadata not certified |
| 110 | LegacySourcePostSealExactCoverV1 | legacy-source-post-seal-exact-cover.v1 | 428,813 | Indexed;full metadata not certified |
| 111 | LegacySourceSealCommitAttestationV1 | legacy-source-seal-commit-attestation.v1 | 809 | Indexed;full metadata not certified |
| 112 | LegacySourceArchiveEvidenceV1 | legacy-source-archive-evidence.v1 | 815 | Indexed;full metadata not certified |
| 113 | FreshRestoreRequirementV1 | fresh-restore-requirement.v1 | 586,780,817,1139 | Indexed;full metadata not certified |
| 114 | FreshRestoreImportManifestV1 | fresh-restore-import-manifest.v1 | 833,1026,1138 | Indexed;full metadata not certified |
| 115 | FreshImportTopologyProjectionV1 | fresh-import-topology-projection.v1 | 831,837,1150 | Indexed;full metadata not certified |
| 116 | StagingImportCapabilityV1 | staging-import-capability.v1 | 588,839 | Indexed;full metadata not certified |
| 117 | StagingImportCapabilityRevocationV1 | staging-import-capability-revocation.v1 | 869 | Indexed;full metadata not certified |
| 118 | StagingImportCapabilityRevocationApplicationV1 | staging-import-capability-revocation-application.v1 | 637,685,859,869,871 | Indexed;full metadata not certified |
| 119 | StagingImportCapabilityRecoveryIntentV1 | staging-import-capability-recovery-intent.v1 | 847 | Indexed;full metadata not certified |
| 120 | RecoverStagingImportCapabilityRequestV1 | recover-staging-import-capability-request.v1 | 851 | Indexed;full metadata not certified |
| 121 | StagingImportCapabilityRecoveryV1 | staging-import-capability-recovery.v1 | 853 | Indexed;full metadata not certified |
| 122 | StagingImportCapabilityRecoveryApplicationV1 | staging-import-capability-recovery-application.v1 | 637,855,859 | Indexed;full metadata not certified |
| 123 | AcquireFreshV7StagingExclusionRequestV1 | acquire-fresh-v7-staging-exclusion-request.v1 | 841 | Indexed;full metadata not certified |
| 124 | FreshV7StagingExclusionLeaseV1 | fresh-v7-staging-exclusion-lease.v1 | 841 | Indexed;full metadata not certified |
| 125 | FreshRestoreImportApplicationV1 | fresh-restore-import-application.v1 | 637,653,685,845,871,1105,1145 | Indexed;full metadata not certified |
| 126 | AcquireStagingDatabaseCommitChallengeRequestV1 | acquire-staging-database-commit-challenge-request.v1 | 871,873 | Indexed;full metadata not certified |
| 127 | StagingDatabaseCommitChallengeV1 | staging-database-commit-challenge.v1 | 875 | Indexed;full metadata not certified |
| 128 | StagingDatabaseCommitAttestationV1 | staging-database-commit-attestation.v1 | 301,871,875 | Indexed;full metadata not certified |
| 129 | ReleaseFreshV7StagingExclusionRequestV1 | release-fresh-v7-staging-exclusion-request.v1 | 877 | Indexed;full metadata not certified |
| 130 | FreshV7StagingExclusionReleaseV1 | fresh-v7-staging-exclusion-release.v1 | 877 | Indexed;full metadata not certified |
| 131 | AbortFreshV7StagingExclusionRequestV1 | abort-fresh-v7-staging-exclusion-request.v1 | 879 | Indexed;full metadata not certified |
| 132 | FreshV7StagingExclusionAbortV1 | fresh-v7-staging-exclusion-abort.v1 | 879 | Indexed;full metadata not certified |
| 133 | FreshTargetInventoryV1 | fresh-target-inventory.v1 | 881,883 | Indexed;full metadata not certified |
| 134 | FreshV7StagingEvidenceV1 | fresh-v7-staging-evidence.v1 | 889 | Indexed;full metadata not certified |
| 135 | RetireEnvironmentInventoryMembershipRequestV1 | retire-environment-inventory-membership-request.v1 | 991 | Indexed;full metadata not certified |
| 136 | EnvironmentInventoryMembershipRetirementV1 | environment-inventory-membership-retirement.v1 | 991 | Indexed;full metadata not certified |
| 137 | ProviderProtocolDowngradeRetirementAuthorizationV1 | provider-protocol-downgrade-retirement-authorization.v1 | 993 | Indexed;full metadata not certified |
| 138 | ProviderProtocolHistoryZeroProjectionV1 | provider-protocol-history-zero-projection.v1 | 999 | Indexed;full metadata not certified |
| 139 | RetireProviderProtocolForDowngradeRequestV1 | retire-provider-protocol-for-downgrade-request.v1 | 995 | Indexed;full metadata not certified |
| 140 | ProviderProtocolDowngradeRetirementV1 | provider-protocol-downgrade-retirement.v1 | 1001 | Indexed;full metadata not certified |
| 141 | ProviderProtocolDowngradeRetirementSetV1 | provider-protocol-downgrade-retirement-set.v1 | 1003 | Indexed;full metadata not certified |
| 142 | PristineDowngradeInventoryV1 | pristine-downgrade-inventory.v1 | 1005 | Indexed;full metadata not certified |
| 143 | PristineDowngradeAuthorizedStateV1 | pristine-downgrade-authorized-state.v1 | 1007 | Indexed;full metadata not certified |
| 144 | PristineDowngradeConsumeReceiptV1 | pristine-downgrade-consume-receipt.v1 | 1009 | Indexed;full metadata not certified |
| 145 | AuthorityProtocolMigrationLatchV1 | authority-protocol-migration-latch.v1 | 537 | Indexed;full metadata not certified |
| 146 | AuthorityProtocolDowngradeAuthorizationV1 | authority-protocol-downgrade-authorization.v1 | 1007 | Indexed;full metadata not certified |

## Recovery

Use latest codex/c12-b01-task9-coordinator tip,not main. Source/name index is
portable;raw SDK/cache/binary/provider state is not restored by clone. Resume
at the grouped C2 design decision,not code execution or historical Task11 fuzz
commands. Earlier147 references in historical entries are superseded by this
explicit count correction,not evidence that one protocol object was removed.
