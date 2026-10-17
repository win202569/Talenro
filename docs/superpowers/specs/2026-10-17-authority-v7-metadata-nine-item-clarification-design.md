# Task11：九项元数据统一澄清附录

状态：用户在4fa0e16b后确认五项候选裁决及四项请求“不作独立证据”的建议，
并在8a8fc234后确认本书面附录；**九项本地设计已批准，未授权实施或部署**。
基线：4fa0e16b098424d1d219c0d9e1e989ca04cb99a6。
本次批准只涉及这九项的书面设计，不是其余44项、完整146项或canonical集合批准。

## 1. 目的、来源和范围

为后续元数据对照明确九个schema的角色与独立证据形式，区分有来源的规范解释
和新增的本地用途限制。沿用已本地批准的[C1请求认证分离](2026-10-16-authority-v7-request-auth-clarification.md)
与[C2准入方法](2026-10-16-authority-v7-metadata-admission-clarification.md)，
以[53项集中复核](../../roadmap/2026-10-16-task11-metadata-consolidated-reassessment.md)
及[原146项初审](../../roadmap/2026-10-16-task11-registry-metadata-audit.md)为审查输入。
下文L编号指[原v7规格](2026-08-24-nodecontrol-authority-v7-upgrade-design.md)在本基线的行号。

本附录不新增角色、wire字段、证据形式、认证渠道、运行时接口或生产配置。
不修改已交付的名称目录，不实现codec、验签、bundle verifier、provider或readiness。
不改变body digest的talenro.c12.前缀、signature metadata域或历史schema排除规则。

采用逐项、有来源的局部澄清，不采用所有Request共用新角色或unknown默认值。
推迟全部元数据而先做codec仍是另一种范围选择，不是本附录的实施授权。

## 2. 九项精确元数据

表中[]是显式空集合；非空集合逐字只含所列一种形式。null只表示该schema不采用
通用外部签名envelope，不表示请求免认证、调用方可信或可跳过本地可信读取。
本表是设计字面量，不是可以供产品导入的部分注册表。

| 原编号 / object | schema | SignerRoleOrNull | EvidenceKinds | 裁决性质 |
| --- | --- | --- | --- | --- |
| 40 AuthorityEpochTransitionCancellationV1 | authority-epoch-transition-cancellation.v1 | null | [database_immutable_body] | 原规格解释 |
| 44 AuthorityEpochTransitionRecoveryIntentV1 | authority-epoch-transition-recovery-intent.v1 | null | [database_immutable_body] | 原规格解释 |
| 47 AuthorityEpochTransitionRecoveryPrefixDecisionV1 | authority-epoch-transition-recovery-prefix-decision.v1 | null | [database_immutable_body] | 原规格解释 |
| 60 FenceDatabaseCommitAttestationArchiveSemanticIDJournalRecordV1 | database-commit-archive-fence-semantic-id-journal-record.v1 | null | [] | 原规格特殊本地信任路径解释 |
| 86 ClaimV1ReservationV1 | claim-v1-reservation.v1 | claim_v1_provider | [external_signed_envelope] | 原规格解释 |
| 100 RetireLegacySourceMembershipRequestV1 | retire-legacy-source-membership-request.v1 | legacy_source_retirement_admin | [] | 已有角色＋新增明确独立用途限制 |
| 104 RetireLegacyEnvironmentRequestV1 | retire-legacy-environment-request.v1 | legacy_source_retirement_admin | [] | 已有角色＋新增明确独立用途限制 |
| 106 RetireLegacyProviderNamespaceRequestV1 | retire-legacy-provider-namespace-request.v1 | legacy_source_retirement_admin | [] | 已有角色＋新增明确独立用途限制 |
| 135 RetireEnvironmentInventoryMembershipRequestV1 | retire-environment-inventory-membership-request.v1 | provider_downgrade_retirement_admin | [] | 已有角色＋新增明确独立用途限制 |

角色和EvidenceKinds是独立维度。所有正值只代表形式资格，不是验签成功、DB提交
证明或message admission成功；message-specific exact-cover、身份、时效、链、
trusted key/root/policy/algorithm及业务状态检查保持独立且mandatory。

## 3. 五项解释的来源与限定

1. **Cancellation（#40）**：L444要求与runtime result、terminal application同一
   transaction提交，L537明确append-only，L734/740规定取消链的preimage消费。
   DB body只在完整cancelled分支中有意义；它不能替代provider的signed rebind
   tombstone，也不能证明同transaction的另外两row存在。
2. **Recovery intent（#44）**：L452规定append-only、禁止UPDATE/DELETE及插入前
   Head绑定，L454、604/610明确提交后的row-bound proof和exact Recover consumer。
   目录登记不证明提交，旧proof也不能绕过当前holder/candidate/Head检查。
3. **Prefix decision（#47）**：L476/480规定apply和replacement_rebind互斥及
   不可变仲裁。L606/608、740分别约束apply的two-key closure与replacement proof。
   apply不能独立取得commit challenge；其body只能按相应application的related-row
   closure消费。replacement才使用独立prefix-decision row-kind通道。schema上的
   DB形式资格不把两种decision_kind变成同一种admission能力。
4. **Semantic-ID journal（#60）**：L478、639/641要求签发方自己的rollback-resistant
   journal、authenticated reservation、digest唯一索引、subject-slot CAS、反向唯一
   约束、seal及rollback floor。它不是普通DB row或外部signed evidence；caller
   提交相同bytes、record ID或signature都不能建立信任。[]不删除该特殊本地读取
   通道，也不新增第三种CanonicalEvidenceBundleV1 evidence_kind。
5. **Reservation（#86）**：L716明确provider envelope，L722固定Reserve响应角色，
   L732/734明确bundle携带exact Reservation envelope。DB存储这份envelope不能使
   其变为裸DB body；其context、sequence、lease baseline及连续prefix仍须重验。

## 4. 四项请求的明确用途限制

请求角色来源分别是L335/337、343、345和991；不借用其response角色，不变更
请求body、nonce、签名域、admin policy或原有认证义务。本附录新增的限定仅为：
这四个schema不能作为可复用的独立bundle evidence项，EvidenceKinds显式为[]。
这不是声称原规格已逐项写出这一禁止规则。

即便请求签名在对应可信admin policy下有效，也必须拒绝将该请求作为
external_signed_envelope或database_immutable_body独立项提交给bundle verifier。
把请求永久保存到DB、provider archive或审计日志，不改变这一用途限制。
该拒绝不能关闭正常请求认证、合法调用或exact retry，也不能把request本身
替代为已完成的membership/environment/provider retirement response或tombstone。

**保留原文展开通道**：L639涉及的provider request原文只能由同组signed provider
response的request_digest及immutable Inspect共同认证。只有对应message契约
明确要求的request原文才能在该契约规定的archive/展开位置被验证；不新增或改写
承载格式，不把它登记成独立bundle项，不允许caller-only bytes建立信任。
[]不禁止按既有契约保存、Inspect或展开request，也不增补其他message的传输权限。
若后续发现该通道的既有承载格式无法满足约束，必须保留为兼容阻塞，不能临时
塞进一种新的evidence_kind或泛化为“所有签过名的请求均可作证据”。

本组不改response的证据资格，不新增实际key、trust root、policy version，
不复用C1的security-admin请求角色到其他schema。部署兼容性仍为UNKNOWN。

## 5. 原审查、剩余44项与停止规则

原审查93/53与复核时的5+48都是各自时点记录，不回写为已批准全表。
本次方向确认覆盖5项解释＋4项用途限制，共9项；其余44项未获裁决。
本书面附录现已获本地批准，可以在新的本地对照表中记为93+9=102项局部分类、
44项未决；这是批准时点的记录，不是102项独立认证，也不让canonical集合或
生产策略自动生效。随后按既有C2方法得到的新来源复核结论须另记，不改本次批准范围。

44项精确沿用集中复核的R组27项和P组17项（以原初审编号定位）：

- R：12、20、34、38、42、45、52、57、61、63、66、68、69、71、74、78、80、83、
  85、87、88、120、123、126、129、131、139。
- P：29、30、32、49、89、90、91、93、108、113、115、132、141、142、143、144、145。

R组须逐行明确请求认证来源、是否采用通用封装、精确角色和独立证据资格；
P组须逐行明确保存/读取/传输与消费路径。Receipt已有角色不等于独立evidence资格，
临时或可被Down消费的row也不自动变成immutable证据。不能用null/[]遮掩这些问题。

全146项闭合前，不发布部分生产metadata registry，不创建占位descriptor/fixture，
不把unknown变为运行时fallback，也不为完整元数据交付编写“实施就绪”计划。
下一步仍是集中澄清这44项，不是重跑已完成名称目录验证或进入main合并。

## 6. 后续验证要求与本轮证据边界

以下是未来相应切片的验收要求，本轮没有执行：

- 独立字面量逐字覆盖九项schema/object/role/kind；错误角色、DB/EXT互换及独立
  请求证据一律拒绝。不得用产品查询生成自己的expected，计数相等也不能替代集合核对。
- #40的partial三row group、#44的未提交/错误Head proof、#47的apply独立challenge
  或错误related decision、#60的caller-only journal、#86的裸DB body分别为反例。
- 四个请求在正常认证路径与独立evidence路径分别验证：合法调用不因[]被禁用，
  合法请求签名不因验签成功而取得证据资格；archive展开仍须response＋Inspect绑定。
- 目录变异测试不声称已实现认证、验签或mutation0。真实consumer的副作用和
  可信环境验证须在其各自计划中另行冻结，不能用名称目录测试代替。

本轮只验证文档九行身份、原53项分割9+44、来源链接、范围与审批用语及Git差异。
没有产品代码、fixture、build/test、native诊断、Docker或DB运行，没有独立复审。
本地设计的确认不更改canonical approved spec-set、摘要、provider兼容性或部署状态。
Task11、物理/PITR、C12安全、生产及main门槛均保持开放。

## 7. 书面审阅与恢复

本附录书面审阅已获确认；继续44项分组契约澄清，不直接进入代码。
最新逐项状态见[更新对照与剩余契约分组](../../roadmap/2026-10-17-task11-metadata-current-reconciliation.md)。
只有完整元数据规则就绪后，才写对应实施计划并另行选择执行方式与验证预算。
不重放已消耗的leaf/name-directory/null-fix阶段，不依赖旧电脑的ignored日志或缓存。
必要文档保存到codex/c12-b01-task9-coordinator；新电脑从最新开发分支提交和
[恢复记录](../../runbooks/repository-recovery.md)继续，不能假定main包含这些进度。
