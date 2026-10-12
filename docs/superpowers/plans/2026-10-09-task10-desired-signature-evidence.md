# Desired Signature Evidence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** 增加隔离、可清理的签名关系证据，不产生生产授权能力。

**Architecture:** stateformat 是唯一 wire 解析者，保留并复制规范化材料；新 stateverify 使用普通 Ed25519 验证 exact transcript、连续双阈值轮换及 metadata/desired 关系。结果只相对于调用者独立提供的初始根成立，不证明该根可信、当前有效或未回滚。

**Tech Stack:** Windows PowerShell；固定 Go 1.26.5；标准库 crypto/ed25519；现有缓存 uuid/JCS/contracts；无新依赖。

**Spec:** `docs/superpowers/specs/2026-10-09-task10-desired-signature-evidence-design.md`（用户已确认）；格式依据 `docs/superpowers/specs/2026-10-09-node-state-format-contract-addendum.md`。

状态：已按批准范围完成 Native 局部实现及一次独立审查；审查的 1 项非阻塞 Minor 已在后续 TDD 修复中关闭，限定回归通过，未做第二次独立审查；未提交推送。本文件不表示整体验收完成。

## Global Constraints

- 工作目录固定 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`，分支 `codex/c12-b01-task9-coordinator`；保留所有已有改动。
- 计划已获批；按以下边界实施。执行中不暂存、提交、推送、合并或清理分支。
- 本计划明确请求新增 stateverify 离线纯单元测试许可；用户批准本计划后才运行。其余测试仅限下方允许清单。
- 禁止真实 DB/Docker、C11/C12/bootstrap/PITR/race/全仓测试、依赖下载、安全设置修改及绕过既有安全拒绝。
- 0..32 个 rotation，四个单体加轮换最多 36 个 envelope；所有 envelope 非空且 ≤262144 bytes，payload ≤65536 bytes；先完成全批数量/外层长度检查再解析。
- ExpectedEpoch 为 1..9223372036854775807；ExpectedNodeID 为 canonical、非 nil、RFC4122 variant UUID，不额外限制版本。
- 初始根 version=1、单 key、threshold=1；根 key 1..5，metadata key 1..64，shares 0..64；原格式规则、排序、唯一性、wire bytes 和旧独立向量不变。
- 不新增 JSON schema、算法协商、可替换 verifier、trusted bool、time.Now、caller now、生产调用者、网络或持久化。
- serving/authority/store/SQL/API/四锁/基础规范/批准集不改；部署兼容性 UNKNOWN，历史两项 Minor、安全及整体验收门禁继续开放。
- 手工文件修改仅用 apply_patch。取消不承诺单次 Ed25519 操作可抢占；所有者不可复制或与 Clear 并发使用。

## Review Focus

1. 已达到 threshold 后出现已知坏签名仍必须失败：Task 3 `TestThresholdKnownBadAfterQuorum`。
2. 同 key 的兼容 metadata 替换可以成立，但证据必须记录不同身份：Task 4 `TestVerifyCompatibleMetadata`。
3. 另一整套自签数据或过期数据仅产生条件性证据：Task 4 `TestVerifyConditionalOnly`。
4. 空合法 shares 与错误 Kind 不可混淆，副本变异不可影响所有者：Task 1 `TestMaterialKindAndOwnership`。
5. 最后一步失败、取消或 panic 不得泄露局部结果和自有缓冲：Task 5 `TestOwnerExitCleanup`。

---

## 文件及接口地图

路径均相对于上述 worktree。新增 stateformat/material.go、material_test.go；有限修改 artifact.go、envelope.go、root.go、metadata.go、strict.go 及相应测试。新增 stateverify/doc.go、errors.go、evidence.go、owner.go、threshold.go、verify.go、fixtures_test.go、threshold_test.go、verify_test.go、owner_test.go、boundary_test.go、testdata/*。不改旧 stateformat/testdata。

执行记录新增 `docs/roadmap/2026-10-09-desired-signature-execution.md`；结束时仅据实际证据更新该记录、`docs/roadmap/current-status.md` 和本计划勾选状态。

### 冻结的 stateformat 接口

以下为公开材料值；仅 ParsedArtifact 能提供经过本次解析的来源。固定数组按值复制，所有 slices 深复制；各容器增加 nil-safe、幂等 `Clear()`。

```go
type PublicKeyMaterial struct { KeyID contracts.Digest; PublicKey [32]byte }
type RootMaterial struct { Version, Threshold uint64; Keys []PublicKeyMaterial }
type MetadataKeyMaterial struct {
    PublicKeyMaterial
    Status string // active / retiring / revoked；非算法协商
    NotBefore, NotAfter time.Time
    HasRevocation bool
    RevokedAt time.Time
    RevocationReason string
    HasIncident bool
    IncidentID uuid.UUID
}
type MetadataMaterial struct {
    Version, RootVersion uint64
    Keys []MetadataKeyMaterial
    RevokedIDs []contracts.Digest
}
type SignatureShare struct { KeyID contracts.Digest; Signature [64]byte }
type SignatureMaterial struct { Current, New, Metadata []SignatureShare }
func (*ParsedArtifact) RootMaterial() (RootMaterial, error)
func (*ParsedArtifact) MetadataMaterial() (MetadataMaterial, error)
func (*ParsedArtifact) SignatureMaterial() (SignatureMaterial, error)
func (*RootMaterial) Clear()
func (*MetadataMaterial) Clear()
func (*SignatureMaterial) Clear()
```

Root accessor 只接受 InitialRoot/RootRotation；Metadata 只接受 Metadata；Signature 只接受 RootRotation/Metadata。nil/zero/cleared/wrong Kind 均返回零值及 ErrInvalidFormat。合法空 shares 返回成功，非对应组保持 nil。HasRevocation/HasIncident 保留原 nullable 语义；不承诺擦除 string 后端。

### 冻结的 stateverify 接口

```go
type Batch struct {
    ExpectedNodeID string
    ExpectedEpoch uint64
    AnchorEnvelope []byte
    Rotations [][]byte
    ActiveRootEnvelope, MetadataEnvelope, DesiredEnvelope []byte
}
type ArtifactIdentity struct { PayloadDigest, EnvelopeDigest contracts.Digest }
type EvidenceCoordinates struct {
    ExpectedNodeID uuid.UUID
    ExpectedEpoch, FinalRootVersion, MetadataVersion uint64
    DesiredGeneration, DesiredSequence uint64
    DesiredKeyID contracts.Digest
}
type SignatureEvidence struct { /* 仅私有自有字段 */ }
func Verify(ctx context.Context, batch Batch) (*SignatureEvidence, error)
func (*SignatureEvidence) Coordinates() EvidenceCoordinates
func (*SignatureEvidence) Anchor() ArtifactIdentity
func (*SignatureEvidence) Rotations() []ArtifactIdentity
func (*SignatureEvidence) Metadata() ArtifactIdentity
func (*SignatureEvidence) Desired() ArtifactIdentity
func (*SignatureEvidence) Clear()
```

四个 errors.New sentinel：ErrInvalidInput=`stateverify: invalid input`；ErrBindingMismatch=`stateverify: binding mismatch`；ErrSignatureMismatch=`stateverify: signature mismatch`；ErrCanceled=`stateverify: canceled`。不包装动态库错误；nil ctx 为 InvalidInput，非 nil 已取消优先。格式/参数错误归 InvalidInput，合法格式的对象关系或静态窗口不符归 BindingMismatch，密码学或份额不足归 SignatureMismatch。

## 验证命令与记录

执行时 PowerShell 设置 GOENV=off、GOWORK=off、GOTOOLCHAIN=local、GOPROXY=off、GOSUMDB=off、GOAUTH=off、GOVCS=all:off、GOFLAGS=-mod=readonly、GOPATH=`C:/Users/Lenovo/go`、GOMODCACHE=`C:/Users/Lenovo/go/pkg/mod`；工具路径：

```powershell
$signatureGo = 'C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
& $signatureGo test ./internal/nodecontrol/stateformat -count=1 -timeout=60s
& $signatureGo test ./internal/nodecontrol/stateverify -count=1 -timeout=60s
& $signatureGo test ./internal/nodecontrol/serving ./internal/nodecontrol/contracts ./internal/readiness -count=1 -timeout=60s
& $signatureGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1 -timeout=60s
& $signatureGo test ./internal/store -run '^$' -count=1 -timeout=60s
```

Task 中 selector 均替换相应包命令的 `-run`，保留 -count=1/-timeout=60s。成功要求退出码 0 且目标测试实际运行；RED 要记录预期断言失败，只有缺符号/编译错误不算行为 RED。可先补最小失败闭合骨架，再观察行为失败。超时/取消/环境失败不能标成验收通过，不扩大测试范围重试。

### Task 1: 自有格式材料和两个新 Minor

**Files:** 上述 stateformat 修改集及 material.go/material_test.go、strict_test.go、artifact_test.go。

**Interfaces:** 消费现有 ParseEnvelope/Bindings；产出冻结的三个材料访问器和 Clear。私有解析 helper 冻结为 `decodeRootPayload([]byte) (artifactBindings, RootMaterial, error)`、`decodeRotationPayload([]byte) (artifactBindings, RootMaterial, error)`、`decodeMetadataPayload([]byte) (artifactBindings, MetadataMaterial, error)`；保留旧 parseXPayload 签名为调用 helper 并清理材料的薄包装。ParseEnvelope 直接消费 helper；每次 payload 仅一套 DTO 解析/语义检查。

- [x] 建立本阶段基线：记录 HEAD、已有变更清单；哈希保护所有范围外已有文件及旧 stateformat/testdata，记录可修改文件的初始 hash，不把上一阶段 156 文件列表误当新基线。
- [x] 写 `TestMaterialKindAndOwnership`：真实格式样本逐字段比较版本/公钥/nullable 撤销字段；逐次变异副本立即断言 artifact 不变；测试所有 Kind/nil/zero/clear 组合和合法空 shares。写 `TestMaterialFailureCleanup` 保留私有 owner 的 slices 别名，失败后断言清零。
- [x] 写 `TestCanonicalOwnedBufferClear` 覆盖相等/不等缓冲清零及 body 不变；写 `TestCanonicalExactSizeLimit` 使用 canonical JSON 字符串长度恰为 limit 和 limit+1：前者成功后者 ErrInvalidFormat，不用尾随空白制造负例。
- [x] 运行 stateformat selector `Test(Material|Canonical)`，记录材料/清理行为 RED；大小测试若原实现已通过，临时仅将大小 guard 放宽一字节，确认新断言失败后立即恢复，记录为测试有效性 mutation，不把原正确 guard 声称为修复。
- [x] 实现私有规范化材料存储和单次解析路径；构造立即登记清理，成功才转移所有权。Artifact.Clear 先清理材料再归零；所有 accessor 只 clone、不解码。strict.go 增加私有 `equalCanonicalAndClear(body, canonical []byte) bool`，defer clear(canonical)，JCS 返回后无论 err 都调用再判错，确保其直接持有的临时缓冲清理。
- [x] 同 selector GREEN，再运行 stateformat 全包；固定旧格式向量 hash 不变。记录两项新 Minor 的证据，历史两项仍开放。

### Task 2: 固定证据类型与独立真实签名向量

**Files:** stateverify/doc.go、input.go、errors.go、evidence.go、fixtures_test.go、evidence_test.go、testdata/anchor.envelope.json、rotation.envelope.json、metadata.envelope.json、desired.envelope.json、rotation.transcript、metadata.transcript、desired.transcript、vectors.json。

**Interfaces:** input.go 先定义冻结的 Batch，产出上述 evidence API、sentinel；测试 helper `signedBatch(t *testing.T, rotations int) Batch` 供后续任务使用，仅在测试代码构造扩展样本。

- [x] 写 `TestEvidenceOwnership`：私有字段装入已知值，accessor 返回精确值；变异 Rotations 副本不能影响 owner；nil/zero/两次 Clear 全零，持有内部 slice 别名见清零。写 `TestFrozenSignatureVectors` 逐字比较固定 transcript 和 P/E SHA256，直接标准库 Verify 验证三类真实签名。
- [x] 独立冻结向量：测试专用固定 seed 和手写 canonical payload；离线一次性辅助生成不得导入 stateformat/stateverify，不使用产品 encoder/Transcript。记录 seed 仅为测试、完整域+NUL+P、pubkey/hash/签名期望与生成来源到 vectors.json；辅助源保存在 testdata/generate_vectors.go 并标记 build ignore，不进入生产。先以错误摘要/签名副本确认向量断言能失败，再恢复期望；不把 round-trip 当独立 oracle。
- [x] 运行 stateverify selector `Test(Evidence|Frozen)`，记录缺少复制/清理实现时的行为 RED；实现 evidence 私有字段和值访问器，Clear 清理 rotations 底层及固定字段；doc 明确 conditional-only，固定 sentinel。
- [x] 同 selector GREEN；固定文件后后续测试不得运行时重写 oracle。signedBatch 的衍生案例可在测试内用普通 Ed25519 签名，但固定向量的期望不得依赖它。

### Task 3: 完整检查的阈值验证

**Files:** stateverify/threshold.go、threshold_test.go。

**Interfaces:** 消费 stateformat.RootMaterial/SignatureShare；产出私有 `verifyThreshold(ctx context.Context, root stateformat.RootMaterial, shares []stateformat.SignatureShare, transcript []byte) error`，只由本包已解析路径调用，不对外接受材料。

参数只在调用期间借读，不保留、不变异、不 Clear；材料仍由调用方 owner 唯一清理。关键测试断言形如：

```go
if err := verifyThreshold(ctx, root, knownBadAfterQuorum, transcript); err != ErrSignatureMismatch {
    t.Fatalf("known bad share after quorum: %v", err)
}
```

- [x] 写 `TestThresholdMatrix`：root size 1..5，每个 threshold 1..size，有效票恰够/超过均成功，不足失败；未知结构正确份额不计数；已知错误 key/domain/hash-message/非零假签名失败。固定向量证明 ordinary Ed25519，而不是 ph/context。
- [x] 写 `TestThresholdKnownBadAfterQuorum`：排序后最后一个已知 key 坏签名，前面已经达到 threshold，仍为 ErrSignatureMismatch；`TestThresholdSeparateGroups`：同公钥可在两代各计一次，但两组不可互借。
- [x] 运行 stateverify selector `TestThreshold` 记录 RED；实现 keyID 查找，全部已知份额逐个 Verify，每次调用前后查 ctx，完成整组再比较 threshold。未知份额不验密码学、不计票，格式验证仍是上游强制前提。
- [x] 同 selector GREEN；增加已取消输入返回固定 ErrCanceled，无动态错误、无 goroutine。

### Task 4: 完整批次关系与单一交付点

**Files:** stateverify/verify.go、verify_test.go、owner.go；必要测试材料仅改 fixtures_test.go。

**Interfaces:** 实现 Verify；消费 Task 1–3 的冻结接口。引入私有 `verificationOwner`，其 `Clear()` 清理已登记 artifacts、材料、transcripts 和未转移 evidence；对象取得后立即登记并 defer Clear，成功仅在最后 detach evidence。

- [x] 写 `TestVerifyInputAndChain`：0/1/32 正向，33/空/262145 bytes/坏 UUID/越界 epoch/缺 anchor 负向；缺段/乱序/分叉/跨 epoch/不递增 sequence/替换不匹配 anchor/active E 非逐字链尾均失败且 result=nil。单独无轮换 active 必须等于 anchor；current/new 各不足都拒绝；非法 shares、重复 ID、错 Kind/C11、本体改变未重签全部拒绝。
- [x] 写 `TestVerifyBindingsAndWindows`：metadata root/epoch、desired node/epoch/key；active/retiring 成功，revoked/缺 key 失败；issued==not_before 通过、issued==not_after 失败；until==metadata/key 上限可通过、晚一秒失败；合法较新 metadata 验旧 retiring desired 通过。新完整快照缺 key 不从旧样本补入；不同对象流 sequence 不人为比较。
- [x] 写 `TestVerifyCompatibleMetadata`：同 key 的两份签名正确兼容 metadata 都能通过，但 Metadata identity 精确不同；不声称 desired 签入不存在的 metadata 版本字段。写 `TestVerifyConditionalOnly`：独立另一整套自签批次和固定过去日期样本只输出条件性证据，不访问系统时钟或改变任何 reader。
- [x] 运行 stateverify selector `TestVerify` 记录行为 RED；实现先全批边界后固定 Kind 解析，anchor version1/1key/threshold1，逐轮 checked +1（不得超过 MaxInt64）、epoch/sequence/双阈值验证后推进；active 原始 E 相等；metadata 完整快照阈值；desired 精确 key/静态窗口/签名。
- [x] 成功最后填写所有 P/E identity、按序 rotation identity 和 coordinates；每对象前后、每次验签前后、交付前检查取消。任何错误 result=nil，无缩短链重试或 fallback。运行同 selector GREEN 及 stateverify 全包。

### Task 5: 所有权退出路径与依赖边界

**Files:** stateverify/owner_test.go、boundary_test.go；必要修改 owner.go/verify.go；有限追加 stateformat 材料失败清理测试。

**Interfaces:** 沿用 verificationOwner.Clear，不新增观察 hook、验签回调或生产开关；nil/zero evidence accessor 全零。

- [x] 写 `TestOwnerExitCleanup`：测试内构造真实 owner 并保留每个自有 backing slice 别名，经 defer Clear 分别走 error/cancel/panic（测试外层 recover），断言别名清零、未交付 evidence 清理；逐项变异立即断言，避免双 XOR 抵消。直接测试 owner 不冒称注入了产品内部 panic。
- [x] 写 `TestVerifyLateFailureAndCancel`：最后 desired 签名坏 result=nil；已取消优先于坏输入；nil ctx InvalidInput；测试专用 context 在固定 Err 调用次数取消以覆盖对象边界、验签边界、最终交付，不用 sleep 或替换 verifier。结合 owner 测试和 defer 源码边界检查覆盖内部退出清理。
- [x] 写 `TestStateverifyBoundary`：AST 只允许生产直接 imports bytes/context/crypto/ed25519/errors/time、uuid、contracts、stateformat；time 只允许静态时间类型/比较，禁止 Now；禁止 JSON/JCS、网络/DB/store/authority/serving/trust/签名调用和 go 语句。扫描本仓非测试 Go imports，拒绝 stateverify 生产调用者；stateformat 不得反向导入 stateverify。检查 Verify 的 owner defer 在取得材料之前登记。
- [x] 运行 stateverify selector `Test(Owner|VerifyLate|StateverifyBoundary)` 记录 RED；修复必要清理/取消实现，不为测试增加 hook。无法通过无侵入测试覆盖的库内 panic/不可擦除副本明确写为保证范围外。
- [x] 同 selector GREEN，运行完整允许清单；比较保护文件与旧向量 hash、检查 diff 仅在批准范围。任何不符停止交付，不改变安全门禁。

## 最终审查与交付（Native）

- [x] 主代理完成五个任务后，按 requesting-code-review 技能要求交由一个全新、当时可用最强模型 reviewer 审查本阶段增量及证据；不给它改代码权限，不自动扩大测试范围，不把整个历史分支差异当本次增量。
- [x] 有 Critical/Important 则按 receiving-code-review/TDD 修复并重跑受影响的允许测试；未解决则标明阻塞，不能宣称完成。Minor 分别记录处理或明确保留，不隐式关闭历史问题。
- [x] 执行记录写明真实命令/退出码、RED/GREEN、固定向量来源、保护 hash 对照、审查结果与剩余门禁；更新 current-status 只描述本地隔离能力。不宣称 Task10/B01/C12/main/security 完成，不提交推送。

## 计划自查

设计 §1–3 对应约束/API/Task4；§4 对应 Task1；§5 对应 Task3；§6–7 对应 Task4；§8 对应 Task2/4/5；§9–10 对应所有测试及交付门禁。五项 Review Focus 都有具名测试。材料保留单次 wire 解析、签名阈值无 early success、兼容 metadata 不误判绑定、过期链不当授权、错误无部分交付均已覆盖。执行方式延续 Native；用户已确认计划及新增纯单元测试范围，最终独立审查已完成：无 Critical/Important，1 项错误分类 Minor 延后处理；参见执行记录。
