# Main State Selected Leaf Projection Binding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增离线 nodeleafbinding，将 Proposed 主状态的节点、选中 leaf DER、公钥摘要和证书窗口精确绑定到一次真实 nodecerttime 判定，不产生完整身份或运行授权。

**Architecture:** 全量预复制预算后取得自有完整 Input，调用真实 nodecerttime.Evaluate 恰一次并直接捕获受保护的完整摘要，然后核对 Proposed 投影。沿用下层 ObservedAt/ValidUntil 和所有责任；共享 owner 生命周期与独立摘要深复制保证调用者隔离。

**Tech Stack:** 缓存 Go 1.26.5、Windows PowerShell、现有 nodecerttime 和限定值类型；无新依赖、wire、平台适配器或运行消费者。

**Spec:** [已批准的主状态选中叶证书投影设计](../specs/2026-10-12-main-state-selected-leaf-binding-design.md)。执行者必须完整阅读规格及本计划。

状态：用户回复“批准”批准上述设计，授权编写本计划。规格的“待书面审阅”
保留起草历史，不修改批准正文。**本计划待审阅；本轮只新增本文，不实施、
运行测试、生成材料/实证、暂存、提交、推送或启动代理。**
沿用此前 Native 方法，计划获批后使用 executing-plans 逐任务实现，末尾一次
fresh 只读质量/安全 reviewer；不派并行实现者，不把设计批准当计划批准。

基线：D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator，
branch=codex/c12-b01-task9-coordinator，
HEAD=a820f550a4e1662143676dc3f58d4f385c14f5a1。
规格 SHA256=3FCC3A2977B55C077660EA0F6D17AB2E2D42EA52A0ABAC45EC1BCC35EE621267。
规格与本文均本机未跟踪，不是远端恢复检查点。

## Global Constraints

- 新源码仅 internal/nodecontrol/nodeleafbinding；旧源码只改§1列出的九份测试例外。产品/API/旧 fixture/AST schema、SQL/module/lock/配置及历史设计不改。
- 局部 profile 恰 main-selected-leaf.der-spki-window.v1，无 wire 或运行开关；PublicKey 显式等于 exact SPKI DER 摘要，只是本模型合同，生产兼容 UNKNOWN。
- Input 为 nodecerttime.Input 类型别名；Proof uint8，None=0、SelectedLeafProjectionConditionalTimeBound=1；Projection、Summary、Model 恰为规格§4公开形状。
- Proposed.Scope.NodeID 对应 leaf.NodeID.String()；Identity.Header=DomainIdentity/Present；Leaf.Kind=CertificateDER、Payload=LeafDERDigest、Envelope=零；PublicKey=SPKIDERDigest。
- Window 两整数范围 [-62135596800,253402300799]、Issued<Until；先检查再 time.Unix(n,0).UTC()，用 Equal 对应 NotBefore/NotAfter，不以 Unix 截断证书时间。
- Certificate 四 raw 上限 65,536/512/65,536/8,192，合计139,776；TrustDomain253 bytes。两张表各 Materials/Expected128，各自 RootRotation32；Target.Links/Main.RootTransition32。
- RootSet raw65,536，其他 raw及独立 Anchor262,144；每表 raw16MiB，完整 Binding raw32MiB+262,144；Expectation root keys/metadata keys/revoked/desired slots5/64/4,096/8。
- 两份 Main 各 root keys/metadata keys/revoked5/64/4,096；bootstrap/agent deauthorized各128；desired slots/LKG/local faults/supervisor faults8/8/64/16。
- Batch.Rotations32、每 raw262,144、Batch raw36×262,144；完整 Policy raw32MiB+37×262,144。Old.Roots33、每 root public keys5；Old.Keys/Old.RevokedIDs/History各4,096；Status/RevocationReason/IncidentKind bytes8/32/64。
- Time 两信封各262,144、合计524,288；完整 Time raw43,778,048；完整 Input raw43,917,824。各位置按出现次数计数，用 size>limit-total 等安全比较，不按 backing pointer 去重。
- 全部容量检查先于五必需 raw 非空和复制；五 raw 为四份 Certificate 及 Time.Policy.Binding.Proposed.Anchor.Raw。其他语义/缺失/未知 kind 保持真实下层决定。
- 完整十二域、两表、Batch/Old/History、观测和raw深复制；仅空集合为nil、两表顶层按Kind/P排序、time.Round(0).UTC()，其他有序字段不改；保留纳秒/真正零值。
- 下层恰一次 nodecerttime.Evaluate，返回 owner 即登记 Clear，先于 context/error；完整 Summary 直接捕获到已保护 owner.summary.Evidence，Links/Coverage 每次返回另复制。
- 下层错误原样返回，包含 PolicyFailure和错误链；当前检测到取消/超时优先；下层失败先于本层投影错配；成功前最终检查context，失败nil、panic清理后传播。
- ObservedAt/ValidUntil非零、UTC年1–9999、ObservedAt<ValidUntil；ObservedAt==Evidence.Time.Policy.Time.Now；不要求 ValidUntil==Evidence.Time.ValidUntil，证书可更早到期。
- 输出 ObservedAt/ValidUntil精确沿用Evidence，全部Proof/Pending/Obligations不删。五新责任与四固定错误见Task1；不新增Recheck/State/Commit/Ready/Authorize、parser/摘要重算、可信bool或provider。
- 浅拷贝共享owner，任一Clear使所有副本失效；nil/零值/清理后Summary零；禁止并发访问，清理不损坏caller/已返回摘要，不承诺runtime隐藏副本擦除或硬实时。
- 只绑定Proposed四类投影，不认证Previous证书、issuer/serial/key-handle/epoch/lineage/history/row；不猜issuer/serial preimage，不证明私钥/平台来源/部署CA；相等SKI数值不可证明来源。
- 继承Old存在、非FreshInstall、旧新Desired存在、Root/Metadata/floor不变等完整下层限制；不接入真实DB/Docker、TLS/provider、硬件/reboot/PITR、整仓/race/Linux或main。
- Task10/B01/C12/main阻塞与生产兼容UNKNOWN保持；本计划批准不授权Git写入、推送、PR、合并或清理。每任务的提交步骤替换为本地证据检查点，待独立Git授权。

## Review Focus

- 同leaf两侧声明只改一侧，旧statecontract先拒绝而遮蔽新层：Task3/4 TestLeafBindingProjectionControls先证完整nodecerttime成功，再断言本层错配。
- leaf SKI恰等于SPKI摘要，误声称能拒绝相同bytes来源：Task4 TestLeafBindingEqualSKIValue允许相等值、责任不消失。
- year-1窗口与Go零时刻混淆，或把纳秒Now舍入到秒：Task4 TestLeafBindingWindowEdges/YearOneWindow/TemporalInheritance区分合法端点和缺失观测。
- 共用底层slice、半途失败清理caller或摘要捕获漏保护：Task2 TestLeafBindingOwnedAliasIsolation，Task5所有权mutation与Task4取消/panic逐项覆盖。
- 短命lower owner返回后先调用context，或最终检查前设置valid：Task3 TestLeafBindingLowerOwnership、Task5 TestLeafBindingOwnershipBoundary拒绝重排，Task4新鲜观测重评估验证不复用摘要。

---

## 1. 文件库存、接口和执行环境

新目录恰33个Go文件：16产品、17测试，无子目录或永久生成器。下表路径
均以 internal/nodecontrol/nodeleafbinding/ 为前缀，任务引用同名文件即该路径。

| 产品文件 | 唯一职责 | 测试 |
| --- | --- | --- |
| doc.go/input.go/errors.go | 局部合同、公开类型、四错误 | input_test.go |
| budget.go | 全输入长度/累计预算/必需非空 | budget_test.go |
| owner.go | owner三字段、预检、复制移交 | model_test.go |
| owned_certificate.go | 四证书raw复制/清理 | owned_certificate_test.go |
| owned_binding.go | 完整Main/两表/目标/锚/轮换 | owned_binding_test.go |
| owned_content.go | 十二域的值与集合 | owned_content_test.go |
| owned_table.go | 两表及Expectation完整复制/规范化 | owned_table_test.go |
| owned_policy.go | Batch/Old/History/PolicyBinding | owned_policy_test.go |
| owned_time.go | 两信封与全部观测 | owned_time_test.go |
| owned_summary.go | 完整Evidence及两可变集合 | owned_summary_test.go |
| model.go | Summary/Clear共享生命周期 | model_test.go |
| lower.go | 一次真实判定、阶段错误、摘要保护 | lower_test.go |
| projection.go | 精确投影及秒窗口对应 | projection_test.go |
| evaluate.go | 串联/最后context/发布 | evaluate_test.go |

其余测试恰 fixtures_test.go、boundary_test.go、ownership_boundary_test.go、
consumer_boundary_test.go。测试辅助无产品hook。33库存包含input/budget和
七owned测试、model、fixtures、lower、projection、evaluate、三个guard。

旧测试仅以下九处helper增加exact新包消费例外和正反例；定位按名字，不
假定历史行号稳定，原消费者规则/原反例保留：

| internal/nodecontrol 下的文件 | helper |
| --- | --- |
| nodecerttime/consumer_boundary_test.go | certTimeConsumerAllowed |
| nodeissuer/boundary_test.go | issuerConsumerAllowed |
| statetimebinding/boundary_test.go | timeBindingConsumerAllowed |
| statepolicybinding/boundary_test.go | policyBindingConsumerAllowed |
| statebinding/boundary_test.go | bindingConsumerAllowed |
| statecontract/boundary_test.go | contractConsumerAllowed |
| statematerials/boundary_test.go | materialsConsumerAllowed |
| statepolicy/boundary_test.go | policyConsumerAllowed |
| stateverify/boundary_test.go | isolatedImportAllowed（仅statepolicy依赖） |

首次执行前确认AGENTS规则、分支/HEAD、Git范围及批准规格hash；若基线变更
先只读辨识，不能覆盖用户修改。建立全nodecontrol Go/SQL、go.mod/go.sum及
上述旧测试的前后hash清单；未准许旧文件必须保持。创建本增量专用忽略实证
目录 .superpowers/sdd/2026-10-12-main-state-selected-leaf-binding，不能覆盖旧轮。
允许新增完成后的 docs/roadmap/2026-10-12-main-state-selected-leaf-binding-execution.md；
仅最终真实结果后追加current-status，不回写旧验收或旧review结论。

工具仅进程环境；不修改Go/Git持久配置。每命令记录可执行路径、参数、
开始/结束UTC、stdout/stderr、真实exit、前后源码hash、失败原因；同RunId
拒绝覆写。缓存/生成器/原始实证放本增量忽略目录，不写进产品或旧归档。

```powershell
$leafBindingGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
$leafBindingFmt='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/gofmt.exe'
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
$env:GOCACHE='D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/.superpowers/sdd/2026-10-12-main-state-selected-leaf-binding/go-cache'
```

## 2. 顺序任务与RED/GREEN

### Task 1: 公开合同、完整预算和九处消费例外

**Files:** Create doc.go/input.go/errors.go/budget.go、owner.go（只声明owner三字段）、input_test.go/budget_test.go；Modify仅§1九旧测试。

**Interfaces:** Consumes完整nodecerttime.Input及既有字段；produces
`preflight(context.Context, Input) error`、`addRaw(total,size,limit int)(int,bool)`、
`certificateBudget(nodeissuer.Input)(int,error)`、`contentBudget(statecontract.MainStateContentModel) bool`、
`bindingBudget(statebinding.Input)(int,error)`、`timeBudget(statetimebinding.Input)(int,error)`。
预算helpers与已批准基线一一对应，只换本层sentinel，不调用旧私有helper。

- [x] **Step 1: 写失败测试。** TestLeafBindingAPI固定Input alias与全部公开字段/Proof编号；TestLeafBindingBudget逐项测试Global Constraints容量equal/+1，重复raw按位置累计；TestLeafBindingPreflightPriority同时空锚/后段超限返回ErrLimitExceeded、nil ctx返回ErrInvalidInput、预取消优先。九旧TestNodeLeafBindingConsumerException覆盖绝对/相对exact路径及nested/sibling/其他根/错误包/external-test/subpath拒绝。

断言范式（limit案例由表驱动独立装配，不误称下层语义成功）：
```go
if e:=preflight(context.Background(), overLimitWithEmptyAnchor); e!=ErrLimitExceeded { t.Fatal(e) }
if _,ok:=addRaw(43917824,1,43917824); ok { t.Fatal("overflow budget accepted") }
```

- [x] **Step 2: 运行RED。** `& $leafBindingGo test ./internal/nodecontrol/nodeleafbinding -run '^TestLeafBinding(API|Budget|PreflightPriority)$' -count=1 -p=1 -timeout=120s`；预期未定义API/预算helper而失败。再对九旧包分别用`-run '^TestNodeLeafBindingConsumerException$'`，新例外尚未准许应失败；不能把旧测试基线成功当RED。
- [x] **Step 3: 实现合同和预算。** 公共类型按规格§4，Model类型在input.go声明，方法留Task2；owner.go仅声明input Input/summary Summary/valid bool供本任务编译。TestLeafBindingAPI不调用尚未实施的Evaluate/方法。private profile常量固定字符串。错误恰ErrInvalidInput/ErrLimitExceeded/ErrInvalidEvidence/ErrBindingMismatch，文字分别`nodeleafbinding: invalid input/limit exceeded/invalid evidence/binding mismatch`。Obligation零UnknownObligation，1–5严格按规格名单。克隆前所有容量检查，无提前空锚错误；消费例外exact根/目录/package/import，不扩大其他包能力。
- [x] **Step 4: 重跑Step2至GREEN。** 取得实际exit0；容量控制与消费控制分别报告。
- [x] **Step 5: 本地检查点。** hash与九旧diff核对，仅登记任务结果，不暂存/提交。

### Task 2: 完整Input所有权与共享Model生命周期

**Files:** Modify owner.go补函数；Create七owned文件、model.go，以及对应七owned测试、model_test.go。

**Interfaces:** Consumes Task1 preflight；produces `owner{input Input; summary Summary; valid bool}`，
`ownInput(context.Context,Input)(*owner,error)`、`(*owner).clear()`；
`copySlice[T any]([]T)[]T`、`copyRaw([]byte)[]byte`、`normalizeTime(time.Time)time.Time`；
`ownCertificate(context.Context,nodeissuer.Input)(nodeissuer.Input,error)`、`clearCertificate(*nodeissuer.Input)`；
`ownBinding(context.Context,statebinding.Input)(statebinding.Input,error)`、`clearBinding(*statebinding.Input)`；
`ownContent(context.Context,statecontract.MainStateContentModel)(statecontract.MainStateContentModel,error)`、`clearContent(*statecontract.MainStateContentModel)`；
`ownTable(context.Context,statematerials.Input)(statematerials.Input,error)`、`clearTable(*statematerials.Input)`；
`ownPolicy(context.Context,statepolicy.ModelInput)(statepolicy.ModelInput,error)`、`clearPolicy(*statepolicy.ModelInput)`；
`ownPolicyBinding(context.Context,statepolicybinding.Input)(statepolicybinding.Input,error)`、`clearPolicyBinding(*statepolicybinding.Input)`；
`ownTimeBinding(context.Context,statetimebinding.Input)(statetimebinding.Input,error)`、`clearTimeBinding(*statetimebinding.Input)`；
`ownEvidence(nodecerttime.Summary)nodecerttime.Summary`、`clearEvidence(*nodecerttime.Summary)`；
`(*Model).Summary() Summary`、`(*Model).Clear()`。

- [x] **Step 1: 写失败测试。** TestLeafBindingOwnedCertificate/Binding/Content/Table/Policy/Time/Summary逐一用本任务合成模型填充所有嵌套字段（不依赖Task3签名fixture），比较规范化期望且修改每个raw/集合/Old不影响另一侧。TestLeafBindingOwnedAliasIsolation以重复backing/共享raw/Old检查失败清理不伤caller。TestLeafBindingModel用测试内显式owner覆盖nil/zero、共享包装Clear、重复Clear、独立owner/摘要集合，不调用尚未实现的Evaluate。
```go
copyModel:=*model
model.Clear()
if !reflect.DeepEqual(copyModel.Summary(),Summary{}) { t.Fatal("copy retained capability") }
```
- [x] **Step 2: 运行RED。** `& $leafBindingGo test ./internal/nodecontrol/nodeleafbinding -run '^TestLeafBinding(Owned|Model)' -count=1 -p=1 -timeout=120s`；缺少own/model helper或断言失败，不接受偶然fixture错误。
- [x] **Step 3: 实现完整有界复制/清理。** 参考基线nodecerttime相应私有函数形状，在新包实现，不导出旧helper。全部raw/集合/Old复制到解除别名的字段后登记清理；normalizeTime仅Round(0).UTC()。ownEvidence先复制值、将两个slice置nil，再建立失败保护后分别克隆；clearEvidence只清零自有Links/Coverage后清整个摘要。Model.Summary带返回失败清理，Clear共享owner。
- [x] **Step 4: 重跑Step2取得GREEN。** 可达backing清零与不可达runtime限制分别陈述；不读取caller私钥或声称并发安全。
- [x] **Step 5: 本地检查点。** 完整输入字段覆盖清单对照实际类型，hash/ownership断言登记；不Git写入。

### Task 3: 固定材料、一次真实lower判定与错误保真

**Files:** Create fixtures_test.go、lower.go、lower_test.go。

**Interfaces:** Consumes Task2 owner；produces `stageError(context.Context,error)error`、
`evaluateLower(context.Context,*owner)error`、`validateEvidence(nodecerttime.Summary)error`。
Test-only `leafBindingFixture(testing.TB,string)Input`、
`assertLeafBindingLower(testing.TB,Input)nodecerttime.Summary`。

- [x] **Step 1: 冻结测试材料及独立核对。** 用本轮忽略目录一次性脚本通过Go AST提取原certTimeFrozenInputs/Manifest的字符串值，不能靠脆弱regex或改旧文件；复制原34输入常量到新fixtures_test.go并保留原摘要检查。在一次性生成器中用真实nodecerttime输出设置两侧Identity.Leaf/PublicKey/Window四字段构造本层匹配输入，其他未覆盖字段保持原值；这些是未签名模型声明，不重签raw。冻结新的匹配Input及独立预期Projection/摘要/raw hash到本层测试常量；leafBindingFixture只解码冻结Input，普通测试不得动态重建匹配字段或根据受测本层输出推导期望。

独立checker使用既有固定raw，核对leaf DER/SPKI摘要、解析窗口、原数字序列号及真实nodecerttime成功；记录所有新匹配声明的变化路径、旧raw完全不变。若需新增同SPKI不同DER或year-1证书，可仅在专属忽略生成器用固定测试key一次性生成并独立核对，再作为冻结常量放新测试；不改旧fixture/生产材料或普通测试随机重签。

- [x] **Step 2: 写lower失败测试及运行RED。** TestLeafBindingFixtureControls核对新/旧hash；TestLeafBindingLowerErrors在测试自行保护的owner上调用evaluateLower，对真实坏签名/过期/time PolicyFailure与context断言原错误分类/错误链且owner.valid仍false，未实现Evaluate前不调用它。TestLeafBindingLowerOwnership验证下层返回后即Clear、Evidence捕获到保护owner、parent Time.ValidUntil可晚于证书期限；TestLeafBindingProjectionControls先证明匹配输入及后续公钥/窗口错配输入仍完整nodecerttime通过。
Run: `& $leafBindingGo test ./internal/nodecontrol/nodeleafbinding -run '^TestLeafBinding(FixtureControls|Lower|ProjectionControls)' -count=1 -p=1 -timeout=120s`；预期未实现lower或关键断言失败。
- [x] **Step 3: 实现lower。** 直接`nodecerttime.Evaluate(ctx,o.input)`恰一次；owner返回立即defer Clear，再阶段context/error；直接`o.summary.Evidence=lower.Summary()`，本层owner已受保护；验证非nil、exact Proof、非零/year1–9999/ObservedAt<ValidUntil及ObservedAt==Time.Policy.Time.Now。validateEvidence为私有摘要一致性函数，测试可直接构造矛盾值，无产品替换判定器hook。
- [x] **Step 4: 重跑Step2取得GREEN。** 失败保留，不把checker/raw常量核对等同完整平台认证；固定材料不依赖旧测试的运行时文件读取。
- [x] **Step 5: 本地检查点。** 原rawhash、固定新声明、实际lower错误/cleanup证据冻结，旧fixtures保持。

### Task 4: 精确Proposed投影、全流程发布和取消

**Files:** Create projection.go、evaluate.go、projection_test.go、evaluate_test.go。

**Interfaces:** Consumes Tasks1–3；produces
`bindProjection(context.Context,Input,nodecerttime.Summary)(Projection,error)`、
`windowMatches(statecontract.TimeWindow,time.Time,time.Time)bool`、
`Evaluate(context.Context,Input)(*Model,error)`。

- [x] **Step 1: 写投影/时窗失败测试。** TestLeafBindingProjection逐一错node/header/kind/P/E、公钥、Issued/Until。对可达错配同时一致改旧新声明并assertLeafBindingLower成功，再要求ErrBindingMismatch；node/header/kind/E非法由lower先拒绝时单独标作前置负例，不能声称达到新层。TestLeafBindingSameKeyDifferentDER用同SPKI不同有效签名leaf拒绝原DER引用。
```go
assertLeafBindingLower(t, wrongPublicKeyButCoherentMains)
m,e:=Evaluate(context.Background(),wrongPublicKeyButCoherentMains)
if e!=ErrBindingMismatch || m!=nil { t.Fatal("unbound projection",e) }
```
TestLeafBindingWindowEdges按两个端点equal/-1s/+1s；YearOneWindow允许合法year1证书起点与Main秒值；超范围/反向窗口分为直接windowMatches拒绝及真实lower前置拒绝，不伪造可达。
- [x] **Step 2: 写责任/时间/取消失败测试。** TestLeafBindingEqualSKIValue对恰等SPKI摘要的SKI数值允许、责任完整；UncoveredReferences对issuer/serial/keyhandle/lineage/history在下层允许范围内一致改两侧仍有限成功；TemporalInheritance检查纳秒Now、父Time.ValidUntil晚于CA/leaf、旧新观测不同需完整Evaluate；ErrorPriority检查lower过期先于公钥错配。CancellationAndPanic通过计数context自动枚举本层与同步下层所有可观察Err调用点，分别注入Canceled/DeadlineExceeded/panic：nil输出或清理后panic、caller不变，含最终发布；不固定声称与旧轮602点相同或硬实时。
- [x] **Step 3: 运行RED。** `& $leafBindingGo test ./internal/nodecontrol/nodeleafbinding -run '^TestLeafBinding(Projection|SameKeyDifferentDER|WindowEdges|YearOneWindow|EqualSKIValue|UncoveredReferences|TemporalInheritance|ErrorPriority|CancellationAndPanic)$' -count=1 -p=1 -timeout=120s`；预期未实现或指定行为失败。
- [x] **Step 4: 实现投影与Evaluate。** windowMatches先整数范围/order检查，time.Unix(n,0).UTC().Equal端点。bindProjection按规格表顺序严格对应，返回仅已绑定四字段。Evaluate保护ownInput owner后调用evaluateLower，再bindProjection；填五责任和原期限，最后stageError(ctx,nil)成功才valid=true并移交。无extra Now/缓存/重试/新crypto。
- [x] **Step 5: 重跑Step3及完整新包。** `& $leafBindingGo test ./internal/nodecontrol/nodeleafbinding -count=1 -p=1 -timeout=120s`实际PASS/exit0；登记取消观察点实际数量、正常panic清理边界与未认证责任，不Git写入。

### Task 5: 能力/所有权/消费guard、最终矩阵与fresh审查

**Files:** Create/Test boundary_test.go、ownership_boundary_test.go、consumer_boundary_test.go；发现问题只修批准新文件/九旧测试；真实完成后新增执行记录并追加current-status。

**Interfaces:** Consumes精确33文件库存、所有公开/私有接口；produces test-only
`leafBindingProductAllowed([]byte)bool`、`leafBindingCapabilityAllowed(map[string][]byte)bool`、
`leafBindingOwnershipAllowed(map[string][]byte)bool`、`leafBindingConsumerAllowed(path,pkg,imported string)bool`；
源码mutation须实际改变一处，否则测试控制本身失败。

- [x] **Step 1: 写guard失败测试并运行RED。** TestLeafBindingProductBoundary/CapabilityBoundary/OwnershipBoundary/Consumers分别检查实际源码、精确API/33库存/四错误/profile/Proof/owner三字段、共享清理和消费者。mutation分别删除/后移预算、完整input保护、部分clone失败保护、lower即Clear、受保护Evidence捕获、两slice深复制、最终context/valid顺序、各投影比较、范围先于Unix、Equal变成Unix截断；均须拒绝。良性格式/局部变量命名/命名import别名允许。禁止能力、方法值/函数值逃逸、shadowing、dot/blank/重复imports、extra导出/globalhook/子目录拒绝。
Run: `& $leafBindingGo test ./internal/nodecontrol/nodeleafbinding -run '^TestLeafBinding(ProductBoundary|CapabilityBoundary|OwnershipBoundary|Consumers)$' -count=1 -p=1 -timeout=120s`。
- [x] **Step 2: 实现精确AST守卫并重跑GREEN。** 下表为package成员白名单；类型嵌套字段访问限于完整值复制/预算或显式投影，不默认开放整包。所有产品helper调用只允许库存中声明的函数/方法；generic copy推导未直接导入嵌套类型。

| import（本地前缀 talenro.local/platform/internal/nodecontrol/） | package成员 |
| --- | --- |
| context/errors/bytes/sort/time | Context / New / Clone,Compare / Slice / Time,Unix |
| contracts | Digest |
| nodecerttime | Input,Summary,Evaluate,NodeCertificateConditionalTimeBound |
| nodeissuer | Input |
| statetimebinding | Input,Summary |
| statepolicybinding/statebinding | Input |
| statecontract | MainStateContentModel,TimeWindow,DomainIdentity,Present,CertificateDER |
| statematerials | Input,RootSet,RootRotation |
| statepolicy | ModelInput,StateModel |

方法仅context.Err，time.Round(0)/UTC/IsZero/Year/Before，projection.go中
窗口Equal与time.Unix(已检查整数,0)，lower.Summary/Clear，UUID.String仅
projection.go，已拥有Old.Clear及本包私有方法。只在lower.go直接调用一次
nodecerttime.Evaluate；产品及新测试不导入stateverify/stateformat/
statecomposition/statetime/nodeca/nodeleaf/nodeleafkeys，新增测试可用
nodeissuer/statetimebinding类型但不额外调用它们Evaluate来替代整lower控制。
fixture独立parser/crypto核对工具仅专属忽略目录，不放产品。
新包自身消费者仅exact本目录/package的_test.go；产品/外部测试/sibling/
其他根/subpath均拒绝，不能反向授权serving/provider。

- [x] **Step 3: 格式和24个最终命令。** 仅33新Go+九旧测试gofmt -l，期望空输出/exit0；必要gofmt -w限这些路径。以下每行单独记录，readiness路径已只读核对存在，不用./...。

```powershell
& $leafBindingGo test ./internal/nodecontrol/nodeleafbinding -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/nodecerttime -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/nodeissuer -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/nodeca -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/nodeleaf -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/nodeleafkeys -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/contracts -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/serving -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/statebinding -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/statecomposition -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/statecontract -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/stateformat -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/statematerials -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/statepolicy -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/statepolicybinding -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/staterecovery -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/statetime -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/statetimebinding -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/stateverify -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/readiness -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/store -run '^$' -count=1 -p=1 -timeout=120s
& $leafBindingGo test ./internal/nodecontrol/nodeissuer -run '^$' -fuzz '^FuzzIssuerEnvelope$' -fuzztime=5s -parallel=1 -p=1 -timeout=30s
& $leafBindingGo test ./internal/nodecontrol/nodeissuer -run '^$' -fuzz '^FuzzIssuerPair$' -fuzztime=5s -parallel=1 -p=1 -timeout=30s
```

期望24实际exit0：20完整包、authority指定测试、store compile/no tests、两
既有短fuzz；fuzz实际基线/次数/进程时长如实记，5s不是整个进程期限。
失败先systematic-debugging，不跳过当通过、擅改timeout/guard或扩大产品。

- [x] **Step 4: 冻结并一次fresh只读审查。** Native实现完后派一位新上下文reviewer，按届时可用工具的最强模型/支持的高推理级审本增量相对a820f550实际diff、全部新文件、规格/计划及原始实证；不把旧实现当新贡献。审查只读不Git写入，不扩大运行验收。报告真实Critical/Important/Minor，修复须有对应RED/GREEN，重跑影响项和最终矩阵；同reviewer复查，不预写零问题。本计划批准将允许该末尾独立review步骤，不允许并行实现或另建用户chat。
- [x] **Step 5: 如实交付本机检查点。** 逐任务结果/材料核对/42Go格式/24矩阵/源码前后hash/review结论记录到执行文档，只标实际完成checkbox；未解决阻塞和取消/失败保留。没有独立Git授权时停在本机待保存，明确不是远端恢复点，不推送main或合并。

## 3. 自检与审批交接

执行勘误（保留上方批准时的原命令）：Task5 store 的真实目录为
`./internal/store`，不是 `./internal/nodecontrol/store`。071 原路径实际
找不到目录；后续只对同一 store 包按原 `-run '^$'` 和 120s 限制编译，
不运行测试，不扩大验收。证据和代价记录于执行账本 R04。

规格覆盖映射：§1–4→Task1合同/Task4投影和责任；§5→Task1完整预算与
Task2所有权、Task5mutation；§6→Task3原错误/真实lower与Task4发布；
§7→Task1九旧例外/Task5新能力guard；§8→Task3固定控制、Task4边界/
取消/未覆盖责任、Task5有限矩阵/fresh审查；§9→本计划审批与Git门槛。
五Review Focus均有任务命名测试；库存33=16产品+17测试，九旧测试共42
格式路径，矩阵20+1+1+2=24。readiness真实路径只读核对，不复制旧路径错误。
作者自检不是独立审查或新增测试成功；本轮没有执行任何Go命令。

请用户审阅并批准本计划，确认其范围符合预期；保留此前Native执行选择，
不用重新选择方法。批准后才使用executing-plans开始Task1，执行前重读适用
技能与规则。Git保存/推送、main合并和生产验收仍须分别满足其原门槛。
