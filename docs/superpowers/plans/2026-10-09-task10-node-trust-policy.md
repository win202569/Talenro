# Node Trust Policy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现隔离、条件性的节点信任状态判定模型，不产生生产授权。

**Architecture:** 新建 statepolicy，复用 stateverify 验签和 stateformat 唯一解析器。平台只作为显式模型观测输入，核心不执行外部 I/O；候选绑定旧状态、原始证据和时间条件，不能作为提交回执。

**Tech Stack:** Go 1.26.5、标准库、现有 uuid/contracts/stateformat/stateverify；PowerShell；离线测试，无新增依赖。

**Spec:** `docs/superpowers/specs/2026-10-09-task10-node-trust-policy-design.md`（用户已批准）。

**状态:** 用户已批准，Native 隔离增量已完成；详细实证见 `../../roadmap/2026-10-09-node-trust-policy-execution.md`。独立审查的两项 Important 已由作者 RED/GREEN 修复，未二次审查。未提交、推送或生产验收。

## Global Constraints

- 工作目录固定为 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`，保留全部现有未提交工作；不新建工作区，不覆盖现有 fixture。
- 不新增网络协议、wire schema、依赖或锁文件，不改变已有签名接受语义。
- 现有 stateverify 接受从初始 v1 开始的完整链，最多 32 次轮换。
- 本阶段不实现 TPM、可信部署认证、跨重启时间源、双槽持久化、硬件提交、TLS、LocalSecurityLatch 或生产接入。
- 模拟适配器只置于测试代码，不注册为运行时默认提供方。
- 失败或取消不返回部分候选；重复输入也必须检查时间和保障条件。
- 验收成功只能表述为“在显式模型假设下，隔离判定规则通过指定测试”。
- 不暂存、提交、推送、合并或清理；本约束覆盖技能中的默认 commit 步骤。
- 不运行真实 DB/Docker、全仓、C11/C12、bootstrap、PITR 或 race 测试，不改变安全设置，不绕过已有拒绝。

## Review Focus

以下五项已分别纳入所属任务的 RED/GREEN 测试，而非只留给审查者判断：

1. 已取消 context 与 malformed 输入同时存在：取消优先且无部分输出（任务 1、5）。
2. 时刻恰好等于 expiry 或 now+5m：半开区间拒绝 expiry，5m 边界允许，溢出拒绝（任务 2）。
3. 不同合法签名封装的相同 payload：重新验签后幂等，不误报内容分叉（任务 3）。
4. 缺失后重新出现的 key、跳过 retiring 版本：保留历史约束，未知不能默认通过（任务 4）。
5. 候选旧状态 token 不变但状态内容变了：纯模型重检仍拒绝，不能仅凭 token 相等通过（任务 5）。

## 文件结构与接口契约

所有下列代码路径相对于工作目录；新增文件集中在 `internal/nodecontrol/statepolicy/`。

| 文件 | 责任 |
| --- | --- |
| doc.go、input.go、errors.go、model.go | 条件性边界、输入、有限错误分类及模型类型 |
| facts.go、owner.go | 调用现有验签、解析并绑定完整事实，管理拥有的数据 |
| time.go | 显式时间观测及有效期判定 |
| transition.go | 根前缀、对象高水位、幂等及冲突判定 |
| history.go | 密钥属性、状态、撤销及保留历史 |
| evaluate.go、candidate.go | 组合判定、候选所有权、纯重检 |
| 对应 *_test.go、fixtures_test.go、boundary_test.go | 定向规则、固定样例、测试模型及生产依赖边界 |

不计划修改 stateformat/stateverify 产品代码。新包测试只读使用 `../stateverify/testdata/` 的固定样例；可在新包 fixtures_test.go 内移植必要的 testPrivate/mint/signedBatch 测试辅助逻辑，注明来源，不导出现有测试 helper、不运行 fixture 生成脚本。

### 公共接口（全部仅模型语义）

```go
func EvaluateModel(ctx context.Context, in ModelInput) (*Candidate, error)
func RecheckModel(ctx context.Context, c *Candidate, in ModelInput) error
func (c *Candidate) NoChange() bool
func (c *Candidate) State() *StateModel // 深拷贝；nil/已 Clear 返回 nil
func (c *Candidate) ValidUntil() time.Time
func (c *Candidate) Clear()
func (s *StateModel) Clear()
```

`Candidate` 全部字段私有，不可按值复制 owner；不存在 Accepted/Authorized/Commit 方法。
私有固定 model 标记用于拒绝零值候选，不被描述为不可伪造的安全凭据。

`ModelInput` 字段固定：`Batch stateverify.Batch`、`Deployment DeploymentObservation`、
`Time TimeObservation`、`Old *StateModel`、`History []KeyHistoryObservation`。
输入在调用中只借用且须保持不变，不支持并发 mutation/Clear。

- `ObservationStatus uint8`：Unknown=0、Available、Unavailable、Rollback；零值不能通过。
- `DeploymentObservation`：Status、`SourceID/ObservationID/AnchorPayload contracts.Digest`、`NodeID uuid.UUID`、`Epoch uint64`、`FreshInstall bool`。FreshInstall 是模型前提，不证明部署认证。
- `TimeObservation`：Status、`ProviderID/ObservationID contracts.Digest`、`Now/Floor time.Time`。Available 还要求身份非零、时刻有效、Now>=Floor。
- `HighWater`：`Version/Sequence uint64`、`Payload contracts.Digest`。
- `RootRecord`：`Mark HighWater`、`Root stateformat.RootMaterial`；保存初始根及所有已接受轮换，比较材料而非只比较根版本。
- `KeyRecord`：`Material stateformat.MetadataKeyMaterial`、`FirstRevokedAt time.Time`、`LastSnapshotUntil/LastCompositeDeadline time.Time`、`DeadlinesKnown bool`。旧记录按 key ID 排序、唯一；缺失于当前 metadata 不删除历史记录。
- `StateModel`：Status、`StateID/CounterID/DeploymentSourceID/TimeProviderID/AnchorPayload contracts.Digest`、`NodeID uuid.UUID`、`Epoch/StateEpoch uint64`、`TimeFloor time.Time`、`Roots []RootRecord`、`Metadata/Desired HighWater`、`Keys []KeyRecord`、`RevokedIDs []contracts.Digest`。
- `KeyHistoryObservation`：Status、`ObservationID/BaseStateID/MetadataPayload/KeyID contracts.Digest`、`RetiringKnown/DeadlinesKnown bool`、`LastSnapshotUntil/LastCompositeDeadline time.Time`、`IncidentID uuid.UUID`、`IncidentKind string`、`IncidentOpen bool`。只对当前旧状态、目标 metadata 和对应 key 生效；未知/重复/错作用域不得成为证明。

StateID/StateEpoch 是提供方旧快照的身份，不由本包生成硬件提交身份。Candidate.State() 中保留旧身份并携带拟议业务字段；首次模型身份保持未提交状态。返回值不能直接用作已持久化旧状态；测试模拟提交须显式赋予新身份。RecheckModel 同时比较捕获的完整旧状态，不能只比较 StateID。

新包只定义观测读取边界，不实现或在核心调用它们：
`DeploymentObserver.ObserveDeployment(context.Context) (DeploymentObservation,error)`、
`TimeObserver.ObserveTime(context.Context) (TimeObservation,error)`、
`StateObserver.ObserveState(context.Context) (*StateModel,error)`。
接口成功不代表观测真实；独立来源认证、一致快照和原子提交仍是未来工作。

错误为 `*PolicyError{Code Code, Reason Reason}`，Code 为 InvalidInput、Rejected、AssuranceMissing、SecurityFault、Canceled；Reason 使用有限枚举，不能包含原始 payload/凭据。测试用 `errors.As` 判别。非法/缺失材料和已认证冲突严格区分。

## 执行与测试环境

用户批准计划后，首先读取 executing-plans、TDD 及适用工作区指令。记录 scoped 文件 hash 基线和 Git 状态，不改 safe.directory/权限；Git 读受限时仅申请精确只读提权，不能借此扩大写权限。

所有测试在上述工作目录运行；每个独立 PowerShell 调用先设置：

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
$policyGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
```

下述命令都以退出码 0 且测试 PASS 为 GREEN；RED 必须是预期缺失接口或断言失败，不能把依赖、环境或语法错误当作行为 RED。首次新增接口的 undefined 编译失败可记录为接口 RED，后续已有接口变更必须得到断言 RED。不下载缺失依赖；环境阻塞单独报告。

## Task 1：输入、拥有关系与精确证据绑定

**Files:** 新建 doc.go、input.go、errors.go、model.go、facts.go、owner.go、facts_test.go、fixtures_test.go。

**Interfaces:** 消费 `stateverify.Verify(ctx, Batch)`、`stateformat.ParseEnvelope(kind, []byte)`；产出
`loadFacts(ctx context.Context, b stateverify.Batch) (*verifiedFacts,error)`、`(*verifiedFacts).Clear()`。
verifiedFacts 私有：anchor、rotations、metadata、desired 的 Bindings、ArtifactIdentity、根/密钥材料；仅成功完成 Verify 和逐项 digest/坐标绑定后可返回。

- [x] 写 TestLoadFactsBindings、TestLoadFactsCancellation、TestModelOwnership：替换每类原始对象、nil context、取消优先、清理两次和嵌套 slice 不共享；使用固定有效签名 Batch。
- [x] RED：`& $policyGo test ./internal/nodecontrol/statepolicy -run '^(TestLoadFacts|TestModelOwnership)' -count=1`；记录预期缺失接口或绑定断言失败。
- [x] 按上述类型契约实现模型和 `loadFacts`；先检查取消、尺寸/计数，再 Verify，逐项复用 ParseEnvelope，owner 在取得资源前登记清理。保持 262144-byte envelope、65536-byte payload、32 rotations 既有限制；超过链范围返回 AssuranceMissing/UnsupportedRange，其余格式尺寸违规 InvalidInput。
- [x] GREEN：重跑同一命令；明确坏签名不生成 verifiedFacts。加入断言 `if err != nil && facts != nil { t.Fatal("partial facts") }`。
- [x] 记录本任务变更清单和 RED/GREEN，不提交。

## Task 2：可信时间观测与有效期

**Files:** 新建 time.go、time_test.go。

**Interfaces:** 消费 verifiedFacts、TimeObservation；产出
`checkTime(f *verifiedFacts, observation TimeObservation, old *StateModel, acceptedRoots int) (time.Time,error)`。
acceptedRoots 包含初始根，首次为 0；不重复检查已接受轮换的当前有效期，始终检查当前 metadata/desired。返回 desired.ValidUntil，不产生新的 floor。

- [x] 写 TestPolicyTimeBoundaries、TestPolicyTimeUnavailable、TestPolicyTimeNoRatchet；时间表覆盖 issued-1ns/issued、until-1ns/until、desired issued=now+5m/+5m+1ns，历史轮换过期与新增轮换过期。加入 `if !deadline.Equal(desiredUntil) { t.Fatal("deadline extended") }`；输入时间观测调用前后完全相同。
- [x] RED：`& $policyGo test ./internal/nodecontrol/statepolicy -run '^TestPolicyTime' -count=1`。
- [x] 实现 checkTime：元数据/新增轮换 `issued<=now<until`，desired `issued<=now+5m` 且 now<until；复核 24h/meta/key 上界。Available 外非通过；Rollback 为 SecurityFault，Unavailable/Unknown 为 AssuranceMissing；old 非 nil 时观测 provider 必须等于 old.TimeProviderID，Floor 不得低于 old.TimeFloor；Now 不低于观测 Floor。时刻限有效 UTC 年 1..9999，所有加法先检查结果范围，零值拒绝。
- [x] GREEN：同命令；额外证明新 metadata 可以授权更早签发但仍有效的 desired，且 payload 时间不能改变 Now/Floor。
- [x] 记录测试证据，不提交。

## Task 3：部署、根前缀及独立高水位

**Files:** 新建 transition.go、transition_test.go。

**Interfaces:** `checkTransition(f *verifiedFacts, in ModelInput) (acceptedRoots int, changed bool, err error)`；验证部署观测、旧状态结构、根前缀和 metadata/desired 高水位，不修改输入。

- [x] 写 TestPolicyDeployment、TestPolicyRootPrefix、TestPolicyHighWater、TestPolicyEnvelopeIdentity：初次无部署/非 FreshInstall 拒绝、已存在状态下 FreshInstall 不重置、跨 epoch 拒绝；0/1/32 链、缺前缀/替换/截断；两个对象各自的低值、同坐标内容冲突、正常递增和 envelope-only 差异。断言各对象 stream 的无关 sequence 大小不影响结果。
- [x] RED：`& $policyGo test ./internal/nodecontrol/statepolicy -run '^TestPolicy(Deployment|RootPrefix|HighWater|EnvelopeIdentity)' -count=1`。
- [x] 实现 checkTransition：身份检查在内容冲突诊断前；旧模型计数器/状态身份非零且 Status Available、根列表连续；比较旧前缀 payload/sequence/根材料。检查 Metadata 和 Desired 自身双高水位；等版本或等 sequence 却改变内容为 SecurityFault；更低且不构成复用冲突为 Rejected。新增根 sequence 顺序沿用 Verify 保证。
- [x] GREEN：同命令；同内容另一个有效 envelope 必须重新验签后 changed=false；无效替代封装应 Rejected 而非 SecurityFault。
- [x] 记录断言证据，不提交。

## Task 4：密钥历史、撤销与保留期限

**Files:** 新建 history.go、history_test.go。

**Interfaces:** `checkHistory(f *verifiedFacts, in ModelInput) (keys []KeyRecord, revoked []contracts.Digest, err error)`；结果为独立深拷贝，失败返回 nil slices。以 key ID 排序合并，不能凭 map 迭代顺序选择错误。

- [x] 写 TestPolicyKeyHistory、TestPolicyRevocation、TestPolicyRetention：属性变更、retiring->active、revoked 复活、缺失后出现、revoked 集合删项、缺历史期限、过早 scheduled、incident 缺失/关闭/类型或 ID 错、观察绑定旧状态/metadata 错误、重复 key；保留 30d 的边界和最后 snapshot.until 相等时不可删除。
- [x] RED：`& $policyGo test ./internal/nodecontrol/statepolicy -run '^TestPolicy(KeyHistory|Revocation|Retention)' -count=1`。
- [x] 实现 checkHistory：正常 active->retiring->revoked，重复状态允许；跳版无法证明中间状态时必须有 exact-scoped RetiringKnown 观测，否则 AssuranceMissing。撤销期限未知不猜测；scheduled 的 revoked_at 不早于已知 composite deadline。emergency 绑定 exact open online_signer_equivocation，成立也返回 SecurityFault/EmergencyRevocation 而非正常候选。
- [x] 实现保留：FirstRevokedAt 不因更新延后/重置；删除 tombstone 要同时满足 now>=FirstRevokedAt+30d 且 now>LastSnapshotUntil，保留累计 ID 和历史约束。测试 ledger 4095/4096/4097；达到需重新部署的容量状态返回阻塞，不能清空。旧密钥历史超过可表达边界时阻塞，不静默丢失；本模型历史记录上限 4096，超出视 UnsupportedRange，不作为生产保留策略。
- [x] GREEN：同命令；失败输入与旧 key records 不变；缺失密钥不继续授权、历史仍保留；历史片段只对目标 key/旧 StateID/目标 metadata payload 生效。
- [x] 记录测试证据，不提交。

## Task 5：完整组合、陈旧候选重检、定向回归及独立审查

**Files:** 新建 evaluate.go、candidate.go、evaluate_test.go、candidate_test.go、boundary_test.go；更新 `docs/roadmap/current-status.md`，新增 `docs/roadmap/2026-10-09-node-trust-policy-execution.md`；仅在执行时勾选本计划。

**Interfaces:** 实现前述 EvaluateModel、RecheckModel 和 Candidate 方法。RecheckModel 消费同一原始 Batch 与新的观测/当前旧状态；不是提交函数。

- [x] 写 TestEvaluateModel、TestCandidateRecheck、TestPolicyBoundary：完整正例、重复 no-change、所有分类错误 nil candidate；替换 Batch、旧 token 变动、token 未变但旧内容改变、时间过期/provider 改变、零值/Clear 后 candidate、修改返回 State 拷贝不影响 owner。核心断言：`if err != nil && c != nil { t.Fatal("partial candidate") }`，陈旧候选必须非 nil error。
- [x] RED：`& $policyGo test ./internal/nodecontrol/statepolicy -run '^(TestEvaluateModel|TestCandidateRecheck|TestPolicyBoundary)' -count=1`。
- [x] 实现 EvaluateModel 固定顺序：入口取消/结构和范围 → loadFacts → checkTransition → checkTime → checkHistory → 最终取消检查 → 分离拥有的 candidate。阶段内按固定对象顺序首个失败；库 ErrCanceled 映射 Canceled，格式错误映射 InvalidInput，签名/绑定错误映射 Rejected。任何路径不输出部分候选。NoChange 按完整业务高水位/根前缀/密钥历史/撤销集合比较，不能仅用 checkTransition 的 changed；观测 ID 或采样时刻变化不构成安全状态推进，时间 floor 仍由平台负责，不作为本包计数器写入理由。
- [x] 实现 RecheckModel：拒绝无有效 model 标记对象，逐字段深比较捕获的旧状态和输入旧状态，再重新 EvaluateModel；要求原始证据身份、部署身份、提供方身份及拟议业务状态与原候选一致。允许刷新时间观测 ID，但 Now 不低于原观测、Floor 不低于原 floor、Now 未到期；不推进候选、状态或计数器。
- [x] 实现边界测试用 Go AST 扫描新包非测试源文件，无网络/文件/数据库/硬件 I/O、time.Now、time.Since、time.Until、生产 observer 实现；全库源码搜索确认新包无生产调用方。规则不得禁止现有纯标准库时间类型或测试读取固定 fixture。
- [x] GREEN：`& $policyGo test ./internal/nodecontrol/statepolicy -count=1`；再执行现有允许范围：

```powershell
& $policyGo test ./internal/nodecontrol/stateformat -count=1
& $policyGo test ./internal/nodecontrol/stateverify -count=1
& $policyGo test ./internal/nodecontrol/serving -count=1
& $policyGo test ./internal/nodecontrol/contracts -count=1
& $policyGo test ./internal/readiness -count=1
& $policyGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1
& $policyGo test ./internal/store -run '^$' -count=1
```

- [x] 使用固定 gofmt 仅格式化本增量 Go 文件；确认格式检查为空、diff whitespace 检查通过、原有非本增量文件 hash 未变。每条测试记录独立退出码，不能仅采用命令序列末尾退出码。未完成测试不得写 PASS。
- [x] 按 Native 安排一次新上下文最强可用模型独立审查（当前 gpt-6-astra）；审查只读，范围为本增量加依赖契约，不把分支全部历史差异当本次实现。告知其测试证据与未验证限制，禁止扩展测试/改安全设置。发现问题则主代理修复并重跑受影响获准测试。
- [x] 更新状态文档与执行记录：列明实际 RED/GREEN、审查结论、遗留问题、真实平台阻塞、未提交/未推送。保存审查记录到本增量工作记录，不宣称整体 Task 10/B01/C12 或 main 验收。

## 自检与交接

已将设计第 3 节映射到任务 1/5，第 4 节映射到任务 3/4，第 5 节映射到任务 2，第 6 节映射到任务 5，第 7 节映射到各任务与定向回归。五项 Review Focus 均有对应测试；签名事实、观测、旧状态、候选类型边界一致。

本计划没有实现真实平台接口，也未承诺硬件或持久化测试；明确保留相应验收阻塞。测试 fixtures 与平台模拟限定于测试文件。代码增量不得顺便修复此前两个无关测试 Minor。

请用户审阅本计划是否准确表达需求。用户批准后按已选 Native 执行，不再询问执行方式；批准本计划不授权提交、推送、生产接入或扩大受阻测试范围。
