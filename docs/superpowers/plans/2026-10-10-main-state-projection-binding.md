# Main-state Dual-snapshot Projection Binding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将完整旧/新 L0 声明分别绑定到同一次自有快照中的两张材料表，并将新 desired 精确绑定到现有单目标条件性验签结果，不产生生产授权。

**Architecture:** 新增隔离纯包 statebinding，调用 statecontract.EvaluateModel、statematerials.Evaluate（旧表）及 statecomposition.Evaluate（新表）。不合并两张表，不把当前 root transition 与目标原授权链混用；完整 Recheck 保存所有声明、材料和未验证字段，下层责任不删除。

**Tech Stack:** 固定 Go 1.26.5、现有 contracts/statecontract/statematerials/statecomposition、PowerShell、apply_patch、离线缓存；无新依赖。

**Spec:** `docs/superpowers/specs/2026-10-10-main-state-projection-binding-design.md`，用户于 2026-10-10 批准；执行前完整阅读设计与本计划。

状态：用户于 2026-10-10 已批准书面设计及本计划，Native 五项任务已完成本地交付。唯一一次全新只读增量审查为 0 Critical / 0 Important / 2 Minor；两项均由作者修正，完整十四条验证重跑全部 exit 0，无第二次独立复审。不写 Git，整体 Task10/B01/C12/main 仍未验收；见执行记录。

## Global Constraints

- 工作目录仅 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`，复用现有分支并保留既有未提交内容；不新建/清理 worktree。
- 不暂存、提交、推送、合并、重置或删除分支，不修改 safe.directory/永久环境；记录步骤替代技能模板的 commit/清理步骤。
- 旧、新表各自 Materials/Expected ≤128；各自材料/预期 RootRotation ≤32；Main.RootTransition/Proposed.Target.Links 各 ≤32。
- 每份 RootSet raw ≤65,536 bytes；其他材料及独立锚各 ≤262,144 bytes；每表 raw ≤16 MiB；外来 raw 总量按出现次数计 ≤32 MiB + 262,144 bytes，不按别名去重，不承诺峰值内存。
- 深复制前检查两表每份 expectation 全部分支集合：root keys 5、metadata keys 64、revoked IDs 4096、desired slots 8，包括非活动分支。
- Main 旧、新各自九个集合先查长度：root keys 5、metadata keys 64、revoked 4096、两个 purpose deauthorized 各 128、desired slots 8、LKG entries 8、local faults 64、supervisor faults 16；连 RootTransition 共十九个 L0 集合。
- RootSet E=0；其他支持的签名材料及独立锚 P/E 非零。封闭 Kind 映射，不直接强转 enum，不用锚/轮换 E 代替 RootSet E。
- 新 Desired 必须 Present 且等于明确目标；旧 Desired 可合法 NeverEstablished。仍只支持 v1/单 root key/threshold 1 独立锚和最多 32 次连续轮换，不补链、截链、排序链、替换锚或自动换最新 authorizer。
- 全部输入先拥有；仅两表顶层按 Kind/P 排序、空集合规范化 nil；Main 有序集合、RootTransition、Target.Links 不排序、不补齐、不截断。
- 下层产品 API/wire/算法/规则、stateverify 守卫、statepolicy、SQL、模块、锁文件和冻结向量不变；仅三个指定旧 boundary_test.go 可改变。
- 产品无自行解析/hash/验签/JSON、当前时间、I/O、网络、环境、随机、硬件、provider、重试或 goroutine；禁止生产消费者。
- 成功仅 ConditionalProjectionBound。旧基线可信性、历史原始授权、历史集合摘要/平台来源、完整闭包/组合、物理提交、运行授权仍未证明；下层及本层义务原样保留。
- 失败/取消返回 nil 模型；固定错误不回显输入；panic 清理后传播。调用期间不允许并发修改输入，模型不允许浅复制或并发 Clear/访问；不承诺字符串/不可达/调用者副本擦除。
- 只运行本计划列出的 12 个完整包、指定 authority 和 store 编译；不跑全仓库、真实 DB/Docker/race/硬件/重启/C11/C12/PITR，不延长 deadline、不下载依赖。
- 手工文件编辑仅 apply_patch；gofmt 只限新包及三处指定旧测试。所有已启动会话必须等结束，不能因输出延迟重复启动或假称超时。

## Review Focus

1. 旧/新表共享调用者底层存储且出现同 P 不同合法 E：必须分别拥有并精确绑定各自 E，不能合并、借用或误拒绝（任务 1、2、4）。
2. 当前 root 更新链与目标原授权链不同、当前 metadata 更新：目标保留原 authorizer，当前/transition 未选材料不得提升验签等级（任务 2–4）。
3. 旧 L0 的 root keys 等声明与旧表不一致，但 L0 和两表各自仍合法：旧侧也必须 BindingMismatch，不能只验证新侧（任务 2、3）。
4. 未投影历史摘要/平台声明或两表额外材料合法变化：新 Evaluate 可成立，旧模型完整 Recheck 必须拒绝，不把“输入没变”当来源证明（任务 4）。
5. 最终摘要已经拥有后才取消/Context.Err panic，以及伪装消费者：不能交付模型、清理要早登记、三个消费例外只允许精确路径/包名（任务 1、4、5）。

---

## 文件分工与冻结接口

新目录 `internal/nodecontrol/statebinding/`：doc.go（证明边界）、input.go（输入）、errors.go（固定映射）、budget.go（全部预算）、owned.go（复制/规范化/清理）、projection.go（双侧投影/精确查找）、target.go（本次 transition/原授权目标）、evaluate.go（组合入口）、model.go（私有所有者/摘要/重检）。

新测试文件：owned_test.go、fixtures_test.go、projection_test.go、evaluate_test.go、model_test.go、boundary_test.go。旧文件仅 `internal/nodecontrol/{statecontract,statematerials,statecomposition}/boundary_test.go`。文档交付为本计划、已批准设计、current-status 和后续 `docs/roadmap/2026-10-10-main-state-projection-binding-execution.md`；保留 scratch。

```go
type Input struct {
    Main statecontract.ModelInput
    Previous statematerials.Input
    Proposed statecomposition.Input
}
type ResultStatus uint8 // 0 InvalidResult; 1 ConditionalProjectionBound
type Obligation uint8 // 0 UnknownObligation; nonzero: spec §7 的八项，顺序不变
type Summary struct {
    Status ResultStatus
    Contract statecontract.ContractSummary
    Previous statematerials.Summary
    Proposed statecomposition.Summary
    Obligations [8]Obligation
}
func Evaluate(ctx context.Context, in Input) (*BindingModel, error)
func Recheck(ctx context.Context, model *BindingModel, in Input) error
func (model *BindingModel) Summary() Summary
func (model *BindingModel) Clear()
```

BindingModel 私有字段固定为 `valid bool`、`input Input`、`contractSummary statecontract.ContractSummary`、`previousSummary statematerials.Summary`、`proposedSummary statecomposition.Summary`；没有下层 owner、provider、外来 receipt 或第二种成功模型。下层公开结构中的嵌套可变集合只在 proposedSummary.Target.Links/Coverage，必须单独拥有/清理并在返回时再复制。

errors.go 定义 spec §7 的 11 个固定 sentinel，不追加输入，文字逐项冻结：

| Sentinel | 固定文字 |
| --- | --- |
| ErrInvalidInput | `statebinding: invalid input` |
| ErrLimitExceeded | `statebinding: limit exceeded` |
| ErrIncompleteCoverage | `statebinding: incomplete coverage` |
| ErrInvalidMaterial | `statebinding: invalid material` |
| ErrMissingMaterial | `statebinding: missing material` |
| ErrDuplicateMaterial | `statebinding: duplicate material` |
| ErrBindingMismatch | `statebinding: binding mismatch` |
| ErrSignatureMismatch | `statebinding: signature mismatch` |
| ErrRollback | `statebinding: rollback` |
| ErrConflict | `statebinding: conflict` |
| ErrClearedModel | `statebinding: cleared model` |

`mapError(ctx context.Context, err error) error` 按设计逐类 errors.Is 映射，当前 ctx.Err 优先、nil 未取消返回 nil、未知非 nil→InvalidInput。投影查找缺失是 MissingMaterial；先发生的下层非法空表等仍按其实际错误映射，不改下层优先级。

## 固定环境、基线和记录

每个新 PowerShell 会话设置以下进程级环境；不修改永久设置。每条测试独立运行，记录完整输出与实际退出码，不用后续命令的成功覆盖前面的失败：

```powershell
$env:GOENV='off'; $env:GOWORK='off'; $env:GOTOOLCHAIN='local'
$env:GOPROXY='off'; $env:GOSUMDB='off'; $env:GOAUTH='off'
$env:GOVCS='all:off'; $env:GOFLAGS='-mod=readonly'
$env:GOPATH='C:/Users/Lenovo/go'; $env:GOMODCACHE='C:/Users/Lenovo/go/pkg/mod'
$bindingGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
$bindingFmt='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/gofmt.exe'
```

执行前在 `.superpowers/sdd/2026-10-10-main-state-projection-binding/` 用 apply_patch 保存 progress.md 和 baseline.json（实际 HEAD/分支/索引 SHA256、Go/SQL/module/lock 各文件 SHA256）。规划点为 432 份既有文件；实施时重新记录实际数量，不重置用户变化。仅三个旧测试可变，其余旧文件须哈希不变；保留完整既有工作树读数。普通 Git 读受阻时只请求明确的只读诊断权限，不改安全配置/索引。记录 RED 必须是预期缺接口/断言失败，环境或路径错误不算 RED。

## Task 1: 有界自有双快照与精确隔离边界

**Files:** 创建 doc.go/input.go/errors.go/budget.go/owned.go、owned_test.go/boundary_test.go；仅修改三个指定旧 boundary_test.go。基线/ledger 随本任务创建并保留。

**Interfaces:** 消费三种下层 Input 及 contracts.Digest；产生本层 Input/11 个错误和以下私有函数（守卫为 test-only）：

```go
func checkBudget(in Input) error
func ownInput(ctx context.Context, in Input) (Input, error)
func clearInput(in *Input)
func bindingImportsAllowed(source []byte) bool
func bindingConsumerAllowed(path, pkg, imported string) bool
```

statecontract 测试中的 `contractConsumerAllowed` 改为相同三参数签名，扫描点传入 repo-relative path/实际 package/import path；其余两个旧 consumer helper 保留现有签名，仅追加精确例外。

- [x] 保存实施前基线及 progress；确认旧进度不被覆盖，新包路径没有未知重叠。台账裁定：PowerShell/记录替代 Bash/commit/cleanup，限定离线范围替代全仓库验证；代价是本机未提交进度和整体兼容性未知。
- [x] 写 TestBindingBudget：每个表/分支/十九个 L0 集合/两条链/单份及总 raw 的 equal 与 +1 独立断言；使用不要求格式合法的预算输入，预取消优先。核心断言为 `if err := checkBudget(in); err != nil { t.Fatal(err) }`；只变一项为上限+1 后 `if err := checkBudget(in); err != ErrLimitExceeded { t.Fatal(err) }`，独立锚 262144/262145、两表总 raw 32MiB+262144 明确包含。
- [x] 写 TestBindingOwnedInput/TestBindingBorrowedAliases：分别修改原两表 raw/每份预期四个数组、十九个 L0 集合、锚及 links，断言 ownInput 不变；故意让旧/新指向同一调用者存储，两个 owned 副本仍独立。clearInput 清零保留的自有 raw/数组别名而不改调用者；顶层表排序/空→nil，L0 和链顺序保持。空锚 InvalidInput；此任务不做材料格式检查。
- [x] 写 TestBindingBoundary、TestContractBindingBoundary、TestSubsetBindingBoundary、TestCompositionBindingBoundary：精确 `internal/nodecontrol/statebinding/x.go`/`statebinding` true，reader/serving、子目录、相似前缀、错误包名 false；保留所有原禁止样例和 statematerials→statecomposition 例外。新包产品能力守卫拒绝 dot/blank/重复路径/冲突别名/GoStmt/禁用成员。
- [x] 运行 `& $bindingGo test ./internal/nodecontrol/statebinding ./internal/nodecontrol/statecontract ./internal/nodecontrol/statematerials ./internal/nodecontrol/statecomposition -run '^(TestBindingBudget|TestBindingOwnedInput|TestBindingBorrowedAliases|TestBindingBoundary|TestContractBindingBoundary|TestSubsetBindingBoundary|TestCompositionBindingBoundary)$' -count=1`，记录缺少 API/新签名/精确例外的 RED。
- [x] 实现 budget.go/owned.go：先 context/所有长度与安全 raw 累加，再深复制整个 Input；拥有数据后立即登记失败/panic 清理，逐段检查取消，仅顶层表规范化。实现三处消费守卫例外，适配 statecontract 原调用/扫描点，不更改下层产品导入能力或算法。
- [x] 固定新包成员白名单：context.Context/Canceled/DeadlineExceeded，errors.New/Is，bytes.Clone/Compare/Equal，sort.Slice，reflect.DeepEqual，contracts.Digest；statecontract 仅 ModelInput/MainStateContentModel/MainStateScope/RotationRef/ArtifactRef/ArtifactKind/Coverage/Present/NeverEstablished/RootPayload/RootRotation/Metadata/DesiredSnapshot/ContractModel/ContractSummary/ContractShapeValid/EvaluateModel 及六个固定错误；statematerials 仅 Input/Scope/Material/Expectation/MaterialRef/MaterialKind/四个 Kind/MaterialSubsetModel/Summary/SubsetConsistent/Evaluate 及七个固定错误；statecomposition 仅 Input/AnchorInput/Target/RotationLink/Coverage/CompositionModel/Summary/ConditionalTargetVerified/Evaluate 及八个固定错误。不给整包能力、stateformat/stateverify 或反射其他成员放行。
- [x] 重跑 targeted 为 GREEN，再分别完整测试 statebinding、statecontract、statematerials、statecomposition，每次 exit 0；记录只允许三处既有测试差异，不提交。

## Task 2: 双侧字段投影、独立链语义与明确目标

**Files:** 创建 projection.go/target.go、fixtures_test.go/projection_test.go。

**Interfaces:** 消费 Task 1 的自有 Input 和已由原层检查的完整两表；产生：

```go
func projectRef(ref statecontract.ArtifactRef) (statematerials.MaterialRef, error)
func findExpected(ctx context.Context, in statematerials.Input, ref statematerials.MaterialRef) (statematerials.Expectation, error)
func bindProjection(ctx context.Context, in Input) error
func bindTransitions(ctx context.Context, in Input) error
func bindTarget(ctx context.Context, in Input, proposed statecomposition.Summary) error
```

测试 helper 固定 `bindingFixture(t *testing.T, rotations int) Input`、`copyBindingInput(t *testing.T, in Input) Input`、`checkBindingLayers(t *testing.T, in Input) statecomposition.Summary`。它们只在 fixtures_test.go：独立 JSON/SHA/测试签名和显式 L0 构造，不调用本层投影/Summary 得到期望值；checkBindingLayers 调用真实三层并登记清理，返回 detached 新摘要。

- [x] 构造 0/1/32 链；n=1 只读复用 `../stateverify/testdata` 冻结向量，其余仅在新测试文件用固定测试 seed 构造，不增加生成器/改向量。补齐十二域 L0：未覆盖域仅合法引用声明，不假造已验证材料；root/metadata/desired 的期望事实从独立测试侧来源生成。正常 baseline 固定 node=`12345678-1234-4234-8234-123456789abc`、epoch=19、目标 root version=n+1；先 checkBindingLayers 证明三层各自合法。
- [x] 写 TestBindingProjection/TestBindingDualEnvelope/TestBindingPreviousProjectionMismatch：spec §6 表格每个字段两侧逐项错配（含同数量不同 key/revoked/slot）；精确 Kind/P/E 查找，RootSet E=0、Expectation epoch/sequence。合法新旧同 P 不同 E 分表成功；仅 Main 引用换 E 则 BindingMismatch；不存在投影对象 MissingMaterial。旧 root keys 换成另一合法有序集合，先证明 L0/两表仍合法，再要求投影 BindingMismatch。
- [x] 写 TestBindingTransitions/TestBindingTarget/TestBindingSeparatedChains：当前旧→新链独立于锚→历史 authorizer 链；逐项错配 rotation E、sequence、前/后版本、old/new payload 与端点，原顺序不变。目标必须是 Proposed 主 desired，metadata/root 是该 desired 原声明；修改真实下层 detached 摘要的状态/scope/目标/P/E/链尾/各坐标独立拒绝。新 Desired NeverEstablished InvalidInput，旧合法 NeverEstablished 无伪造 desired；不自动换最新 metadata 或延长目标链。

  独立链基线必须先有三层合法证明，再有以下具体断言：

  ```go
  proof := checkBindingLayers(t, in)
  if in.Main.Proposed.Desired.MetadataPayload == in.Main.Proposed.Metadata.Mark.Material.Payload {
      t.Fatal("fixture did not retain historical authorizer")
  }
  if err := bindTransitions(context.Background(), in); err != nil { t.Fatal(err) }
  if err := bindTarget(context.Background(), in, proof); err != nil { t.Fatal(err) }
  ```

- [x] 运行 `& $bindingGo test ./internal/nodecontrol/statebinding -run '^(TestBindingProjection|TestBindingDualEnvelope|TestBindingPreviousProjectionMismatch|TestBindingTransitions|TestBindingTarget|TestBindingSeparatedChains)$' -count=1`，记录缺少私有投影函数的 RED。
- [x] 实现 projection.go：封闭四类 Kind 映射，findExpected 同时精确命中材料和预期的完整 ref；Kind/P 不存在为 MissingMaterial，Kind/P 已存在而 E 不同为 BindingMismatch。依设计逐字段比较两侧 scope/root/metadata/已建立 desired，包括完整有序集合，不借用另一表或只看数量。预算/格式与有序性由前序检查保证，不自行解析/hash。
- [x] 实现 target.go：RootTransition 每个 ref/坐标及端点绑定新表与旧/新主 root；bindTarget 回核真实摘要状态/scope、原 Target/独立锚身份、FinalRoot 及 Coordinates.ExpectedNodeID.String()/ExpectedEpoch/FinalRootVersion/MetadataVersion/DesiredGeneration/DesiredSequence/DesiredKeyID。版本取选中 authorizer 预期，不误取当前 metadata；没有 proof grade 升级。
- [x] 重跑 targeted 和完整 statebinding，各 exit 0，记录 fixture/映射/链范围，不提交。

## Task 3: 同快照 Evaluate、早清理所有者与固定错误

**Files:** 创建 evaluate.go、model.go/evaluate_test.go；补齐 errors.go。

**Interfaces:** 消费 Task 1 ownInput/clearInput、Task 2 三个 bind 函数及真实下层入口；产生公共 Evaluate、冻结私有 BindingModel、Clear，以及 `mapError(ctx context.Context, err error) error`、`clearCompositionSummary(s *statecomposition.Summary)`。Evaluate 实现使用命名返回值 `func Evaluate(ctx context.Context, in Input) (result *BindingModel, err error)`，使失败/panic 的延迟清理可检查 `result == nil`；公开调用签名不变。Task 4 才增加本层 Summary/Recheck。

- [x] 写 TestBindingEvaluate/TestBindingErrorMapping：0/1/32、旧 NeverEstablished、同 P 不同 E、历史 authorizer/独立 current transition 正向成功；L0 缺域/rollback/conflict、旧表损坏/缺失关系、新目标假签名、两侧投影/transition ref 错配都返回 nil 模型及固定错误。单独测 wrapped ContractError/各类下层 sentinel、未知动态文本、nil/取消/超时，逐行 errors.Is。

  两个集成子测名称固定为 `TestBindingEvaluate/previous_root_keys` 和 `TestBindingEvaluate/transition_envelope`。前者仅把 Main.Previous.Root.KeyIDs 换成另一合法有序集合，旧材料表保持不变；后者仅把 Main.RootTransition 中一项 Material.Envelope 换成另一非零 E，新材料表和选中目标保持不变。每个子测先调用 `checkBindingLayers(t, in)`，证明变更后的 L0/两张材料表及现有组合各自仍合法，再要求 Evaluate 返回 ErrBindingMismatch 和 nil 模型。

  每个成功/失败子测使用同一断言形状，不让更早错误遮蔽目标守卫：

  ```go
  model, err := Evaluate(context.Background(), in)
  if want == nil {
      if err != nil || model == nil { t.Fatalf("missing result: %v", err) }
      t.Cleanup(model.Clear)
  } else if !errors.Is(err, want) || model != nil {
      t.Fatalf("want %v, model=%v err=%v", want, model != nil, err)
  }
  ```

- [x] 写 TestBindingProofStrength：旧表格式正确坏签名可保持材料一致；新表未选 current metadata/transition 仅 MaterialOnly，只有目标保持真实下层等级。Metadata.History/RevokedSetDigest、scope.TrustAnchorID 与外部锚无未定义强制等式；正常 opaque 不同 ID 可成立，但不获得来源证明。坏 schema/raw/预期/关系仍失败。
- [x] 运行 `& $bindingGo test ./internal/nodecontrol/statebinding -run '^(TestBindingEvaluate|TestBindingErrorMapping|TestBindingProofStrength)$' -count=1`，记录缺少 Evaluate/模型/映射的 RED。
- [x] 实现 Evaluate：ownInput→私有模型接管快照并立即登记 result==nil 时 model.Clear→完整 L0→完整旧材料→新 composition→bindProjection/bindTransitions→捕获三个摘要到已受清理保护的私有字段→bindTarget/摘要状态与 scope 回核→最终 context→valid/返回。每个下层返回后立即 defer Clear，先于 mapError 的 ctx.Err；失败与 panic 不转移模型。新摘要数组即刻由模型拥有，错误时 clearCompositionSummary 清零。
- [x] 实现固定 mapError 和 nil-safe/幂等 Clear；清理模型 input、proposedSummary.Target.Links/Coverage 和所有私有值，不清除调用者/detached 摘要。
- [x] 重跑 targeted，再分别完整 statebinding、statecontract、statematerials、statecomposition，全部 exit 0。
- [x] 临时省略旧侧投影，用 `& $bindingGo test ./internal/nodecontrol/statebinding -run '^TestBindingEvaluate/previous_root_keys$' -count=1` 要求该集成子测断言 RED；立即 apply_patch 恢复，核对修改前源码 SHA256。
- [x] 临时省略 Evaluate 的 bindTransitions 调用，用 `& $bindingGo test ./internal/nodecontrol/statebinding -run '^TestBindingEvaluate/transition_envelope$' -count=1` 要求该集成子测断言 RED；立即 apply_patch 恢复，核对修改前源码 SHA256。两个 probe 都不保留、不提交，不为下层已蕴含的比较制造假 RED。
- [x] 在两次恢复后原样完整运行 statebinding，要求 GREEN 并保存实际退出码。

## Task 4: 限定摘要、全部输入 Recheck 与取消/panic

**Files:** 完善 model.go，新建 model_test.go；扩展 evaluate_test.go。

**Interfaces:** 消费 Task 3 私有字段/Clear，产生冻结 ResultStatus/Obligation/Summary、Summary/Recheck，以及 `cloneCompositionSummary(s statecomposition.Summary) statecomposition.Summary`（复制 Links/Coverage）。

- [x] 写 TestBindingSummary：唯一 ConditionalProjectionBound；三个下层摘要和全部 5/6/7/嵌套6 义务不变，本层八项顺序为 ExternalAnchorAndPlatformProvenance、PreviousSignaturesAndUnselectedHistory、UnverifiedHistoryAndSetDigests、DomainAndCurrentTime、FullStateMaterialClosure、FullStateComposition、PhysicalCommit、RuntimeAuthorization。新 coverage 原样、旧表不升级；nil/zero/cleared 结果完整零值，Recheck 为 ClearedModel。
- [x] 写 TestBindingModelOwnership：改变调用者十九个 L0 集合、两表所有 raw/expectation 数组、锚/links 和每次 Summary 的 Links/Coverage；私有 input/二次摘要不变。保留私有自有别名，Clear 后清零，调用者与已返回摘要不变；共享调用者底层存储也不得串改。
- [x] 写 TestBindingRecheck/TestBindingUnmappedRecheck：相同规范输入、顶层表重排与 nil/empty 通过；链/有序域乱序拒绝。两侧合法同 P 不同 E、旧额外材料/新未选材料替换、主 metadata 历史引用/集合摘要、合法未投影身份字段及 coherent 两侧 scope.TrustAnchorID 变化，都先证明新 Evaluate 成功，再断言旧 Recheck BindingMismatch。完整 L0 两侧每个字段独立变更必须拒绝（允许更早的合法性错误），不只测目标字段。

  子测 `TestBindingUnmappedRecheck/coherent_authorizer_rewrite` 独立构造另一份合法兼容 metadata（授权同一 desired key），保持 selected desired raw/P/E 不变，一致更新两份 L0 中已建立 desired 的 MetadataPayload、相应两表 desired expectations 的 MetadataPayload 及新 Target.Metadata；其余可见关联声明按 L0 规则一致更新，不把新关联冒充 wire 已签入。先 `checkBindingLayers(t, replacement)`，再用以下完整 Evaluate/旧 Recheck 断言证明：新条件性模型可成立，旧模型身份必须拒绝。测试只在新 fixtures_test.go 构造材料，不修改下层 helper/API。

  合法替换反例核心断言：

  ```go
  fresh, err := Evaluate(context.Background(), replacement)
  if err != nil || fresh == nil { t.Fatalf("replacement not valid: %v", err) }
  defer fresh.Clear()
  if err := Recheck(context.Background(), old, replacement);
      !errors.Is(err, ErrBindingMismatch) { t.Fatal(err) }
  ```

- [x] 写 TestBindingCancellation/TestBindingLateCancellation/TestBindingPanicCleanup：测试专用计数 Context.Err 覆盖所有实际观测 Evaluate 检查点（context.Canceled/context.DeadlineExceeded 两种）及 Recheck 最终检查；摘要捕获后取消仍 nil 模型。拥有快照后阶段/最后 panic 传播，调用者不变；结合别名 Clear 与 Task 5 AST 清理登记证明，不假称观察下层未返回/所有不可达副本，不加产品 hook。
- [x] 运行 `& $bindingGo test ./internal/nodecontrol/statebinding -run '^(TestBindingSummary|TestBindingModelOwnership|TestBindingRecheck|TestBindingUnmappedRecheck|TestBindingCancellation|TestBindingLateCancellation|TestBindingPanicCleanup)$' -count=1`，记录缺少摘要/重检 API 的 RED。
- [x] 实现 Summary 的固定状态/责任和 detached 嵌套集合；Recheck 新完整 Evaluate、立即 defer fresh.Clear、reflect.DeepEqual 整个规范 input、比较后再检查取消；不同 BindingMismatch。
- [x] 重跑 targeted 和完整 statebinding，全部 exit 0。
- [x] 临时将 Recheck 完整输入比较省略一次，要求合法替换反例 RED，立即恢复并核对原 SHA256，原样完整 statebinding GREEN；不保留 probe，不提交。

## Task 5: 限定回归、唯一一次最终审查与交付

**Files:** 完善 boundary_test.go；维护基线/progress，创建 execution.md，更新计划/设计/current-status；不新增产品行为。

**Interfaces:** 消费任务 1–4 全部冻结 API/模型。新增 test-only `bindingCleanupRegistered(source []byte) bool`，检查 Evaluate 的模型清理先于下层/后续 ctx.Err，每个下层 owner 清理先于 mapError/ctx.Err；捕获摘要写入已受保护的模型。反例为 late/conditional/missing/wrong-owner defer；不把字符串包含检查当完整 AST 验证。

- [x] 写 TestBindingCleanupRegistration/TestBindingConsumers/TestBindingProductImports；扫描全仓库所有非测试 Go 消费者、验证完整 import 成员白名单，三个旧守卫仍保留原禁止消费者。
- [x] 运行 `& $bindingGo test ./internal/nodecontrol/statebinding ./internal/nodecontrol/statecontract ./internal/nodecontrol/statematerials ./internal/nodecontrol/statecomposition -run '^(TestBindingCleanupRegistration|TestBindingConsumers|TestBindingProductImports|TestContractBindingBoundary|TestSubsetBindingBoundary|TestCompositionBindingBoundary)$' -count=1`，记录缺少 AST helper 或预期守卫断言的 RED。
- [x] 最小实现 test-only AST helper，保持冻结产品接口不变。
- [x] 重跑本任务 targeted 命令，要求 GREEN 并保存实际退出码。
- [x] 仅格式化十五个新 Go 文件及三处指定旧测试，下面数组必须恰为 18 个、且新包文件名与本计划分工一致；单独记录格式化退出码，失败先处理，不用后续命令掩盖。

  ```powershell
  $bindingFormatFiles = @((Get-ChildItem -LiteralPath 'internal/nodecontrol/statebinding' -Filter '*.go').FullName) + @(
      'internal/nodecontrol/statecontract/boundary_test.go',
      'internal/nodecontrol/statematerials/boundary_test.go',
      'internal/nodecontrol/statecomposition/boundary_test.go'
  )
  if ($bindingFormatFiles.Count -ne 18) { throw 'Unexpected formatting scope' }
  & $bindingFmt -w $bindingFormatFiles
  ```

- [x] 独立运行 `& $bindingFmt -l $bindingFormatFiles`，实际 exit 0 且输出必须空。
- [x] 检查旧基线只有三个批准例外、其他文件不变；只读 HEAD/分支/索引不变、diff-check exit 0。保存 post-format source/test freeze。
- [x] 每个完整包独立运行下表命令、逐次保存退出码；不能用一个最后成功代替全部结果。全部必须 exit 0，store 仅编译。等待所有会话结束、复核 freeze 后才审查。

| 范围 | 固定命令 |
| --- | --- |
| statebinding | `& $bindingGo test ./internal/nodecontrol/statebinding -count=1` |
| statecontract | `& $bindingGo test ./internal/nodecontrol/statecontract -count=1` |
| statematerials | `& $bindingGo test ./internal/nodecontrol/statematerials -count=1` |
| statecomposition | `& $bindingGo test ./internal/nodecontrol/statecomposition -count=1` |
| stateformat | `& $bindingGo test ./internal/nodecontrol/stateformat -count=1` |
| stateverify | `& $bindingGo test ./internal/nodecontrol/stateverify -count=1` |
| statepolicy | `& $bindingGo test ./internal/nodecontrol/statepolicy -count=1` |
| staterecovery | `& $bindingGo test ./internal/nodecontrol/staterecovery -count=1` |
| statetime | `& $bindingGo test ./internal/nodecontrol/statetime -count=1` |
| serving | `& $bindingGo test ./internal/nodecontrol/serving -count=1` |
| contracts | `& $bindingGo test ./internal/nodecontrol/contracts -count=1` |
| readiness | `& $bindingGo test ./internal/readiness -count=1` |
| authority（指定） | `& $bindingGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1` |
| store（仅编译） | `& $bindingGo test ./internal/store -run '^$' -count=1` |

- [x] 沿用 Native：作者完成所有实现，无实现代理/逐项独立审查；验证后仅派一次全新、只读、最强可用模型 reviewer 审查本增量及三处跨包边界，包含批准设计/计划、五项 Review Focus、基线例外/freeze、全部退出码、probe 恢复与所有未验证责任；不让 reviewer 改文件/跑测试/联网/再委派，不称为全分支安全验收。代价：既有未提交改动不在本轮独立审查范围。
- [x] 作者逐项裁定审查：Critical/Important 必须复现、最小修复并重跑全部批准回归；设计外扩展先报告。Minor 在本范围内补测试/文档/限定修正或有理由保留，每项明确记录，代价是可能增加定向测试时间。产品改动重跑全部，纯测试至少重跑受影响包和三个守卫；不隐藏阻断、不扩大能力、无第二次独立复审，修复后的独立审查覆盖不因此扩大。
- [x] apply_patch 完成执行记录/checkbox/current-status，报告每项裁定及误用代价、遗留 Minor（若有）、准确测试退出码和所有仍未验证责任；最终哈希/HEAD/索引核对。保留分支/worktree/scratch，不触发合并/清理菜单；没有提交推送，整体 Task10/B01/C12/main 仍未验收。

## 作者自查与执行交接

设计 §1–2/§8 → 任务 1/5；§3–4 → 任务 1/3/4；§5 → 任务 3/5；§6/6.1/6.2 → 任务 2/3；§6.3 → 任务 3/4；§7 → 任务 3/4；§9 → 五项测试循环和最终 14 条命令；§10 → 本计划自查和用户审阅。五项 Review Focus 各有指定测试，所有跨任务接口/私有字段/状态名一致；计划不实现独立 verifier/摘要算法或扩大类型证明。

作者自查不是独立审查。批准的 Native 执行已完成五项任务、全部 RED/GREEN、三个已恢复 probe、一次只读增量审查及两个 Minor 的作者修复与十四条重跑。所有 checkbox 和专用 ledger 已记录真实范围、退出码、最终冻结及保留责任；没有提交推送、第二次独立复审或整体生产验收。
