# Main-state Selected-current-target Policy Binding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将主状态所选当前目标、旧策略投影及显式时间观测绑定到同一自有输入，返回可完整重检的条件性模型，不产生提交或运行权限。

**Architecture:** 新增隔离纯包 statepolicybinding，顺序组合 statebinding.Evaluate 与 statepolicy.EvaluateModel；仅支持已初始化、当前 root/metadata 目标和不变的主状态 floor。明确区分轮换证明对象与 RootSet 身份，保留下层完整检查及全部未完成责任，不改历史目标支持或下层算法。

**Tech Stack:** 固定 Go 1.26.5、现有 contracts/statebinding/statepolicy/stateformat 及下层值类型、PowerShell、apply_patch、离线缓存；无新依赖。

**Spec:** `docs/superpowers/specs/2026-10-10-main-state-selected-policy-binding-design.md`，用户于 2026-10-10 批准；执行前完整阅读设计及本计划。

状态：用户于 2026-10-10 批准本实施计划；五项 Native 任务及 41 步均已实际完成。五个 RED/GREEN 循环、十五项限定验证、唯一一次只读审查及最终源哈希/Git 核对全部完成；审查 0 Critical / 0 Important / 0 Minor，源码未再改变。执行记录及作者裁定已保存；不暂存/提交/推送，C12/main 未整体验收。

## Global Constraints

- 仅复用 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator` 及 `codex/c12-b01-task9-coordinator`；不新建或清理 worktree，保留既有未提交进度。
- 不暂存、提交、推送、合并、重置、删除分支或修改 safe.directory/永久环境；记录步骤替代技能模板中的 commit/清理。
- Binding 两表各 Materials/Expected ≤128；各材料/预期 RootRotation ≤32；Main.RootTransition/Target.Links 各 ≤32；独立 v1 锚仍只支持单 key/threshold 1。
- Binding 每份 RootSet raw ≤65,536 bytes；其他材料/锚各 ≤262,144 bytes；每表 raw ≤16 MiB；Binding raw 总量按出现次数计 ≤32 MiB + 262,144 bytes。
- 每份 expectation 全部分支预查：root keys 5、metadata keys 64、revoked 4096、slots 8，包括非活动分支。
- Main 旧、新各九个集合预查：root keys 5、metadata keys 64、revoked 4096、两个 purpose deauthorized 各 128、desired slots 8、LKG entries 8、local faults 64、supervisor faults 16；连 RootTransition 共十九个 L0 集合。
- Policy.Batch.Rotations ≤32；各次 anchor/rotation/active root/metadata/desired raw ≤262,144 bytes；Batch raw 总量 ≤36 × 262,144 bytes；Binding 加 Batch 总量 ≤32 MiB + 37 × 262,144 bytes，重复出现的 active root 也计数，不按别名去重。
- Old.Roots ≤33，每个 root Keys ≤5；Old.Keys/Old.RevokedIDs/History 各 ≤4096；History.IncidentKind ≤64 bytes；Old key Status/RevocationReason ≤8/32 bytes。预算不改变 policy 对等长输入的语义拒绝，也不承诺峰值内存。
- 已初始化 Old 非 nil，FreshInstall=false，旧/新 Desired 均 Present；所选 metadata/链尾 RootSet 与新 Main 当前引用精确相等，主状态可选 TimeFloor 旧/新完全不变。
- RootSet E=0；v1 策略根 Mark.Payload 对应 anchor P，轮换后的策略根 Mark.Payload 对应 rotation P，不等于其 NewRoot P；完整引用按封闭 Kind 映射，不直接转换 enum。
- 仅两表顶层 Materials/Expected 按 Kind/P 排序；全部空集合→nil，其他集合/链/History 保留原顺序。全部 time.Time 去单调表示并转 UTC，保留纳秒及零值；不取整、补时间或取系统当前时间。
- 三种 floor：主状态可选约束不变；存在时 Old.TimeFloor ≥ 主约束；ObservedFloor ≥ Old.TimeFloor，Now ≥ ObservedFloor。policy 新部分状态仍保留旧 floor，不将观测写回主 guard。
- 不推断 DeploymentID=SourceID、TrustAnchorID=锚 P/E、主 guard=Policy.CounterID、checkpoint=StateEpoch；不认证旧基线同一物理快照或完整历史。
- 产品唯一判定入口是 statebinding.Evaluate/statepolicy.EvaluateModel；stateformat 只用既有 parser/material；不直接调用其他下层 Evaluate 或导入 stateverify/statetime/staterecovery。
- 失败/取消返回 nil；panic 清理后传播。所有可变所有者在后续 context/下层调用前登记清理；不加产品 hook，不声明不可返回副本已直接观测，不承诺 OOM/致命终止、字符串/不可达/全进程擦除或并发访问。
- 下层全部产品源码/接口/wire/算法、冻结向量、依赖/锁文件、SQL/生成文件不变；既有测试仅六份指定消费者守卫可变。
- 无生产消费者、provider 实现/调用、外部 I/O、网络、系统时钟、环境、随机、硬件、重试、goroutine、动态反射；成功只为 ConditionalSelectedPolicyBound，下层及附加责任固定保留。
- 重检固定整个时间快照；Now/Floor/ObservationID 真正改变，即使向前且仍合法，也须新 Evaluate，不能沿用旧模型。Recheck 不证明真实当前时间或提交原子性。
- 手工编辑只用 apply_patch；gofmt 只限二十个新 Go 文件及六份指定旧测试。所有启动会话等待实际结束，不因输出发布延迟重复启动或将其伪称测试超时。
- 测试只用离线固定样例/显式时间/测试模型；精确命令见各任务及最终十五条矩阵，单命令 Go 测试超时固定 120s，不静默增加重试、超时或测试范围；无全仓/race/真实 DB/Docker/硬件/C11/C12/bootstrap/PITR。

## Review Focus

1. Binding 两表与 Policy.Batch 共享调用者 raw，active root 又别名最后 rotation：必须独立拥有每次出现，清理任一自有副本不得破坏调用者或其他副本（任务 1/4）。
2. 旧策略根的 rotation P 与主状态 RootSet P 不同，但完整 key 材料相同：必须作类型化路径映射，不能直接对 P 判等或只看 keys 数量（任务 2/3）。
3. Old.Keys 包含旧当前 metadata 不再列出的合法历史/tombstone：必须保留并交 policy 判定；不能删多余记录、遗漏当前 key 全字段或混淆 ledger 与当前 KeyIDs（任务 3）。
4. 两个下层各自成功的另一份 Batch，或合法历史目标/未投影主状态变化：组合必须拒绝 Batch 错绑、只对本包历史子范围拒绝，SelectedPolicyNoChange 不能表示十二域不变（任务 2/3/4）。
5. 时间仅前进 1ns、观测 ID 替换、最终摘要捕获后取消/panic：旧模型不得继续使用，必须完整重检和早清理；时区表示差异不能被误判成真实观测变更（任务 1/4/5）。

---

## 文件分工与冻结接口

新包路径 `internal/nodecontrol/statepolicybinding/`，下列文件名是格式化/冻结范围：

| 产品文件（11） | 单一职责 |
| --- | --- |
| doc.go / input.go | 条件性边界 / 完整输入 |
| errors.go | 固定 sentinel、policy 枚举诊断和映射 |
| budget.go / owned.go | 全输入预算 / 深复制、时间规范化、清理 |
| selection.go | 支持范围、精确 Batch、封闭引用查找 |
| roots.go | 根路径 owner、旧前缀/新完整链、本次后缀 |
| projection.go / time.go | 旧/新部分状态与当前 keys/ledger / provider 与主约束 |
| evaluate.go / model.go | 同快照组合 / 私有 owner、有限摘要、完整重检 |

新测试（9）：owned_test.go、fixtures_test.go、selection_test.go、roots_test.go、
projection_test.go、time_test.go、evaluate_test.go、model_test.go、boundary_test.go。
既有测试仅 `internal/nodecontrol/{statebinding,statepolicy,stateverify,statecontract,statematerials,statecomposition}/boundary_test.go` 六份。

后续执行记录为 `docs/roadmap/2026-10-10-main-state-selected-policy-binding-execution.md`，
台账在 `.superpowers/sdd/2026-10-10-main-state-selected-policy-binding/`；实施时才
更新 current-status、设计/计划 checkbox 和执行记录，不提前写完成。

```go
type Input struct { Binding statebinding.Input; Policy statepolicy.ModelInput }
type ResultStatus uint8 // 0 InvalidResult; 1 ConditionalSelectedPolicyBound
type Obligation uint8 // 0 UnknownObligation; 1..3 按设计的三项顺序
type SourceKind uint8 // 0 UnknownSource; 1 InitialAnchor; 2 RootRotationProof
type SourceIdentity struct { Kind SourceKind; Payload, Envelope contracts.Digest }
type RootProjection struct {
    RootSet statematerials.MaterialRef
    Version, Sequence uint64
    Source SourceIdentity
}
type PolicyBase struct {
    StateID, CounterID contracts.Digest
    StateEpoch uint64
    DeploymentSourceID, DeploymentObservationID contracts.Digest
}
type PolicyProjection struct {
    Root RootProjection
    Metadata, Desired statepolicy.HighWater
    KeyCount, RevokedCount int
}
type TimeSnapshot struct {
    ProviderID, ObservationID contracts.Digest
    Now, OldFloor, ObservedFloor time.Time
    MainFloor statecontract.OptionalTime
}
type Summary struct {
    Status ResultStatus
    Binding statebinding.Summary
    Base PolicyBase
    Previous, Proposed PolicyProjection
    HistoryCount int
    Time TimeSnapshot
    ValidUntil time.Time
    SelectedPolicyNoChange bool
    Obligations [3]Obligation
}
func Evaluate(ctx context.Context, in Input) (*Model, error)
func Recheck(ctx context.Context, model *Model, in Input) error
func (model *Model) Summary() Summary
func (model *Model) Clear()
```

Model 只有 `valid bool`、`input Input`、`summary Summary` 三个私有字段，
不保留下层 owner/State()，不暴露 State/Commit/Authorize/Ready；可变摘要
仅 Binding.Proposed.Target.Links/Coverage，必须拥有并独立复制。

errors.go 定义 13 个 `Err<Class>` sentinel，Class 依次为 InvalidInput、
LimitExceeded、IncompleteCoverage、InvalidMaterial、MissingMaterial、
DuplicateMaterial、BindingMismatch、SignatureMismatch、Rollback、Conflict、
ClearedModel、UnsupportedPolicyShape、PolicyRejected；固定文字为
`statepolicybinding: ` 加对应小写空格文字（例如 `unsupported policy shape`）。
不打印输入。`PolicyFailure struct { Code statepolicy.Code; Reason statepolicy.Reason }`
的 Error 固定同 ErrPolicyRejected，Unwrap 返回 ErrPolicyRejected；不携带 raw
error 或秘密；方法签名为 `func (e *PolicyFailure) Error() string` 和
`func (e *PolicyFailure) Unwrap() error`。mapError 用 errors.Is/As 保留已知类别/枚举，未知→InvalidInput，
当前 ctx.Err 优先；下层 Canceled/Cancellation 且当前无错误时→context.Canceled。

## 固定环境、基线及暂停条件

每个执行测试的新 PowerShell 会话设置以下进程级环境；不修改永久配置：

```powershell
$env:GOENV='off'; $env:GOWORK='off'; $env:GOTOOLCHAIN='local'
$env:GOPROXY='off'; $env:GOSUMDB='off'; $env:GOAUTH='off'
$env:GOVCS='all:off'; $env:GOFLAGS='-mod=readonly'
$env:GOPATH='C:/Users/Lenovo/go'; $env:GOMODCACHE='C:/Users/Lenovo/go/pkg/mod'
$policyBindingGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
$policyBindingFmt='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/gofmt.exe'
```

规划时两工具存在，源码范围 448 份无上轮冻结漂移，HEAD 为
`6752b9612a00b9d6b0bf87ae8dd95063f6b056ee`，索引 SHA256 为
`1CBA46BEF4C1AFDC9AE7938146909F0D2D641A0D712CE18036B60FE5B8673D12`。
实施前重新读取实际状态，apply_patch 保存 baseline.json、progress.md、
各命令原始输出/退出码；冻结范围包括 Go/SQL/module/lock 及既有向量。
当前未提交内容不是本轮可覆盖的基线。仅六份既有测试允许变化，其他旧
文件哈希须保持，新增目录不得有未知重叠。Git 诊断只读、不写索引。

若需改变设计/下层产品/依赖、接入生产/外部系统或扩大测试，先暂停报告
范围问题；未授权不越界。环境/额度/工具拒绝不是测试 RED，不绕过限制。
真实测试失败保留阻塞，按证据排查本范围原因，不增加截止时间冒充通过。
每个任务记录已完成步骤、已知代价及仍未验证责任；不做模板 commit。

## Task 1: 完整有界 owner 与六份精确消费者边界

**Files:** 创建 doc.go/input.go/errors.go/budget.go/owned.go、model.go 的类型及 Clear，
owned_test.go/boundary_test.go；只修改六份已列旧 boundary_test.go。基线/台账随此任务建立。

**Interfaces:** 消费现有完整输入/值类型；产生本计划冻结类型、错误、Model.Clear，以及：

```go
func checkBudget(in Input) error
func bindingBudget(in statebinding.Input) (rawTotal int, err error)
func ownInput(ctx context.Context, in Input) (out Input, err error)
func ownBinding(ctx context.Context, in statebinding.Input) (statebinding.Input, error)
func ownPolicy(ctx context.Context, in statepolicy.ModelInput) (statepolicy.ModelInput, error)
func normalizeTime(v time.Time) time.Time
func clearInput(in *Input)
func clearBinding(in *statebinding.Input)
func clearPolicy(in *statepolicy.ModelInput)
func clearSummary(s *Summary)
```

- [x] 保存最新基线/progress，记录 Native、本机未提交检查点、部分验收限制；确认路径没有未知重叠，保留原 448 份进度。
- [x] 写 TestPolicyBindingBudget：Global Constraints 每个长度/raw 上限独立 equal/+1，包括十九个 L0 集合、expectation 非活动分支、Old 每个 root keys、History 字符串和重复 active raw；无格式要求，assert 形状如下。另测预取消优先于超限、nil context InvalidInput。

  ```go
  if err := checkBudget(atLimit); err != nil { t.Fatal(err) }
  if err := checkBudget(overLimit); !errors.Is(err, ErrLimitExceeded) { t.Fatal(err) }
  got, err := ownInput(canceledContext, overLimit)
  if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(got, Input{}) { t.Fatal(err) }
  ```

- [x] 写 TestPolicyBindingOwnedInput/TestPolicyBindingAliases/TestPolicyBindingNormalization：修改调用者全部 raw/数组/Old 根 public key、keys/ledger/history 不影响 owned；故意共享两表/Batch/active-root 底层存储，修改一个 owned 副本不影响其余。clearInput 清零 held 自有别名，不改调用者；两表顶层排序、nil/empty 等价，链/History/有序域保留；时间 UTC/去单调且纳秒不变、零值不补齐。旧/新 metadata 同 P 不同 E 分表独立拥有。
- [x] 写 TestPolicyBindingBoundary 及六个旧正反例测试：TestBindingPolicyBindingBoundary、TestPolicyBindingConsumerBoundary、TestStateverifyPolicyBindingBoundary、TestContractPolicyBindingBoundary、TestSubsetPolicyBindingBoundary、TestCompositionPolicyBindingBoundary。精确目录/包名 true；子目录、相似前缀、错误包名、原 reader/serving 等 false；新包直消费 stateverify false。保留各旧生产能力白名单和清理守卫。
- [x] 运行下命令记录缺接口/预期守卫断言 RED；路径、环境或依赖失败不算 RED。

  `& $policyBindingGo test ./internal/nodecontrol/statepolicybinding ./internal/nodecontrol/statebinding ./internal/nodecontrol/statepolicy ./internal/nodecontrol/stateverify ./internal/nodecontrol/statecontract ./internal/nodecontrol/statematerials ./internal/nodecontrol/statecomposition -run '^(TestPolicyBindingBudget|TestPolicyBindingOwnedInput|TestPolicyBindingAliases|TestPolicyBindingNormalization|TestPolicyBindingBoundary|TestBindingPolicyBindingBoundary|TestPolicyBindingConsumerBoundary|TestStateverifyPolicyBindingBoundary|TestContractPolicyBindingBoundary|TestSubsetPolicyBindingBoundary|TestCompositionPolicyBindingBoundary)$' -count=1 -timeout 120s -p=1`

- [x] 实现 budget/own/clear/type/error：完整预算在任何复制/索引前完成；Binding 借鉴原 owned.go，复制十九个 L0 集合及两表/锚/links，不导出或改下层私有 helper；Policy 完整复制二维 raw、Old 嵌套 keys 和历史。全部 time.Time 经 normalizeTime：任何 `v.IsZero()` 瞬间规范化为 `time.Time{}`（包括零瞬间的其他时区表示），其他值用 `v.Round(0).UTC()`，不补时间或丢纳秒。部分复制立即有 defer 清理，取消时不返回部分 Input。
- [x] 修改六份旧消费者判断：statepolicy 新增 test-only `policyConsumerAllowed(path, pkg, imported string) bool`，扫描处计算 repo-relative path；其他 helper 保留当前三参数签名。只追加精确 `internal/nodecontrol/statepolicybinding`/`statepolicybinding`，stateverify 的例外只针对其 statepolicy 检查分支，原 Verify 权限不变。
- [x] 实现新 test-only import/消费者守卫：`policyBindingImportsAllowed(source []byte) bool`、`policyBindingConsumerAllowed(path, pkg, imported string) bool`，产品成员白名单见下节。
- [x] 重跑完全相同 targeted 命令 GREEN，再分别完整测试本任务七包（按最终矩阵命令），逐条 exit 0，保存检查点/旧差异六份，不提交。

### 新包固定能力白名单

标准库：context 的 Context/Canceled/DeadlineExceeded；errors 的 New/Is/As；
bytes 的 Clone/Compare/Equal；sort.Slice；reflect.DeepEqual；time.Time/Unix；
contracts.Digest。仅允许 time 值的零值/比较/UTC/Round/Unix/Year 操作；不导入
uuid（NodeID 使用下层 UUID 值的 String），不增加时间采样或控制 context 的能力。

statebinding 只允许 Input/Summary/Evaluate/ConditionalProjectionBound 及 11 个
既有固定错误；statepolicy 只允许 ModelInput/StateModel/RootRecord/KeyRecord/
HighWater/DeploymentObservation/TimeObservation/KeyHistoryObservation/Candidate/
EvaluateModel/Available/Unknown/PolicyError/Code/Reason/Canceled/Cancellation。
不允许下层 RecheckModel、observer 类型或其他判定入口。

statecontract 只允许复制/投影所需的 MainStateContentModel/OptionalTime/
ArtifactRef/ArtifactKind/Coverage/Present/NeverEstablished/RootPayload/RootRotation/
Metadata/DesiredSnapshot；statematerials 只允许 Input/Scope/Material/Expectation/
MaterialRef/MaterialKind/RootSet/RootRotation/Metadata/Desired；statecomposition
只允许 Summary/RotationLink/Coverage。stateformat 只允许 ParseRootSet/
ParsedRootSet/RootMaterial/PublicKeyMaterial/ParseEnvelope/ParsedArtifact/Metadata/
MetadataMaterial/MetadataKeyMaterial。成员返回的 owner 仅用既有事实访问及 Clear。

若实现可不使用上述某成员，保持子集；不因编译方便放行整包。类型/方法
之外的下层能力、新 import 和 direct Evaluate 都须停止并报告，不放宽白名单。

## Task 2: 所选 Batch、根身份映射与当前根后缀

**Files:** 创建 selection.go/roots.go、fixtures_test.go/selection_test.go/roots_test.go。

**Interfaces:** 消费 Task 1 Input/固定错误；产生：

```go
func projectRef(ref statecontract.ArtifactRef) (statematerials.MaterialRef, error)
func findMaterial(ctx context.Context, table statematerials.Input, ref statematerials.MaterialRef) (statematerials.Material, error)
func findExpected(ctx context.Context, table statematerials.Input, ref statematerials.MaterialRef) (statematerials.Expectation, error)
func bindSelectedScope(ctx context.Context, in Input, proof statebinding.Summary) error
func bindBatch(ctx context.Context, in Input) error
type rootFact struct { projection RootProjection; material stateformat.RootMaterial }
type rootPath struct { items []rootFact }
func buildRootPath(ctx context.Context, in Input) (*rootPath, error)
func (path *rootPath) Clear()
func bindRoots(ctx context.Context, in Input, path *rootPath, state *statepolicy.StateModel, proposed bool) (RootProjection, error)
func bindRootTransition(ctx context.Context, in Input, path *rootPath) error
```

这些 helper 消费刚通过完整 Binding 的自有输入；find 返回借用，不清理
借用 Material。buildRootPath 独立拥有解析所得 RootMaterial；Clear 清零
items/material，失败路径不泄露部分 path。bindRoots 的 proposed=false 绑定
Old 前缀/旧 Main，true 绑定完整新链/新 Main，不排序根或猜测其位置。

测试 helper 固定 `policyBindingFixture(t *testing.T, rotations, previousRotations int) Input`、
`copyPolicyBindingInput(t *testing.T, in Input) Input`、
`checkBindingInput(t *testing.T, in Input) statebinding.Summary`、
`checkPolicyInput(t *testing.T, in Input)`，仅在 fixtures_test.go。

- [x] 建立独立 fixtures：只读复用 stateverify/testdata 冻结向量，固定测试 seed 签名其他样例；node=`12345678-1234-4234-8234-123456789abc`、epoch=19、Now=`2020-01-01T01:00:00Z`、OldFloor=`2019-12-31T00:00:00Z`。rotations/previousRotations 为各自轮换次数，0≤previous≤rotations≤32；相等时业务 no-change；前进时 mint 新 metadata version/sequence=7/6 和 desired generation/sequence=8/10，旧基线=6/5、7/9，root 各自身序列不跨流比较。补齐十二域 L0 与旧表/新表，Main.RootTransition 只含本次后缀。Old 从测试侧原 payload/固定公钥独立构造，不从待测上层或下层 Candidate.State 反推期望；下层 sanity helpers 只核验，不生成 oracle。测试复制用独立 test-only 深复制，不调用产品 ownInput。
- [x] 写 TestPolicyBindingSelectedScope/TestPolicyBindingBatch：0/1/32 当前目标成功；合法历史 authorizer 先 checkBindingInput 成功，范围 UnsupportedPolicyShape；Old=nil、FreshInstall=true、旧 Desired NeverEstablished、合法主 floor 提高/新建同类拒绝。逐一改变 Batch scope/anchor/每个 rotation/数量/顺序/active/meta/desired，assert ErrBindingMismatch；两份独立合法信封分别通过下层仍不能错绑。

  ```go
  proof := checkBindingInput(t, in)
  if err := bindSelectedScope(context.Background(), in, proof); err != nil { t.Fatal(err) }
  checkPolicyInput(t, replacement) // independently valid Policy.Batch
  if err := bindBatch(context.Background(), replacement); !errors.Is(err, ErrBindingMismatch) { t.Fatal(err) }
  ```

- [x] 写 TestPolicyBindingRootPath/TestPolicyBindingRoots/TestPolicyBindingRootSuffix：v1/轮换两种 SourceKind/P/E 分开；断言 rotation P != NewRoot P；完整 key/public-key/threshold/version/sequence、前缀长度/位置、末端引用错配分别拒绝，32/33 路径边界；Main.RootTransition 与旧根后的 Links 后缀逐项完整一致，空 suffix 不生成转换。额外未选根不得被选作旧基线。

  ```go
  if path.items[1].projection.Source.Payload == path.items[1].projection.RootSet.Payload { t.Fatal("conflated identities") }
  if in.Policy.Old.Roots[1].Mark.Payload != path.items[1].projection.Source.Payload { t.Fatal("wrong policy root identity") }
  _, err := bindRoots(context.Background(), wrongRootState, path, wrongRootState.Policy.Old, false)
  if !errors.Is(err, ErrBindingMismatch) { t.Fatal(err) }
  ```

- [x] 运行 `& $policyBindingGo test ./internal/nodecontrol/statepolicybinding -run '^(TestPolicyBindingSelectedScope|TestPolicyBindingBatch|TestPolicyBindingRootPath|TestPolicyBindingRoots|TestPolicyBindingRootSuffix)$' -count=1 -timeout 120s -p=1`，记录预期缺 helper/映射断言 RED。
- [x] 实现 selection.go：封闭四类 ref 映射；Kind/P 缺失→MissingMaterial，已有但 E 不符→BindingMismatch；范围检查在完整 Binding 之后，只核对 Binding 顶层成功状态及 summary 的作用域/目标/FinalRoot 与自有输入，不创造或升级子层证明。按设计 §6 原字节逐项比较，不生成替代 Batch、hash、证明或重试。
- [x] 实现 roots.go：anchor RootSet 和 Links.NewRoot 分别通过 ParseRootSet，立即登记解析 owner/material 清理；Source 初始取 Anchor P/E，其他取 Rotation ref P/E。逐项 binds policy Mark（Source P）及完整材料，再将对应 RootSet/版本/序列绑定 Main，后缀同时绑定 rotation ref/expectation 的 old/new root/版本/序列。path 只暂存 root materials，不携带授权。
- [x] 重跑同 targeted GREEN，再完整新包和六守卫包，各 exit 0；记录新目标限制只在上层，不改旧历史能力，不提交。

## Task 3: 旧/新当前 key 投影、三种 floor 与完整 Evaluate

**Files:** 创建 projection.go/time.go/evaluate.go、projection_test.go/time_test.go/evaluate_test.go；完善 errors.go。

**Interfaces:** 消费 Task 1 owner/Clear、Task 2 scope/Batch/root helpers、现有真实下层入口；产生：

```go
func bindStateProjection(ctx context.Context, table statematerials.Input, main statecontract.MainStateContentModel, state *statepolicy.StateModel) error
func bindKeyProjection(ctx context.Context, table statematerials.Input, metadata statecontract.ArtifactRef, state *statepolicy.StateModel) error
func bindStateIdentity(ctx context.Context, in Input, state *statepolicy.StateModel, proposed bool) error
func bindTime(ctx context.Context, in Input) error
func mapError(ctx context.Context, err error) error
func makeSummary(in Input, binding statebinding.Summary, prior, next RootProjection, nextState *statepolicy.StateModel, noChange bool, validUntil time.Time) Summary
func Evaluate(ctx context.Context, in Input) (result *Model, err error)
```

makeSummary 只复制已受保护的摘要和值，不再申请/解析材料、不调用 context。
带命名返回值的 Evaluate 公开调用类型不变，用 result==nil 的 defer 判定
未转移所有权；Task 4 再开放 Summary/Recheck。

- [x] 写 TestPolicyBindingProjection/TestPolicyBindingHistoricalKeys：旧/新 metadata/desired 全高水位匹配；当前每个 key 的完整 public-key/状态/时间/撤销/incident 必须唯一相等，ledger 完整有序相同；合法额外 active 历史记录/retained tombstone 不删除。缺当前 key、同数量换 key、仅状态或有效窗变化、ledger 冲突各 assert BindingMismatch；Old 的 deadline/历史字段只完整保留并交 policy，不假造已认证来源。
- [x] 写 TestPolicyBindingTime/TestPolicyBindingIdentity：Main floor 缺席/存在且不变；OldFloor=约束或更高、ObservedFloor=Old 或更高、Now=Observed 或更高合法；OldFloor 低于主约束→BindingMismatch；观测低于 Old/Now 低于观测沿 policy 枚举拒绝。provider、node、epoch、anchor 错配，nil/零/越界时间、纳秒、checked 范围、半开窗口和重复过期不输出模型。合法 Main floor 提高 UnsupportedPolicyShape；降低/清除必须先保留 Binding.Rollback。选择互不相等的 DeploymentID/SourceID、TrustAnchorID/anchor、main guard/Policy.CounterID 仍可成立，明确这些 ID 没有转换契约。
- [x] 写 TestPolicyBindingEvaluate/TestPolicyBindingPolicyFailure/TestPolicyBindingErrorMapping：0/1/32 no-change 和后缀更新；两个独立下层合法的错绑 Batch→nil/BindingMismatch。旧 metadata 坏签名但静态格式合法不被升级成已验签；当前目标坏签名仍失败。每个历史负例先证明 Binding 合法、独立 policy 返回预期 Code/Reason，避免其他签名/时间错误遮蔽；覆盖 missing history、scheduled deadline/tombstone、SecurityFault/EmergencyRevocation。固定诊断不回显动态输入，wrapped sentinel/PolicyError、未知错误、nil/取消/超时映射均明确；typed-nil PolicyError 必须 InvalidInput、不解引用 panic，当前取消仍优先。

  ```go
  model, err := Evaluate(context.Background(), in)
  if want == nil {
      if err != nil || model == nil { t.Fatal(err) }; defer model.Clear()
      if model.summary.Status != ConditionalSelectedPolicyBound { t.Fatal("wrong status") }
  } else if model != nil || !errors.Is(err, want) { t.Fatal(err) }
  var pf *PolicyFailure
  if !errors.As(policyErr, &pf) || pf.Code != wantCode || pf.Reason != wantReason { t.Fatal(policyErr) }
  ```

- [x] 运行 `& $policyBindingGo test ./internal/nodecontrol/statepolicybinding -run '^(TestPolicyBindingProjection|TestPolicyBindingHistoricalKeys|TestPolicyBindingTime|TestPolicyBindingIdentity|TestPolicyBindingEvaluate|TestPolicyBindingPolicyFailure|TestPolicyBindingErrorMapping)$' -count=1 -timeout 120s -p=1`，记录预期缺接口/组合断言 RED。
- [x] 实现 projection/time：metadata 精确 Raw 经现有 ParseEnvelope(Metadata)，核对 P/E 并立即清理 parser/material；当前 key 逐项唯一全字段比较，允许额外历史 key；高水位精确投影。TimeProviderID 四方相等，主 floor UnixSeconds 在 `[-62135596800,253402300799]` 后 time.Unix(...,0).UTC；仅约束 OldFloor，不更新三种 floor。state identity 比较 node/epoch/anchor/source/provider；新状态须 Unknown 且物理 ID/StateEpoch/TimeFloor 仍等 Old，不擅自提升可信性。
- [x] 实现 Evaluate：ownInput→Model 接管并立即 defer(result==nil 时 Clear)→Binding.Evaluate/立即 defer Clear→Summary 直接放入受保护字段→scope→Batch→buildRootPath/立即 defer Clear→旧 identity/roots/transition/projection/time→Policy.EvaluateModel/立即 defer Clear→State()/立即 defer Clear→新 identity/roots/projection→makeSummary→最终 context→valid/return。下层 Clear 登记先于 mapError 的 context 回调；失败不返回部分 Model，不保留下层 owner。
- [x] 完善 mapError/PolicyFailure：errors.As 成功后先检查非 nil，再复制已知 policy 枚举，不保留 error 指针；错误文字固定。
- [x] 重跑同 targeted GREEN，再独立完整新包及六守卫包，各 exit 0，记录此时仍无公开运行/提交接口，不提交。

## Task 4: 条件性摘要、整个规范化输入重检及取消/panic

**Files:** 完善 model.go/evaluate_test.go，创建 model_test.go；复用 owned_test.go/boundary_test.go。

**Interfaces:** 消费 Task 3 私有 input/summary；产生公共 Summary/Recheck 及
`func cloneSummary(in Summary) (out Summary)`。cloneSummary 独立复制 Binding
摘要的 Links/Coverage，建立失败清理；不修改内部有效模型。

- [x] 写 TestPolicyBindingSummary/TestPolicyBindingNoWholeStateChange：逐字段核对固定 Summary、RootProjection 的两种身份、所有下层八项/嵌套责任及附加三项。只增加 Main.Proposed.Checkpoint.Sequence 且其余合法时，SelectedPolicyNoChange 可仍为 true，但输入明显非十二域无变化，摘要不得提供 WholeStateNoChange/Ready。nil/零/cleared Summary 为完整零值。
- [x] 写 TestPolicyBindingModelOwnership/TestPolicyBindingModelClear：改变调用者各可变输入、每次 detached Links/Coverage 不影响模型；保留内部 raw/集合/历史/根公钥及摘要别名，Clear 两次均零、独立摘要/调用者不变；不直接观测未返回副本或字符串擦除。
- [x] 写 TestPolicyBindingRecheck/TestPolicyBindingUnmappedRecheck/TestPolicyBindingFixedTime：完整相同输入、两表顶层重排、nil/empty、等瞬时时区通过；合法时间向前 1ns、floor 提高、ObservationID、部署 ID、Old.StateID/CounterID/StateEpoch、History 顺序/内容、合法未选 raw/expectations、未投影域/metadata 历史和集合摘要改变均拒绝旧模型。每个“合法替换”先验证新 Evaluate 成功，不以更早格式失败冒充重检身份覆盖。

  ```go
  fresh, err := Evaluate(context.Background(), replacement)
  if err != nil || fresh == nil { t.Fatal(err) }; defer fresh.Clear()
  if err := Recheck(context.Background(), prior, replacement); !errors.Is(err, ErrBindingMismatch) { t.Fatal(err) }
  ```

- [x] 写 TestPolicyBindingCancellation/TestPolicyBindingLateCancellation/TestPolicyBindingPanicCleanup：test-only 计数 Context.Err 按实际成功调用轨迹逐点取消/超时；Summary 捕获后及 Recheck 比较后的最终点覆盖。普通 panic 必须传播、调用者不变、模型 nil；与持有 backing-store 的 Clear 测试及下一任务 AST 早清理检查分别报告证据。Context 轨迹数量来自本次运行，不硬编码下层调用次数、不新增产品测试 hook。
- [x] 运行 `& $policyBindingGo test ./internal/nodecontrol/statepolicybinding -run '^(TestPolicyBindingSummary|TestPolicyBindingNoWholeStateChange|TestPolicyBindingModelOwnership|TestPolicyBindingModelClear|TestPolicyBindingRecheck|TestPolicyBindingUnmappedRecheck|TestPolicyBindingFixedTime|TestPolicyBindingCancellation|TestPolicyBindingLateCancellation|TestPolicyBindingPanicCleanup)$' -count=1 -timeout 120s -p=1`，记录缺 accessor/Recheck 的 RED。
- [x] 实现 Summary/cloneSummary：只在 valid 时独立返回，否则零值；保持全部义务。Recheck nil/取消/清理检查→完整新 Evaluate→立即 defer fresh.Clear→reflect.DeepEqual 整个规范 input→最终 context→不同 BindingMismatch，不调用下层宽松时间 RecheckModel、不只比摘要。
- [x] 重跑同 targeted GREEN 和完整新包，exit 0；检查无额外生产 API、未抹掉来源/历史/全域责任，保存阶段证据，不提交。

## Task 5: 精确能力/清理回归、十五项验证及一次 Native 复核

**Files:** 完善新 boundary_test.go；维护基线/progress/freeze，实施后创建执行记录并更新 current-status/批准设计/计划状态；不新增产品行为。

**Interfaces:** 消费任务 1–4 冻结 API；test-only
`func policyBindingCleanupRegistered(source []byte) bool`、
`func policyBindingImportsAllowed(source []byte) bool`、
`func policyBindingConsumerAllowed(path, pkg, imported string) bool`。

- [x] 写 TestPolicyBindingCleanupRegistration/TestPolicyBindingProductImports/TestPolicyBindingConsumers：AST 要求拥有输入后立即登记 Model 清理，lower/路径/State()/parser/material owner 的 Clear 先于后续 context/判定，摘要写入受保护的 Model；late/missing/conditional/wrong-owner defer、早 ownership transfer 均 false。守卫检查固定支持结构，不声称完整控制流证明；全非测试 Go 消费扫描无本包产品消费者，原精确例外不扩散到子目录/其他能力。
- [x] 运行 `& $policyBindingGo test ./internal/nodecontrol/statepolicybinding -run '^(TestPolicyBindingCleanupRegistration|TestPolicyBindingProductImports|TestPolicyBindingConsumers)$' -count=1 -timeout 120s -p=1`，记录缺 AST helper 或故意错误结构的 RED，不得省略已由下层蕴含的检查来制造假 mutation RED。
- [x] 实现 test-only 最小 AST 清理 helper，其他已建立的能力/消费者 helper 保持任务 1 契约。
- [x] 重跑同一 Task 5 targeted 命令，要求 GREEN 并保存实际退出码。
- [x] 六个旧边界正反例重跑任务 1 targeted 命令，保存 exit 0。严格只格式化二十个新 Go 文件和六旧测试，先核对数组恰为 26，gofmt -w 实际 exit 0，再独立 gofmt -l 实际 exit 0 且空输出。

  ```powershell
  $policyBindingFormatFiles = @((Get-ChildItem -LiteralPath 'internal/nodecontrol/statepolicybinding' -Filter '*.go').FullName) + @(
      'internal/nodecontrol/statebinding/boundary_test.go', 'internal/nodecontrol/statepolicy/boundary_test.go',
      'internal/nodecontrol/stateverify/boundary_test.go', 'internal/nodecontrol/statecontract/boundary_test.go',
      'internal/nodecontrol/statematerials/boundary_test.go', 'internal/nodecontrol/statecomposition/boundary_test.go'
  )
  if ($policyBindingFormatFiles.Count -ne 26) { throw 'Unexpected formatting scope' }
  & $policyBindingFmt -w $policyBindingFormatFiles
  ```

- [x] 独立 `& $policyBindingFmt -l $policyBindingFormatFiles`，核对所有旧文件仅六批准例外、HEAD/分支/索引不变、只读 diff-check exit 0；保存 post-format-source-test-freeze.json。新路径文件数/名字超范围先报告，不放宽数组门槛。
- [x] 逐条独立运行下面十五条固定命令，保存原始输出/实际退出码/开始结束；全部 exit 0，所有会话确实结束，冻结无漂移后才复核。单包 Go 实际 timeout 120s，不把工具输出发布延迟当作超时，也不重复启动 pending 命令。

| 范围 | 固定命令 |
| --- | --- |
| statepolicybinding | `& $policyBindingGo test ./internal/nodecontrol/statepolicybinding -count=1 -timeout 120s -p=1` |
| statebinding | `& $policyBindingGo test ./internal/nodecontrol/statebinding -count=1 -timeout 120s -p=1` |
| statecontract | `& $policyBindingGo test ./internal/nodecontrol/statecontract -count=1 -timeout 120s -p=1` |
| statematerials | `& $policyBindingGo test ./internal/nodecontrol/statematerials -count=1 -timeout 120s -p=1` |
| statecomposition | `& $policyBindingGo test ./internal/nodecontrol/statecomposition -count=1 -timeout 120s -p=1` |
| stateformat | `& $policyBindingGo test ./internal/nodecontrol/stateformat -count=1 -timeout 120s -p=1` |
| stateverify | `& $policyBindingGo test ./internal/nodecontrol/stateverify -count=1 -timeout 120s -p=1` |
| statepolicy | `& $policyBindingGo test ./internal/nodecontrol/statepolicy -count=1 -timeout 120s -p=1` |
| staterecovery | `& $policyBindingGo test ./internal/nodecontrol/staterecovery -count=1 -timeout 120s -p=1` |
| statetime | `& $policyBindingGo test ./internal/nodecontrol/statetime -count=1 -timeout 120s -p=1` |
| serving | `& $policyBindingGo test ./internal/nodecontrol/serving -count=1 -timeout 120s -p=1` |
| contracts | `& $policyBindingGo test ./internal/nodecontrol/contracts -count=1 -timeout 120s -p=1` |
| readiness | `& $policyBindingGo test ./internal/readiness -count=1 -timeout 120s -p=1` |
| authority（限定） | `& $policyBindingGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1 -timeout 120s -p=1` |
| store（仅编译） | `& $policyBindingGo test ./internal/store -run '^$' -count=1 -timeout 120s -p=1` |

- [x] 沿用 Native：作者完成五任务，无实现代理/逐任务独立复审。验证后只派一次全新独立上下文、只读最强可用源码 reviewer（当前可用 gpt-6-astra / xhigh）审查本增量及六边界；提供已批准设计/计划、五项 Review Focus、旧基线/实际差异、冻结、RED/GREEN及十五实际结果、全部未完成义务。reviewer 仅只读查看范围内文件、Git 差异和已保存证据，不改文件、不运行测试或额外诊断、不联网、不再委派。只审本增量，不假称全脏分支/项目安全验收；本轮计划撰写不派 reviewer。
- [x] 作者逐项裁定 Critical/Important/Minor，Critical/Important 在范围内最小修复并重跑十五条；范围外先暂停。Minor 最小修复或记录保留理由和代价；产品变化重跑全部，纯测试至少重跑受影响包及六守卫。无第二次独立审查，作者修复后的版本独立覆盖不会因此扩大；任何遗留阻断须明确保留。
- [x] apply_patch 更新执行记录、真实 checkbox、progress、current-status，逐项列实际结果、审查裁定/限制及平台来源、旧物理快照关联、时间证明/全历史、其他域、完整闭包/组合、提交/运行和 C12/main 未验收。最终源哈希/分支/HEAD/索引核对；保留全部既有修改/worktree/scratch，不提交、推送、合并或清理。

## 作者自查及执行交接

设计 §1–2/§10 → 任务 1/2/5；§3–4 → 任务 1/4；§5 → 任务 3/5；
§6 → 任务 2/3；§7.1 → 任务 3/4；§7.2 → 任务 2/3；§7.3–8 → 任务 3/4；
§9 → 任务 3/4；§11 → 五个 RED/GREEN 循环和十五条矩阵；§12 → 本轮文档
及用户审阅门。五项 Review Focus 均有明确命名测试，跨任务接口/字段/状态
名一致，二十个新 Go 文件/六处旧例外与格式化范围一致。预算与支持范围
不改变下层语义，没有独立 verifier/时间证明、codec 或生产消费者。

自查是作者的文档检查，不是实现、测试或独立审查。用户批准本计划后才
使用 executing-plans 按已选 Native 开始任务 1；无需重新选择方法。
计划批准也不授权 Git 集成、其他安全域或真实平台验收；本机文档尚未提交
推送，不构成跨电脑恢复检查点。
