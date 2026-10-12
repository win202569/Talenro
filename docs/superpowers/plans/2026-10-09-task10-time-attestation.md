# Time Attestation Isolated Verification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Native execution was already selected; do not switch to subagent implementation. Steps use checkbox syntax for tracking.

**Goal:** 实现时间证据的离线格式、条件性签名验证及 freshness/checkpoint 判定，不输出生产授权。

**Architecture:** stateformat 独占严格 wire 解析；新增 statetime 消费原始封装和显式当前 metadata/时间/请求/checkpoint 模型。核心无 I/O、无真实时钟、不更新可信状态；重检重新验签。

**Tech Stack:** Go 1.26.5、标准库 Ed25519、已有 JCS/UUID 库、PowerShell；离线且无新增依赖。

**Spec:** `docs/superpowers/specs/2026-10-09-task10-time-attestation-design.md`，已批准，执行者必须完整阅读。

状态：五项实施已完成；修复后限定回归全部通过。独立审查两项轻微测试覆盖建议明确待补，详见执行记录。

## Global Constraints

- 复用现有工作树 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`，保护已有未提交成果，不另建工作树。
- payload 上限 64 KiB，envelope 上限 256 KiB；不扩宽共享限制。
- checkpoint 使用 1..MaxInt64；尚无可信本地 checkpoint 不能用零隐式初始化。
- 格式层检查 `issued_at < valid_until` 且窗口不超过 5 分钟，不查询 now。
- 任何成功返回都不产生 production capability；metadata 根签名/平台真实性不是本模块证明范围。
- 无运行时消费者，无 TLSValidated/NonceConsumed 布尔替代物；不接网络、时钟、持久化、TPM、签名服务。
- 不新增依赖、数据库字段、OpenAPI 或外部网络协议发布；不修改旧批准集及其 hash。
- 不暂存、提交、推送、合并或清理分支。任务末尾记录检查点代替技能模板的自动提交。
- 不自动扩大至真实 DB/Docker、整仓、C11/C12、bootstrap、PITR 或 race；生产兼容性 UNKNOWN。
- Native 主代理实现；全部任务后一次全增量独立审查，记录未判断领域，不把结果称作全分支验收。

## Review Focus

1. nonce 全零仍是合法 32-byte 编码，不能误用拒绝全零的公钥 helper（任务 1）。
2. 极远时间差、年份边界和亚秒 Now 不应溢出或被取整成可接受（任务 3）。
3. 同 metadata P、不同 E 不能偷偷替换当前观测；相同 checkpoint 的不同合法请求不是内容分叉（任务 2、3）。
4. 重检不得只看观测 ID：保持 ID 而改变 nonce/checkpoint/状态仍必须失败（任务 4）。
5. 同一未消费模型重复调用可成功，必须明确这不是防重放；别名导入也不能绕过无时钟约束（任务 3、5）。

## 文件与接口安排

所有下面相对路径以工作树为根。只修改 stateformat 的 `artifact.go`、
`envelope.go`，新增 `time_attestation.go` 和同名测试；必要的新增向量
放 `stateformat/testdata/time-attestation-v1.json`。不修改旧四类格式的接受规则。

新包 `internal/nodecontrol/statetime/`：

- `doc.go`：条件性、来源真实性及生产集成限制。
- `input.go`：固定值观测、ModelInput，零值不可信。
- `errors.go`：固定错误，不插入输入数据。
- `verify.go`：解析、观测绑定、选键及 Ed25519。
- `policy.go`：显式时间和 checkpoint 比较。
- `evidence.go`：独立持有的 EvidenceModel、摘要访问器和 Clear。
- `evaluate.go`：EvaluateModel、RecheckModel。
- `fixtures_test.go`、`verify_test.go`、`policy_test.go`、`evidence_test.go`、
  `boundary_test.go`：分别对应构造、验签、规则、重检/所有权和隔离。

固定公共接口（在各任务标明实现时点）：

```go
type ObservationStatus uint8 // Unknown=0, Available, Unavailable, Rollback
type CurrentMetadataObservation struct {
    Status ObservationStatus
    SourceID, ObservationID, StateID contracts.Digest
    Payload, Envelope contracts.Digest
    NodeID uuid.UUID
    Epoch, Version uint64
}
type TimeObservation struct {
    Status ObservationStatus
    ProviderID, ObservationID contracts.Digest
    Now, Floor time.Time
}
type RequestObservation struct {
    Status ObservationStatus
    RequestID, ObservationID contracts.Digest
    NodeID uuid.UUID
    Epoch uint64
    ExpectedNonce [32]byte
}
type CheckpointObservation struct {
    Status ObservationStatus
    SourceID, ObservationID, StateID contracts.Digest
    NodeID uuid.UUID
    Epoch, Sequence uint64
}
type ModelInput struct {
    AttestationEnvelope, MetadataEnvelope []byte
    Metadata CurrentMetadataObservation
    Time TimeObservation
    Request RequestObservation
    Checkpoint CheckpointObservation
}
func EvaluateModel(context.Context, ModelInput) (*EvidenceModel, error)
func RecheckModel(context.Context, *EvidenceModel, ModelInput) error
func (*EvidenceModel) Summary() EvidenceSummary
func (*EvidenceModel) Clear()
```

EvidenceSummary 为纯值：AttestationPayload/AttestationEnvelope、
MetadataPayload/MetadataEnvelope、KeyID、RequestID（均 Digest），NodeID、
Epoch、Nonce[32]、OldCheckpoint、ProposedCheckpoint、AdvanceRequired、
ValidUntil。完整观测快照及有效标记保留为 EvidenceModel 私有字段，不提供
手工构造成功结果的函数；nil/清除后 Summary 返回零值。

错误 sentinel 固定名：ErrInvalidInput、ErrBindingMismatch、
ErrSignatureMismatch、ErrUnauthorizedKey、ErrAssuranceMissing、
ErrInvalidTime、ErrExpired、ErrCheckpointRollback、ErrStaleEvidence。
取消返回 context 的取消/截止错误，可用 errors.Is 检查；nil ctx 为输入错误。

## 验证命令约定

工作目录固定为上述工作树；使用每个终端单独设置的离线环境：

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
$attestationGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
```

下文命令在此前缀之后执行；GREEN 必须 exit 0 且输出 ok，禁止只凭缓存
结论（使用 -count=1）。RED 应明确来自本任务缺失行为，不把环境失败作 RED。
有疑似挂起先检查执行状态，不同时启动重复测试；保持简短进度反馈。

### Task 1：时间证据严格格式和独立向量

**Files:** stateformat/artifact.go、envelope.go；新增 time_attestation.go、
time_attestation_test.go、testdata/time-attestation-v1.json。
**Interfaces:** 沿用 ParseEnvelope(Kind, []byte)；追加 Kind TimeAttestation、
Bindings.CheckpointSequence uint64、Bindings.RequestNonce [32]byte。
新增私有 parseTimeAttestationPayload([]byte) (artifactBindings,error)。

- [x] 记录只读基线：工作树/HEAD、修改清单、所有既有受保护源码和锁文件 hash；确认目标新文件不存在，已有同名文件则先核对，不覆盖。
- [x] 写 TestTimeAttestationFormat、TestTimeAttestationNonce、TestTimeAttestationVector、TestTimeAttestationOwnership；断言如 `b.Sequence == 0`、`b.CheckpointSequence == 7`、全零 nonce 可解析、`valid_until-issued_at == 5*time.Minute` 可解析而 +1 秒失败。加入 unknown/duplicate/escaped-duplicate、错误 schema/Kind、数值整数、MaxInt64+1、非法 UTF-8、尾随值、padding、非零尾比特和错误摘要反例。
- [x] 运行 `& $attestationGo test ./internal/nodecontrol/stateformat -run '^TestTimeAttestation' -count=1`，记录 RED。
- [x] 按设计 §3 实现；单签名封装仅在 Desired 或 TimeAttestation 分支读 key/signature，不能落入 root 元字段路径。nonce helper 独立于 parseKey；不增加生产 builder。
- [x] 固定向量直接给出原始 P/E、SHA256 和 domain+NUL+P，期望值不得由待测 Transcript 生成；大小限制测试分别用 canonical 超限值覆盖 65536+1、262144+1，并以现有共享边界测试保证上限不变。清理访问器副本、原输入变更都不得修改对象。
- [x] 运行同组及完整 stateformat，均 GREEN；记录受保护旧格式回归，无提交。

### Task 2：绑定模型与独立签名核验

**Files:** 新 statetime/doc.go、input.go、errors.go、verify.go、fixtures_test.go、verify_test.go。
**Interfaces:** 定义上述全部输入类型和错误。私有
`verifyModel(ctx context.Context, in ModelInput) (*verifiedFacts,error)`；
verifiedFacts 拥有 ParsedArtifact 两份、MetadataMaterial 和私有 transcript，
提供 nil-safe Clear；不供包外调用，不返回生产授权。

- [x] 写 TestTimeVerifyBindings、TestTimeVerifySignature、TestTimeVerifyKeyStatus：固定种子 Ed25519 测试签名直接构造域消息；好签名成功，错域/签名/公钥失败；metadata E 观测不匹配已覆盖；真实同 P 不同 E 原始封装替换的直接测试按最终审查 Minor 明确待补；伪造全部观测仍只是模型，不承诺根签名验证。
- [x] 表驱动覆盖每项观测的 Unknown/Unavailable/Rollback/非法状态、零身份、错误节点/epoch/版本/StateID、0/MaxInt64+1 checkpoint、请求 nonce 不同、active/retiring 正例和 revoked/账本/缺失键拒绝。预取消与 nil context 独立断言。
- [x] 运行 `& $attestationGo test ./internal/nodecontrol/statetime -run '^TestTimeVerify' -count=1`，记录 RED。
- [x] 实现 verifyModel，先检查长度及显式值再解析；绑定精确 P/E 后从当前 metadata 选择密钥，用 crypto/ed25519.Verify 验证原始 Transcript，前后检查 ctx。失败统一清理临时所有者，错误不含 payload/nonce。不查询真实来源，不调用 stateverify/statepolicy，不证明 metadata 根签名。
- [x] 运行同组 GREEN；错误路径断言 `facts == nil`，成功路径显式 Clear，记录检查点。

### Task 3：时间与 checkpoint 条件性结果

**Files:** 新 statetime/policy.go、evidence.go、evaluate.go、policy_test.go。
**Interfaces:** 完成 EvaluateModel、EvidenceSummary、Summary/Clear；私有
`evaluateFacts(in ModelInput, facts *verifiedFacts) (EvidenceSummary,error)`。
EvaluateModel 拥有并 defer 清理 verifiedFacts，成功只转移固定值快照。

- [x] 写 TestTimePolicyWindows、TestTimePolicyCheckpoint、TestTimePolicyNoAuthority：固定 now，检查起点接受、终点拒绝、now 位于终点前 1ns 接受；metadata/key 边界独立覆盖；未来 issued 即使仅 +1 秒仍拒绝。
- [x] 增加 Review Focus：年份 1/9999 的极远差与零时间拒绝；不得对溢出的绝对 duration 取负而接受。issued 为 now-5m 时因最大 5m 窗口已结束仍拒绝；分别测试偏差边界与最终窗口判定，不能编造同时合法的边界正例。
- [x] checkpoint 6/7/8 对本地 7 分别拒绝/不推进/拟议推进；metadata creation sequence 任意合法更大值不影响比较。新请求同 checkpoint 可成功；同一未消费请求连续 EvaluateModel 可成功，但输入时间、floor、checkpoint 不变，无真实消费声明。
- [x] 运行 `& $attestationGo test ./internal/nodecontrol/statetime -run '^TestTimePolicy' -count=1`，记录 RED。
- [x] 实现 evaluateFacts 和 EvaluateModel，错误时间范围/低于 Floor 为 ErrInvalidTime，合法时间但不在当前窗口为 ErrExpired；较低 checkpoint 为 ErrCheckpointRollback。其他绑定/签名检查先于 policy；有效期输出为 attestation.ValidUntil（已验证不超 metadata/key）。
- [x] 运行同组及完整 statetime GREEN；验证 nil/清除 Summary 零值和无可变切片外泄，记录检查点。

### Task 4：重检、取消和所有权

**Files:** 修改 statetime/evaluate.go、evidence.go；新 evidence_test.go。
**Interfaces:** 完成 RecheckModel；快照保存完整固定值观测及初次时间，
Recheck 不更新初次证据，不消费 nonce；调用者重新 Evaluate 才产生新证据。

- [x] 写 TestTimeEvidenceRecheck：同输入成功，时间观测 ID 更新且 Now/Floor 不下降成功；时间提供方改变、过期、Floor/Now 下降失败。保持原 ObservationID 而改 RequestID/nonce/checkpoint/metadata 状态逐项失败。
- [x] 写 TestTimeEvidenceOwnership、TestTimeEvidenceCancellation：外部 envelope 修改不影响既有 Summary，但重检篡改输入失败；Clear 两次及 nil 安全；清除后重检 ErrStaleEvidence；原始结果与重检临时结果均无借用，取消绝不返回成功。测试取消使用确定性 context，而非 sleep 竞态。
- [x] 运行 `& $attestationGo test ./internal/nodecontrol/statetime -run '^TestTimeEvidence' -count=1`，记录 RED。
- [x] 实现 RecheckModel：先检查 context/证据有效性，再调用 EvaluateModel 重新解析验签；成功结果 defer Clear，比较所有非时间观测/全部 P/E/完整摘要。允许新的时间观测 ID，其余必须满足设计 §7；不匹配返回 ErrStaleEvidence，重新验证本身失败则保留其固定错误。
- [x] 运行同组和完整 statetime GREEN，记录所有权/取消限制，不声称并发 Clear 安全。

### Task 5：隔离门禁、限定回归与最终审查

**Files:** 新 statetime/boundary_test.go；完成后更新本计划、设计状态、
docs/roadmap/current-status.md，新增 docs/roadmap/2026-10-09-time-attestation-execution.md。
**Interfaces:** 不增加产品接口；AST helper `timeImportsAllowed([]byte) bool`
和 `timeConsumerAllowed(path, packageName, importPath string) bool` 只在测试中。

- [x] 写 TestTimeBoundary 及正反 helper 案例：允许纯类型 time.Time/显式时间计算；拒绝 time.Now/Since/Until/timers（包含别名和函数值获取）、dot import、net/os/database、statepolicy/stateverify；禁止任意非测试文件导入 statetime，精确比较路径而非前缀。遍历仓库产品 Go 文件，跳过 .git、.worktrees、vendor，不漏 cmd。
- [x] 先观察 helper 缺失或注入禁止样例的 RED，再实现测试 helper；不向产品源临时注入危险调用。运行 `& $attestationGo test ./internal/nodecontrol/statetime -run '^TestTimeBoundary' -count=1` GREEN。
- [x] 对本次改动 Go 文件运行固定工具链 gofmt；重新比对受保护 hash，仅设计授权的文件允许变化。检查新文件及 tracked diff 的空白错误，不修改安全目录设置。
- [x] 运行 `& $attestationGo test ./internal/nodecontrol/statetime ./internal/nodecontrol/stateformat ./internal/nodecontrol/stateverify ./internal/nodecontrol/statepolicy ./internal/nodecontrol/serving ./internal/nodecontrol/contracts ./internal/readiness -count=1`，须所有包 ok、exit 0。
- [x] 运行 `& $attestationGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1` 及 `& $attestationGo test ./internal/store -run '^$' -count=1`；后者明确只编译、no tests to run。
- [x] 按 requesting-code-review 技能提供批准设计、计划、精确增量、基线与测试证据；Native 仅此一次 fresh 最强可用模型最终独立审查。审查只读、不扩大测试权限。修复发现的问题需先复现再回归，实质设计变化先请求批准。
- [x] 写执行记录：每次 RED/GREEN、命令/退出码、作者与审查者各自实际验证范围、发现及修复、未判断项、未完成生产门禁；不以有条件模型测试冒充 TLS/防重放/硬件/整仓验收。
- [x] 更新状态报告仅新增本增量，保留既有历史；再次保护 hash/格式/差异检查。逐项标完成，仅实际通过才勾选。交付文件与事实，不提交或推送。

## 计划自查与交接

覆盖：设计 §3→任务 1；§4–5、§8 错误→任务 2；§6→任务 3；
§7–8 所有权→任务 4；§8 隔离及 §9–10→任务 5。
Review Focus 五项均已映射到具体测试。无独立新子系统或生产集成混入。
沿用此前 Native 选择及 executing-plans；实际证据见执行记录。
