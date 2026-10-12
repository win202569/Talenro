# Task 10 Desired Canonical Format and Binding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现隔离的 C1.2 严格格式组件和私有 desired 投影绑定，不宣称验签或生产授权。

**Architecture:** 新 stateformat 纯包独占四类制品的 wire 解析、局部语义、封装和 transcript。serving 新增纯绑定函数消费格式结果，不修改现有 reader/adapter 的成功语义。全部结果明确为格式证据，失败丢弃自有暂存。

**Tech Stack:** 固定 Go 1.26.5，已锁定 github.com/gowebpki/jcs v1.0.1、google/uuid，标准库，Windows PowerShell 离线测试。

**Spec:** 完整阅读 `docs/superpowers/specs/2026-10-09-task10-desired-canonical-binding-design.md` 与 `docs/superpowers/specs/2026-10-09-node-state-format-contract-addendum.md`。用户已批准两份书面设计；本实施计划已于本会话获用户确认并完成本地执行（两项新 Minor 延期）。精确 wire 字段以附录为准。

## Global Constraints

- 复用 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`；保持既有 dirty 文件，分支不得是 main。
- 不暂存/提交/推送/合并/删除 worktree。PowerShell 记录 ledger，覆盖技能默认 Bash、commit 和清理步骤。
- 附录批准是局部开发依据，不等于生产规范登记。只读查发布记录；没有权威记录时记 UNKNOWN，继续仅隔离 codec/测试和私有纯绑定，禁止 DB 写入、网络发布或旧制品迁移。
- 只有下列文件清单中的新增源码/测试和进度文档可变；SQL/schema/store、现有 authority/certificate/reader/adapter、strictjson/trust、API/generated、四个依赖锁、基础规范/B02 旧计划/c12-spec-set 保持不变。
- 不运行 DB/Docker/C11/C12/bootstrap/PITR/race/全仓测试，不改变安全设置或绕过拒绝，不下载依赖。
- 本计划新增明确的测试许可请求：stateformat 全包离线单元测试；其余仍限 serving/contracts/readiness、authority `^(TestBoundRead|TestCoordinator)`、store `-run '^$'` 编译。
- P ≤65536 bytes；E ≤262144 bytes；payload_b64url ≤87382 chars；base64url(E) ≤349526 chars。实际解码长度与 canonical re-encode 同时检查；网络入口仍不接线。
- 根容器深度=1，最大16；所有字段 required，只有 metadata 指定 revocation 分支可 null；禁止未知/重复/折叠别名、非法 Unicode、尾随值和非规范原文。
- 所有整数 wire 为精确十进制 string，范围及字段局部上限按附录；没有 float64 中转或 number/string 双接受。
- root 1..5 keys，初始 root=1 且1 key/threshold1；rotation body new=previous+1；metadata 1..64 keys、0..4096 revoked IDs；每组 share 0..64；desired 0..8 processes、每 slot 1..64 metrics。
- D/K/S/T/UUID/ID/Build、十项 capacity bounds、时间窗口/coverage/null 分支完全采用附录 §3–5。格式通过不证明签名、当前授权、库存/profile、历史或 deployment 信任。
- 只保留 bytes/固定字段的自有副本；不能声称擦除不可达副本或不可变字符串。无 production hook、可替换 verifier 或 caller trusted bool。

## Review Focus

1. metadata 中显式 null 与缺字段、required=false 与未提供不能混淆；Task 1/3 exact-presence 测试。
2. >2^53 的字符串整数、转义后重复键、孤立 surrogate 不能被标准 decoder 修复；Task 1 primitive/strict 测试。
3. 合法容器 count 不代表满足 byte budget；过长 revoked ledger 不可被截断，outer 的 base64 两层边界不可混算；Task 3/4 size 测试。
4. 两份各自合法对象拼接、或 root payload 存 new_root_set 而非 rotation body，必须拒绝；Task 5 splice 测试。
5. 最后一个对象失败/取消/panic 时，之前解析成功的对象仍需清理；格式成功不可被接入旧 reader 作为已验签事实；Task 5 lifecycle 与 Task 6 import/caller 检查。

## 文件与接口总图

新建 `internal/nodecontrol/stateformat/`：

- `doc.go`：证据边界；`errors.go`：唯一 `ErrInvalidFormat` 固定无值 sentinel。
- `strict.go` / `strict_test.go`：token/presence/Unicode/depth/原文 JCS 比较。
- `primitives.go` / `primitives_test.go`：U/D/K/S/T/UUID/ID/Build。
- `desired.go` / `desired_test.go`：desired/ProcessSpec/capacity 私有 wire DTO 和校验。
- `root.go` / `root_test.go`：root set/rotation 私有 DTO 和校验。
- `metadata.go` / `metadata_test.go`：metadata/private key DTO、window coverage。
- `envelope.go` / `envelope_test.go`：四类 outer、严格绑定、desired 封装构造。
- `artifact.go` / `artifact_test.go`：拥有数据的结果、摘要/transcript/清理。
- `boundary_test.go`：纯包 import 边界的只读源码测试。
- `vectors_test.go`、`testdata/{desired,root-initial,root-rotation,metadata}.{payload,envelope}.json`、`testdata/vectors.json`：独立字面向量。

新建 serving `desired_format.go`、`desired_format_test.go`、`desired_format_fixture_test.go`。
更新 `docs/roadmap/current-status.md`、`2026-10-08-task10-read-boundary.md`，新建
`docs/roadmap/2026-10-09-desired-canonical-execution.md`；更新本计划完成勾选。

内部 token schema 固定为 `wireSchema`（私有），静态描述 object exact fields、
primitive kind、nullable、array element/count。API：
`strictCanonical(body []byte, limit int, schema *wireSchema) error`。
它先 bounded token walk 并验证 UTF-8/Unicode escape，再调用已锁定 JCS，
与 body 比较。typed decode 只发生在成功之后；禁用自定义解码器绕过 token schema。
schema 只由包内静态声明，不接受用户传来的动态 schema。不得修改共享 strictjson。

Task 2/3 私有 payload parser 均签名 `(body []byte) (artifactBindings, error)`：
`parseDesiredPayload`、`parseRootPayload`、`parseRotationPayload`、`parseMetadataPayload`。
wire DTO 私有且仅用于编解码；本计划不提前创建缺失的 B02 逻辑 ProcessSpec IR。

固定跨包 API（Task 4 提供，Task 5 消费）：

```go
type Kind uint8 // Invalid=0, Desired=1, InitialRoot=2, RootRotation=3, Metadata=4
type Bindings struct {
    Kind Kind
    Epoch, Sequence, RootVersion, PreviousRootVersion, MetadataVersion uint64
    NodeID uuid.UUID
    Generation, InventoryVersion, ResourceEnvelopeVersion uint64
    ResourceEnvelopeDigest, SigningKeyID contracts.Digest
    Signature [64]byte
    IssuedAt, ValidUntil time.Time
}
type ParsedArtifact struct { /* private owned fields only */ }
func ParseEnvelope(kind Kind, envelope []byte) (*ParsedArtifact, error)
func BuildDesiredEnvelope(payload []byte, digest, keyID contracts.Digest, signature [64]byte) (*ParsedArtifact, error)
func (a *ParsedArtifact) Bindings() Bindings
func (a *ParsedArtifact) Payload() []byte
func (a *ParsedArtifact) Envelope() []byte
func (a *ParsedArtifact) PayloadDigest() contracts.Digest
func (a *ParsedArtifact) EnvelopeDigest() contracts.Digest
func (a *ParsedArtifact) Transcript() ([]byte, error)
func (a *ParsedArtifact) Clear()
```

`artifactBindings` 是包内 Bindings 的 type alias，不是第二套字段；无关字段为零。
Payload/Envelope/Transcript 返回副本；Bindings 固定字段按值返回。
nil/zero/cleared artifact 的访问器返回零值，Transcript 返回 ErrInvalidFormat；
InitialRoot 没有 transcript，同样错误。Clear 可重复且 nil-safe，清 owned bytes
和固定字段；不可通过复制结构体或共享指针建立第二所有者，不支持并发 Clear。
Parse/Build 错误返回 nil，永不返回部分对象。无可导出字段或 trusted 构造器。

### Task 1: 严格格式基础与执行基线

**Files:** doc/errors/strict/primitives 与对应测试、ledger。
**Interfaces:** 提供 strictCanonical、wireSchema、ErrInvalidFormat；primitive helpers
`parseU(string)(uint64,error)`、`parseDigest(string)(contracts.Digest,error)`、
`parseKey(string)([32]byte,error)`、`parseSignature(string)([64]byte,error)`、
`parseTime(string)(time.Time,error)`、`parseUUID(string)(uuid.UUID,error)`、
`validID(string,int)bool`、`validBuild(string)bool`。

- [x] **Step 1 — 前置检查。** 读 AGENTS/分支/HEAD/dirty 与部署记录；UNKNOWN 走 Global Constraints 隔离分支。保存 SQL、全部 store、四锁、API、strictjson/trust、现有 serving/authority、批准集及其所列文件的 filename/hash 清单。记录 Task1→2/3→4→5 API 一致性。运行既有三组限定基线命令（见末尾，首组先排除尚不存在的 stateformat）；exit0 才实施，缺缓存不下载。
- [x] **Step 2 — primitives RED。** `TestPrimitiveUExact` 对 1、2^53−1/2^53/2^53+1、MaxInt64 的 string 逐值比较；零、overflow、前导零、符号/指数/小数失败。`TestPrimitiveEncodings` 覆盖 D/K/S 重编码、尾 pad bits、全零，T 日期/offset/秒/零时间，UUID variant且任意 version nibble，ID长度字符、Build三分量边界。
- [x] **Step 3 — strict RED。** `TestStrictPresence` 使用静态小测试 schema：显式 false 和合法 null 通过、缺字段失败；嵌套重复（含 `a` 与 `\u0061`）、unknown/case alias、number-as-string、尾随值、null-array 失败。`TestStrictUnicodeCanonical` 拒绝非法 UTF-8/孤立 surrogate/非 canonical whitespace、键序、转义；合法非 ASCII 字符不得被 normalize。`TestStrictLimits` 私有 schema 中构造深度16/17、limit/limit+1 边界，不让业务 DTO 深度不足掩盖该测试。
- [x] **Step 4 — 运行 RED。** 按离线环境 `test ./internal/nodecontrol/stateformat -run '^(TestPrimitive|TestStrict)' -count=1 -timeout=2m`。先记录 missing symbols，再加 fail-closed declarations 验证合法输入出现行为失败；仅编译/环境错误不算行为 RED。
- [x] **Step 5 — 实现 GREEN。** 新包严格 token walk 用 json.Decoder.UseNumber，仅作结构检查；U 自己精确 ParseUint/range，JSON number 由 schema 拒绝。Unicode escape 在 typed decode 前检查，presence 逐层追踪，不靠零值。重复键按解码后的名字识别。按附录编码规则实现 helpers，错误无输入值。
- [x] **Step 6 — 验证并记录。** 同 selector exit0；临时移除 required 检查使 TestStrictPresence RED 后立即恢复，重跑 stateformat 全包 exit0，记录 ledger，不提交。

### Task 2: desired 与 root payload

**Files:** desired/root 与测试，artifact.go 中 Bindings/Kind 声明，四类中的前三个 payload 字面文件。
**Interfaces:** 消费 Task1；提供 parseDesiredPayload、parseRootPayload、parseRotationPayload。

- [x] **Step 1 — 固定正向样本。** 手写 JCS 排序字面文件，不以产品 encoder 生成期望；payload 不加终端换行。使用不同 node/key/public-key材料及非零假 signature，注明非 crypto 证据。保留 fixture provenance；public-key digest 通过独立 .NET SHA256 核对。计划禁止使用 fixture 伪装生产证据。
- [x] **Step 2 — desired RED。** `TestDesiredPayload` 覆盖全部附录 §4.1 字段、0/8/9 processes、1/64/65 metrics、各 ID/enum/build、reason missing、required=false/missing/null、排序/重复。`TestDesiredCapacityBounds` 对十项 min−1/min/max/max+1 表测。`TestDesiredWindow` 测相等、逆序、24h/24h+1s、年份上界。每个合法 fixture 的每个 required 层做删除/unknown/null/type mutation。
- [x] **Step 3 — root RED。** `TestRootPayload` 测1..5/6 keys、threshold0/1/count/count+1、key ID hash错/重复、初始版本必须1key/threshold1。`TestRotationPayload` 测 previous+1 overflow、epoch/sequence 两层错配、7d/7d+1s、不接受0→1或只传 new_root_set。
- [x] **Step 4 — 运行并实现。** selector `^(TestDesired|TestRoot|TestRotation)` 行为 RED 后实现 private DTO/schema 和 parser。全过程按附录字段，不新增通用 payload map API；时间加法后需检查结果可用和顺序，不读 time.Now。
- [x] **Step 5 — GREEN。** selector 和 stateformat 全包 exit0。比较 fixtures 输入未变；Bindings 含全部后续绑定字段且无关字段零。ledger 记录，本地保留。

### Task 3: metadata 局部语义

**Files:** metadata.go/test、metadata.payload.json。
**Interfaces:** 消费 Task1/2 基础类型，提供 parseMetadataPayload。

- [x] **Step 1 — 写 RED。** `TestMetadataPayload` 覆盖 §4.4 exact字段、role、48h/7d边界与next_refresh，1/64/65 keys、排序唯一、公钥摘要。`TestMetadataRevocationPresence` 区分全部 required null/absent，scheduled/security_incident 两分支、revoked_at≤issued_at、revoked membership；不查外部 incident。
- [x] **Step 2 — coverage RED。** `TestMetadataActiveCoverage` 测 active在起点覆盖、首尾缺口、中间1秒洞、相邻半开区间、包含/重叠、retiring不能补active洞、active/retiring必须与window相交。排序key ID不能被误当window时间排序。
- [x] **Step 3 — 数量与字节双边界。** `TestMetadataLedgerBudget` 有效数量但超过65536 bytes必须拒绝、输入和ledger不被截断。私有 revoked-list validator 单测0/4096/4097，说明4096项不能全部塞入真实64KiB payload，不构造假的“全上限正向”整包测试。
- [x] **Step 4 — RED→GREEN。** selector `^TestMetadata`；先观察合法/边界行为 RED，再实现窗口按 not_before排序的临时副本并作无洞 union。不得改变原key排序；不实施跨版本保留/授权。
- [x] **Step 5 — 回归。** stateformat 全包 exit0；一个撤销null检查或coverage洞检查临时失效应被测出，立即恢复后重跑。记录 ledger。

### Task 4: 四类完整封装、摘要与所有权

**Files:** envelope/artifact/vector 文件与测试、四个 envelope fixtures、vectors.json。
**Interfaces:** 产出总图中完整跨包 API；消费四个 private payload parsers。

- [x] **Step 1 — 冻结向量。** vectors.json 记录每个 kind 对应 payload/envelope 文件路径、SHA256 lowercase hex、transcript hex（初始根为 null）。由手工核对字段的字面文件及 .NET SHA256/UTF8/base64 独立计算值，用 apply_patch 保存；不调用产品实现来生成期待输出。
- [x] **Step 2 — outer RED。** `TestEnvelopeExactFields` 覆盖四类 schema 的全部 required/extra/null/type；`TestEnvelopeBindings` 逐个改 outer authority/root/previous、payload digest、嵌套 body kind；`TestSharesFormat` 0/64/65、排序/重复/坏signature/结构正确未知ID；空shares可格式通过但不作阈值承诺。
- [x] **Step 3 — 大小/摘要 RED。** `TestEnvelopeSizeLayers` 私有payload decoder检查87382及真实65536/65537、错误pad bits；strict outer limit262144/262145；真实合法metadata envelope超过64KiB仍可通过（使用大量但未超过payload64KiB的revoked IDs）。不要要求exact262144合法业务对象存在。`TestArtifactVectors` 精确对比P/E摘要和domain+NUL+P；InitialRoot.Transcript错误。
- [x] **Step 4 — 构造/所有权 RED。** `TestBuildDesiredEnvelope` 与字面E全等，错误digest/key/signature或不规范P失败，非零假签名正常格式可通过。`TestArtifactOwnership` 修改输入和每份 accessor 副本立即断言；内部保留owned aliases检查Clear清零，nil/zero/重复Clear行为固定。任何Parse/Build失败必须nil，无部分材料。
- [x] **Step 5 — 实现并验证。** `ParseEnvelope` kind参数固定schema，不按输入尝试其他kind；长字符串先字符长度再decode并验原长度；逐层canonical比较。Build先parse P和核摘要，再唯一编码E，调用真实ParseEnvelope回读。所有中途分配由单一生命周期清理，成功才转移。selector `^(TestEnvelope|TestShares|TestArtifact|TestBuildDesired)` GREEN后全包GREEN。
- [x] **Step 6 — 敏感性。** 临时去掉payload digest比较或所有权clone，具名测试RED后恢复；全包exit0并ledger。没有验签器或信任来源调用。

### Task 5: 私有投影绑定与失败清理组合

**Files:** serving/desired_format.go、desired_format_test.go、desired_format_fixture_test.go。
**Interfaces:** `func bindDesiredFormat(ctx context.Context, p desiredProjection) (desiredFormat, error)`；
`desiredFormat` 私有，独占 desired/root/metadata 的 `*stateformat.ParsedArtifact` 和
rotation指针slice；`func discardDesiredFormat(f *desiredFormat)` 可重复。无新公开reader。

- [x] **Step 1 — 新fixture。** 独立构造0/1/32 rotations 的canonical payload/envelope并配套结构投影，不改既有opaque desiredRowsFixture。期望从附录字面样本/显式字段建立，不用被测binder作oracle。根链尾和active root字节一致、对象拥有独立内存；固定epoch19、角色不同UUID/sequence。
- [x] **Step 2 — RED绑定。** `TestDesiredFormatBinding` 对所有重复字段逐项改payload/outer/projection；重算被改对象的合法摘要，确保失败原因是绑定而非坏hash。`TestDesiredFormatSplice` 交换两份合法envelope、误传new_root_set作root P、初始root作rotation、链中间错配；0/1/32成功，33链/零投影拒绝。
- [x] **Step 3 — 最小实现。** 首先取消→Canceled，nil ctx→Unavailable；再以既有newDesiredProjection重检结构并清理其临时副本（不改输入）。逐一ParseEnvelope核对P/E及摘要和kind/Bindings，desired使用BuildDesiredEnvelope；所有root历史均查，metadata版本/根绑定 exact。结构/格式错→Conflict；每对象前后及交付前检查ctx。失败返回零desiredFormat。不得复制EffectiveDeadline进入P或修改readDesiredProjection。
- [x] **Step 4 — 生命周期 RED/组合。** `TestDesiredFormatLifetime` 保留中途ParsedArtifact的owned aliases，通过私有无hook的owner方法验证最后对象失败、取消及panic时清零；明确直接helper测试不声称普通codec自然会panic。`TestDesiredFormatGuardComposition` 测试中用真实guard+adapter+binder，后检reject/cancel/sourcepanic清零、成功保留；不新增产品reader、callback观察hook。原projection/fixture不变。
- [x] **Step 5 — GREEN/敏感性。** `test ./internal/nodecontrol/serving -run '^TestDesiredFormat' -count=1 -timeout=2m` exit0；临时在测试组合传no-op discard，reject/cancel/panic断言RED后恢复。新stateformat及serving全包exit0。旧测试不得为接受新格式而放宽或重写。

### Task 6: 限定验证、独立审查和交付

**Files:** 两份roadmap、执行ledger、本计划；仅文档状态。
**Interfaces:** 消费 Tasks1–5 证据，无新增 API。

- [x] **Step 1 — 新鲜限定回归。** 执行下面三组命令，全部exit0；store明确no tests。gofmt -l新Go文件为空；git diff --check和未跟踪文件空白检查通过。依赖/编译错误不是扩大执行或下载的许可。
- [x] **Step 2 — 边界测试。** `TestFormatDependencyBoundary` 放在stateformat/boundary_test.go，go/parser读取非测试imports，拒绝state/serving/authority/store/trust/网络/数据库/crypto验签导入，且只允许标准库、uuid、jcs、contracts；测试其私有checker对合成禁止import片段先RED后GREEN，不向产品文件临时导入网络库。最终只读go list确认新依赖无意外扩大。全仓rg确认bindDesiredFormat仅测试调用，reader/adapter不变。禁止以名称Contains误判包前缀，比较完整import path。重跑stateformat全包，exit0后才最终审查。
- [x] **Step 3 — 新上下文最终审查。** 沿用Native方式，一名最强可用审查者只读检查新增实现、spec、plan、ledger和具名依赖；不得把历史分支全部视作通过。Important/Critical复现RED后一次修复与限定全回归；Minor记录延期。所有excluded行为由主执行者逐项裁定并写入ledger。
- [x] **Step 4 — 不变性与完成记录。** 与Task1 hash清单比对所有保护文件；列新增文件、测试结果、修复/未关闭项、无独立重复执行之证据边界。状态注明本地未提交未推送、G1–G5仅局部开发契约已落实，不代表生产登记/部署兼容。若真实规范冲突出现，停止受影响实现，不暗改schema。
- [x] **Step 5 — 交付。** 勾选实际完成项，保留ledger/worktree。Task10、签名/信任/可信时间、B02/B03/HTTP同步、部署记录、v7/DB/PITR/security/B01/C12/main门禁继续开放；两项旧测试minor不自动关闭。

## 固定离线验证命令（已按许可在执行阶段运行）

在上述worktree，所有selector沿用下列环境。测试需要缓存写权限时仅申请
限定提权；不可因权限/缓存失败删缓存、下载或使用被安全系统阻止的路径。

```powershell
$env:GOENV='off'; $env:GOWORK='off'; $env:GOTOOLCHAIN='local'
$env:GOPROXY='off'; $env:GOSUMDB='off'; $env:GOAUTH='off'
$env:GOVCS='all:off'; $env:GOFLAGS='-mod=readonly'
$env:GOPATH='C:/Users/Lenovo/go'; $env:GOMODCACHE='C:/Users/Lenovo/go/pkg/mod'
$taskGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
& $taskGo test ./internal/nodecontrol/stateformat ./internal/nodecontrol/serving ./internal/nodecontrol/contracts ./internal/readiness -count=1 -timeout=2m
& $taskGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1 -timeout=2m
& $taskGo test ./internal/store -run '^$' -count=1 -timeout=2m
```

## 自查与执行交接

附录§3→Task1，§4 desired/root→Task2、metadata→Task3，§5–7→Task4，
投影绑定/清理→Task5，归属/许可/兼容性→Task1/6。五项Review Focus均有测试。
本计划不承诺实际发布格式已兼容，没有修改批准集或生成生产验收证据。

沿用此前明确选择的 Native：当前会话实现，最后一次独立审查，不逐任务
派实现代理。用户须先审阅并确认本计划；确认也明确覆盖新增stateformat
包的离线纯单元测试范围，不授权DB/网络/生产集成或提交推送。
