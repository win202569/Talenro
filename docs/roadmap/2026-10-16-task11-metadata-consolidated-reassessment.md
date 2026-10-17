# Task11：53项元数据例外集中复核

后续（2026-10-17）：用户在4fa0e16b后确认五项候选及四项E用途限制的方向。
见[九项统一书面附录](../superpowers/specs/2026-10-17-authority-v7-metadata-nine-item-clarification-design.md)，
该附录随后在8a8fc234后获书面批准；不等于其余44项已裁决。
最新[更新对照](2026-10-17-task11-metadata-current-reconciliation.md)另有#93来源复核，
当前103/43；以下5+48保留为复核时点记录，不是当前待审批状态。

状态：来源复核与候选裁决，待用户确认；不是获批的新契约、产品注册表或实施计划。
复核基线：bdebb8461ee913e2608d92b4f8378a624ebb0237。
用户要求按建议继续：先集中核对53项，再确认规则，最后另行规划实施。
本轮仅文档；没有改产品、fixture、测试、canonical规格集、信任策略或密钥，
没有build、测试、Docker、DB或既有诊断阶段重跑。

## 1. 目的与结论

在已验证的146项名称目录之上，厘清“允许哪个通用envelope角色”与“允许哪种
独立bundle evidence”两个维度；请求认证仍是第三个独立维度。
依据是已本地批准的[C1](../superpowers/specs/2026-10-16-authority-v7-request-auth-clarification.md)、
[C2](../superpowers/specs/2026-10-16-authority-v7-metadata-admission-clarification.md)、
[原146项初审](2026-10-16-task11-registry-metadata-audit.md)和
[原v7规格](../superpowers/specs/2026-08-24-nodecontrol-authority-v7-upgrade-design.md)。
下文L编号以该原规格在本基线的行号为准；对象/schema仍以原初审相同编号定位。

53项逐行重分为27项R（请求认证/封装未闭合）、4项E（请求角色已有明确来源，
独立证据资格未闭合）和22项P（其他对象）。这不是53个已证明的协议缺陷。
本轮从P组提出5项有具体证据链的候选裁决，剩余48项集中列出具体缺口。
原初审93/53保留为历史事实，不静默改成已认证98/48；即使用户接受这5项，
其余48项仍阻塞完整元数据注册表，且不构成独立审查或部署批准。

## 2. 五项候选裁决

以下均是执行者的来源解释，未写入产品表；请作为一组确认。
DB表示database_immutable_body，EXT表示external_signed_envelope。
null明确仅表示不采用通用外部签名封装，不表示免认证或任意来源可信。

| 原编号 / object | 候选SignerRoleOrNull | 候选EvidenceKinds | 肯定依据与不能省略的限制 |
| --- | --- | --- | --- |
| 40 AuthorityEpochTransitionCancellationV1 | null | DB | L444同一transaction提交cancellation/result/terminal三row，L537明确append-only，L734/740明确消费cancellation preimage；只可按对应cancelled分支的完整链使用，不是provider签名tombstone的替代品。 |
| 44 AuthorityEpochTransitionRecoveryIntentV1 | null | DB | L452明确append-only、禁止UPDATE/DELETE并绑定插入前Head；L454及604/610明确row-bound commit proof与exact Recover consumer；不因row存在就证明外部提交已获认可。 |
| 47 AuthorityEpochTransitionRecoveryPrefixDecisionV1 | null | DB | L476/480明确append-only与apply/replacement互斥；L606/608、740明确two-key apply closure与replacement proof消费。apply不得独立申请commit challenge，replacement才有独立row-kind通道；schema资格不抹平decision_kind或message-specific限制。 |
| 60 FenceDatabaseCommitAttestationArchiveSemanticIDJournalRecordV1 | null | [] | L639/641、478明确签发方自身rollback-resistant journal与authenticated reservation、digest唯一索引及subject-slot CAS；不接受caller提供record或signature。它有内部可信读取通道，不是普通bundle DB/外部签名证据；相同bytes或lookup hint不能建立信任。 |
| 86 ClaimV1ReservationV1 | claim_v1_provider | EXT | L716明确provider envelope，L722明确Reserve响应角色，L732/734明确bundle携带exact Reservation envelope；角色不是从附近lease段落猜测，DB保存也不得降为DB形式。 |

尤其不把#60的[]解释为删除journal验证；它恰恰要求使用规格规定的特殊本地信任
路径。也不把#47的DB解释为每一种decision都能独立提交任意消费者。

## 3. 剩余48项：一份完整问题清单

表内保留原编号，精确schema见原初审同一行。以下各项仍未决，不为它们填写生产
null/[]，不继承响应/父对象角色，不新增默认请求角色。

### 3.1 R：27项请求

共同缺口：调用中的challenge、credential、proof、capability或受信channel，不自动
等于该请求采用§4.1通用envelope。原规格所列响应角色也不是请求角色。
每一行须明确其请求认证来源、是否采用通用封装、采用时精确角色，以及独立证据
资格。下列来源是审查入口，不声称已完整冻结认证wire协议。

| 原编号 / object | 来源入口 | 必须保留/补明的边界 |
| --- | --- | --- |
| 12 AcquireDatabaseIncarnationRegistrationChallengeRequestV1 | L355 | 首次registration前的调用方身份与请求封装，不能从provider challenge签名反推。 |
| 20 RegisterDatabaseIncarnationRequestV1 | L367/369 | challenge与三份attestor envelope是bundle；请求自身身份/封装须另定，production ID仍来自intent。 |
| 34 RequestAuthorityEpochTransitionV1 | L432/434 | DB intent、fresh attestation/proof、current lease不自动赋予请求通用角色。 |
| 38 ResolveAuthorityEpochTransitionRequestV1 | L438 | resolution commit attestation与pending CAS不代替请求认证契约。 |
| 42 InspectAuthorityEpochTransitionRequestV1 | L448 | 只读不等于免认证；caller_nonce防响应重放不等于请求签名。 |
| 45 RecoverAuthorityEpochTransitionRequestV1 | L454/610/683 | intent commit proof、exact-one ordinary/gap与transcript ID不定义请求角色。 |
| 52 PrepareClaimV1GenesisRequestV1 | L496 | closed precursor与exact DB attempt/runtime证据保留；不继承preparation响应角色。 |
| 57 AcquireDatabaseCommitChallengeRequestV1 | L604/608 | row-kind、purpose、candidate及archive generation必须精确；请求封装与这些body检查分开。 |
| 61 FenceDatabaseCommitAttestationArchiveRequestV1 | L637/639/641 | reason-specific授权、trusted journal与precondition Inspect不能被caller-only bytes替代。 |
| 63 InspectDatabaseCommitAttestationArchiveRequestV1 | L659 | candidate相互认证与fresh trusted parent保持；明确谁可请求及是否另有envelope。 |
| 66 CompleteClaimV1GenesisRequestV1 | L675 | commit attestation和current proof不代表请求可用provider响应角色。 |
| 68 PrepareClaimV1GenesisReleaseRequestV1 | L677 | “同一signer role”修饰response；不能把它扩大到request。 |
| 69 OpenClaimV1GenesisRequestV1 | L679 | open nonce、release proof、mode矩阵与调用方权限分离。 |
| 71 InspectClaimV1GenesisRequestV1 | L681/683 | immutable stored request/bundle/response导出仍需明确请求访问边界。 |
| 74 AcquireDatabaseIncarnationProofRequestV1 | L709/711 | verifier单次challenge、purpose及holder proof不自动定义请求envelope角色。 |
| 78 RenewClaimV1ServingLeaseRequestV1 | L730/734/738/740 | attestor记账与三种renewal矩阵保留；请求身份、外部封装、evidence用途分别裁决。 |
| 80 RevokeClaimV1ServingLeaseRequestV1 | L744 | attestor在调用前保存ID/digest/nonce的要求保留；须明确请求认证及通用封装角色，不能从revocation响应或C1 policy请求类推。 |
| 83 RebindClaimV1RuntimeInstanceRequestV1 | L746/752 | authorization、termination、lineage与candidate-bound evidence均须独立验证；operator authorization签名不自动等于request envelope角色。 |
| 85 ClaimV1ReserveRequestV1 | L713/715/722 | ordinary context与credential admission不是provider响应签名角色。 |
| 87 ClaimV1FinalizeRequestV1 | L713/717/722 | reservation/effect/DB point和terminalization权限不能靠目录查找授予。 |
| 88 ClaimV1AbortRequestV1 | L713/718/722 | abort封闭枚举及reservation绑定保留，不继承Receipt角色。 |
| 120 RecoverStagingImportCapabilityRequestV1 | L851/853 | archived acquire与intent proof、held_recovery状态不替代请求认证定义。 |
| 123 AcquireFreshV7StagingExclusionRequestV1 | L841 | capability/commit proof/single-use challenge与请求认证分别冻结。 |
| 126 AcquireStagingDatabaseCommitChallengeRequestV1 | L871/873 | candidate preimage不是已提交证明；request封装不能从response推导。 |
| 129 ReleaseFreshV7StagingExclusionRequestV1 | L877/879 | held-only、no recovery与terminal互斥不能代替调用方权限。 |
| 131 AbortFreshV7StagingExclusionRequestV1 | L879 | held/held_recovery_abort_only与exact revocation proof保持；不按Release角色推断请求。 |
| 139 RetireProviderProtocolForDowngradeRequestV1 | L993/995/1001 | 独立admin authorization有角色，但不是本request已有同角色的证明。 |

### 3.2 E：4项已知请求角色

| 原编号 / object | 已有角色 / 来源 | 未决问题与推荐待批规则 |
| --- | --- | --- |
| 100 RetireLegacySourceMembershipRequestV1 | legacy_source_retirement_admin；L335/337 | 推荐独立EvidenceKinds=[]，请求认证及后续response request_digest绑定仍mandatory。 |
| 104 RetireLegacyEnvironmentRequestV1 | legacy_source_retirement_admin；L343 | 同上；不得把请求签名当作永久环境retirement response。 |
| 106 RetireLegacyProviderNamespaceRequestV1 | legacy_source_retirement_admin；L345 | 同上；不能把请求当作provider永久tombstone。 |
| 135 RetireEnvironmentInventoryMembershipRequestV1 | provider_downgrade_retirement_admin；L991 | 同上；不能把请求当作release inventory authority的membership tombstone。 |

这四项[]是建议新增的明确用途限制，不冒充原规格已经逐项禁止。L639的特殊
Fence授权展开需要provider request原文时，仍通过同组signed response及immutable
Inspect认证，不能因此把所有request升级为普通独立bundle evidence。若接受本组
用途限制，须在后续书面澄清中保留这种message-specific展开通道。

### 3.3 P：17项其他对象

| 原编号 / object | 来源入口 | 尚缺的精确裁决 |
| --- | --- | --- |
| 29 DatabaseLegacyAbsenceProjectionV1 | L409/411/492/599 | 稳定可重算投影与attempt引用已明确；跨边界preimage究竟经固定immutable保存/独立DB项还是嵌入传递，须明示，不能仅凭投影名称判[]。 |
| 30 DatabaseLegacyAbsenceObservationV1 | L413/490/492/599 | attempt与activation阶段不同；须明示每份observation的不可变保存位置、独立消费或仅父对象绑定。 |
| 32 AuthorityGenesisEpochEvidenceV1 | L421/428/728/825/918 | immutable epoch evidence与lease/Head一致性明确；独立证据的保存/传输形式尚须闭合，不从Evidence后缀赋权。 |
| 49 EmptyDatabaseInventoryV1 | L490/492/496 | attempt-time object且有digest引用；须明确完整preimage保存和Prepare消费通道，不将digest引用当DB immutable body证明。 |
| 89 ClaimV1ReceiptV1 | L719/720/722 | provider envelope与claim_v1_provider角色可由Finalize/Abort契约确认；但Record嵌入完整Receipt不自动授予Receipt独立bundle项资格，仍须明确独立消费位置，不能仅凭role就登记EXT。 |
| 90 ClaimV1RecordV1 | L720/722/732/734 | Record envelope独立消费有据，但须把它与Inspect provider response的精确object/角色对应明确；不能仅因包含合法Receipt就继承其角色。 |
| 91 IndeterminateSourceSealV1 | L537/782/784 | 本地永久freeze与同transaction提交已明示；缺独立跨边界证据消费资格，不能因为DB row就自动登记DB形式。 |
| 93 LegacySourceInventoryV1 | L795/797 | database_fence唯一evidence schema已有据；还须明确stage-specific稳定内容与完整observation body如何immutable保存/可信读取，不将stage/time不同的body混同。 |
| 108 LegacySourceRetirementSetV1 | L347/428/786 | mechanically derived exact-cover set已有据；须明确集合本身的传输及独立资格，不能继承成员签名。 |
| 113 FreshRestoreRequirementV1 | L537/784/817 | source seal同transaction、不可变及freeze已有据；本地守卫用途不自动产生独立bundle资格。 |
| 115 FreshImportTopologyProjectionV1 | L831/837 | 纯投影且明确排除tx/time/control/evidence；须明确是受信重算或传输对象，不把expected-post摘要当DB提交证明。 |
| 132 FreshV7StagingExclusionAbortV1 | L877/879 | 与Release并列的terminal response明确；Abort精确role需直接绑定，不能仅因前一段Release有provider role就继承。 |
| 141 ProviderProtocolDowngradeRetirementSetV1 | L1003/1005 | set由B01验证后经context传给同一transaction；bundle具体携带members的envelopes，集合自身独立资格须明确。 |
| 142 PristineDowngradeInventoryV1 | L1005/1007 | transaction-local pre-authorization inventory，有stored preimage检查；不是已提交不可变DB证据的自动证明。 |
| 143 PristineDowngradeAuthorizedStateV1 | L1007 | 同transaction授权插入后的阶段投影，不能继承B11 authorization角色或伪称已提交row。 |
| 144 PristineDowngradeConsumeReceiptV1 | L1009 | guard consume后latch/auth已删除，整体transaction还能回滚；Receipt名称不代表外部签名或committed DB evidence。 |
| 145 AuthorityProtocolMigrationLatchV1 | L537/1005/1009 | Up插入但Down原子consume；不具普遍不可删的immutable row语义，须明确DB-local guard preimage与独立证据资格分离。 |

## 4. 集中决策建议与替代方案

推荐：按有据的五项先确认局部解释；四项E的[]限制作为一组新的书面澄清提案；
27项R先冻结请求认证/封装矩阵，17项P先冻结保存/消费路径。全部完成前不出版
完整元数据表，不以部分descriptor或unknown运行时枚举绕过阻塞。

替代一：现在统一给所有请求新增一个签名角色。虽然表面统一，但不同admin、
普通caller、attestor channel、provider recovery及只读Inspect的权限边界不同，
会引入新的wire/trust/bootstrap契约，本轮没有来源或权限替用户选择它。

替代二：再后置元数据、优先body codec。可以另作切片，但不能解决本轮53项，
也不能让通用验签器在role未定时开始接线；需要另行范围决定。

本轮不选择新的请求认证协议，不写实现计划。下一步是用户确认上述分组方向后，
写一份明确区分原规范解释与新增契约的统一附录；其书面审阅通过后才进入实施规划。
不逐个对象另起审批循环，也不把一次方向批准扩大为48项内容已经认证。

## 5. 恢复与核对

新电脑从现有开发分支的最新文档检查点开始。已完成的名称目录实现、null修复及
受限验证不重跑。原初审、原规格与本报告都保留，canonical approved spec-set
及摘要不变；兼容性仍UNKNOWN。完整Task11/C12、物理环境、生产与main门槛开放。

本报告只做文档集合/引用/差异核对，不宣称完成验签、codec、真实provider或测试。
提交前须确认53个原编号一对一覆盖、5+48分类精确、引用对象拼写正确、无占位行，
并确认只有本报告及进度/恢复文档变更。未经接受的候选不能写成已批准策略。
