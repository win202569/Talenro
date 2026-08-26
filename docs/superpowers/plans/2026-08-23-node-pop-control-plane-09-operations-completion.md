# Talenro C1.2 Operations and Completion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付完整 C1.2 threat model/runbooks、唯一 production Authority Protocol v7 upgrade/restore/Down orchestrator、真实 Linux platform 与 OperatorClientTrustGuard conformance、真实 provider/DB/PITR/destructive-restore conformance，以及只在四个同构建 scope 全部有效时才能生成的完成清单和路线图推进。

**Architecture:** 文档门先把安全事件、registered migration/shutdown procedure、有限操作、验证与停止条件冻结；`nodecontrol-authority-operations` 是唯一 production upgrade/restore/Down entrypoint，只编排 B01 database primitives 与 B03 provider/attestor authenticated APIs，并在所有外部调用前通过 rollback-resistant journal 预分配 semantic IDs。Linux-native gate 在 attested `linux/amd64` TPM/vTPM runner 上跨 cold reboot 验证三个独立 anti-rollback identity、production secure time、host memory policy和真实外部攻击，再验证 production operator trust guard。独立 authority gate 使用 production provider class 的隔离 conformance tenant 和真实 PostgreSQL PITR，覆盖 epoch recovery、staging、source archive 与 disposable Down；最终 validator 重验四份 scope 的签名、时效、canonical spec-set、输入一致性、license 与 cleanup，不能拼接、提升 fake evidence或把 `fresh_v7_staging_closed` 误报成已完成 legacy upgrade。

**Tech Stack:** Go 1.26.5、Linux `/usr/bin/bash`、TPM/vTPM NV counter与keystore、production TrustedTimeSource、SELinux enforcing、Yama 3、BPF lockdown、systemd collector mask、pidfd/perf/BPF/ptrace negative harness、PostgreSQL 18.4 PITR、ControlPlaneAuthorityFence、OperatorClientTrustGuard、RFC 8785 JCS、Ed25519 runner attestation、SHA-256。

**Spec:** [Approved C1.2 node and POP control-plane base design](../specs/2026-08-23-node-pop-control-plane-design.md), approved SHA-256 `B0B0DDBBD11546FC07A25CB76992E375481192A3261270FF442095501CC2B4B5`, especially sections 6.1, 9.1, 12, 17.4, 18–21; [approved authority Abort/serving amendment](../specs/2026-08-24-nodecontrol-authority-abort-serving-design.md), approved content SHA-256 `86996084462A5DE1E7667D56A135E93099EE38E1089464CCDFCFF7284EA0D1D7`, especially §§5.5, 7–10; [approved Authority Protocol v7 additive upgrade amendment](../specs/2026-08-24-nodecontrol-authority-v7-upgrade-design.md), approved content SHA-256 `EFAEBE52BDC3D70BDA8737C893B02752E60ACEC0A08441813CADF1425079FC8E`, especially §10 B11 and §11; canonical set manifest [c12-spec-set.v1.json](../specs/c12-spec-set.v1.json).

## Global Constraints

- 本册只实现 `C1.2-B11`；只有在 B01–B10 全部关闭后执行。
- B11 只消费 Plan 08 `BuildTrackedC12SpecSet`/`ValidateTrackedC12SpecDigest` 的唯一结果，不实现第二个 manifest reader、member hash/JCS builder 或 `SpecDigest` override；四 scope 与 completion manifest 必须绑定同一 tracked tree 上含 base、Abort/serving amendment、Authority Protocol v7 amendment 的 exact canonical set。
- `cmd/nodecontrol-authority-operations` 是唯一 production upgrade、fresh-restore staging/recovery、legacy-source retirement/archive 与 registered 00007 Down orchestrator。普通 Goose CLI、SQL shell、runbook shell snippet、test harness和 control-api handler都不能执行这些流程；它们只能调用该 binary 的 typed subcommand 或只读 Inspect。
- Orchestrator 必须在首次相关外部调用前，通过 commit-attestor/Fence signer 的 authenticated reservation API 在 DB/PITR failure domain 外的 rollback-resistant journal 预分配 upgrade/registration、source/environment/provider retirement、epoch terminal/recovery/deferred-suffix/prefix-decision/application、staging capability/exclusion/recovery/revocation/application、Fence request/attempt 等所有 IDs/nonces。epoch prefix subject-slot、decision ID反向唯一与 reason-specific semantic record由 attestor 原子 CAS；B11 不直接写 journal 或 archive。
- 所有 provider、commit-attestor、Fence signer、inventory/release operator 与 evidence signer 调用都发生在数据库 transaction 之外；唯一例外是冻结 v7 §9.3 的 `authority_protocol_downgrade_authorizer`，它恰在 registered disposable-Down Goose transaction 内签入实际 txid。数据库 transaction 除该一次限界调用外只调用 B01 typed repository/migration functions；B11 不创建 canonical schema、00007 table/trigger/function、global latch/owner、generation、Fence、challenge、Inspect、provider transcript或 attestation实现，且任何其他 flow/signer 不得复用该例外。
- Epoch PITR recovery 固定顺序为 fresh terminal Inspect → durable recovery intent → optional deferred rebind suffix → suffix-bound commit proof → exact provider recovery transcript → same-prefix apply-vs-replacement arbitration → 必要时对未归档 old generation 先调用 atomic multi-key Fence → restricted DB exact historical rehydrate/recovery application → 每个 live candidate 的 fresh proof与唯一 first-consumer applied result → optional pre/post-catchup applied rebind → terminal-application catchup或dedicated reconcile。任何步骤不能跳过、重排、改ID或用 current transaction 重算 provider-consumed historical body。
- Staging 固定编排 capability registration commit proof → provider acquire → import application或revocation application commit proof → Release/Abort；PITR recovery固定为 optional held-preserving/lineage rebind → durable recovery intent+commit proof → provider recovery CAS → DB intent/capability zero-or-exact-one rehydrate+recovery application → generic commit proof → signed recovery revocation+revocation application → fresh staging challenge/attestation → Abort。未attest且被PITR丢失的 `FreshRestoreImportApplicationV1` 永远不得 Fence或重做import，只能走 intent-backed recovery revocation/Abort；staging outcome Fence仅允许预分配的 revocation application。
- `LegacySourceRetirementSetV1`、`LegacySourcePostSealExactCoverV1`、`LegacySourceArchiveEvidenceV1`、`FreshTargetInventoryV1` 与 `FreshV7StagingEvidenceV1` 只由B11生产 signer生成/签署/验证。`FreshV7StagingEvidenceV1.statement=safe_staging_established_and_closed` 只证明安全 staging/restore_incomplete，不能授权route、serving、destructive restore或 legacy-upgrade completion。
- Registered 00007 Down只允许 `installation_kind=disposable_fixture`，且在 transaction 外完成final inventory membership与全部provider永久retirement/zero projection。Goose callback 在唯一 owned `*sql.Tx` 内取得 actual transaction identity并存储 pristine `1/0` inventory；随后恰调用一次 B11 `authority_protocol_downgrade_authorizer`，其签名逐字绑定实际 txid、transaction nonce、inventory/retirement/catalog/anchor preimages，再由 B01 verifier消费同-tx authorization，执行 `1/0 -> 1/1 -> 0/0` guard与DDL/version。除该 signer 外 transaction 内没有 provider/attestor/其他 external call。Production intent、任一source/attempt/protocol row或非零provider history永久拒绝Down。
- B11 conformance必须覆盖production provider、真实PostgreSQL/PITR、source retirement/archive、fresh staging/recovery/Abort、disposable Down与destructive-restore hard blockers；当前 unsupported `trust_bundle_publish` 与 `operator_authorizer_change` 必须使 destructive restore停在 `fresh_v7_staging_closed`，不能通过测试fixture或operator声明绕过。
- Production completion 恰好需要四个 scope：`windows_powershell_docker`、`windows_git_bash_docker`、`linux_platform_operator_trust`、`authority_fence_pitr`。
- 四个 scope 必须有不同 run ID，但 repo commit、tracked-tree、approved spec、toolchain、`ReleaseDigests[3]`、完整 `AuthorityOperationsArtifactTupleV1`、其组合摘要、operations supply-chain record-set及canonical bundle、唯一 artifact catalog、process policy、full `ArtifactScanDigest` 和全部 image digest逐字相同；nested platform/authority evidence与validated scanner receipt也必须携带同一组 scanner-derived 值，任一证据最长 72 小时。
- 四类 final runner 与一个 nonpublishing diagnostic runner 都只能调用 absolute-path、catalog/toolchain hash+OS/arch 已验证的预构建 native helper public parent mode：Windows为零参数 `run-windows-final-scopes`，平台为 `run-platform-phase-a`/冷重启后的 `resume-platform-phase-b`，authority为 `run-authority-final`，Plan 08 B10 的 Task 6A 诊断仅为 `capture-staged-and-build-closed-set`，completion为 `completion-consolidate-create`/`completion-revalidate`。每个 fresh launch generation 只有实际 public parent 可以恰调用一次 `OpenInstalledNativeRunnerSession` 并消费返回的 one-use launch session；安装/预检流程只能调用 non-consuming `VerifyFixedNativeRunnerInstallation(ctx,role)`，不得以 open/destroy session 代替验证。平台 Phase A 的 launch session 在 durable resume-manifest adoption 时原子转移为 resume-only successor；Phase B 只经 `OpenResumeManifest` 恢复该 exact installed executable/provider/session projection，绝不第二次 open 同一 role launch generation。Completion create/revalidate 的进程崩溃也分别只经 fixed role-state 的 `RecoverCompletionPublishPending`/`RecoverCompletionRetainedRevalidation` 恢复相同 generation，绝不第二次 Open；E2 后 create 没有 recovery，只能清除旧 reservation 后读终态 audit receipt。每个 final parent 独占snapshot/session、logical-WAL reservation+actual、两个opaque roots、protected trusted-time/provider/runner-attestor handles、canonical evidence bytes、atomic channel handoff和finally cleanup；diagnostic parent 只持其 profile 允许的 ODB/external-root/recovery/helper/Git/build-tool handles，不能取得 script、provider、channel、signer、evidence-root 或 publishing handle。shell只作为 final parent 从其read-only snapshot按固定file identity启动的一次性finite stage，接收无秘密的closed stage enum并返回bounded result，不得选择repo/script/profile/receipt、持handle或签nested/outer evidence。Parent以显式 Windows inheritance list或Unix FD table直接spawn同一预构建helper的private producer/completion child；operator、shell和runbook没有 `--tree-facts-handle`、`--tracked-tree-digest`或build-root参数。Private `build-closed-set --tree-facts-handle ... --parent-session-handle ...` transcript不在public help且只接受parent直接继承的handles。任何 `go run`/`go tool` 临时 executable、PATH binary、shell转发或未锁中间进程都不得接收该数值。repo root只定位 canonical Git object database；producer只读handle证明的 exact snapshot。两个build root使用独立128-bit opaque child、repository-external、空且独占并由capsule/WAL exact-clean；禁止direct scanner、单根/ambient/alternate producer。staged空CommitOID固定deferred；四份final scope只接受committed locator。
- 上述 completion recovery 只覆盖 E2 前 create active phase 与 revalidate 的 `retained_complete_active`。E2 已把 create reservation 标成 `exiting` 后，旧 containment 必须先 absence-proof+clear；重复 `completion-consolidate-create` 只能以新的不消费 generation 的 `terminal_audit` reservation 调用只读 `InspectCompletionRetainedFinalization`，不得再调用 `RecoverCompletionPublishPending`、`RecoverFixedCompletionRoleState` 或 Open。
- 五类 runner profile 都只是 Plan 08 `internal/c12runnerprofile` 所有的 strict `C12NativeRunnerProfileV1` 实例：四个 final roles `windows_scopes|platform|authority|completion` 加一个 nonpublishing `staged_diagnostic` role。P09 不复制 schema、JCS、loader、bootstrap 或 install-receipt validator。Windows 与 diagnostic 实例由 Plan 08 创建，锁定 token 分别为 `testdata/c12/windows/runner-profile.v1.json` 与 `testdata/c12/diagnostic/runner-profile.v1.json`；platform/authority/completion 实例分别由本册 Tasks 4/6B/7 创建。最终 relock 必须原子重写并提交五份实例，但 completion channel 仍只能 adopt 三个 final sender transactions，永远没有 diagnostic transaction。
- 最终 catalog/toolchain/full scan 必须 exact-cover 同源但 distinct digest/role 的 `c12_runner_helper_windows_amd64` PE 与 `c12_runner_helper_linux_amd64` static ELF、B10-frozen `c12_runner_launcher_windows_amd64` PE 与 `c12_runner_launcher_linux_amd64` ELF及各自exact binary/SBOM/provenance triple、独立预构建 `talenro_artifact_scan_windows_amd64` private completion child，以及 PostgreSQL 18.4、Redis 8.8.1、NATS 2.14.3 integration-only OCI layouts；launcher仍是revalidated input而非B11 rewrite output/role receipt。Windows/Linux 执行前各验证正确 OS/arch/build digest，不得 `go run`、helper-as-launcher alias或跨 OS 复用一份 binary。
- 上一条中的 operator 调用一律落在受 OS 基线认证的固定 `talenro-c12-runner-launcher`，不是稳定 helper 路径。Launcher 通过 Plan 08 `RunFixedNativeRunnerMode` 读取 guard-selected install record、创建受控 execution reservation/Job或cgroup并直接启动版本化 helper；只有该 helper 是实际 public parent。无 launcher 绑定直接运行 helper、旧 generation helper、并发 launcher 或恢复前未证明旧 helper tree absent 都失败。每份 profile 的 `BootstrapLauncherCatalogRole` 必须按 OS 绑定 launcher role，machine-local `C12FixedRunnerLauncherAttestationV1` 是 role install 之前的外部基线前提。
- `C12NativeRunnerInstallReceiptV1` 只覆盖 role-generation native helper、解释器、Docker/Git/PITR/构建工具及 private completion child；machine-base launcher 与 tracked snapshot scripts 永远不属于 role install receipt。每个 public parent 只能从已认证 committed snapshot 按 tracked token 重验 exact blob bytes、Git mode、snapshot root/file identity 与 no-follow one-link regular-file约束后派生 absolute script argv；机器 receipt、caller path 或同名已安装副本都不能替代该证明。
- production Linux/authority scope 不得只引用 nested evidence 路径：`linux_platform_operator_trust.TestResultDigest` 必须 domain-bind 完整 signed `PlatformEvidenceV1` envelope digest，`authority_fence_pitr.TestResultDigest` 必须用独立 domain 绑定完整 signed `AuthorityFenceEvidenceV1` envelope digest。Create/revalidate completion 两种模式都必须显式接收和重验这两份 nested envelope。
- Container deterministic fake、Windows Job Object、Git Bash 与 Docker test-host 都不能满足真实 Linux/platform/provider 或 authority-fence scope。
- Linux platform/authority parent 必须在固定、attested `linux/amd64` runner 上按catalog identity启动native `/usr/bin/bash` one-shot stages；Bash不是parent/finalizer，production target不扩展到Windows。
- Linux parent 只能以验证并 pin 住的 `/usr/bin/bash --noprofile --norc <parent-derived absolute snapshot script>` 启动 finite stage，并以 locked snapshot directory 为 cwd、以 native parent 构造的 closed sanitized environment 运行；script path/cwd/env 不是 caller input，stdin/stdout 只传 closed stage enum 与 bounded result。Bash 不得重扫/装载 OCI、访问 provider/scanner/deleter、比较或签 receipt/bundle/evidence、持 FD/HANDLE 或决定 cleanup/publish。
- Agent main、supervisor 与 LocalSecurityLatch 分配三个独立 run-scoped NV counter/keystore identity；任一 identity 相等、handle cross-swap、blob/counter不匹配或 rollback 都 fail closed。
- Protected WAL-key 是 capsule 前唯一允许的 durable backend create，且必须 intent-first：先由 Plan 08 guard 选择绑定 deterministic backend reservation 的 `protected_key_intent`，再 exact-idempotent CreateOrRecover；随后高层 composite 从不含 handle/final-WAL digest 的模板内部构造完整 `C12CleanupOnlyOwnershipCapsuleV1`，file+directory fsync 并以唯一 guard CAS 消费 intent、选择 exact capsule。每次重启先执行共享三路 dispatch：已 claim active generation 但尚无 selected intent/capsule/WAL/resource 时只走零创建 `abort_unstarted` 并终结该 generation；selected intent 时只走 `abort_intent`，exact 删除候选、销毁/验证该一把 key 并进入固定非肯定 target；只有已认证 selected role state 且两种 abort 条件均不存在时才走 `role_dispatch`。任何 inspect/abort 错误不能回退到 role/workload，只有 selected capsule 后才允许第一个 ordinary producer/provider/resource intent。完整 cross-reboot resume manifest 在访问任何 NV handle、build root 或 resume path 前验证 authenticated seal、`StartedAt`/`ExpiresAt`、runner/provider identity、phase、run ID、exact handle digest、counter/digest floor、commit/Git-tree locator/canonical tracked-tree/spec/release/toolchain/scanner facts与ownership/capsule/cleanup digests；manifest durable adoption 只有在candidate durable后由guard compare-and-advance原子承接该 slot 并作废 capsule，pre-guard candidate无authority，absent/partial/mismatched manifest 的重启只能依guard-selected unstarted/intent/capsule exact-abort/clean或fail-closed quarantine，禁止 TPM 全局 enumeration。
- 本计划中 platform、staged-diagnostic 与 authority facade “derive/build cleanup plan”的每一处都严格表示调用 Plan 08 高层 `BindCleanupOnlyOwnershipPlan(ctx,session,template)`；只有 `internal/c12evidence/cleanup.go` 可继续调用 raw runnerprofile `BindFixedCleanupOnlyOwnershipPlan|InspectFixedCleanupOnlyOwnershipPlan`，role files 和 commands 均不得直接引用。
- Phase A 与 Phase B 必须跨受控 cold reboot、保持同一 TPM/VM/provider identity并使用不同 boot ID；旧 sealed state、transition capsule、latch与 counter交叉回放全部拒绝。
- Production TrustedTimeSource/provider unavailable、wall-clock rollback与允许时的VM snapshot rollback必须阻止TLS authorization和所有core启动，不能以signed time hint或文件时间代替。
- Unsafe host phase逐项覆盖Yama非3、`unprivileged_bpf_disabled`非1、`fs.suid_dumpable`非0、非空/pipe core pattern、`core_uses_pid`非0、collector unmask、SELinux permissive或policy/domain/perf-BPF权限漂移；该phase不得生成test credential。
- Positive host phase来自fresh attested VM；Yama 3与unprivileged BPF disabled在boot早期设置并不得在线降低，五个sysctl、SELinux policy/domain-map与collector mask在reboot前后逐字一致。
- External attacker必须在target seccomp之外，以相同core UID/domain及不同UID/core-domain分别尝试user-stack/register perf、ptrace、process_vm、`/proc/<pid>/mem`、pidfd_getfd与BPF load；不得用更严格attacker sandbox制造假通过。
- Approved core、native-check与smoke-client都需要post-start image/isolation receipt和canary crash；不得要求外部target伪装为dumpable 0。
- OperatorClientTrustGuard必须使用真实production provider，覆盖较低package、same-version fork、disk rollback/restart、provider restart/unavailable、normal overlap、emergency removal、nonce replay和attestation expiry。
- Authority-fence gate使用与production相同provider class/policy、独立attested conformance tenant/namespace和run-scoped credential；不得访问production namespace。
- Provider reservation/receipt是append-only，结束时只能变成terminal committed/aborted并按retention保留，不能删除；任何pending operation使gate失败。
- Authority provider必须与PostgreSQL backup位于独立failure/restore domain；真实PITR必须从revoke/disable前恢复并证明listener/signer关闭、旧node/operator cert与旧desired不可服务。
- External VM、NV handles、PITR数据库/恢复timeline、tracked/build/run roots、resume blob、临时cert/credential/state都使用 Plan 08 B10 已冻结的 closed `ResourceType` 与 typed inspector/deleter API写 authenticated ownership WAL，并按 exact ID+provider identity+creation/start token 清理；跨 reboot 的 WAL key only through sealed helper handle 恢复。任何 intent/actual/cleaned 缺口、path-only parent、全局枚举或 generic shell deleter 都阻止 nested/outer signature。provider/journal/archive digest record按policy保留且不得含secret，也不得伪装成可删除 resource。
- Completion create/revalidate 不信任 consolidation host wall clock。两个零参数native parent只进入 `internal/c12completionpublication` 的opaque facade；该 composite 的受保护 child wrapper才从已验证runner bootstrap/snapshot解析toolchain/runner-profile锁定的唯一 production provider profile token，并以受保护 credential/provider handle 构造本次专用 sealed `AuthenticatedTrustedTimeClient`。parent、command与caller都不能接收或实现该client，且没有profile/endpoint/transport flag或环境覆盖。该 client 以本次 canonical input challenge取得 fresh signed `C12TrustedTimeEvidenceV1`，验证 source identity、signature、challenge、UTC、monotonic sequence/floor、issued/validity window 和 rollback resistance。缺失、重放、回拨、source identity drift、caller endpoint/transport override、package global/context ambient client 或用文件 mtime/本机时间替代都拒绝。
- Attested channel sender 只可对其 profile 固定 slot 调用 `Prepare`、`Commit`、`InspectSender(nonce)`；receiver 不能枚举或调用 sender Inspect。Completion 必须先持有 authenticated outer `publish_pending` transaction capsule，再通过 Plan 08 opaque `EvidenceAdoptionSet` 调用 `BeginAdoptionSet`/`RecoverAdoptionSet`。`AdoptCommitted(set,kind,expectedSlot,expectedSequence)` 与 `Acknowledge(set,lease)` 绑定同一 set 及其 durable actual；`RecoverAcknowledge(set)` 则不接受 kind、slot、sequence、nonce、digest 或 lease，而是只读取该 set 唯一 authenticated pending ACK intent 及 matching adoption actual。每次 adoption 前先 durable intent，返回 nonce/envelope 后先 durable actual 才可进入下一个 binding；任何 crash gap 只可由 `RecoverAdoptionSet` exact replay，不能重新 Begin、枚举或猜 nonce。Platform/authority sender 不得读取、采用或比较 Windows/彼此 transaction；三方 shared receipt/bundle equality 只在 completion adoption 后验证。
- Plan 08 拥有 cleanup-only capsule、leaf finalization receipt 与 cleaned-tombstone state machine。Windows/platform/authority 的每条路径都必须按唯一顺序执行：每个资源 exact absent → terminal WAL file+directory fsync → leaf `PrepareFinalization` → strict tombstone stage/file+directory fsync + first guarded CAS/reread into origin-specific pending carrying the complete receipt and sole fixed target → leaf Recover/Commit → destroy/verify the exact WAL key → **while pending remains selected** unlink+parent-directory-fsync and verify the tombstone absent → final guarded CAS/reread into that target。target CAS后没有filesystem retirement。普通及 abort-origin cleanup 的 target 是 `terminal-inactive/next generation`；Windows/platform/authority success-origin 的 target 分别是同 generation 的 `windows_post_cleanup_publication|platform_post_cleanup_publication|authority_post_cleanup_publication`，只有该 successor 才能取得 nested/outer/channel one-shot attestors。它们在两份 evidence 与 exact transaction `Prepare→Commit→Inspect` 全部收敛后才 final-CAS 到 terminal-inactive/next；cleanup failure、abort 或 active crash 无签名入口。Channel attestor 在 `Prepare` 前的一次签名中被消费并销毁，之后 sender 只保留 nonce 与 immutable signed envelope 用于 `Commit`/`InspectSender`。Completion 只有一个逻辑 fixed `C12CompletionRoleStateV1` recovery-slot record：其 `publish_pending` variant 持有 strict eight-file root、manifest temp/final、`completed` marker、adoption/ACK reservations与恢复权限；snapshot/provider/private-child/session/WAL/run-root 的 temporary cleanup authority 是同一 record 内、在 create `publish_pending` 与 revalidate `retained_complete_active` 下都严格闭合为 `none|protected_key_intent|inner_bootstrap_restart_pending|cleanup_capsule|finalization_pending` 的 least-authority **inner substate**，不是第二文件/slot，retained publish objects永不进入 inner absent set。它由Plan 08每role独立的rollback-resistant `(StateEpoch,current digest)` guard和固定A/B cells防止旧合法状态回放；guard不含phase/run/launch/next/publication/payload。只有 manifest durable 且三个 set-bound ACK 均 durable 后，一个逻辑终态事务才以 `E→E1 finalize_ready` 和 proof-bound `E1→E2 retained_complete` 两次 guard advance 完成；Prepare 不构造 terminal cell，E2 才是唯一终态 linearization 并把 matching execution reservation 标为 `exiting`。该状态不再具有 mutation、adoption或ACK权限，`completion-revalidate` 只能读取 immutable payload并在其active launch envelope内维护自己的temporary inner cleanup substate。
- Completion inner substate 的闭合 union 在 create `publish_pending` 与 revalidate `retained_complete_active` 下都恰为 `none|protected_key_intent|inner_bootstrap_restart_pending|cleanup_capsule|finalization_pending`；每次变化都更新 full `StateDigest`/`PreviousStateDigest`，但只含 temporary cleanup authority，必须保持 outer `PublishBindingDigest` 逐字不变。`none` 本身不是恢复结论。Coordinator 必须同时认证 owner-specific outer progress 与 adoption-set pending transition并返回闭合五路 `begin_required|session_validation_recovery|session_cleanup_recovery|finalization_recovery|complete`：无 set pending 的 genuine pre-manifest `none` 才可 Begin；pre-manifest selected state只能返回sealed validation-required session；create capsule一旦有固定set-first pending或outer已选`manifest_intent|manifest_actual`，只能先无filesystem地收敛该barrier并返回sealed cleanup-only session，禁止provider/child/validator/Stage；pending finalization仅走no-session恢复；只有create的`manifest_intent`-or-later加fixed-target-completed `none`才是actionless complete。pre-capsule selected-state recovery由 coordinator 内部 intent→destroy→restart-pending→replacement-intent→capsule 状态机收敛，最终只返回非零 inner handle；首个 inner CAS 前的 no-authority candidate 则只由 `begin_required` 的重检 Begin target 收敛，不向 caller 暴露分支或允许凭 error 调 Begin。
- Plan 07 的 protected verifier result-slot 身份、打开与一次消费语义仍由 Plan 07/08 消费边界唯一所有；P09 不引入 result path、path digest、caller-selected slot 或第二种 verifier-result locator。
- 本页出现的 receiver/binding/set/lease/sink/view 低层调用全部只描述 `internal/c12completionpublication/operator.go` 的私有实现；public parent、command 与 runner helper 只持有本页冻结的 opaque composite handle/operator/inner-session，不能命名、接收或调用这些低层能力。
- Threat model/runbook不能要求直接修改数据库、运行远程shell、降低host policy或绕过provider；root/kernel/hypervisor/deployment主体失陷时唯一支持恢复是外部隔离、重建和reenrollment。
- Xray MPL-2.0、sing-box GPLv3+、三个outer image、verifier test-assets及authority-operations auxiliary binary/image各自的SBOM/source/provenance/vulnerability/license/notice/secret-scan obligation必须处于通过状态；辅助分类不构成第四release或许可证豁免，Redis 8.8.1 production block不得被C1.2完成绕过。
- Roadmap只在真实四scope manifest验证成功、文档门通过、license未绕过、cleanup成功且worktree等于baseline后更新。
- 每个 task 按 RED/GREEN/REFACTOR 独立提交；不得 stage 或删除用户未跟踪目录。

---

## File structure and frozen B11 evidence

```text
docs/security/
├── c12-threat-model.md
└── c12-completion-record.md
docs/runbooks/
├── c12-identity-trust-authority.md
├── c12-node-state-quarantine.md
└── c12-host-isolation-core-release.md
internal/c12evidence/
├── docs_test.go
├── platform.go
├── platform_test.go
├── authority.go
├── authority_test.go
├── completion.go
└── completion_test.go
internal/nodecontrol/hostevidence/operator_guard_conformance_test.go
internal/nodecontrol/authority/production_conformance_test.go
deploy/c12-authority/authority-operations.Dockerfile
docs/licenses/c12-authority-operations.md
internal/nodecontrol/operations/
├── orchestrator.go                 # sole production entry and phase persistence
├── activation.go                   # runtime registration and five-stage genesis
├── epoch_transition.go             # normal resolve/cancel terminal protocol
├── epoch_recovery.go               # terminal/suffix/prefix/Fence/catchup orchestration
├── staging.go                      # capability/import/recovery/revocation/Release/Abort
├── source_down.go                  # source exact-cover/archive and disposable Down
├── orchestrator_test.go
├── activation_test.go
├── epoch_transition_test.go
├── epoch_recovery_test.go
├── staging_test.go
└── source_down_test.go
cmd/nodecontrol-authority-operations/
├── main.go
└── main_test.go
scripts/
├── verify-c12-platform.sh
├── verify-c12-authority-v7-operations.sh
└── verify-c12-authority-fence.sh
testdata/c12/diagnostic/runner-profile.v1.json
testdata/c12/platform/runner-profile.v1.json
testdata/c12/platform/trusted-time-provider-profile.v1.json
testdata/c12/authority/runner-profile.v1.json
testdata/c12/authority/operations-build-policy.v1.json
testdata/c12/completion/runner-profile.v1.json
testdata/c12/artifact-catalog.v1.json
docs/roadmap/implementation-sequence.md
```

The production evidence boundaries are:

```go
package c12evidence

type PlatformEvidenceV1 struct {
	SchemaVersion              string
	RunID                      uuid.UUID
	StartedAt                  time.Time
	FinishedAt                 time.Time
	ExpiresAt                  time.Time
	TrustedTimeSourceIdentity  contracts.Digest
	TrustedTimeAttestationDigest contracts.Digest
	RunnerAttestation          []byte
	RunnerIdentity             contracts.Digest
	ProviderIdentitySet        contracts.Digest
	PCRMeasurementPolicy       contracts.Digest
	NegativeVMIdentity         contracts.Digest
	PositiveVMIdentity         contracts.Digest
	PhaseABootID               string
	PhaseBBootID               string
	CounterHandleSetDigest     contracts.Digest
	CounterFloorBefore         [3]uint64
	CounterFloorAfter          [3]uint64
	TrustedTimeFloorBefore     time.Time
	TrustedTimeFloorAfter      time.Time
	HostMemoryPolicyDigest     contracts.Digest
	KernelDigest               contracts.Digest
	SELinuxPolicyDigest        contracts.Digest
	SELinuxDomainMapDigest     contracts.Digest
	SysctlReadbackDigest       contracts.Digest
	CollectorMaskDigest        contracts.Digest
	ReplayResultDigest         contracts.Digest
	ExternalAttackerDigest     contracts.Digest
	PostStartReceiptSetDigest  contracts.Digest
	OperatorGuardResultDigest  contracts.Digest
	RepoCommit                 string
	TrackedTreeDigest          contracts.Digest
	SpecDigest                 contracts.Digest
	ToolchainDigest            contracts.Digest
	ReleaseDigests             [3]contracts.Digest
	OwnershipWALReservationDigest contracts.Digest
	OwnershipWALDigest         contracts.Digest
	BuildRootIdentitySetDigest contracts.Digest
	CleanupPlanDigest          contracts.Digest
	CleanupOnlyOwnershipCapsuleDigest contracts.Digest
	RecoveryAuthoritySlotIdentityDigest contracts.Digest
	AuthorityOperationsArtifact AuthorityOperationsArtifactTupleV1
	AuthorityOperationsBinaryAndImageDigest contracts.Digest
	AuthorityOperationsSupplyChainRecordSetDigest contracts.Digest
	SupplyChainEvidenceBundleDigest contracts.Digest
	ArtifactCatalogDigest      contracts.Digest
	ExpectedProcessImagePolicyDigest contracts.Digest
	ImageSetDigest             contracts.Digest
	ArtifactScanDigest         contracts.Digest
	CleanupResultDigest        contracts.Digest
}

type PlatformTrustPolicy struct {
	ScopePolicy                         TrustPolicy
	ExpectedTrustedTimeSourceIdentity   contracts.Digest
	ExpectedRunnerIdentity              contracts.Digest
	ExpectedProviderIdentitySet         contracts.Digest
	ExpectedPCRMeasurementPolicy        contracts.Digest
	ExpectedNegativeVMIdentity          contracts.Digest
	ExpectedPositiveVMIdentity          contracts.Digest
	ExpectedCounterHandleSetDigest      contracts.Digest
	ExpectedHostMemoryPolicyDigest      contracts.Digest
	ExpectedKernelDigest                contracts.Digest
	ExpectedSELinuxPolicyDigest         contracts.Digest
	ExpectedSELinuxDomainMapDigest      contracts.Digest
	ExpectedSysctlReadbackDigest        contracts.Digest
	ExpectedCollectorMaskDigest         contracts.Digest
	ExpectedReplayResultDigest          contracts.Digest
	ExpectedExternalAttackerDigest      contracts.Digest
	ExpectedPostStartReceiptSetDigest   contracts.Digest
	ExpectedOperatorGuardResultDigest   contracts.Digest
	ExpectedOwnershipWALReservationDigest contracts.Digest
	ExpectedOwnershipWALDigest          contracts.Digest
	ExpectedBuildRootIdentitySetDigest  contracts.Digest
	ExpectedCleanupPlanDigest           contracts.Digest
	ExpectedCleanupOnlyOwnershipCapsuleDigest contracts.Digest
	ExpectedRecoveryAuthoritySlotIdentityDigest contracts.Digest
	ExpectedCleanupResultDigest         contracts.Digest
}

type AuthorityFenceEvidenceV1 struct {
	SchemaVersion                    string
	RunID                            uuid.UUID
	StartedAt                        time.Time
	FinishedAt                       time.Time
	ExpiresAt                        time.Time
	TrustedTimeSourceIdentity        contracts.Digest
	TrustedTimeAttestationDigest     contracts.Digest
	ProviderIdentityPolicyVersion    string
	ConformanceTenant                string
	ControlAPIBinaryAndImageDigest   contracts.Digest
	AuthorityOperationsArtifact      AuthorityOperationsArtifactTupleV1
	AuthorityOperationsBinaryAndImageDigest contracts.Digest
	AuthorityOperationsSupplyChainRecordSetDigest contracts.Digest
	SupplyChainEvidenceBundleDigest contracts.Digest
	ArtifactCatalogDigest             contracts.Digest
	ExpectedProcessImagePolicyDigest  contracts.Digest
	ImageSetDigest                    contracts.Digest
	ArtifactScanDigest                contracts.Digest
	AuthorityClientProtocolVersion   string
	AuthorityRulesetDigest           contracts.Digest
	DBSystemTimelineDigest           contracts.Digest
	DBSchemaOrderedMigrationDigest   contracts.Digest
	PITRBackupRestoreDigest          contracts.Digest
	RepoCommit                       string
	TrackedTreeDigest                contracts.Digest
	SpecDigest                       contracts.Digest
	ToolchainDigest                  contracts.Digest
	ReleaseDigests                   [3]contracts.Digest
	AuthorityProtocolV7ResultDigest  contracts.Digest
	LegacySourceArchiveResultDigest  contracts.Digest
	FreshStagingRecoveryResultDigest contracts.Digest
	DisposableDownResultDigest       contracts.Digest
	DestructiveRestoreBlockDigest    contracts.Digest
	LegacyUpgradeStatus              string
	TestCaseResultDigest             contracts.Digest
	TerminalProviderRecordDigest     contracts.Digest
	OwnershipWALReservationDigest    contracts.Digest
	OwnershipWALDigest               contracts.Digest
	BuildRootIdentitySetDigest       contracts.Digest
	CleanupPlanDigest                contracts.Digest
	CleanupOnlyOwnershipCapsuleDigest contracts.Digest
	RecoveryAuthoritySlotIdentityDigest contracts.Digest
	CleanupResultDigest              contracts.Digest
	RunnerAttestation                []byte
}

type AuthorityTrustPolicy struct {
	ScopePolicy                              TrustPolicy
	ExpectedTrustedTimeSourceIdentity        contracts.Digest
	ExpectedProviderIdentityPolicyVersion    string
	ExpectedConformanceTenant                string
	ExpectedControlAPIBinaryAndImageDigest   contracts.Digest
	ExpectedAuthorityClientProtocolVersion   string
	ExpectedAuthorityRulesetDigest           contracts.Digest
	ExpectedDBSystemTimelineDigest           contracts.Digest
	ExpectedDBSchemaOrderedMigrationDigest   contracts.Digest
	ExpectedPITRBackupRestoreDigest          contracts.Digest
	ExpectedAuthorityProtocolV7ResultDigest  contracts.Digest
	ExpectedLegacySourceArchiveResultDigest  contracts.Digest
	ExpectedFreshStagingRecoveryResultDigest contracts.Digest
	ExpectedDisposableDownResultDigest       contracts.Digest
	ExpectedDestructiveRestoreBlockDigest    contracts.Digest
	ExpectedLegacyUpgradeStatus              string
	ExpectedTestCaseResultDigest              contracts.Digest
	ExpectedTerminalProviderRecordDigest      contracts.Digest
	ExpectedOwnershipWALReservationDigest     contracts.Digest
	ExpectedOwnershipWALDigest                contracts.Digest
	ExpectedBuildRootIdentitySetDigest        contracts.Digest
	ExpectedCleanupPlanDigest                 contracts.Digest
	ExpectedCleanupOnlyOwnershipCapsuleDigest contracts.Digest
	ExpectedRecoveryAuthoritySlotIdentityDigest contracts.Digest
	ExpectedCleanupResultDigest               contracts.Digest
}

func CanonicalPlatformEvidenceUnsignedPayload(PlatformEvidenceV1) ([]byte, error)
func CanonicalPlatformEvidence(PlatformEvidenceV1) ([]byte, contracts.Digest, error)
func ValidatePlatformEvidence(PlatformEvidenceV1, PlatformTrustPolicy) error
func CanonicalAuthorityFenceEvidenceUnsignedPayload(AuthorityFenceEvidenceV1) ([]byte, error)
func CanonicalAuthorityFenceEvidence(AuthorityFenceEvidenceV1) ([]byte, contracts.Digest, error)
func ValidateAuthorityFenceEvidence(AuthorityFenceEvidenceV1, AuthorityTrustPolicy) error

type C12TrustedTimeEvidenceV1 struct {
	SchemaVersion  string
	SourceIdentity contracts.Digest
	Purpose        string
	ChallengeDigest contracts.Digest
	IssuedAt       time.Time
	ValidUntil     time.Time
	Sequence       uint64
	Floor          time.Time
	Attestation    []byte
}

type RetainedTrustedTimePolicy struct {
	ExpectedSourceIdentity contracts.Digest
	ExpectedPurpose        string
	ExpectedChallengeDigest contracts.Digest
	MinimumSequence       uint64
	MinimumFloor          time.Time
}

type FreshTrustedTimePolicy struct {
	Retained              RetainedTrustedTimePolicy
	MaximumRoundTrip      time.Duration
}

type TrustedTimeNativeClientBuildV1 struct {
	PurposeClass     string // platform_linux_amd64 | authority_linux_amd64 | completion_windows_amd64
	OS               string
	Arch             string
	CatalogRole      string
	ExecutableDigest contracts.Digest
}

type TrustedTimeProviderProfileV1 struct {
	SchemaVersion         string
	SourceIdentity        contracts.Digest
	Endpoint              string
	ServerSPKIDigest      contracts.Digest
	TransportPolicyDigest contracts.Digest
	ClientBuildDigest     contracts.Digest // OS-neutral trusted-time protocol/source build
	NativeClientBuilds    [3]TrustedTimeNativeClientBuildV1
}

type ValidatedTrustedTimeProviderProfile interface {
	validatedTrustedTimeProviderProfile() // sealed nonzero profile/source/transport/native-client binding
}
type ValidatedTrustedTimeNativeClientIdentity interface {
	validatedTrustedTimeNativeClientIdentity() // minted only by catalog/toolchain-locked native parent
}
type ProtectedTrustedTimeProviderHandle interface {
	protectedTrustedTimeProviderHandle() // sealed to native-helper-owned implementations and package tests
}

type TrustedTimeClient interface {
	FetchAndValidateFresh(context.Context, FreshTrustedTimePolicy) (C12TrustedTimeEvidenceV1, time.Time, error)
}

type AuthenticatedTrustedTimeClient interface {
	TrustedTimeClient
	authenticatedTrustedTimeClient() // sealed production provenance; package-local fakes only
}

type InheritedCompletionTrustedTimeRuntime interface {
	inheritedCompletionTrustedTimeRuntime() // sealed one-use child-only profile/provider/client provenance
}

func LoadTrustedTimeProviderProfile(string, contracts.Digest, ValidatedTrustedTimeNativeClientIdentity) (ValidatedTrustedTimeProviderProfile, error)
func NewTrustedTimeClient(ValidatedTrustedTimeProviderProfile, ProtectedTrustedTimeProviderHandle) (AuthenticatedTrustedTimeClient, error)
func BindPlatformPhaseATrustedTimeClient(context.Context, *c12runnerprofile.AuthenticatedNativeRunnerSession) (AuthenticatedTrustedTimeClient, error)
func BindAuthorityTrustedTimeClient(context.Context, *c12runnerprofile.AuthenticatedNativeRunnerSession) (AuthenticatedTrustedTimeClient, error)
func OpenInheritedCompletionTrustedTimeRuntime(context.Context) (InheritedCompletionTrustedTimeRuntime, error)
func ValidateRetainedTrustedTimeEvidence(C12TrustedTimeEvidenceV1, RetainedTrustedTimePolicy) (time.Time, error)
func CanonicalTrustedTimeEvidenceUnsignedPayload(C12TrustedTimeEvidenceV1) ([]byte, error)
func CanonicalTrustedTimeEvidence(C12TrustedTimeEvidenceV1) ([]byte, contracts.Digest, error)
```

Each Platform/Authority `Canonical*UnsignedPayload` is the RFC 8785 JCS of one flat closed projection containing **exactly every serialized member of its evidence type except `RunnerAttestation`**, preserving field names/types and allowing no `omitempty`、default injection、nested wrapper or caller-selected exclusion. Platform signs `TALENRO-C12-PLATFORM-EVIDENCE-V1\x00 || CanonicalPlatformEvidenceUnsignedPayload(evidence)`；authority signs `TALENRO-AUTHORITY-FENCE-EVIDENCE-V1\x00 || CanonicalAuthorityFenceEvidenceUnsignedPayload(evidence)`. The corresponding `Canonical*Evidence` is never a signing input：after strict decode and signature verification it encodes the complete flat envelope including `RunnerAttestation`, with digest domain `TALENRO-C12-PLATFORM-EVIDENCE-ENVELOPE-V1` or `TALENRO-AUTHORITY-FENCE-EVIDENCE-ENVELOPE-V1` followed by `0x00 || canonical_envelope`.

`C12TrustedTimeEvidenceV1` is owned only by Task 4 and has exact schema `talenro-c12-trusted-time-evidence/v1`. Its unsigned projection excludes only `Attestation` and is signed over `TALENRO-C12-TRUSTED-TIME-EVIDENCE-V1\x00 || JCS(unsigned)` by the toolchain/runner-profile locked production source identity；the complete envelope digest uses `TALENRO-C12-TRUSTED-TIME-EVIDENCE-ENVELOPE-V1\x00`. `Purpose` is closed to `platform_phase_a`、`platform_phase_b`、`authority_evidence`、`completion_create` or `completion_revalidate`, so the cross-reboot gate cannot be forced through a completion-only challenge. Strict validation requires nonzero identity/challenge/signature, canonical UTC source claims `Floor <= IssuedAt <= ValidUntil <= IssuedAt+5m`, nonzero monotonically increasing `Sequence`, exact expected source/purpose and exact challenge `SHA-256("TALENRO-C12-TRUSTED-TIME-CHALLENGE-V1" || 0x00 || purpose || canonical_input_digest || fresh_command_nonce)`. Each purpose has a closed canonical input projection owned by its caller；for platform this includes the sealed resume-manifest digest/run/phase/previous sequence and for authority it includes the run/nested-evidence input digest, while completion uses the fixed eight-file/projection digest. The source call is online and nonce-bound；after signature/challenge verification the command uses `IssuedAt` as its sole trusted current time and never consults the host clock.

Task 4 exposes exactly two evidence-validation entry points. `ValidateRetainedTrustedTimeEvidence` revalidates a retained historical envelope's strict schema、locked source signature、purpose/challenge、canonical UTC interval、sequence/floor and canonical envelope digest without requiring or pretending to reconstruct an old transport round-trip proof. `TrustedTimeClient.FetchAndValidateFresh` owns the online source call and its unexported same-call monotonic start/end proof, applies that retained validator plus the closed maximum-round-trip rule, and returns the validated envelope and its `IssuedAt`; no lower-level fetch、round-trip proof constructor or alternate validator is public. `TrustedTimeProviderProfileV1` has exact schema `talenro-c12-trusted-time-provider-profile/v1`. Its scalar `ClientBuildDigest` is the OS-neutral digest of the exact trusted-time protocol/client source+recipe, never a PE/ELF executable hash；the closed ordered `NativeClientBuilds` separately maps platform Linux helper、authority Linux helper and Windows completion child purpose classes to exact OS/arch/catalog role/executable digest. `LoadTrustedTimeProviderProfile` accepts only the fixed tracked profile path plus its expected toolchain/runner-profile digest and an opaque native identity minted after the parent validates its actual executable against the matching catalog tuple；it strict-decodes all fields, confines the HTTPS endpoint to the locked origin and constant-time verifies source-key、server-SPKI、TLS/transport、OS-neutral build and exact purpose/OS/arch executable projections before returning the opaque profile. Production obtains a client only from `NewTrustedTimeClient` with that opaque profile plus a non-exportable native-helper-delivered `ProtectedTrustedTimeProviderHandle` for the matching credential/provider session. Neither `FreshTrustedTimePolicy` nor `CompletionPolicy` carries an endpoint、credential or transport, and no package global、context value、environment fallback or caller HTTP client may replace the constructed transport. Task 4 tests use a fake implementing this same interface；production commands reject a fake/unvalidated profile/client provenance、wrong purpose/OS/arch tuple、PE-as-ELF substitution or executable digest splice before a network call. Create uses the client method. Revalidation first uses the retained function on the embedded creation statement, then uses the same profile-constructed client method for a new challenge with greater sequence and nondecreasing floor. Task 7 may only compose these Task 4 APIs and must not implement a second signature/challenge/time validator. No caller-provided scalar time or replayable standalone time file enters any expiry check.

`TrustedTimeClient` remains the package's low-level behavior abstraction, but every production constructor and completion runtime consumes only sealed `AuthenticatedTrustedTimeClient`. Only `internal/c12evidence` package tests may implement that sealed subtype；cross-package tests reach it through the package-owned fake-provider bridge. The completion composite exposes no client parameter and never promotes caller-defined behavior to authenticated provenance.

`ValidatedTrustedTimeProviderProfile` is likewise a sealed interface, not a zero-size public value. `LoadTrustedTimeProviderProfile` is its sole production mint and binds the validated canonical profile、source/SPKI/transport、toolchain/runner profile and exact native-client identity；nil、zero、caller implementation or cross-profile reuse fails before `NewTrustedTimeClient` accesses a provider handle.

`BindPlatformPhaseATrustedTimeClient` and `BindAuthorityTrustedTimeClient` are the only exported production adapters for the non-completion fresh native roles. Each accepts only the authenticated Plan 08 session and hard-codes exactly one role/purpose (`platform→platform_phase_a` or `authority→authority_evidence`)；neither accepts a purpose selector, and the P08 session exposes no caller-readable role scalar. Inside `internal/c12evidence/trustedtime.go` each binds and Inspect-validates the raw runtime, privately wraps its role/profile/native-client/provider/session projections in the package's unexported native-identity/provider implementations, and invokes the same `LoadTrustedTimeProviderProfile`/`NewTrustedTimeClient` core. The returned client is one-use and no caller can extract a raw runtime/result. Both accept no path、profile digest、endpoint、provider handle or transport option. Phase B instead uses an unexported platform-facade helper over the pathlessly recovered successor；it recovers a durably recorded result through Plan 08 after response loss and never refetches. Original and recovered results must carry the same sealed native monotonic elapsed duration and measurement binding；only that durable record plus fixed `MaximumRoundTrip` may mint the private fresh-round-trip proof, while an absent/zero/negative/overflowed/spliced measurement downgrades to cleanup before promotion.

`OpenInheritedCompletionTrustedTimeRuntime` is the sole production mint route for the completion child's otherwise package-private `ValidatedTrustedTimeNativeClientIdentity` and `ProtectedTrustedTimeProviderHandle` implementations. It accepts no input and internally calls Plan 08's zero-argument child-session Open、Inspect and bounded evidence/tree-read bridges, verifies the fixed completion purpose/profile/client-build/install/tree projections plus exact eight-member/optional-retained-manifest binding, wraps that sealed session in c12evidence-owned private identity/transport types, and then calls the same `LoadTrustedTimeProviderProfile` and `NewTrustedTimeClient` implementations used by the other native roles. It returns only one sealed runtime；the child `main` package cannot implement either private marker or supply a handle/path/profile/policy. Task 4 deliberately does not implement the future manifest-aware child runner or its DTO: `RunInheritedCompletionValidationChild`、`CompletionValidationChildResponseV1` and their canonical codec are Task 7 `completion.go` owners, after `C12CompletionManifestV1` and the retained semantic validator exist. Task 4 tests `TestOpenInheritedCompletionTrustedTimeRuntimeIsChildOnlyAndFixed` and `TestInheritedCompletionTrustedTimeRuntimeMintsSealedProvenanceAndConsumesOnce` compile only this lower provenance bridge; Task 7 adds the policy-binding/run/response tests against the real manifest schema. Direct/go-run/unbound launch、a changed session/tree/profile binding or alternate client rejects before network access.

Both nested envelopes carry explicit UTC `StartedAt`/`FinishedAt`/`ExpiresAt`, locked trusted-time source identity and signed time-statement digest, plus the exact `RepoCommit`, canonical 64-hex `TrackedTreeDigest`, three-member `SpecDigest`, `ToolchainDigest` and ordered `ReleaseDigests[3]`；no opaque combined repo/tree or release/toolchain digest may substitute. `AuthorityOperationsArtifact`、the two `AuthorityOperations*Digest` members、`SupplyChainEvidenceBundleDigest`、`ArtifactCatalogDigest` and `ArtifactScanDigest` are the Authority Protocol v7 additive refinement of the base-design §18 evidence boundary under the new three-member `SpecDigest`. The tuple exposes the exact binary SHA、OCI manifest、build receipt and dependency closure needed for immutable execution；the combined digest is distinct from and cannot alias `ControlAPIBinaryAndImageDigest`；the supply-chain digest binds the subject-checked closed record references while the bundle digest preserves their bounded canonical auditable bodies；the catalog digest proves exact input-role coverage；the full-scan digest binds the entire canonical deterministic receipt including ruleset, all releases/images/assets and expected-process policy. Actual run-specific process receipts remain in run evidence. An old V1 payload lacking any one is rejected by strict decoding/current-policy validation rather than upgraded in place. `ReleaseDigests` remains exactly three.

Plan 09 consumes the Plan 08 native-runner contract without moving machine facts into Git. Each tracked policy instance has schema `talenro-c12-native-runner-profile/v1` and the Plan 08-owned closed projection `SchemaVersion`、`Role`、ordered `AllowedTrackedScriptTokens`、ordered `RequiredToolRoles`、`HelperCatalogRole`、`BootstrapLauncherCatalogRole`、`RecoverySlotPurpose`、`ExternalRootPurpose`、ordered `ChannelBindings []C12NativeRunnerChannelBindingV1`、`ProviderCredentialPurpose`、`EvidenceRootPurpose` and `InstallReceiptPolicy`. `BootstrapLauncherCatalogRole` is exactly `c12_runner_launcher_windows_amd64` for Windows roles and `c12_runner_launcher_linux_amd64` for Linux roles, is never a helper alias/receipt role and has no machine path in Git. Each binding has exactly `TransactionKind`、`Access`、`SlotPurpose` and `SequencePolicy`；`SequencePolicy` is always `next`. Windows/platform/authority each declare exactly one sender binding, completion declares exactly three receiver bindings in order `windows_scopes`、`platform_scope`、`authority_scope`, and `staged_diagnostic` declares zero bindings. The Plan 08-owned diagnostic instance at fixed token `testdata/c12/diagnostic/runner-profile.v1.json` also fixes empty scripts, the exact Linux static-helper/locked-Git/reproducible-build tool roles, `ProviderCredentialPurpose=none` and `EvidenceRootPurpose=nonpublishing`; it has no provider/channel/result/scope signing or publishing capability. Every tracked instance contains no absolute machine path、object-database/root/slot identity、installed executable digest、launcher attestation、transaction nonce、secret or evidence/output digest. Unknown、duplicate、null、default、missing/wrong-OS/helper-aliased launcher role、noncanonical token order、another role's script/tool/channel/provider purpose or role/token mismatch rejects. Protected provisioning calls consuming `InstallFixedNativeRunnerProfile`, which destroys its installer session/candidate handle before every return, followed only by non-consuming `VerifyFixedNativeRunnerInstallation(context.Context, NativeRunnerRole)`；only the launcher-selected actual helper invokes `OpenInstalledNativeRunnerSession(context.Context, NativeRunnerRole)` exactly once for a fresh generation. A crash restart first proves the old bound containment absent and uses only its sealed pre-terminal same-generation successor；a terminal create result instead requires old-reservation clear plus the separate read-only audit. There is no profile/store/path/options overload.

The selected private install record carries the closed Plan 08 variant `b10_pre_catalog|b11_final_catalog`. Non-consuming Verify dispatches solely on that authenticated record：B10 requires catalog absent/zero with exact toolchain-native-role/profile/launcher projections；B11 requires the nonzero final catalog exact cover and relock commit. P09 supplies no variant option, and present-at-B10 or absent-at-B11 catalog state rejects.

The compile-time OS adapter opens only Plan 08's fixed store：`/var/lib/talenro/c12/runner-bootstrap.v1.sealed` on Linux or `HKLM\SOFTWARE\Talenro\C12\RunnerBootstrapV1` on Windows. That machine-sealed record—not the tracked profile—binds the canonical Git object-database locator+observed identity, owner-only external root, recovery slot, exact absolute helper/interpreter/tool/private-child `C12NativeRunnerInstallReceiptV1` set and role-dependent evidence-root/provider/attestor/channel capabilities plus profile/catalog/toolchain projections. Snapshot scripts remain tracked tokens only：after capture the parent derives each absolute path under its retained committed snapshot and requires exact committed blob bytes、Git mode、snapshot root/file identity、no-follow regular file and one link before spawn；no script receipt exists. Final-role channel capabilities expose only nonzero `ChannelNamespaceDigest`、`ChannelSlotIdentityDigest` and `SlotSequence` for their tracked bindings；each sender gets exact one and completion gets exact three ordered bindings. A `staged_diagnostic` session instead returns only ODB、external-root、recovery-slot and exact Linux helper/Git/build-tool handles, with no script/evidence-root/provider/attestor/channel capability. `VerifyFixedNativeRunnerInstallation` strict-validates/reopens every installed receipt and all profile/catalog/toolchain projections without consuming a launch generation or returning capability；`OpenInstalledNativeRunnerSession` repeats those checks, atomically consumes exactly one launch generation and returns only opaque one-use typed handles to the actual public parent. Caller paths、raw receipt/bootstrap bytes、copied session fields、numeric handles、provision-time open/destroy or a second open of the same launch generation cannot initialize a parent. The Windows/diagnostic tracked instances and fixed stores are Plan 08-owned and are consumed, not recreated, below.

The fixed launcher is separately reproducibly built/scanned as `c12_runner_launcher_windows_amd64|c12_runner_launcher_linux_amd64`, frozen at P08 B10 and machine-base-attested before role installation. B11 and P09's artifact catalog revalidate those exact binary/SBOM/provenance identities byte-for-byte；they never upgrade or reinstall the launcher. Its protected execution reservation binds role/mode/install record、launcher/helper PID+start、containment and attested boot identity. Launcher death kills the exact tree；a successor must prove same-boot absence or an authenticated different boot before replacing the reservation. Completion create recovers only `active_unstarted|publish_pending`, revalidate recovers only `retained_complete_active`, and platform recovers only `resume_only_successor`, all without a second Open. E2/normal terminal business mutation marks the reservation exiting；only launcher wait/reap+empty proof followed by reservation-only guarded clear enables the next run or the distinct create-mode `terminal_audit` reservation.

Plan 09 consumes Plan 08's opaque `EvidenceAdoptionSet` and frozen acyclic digest graph without redefining its strict schemas、JCS projections or domains. Immutable `AdoptionSetIdentityDigest I` binds one completion receiver/run/generation/profile/set reservation and the exact three ordered expected slot/sequence capabilities, but no epoch、state/publish or mutable set-record digest. Every ordinary begin/adopt/publication/ACK semantic change derives `SetTransitionDigest T` only from `{I,previous T,source PublishBindingDigest,closed transition kind,canonical semantic delta}` and follows the exact one-way order `pending Sp → outer guard CAS/reread producing Pnew from T+delta → actual Sa carrying Sp+T+new epoch/full/Pnew/result`. T/Pnew/view exclude Sp/Sa and every mutable/full set-record digest. Recovery accepts exactly source-P with the stored delta not yet applied, or deterministic Pnew with the same selected I/T/delta/result already applied；an execution/ordinary-inner-only successor may rebase the current epoch/full anchor without changing P, and a present Sa remains immutable while an absent Sa binds the current selected anchor. `finalize_ready` instead follows its distinct terminal branch `source-only pending Fp → outer prepare CAS/reread producing E1/D1/P1 → source-only actual marker Fa`; Fp/Fa bind I/T、the source E/D/P and frozen E2-target/audit digests but contain neither post-prepare E1/D1/P1 nor a future E2/full-cell digest. Recovery authenticates/converges Fp/Fa and accepts only the same selected I/T/delta/outer business、inner `none` and P1；after launcher death it may rebase the proof across zero or more guard-selected execution-reservation-only successors to current E'/D'/P1, then Commit rechecks that current proof. Any inner or outer semantic drift rejects. `BeginAdoptionSet` starts the ordinary graph；`RecoverAdoptionSet` can reopen only its exact partially progressed transition. Every hook validates the current guard epoch/full digest and source publish digest. `RecoverAcknowledge(set)` accepts no binding selector and converges only the set's sole authenticated pending ACK intent plus matching adoption actual. All reject a foreign identity/set、reordered binding、self/new-P/set-record-digest preimage、unrepresented business drift、second Begin/adopt/ACK、nonce supplied before the matching actual or an outer-capsule digest mismatch.

### Task 1: Threat model with executable coverage contract

**Files:**
- Create: `docs/security/c12-threat-model.md`
- Create: `internal/c12evidence/docs_test.go`
- Test: `internal/c12evidence/docs_test.go`

**Interfaces:**
- Consumes: approved-spec threat list, C1.2 trust/authority/agent/supervisor/adapter/evidence boundaries and B08/B09 license records.
- Produces: one threat model whose scenario IDs are consumed by all three runbooks and final completion validator.

- [ ] **Step 1: RED — add required threat-ID and boundary tests**

```go
func TestThreatModelCoversApprovedBoundary(t *testing.T) {
	required := []string{
		"C12-T01 enrollment-grant-replay", "C12-T02 certificate-and-key-compromise",
		"C12-T03 listener-trust-confusion", "C12-T04 signed-state-replay-fork",
		"C12-T05 trusted-time-rollback", "C12-T06 server-ca-package-rollback",
		"C12-T07 operator-guard-replay", "C12-T08 authority-fence-pitr",
		"C12-T09 root-share-confusion", "C12-T10 release-config-toctou",
		"C12-T11 orphan-and-ownership", "C12-T12 core-sandbox-escape",
		"C12-T13 memory-perf-bpf-dump", "C12-T14 supervisor-protocol-bypass",
		"C12-T15 dual-latch-crash", "C12-T16 resource-exhaustion",
		"C12-T17 aggregate-oversubscription", "C12-T18 dependency-outage",
		"C12-T19 observation-cardinality", "C12-T20 core-supply-chain-license",
	}
	assertDocumentContainsEach(t, "../../docs/security/c12-threat-model.md", required)
}
```

- [ ] **Step 2: Run the document contract and verify RED**

Run: `go test ./internal/c12evidence -run '^TestThreatModel' -count=1`

Expected: FAIL because `docs/security/c12-threat-model.md` does not exist.

- [ ] **Step 3: GREEN — write assets, trust boundaries and attacker capabilities**

Define control-plane DB/fence, node/operator/server PKI, deployment/root/signer keys, agent/supervisor/latch state, core release/config/credentials, evidence and license assets. Draw explicit boundaries for public/bootstrap/agent/operator/metrics listeners, PostgreSQL versus external fence, agent versus root-owned supervisor, each core UID/cgroup, host versus nested Docker, and four evidence scopes.

- [ ] **Step 4: GREEN — write all 20 threat scenarios with fixed fields**

Each `C12-Tnn` section contains `Asset`, `Adversary capability`, `Entry point`, `Prevention`, `Detection`, `Fail-closed state`, `Recovery runbook`, `Residual risk`, and `Verification gate`. State explicitly that full root/Administrator/kernel/hypervisor/deployment compromise exceeds application isolation and requires external isolation/rebuild/reenrollment.

- [ ] **Step 5: Run the document contract and verify GREEN**

Run: `go test ./internal/c12evidence -run '^TestThreatModel' -count=1`

Expected: PASS; every scenario maps to a named test/gate and runbook, and the document contains no instruction to mutate DB directly, use remote shell or relax host security policy.

- [ ] **Step 6: REFACTOR — check links, finite IDs and diff**

Run: `go test ./internal/c12evidence -run '^TestThreatModel|^TestSecurityDocumentLinks' -count=1`

Run: `git diff --check -- docs/security/c12-threat-model.md internal/c12evidence/docs_test.go`

Expected: PASS.

- [ ] **Step 7: Commit the threat model**

```bash
git add docs/security/c12-threat-model.md internal/c12evidence/docs_test.go
git commit -m "docs: add C1.2 threat model"
```

### Task 2: Identity, trust and authority operations runbook

**Files:**
- Create: `docs/runbooks/c12-identity-trust-authority.md`
- Modify: `internal/c12evidence/docs_test.go`
- Test: `internal/c12evidence/docs_test.go`

**Interfaces:**
- Consumes: threat IDs C12-T01–T09, node/operator certificate APIs, root/metadata workflows, host-deployed trust packages, OperatorClientTrustGuard, B01 typed migration/database functions, B03 provider/attestor APIs and the sole B11 production orchestrator.
- Produces: bounded procedures for certificate compromise, signer/root/CA/trust-package/operator-guard/deployment-authority incidents, registered v7 migration/shutdown, epoch PITR recovery, legacy source archive, fresh staging/Abort, disposable Down and administrative restore.

- [ ] **Step 1: RED — add runbook scenario and section tests**

Require these exact procedure IDs: `C12-R01 node certificate compromise`, `R02 operator certificate compromise`, `R03 signer rotation`, `R04 root emergency ceremony`, `R05 CA issuer outage`, `R06 server CA rotation/removal`, `R07 operator guard outage/rollback`, `R08 deployment authority compromise`, `R09 fence outage/PITR`, `R10 administrative disable/restore`. Under R09 require exact registered subprocedures `C12-R09A production Up and classification`, `C12-R09B epoch terminal PITR recovery`, `C12-R09C legacy source retirement and archive`, `C12-R09D fresh target staging recovery and Abort`, `C12-R09E disposable registered Down`, and `C12-R09F destructive restore hard stop`.

- [ ] **Step 2: Run the runbook contract and verify RED**

Run: `go test ./internal/c12evidence -run '^TestIdentityTrustAuthorityRunbook$' -count=1`

Expected: FAIL because the runbook is absent.

- [ ] **Step 3: GREEN — write the ten bounded procedures**

Every procedure uses exact headings `Trigger signals`, `Required role and scope`, `Finite operator actions`, `Verification`, `Stop conditions`, `Escalation`, and `Evidence retained`. R09A–R09F may invoke only exact `nodecontrol-authority-operations` typed subcommands and its read-only Inspect; they never invoke ordinary Goose CLI, issue SQL, mutate a database/archive/journal row, run remote shell, reuse old certificates, accept a lower trust version or bypass uncached authorization. The runbook names the B01/B03 dependencies and requires every external call outside DB transactions except the one exact §9.3 `authority_protocol_downgrade_authorizer` call inside registered disposable Down.

- [ ] **Step 4: GREEN — encode the registered v7 upgrade/restore/Down procedures**

`C12-R09A` first proves LocalRuntimeIsolation/LegacyRuntimeShutdown, calls the registered migration executor, records the immutable production intent, and uses authenticated reservation before any provider/attestor call; dependency ambiguity produces IndeterminateSourceSeal/FreshRestoreRequirement and remains closed. `C12-R09B` lists the exact terminal→intent→suffix→commit-proof→provider-transcript→same-prefix arbitration→optional Fence→restricted rehydrate/application→fresh candidate proof→pre/post-catchup applied rebind→catchup/reconcile sequence and its Inspect-only restart points. It explicitly forbids current-transaction reconstruction of provider-consumed rows and direct archive writes.

`C12-R09C` requires permanent membership plus every environment/database/provider retirement, exact `LegacySourceRetirementSetV1`, post-seal observations/exact cover and archive evidence before a target admission; partial completion only exact-retries preallocated IDs. `C12-R09D` records capability/intent/recovery-application proofs and the Release/Abort branches, requires recovery revocation with its journal-bound `revocation_application_id`, and states that a missing fresh import application is never fenced or reimported. `C12-R09E` permits only disposable fixture Down after environment membership and every provider are permanently `down_retired`, then drives the B01 `1/0 -> 1/1 -> 0/0` transaction; production and raw Goose remain prohibited. `C12-R09F` stops at signed `FreshTargetInventoryV1`/`FreshV7StagingEvidenceV1` while the two approved effect kinds remain unsupported, retains `restore_incomplete`, and never reopens routes/listeners/signers.

`C12-R09` keeps listeners/signers closed after PITR until the exact recovery/catchup chain proves a nondecreasing provider/database head. `C12-R10` requires completed stopped reenrollment, fresh host evidence, two distinct uncached security-admin credentials within 15 minutes, same epoch/effect and a new higher desired generation.

- [ ] **Step 5: Run runbook tests and verify GREEN**

Run: `go test ./internal/c12evidence -run '^TestIdentityTrustAuthorityRunbook|^TestRunbooksForbidDirectMutation$' -count=1`

Expected: PASS; all required roles/actions/stops/evidence fields and R09A–R09F sequences exist；direct SQL/Goose/archive/journal mutation, any in-transaction provider/attestor or non-Down signer call, any second/wrong-phase Down-authorizer call, fresh-import Fence/reimport, production Down and `fresh_v7_staging_closed=upgrade_complete` patterns are absent.

- [ ] **Step 6: REFACTOR — validate threat backlinks and links**

Run: `go test ./internal/c12evidence -run '^TestRunbookThreatBacklinks|^TestSecurityDocumentLinks' -count=1`

Expected: PASS; every procedure references at least one C12 threat ID and a concrete observable/API.

- [ ] **Step 7: Commit the trust/authority runbook**

```bash
git add docs/runbooks/c12-identity-trust-authority.md internal/c12evidence/docs_test.go
git commit -m "docs: add C1.2 trust authority runbook"
```

### Task 3: Node quarantine, host isolation and external-core runbooks

**Files:**
- Create: `docs/runbooks/c12-node-state-quarantine.md`
- Create: `docs/runbooks/c12-host-isolation-core-release.md`
- Modify: `internal/c12evidence/docs_test.go`
- Test: `internal/c12evidence/docs_test.go`

**Interfaces:**
- Consumes: threat IDs C12-T10–T20, signed desired/recovery state, dual latch, resource envelope, manifest/map, host memory policy, B08/B09 license/release locks and ownership evidence.
- Produces: bounded node-state/supervisor recovery and host/core isolation/rebuild procedures.

- [ ] **Step 1: RED — add exact procedure coverage tests**

The node-state runbook must contain `C12-R11 desired rollback/conflict`, `R12 node quarantine/release`, `R13 control-plane outage/LKG expiry`, `R14 trusted-time loss/offline reboot`, `R15 agent-state corruption`, `R16 supervisor compromise`, `R17 pending fault/dual latch`, and `R18 resource-envelope rotation/oversubscription`.

The host/core runbook must contain `C12-R19 crash-dump/SELinux drift`, `R20 manifest/map rollback/TOCTOU`, `R21 external core provenance/CVE`, and `R22 privileged-host isolation/rebuild`.

- [ ] **Step 2: Run document contracts and verify RED**

Run: `go test ./internal/c12evidence -run '^TestNodeStateRunbook|^TestHostCoreRunbook' -count=1`

Expected: FAIL because both documents are absent.

- [ ] **Step 3: GREEN — write node-state and dual-latch procedures**

For rollback/fork, retain highest seen, stop/quarantine and require a higher signed generation. For LKG expiry/time loss/state corruption, make the node non-accepting and stop within the specified normal/security deadline. For supervisor fault, require stopped recovery snapshot, exact HostRemediationEvidence, supervisor latch clear, local latch clear, server attestation and explicit resume; one latch or one restart never authorizes core.

- [ ] **Step 4: GREEN — write host/core isolation and release procedures**

On SELinux/sysctl/collector drift or sandbox escape, stop/quarantine and request independent orchestration/network isolation; never lower policy. Manifest/map/TOCTOU and integrity incidents prohibit LKG. CVE without exploit evidence uses draining and reviewed replacement; suspected exploit uses security quarantine and five-second forced stop. Full privileged-host compromise requires rebuild/reenrollment, not local evidence repair.

- [ ] **Step 5: Run runbook tests and verify GREEN**

Run: `go test ./internal/c12evidence -run '^TestNodeStateRunbook|^TestHostCoreRunbook|^TestRunbooksForbidDirectMutation$' -count=1`

Expected: PASS with fixed signals, finite actions, verification, stop conditions and retained evidence for all 12 procedures.

- [ ] **Step 6: REFACTOR — verify license and external-process boundaries**

Run: `go test ./internal/c12evidence -run '^TestRunbookLicenseBoundary|^TestRunbookThreatBacklinks|^TestSecurityDocumentLinks' -count=1`

Expected: PASS; Xray/sing-box remain separate releases and process isolation is not presented as a license conclusion.

- [ ] **Step 7: Commit node/host/core operations docs**

```bash
git add docs/runbooks/c12-node-state-quarantine.md docs/runbooks/c12-host-isolation-core-release.md internal/c12evidence/docs_test.go
git commit -m "docs: add C1.2 node and host recovery runbooks"
```

### Task 4: Platform evidence and cross-reboot authenticated resume contract

**Files:**
- Create: `internal/c12evidence/trustedtime.go`
- Create: `internal/c12evidence/trustedtime_test.go`
- Create: `internal/c12evidence/platform.go`
- Create: `internal/c12evidence/platform_test.go`
- Create: `scripts/invoke-exact-platform-semantic-gate.ps1`
- Create: `testdata/c12/platform/runner-profile.v1.json`
- Create: `testdata/c12/platform/trusted-time-provider-profile.v1.json`
- Test: `internal/c12evidence/platform_test.go`

**Interfaces:**
- Consumes: `C12ScopeEvidenceV1`, B10 WAL, three production counter/sealer identities, locked production TrustedTimeSource client/identity, runner attestor and host policy package, plus Plan 08 `internal/c12runnerprofile` fixed-token loader、sealed-bootstrap opener、`C12NativeRunnerProfileV1` and `C12NativeRunnerInstallReceiptV1` validator. This task creates only the `platform` instance；it does not own a second profile schema/loader/session.
- Produces: sole `C12TrustedTimeEvidenceV1` schema/strict validator/canonical domains and source-client adapter, plus `PlatformEvidenceV1`, `CrossRebootResumeManifestV1`, `ValidatePlatformEvidence`, sink-bound `SealResumeManifest`, successor-only `OpenResumeManifest`, sealed workload/cleanup facades and the success-only post-clean publication facade. There is no bytes+policy cleanup opener. Task 7 only consumes this Task 4 API；it does not implement another clock/source validator.

- [ ] **Step 1: RED — add three-identity and pre-handle seal tests**

```go
func TestResumeManifestRejectsCounterIdentityReuse(t *testing.T) {
	t.Parallel()
	manifest := validResumeManifest()
	manifest.CounterIdentityDigests[1] = manifest.CounterIdentityDigests[0]
	builder := authenticatedPlatformResumeBuilder(t)
	if _, err := SealResumeManifest(testContext(), builder, manifest); !errors.Is(err, ErrIdentityReuse) {
		t.Fatal("resume manifest reused an anti-rollback identity")
	}
}

func TestOpenResumeManifestAuthenticatesBeforeHandleAccess(t *testing.T) {
	audit := installTamperedPlatformResumeRecoveryState(t)
	_, err := OpenResumeManifest(testContext())
	if err == nil || audit.NonTimeHandleAccesses() != 0 || audit.NativeSessionOpens() != 0 {
		t.Fatal("tampered resume manifest reached an NV handle")
	}
}
```

Add a table-driven RED matrix that mutates commit/Git-tree locator/canonical tracked-tree/spec/toolchain/release, each operations tuple component、combined digest、supply-chain record-set/bundle、catalog、process-image policy、full image set and `ArtifactScanDigest` independently in the guard-selected resume record, enclosing platform evidence and final scope policy. Every mutation must reject before NV-handle/build-root access；a valid tuple under another scope's process policy, a missing/extra runtime role and same operations tuple under a different full receipt also reject. Add zero/non-UTC/reversed timestamps, `ExpiresAt > StartedAt+72h`, Phase B after expiry and a fresh outer scope wrapping an expired nested platform envelope；every workload/resume case rejects before NV/WAL/root/resume-child access. Separately require pathless `OpenPlatformResumeCleanup(ctx)` after expiry or trusted-time-provider outage to obtain only Plan 08's fixed sealed cleanup successor internally, derive its local policy internally and expose only the high-level execute/finalize facade, never accepting or returning bytes、selector、policy、capability、proof、resume/sign/create/general WAL authority.

Add signature-projection RED cases proving `RunnerAttestation` is the sole excluded member：changing it leaves `CanonicalPlatformEvidenceUnsignedPayload` byte-identical but changes the complete envelope and fails verification；changing any other field changes the unsigned payload and invalidates the old signature. Signing the full envelope、omitting another member、nesting the payload or accepting a caller-selected projection must fail.

Add trusted-time RED vectors for wrong source/purpose/challenge, replayed command nonce, equal/decreasing sequence on revalidation, floor before manifest creation, non-UTC/reversed/>5m source interval, unknown/duplicate/null field, full-envelope signature and caller wall-clock scalar. Exercise both sole evidence-validation APIs：a retained creation envelope validates without an impossible historical round-trip token, while only the same profile-constructed client's same-call fresh fetch can satisfy the opaque internal monotonic timeout proof. The interface fake must be injectable in tests, while a production constructor rejects an untracked profile、endpoint/source-key/transport/client-build digest drift、wrong protected handle、caller transport、context/global override or missing credential before a network call. A fake host clock jump in either direction must not affect the accepted `IssuedAt`; a source timeout beyond the monotonic round-trip cap fails. Cross-reboot tests splice a separately valid Phase-B source identity/sequence/floor and require rejection before NV/WAL/root access.

The same matrix adds `TestOpenInheritedCompletionTrustedTimeRuntimeIsChildOnlyAndFixed`, `TestInheritedCompletionTrustedTimeRuntimeMintsSealedProvenanceAndConsumesOnce` and `TestInheritedCompletionTrustedTimeRuntimeRejectsPolicyBindingSplice` against Plan 08's real inherited child-session API. They prove an external `main` package cannot implement the private identity/provider markers, while the c12evidence-owned adapter can mint them only from the authenticated fixed child binding and can perform exactly one matching fetch.

Add platform-runner-profile RED vectors through Plan 08's generic loader: wrong schema/role, unknown/duplicate/null/default member, missing/extra/reordered tracked script token or tool role, wrong helper catalog role/recovery/external/evidence-root/provider purpose or install-receipt policy, and a channel policy not exactly one `{TransactionKind: platform_scope, Access: sender, SlotPurpose: platform_scope_next, SequencePolicy: next}` binding. Separately mutate the sealed session's canonical ODB/external root/recovery slot/helper/interpreter/tool install receipt、`ChannelNamespaceDigest`/`ChannelSlotIdentityDigest`/`SlotSequence`、provider credential or attestor capability, and try copied raw bootstrap/receipt plus reused/foreign opaque session. Script vectors replace the committed blob、Git mode、snapshot root/file identity、link count or no-follow result and prove an installed same-name script/receipt cannot satisfy the tracked token. Each failure must occur before WAL、provider、script or channel access. The positive fixture uses the fixed Linux sealed store and proves `VerifyFixedNativeRunnerInstallation` is non-consuming, only the public parent can open once, and the tracked profile/session expose no machine path or protected handle to the caller.

The platform-profile matrix also removes/empties/aliases `BootstrapLauncherCatalogRole`, swaps in the Windows launcher role or adds the launcher to role receipts；all reject before store/WAL access. The positive token fixes `BootstrapLauncherCatalogRole=c12_runner_launcher_linux_amd64`, while its machine-local launcher attestation remains outside the tracked JSON.

Add `TestPlatformRawRunnerprofileCallsOwnedOnlyByEvidenceFacade`. It parses every module-local production `.go` file with `go/parser` regardless of build tag, resolves aliases for the exact `internal/c12runnerprofile` import and rejects dot imports. Its literal platform-symbol set covers Begin/Seal、`InspectFixedPlatformResumeRecoveryDispatch`、every phase-specific recover/Inspect、Phase-B trusted-time bind/recover/promote、workload/result、all three cleanup origins/recovery/views/finalizers including `DescribeFixedPlatformResumeCleanup|OpenFixedPlatformResumeCleanupCapability`、post-clean Inspect/attestors/sender/finalizer；only `internal/c12evidence/platform.go` may reference those selectors. Its literal shared native-time set covers Bind/Inspect/Exchange/InspectResult/Consume and permits references only in `internal/c12evidence/trustedtime.go`. It inspects every `SelectorExpr`, not only direct calls, so function values and forwarding wrappers cannot bypass it；all `cmd/**`、runner-helper and alternate packages must have zero reference. It also rejects any exported platform/trusted-time c12evidence signature containing a raw P08 sink/successor/runtime/exchange/workload-result/cleanup-descriptor/post-clean/sender type；within those signatures the only allowed P08 capability parameter is the authenticated fresh session accepted by the two role-fixed client binders and manifest builder. The exported post-clean Publish signature is wrapper-only and may accept neither caller evidence structs nor canonical `[]byte`; `platform.go` alone reads the current post-clean defensive view and reconstructs them. The separately scanned authority facade permits the authenticated fresh session only in the named unsigned-facts validator、sealed closed-set binder and abort entrypoint frozen below；Complete consumes only the resulting sealed token. That exception does not widen the platform set.

Add `TestPlatformEvidencePhaseBRecoveryDispatchIsClosedAndNoFallback`, `TestPlatformEvidenceCleanupRecoveryRestoresOriginAndSuccessResult`, `TestPlatformEvidenceCleanupUnknownOriginFailsClosed`, `TestPlatformEvidencePostCleanupSuccessFacadeOnlyAfterExactLeafReceipt`, `TestPlatformEvidenceAbortCleanupCannotPublish`, `TestPlatformEvidencePostCleanupRecoveryNeverResignsOrRebinds` and `TestPlatformEvidenceFacadeHasNoRawAttestorSenderNonceOrTransactionSurface`. They use a genuinely fresh process after success-origin CAS response loss、cleanup target CAS and every attestation intent/result/sender write, recompute and exact-match `PublicationInputsDigest` only by calling P08's sole `c12runnerprofile.FixedPostCleanupPublicationInputsDigest` implementation with its exact `uint8(role length)||role||uint8(member count)||Σ(uint64be(member length)||member)` framing, strict-decode/cross-check the defensive manifest/result/receipt/bundle tuple without caller evidence/bytes, and prove the high-level parent selects exactly one authenticated route, every target rechecks it, pathless recovery restores the exact success result、never probes/falls through finalizers、never re-signs/rebinds and never exposes a raw bridge.

Task 4 freezes a 16-root OS-neutral semantic manifest. `internal/c12evidence/platform_test.go` owns `TestPlatformSemanticGateExactCoversDeclaredTests`、`TestPlatformRawRunnerprofileCallsOwnedOnlyByEvidenceFacade`、the seven `TestPlatformEvidence...` roots above、`TestResumeManifestCompleteSemanticMatrix`、`TestResumeManifestRejectsCounterIdentityReuse` and `TestOpenResumeManifestAuthenticatesBeforeHandleAccess`. `internal/c12evidence/trustedtime_test.go` owns `TestTrustedTimeEvidenceCompleteSemanticMatrix`、`TestOpenInheritedCompletionTrustedTimeRuntimeIsChildOnlyAndFixed`、`TestInheritedCompletionTrustedTimeRuntimeMintsSealedProvenanceAndConsumesOnce` and `TestInheritedCompletionTrustedTimeRuntimeRejectsPolicyBindingSplice`. The two aggregate roots contain literal subcase tables for every otherwise prose-defined resume/trusted-time mutation；no other top-level `TestPlatformEvidence*|TestResumeManifest*|TestOpenResumeManifest*|TestTrustedTimeEvidence*|TestOpenInheritedCompletionTrustedTimeRuntime*|TestInheritedCompletionTrustedTimeRuntime*` definition is permitted.

`scripts/invoke-exact-platform-semantic-gate.ps1` initially contains that literal `{Package,OwnerFile,Test}` table and accepts no arguments or environment override. It builds only per-package anchored literal alternations, requires `go test <package> -list <pattern>` to equal the sorted table exactly, then uses `go test <package> -json -run <pattern> -count=1` to require one root `run`、one root `pass` and zero root/descendant `skip` for all 16 names. The meta-test exact-compares its own literal table to the script, parses every owner file independent of build OS, requires one top-level definition per member and uses `go/types` to reject reachable `Skip|Skipf|SkipNow` or unresolved dynamic package-local helper targets. Missing、renamed、duplicate、moved、extra prefixed or zero-matched roots fail closed. Task 5 extends this same script with a second immutable parent table；it may not weaken or replace the Task 4 entries.

- [ ] **Step 2: Run platform-contract tests and verify RED**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-platform-semantic-gate.ps1`

Expected: FAIL because platform evidence and resume types are undefined.

- [ ] **Step 3: GREEN — implement the cross-reboot manifest**

```go
type CrossRebootResumeManifestV1 struct {
	SchemaVersion          string
	RunID                  uuid.UUID
	StartedAt              time.Time
	ExpiresAt              time.Time
	RepoCommit             string
	GitTreeLocator         GitTreeLocatorV1
	TrackedTreeDigest      contracts.Digest
	SpecDigest             contracts.Digest
	ReleaseDigests         [3]contracts.Digest
	RunnerIdentity         contracts.Digest
	ProviderIdentitySet    contracts.Digest
	Phase                   string
	CounterHandleDigests    [3]contracts.Digest
	CounterIdentityDigests  [3]contracts.Digest
	ExpectedCounterFloors   [3]uint64
	ExpectedStateDigests    [3]contracts.Digest
	TrustedTimeSourceIdentity contracts.Digest
	PhaseATrustedTimeStatementDigest contracts.Digest
	TrustedTimeSequence     uint64
	TrustedTimeFloor        time.Time
	ToolchainDigest         contracts.Digest
	AuthorityOperationsArtifact AuthorityOperationsArtifactTupleV1
	AuthorityOperationsBinaryAndImageDigest contracts.Digest
	AuthorityOperationsSupplyChainRecordSetDigest contracts.Digest
	SupplyChainEvidenceBundleDigest contracts.Digest
	ArtifactCatalogDigest   contracts.Digest
	ExpectedProcessImagePolicyDigest contracts.Digest
	ImageSetDigest          contracts.Digest
	ArtifactScanDigest      contracts.Digest
	OwnershipWALReservationDigest contracts.Digest
	OwnershipWALDigest     contracts.Digest
	BuildRootIdentitySetDigest contracts.Digest
	CleanupPlanDigest      contracts.Digest
	CleanupOnlyOwnershipCapsuleDigest contracts.Digest
	RecoveryAuthoritySlotIdentityDigest contracts.Digest
	InstalledExecutableProjectionDigest contracts.Digest
	ProviderProjectionDigest contracts.Digest
	ResumeOnlySessionSuccessorDigest contracts.Digest
	PhaseABootID            string
}

type OpenedResumeManifest interface {
	openedResumeManifest() // sealed validated Phase-B workload facade; owns exact P08 successor/result binding
}

type PlatformResumeManifestBuilder interface {
	platformResumeManifestBuilder() // opaque Phase-A holder; concrete privately owns the P08 sink
}

type PlatformPhaseBRecoveryDispatch string

const (
	PlatformPhaseBResumeValidationFresh    PlatformPhaseBRecoveryDispatch = "resume_validation_fresh"
	PlatformPhaseBResumeValidationRecovery PlatformPhaseBRecoveryDispatch = "resume_validation_recovery"
	PlatformPhaseBRecoverWorkload         PlatformPhaseBRecoveryDispatch = "workload"
	PlatformPhaseBRecoverCleanup          PlatformPhaseBRecoveryDispatch = "cleanup"
	PlatformPhaseBRecoverOrdinaryFinalization PlatformPhaseBRecoveryDispatch = "ordinary_finalization"
	PlatformPhaseBRecoverPublication      PlatformPhaseBRecoveryDispatch = "publication"
)

type PlatformResumeCleanupDisposition string

const (
	PlatformResumeCleanupAfterSuccess PlatformResumeCleanupDisposition = "success"
	PlatformResumeCleanupAfterAbort   PlatformResumeCleanupDisposition = "abort"
)

type PlatformResumeCleanupSession interface {
	platformResumeCleanupSession() // sealed successor+leaf capability/proof/receipt composite
}

type PlatformPostCleanupPublication interface {
	platformPostCleanupPublication() // sealed exact-clean success-only P08 successor wrapper
}

func BeginPlatformResumeManifest(context.Context, *c12runnerprofile.AuthenticatedNativeRunnerSession) (PlatformResumeManifestBuilder, error)
func SealResumeManifest(context.Context, PlatformResumeManifestBuilder, CrossRebootResumeManifestV1) (contracts.Digest, error)
func InspectPlatformPhaseBRecoveryDispatch(context.Context) (PlatformPhaseBRecoveryDispatch, error)
func OpenResumeManifest(context.Context) (OpenedResumeManifest, error)
func RecoverOpenedResumeManifest(context.Context) (OpenedResumeManifest, error)
func RunResumeManifestWorkload(context.Context, OpenedResumeManifest) (PlatformResumeCleanupSession, error)
func OpenPlatformResumeCleanup(context.Context) (PlatformResumeCleanupSession, error)
func RecoverPlatformResumeCleanup(context.Context) (PlatformResumeCleanupSession, error)
func RecoverPlatformResumeOrdinaryFinalization(context.Context) error
func InspectPlatformResumeCleanupDisposition(PlatformResumeCleanupSession) (PlatformResumeCleanupDisposition, error)
func ExecuteAndFinalizePlatformResumeAbortCleanup(context.Context, PlatformResumeCleanupSession) error
func ExecuteAndFinalizePlatformResumeSuccessCleanup(context.Context, PlatformResumeCleanupSession) (PlatformPostCleanupPublication, error)
func RecoverPlatformPostCleanupPublication(context.Context) (PlatformPostCleanupPublication, error)
func PublishPlatformPostCleanupEvidence(context.Context, PlatformPostCleanupPublication) error
```

Only phase `awaiting_cold_reboot` is serializable. Phase A obtains `StartedAt` through `BindPlatformPhaseATrustedTimeClient` from its authenticated P08 session and freezes `ExpiresAt <= StartedAt+72h`. `BeginPlatformResumeManifest` consumes that same session into an opaque builder whose concrete privately owns the P08 sink. `SealResumeManifest` accepts only the builder, first strict-validates/canonicalizes the unsigned manifest, then privately calls `SealAndSelectFixedPlatformResumeManifest(ctx,sink,canonicalUnsigned)`；the sink-bound machine attestor supplies the production signature/seal, recomputes both the unsigned projection digest and authenticated envelope digest, consumes itself and transfers the Phase-A session. No command/test/caller receives a raw sink, successor, runtime/result, sealer、signature、digest、path or provider. `InspectPlatformPhaseBRecoveryDispatch` is the sole high-level parent router and maps the authenticated P08 enum exactly as `validate_resume_fresh→resume_validation_fresh`, `validate_resume_recovery→resume_validation_recovery`, `workload_ready→workload`, `cleanup_only→cleanup`, `ordinary_finalization_pending→ordinary_finalization`, `post_cleanup_publication→publication`. Fresh is returned for the original Phase-B reservation or an absence-proved replacement reservation exactly when no exchange intent exists；the durable-intent-before-provider rule proves provider call count zero. Recovery is returned once an intent exists, including an incomplete/ambiguous external call. The fixed parent calls exactly one matching high-level route：fresh alone calls pathless `OpenResumeManifest`; validation recovery and workload call `RecoverOpenedResumeManifest`; cleanup calls `RecoverPlatformResumeCleanup`; ordinary finalization calls only `RecoverPlatformResumeOrdinaryFinalization`; publication calls `RecoverPlatformPostCleanupPublication`. Each target repeats the high-level dispatch check before touching its private raw successor. On fresh Phase B, `OpenResumeManifest` internally calls only `RecoverFixedPlatformResumeSuccessor`, Inspect-matches defensive `CanonicalUnsignedManifest` and its digest separately from `AuthenticatedResumeEnvelopeDigest`, strict-decodes the former, binds the successor-only trusted-time runtime, performs the unique nonce-bound exchange, validates the sealed monotonic elapsed record against fixed `MaximumRoundTrip` plus source/sequence/floor/expiry/session projections and promotes with that same result. `RecoverOpenedResumeManifest` is selector-free and takes no consumed predecessor；under `resume_validation_recovery` it recovers the same successor and only an already durable exact exchange response, revalidates/promotes once, while under `workload` it calls only `RecoverFixedPlatformResumeWorkloadSuccessor`. An absent/incomplete exchange intent or absent/invalid/spliced measurement on the recovery route downgrades to cleanup and never refetches；absence of intent can never enter recovery. The ordinary target delegates only to common `RecoverCleanedOwnership` and returns no session/capability. All targets reject every other phase without returning another route；unknown/corrupt/reservation-mismatched dispatch is terminal, and no error authorizes probing/falling through. Thus first Phase B and crash-before-intent replacement remain reachable, an ambiguous provider call cannot be repeated, and a crash after promotion CAS cannot resurrect the consumed resume successor. No `OpenInstalledNativeRunnerSession`、NV、WAL、root or other provider call occurs before promotion. Tests independently mutate unsigned bytes/digest、authenticated envelope digest and monotonic duration/provenance, and inject before intent reservation replacement、after intent、provider response、response durability and promotion CAS. Exact AST tests enumerate every platform-specific raw P08 symbol with sole production owner `internal/c12evidence/platform.go` (plus closed shared ordinary-finalization and trusted-time subsets owned only by `cleanup.go` and `trustedtime.go`); `cmd/**`、runner helpers and alternate packages must have zero reference/call.

Implement `testdata/c12/platform/runner-profile.v1.json` only as the strict `Role=platform` policy instance frozen above. `AllowedTrackedScriptTokens=[scripts/verify-c12-platform.sh]`; `RequiredToolRoles` is the fixed ordered Linux Bash/Docker/Git/helper/producer/trusted-time-client role set；`HelperCatalogRole=c12_runner_helper_linux_amd64`；`RecoverySlotPurpose=platform_cross_reboot_recovery`；`ExternalRootPurpose=platform_work_root`；`EvidenceRootPurpose=platform_nonpublishing_evidence`；`ProviderCredentialPurpose=platform_trusted_time_host_operator_guard`；its sole binding is `{TransactionKind: platform_scope, Access: sender, SlotPurpose: platform_scope_next, SequencePolicy: next}`；and `InstallReceiptPolicy` closes Linux/amd64 plus helper、Bash interpreter、Docker/Git/producer/trusted-time-client tool catalog roles and application-control purposes, never the snapshot script token. It contains none of the machine ODB/root/slot/absolute path/receipt/attestor/transaction facts. At this task Plan 08's fixed-token loader validates the policy instance and synthetic sealed-session fixtures；only Task 7 relock plus provisioning later lets `VerifyFixedNativeRunnerInstallation(..., platform)` validate the machine set non-consumingly, while Task 8 Phase A alone may call `OpenInstalledNativeRunnerSession(..., platform)` once. No output/evidence digest is embedded, so the instance has no source/output cycle.

`OpenPlatformResumeCleanup` is the only fresh cleanup opener and accepts no successor or selector；it internally obtains only P08's fixed pathless cleanup successor and calls both `InspectFixedPlatformResumeCleanupSuccessor` and `DescribeFixedPlatformResumeCleanup`, strict-decodes the authenticated defensive capsule envelope, derives `c12cleanup.CleanupOnlyPolicy` exclusively from that envelope, requires its embedded role、launch generation and machine-seal locator to exact-match the fixed platform role plus descriptor/selected successor, and recomputes `CleanupPolicyBindingDigest`. Only after those defensive checks does it call P08 `OpenFixedPlatformResumeCleanupCapability(ctx,successor)`；that guarded method rereads current selected state and privately opens the assignment-identical leaf capability. `RecoverPlatformResumeCleanup` obtains that same pathless P08 successor and reconstructs the same composite；neither API returns descriptor bytes、policy、capability、proof、path or selector. The session records a sealed success/abort disposition copied from the P08 origin and, only for success, the exact canonical pre-clean workload result/digest needed after process loss. It closed-maps `complete=>success` and `abort|downgrade=>abort`; an unknown origin、abort with nonzero result、success with absent/mismatched result or view/descriptor role/generation/machine-seal/policy digest mismatch fails before guarded Open/leaf access and never probes one finalizer to choose another. Execute obtains the exact leaf proof and prepared finalization receipt. The abort finalizer accepts only abort-origin and routes the receipt through ordinary `FinalizeFixedCleanedOwnership`; the success finalizer accepts only success-origin and routes it through `FinalizeFixedPlatformResumeCleanupToPublication`, returning a high-level post-clean wrapper. A mismatched method rejects without consuming the session/receipt. Expiry/source outage therefore never makes resources immortal, yet cannot create resume/sign/publication authority from an abort/downgrade successor.

Phase B enters only through the successor facade and never consumes another launch generation. `RunResumeManifestWorkload` calls P08's fixed workload facade, inspects the sealed P08 result's canonical workload、scan-receipt and supply-chain-bundle projections, applies the owning semantic validators, exact-cross-checks their repeated facts/digests against the authenticated manifest and package-derived `PlatformTrustPolicy`, and immediately converts the retained P08 successor/result to a success-origin cleanup session；any provider/workload/semantic error converts it to abort-origin cleanup before returning the error. It never exposes a raw result or lets a caller relabel the disposition. Only after the full gate succeeds may the P08 facade open declared counter/provider/WAL/root resources. The retained sealed result binds and carries the manifest's exact time/repo/tree/spec/toolchain/releases、every scanner-derived value and the exact receipt/bundle projections；success completion transfers all three into the P08 success-origin state before returning. After exact cleanup, `PublishPlatformPostCleanupEvidence` accepts only the success-only wrapper. It reads defensive bounded canonical manifest、workload-result and receipt/bundle projections from that wrapper's current guard-selected P08 successor, independently calls P08's sole `c12runnerprofile.FixedPostCleanupPublicationInputsDigest` with role `platform` and the exact four-member order/framing, constant-time exact-matches the stored digest, strict-decodes every projection, cross-checks all repeated receipt/bundle/scanner/manifest facts and digests, then strict-reconstructs nested/outer values with empty attestation fields, and rebuilds/validates all unsigned nested and outer facts from the manifest/result/cleanup receipt, calls the two fixed P08 attestors in order, constructs the fixed four-member transaction internally, binds the P08 sender, and converges `Prepare→Commit→Inspect→Finalize`. It accepts no evidence struct、canonical bytes、transaction、nonce、attestor、sender、policy or signature. `RecoverPlatformPostCleanupPublication` performs that same sole-digest call and independent strict decode/cross-check before resuming the exact recorded attestation/sender state, without re-signing or selecting another slot. Abort cleanup has no route to this facade.

For avoidance of doubt, the strict platform profile created above contains `BootstrapLauncherCatalogRole=c12_runner_launcher_linux_amd64`, and `InstallReceiptPolicy` excludes both that launcher and the snapshot script. Task 8 invokes the fixed launcher, which selects the versioned platform helper and passes the protected execution binding；only that helper may perform the single fresh Phase-A Open.

- [ ] **Step 4: GREEN — implement full platform evidence validation**

Require schema `talenro-c12-platform-evidence/v1`, distinct VM and boot IDs, UTC `StartedAt <= FinishedAt <= ExpiresAt <= StartedAt+72h` measured by the locked production trusted-time source, nonzero matching source/time-attestation digests, explicit repo/canonical tracked tree/spec/toolchain/three releases, three increasing counter floors, nondecreasing trusted time, exact signed host policy, all five sysctl/unit/policy digests, replay/attacker/post-start/operator-guard/cleanup digests, production provider set, nonzero WAL-reservation/ownership-WAL/build-root-set/cleanup-plan/capsule/slot digests equal to the sealed resume manifest and final exact-clean receipt, every nonzero scanner-derived tuple/combined/record-set/bundle/catalog/process-policy/image/full-scan value equal to the resume manifest/enclosing scope policy, and runner attestation over the exact domain-separated `CanonicalPlatformEvidenceUnsignedPayload` transcript frozen above. The timestamps、trusted-time fields、explicit build inputs and ownership projections are ordinary signed members and cannot be omitted/defaulted. Compute the full canonical envelope only after verifying that attestation, and reject it at both creation and later completion revalidation when signed trusted current time is after `ExpiresAt`. The enclosing `linux_platform_operator_trust` scope must set `TestResultDigest = SHA-256(ASCII("TALENRO-C12-NESTED-PLATFORM-RESULT-V1") || 0x00 || CanonicalPlatformEvidenceDigest)`，where the digest is the verified complete signed envelope digest, not a path, unsigned payload or caller digest. `PostStartReceiptSetDigest` signs the observed required/allowed exact-cover plus the selected `linux_platform_operator_trust` process-policy projection；missing/extra/role-swapped process、another scope's tuple or any artifact/record/bundle/catalog/full-receipt splice is diagnostic-only and returns `ErrIncompletePlatformGate`.

- [ ] **Step 5: Run platform tests and verify GREEN**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-platform-semantic-gate.ps1`

Expected: PASS; missing replay, same boot, swapped counter, fake provider, missing attacker, stale policy and failed cleanup cases all reject.

- [ ] **Step 6: REFACTOR — fuzz resume/evidence strict decoding**

Run: `go test ./internal/c12evidence -run '^$' -fuzz '^FuzzPlatformEvidence$' -fuzztime=10s -timeout 30s`

Expected: PASS without panic, input echo or handle access on unauthenticated input.

- [ ] **Step 7: Commit production platform evidence types**

```bash
git add internal/c12evidence/trustedtime.go internal/c12evidence/trustedtime_test.go internal/c12evidence/platform.go internal/c12evidence/platform_test.go scripts/invoke-exact-platform-semantic-gate.ps1 testdata/c12/platform/runner-profile.v1.json testdata/c12/platform/trusted-time-provider-profile.v1.json
git commit -m "feat: define C1.2 platform evidence"
```

### Task 5: Real Linux platform, memory isolation and OperatorClientTrustGuard gate

**Files:**
- Create: `scripts/verify-c12-platform.sh`
- Create (first line `//go:build c12_platform`): `internal/c12evidence/platform_gate_test.go`
- Create (first line `//go:build c12_platform`): `internal/nodecontrol/hostevidence/operator_guard_conformance_test.go`
- Modify: `internal/c12evidence/platform_test.go`
- Modify: `scripts/invoke-exact-platform-semantic-gate.ps1`
- Modify: `cmd/c12-fixture/main.go`
- Modify: `internal/c12evidence/runnerhelper.go`
- Modify: `internal/c12evidence/runnerhelper_test.go`
- Modify: `cmd/talenro-c12-runner-helper/main.go`
- Modify: `cmd/talenro-c12-runner-helper/main_test.go`
- Test: `internal/c12evidence/platform_gate_test.go`
- Test: `internal/nodecontrol/hostevidence/operator_guard_conformance_test.go`

**Interfaces:**
- Consumes: the Task 4 strict `platform` runner-profile instance and its fixed sealed Linux bootstrap/opaque install session, production counter/sealer/time/host-policy/operator-guard providers, B10 WAL/native Linux helper, the sole Plan 08 private producer, approved releases and `PlatformEvidenceV1`. The prebuilt、catalog/profile-locked Linux helper is the only public parent：`run-platform-phase-a` owns initial capture/build and `resume-platform-phase-b` owns authenticated resume. Plan 08 B10 already owns、locks、installs and tests the distinct zero-argument `capture-staged-and-build-closed-set` binary under its fixed `staged_diagnostic` profile；Task 5 neither upgrades nor replaces that installed diagnostic. Task 5 platform-parent source changes take effect only after the final Task 7 rebuild/relock/provisioning and cannot become a new Task 6A dependency. Each platform parent holds its snapshot、cleanup/resume authority、native trusted-time/provider handle、role-bound attestors、canonical receipt/bundle and output handles；the fixed Bash script is only a one-shot finite stage launched from the retained read-only snapshot with a fixed role on standard input and a bounded result record on standard output. Its sole path argv is the parent-derived、identity-verified absolute snapshot script path；no caller-selected repo/profile/script/provider-config/receipt/bundle/output path、digest or numeric FD/HANDLE enters argv/env, cwd and the sanitized environment are parent-fixed, and the shell cannot sign evidence or select another profile/script/output.
- Produces: the production-capable platform evidence/gate path；in the final Task 8 run the Phase B native parent canonicalizes and signs both nested `PlatformEvidenceV1` and attested `linux_platform_operator_trust` `C12ScopeEvidenceV1` with `provider_class=production` and the shared operations tuple/combined/record-set/catalog/process-policy/image/full-scan bindings, then atomically commits its four-member attested channel transaction. The final enclosing scope also requires the Task 6A operations artifact receipt, so Task 5 implements and contract-tests the gate but Task 8 performs the reusable production run after Task 6A.

- [ ] **Step 1: RED — add phase, replay and external-attacker contract tests**

Require cases `unsafe-preflight`, `phase-a`, `phase-b`, `same-uid-attacker`, `different-uid-attacker`, `canary-crash`, `operator-guard`, and `cleanup`. The fake production-provider fixture must be rejected before any test case runs.

Add `TestPlatformPhaseATransfersRunnerSessionExactlyOnce` and `TestPlatformPhaseBRejectsSecondRunnerSessionOpen`. The positive row enters through the Linux launcher, opens one platform launch session in Phase A, atomically cross-binds its installed executable/provider/session projection into the guard-selected durable resume-manifest successor, and lets Phase B recover it only through `OpenResumeManifest`. Crash before/after manifest signing/candidate write/fsync、guard compare-and-advance、selected-state reread or execution-reservation replacement, a second `OpenInstalledNativeRunnerSession(..., platform)` in Phase B, a copied session, changed install/launcher/boot identity or cleanup facade used as resume authority must reject；expiry/provider outage may obtain only the sealed `PlatformResumeCleanupSession` from the P08 cleanup successor.

Task 5 extends the frozen platform semantic script with exactly six `internal/c12evidence/runnerhelper_test.go` roots：`TestPlatformParentSemanticGateExactCoversDeclaredTests`、the two Phase-A/Phase-B roots above、`TestPlatformParentClosedLauncherAndRecoveryLifecycle`、`TestPlatformParentCleanupExactAbsenceAndNoPendingHandle` and `TestFixturePlatformStageClosedContract`. The three aggregates own literal cases for the launcher/help/private-helper surface、old-containment absence and same-generation recovery、fixed script/stdio/environment、fixture stage framing、all cleanup identities and zero pending handle. In the same change, Task 5 edits the existing `TestPlatformSemanticGateExactCoversDeclaredTests` literal table to expect the complete 22-entry script while requiring its first 16 `{Package,OwnerFile,Test}` entries to remain byte-for-byte equal to the Task 4 core partition. The new parent meta exact-compares the same complete union, exact-covers every top-level `TestPlatformParent*|TestPlatformPhaseA*|TestPlatformPhaseB*|TestFixturePlatformStage*` definition and applies the same unique-owner/go-types/reachable-helper/no-skip rules. Thus the Task 4 meta cannot fail merely because the authorized parent partition was appended, yet neither table can accept a unilateral or reordered extension. No parent root is build-tagged；the tagged production aggregators remain a separate Linux gate.

- [ ] **Step 2: Run tagged platform contracts and verify RED**

Run on the attested Linux runner: `/usr/bin/bash --noprofile --norc -c 'go test -tags=c12_platform ./internal/c12evidence ./internal/nodecontrol/hostevidence -count=1 -timeout 20m'`

Expected: FAIL because `verify-c12-platform.sh` and production cases are absent.

- [ ] **Step 3: GREEN — implement unsafe negative VM preflight**

Run each unsafe profile as a separate fresh boot: Yama absent/non-3, unprivileged BPF non-1, suid dumpable nonzero, nonempty regular core pattern, pipe core pattern, core_uses_pid nonzero, collector unmasked, SELinux permissive, wrong policy digest, wrong domain map and perf/BPF permission. Assert agent/supervisor fail before secret read, credential generation or child create; record only finite case/result digests, then destroy the VM and WAL-record its exact destroy receipt.

- [ ] **Step 4: GREEN — implement Phase A and authenticated cold-reboot handoff**

Consume the three already cleanup-capsule-declared handle reservations only from the reserved conformance namespace, verify pairwise distinct provider identities, write agent/supervisor/latch state, active local/supervisor faults and transition capsule, and record each intent/actual. Then construct the resume manifest with the exact cleanup-only capsule digest/slot identity and identical run/runner/provider/WAL-key/resource/cleanup projections. The helper writes/caps/fsyncs a bounded manifest candidate and may rename it into its fixed payload name, but those bytes have zero authority until one guarded whole-record compare-and-advance binds their identity+digest and changes the same role/run from `active_run` to `resume_only_successor`; post-advance reread is mandatory. That guard advance alone adopts the WAL authority and invalidates the cleanup-only state. Only after the selected successor and its `exiting` Phase-A execution reservation are durable may the launcher accept the controlled cold reboot. Do not enumerate TPM handles and do not emit key/state bytes.

- [ ] **Step 5: GREEN — implement Phase B replay and time tests**

After reboot, authenticate the role guard、selected cell、fixed recovery payload and new attested boot identity before exact handle access. A guard-selected `cleanup_only` capsule permits only exact cleanup and can never run Phase B/workload；a guard-selected adopted resume manifest permits Phase B and makes the old capsule nonce/digest unusable. Reject old agent/supervisor/latch blob replay, counter/blob mismatch, every pairwise cross-swap, highest-verified/highest-applied/rollback-capsule replay, power-loss window and latch damage. Inject cold reboot after cleanup-capsule binding, producer return, every intent/actual and each manifest candidate create/write/fsync/rename、guard advance/reread and execution-reservation seam；every pre-guard seam exact-cleans only through the selected capsule, every post-guard seam resumes or exact-cleans only through the selected manifest, and no snapshot/build/run/VM/NV/credential residue remains. Phase B replaces the old reservation only after the authenticated new boot proves the prior process namespace ended；same numeric PID/start/cgroup reuse or cross-boot kill rejects. Exercise wall-clock rollback, permitted VM snapshot rollback and production time-provider unavailable; no core may run before both latches clear and server finalization/explicit resume.

- [ ] **Step 6: GREEN — implement positive host policy and post-start receipts**

Create a fresh positive VM from the same attested image/toolchain. Before secrets, require SELinux enforcing, exact policy/domain-map, Yama 3, unprivileged BPF disabled, suid dumpable 0, empty core pattern, core_uses_pid 0 and collector mask; compare before/after reboot readbacks. For approved core, native-check and smoke-client record pinned executable FD, post-start pidfd image, UID/GID/capability, SELinux domain, NoNewPrivs, seccomp, cgroup and prlimit receipts; crash each canary and prove no handler/artifact/secret result.

- [ ] **Step 7: GREEN — implement external attacker cases outside target seccomp**

Launch one attacker with exact target core UID/domain and another with different UID/allowed core domain, both different PID/cgroup and neither inheriting target seccomp. Attempt `PERF_SAMPLE_STACK_USER`, user registers/high-frequency sampling, ptrace, both `process_vm_*`, `/proc/<pid>/mem`, pidfd_getfd, BPF map and program load. Each must fail at LSM/syscall boundary, produce no perf sample/BPF object, and leave target running.

- [ ] **Step 8: GREEN — implement production OperatorClientTrustGuard conformance**

Against the real provider, test lower package, same-version fork, client disk rollback then restart, provider restart, normal CA overlap, emergency removal, provider unavailable, nonce replay and expired attestation. Current attested digest is the only accepted state; both existing and new transports close on unavailable/replay/expiry.

- [ ] **Step 9: Freeze the production native parents and run the tagged gate contracts**

Freeze the following final behavior now；Task 5 unit/tagged contracts use synthetic sealed fixtures and the policy-only profile, while only Task 8 may invoke it after Task 7 relock/provisioning. On the final platform runner, the fixed Linux launcher validates its base attestation、guard-selected install record and execution containment, spawns the exact versioned helper, and that actual `run-platform-phase-a` parent calls `OpenInstalledNativeRunnerSession(..., platform)` exactly once. The Linux adapter opens `/var/lib/talenro/c12/runner-bootstrap.v1.sealed`, validates the final helper/interpreter/tool `C12NativeRunnerInstallReceiptV1` set, and consumes its opaque session to authenticate the canonical ODB、external root、recovery slot、provider/attestor policies and single sender channel binding. Its unexported role facade then generates the run ID、all opaque snapshot/build/run/resume names and every VM/NV/credential reservation without creating any resource, derives the sealed `CleanupOnlyOwnershipPlan` from those values plus the session, and executes `BeginProtectedWALKeyBootstrap(ctx,session,plan) → CreateOrRecoverProtectedWALKeyBootstrap → SealAndSelectProtectedWALKeyBootstrap({IssuedAt,ExpiresAt})`. The composite fills every machine/run/slot/provider/parent/resource/WAL projection internally, derives the key/final-WAL reservation digests, seals+file/directory-fsyncs the capsule, atomically consumes the intent and returns opaque selected ownership；the parent never receives a raw binding、intent、key handle or capsule signer. A claimed generation before its first intent routes only to zero-resource `abort_unstarted`; a selected intent permits only its one provisional key and routes only to exact `abort_intent`; an unselected capsule has no authority. Only a selected capsule is the ordinary-resource barrier；the same facade must immediately call the sole owning consumer `OpenSelectedOwnershipWAL(ctx,selected)`, which internally exact O_EXCL-creates/reopens the reserved WAL, writes+file/directory-fsyncs its bootstrap and observed file-identity actual, consumes selected ownership and returns only sealed `SelectedOwnershipWAL`. Only after that return may the parent use the sealed WAL to record intents/actuals, create the run root, capture/materialize the exact committed snapshot and invoke its private producer with an explicit FD table. It creates/WAL-records `<platform-run-root>/<opaque-a>` and `<platform-run-root>/<opaque-b>` where each opaque child is an independently generated 128-bit lowercase-hex name already bound by the plan/capsule, requires byte-identical complete builds/receipts/bundles, and retains root A read-only through reboot. Before any role-specific Phase-A recovery route, the zero-argument parent exact-dispatches `abort_unstarted|abort_intent|role_dispatch`; either abort exits after its exact pre-resource terminalization and cannot fall through. The platform script has no install receipt：the parent derives it from the retained committed snapshot and verifies exact blob bytes、Git mode、snapshot/file identity、no-follow regular-file and one-link constraints before use. All OCI layouts are already present by digest；network/pull/tag/registry fallback fails.

The only operator commands are zero-argument public parents；all repo/profile/script/recovery-slot/channel identities come from the validated fixed runner bootstrap, not cwd、PATH、argv or environment:

```text
/usr/libexec/talenro-c12/talenro-c12-runner-launcher run-platform-phase-a
```

After the controlled cold reboot:

```text
/usr/libexec/talenro-c12/talenro-c12-runner-launcher resume-platform-phase-b
```

The Phase A parent constructs the purpose-bound platform trusted-time client only through `BindPlatformPhaseATrustedTimeClient` from the authenticated session and profile's exact `platform_linux_amd64` projection. It pins the installed `/usr/bin/bash` interpreter and derives the exact tracked platform script from its retained snapshot, sets that snapshot as cwd, builds the closed sanitized environment, and starts only `/usr/bin/bash --noprofile --norc <parent-derived absolute snapshot script>`；path/cwd/env are never accepted from the caller. The one-shot child's standard input/output carry only a closed stage enum and bounded authenticated result record；no numeric descriptor is placed in argv/env. Each child exits after one stage；the parent validates PID/start token/order, kills the full process group and rejects a grandchild writer or held-open stdio. The parent alone owns Docker/provider/producer/scanner/WAL/canonical receipt/bundle/signing/deletion state；Bash cannot inspect or compare any channel transaction. Phase A's local cleanup defer remains armed until `SealResumeManifest` returns the sink-authenticated envelope digest, the guard-selected successor rereads successfully and the launcher accepts reboot. Only then may Phase A return without deleting declared resources. `resume-platform-phase-b` never calls Open again；it calls only the successor facade. An expired/provider-unavailable manifest is downgraded by P08 and opened only as the high-level cleanup session, which can exact-clean but never workload or sign.

After Phase B, the native parent gives the sealed workload facade to `RunResumeManifestWorkload`; that facade validates the result and returns a sealed success- or abort-origin cleanup session without exposing any deleter/policy/proof. The matching execute/finalize method proves every VM/NV/credential/resume/snapshot/build/run resource absent, persists the terminal WAL and leaf finalization receipt, and drives the P08 origin-bound two-CAS finalizer. Abort selects ordinary terminal-inactive and has no signing route. Success selects `platform_post_cleanup_publication` in the same generation and returns only `PlatformPostCleanupPublication`; the next generation is not available yet. `PublishPlatformPostCleanupEvidence` takes only that wrapper, recovers the exact manifest/workload/receipt/bundle bytes retained in the current P08 state, exact-recomputes `PublicationInputsDigest` only through P08 `c12runnerprofile.FixedPostCleanupPublicationInputsDigest(platform, manifest, workload, receipt, bundle)` and its frozen unsigned big-endian framing, then strict-reconstructs and validates nested/outer projections with empty attestations internally, consumes/self-verifies the two P08 attestors in fixed order, constructs the four-member transaction itself and drives the P08 sender through `Prepare`/`Commit`/`InspectSender` before selector-free finalization. No raw attestor、sender、nonce or transaction crosses the facade. Lost response calls `RecoverPlatformPostCleanupPublication`, which repeats the same digest call plus independent strict decode/cross-check and can recover only the exact retained attestation/sender state；it never opens、creates、re-signs or selects another slot. Only the final committed-sender CAS consumes the successor, marks the reservation exiting and advances the next generation；launcher wait/reap+absence proof then clears it. Contract tests use a fresh process after the cleanup target CAS and inject every leaf receipt/pending/commit/guard seam plus both attestation intent/result response-loss points and every sender transition, proving a cleanup/sign failure publishes none, a post-Commit crash leaves only the exact transaction, abort cannot mint publication, and at most one helper owns the same generation.

At Task 5, run the tagged contract/fake-provider rejection suite and prove both phases refuse a missing/zero/spliced tuple、combined/record-set/bundle/catalog/process-policy/image/full scanner receipt, a bundle body/reference/digest mismatch, or mismatched/read-write OCI layout before seal/scope signing；do not create reusable production scope evidence. In Task 8, run these two parents with the final Task 6A artifact projection；both must exit 0, boot IDs differ, provider/VM identity remains valid, all negative/positive/attacker/operator-guard cases pass, observed process receipts exact-cover the platform scope policy, and the channel output validates as unexpired production `linux_platform_operator_trust` with all shared scanner-derived bindings.

Before the owning tagged command, freeze exactly one substantive top-level aggregator in each tagged owner：`TestPlatformProductionConformanceMatrix` in `internal/c12evidence/platform_gate_test.go` and `TestPlatformOperatorGuardConformanceMatrix` in `internal/nodecontrol/hostevidence/operator_guard_conformance_test.go`. Add `TestPlatformTaggedGateExactCoversDeclaredTests` to the first file. The production aggregator's literal registration table is exactly `unsafe-preflight|phase-a|phase-b|same-uid-attacker|different-uid-attacker|canary-crash|cleanup`; the guard aggregator's table is exactly `operator-guard`. The meta-test parses all test files regardless of tag, exact-checks those three name→owner definitions、the two first-line tag owners and both case-name→aggregator tables, rejects duplicates/extra tagged top-level tests/case registrations and proves each aggregator ranges its exact table once without a condition、filter or dynamic name. With `go/types` call resolution it rejects `Skip`、`Skipf` or `SkipNow` in each aggregator and every package-local helper transitively reachable from a registered case. Every required case below is one direct named subtest of its owning aggregator. After creating both owners and before the tracked assertion, stage exactly those two files；Step 11's complete add restages them with the rest of the task. Then run this exact static/list/JSON gate:

```bash
set -euo pipefail
platform_tag_files=(internal/c12evidence/platform_gate_test.go internal/nodecontrol/hostevidence/operator_guard_conformance_test.go)
git add -- "${platform_tag_files[@]}"
git ls-files --error-unmatch "${platform_tag_files[@]}" >/dev/null
for file in "${platform_tag_files[@]}"; do
  [[ "$(sed -n '1p' "$file")" == '//go:build c12_platform' ]]
done
platform_tag_packages=(./internal/c12evidence ./internal/nodecontrol/hostevidence)
platform_tag_pattern='^(TestPlatformOperatorGuardConformanceMatrix|TestPlatformProductionConformanceMatrix|TestPlatformTaggedGateExactCoversDeclaredTests)$'
platform_tag_expected=(
  TestPlatformOperatorGuardConformanceMatrix
  TestPlatformProductionConformanceMatrix
  TestPlatformTaggedGateExactCoversDeclaredTests
)
platform_tag_case_expected=(
  'TestPlatformOperatorGuardConformanceMatrix/operator-guard'
  'TestPlatformProductionConformanceMatrix/canary-crash'
  'TestPlatformProductionConformanceMatrix/cleanup'
  'TestPlatformProductionConformanceMatrix/different-uid-attacker'
  'TestPlatformProductionConformanceMatrix/phase-a'
  'TestPlatformProductionConformanceMatrix/phase-b'
  'TestPlatformProductionConformanceMatrix/same-uid-attacker'
  'TestPlatformProductionConformanceMatrix/unsafe-preflight'
)
platform_tag_expected_text="$(printf '%s\n' "${platform_tag_expected[@]}" | LC_ALL=C sort)"
platform_tag_actual_text="$(go test -tags=c12_platform "${platform_tag_packages[@]}" -list "$platform_tag_pattern" | grep -E '^Test[A-Za-z0-9_]+$' | LC_ALL=C sort)"
[[ "$platform_tag_actual_text" == "$platform_tag_expected_text" ]]
platform_tag_json="$(mktemp)"
trap 'rm -f -- "$platform_tag_json"' EXIT
go test -tags=c12_platform "${platform_tag_packages[@]}" -json -run "$platform_tag_pattern" -count=1 -timeout 30m >"$platform_tag_json"
for name in "${platform_tag_expected[@]}"; do
  run_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"run\""){c++} END{print c+0}' "$platform_tag_json")"
  pass_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"pass\""){c++} END{print c+0}' "$platform_tag_json")"
  skip_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"skip\""){c++} END{print c+0}' "$platform_tag_json")"
  [[ "$run_count" == 1 && "$pass_count" == 1 && "$skip_count" == 0 ]]
done
for name in "${platform_tag_case_expected[@]}"; do
  run_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"run\""){c++} END{print c+0}' "$platform_tag_json")"
  pass_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"pass\""){c++} END{print c+0}' "$platform_tag_json")"
  skip_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"skip\""){c++} END{print c+0}' "$platform_tag_json")"
  [[ "$run_count" == 1 && "$pass_count" == 1 && "$skip_count" == 0 ]]
done
! awk 'index($0,"\"Action\":\"skip\"") && (index($0,"\"Test\":\"TestPlatformProductionConformanceMatrix/") || index($0,"\"Test\":\"TestPlatformOperatorGuardConformanceMatrix/")) { found=1 } END { exit(found ? 0 : 1) }' "$platform_tag_json"
rm -f -- "$platform_tag_json"
trap - EXIT
```

Require this block's exit status zero. Exact `-list` comparison makes a missing/renamed/extra aggregator fail；JSON requires one run+pass and zero skip for every top-level owner and all eight literal full subtest names, then rejects any skip event anywhere below either aggregator tree. The meta-test prevents weakening the literal owner/name/case maps or hiding a skip in a called helper. A missing/untracked file、deleted/renamed case or wrong literal first-line tag cannot be hidden by a passing parent aggregator.

Run the owning parent surface and package gates:

```text
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-platform-semantic-gate.ps1
go test ./internal/c12runnerprofile -run '^(TestNativeRunnerPlatformRecoveryUsesAttestedBootIdentity|TestNativeRunnerLauncherDeathCannotLeaveTwoLiveParents|TestNativeRunnerRecoveryProvesBoundParentAbsentBeforeSuccessor|TestNativeRunnerLaunchReservationCrashSeams)$' -count=1 -timeout 30m
go vet ./cmd/c12-fixture ./cmd/talenro-c12-runner-launcher ./cmd/talenro-c12-runner-helper ./internal/c12runnerprofile ./internal/c12evidence
```

Expected: PASS；the Linux launcher's help contains `capture-staged-and-build-closed-set`、`run-platform-phase-a`、`resume-platform-phase-b` and `run-authority-final` exactly once plus only its three provision modes；the versioned helper's unbound help contains none of those machine modes while its protected inherited dispatcher runs the selected implementation. Phase A opens one launch session and Phase B performs zero native-session opens；launcher death/reboot proves the old containment absent before the same-generation successor, and all second-open、wrong boot/launcher/helper、direct-helper/shell/PATH/provider-config/output-path/fake-provider/ambient-cwd and escaped-grandchild attacks fail before a provider or resource call. A compatibility assertion validates but does not rebuild/upgrade Plan 08's already installed diagnostic role generation or launcher base attestation；Task 5's helper source digest must differ from that installed B10 diagnostic until the final relock/provisioning, while the frozen launcher digest must remain identical.

- [ ] **Step 10: REFACTOR — verify exact cleanup and no pending handle**

Rerun `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-platform-semantic-gate.ps1` and the exact `TestNativeRunnerLaunchReservationCrashSeams` root above；`TestPlatformParentCleanupExactAbsenceAndNoPendingHandle` is the sole launcher→helper parent-owned cleanup-verification aggregate, and there is no operator-callable Bash cleanup/evidence-path command.

At Task 5 validate this command against the run-scoped cleanup fixture；at Task 8 run it against the real final evidence. Expected: PASS only after exact owner revalidation, undefine of all three run handles, run-key/state deletion, both VM destroy receipts, no pending WAL intent and cleanup digest success. The two one-way sysctls are not lowered online.

- [ ] **Step 11: Commit the Linux/operator production gate**

```bash
git add scripts/verify-c12-platform.sh scripts/invoke-exact-platform-semantic-gate.ps1 internal/c12evidence/platform_test.go internal/c12evidence/platform_gate_test.go internal/nodecontrol/hostevidence/operator_guard_conformance_test.go cmd/c12-fixture/main.go internal/c12evidence/runnerhelper.go internal/c12evidence/runnerhelper_test.go cmd/talenro-c12-runner-helper/main.go cmd/talenro-c12-runner-helper/main_test.go
git commit -m "test: add production Linux C1.2 platform gate"
```

### Task 6A: Sole production Authority Protocol v7 operations orchestrator

**Files:**
- Create: `internal/nodecontrol/operations/orchestrator.go`
- Create: `internal/nodecontrol/operations/activation.go`
- Create: `internal/nodecontrol/operations/epoch_transition.go`
- Create: `internal/nodecontrol/operations/epoch_recovery.go`
- Create: `internal/nodecontrol/operations/staging.go`
- Create: `internal/nodecontrol/operations/source_down.go`
- Create: `internal/nodecontrol/operations/orchestrator_test.go`
- Create: `internal/nodecontrol/operations/activation_test.go`
- Create: `internal/nodecontrol/operations/epoch_transition_test.go`
- Create: `internal/nodecontrol/operations/epoch_recovery_test.go`
- Create: `internal/nodecontrol/operations/staging_test.go`
- Create: `internal/nodecontrol/operations/source_down_test.go`
- Create: `cmd/nodecontrol-authority-operations/main.go`
- Create: `cmd/nodecontrol-authority-operations/main_test.go`
- Create: `deploy/c12-authority/authority-operations.Dockerfile`
- Create: `testdata/c12/artifact-catalog.v1.json`
- Create: `testdata/c12/authority/operations-build-policy.v1.json`
- Test (created, locked and installed by Plan 08): `testdata/c12/diagnostic/runner-profile.v1.json`
- Create: `docs/licenses/c12-authority-operations.md`
- Test: `internal/nodecontrol/operations/*_test.go`
- Test: `cmd/nodecontrol-authority-operations/main_test.go`

**Interfaces:**
- Consumes: B01 canonical V1 contracts, sealed-grant verifiers, provider-scoped registered 00007 migration/Down executor and typed DB repositories; B02 `FreshRestoreImportConsumer`; B03 production provider, runtime/lineage/source attestors, authenticated semantic-ID reservation, atomic Fence, row/staging challenge, archive Inspect and fixed-role response-signing clients; Plan 08 `ValidateTrackedC12SpecDigest` plus command-validated `ScanReceiptV1`; Plan 08's already frozen fixed-token `C12OperationsBuildPolicyV1` loader; and Plan 08 B10's fixed `Role=staged_diagnostic` token `testdata/c12/diagnostic/runner-profile.v1.json` plus its already locked/installed Linux sealed-bootstrap session. This task creates only the strict instance `testdata/c12/authority/operations-build-policy.v1.json` selecting B10-frozen Go/OCI recipe roles；it neither changes either generic schema/loader nor accepts a policy/profile path. Task 6A never opens the not-yet-created `authority` profile and never obtains provider、channel、signer、script、evidence-root or publishing capability. At production command startup the sole Plan 08 validator reads the attested implementation tree from fixed read-only `/run/talenro-c12/implementation-tree` and the receipt from fixed read-only `/run/talenro-c12/closed-set-receipt.json`; neither path is a CLI/config override. It does not own any B01 schema/function、B02 import repository/store或B03 journal/archive/provider implementation.
- Produces: `NewOrchestrator(Dependencies) (*Orchestrator, error)` and exact methods `InstallAndClassify(context.Context, UpgradeV7Request)`, `ActivateClaimV1Genesis(context.Context, GenesisActivationRequest)`, `AdvanceAuthorityEpoch(context.Context, EpochTransitionRequest)`, `RecoverEpochTransition(context.Context, EpochRecoveryRequest)`, `EstablishFreshStaging(context.Context, FreshStagingRequest)`, `AbortFreshStaging(context.Context, StagingAbortRequest)`, `RetireAndArchiveLegacySource(context.Context, LegacySourceRetirementRequest)`, and `RunDisposableDown(context.Context, DisposableDownRequest)`. The command exposes only `install-classify`, `activate-genesis`, `advance-epoch`, `recover-epoch`, `stage-restore`, `abort-staging`, `retire-archive-source`, `down-disposable`, and read-only `inspect`. It is built as one static auxiliary binary and one dedicated non-shell OCI image, scanned by Plan 08 and bound as the exact `AuthorityOperationsArtifactTupleV1` plus combined/record-set/catalog/full-scan digests；it is not a fourth proprietary release. This task also creates the sole final tracked artifact catalog, a digest-free approved operations build policy and license/notice obligation document；final tree/output-bound SBOM/provenance/vulnerability/secret records are generated only after commit in Task 8.

- [ ] **Step 1: RED — freeze the sole-entrypoint, reservation and transaction boundaries**

```go
func TestOrchestratorReservesEverySemanticIDBeforeExternalUse(t *testing.T) {
	t.Parallel()
	deps := newOrderedFakeDependencies(t)
	orchestrator := mustNewOrchestrator(t, deps)
	_, err := orchestrator.RecoverEpochTransition(t.Context(), validEpochRecoveryRequest(t))
	if err != nil {
		t.Fatal(err)
	}
	deps.AssertBefore(t, "authenticated-semantic-reservation", "terminal-inspect")
	deps.AssertNoExternalCallWhileDBTransactionOpen(t)
}

func TestCommandRejectsRawGooseAndCallerSelectedArchiveInputs(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{
		{"goose", "down"}, {"recover-epoch", "--decision-id", newID()},
		{"recover-epoch", "--journal-record", "caller.json"},
		{"stage-restore", "--archive-body", "caller.json"},
	} {
		assertCommandCategory(t, args, "unsupported_or_untrusted_input")
	}
}
```

Tests require authenticated reservations for every future ID/nonce before the first corresponding provider/attestor call; epoch pair/replacement share the same prefix subject-slot decision ID and reverse-unique index, while recovery application stays intent-preallocated. They also require exact retry to reuse reservations, forbid caller-selected IDs/journal/archive/Fence generations, forbid a direct archive writer dependency, and instrument every external client to fail if a DB transaction is open except the exact one-shot actual-txid-bound `authority_protocol_downgrade_authorizer` at the registered Down stage. A provider、attestor、other signer、second call or wrong-stage call still fails immediately.

Add build-boundary tests requiring `GOOS=linux GOARCH=amd64 CGO_ENABLED=0`, `-trimpath -buildvcs=false -buildmode=exe`, no external linker/custom extldflags, one non-root fixed entrypoint and no shell/package manager in `deploy/c12-authority/authority-operations.Dockerfile`. Its build context is exactly the separate `deploy/c12-authority/` tree；it must not add or read any file under Plan 08's final-locked `deploy/c12/` outer-image context. Consume without modifying Plan 08 Task 7's already final toolchain lock, whose generic Go/OCI builder identities, `SOURCE_DATE_EPOCH` value/schema, OCI config/history/layer timestamps, tar order/uid/gid/mode/xattrs and compression parameters were frozen for the future operations role. The P09 file is only a strict `C12OperationsBuildPolicyV1` instance at the fixed token；Plan 08's loader rejects alternate paths、unknown/duplicate/null/default fields、an unfrozen builder role、changed environment/timestamp/tar/compression projection or any output/tree/receipt/provenance digest before a build. The digest-free instance selects only those frozen values and never changes the loader；neither tracked file contains the eventual repo tree、output、build-receipt or provenance digest. Build the same tracked tree from two different absolute roots and require identical tuple components, manifest/config/layer and combined digest values；any ambient timestamp/path/ownership drift fails the receipt. Catalog tests require this command/Dockerfile/build-policy/license document to fill the one exact auxiliary binary/image role and require every other tracked build/OCI/release/license declaration exactly once, including distinct Windows/amd64 PE and Linux/amd64 static-ELF runner-helper roles、distinct frozen `c12_runner_launcher_windows_amd64` PE and `c12_runner_launcher_linux_amd64` ELF roles with each launcher's exact binary/SBOM/provenance triple、the distinct prebuilt `talenro_artifact_scan_windows_amd64` private completion child, plus exact PostgreSQL 18.4、Redis 8.8.1、NATS 2.14.3 integration-only OCI layouts from the tracked integration manifest. The Plan 08 scanner includes all of them in the role-bound image/artifact set and rejects a missing/duplicate/role-swapped completion child or launcher record, launcher binary without matching SBOM/provenance, helper-as-launcher alias, cross-OS helper/launcher substitution, any attempt to append an auxiliary role to `ReleaseInputDigests[3]`, or a caller catalog/path/digest. The launchers remain byte-identical B10-frozen inputs and are never Task 6A build receipts、outputs or upgrades. Any attempted Task 6A toolchain-lock or `deploy/c12/` context change is a hard failure and requires returning to Plan 08 Task 7's frozen two-commit boundary：Step 7 source commit → Step 8 clean-tree rebuild/relock → Step 9 exact-five projection commit.

Add startup tests proving the dedicated image has only its measured binary and finite runtime files, while the launcher binds the already attested implementation tree and canonical scanner receipt to the two fixed read-only paths above. Missing/read-write/wrong-tree mounts, receipt/tree/commit/spec/catalog mismatch, symlink/reparse escape or any CLI/config path override must fail before `NewOrchestrator` or an external provider call.

Add native-entry RED vectors requiring `capture-staged-and-build-closed-set` to call only `OpenInstalledNativeRunnerSession(..., staged_diagnostic)` at the fixed token/store. An authority/platform/other final-role session request, copied profile, caller profile path, diagnostic profile with any script/provider/channel/signer/evidence/publishing capability, or missing Linux helper/Git/build-tool receipt must reject before capsule、WAL、tree capture or producer creation. The positive synthetic session exposes only canonical ODB、owner-only external root、recovery slot and exact installed Linux helper/Git/build-tool handles.

- [ ] **Step 2: Run orchestrator boundary tests and verify RED**

Before the first Task 6A test or command that is allowed to open the new final `testdata/c12/artifact-catalog.v1.json`, stage exactly the Task 6A owned set (and no other path), then invoke Plan 08's locked native helper `CaptureStagedTrackedTree`. The helper alone uses the absolute hash-locked Git executable to require cached-name equality、worktree-to-index byte equality and no untracked file under an owned directory, invoke `git write-tree`, verify/materialize that exact index tree and compute the separate canonical tree digest. The catalog parser accepts only the helper's protected staged-tree locator/snapshot；opening the ambient worktree catalog, parsing before the helper's `write-tree`, or treating an unstaged catalog as tracked is RED. Repeat this candidate capture after each RED/GREEN edit and immediately before every catalog-aware test；non-catalog unit tests may run without it.

Run: `go test ./internal/nodecontrol/operations ./cmd/nodecontrol-authority-operations -run '^TestOrchestrator|^TestCommand' -count=1`

Expected: FAIL because the B11 package and unique production command do not exist.

- [ ] **Step 3: GREEN — implement the production shell and authenticated reservation plan**

Construct `Orchestrator` only when every B01/B02/B03 dependency is production-class, the fixed mounted tree's exact `SpecDigest` validates against the command-validated receipt, all receipt tuple/record/catalog/image/full-scan facts validate, and no alternate migration/archive writer or direct staging-import store is registered. The authoritative migration executor imports package `db/migrations` and constructs a scoped Goose provider with `goose.WithDisableGlobalRegistry(true)` and `goose.WithGoMigrations(migrations.NodeControlAuthorityV7Migration())`; it obtains opaque Up/Down grants only from B01 verification, consumes them through B01's exported sealed-value use APIs, does not duplicate the migration factory under `authority`, and never registers 00007 globally. Each mutating request first creates or exact-restores one rollback-resistant reservation plan containing all flow IDs/nonces and the exact Fence semantic-key contexts. Epoch pair/replacement reservations call the attestor's atomic subject-slot+reason-record operation; staging recovery/outcome reservations bind recovery/revocation application IDs to the exact intent/provider/signed-revocation preimages. Resume state is derived only from B01 immutable rows plus B03 provider/journal/archive Inspect; B11 adds no checkpoint schema and never copies signer keys, journal bytes or archive bodies into local state.

- [ ] **Step 4: RED — freeze runtime registration and the five-stage closed-until-open path**

Add tests for production Up intent, provider runtime registration, DB `AuthorityRuntimeRegistrationResultV1`, durable `AuthorityProtocolUpgradeAttemptV1`, Prepare, DB activation + row-bound proof, Complete, DB completion + proof, PrepareRelease/open nonce, DB release + proof and Open/Inspect. Inject response loss and holder restart at every boundary. Assert provider ordinary APIs/listeners stay closed through all predecessors, no stage calls provider while a DB transaction is open, and `fresh_restore_target` opens only to `fresh_v7_staging_closed` with null serving lease.

Run: `go test ./internal/nodecontrol/operations -run '^TestGenesisActivation' -count=1 -timeout 10m`

Expected: FAIL because `activation.go` and the five-stage state machine are absent.

- [ ] **Step 5: GREEN — implement runtime registration, durable attempt and five-stage activation**

Use B01 verification to obtain the sealed production Up grant, construct only the scoped Goose provider above, and install/classify 00007. If the grant is consumed and Up then rolls back, first use signed Inspect/catalog/version/latch evidence to prove not-committed, then re-run the one production verifier over the unchanged upgrade evidence to obtain a fresh one-use Up grant before a new Goose transaction；commit uncertainty always inspects first, and a committed Up is never re-granted or blindly replayed. Outside DB transactions, register the production runtime with B03；then commit the exact DB registration result before creating the durable attempt. For each activation/completion/release row, open one short B01 transaction, commit the immutable row, close it, then obtain a fresh B03 challenge/read-back attestation before invoking the next provider transition. Persist and exact-retry preparation/completion/release/open IDs from the reservation plan. `Open` consumes the provider-generated nonce only after release attestation；empty-in-place becomes active with one serving lease, while fresh target stays staging-closed and routes remain shut.

- [ ] **Step 6: RED — add the normal epoch terminal protocol and winner matrix**

Cover DB intent→provider Request→same-transaction application+resolution→row-bound resolution proof→provider Resolve CAS→DB terminal application, plus old-holder death cancellation rebind→same-transaction rebind-result+cancellation+terminal-application. Race Resolve against cancellation with two goroutines and observable barriers. Require one terminal-chain winner, exact Inspect/retry, serving lease invalidation and terminal-application catchup before ordinary readiness; reject TTL/operator cancel, second transition, partial DB pairs/triples and next-epoch before terminal application.

Run: `go test ./internal/nodecontrol/operations -run '^TestNormalEpochTransition' -count=1 -timeout 10m`

Expected: FAIL because `epoch_transition.go` is absent.

- [ ] **Step 7: GREEN — implement normal epoch advance and cancellation terminalization**

Preallocate transition/resolution/cancellation/terminal IDs, commit the intent under the B01 lock order, and call Task 11 provider only outside the transaction. On the resolved branch atomically insert application+resolution, obtain the exact resolution commit proof, consume it in the provider Resolve CAS, then commit the terminal application and perform context-descendant lease renew. On holder failure call only the B03 cancellation rebind path; after its winning provider CAS atomically commit result+cancellation+terminal application and require terminal-application catchup. Every uncertain response first uses provider/DB Inspect, and no recovery intent is created unless a real later PITR loses provider-consumed DB preimages.

- [ ] **Step 8: RED — add the complete crash/concurrency matrix**

Use observable DB/provider/attestor barriers, response-loss injection and two real goroutines/connections; never use `sleep` to infer a winner.

| Flow | Injected seam or race | Required restart result |
| --- | --- | --- |
| genesis | runtime registration, DB registration result, durable attempt, Prepare, activation proof, Complete, completion proof, PrepareRelease, release proof and Open before/after response | same reserved IDs and exact Inspect; each committed predecessor occurs once; ordinary writers/readers/listeners/signers remain closed until Open |
| normal epoch | intent, Request, application+resolution, resolution proof, Resolve CAS, terminal application and lease catchup before/after response | resolved branch exact-retries one chain; no next epoch or readiness before terminal application/catchup |
| normal epoch | Resolve CAS versus old-holder-death cancellation rebind | exactly one resolved-or-cancelled terminal chain; cancellation commits rebind result+cancellation+terminal application atomically; loser only Inspects |
| epoch | terminal Inspect before/after return; recovery-intent commit before/after response | same reserved IDs; zero-or-exact-one intent; no suffix call before committed intent |
| epoch | each deferred rebind response; holder death at every ordinal; suffix-bound proof before/after archive | same suffix/ordinal chain; deferred DB results remain absent until restricted adoption |
| epoch | provider transcript response loss; apply-vs-replacement prefix race | Inspect exact transcript; one subject-slot decision ID and one winning decision kind |
| epoch | old append versus pair/replacement Fence; Fence response loss | append winner means exact restore only; Fence winner means old generation stale; no partial generation advance |
| epoch | restricted rehydrate/application commit unknown; fresh candidate death; first-consumer/applied-result/catchup response loss | exact historical tx preimages or total rollback; fresh proof per candidate; one consumer; exact catchup/reconcile |
| staging | capability commit/proof/acquire, admission consume-before-SQL-error, import commit invocation/response loss and challenge/attestation at every boundary | definite no-commit may obtain one fresh admission after exact Inspect; uncertain/committed/PITR-loss never reimports; no duplicate application |
| staging | PITR loses unarchived fresh import application | no Fence for `FreshRestoreImportApplicationV1`; intent-backed recovery revocation then Abort |
| staging | recovery intent/provider CAS/rehydrate/recovery-application/generic proof at every boundary | zero-or-exact-one rehydrate, same recovery application ID and exact generic Fence/restore rule |
| staging | revocation sign/application/challenge/attestation/archive/Abort at every boundary | signed/journal/row/challenge ID equality; only revocation outcome Fence; exact Abort Inspect |
| source | each membership/database/environment/provider retirement, seal, post-seal observation, exact cover and archive response loss | permanent tombstones plus exact retry of remaining preallocated members; never reopen source |
| Down | membership/provider retirement race, preflight-grant consume then rollback, `1/0`, `1/1`, consume/DDL/version/commit response | production always rejects; disposable failure rolls back to `1/0`; known-not-committed retry re-verifies permanent evidence into a fresh grant and new txid authorization；uncertain/committed version/catalog is inspected, never blindly rerun |

- [ ] **Step 9: Run the epoch recovery matrix and verify RED**

Run: `go test ./internal/nodecontrol/operations -run '^TestEpochRecovery' -count=1 -timeout 10m`

Expected: FAIL because the terminal/suffix/prefix/Fence/application/catchup state machine is absent.

- [ ] **Step 10: GREEN — implement exact epoch recovery orchestration**

Drive fresh terminal Inspect, then a short B01 transaction for `AuthorityEpochTransitionRecoveryIntentV1`; append zero to 64 deferred recovery rebinds only after the intent, obtain a fresh suffix-bound commit proof, recover the exact provider transcript, and execute same-prefix apply-vs-replacement arbitration with the journal-reserved decision ID. If an unarchived candidate body must change, call the B03 atomic multi-key Fence outside the DB transaction and wait for its exact response/Inspect before opening the restricted transaction. That transaction restores provider-consumed original bodies/txid/snapshot/point/time byte-for-byte, adopts deferred results by ordinal and appends the recovery application atomically. Afterwards require a new live-candidate proof for every candidate attempt, serialize first-consumer/history-tail CAS, optional pre/post-catchup applied rebind result, terminal-application catchup or dedicated unconsumed-tail reconcile. Never use a Fence response as commit proof.

- [ ] **Step 11: Run the epoch recovery matrix and verify GREEN**

Run: `go test ./internal/nodecontrol/operations -run '^TestEpochRecovery' -count=1 -timeout 10m`

Expected: PASS for resolved/cancelled, zero/one historical row sets, suffix counts 0/1/64, apply/replacement races, same-holder retry, higher-generation candidate and every response-loss seam; cross-prefix/ID/body/generation/candidate splice and external-call-in-transaction all fail closed.

- [ ] **Step 12: RED — add staging, source/archive and Down contract tests**

Run: `go test ./internal/nodecontrol/operations -run '^TestFreshStaging|^TestLegacySource|^TestDisposableDown' -count=1 -timeout 10m`

Expected: FAIL because staging recovery, signed exact-cover/archive and registered Down orchestration are absent.

- [ ] **Step 13: GREEN — implement staging Release/Abort and PITR recovery**

Register capability in the B01 fixed lock order, obtain its row-bound proof and acquire exclusion. For the import branch, submit the complete manifest/capability/held-lease/current-Head/identity-lineage-runtime/route/pre-post facts to B01's verifier, receive the opaque `authority.VerifiedFreshRestoreImportAdmission`, and call the injected B02 `FreshRestoreImportConsumer` exactly once for that admission；B11 never calls `begin_staging_import`, a generated store function or a B01 import repository directly. Consumption is process-local and is not undone by SQL rollback. After a definite pre-Commit rollback, retry is permitted only when a fresh signed B01 Inspect plus exact application lookup/external history proves the immutable attempt never committed and the same capability remains held/current with no newer state；B11 then reruns the full B01 verifier outside the transaction to obtain a fresh one-use admission before invoking the consumer again. Once Commit was invoked or its result is unknown, inspect the exact application and external history first: a committed application returns its exact terminal result, while inability to prove absence—or PITR loss of an application known externally to have committed—permanently forbids reimport and converges through the recovery revocation path to Abort. The same/copy admission is never reused. Then obtain the external-admission challenge/attestation and Release. The non-import branch produces the signed revocation application plus challenge/attestation and Abort. For held PITR recovery, perform only allowed holder/lineage recovery, commit recovery intent and proof, call provider recovery CAS, exact-rehydrate zero-or-one capability plus recovery application, obtain the generic proof, create recovery revocation with `revocation_application_id` equal to trusted journal/intent/request/body/outcome IDs, then obtain fresh `archive_continuity` challenge/attestation and Abort. If the lost outcome is fresh import, skip Fence and converge through recovery revocation；never reimport after uncertain/committed/PITR-loss evidence. Only generic recovery application and revocation application can invoke their exact registered Fence reasons.

- [ ] **Step 14: GREEN — implement source evidence and disposable Down orchestration**

Preallocate and permanently retire the complete environment membership plus each database/environment/provider action, then have the B11 production signers mechanically build `LegacySourceRetirementSetV1`, each trusted seal/post-seal observation, `LegacySourcePostSealExactCoverV1` and `LegacySourceArchiveEvidenceV1`. Target staging builders must use locked snapshots plus pre/post proofs to sign `FreshTargetInventoryV1` and `FreshV7StagingEvidenceV1`, preserving `restore_incomplete` and closed route.

For Down, first complete external inventory membership and exact per-provider `down_retired` records/zero projections, construct the retirement set, and obtain B01's sealed preflight grant；it cannot guess a future database txid. Enter B01's registered Goose-owned transaction with that typed evidence bundle, consume the grant, obtain the actual transaction identity from the owned `*sql.Tx`, and store the exact pristine inventory at `1/0`. At that one point call the injected B11 `authority_protocol_downgrade_authorizer` exactly once；its response must bind the actual txid、transaction nonce、pristine inventory、retirement set、anchor/catalog digests and two-minute window, and B01 must verify it into the sealed authorization before the restricted insert reaches `1/1`. Consume it to `0/0`, perform DDL/version and commit atomically. On ambiguous commit, inspect catalog/version/latch before any retry；a committed transaction never reissues authorization or preflight grant. If Inspect proves the consumed transaction did not commit, re-run B01's sole Down verifier outside a transaction over the unchanged permanent retirement/preflight evidence to obtain a fresh one-use grant, then open a new transaction and sign its new actual txid. A copied/consumed grant is never reusable merely because PostgreSQL rolled back. Reject production、ordinary Goose、partial provider coverage、any nonzero fixed projection、any provider/attestor/other signer call inside the transaction, or a second/wrong-phase Down-authorizer call.

- [ ] **Step 15: Run all operations tests and command contracts**

Run: `go test ./internal/nodecontrol/operations ./cmd/nodecontrol-authority-operations -count=1 -timeout 15m`

Immediately before the final catalog parser/producer, stage exactly the Task 6A commit set—and no other path—using this same list:

```bash
git add internal/nodecontrol/operations/orchestrator.go internal/nodecontrol/operations/activation.go internal/nodecontrol/operations/epoch_transition.go internal/nodecontrol/operations/epoch_recovery.go internal/nodecontrol/operations/staging.go internal/nodecontrol/operations/source_down.go internal/nodecontrol/operations/orchestrator_test.go internal/nodecontrol/operations/activation_test.go internal/nodecontrol/operations/epoch_transition_test.go internal/nodecontrol/operations/epoch_recovery_test.go internal/nodecontrol/operations/staging_test.go internal/nodecontrol/operations/source_down_test.go cmd/nodecontrol-authority-operations/main.go cmd/nodecontrol-authority-operations/main_test.go deploy/c12-authority/authority-operations.Dockerfile testdata/c12/artifact-catalog.v1.json testdata/c12/authority/operations-build-policy.v1.json docs/licenses/c12-authority-operations.md
```

The locked native helper mechanically requires the cached name set to equal that list exactly, uses locked Git's equivalent of `git diff --exit-code -- <the same exact list>` to prove worktree bytes equal index bytes, rejects untracked files under every owned directory, and invokes `git write-tree` to produce `candidate_git_tree_oid` in the canonical repository. No plan script/parser directly invokes PATH Git for these evidence facts. This OID is only a Git object locator：for the current SHA-1 repository it is 40 lowercase hex and must never be placed in a `TrackedTreeDigest` field or 64-hex CLI argument. `CaptureStagedTrackedTree` verifies the OID/object graph, materializes that exact staged tree to its WAL-owned repository-external snapshot, and separately computes `candidate_tracked_tree_sha256` using Plan 08's `TALENRO-C12-TRACKED-TREE-V1` full path/mode/blob-SHA-256 formula. The producer reads blobs/catalog only from that protected staged snapshot and canonical object locator；it never reads an untracked/ambient worktree catalog. This makes the new catalog a tracked candidate input without pretending the uncommitted tree is authoritative；the production loader must still return `deferred_not_authoritative` because `GitTreeLocatorV1.CommitOID` is empty and no implementation commit names it.

Use only Plan 08 B10's already locked Linux machine-base launcher plus its guard-selected installed `staged_diagnostic` helper generation. The zero-argument launcher mode authenticates its base attestation、selected install record and Linux containment, installs the execution reservation, then directly spawns that B10 helper；before any capsule or resource action the actual helper calls only `OpenInstalledNativeRunnerSession(..., staged_diagnostic)`. The adapter loads Plan 08's fixed `testdata/c12/diagnostic/runner-profile.v1.json`, opens `/var/lib/talenro/c12/runner-bootstrap.v1.sealed`, validates the helper/Git/build-tool receipts plus Linux launcher projection and returns only opaque ODB/external-root/recovery/executable handles. It cannot open `authority`, and the diagnostic session contains no script/provider/channel/signer/evidence-root/publishing handle. Task 5 source changes are neither loaded nor required, and no newly built/unlocked helper or changed launcher may substitute. The parent also strict-loads the one `C12OperationsBuildPolicyV1` token through Plan 08's frozen loader and rejects any caller path/field override. Before any ordinary resource, its unexported diagnostic facade authenticates that session/recovery slot, generates the run/parent/independent opaque-child reservations, derives the sealed complete cleanup plan and executes `BeginProtectedWALKeyBootstrap(ctx,session,plan) → CreateOrRecoverProtectedWALKeyBootstrap → SealAndSelectProtectedWALKeyBootstrap({IssuedAt,ExpiresAt})`. The selected intent precedes the sole provisional key；the selected capsule consumes it and is the ordinary-resource barrier, with every machine/run/slot/provider/parent/resource/WAL field supplied internally rather than by the parent. Pre-intent active recovery proves key/WAL/ordinary creates zero and routes only to `abort_unstarted`; selected intent permits only one exact key and routes only to `abort_intent`; an unselected capsule permits no ordinary create. The facade then calls only `OpenSelectedOwnershipWAL(ctx,selected)`；that unique consumer internally exact O_EXCL-creates/reopens the reserved WAL, writes+file/directory-fsyncs its bootstrap and observed identity actual, consumes selected ownership and returns a sealed WAL with no path/key/create primitive. Only after that return may the parent append through the sealed WAL, create the repository-external run root and exactly two empty exclusive children `<task6a-run-root>/<opaque-a>` and `<task6a-run-root>/<opaque-b>` with independent 128-bit lowercase-hex names and distinct file identities, then directly launch its private producer child. Every zero-argument restart first maps the shared bootstrap dispatch；both abort targets exit after exact pre-resource terminalization and only `role_dispatch` may enter diagnostic recovery. The sole operator command is zero-argument at the fixed launcher path；repo/snapshot/staging-list/policy/WAL/run-root/output facts come only from the authenticated selected state, fixed token and retained in-process handles:

```text
/usr/libexec/talenro-c12/talenro-c12-runner-launcher capture-staged-and-build-closed-set
```

The operator-visible absolute path is the machine-base launcher；the selected catalog/toolchain-locked B10 helper performs staged capture itself, passes the staged-tree FD only to its exact private child and uses validated `C12OperationsBuildPolicyV1` to materialize/compare every catalog role twice and run both scanner passes. Because the candidate tree is not yet named by a final implementation commit, only `deferred_not_authoritative` is permitted；named receipt/bundle/scope outputs remain absent and no alternate builder/materializer exists. The non-resumable diagnostic keeps its capsule in `cleanup_only`; after crash/power loss/expiry the launcher proves containment absent and may spawn only same-generation exact-clean recovery, never restart/sign. Cleanup is resource absent → terminal WAL file+directory fsync → leaf Prepare → tombstone stage/fsync + first `cleanup_finalization_pending(receipt,ordinary_target)` CAS/reread → leaf Recover/Commit → key absence → tombstone unlink+parent-directory fsync+exact absence **while pending** → final CAS/reread to terminal-inactive/next. Target selection has no later filesystem action. Seam tests cover launcher/helper/policy/capsule/WAL/resource/producer、leaf receipt/pending/key、unlink/parent-fsync/absence、final CAS、terminal/reservation clear and boot/PID reuse；pre-barrier creates nothing, later seams converge without enumeration/file-ID splice/dual parent. Authoritative bodies first arise only from Task 8's committed implementation tree.

Expected: PASS for the full matrix, including admission-consumed-before-SQL-error、Commit-response-loss and PITR-loss-of-committed-import cases, the two-root reproducibility check and the no-receipt focused static/OCI preflight. The command emits only bounded phase/result IDs and signed evidence digests；it never emits journal/archive bytes, credentials or raw evidence bodies. No full receipt/scope exists yet and no step promotes the auxiliary artifact to a fourth release.

- [ ] **Step 16: REFACTOR — run race/static gates and dependency ownership audit**

Run: `go test -race ./internal/nodecontrol/operations ./cmd/nodecontrol-authority-operations -count=1 -timeout 15m`

Run: `go vet ./internal/nodecontrol/operations ./cmd/nodecontrol-authority-operations`

Run: `go tool golangci-lint run ./internal/nodecontrol/operations/... ./cmd/nodecontrol-authority-operations/...`

Expected: PASS; dependency audit proves B11 imports only B01 public contracts plus the `db/migrations` factory, the B02 consumer surface and B03 public contracts. It has no SQL/migration DDL, `internal/store` import, B01 `begin_staging_import`/direct import-repository call, archive/journal storage or provider implementation, and does not import `internal/c12evidence/specset.go` alternatives.

- [ ] **Step 17: Commit the orchestrator without conformance evidence**

```bash
git commit -m "feat: add Authority Protocol v7 operations orchestrator"
```

Immediately before commit, repeat the exact cached-name equality and worktree-to-index byte check through the locked native helper and require its locked-Git `write-tree` result to equal the same `candidate_git_tree_oid`；do not restage, add or modify any path between the deferred producer run and this commit. Immediately after commit, have the helper require `HEAD^{tree}` to equal that OID, capture the new exact `HEAD` commit+tree locator through `CaptureCommittedTrackedTree`, and require its separately recomputed canonical 64-hex `TrackedTreeDigest` to equal `candidate_tracked_tree_sha256`. The commit hash/tree OID locate immutable Git objects；neither substitutes for the canonical evidence digest. Task 8 uses only this committed locator/snapshot ordering: final catalog in index tree → verified candidate tree/OID+canonical digest → commit naming that exact tree → authoritative producer.

### Task 6B: Real provider, PostgreSQL PITR and destructive-restore conformance gate

**Files:**
- Create: `internal/c12evidence/authority.go`
- Create: `internal/c12evidence/authority_test.go`
- Create (first line `//go:build c12_authority`): `internal/nodecontrol/authority/production_conformance_test.go`
- Create (first line `//go:build c12_authority`): `internal/nodecontrol/operations/production_conformance_test.go`
- Create: `scripts/verify-c12-authority-v7-operations.sh`
- Create: `scripts/verify-c12-authority-fence.sh`
- Create: `scripts/invoke-exact-authority-semantic-gate.ps1`
- Create: `testdata/c12/authority/runner-profile.v1.json`
- Modify: `internal/c12evidence/runnerhelper.go`
- Modify: `internal/c12evidence/runnerhelper_test.go`
- Modify: `cmd/talenro-c12-runner-helper/main.go`
- Modify: `cmd/talenro-c12-runner-helper/main_test.go`
- Test: `internal/c12evidence/authority_test.go`
- Test: `internal/nodecontrol/authority/production_conformance_test.go`
- Test: `internal/nodecontrol/operations/production_conformance_test.go`

**Interfaces:**
- Consumes: Task 6A sole orchestrator and clean auxiliary-artifact receipt, frozen B01/B02/B03 production interfaces, real production provider/attestor/signers in an isolated tenant, PostgreSQL PITR/destructive-restore harness, Plan 08 spec-set/scope/artifact validators, WAL and runner attestor, plus Plan 08's generic fixed-token runner-profile/Bootstrap/install-receipt APIs and this task's strict `Role=authority` instance. The only final production entry is the catalog/profile-locked native helper parent `run-authority-final`；it holds the committed snapshot、cleanup authority、private producer、native authority trusted-time/provider handle、runner attestor、canonical evidence and output handles. Both Bash scripts are fixed one-shot finite stages launched by that parent as `/usr/bin/bash --noprofile --norc <parent-derived absolute snapshot script>` from the locked snapshot cwd with sanitized env and closed stdin/stdout records；they receive no caller-selected path/profile/provider/output or numeric FD and own no OCI/provider/scanner/deleter/signing operation.
- Produces: exact `AuthorityFenceEvidenceV1` validator、a build-tagged non-authoritative real-dependency diagnostic harness、receipt-gated one-shot stage scripts and the native finalizer that canonicalizes/signs nested authority plus attested `authority_fence_pitr` scope only after exact cleanup, then commits its four-member attested channel transaction. Task 6B itself cannot mint reusable evidence from an uncommitted tree；Task 8 performs the first positive receipt-pinned production parent run after the final tree is committed. It does not add a second production orchestrator or a shell evidence builder.

```go
type AuthorityFinalWorkloadResult interface {
	authorityFinalWorkloadResult() // package-private session-consuming validated result; no exported constructor
}

type AuthorityUnsignedWorkloadFacts interface {
	authorityUnsignedWorkloadFacts() // package-private typed matrix/time/nested+outer unsigned facts; no bytes or attestor
}

type AuthorityCleanupDisposition string

const (
	AuthorityCleanupAfterSuccess AuthorityCleanupDisposition = "success"
	AuthorityCleanupAfterAbort   AuthorityCleanupDisposition = "abort"
)

type AuthorityRecoveryDispatch string

const (
	AuthorityRecoverCleanup     AuthorityRecoveryDispatch = "cleanup"
	AuthorityRecoverOrdinaryFinalization AuthorityRecoveryDispatch = "ordinary_finalization"
	AuthorityRecoverPublication AuthorityRecoveryDispatch = "publication"
)

type AuthorityCleanupSession interface {
	authorityCleanupSession() // sealed P08 successor + exact leaf capability/proof/receipt composite
}

type AuthorityPostCleanupPublication interface {
	authorityPostCleanupPublication() // sealed success-only same-generation publication facade
}

func ValidateAndSealAuthorityUnsignedWorkloadFacts(context.Context, *c12runnerprofile.AuthenticatedNativeRunnerSession, AuthorityFenceEvidenceV1, C12ScopeEvidenceV1, C12TrustedTimeEvidenceV1) (AuthorityUnsignedWorkloadFacts, error)
func BindAuthorityFinalWorkloadResult(context.Context, *c12runnerprofile.AuthenticatedNativeRunnerSession, AuthorityUnsignedWorkloadFacts, FinalClosedSetProjection) (AuthorityFinalWorkloadResult, error)
func CompleteAuthorityWorkloadToCleanup(context.Context, AuthorityFinalWorkloadResult) (AuthorityCleanupSession, error)
func AbortAuthorityWorkloadToCleanup(context.Context, *c12runnerprofile.AuthenticatedNativeRunnerSession) (AuthorityCleanupSession, error)
func InspectAuthorityRecoveryDispatch(context.Context) (AuthorityRecoveryDispatch, error)
func RecoverAuthorityCleanup(context.Context) (AuthorityCleanupSession, error)
func RecoverAuthorityOrdinaryFinalization(context.Context) error
func InspectAuthorityCleanupDisposition(AuthorityCleanupSession) (AuthorityCleanupDisposition, error)
func ExecuteAndFinalizeAuthorityAbortCleanup(context.Context, AuthorityCleanupSession) error
func ExecuteAndFinalizeAuthoritySuccessCleanup(context.Context, AuthorityCleanupSession) (AuthorityPostCleanupPublication, error)
func RecoverAuthorityPostCleanupPublication(context.Context) (AuthorityPostCleanupPublication, error)
func PublishAuthorityPostCleanupEvidence(context.Context, AuthorityPostCleanupPublication) error
```

`ValidateAndSealAuthorityUnsignedWorkloadFacts` is the only typed pre-clean authority semantic mint. It accepts the authenticated authority session plus concrete c12evidence-owned `AuthorityFenceEvidenceV1`、`C12ScopeEvidenceV1` and `C12TrustedTimeEvidenceV1` candidates；requires both nested/outer attestation fields empty；derives and exact-validates the complete authoritative matrix、PITR/provider/terminal records、trusted-time statement and every repeated repo/tree/spec/toolchain/release/scanner fact against session-private stage state；and returns one private-concrete `AuthorityUnsignedWorkloadFacts`. It accepts no bytes、path、digest override、scope enum、callback、permit or caller implementation. Independently, the fixed parent calls `BindFinalClosedSetProjectionSink` on that same session, gives the opaque sink and typed clean receipt/bundle only to `artifactscan.ValidateAndSealFinalClosedSetProjection`, then passes the resulting private `FinalClosedSetProjection` plus the unsigned-facts token and same session to `BindAuthorityFinalWorkloadResult`. That binder exact-matches their run/tree/spec/toolchain/scanner/bundle identities, consumes both sealed inputs and the active session, and privately retains a fixed-order triple：strict-canonical workload-result projection capped at 64 KiB、strict-canonical scan-receipt projection capped at 64 KiB and strict-canonical supply-chain-bundle projection capped at 4 MiB, with aggregate capped by P08 `MaxFixedPostCleanupPublicationInputsBytes`; it excludes `CleanupResultDigest`、terminal cleanup proof and all evidence signatures. Command、shell、test-tag diagnostic output and caller bytes cannot construct any token. `CompleteAuthorityWorkloadToCleanup` consumes only `AuthorityFinalWorkloadResult` and is the sole high-level unwrap：inside `authority.go` it recovers the token-private session and passes exactly those three defensive byte slices to raw `CompleteFixedAuthorityWorkloadToCleanup`. `AbortAuthorityWorkloadToCleanup` alone consumes a still-active pre-mint session into the mutually exclusive abort origin. The real positive chain is `fixed authority parent → Bind sink → artifactscan typed Validate+Seal → ValidateAndSealAuthorityUnsignedWorkloadFacts → BindAuthorityFinalWorkloadResult → CompleteAuthorityWorkloadToCleanup`; no import cycle or production caller-defined interface exists. `InspectAuthorityRecoveryDispatch` is the sole high-level recovery router：it maps authenticated P08 `cleanup_only→cleanup`, `ordinary_finalization_pending→ordinary_finalization` and `post_cleanup_publication→publication` exactly, returns no raw state/capability and lets P08 atomically default an active-parent crash to abort cleanup. The fixed parent calls exactly one matching pathless route；`RecoverAuthorityCleanup`, zero-capability `RecoverAuthorityOrdinaryFinalization` and `RecoverAuthorityPostCleanupPublication` each repeat the high-level dispatch check before touching their private route. The ordinary target delegates only to common `RecoverCleanedOwnership` and can never return/remint a cleanup session. Unknown/corrupt/reservation-mismatched state is terminal, and no error authorizes probing or falling through to another route. The cleanup composite opens the P08 descriptor internally, derives the policy from the authenticated capsule, supplies only the fixed authority role, exact-matches the selected launch generation and machine-seal locator and recomputes the complete policy binding；only then does it call `OpenFixedAuthorityCleanupCapability(ctx,successor)`, whose current-state reread and private leaf Open return the capability. The composite retains capability/proof/finalization receipt without exposing any of them.

Because these are final-schema structs reused only as typed pre-clean candidates, the mint additionally requires `AuthorityFenceEvidenceV1.CleanupResultDigest == zero`、`C12ScopeEvidenceV1.CleanupDigest == zero`、`C12ScopeEvidenceV1.TestResultDigest == zero` and both attestation fields empty. All five pre-clean placeholders are excluded from the sealed workload projection；nonzero caller-prefill、an omitted zero check or a value copied from another run rejects before binding or cleanup. Post-clean construction is uniquely ordered：assignment-identical terminal leaf views populate both cleanup digests，the fixed nested attestor signs the now-complete `AuthorityFenceEvidenceV1`，the facade computes outer `TestResultDigest` from that complete signed nested-envelope digest under the frozen authority-result domain，and only then may the fixed outer attestor sign `C12ScopeEvidenceV1`. No caller value can fill any step. The existing authority semantic aggregate adds literal subcases for each nonzero prefill、an omitted zero check、outer result derivation before the nested signature and any implementation that expects a cleanup/result digest before the success-origin CAS, without increasing the frozen top-level root count.

The linear failure rule is exact：each sealed unsigned/projection token is consumed on every `BindAuthorityFinalWorkloadResult` return；the authenticated session transfers into `AuthorityFinalWorkloadResult` only on success. On a binder error it remains valid solely for the immediate no-data `AbortAuthorityWorkloadToCleanup` call and cannot mint another sink、unsigned token or Bind attempt；a crash instead follows P08's authenticated active-parent default-to-abort route. A successful Bind leaves no separate session authority, and Complete consumes its result token on every return.

The abort execute/finalize method accepts only abort origin and drives ordinary terminal-inactive/next. The success method accepts only success origin, exact-cleans and passes the assignment-identical leaf receipt to `FinalizeFixedAuthorityCleanupToPublication`, returning a same-generation high-level wrapper. Mismatched methods reject without relabeling origin or consuming a valid receipt. `PublishAuthorityPostCleanupEvidence` accepts only that wrapper. It reads defensive bounded canonical workload-result and receipt/bundle projections from the wrapper's current guard-selected P08 successor, independently calls P08's sole `c12runnerprofile.FixedPostCleanupPublicationInputsDigest(authority, workload, receipt, bundle)` implementation using exact `uint8(role length)||role||uint8(member count)||Σ(uint64be(member length)||member)` framing, constant-time exact-matches the stored digest, strict-decodes every projection, cross-checks all repeated receipt/bundle/scanner facts and digests, then strict-reconstructs nested/outer values with empty attestation fields, validates nested/outer unsigned facts from the pre-clean workload-result plus terminal cleanup views, calls the two fixed P08 attestors, binds only the P08 authority post-clean sender, constructs the fixed transaction and converges `Prepare→Commit→Inspect→Finalize`. It accepts no evidence struct、canonical bytes、attestor、sender、nonce、transaction、signature、policy or cleanup proof. Recovery after success-origin CAS loss、cleanup target CAS or either attestation intent/result seam starts a fresh process, repeats that same sole digest call and independent strict decode/cross-check of the defensive result/receipt/bundle tuple, and returns only the exact recorded signature/sender state without producer/provider/validator rerun；it never re-signs or selects a slot. Tests `TestAuthorityParentRecoveryDispatchIsClosedAndNoFallback`, `TestAuthorityParentSuccessSignsOnlyFromPostCleanupPublication`, `TestAuthorityParentAbortCannotMintPublication`, `TestAuthorityParentCleanupFailureCannotSignOrPrepareChannel`, `TestAuthorityParentPostCleanupRecoveryIsSelectorFree`, `TestAuthorityParentPostCleanupResponseLossConvergesExactTransaction` and `TestAuthorityParentHasNoRawAttestorSenderNonceOrTransactionSurface` freeze this facade.

- [ ] **Step 1: RED — add authority evidence completeness tests**

```go
func TestAuthorityEvidenceRejectsExpiredOrPendingProviderState(t *testing.T) {
	t.Parallel()
	evidence := validAuthorityEvidence()
	evidence.ExpiresAt = evidence.FinishedAt.Add(72*time.Hour + time.Nanosecond)
	if err := ValidateAuthorityFenceEvidence(evidence, testAuthorityPolicy()); !errors.Is(err, ErrExpiry) {
		t.Fatal("authority evidence exceeded maximum lifetime")
	}
	evidence = validAuthorityEvidence()
	evidence.TerminalProviderRecordDigest = contracts.Digest{}
	if err := ValidateAuthorityFenceEvidence(evidence, testAuthorityPolicy()); !errors.Is(err, ErrPendingProviderRecord) {
		t.Fatal("authority evidence accepted pending provider state")
	}
}

func TestAuthorityEvidenceRejectsStagingAsCompletedLegacyUpgrade(t *testing.T) {
	t.Parallel()
	evidence := validAuthorityEvidence()
	evidence.LegacyUpgradeStatus = "fresh_v7_staging_closed_as_complete"
	if err := ValidateAuthorityFenceEvidence(evidence, testAuthorityPolicy()); !errors.Is(err, ErrRestoreIncomplete) {
		t.Fatal("authority evidence promoted staging-closed to completed legacy upgrade")
	}
}

func TestAuthorityEvidenceRejectsOperationsArtifactSplice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*AuthorityFenceEvidenceV1, *AuthorityTrustPolicy)
	}{
		{"binary-component", func(e *AuthorityFenceEvidenceV1, _ *AuthorityTrustPolicy) { e.AuthorityOperationsArtifact.BinarySHA256[0] ^= 1 }},
		{"image-component", func(e *AuthorityFenceEvidenceV1, _ *AuthorityTrustPolicy) { e.AuthorityOperationsArtifact.ImageManifestDigest[0] ^= 1 }},
		{"build-receipt-component", func(e *AuthorityFenceEvidenceV1, _ *AuthorityTrustPolicy) { e.AuthorityOperationsArtifact.BuildReceiptDigest[0] ^= 1 }},
		{"dependency-component", func(e *AuthorityFenceEvidenceV1, _ *AuthorityTrustPolicy) { e.AuthorityOperationsArtifact.DependencyClosureDigest[0] ^= 1 }},
		{"combined", func(e *AuthorityFenceEvidenceV1, _ *AuthorityTrustPolicy) { e.AuthorityOperationsBinaryAndImageDigest[0] ^= 1 }},
		{"supply-chain", func(e *AuthorityFenceEvidenceV1, _ *AuthorityTrustPolicy) { e.AuthorityOperationsSupplyChainRecordSetDigest[0] ^= 1 }},
		{"supply-chain-bundle", func(e *AuthorityFenceEvidenceV1, _ *AuthorityTrustPolicy) { e.SupplyChainEvidenceBundleDigest[0] ^= 1 }},
		{"catalog", func(e *AuthorityFenceEvidenceV1, _ *AuthorityTrustPolicy) { e.ArtifactCatalogDigest[0] ^= 1 }},
		{"process-policy", func(e *AuthorityFenceEvidenceV1, _ *AuthorityTrustPolicy) { e.ExpectedProcessImagePolicyDigest[0] ^= 1 }},
		{"image-set", func(e *AuthorityFenceEvidenceV1, _ *AuthorityTrustPolicy) { e.ImageSetDigest[0] ^= 1 }},
		{"full-scan", func(e *AuthorityFenceEvidenceV1, _ *AuthorityTrustPolicy) { e.ArtifactScanDigest[0] ^= 1 }},
		{"policy-full-scan", func(_ *AuthorityFenceEvidenceV1, p *AuthorityTrustPolicy) { p.ScopePolicy.ExpectedArtifactScanDigest[0] ^= 1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evidence, policy := validAuthorityEvidence(), testAuthorityPolicy()
			tt.mutate(&evidence, &policy)
			if err := ValidateAuthorityFenceEvidence(evidence, policy); !errors.Is(err, ErrBuildMismatch) {
				t.Fatal("authority evidence accepted a scanner-derived binding splice")
			}
		})
	}
}
```

The RED matrix must also mutate explicit repo commit/canonical tracked tree/spec/toolchain/each release、trusted-time source/statement、each UTC timestamp and ownership/build-root/cleanup digest in the nested authority envelope and enclosing scope/policy；all splices reject. Freeze the authority signature projection: changing `RunnerAttestation` alone leaves `CanonicalAuthorityFenceEvidenceUnsignedPayload` byte-identical but changes the full envelope and fails verification；changing any other field changes the unsigned payload and invalidates the old signature. A full-envelope signature、a projection omitting any second field、a nested payload or a caller-selected projection must reject.

Add authority-runner-profile RED vectors through Plan 08's loader：wrong role/schema/token、unknown/duplicate/null/default field、missing/extra/reordered tracked script token or tool role、wrong helper catalog role/recovery/external/evidence-root/provider purpose/install-receipt policy, or a channel policy other than the sole `{TransactionKind: authority_scope, Access: sender, SlotPurpose: authority_scope_next, SequencePolicy: next}` binding. Separately splice the sealed ODB/external root/recovery slot/helper/interpreter/tool install receipt、`ChannelNamespaceDigest`/`ChannelSlotIdentityDigest`/`SlotSequence`、provider credential/attestor capability, copied raw bootstrap/receipt、foreign/reused opaque session and PE/Windows identity substituted for the Linux helper. Script vectors mutate either committed blob、Git mode、snapshot root/file identity、no-follow/link-count proof or substitute an installed same-name script/receipt；all reject. Verify that provisioning-time `VerifyFixedNativeRunnerInstallation(..., authority)` does not consume the launch generation, while `run-authority-final` opens it exactly once and a second open fails before WAL、provider、PITR、script or channel access.

The authority-profile RED matrix also removes/empties/aliases `BootstrapLauncherCatalogRole`, swaps in the Windows launcher or places either launcher in role receipts；all reject before WAL/provider/PITR access. The positive token fixes `BootstrapLauncherCatalogRole=c12_runner_launcher_linux_amd64`, and the launcher attestation remains machine-local.

Add `TestAuthorityRawRunnerprofileCallsOwnedOnlyByEvidenceFacade`. It applies the same all-build-tag AST/alias/SelectorExpr scan to the literal authority Complete/Abort/`InspectFixedAuthorityRecoveryDispatch`/Recover/Inspect/`DescribeFixedAuthorityCleanup|OpenFixedAuthorityCleanupCapability`/receipt-finalize/post-clean Inspect/attest/sender/finalizer set and permits production references only in `internal/c12evidence/authority.go`; every command、runner helper or forwarding package is forbidden. It also proves the exact exported producer surface above：the unsigned-facts validator alone may accept the fresh session plus concrete `AuthorityFenceEvidenceV1|C12ScopeEvidenceV1|C12TrustedTimeEvidenceV1`; the binder alone may accept that same session plus sealed `AuthorityUnsignedWorkloadFacts|FinalClosedSetProjection`; Complete accepts only `AuthorityFinalWorkloadResult`; Abort accepts only the fresh session；and recovery/publication use only sealed `AuthorityCleanupSession|AuthorityPostCleanupPublication`. No signature contains a raw P08 cleanup/post-clean/sender capability、canonical `[]byte`、callback、permit or caller-implementable token. A `go/types` positive call-graph submatrix compiles the actual fixed-parent chain `BindFinalClosedSetProjectionSink → artifactscan.ValidateAndSealFinalClosedSetProjection → ValidateAndSealAuthorityUnsignedWorkloadFacts → BindAuthorityFinalWorkloadResult → CompleteAuthorityWorkloadToCleanup`, requires `artifactscan→c12evidence` and c12evidence↛artifactscan, requires the sole `SealValidatedClosedSet` call in `internal/artifactscan/scan.go`, and rejects a missing/dynamic/unresolved edge. Thus only `authority.go` may unwrap the sealed pre-clean token and pass its bounded canonical bytes to the import-cycle-free P08 bridge.

The same root exact-freezes a second, high-level production owner table. Outside their definitions in `internal/c12evidence/authority.go`, every reference/call to `ValidateAndSealAuthorityUnsignedWorkloadFacts|BindAuthorityFinalWorkloadResult|CompleteAuthorityWorkloadToCleanup|AbortAuthorityWorkloadToCleanup|InspectAuthorityRecoveryDispatch|RecoverAuthorityCleanup|RecoverAuthorityOrdinaryFinalization|InspectAuthorityCleanupDisposition|ExecuteAndFinalizeAuthorityAbortCleanup|ExecuteAndFinalizeAuthoritySuccessCleanup|RecoverAuthorityPostCleanupPublication|PublishAuthorityPostCleanupEvidence` is permitted only in the fixed authority-parent implementation in `cmd/talenro-c12-runner-helper/main.go`; each required fresh/success/abort/recovery edge has the exact literal multiplicity frozen by the call table, and no other production file may mention one. The shared `BindFinalClosedSetProjectionSink` and `artifactscan.ValidateAndSealFinalClosedSetProjection` production call tables contain exactly the two already declared fixed parents—Windows final scopes and Authority final—and no third caller. The all-build-tag scan resolves import aliases and every `Ident|SelectorExpr|MethodExpr` plus `go/types` call object, so a function value、alias variable、forwarding wrapper、reflection/dynamic dispatch or unresolved edge cannot evade the count. Definition-only references and test fixtures are separately classified；an extra static edge、missing edge or wrong owner is RED without adding a new top-level root.

Step 1 defines an initial five-root OS-neutral authority semantic manifest in `internal/c12evidence/authority_test.go`：`TestAuthoritySemanticGateExactCoversDeclaredTests`、`TestAuthorityRawRunnerprofileCallsOwnedOnlyByEvidenceFacade`、`TestAuthorityEvidenceRejectsExpiredOrPendingProviderState`、`TestAuthorityEvidenceRejectsStagingAsCompletedLegacyUpgrade` and `TestAuthorityEvidenceRejectsOperationsArtifactSplice`. The three evidence roots use literal subcase tables to own every schema/time/provider/operations/spec/staging/cleanup/signature mutation required above；no other top-level `TestAuthorityEvidence*` definition is permitted.

`scripts/invoke-exact-authority-semantic-gate.ps1` initially stores that literal `{Package,OwnerFile,Test}` table, accepts no arguments or environment override, uses only anchored literal alternations, exact-compares `go test -list`, and requires the JSON stream to contain exactly one root run/pass and zero root/descendant skip for each name. The meta-test exact-compares its table with the script, parses every owner file independent of build OS, requires one unique top-level definition and uses `go/types` to reject reachable `Skip|Skipf|SkipNow` or unresolved dynamic package-local helper targets. Missing、renamed、duplicate、moved、extra prefixed or zero-matched roots fail closed. Step 9 extends this same script with the final parent manifest；the initial five entries may not change.

- [ ] **Step 2: Run authority evidence tests and verify RED**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-authority-semantic-gate.ps1`

Expected: FAIL because authority evidence validator is undefined.

- [ ] **Step 3: GREEN — implement strict authority evidence validation**

Require schema `talenro-authority-fence-evidence/v1`, production provider identity/policy/version, isolated tenant, exact control API binary+image, UTC `StartedAt <= FinishedAt <= ExpiresAt <= StartedAt+72h`, locked nonzero trusted-time source/time-attestation digests, explicit repo commit/canonical tracked tree/spec/toolchain/three releases, the complete scanner-derived operations tuple/combined/record-set/bundle/catalog/process-policy/image/full-scan bindings, client protocol/ruleset/DB system+timeline/schema/migration/PITR/test/terminal/ownership/cleanup digests, all five v7/source/staging/Down/destructive-block result digests, and trusted runner attestation over the exact domain-separated `CanonicalAuthorityFenceEvidenceUnsignedPayload` transcript frozen above. Compute the complete canonical envelope only after signature verification. The enclosing `authority_fence_pitr` scope must set `TestResultDigest = SHA-256(ASCII("TALENRO-C12-NESTED-AUTHORITY-RESULT-V1") || 0x00 || CanonicalAuthorityFenceEvidenceDigest)`，where the digest is the verified complete signed envelope digest, not a path, unsigned payload or caller digest. Recompute `SpecDigest` only with Plan 08 `ValidateTrackedC12SpecDigest` over the exact committed locator/snapshot. `authority.go` requires every component/digest nonzero, constant-time equal its `AuthorityTrustPolicy.Expected*` value and equal the enclosing scope；it does not import `internal/artifactscan` or recompute that package's domains. The command layer validates `CanonicalScanReceipt` plus the canonical bundle, recomputes tuple/combined/record/bundle/catalog/full image membership, selects the `authority_fence_pitr` expected-process policy and then supplies all expected values in policy. The authority result digest binds observed pre/post process receipts and exact-covers that selected policy. Reject a single-spec/old-set/caller digest, caller operations/scan digest, binary-only receipt, any tuple/combined mismatch, same-binary/different-image, same-image/different-binary, record-set/bundle/catalog/image/process-policy splice or same-operations/different-full-receipt splice. `LegacyUpgradeStatus` is closed to `not_attempted` or `fresh_v7_staging_closed_restore_incomplete`; no completed value exists. `DestructiveRestoreBlockDigest` proves the approved unsupported effects kept the target `restore_incomplete`; it is not an upgrade-completion digest.

- [ ] **Step 4: Run authority evidence tests and verify GREEN**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-authority-semantic-gate.ps1`

Expected: PASS; fake provider, production tenant, stale build/spec set, old protocol/ruleset, missing epoch/source/staging/Down case, missing hard-block proof, failed cleanup and tampered attestation reject.

- [ ] **Step 5: RED — add production provider crash/PITR conformance cases**

Cases must cover the baseline concurrent total order/no duplicate sequence, same-operation idempotency, changed scope/effect conflict, receipt replay/tamper, illegal abort and provider outage plus the complete Task 6A fault table against real provider/attestor/PostgreSQL. Required matrices include both empty-in-place and fresh-target runtime registration, durable attempt and every five-stage genesis boundary；normal epoch resolved and old-holder-death-cancelled terminal paths including their race and terminal-application catchup；resolved/cancelled epoch recovery with suffix 0/1/64 and apply/replacement；all four Fence reasons with only the registered semantic keys；staging Release, normal Abort and recovery Abort；source membership/retirement/seal/post-seal exact cover/archive；and disposable Down. Required negatives include premature genesis Open/serving, normal-epoch TTL/operator cancel or next transition before terminal application, fresh-import Fence/reimport, prefix/decision/journal/context splice, same-holder once-row bypass, direct archive write, production Down, partial source/provider coverage and staging-closed completion. At Task 6B these are non-authoritative diagnostics under `-tags=c12_authority`；the same complete matrix becomes authoritative only when Task 8 reruns it through the final receipt-pinned OCI entrypoint.

- [ ] **Step 6: Run tagged provider conformance and verify RED**

Freeze exactly one substantive top-level aggregator in each tagged owner：`TestProductionAuthorityConformanceMatrix` in `internal/nodecontrol/authority/production_conformance_test.go` and `TestProductionAuthorityV7OperationsConformanceMatrix` in `internal/nodecontrol/operations/production_conformance_test.go`. The authority aggregator's literal cases are exactly `concurrent-total-order|same-operation-idempotency|changed-scope-effect-conflict|receipt-replay-tamper|illegal-abort|provider-outage|four-fence-reasons`. The operations aggregator's literal cases are exactly `genesis-empty-in-place|genesis-fresh-target|genesis-five-stage-crash-matrix|normal-epoch-resolved|normal-epoch-old-holder-death-cancelled|normal-epoch-terminal-race-catchup|epoch-recovery-resolved|epoch-recovery-cancelled|epoch-recovery-suffix-0|epoch-recovery-suffix-1|epoch-recovery-suffix-64|epoch-recovery-apply-vs-replacement|staging-release|staging-normal-abort|staging-recovery-abort|source-retirement-seal-postseal-archive|disposable-down|premature-genesis-open-serving|normal-epoch-illegal-cancel-or-next|fresh-import-fence-or-reimport|prefix-decision-journal-context-splice|same-holder-once-row-bypass|direct-archive-write|production-down|partial-source-provider-coverage|staging-closed-is-not-completion|real-pitr-response-loss-refusal`. The latter file also owns `TestC12AuthorityTaggedGateExactCoversDeclaredTests`, which parses both owners regardless of tag, exact-checks the three name→owner entries、two first-line tags and both exact case-name→aggregator tables, rejects duplicate/extra tagged top-level tests/case registrations and proves each aggregator ranges its table exactly once without a condition、filter or dynamic name. With `go/types` call resolution it rejects `Skip`、`Skipf` or `SkipNow` in each aggregator and every package-local helper transitively reachable from a registered case. Every required matrix below is one direct named subtest of its fixed aggregator. After creating both owners, stage exactly those two files before the tracked assertion；Step 11's complete add restages them with the rest of the task. Run on the authority runner:

```bash
set -euo pipefail
authority_tag_files=(internal/nodecontrol/authority/production_conformance_test.go internal/nodecontrol/operations/production_conformance_test.go)
git add -- "${authority_tag_files[@]}"
git ls-files --error-unmatch "${authority_tag_files[@]}" >/dev/null
for file in "${authority_tag_files[@]}"; do
  [[ "$(sed -n '1p' "$file")" == '//go:build c12_authority' ]]
done
authority_tag_packages=(./internal/nodecontrol/authority ./internal/nodecontrol/operations)
authority_tag_pattern='^(TestC12AuthorityTaggedGateExactCoversDeclaredTests|TestProductionAuthorityConformanceMatrix|TestProductionAuthorityV7OperationsConformanceMatrix)$'
authority_tag_expected=(
  TestC12AuthorityTaggedGateExactCoversDeclaredTests
  TestProductionAuthorityConformanceMatrix
  TestProductionAuthorityV7OperationsConformanceMatrix
)
authority_tag_case_expected=(
  'TestProductionAuthorityConformanceMatrix/changed-scope-effect-conflict'
  'TestProductionAuthorityConformanceMatrix/concurrent-total-order'
  'TestProductionAuthorityConformanceMatrix/four-fence-reasons'
  'TestProductionAuthorityConformanceMatrix/illegal-abort'
  'TestProductionAuthorityConformanceMatrix/provider-outage'
  'TestProductionAuthorityConformanceMatrix/receipt-replay-tamper'
  'TestProductionAuthorityConformanceMatrix/same-operation-idempotency'
  'TestProductionAuthorityV7OperationsConformanceMatrix/direct-archive-write'
  'TestProductionAuthorityV7OperationsConformanceMatrix/disposable-down'
  'TestProductionAuthorityV7OperationsConformanceMatrix/epoch-recovery-apply-vs-replacement'
  'TestProductionAuthorityV7OperationsConformanceMatrix/epoch-recovery-cancelled'
  'TestProductionAuthorityV7OperationsConformanceMatrix/epoch-recovery-resolved'
  'TestProductionAuthorityV7OperationsConformanceMatrix/epoch-recovery-suffix-0'
  'TestProductionAuthorityV7OperationsConformanceMatrix/epoch-recovery-suffix-1'
  'TestProductionAuthorityV7OperationsConformanceMatrix/epoch-recovery-suffix-64'
  'TestProductionAuthorityV7OperationsConformanceMatrix/fresh-import-fence-or-reimport'
  'TestProductionAuthorityV7OperationsConformanceMatrix/genesis-empty-in-place'
  'TestProductionAuthorityV7OperationsConformanceMatrix/genesis-five-stage-crash-matrix'
  'TestProductionAuthorityV7OperationsConformanceMatrix/genesis-fresh-target'
  'TestProductionAuthorityV7OperationsConformanceMatrix/normal-epoch-illegal-cancel-or-next'
  'TestProductionAuthorityV7OperationsConformanceMatrix/normal-epoch-old-holder-death-cancelled'
  'TestProductionAuthorityV7OperationsConformanceMatrix/normal-epoch-resolved'
  'TestProductionAuthorityV7OperationsConformanceMatrix/normal-epoch-terminal-race-catchup'
  'TestProductionAuthorityV7OperationsConformanceMatrix/partial-source-provider-coverage'
  'TestProductionAuthorityV7OperationsConformanceMatrix/prefix-decision-journal-context-splice'
  'TestProductionAuthorityV7OperationsConformanceMatrix/premature-genesis-open-serving'
  'TestProductionAuthorityV7OperationsConformanceMatrix/production-down'
  'TestProductionAuthorityV7OperationsConformanceMatrix/real-pitr-response-loss-refusal'
  'TestProductionAuthorityV7OperationsConformanceMatrix/same-holder-once-row-bypass'
  'TestProductionAuthorityV7OperationsConformanceMatrix/source-retirement-seal-postseal-archive'
  'TestProductionAuthorityV7OperationsConformanceMatrix/staging-closed-is-not-completion'
  'TestProductionAuthorityV7OperationsConformanceMatrix/staging-normal-abort'
  'TestProductionAuthorityV7OperationsConformanceMatrix/staging-recovery-abort'
  'TestProductionAuthorityV7OperationsConformanceMatrix/staging-release'
)
authority_tag_expected_text="$(printf '%s\n' "${authority_tag_expected[@]}" | LC_ALL=C sort)"
authority_tag_actual_text="$(go test -tags=c12_authority "${authority_tag_packages[@]}" -list "$authority_tag_pattern" | grep -E '^Test[A-Za-z0-9_]+$' | LC_ALL=C sort)"
[[ "$authority_tag_actual_text" == "$authority_tag_expected_text" ]]
authority_tag_json="$(mktemp)"
trap 'rm -f -- "$authority_tag_json"' EXIT
go test -tags=c12_authority "${authority_tag_packages[@]}" -json -run "$authority_tag_pattern" -count=1 -timeout 45m >"$authority_tag_json"
for name in "${authority_tag_expected[@]}"; do
  run_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"run\""){c++} END{print c+0}' "$authority_tag_json")"
  pass_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"pass\""){c++} END{print c+0}' "$authority_tag_json")"
  skip_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"skip\""){c++} END{print c+0}' "$authority_tag_json")"
  [[ "$run_count" == 1 && "$pass_count" == 1 && "$skip_count" == 0 ]]
done
for name in "${authority_tag_case_expected[@]}"; do
  run_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"run\""){c++} END{print c+0}' "$authority_tag_json")"
  pass_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"pass\""){c++} END{print c+0}' "$authority_tag_json")"
  skip_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"skip\""){c++} END{print c+0}' "$authority_tag_json")"
  [[ "$run_count" == 1 && "$pass_count" == 1 && "$skip_count" == 0 ]]
done
! awk 'index($0,"\"Action\":\"skip\"") && (index($0,"\"Test\":\"TestProductionAuthorityConformanceMatrix/") || index($0,"\"Test\":\"TestProductionAuthorityV7OperationsConformanceMatrix/")) { found=1 } END { exit(found ? 0 : 1) }' "$authority_tag_json"
rm -f -- "$authority_tag_json"
trap - EXIT
```

Expected: FAIL because the production v7 operations conformance harness is absent. Once GREEN, every one of the 34 literal full subtest names must run+pass exactly once with zero skip, any skip event below either aggregator tree fails, and a deleted/renamed case cannot be hidden by a passing parent.

- [ ] **Step 7: GREEN — implement isolated tenant and crash matrix**

The native authority parent alone opens the fixed protected provider credential selected by the validated runner profile, requires the conformance tenant marker and rejects production namespace, then creates run-scoped provider sessions and WAL records. No script reads `/run/secrets/talenro-c12-authority/provider.json` or any provider configuration. Implement two non-overlapping layers. The `c12_authority` `_test.go` diagnostic layer may call Task 6A's sole orchestrator through a test-file-only constructor that injects already-validated typed artifact facts and real isolated production dependencies；the symbol is absent from non-test builds, emits no receipt/evidence/attestation and cannot satisfy a production entrypoint startup. The production native parent exposes no such constructor and must fail before `NewOrchestrator` until Task 8 supplies a final-tree `ScanReceiptV1` and canonical bundle；the finite Bash stages can only return bounded case results and cannot start the production entrypoint themselves. Static/build tests prove the test constructor and diagnostic capability are unreachable from `cmd/nodecontrol-authority-operations`.

For Task 8's authoritative run, every positive/crash/PITR case starts only the dedicated OCI entrypoint whose exact manifest digest and binary SHA-256 are explicit components of the validated scanner receipt. Before each start, the native parent validates the canonical supply-chain bundle against every receipt reference/body/digest, rescans the read-only OCI layout, requires its config/layer set and embedded entrypoint binary SHA-256 to equal those receipt components, recomputes the combined digest, loads by content digest and invokes by immutable image ID/digest；after exit, that same parent re-inspects the runtime image/process receipt and requires the same tuple under the authority-scope process policy. The parent alone uses its attested final-tree snapshot, proves commit/tree/spec/catalog equality to the receipt, and mounts that snapshot plus the receipt at the two fixed read-only Task 6A paths；the entrypoint revalidates both before `NewOrchestrator`. Bash receives none of the snapshot、receipt、bundle、layout、mount、provider、scanner or runtime handles. All application/PostgreSQL/Redis/NATS dependency layouts are the receipt-bound local digest-addressed layouts；container/build network is disabled and pull/tag/registry fallback is a hard failure. Reject missing/changed bundle bodies, body/reference/digest disagreement, component/combined disagreement, same binary under another manifest, same manifest with another binary, a caller-selected combined digest, missing/read-write/wrong implementation-tree mount, an in-process API path, `go run`, PATH-resolved binary, mutable tag, rewritten layout, alternate entrypoint or direct host build. Both the diagnostic matrix now and the receipt-pinned matrix in Task 8 drive baseline `Reserve/Finalize/Abort/Inspect/Head`, semantic reservation, both genesis modes through runtime registration/durable attempt/Prepare→activation proof→Complete→completion proof→PrepareRelease/release proof→Open, normal epoch Resolve-vs-cancel terminalization, Fence/challenge/Inspect, epoch recovery, staging, source retirement/archive and disposable Down through every observable seam. Assert DB visibility matches provider/attestor terminal records；no ordinary path opens before the exact genesis Open or normal-epoch terminal application/catchup；prefix slot/semantic reservation/once-row exact retry never aliases；no changed effect/body reuses an operation；all provider/attestor calls and every signer except the one exact registered-Down authorizer stay outside DB transactions, and that authorizer occurs once only after pristine `1/0` capture with the actual txid.

- [ ] **Step 8: GREEN — implement real PITR refusal proof**

Create backups before and after runtime registration, durable attempt and every Prepare/activation/Complete/completion/PrepareRelease/release/Open boundary for both genesis modes；before/after normal-epoch intent, Request, application+resolution, proof, Resolve/cancellation CAS, terminal application and catchup；before a real node/operator revoke and desired disable；before/within epoch terminal recovery；and at each staging held/recovery/outcome phase. At each provider/attestor response-loss seam, promote the restored run-owned PostgreSQL timeline while retaining external journals. In Task 6B restart only the non-authoritative diagnostic harness identity and require Inspect to converge without duplicate call/effect；implement the production native-parent path and prove it refuses missing/non-final receipt before any finite Bash stage. In Task 8 repeat every seam by restarting the same digest-pinned OCI entrypoint and require its `inspect` subcommand to converge to the exact same terminal chain. Assert baseline listeners/signers remain closed and old cert/desired cannot serve；then run the exact terminal→intent→suffix→proof→transcript→arbitration→optional Fence→restricted application→fresh candidate→catchup/reconcile recovery and staging recovery/Abort flows. Separately build a complete retired/archive source and a disposable pristine fixture for Down. A destructive-restore attempt with unsupported `trust_bundle_publish`/`operator_authorizer_change` must end only in signed `FreshTargetInventoryV1`/`FreshV7StagingEvidenceV1`, closed route and `restore_incomplete`; it must never produce a completion/open record. Diagnostic records omit pre/post execution-image claims；Task 8 alone records them from the final pinned image. Never alter provider or attestor history to match the old DB.

- [ ] **Step 9: Run non-authoritative diagnostics and freeze the final native authority parent**

Before implementing the final parent, add `TestAuthorityParentSuccessSignsOnlyFromPostCleanupPublication`, `TestAuthorityParentAbortCannotMintPublication`, `TestAuthorityParentCleanupFailureCannotSignOrPrepareChannel`, `TestAuthorityParentPostCleanupRecoveryIsSelectorFree`, `TestAuthorityParentPostCleanupResponseLossConvergesExactTransaction` and `TestAuthorityParentHasNoRawAttestorSenderNonceOrTransactionSurface`. They prove success/abort origin is fixed before cleanup, only an exact leaf receipt can mint same-generation post-clean publication, active crash defaults to abort, attestation response loss never re-signs, and terminal-inactive/next is selected only after exact committed-sender finalization.

At this step extend the authority semantic script with exactly nine more `internal/c12evidence/authority_test.go` roots：`TestAuthorityParentSemanticGateExactCoversDeclaredTests`、the six roots just named、the already frozen `TestAuthorityParentRecoveryDispatchIsClosedAndNoFallback` and `TestAuthorityParentCleanupExactAbsenceAndAppendOnlyRetention`. The cleanup aggregate owns the exact resource/PITR/cert/credential/provider absence checks while proving append-only journal/Fence/archive/tombstone retention. In the same change, edit the existing `TestAuthoritySemanticGateExactCoversDeclaredTests` table to expect the complete 14-entry script while requiring its initial five entries to remain byte-for-byte equal to the Step 1 core partition. The parent meta exact-compares the same complete union, exact-covers every top-level `TestAuthorityParent*` definition and applies the same unique-owner/go-types/reachable-helper/no-skip rules；neither script nor either meta may accept a unilateral/reordered extension. No parent root is build-tagged；the two tagged production conformance aggregators remain under their separate Linux gate.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-authority-semantic-gate.ps1`

Expected: FAIL because the high-level post-clean authority parent facade and its fixed runnerprofile continuation are not yet wired.

The strict authority token created here sets `BootstrapLauncherCatalogRole=c12_runner_launcher_linux_amd64`; its `InstallReceiptPolicy` excludes both launcher and snapshot scripts. The sole operator entry is the fixed Linux launcher, whose guard-selected versioned helper is the actual authority parent. Direct versioned-helper machine-mode invocation is not a public surface.

Run the tagged real-provider/PostgreSQL diagnostic harness from Step 6 now that Step 7–8 are implemented. It may inspect the two-root-identical OCI layout produced by the focused Plan 08 build reproducibility test and must run production-entrypoint startup-refusal cases, but it must not positively start that entrypoint because no final-tree receipt exists. It writes bounded non-attested per-case/ownership/cleanup digests to `/var/lib/talenro-c12-authority/v7-operations-diagnostic.json` and cannot create、serialize or retain `ScanReceiptV1`, `AuthorityFenceEvidenceV1` or `C12ScopeEvidenceV1` from an uncommitted tree. A static test rejects any diagnostic record consumed by the scope/completion validators.

Implement `testdata/c12/authority/runner-profile.v1.json` as the strict `Role=authority` generic policy instance. `AllowedTrackedScriptTokens=[scripts/verify-c12-authority-v7-operations.sh,scripts/verify-c12-authority-fence.sh]` in that order；`RequiredToolRoles` is the fixed ordered Linux Bash/Git/Docker/PITR/helper/producer/trusted-time-client role set；`HelperCatalogRole=c12_runner_helper_linux_amd64`；`RecoverySlotPurpose=authority_cleanup_recovery`；`ExternalRootPurpose=authority_work_root`；`EvidenceRootPurpose=authority_nonpublishing_evidence`；`ProviderCredentialPurpose=authority_trusted_time_provider_pitr`；its sole binding is `{TransactionKind: authority_scope, Access: sender, SlotPurpose: authority_scope_next, SequencePolicy: next}`；and `InstallReceiptPolicy` closes Linux/amd64 plus helper、Bash interpreter、Git/Docker/PITR/producer/trusted-time-client tool catalog roles and application-control purposes, never either snapshot script token. It contains no machine ODB/root/slot/absolute path/receipt/attestor/transaction fact. Task 6B validates the policy and synthetic session fixtures；Task 7 relock/provisioning later calls only non-consuming `VerifyFixedNativeRunnerInstallation(..., authority)`. In the final run, before any ordinary WAL/run-root/snapshot/producer/provider/PITR/credential/certificate resource create, `run-authority-final` calls `OpenInstalledNativeRunnerSession(..., authority)` exactly once；the Linux adapter opens `/var/lib/talenro/c12/runner-bootstrap.v1.sealed`, validates the exact helper/interpreter/tool receipts, consumes the opaque session and verifies every policy/bootstrap projection. Its unexported authority facade generates all stable parent/O_EXCL child reservations, binds the opaque fixed cleanup plan from those reservations plus the session without receiving a deleter/adapter set, and executes `BeginProtectedWALKeyBootstrap(ctx,session,plan) → CreateOrRecoverProtectedWALKeyBootstrap → SealAndSelectProtectedWALKeyBootstrap({IssuedAt,ExpiresAt})`. Intent selection/file+directory fsync precedes creation/exact-reopen of the one deterministic nonexportable key；the runnerprofile bootstrap fills every machine/run/slot/provider/parent/resource/WAL field, seals+file/directory-fsyncs the capsule, guard-selects it and returns opaque selected ownership without exposing a binding、intent、key handle or signer. A pre-intent claimed generation routes only to zero-resource `abort_unstarted`; a selected intent permits only its one provisional key and routes only to exact abort/destroy；an unselected capsule has no authority. The facade then invokes the sole owning consumer `OpenSelectedOwnershipWAL(ctx,selected)`；it internally exact O_EXCL-creates/reopens the planned WAL, writes+file/directory-fsyncs bootstrap plus observed file-identity actual, consumes selected ownership and returns only sealed `SelectedOwnershipWAL`. Only after that return may the parent append through the sealed WAL, create the run root, capture the exact committed snapshot and create/WAL-record two repository-external exclusive build children with independent 128-bit lowercase-hex names and identities before directly launching its private producer. Every zero-argument restart first exact-dispatches `abort_unstarted|abort_intent|role_dispatch`; an abort error never permits authority recovery or fresh Begin. The sole operator command is:

```text
/usr/libexec/talenro-c12/talenro-c12-runner-launcher run-authority-final
```

The installed helper ELF、Bash interpreter and private producer/trusted-time client roles are exact catalog/toolchain/authority-runner-profile members. The parent retains the capture/session handles, builds/scans both complete roots byte-identically, retains only helper-designated root A read-only, constructs the purpose-bound trusted-time client only from the protected provider handle plus the exact `authority_linux_amd64` native-client tuple, pins the installed `/usr/bin/bash` interpreter and derives both tracked scripts only from its committed snapshot. Before each fixed stage it requires the exact committed script blob bytes、Git mode、snapshot root/file identity、no-follow regular-file and one-link proof, sets the locked snapshot cwd and closed sanitized environment, then invokes only `/usr/bin/bash --noprofile --norc <parent-derived absolute snapshot script>` with the bounded result channel as standard stdin/stdout and a closed role enum. The parent-derived、identity-verified absolute snapshot script is the sole path placed in argv；no caller-selected repo/profile/script/provider-config/receipt/bundle/layout/WAL/run-root/result/evidence path、digest or numeric handle enters argv/env, and an installed script or script receipt cannot substitute. The shell cannot invoke a provider client、canonicalizer、signer、scanner、OCI loader/mounter or deleter, cannot read/compare any other channel transaction, and a grandchild writer/held-open stdio is rejected. Direct Bash、`scan-closed-set`、private-child invocation、one-root/prebuilt materialization、alternate producer、PATH binary、`go run` or ambient endpoint rejects.

The guard-selected cleanup-only capsule remains the sole crash-recovery authority through the non-resumable authority workload. After every helper/launcher/power seam, including capsule candidate/guard/reread、WAL/run-root/build actuals、producer return、each provider/PITR/resource intent/actual and each stage result, a fixed launcher successor first proves the old bound containment/process namespace absent and may start only the same-generation exact-clean helper；expiry or trusted-time outage still permits delete-only cleanup and never workload/signing. Only after the complete matrix、typed receipt/bundle、trusted-time and nested/outer unsigned semantic candidates validate through the sealed producer chain does package-private `BindAuthorityFinalWorkloadResult` consume the same session plus `AuthorityUnsignedWorkloadFacts|FinalClosedSetProjection`; only the resulting `AuthorityFinalWorkloadResult` lets the parent guard-select success-origin cleanup and transfer the canonical result plus three attestors into inaccessible state. Any explicit failure or active crash before that CAS selects/defaults to abort-origin cleanup and destroys the attestors.

Both origins first prove snapshot/build/PITR/credential/run/provider/session and every declared resource absent, fsync terminal WAL and prepare the assignment-identical leaf receipt. The matching finalizer stages/fsyncs the strict tombstone and first selects `authority_cleanup_finalization_pending(receipt,fixed_target)` (or abort pending), then leaf Recover/Commit、key destruction/absence、tombstone unlink+parent-directory fsync+exact absence all complete while pending remains selected；only afterward does the final guarded CAS/reread select the target. Abort targets terminal-inactive/next and never signs；success targets same-generation `authority_post_cleanup_publication`, with no post-target tombstone work. Only that successor may revalidate facts, sign/self-verify authority+outer evidence, bind sender and drive `Prepare/Commit/InspectSender`; response loss never re-signs. Exact committed sender inspection alone permits the final next-generation CAS and reservation clear. Cross-domain/sender reads reject；pre-Commit failure publishes none, post-Commit retains only the immutable transaction. Crash tests cover origin/dispatch、resource/WAL/receipt/pending/key、unlink/parent-fsync/absence、target CAS、both signatures、channel/final CAS/reservation clear, requiring selector-free convergence, no re-sign and at most one live parent.

Implementation-time expected: the tagged diagnostic harness executes every real-provider/PostgreSQL case, all production entrypoint attempts without the final receipt fail before `NewOrchestrator`, every reserved provider/attestor record is terminal/exactly recoverable, both five-stage genesis modes、normal resolved/cancelled epoch、epoch recovery、source、staging、Down and every PITR/response-loss refusal succeed diagnostically, cleanup succeeds, and no reusable receipt/scope file is written. After every implementation/evidence/scanner file is committed, Task 8 first creates the final full receipt/bundle and only then runs `run-authority-final`；that is the first time the exact OCI entrypoint executes the complete matrix with matching pre/post image receipts and may produce unexpired production evidence. Parent tests exact-cover public help and reject direct-shell/PATH/go-run/provider-config/fake/ambient-cwd attacks, replacement WAL/capsule identity and every cleanup seam.

Before either owning tagged suite is run, use the single frozen fail-closed command above and require its exit status zero. A missing/untracked file, tag on a later line, wrong literal tag or failure of the first package blocks Task 6B and cannot be masked by the second check/package.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-authority-semantic-gate.ps1`

Run: `go test ./internal/c12runnerprofile -run '^(TestNativeRunnerLaunchReservationCrashSeams|TestNativeRunnerLauncherDeathCannotLeaveTwoLiveParents)$' -count=1 -timeout 45m`

Run: `go vet ./cmd/talenro-c12-runner-launcher ./cmd/talenro-c12-runner-helper ./internal/c12runnerprofile ./internal/c12evidence`

Expected: PASS；the Linux launcher's OS-valid help contains `run-authority-final` exactly once, the versioned helper's unbound help omits it, private finalizer/producer/client modes remain absent, launcher death cannot leave a live authoritative tree, and the selected helper parent owns the full nested/outer signature and cleanup/channel sequence.

- [ ] **Step 10: REFACTOR — verify exact cleanup without deleting append-only evidence**

Run diagnostic cleanup assertions inside the owning tagged test. Rerun `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-authority-semantic-gate.ps1`；final cleanup verification is frozen inside `run-authority-final` and the exact `TestAuthorityParentCleanupExactAbsenceAndAppendOnlyRetention` aggregate, with no operator-callable Bash cleanup/evidence-path command.

At Task 6B assert the scope path remains absent；at Task 8 run the frozen command against real final evidence. Expected: PASS after exact PITR DB/cert/credential/state/fixture cleanup；provider, semantic-ID journal, Fence once-row and archive records remain retained as non-secret append-only digests；source/down tombstones remain permanent；and no non-run operation was inspected or changed.

- [ ] **Step 11: Commit the authority/PITR gate**

```bash
git add internal/c12evidence/authority.go internal/c12evidence/authority_test.go internal/nodecontrol/authority/production_conformance_test.go internal/nodecontrol/operations/production_conformance_test.go scripts/verify-c12-authority-v7-operations.sh scripts/verify-c12-authority-fence.sh scripts/invoke-exact-authority-semantic-gate.ps1 testdata/c12/authority/runner-profile.v1.json internal/c12evidence/runnerhelper.go internal/c12evidence/runnerhelper_test.go cmd/talenro-c12-runner-helper/main.go cmd/talenro-c12-runner-helper/main_test.go
git commit -m "test: add production Authority Protocol v7 PITR gate"
```

### Task 7: Four-scope completion manifest and license gate

**Files:**
- Create: `internal/c12evidence/completion.go`
- Create: `internal/c12evidence/completion_test.go`
- Create: `internal/c12completionpublication/operator.go`
- Create: `internal/c12completionpublication/operator_test.go`
- Modify: `cmd/talenro-artifact-scan/main.go`
- Modify: `cmd/talenro-artifact-scan/main_test.go`
- Modify: `internal/c12evidence/runnerhelper.go`
- Modify: `internal/c12evidence/runnerhelper_test.go`
- Modify: `cmd/talenro-c12-runner-helper/main.go`
- Modify: `cmd/talenro-c12-runner-helper/main_test.go`
- Modify: `deploy/c12/compose.outer.yaml`
- Modify: `testdata/c12/integration-manifest.v1.json`
- Modify: `testdata/c12/toolchain-lock.v1.json`
- Modify: `testdata/c12/platform/trusted-time-provider-profile.v1.json`
- Modify (created by Plan 08): `testdata/c12/windows/runner-profile.v1.json`
- Modify (created by Plan 08): `testdata/c12/diagnostic/runner-profile.v1.json`
- Modify: `testdata/c12/platform/runner-profile.v1.json`
- Modify: `testdata/c12/authority/runner-profile.v1.json`
- Create: `testdata/c12/completion/runner-profile.v1.json`
- Create: `scripts/invoke-exact-completion-semantic-gate.ps1`
- Test (read-only; owned by Tasks 4–5): `scripts/invoke-exact-platform-semantic-gate.ps1`
- Test (read-only; owned by Task 6B): `scripts/invoke-exact-authority-semantic-gate.ps1`
- Test (read-only input; never rewritten or staged by this task): `testdata/c12/artifact-catalog.v1.json`
- Test (read-only input; created and owned by Plan 08, never rewritten or staged by this task): `testdata/c12/launcher-import-closure.v1.json`
- Test: `internal/c12evidence/completion_test.go`

**Interfaces:**
- Consumes: exactly four scopes、two explicit nested envelopes、the fixed spec validator、validated receipt/bundle/license/document/provider projections、Task 4's signed time semantics and sealed `AuthenticatedTrustedTimeClient` provenance, plus Plan 08's native profile/bootstrap/install/phase/audit and low-level channel/state machinery. Each zero-argument public mode first uses the non-consuming phase router；create fresh-Opens only `idle`, revalidate fresh-Opens only inactive `retained_complete`, and only the matching active phase obtains same-generation recovery. Already-terminal create uses only the separate read-only audit after old-reservation clear. `completion.go` owns semantic validation only. `internal/c12completionpublication` is the sole private join point：it alone receives the coordinator/receiver/binding/set/sink/view, derives fixed policy/provider facts, authenticates child time results, privately holds the validated token and stages canonical bytes. Public commands receive only opaque high-level handles/operators/sessions. `CompletionPolicy.ExpectedTrustedTimeSourceIdentity` is derived from the committed provider profile and is never caller-overridable；no command/client/policy/path/store/raw-state injection exists.
- Produces: `C12CompletionManifestV1`, opaque-token `ValidateCompletionInputs`/`CanonicalValidatedCompletionManifest`, the sole `internal/c12completionpublication` composite operator, the strict `Role=completion` profile instance, the catalog/toolchain-locked native helper public parent modes `completion-consolidate-create`/`completion-revalidate`, and the parent-authenticated private prebuilt artifact-scan completion child. P09 only consumes Plan 08's channel schema/state machine and B10-frozen private manifest wire decoder；it adds no second transaction/JCS/domain or low-level sink. No operator-facing path-list、provider-profile flag or `go run` completion path exists.

- [ ] **Step 1: RED — add exact-four and same-build tests**

```go
func TestCompletionRequiresEachScopeExactlyOnce(t *testing.T) {
	t.Parallel()
	inputs := validCompletionInputs()
	inputs.Scopes[3] = inputs.Scopes[0]
	if _, err := ValidateCompletionInputs(inputs); !errors.Is(err, ErrScopeSet) {
		t.Fatal("completion accepted duplicate and missing scope")
	}
}

func TestCompletionRejectsScopeRunIDReuseAcrossDistinctScopes(t *testing.T) {
	t.Parallel()
	for i := 0; i < len(validCompletionInputs().Scopes); i++ {
		for j := i + 1; j < len(validCompletionInputs().Scopes); j++ {
			inputs := validCompletionInputs()
			inputs.Scopes[j].RunID = inputs.Scopes[i].RunID
			if _, err := ValidateCompletionInputs(inputs); !errors.Is(err, ErrScopeRunIDReuse) {
				t.Fatalf("completion accepted RunID reuse across scope indexes %d and %d", i, j)
			}
		}
	}
}

func TestCompletionRejectsCrossBuildSplice(t *testing.T) {
	t.Parallel()
	inputs := validCompletionInputs()
	inputs.Scopes[2].TrackedTreeDigest[0] ^= 1
	if _, err := ValidateCompletionInputs(inputs); !errors.Is(err, ErrBuildMismatch) {
		t.Fatal("completion accepted cross-build evidence")
	}
}

func TestCompletionRejectsFreshV7StagingAsLegacyUpgradeComplete(t *testing.T) {
	t.Parallel()
	inputs := validCompletionInputs()
	inputs.Authority.LegacyUpgradeStatus = "fresh_v7_staging_closed_as_complete"
	if _, err := ValidateCompletionInputs(inputs); !errors.Is(err, ErrRestoreIncomplete) {
		t.Fatal("completion promoted safe staging to completed legacy upgrade")
	}
}

func TestCompletionRejectsEveryAuthorityOperationsArtifactSplice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*CompletionInputs)
	}{
		{"scope-tuple", func(in *CompletionInputs) { in.Scopes[1].AuthorityOperationsArtifact.BinarySHA256[0] ^= 1 }},
		{"scope-combined", func(in *CompletionInputs) { in.Scopes[1].AuthorityOperationsBinaryAndImageDigest[0] ^= 1 }},
		{"scope-record-set", func(in *CompletionInputs) { in.Scopes[1].AuthorityOperationsSupplyChainRecordSetDigest[0] ^= 1 }},
		{"scope-bundle", func(in *CompletionInputs) { in.Scopes[1].SupplyChainEvidenceBundleDigest[0] ^= 1 }},
		{"scope-catalog", func(in *CompletionInputs) { in.Scopes[1].ArtifactCatalogDigest[0] ^= 1 }},
		{"scope-process-policy", func(in *CompletionInputs) { in.Scopes[1].ExpectedProcessImagePolicyDigest[0] ^= 1 }},
		{"scope-image-set", func(in *CompletionInputs) { in.Scopes[1].ImageSetDigest[0] ^= 1 }},
		{"scope-full-scan", func(in *CompletionInputs) { in.Scopes[1].ArtifactScanDigest[0] ^= 1 }},
		{"nested-platform-tuple", func(in *CompletionInputs) { in.Platform.AuthorityOperationsArtifact.ImageManifestDigest[0] ^= 1 }},
		{"nested-platform-combined", func(in *CompletionInputs) { in.Platform.AuthorityOperationsBinaryAndImageDigest[0] ^= 1 }},
		{"nested-platform-record-set", func(in *CompletionInputs) { in.Platform.AuthorityOperationsSupplyChainRecordSetDigest[0] ^= 1 }},
		{"nested-platform-bundle", func(in *CompletionInputs) { in.Platform.SupplyChainEvidenceBundleDigest[0] ^= 1 }},
		{"nested-platform-catalog", func(in *CompletionInputs) { in.Platform.ArtifactCatalogDigest[0] ^= 1 }},
		{"nested-platform-process-policy", func(in *CompletionInputs) { in.Platform.ExpectedProcessImagePolicyDigest[0] ^= 1 }},
		{"nested-platform-image-set", func(in *CompletionInputs) { in.Platform.ImageSetDigest[0] ^= 1 }},
		{"nested-platform-full-scan", func(in *CompletionInputs) { in.Platform.ArtifactScanDigest[0] ^= 1 }},
		{"nested-authority-tuple", func(in *CompletionInputs) { in.Authority.AuthorityOperationsArtifact.BuildReceiptDigest[0] ^= 1 }},
		{"nested-authority-combined", func(in *CompletionInputs) { in.Authority.AuthorityOperationsBinaryAndImageDigest[0] ^= 1 }},
		{"nested-authority-record-set", func(in *CompletionInputs) { in.Authority.AuthorityOperationsSupplyChainRecordSetDigest[0] ^= 1 }},
		{"nested-authority-bundle", func(in *CompletionInputs) { in.Authority.SupplyChainEvidenceBundleDigest[0] ^= 1 }},
		{"nested-authority-catalog", func(in *CompletionInputs) { in.Authority.ArtifactCatalogDigest[0] ^= 1 }},
		{"nested-authority-process-policy", func(in *CompletionInputs) { in.Authority.ExpectedProcessImagePolicyDigest[0] ^= 1 }},
		{"nested-authority-image-set", func(in *CompletionInputs) { in.Authority.ImageSetDigest[0] ^= 1 }},
		{"nested-authority-full-scan", func(in *CompletionInputs) { in.Authority.ArtifactScanDigest[0] ^= 1 }},
		{"scanner-policy-tuple", func(in *CompletionInputs) { in.Policy.ScopePolicies[3].ExpectedAuthorityOperationsArtifact.DependencyClosureDigest[0] ^= 1 }},
		{"scanner-policy-combined", func(in *CompletionInputs) { in.Policy.ScopePolicies[3].ExpectedAuthorityOperationsBinaryAndImageDigest[0] ^= 1 }},
		{"scanner-policy-record-set", func(in *CompletionInputs) { in.Policy.ScopePolicies[3].ExpectedAuthorityOperationsSupplyChainRecordSetDigest[0] ^= 1 }},
		{"scanner-policy-bundle", func(in *CompletionInputs) { in.Policy.ScopePolicies[3].ExpectedSupplyChainEvidenceBundleDigest[0] ^= 1 }},
		{"scanner-policy-catalog", func(in *CompletionInputs) { in.Policy.ScopePolicies[3].ExpectedArtifactCatalogDigest[0] ^= 1 }},
		{"scanner-policy-process", func(in *CompletionInputs) { in.Policy.ScopePolicies[3].ExpectedProcessImagePolicyDigest[0] ^= 1 }},
		{"scanner-policy-image-set", func(in *CompletionInputs) { in.Policy.ScopePolicies[3].ExpectedImageSetDigest[0] ^= 1 }},
		{"scanner-policy-full-scan", func(in *CompletionInputs) { in.Policy.ScopePolicies[3].ExpectedArtifactScanDigest[0] ^= 1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputs := validCompletionInputs()
			tt.mutate(&inputs)
			if _, err := ValidateCompletionInputs(inputs); !errors.Is(err, ErrBuildMismatch) {
				t.Fatal("completion accepted an authority operations artifact splice")
			}
		})
	}
}

func TestCompletionRejectsParallelCleanOperationsLicenseSet(t *testing.T) {
	t.Parallel()
	inputs := validCompletionInputs()
	inputs.Licenses = anotherCleanLicenseProjectionForSameArtifact(t)
	if _, err := ValidateCompletionInputs(inputs); !errors.Is(err, ErrLicenseMismatch) {
		t.Fatal("completion accepted a clean license set not projected from the scanned typed records")
	}
}
```

Add RED matrices for every explicit nested repo/tree/spec/toolchain/release/timestamp/trusted-time field, the six independent outer/nested expiry candidates and `CompletionPolicy.NotAfter`. Add an online time-source fake that signs the exact per-command purpose/challenge；wrong identity/purpose/challenge, replayed or zero stored nonce, changed stored canonical input digest, missing/tampered embedded creation envelope, embedded-envelope/digest/sequence/source mismatch, `CreatedAt != IssuedAt`, stale/equal fresh revalidation sequence, floor rollback, host wall-clock forward/backward jump and source timeout must fail or remain irrelevant exactly as frozen above. Independently splice a valid different source into Platform、Authority、embedded creation and fresh revalidation evidence and require every case to fail constant-time against `CompletionPolicy.ExpectedTrustedTimeSourceIdentity`; one source's statement may not judge another source's `ExpiresAt`. No test passes `time.Now()` into `completion.go`.

Add completion-runner-profile RED vectors through Plan 08's loader：wrong role/schema/token、unknown/duplicate/null/default field、missing/extra/reordered tracked script token or tool role、wrong helper catalog role/recovery/external/evidence-root/provider purpose/install-receipt policy、a sender binding, or a channel set other than the exact ordered receiver bindings `{windows_scopes, windows_scopes_next, next}`、`{platform_scope, platform_scope_next, next}`、`{authority_scope, authority_scope_next, next}`. Separately splice the sealed canonical ODB/evidence/external root、absolute native helper/private-child/interpreter/tool path or native install receipt、provider credential/attestor capability or actual `ChannelNamespaceDigest`/`ChannelSlotIdentityDigest`/next `SlotSequence`; any script path field or script install receipt also rejects because completion has `AllowedTrackedScriptTokens=[]` and receipts cover native executables only. The Windows adapter may open only `HKLM\SOFTWARE\Talenro\C12\RunnerBootstrapV1`; copied/raw receipt/bootstrap、foreign/reused opaque session or mismatched Windows helper/child identity must reject before channel/provider/evidence-root access. Package-local channel RED tests still prove the low-level receiver cannot list/Inspect、choose nonce or adopt prepared/quarantined/wrong-sequence input and that every mutation requires the same held set. Cross-package RED/compile-fail tests instead prove production command/helper code can name only `CompletionPublishPendingHandle` and call `AdoptExpectedEvidence(ctx)`；it cannot name or recover a receiver、binding、set、slot、sequence、nonce、lease、sink or view, and there is no facade overload accepting any of them.

Add `TestCompletionPublishPendingPrecedesAdoptionSet`、`TestCompletionPublishPendingFreshProcessRecoveryDoesNotOpenAgain`、`TestCompletionPublishPendingUsesSingleRoleRecoverySlot`、`TestCompletionRoleStateRejectsAuthenticatedRollback`、`TestCompletionRoleStateGuardCrashSeams`、`TestCompletionAdoptionSetCrashRecovery`、`TestCompletionMaterializedViewExactEightMembers`、`TestCompletionOpaqueHandleHidesReceiverSetBindingSinkAndView`、`TestCompletionLowLevelChannelCallsOwnedOnlyByCompositePackage`、`TestCompletionRawRunnerprofileCallsOwnedOnlyByCompositePackage`、`TestCompletionCreateDispatchIsClosedAndNoFallback`、`TestCompletionPublicationOperatorHasNoCallerPathKindOrDigest`、`TestCompletionPublicationOperatorIsSoleStageCanonicalManifestCaller`、`TestCompletionPublicationSinkManifestProjectionMatchesFrozenWireShape`、`TestCompletionPublicationProgressCrossPersistsEachOuterSeam`、`TestCompletionRecoverPublicationProgressClosesEveryCrashSeam`、`TestCompletionInnerCleanupAdapterPairsBothCapabilities`、`TestCompletionInnerCleanupExcludesPublishState`、`TestCompletionInnerCleanupDoesNotTerminalizeLaunchGeneration`、`TestCompletionValidationRejectsCallerPolicyClientAndRawChildBytes`、`TestCompletionFinalizeReadyBindsImmutableRetainedPayloadDigest`、`TestCompletionFinalizeReadyHasNoDigestCycleOrPrematureCell`、`TestCompletionRecoverFinalizationBeforeAndAfterGuardCAS`、`TestCompletionTerminalFinalizationAuditIsPathlessReadOnly`、`TestCompletionTerminalAuditReceiptCannotDriveRevalidation`、`TestCompletionTerminalAuditSurvivesCompletedRevalidations`、`TestCompletionModePhaseCrossClaimRejects`、`TestCompletionRetainedTransitionHasNoSplitBrainWindow`、`TestCompletionRetainedCompleteAdvancesLaunchGenerationOnce` and `TestCompletionRetainedCompleteIsImmutable`. The role record must reach durable `publish_pending(N)` before the composite returns its handle. `handle.AdoptExpectedEvidence(ctx)` then privately Begin/Recovers one exact set and drives the three P08 set-bound adoptions；only after all actuals may `BindCompletionPublicationOperator(handle)` privately acquire the sink/view. The operator must build inputs only from that view and fixed authenticated dependencies, retain the semantic token internally and be the sole Stage caller. Inject real fresh-process crashes from Open through every **pre-E2** guarded intent/actual seam；restart first uses the closed high-level create dispatch and then only its one matching `RecoverCompletionPublishPending(ctx)`/terminal-audit target, the handle's idempotent narrow methods and rebound operator recovery, never another Open、caller locator、raw coordinator or low-level set. Copy each earlier valid phase cell over either A/B cell and mutate/cross-swap guard identity/epoch/digests；all cases return zero authority or quarantine, never a second parent or restored publish/ACK capability. Post-E2 cases are tested only through old-reservation absence/clear plus terminal audit, and active revalidation only through its distinct handle.

`TestCompletionRecoverPublicationProgressClosesEveryCrashSeam` must use a manifest whose signed trusted-time envelope、nonce and canonical bytes would differ if rebuilt. It crashes after the fixed set-pending、outer-CAS and set-actual steps of reciprocal `manifest_intent`, and with the reserved temp absent、zero-length、at every partial prefix、full-but-unflushed、file-fsynced and immediately before/after both sides of `manifest_actual`. It additionally power-cuts after both actuals while simulating loss of the un-fsynced temp directory entry, and before/after rename、parent-directory fsync and each reciprocal `manifest_durable` side. When inner authority remains, fresh continuation must first take only `session_cleanup_recovery`, finish any set-first state barrier without filesystem work, exact-clean/finalize and reread `complete`; the outer continuation may then invoke only rebound `RecoverPublication`. It must exact-rewrite or verify the durable bytes, recreate the reserved temp/rename when necessary, preserve their digest/nonce/time envelope byte-for-byte and leave provider/private-child/validator/canonicalizer call counts unchanged from intent selection onward. Zero、more-than-64-KiB、outer-first、one-sided-unrepresented、digest/view/set/identity-spliced intents and a foreign temp identity all fail closed. The test also proves the intent barrier precedes the first temp write, both byte copies remain through reciprocal `manifest_actual`, and they disappear only after reciprocal `manifest_durable` while the digest/identities remain.

Also add `TestCompletionValidationChildResponseStrictWire`, `TestCompletionValidationChildResponseMatchesConsumedRuntimeResult`, `TestCompletionValidationChildResponseCannotMintSemanticToken`, `TestCompletionValidationChildRuntimeIsOnlyCapabilityRoute`, `TestCompletionRetainedRevalidationFailureClosesWithAuthenticatedAbort`, `TestCompletionRetainedRevalidationOutcomeRaceIsSingleWinner` and `TestCompletionTerminalAuditSurvivesAbortedRevalidations`. They compile the real P08 runtime、c12evidence inherited adapter/wire decoder and composite, inject failure/crash at every Bind/run/result-consume/fetch/decode/cleanup/outcome-intent seam, and require zero token on raw/mutated response plus one closed inactive generation for either completed or aborted outcome.

The same RED matrix requires composite-owned `BeginCompletionInnerCleanup`/`InspectCompletionInnerCleanupRecoveryDispatch`/`RecoverCompletionInnerCleanup`/`RecoverCompletionInnerFinalization`/`FinalizeCompletionInnerCleanup` to pair one runnerprofile `CompletionInnerCleanupHandle` with one c12evidence cleanup-only capability inside an opaque `CompletionInnerCleanupSession` only on the three handle-producing begin/validation-recovery/cleanup-recovery routes；neither raw half escapes and runnerprofile never imports c12evidence. Every fresh or recovered owner first calls the five-way Inspect：only `begin_required` may call Begin and yields a validation-required session；only `session_validation_recovery` may Recover a validation-required session；only create `session_cleanup_recovery` may finish an authenticated set-first reciprocal manifest-intent barrier and Recover a cleanup-only session；only `finalization_recovery` may call no-session finalization；and `complete` returns no capability and permits only outer progress recovery. Each session is returned only after the coordinator has durably converged `protected_key_intent→cleanup_capsule` and selected one nonzero handle. Its sealed resume class is enforced by the composite：validation-required alone may invoke provider/private-child/validator and Stage, while cleanup-only rejects all four and may only exact-clean/finalize before rebound `RecoverPublication`. Both exact-cover only the temporary objects and exclude the retained root、manifest、marker and channel reservations. Finalization recovery returns no session and can close the stored receipt/key/target after the key is already absent. After manifest durability and `handle.AcknowledgeExpectedEvidence(ctx)`, `handle.FinalizeRetainedComplete(ctx)` privately freezes an exact E2 target-transition digest plus the separate immutable payload/finalization-set audit digest, runs E→E1 Prepare/Recover and proof-bound E1→E2 Commit, and returns no low-level proof. E2 ends same-generation recovery. A later pathless `InspectCompletionRetainedFinalization(ctx)` returns only the distinct audit receipt after old-reservation clear and validates authenticated lineage across closed completed/aborted revalidations；`OpenCompletionRetainedRevalidation`/`RecoverCompletionRetainedRevalidation` alone return the cleanup/validation-capable revalidation handle. A failed revalidation must exact-clean and select the fixed abort target through its receipt-taking finalizer；after `finalization_pending`, only no-session `RecoverCompletionInnerFinalization` can consume the generation and append the non-affirming edge. It may not call the success finalizer, accept caller outcome bytes or leave `retained_complete_active` wedged.

Inject crashes before/after outer creation、adoption/publication edges、inner plan/core/intent candidate/write/fsync/first CAS、key CreateOrRecover/abort/absence、restart marker/replacement capsule、selected-WAL binding marshal/store/recover/open/bootstrap/actual/file+parent fsync、resource cleanup、`finalization_pending`、leaf Recover/Commit、key absence、tombstone unlink/parent-directory fsync/exact absence、the **final** target CAS/reread、manifest/marker、ACK、both frozen digests、E1 proof/E2 guard、reservation/audit and child runtime/wire seams, once under create、completed and both abort owners. Before first inner CAS key/WAL/resource/provider/child counts are zero；no validation session returns before WAL readiness. Pre-manifest selected state routes only validation recovery. A capsule plus either reciprocal manifest-intent side routes only create cleanup recovery；it finishes the state barrier from stored bytes without filesystem/provider/child/validator/canonicalizer, exact-cleans through pre-target tombstone retirement, then reaches actionless complete and stored-byte `RecoverPublication`. Genuine cleaned pre-manifest retry without set pending is Begin, never false complete.

The completion-profile matrix also removes/empties/aliases `BootstrapLauncherCatalogRole`, supplies the Linux launcher or includes the launcher in role receipts；all reject before channel/provider/root access. The positive token fixes `BootstrapLauncherCatalogRole=c12_runner_launcher_windows_amd64`. Add `TestCompletionPublishBindingSurvivesExecutionAndInnerCleanupCAS`, `TestCompletionPublishBindingRejectsOuterBusinessDrift`, `TestCompletionLauncherDeathBetweenEveryAdoptAndACKSeam` and `TestCompletionCrossBootRecoveryNeverOpensAgain`. The first requires execution-reservation-only and every inner-union transition among `none|protected_key_intent|inner_bootstrap_restart_pending|cleanup_capsule|finalization_pending` to advance guard epoch/full `StateDigest`/`PreviousStateDigest` while preserving `PublishBindingDigest` byte-identically under both create `publish_pending` and revalidate `retained_complete_active`；manifest/materialization/completed/adopt/ACK/finalize-ready mutations must derive and cross-persist a new reciprocal digest. It separately injects response loss at plan/core/intent candidate/fsync/CAS、key CreateOrRecover/abort/absence、restart marker、replacement intent/capsule、selected WAL bootstrap/actual and finalization seams and rejects any owner/payload/head drift. Kill/reboot the launcher after Begin and between every pre-E2 adoption/publication/inner/manifest/ACK/finalize seam；the successor first proves the old containment absent and uses only the exact opaque publish-handle recovery path. Kill/reboot after E2 instead permits only clearing `exiting` and installing a distinct terminal-audit reservation；kill/reboot during `retained_complete_active` permits only `RecoverCompletionRetainedRevalidation`. Any unrepresented outer drift、second Open、dual live helper、post-E2 publish recovery or old PID kill rejects.

Also add `TestCompletionRetainedSemanticValidatorFreshAndRecoveredRoute`, `TestCompletionLatestTrustedTimeHeadRejectsRollbackAfterCompletedAndAborted`, `TestCompletionRetainedRevalidationOppositeOutcomeReturnsConflict`, `TestCompletionRetainedRevalidationPendingUsesNoSessionInnerFinalization`, `TestCompletionRetainedRevalidationPostTerminalResponseLossAudit`, `TestCompletionRevalidationOutcomeDeliveryMustAckBeforeNextOpen`, `TestCompletionAbortAfterValidTimeMustAdvanceHead`, `TestCompletionValidationChildResponseAbsentAndPresentHeadMatrix`, `TestCompletionRetainedCreationTimeBaselineRequiresSignedEnvelope`, `TestCompletionRetainedRevalidationDispatchIsClosedAndGuarded` and `TestCompletionRetainedRevalidationActiveCrashRoutesOnlyToRecovery`. Do **not** redeclare or edit the P08-owned unique roots `TestCompletionInnerFinalizersRequireExactLeafReceipt`、`TestCompletionInnerFinalizationReceiptCrashSeams` or `TestCompletionRetainedOutcomeConsumesSameCleanupReceipt`；P09 consumes and reruns those owning definitions unchanged, while every P09 composite scenario remains in the distinct Task 7 roots named above. The exact manifest still requires one global definition each. Together the P08 roots and distinct P09 roots use real signed envelopes/heads/outcomes、the exact leaf receipt、all revalidation and five inner dispatch values、both session classes、replacement reservation、delivery fsync/ACK seams. The raw selected-WAL crash root and high-level matrix cover Begin through WAL readiness. The P08 finalization crash root closes its leaf/runnerprofile seams, while the distinct P09 recovery roots run create、completed、pre-time abort and with-head abort across composite pending、response loss、key absence、tombstone unlink/parent-fsync/absence and final target CAS；only no-session finalization runs until CAS, then complete is truly actionless. Every target rejects the other four values；no absent/skip/old-arity、rollback、fallback、cleanup-only validation、complete recovery、audit Open or active→fresh/audit transition may pass.

Define `TestCompletionConsolidationRejectsOnlyWindowsTransaction` now in Task 7 inside `internal/c12completionpublication/operator_test.go`. It publishes only the isolated Windows member, calls the high-level held-set adoption path and requires finite `scope_set_incomplete` before trusted-time/provider/materialization with no enumeration、manifest、ACK or retained transition. Task 8 only reruns this already committed root and may not create or edit test source.

Freeze one literal completion semantic manifest with exactly 62 substantive roots. The manifest entries are triples `{Package,OwnerFile,Test}` and are hard-coded in `TestCompletionSemanticGateExactCoversDeclaredTests` in `internal/c12completionpublication/operator_test.go`; they are not discovered from source or a prefix. The six entries owned by `./internal/c12evidence`, file `internal/c12evidence/completion_test.go`, are:

`TestCompletionRejectsCrossBuildSplice|TestCompletionRejectsEveryAuthorityOperationsArtifactSplice|TestCompletionRejectsFreshV7StagingAsLegacyUpgradeComplete|TestCompletionRejectsParallelCleanOperationsLicenseSet|TestCompletionRejectsScopeRunIDReuseAcrossDistinctScopes|TestCompletionRequiresEachScopeExactlyOnce`.

The three unchanged P08 entries owned by `./internal/c12runnerprofile`, file `internal/c12runnerprofile/completion_state_test.go`, are:

`TestCompletionInnerFinalizationReceiptCrashSeams|TestCompletionInnerFinalizersRequireExactLeafReceipt|TestCompletionRetainedOutcomeConsumesSameCleanupReceipt`.

The remaining 53 entries are owned by `./internal/c12completionpublication`, file `internal/c12completionpublication/operator_test.go`:

`TestCompletionAbortAfterValidTimeMustAdvanceHead|TestCompletionAdoptionSetCrashRecovery|TestCompletionConsolidationRejectsOnlyWindowsTransaction|TestCompletionCreateDispatchIsClosedAndNoFallback|TestCompletionCrossBootRecoveryNeverOpensAgain|TestCompletionFinalizeReadyBindsImmutableRetainedPayloadDigest|TestCompletionFinalizeReadyHasNoDigestCycleOrPrematureCell|TestCompletionInnerCleanupAdapterPairsBothCapabilities|TestCompletionInnerCleanupDoesNotTerminalizeLaunchGeneration|TestCompletionInnerCleanupExcludesPublishState|TestCompletionLatestTrustedTimeHeadRejectsRollbackAfterCompletedAndAborted|TestCompletionLauncherDeathBetweenEveryAdoptAndACKSeam|TestCompletionLowLevelChannelCallsOwnedOnlyByCompositePackage|TestCompletionMaterializedViewExactEightMembers|TestCompletionModePhaseCrossClaimRejects|TestCompletionOpaqueHandleHidesReceiverSetBindingSinkAndView|TestCompletionPublicationOperatorHasNoCallerPathKindOrDigest|TestCompletionPublicationOperatorIsSoleStageCanonicalManifestCaller|TestCompletionPublicationProgressCrossPersistsEachOuterSeam|TestCompletionPublicationSinkManifestProjectionMatchesFrozenWireShape|TestCompletionPublishBindingRejectsOuterBusinessDrift|TestCompletionPublishBindingSurvivesExecutionAndInnerCleanupCAS|TestCompletionPublishPendingFreshProcessRecoveryDoesNotOpenAgain|TestCompletionPublishPendingPrecedesAdoptionSet|TestCompletionPublishPendingUsesSingleRoleRecoverySlot|TestCompletionRawRunnerprofileCallsOwnedOnlyByCompositePackage|TestCompletionRecoverFinalizationBeforeAndAfterGuardCAS|TestCompletionRecoverPublicationProgressClosesEveryCrashSeam|TestCompletionRetainedCompleteAdvancesLaunchGenerationOnce|TestCompletionRetainedCompleteIsImmutable|TestCompletionRetainedCreationTimeBaselineRequiresSignedEnvelope|TestCompletionRetainedRevalidationActiveCrashRoutesOnlyToRecovery|TestCompletionRetainedRevalidationDispatchIsClosedAndGuarded|TestCompletionRetainedRevalidationFailureClosesWithAuthenticatedAbort|TestCompletionRetainedRevalidationOppositeOutcomeReturnsConflict|TestCompletionRetainedRevalidationOutcomeRaceIsSingleWinner|TestCompletionRetainedRevalidationPendingUsesNoSessionInnerFinalization|TestCompletionRetainedRevalidationPostTerminalResponseLossAudit|TestCompletionRetainedSemanticValidatorFreshAndRecoveredRoute|TestCompletionRetainedTransitionHasNoSplitBrainWindow|TestCompletionRevalidationOutcomeDeliveryMustAckBeforeNextOpen|TestCompletionRoleStateGuardCrashSeams|TestCompletionRoleStateRejectsAuthenticatedRollback|TestCompletionTerminalAuditReceiptCannotDriveRevalidation|TestCompletionTerminalAuditSurvivesAbortedRevalidations|TestCompletionTerminalAuditSurvivesCompletedRevalidations|TestCompletionTerminalFinalizationAuditIsPathlessReadOnly|TestCompletionValidationChildResponseAbsentAndPresentHeadMatrix|TestCompletionValidationChildResponseCannotMintSemanticToken|TestCompletionValidationChildResponseMatchesConsumedRuntimeResult|TestCompletionValidationChildResponseStrictWire|TestCompletionValidationChildRuntimeIsOnlyCapabilityRoute|TestCompletionValidationRejectsCallerPolicyClientAndRawChildBytes`.

`TestCompletionSemanticGateExactCoversDeclaredTests` is the separate 63rd execution root. It parses all module-local Go test files independent of the planning OS, requires every substantive name to have exactly one top-level definition in its literal owner package/file, and exact-covers every top-level `TestCompletion*` definition in `internal/c12evidence/completion_test.go` against its six entries. For `internal/c12completionpublication/operator_test.go`, the exact set is its 53 substantive entries、this semantic meta root and the literal nonsemantic exclusion set `{TestCompletionLauncherImportClosureRemainsB10Frozen,TestCompletionStaticGateExactCoversDeclaredTests}`；those two exclusions are mandatory, may not grow, and are separately exact-listed/JSON-run with reachable-helper no-skip by Step 8's four-root static gate. The P08 file is checked only for the three named unique definitions because it also owns earlier frozen roots. The semantic meta resolves each of its 63 execution roots' statically reachable package-local functions、methods、function values and nested subtest literals with `go/types`; any reachable `Skip|Skipf|SkipNow`, unresolved dynamic local call target, renamed/missing/duplicate root, owner move or nonliteral manifest construction fails closed. It also exact-compares its triple table with the literal table in `scripts/invoke-exact-completion-semantic-gate.ps1` so neither runner nor meta-test can silently omit a member.

The PowerShell script groups that same literal table by package, builds only anchored literal alternations, requires `go test <package> -list <pattern>` to equal the sorted expected roots exactly, and then runs `go test <package> -json -run <pattern> -count=1`. For every one of the 63 roots it requires exactly one root `run` and one root `pass`, and zero `skip` events for the root or any descendant；a package error、extra listed root、missing root、duplicate event or descendant skip fails nonzero. It accepts no caller package、pattern、root or skip override. All Task 7/8 and final completion semantic commands invoke this script rather than a broad prefix.

Step 1 also defines both mandatory nonsemantic exclusion roots in `internal/c12completionpublication/operator_test.go`, so every Task 7 semantic-meta execution can prove that the exact exclusion set exists even though the 63-root semantic runner does not execute it. `TestCompletionLauncherImportClosureRemainsB10Frozen` strict-loads the Plan 08-owned read-only `testdata/c12/launcher-import-closure.v1.json` and requires exactly the ordered `{GOOS:windows,GOARCH:amd64,GOAMD64:v1,CGO:false,BuildTags:[],GOFLAGS:"",GOWORK:off,GOENV:off,GOEXPERIMENT:"",GOTOOLCHAIN:local}` and identical Linux record. For each record it invokes the locked Go tool from an allowlist-sanitized environment with every listed value explicit, passes no `-tags`、sorts the complete module-local `go list -deps` set and requires exact package-list plus `ClosureDigest` equality. Ambient GOOS/GOARCH/GOAMD64/CGO/GOFLAGS/GOWORK/GOENV/GOEXPERIMENT/GOTOOLCHAIN/build tags or any other `GO*` build selector、an absent/reordered/defaulted target or deriving an expected list from the current host fails. Both closures explicitly reject `internal/c12evidence` and `internal/c12completionpublication`. The launcher reaches only the P08 runnerprofile bridge, which spawns the selected versioned helper process；that helper—not the launcher—links the composite, and the composite invokes the separately frozen P08 private-child runtime. Any other module-local dependency drift fails even if those two packages remain absent.

The second Step 1 exclusion root is `TestCompletionStaticGateExactCoversDeclaredTests`. It parses every package `_test.go` with `go/parser`, requires exactly one top-level definition of itself、`TestCompletionLowLevelChannelCallsOwnedOnlyByCompositePackage`、`TestCompletionRawRunnerprofileCallsOwnedOnlyByCompositePackage` and `TestCompletionLauncherImportClosureRemainsB10Frozen`, and rejects duplicate definitions. It then loads the package with `go/types`, resolves the package-local call graph (including methods、function values whose concrete target is statically resolvable and nested function literals) from each of those four roots, and rejects `Skip`/`Skipf`/`SkipNow` in every root or transitively reachable helper；an unresolved dynamic test-helper call fails closed rather than escaping the scan. It also verifies that exactly the three substantive tests are present in the frozen static-gate member list and exact-compares the raw completion symbol/method/owner tables to the P08 declaration set；the meta-test is not allowed to weaken or replace any check. Task 7 need only compile these two exclusions and prove their mandatory definitions through the semantic meta；Step 8 is their first exact list/JSON execution and must pass both.

- [ ] **Step 2: Run completion tests and verify RED**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-completion-semantic-gate.ps1`

Expected: FAIL because completion types and validator are undefined.

- [ ] **Step 3: GREEN — implement fixed-order completion references**

```go
type ScopeReferenceV1 struct {
	Scope         Scope
	EvidenceDigest contracts.Digest
}

type CompletionPolicy struct {
	ScopePolicies                    [4]TrustPolicy // PowerShell, Git Bash, platform, authority
	PlatformPolicy                   PlatformTrustPolicy
	AuthorityPolicy                  AuthorityTrustPolicy
	ExpectedTrustedTimeSourceIdentity contracts.Digest
	ExpectedLicenseRecordSetDigest     contracts.Digest
	ExpectedDocumentationDigest        contracts.Digest
	NotAfter                           time.Time
}

type CompletionArtifactProjection struct {
	ScanReceiptDigest          contracts.Digest
	SupplyChainBundleDigest    contracts.Digest
}

type CompletionLicenseProjection struct {
	LicenseRecordSetDigest contracts.Digest
	DocumentationDigest    contracts.Digest
}

type CompletionTrustedTimeInput struct {
	Evidence             C12TrustedTimeEvidenceV1
	CanonicalInputDigest contracts.Digest
	ChallengeNonce       [32]byte
}

type CompletionTrackedTreeInput struct {
	Reader        c12runnerprofile.CompletionValidationChildTrackedTreeReader
	BindingDigest contracts.Digest
}

// Task 7 owns this sole child/parent wire DTO and codec after the retained
// manifest schema exists; Task 4 owns only its trusted-time primitives.
type CompletionValidationChildResponseV1 struct {
	SchemaVersion                     string
	Purpose                           c12runnerprofile.CompletionValidationPurpose
	CanonicalInputDigest              contracts.Digest
	MaterializedEvidenceBindingDigest contracts.Digest
	ChallengeNonce                    [32]byte
	TrustedTimeEvidence               C12TrustedTimeEvidenceV1
	TrustedTimeEvidenceDigest         contracts.Digest
	IssuedAt                          time.Time
	RunnerProfileDigest               contracts.Digest
	ProviderProfileDigest             contracts.Digest
	InstalledChildProjectionDigest    contracts.Digest
	ChildProcessBindingDigest         contracts.Digest
	ProviderSessionBindingDigest      contracts.Digest
	TrackedTreeBindingDigest          contracts.Digest
	ValidatedTrackedFactsDigest       contracts.Digest
	PreviousTrustedTimeHeadDigest     contracts.Digest
	PreviousTrustedTimeHeadGeneration uint64
	RuntimeBindingDigest              contracts.Digest
}

// Semantic in-process input only; never a wire/path/provider/client surface.
type CompletionInputs struct {
	EvidenceSetBindingDigest contracts.Digest
	EvidenceMembers          [8]CompletionEvidenceReferenceV1
	Scopes                   [4]C12ScopeEvidenceV1
	Platform                 PlatformEvidenceV1
	Authority                AuthorityFenceEvidenceV1
	Policy                   CompletionPolicy
	Artifacts                CompletionArtifactProjection
	Licenses                 CompletionLicenseProjection
	TrustedTime              CompletionTrustedTimeInput
	TrackedTree              CompletionTrackedTreeInput
}

// Retained validation is a separate nonpublication semantic path. It cannot
// produce canonical manifest bytes or be passed to the P08 publication sink.
type RetainedCompletionInputs struct {
	Evidence                  c12runnerprofile.CompletionValidationChildEvidenceView
	RetainedManifestCanonical []byte
	Policy                    CompletionPolicy
	FreshTrustedTime          CompletionTrustedTimeInput
	ValidatedCreationBaseline ValidatedCompletionCreationTimeBaseline
	ValidatedFreshTimeHead    ValidatedCompletionTrustedTimeHead
	PreviousTrustedTimeHead   c12runnerprofile.CompletionRetainedTrustedTimeHeadView
	TrackedTree               CompletionTrackedTreeInput
	ChildResponse             CompletionValidationChildResponseV1
}

// internal/c12evidence/completion.go: public semantic type whose canonical
// wire shape must exact-match P08's already-frozen private sink projection.
const MaxCanonicalCompletionManifestBytes = 64 << 10

type CompletionEvidenceReferenceV1 struct {
	Role            c12runnerprofile.CompletionEvidenceMemberRole
	CanonicalDigest contracts.Digest
}

type C12CompletionManifestV1 struct {
	SchemaVersion          string
	EvidenceSetBindingDigest contracts.Digest
	EvidenceMembers        [8]CompletionEvidenceReferenceV1
	CreatedAt              time.Time
	ExpiresAt              time.Time
	TrustedTimeSourceIdentity contracts.Digest
	CreationTrustedTimeInputDigest contracts.Digest
	CreationTrustedTimeChallengeNonce [32]byte
	CreationTrustedTimeEvidence C12TrustedTimeEvidenceV1
	CreationTrustedTimeEvidenceDigest contracts.Digest
	CreationTrustedTimeSequence uint64
	Scopes                 [4]ScopeReferenceV1
	RepoCommit             string
	TrackedTreeDigest      contracts.Digest
	SpecDigest             contracts.Digest
	ToolchainDigest        contracts.Digest
	ReleaseDigests         [3]contracts.Digest
	AuthorityOperationsArtifact AuthorityOperationsArtifactTupleV1
	AuthorityOperationsBinaryAndImageDigest contracts.Digest
	AuthorityOperationsSupplyChainRecordSetDigest contracts.Digest
	SupplyChainEvidenceBundleDigest contracts.Digest
	ArtifactCatalogDigest  contracts.Digest
	ExpectedProcessImagePolicyDigest contracts.Digest
	ImageSetDigest         contracts.Digest
	ArtifactScanDigest     contracts.Digest
	LicenseRecordSetDigest contracts.Digest
	DocumentationDigest    contracts.Digest
}

type ValidatedCompletionManifest interface {
	validatedCompletionManifest() // opaque canonical manifest + exact view/set/policy semantic-validation binding
}

type ValidatedRetainedCompletion interface {
	validatedRetainedCompletion() // opaque nonpublication success bound to one active revalidation/result/head candidate
}

type ValidatedCompletionTrustedTimeHead interface {
	validatedCompletionTrustedTimeHead() // opaque same-source/latest-head successor proof; no overall semantic success
}

type ValidatedCompletionCreationTimeBaseline interface {
	validatedCompletionCreationTimeBaseline() // sealed retained creation envelope/source/challenge/sequence/floor baseline
}

func ValidateCompletionInputs(CompletionInputs) (ValidatedCompletionManifest, error)
func CanonicalValidatedCompletionManifest(ValidatedCompletionManifest) ([]byte, C12CompletionManifestV1, error) // nonempty, <= MaxCanonicalCompletionManifestBytes
func RunInheritedCompletionValidationChild(context.Context, InheritedCompletionTrustedTimeRuntime) (CompletionValidationChildResponseV1, error)
func CanonicalCompletionValidationChildResponse(CompletionValidationChildResponseV1) ([]byte, contracts.Digest, error)
func StrictDecodeCompletionValidationChildResponse([]byte) (CompletionValidationChildResponseV1, contracts.Digest, error)
func ValidateRetainedCompletionCreationTimeBaseline([]byte, contracts.Digest, CompletionPolicy) (ValidatedCompletionCreationTimeBaseline, error)
func ValidateCompletionTrustedTimeHead(CompletionValidationChildResponseV1, c12runnerprofile.CompletionRetainedTrustedTimeHeadView, ValidatedCompletionCreationTimeBaseline) (ValidatedCompletionTrustedTimeHead, error)
func ValidateRetainedCompletionInputs(RetainedCompletionInputs) (ValidatedRetainedCompletion, error)
```

In `internal/c12completionpublication/operator.go`, define the sole high-level composite:

```go
type CompletionPublicationOperator interface {
	completionPublicationOperator() // opaque; privately owns one P08 sink, exact view and validated token
	MaterializeAdoptedEvidence(context.Context) error
	ValidateMaterializedEvidence(context.Context, CompletionInnerCleanupSession) error
	StageValidatedManifest(context.Context) error
	PublishManifest(context.Context) error
	PublishCompletedMarker(context.Context) error
	RecoverPublication(context.Context) error
}

type CompletionInnerCleanupOwner interface {
	completionInnerCleanupOwner() // implemented only by package-owned create/revalidation handles
}

type CompletionInnerCleanupRecoveryDispatch string

const (
	CompletionInnerCleanupRecoveryBeginRequired     CompletionInnerCleanupRecoveryDispatch = "begin_required"
	CompletionInnerCleanupRecoverySessionValidation CompletionInnerCleanupRecoveryDispatch = "session_validation_recovery"
	CompletionInnerCleanupRecoverySessionCleanup    CompletionInnerCleanupRecoveryDispatch = "session_cleanup_recovery"
	CompletionInnerCleanupRecoveryFinalization      CompletionInnerCleanupRecoveryDispatch = "finalization_recovery"
	CompletionInnerCleanupRecoveryComplete          CompletionInnerCleanupRecoveryDispatch = "complete"
)

type CompletionInnerCleanupSession interface {
	completionInnerCleanupSession() // opaque handle+leaf capability/proof/receipt plus sealed validation_required|cleanup_only resume class
}

type CompletionPublishPendingHandle interface {
	CompletionInnerCleanupOwner
	completionPublishPendingHandle() // opaque composite-package holder; no low-level accessor
	AdoptExpectedEvidence(context.Context) error // idempotently Begin/Recover one set and adopt exact three fixed bindings
	AcknowledgeExpectedEvidence(context.Context) error // idempotently reacquire/recover and ACK exact three after completed_durable
	FinalizeRetainedComplete(context.Context) (CompletionRetainedCompleteView, error) // private Prepare/Recover/Commit over exact held set
}

type CompletionRetainedCompleteView interface {
	completionRetainedCompleteView() // opaque, immutable read-only retained state
}

type CompletionTerminalAuditReceipt interface {
	completionTerminalAuditReceipt() // opaque read-only audit result; never a cleanup/revalidation owner
}

type CompletionRetainedRevalidationOutcomeAuditReceipt interface {
	completionRetainedRevalidationOutcomeAuditReceipt() // exact P08 pending-delivery receipt; only durable-delivery+ACK authority
}

type CompletionRetainedRevalidationOutcomeDeliveryReceipt interface {
	completionRetainedRevalidationOutcomeDeliveryReceipt() // protected launcher result-slot durable-accept proof
}

type CompletionCreateRecoveryDispatch string

const (
	CompletionCreateFreshOpen       CompletionCreateRecoveryDispatch = "fresh_open"
	CompletionCreatePublishRecovery CompletionCreateRecoveryDispatch = "publish_recovery"
	CompletionCreateTerminalAudit   CompletionCreateRecoveryDispatch = "terminal_audit"
)

type CompletionRetainedRevalidationDispatch string

const (
	CompletionRetainedRevalidationFreshOpen      CompletionRetainedRevalidationDispatch = "fresh_open"
	CompletionRetainedRevalidationActiveRecovery CompletionRetainedRevalidationDispatch = "active_recovery"
	CompletionRetainedRevalidationOutcomeAudit   CompletionRetainedRevalidationDispatch = "outcome_audit"
)

type CompletionRetainedRevalidationHandle interface {
	CompletionInnerCleanupOwner
	completionRetainedRevalidationHandle() // opaque active retained generation; no channel/publication authority
	ValidateRetainedCompletion(context.Context, CompletionInnerCleanupSession) error
	FinalizeRetainedCompletion(context.Context, CompletionInnerCleanupSession) (CompletionRetainedCompleteView, error)
	AbortRetainedRevalidation(context.Context, CompletionInnerCleanupSession) (CompletionRetainedCompleteView, error)
}

func BindCompletionPublicationOperator(CompletionPublishPendingHandle) (CompletionPublicationOperator, error)
func InspectCompletionCreateRecoveryDispatch(context.Context) (CompletionCreateRecoveryDispatch, error)
func BeginCompletionPublishPending(context.Context, *c12runnerprofile.AuthenticatedNativeRunnerSession) (CompletionPublishPendingHandle, error)
func RecoverCompletionPublishPending(context.Context) (CompletionPublishPendingHandle, error)
func BeginCompletionInnerCleanup(context.Context, CompletionInnerCleanupOwner) (CompletionInnerCleanupSession, error)
func InspectCompletionInnerCleanupRecoveryDispatch(context.Context, CompletionInnerCleanupOwner) (CompletionInnerCleanupRecoveryDispatch, error)
func RecoverCompletionInnerCleanup(context.Context, CompletionInnerCleanupOwner) (CompletionInnerCleanupSession, error)
func RecoverCompletionInnerFinalization(context.Context, CompletionInnerCleanupOwner) error
func FinalizeCompletionInnerCleanup(context.Context, CompletionPublishPendingHandle, CompletionInnerCleanupSession) error
func InspectCompletionRetainedFinalization(context.Context) (CompletionTerminalAuditReceipt, error)
func InspectCompletionRetainedRevalidationDispatch(context.Context) (CompletionRetainedRevalidationDispatch, error)
func InspectCompletionRetainedRevalidationOutcome(context.Context) (CompletionRetainedRevalidationOutcomeAuditReceipt, c12runnerprofile.CompletionRetainedRevalidationOutcome, CompletionRetainedCompleteView, error)
func PersistCompletionRetainedRevalidationOutcomeDelivery(context.Context, CompletionRetainedRevalidationOutcomeAuditReceipt) (CompletionRetainedRevalidationOutcomeDeliveryReceipt, error)
func AcknowledgeCompletionRetainedRevalidationOutcome(context.Context, CompletionRetainedRevalidationOutcomeAuditReceipt, CompletionRetainedRevalidationOutcomeDeliveryReceipt) error
func OpenCompletionRetainedRevalidation(context.Context, *c12runnerprofile.AuthenticatedNativeRunnerSession) (CompletionRetainedRevalidationHandle, error)
func RecoverCompletionRetainedRevalidation(context.Context) (CompletionRetainedRevalidationHandle, error)
```

`CompletionInputs` is an exact in-process semantic value, not serialized input. `CompletionPolicy.ScopePolicies` is exactly four entries in PowerShell、Git Bash、platform、authority order；`PlatformPolicy` and `AuthorityPolicy` are the complete package-derived nested policies and must respectively contain byte-identical `ScopePolicy` values to `ScopePolicies[2]` and `ScopePolicies[3]`；each entry fixes its own scope/provider/run/runner/daemon/test/cleanup values, while all four `RunID` values must be nonzero and pairwise distinct even when every other shared build fact is byte-identical. All six pair substitutions reject as `ErrScopeRunIDReuse` before nested validation or publication. The four scopes must byte-equal on repo/tree/spec/toolchain/three releases and every scanner-derived tuple/record/bundle/catalog/process-policy/image/full-scan field. One anonymous single-scope `TrustPolicy` cannot validate the array, and neither nested signed payload can self-supply its omitted platform/authority policy fields. Creation and every revalidation must call the sole complete `ValidatePlatformEvidence(Platform,Policy.PlatformPolicy)` and `ValidateAuthorityFenceEvidence(Authority,Policy.AuthorityPolicy)` entrypoints before deriving their signed envelope digests. Its two artifact digests must equal the exact `scan_receipt` and `supply_chain_bundle` member digests in the same ordered view；the license/document pair must equal both the fixed policy and their receipt-bound projections；the trusted-time triple must recompute the signed purpose/challenge and supplies no caller scalar time. `TrackedTree` must be the sealed P08 reader from the same consumed result and binding digest; the validator recomputes the complete canonical tracked-tree inventory and reads only the closed fixed roles to validate the three-member spec set、catalog、toolchain、obligation documents、threat model/three runbooks and provider/runner profiles. Every field is mandatory/nonzero except values whose nested strict schema explicitly permits otherwise. Production constructs the whole value only inside the composite from its consumed P08 evidence/tree view and authenticated child response；tests in `internal/c12evidence/completion_test.go` may build package-local fixtures, while facade/ownership tests live in `internal/c12completionpublication/operator_test.go` to avoid a test import cycle.

`RetainedCompletionInputs` is the distinct revalidation path. It requires the P08 result's defensive exact-eight bytes、same binding/tree reader、nonzero retained manifest bytes/digest、decoded child response and fresh trusted-time input to be byte/digest-equal. Before judging the fresh response, the composite calls `ValidateRetainedCompletionCreationTimeBaseline(retainedCanonical,retainedDigest,derivedPolicy)`. That function strict-decodes the manifest, revalidates the complete signed `completion_create` envelope and challenge preimage, exact-matches `CreatedAt`、source、sequence、floor、input digest/nonce and returns a sealed baseline；a lone manifest or digest cannot mint it. The composite then strict-matches the DTO/P08 result, calls Task 4's signed fresh-evidence validator and `ValidateCompletionTrustedTimeHead(response,previousHead,baseline)`, and immediately retains that head token together with the same P08 `CompletionValidatedTrustedTimeHeadCandidate` before any remaining scope/license/tree semantics. When previous head is absent, the fresh response must be a strict successor of the sealed creation baseline；when present, its canonical bytes/digest/generation must exact-match P08 state and it must itself validate as a same-source successor of the baseline before the new response can be newer. Sequence and IssuedAt strictly increase and Floor never decreases. A later overall semantic failure therefore still uses the with-head abort route, while a failure before this barrier uses preserve-head abort. `ValidateRetainedCompletionInputs` then revalidates every retained scope/nested signature、tracked fact and current policy using both sealed time tokens and returns only `ValidatedRetainedCompletion`; it has no canonicalizer/staging route. Tests exercise real signed creation envelopes for first revalidation and for the successor after both completed and aborted edges, rejecting any digest-only baseline or rollback behind the latest head.

`CompletionValidationChildResponseV1` is the sole child/parent wire DTO with exact schema `talenro-c12-completion-validation-child-response/v1` and a hard 64 KiB canonical limit. The private child obtains every binding field from `OpenInheritedCompletionTrustedTimeRuntime`, performs its sole fresh fetch, requires `IssuedAt` and the complete signed time-envelope digest to match that result, canonical-encodes this DTO and writes only those bytes. Strict decode rejects unknown/duplicate/null/defaulted fields、noncanonical UTC、unknown purpose、oversize/noncanonical encoding or redundant-field mismatch. Every ordinary binding is nonzero. The previous-head pair is conditionally exact：`completion_create` and the first `completion_revalidate` require `(PreviousTrustedTimeHeadDigest=zero,PreviousTrustedTimeHeadGeneration=0)` and byte-match an absent P08 view；a later revalidation requires both nonzero and exact-match the P08 head；one-zero/one-nonzero、a nonzero create head or a zero pair when P08 has a head rejects. Parent-side composite first consumes P08's sealed runtime result, requires its purpose/input/view/nonce/profile/child/process/provider/tree/runtime/head fields and response digest to constant-time equal this decoded DTO, then calls Task 4's existing signed-evidence validator. It performs no alternate signature/time semantics. Tests `TestCompletionValidationChildResponseStrictWire`, `TestCompletionValidationChildResponseAbsentAndPresentHeadMatrix`, `TestCompletionValidationChildResponseMatchesConsumedRuntimeResult` and `TestCompletionValidationChildResponseCannotMintSemanticToken` own the exact wire boundary and composite binding.

Every fresh or recovered owner first invokes `InspectCompletionInnerCleanupRecoveryDispatch(ctx,owner)`, which maps authenticated owner/tag/outer progress plus same-set pending transition exactly to `begin_required|session_validation_recovery|session_cleanup_recovery|finalization_recovery|complete` and mints no authority. Begin alone drives plan/core/intent/key/Seal/capsule selection, stores the sealed leaf WAL binding, and internally exact-opens/reopens the selected WAL；it returns a validation-required session only after WAL bootstrap+observed-identity actual are file+parent-directory durable. Validation recovery converges intent/restart/replacement-capsule seams and performs the same WAL recovery barrier before returning. Create-only cleanup recovery first converges any set-first manifest-intent state barrier without filesystem/provider/child/validator/canonicalizer work, then revalidates the existing capsule/WAL and returns only cleanup-only. Neither returns `(nil,nil)` or exposes the leaf binding encoding/recorder. The composite calls the sole descriptor consumer, which derives policy from the authenticated capsule, supplies only the fixed completion role, exact-matches its embedded nonzero launch generation and machine-seal locator to the descriptor/owner/coordinator and recomputes the complete policy binding；only then does it call `OpenCompletionInnerCleanupCapability(ctx,rawHandle)`. That runnerprofile method rereads current owner/tag/epoch/full digest/capsule/policy, linearly consumes the cleanup-open claim and privately returns the assignment-identical delete-only capability. The composite pairs that capability with the same WAL-ready raw handle；validation-required alone may bind the P08 runtime, whose fixed resource creators append/fsync intent and observed actual through the handle-private recorder before provider/process use, while cleanup-only rejects runtime/validation/Stage. Finalization executes cleanup, prepares the leaf receipt and first selects `finalization_pending` with receipt+tombstone+fixed target. The live call or selector-free `RecoverCompletionInnerFinalization` then closes leaf Commit、key absence、tombstone unlink+parent-directory-fsync/exact absence and finally the target CAS；it returns no session/handle. Thus a crash anywhere before target selection stays in `finalization_recovery`, while create target-completed `none` plus manifest-intent-or-later is the truly actionless `complete` route. Genuine pre-manifest `none` remains Begin. Every target repeats all five checks；wrong resume class、role/generation/machine-seal/policy splice、pre-CAS/stale descriptor、missing WAL barrier、receipt/key/tombstone splice、target CAS before retirement、error fallback or retained/create cross-use rejects.

`internal/c12evidence/completion.go` owns only the manifest semantic schema、strict validator and opaque `ValidatedCompletionManifest` token；it does not own filesystem progress or a publish handle. `internal/c12completionpublication/operator.go` owns the concrete `CompletionPublishPendingHandle` and high-level operator containers, but implements none of `EvidencePublishPendingBinding`、`EvidenceChannelReceiver`、`CompletionPublicationSink` or `CompletionMaterializedEvidenceView`. No test、runner or alternate package may construct any of those Plan 08 types. The operator package is the only legal join point importing both `c12evidence` and `c12runnerprofile`; neither dependency imports it. The Plan 08 coordinator stores top-level schema `talenro-c12-completion-role-state/v1` as one logical record in the completion role's fixed A/B recovery cells, with closed variants `idle|active_unstarted|publish_pending|retained_complete|retained_complete_active`. Every variant contains nonzero `StateEpoch` and role-unique `CounterIdentityDigest` plus `PreviousStateDigest` and its launch/phase payload；`PreviousStateDigest` is zero only for first-install epoch 1 and otherwise equals the immediately expected selected state digest at a guarded mutation. Current-state loading validates the guard-selected cell directly and never depends on the inactive/history cell. The guard stores only epoch/current digest and is not another completion record. Both create `publish_pending` and revalidate `retained_complete_active` contain exactly one nested inner-cleanup union `none|protected_key_intent|inner_bootstrap_restart_pending|cleanup_capsule|finalization_pending`; no other phase can carry it, and there is no sibling outer/inner-state file or separately committed active bit. Every ordinary inner transition rewrites the complete opposite cell, sets `PreviousStateDigest` to the exact prior full digest, advances `StateEpoch`/full `StateDigest` and preserves the same owner-specific outer business plus byte-identical `PublishBindingDigest`; the sole cleanup-route exception invokes the fixed set-bound manifest-intent barrier before returning its session, not an arbitrary inner mutation. The `publish_pending` payload otherwise binds phase、run/role/launch-generation/session digest、the exact three ordered kind/slot/sequence expectations、stable strict-eight-file root identity、manifest temp/final identities、`completed` marker identity、three ACK reservation identities、immutable `AdoptionSetIdentityDigest`、latest `LastSetTransitionDigest`、per-binding semantic adoption/ACK intent+actuals and a closed recovery plan. It never stores a mutable set pending/actual/full-record digest in the outer publish projection. While publication progress is `manifest_intent|manifest_actual`, the guarded outer payload and reciprocal set pending/actual record each durably carry the same nonempty at-most-64-KiB canonical manifest bytes、their digest、exact view/set binding and reserved temp/final identities. The manifest T hashes those semantic values with identity、previous T and source P only；outer Pnew covers T+bytes+semantic fields but excludes the set-record digest/new-P copy, while the set actual one-way binds T+Pnew. The byte copies remain through reciprocal `manifest_actual` and may be erased only by reciprocal `manifest_durable` after rename plus parent-directory fsync；they may not exist under any other progress tag. This preserves a reconstruction source if power loss drops the temp directory entry or an un-fsynced rename without creating a digest fixed point；the retained-active owner instead preserves the immutable payload/audit/latest-head projection. Composite `BeginCompletionPublishPending` consumes/invalidates the supplied in-memory session with `OpenCompletionRoleStateForSession`, strict-validates its `active_unstarted(N)` business projection, calls only coordinator `BeginPublishPending`, then obtains the concrete binding/receiver only through `BindCompletionEvidenceReceiver` before returning their holder. `RecoverCompletionPublishPending(ctx)` accepts no session/path/options：it calls only `RecoverFixedCompletionRoleState` followed by coordinator `RecoverPublishPending`, authenticates the guard and unique selected cell, and pathlessly recovers either the exact pre-barrier `active_unstarted(N)` after an immediate process crash or the existing `publish_pending(N)` variant；it then binds the same receiver capability and never calls Open、claims another generation、selects a run or creates a replacement set. An earlier valid selected cell replay, stale epoch/digest or guard/current-cell ambiguity returns no handle and quarantines；an incomplete inactive candidate before guard advance has zero authority and cannot invalidate the selected state.

`InspectCompletionCreateRecoveryDispatch` is the sole non-authorizing create-mode router and is itself implemented only inside the composite package. It maps the authenticated P08 role state plus the fixed launcher reservation exactly to `fresh_open` for `idle`, `publish_recovery` for absence-proved same-generation `active_unstarted|publish_pending`, and `terminal_audit` for inactive `retained_complete` only after the old `exiting` reservation has been cleared and a distinct read-only audit reservation is selected. Unknown/corrupt/reservation-mismatched state、`retained_complete_active` or any revalidation phase returns no route. The fixed helper calls exactly one target. Fresh work performs the one permitted generic session Open and passes that session immediately to `BeginCompletionPublishPending`, which rechecks the same reservation and exact `idle→active_unstarted` lineage；publish recovery calls `RecoverCompletionPublishPending`, and terminal audit calls `InspectCompletionRetainedFinalization`, each of which rechecks its own dispatch phase before accessing raw P08 state. No error authorizes probing or falling through to another target.

`handle.FinalizeRetainedComplete(ctx)` is the only create-side mutable exit. Its concrete composite implementation privately uses the held receiver/set/binding to Prepare or recover the reciprocal E→E1 `finalize_ready` transition, freezing both the exact one-use `retained_complete(active=nil,next=N+1)` E2 target-transition digest and a distinct immutable retained-payload/finalization-set audit digest that excludes `next`、`active`、execution reservation and inner cleanup. It accepts no caller set、digest or bytes. Only the internally held E1/D1/P1 proof can Commit：Commit constructs/file+directory-fsyncs the exact opposite-cell E2 target, advances+rereads the guard and marks the matching execution reservation `exiting`; no terminal cell exists before that proof-consuming call. A pre-E2 crash recovers only through the same opaque publish handle. After E2, same-generation recovery is forbidden：the launcher first proves the old containment absent and clears its exiting reservation, then a later create-mode invocation installs a distinct read-only `terminal_audit` reservation and calls `InspectCompletionRetainedFinalization(ctx)`. That function returns only `CompletionTerminalAuditReceipt`, wraps P08's read-only receipt and authenticates the immutable audit digest plus the closed lineage from original E2 through zero or more closed `completed|aborted` revalidation edges to current `retained_complete(active=nil)`；it rejects `retained_complete_active` and exposes no cleanup/revalidation/channel/mutation authority. The original set remains a powerless audit record. `OpenCompletionRetainedRevalidation` converts only an actual revalidate-mode parent session that privately claims the next generation；pathless `RecoverCompletionRetainedRevalidation` rehydrates only exact `retained_complete_active` after absence proof and never Opens again. Both return `CompletionRetainedRevalidationHandle`, not an audit receipt. Its inner-cleanup session and `ValidateRetainedCompletion` are composite-owned. Validation retains the distinct `ValidatedRetainedCompletion` token together with the same sealed P08 trusted-time-head candidate. After exact cleanup, `FinalizeRetainedCompletion` alone may consume that pair and select `finalization_pending` with a fixed `completed` target；`AbortRetainedRevalidation` accepts no outcome/reason bytes and selects the same pending tag with only a fixed non-affirming `aborted` target, automatically carrying the candidate only if trusted-time validation had succeeded. The first guarded pending CAS wins. A live same-target replay exact-matches；the opposite target returns `ErrRetainedRevalidationOutcomeConflict` plus no view, never nil success. After any crash with `finalization_pending`, only the five-way inner router's selector-free, no-session `RecoverCompletionInnerFinalization` may replay the stored receipt/key/fixed target and finish the second CAS；there is no handle/session-bound retained-outcome recovery API. That second CAS consumes the active generation, preserves the immutable payload/audit digest, updates or preserves the latest trusted-time head by P08's fixed rule, installs `outcome_delivery_pending` and marks the reservation `exiting`. Response loss then returns only through `InspectCompletionRetainedRevalidationOutcome`、protected result delivery and receipt-bound ACK after containment absence/clear；the audit reports the stored outcome class without reopening an active handle. No retained handle implements a channel interface or carries publication、adoption or ACK authority.

The zero-argument revalidate mode first calls `InspectCompletionRetainedRevalidationDispatch`, which maps the authenticated P08 enum one-to-one as `fresh_open→fresh_open`、`active_recovery→active_recovery` and `outcome_audit→outcome_audit`. Exact `retained_complete_active` after old-containment absence proof and installation+reread of the same-generation replacement reservation routes only to selector-free `RecoverCompletionRetainedRevalidation`; inactive `retained_complete` with durable `outcome_delivery_pending` routes only to `InspectCompletionRetainedRevalidationOutcome`; inactive with pending absent routes only to `OpenCompletionRetainedRevalidation`. The fixed launcher—not any facade caller—installs the matching `revalidation_recovery`、`revalidation_outcome_audit` or fresh-work reservation before spawning the helper. Every target repeats the same guarded phase/reservation/generation check before accessing its private raw state and rejects the other two dispatch values；an inspect/target error is terminal and never permission to probe or fall through. The recovery facade calls only P08 `RecoverFixedCompletionRoleState` and coordinator `RecoverRetainedComplete`, rehydrates the original active generation and never Opens or claims a new generation. The audit facade only authenticates/consumes the selected reservation, wraps the sealed P08 receipt in `CompletionRetainedRevalidationOutcomeAuditReceipt`, exact-matches outcome/generation/payload/latest-head to the returned read-only complete view and performs no provider/child/cleanup call. Before ACK, the parent must call `PersistCompletionRetainedRevalidationOutcomeDelivery`；that zero-path facade asks P08 to write/fsync/reopen the bounded canonical outcome in the reservation-bound protected launcher result slot and returns an opaque delivery receipt only after file and parent-directory durability. `AcknowledgeCompletionRetainedRevalidationOutcome` requires the assignment-identical audit+delivery pair and delegates to P08's guarded pending-only ACK；same pair is idempotent and any missing/older/different receipt rejects. The generation-bound result record remains replay-readable after ACK/helper/launcher response loss and cannot be overwritten by later work. Until durable acceptance、ACK and launcher absence/clear all complete, fresh Open is forbidden. Active handle、inner session、post-terminal outcome audit and create terminal audit remain disjoint, so neither an active crash nor post-terminal response loss can be mistaken for a new generation or lose the previous result.

Plan 09 consumes Plan 08's sole transaction/channel/set/publication/finalization machinery without redefining it, but only `internal/c12completionpublication/operator.go` may name or call the low-level `EvidencePublishPendingBinding`、`EvidenceChannelReceiver`、`EvidenceAdoptionSet`、`CompletionPublicationSink`、`CompletionMaterializedEvidenceView` or their methods. Sender profiles still bind one symbolic sender policy and the completion profile fixes three ordered nonenumerable receiver bindings；diagnostic has none. The create command receives only `CompletionPublishPendingHandle` and `CompletionPublicationOperator`. `handle.AdoptExpectedEvidence(ctx)` privately begins or exact-recovers the one held set and adopts the three fixed bindings in profile order, with each P08 intent/actual and reciprocal digest durable before advancing；it accepts no kind、slot、sequence、nonce or set. `BindCompletionPublicationOperator(handle)` privately mints and retains the low-level sink/view only after those three actuals. The operator materializes the exact eight members, accepts only a matching `CompletionInnerCleanupSession`, runs and authenticates the protected child internally, constructs `CompletionInputs` from the retained view plus package-derived fixed policy/provider facts and holds the resulting `ValidatedCompletionManifest` token privately. `StageValidatedManifest(ctx)` alone obtains the nonempty at-most-64-KiB canonical bytes and calls the low-level sink；that call does not return until the same bytes、digest、view/set binding and reserved identities are durably reciprocal in both `manifest_intent` records before any temp write, so loss of the in-memory token is harmless. No command/helper receives a token、receiver、binding、set、lease、sink、view、child result or caller policy/client. After `completed_durable`, `handle.AcknowledgeExpectedEvidence(ctx)` privately reacquires or recovers the exact three ACKs without a selector；`handle.FinalizeRetainedComplete(ctx)` privately owns Prepare/Recover/Commit. Every pre-E2 crash reopens only the same opaque handle through `RecoverCompletionPublishPending`, then rebinds the high-level operator and calls `RecoverPublication` after any required composite inner-cleanup recovery. From `manifest_intent` onward that recovery uses only its authenticated durable byte source and never invokes the provider、child、validator or canonicalizer. AST/compile-fail tests reject every production reference to those low-level types or calls outside `internal/c12completionpublication/operator.go`, every alternate Stage caller and every command import that would expose them.

Add `TestCompletionRawRunnerprofileCallsOwnedOnlyByCompositePackage`. Across every module-local production `.go` file and both frozen Windows/Linux build projections, it combines import-alias/`SelectorExpr`/`MethodExpr` AST checks with `go/types` selection resolution. Its literal P08 parent-side completion symbol set is exact：`OpenCompletionRoleStateForSession`、`RecoverFixedCompletionRoleState`、`InspectFixedCompletionRoleState`、`InspectFixedCompletionRevalidationDispatch`、`InspectFixedCompletionRetainedFinalization`、`InspectLastFixedCompletionRevalidationOutcome`、`InspectCompletionRetainedRevalidationOutcomeReceipt`、`PersistLastFixedCompletionRevalidationOutcomeDelivery`、`AcknowledgeLastFixedCompletionRevalidationOutcome`、`BindCompletionValidationChildRuntime`、`RunCompletionValidationChild` and `ConsumeCompletionValidationChildResult`; its coordinator-method set is exactly `ReadCurrent|BeginPublishPending|RecoverPublishPending|BeginInnerCleanup|InspectInnerRecoveryDispatch|RecoverInnerCleanup|RecoverInnerFinalization|FinalizeInnerCleanup|OpenRetainedComplete|RecoverRetainedComplete|FinalizeRetainedRevalidation|AbortRetainedRevalidation|AbortRetainedRevalidationWithTrustedTimeHead`. Only `internal/c12completionpublication/operator.go` may name/call those symbols, types or resolved methods. The separately enumerated cleanup exception set is exactly `{DescribeCompletionInnerCleanup,OpenCompletionInnerCleanupCapability}` and both selectors are permitted only in `internal/c12evidence/cleanup.go`; that file passes only the same raw handle to the one-argument guarded Open and never a policy/adapter/set. The test exact-compares these literal sets to the complete exported P08 bridge declarations and coordinator method set, rejects dot imports、aliases/function values/forwarders and rejects any raw coordinator、inner handle、child runtime/result or terminal/outcome audit capability in another package's exported signature. It also freezes the high-level inner quartet `BeginCompletionInnerCleanup|InspectCompletionInnerCleanupRecoveryDispatch|RecoverCompletionInnerCleanup|RecoverCompletionInnerFinalization`, proves each of the four action targets rechecks its exact five-value enum、`complete` calls none of them, exact-checks both sealed session resume classes and rejects error-fallback or raw-state branching. `internal/c12evidence/completion.go`、all `cmd/**`、runner helpers and alternate packages must have zero reference. The existing child-side provenance ownership test remains the exact allowlist for inherited-session read/exchange APIs. Thus the selected helper can invoke only the high-level dispatch/handle/operator facade even though it runs in the same authenticated process.

`C12CompletionManifestV1.EvidenceMembers` is exactly eight ordered references matching the P08 roles：`windows_powershell_scope`、`windows_git_bash_scope`、`linux_platform_scope`、`authority_fence_scope`、`platform_evidence`、`authority_evidence`、`scan_receipt`、`supply_chain_bundle`. `EvidenceSetBindingDigest` equals the opaque materialized view's P08-frozen `TALENRO-C12-COMPLETION-MATERIALIZED-EVIDENCE-BINDING-V1` digest over exactly `{CounterIdentityDigest,I,materialization-intent T,its selected post-intent P,run,generation,profile,root,ordered exact-eight role/object/canonical digests}`. It is computed before materialization-actual T and excludes current/future epoch/full digest、actual-or-later T/P、all set-record digests and inner/execution state；materialization-actual then binds this digest/result one-way. Every manifest `CanonicalDigest` equals that view role's copied canonical bytes. `ValidateCompletionInputs` places these values into the opaque token only from the operator-constructed view inputs；caller overrides, reordered/duplicate/missing roles, token/view/set/run/profile splice or a semantically valid manifest with a different structural binding rejects before staging.

The fixed scope order is PowerShell, Git Bash, Linux platform/operator trust, authority fence/PITR. In each mode the composite derives the canonical input/materialized-view digests from its exact retained view and fixed projections, then privately binds Plan 08's sealed completion-child runtime to the same coordinator+inner-cleanup pair and calls it with the closed create/revalidate purpose. Plan 08 generates the fresh nonce、spawns/authenticates the installed child and returns an immutable sealed result；inside that exact child, c12evidence's zero-argument inherited runtime adapter alone mints the private native-client/provider provenance, loads the fixed provider profile and performs one signed fetch. Parent-side composite atomically consumes the P08 result, strict-decodes the sole `CompletionValidationChildResponseV1`, constant-time matches every purpose/input/view/nonce/profile/executable/process/provider/runtime field and response digest, then calls Task 4 `ValidateRetainedTrustedTimeEvidence` before using `IssuedAt` as the sole `trusted_now`; raw pipe bytes or a mutated non-authoritative view never mint a Go token. It then constructs `CompletionInputs` in-process from the same view and package-derived `CompletionPolicy`, calls the sole semantic validator and privately retains its token. Revalidate every outer and nested signature/attestation and expiry against that time；rerun Plan 08 `ValidateTrackedC12SpecDigest` for the exact shared committed tree locator/canonical digest；and require the same builder-derived `SpecDigest`, repo/toolchain/three releases, full operations tuple/combined/record-set/bundle/catalog/process-policy/image/full-scan bindings, successful cleanup and production provider class for last two scopes. The nested Platform and Authority source identities, embedded creation statement source, manifest projection and every fresh revalidation statement must constant-time equal the package-derived policy identity；different valid sources cannot share or compare expiry domains. Create sets `CreatedAt = trusted_now` and `ExpiresAt = min(CreatedAt+72h, CompletionPolicy.NotAfter, Scopes[0..3].ExpiresAt, PlatformEvidenceV1.ExpiresAt, AuthorityFenceEvidenceV1.ExpiresAt)` exactly；it may never outlive the policy cap or any of those six signed envelopes.

The manifest embeds the complete bounded signed creation-time `C12TrustedTimeEvidenceV1` envelope (including its public attestation), the exact canonical input digest and 32-byte nonce needed to recompute its typed `completion_create` challenge, plus redundant source identity、canonical envelope digest and sequence fields. Canonical manifest decoding requires those projections byte/digest-equal and nonzero；the nonce is public anti-replay data, not a credential. Revalidation first reconstructs the original fixed input digest from the retained eight files/committed projections, recomputes the original challenge from stored purpose/input digest/nonce, calls Task 4 `ValidateRetainedTrustedTimeEvidence` on the embedded envelope, revalidates its full envelope digest, and requires `CreatedAt == CreationTrustedTimeEvidence.IssuedAt`, source/sequence equality、policy-source equality and `Floor <= CreatedAt <= ValidUntil`. It then strict-decodes the P08 guard-authenticated latest accepted response head when present and requires it to be a valid same-source successor of creation. Only then does it call the same profile-constructed client's Task 4 `FetchAndValidateFresh` method for a new `completion_revalidate` purpose/challenge from that same source with sequence strictly greater than and floor/IssuedAt nondecreasing from the latest head (or creation when no head exists), use the fresh `IssuedAt` for current expiry checks, and reject at one nanosecond after the stored minimum expiry. A completed edge and an aborted edge reached after valid time both advance that head；an abort before valid time preserves it. A manifest that retains only a digest/sequence without the signed creation envelope and challenge preimage, or a later generation that compares only against creation and rolls back behind the latest head, is invalid.

All four scopes, the explicitly supplied nested PlatformEvidence, explicitly supplied nested AuthorityFenceEvidence, the completion manifest and package-derived `CompletionPolicy` must carry constant-time identical repo/tree/spec/toolchain/releases and scanner-derived values. `PlatformEvidenceV1.TrustedTimeSourceIdentity`、`AuthorityFenceEvidenceV1.TrustedTimeSourceIdentity`、`C12CompletionManifestV1.TrustedTimeSourceIdentity`、the embedded creation envelope source and every fresh revalidation envelope source must also constant-time equal the policy's sole `ExpectedTrustedTimeSourceIdentity`. Recompute both complete signed nested envelope digests, require the Linux scope's `TestResultDigest` to equal the exact `TALENRO-C12-NESTED-PLATFORM-RESULT-V1` domain formula and the authority scope's `TestResultDigest` to equal the exact `TALENRO-C12-NESTED-AUTHORITY-RESULT-V1` domain formula；a path digest, unsigned payload, embedded duplicate, missing nested file、swapped domain or time-source splice rejects. Tests independently make each of the six envelope expiries and the policy cap earliest, require the manifest to select exactly that instant, then reject revalidation one nanosecond after it. The composite validates the clean closed-set receipt and canonical bundle through Plan 08, derives the policy only from those authenticated fixed projections plus the provider profile and never accepts it from command/child bytes；`completion.go` neither imports artifactscan nor parses/recomputes its receipt/domain. There is no Plan 09 spec-set/artifact builder or digest override.

- [ ] **Step 4: GREEN — implement non-bypassable scanner/license/document gates**

Require scanner version/ruleset and clean closed-set receipt；distinct Xray MPL-2.0, sing-box GPLv3+, context-init, inner-daemon, verifier, verifier-test-assets and authority-operations auxiliary binary/image SBOM+source/provenance+vulnerability+license/notice/secret-scan records；Redis production status `blocked_pending_written_approval_or_commercial_license`；threat model and all three runbooks must pass their contract digest. For operations, artifactscan revalidates the nine typed record bodies in `SupplyChainEvidenceBundleV1` against the receipt's fixed references, aggregate record-set digest and bundle digest；`LicenseRecordSetDigest` is the canonical projection of those exact license/notice references plus their referenced tracked obligation-document digests and the other catalog license records. A parallel `--license-root` set or supply-chain bundle that is clean but not byte/digest-equal to this receipt-bound projection rejects. Auxiliary classification is not a fourth release and grants no license exemption. `FreshV7StagingEvidenceV1` may be retained only with `statement=safe_staging_established_and_closed` and `restore_incomplete`; it is never an accepted legacy-upgrade-completion input or status. Any missing/changed/license-bypassed record, bundle splice, staging-as-complete claim or reopened route rejects completion.

- [ ] **Step 5: Run completion tests and verify GREEN**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-completion-semantic-gate.ps1`

Expected: PASS for one same-build four-scope implementation fixture；duplicate/missing scope, cross-build/spec-set/tuple/component/combined/record-set/bundle/catalog/process-policy/image/full-scan splice, old SpecDigest, zero/binary-only operations digest, nested platform/authority mismatch, same artifact with a different clean record/license set, expired evidence, fake production scope, cleanup failure, stale scanner, missing core license, Redis bypass and `fresh_v7_staging_closed` legacy-completion claim all fail. The transaction tests additionally pass only when outer `publish_pending` durability precedes `BeginAdoptionSet`, every adopt/ACK transition is same-set and durable, crash recovery exact-replays through `RecoverAdoptionSet`/`RecoverAcknowledge`, inner cleanup cannot name retained publish objects, and `retained_complete` rejects every mutation/cleanup/adoption/ACK attempt.

- [ ] **Step 6: GREEN — expose only the locked native consolidation/completion parents**

Freeze two mutually exclusive zero-argument operator modes on the exact machine-base Windows dispatcher `C:\Program Files\Talenro\C12\talenro-c12-runner-launcher.exe`；it guard-selects and directly spawns the exact versioned completion helper before that helper becomes the actual parent:

- The fixed Windows launcher first authenticates its base attestation and guard-selected completion install record, creates the helper suspended in a kill-on-close Job, and guard-installs+rereads the execution reservation before resume. `completion-consolidate-create` then calls only high-level `InspectCompletionCreateRecoveryDispatch(ctx)`. On `fresh_open`, this actual selected helper calls `OpenInstalledNativeRunnerSession(..., completion)` exactly once and immediately passes the opaque session to `BeginCompletionPublishPending`；the adapter consumes the matching launcher binding/reservation, rechecks the exact fresh lineage, validates helper/private-child/PowerShell/Git/tool receipts, guarded-claims `active_unstarted(N)` and drives the transition-closed same-record replacement to `publish_pending` before first channel/provider/root access. There is no launcher or tracked-script role receipt. On `publish_recovery`, the launcher must first prove the prior bound Job/process tree absent (or a different attested boot), replace only the execution reservation and then let the selected helper call only pathless `RecoverCompletionPublishPending(ctx)`；the composite privately invokes raw P08 recovery and never Opens again. On `terminal_audit`, the launcher first proves and clears the prior `exiting` reservation, installs a distinct non-consuming audit reservation, and only then lets the helper call high-level `InspectCompletionRetainedFinalization(ctx)` to report the already-completed create generation. Every target rechecks the high-level route. `retained_complete_active` and every revalidation-only phase return no route. The helper、launcher and command have zero reference to `InspectFixedCompletionRoleState`、`RecoverFixedCompletionRoleState`、`InspectFixedCompletionRetainedFinalization` or a raw coordinator. The closed state binds full epoch/previous/counter/install/boot/execution facts plus run/session/generation、the exact ordered three transaction kind/slot/sequence expectations、stable strict-eight-file root、manifest temp/final、`completed`、three ACK reservations、adoption-set reservation and recovery plan. `PublishBindingDigest` covers the outer publication subset but excludes execution and inner-cleanup substates. An authenticated older cell、unknown/default/extra field、dual-live parent or a resource create before this barrier fails with zero channel/provider/root access.
- Only after the guard-selected outer role-state is durable may the parent call `handle.AdoptExpectedEvidence(ctx)`. That zero-selector facade privately starts or recovers one exact adoption set and drives the three fixed channel bindings in order；P08 still durably cross-persists every intent/actual and reciprocal publish digest, but receiver、binding、set、slot、sequence、nonce and lease never leave the composite. It never pre-Inspects、lists、begins another set or selects a transaction. Only completion verifies all three commit/tree/source-runner attestations and byte-identical canonical receipt/bundle；missing/duplicate/renamed/extra roles、prepared/quarantined/wrong-sequence/foreign-set transaction、receipt/bundle difference or channel/source splice rejects before trusted-time/provider access while the opaque `publish_pending` handle retains exact recovery authority.
- After all three adoption actuals, `c12completionpublication.BindCompletionPublicationOperator(handle)` is the only high-level factory. `MaterializeAdoptedEvidence(ctx)` privately drives the low-level materialization and retains the immutable exact-eight view. Before any temporary Git snapshot、provider credential、private child、parent/child session、WAL or run root, every fresh or recovered parent first calls `InspectCompletionInnerCleanupRecoveryDispatch(ctx,handle)` and follows exactly one five-way target. `begin_required` alone calls `BeginCompletionInnerCleanup` and returns a new validation-required paired session. `session_validation_recovery` alone calls `RecoverCompletionInnerCleanup` and returns a same-owner validation-required session. Create-only `session_cleanup_recovery` alone calls that facade；inside runnerprofile it first finishes a set-first pending reciprocal `manifest_intent` from the stored exact bytes without touching the manifest filesystem or any semantic dependency, then returns the same capsule as a cleanup-only session. `finalization_recovery` alone calls no-session `RecoverCompletionInnerFinalization`. `complete` calls no inner method and proceeds only to rebound outer `RecoverPublication`. The composite privately pairs the runnerprofile handle and c12evidence cleanup capability, seals the resume class and exact-excludes all retained publish objects. Only validation-required may enter `ValidateMaterializedEvidence(ctx,inner)`, which accepts no request/client/policy/member/token；internally the composite binds/runs the P08 child runtime, atomically consumes its sealed result, strict-decodes and cross-matches the sole c12evidence wire DTO, validates the signed fresh time envelope, constructs `CompletionInputs` from the same view plus fixed authenticated projections and retains the opaque validator token. `StageValidatedManifest(ctx)` alone extracts canonical bytes from that token and calls the low-level sink, whose frozen structural decoder independently checks the view binding. Cleanup-only rejects both calls and proceeds directly to exact cleanup/finalization；after that only `RecoverPublication` may replay the durable bytes. The coordinator derives the five-way result from authenticated inner tag、outer progress and same-set pending transition：a genuinely pre-manifest `none` with no pending transition must Begin/retry；a selected pre-manifest state requires validation recovery；a selected create capsule with either set-first pending or outer `manifest_intent|manifest_actual` requires cleanup recovery；only `manifest_intent`-or-later plus target-completed `none` is complete. Every action target rechecks all five values before work；no raw cleanup/coordinator/runtime/result handle escapes and no error falls back to another target.
- The native `finally` stops/verifies the child and exact-cleans the inner temporary set through `cleanup_capsule/WAL-ready handle → resource absent → terminal WAL → finalization_pending(receipt,tombstone,fixed target) → leaf Recover/Commit → key absence → tombstone unlink+parent-directory fsync+exact absence → final target CAS`. Target selection has no later filesystem work. Pre-capsule crashes converge only through `protected_key_intent → destroy/absence → inner_bootstrap_restart_pending → replacement intent → cleanup_capsule → selected-WAL Recover/Open/bootstrap/actual`; before first intent CAS both key/WAL counts are zero. `FinalizeCompletionInnerCleanup` consumes either sealed class only after every proof and, apart from set-first state-only recovery, preserves outer fields、canonical bytes、P and generation. Only afterward may publication/marker reach durable, then fixed ACK/finalization proceed. Crash recovery reopens the opaque publish handle, maps the five-way dispatch and invokes exactly Begin-validation、session-validation、session-cleanup、no-session-finalization or actionless complete. Cleanup-only cannot Validate/Stage and uses only stored-byte publication recovery；after pending selection, only no-session finalization may finish key/tombstone/target, never capsule reopen or Begin. Genuine cleaned pre-manifest retry with no set pending is Begin；proved complete rejects Begin forever. No route exposes a leaf binding、recorder or raw capability.
- The outer `publish_pending` variant remains the sole logical recovery/mutation authority through finalize-ready. Provider/child/signature/cleanup failure leaves no affirmed manifest；post-rename recovery never regenerates different bytes. Only after `completed_durable` and facade ACK success does `handle.FinalizeRetainedComplete(ctx)` freeze the exact E2 target transition and separate immutable retained-payload/finalization-set audit digest, privately drive E→E1 Prepare/Recover and proof-bound E1→E2 Commit, and mark the matching reservation `exiting`. After E2 the publish handle is invalid. A later create audit first clears that old reservation, installs a distinct read-only reservation and obtains only `CompletionTerminalAuditReceipt`; it verifies the immutable digest plus original-E2→zero-or-more-closed-`completed|aborted`-revalidation→current-inactive-retained lineage, never the mutable `next` envelope. There is no prewritten terminal candidate or second active bit, and `retained_complete_active` rejects audit.
- The sole crash-stable source for an unpublished completion manifest is the reciprocal bytes carried by `manifest_intent|manifest_actual`. `StageValidatedManifest` must finish the bounded exact-byte intent barrier before touching the temp object；afterward `RecoverPublication` reads no in-memory token or semantic input, rewrites/verifies only those stored bytes against their stored digest and reserved identity, and never reruns provider、child、validation or canonicalization. Both copies remain through reciprocal `manifest_actual`, survive temp-directory-entry or un-fsynced-rename loss, and are removed only after rename、parent-directory fsync and both sides durably select `manifest_durable`.
- The fixed launcher prepares the matching reservation before `completion-revalidate` calls the sole three-way phase router. Exact `retained_complete_active` is eligible only after old-containment absence proof and the same-generation `revalidation_recovery` replacement reservation is installed+reread；the router returns only `active_recovery` and forces selector-free `RecoverCompletionRetainedRevalidation`. At inactive `retained_complete`, durable `outcome_delivery_pending` returns only `outcome_audit` and forces `InspectCompletionRetainedRevalidationOutcome` → protected result-slot file+directory fsync → explicit assignment-identical audit+delivery-receipt ACK → reservation absence/clear；pending absent returns only `fresh_open`, after which the separately guarded Open may claim the next generation. All three targets repeat the exact phase/reservation/generation check and reject the other two values；a dispatch/target error never falls through. Before temporary validation resources the parent obtains the opaque inner session. The parent first validates the retained creation baseline, then runs/consumes the protected child and validates the fresh head. On semantic success `FinalizeRetainedCompletion` exact-cleans, prepares the leaf finalization receipt and passes receipt+validated token+matching head candidate to the P08 completed method. Every fail-closed route exact-cleans/prepares the same receipt and invokes exactly one split abort method：pre-time preserves the prior head；post-time supplies the candidate and advances it. The P08 inner pending state stores receipt view plus fixed outcome target before leaf commit/final CAS. Opposite outcome returns conflict and active crash recovery is selector-free. Terminal CAS preserves payload/audit digest, appends the closed edge, installs outcome-delivery pending, returns inactive and marks the reservation exiting. A second revalidation is impossible until durable delivery+ACK and, afterward, must be strictly newer than the latest head even when the prior edge aborted. The child and parent independently recompute the exact tracked tree from the same sealed reader；neither route rewrites retained publication objects.

Implement `testdata/c12/completion/runner-profile.v1.json` as the strict `Role=completion` generic policy instance. `AllowedTrackedScriptTokens=[]` because both public modes spawn only the prebuilt private child；`RequiredToolRoles=[c12_runner_helper_windows_amd64,talenro_artifact_scan_windows_amd64,powershell,git]` in the fixed Plan 08 order；`HelperCatalogRole=c12_runner_helper_windows_amd64`；`BootstrapLauncherCatalogRole=c12_runner_launcher_windows_amd64`；`RecoverySlotPurpose=completion_role_state_recovery`；`ExternalRootPurpose=completion_work_root`；`EvidenceRootPurpose=completion_eight_file_root`；`ProviderCredentialPurpose=completion_trusted_time`；and `InstallReceiptPolicy` closes Windows/amd64 plus those role-generation application-control roles but excludes the machine-base launcher. `ChannelBindings` is exactly three ordered receiver records：`{TransactionKind: windows_scopes, Access: receiver, SlotPurpose: windows_scopes_next, SequencePolicy: next}`、`{TransactionKind: platform_scope, Access: receiver, SlotPurpose: platform_scope_next, SequencePolicy: next}`、`{TransactionKind: authority_scope, Access: receiver, SlotPurpose: authority_scope_next, SequencePolicy: next}`. The tracked file contains no canonical ODB、machine root/slot、guard counter/digest/epoch、absolute path、launcher attestation、receipt、namespace、slot identity/sequence、credential、attestor、nonce or output fact. Step 9 commits this policy token；Step 10 rewrites only its declared catalog/toolchain role projections, and Step 12 seals the machine-local facts/receipts/guard identities.

Each public parent validates its own identity and the exact prebuilt child against the final catalog/toolchain and install receipts before reading evidence, but invokes only the composite facade. The composite alone asks Plan 08's coordinator+inner-bound runtime to spawn private `verify-completion-child` with the fixed inherited-handle table；it never receives the provider handle or a spawn primitive. The child authenticates the inherited P08 session, then only c12evidence's `OpenInheritedCompletionTrustedTimeRuntime` may privately mint the sealed native-client/provider provenance and perform one matching fetch. It canonical-encodes only `CompletionValidationChildResponseV1` through the inherited pipe；a Go interface/token never crosses the process boundary. The parent-side composite consumes the immutable P08 runtime result once, strict-decodes the DTO, matches response length/digest、pipe/session identity、child exit、profile/input/view/nonce/process/provider/runtime bindings and revalidates the signed trusted-time evidence before constructing `CompletionInputs` and minting the package-local semantic token. Raw child bytes、caller `CompletionPolicy`、caller client、temporary executable、PATH lookup、reopened/numeric handle or a second child can never mint that token. The P08 sink alone owns the reserved manifest temp, and neither child nor command receives its path/write handle. Cleanup failure is nonzero and cannot publish/affirm a manifest. Tests inject every profile/bootstrap/install receipt、guard/cell/epoch/digest、outer facade、snapshot/root/credential/child-response/token/manifest-temp/paired-cleanup/publish/ACK/E2/outcome/lineage seam and assert no undeclared resource or replay-restored authority remains. The private child mode stays absent from public help and has no public/go-run ingress.

Both modes reject stdin、any argument、glob/path/environment override、symlink/reparse escape、mixed mode and any output inside the repo. Tests cover each nonenumerating adoption/lease/ACK state, missing/wrong/swapped nested evidence, nested signature/domain/runner-identity mismatch, path-digest substitution, missing/wrong bundle, cross-transaction receipt/bundle disagreement, same artifact with a different clean bundle/license projection, absent parent commit, 40-hex Git-OID-as-digest/padded-SHA-1/current-tree substitution, wrong/replayed/rollback trusted-time response and no pre-clean final output. Parent/child CLI tests additionally cover absent/non-inherited/reused protected handle、wrong parent session/PID/start token、child binary/client-build/profile/catalog/toolchain splice、ambient global/context/HTTP-client injection and direct `go run`；the fake provider is reachable only through the test implementation of the sealed native bridge.

- [ ] **Step 7: Run CLI tests and verify GREEN**

Run first, after the last Task 7 runnerhelper/composite source change:

```text
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-platform-semantic-gate.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-authority-semantic-gate.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-completion-semantic-gate.ps1
```

Then run the eight P08/launcher upstream roots through an independent global exact-list/JSON gate, not a broad prefix:

```powershell
$expected = @(
  'TestEvidenceChannelBindingRejectsBusinessDriftWithFreshStateEpoch',
  'TestEvidenceChannelBindingSurvivesExecutionReservationOnlyCAS',
  'TestEvidenceChannelBindingSurvivesInnerCleanupOnlyCAS',
  'TestNativeRunnerCompletionRevalidationAbortClosesGeneration',
  'TestNativeRunnerCompletionTerminalAuditSurvivesAbortedRevalidations',
  'TestNativeRunnerCompletionTerminalAuditSurvivesCompletedRevalidations',
  'TestNativeRunnerLauncherDeathCannotLeaveTwoLiveParents',
  'TestNativeRunnerLaunchReservationCrashSeams'
)
$pattern = '^(' + (($expected | ForEach-Object { [regex]::Escape($_) }) -join '|') + ')$'
$listed = @(& go test ./... -list $pattern)
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$actual = @($listed | Where-Object { $_ -cmatch '^Test[A-Za-z0-9_]+$' } | Sort-Object)
$delta = @(Compare-Object -CaseSensitive $expected $actual)
if ($delta.Count -ne 0) { $delta | Out-String | Write-Error; exit 1 }
$jsonLines = @(& go test ./... -json -run $pattern -count=1)
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$events = @($jsonLines | ForEach-Object { $_ | ConvertFrom-Json })
foreach ($name in $expected) {
  $run = @($events | Where-Object { $_.Test -ceq $name -and $_.Action -ceq 'run' })
  $pass = @($events | Where-Object { $_.Test -ceq $name -and $_.Action -ceq 'pass' })
  $skip = @($events | Where-Object {
    $_.Action -ceq 'skip' -and ($_.Test -ceq $name -or (($_.Test -is [string]) -and $_.Test.StartsWith("$name/", [System.StringComparison]::Ordinal)))
  })
  if ($run.Count -ne 1 -or $pass.Count -ne 1 -or $skip.Count -ne 0) { throw "upstream completion root did not run/pass exactly once without skip: $name" }
}
```

Expected: PASS with sanitized finite errors and no partial output on failure. Windows launcher help exact-covers its three run/two provision modes；unbound helper direct invocation rejects. A fresh create/revalidate generation opens exactly one completion session；a recovered active generation proves prior Job/process absence and uses only its mode-specific opaque recovery handle, never Open. Commands cannot name low-level channel/sink/view/runtime or inject client/policy/child bytes. Outer `publish_pending` precedes facade adoption；execution/inner-only CAS preserves `PublishBindingDigest`, outer publication changes cross-persist it, and launcher death/reboot between every pre-E2 seam converges without dual parent. Three ACK actuals precede the two guarded E→E1→E2 terminal advances；post-E2 create can only audit the immutable payload/set digest after reservation clear, including after any closed completed/aborted revalidation lineage. A failed revalidation exact-cleans and returns inactive through `AbortRetainedRevalidation`, never remains wedged or records success.

- [ ] **Step 8: REFACTOR — fuzz completion decoding and run static gates**

Run: `go test ./internal/c12evidence -run '^$' -fuzz '^FuzzCompletionManifest$' -fuzztime=10s -timeout 30s`

Rerun the Step 1-defined `TestCompletionLauncherImportClosureRemainsB10Frozen` from `internal/c12completionpublication/operator_test.go`; its frozen target matrix、sanitized locked-tool invocation、exact dependency closures and forbidden launcher dependencies must remain unchanged after the Task 7 GREEN implementation.

Rerun the Step 1-defined `TestCompletionStaticGateExactCoversDeclaredTests`; Step 8 may strengthen its implementation but may not add a new root、move either exclusion definition or change the frozen four-root member list.

Run this exact fail-closed PowerShell gate from the repository root:

```powershell
$pattern = '^(TestCompletionStaticGateExactCoversDeclaredTests|TestCompletionLowLevelChannelCallsOwnedOnlyByCompositePackage|TestCompletionRawRunnerprofileCallsOwnedOnlyByCompositePackage|TestCompletionLauncherImportClosureRemainsB10Frozen)$'
$expected = @(
  'TestCompletionLauncherImportClosureRemainsB10Frozen',
  'TestCompletionLowLevelChannelCallsOwnedOnlyByCompositePackage',
  'TestCompletionRawRunnerprofileCallsOwnedOnlyByCompositePackage',
  'TestCompletionStaticGateExactCoversDeclaredTests'
)
$listed = @(& go test ./internal/c12completionpublication -list $pattern)
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$actual = @($listed | Where-Object { $_ -match '^Test' } | Sort-Object)
$delta = @(Compare-Object -CaseSensitive $expected $actual)
if ($delta.Count -ne 0) { $delta | Out-String | Write-Error; exit 1 }
$jsonLines = @(& go test ./internal/c12completionpublication -json -run $pattern -count=1)
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$events = @($jsonLines | ForEach-Object { $_ | ConvertFrom-Json })
foreach ($name in $expected) {
  $run = @($events | Where-Object { $_.Test -ceq $name -and $_.Action -ceq 'run' })
  $pass = @($events | Where-Object { $_.Test -ceq $name -and $_.Action -ceq 'pass' })
  $skip = @($events | Where-Object {
    $_.Action -ceq 'skip' -and
    ($_.Test -ceq $name -or (($_.Test -is [string]) -and $_.Test.StartsWith("$name/", [System.StringComparison]::Ordinal)))
  })
  if ($run.Count -ne 1 -or $pass.Count -ne 1 -or $skip.Count -ne 0) {
    Write-Error "static gate test did not run+pass exactly once or has a skipped descendant: $name"
    exit 1
  }
}
```

The exact four-name comparison makes zero/partial/extra matches fail before execution；the JSON event check requires each root test to run and pass exactly once and rejects a skip on the root or any `$name/...` descendant, while the meta-test proves definitions are unique and no transitively reachable package-local helper can skip.

Add `TestSharedOrdinaryFinalizationRawRunnerprofileCallsOwnedOnlyByCleanupFacade` and `TestFinalProducerRawBridgeStaticGateExactCoversDeclaredTests` in `internal/c12evidence/completion_test.go`. The shared test freezes the exact raw set `{FinalizeFixedCleanedOwnership,RecoverFixedCleanedOwnership}`, permits production references only in `internal/c12evidence/cleanup.go`, requires exported high-level `FinalizeCleanedOwnership`/`RecoverCleanedOwnership` to return no raw capability/receipt/session and rejects every command、runner-helper or alternate forwarding reference. The meta-test uses `go/parser` to require exactly one top-level definition of itself、that shared test、`TestPlatformRawRunnerprofileCallsOwnedOnlyByEvidenceFacade`、`TestAuthorityRawRunnerprofileCallsOwnedOnlyByEvidenceFacade` and P08's `TestProtectedWALKeyBootstrapRawRunnerprofileCallsOwnedOnlyByCleanupFacade`, rejecting duplicates. It then uses `go/types` call resolution to reject `Skip`/`Skipf`/`SkipNow` in all five roots plus every transitively reachable package-local helper/function literal, with unresolved dynamic helper dispatch failing closed. Its frozen substantive member list contains exactly platform、authority、shared ordinary-finalization and shared protected-key-bootstrap. It exact-compares their literal role-specific platform、shared trusted-time、authority、shared ordinary-finalization and protected-key raw/high-level owner sets against the complete categorized P08 bridge API, including Task 7's sender/finalizer symbols、high-level `BindCleanupOnlyOwnershipPlan` with fixed role-builder callers、raw `BindFixedCleanupOnlyOwnershipPlan|InspectFixedCleanupOnlyOwnershipPlan` with sole `cleanup.go` ownership and the guarded selected-WAL consumer, so a newly added、misowned or silently omitted bridge API fails closed.

That complete categorization includes the exact shared post-clean digest singleton `{FixedPostCleanupPublicationInputsDigest}`. Its domain/hash implementation exists only in `internal/c12runnerprofile/profile.go`; outside that definition, the complete production owner set is exactly the three P08 role-state owners `internal/c12runnerprofile/{windows_cleanup,platform_resume,authority_cleanup}.go` plus the three high-level consumers `internal/c12evidence/{runnerhelper,platform,authority}.go`. For each role, the frozen resolved call graph requires both the success-origin CAS and current-state Inspect validator in its runnerprofile owner, plus both Publish and Recover in its c12evidence owner, to reach that same function with the fixed role/arity/order；one unexported role-local wrapper may provide the single direct reference for its two high-level paths, but no command、alternate package or seventh consumer file may reference it. The meta rejects an inline/duplicate hasher、function value、forwarder or digest-only bypass. It also exact-cross-checks the shared sealed-producer category against P08's Windows owner root and the Authority high-level owner table above：`BindFinalClosedSetProjectionSink` plus `artifactscan.ValidateAndSealFinalClosedSetProjection` have exactly the two Windows/Authority fixed-parent edges, while `SealValidatedClosedSet` remains called only in `internal/artifactscan/scan.go`. These are submatrices of the existing meta root and do not change either frozen manifest count.

Run this second exact fail-closed PowerShell gate from the repository root:

```powershell
$bridgePattern = '^(TestAuthorityRawRunnerprofileCallsOwnedOnlyByEvidenceFacade|TestFinalProducerRawBridgeStaticGateExactCoversDeclaredTests|TestPlatformRawRunnerprofileCallsOwnedOnlyByEvidenceFacade|TestProtectedWALKeyBootstrapRawRunnerprofileCallsOwnedOnlyByCleanupFacade|TestSharedOrdinaryFinalizationRawRunnerprofileCallsOwnedOnlyByCleanupFacade)$'
$bridgeExpected = @(
  'TestAuthorityRawRunnerprofileCallsOwnedOnlyByEvidenceFacade',
  'TestFinalProducerRawBridgeStaticGateExactCoversDeclaredTests',
  'TestPlatformRawRunnerprofileCallsOwnedOnlyByEvidenceFacade',
  'TestProtectedWALKeyBootstrapRawRunnerprofileCallsOwnedOnlyByCleanupFacade',
  'TestSharedOrdinaryFinalizationRawRunnerprofileCallsOwnedOnlyByCleanupFacade'
)
$bridgeListed = @(& go test ./internal/c12evidence -list $bridgePattern)
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$bridgeActual = @($bridgeListed | Where-Object { $_ -match '^Test' } | Sort-Object)
$bridgeDelta = @(Compare-Object -CaseSensitive $bridgeExpected $bridgeActual)
if ($bridgeDelta.Count -ne 0) { $bridgeDelta | Out-String | Write-Error; exit 1 }
$bridgeJSONLines = @(& go test ./internal/c12evidence -json -run $bridgePattern -count=1)
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$bridgeEvents = @($bridgeJSONLines | ForEach-Object { $_ | ConvertFrom-Json })
foreach ($name in $bridgeExpected) {
  $run = @($bridgeEvents | Where-Object { $_.Test -ceq $name -and $_.Action -ceq 'run' })
  $pass = @($bridgeEvents | Where-Object { $_.Test -ceq $name -and $_.Action -ceq 'pass' })
  $skip = @($bridgeEvents | Where-Object {
    $_.Action -ceq 'skip' -and
    ($_.Test -ceq $name -or (($_.Test -is [string]) -and $_.Test.StartsWith("$name/", [System.StringComparison]::Ordinal)))
  })
  if ($run.Count -ne 1 -or $pass.Count -ne 1 -or $skip.Count -ne 0) {
    Write-Error "raw bridge gate did not run+pass exactly once or has a skipped descendant: $name"
    exit 1
  }
}
```

The bridge list comparison rejects zero/partial/extra matches, the JSON events require all five root tests to run and pass exactly once with no root/descendant skip, and the go/types meta-test makes the reachable-helper rule、owner maps and categorized raw symbol sets non-weakenable.

Run: `go vet ./internal/c12evidence ./internal/c12completionpublication ./internal/c12runnerprofile ./cmd/talenro-artifact-scan ./cmd/talenro-c12-runner-launcher ./cmd/talenro-c12-runner-helper`

Expected: PASS.

- [ ] **Step 9: Commit completion validation before claiming completion**

```bash
set -euo pipefail
b11_source_paths=(
  internal/c12evidence/completion.go
  internal/c12evidence/completion_test.go
  internal/c12completionpublication/operator.go
  internal/c12completionpublication/operator_test.go
  internal/c12evidence/runnerhelper.go
  internal/c12evidence/runnerhelper_test.go
  cmd/talenro-artifact-scan/main.go
  cmd/talenro-artifact-scan/main_test.go
  cmd/talenro-c12-runner-helper/main.go
  cmd/talenro-c12-runner-helper/main_test.go
  testdata/c12/completion/runner-profile.v1.json
  scripts/invoke-exact-completion-semantic-gate.ps1
)
b11_relock_outputs=(
  deploy/c12/compose.outer.yaml
  testdata/c12/integration-manifest.v1.json
  testdata/c12/toolchain-lock.v1.json
  testdata/c12/platform/trusted-time-provider-profile.v1.json
  testdata/c12/windows/runner-profile.v1.json
  testdata/c12/diagnostic/runner-profile.v1.json
  testdata/c12/platform/runner-profile.v1.json
  testdata/c12/authority/runner-profile.v1.json
  testdata/c12/completion/runner-profile.v1.json
)
b11_preexisting_relock_outputs=(
  deploy/c12/compose.outer.yaml
  testdata/c12/integration-manifest.v1.json
  testdata/c12/toolchain-lock.v1.json
  testdata/c12/platform/trusted-time-provider-profile.v1.json
  testdata/c12/windows/runner-profile.v1.json
  testdata/c12/diagnostic/runner-profile.v1.json
  testdata/c12/platform/runner-profile.v1.json
  testdata/c12/authority/runner-profile.v1.json
)
[[ "${#b11_source_paths[@]}" -eq 12 ]]
[[ "${#b11_relock_outputs[@]}" -eq 9 ]]
[[ "${#b11_preexisting_relock_outputs[@]}" -eq 8 ]]
b11_source_expected="$(printf '%s\n' "${b11_source_paths[@]}" | LC_ALL=C sort)"
b11_relock_expected="$(printf '%s\n' "${b11_relock_outputs[@]}" | LC_ALL=C sort)"
b11_preexisting_expected="$(printf '%s\n' "${b11_preexisting_relock_outputs[@]}" | LC_ALL=C sort)"
[[ "$(printf '%s\n' "${b11_source_paths[@]}" | LC_ALL=C sort -u)" == "$b11_source_expected" ]]
[[ "$(printf '%s\n' "${b11_relock_outputs[@]}" | LC_ALL=C sort -u)" == "$b11_relock_expected" ]]
[[ "$(printf '%s\n' "${b11_preexisting_relock_outputs[@]}" | LC_ALL=C sort -u)" == "$b11_preexisting_expected" ]]
b11_overlap="$(comm -12 <(printf '%s\n' "${b11_source_paths[@]}" | LC_ALL=C sort -u) <(printf '%s\n' "${b11_relock_outputs[@]}" | LC_ALL=C sort -u))"
[[ "$b11_overlap" == 'testdata/c12/completion/runner-profile.v1.json' ]]
b11_nonpreexisting_outputs="$(comm -23 <(printf '%s\n' "${b11_relock_outputs[@]}" | LC_ALL=C sort -u) <(printf '%s\n' "${b11_preexisting_relock_outputs[@]}" | LC_ALL=C sort -u))"
[[ "$b11_nonpreexisting_outputs" == 'testdata/c12/completion/runner-profile.v1.json' ]]
b11_source_actual="$(
  {
    git diff --no-renames --name-only HEAD --
    git ls-files --others --exclude-standard
  } | LC_ALL=C sort -u
)"
[[ "$b11_source_actual" == "$b11_source_expected" ]]
git diff HEAD --exit-code -- "${b11_preexisting_relock_outputs[@]}"
git diff --check
git add -- "${b11_source_paths[@]}"
b11_cached_actual="$(git diff --cached --no-renames --name-only -- | LC_ALL=C sort)"
[[ "$b11_cached_actual" == "$b11_source_expected" ]]
git diff --exit-code
git diff --cached --check
[[ -z "$(git ls-files --others --exclude-standard)" ]]
b11_source_candidate_tree="$(git write-tree)"
git commit -m "feat: validate four-scope C1.2 completion"
[[ "$(git rev-parse 'HEAD^{tree}')" == "$b11_source_candidate_tree" ]]
b11_source_commit="$(git rev-parse HEAD)"
b11_source_tree="$(git rev-parse 'HEAD^{tree}')"
git diff HEAD --exit-code
git diff --cached --exit-code
[[ -z "$(git ls-files --others --exclude-standard)" ]]
```

The source-candidate union is exactly the twelve declared paths, and the source/relock intersection is exactly `testdata/c12/completion/runner-profile.v1.json`. Therefore the pre-source gate requires the other eight relock outputs byte-identical to `HEAD`, while the overlapping completion profile is intentionally committed here once as a policy/digest-free source and rewritten only by Step 10. The cached tree must equal the verified candidate tree, and the source commit returns tracked worktree、index and untracked state all clean. Steps 9–12 are excerpts of one long-lived protected central Bash orchestration；it retains `b11_source_commit|b11_source_tree` and later `b11_projection_commit|b11_projection_tree` as shell-local values while synchronously driving the separate authenticated host transactions, never reconstructing them from caller environment or a repository file. This commit records only the policy/digest-free completion profile：role、tracked script/tool/helper tokens、purpose tokens、install policy and the three ordered channel policy bindings are canonical, while machine paths/identities、actual namespace/slot/sequence、receipts、credentials and attestators remain absent by schema. Plan 08 may strict-decode/test the tracked policy, but `OpenInstalledNativeRunnerSession` cannot mint a production session until the next atomic relock projection is committed and Step 12 seals matching machine facts.

The completion profile committed in Step 9 includes `BootstrapLauncherCatalogRole=c12_runner_launcher_windows_amd64`; its role receipts exclude the launcher. Its guard-selected state exposes both full `StateDigest` and outer-only `PublishBindingDigest` under the P08 derivation, with execution reservation and inner temporary cleanup excluded only from the latter.

- [ ] **Step 10: Rebuild and atomically relock all final plus diagnostic execution projections**

The Step 9 commit is the last B01–B11 source/helper/client/policy change. From that clean committed tree, repeat Plan 08's sole atomic rebuild/relock after rebuilding both runner-helper PE/ELF roles、the prebuilt Windows artifact-scan completion child and verifier image. Rebuild both machine-base launchers from the same frozen Task 2 source/recipe in two independent absolute roots and require their PE/ELF bytes、SBOM/provenance and native-role projections to remain byte-identical to the B10-frozen identities and existing base attestations；this is revalidation, never a launcher upgrade or receipt/output rewrite. An import-graph gate proves `cmd/talenro-c12-runner-launcher` still reaches only the B10-frozen runnerprofile graph and does not import `internal/c12evidence` or the new `internal/c12completionpublication` composite；P09 changed neither runnerprofile nor launcher source to add those packages. Before building, strict-load the three fixed tracked `deploy/c12/{context-init,inner-daemon,verifier}.build-context.v1.json` tokens, exact-cover their path/mode/blob digests, reject every rewrite output as a member, and retain their canonical source/provenance digests；the transaction revalidates but never rewrites those manifests. It also strict-loads the digest-free catalog and all five fixed runner-profile tokens, including the Plan 08-created Windows/diagnostic instances and the Task 7 completion instance. The transaction binds the canonical catalog/role/context projections and atomically rewrites exactly nine outputs：Compose、integration manifest、toolchain lock、trusted-time provider profile, then Windows/diagnostic/platform/authority/completion runner profiles.

```bash
set -euo pipefail
git diff HEAD --exit-code
git diff --cached --exit-code
[[ -z "$(git ls-files --others --exclude-standard)" ]]
[[ "$(git rev-parse HEAD)" == "$b11_source_commit" ]]
[[ "$(git rev-parse 'HEAD^{tree}')" == "$b11_source_tree" ]]
b11_relock_outputs=(
  deploy/c12/compose.outer.yaml
  testdata/c12/integration-manifest.v1.json
  testdata/c12/toolchain-lock.v1.json
  testdata/c12/platform/trusted-time-provider-profile.v1.json
  testdata/c12/windows/runner-profile.v1.json
  testdata/c12/diagnostic/runner-profile.v1.json
  testdata/c12/platform/runner-profile.v1.json
  testdata/c12/authority/runner-profile.v1.json
  testdata/c12/completion/runner-profile.v1.json
)
[[ "${#b11_relock_outputs[@]}" -eq 9 ]]
b11_relock_expected="$(printf '%s\n' "${b11_relock_outputs[@]}" | LC_ALL=C sort)"
[[ "$(printf '%s\n' "${b11_relock_outputs[@]}" | LC_ALL=C sort -u)" == "$b11_relock_expected" ]]
go run ./cmd/talenro-artifact-scan lock-outer-images --compose deploy/c12/compose.outer.yaml --context-root deploy/c12 --integration-manifest testdata/c12/integration-manifest.v1.json --output-lock testdata/c12/toolchain-lock.v1.json --artifact-catalog testdata/c12/artifact-catalog.v1.json --trusted-time-provider-profile testdata/c12/platform/trusted-time-provider-profile.v1.json --windows-runner-profile testdata/c12/windows/runner-profile.v1.json --diagnostic-runner-profile testdata/c12/diagnostic/runner-profile.v1.json --platform-runner-profile testdata/c12/platform/runner-profile.v1.json --authority-runner-profile testdata/c12/authority/runner-profile.v1.json --completion-runner-profile testdata/c12/completion/runner-profile.v1.json --rewrite-compose --rewrite-integration-manifest --rewrite-trusted-time-provider-profile --rewrite-windows-runner-profile --rewrite-diagnostic-runner-profile --rewrite-platform-runner-profile --rewrite-authority-runner-profile --rewrite-completion-runner-profile
b11_relock_actual="$(git diff --no-renames --name-only HEAD -- | LC_ALL=C sort)"
[[ "$b11_relock_actual" == "$b11_relock_expected" ]]
git diff --cached --exit-code
git diff --check
[[ -z "$(git ls-files --others --exclude-standard)" ]]
```

The child/client binaries do not embed final lock、catalog、profile or install-receipt digests；they are built only from the committed source/recipe, tracked profiles contain only `InstallReceiptPolicy` plus role/purpose tokens, and runtime validates machine-local receipts from the sealed bootstrap. Thus `source/recipe -> executable tuple -> provider profile -> five runner profiles -> toolchain lock`, followed only after commit by OS provisioning receipts, is acyclic. Compose and all eight other rewrite outputs are outside the three image contexts, so the image-digest graph is likewise acyclic. The atomic failure matrix injects failure before/after every one of the nine output rewrites, each fixed context-manifest validation, every profile/catalog read and concurrent-change check；it requires all three context manifests、all nine outputs and the catalog byte-identical to their pre-run bytes with no temp residue. A successful transaction must leave the index and untracked set empty while the working-tree delta is exactly the nine declared outputs；a missing、extra、renamed or staged path fails before the projection commit. Require the rewritten lock/catalog/provider/profile projections to exact-cover both helper roles、both frozen launcher roles and each exact binary/SBOM/provenance triple、all three trusted-time purpose/OS/arch client mappings、private completion child、verifier image、outer images、their three context digests、integration layouts, all five profile roles, install-receipt policies、tracked script/tool roles, the diagnostic empty-script/zero-channel/none-provider/nonpublishing projection and the exact ordered final sender/receiver channel policy bindings. Launchers remain B10 inputs, not a tenth rewrite output or install-receipt role. Absolute paths、actual slot identities/sequences and receipt envelopes are deliberately absent from every tracked output.

- [ ] **Step 11: Commit the final relock projection and freeze the implementation tree**

```bash
set -euo pipefail
b11_relock_outputs=(
  deploy/c12/compose.outer.yaml
  testdata/c12/integration-manifest.v1.json
  testdata/c12/toolchain-lock.v1.json
  testdata/c12/platform/trusted-time-provider-profile.v1.json
  testdata/c12/windows/runner-profile.v1.json
  testdata/c12/diagnostic/runner-profile.v1.json
  testdata/c12/platform/runner-profile.v1.json
  testdata/c12/authority/runner-profile.v1.json
  testdata/c12/completion/runner-profile.v1.json
)
[[ "${#b11_relock_outputs[@]}" -eq 9 ]]
b11_relock_expected="$(printf '%s\n' "${b11_relock_outputs[@]}" | LC_ALL=C sort)"
[[ "$(printf '%s\n' "${b11_relock_outputs[@]}" | LC_ALL=C sort -u)" == "$b11_relock_expected" ]]
b11_relock_actual="$(git diff --no-renames --name-only HEAD -- | LC_ALL=C sort)"
[[ "$b11_relock_actual" == "$b11_relock_expected" ]]
git diff --cached --exit-code
git diff --check
[[ -z "$(git ls-files --others --exclude-standard)" ]]
[[ "$(git rev-parse HEAD)" == "$b11_source_commit" ]]
[[ "$(git rev-parse 'HEAD^{tree}')" == "$b11_source_tree" ]]
git add -- "${b11_relock_outputs[@]}"
b11_cached_actual="$(git diff --cached --no-renames --name-only -- | LC_ALL=C sort)"
[[ "$b11_cached_actual" == "$b11_relock_expected" ]]
git diff --exit-code
git diff --cached --check
[[ -z "$(git ls-files --others --exclude-standard)" ]]
b11_projection_candidate_tree="$(git write-tree)"
git commit -m "build: relock final C1.2 execution identities"
[[ "$(git rev-parse HEAD^)" == "$b11_source_commit" ]]
[[ "$(git rev-parse 'HEAD^{tree}')" == "$b11_projection_candidate_tree" ]]
git diff HEAD --exit-code
git diff --cached --exit-code
[[ -z "$(git ls-files --others --exclude-standard)" ]]
b11_projection_commit="$(git rev-parse HEAD)"
b11_projection_tree="$(git rev-parse 'HEAD^{tree}')"
```

After this projection commit, the locked helper must capture the committed tree and prove its Compose/context/integration/toolchain/catalog/profile projections equal the rebuilt outputs. The authenticated central orchestration retains `b11_projection_commit` and `b11_projection_tree` outside caller control through Step 12 and exact-compares them again before Task 8；they are not read from an environment or caller file. No `internal/c12evidence`、`internal/c12completionpublication`、artifactscan、runner-helper、runnerprofile、`cmd/talenro-c12-runner-launcher`、Dockerfile/entrypoint、Compose、integration、catalog、toolchain or trusted-time/runner-profile source may change before Task 8；any change invalidates the projection commit and forces the complete Step 9 source commit → Step 10 rebuild/relock → Step 11 projection commit sequence again. Task 6A's earlier staged candidate was committed before this final relock；Task 8 authoritative scopes capture only this later clean projection commit.

- [ ] **Step 12: Provision and seal all native runner installations**

Using only the Step 11 committed catalog/toolchain/five-profile projections distributed byte-identically to the target machines, the authenticated OS deployment service opens five independent role/host-local deployment transactions. Each transaction exact-binds one role、OS/arch、projection commit and candidate package, verifies the already-installed machine-base launcher's `C12FixedRunnerLauncherAttestationV1`, and supplies its nonexportable candidate-package handle directly to that attested launcher process. The operator shell cannot mint、serialize、name or pass the handle；a launcher command outside its matching live transaction, a cross-host/replayed handle or a handle for another role/projection fails before installer-session mint or any guard/cell write. Inside each transaction invoke exactly one Plan 08 protected zero-extra-argument provision mode on its own fixed runner host：Windows-scope `provision-windows-scopes-runner`, Windows completion `provision-completion-runner`, Linux diagnostic `provision-staged-diagnostic-runner`, Linux platform `provision-platform-runner`, and Linux authority `provision-authority-runner`. A runbook never calls `OpenFixedMachineInstallerSession` or `InstallFixedNativeRunnerProfile` directly. Each mode internally and in order performs non-consuming `VerifyFixedMachineStateGuardProvider`、opens one process/mode/candidate-handle-bound installer session、calls consuming `InstallFixedNativeRunnerProfile` for only its compile-time role—which invalidates the session immediately and destroys session+handle before every success/error return—then calls non-consuming `VerifyFixedNativeRunnerInstallation` and emits one bounded role/OS/arch/projection-commit/install-generation/state-epoch/install-record-digest success record with no path/receipt/key bytes. There is no separate caller-visible destroy operation；response loss or retry must open a fresh deployment transaction/session. Any nonzero exit, missing/duplicate role record or projection-commit mismatch stops before Task 8；sessions、handles and records never cross hosts.

Within each mode the provisioner stages the role's exact PE/ELF helper、native interpreter/Git/Docker/scanner/build tools and private Windows completion child beneath installer-derived immutable generation paths under fixed application-control policy and never overwrites the selected generation. Tracked snapshot scripts are never installed and receive no install receipt；an actual public parent later derives each allowed script only from its authenticated committed read-only snapshot and revalidates exact blob SHA/mode、snapshot root/file identity and no-follow one-link regular-file constraints before receipt-pinned interpreter execution. The provisioner creates one strict machine-local `C12NativeRunnerInstallReceiptV1` per staged native executable, durably creates the complete immutable install record plus profile/catalog/toolchain and role-dependent ODB/root/slot/channel/provider/attestor projections, and binds that digest in the candidate role-state cell. No receipt is tracked or printed. The fixed bootstrap retains only stable store/guard/cell/install-root identities and no active-install pointer；one role-guard advance solely selects both install generation and runtime state. First install begins at authenticated `(epoch=0,digest=zero)` with the role still uninstalled/retryable；a pre-advance crash leaves no selected generation. It creates the role-unique non-resettable guard/fixed A/B cells, stages/fsyncs generation-1 files/receipts/install record plus epoch-1 inactive state (`idle(next=1)` for completion), advances `0/zero -> 1/initialDigest` and rereads the selected cell/install record. Reinstall is allowed only from an already selected ordinary terminal-inactive or completion `idle|retained_complete(active=nil)` state and preserves guard/cell identities、next launch generation and immutable retained payload；its pre-advance crash keeps the old selected set, while a first-install pre-advance crash does not pretend an old set exists. Active/pending or missing/reset/reprovisioned guard rejects. Every post-advance crash exact-forward-recovers only the new set. Each final Verify must return nil, mint no runtime handle/session, leave the role launch state inactive and consume no guard/launch generation. Wrong/missing/extra/stale native receipt、script receipt、pre-relock helper/launcher、cross-role/OS candidate handle、profile drift、diagnostic capability expansion、writable executable or guard/cell/install-record rollback blocks Task 8. Re-run `git diff --exit-code` and require the implementation tree, including `cmd/talenro-c12-runner-launcher`, still equals Step 11.

On each respective Windows host, run the matching provision command only after this Windows-tagged adapter/mode gate passes against the frozen launcher source:

```powershell
& powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-b10-gates.ps1
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$provisionPackages = @('./internal/c12runnerprofile', './cmd/talenro-c12-runner-launcher')
$provisionExpected = @(
  'TestInstallFixedNativeRunnerProfileConsumesSessionAndDestroysCandidateOnEveryReturn',
  'TestNativeRunnerBootstrapLauncherClosesFirstInstallAndReinstall',
  'TestNativeRunnerBootstrapLauncherUsesGuardSelectedInstallRecord',
  'TestNativeRunnerCandidatePackageHandleIsRoleProjectionAndHostBound',
  'TestNativeRunnerCompletionModePhaseCrossClaimRejects',
  'TestNativeRunnerCompletionRoleStateCoordinatorRejectsIllegalPhaseGenerationAndPayload',
  'TestNativeRunnerCompletionRoleStateCoordinatorRequiresSessionOrFixedRecovery',
  'TestNativeRunnerCompletionRoleStateInspectIsNonConsuming',
  'TestNativeRunnerCompletionTerminalFinalizationAuditIsPathlessReadOnly',
  'TestNativeRunnerFirstInstallInitializesRoleState',
  'TestNativeRunnerFixedLauncherAttestationBindsMachineBase',
  'TestNativeRunnerFixedProvisionModesAreRoleAndOSClosed',
  'TestNativeRunnerOpenFixedMachineInstallerSessionRequiresProtectedProvisioner',
  'TestNativeRunnerProvisionOutsideDeploymentTransactionRejects',
  'TestNativeRunnerReinstallPreservesLaunchStateAndRetainedPayload',
  'TestNativeRunnerWindowsBootstrapLauncherProductionAdapterRejectsFallback',
  'TestNativeRunnerWindowsStateGuardProductionAdapterRejectsFallback'
)
$provisionPattern = '^(' + (($provisionExpected | ForEach-Object { [regex]::Escape($_) }) -join '|') + ')$'
$provisionListed = @(& go test @provisionPackages -list $provisionPattern)
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$provisionActual = @($provisionListed | Where-Object { $_ -cmatch '^Test[A-Za-z0-9_]+$' } | Sort-Object)
$provisionDelta = @(Compare-Object -CaseSensitive $provisionExpected $provisionActual)
if ($provisionDelta.Count -ne 0) { $provisionDelta | Out-String | Write-Error; exit 1 }
$provisionJSONLines = @(& go test @provisionPackages -json -run $provisionPattern -count=1 -timeout 30m)
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$provisionEvents = @($provisionJSONLines | ForEach-Object { $_ | ConvertFrom-Json })
foreach ($name in $provisionExpected) {
  $run = @($provisionEvents | Where-Object { $_.Test -ceq $name -and $_.Action -ceq 'run' })
  $pass = @($provisionEvents | Where-Object { $_.Test -ceq $name -and $_.Action -ceq 'pass' })
  $skip = @($provisionEvents | Where-Object { $_.Test -ceq $name -and $_.Action -ceq 'skip' })
  $descendantSkip = @($provisionEvents | Where-Object { $_.Action -ceq 'skip' -and $_.Test -and $_.Test.StartsWith("$name/", [System.StringComparison]::Ordinal) })
  if ($run.Count -ne 1 -or $pass.Count -ne 1 -or $skip.Count -ne 0 -or $descendantSkip.Count -ne 0) { throw "pre-provision root did not run+pass exactly once without skip: $name" }
}
```

Inside the Windows-scope host's live deployment transaction, its exact transaction payload is:

```powershell
& 'C:\Program Files\Talenro\C12\talenro-c12-runner-launcher.exe' provision-windows-scopes-runner
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
```

Inside the distinct Windows completion host's live deployment transaction, its exact transaction payload is:

```powershell
& 'C:\Program Files\Talenro\C12\talenro-c12-runner-launcher.exe' provision-completion-runner
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
```

On each respective Linux host, run the matching provision command only after this Linux-tagged adapter/mode gate passes against the frozen launcher source:

```bash
set -euo pipefail
bash scripts/invoke-exact-b10-linux-adapter-gate.sh
provision_packages=(./internal/c12runnerprofile ./cmd/talenro-c12-runner-launcher)
provision_expected=(
  TestInstallFixedNativeRunnerProfileConsumesSessionAndDestroysCandidateOnEveryReturn
  TestNativeRunnerBootstrapLauncherClosesFirstInstallAndReinstall
  TestNativeRunnerBootstrapLauncherUsesGuardSelectedInstallRecord
  TestNativeRunnerCandidatePackageHandleIsRoleProjectionAndHostBound
  TestNativeRunnerCompletionModePhaseCrossClaimRejects
  TestNativeRunnerCompletionRoleStateCoordinatorRejectsIllegalPhaseGenerationAndPayload
  TestNativeRunnerCompletionRoleStateCoordinatorRequiresSessionOrFixedRecovery
  TestNativeRunnerCompletionRoleStateInspectIsNonConsuming
  TestNativeRunnerCompletionTerminalFinalizationAuditIsPathlessReadOnly
  TestNativeRunnerFirstInstallInitializesRoleState
  TestNativeRunnerFixedLauncherAttestationBindsMachineBase
  TestNativeRunnerFixedProvisionModesAreRoleAndOSClosed
  TestNativeRunnerLinuxBootstrapLauncherProductionAdapterRejectsFallback
  TestNativeRunnerLinuxStateGuardProductionAdapterRejectsFallback
  TestNativeRunnerOpenFixedMachineInstallerSessionRequiresProtectedProvisioner
  TestNativeRunnerProvisionOutsideDeploymentTransactionRejects
  TestNativeRunnerReinstallPreservesLaunchStateAndRetainedPayload
)
provision_pattern="^($(IFS='|'; printf '%s' "${provision_expected[*]}"))$"
provision_expected_text="$(printf '%s\n' "${provision_expected[@]}" | LC_ALL=C sort)"
provision_actual_text="$(go test "${provision_packages[@]}" -list "$provision_pattern" | grep -E '^Test[A-Za-z0-9_]+$' | LC_ALL=C sort)"
[[ "$provision_actual_text" == "$provision_expected_text" ]]
provision_json="$(mktemp)"
trap 'rm -f -- "$provision_json"' EXIT
go test "${provision_packages[@]}" -json -run "$provision_pattern" -count=1 -timeout 30m >"$provision_json"
for name in "${provision_expected[@]}"; do
  run_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"run\""){c++} END{print c+0}' "$provision_json")"
  pass_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"pass\""){c++} END{print c+0}' "$provision_json")"
  skip_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n) && index($0,"\"Action\":\"skip\""){c++} END{print c+0}' "$provision_json")"
  [[ "$run_count" == 1 && "$pass_count" == 1 && "$skip_count" == 0 ]]
done
rm -f -- "$provision_json"
trap - EXIT
```

Inside the Linux diagnostic host's live deployment transaction, its exact transaction payload is `/usr/libexec/talenro-c12/talenro-c12-runner-launcher provision-staged-diagnostic-runner`；inside the distinct Linux platform host transaction it is `/usr/libexec/talenro-c12/talenro-c12-runner-launcher provision-platform-runner`；inside the distinct Linux authority host transaction it is `/usr/libexec/talenro-c12/talenro-c12-runner-launcher provision-authority-runner`. Each authoritative Bash wrapper already has `set -euo pipefail`, so the gate and its one role command stop immediately on nonzero exit. Do not add source after the Step 9 freeze. Instead, before that source commit P08's already frozen roots must cover these negatives exactly：`TestNativeRunnerProvisionOutsideDeploymentTransactionRejects` owns outside-transaction；`TestNativeRunnerCandidatePackageHandleIsRoleProjectionAndHostBound` owns wrong-role/projection and cross-host/replayed handles；`TestNativeRunnerFixedLauncherAttestationBindsMachineBase` owns wrong base-launcher attestation；`TestInstallFixedNativeRunnerProfileConsumesSessionAndDestroysCandidateOnEveryReturn` owns post-destruction reuse. Step 12 reruns all four through the exact pre-provision gate；each must fail before installer-session/state access. Task 8 exact-covers the five bounded status records by role/host/projection commit before starting any evidence parent.

After all five authenticated provisioning transactions and before Task 8, the same protected central orchestration that retained the two Step 11 scalars runs the final repository proof；it neither reloads those scalars from caller environment nor permits a new commit between the comparisons:

```bash
set -euo pipefail
git diff --check
git diff HEAD --exit-code
git diff --cached --exit-code
[[ -z "$(git ls-files --others --exclude-standard)" ]]
[[ "$(git rev-parse HEAD)" == "$b11_projection_commit" ]]
[[ "$(git rev-parse 'HEAD^{tree}')" == "$b11_projection_tree" ]]
```

Thus Task 8 begins only from the exact Step 11 projection commit/tree with tracked worktree、index and untracked state all clean；a provisioning-side repository write or any later source/output change forces the full Step 9 → Step 10 → Step 11 sequence again.

### Task 8: Run final four-scope acceptance and advance the roadmap

**Files:**
- Create: `docs/security/c12-completion-record.md`
- Modify: `docs/roadmap/implementation-sequence.md`
- Test: `docs/security/c12-completion-record.md`
- Test: `docs/roadmap/implementation-sequence.md`

**Interfaces:**
- Consumes: four fresh same-build scope files, the two explicitly transferred signed nested `PlatformEvidenceV1`/`AuthorityFenceEvidenceV1` files, the byte-identical final scanner receipt and canonical supply-chain bundle containing the authority-operations tuple/records/catalog bindings, license/document digests, both native completion parent modes from Task 7 and exactly two independent final-review receipts bound to the frozen implementation tree and completion manifest.
- Produces: one external `C12CompletionManifestV1`, two external independent-review receipts, one external combined review-gate receipt, a digest-only completion record, C1.2=`complete`, and C1.3=`current design`.

- [ ] **Step 1: RED — prove consolidation fails with only a Windows channel transaction**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-completion-semantic-gate.ps1`

The fixture publishes only the non-authoritative Windows transaction into an isolated package test channel；no production mode accepts fixture paths. The attempt creates a durable isolated `CompletionPublishPendingHandle` and calls only `handle.AdoptExpectedEvidence(ctx)`. Its private composite implementation may mutate only the exact held set after matching outer intent；the missing second binding fails without enumeration, low-level capability exposure, manifest/ACK or retained transition. Expected: PASS only when the attempt returns finite `scope_set_incomplete` before trusted-time/provider access, no evidence root/manifest exists, and teardown exact-cleans that isolated handle's pending set/outer state.

- [ ] **Step 2: Run the independent locked repository final gate before evidence collection**

Before opening any native-runner sealed bootstrap、cleanup capsule、tree-facts session、provider/attestor handle、evidence root or channel slot, run P01's final repository gate once against the clean Step 11 relock commit. The gate controller pins the final toolchain-approved PowerShell executable and the exact tracked `scripts/run-c12-integration.ps1` file identity from that commit, fixes the repository root as cwd, sanitizes environment, passes only the literal closed arguments below, and owns/cleans its ephemeral PostgreSQL/Redis/NATS dependency set. It is explicitly non-evidence：it receives no runner profile/session/receipt/bundle/slot/signing authority, publishes no scope/member, and must exit with all dependencies/processes/temp state absent before Step 3 begins.

```powershell
& 'C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe' -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Suite final -Timeout 240m
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
```

`final` executes only Windows-compatible integration manifests/gates and statically verifies the complete literal custom-tag owner map, including both `c12_platform`、both `c12_authority`、both `c12_docker` and both integration-tagged infrastructure files；it never executes the real platform/authority tagged suites on Windows. Require a terminal self-clean receipt from P01's non-evidence dependency controller and recheck the implementation tree unchanged. Failure or residue blocks all protected parents；`run-windows-final-scopes` may never launch or absorb this gate.

- [ ] **Step 3: Build/scan the auxiliary artifact and run both Windows wrapper scopes from the final tracked tree**

After Tasks 1–7 and the final relock projection commit are complete, require that exact relock commit to be the clean implementation tree in which every B10/B11 source、three image build-context manifests、helper、custom-tag test and script already exists, is committed and tracked. Invoke only Plan 08's exact installed zero-argument native parent:

```powershell
& 'C:\Program Files\Talenro\C12\talenro-c12-runner-launcher.exe' run-windows-final-scopes
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
```

The Windows native parent launches only Plan 08's exact snapshot `verify-c12.ps1` and `verify-c12.sh` finite portability stages; it never launches `run-c12-integration.ps1`. Those stages receive only closed enum/result stdio and cannot call Docker、Git、scanner、provider、cleanup or signing APIs. For each distinct run the Windows parent drives the complete outer Compose and two-root build/equality flow, while the nested long-lived Linux verifier parent alone performs Plan 07's exact protected-slot sequence `NewCoreSmokeResultParentSession -> direct child FD 3 -> BindStartedChild -> wait/reap -> AdoptAfterChildExit -> AdoptCoreSmokeResultFromProtectedSlot` and signs the resulting `C12VerifierResultV1`. The Windows parent sees only that signed result and can never access the FD3 slot、trusted source or receipt set. After both lifecycles freeze their bounded pre-clean composite, Plan 08 selects success-origin cleanup, exact-cleans and commits the assignment-identical leaf receipt into same-generation `windows_post_cleanup_publication`. Only that successor may durably intent/at-most-once sign/recover the two Windows scopes, bind the fixed sender, consume the channel attestor before `Prepare` and converge the fixed four-member transaction；lost acknowledgement uses only the retained nonce and `InspectSender(nonce)`. Terminal-inactive/next is selected only after exact committed inspection；abort、cleanup/sign failure or active crash cannot sign or bind. There is no per-wrapper command/path/copy and no operator-visible private child、tree-fact handle、result path/digest、catalog/path-list/alternate-producer override.

Expected: two fresh, attested, successful `container_deterministic` scope files with distinct run IDs, byte-identical scan receipts/bundles and the same daemon/repo/tree/spec/toolchain/three-release/authority-operations/record/catalog/process-policy/image inputs.

- [ ] **Step 4: Run the final Linux platform/operator-trust scope**

On the platform runner invoke only Task 5's zero-argument `run-platform-phase-a`, perform the controlled cold reboot, then invoke only `resume-platform-phase-b`. The two parents capture/resume the same helper-owned committed snapshot/root-A layout and private producer facts, validate their own canonical receipt/bundle, exact-clean through the cleaned-tombstone terminal sequence, sign nested/outer evidence, consume the channel attestor before `Prepare`, and commit only their fixed platform transaction. They neither open nor compare the Windows/authority slots；cross-run shared receipt/bundle equality is deferred to completion. No shell or operator sees the receipt/layout/provider/cleanup/signing handles.

Expected: fresh production `linux_platform_operator_trust` evidence with all replay/time/host-policy/attacker/post-start/operator-guard cases and cleanup successful.

- [ ] **Step 5: Run the final authority-fence/PITR scope**

On the authority runner invoke only Task 6B's zero-argument `run-authority-final`. Its native parent owns capture/private production、fixed one-shot authority stages、trusted-time/provider handles and exact cleanup, validates its own canonical receipt/bundle and selects success-origin cleanup only with the package-private final-workload token. On every same-generation restart it first calls the closed high-level authority dispatch and executes exactly the returned cleanup、ordinary-finalization or publication route；each route rechecks the authenticated phase, with no recovery probing/error fallback, and ordinary-finalization returns no session/capability. The receipt-bound two-CAS leaf finalizer enters same-generation `authority_post_cleanup_publication` only after exact absence/key destruction；that successor alone exposes the nested/outer/channel attestors. The high-level facade signs/self-verifies nested and outer evidence, consumes the channel attestor before `Prepare`, converges only its fixed authority transaction and finalizes terminal-inactive/next only after exact committed-sender inspection. Abort、active crash or cleanup failure goes ordinary terminal-inactive and cannot sign. It never opens or compares Windows/platform slots；completion alone checks cross-run equality. No relative script、provider-config、receipt/layout/WAL/run-root/evidence path、attestor、sender、nonce or transaction is operator-visible.

Expected: fresh production `authority_fence_pitr` evidence with the identical scanner-derived tuple/combined/record-set/bundle/catalog/process-policy/image/full-scan bindings, the Task 6A/6B genesis/normal-epoch/recovery/source/staging/Down matrices, real PITR/response-loss refusal, all provider/attestor records terminal or exact-recoverable, destructive restore explicitly blocked at safe staging, and cleanup successful.

- [ ] **Step 6: GREEN — create the four-scope completion manifest**

The locked native consolidation parent authenticates the guard/current cell and obtains only `CompletionPublishPendingHandle` from fresh Begin or pre-E2 pathless recovery. It calls `handle.AdoptExpectedEvidence(ctx)`；inside the composite, P08's private proof/mutation hooks durably record each exact outer+set intent and actual, while the parent can neither name nor choose receiver/binding/set/slot/sequence/nonce/lease. It then calls `BindCompletionPublicationOperator(handle)` and `MaterializeAdoptedEvidence(ctx)`. Before provider/child resources every invocation calls `InspectCompletionInnerCleanupRecoveryDispatch(ctx,handle)` and executes exactly one five-way target. `begin_required` alone enters `BeginCompletionInnerCleanup` and yields validation-required. Selected intent/restart/capsule before any manifest intent maps only to `session_validation_recovery` and yields validation-required through `RecoverCompletionInnerCleanup`. Create `cleanup_capsule` plus set-first pending or outer selected `manifest_intent|manifest_actual` maps only to `session_cleanup_recovery`；the same facade first finishes only that reciprocal state barrier from stored bytes and yields cleanup-only. `finalization_pending` maps only to no-session `finalization_recovery` and `RecoverCompletionInnerFinalization`, including after the key has already been destroyed. Create target-completed `none` maps to actionless `complete` only when authenticated reciprocal `manifest_intent`-or-later progress proves the cleanup target；a pre-first-CAS or genuinely cleaned pre-manifest `none` with no set pending is instead `begin_required`. Only validation-required may call parameter-free `ValidateMaterializedEvidence(ctx,inner)` and `StageValidatedManifest(ctx)` before exact cleanup；cleanup-only must skip both and directly exact-clean through `FinalizeCompletionInnerCleanup(ctx,handle,inner)`, then use only stored-byte `RecoverPublication`. Every action target rechecks all five values and no nil-success/error fallback calls Begin. Any returned session is consumed before outer publication continuation, while successful no-session finalization or `complete` proceeds only after its target/progress reread. The parent then calls only `handle.AcknowledgeExpectedEvidence(ctx)` and `handle.FinalizeRetainedComplete(ctx)`；the latter privately freezes the exact E2 target plus immutable payload/finalization-set audit digest and owns E→E1 Prepare/Recover and proof-bound E1→E2 Commit. No terminal cell exists before Commit. After E2, success/duplicate handling uses only old-reservation clear plus the distinct audit receipt, including after completed revalidations. Every pre-E2 crash resumes through the same opaque handle/operator；partial/mismatched/replayed state cannot be enumerated into success. The capsule envelope and tombstone payload remain distinct authenticated values, with the latter carried only by `finalization_pending`; no credential、provider config、raw child output、path or low-level capability enters the command surface.

For create, selecting `manifest_intent` means the outer role record and adoption set already contain the same bounded canonical bytes and digest before any temp write. Therefore the post-cleanup actionless `complete` branch is recoverable even when the temp is absent or partial：its only next action is `RecoverPublication`, which replays those stored bytes and cannot call the provider、child、validator or canonicalizer again.

An original helper that receives `handle.FinalizeRetainedComplete(ctx)` success may return normally；the terminal audit is only the response-loss/later-duplicate create path after that helper's `exiting` reservation has been reaped or absence-proved and cleared. It never replaces the normal facade return and never yields a revalidation handle.

Run:

```powershell
& 'C:\Program Files\Talenro\C12\talenro-c12-runner-launcher.exe' completion-consolidate-create
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
```

Expected: PASS；the profile-owned manifest is a sibling, never a ninth member of the strict eight-file evidence root. The same-set three adoption actuals、inner cleanup、manifest/`completed` durability and three ACK actuals precede one logical retained-finalization transaction implemented by exactly two guarded advances `E→E1 finalize_ready→E2 retained_complete`；no second Open or old unbound adoption API is reachable, and replay of any earlier authenticated cell yields zero authority. The manifest has exactly four ordered scope digests, equal explicit nested/outer build inputs including every scanner-derived operations/bundle/catalog/full-scan value, the one Plan 08 builder-derived current `SpecDigest`, the complete signed creation-time envelope with its source identity/input digest/nonce/digest/sequence projections, and expiry equal to the exact minimum policy/outer/nested instant；all evidence is unexpired at source-attested current time, no fake is promoted to a production scope, and no safe-staging result is promoted to legacy-upgrade completion.

- [ ] **Step 7: Run two independent whole-suite reviews and close their fail-closed gate**

Freeze the Step 11 implementation commit/Git-tree locator、canonical tracked-tree digest、current three-entry `SpecDigest`, final repository-gate result digest、four ordered scope evidence digests and the Step 6 completion-manifest digest as one immutable review assignment. Dispatch that same assignment to exactly two fresh reviewers who did not author the implementation or either other's review：one fixed role is `quality_correctness` and the other is `security_fail_safe`. Their authenticated reviewer identities and roles must be distinct. Neither reviewer may receive a mutable worktree, signing/provider credential, cleanup/recovery capability, native-runner session, evidence channel, raw secret-bearing output or a path/selector that can change the assignment.

Each reviewer inspects the complete Plan 08/09 implementation and its executable/static/tagged gate coverage against the frozen specs and emits one canonical external `C12IndependentReviewReceiptV1`. Its bounded schema contains only `SchemaVersion`、`ReviewKind`、`ReviewerIdentity`、`ReviewerRole`、implementation commit/Git-tree locator/canonical tracked-tree digest、`SpecDigest`、final-gate/ordered-scope/manifest digests、`CriticalOpen`、`ImportantOpen`、`Outcome`、`ReviewedAt`、a digest of the externally retained detailed review and the reviewer-authenticated signature. The canonical receipt digest uses a fixed `TALENRO-C12-INDEPENDENT-REVIEW-RECEIPT-V1` domain over every preceding field；Git never stores the detailed review、raw test output or credentials.

The fixed external review-gate controller receives only two deployment-provided nonenumerable receipt handles. It strict-decodes/canonicalizes both receipts, verifies their signatures against the independently administered reviewer registry, exact-matches the frozen assignment, requires the two fixed roles and distinct identities, rejects missing/duplicate/extra receipts and requires `CriticalOpen=0`、`ImportantOpen=0` and `Outcome=pass` in both. There is no skip、waiver、severity downgrade、caller-supplied count/path or fallback route. It emits one bounded canonical `C12IndependentReviewGateReceiptV1` containing the assignment digest, the two ordered receipt digests, zero Critical/Important totals, `Outcome=pass` and an authenticated combined digest; it exposes no review content or mutable authority.

Any unresolved Critical/Important finding blocks C1.2. Any finding-driven source/spec/helper/client/policy/manifest/toolchain/release input change invalidates every prior scope、manifest and review receipt and forces the complete Task 7 source-commit → rebuild/relock → projection-commit/provision sequence followed by Task 8 Steps 1–7 again. The two detailed reviews and all three authenticated receipts remain external until the longest evidence/audit retention deadline. Low/Medium observations may remain only when the two reviewers classify them independently, their exact detail digests are retained and neither implies a Critical/Important safety or correctness gap.

- [ ] **Step 8: GREEN — write the digest-only completion record**

`docs/security/c12-completion-record.md` records the clean attested implementation commit/Git-tree locator plus separate canonical tracked-tree digest captured immediately after Task 7's final projection commit—not the later self-referential record commit—plus canonical spec-set/toolchain/three-release/operations tuple/combined/record-set/bundle/catalog/process-policy/image-set/full-scanner/license/document digests, four scope evidence digests, manifest digest, trusted-time source identity and creation-statement digest/sequence, run completion timestamps/minimum expiry, cleanup success and the exact constrained scanner claim. It also records the two fixed review roles、two distinct bounded reviewer identities、both canonical review-receipt digests、both zero Critical/Important counts/statuses and the combined review-gate assignment/receipt digest and `pass` status；it never copies the detailed reviews or treats their absence as zero findings. It states that the operations artifact is an independently scanned auxiliary executable, not a fourth proprietary release, and that the eventual two-file documentation commit is a post-attestation administrative record rather than an attested build. The external canonical bundle、detailed reviews and authenticated review receipts are retained with the evidence package until the longest evidence/audit retention deadline；Git records only their bounded digests/statuses. If v7 staging conformance evidence is cited, record only that it proves `safe_staging_established_and_closed`/`restore_incomplete` and does not prove a completed legacy upgrade or destructive restore. Do not copy attestation, provider credential, cert/key, core asset, config, absolute runner path or raw test output into Git.

- [ ] **Step 9: REFACTOR — run final clean-tree verification**

Run the three no-argument exact semantic gates before any broad package command；each must still exact-list and JSON-run/pass every declared root once with zero root/descendant skip:

```text
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-platform-semantic-gate.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-authority-semantic-gate.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-completion-semantic-gate.ps1
```

Run: `go test ./... -count=1 -timeout 30m`

Run: `go test -race ./... -count=1 -timeout 30m`

Run: `go vet ./...`

Run: `go tool golangci-lint run ./...`

Run: `git diff --check`

Run in PowerShell:

```powershell
& 'C:\Program Files\Talenro\C12\talenro-c12-runner-launcher.exe' completion-revalidate
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
```

Expected: all PASS；the validator resolves and revalidates the implementation commit/tree recorded in the manifest rather than substituting current working bytes；the diff from that implementation tree is exactly the one intended `docs/security/c12-completion-record.md` file and touches no source/spec/manifest/toolchain/scanner/license input；and the manifest remains unexpired. The external review-gate controller revalidates both reviewer signatures、distinct fixed identities/roles、the exact unchanged assignment、both zero Critical/Important counts and the combined receipt after `completion-revalidate`; any mismatch、expiry-invalidating delay or newly changed implementation/artifact blocks the roadmap.

- [ ] **Step 10: GREEN — update the roadmap only after review and final verification**

Change C1.2 from `current design` to `complete`; change C1.3 to `current design`; preserve the C1.2 production target, Redis production block and Xray/sing-box license/release caveats. Do not claim C1.3 implementation. No roadmap byte may change before Steps 7–9 all pass.

- [ ] **Step 11: Commit the completion record and roadmap transition**

```bash
git add docs/security/c12-completion-record.md docs/roadmap/implementation-sequence.md
git commit -m "docs: record C1.2 control plane completion"
```

- [ ] **Step 12: Verify the post-attestation administrative commit**

Run: `git diff --exit-code`

Run: `git status --porcelain=v1 --untracked-files=all`

Run: `git diff --name-only HEAD^ HEAD`

Expected: the first and second commands both produce empty output, including no untracked path；the final name-only command lists exactly `docs/security/c12-completion-record.md` and `docs/roadmap/implementation-sequence.md`；`HEAD^` equals the manifest's attested implementation commit；the recorded external completion manifest still verifies against that parent tree and is not rewritten to claim the documentation commit.

## B11 and C1.2 final exit gate

B11 closes only after Tasks 1–5, 6A, 6B and 7 are committed and Task 8 has produced a valid, fresh four-scope manifest from one build using the current Plan 08 canonical `SpecDigest`；two distinct independent reviewers have issued assignment-identical authenticated quality/correctness and security/fail-safe receipts；the combined review gate has revalidated both with zero unresolved Critical/Important findings；and the final verification has passed before any roadmap edit. The implementation/conformance milestone does not assert that any legacy deployment completed destructive restore: `fresh_v7_staging_closed` remains safe staging with `restore_incomplete`. A blocked provider, unavailable runner, expired evidence, missing/invalid/duplicate review receipt, unresolved Critical/Important finding, failed cleanup, incomplete license record, Redis license bypass, local resource shortage, fake-only result or staging-as-complete claim leaves C1.2 at `current design`; none is a reason to weaken or relabel the gate.
