# Node CA Material Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增隔离的 nodeca，检查 owned CA 材料子集、与声明 subject SPKI 的 exact 对应及固定 32-byte SKI 派生，不建立信任或运行授权。

**Architecture:** 独立包只依赖标准库与 contracts.Digest，不导入 nodeleaf/nodeleafkeys。固定检查顺序为预算、所有权、声明 SPKI、Go certificate parser/raw scanner、CA 字段/四扩展、SPKI/SKI 绑定、partial Summary；共享 owner 保持统一 Clear 生命周期。未知签名算法、AKI parent 对应、可信时间、来源/用途、bundle 状态和 leaf/runtime 关系仍未完成。

**Tech Stack:** 固定 Go 1.26.5、crypto/x509、ECDSA P-256、SHA-256、有限 DER scanner、Windows PowerShell；不增加依赖。

**Spec:** [已批准的 CA 材料设计](../specs/2026-10-11-node-ca-material-design.md)，执行者必须完整阅读 spec 和本计划。

状态：用户于 2026-10-11 回答“1”，批准本计划并选择 Native；五项任务完成，限定最终五命令全部 exit0，fresh 只读整增量审查无待修复项，按部分材料层本机交付（当地 2026-10-12）。此前“本轮仅规划”等段落为规划阶段历史描述；执行记录见 docs/roadmap/2026-10-11-node-ca-material-execution.md。无 Git 写入授权，六项 CA Pending、旧证明等级与整体验收阻塞保留。

## Global Constraints

- 只在既有 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`、branch=`codex/c12-b01-task9-coordinator` 实施，不创建/切换 worktree。
- `DER []byte` 为 1–65,536 bytes；`DeclaredSPKI []byte` 为 1–8,192 bytes；输入总预算 73,728 bytes，不是总堆内存上限。预算检查先于任何克隆或解析。
- 声明 SPKI 是 CA subject key，不是 CA parent key；首版有效 canonical P-256 SPKI 恰好 91 bytes，不支持其他 key 算法或表示。
- 一个完整 X.509 v3 CA；nonempty issuer/subject raw DN 各最多 4,096 bytes；不规定文本归一化或命名授权；禁止 unique identifiers、拼接及尾随。
- 最小正数 serial 编码 content 最多 20 octets，包含 sign padding；时间为 UTC 整秒、1–9999 年，1950–2049 用 UTCTime、其他年份用 GeneralizedTime，NotBefore < NotAfter；CA 不套 leaf 的 24-hour/backdate 规则。
- 恰好 BC/KU/SKI/AKI 四种不同扩展；BC/KU critical=true，SKI/AKI critical=false 且 false 省略；未知、额外、重复扩展拒绝。扩展顺序不影响普通错误优先级。
- BC 为 cA=true，pathLen 省略或 0–255；KU 内层恰为 `03 02 02 04` 或 `03 02 01 06`；SKI 内层 `04 20`+32 bytes，AKI 内层 `30 22 80 20`+32 bytes，仅 `[0]`。
- `CA_SKI = SHA256(exact owned CA RawSubjectPublicKeyInfo DER)`，完整 32 bytes；不截断、不换 preimage、不依赖 Go 自动标识或 GODEBUG、不修改既有 CA/fixture。
- CA signature BIT STRING 非空且 unused bits=0，内外 AlgorithmIdentifier 逐字一致；不实施验签或 signer 算法许可政策，parser 映射为 Unknown 不等于可信。
- 保留 spec §3 固定六项 CA Pending 及顺序；不替换旧两层的六项 Pending，不升级旧 CA compatibility UNKNOWN。
- Parse 只解析 owned 输入；cleanup-before-copy；失败/取消/正常 panic unwind 清理，panic 向上传播。不同包装共享 owner；无并发修改输入或并发 Clear/read 保证，无硬实时取消承诺。
- 公开入口只有 Parse、Material.Summary/Clear 和固定值类型/五个 sentinel；无 DER/SPKI 输出、Verify、Ready、Authorize、序列化或安装接口。
- 不修改任何旧源码、测试、testdata、SQL/迁移、配置或公开接口；不放宽旧守卫。不增加 product consumer、provider、I/O、私钥、随机数、时钟、goroutine、系统 roots 或 runtime 接入。
- 仅本机交付；无 Git 配置/索引/HEAD 写入、暂存、提交、推送、合并或清理。本计划的阶段留痕替代技能模板中的频繁 commit 步骤。

## Review Focus

- 声明 CA subject key 被误当 parent key：Task 4 的 TestCAPartialProofParentRoles 必须证明任意合形状 AKI 不执行 subject 派生，实际 parent/签名仍未认证。
- pathLen 省略被折叠为零：Task 3 的 TestCAPathLenPresence 固定 HasPathLen=false/0 与 true/0 的不同结果。
- parser 接受但 raw 值未完整消费：Task 2 的 TestCARawNestedTails 与 Task 3 的 TestCAExtensionCanonicality 覆盖 DN、BC、KU、SKI、AKI 的嵌套尾随/非最小表示。
- Go 前置拒绝被改写成后置 profile 类别：Task 4 的 TestCAParserPriority 与 TestCACombinedFailurePriority 固定完整检查顺序及不泄露动态错误。
- 清理/包装别名与取消产生旧有效结果：Task 1 的 TestCAMaterialLifecycle 及 Task 4 的 TestCACancellation 固定 shared owner、detached 值及失败无半成品。

---

## 1. 文件与类型接口

唯一新源码目录 `internal/nodecontrol/nodeca/`，24 个 Go 文件：

| 产品文件（12） | 单一职责 |
| --- | --- |
| doc.go | 部分事实、非目标及并发限制 |
| input.go | Input、Proof、Obligation、KeyIdentifier、Summary |
| errors.go | 五个固定 sentinel，无动态 error wrapping |
| budget.go | preflight，不复制、不解析 |
| owned.go | 两份独立 owned buffers、共享 owner 及 clear |
| material.go | Material.Summary/Clear，无其他方法 |
| spki.go | 声明及 CA raw P-256 SPKI 编码/公钥检查 |
| der.go | 最小 TLV、完整消费、INTEGER/OID 读取 |
| raw_certificate.go | CA 外层/TBS/Name/validity/extension 原始结构 |
| ca_fields.go | 版本、名字、serial、时间与 raw/parser 对应 |
| extensions.go | 四扩展形状与对应，不做 SKI 派生比较 |
| parse.go | 完整固定阶段、绑定、摘要及 owner 移交 |

测试文件（12）：fixtures_test.go、budget_test.go、owned_test.go、material_test.go、spki_test.go、der_test.go、raw_certificate_test.go、ca_fields_test.go、extensions_test.go、parse_test.go、boundary_test.go、fuzz_test.go。不主动增加 generator command 或新 testdata 目录；新 golden 和测试辅助材料均放在 fixtures_test.go，不改全部旧 testdata/向量路径。若 Go fuzz 失败自动写入本包 testdata/fuzz，保留种子及真实失败证据，报告新增路径；未经范围修订不得删除、忽略失败或宣称固定文件交付已经完成。

公开契约：

```go
type Input struct { DER, DeclaredSPKI []byte }
type Proof uint8 // None=0; NodeCAMaterialShapeAndDeclaredKeyBound=1
type Obligation uint8 // UnknownObligation=0; spec §3 六项依顺序为 1–6
type KeyIdentifier [32]byte // 命名类型，非 alias
type Summary struct {
    Proof Proof
    Pending [6]Obligation
    CADERDigest, SPKIDERDigest, SubjectDERDigest, IssuerDERDigest contracts.Digest
    SubjectKeyIdentifier, AuthorityKeyIdentifier KeyIdentifier
    Serial [20]byte
    SerialLength uint8
    NotBefore, NotAfter time.Time
    HasPathLen bool
    PathLen uint8
    CRLSignAllowed bool
}
type Material struct { owned *owner }
func Parse(context.Context, Input) (*Material, error)
func (*Material) Summary() Summary
func (*Material) Clear()
```

五个 error 名字及固定消息为 ErrInvalidInput=`nodeca: invalid input`、ErrLimitExceeded=`nodeca: limit exceeded`、ErrInvalidMaterial=`nodeca: invalid material`、ErrUnsupportedCAProfile=`nodeca: unsupported CA profile`、ErrBindingMismatch=`nodeca: binding mismatch`；没有跨包 error alias。只有这些 error vars 为包级可变值，无 profile registry、cache 或全局 hook。

私有类型决定跨任务接口：`owner{der,declaredSPKI []byte; summary Summary; valid bool}`；`derElement{tag byte; full,content []byte}`；`rawExtension{id asn1.ObjectIdentifier; critical bool; value []byte}`；`rawCA` 包含 derElement 字段 `tbs,serial,issuer,subject,spki,notBefore,notAfter,innerSignatureAlgorithm,outerSignatureAlgorithm,signature` 及 `extensions []rawExtension`。`fieldSummary{Serial [20]byte; SerialLength uint8; NotBefore,NotAfter time.Time}`；`extensionSummary{SubjectKeyIdentifier,AuthorityKeyIdentifier KeyIdentifier; HasPathLen bool; PathLen uint8; CRLSignAllowed bool}`。它们只在一次检查中使用，不保留到成功 owner。

## 2. 执行环境、基线与留痕

规划前只读快照：HEAD=`6752b9612a00b9d6b0bf87ae8dd95063f6b056ee`，index SHA256=`1CBA46BEF4C1AFDC9AE7938146909F0D2D641A0D712CE18036B60FE5B8673D12`；734 个非忽略文件、339 条 dirty status。它不是实施基线或测试证据；批准本计划后必须重新 capture，包含本次计划和 spec 留痕。

执行时先用 apply_patch 建立专属 `.superpowers/sdd/2026-10-11-node-ca-material/baseline.json` 与 `ledger.md`；逐次命令记录使用 `attempt-0001.md` 起的连续四位编号。目录必须已由只读 ignore 检查确认为忽略；如不是，报告并停在执行基线，不改 ignore 配置。不存在证据目录才创建；若已有内容，先审阅归属，不覆盖。不得读写其他增量的 scratch。

所有 Go 命令都使用下面完整 PowerShell 环境，workdir 为上述 worktree；本轮仅写下命令，不执行或创建 cache：

```powershell
$env:GOTOOLCHAIN='local'
$env:GOPROXY='off'
$env:GOSUMDB='off'
$env:GOAUTH='off'
$env:GOVCS='all:off'
$env:GOFLAGS='-mod=readonly'
$env:GOPATH='C:/Users/Lenovo/go'
$env:GOMODCACHE='C:/Users/Lenovo/go/pkg/mod'
$env:GOCACHE='D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/.superpowers/sdd/2026-10-11-node-ca-material/go-cache'
$caGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
$caFmt='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/gofmt.exe'
```

每次 RED/GREEN、格式检查、失败、补充检查和最终命令均在开始前记录环境（包含实际继承的 GODEBUG，不静默覆盖）、完整命令、workdir、实际开始 UTC、批准文档及当前源码 hashes；完成后记录实际结束 UTC、完整输出、最终退出码和 post hashes。标识值不依赖 GODEBUG 或 Go 自动生成。日志用 apply_patch 保存；不能补造缺失时间或省略失败。RED 必须由本任务缺声明或真正行为断言失败导致，语法/fixture/权限/缺依赖失败不算 RED。

普通 test 固定 `-count=1 -p=1 -timeout=120s`；fuzz 另固定 `-run=^$ -count=1 -p=1 -parallel=1 -fuzztime=5s -timeout=30s`。返回运行会话就等待原会话的实际退出码，保持进度更新；不得重复启动来替代等待。超时、非预期失败、缺离线依赖或拒绝须记录后诊断，不能静默加时、下载依赖或绕过平台限制。

只读 Git 使用 `GIT_OPTIONAL_LOCKS=0`、`core.fsmonitor=false`，必要时通过权限机制提权读取；不写 safe.directory。每阶段核对所有旧文件和旧向量 hashes、branch/HEAD/index/status；除批准的状态文档/留痕，不得改旧文件。测试 baseline 与证据不是跨电脑恢复检查点。

## Task 1: 输入合同、预算与两 buffer 生命周期

**Files:** 新建 doc.go/input.go/errors.go/budget.go/owned.go/material.go；测试 budget_test.go/owned_test.go/material_test.go。

**Interfaces:** 产生 §1 公开类型、error、两个 Material 方法；私有 `preflight(ctx context.Context,in Input) error`、`ownInput(in Input) *owner`、`(*owner).clear()`。此任务不提供 Parse stub，不把手造 owner 当 CA 正例。

- [x] **Step 1: 保存获准执行基线。** 核对计划批准与方法选择，完整阅读 spec/plan；记录当前全部既有文件及向量哈希、Git 基线、两个批准文档哈希，确认新源码/证据目录归属，然后建立专属证据/cache。
- [x] **Step 2: 写失败测试。** TestCAPreflight 覆盖 DER 0/1/65536/65537、SPKI 0/1/8192/8193；顺序为 nil context→已取消→任一超限→空值。TestCAFixedErrors 检查五个独立固定 sentinel/messages。TestCAOwnedInput 用共享 backing/subslice 检查两份独立克隆与 clear；TestCAMaterialLifecycle 检查 nil、零值、包装拷贝、detached Summary、幂等与全部字段清零。
  ```go
  if !errors.Is(preflight(context.Background(), Input{DeclaredSPKI: make([]byte,8193)}), ErrLimitExceeded) { t.Fatal("budget priority") }
  o := ownInput(Input{DER: []byte{1,2}, DeclaredSPKI: []byte{3,4}})
  a,b := o.der,o.declaredSPKI; o.valid=true
  m := &Material{owned:o}; copied:=*m; m.Clear()
  if copied.Summary()!=(Summary{}) || !bytes.Equal(a,[]byte{0,0}) || !bytes.Equal(b,[]byte{0,0}) { t.Fatal("shared lifecycle") }
  ```
- [x] **Step 3: 记录真实 RED。** Run: `& $caGo test ./internal/nodecontrol/nodeca -run '^(TestCAPreflight|TestCAFixedErrors|TestCAOwnedInput|TestCAMaterialLifecycle)$' -count=1 -p=1 -timeout=120s`；预期缺本任务声明而失败。
- [x] **Step 4: 最小实现。** preflight 不复制/解析；ownInput 内部先建立失败/panic 清理，再逐份 bytes.Clone，初始 valid=false；Material 方法 nil-safe，共享 owner。成功 handoff 和 clear 不保留调用方 slice。
- [x] **Step 5: 记录 GREEN。** 重跑 Step 3 同命令，预期 exit0；再 `& $caGo test ./internal/nodecontrol/nodeca -count=1 -p=1 -timeout=120s`，预期当前合同测试 exit0，不称为 CA Parse 完成。
- [x] **Step 6: 保存阶段检查点。** 记录 RED/GREEN/源码 hashes，核对旧基线，更新本计划本任务 checkbox；不暂存或提交。

## Task 2: 声明 P-256 SPKI 与 CA 原始编码

**Files:** 新建 spki.go/der.go/raw_certificate.go；测试 fixtures_test.go/spki_test.go/der_test.go/raw_certificate_test.go。

**Interfaces:** 消费 Task 1 errors；产生 `checkDeclaredSPKI(der []byte) error`、`checkCASPKI(cert *x509.Certificate,raw rawCA) error`、`readElement(raw []byte)(derElement,[]byte,error)`、`singleElement(raw []byte,tag byte)(derElement,error)`、`takeElement(raw *[]byte,tag byte)(derElement,error)`、`minimalInteger(raw []byte) bool`、`parseRawCA(der []byte)(rawCA,error)`；测试辅助入口 `fixtureInput(parentSigned bool) Input` 返回新克隆的冻结 golden 原始输入，供后续任务使用。scanner 只读 owned 数据，完整消费，不做签名验证或 profile 绑定。

- [x] **Step 1: 建立独立测试材料。** fixtures_test.go 固定已批准附录的 RFC 7093 91-byte SPKI 和 SHA256=`6d20896ab8bd833b6b66554bd59b20225d8a75a296088148399d7bf763d57405`。在该文件冻结新的测试专用自签/不同 parent CA golden DER、显式 SKI/AKI 和独立摘要；摘要以 PowerShell SHA256 对冻结原始 bytes 计算并留痕，不调用产品函数作 oracle。golden 运行时不重签；测试私钥只在本包测试辅助函数中使用。
- [x] **Step 2: 写失败测试。** TestCADeclaredSPKI/TestCAWireSPKI 固定 P-256 exact/canonical、尾随/PEM/压缩/显式参数/非法点/合法不同算法拒绝及 cert raw SPKI 对应。TestCADERTLV 检查短/长长度边界、非最小/indefinite/high-tag/溢出与 trailing；TestCARawCertificate/TestCARawNestedTails 覆盖 TBS、Name SET/ATV、validity、extension 各层完整消费、unique IDs、raw/parser mismatch 和非空 unused=0 signature。
  ```go
  v,tail,err := readElement([]byte{0x04,1,0x7f,0xff})
  if err!=nil || len(v.content)!=1 || !bytes.Equal(tail,[]byte{0xff}) { t.Fatal("TLV cursor") }
  if _,err:=singleElement([]byte{0x04,0x81,1,0x7f},4); !errors.Is(err,ErrInvalidMaterial) { t.Fatal("nonminimal length") }
  ```
- [x] **Step 3: 记录 RED。** Run: `& $caGo test ./internal/nodecontrol/nodeca -run '^(TestCADeclaredSPKI|TestCAWireSPKI|TestCADERTLV|TestCARawCertificate|TestCARawNestedTails)$' -count=1 -p=1 -timeout=120s`；预期缺对应 helper 或行为失败。
- [x] **Step 4: 实现编码检查。** 两种 SPKI 检查均 ParsePKIXPublicKey、命名 P256 合法点、MarshalPKIXPublicKey exact compare；临时 marshal buffer 立即 defer clear。声明 malformed/noncanonical 为 ErrInvalidInput，CA raw malformed/noncanonical 为 ErrInvalidMaterial；合法非 P256 为 ErrUnsupportedCAProfile。TLV 在累乘之前以实际剩余长度约束 accumulator；Name/OID/AlgorithmIdentifier/extension 内层完整消费；不修正或保留重编码、Go key/parser 对象。rawCA 的 signature 为 BIT STRING 外层原始值，保留 inner/outer algorithm exact bytes 用于对应。原始 BIT STRING malformed 为 ErrInvalidMaterial；可解释但空 signature 或 unused bits 非零为 ErrUnsupportedCAProfile，Go 前置拒绝仍优先。raw scanner 不把合法空 Name SEQUENCE 当 malformed；CA 非空 Name/4096-byte 限制由 Task 3 fields 执行，Name SET/ATV 结构不完整仍为 ErrInvalidMaterial。
- [x] **Step 5: 记录 GREEN。** 重跑 Step 3，再完整本包 test，固定 flags、均 exit0。fixture/语法错误须修正并重新取得真实 RED，不能将其登记为行为 RED。
- [x] **Step 6: 保存阶段检查点。** 固定新 golden/独立 oracle 哈希，核对旧文件未变，不提交。

## Task 3: CA 字段与四扩展材料 profile

**Files:** 新建 ca_fields.go/extensions.go；测试 ca_fields_test.go/extensions_test.go；只扩充新 fixtures_test.go 的突变辅助函数。

**Interfaces:** 消费 Task 2 rawCA/SPKI helper；产生 `checkCAFields(cert *x509.Certificate,raw rawCA)(fieldSummary,error)`、`canonicalTime(raw derElement,parsed time.Time) bool`、`checkCAExtensions(cert *x509.Certificate,raw []rawExtension)(extensionSummary,error)`。fieldSummary 只有 unsigned serial/length、UTC times；extensionSummary 只有两个 KeyIdentifier、HasPathLen/PathLen、CRLSignAllowed。CA SPKI 检查包含在 fields 首段；扩展函数只检查形状/对应，SKI hash 比较留给 Task 4。

- [x] **Step 1: 写字段失败测试。** TestCAFields 覆盖 v3、非空名字/4096与4097-byte预算、Name raw/parser 对应、正数/零/负/20及21-octet serial 与必要 padding；时间1/9999与0/10000边界、1950/2049/2050切换、UTC整秒/offset/fraction/倒置/相等。TestCANotLeafLifetime 接受合规的非24h CA窗口，不获取当前时间。
- [x] **Step 2: 写扩展失败测试。** TestCAExtensions 为四种扩展各自缺失/重复/criticality/值/额外已知未知做表；TestCAPathLenPresence 固定省略/0/1/255与-1/256/非最小整数；TestCAExtensionCanonicality 固定 KU 两值与额外bits/尾零、SKI/AKI 1/20/31/32/33/64、AKI constructed/issuer/serial/尾随；合法排列保持同一结果。
  ```go
  // 对同一合规 cert/raw 使用包内新 fixture，只比较 detached projection。
  if absent.HasPathLen || absent.PathLen!=0 || !zero.HasPathLen || zero.PathLen!=0 { t.Fatal("pathLen presence") }
  if certOnly.CRLSignAllowed || !certAndCRL.CRLSignAllowed { t.Fatal("KU projection") }
  ```
- [x] **Step 3: 记录 RED。** Run: `& $caGo test ./internal/nodecontrol/nodeca -run '^(TestCAFields|TestCANotLeafLifetime|TestCAExtensions|TestCAPathLenPresence|TestCAExtensionCanonicality)$' -count=1 -p=1 -timeout=120s`；预期缺声明或行为失败。
- [x] **Step 4: 实现 profile。** fields 按 version/key→Name→serial→time 固定顺序，raw与parser不一致 ErrInvalidMaterial，合解释但超子集 ErrUnsupportedCAProfile。扩展先固定 inventory/criticality，再按 BC→KU→SKI形状→AKI形状；非最小/尾随/结构 malformed 为 ErrInvalidMaterial，能解释的值/长度/profile 不符为 ErrUnsupportedCAProfile。BC 与 IsCA/BasicConstraintsValid/MaxPathLen/MaxPathLenZero、KU 与 KeyUsage、SKI/AKI 与解析 slices 均 exact 对应；不靠 extension 顺序/map迭代定首错。
- [x] **Step 5: 记录 GREEN。** 重跑 Step 3，再完整本包 test，均 exit0；内部 profile helper 的 duplicate 拒绝类别不替代整入口 Go 前置解析的类别。
- [x] **Step 6: 保存阶段检查点。** 记录字段/扩展矩阵与结果，旧两层/向量保持冻结，不提交。

## Task 4: 完整 Parse、声明绑定与 partial 结果

**Files:** 新建 parse.go；测试 parse_test.go；扩充新 fixtures_test.go。

**Interfaces:** 消费 Tasks 1–3，产生唯一 `Parse(ctx context.Context,in Input)(*Material,error)`。没有 leaf/parent 输入或输出访问器；context 在各同步阶段前后检查，取消不被普通错误覆盖。

- [x] **Step 1: 写失败的入口/绑定测试。** TestCAParse 检查两个不同签发关系 golden、KU/pathLen/扩展排列及所有15个 Summary 字段；TestCABindings 用不同合规P256声明与错32-byteSKI固定两类 mismatch、SPKI首错，确认摘要以原始DER/SPKI/Name为preimage。
- [x] **Step 2: 写不升级与取消失败测试。** TestCAPartialProof/TestCAPartialProofParentRoles 覆盖改变32-byteAKI、坏签名bytes、Unknown signer algorithm、过期/未来CA、Subject==Issuer/SKI==AKI，仍可通过材料层且六Pending顺序不变；非空/unused=0以外 signature拒绝。TestCACancellation 用只在 Err 调用阈值后取消的包内 context，对可达阶段逐阈值迭代，均无半成品；无生产 hook/goroutine，不假称克隆panic清零被动态观测。
- [x] **Step 3: 写错误优先级失败测试。** TestCAParserPriority/TestCACombinedFailurePriority 固定 nil/预取消→任一超限→空值→声明SPKI→Go parser→raw scanner→字段→四扩展→SPKI mismatch→SKI mismatch→最终context。Go重复扩展前置拒绝为 ErrInvalidMaterial；完整raw结构中非最小值/内层尾随为 ErrInvalidMaterial；可解释profile/ID长度不符为 ErrUnsupportedCAProfile。错误仅固定sentinel或context，不包含输入字节或动态parser文本。
  ```go
  if got.Proof!=NodeCAMaterialShapeAndDeclaredKeyBound || got.Pending!=([6]Obligation{CertificateSignatureAndIssuerChain,AuthorityIdentifierAndParentBinding,TrustedCurrentTime,DeploymentAuthorizationAndPurpose,BundleStatusAndRollback,LeafIssuerRelationAndRuntimeUse}) { t.Fatal("partial strength") }
  if material!=nil && err!=nil { t.Fatal("partial material on error") }
  ```
- [x] **Step 4: 记录 RED。** Run: `& $caGo test ./internal/nodecontrol/nodeca -run '^(TestCAParse|TestCABindings|TestCAPartialProof|TestCAPartialProofParentRoles|TestCACancellation|TestCAParserPriority|TestCACombinedFailurePriority)$' -count=1 -p=1 -timeout=120s`；预期缺 Parse 或行为失败。
- [x] **Step 5: 最小实现完整阶段。** preflight→ownInput及失败defer→声明SPKI→ParseCertificate及Raw exact→parseRawCA→fields→extensions→SPKI/SKI绑定→四摘要/固定Pending→最终context→valid与owner移交。所有后续读取来自owned；临时对象不进入owner，阶段末context优先，panic不recover成普通成功。hash只使用owned exact原始片段。
- [x] **Step 6: 记录 GREEN。** 重跑 Step 4，再完整本包 test，均 exit0；成功 Clear、失败、取消与detached结果独立验证，不作chain/time/运行结论。
- [x] **Step 7: 保存阶段检查点。** 记录完整入口与错误矩阵，核对旧基线，不提交。

## Task 5: 隔离守卫、fuzz、最终验证与独立审查

**Files:** 新建 boundary_test.go/fuzz_test.go；仅在新包内必要修正。完成后允许新增 `docs/roadmap/2026-10-11-node-ca-material-execution.md`、更新 current-status 顶部，以及本 spec/plan 的顶部状态/checkbox；不改旧需求正文/执行记录。

**Interfaces:** 产品公开 surface 固定为 §1；无消费者开放。FuzzCADER/FuzzCASPKI 是测试入口，其他辅助函数不导出。

- [x] **Step 1: 写守卫失败测试及fuzz测试入口。** TestCAImportsBoundary/TestCAAPI/TestCAConsumers/TestCAFileInventory/TestCAParseOwnershipFlow/TestCANoRuntimeCalls。import/member 白名单仅 bytes(Clone/Equal)、context(Context)、errors(New)、crypto/sha256(Sum256)、crypto/ecdsa(PublicKey)、crypto/elliptic(P256)、crypto/x509(Certificate/ParseCertificate/ParsePKIXPublicKey/MarshalPKIXPublicKey/KeyUsageCertSign/KeyUsageCRLSign)、encoding/asn1(ObjectIdentifier/Unmarshal/RawValue)、time(Time)、contracts(Digest)；必要key/time/oid实例方法仅执行本计划检查，不增加签名/I/O能力。
  API/字段/枚举/error vars与24文件精确核对。产品禁止 Verify/CheckSignature*、pool/签发/时钟/随机/私钥/并发/serialization及旧材料包import。Input.DER 是必需输入字段，不可因粗糙selector禁令误判为导出DER accessor。对导入别名/方法值/局部遮蔽/伪装selector、额外API/兄弟包/外部测试consumer各给真实拒绝case；同目录同package测试才有测试例外。
  同一步先建立 fuzz_test.go：FuzzCADER 以新golden及空/嵌套长度种子调用Parse并Clear成功结果；FuzzCASPKI 以RFC SPKI及新key/非法参数种子调用checkDeclaredSPKI。超外层预算样本验证preflight分类，不在产品克隆超限bytes；成功Material必有六Pending，error必无有效Material。固定24文件检查的RED不以遗漏fuzz文件或语法错误冒充guard行为失败。
- [x] **Step 2: 记录守卫 RED。** Run: `& $caGo test ./internal/nodecontrol/nodeca -run '^(TestCAImportsBoundary|TestCAAPI|TestCAConsumers|TestCAFileInventory|TestCAParseOwnershipFlow|TestCANoRuntimeCalls)$' -count=1 -p=1 -timeout=120s`；写case先于guard实现，预期缺guard helper或注入违规被错误接受，不人为破坏旧守卫。
- [x] **Step 3: 实现测试守卫。** guard 精确目录/package/path，read-only walk扫描消费者；失败清理责任在首个clone前、Parse移交前后上下文/有效状态的有限AST检查。测试guard本身可以读仓库，产品不能I/O；它只给出有限AST证据，不声称形式证明或观测克隆时OOM。
- [x] **Step 4: 记录守卫 GREEN。** 重跑 Step 2 同命令，预期 exit0；随后只格式化24个新Go文件。精确文件集合及格式化命令：
  ```powershell
  $caNames='doc.go input.go errors.go budget.go owned.go material.go spki.go der.go raw_certificate.go ca_fields.go extensions.go parse.go fixtures_test.go budget_test.go owned_test.go material_test.go spki_test.go der_test.go raw_certificate_test.go ca_fields_test.go extensions_test.go parse_test.go boundary_test.go fuzz_test.go'.Split(' ')
  [string[]]$caPaths=@($caNames | ForEach-Object { Join-Path 'internal/nodecontrol/nodeca' $_ })
  [Array]::Sort($caPaths,[StringComparer]::Ordinal)
  & $caFmt -w $caPaths
  ```
  记录真实前后hash及退出码，不格式化旧文件或不受限目录。
- [x] **Step 5: 记录格式检查。** 使用 Step 4 同一精确路径数组，Run: `& $caFmt -l $caPaths`；预期无输出且exit0；有输出即不通过，不能只看退出码。
- [x] **Step 6: 运行五条最终限定Go命令。** 全部使用 §2 完整离线wrapper，逐条记录真实退出码；没有外推全仓库、DB/Docker、race、硬件或production验收。
  ```powershell
  & $caGo test ./internal/nodecontrol/nodeca -count=1 -p=1 -timeout=120s
  & $caGo test ./internal/nodecontrol/nodeleaf -count=1 -p=1 -timeout=120s
  & $caGo test ./internal/nodecontrol/nodeleafkeys -count=1 -p=1 -timeout=120s
  & $caGo test ./internal/nodecontrol/nodeca -run '^$' -fuzz '^FuzzCADER$' -count=1 -p=1 -parallel=1 -fuzztime=5s -timeout=30s
  & $caGo test ./internal/nodecontrol/nodeca -run '^$' -fuzz '^FuzzCASPKI$' -count=1 -p=1 -parallel=1 -fuzztime=5s -timeout=30s
  ```
  每条预期 exit0，fuzz无发现的crash/行为断言失败；限定fuzz通过不等于没有所有潜在parser缺陷。
- [x] **Step 7: 冻结增量并请求一次 fresh 只读审查。** Native 下作者实现五任务后，由当前最有能力可用模型 gpt-6-astra、xhigh 的新审查者检查 spec/plan、24文件、完整证据/旧基线与五条最终命令。本步是计划内未来 delegation，不在规划时启动。审查对象为本增量，不冒充重新审完所有历史dirty改动；禁止reviewer改代码/Git或降低Pending。记录Critical/Important/Minor及证据限制。SDD方法按其技能增加各任务独立门槛，不取消最终整增量审查。
- [x] **Step 8: 处理审查结果。** Critical/Important 未解决不得宣称完成；Minor也记录并修复本范围。源码改变则原冻结hash/审查失效，重跑受影响检查及全部五条最终命令、再次fresh审查；跨范围需求先修订审批，不能借修复改旧profile/provider。无审查资源时保存未审开发状态，不称完成或自行降级模型冒充原审查。
- [x] **Step 9: 保存真实交付记录。** execution 文档逐项列环境、所有attempt、RED/GREEN、最终五命令、格式/source hashes、审查及限制；current-status只新增本增量顶部段，保留CA/leaf实际信任、原六Pending和Task10/B01/C12/main阻塞。spec/plan只更新状态/checklists。核对所有旧文件/testdata/向量、branch/HEAD/index/status及额外文件exact范围；记录均真实后交付，不提交、推送、合并或删除证据。

## 3. 计划书面自检与交接

覆盖：spec §1–3→各任务partial边界及Task4；§4/§7生命周期→Task1/4；§5.1/5.2 raw→Task2/3；§5.3/§6→Task3/4；§7错误顺序→Task4；§8独立向量→Task2、隔离/fuzz→Task5；§9范围及阶段→全局约束和本节。五项Review Focus各有指定测试。没有第二条信任通道、旧消费者例外、算法扩展、未定义跨任务函数或占位实现步骤。

本轮没有运行上述命令，没有形成RED/GREEN或代码审查结果。计划批准与执行方法选择是下一门槛，用户对spec的“是”不批准本计划或代码。

执行建议：**1 Native**，作者在当前会话依次实现五任务，结束后一次fresh独立审查；新包内接口顺序依赖强，且没有旧代码修改，适合此方法。**2 Subagent-driven**，每任务独立实现/审查后推进，再最终审查，上下文与审查成本更高。请用户先审阅计划，再明确选择 1 或 2；当前不启动任何子代理。
