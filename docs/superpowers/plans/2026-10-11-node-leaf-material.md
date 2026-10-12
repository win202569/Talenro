# Node Leaf Material Validation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增独立 nodeleaf 纯材料校验包，检查单张 exact DER 与固定节点/domain/SPKI 对应，始终保留六项未完成责任，不生成证书信任或运行授权。

**Architecture:** 先做预算与所有权，再用 Go X.509 解析和私有有限 DER 检查验证封闭子集，最后逐字绑定 URI/SPKI 并移交结果。所有实现只在新包中，现有产品不接入；没有链验证、时间提供器、数据库或主状态转换。

**Tech Stack:** 本地 Go 1.26.5、标准库、既有 contracts.Digest 与 github.com/google/uuid v1.6.0；PowerShell、apply_patch、离线缓存；无新依赖。

**Spec:** [已批准书面设计](../specs/2026-10-11-node-leaf-material-design.md)。用户于 2026-10-11 批准；执行者必须完整阅读设计和本计划。设计中的“待批准”行保留上轮交付记录，本轮不改设计文件。

状态：计划已获用户批准并选择方案 1 Native；五任务完成限定本机实现、验证与一次 fresh-context 独立只读审查。checkbox 仅据实际执行勾选，未提交或推送。

原文的“本轮”保留计划编写阶段说明；实施结果见 [执行记录](../../roadmap/2026-10-11-node-leaf-material-execution.md)。最终 17 项 exit 0，store 仅编译；0 Critical / 0 Important / 2 Minor 历史留痕限制已披露。六项 Pending 与既有整体验收阻塞保持，不升格证书认证/运行权限。

## Global Constraints

- 复用 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`，分支 `codex/c12-b01-task9-coordinator`；不新建 worktree，不清理旧文件，保留全部既有未提交进度。
- 不暂存、提交、推送、合并、重置、删分支或修改 safe.directory/永久环境；每任务的证据记录替代技能模板 Commit 步骤。
- `DER []byte`：单张 exact DER，1–65,536 bytes；`ExpectedSPKI []byte`：单张 exact DER SPKI，1–512 bytes。空值 InvalidInput，外层超限 LimitExceeded，全部预算先于复制/解析。
- `NodeID` 非零且 RFC4122 variant；domain 小写 ASCII DNS，总长 1–253，每 label 1–63，首尾 `[a-z0-9]`，内部仅 `[a-z0-9-]`；不 trim/case-fold/IDNA/URL 规范化。
- exact URI 为 `spiffe://<TrustDomain>/node/<NodeID.String()>`，最大 304 ASCII bytes；exact leaf SPKI 等于 expected SPKI，不能只比较点或摘要。
- X.509 v3、empty subject exact `30 00`、命名曲线 P-256 ECDSA 合法点；不限制 issuer 算法或验证签名，不接受 unique identifiers/额外 TBS 尾字段。
- serial 数学值 >0；最小正数 DER INTEGER 内容含必要 padding 总长 ≤20；导出最小 unsigned magnitude 1–20，零填充 `[20]byte` 其余区。20-byte 高位 magnitude 需要 21-byte INTEGER 时拒绝，不改旧 lookup 规则。
- raw 时间含秒且 Z 结尾，无 offset/小数秒；1950–2049 用 UTCTime，否则 GeneralizedTime；年份 1–9999；UTC 且 NotBefore<NotAfter，exact 24h。无系统时间、backdate 或 freshness 证明。
- exact 六项 OID/criticality：BC `2.5.29.19` true、KU `2.5.29.15` true、EKU `2.5.29.37` false、SAN `2.5.29.17` true、SKI `2.5.29.14` false、AKI `2.5.29.35` false；顺序可变，但 inventory/检查顺序固定。
- BC 值 `30 00`；KU `03 02 07 80`；EKU 只有 clientAuth `1.3.6.1.5.5.7.3.2`；SAN 恰一个 primitive [6]。raw 内值最短 DER 且完全消费，未知 noncritical、重复或额外项同样拒绝。
- SKI 非空 OCTET STRING 1–64；AKI 恰一个 primitive [0] keyIdentifier 1–64；不选择派生算法。keyIdentifier-only/64-byte 是本地保守子集，不是完整 profile 或 RFC 通用规则。
- 先拥有 DER/SPKI 才解析；临时 x509.Certificate/公钥/大整数不保留、不导出。Material 包装复制共享 owner，Clear 后全部包装失效；Summary/DER detached，nil-safe、幂等清理。
- 错误、取消返回 nil Material；可正常展开的 panic 清理后传播。清理登记先于后续复制/解析/Err 调用，不承诺字符串、不可达分配、fatal OOM 或全进程擦除，不支持并发读写/Clear。
- 有效 Proof 只为 NodeLeafMaterialShapeValid，Pending 固定六项；None 零结果不是 Ready。LeafDERDigest/SPKIDERDigest 的 SHA-256 preimage 分别是 exact DER/SPKI，不转换旧 PublicKeyDigest、issuer/serial digest。
- 新产品包不导入旧隔离证据包，其他产品包不消费 nodeleaf；旧源码/守卫/向量/SQL/module/lock 全部冻结。
- 不提供 Recheck/Verify/State/Ready/Authorize/Commit/序列化入口；无 Verify/CheckSignature*、CertPool/system roots、TLS、签发、私钥读取、网络、DB、文件 I/O、时钟、随机、hook 或 goroutine。
- 产品有限 DER helper 不发展成通用 ASN.1 框架；不重新编码或修正原证书以接受。测试可真实签发/变异，但不导入生产 helper 来生成期望真值。
- 本增量仅离线局部验证；无全仓/race/真实 DB/Docker/硬件/C11/C12 PITR/生产 conformance。现有验收阻塞不解除。
- 手工编辑用 apply_patch；gofmt 仅指定 20 个新 Go 文件；每个测试会话必须实际结束，不能重复启动 pending 测试或用发布延迟判断失败。

## Review Focus

1. URI+otherName/registeredID 仍可能只有一个公开 URIs 项；必须查 raw GeneralNames（Task 3/4：TestLeafIgnoredSANArm）。
2. 20-byte 高位 serial 与大 DER length 的 padding/整数溢出；不能用 magnitude 长度或截断后的 length 接受（Task 2：TestLeafSerialPadding、TestLeafLengthOverflow）。
3. 公钥点相同但 SPKI 不合规/有尾随，以及 DER/SPKI 摘要角色混淆；只能接受 exact 最短 P-256 SPKI（Task 2/4：TestLeafExpectedSPKI、TestLeafSPKIBinding）。
4. 第二次复制前的早清理登记、复制后/解析后/最终返回前的取消或 panic，以及复制包装再 Clear；不能泄露 owner 或陈旧 summary（Task 1/4：TestLeafMaterialLifecycle、TestLeafLateCancellation、TestLeafCleanupRegistration）。
5. 有效形状但过期/未来、坏签名或错误 key ID 派生；不扩大认证，成功仍保留六项 Pending（Task 4：TestLeafPartialProofOnly）。

---

## 文件结构与固定接口

新目录 `internal/nodecontrol/nodeleaf/` 固定 20 个 Go 文件，不新增 testdata：

| 产品文件（10） | 责任 |
| --- | --- |
| doc.go / input.go | 部分证据边界 / 输入、Proof、Pending、Summary 类型 |
| errors.go / budget.go | 固定错误 / 预算、UUID/domain 纯值检查、期望 URI |
| owned.go / material.go | 双 buffer 所有权、清理 / 私有 owner 包装与 detached 输出 |
| der.go | 单 DER、TBS/扩展 raw 范围、最短有限 TLV 检查 |
| leaf_fields.go | expected SPKI、leaf 基本字段、serial/time 子集 |
| extensions.go | 六项 inventory/criticality/原始内值与 parsed 字段对应 |
| parse.go | 唯一 public Parse、固定阶段顺序、URI/SPKI 绑定、摘要移交 |

测试（10）：`fixtures_test.go`、`errors_test.go`、`budget_test.go`、`owned_test.go`、`material_test.go`、`der_test.go`、`leaf_fields_test.go`、`extensions_test.go`、`parse_test.go`、`boundary_test.go`。每份责任对应相邻产品文件；fixture/守卫仅测试代码。

公开接口固定如下；Proof/Obligation 是描述值，不是验证 token：

```go
type Input struct {
    DER []byte
    NodeID uuid.UUID
    TrustDomain string
    ExpectedSPKI []byte
}
type Proof uint8 // None=0; NodeLeafMaterialShapeValid=1
type Obligation uint8 // UnknownObligation=0; 以下六项按 1..6 顺序
// LeafSignatureAndIssuerTrust, KeyIdentifierDerivationAndIssuerBinding,
// TrustedCurrentTime, IdentityAndAuthorityProvenance,
// PrivateKeyPossession, MainStateAndRuntimeBinding
type Summary struct {
    Proof Proof
    Pending [6]Obligation
    NodeID uuid.UUID
    TrustDomain string
    LeafDERDigest, SPKIDERDigest contracts.Digest
    Serial [20]byte
    SerialLength uint8
    NotBefore, NotAfter time.Time
}
type Material struct { owned *owner }
func Parse(ctx context.Context, in Input) (*Material, error)
func (m *Material) Summary() Summary
func (m *Material) DER() []byte
func (m *Material) Clear()
```

私有 `owner` 只含 `der, expectedSPKI []byte; summary Summary; valid bool`。错误变量固定为 ErrInvalidInput、ErrLimitExceeded、ErrInvalidMaterial、ErrUnsupportedLeafProfile、ErrBindingMismatch，分别使用固定消息 `node leaf: invalid input`、`node leaf: limit exceeded`、`node leaf: invalid material`、`node leaf: unsupported leaf profile`、`node leaf: binding mismatch`；不包装 parser 错误或输入值。

实施期间才允许新增专属证据目录 `.superpowers/sdd/2026-10-11-node-leaf-material/`，包含 baseline.json、progress.md、逐次命令证据、冻结/审查记录和 go-cache/；未知已有文件先查归属，不覆盖。最终新增 `docs/roadmap/2026-10-11-node-leaf-material-execution.md`，有限更新 current-status 的本增量摘要及本设计/计划状态。旧执行记录不改。本轮写计划不创建这些目录或更新上述文件。

## 环境、基线、证据与停止条件

所有未来测试从上述 worktree、PowerShell、login=false 执行，每进程设置：

```powershell
$env:GOENV='off'
$env:GOWORK='off'
$env:GOTOOLCHAIN='local'
$env:GOPROXY='off'
$env:GOSUMDB='off'
$env:GOAUTH='off'
$env:GOVCS='all:off'
$env:GOFLAGS='-mod=readonly'
$env:GOPATH='C:/Users/Lenovo/go'
$env:GOMODCACHE='C:/Users/Lenovo/go/pkg/mod'
$env:GOCACHE='D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/.superpowers/sdd/2026-10-11-node-leaf-material/go-cache'
$leafGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
$leafFmt='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/gofmt.exe'
```

Go/gofmt 路径写计划时已确认存在，但未执行。HEAD 已核对为 `6752b9612a00b9d6b0bf87ae8dd95063f6b056ee`；索引 SHA256 `1CBA46BEF4C1AFDC9AE7938146909F0D2D641A0D712CE18036B60FE5B8673D12`；批准设计 SHA256 `2CAE43DE6CCEEB3416E68116E941EA0D7E6B7DAE62A94DA7ECC3ABEAB20CE2CC`。实施前重新读取本次全部源码/旧守卫/SQL/module/lock/捕获向量、branch/HEAD/index/未提交清单及审批两文件；当前 worktree 既有脏内容不是本轮可替换基线。

每次 RED/GREEN/补充尝试/最终回归均保存完整命令与 selector、开始/结束 UTC、实际输出/退出码、任务 ID、前后源哈希。后台返回 session ID 时等待该 session 结束，记录最终退出码；存档用 apply_patch，不 shell 重定向。无退出码/未结束不得验收，早期记录缺失不得靠回忆补造。

每条 Go test 固定 `-count=1 -timeout 120s -p=1`；使用同任务 selector 的 RED 与 GREEN，整包测试独立记录。RED 只能来自预期缺接口或真实行为断言，不是 fixture 错误、语法错误、缓存/依赖/权限拒绝。超时或非预期失败先保存真实证据再诊断；不静默加时、无限重试、下载依赖或扩大范围。

只读 Git 用 GIT_OPTIONAL_LOCKS=0、core.fsmonitor=false；必要提权按权限机制审查，不改 safe.directory。任何范围外缺陷、算法/能力新增、旧守卫需放宽、工具拒绝或资料矛盾，完成不受影响检查后报告并提出修订；不绕过平台拒绝，不把局部成功当整体验收。

## Task 1: 固定类型、预算与共享 owner 生命周期

**Files:** Create doc.go、input.go、errors.go、budget.go、owned.go、material.go；Create errors_test.go、budget_test.go、owned_test.go、material_test.go，全部位于新目录；不创建 Parse stub。

**Interfaces:** 产生上述公开类型、错误和三个 Material 方法；私有 `preflight(ctx context.Context, in Input) error`、`expectedURI(in Input) string`、`ownInput(in Input) *owner`、`(*owner).clear()`。preflight 不解析、不复制；ownInput 对两 buffer 独立复制，valid 初始 false。

- [x] **Step 1: 建立实施基线。** 核对授权/执行方法/既有 worktree，保存 baseline 和空任务台账；仅创建专属证据/cache，不改旧文件。
- [x] **Step 2: 写失败测试。** TestLeafBudget 覆盖 0/1/65,536/65,537 DER、0/1/512/513 SPKI，context 优先于预算、预算优先于 UUID/domain；TestLeafScope 覆盖 nil/错误 variant UUID、253/254 总长、63/64 label、大小写/空白/尾点/Unicode，不加版本约束；TestLeafFixedErrors 检查五类 errors.Is 和固定消息。TestLeafOwnedAliases 检查 DER/SPKI 共 backing/subslice 仍独立拥有；TestLeafMaterialLifecycle/TestLeafNilMaterial 检查共享包装 Clear、零 summary、detached 输出。
  ```go
  err := preflight(context.Background(), Input{DER: make([]byte, 65537), ExpectedSPKI: []byte{1}})
  if !errors.Is(err, ErrLimitExceeded) { t.Fatal(err) }
  o := ownInput(Input{DER: []byte{1, 2}, ExpectedSPKI: []byte{3, 4}})
  backing := o.der
  o.valid = true // 仅内部生命周期实验，不是 Parse 正例
  m := &Material{owned: o}; copied := *m; m.Clear()
  if copied.Summary() != (Summary{}) || copied.DER() != nil || !bytes.Equal(backing, []byte{0, 0}) { t.Fatal("stale owner") }
  ```
- [x] **Step 3: 验证 RED。** 运行 `& $leafGo test ./internal/nodecontrol/nodeleaf -run '^(TestLeafBudget|TestLeafScope|TestLeafFixedErrors|TestLeafOwnedAliases|TestLeafMaterialLifecycle|TestLeafNilMaterial)$' -count=1 -timeout 120s -p=1`；预期缺声明或真实断言失败，记录实际证据。
- [x] **Step 4: 最小实现。** budget.go 按 context/空值及上限/UUID/domain 检查；expectedURI 仅用于 preflight 已接受值。ownInput 创建 owner 后、首次 bytes.Clone 前登记失败清理，第二 copy 失败也清第一份；成功移交责任。Material 方法只读共享 owner，invalid/nil 返回零值，Clear 清零两 buffer、summary、valid，不清 caller/已分离输出。doc.go 明确纯材料和六项责任。
- [x] **Step 5: 验证 GREEN。** 同 selector 及 `& $leafGo test ./internal/nodecontrol/nodeleaf -count=1 -timeout 120s -p=1` 分别 exit 0；仅内部 owner 实验，不声称材料 Parse 完成。
- [x] **Step 6: 保存阶段证据。** 更新 task 状态/UTC/命令/源哈希；不提交，按所选方法进入下一任务。

## Task 2: 真实离线 fixture、有限 DER 与基本字段

**Files:** Create fixtures_test.go、der.go、der_test.go、leaf_fields.go、leaf_fields_test.go；必要时仅修正 Task 1 新文件。

**Interfaces:** 消费 Input/固定错误；产生私有 `derElement{tag byte; full, content []byte}`、`rawExtension{id asn1.ObjectIdentifier; critical bool; value []byte}`、`rawLeaf{serial,subject,spki derElement; notBefore,notAfter derElement; extensions []rawExtension}`；`readElement(raw []byte) (derElement, []byte, error)`、`parseRawLeaf(der []byte) (rawLeaf,error)`、`checkExpectedSPKI(der []byte) error`、`checkLeafFields(cert *x509.Certificate, raw rawLeaf) ([20]byte,uint8,error)`。临时 raw/parsed 只引用 owned 材料，不保留到 owner。

测试 helper 固定为 `leafFixture(t testing.TB) Input`、`leafWithTemplate(t testing.TB, edit func(*x509.Certificate)) Input`、`rewriteFixture(t testing.TB, in Input, path string, replacement []byte) Input`。rewriteFixture 仅支持 certificate、tbs、version、serial、notBefore、notAfter、subject、spki、extensions、signature 路径；用测试专属 ASN.1 RawValue 定位/重建长度，不调用产品 DER helper。变异允许签名失效，不误称真实重新验签。

- [x] **Step 1: 写 fixture 与失败测试。** 固定测试 issuer/leaf scalar=1/2、node `11111111-2222-4333-8444-555555555555`、domain `node.test`、serial=128、NotBefore `2026-10-11T00:00:00Z`、NotAfter 次日同时间；test-only CreateCertificate 显式六项 ExtraExtensions，SKI bytes 1..20、AKI bytes 21..40，仅作形状。TestLeafFixtureControls 用标准库独立核对 v3、精确时间/SPKI/serial、六 OID/criticality 和 raw 值，并以固定 issuer 公钥离线验证正例签名；不能用本包得出期望。签名可以随机，每份 fixture 的字段与原始材料摘要独立核对，不承诺跨运行 exact DER 相同；两个测试 key 仅测试用，不作为生产配置。
- [x] **Step 2: 写 raw/字段断言。** TestLeafDERGrammar/TestLeafLengthOverflow 覆盖 empty/截断/indefinite/非最短 length/high-tag/超大 length/outer 与 TBS 尾随/unique IDs；合法 v1/v2 为 unsupported，非法 tag 为 material。TestLeafExpectedSPKI 覆盖 exact P-256、尾随、非最短、RSA/P-384/非法点、同点不同非合规编码→InvalidInput。TestLeafSerialPadding 用 INTEGER 内容 `00 80`→magnitude `80`，20-byte `7f...` 接受，`00 80...` 共21 bytes 拒绝；零/负/非最小值拒绝。TestLeafFieldWindow 覆盖倒置/23h59m59s/24h1s、1950/2049/2050/9999 边界、offset/小数秒/错误 tag、年份0或10000，不截断或补时间。
  ```go
  v, tail, err := readElement([]byte{0x04, 0x81, 0x01, 0x01})
  _ = v; _ = tail
  if !errors.Is(err, ErrInvalidMaterial) { t.Fatal("nonminimal length accepted") }
  in := leafFixture(t)
  cert, err := x509.ParseCertificate(in.DER); if err != nil { t.Fatal(err) }
  raw, err := parseRawLeaf(in.DER); if err != nil { t.Fatal(err) }
  serial, n, err := checkLeafFields(cert, raw)
  if err != nil || n != 1 || serial[0] != 0x80 { t.Fatal("serial padding") }
  ```
- [x] **Step 3: 验证 RED。** `& $leafGo test ./internal/nodecontrol/nodeleaf -run '^(TestLeafFixtureControls|TestLeafDERGrammar|TestLeafLengthOverflow|TestLeafExpectedSPKI|TestLeafSerialPadding|TestLeafFieldWindow)$' -count=1 -timeout 120s -p=1`；fixture 若失败先修 fixture，不把它计为产品 RED。
- [x] **Step 4: 实现有限 raw 与字段检查。** readElement 只支持本 grammar 的单 byte tag，累积长度前检查 overflow/剩余字节，full/content/tail 精确对应；parseRawLeaf 消费单外层证书/TBS/字段/扩展，检查 minimal lengths、OID、critical BOOLEAN canonical/default，拒绝额外字段；不通用递归验全部 ASN.1。基本字段按 version→subject→P-256 点→serial→时间顺序；raw 时间逐字与 parsed UTC 的指定格式对应，不接受规范化后的偏差。checkExpectedSPKI 先单最短 DER/完整消费，再 ParsePKIXPublicKey/点检查，不验证签名。
- [x] **Step 5: 验证 GREEN。** 同 selector、完整新包分别 exit 0；raw 子字段与 Go parsed 值独立对应，所有 six extension 语义仍待 Task 3。
- [x] **Step 6: 保存本阶段 UTC、完整 selector/输出/哈希。** 不新增 vectors、不提交，不把 DER-only 检查描述为完整材料入口。

## Task 3: 六项扩展集合与原始内值

**Files:** Create extensions.go、extensions_test.go；只扩充新 fixtures_test.go/der_test.go 中该任务需要的变体。

**Interfaces:** 消费 Task 2 rawLeaf/rawExtension 与 x509.Certificate；产生私有 `checkExtensions(raw []rawExtension, cert *x509.Certificate) ([]byte,error)`，返回唯一 raw URI（临时 owned DER 切片），此时不与 NodeID/domain 比较；错误仅固定类别。

- [x] **Step 1: 写失败测试。** TestLeafExtensionInventory 覆盖六项任意排列、每项缺失/重复/额外、unknown critical/noncritical、六项逐一错误 criticality；TestLeafExtensionValues 逐一 BC/KU/EKU/SAN/SKI/AKI 闭合表负例：explicit FALSE/default、pathLen、extra KU/EKU/Any/serverAuth、0/65-byte IDs、AKI issuer/serial arm/重复 [0]、inner trailing/nonminimal/wrong tag；每项最短边界1/64 IDs合规通过。TestLeafIgnoredSANArm 明确 URI+otherName 与 URI+registeredID 先在 Go parser 得到 len(URIs)=1，却必须被本 helper 拒绝；DNS/IP/email/第二 URI同样拒绝。
  ```go
  in := leafWithTemplate(t, func(c *x509.Certificate) {
      for i := range c.ExtraExtensions {
          if c.ExtraExtensions[i].Id.String() == "2.5.29.15" { c.ExtraExtensions[i].Value = []byte{3, 2, 6, 0xc0} }
      }
  })
  cert, err := x509.ParseCertificate(in.DER); if err != nil { t.Fatal(err) }
  raw, err := parseRawLeaf(in.DER); if err != nil { t.Fatal(err) }
  if _, err := checkExtensions(raw.extensions, cert); !errors.Is(err, ErrUnsupportedLeafProfile) { t.Fatal("extra KU") }
  ```
- [x] **Step 2: 验证 RED。** `& $leafGo test ./internal/nodecontrol/nodeleaf -run '^(TestLeafExtensionInventory|TestLeafExtensionValues|TestLeafIgnoredSANArm)$' -count=1 -timeout 120s -p=1`；预期缺 helper 或真实拒绝逻辑缺失。
- [x] **Step 3: 实现封闭集合。** 先按 raw 原顺序建立六个固定槽，检查数量/OID/重复/criticality，再固定 BC→KU→EKU→SAN→SKI→AKI；内值完整消费、minimal DER、parsed BasicConstraints/KeyUsage/EKU 与 raw 对应。URI 仅检查 primitive [6]、ASCII、1–304 与唯一性；不解析 URL 等价。任意 SKI/AKI bytes 只证明形状，不能派生/比 issuer key。
- [x] **Step 4: 验证 GREEN。** 同 selector 与完整新包分别 exit 0。Go 在 helper 前拒绝的变体使用原始 helper 单元样例查 own rule，端到端错误类别留给 Task 4；不伪称 Go 解析成功。
- [x] **Step 5: 保存阶段证据与源冻结。** 此任务仍没有 public Parse 接入；不放宽旧 guard、不提交。

## Task 4: 完整 Parse、精确绑定与清理/能力守卫

**Files:** Create parse.go、parse_test.go、boundary_test.go；扩充新 owned_test.go/material_test.go/fixtures_test.go；必要修正仅前述新产品文件。

**Interfaces:** 消费 Task 1–3 所有私有 helper；产生 public Parse。test-only helper `leafImportsAllowed(source []byte) bool`、`leafCleanupRegistered(source []byte) bool`、`leafConsumerAllowed(path,pkg,imported string) bool`；没有新产品 callback/global hook。

- [x] **Step 1: 写完整入口失败测试。** TestLeafParse/TestLeafURIBinding/TestLeafSPKIBinding 检查 real fixture 成功、raw DER detached、两个独立 SHA-256 preimage、[20]byte padding、fixed Proof/Pending；canonical URI 的 case/percent/port/userinfo/query/fragment/extra path/末尾斜杠均拒绝（若 Go 已拒绝，InvalidMaterial；Go 可解析单 ASCII URI不同则 BindingMismatch）。合法不同 P-256 leaf/expected SPKI→BindingMismatch，非法 expected SPKI→InvalidInput。TestLeafParseRejections 重放 Task 2/3 全部拒绝变体，结果 nil；Go 提前拒绝不包装原错误。
  ```go
  in := leafFixture(t)
  m, err := Parse(context.Background(), in)
  if err != nil || m == nil { t.Fatal(err) }; defer m.Clear()
  s := m.Summary()
  want := [6]Obligation{LeafSignatureAndIssuerTrust, KeyIdentifierDerivationAndIssuerBinding, TrustedCurrentTime, IdentityAndAuthorityProvenance, PrivateKeyPossession, MainStateAndRuntimeBinding}
  if s.Proof != NodeLeafMaterialShapeValid || s.Pending != want { t.Fatal("proof enlarged") }
  if s.LeafDERDigest != contracts.Digest(sha256.Sum256(in.DER)) || s.SPKIDERDigest != contracts.Digest(sha256.Sum256(in.ExpectedSPKI)) { t.Fatal("preimage swapped") }
  ```
- [x] **Step 2: 写非认证与取消测试。** TestLeafPartialProofOnly 使用固定早/晚日期、合法结构但翻转 signature BIT STRING 一位、不同非空 SKI/AKI bytes，Parse 仍成功且六责任不变；不把签名语法损坏当坏签名。TestLeafNegativeSerialCompatibility 以 raw 变体分别测试 parser 默认及 test-only t.Setenv("GODEBUG","x509negativeserial=1")，两者都不得成功，测试结束自动恢复、不 Parallel。TestLeafCancellation/TestLeafLateCancellation 用 test-only Context.Err 先记录一次完整成功轨迹，再按检查点逐一取消/到期/正常 panic，含复制后、解析后、摘要后/最终返回前；不硬编码调用次数、不承诺即时打断标准 parser。TestLeafParseOwnership 检查源与 subslice 修改、包装复制、detached 数据不影响 owner。
- [x] **Step 3: 写守卫失败测试。** TestLeafCapabilityBoundary/TestLeafProductImports/TestLeafConsumers 检查下文 import/member 表、20 文件清单、exported types/fields/五错误/3方法/唯一 Parse；内存正反例识别合法 named alias，拒绝 dot/blank/duplicate imports、alias 名冲突、forbidden members 及其方法值转存、其他 evidence/serving/runtime 消费者、goroutine。TestLeafCleanupRegistration 检查 preflight 在 ownInput 前、ownInput 首次 clone 前 defer、Parse 取得 owner 后立即 defer、最后 valid/transfer/return；内存变异 missing/late/conditional/wrong-owner defer、未保护 copy、提前 transfer 必须失败。
- [x] **Step 4: 验证 RED。** `& $leafGo test ./internal/nodecontrol/nodeleaf -run '^(TestLeafParse|TestLeafURIBinding|TestLeafSPKIBinding|TestLeafParseRejections|TestLeafPartialProofOnly|TestLeafNegativeSerialCompatibility|TestLeafCancellation|TestLeafLateCancellation|TestLeafParseOwnership|TestLeafCapabilityBoundary|TestLeafProductImports|TestLeafConsumers|TestLeafCleanupRegistration)$' -count=1 -timeout 120s -p=1`；不是直接改真实源码制造故障。
- [x] **Step 5: 实现 Parse 与固定守卫。** 按设计第6节顺序调用 preflight→ownInput（立即失败 defer）→owned expected SPKI→Go ParseCertificate→parseRawLeaf→checkExtensions→checkLeafFields→exact URI→exact SPKI→protected summary→最终 ctx→valid/transfer/Material。raw helper 先完成范围/语法，field subset 在扩展后；无 parsed 指针进入 owner。各复制/解析/子集/绑定阶段边界检查 ctx，standard parser 内部不插 hook；错误只 sentinel/ctx.Err，panic 用 defer 自然传播。清理守卫固定有限源码形状，不声称形式化证明。
- [x] **Step 6: 验证 GREEN。** 同 selector 与完整新包分别 exit 0；Clear captured backing 的可观测实验与失败私有 owner 的源码证据分开，不伪称测到不可达内存擦除。
- [x] **Step 7: 保存全部取消/失败尝试与源冻结。** 不接入任何现有入口，不提交。

产品 imports/member 白名单仅：bytes(Clone/Equal)、context(Context)、errors(New)、crypto/sha256(Sum256)、crypto/ecdsa(PublicKey)、crypto/elliptic(P256)、crypto/x509(Certificate/ParseCertificate/ParsePKIXPublicKey/KeyUsageDigitalSignature/ExtKeyUsageClientAuth)、encoding/asn1(ObjectIdentifier/Unmarshal/RawValue)、time(Time/Hour)、github.com/google/uuid(UUID/Nil/RFC4122)、contracts(Digest)。cert/key/time/serial 的普通字段/安全方法允许，但任何 selector Verify/CheckSignature*/State/Ready/Authorize/Commit、时钟或池/签发能力拒绝；不能因为 aliased 或存成方法值而放行。包级可变值只允许五个固定 error 变量，OID/Pending 表在函数内生成；hook、unsafe/反射、额外产品文件拒绝；其他产品消费者一律拒绝，不改旧 guard。

## Task 5: 十七项有界回归、独立审查与本机交付

**Files:** 无新产品能力；只维护专属证据、最终执行文档、设计/本计划状态及 current-status 本增量摘要，不改旧执行记录/旧源码/向量。

**Interfaces:** 消费20新文件及真实 RED/GREEN；交付限定验证、审查裁定、最终冻结；不生成生产或 Git 集成权限。

- [x] **Step 1: 检查格式范围。** 显式 $leafFormatFiles 列表仅上述20新Go文件，断言 Count=20、无未知文件，运行 `& $leafFmt -w $leafFormatFiles`；没有任何旧 Go 文件参与。
- [x] **Step 2: 检查格式结果。** `& $leafFmt -l $leafFormatFiles` exit 0 且输出为空；保存格式后源/全部旧文件与向量冻结，确认旧哈希无漂移。
- [x] **Step 3: 逐条执行下列17项。** 每条独立进程、环境、UTC、selector、实际输出/退出码；15个完整包+authority限定+store仅编译，全部会话结束且 exit 0 才过门；不将 store 编译称为运行验收。

| 范围 | 固定命令 |
| --- | --- |
| nodeleaf | `& $leafGo test ./internal/nodecontrol/nodeleaf -count=1 -timeout 120s -p=1` |
| statetimebinding | `& $leafGo test ./internal/nodecontrol/statetimebinding -count=1 -timeout 120s -p=1` |
| statepolicybinding | `& $leafGo test ./internal/nodecontrol/statepolicybinding -count=1 -timeout 120s -p=1` |
| statebinding | `& $leafGo test ./internal/nodecontrol/statebinding -count=1 -timeout 120s -p=1` |
| statecontract | `& $leafGo test ./internal/nodecontrol/statecontract -count=1 -timeout 120s -p=1` |
| statematerials | `& $leafGo test ./internal/nodecontrol/statematerials -count=1 -timeout 120s -p=1` |
| statecomposition | `& $leafGo test ./internal/nodecontrol/statecomposition -count=1 -timeout 120s -p=1` |
| stateformat | `& $leafGo test ./internal/nodecontrol/stateformat -count=1 -timeout 120s -p=1` |
| stateverify | `& $leafGo test ./internal/nodecontrol/stateverify -count=1 -timeout 120s -p=1` |
| statepolicy | `& $leafGo test ./internal/nodecontrol/statepolicy -count=1 -timeout 120s -p=1` |
| staterecovery | `& $leafGo test ./internal/nodecontrol/staterecovery -count=1 -timeout 120s -p=1` |
| statetime | `& $leafGo test ./internal/nodecontrol/statetime -count=1 -timeout 120s -p=1` |
| serving | `& $leafGo test ./internal/nodecontrol/serving -count=1 -timeout 120s -p=1` |
| contracts | `& $leafGo test ./internal/nodecontrol/contracts -count=1 -timeout 120s -p=1` |
| readiness | `& $leafGo test ./internal/readiness -count=1 -timeout 120s -p=1` |
| authority（限定） | `& $leafGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1 -timeout 120s -p=1` |
| store（仅编译） | `& $leafGo test ./internal/store -run '^$' -count=1 -timeout 120s -p=1` |

- [x] **Step 4: 按用户选定方法做最终独立只读审查。** Native：作者逐任务实现，不做逐任务独立审查；17项过门后才派一名全新上下文、最强可用 reviewer（派发时核对名单；当前 gpt-6-astra/xhigh），只读本增量。Subagent-driven：每任务新实现者与独立审查，通过后下一任务，最终另做完整增量审查。提供批准spec/plan、20文件、基线实际差异、全部证据及五项 Review Focus。reviewer 不改文件、不运行测试/联网/外部诊断、不委派；旧审查不能替代。本轮不派代理。
- [x] **Step 5: 作者裁定审查问题。** Critical/Important 仅在新包最小修正，每次代码/测试变更重跑完整17项并更新冻结；超范围先修订批准。Minor 修复或注明依据/限制；修正后让同 reviewer 只读复查修正点，记录覆盖差异，不把旧审查宣称覆盖后改源码。
- [x] **Step 6: 写真实执行交付。** apply_patch 保存完整失败/通过/实际退出、审查裁定、UTC/哈希、格式检查；有限更新本增量状态。明确未认证、六项 Pending、旧总体验收阻塞及无跨电脑 Git 检查点；缺失证据显式列出，不补造。
- [x] **Step 7: 最终只读核对与交付。** 20新源码/旧文件/向量与最后review/test冻结相同，branch/HEAD/index保持原基线；所有会话结束。仅本机成果，不暂存/提交/推送/合并/删除/清理。

## 作者自检与执行交接

覆盖：设计第1节→Task 1/4/5；第2节→Task 4/5；第3节→Task 1/2/4；第4.1节→Task 2/4；第4.2节→Task 3/4；第5节→Task 1/4；第6节→Task 1/2/4；第7节→各任务测试/Task5矩阵；第8–9节→本轮仅计划及最终限制。五项 Review Focus 均有命名测试；固定类型、私有helper、文件与selector一致，无 Parse stub/新能力或未定义的邻任务接口。

五任务各自可测，10产品+10测试=20新Go文件，旧Go修改数=0，15完整包+2限定项=17命令。每任务有真实 RED/GREEN 和证据步骤；没有提前勾选，没有把 cleanup 源码检查当内存或认证证明。

自检是计划质量检查，不是执行结果。请用户审阅本计划并选择执行方式：建议 Native，因为五任务共享有限 DER/owner 接口、范围仅隔离包且没有运行接入；可减少逐任务上下文成本，最终仍保留独立审查。也可选择 Subagent-driven，换取逐任务独立审查。选定前不调用实施技能、不写产品代码、不跑测试、不派代理；批准计划不包含 Git 集成、生产接入或扩大验收。

