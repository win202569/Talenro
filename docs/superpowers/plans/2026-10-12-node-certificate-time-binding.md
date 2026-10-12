# Node Certificate Conditional Time Binding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增离线 nodecerttime，以两次真实下层判定、同节点/时间快照关系和 CA→leaf 半开时间窗，产生不升级信任或运行权限的条件证据。

**Architecture:** 完整预算后克隆并保护整个 Input，顺序调用 nodeissuer.Evaluate 与 statetimebinding.Evaluate，保留所有嵌套责任。唯一 Now 来自已核对的时间摘要；ValidUntil 取 time/CA/leaf 三者最小到期时刻。共享 owner 管理包装生命周期，返回摘要的两个可变集合独立复制。

**Tech Stack:** 已缓存 Go 1.26.5、Windows PowerShell、现有两个下层包及值类型；无新依赖、wire、数据库、provider 或运行消费者。

**Spec:** [已批准的节点证书与条件时间证据绑定设计](../specs/2026-10-12-node-certificate-time-binding-design.md)。执行者须完整阅读规格和本计划。

状态：用户最新回复“批准”批准上述书面设计，允许编写本计划；规格中的
“待书面审阅/本轮”是上一阶段留痕，不修改已批准正文。**本计划尚待审阅及
执行方法确认。本轮只新增本文，不实现、不测试、不生成材料、不提交或推送。**

规划基线：既有 worktree `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`，
branch=`codex/c12-b01-task9-coordinator`，HEAD=`f5d731bef9d767d57ca32f0f4c3a0f532a4bc322`。
规格 SHA256=`595A4471C642727AE60D8C816BDEDAA65FD3B9F21EF812B503CDD12F91CE694B`；
规格在本机尚未跟踪，本计划也不是远端恢复检查点。

## Global Constraints

- 新源码仅在 `internal/nodecontrol/nodecerttime/`；旧源码只改规格 §9 的八份 boundary_test.go 消费例外及新增案例。旧产品/API/测试材料、模块、配置、SQL、既有完成记录不变。
- 公开 Input 恰为 `Certificate nodeissuer.Input`、`Time statetimebinding.Input`；Proof 为 uint8，恰 `None=0`、`NodeCertificateConditionalTimeBound=1`。Summary 恰为 Proof、Certificate、Time、ObservedAt、ValidUntil；Model 恰 `owned *owner`。
- Certificate 四份 raw 上限依次为 65,536/512/65,536/8,192 bytes，合计 139,776；TrustDomain 仅容量上限 253 bytes，语法由真实下层检查。
- 两张表各 Materials/Expected 128；各自 RootRotation 32；Target.Links/Main.RootTransition 32；RootSet raw 65,536、其他 raw/独立锚 262,144；每表 raw 16 MiB，Binding raw 32 MiB + 262,144。
- 每个 Expectation root keys/metadata keys/revoked/desired slots 为 5/64/4,096/8；两份 Main 各 root keys/metadata keys/revoked 为 5/64/4,096，bootstrap/agent deauthorized 各 128，desired slots/LKG/local faults/supervisor faults 为 8/8/64/16。
- Batch.Rotations 32、每份 raw 262,144、Batch raw 36 × 262,144；完整 Policy raw 32 MiB + 37 × 262,144。Old.Roots 33、每 root keys 5；Old.Keys/Old.RevokedIDs/History 各 4,096；Status/RevocationReason/IncidentKind 字节为 8/32/64。
- Time 两份信封各 262,144、合计 524,288；Time raw 43,778,048；完整 Input raw 43,917,824。按出现次数累计，`len > limit-total` 防溢出；全部容量检查先于五项必需 raw 非空检查及任何输入克隆。
- 必需 raw 仅 Certificate 四项与 Time.Policy.Binding.Proposed.Anchor.Raw；其他空值、未知 kind、域值/状态/Old 语义交由真实下层。预算通过不等于语义接受，不保证 heap 峰值或硬实时。
- 空集合规范为 nil；两张表仅顶层 Materials/Expected 按 Kind/Payload 排序，其他顺序保留。全部既有 time.Time 用 Round(0).UTC() 保留纳秒、真正零值；raw/签名/preimage/主状态 UnixSeconds 不改。
- 顺序恰为 context→全部预算→必需 raw→owned Input→真实 issuer→真实 timebinding→节点/时间快照→CA 窗→leaf 窗→最小到期→最终 context→发布。检测到的取消/超时优先于当前阶段错误。
- Now 唯一为 Summary.Time.Policy.Time.Now；两窗均 `NotBefore <= Now < NotAfter`，CA 先于 leaf。无 skew、整秒截断、5 分钟延长、leaf 全窗包含于 CA、签发模板检查或缓存许可。
- Now/Time.ValidUntil 非零、UTC 年 1–9999、Now 严格早于 Time.ValidUntil；证书端点只要求 UTC 年 1–9999、start<end，不以 IsZero 拒绝合法 year-1 起点。
- ObservedAt=Time.Policy.Time.Now；ValidUntil=min(Time.ValidUntil, Certificate.CA.NotAfter, Certificate.Leaf.Leaf.NotAfter)。六个固定错误见 Task 1；真实下层非 nil 错误原样传递，包括 PolicyFailure 类型、Code/Reason/错误链。
- 两次真实下层 owner 返回即登记 Clear，早于 context/错误处理；Time.Summary 直接捕获到已保护的 owner.summary.Time。失败/取消清理并返回 nil，正常 panic 清理后传播；产品禁止 recover。
- 包装拷贝共享 Clear 生命周期；独立 Evaluate 和每次 Summary 的可变集合独立。清零仅限可达自有值，不擦除调用者、已交出的摘要或别的 owner；不承诺字符串/不可达副本擦除、并发安全、致命 OOM 恢复。
- 原两张六项 Pending、原 Proof/Status、所有嵌套 Obligations 原样保留；无第三张责任表、Recheck、原始字节访问器、可信 bool、额外 Now、私钥、回调、provider、State/Commit/Ready/Authorize。
- 仍要求 Old 存在、非 FreshInstall、旧新 Desired 存在、完整 Root/Metadata 不变、无新 root transition、既有历史/floor/checkpoint 规则成立；由真实 statetimebinding 检查，不复制业务判定器。
- Identity.Leaf/issuer/serial/public-key/key-handle/epoch/lineage/行状态/部署 CA 来源与 trust bundle 不绑定；时间 transcript 不含证书摘要。可信时钟、TLS、nonce 消费、物理提交、生产兼容 UNKNOWN 和 Task10/B01/C12/main 阻塞不变。
- 只用 PowerShell、离线只读模块和现有 worktree，不安装依赖或修改持久 Go/Git 配置。计划批准不自动批准暂存/提交/推送、PR、main 合并、分支/worktree 清理或扩展验收范围。

## Review Focus

- 空锚与后半输入同时超限：Task 1 `TestCertTimePreflightPriority` 固定所有容量先于必需非空，不照搬旧 bindingBudget 的提前空锚返回。
- 输入共享底层切片、半途清理误伤调用者：Task 2 `TestCertTimeOwnedAliasIsolation` 与 Task 5 `TestCertTimeOwnershipBoundary` 分别证明可观察隔离及清理登记源码结构。
- 跨节点负例实际上坏在签名/下层时间：Task 3 `TestCertTimeCrossNodeControls` 先证明两份下层各自成立，再断言新绑定错误。
- 纳秒与合法 Go 零时刻的证书端点混淆：Task 4 `TestCertTimeWindowEdges`、`TestCertTimeYearOneEndpoints` 区分半开窗口、合法端点和缺失观测。
- 取消或 panic 出现在摘要捕获/移交附近：Task 4 `TestCertTimeCancellationAndPanic` 与 Task 5 所有权 mutation controls 固定最终 context、失败 nil、原输入不变和受保护捕获；不伪称观测不可返回 owner 的擦除。

---

## 1. 文件、接口与执行约束

新目录恰 35 个 Go 文件，17 产品文件、18 测试文件，无子目录或永久生成器。
下表路径均以 `internal/nodecontrol/nodecerttime/` 为前缀；任务中引用同名文件
均指此完整路径，不允许顺手重构其他包。

| 产品文件 | 唯一职责 | 主要测试文件 |
| --- | --- | --- |
| doc.go | 条件范围和非目标 | fixtures_test.go（材料、测试辅助） |
| input.go | Input、Proof、Summary | input_test.go |
| errors.go | 六个 sentinel | budget_test.go |
| budget.go | 完整容量、非空顺序、累计 | budget_test.go |
| owner.go | owner、通用复制、清理移交 | model_test.go |
| owned_certificate.go | Certificate 四份 raw 克隆/清零 | owned_certificate_test.go |
| owned_binding.go | 两份 Main、表、目标/锚/轮换 | owned_binding_test.go |
| owned_content.go | 十二域值和所有集合 | owned_content_test.go |
| owned_table.go | Materials/Expected 深复制与排序 | owned_table_test.go |
| owned_policy.go | Batch/Old/History、PolicyBinding | owned_policy_test.go |
| owned_time.go | 两份信封与四组观测 | owned_time_test.go |
| owned_summary.go | Time 摘要的 Links/Coverage 克隆/清零 | owned_summary_test.go |
| model.go | nil-safe Summary/Clear、共享包装 | model_test.go |
| lower.go | 阶段错误、恰两次真实判定/捕获 | lower_test.go |
| binding.go | 新层节点和时间快照等式 | binding_test.go |
| window.go | 必要时间、CA→leaf、最小到期 | window_test.go |
| evaluate.go | 全流程与最终发布 | evaluate_test.go |

其余三个测试文件恰为 `boundary_test.go`、`ownership_boundary_test.go`、
`consumer_boundary_test.go`，分别守卫能力/API/库存、清理流程、新包消费者。

八个旧修改定位（行号是规划基线，执行按 helper 名精确定位）：

| internal/nodecontrol 下的文件 | 只新增 nodecerttime 例外的 helper |
| --- | --- |
| nodeissuer/boundary_test.go:614 | issuerConsumerAllowed |
| statetimebinding/boundary_test.go:15 | timeBindingConsumerAllowed |
| statepolicybinding/boundary_test.go:15 | policyBindingConsumerAllowed |
| statebinding/boundary_test.go:95 | bindingConsumerAllowed |
| statecontract/boundary_test.go:75 | contractConsumerAllowed |
| statematerials/boundary_test.go:158 | materialsConsumerAllowed |
| statepolicy/boundary_test.go:14 | policyConsumerAllowed |
| stateverify/boundary_test.go:95 | isolatedImportAllowed 的 statepolicy 分支，stateverify 分支不开放 |

其他允许的文档变化仅本计划 checkbox、新增
`docs/roadmap/2026-10-12-node-certificate-time-binding-execution.md` 的实际执行记录，
及完成阶段 `docs/roadmap/current-status.md` 顶部本增量的事实性追加。
两者沿用现有进度记录位置，不覆盖已有历史；规格和既有历史正文不改。

公开签名恰为 `Evaluate(context.Context, Input) (*Model, error)`、
`(*Model).Summary() Summary`、`(*Model).Clear()`；其余接口全部私有。
每任务先写命名测试并取得真实 RED，再最小实现取得 GREEN。未定义接口造成的
编译失败可作第一轮 RED；fixture 构建失败/环境失败不能充当行为 RED。
本阶段默认“提交”检查点改为保留原始证据；Git 写入须另获明确批准。

## 2. 未来执行环境与证据（本轮不运行）

先核对 worktree/HEAD、八个旧文件和新增文件范围、已批准 spec hash；已有
修改保留，不重置。读取适用技能；Native 用 executing-plans，本计划最终
鲜活审查仍用 requesting-code-review/verification-before-completion。

执行时才建立专属忽略证据目录 `.superpowers/sdd/2026-10-12-node-certificate-time-binding/`，
先验证该目录归属及忽略规则。记录仅白名单环境，不打印全部环境或凭据：

```powershell
$env:GOOS='windows'
$env:GOARCH='amd64'
$env:GOTOOLCHAIN='local'
$env:GOPROXY='off'
$env:GOSUMDB='off'
$env:GOAUTH='off'
$env:GOVCS='all:off'
$env:GOFLAGS='-mod=readonly'
$env:GOPATH='C:/Users/Lenovo/go'
$env:GOMODCACHE='C:/Users/Lenovo/go/pkg/mod'
$env:GOCACHE='D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/.superpowers/sdd/2026-10-12-node-certificate-time-binding/go-cache'
$certTimeGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
$certTimeFmt='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/gofmt.exe'
```

检测并记录继承 GODEBUG/GOEXPERIMENT 是否影响试验，不静默改持久配置。
每次命令分别保存原命令、UTC 起止、完整 stdout/stderr、实际 exit code 和
源码前后 SHA256；RED/GREEN、生成/核对、格式、矩阵、取消与失败都保留。
不把中途工具超时/输出丢失记成测试失败或通过，不回填未运行结果；同一
真实测试 session 等待结束，持续交流，不自动扩大 timeout 或整仓范围。

下文命令从该 worktree 根运行。Go 执行前用上述环境；Task 1–4 的局部命令
与最终矩阵是不同 executions，不复用旧增量的 PASS 或 ledger。

### Task 1: 冻结公开合同、完整预复制预算与八处消费范围

**Files:** Create `doc.go`、`input.go`、`errors.go`、`budget.go`；Test `input_test.go`、`budget_test.go`；Modify 上表八份旧 boundary_test.go，仅指定 helper 的新例外和新增测试。

**Interfaces:**
- Consumes: `nodeissuer.Input/Summary`、`statetimebinding.Input/Summary`，上述完整预算，不调用判定器。
- Produces: 公开 Input/Proof/Summary；`preflight(context.Context, Input) error`、`certificateBudget(nodeissuer.Input) (int, error)`、`bindingBudget(statebinding.Input) (int, error)`、`contentBudget(statecontract.MainStateContentModel) bool`、`timeBudget(statetimebinding.Input) (int, error)`、`addRaw(total, size, limit int) (int, bool)`。预算 helper 只做容量，无空值/语义修补。

- [x] **Step 1: 写合同、容量和顺序失败测试。** `TestCertTimePublicContract` 固定字段/枚举；`TestCertTimeFixedErrors` 逐项比较下面文字且六个 sentinel 独立：

```go
// None == 0; NodeCertificateConditionalTimeBound == 1; Proof 为 uint8。
// ErrInvalidInput.Error() == "nodecerttime: invalid input"
// ErrLimitExceeded.Error() == "nodecerttime: limit exceeded"
// ErrInvalidEvidence.Error() == "nodecerttime: invalid evidence"
// ErrBindingMismatch.Error() == "nodecerttime: binding mismatch"
// ErrNotYetValid.Error() == "nodecerttime: certificate not yet valid"
// ErrExpired.Error() == "nodecerttime: certificate expired"
if preflight(nil, Input{}) != ErrInvalidInput { t.Fatal("nil context") }
// TestCertTimePreflightPriority：五个必需 raw 逐个置空 + 后半 History 4097。
if got := preflight(context.Background(), overLimitAndEmpty); got != ErrLimitExceeded { t.Fatal(got) }
// 去掉超限，五个空值逐个返回 ErrInvalidInput；预取消返回 context.Canceled。
if total, ok := addRaw(43917823, 1, 43917824); !ok || total != 43917824 { t.Fatal("total edge") }
if _, ok := addRaw(43917823, 2, 43917824); ok { t.Fatal("total overflow accepted") }
```

`TestCertTimeBudgetEdges` 对 Global Constraints 每项两位置、每个嵌套元素做
等长与 +1：等长容量 helper 成功，+1 恰 ErrLimitExceeded；等长 raw 用共享
backing 重复填充，断言仍按位置累计。`TestCertTimeBudgetTotals` 固定
139776、16*1024*1024、32*1024*1024+262144、36*262144、32*1024*1024+37*262144、
524288、43778048、43917824；对 addRaw 做 total==limit/size==0 成功、
size==limit-total 成功、size==limit-total+1 失败，及整数最大值附近不溢出。
完整输入 +1 可能先触发分项上限，不能声称隔离到不可独立超过的总预算。
未知 kind 使用其他 raw 容量分类，不在 preflight 中拒绝；空但非必需值仍通过容量。

- [x] **Step 2: 分别取得新包与旧消费案例的真实 RED。** 新包运行 `& $certTimeGo test ./internal/nodecontrol/nodecerttime -run '^TestCertTime(PublicContract|FixedErrors|Budget.*|PreflightPriority)$' -count=1 -p=1 -timeout=120s`；预期缺接口编译失败。八个旧文件各新增 `TestNodeCertTimeConsumerException`，逐包运行 `& $certTimeGo test ./internal/nodecontrol/<package> -run '^TestNodeCertTimeConsumerException$' -count=1 -p=1 -timeout=120s`；预期只在新 exact 消费正例失败，保留全部旧案例。

- [x] **Step 3: 在四个产品文件实现上述类型、六错误和容量接口。** `addRaw` 拒绝负 total/size/limit、total>limit 或 size>limit-total 后才相加；全部预算扫描结束才做五项非空。全量 scan 的 context 检测不能改掉已检测取消优先规则；不调用旧私有预算或照搬旧空锚提前退出。

- [x] **Step 4: 在八个旧 helper 加精确例外。** 同一测试名分别覆盖相对/当前根绝对路径、产品/同包测试允许；nested、相似前缀、else 后缀、另一个根、外部、wrong package、nodecerttime_test、sibling、目标 import/child 拒绝。目标 import 恰为该旧包，唯 stateverify 文件新目标恰为 statepolicy；stateverify 自身 import 明确仍拒绝。绝对/相对均先归一到当前仓库根，再比较直接目录和 package，不用宽泛前缀。

- [x] **Step 5: 重跑 Step 2 九条局部命令取得 GREEN。** 所有命名测试 PASS/exit 0；保存预复制预算等长不是下层接受的说明。检查旧八文件只增加新例外及案例，无其他旧源码变化。

### Task 2: 完整 owned Input、时间规范化、摘要隔离和共享生命周期

**Files:** Create `owner.go`、`owned_certificate.go`、`owned_binding.go`、`owned_content.go`、`owned_table.go`、`owned_policy.go`、`owned_time.go`、`owned_summary.go`、`model.go`；Test 对应七个 owned_*_test.go 和 `model_test.go`。

**Interfaces:**
- Consumes: Task 1 的 Input/Summary/preflight；类型复制不消费 parser/Evaluate。
- Produces: `type owner struct { input Input; summary Summary; valid bool }`、`type Model struct { owned *owner }`；`ownInput(context.Context, Input) (*owner, error)`、`(*owner).clear()`、`copySlice[T any]([]T) []T`、`copyRaw([]byte) []byte`、`normalizeTime(time.Time) time.Time`、公开 Summary/Clear。
- Produces: `ownCertificate(context.Context, nodeissuer.Input) (nodeissuer.Input, error)` / `clearCertificate(*nodeissuer.Input)`；`ownBinding(context.Context, statebinding.Input) (statebinding.Input, error)` / `clearBinding(*statebinding.Input)`；`ownContent(context.Context, statecontract.MainStateContentModel) (statecontract.MainStateContentModel, error)` / `clearContent(*statecontract.MainStateContentModel)`；`ownTable(context.Context, statematerials.Input) (statematerials.Input, error)` / `clearTable(*statematerials.Input)`。
- Produces: `ownPolicy(context.Context, statepolicy.ModelInput) (statepolicy.ModelInput, error)` / `clearPolicy(*statepolicy.ModelInput)`；`ownPolicyBinding(context.Context, statepolicybinding.Input) (statepolicybinding.Input, error)` / `clearPolicyBinding(*statepolicybinding.Input)`；`ownTimeBinding(context.Context, statetimebinding.Input) (statetimebinding.Input, error)` / `clearTimeBinding(*statetimebinding.Input)`；`ownTimeSummary(statetimebinding.Summary) statetimebinding.Summary` / `clearTimeSummary(*statetimebinding.Summary)`。

- [x] **Step 1: 写完整字段复制、别名及生命周期失败测试。** `TestCertTimeOwnedComplete` 用纯值输入填满十二域、Expected 嵌套集合、Target/Anchor、Batch/二维 rotations、Old/每 root keys/每 key、History、两信封和所有观测，逐字段断言复制后保留。`TestCertTimeOwnedAliasIsolation` 故意共享 raw/重叠 subslice/两个 Main 集合，修改 owned 一位置不改变原输入或另一个 owned 位置；两个独立 ownInput 不共享 backing。

```go
// TestCertTimeNormalizeTimes：同 instant 的 offset 表示与非零纳秒。
if got := normalizeTime(withOffset); got != withOffset.Round(0).UTC() { t.Fatal(got) }
if normalizeTime(time.Time{}) != (time.Time{}) { t.Fatal("missing time repaired") }
// TestCertTimeModelLifecycle：由本包测试直接构造私有有效 owner，不称真实 Evaluate。
m := &Model{owned: o}; copied := *m
first, second := m.Summary(), m.Summary()
// 修改 first.Time.Policy.Binding.Proposed.Target.Links/Coverage，second 与 o 均不变。
copied.Clear(); copied.Clear()
if !reflect.DeepEqual(m.Summary(), Summary{}) { t.Fatal("shared lifecycle") }
// retained owned raw/slice backing 全零；原输入、first、second 和另一个 owner 不变。
// nil/zero Model Summary 完整零值，Clear 不 panic。
```

各 `TestCertTimeOwnedCertificate/Binding/Content/Table/Policy/Time/Summary` 负责
该文件全部可变字段、empty→nil、原集合顺序；只有两表顶层按 Kind/P 排序。
时间字段清单为 Policy.Time.Now/Floor、Old.TimeFloor、Old.Keys[].Material.
NotBefore/NotAfter/RevokedAt、FirstRevokedAt/LastSnapshotUntil/LastCompositeDeadline、
History 两 deadline、外层 Time.Time.Time.Now/Floor；主状态整数时间不动。

- [x] **Step 2: 运行取得 RED。** `& $certTimeGo test ./internal/nodecontrol/nodecerttime -run '^TestCertTime(Owned.*|NormalizeTimes|ModelLifecycle)$' -count=1 -p=1 -timeout=120s`；预期缺 owned/Model 接口，或明确复制/规范化断言失败。

- [x] **Step 3: 在指定文件实现完整输入拥有与清理接口。** 每个 helper 先将借用的全部可变字段脱钩为 nil，再登记失败/panic 清理，再逐项克隆；二维外层先脱钩内层 raw，Old 新指针先脱钩 Roots/Keys/RevokedIDs，每 Root.Keys 也先脱钩。每次 context 回调前保护已复制部分；不清零仍借用的 input。`ownInput` 调用 preflight 并仅一次转移完整 owner；不在类型复制路径调用额外判定器。

- [x] **Step 4: 实现受保护的摘要复制与包装生命周期。** `ownTimeSummary` 只深复制 Links/Coverage，并在第一次分配前登记清理；其余全部值原样复制。Model.Summary 仅对 valid owner 返回完整深复制 Summary；Model.Clear 经共享 owner.clear 清零 input/summary/valid，不仅切断当前包装指针。只有已拥有的 Old 可调用 Clear，不清理调用者输入。

- [x] **Step 5: 重跑 Step 2 取得 GREEN 并运行 Task 1 的新包测试。** PASS/exit 0；记录可观察 backing 清零和原输入隔离。分配失败/不可返回 owner 的内存擦除不作运行证明，后续 Task 5 固定源码登记次序。

### Task 3: 固定真实签名材料、两次 lower 判定及同快照绑定

**Files:** Create `lower.go`、`binding.go`；Test `fixtures_test.go`、`lower_test.go`、`binding_test.go`。一次性生成/独立核对源码仅放专属忽略证据目录，不进新包 inventory；旧 fixture/testdata 不改。

**Interfaces:**
- Consumes: Task 2 `owner`、owned/clear 接口；`nodeissuer.Evaluate(context.Context, nodeissuer.Input) (*nodeissuer.Material, error)`、`statetimebinding.Evaluate(context.Context, statetimebinding.Input) (*statetimebinding.Model, error)`，以及它们的 Summary/Clear。
- Produces: `stageError(context.Context, error) error`、`evaluateLower(context.Context, *owner) error`、`checkLowerProofs(Summary) error`、`bindSnapshots(context.Context, Input, Summary) error`。
- Produces test-only: `certTimeFixture(testing.TB, string) Input` 从固定完整 Input 常量解码独立值；`assertCertTimeLowerControls(testing.TB, Input) Summary` 实际调用两下层、立即 defer Clear 后返回独立摘要。`certTimeFixture` 的名字列表与逐份 raw/摘要 hash manifest 一并冻结，不随机签发。

- [x] **Step 1: 生成并独立核对冻结材料，建立真实 lower 正控制。** 参考 statetimebinding/fixtures_test.go 的 compositionFixture/declarationFor/policyBindingFixture/attachTime 签名与引用关系，但在专属生成器构建自洽新材料，不调用其他包私有测试函数。基础 node=`12345678-1234-4234-8234-123456789abc`、TrustDomain=`node.test`、authority epoch=19、Now=`2020-01-01T01:00:00Z`；基础 CA=`2019-01-01T00:00:00Z` 至 `2030-01-01T00:00:00Z`，leaf=`2020-01-01T00:55:00Z` 至 `2020-01-02T00:55:00Z`，time attestation=`01:00:00Z` 至 `01:05:00Z`。

基础选择既有一段历史 root rotation（Old.Roots 两项、Target.Links 一项），
但 Main.RootTransition 为空、完整旧新 Root/Metadata 相同；另冻结 0/32 段
既有历史正控制，确保非空 Links 的真实摘要隔离也被动态测试。
使用固定测试 P-256 标量 1/2（CA/leaf），Ed25519 root seed 为 32 个 0x01、
metadata/time signer seed 为 32 个 0x02；这些只用于生成测试材料。签名可在
一次性生成阶段产生后冻结，不把生产钥匙/生成器或随机重签放进普通测试。
生成器自行计算并冻结 exact DER、SPKI、TBS 和 JSON/JCS transcript 的摘要，
独立核对者重新检查字节、签名、references/Expected/Old/Main 坐标及所有 raw
hash；共享同一摘要函数输出不称独立核对。普通测试仅解码冻结材料并做
明确的输入 mutation，不重新签名。test-only 可用 JSON/crypto/hash/reflect；
产品及新测试都不直接 import 规格禁止新增例外的七包。

冻结病例包括基础、另一节点（全部时间/desired/声明自洽）、CA/leaf 端点
前后 1ns 六点各自可成立的时间证据、两种双窗故障、三者最早和并列、
leaf 跨 CA 全窗但 Now 同时有效、CA year-1 起点与 leaf year-1 起点的真实
issuer 正控制、坏签名/nonce/metadata、rollback/缺失观测/PolicyFailure。
每个 joint window 病例的 attestation 在三点 Now 下都有效；所需 metadata/key/
desired/window/floor 随病例自洽冻结，不拿过期 time attestation 冒充新窗错误。
year-1 的 issuer-only 控制不冒称完整时间组合已接受。`TestCertTimeFixtureControls`
逐例证明标明应通过的真实 lower，并核对 manifest；fixture 不成立先修 fixture。

- [x] **Step 2: 写 lower/snapshot 失败测试并运行取得 RED。** `TestCertTimeLowerErrorIdentity` 对同一坏输入分别取得真实下层错误再调用 evaluateLower，断言 sentinel `got==want`；PolicyFailure 用 errors.As 核对同型 Code/Reason/unwrap 链，不要求两次独立调用分配的 error 指针相等。`TestCertTimeStageError` 固定 nil context、Canceled/DeadlineExceeded 优先与正常原错误透传。`TestCertTimeCrossNodeControls` 先取得真实两 lower 成功摘要，然后：

```go
if got := bindSnapshots(context.Background(), crossNodeInput, checkedSummary); got != ErrBindingMismatch { t.Fatal(got) }
// TestCertTimeSnapshotRelations：逐改证书摘要 node、time node、输入三组 node/Main 两 Scope。
// 逐改 ProviderID/ObservationID/Now(1ns)/ObservedFloor，对应关系恰 ErrBindingMismatch。
// TestCertTimeLowerProofs：错误 issuer.Proof 或 time.Status -> ErrInvalidEvidence。
// TestCertTimeTimeTupleLowerControls：输入 tuple 不一致时真实 timebinding 自身拒绝。
```

Run: `& $certTimeGo test ./internal/nodecontrol/nodecerttime -run '^TestCertTime(FixtureControls|Lower.*|StageError|CrossNodeControls|SnapshotRelations|TimeTupleLowerControls)$' -count=1 -p=1 -timeout=120s`；预期尚缺新接口或新绑定断言失败，FixtureControls 本身必须通过才归类行为 RED。

- [x] **Step 3: 实现 lower.go 的四个接口中前三个。** `evaluateLower` 输入只能是已保护 owner：真实 issuer 调用后立即 defer Clear，再 context/error、非 nil 检查、写 owner.summary.Certificate；真实 timebinding 调用后同样立即 defer Clear，再 context/error、非 nil 检查，**直接** `o.summary.Time = timeModel.Summary()`，然后检查恰有旧 Proof/Status。禁止临时未保护可变 Time 摘要、替换函数/别名/回调和其他下层入口。此步骤不设置 o.valid 或新 Proof。

- [x] **Step 4: 在 binding.go 实现 bindSnapshots。** 等式逐项按规格 §6：Certificate.Leaf.Leaf.NodeID=owned Certificate.Leaf.NodeID=Time.Time.NodeID；time node 对应 Metadata/Request/Checkpoint.NodeID 和 Main 两 Scope 的 canonical UUID.String。Time.Policy.Time 的 ProviderID/ObservationID/Now/ObservedFloor 对应 Time.Time.Time 与 Policy.Policy.Time 的 tuple；time.Time 均已规范，用值相等保留 ns。其他字段只沿用真实下层，不新增域=metadata、provider=CA digest、epoch 签入证书等判断。

- [x] **Step 5: 重跑 Step 2 取得 GREEN。** 全部 PASS/exit 0；真实 lower 错误与新 snapshot helper 人为摘要 mutation 控制分别标明，不把 caller-built Summary 视为公开输入或产品信任路径。

### Task 4: 半开窗、最小期限、取消/正常 panic 和公共 Evaluate

**Files:** Create `window.go`、`evaluate.go`；Test `window_test.go`、`evaluate_test.go`。

**Interfaces:**
- Consumes: Task 1–3 的 preflight/ownInput/evaluateLower/bindSnapshots/stageError 和真实 fixture 控制。
- Produces: `certificateWindow(now, start, end time.Time) error`、`checkWindows(Summary) (time.Time, error)`；公开 `Evaluate(context.Context, Input) (*Model, error)`。checkWindows 返回唯一最小 ValidUntil，不读取时钟或独立 Now。

- [x] **Step 1: 写窗口与必要时间失败测试。** `TestCertTimeWindowEdges` 对 CA/leaf 的 start-1ns/start/start+1ns、end-1ns/end/end+1ns 分别先跑真实 lower 正控制，期望如下；unit helper 与完整 Evaluate 两类分开标记：

```go
// 每个端点使用 fixtures_test.go 冻结的相应自洽 Input，不事后改签名 raw。
// start-1ns -> ErrNotYetValid; start/start+1ns -> nil。
// end-1ns -> nil; end/end+1ns -> ErrExpired。
if got := certificateWindow(start.Add(-time.Nanosecond), start, end); got != ErrNotYetValid { t.Fatal(got) }
if got := certificateWindow(start, start, end); got != nil { t.Fatal(got) }
if got := certificateWindow(end.Add(-time.Nanosecond), start, end); got != nil { t.Fatal(got) }
if got := certificateWindow(end, start, end); got != ErrExpired { t.Fatal(got) }
// TestCertTimeYearOneEndpoints：合法 start=time.Time{}、end>start，不误判缺失端点。
if got := certificateWindow(yearOneNonzeroNow, time.Time{}, yearOneEnd); got != nil { t.Fatal(got) }
// TestCertTimeNecessaryTimes：Now/Time.ValidUntil zero，UTC year 0/10000，
// Now>=Time.ValidUntil，证书 start>=end/超年界 -> ErrInvalidEvidence。
// TestCertTimeDeadlineMinimum：time/CA/leaf 最早和每种并列，got == exact min。
// CA expired + leaf future -> ErrExpired；CA future + leaf expired -> ErrNotYetValid。
```

`TestCertTimeNonContainingLeafWindow` 真实两 lower 均成功且 Now 同时有效，
leaf 全窗不包含于 CA 仍接受，期限取较早者。`TestCertTimeFreshObservation`
改变 Now 1ns 后重新完整 Evaluate，ObservedAt 保留 ns；原 owner 摘要不被新
调用改写，不声称旧结果为新观测提供认证。

- [x] **Step 2: 写完整流程与责任保留失败测试并运行 RED。** `TestCertTimeEvaluateGolden` 比较新 Proof、ObservedAt、最小期限及两个完整真实下层 Summary；`TestCertTimeObligationsRetained` 用 reflect.DeepEqual 比较所有嵌套责任，原两张 Pending 各六项保留。`TestCertTimeUnboundIdentity` 仅改变下层不认证的 Identity 引用、CA 自身签名/未认证来源，先证明 lower 控制再检查相同条件结果与未完成责任，不伪造 full-main-state 授权。

`TestCertTimeEvaluatePriority` 固定容量>空值>issuer>time>新绑定>CA>leaf；
`TestCertTimeIndependentLifecycle` 两次真实 Evaluate、两次 Summary、包装 Clear
不互相误伤。`TestCertTimeCancellationAndPanic` 用 test-only counting context
先量测真实 successful Err 调用点，再逐点检测取消/DeadlineExceeded/正常
panic，断言取消时 nil Model+原 context 错误、panic 原值传播、原输入不变。
test-only defer/recover 可用，产品不得 recover；保存能取得的 owned backing
控制，公共 Evaluate 的不可返回 owner 清理只以 Task 5 AST 控制支持。

Run: `& $certTimeGo test ./internal/nodecontrol/nodecerttime -run '^TestCertTime(Window.*|YearOneEndpoints|NecessaryTimes|DeadlineMinimum|NonContainingLeafWindow|FreshObservation|Evaluate.*|ObligationsRetained|UnboundIdentity|IndependentLifecycle|CancellationAndPanic)$' -count=1 -p=1 -timeout=120s`；预期缺接口或明确窗口/移交断言失败，不把 fixture 失败作 RED。

- [x] **Step 3: 实现 window.go 两个接口。** checkWindows 验证必要 observation 时间和证书 start/end 年范围及严格升序；观察零值无效，但证书起点零时刻可有效。按 CA 后 leaf 调用 certificateWindow，仅用 Before/Year/IsZero；返回 min 三个 exact time.Time，不通过 Unix/Seconds/Duration 舍入、Add 延长或叶全窗包含判断。

- [x] **Step 4: 实现 evaluate.go 的完整 Evaluate。** `ownInput` 返回后在下一 context 回调前登记 owner.clear；经 stageError 处理每阶段，运行两个 lower、绑定、窗口，写 ObservedAt/ValidUntil，最后 context 成功才设置新 Proof、valid 并转移包装。任何失败返回 nil；正常 panic unwind 不吞错，外部输入始终未改。

- [x] **Step 5: 重跑 Step 2 取得 GREEN，并执行新包完整测试。** `& $certTimeGo test ./internal/nodecontrol/nodecerttime -count=1 -p=1 -timeout=120s`；所有测试 PASS/exit 0。标明计数 context 测的是可检测取消，不是硬实时取消/平台原子快照保证。

### Task 5: 精确能力/所有权/消费者守卫、23 命令矩阵与 fresh 审查

**Files:** Create/Test `boundary_test.go`、`ownership_boundary_test.go`、`consumer_boundary_test.go`；只在发现本增量问题时修指定新文件/八旧测试，遵循相应修复 TDD；完成后如实更新本计划与本增量执行记录。

**Interfaces:**
- Consumes: 完整 35 文件库存、公开 API 和前四任务私有接口/实际源码；不改变产品接口。
- Produces test-only: `certTimeProductAllowed([]byte) bool`、`certTimeCapabilityAllowed(map[string][]byte) bool`、`certTimeOwnershipAllowed(map[string][]byte) bool`、`certTimeConsumerAllowed(path, pkg, imported string) bool`；测试辅助 `certTimeSources(testing.TB) map[string][]byte`、`certTimeMutation(t testing.TB, sources map[string][]byte, file, old, replacement string) map[string][]byte`。mutation 必须改变实际源码，否则控制本身失败。

- [x] **Step 1: 写精确 guard 的正反失败测试并运行 RED。** `TestCertTimeProductBoundary/CapabilityBoundary/OwnershipBoundary/Consumers` 检查真实新包；正例含安全命名别名、Round(0).UTC()、同目录同包测试 import。负例逐项覆盖禁止能力/额外导出或全局 hook/Proof 改号/新文件子目录/对旧包的违规入口、时间方法参数绕过、方法值/判定器函数值逃逸、别名遮蔽、dot/blank/重复 import、goroutine/recover；消费者 subpath、外部测试包及其他根也拒绝。

所有权 mutation 逐项删除/后移完整输入 defer、部分 clone cleanup、lower
立即 Clear、可变 Time Summary 直接受保护捕获、Summary 两集合复制、最终
context/valid/transfer 顺序及 backing 清零，均须被拒绝。良性语法/格式/命名
别名仍接受；不把字符串扫描或单个 guard 声称为形式安全证明。
Run: `& $certTimeGo test ./internal/nodecontrol/nodecerttime -run '^TestCertTime(ProductBoundary|CapabilityBoundary|OwnershipBoundary|Consumers)$' -count=1 -p=1 -timeout=120s`；预期 guard 未定义或指定 mutation 未拒绝。

- [x] **Step 2: 实现 AST/源码库存守卫并重跑 GREEN。** 产品 import/member 精确白名单为下表；测试源自身可用 AST/文件读取工具，仅作本机测试。最终守卫同时核对公开字段/签名、owner 三字段、六 sentinel 文字、无其他全局可变值/导出能力和 35 文件库存，不默认允许整包。

| import（本地前缀 talenro.local/platform/internal/nodecontrol/） | 产品允许的 package 成员 |
| --- | --- |
| context / errors / bytes / sort / time | Context / New / Clone,Compare / Slice / Time |
| nodeissuer | Input, Summary, Evaluate, NodeLeafDirectIssuerSignatureBound |
| statetimebinding | Input, Summary, Evaluate, ConditionalTimeEvidenceBound |
| statepolicybinding | Input |
| statebinding | Input |
| statecontract | MainStateContentModel |
| statematerials | Input, RootSet, RootRotation |
| statepolicy | ModelInput, StateModel |

不需要 contracts 的直接 import；若实现需要白名单以外的成员，先书面修订，
不放宽整包。明确拒绝 stateformat/statecomposition/stateverify/statetime/nodeca/
nodeleaf/nodeleafkeys 的直接 import（产品及新测试）；泛型推导其嵌套值。
方法只允许 context.Err、time.Time 的 Round(0)/UTC/IsZero/Year/Before、
用于节点 canonical 对比的 UUID.String、两个真实 lower 的 Summary/Clear、
已拥有 Old.Clear 及本包声明的私有生命周期 helper；不开放其他 method call。
两次 Evaluate 必须直接调用在 lower.go，UUID.String 仅 binding.go 节点等式，
时间规范化仅 normalizeTime；AST guard 对方法值、别名赋值及隐藏调用做反控制。
新包消费者唯一例外为 exact 当前根直接目录/package=nodecerttime 的 _test.go，
其产品文件、外部测试、子目录/子路径/sibling/其他根一律拒绝。

- [x] **Step 3: 格式检查并执行 23 个独立最终命令。** `& $certTimeFmt -l` 加本增量 35 个新 Go 路径和八个旧测试路径，期望空输出/exit 0；需要格式化只限这些文件，不整仓重写。下列每行单独计数/留证，不使用 ./...、不把 store 记成已跑测试：

```powershell
& $certTimeGo test ./internal/nodecontrol/nodecerttime -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/nodeissuer -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/nodeca -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/nodeleaf -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/nodeleafkeys -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/contracts -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/serving -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/statebinding -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/statecomposition -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/statecontract -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/stateformat -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/statematerials -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/statepolicy -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/statepolicybinding -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/staterecovery -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/statetime -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/statetimebinding -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/stateverify -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/readiness -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/store -run '^$' -count=1 -p=1 -timeout=120s
& $certTimeGo test ./internal/nodecontrol/nodeissuer -run '^$' -fuzz '^FuzzIssuerEnvelope$' -fuzztime=5s -parallel=1 -p=1 -timeout=30s
& $certTimeGo test ./internal/nodecontrol/nodeissuer -run '^$' -fuzz '^FuzzIssuerPair$' -fuzztime=5s -parallel=1 -p=1 -timeout=30s
```

期望 23 个实际 exit code 全为 0；19 完整包与 authority 指定测试 PASS，
store 仅 compile/no tests to run，两 fuzz 完成原配置。记录 fuzz 基线/实际
时长和启动等待，不把 5s 说成整个进程只运行 5s。任何失败先按
systematic-debugging 查因；不能跳过变成通过、擅改 timeout 或扩展旧产品。

- [x] **Step 4: 冻结增量并取得一次 fresh 只读质量/安全审查。** Native 执行完后按照该方法要求，派一位新上下文最强可用 reviewer，审本增量相对 f5d731be 的实际修改与新增未跟踪文件、已批准规格/计划、原始证据，不审旧提交为新贡献；这是执行方法的明确审查步骤，不是本轮已启动的代理。审查者不改文件/Git、不替作者补证、不跑扩展验收；发现问题修复并重测影响项及最终矩阵，报告未解决项，不能预填“零问题”。

- [x] **Step 5: 仅在证据与审查完成后记录实际交付并请求下一授权。** 执行文档记逐任务 RED/GREEN、fixture 核对、格式、23 矩阵、源码前后 hash、最终审查真实结论和所有未认证责任；计划 checkbox 只标实际完成步骤。任何取消/失败/验收阻塞保留，不把有限测试/短 fuzz/AST/单审查扩称整仓、race/Linux、DB/Docker、TLS/provider、硬件/reboot/PITR 或 C12 权威验收。没有后续 Git 批准时停在本机修改与待保存状态，明确不能换电脑恢复新增进度。

## 3. 作者自检与执行交接

作者自检映射：规格 §1–3/8 为 Task 1 公开合同与 Task 4 责任；§4 为 Task 1
每项预算/总量/空值顺序；§5 为 Task 2 深复制与 Task 5 cleanup mutations；
§6 为 Task 3 真实调用和关系、Task 4 发布；§7 为 Task 4 半开窗/纳秒/year-1/
最小期限；§9 为 Task 1 八处旧例外、Task 5 新守卫；§10 为 Task 3–5 的真实
正控制、有限矩阵与 fresh 审查；§11 的批准门槛见本文状态与交接。
五项 Review Focus 都有命名测试。接口/文件/API 数量与白名单逐项核对，
未把测试辅助或 reviewer 执行预写成完成事实；文档自检不是独立源码审查。

建议继续此前 Native 方式：五个任务依赖同一 owned 输入及 lower 摘要接口，
由当前会话顺序实施便于保持合同一致，最后一次 fresh reviewer 审整个增量。
需要用户审阅本计划并确认执行方法之后才实施；也可选择逐任务独立实现/
审查的 Subagent-driven。未批准前不创建材料/证据目录、不运行 Go 测试、
不改八份旧测试、不启动实现或审查代理，更不进行任何 Git 写入。
