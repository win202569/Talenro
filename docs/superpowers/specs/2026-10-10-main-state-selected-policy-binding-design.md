# Task 10：主状态所选当前目标的条件性策略绑定设计

日期：2026-10-10

状态：书面设计及其五任务实施计划均已由用户于 2026-10-10 批准；五项 Native 本机实施、十五项限定验证、一次只读增量审查、文档及最终冻结核对均已完成。
审查 0 Critical / 0 Important / 0 Minor；无源码修正。Git 集成与整体 C12/main 验收仍未获本轮授权。
下文 §11–12 的“本轮”描述原设计阶段边界；后续实施证据与限制见 [执行记录](../../roadmap/2026-10-10-main-state-selected-policy-binding-execution.md)，不能把局部模型通过升级为生产验收。

## 1. 目的、当前基础及成功含义

在已完成的主状态投影绑定之上，将同一份所选材料、旧策略高水位、
密钥历史和显式时间观测核对起来，减少“分别通过，但不是同一个输入”
的组合缺口。交付供后续完整主状态组合使用的隔离模型，不接入运行时。

现有基础是：

- `statecontract` 检查十二域声明及有限转换，不认证持久状态。
- `statematerials` 检查完整输入表的有限材料子集，不自行验签。
- `statecomposition` 核验一个显式目标的锚→根链→metadata→desired。
- `statebinding` 绑定上述结果与旧、新主状态投影，允许目标使用历史
  授权根/metadata，保留完整输入重检及未完成责任。
- `statepolicy` 对一份 Batch 及外部模型观测判定高水位、时间和密钥历史；
  它的候选仍是部分状态，不是完整主状态或提交授权。

本阶段成功仅表示：在显式模型假设下，所选**当前**目标与旧策略投影、
时间提供者和策略判定属于同一自有输入快照。不是可信平台证明、全域
策略通过、可以提交或可以运行；不改变 C12/main 未整体验收的状态。

依据：

- [主状态契约设计](2026-10-10-task10-main-state-contract-design.md)。
- [投影绑定设计](2026-10-10-main-state-projection-binding-design.md)。
- [节点信任策略设计](2026-10-09-task10-node-trust-policy-design.md)。

## 2. 方案及本轮边界

采用新增纯组合包 `internal/nodecontrol/statepolicybinding`。组合现有
`statebinding.Evaluate` 与 `statepolicy.EvaluateModel`，不改变下层算法、
接口、历史目标支持、wire 格式或既有证明等级。

备选方案是直接扩展 statebinding，或同时接入 statetime 时间证明。
前者会把历史静态投影与当前策略混在一个入口；后者还需定义 nonce、
checkpoint、证明 metadata 与当前 metadata 的关系及消费规则。选择
独立包先完成当前目标策略绑定，成本是重复的有界快照、解析和验签，
以及明确拒绝本包范围以外的合法历史目标。

本包只接受：

1. 已初始化的 `Policy.Old != nil`，部署观测 `FreshInstall == false`。
2. 旧、新主状态 Desired 均为 Present；不适配首次建立 Desired。
3. 同节点、同 authority epoch、既有 L0 允许的同部署更新。
4. 所选 metadata 完整引用等于新主状态的当前 metadata；所选链尾
   RootSet 完整引用等于新主状态的当前 root；所选 Desired 等于新主状态
   Desired，原关联声明仍必须一致。
5. 主状态 Scope.TimeFloor 原值不变，包括 Present 标记及 UnixSeconds。

通过下层 Binding 的合法历史目标、未初始化场景、首次 Desired 或主状态
floor 变更，在本包返回范围不支持；下层仍先拒绝结构错误、floor 清除/
降低等回退，不将这些失败改为范围诊断。不换材料、不延长目标链、
不改写声明、不自动重试。
现有 statebinding 仍可单独处理其既有历史模型，本包的拒绝不回写下层。

不纳入：statetime、时间证明/nonce/checkpoint 消费、TLS、平台观察器实现、
完整历史认证、其他安全域验证器、完整主状态编码、物理提交、运行授权、
真实服务/数据库/Docker、全仓或 C12 验收、Git 集成和清理。

## 3. 对外边界及输出

公开输入仅包含完整下层输入，不接受外来成功标记、缓存证据或候选：

```go
type Input struct {
    Binding statebinding.Input
    Policy  statepolicy.ModelInput
}

func Evaluate(ctx context.Context, in Input) (*Model, error)
func Recheck(ctx context.Context, model *Model, in Input) error
func (model *Model) Summary() Summary
func (model *Model) Clear()
```

Model 私有拥有完整规范化输入及本次独立摘要；不得浅复制，不支持并发
访问/Clear。不给出 State、Commit、Authorize 或 Ready 方法，不导出下层
Candidate 或把部分 StateModel 作为完整主状态输出。

本层状态固定为零值 InvalidResult 和 ConditionalSelectedPolicyBound。
Summary 包含：

- 本层状态，以及原样保留的 `statebinding.Summary` 与其所有嵌套责任。
- 本次保留并完整重检的旧策略 StateID、CounterID、StateEpoch、部署来源及观测 ID；
  明确它们只是调用者提供的身份，不是主状态物理提交身份。
- 旧、新当前 RootSet 的完整引用、版本/sequence，及对应初始锚或轮换
  证明对象的完整身份；不把两种 payload 摘要合并为一个字段。
- 旧、新 metadata/desired 高水位，历史/密钥/撤销集合数量。
- TimeProviderID、时间 ObservationID、显式 Now、旧策略 floor、观测 floor、
  原样保留的可选主状态 floor，以及本次下层 ValidUntil。
- `SelectedPolicyNoChange`：只转述下层部分策略的 NoChange，不表示十二域
  主状态不变，不建议是否推进主 guard。
- 本层固定三项附加责任：ExplicitObservationsAndBaseProvenance、
  TimeAttestationAndFreshness、OtherDomainPolicyCoverage。

Summary 所含 Links/Coverage 等可变集合每次均独立复制，调用者可清理自己的
副本；返回前发生 panic 时清理已建立的自有副本。nil/零值/已清理模型返回
完整零值摘要。摘要不携带 raw 材料、完整历史账本或隐含生产 capability。

## 4. 自有快照、预算与规范化

预取消/超时优先，nil context 拒绝。任何深复制或索引分配前，检查全部
Binding 预算，**完全沿用投影绑定设计第 4 节的数值和所有分支**，不得先
只检查所选项。另检查以下 Policy 输入：

| 范围 | 预复制上限 |
| --- | --- |
| Batch.Rotations | 32 项 |
| Batch 每次出现的 anchor、rotation、active root、metadata、desired raw | 各 262,144 bytes |
| Batch raw 总量，按出现次数计，包括重复的 active root | 36 × 262,144 bytes |
| Binding 加 Batch 外来 raw 总量 | 32 MiB + 37 × 262,144 bytes |
| Old.Roots / 每个 root 的 Keys | 33 / 5 项 |
| Old.Keys / Old.RevokedIDs / History | 各 4096 项 |
| 每个 History.IncidentKind | 64 bytes |
| 每个 Old key 的 Status / RevocationReason | 8 / 32 bytes |

预算接受等长输入不等于语义接受。例如既有 policy 仍可能拒绝 4096 项
revoked ledger；本包不扩大该算法范围。累计用 `len > limit-total` 形式
避免加法溢出；不按指针别名去重，不忽略未使用的输入。预算不是峰值内存
或性能承诺，下层快照和临时解析可能同时存在。

深复制 Binding 全部字段/集合/raw、Policy 全部 Batch raw 和二维切片、
Old 的 roots/public keys、keys、revoked、History。固定数组和值字段逐项
复制；输入调用期间不可并发修改。只规范化：

- Binding 两表的顶层 Materials/Expected 按 Kind/P 排序；空集合为 nil。
- Policy 空集合为 nil；roots、keys、revoked、History、rotations 的原顺序
  保留，不排序掩盖非法顺序，也不将不同 History 排列当作同一重检输入。
- 全部 time.Time 去掉单调时钟表示并转 UTC，保留纳秒和零值语义；不取整、
  不生成当前时间、不补缺失时间。主状态 UnixSeconds 不变。

复制中定期检查取消，每份新增可变数据立即挂到清理责任下。完整拥有
输入后，在调用下一次 context 回调或下层前，先登记失败/panic 清理。
下层 owner、State() 的临时独立副本、解析 owner 和摘出的 material 也在
返回后立即登记清理；摘要先交给已受保护的本层 owner，再做后续检查。

只有最终 context 检查通过后才转移 Model 所有权。普通失败/取消返回 nil
模型；panic 经 defer 清理后传播，不 recover 成成功。Clear nil-safe、
幂等，清零可达自有 raw/集合及摘要后将模型置零，不修改调用者输入或已
交出的独立摘要。不承诺 OOM/致命终止、不可达或运行时副本、字符串和
全进程内存擦除，也不声称对所有任意控制流有形式化证明。

## 5. 同快照执行顺序

1. 检查 context/全部预算，建立自有输入并登记失败清理。
2. 对完整自有 Binding 调用现有 statebinding.Evaluate，立即登记 Clear；
   捕获并保护完整摘要，核对 ConditionalProjectionBound 和摘要坐标。
3. 检查第 2 节的支持范围，再检查第 6 节 Batch 原字节绑定。
4. 按第 7–8 节绑定旧策略投影与时间提供者/floor。读取原始格式事实
   只使用 stateformat 既有 parser，不另写 JSON/JCS/digest/签名算法。
5. 对同一自有 Policy 调用 statepolicy.EvaluateModel，立即登记 Clear；
   获取临时 State() 副本并立即登记其 Clear，按第 7 节核对新投影。
6. 将独立摘要放入已保护的 Model，核对 SelectedPolicyNoChange 和
   ValidUntil；最终复查 context，才返回条件性结果。

没有重试、provider 回调、外部 I/O、环境/系统时钟读取、后台任务或
goroutine。所有下层完整输入检查继续运行，不以摘要相等替代判定。

## 6. 精确 Batch 绑定

成功的下层 Binding 不允许 Policy 使用另一份单独有效的材料。
在已检查的 Proposed 表内，以完整 Kind/P/E 唯一命中每个所选引用，
将 Policy.Batch 的以下字段与同快照原字节逐项比较：

| Policy.Batch 字段 | 唯一来源及比较 |
| --- | --- |
| ExpectedNodeID / ExpectedEpoch | Proposed.Materials.Scope，及 Main.Proposed.Scope 的 node/authority epoch |
| AnchorEnvelope | Proposed.Anchor.Raw，精确 bytes 相等 |
| Rotations[i] | Proposed.Target.Links[i].Rotation 在 Proposed 表的 Raw；数量和顺序相同 |
| ActiveRootEnvelope | 无轮换时为同一 Anchor.Raw；否则为最后一项所选 rotation.Raw |
| MetadataEnvelope / DesiredEnvelope | 所选 Target.Metadata / Target.Desired 在 Proposed 表的 Raw |

不得仅比 payload、以有效替代信封替换、跨 Previous 表取字节或把 RootSet
payload 当成 ActiveRootEnvelope。Binding 已检查独立锚 P/E 及原 root payload
关系；Policy 再沿用既有完整验签路径，不接受本层伪造 SignatureEvidence。

未选项仍由 Binding 完整核对、拥有和重检。对 Batch 的任何有效替换须
作为新 Evaluate 处理，不能沿用旧 Model。

## 7. 策略投影：明确分开 root-set 与 rotation 身份

### 7.1 节点、锚及旧策略身份

Old.NodeID、Deployment.NodeID、Batch 节点及两份 Main.Scope.NodeID 必须
一致；对应 epoch 相等。Old.AnchorPayload 与 Deployment.AnchorPayload
绑定 Proposed.Anchor.Payload，部署 SourceID 和时间 ProviderID 的连续性
仍交现有 policy 检查。

Old.StateID、CounterID、StateEpoch 全部保留并重检，遵守下层非零/范围
规则。不把 Old.StateEpoch 当作 checkpoint sequence，不把 Policy.CounterID
直接当作主/监督/latch guard ID。不规定 Main.DeploymentID 等于部署
SourceID，也不规定 TrustAnchorID 等于锚 P/E。现有类型没有这些转换契约。
本包不证明 Old 与 Main.Previous 来自同一物理快照；该缺口明确归于
ExplicitObservationsAndBaseProvenance 和原 FullStateComposition 责任。

### 7.2 根链投影

建立一个仅用于本次核对的有序映射：第 0 项为独立初始锚及 Target.AnchorRoot；
后续第 i 项为 Target.Links[i-1] 的 Rotation 和 NewRoot。每项 RootSet 均在
Proposed 表内完整命中，并使用现有 ParseRootSet 读取版本、sequence、
threshold 和完整 public-key 材料。

| 根项 | statepolicy.RootRecord.Mark.Payload | 主状态 / RootSet 身份 |
| --- | --- | --- |
| v1 初始根 | 独立初始锚的 payload 摘要 | Target.AnchorRoot.Payload，与初始 root payload 一致 |
| 第 i 次轮换产生的根 | 对应 Rotation.Payload，**不是** NewRoot.Payload | 对应 NewRoot.Payload，即嵌入并独立列出的 root-set payload |

Old.Roots 必须是所选完整链的非空前缀。每项 Mark 的 version/sequence/
证明对象 payload 精确相符，RootMaterial 的版本、threshold、所有有序
key ID/public key 字节与对应 RootSet 完全相符；不只比较数量或摘要。
下层 policy 再将旧前缀与验签事实精确核对，不更改它保存的 Mark 含义。

旧前缀最后一个 **RootSet** 完整引用、版本及 sequence 绑定 Main.Previous.Root；
其材料也在 Previous 表被 Binding 检查。完整链最后一个 RootSet 绑定
Main.Proposed.Root。初始根和轮换所含 root-set 不能互相当 envelope 使用。

`Main.RootTransition` 须等于目标链从旧根之后开始的有序后缀投影：
逐项核对 rotation 完整 P/E、old/new root payload、前后版本及 sequence。
旧根未变化时后缀和 Main.RootTransition 均为空。该相等规则只属于本包
的“所选当前根”子范围，不修改 statebinding 历史目标中两条链不同的规则。

新策略 State() 的 Roots 与上述完整映射逐项核对，最后根再次绑定新 Main；
其 Status 必须仍是 Unknown。StateID/CounterID/StateEpoch 和旧策略 floor
按下层现有语义保留，绝不据此宣称发生新物理提交。

### 7.3 metadata、desired、keys 和 ledger

Old.Metadata / Old.Desired 的 version、sequence、payload 分别精确等于
Main.Previous.Metadata / Desired 的对应投影；策略没有 envelope 字段时，
不填造 E，原完整 E 仍由 Binding 拥有和核对。旧 Desired 可以原本使用
历史 authorizer；不把它改为当前 metadata，也不要求旧材料今天仍有效。

用 Previous 表中**旧当前 metadata 的精确原信封**读取 MetadataMaterial：
每个当前 key 必须在 Old.Keys 中唯一出现，完整 MetadataKeyMaterial 相同，
包括 public key、状态、有效窗及撤销/incident 字段；Old.RevokedIDs 精确
等于旧当前 metadata 的完整有序 revoked ledger。不能用 key ID/数量
替代材料一致性。

Old.Keys 可包含该 snapshot 已不再列出的历史 key/tombstone，不能强制
整个 Keys 等于当前 KeyIDs，也不能删除多出的历史记录。其 FirstRevokedAt、
DeadlinesKnown、最后 snapshot/composite deadline 等保持原值，交现有
policy 校验，并进入完整重检；来源真实性及历史完整性仍未认证。

新 State() 的 metadata/desired 高水位精确绑定 Main.Proposed，使用 Proposed
表中当前 metadata 再作同样的“当前 keys 精确材料 / 完整 ledger”检查。
保留 policy 产生的额外历史记录，不把它们回填 Main.Metadata.KeyIDs。
新身份的 node/epoch/anchor、部署来源和时间提供者同时核对，不只比业务
高水位。History 的 BaseStateID、MetadataPayload 以及每项条件检查全部
沿用下层；不新增默认历史、不绕过紧急撤销诊断。

旧 metadata/desired 的签名及其他历史项不会因此获得 SignatureChecked。
Metadata.History、RevokedSetDigest 仍无完整对应 codec/历史证明。本轮
关联声明仍不是新增签入 wire 的关系；独立一致改写后可形成另一个条件性
输入，不能称其原关联已经获得认证。旧 Model 重检必须拒绝这种变化。

## 8. 三种 floor 的明确关系

同一 TimeProviderID 精确绑定 Main 旧/新 Scope、Policy.Old 和 Time.ProviderID；
Available/ObservationID/回退状态及时间规则均沿用 policy，不读取系统时钟。

| 值 | 本包的含义与规则 |
| --- | --- |
| Main.Scope.TimeFloor | 可选、秒精度的主状态已有安全约束；本包只保留，旧/新完全相同 |
| Policy.Old.TimeFloor | 旧策略显式模型的必需时间下限，非零且在下层有效范围 |
| Policy.Time.Floor / Now | 本次显式 provider 观测；Floor ≥ Old.TimeFloor，Now ≥ Floor |

Main floor 为 Present 时，安全转换 UnixSeconds 为 UTC 瞬间并检查范围；
要求 Old.TimeFloor ≥ Main floor。三者不是相等字段；因此观测 floor
可以更高，但不会提高主状态 floor，也不把新 State().TimeFloor 自动
改为观测 floor。Main floor 缺席时，不合成零时间或新的主状态 floor，
仍要求 policy 的旧 floor 和本次观测有效。

本包既不允许主状态 floor 提高，也不允许清除/降低/新建。不同 floor
推进、持久化和主 guard 关系另行设计。签名材料、时间证明、HTTP Date
不产生时间观测，ValidUntil 也不能作为延长授权的期限。

既有 metadata、desired、key、轮换半开时间窗和 checked 时间运算保持不变。
新接受轮换后缀需当前有效；已接受前缀仍核验签名但不因今天过期拒绝。
重复业务状态也必须经过当前显式时间规则，不能用 NoChange 跳过过期。

## 9. 重检、错误及未完成责任

Recheck 处理 nil/预取消 context、缺失/清理模型后，重新执行完整 Evaluate；
fresh owner 返回后立即登记 Clear，再精确比较全部规范化 Input，最后
复查取消。覆盖 Binding 十二域、全部表项、未选项、目标及锚，Policy
所有原字节、Old 全字段、部署/时间/History 全字段，不只比较摘要。

本轮重检**固定时间快照**：Now、Floor、ObservationID 的真实改变即使
向前且仍合法，也不属于旧 Model，返回绑定不符；需新的 Evaluate。
仅时区/单调表示变化按第 4 节规范化为同一瞬间。下层 policy 允许时间
向前重检的 API 不改，但本包不借此放宽输入身份。Recheck 不证明当前
真实时间未到期，不创建 freshness capability 或原子提交锁。

本层错误采用固定分类：InvalidInput、LimitExceeded、IncompleteCoverage、
InvalidMaterial、MissingMaterial、DuplicateMaterial、BindingMismatch、
SignatureMismatch、Rollback、Conflict、ClearedModel、UnsupportedPolicyShape、
PolicyRejected。错误文字固定以 `statepolicybinding: ` 开头，不含输入
值、UUID、摘要、raw、路径或秘密。

Binding 已知类别同义映射；statepolicy.PolicyError 保留 Code/Reason
作为 PolicyRejected 的只读枚举诊断，不把 AssuranceMissing、SecurityFault
或 EmergencyRevocation 改成普通成功。未知非 nil 错误映射 InvalidInput。
当前 ctx.Err 优先，并保留 errors.Is(Canceled/DeadlineExceeded)。若当前
ctx.Err 为 nil，但下层仍报告 Canceled/Cancellation，则返回
context.Canceled，不尝试恢复或重新调用；取消错误不要求本层文字前缀。

确定顺序为第 5 节阶段顺序，同阶段按保留的集合顺序检查，不依赖 map
遍历。预算超限先于复制；下层合法历史目标的本层范围拒绝先于策略高水位
比较，不能错说它已经发生真实存储回滚。全失败路径无部分输出、不改输入。

原 Binding 的全部八项及所有下层嵌套责任原样保留。DomainAndCurrentTime
不删除：这里新增的时间判定只覆盖所选当前策略，不认证时钟或其他域。
FullStateComposition、PhysicalCommit 和 RuntimeAuthorization 仍未完成。
附加三项责任固定保留，不因任何模型值改为“已满足”。

## 10. 隔离及未来允许修改的范围

未来实现仅新增 statepolicybinding 产品/测试文件，并精准更新以下六个
既有消费者边界测试，以允许**准确目录且包名为 statepolicybinding**：

- statebinding/boundary_test.go：唯一新增条件性消费者。
- statepolicy/boundary_test.go：唯一新增条件性消费者。
- stateverify/boundary_test.go：允许该包消费 statepolicy；不允许本包直接
  调用 Verify，也不改变原 policy/composition 的 verifier 权限。
- statecontract/boundary_test.go、statematerials/boundary_test.go、
  statecomposition/boundary_test.go：只为本包的类型、复制、摘要/引用核对
  放行；本包不直接调用它们的 Evaluate。

新增目录匹配不能用前缀、通配或同名任意目录；反例包括子目录、相似前缀、
目录对但包名错、runtime/reader/serving/authority/store 等原禁用消费者。
本包自身仍禁止任何非测试产品消费者。

本包采用精确 import/成员能力白名单：statebinding、statepolicy 是唯一
判定入口；stateformat 只作既有格式解析及 material 访问，其他下层包只
用限定类型/常量。纯标准库只用于 context、复制、排序、值比较、固定错误
及显式时间转换。不导入 stateverify、statetime、staterecovery、数据库、
网络、文件/进程、硬件、随机或环境能力；禁止系统时钟、provider 调用、
goroutine、dot/blank/重复 import 及动态反射操作。AST 守卫是固定源码
形状与能力回归检查，不是任意 Go 程序安全性的形式化证明。

六份既有测试只变更上述消费者判断及精确正反例，不放宽产品能力白名单
或异常清理规则。下层所有产品源码、其他既有测试、依赖、SQL 和生成文件
保持不变。如实现发现必须改变它们，停止并单独提出范围修订。

## 11. 未来测试及局部验收

实施计划需以测试先行覆盖：

1. 当前目标的 no-change、metadata/desired 前进、0/1/32 轮换及旧前缀。
2. 合法历史目标在 statebinding 仍成立，但本包范围拒绝；不改写/降级。
3. Batch 每个字段、每个 raw、链数量/顺序、active root、完整 E 错配；
   另一份独立有效 Batch 也不能通过旧绑定。
4. v1 身份、轮换 P 与 root-set P 分离；前缀根材料/sequence/threshold/
   public-key 替换、旧当前根位置、RootTransition 后缀错配。
5. 旧、新 metadata/desired 高水位，完整当前 key 材料及 ledger；合法
   额外历史记录保留，历史冲突/缺证、早删 tombstone、紧急撤销不成功。
6. 节点/epoch/anchor、provider 及观测身份错配；Old 缺失、FreshInstall、
   旧 Desired 缺席不进入首次建立路径；不得借测试发明 opaque ID 等式。
7. floor 缺席/存在、相等/更高观测、低于约束、纳秒、时间范围及半开
   窗口；所有主状态 floor 变更拒绝，无 wall-clock fallback。
8. Summary 的条件状态及全部责任、独立集合、清理后零值；其他主状态
   域变化不能被 SelectedPolicyNoChange 误称为完整主状态无变化。
9. 全输入重检，包括合法未选项、未投影域、历史字段、观测 ID、时间
   向前变化；两表排列/空集合/等瞬时时区仅按既定规范化处理。
10. 各集合/原字节预算等长与超限、复制隔离、取消及正常 panic 清理、
    无部分 owner 返回；源码清理登记检查及可达 backing-store 清零检查
    分别记录，不假称直接观测到了不可返回的私有副本。
11. 六份旧守卫及新包守卫的精确正反例；不增加产品接入或额外能力。

未来验证使用固定签名样例、显式时间及测试模型；仅离线隔离新包测试与
既有定向回归：statebinding、statecontract、statematerials、statecomposition、
stateformat、stateverify、statepolicy、staterecovery、statetime、serving、
contracts、readiness，authority 仅 TestBoundRead/TestCoordinator，store
仅编译。精确命令、超时和证据冻结在书面计划中确定，本轮不运行测试。

不扩大到 race、全仓、真实 DB/Docker、C11/C12、bootstrap/PITR 或硬件测试。
局部通过仍保留平台来源、跨启动连续性、全域材料/策略、物理耐久性、
独立运行时授权及整体 main 验收阻塞。实施后的源码复核与局部测试不能
被表述为本轮设计已经获得独立安全审查或生产验收。

## 12. 本轮交付、审批及自检

本轮只新增本文，不改源码、既有路线图完成记录或实施计划，不安装依赖、
运行代码测试、暂存、提交、推送、合并或清理。brainstorming 的默认提交
步骤受已确认的本轮“仅文档、不提交”边界约束；本文目前只有本机副本，
尚未形成跨电脑恢复检查点。

用户批准书面设计后，下一步仅使用 writing-plans 编写书面实施计划；
计划审阅批准后才进入实现。沿用先前 Native 执行选择，不重复要求选择
方法，仍以届时明确批准的计划为准。本轮自检不是独立源码/安全审查。

自检已核对：方向与目的、当前/历史目标隔离、rotation/root-set 不同身份、
三种 floor、旧当前 keys 与历史账本区分、opaque ID 无转换推断、完整重检、
所有权/异常清理、六份精确测试例外、局部验证范围及审批边界。无未决定
占位项；不宣称下层责任已消失或全项目完成。
