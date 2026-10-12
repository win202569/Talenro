# Node Leaf Direct Issuer Relation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现独立 nodeissuer，对自有 CA/leaf 原始材料证明直接公钥、名称、标识及 ECDSA/SHA-256 签名关系，不建立受信来源或运行授权。

**Architecture:** 真实调用 nodeca.Parse 与 nodeleafkeys.Evaluate，随后在同一自有 DER 上检查临时字段、签名表示、issuer 关系和实际验签。共享 owner 只保存四项输入和固定 Summary，全部 lower Material 在返回前清理。原两张六项 Pending、旧产品能力和生产兼容性 UNKNOWN 保持不变。

**Tech Stack:** Go1.26.5、crypto/x509、crypto/ecdsa、SHA-256、有界 DER envelope scanner、Windows PowerShell；无新依赖。

**Spec:** [已批准的直接签发关系设计](../specs/2026-10-12-node-leaf-issuer-relation-design.md)。执行者必须完整阅读 spec 和本计划。

状态：用户于2026-10-12回复“批准”，批准上述书面设计并允许编写本计划。
spec 的“待审草案/本轮”文字是上一阶段历史，不重写需求正文；批准留痕以本段
及用户消息为准。**本计划尚待审阅，执行方法尚未选择；本轮不实施、不测试、
不暂存/提交/推送，不改变旧文件或 Git 配置。**

规划只读基线：branch=`codex/c12-b01-task9-coordinator`，
HEAD=`764dcb8b6b682ecce03fd6de02dbe339c66ff985`，非忽略工作树/暂存区无修改。
设计文件 SHA256=`B7380C6C4633A339BA28D9E25C884F63BE2D54D26DF6DE85BFA3662F711262B7`。
HEAD 是本机文档提交，最新已推送恢复检查点仍为2877f89c；不能假定新设计已在远端。

## Global Constraints

- 只复用既有 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`，不创建或切换 worktree/branch。
- `Leaf.DER`：1–65,536；`Leaf.ExpectedSPKI`：1–512；`CA.DER`：1–65,536；`CA.DeclaredSPKI`：1–8,192；四项 byte 输入合计最多 **139,776 bytes**。
- TrustDomain 仍为1–253-byte小写ASCII DNS声明，每label 1–63、首尾为`[a-z0-9]`、内部仅`[a-z0-9-]`；NodeID非零且RFC4122 variant。超限检查先于空/纯值检查和任何克隆/解析。
- CA.DeclaredSPKI是该CA的subject key，而非其parent key；两种声明公钥首版都必须是canonical named P-256 SPKI，实际有效长度为91 bytes。
- Input恰为`Leaf nodeleaf.Input`、`CA nodeca.Input`；lower keys的DeclaredIssuerSPKI恰用owned CA.DeclaredSPKI，不新增第五项输入、Material/Summary输入、可信flag或callback。
- 旧三包的产品实现、公开API、Proof、全部Pending不变；新Proof恰为`None=0`和`NodeLeafDirectIssuerSignatureBound=1`，输出恰为spec §4五字段Summary，无第三张Pending表。
- leaf `RawIssuer == CA.RawSubject`按完整raw DER逐字比较；leaf AKI == CA SKI == SHA256(exact CA raw subject SPKI)，完整32 bytes；CA AKI仍指向parent，不做subject派生。
- 首版仅接受leaf `ecdsa-with-SHA256`，OID `1.2.840.10045.4.3.2`，inner/outer AlgorithmIdentifier恰为`30 0a 06 08 2a 86 48 ce 3d 04 03 02`，参数省略。
- signature unused-bits byte恰为0；payload为exact DER的两个最小INTEGER，8–72 bytes，content最多33 bytes，`1 <= r,s < ffffffff00000000ffffffffffffffffbce6faada7179e84f3b9cac2fc632551`；不新增low-S政策，不修复输入。
- 实际验签只用同一临时CA的P-256 subject public key及`SHA256(exact owned leaf RawTBSCertificate DER)`，调用`ecdsa.VerifyASN1`；不marshal TBS/Name、不改preimage、不fallback。
- CA自身签名、parent/root/chain/path、可信当前时间、部署purpose、bundle状态/high-water、issuer/serial唯一性、身份/authority、私钥持有、主状态及runtime仍未认证；Task10/B01/C12/main阻塞保留。
- cleanup-before-copy，四buffer独立；失败/取消/正常panic unwind清理，panic向上传播。共享包装Clear；无输入并发修改、并发Summary/Clear、硬实时、常数时间或不可达内存擦除承诺。
- 唯一旧源码例外是spec §7三个boundary_test.go消费者helper及新增回归案例；不得放宽其余guard、旧API、旧产品import/member或原测试数据。
- 产品禁止私钥/随机数/签发、time.Now、文件/网络/数据库/provider、root pool、x509 Verify/CheckSignature*、安装/主状态/serving/runtime、序列化、goroutine、callback、全局可变注入点及泛算法协商。
- Windows操作仅PowerShell；缓存Go1.26.5、离线只读模块，不安装依赖、不改变持久Go/Git配置。有限测试不等于全仓库、race/Linux/DB/Docker/hardware/reboot/PITR/provider验收。
- 执行阶段推荐先用原始证据保存逐任务检查点，完成独立审查后再另行请求Git保存/推送授权；本计划不自行批准提交、推送、main合并、PR或清理。

## Review Focus

- 相同Name/key出现在另一CA：Task4 `TestIssuerCertificateAmbiguity`应接受直接数学关系但保留CA DER差异及全部Pending，不证明可信issuer_id或唯一来源。
- lower更早拒绝与新层错误顺序混淆：Task2 `TestIssuerLowerErrorIdentity`和Task4 `TestIssuerCombinedFailurePriority`分别固定原sentinel身份及仅到达新阶段后的优先级。
- 以重编码或TBS content替代原TBS：Task3 `TestIssuerSignaturePreimage`做不同字节preimage负例；Task5 `TestIssuerProductBoundary`禁止marshal/重建路径，不夸大动态证明。
- overlapping subslice/共享包装Clear误伤另一Evaluate：Task1 `TestIssuerOwnedInput`和Task4 `TestIssuerIndependentLifecycle`固定独立buffer与独立owner，不承诺并发安全。
- 已观察到取消但普通错误/旧成功仍被返回：Task2 `TestIssuerStageError`及Task4 `TestIssuerCancellation`固定context优先、nil Material与最终移交边界；不把耗时短测当硬实时保证。

---

## 1. 文件与跨任务接口

唯一新源码目录`internal/nodecontrol/nodeissuer/`，恰24个Go文件，无新testdata或
永久生成命令。下表为完整inventory；每行的测试文件负责该产品文件及关联流程。

| 产品文件（12） | 职责 | 测试文件（12） |
| --- | --- | --- |
| doc.go | 部分事实与非目标 | fixtures_test.go（全包独立材料） |
| input.go | 两字段Input、Proof、五字段Summary | lower_test.go（两级旧Proof/Pending） |
| errors.go | 六个独立固定sentinel | budget_test.go（错误与预算） |
| budget.go | preflight/纯值规则 | owned_test.go（预算后复制与清理） |
| owned.go | 四buffer与clear、重建owned Input | material_test.go（共享包装） |
| material.go | Summary/Clear | snapshot_test.go（固定描述对应） |
| lower.go | 真实lower调用及描述一致性 | boundary_test.go（全包隔离） |
| snapshot.go | 临时CA/leaf parser与lower对应 | fuzz_test.go（有界输入不变量） |
| der.go | leaf envelope、最小TLV/OID/INTEGER | der_test.go |
| signature.go | AlgorithmIdentifier/r/s政策及VerifyASN1 | signature_test.go |
| relation.go | raw公钥/Name/标识对应 | relation_test.go |
| evaluate.go | 固定阶段和成功owner移交 | evaluate_test.go |

三个旧修改路径仅为：
`internal/nodecontrol/nodeleaf/boundary_test.go`的leafConsumerAllowed（规划行111）、
`internal/nodecontrol/nodeleafkeys/boundary_test.go`的keysConsumerAllowed（行271）、
`internal/nodecontrol/nodeca/boundary_test.go`的guardConsumer（行474）及各自消费者
测试案例。行号仅作定位，执行时以helper名识别，不覆盖整个文件。
除此之外只允许本计划checkbox和最终本增量执行/状态文档的明确更新；需求正文、
旧spec/plan/execution、旧testdata/SQL/config/go.mod/go.sum保持不变。

公开契约按spec §3–4实现，不添加exported类型/字段/函数：

```go
type Input struct { Leaf nodeleaf.Input; CA nodeca.Input }
type Proof uint8 // None=0; NodeLeafDirectIssuerSignatureBound=1
type Summary struct {
    Proof Proof
    Leaf nodeleafkeys.Summary
    CA nodeca.Summary
    LeafTBSDERDigest, LeafIssuerDERDigest contracts.Digest
}
type Material struct { owned *owner }
func Evaluate(context.Context, Input) (*Material, error)
func (*Material) Summary() Summary
func (*Material) Clear()
```

六个error名字和固定消息完全使用spec §6.2，不导出lower error alias。
唯一包级可变值就是这六个error；profile字串用私有const
`issuerRelationProfile = "node-leaf-direct-issuer.p256-sha256.v1"`。
固定bytes用函数内数组/literal或私有string const，不增加可变全局数组/cache。

私有跨任务类型固定如下；certificate/raw类型只在一次Evaluate栈内暂存：

```go
type owner struct {
    leafDER, expectedSPKI, caDER, declaredSPKI []byte
    summary Summary
    valid bool
}
type lowerSummaries struct { ca nodeca.Summary; leaf nodeleafkeys.Summary }
type certificateSnapshots struct { ca, leaf *x509.Certificate }
type derElement struct { tag byte; full, content []byte }
type signedLeaf struct {
    tbs, issuer, subject, spki derElement
    innerAlgorithm, outerAlgorithm, signature derElement
}
```

测试辅助接口均定义在fixtures_test.go，不作为产品API：
`hexBytes(t testing.TB, text string) []byte`、`fixtureInput(t testing.TB) Input`、
`fixtureVariant(t testing.TB, name string) Input`、
`fixtureParts(t testing.TB, der []byte) [][]byte`、
`fixtureTLV(tag byte, content []byte) []byte`、`fixtureJoin(parts [][]byte) []byte`、
`fixtureRewriteLeaf(t testing.TB, in Input, field string, replacement []byte) Input`、
`fixtureRewriteCA(t testing.TB, in Input, field string, replacement []byte) Input`、
`fixtureSetKeyID(t testing.TB, in Input, role string, value []byte) Input`。
rewrite field限certificate/tbs/signature/innerAlgorithm/outerAlgorithm/issuer/subject/
spki/serial/validity/extensions；role限leafSKI/leafAKI/caSKI/caAKI；未知值测试失败，
无自动修复产品输入。每次fixture返回四个独立克隆。

## 2. 执行基线、环境与证据

本节仅规定未来执行，当前没有运行任何Go测试、生成golden或创建cache。
批准计划后再capture真实Git status/index/HEAD、全部非忽略文件长度/hash和旧
testdata清单；规划HEAD/hash不能冒充执行基线。保留所有现有改动。
只读确认`.superpowers/sdd/2026-10-12-node-leaf-issuer-relation/`被忽略且无不明
既有内容后，再以apply_patch建立baseline.json/ledger.md/verify.ps1。存在记录就先
核对归属、续用不重复的attempt编号；不覆盖，不改ignore配置、不读写其他增量台账。

完整PowerShell进程环境如下，禁止替换为系统Go1.25.5或Bash：

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
$env:GOCACHE='D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/.superpowers/sdd/2026-10-12-node-leaf-issuer-relation/go-cache'
$issuerGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
$issuerFmt='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/gofmt.exe'
```

执行目录固定为既有worktree。每次回执记录以上完整wrapper、实际arguments、
workdir、开始/结束UTC、原始完整stdout/stderr、真实exit，以及命令前后source
inventory hashes。只记录白名单环境（含继承GODEBUG），不泄露其他环境凭据；
非空GODEBUG若影响材料规则须停下说明，不静默改全局设置或依赖其默认行为。

原始进程运行中继续等同一session至实际退出；超过120s的test超时或30s fuzz
超时就是失败，不扩大timeout、重复启动来制造成功。每次等待最长60s并给用户
简短更新。输出丢失/截断保留真实诊断，不补造时间、exit或hash；之后重跑是新的
attempt而非补写原记录。文本文件编辑/回执保存用apply_patch，不用shell写入捷径。

RED必须为计划中预期的缺声明或断言行为；fixture语法、错误oracle、环境启动
失败不能冒充行为RED。修复后取得真实RED再GREEN。每任务最后核对旧基线，
更新checkbox和证据检查点，不进行Git bookkeeping写入。新fuzz若生成testdata/
fuzz失败种子，保留文件与失败回执并报告范围变化，不能删掉它来维持24文件清单。

## Task 1: 输入、共享生命周期与三个精确消费例外

**Files:** Create `internal/nodecontrol/nodeissuer/{doc.go,input.go,errors.go,budget.go,owned.go,material.go,budget_test.go,owned_test.go,material_test.go}`；Modify三个§1列出的旧boundary_test.go。

**Interfaces:** Produces `preflight(ctx context.Context,in Input) error`、
`ownInput(in Input) *owner`、`(*owner).input(nodeID uuid.UUID,trustDomain string) Input`、
`(*owner).clear()`及公开Input/Proof/Summary/Material方法、六个error。
本任务尚不声明Evaluate，不能用占位Evaluate返回成功。

- [x] **Step 1: 先写三处旧消费者回归。** 原案例原样保留；新增exact nodeissuer/a.go与a_test.go、规范绝对路径允许，package必须nodeissuer。新增nodeissuer_extra、nested、另一个根下同后缀目录、outside repo、错package、nodeissuer_test、任意兄弟_test.go及nodeca子路径拒绝。

- [x] **Step 2: 记录旧guard真实RED。** Run下列命令，预期新增允许案例被旧guard拒绝：

  ```powershell
  & $issuerGo test ./internal/nodecontrol/nodeleaf ./internal/nodecontrol/nodeleafkeys ./internal/nodecontrol/nodeca -run '^(TestLeafConsumers|TestKeysConsumers|TestCAConsumers)$' -count=1 -p=1 -timeout=120s
  ```

- [x] **Step 3: 仅增加三个批准的exact例外。** 先归一化到各guard已有仓库根/相对路径模型，检查直接目录与package。nodeca的例外只针对exact nodeca import，不能准许target子路径。

- [x] **Step 4: 记录旧guard GREEN。** 重跑Step2，预期exit0，原正反例全部保留且通过。

- [x] **Step 5: 写合同失败测试。** TestIssuerPreflight覆盖nil/已取消/已到期、四字段空/limit/limit+1、UUID/domain各边界、长度优先；TestIssuerFixedErrors逐一比较spec固定文本及不同包sentinel身份。TestIssuerOwnedInput覆盖重叠subslice独立克隆、调用方后改不影响owner。TestIssuerMaterialLifecycle覆盖nil/零值、共享包装、Clear幂等与detached Summary。

  ```go
  in := Input{CA: nodeca.Input{DeclaredSPKI: make([]byte,8193)}}
  if !errors.Is(preflight(context.Background(),in),ErrLimitExceeded) { t.Fatal("budget before empty") }
  o := ownInput(Input{Leaf:nodeleaf.Input{DER:[]byte{1,2},ExpectedSPKI:[]byte{3}},CA:nodeca.Input{DER:[]byte{4},DeclaredSPKI:[]byte{5}}})
  a,b,c,d := o.leafDER,o.expectedSPKI,o.caDER,o.declaredSPKI
  o.valid=true; o.summary.Proof=NodeLeafDirectIssuerSignatureBound
  m:=&Material{owned:o}; copyM:=*m; detached:=m.Summary(); m.Clear()
  if copyM.Summary()!=(Summary{}) || detached.Proof!=NodeLeafDirectIssuerSignatureBound { t.Fatal("lifecycle") }
  for _,buf:=range [][]byte{a,b,c,d} { for _,v:=range buf { if v!=0 { t.Fatal("owned bytes not cleared") } } }
  ```

- [x] **Step 6: 记录合同RED。** Run `& $issuerGo test ./internal/nodecontrol/nodeissuer -run '^(TestIssuerPreflight|TestIssuerFixedErrors|TestIssuerOwnedInput|TestIssuerMaterialLifecycle)$' -count=1 -p=1 -timeout=120s`，预期缺本任务声明。
- [x] **Step 7: 最小实现合同。** 按spec实现；ownInput首次Clone前登记defer清理，四项分别Clone；owner.input只组装自有byte与纯值NodeID/domain，无Clone/parser，无caller byte读取。
- [x] **Step 8: 记录GREEN。** 重跑Step6；再完整新包及旧三包test（同count/p/timeout），预期全部exit0。不把此时的合同测试称为完整Evaluate完成。
- [x] **Step 9: 保存阶段检查点。** 原始RED/GREEN、三个旧guard最小diff、旧文件余部/旧testdata hash及本任务源码留痕；只更新本任务checkbox，不提交/推送。

## Task 2: 冻结测试材料、真实lower组合与快照对应

**Files:** Create `internal/nodecontrol/nodeissuer/{fixtures_test.go,lower.go,snapshot.go,lower_test.go,snapshot_test.go}`。

**Interfaces:** Consumes Task1 owned Input/errors。Produces
`stageError(ctx context.Context,err error) error`、
`evaluateLower(ctx context.Context,in Input) (lowerSummaries,error)`、
`checkLowerProofs(s lowerSummaries) error`、
`parseSnapshots(ctx context.Context,in Input,s lowerSummaries) (certificateSnapshots,error)`。
in必须是Task1重建的owned输入；两类返回值仅是内部临时信息，不能保存到owner。

- [x] **Step 1: 制作并独立冻结golden。** 在专属忽略scratch以apply_patch建立一次性fixture-generate.go，使用标准库Go离线签发，禁止product helper或旧fixture导入。基准测试CA为P-256 scalar=1、leaf scalar=2；CA serial=128、DN CN=`nodeissuer offline CA`、2026-01-01T00:00:00Z至2036-01-01T00:00:00Z、BC=true无pathLen、KU=certSign；leaf serial=129、empty subject、2026-10-12T00:00:00Z至次日、NodeID=`11111111-2222-4333-8444-555555555555`、domain=`node.test`。两签名算法显式ECDSAWithSHA256；CA四扩展和leaf六扩展完全按旧获批profile显式设置，SKI/AKI都是exact SPKI全SHA256，不依赖自动行为。

  一次生成后把DER/SPKI/Name/TBS/signature与独立摘要冻结为fixtures_test.go的hex常量；Windows SHA256另核对原始bytes，不调用产品hash/parser作期望值。基准真实签名用标准库x509.CheckSignatureFrom作为test-only对照，错误不被转换为期望“有效”。普通测试不运行generator、不随机重签。若首次helper失败，保留回执，修正并以新attempt生成，不假称固定bytes早已存在。

  冻结以下fixtureVariant名字：base；caCRL（KU certSign+CRLSign）；caPathZero（显式0）；caReordered/leafReordered（各扩展逆序）；alternateDN（CN=`nodeissuer alternate CA`，leaf issuer同raw）；expired（CA2010-01-01至2011-01-01，leaf2010-10-12一日）；future（CA2040-01-01至2050-01-01，leaf2041-10-12一日）；crossCAWindow（base leaf，CA2020-01-01至2021-01-01）；sameCAKeyAndNameOtherCert（base CA改serial=130与parent AKI32bytes全0x42）；sameKeyRoles（leaf scalar=1、相应SPKI/SKI改变）；highS/lowS（同一base leaf的s/N-s两个正确DER表示，TBS不变）。变体仍真实签发leaf；CA自身变化可保留旧签名，因本层不认证CA parent。变体含义按其名字固定，不读取time.Now。

- [x] **Step 2: 写lower与快照失败测试。** TestIssuerLowerProofs逐一翻转三Proof、两个Pending任一条/顺序，须ErrInvalidMaterial。TestIssuerLowerErrorIdentity：CA malformed+leaf malformed先返回nodeca.ErrInvalidMaterial；CA合规+leaf malformed返回原nodeleaf.ErrInvalidMaterial；leaf AKI32bytes错误返回原nodeleafkeys.ErrBindingMismatch而非本层relation error；CA错误声明key返回原nodeca.ErrBindingMismatch。TestIssuerStageError对已取消ctx与普通错误返回context.Canceled，对正常ctx保留同一个error。

  TestIssuerSnapshots对真实lower结果核对所有字段；内部单元控制逐一改变CA四摘要、两ID、serial/长度、时间、pathLen标志/值和CRLSignAllowed，以及leaf DER/SPKI/声明issuer摘要、SKI/AKI、serial/长度、时间、NodeID/domain，每项必须ErrInvalidMaterial。公开入口不接受这些伪造Summary。

  ```go
  in:=fixtureInput(t); o:=ownInput(in); defer o.clear()
  owned:=o.input(in.Leaf.NodeID,in.Leaf.TrustDomain)
  s,err:=evaluateLower(context.Background(),owned); if err!=nil { t.Fatal(err) }
  if err:=checkLowerProofs(s); err!=nil { t.Fatal(err) }
  broken:=s; broken.ca.CADERDigest[0]^=1
  if _,err:=parseSnapshots(context.Background(),owned,broken); !errors.Is(err,ErrInvalidMaterial) { t.Fatal("snapshot") }
  ```

- [x] **Step 3: 记录RED。** Run `& $issuerGo test ./internal/nodecontrol/nodeissuer -run '^(TestIssuerFixtureControls|TestIssuerLowerProofs|TestIssuerLowerErrorIdentity|TestIssuerStageError|TestIssuerSnapshots)$' -count=1 -p=1 -timeout=120s`，预期缺真实lower/snapshot helper而失败。fixture controls须自身有效，不把生成或解码错误当行为RED。
- [x] **Step 4: 实现真实lower与字段对应。** evaluateLower按CA后keys顺序真实调用、每次立即defer Clear、各阶段context优先；keys的DeclaredIssuerSPKI仅用owned CA声明。checkLowerProofs按spec §4原数组逐字比较。parseSnapshots按CA后leaf分别ParseCertificate并在前后检查取消，Raw完整对应；固定摘要、ID、serial、UTC时间与lower对应，CA pathLen映射要求省略=-1/Zero=false、显式0=0/Zero=true、正值=对应整数/Zero=false；CRLSignAllowed对应KeyUsageCRLSign。只检查快照描述，不重实现旧扩展profile，不调用DER accessor，不保留parser对象到owner。
- [x] **Step 5: 记录GREEN。** 重跑Step3及完整新包/旧三包，预期exit0；冻结generator源、golden manifest和独立摘要。两层临时副本清理由源码流和旧Clear回归支持，不声称动态观察失败私有owner擦除。
- [x] **Step 6: 保存阶段检查点。** 完整回执/前后hash/基线核对；除Task1三个guard例外没有旧source drift，不提交。

## Task 3: leaf envelope、签名表示与精确验签原语

**Files:** Create `internal/nodecontrol/nodeissuer/{der.go,signature.go,der_test.go,signature_test.go}`。

**Interfaces:** Produces
`readElement(raw []byte) (derElement,[]byte,error)`、
`singleElement(raw []byte,tag byte) (derElement,error)`、
`takeElement(raw *[]byte,tag byte) (derElement,error)`、
`minimalInteger(raw []byte) bool`、`minimalOID(raw []byte) bool`、
`parseLeafEnvelope(der []byte) (signedLeaf,error)`、
`checkLeafSignatureProfile(cert *x509.Certificate,raw signedLeaf) error`、
`checkSignatureDER(sig []byte) error`、
`verifyLeafSignature(ca *x509.Certificate,raw signedLeaf) error`。
profile检查负责raw/parser对应及算法/表示；verify只做实际key/hash/验签。

- [x] **Step 1: 冻结公开独立向量。** 从spec §8.1取Qx/Qy/r/s与sample消息摘要；完整SPKI是下列91bytes，DER签名为72bytes。N-s独立数学值在规划时用Windows BigInteger核对；它不是新的验签结果。

  ```text
  SPKI = 3059301306072a8648ce3d020106082a8648ce3d0301070342000460fed4ba255a9d31c961eb74c6356d68c049b8923b61fa6ce669622e60f29fb67903fe1008b8bc99a41ae9e95628bc64f2f1b20c2d7e9f5177a3c294d4462299
  signature = 3046022100efd48b2aacb6a8fd1140dd9cd45e81d69d2c877b56aaf991c34d0ea84eaf3716022100f7cb1c942d657c41d436c7a1b6e29f65f3e900dbb9aff4064dc4ab2f843acda8
  N-s = 0834e36ad29a83bf2bc9385e491d6099c8fdf9d1ed67aa7ea5f51f93782857a9
  ```

  公开向量仅测试原语。基准项目pair/TBS才是verifyLeafSignature端到端正例，
  不把ASCII sample包装成certificate成功。

- [x] **Step 2: 写失败测试。** TestIssuerDERTLV固定最短长度、截断/溢出/indefinite/high-tag及cursor；TestIssuerEnvelope完整v3字段、外层三字段、unique IDs/尾随/多字段拒绝及所有raw范围。TestIssuerSignatureProfile、TestIssuerSignatureDER按下表固定错误；测试同时覆盖直接helper和公开阶段可到达的输入，lower/parser先拒绝的仍用原error。

  | 输入/控制 | 本层helper期望 |
  | --- | --- |
  | exact无参数OID，parser ECDSAWithSHA256，raw范围对应 | nil |
  | exact OID但手工控制parser枚举不对应 | ErrInvalidMaterial |
  | inner/outer不同、range与parser不对应 | ErrInvalidMaterial |
  | 正确NULL `05 00`或相同完整SHA1/SHA224/SHA384/SHA512/RSA/Ed25519/DSA/Unknown OID | ErrUnsupportedSignatureProfile |
  | NULL带payload、缺OID、多参数、未终止/冗余base128 OID、param TLV截断/非最小length | ErrInvalidMaterial |
  | opaque完整非NULL参数TLV | ErrUnsupportedSignatureProfile；不检查未知内部schema |
  | signature unused!=0、空/截断、额外INTEGER/尾随、冗余sign zero、错误tag、P1363不是DER | ErrInvalidMaterial |
  | canonical r/s为0、负数、N、N+1或超范围 | ErrUnsupportedSignatureProfile |
  | r/s为1或N-1、必要padding、8/72bytes合法边界 | 表示nil；不声称数学签名有效 |
  | 非最小负数或超范围同时存在尾随 | ErrInvalidMaterial，先完整表示检查 |

  TestIssuerSignaturePrimitive断言公开r/s、s替为N-s后均有效，bit flip、换key、
  换hash无效；公开消息摘要固定，不用产品函数产生期望。TestIssuerSignaturePreimage
  对项目base raw.tbs.full成功，单元控制替为整张leaf或仅TBS content须ErrInvalidSignature。

  ```go
  in:=fixtureInput(t); s,err:=evaluateLower(context.Background(),in); if err!=nil { t.Fatal(err) }
  certs,err:=parseSnapshots(context.Background(),in,s); if err!=nil { t.Fatal(err) }
  raw,err:=parseLeafEnvelope(in.Leaf.DER); if err!=nil { t.Fatal(err) }
  if err:=checkLeafSignatureProfile(certs.leaf,raw); err!=nil { t.Fatal(err) }
  if err:=verifyLeafSignature(certs.ca,raw); err!=nil { t.Fatal(err) }
  wrong:=raw; wrong.tbs.full=raw.tbs.content
  if err:=verifyLeafSignature(certs.ca,wrong); !errors.Is(err,ErrInvalidSignature) { t.Fatal("exact TBS preimage") }
  ```

- [x] **Step 3: 记录RED。** Run `& $issuerGo test ./internal/nodecontrol/nodeissuer -run '^(TestIssuerDERTLV|TestIssuerEnvelope|TestIssuerSignatureProfile|TestIssuerSignatureDER|TestIssuerSignaturePrimitive|TestIssuerSignaturePreimage)$' -count=1 -p=1 -timeout=120s`；预期缺helper或预期行为失败。
- [x] **Step 4: 实现最小scanner与运算。** TLV累乘前以真实remaining约束，读单byte tag、完整最短length；固定TBS v3字段次序，extensions边界消费但不复制旧extension规则。AlgorithmIdentifier只做spec §5.3已冻结的外层/OID/NULL规则，不扩大opaque schema。先完整消费r/s两个INTEGER及最小编码，再按r后s范围；最小负数属于unsupported，非最小表示优先malformed，长度上限不能遮盖可见尾随。r/s以去必要padding后的fixed32bytes与固定N逐字序比较；没有产品math/big或私钥。verify只取ca.PublicKey真实*ecdsa.PublicKey、P256/坐标/点检查（异常ErrInvalidMaterial），SHA256(raw.tbs.full)，VerifyASN1 false为ErrInvalidSignature，允许high-S、不fallback。
- [x] **Step 5: 记录GREEN。** 重跑Step3及完整新包/旧三包，均exit0；确认所有error纯固定，公开vector“有效”不是由被测函数动态反推。
- [x] **Step 6: 保存阶段检查点。** 新scanner/signature原始RED/GREEN及golden/源码hash；不把此任务独立helper通过称为完整issuer认证，不提交。

## Task 4: 直接关系与完整Evaluate组合

**Files:** Create `internal/nodecontrol/nodeissuer/{relation.go,evaluate.go,relation_test.go,evaluate_test.go}`。

**Interfaces:** Consumes Tasks1–3接口。Produces
`checkIssuerRelation(in Input,s lowerSummaries,certs certificateSnapshots) error`及
`Evaluate(ctx context.Context,in Input) (*Material,error)`。
关系只比较raw公钥、Name、32-byte标识/派生与真实lower；不验CA自身签名、parent
AKI、不检查chain/time/status。所有使用Input的内部helper调用均为owned输入。

- [x] **Step 1: 写关系与公开组合失败测试。** TestIssuerRelation逐一控制公钥、Name、标识不同，顺序key→Name→ID；内部伪造certs/s只能用于helper单元控制。TestIssuerEvaluateGolden遍历Task2全部正例variant，要求新Proof、两级leaf旧Proof、CA旧Proof、spec两张完整Pending、固定raw摘要/serial/time/pathLen/ID，Clear后所有Proof为None；生产兼容性无新声明。

  ```go
  m,err:=Evaluate(context.Background(),fixtureInput(t)); if err!=nil { t.Fatal(err) }; defer m.Clear()
  got:=m.Summary()
  if got.Proof!=NodeLeafDirectIssuerSignatureBound || got.Leaf.Proof!=nodeleafkeys.NodeLeafDeclaredKeyIdentifiersBound || got.Leaf.Leaf.Proof!=nodeleaf.NodeLeafMaterialShapeValid || got.CA.Proof!=nodeca.NodeCAMaterialShapeAndDeclaredKeyBound { t.Fatal("partial proof") }
  if got.Leaf.Leaf.Pending!=([6]nodeleaf.Obligation{nodeleaf.LeafSignatureAndIssuerTrust,nodeleaf.KeyIdentifierDerivationAndIssuerBinding,nodeleaf.TrustedCurrentTime,nodeleaf.IdentityAndAuthorityProvenance,nodeleaf.PrivateKeyPossession,nodeleaf.MainStateAndRuntimeBinding}) { t.Fatal("leaf pending") }
  if got.CA.Pending!=([6]nodeca.Obligation{nodeca.CertificateSignatureAndIssuerChain,nodeca.AuthorityIdentifierAndParentBinding,nodeca.TrustedCurrentTime,nodeca.DeploymentAuthorizationAndPurpose,nodeca.BundleStatusAndRollback,nodeca.LeafIssuerRelationAndRuntimeUse}) { t.Fatal("CA pending") }
  ```

  TestIssuerEvaluateRejects：改leaf形状合法serial/TBS但保留签名、错key的合法r/s签名→ErrInvalidSignature；CA同key但改raw subject Name、文本相同不同String tag或RDN bytes→ErrIssuerRelationMismatch；坏ID/声明SPKI→真实lower原error。错误均返回nil Material，不泄露动态文本。

- [x] **Step 2: 写组合优先级与不升级测试。** TestIssuerCombinedFailurePriority固定CA错误先于leaf、leaf lower错误先于新signature、表示错误先于Name、Name先于数学错签名；Name/ID双错Name-first仅对关系helper成立，公开lower坏ID先拒绝。TestIssuerCertificateAmbiguity对sameCAKeyAndNameOtherCert通过、CADigest不同且全部Pending；TestIssuerUntrustedFactsRetained对CA自身signature bit flip保持shape、CA AKI任意32bytes、expired/future/crossCAWindow/sameKeyRoles通过但Pending完整，固定测试参考日2026-10-12不是产品可信时间。

  TestIssuerIndependentLifecycle创建两次Evaluate，清理第一次不改第二次或caller
  inputs；TestIssuerCancellation覆盖nil/预取消/预到期，内部stageError取消优先，
  有界工作中取消不返回有效半成品（允许在实际取消前正常返回的竞态单独记为
  未观察取消，不能强行当失败/通过），最终ctx检查与移交由Task5源码guard固定。

- [x] **Step 3: 记录RED。** Run `& $issuerGo test ./internal/nodecontrol/nodeissuer -run '^(TestIssuerRelation|TestIssuerEvaluateGolden|TestIssuerEvaluateRejects|TestIssuerCombinedFailurePriority|TestIssuerCertificateAmbiguity|TestIssuerUntrustedFactsRetained|TestIssuerIndependentLifecycle|TestIssuerCancellation)$' -count=1 -p=1 -timeout=120s`，预期缺Evaluate/关系行为失败。
- [x] **Step 4: 实现固定组合。** Evaluate严格按spec §6.1：preflight、ownInput立即defer失败清理、owned Input、evaluateLower、parseSnapshots、parseLeafEnvelope/profile、checkIssuerRelation、verifyLeafSignature、固定Summary、最终ctx检查、valid/唯一transfer、返回Material。每个阶段前后ctx检查；lower全部在其helper返回前Clear。关系按exact bytes与完整hash核对，不以digest替代Name/key，不错误比较CA parent AKI；Summary只取真实lower及SHA256(raw.tbs.full/raw.issuer.full)。panic不recover成success，owner不留parser/key/raw引用。
- [x] **Step 5: 记录GREEN。** 重跑Step3及完整新包/旧三包，均exit0；清理共享owner可达bytes的观察证据与失败不可达内存限制分开。
- [x] **Step 6: 保存阶段检查点。** 全部组合控制的RED/GREEN/固定error优先级/源码hash；记录未证明责任，不提交。

## Task 5: 全边界守卫、限定验证、一次独立审查与收尾

**Files:** Create `internal/nodecontrol/nodeissuer/{boundary_test.go,fuzz_test.go}`；
完成后Create `docs/roadmap/2026-10-12-node-leaf-issuer-relation-execution.md`，
仅在审查及限定验证通过后更新`docs/roadmap/current-status.md`顶部本增量状态和
本计划checkbox。旧设计状态/需求正文不重写，审批留痕在本计划与执行记录中。

**Interfaces:** Test-only guard入口固定为
`issuerProductAllowed(source []byte) bool`、
`issuerCapabilityAllowed(sources map[string][]byte) bool`、
`issuerConsumerAllowed(path,pkg,imported string) bool`、
`issuerOwnershipAllowed(sources map[string][]byte) bool`。
Fuzz入口固定`FuzzIssuerEnvelope(*testing.F)`、`FuzzIssuerPair(*testing.F)`。
本任务不增加产品接口或更强Proof。

- [x] **Step 1: 写guard失败测试及其拒绝控制。** TestIssuerProductBoundary、TestIssuerCapabilityBoundary、TestIssuerConsumers、TestIssuerOwnershipFlow固定24文件、全部公开契约/六error字串/profile、owner四buffer/Summary/valid，以及下面的完整allowed imports/members；额外export、mutable包级值、DER accessor、Material/Summary输入、lower清理延迟/省略、先Clone后defer、错误Clear目标、提前valid/transfer、caller byte后读及无最终ctx检查都拒绝。原source不被故意破坏，拒绝控制使用内存变体；尚未实现checker时缺声明或拒绝行为失败构成RED。

  | Product import | 唯一允许package members |
  | --- | --- |
  | bytes | Clone Equal Compare |
  | context | Context |
  | errors | New |
  | crypto/sha256 | Sum256 |
  | crypto/ecdsa | PublicKey VerifyASN1 |
  | crypto/elliptic | P256 |
  | crypto/x509 | Certificate ParseCertificate ECDSAWithSHA256 KeyUsageCRLSign |
  | github.com/google/uuid | UUID Nil RFC4122 |
  | talenro.local/platform/internal/nodecontrol/contracts | Digest |
  | talenro.local/platform/internal/nodecontrol/nodeleaf | Input Summary Proof Obligation None NodeLeafMaterialShapeValid及spec §4原六项Obligation名 |
  | talenro.local/platform/internal/nodecontrol/nodeleafkeys | Input Summary Proof None NodeLeafDeclaredKeyIdentifiersBound Evaluate |
  | talenro.local/platform/internal/nodecontrol/nodeca | Input Summary Proof Obligation None NodeCAMaterialShapeAndDeclaredKeyBound Parse及spec §4原六项Obligation名 |

  普通receiver方法只准本任务需要的ctx.Err、lower.Summary/Clear、serial.Sign/Bytes、
  time.UTC/Equal、Curve.IsOnCurve、Material/owner固定方法；不能借method-value绕过
  x509 Verify/CheckSignature*、DER、Marshal/Ready/Authorize/State/Commit/I/O。
  标准库或lower显式alias允许；点/空白alias、重复import/alias、局部definition/range/
  type/field遮蔽alias、伪装receiver selector、非原语VerifyASN1/Verify方法值拒绝。
  验签原语只允许signature.go的精确crypto/ecdsa.VerifyASN1调用，不能变成通用
  callback。guard不声称完整Go语义或形式证明。

  新consumer仅准仓库exact nodeissuer目录同package的_test.go；产品和兄弟/
  nested/outside/外部test package全部拒绝；下层三个新例外与所有原案例一起回归。
  用内存拒绝控制和正常source对照约束checker，不为guard专门扩权。

- [x] **Step 2: 记录guard RED。** Run `& $issuerGo test ./internal/nodecontrol/nodeissuer -run '^(TestIssuerProductBoundary|TestIssuerCapabilityBoundary|TestIssuerConsumers|TestIssuerOwnershipFlow)$' -count=1 -p=1 -timeout=120s`，预期缺checker或固定拒绝控制失败；源代码不作故意破坏。

- [x] **Step 3: 实现test-only checker和两个fuzz入口。** 四个checker按Step1完整闭包实现。Envelope fuzz callback为`func(t *testing.T,der []byte)`，输入最多65,536bytes，分别检查envelope/signature helper的无panic/固定error不变量；seeds含golden certificate、合法signature、非最小length、错unused bits、截断。Pair callback为`func(t *testing.T,leafDER,caDER,expectedSPKI,declaredSPKI []byte)`，分别限65,536/65,536/512/8,192bytes，固定合法NodeID/domain，再clone四输入调用Evaluate；seeds含base及错signature/Name/AKI。成功必须断言全Proof/Pending且Clear归零，failure必须nil Material；不recover panic、不写产品source。

- [x] **Step 4: 记录guard GREEN。** 重跑Step2及完整新包test，预期exit0；normal source允许、全部内存变体正确拒绝，两fuzz seed corpus均运行，不声称已经完成短fuzz探索。

- [x] **Step 5: 格式化仅本次27个Go路径并冻结。** 24个新文件和三个旧guard exact名单，Ordinal排序；`& $issuerFmt -w @issuerSourcePaths`后用`-l`要求exit0且无输出。复核三旧文件的其余guards/API/案例不变、所有旧testdata/源码基线不变；冻结final-sources-1.json包含exact bytes长度/hash。格式错误不扩展为改旧生成脚本/文档。此冻结之后的任何source改动使验证与审查失效。

- [x] **Step 6: 完整限定矩阵。** 使用§2完整环境，逐条记录原始回执与同一冻结revision的前后hash；下面22条命令必须各自exit0，fuzz同时报告实际executions/seed数，不用总elapsed冒充fuzztime。18个完整包为原17加nodeissuer。

  ```powershell
  & $issuerGo test ./internal/nodecontrol/nodeissuer -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/nodeca -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/nodeleaf -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/nodeleafkeys -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/contracts -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/serving -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/statebinding -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/statecomposition -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/statecontract -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/stateformat -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/statematerials -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/statepolicy -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/statepolicybinding -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/staterecovery -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/statetime -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/statetimebinding -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/stateverify -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/readiness -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/store -run '^$' -count=1 -p=1 -timeout=120s
  & $issuerGo test ./internal/nodecontrol/nodeissuer -run '^$' -fuzz '^FuzzIssuerEnvelope$' -count=1 -p=1 -parallel=1 -fuzztime=5s -timeout=30s
  & $issuerGo test ./internal/nodecontrol/nodeissuer -run '^$' -fuzz '^FuzzIssuerPair$' -count=1 -p=1 -parallel=1 -fuzztime=5s -timeout=30s
  ```

  上面共22条：18完整包、1指定authority、1store编译、2fuzz。
  store不跑测试、authority不启用integration tag，不称整仓。
  最终保存新包与三guard的exact差异空白检查；旧历史空白/格式欠账不能说通过。

- [x] **Step 7: fresh独立质量/安全审查。** Native选择下，由作者实现五项后，派一个全新上下文reviewer，使用执行时可用最强审查模型（目前gpt-6-astra/xhigh），fork_turns=none、只读，提供完整spec/plan、真实基线、全部freeze/source和原始证据。不得用旧CA reviewer结论代替新组合层审查。审查覆盖原TBS、公钥/Name/ID、算法/sentinel顺序、完整Pending、隔离例外、清理/取消与所有保全项；无新source写入或外部副作用授权。

  Critical/Important必须修复并取得真实RED/GREEN，source改变就重冻结、重跑
  上面全部22条及format，回到同一reviewer复核；Minor修复或列出具体风险请用户
  决定，不能私自跳过。不自动新开reviewer来替换不利结论，模型不可用不静默降级。
  Subagent-driven若被用户选择则按其技能逐任务实现/审查，最终仍对整增量检查；
  未选择执行方法之前禁止派implementer或reviewer。

- [x] **Step 8: 最终保全与正式执行记录。** 待审查无未决项、最终命令均通过且同freeze后，用apply_patch写执行文档与current-status顶部增量；记录实际审批/UTC、每次attempt真实状态、审查逐项处置、修订、最终源码hash和全部限制。保留旧CA/leaf/progress文字为历史，不把partial改为Trusted/Ready或C12完成。正式文档完成后核对只有24新Go、三个guard最小变更、允许的plan/status/execution变更；无新增consumer、旧testdata/source漂移、Git config/HEAD/index变化。若某命令失败或review缺失，写真实未验收状态，不勾该任务完成。
- [x] **Step 9: 用户交付与下一保存门。** 报告本增量实际完成范围、所有未决责任与失败/证据限制；Native执行技能要求的Ruling/cost清单按当时技能完整列出。源码收尾之后，暂存/提交/普通开发分支推送另行请求明确授权；不自动合并main、PR、强推、删除分支/worktree或迁移provider。

## 3. 覆盖、自检与执行方法交接

| spec要求 | 所属任务/固定检查 |
| --- | --- |
| §1–4目标、原输入/Proof/Pending、budget/API | Task1合同；Task2真实lower；Task4完整结果；Task5固定API |
| §5.1同一原始材料与所有描述对应 | Task2快照；Task3raw envelope；Task5源码流 |
| §5.2公钥/Name/ID、同Name/key不等唯一issuer | Task4关系/ambiguity/untrusted facts |
| §5.3 OID/参数、signature DER/r/s、high-S | Task3固定错误表与公开oracle；Task4golden high/low-S |
| §5.4原TBS/实际key/VerifyASN1、不验CA/chain/time | Task3preimage；Task4保留未决；Task5产品guard |
| §6.1–6.3确定阶段、原error身份、取消与清理 | Tasks1–4对应单元/组合；Task5ownership静态控制 |
| §7三个exact例外/全部旧边界 | Task1旧守卫正反例；Task5consumer/import/inventory |
| §8.1–8.3独立向量/拒绝/不升级/fuzz/回归/独立审查 | Task2golden；Task3RFC oracle；Tasks4–5完整验收 |
| §9阶段/保全/不推送合并 | 当前文档仅规划；Task5正式保全与独立Git门 |

作者自检：逐项覆盖上表；接口名/类型与任务消费一致；五项Review Focus均有
具名测试；不存在产品占位实现、开放profile或待执行者猜测的未知类型；golden
生成的具体输出必须在执行时冻结/独立核对，当前没有伪造future golden hashes。
有限AST/fuzz和作者自检都不是正式证明或独立代码审查。

推荐执行方式为 **Native**：五项任务紧密依赖同一owned材料及固定错误顺序，
在本会话由作者顺序实现能减少接口传递成本，最后一个fresh最强模型审查整增量。
另可选择Subagent-driven，逐任务fresh实现/审查，成本更高但检查更早。
两种方法都不能跳过最终审查、失败修复或保全检查；用户尚未选择，当前不执行。

本轮只新增本计划并做文档/只读保全核对，不更新旧spec/current-status、不生成
fixture、运行Go测试、建立实施scratch、修改source、暂存/提交/推送、合并或清理。
用户须先审阅本计划，并选择执行方法；批准书面设计不等于此计划已经获准执行。
