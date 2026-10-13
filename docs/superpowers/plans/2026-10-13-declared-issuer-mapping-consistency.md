# Declared Issuer Mapping Consistency Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. 本计划建议Native顺序执行；只有用户选定方法并批准本文后才实施。

**Goal:** 增加一个只检查登记/签发声明与真实证书材料对应关系的纯组件；成功仅表示声明一致，不表示生产来源可信或运行授权。

**Architecture:** 新增nodeissuermapping，唯一真实下层为nodeissuerbinding.Evaluate。
新层自有复制声明，下层独立预算/复制完整材料；直接捕获下层完整证据并及时Clear。
没有登记发布、数据库读取、迁移或v7生产能力；所有可信来源及唯一性责任保留。

**Tech Stack:** Go1.26.5、既有标准库及github.com/google/uuid v1.6.0；
Windows PowerShell、离线缓存、只读模块，不新增依赖。

**Spec:** [获批规格](../specs/2026-10-13-production-issuer-serial-mapping-compatibility-design.md)
SHA256：4AEB4D41F72066CDE1B129D3A42ECC81D33BFD4CD6D413E7849F67DD0F45857C。

## Global Constraints

- 基线d2829bfaa93ae2a119162a5ce6057fbe223d2cb7；分支codex/c12-b01-task9-coordinator；
  复用.worktrees/c12-b01-task9-coordinator，不新建/清理worktree。
- 仅实施规格§10首个“登记记录/版本关联的一致性与拒绝合同”增量。
  真实发布、签发来源、SQL持久化和v7接入不属于本计划，不得补临时可信格式。
- 唯一新产品目录internal/nodecontrol/nodeissuermapping；精确25个Go文件。
  唯一旧Go修改internal/nodecontrol/nodeissuerbinding/consumer_boundary_test.go，
  仅增加具名精确消费例外及控制，不修改任何旧产品/fixture/SQL/module。
- 新公开入口仅CheckDeclared；不增加Evaluate/Compatible/Ready/TrustedRegistry/
  生产reader工厂、数据库/provider/lease/epoch回调或verified flag。
- 输入登记/版本/签发身份均为调用方声明。即使全部字段一致，仍不能证明来源、
  当前发布头、防回退、真实签发关联、数据库唯一性、行状态或生产资格。
- 声明profile固定declared-issuer-serial.registry-version-link.v1；
  下层profile保持main-selected-issuer-serial.ca-der-bound.v1，
  不改其两个域/编码/原错误/时间/责任，不新增产品hash或登记wire格式。
- “生产IssuerID继续是现有字符串，逐byte匹配；不trim、不折叠大小写”；
  精确选择IssuerID+CA版本，不取latest/最大值、不按Name/SPKI/验签结果选CA。
- 保留(issuer_id,serial_bytes)原唯一性命名空间；不改索引或自动回填/改名/删行。
  本组件没有DB访问，不能声称它验证了真实唯一性或运行写入冲突。
- 上层只保存自有Claims与Summary，不保存浅拷贝Material或借用caller raw。
  Material预算/完整复制由唯一真实lower负责；不声称新层在入口已复制全部raw。
  不支持caller并发修改、不承诺隐藏副本擦除或硬实时取消。
- 新声明预算：Catalog.Entries最多128；每个IssuerID最多128 ASCII bytes，
  Association.NodeID最多36 bytes、SerialBytes最多20；所有位置计费，
  声明字符串/raw合计最多16568 bytes，溢出安全。版本最大9223372036854775807。
  empty/0坐标作为缺失声明进入后续不足分类；超限先于缺失/形状错误。
- 新层metadata预算/形状在lower之前；lower材料错误在声明对应错误之前。
  当前可观察取消/截止优先；所有失败返回nil模型，普通panic清理后继续传播。
- Model浅复制共享Clear；Summary复制全部值并深复制两Evidence slices。
  原始失败回执保留，不把无测试运行、fixture错误或守卫语法错误叫作目标RED。
- 独立最终审查限一次fresh reviewer；修复后重跑完整26矩阵并同reviewer复查。
  不并行实现、不创建用户chat；执行选项改变须由用户选择。
- 本轮只写计划。将来实施阶段也无Git写入/提交/推送/PR/main/cleanup/install/
  安全设置改动；这些仍须独立授权。历史C12/B01/Task10/Task15/I2/I3阻塞不关闭。

## Review Focus

1. 全部伪造的声明可以自洽：Task4 FabricatedClaimsAreNotTrusted必须明确只得声明Proof，
   保留全部五责任、无生产成功标志/工厂；不能用测试自洽建立可信来源。
2. 两张同密钥同Name的CA都能验签：Task3 SameKeyRenewal要求按显式CA版本/
   CADER对应，不任选其一；两种自洽选择都只是条件声明，不证明历史原选版本。
3. 重复/冲突记录藏在非选中位置：Task3 UnselectedDuplicates检查整个声明表；
   完全相同重复也拒绝，不能排序去重或只检查第一次命中。
4. 一致的旧快照可伪装“当前”：Task3 DeclaredHeadMismatch/Task4 CoherentRollbackClaim
   验证坐标不一致拒绝；所有坐标一致倒退也不声称防回退已证明。
   Task4 CoherentChangedCatalogClaim明确同head/digest下改非选中声明仍不能证明
   内容摘要已核验；本层不定义登记编码或真实防回退算法。
5. caller复用slice且半途取消：Task2 OwnedAliasIsolation/Task4 CancellationAndPanic
   验证声明自有、两返回slices隔离、caller不变及完整真实lower生命周期。

---

## 1. 文件与合同决策

25个新Go精确库存，文件均在internal/nodecontrol/nodeissuermapping/：

- 12产品：doc.go input.go errors.go budget.go owner.go owned_claims.go
  owned_summary.go model.go claims.go lower.go projection.go check.go。
- 13测试：input_test.go budget_test.go owned_claims_test.go owned_summary_test.go
  model_test.go fixtures_test.go claims_test.go lower_test.go projection_test.go
  check_test.go boundary_test.go ownership_boundary_test.go consumer_boundary_test.go。

doc合同/预算由Task1负责；声明及证据生命周期Task2；
声明查找/真实lower/对应Task3；公开入口Task4；三守卫及交付Task5。
唯一旧测试文件加新库存=26个格式路径；最终独立命令也恰26，但两计数含义不同。

### 公开类型：input.go（仅声明，不是生产能力）

下列名称、顺序和类型固定；contracts.Digest为既有[32]byte，uuid.UUID为既有类型。
Head坐标只表示声明值，无真实性或antirollback证明。
私有常量名declaredMappingProfile固定为新声明profile，materialReferenceProfile
固定为下层profile；maxRegistrationClaims=128、maxDeclaredRawBytes=16568、
maxDeclaredVersion uint64=9223372036854775807，不引入格式选择器。

~~~go
type Proof uint8
const (
    None Proof = iota
    DeclaredIssuerSerialCorrespondenceConditionalTimeBound
)
type RegistrationState uint8
const (
    UnknownRegistrationState RegistrationState = iota
    DeclaredEligible
    DeclaredWithdrawn
)
type Obligation uint8
const (
    UnknownObligation Obligation = iota
    AuthenticatedRegistryPublicationAndRollbackProtection
    AuthenticatedIssuanceLinkAndRowStatus
    GuardedV7SnapshotAndLease
    CrossRenewalDatabaseUniqueness
    FullIdentityCompositionCommitAndRuntimeAuthorization
)
type SnapshotHead struct {
    Deployment contracts.Digest
    Version uint64
    Digest contracts.Digest
}
type RegistrationClaim struct {
    IssuerID string
    CAVersion uint64
    CADERDigest contracts.Digest
    State RegistrationState
}
type CatalogClaim struct {
    Head SnapshotHead
    Entries []RegistrationClaim
}
type AssociationClaim struct {
    Head SnapshotHead
    CertificateID, IssuanceID uuid.UUID
    NodeID, IssuerID string
    CAVersion uint64
    LeafDER, PublicKey contracts.Digest
    SerialBytes []byte
}
type Claims struct {
    ExpectedHead SnapshotHead
    Catalog CatalogClaim
    Association AssociationClaim
}
type Input struct {
    Material nodeissuerbinding.Input
    Claims Claims
}
type Projection struct {
    Head SnapshotHead
    IssuerID string
    CAVersion uint64
    CertificateID, IssuanceID uuid.UUID
    Local nodeissuerbinding.Projection
}
type Summary struct {
    Proof Proof
    Profile string
    Evidence nodeissuerbinding.Summary
    Proposed Projection
    ObservedAt, ValidUntil time.Time
    Obligations [5]Obligation
}
type Model struct { owned *owner }
~~~

owner.go固定owner {claims Claims; summary Summary; valid bool}。
CheckDeclared(context.Context,Input)(*Model,error)、(*Model).Summary()Summary、
(*Model).Clear()是全部公开函数/方法；不导出复制/选择/lower helper。

errors.go恰五个固定errors.New，前缀nodeissuermapping：
ErrInvalidInput="nodeissuermapping: invalid input"；
ErrLimitExceeded="nodeissuermapping: limit exceeded"；
ErrInvalidEvidence="nodeissuermapping: invalid evidence"；
ErrEvidenceInsufficient="nodeissuermapping: evidence insufficient"；
ErrClaimConflict="nodeissuermapping: claim conflict"。
它们不冒充authority.NotFound/Conflict或可信数据库分类。

### 精确语义

preflight只处理ctx及声明预算/形状，不重新实现lower材料预算。
非空IssuerID必须满足既有ASCII格式；NodeID非空时必须为36-byte，
其实际canonical值在绑定时逐byte与真实lower节点比较；
非零UUID必须RFC4122 variant。未知State>2或版本>MaxInt64属于InvalidInput。
0/空/UnknownState保留为缺失声明，不凭形状检查补值。

validateDeclarations检查ExpectedHead/Catalog.Head/Association.Head全坐标非零并
完全相等。缺字段/空表/任意记录缺IssuerID/版本/CADER/已声明资格为
EvidenceInsufficient；同键重复（即使完全相同）或矛盾head为ClaimConflict。
全表按(issuer字符串,CAVersion)检查重复，不跨版本去重。
精确选Association.IssuerID/CAVersion；无命中为EvidenceInsufficient，不是
可信NotFound。选中DeclaredWithdrawn为ClaimConflict；不自动选其他Eligible版本。
已声明签发UUID/NodeID/LeafDER/PublicKey/serial缺失亦为EvidenceInsufficient。
该helper阶段顺序固定：ctx→三个head完整性→head相等→表非空及全部记录
完整性→全表重复→Association身份/材料字段完整性→精确选中记录→选中资格→
最终ctx。每个阶段的当前可观察取消优先，前一阶段的领域错误优先于后者。

bindProjection先接收validateDeclarations结果，比较选中CADER与真实lower；
再逐项比较NodeID、LeafDER、PublicKey及serial长度/实际bytes。
lower校验保证slice前L在1..20、首非零、未用尾全零。
有值错配为ClaimConflict，成功Local直接等于lower.Proposed，不再派生hash。
不核对真实SQL索引、登记摘要wire或数据库行来源；这些不得在文档中暗示已做。

## 2. 执行前记录（实施时才运行）

仅读获批规格/计划、现有workspace与全量源码基线；自己的忽略scope为
.superpowers/sdd/2026-10-13-declared-issuer-mapping-consistency/。
保存plan-path、每task的完整brief及BASE、进度、原始回执和before/after源hash。
不复用/改写之前增量的私有证据目录。基线预计676个Go/SQL/module文件，
以实际盘点为准；唯一允许旧Go差异为上文一个测试。

Go程序：C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe；
gofmt为同目录gofmt.exe。每命令固定GOOS=windows、GOARCH=amd64、
GOTOOLCHAIN=local、GOPROXY=off、GOSUMDB=off、GOAUTH=off、GOVCS=all:off、
GOFLAGS=-mod=readonly；GOPATH=C:/Users/Lenovo/go，GOMODCACHE其pkg/mod，
GOCACHE仅自己scope/go-cache。记录实际程序、参数、环境、UTC、exit和完整输出。
PowerShell变量issuerMappingGo指向该程序；不覆盖HOME/CODEX_HOME。

## Task 1: 声明API、预算与唯一消费例外

**Files:** Create doc.go/input.go/errors.go/budget.go/owner.go（三字段声明）；Test input_test.go/budget_test.go；
Modify internal/nodecontrol/nodeissuerbinding/consumer_boundary_test.go。
**Interfaces:** 产生公开类型和错误；addRaw(total,size,limit int)(int,bool)，
claimBudget(Claims)error，validateClaimShape(Claims)error（budget.go）、
preflight(context.Context,Input)error。Model在input.go引用owner；
Task1 owner.go提供三字段声明，Task2增加方法，不编造CheckDeclared。

- [x] **Step 1: 写失败测试。** TestIssuerMappingAPI/FixedErrors/Budget/Preflight
  核对类型/字段/enum/profile及五错误；128/129 records、每字符串/raw等长/+1、
 16568总量、溢出/重复位置、nilctx/当前取消、预算先于缺失与非法字段。
  旧测试新增TestIssuerBindingMappingBoundary，验证exact新目录/pkg的.go和_test.go
  真，external_test/subdir/sibling/其他root/伪package/目标子导入假。
~~~go
if ErrEvidenceInsufficient.Error()!="nodeissuermapping: evidence insufficient" { t.Fatal("error") }
if DeclaredEligible!=1 || DeclaredWithdrawn!=2 { t.Fatal("enum") }
// TestIssuerMappingPreflight必须恰此名称，不能用Priority后缀漏出下列选择器。
~~~
- [x] **Step 2: 记录两次RED。** 分别新包和旧精确例外，不接受无运行测试。
~~~powershell
& $issuerMappingGo test ./internal/nodecontrol/nodeissuermapping -run '^TestIssuerMapping(API|FixedErrors|Budget|Preflight)$' -count=1 -p=1 -timeout=120s
& $issuerMappingGo test ./internal/nodecontrol/nodeissuerbinding -run '^TestIssuerBindingMappingBoundary$' -count=1 -p=1 -timeout=120s
~~~
  预期分别缺API/预算和缺exact例外，不把测试自身编译错误混作目标RED。
- [x] **Step 3: 实现本任务合同。** 固定上述字段/预算/错误及私有profile；
  旧issuerBindingConsumerAllowed只新增exact nodeissuermapping目录/package/
  nodeissuerbinding目标import分支，保留全部旧规则；不放宽整包或其他旧consumer。
- [x] **Step 4: 原两命令GREEN。** 两次exit0，预算全部覆盖；规格/旧产品hash不变。
- [x] **Step 5: 本地检查点。** 记录实际diff、RED/GREEN、新固定API，无Git写入。

## Task 2: 声明/证据自有复制与共享模型生命周期

**Files:** Update owner.go；Create owned_claims.go/owned_summary.go/model.go；
Test owned_claims_test.go/owned_summary_test.go/model_test.go。
**Interfaces:** copySlice[T any]([]T)[]T、copyRaw([]byte)[]byte、normalizeTime(time.Time)time.Time；
ownClaims(context.Context,Claims)(Claims,error)、clearClaims(*Claims)；
ownEvidence(nodeissuerbinding.Summary)nodeissuerbinding.Summary、
clearEvidence(*nodeissuerbinding.Summary)；(*owner).clear()、Model两方法。

- [x] **Step 1: 写失败测试。** TestIssuerMappingOwnedClaims/OwnedSummary/Model/
  OwnedAliasIsolation/OwnedFailureEnumeration/NormalizeTime，以完整非零合成字段
  覆盖Entries和SerialBytes两可变输入集合、所有坐标/字符串值、重复backing；
  lower证据两slices位于Evidence.Evidence.Time.Policy.Binding.Proposed.Target.Links
  和Coverage（路径以nodeissuerbinding.Summary为根），返回Summary再增加外层Evidence。
  nil/zero、独立owner、共享Clear、正常/取消/截止/panic均不清caller。
~~~go
copied := *model
model.Clear()
if !reflect.DeepEqual(copied.Summary(),Summary{}) { t.Fatal("shared lifecycle") }
~~~
- [x] **Step 2: RED。**
~~~powershell
& $issuerMappingGo test ./internal/nodecontrol/nodeissuermapping -run '^TestIssuerMapping(Owned|Model|NormalizeTime)' -count=1 -p=1 -timeout=120s
~~~
- [x] **Step 3: 实现私有复制。** ownClaims先脱离两个slice再保护清理，保留空值
  不修正声明；strings仅值复制，不宣称清掉不可变backing。ownEvidence先把两
  slices置nil并保护，再copy。owner.clear清声明/完整Evidence再zero；
  Summary取完整值后脱离Evidence并复制；不存Material或lower Model句柄。
- [x] **Step 4: GREEN。** 完整反射覆盖和reachable backing清理、动态Err点枚举通过，
  不套用旧39/710点；说明Material的复制责任在真实lower而非此阶段。
- [x] **Step 5: 本地检查点。** 记录自有范围/隐藏runtime及并发限制，无Git写入。

## Task 3: 固定材料、全量声明检查、唯一真实lower与投影

**Files:** Create fixtures_test.go/claims.go/claims_test.go/lower.go/lower_test.go/
projection.go/projection_test.go。
**Interfaces:** validateDeclarations(context.Context,Claims)(RegistrationClaim,error)；
stageError(context.Context,error)error、evaluateLower(context.Context,*owner,nodeissuerbinding.Input)error、
validateEvidence(nodeissuerbinding.Summary)error；
bindProjection(context.Context,Claims,nodeissuerbinding.Summary)(Projection,error)；
test-only issuerMappingFixture(testing.TB,string)Input，
assertIssuerMappingLower(testing.TB,Input)nodeissuerbinding.Summary。

- [x] **Step 1: 冻结独立controls。** 一次性AST提取现有nodeissuerbinding固定39份
  Material及manifest，旧raw不变。只在本层增添synthetic声明：固定非零scope/head、
  canonical UUID、显式CA版本、ASCIIissuer及真实serial/DER/SPKI对应。
  冻结完整Input/raw/lower摘要/声明Projection manifest，不调用新CheckDeclared、
  bindProjection或validateDeclarations造golden。真实lower结果和独立X.509/hash
  各自核对；保留原低层正/负例及同key/name续签、同serial另一CA控制。
  普通测试不签发，不读旧源；生成器byte-identical归档.go.txt及hash，保留失败。
  先独立跑TestIssuerMappingFixtureControls取得exit0，再写依赖产品的失败测试。
- [x] **Step 2: 写失败测试。** Claims缺字段/Unknown资格/空表/不存在版本不足；
  DuplicateKeys及UnselectedDuplicates全表含相同/冲突重复；DeclaredHeadMismatch、
  仅Association改变Issuer大小写而无精确命中时为不足，禁止折叠后命中；
  withdrawn无fallback为冲突。LowerErrors/Ownership/Consistency
  测原sentinel、typed concrete type/字段/Unwrap链、即时Clear/完整capture、
  exactproof/profile/正常UTC值/Obs<Until、两层继承时间/Local.Leaf、
  CADER非零及serial首尾/长度；private伪Summary不进入公开Input。
  Projection/SameKeyRenewal/SameSerialDifferentIssuer/SerialEncoding先assert
  真实lower成功再断言声明错配；21-byte DER lower拒绝明确是前置失败。
- [x] **Step 3: 产品RED。**
~~~powershell
& $issuerMappingGo test ./internal/nodecontrol/nodeissuermapping -run '^TestIssuerMapping(Claims|DuplicateKeys|UnselectedDuplicates|DeclaredHeadMismatch|Lower|Projection|SameKeyRenewal|SameSerialDifferentIssuer|SerialEncoding)' -count=1 -p=1 -timeout=120s
~~~
  必须先有正确fixture；这里失败须是未实现helper/目标语义，不是材料不合法。
- [x] **Step 4: 实现claims/lower/projection。** 按§1精确缺失/冲突分类，128²有界
  全表重复检查且每外层循环/final检查ctx；不排序去重/选latest。
  validateDeclarations使用显式ctx并检查每外层循环及返回前取消；bindProjection
  调用validateDeclarations(ctx,claims)。evaluateLower只调用nodeissuerbinding.Evaluate(ctx,material)一次，收到模型
  立即defer Clear；ctx/原err检查后直接捕获o.summary.Evidence=lower.Summary()，
  owner由Task4在调用前保护。validateEvidence先校验固定profile/proof及正常时间、
  完整继承值和canonical serial，再任何slice。bindProjection按§1比较，
  返回Projection.Local=evidence.Proposed和原声明标识，不新hash/改输入。
- [x] **Step 5: 原选择器GREEN。** typed错误不引入更低包import：
  reflect.TypeOf/DeepEqual及errors.Unwrap核对真实两次lower结果，不绕过生产lower。
  当前自洽CA版本仅表明声明对应，不能证明真实签发版本；原raw/旧manifest不变。
- [x] **Step 6: 本地检查点。** 记录39 Material控制/独立golden、真实lower清理、
  非选择项重复拒绝和档案hash；所有可信来源/DB/v7门槛保持开放。

## Task 4: CheckDeclared入口、责任、取消与有限结论

**Files:** Create check.go/check_test.go。
**Interfaces:** CheckDeclared(context.Context,Input)(*Model,error)，private owner及前述helper。
- [x] **Step 1: 写失败测试。** TestIssuerMappingCheckDeclared/TemporalInheritance/
  ErrorPriority/FabricatedClaimsAreNotTrusted/CoherentRollbackClaim/CoherentChangedCatalogClaim/
  CancellationAndPanic。函数签名、fixedprofile/proof、完整Evidence/Projection/五
  责任/原Obs/Until；纳秒新观察独立模型。全部伪造但自洽声明或一致旧head即使
  成功也只得Declared Proof、无生产可信位且责任不消失。改变一个非选中、
  不重复且形状合法的登记声明，但保持自报head/digest一致，也必须明确本层
  没有核验真实登记编码/hash，不能将Declared Proof当内容完整性证明。
  材料过期/typed失败与声明冲突双故障保留原材料错误；metadata预算/形状优先于lower。
  动态真实Err点取消/截止/panic无模型/caller hash不变；返回快照共享Clear受控。
- [x] **Step 2: RED。**
~~~powershell
& $issuerMappingGo test ./internal/nodecontrol/nodeissuermapping -run '^TestIssuerMapping(CheckDeclared|TemporalInheritance|ErrorPriority|FabricatedClaimsAreNotTrusted|CoherentRollbackClaim|CoherentChangedCatalogClaim|CancellationAndPanic)$' -count=1 -p=1 -timeout=120s
~~~
- [x] **Step 3: 实现check.go。** preflight→新owner并登记失败/panic清理→ownClaims→
  evaluateLower(ctx,o,in.Material)→bindProjection(ctx,o.claims,o.summary.Evidence)→
  fixedprofile/Declared Proof/五责任及lower原时间→最终ctx→valid/移交。
  不在owner存Material，不接受Ready/lease/provider/预计算Summary或Lower替身。
  五责任顺序恰§1；来源声明不足不能调用自己制造的生产fallback。
- [x] **Step 4: 选择器与全新包分别GREEN。**
~~~powershell
& $issuerMappingGo test ./internal/nodecontrol/nodeissuermapping -count=1 -p=1 -timeout=120s
~~~
  明确本层没有可信来源前后检查；不将声明head比对称为真实antirollback。
  所有更高生产结果始终不存在。记录实际动态取消点和范围，而非旧点数。
- [x] **Step 5: 本地检查点。** 记录真实结果/所有义务，无生产factory/Git写入。

## Task 5: 三守卫、26最终矩阵、fresh审查与交付

**Files:** Create boundary_test.go/ownership_boundary_test.go/consumer_boundary_test.go；
最终才新增docs/roadmap/2026-10-13-declared-issuer-mapping-consistency-execution.md、
追加docs/roadmap/current-status.md，更新本计划实际checkbox。
**Interfaces:** issuerMappingSources(testing.TB)map[string][]byte；
issuerMappingMutation(testing.TB,map[string][]byte,string,string,string)map[string][]byte；
issuerMappingProductAllowed([]byte)bool；
issuerMappingCapabilityAllowed(map[string][]byte)bool；
issuerMappingOwnershipAllowed(map[string][]byte)bool；
issuerMappingConsumerAllowed(path,pkg,imported string)bool；
issuerMappingASTHash([]byte)string、issuerMappingSchemas()map[string]string。

- [x] **Step 1: 写guard失败测试。** ProductBoundary/CapabilityBoundary/
  OwnershipBoundary/Consumers覆盖25库存/12产品schema/API/三owner字段、
  Claim预算/全表duplicates/显式CA版本/原serialslice、真实lower一次/立即Clear、
  完整证据两slice复制/最后ctx/Declared Proof/五责任。
  每个mutation先证明源bytes确改变；删除/后移保护或选择latest/吞错误/
  新hash/Material句柄外逸/生产标志/额外能力/导出/子目录均拒绝。
  良性格式/local rename/命名import alias通过。新consumer仅exact本目录
  同package _test.go；external_test/production/sibling/subdir/other-root/subimport拒绝。
- [x] **Step 2: 未实现guard RED。**
~~~powershell
& $issuerMappingGo test ./internal/nodecontrol/nodeissuermapping -run '^TestIssuerMapping(ProductBoundary|CapabilityBoundary|OwnershipBoundary|Consumers)$' -count=1 -p=1 -timeout=120s
~~~
- [x] **Step 3: 实现闭合guard，schema初始为空。** 生产成员仅context.Context、
  errors.New、bytes.Clone/Equal、time.Time、uuid.UUID/Nil/RFC4122、
  contracts.Digest、nodeissuerbinding.Input/Summary/Projection/Evaluate/
  SelectedIssuerSerialReferencesConditionalTimeBound。方法仅ctx.Err、
  time.Round(0)/UTC/IsZero/Year/Before、uuid.Variant、真实lower.Summary/Clear及
  固定私有clear。无parser/newhash/append/编码/strings生产归一化、方法值/
  callback能力/shadow/dot/blank/重复import/goroutine/recover。
  全量schema锁定字段/直接调用/控制流，且总数12；声明不能添加可信位。
- [x] **Step 4: schema RED。** 原命令应仅因空集合拒绝12实际产品schemas；
  保存规范化AST摘要，不把白名单/fixture错误记作schemaRED。
- [x] **Step 5: 逐源审计后冻结12认可schema。** 对照原错误/证据capture/预算/
  选择/清理/结果限定及先前RED/GREEN，不能盲目刷新。
- [x] **Step 6: 原guard命令GREEN。** 实际负mutation均拒绝，三个良性控制通过。
- [x] **Step 7: 格式26paths并运行26独立命令。** 精确25新Go+一个旧测试gofmt -l
  空/exit0；必要-w限这些文件。保存独立raw格式回执/前后hash；所有测试同一
  定型Go/SQL/module冻结，忽略scope保留全部失败。
~~~powershell
$mappingFull = @('nodeissuermapping','nodeissuerbinding','nodeleafbinding','nodecerttime','nodeissuer','nodeca','nodeleaf','nodeleafkeys','contracts','serving','statebinding','statecomposition','statecontract','stateformat','statematerials','statepolicy','statepolicybinding','staterecovery','statetime','statetimebinding','stateverify')
foreach ($package in $mappingFull) {
    & $issuerMappingGo test ("./internal/nodecontrol/"+$package) -count=1 -p=1 -timeout=120s
    # 每一项用独立RunId与raw回执执行；非零即停止/调查，不继续当作全矩阵通过。
}
& $issuerMappingGo test ./internal/readiness -count=1 -p=1 -timeout=120s
& $issuerMappingGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1 -p=1 -timeout=120s
& $issuerMappingGo test ./internal/store -run '^$' -count=1 -p=1 -timeout=120s
& $issuerMappingGo test ./internal/nodecontrol/nodeissuer -run '^$' -fuzz '^FuzzIssuerEnvelope$' -fuzztime=5s -parallel=1 -p=1 -timeout=30s
& $issuerMappingGo test ./internal/nodecontrol/nodeissuer -run '^$' -fuzz '^FuzzIssuerPair$' -fuzztime=5s -parallel=1 -p=1 -timeout=30s
~~~
  恰22完整包+authority指定+store仅编译+两fuzz=26；有限counts/实际耗时如实记。
  失败先systematic-debugging，不放宽时限/旧守卫/安全门槛或改旧产品掩盖。
- [x] **Step 8: 一次fresh只读最终质量/安全审查。** 本计划获批允许一位独立
  reviewer按执行时最强可用模型/支持高推理审全部25新文件、一个旧diff、
  spec/plan和raw回执，相对d2829bfa，不计旧组件为新贡献、不派并行实现/新chat。
  C/I/M真实记录，修复需对应回归及完整26重矩阵/同reviewer复查。
  每个declined领域由作者逐项裁定与代价，不能沉默丢掉；作者不替代fresh审查。
- [x] **Step 9: 如实本机交付。** 执行记录写每task、原raw保持、所有拒绝/条件
  结果、格式与26矩阵hash、review/修复及责任；current-status仅追加实际状态。
  检查旧源码/SQL/module/获批spec/HEAD/index不变；仅旧测试授权差异。
  不宣传生产兼容、登记可信、唯一性/v7验收、远端恢复或C12整体完成。
  无Git授权则本机保留并停止待保存。

## 3. 规格覆盖与明确延后门槛

- Spec§§3/6/7的声明坐标、显式CA、byte匹配、旧数据不补值、时限/责任：
  Tasks1–4；真正来源未证明时只能得到声明结果。
- Spec§§4/5/8的真实登记发布、签发关联来源及v7同快照门槛：
  不在本计划接入，Task4/5禁止生产结果/接口，五责任明确保留；
  受控来源/真实读取另有独立规格、计划和验收。
- Spec§7真实数据库唯一冲突与迁移：没有DB行为，不新增模拟“唯一性已验证”
  能力，不改SQL；另行专项授权/真实测试，Task5审查保留其UNKNOWN。
- Spec§9所有与声明组件有关的类由Task1–5测试覆盖；真实DB/provider/lease/
  回调读取/头持久化/撤回安全/生产工厂的集成验收明确延后，不能以fake证明。
- 字段表示、限量、profile是本计划的本地声明API决策，不建立生产wire/
  registry摘要算法或新的部署/数据库命名空间，也不改获批spec。

## 4. 作者自检及执行选择

自检：五task接口/字段一致、25/12库存、26格式/命令计数、精确选择器、
五ReviewFocus都有task测试；预算分层/声明Proof/来源缺口明确。
所有checkbox保持未完成；当前未运行Go、未实施、未进行独立实现审查。
技能默认频繁commit被无Git约束替代本地检查点；代价是未远端保存。

推荐Native顺序：五任务强依赖同一声明ABI与完整下层证据，单作者实现避免
反复交接；末尾一位fresh reviewer保留独立门槛。成本是中途没有独立审查。
可选Subagent-driven：每任务fresh实现/审查、末尾总审，成本更高；
仅用户选该方法后才适用相应工作流，仍不得另建用户chat。

请用户审阅本计划并选择方法。批准规格不等于批准本计划或代码实现；
选择Native/Subagent-driven亦不能替代对本文具体内容的批准。
