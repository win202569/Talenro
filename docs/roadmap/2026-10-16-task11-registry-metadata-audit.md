# Task11：146项元数据逐项初审与集中例外

基线f6814efacad1cd7fe3787068611ab26f86976b39。用户随后批准
[C2统一准入规则](../superpowers/specs/2026-10-16-authority-v7-metadata-admission-clarification.md)，
本轮按该规则逐项登记现有证据与未决问题。这里只是计划审查附表，**不是产品注册表、
完整规范认证或可执行实施计划**；没有代码、fixture、测试或真实负载执行。

## 初审结果及其边界

完整146项名称/object均有一行，93项已有局部分类依据，53项尚未闭合。
其中4项保留已知请求角色但不猜测其证据资格。未决数量是本轮审查状态，**不是
已证明的协议缺陷数量，也不表示后续审查无法从已有规格继续关闭其中部分项目**。
已裁决项是作者依原规格与本地已批C1/C2的判断，未作独立审查，不是生产信任批准。

名称集合来自原v7 §4.1，仍是146个V1对象；旧147计表头错误已在来源索引更正。
角色来自具体对象定义或明确“同一角色”的先行引用，绝不从同段所有角色词自动分配。
external_signed_envelope表示类型可用形式，不是所有message都要/允许该项；
每message exact-cover、key/root/policy/algorithm和时效/状态仍独立验证。
database_immutable_body行还须对应真实已提交不可变preimage与特定消费契约，
临时投影、DB存储位置或摘要本身不够。null/[]行不赋予请求免认证权。
“未决”文字只存在于此审查表，不得进入生产descriptor或变成运行时fallback。

## 来源与读法

原规格：[v7升级修订](../superpowers/specs/2026-08-24-nodecontrol-authority-v7-upgrade-design.md)。
L编号以该基线文件为准；“全部引用”来自[完整来源索引](2026-10-16-task11-registry-reconciliation.md)，
可能含cross-reference而非定义。分类依据列指出实际判断，不把索引命中当policy证明。
C1仅为本地设计依据，未完成canonical集合登记/部署兼容；C2也不改变原批准规格字节。

| # | Object / schema | 原规格全部引用行 | SignerRoleOrNull | EvidenceKinds | 判定依据/尚缺事项 | 状态 |
| ---: | --- | --- | --- | --- | --- | --- |
| 1 | CanonicalEvidenceBundleV1<br>canonical-evidence-bundle.v1 | 333,367,375,381,732,734,803,871,999,1003 | null | [] | §4.1 container，不为自身生成DB/外部签名证据；完整类别定义：只作嵌入/集合/临时投影；未继承父角色 | 有据已裁决 |
| 2 | EnvironmentInventoryV1<br>environment-inventory.v1 | 282,284,309,319 | release_deployment_operator | external_signed_envelope | §4.1 + L284；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 3 | EnvironmentInventoryRecordV1<br>environment-inventory-record.v1 | 291,295 | null | [] | L291/295：inventory嵌入体及其内容摘要；完整类别定义：只作嵌入/集合/临时投影；未继承父角色 | 有据已裁决 |
| 4 | DeploymentEnvironmentAttestationV1<br>deployment-environment-attestation.v1 | 293,295 | deployment_environment_attestor | external_signed_envelope | §4.1 + L293；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 5 | EnvironmentInventoryAnchorV1<br>environment-inventory-anchor.v1 | 305 | claim_v1_inventory_anchor | external_signed_envelope | §4.1 + L305；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 6 | EnvironmentInventoryAnchorSetV1<br>environment-inventory-anchor-set.v1 | 307 | null | [] | L307：集合引用anchor；签名资格属于成员；完整类别定义：只作嵌入/集合/临时投影；未继承父角色 | 有据已裁决 |
| 7 | PublishClaimV1CredentialPolicyRequestV1<br>publish-claim-v1-credential-policy-request.v1 | 325 | claim_v1_security_admin_request | [] | 本地已批C1 §3–5；请求认证保持mandatory；无bundle证据资格 | 有据已裁决 |
| 8 | ClaimV1CredentialPolicyV1<br>claim-v1-credential-policy.v1 | 321,325,734 | claim_v1_security_policy | external_signed_envelope | §4.1 + L325；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 9 | LocalRuntimeIsolationV1<br>local-runtime-isolation.v1 | 315,317,331,351,782 | deployment_isolation_operator | external_signed_envelope | §4.1 + L315；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 10 | LegacyRuntimeShutdownV1<br>legacy-runtime-shutdown.v1 | 322,331 | release_deployment_operator | external_signed_envelope | §4.1 + L331；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 11 | LegacyRuntimeShutdownSetV1<br>legacy-runtime-shutdown-set.v1 | 333 | null | [] | L333：bundle携带成员shutdown/isolation envelopes；完整类别定义：只作嵌入/集合/临时投影；未继承父角色 | 有据已裁决 |
| 12 | AcquireDatabaseIncarnationRegistrationChallengeRequestV1<br>acquire-database-incarnation-registration-challenge-request.v1 | 355 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 13 | DatabaseIncarnationRegistrationChallengeV1<br>database-incarnation-registration-challenge.v1 | 355 | claim_v1_provider | external_signed_envelope | §4.1 + L355；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 14 | DatabaseIncarnationAttestationV1<br>database-incarnation-attestation.v1 | 353,357 | database_incarnation_attestor | external_signed_envelope | §4.1 + L357；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 15 | DatabaseRuntimeInstanceBindingV1<br>database-runtime-instance-binding.v1 | 359 | database_runtime_attestor | external_signed_envelope | §4.1 + L359；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 16 | RenewDatabaseRuntimeInstanceLeaseRequestV1<br>renew-database-runtime-instance-lease-request.v1 | 361 | null | [] | L361：mutually-authenticated local attestor channel；明确本地holder认证渠道；不转成独立签名证据，不取消认证 | 有据已裁决 |
| 17 | DatabaseRuntimeInstanceLeaseV1<br>database-runtime-instance-lease.v1 | 363 | database_runtime_attestor | external_signed_envelope | §4.1 + L363；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 18 | DatabaseRuntimeInstanceLeaseTerminationV1<br>database-runtime-instance-lease-termination.v1 | 365 | database_runtime_attestor | external_signed_envelope | §4.1 + L365；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 19 | DatabaseTimelineLineageAttestationV1<br>database-timeline-lineage-attestation.v1 | 297,393,537 | database_timeline_lineage_attestor | external_signed_envelope | §4.1 + L297；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 20 | RegisterDatabaseIncarnationRequestV1<br>register-database-incarnation-request.v1 | 367 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 21 | DatabaseIncarnationRegistrationV1<br>database-incarnation-registration.v1 | 367 | claim_v1_incarnation_registry | external_signed_envelope | §4.1 + L367；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 22 | DatabaseIncarnationRebindingRegistrationV1<br>database-incarnation-rebinding-registration.v1 | 395,752,754,1111 | claim_v1_incarnation_registry | external_signed_envelope | §4.1 + L395；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 23 | AuthorityRuntimeRegistrationResultV1<br>authority-runtime-registration-result.v1 | 369,596 | null | database_immutable_body | L369,L713；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 24 | AuthorityRuntimeRebindResultV1<br>authority-runtime-rebind-result.v1 | 301,393,444,454,460,596,736,740,754 | null | database_immutable_body | L393,L537,L713；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 25 | DatabaseAuthorityHeadV1<br>database-authority-head.v1 | 303,371,375 | null | [] | L375：显式禁止把临时Head伪装为DB immutable body；完整类别定义：只作嵌入/集合/临时投影；未继承父角色 | 有据已裁决 |
| 26 | DatabaseAuthorityHeadAttestationV1<br>database-authority-head-attestation.v1 | 301,375,454,458,474,734,1116 | database_incarnation | external_signed_envelope | §4.1 + L375；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 27 | DatabaseAuthorityRebindGapAttestationV1<br>database-authority-rebind-gap-attestation.v1 | 377,450,454,480,537,608,711,1015 | database_runtime_attestor | external_signed_envelope | §4.1 + L377；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 28 | DatabaseAuthorityPostRecoveryRebindAttestationV1<br>database-authority-post-recovery-rebind-attestation.v1 | 301,399,537,711,1015 | database_runtime_attestor | external_signed_envelope | §4.1 + L399；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 29 | DatabaseLegacyAbsenceProjectionV1<br>database-legacy-absence-projection.v1 | 409,599 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 30 | DatabaseLegacyAbsenceObservationV1<br>database-legacy-absence-observation.v1 | 413 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 31 | ProviderNamespaceAbsenceV1<br>provider-namespace-absence.v1 | 415,498,1079 | claim_v1_provider | external_signed_envelope | §4.1 + L415；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 32 | AuthorityGenesisEpochEvidenceV1<br>authority-genesis-epoch-evidence.v1 | 419,421,728,825,918 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 33 | AuthorityEpochTransitionIntentV1<br>authority-epoch-transition-intent.v1 | 432 | null | database_immutable_body | L432,L734；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 34 | RequestAuthorityEpochTransitionV1<br>request-authority-epoch-transition.v1 | 434 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 35 | AuthorityEpochTransitionV1<br>authority-epoch-transition.v1 | 434,448 | claim_v1_provider | external_signed_envelope | §4.1 + L434；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 36 | AuthorityEpochTransitionApplicationV1<br>authority-epoch-transition-application.v1 | 436 | null | database_immutable_body | L436,L734；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 37 | AuthorityEpochTransitionResolutionV1<br>authority-epoch-transition-resolution.v1 | 436 | null | database_immutable_body | L436,L734；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 38 | ResolveAuthorityEpochTransitionRequestV1<br>resolve-authority-epoch-transition-request.v1 | 438 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 39 | AuthorityEpochTransitionResolutionReceiptV1<br>authority-epoch-transition-resolution-receipt.v1 | 438,448 | claim_v1_provider | external_signed_envelope | §4.1 + L438；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 40 | AuthorityEpochTransitionCancellationV1<br>authority-epoch-transition-cancellation.v1 | 444,460,740,756 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 41 | AuthorityEpochTransitionTerminalApplicationV1<br>authority-epoch-transition-terminal-application.v1 | 371,432,442,740,756 | null | database_immutable_body | L442,L734；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 42 | InspectAuthorityEpochTransitionRequestV1<br>inspect-authority-epoch-transition-request.v1 | 448 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 43 | InspectAuthorityEpochTransitionResponseV1<br>inspect-authority-epoch-transition-response.v1 | 383,448 | claim_v1_provider | external_signed_envelope | L448，同一角色追溯L438 + §4.1；明示签名/同一角色引用；message-specific exact-cover另验 | 有据已裁决 |
| 44 | AuthorityEpochTransitionRecoveryIntentV1<br>authority-epoch-transition-recovery-intent.v1 | 452 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 45 | RecoverAuthorityEpochTransitionRequestV1<br>recover-authority-epoch-transition-request.v1 | 379,454,610,683,736 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 46 | AuthorityEpochTransitionRecoveryV1<br>authority-epoch-transition-recovery.v1 | 379,456 | claim_v1_provider | external_signed_envelope | §4.1 + L456；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 47 | AuthorityEpochTransitionRecoveryPrefixDecisionV1<br>authority-epoch-transition-recovery-prefix-decision.v1 | 476,608,637 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 48 | AuthorityEpochTransitionRecoveryApplicationV1<br>authority-epoch-transition-recovery-application.v1 | 371,464,637 | null | database_immutable_body | L464,L466,L537；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 49 | EmptyDatabaseInventoryV1<br>empty-database-inventory.v1 | 486,490 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 50 | AuthorityProtocolUpgradeIntentV1<br>authority-protocol-upgrade-intent.v1 | 488 | null | database_immutable_body | L488,L492；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 51 | AuthorityProtocolUpgradeAttemptV1<br>authority-protocol-upgrade-attempt.v1 | 492,597 | null | database_immutable_body | L492,L496；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 52 | PrepareClaimV1GenesisRequestV1<br>prepare-claim-v1-genesis-request.v1 | 496 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 53 | ClaimV1GenesisPreparationV1<br>claim-v1-genesis-preparation.v1 | 494,496 | claim_v1_provider | external_signed_envelope | §4.1 + L496；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 54 | AuthorityProtocolActivationV1<br>authority-protocol-activation.v1 | 553,599 | null | database_immutable_body | L553,L599；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 55 | AuthorityProtocolActivationCompletionV1<br>authority-protocol-activation-completion.v1 | 555,600 | null | database_immutable_body | L555,L600；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 56 | AuthorityProtocolActivationReleaseV1<br>authority-protocol-activation-release.v1 | 557,601 | null | database_immutable_body | L557,L601；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 57 | AcquireDatabaseCommitChallengeRequestV1<br>acquire-database-commit-challenge-request.v1 | 604 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 58 | DatabaseCommitChallengeV1<br>database-commit-challenge.v1 | 604,839 | claim_v1_provider | external_signed_envelope | §4.1 + L604；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 59 | DatabaseCommitAttestationV1<br>database-commit-attestation.v1 | 301,438,454,458,466,599,601,629,839,855,875 | database_incarnation | external_signed_envelope | §4.1 + L629；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 60 | FenceDatabaseCommitAttestationArchiveSemanticIDJournalRecordV1<br>database-commit-archive-fence-semantic-id-journal-record.v1 | 641 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 61 | FenceDatabaseCommitAttestationArchiveRequestV1<br>fence-database-commit-attestation-archive-request.v1 | 637 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 62 | FenceDatabaseCommitAttestationArchiveResponseV1<br>fence-database-commit-attestation-archive-response.v1 | 606,657 | database_incarnation | external_signed_envelope | §4.1 + L657；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 63 | InspectDatabaseCommitAttestationArchiveRequestV1<br>inspect-database-commit-attestation-archive-request.v1 | 659 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 64 | InspectDatabaseCommitAttestationArchiveResponseV1<br>inspect-database-commit-attestation-archive-response.v1 | 606,661 | database_incarnation | external_signed_envelope | §4.1 + L661；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 65 | ClaimV1GenesisCompletionV1<br>claim-v1-genesis-completion.v1 | 675 | claim_v1_provider | external_signed_envelope | §4.1 + L675；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 66 | CompleteClaimV1GenesisRequestV1<br>complete-claim-v1-genesis-request.v1 | 675 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 67 | ClaimV1GenesisReleasePreparationV1<br>claim-v1-genesis-release-preparation.v1 | 677 | claim_v1_provider | external_signed_envelope | §4.1 + L677；同一角色追溯L675；外部响应 | 有据已裁决 |
| 68 | PrepareClaimV1GenesisReleaseRequestV1<br>prepare-claim-v1-genesis-release-request.v1 | 677 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 69 | OpenClaimV1GenesisRequestV1<br>open-claim-v1-genesis-request.v1 | 679 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 70 | ClaimV1GenesisOpenV1<br>claim-v1-genesis-open.v1 | 679,689,769,967 | claim_v1_provider | external_signed_envelope | §4.1 + L679；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 71 | InspectClaimV1GenesisRequestV1<br>inspect-claim-v1-genesis-request.v1 | 681 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 72 | InspectClaimV1GenesisResponseV1<br>inspect-claim-v1-genesis-response.v1 | 681 | claim_v1_provider | external_signed_envelope | §4.1 + L681；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 73 | ClaimV1ProviderHeadV1<br>claim-v1-provider-head.v1 | 639,687,689,922 | claim_v1_provider | external_signed_envelope | §4.1 + L689；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 74 | AcquireDatabaseIncarnationProofRequestV1<br>acquire-database-incarnation-proof-request.v1 | 709 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 75 | DatabaseIncarnationProofV1<br>database-incarnation-proof.v1 | 397,454,709,724,915,937 | claim_v1_provider | external_signed_envelope | §4.1 + L709；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 76 | DatabaseRuntimeInstanceLeaseDescendantChainV1<br>database-runtime-instance-lease-descendant-chain.v1 | 707 | database_runtime_attestor | external_signed_envelope | §4.1 + L707；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 77 | AuthorityProtocolContextV1<br>authority-protocol-context.v1 | 713,719 | null | [] | L713/719：请求字段/receipt嵌入context，非独立证据；完整类别定义：只作嵌入/集合/临时投影；未继承父角色 | 有据已裁决 |
| 78 | RenewClaimV1ServingLeaseRequestV1<br>renew-claim-v1-serving-lease-request.v1 | 730 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 79 | ClaimV1ServingLeaseV1<br>claim-v1-serving-lease.v1 | 728,734,742 | claim_v1_provider | external_signed_envelope | §4.1 + L728；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 80 | RevokeClaimV1ServingLeaseRequestV1<br>revoke-claim-v1-serving-lease-request.v1 | 744 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 81 | ClaimV1ServingLeaseRevocationV1<br>claim-v1-serving-lease-revocation.v1 | 744 | claim_v1_provider | external_signed_envelope | §4.1 + L744；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 82 | DatabaseRuntimeRebindAuthorizationV1<br>database-runtime-rebind-authorization.v1 | 381 | runtime_rebind_operator | external_signed_envelope | §4.1 + L381；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 83 | RebindClaimV1RuntimeInstanceRequestV1<br>rebind-claim-v1-runtime-instance-request.v1 | 746,752 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 84 | ClaimV1RuntimeInstanceRebindV1<br>claim-v1-runtime-instance-rebind.v1 | 395,448,691,746 | claim_v1_provider | external_signed_envelope | §4.1 + L746；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 85 | ClaimV1ReserveRequestV1<br>claim-v1-reserve-request.v1 | 715 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 86 | ClaimV1ReservationV1<br>claim-v1-reservation.v1 | 716,719,732 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 87 | ClaimV1FinalizeRequestV1<br>claim-v1-finalize-request.v1 | 717 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 88 | ClaimV1AbortRequestV1<br>claim-v1-abort-request.v1 | 718 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 89 | ClaimV1ReceiptV1<br>claim-v1-receipt.v1 | 719,720 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 90 | ClaimV1RecordV1<br>claim-v1-record.v1 | 720,732 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 91 | IndeterminateSourceSealV1<br>indeterminate-source-seal.v1 | 586,782,784,1139 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 92 | LegacySourceSealV1<br>legacy-source-seal.v1 | 586,786 | null | database_immutable_body | L786,L813；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 93 | LegacySourceInventoryV1<br>legacy-source-inventory.v1 | 795 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 94 | LegacyEpochSourceObservationV1<br>legacy-epoch-source-observation.v1 | 803,805 | legacy_epoch_auditor | external_signed_envelope | §4.1 + L803；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 95 | LegacyProviderEpochTranscriptV1<br>legacy-provider-epoch-transcript.v1 | 799 | legacy_epoch_provider | external_signed_envelope | §4.1 + L799；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 96 | LegacyDeploymentEpochManifestV1<br>legacy-deployment-epoch-manifest.v1 | 801 | release_deployment_operator | external_signed_envelope | §4.1 + L801；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 97 | LegacyEpochSourceSetV1<br>legacy-epoch-source-set.v1 | 428,805 | legacy_epoch_auditor | external_signed_envelope | §4.1 + L805；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 98 | LegacyEpochMaximumEvidenceV1<br>legacy-epoch-maximum-evidence.v1 | 807 | legacy_epoch_auditor | external_signed_envelope | §4.1 + L807；同一角色追溯L805；外部签名对象 | 有据已裁决 |
| 99 | LegacySourceRetirementAuthorizationV1<br>legacy-source-retirement-authorization.v1 | 335 | legacy_source_retirement_admin | external_signed_envelope | §4.1 + L335；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 100 | RetireLegacySourceMembershipRequestV1<br>retire-legacy-source-membership-request.v1 | 337 | legacy_source_retirement_admin | 未决（不是[]） | L337，同一角色追溯L335；E：请求角色已明示；独立bundle证据资格未闭合 | 未决 |
| 101 | LegacySourceMembershipRetirementV1<br>legacy-source-membership-retirement.v1 | 337,428,786 | release_deployment_operator | external_signed_envelope | §4.1 + L337；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 102 | LegacyDatabaseSourceRetirementV1<br>legacy-database-source-retirement.v1 | 341,786 | null | database_immutable_body | L341,L803；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 103 | LegacyDatabaseSourceRetirementCommitAttestationV1<br>legacy-database-source-retirement-commit-attestation.v1 | 341 | legacy_source_archive_attestor | external_signed_envelope | §4.1 + L341；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 104 | RetireLegacyEnvironmentRequestV1<br>retire-legacy-environment-request.v1 | 343 | legacy_source_retirement_admin | 未决（不是[]） | L343；E：请求角色已明示；独立bundle证据资格未闭合 | 未决 |
| 105 | LegacyEnvironmentRetirementV1<br>legacy-environment-retirement.v1 | 343 | deployment_environment_retirement_authority | external_signed_envelope | §4.1 + L343；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 106 | RetireLegacyProviderNamespaceRequestV1<br>retire-legacy-provider-namespace-request.v1 | 345 | legacy_source_retirement_admin | 未决（不是[]） | L345；E：请求角色已明示；独立bundle证据资格未闭合 | 未决 |
| 107 | LegacyProviderNamespaceRetirementV1<br>legacy-provider-namespace-retirement.v1 | 345 | claim_v1_provider | external_signed_envelope | §4.1 + L345；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 108 | LegacySourceRetirementSetV1<br>legacy-source-retirement-set.v1 | 347,428,786 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 109 | LegacySourcePostSealObservationV1<br>legacy-source-post-seal-observation.v1 | 811 | legacy_source_archive_attestor | external_signed_envelope | §4.1 + L811；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 110 | LegacySourcePostSealExactCoverV1<br>legacy-source-post-seal-exact-cover.v1 | 428,813 | release_deployment_operator | external_signed_envelope | §4.1 + L813；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 111 | LegacySourceSealCommitAttestationV1<br>legacy-source-seal-commit-attestation.v1 | 809 | legacy_source_archive_attestor | external_signed_envelope | L809 + §4.1；明示签名/同一角色引用；message-specific exact-cover另验 | 有据已裁决 |
| 112 | LegacySourceArchiveEvidenceV1<br>legacy-source-archive-evidence.v1 | 815 | release_deployment_operator | external_signed_envelope | §4.1 + L815；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 113 | FreshRestoreRequirementV1<br>fresh-restore-requirement.v1 | 586,780,817,1139 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 114 | FreshRestoreImportManifestV1<br>fresh-restore-import-manifest.v1 | 833,1026,1138 | fresh_restore_export_operator | external_signed_envelope | §4.1 + L833；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 115 | FreshImportTopologyProjectionV1<br>fresh-import-topology-projection.v1 | 831,837,1150 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 116 | StagingImportCapabilityV1<br>staging-import-capability.v1 | 588,839 | fresh_restore_staging_import_authorizer | external_signed_envelope | §4.1 + L839；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 117 | StagingImportCapabilityRevocationV1<br>staging-import-capability-revocation.v1 | 869 | fresh_restore_staging_import_revoker | external_signed_envelope | §4.1 + L869；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 118 | StagingImportCapabilityRevocationApplicationV1<br>staging-import-capability-revocation-application.v1 | 637,685,859,869,871 | null | database_immutable_body | L869,L875；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 119 | StagingImportCapabilityRecoveryIntentV1<br>staging-import-capability-recovery-intent.v1 | 847 | null | database_immutable_body | L847,L851；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 120 | RecoverStagingImportCapabilityRequestV1<br>recover-staging-import-capability-request.v1 | 851 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 121 | StagingImportCapabilityRecoveryV1<br>staging-import-capability-recovery.v1 | 853 | claim_v1_provider | external_signed_envelope | §4.1 + L853；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 122 | StagingImportCapabilityRecoveryApplicationV1<br>staging-import-capability-recovery-application.v1 | 637,855,859 | null | database_immutable_body | L855,L859；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 123 | AcquireFreshV7StagingExclusionRequestV1<br>acquire-fresh-v7-staging-exclusion-request.v1 | 841 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 124 | FreshV7StagingExclusionLeaseV1<br>fresh-v7-staging-exclusion-lease.v1 | 841 | claim_v1_provider | external_signed_envelope | §4.1 + L841；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 125 | FreshRestoreImportApplicationV1<br>fresh-restore-import-application.v1 | 637,653,685,845,871,1105,1145 | null | database_immutable_body | L845,L875；明确提交/不可变行及后续preimage消费；非DB自签 | 有据已裁决 |
| 126 | AcquireStagingDatabaseCommitChallengeRequestV1<br>acquire-staging-database-commit-challenge-request.v1 | 871,873 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 127 | StagingDatabaseCommitChallengeV1<br>staging-database-commit-challenge.v1 | 875 | claim_v1_provider | external_signed_envelope | §4.1 + L875；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 128 | StagingDatabaseCommitAttestationV1<br>staging-database-commit-attestation.v1 | 301,871,875 | database_incarnation | external_signed_envelope | §4.1 + L875；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 129 | ReleaseFreshV7StagingExclusionRequestV1<br>release-fresh-v7-staging-exclusion-request.v1 | 877 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 130 | FreshV7StagingExclusionReleaseV1<br>fresh-v7-staging-exclusion-release.v1 | 877 | claim_v1_provider | external_signed_envelope | L877 + §4.1；明示签名/同一角色引用；message-specific exact-cover另验 | 有据已裁决 |
| 131 | AbortFreshV7StagingExclusionRequestV1<br>abort-fresh-v7-staging-exclusion-request.v1 | 879 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 132 | FreshV7StagingExclusionAbortV1<br>fresh-v7-staging-exclusion-abort.v1 | 879 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 133 | FreshTargetInventoryV1<br>fresh-target-inventory.v1 | 881,883 | fresh_target_inventory_operator | external_signed_envelope | §4.1 + L883；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 134 | FreshV7StagingEvidenceV1<br>fresh-v7-staging-evidence.v1 | 889 | fresh_restore_evidence_operator | external_signed_envelope | §4.1 + L889；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 135 | RetireEnvironmentInventoryMembershipRequestV1<br>retire-environment-inventory-membership-request.v1 | 991 | provider_downgrade_retirement_admin | 未决（不是[]） | L991；E：请求角色已明示；独立bundle证据资格未闭合 | 未决 |
| 136 | EnvironmentInventoryMembershipRetirementV1<br>environment-inventory-membership-retirement.v1 | 991 | release_deployment_operator | external_signed_envelope | §4.1 + L991；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 137 | ProviderProtocolDowngradeRetirementAuthorizationV1<br>provider-protocol-downgrade-retirement-authorization.v1 | 993 | provider_downgrade_retirement_admin | external_signed_envelope | §4.1 + L993；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 138 | ProviderProtocolHistoryZeroProjectionV1<br>provider-protocol-history-zero-projection.v1 | 999 | claim_v1_provider_history_auditor | external_signed_envelope | §4.1 + L999；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 139 | RetireProviderProtocolForDowngradeRequestV1<br>retire-provider-protocol-for-downgrade-request.v1 | 995 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；R：请求认证/通用envelope角色或独立证据形式未同时闭合 | 未决 |
| 140 | ProviderProtocolDowngradeRetirementV1<br>provider-protocol-downgrade-retirement.v1 | 1001 | claim_v1_provider | external_signed_envelope | §4.1 + L1001；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |
| 141 | ProviderProtocolDowngradeRetirementSetV1<br>provider-protocol-downgrade-retirement-set.v1 | 1003 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 142 | PristineDowngradeInventoryV1<br>pristine-downgrade-inventory.v1 | 1005 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 143 | PristineDowngradeAuthorizedStateV1<br>pristine-downgrade-authorized-state.v1 | 1007 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 144 | PristineDowngradeConsumeReceiptV1<br>pristine-downgrade-consume-receipt.v1 | 1009 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 145 | AuthorityProtocolMigrationLatchV1<br>authority-protocol-migration-latch.v1 | 537 | 未决（不是null） | 未决（不是[]） | 来源索引中该object全部引用；P：精确角色/独立证据资格或不变preimage消费链未同时闭合 | 未决 |
| 146 | AuthorityProtocolDowngradeAuthorizationV1<br>authority-protocol-downgrade-authorization.v1 | 1007 | authority_protocol_downgrade_authorizer | external_signed_envelope | §4.1 + L1007；明示外部签名对象；bundle形式须再受message exact-cover约束 | 有据已裁决 |

## 一次性集中例外清单

R表示请求认证/通用封装/独立证据通道尚未全部闭合；P表示对象角色或独立证据/
不可变preimage消费链尚未全部闭合；E表示角色已明示但请求证据资格尚未闭合。
分组标签不是新协议枚举，不断言每一项都需新增契约。先复核精确定义/引用，再区分
“已有条款可解决”与“确实需新增裁决”；不能凭所属组批量填null或[]。

- AcquireDatabaseIncarnationRegistrationChallengeRequestV1（acquire-database-incarnation-registration-challenge-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行355。
- RegisterDatabaseIncarnationRequestV1（register-database-incarnation-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行367。
- DatabaseLegacyAbsenceProjectionV1（database-legacy-absence-projection.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行409,599。
- DatabaseLegacyAbsenceObservationV1（database-legacy-absence-observation.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行413。
- AuthorityGenesisEpochEvidenceV1（authority-genesis-epoch-evidence.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行419,421,728,825,918。
- RequestAuthorityEpochTransitionV1（request-authority-epoch-transition.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行434。
- ResolveAuthorityEpochTransitionRequestV1（resolve-authority-epoch-transition-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行438。
- AuthorityEpochTransitionCancellationV1（authority-epoch-transition-cancellation.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行444,460,740,756。
- InspectAuthorityEpochTransitionRequestV1（inspect-authority-epoch-transition-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行448。
- AuthorityEpochTransitionRecoveryIntentV1（authority-epoch-transition-recovery-intent.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行452。
- RecoverAuthorityEpochTransitionRequestV1（recover-authority-epoch-transition-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行379,454,610,683,736。
- AuthorityEpochTransitionRecoveryPrefixDecisionV1（authority-epoch-transition-recovery-prefix-decision.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行476,608,637。
- EmptyDatabaseInventoryV1（empty-database-inventory.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行486,490。
- PrepareClaimV1GenesisRequestV1（prepare-claim-v1-genesis-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行496。
- AcquireDatabaseCommitChallengeRequestV1（acquire-database-commit-challenge-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行604。
- FenceDatabaseCommitAttestationArchiveSemanticIDJournalRecordV1（database-commit-archive-fence-semantic-id-journal-record.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行641。
- FenceDatabaseCommitAttestationArchiveRequestV1（fence-database-commit-attestation-archive-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行637。
- InspectDatabaseCommitAttestationArchiveRequestV1（inspect-database-commit-attestation-archive-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行659。
- CompleteClaimV1GenesisRequestV1（complete-claim-v1-genesis-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行675。
- PrepareClaimV1GenesisReleaseRequestV1（prepare-claim-v1-genesis-release-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行677。
- OpenClaimV1GenesisRequestV1（open-claim-v1-genesis-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行679。
- InspectClaimV1GenesisRequestV1（inspect-claim-v1-genesis-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行681。
- AcquireDatabaseIncarnationProofRequestV1（acquire-database-incarnation-proof-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行709。
- RenewClaimV1ServingLeaseRequestV1（renew-claim-v1-serving-lease-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行730。
- RevokeClaimV1ServingLeaseRequestV1（revoke-claim-v1-serving-lease-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行744。
- RebindClaimV1RuntimeInstanceRequestV1（rebind-claim-v1-runtime-instance-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行746,752。
- ClaimV1ReserveRequestV1（claim-v1-reserve-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行715。
- ClaimV1ReservationV1（claim-v1-reservation.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行716,719,732。
- ClaimV1FinalizeRequestV1（claim-v1-finalize-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行717。
- ClaimV1AbortRequestV1（claim-v1-abort-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行718。
- ClaimV1ReceiptV1（claim-v1-receipt.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行719,720。
- ClaimV1RecordV1（claim-v1-record.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行720,732。
- IndeterminateSourceSealV1（indeterminate-source-seal.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行586,782,784,1139。
- LegacySourceInventoryV1（legacy-source-inventory.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行795。
- RetireLegacySourceMembershipRequestV1（retire-legacy-source-membership-request.v1）：E：请求角色已明示；独立bundle证据资格未闭合；源行337。
- RetireLegacyEnvironmentRequestV1（retire-legacy-environment-request.v1）：E：请求角色已明示；独立bundle证据资格未闭合；源行343。
- RetireLegacyProviderNamespaceRequestV1（retire-legacy-provider-namespace-request.v1）：E：请求角色已明示；独立bundle证据资格未闭合；源行345。
- LegacySourceRetirementSetV1（legacy-source-retirement-set.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行347,428,786。
- FreshRestoreRequirementV1（fresh-restore-requirement.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行586,780,817,1139。
- FreshImportTopologyProjectionV1（fresh-import-topology-projection.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行831,837,1150。
- RecoverStagingImportCapabilityRequestV1（recover-staging-import-capability-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行851。
- AcquireFreshV7StagingExclusionRequestV1（acquire-fresh-v7-staging-exclusion-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行841。
- AcquireStagingDatabaseCommitChallengeRequestV1（acquire-staging-database-commit-challenge-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行871,873。
- ReleaseFreshV7StagingExclusionRequestV1（release-fresh-v7-staging-exclusion-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行877。
- AbortFreshV7StagingExclusionRequestV1（abort-fresh-v7-staging-exclusion-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行879。
- FreshV7StagingExclusionAbortV1（fresh-v7-staging-exclusion-abort.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行879。
- RetireEnvironmentInventoryMembershipRequestV1（retire-environment-inventory-membership-request.v1）：E：请求角色已明示；独立bundle证据资格未闭合；源行991。
- RetireProviderProtocolForDowngradeRequestV1（retire-provider-protocol-for-downgrade-request.v1）：R：请求认证/通用envelope角色或独立证据形式未同时闭合；源行995。
- ProviderProtocolDowngradeRetirementSetV1（provider-protocol-downgrade-retirement-set.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行1003。
- PristineDowngradeInventoryV1（pristine-downgrade-inventory.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行1005。
- PristineDowngradeAuthorizedStateV1（pristine-downgrade-authorized-state.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行1007。
- PristineDowngradeConsumeReceiptV1（pristine-downgrade-consume-receipt.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行1009。
- AuthorityProtocolMigrationLatchV1（authority-protocol-migration-latch.v1）：P：精确角色/独立证据资格或不变preimage消费链未同时闭合；源行537。

## 实施计划为何仍未就绪

C2规定有未决行就不发布完整元数据注册表。故本轮没有把93项拼成部分生产表，
没有创建产品schema-registry.v1.json或把未决塞入测试placeholder，也没有声称
完整角色/证据map已冻结。CanonicalBodyDigest/验签/bundle引擎、provider和生产
readiness仍未实现于这个切片；历史验证/物理/生产/main门槛保持。

下一步建议先把这份完整审查附表作为共同输入：不再逐条另起审批循环；对例外
集中决定一次后再冻结完整元数据计划。如果优先需要可交付代码，可明确批准将
首切片改为**146项精确schema/object名称目录＋独立名称/摘要framing向量**，把角色/
EvidenceKinds整组后置；这会改变已批准设计，不能由执行者自行选择。它不降低
运行时安全门槛，因为该名称目录不提供任何授权/验签/证据资格API。

本轮没有新契约提案自动生效；实施方法和测试范围也未获批准。
必要文档保存到现有开发分支；Git不携带本机SDK/cache/binary/provider状态。
