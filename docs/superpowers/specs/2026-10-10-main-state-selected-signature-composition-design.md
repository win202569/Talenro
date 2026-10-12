# 主状态材料与签名证据：单目标组合校验设计

日期：2026-10-10。

状态：用户于 2026-10-10 批准本文及对应五项实施计划；Native 本地五项任务
和全部限定回归完成，唯一一次独立增量审查为 0 Critical / 0 Important / 0 Minor，
无遗留 Minor。沿用现有 coordinator worktree 和 PowerShell；不暂存、提交、推送或合并。
审查后仅更新文档，最终只读核对确认源码/测试与基线边界未变。
本次仅交付单目标条件性组合增量，不是完整主状态或生产验收；证据见对应 roadmap execution 记录。

## 1. 意图与成功标准

上一增量的 statematerials 能证明根集合、根轮换、metadata 和 desired
的原始材料与本批预期记录一致，但不验证签名。现有 stateverify 能验证
相对于显式初始根的连续轮换、metadata 和一个 desired 的签名关系，但
不消费材料表，也不证明调用者声明的原始 metadata/root 关联保持不变。

本增量连接这两项检查，回答一个有限问题：

> 本批材料全部符合其声明；其中明确选定的一个 desired、它声明的
> metadata 及从显式初始根到该 metadata 根的轮换链，是否使用同一份
> 原始材料快照，通过了既有条件性签名验证？

成功必须同时满足完整材料子集检查、精确选择与关联绑定、同快照验签和
证据回核。两套分别合法却使用不同材料的输入不得拼成成功结果。错误时
不返回部分模型。对任意额外历史材料，不因选定目标成功而提升其证明状态。

本层不认证初始根来源，不判断当前时间有效，不证明最高已接受版本、
历史原始关联、防回滚、完整主状态材料闭包或生产权限。成功不是 Trusted、
Authorized、AcceptableState、恢复放行凭据或提交能力。

## 2. 依据与方案选择

依据为父设计 `2026-08-23-node-pop-control-plane-design.md`、主状态 L0
契约设计、`2026-10-10-main-state-materials-subset-design.md` 和既有
desired 签名关系设计。实现事实以当前 stateformat/statematerials/
stateverify 源码为准，历史设计文件中的阶段状态不替代后续执行记录。

采用新增纯内存包 `internal/nodecontrol/statecomposition`。另外两条
路线是将验签并入 statematerials，或直接组合可信时间、策略与生产入口。
前者改变已批准的材料层职责，后者涉及尚未提供的平台来源及运行验收；
本批选择单独、可重检的单目标组合层，保留原有层的证据边界。

依赖方向固定为：

- statecomposition → statematerials：检查完整输入子集的格式、预期字段和关系。
- statecomposition → stateverify：唯一签名验证算法，不另写阈值或 Ed25519 逻辑。
- statecomposition → stateformat：解析独立初始根并核对其 payload 与材料表根的字节身份。
- 值类型依赖 contracts.Digest；标准库只用于上下文、错误、复制、排序及比较等纯计算。

不导入 statecontract、statepolicy、statetime、staterecovery、serving、
authority、store 或 trust。不读取其他模型的私有状态，不创建生产适配器。
不引入 provider、数据库、网络、磁盘、环境变量、随机数、时钟、硬件或
后台 goroutine。没有新 wire schema、摘要编码、签名域或依赖下载。

stateformat 仍是唯一 wire 解析者和原始 payload/envelope 摘要计算者。
本批不需要扩展其产品 API；不复制 JSON 校验、不 marshal 重建原始材料，
不直接计算材料摘要，也不允许可替换 verifier 或调用者传入“已验签”结果。

## 3. 输入：材料表、独立锚及明确选择

输入只包含以下三个部分。具体 Go 类型及方法名交由实施计划冻结，
但不得改变本节字段含义或增加可跳过检查的可选参数。

### 3.1 完整材料子集

一份现有 statematerials.Input，包含 scope、材料表和对应预期记录。
四种类型、字段核对、唯一性、严格顺序和历史关系继续遵循其原契约。
材料表可包含不属于本次验签目标的其他合法历史记录；这些记录也必须
通过完整材料子集检查，不能因未选中而跳过格式或引用校验。

### 3.2 独立提供的初始锚

一份 InitialRoot envelope 原始字节，以及调用者独立提供的预期 node/
epoch、anchor payload digest 和 anchor envelope digest。预期摘要均非零；
预期 scope 必须与材料子集 scope 完全一致。

本层不从候选材料、metadata、链尾或响应中选择锚，也不把解析出的摘要
直接当作独立预期。调用者必须从独立上下文提供比较目标；本 API 只能
检查这些显式值相符，无法证明调用者真正独立取得它们或认证其来源。
不添加 trusted bool、来源认证标记或由候选数据自我证明的证书。

锚严格使用既有 InitialRoot 入口：root version=1、一个 root key、
threshold=1，epoch 与独立预期相符。初始根没有签名；不发明自签证明。

### 3.3 单目标选择

选择字段只使用现有 MaterialRef 的 Kind/P/E 身份，不携带另一套原始字节：

| 选择 | 必须命中的材料 | 约束 |
| --- | --- | --- |
| AnchorRoot | RootSet | P 等于独立锚 payload digest，E=0；原始根字节等于锚 payload |
| OrderedLinks | 0..32 个有序链接 | 每项显式给出 RootRotation ref 和 NewRoot RootSet ref |
| Metadata | 一项 Metadata | 原预期记录的 RootPayload 等于所选链尾 RootSet 的 P |
| Desired | 一项 Desired | 原预期记录的 MetadataPayload/RootPayload 等于所选 metadata P 和链尾 root P |

每个选择必须命中材料表及其对应预期记录的完整 ref，而不是仅按版本、
key ID 或 payload 查找后接受任意信封。RootSet 的零 E 与签名制品的非零
E 规则保持；InitialRoot envelope 的 E 只属于独立锚，不填入 RootSet ref。

链从 AnchorRoot 开始。每个 rotation 的原预期 OldRootPayload 必须等于
前一个选中根 P，NewRootPayload 必须等于本项 NewRoot 的 P。原有材料层
已检查内嵌新根与独立 RootSet 的 exact bytes/P 相等；组合层只能使用
刚通过该检查的同一份快照，不能再接受调用者另给的根材料。

无轮换时链尾是 AnchorRoot；否则是最后一个 NewRoot。连续性、epoch、
版本 +1、sequence、双阈值等规则仍由已有材料层和验签层执行。不得排序
链来修复调用者输入、重复链接、跨片段自动补链或在失败后缩短链重试。

保留现有验签 API 的限制：只能从第 1 版根起进行最多 32 次连续轮换，
本次可达根版本不超过 33。材料层接受的高版本根或断开历史片段可以
存在于额外材料中，但不能被本批选作新的锚、替代链尾或冒充已验证链。
此限制是当前 API 能力，不是全系统根版本或生命周期上限。

不接受第二份 DesiredEnvelope/MetadataEnvelope、ActiveRootEnvelope、
预先构造的 stateverify.Batch、SignatureEvidence 或 MaterialSubsetModel。
组合层必须自行从选中的自有原始材料组装 Batch；链尾信封只能是自有
锚 envelope 或最后一个选中 rotation 的原始信封，不能另外指定。

## 4. 输入预算、所有权与同快照语义

在任何输入深复制、解析或索引分配前，先检查 nil context、已有取消和
预算。材料与预期记录各最多 128 项；全部材料原始字节最多 16 MiB；
RootSet raw ≤65,536 bytes，其余三类 raw ≤262,144 bytes。材料子集及
预期中的 rotation 各最多 32 项；OrderedLinks 最多 32 项。

独立锚必须非空且 ≤262,144 bytes，单独计入组合输入预算；因此全部
外来 raw bytes 的总上限为 16 MiB + 262,144 bytes，不改变原材料层
16 MiB 上限。Target refs 不携带 raw，不重复计算 Batch 的内部借用。

深复制前也检查预期记录所有分支的集合上限：root keys 5、metadata
keys 64、revoked IDs 4096、desired slots 8，包括最终会被拒绝的非活动
分支。累计长度计算必须溢出安全。这里只做有界复制所需的形状预检查；
完整字段、wire、唯一性及关联规则仍交给原层，不形成第二套协议规则。

通过预算检查后，组合层创建一个私有、自有的完整输入快照：独立锚 raw、
材料 raw、所有预期嵌套集合、OrderedLinks 和固定选择均独立复制。随后
两个下层只借用这份快照中的材料。下层因原契约产生自己的解析副本或
内部副本是允许的；“同快照”指相同字节与声明，不要求共用解析对象。
输入预算不是总峰值内存承诺，不宣称这些副本合计只占 16 MiB。

顶层材料表与预期记录可按 Kind/P 规范排序以供完整身份比较；零长度
集合规范化为 nil。嵌套规范列表不得重排，OrderedLinks 必须保留原顺序。
调用期间调用者不得并发修改输入；所有者不得浅复制或与 Clear 并发访问。

## 5. 评估流程与证据回核

入口概念上为 Evaluate(ctx, Input)，顺序如下：

1. 执行取消及预算预检查，建立完整自有快照，立即登记错误/取消/panic 清理。
2. 对快照中的完整 statematerials.Input 调用 Evaluate。失败即终止；
   不仅检查 Target 命中的材料，不接受以前调用的缓存模型作为本次证据。
   成功后建立只引用自有材料及原预期记录的有界索引。
3. 用 stateformat.ParseEnvelope(InitialRoot, 自有锚 raw) 解析独立锚，
   核对预期 P/E、epoch、version、key/threshold 和 scope。取得独立复制
   payload，逐字比较 AnchorRoot 材料 raw；比较后清理临时 payload。
4. 使用该索引核对全部 Target ref，
   有序逐项检查 old/new root 选择、metadata 的根关联和 desired 的原始
   metadata/root 关联；不存在的材料、不同 E 或错误 Kind 均终止。
5. 从上述同一快照组装 stateverify.Batch：ExpectedNodeID/Epoch 使用
   已核对的独立 scope；anchor 使用自有锚，Rotations 使用有序选中 raw，
   ActiveRootEnvelope 使用锚或最后一个 rotation raw，metadata/desired
   使用精确选中 raw。只调用既有 Verify，不增加宽松路径或候补尝试。
6. 验签成功后仍逐项回核 SignatureEvidence：anchor P/E 等于独立预期，
   rotation P/E 按顺序与所选 refs 一致，metadata/desired P/E 精确一致。
   回核 node/epoch、链尾 root version、metadata version、desired generation/
   sequence/signing key ID 与同快照预期相等。不得只核对数量或最终版本。
7. 构造私有结果与确定性的材料证明清单，最后再次检查取消；只在所有
   检查完成后转移结果所有权。临时格式对象、子集模型和签名证据均清理。

组合层返回自己的私有字段结果类型，不导出或共享下层临时所有者指针。
不得把材料层 Summary 的 SubsetConsistent 直接改名为“全部验签通过”。

### 5.1 声明关联不变不等于 wire 签入了关联

Desired wire 没有 root/metadata payload digest；metadata wire 也没有
root payload digest。组合层证明的是本次完整声明中的这些关系与所选
验签材料一致，不声称签名覆盖了 wire 中不存在的字段。

若另一份合法 metadata 也能验证同一个 desired，但不是本次声明的
metadata ref，选择它必须失败。若调用者同时改写声明与材料并新建一次
Evaluate，而所有新声明关系合法，本层没有旧状态，不能断言检测出历史
原始授权者被改写。已有模型的 Recheck 会拒绝这种变化；跨状态原始关联
保持仍需后续与完整 L0/持久历史的设计，不能凭本层成功消除该责任。

## 6. 历史、撤销与条件性结果

本批不读取时钟，不要求所有历史材料当前有效。选中的历史链、metadata
和 desired 如满足现有静态窗口及签名关系，即使相对于当前日期过期，也
只能形成条件性证据；不能用于网络补链、恢复、部署或当前运行授权。

不改变 stateverify 的 key 状态及撤销语义：所选 metadata 必须含相应
active/retiring key，且该 key 不在它的 revoked IDs 中，并满足原有
签发/有效期静态边界。选择一份已撤销该 key 的 metadata 仍须失败，不能
以“这是历史 desired”跳过验签层规则。

若历史 desired 的声明关联的是原来仍含适用 key 的旧 metadata，可相对
该旧 metadata 验证；另有较新、包含撤销记录的合法 metadata 可以保留在
材料表中，但它不属于本次选中证据。此成功不证明 desired 当前可使用，
也不认证旧 metadata 是真实签发时授权者。材料层的原有历史保留规则不变。

私有 CompositionModel 的只读摘要必须同时呈现以下不同事实：

| 对象 | 本次可报告的证明状态 |
| --- | --- |
| 完整材料表 | 原 statematerials 子集一致性摘要，固定未完成责任保持原样 |
| 所选初始 RootSet | 与显式锚 payload 逐字绑定；锚来源未认证，不标为已签名 |
| 所选 rotation 及其新 RootSet | rotation 相对给定锚满足双阈值；新根由该选中链导出 |
| 所选 metadata、desired | 仅对所选链/metadata 的条件性签名关系已检查 |
| 其他每项材料 | 仅材料一致，未纳入本次条件性签名证明 |

结果清单逐项覆盖整个材料表，按 Kind/P 确定性排序；类型与 P/E 身份
一起显示，不把 RootSet P 换成 rotation P，也不把锚 E 填进 RootSet。
锚本身还报告独立 InitialRoot 的 P/E；有序轮换及最终根身份独立保留。

成功状态命名限定为 ConditionalTargetVerified 一类，不使用 Trusted/
Authorized/Ready。摘要不携带完整 raw、公钥材料、下层所有者或放行函数。
所有清单访问器必须返回 detached copies；nil/zero/cleared 模型摘要为
全零无效结果，不能让零值伪装为“未选中材料已证明不存在”。

无论本次额外材料是否为空，固定保留：外部锚来源、未覆盖签名及完整
历史、域业务与可信当前时间、完整主状态材料闭包、完整主状态组合复核、
物理提交和运行授权等未完成责任。单目标重检只完成本层的重检，不关闭
statematerials 原有 CompositionRecheck 或主状态 L0 的组合义务。

## 7. 完整重检、清理与错误

概念 Recheck(ctx, Model, Input) 必须重新执行完整 Evaluate：重新解析
全材料子集、核对锚和 Target、重新验签、回核证据。随后比较保存的完整
规范输入与本次新模型的规范输入，包含未选中材料、每个预期字段、独立
scope/P/E/raw、所有选择及有序链接。不得仅比较所选 desired P、摘要清单
或缓存 SignatureEvidence。合法的材料表重排、nil/empty 等价可通过；
改变链顺序、任何完整记录或未选中材料的合法替换也必须拒绝为模型不符。

Clear nil-safe、幂等，清理所有可达自有 raw、预期集合、链接、清单和
固定字段；模型清理后不可重检成功。Clear 不修改调用者原始输入或先前
返回的 detached 摘要。失败、取消或 panic 清理尚未交付结果和临时对象；
panic 在清理后传播，不转成成功，也不新增产品测试 hook。

不承诺擦除不可变 string、Go 运行时副本、解析库内部存储或其他所有者
持有的 detached copies；不支持同一模型并发 Clear/访问或并发修改输入。

错误契约必须区分 InvalidInput、LimitExceeded、InvalidMaterial、
MissingMaterial、DuplicateMaterial、BindingMismatch、SignatureMismatch、
ClearedModel 及 context 取消/超时。下层固定错误映射为本层固定类别，
不回显原始字节、key/摘要值、JSON 错误或其他动态敏感内容。具体 sentinel
名称与映射表由计划冻结，不能用一个通用失败替代所有关系/验签分类。

nil context 为 InvalidInput；已有 ctx.Err() 优先。底层 stateverify 的
Canceled 不得映射成签名错误；标准 context 取消或超时保留实际 ctx.Err()
以便 errors.Is 识别。检查点至少位于入口、预算/复制/选择遍历、每阶段
下层调用前后、回核及最终交付前。不承诺抢占一次有界解析/标准验签，
不启动后台 goroutine 或声称实时取消 SLA。

## 8. 依赖边界的唯一例外与允许变更

当前两个下层的测试守卫分别禁止 statematerials 的所有非测试消费者，
以及除 statepolicy 外的 stateverify 非测试消费者。未来实施需要以下
明确且仅限测试守卫的例外；当前文档阶段不修改这些守卫：

- statematerials 仅允许文件目录恰为 internal/nodecontrol/statecomposition、
  Go package 名恰为 statecomposition 的非测试文件消费它。
- stateverify 保留已有同路径/同包名的 statepolicy 例外，额外仅允许
  上述精确 statecomposition 路径/包名；仍禁止生产消费者导入 statepolicy。
- 新 statecomposition 的守卫禁止任何非测试消费者，且限制自身纯计算
  依赖和所用导出成员，不允许 JSON、hash、crypto、time、os、net、数据库、
  随机、unsafe、反射调用能力、goroutine 或 provider 绕过。

必须有 AST 负向测试拒绝类似前缀目录、子目录、同路径错误包名、其他
reader/adapter、点导入、空导入、重复路径/混淆别名及禁用成员。reflect
若用于私有规范输入 DeepEqual，只允许该纯值比较，不允许一般反射能力。
不能用字符串包含、全 nodecontrol 白名单或仅允许包名来替代精确边界。

可实施变更仅为新 statecomposition 包和测试、statematerials/
stateverify 的上述边界测试及相应设计/计划/执行/进度文档。原有材料层、
验签层和 stateformat 的产品实现/API/成功规则保持不变；statecontract
及其生产消费者守卫不变。不得额外修改 serving/authority/store/SQL/
schema/生成代码/锁文件/基础规范/API/平台 provider，或顺带修复旧问题。

若实现确实需要下层产品 API 改动、多个 desired、状态策略、可信时间或
主状态自动投影，须先报告超出此增量范围，不能默默扩展设计。

## 9. 测试与验收要求

沿用 Native 执行：实施计划安排先失败测试，再最小实现及限定回归；不
因当前文档批准自动开始实现。测试必须覆盖：

- 0/1/32 次真实签名轮换、33 次超限，选择顺序、连续链、缺失/重复/
  跨 epoch/错误版本，材料表另有断开历史片段但不得替代目标链。
- 独立锚 P/E/scope 与 Raw、AnchorRoot exact bytes 的逐项不符；高版本
  RootSet 格式合法但不能当 InitialRoot；另一套自签数据只有在另给它的
  独立比较目标时可形成条件性证据，仍显式保留锚来源未认证责任。
- 每种选择的 Kind/P/E 完整绑定，RootSet/rotation P 不可互换，锚 E
  与 RootSet E 不可互换；合法同 P 不同 E 信封不能用于替代原选择或重检。
- 所选 metadata/desired 与原预期关联、全部证据 P/E 及坐标逐项相等。
  相容但不同 metadata 可通过单独验签，却不得绕过本层原关联；同时
  合法改写声明的全新 Evaluate 与已有模型 Recheck 有不同、明确的结果。
- 格式正确的假签名、缺失/撤销 key、已知坏份额及阈值不足拒绝；原
  验签双阈值及完整 transcript 规则不变，不以本层 mock verifier 代替。
- 历史过期材料可产生条件性证据；旧 metadata 的合法 key 与额外新撤销
  metadata 共存，摘要仍把后者标为未选中、当前授权未证明；选中 revoked
  metadata 必须失败。额外未选中坏签名可保持格式合法但绝不被标为验签。
- 未选中材料仍须完成所有格式、预期与关系检查；四类齐全/零额外材料
  均不能移除固定义务。每项材料在清单中恰好一次，逐项证明状态精确。
- 材料/预期计数、每项 raw、总 raw、独立锚、链接及每种嵌套集合预算的
  边界与超限；预检查发生在深复制前，无界非活动分支也不能绕过预算。
- 所有 raw/嵌套集合/链接/清单独立所有权，逐次单独变异立即断言；输入
  修改不影响已有模型，Clear 不改调用方或 detached 摘要，nil/zero/
  cleared 无成功结果。Recheck 重新验签且比较完整快照，包括未选中材料。
- 预取消、阶段之间取消、验签后回核/最终交付取消、超时、失败清理和
  panic 清理；全失败路径返回 nil 模型，不留下可消费的部分证据。
- 精确依赖/调用例外与反例、两个下层原逻辑/格式向量不变，L0 和生产
  消费者仍被禁止。不新增为测试服务的产品 hook 或可关闭校验配置。

正向 fixture 使用既有独立真实签名向量或冻结的测试专用种子；关键期望
P/E/原始字节不能只从被测组合层或解析器即时生成，避免自验证循环。
测试代码可以使用测试专用签名构造材料，产品代码不得获得签名能力。
对关键绑定/清单/重检负向测试做受控敏感性检查，临时变异须精确恢复。

未来计划请求的离线回归只包括新包及既有 statematerials、stateformat、
stateverify、statepolicy、statecontract、staterecovery、statetime、
serving、contracts、readiness；authority 仅 TestBoundRead/TestCoordinator，
store 仅编译，不跑测试。沿用固定离线 Go 及缓存依赖、PowerShell。
不运行全仓库测试、真实 DB/Docker、race、硬件、重启或 C11/C12/PITR 验收，
不改变安全配置、绕过既有安全限制或宣称生产兼容性。

最终审查与测试策略由 Native 实施计划固定。本文件只有作者设计自查，
不是独立安全审查或产品测试报告；不因为本文存在而关闭以前的验收阻塞。

## 10. 当前交付、后续阶段与自查

本轮交付本文及 current-status 的文档阶段记录。保留已有未提交内容，
不执行任何 Git 写入；brainstorming 技能的设计提交建议不覆盖当前已
批准的“不暂存/提交/推送/合并”范围。

用户审阅批准本文后，再编写逐任务实施计划，冻结 Go API、文件拆分、
错误映射、测试命令、Native 执行和唯一边界例外；计划审阅后才实现。
不重复询问已选 Native 方法，不把对话范围批准当成尚未存在的计划批准。

作者已按实际接口自查：同一自有快照、仅一个选中 desired、独立锚 P/E、
RootSet 与信封身份分离、声明关联与签名字段分离、原撤销语义、完整重检、
未选中材料逐项标记、精确依赖例外及固定未完成责任均已明确。此自查
不代表新组件已实现、测试通过、主状态闭包或 Task10/B01/C12/main 验收完成。
