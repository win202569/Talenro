# Task 10：主状态当前目标的条件性时间证据绑定设计

日期：2026-10-11

状态：用户于 2026-10-11 批准本文及后续五任务实施计划；限定本机实现、
修正后十六项验证和同一只读 reviewer 复查完成。首次 Important 分类缺陷
已修复，早期记录缺项作为 M1 审计限制保留；无 Git 集成或 C12/main 验收。
见[执行记录](../../roadmap/2026-10-11-main-state-time-evidence-binding-execution.md)。
以下“未来/本轮”措辞保留书面设计阶段的历史语境，不代替执行记录。

## 1. 目的、基础与成功含义

在已完成的 `statepolicybinding` 之上，将同一份当前授权材料、旧策略
快照、显式时间、请求 nonce、旧检查点及签名的新检查点绑定起来，关闭
“各模块分别通过，却使用不同输入”的条件性组合缺口。交付仍是隔离的
纯模型，不是可信时钟、请求消费、物理提交或运行授权。

现有基础：

- `statepolicybinding.Evaluate` 已绑定完整主状态投影、所选当前目标、
  Batch 原字节、旧策略高水位/完整密钥历史和显式时间。
- `statetime.EvaluateModel` 核验调用者指定的当前 metadata 下的时间
  证明签名、请求 nonce、节点/epoch、时间窗和检查点；不会认证 metadata
  根链、实际 TLS/时钟/存储来源，也不会消费 nonce。
- 两者的观察身份仍来自外部模型。把条件模型连起来不能补成平台证明。

成功仅表示：在一份自有、完整、固定观测快照中，两次完整下层判定及
本层对应关系都成立。C12/main 的整体验收、跨启动连续性、真实观察器、
完整主状态认证和物理耐久性仍未完成。

设计依据：

- [所选当前目标策略绑定设计](2026-10-10-main-state-selected-policy-binding-design.md)。
- [时间证明设计](2026-10-09-task10-time-attestation-design.md)。
- [主状态投影绑定设计](2026-10-10-main-state-projection-binding-design.md)。

## 2. 方案、适用范围与关键取舍

采用独立包 `internal/nodecontrol/statetimebinding`，只组合
`statepolicybinding.Evaluate` 和 `statetime.EvaluateModel`。不改下层产品
源码、接口、判定顺序、wire 格式或既有证明等级。

备选是扩展现有策略绑定包；这会把尚未解决的请求/检查点责任并入已冻结
接口。独立层便于明确范围和保留下层责任，成本是重复的有界复制及判定。

**本次仅支持当前授权根与 metadata 均不变化的场景。** statetime 使用的
是调用者指定的已提交当前 metadata，不能把尚未提交的候选当作来源。
因此本层要求规范化后的 `Main.Previous.Root` 与 `Main.Proposed.Root` 全
字段相同，两个 `Metadata` section 也全字段相同：包含 header、完整
Kind/P/E、版本/sequence、threshold、所有有序 key/revoked 集合、根关联、
history、set digest 和 window，而不只比较 payload 或版本。

同时完整继承 selected-policy 层的范围：Old 非 nil、非 FreshInstall、
旧/新 Desired 为 Present、同节点/authority epoch/既有部署、所选目标是
新主状态当前 root/metadata/desired、主状态可选 floor 完全不变。
在 root/metadata 不变的本层范围中，所选 root/metadata 也必须是旧主状态
当前项；`RootTransition` 为空，旧策略根链是完整目标链，不能有新轮换
后缀。已有接受链可含 0、1 或 32 次历史轮换，不因其存在而拒绝。

允许下层原规则所允许的 Desired 更新或不变、检查点等值或前进，以及
其他 L0 声明更新；这不补成其他域策略判定，也不把 SelectedPolicyNoChange
解释为十二域主状态不变。

合法的 root/metadata 安装、历史目标、首次安装/Desired 建立和主状态
floor 变化不属于本层。先执行下层完整策略绑定，再作本层范围诊断。
下层错误原样分类，不能把回退、验签失败或紧急撤销拒绝改成范围提示。
已有下层独立支持不变；需要随响应安装新 root/metadata 的时间证明，
必须以后另行设计，不能在本包将候选 Status 改为 Available 或 committed。

不纳入：真实时钟/TLS/provider、随机 nonce 生成、消费/防重放存储、
检查点预留/持久化/guard seal、跨启动历史认证、运行时接入、完整主状态
组合、数据库/Docker/硬件测试、整体 C12 验收、提交推送合并或清理。

## 3. 输入、输出与条件责任

不接受缓存成功标记、外来下层 owner、已解析材料或“可提交”证明。

```go
type Input struct {
    Policy statepolicybinding.Input
    Time   statetime.ModelInput
}

func Evaluate(ctx context.Context, in Input) (*Model, error)
func Recheck(ctx context.Context, model *Model, in Input) error
func (model *Model) Summary() Summary
func (model *Model) Clear()
```

Model 私有拥有完整规范化 Input 和独立 Summary；不能浅复制或与 Clear
并发使用。不提供 State、Commit、Authorize、Ready 或 nonce 消费接口，
不返回下层 Candidate/EvidenceModel。结果状态只有零值 InvalidResult 和
ConditionalTimeEvidenceBound。
ResultStatus 和 Obligation 均为 uint8 闭合枚举；Obligation 的零值为
UnknownObligation，只有有效摘要才填写五项固定责任。

Summary 精确包含：

- `Status ResultStatus`：只表达本层条件性结果。
- `Policy statepolicybinding.Summary`：原样保留完整嵌套摘要及责任。
- `Time statetime.EvidenceSummary`：本次真实下层判定的完整固定值摘要。
- `Observations ObservationBinding`：BaseStateID，MetadataSourceID、
  MetadataObservationID、RequestObservationID、CheckpointSourceID、
  CheckpointObservationID；类型均为既有 contracts.Digest。其余身份已在
  Policy.Base、Policy.Time 和 Time.RequestID 中保留，不再合成新身份。
- `ValidUntil`：严格为 Policy.ValidUntil 与 Time.ValidUntil 的较早瞬间。
- 固定 `[5]Obligation`：ObservationAndBaseProvenance、TrustedTimeAndTLS、
  AtomicRequestConsumption、DurableCheckpointCommit、
  FullMainStateAndRuntimeAuthorization。

五项责任及所有嵌套责任始终保留。尤其不删除原
Policy.Obligations 中的 TimeAttestationAndFreshness：本层覆盖签名与模型对应关系，不覆盖
真实来源、请求原子消费或实际新鲜性。不能用成功标记把原责任改成已满足。
不新增 ReplayProtected、NonceConsumed、TLSValidated、Committed 等布尔值。

Summary 不含 raw、完整账本、provider 句柄或隐含生产 capability；所有
Links/Coverage 等可变集合每次独立复制，调用者可清理自己的副本。
nil/零值/已清理 Model 返回完整零值摘要。ValidUntil 不是授权延期、实际
当前时间或提交时效保证。Time.AdvanceRequired 不意味着真的推进了存储。

## 4. 完整所有权、预算与规范化

nil context 拒绝；预取消/超时优先。任何输入深复制或索引分配之前，
检查完整 Policy 和 Time 的所有分支，不能只看所选材料。沿用既有
selected-policy 与 projection-binding 数值预算，明确为：

| 范围 | 预复制上限 |
| --- | --- |
| 每个 Binding 表的 Materials / Expected | 各 128 项 |
| 每表 Materials / Expected 的 RootRotation；Target.Links / RootTransition | 各 32 项 |
| 表中每个 RootSet raw / 其他 raw；独立 Anchor.Raw | 65,536 / 262,144；262,144 bytes |
| 每表 raw；完整 Binding raw 含独立锚 | 16 MiB；32 MiB + 262,144 bytes |
| 每个 Expectation 的 root keys / metadata keys / revoked / desired slots | 5 / 64 / 4096 / 8 项 |
| 两份 Main 各自 root keys / metadata keys / revoked | 5 / 64 / 4096 项 |
| 两份 Main 各自 bootstrap / agent deauthorized | 各 128 项 |
| 两份 Main 各自 Desired slots / LKG entries / local faults / supervisor faults | 8 / 8 / 64 / 16 项 |
| Policy.Batch.Rotations；Batch 每次出现的 raw | 32 项；262,144 bytes |
| Policy.Batch raw 总量；完整 Policy Input raw | 36 × 262,144；32 MiB + 37 × 262,144 bytes |
| Policy.Old.Roots / 每个 root public keys | 33 / 5 项 |
| Policy.Old.Keys / Old.RevokedIDs / History | 各 4096 项 |
| 每个旧 key 的 Status / RevocationReason；每个 History.IncidentKind | 8 / 32；64 bytes |
| Time.AttestationEnvelope / Time.MetadataEnvelope | 各 262,144 bytes |
| 两个 Time raw 合计；完整本层 Input raw | 524,288；32 MiB + 39 × 262,144 bytes |

按出现次数计 raw，包括重复的 metadata/active root 和跨输入别名，不按
指针去重。累计采用 `len > limit-total`，避免加法溢出；等长预算通过不
表示下层语义通过。不增加 v1 锚语义、共享 wire 限制或最大轮换数。
预算不是峰值内存、性能或抗所有资源耗尽攻击的承诺。

完整深复制 Policy 的 Binding 全部十二域/表项/期望/目标/锚/raw、Batch
所有 raw/二维 rotations、Old 全字段及 roots/public keys/keys/revoked、
完整 History；另外复制 Time 两个 raw 和全部观测值。即使同一 backing
store 被多个位置共享，每个自有位置也独立复制。不得为省内存丢弃历史
key/tombstone、未选材料或未投影域。固定数组、UUID、nonce 按值保留。

只作既有规范化：两张 Binding 表顶层 Materials/Expected 按 Kind/P 排序；
空集合为 nil；其他集合顺序不变；全部 time.Time 用 Round(0).UTC() 保留
纳秒，零值仍为真正零值。主状态 UnixSeconds 和所有 raw 字节不变。不
重编码材料，不生成 Now，不补缺失时间，不用排序掩盖非法历史顺序。

复制阶段每份可变数据立即归属失败/panic 清理责任；成功取得整个 Input
后，在下一次 context 回调或下层调用之前登记本层 owner 清理。取得下层
owner 后立即登记其 Clear，先于错误映射/context 回调；获取 Policy.Summary
时直接放入已保护的本层字段，不先建立未保护的可变临时摘要。

仅最终 context 检查通过后才设置 valid 并转移所有权。错误/取消返回 nil
模型，无部分输出；panic 清理后传播，不 recover 为成功。Clear nil-safe、
幂等，清零可达自有 raw、集合及摘要并将 owner 置零；不破坏调用者输入
或已交出的独立 Summary。不承诺字符串/不可达副本/运行时副本擦除、
OOM/致命终止恢复、并发安全或任意控制流的形式化安全证明。

## 5. 同快照执行顺序

1. context、完整预算检查，建立完整自有 Input 并保护本层 owner。
2. 调用 `statepolicybinding.Evaluate(ctx, owned.Policy)`，立即登记 Clear；
   直接捕获保护 Policy.Summary，检查 ConditionalSelectedPolicyBound 及
   第 6 节所用摘要坐标与自有输入一致。
3. 核对第 2 节 root/metadata 不变范围及第 6–7 节的材料、共同基态、
   节点/epoch、时间快照和旧检查点关联。
4. 对完整自有 Time 调用 `statetime.EvaluateModel`，立即登记 Clear；
   将固定值摘要放入受保护字段。
5. 核对时间摘要的 metadata P/E、node/epoch、RequestID/nonce、旧/新
   检查点、AdvanceRequired，建立 ObservationBinding、较早 ValidUntil
   及固定责任；检查两个下层的非零有效期限。
6. 最终检查 context，返回 ConditionalTimeEvidenceBound。

不直接调用其他下层 Evaluate/Verify/RecheckModel，不读取 Candidate.State，
不重新实现 parser、签名、digest 或时间证明算法。没有自动重试、替换
材料、provider 回调、环境/系统时钟、网络、存储、随机、后台任务或
goroutine。所有下层完整检查仍执行，不以摘要相等替代算法判定。

## 6. 当前 metadata、节点与共同基态绑定

令本层 `in.Policy.Binding.Main` 为 Main，`in.Policy.Policy` 为现有
statepolicy.ModelInput；下文的 `Policy.Policy` 始终指这一内层完整输入，
不把它与返回的 `Summary.Policy` 混作同一种模型。
摘要中的 Previous/Proposed.Metadata 高水位必须对应这份 Main 的相应
版本/sequence/payload；Time.Metadata 的 version/epoch/P/E 则精确对应
两个未变化的当前 metadata mark 和两份 Main.Scope。节点 UUID 的规范
String 必须等于两份 Main.Scope.NodeID；epoch 必须等于两份 authority
epoch，且 Request/Checkpoint 的 UUID 与 epoch 相同。不得直接强制转换
跨包状态枚举；Available 关系用明确分支核对。
本层核对发现 Time 四种观测中任一 status 不是 Available 时，返回
AssuranceMissing，不把缺证诊断改成普通绑定差异；其余 statetime 格式、
非零身份、key、签名及窗口检查仍完整执行。

`Time.MetadataEnvelope` 必须与 `Policy.Policy.Batch.MetadataEnvelope`
精确 bytes 相等。该 Batch 已由下层绑定所选当前目标的原信封；本层还
在 Previous 和 Proposed 表中，以相同完整 Kind/P/E 唯一命中当前 metadata，
要求两处 raw 均与 Time.MetadataEnvelope 相同。重复 Kind/P、缺失项或
另一份同 payload 不同 envelope 的有效材料不能冒充同快照。

所选 target 的 metadata 与链尾 root 同时对应旧/新当前项；只取已检查
的原字节和固定值，不重建 envelope。metadata 根链及所选策略授权仍由
完整 selected-policy 判定认证，时间签名/key/window 仍由 statetime 判定。
不把这两者升级为旧历史签名、完整关联 wire 认证或存储读取真实性。

**本层新增、供本次书面审阅的共同基态模型约定：**

```text
Time.Metadata.StateID == Time.Checkpoint.StateID == Policy.Policy.Old.StateID
```

三个字段在本组合 Input 中均被定义为同一个调用者声明的已提交旧基态
标识，必须非零且精确相同；此约定不是既有两个独立包已经保证的事实。
不得靠摘要相似推断任意 opaque ID 等价，也不证明它们来自一次物理
读取/锁定/提交。独立、一致替换后仍可成为另一份条件输入，旧 Model
重检必须拒绝这种变化。来源真实性由 ObservationAndBaseProvenance 保留。

Metadata.SourceID、Checkpoint.SourceID、Deployment.SourceID 和主状态
DeploymentID 各保留原有角色，不要求相等；各 ObservationID 也不相互
合并。不将 StateID 解释为材料摘要/封印，不将 StateEpoch 解释为检查点，
不把 CounterID 转换成主/监督/latch guard。所有原身份进入完整重检。

## 7. 显式时间、请求与检查点

### 7.1 同一个时间观测，不提升 floor

Policy.Policy.Time 与 Time.Time 必须都是各包明确的 Available，ProviderID、
ObservationID、规范化 Now 和 Floor 精确相同。ProviderID 同时对应
Policy.Policy.Old 及两份 Main.Scope 的现有 provider 连续性。下层已有关系不变：
主状态可选 floor 不变；存在时旧策略 floor 不低于它；ObservedFloor 不
低于 OldFloor；Now 不低于 ObservedFloor。三种 floor 不强制相等。

签名的 IssuedAt、HTTP Date、ValidUntil 不生成 Now/Floor/provider，也不
提高主状态 floor 或 guard。原 policy 半开窗口、历史前缀规则，以及
statetime 的 metadata/key/attestation 窗口和五分钟 freshness 限制全部
沿用原算法。NoChange 不跳过时间判定。没有缺证时的 wall-clock fallback。

### 7.2 请求：精确 nonce 关联，不宣称防重放

Time.Request 的 status、RequestID、ObservationID、NodeID、epoch 和
ExpectedNonce 全部保留；完整 statetime 判定要求签名 nonce 与
ExpectedNonce 相同。全零 32-byte nonce 是既有格式允许的值，不能把它
当作密钥而新增非零要求，也不能因此宣称它经过随机生成。

RequestID/ObservationID 是显式观测身份，不是新增签入 wire 的字段。
摘要记录 RequestID 不代表它本身获得签名认证。本层不读取网络请求，
不发放 nonce，不验证 TLS，不消费或锁定请求；相同未消费模型重复
Evaluate 可以重复成功，必须明确保留 AtomicRequestConsumption。

### 7.3 检查点：旧观测与新声明精确对应

```text
Time.Checkpoint.Sequence == Main.Previous.Checkpoint.Sequence
Time.Summary.OldCheckpoint == Time.Checkpoint.Sequence
Time.Summary.ProposedCheckpoint == Main.Proposed.Checkpoint.Sequence
Time.Summary.AdvanceRequired == (ProposedCheckpoint > OldCheckpoint)
```

这里 Time.Summary 是本次下层返回的 EvidenceSummary，不是外来输入。
检查点节点/epoch 同时对应两份 Main.Scope 和 Checkpoint section。
签名检查点等值时 Main 检查点不变，签名前进时 Main 新 sequence 必须
恰等于签名值；回退由原 statetime/契约规则拒绝。不采用 max、clamp、
自动补写或选择另一份证明，也不把 artifact 创建 sequence/root version/
metadata version/StateEpoch 当作检查点。允许 Desired 不变但检查点前进，
不能用下层 SelectedPolicyNoChange 抹掉它。

这些关系只绑定 proposed 声明：没有 checkpoint 预留、实际 disk commit、
计数器/封印写入或中断恢复保证。DurableCheckpointCommit 始终保留。

## 8. 固定快照重检与固定错误

Recheck 先处理 nil/预取消 context、nil/清理 Model，再完整重新 Evaluate
新 Input；fresh owner 立即登记 Clear。比较所有规范化 Input 字段，最后
复查 context，不只比较摘要、所选项或有效期限。

覆盖所有十二域、未选表项、全部 raw、旧策略/历史/部署/观测身份和
Time 全观测/nonce。Now/Floor/ObservationID 改变即使向前且仍合法，也
不是旧模型，返回 BindingMismatch，必须新 Evaluate；仅既定时区/单调
表示和顶层表排列/空集合规范化可视为同输入。不调用允许时间向前的
下层 RecheckModel。重检不是真实当前时间检测或原子提交锁。

本层使用 20 个固定 sentinel，错误文字以 `statetimebinding: ` 开头，
不含输入 UUID、摘要、raw、路径、时间值或秘密：

- 原 selected-policy 的 13 类同义保留：InvalidInput、LimitExceeded、
  IncompleteCoverage、InvalidMaterial、MissingMaterial、DuplicateMaterial、
  BindingMismatch、SignatureMismatch、Rollback、Conflict、ClearedModel、
  UnsupportedPolicyShape、PolicyRejected。
- 本层新增 UnsupportedBindingShape，专用于第 2 节新增的范围拒绝。
- statetime 其余六类同义保留：UnauthorizedKey、AssuranceMissing、
  InvalidTime、Expired、CheckpointRollback、StaleEvidence；它的
  InvalidInput/BindingMismatch/SignatureMismatch 映射上述共用分类。

selected-policy 的 PolicyFailure 只复制合法的 Code/Reason 闭合枚举，
转为本层同名只读诊断类型，Unwrap 为本层 PolicyRejected；Error 文本
固定，不保留下层错误对象/动态文字。Code 1–5、Reason 1–14 的既有范围
不扩大，typed-nil 诊断不得解引用；未知非 nil 错误归 InvalidInput。
处理已知取消后，先检查并复制非 nil PolicyFailure，再作普通 sentinel
映射；不能先匹配它的 Unwrap 而丢掉 Code/Reason。typed-nil PolicyFailure
和超出枚举范围的诊断归 InvalidInput。

当前 ctx.Err 优先，保留 errors.Is(Canceled/DeadlineExceeded)。当当前
context 尚未取消但下层报告这两个固定取消类别时，仍返回相应标准
context sentinel，不尝试恢复或重试；取消错误不要求本层文字前缀。
错误顺序固定为第 5 节阶段顺序，同阶段按保留的集合顺序检查，不依赖
map 遍历。整个失败路径不改调用者输入，不返回部分 owner。

## 9. 隔离与未来允许修改的范围

未来仅新增 statetimebinding 产品/测试文件，最多调整以下八个既有
`boundary_test.go` 的精确消费者例外及其正反例，不改变产品能力守卫：

1. statepolicybinding：允许新包唯一调用完整 Evaluate、Summary、Clear。
2. statetime：允许新包唯一调用完整 EvaluateModel、Summary、Clear。
3. statebinding：仅 Input/Summary 等类型、完整复制和摘要核对。
4. statepolicy：仅模型值复制/清理、状态常量和 PolicyFailure 枚举诊断。
5. stateverify：只改其 statepolicy 消费者判断；不允许新包导入 verifier。
6. statecontract：仅主状态类型、常量及复制。
7. statematerials：仅表/引用/期望类型、常量及复制核对。
8. statecomposition：仅嵌套 Summary/Links/Coverage 类型及独立复制。

例外必须精确到仓库相对目录 `internal/nodecontrol/statetimebinding` 且
package 名 `statetimebinding`。statetime 现有 WalkDir 原始相对路径须先
转换为仓库相对路径再做精确判断；不以 `../../../`、字符串前缀或同名
包放宽匹配。子目录、相似前缀、正确目录错误包名，以及 reader/serving/
authority/store 等消费者仍拒绝。新包自身禁止非测试产品消费者。

新包使用精确 import/成员能力白名单。判定入口仅上述两包；其他下层
仅限定类型/常量及自有值清理。contracts 仅 Digest；不直接导入
stateformat，根公钥等嵌套类型用有界泛型值复制，不调用 parser、crypto、
digest、JSON/JCS 或 State()。标准库只用于 context、固定错误、bytes
复制/比较、排序、DeepEqual 和显式 time.Time 规范化。禁止系统时钟、
provider、文件/进程、数据库、网络、随机、环境、unsafe、goroutine、
dot/blank/重复 import、动态反射及直接调用其他判定器。

AST 清理守卫分别验证固定 owner 登记/摘要保护/最后转移的源码形状，
不是任意 Go 程序的形式化证明。八份旧测试只变消费者判断及精确正反例；
下层全部产品、其他既有测试、依赖、SQL、生成文件和 captured vectors
保持不变。若必须扩大范围，停止并提出书面修订，不默认为本设计已授权。

## 10. 未来测试与有限验收

实施计划应测试先行，使用离线固定签名 key、严格格式的真实下层材料、
显式时间和模型观测，不加产品测试 hook。组合负例应有对应下层分别
成立的正控制，区分“组合不一致”与“下层本来无效”。覆盖：

1. 不变当前 root/metadata；Desired 不变/更新；检查点等值/前进；完整
   已接受 0/1/32 轮换历史链。合法新 root/metadata 仍在下层成立，本层
   UnsupportedBindingShape；不把候选标成已提交。
2. metadata 原字节跨 Batch/两表/Time 四处关联；同 P 不同 E 的独立有效
   替代；完整 section 中 key、history、ledger、window 或引用变化。
3. node/epoch、共同 BaseStateID、独立来源/ObservationID、provider 错配；
   来源 ID 不新增无依据等式；旧模型保留不投影字段，其他域仍有未完成责任。
4. 两份时间 Now/Floor/ObservationID 精确关系、纳秒/时区/零值、三种
   floor、下层过期/无观察/未授权 key/签名错误；不得补时间或抬 floor。
5. old/signed new/Main new 检查点逐项错配、回退、等值、前进和
   AdvanceRequired；区分 artifact sequence/StateEpoch，不做自动推进。
6. 全零 nonce 合法正例；错误 nonce；一致替换请求与签名证明新 Evaluate
   可成立但旧 Recheck 不同；未消费输入重复成功明确不是 replay protection。
7. 下层历史缺证/冲突、紧急撤销及其他策略拒绝继续失败；完整错误分类、
   PolicyFailure 闭合值、typed-nil、未知错误、context 优先及固定文字。
8. Summary 全嵌套责任和五项固定责任、身份/有效期限较早值、独立可变
   集合；无 State/Commit/Ready；检查点变化不被策略 NoChange 掩盖。
9. 全输入 Recheck 包括未选材料、未投影域、Old/History、全部观察身份，
   正向时间 1 ns 仍需新 Evaluate；仅既定规范化等价可重检通过。
10. 每项预算等长/+1，包括额外 Time raw、总量及别名按出现次数计；
    全部嵌套复制隔离；Clear 可达 backing store 清零不伤调用者/独立摘要。
11. 复制/两个下层/捕获可变摘要后/最终转移前的取消和正常 panic 清理；
    源码登记守卫与可观察 backing store 测试分开记录，不伪称已观测
    不可返回的私有 owner，也不增加测试专用产品入口。
12. 八份旧守卫及新包精确正反例；下层判定/时钟/副作用能力均无放宽。

未来局部验证矩阵为 16 个独立命令：新包加既有 statepolicybinding、
statebinding、statecontract、statematerials、statecomposition、stateformat、
stateverify、statepolicy、staterecovery、statetime、serving、contracts、
readiness 共 14 个完整包；authority 仅 TestBoundRead/TestCoordinator；
store 仅编译。Go 1.26.5、离线依赖、`-count=1 -timeout 120s -p=1`，如
需缓存仅设本进程所属临时目录，不改永久环境配置。精确命令、任务/文件
边界和证据冻结规则在书面计划中确定，不将此处当作已执行结果。

不包含 race、全仓、真实 DB/Docker、C11/C12、bootstrap/PITR 或硬件
验证。实现后的只读增量审查及这些局部测试不能称为完整安全审查或生产
验收；已有整体验收阻塞不因新模块通过而消失。

## 11. 本轮交付、审批与自检

本轮仅新增本文，不修改源码、路线图/既有完成记录、实施计划或 Git
索引，不安装依赖，不运行代码测试，不暂存/提交/推送/合并/清理。
brainstorming 默认提交步骤受已确认的“仅书面设计、不提交”范围约束；
本文只有本机副本，尚不是跨电脑恢复检查点。

用户审核批准本文之后，下一步仅使用 writing-plans 编写实施计划；
计划批准后才实现。沿用此前 Native 本机执行选择，不重复选择执行方式，
仍以届时批准的精确计划为准。当前批准方向不等于批准尚未存在的计划。

自检范围：已提交/候选边界、完整 root/metadata 不变限制、共同 StateID
约定与独立来源区分、原字节绑定、固定时间与 floor、nonce 不消费、
检查点只绑定声明、完整责任/重检/清理、20 类固定错误、八份测试例外、
16 项未来有限验证及审批边界。本文不含未决定占位项；文档自检不是
独立源码/安全审查，也没有本轮测试通过声明。
