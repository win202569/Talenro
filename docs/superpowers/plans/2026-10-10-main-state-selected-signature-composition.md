# Main-state Selected-signature Composition Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让完整材料子集检查与单个选中 desired 的条件性验签使用同一份自有快照，并能完整重检，不产生生产授权。

**Architecture:** 新增隔离纯包 statecomposition，调用 statematerials.Evaluate 和 stateverify.Verify；独立锚只由 stateformat 解析，所有验签字节均来自本次自有输入。逐项回核证据，明确区分显式锚、链导出的根、相对锚验签的制品与仅材料一致的历史。下层产品实现不变，仅增加两处精确的消费边界测试例外。

**Tech Stack:** 固定 Go 1.26.5、现有 contracts/stateformat/statematerials/stateverify、PowerShell、离线缓存；无新依赖。

**Spec:** `docs/superpowers/specs/2026-10-10-main-state-selected-signature-composition-design.md`（用户于 2026-10-10 批准）。执行前完整阅读设计和本计划。

状态：用户于 2026-10-10 批准本计划；Native 本地五项任务完成，全部限定回归通过，唯一一次独立增量审查为 0 Critical / 0 Important / 0 Minor。最终只读核对确认源码/测试冻结、基线例外及 HEAD/分支/索引均符合批准范围。不写 Git 索引、不提交或推送；所有既有进度保留，整体验收仍未完成。

## Global Constraints

- 工作目录仅 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`；复用现有分支，保留已有未提交内容。
- 不暂存、提交、推送、合并、重置、删除分支、清理或新建 worktree；记录步骤替代技能模板中的 commit。
- 材料与预期记录各最多 128 项；全部材料原始字节最多 16 MiB；材料、预期中的轮换和目标有序链接各最多 32 项。
- RootSet raw ≤65,536 bytes；其他材料和独立锚 raw ≤262,144 bytes；独立锚另计，全部外来 raw 总上限 16 MiB + 262,144 bytes。
- 深复制前检查所有预期分支集合：root keys 5、metadata keys 64、revoked IDs 4096、desired slots 8；预算不是峰值内存承诺。
- RootSet 的 E=0，签名制品的 P/E 非零；独立 InitialRoot P/E 非零，不能把锚 E 填进 RootSet ref。
- 只选一个 desired；v1/单 root key/threshold 1 的显式锚起，最多 32 次连续轮换；不自动选锚、补链、排序链、缩短链重试或接受外来缓存证据。
- 下层产品 API、wire、签名算法与静态窗口/撤销规则不变；statecontract 隔离守卫不变，无生产调用者。
- 不读取当前时间、磁盘、网络、环境、随机数或硬件；产品无 JSON/hash/crypto/goroutine/provider；测试及证据记录可读取本地文件。
- 当前时间、锚来源、未覆盖历史、完整主状态闭包/组合、物理提交及运行授权均不在本批证明范围；结果只能 ConditionalTargetVerified。
- 全部失败返回 nil 模型，固定错误不回显材料；输入调用期间不可并发修改，模型不可浅复制或并发 Clear/访问。
- 仅运行本计划列出的离线包、authority 两类测试和 store 编译；不跑全仓库/真实 DB/Docker/race/硬件/重启/C11/C12/PITR。
- 手工编辑仅 apply_patch；格式化可用固定工具链 gofmt。不得绕过既有安全限制或宣称生产兼容性。

## Review Focus

以下五种接口使用情形各补入指定任务的断言，不让合法输入变化变成无声的证据升级：

1. 同 P、不同 E 的合法信封：选择必须命中本批完整 ref，已有模型重检不能接受替换（任务 2、4）。
2. 原声明与选择保持一致但一起合法换成兼容 metadata：新 Evaluate 可成立，不代表历史绑定；旧模型 Recheck 必须拒绝（任务 3、4）。
3. 未选中材料仍有效且可合法替换：它不得绕过材料检查，也不得绕过完整快照重检（任务 3、4）。
4. 最后一阶段才取消、或 Context.Err 在已拥有输入后 panic：不能交付模型，清理要早注册，不修改调用者存储（任务 4、5）。
5. 伪装成允许路径/包名的消费者及重复别名：只允许精确两处例外，绝不能给 reader、子目录、相似前缀或任意反射能力放行（任务 1、5）。

## 文件分工与固定公共 API

新目录 `internal/nodecontrol/statecomposition/`：`doc.go`（证据边界）、`input.go`（输入）、`errors.go`（错误映射）、`owned.go`（预算与完整自有快照）、`selection.go`（严格选择与 Batch）、`evidence.go`（回核）、`evaluate.go`（组合入口）、`model.go`（摘要/清理/重检）。

测试：`owned_test.go`、`fixtures_test.go`、`selection_test.go`、`evaluate_test.go`、`model_test.go`、`boundary_test.go`。不新建 wire fixture 生成器；只读复用 stateverify/testdata，变体构造仅在新测试文件。
既有文件仅修改 `internal/nodecontrol/statematerials/boundary_test.go` 和 `internal/nodecontrol/stateverify/boundary_test.go` 的精确消费例外及反例。其余 Go/SQL/模块/锁文件均受基线保护。

```go
type AnchorInput struct {
    Scope statematerials.Scope
    Payload, Envelope contracts.Digest
    Raw []byte
}
type RotationLink struct { Rotation, NewRoot statematerials.MaterialRef }
type Target struct {
    AnchorRoot statematerials.MaterialRef
    Links []RotationLink
    Metadata, Desired statematerials.MaterialRef
}
type Input struct {
    Materials statematerials.Input
    Anchor AnchorInput
    Target Target
}
type ResultStatus uint8 // 0 InvalidResult; 1 ConditionalTargetVerified
type ProofStatus uint8 // 0 InvalidProof; 1 MaterialOnly; 2 AnchorAssumed;
                       // 3 ChainDerived; 4 SignatureChecked
type Coverage struct { Ref statematerials.MaterialRef; Status ProofStatus }
type Obligation uint8 // 0 UnknownObligation; then the seven below in this order
// ExternalAnchorProvenance, UnselectedSignaturesAndHistory, DomainAndCurrentTime,
// FullStateMaterialClosure, FullStateComposition, PhysicalCommit, RuntimeAuthorization
type Summary struct {
    Status ResultStatus
    Scope statematerials.Scope
    Subset statematerials.Summary
    Anchor stateverify.ArtifactIdentity
    Target Target
    FinalRoot statematerials.MaterialRef
    Coordinates stateverify.EvidenceCoordinates
    Coverage []Coverage
    Obligations [7]Obligation
}
func Evaluate(ctx context.Context, in Input) (*CompositionModel, error)
func Recheck(ctx context.Context, model *CompositionModel, in Input) error
func (model *CompositionModel) Summary() Summary
func (model *CompositionModel) Clear()
```

所有者 CompositionModel 字段私有，固定为 `valid bool`、`input Input`、`subset statematerials.Summary`、`coordinates stateverify.EvidenceCoordinates`、`finalRoot statematerials.MaterialRef`；不保存下层临时所有者。Summary 从这些自有值生成，Target.Links/Coverage 每次独立复制；Subset 中既有固定未完成责任保持不变。

错误在 errors.go 定义八个固定 sentinel：ErrInvalidInput、ErrLimitExceeded、ErrInvalidMaterial、ErrMissingMaterial、ErrDuplicateMaterial、ErrBindingMismatch、ErrSignatureMismatch、ErrClearedModel；文字固定以 `statecomposition: ` 开头，不附加输入值。

私有 `mapError(ctx context.Context, err error) error` 首先保留实际 ctx.Err()；随后按 errors.Is 映射：

| 下层错误 | 本层错误 |
| --- | --- |
| statematerials.ErrInvalidInput / ErrLimitExceeded / ErrInvalidMaterial / ErrMissingMaterial / ErrDuplicateMaterial / ErrClearedModel | 相应同名 sentinel，前述各成员均属于 statematerials |
| statematerials.ErrReferenceMismatch 或 stateverify.ErrBindingMismatch | ErrBindingMismatch |
| stateformat.ErrInvalidFormat 或 stateverify.ErrInvalidInput | ErrInvalidMaterial（组装调用前参数和 Kind 已检查） |
| stateverify.ErrSignatureMismatch | ErrSignatureMismatch |
| context.Canceled / DeadlineExceeded | 相应 context 错误 |
| stateverify.ErrCanceled 且 ctx.Err 尚为空 | context.Canceled，不伪装为验签错误 |
| 其他非 nil 错误 | ErrInvalidInput，不透传动态文本 |
| nil 且未取消 | nil |

本层参数错误（nil context、空锚/零锚摘要、Target 错误 Kind/零摘要/非法 E）是 ErrInvalidInput；合法 scope 不同、锚 P/E/根字节不符、完整 ref 的 E 不同、链/声明/证据不符是 ErrBindingMismatch；合法 ref 的 Kind/P 未命中是 ErrMissingMaterial；重复表记录由材料层映射 ErrDuplicateMaterial，重复目标链接按链关系判 ErrBindingMismatch。格式损坏为 ErrInvalidMaterial。预算超限优先于深复制和格式处理；已有取消优先于预算。

## 固定执行环境与证据记录

每个新 PowerShell 会话先设置以下任务环境，命令独立执行并逐次记录退出码，不修改永久环境或让最后一个命令掩盖前一个失败：

```powershell
$env:GOENV='off'; $env:GOWORK='off'; $env:GOTOOLCHAIN='local'
$env:GOPROXY='off'; $env:GOSUMDB='off'; $env:GOAUTH='off'
$env:GOVCS='all:off'; $env:GOFLAGS='-mod=readonly'
$env:GOPATH='C:/Users/Lenovo/go'; $env:GOMODCACHE='C:/Users/Lenovo/go/pkg/mod'
$compositionGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
$compositionFmt='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/gofmt.exe'
```

首次实现前在 `.superpowers/sdd/2026-10-10-main-state-selected-signature-composition/` 用 apply_patch 保存 HEAD/工作树读数、现有 Go/SQL/模块/锁文件 SHA-256 baseline.json 与 progress.md；每步记录 RED/GREEN、退出码、范围及剩余义务。此 scratch 保留，不清理，不成为提交。
工具输出延迟不等于测试超时；已启动会话必须等结束，不重复启动同一组测试。RED 必须是预期接口缺失/断言失败，环境错误不算 RED；已经正确的行为只记录既有覆盖。

## Task 1: 有界自有输入与精确隔离边界

**Files:** 新建 doc.go/input.go/errors.go/owned.go、owned_test.go/boundary_test.go；修改上述两个既有 boundary_test.go。保存执行前基线及 ledger。

**Interfaces:** 消费原 statematerials.Input/Scope/MaterialRef、contracts.Digest。产生固定输入/错误类型及以下私有函数，供后续任务使用：

```go
func checkBudget(in Input) error
func ownInput(ctx context.Context, in Input) (Input, error)
func clearInput(in *Input)
func compositionImportsAllowed(source []byte) bool // test only
func compositionConsumerAllowed(path, pkg, imported string) bool // test only
// Add the same signature as materialsConsumerAllowed in statematerials tests;
// retain/extend stateverify's existing isolatedImportAllowed.
```

- [x] 记录执行前基线；用 apply_patch 创建 ledger，明确只有两处既有测试允许变化，不改 Git 配置或索引。
- [x] 写 TestCompositionBudget 与 TestCompositionOwnedInput：对每个 128/32/65536/262144/16MiB/5/64/4096/8 上限独立测试 equal 与 +1；equal 仅要求 `checkBudget(in)==nil`，不假称非协议字节能通过 Evaluate。超大非活动分支也须 LimitExceeded；空锚 InvalidInput、预取消优先；每个 raw/嵌套 slice/Links 分别修改后自有副本不变，clearInput 清零保留的自有别名但不改原输入；nil/empty 和顶层表顺序规范化，链顺序不改。

  预算测试的独立锚断言固定为（其他预算用同样的 equal/+1 表驱动，不先进入格式层）：

  ```go
  in := Input{Anchor: AnchorInput{Raw: make([]byte, 262144)}}
  if err := checkBudget(in); err != nil { t.Fatal(err) }
  in.Anchor.Raw = make([]byte, 262145)
  if err := checkBudget(in); err != ErrLimitExceeded { t.Fatal(err) }
  ```

- [x] 写 TestCompositionBoundary、TestSubsetCompositionBoundary、TestStateverifyCompositionBoundary：精确目录/正确包名例外 true；reader、子目录、相似前缀、错误包名、消费 statepolicy/statecomposition false；点/空/重复路径/冲突别名、禁用导出成员/GoStmt 拒绝。核心断言例如 `if compositionConsumerAllowed("internal/nodecontrol/serving/x.go", "serving", compositionPath) { t.Fatal("runtime consumer") }`，测试常量 compositionPath 固定为 `talenro.local/platform/internal/nodecontrol/statecomposition`。
- [x] 运行 `& $compositionGo test ./internal/nodecontrol/statecomposition ./internal/nodecontrol/statematerials ./internal/nodecontrol/stateverify -run '^(TestCompositionBudget|TestCompositionOwnedInput|TestCompositionBoundary|TestSubsetCompositionBoundary|TestStateverifyCompositionBoundary)$' -count=1`，记录 RED。
- [x] 实现 ownInput：检查 context 与 checkBudget 后深复制全部 raw/预期数组/Links，登记失败清理，逐项检查取消，顶层按 Kind/P 排序，所有空列表规范化 nil；clearInput 幂等清零自有可变存储。预算累加采用 `len(raw)>limit-total`，不先分配索引，不重新解释 wire。此任务只交付输入准备，不提供总是成功的 Evaluate scaffold。
- [x] 实现两处精确消费例外及新包 AST 成员白名单。仅允许 context.Context/Canceled/DeadlineExceeded、errors.New/Is、bytes.Clone/Compare/Equal、sort.Slice、reflect.DeepEqual、contracts.Digest；下层允许固定 API 中类型、Evaluate/Verify、相应固定错误，以及 stateformat.ParseEnvelope/InitialRoot/ParsedArtifact/RootMaterial/ErrInvalidFormat。statematerials 允许 Input/Scope/MaterialRef/MaterialKind、四个 Kind、Material/Expectation/MaterialSubsetModel/Summary；stateverify 允许 Batch/SignatureEvidence/EvidenceCoordinates/ArtifactIdentity。不得放开整个标准库包的能力，拒绝重复 import 路径，即使别名不同。
- [x] 重跑 targeted，随后分别完整运行 statecomposition、statematerials、stateverify；全部 exit 0。记录结果和仅两处既有测试差异，不提交。

## Task 2: 独立锚及同快照 Batch 选择

**Files:** 新建 selection.go、fixtures_test.go、selection_test.go。

**Interfaces:** 消费 Task 1 的自有 Input、原层验证过的完整材料表。产生只借用此快照字节的私有 selection：

```go
type selection struct {
    batch stateverify.Batch
    finalRoot statematerials.MaterialRef
    rootVersion, metadataVersion, desiredGeneration, desiredSequence uint64
    desiredKeyID contracts.Digest
}
func selectBatch(ctx context.Context, in Input) (*selection, error)
func (selected *selection) Clear() // clears outer refs/fields, not borrowed raw
```

selectBatch 的前置条件是 in 已自有且刚通过 statematerials.Evaluate。索引以 (Kind,P) 为唯一键，查中后必须比较完整 E；索引值引用同快照 Material 和 Expectation，不能读取输入外对象。
测试 helper 固定 `fixture(t *testing.T, rotations int) Input`、`requireOutcome(t *testing.T, model *CompositionModel, err, want error)`（Task 3 加入后一函数）：错误用 errors.Is，失败模型 nil，成功非 nil 并注册 Clear。

- [x] 创建测试 fixture：只读 `../stateverify/testdata/{anchor,rotation,metadata,desired}.{payload,envelope}.json` 与 vectors.json；1 次链使用冻结原始向量，独立 scope 固定 node `12345678-1234-4234-8234-123456789abc`/epoch 19。对 n=0/32 及变体，参考既有 fixtures_test.go 的测试专用固定种子签名方法；原有 vectors/生成器不变。根 raw 从冻结 payload/测试侧 json.RawMessage 取得，不调用产品 parser 生成全部期望。metadata version 6、desired generation 7/sequence 9 等固定字段直接核对，P/E 用冻结向量或测试侧独立 SHA-256；不得由组合层输出反推期望。
- [x] 写 TestCompositionSelection、TestCompositionAnchor、TestCompositionTargetRefs：先对 fixture 的材料子集实际 Evaluate 并清理，再 ownInput/selectBatch。0/1/32 链成功；断开 100→101 历史可额外保留但不可选作锚/目标链。`bytes.Equal(selected.batch.DesiredEnvelope, owned target raw)` 且元数据、每个轮换、链尾逐字来自正确表项；0 链尾=锚，其他=最后一个 rotation。33 链先 LimitExceeded。

  正向 TestCompositionSelection 对每个 n ∈ {0,1,32} 至少固定以下坐标，借用字节另按完整 ref 查到的表项比较：

  ```go
  if selected.rootVersion != uint64(n+1) ||
      selected.batch.ExpectedEpoch != 19 ||
      selected.batch.ExpectedNodeID != "12345678-1234-4234-8234-123456789abc" ||
      len(selected.batch.Rotations) != n { t.Fatal("selected coordinates") }
  ```

- [x] 各自单独变更锚 P/E/scope、锚 payload/根 Raw、RootSet E、每种 Target Kind/P/E、新根/旧根关联、links 顺序/重复/缺段、目标 metadata/desired 关联；断言对应 InvalidInput/MissingMaterial/BindingMismatch，损坏或高版本 InitialRoot 为 InvalidMaterial。合法同 P 不同 E 的 envelope 用结构合法的额外未知 share 构造，确认材料表唯一性不被绕过、选择不能只看 P。
- [x] 运行 `& $compositionGo test ./internal/nodecontrol/statecomposition -run '^(TestCompositionSelection|TestCompositionAnchor|TestCompositionTargetRefs)$' -count=1`，记录 RED。
- [x] 实现 selectBatch：独立锚固定 Kind 解析、核对 P/E/scope/version/key/threshold，清理 RootMaterial/payload 临时副本，比较 AnchorRoot exact raw；有序核对每个选择和原预期关联。Batch 的 raw 全从同快照借用，ActiveRootEnvelope 仅从锚/链尾设定；没有外来 Batch 参数。错误清理外层 refs，保留借用 raw 给输入所有者统一清理。
- [x] 重跑 targeted 和完整 statecomposition；exit 0，记录测试及 source 差异，不提交。

## Task 3: 组合 Evaluate、真实验签与逐项证据回核

**Files:** 新建 evaluate.go、evidence.go、evaluate_test.go；扩展 errors.go/fixtures_test.go；model.go 此步只提供私有字段所有者及 Clear，摘要/重检留任务 4。

**Interfaces:** 消费 ownInput/selectBatch 及 `statematerials.Evaluate(ctx,Input)`、`stateverify.Verify(ctx,Batch)`。产生公共 Evaluate、CompositionModel/Clear，以及私有 `checkEvidence(ctx context.Context, in Input, selected *selection, evidence *stateverify.SignatureEvidence) error` 和固定 mapError。

- [x] 写 TestCompositionEvaluate 与 TestCompositionErrorMapping：真实 0/1/32 链返回非 nil；独立映射表每行用 errors.Is 断言（含 wrapped 下层错误、未知动态错误被固定化、超时保留）。格式正确假签名、current/new 任一不足、已知坏份额、缺失/撤销 key、静态窗不符拒绝且模型 nil；不要让更早的格式损坏遮蔽想测试的密码学失败。
- [x] 写 TestCompositionEvidenceBinding：使用实际 Verify 得到的只读证据，对私有 checkEvidence 的独立比较目标逐项变更 anchor P/E、每个 rotation P/E/数量/顺序、metadata/desired P/E、node/epoch/root version/metadata version/desired generation/sequence/key；每次 BindingMismatch，nil/cleared evidence 不成功。每项测试重建合法基线，不用双重变异相互抵消，不加入产品 mock verifier。
- [x] 写 TestCompositionDeclaredAssociation、TestCompositionHistoricalTarget：两份签名正确且含适用 key 的兼容 metadata，固定旧声明时选另一份 BindingMismatch；同时改写完整声明的新 Evaluate 可条件成功，声明保持不是 wire 签入了 metadata P。Jan 2020 过期向量仍条件成功；旧可用 metadata 与额外新撤销 metadata 并存；选中 revoked snapshot 失败。未选中材料的坏签名可保持材料一致，但未知 schema/损坏 raw/预期不符/缺失关系仍失败。

  TestCompositionEvaluate 对正常 fixture 及表驱动负向输入都使用同一 helper，核心断言固定如下：

  ```go
  if want == nil {
      if err != nil || model == nil { t.Fatalf("missing success: %v", err) }
      t.Cleanup(model.Clear)
  } else if !errors.Is(err, want) || model != nil {
      t.Fatalf("want %v, got model=%v err=%v", want, model != nil, err)
  }
  ```

- [x] 运行 `& $compositionGo test ./internal/nodecontrol/statecomposition -run '^(TestCompositionEvaluate|TestCompositionErrorMapping|TestCompositionEvidenceBinding|TestCompositionDeclaredAssociation|TestCompositionHistoricalTarget)$' -count=1`，记录 RED。
- [x] 实现 Evaluate：完整 owned 快照→完整材料 Evaluate→选择→唯一 Verify→逐项 checkEvidence→最终 context 检查→转移模型所有权。临时子集模型、selection、格式对象、签名证据注册清理；保存 Subset.Summary、证据坐标和 finalRoot 固定值，不返回下层所有者。失败/取消无模型，panic 清理后传播。
- [x] 实现 checkEvidence 逐项完整 P/E 和坐标核对，不重算 hash，不只比较数量或版本；实现 errors.go 固定映射和 model.Clear。重跑 targeted 与完整 statecomposition、statematerials、stateverify；全部 exit 0，记录结果，不提交。

## Task 4: 条件性摘要、全部材料标记及完整 Recheck

**Files:** 完善 model.go，新建 model_test.go；扩展 evaluate_test.go 的取消/panic 测试。

**Interfaces:** 消费 Task 3 的私有 CompositionModel。产生固定 Summary/状态/义务/Coverage 类型及公共 Summary、Recheck；Clear 继续满足 Task 3 语义。

- [x] 写 TestCompositionSummary：Status=ConditionalTargetVerified，Anchor=P/E 独立 InitialRoot，FinalRoot 为 RootSet ref；目标 anchor root=AnchorAssumed，选中新根=ChainDerived，选中 rotations/metadata/desired=SignatureChecked，所有其他项=MaterialOnly。`len(summary.Coverage)==len(in.Materials.Materials)` 且每 ref 恰一次、Kind/P 顺序规范；Target.Links 顺序保留，Subset 固定 6 义务与本层固定 7 义务原样，即使无额外历史也保留。
- [x] 写 TestCompositionModelOwnership：逐项改变原 raw、预期每种嵌套集合、Links 和返回 Summary 的 Links/Coverage，二次读取及模型私有 input 不变；保留私有 raw/集合别名后 Clear 确认清零，原输入和 detached summary 不变。nil/zero/cleared Summary 深等于 `Summary{}`，Recheck 为 ErrClearedModel，Clear 幂等。
- [x] 写 TestCompositionRecheck：原输入、顶层表重排与 nil/empty 等价通过；链乱序拒绝。合法同 P 不同 E、独立锚/预期 scope、完整原声明合法换 metadata、未选中材料合法替换、任何完整预期字段改变都不得通过旧模型重检。新输入本身合法时先证明其 Evaluate 成功，再要求旧模型 Recheck BindingMismatch，区分真实模型身份拒绝和更早校验错误。

  TestCompositionSummary 固定义务断言及 TestCompositionRecheck 的合法替换断言如下（old 为原输入的成功模型，replacement 为自有合法替换输入）：

  ```go
  expected := [7]Obligation{ExternalAnchorProvenance, UnselectedSignaturesAndHistory,
      DomainAndCurrentTime, FullStateMaterialClosure, FullStateComposition,
      PhysicalCommit, RuntimeAuthorization}
  if got.Status != ConditionalTargetVerified || got.Obligations != expected ||
      len(got.Coverage) != len(in.Materials.Materials) { t.Fatal("coverage or obligations") }
  fresh, err := Evaluate(context.Background(), replacement)
  if err != nil || fresh == nil { t.Fatalf("replacement not valid: %v", err) }
  defer fresh.Clear()
  if err := Recheck(context.Background(), old, replacement);
      !errors.Is(err, ErrBindingMismatch) { t.Fatal(err) }
  ```

- [x] 写 TestCompositionCancellation/TestCompositionPanicCleanup：测试专用计数 Context.Err（仅 context 输入，无产品 hook）覆盖预取消、own/选择阶段、验签结束回核、最终返回前取消/超时；失败模型 nil、errors.Is 匹配，原输入不变。用 panic Context.Err 触发拥有输入后的 panic，确认传播与早登记 defer；私有 clearInput/Clear 的保留别名检查证明清零，配合任务 5 AST/源码清理检查，不声称从外部观察了所有不可达运行时副本。
- [x] 运行 `& $compositionGo test ./internal/nodecontrol/statecomposition -run '^(TestCompositionSummary|TestCompositionModelOwnership|TestCompositionRecheck|TestCompositionCancellation|TestCompositionPanicCleanup)$' -count=1`，记录 RED。
- [x] 实现 Summary：只从私有成功模型生成固定状态、原子集摘要、独立锚 P/E、目标/坐标/链尾及逐项 coverage；返回独立 Links/Coverage，不输出 raw/公钥/下层 owner。Recheck 调用新的完整 Evaluate（不是缓存 evidence），defer 清理 fresh；reflect.DeepEqual 比较完整规范 input，包含额外材料和所有声明，比较后再检查取消，不相同 BindingMismatch。
- [x] 重跑 targeted 和完整 statecomposition；exit 0。对可观察的关键 metadata 选择、MaterialOnly 标记及 Recheck 完整比较各做一次临时省略变异，记录相应测试 RED，逐个恢复并核对产品哈希，再运行原样 GREEN。不要求对下层重复蕴含的守卫制造假 RED，不提交。

## Task 5: 限定回归、一次最终独立审查与交付

**Files:** 完善 boundary_test.go 清理/消费者检查；新增 `docs/roadmap/2026-10-10-main-state-selected-signature-composition-execution.md`，更新计划/设计/current-status；维护 scratch 记录。

**Interfaces:** 消费任务 1–4 全部 API/结果及只读基线。只交付此增量的测试与审查证据，不新增产品功能或授权状态。

- [x] 增加 TestCompositionCleanupRegistration：AST 验证 Evaluate 在调用下层前登记自有输入清理，禁止 late-defer 反例；验证消费者扫描覆盖所有非测试 Go 文件，两个旧边界守卫各保留原禁止样例。运行针对 guard 的 RED（缺少该断言/检查时），最小补测试守卫，再 targeted GREEN。
- [x] 格式化只限新包与两处批准的测试文件；用固定 gofmt 的 list 输出检查为空。基线中只允许上述两处旧测试改变；其余 Go/SQL/模块/锁文件哈希不变，HEAD/索引保持未写入。普通 Git 读若权限阻塞，只请求明确的只读诊断权限，不修改 safe.directory 或 Git 索引。
- [x] 完整回归分包独立运行并记录所有退出码：`& $compositionGo test ./internal/nodecontrol/statecomposition -count=1`，再相同形式逐个运行 statematerials、stateformat、stateverify、statepolicy、statecontract、staterecovery、statetime、serving、contracts、readiness；全部必须 exit 0。
- [x] 单独运行 `& $compositionGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1`；单独运行 `& $compositionGo test ./internal/store -run '^$' -count=1`（仅编译，不是 store 测试通过）。等待所有会话结束，保留原环境与范围。
- [x] Native 作者执行所有实现；完成限定验证后仅派出一次全新、只读、最强可用模型 reviewer，对该增量所有任务和跨包边界做整体源码审查。审查包包含设计/计划、基线例外、准确测试证据、临时变异恢复哈希及所有未完成责任；无实现代理、无逐任务审查、无第二次独立复审，不能称为全分支/项目安全验收。
- [x] 作者裁定发现：Critical/Important 必须先复现，再最小修复、限定重测；设计外扩展先报告。Minor 仅在原批准范围内处理（补测试/文档或限定缺陷修正），每项记录接受/修复/有理由保留；不隐藏阻断，也不临时放开边界。审查后有产品修复须重新跑全部上述批准回归；只有新测试变化至少重跑受影响包及两处边界。
- [x] 用 apply_patch 完成执行记录、任务 checkbox 和 current-status，保留历史证据。汇报单目标条件性组合进度、审查结论、所有最终退出码和仍未验证责任；Git 未提交/推送，生产兼容性 UNKNOWN，完整 Task10/B01/C12/main 验收仍未完成，不触发合并/清理菜单。

## 作者自查与执行交接

设计 §1–2/§8 → 任务 1/5；§3–4 → 任务 1/2；§5/§5.1 → 任务 2/3；§6 → 任务 3/4；§7 → 任务 3/4/5；§9 → 五项测试循环及最终限定回归；§10 → 本计划审阅交接和最终记录。
五项 Review Focus 均有具体测试；输入/私有 helper/结果字段在上述 Interfaces 中已冻结，未定义第二个摘要或 verifier。不得把预算 equal 的单元测试、额外坏签名的 MaterialOnly 或作者取消测试说成生产验收。

批准前的计划自查仅覆盖类型一致性、步骤明确性、范围和测试覆盖，不是独立审查或产品测试。批准后已按 executing-plans 的 Native 方法完成五项任务、限定验证和唯一一次独立增量审查；准确结果与全部范围裁定见 `docs/roadmap/2026-10-10-main-state-selected-signature-composition-execution.md`。无遗留 Minor；审查后仅更新文档，未修改源码或测试。
