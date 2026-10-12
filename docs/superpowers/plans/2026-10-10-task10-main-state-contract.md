# Task 10 Main State Contract L0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task using the preserved Native method. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现十二域主状态的逻辑覆盖和结构一致性模型，始终保留 L1–L5 未完成责任，不生成提交或启动权限。

**Architecture:** 新建纯包 `internal/nodecontrol/statecontract`，分离字段形状、跨域引用、previous/proposed 基本高水位比较和私有结果所有权。所有材料只作显式类型化引用，不解析/验签原始材料，不把结构齐全解释为完整 sealed blob。

**Tech Stack:** 固定 Go1.26.5、PowerShell、既有 UUID/contracts 基础值类型；无新增依赖。

**Spec:** `docs/superpowers/specs/2026-10-10-task10-main-state-contract-design.md`，用户已批准。

状态：2026-10-10 Native 本地实现、一次独立源码审查及一次修复已完成，批准范围回归通过。保留两项审查Minor（集合所有权测试覆盖、嵌套覆盖错误分类），未提交或推送；不代表生产或整体项目验收。详见执行记录。

后续更新：同日独立小修补已关闭上述两项Minor，见 `docs/roadmap/2026-10-10-main-state-contract-followup.md`。19个集合逐一验证，错误分类先复现再修复，限定回归通过；未再次独立审查，历史状态保留。

## Global Constraints

- 工作目录 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`；保留既有修改，不暂存、提交、推送、合并或清理工作区。
- 本增量仅 L0；不实现 L1–L5、codec/hash/seal、原始材料读取、网络、provider、磁盘/硬件或启动编排。
- 模型版本固定为 1；authority epoch、流版本/generation/creation sequence 取 `1..MaxInt64`。本包不接收或推进 NV counter，不能借用其 uint64 数域接受超界版本。
- UUID 非零、小写规范 RFC4122；身份摘要为非零32字节。三个 guard ID 彼此不同，三个 key ID 彼此不同；这种声明检查不证明平台隔离。
- 同 node/authority/identity epoch 和部署/trust-anchor 范围内比较。初始化、迁移、reenroll/reset/key replacement 不由本包授权。
- 不依赖或修改 statepolicy/statetime/staterecovery/stateformat/stateverify；不解除它们的隔离守卫，不添加生产消费者或候选转换函数。
- Root keys1–5、threshold1..key数、单次 rotation chain0–32、metadata keys1–64、revoked IDs0–4096、每 CA 累计集合0–128、desired/LKG slots0–8、recovery local/supervisor binding0–64/0–16。
- 限长先于复制；排序唯一输入不得自动修正、截断。时间是显式逻辑值，不调用时钟，不作当前有效性判断。
- nil context 拒绝，预取消优先，成功返回前再次检查；错误返回 nil 结果，无副作用。
- 结果持有独立快照、Clear nil-safe且幂等，不保证并发修改/Clear或运行时所有副本擦除。
- Native 主代理实现，最终一次 fresh 最强可用模型只读审查。测试只在已批准范围内，不跑整仓、DB/Docker、race、硬件/reboot、C11/C12。
- 生产兼容性 UNKNOWN；本增量不解除 Task10/B01/C12/main 验收门禁。技能模板的 commit 步骤用本地证据检查点代替。

## Review Focus

1. Present→NeverEstablished 或清空当前 LKG 时遗失高水位：任务3 `TestContractCoverageRegression`、`TestContractLKGHistory`。
2. 同版本同序列换 payload，以及同一材料身份被赋予不同版本：任务3 `TestContractForks`；不得因没有原始 bytes 忽略可见矛盾。
3. 比较/重检时 nil 与显式空集合误报变化、输入切片借用：任务4 `TestContractOwnership`、`TestContractRecheck`。
4. checkpoint 被错误当成每个流的创建序列，或 certificate serial 被当作版本：任务2/3 `TestContractIndependentCounters`。
5. 齐全结构返回可提交标志、空责任清单，或包被生产导入：任务4/5 `TestContractNoAuthority`、`TestContractBoundary`。

## 文件映射

全部新 Go 文件位于 `internal/nodecontrol/statecontract/`：

| 文件 | 职责 |
| --- | --- |
| doc.go、input.go、errors.go | 不授权声明、公共固定类型、固定错误 |
| shape.go | 单个内容模型的形状与限长检查 |
| bindings.go | 当前模型内的跨域引用一致性 |
| transition.go | 同作用域 previous/proposed 基本单调性与冲突 |
| model.go、evaluate.go | 拥有的快照、摘要、清理、重检和入口 |
| fixtures_test.go、shape_test.go | 独立手工值夹具与结构反例 |
| bindings_test.go、transition_test.go | 引用/高水位反例 |
| model_test.go、boundary_test.go | 取消/所有权/不授权/AST隔离 |

交付时新增 `docs/roadmap/2026-10-10-main-state-contract-execution.md`，更新本计划、设计状态和 `docs/roadmap/current-status.md`，保留历史。

## 固定公共接口与数据契约

以下是逻辑模型类型，不是生产 wire schema。表中 `Digest` 均写为既有 `contracts.Digest`；类型声明中不创建可与其他身份互换的新通用字符串字典。

### 基础类型

- `Coverage uint8`：Unknown=0、Present、NeverEstablished。`Domain uint8`：DomainUnknown=0，再按下方十二域表顺序定义 DomainCheckpoint 至 DomainLKG。
- `SectionHeader { Domain Domain; Coverage Coverage }`，每个域以具名字段 Header 持有；检查 tag 与所在顶层字段精确相等。
- `GuardBinding { CounterID, SealKeyID Digest }`。
- `MainStateScope { NodeID string; AuthorityEpoch uint64; DeploymentID, TrustAnchorID, DeploymentKeySetID, TimeProviderID Digest; Main, Supervisor, Latch GuardBinding; TimeFloor OptionalTime }`。
- `OptionalTime { Present bool; UnixSeconds int64 }`；Present=false 时值必须0。时间使用显式 UTC Unix 秒，合法范围 `-62135596800..253402300799`；1970的0秒是合法时间，不作缺失标记。`TimeWindow { Issued, Until int64 }` 要求范围内且 Issued<Until，不判断是否当前过期。
- `ArtifactKind uint8`：ArtifactUnknown=0；RootPayload、RootRotation、Metadata、MetadataHistory、CertificateDER、IdentityHistory、TrustBundle、ApprovedManifest、InstalledMap、ResourceEnvelope、HostMemoryPolicy、DesiredSnapshot、RecoverySnapshot、RecoverySession、SupervisorState、TransitionCapsule、LKGHistory、ProcessSpec、ConfigMaterial、ReleaseMaterial。
- `ArtifactRef { Kind ArtifactKind; Payload, Envelope Digest }`；RootRotation/Metadata/TrustBundle/ResourceEnvelope/HostMemoryPolicy/DesiredSnapshot/RecoverySnapshot 为签名引用，P/E 均非零；其余为完整材料或payload引用，Payload非零、Envelope必须0。此分类不暗示本包已经验签或认可各物理格式。
- `StreamMark { Version, Sequence uint64; Material ArtifactRef }`；`VersionMark { Version uint64; Material ArtifactRef }`。两者分开，禁止给 manifest/map 编造 authority sequence。
- `DesiredBinding { Generation, Sequence uint64; Payload Digest }`，用于 exact desired 绑定；不是 ArtifactRef，无 envelope，完整材料身份另由对应引用承载。

### 十二域类型及最低结构检查

每个类型都有 `Header SectionHeader`；下面只列其余字段。除有明确 NeverEstablished 分支，列出的摘要/版本/UUID均必需。绑定字段的业务真实性留在未完成责任中。

| MainStateContentModel 字段 / 类型 | 固定字段 | L0 规则 |
| --- | --- | --- |
| Checkpoint / CheckpointSection | `NodeID string; AuthorityEpoch, Sequence uint64` | node/epoch与Scope相同；Sequence独立正整数 |
| Root / RootSection | `Mark StreamMark; Threshold uint8; KeyIDs []Digest` | kind RootPayload；1–5非零排序唯一keys；有效threshold；anchor身份在Scope |
| Metadata / MetadataSection | `Mark StreamMark; RootVersion uint64; RootPayload Digest; KeyIDs, RevokedIDs []Digest; RevokedSetDigest Digest; History ArtifactRef; Window TimeWindow` | kind Metadata；History kind MetadataHistory，Payload非零、Envelope为零；keys1–64、revoked0–4096；不凭keys ID列表推导key状态 |
| Identity / IdentitySection | `IdentityEpoch uint64; LineageID string; LineageUntil int64; IssuerID, Serial, PublicKey, KeyHandle Digest; Leaf ArtifactRef; History ArtifactRef; Window TimeWindow` | Leaf kind CertificateDER；History kind IdentityHistory；Serial为opaque精确标识摘要，不比较数值；证书状态/切换/重试在History完整材料责任内 |
| ServerTrust / ServerTrustSection | `Bootstrap, Agent BundleRef` | 两个字段固定purpose，不能省略或互换 |
| Releases / ReleaseSection | `BuildID Digest; Manifest, Map VersionMark` | kind ApprovedManifest/InstalledMap，不接受poll来源变更授权声明 |
| Resources / ResourceSection | `Mark StreamMark; NodeID string; AuthorityEpoch uint64; DeploymentKeyID Digest` | kind ResourceEnvelope；绑定Scope；只记录key身份，不证明role正确 |
| HostPolicy / HostPolicySection | `Mark StreamMark; NodeID string; AuthorityEpoch uint64; DeploymentKeyID Digest` | kind HostMemoryPolicy；同上，不执行系统readback |
| Desired / DesiredSection | `Mark StreamMark; InventoryVersion, ResourceVersion uint64; ResourcePayload, RootPayload, MetadataPayload, SigningKeyID Digest; Window TimeWindow; SlotIDs []string` | kind DesiredSnapshot；slots0–8，排序唯一ASCII ID |
| Recovery / RecoverySection | `Mark StreamMark; IdentityEpoch uint64; RecoveryID string; SessionVersion uint64; Session ArtifactRef; RootPayload, MetadataPayload, SigningKeyID Digest; Window TimeWindow; Local []LocalFaultBinding; Supervisor []SupervisorFaultBinding` | kind RecoverySnapshot/RecoverySession；session完整材料承诺reason/action/完整incident集合，L0不解析或执行；binding0–64/0–16 |
| Supervisor / SupervisorSection | `Identity Digest; Guard GuardBinding; State ArtifactRef; Verified, Applied AppliedRecord; Transition TransitionRef` | State kind SupervisorState；Guard等于Scope.Supervisor；不持有fault清除或易失lease表 |
| LKG / LKGSection | `HighestAccepted DesiredBinding; History ArtifactRef; Entries []LKGEntry` | History kind LKGHistory；entries0–8；Present即使entries为空也保留HighestAccepted/History |

辅助类型：

- `BundlePurpose uint8`：PurposeUnknown=0、BootstrapServer、AgentServer；`BundleRef { Purpose BundlePurpose; TrustDomain string; AuthorityEpoch uint64; Mark StreamMark; DeploymentKeyID Digest; Deauthorized []Digest; DeauthorizedSetDigest Digest }`。kind TrustBundle；AuthorityEpoch等于Scope；domain1–253字节，规范小写DNS标签，每标签1–63、仅a-z/0-9/内部连字符，不允许空标签/尾点；另拒绝四个纯十进制标签且各值0–255的IPv4形式，IPv6的冒号已被字符规则拒绝。累计集合0–128排序唯一。完整CA entries/DER/status由材料责任覆盖，不冒充已验证。
- `AppliedRecord { Coverage Coverage; Desired DesiredBinding }`；NeverEstablished要求Desired全零，Present要求完整绑定。
- `TransitionRef { Coverage Coverage; Material ArtifactRef; PriorApplied, CandidateVerified DesiredBinding }`；NeverEstablished剩余全零，Present要求kind TransitionCapsule与两条完整绑定；只作逻辑链接，不授权rollback。
- `LocalFaultBinding { FaultID, IncidentID string }`；`SupervisorFaultBinding { FaultID, IncidentID string; Evidence Digest }`。UUID规范非零；按FaultID排序唯一；不要求incident唯一（可多fault对应同incident）。
- `LKGEntry { SlotID string; Desired DesiredBinding; Process, Config, Release ArtifactRef; RootPayload, MetadataPayload, SigningKeyID Digest; Deadline int64 }`。三个kind分别 ProcessSpec/ConfigMaterial/ReleaseMaterial；SlotID1–64 ASCII `[a-zA-Z0-9_-]`，按bytewise严格升序。历史授权材料由History承诺，不强制历史root/metadata等于当前版本。
- `RotationRef { PreviousVersion, NewVersion, Sequence uint64; OldRootPayload, NewRootPayload Digest; Material ArtifactRef }`，kind RootRotation，作为输入的更新证明形状，不当作终身root历史。

`MainStateContentModel { ModelVersion uint32; Scope MainStateScope; Checkpoint CheckpointSection; Root RootSection; Metadata MetadataSection; Identity IdentitySection; ServerTrust ServerTrustSection; Releases ReleaseSection; Resources ResourceSection; HostPolicy HostPolicySection; Desired DesiredSection; Recovery RecoverySection; Supervisor SupervisorSection; LKG LKGSection }`。

`ModelInput { Previous, Proposed MainStateContentModel; RootTransition []RotationRef }`。两份内容必须各自结构有效；缺失previous不当作初始化。无 StateEpoch、ScanID、nonce、path、secret、raw bytes 或caller验证成功字段。

注意：History/Session/State等引用保留的是完整材料责任，不能把一个digest称为已经获得完整历史。MetadataHistory是独立的非签名历史材料引用，不复用metadata网络格式。

### 入口、结果与错误

- `EvaluateModel(ctx context.Context, in ModelInput) (*ContractModel, error)`。
- `RecheckModel(ctx context.Context, model *ContractModel, in ModelInput) error`：context检查后，nil/cleared模型返回StaleModel；重新验证输入，失败沿用原校验错误，合法但与已拥有输入不同才返回StaleModel。逐字段比较，不允许调用方在相同ID下换值。它不更新模型，也不代表真实新鲜读取。
- `(*ContractModel).Summary() ContractSummary`、`Clear()`；私有有效标记、独立owned input。nil/cleared Summary返回全零。
- `ContractStatus uint8`：ContractUnknown=0、ContractShapeValid；`Obligation uint8`：ObligationUnknown=0、MaterialClosure、DomainSemantics、CompositionRecheck、PhysicalCommit、RuntimeAuthorization。
- `ContractSummary { Status ContractStatus; Scope MainStateScope; Coverage [12]Coverage; Outstanding [5]Obligation }`。成功固定按以上顺序返回五项责任；没有生产授权、完整state digest或proposal字段。
- 哨兵 `ErrInvalidInput / ErrIncompleteCoverage / ErrBindingMismatch / ErrRollback / ErrConflict / ErrStaleModel`，支持errors.Is。可用 `ContractError { Domain Domain; Cause error }` 包装固定哨兵，Error只返回固定分类文本，Unwrap返回Cause；不得将输入拼接进错误。

错误优先：先context，再模型版本/scope/两份结构，再跨域引用，再previous/proposed比较，最后context。缺域/Unknown/非法NeverEstablished→IncompleteCoverage；kind/范围/排序/字段零值→InvalidInput；tag/作用域/跨引用不符→BindingMismatch；降低→Rollback；同水位不同payload/不可能身份映射→Conflict；失效模型或重检依赖替换→StaleModel。精确双故障的先后顺序不承诺，仅测试单故障分类及context优先。

## 具体跨引用和转换规则

1. 最终审查裁定修正：Metadata可保留较早RootVersion/RootPayload；不得引用未来版本，同版本必须同payload、同payload不得对应不同版本。Desired资源绑定按同样规则允许较早版本；已接受Desired/Recovery的历史root/metadata/key不强求等于当前活跃材料，也不因当前撤销而无法保存。若metadata引用恰等于已给出的当前metadata，则root/key必须匹配其可见声明；未知历史引用继续欠L1/L2。Recovery IdentityEpoch仍相等。相同Desired/Recovery payload在旧新状态间不得改写其原始授权绑定；Desired原始resource/inventory绑定同样不可改写。保存历史绝非新操作授权。
2. Supervisor的Applied Present要求Verified Present，generation/sequence均不高于Verified；同generation必须same sequence/payload。Verified不得超过当前Desired的generation/sequence；同generation匹配exact desired。Desired NeverEstablished要求verified/applied/LKG均NeverEstablished。Transition Present要求PriorApplied等于当前Applied、CandidateVerified等于Verified；其合法性仍未验证。
3. LKG有条目要求Applied Present，条目Desired必须等于Applied或Present transition的PriorApplied；HighestAccepted不得低于任一条目或前一份LKG.HighestAccepted。History是已接受材料闭包责任，不以空entries清空。LKG中旧root/metadata/key不强制等于当前，但仍欠L2授权/撤销校验。
4. 两份Scope除TimeFloor外必须相等，IdentityEpoch必须相同；各固定purpose的TrustDomain不能替换；floor已Present不能消失/降低，false→true仅检查范围且保留时间来源未验证责任，不产生NV写入。
5. Checkpoint不降低。Root、Metadata、Resources、HostPolicy、各Bundle、Desired/Recovery的StreamMark：Version/Sequence任一降低→Rollback；同Version要求同Sequence和Payload；提升Version要求严格提升Sequence。相同Payload被赋予不同Version/Sequence→Conflict；同P的新合法E不据此判定fork，但E变化应使精确Recheck失败。
6. Manifest/Map VersionMark：版本不降低，同版本不同Payload或同Payload不同版本→Conflict；不编造sequence。BuildID改变不能替代上述规则，也不证明新build可信。
7. Present→NeverEstablished拒绝。revoked/deauthorized集合不能删项；不验证仅靠ID无法得知的key/CA状态转换。Identity serial/key/leaf变化不按大小比较，真实性和切换授权仍未履行。同一Leaf.Payload被声明不同issuer/serial/public-key/window→Conflict。
8. Root不变要求RootTransition为空；提升时至多32条，第一条previous/rootpayload接旧Root，逐条new=previous+1（checked）、sequence严格增，每条old/new payload串联，末条new/version/sequence/payload等于新Root。历史版本本身可以大于32，不要求终身历史数组。
9. Supervisor verified/applied及LKG.HighestAccepted按DesiredBinding检查不降低、同generation/sequence不同payload冲突；不会以transition降低verified。Recovery session ID变化不重置generation，不能据此恢复运行。

最终审查补强：形状校验后、一般跨引用前，统一核对两份状态内全部可见DesiredBinding（含transition及每条LKG）：相同generation、sequence或payload必须对应同一完整绑定，generation与sequence相对顺序一致，矛盾为Conflict。完整root chain及旧endpoint不得重复使用root payload，不同rotation link不得复用同一rotation payload。两项均不依赖摘要重算。

## 运行配置与证据

所有测试在固定worktree根执行；每次PowerShell会话使用以下前缀：

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
$contractGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
```

每个任务保存真实RED/GREEN命令、退出码和输出；新接口尚不存在的编译RED如实标记，不伪称每个case都经历行为RED。旧代码正确时不故意破坏它。未结束的session只等待，不重复启动。

### Task 1：固定十二域结构与限长

**Files:** 新doc.go、input.go、errors.go、shape.go、fixtures_test.go、shape_test.go。
**Interfaces:** 定义全部上述输入类型/错误；私有 `validateShape(in MainStateContentModel) error`。任务2/3消费已通过该函数的值，仍不得假设L1已完成。

- [x] 核实已有隔离worktree、分支/HEAD、AGENTS及本计划ledger；保存既有Go/SQL/module文件hash基线，保留全部历史修改。不修改safe.directory。
- [x] 写 `TestContractCoverage`：完整夹具通过；十二域逐一Unknown/错tag拒绝；NeverEstablished只允许三个顶层域及指定子记录，任何残留字段拒绝；零值完整模型拒绝。
- [x] 写 `TestContractShapeBounds`：各数组最大值通过、最大+1拒绝，乱序/重复/零摘要拒绝；root keys0、threshold0/>count拒绝；UUID大小写/非RFC4122、authority0/MaxInt64+1、非法kind、P/E角色混淆拒绝；每个字段域时间范围/倒置window测试。
- [x] 写 `TestContractGuardSeparation`：逐对复用三个counter或key拒绝；不同声明值可通过结构检查但不带平台可信结论。显式空累计集合与Unknown区分；无原始bytes也不能消除后续责任。
- [x] 运行 `& $contractGo test ./internal/nodecontrol/statecontract -run '^TestContract(Coverage|ShapeBounds|GuardSeparation)$' -count=1`，记录RED。
- [x] 实现validateShape及每域小型helper：先所有集合限长再遍历，不复制、不排序、不截断；十二字段分别检查精确tag/coverage/kind。引用历史不展开或验签。
- [x] 同命令GREEN；记录任务检查点，不提交。

### Task 2：模型内跨域引用一致性

**Files:** 新bindings.go、bindings_test.go。
**Interfaces:** 消费任务1类型/validateShape；产出私有 `validateBindings(in MainStateContentModel) error`。不调用provider或其他模型包。

- [x] 写 `TestContractBindings`：逐项替换metadata/root、desired/resource/root/metadata/key、recovery/identity、Supervisor guard、Bundle purpose/domain造成明确拒绝；RecoverySnapshot不得填入Desired。
- [x] 写 `TestContractAppliedAndLKG`：无desired时子状态不能伪装已建立；applied超过verified拒绝；同generation不同payload拒绝；LKG引用非applied拒绝；合法空entries保留History/HighestAccepted；保留旧授权材料的LKG结构不因不是当前root而误拒绝。
- [x] 写 `TestContractIndependentCounters`：checkpoint=100、root sequence=4、metadata=8、desired=20各自合法；不要求sequence全部等于100。保持同identity epoch的两张不同serial证书均可结构有效，不能据serial大小判断回退。
- [x] 运行 `& $contractGo test ./internal/nodecontrol/statecontract -run '^TestContract(Bindings|AppliedAndLKG|IndependentCounters)$' -count=1`，记录RED。
- [x] 实现validateBindings，按九条跨域/转换规则中1–3落实当前模型关系；仅检查可见引用，不将调用方digest变成验签结果。
- [x] 同命令GREEN，完整新包回归；记录检查点，不提交。

### Task 3：旧新模型的可判定高水位约束

**Files:** 新transition.go、transition_test.go。
**Interfaces:** 消费任务1/2；产出私有 `validateTransition(in ModelInput) error`。只在两份内容shape/bindings通过后调用，负责规则4–9。

- [x] 写 `TestContractCoverageRegression`、`TestContractLKGHistory`：Present降为NeverEstablished拒绝；合法停止/移除当前LKG不清高水位；两累计集合删项拒绝，显式空到非空可追加。
- [x] 写 `TestContractForks`：所有StreamMark/VersionMark逐字段降低、同水位换P、同P换水位拒绝；同P不同E不当作payload fork，但不声称E有效。同leaf身份不同issuer/serial/key/window拒绝。
- [x] 写 `TestContractRootTransition`：版本100→101的一条链通过；缺失、跳号、错old/new摘要、sequence不增、超过32及整数边界拒绝；相同root非空chain拒绝；不得把版本100本身当超限历史。
- [x] 写 `TestContractTransitionScope`：node/epoch/deployment/guard/key/time-provider/identity-epoch替换拒绝；floor消失/降低拒绝；恢复session变化不允许generation回退；certificate serial数值顺序不影响独立结构比较结论。
- [x] 运行 `& $contractGo test ./internal/nodecontrol/statecontract -run '^TestContract(CoverageRegression|LKGHistory|Forks|RootTransition|TransitionScope)$' -count=1`，记录RED。
- [x] 实现validateTransition，复用单一StreamMark比较器但只在指定域调用；证书/时间/manifest等分别判定，禁止一刀切max。Root链逐条checked比较，集合用排序输入的包含关系判定，不删减历史。
- [x] 同命令GREEN，完整新包回归；记录检查点，不提交。

### Task 4：私有结果、固定责任与精确重检

**Files:** 新model.go、evaluate.go、model_test.go。
**Interfaces:** 产出EvaluateModel、RecheckModel、ContractModel.Summary/Clear和固定结果类型；消费前三任务helper。私有 `cloneInput(in ModelInput) ModelInput` 拥有所有slice。

- [x] 写 `TestContractNoAuthority`：完整输入只得到ContractShapeValid和固定五项责任；修改返回的数组/摘要不影响下次Summary；调用方没有清责任字段；invalid结果nil、nil/cleared摘要全零。
- [x] 写 `TestContractOwnership`、`TestContractRecheck`：逐个修改输入数组/成员不改变模型；相同ID下换任一依赖/材料E/窗口后Recheck拒绝；相同值重检通过且不更新模型；nil和空合法集合视为同一逻辑输入，不制造stale。Clear双次、nil、清理后重检拒绝。后续 `TestContractOwnershipEveryCollection` 已补齐全部集合。
- [x] 写 `TestContractCancellationAndErrors`：nilcontext、预取消优先、确定性末尾取消无成功结果；错误Is分类正确、不含UUID/输入摘要。无需sleep/真实时钟。
- [x] 运行 `& $contractGo test ./internal/nodecontrol/statecontract -run '^TestContract(NoAuthority|Ownership|Recheck|CancellationAndErrors)$' -count=1`，记录RED。
- [x] 实现EvaluateModel按固定顺序校验后复制输入；合法空slice统一规范化为nil，但不排序非空输入。Recheck重新Evaluate并清理临时结果，再比较完整规范化快照，可用reflect.DeepEqual，不按几个ID代替全值比较。
- [x] 同命令GREEN和完整新包回归；逐一核对每个slice都深复制、Clear置零；记录检查点，不提交。

### Task 5：隔离边界、限定回归和一次最终审查

**Files:** 新boundary_test.go；更新本计划列明的交付文档。
**Interfaces:** 测试专用 `contractImportsAllowed(src []byte) bool`、`contractConsumerAllowed(importPath string) bool`，无新增产品接口。

- [x] 写 `TestContractBoundary`：产品导入只允许context/errors/bytes/strings/reflect、google/uuid和contracts；按所有别名限制context仅Context/Canceled/DeadlineExceeded，uuid仅Parse/Nil/RFC4122，reflect仅DeepEqual，contracts仅Digest，errors仅New；bytes只Compare/Equal，strings只纯字符串检查（Contains/HasPrefix/HasSuffix/Split/ToLower）。拒绝dot/blank import、time、rand、os/net/unsafe/syscall以及现有state模型包。
- [x] 写多别名、函数值取得禁止UUID随机能力、无法解析代码等反例及合法类型正例。遍历仓库全部非测试Go含cmd，跳过.git/.worktrees/vendor，禁止精确statecontract导入；test消费者不在禁止范围。运行 `& $contractGo test ./internal/nodecontrol/statecontract -run '^TestContractBoundary$' -count=1` 观察helper缺失RED，实现后GREEN，不往产品代码注入坏代码。
- [x] 对本次文件执行固定gofmt；核对protected hash、tracked/untracked差异和空白；只有本包与指定文档可变，依赖锁/DB/其他产品保持原样。
- [x] 运行 `& $contractGo test ./internal/nodecontrol/statecontract ./internal/nodecontrol/staterecovery ./internal/nodecontrol/statetime ./internal/nodecontrol/stateformat ./internal/nodecontrol/stateverify ./internal/nodecontrol/statepolicy ./internal/nodecontrol/serving ./internal/nodecontrol/contracts ./internal/readiness -count=1`，所有包ok且exit0。
- [x] 运行 `& $contractGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1` 与 `& $contractGo test ./internal/store -run '^$' -count=1`，exit0；store明确只编译。
- [x] 使用requesting-code-review，向一次fresh最强可用模型只读审查提供批准spec/plan、精确增量、基线、Review Focus、实际证据和范围禁令。按Native流程对Important/Critical先复现再修复并重跑限定回归；Minor列明，不能当作整体通过而隐藏。
- [x] 写执行记录并更新状态，区分作者实测与审查者实际检查；保留历史、所有裁定/风险和未完成层。仅实际完成勾选。保留工作区，不自动暂存、提交或推送。

## 自查与交接

设计§1–4→约束/基础类型/隔离；§5–6→任务1–3；§7结果强度→任务4；§8–9→任务4/5；§10→各命名测试及限定回归。五项Review Focus都有归属。

完整材料的raw bytes、证书/CA真实状态、key窗口/撤销期限、recovery action语义、supervisor fault和LKG启动授权均没有偷偷变为L0通过项；只保留类型化引用和非空验证责任。历史引用的摘要不证明历史存在或完整。

本计划已按Native实施；审查裁定见规则1和补强规则。原有两项Minor已在后续小修补关闭，原审查历史见执行记录。没有新增生产消费者、存储格式或平台门禁豁免。
