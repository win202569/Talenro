# 主状态声明与已验证材料：双快照投影绑定设计

日期：2026-10-10。

状态：用户已选择新增隔离层方案、确认聊天设计，并于 2026-10-10 批准本文。
对应五项实施计划已于 2026-10-10 获用户批准，Native 五项局部实施已完成。
一次只读增量审查为 0 Critical / 0 Important / 2 Minor；两项作者修正并完成
十四条限定重跑，全部 exit 0，无第二次独立复审。规范证明边界未改变；
见 ../../roadmap/2026-10-10-main-state-projection-binding-execution.md。
执行不暂存、提交、推送或合并；进度与验收以计划和专用台账为准。

## 1. 意图、依据和成功标准

statecontract 已检查旧、新主状态的十二域覆盖、结构、声明间关系及有限
高水位转换。statecomposition 已在自有材料快照上检查完整输入子集，
并对一个明确选中的 desired 做相对于独立锚的条件性验签。二者目前没有
共同输入所有者，也没有证明主状态字段对应那一批实际解析/验签的材料。

本增量新增纯内存 `internal/nodecontrol/statebinding`，回答以下有限问题：

> 同一份自有事务快照中的旧、新主状态，是否通过完整 L0 检查；它们的
> root、metadata、已建立 desired 和本次 root transition 的可映射字段，
> 是否分别对应各自材料表；新状态明确选中的 desired 是否就是本次
> statecomposition 条件性验签的目标？

依据为以下已批准设计和实际公开类型：

- `2026-10-10-task10-main-state-contract-design.md` §4–9。
- `2026-10-10-main-state-materials-subset-design.md`。
- `2026-10-10-main-state-selected-signature-composition-design.md`。
- statecontract.ModelInput、statematerials.Input、statecomposition.Input。

成功结论只能是 ConditionalProjectionBound。它不是完整材料闭包、全域
语义验证、可信旧基线、已签入的历史关联、可提交候选或运行许可。保留
全部下层未完成责任；不因做了本批重检就删除 FullStateComposition。

## 2. 已选择方案与实施范围

已比较三条路线：新增隔离层；扩展 statecomposition 承载完整主状态；
先补齐证书/CA、部署、recovery、LKG 等所有材料校验器。用户选择第一条，
使本次只解决明确的投影绑定接口，不改变下层产品 API 或扩大单目标证明。

本增量限于新 statebinding 包、其测试/文档，以及下列三个既有测试文件
中的精确消费例外与反例：

- `internal/nodecontrol/statecontract/boundary_test.go`。
- `internal/nodecontrol/statematerials/boundary_test.go`。
- `internal/nodecontrol/statecomposition/boundary_test.go`。

不修改下层产品代码、stateverify 守卫、statepolicy 规则、wire/schema、
签名算法、原格式向量、SQL、模块或锁文件。无生产消费者，不接入 serving、
reader、启动编排、存储、seal、guard/counter 或 provider。

工作区沿用 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`，
保留既有未提交进度。使用 PowerShell、apply_patch 和固定 Go 1.26.5 离线
缓存；不新建/清理 worktree，不写 Git 索引，不提交、推送或合并。
这是用户范围对技能默认提交/清理步骤的覆盖；代价是文档和进度仍仅在本机，
不构成异机恢复检查点，整体兼容性和生产验收仍未知。

## 3. 输入、公开入口与所有者

固定输入组成，不接收外来成功布尔值、下层模型、缓存 evidence 或摘要：

```go
type Input struct {
    Main     statecontract.ModelInput
    Previous statematerials.Input
    Proposed statecomposition.Input
}
func Evaluate(ctx context.Context, in Input) (*BindingModel, error)
func Recheck(ctx context.Context, model *BindingModel, in Input) error
func (model *BindingModel) Summary() Summary
func (model *BindingModel) Clear()
```

Main.Previous 与 Previous 材料表绑定；Main.Proposed 与 Proposed.Materials
绑定。Proposed 的 Anchor 和 Target 原样使用现有公开类型。新状态 Desired
必须 Present，且 Target.Desired 精确等于该域的材料引用。旧状态 Desired
可 Present 或合法 NeverEstablished；后者不虚构旧 desired、旧目标或签名。
新状态 Desired 为 NeverEstablished、不含 desired 的更新、多个目标均不在本接口范围，
拒绝而不自动挑选目标。这是本批能力限制，不改变 L0 对合法状态的定义。

两张材料表各自保留现有 Kind/P 唯一性。同一 P 在旧、新表中可以具有不同
合法 E；不得合并为一张表、按 P 丢弃旧 E、挑选任一 envelope 或回退重试。
同表重复仍由材料层拒绝。整体“同快照”指一次拥有以上全部输入，不是
声称从外部平台原子读取了真实状态。

Proposed 仍受下层显式 v1/单 root key/threshold 1 初始锚及最多 32 次连续
轮换限制；不引入高版本替代锚、自动补链或链截短。各表的 32 次轮换预算
不是允许截断终身 root 历史；完整历史仍在未完成责任中。

BindingModel 字段私有，拥有完整规范化 Input 及从本次调用捕获的三个
下层摘要，不持有下层临时所有者。摘要中所有可变 Links/Coverage 独立
保存并在每次公开 Summary 时再复制，不能泄露模型内部集合。接口类型、
私有字段和导入成员清单在实施计划中逐项固定，不增加第二种成功模型。

## 4. 预算、复制、规范化和清理

nil context 拒绝；预取消/超时优先。任何深复制或索引分配前，检查全部
输入集合，包括当前 Kind 不使用的 expectation 分支：

| 范围 | 固定上限 |
| --- | --- |
| 旧、新材料表各自的 Materials / Expected | 各 128 项 |
| 两表各自的材料及预期 RootRotation 数 | 各 32 项 |
| Proposed.Target.Links / Main.RootTransition | 各 32 项 |
| 每份 RootSet raw | 65,536 bytes |
| 其他材料及 Proposed.Anchor raw | 各 262,144 bytes |
| 每张材料表 raw 总量 | 各 16 MiB |
| 外来 raw 总量，按出现次数计、不按别名去重 | 32 MiB + 262,144 bytes |
| 每份 expectation 所有分支的 keys/revoked/slots | root 5、metadata 64、revoked 4096、slots 8 |
| Main 旧、新各自九个集合 | root keys 5、metadata keys 64、revoked 4096、两 purpose deauthorized 各 128、desired slots 8、LKG entries 8、local faults 64、supervisor faults 16 |

上限 equal 只证明预算接受，不代表任意等长字节满足格式或语义。累计使用
`len(raw) > limit-total`，不先求可能溢出的和。预算是外来输入界限，不是
峰值内存承诺；本层快照、下层复制和临时解析对象会同时占用内存。

通过预算后深复制两表所有 raw/expectation 嵌套集合、独立锚、目标链接，
以及 Main 旧、新所有九个集合和 RootTransition，共十九个 L0 集合。
复制期间检查取消；从拥有数据开始立即登记失败/取消/panic 清理。

两表只规范化顶层 Materials/Expected 为 Kind/P 顺序；所有长度为零的
集合规范化为 nil。Main 各有序集合、RootTransition 和 Target.Links 不
排序、不补齐、不截断，保留原顺序交给下层检查。不得用排序掩盖非法声明。

Clear nil-safe、幂等，清零可达自有 raw/集合及内部摘要 Links/Coverage，
随后模型为零值；不修改调用者输入或已返回的独立摘要。不承诺字符串、
不可达运行时副本或调用者副本擦除。输入在调用期间不可并发修改；模型
不可浅复制，不支持并发 Clear/访问。违约可能产生竞态或无效状态。

## 5. 同快照检查顺序

1. 检查 context/全部预算，建立完整自有快照，立即登记失败清理。
2. 对自有 Main 调用 statecontract.EvaluateModel，验证全部十二域及旧→新
   声明转换，不只抽出三个域另跑一个宽松 L0。
3. 对自有 Previous 调用 statematerials.Evaluate，检查完整旧材料表。
4. 对自有 Proposed 调用 statecomposition.Evaluate，使用唯一现有路径
   检查完整新材料表、独立 v1 锚、有序链、真实验签及证据回核。
5. 按 §6 精确绑定两份主状态、RootTransition 和选中目标；捕获下层摘要并
   立即登记其可变集合的失败清理，核对摘要
   状态分别为 ContractShapeValid、SubsetConsistent、ConditionalTargetVerified，
   摘要 scope 与对应自有输入一致。
6. 最终复查 context，转移快照/独立摘要所有权给 BindingModel。

每个下层所有者返回后，先登记 Clear，再调用任何后续 context 检查或阶段。
全部普通失败/取消返回 nil 模型；panic 清理后传播，不吞掉或伪装为成功。
本层不自行解析、计算材料摘要、验签、读取当前时间、磁盘、网络、环境、
随机数或硬件，不增加重试、后台任务或 goroutine。

## 6. 精确投影规则与不证明的字段

映射按封闭 Kind 转换，不能将不同类型 enum 直接强转或把 RootSet 的
E=0 换成 InitialRoot/rotation 的 envelope。每个被引用对象必须在对应
表的 Materials 和 Expected 中命中完整 Kind/P/E。其他表项仍参与完整
下层检查与重检，不因不参与主状态投影而忽略。

| 主状态声明 | 与对应已检查 expectation 的比较 |
| --- | --- |
| Scope.NodeID / AuthorityEpoch | 表 Scope.NodeID / Epoch；被投影 expectation.Epoch 同时一致 |
| Root.Mark.Material / Version / Sequence | RootPayload→RootSet，完整 P/E（E=0）、Root.Version、Sequence |
| Root.Threshold / KeyIDs | Root.Threshold（精确数值）、完整有序 KeyIDs |
| Metadata.Mark / RootVersion / RootPayload | Metadata Kind 的完整 P/E、Version、Sequence、RootVersion、RootPayload |
| Metadata.KeyIDs / RevokedIDs / Window | 完整有序集合、Issued/Until，不只比较数量 |
| Present Desired.Mark / InventoryVersion / ResourceVersion | DesiredSnapshot→Desired，完整 P/E、Generation、Sequence、InventoryVersion、ResourceVersion |
| Present Desired.ResourcePayload / RootPayload / MetadataPayload / SigningKeyID / Window / SlotIDs | 所有对应声明值、节点身份、完整有序 slots 与时间窗 |

上述规则分别应用于 Previous 与 Proposed；旧 desired 未曾建立时不生成
默认成功记录。metadata/desired 所需的历史 root/metadata 由材料层现有
关系检查确保存在，绝不把它们自动替换为当前主状态的 active root/metadata。

### 6.1 本次 RootTransition

Main.RootTransition 保留 L0 的精确有序转换语义。每个 RotationRef 的
RootRotation 完整 P/E、Sequence、PreviousVersion、NewVersion、OldRootPayload、
NewRootPayload 必须对应 Proposed.Materials 中的轮换 expectation；两端
RootSet 必须存在且版本、payload 和序列满足下层材料关系。转换起点绑定
Main.Previous.Root 的身份/版本/序列，终点绑定 Main.Proposed.Root。
根未更新时必须遵循 L0 的空 transition 规则，不自动寻找或省略链接。

Main.RootTransition 与 Proposed.Target.Links 不是同一条概念链：前者是
旧→新 active root 的本次更新，后者是独立锚→目标原授权根的验签路径。
不要求二者列表相等，不强制验签路径延长到新的 active root。
落在目标路径上的材料保留下层相应证明等级；其他本次更新材料即使字段
绑定成功，仍只 MaterialOnly，不能把投影成功提升成轮换签名全部已验证。

### 6.2 新状态的明确目标

Proposed.Target.Desired 必须等于 Main.Proposed.Desired.Mark.Material 的
类型化完整引用。其 Metadata.Payload 必须等于该 Desired.MetadataPayload；
完整 metadata E 从新表原 expectation/选中 ref 精确命中，不能按 P
任选信封。验签链尾 RootSet 必须等于该 Desired.RootPayload，并和选中
metadata 的原 root 声明一致；版本及签名 key 等坐标回核沿用下层结果。

目标可以沿用历史授权 metadata/root，且不等于 Main.Proposed 当前 active
metadata/root。不得为了匹配新状态而改写它的原声明、换成最新 metadata、
跨表借用信封、自动选更短链或重试。静态合法但已过期的材料可以用于此
历史/恢复模型，仍不产生当前运行授权。

### 6.3 未覆盖字段保持明确未验证

Metadata.RevokedSetDigest 和 Metadata.History 尚无本批对应的完整材料/
编码契约，本层不重算、不假称与 revoked 列表或历史内容绑定。其 L0 结构
与转换规则照常检查，完整值进入旧模型 Recheck，语义义务仍保留。

MainStateScope 的 DeploymentID、TrustAnchorID、DeploymentKeySetID、
TimeProviderID、三个 guard/key 身份及 TimeFloor 只获得既有 L0 检查，
不证明平台来源或实际隔离。尤其不擅自规定 TrustAnchorID 等于本次独立锚
的 P/E：现有类型没有该转换契约；它与外部锚的部署信任关系仍待独立设计。
“scope 绑定”在本批仅指 node/authority epoch 的材料坐标对应，不泛化为
全部平台身份已经验证。

其余安全域的原始证书/CA、部署/资源/主机策略、recovery、supervisor、
LKG 和各类完整历史没有本批验证器。完整 L0 的合法声明不能变为这些域的
材料证据。所有字段仍保存并参与 Recheck；这只证明输入未变，不证明来源。

desired 的 metadata/root payload 关联、metadata 的 root payload 关联并
非现有 wire 已签入字段。新 Evaluate 中所有关联声明连同旧、新 L0 合法
一致改写后可能成立；这仍不是“原历史关联已获授权”。旧模型重检必须拒绝
这类变化。旧状态来自调用者，不是已经认证的持久基线；L0 拒绝声明回滚
不能被描述成真实存储防回滚或可信安装历史。

## 7. 结果、责任、重检和错误

Summary 的固定内容为本层状态、ContractSummary、旧材料 Summary、新
Composition Summary 及本层固定八项未完成责任。下层 5/6/7 项责任及新
composition 内原子集的 6 项责任原样保留，不重标 coverage，不删除义务：

1. ExternalAnchorAndPlatformProvenance。
2. PreviousSignaturesAndUnselectedHistory。
3. UnverifiedHistoryAndSetDigests。
4. DomainAndCurrentTime。
5. FullStateMaterialClosure。
6. FullStateComposition。
7. PhysicalCommit。
8. RuntimeAuthorization。

本层状态 enum 0=InvalidResult，1=ConditionalProjectionBound；义务 enum
0=UnknownObligation，非零项按以上顺序。nil/零值/已清理模型的 Summary
为完整零值。旧表没有条件性验签结论；新表的 AnchorAssumed/ChainDerived/
SignatureChecked/MaterialOnly 完全沿用本次下层结果。主状态 active root、
当前 metadata 或 root transition 被绑定，不额外取得 SignatureChecked。

Recheck 首先处理 nil/预取消 context 和缺失/已清理模型。随后对新输入
重新运行完整 Evaluate，立即登记 fresh.Clear，精确比较整个规范化 Input：
Main 旧/新全部字段与 RootTransition、两表所有材料/expectation、独立锚
scope/P/E/raw、目标与有序 links，包括未选和未投影内容。比较后再次检查
取消；不同返回 ErrBindingMismatch。不只比较摘要、选中 P 或几个高水位，
也不调用外来缓存证据。新输入本身合法不代表它属于旧模型。

固定错误分类为 ErrInvalidInput、ErrLimitExceeded、ErrIncompleteCoverage、
ErrInvalidMaterial、ErrMissingMaterial、ErrDuplicateMaterial、ErrBindingMismatch、
ErrSignatureMismatch、ErrRollback、ErrConflict、ErrClearedModel，文字固定
以 `statebinding: ` 开头，不附输入值、UUID、摘要、raw、路径或秘密。

当前 ctx.Err 优先；context.Canceled/DeadlineExceeded 保留 errors.Is。
statecontract 的 InvalidInput/IncompleteCoverage/BindingMismatch/Rollback/
Conflict/StaleModel 分别映射本层 InvalidInput/IncompleteCoverage/
BindingMismatch/Rollback/Conflict/ClearedModel；材料层各类别同义映射，
ReferenceMismatch 映射 BindingMismatch；composition 各类别同义映射。
未知非 nil 错误固定为 InvalidInput，不透传动态文本。新 desired 合法
NeverEstablished 或非法目标 Kind 是 InvalidInput；缺少投影对象是
MissingMaterial，已存在但字段/完整 E/关联不符是 BindingMismatch。
本层预算超限先于深复制和下层格式错误，预取消仍先于预算。

## 8. 隔离边界和反例

新包仅依赖 context/errors/bytes/sort/reflect 的必要成员、contracts.Digest
和上述三个下层的固定类型/入口/摘要/错误。reflect 只允许 DeepEqual；
不导入 stateformat/stateverify/statepolicy/statetime/staterecovery/provider，
不开放 JSON/hash/crypto/I/O/time/random/unsafe 或 goroutine 能力。

三个旧守卫各仅允许非测试文件目录精确为
`internal/nodecontrol/statebinding` 且 Go package 名为 `statebinding`
的消费例外；保留 statematerials 已有精确 statecomposition 例外。
旧下层之间的禁用依赖不放宽。新包禁止任何非测试消费者。

反例必须覆盖 reader/serving、子目录、相似前缀、错误包名、点/空 import、
重复 import 路径、冲突别名、禁用导出成员和 GoStmt。全仓库非测试文件
扫描验证没有新消费者；清理 AST/源码检查确认拥有快照及每个临时 owner
后，在下层调用或后续 ctx.Err 前登记清理。测试可以读取冻结向量、构造
签名和检查源码；这些能力不进入产品白名单，不重写既有 fixture 生成器。

## 9. 测试设计和验收边界

实施计划必须给下列场景明确命名测试、独立断言和 RED/GREEN 记录：

1. 两套正常材料与完整 L0 映射；旧 desired Present/合法 NeverEstablished；
   目标 0/1/32 链、主 root 无变化/连续变化。新 Desired 的 NeverEstablished 拒绝。
2. 每个投影字段独立错配：node/epoch、Kind/P/E、version/sequence、keys/
   revoked/slots、window、资源/授权引用、RootTransition 各坐标。避免早先
   错误遮蔽要测试的绑定分支；新旧两侧分别断言。
3. 同 P 不同合法 E 分别在两表成立并精确对应各自主状态；错误 E 不接受。
   新完整输入可 Evaluate 的信封替换仍不通过旧模型 Recheck。
4. 目标原授权 root/metadata 比当前主状态旧仍可条件成立；不得自动换成
   当前材料。未选 current metadata 或 transition 的格式正确坏签名不升级
   coverage；未选/旧表材料损坏、缺失、预期不符仍使整个调用失败。
5. 合法一致改写 metadata/root 声明的新 Evaluate 与旧 Recheck 身份拒绝
   分开证明；不把自报旧 L0 或公开 trust-anchor ID 当成历史授权来源。
6. 每项预算 equal/+1，全部十九个 L0 集合及两表四类 expectation 分支
   先查长度；总 raw、单份 raw、轮换/链接各自上限，预取消优先。
7. 原输入与每次摘要独立；nil/zero/Clear 幂等；可达自有别名清零；全部
   未投影字段、旧表额外材料和新表未选材料改变均影响完整 Recheck。
   仅顶层表重排/nil-empty 规范化等价，链接或有序域集合乱序不等价。
8. 分阶段及最后取消/超时返回 nil 模型；拥有快照后的 panic 清理后传播，
   不修改调用者；保留别名清理断言与早登记 AST 证据，不假称观察所有
   不可达运行时副本。固定错误与 context 映射逐项断言。
9. 每次成功均保留全部责任，只有 ConditionalProjectionBound；旧材料
   不被提升验签、目标外材料不被提升覆盖；三个精确消费例外及反例。

实施后仅跑固定离线范围：完整 statebinding、statecontract、statematerials、
statecomposition、stateformat、stateverify、statepolicy、staterecovery、
statetime、serving、contracts 和实际 `internal/readiness` 包；指定 authority
的 TestBoundRead/TestCoordinator；store 仅编译（不作 store 测试验收）。
具体命令、固定环境、分包退出码及测试后源码冻结在实施计划中列明。
不跑全仓库、真实 DB/Docker、race、硬件、重启或 C11/C12/PITR，不改测试
deadline，不下载依赖，不用输出延迟当作测试超时或重复启动未结束会话。

目前只是设计：没有本增量实现/测试/独立源码审查结论。实施计划需冻结
允许变更的三个旧测试文件及所有旧 Go/SQL/module/lock 哈希基线，保留
准确退出码和仍未完成的整体 Task10/B01/C12/main 验收条件。新增源码
审查的执行方法由随后实施计划呈现，不能把本次作者设计自查称为源码审查。

## 10. 作者自查与审批交接

已针对现有公开类型逐项自查：两表避免同 P 不同 E 的合法跨状态变化被
错误合并；主状态 root 更新和目标原授权链没有混用；同快照没有冒充真实
平台原子读取；旧 L0 没有冒充可信基线；历史集合摘要和 trust-anchor ID
没有虚构 wire 对应关系。十九个 L0 集合、两表全部 raw/嵌套集合及目标
链接均在所有权/预算/重检范围，公开摘要不泄露内部集合。

范围保持单个投影绑定增量，证书、CA、recovery、部署、LKG 的完整材料
及业务验证器仍需后续独立设计；完整状态格式、提交和运行验收未解除。
本文没有保留待填写字段或默认成功出口；精确代码文件分工与测试命令
属于下一份实施计划，不在此阶段执行。

用户于 2026-10-10 已批准本文；已使用 writing-plans 编写对应
`docs/superpowers/plans/2026-10-10-main-state-projection-binding.md`。
用户随后已批准对应计划，保留 Native 执行方法；不改变本设计的证明边界。

