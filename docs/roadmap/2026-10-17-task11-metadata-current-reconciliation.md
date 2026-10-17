# Task11：更新元数据对照与剩余契约分组

基线：8a8fc234a1a8db17193dbff843aeaab3f2220870；用户随后确认
[九项书面澄清](../superpowers/specs/2026-10-17-authority-v7-metadata-nine-item-clarification-design.md)。
本轮只做本地设计记账与来源复核；没有代码、fixture、build/test、Docker、DB、
provider操作或旧诊断阶段重跑，没有新的独立审查。

## 1. 时点、计数与权限

| 时点 | 有局部依据的分类 | 未决 | 含义 |
| --- | ---: | ---: | --- |
| 原初审 | 93 | 53 | 执行者初审，非独立认证；原表保留。 |
| 九项书面附录获批 | 102 | 44 | 93项既有分类＋9项获批局部澄清；只批准这9项。 |
| 本轮补充来源复核 | 103 | 43 | #93取得完整保存/消费依据，新增1项执行者来源判断；不是新增契约或用户批准其余44项。 |

146个精确schema/object名称不变。103不是实现数量、独立安全认证数量或生产
可接受证据数量；所有运行时接线、canonical登记、兼容性与整体验收门槛仍开放。
既有93项沿用原初审时的局部判断，不借本次确认宣称逐项重新认证。

## 2. 可恢复的完整对照关系

这是一份文档审查增量，不是部分产品注册表。完整146项身份、原角色/形式/来源
仍在[原初审表](2026-10-16-task11-registry-metadata-audit.md)。以该表为审查基表，
仅将下列10项替换为本节明确结论，其他136项原样继承；未决项不填null/[]。
本节来源全部随Git保存，不要求旧电脑的日志、缓存或SDK，也不被产品运行时加载。

| 原编号 / object | schema | SignerRoleOrNull | EvidenceKinds | 来源与状态 |
| --- | --- | --- | --- | --- |
| 40 AuthorityEpochTransitionCancellationV1 | authority-epoch-transition-cancellation.v1 | null | [database_immutable_body] | 已批九项附录§2–3；完整取消三row链限制不变。 |
| 44 AuthorityEpochTransitionRecoveryIntentV1 | authority-epoch-transition-recovery-intent.v1 | null | [database_immutable_body] | 已批九项附录§2–3；commit proof与exact consumer仍必需。 |
| 47 AuthorityEpochTransitionRecoveryPrefixDecisionV1 | authority-epoch-transition-recovery-prefix-decision.v1 | null | [database_immutable_body] | 已批九项附录§2–3；apply不能独立取得challenge。 |
| 60 FenceDatabaseCommitAttestationArchiveSemanticIDJournalRecordV1 | database-commit-archive-fence-semantic-id-journal-record.v1 | null | [] | 已批九项附录§2–3；保留签发方可信本地journal读取。 |
| 86 ClaimV1ReservationV1 | claim-v1-reservation.v1 | claim_v1_provider | [external_signed_envelope] | 已批九项附录§2–3；不降为裸DB body。 |
| 93 LegacySourceInventoryV1 | legacy-source-inventory.v1 | null | [database_immutable_body] | 本轮执行者来源复核：原v7 L795/797/803/809/811，详见§3，非独立认证。 |
| 100 RetireLegacySourceMembershipRequestV1 | retire-legacy-source-membership-request.v1 | legacy_source_retirement_admin | [] | 已批九项附录§2/4；认证及受约束的archive展开保留。 |
| 104 RetireLegacyEnvironmentRequestV1 | retire-legacy-environment-request.v1 | legacy_source_retirement_admin | [] | 已批九项附录§2/4；不是environment tombstone。 |
| 106 RetireLegacyProviderNamespaceRequestV1 | retire-legacy-provider-namespace-request.v1 | legacy_source_retirement_admin | [] | 已批九项附录§2/4；不是provider tombstone。 |
| 135 RetireEnvironmentInventoryMembershipRequestV1 | retire-environment-inventory-membership-request.v1 | provider_downgrade_retirement_admin | [] | 已批九项附录§2/4；不是membership response。 |

## 3. #93的新依据：不是只看名称或存储位置

来源：[原v7规格](../superpowers/specs/2026-08-24-nodecontrol-authority-v7-upgrade-design.md)
在本基线的L795定义retirement/seal/post_seal阶段的完整body；L797将database_fence
的唯一evidence schema固定为legacy-source-inventory.v1；L803要求对应observation
的CanonicalEvidenceBundleV1携带exact preimage/envelope。尤其L809要求archive
attestor通过绑定source DB identity的可信只读连接，自行读取唯一immutable seal
及inventory row，确认另一个transaction已提交可见并重算digest；L811再trusted-read
post-seal inventory并检查stable content相等。这补齐原初审只引用L795时遗漏的
保存、可信读取和独立消费链。按C2，局部分类为null＋database_immutable_body。

该分类不新增外部签名角色，不让inventory继承archive attestor或epoch auditor角色。
retirement/seal/post_seal的body包含不同stage/snapshot/time/point，即使stable content
相同也不能交换完整body或digest。消费者仍须绑定exact source DB、membership、
retirement、stage、observation及对应seal/post-seal链；裸caller bytes、只有stable
content摘要或无可信提交来源均不能替代这些检查。原L809所述WAL point不升级为
§6.1的通用exact COMMIT witness，本轮不改变物理提交证明规则。

这是按已有规则作的执行者局部来源判断，不是新批准的契约，不减少独立复审要求。

## 4. 剩余43项的精确集合

R组仍27项：

- R：12、20、34、38、42、45、52、57、61、63、66、68、69、71、74、78、80、83、
  85、87、88、120、123、126、129、131、139。

P组从17减为16项，仅移除本轮有新来源的#93：

- P：29、30、32、49、89、90、91、108、113、115、132、141、142、143、144、145。

每项object/schema与[集中复核](2026-10-16-task11-metadata-consolidated-reassessment.md)
同编号一致。这里的R/P只是文档问题分组，不进入产品enum或fallback。

## 5. R组：七个责任组，不是七个已批准角色

原v7 L325/327已有credential隔离与accepted key set的要求，L278/639要求保存和
认证exact request/archive；这些事实不能单独推出每个request的通用envelope角色。
基础设计§9.1规定provider独立性、fence及恢复边界，也不能替代v7请求的wire认证
定义。此次阅读没有发现足以统一填满27行的明确schema→request-role映射。

以下按接收方和业务责任集中提出问题，**不默认生成七个角色，不分配生产key或policy**。

| 责任组 | 原编号 | 来源入口与应冻结的问题 |
| --- | --- | --- |
| registration/genesis | 12,20,52,66,68,69 | v7 L355/367/496/675/677/679及L325/327；明确首次注册与后续genesis分别由谁认证、是否复用既有调用通道；不能用response的claim_v1_provider作请求角色。 |
| epoch/runtime/lease | 34,38,45,78,80,83 | L434/438/454/730/744/746/752；分别保留intent/commit proof、attestor记账、rebind authorization及candidate限制；这些证明不是对request envelope角色的定义。 |
| ordinary mutation | 85,87,88 | L713–724；保留credential policy、current proof/context与terminalization规则；明确caller签名/传输身份如何绑定exact request，不能靠响应签名授权调用。 |
| read/proof | 42,71,74 | L448/681/709；Inspect与proof不授予mutation权限，仍需明确访问者身份、nonce/request绑定与封装资格。 |
| challenge/archive | 57,61,63,126 | L604/637/639/641/659/871/873；区分向provider申请challenge与向archive signer请求Fence/Inspect，保留可信journal和candidate双向认证，不能让caller选择信任来源。 |
| staging hold/recovery | 120,123,129,131 | L841/851/877/879；明确谁可提交capability/recovery/terminal请求，签名认证不能替代held状态、exact proof与single-use CAS。 |
| disposable downgrade | 139 | L993/995/1001；独立admin authorization已明确，但它不自动让retirement request使用同一个角色或免认证。 |

每行未来必须给出：精确object/schema、调用方/接收方、认证来源、请求body绑定、
通用envelope有无、采用时唯一role、独立evidence资格及兼容/拒绝边界。
同组不同调用不能因组名相同而继承权限；缺定义的字段仍是文档未决。

## 6. P组：16项具体数据通道问题

| 原编号 / object | 来源入口 | 待冻结的责任边界；本轮不擅自裁决 |
| --- | --- | --- |
| 29 DatabaseLegacyAbsenceProjectionV1 | L409/411/492/599 | stable重算对象的跨边界preimage如何保存/传输；不能把可重算直接等同已提交immutable row。 |
| 30 DatabaseLegacyAbsenceObservationV1 | L413/490/492/599 | attempt与activation observation分别绑定哪份保存的body、哪些consumer可独立使用。 |
| 32 AuthorityGenesisEpochEvidenceV1 | L421/428/728/825/918 | immutable epoch evidence的完整body保存与Prepare/lease消费形式，不能靠Evidence后缀赋权。 |
| 49 EmptyDatabaseInventoryV1 | L490/492/496 | attempt-time完整inventory如何成为可信preimage；不能仅凭attempt里有digest就登记DB形式。 |
| 89 ClaimV1ReceiptV1 | L719/720/722 | role已有依据；仍须明确独立bundle消费还是只在Record中嵌入，不能靠角色已知自动登记EXT。 |
| 90 ClaimV1RecordV1 | L720/722/732/734 | 独立Record envelope的消费已有依据；需明确exact object与Inspect response的角色对应，而非从Receipt继承。 |
| 91 IndeterminateSourceSealV1 | L537/782/784 | 本地freeze row有提交依据；需明确是否有独立跨边界消费，minimal seal本身不能授权archive或target。 |
| 108 LegacySourceRetirementSetV1 | L347/428/786 | mechanical exact-cover集合自己的保存/传输与证据形式，不继承成员签名。 |
| 113 FreshRestoreRequirementV1 | L537/784/817 | source seal同transaction的本地守卫row与独立evidence资格分离。 |
| 115 FreshImportTopologyProjectionV1 | L831/837/883 | 纯投影由谁可信重算/传递，不能把expected-post摘要当提交证明或继承inventory签名。 |
| 132 FreshV7StagingExclusionAbortV1 | L877/879 | terminal provider response的精确role须明示；不从邻段Release角色自动复制。 |
| 141 ProviderProtocolDowngradeRetirementSetV1 | L1003/1005 | B01验证的集合经context传入transaction；集合自身与成员envelope的传输/资格分离。 |
| 142 PristineDowngradeInventoryV1 | L1005/1007 | pre-authorization transaction-local inventory与已提交DB证据不是同一形式。 |
| 143 PristineDowngradeAuthorizedStateV1 | L1007 | 同transaction auth=1阶段投影，不能继承B11 authorization签名或假称独立committed row。 |
| 144 PristineDowngradeConsumeReceiptV1 | L1009 | guard consume仍可能整体回滚；明确同transaction内部传递而非凭Receipt名称赋签名能力。 |
| 145 AuthorityProtocolMigrationLatchV1 | L537/1005/1009 | Up写入、Down原子consume的latch如何被可信本地检查；不能登记成普遍不可删除的DB证据。 |

这16项尚无新增契约自动生效。尤其没有将“看起来只在本地用”默写为null/[]。

## 7. 下一步的架构选择（建议，尚未批准）

推荐**按契约分组的混合认证策略**：先写27项请求认证矩阵，已有明示专用通道的
保留其通道；确需通用signed request的，另行明示专用请求role与body/schema/
policy绑定，不复用响应角色。不根据组名自动授予权限，不把C1复制给其他请求。
每个schema在矩阵中最终只有明确的封装选择；认证与独立bundle资格仍分开，
任何新role、用途限制、wire或bootstrap要求均显式标为新增提案。P组数据通道
另外按consumer展开，不用认证方案冒充其存储/证据决定。

替代一是把27项全部改为通用signed request并统一规划角色，容易形成一致表面，
但会改动已有channel及bootstrap/兼容边界，当前没有依据或批准直接采用。
替代二是为27项全部填null、只依赖既有proof/传输身份，变更较少，但没有逐行
明确认证与body绑定就会把未知伪装成无通用封装，也不采用。

下一次需要确认的是上述混合方向，而不是逐对象重复审批。方向确认只授权写
完整矩阵和书面契约提案；提案本身仍需审阅，不能一并当作27项内容已被批准。
43项未决全部闭合前，不产生实施就绪的完整元数据计划或部分产品表。

## 8. 核对与恢复

本轮核对目标：原146项身份唯一；9项已批映射及#93增量无重号/错名；原53项
精确分割为9＋1＋27＋16；七个R责任组无缺漏/重复；原93项分类与原审查时点
记录保留。局部判断不是运行时验证成功，本轮没有执行产品测试或负载。

新电脑从最新开发分支及[恢复记录](../runbooks/repository-recovery.md)继续。
不要重放已完成的name-directory/null-fix/leaf阶段，不把canonical登记、兼容性、
真实环境/PITR、Task11/C12、生产或main门槛算作通过。
