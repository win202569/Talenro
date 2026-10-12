# Task 10：双槽提交与崩溃恢复隔离模型设计

日期：2026-10-09。

状态：用户已批准本书面设计及对应实施计划；隔离模型已实现、限定回归通过，最终审查 Important 已修复；一项 Minor 测试覆盖待办保留。
只实施本设计批准的离线模型，不修改 Git 索引；不提交、推送。

## 1. 目标、依据与已有成果

验证父设计 `2026-08-23-node-pop-control-plane-design.md` §12 的顺序：
完整候选写入 inactive slot → file+directory fsync → 原子递增硬件计数器
→ 可选 pointer。通过确定性中断测试明确重启后应选择什么、何时必须
拒绝恢复。不接触实际磁盘、进程、TPM 或运行时状态。

现有 statepolicy 只判断部分节点信任状态，statetime 只判断条件性时间
证据。它们既未持久提交，也未覆盖生产 sealed blob 所有字段。本文
不把 Candidate.State() 从 Unknown 改为 Available，不增加生产消费者，
不把时间采样/普通报告当成应递增计数器的安全状态变化。

成功标准是提交顺序和恢复选择的模型级可证伪测试，不是断电持久性、
硬件反回放、完整 high-water 语义或节点启动资格的证明。

## 2. 方案选择与非目标

采用独立纯包 `internal/nodecontrol/staterecovery`，包含恢复选择器和
无 I/O 的提交步骤状态机。两者仅消费显式固定值观测，返回条件性结果。
真实文件双槽写入器和硬件适配器是另一条路线；缺少真实平台验收条件，
本阶段不实施。也不把恢复逻辑并入 statepolicy，以免局部 trust model
被误用为完整生产 sealed state。

明确不实现：磁盘格式/序列化、MAC/seal 算法、密钥管理、fsync、计数器
驱动、自动重试、实际删除、initial provision、reenroll、跨 epoch
迁移、LocalSecurityLatch、supervisor 状态、TLS、真实时钟或启动 LKG。
不新增依赖、DB schema、网络接口或部署规范登记。

## 3. 模型输入与来源限制

所有正向状态都显式称为 Model/Observation；Available/Intact 等标签
只是调用方提供的假设，不证明来源。禁止添加“已持久化/可信”布尔值
后把普通内存输入包装成生产回执。

### 3.1 作用域和计数器

ExpectedScope 包含非零 canonical node UUID、control-plane authority
epoch、GuardID、SealKeyID、DeploymentID。各 ID 使用非零 32-byte digest；
authority epoch 保持 1..MaxInt64。该模型只处理同一作用域，禁止网络
输入建立新作用域或重置计数器。

CounterObservation 包含状态（Unknown/Available/Unavailable/Rollback）、
SourceID、ObservationID、ScanID、Scope、Value、Maximum。计数值使用
uint64，已初始化状态要求 `1 <= Value <= Maximum`；Maximum 是显式
provider 容量假设，范围 1..MaxUint64，不借用 authority epoch 的数域。
Value==0 不自动成为首次初始化；初始化设计留待后续。递增前要求
Value < Maximum，并 checked +1，禁止溢出或跳号。

### 3.2 两个固定 slot

SlotObservation 固定为 A、B 两个具名位置，不接受可变长度列表或文件
路径。每个观测携带相同 ScanID、非零 ObservationID/SourceID 和状态：
Unknown、Absent、Intact、Corrupt、Unsupported、Unavailable。

Intact 携带 scope、StateEpoch、ContentDigest、BlobDigest 和模型版本
ModelVersion=1。ContentDigest 代表完整业务状态的身份（不含提交 epoch）；
BlobDigest 代表包含 scope、state epoch 和内容的完整 sealed blob 身份。
二者不是可互换摘要。本模型没有 bytes，因此无法重新计算摘要或验证
seal；Intact 的真实性仍由未来适配器证明。Absent 不得携带 blob 字段，
其他异常状态不允许通过填充字段变成可恢复状态。

ScanID 相等只表达“一致扫描”的模型前提，不证明真实扫描原子性。
未来适配器必须在独占写入锁下读取计数器/两个 slot 并复核计数器未变，
同时保证锁和来源不可伪造；这里不实现或声称拥有该能力。

输入附带可选 PointerHint（整数位置提示），选择器完全不读取其值，
即使越界或与结果冲突也不改变选择。pointer 无权覆盖任何校验。

### 3.3 完整状态不在本阶段编造

ContentDigest 是不透明且未经本模块验证的完整状态承诺，不是把
statepolicy.StateModel 序列化后的摘要。未来完整 sealed state 必须
覆盖父设计列出的 checkpoint、root/metadata/desired/recovery、撤销账本、
身份/证书/CA bundle、manifest/map、host policy、supervisor/LKG 等全部
high-water，并在持久化前完成各自语义校验。

本模块既不能证明内容完整，也不能根据两个摘要推导 high-water 未降低。
拟议状态的授权与单调性是外部假设；不声明本模型提供业务回退防护。

## 4. 恢复选择规则

建议入口 `RecoverModel(ctx, RecoveryInput) (*RecoveryModel, error)`。
首先验证 context、作用域、观测来源字段及完整快照，再检查两个 slot。
未知/不可用计数器或 Rollback 状态直接拒绝；不改用文件 epoch 猜测值。

逐个检查两个 slot，不能找到一个好 slot 后跳过另一个：

1. 任一 Unknown/Unavailable：缺少观测，拒绝。
2. 任一 Corrupt/Unsupported、未知状态、错误版本、作用域/密钥/guard
   不匹配、无效摘要或 epoch：安全故障，拒绝，即使另一槽是好副本。
3. 两槽 Intact 且 StateEpoch 相同：重复/歧义故障，拒绝。相同摘要也
   不任选其一；不同摘要明确为分叉。对应父设计“重复”“唯一 blob”。
4. 必须且只能有一个 Intact slot 的 StateEpoch == Counter.Value。
   没有则 counter/blob mismatch；不能降级选择旧槽，也不能推进 counter
   来迎合未来槽。
5. 另一槽可以 Absent、完整旧 epoch，或完整的紧邻未来 epoch Value+1。
   完整旧槽不成为备选。未来 epoch > Value+1 或超过 Maximum 均故障。
   Value==Maximum 时不允许未来槽。

成功结果私有字段保存完整依赖快照；只暴露按值 summary：选中槽、
epoch、内容/封装身份、作用域、完整未来槽的可丢弃标记。该标记是
诊断建议，不是删除命令；本模块从不执行删除。

错误分类固定为 InvalidInput、AssuranceMissing、SecurityFault、
CounterExhausted、StaleModel、InvalidTransition 和 context 取消错误。
SecurityFault 附固定原因：CorruptSlot、UnsupportedVersion、ScopeMismatch、
DuplicateEpoch、Fork、CounterMismatch、UnexpectedFuture。错误不输出
原始身份或状态内容。不返回部分成功或“旧状态也许可用”。

## 5. 提交步骤状态机

建议 `PrepareModel(ctx, RecoveryInput, ProposalObservation)` 先重跑恢复
选择，再绑定非零 ProposalID、AttemptID、BaseBlobDigest、ProposedContentDigest、
ExpectedScope 和 BaseStateEpoch。BaseBlobDigest/BaseStateEpoch 必须匹配恢复选中槽。
ProposalObservation 是外部业务校验模型，不是可信授权回执。

若 ProposedContentDigest 与当前相同，返回 NoChange，不产生任何写入或
counter 推进建议，重复 sampling/poll 可用测试证明零步更新。若不同，
须 counter 未耗尽；记录目标为另一槽、target epoch=current+1、完整
旧快照及提案身份。NoChange 为终态，不接受 Advance/Confirm；未知、
零值或被 Clear 的模型不可继续。

状态机无 provider 调用，只在输入匹配的观测事件时返回下一步骤：

| 阶段 | 接受的模型事件 | 后继 |
| --- | --- | --- |
| Prepared | CandidateWritten：目标槽、scope、target epoch、提案内容及新 blob 身份全部匹配 | Written |
| Written | FileSynced：同一候选身份 | FileDurable |
| FileDurable | DirectorySynced：同一候选身份 | CandidateDurable |
| CandidateDurable | BaseRechecked：当前 counter/source/scope、旧 active blob 和目标候选仍匹配 | ReadyToAdvance |
| ReadyToAdvance | CounterAdvanced：同一 counter、old→old+1，候选身份未变 | NeedsRecovery |
| 任一非终态提交阶段 | Failed 或 OutcomeUnknown（带有效阶段/尝试绑定） | NeedsRecovery |

所有事件携带 AttemptID、非零 EventID、非零 SourceID、完整 scope 和
绑定值。AttemptID 随 Prepare 由调用方显式给出，模型不生成随机数。
每个事件只能用于当前允许的阶段；重复、跳步、跨尝试、交叉槽或换摘要
均 InvalidTransition/StaleModel，不忽略错误继续前进。
同一次线性模型链禁止复用 EventID；NeedsRecovery 不再接受 Advance。
BaseRechecked 携带新的 ScanID，必须异于初始扫描，并检查计数器来源、
容量与初始值未变；后续确认扫描也不得复用初始或该重检 ScanID。

建议 `AdvanceModel(ctx, *CommitModel, Event)` 返回新的独立模型，不修改
旧模型。旧模型可能被调用者重复使用，因此这只是顺序判定，并非
并发控制或 exactly-once 执行器。失败时无部分下一模型。

每一事件仅代表阶段完成假设。底层操作实际发生且响应丢失时必须报告
OutcomeUnknown，不能直接重发递增。状态进入 NeedsRecovery 后，唯一
允许的确认路径是提供全新的 scan 给 `ConfirmModel`，由 RecoverModel
独立选择，并要求同一作用域/provider、与初始相同的 Maximum 和该尝试绑定：

- 若递增未获确定成功确认，counter 仍为 old，选中原 blob：返回
  NotCommitted；不自动重试。
- 若 CounterAdvanced 已明确确认成功，却重新观察到 old：回退故障，
  不把它降格为 NotCommitted。
- 只有递增步骤已进入 ReadyToAdvance 后报告 Failed/OutcomeUnknown，
  或已收到 CounterAdvanced，才允许确认 old+1。选中目标槽且预期
  blob/content 完全匹配时返回 CommittedModel；更早阶段出现 old+1
  是无法归属本次尝试的状态变化，拒绝。
- 其余值、槽、摘要或恢复错误：拒绝，不据 ACK 猜测结果。

新 ScanID 必须不同于初次扫描，观测齐全；陈旧扫描不能确认。若候选
blob 身份尚未由匹配的 CandidateWritten 事件建立，不能仅凭相同内容
接受 old+1；返回保障不足。保守拒绝不会尝试递减硬件 counter。

CommittedModel 仍是模型结果，不输出 provider 回执，不更新任何现有
statepolicy/statetime 观测。可选 pointer 写入不纳入提交成功条件，
也不创建允许调用者修改真实 pointer 的接口。

## 6. 崩溃点和可观察结果

测试驱动程序只在 `_test.go` 中模拟磁盘可见性与 counter 原子性；
以下每一行独立构造重启扫描，不把内存中的阶段标签当成存储事实。

| 中断点/重启观测 | 恢复结果 |
| --- | --- |
| 尚未触及 inactive slot；counter=old | 选择唯一完整 old blob |
| 写候选过程中，只留下截断/损坏槽 | 拒绝恢复，即使旧槽完整 |
| 候选写完、尚未确认 file/dir fsync | 分别枚举旧内容仍在、Absent、完整 future、损坏；依第 4 节判定，不能假定任一结果必然发生 |
| file+directory fsync 已完成；counter=old | 选择 old；完整 future 仅可丢弃诊断 |
| counter 请求结果未知 | 重新读取；只允许看到 old 或 old+1 对应的确定选择，否则拒绝 |
| counter 已为 old+1；pointer 未更新或指向旧槽 | 选择完整新 blob，不依赖 pointer |
| counter 已为 old+1；新 blob 丢失/损坏 | 拒绝，不回退到完整旧 blob |
| 两槽同 epoch、跨 guard/key、缺失一致观测 | 拒绝，无 LKG 授权 |

模型中的 durable 事件是测试假设，不证明真实 fsync 在具体文件系统、
设备缓存、虚拟化层或突然断电条件下的效果。损坏候选导致停止是遵守
父设计的保守行为，不为提高可用性而新增“忽略损坏 inactive slot”。

## 7. 所有权、取消和隔离

输入/summary 优先固定值，不携带任意 blob bytes、文件路径、secret。
模型持有独立快照，Clear 幂等且 nil 安全；清除后不能继续或确认。
调用期间不得并发修改输入或 Clear；不承诺运行时副本擦除。

nil context 拒绝，预取消优先；解析/选择/转换返回前再检查取消。
所有错误无部分成功，无状态写回，无自动 quarantine 或清理副作用。
“必须停止/隔离”是未来运行时消费错误的要求，而非本模型已执行的动作。

纯包依赖仅限明确允许的标准库、既有 UUID/contracts 类型；不依赖
statepolicy/statetime，不解除其隔离 guard。不使用时钟、网络、文件、
进程或随机源。AST 正反测试禁止非测试运行时代码导入本包，测试多别名
绕过和点导入；无默认 adapter、无生产运行入口。

## 8. 验证计划边界

后续实施计划必须为以下内容安排独立、可检查的测试：

- 作用域、两槽完整检查、Absent/Unknown/Corrupt/Unsupported 差别、
  相同 epoch 同摘要与分叉、counter unavailable/rollback/零值。
- 正常选择、完整旧/未来槽、异常未来槽、pointer 任意值不影响判定、
  最大计数值下 NoChange 与拒绝增量、checked +1。
- 每个合法状态转换及所有跳步/重复/错 attempt/错槽/错 blob 反例。
- 相同业务内容不安排写入；只改 pointer 不推进计数器。
- crash table 全部场景、counter ACK 丢失后双向确认、旧 scan 拒绝、
  old+1 不匹配候选时绝不回退或盲目重试。
- nil/零模型、取消、Clear、独立副本、错误类别与不泄露输入。
- 运行时无消费者/无 I/O 的正反边界测试。

本阶段不使用 Docker 故障注入或真实文件模拟持久化；测试 helper 明确
标为进程内假模型，不称为父设计生产 provider 或 container gate 证据。
限定回归沿用现有包加 staterecovery；authority 仅
TestBoundRead/TestCoordinator，store 仅编译。无整仓、真实 DB/Docker、
C11/C12、bootstrap、PITR、race 或硬件/重启测试。

## 9. 交接与验收限制

本文经用户审阅批准后，使用 writing-plans 编写实施计划；计划再次
获批后才实现。沿用 Native 方式及最终一次独立审查，不自动提交或推送。
实施计划已获批准；不执行技能模板的自动提交。

未来真实集成至少还须证明：完整 state schema 与业务单调性、seal/key
来源、独占锁与一致读、实际 fsync/原子 increment 语义、ACK 丢失处理、
三个独立 guard 的隔离、LocalSecurityLatch 优先检查、受控重启与回放
拒绝。当前成功不能解除任何生产门禁或声明整体 Task10/B01/C12/main 验收。
