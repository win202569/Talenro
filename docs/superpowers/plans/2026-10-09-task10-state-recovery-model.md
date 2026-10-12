# Task 10 State Recovery Model Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task using the preserved Native method. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** 实现可独立证伪的双槽恢复选择和提交顺序模型，不提供真实持久化或生产启动资格。

**Architecture:** 新建纯包 `internal/nodecontrol/staterecovery`，消费显式固定值观测，返回私有快照支持的条件性结果。恢复选择器先检查两槽；提交状态机以不可变后继和新扫描确认处理不确定结果。所有中断模拟只存在于测试。

**Tech Stack:** Go 1.26.5、PowerShell、既有 google/uuid 与 contracts.Digest；无新增依赖。

**Spec:** `docs/superpowers/specs/2026-10-09-task10-state-recovery-model-design.md`（2026-10-09 用户批准）；父设计 `docs/superpowers/specs/2026-08-23-node-pop-control-plane-design.md` §12。

状态：用户已批准计划，Native 本地实施完成；限定回归通过，一次独立审查的 Important 已复现修复；一项 Minor 测试覆盖待办保留，未提交/推送。

## Global Constraints

- 工作目录固定为 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`；保留既有修改、分支和 HEAD，不自动暂存、提交或推送。
- 只新增 staterecovery 产品/测试文件及本增量记录；不得修改 statepolicy/statetime 的观测或解除既有隔离检查。
- Counter 为 uint64，`1 <= Value <= Maximum`，Maximum 为 `1..MaxUint64`；authority epoch 为 `1..MaxInt64`；ModelVersion=1。
- NodeID 使用 canonical、非零 RFC4122 UUID 字符串；其余身份为非零 32-byte contracts.Digest。不得增加真实 blob、路径、secret 或任意输入日志。
- 两个固定槽 A/B；pointer 是完全忽略的整数提示；不提供初始化、自动重试、删除、回退到备份或 LKG 授权。
- 无 I/O、时钟、网络、进程、随机源、seal/crypto 实现、硬件适配器、真实 provider 回执和生产消费者。
- 同一 ScanID 和 Intact/Available 仅为模型假设，不证明来源、原子扫描、持久性或完整业务 high-water 单调性。
- 所有模型持有私有固定值快照；Clear nil-safe、幂等；不保证并发 Clear 或清除运行时副本。旧模型可重放，不保证 exactly-once。
- nil context 拒绝；预取消优先于其他输入错误；返回成功前复查取消；错误时结果为 nil，无部分成功。
- 不下载、不改依赖锁；不运行整仓、真实 DB/Docker、C11/C12、bootstrap、PITR、race 或硬件/重启测试。
- Native 主代理实施，完成后一次 fresh 最强可用模型独立只读审查；本轮仅计划自查，不派实施或审查代理。
- 所有成功结论只限隔离模型；生产兼容性 UNKNOWN，整体 Task10/B01/C12/main 验收未解除。

## Review Focus

1. 已收到 CounterAdvanced 成功事件却新扫描回到 old：必须安全故障，不能 NotCommitted；任务 4 `TestRecoveryConfirmAcknowledgedRollback`。
2. 容量或来源在重检/确认时替换：相同数值仍不得接受；任务 3/4 `TestRecoveryRecheckBindings`、`TestRecoveryConfirmBindings`。
3. 同 epoch 同 digest 的两个副本也违反唯一性：任务 1 `TestRecoveryDuplicateEpoch`，区分 DuplicateEpoch 与 Fork。
4. 最大计数值下同内容为合法 NoChange，不同内容不能溢出：任务 2 `TestRecoveryPrepareCapacity`。
5. 分别使用多个别名或函数值取得 UUID 随机/外部来源能力：测试守卫不能只检查最后一个别名；任务 5 `TestRecoveryBoundaryAliases`。

## 文件与职责

下列新代码文件均位于 `internal/nodecontrol/staterecovery/`：

| 文件 | 职责 |
| --- | --- |
| doc.go、input.go、errors.go | 模型限制、固定值输入/枚举、固定错误分类 |
| recover.go、recovery_model.go | 完整扫描验证、唯一选择、恢复摘要及 Clear |
| prepare.go、commit_model.go | 提案/容量检查、私有提交状态及摘要 |
| advance.go | 事件绑定和有序不可变转换 |
| confirm.go | 独立新扫描与条件性最终结果 |
| fixtures_test.go、recover_test.go | 明确标记为内存假模型的夹具和选择测试 |
| prepare_test.go、advance_test.go、confirm_test.go | 提案、事件及确认测试 |
| crash_test.go、ownership_test.go、boundary_test.go | 中断矩阵、所有权/取消、AST 隔离 |

文档：实施结束新增 `docs/roadmap/2026-10-09-state-recovery-model-execution.md`，更新 `docs/roadmap/current-status.md`、本计划复选框和本设计状态，不覆盖既有历史。

## 固定接口契约（各任务共用）

这些名称/字段是实现决策；私有辅助函数由实施者保持最小化。

- `Scope { NodeID string; AuthorityEpoch uint64; GuardID, SealKeyID, DeploymentID contracts.Digest }`。用 uuid.Parse、RFC4122 Variant、非 Nil 和 String()==NodeID 检查规范形式，不生成 UUID。
- `CounterStatus uint8`：CounterUnknown=0、CounterAvailable、CounterUnavailable、CounterRollback。
- `CounterObservation { Status CounterStatus; SourceID, ObservationID, ScanID contracts.Digest; Scope Scope; Value, Maximum uint64 }`。
- `SlotID uint8`：SlotUnknown=0、SlotA、SlotB；`SlotStatus uint8`：SlotUnknownStatus=0、SlotAbsent、SlotIntact、SlotCorrupt、SlotUnsupported、SlotUnavailable。
- `SlotObservation { Status SlotStatus; SourceID, ObservationID, ScanID contracts.Digest; Scope Scope; StateEpoch uint64; ContentDigest, BlobDigest contracts.Digest; ModelVersion uint32 }`。
- `RecoveryInput { ExpectedScope Scope; Counter CounterObservation; A, B SlotObservation; HasPointerHint bool; PointerHint int }`。不通过指针表达可选提示，所有字段按值复制；两个 pointer 字段均不影响验证或结果。
- `RecoverySummary { Scope Scope; SelectedSlot SlotID; StateEpoch uint64; ContentDigest, BlobDigest contracts.Digest; DiscardFuture bool; FutureSlot SlotID }`。无未来槽时 FutureSlot=SlotUnknown。
- `RecoverModel(ctx context.Context, in RecoveryInput) (*RecoveryModel, error)`；`(*RecoveryModel).Summary() RecoverySummary`、`Clear()`。私有有效标记及完整快照；nil/无效摘要为零值。
- `ProposalObservation { ProposalID, AttemptID, BaseBlobDigest, ProposedContentDigest contracts.Digest; ExpectedScope Scope; BaseStateEpoch uint64 }`。
- `CommitPhase uint8`：PhaseUnknown=0、NoChange、Prepared、Written、FileDurable、CandidateDurable、ReadyToAdvance、NeedsRecovery。
- `CommitSummary { Phase CommitPhase; Scope Scope; ProposalID, AttemptID contracts.Digest; BaseSlot, TargetSlot SlotID; BaseStateEpoch, TargetStateEpoch uint64; BaseBlobDigest, ProposedContentDigest, CandidateBlobDigest contracts.Digest }`。
- `PrepareModel(ctx context.Context, in RecoveryInput, proposal ProposalObservation) (*CommitModel, error)`；`(*CommitModel).Summary() CommitSummary`、`Clear()`。NoChange 的 TargetSlot/TargetStateEpoch/CandidateBlobDigest 为零，不安排动作。
- `EventKind uint8`：EventUnknown=0、CandidateWritten、FileSynced、DirectorySynced、BaseRechecked、CounterAdvanced、Failed、OutcomeUnknown。
- `Event { Kind EventKind; Phase CommitPhase; AttemptID, EventID, SourceID contracts.Digest; Scope Scope; TargetSlot SlotID; TargetStateEpoch uint64; ContentDigest, BlobDigest contracts.Digest; Scan RecoveryInput; CounterBefore, CounterAfter uint64 }`。
- `AdvanceModel(ctx context.Context, model *CommitModel, event Event) (*CommitModel, error)`。模型额外私存初始快照、重检 ScanID、事件 ID 固定数组、候选身份、是否已明确 ACK 和失败前阶段；最多五个成功事件，或四个成功事件加失败，容量五。
- `ConfirmationInput { AttemptID contracts.Digest; Scan RecoveryInput }`；`ConfirmationOutcome uint8`：ConfirmationUnknown=0、NotCommitted、CommittedModel。
- `ConfirmationSummary { Outcome ConfirmationOutcome; AttemptID contracts.Digest; Recovery RecoverySummary }`；`ConfirmModel(ctx context.Context, model *CommitModel, in ConfirmationInput) (*ConfirmationModel, error)`；`(*ConfirmationModel).Summary() ConfirmationSummary`、`Clear()`。
- 错误哨兵：ErrInvalidInput、ErrAssuranceMissing、ErrSecurityFault、ErrCounterExhausted、ErrStaleModel、ErrInvalidTransition；支持 errors.Is。`SecurityReason uint8`：ReasonUnknown=0、CorruptSlot、UnsupportedVersion、ScopeMismatch、DuplicateEpoch、Fork、CounterMismatch、UnexpectedFuture。`SecurityError { Reason SecurityReason }` 的 Error 仅固定文本，Unwrap 返回 ErrSecurityFault；使用 errors.As 断言原因。

错误映射：nil context/非法 scope/零身份/不一致 ScanID/非法 counter 枚举或范围→InvalidInput；未知或不可用 counter/slot→AssuranceMissing；CounterRollback→SecurityFault/CounterMismatch。Intact 槽非法 digest/epoch、未知槽枚举或 Absent 携带 blob 字段→SecurityFault/CorruptSlot；错版本或 Unsupported→UnsupportedVersion；槽 scope 不同→ScopeMismatch。提案 base 或后续合法观测的绑定替换→StaleModel；非法阶段/事件/复用 EventID→InvalidTransition。恢复验证错误原样传播，不转成较弱的成功结果。

### 事件字段与确认约束

所有事件验证 EventID/SourceID 非零、当前 Phase、AttemptID、Scope、TargetSlot、TargetStateEpoch、ProposedContentDigest；候选建立后 BlobDigest 必须一致。CandidateWritten 的 BlobDigest 非零且不得等于旧 BlobDigest（epoch 改变）；Prepared 阶段 Failed/OutcomeUnknown 的 BlobDigest 必须零。其他阶段须已有候选身份。

- CandidateWritten/FileSynced/DirectorySynced：SourceID 匹配初始目标槽 SourceID；Scan 及两个 Counter 字段为零。
- BaseRechecked：SourceID 匹配 counter SourceID；Scan 是新的完整输入，用 RecoverModel 重验，选中旧 active，另一槽必须为精确候选。保持 ExpectedScope、所有 SourceID、counter Value/Maximum、旧 active blob 字段；ObservationID 可更新，三个 ScanID 相同且异于初始；CounterBefore/After 为零。
- CounterAdvanced：SourceID 匹配 counter SourceID；CounterBefore=old、CounterAfter=old+1；Scan 为零，精确候选身份沿用事件绑定。只从 ReadyToAdvance 进入 NeedsRecovery，记录明确 ACK。
- Failed/OutcomeUnknown：Prepared/Written/FileDurable 的 SourceID 匹配目标槽；CandidateDurable/ReadyToAdvance 匹配 counter SourceID；Scan 和 CounterBefore/After 为零。保存失败前阶段，统一进入 NeedsRecovery，不猜测实际副作用。
- 其他不使用字段必须零，避免含糊的第二套观测。拒绝后不改变原模型。NeedsRecovery 和 NoChange 均拒绝 Advance；Confirm 仅接受 NeedsRecovery。
- Confirm 扫描再次 RecoverModel；AttemptID、Scope、全部 SourceID、Maximum 保持一致；ScanID 不得复用初始或重检扫描。ObservationID 可更新，不以 ID 相同证明内容没变。
- counter=old：必须选中原槽、原 StateEpoch/ContentDigest/BlobDigest；已 ACK 则 CounterMismatch，否则 NotCommitted。
- counter=old+1：没有已建立候选则 AssuranceMissing；只有 ReadyToAdvance 的失败/未知或明确 ACK 可归属本尝试；过早失败后观察到增量为 StaleModel。合法时必须选中目标槽且 epoch/content/blob 精确匹配，否则 StaleModel，不按内容相同推断。
- 其他 counter 值拒绝（CounterMismatch）。Confirm 不修改 CommitModel；重复对同一条件扫描求值不代表消费回执或自动重试。

## 离线运行配置

每次新 PowerShell 会话先设以下配置，命令从固定 worktree 根执行：

```powershell
$env:GOENV='off'
$env:GOWORK='off'
$env:GOTOOLCHAIN='local'
$env:GOPROXY='off'
$env:GOSUMDB='off'
$env:GOAUTH='off'
$env:GOVCS='all:off'
$env:GOFLAGS='-mod=readonly'
$env:GOPATH='C:/Users/Lenovo/go'
$env:GOMODCACHE='C:/Users/Lenovo/go/pkg/mod'
$recoveryGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
```

RED 只来自新测试尚无接口或确实暴露缺陷，不修改正确产品行为制造失败；每组保存实际命令/退出码和断言。命令仍运行时等待原 session，不重复启动。所有检查点替代技能模板的 commit；提交不在批准范围内。

### Task 1：两槽完整验证与恢复选择

**Files:** 新 doc.go、input.go、errors.go、recover.go、recovery_model.go、fixtures_test.go、recover_test.go。
**Interfaces:** 产出以上 Scope/Counter/Slot/Recovery 类型、错误以及 RecoverModel/Summary/Clear；后续任务仅使用这些稳定接口，不自行省略恢复验证。

- [x] 在修改前保存既有产品、测试、go.mod/go.sum 和 DB 文件 hash 基线及当前工作区状态；记录受保护文件，禁止将历史修改算成本次实现。
- [x] 写 `TestRecoverySelect`、`TestRecoveryObservations`：counter=7、A=完整 epoch7、B=Absent 成功选择 A；分别检查完整旧6、未来8；两槽交换同样成立。逐项构造所有 counter/slot 状态、零身份、UUID 非 canonical/非 RFC4122、authority 0/MaxInt64+1、scope 各字段替换，断言 nil 结果及契约中的精确错误。
- [x] 写 `TestRecoveryDuplicateEpoch`、`TestRecoveryFutureAndPointer`：7/7 同摘要为 DuplicateEpoch、不同摘要为 Fork；仅6/8 而 counter7 为 CounterMismatch；未来9 或超过 Maximum 为 UnexpectedFuture。pointer 缺失、指 A/B、负数和最大 int 不改变结果；有效 A 旁的损坏/缺失观测 B 不能被忽略。
- [x] 运行 `& $recoveryGo test ./internal/nodecontrol/staterecovery -run '^TestRecovery(Select|Observations|DuplicateEpoch|FutureAndPointer)$' -count=1`，记录 RED。
- [x] 实现 RecoverModel 及固定类型/错误。先校验整体输入再完整检查两槽，重复 epoch 在选择前拒绝；仅在 Value<Maximum 时计算 Value+1；Absent 的 scope/epoch/digests/version 必须全零。快照中规范化忽略 pointer，不把提示保存为权威依赖。
- [x] 运行同一命令得到 GREEN；补全零/清除 RecoveryModel 摘要测试，记录独立检查点，不提交。

### Task 2：提交准备、无变更与容量边界

**Files:** 新 prepare.go、commit_model.go、prepare_test.go；input.go 增加提案、阶段和摘要类型。
**Interfaces:** 消费任务 1 RecoverModel；产出 PrepareModel、CommitModel.Summary/Clear；私有有效标记与完整初始快照为任务 3/4 使用，不暴露 setter。

- [x] 写 `TestRecoveryPrepareBindings`、`TestRecoveryPrepareNoChange`：ProposalID/AttemptID/摘要为零拒绝；base epoch/blob、scope 不匹配为 StaleModel；正常选 A 则目标 B/8，反向亦然。相同 content 返回 NoChange，目标三字段全零，输入及 counter 不变；仅 pointer 改变也不产生写入建议。
- [x] 写 `TestRecoveryPrepareCapacity`：Maximum=7、Value=7 以及 Maximum=MaxUint64、Value=MaxUint64 下，同内容均 NoChange，不同内容为 CounterExhausted。MaxUint64-1 到 MaxUint64 是有效目标；不得溢出到0。非法容量和值先被 RecoverModel 拒绝。
- [x] 运行 `& $recoveryGo test ./internal/nodecontrol/staterecovery -run '^TestRecoveryPrepare' -count=1`，记录 RED。
- [x] 实现 PrepareModel：先完整恢复，再验证提案绑定，先判 NoChange 后检查容量；固定值转移到独立 CommitModel，临时 RecoveryModel 清理不影响结果。不给 NoChange 任何可执行动作。
- [x] 同命令 GREEN，完整包回归；记录检查点，不提交。

### Task 3：不可变步骤转换与递增前重检

**Files:** 新 advance.go、advance_test.go；input.go 加 Event 类型，commit_model.go 加私有跟踪字段。
**Interfaces:** 消费任务 1/2；产出 AdvanceModel。重检调用 RecoverModel；绑定表和错误分类使用本计划固定接口契约。

- [x] 写 `TestRecoveryAdvanceSequence`：逐阶段断言 Prepared→Written→FileDurable→CandidateDurable→ReadyToAdvance→NeedsRecovery；原模型的 Summary 在每次转换后不变。每一非终态各自接受 Failed/OutcomeUnknown 并进入 NeedsRecovery；NoChange/NeedsRecovery/零模型拒绝。
- [x] 写 `TestRecoveryAdvanceRejections`：枚举所有不合法 phase×kind、重复 EventID（包含早先阶段 ID）、跨 AttemptID、零来源、错来源、错槽、错epoch、换内容/blob、CandidateWritten 沿用旧 blob、未用字段非零；错误结果 nil、原模型无变化。
- [x] 写 `TestRecoveryRecheckBindings`：新 ScanID 正例；初始 ScanID、counter Value/Maximum/SourceID、槽 SourceID、旧 active 身份或候选身份逐项替换拒绝；相同 ObservationID 下改值也不能绕过。新 ObservationID 且所有依赖正确可接受。
- [x] 运行 `& $recoveryGo test ./internal/nodecontrol/staterecovery -run '^TestRecovery(Advance|Recheck)' -count=1`，记录 RED。
- [x] 实现 AdvanceModel 的事件字段表、固定事件 ID 数组及按值后继；每次事件均先检查当前模型有效性和阶段。CounterAdvanced 只记录确定 ACK；失败或未知只记录先前阶段，不推断计数器状态，不实现任何真实调用或重试。
- [x] 同命令 GREEN，再跑完整包；记录阶段矩阵检查点，不提交。

### Task 4：新扫描确认和全部中断点

**Files:** 新 confirm.go、confirm_test.go、crash_test.go、ownership_test.go；input.go 加确认类型。
**Interfaces:** 消费 RecoverModel、CommitModel 私有快照；产出 ConfirmModel、ConfirmationModel.Summary/Clear。不增加 statepolicy/statetime 接口。

- [x] 写 `TestRecoveryConfirmOutcomes`、`TestRecoveryConfirmAcknowledgedRollback`：ReadyToAdvance 结果未知后，扫描 old→NotCommitted、精确 new→CommittedModel；明确 ACK 后 old→SecurityFault/CounterMismatch。候选未建立时 new→AssuranceMissing；较早失败时 new→StaleModel；新槽内容相同而 blob 不同仍拒绝。
- [x] 写 `TestRecoveryConfirmBindings`：错 attempt、初始/重检 ScanID 重用、provider/容量改变、选错槽或 blob、counter 跳号均无成功结果；未获 ACK 且 old 选中非原 blob 也拒绝；保持真实恢复错误，不退回旧槽。
- [x] 写 `TestRecoveryCrashMatrix`：每个用独立扫描夹具覆盖设计 §6 所有行。写入未同步时分别枚举旧槽仍在、Absent、完整 future、Corrupt；counter=new 且 pointer 旧/缺失照选新；new blob 丢失/损坏不能选 old。fixture 名称说明进程内模型，绝不运行真实文件注入。
- [x] 写 `TestRecoveryOwnership`、`TestRecoveryCancellation`：输入与摘要副本修改不改变各模型；清理旧模型不影响后继；nil/重复 Clear、清理后 Advance/Confirm、nil context、预取消和确定性末尾取消均拒绝且 nil。错误文本不含输入 UUID/digest；errors.Is/As 精确断言。
- [x] 运行 `& $recoveryGo test ./internal/nodecontrol/staterecovery -run '^TestRecovery(Confirm|Crash|Ownership|Cancellation)' -count=1`，记录 RED。
- [x] 实现 ConfirmModel 与确认结果所有权；先检查模型/尝试绑定再独立 RecoverModel，按已 ACK/可能递增/早期失败分别判断；新 ScanID 必须不同于初始及重检。完成所有入口末尾取消检查和 nil-safe Clear，不原地修改或消费提交模型。
- [x] 同命令 GREEN，再跑完整包；逐行核对 crash matrix 与错误原因，记录检查点，不提交。

### Task 5：无生产入口的隔离、限定回归及一次最终审查

**Files:** 新 boundary_test.go；更新本计划列出的文档，不改其他产品文件。
**Interfaces:** 测试专用 `recoveryImportsAllowed(src []byte) bool`、`recoveryConsumerAllowed(importPath string) bool`；产品接口不增加。

- [x] 写 `TestRecoveryBoundary`、`TestRecoveryBoundaryAliases`：产品允许导入仅 context/errors、google/uuid、contracts；uuid 所有别名只允许 Parse、Nil、RFC4122，contracts 仅 Digest，context 仅 Context/Canceled/DeadlineExceeded，errors 仅 New。禁止 dot/blank import、UUID New/NewRandom、函数值引用、time/net/os/unsafe/syscall、statepolicy/statetime；多别名样例必须全部检查，合法纯样例通过。对未知导入和无法解析 AST 默认拒绝。
- [x] AST 消费者扫描覆盖仓库全部非测试 Go 文件（含 cmd，跳过 .git/.worktrees/vendor），任何导入精确 staterecovery 路径都失败；测试可消费。正反样例测试 helper，不往产品文件临时注入危险代码。运行 `& $recoveryGo test ./internal/nodecontrol/staterecovery -run '^TestRecoveryBoundary' -count=1` 记录 helper 尚缺的 RED，实现后记录 GREEN。
- [x] 对本次新 Go 文件用固定工具链 gofmt；核对 protected hash、未跟踪文件和空白错误，确保既有代码及依赖锁未改。Git 只读如需提权按授权机制处理，不改 safe.directory 或安全设置。
- [x] 运行 `& $recoveryGo test ./internal/nodecontrol/staterecovery ./internal/nodecontrol/statetime ./internal/nodecontrol/stateformat ./internal/nodecontrol/stateverify ./internal/nodecontrol/statepolicy ./internal/nodecontrol/serving ./internal/nodecontrol/contracts ./internal/readiness -count=1`；要求每包 ok、退出0。
- [x] 运行 `& $recoveryGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1` 和 `& $recoveryGo test ./internal/store -run '^$' -count=1`；退出0，store 明确只编译、no tests to run。
- [x] 使用 requesting-code-review 技能组织一次 Native 最终 fresh 最强可用模型只读审查；提供已批准设计/计划、精确增量、基线、实际测试证据及范围禁令。修复发现须先复现再回归；实质设计变化先求批准；不以自动追加审查扩大范围。
- [x] 执行记录逐项写入实际 RED/GREEN 命令/退出码、审查发现及处理、未验证事项；更新当前状态及计划，仅实际完成才勾选。注明 sealed 内容完整性/业务授权/持久性/三个独立 guard/LocalSecurityLatch/受控重启仍待真实集成证明。
- [x] 最后复查受保护 hash 与变更范围，交付事实和文件；不暂存、提交、推送或声明总体完成。

## 计划自查与交接

覆盖映射：设计 §1–3→约束、接口及任务1/2；§4→任务1；§5→任务2/3/4；§6→任务4；§7→任务4/5；§8–9→任务5。五项 Review Focus 已逐项落到命名测试。

已自查：来源/容量重绑、重复 epoch、未知 ACK 与确定 ACK 的区别、候选未建立时的拒绝、固定数组容量、pointer 无权威性、取消及纯包导入闭包；接口名称与任务引用一致。无真实适配器、持久化或业务 high-water 语义混入。

用户已批准本计划，按 executing-plans 沿用 Native 实施；实际命令与证据记录在本增量执行记录中。
