# Node Leaf Declared Key Identifiers Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增独立 nodeleafkeys 离线关系校验包，证明 SKI/AKI 与 exact subject/声明 issuer SPKI 的固定派生关系，不证明签发者可信，始终保留旧六项 Pending。

**Architecture:** 先检查外层预算并独立拥有三个输入 buffer，组合既有 nodeleaf.Parse，检查声明 issuer SPKI 的固定表示子集，再比较完整 SPKI SHA-256 的两个 32-byte 标识。旧 nodeleaf 只增加针对 exact 新包的一处消费者例外；没有 CA/链/时间/provider/主状态或运行接入。

**Tech Stack:** Go 1.26.5、标准库、既有 nodeleaf/contracts.Digest/github.com/google/uuid v1.6.0；PowerShell、apply_patch、离线缓存；无新依赖。

**Spec:** [已批准书面附录](../specs/2026-10-11-node-leaf-key-identifiers-design.md)，用户于 2026-10-11 批准；同时完整阅读其引用的 [旧 nodeleaf 设计](../specs/2026-10-11-node-leaf-material-design.md)。旧设计的算法未冻结/禁止外部消费者规则，只在本附录明确列出的新关系层和唯一例外上得到补充，不改旧形状语义。

状态：用户于 2026-10-11 选择 1，批准本计划并按 Native 执行。五任务已完成；18 项最终限定命令 exit 0，fresh 独立只读审查 0 Critical / 0 Important / 0 Minor；仅本机交付，无 Git 写入。正文的“本轮”“未来”“选定前”保留起草时说明，当前状态与已执行 checkbox 以本页首及执行记录为准。需求正文与范围不变。见 [真实执行记录](../../roadmap/2026-10-11-node-leaf-key-identifiers-execution.md)，包含全部 43 次测试及两次格式检查的实际结果和两次诊断，不把失败或预检计为最终验收。

## Global Constraints

- 复用 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`，分支 `codex/c12-b01-task9-coordinator`；保留全部旧未提交进度，不新建/清理 worktree。
- 不暂存、提交、推送、合并、重置、删分支、修改 safe.directory 或永久环境；每任务的真实证据记录替代技能模板 Commit 步骤。跨电脑远程检查点未产生。
- `Leaf nodeleaf.Input` 与 `DeclaredIssuerSPKI []byte` 是全部输入；不接受 Summary/verified/预计算 AKI/issuer ID/CA 链或回调代替材料。
- leaf DER 1–65,536 bytes、expected SPKI 1–512 bytes、声明 issuer SPKI 1–8,192 bytes；总输入最多 74,240 bytes，不是总堆内存上限。context 优先，其次任一超限，再检查空值与 UUID/domain；全部外层预算先于克隆/解析。
- NodeID 非零、RFC4122 variant；domain 小写 ASCII DNS，总长 1–253，每 label 1–63，首尾 `[a-z0-9]`、内部 `[a-z0-9-]`。不 trim/case-fold/IDNA；原 lower URI/SPKI/v3/serial/time/六扩展规则全部不变。
- profile 标记 `node-key-id.spki-sha256.v1`；SKI=SHA256(exact leaf RawSubjectPublicKeyInfo DER)，AKI=SHA256(exact DeclaredIssuerSPKI DER)，均为完整 32-byte 二进制值，无截断、前缀或算法协商/fallback。
- extension 内值固定 SKI=`04 20`+32 bytes，AKI=`30 22 80 20`+32 bytes；outer inventory/noncritical/default/完整消费由 lower 保证。相同 SPKI 的 SKI/AKI 可以相同。
- issuer 表示子集仅 RSA、ECDSA P-256/P-384/P-521、Ed25519；Go 完整解析后 MarshalPKIXPublicKey 必须与 owned 输入逐字一致，重编码仅用于拒绝不一致，绝不修复/替代输入。leaf 仍仅 P-256。
- 不接受 DSA/X25519/P-224/显式 EC 参数/压缩 EC 点/未知类型/PEM/裸 PKCS#1/拼接/尾随；不新增 RSA 位数或 CA 算法政策。支持表示不等于批准该 CA。
- 两个 key ID 先按 SKI、AKI 检查长度，再按 SKI、AKI 比较；错误不包装 parser 动态信息、不输出输入/标识/摘要、不记录日志。
- `Proof=NodeLeafDeclaredKeyIdentifiersBound` 只证明声明对应；nested Leaf 保持 `NodeLeafMaterialShapeValid` 和原固定六项 Pending，包含 KeyIdentifierDerivationAndIssuerBinding。不得删除/拆空责任或输出 Ready/Trusted/Authorize。
- 现有 CA 的 20-byte/其他算法 SKI 不会自动迁移；本包没有 CA DER，不能证明 CA SKI 兼容。未来 CA profile/provider/签名/信任包认证须另行设计，不在本计划完成。
- 先登记 owner 清理再复制；共享包装生命周期、detached Summary、nil-safe/幂等 Clear；失败/取消/正常 panic unwind 清理自有可达 buffers/summary，panic 传播。不承诺不可达分配/immutable strings/调用方/fatal OOM 擦除或并发安全。
- 唯一旧 Go 文件是 `internal/nodecontrol/nodeleaf/boundary_test.go`：仅实现 exact nodeleafkeys 目录且 package=nodeleafkeys 的消费者例外，并为同一例外补充回归 cases；原例外、原负例和其他旧守卫不变。没有旧产品改动。
- 新包不得被任何其他产品包消费；禁止 Verify/CheckSignature*/CertPool/system roots/TLS/签发/私钥/网络/DB/文件 I/O/时钟/随机/hook/goroutine/主状态转换。测试专属签发、读取源码和 Context 实验不算产品能力。
- 仅限定离线回归；不跑全仓、race、DB、Docker、硬件、C11/C12 PITR 或生产 conformance。Task10/B01/C12/main 的既有验收阻塞不解除。
- 编辑用 apply_patch；gofmt 仅明确列出的 17 个新 Go 文件加上述 1 个旧测试文件；未结束的测试进程必须等实际退出，不能重复启动同项或提前声称通过。

## Review Focus

1. Go parser 可接受的 inner trailing/参数偏差不能因解析成功而放行；canonical marshal 只比较，不替代输入（Task 2：TestKeysIssuerCanonical）。
2. subject 与 issuer 恰好同 key 时 ID 相等不应被误拒绝，也不应升格成 self-signing/信任（Task 3/4：TestKeysSameSPKI、TestKeysPartialProofOnly）。
3. malformed lower leaf 加 malformed issuer、错误 SKI 加短 AKI、操作结束时取消，必须保持固定错误优先级（Task 4：TestKeysErrorPrecedence、TestKeysLateCancellation）。
4. 三份共享 backing/subslice、复制包装、第二/第三次复制前 panic 及 Clear 后读取，不能泄漏/保留有效 owner（Task 1/4：TestKeysOwnedAliases、TestKeysMaterialLifecycle、TestKeysCleanupRegistration）。
5. exact 新包例外不能成为路径后缀/前缀通配、外部测试包或下游运行消费者的通行证（Task 1/4：旧 TestLeafConsumers 新 cases、TestKeysConsumers、TestKeysCapabilityBoundary）。

---

## 文件结构与固定接口

所有新 Go 文件都在 `internal/nodecontrol/nodeleafkeys/`；不新增 testdata/捕获向量。固定 9 产品+8测试=17新文件：

| 产品文件 | 责任 |
| --- | --- |
| doc.go / input.go | 部分证据边界 / Input、KeyIdentifier、Proof、Summary、私有 profile 标记 |
| errors.go / budget.go | 五个 lower error 同身份别名及一个新错误 / 无复制解析的 preflight、domain 检查 |
| owned.go / material.go | 三 buffer 所有权及失败清理 / 共享 owner、detached Summary、Clear |
| issuer_spki.go | 声明 issuer 表示子集、完整解析与 canonical marshal 比较 |
| key_identifiers.go | 检查 lower Summary、同 owned leaf snapshot 与固定派生/比较 |
| evaluate.go | 唯一 public Evaluate、阶段 context、组合 lower 和成功移交 |

测试文件：`fixtures_test.go`、`budget_test.go`、`owned_test.go`、`material_test.go`、`issuer_spki_test.go`、`key_identifiers_test.go`、`evaluate_test.go`、`boundary_test.go`。

唯一旧 Go 修改：`internal/nodecontrol/nodeleaf/boundary_test.go` 的 `leafConsumerAllowed` 及 `TestLeafConsumers` 中验证同一例外的新增 cases；旧 cases 全保留，不改变其期望。这是附录 §8 所列例外的实现与验证，不是额外允许修改其他守卫。

公开类型与签名固定为：

```go
type Input struct {
    Leaf nodeleaf.Input
    DeclaredIssuerSPKI []byte
}
type KeyIdentifier [32]byte
type Proof uint8 // None=0; NodeLeafDeclaredKeyIdentifiersBound=1
type Summary struct {
    Proof Proof
    Leaf nodeleaf.Summary
    SubjectKeyIdentifier, AuthorityKeyIdentifier KeyIdentifier
    DeclaredIssuerSPKIDERDigest contracts.Digest
}
type Material struct { owned *owner }
func Evaluate(ctx context.Context, in Input) (*Material, error)
func (m *Material) Summary() Summary
func (m *Material) Clear()
```

私有 `owner` 仅有 `der, expectedSPKI, declaredIssuerSPKI []byte; summary Summary; valid bool`。三个 buffer 独立克隆，不能以 slice 别名复用；临时 lower Material/解析指针均不保留。`const keyIdentifierProfile = "node-key-id.spki-sha256.v1"`；没有导出算法选择器或 Derive 接口。

`ErrInvalidInput/ErrLimitExceeded/ErrInvalidMaterial/ErrUnsupportedLeafProfile/ErrBindingMismatch` 分别直接引用同名 nodeleaf sentinel；新错误消息固定为 `node leaf keys: unsupported key identifier profile`。全部 errors.Is 身份固定，不包装 parser 原错误。lower Proof/Pending 或 snapshot 摘要矛盾统一 ErrInvalidMaterial；不是新增第七个领域错误。

实施期间才新增 `.superpowers/sdd/2026-10-11-node-leaf-key-identifiers/` 专属证据/cache 和 `docs/roadmap/2026-10-11-node-leaf-key-identifiers-execution.md`。最终只有限更新本附录/本计划状态、current-status 的本增量摘要；其他旧执行记录/设计/计划不改。本轮不创建上述实施产物。

## 环境、真实留痕与停止条件

未来测试均在上述 worktree、PowerShell、login=false，每进程使用完整 wrapper：

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
$env:GOCACHE='D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/.superpowers/sdd/2026-10-11-node-leaf-key-identifiers/go-cache'
$keysGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
$keysFmt='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/gofmt.exe'
```

路径已只读确认存在，未运行。规划前 HEAD=`6752b9612a00b9d6b0bf87ae8dd95063f6b056ee`、index SHA256=`1CBA46BEF4C1AFDC9AE7938146909F0D2D641A0D712CE18036B60FE5B8673D12`；319 条原 dirty status、714 个非忽略文件组成当前只读核对面。实施开始必须重新 capture 全部基线，不能把此规划快照或历史 17 项通过结果当本增量的执行证据。

每次 RED/GREEN/失败/补充/最终回归都在启动前记录 source hashes 与开始 UTC，完成后记录 source hashes、结束 UTC、完整环境+命令、workdir/login、selector、实际输出/最终退出码和任务 ID。用 apply_patch 保存；记录较晚应披露实际时间，不能补造缺失证据。已有审查和旧增量的测试不代替本次结果。

每条 test 固定 `-count=1 -timeout 120s -p=1`，命名 selector 的 RED/GREEN 相同，整包另行记录。预期 RED 只能来自缺本任务声明或真实行为断言，不能是 fixture/语法/缓存/权限错误。会话未结束就等待原会话；没有实际退出码不验收。超时或非预期失败保留证据后诊断，不能静默加时、无界重试或下载依赖。

Git 只读使用 GIT_OPTIONAL_LOCKS=0、core.fsmonitor=false；需要提权按权限机制，不改 Git 信任配置。任何超范围旧文件需求、算法/接口变化、工具拒绝或规格矛盾，完成不受影响工作后报告并修订批准范围；不绕过平台拒绝。

## Task 1: 固定类型、三 buffer 生命周期与狭窄消费者例外

**Files:** 新建 doc.go/input.go/errors.go/budget.go/owned.go/material.go、budget_test.go/owned_test.go/material_test.go；修改唯一旧 boundary_test.go 的例外和对应 cases。不存在 Evaluate stub。

**Interfaces:** 产生上列公开类型/error/Material 方法；私有 `preflight(ctx context.Context, in Input) error`、`validTrustDomain(value string) bool`、`ownInput(in Input) *owner`、`(*owner).clear()`。preflight 不解析/克隆；ownInput 初始 valid=false。

- [x] **Step 1: 建立获准执行基线。** 核对书面计划批准与方法选择，完整读两 spec/本计划；保存全部既有文件/向量哈希、Git branch/HEAD/index/status、批准文档哈希与空台账，确认新目录/证据目录无未知既有文件。仅此时建立专属证据/cache。
- [x] **Step 2: 写生命周期/预算失败测试。** TestKeysBudget 覆盖 DER 0/1/65536/65537、expected 0/1/512/513、issuer 0/1/8192/8193，及 context→任一超限→空值/UUID/domain；TestKeysScope 覆盖 RFC4122 variant、253/254 domain 和63/64 label、大小写/Unicode/空白/尾点；TestKeysFixedErrors 检查五 alias 身份与新消息。TestKeysOwnedAliases 对三输入共享 backing/subslice 检查独立所有权；TestKeysMaterialLifecycle/TestKeysNilMaterial 检查共享包装、detached Summary、两 Proof 清零及幂等/nil Clear。
  ```go
  if ErrInvalidInput != nodeleaf.ErrInvalidInput || ErrBindingMismatch != nodeleaf.ErrBindingMismatch { t.Fatal("lower identity changed") }
  err := preflight(context.Background(), Input{DeclaredIssuerSPKI: make([]byte, 8193)})
  if !errors.Is(err, ErrLimitExceeded) { t.Fatal("limit must precede empty leaf") }
  o := ownInput(Input{Leaf: nodeleaf.Input{DER: []byte{1, 2}, ExpectedSPKI: []byte{3, 4}}, DeclaredIssuerSPKI: []byte{5, 6}})
  a, b, c := o.der, o.expectedSPKI, o.declaredIssuerSPKI
  o.valid = true // 仅内部 lifecycle 实验，不是 Evaluate 正例
  m := &Material{owned: o}; copied := *m; m.Clear()
  if copied.Summary() != (Summary{}) || !bytes.Equal(a, []byte{0,0}) || !bytes.Equal(b, []byte{0,0}) || !bytes.Equal(c, []byte{0,0}) { t.Fatal("stale owner") }
  ```
- [x] **Step 3: 为唯一旧例外先加失败 cases。** 在旧 TestLeafConsumers 原表后追加 exact worktree-root-relative 和 absolute nodeleafkeys/a.go（package=nodeleafkeys）通过；nodeleafkeys/a_test.go 同包通过；nodeleafkeys_extra、nodeleafkeys/child、else/internal/nodecontrol/nodeleafkeys、nodeleafkeys_test package、任意兄弟/外部测试路径拒绝；原表不删除/改值。
- [x] **Step 4: 分别记录真实 RED。** 新包 selector=`^(TestKeysBudget|TestKeysScope|TestKeysFixedErrors|TestKeysOwnedAliases|TestKeysMaterialLifecycle|TestKeysNilMaterial)$`；旧包 selector=`^TestLeafConsumers$`。均用固定 test flags；新包预期缺声明，旧例外正例预期被现状拒绝。不是制造其他旧测试失败。
- [x] **Step 5: 最小实现合同与例外。** 新文件实现 preflight、三 buffer cleanup-before-copy、shared owner 与零值 Summary。旧 leafConsumerAllowed 保留原分支，仅新增：filepath.Abs/Clean 后相对实际 walk root（旧测试的 ../../../）做 filepath.Rel，要求 rel 的目录恰为 internal/nodecontrol/nodeleafkeys、package 恰 nodeleafkeys；Rel/路径解析错误 fail closed，不用 HasPrefix/HasSuffix 放宽目录。
- [x] **Step 6: 记录 GREEN。** 重跑两个同 selector，再分别完整新包和完整旧 nodeleaf，均 exit 0。新包此时只证明合同/lifecycle，不声称材料关系入口完成。
- [x] **Step 7: 保存阶段证据。** 记录实际 RED/GREEN/UTC/pre-post hashes及唯一旧文件 diff；除此旧源码/向量全冻结，不提交。

## Task 2: 独立离线 fixture 与声明 issuer SPKI 表示

**Files:** 新建 fixtures_test.go、issuer_spki.go、issuer_spki_test.go；必要修正仅 Task 1 新文件。

**Interfaces:** 消费固定错误；产生私有 `checkDeclaredIssuerSPKI(der []byte) error`。test-only helpers：`keysFixture(t testing.TB) Input`、`keysIssuerSPKI(t testing.TB, kind string) []byte`、`keysWithIDs(t testing.TB, in Input, ski, aki []byte) Input`、`keysWithSignatureBitFlipped(t testing.TB, in Input) Input`、`rfcSPKI(t testing.TB) []byte`（解码 spec §7.1 固定 hex，不调用产品 helper）。kind 仅 p256/p384/p521/rsa/ed25519；后两个只作编码材料，不作生产 CA。

- [x] **Step 1: 先写并独立核对 fixture。** subject/真实测试 issuer P-256 scalar=2/1；node=11111111-2222-4333-8444-555555555555、domain=node.test、serial=128、NotBefore=2026-10-11T00:00:00Z、NotAfter=次日；显式六 ExtraExtensions 与32-byte IDs，同时显式设置测试 issuer 模板 SubjectKeyId 为该 issuer SPKI 的全32-byte摘要，不依赖 Go 自动生成。expected/issuer SPKI 由测试标准库生成，ID 期望由独立标准库 SHA-256 求值，不调用产品 helper；RFC SPKI/hash 常量直接复制批准 spec §7.1。TestKeysFixtureControls 检查原始 SPKI/六内值/日期/独立摘要并离线验签正例；允许签名随机，不声称每次 leaf DER 固定。不改旧 fixture。
- [x] **Step 2: 先运行 fixture control。** `& $keysGo test ./internal/nodecontrol/nodeleafkeys -run '^TestKeysFixtureControls$' -count=1 -timeout 120s -p=1` 必须 exit 0 后才新增调用缺失产品 helper 的测试；fixture 失败不算产品 RED。
- [x] **Step 3: 写 issuer 失败测试。** TestKeysIssuerKinds 五种表示接受，P-224/X25519/DSA/未知类型拒绝；TestKeysIssuerCanonical 覆盖尾随、PEM/裸 PKCS#1、非最小 length、EC 参数 trailing/显式参数/压缩点/非法点、RSA key 内 trailing、Ed25519 参数/长度、BIT STRING 未使用位。原始变体通过测试 ASN.1 RawValue 独立重建；至少一个 Go ParsePKIXPublicKey 已接受但 marshal 与原字节不同的变体，helper 必须 ErrInvalidInput；不将预期值建立在产品 helper 上。TestKeysIssuerBudgetEdge 用 test-only RSA 表示构造 exact 8192-byte SPKI（8154-byte modulus magnitude，首字节80末字节01，E=65537），独立断言实际长度与 parse/marshal 完整对应；8193 预算拒绝由 Task1 preflight 测试，不在 helper 暗增预算。
  ```go
  for _, kind := range []string{"p256", "p384", "p521", "rsa", "ed25519"} {
      raw := keysIssuerSPKI(t, kind)
      if err := checkDeclaredIssuerSPKI(raw); err != nil { t.Fatalf("%s: %v", kind, err) }
      if err := checkDeclaredIssuerSPKI(append(bytes.Clone(raw), 0)); !errors.Is(err, ErrInvalidInput) { t.Fatal("trailing SPKI accepted") }
  }
  ```
- [x] **Step 4: 记录 RED。** selector=`^(TestKeysIssuerKinds|TestKeysIssuerCanonical|TestKeysIssuerBudgetEdge)$`；固定 flags，预期缺 helper 或真实断言失败。
- [x] **Step 5: 最小实现 issuer_spki.go。** ParsePKIXPublicKey 后有限 type switch：RSA、指定 named ECDSA（合法非零点）、32-byte Ed25519；未知/其他失败。MarshalPKIXPublicKey 后立即登记 buffer 清零责任，逐字比较 owned 原 bytes；只返回固定 ErrInvalidInput。不保留解析对象或改变材料；不验证签名，不增 RSA/CA policy。
- [x] **Step 6: 记录 GREEN 与冻结。** 同 selector及完整新包分别 exit 0；保存真实试验/parser接受事实、marshal差异及源哈希，不提交。

## Task 3: exact leaf snapshot、固定派生与双标识关系

**Files:** 新建 key_identifiers.go、key_identifiers_test.go；仅扩充新 fixtures_test.go 必要变体。

**Interfaces:** 消费 lower Summary、owned DER/expected SPKI 和已检查的声明 SPKI；产生私有 `checkLowerSummary(s nodeleaf.Summary) error`、`checkLeafSnapshot(der, expectedSPKI []byte, lower nodeleaf.Summary) (*x509.Certificate, error)`、`keyIdentifierForSPKI(der []byte) KeyIdentifier`、`checkKeyIdentifiers(cert *x509.Certificate, declaredIssuerSPKI []byte) (KeyIdentifier, KeyIdentifier, contracts.Digest, error)`。临时 Certificate 只活在调用阶段，不能进入 owner。

- [x] **Step 1: 写 snapshot 与 oracle 失败测试。** TestKeysLowerSummary 要求原 Proof/六 Pending 顺序 exact，零/缺项/重排/未知项→InvalidMaterial；TestKeysLeafSnapshot 要求 cert.Raw==owned DER、raw SPKI==owned expected、SHA256两 preimage==lower摘要，尾随或替换材料/摘要→InvalidMaterial。TestKeysRFC7093 用 spec 91-byte SPKI 与固定 hash 检查产品派生，不用产品生成期望。
  ```go
  want := "6d20896ab8bd833b6b66554bd59b20225d8a75a296088148399d7bf763d57405"
  got := keyIdentifierForSPKI(rfcSPKI(t)) // test-only helper，固定 spec bytes
  if hex.EncodeToString(got[:]) != want { t.Fatal("not full SPKI SHA-256") }
  ```
- [x] **Step 2: 写关系/顺序失败测试。** TestKeysIdentifierLengths 覆盖两者各1/20/31/32/33/64；32字节真值成功，其余 UnsupportedKeyIdentifierProfile。TestKeysIdentifierBinding 分别错SKI/AKI、交换公钥 preimage、SHA1/截断SHA256/完整BITSTRING/EC-point/leaf-DER/加前缀/hex/base64变体；32-byte错误→BindingMismatch，其他长度→UnsupportedKeyIdentifierProfile。TestKeysSameSPKI 让两个输入SPKI相同且标识相同，关系接受；TestKeysIdentifierPrecedence 对错32-byte SKI+20-byte AKI先返回 profile 错误，两个32-byte错误返回BindingMismatch，错误时返回zero IDs/digest。相同错误类别不能由输出区分两次比较的先后，SKI优先顺序还由Task4有限源码守卫和fresh review检查，不伪称公共输出证明该顺序。
- [x] **Step 3: 记录 RED。** selector=`^(TestKeysLowerSummary|TestKeysLeafSnapshot|TestKeysRFC7093|TestKeysIdentifierLengths|TestKeysIdentifierBinding|TestKeysSameSPKI|TestKeysIdentifierPrecedence)$`；固定 flags，预期缺声明/真实行为缺失。
- [x] **Step 4: 最小实现 key_identifiers.go。** 检查原 lower 固定 Proof/Pending，完整解析同 owned leaf、两个 exact 字节与摘要对应；不复制完整 leaf profile。私有派生仅 SHA256.Sum256 原始 SPKI；length两项先于compare两项。返回 Subject/Authority ID 与声明 SPKI Digest，不能导出算法函数或自动转换旧数据库/主状态引用。
- [x] **Step 5: 记录 GREEN 与冻结。** 同 selector和完整新包分别 exit 0；记录独立向量/长度/错误顺序，保留 CA trust 责任，不提交。

## Task 4: 唯一 Evaluate、取消清理与 API/消费者守卫

**Files:** 新建 evaluate.go、evaluate_test.go、boundary_test.go；扩充新 owned/material/fixture 测试；必要修正仅已列新产品文件或已批准的唯一旧例外。

**Interfaces:** 消费Task1–3全部私有helper；产生public Evaluate。test-only helper `keysImportsAllowed(source []byte) bool`、`keysCleanupRegistered(source []byte) bool`、`keysConsumerAllowed(path, pkg, imported string) bool`，只用于有限守卫，不增加产品 hook。

- [x] **Step 1: 写入口失败测试。** TestKeysEvaluate/TestKeysEvaluateIssuerKinds 要求 real fixture成功、两个Proof/nested六Pending/ID32/三个摘要正确，返回Summary detached；每种issuer表示分别构造匹配AKI后接受，不能误要求声明issuer验签。TestKeysEvaluateRejections 重放issuer与ID负例，并保留lower URI/SPKI/profile/parser拒绝类型；空结果不带半成品。
  ```go
  in := keysFixture(t)
  m, err := Evaluate(context.Background(), in)
  if err != nil || m == nil { t.Fatal(err) }; defer m.Clear()
  s := m.Summary()
  if s.Proof != NodeLeafDeclaredKeyIdentifiersBound || s.Leaf.Proof != nodeleaf.NodeLeafMaterialShapeValid { t.Fatal("proof") }
  want := [6]nodeleaf.Obligation{nodeleaf.LeafSignatureAndIssuerTrust, nodeleaf.KeyIdentifierDerivationAndIssuerBinding, nodeleaf.TrustedCurrentTime, nodeleaf.IdentityAndAuthorityProvenance, nodeleaf.PrivateKeyPossession, nodeleaf.MainStateAndRuntimeBinding}
  if s.Leaf.Pending != want || s.DeclaredIssuerSPKIDERDigest != contracts.Digest(sha256.Sum256(in.DeclaredIssuerSPKI)) { t.Fatal("pending/preimage") }
  ```
- [x] **Step 2: 写非认证、优先级、取消与别名测试。** TestKeysPartialProofOnly 使用bad-crypto-signature/固定expired或future窗口/未经信任包但AKI匹配的声明issuer，以及same-key，始终只得partial六Pending；不将损坏签名语法当密码学签名错误。TestKeysErrorPrecedence 覆盖bad lower+bad issuer、lower binding mismatch+bad issuer、valid lower+bad issuer+bad IDs、错SKI+短AKI，固定类型。TestKeysCancellation/TestKeysLateCancellation 用test-only Context.Err先记录完整成功轨迹，再对每个实际检查点分别返回Canceled/DeadlineExceeded或panic，验证最终检查点及取消覆盖阶段普通错误；不硬编码调用次数、不Parallel、不用产品callback。TestKeysEvaluateOwnership 对共享subslice和Evaluate后caller修改，Summary/owner不变；Clear不影响caller与detached值。
- [x] **Step 3: 写能力与清理守卫失败测试。** TestKeysProductImports/TestKeysCapabilityBoundary 固定17文件、5公开类型/确切字段、6error vars/2Proof常量、唯一Evaluate/2Material方法；dot/blank/duplicate imports、alias冲突、forbidden方法值、goroutine/global hook/序列化或DER方法拒绝；有限源码检查两项length先于两项compare、SKI先于AKI，内存逆序变体拒绝。TestKeysConsumers 只允许本新包同包测试导入nodeleafkeys，其他product或external_test拒绝；对nodeleaf导入仅允许其原测试例外与exact新包。TestKeysCleanupRegistration 检查preflight先于ownInput，ownInput首次clone前清理、Evaluate得owner和lower后立即登记清理、marshal buffer清理、最后ctx→valid/transfer→return；内存source变体missing/late/conditional/wrong-owner defer及提前transfer都拒绝，不能改真实源码制造RED。
- [x] **Step 4: 记录 RED。** selector=`^(TestKeysEvaluate|TestKeysEvaluateIssuerKinds|TestKeysEvaluateRejections|TestKeysPartialProofOnly|TestKeysErrorPrecedence|TestKeysCancellation|TestKeysLateCancellation|TestKeysEvaluateOwnership|TestKeysProductImports|TestKeysCapabilityBoundary|TestKeysConsumers|TestKeysCleanupRegistration)$`；固定 flags，真实缺入口/守卫或行为断言失败。
- [x] **Step 5: 最小实现 Evaluate 与守卫。** 按spec §6.1固定七阶段：preflight→owned三buffer→lower Parse/立即defer Clear与Summary检查→issuer表示→同owned snapshot→两length/两compare与摘要→最终ctx/有效owner移交。lower Parse使用owned bytes+原纯值NodeID/domain，绝不caller slices。操作完成先查ctx再返回该阶段error，所有temporary解析值不保留；shared Clear与正常panic清理不改变partial强度。守卫扫描imports/API/consumer/cleanup有限源码形状，不冒称形式化证明。
- [x] **Step 6: 记录 GREEN。** 同selector、完整新包、完整nodeleaf分别exit0；failed private-owner erasure主要是源码责任检查，和成功owner captured backing实际清零证据分开报告。
- [x] **Step 7: 保存入口与源冻结。** 17新文件及唯一旧守卫差异完整记录，不接入任何consumer/provider/state/runtime，不提交。

产品 imports/member 白名单只允许：bytes(Clone/Equal)、context(Context)、errors(New)、crypto/sha256(Sum256)、crypto/rsa(PublicKey)、crypto/ecdsa(PublicKey)、crypto/ed25519(PublicKey/PublicKeySize)、crypto/elliptic(P256/P384/P521)、crypto/x509(Certificate/ParseCertificate/ParsePKIXPublicKey/MarshalPKIXPublicKey)、uuid(UUID/Nil/RFC4122)、contracts(Digest)、nodeleaf(Input/Summary/Obligation/Parse/两个Proof值/六Obligation值/五error)。cert/key/point的普通字段和上述关系检查所需的IsOnCurve/Sign方法允许；lower Material只用Summary/Clear，不获取lower DER、不保留对象。禁止Verify/CheckSignature*/池/签发/时钟及前述I/O能力，包括alias或方法值绕过。包级可变值仅6error vars；类型表/Pending数组在函数中生成。

## Task 5: 十八项限定回归、一次 fresh review 与本机交付

**Files:** 不增加产品能力；维护专属真实证据、最终执行记录及本增量状态文档；旧执行记录/旧源码/向量除单一例外全部冻结。

**Interfaces:** 消费Task1–4实际RED/GREEN与18个Go文件差异；输出限定验证/审查裁定/最终冻结，不赋予Git或生产集成权限。

- [x] **Step 1: 精确格式范围。** 明确列出17新文件+唯一旧boundary_test.go共18路径，断言无其他path/unknown file；运行 `& $keysFmt -w $keysFormatFiles`。不能格式化其他旧文件。
- [x] **Step 2: 检查格式与旧冻结。** `& $keysFmt -l $keysFormatFiles` exit0且输出为空；保存18文件source freeze，旧guard diff仍只例外/cases，其余基线文件/向量无漂移。
- [x] **Step 3: 每条独立运行下列18项。** 所有条目应用完整离线wrapper与固定flags，记录每条实际最终退出码与pre/post hashes；不并发写同证据/cache或重复未结束会话。16完整包+authority限定+store仅编译必须全部exit0；store不算运行验收。

| 范围 | 固定命令 |
| --- | --- |
| nodeleafkeys | `& $keysGo test ./internal/nodecontrol/nodeleafkeys -count=1 -timeout 120s -p=1` |
| nodeleaf | `& $keysGo test ./internal/nodecontrol/nodeleaf -count=1 -timeout 120s -p=1` |
| statetimebinding | `& $keysGo test ./internal/nodecontrol/statetimebinding -count=1 -timeout 120s -p=1` |
| statepolicybinding | `& $keysGo test ./internal/nodecontrol/statepolicybinding -count=1 -timeout 120s -p=1` |
| statebinding | `& $keysGo test ./internal/nodecontrol/statebinding -count=1 -timeout 120s -p=1` |
| statecontract | `& $keysGo test ./internal/nodecontrol/statecontract -count=1 -timeout 120s -p=1` |
| statematerials | `& $keysGo test ./internal/nodecontrol/statematerials -count=1 -timeout 120s -p=1` |
| statecomposition | `& $keysGo test ./internal/nodecontrol/statecomposition -count=1 -timeout 120s -p=1` |
| stateformat | `& $keysGo test ./internal/nodecontrol/stateformat -count=1 -timeout 120s -p=1` |
| stateverify | `& $keysGo test ./internal/nodecontrol/stateverify -count=1 -timeout 120s -p=1` |
| statepolicy | `& $keysGo test ./internal/nodecontrol/statepolicy -count=1 -timeout 120s -p=1` |
| staterecovery | `& $keysGo test ./internal/nodecontrol/staterecovery -count=1 -timeout 120s -p=1` |
| statetime | `& $keysGo test ./internal/nodecontrol/statetime -count=1 -timeout 120s -p=1` |
| serving | `& $keysGo test ./internal/nodecontrol/serving -count=1 -timeout 120s -p=1` |
| contracts | `& $keysGo test ./internal/nodecontrol/contracts -count=1 -timeout 120s -p=1` |
| readiness | `& $keysGo test ./internal/readiness -count=1 -timeout 120s -p=1` |
| authority（限定） | `& $keysGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1 -timeout 120s -p=1` |
| store（仅编译） | `& $keysGo test ./internal/store -run '^$' -count=1 -timeout 120s -p=1` |

- [x] **Step 4: 按选定方法独立只读审查。** Native：作者逐任务实现，18项过门后才派一次fresh-context最强可用reviewer；Subagent-driven：每任务新实现者/独立reviewer过门，再最终完整增量审查。派发前核对当前model名单，reviewer只读批准spec/plan、17新文件+旧guard diff、基线/全证据、五Review Focus；不改文件、跑测试、联网、委派或用旧审查代替。本轮不派代理。
- [x] **Step 5: 作者裁定并验证修正。** Critical/Important仅在获准18Go文件内最小修复；每次产品/测试改动重跑完整18项、重新freeze，同reviewer只读复查修改点。Minor修复或保留有依据的限制，不补造历史证据；超范围先修订批准。明确最后review覆盖的文件hash，不能把review前结论应用到后改源码。
- [x] **Step 6: 写真实最终执行记录。** 用apply_patch记录所有尝试/UTC/full command/workdir/login/输出exit/hash、格式结果、review与裁定；有限更新本spec/plan状态和current-status本增量摘要。标明六Pending、声明来源非信任、CA兼容UNKNOWN、总体验收仍阻塞、无远程Git检查点；未运行/缺证据明确披露。
- [x] **Step 7: 最终只读交付。** 最后test/review冻结相同，唯一旧guard diff合规，其他旧文件/向量与执行基线一致，branch/HEAD/index保持原样，全部会话结束。只交本机成果，不暂存/提交/推送/合并/删除/清理。

## 作者自检与执行交接

覆盖：spec §1→Task3/4/5；§2–3→Task3/4；§4→Task4非认证测试与Task5限制；§5→Task1/2/4；§6→Task1/3/4；§7→Task2独立fixture/RFC向量及Task3/4全部拒绝面；§8→Task1例外及Task4consumer守卫；§9→阶段约束/Task5交付。五Review Focus均有命名测试，接口/文件/selector一致，只有一处旧守卫功能变化及同例外cases，无未知neighbor helper或新增运行能力。

五任务各有独立测试周期与证据记录；9产品+8测试=17新Go文件，唯一旧测试文件=1，共18格式/review文件；16完整包+2限定项=18最终命令。上述 checkbox 已依据本次真实执行证据更新；起草时的计划自检不是测试或审查证据。

执行选择留痕：用户已选择下列 1（Native）；以下为原交接选项：

1. **Native（推荐）**：作者在当前会话逐任务实现，最终一次fresh独立只读审查。五任务共享owner/snapshot接口，且不进入生产运行，适合减少逐任务上下文成本；仍保留全部TDD与最终审查。
2. **Subagent-driven**：每任务fresh实现者与独立reviewer，通过再进行下一任务，最终完整审查；逐任务独立性更高，成本也更高。

选定前不调用实施技能、不写产品代码、不跑测试、不派代理。批准本计划/方法不批准Git写入、扩展旧修改范围或生产接入；本轮设计批准不替代对本计划的书面审阅。
