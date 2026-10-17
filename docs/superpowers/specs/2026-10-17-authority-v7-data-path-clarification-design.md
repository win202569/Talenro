# Task11：16项数据通道统一澄清附录

状态：用户在0ec812cd后确认本书面附录全文；**本地设计已批准，不授权实现、迁移或部署**。
此前399f3d21后的确认只批准方向；本次书面确认保留5项DB原文、3项provider
envelope、8项本地可信通道的全部具体约束。
基线：399f3d211a7ef0a89962a2f42af41cca4aa3ff16。
最终[146项本地对照](../../roadmap/2026-10-17-task11-final-metadata-reconciliation.md)
为146/0；这是局部设计一致性结果，不是独立认证或产品验收。

## 1. 目标、依据和非目标

明确16个schema的角色/独立证据资格，以及相应原文保存、可信读取和消费责任。
来源是[原v7规格](2026-08-24-nodecontrol-authority-v7-upgrade-design.md)、
[C2准入方法](2026-10-16-authority-v7-metadata-admission-clarification.md)、
[已批九项澄清](2026-10-17-authority-v7-metadata-nine-item-clarification-design.md)、
[已批27项请求矩阵](2026-10-17-authority-v7-request-auth-matrix-design.md)及
[16项来源审查](../../roadmap/2026-10-17-task11-data-path-review.md)。
下文L编号以原v7规格在本基线的行号为准。

选择typed immutable companion原文＋既有provider envelope＋明确本地用途限制。
不新增projection signer、第三种bundle evidence、canonical object/schema或body字段，
不改talenro.c12.摘要前缀、signature metadata域或历史schema排除规则。
不能把全部对象填null/[]来隐藏跨边界原文需求，也不为集合默认继承成员签名。

本文提出新的持久化/owner关系、部分角色用途明示与独立资格限制，不声称原SQL、
manifest或consumer已经具备这些行为。物理存储与consumer后续单独规划；本轮
不创建table、SQL、fixture、codec、API、key/policy或产品注册表，不运行负载。

## 2. 16项精确metadata（书面已批准）

null表示不采用通用外部envelope，不表示无需可信来源；[]是明确无独立bundle
资格，不表示对象可删除或无需本地校验。非空集合逐字只含表中所列一种形式。
所有值只表达协议形式，不证明编码、验签、DB提交或业务admission已经成功。

| 原编号 / object | schema | SignerRoleOrNull | EvidenceKinds | 通道 |
| --- | --- | --- | --- | --- |
| 29 DatabaseLegacyAbsenceProjectionV1 | database-legacy-absence-projection.v1 | null | [database_immutable_body] | A |
| 30 DatabaseLegacyAbsenceObservationV1 | database-legacy-absence-observation.v1 | null | [database_immutable_body] | A |
| 32 AuthorityGenesisEpochEvidenceV1 | authority-genesis-epoch-evidence.v1 | null | [database_immutable_body] | A |
| 49 EmptyDatabaseInventoryV1 | empty-database-inventory.v1 | null | [database_immutable_body] | A |
| 89 ClaimV1ReceiptV1 | claim-v1-receipt.v1 | claim_v1_provider | [external_signed_envelope] | B |
| 90 ClaimV1RecordV1 | claim-v1-record.v1 | claim_v1_provider | [external_signed_envelope] | B |
| 91 IndeterminateSourceSealV1 | indeterminate-source-seal.v1 | null | [] | C |
| 108 LegacySourceRetirementSetV1 | legacy-source-retirement-set.v1 | null | [database_immutable_body] | A |
| 113 FreshRestoreRequirementV1 | fresh-restore-requirement.v1 | null | [] | C |
| 115 FreshImportTopologyProjectionV1 | fresh-import-topology-projection.v1 | null | [] | C |
| 132 FreshV7StagingExclusionAbortV1 | fresh-v7-staging-exclusion-abort.v1 | claim_v1_provider | [external_signed_envelope] | B |
| 141 ProviderProtocolDowngradeRetirementSetV1 | provider-protocol-downgrade-retirement-set.v1 | null | [] | C |
| 142 PristineDowngradeInventoryV1 | pristine-downgrade-inventory.v1 | null | [] | C |
| 143 PristineDowngradeAuthorizedStateV1 | pristine-downgrade-authorized-state.v1 | null | [] | C |
| 144 PristineDowngradeConsumeReceiptV1 | pristine-downgrade-consume-receipt.v1 | null | [] | C |
| 145 AuthorityProtocolMigrationLatchV1 | authority-protocol-migration-latch.v1 | null | [] | C |

A组5项的不可变保存与owner要求是新增契约，不是“有digest即已持久化”的解释。
B组#89独立形式资格、#90的ordinary Inspect object/role对应及#132的明确角色
在本文逐项固定；不能只引用相邻段落作为批准依据。C组的独立用途限制也须明确
批准，不把“原文未提及外部消费”说成原规范已禁止所有未来用途。

## 3. A组共同存储/信任规则

完整canonical body与其digest须作为**有owner的typed companion**持久保存。
owner kind/id、schema和slot由固定协议分支机械决定，不接受caller指定任意owner。
逻辑唯一键为：实际DB identity、activation ID、owner kind、owner ID、object schema、
object slot；slot只用于本节明确的projection/observation/epoch/inventory/retirement-set
用途。它们是数据库存储关系，不新增wire字段、签名domain或第147个canonical schema。

每个companion值保存exact canonical bytes和按原§4.1重算的body digest。owner
存在/唯一、对应body引用、activation/DB identity和slot/stage必须逐字匹配；即使
caller给出相同digest，也不能跳过owner关系或把另一个activation的body挂入此处。
一次原子提交owner与其规定companion，随后禁止UPDATE/DELETE/TRUNCATE和owner
重挂；same key/same bytes exact retry可复用，different bytes、partial owner组、
duplicate、orphan或不完整preimage固定关闭该路径，不允许后台补猜或清理伪装通过。

写入只允许原§4.1 B01 typed verifier和authoritative Go executor控制的受限路径，
不能授予普通runtime/Goose CLI/operator SQL。SQL须重验本地identity、ACL、locks、
owner/row/cardinality与只追加约束；数据库不声称执行JCS或公钥验证。所有外部
provider/attestor调用仍在DB锁/事务之外；事务只消费此前验证且仍满足原规则的证据。

读取者必须使用自身受信路径从exact DB/owner读取完整body并比较，不接受caller
自报“已经trusted-read”的标记。来源DB从受信inventory/registration/lineage确定，
不能从caller URI、可变CAS或未验连接串推导。已有source archive管理只读路径可以
按本附录受控读取publication，不得重开被关闭的普通route或复活被撤销credential；
没有合格管理读路径时该consumer保持关闭，而不是只验一个裸body digest继续。

跨边界只在原message明确要求的preimage集合中使用database_immutable_body，
依§4.1携带exact body；接收方仍需重算、验证owner来源与全部相关外部证据并执行
message-specific exact-cover。类型资格本身不证明commit，不授权任意message加项。

## 4. Attempt/activation四项的owner、事务与消费

以下owner均为既有canonical对象的真实DB行，不是从同digest临时造出的owner。
owner只引用已经构造的子body digest；companion存储边在body外，避免子对象再
引用owner digest而形成固定点。owner与子body必须在同一个受限transaction提交。
owner body没有直接DB identity字段时，从其已验证的registration/runtime/evidence
链机械取得并与实际连接核对，不能为方便比较向原owner body补字段或信任caller hint。

| Companion用途 | owner与固定引用 | 原子保存与消费责任 |
| --- | --- | --- |
| A29 projection | AuthorityProtocolUpgradeAttemptV1的attempt_id；database_legacy_absence_projection_digest | 按L409/411/492在锁定扫描中构造，随attempt提交；Prepare读取该owner原文。activation按L553重算stable projection并与之byte-identical比较，不能覆盖旧body。 |
| A30 attempt observation | 同一attempt；attempt_database_observation_digest | stage=attempt，保存原snapshot/point/time，随attempt提交；inventory及后续activation以这个exact观察为基线。 |
| A30 activation observation | AuthorityProtocolActivationV1的activation_id；activation_database_observation_digest | stage=activation，按L413/553与activation同transaction提交；新的snapshot/time可不同，projection须相同、database point不得早于attempt observation。 |
| A32 epoch evidence | 同一attempt；epoch_evidence_digest | 按L421/428在写attempt前完成mode-specific derivation，与attempt同transaction保存；Prepare及后续lease/Head沿用该body/root，不在activation时重新选择epoch。 |
| A49 empty inventory | 同一attempt；database_inventory_digest | 按L490/492构造attempt-time原文并同transaction保存；Prepare消费，activation的attempt_database_inventory_digest只引用这份body，不重造“相同MVCC”inventory。 |

activation还必须通过其attempt_digest指向上述唯一attempt，核对projection、
attempt inventory、epoch evidence及genesis root的原有引用。缺任一companion不能
用现场scan取代已保存历史；需要现场scan的约束则另外执行，不能反向改写旧body。
L553的activation新观察不等于旧观察，这是两个owner/slot，不是同一body的更新。

## 5. A108：先于epoch签发的source retirement-set publication

本节新增一个**仅存储层的retirement-set publication owner**，不新增canonical
object。owner在受信完整environment inventory派生的每个distinct source DB中各
有一份，唯一键为该source DB identity、activation ID和retirement_plan_id，永久
绑定唯一source_membership_retirement_digest；不能把membership digest加入可变
key来为同plan生成第二owner。禁止由caller任选一个发布库替代全部source。

owner的创建条件：永久membership tombstone已验证，完整inventory/anchors及
action plan exact，全部member DB/environment/provider retirement均已terminal，
所有签名、commit attestation与coverage约束满足L347/786。不用短时shutdown或
partial retirement作为发布资格，也不以尚未存在的epoch source-set/完整seal为前提。

B11在第一次publication之前，把机械构造的完整LegacySourceRetirementSetV1 body、
created_at和digest在既有rollback-resistant retirement-plan journal中一次固定。
这是新增journal保存责任；retry复用原文和时间，same plan/different body拒绝。
journal不是独立bundle证据，也不因为保存了原文就取代DB publication。

随后在每个source DB中，以受限source-retirement publication transaction原子插入
publication owner及其唯一A108 companion。owner永久绑定本地已存在的DB retirement
row、same global membership/plan/inventory以及fixed set digest，companion保存上述
完整原文；不等待source seal、不改变任何retirement tombstone或既有body字段。
各source副本的canonical body必须byte-identical，共享provider tombstone可复用，
但不能按source删减set成员。owner没有epoch/source-set/seal digest反向依赖。

不要求跨DB分布式transaction，也没有TTL或rollback-unretire。每个source独立
atomic append并exact retry；**全部source publication已提交且可信读取一致之前，
不得签发依赖该set的epoch observation/source-set/maximum或完整source seal**。
任何source不可访问、publication缺失/不同/partial、journal不可验证均保持全scope
closed，不自动换库、重建body、选择多数副本或从caller附件补齐。

依赖顺序固定为：

```text
permanent membership + complete member retirement evidence
  -> B11 fixed set body -> per-source publication owner + exact body (commit/read)
  -> epoch observations/source-set/maximum -> full source seal -> archive/target
```

这使L805的source-set先取得有owner的set preimage，L786/809的完整seal随后引用
同一digest，不再以未来seal作首次保存的owner。后续source archive reader核对
publication、seal和原始member evidence；publication不代替任何外部签名或DB提交证明。

## 6. 恢复、物理存储与commit-proof边界

A的DB形式不是新commit proof，五项均**不加入§6.1固定row-kind registry**，
不得冒充activation/resolution等已有kind申请attestation，不给它们增设通用签名role。
trusted读取已提交body只证明相应受信读取事实；server WAL/flush/observation point
不能冒充exact COMMIT/COMMIT PREPARED record end LSN。

DB中完整owner/companion组仍在时只可逐字复用；组缺失、partial或different时不能
从provider archive、journal或当前transaction猜回历史DB事实。现有已批恢复分支
只有在其明确覆盖的类型/owner、原文及必要fork-cut witness全部成立时才可使用；
本附录**不新增针对这些companion/publication的缺失行恢复权限**。没有对应已批
恢复能力就保持关闭，即使外部存有相同bytes或签名也不能自行回填或重新执行effect。
需要新增恢复能力时须另行明确契约，不在metadata目录里隐藏这个环境门槛。

未来物理实施须把companion/publication及owner完整纳入固定catalog、lock manifest、
ACL、guard、backup/恢复与非control protocol row检查。它们不是额外的Down control
table；原control_table_count=2及latch/auth三阶段不变。有orphan/partial或非空协议
记录时保持Down/readiness关闭，不能从fixed forbidden registry排除它们以凑零行。
source stable legacy-content投影对新增protocol存储的分类必须显式、稳定，不能让
publication本身破坏retirement→seal→post-seal的旧内容相等性，也不能借此漏扫领域表。

这些是新增物理存储责任，不声称现manifest已包含。部署前须确定获批schema版本、
迁移/兼容向量与guard覆盖；不得就地改写已发布/shared migration或旧canonical摘要。
本轮不选择或执行迁移命令，未来的metadata-only计划也不自动获得这些SQL变更权限。

## 7. B组：三个provider envelope的精确责任

三项共同取claim_v1_provider、EvidenceKinds=[external_signed_envelope]。这是逐项
固定的角色/形式，不从邻居或内嵌对象继承；不改变任意message的exact-cover清单。

- **B89 Receipt**：按L719/722由Finalize/Abort的provider生成并签名，保持完整
  reservation_envelope、terminalization context及terminal状态组合。L720的Record
  可以嵌入其完整envelope；这不使Record或caller获得Receipt签名能力。独立形式
  资格只在明确消费它的契约中使用，不能在其他bundle中添加“额外合法receipt”。
- **B90 Record**：按L720/722明确ordinary Inspect返回的ClaimV1RecordV1 envelope
  由provider签名，L732/734的prefix replay消费该exact envelope；它不是genesis
  InspectClaimV1GenesisResponseV1。保留reservation与terminal receipt的完整验证，
  不能因外层Record签名有效就信任任意内层签名/引用，也不继承caller当前context。
- **B132 Abort**：显式补明FreshV7StagingExclusionAbortV1为provider签名响应，
  role固定claim_v1_provider，而非靠L877 Release的角色词推断。L879的held/
  held_recovery_abort_only、proof、terminal CAS和Release互斥全部保持；请求签名
  或timeout不产生Abort响应，不让held_recovery_required跳过Recover。

DB存储这些envelope不改变EXT形式；替换成裸DB body拒绝。Record/Receipt历史
digest不因lease续租改变，签名可信时间/撤销及历史验证继续按原规格，而非按
caller当前context重签或把旧body修订为新holder的事实。不生成新key或root。

## 8. C组：八项本地可信通道

C组共同取null、EvidenceKinds=[]，须拒绝两种独立bundle包装。此为明确用途
限制，不删除本地对象、row或guard，不允许raw caller JSON伪装可信typed值。

- **C91 minimal seal**：可信source reader按L537/782/784从本DB读取唯一row及
  same-transaction requirement，重验classification/failed/missing/identity。后续
  完整seal按L809在本地核对引用。该对象只维持freeze，不授权archive/epoch/target。
- **C113 requirement**：按L817读取与对应source seal同transaction的immutable
  requirement，核对branch、nullable、reason和完整blocked capabilities；不把其
  digest或[]解释为可以恢复普通服务或绕过完整source archive。
- **C115 import projection**：按L831/837，由manifest signer和consumer各自从
  exact objects构造；重算pre/expected-post/actual-post，保持排除tx/time/control/
  evidence的规则。独立重算与digest比对是本地信任通道，不把其包装成提交证明。
- **C141 downgrade set**：B01从原L1003的完整inventory和全部retirement envelopes
  验证/构造set，B11为本次受控Down固定完整body及created_at并保存到其幂等执行
  journal；timestamp不是从成员签名机械推导。B01重算成员/coverage和body digest后，
  只经私有typed context交给L1005同一Goose callback；caller自造context、裸set
  或只有digest都拒绝。context绑定installation/DB/latch/inventory及本次执行身份，
  不能跨连接/transaction重用；成员envelopes的独立验证不因set本地可信而省略。
- **C142 pre-auth inventory**：L1005的同一transaction/完整locks下从fixed manifest
  构造并保存该transaction的pre-auth canonical body；latch/auth=1/0、zero-count、
  catalog、txid/nonce核对后才供L1007授权。不得跨transaction沿用该snapshot。
- **C143 authorized state**：只在该transaction插入exact auth后构造，latch/auth=1/1，
  引用原pre-auth body与auth digest；不把auth=1的状态反推成旧pre-auth digest，
  不继承authority_protocol_downgrade_authorizer的签名能力。
- **C144 consume receipt**：embedded guard原子删除exact auth/latch后，在同一
  callback中构造0/0结果，继续DDL/version更新；未提交前仍可整体回滚为1/0，
  Receipt不是外部signed或committed DB evidence，不能独立拿去证明Down成功。
- **C145 latch**：只从本地固定表锁定读取唯一真实latch，核对installation kind、
  DB identity、version/catalog及原L537/1009守卫。只有原受控Down可consume；它
  不具普遍“永久不可删除”语义，不接受caller body或[]作为缺latch的替代。

八条通道分别检查自己的owner/source/transaction/stage，不新增通用“null对象可信”
接受器。本地读取结果不得被当作已签bundle证据跨边界复用；未来若新增跨边界
consumer，须先修订其契约，不能增加第三种evidence kind或默改这些16行。

## 9. 兼容性、阶段和计数

本文新增A的持久化/owner/journal/读取约束、B的逐项角色/形式明示和C的本地用途
限制，均须整体书面审阅。现有部署兼容性UNKNOWN，canonical approved spec-set
及摘要不变；原migration、signer policy与实际数据库不能因文档存在而自动生效。

批准前为130项局部分类、16项未决。本文完整获批后，最终146项精确对照未发现
未决metadata单元，记录146/0的局部设计结果；这不是独立认证、所有codec已实现
或C12安全验收完成。未来若发现冲突仍须保留具体阻塞，不能靠总数146掩盖它。

后续分开规划：只读metadata/独立向量是一项；A存储与consumer、真正认证/codec、
恢复、canonical登记与物理环境各有自己的实施/审批边界。批准本文不直接授权
任何产品代码、SQL/manifest迁移、key管理、workload或main合并。

## 10. 后续验证要求与本轮核对

未来相应实施必须覆盖，以下本轮未运行：

- 16项精确schema/object/role/kind，5/3/8集合；DB/EXT互换、C组两种独立包装、
  错role、same-count missing+extra、未知/重复字段均拒绝。独立expected不由产品生成。
- A owner不存在/重复、partial companion、不同DB/activation/stage、变更旧observation、
  更换epoch/body/time、current scan代替历史body、caller-only可信标记均拒绝；
  原子提交/回滚、不同body exact-retry冲突和orphan guard需DB层验证。
- A108部分source publication、different全局set、改created_at、publication缺owner、
  source-set抢先签发、用未来seal作owner、shared provider漏覆盖、journal回滚均拒绝；
  exact retry只能补剩余publication，不重复retirement或重开scope。
- 不向§6.1增加row_kind，不把读到body或观察WAL point当exact COMMIT；PITR缺失
  无已批恢复能力时保持关闭，不用archive bytes伪造新transaction中的历史记录。
- B三项错误role/schema、Record与genesis Inspect混用、内层Receipt无效、EXT降为DB、
  历史context改写及非法Abort状态均拒绝。C组三阶段/owner/transaction错配、
  任意context注入、缺latch、把未提交consume receipt当成功及本地对象跨界复用均拒绝。

本轮只做文档16行身份/分组/metadata核对、原146项增量集合、owner引用字段、
依赖DAG与链接/差异检查；没有代码测试、DB transaction、签名或负载验证，也没有
独立复审。自检不得留“owner将来再选”或把物理实现缺失伪称已执行的placeholder。

## 11. 书面审阅和恢复

本附录书面已确认，146项最终对照已记录；下一步审阅[metadata-only实施计划](../plans/2026-10-17-authority-v7-metadata-registry.md)
并选择执行方法/验证预算，不直接运行旧Task11命令。所有新增存储责任虽在
本地设计中明确，实际迁移、可信读取配置和恢复能力仍未实施，不计为生产通过。
必要文件保存到现有开发分支；新电脑从[恢复记录](../../runbooks/repository-recovery.md)
和最新提交继续，不重放已消耗的leaf/name-directory/null-fix验证阶段。
