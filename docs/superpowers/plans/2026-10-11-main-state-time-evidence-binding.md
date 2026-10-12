# Main-state Current-target Time Evidence Binding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将当前目标策略、当前 metadata、显式时间、请求 nonce 和旧/新检查点绑定到完整自有快照，返回可固定快照重检的条件性模型，不产生提交、消费或运行权限。

**Architecture:** 新增纯包 statetimebinding，依次调用 statepolicybinding.Evaluate 和 statetime.EvaluateModel，之间核对当前材料、共同基态和观察关联。只支持授权 root/metadata 及主状态 floor 不变，完整保留下层责任，不改下层算法或接口。

**Tech Stack:** Go 1.26.5、既有 contracts/statepolicybinding/statetime 及值类型、PowerShell、apply_patch、离线缓存；无新依赖。

**Spec:** [已批准书面设计](../specs/2026-10-11-main-state-time-evidence-binding-design.md)，用户于 2026-10-11 在对话中批准；执行前完整阅读设计和本计划。设计文件的待批准状态行属于上轮交付时记录，本轮不改该文件，实施交付时才更新状态。

状态：用户于 2026-10-11 批准本计划；五任务限定 Native 本机实现、修正后十六项验证和只读修正复查完成。checkbox 表示当前实现/验证与交付事项，不宣称早期阶段记录完整：缺失 UTC/逐阶段哈希及补充命令 selector 保留为 M1 审计限制。无 Git 写入或 C12/main 验收；见 [执行记录](../../roadmap/2026-10-11-main-state-time-evidence-binding-execution.md)。

## Global Constraints

- 仅复用 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`、分支 `codex/c12-b01-task9-coordinator`；保留全部既有未提交内容，不新建或清理 worktree。
- 不暂存、提交、推送、合并、重置、删分支或修改 safe.directory/永久环境；阶段证据记录替代技能模板的 Commit 步骤。
- 已初始化 Old 非 nil、FreshInstall=false、旧/新 Desired 为 Present；同节点/authority epoch/既有部署；所选当前 root/metadata/desired；主状态可选 floor 全字段不变。
- 规范化的两个 Main.Root section 完全相同，两个 Main.Metadata section 完全相同，RootTransition 为空；允许完整已接受 0/1/32 历史轮换链，不接受新轮换后缀。
- 完整 metadata 原字节在 Time、Batch、Previous 表、Proposed 表四处相同；完整 Kind/P/E 唯一命中，不重编码，不允许仅比较 P 或更换有效信封。
- 本组合中 `Time.Metadata.StateID == Time.Checkpoint.StateID == Policy.Policy.Old.StateID`，非零；是共同旧基态的显式模型约定，不证明一次物理读取/提交。其他 SourceID/ObservationID 保留独立角色。
- 不推断 DeploymentID=SourceID、TrustAnchorID=锚 P/E、主 guard=CounterID、checkpoint=StateEpoch 或材料创建 sequence；不用直接 enum 转换。
- 两个时间观测明确 Available，ProviderID/ObservationID/规范化 Now/Floor 精确相同；OldFloor≥主可选约束、ObservedFloor≥OldFloor、Now≥ObservedFloor，三种 floor 不强制相等或写回。
- Time 四种观测任一非 Available→AssuranceMissing；完整 statetime 检查仍负责格式、非零身份、key、签名、半开窗及五分钟 freshness；签名/HTTP Date 不生成 Now 或 floor。
- Time 旧检查点=Main.Previous.Checkpoint；已验签新检查点=Main.Proposed.Checkpoint；AdvanceRequired 恰等于新>旧。无 max/clamp/自动写入。
- 全零 32-byte nonce 仍是合法格式；RequestID/ObservationID 未新增签入 wire。相同未消费模型重复成功不代表 TLS、防重放或请求消费。
- Binding 两表各 Materials/Expected≤128；每表各类 RootRotation≤32，Target.Links/RootTransition 各≤32；v1 锚语义及原 wire 限制不扩大。
- 每个表 RootSet raw≤65,536 bytes，其他 raw/独立锚≤262,144 bytes；每表≤16 MiB；Binding 总 raw≤32 MiB + 262,144 bytes。
- 每个 Expectation 全部分支 root keys/metadata keys/revoked/slots≤5/64/4096/8；两份 Main 各自 root keys/metadata keys/revoked≤5/64/4096，bootstrap/agent deauthorized 各≤128，Desired slots/LKG entries/local faults/supervisor faults≤8/8/64/16。
- Batch rotations≤32，每次出现 raw≤262,144 bytes，总 Batch≤36 × 262,144 bytes；完整 Policy 输入≤32 MiB + 37 × 262,144 bytes。
- Old.Roots≤33，每个 root keys≤5；Old.Keys/Old.RevokedIDs/History 各≤4096；key Status/RevocationReason≤8/32 bytes，History.IncidentKind≤64 bytes。
- Time 两个 raw 各≤262,144 bytes，合计≤524,288 bytes；本层总 raw≤32 MiB + 39 × 262,144 bytes。按每次出现计，不按别名去重，累计使用 `len > limit-total`，预算不是峰值内存承诺。
- 任何深复制/索引分配前完整预查上述所有分支；不能丢未选项、未投影域、历史 key/tombstone。仅两表顶层 Materials/Expected 按 Kind/P 排序，空集合→nil，其余顺序不变。
- 全部 time.Time 去单调表示并转 UTC，保留纳秒及真正零值；不取整、补时间或读取系统时钟，主状态 UnixSeconds/raw 不变。
- 产品唯一判定入口是 statepolicybinding.Evaluate/statetime.EvaluateModel；不读 Candidate.State，不直接调用其他 Evaluate/Verify/RecheckModel，不直接导入 stateformat/crypto/JSON/JCS。
- 失败/取消返回 nil，panic 清理后传播；清理登记先于后续 context/判定/映射，摘要直接交受保护 owner；不加产品 hook，不承诺 OOM/致命终止、字符串/不可达/全进程擦除或并发访问。
- Model 只含 valid/input/summary，状态 ConditionalTimeEvidenceBound；所有下层责任及五项附加责任固定保留，ValidUntil 取两个下层较早值，无 Ready/State/Commit/Authorize API。
- Recheck 重新完整 Evaluate 并比较所有规范化输入；Now/Floor/ObservationID 即使合法前进 1 ns 也需新 Evaluate；不借下层宽松时间重检接受变化。
- 下层全部产品/接口/wire/算法、其他既有测试、依赖/锁文件、SQL/生成文件、captured vectors 不变；既有代码修改仅八份指定消费者边界测试。
- 无产品消费者/provider 实现、外部 I/O/网络/存储/随机/环境/硬件/重试/goroutine/动态反射；无全仓/race/真实 DB/Docker/C11/C12/bootstrap/PITR 验证。
- 手工编辑仅 apply_patch，gofmt 仅 20 个新 Go 文件和八份旧边界；所有启动会话等待实际退出，不把输出发布延迟当成测试超时或重复启动 pending 测试。

## Review Focus

1. 两套现有测试的同名 private(n) 使用不同 seed 构造；时间证据必须使用所选 metadata 的真实 key，而非另套独立合法样例（任务 1：TestTimeBindingFixtureControls）。
2. 四处 metadata/二维 rotations 共享调用者 backing store，且同 P 不同 E 可独立验签成功；每次出现必须独立拥有，精确信封不得替换（任务 1/2：TestTimeBindingOwnedAliases、TestTimeBindingAlternateMetadataEnvelope）。
3. root/metadata 的 history/set digest/window/有序集合变化而高水位未变；本层必须比较完整 section，不能把候选标为已提交（任务 2：TestTimeBindingUnsupportedSections）。
4. 只改变未签入 wire 的 RequestID/ObservationID 或一致替换 nonce/证明；新 Evaluate 可成立，但旧 Recheck 不同，重复未消费输入不具防重放能力（任务 3/4：TestTimeBindingRequestSemantics、TestTimeBindingObservationRecheck）。
5. 摘要捕获后或最终检查处取消/panic；不得泄出部分 owner、延迟登记 cleanup 或提前转移，固定源码形状检查与可观察擦除证据必须分开（任务 4：TestTimeBindingLateCancellation、TestTimeBindingPanicPropagation、TestTimeBindingCleanupRegistration）。

---

## 文件分工、接口与证据范围

新增目录 `internal/nodecontrol/statetimebinding/`，固定 20 个 Go 文件：

| 产品文件（10） | 责任 |
| --- | --- |
| doc.go / input.go | 条件边界 / 完整输入与固定摘要类型 |
| errors.go | 20 类固定错误、PolicyFailure 与闭合映射 |
| budget.go / owned.go | 完整预复制预算 / 拥有、规范化、清理及摘要复制 |
| binding.go | 支持范围、完整 section、精确原材料及下层摘要坐标 |
| observations.go | 共同基态、UUID/epoch、provider/时间/观察身份 |
| checkpoint.go | 旧检查点关联、已验签摘要/新检查点对应 |
| evaluate.go / model.go | 完整组合流程 / 私有 owner、Summary、Clear、Recheck |

新测试（10）：fixtures_test.go、budget_test.go、owned_test.go、errors_test.go、
binding_test.go、observations_test.go、checkpoint_test.go、evaluate_test.go、
model_test.go、boundary_test.go。测试签名/变体只在内存中构造，不新增向量文件。

八份旧测试（完整冻结修改列表）：

```text
internal/nodecontrol/statepolicybinding/boundary_test.go
internal/nodecontrol/statetime/boundary_test.go
internal/nodecontrol/statebinding/boundary_test.go
internal/nodecontrol/statepolicy/boundary_test.go
internal/nodecontrol/stateverify/boundary_test.go
internal/nodecontrol/statecontract/boundary_test.go
internal/nodecontrol/statematerials/boundary_test.go
internal/nodecontrol/statecomposition/boundary_test.go
```

只增加准确目录且包名 statetimebinding 的消费者判断/正反例；stateverify
只变其 statepolicy 分支。statetime 扫描路径先 filepath.Rel 转仓库相对值。
不扩大既有产品 import/member 白名单或 cleanup 规则。

八份新正反例测试名依次固定为 TestPolicyBindingTimeBindingBoundary、
TestTimeBindingBoundary、TestBindingTimeBindingBoundary、
TestPolicyTimeBindingBoundary、TestStateverifyTimeBindingBoundary、
TestContractTimeBindingBoundary、TestSubsetTimeBindingBoundary、
TestCompositionTimeBindingBoundary，对应上述八个包的相同顺序。
每个包分别用 `& $timeBindingGo test $timeBindingGuardPackage -run 'TimeBindingBoundary$' -count=1 -timeout 120s -p=1`；
timeBindingGuardPackage 只能逐项取上述八个目录（以 `./` 开头、去掉
`/boundary_test.go`），每项独立命令/输出/退出码，不能用一个汇总退出码
代替八项结果。

公共接口固定为：

```go
type Input struct { Policy statepolicybinding.Input; Time statetime.ModelInput }
type ResultStatus uint8 // 0 InvalidResult; 1 ConditionalTimeEvidenceBound
type Obligation uint8 // 0 UnknownObligation; 1..5 按设计顺序
type ObservationBinding struct {
    BaseStateID, MetadataSourceID, MetadataObservationID contracts.Digest
    RequestObservationID, CheckpointSourceID, CheckpointObservationID contracts.Digest
}
type Summary struct {
    Status ResultStatus
    Policy statepolicybinding.Summary
    Time statetime.EvidenceSummary
    Observations ObservationBinding
    ValidUntil time.Time
    Obligations [5]Obligation
}
func Evaluate(ctx context.Context, in Input) (*Model, error)
func Recheck(ctx context.Context, model *Model, in Input) error
func (model *Model) Summary() Summary
func (model *Model) Clear()
```

五项 Obligation 依次 ObservationAndBaseProvenance、TrustedTimeAndTLS、
AtomicRequestConsumption、DurableCheckpointCommit、
FullMainStateAndRuntimeAuthorization。Model 私有字段仅 `valid bool`、
`input Input`、`summary Summary`；不保留下层 owner。

errors.go 公开 `Err<Class>`：InvalidInput、LimitExceeded、IncompleteCoverage、
InvalidMaterial、MissingMaterial、DuplicateMaterial、BindingMismatch、
SignatureMismatch、Rollback、Conflict、ClearedModel、UnsupportedPolicyShape、
PolicyRejected、UnsupportedBindingShape、UnauthorizedKey、AssuranceMissing、
InvalidTime、Expired、CheckpointRollback、StaleEvidence，共 20 类，不合并
UnsupportedPolicyShape/UnsupportedBindingShape 或 Rollback/CheckpointRollback。
文字为 `statetimebinding: ` 加类别小写空格词；不携带输入。
`PolicyFailure struct { Code statepolicy.Code; Reason statepolicy.Reason }`，
`Error() string` 固定同 ErrPolicyRejected，`Unwrap() error` 返回本层 sentinel。
mapError 先当前 context、标准取消，再 errors.As 复制非 nil 合法下层
PolicyFailure（Code 1–5、Reason 1–14），最后同义 sentinel；typed-nil/未知/
越界诊断 InvalidInput。不能先匹配 Unwrap 而丢诊断。

仅实施后允许维护：本计划真实 checkbox、设计状态与执行记录链接、
`docs/roadmap/2026-10-11-main-state-time-evidence-binding-execution.md`、
current-status.md 的本增量摘要，以及专属台账
`.superpowers/sdd/2026-10-11-main-state-time-evidence-binding/`。
不提前写完成，不覆盖旧执行记录或旧 scratch。

## 固定环境、基线与停止条件

所有命令从上述 worktree 执行，PowerShell、login=false。规划时 Go/gofmt
文件存在，但本轮没有执行它们。每个未来测试会话显式设置进程环境：

```powershell
$env:GOENV='off'; $env:GOWORK='off'; $env:GOTOOLCHAIN='local'
$env:GOPROXY='off'; $env:GOSUMDB='off'; $env:GOAUTH='off'
$env:GOVCS='all:off'; $env:GOFLAGS='-mod=readonly'
$env:GOPATH='C:/Users/Lenovo/go'; $env:GOMODCACHE='C:/Users/Lenovo/go/pkg/mod'
$env:GOCACHE='D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/.superpowers/sdd/2026-10-11-main-state-time-evidence-binding/go-cache'
$timeBindingGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
$timeBindingFmt='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/gofmt.exe'
```

专属台账/cache 在实施启动时才创建；已有则读取归属/进度后续用，绝不
覆盖未知文件。环境仅限该进程，不修改系统/Powershell profile。缓存不可
用则先报告实际环境限制，不静默改超时、下载依赖或伪称测试 RED。

最近一次已核对的 HEAD 为 `6752b9612a00b9d6b0bf87ae8dd95063f6b056ee`，
索引 SHA256 为 `1CBA46BEF4C1AFDC9AE7938146909F0D2D641A0D712CE18036B60FE5B8673D12`。
执行前重新核对 branch/HEAD/index、已有全部源/SQL/module/lock/向量文件
哈希、当前未提交清单及两份审批文档；已有 26 份旧增量冻结并非本轮可
替换的基线，八个旧守卫也先保留本轮开始时原字节。

只读 Git 诊断设置 GIT_OPTIONAL_LOCKS=0、core.fsmonitor=false；若需要提权
依授权审查机制处理，不设置 safe.directory、不写索引。诊断拒绝不是
源码 RED；完成不受影响工作，明确剩余阻塞，不能伪称验证成功。

所有阶段将工具实际输出、退出码、开始/结束、源码哈希和 RED/GREEN
结果用 apply_patch 存档；不借 shell 重定向写文件。测试 RED 只能是预期
缺接口/真实断言失败，不包括缺依赖/工具拒绝/语法或签名样例错误。
下层行为矛盾、范围外缺陷、新增能力/依赖需求，均先停止并报告书面修订；
不得绕过平台拒绝、复制既有审查结论或用无限重试换成功。

## Task 1: 完整拥有输入、固定错误与精准隔离

**Files:** Create doc.go、input.go、errors.go、budget.go、owned.go，以及
fixtures_test.go、budget_test.go、owned_test.go、errors_test.go、boundary_test.go；
Modify 上述八份旧消费者测试。全部新文件均位于固定新包目录。

**Interfaces:** 无前置任务。产生公共值类型和 PolicyFailure，内部：

```go
func checkBudget(in Input) error
func bindingBudget(in statebinding.Input) (rawTotal int, err error)
func contentBudget(in statecontract.MainStateContentModel) bool
func ownInput(ctx context.Context, in Input) (out Input, err error)
func ownPolicyBinding(ctx context.Context, in statepolicybinding.Input) (out statepolicybinding.Input, err error)
func ownBinding(ctx context.Context, in statebinding.Input) (out statebinding.Input, err error)
func ownPolicy(ctx context.Context, in statepolicy.ModelInput) (out statepolicy.ModelInput, err error)
func ownTime(ctx context.Context, in statetime.ModelInput) (out statetime.ModelInput, err error)
func copyContent(ctx context.Context, in statecontract.MainStateContentModel) (out statecontract.MainStateContentModel, err error)
func copyTable(ctx context.Context, in statematerials.Input) (out statematerials.Input, err error)
func copySlice[T any](in []T) []T
func copyRaw(in []byte) []byte
func normalizeTime(in time.Time) time.Time
func clearInput(in *Input)
func clearPolicyBinding(in *statepolicybinding.Input)
func clearBinding(in *statebinding.Input)
func clearPolicy(in *statepolicy.ModelInput)
func clearTime(in *statetime.ModelInput)
func clearContent(in *statecontract.MainStateContentModel)
func clearTable(in *statematerials.Input)
func mapError(ctx context.Context, err error) error
```

本任务同时产生 test-only `func timeBindingImportsAllowed(source []byte) bool`
与 `func timeBindingConsumerAllowed(path, pkg, imported string) bool`，任务 4
补全精确产品文件扫描及清理结构守卫，不再另起同名 helper。

复制仅复用现有 owned.go/budget.go 的值处理模式，不导出/改下层私有 helper。
clearPolicy 使用自有 Old.Clear；顺序、所有字段及规范化覆盖不得缩水。
时间证明测试 builder 沿用 statetime/fixtures_test.go 的 canonical payload、
两层 schema 与 `TALENRO-NODE-TIME-ATTESTATION-V1` 加 NUL 的 transcript；
只换成所选 metadata 的已知 fixture key，不重写产品验签算法。

- [x] 保存本轮 baseline.json、八旧守卫原字节/哈希和 progress.md；确认新目录不重叠、旧进度无漂移、两份文档审批已真实完成。
- [x] 创建 test-only fixture helper `timeBindingFixture(t *testing.T, rotations int, proposedCheckpoint uint64) Input`、`cloneFixture(t *testing.T, in Input) Input`、`signAttestation(t *testing.T, in *Input, sequence uint64, nonce [32]byte)`、`checkLowerModels(t *testing.T, in Input) statepolicybinding.Summary`；只改新 fixtures_test.go。最后一个 helper 分别完整运行两下层并立即登记 Clear，返回独立 Policy.Summary，Time 下层必须同时成功。适配既有 policyBindingFixture(t,n,n)，不用下层私有函数跨包调用；同目录内存复制/改名其必要测试 builder，不移动旧文件。节点 `12345678-1234-4234-8234-123456789abc`、epoch 19、旧检查点 100、Now `2020-01-01T01:00:00Z`、旧/观测 floor `2019-12-31T00:00:00Z`；时间证明 IssuedAt=Now、ValidUntil=Now+5m。使用 statepolicybinding 样例 seed 首字节 2、其余零的已授权 metadata key，不能采用 statetime 的重复 byte seed。三个 StateID 等于真实 fixture.Old.StateID，其余来源 ID 明确非零独立；Time.MetadataEnvelope 是同一 current Batch raw，ExpectedNonce 默认全零。
- [x] 写 TestTimeBindingFixtureControls：0/1/32 完整历史链、100/101 检查点时分别完整运行两个下层，二者各返回非 nil/err=nil，signer public key/key ID 精确存在于当前 metadata；下层 owner 立即 Clear。fixture 校验失败先修测试材料，不记为本层 RED。
- [x] 写 TestTimeBindingBudget/TestTimeBindingBudgetAllBranches/TestTimeBindingOwnedAliases/TestTimeBindingNormalization/TestTimeBindingErrorMapping 和八份 Test*TimeBindingBoundary：预算每项等长/+1，包括所有 expectation 非活动分支、Main 十九集合、Old 各 root key、History 字符串及两个 Time raw。合法语义不是 budget 断言前提；原 bytes 共享四处 metadata/active root/rotation，复制后各自独立且 caller 不变，清理归零；零时间保留、纳秒保留、两表顶层重排等价、其他顺序保持。错误覆盖全部 20 类、wrapped/typed-nil/越界 PolicyFailure、未知带敏感文字 error、context 取消/超时优先。

  ```go
  over := cloneFixture(t, timeBindingFixture(t, 1, 100))
  over.Time.MetadataEnvelope = make([]byte, 262145)
  if !errors.Is(checkBudget(over), ErrLimitExceeded) { t.Fatal("missing Time raw limit") }
  var nilPolicy *statepolicybinding.PolicyFailure
  if !errors.Is(mapError(context.Background(), nilPolicy), ErrInvalidInput) { t.Fatal("typed nil") }
  owned, err := ownInput(context.Background(), in)
  if err != nil { t.Fatal(err) }; defer clearInput(&owned)
  before := bytes.Clone(in.Time.MetadataEnvelope)
  owned.Time.MetadataEnvelope[0] ^= 1
  if !bytes.Equal(in.Time.MetadataEnvelope, before) { t.Fatal("caller alias") }
  if bytes.Equal(owned.Time.MetadataEnvelope, owned.Policy.Policy.Batch.MetadataEnvelope) { t.Fatal("owned occurrence alias") }
  ```

- [x] 运行 RED：`& $timeBindingGo test ./internal/nodecontrol/statetimebinding -run '^(TestTimeBindingFixtureControls|TestTimeBindingBudget.*|TestTimeBindingOwnedAliases|TestTimeBindingNormalization|TestTimeBindingErrorMapping)$' -count=1 -timeout 120s -p=1`，预期缺 checkBudget/ownInput/mapError 或真实缺失行为；八份新消费者正例分别运行 `-run 'TimeBindingBoundary$'`，预期旧守卫拒绝准确目录。逐包独立存档，不把环境失败当 RED。
- [x] 在 doc.go/input.go/errors.go 实现固定边界、类型和 20 错误/mapError；先处理 context，再 PolicyFailure 再普通 sentinel，文字和枚举范围严格按冻结接口。
- [x] 在 budget.go/owned.go 实现完整预算、owner/cleanup 和规范化，采用本任务逐项签名；全部分支预算先于复制，每个可变副本立即受清理保护。
- [x] 只修改八处精确消费者例外，保留子目录/相似前缀/错误包名/runtime/reader/serving/authority/store 反例，stateverify 的直接 verifier 权限仍拒绝。实现本任务的新 boundary 基础 helper，能力仅设计 §9，不预留新 parser/时钟权限。
- [x] 重跑 Task 1 同样命令 GREEN，要求 fixture 两下层正控制、owner/预算/错误及八旧正反例均 exit 0；记录阶段哈希/实际证据，不提交。

## Task 2: 当前材料、共同基态、时间和检查点的纯对应

**Files:** Create binding.go、observations.go、checkpoint.go、binding_test.go、
observations_test.go、checkpoint_test.go；扩充新 fixtures_test.go，不改下层。

**Interfaces:** 消费 Task 1 的完整 owned Input 和固定错误；产生：

```go
func findMaterial(ctx context.Context, table statematerials.Input, ref statematerials.MaterialRef) (statematerials.Material, error)
func currentMetadataRef(in Input) statematerials.MaterialRef
func bindScope(ctx context.Context, in Input, policy statepolicybinding.Summary) error
func bindMetadata(ctx context.Context, in Input, policy statepolicybinding.Summary) error
func bindObservations(ctx context.Context, in Input, policy statepolicybinding.Summary) error
func bindOldCheckpoint(ctx context.Context, in Input) error
func bindEvidence(ctx context.Context, in Input, evidence statetime.EvidenceSummary) error
```

currentMetadataRef 仅封闭 Metadata 类型投影，调用前已有完整下层成功；
findMaterial 返回同一 owned raw 的只读借用，不清理它、不再复制/解析。
Kind/P 唯一命中且 E 精确相等；先前预算禁止无界索引，使用有界顺序查找。

- [x] 写 TestTimeBindingScope/TestTimeBindingUnsupportedSections：0/1/32 已接受链成立；合法新增 root/metadata 在 selected-policy 层先成立，本层 UnsupportedBindingShape，合法下层历史/首次/floor 场景保留 UnsupportedPolicyShape；下层回退/签名失败不被重分类。两个完整 Root/Metadata section 任一字段变化均不能仅靠 mark 判等；合法 history/set digest 声明变化先证下层成立，不伪造成功控制。
- [x] 写 TestTimeBindingMetadata/TestTimeBindingAlternateMetadataEnvelope：四处 raw、Kind/P/E、两表唯一命中和版本/高水位；使用旧 multiRootPolicyBindingFixture 的两 key/threshold 1 测试构造，另一授权 root key 3 对相同 metadata payload 重新签名，Time 单独和 selected-policy 单独都有效而组合错 E/raw；不能用破损 JSON/无效签名替代这个负例。

  ```go
  policySummary := checkLowerModels(t, replacement)
  if replacement.Time.Metadata.Payload != original.Time.Metadata.Payload { t.Fatal("changed P") }
  if replacement.Time.Metadata.Envelope == original.Time.Metadata.Envelope { t.Fatal("unchanged E") }
  if !errors.Is(bindMetadata(context.Background(), replacement, policySummary), ErrBindingMismatch) { t.Fatal("alternate E accepted") }
  ```

- [x] 写 TestTimeBindingObservations/TestTimeBindingTimeSnapshot/TestTimeBindingCheckpoint：两个 Time.StateID 一起变仍让 statetime 合法，却不同 Old.StateID→BindingMismatch；各独立 SourceID 可不同、观测缺席→AssuranceMissing，node/epoch/provider/time ObservationID/Now/Floor 一方变化→BindingMismatch。Time 旧 checkpoint 100、Main 旧 100；signed/Main 新 100/101；不相等、错误 AdvanceRequired/RequestID/nonce、越过旧/新身份都失败。StateEpoch=82 与 checkpoint=100 不强求相等；全零 nonce 合法。只签摘要的新值不能补写 Main。
- [x] 运行 `& $timeBindingGo test ./internal/nodecontrol/statetimebinding -run '^(TestTimeBindingScope|TestTimeBindingUnsupportedSections|TestTimeBindingMetadata|TestTimeBindingAlternateMetadataEnvelope|TestTimeBindingObservations|TestTimeBindingTimeSnapshot|TestTimeBindingCheckpoint)$' -count=1 -timeout 120s -p=1`，记录上述 helper 缺失或对应断言 RED。
- [x] 在 binding.go 实现 scope 的完整 section DeepEqual/空 RootTransition/旧链完整对应，先确认真实 lower status、Main/selected/current 坐标；metadata 原 bytes/完整 ref/两表版本对应，采用有界查找，不解析/验签。
- [x] 在 observations.go 实现闭合状态分支、共同旧 StateID、规范 UUID.String/node/epoch、四方 provider、精确 time 快照，不发明独立来源等式或自动补时间。
- [x] 在 checkpoint.go 实现 old checkpoint 和 evidence 逐字段规则，严格按设计 §6–7，不自动补写/截断/选择另一份证明。
- [x] 重跑同样 targeted GREEN 及新包现有完整测试，均 exit 0；保留原调用者及完整下层义务，无提交。

## Task 3: 完整顺序组合、条件摘要与 Model 清理

**Files:** Create evaluate.go、model.go、evaluate_test.go、model_test.go；
完善新 boundary_test.go 的判定入口白名单，不扩展产品权限。

**Interfaces:** 消费 Task 1–2 helper；产生公共 Evaluate、Summary、Clear；
Evaluate 实现使用命名返回值 `(result *Model, err error)`，公开函数类型不变。
辅助：`func finishSummary(in Input, policy statepolicybinding.Summary, evidence statetime.EvidenceSummary) Summary`、
`func cloneSummary(in Summary) (out Summary)`、`func clearSummary(in *Summary)`、
`func clonePolicySummary(in statepolicybinding.Summary) (out statepolicybinding.Summary)`、
`func clearPolicySummary(in *statepolicybinding.Summary)`。
仅复制 Policy.Binding.Proposed.Target.Links/Coverage 等可变摘要；Time 是值。
finishSummary 只作固定值赋值与较早期限选择，无分配/回调/解析。

- [x] 写 TestTimeBindingEvaluate/TestTimeBindingLowerRejections：0/1/32 链、Desired no-change/合法更新、checkpoint 100/101；完整 metadata/root/floor 原值不变。所有下层错误返回 nil Model，含历史缺证/冲突、紧急撤销、过期、key 未授权、坏 signature/格式、checkpoint 回退；先证对应 lower 控制的预期类别，不以其他错误掩盖。
- [x] 写 TestTimeBindingSummary/TestTimeBindingRequestSemantics：嵌套 Policy 摘要逐值/集合/责任保留、Time 原摘要、ObservationBinding 六字段精确、ValidUntil 恰 min；五项义务固定，不抹掉 TimeAttestationAndFreshness。全零 nonce 正例、重复 Evaluate 可重复成功；RequestID/ObservationID 只是观测身份，checkpoint 前进仍可 Policy.SelectedPolicyNoChange=true，绝不输出 WholeStateNoChange/Consumed/Ready。

  ```go
  in := timeBindingFixture(t, 1, 101)
  model, err := Evaluate(context.Background(), in)
  if err != nil || model == nil { t.Fatal(err) }; defer model.Clear()
  s := model.Summary()
  if s.Status != ConditionalTimeEvidenceBound || s.Time.OldCheckpoint != 100 || s.Time.ProposedCheckpoint != 101 || !s.Time.AdvanceRequired { t.Fatal("checkpoint binding") }
  want := [5]Obligation{ObservationAndBaseProvenance, TrustedTimeAndTLS, AtomicRequestConsumption, DurableCheckpointCommit, FullMainStateAndRuntimeAuthorization}
  if s.Obligations != want { t.Fatal("lost responsibility") }
  wantUntil := s.Policy.ValidUntil
  if s.Time.ValidUntil.Before(wantUntil) { wantUntil = s.Time.ValidUntil }
  if !s.ValidUntil.Equal(wantUntil) { t.Fatal("not the exact minimum deadline") }
  ```

- [x] 写 TestTimeBindingSummaryOwnership/TestTimeBindingModelClear：每份 detached Links/Coverage 都独立；改变调用者所有可变集合/raw/Old/history 不影响模型。持有私有 input/summary 可达 backing-store 引用，Clear 两次全部零且 caller/独立摘要不变；nil/zero/cleared Summary 全零。不宣称擦掉 strings/不可达副本。
- [x] 运行 `& $timeBindingGo test ./internal/nodecontrol/statetimebinding -run '^(TestTimeBindingEvaluate|TestTimeBindingLowerRejections|TestTimeBindingSummary|TestTimeBindingRequestSemantics|TestTimeBindingSummaryOwnership|TestTimeBindingModelClear)$' -count=1 -timeout 120s -p=1`，记录缺完整 Evaluate/owner 或对应行为 RED。
- [x] 实现 Evaluate：ownInput→Model 接管并立即 defer(result==nil 时 Clear)→statepolicybinding.Evaluate/立即 defer Clear→Summary 直接写 model.summary.Policy→scope→metadata→observations→oldCheckpoint→statetime.EvaluateModel/立即 defer Clear→Time 固定值捕获→bindEvidence→finishSummary→最终 context→valid/return。下层 owner 登记必须先于 mapError/context。期限均非零有效，不把候选当 committed。
- [x] 实现 Summary/cloneSummary/Clear：新复制立刻受失败 defer 保护，只在 valid 返回独立摘要；Clear 清零完整 owner，nil-safe/幂等。不加入 Recheck stub/Ready 等临时 API，不保留下层 owner。
- [x] 重跑同 targeted 和新包完整测试 GREEN，记录 source freeze 与两下层责任未消除；无提交。

## Task 4: 全输入固定快照重检、取消/panic 与源码守卫

**Files:** Modify 新 model.go；扩充新 model_test.go、owned_test.go、
evaluate_test.go、boundary_test.go。产品修正只能在固定新包内，八旧测试
若需修正仅原批准消费者范围，不能增加下层产品/test hook。

**Interfaces:** 消费 Task 3 有效 Model 和 Task 1 的能力/消费者 helper；
产生公共 Recheck 与 test-only
`func timeBindingCleanupRegistered(source []byte) bool`；不是产品 API。

- [x] 写 TestTimeBindingRecheck/TestTimeBindingObservationRecheck/TestTimeBindingFixedTime：全相同输入、两表顶层重排、nil/empty、同瞬时时区/单调表示通过；合法全套时间前进 1ns/提高 Floor/改同一 time ObservationID、新请求 ID/ObservationID/nonce+重签、新共同 StateID、独立来源、Old CounterID/StateEpoch、未选合法材料/期望、未投影域、History 字段/有意义顺序变动拒绝旧模型。每个合法替换先新 Evaluate 成功；若下层拒绝该变化，保留下层类别，不伪称比较覆盖。

  ```go
  replacement := cloneFixture(t, in)
  replacement.Time.Time.Now = replacement.Time.Time.Now.Add(time.Nanosecond)
  replacement.Policy.Policy.Time.Now = replacement.Time.Time.Now
  fresh, err := Evaluate(context.Background(), replacement)
  if err != nil || fresh == nil { t.Fatal(err) }; defer fresh.Clear()
  if !errors.Is(Recheck(context.Background(), prior, replacement), ErrBindingMismatch) { t.Fatal("forward time reused") }
  if !errors.Is(Recheck(context.Background(), nil, in), ErrClearedModel) { t.Fatal("nil owner") }
  ```

- [x] 写 TestTimeBindingCancellation/TestTimeBindingLateCancellation/TestTimeBindingPanicPropagation：test-only Context.Err 根据本次成功轨迹逐点取消/超时/正常 panic，包含复制、两个 lower、可变摘要捕获后、最终返回前及 Recheck 比较后。取消/失败 nil Model，panic 传播且 caller 不变；不硬编码下层调用次数。私有失败 owner 不可观测时仅记录源码早清理证据，另用 Clear 的 backing-store 实验，不伪造擦除实测。
- [x] 写 TestTimeBindingCleanupRegistration/TestTimeBindingProductImports/TestTimeBindingConsumers：AST 固定检查输入保护、两个 lower 立即 defer Clear、可变 Policy.Summary 直接交 protected Model、copy/clone 初始失败清理、最后 valid/return、Recheck fresh 立即 cleanup；missing/late/conditional/wrong-owner defer、未保护临时摘要、提前 transfer 均拒绝。产品 member/文件名白名单恰十份，二十新 Go 文件最终清单准确；无非测试产品消费者。两份 lower 仍唯一判定入口，无 State()/其他 Evaluate/parser/clock/副作用，dot/blank/duplicate import 和 goroutine 反例仍拒绝。
- [x] 运行 `& $timeBindingGo test ./internal/nodecontrol/statetimebinding -run '^(TestTimeBindingRecheck|TestTimeBindingObservationRecheck|TestTimeBindingFixedTime|TestTimeBindingCancellation|TestTimeBindingLateCancellation|TestTimeBindingPanicPropagation|TestTimeBindingCleanupRegistration|TestTimeBindingProductImports|TestTimeBindingConsumers)$' -count=1 -timeout 120s -p=1`，记录缺 Recheck/固定 AST helper 或真实缺清理结构 RED；故意改写源码字符串为 unsafe mutation，仅在 test 内存里，不改真实源码来制造 RED。
- [x] 在 model.go 实现 Recheck：context/nil/cleared→完整 Evaluate→立即 defer fresh.Clear→DeepEqual 两个完整规范 Input→最终 context→差异 BindingMismatch。不调用 lower RecheckModel，不只比 Summary。
- [x] 在 boundary_test.go 实现清理 AST/helper 并补全能力/文件/消费者扫描，明确固定结构限制，不假称形式化证明；若真实源码登记缺失，先保留 RED，再仅在新包最小补齐，不放宽守卫。
- [x] 重跑同 targeted 和完整新包 GREEN；八旧消费者包逐包 `-run 'TimeBindingBoundary$'` 仍 exit 0，保存阶段证据，不提交。

## Task 5: 十六项局部验证、一次 Native 只读复核与交付记录

**Files:** 无新增产品行为；维护专属 baseline/progress/命令证据/冻结；实施后
创建执行记录并有限更新设计/计划状态、current-status。无其他仓库变更。

**Interfaces:** 消费 Task 1–4 冻结 API/20 新文件/八旧守卫及真实 RED/GREEN；
交付限定验证、增量审查裁定和最终不变性核对，不赋予生产/Git 权限。

- [x] 核对新目录恰 20 个指定 Go 文件，无新 vectors/未知文件；旧代码只八份批准守卫改变。构造显式 formatting 数组（20 新文件+八旧文件）并断言 Count=28，才运行 `& $timeBindingFmt -w $timeBindingFormatFiles`；其他源码/文档不格式化。
- [x] 独立运行 `& $timeBindingFmt -l $timeBindingFormatFiles`，实际 exit 0 且输出为空；保存 post-format-source-test-freeze.json（28 文件及旧全范围/vector 哈希），只读 Git/索引核对，不提交。
- [x] 下列 16 条逐条独立执行，均使用固定进程环境；记录完整实际退出码和输出，所有会话确实结束才验收。单命令 timeout=120s，不静默提高超时、加重试/race或扩大测试。

| 范围 | 固定命令 |
| --- | --- |
| statetimebinding | `& $timeBindingGo test ./internal/nodecontrol/statetimebinding -count=1 -timeout 120s -p=1` |
| statepolicybinding | `& $timeBindingGo test ./internal/nodecontrol/statepolicybinding -count=1 -timeout 120s -p=1` |
| statebinding | `& $timeBindingGo test ./internal/nodecontrol/statebinding -count=1 -timeout 120s -p=1` |
| statecontract | `& $timeBindingGo test ./internal/nodecontrol/statecontract -count=1 -timeout 120s -p=1` |
| statematerials | `& $timeBindingGo test ./internal/nodecontrol/statematerials -count=1 -timeout 120s -p=1` |
| statecomposition | `& $timeBindingGo test ./internal/nodecontrol/statecomposition -count=1 -timeout 120s -p=1` |
| stateformat | `& $timeBindingGo test ./internal/nodecontrol/stateformat -count=1 -timeout 120s -p=1` |
| stateverify | `& $timeBindingGo test ./internal/nodecontrol/stateverify -count=1 -timeout 120s -p=1` |
| statepolicy | `& $timeBindingGo test ./internal/nodecontrol/statepolicy -count=1 -timeout 120s -p=1` |
| staterecovery | `& $timeBindingGo test ./internal/nodecontrol/staterecovery -count=1 -timeout 120s -p=1` |
| statetime | `& $timeBindingGo test ./internal/nodecontrol/statetime -count=1 -timeout 120s -p=1` |
| serving | `& $timeBindingGo test ./internal/nodecontrol/serving -count=1 -timeout 120s -p=1` |
| contracts | `& $timeBindingGo test ./internal/nodecontrol/contracts -count=1 -timeout 120s -p=1` |
| readiness | `& $timeBindingGo test ./internal/readiness -count=1 -timeout 120s -p=1` |
| authority（限定） | `& $timeBindingGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1 -timeout 120s -p=1` |
| store（仅编译） | `& $timeBindingGo test ./internal/store -run '^$' -count=1 -timeout 120s -p=1` |

- [x] 16 项全部 exit 0 且冻结无漂移后，沿用 Native：作者本机完成，无实现代理/逐任务独立审查；仅此时派一个全新独立上下文、只读最强可用 reviewer 审本增量（目前可用 gpt-6-astra / xhigh，派发时重新核对工具可用名单）。提供批准设计/计划、五项 Review Focus、20 新文件和八旧守卫实际基线差异、阶段结果/16 条实际证据及全部限制。reviewer 不改文件、不跑测试/外部诊断、不联网、不委派，不用旧 reviewer/旧审查代替。本轮写计划不派代理。
- [x] 作者逐项裁定 Critical/Important/Minor。Critical/Important 必须在既定范围最小修复并重跑完整 16 项；范围外先暂停提出修订。Minor 修复或记录原因/代价；任何源/测试变化均重跑最终 16 项、更新冻结，同一 reviewer 可只读复查修正点。留存首次审查覆盖及修正版覆盖事实，不能声称旧审查自动覆盖后改源码；未解决阻断明确保留。
- [x] apply_patch 保存真实执行记录、progress、checkbox、设计状态和 current-status 的本增量摘要。列所有实际命令/失败尝试/退出码、审查裁定/限制、nonce 未消费/时间与旧基态来源未证/完整历史与其他域/实际 checkpoint commit/运行授权及 C12/main 未验收。不得把部分政策 NoChange 或局部测试升级成项目完成。
- [x] 最终读取 source/八旧守卫/向量/hash、branch/HEAD/index 及文档，确认最后审查/测试后无源码漂移；所有会话结束，只保留本机成果与证据。不暂存/提交/推送/合并/删除/清理，明确尚无跨电脑恢复检查点。

## 作者自检与执行交接

设计 §1–2→任务 2/3/5；§3→任务 1/3；§4→任务 1/3/4；§5→任务 3/4；
§6→任务 2；§7→任务 2/3/4；§8→任务 1/4；§9→任务 1/4/5；
§10→四个实现 RED/GREEN 循环、八旧正反例、16 条最终验证与审查；
§11→本轮文档及计划审阅门。五项 Review Focus 均有明确命名测试。

类型/函数/文件/状态名称一致；20 新文件+八旧守卫=28 格式化/增量冻结，
14 完整包+authority 限定+store 编译=16 验证。任务以所有权、对应关系、
完整组合和重检为独立可测试交付，不引入候选提交、parser、运行时消费者
或新的下层算法。每项 checkbox 都有可核对结果，没有提前勾选完成。

以上自检保留计划写作阶段的检查记录，不能当作执行结果。计划及设计
现已获批并按 executing-plans 的 Native 方式完成限定本机增量；实际失败、
通过、I1 修正和 M1 审计限制见执行记录。审批从未授权本轮 Git 集成、
生产接入或扩大验收，现有整体验收门仍未通过。
