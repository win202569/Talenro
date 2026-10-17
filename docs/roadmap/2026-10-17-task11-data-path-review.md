# Task11：27项矩阵获批后的16项数据通道审查

最新：用户在0ec812cd后确认完整16项书面附录；
[最终146项本地对照](2026-10-17-task11-final-metadata-reconciliation.md)为146/0。
metadata-only实施计划已写出待审阅，存储和消费仍未实施。以下130/16、方向批准
与书面待审阅描述均为历史时点，不是最新状态或独立认证。

后续：用户在399f3d21后确认A/B/C方向。
[完整16项书面附录](../superpowers/specs/2026-10-17-authority-v7-data-path-clarification-design.md)
已写出，明确attempt/activation companion及先于epoch签发的逐source publication owner，
正文仍待审阅；当前130/16不变。以下“方向未批”文字保留为本次来源审查时点。

基线：a40c7b45736197a41b4a96415eb840f4f9aaef25。用户随后确认
[27项请求矩阵](../superpowers/specs/2026-10-17-authority-v7-request-auth-matrix-design.md)，
本地设计状态更新为130项已有局部分类、16项未决；不是实现/验收数量。
本轮核查这16项的保存、传输及消费路径，**本报告中的三类建议尚未获批准**。
没有产品代码、fixture、SQL/manifest、key/policy、canonical集合或运行时变更，
没有build/test、Docker/DB、provider操作、独立复审或旧验证阶段重跑。

## 1. 可在新电脑恢复的完整分类关系

146个精确schema/object仍来自[原初审](2026-10-16-task11-registry-metadata-audit.md)，
原93/53记录不回写。当前文档对照为：

1. 保留原93项执行者有据分类，不借新批准宣称已独立认证。
2. 应用[九项已批澄清](../superpowers/specs/2026-10-17-authority-v7-metadata-nine-item-clarification-design.md)
   的9项增量；加上[补充来源复核](2026-10-17-task11-metadata-current-reconciliation.md)中
   #93的执行者来源判断，得到103/43。
3. 再应用已批请求矩阵中的27个精确schema→profile→role映射，各自EvidenceKinds=[]，
   得到130/16。它们与前10项无重叠。其余109项仍保留原表状态，16项不填默认值。

这些是审查文档之间的关系，不是产品运行时overlay、部分registry或fixture。
全部必要身份、映射、来源与待办随Git保存，不依赖原电脑缓存或对话。

## 2. 三类建议与精确集合（待确认方向）

DB表示database_immutable_body，EXT表示external_signed_envelope；表中值均是
候选方向，不是当前已批准的16行元数据。null/[]必须以未来书面契约的肯定约束
为依据，而不是因为没有找到签名或只见到一个digest引用。

- A：29、30、32、49、108。建议补齐不可变DB preimage保存及owner/consumer关系，
  候选null＋DB，不新增签名角色；不能仅凭投影或集合名称获得DB资格。
- B：89、90、132。建议逐项明确provider envelope及claim_v1_provider角色，候选
  claim_v1_provider＋EXT；不继承父/邻居角色，不开放任意message额外携带证据。
- C：91、113、115、141、142、143、144、145。建议明确本地可信读取/重算通道，
  候选null＋[]；保留所有本地守卫，不使其变成可复用的独立bundle项。

## 3. 16项逐行来源与必须补明的内容

来源L编号指[原v7规格](../superpowers/specs/2026-08-24-nodecontrol-authority-v7-upgrade-design.md)
在本基线的行号。来源只说明现有body/引用/消费；建议新增的限制不冒称原文规则。

| 原编号 / object | schema | 类别 | 已有来源、现有用途与尚缺契约 |
| --- | --- | --- | --- |
| 29 DatabaseLegacyAbsenceProjectionV1 | database-legacy-absence-projection.v1 | A | L409/411/492/553/599：stable projection跨attempt/activation重算并要求byte-identical；须明确完整canonical body在哪个不可变owner下保存和供Prepare读取，不能只保留digest或每次从可变DB猜原文。 |
| 30 DatabaseLegacyAbsenceObservationV1 | database-legacy-absence-observation.v1 | A | L413/492/553：attempt与activation各自产生不同body/digest；须按对应owner原子保存，不能把activation snapshot/time覆盖attempt observation，需保持stage和database point先后约束。 |
| 32 AuthorityGenesisEpochEvidenceV1 | authority-genesis-epoch-evidence.v1 | A | L421/428/492/496/728：mode-specific推导与genesis root贯穿Prepare/lease；须明确attempt前构造、与attempt原子保存、之后不改写原body及完整source-chain消费，不以Evidence后缀或数学摘要当可信提交。 |
| 49 EmptyDatabaseInventoryV1 | empty-database-inventory.v1 | A | L490/492/496/553：仅attempt-time object，Prepare引用，activation引用attempt inventory；须保存同一次attempt的完整原文，不能在activation时重造inventory并伪称同一MVCC观察。 |
| 89 ClaimV1ReceiptV1 | claim-v1-receipt.v1 | B | L719/720/722：provider receipt envelope及Finalize/Abort响应角色已有依据；建议明确其EXT形式资格和允许消费位置，保留reservation/context绑定；独立目录资格不改变message-specific exact-cover。 |
| 90 ClaimV1RecordV1 | claim-v1-record.v1 | B | L720/722/732/734：稳定历史Record、ordinary Inspect响应与独立Record envelope消费；建议补明ordinary Inspect返回的本object使用claim_v1_provider，不与InspectClaimV1GenesisResponseV1混为一物，不从内嵌Receipt继承角色。 |
| 91 IndeterminateSourceSealV1 | indeterminate-source-seal.v1 | C | L537/782/784/809：与requirement同transaction提交、冻结DB，可由后续完整seal引用；建议只由本地可信source读取核对，不提供独立bundle资格，不能用minimal seal授权archive/epoch/target。 |
| 108 LegacySourceRetirementSetV1 | legacy-source-retirement-set.v1 | A | L347/428/786/805/809：global mechanical exact-cover set进入epoch/source/完整seal链；须新增明确的不可变发布owner及发布时点，必须早于需要它的epoch source-set签发，不能等完整seal提交后才保存。 |
| 113 FreshRestoreRequirementV1 | fresh-restore-requirement.v1 | C | L537/784/817：与source seal同transaction、不可改写、固定blocked capabilities；建议本地可信guard读取，[]不取消freeze或允许取值替代完整source archive。 |
| 115 FreshImportTopologyProjectionV1 | fresh-import-topology-projection.v1 | C | L831/837/883：manifest signer/consumer从完整objects独立构造纯投影，排除tx/time/control/evidence；建议保持确定性重算与digest比对，不另造独立DB/signed证据，不能继承FreshTargetInventory签名。 |
| 132 FreshV7StagingExclusionAbortV1 | fresh-v7-staging-exclusion-abort.v1 | B | L877/879：terminal provider Abort响应与CAS明确，但不能仅凭邻段Release角色登记；建议显式补明本schema的claim_v1_provider角色与EXT形式，不改变held/recovery/Release互斥矩阵。 |
| 141 ProviderProtocolDowngradeRetirementSetV1 | provider-protocol-downgrade-retirement-set.v1 | C | L1003/1005：成员envelopes进入bundle，B01验证exact set后经context传给同一Goose transaction；建议明确集合本身只通过可信typed context/重构传递，不继承成员签名，不让caller用任意set body取代验证结果。 |
| 142 PristineDowngradeInventoryV1 | pristine-downgrade-inventory.v1 | C | L1005/1007：pre-authorization的transaction-local inventory与stored pre-auth检查；建议同一tx/locks/nonce内可信构造和使用，不冒充独立committed row或继承B11签名。 |
| 143 PristineDowngradeAuthorizedStateV1 | pristine-downgrade-authorized-state.v1 | C | L1007：同一transaction auth=1阶段投影；建议只供embedded guard核对，不把auth=1重构成pre-auth body，不提供跨transaction复用资格。 |
| 144 PristineDowngradeConsumeReceiptV1 | pristine-downgrade-consume-receipt.v1 | C | L1009：consume后latch/auth=0/0，DDL/version与guard同transaction可整体回滚；建议仅在该callback内部使用，Receipt名称不构成签名或提交证明。 |
| 145 AuthorityProtocolMigrationLatchV1 | authority-protocol-migration-latch.v1 | C | L537/1005/1009：Up写入、Down受控consume；建议本地可信锁定读取，不能当作永不删除的通用DB证据，不能因[]绕过latch/installation kind检查。 |

## 4. 为什么A组不能只改元数据

原契约中的digest引用不等于完整preimage已持久保存。若选择A方向，书面澄清须
冻结每个schema的owner、创建transaction、唯一性/不可变规则、可信读取、跨边界
consumer及恢复责任。可以讨论typed companion storage，但**不能建立无owner的
可变content-addressed store**，也不能让provider从任意URI猜历史body。
既有body字段和digest不能为补存储而改变；新增DB存储/manifest/ACL/guard要求
必须明确列作后续新增责任，不暗称原migration已经支持。

29/30/32/49的候选关系以既有attempt/activation为锚：attempt事务保存当次完整
projection、attempt observation、epoch evidence和inventory；activation产生新的
observation并引用原attempt对象，不能重写旧snapshot、time或body。具体owner关系
和stage唯一性仍须书面冻结，不能把所有对象按单一digest无条件混存。

**#108的依赖顺序**：永久membership＋全部member retirement evidence → 完整set
原文的不可变发布 → epoch source-set/maximum签发 → 完整source seal及archive。
原L805的source-set依赖retirement set，L786的完整seal又依赖source-set；所以不能
把“第一次保存retirement set”的owner放在该完整seal之后。需要单独明确一个在
source-set签发前已提交的受控发布owner，以及global set与每个source DB/retirement
plan的关系。本文不替用户指定任意数据库或创建新的存储表；这仍是书面设计需解
的问题，不能用“owner以后补”宣称该行已通过。

A组的DB形式仅是preimage形式资格，不是新commit proof。尤其不能把这五项加入
原§6.1固定row-kind registry，或把它们伪装成现有某种row_kind来申请attestation。
如果某consumer确实需要新的独立提交见证，必须另作明确契约澄清；不能以普通
DB preimage、父对象签名、后来WAL point或数据库里查到同digest替代。

## 5. B/C两组的边界

B组不引入新角色字符串或新authority，建议只是明确各自provider责任和证据形式；
#132的角色明示、#89的独立资格等超出已明确原文的部分仍是新增本地澄清，待批准。
DB保存external envelope不能把它降为裸DB body；历史Record/Receipt不继承caller
当前context，也不因lease renewal改变历史digest。对任意message而言，目录中的
EXT形式资格不允许extra项，缺失/重复/错误preimage仍按原exact-cover拒绝。

C组[]是建议明确的用途限制，不是删除这些对象或让其免校验。它们各有本地可信
读取、typed context或确定性重算通道，不能用caller JSON、任意context值或相同
digest冒充受信来源。本地source freeze、import投影及Down三阶段是不同职责，
不能因为都取null/[]就共用一个不检查owner/transaction/stage的通用接受路径。

若未来某一原本本地消费实际跨越了provider/attestor边界，必须停下来明确真实
传输和验证责任；不新增第三种bundle evidence，不把本地projection偷偷标为DB
immutable body，也不为了通过目录校验给它套一个邻近签名role。

## 6. 方案比较与推荐方向

推荐按A/B/C三类继续写统一澄清：A五项解决不可变原文与owner/DAG，B三项明确
provider envelope，C八项收紧为本地可信通道。代价是必须正视A的持久化责任，
但可保持现有两种evidence kind、原body/digest和既有签发authority，不凭空增加
一套“通用projection signer”。

另一方案是为全部缺口新增外部attestation，让每份投影/集合都签名。这会扩张
trust/key/authority、职责和wire，且签名不能自动证明DB提交；当前不推荐。
全部填null/[]则无法解释已有的跨边界digest/preimage消费，也不推荐。

本报告仅请求确认三类方向；确认后才写完整16项书面澄清，尤其要闭合A的owner
与依赖顺序。方向确认不等于16行已裁决，当前仍130/16，不能生成完整metadata
实施就绪计划、产品表或“146项已经通过”的声明。

## 7. 核对与恢复

文档核对目标：原146项身份一一对应；原53项精确分割为9项已批、1项来源判断、
27项已批请求和16项未决；11个请求角色/profile允许集不变；A/B/C为5/3/8且无
缺漏/重复。核对文档不证明真实存储、验签、认证或DB提交已经实现。

必要文件保存到现有codex/c12-b01-task9-coordinator分支；新电脑从最新提交和
[恢复记录](../runbooks/repository-recovery.md)恢复。canonical登记/兼容性、物理
环境/PITR、完整Task11/C12、安全/生产及main门槛不变，不重放已消费旧验证阶段。
