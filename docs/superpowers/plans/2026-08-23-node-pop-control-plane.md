# Talenro C1.2 Node and POP Control Plane Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有 Talenro Go 模块化单体中交付可审计的节点/POP 库存、外部权威防回滚、仅出站 mTLS node agent、签名 desired/recovery state、独立 node-core-supervisor、fixture/Xray/sing-box 外部进程适配、健康容量事实，以及绑定 claim-v1 genesis/epoch/staging/source/Down 协议的四作用域权威完成证据。

**Architecture:** PostgreSQL 保存库存、身份、签名意图、恢复、审计与最新 observation 等领域事实；registered Go `00007` 以默认关闭的 additive catalog、migration latch、runtime/incarnation result 链、genesis/epoch/source/staging barrier 与三阶段 Down guard 管理 v6→claim-v1 单向过渡。独立 rollback-resistant provider/attestor 以 stable Head、epoch/sequence、lease、timeline lineage、commit archive/Fence/Inspect 封住 PostgreSQL PITR 回放；B01 只提供 canonical contracts/verifiers/schema和deterministic fakes，B03 实现 production provider/WAL archive runtime，B11 编排外部调用。B10–B11 另以固定 runner profile、machine-sealed bootstrap/install receipt、最小 Git object database、cleanup-first WAL/capsule 和 nonenumerable attested channel 把五类 native runner 的机器能力与 tracked policy 分离。`control-api`、agent、supervisor 与四作用域验收仍按隔离边界交付。

**Tech Stack:** Go 1.26.5、PostgreSQL 18.4、Redis 8.8.1、NATS 2.14.3、pgx 5.10.0、OpenAPI 3.0.3、oapi-codegen 2.8.0、Protobuf Go 1.36.11、Buf 1.72.0、sqlc 1.31.1、goose 3.27.1、Prometheus Go client 1.23.2、Ed25519、ECDSA P-256、RFC 8785 JCS、TLS 1.3、Linux cgroup v2/pidfd/seccomp/nftables/SELinux、Docker Compose。

**Spec:** [Approved C1.2 node and POP control-plane base design](../specs/2026-08-23-node-pop-control-plane-design.md), approved SHA-256 `B0B0DDBBD11546FC07A25CB76992E375481192A3261270FF442095501CC2B4B5`; [approved authority Abort/serving amendment](../specs/2026-08-24-nodecontrol-authority-abort-serving-design.md), approved content SHA-256 `86996084462A5DE1E7667D56A135E93099EE38E1089464CCDFCFF7284EA0D1D7`; [approved authority v7 upgrade amendment](../specs/2026-08-24-nodecontrol-authority-v7-upgrade-design.md), approved content SHA-256 `EFAEBE52BDC3D70BDA8737C893B02752E60ACEC0A08441813CADF1425079FC8E`; [approved canonical authority evidence/dispatcher addendum](../specs/2026-08-28-nodecontrol-authority-canonical-dispatcher-addendum-design.md), approved content SHA-256 `7D480607A92214627C1CEF3E81AF76610EC861508CE8D510D5BD046E82600FE2`; canonical set manifest [c12-spec-set.v1.json](../specs/c12-spec-set.v1.json) lists the base followed by those three amendments in bytewise path order.

## Global Constraints

- 本计划只实现 C1.2。禁止加入 C1.3 账本、权益、配额租约，禁止加入 C1.4 调度、生产隧道凭据、客户端候选集或真实公网代理流量。
- v7 不支持 mixed-version rolling upgrade。任一 durable/shared 环境在 registered `00007` Up 前必须先验证 `LocalRuntimeIsolationV1`并排空旧 writer/listener/signer/provider-mutating job/session/route/credential/capability；普通 Goose CLI 不注册 00007。
- `00007` 只安装默认关闭的兼容 schema。只有 complete activation/completion/release、current runtime/incarnation result 链、active provider Head 与有效 serving lease 全部相等时才可开放 ordinary writer/reader/listener/signer。
- `fresh_v7_staging_closed` 是 restore-incomplete 安全停靠状态，不是 legacy upgrade 完成；B01 不声称完成 trust/authorizer rotation、legacy destructive restore 或 production provider runtime。
- 所有 v7 body/envelope/evidence bundle 使用固定 schema↔Go/SQL 映射、RFC 8785 JCS、domain-separated SHA-256 与封闭 signer-role/policy 验证；历史 `AuthorityEffectCommitmentV1`、`AuthorityEffectResolutionV1`、`ActivationDecisionEvidenceV1` 仍属 Abort/serving amendment，不得重复登记为 v7 新 schema。
- Provider/attestor 调用不得发生在 SQL transaction 或 DB lock 之内；所有跨边界恢复只使用预分配 immutable ID、canonical digest/preimage、fresh challenge/Inspect 与 exact retry。
- `node-agent` 和 `node-core-supervisor` 的 production support matrix 只含 `linux/amd64`；Windows 只运行控制面开发入口与 Job Object portability fixture。
- `control-api` 继续是 modular monolith，但 public、bootstrap、agent、operator、metrics 五个 listener 使用独立 server、端口、认证和 resource cap。
- PostgreSQL 是唯一数据库权威；Redis 只能保存可重建的短期状态；NATS JetStream 仍是至少一次投递并按 `event_id` 去重。
- 所有授权、撤销、identity epoch、trust/root/metadata、desired/recovery activation、operator transition 与 authorizer record 都经过 `ControlPlaneAuthorityFence`。Observation 与纯读不消耗 sequence。
- Abort 只允许 exact unbound/unclaimed fence 与同一 `READ COMMITTED` DBTX resolver 证明的 `EffectAbsent`；durable claim、所有 authority-bearing writer 的 fence-first lock/trigger、provider Abort 和 terminalization逐阶段分离。`EffectPrepared` 后才确定的 signer/issuer/provider failure、deadline 或 supersession 必须先用 B01 canonical helper 写 immutable `final_not_applied` commitment，使 resolver 返回 `EffectCommitted`，再走同一 `Coordinator.Finalize/Recover`；不得把已准备工作改判为 absent 或 Abort。
- Provider committed Receipt、stored Head/checkpoint、全部 contextual-validation preimage、evidence/resolution、DB fence terminalization、领域 disposition/active pointer与 required audit/outbox必须由 Coordinator 的 transaction-bound activator一次提交；rollback-resistant capture在首个 evidence-related external observation前开始，首次 consume 无论成功或失败都 burn，并只在所有写后紧邻唯一 `Commit`；none-time proof无 token且绝不 consume。Commit uncertainty先从同一 locked snapshot解析/验证 stored preimage和领域 terminal outcome，不用 mutable current Head替代。
- Production fence、certificate issuer、NodeStateSigner、root-share provider、OperatorAuthorizer、OperatorClientTrustGuard、RollbackGuard、SecurityLatchGuard、TrustedTimeSource 和 host evidence verifier 都必须是外部 provider；local/test provider 名称必须显式含 `Deterministic` 或 `LocalTest`，不能进入 production profile。
- C1.1 `internal/trust`、C1.1 root schema、config signer key 与 device enrollment token domain 不得复用为 C1.2 node-state、node certificate 或 node grant authority。
- C1.2 canonical JSON 使用现有严格 JSON/JCS 基础，但每个 schema 使用自己的固定 domain-separated transcript；V1 不协商算法。
- Node、operator、server leaf 固定 ECDSA P-256、精确 SAN/EKU/KeyUsage/BasicConstraints/extension profile；TLS 仅 1.3 + `h2`，禁止 tickets、0-RTT、renegotiation和 `InsecureSkipVerify`。
- Bootstrap listener 不请求客户端证书；agent/operator listener 必须 `RequireAndVerifyClientCert`，并在每个 request 入场和 response commit 前重新执行 exact certificate row 与状态授权。
- Node desired/recovery/root/metadata、trust bundle、resource envelope 与 host memory policy 都执行 authority epoch/sequence、同流 version、digest、cumulative revoke 与 same-value fork 检查。`node_authority_checkpoint` 在 node request 中只是 scalar sequence，不是 artifact-backed `VersionedDigest`；client-ahead 只走 headerless global authority-readiness/Head/Inspect recovery，绝不创建 per-node trust-conflict incident、notice 或 latch。
- Agent 的 TLS、snapshot validity 与 lease deadline 全部来自 production `TrustedTimeSource`；HTTP `Date`、signed time evidence、wall clock 和 RollbackGuard counter 都不能设置可信时间。
- Agent main RollbackGuard、supervisor RollbackGuard 与 LocalSecurityLatch 使用三个独立 counter/keystore identity；任何文件、MAC、数据库 sequence 或 wall clock 都不能冒充 production anti-rollback。
- Host-deployed `DeploymentAuthorityKeySetV1` 至少四把不同 Ed25519 key，角色精确为 `trust_bundle`、`host_remediation`、`operator_trust_guard`、`node_resource_envelope`，一把 key 不能占两个角色。
- `ApprovedReleaseManifestV1` / `InstalledReleaseMapV1` canonical types、strict verifiers、domain digests和immutable verified handles由 Plan 02 Task 5唯一拥有；manifest bytes随 agent/supervisor build固定，Plan 07只从 reviewed locks原子重生成其固定embedded instance。Agent与supervisor各自验证embedded bytes，并各自no-follow读取固定root-owned map及每个immutable release root后才派生/提交`LocalVersionedDigestV1`；caller tuple/path/bytes/boolean不可替代。Map只把release ID映射到不可写absolute release root，不能覆盖digest、argv、adapter、dependency closure或license。
- Trust-conflict禁用/隔离后的exact incident-bound current certificate只保留`trust_conflict_evidence|recovery_poll|recovery_attestation`三类端点；重启notice/latch模式必须同时运行bounded evidence retry和专用recovery poll/attestation worker，desired/report/ordinary reconcile/LKG仍关闭，避免signed clear永不可达。
- Desired state 最多 8 个 slot、canonical payload 最大 64 KiB；API 普通 body 最大 64 KiB，明确的 trust-conflict artifact endpoint 才允许 1 MiB/两个 artifact。
- Agent 永不接收 path、argv、environment、shell、原始 core config、production credential、下载 URL 或动态插件。Supervisor 本地协议也不接受这些字段。
- Xray 与 sing-box 始终是独立 OS 进程、独立上游制品和独立发布批次；不得 import、vendor、link 或复制进三个专有 release binary/production image。
- C1.2 core profile 仅是固定 loopback test profile；真实 ingress/egress/TUN/生产 credential 属于后续已批准规格。
- P07 core-smoke runner 与 P08 native verifier parent 的跨进程结果只经一次性 Linux protected result slot：父进程创建 64 KiB sealed `memfd` 并仅以 direct-child inherited FD 3 交付 writer；child 在 11 项 exact cleanup 与 WAL terminal 后调用同一 `RunnerReceiptFinalizer` 写入，父进程 reap 后一次 adopt、mint opaque `TrustedCoreSmokeReceiptSource` 并再次调用唯一 semantic validator。bytes、path、digest、FD scalar、reopened FD 或 shell 中转都不能构造可信源。
- `C12NativeRunnerProfileV1` 固定为四个 final roles `windows_scopes|platform|authority|completion` 加一个零发布能力的 `staged_diagnostic`；tracked profile 只保存 role/purpose/channel policy，absolute path、Git object-database identity、installed native executable identity、actual slot/sequence、provider/attestor capability 只存在固定 role-indexed machine store/state cells 与 `C12NativeRunnerInstallReceiptV1`。receipt 只覆盖 native helper/interpreter/tool/private child，snapshot scripts 只从 authenticated committed snapshot 派生且没有 receipt。Windows/Linux 各有一个由 OS image/deployment baseline 安装并以 `C12FixedRunnerLauncherAttestationV1` 固定的 machine-base launcher，位置只能是 `C:\Program Files\Talenro\C12\talenro-c12-runner-launcher.exe` 或 `/usr/libexec/talenro-c12/talenro-c12-runner-launcher`；launcher不是role receipt或versioned helper，catalog精确覆盖两种launcher binary/SBOM/provenance。每个 role 独立保持 install generation、monotonic launch generation 与至多一个 active run，并由role-unique不可重置rollback-resistant `(StateEpoch,current state digest)` guard选择固定A/B cell中的唯一当前逻辑状态及其immutable install-record generation；bootstrap只保存稳定guard/cell/install-root identity而无active pointer。guard不含phase/run/launch/next/publication/payload且不能由文件/DPAPI/MAC/其他role counter替代。每次run由launcher先把suspended/stopped helper原子放入kill-on-close Job或受监管cgroup，guard-select带boot/PID/start/containment的execution reservation后才resume；launcher死亡、重启、PID复用或cross-boot recovery必须先证明原containment为空。active phase只能替换reservation后走同generation pathless recovery；pre-Open/terminal-inactive reservation只能清除，fresh work还须新reservation。Completion create 在 `retained_complete` 上不能伪装 recovery：旧 `exiting` reservation 必须先 absence-proof+clear，再由独立不消费generation的 `terminal_audit` reservation调用只读终态审计。Revalidate 必须先由 guarded dispatch 区分 `outcome_delivery_pending` 与 fresh work；pending 只能经 `revalidation_outcome_audit` reservation、受保护result-slot文件+目录fsync回执和assignment-identical ACK，任何dispatch错误不得回退尝试Open。只有ACK+reservation clear后才可fresh Open，且新trusted-time head必须严格晚于creation baseline与所有已完成/已中止但已接受时间的最新head。diagnostic 只有 ODB/external-root/recovery/Linux-helper/Git/build-tool handle，completion 永远只接收三个 final sender transaction。
- Protected WAL-key bootstrap 固定为 intent-first：在任何 durable backend key create 前，rollback-resistant guard先选择绑定run/role/generation/recovery-slot与deterministic backend reservation的 `protected_key_intent`；leaf只可对该binding执行exact-idempotent CreateOrRecover。cleanup capsule durable后，唯一guard CAS原子消费intent并选择 `cleanup_only`。每次重启先三路 exact-dispatch：active generation 尚无 selected intent/capsule/WAL/resource 时只走零创建 `abort_unstarted` 并终结；selected intent 时只走 pathless `abort_intent`，exact删除候选、销毁/验证该一把key并回到origin-specific固定abort target；仅两种abort条件都不存在的 authenticated selected role state 可走 `role_dispatch`。任何错误不能枚举、创建第二把key或回退到role/workload route。Completion inner cleanup由已有coordinator从已消费session的sealed投影内部执行同一协议：create恢复原`publish_pending`，revalidate恢复原`retained_complete_active`。
- 每个非 completion native parent 在第一个 **ordinary** resource/provider/producer intent 前，由 unexported c12evidence role facade 提供 authenticated session 的不透明句柄与内部稳定 resource reservations；固定 role builder 只调用 `cleanup.go` 的高层 `BindCleanupOnlyOwnershipPlan`，由它独占 raw runnerprofile `BindFixedCleanupOnlyOwnershipPlan`，内部绑定完整 sealed plan。高层随后执行 `BeginProtectedWALKeyBootstrap → CreateOrRecoverProtectedWALKeyBootstrap → SealAndSelectProtectedWALKeyBootstrap → OpenSelectedOwnershipWAL`；runnerprofile 执行 guarded raw key Open/Create/Seal 并返回唯一 opaque fixed-selected token，最后一步把同一 token 一次交给 guarded `OpenFixedSelectedOwnershipWAL`，current-state reread 后才可 raw exact-create/reopen WAL，并在返回前 durable flush bootstrap、observed identity 与 file+directory。command 与 c12evidence 都拿不到 leaf binding、intent、key handle、path或signer。`internal/c12cleanup` 独占 resource types、binding/key leaf、16-KiB machine-sealed selected-WAL binding envelope、sealed WAL recorder、cleanup capability/proof/receipt与Recover/Commit；`ResourceRef`和每条`WALRecordV1`都以nonzero declaration digest唯一选择capsule declaration，再冗余核对type/name，same type+name/different parent永不别名；c12evidence只持非completion plan/bootstrap/fixed-selected 的 opaque runnerprofile wrappers与shared dispatch，不能Marshal/Recover leaf binding。所有普通/abort/success cleanup 都在guard-selected pending内按 receipt commit → key absent → tombstone unlink+parent-fsync+exact absence收敛，最后CAS才选择terminal或same-generation post-clean target；target后无filesystem work。Windows/authority 的正向入口先由 artifactscan 唯一 typed validator 通过一次性 sink 密封 receipt+bundle，再与各自 private unsigned semantic token 及同一 session 绑定；command 不能传 bytes/DTO/callback/permit。Windows/platform/authority 的success-origin CAS还把result/receipt/bundle（platform另含manifest）的有界canonical bytes按固定顺序直接写入guard-selected A/B cell，并只用 `FixedPostCleanupPublicationInputsDigest` 的 `uint8(role length)||role||uint8(member count)||Σ(uint64be(member length)||member)` unsigned big-endian framing计算`PublicationInputsDigest`；cleanup pending/post-clean target逐字承接到final generation CAS，fresh parent只能从current post-clean view defensive copy、调用同一 digest 函数并由wrapper-only facade重建/签名，caller无evidence struct/bytes。Success successor全部nested/outer/channel提交后才advance next，abort、active crash或cleanup failure无签名器。
- Completion launch、outer `publish_pending`、adoption/publication/ACK、`finalize_ready` 与临时 inner authority 共用一个 guard-selected `C12CompletionRoleStateV1`；inner union严格为 `none|protected_key_intent|inner_bootstrap_restart_pending|cleanup_capsule|finalization_pending`。Coordinator在 capsule selection 后独占Marshal/Recover leaf selected-WAL binding，并在 Begin、validation recovery、cleanup recovery返回handle/session前 exact-open/reopen同一WAL、file+parent-fsync bootstrap/actual；handle私持唯一recorder，P08 runtime在每个provider/process/temp create前后记录durable intent/observed actual，cleanup-only禁用新append与Validate/Stage。full `StateDigest`覆盖 execution/inner，outer-only P排除二者。P08独占 I/T 与严格 Sp/Sa/Fp/Fa：ordinary durable `Sp → outer Pnew → Sa`，hash边 `T→{Sp,Pnew}`、`{Sp,T,Pnew,new epoch/full,result}→Sa`；finalize专走 `Fp → outer E1/D1/P1 → source-only Fa`，launcher replacement只可将proof rebase到相同I/T/delta/P1、inner none的current E'/D'/P1。P/outer/view不含set-record digest；V只绑定materialization-intent前驱和exact-eight结果。五路router仅允许 genuine none→Begin、pre-manifest selected→validation session、set-first/manifest intent capsule→state-barrier后cleanup session、pending→no-session finalization、target-completed+manifest-intent-or-later→actionless complete。Manifest intent在任何temp write前持久化两份同一≤64-KiB exact bytes并保留至rename+parent-fsync后的reciprocal durable；恢复不重跑provider/child/validator/canonicalizer。只有c12completionpublication消费低层publication能力，command只持opaque facade。
- 三个ACK actual后冻结 exact E2 target-transition digest 与独立 immutable retained-payload/finalization-set audit digest。Prepare只完成E→E1，proof可在合法execution-reservation replacement后从current selected E'/D'/P1重建；Commit才构造/fsync E2、advance guard并标记reservation `exiting`。E2后只读audit验证immutable digest与closed completed/aborted lineage。每个inner terminal route消费assignment-identical leaf receipt；首个CAS选择fixed `finalization_pending`，no-session recovery在该pending内完成leaf Commit、key absence、tombstone unlink+parent-fsync+absence，最后CAS才安装create cleanup target或retained `outcome_delivery_pending`。CAS后无retirement；response loss仅经audit、protected delivery与receipt-bound ACK。sender只可Prepare/Commit/Inspect，不得枚举/猜nonce/按path重开。
- B10 先提交最终 Task 7 source/test/script/owner-map commit，再从其 clean tree 执行唯一原子 relock，恰好重写 Compose、integration manifest、toolchain lock、Windows profile、diagnostic profile 五个 tracked outputs并单独提交projection commit；host gates、provisioning与exit只运行在该clean projection tree。B11 同样先证明 source candidate union 恰好12项并提交、三态clean，再从该 clean source commit 原子重写且只重写 Compose、integration、toolchain、trusted-time provider 与五份 runner profile共九个 outputs；12-source与9-output集合交集恰好只有digest-free completion profile，source commit前其余8 outputs必须不变。B11 projection commit还要证明working/cached集合恰好9项、worktree=index、candidate tree等于commit tree、父提交恰为source commit及提交后三态clean；provisioning后再次匹配保留的projection commit/tree并三态clean，才可进入Task 8。两种machine-base launcher及其SBOM/provenance是byte-identical frozen inputs，不增加第六/第十个输出或role receipt。任一 replacement 失败保持整组 byte-identical；随后每个role host必须在匹配的authenticated OS deployment transaction内，以不可导出且role/host/projection-bound candidate-package handle调用固定launcher provision mode，再用non-consuming `VerifyFixedNativeRunnerInstallation`审计。authenticated selected install record密封且只允许`ProjectionVariant=b10_pre_catalog|b11_final_catalog`：B10 variant要求catalog absent/zero并只绑定toolchain native-role/profile/launcher attestation，B11 variant要求final catalog exact cover并另绑定final relock commit；Verify没有caller variant参数，只按该selected record分派。`InstallFixedNativeRunnerProfile`线性消费session并在每个返回前销毁session+candidate handle。first install在guard advance前保持`(0,zero)` uninstalled/retryable；reinstall advance前仍只选旧generation。只有实际零参数public parent在fresh inactive generation且已有guard-selected execution reservation时可调用一次 `OpenInstalledNativeRunnerSession`；`active_unstarted|publish_pending|retained_complete_active`或platform resume只能在旧containment absence proof后走fixed pathless recovery，安装/预检不得Open，任何generation都不得第二次Open。
- Supervisor 在 child create 前验证 resource envelope、aggregate reservation、manifest/map、memory policy、exact executable FD 与全部 finite cgroup/rlimit；任何 read-back 差异 fail closed。
- Production Linux V1 必须 SELinux enforcing，并要求 Yama 3、unprivileged BPF disabled、空 core pattern、suid dumpable 0、core_uses_pid 0 与 collector mask。Docker fake 不得声称证明 host policy。
- 所有 stdout/stderr、HTTP 错误、日志、metrics、outbox、observation 与 evidence 使用有限 enum/allowlist，禁止 grant、CSR、cert DER/serial、key、signed payload、core output、credential、path、IP 原文和 node ID metric label。
- 所有生成物提交仓库；`scripts/generate.*` 后 `api/`、`gen/`、`internal/store/` 必须无 drift。`internal/store/models.go` 与 `internal/store/querier.go` 是按批次顺序串行生成的 shared sqlc outputs，不属于任一批次可手写的独占 API；每个修改 query 的 task 只可由 pinned generator 改写它们，并以 first-generation exact stage → second generation → worktree-vs-index empty diff 证明没有跨 task 漂移。
- 每个实施任务遵循 RED → GREEN → REFACTOR：先运行 focused failing test，再做最小实现，再运行 focused、package 和受影响 integration test，最后独立提交。
- 普通 Go build/test 固定 `CGO_ENABLED=0`，race 固定 `CGO_ENABLED=1`；三个 release 固定 `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 -buildmode=exe`，禁止 external linker、`-linkshared` 和自定义 extldflags。
- `cmd/nodecontrol-authority-operations` 是安全关键的独立辅助制品而不是第四个专有 release；最终四个 scope 都必须另外绑定同一 exact binary/image/build/dependency tuple、组合摘要、subject-checked supply-chain record-set及canonical bundle、唯一 artifact catalog、process-image policy、full image set 与 `ArtifactScanDigest`。scanner 对其静态 binary、专用 OCI image及依赖闭集执行与production集合相同的禁止-core/secret/license检查。
- PowerShell 与 repository-external Git-for-Windows Bash wrapper 必须调用同一个 canonical Docker verifier；Linux platform/provider 与 authority-fence 是另外两个作用域。
- 每个 authority evidence 最长 72 小时，绑定 repo commit、tracked-tree、canonical `C12SpecSetV1` digest、toolchain、三个 release binary、全部 scanner-derived operations/bundle/catalog/process-policy/full-image/full-scan facts 与 cleanup digest。Spec set只从 tracked `docs/superpowers/specs/c12-spec-set.v1.json` 构造；artifact set只从 tracked `testdata/c12/artifact-catalog.v1.json` exact-cover；最终 manifest 要求四个作用域恰好各一且这些输入逐字相同。
- Test harness 只清理本 run 的 exact deterministic name/ID 与已验证临时根，不枚举或删除既有用户资源；主失败与 cleanup 失败分别保留 exit bit。
- 用户现有未跟踪目录 `.cache/`、`.superpowers/`、`.task19-go/` 不属于本计划，不得 stage、删除、扫描凭据或用作 test root。
- Redis 8.8.1 production license block、sing-box GPLv3+ gate 与 Xray MPL-2.0 gate 保持有效；进程隔离不是法律结论。

---

## Scope check and plan suite

设计第 21 节包含数据库权威、身份信任、节点 TCB、主机隔离、两个独立 core adapter 和四作用域验收，不能作为一个可审查提交序列实现。本索引冻结跨分册类型、文件责任与依赖；九个分册保留全部十一个实施批次：

| Batch | Plan | Independent exit |
| --- | --- | --- |
| `C1.2-B01` | [01 contracts, schema, authority](2026-08-23-node-pop-control-plane-01-contracts-schema-authority.md) | 三份 OpenAPI、nodecontrol events、Abort/serving 原语，v7 全部 canonical contracts/verifiers/golden，base 25 + additive 26 的 51-table catalog、registered Go 00007/Down guard、exact-epoch repository/readiness、`BoundAuthorityReadSource` 与 catalog/crash/PITR matrix 通过 |
| `C1.2-B02` | [02 operator state and signing](2026-08-23-node-pop-control-plane-02-operator-state-signing.md) | inventory/cursor/audit、desired/recovery signer saga、root threshold publisher，以及 Task 10 仅经 B01 opaque admission/repository 的 disabled-only fresh-restore consumer 通过 |
| `C1.2-B03` | [03 node identity and mTLS](2026-08-23-node-pop-control-plane-03-node-identity-mtls.md) | enrollment/rotation/receipt、bounded recovery/remediation/two-person restore、trust package/operator guard/三个 TLS listener，以及 production claim-v1 provider、incarnation/runtime/timeline attestor、trusted WAL commit archive/Fence/Inspect/first-consumer adapter 通过 |
| `C1.2-B04` | [04 node agent](2026-08-23-node-pop-control-plane-04-node-agent.md) | agent guards、signed pull/recovery、resource envelope 与 single-writer reconcile 通过 |
| `C1.2-B05` | [05 node-core supervisor](2026-08-23-node-pop-control-plane-05-node-core-supervisor.md) | strict socket、lease/seal、双 latch、Linux sandbox/aggregate reservation 通过 |
| `C1.2-B06–B07` | [06 fixture, health, load](2026-08-23-node-pop-control-plane-06-fixture-health-load.md) | controlled process、capacity reducer、privacy 与 1,000-agent reference load 通过 |
| `C1.2-B08–B09` | [07 Xray and sing-box adapters](2026-08-23-node-pop-control-plane-07-xray-singbox-adapters.md) | 两个独立 fixed-profile real loopback smoke 依次通过 |
| `C1.2-B10` | [08 verifier and supply chain](2026-08-23-node-pop-control-plane-08-verifier-supply-chain.md) | 唯一 `C12SpecSetV1`/`SpecDigest` builder、tracked artifact-catalog schema/exact-cover scanner、strict build-context/operations-policy loader、最小 Git object database、canonical supply-chain bundle、双 wrapper contract/fixture、nested Docker、cleanup capsule/WAL/tombstone、attested channel 与 runner profile/bootstrap/install-receipt schema通过；B10 原子 relock 五个 outputs 并安装 Windows+diagnostic roles，B11 final catalog/artifact落地前不生成可复用权威scope或upgrade/staging/recovery evidence |
| `C1.2-B11` | [09 operations and completion](2026-08-23-node-pop-control-plane-09-operations-completion.md) | 唯一 production upgrade/restore/Down orchestrator、registered migration/shutdown runbook、真实 Linux/operator guard、provider/DB/PITR/destructive-restore conformance通过；最终原子 relock 九个 outputs、安装全部五类 runner，completion 只 adopt 三个 final transactions，并由 Plan 09 Task 8 运行四作用域 manifest |

严格按表顺序执行。合并分册中的 `B06/B07` 与 `B08/B09` 仍各自拥有独立 RED/GREEN/提交和 exit gate；不能把 fixture 通过等同 health/load 通过，也不能让一个 core adapter 冒充另一个。

## Repository map and ownership

```text
api/
├── openapi/
│   ├── control-api.v1.yaml                       # 现有 public/C1.1 API，不承载 node route
│   ├── node-bootstrap-api.v1.yaml                # server TLS + grant claim
│   ├── node-agent-api.v1.yaml                    # node mTLS poll/report/recovery
│   ├── node-operator-api.v1.yaml                 # operator mTLS inventory/actions/audit
│   └── node-*-oapi-codegen.yaml                  # 三个独立生成配置
└── proto/talenro/
    ├── nodecontrol/v1/events.proto               # nodecontrol outbox contract
    └── nodesupervisor/v1/protocol.proto          # 本地 message schema，不生成 gRPC
cmd/
├── control-api/                                  # public + bootstrap + agent + operator + metrics
├── node-agent/                                   # linux/amd64 outbound agent
├── node-core-supervisor/                         # root-owned local TCB
├── nodecontrol-authority-operations/             # sole production v7 upgrade/restore/Down entrypoint
├── c12-fixture/                                  # 独立 controlled process
├── c12-load/                                     # 1,000-agent load generator
├── c12-load-runner/                              # B07 run-owned authoritative Linux substrate
├── talenro-core-lock/                            # B08/B09 offline release-lock/dependency tool
└── talenro-artifact-scan/                        # versioned binary/OCI scanner
db/
├── migrations/00006_nodecontrol.sql              # base 25 tables；v7 只补每个 PL/pgSQL block 的 Goose annotations
├── migrations/00007_nodecontrol_authority_abort_serving.go  # authoritative executor 才注册的 Go migration
├── migrations/assets/nodecontrol_authority_v7_{up,down}.sql # embedded literal assets; 26 additive protocol tables
├── schema/nodecontrol.v1.yaml                    # final 51-table exact catalog: base 25 + v7 26
└── queries/nodecontrol_{authority,serving,inventory,identity,state,recovery,resource_envelope,observation}.sql
internal/
├── nodecontrol/
│   ├── contracts/                                # pure canonical DTO/enums plus every v7 body/envelope/evidence/golden mapping
│   ├── authority/                                # fence/claim, v7 verifier/repository/readiness; production provider runtime remains B03
│   ├── serving/                                  # bound same-connection certificate/desired guarded reads
│   ├── inventory/                                # POP/node/slot/capacity/resource envelope
│   ├── operator/                                 # exact credential authorization/cursor/audit
│   ├── state/                                    # root/metadata/desired/recovery/signer workflows
│   ├── identity/                                 # grant/x509/issuance/receipt/certificate auth
│   ├── hostevidence/                             # trust package/guard/remediation verification + fenced resource-envelope publication
│   ├── recovery/                                 # incident/session/clear/resume/restore state machines
│   ├── claimv1/                                  # production provider Head/epoch/staging/source state
│   ├── incarnation/                              # runtime/timeline/rebind production attestors
│   ├── commitarchive/                            # trusted WAL archive, journal, Fence and Inspect
│   ├── authorityadapter/                         # B01 contracts to B03 provider adapters
│   ├── operations/                               # B11 sole production v7 orchestration
│   └── observation/                              # report/reducer/health transitions
├── nodebootstrapapi/                             # generated bootstrap adapter
├── nodeagentapi/                                 # generated agent adapter and waiter cancelation
├── nodeoperatorapi/                              # generated operator adapter
├── localrelease/                                 # B02 build-embedded approved manifest byte source；P07 later regenerates fixed instance
├── nodeagent/
│   ├── localstate/                               # RollbackGuard/latch/double-slot/time
│   ├── transport/                                # trust-guarded TLS claim/rotate/poll/report
│   ├── releasecatalog/                           # independent B02 manifest/map/root verifier and immutable handle
│   ├── trust/                                    # independent C1.2 signed-package verifier
│   ├── reconcile/                                # single writer and LKG transition
│   └── adapter/{fixture,xray,singbox}/            # typed previews and probes
├── nodesupervisor/
│   ├── wire/                                     # strict canonical Protobuf framing + SCM_RIGHTS
│   ├── state/                                    # independent guard, lease and seals
│   ├── releasecatalog/                           # separate B02 manifest/map/root verifier and immutable handle
│   ├── sandbox/                                  # Linux execution/isolation/read-back
│   └── adapter/                                  # deterministic config compiler
├── c12evidence/                                  # strict spec-set builder, WAL, evidence schema and completion validator
└── artifactscan/                                 # Go deps, ELF/PE and OCI layout scanner
deploy/c12/                                       # audited outer Compose/images/policies
scripts/
├── run-c12-integration.ps1                       # B01-owned ephemeral PostgreSQL/Redis/NATS dependency set + exact tagged Go-test runner
├── verify-c12.ps1                                # fixed PowerShell wrapper
├── verify-c12.sh                                 # repository-external Git Bash wrapper
├── c12-authority.sh                              # canonical inner Bash verifier
├── verify-c12-platform.sh                        # Linux-native attested host gate
├── verify-c12-authority-v7-operations.sh         # real v7 genesis/epoch/staging/source/Down matrix
└── verify-c12-authority-fence.sh                 # provider/PITR gate
testdata/c12/
├── artifact-catalog.v1.json                     # sole final tracked exact-role catalog
├── integration-contracts-schema-authority.v1.json # B01 exact integration groups/profiles
├── integration-operator-state-signing.v1.json   # B02 exact integration groups/profiles
├── integration-node-identity-mtls.v1.json       # B03 exact integration groups/profiles
├── authority/operations-build-policy.v1.json    # digest-free deterministic auxiliary recipe/policy
└── ...                                          # locks, manifests, vectors and negative scanner fixtures
docs/licenses/c12-authority-operations.md         # auxiliary license/notice/source obligation
docs/security/c12-threat-model.md
docs/runbooks/c12-*.md
```

Plans 04 and 05 additionally fix `internal/nodeagent/releasecatalog/catalog.go` and `internal/nodesupervisor/releasecatalog/catalog.go` plus their tests as two independent consumers of B02's build-embedded manifest and installed map/root facts；`cmd/node-agent/trust_conflict_recovery_e2e_test.go` is an untagged startup/restart E2E for the retained-notice `409 → restart → signed clear → attestation ACK` path. These files introduce no second manifest type、caller tuple or test-only recovery route.

Plan 01 在上述目录内进一步固定 `internal/nodecontrol/contracts/{canonical,signature_envelope,evidence_bundle,authority_v7_genesis,authority_v7_epoch,authority_v7_staging,authority_v7_source,authority_v7_down,commit_archive}.go`、`internal/nodecontrol/authority/{v7_verifier,database_head,exact_epoch_readiness}.go` 与 `testdata/c12/authority-v7/*.json`。这些文件只提供 pure canonical body/envelope/evidence、typed verifier、repository input 与 golden vectors；production provider、timeline/runtime attestor、trusted WAL decoder 与 rollback-resistant archive 的可变实现在 Plan 03，不在这些文件中。

## Frozen cross-plan interfaces

The pure shared authority primitives are created in `internal/nodecontrol/contracts/authority.go`:

```go
package contracts

type Digest [32]byte

type AuthorityVersion struct {
	Epoch    uint64
	Sequence uint64
}

type VersionedDigest struct {
	Version           uint64
	AuthoritySequence uint64
	Digest            Digest
}

type LocalVersionedDigestV1 struct {
	Version uint64
	Digest  Digest
}

func (v AuthorityVersion) Validate() error
func CompareVersionedDigest(current, candidate VersionedDigest) (Comparison, error)
func CompareLocalVersionedDigest(current, candidate LocalVersionedDigestV1) (Comparison, error)
```

`Comparison` is a closed enum with only `same`, `advance`, `rollback`, and `fork`; callers must switch exhaustively.

The Abort/serving-amendment external authority boundary is created in `internal/nodecontrol/authority/provider.go`:

```go
package authority

type Provider interface {
	Reserve(context.Context, ReserveRequest) (Reservation, error)
	Finalize(context.Context, FinalizeRequest) (Receipt, error)
	Abort(context.Context, AbortRequest) (Receipt, error)
	Inspect(context.Context, uuid.UUID) (Record, error)
	Head(context.Context) (Head, error)
	CommittedNodeCheckpoint(context.Context, contracts.Digest) (NodeCheckpoint, error)
}

type ReserveRequest struct {
	OperationID uuid.UUID
	Kind        EffectKind
	ScopeKind   ScopeKind
	ScopeDigest contracts.Digest
}

type FinalizeRequest struct {
	OperationID  uuid.UUID
	EffectDigest contracts.Digest
	DBSystemID   uint64
	DBTimeline   uint32
	RequiredLSN  WALPosition
}
```

`ScopeKind` is closed to `node`, `global_node_trust`, and `global_operator_trust`. `Reserve` is idempotent by operation ID; any request-field change is `ErrConflict`. `Finalize` and `Abort` are terminal and mutually exclusive. `CommittedNodeCheckpoint` includes only the exact node scope plus committed global-node-trust effects and returns the exact latest receipt digest. No provider request contains grant, certificate, desired payload, key or credential bytes.

`Provider` above remains the approved legacy_v6/ordinary Reserve/Finalize/Abort surface and is not widened into a caller-selectable v7 mega-interface. Plan 01 separately owns exact canonical v7 request/response/envelope types plus pure verifiers for genesis, epoch, staging, source retirement and Down. Plan 03 must expose `authorityadapter.OrdinaryProvider` as the narrow production view over `claimv1.ClaimV1Provider`, prove it implements this exact six-method `authority.Provider`, and pass that one view to the sole production Coordinator; its v7 provider/attestor interfaces remain distinct typed adapters and are never reachable through the ordinary view. Plan 09 owns sequencing and may call the separate v7 interfaces only outside DB transactions. Plan 08 alone computes `SpecDigest` from `c12-spec-set.v1.json` and must not mint any of those protocol evidences.

All domain mutations consume the Batch 01 coordinator rather than constructing provider database coordinates. Read-only readiness may keep a non-locking resolver, but Abort and final visibility use the exact transaction-bound dispatcher:

```go
type CoordinatorFinalizeRequest struct {
	OperationID  uuid.UUID
	EffectDigest contracts.Digest
}

type TransactionalEffectQuery struct {
	registeredKind EffectKind
	expected       Reservation
}

func (q TransactionalEffectQuery) RegisteredKind() EffectKind
func (q TransactionalEffectQuery) Expected() Reservation

type TransactionalResolvedEffect struct {
	OperationID uuid.UUID
	Epoch       uint64
	Sequence    uint64
	Effect      ResolvedEffect
}

type RegisteredEffectResolver interface {
	ResolveRegisteredAuthorityEffectForUpdate(context.Context, store.DBTX, TransactionalEffectQuery) (TransactionalResolvedEffect, error)
}

type TransactionalEffectResolver interface {
	ResolveAuthorityEffectForUpdate(context.Context, store.DBTX, Reservation) (ResolvedEffect, error)
}

type RegisteredEffectActivator interface {
	CaptureActivationDecisionMaterial(context.Context, Receipt) (ActivationDecisionMaterial, error)
	ActivateAuthorityEffect(context.Context, store.DBTX, Receipt, ValidatedActivationDecisionEvidence) error
	ValidatePersistedAuthorityEffect(context.Context, store.DBTX, Receipt, ValidatedActivationDecisionEvidence, AuthorityEffectResolution) error
}

type TransactionalEffectActivator interface {
	RegisteredEffectActivator
}

type EffectDispatcher interface {
	TransactionalEffectResolver
	TransactionalEffectActivator
	authorityEffectDispatcher()
}

type EffectRegistration struct {
	Kind      EffectKind
	Resolver  RegisteredEffectResolver
	Activator RegisteredEffectActivator
}

func NewEffectDispatcher([]EffectRegistration) (EffectDispatcher, error)
func NewCoordinator(Provider, Repository, EffectDispatcher) (*Coordinator, error)
func (c *Coordinator) Reserve(context.Context, ReserveRequest) (Reservation, error)
func (c *Coordinator) Finalize(context.Context, CoordinatorFinalizeRequest) (Receipt, error)
func (c *Coordinator) Abort(context.Context, AbortRequest) (Receipt, error)
func (c *Coordinator) Recover(context.Context, uuid.UUID) (Receipt, error)
func (c *Coordinator) CheckReady(context.Context) (Readiness, error)
```

Identity、state/root publishing、recovery and resource-envelope repositories each expose one `RegisteredEffectResolver` plus the three-method `RegisteredEffectActivator` for a disjoint closed kind set. `NewEffectDispatcher` alone constructs opaque per-registration queries, probes all 13 supported handlers in bytewise effect-kind order, validates their operation/epoch/sequence echoes, and keeps `operator_authorizer_change`/`trust_bundle_publish` fixed unsupported. `NewCoordinator` accepts only the exact non-nil private concrete dynamic type returned by `NewEffectDispatcher`; wrapper、embedded、proxy or alternate dispatcher types fail before any method call. Only the coordinator owns raw `Provider`, obtains the locked `Reservation`, captures Receipt/Head/checkpoint preimages after every earlier lock is released, constructs and context-validates the branded proof, atomically commits fence plus domain resolution/proof preimages, or exposes a same-repository bound read source.

| Registration owner | Exact effect kinds |
| --- | --- |
| Identity handler | `grant_create`, `grant_claim`, `certificate_activate`, `certificate_revoke`, `identity_epoch_advance` |
| State/root handler | `root_publish`, `metadata_publish`, `desired_activate`, `recovery_activate` |
| Recovery handler | `security_incident_open`, `security_incident_resolve`, `operator_transition` |
| Resource-envelope handler | `resource_envelope_activate` |
| Fixed unsupported registrations (nil resolver + nil activator) | `trust_bundle_publish`, `operator_authorizer_change` |

Canonical effect helpers are owned only by `internal/nodecontrol/authority`:

```go
func NewAuthorityEffectCommitment(AuthorityEffectCommitmentInput) (AuthorityEffectCommitment, error)
func ParseAuthorityEffectCommitment([]byte) (AuthorityEffectCommitment, error)
func NewAuthorityProviderHeadSnapshot(Head) (AuthorityProviderHeadSnapshot, error)
func ParseAuthorityProviderHeadSnapshot([]byte) (AuthorityProviderHeadSnapshot, error)
func AuthorityProviderHeadDigest(Head) (contracts.Digest, error)
func NewAuthorityCheckpointAnchor(AuthorityCheckpointAnchorInput) (AuthorityCheckpointAnchor, error)
func ParseAuthorityCheckpointAnchor([]byte) (AuthorityCheckpointAnchor, error)
func NewAuthorityEffectResolution(AuthorityEffectResolutionInput) (AuthorityEffectResolution, error)
func ParseAuthorityEffectResolution([]byte) (AuthorityEffectResolution, error)
func ValidateAuthorityEffectResolution(AuthorityEffectResolution, AuthorityEffectCommitment, ValidatedActivationDecisionEvidence) error
func NewActivationDecisionEvidence(ActivationDecisionEvidenceInput) (ActivationDecisionEvidence, error)
func ParseActivationDecisionEvidence([]byte) (ActivationDecisionEvidence, error)
func ValidateActivationDecisionEvidence(ActivationDecisionEvidence, ActivationDecisionEvidenceInput) (ValidatedActivationDecisionEvidence, error)
func BeginActivationEvidenceCapture() *ActivationEvidenceCapture
func (c *ActivationEvidenceCapture) Complete(ActivationDecisionEvidenceInput) (ActivationDecisionEvidence, error)
func (v ValidatedActivationDecisionEvidence) Evidence() ActivationDecisionEvidence
func (v ValidatedActivationDecisionEvidence) Input() ActivationDecisionEvidenceInput
```

B02/B03 authenticate finite domain/trusted-time inputs and return `ActivationDecisionMaterial`, but never receive raw Provider or reimplement JCS/digest transcripts. Coordinator supplies the exact committed Receipt、same-Provider `AuthorityProviderHeadSnapshot` and optional typed checkpoint anchor, then obtains `ValidatedActivationDecisionEvidence`. A fresh rollback-resistant proof shares one non-serializable operation/commitment/evidence-bound admission state across copies；its first consume attempt burns the state even on cancellation、expiry or binding failure. Parsed and none-time proofs never acquire a token, and only parsed-origin proofs may enter zero-write persisted-outcome validation.

Production certificate/desired reads consume only the bound capability below; an arbitrary pool plus independent availability check is not a valid constructor input:

```go
package serving

type BoundAuthorityReadSource interface {
	WithConsistentReadyRead(context.Context, func(context.Context, store.DBTX) error) error
}

func NewReader(BoundAuthorityReadSource) (*Reader, error)
```

The source uses one physical PostgreSQL connection for DB readiness pre-check, exact query and DB readiness post-check; provider calls occur outside SQL statements/locks. Results and domain errors remain private until exact head equality passes.

The state-signing boundaries are created in `internal/nodecontrol/state/provider.go`:

```go
package state

type NodeStateSigner interface {
	Sign(context.Context, SignRequest) (SignResult, error)
}

type RootShareProvider interface {
	SignShare(context.Context, RootShareRequest) (RootShareResult, error)
}

type SignRequest struct {
	SigningID    uuid.UUID
	Kind         SigningKind
	KeyID        string
	PayloadDigest contracts.Digest
	Transcript   []byte
}
```

`SigningKind` contains only `desired`, `recovery`, and `time_attestation`; root/metadata signatures cannot be requested through `NodeStateSigner`.

The certificate and operator boundaries are created in `internal/nodecontrol/identity/provider.go` and `internal/nodecontrol/operator/authorizer.go`:

```go
package identity

type NodeCertificateIssuer interface {
	Issue(context.Context, IssueRequest) (IssueResult, error)
}
```

```go
package operator

type OperatorAuthorizer interface {
	Authorize(context.Context, Credential, Action, Target) (Authorization, error)
}
```

Issuer requests contain an immutable issuance ID, exact issuer ID, public key, server-constructed leaf template and template digest. Authorization credentials contain issuer ID, serial bytes, exact leaf DER/public-key digests, URI SAN and authority epoch; security-admin actions are never served from cache.

The host-remediation boundary is created in `internal/nodecontrol/hostevidence/remediation.go` and consumed only by the recovery application:

```go
package hostevidence

type HostRemediationAction string

const (
	HostRemediationRegisterIncident       HostRemediationAction = "register_host_security_incident"
	HostRemediationCompleteReenrollment   HostRemediationAction = "complete_reenrollment"
	HostRemediationClearSecurityLatches   HostRemediationAction = "clear_security_latches"
	HostRemediationReauthorizeAfterRestore HostRemediationAction = "reauthorize_after_restore"
	HostRemediationRegisterResourceEnvelope HostRemediationAction = "register_resource_envelope"
)

type GuardCounterEvidenceV1 struct {
	CounterIdentity string
	CounterValue    uint64
	StateDigest     contracts.Digest
}

type HostRemediationEvidenceV1 struct {
	SchemaVersion             string
	EvidenceID                uuid.UUID
	NodeID                    uuid.UUID
	IncidentID                uuid.UUID
	Action                    HostRemediationAction
	AgentBuildDigest          contracts.Digest
	SupervisorBuildDigest     contracts.Digest
	AgentRollbackGuard        GuardCounterEvidenceV1
	SupervisorRollbackState   GuardCounterEvidenceV1
	SecurityLatchGuard        GuardCounterEvidenceV1
	TrustedTimeProviderID     string
	TrustedTimeFloor          time.Time
	InstalledMapVersion       uint64
	InstalledMapDigest        contracts.Digest
	CompletedAt               time.Time
	DeploymentKeyID           string
	Algorithm                 string
	DeploymentAuthoritySig    []byte
}

type HostRemediationVerifier interface {
	Verify(context.Context, HostRemediationEvidenceV1) (VerifiedHostRemediation, error)
}
```

The five literals above are the complete action universe: `RegisterHostSecurityIncident`、`CompleteReenrollment`、`ClearSecurityQuarantine`、`ReauthorizeAfterRestore` and `ResourceEnvelopeService.Register` accept only their same-order respective literal. The verifier accepts only the exact role=`host_remediation` Ed25519 transcript and returns a defensive, digest-bound `VerifiedHostRemediation`; the recovery/resource-envelope repositories enforce 15-minute freshness, one-time use and exact node/incident/action binding. Empty、case/alias、unknown and every cross-action pairing reject before mutation.

The agent platform boundaries are created in `internal/nodeagent/localstate/providers.go`:

```go
package localstate

type MonotonicCounter interface {
	Identity(context.Context) (string, error)
	Read(context.Context) (uint64, error)
	Increment(context.Context, uint64) error
}

type AuthenticatedSealer interface {
	Identity(context.Context) (string, error)
	Seal(context.Context, string, []byte) ([]byte, error)
	Open(context.Context, string, []byte) ([]byte, error)
}

type TrustedTimeSource interface {
	Identity(context.Context) (string, error)
	Now(context.Context) (TrustedInstant, error)
	AttestFloor(context.Context) (TimeFloorAttestation, error)
}
```

Agent main, supervisor and latch construction each receive distinct counter/sealer identities and reject equality before reading key/config or starting a child.

The local supervisor client surface is created in `internal/nodesupervisor/wire/client.go`:

```go
package wire

type Client interface {
	Handshake(context.Context, HandshakeRequest) (HandshakeResponse, error)
	Prepare(context.Context, PrepareRequest, CredentialFD) (PrepareResponse, error)
	Check(context.Context, CheckRequest) (CheckResponse, error)
	Start(context.Context, StartRequest) (StartResponse, error)
	Probe(context.Context, ProbeRequest, CredentialFD) (ProbeResponse, error)
	Drain(context.Context, DrainRequest) (DrainResponse, error)
	Stop(context.Context, StopRequest) (StopResponse, error)
	Renew(context.Context, RenewRequest) (RenewResponse, error)
	Rollback(context.Context, RollbackRequest) (RollbackResponse, error)
	ListFaults(context.Context, ListFaultsRequest) (ListFaultsResponse, error)
	ClearFault(context.Context, ClearFaultRequest) (ClearFaultResponse, error)
}
```

No request includes executable path, argv, environment, shell, arbitrary config bytes, PID or signal. `CredentialFD` is valid only for exact test-profile prepare/probe roles; its zero value means no FD. `StopRequest` is one strict discriminated union: `AllOwned=false` requires the complete exact slot/generation/authority-sequence/desired/boot binding and nonzero lease ID；`AllOwned=true` is allowed only on the verified persistent peer for the configured node/current supervisor boot, requires empty lease/slot/generation/authority-sequence/desired fields, and may inspect/stop only exact entries in that supervisor boot's node-scoped ownership ledger/cgroup root. Cross-node/boot input、mixed forms、PID/pattern/name enumeration and caller-selected targets reject before mutation.

## Canonical artifact ownership and design traceability

These names are normative; implementations must not shorten, alias or redefine them in a second package:

| Canonical artifact or field | Owning file/task | Required consumers |
| --- | --- | --- |
| `contracts.Adapter`, `contracts.DesiredReasonV1`, `contracts.Nonce32`, `contracts.ProcessSpecV1`, `contracts.CapacityLimitsV1` | `internal/nodecontrol/contracts/process_spec.go`, Plan 02 Task 5 | desired/time signer, recovery attestation, agent preview/reconciler, supervisor compiler, Xray and sing-box adapters |
| `contracts.LocalVersionedDigestV1` / `CompareLocalVersionedDigest` | `internal/nodecontrol/contracts/authority.go`, Plan 01 Task 1 | Plan 04 agent and Plan 05 supervisor approved-manifest/installed-map rollback guards only；never an authority-bearing stream |
| `ApprovedReleaseManifestV1` / `InstalledReleaseMapV1`, strict verifiers and build-embedded raw source | `internal/nodecontrol/contracts/local_release.go` + `internal/localrelease`, Plan 02 Task 5；fixed instance updated only by Plan 07 `lock-core` | Plan 04 and Plan 05 independently verify manifest/map/root facts and derive the two local tuples；Plan 07 final locks and P08 exact-cover scanner consume the same manifest digest |
| `NodeTimeAttestationV1` | `internal/nodecontrol/state/contracts.go`, Plan 02 Tasks 5/8 | bootstrap and poll responses, agent trusted-time verifier and rollback guard |
| `NodeResourceEnvelopeV1`, `NodeResourceEnvelopePackageV1` | `internal/nodecontrol/contracts/resource_envelope.go`, Plan 02 Task 5 | Plan 03 unique fenced resource-envelope publisher/inventory pointer, agent trust verifier, supervisor reservation/sandbox |
| `HostMemoryIsolationPolicyV1`, `HostMemoryIsolationPolicyPackageV1` | `internal/nodecontrol/contracts/host_memory_policy.go`, Plan 04 Task B04-T04 | agent verifier, supervisor sandbox, Linux platform evidence |
| `RecoveryStateSnapshotV1` | `internal/nodecontrol/state/contracts.go`, Plan 02 Task 5 | recovery signer saga, agent recovery verifier/latch, API authorizers |
| generated `nodeagentv1.SecurityFaultReceiptV1` | Plan 01 Task 2 OpenAPI/generator output | agent recovery wire response and specialized receipt commit gate |
| `DeliverableSecurityFaultReceipt` | `internal/nodecontrol/recovery/types.go`, Plan 03 Task 7 | fence-finalized domain receipt；Plan 03 maps it field-by-field to the generated wire receipt |
| `HostRemediationAction`, `HostRemediationEvidenceV1`, `HostRemediationVerifier` | `internal/nodecontrol/hostevidence/remediation.go`, Plan 03 Task 7 | exact five-action matrix for incident registration, reenrollment completion, typed latch clear, restore reauthorization and resource-envelope registration |
| generated `nodeagentv1.TrustConflictEvidenceRequestV1` / `TrustConflictEvidenceAckV1`, high-water-conflict and ACK commit bindings | Plan 01 Task 2 generated API；`internal/nodecontrol/{recovery,identity}` and `internal/nodeagentapi`, Plan 03 Tasks 6–8 | node-local bounded unverified incident, independent full-artifact verification queue and atomic typed escalation；never globalizes a client claim |
| `LeaseRefreshSealV1`, `RollbackSealV1` | `internal/nodesupervisor/runtime/types.go`, Plan 05 Task B05-T06 | supervisor runtime only; agent may receive only typed result IDs |
| agent `ProfileV1`/`PreviewV1`/`Register`/`Preview`; supervisor `CompileRequestV1`/`CompiledConfigV1`/`CompileFunc`/`RegisterCompiler` | `internal/nodeagent/adapter/{types,registry}.go`, Plan 06 Task B06-T02；`internal/nodesupervisor/adapter/{types,compiler}.go`, Plan 05 Task B05-T03 | fixture、Xray、sing-box register into these sole registries；P07 owns no shadow surface |
| offline `talenro-core-lock`, `CoreSmokeBuildReceiptV1`, `RunnerReceiptFinalizer`, protected result slot and full `ValidatedCoreSmokeReceiptSetV1` | `internal/coreartifactlock` + `cmd/talenro-core-lock`, Plan 07 Task 2；`internal/c12acceptance/coresmoke/{receipts,result}.go` + `internal/c12test/coresmokerunner/{types,channel_linux}.go`, Plan 07 Tasks 2/5 | tracked runner manifest pins recipe/toolchain/closed source-selection policy but no materialized source or final output hash；same-committed-tree helper receipt uniquely carries exact selected source closure plus broker/test/test2json/image hashes；P08 native parent只能在 exact child cleanup 后从 inherited FD 3 sealed slot 一次 adopt 并重验完整 set，不能从 bytes/path/digest/FD scalar mint trusted source |
| `C12NativeRunnerProfileV1`, `C12NativeRunnerInstallReceiptV1`, `C12FixedRunnerLauncherAttestationV1`, fixed sealed bootstrap, per-role rollback-resistant state guard/A-B cells, staged-diagnostic/Windows/platform/authority cleanup successors plus Windows/platform/authority post-clean successors, guarded cleanup finalization, completion state/child runtime and opaque one-use native sessions | `internal/c12runnerprofile/{profile,bootstrap,state_guard,staged_diagnostic_cleanup,cleanup_finalization,windows_cleanup,platform_resume,authority_cleanup,native_trusted_time_runtime,completion_state,completion_child_runtime,channel}.go` + `cmd/talenro-c12-runner-launcher`, Plan 08 Tasks 2/6/7；five fixed profile instances in Plans 08–09 | 四个 final roles 与一个 `staged_diagnostic` 的 tracked policy；receipt只覆盖native executables，snapshot scripts和machine-base launcher无receipt；OS adapter只开固定role-indexed machine store并以不可重置guard的epoch+digest唯一选择runtime cell及immutable install record，bootstrap无active pointer；attested launcher先建containment与execution reservation再启动helper。installer首次在`(0,zero)`保持uninstalled直到sole advance，inactive-only重装在guard前保留旧generation、guard后只选新generation且保留next/retained payload；每个provision mode必须持有OS deployment transaction的host/role/projection-bound candidate handle，install线性消费并销毁session/handle；fresh Open一次，active/reboot只可pathless recovery。所有非completion restart先走 `abort_unstarted|abort_intent|role_dispatch`；Windows/platform/authority success-origin CAS已把有界canonical publication-input tuple完整写入A/B cell；exact-clean后停在same-generation post-clean successor，`PostCleanupPublicationView`只在current-state reread后返回defensive bytes+`PublicationInputsDigest`，wrapper-only facade据此cold-rebuild并attest/send，staged-diagnostic 的成功、失败或崩溃 cleanup 均走 ordinary terminal 且无 attestor/sender/publisher，其他 abort/crash 也走 ordinary terminal。Completion E2 create只能走独立只读audit，revalidation pending outcome必须先durable-deliver+ACK |
| `C12OuterImageBuildContextV1`, `C12OperationsBuildPolicyV1` and minimal exact-commit Git object database capsule | `internal/artifactscan/buildinputs.go`, Plan 08 Task 3；fixed context manifests and Plan 09 Task 6A policy instance | 三个 outer image context strict exact-cover且排除 rewrite outputs；nested verifier只挂只读 `/run/c12-input/object-db` 并由固定 parent重验 locator/identity/closure，不能依赖 host ambient repo或无 object database 的 snapshot |
| `ProtectedWALKeyBindingCoreV1`/`ProtectedWALKeyBindingV1`, leaf key intent/handle + sealed selected-WAL binding/recorder, cleanup capsule/plan/bootstrap/fixed-selected wrapper, cleanup capability/proof/receipt/tombstone, `C12AttestedEvidenceTransactionV1`, opaque adoption set and completion coordinator | `internal/c12cleanup/{core,protected_key_*}.go` + `internal/c12evidence/{wal,cleanup}.go` + `internal/c12runnerprofile/{bootstrap,staged_diagnostic_cleanup,cleanup_finalization,windows_cleanup,platform_resume,authority_cleanup,completion_state,channel}.go` + `internal/c12completionpublication/operator.go`, Plan 08 Tasks 2/6/7 and Plan 09 Tasks 4/6B/7 | c12cleanup leaf独占resource、key、≤16-KiB machine-sealed selected-WAL binding envelope、recorder、cleanup capability/proof/receipt及exact codec/engine；c12evidence只拥有非completion high-level plan/bootstrap/fixed-selected opaque wrappers与shared dispatch，零 raw leaf key/selected-WAL/cleanup Bind/Open 调用。runnerprofile guarded bootstrap 从 session+reservations 内部绑定完整 plan，独占 raw key Open/Create/Destroy/Seal，并在 current-state reread 后由 opaque fixed-selected token 独占 raw selected-WAL Open；五个 cleanup successor/handle owner 私下重建 declaration-bound adapter set、Bind 后 guarded Open。completion coordinator另独占binding Marshal/store/Recover/Open，且Begin/两种session recovery只在WAL bootstrap/actual durable后返回私持recorder的handle；runtime资源创建强制通过它。所有finalization在selected pending内由两个私有 current-state gate 完成leaf Recover/Commit、key absent、tombstone unlink+parent-fsync+absence，最后CAS才选target，CAS后无filesystem。runnerprofile另拥有五路inner dispatch、两种session class、I/T/Sp/Sa/Fp/Fa DAG、sink/view/decoder；c12completionpublication唯一配对handle+delete capability并消费低层publication facade。任何raw token/binding/recorder/encoding escape、receipt/target替换、abort→publication或success→ordinary finalizer都拒绝 |
| `C12ScopeEvidenceV1` | `internal/c12evidence/scope.go`, Plan 08 Task 1 | both Windows wrappers and both production-provider gates |
| `C12SpecSetV1` / `SpecDigest` | `internal/c12evidence/specset.go`, Plan 08 Task 1 | every scope evidence builder/validator and Plan 09 completion consumer; no second builder |
| `C12TrustedTimeEvidenceV1`, sealed authenticated trusted-time client/runtime adapters, `PlatformEvidenceV1`, `PlatformTrustPolicy`, cross-reboot manifest builder/open/cleanup/post-clean facades | `internal/c12evidence/{trustedtime,platform}.go` + `internal/c12runnerprofile/{native_trusted_time_runtime,platform_resume}.go`, Plan 08 Task 2 and Plan 09 Task 4 | P09 c12evidence owns the sole trusted-time schema/semantic validators and opaque high-level platform facades；raw P08 sink/successor/runtime/result/cleanup/post-clean bridge selectors are exact-AST confined to those two c12evidence files, commands/helpers receive none. Completion carries the full package-derived nested platform policy and calls the same complete validator on create/revalidate |
| `AuthorityFenceEvidenceV1`, `C12CompletionManifestV1`, `ValidatedCompletionManifest`, `ValidatedRetainedCompletion`, sealed creation-time baseline/latest trusted-time head, `C12CompletionRoleStateV1`, terminal/outcome audit and durable-delivery receipts | `internal/c12evidence/{authority,completion}.go` + `internal/c12runnerprofile/{authority_cleanup,completion_state,channel}.go` + `internal/c12completionpublication/operator.go`, Plans 08 Tasks 2/7 and 09 Tasks 6B/7 | c12evidence owns authority/completion semantic validation plus distinct create/retained tokens；runnerprofile owns sole logical state、latest accepted head、closed `completed|aborted` lineage、outcome-delivery pending/dispatch/result-slot receipt、低层 sink/view 及 B10-frozen structural decoder；P09 composite owns opaque publish/revalidation handle、high-level operator、paired inner-cleanup session，并将 terminal-audit、outcome-audit、delivery receipt 与 active revalidation handle 保持为不可互换类型。First revalidation必须从完整signed create envelope生成baseline，后续head严格单调；post-terminal audit durable-deliver+ACK后方可fresh Open，任何dispatch错误fail closed |
| `AuthorityOperationsArtifactTupleV1`、`AuthorityOperationsBinaryAndImageDigest` | `internal/c12evidence/scope.go` + `internal/artifactscan/scan.go`, Plan 08 Tasks 1/3；final producer Plan 09 Task 6A | scanner receipt、all four scopes、cross-reboot/platform/authority evidence、immutable OCI launcher、CompletionPolicy/manifest；the tuple exposes binary SHA、image manifest、build receipt and dependency closure, while the combined digest is recomputed and remains outside `ReleaseDigests[3]` |
| `AuthorityOperationsSupplyChainRecordSetDigest`、`SupplyChainEvidenceBundleDigest` | `internal/artifactscan/scan.go`, Plan 08 Task 3；consumers in `internal/c12evidence/{scope,platform,authority,completion}.go`, Plans 08–09 | subject-checked nine-record references、bounded canonical retained bundle、four scopes、nested platform/authority evidence、receipt-bound license projection、completion/revalidation；a parallel clean record/license set cannot substitute |
| `ArtifactCatalogDigest`、`ExpectedProcessImagePolicyDigest`、`ImageSetDigest`、`ArtifactScanDigest` | sole tracked `testdata/c12/artifact-catalog.v1.json` + `internal/artifactscan/scan.go`, Plans 08 Task 3 and 09 Task 6A | exact-cover scanner、four scope policies、nested platform/authority evidence、completion manifest/revalidation、Windows/Linux helper roles and both machine-base launcher binary/SBOM/provenance triples；launchers are B10/B11 frozen inputs rather than rewrite outputs or runtime receipts, and runtime receipts remain scope-specific without making the canonical receipt run-dependent |
| `authority.NodeCheckpoint` / `node_authority_checkpoint_sequence` | `internal/nodecontrol/authority/provider.go` and `repository.go`, Plan 01 Tasks 5–6; signed field in Plan 02 Tasks 5/8 | poll high-water and time attestation, desired/recovery audience, agent and supervisor rollback state |
| `AuthorityEffectCommitmentV1`, `AuthorityProviderHeadV1`, `AuthorityCheckpointAnchorV1`, `ActivationDecisionEvidenceV1`, `AuthorityEffectResolutionV1`, sealed dispatcher and branded contextual proof | `internal/nodecontrol/authority/{commitment,resolution,activation_evidence,effect_dispatcher}.go`, Plan 01 Tasks 7–9 | B02/B03 registered resolvers/three-method activators；Coordinator same-Provider capture、atomic activation and parsed persisted-outcome recovery |
| `serving.BoundAuthorityReadSource` / guarded certificate/desired facts | `internal/nodecontrol/serving`, Plan 01 Tasks 10 and 15 | B02 desired reader and B03 mTLS authorizer only |
| v7 `CanonicalBodyDigest`, `ParseCanonicalBody`, `VerifiedEnvelope`, `VerifiedEvidenceBundle`, `VerifySignatureEnvelope`, `VerifyCanonicalEvidenceBundle` and fixed schema/role registry | `internal/nodecontrol/contracts/{canonical,signature_envelope,evidence_bundle,schema_registry}.go`, Plan 01 Task 11 | B03 provider/attestor adapters and B11 orchestrator; B10 only hashes the approved spec-set |
| `DatabaseAuthorityRebindGapAttestationV1`, `DatabaseAuthorityPostRecoveryRebindAttestationV1`, genesis/epoch/lease/staging/source/Down contract families | `internal/nodecontrol/contracts/authority_v7_*.go`, Plan 01 Tasks 11–13 | B03 production signer/provider implementations; B11 supplies exact preimages and sequencing |
| generic related-row set, purpose/generation evidence, multi-key Fence, semantic journal/epoch subject slot/stable-holder once-row, archive Inspect/history/staging evidence modes | `internal/nodecontrol/contracts/commit_archive.go` and `internal/nodecontrol/authority/commit_archive_verifier.go`, Plan 01 Task 14 | B03 rollback-resistant archive runtime; B11 orchestration only |
| 26 additive v7 protocol tables, ACL/guard/lock registries and `DatabaseAuthorityHeadV1` exact-epoch projection | registered Go `00007`, `db/schema/nodecontrol.v1.yaml`, Plan 01 Tasks 8 and 15 | all domain repositories/readiness; B02 imports through B01 functions and creates no schema |

Spec coverage remains explicit across the suite:

| Approved design sections | Plan ownership |
| --- | --- |
| §2–5 boundary, decision and architecture | this index plus Plan 01 Tasks 1–3 |
| §6 identity, rotation, application authorization and recovery | Plan 03 Tasks 1–7 |
| §7 signing, metadata and root threshold workflow | Plan 02 Tasks 5–9 |
| §8 isolated APIs, endpoint matrix, bounds and events | Plan 01 Tasks 2–3; Plan 03 Tasks 6, 8–10; Plan 06 Task B07-T03 |
| §9 PostgreSQL authority and PITR fence plus approved Abort/serving amendment §§4–8 | Plan 01 Tasks 4–10; Plans 02–03 per-kind activators; Plan 09 Tasks 6A–6B |
| §10 inventory and two-level quarantine | Plan 02 Tasks 1–4; Plan 03 Task 7; Plans 04–06 |
| §11 desired/recovery snapshots | Plan 02 Tasks 5–8; Plans 04–05 |
| §12 local guards, trusted time and host policy | Plan 04 Tasks B04-T01–T04; Plan 09 Task 5 |
| §13 reconciler, supervisor and adapters | Plans 04–05; Plan 07 |
| §14–16 observation, health, capacity, privacy | Plan 06 Tasks B07-T01–T05 |
| §17 deterministic, integration, mTLS, process and load tests | every plan's exit gate; Plans 06–08 own the composed harness |
| §18 four-scope authority completion plus canonical spec-set binding | Plans 08–09 |
| §19 threat model and runbooks | Plan 09 Tasks 1–3 |
| §20 license and supply chain | Plans 07–09 |
| v7 amendment §1–9 protocol/profile/genesis/epoch/source/staging/Down | Plan 01 Tasks 8 and 11–18 own contracts/schema/verifiers/tests; Plan 03 owns production provider/attestor/archive; Plan 09 owns orchestration/conformance |
| v7 amendment §10 ownership and §11 TDD matrix | this index ownership table plus Plan 01 Tasks 8 and 11–18; Plans 02/03/08/09 consume only their assigned interfaces |

Completion keeps `ReleaseDigests` exactly `[3]`. The measured operations tuple、recomputed combined digest、subject-checked record-set、canonical bundle、artifact catalog、process-image policy、full role-bound image set and `ArtifactScanDigest` must all be nonzero and byte-identical in all four `C12ScopeEvidenceV1` documents, nested `PlatformEvidenceV1`/`AuthorityFenceEvidenceV1`, the validated full `ScanReceiptV1`/bundle and `C12CompletionManifestV1`；any missing field, alias to the control-api digest, component/combined disagreement, omitted catalog role, parallel clean license set or cross-artifact/full-receipt splice fails before completion signing.

## Generation, verification, and commit discipline

Every task uses these exact local commands unless its plan specifies a narrower package first:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1
$env:CGO_ENABLED = '0'
go test ./... -count=1 -timeout 30m
$env:CGO_ENABLED = '1'
go test -race ./... -count=1 -timeout 30m
$env:CGO_ENABLED = '0'
go vet ./...
go tool golangci-lint run ./...
```

For every task that changes generated artifacts, its batch plan's exact path list is normative: run the first generation, stage only those generated paths, run generation a second time, require `git diff --exit-code -- <the same explicit paths>` and the matching untracked-path query to be empty, then commit that task. The suite-level `git diff HEAD` gate is run only at batch exit after all task commits；using it between the first generation and the task commit would incorrectly compare legitimate output to the prior task's `HEAD`.

The environment assignment is normative for every narrower command in the batch plans as well: each non-race Go build/test/vet/lint process is launched with `CGO_ENABLED=0`, and each race process with `CGO_ENABLED=1`；executors must not inherit an ambient value. Every planned `*_integration_test.go`, including both composed B03 E2E files, has `//go:build integration` as its first line and is excluded from untagged package gates. The full static block below runs only at the final suite gate, after B03、B10 and B11 have created every named file；the B03/B10/B11 owning exits repeat their exact existing-file subsets, while earlier batches scan only then-existing `*_integration_test.go` files. It compares literal first lines and requires no output；`rg -L` is not used because in ripgrep it means `--follow`, not files-without-match. Explicit untagged package lists must not name an integration-only directory.

```powershell
$integrationTaggedFiles = @(
    @(rg --files -g '*_integration_test.go')
    'internal/e2e/nodecontrol_mtls_test.go'
    'internal/e2e/nodecontrol_authority_v7_test.go'
) | Sort-Object -Unique
$badIntegrationTags = @($integrationTaggedFiles | Where-Object {
    -not (Test-Path -LiteralPath $_ -PathType Leaf) -or
    (Get-Content -LiteralPath $_ -TotalCount 1) -cne '//go:build integration'
})
if ($badIntegrationTags.Count -ne 0) {
    $badIntegrationTags
    throw 'integration test missing exact first-line build tag'
}

$customBuildTags = [ordered]@{
    'internal/nodesupervisor/wire/peer_stub.go' = '//go:build !linux'
    'internal/nodesupervisor/wire/fd_stub.go' = '//go:build !linux'
    'internal/nodesupervisor/wire/platform_stub_test.go' = '//go:build !linux'
    'internal/nodesupervisor/sandbox/platform_stub.go' = '//go:build !linux'
    'internal/nodesupervisor/sandbox/exec_stub.go' = '//go:build !linux'
    'internal/nodesupervisor/sandbox/exec_stub_test.go' = '//go:build !linux'
    'internal/c12test/core_smoke_credentials_linux.go' = '//go:build linux'
    'internal/c12test/core_smoke_credentials_stub.go' = '//go:build !linux'
    'internal/c12test/load/runner_linux.go' = '//go:build linux'
    'internal/c12test/load/runner_stub.go' = '//go:build !linux'
    'internal/c12test/coresmokerunner/runner_linux.go' = '//go:build linux'
    'internal/c12test/coresmokerunner/channel_linux.go' = '//go:build linux'
    'internal/c12test/coresmokerunner/runner_stub.go' = '//go:build !linux'
    'internal/c12test/coresmokerunner/channel_linux_test.go' = '//go:build linux'
    'internal/c12test/coresmokebroker/broker_linux.go' = '//go:build linux'
    'internal/c12test/coresmokebroker/broker_stub.go' = '//go:build !linux'
    'internal/c12test/coresmokebroker/broker_linux_test.go' = '//go:build linux'
    'internal/e2e/c12_controlled_process_test.go' = '//go:build c12fixture'
    'internal/e2e/c12_lkg_expiry_test.go' = '//go:build c12fixture'
    'internal/e2e/c12_crash_recovery_test.go' = '//go:build c12fixture'
    'internal/e2e/c12_security_latch_test.go' = '//go:build c12fixture'
    'internal/e2e/c12_resource_faults_test.go' = '//go:build c12fixture && linux'
    'internal/e2e/c12_sandbox_faults_test.go' = '//go:build c12fixture && linux'
    'internal/e2e/c12_cleanup_test.go' = '//go:build c12fixture'
    'internal/e2e/c12_privacy_test.go' = '//go:build c12fixture'
    'internal/e2e/c12_api_bounds_test.go' = '//go:build c12fixture'
    'internal/c12acceptance/coresmoke/xray_test.go' = '//go:build c12_core_smoke && linux'
    'internal/c12acceptance/coresmoke/singbox_test.go' = '//go:build c12_core_smoke && linux'
    'internal/c12runnerprofile/completion_child_runtime_windows.go' = '//go:build windows'
    'internal/c12runnerprofile/completion_child_runtime_windows_test.go' = '//go:build windows'
    'internal/c12runnerprofile/completion_child_runtime_unsupported.go' = '//go:build !windows'
    'internal/c12runnerprofile/native_trusted_time_runtime_linux.go' = '//go:build linux'
    'internal/c12runnerprofile/native_trusted_time_runtime_linux_test.go' = '//go:build linux'
    'internal/c12runnerprofile/native_trusted_time_runtime_unsupported.go' = '//go:build !linux'
    'internal/c12runnerprofile/state_guard_windows.go' = '//go:build windows'
    'internal/c12runnerprofile/state_guard_windows_test.go' = '//go:build windows'
    'internal/c12runnerprofile/state_guard_linux.go' = '//go:build linux'
    'internal/c12runnerprofile/state_guard_linux_test.go' = '//go:build linux'
    'internal/c12cleanup/protected_key_windows.go' = '//go:build windows'
    'internal/c12cleanup/protected_key_windows_test.go' = '//go:build windows'
    'internal/c12cleanup/protected_key_linux.go' = '//go:build linux'
    'internal/c12cleanup/protected_key_linux_test.go' = '//go:build linux'
    'cmd/talenro-c12-runner-launcher/launcher_windows.go' = '//go:build windows'
    'cmd/talenro-c12-runner-launcher/launcher_windows_test.go' = '//go:build windows'
    'cmd/talenro-c12-runner-launcher/launcher_linux.go' = '//go:build linux'
    'cmd/talenro-c12-runner-launcher/launcher_linux_test.go' = '//go:build linux'
    'cmd/talenro-c12-runner-helper/runnerhelper_windows.go' = '//go:build windows'
    'cmd/talenro-c12-runner-helper/runnerhelper_windows_test.go' = '//go:build windows'
    'cmd/talenro-c12-runner-helper/runnerhelper_linux.go' = '//go:build linux'
    'cmd/talenro-c12-runner-helper/runnerhelper_linux_test.go' = '//go:build linux'
    'internal/c12evidence/platform_gate_test.go' = '//go:build c12_platform'
    'internal/nodecontrol/hostevidence/operator_guard_conformance_test.go' = '//go:build c12_platform'
    'internal/nodecontrol/authority/production_conformance_test.go' = '//go:build c12_authority'
    'internal/nodecontrol/operations/production_conformance_test.go' = '//go:build c12_authority'
    'internal/c12evidence/compose_c12_docker_test.go' = '//go:build c12_docker'
    'internal/c12evidence/authority_script_c12_docker_test.go' = '//go:build c12_docker'
    'internal/testinfra/c12_authority_pitr_integration.go' = '//go:build integration'
    'internal/nodecontrol/authority/v7_opaque_integration_fixture.go' = '//go:build integration'
}
$badCustomTags = @($customBuildTags.GetEnumerator() | Where-Object {
    -not (Test-Path -LiteralPath $_.Key -PathType Leaf) -or
    (Get-Content -LiteralPath $_.Key -TotalCount 1) -cne $_.Value
} | ForEach-Object Key)
if ($badCustomTags.Count -ne 0) {
    $badCustomTags
    throw 'custom conformance test missing exact first-line build tag'
}
```

P01 Task 4 owns `scripts/run-c12-integration.ps1`; P01 Task 8 extends it with registered disposable-v7 and infrastructure-only PITR profiles before P02/P03 execute. Because Windows PowerShell 5.1 `-File` cannot bind an array-valued script parameter, focused mode accepts only `-Profile base|authority-v7|authority-v7-pitr`, one strictly validated pipe-delimited `-Packages` string (each segment is one explicit `./package`, never `./...`, with no empty/duplicate segment or shell metacharacter), optional anchored `-Run`, finite `-Timeout`, and `-Race`. Suite mode is mutually exclusive and accepts only `-Suite batch01|batch02|batch03|final` plus a finite global `-Timeout`. It loads the fixed ordered manifests `testdata/c12/integration-contracts-schema-authority.v1.json`, then additionally `integration-operator-state-signing.v1.json`, then additionally `integration-node-identity-mtls.v1.json` for both `batch03` and `final`. `batch03` closes B03 while future-batch custom-tag files may be absent；`final` additionally performs only the complete literal `$customBuildTags` existence/first-line static validation after P04–P09 have created every listed file. It does **not** execute those custom-tag suites on Windows：`c12fixture`/core-smoke/`c12_platform`/`c12_authority`/`c12_docker` owning positives remain on their explicitly approved Linux、authority or Docker runners, and final completion consumes their attested results. A manifest group contains only a unique ID, one explicit package, `base|authority-v7|authority-v7-pitr`, a sorted non-empty exact test-name set and a bounded timeout；regexes, free-form flags, package globs and shared-database markers are forbidden.

Before executing a suite, a Go/parser-based validator scans every tracked `_test.go` whose literal first line is `//go:build integration`, extracts every top-level `Test*` function, and requires the selected manifest set to cover each `{package,test}` exactly once with no missing/extra/duplicate entry. The sole profile initializer `TestPrepareC12AuthorityV7Database` is explicitly excluded from that test universe and may run only through the runner's private initialization phase；no manifest test may call `t.Skip`. The runner captures `go test -json` and accepts a group only when every listed test has exactly one terminal `pass`, zero `skip`/missing/extra test events and the package passes. The validator also requires `sum(group timeout + fixed per-group setup/cleanup budget) <=` the exact suite budget (`batch01=120m`, `batch02=180m`, `final=240m`), so legal groups cannot overrun an undersized global deadline. The runner executes groups serially with `-p=1`; focused mode likewise handles each listed package as a separate group. Every ordinary group creates a fresh randomly named dependency set: `postgres:18.4-alpine3.23`, `redis:8.8.1-alpine3.23`, and `nats:2.14.3-alpine3.22` with JetStream enabled；the PITR profile additionally uses only its frozen verified primary/candidate/volume state machine. Each container receives a Docker-assigned loopback port；the runner captures and validates all exact names/IDs/health-or-protocol probes/ports, migrates database base through ordinary Goose, optionally applies registered 00007 only through the integration-tagged sealed-grant fixture, and sets group-local dependency variables. It splits validated input only after validation, constructs bounded argument arrays, invokes native executables with Windows PowerShell 5.1-compatible call operators (`& docker @dockerArgs`, `& go @goArgs`) and checks `$LASTEXITCODE` after every call；`Invoke-Expression`, `cmd /c`, `powershell -Command`, `.Arguments` string construction and shell interpolation are forbidden and statically tested absent. Each group `finally` re-inspects and removes only its verified resources and reports test and per-resource cleanup failures separately before the next group. The runner refuses caller dependency URLs/addresses, fixed ports/services, production installation kind, unknown package/test/flag, cleanup target mismatch or unbounded/per-group-over-global timeout. This isolates base migration/Down tests from preinstalled-v7/PITR tests and prevents cross-package schema/PITR races；B01–B03 never depend on B10 or a user service.

At B11 final acceptance this P01-owned `-Suite final` invocation remains a separate locked repository gate：it must finish and clean its disposable Windows dependency set before any protected native scope parent starts；`run-windows-final-scopes` never launches、wraps或attests this PowerShell suite；四作用域 channel/manifest只消费各自 native parent 产生的证据，不能把该独立 repository gate 冒充 `windows_powershell_docker` scope。

`batch03` has the same exact 240-minute sum/setup/global budget as `final`; only the future-file custom-tag phase differs. The budget formula uses exactly three minutes setup+cleanup allowance for each ordinary group and eight minutes for each PITR group, in addition to its manifest timeout；the runner enforces the same phase deadlines. In the parser rule, “every tracked” is scoped to the closed union of explicit package paths named by the selected B01–B03 manifests；within that package universe every first-line integration test must be covered, but a later batch's separate package (for example B06 observation) is excluded and must pass its own owning integration gate. The full first-line/static custom-tag scan remains repository-wide at final.

For the JSON rule, “extra” means another top-level test name. Events beneath a listed parent are allowed only as a nonempty canonical bounded `Parent/subtest` subtree；the parent must pass exactly once, and any malformed/unbounded suffix or skipped descendant rejects.

After all P01–P03 packages exist, the full integration gate is exactly:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Suite final -Timeout 240m
```

At each batch exit, after its generated-output tasks have committed, rerun generation and require the committed tree plus untracked set to remain clean:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1
git diff --exit-code HEAD -- api gen internal/store
if (git ls-files --others --exclude-standard -- api gen internal/store) { throw 'untracked generated artifact' }
```

Do not run tagged integration/E2E against fixed user services. Each plan names the owned harness entrypoint. A task commit stages only its explicit file list; never use `git add .`, `git add -A`, or a path that includes user-owned untracked directories.

## Completion semantics

Implementation completion is not inferred from unit tests or Docker alone. `C12CompletionManifestV1` must validate four unexpired, attested and build-identical scope digests:

1. `windows_powershell_docker`;
2. `windows_git_bash_docker`;
3. `linux_platform_operator_trust`;
4. `authority_fence_pitr`.

The first two may contain `container_deterministic` fake evidence but cannot satisfy the third or fourth scope. The final validator also requires the exact three-entry approved spec set, zero unresolved v7 runtime/epoch/staging/source/Down state, complete activation/completion/release or an explicitly non-complete fresh-restore result, exact-epoch DB/Provider Head equality, and zero Critical/Important independent-review findings. `fresh_v7_staging_closed` by itself is always `restore_incomplete` and can never satisfy C1.2 completion. Only after all four scopes pass, all licenses remain non-bypassed, cleanup/worktree equality succeeds, and the C1.2 runbooks/destructive-restore follow-up are complete may Batch 11 change the roadmap from C1.2 `current design` to `complete` and C1.3 to `current design`.
