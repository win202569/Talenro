# Main State Selected Issuer/Serial Local Reference Binding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增 nodeissuerbinding，把 Proposed 的 issuer/serial 摘要按明确版本化的本地规则对应到一次真实 nodeleafbinding 判定，不产生生产标识兼容或运行授权。

**Architecture:** 完整预复制预算和自有 Input 先于真实下层；完整 Evidence 捕获到保护 owner 后，使用固定域和固定容量 scratch 派生两个引用，并逐项比较 Proposed。完整保留原叶投影、时间、错误与责任，最后 context 检查后才发布；模型共享 owner，摘要独立深复制。

**Tech Stack:** Windows PowerShell、缓存 Go 1.26.5、既有 nodeleafbinding/限定值类型、标准库 sha256.Sum256；无新依赖、parser、provider、wire 或运行消费者。

**Spec:** [已批准书面规格](../specs/2026-10-13-main-state-selected-issuer-serial-binding-design.md)。执行者完整阅读规格和本计划。

状态：用户“批准”批准书面规格，授权编写本计划；规格正文的“待书面审阅”
保留起草历史。本文待审阅，本轮只新增本文，不实现、执行 Go、生成材料、
创建执行实证、启动代理、暂存、提交或推送。沿用此前 Native：计划批准后
使用 executing-plans 逐任务实现，末尾一位 fresh 只读增量 reviewer；
不派并行实现者，不将规格批准当作本计划批准。

工作区 D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator，
branch=codex/c12-b01-task9-coordinator，
HEAD=649cc0891011334dd032bcc4a9618110b0a9c5ce。
规格SHA256=A4F48076299E2C13C0C37C03276860262AACB0C6420D7B534CA18345CD2E3BF9。
新规格/计划仍是本机文件；此前远端649cc089不是这份新计划的恢复点。

## Global Constraints

- 唯一新产品目录 internal/nodecontrol/nodeissuerbinding，无子目录；旧源码仅§1十个测试例外，旧产品及其AST schema/fixtures/SQL/模块/配置/历史设计均不改。
- 固定 Profile=main-selected-issuer-serial.ca-der-bound.v1；Input=nodeleafbinding.Input 别名，无格式选择器、旧格式fallback或生产issuer字符串输入。
- DI="talenro.main-selected-issuer-ref.v1\x00"，DS="talenro.main-selected-serial-ref.v1\x00"，各ASCII36 bytes，末尾恰一个NUL，不是可见四字符。
- C=真实下层 E.Evidence.Certificate.CA.CADERDigest；S/L=E.Evidence.Certificate.Leaf.Leaf.Serial/SerialLength；不能使用CA自身serial或调用方摘要替代真实材料。
- I=SHA256(DI||C)，68-byte preimage；R=SHA256(DS||I||U8(L)||S[0:L])，70–89 bytes；只 hash实际长度，不是hex/JSON/DER INTEGER/数字字符串/整容量。
- C非零；L为1–20，S[0]非零，S[L:20]全零；首字节>=0x80允许。先范围再索引，不修复表示；[68]/[89]scratch先保护后填，用后清零，正常成功路径各一次Sum256。
- Proposed.Identity.IssuerID==I，Serial==R；Proposed.Leaf完整沿用下层Projection，不写回Input、不认证Previous、不改原同Leaf冲突规则。
- CA续签即使同key/name仍是不同exact DER preimage；本地引用不映射存储issuer字符串，不证明部署来源/唯一性/私钥；生产兼容UNKNOWN。
- Certificate raw上限65,536/512/65,536/8,192，合计139,776；TrustDomain253 bytes；两表各Materials/Expected128，各RootRotation32；Main.RootTransition/Target.Links各32。
- RootSet raw65,536，其余raw/Anchor262,144；每表累计16MiB；Binding总量32MiB+262,144；Expected root keys/metadata keys/revoked/desired slots=5/64/4096/8。
- 两份Main各root keys5、metadata keys64/revoked4096、bootstrap/agent deauthorized各128、Desired/LKG/RecoveryLocal/RecoverySupervisor=8/8/64/16。
- Batch.Rotations32；四固定和各rotation raw262,144；Batch累计36×262,144；Binding+Batch总量32MiB+37×262,144。
- Old.Roots33，每root keys5；Old.Keys/Old.RevokedIDs/History各4096；Status/RevocationReason/IncidentKind长度8/32/64 bytes。
- 两time raw各262,144，合计524,288；Time路径总量43,778,048；完整Input raw43,917,824。每个出现位置计数、溢出安全、不按backing去重；全部预算先于五raw必需非空及任何复制。
- 完整十二域/表/Batch/Old/History/证书/时间复制；仅两表顶层按Kind/P排序、nil-empty规范化和time.Round(0).UTC()，其他有序字段不改，零值和纳秒保持。
- 一次真实nodeleafbinding.Evaluate；owner返回即defer Clear，先于context/error；完整Summary直接存入受保护owner.summary.Evidence；不调用更低Evaluate/Verify/parser。
- Evidence所有Proof/Pending/Obligations保留；不得删除下层IssuerAndSerialReferenceMapping。新增Proof只完成新本地关系，五本层责任固定为规格§4顺序。
- E.Proof恰SelectedLeafProjectionConditionalTimeBound；ObservedAt/ValidUntil非零、UTC年1–9999、ObservedAt<ValidUntil；E.ObservedAt==E.Evidence.ObservedAt==E.Evidence.Time.Policy.Time.Now，E.ValidUntil==E.Evidence.ValidUntil，规范化完整值比较，不用Equal。
- E.Proposed与自有Main选定四字段相同；不要求E.ValidUntil==E.Evidence.Time.ValidUntil。输出时间精确沿用E，新观测完整重评估，无缓存许可。
- 两可变slice在Summary.Evidence.Evidence.Time.Policy.Binding.Proposed下的Target.Links和Coverage；返回摘要深复制并保护失败返回，清理只影响自有backing。
- Model浅拷贝共享owner；Clear nil-safe/幂等并使全部副本无效；nil/zero/清理后完整零摘要含Profile=""；不支持并发，不承诺隐藏runtime副本/物理擦除/硬实时/OOM/crash恢复。
- 当前可观察context错误优先；下层非nil错误原样保留含typed PolicyFailure/错误链；下层失败先于新错配；最终context后才valid，失败nil、正常panic清理后传播。
- 公开API/五责任/四固定错误严格按Task1；无Ready/Authorize/Commit/Recheck、raw/owner访问器、hook或额外输出。只用限定context/errors/bytes/sort/time和sha256.Sum256，无其他crypto/编码/反射/I/O/环境/时钟/网络/provider/unsafe/goroutine/recover。
- 非认证key-handle/epoch/lineage/history/row、部署issuer映射、平台/观测来源、完整identity/commit/runtime责任保留；同bytes不能证明赋值来源，有限hash/AST测试不证明无碰撞或形式安全。
- 25限定命令和45 Go路径格式检查；不整仓/race/Linux、真实DB/Docker/TLS/provider/硬件/reboot/PITR，不解除C12/B01/Task10/main/I2/I3及历史失败原因阻塞。
- 本计划批准不授权Git写入、PR、推送、main合并或清理。技能默认提交步骤改为本地实证检查点，另需独立Git授权。

## Review Focus

- serial首byte>=0x80误当负数或加入DER符号00、hash错误容量：Task3 TestIssuerBindingCanonicalSerial/ReferenceVectors 固定长度1/20和高位正例；20-byte高位仅编码helper正例，不能误称真实lower接受。
- 同key/name的CA续签被当同issuer、公钥/AKI冒充DER引用：Task3固定真实材料，Task4 TestIssuerBindingSameKeyDifferentCAReferences先证明下层成功再拒绝旧引用。
- 同leaf只改一侧被旧transition先拒绝而伪称新层成功防护：Task4 TestIssuerBindingProjectionControls一致改两侧且先断言真实lower通过。
- 同instant不同location误过完整值时间一致性：Task3 TestIssuerBindingLowerConsistency构造Equal为真而完整值不同，要求ErrInvalidEvidence。
- 共享slice/部分失败伤caller、lower返回后ctx先于保护或scratch清理遗漏：Task2 OwnedAliasIsolation，Task3 ReferenceCleanup，Task4 CancellationAndPanic及Task5 OwnershipBoundary实际mutation。

---

## 1. 精确文件库存、跨任务接口与环境

新目录恰35个Go文件：17产品、18测试。下表文件名前缀为
internal/nodecontrol/nodeissuerbinding/；无永久生成器或子目录。

| 产品文件 | 职责 | 对应测试 |
| --- | --- | --- |
| doc.go / input.go / errors.go | 局部合同/API/四错误 | input_test.go |
| budget.go | 完整预复制预算 | budget_test.go |
| owner.go | 三字段owner/复制移交 | model_test.go |
| owned_certificate.go | 四certificate raw | owned_certificate_test.go |
| owned_binding.go | Main/两表/target/anchor | owned_binding_test.go |
| owned_content.go | 十二域值/集合 | owned_content_test.go |
| owned_table.go | Materials/Expected | owned_table_test.go |
| owned_policy.go | PolicyBinding/Batch/Old/History | owned_policy_test.go |
| owned_time.go | 两time raw及观测 | owned_time_test.go |
| owned_summary.go | 完整nodeleafbinding.Summary/两slices | owned_summary_test.go |
| model.go | detached Summary/共享Clear | model_test.go |
| lower.go | 一次真实lower/阶段错误/一致性 | lower_test.go |
| references.go | 固定域/scratch/两个hash | references_test.go |
| projection.go | issuer后serial逐项匹配 | projection_test.go |
| evaluate.go | 最后context与发布 | evaluate_test.go |

其余四测试：fixtures_test.go、boundary_test.go、
ownership_boundary_test.go、consumer_boundary_test.go。
Task1只声明owner/Model，方法在Task2、公开Evaluate在Task4；API完整
方法/函数检查分别归到这些任务及Task5，不能在Task1引用未定义实现。

十个旧测试只新增exact消费者例外/正反控制，不改旧产品或旧规则：

| internal/nodecontrol下路径 | helper |
| --- | --- |
| nodeleafbinding/consumer_boundary_test.go | leafBindingConsumerAllowed |
| nodecerttime/consumer_boundary_test.go | certTimeConsumerAllowed |
| nodeissuer/boundary_test.go | issuerConsumerAllowed |
| statetimebinding/boundary_test.go | timeBindingConsumerAllowed |
| statepolicybinding/boundary_test.go | policyBindingConsumerAllowed |
| statebinding/boundary_test.go | bindingConsumerAllowed |
| statecontract/boundary_test.go | contractConsumerAllowed |
| statematerials/boundary_test.go | materialsConsumerAllowed |
| statepolicy/boundary_test.go | policyConsumerAllowed |
| stateverify/boundary_test.go | isolatedImportAllowed，仅statepolicy |

执行前重读适用技能/AGENTS、批准spec/plan和现有类型；只读核对已隔离
worktree/branch/HEAD/spec hash，不另造worktree、重置或覆盖用户改动。
若基线/文件改变，先辨识影响，不默认为当前规格仍获批准。
冻结旧Go/SQL/module及十旧测试；每步核对批准差异和未准许旧产品。
本增量实证目录为 .superpowers/sdd/2026-10-13-main-state-selected-issuer-serial-binding，
独立于前轮，不覆盖其生成器、缓存、归档或回执。准备只能在实施获批后做。

每Go命令独立记录原可执行路径/参数/进程环境、UTC始末、实际exit、
完整stdout/stderr及源码前后hash；RunId禁止覆写，RED/失败不得当PASS。
使用PowerShell对应Bash-only技能记账，保留plan-path/brief/base/receipt，
不以适配为理由修改Go/Git持久配置、安装依赖或扩大测试。

~~~powershell
Set-Location -LiteralPath 'D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator'
$issuerBindingGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
$issuerBindingFmt='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/gofmt.exe'
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
$env:GOCACHE='D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/.superpowers/sdd/2026-10-13-main-state-selected-issuer-serial-binding/go-cache'
~~~

缓存可执行文件须先只读确认存在；缺失则记录阻塞，不自动下载。
未来完成后只新增 docs/roadmap/2026-10-13-main-state-selected-issuer-serial-binding-execution.md，
并追加 docs/roadmap/current-status.md真实结果；不回写旧review/验收。

## 2. 五个顺序任务

### Task 1: 合同、预算和十个精确消费例外

**Files:** Create doc.go/input.go/errors.go/budget.go/owner.go；
Test input_test.go/budget_test.go；Modify §1十个旧测试的具名helper及新控制。

**Interfaces:** Produces规格§4的Input/Proof/Obligation/Projection/Summary/Model，
私有owner {input Input; summary Summary; valid bool}；budget.go中：
addRaw(total,size,limit int)(int,bool)，certificateBudget(nodeissuer.Input)(int,error)，
contentBudget(statecontract.MainStateContentModel)bool，
bindingBudget(statebinding.Input)(int,error)，timeBudget(statetimebinding.Input)(int,error)，
preflight(context.Context,Input)error。不定义假Evaluate或导出旧helper。

- [x] **Step 1: 写失败测试。** TestIssuerBindingAPI反射类型别名/字段/uint8及enum顺序，暂不引用Evaluate/Model方法。TestIssuerBindingFixedErrors核对四固定文字；Budget/Preflight覆盖Global Constraints每项等长/+1、总量/溢出/重复位置/未选材料和预算先于空raw。十旧测试分别新增唯一名以IssuerBindingBoundary结尾的测试；exact目录/package/导入为真，external_test/subdir/sibling/other-root/subpath/伪package为假，旧正反例保留。
~~~go
if reflect.TypeOf(Input{}) != reflect.TypeOf(nodeleafbinding.Input{}) { t.Fatal("alias") }
if uint8(None)!=0 || uint8(SelectedIssuerSerialReferencesConditionalTimeBound)!=1 { t.Fatal("proof") }
if ErrBindingMismatch.Error()!="nodeissuerbinding: binding mismatch" { t.Fatal("fixed text") }
// 一个材料在两个位置出现时都计入预算，不能因共享backing而减半。
~~~
- [x] **Step 2: 运行RED。** 新包用下列选择器，预期类型/预算未定义；十旧包各单独用 -run 'IssuerBindingBoundary$'，预期缺精确例外而失败，记录11回执，不接受任意fixture故障。
~~~powershell
& $issuerBindingGo test ./internal/nodecontrol/nodeissuerbinding -run '^TestIssuerBinding(API|FixedErrors|Budget|Preflight)$' -count=1 -p=1 -timeout=120s
~~~
十旧package命令目标恰为§1表的目录，其他参数 -count=1 -p=1 -timeout=120s。
- [x] **Step 3: 实现合同/预算/例外。** input.go私有常量selectedIssuerSerialProfile固定为Global Constraints的Profile，公开API及两enum恰规格§4。errors.go恰四errors.New：ErrInvalidInput="nodeissuerbinding: invalid input"，ErrLimitExceeded="nodeissuerbinding: limit exceeded"，ErrInvalidEvidence="nodeissuerbinding: invalid evidence"，ErrBindingMismatch="nodeissuerbinding: binding mismatch"。budget.go参考基线完整限制逐项实现，不调用旧私有helper；owner.go先仅声明三字段。十旧helper仅增加exact新目录/package/目标导入分支，stateverify仅statepolicy，不开放整包能力。
- [x] **Step 4: 重跑Step2取得GREEN。** 11独立命令都实际exit0；每一个预算限制有对应测试，确认其他旧文件/原规则、规格hash不变。
- [x] **Step 5: 本地检查点。** 保存前后hash/RED-GREEN/仅十旧测试的实际diff；不Git写入，不称此阶段完整Evaluate可用。

### Task 2: 完整自有复制、摘要和模型生命周期

**Files:** Create七个owned_*.go、model.go；Update owner.go；
Test七个owned_*_test.go、model_test.go。

**Interfaces:** Consumes Task1 preflight/owner/Input；produces以下私有接口，
类型前缀均为既有对应包，不创建新别名隐藏漏字段：
~~~text
copySlice[T any]([]T) []T; copyRaw([]byte) []byte; normalizeTime(time.Time) time.Time
ownInput(context.Context,Input)(*owner,error); (*owner).clear()
ownCertificate(context.Context,nodeissuer.Input)(nodeissuer.Input,error); clearCertificate(*nodeissuer.Input)
ownContent(context.Context,statecontract.MainStateContentModel)(statecontract.MainStateContentModel,error); clearContent(*statecontract.MainStateContentModel)
ownTable(context.Context,statematerials.Input)(statematerials.Input,error); clearTable(*statematerials.Input)
ownBinding(context.Context,statebinding.Input)(statebinding.Input,error); clearBinding(*statebinding.Input)
ownPolicy(context.Context,statepolicy.ModelInput)(statepolicy.ModelInput,error); clearPolicy(*statepolicy.ModelInput)
ownPolicyBinding(context.Context,statepolicybinding.Input)(statepolicybinding.Input,error); clearPolicyBinding(*statepolicybinding.Input)
ownTimeBinding(context.Context,statetimebinding.Input)(statetimebinding.Input,error); clearTimeBinding(*statetimebinding.Input)
ownEvidence(nodeleafbinding.Summary) nodeleafbinding.Summary; clearEvidence(*nodeleafbinding.Summary)
(*Model).Summary() Summary; (*Model).Clear()
~~~

- [x] **Step 1: 写失败测试。** OwnedCertificate/Binding/Content/Table/Policy/Time/Summary逐字段构造非零合成模型，不依赖Task3签名fixture；必要raw可为非空合成bytes，复制不等于语义验证。OwnedComplete用反射对照完整嵌套类型与集合；OwnedAliasIsolation覆盖重复raw/backing/Old/半途失败、caller不变。Model覆盖nil/zero、独立owner、共享包装Clear、返回摘要两slices隔离和完整Profile零值，不调用未实现Evaluate。
~~~go
copyModel:=*model
model.Clear()
if !reflect.DeepEqual(copyModel.Summary(),Summary{}) { t.Fatal("shared owner survived") }
// 两slice路径：Evidence.Evidence.Time.Policy.Binding.Proposed.Target.Links / Coverage。
~~~
- [x] **Step 2: 运行RED。** & $issuerBindingGo test ./internal/nodecontrol/nodeissuerbinding -run '^TestIssuerBinding(Owned|Model|NormalizeTime)' -count=1 -p=1 -timeout=120s；预期缺复制/模型方法或指定隔离断言失败。
- [x] **Step 3: 实现完整所有权。** 七owned文件参考基线相应字段/排序/清理形状，在新包本地实现Task2接口；owner.go先preflight再all-copy。解除别名后、潜在ctx/失败/panic前登记各级清理。ownEvidence先复制值并把两slices置nil，保护失败后分别复制；model.go清空临时Evidence后保护返回，再取独立ownEvidence。不改下层产品或遗漏其字段。
- [x] **Step 4: 重跑Step2取得GREEN。** 所有raw/集合/Old及Summary覆盖清单与真实类型一致；回调枚举复制可观察取消/panic点，保留caller hash，可达自有清零与隐藏runtime限制分别记录。
- [x] **Step 5: 本地检查点。** 固定完整字段覆盖/失败保护证据；类型和方法形状供Task3/4使用，不Git写入。

### Task 3: 固定材料、一次真实lower与exact引用派生

**Files:** Create fixtures_test.go、lower.go/lower_test.go、
references.go/references_test.go。

**Interfaces:** Consumes完整owner/Input/clearEvidence；produces：
stageError(context.Context,error)error，
evaluateLower(context.Context,*owner)error，
validateEvidence(Input,nodeleafbinding.Summary)error，
referenceScratch struct {issuer [68]byte; serial [89]byte}，
(*referenceScratch).clear()，
deriveReferences(context.Context,nodeleafbinding.Summary,*referenceScratch)(contracts.Digest,contracts.Digest,error)。
仅私有derive可供同包测试检查scratch；公开接口不接受Summary或scratch。
Test-only issuerBindingFixture(testing.TB,string)Input、
assertIssuerBindingLower(testing.TB,Input)nodeleafbinding.Summary。

现有nodeleaf/leaf_fields.go的checkLeafFields先限制DER INTEGER content
最多20 bytes，再去合法前导符号00。本合同不放宽它：完整成功用L=1、
S[0]=0x80（DER content为00 80），及L=20、S[0]=0x7f；L=20且S[0]=0x80
的canonical magnitude只在私有derive单元测试证明编码允许。
相应真实21-byte DER content作为下层先拒绝的前置控制，不能记作本层
成功或改旧parser达成成功；它的错误由完整nodeleafbinding取得并原样比对。

- [x] **Step 1: 冻结材料并独立核对。** 隔离一次性AST读取基线leafBindingOriginalInputs/Manifest及leafBindingFrozenInputs/Manifest四常量，不改原文件。新常量命名issuerBindingOriginalInputs/Manifest、issuerBindingFrozenInputs/Manifest；原始packet/raw hash完整保留。按真实nodeleafbinding结果和独立§3编码仅一致改变两侧IssuerID/Serial，其他字段不变；固定expected I/R、preimages、全lower摘要hash和所有变化路径。必须增加真实同key/name不同CA DER、同leaf serial不同CA、serial长1/20及高位正例；固定测试key在一次性生成器生成新增有效签名材料，独立x509/sha256解析核对后冻结，普通测试不签发。生成器不调用新产品helper/Evaluate推导golden，用后byte-identical归档为非Go源码，记录hash和重放代价；全矩阵前不得留下旧guard会扫描的非法Go消费者。
- [x] **Step 2: 写失败测试。** FixtureControls核对旧raw/新增raw/全部声明变化；LowerErrors核对真实坏签名/过期/leaf投影/PolicyFailure原错误及链。LowerConsistency直接测wrongProof/零或越界时间/期限/投影，以及同instant不同location完整值不等；LowerOwnership测直接完整捕获、lower及时Clear、owner尚未valid。ReferenceVectors/CanonicalSerial核对两个域各36、68及69+L长度、I/R冻结值；L0/21、零C/首零/尾非零拒绝，首byte>=0x80合法。ReferenceCleanup以预填非零scratch测成功、取消/超时、错误及panic后两数组全零；wrong用途/版本/NUL/hex/顺序/整容量/CAserial向量的值先证不同再比较。
~~~go
bad:=expected
bad.ObservedAt=bad.ObservedAt.In(time.FixedZone("same-instant",3600))
if !bad.ObservedAt.Equal(expected.ObservedAt) || bad.ObservedAt==expected.ObservedAt { t.Fatal("control") }
if validateEvidence(in,bad)!=ErrInvalidEvidence { t.Fatal("normalized full value") }
// golden I/R由Step1独立冻结，不使用受测deriveReferences临时计算expected。
~~~
- [x] **Step 3: 运行RED。** & $issuerBindingGo test ./internal/nodecontrol/nodeissuerbinding -run '^TestIssuerBinding(FixtureControls|Lower|Reference|CanonicalSerial)' -count=1 -p=1 -timeout=120s；预期lower/derive未定义或对应断言失败；FixtureControls若自身出错先修独立控制，不把它记为产品RED。
- [x] **Step 4: 实现lower与references。** lower.go直接nodeleafbinding.Evaluate(ctx,o.input)一次，立即defer返回owner.Clear，阶段ctx/error后直接o.summary.Evidence=lower.Summary()；validateEvidence对照规格§6完整Proof/时间及自有Main四字段，不用Equal。references.go私有常量issuerReferenceDomain=DI、serialReferenceDomain=DS恰Global Constraints两个literal；derive在非nilscratch上先defer清理再可能ctx/失败，nilscratch为ErrInvalidEvidence、nilctx为ErrInvalidInput，当前ctx优先。先C/L/首尾检查，再填68/89数组、Sum256两个实际slice；两计算之间及返回前阶段ctx检查，返回digest值，不输出buffer，不append/reparse。
- [x] **Step 5: 重跑Step3取得GREEN。** 私有一致性和scratch正常/失败/panic清理全通过；真实lower失败仍原类型，成功证据直接受保护且全部Pending/Obligations相同。
- [x] **Step 6: 本地检查点。** 完整packet manifest、独立golden、原raw不变、新增证书关系及真实lower控制保存；生成器归档完成，不称来源/production mapping认证，不Git写入。

### Task 4: 精确Proposed引用、发布与取消

**Files:** Create projection.go/projection_test.go、
evaluate.go/evaluate_test.go。

**Interfaces:** Consumes Tasks1–3；
produces bindProjection(context.Context,Input,nodeleafbinding.Summary)(Projection,error)
及 Evaluate(context.Context,Input)(*Model,error)。
bindProjection使用本地referenceScratch传给Task3 deriveReferences，
Proposed.Leaf直接来自已核对的E.Proposed，不重新绑定/解析旧四字段。

- [x] **Step 1: 写失败测试。** Projection/ProjectionControls逐项错IssuerID、Serial、两者同时错：非零错值一致改Previous/Proposed、先完整assertIssuerBindingLower成功、再要求nil/ErrBindingMismatch；Issuer优先于Serial。Lower已拒绝的zero/同leaf单侧冲突独立标作前置控制。SameKeyDifferentCAReferences/SameSerialDifferentIssuer使用Task3真实固定材料，prove两种lower均通过，但old引用套新CA失败；bareDER/SPKI/Name/SKI/AKI、无域serial及CAserial替代各有数值不同控制。
evaluate_test同时写API函数签名、固定Profile/Proof/五责任、全Evidence相等、纳秒/原deadline、新观测完整Evaluate、下层过期加新错配原过期优先、未覆盖引用合法变化仍有限成功；动态枚举完整Err调用点注入取消/超时/panic，失败无模型且caller不变，不固定655点。
~~~go
assertIssuerBindingLower(t,wrongButCoherentMains)
m,e:=Evaluate(context.Background(),wrongButCoherentMains)
if m!=nil || e!=ErrBindingMismatch { t.Fatal("unbound local references",e) }
~~~
- [x] **Step 2: 运行RED。** & $issuerBindingGo test ./internal/nodecontrol/nodeissuerbinding -run '^TestIssuerBinding(Projection|ProjectionControls|SameKeyDifferentCAReferences|SameSerialDifferentIssuer|Evaluate|TemporalInheritance|ErrorPriority|UncoveredReferences|CancellationAndPanic)$' -count=1 -p=1 -timeout=120s；预期未定义Evaluate/bindProjection或对应引用/发布断言失败。
- [x] **Step 3: 实现projection和Evaluate。** projection.go先用Task3派生，再Issuer后Serial相等；成功Projection{Leaf:E.Proposed,IssuerID:I,Serial:R}。evaluate.go保护完整ownInput owner后一次evaluateLower、bindProjection、固定Proof/Profile/五责任和原时间；最终stageError(ctx,nil)后才valid=true并移交。任意失败/正常panic清理自身，不能recover为成功或重试。
- [x] **Step 4: 重跑Step2及完整新包取得GREEN。** Step2选择器和 & $issuerBindingGo test ./internal/nodecontrol/nodeissuerbinding -count=1 -p=1 -timeout=120s 分别记录；确认fullEvaluate与完整Golden、时间、责任、shared Clear及实际取消点；未认证责任不因本地两个引用已匹配而消失。
- [x] **Step 5: 本地检查点。** 函数/模型API与Task1规格一致；证明只能达到conditional local references，不能暗示签发唯一性、Previous或production/provenance有效，不Git写入。

### Task 5: 三类守卫、最终25矩阵、fresh审查与本机交付

**Files:** Create boundary_test.go、ownership_boundary_test.go、
consumer_boundary_test.go；问题修复仅35新Go/十旧测试；
最后才新增执行记录并追加current-status。

**Interfaces:** Consumes全部35库存/API/owners/两个固定域及流程；
test-only issuerBindingProductAllowed([]byte)bool，
issuerBindingCapabilityAllowed(map[string][]byte)bool，
issuerBindingOwnershipAllowed(map[string][]byte)bool，
issuerBindingConsumerAllowed(path,pkg,imported string)bool。

- [x] **Step 1: 写guard失败测试。** ProductBoundary/CapabilityBoundary/OwnershipBoundary/Consumers针对真实源码、exact35库存/17产品schema、Input别名/公开API、owner三字段、scratch两数组、五责任/四错误/Profile/DI/DS。实际mutation删除/后移预算/全输入保护/lower即时Clear/直接Evidence捕获/两slice复制/scratch保护/最终ctx、改变首尾/长度/来源/域/NUL/两Sum256实际slice、错Issuer或Serial比较、Profile或Proof；每次先断言源码确改变。禁止更低Evaluate/parser、额外hash、append/编码、方法值、shadow、dot/blank/重复import、导出/子目录。良性格式/局部命名/命名import别名可接受；消费者仅exact本目录/package_test.go。
- [x] **Step 2: 运行RED。** & $issuerBindingGo test ./internal/nodecontrol/nodeissuerbinding -run '^TestIssuerBinding(ProductBoundary|CapabilityBoundary|OwnershipBoundary|Consumers)$' -count=1 -p=1 -timeout=120s；预期未定义guard/对应拒绝断言失败。
- [x] **Step 3: 实现守卫。** 按规格§7和下表闭合AST成员/字段/方法库存，固定正常成功流程中只一直接nodeleafbinding.Evaluate、两sha256.Sum256及已检查长度slice；其他包仅复制预算类型，Old.Clear仅自有。Capability/Ownership用17产品规范化AST schemas和exact35库存；初始schema集合为空、闭合拒绝，不用TODO值。消费者实际扫描root，排除.git/.superpowers/.worktrees/vendor/node_modules等约定目录，保留原规则。

| import（本地包以talenro.local/platform/internal/nodecontrol/为前缀） | 产品package成员白名单 |
| --- | --- |
| context / errors / bytes / sort / time / crypto/sha256 | Context / New / Clone,Compare / Slice / Time / Sum256 |
| nodeleafbinding | Input,Summary,Projection,Evaluate,SelectedLeafProjectionConditionalTimeBound |
| contracts / nodecerttime / nodeissuer | Digest / Input,Summary / Input |
| statetimebinding / statepolicybinding / statebinding | Input,Summary / Input / Input |
| statecontract / statematerials / statepolicy | MainStateContentModel / Input,RootSet,RootRotation / ModelInput,StateModel |

字段只限完整复制/预算和规格的摘要/引用路径；方法仅context.Err、time.Round(0)/UTC/IsZero/Year/Before、真实lower.Summary/Clear、已自有Old.Clear和本包固定私有方法。不准time.Equal/Unix、UUID.String、更低Evaluate或sha256其他成员；测试独立编码/manifest可用标准库测试工具，但不调用被禁止的更低判定器替代完整nodeleafbinding控制。
- [x] **Step 4: 运行schema RED。** 同Step2选择器记录实际库存/schema拒绝及17产品规范化摘要；不得因fixture失败/白名单错误宣称schema RED。
- [x] **Step 5: 冻结认可schema。** 对照实际源代码、先前RED/GREEN和规格审计，填入17个真实固定摘要，保持其余能力/消费/所有权拒绝规则，不盲目刷新或放宽。
- [x] **Step 6: 重跑Step2取得GREEN。** 四guard及良性/实际负mutation全部通过；定型源码相对baseline仅批准范围，无额外产品或放宽consumer。
- [x] **Step 7: 格式并运行25最终命令。** 仅35新Go+十旧测试共45路径gofmt -l，期望空输出/exit0，必要gofmt -w只限这些路径；下列各命令独立原始回执，源码前后hash都等于最终同一冻结。
~~~powershell
& $issuerBindingGo test ./internal/nodecontrol/nodeissuerbinding -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/nodeleafbinding -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/nodecerttime -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/nodeissuer -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/nodeca -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/nodeleaf -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/nodeleafkeys -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/contracts -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/serving -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/statebinding -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/statecomposition -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/statecontract -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/stateformat -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/statematerials -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/statepolicy -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/statepolicybinding -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/staterecovery -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/statetime -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/statetimebinding -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/stateverify -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/readiness -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/store -run '^$' -count=1 -p=1 -timeout=120s
& $issuerBindingGo test ./internal/nodecontrol/nodeissuer -run '^$' -fuzz '^FuzzIssuerEnvelope$' -fuzztime=5s -parallel=1 -p=1 -timeout=30s
& $issuerBindingGo test ./internal/nodecontrol/nodeissuer -run '^$' -fuzz '^FuzzIssuerPair$' -fuzztime=5s -parallel=1 -p=1 -timeout=30s
~~~
期望25实际exit0：21完整包、指定authority、store仅编译、两既有短fuzz。
次数/新增interesting/基线/实际Go与整体进程时长如实记；不是5s墙钟保证。
失败先systematic-debugging，不跳过、加大timeout、改旧产品或guard掩盖。
- [x] **Step 8: 冻结后一次fresh只读质量/安全审查。** 按执行时最强可用模型/支持高推理级审本增量相对649cc089的十旧测试diff、全部35新文件、规格/计划和raw回执，不把旧实现算新贡献。Critical/Important/Minor真实报告；修复须对应RED/GREEN并重跑完整25矩阵、同reviewer复查。每项未判断领域由作者逐一明确范围/代价，不默认为已验收。本计划批准允许这一末尾独立审查，不允许并行实现或新用户chat；没有工具时报告缺门槛，不用作者自检冒充独立review。
- [x] **Step 9: 如实本机交付。** 执行记录登记各任务、真实材料/原raw保持、45格式、25矩阵前后hash、review及修复、取消点、失败和全部裁定/代价；current-status仅追加最后实际结果，勾选仅实际完成步骤。规格/未授权旧产品/SQL/module/HEAD/index保持；无Git授权则停本机待保存，不推送或合并，不宣传远端恢复或C12整体完成。

## 3. 作者自检与审批交接

规格覆盖：§1–2→Task1合同/Task4责任与边界；§3→Task3 exact引用、Task4匹配；
§4→Task1类型、Task2方法、Task4Evaluate/摘要；§5→Task1完整预算、Task2完整
复制、Task3scratch、Task5mutation；§6→Task3一次lower/一致性与Task4发布；
§7→Task1十旧例外和Task5能力/消费；§8→Task3独立固定材料、Task4可达控制/
取消、Task5矩阵/审查；§9→书面审批和本机/Git门槛。

五Review Focus均落到具名测试。35=17产品+18测试，格式35+10=45；
最终21+1+1+2=25。Task2 helper与真实类型/Task3摘要嵌套路径一致，
Task3固定域36/issuer68/serial70–89，Task4仅调用已定义跨任务接口。
Task1未提前引用Evaluate/Model方法；Task3独立fixture不依赖新Evaluate。
代码片段只固定API/编码/断言，未将函数实现写成计划，未留待填项。

本文自检不是执行、测试成功或独立review。本轮只新增本计划，批准规格
正文保持不变；未创建new产品、fixture、实证或缓存，不Go/Git写入。
按用户阶段约束，将技能“频繁提交”改为各任务本地检查点；代价是缺少
远端保存及独立阶段review，Native末尾才有fresh检查，整体仍未验收。

请用户审阅计划是否准确覆盖所需范围。沿用Native，无需重新选择方法；
计划书面批准后才读executing-plans等执行技能开始Task1。
实施完成后的Git保存/推送、main合并和整体验收继续各自独立审批。
