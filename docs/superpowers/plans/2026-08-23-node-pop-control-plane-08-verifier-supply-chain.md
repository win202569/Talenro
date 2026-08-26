# Talenro C1.2 Verifier and Supply Chain Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付 canonical C1.2 spec-set 的唯一 builder/validator、唯一 canonical nested-Docker verifier、一个受锁 native Windows parent 驱动的 PowerShell 与 repository-external Git Bash 两个独立 portability stage、可恢复 ownership WAL、封闭制品 scanner，以及不能冒充真实 Linux/provider 证据的两个 Windows scope evidence。

**Architecture:** `internal/c12evidence/specset.go` 只从固定 tracked `docs/superpowers/specs/c12-spec-set.v1.json` 及该 tracked tree 中列出的三份 approved Markdown 构造 canonical set，并成为所有 scope/completion evidence 的唯一 `SpecDigest` builder/validator。一个 catalog/toolchain-locked、预构建、长驻的 native Windows parent 独占 tree/session/WAL/nonce/attestor/output/cleanup handles，连续驱动两次 digest-pinned outer Compose；它分别直接启动固定 PowerShell 与 repository-external Git Bash one-shot portability stage，shell只返回有限结果，不能调用Docker/helper/scanner或持有受保护句柄。所有 PostgreSQL、Redis、NATS、Linux fixture、real-core smoke、load 与扫描判定都在 `c12-verifier` 中由另一个长驻 native Linux parent完成，Bash authority script同样只运行固定ordinary stages。内层只产生并签署 strict `C12VerifierResultV1`；native parent将两个独立shell事实各自与本次 runner/daemon/Job/WAL/cleanup组合成两份distinct attested `C12ScopeEvidenceV1`，不共享或转签内层 scope。Outer rootless inner daemon 与 host daemon 隔离；append-only WAL 在资源创建前写 intent、创建后写 actual，scanner 对三个 proprietary release、独立的 `nodecontrol-authority-operations` 辅助 binary+image、runner-helper、内层 integration 依赖布局、其余production/outer OCI 和 verifier test assets 做封闭集合比较。B10只交付该通用scanner/evidence/native-parent能力；B11生成辅助制品后，Plan 09 Task 8才运行四份最终权威scope。

**Tech Stack:** Go 1.26.5、PowerShell 5.1+、Git for Windows Bash 5+、Linux `/usr/bin/bash`、Docker Engine/Compose v2、rootless Docker-in-Docker、mTLS、OCI image layout、`debug/elf`、`debug/pe`、`debug/buildinfo`、RFC 8785 JCS、SHA-256、HMAC-SHA-256、Windows DPAPI。

**Spec:** [Approved C1.2 node and POP control-plane base design](../specs/2026-08-23-node-pop-control-plane-design.md), approved SHA-256 `B0B0DDBBD11546FC07A25CB76992E375481192A3261270FF442095501CC2B4B5`, especially sections 17.5–17.6, 18, 20, and 21 batch 10; [approved authority Abort/serving amendment](../specs/2026-08-24-nodecontrol-authority-abort-serving-design.md), approved content SHA-256 `86996084462A5DE1E7667D56A135E93099EE38E1089464CCDFCFF7284EA0D1D7`, especially §9; [approved Authority Protocol v7 additive upgrade amendment](../specs/2026-08-24-nodecontrol-authority-v7-upgrade-design.md), approved content SHA-256 `EFAEBE52BDC3D70BDA8737C893B02752E60ACEC0A08441813CADF1425079FC8E`, especially §10 B10 and §11; canonical set manifest [c12-spec-set.v1.json](../specs/c12-spec-set.v1.json).

## Global Constraints

- 本册只实现 `C1.2-B10`，并消费已独立通过的 B08 Xray 与 B09 sing-box result；缺任一结果时 verifier fail closed。
- canonical spec set 的唯一输入路径固定为 tracked `docs/superpowers/specs/c12-spec-set.v1.json`，当前 exact membership 恰为 base `B0B0DDBBD11546FC07A25CB76992E375481192A3261270FF442095501CC2B4B5`、Abort/serving amendment `86996084462A5DE1E7667D56A135E93099EE38E1089464CCDFCFF7284EA0D1D7` 和 Authority Protocol v7 amendment `EFAEBE52BDC3D70BDA8737C893B02752E60ACEC0A08441813CADF1425079FC8E`；manifest 不含自身 hash。
- `internal/c12evidence/specset.go` 独占 strict decode、repository-relative path confinement、tracked regular-file read、Markdown approved-state check、SHA-256/JCS 构造与 `SpecDigest` 验证。其他 package、scanner command、B10/B11 wrapper 只能调用这个 helper，禁止第二套 member list、文件 hash、JCS 或 digest 计算。
- `SpecDigest` 固定为 `SHA-256(ASCII("talenro.c12.spec-set.v1") || 0x00 || JCS(C12SpecSetV1))`。base 唯一，amendments 按 repository-relative forward-slash path bytewise 升序且无重复，path 为 canonical UTF-8，hash 为 lowercase 64-hex；Authority Protocol v7 amendment 入 set 后只能由该 builder 计算新的单一 `SpecDigest`，旧 evidence 固定失效。
- Builder 必须从 native helper 已捕获并 repository-external materialize 的 exact Git tree 读取 manifest 与三份 Markdown，并证明 `HEAD` commit/tree、index tree、执行/制品源闭集的 tracked mode/bytes 及 relevant-untracked-zero 条件都与该 snapshot 相同；manifest path/member/hash、tracked mode/tree 或第二 amendment 任一 tamper 都在 Docker/VM/任何资源创建及任何 scope evidence 前失败。production API/CLI 不接受 manifest path、member、hash、Git tree OID、canonical bytes 或 caller-supplied `TrackedTreeDigest`/`SpecDigest` override。
- B10 只生成两个 container-deterministic scope evidence 及 scanner/cleanup receipts；不得生成或签发 upgrade、legacy-source、epoch-recovery、fresh-staging、destructive-restore 或 Down evidence，也不得直接调用 commit archive/Fence semantic-ID journal。
- Windows scope 恰好为 `windows_powershell_docker` 与 `windows_git_bash_docker`；Git for Windows Bash 不能称为 Linux platform scope。
- Nested Docker provider 名称必须精确为 `DeterministicRollbackGuardFakeV1`、`DeterministicSecurityLatchGuardFakeV1`、`DeterministicTrustedTimeFakeV1`、`DeterministicHostMemoryIsolationPolicyFakeV1` 与 `DeterministicControlPlaneAuthorityFenceFakeV1`。
- Container fake evidence 的 provider class 固定为 `container_deterministic`；completion validator 必须拒绝它满足 `linux_platform_operator_trust` 或 `authority_fence_pitr`。
- Wrapper 先解析并验证 PowerShell/Git Bash/docker/Compose 的 canonical absolute path、SHA-256、签名/版本与 OS/arch；禁止 PATH alias、shim 和自动升级。
- 两个 wrapper 显式使用同一个 locked Docker context/endpoint；`DOCKER_HOST`、`DOCKER_CONTEXT`、`COMPOSE_FILE`、`COMPOSE_PROJECT_NAME` 必须为空或逐字等于 lock。
- Outer Compose 恰好三个 service：`c12-context-init`、`c12-inner-daemon`、`c12-verifier`；恰好一个 internal network 和三个 named volume：`inner-data`、`context-secrets`、`verifier-output`。
- 三个 outer image 必须以 OCI digest 固定并声明 `pull_policy: never`；wrapper 使用 `--pull never --no-build`，禁止 fallback pull/build/tag replacement。
- Host Docker allowlist 仅包含 locked context inspect/version/info、三个 exact image inspect、offline compose config、exact project up/stop/rm/down `--volumes`、对三个 exact project/service 分别执行 `compose ps --all --quiet c12-verifier|c12-context-init|c12-inner-daemon`、WAL exact container-ID inspect、以及仅对 authenticated WAL 已记录 deterministic name 的 one-result `docker network inspect talenro-c12-<run-id>-network` 和三次 `docker volume inspect talenro-c12-<run-id>-{inner-data,context-secrets,verifier-output}`（包括 crash recovery 与 down 后 not-found 验证）。三个 `ps` 各必须只返回一个 stopped container ID，且 inspect 的 project/service/run labels、image digest、creation time、exit state/code 必须匹配；network/volume inspect 必须恰好一个结果并匹配 name、Compose project/run labels、driver/options/scope、creation token，network还匹配 exact ID/internal=true，volume还匹配各自用途。所有 intent 在 `up` 前 fsync；successful `up` 返回后立即把三 container ID、network ID和三 volume identity digest写 actual，crash recovery只可从 authenticated intent 的 exact stable name补 actual。禁止默认 running-only `compose ps`、`ps -q`、按 container name 查询、Docker list或模糊枚举。四个容器内源只能是 verifier 的 `/run/c12-export/closed-set-receipt.json`、`/run/c12-export/supply-chain-evidence-bundle.json`、`/run/c12-export/verifier-result.json` 和 context-init 的 `/run/c12-public/verifier-result-trust.json`；禁止从 named volume/container path 枚举或拷贝。除这四次 exact-ID/fixed-file copy 外，禁止 pull/build/load/save/run/exec/cp/logs/events/system prune、list 与模糊枚举。
- Verifier 不挂 host Docker socket，只通过 per-run mTLS 连接 inner daemon；repo 只读、output volume 可写，inner daemon 独占空 data volume。
- Core/base OCI layouts 只存在 verifier image 的 test-assets layer；PostgreSQL 18.4、Redis 8.8.1 与 NATS 2.14.3 也必须作为 exact digest-addressed integration-only OCI layouts 进入封闭 catalog、SBOM/provenance/vulnerability/license/notice/source 记录与 `ImageSetDigest`。Redis 的 integration-only 测试使用不改变 production license block。inner daemon 无 registry 网络，只允许 streaming `ImageLoad`，`ImagePull` 必须被策略和测试拒绝。
- Real-core smoke只能通过 Plan 07-owned `c12-core-smoke-runner` 的同run两阶段协议消费：test child仅从固定 FD 9 `LiveReceiptSource` 得到 opaque provisional handle；core-smoke child在完整cleanup并证明11个WAL resources全部absent后才通过同一 `RunnerReceiptFinalizer` 生成 `ValidatedCoreSmokeReceiptSetV1`。P08 locked Linux verifier parent对每个adapter/run必须执行唯一序列 `NewCoreSmokeResultParentSession` → direct-child inherited literal FD 3 → `BindStartedChild` → wait/reap exact child → `AdoptAfterChildExit` → `AdoptCoreSmokeResultFromProtectedSlot`；只有最后一步在parent内mint `TrustedCoreSmokeReceiptSource`并再次调用唯一semantic validator。该完整 set 必须含 same-committed-tree helper-authenticated `CoreSmokeBuildReceiptV1` 及其 digest，逐项绑定 recipe-only runner manifest、tracked tree、helper identity和 broker/test/test2json/test-host 实际制品 hash；manifest 本身不得冻结 final output hash。Windows outer parent永不接触FD3 slot、adoption、trusted source或receipt set，只消费已由nested Linux parent签名的 `C12VerifierResultV1`。bytes、path、digest、FD scalar、reopened FD、socket/pipe/`SCM_RIGHTS`、shell或另一“private channel”都不能替代该序列。C12 verifier/helper必须重验这些绑定，并把完整 set canonical digest、独立 build-receipt digest及final cleanup digest绑定进 `VerifierTestResultDigest`/最终 `TestResultDigest`；旧 full set、manifest内final hash、旧result/digest path、provisional handle、仅child success或另一run的clean set都不能替代。
- 每次入口生成 128-bit random run ID，project 固定为 `talenro-c12-<32 lowercase hex>`，临时根必须重新解析为 OS temp 的直接子目录。
- WAL 是 append-only、hash-chained、HMAC-authenticated；intent 与 actual 都必须 file+directory fsync。HMAC key 使用 DPAPI 或 mode `0600` 文件保护且不写日志。
- B10/B11 relock 前只认证 candidate native build identity、tracked policy、toolchain native-role projection（B11另含final catalog exact cover），不得预开bootstrap/session或要求install receipt。B10 candidate handle不依赖尚不存在的production catalog；B11 handle另绑定final catalog。五/九个输出全部durable后才逐role `InstallFixedNativeRunnerProfile`，该调用线性消费installer session并在每个返回前销毁session+candidate handle，随后只用精确的 `VerifyFixedNativeRunnerInstallation(context.Context, NativeRunnerRole) error` 做不消费launch generation的nil/error审计；只有实际production parent可调用`OpenInstalledNativeRunnerSession`。每个role独立保持monotonic launch generation与at-most-one active run，同role active并发拒绝，terminal cleaned后下一generation无需重装即可开放，跨role互不消费。每个role另有一个不可重置、不可复用的machine-provisioned rollback-resistant monotonic state guard；它只锚定`StateEpoch + current role-state digest`，不保存phase/run/launch generation/next/publication/payload，也不能成为第二份业务真相。
- Completion receiver在任何slot mutation前必须用P09 durable `completion_publish_pending`绑定`BeginAdoptionSet`并fsync exact-three intent；每次adopt都按intent fsync→idempotent slot mutation→nonce/envelope/object actual fsync后才返回。`RecoverAdoptionSet`只恢复同run/generation/set，ACK只在同一binding的completed marker durable后发生；`RecoverAcknowledge(set)`只读唯一pending ACK intent+matching adoption actual且不接收selector。禁止枚举、cross-bind或caller构造set/nonce/digest。
- Recovery 只查询 WAL 中 exact deterministic name/actual ID，禁止 label/name list；pre-existing collision、creation time/label/image mismatch 都停止并要求人工隔离。
- Cleanup deadline 固定 120 秒；exit bit 1 表示主测试失败，bit 2 表示 cleanup/ownership failure，两者可以同时保留。
- wrapper 不直接调用 PATH 中的 Git，也不把 ambient worktree 挂入 verifier。受锁 native helper 独占 absolute Git executable identity/hash/version、`HEAD`/index/tree-object 读取和 snapshot materialization：资源创建前要求 `HEAD` commit 的 Git tree OID 等于 index `write-tree` OID、执行/构建/catalog/spec/toolchain/license 闭集逐 mode/byte 匹配且相关 untracked 为零，计算 canonical 64-hex `TrackedTreeDigest`，把该 exact tree materialize 到 WAL-owned repository-external read-only snapshot；Compose 只挂该 snapshot。结束时 helper 重新验证原 repo 的 HEAD/index/source closure 和 untracked-zero，并 exact-clean snapshot。仅保存并比较 `git status` baseline 不足以签证，来自同一脏 worktree 的“前后相同”必须拒绝。
- Scanner version 固定 `talenro-artifact-scan/v1`；ruleset digest写入 toolchain lock。缺命令、不可解析输出、输入集合不全或 scanner/ruleset漂移都失败。
- 同源 `cmd/talenro-c12-runner-helper` 必须由受锁 toolchain 生成两个 distinct catalog role/build digest：Windows wrappers 只能调用 `c12_runner_helper_windows_amd64` PE，verifier image 只能调用静态 `c12_runner_helper_linux_amd64` ELF。两者都进 toolchain/catalog/image/full-scan exact-cover，并在运行前验证 OS/arch/build digest；禁止 `go run`、环境构建或用其中一份跨 OS 执行。
- 三个 proprietary release 固定 `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 -buildmode=exe`，禁止 `-linkshared`、external linker 和自定义 extldflags；Linux release不得有 `PT_INTERP`、`DT_NEEDED` 或动态 import closure。
- Scanner 不能用自由文本 `strings` 作为结论；必须解析 Go deps/embed、PE/ELF、OCI manifest/config/layer 与 archive payload。
- Xray/sing-box test asset、三个 outer image以及独立 authority-operations auxiliary binary/image各自的SBOM/provenance/vulnerability/license/notice/source obligations必须完整入账；每条 operations 记录必须绑定本次 exact binary/image/build/dependency subject，辅助分类不产生许可证豁免，Redis production block保持不变。
- 生产 runner 的唯一 deterministic 制品入口是受锁、预构建且 OS/arch/build-digest 已验证的 native helper parent mode：committed tree 使用 `talenro-c12-runner-helper capture-and-build-closed-set`，Task 6A staged diagnostic 使用独立的 `capture-staged-and-build-closed-set`，Windows nested verifier 使用 `validate-capsule-and-build-closed-set`。parent在一个受锁session中捕获/验证tree、创建并WAL记录两个roots，再用显式 Windows handle inheritance list或Unix FD duplication直接spawn同一份预构建helper的private `build-closed-set` child；operator/shell从不看见或传入 tree-facts HANDLE/FD。child调用 `internal/artifactscan`，必须同时接收两个不同、run-scoped、repository-external、空且独占的 build root，用受锁 recipe 构建两次，对两根各调用同一 `ScanClosedSet`，要求制品及 canonical receipt/bundle 逐字节相等后才原子发布一对输出。它只从唯一 tracked `testdata/c12/artifact-catalog.v1.json` 解析完整输入角色，并对受跟踪构建/OCI/release/license 声明与两个制品根分别做双向 exact-cover；缺、余、重复、role swap、symlink/reparse 或 caller path-list/catalog/digest override 都失败。`talenro-artifact-scan scan-closed-set` 保留为单根 fixture/diagnostic CLI，wrapper/runbook 不得用它绕过 native helper 的双根构建比较；任何 `go run`、Go tool 临时 executable、shell或另一未锁进程都不得接收/转交 tree-facts HANDLE/FD。
- 每个 scope evidence 最长 72 小时并绑定 repo/tracked tree/spec/toolchain、三个 release、完整 authority-operations artifact tuple、其组合摘要、supply-chain record-set及可审计 bundle digest、artifact catalog、process policy、full scan、全部 image、test result 与 cleanup digest；辅助制品不是第四 release，但缺失或跨构建不一致同样失败。
- 每个 task 使用独立 RED/GREEN/REFACTOR 与 commit；只 stage 明列文件，绝不 stage 用户未跟踪目录。

---

## File structure and frozen B10 interfaces

```text
internal/c12evidence/
├── specset.go                       # sole tracked C12SpecSetV1 builder/validator
├── specset_test.go
├── scope.go                         # scope/provider class/evidence validation
├── scope_test.go
├── canonical.go                     # strict JSON + JCS digest
├── wal.go                           # intent/actual/cleaned append-only WAL
├── wal_test.go
├── cleanup.go                       # exact cleanup plan and exit bits
├── cleanup_test.go
├── runnerhelper.go                  # sole Go bridge for WAL/canonical result-to-scope/attestor operations
├── runnerhelper_test.go
└── wrapper_contract_test.go         # fake Docker and two-shell command contracts
internal/c12runnerprofile/
├── profile.go                       # neutral four-final-plus-diagnostic tracked policy/JCS owner
├── profile_test.go
├── bootstrap.go                     # fixed-location sealed machine bootstrap and install receipt
└── bootstrap_test.go
internal/artifactscan/
├── rules.go                         # talenro-artifact-scan/v1 closed rules
├── buildinputs.go                   # fixed outer-context and operations-policy loaders
├── buildinputs_test.go
├── godeps.go                        # go list/buildinfo/embed inventory
├── executable.go                    # ELF/PE/static-link inventory
├── oci.go                           # OCI manifest/config/layer traversal
├── scan.go                          # closed-set comparison and receipt
└── *_test.go
cmd/talenro-artifact-scan/main.go
cmd/talenro-c12-runner-helper/{main.go,main_test.go}
deploy/c12/
├── compose.outer.yaml
├── context-init.build-context.v1.json
├── inner-daemon.build-context.v1.json
├── verifier.build-context.v1.json
├── context-init.Dockerfile
├── inner-daemon.Dockerfile
├── verifier.Dockerfile
├── context-init-entrypoint.sh
├── inner-daemon-entrypoint.sh
└── verifier-entrypoint.sh
scripts/
├── c12-authority.sh
├── verify-c12.ps1
└── verify-c12.sh
testdata/c12/
├── artifact-catalog.v1.json              # sole tracked final exact-role catalog; created by B11 Task 6A
├── integration-manifest.v1.json          # closed PG/Redis/NATS inner groups and exact OCI roles
├── toolchain-lock.v1.json
├── windows/runner-profile.v1.json        # B10-owned policy-only windows_scopes profile
├── diagnostic/runner-profile.v1.json     # B10-owned nonpublishing staged_diagnostic profile
├── authority/operations-build-policy.v1.json # schema/loader owned here; final instance created by B11 Task 6A
├── evidence/container-scope-valid.json
└── scanner/*.json
```

The evidence types introduced here are frozen for plan 09:

```go
package c12evidence

type C12SpecSetMemberV1 struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type C12SpecSetV1 struct {
	SchemaVersion string                 `json:"schema_version"`
	Base          C12SpecSetMemberV1     `json:"base"`
	Amendments    []C12SpecSetMemberV1   `json:"amendments"`
}

type GitTreeLocatorV1 struct {
	ObjectFormat string // "sha1" or "sha256"
	CommitOID    string // required for authoritative evidence; empty only for the Task 6A staged diagnostic candidate
	TreeOID      string // exact Git object ID, width determined by ObjectFormat
}

func CaptureCommittedTrackedTree(context.Context, string, string) (GitTreeLocatorV1, contracts.Digest, error)
func CaptureStagedTrackedTree(context.Context, string) (GitTreeLocatorV1, contracts.Digest, error)
func BuildCanonicalTrackedTreeDigest(context.Context, string, GitTreeLocatorV1) (contracts.Digest, error)
func BuildTrackedC12SpecSet(context.Context, string, GitTreeLocatorV1) (C12SpecSetV1, contracts.Digest, error)
func ValidateTrackedC12SpecDigest(context.Context, string, GitTreeLocatorV1, contracts.Digest) error

type Scope string

const (
	ScopeWindowsPowerShellDocker Scope = "windows_powershell_docker"
	ScopeWindowsGitBashDocker    Scope = "windows_git_bash_docker"
	ScopeLinuxPlatformTrust      Scope = "linux_platform_operator_trust"
	ScopeAuthorityFencePITR      Scope = "authority_fence_pitr"
)

type ProviderClass string

const (
	ProviderContainerDeterministic ProviderClass = "container_deterministic"
	ProviderProduction             ProviderClass = "production"
)

type AuthorityOperationsArtifactTupleV1 struct {
	BinarySHA256            contracts.Digest
	ImageManifestDigest     contracts.Digest
	BuildReceiptDigest      contracts.Digest
	DependencyClosureDigest contracts.Digest
}

type C12ScopeEvidenceV1 struct {
	SchemaVersion      string
	Scope              Scope
	ProviderClass      ProviderClass
	RunID              uuid.UUID
	StartedAt          time.Time
	FinishedAt         time.Time
	ExpiresAt          time.Time
	RunnerIdentity     contracts.Digest
	DaemonIdentity     contracts.Digest
	RepoCommit         string
	TrackedTreeDigest  contracts.Digest
	SpecDigest         contracts.Digest
	ToolchainDigest    contracts.Digest
	ReleaseDigests     [3]contracts.Digest
	AuthorityOperationsArtifact AuthorityOperationsArtifactTupleV1
	AuthorityOperationsBinaryAndImageDigest contracts.Digest
	AuthorityOperationsSupplyChainRecordSetDigest contracts.Digest
	SupplyChainEvidenceBundleDigest contracts.Digest
	ArtifactCatalogDigest contracts.Digest
	ExpectedProcessImagePolicyDigest contracts.Digest
	ArtifactScanDigest  contracts.Digest
	ImageSetDigest     contracts.Digest
	TestResultDigest   contracts.Digest
	CleanupDigest      contracts.Digest
	Attestation        []byte
}

type C12VerifierResultV1 struct {
	SchemaVersion      string
	RunID              uuid.UUID
	StartedAt          time.Time
	FinishedAt         time.Time
	RepoCommit         string
	TrackedTreeDigest  contracts.Digest
	SpecDigest         contracts.Digest
	ToolchainDigest    contracts.Digest
	ReleaseDigests     [3]contracts.Digest
	AuthorityOperationsArtifact AuthorityOperationsArtifactTupleV1
	AuthorityOperationsBinaryAndImageDigest contracts.Digest
	AuthorityOperationsSupplyChainRecordSetDigest contracts.Digest
	SupplyChainEvidenceBundleDigest contracts.Digest
	ArtifactCatalogDigest contracts.Digest
	ExpectedProcessImagePolicyDigest contracts.Digest
	ImageSetDigest      contracts.Digest
	ArtifactScanDigest  contracts.Digest
	ValidatedCoreSmokeReceiptSetDigest contracts.Digest
	CoreSmokeBuildReceiptDigest contracts.Digest
	CoreSmokeFinalCleanupDigest contracts.Digest
	VerifierTestResultDigest contracts.Digest
	InnerProcessReceiptSetDigest contracts.Digest
	VerifierAttestation []byte
}

type C12VerifierTrustHandoffV1 struct {
	SchemaVersion     string           `json:"schema_version"`
	RunID             uuid.UUID        `json:"run_id"`
	IssuerSPKIDigest  contracts.Digest `json:"issuer_spki_digest"`
	LeafProfileDigest contracts.Digest `json:"leaf_profile_digest"`
	NotBefore         time.Time        `json:"not_before"`
	NotAfter          time.Time        `json:"not_after"`
}

// TrustPolicy is the complete caller-independent expected projection for one
// scope.  ValidateVerifierResult consumes only its shared scanner/tree subset;
// ValidateScopeEvidence requires every member below.
type TrustPolicy struct {
	ExpectedScope              Scope
	ExpectedProviderClass      ProviderClass
	ExpectedRunID              uuid.UUID
	ExpectedRunnerIdentity     contracts.Digest
	ExpectedDaemonIdentity     contracts.Digest
	ExpectedRepoCommit         string
	ExpectedTrackedTreeDigest  contracts.Digest
	ExpectedSpecDigest         contracts.Digest
	ExpectedToolchainDigest    contracts.Digest
	ExpectedReleaseDigests     [3]contracts.Digest
	ExpectedAuthorityOperationsArtifact AuthorityOperationsArtifactTupleV1
	ExpectedAuthorityOperationsBinaryAndImageDigest contracts.Digest
	ExpectedAuthorityOperationsSupplyChainRecordSetDigest contracts.Digest
	ExpectedSupplyChainEvidenceBundleDigest contracts.Digest
	ExpectedArtifactCatalogDigest contracts.Digest
	ExpectedProcessImagePolicyDigest contracts.Digest
	ExpectedImageSetDigest      contracts.Digest
	ExpectedArtifactScanDigest  contracts.Digest
	ExpectedTestResultDigest    contracts.Digest
	ExpectedCleanupDigest       contracts.Digest
}

type VerifierResultPolicy struct {
	ExpectedRunID                    uuid.UUID
	ExpectedContextInitContainerID   string
	ExpectedContextInitImageDigest   contracts.Digest
	ExpectedTrustHandoffDigest       contracts.Digest
	ExpectedRunIssuerSPKIDigest      contracts.Digest
	ExpectedResultLeafProfileDigest  contracts.Digest
	ExpectedRepoCommit               string
	ExpectedTrackedTreeDigest        contracts.Digest
	ExpectedSpecDigest               contracts.Digest
	ExpectedToolchainDigest          contracts.Digest
	ExpectedValidatedCoreSmokeReceiptSetDigest contracts.Digest
	ExpectedCoreSmokeBuildReceiptDigest contracts.Digest
	ExpectedCoreSmokeFinalCleanupDigest contracts.Digest
	ExpectedScopeFacts               TrustPolicy
}

func ValidateScopeEvidence(C12ScopeEvidenceV1, TrustPolicy, time.Time) error
func CanonicalScopeEvidenceUnsignedPayload(C12ScopeEvidenceV1) ([]byte, error)
func CanonicalScopeEvidence(C12ScopeEvidenceV1) ([]byte, contracts.Digest, error)
func ValidateVerifierResult(C12VerifierResultV1, VerifierResultPolicy) error
func CanonicalVerifierResultUnsignedPayload(C12VerifierResultV1) ([]byte, error)
func CanonicalVerifierResult(C12VerifierResultV1) ([]byte, contracts.Digest, error)
func ValidateVerifierTrustHandoff(C12VerifierTrustHandoffV1, uuid.UUID, time.Time) error
func CanonicalVerifierTrustHandoff(C12VerifierTrustHandoffV1) ([]byte, contracts.Digest, error)
```

`GitTreeLocatorV1.TreeOID` is a locator into the canonical repository object database, not an evidence digest. In the repository's current `sha1` object format it is exactly 40 lowercase hex；in a `sha256` object-format repository it is 64 lowercase hex. `TrackedTreeDigest` is always a distinct 32-byte SHA-256 value computed as `SHA-256(ASCII("TALENRO-C12-TRACKED-TREE-V1") || 0x00 || JCS(sorted({path,git_mode,blob_sha256})))` over the recursively complete Git tree, with canonical UTF-8 forward-slash paths sorted bytewise, the exact Git mode, and SHA-256 of each blob's bytes；it is never the SHA-1 tree OID padded/re-encoded or a hash of textual `git status`. The capture APIs use only the locked absolute Git executable, verify every object before materialization, and return the locator and canonical digest as separate typed values. Authoritative evidence requires a nonempty commit OID whose resolved tree equals `TreeOID`; only Task 6A's explicitly deferred staged-candidate diagnostic may use `CaptureStagedTrackedTree` with no commit.

`CanonicalScopeEvidenceUnsignedPayload` returns RFC 8785 JCS for one flat closed projection containing **exactly every serialized `C12ScopeEvidenceV1` member except `Attestation`**, with the same field names/types and with no `omitempty`、default injection、nested wrapper or caller-selected exclusion. The runner signs exactly `TALENRO-C12-SCOPE-EVIDENCE-V1\x00 || CanonicalScopeEvidenceUnsignedPayload(evidence)`. `CanonicalScopeEvidence` is never a signing input：after strict decoding and signature verification it returns JCS for the complete flat envelope including `Attestation`, and its digest is `SHA-256("TALENRO-C12-SCOPE-EVIDENCE-ENVELOPE-V1" || 0x00 || canonical_envelope)`.

`C12VerifierResultV1` 是内层唯一可签输出，不含 `Scope`、`ProviderClass`、runner/daemon/Job/WAL/cleanup 或外层 `Attestation`。它的三个 core-smoke fields只可由same-run完整 `ValidatedCoreSmokeReceiptSetV1` validator结果、其中严格重验的 `CoreSmokeBuildReceiptV1` 和parent final cleanup/11-resource-absence证明填充，并逐字等于 `VerifierResultPolicy`。Build receipt必须绑定recipe-only manifest、同一 committed tree/helper和 broker/test/test2json/test-host actual hashes；旧 full set、manifest内final output hash、旧digest path、provisional handle或caller scalar拒绝。`VerifierTestResultDigest` 的domain-separated preimage必须把这三个 digest作为不同有序成员，不能只保留aggregate set digest。`CanonicalVerifierResultUnsignedPayload` 只排除 `VerifierAttestation`，并由 context-init 为该 run 创建的 verifier-result leaf 对 `TALENRO-C12-VERIFIER-RESULT-V1\x00 || payload` 签名。`VerifierAttestation` 是 strict bounded 结构，只包含签名与公开 DER leaf/issuer chain；它不得自带或选择 trust anchor。Context-init 每 run 从 OS CSPRNG 随机生成独立 issuer 与 result-leaf keypair，issuer private key 只存在 context-init 的私有一次性目录/locked memory，签发本次 leaf 后在 readiness 前 close、zero、no-follow unlink 并 file+directory fsync；任何长期/确定性 issuer private key 都不得嵌入 image、lock、repo 或共享 volume。Context-init 将 strict bounded `C12VerifierTrustHandoffV1{run_id,issuer_spki_digest,leaf_profile_digest,not_before,not_after}` 原子写到其容器可写层的固定非挂载 `/run/c12-public/verifier-result-trust.json`。Wrapper 必须先用 exact-project `compose ps --all --quiet c12-context-init` 获得并 WAL-bind 唯一 stopped context-init container ID，inspect 其 run label/image/creation/exit，再以第四次 exact-ID/fixed-source `docker cp` 取得该 public handoff；对它执行 no-follow/one-link/regular/type/32 KiB cap/canonical digest 验证后，用 exact container/image/run/WAL 事实和 handoff 值构造上述 `VerifierResultPolicy`。Leaf 必须带 exact custom EKU、SAN `urn:talenro:c12:verifier-result:<run-id>`、handoff profile digest 与不超过 run deadline 的 validity。仅当 result public chain 终止于 handoff 的 per-run issuer SPKI，leaf profile/RunID/time 和 exact context-init container/image 全部匹配时才验证 result；caller/self-signed/result内自选 root、handoff/result chain splice 必须拒绝。完整 envelope digest 固定为 `SHA-256("TALENRO-C12-VERIFIER-RESULT-ENVELOPE-V1" || 0x00 || canonical_envelope)`。PowerShell/Git Bash 必须各自验签与重验所有 scanner facts，然后通过 native helper 将该 envelope digest 与本次外层事实 domain-bind 到 `C12ScopeEvidenceV1.TestResultDigest`；内层 result 不是 scope，也不能被两个 wrapper 共享、转签或直接当作 scope evidence。每个 wrapper 的 RunnerAttestor 只能签自己的 exact scope enum/run ID 与 `TALENRO-C12-SCOPE-EVIDENCE-V1\x00 || CanonicalScopeEvidenceUnsignedPayload`，不得签完整 envelope、对方 scope 或复用另一 wrapper 的 attestation。

The handoff schema is exactly `talenro-c12-verifier-trust-handoff/v1`, and Task 1 is the sole schema、strict decoder、validator、JCS encoder and digest owner. Strict decoding rejects unknown/duplicate/null/defaulted fields; digests are nonzero fixed-width values; times are canonical UTC with `NotBefore <= current trusted wrapper time <= NotAfter <= NotBefore+6h`. `CanonicalVerifierTrustHandoff` JCS-encodes exactly the six typed members in the frozen JSON names and derives `SHA-256(ASCII("TALENRO-C12-VERIFIER-TRUST-HANDOFF-V1") || 0x00 || JCS(handoff))`. Context-init only produces this type through the Task 1 API/test vectors；Task 4's `create-verifier-trust` is only a native adapter to that API, and Task 3/6/7 helper/wrappers only consume that same validator/digest and may not parse、serialize or hash a parallel shell schema. The handoff is public trust metadata, not a secret or a trust anchor selected by the result；its ownership is the exact context-init container/run/WAL record, and it is copied, consumed and finally removed only as part of that run.

### Task 1: Canonical spec-set plus scope evidence boundary

**Files:**
- Create: `internal/c12evidence/specset.go`
- Create: `internal/c12evidence/specset_test.go`
- Create: `internal/c12evidence/scope.go`
- Create: `internal/c12evidence/scope_test.go`
- Create: `internal/c12evidence/canonical.go`
- Create: `testdata/c12/evidence/container-scope-valid.json`
- Test: `internal/c12evidence/scope_test.go`

**Interfaces:**
- Consumes: `contracts.Digest`, strict JSON/JCS support from C1.1, helper-captured `GitTreeLocatorV1` plus separately computed canonical `TrackedTreeDigest`, fixed tracked `docs/superpowers/specs/c12-spec-set.v1.json`, the three exact approved Markdown hashes frozen above, and `TrustPolicy.ExpectedAuthorityOperationsArtifact`、`ExpectedAuthorityOperationsBinaryAndImageDigest`、`ExpectedAuthorityOperationsSupplyChainRecordSetDigest`、`ExpectedSupplyChainEvidenceBundleDigest`、`ExpectedArtifactCatalogDigest`、`ExpectedProcessImagePolicyDigest`、`ExpectedImageSetDigest`、`ExpectedArtifactScanDigest` supplied by a higher-level command after canonical closed-set receipt/bundle validation. Task 1 does not import `internal/artifactscan` or parse either artifact.
- Produces: the sole tracked-tree capture/digest and `BuildTrackedC12SpecSet`/`ValidateTrackedC12SpecDigest` APIs, sole `C12VerifierTrustHandoffV1` schema/strict validator/canonical digest, strict `C12VerifierResultV1`/`VerifierResultPolicy` validation/canonicalization, then `Scope`, `ProviderClass`, `C12ScopeEvidenceV1`, `TrustPolicy`, `ValidateScopeEvidence`, `CanonicalScopeEvidenceUnsignedPayload`, and `CanonicalScopeEvidence` exactly as frozen above. No function here emits an upgrade, recovery, staging, source-retirement or Down object；`C12VerifierResultV1` 不得带 scope/provider/outer-cleanup 字段。

- [ ] **Step 1: RED — write tracked spec-set membership and tamper tests**

```go
func TestTrackedSpecSetHasTheThreeApprovedMembers(t *testing.T) {
	t.Parallel()
	locator, trackedDigest := capturedCommittedTree(t)
	set, digest, err := BuildTrackedC12SpecSet(t.Context(), repoRoot(t), locator)
	if err != nil || digest == (contracts.Digest{}) {
		t.Fatalf("build tracked spec set: %v", err)
	}
	if got, err := BuildCanonicalTrackedTreeDigest(t.Context(), repoRoot(t), locator); err != nil || subtle.ConstantTimeCompare(got[:], trackedDigest[:]) != 1 {
		t.Fatalf("canonical tracked tree mismatch: %v", err)
	}
	assertExactApprovedC12Members(t, set)
}

func TestTrackedSpecSetRejectsSecondAmendmentTamper(t *testing.T) {
	t.Parallel()
	repo := cloneTrackedSpecFixture(t)
	repo.ReplaceTrackedBlob("docs/superpowers/specs/2026-08-24-nodecontrol-authority-v7-upgrade-design.md", []byte("tampered"))
	if _, _, err := BuildTrackedC12SpecSet(t.Context(), repo.Root(), repo.OriginalTreeLocator()); !errors.Is(err, ErrTrackedTreeMismatch) {
		t.Fatal("builder accepted a tampered second amendment")
	}
}
```

Add table-driven signature-projection RED matrices for both envelopes. Mutating `Attestation` alone must leave `CanonicalScopeEvidenceUnsignedPayload` byte-identical while changing the complete canonical envelope and failing signature verification；mutating any other scope field must change the unsigned payload and fail the old signature. Likewise only `VerifierAttestation` is excluded from `CanonicalVerifierResultUnsignedPayload`；unknown/missing/extra field, a result containing scope/provider/cleanup, a full-envelope signature, any second omitted member, nested payload or caller-selected projection must fail. Add shared handoff vectors consumed by both shell fixtures: wrong schema/RunID, zero issuer/profile digest, unknown/duplicate/null field, omitted member, non-UTC/reversed/>6h time, expired handoff, self-selected issuer and handoff/result chain splice all fail the one Task 1 validator/digest.

Table-drive wrong schema, wrong/missing/extra base, wrong/missing/extra/duplicate amendment, amendment order swap, backslash/absolute/`..`/non-NFC path, uppercase/short/nonhex hash, symlink/reparse/non-regular file, untracked manifest, index-only replacement, working-tree-only replacement, member hash drift, approved-state removal, tracked-tree digest splice and second-amendment byte/hash tamper. Also freeze a separate tree-identity matrix: 40-hex SHA-1 tree OID passed where a 64-hex canonical digest is required, padded SHA-1, wrong object format/commit/tree pair, index tree differing from `HEAD^{tree}`, dirty-but-status-stable source bytes, relevant untracked source/catalog/license file and ambient-worktree mount all reject before any Docker/resource/signature/output side effect. Every failure must occur before any evidence signature or output file creation.

- [ ] **Step 2: Run spec-set tests and verify RED**

Run: `go test ./internal/c12evidence -run '^TestTrackedSpecSet' -count=1`

Expected: FAIL because the sole tracked builder/validator is undefined.

- [ ] **Step 3: GREEN — implement the only spec-set builder/validator**

Use the fixed manifest path with no caller override. First resolve and verify the typed Git locator, compute the separate canonical tree digest by the frozen full-tree formula, and reject index/HEAD/source-closure/untracked drift before materializing or reading the fixed snapshot. Strict-decode one base and the complete approved amendment array, require bytewise canonical path order and exact current three-member set, read each exact regular-file blob from that captured tree, verify tracked mode and Markdown approval state, and compare SHA-256 to the lowercase manifest value. JCS-encode only the typed `C12SpecSetV1`, prepend `talenro.c12.spec-set.v1\x00`, compute the one `SpecDigest`, and return defensive copies. `ValidateTrackedC12SpecDigest` must rerun the same builder and constant-time compare；it cannot trust a persisted/caller tree/spec digest, confuse a Git OID with `TrackedTreeDigest`, or accept a separately supplied member list.

- [ ] **Step 4: Run spec-set tests and verify GREEN**

Run: `go test ./internal/c12evidence -run '^TestTrackedSpecSet' -count=1`

Expected: PASS for the exact tracked manifest/tree and FAIL for every order/path/hash/tree/approval/tamper mutation, including the Authority Protocol v7 second amendment.

- [ ] **Step 5: RED — write scope completeness and anti-substitution tests**

```go
func TestContainerEvidenceCannotSatisfyProductionScopes(t *testing.T) {
	t.Parallel()
	for _, scope := range []Scope{ScopeLinuxPlatformTrust, ScopeAuthorityFencePITR} {
		evidence := validScopeEvidence(scope)
		evidence.ProviderClass = ProviderContainerDeterministic
		if err := ValidateScopeEvidence(evidence, testTrustPolicy(), evidence.FinishedAt);
			!errors.Is(err, ErrProviderClass) {
			t.Fatalf("container fake satisfied production scope %s", scope)
		}
	}
}

func TestScopeEvidenceExpiresNoLaterThanSeventyTwoHours(t *testing.T) {
	t.Parallel()
	evidence := validScopeEvidence(ScopeWindowsPowerShellDocker)
	evidence.ExpiresAt = evidence.FinishedAt.Add(72*time.Hour + time.Nanosecond)
	if err := ValidateScopeEvidence(evidence, testTrustPolicy(), evidence.FinishedAt); !errors.Is(err, ErrExpiry) {
		t.Fatal("scope evidence exceeded maximum lifetime")
	}
}

func TestScopeEvidenceRejectsMissingOrSplicedAuthorityOperationsArtifact(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*C12ScopeEvidenceV1, *TrustPolicy)
	}{
		{"binary-component", func(e *C12ScopeEvidenceV1, _ *TrustPolicy) { e.AuthorityOperationsArtifact.BinarySHA256[0] ^= 1 }},
		{"image-component", func(e *C12ScopeEvidenceV1, _ *TrustPolicy) { e.AuthorityOperationsArtifact.ImageManifestDigest[0] ^= 1 }},
		{"build-receipt-component", func(e *C12ScopeEvidenceV1, _ *TrustPolicy) { e.AuthorityOperationsArtifact.BuildReceiptDigest[0] ^= 1 }},
		{"dependency-component", func(e *C12ScopeEvidenceV1, _ *TrustPolicy) { e.AuthorityOperationsArtifact.DependencyClosureDigest[0] ^= 1 }},
		{"combined", func(e *C12ScopeEvidenceV1, _ *TrustPolicy) { e.AuthorityOperationsBinaryAndImageDigest[0] ^= 1 }},
		{"supply-chain", func(e *C12ScopeEvidenceV1, _ *TrustPolicy) { e.AuthorityOperationsSupplyChainRecordSetDigest[0] ^= 1 }},
		{"supply-chain-bundle", func(e *C12ScopeEvidenceV1, _ *TrustPolicy) { e.SupplyChainEvidenceBundleDigest[0] ^= 1 }},
		{"catalog", func(e *C12ScopeEvidenceV1, _ *TrustPolicy) { e.ArtifactCatalogDigest[0] ^= 1 }},
		{"process-policy", func(e *C12ScopeEvidenceV1, _ *TrustPolicy) { e.ExpectedProcessImagePolicyDigest[0] ^= 1 }},
		{"image-set", func(e *C12ScopeEvidenceV1, _ *TrustPolicy) { e.ImageSetDigest[0] ^= 1 }},
		{"full-scan", func(e *C12ScopeEvidenceV1, _ *TrustPolicy) { e.ArtifactScanDigest[0] ^= 1 }},
		{"policy-tuple", func(_ *C12ScopeEvidenceV1, p *TrustPolicy) { p.ExpectedAuthorityOperationsArtifact.BinarySHA256[0] ^= 1 }},
		{"policy-combined", func(_ *C12ScopeEvidenceV1, p *TrustPolicy) { p.ExpectedAuthorityOperationsBinaryAndImageDigest[0] ^= 1 }},
		{"policy-supply-chain", func(_ *C12ScopeEvidenceV1, p *TrustPolicy) { p.ExpectedAuthorityOperationsSupplyChainRecordSetDigest[0] ^= 1 }},
		{"policy-supply-chain-bundle", func(_ *C12ScopeEvidenceV1, p *TrustPolicy) { p.ExpectedSupplyChainEvidenceBundleDigest[0] ^= 1 }},
		{"policy-catalog", func(_ *C12ScopeEvidenceV1, p *TrustPolicy) { p.ExpectedArtifactCatalogDigest[0] ^= 1 }},
		{"policy-process", func(_ *C12ScopeEvidenceV1, p *TrustPolicy) { p.ExpectedProcessImagePolicyDigest[0] ^= 1 }},
		{"policy-image-set", func(_ *C12ScopeEvidenceV1, p *TrustPolicy) { p.ExpectedImageSetDigest[0] ^= 1 }},
		{"policy-full-scan", func(_ *C12ScopeEvidenceV1, p *TrustPolicy) { p.ExpectedArtifactScanDigest[0] ^= 1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evidence, policy := validScopeEvidence(ScopeWindowsPowerShellDocker), testTrustPolicy()
			tt.mutate(&evidence, &policy)
			if err := ValidateScopeEvidence(evidence, policy, evidence.FinishedAt); !errors.Is(err, ErrBuildMismatch) {
				t.Fatal("scope evidence accepted a scanner-derived binding splice")
			}
		})
	}
}
```

- [ ] **Step 6: Run the focused scope tests and verify RED**

Run: `go test ./internal/c12evidence -run '^TestContainerEvidence|^TestScopeEvidence' -count=1`

Expected: FAIL because scope evidence types and validator are undefined.

- [ ] **Step 7: GREEN — add the closed enums and strict validator**

Validation must require schema `talenro-c12-scope-evidence/v1`, one of the four exact scopes, nonzero run/times/digests, `StartedAt <= FinishedAt <= ExpiresAt <= FinishedAt+72h`, and a trusted runner attestation over the exact domain-separated `CanonicalScopeEvidenceUnsignedPayload` transcript frozen above；the full canonical envelope is computed only after that verification. It also requires exact expected repo/tree/spec/toolchain/three-release/authority-operations/catalog/process-policy/full-scan/image inputs. Every component of `AuthorityOperationsArtifact`, `AuthorityOperationsBinaryAndImageDigest`, `AuthorityOperationsSupplyChainRecordSetDigest`, `SupplyChainEvidenceBundleDigest`, `ArtifactCatalogDigest`, `ExpectedProcessImagePolicyDigest`, `ImageSetDigest` and `ArtifactScanDigest` must be nonzero and constant-time equal its corresponding `TrustPolicy.Expected*` value；the command layer obtains those expected values only from Task 3's validated canonical clean receipt and bundle. Task 3 derives `ImageSetDigest` over every catalog OCI image/layout role, including the exact operations image manifest；there is no opaque-membership assertion against an unrelated projection. The command also recomputes the combined digest from the tuple and receipt ruleset before constructing policy. Zero, caller-supplied, stale binary-only, same-binary/different-image, same-image/different-binary, build/dependency, supply-chain-record/bundle, catalog, process-policy, image-role or same-operations/different-full-receipt splices reject. `scope.go` neither imports `internal/artifactscan` nor recomputes its domains. `SpecDigest` must equal a fresh `ValidateTrackedC12SpecDigest` result for the exact typed Git tree locator whose separately recomputed canonical digest equals `TrackedTreeDigest`；a caller literal, Git-OID substitution, single-spec hash, old two-member set digest or cross-tree digest rejects. Windows scopes require `container_deterministic`; production scopes require `production`. Unknown JSON fields, duplicate fields, non-UTC timestamps and a scope/attestation mismatch fail closed.

- [ ] **Step 8: Run all evidence tests and verify GREEN**

Run: `go test ./internal/c12evidence -count=1`

Expected: PASS for the canonical fixture; field removal, field duplication, expiry, cross-scope rewrite, stale build and fake-provider mutations all fail.

- [ ] **Step 9: REFACTOR — fuzz strict spec-set and scope decoding**

```go
func FuzzScopeEvidence(f *testing.F) {
	f.Add(readFixture(f, "container-scope-valid.json"))
	f.Fuzz(func(t *testing.T, input []byte) {
		_, _ = DecodeScopeEvidence(bytes.NewReader(input), 64<<10)
	})
}
```

Run: `go test ./internal/c12evidence -run '^$' -fuzz '^FuzzScopeEvidence$' -fuzztime=10s -timeout 30s`

Run: `go test ./internal/c12evidence -run '^$' -fuzz '^FuzzC12SpecSet$' -fuzztime=10s -timeout 30s`

Expected: PASS with no panic, unbounded allocation or raw-input error echo.

- [ ] **Step 10: Commit the sole spec-set and evidence boundary**

```bash
git add internal/c12evidence/specset.go internal/c12evidence/specset_test.go internal/c12evidence/scope.go internal/c12evidence/scope_test.go internal/c12evidence/canonical.go testdata/c12/evidence/container-scope-valid.json
git commit -m "feat: bind C1.2 evidence to canonical spec set"
```

### Task 2: Append-only ownership WAL, exact recovery and cleanup exit bits

**Files:**
- Create: `internal/c12runnerprofile/profile.go`
- Create: `internal/c12runnerprofile/profile_test.go`
- Create: `internal/c12runnerprofile/bootstrap.go`
- Create: `internal/c12runnerprofile/bootstrap_test.go`
- Create: `internal/c12runnerprofile/completion_state.go`
- Create: `internal/c12runnerprofile/completion_state_test.go`
- Create: `internal/c12runnerprofile/completion_child_runtime.go`
- Create: `internal/c12runnerprofile/completion_child_runtime_test.go`
- Create (first line `//go:build windows`): `internal/c12runnerprofile/completion_child_runtime_windows.go`
- Create (first line `//go:build windows`): `internal/c12runnerprofile/completion_child_runtime_windows_test.go`
- Create (first line `//go:build !windows`): `internal/c12runnerprofile/completion_child_runtime_unsupported.go`
- Create: `internal/c12runnerprofile/platform_resume.go`
- Create: `internal/c12runnerprofile/platform_resume_test.go`
- Create: `internal/c12runnerprofile/authority_cleanup.go`
- Create: `internal/c12runnerprofile/authority_cleanup_test.go`
- Create: `internal/c12runnerprofile/staged_diagnostic_cleanup.go`
- Create: `internal/c12runnerprofile/staged_diagnostic_cleanup_test.go`
- Create: `internal/c12runnerprofile/cleanup_finalization.go`
- Create: `internal/c12runnerprofile/native_trusted_time_runtime.go`
- Create: `internal/c12runnerprofile/native_trusted_time_runtime_test.go`
- Create (first line `//go:build linux`): `internal/c12runnerprofile/native_trusted_time_runtime_linux.go`
- Create (first line `//go:build linux`): `internal/c12runnerprofile/native_trusted_time_runtime_linux_test.go`
- Create (first line `//go:build !linux`): `internal/c12runnerprofile/native_trusted_time_runtime_unsupported.go`
- Create: `internal/c12runnerprofile/state_guard.go`
- Create: `internal/c12runnerprofile/state_guard_test.go`
- Create: `internal/c12runnerprofile/state_guard_windows.go`
- Create: `internal/c12runnerprofile/state_guard_windows_test.go`
- Create: `internal/c12runnerprofile/state_guard_linux.go`
- Create: `internal/c12runnerprofile/state_guard_linux_test.go`
- Create: `testdata/c12/windows/runner-profile.v1.json`
- Create: `testdata/c12/diagnostic/runner-profile.v1.json`
- Create: `internal/c12evidence/wal.go`
- Create: `internal/c12evidence/wal_test.go`
- Create: `internal/c12evidence/cleanup.go`
- Create: `internal/c12evidence/cleanup_test.go`
- Create: `internal/c12cleanup/core.go`
- Create: `internal/c12cleanup/core_test.go`
- Create (first line `//go:build windows`): `internal/c12cleanup/protected_key_windows.go`
- Create (first line `//go:build windows`): `internal/c12cleanup/protected_key_windows_test.go`
- Create (first line `//go:build linux`): `internal/c12cleanup/protected_key_linux.go`
- Create (first line `//go:build linux`): `internal/c12cleanup/protected_key_linux_test.go`
- Create: `internal/c12evidence/runnerhelper.go`
- Create: `internal/c12evidence/runnerhelper_test.go`
- Create: `cmd/talenro-c12-runner-helper/main.go`
- Create: `cmd/talenro-c12-runner-helper/main_test.go`
- Create: `cmd/talenro-c12-runner-launcher/main.go`
- Create: `cmd/talenro-c12-runner-launcher/main_test.go`
- Create (first line `//go:build windows`): `cmd/talenro-c12-runner-launcher/launcher_windows.go`
- Create (first line `//go:build windows`): `cmd/talenro-c12-runner-launcher/launcher_windows_test.go`
- Create (first line `//go:build linux`): `cmd/talenro-c12-runner-launcher/launcher_linux.go`
- Create (first line `//go:build linux`): `cmd/talenro-c12-runner-launcher/launcher_linux_test.go`
- Test: `internal/c12evidence/wal_test.go`
- Test: `internal/c12evidence/cleanup_test.go`
- Test: `internal/c12evidence/runnerhelper_test.go`
- Test: `cmd/talenro-c12-runner-helper/main_test.go`
- Test: `cmd/talenro-c12-runner-launcher/main_test.go`

**Interfaces:**
- Consumes: validated 128-bit run ID, OS-keystore protected HMAC key, exact resource inspectors/deleters reconstructed only inside guard-selected runnerprofile successors from their sealed installed/provider/OS runtime capabilities, `contracts.Digest`, and a fixed-token tracked runner policy authenticated against a fixed-location sealed machine bootstrap.
- Produces: the dependency-leaf `internal/c12cleanup` package that owns the cleanup capsule/resource schemas plus sealed `Capability`/`TerminalProof` core, and the neutral `internal/c12runnerprofile` package with the sole strict/JCS/domain implementation of `C12NativeRunnerProfileV1`, `C12NativeRunnerInstallReceiptV1`, `C12FixedRunnerLauncherAttestationV1` and opaque authenticated native session；the fixed machine-base `talenro-c12-runner-launcher` plus sole `RunFixedNativeRunnerMode` launcher→guard-selected-helper bridge；then the high-level `BindCleanupOnlyOwnershipPlan` facade followed by the intent-first `BeginProtectedWALKeyBootstrap`/`CreateOrRecoverProtectedWALKeyBootstrap`/`SealAndSelectProtectedWALKeyBootstrap` chain, its sole `OpenSelectedOwnershipWAL` consumer, sealed `SelectedOwnershipWAL.AppendIntent|AppendActual|AppendCleaned|AppendNotFound`, `RecoverExact`, `ExecuteCleanup`, `ExitBits(primaryFailed, cleanupFailed bool) int`, plus the versioned native `talenro-c12-runner-helper` core. `internal/c12evidence` exposes aliases/thin semantic adapters over the leaf cleanup types, while both `c12evidence` and `c12runnerprofile` may import `c12cleanup`; the leaf imports neither, so the fixed platform finalizer never creates an evidence↔runnerprofile cycle. `ResourceType` is frozen here as a closed enum covering ownership-WAL file, minimal Git object database, outer container/network/volume, process/Windows Job, scheduled task, repository-external tree/build/run root, resume blob, external VM, TPM/vTPM NV handle, PITR database/timeline, ephemeral provider credential/certificate and bounded run path；each type has one typed exact inspector/deleter interface and no generic shell/name-list fallback. At the Task 2 commit the unbound helper exposes only its closed non-machine ingresses `capture-tree`, `wal-intent`, `wal-actual`, `wal-recover`, and `wal-clean`；the launcher exposes only its OS-valid fixed run/provision names, and the helper's machine-native dispatcher is inherited-binding-only. Task 3 extends the prebuilt helper layer with launcher-bound `capture-staged-and-build-closed-set`, private parent-authenticated `build-closed-set` child and private `scope-from-verifier-result` **unsigned projection builder** only after `internal/artifactscan` exists；Task 4 adds the separate nested-container `validate-capsule-and-build-closed-set` ingress、fixture-only `compose-config-contract` and parent-private manifest-aware `create-verifier-trust`/`run-inner-integration` only after those schemas/instances exist. Tasks 6–7 finally add launcher-bound production `run-windows-final-scopes` and the post-clean two-scope signer；there is no public per-shell finalizer or exposed Go object/handle constructor. `internal/c12evidence/runnerhelper.go` never imports `internal/artifactscan` and never duplicates its receipt/bundle schema or digest domains；only the `cmd/talenro-c12-runner-helper` command package may import both packages and own the producer boundary.

Task 2 freezes the runner-profile/bootstrap boundary before any Task 3 command can use it:

```go
package c12runnerprofile

type NativeRunnerRole string

const (
	NativeRunnerWindowsScopes NativeRunnerRole = "windows_scopes"
	NativeRunnerPlatform      NativeRunnerRole = "platform"
	NativeRunnerAuthority     NativeRunnerRole = "authority"
	NativeRunnerCompletion    NativeRunnerRole = "completion"
	NativeRunnerStagedDiagnostic NativeRunnerRole = "staged_diagnostic"
)

type NativeRunnerPublicMode string

const (
	NativeRunnerModeWindowsFinalScopes       NativeRunnerPublicMode = "run-windows-final-scopes"
	NativeRunnerModeStagedDiagnostic         NativeRunnerPublicMode = "capture-staged-and-build-closed-set"
	NativeRunnerModePlatformPhaseA           NativeRunnerPublicMode = "run-platform-phase-a"
	NativeRunnerModePlatformPhaseB           NativeRunnerPublicMode = "resume-platform-phase-b"
	NativeRunnerModeAuthorityFinal           NativeRunnerPublicMode = "run-authority-final"
	NativeRunnerModeCompletionCreate         NativeRunnerPublicMode = "completion-consolidate-create"
	NativeRunnerModeCompletionRevalidate     NativeRunnerPublicMode = "completion-revalidate"
)

type NativeRunnerInstallReceiptPolicyV1 struct {
	RequiredReceiptRoles       []string
	InstallAuthorityPurpose    string
	MaxReceiptAgeSeconds       uint32
	RequireApplicationControl bool
}

type C12NativeRunnerProfileV1 struct {
	SchemaVersion             string
	Role                      NativeRunnerRole
	AllowedTrackedScriptTokens []string
	RequiredToolRoles         []string
	HelperCatalogRole         string
	BootstrapLauncherCatalogRole string
	RecoverySlotPurpose       string
	ExternalRootPurpose       string
	ChannelBindings           []C12NativeRunnerChannelBindingV1
	ProviderCredentialPurpose string
	EvidenceRootPurpose       string
	InstallReceiptPolicy      NativeRunnerInstallReceiptPolicyV1
}

type C12NativeRunnerChannelBindingV1 struct {
	TransactionKind string
	Access          string // sender or receiver
	SlotPurpose     string // symbolic policy purpose; machine identity remains sealed
	SequencePolicy  string // exactly "next"
}

type C12NativeRunnerInstallReceiptV1 struct {
	SchemaVersion                  string
	RunnerRole                     NativeRunnerRole
	CatalogRole                    string
	OS                             string
	Arch                           string
	AbsolutePath                   string
	ExecutableSHA256               contracts.Digest
	FileIdentityDigest             contracts.Digest
	ApplicationControlPolicyDigest contracts.Digest
	InstalledAt                    time.Time
	Attestation                    []byte
}

type C12FixedRunnerLauncherAttestationV1 struct {
	SchemaVersion                  string
	LauncherCatalogRole            string
	OS                             string
	Arch                           string
	AbsolutePath                   string
	ExecutableSHA256               contracts.Digest
	FileIdentityDigest             contracts.Digest
	MachineImageIdentityDigest     contracts.Digest
	ApplicationControlPolicyDigest contracts.Digest
	ToolchainProjectionDigest      contracts.Digest
	Attestation                    []byte
}

type AuthenticatedNativeRunnerSession struct {
	sealedState privateAuthenticatedNativeRunnerState
}

type ProtectedMachineInstallerSession interface {
	protectedMachineInstallerSession() // fixed machine/application-control authority
}

type ProtectedWALKeyBootstrapDispatch string

const (
	ProtectedWALKeyBootstrapRoleDispatch   ProtectedWALKeyBootstrapDispatch = "role_dispatch"
	ProtectedWALKeyBootstrapAbortUnstarted ProtectedWALKeyBootstrapDispatch = "abort_unstarted"
	ProtectedWALKeyBootstrapAbortIntent     ProtectedWALKeyBootstrapDispatch = "abort_intent"
)

type FixedProtectedWALKeyBootstrap interface {
	fixedProtectedWALKeyBootstrap() // runnerprofile-sealed stateful continuation; exposes no leaf intent/handle/backend method
}

type FixedCleanupOnlyOwnershipPlan interface {
	fixedCleanupOnlyOwnershipPlan() // session/runtime-bound exact plan; opaque, linear and nonserializable
}

type FixedCleanupOnlyOwnershipPlanView struct {
	Plan              c12cleanup.CleanupOnlyOwnershipPlanV1
	PlanBindingDigest contracts.Digest
}

type FixedProtectedWALKeyBootstrapView struct {
	PlanBindingDigest              contracts.Digest
	BackendObjectReservationDigest contracts.Digest
	WALKeyProtectedHandleDigest    contracts.Digest
}

type FixedSelectedCleanupOnlyOwnership interface {
	fixedSelectedCleanupOnlyOwnership() // runnerprofile-sealed selected state plus leaf binding; opaque, linear and nonserializable
}

type FixedCleanupOnlySelectionView struct {
	CapsuleEnvelopeDigest         contracts.Digest
	WALKeyProtectedHandleDigest   contracts.Digest
	OwnershipWALReservationDigest contracts.Digest
}

type StagedDiagnosticRecoveryDispatch string

const (
	StagedDiagnosticRecoveryCleanupOnly                 StagedDiagnosticRecoveryDispatch = "cleanup_only"
	StagedDiagnosticRecoveryOrdinaryFinalizationPending StagedDiagnosticRecoveryDispatch = "ordinary_finalization_pending"
)

type StagedDiagnosticCleanupSuccessor interface {
	stagedDiagnosticCleanupSuccessor() // sealed selected diagnostic cleanup owner; no workload/publish authority
}

type WindowsCleanupOrigin string

const (
	WindowsCleanupAfterSuccess WindowsCleanupOrigin = "success"
	WindowsCleanupAfterAbort   WindowsCleanupOrigin = "abort"
)

type WindowsRecoveryDispatch string

const (
	WindowsRecoveryCleanupOnly                  WindowsRecoveryDispatch = "cleanup_only"
	WindowsRecoveryOrdinaryFinalizationPending  WindowsRecoveryDispatch = "ordinary_finalization_pending"
	WindowsRecoveryPostCleanupPublication       WindowsRecoveryDispatch = "post_cleanup_publication"
)

type WindowsCleanupSuccessor interface {
	windowsCleanupSuccessor() // sealed no-workload exact-clean authority
}

type WindowsPostCleanupPublicationSuccessor interface {
	windowsPostCleanupPublicationSuccessor() // success-only same-generation attest/send authority after exact absence
}

type PlatformResumeManifestSink interface {
	platformResumeManifestSink() // sealed Phase-A one-use fixed-slot guarded transfer
}

type PlatformResumeSuccessor interface {
	platformResumeSuccessor() // sealed Phase-B read/validate/resume authority
}

type PlatformResumeRecoveryDispatch string

const (
	PlatformResumeRecoveryValidateResumeFresh PlatformResumeRecoveryDispatch = "validate_resume_fresh"
	PlatformResumeRecoveryValidateResumeRecovery PlatformResumeRecoveryDispatch = "validate_resume_recovery"
	PlatformResumeRecoveryWorkloadReady    PlatformResumeRecoveryDispatch = "workload_ready"
	PlatformResumeRecoveryCleanupOnly      PlatformResumeRecoveryDispatch = "cleanup_only"
	PlatformResumeRecoveryOrdinaryFinalizationPending PlatformResumeRecoveryDispatch = "ordinary_finalization_pending"
	PlatformResumeRecoveryPostCleanup      PlatformResumeRecoveryDispatch = "post_cleanup_publication"
)

type PlatformResumeCleanupSuccessor interface {
	platformResumeCleanupSuccessor() // sealed exact-clean-only downgraded authority
}

type PlatformResumeCleanupOrigin string

const (
	PlatformResumeCleanupAfterSuccess PlatformResumeCleanupOrigin = "success"
	PlatformResumeCleanupAfterAbort   PlatformResumeCleanupOrigin = "abort"
)

type PlatformResumeWorkloadSuccessor interface {
	platformResumeWorkloadSuccessor() // sealed post-gate fixed Phase-B facade
}

type PlatformResumeWorkloadResult interface {
	platformResumeWorkloadResult() // sealed provider/resource/conformance result
}

type PlatformPostCleanupPublicationSuccessor interface {
	platformPostCleanupPublicationSuccessor() // sealed success-only attest/send authority after exact absence
}

type AuthorityCleanupOrigin string

const (
	AuthorityCleanupAfterSuccess AuthorityCleanupOrigin = "success"
	AuthorityCleanupAfterAbort   AuthorityCleanupOrigin = "abort"
)

type AuthorityRecoveryDispatch string

const (
	AuthorityRecoveryCleanupOnly                   AuthorityRecoveryDispatch = "cleanup_only"
	AuthorityRecoveryOrdinaryFinalizationPending   AuthorityRecoveryDispatch = "ordinary_finalization_pending"
	AuthorityRecoveryPostCleanupPublication        AuthorityRecoveryDispatch = "post_cleanup_publication"
)

type AuthorityCleanupSuccessor interface {
	authorityCleanupSuccessor() // sealed no-workload exact-clean authority
}

type AuthorityPostCleanupPublicationSuccessor interface {
	authorityPostCleanupPublicationSuccessor() // success-only same-generation attest/send authority after exact absence
}

const MaxFixedPostCleanupPublicationInputsBytes = 5 << 20

type AuthorityCleanupSuccessorView struct {
	Origin                              AuthorityCleanupOrigin
	CanonicalWorkloadResult             []byte // validated pre-clean result; present only for success origin
	CanonicalWorkloadResultDigest       contracts.Digest
	RunIdentityDigest                   contracts.Digest
	LaunchGeneration                    uint64
	ProviderSessionBindingDigest        contracts.Digest
	RecoveryAuthoritySlotIdentityDigest contracts.Digest
}

type AuthorityCleanupDescriptorV1 struct {
	CanonicalCapsuleEnvelope             []byte
	CapsuleEnvelopeDigest                contracts.Digest
	CleanupPolicyBindingDigest           contracts.Digest
	MachineSealLocatorDigest             contracts.Digest
	CanonicalWorkloadResultDigest        contracts.Digest // zero only for abort origin
	RunIdentityDigest                    contracts.Digest
	LaunchGeneration                     uint64
	RecoveryAuthoritySlotIdentityDigest  contracts.Digest
	sealedState                          privateAuthorityCleanupDescriptor
}

type AuthorityPostCleanupPublicationView struct {
	CanonicalWorkloadResult             []byte
	CanonicalScanReceiptProjection      []byte
	CanonicalSupplyChainBundleProjection []byte
	CanonicalWorkloadResultDigest       contracts.Digest
	PublicationInputsDigest             contracts.Digest
	TerminalCleanupProofDigest          contracts.Digest
	RunIdentityDigest                   contracts.Digest
	LaunchGeneration                    uint64
	SenderBindingDigest                 contracts.Digest
}

type StagedDiagnosticCleanupSuccessorView struct {
	RunIdentityDigest                    contracts.Digest
	LaunchGeneration                    uint64
	RecoveryAuthoritySlotIdentityDigest contracts.Digest
}

type StagedDiagnosticCleanupDescriptorV1 struct {
	CanonicalCapsuleEnvelope             []byte
	CapsuleEnvelopeDigest                contracts.Digest
	CleanupPolicyBindingDigest           contracts.Digest
	MachineSealLocatorDigest             contracts.Digest
	RunIdentityDigest                    contracts.Digest
	LaunchGeneration                     uint64
	RecoveryAuthoritySlotIdentityDigest  contracts.Digest
	sealedState                          privateStagedDiagnosticCleanupDescriptor
}

type PlatformResumeWorkloadResultView struct {
	CanonicalResult                         []byte
	CanonicalScanReceiptProjection          []byte
	CanonicalSupplyChainBundleProjection    []byte
	CanonicalResultDigest                   contracts.Digest
	CanonicalScanReceiptProjectionDigest    contracts.Digest
	CanonicalSupplyChainBundleProjectionDigest contracts.Digest
	CanonicalUnsignedManifestDigest         contracts.Digest
	AuthenticatedResumeEnvelopeDigest       contracts.Digest
	RunIdentityDigest                       contracts.Digest
	LaunchGeneration                        uint64
	ProviderSessionBindingDigest            contracts.Digest
}

type WindowsCleanupSuccessorView struct {
	Origin                          WindowsCleanupOrigin
	CanonicalPreCleanupResult       []byte // bounded two-scope unsigned facts plus exact shared receipt/bundle projection; success only
	CanonicalPreCleanupResultDigest contracts.Digest
	RunIdentityDigest               contracts.Digest
	LaunchGeneration                uint64
	RecoveryAuthoritySlotIdentityDigest contracts.Digest
}

type WindowsCleanupDescriptorV1 struct {
	CanonicalCapsuleEnvelope             []byte
	CapsuleEnvelopeDigest                contracts.Digest
	CleanupPolicyBindingDigest           contracts.Digest
	MachineSealLocatorDigest             contracts.Digest
	CanonicalPreCleanupResultDigest      contracts.Digest // zero only for abort origin
	RunIdentityDigest                    contracts.Digest
	LaunchGeneration                     uint64
	RecoveryAuthoritySlotIdentityDigest  contracts.Digest
	sealedState                          privateWindowsCleanupDescriptor
}

type WindowsPostCleanupPublicationView struct {
	CanonicalPreCleanupResult              []byte
	CanonicalScanReceiptProjection         []byte
	CanonicalSupplyChainBundleProjection   []byte
	CanonicalPreCleanupResultDigest        contracts.Digest
	PublicationInputsDigest                contracts.Digest
	TerminalCleanupProofDigest             contracts.Digest
	RunIdentityDigest                      contracts.Digest
	LaunchGeneration                       uint64
	SenderBindingDigest                    contracts.Digest
}

type PlatformResumeCleanupSuccessorView struct {
	Origin                        PlatformResumeCleanupOrigin
	CanonicalWorkloadResult       []byte // validated pre-clean result; present only for success origin
	CanonicalWorkloadResultDigest contracts.Digest
	Resume                        PlatformResumeSuccessorView
}

type PlatformResumeCleanupDescriptorV1 struct {
	CanonicalCapsuleEnvelope           []byte
	CapsuleEnvelopeDigest              contracts.Digest
	CleanupPolicyBindingDigest         contracts.Digest
	MachineSealLocatorDigest           contracts.Digest
	CanonicalWorkloadResultDigest      contracts.Digest // zero only for abort origin
	RunIdentityDigest                  contracts.Digest
	LaunchGeneration                   uint64
	ResumeSuccessorBindingDigest       contracts.Digest
	RecoveryAuthoritySlotIdentityDigest contracts.Digest
	sealedState                        privatePlatformResumeCleanupDescriptor
}

type PlatformPostCleanupPublicationView struct {
	CanonicalUnsignedManifest              []byte
	CanonicalWorkloadResult                 []byte
	CanonicalScanReceiptProjection          []byte
	CanonicalSupplyChainBundleProjection    []byte
	CanonicalUnsignedManifestDigest         contracts.Digest
	AuthenticatedResumeEnvelopeDigest       contracts.Digest
	CanonicalWorkloadResultDigest           contracts.Digest
	PublicationInputsDigest                 contracts.Digest
	TerminalCleanupProofDigest              contracts.Digest
	RunIdentityDigest                       contracts.Digest
	LaunchGeneration                        uint64
	SenderBindingDigest                     contracts.Digest
}

type NativeTrustedTimePurpose string

const (
	NativeTrustedTimePlatformPhaseA NativeTrustedTimePurpose = "platform_phase_a"
	NativeTrustedTimePlatformPhaseB NativeTrustedTimePurpose = "platform_phase_b"
	NativeTrustedTimeAuthority      NativeTrustedTimePurpose = "authority_evidence"
)

type NativeTrustedTimeRuntime interface {
	nativeTrustedTimeRuntime() // sealed role/purpose/session-bound one-use transport
}

type NativeTrustedTimeRuntimeView struct {
	Purpose                      NativeTrustedTimePurpose
	Role                         NativeRunnerRole
	RunIdentityDigest            contracts.Digest
	LaunchGeneration             uint64
	RunnerProfileDigest          contracts.Digest
	ProviderProfileDigest        contracts.Digest
	NativeClientProjectionDigest contracts.Digest
	ProviderSessionBindingDigest contracts.Digest
	RuntimeBindingDigest         contracts.Digest
}

type NativeTrustedTimeExchangeResult interface {
	nativeTrustedTimeExchangeResult() // sealed request/response/runtime transcript
}

type NativeTrustedTimeExchangeResultView struct {
	Runtime                    NativeTrustedTimeRuntimeView
	CanonicalRequest           []byte
	CanonicalRequestDigest     contracts.Digest
	CanonicalResponse          []byte
	CanonicalResponseDigest    contracts.Digest
	MonotonicElapsed           time.Duration
	MonotonicMeasurementDigest contracts.Digest
}

type PlatformResumeSuccessorView struct {
	CanonicalUnsignedManifest            []byte
	CanonicalUnsignedManifestDigest      contracts.Digest
	AuthenticatedResumeEnvelopeDigest    contracts.Digest
	RunIdentityDigest                    contracts.Digest
	LaunchGeneration                     uint64
	InstalledExecutableProjectionDigest  contracts.Digest
	ProviderProjectionDigest             contracts.Digest
	ResumeOnlySessionSuccessorDigest      contracts.Digest
	CleanupOnlyOwnershipCapsuleDigest     contracts.Digest
	OwnershipWALReservationDigest         contracts.Digest
	RecoveryAuthoritySlotIdentityDigest   contracts.Digest
}

type CompletionRoleStatePhase string

const (
	CompletionRoleStateIdle                   CompletionRoleStatePhase = "idle"
	CompletionRoleStateActiveUnstarted        CompletionRoleStatePhase = "active_unstarted"
	CompletionRoleStatePublishPending         CompletionRoleStatePhase = "publish_pending"
	CompletionRoleStateRetainedComplete       CompletionRoleStatePhase = "retained_complete"
	CompletionRoleStateRetainedCompleteActive CompletionRoleStatePhase = "retained_complete_active"
)

type CompletionRevalidationDispatch string

const (
	CompletionRevalidationDispatchFreshOpen     CompletionRevalidationDispatch = "fresh_open"
	CompletionRevalidationDispatchActiveRecovery CompletionRevalidationDispatch = "active_recovery"
	CompletionRevalidationDispatchOutcomeAudit  CompletionRevalidationDispatch = "outcome_audit"
)

type AuthenticatedCompletionRoleStateView struct {
	StateEpoch               uint64
	StateDigest              contracts.Digest
	PublishBindingDigest     contracts.Digest
	InstallGeneration        uint64
	Phase                    CompletionRoleStatePhase
	LaunchGeneration         uint64
	ImmutablePayloadDigest   contracts.Digest
	sealedState              privateCompletionRoleStateView
}

type CompletionInnerCleanupHandle interface {
	completionInnerCleanupHandle() // opaque exact declared temporary set and one run/generation
}

type CompletionInnerCleanupDescriptorV1 struct {
	CanonicalCapsuleEnvelope   []byte // bounded defensive copy; no path/key/create authority
	CapsuleEnvelopeDigest      contracts.Digest
	CleanupPolicyBindingDigest contracts.Digest
	LaunchGeneration           uint64
	MachineSealLocatorDigest   contracts.Digest
	CoordinatorBindingDigest   contracts.Digest
	sealedState                privateCompletionInnerCleanupDescriptor
}

type CompletionRetainedFinalizationReceipt interface {
	completionRetainedFinalizationReceipt() // opaque read-only terminal audit; no generation/channel/mutation authority
}

type CompletionRetainedRevalidationOutcomeReceipt interface {
	completionRetainedRevalidationOutcomeReceipt() // only exact delivery-ACK authority; no business/generation/channel mutation
}

type CompletionRetainedRevalidationOutcomeDeliveryReceipt interface {
	completionRetainedRevalidationOutcomeDeliveryReceipt() // fixed launcher result-slot file+directory fsync proof
}

type CompletionRetainedRevalidationOutcomeView struct {
	Outcome                      CompletionRetainedRevalidationOutcome
	LaunchGeneration             uint64
	ImmutablePayloadDigest       contracts.Digest
	LatestTrustedTimeHead        CompletionRetainedTrustedTimeHeadView
}

// Task 2 owns this enum because the child runtime needs it before Task 7 adds
// the publication channel. Task 7 consumes this declaration and adds no copy.
type CompletionEvidenceMemberRole string

const (
	CompletionMemberWindowsPowerShellScope CompletionEvidenceMemberRole = "windows_powershell_scope"
	CompletionMemberWindowsGitBashScope    CompletionEvidenceMemberRole = "windows_git_bash_scope"
	CompletionMemberLinuxPlatformScope     CompletionEvidenceMemberRole = "linux_platform_scope"
	CompletionMemberAuthorityFenceScope    CompletionEvidenceMemberRole = "authority_fence_scope"
	CompletionMemberPlatformEvidence       CompletionEvidenceMemberRole = "platform_evidence"
	CompletionMemberAuthorityEvidence      CompletionEvidenceMemberRole = "authority_evidence"
	CompletionMemberScanReceipt            CompletionEvidenceMemberRole = "scan_receipt"
	CompletionMemberSupplyChainBundle      CompletionEvidenceMemberRole = "supply_chain_bundle"
)

type CompletionValidationPurpose string

const (
	CompletionValidationCreate     CompletionValidationPurpose = "completion_create"
	CompletionValidationRevalidate CompletionValidationPurpose = "completion_revalidate"
)

type CompletionValidationChildRuntime interface {
	completionValidationChildRuntime() // opaque one-use child/provider/process capability
}

type CompletionValidationChildInheritedSession interface {
	completionValidationChildInheritedSession() // exact direct-child inherited binding; no caller constructor
}

type CompletionValidationChildRequest struct {
	Purpose                           CompletionValidationPurpose
	CanonicalInputDigest              contracts.Digest
	MaterializedEvidenceBindingDigest contracts.Digest
}

type CompletionValidationChildResult interface {
	completionValidationChildResult() // sealed immutable authenticated execution result
}

type CompletionRetainedTrustedTimeHeadView struct {
	AcceptedGeneration      uint64
	CanonicalResponse       []byte // defensive canonical P09 response; absent only before the first accepted revalidation
	CanonicalResponseDigest contracts.Digest
}

type CompletionValidatedTrustedTimeHeadCandidate interface {
	completionValidatedTrustedTimeHeadCandidate() // sealed, one result/response/generation; not yet semantically affirmed
}

type CompletionRetainedRevalidationOutcome string

const (
	CompletionRetainedRevalidationCompleted CompletionRetainedRevalidationOutcome = "completed"
	CompletionRetainedRevalidationAborted   CompletionRetainedRevalidationOutcome = "aborted"
)

type CompletionValidationChildResultView struct {
	Purpose                           CompletionValidationPurpose
	CanonicalInputDigest              contracts.Digest
	MaterializedEvidenceBindingDigest contracts.Digest
	ChallengeNonce                    [32]byte
	CanonicalResponse                 []byte // bounded defensive copy; no handle/path/token
	CanonicalResponseDigest           contracts.Digest
	RunnerProfileDigest               contracts.Digest
	ProviderProfileDigest             contracts.Digest
	InstalledChildProjectionDigest    contracts.Digest
	ChildProcessBindingDigest         contracts.Digest
	ProviderSessionBindingDigest      contracts.Digest
	TrackedTreeBindingDigest          contracts.Digest
	RuntimeBindingDigest              contracts.Digest
	PreviousTrustedTimeHead           CompletionRetainedTrustedTimeHeadView
	TrustedTimeHeadCandidate          CompletionValidatedTrustedTimeHeadCandidate
	Evidence                          CompletionValidationChildEvidenceView
	TrackedTree                       CompletionValidationChildTrackedTreeReader
}

type CompletionValidationChildInheritedView struct {
	Purpose                           CompletionValidationPurpose
	CanonicalInputDigest              contracts.Digest
	MaterializedEvidenceBindingDigest contracts.Digest
	ChallengeNonce                    [32]byte
	RunnerProfileDigest               contracts.Digest
	ProviderProfileDigest             contracts.Digest
	InstalledChildProjectionDigest    contracts.Digest
	ChildProcessBindingDigest         contracts.Digest
	ProviderSessionBindingDigest      contracts.Digest
	TrackedTreeBindingDigest          contracts.Digest
	RuntimeBindingDigest              contracts.Digest
	PreviousTrustedTimeHead           CompletionRetainedTrustedTimeHeadView
}

type CompletionValidationChildEvidenceMemberV1 struct {
	Role            CompletionEvidenceMemberRole
	CanonicalBytes  []byte
	CanonicalDigest contracts.Digest
}

type CompletionValidationChildEvidenceView struct {
	MaterializedEvidenceBindingDigest contracts.Digest
	TrackedTreeBindingDigest           contracts.Digest
	Members                           [8]CompletionValidationChildEvidenceMemberV1
	RetainedManifestCanonical          []byte // absent only for completion_create
	RetainedManifestDigest             contracts.Digest
}

type CompletionValidationTrackedBlobRole string

const (
	CompletionTrackedSpecSetManifest          CompletionValidationTrackedBlobRole = "spec_set_manifest"
	CompletionTrackedSpecSetMember            CompletionValidationTrackedBlobRole = "spec_set_member"
	CompletionTrackedArtifactCatalog          CompletionValidationTrackedBlobRole = "artifact_catalog"
	CompletionTrackedToolchainLock            CompletionValidationTrackedBlobRole = "toolchain_lock"
	CompletionTrackedObligationDocument       CompletionValidationTrackedBlobRole = "obligation_document"
	CompletionTrackedCompletionContract       CompletionValidationTrackedBlobRole = "completion_contract"
	CompletionTrackedTrustedTimeProfile       CompletionValidationTrackedBlobRole = "trusted_time_profile"
	CompletionTrackedCompletionRunnerProfile CompletionValidationTrackedBlobRole = "completion_runner_profile"
)

type CompletionValidationTrackedBlobRequest struct {
	Role  CompletionValidationTrackedBlobRole
	Index uint8 // fixed role-local ordinal; zero for singleton roles
}

type CompletionValidationTrackedTreeMemberV1 struct {
	RepoRelativePath string
	GitMode          string
	BlobSHA256       contracts.Digest
}

type CompletionValidationTrackedBlobView struct {
	Request        CompletionValidationTrackedBlobRequest
	Member         CompletionValidationTrackedTreeMemberV1
	CanonicalBytes []byte
}

type CompletionValidationChildTrackedTreeView struct {
	BindingDigest                contracts.Digest
	ObjectFormat                 string
	CommitOID                    string
	TreeOID                      string
	CanonicalTrackedTreeDigest   contracts.Digest
	CanonicalInventoryDigest     contracts.Digest
	SnapshotObjectIdentityDigest contracts.Digest
	ObjectDatabaseIdentityDigest contracts.Digest
}

type CompletionValidationChildTrackedTreeReader interface {
	completionValidationChildTrackedTreeReader() // sealed, read-only and bound to one runtime/result
	Inspect(context.Context) (CompletionValidationChildTrackedTreeView, error)
	ReadCanonicalInventory(context.Context) ([]CompletionValidationTrackedTreeMemberV1, error)
	ReadFixedBlob(context.Context, CompletionValidationTrackedBlobRequest) (CompletionValidationTrackedBlobView, error)
}

type CompletionInnerRecoveryDispatch string

const (
	CompletionInnerRecoveryBeginRequired     CompletionInnerRecoveryDispatch = "begin_required"
	CompletionInnerRecoverySessionValidation CompletionInnerRecoveryDispatch = "session_validation_recovery"
	CompletionInnerRecoverySessionCleanup    CompletionInnerRecoveryDispatch = "session_cleanup_recovery"
	CompletionInnerRecoveryFinalization      CompletionInnerRecoveryDispatch = "finalization_recovery"
	CompletionInnerRecoveryComplete          CompletionInnerRecoveryDispatch = "complete"
)

type AuthenticatedCompletionRoleStateCoordinator interface {
	completionRoleStateCoordinator() // opaque; owns private store and closed semantic-permit mint/consume
	ReadCurrent(context.Context) (AuthenticatedCompletionRoleStateView, error)
	BeginPublishPending(context.Context) error
	RecoverPublishPending(context.Context) error
	BeginInnerCleanup(context.Context) (CompletionInnerCleanupHandle, error)
	InspectInnerRecoveryDispatch(context.Context) (CompletionInnerRecoveryDispatch, error)
	RecoverInnerCleanup(context.Context) (CompletionInnerCleanupHandle, error)
	RecoverInnerFinalization(context.Context) error
	FinalizeInnerCleanup(context.Context, CompletionInnerCleanupHandle, c12cleanup.FinalizationReceipt) error
	OpenRetainedComplete(context.Context) error
	RecoverRetainedComplete(context.Context) error
	FinalizeRetainedRevalidation(context.Context, CompletionInnerCleanupHandle, c12cleanup.FinalizationReceipt, CompletionValidatedTrustedTimeHeadCandidate) error
	AbortRetainedRevalidation(context.Context, CompletionInnerCleanupHandle, c12cleanup.FinalizationReceipt) error
	AbortRetainedRevalidationWithTrustedTimeHead(context.Context, CompletionInnerCleanupHandle, c12cleanup.FinalizationReceipt, CompletionValidatedTrustedTimeHeadCandidate) error
}

func CanonicalNativeRunnerProfile(C12NativeRunnerProfileV1) ([]byte, contracts.Digest, error)
func CanonicalNativeRunnerInstallReceipt(C12NativeRunnerInstallReceiptV1) ([]byte, contracts.Digest, error)
func CanonicalFixedRunnerLauncherAttestation(C12FixedRunnerLauncherAttestationV1) ([]byte, contracts.Digest, error)
func VerifyFixedMachineStateGuardProvider(context.Context) error
func OpenFixedMachineInstallerSession(context.Context) (ProtectedMachineInstallerSession, error)
func InstallFixedNativeRunnerProfile(context.Context, NativeRunnerRole, ProtectedMachineInstallerSession) ([]C12NativeRunnerInstallReceiptV1, error)
func VerifyFixedNativeRunnerInstallation(context.Context, NativeRunnerRole) error
func RunFixedNativeRunnerMode(context.Context, NativeRunnerPublicMode) (int, error)
func FixedPostCleanupPublicationInputsDigest(NativeRunnerRole, ...[]byte) (contracts.Digest, error)
func CompleteFixedWindowsScopesToCleanup(context.Context, *AuthenticatedNativeRunnerSession, []byte, []byte, []byte) (WindowsCleanupSuccessor, error) // canonical pre-clean result, scan-receipt projection, supply-chain-bundle projection
func AbortFixedWindowsScopesToCleanup(context.Context, *AuthenticatedNativeRunnerSession) (WindowsCleanupSuccessor, error)
func InspectFixedProtectedWALKeyBootstrapDispatch(context.Context) (ProtectedWALKeyBootstrapDispatch, error)
func AbortFixedUnstartedProtectedWALKeyBootstrap(context.Context) error
func AbortFixedProtectedWALKeyIntent(context.Context) error
func BindFixedCleanupOnlyOwnershipPlan(context.Context, *AuthenticatedNativeRunnerSession, c12cleanup.CleanupOnlyOwnershipPlanTemplateV1) (FixedCleanupOnlyOwnershipPlan, error)
func InspectFixedCleanupOnlyOwnershipPlan(FixedCleanupOnlyOwnershipPlan) (FixedCleanupOnlyOwnershipPlanView, error)
func BeginFixedProtectedWALKeyBootstrap(context.Context, *AuthenticatedNativeRunnerSession, FixedCleanupOnlyOwnershipPlan) (FixedProtectedWALKeyBootstrap, error)
func CreateOrRecoverFixedProtectedWALKeyBootstrap(context.Context, FixedProtectedWALKeyBootstrap) (FixedProtectedWALKeyBootstrapView, error)
func SealAndSelectFixedCleanupOnlyOwnershipCapsule(context.Context, FixedProtectedWALKeyBootstrap, c12cleanup.C12CleanupOnlyOwnershipCapsuleTemplateV1) (FixedSelectedCleanupOnlyOwnership, error)
func InspectFixedSelectedCleanupOnlyOwnership(FixedSelectedCleanupOnlyOwnership) (FixedCleanupOnlySelectionView, error)
func OpenFixedSelectedOwnershipWAL(context.Context, FixedSelectedCleanupOnlyOwnership) (c12cleanup.SelectedOwnershipWAL, error)
func InspectFixedStagedDiagnosticRecoveryDispatch(context.Context) (StagedDiagnosticRecoveryDispatch, error)
func RecoverFixedStagedDiagnosticCleanupSuccessor(context.Context) (StagedDiagnosticCleanupSuccessor, error)
func InspectFixedStagedDiagnosticCleanupSuccessor(context.Context, StagedDiagnosticCleanupSuccessor) (StagedDiagnosticCleanupSuccessorView, error)
func DescribeFixedStagedDiagnosticCleanup(context.Context, StagedDiagnosticCleanupSuccessor) (StagedDiagnosticCleanupDescriptorV1, error)
func OpenFixedStagedDiagnosticCleanupCapability(context.Context, StagedDiagnosticCleanupSuccessor) (c12cleanup.Capability, error)
func InspectFixedWindowsRecoveryDispatch(context.Context) (WindowsRecoveryDispatch, error)
func RecoverFixedWindowsCleanupSuccessor(context.Context) (WindowsCleanupSuccessor, error)
func InspectFixedWindowsCleanupSuccessor(context.Context, WindowsCleanupSuccessor) (WindowsCleanupSuccessorView, error)
func DescribeFixedWindowsCleanup(context.Context, WindowsCleanupSuccessor) (WindowsCleanupDescriptorV1, error)
func OpenFixedWindowsCleanupCapability(context.Context, WindowsCleanupSuccessor) (c12cleanup.Capability, error)
func FinalizeFixedWindowsCleanupToPublication(context.Context, WindowsCleanupSuccessor, c12cleanup.FinalizationReceipt) (WindowsPostCleanupPublicationSuccessor, error)
func RecoverFixedWindowsPostCleanupPublication(context.Context) (WindowsPostCleanupPublicationSuccessor, error)
func InspectFixedWindowsPostCleanupPublication(WindowsPostCleanupPublicationSuccessor) (WindowsPostCleanupPublicationView, error)
func AttestFixedWindowsPowerShellEvidence(context.Context, WindowsPostCleanupPublicationSuccessor, []byte) ([]byte, error)
func AttestFixedWindowsGitBashEvidence(context.Context, WindowsPostCleanupPublicationSuccessor, []byte) ([]byte, error)
func BeginFixedPlatformResumeTransition(context.Context, *AuthenticatedNativeRunnerSession) (PlatformResumeManifestSink, error)
func SealAndSelectFixedPlatformResumeManifest(context.Context, PlatformResumeManifestSink, []byte) (contracts.Digest, error)
func InspectFixedPlatformResumeRecoveryDispatch(context.Context) (PlatformResumeRecoveryDispatch, error)
func RecoverFixedPlatformResumeSuccessor(context.Context) (PlatformResumeSuccessor, error)
func InspectFixedPlatformResumeSuccessor(context.Context, PlatformResumeSuccessor) (PlatformResumeSuccessorView, error)
func BindNativeTrustedTimeRuntime(context.Context, *AuthenticatedNativeRunnerSession, NativeTrustedTimePurpose) (NativeTrustedTimeRuntime, error)
func BindPlatformResumeTrustedTimeRuntime(context.Context, PlatformResumeSuccessor) (NativeTrustedTimeRuntime, error)
func InspectNativeTrustedTimeRuntime(NativeTrustedTimeRuntime) (NativeTrustedTimeRuntimeView, error)
func ExchangeNativeTrustedTime(context.Context, NativeTrustedTimeRuntime, []byte) (NativeTrustedTimeExchangeResult, error)
func RecoverPlatformResumeTrustedTimeExchange(context.Context, PlatformResumeSuccessor) (NativeTrustedTimeExchangeResult, error)
func InspectNativeTrustedTimeExchangeResult(NativeTrustedTimeExchangeResult) (NativeTrustedTimeExchangeResultView, error)
func ConsumeNativeTrustedTimeExchangeResult(NativeTrustedTimeExchangeResult) error
func PromoteFixedPlatformResumeSuccessor(context.Context, PlatformResumeSuccessor, NativeTrustedTimeExchangeResult) (PlatformResumeWorkloadSuccessor, error)
func RecoverFixedPlatformResumeWorkloadSuccessor(context.Context) (PlatformResumeWorkloadSuccessor, error)
func RunFixedPlatformResumeWorkload(context.Context, PlatformResumeWorkloadSuccessor) (PlatformResumeWorkloadResult, error)
func InspectFixedPlatformResumeWorkloadResult(PlatformResumeWorkloadResult) (PlatformResumeWorkloadResultView, error)
func CompleteFixedPlatformResumeWorkloadToCleanup(context.Context, PlatformResumeWorkloadSuccessor, PlatformResumeWorkloadResult) (PlatformResumeCleanupSuccessor, error)
func AbortFixedPlatformResumeWorkloadToCleanup(context.Context, PlatformResumeWorkloadSuccessor) (PlatformResumeCleanupSuccessor, error)
func DowngradeFixedPlatformResumeSuccessorToCleanup(context.Context, PlatformResumeSuccessor) (PlatformResumeCleanupSuccessor, error)
func RecoverFixedPlatformResumeCleanupSuccessor(context.Context) (PlatformResumeCleanupSuccessor, error)
func InspectFixedPlatformResumeCleanupSuccessor(context.Context, PlatformResumeCleanupSuccessor) (PlatformResumeCleanupSuccessorView, error)
func DescribeFixedPlatformResumeCleanup(context.Context, PlatformResumeCleanupSuccessor) (PlatformResumeCleanupDescriptorV1, error)
func OpenFixedPlatformResumeCleanupCapability(context.Context, PlatformResumeCleanupSuccessor) (c12cleanup.Capability, error)
func FinalizeFixedCleanedOwnership(context.Context, c12cleanup.FinalizationReceipt) error
func RecoverFixedCleanedOwnership(context.Context) error
func FinalizeFixedPlatformResumeCleanupToPublication(context.Context, PlatformResumeCleanupSuccessor, c12cleanup.FinalizationReceipt) (PlatformPostCleanupPublicationSuccessor, error)
func RecoverFixedPlatformPostCleanupPublication(context.Context) (PlatformPostCleanupPublicationSuccessor, error)
func InspectFixedPlatformPostCleanupPublication(PlatformPostCleanupPublicationSuccessor) (PlatformPostCleanupPublicationView, error)
func AttestFixedPlatformNestedEvidence(context.Context, PlatformPostCleanupPublicationSuccessor, []byte) ([]byte, error)
func AttestFixedPlatformScopeEvidence(context.Context, PlatformPostCleanupPublicationSuccessor, []byte) ([]byte, error)
func CompleteFixedAuthorityWorkloadToCleanup(context.Context, *AuthenticatedNativeRunnerSession, []byte, []byte, []byte) (AuthorityCleanupSuccessor, error) // canonical workload result, scan-receipt projection, supply-chain-bundle projection
func AbortFixedAuthorityWorkloadToCleanup(context.Context, *AuthenticatedNativeRunnerSession) (AuthorityCleanupSuccessor, error)
func InspectFixedAuthorityRecoveryDispatch(context.Context) (AuthorityRecoveryDispatch, error)
func RecoverFixedAuthorityCleanupSuccessor(context.Context) (AuthorityCleanupSuccessor, error)
func InspectFixedAuthorityCleanupSuccessor(context.Context, AuthorityCleanupSuccessor) (AuthorityCleanupSuccessorView, error)
func DescribeFixedAuthorityCleanup(context.Context, AuthorityCleanupSuccessor) (AuthorityCleanupDescriptorV1, error)
func OpenFixedAuthorityCleanupCapability(context.Context, AuthorityCleanupSuccessor) (c12cleanup.Capability, error)
func FinalizeFixedAuthorityCleanupToPublication(context.Context, AuthorityCleanupSuccessor, c12cleanup.FinalizationReceipt) (AuthorityPostCleanupPublicationSuccessor, error)
func RecoverFixedAuthorityPostCleanupPublication(context.Context) (AuthorityPostCleanupPublicationSuccessor, error)
func InspectFixedAuthorityPostCleanupPublication(AuthorityPostCleanupPublicationSuccessor) (AuthorityPostCleanupPublicationView, error)
func AttestFixedAuthorityNestedEvidence(context.Context, AuthorityPostCleanupPublicationSuccessor, []byte) ([]byte, error)
func AttestFixedAuthorityScopeEvidence(context.Context, AuthorityPostCleanupPublicationSuccessor, []byte) ([]byte, error)
func InspectFixedCompletionRoleState(context.Context) (CompletionRoleStatePhase, error)
func InspectFixedCompletionRevalidationDispatch(context.Context) (CompletionRevalidationDispatch, error)
func InspectFixedCompletionRetainedFinalization(context.Context) (CompletionRetainedFinalizationReceipt, error)
func InspectLastFixedCompletionRevalidationOutcome(context.Context) (CompletionRetainedRevalidationOutcomeReceipt, error)
func InspectCompletionRetainedRevalidationOutcomeReceipt(CompletionRetainedRevalidationOutcomeReceipt) (CompletionRetainedRevalidationOutcomeView, error)
func PersistLastFixedCompletionRevalidationOutcomeDelivery(context.Context, CompletionRetainedRevalidationOutcomeReceipt) (CompletionRetainedRevalidationOutcomeDeliveryReceipt, error)
func AcknowledgeLastFixedCompletionRevalidationOutcome(context.Context, CompletionRetainedRevalidationOutcomeReceipt, CompletionRetainedRevalidationOutcomeDeliveryReceipt) error
func OpenInstalledNativeRunnerSession(context.Context, NativeRunnerRole) (*AuthenticatedNativeRunnerSession, error)
func OpenCompletionRoleStateForSession(context.Context, *AuthenticatedNativeRunnerSession) (AuthenticatedCompletionRoleStateCoordinator, error)
func RecoverFixedCompletionRoleState(context.Context) (AuthenticatedCompletionRoleStateCoordinator, error)
func DescribeCompletionInnerCleanup(context.Context, CompletionInnerCleanupHandle) (CompletionInnerCleanupDescriptorV1, error)
func OpenCompletionInnerCleanupCapability(context.Context, CompletionInnerCleanupHandle) (c12cleanup.Capability, error)
func BindCompletionValidationChildRuntime(context.Context, AuthenticatedCompletionRoleStateCoordinator, CompletionInnerCleanupHandle) (CompletionValidationChildRuntime, error)
func RunCompletionValidationChild(context.Context, CompletionValidationChildRuntime, CompletionValidationChildRequest) (CompletionValidationChildResult, error)
func ConsumeCompletionValidationChildResult(CompletionValidationChildResult) (CompletionValidationChildResultView, error)
func OpenInheritedCompletionValidationChildSession(context.Context) (CompletionValidationChildInheritedSession, error)
func InspectInheritedCompletionValidationChildSession(CompletionValidationChildInheritedSession) (CompletionValidationChildInheritedView, error)
func ReadInheritedCompletionValidationEvidence(CompletionValidationChildInheritedSession) (CompletionValidationChildEvidenceView, error)
func ReadInheritedCompletionValidationTrackedTree(CompletionValidationChildInheritedSession) (CompletionValidationChildTrackedTreeReader, error)
func ExchangeInheritedCompletionTrustedTime(context.Context, CompletionValidationChildInheritedSession, []byte) ([]byte, error)
```

Every cleanup descriptor is a non-authorizing, bounded projection of the already authenticated guard-selected owner. The role is fixed by the describing facade rather than accepted as caller data；all five staged-diagnostic、Windows、platform、authority and completion descriptors carry the selected nonzero `LaunchGeneration` and `MachineSealLocatorDigest`. The c12evidence facade strict-decodes the authenticated capsule, derives the complete `CleanupOnlyPolicy` from that envelope, requires its embedded role to equal the fixed facade role and requires its embedded launch generation、machine-seal locator and capsule digest to exact-match the descriptor/selected successor, and exact-recomputes `CleanupPolicyBindingDigest`. It then passes only the matching sealed successor/handle to its one-argument `OpenFixed*CleanupCapability` method；no c12evidence API receives or constructs an adapter、adapter slice or `CleanupInspectorDeleterSet`. In the same guarded call, runnerprofile rereads the current phase、epoch、full state digest、slot、capsule and policy binding, rejects a pre-CAS candidate or stale post-resume/post-terminal descriptor, pathlessly reopens the successor-private installed/provider/OS capabilities, constructs one fixed production adapter bound to each authenticated declaration, calls leaf `BindCleanupInspectorDeleterSet`, and only then calls the raw leaf Open with that sealed set. It marks that successor/handle's cleanup-open claim and every reconstructed runtime capability consumed on every return, so a descriptor copy is never selection or runtime authority. No descriptor field is independently authorizing and no caller may supply or default a policy member.

All three success-only `windows|platform|authority` post-clean targets retain a role-specific publication-input projection directly inside the authenticated A/B role-state cell before the final cleanup target CAS. `FixedPostCleanupPublicationInputsDigest` is the sole framing implementation and computes exactly `SHA-256(ASCII("TALENRO-C12-POST-CLEANUP-PUBLICATION-INPUTS-V1") || 0x00 || uint8(len(RoleASCII)) || RoleASCII || uint8(member_count) || Σ(uint64be(len(member_i)) || member_i))`, with no terminator or padding. `RoleASCII` is exactly `windows_scopes|platform|authority`; member count/order is respectively `3:{pre-clean result,scan receipt,supply-chain bundle}`、`4:{unsigned resume manifest,workload result,scan receipt,supply-chain bundle}`、`3:{workload result,scan receipt,supply-chain bundle}`. It rejects an unknown role、wrong count/order、empty member、length overflow、an individual/aggregate bound violation or a role byte length exceeding 255；all lengths are unsigned big-endian and the aggregate cap is checked before allocation/hash. The full bytes、this exact framing and digest are covered by the selected full state digest. At the success-origin CAS itself—not later at cleanup finalization—the unique c12evidence role facade supplies the already strict-schema/semantic-validated canonical result、receipt and bundle byte projections (platform transfers the same three through its sealed workload-result token). Runnerprofile deliberately imports no P09/artifactscan semantic schema：it checks only nonempty generic canonical-JCS byte form、the frozen individual/aggregate bounds、role-specific tuple arity/order and the domain digest, defensively copies the exact tuple into the inactive success-cleanup state cell, file/directory-fsyncs it and only then selects+rereads success origin. Every later cleanup-finalization pending and fixed post-clean target candidate copies those exact members byte-for-byte；each inactive cell is durable before its guard CAS. Pending and post-clean recovery preserve the projection unchanged through every leaf/key/tombstone and attestation/sender seam until the final committed-sender generation CAS consumes it；a different byte with a recomputed unrelated field, digest-only target or missing payload rejects. The aggregate strict-canonical projection is nonempty、nonsecret and at most `MaxFixedPostCleanupPublicationInputsBytes`：Windows stores the exact canonical pre-clean result plus scan-receipt and supply-chain-bundle projections；platform stores the canonical unsigned resume manifest、canonical workload result and those receipt/bundle projections；authority stores the canonical workload result and those receipt/bundle projections. The 5-MiB cap accommodates the already frozen 64-KiB receipt/result and 4-MiB bundle bounds without accepting an unbounded state value. Abort/ordinary targets contain zero publication inputs. Each `InspectFixed*PostCleanupPublication` first rereads the exact current post-clean phase/epoch/full digest and validates member order/generic canonical form/individual bounds/aggregate cap, recomputes the same function in constant time and returns only bounded defensive copies plus that digest；a stale/cross-role/missing/oversize/noncanonical/member-spliced view rejects. `TestNativeRunnerProfileStrictFiveRoleTokens` independently freezes three literal known-answer vectors plus 255/256 and 65535/65536 member-length transitions、wrong count/reorder/empty/over-cap cases, so a varint、little-endian、omitted role/count or another package-local implementation fails. The fixed state cell—not a caller、process heap、path or post-target file—is the cold-recovery source, so a fresh parent deterministically reconstructs the typed unsigned nested/outer projections and supplies their canonical bytes only inside c12evidence to the raw role-bound attestors. P09 high-level Publish facades accept only their opaque post-clean wrapper and no evidence struct or byte slice. After exact committed-sender inspection the final state CAS can consume the successor and select terminal/next without deleting an input file；the old cell becomes unselected and has no authority. The three existing `TestNativeRunner*PostCleanupAttestationResponseLossNeverResigns` roots restart a genuinely fresh process after the success-origin CAS response is lost、after the cleanup target CAS and before the first attestation, then before/after each attestation intent、hardware result、result durability and response delivery；they require byte-identical reconstruction solely from the selected state, at-most-once hardware use and zero producer/provider/workload/semantic-validator rerun.

The Windows final-scopes parent no longer advances the role generation before evidence publication. `CompleteFixedWindowsScopesToCleanup` and `AbortFixedWindowsScopesToCleanup` race through one guarded origin CAS. Success is callable only inside `internal/c12evidence/runnerhelper.go` from `CompleteWindowsFinalScopesToTerminal` after its private `WindowsFinalScopesValidatedResult` has consumed the same authenticated session、two sealed unsigned scope facts and an artifactscan-sealed `FinalClosedSetProjection`. The high-level command therefore supplies no bytes and cannot implement the result；the facade alone unwraps and passes exactly three defensive canonical projections to the raw bridge in fixed order：pre-clean result (≤64 KiB)、scan receipt (≤64 KiB) and supply-chain bundle (≤4 MiB), with aggregate ≤`MaxFixedPostCleanupPublicationInputsBytes`；the success-origin CAS freezes all three exact byte slices plus both unsigned scope fact sets and transfers the two scope attestors plus channel attestor into inaccessible success-cleanup state. Abort accepts no reason/bytes and destroys all three attestors. Active-parent crash before the success CAS defaults atomically to abort cleanup. `InspectFixedWindowsRecoveryDispatch` is the sole non-authorizing parent router：after authenticating the guard/cell/reservation it closes an active crash to abort and returns exactly `cleanup_only|ordinary_finalization_pending|post_cleanup_publication`; unknown/corrupt/reservation-mismatched state returns no route. The first route may recover a cleanup successor, the second calls only c12evidence's zero-capability `RecoverCleanedOwnership`, and the third may recover only the post-clean successor. Each target atomically rechecks the same phase before doing work, so an error never authorizes probing or falling through to another route and an ordinary pending state cannot remint cleanup authority. The first origin wins, same-origin replay is exact-idempotent and opposite origin conflicts. Inspect/Describe expose only defensive bounded projections, with the pre-clean result present/nonzero exactly for success. `TestWindowsRawRunnerprofileCallsOwnedOnlyByFinalScopesFacade` scans all module-local production files across build tags and requires every role-specific raw Windows complete/abort/dispatch/recover/inspect/`DescribeFixedWindowsCleanup|OpenFixedWindowsCleanupCapability`/receipt-finalize/post-clean attest/sender/finalize selector to occur only in `internal/c12evidence/runnerhelper.go`; the shared raw ordinary-pending recovery occurs only in `internal/c12evidence/cleanup.go`, while command packages and alternate wrappers have zero references.

Both origins obtain the same leaf cleanup capability only when the fixed c12evidence adapter validates the descriptor and then calls guarded `OpenFixedWindowsCleanupCapability` on that same successor. Success passes only the assignment-identical receipt to `FinalizeFixedWindowsCleanupToPublication`, which selects `windows_cleanup_finalization_pending` with the sole post-clean target, recovers/commits the leaf, proves the key absent, unlinks+parent-directory-fsyncs and verifies the tombstone absent while pending, and only then performs the final CAS into the same-generation successor. Abort uses only the ordinary receipt-bound target；the two finalizers reject the opposite origin. Selector-free recovery converges only the stored receipt/target. The successor is the sole route to two scope attestors and sender；each attestor durably records intent/result around at-most-once invocation. Only after both signatures and committed sender inspection may finalization advance terminal-inactive/next. Cleanup/sign/channel failure cannot publish or open a generation, and target selection has no later filesystem work.

`BeginFixedPlatformResumeTransition` is frozen in Task 2 so Plan 09 never adds a runnerprofile/launcher dependency after B10. It accepts only the actual Phase-A `platform` session already claimed by the fixed `run-platform-phase-a` helper and returns a one-use sink bound to that run/generation、fixed recovery slot、cleanup capsule/WAL reservation、the machine-local resume-manifest attestor and installed executable/provider/session projections. `SealAndSelectFixedPlatformResumeManifest` accepts only c12evidence's bounded canonical **unsigned** manifest projection. It recomputes its digest, signs the fixed domain with the sink-bound attestor, constructs and authenticates the private sealed envelope, stages/file+directory-fsyncs the fixed candidate, binds the envelope identity/digest and every session projection into the inactive role cell, advances+rereads the guard and consumes the sink on every return. It returns only the resulting nonzero envelope digest；there is no caller sealer、signature、digest、path、slot、provider or role input, and no command or test helper can mint a production envelope. `RecoverFixedPlatformResumeSuccessor` authenticates the envelope internally and is the sole pathless mint after reboot, callable only by the fixed `resume-platform-phase-b` child with the matching inherited launch reservation/new boot identity；it never calls `OpenInstalledNativeRunnerSession`. Inspect returns one defensive canonical unsigned-manifest copy, the authenticated envelope digest and only the exact binding digests above. If semantic/time/provider validation cannot authorize resume, `DowngradeFixedPlatformResumeSuccessorToCleanup` linearly consumes the resume successor and guard-selects a distinct cleanup-only variant; the cleanup successor has no provider/workload/signing methods. A crash after downgrade recovers only through `RecoverFixedPlatformResumeCleanupSuccessor`. Cross-role/run/generation/boot replay、a second sink/successor open、resume after downgrade or cleanup-to-resume promotion rejects. `InspectFixedPlatformResumeRecoveryDispatch` is a non-authorizing, reservation-bound, fail-closed discriminator over exactly `validate_resume_fresh|validate_resume_recovery|workload_ready|cleanup_only|ordinary_finalization_pending|post_cleanup_publication`；it authenticates the selected guard/cell plus exchange-intent phase before returning and never exposes a selector/capability. Fresh is legal under the original Phase-B reservation **or an absence-proved replacement reservation** exactly when no trusted-time exchange intent was ever written；because intent is durable before any provider call, the replacement route also proves provider call count zero and may safely begin the one exchange. Once intent exists—even if provider response or result durability is incomplete—only recovery is legal；it can recover an identical durable response or downgrade to cleanup and can never refetch. Promotion selects `workload_ready` atomically. The ordinary-pending target calls only c12evidence `RecoverCleanedOwnership` and returns no session/capability. Every target atomically rechecks the same phase, so P09 never chooses by probing one recovery error and falling through to another、treats absent intent as failed recovery or remints cleanup authority after the first finalization CAS；unknown/corrupt/reservation-mismatched state returns no route. P09's opaque manifest builder、fresh/resume recovery openers and cleanup facade are the sole c12evidence consumers of these sealed capabilities. Every production reference/call to the role-specific raw platform bridge is statically confined to `internal/c12evidence/platform.go`; the shared raw ordinary-pending and trusted-time bridges are confined to `internal/c12evidence/cleanup.go` and `internal/c12evidence/trustedtime.go` respectively. Commands、runner helpers and alternate packages have zero references, and P09's exact AST/JSON gate fails on a missing/renamed/skipped ownership test. Tests `TestNativeRunnerPlatformResumeTransitionIsGuardedAndOneUse`, `TestNativeRunnerPlatformResumeManifestUsesSinkBoundAttestor`, `TestNativeRunnerPlatformPhaseBUsesPathlessSuccessorNotOpen`, `TestNativeRunnerPlatformValidationDispatchSeparatesFreshFromAmbiguousExchange` and `TestNativeRunnerPlatformResumeDowngradeCannotRegainWorkloadAuthority` cover every candidate/sign/fsync/guard/reboot/before-intent replacement/intent/response/promotion/downgrade/finalization seam.

`BindNativeTrustedTimeRuntime` is the sole production provenance bridge for the fresh Linux `platform_phase_a|authority_evidence` purposes and enforces the closed role map `platform→platform_phase_a`, `authority→authority_evidence`; it carves one one-use transport from the authenticated session without exposing or consuming the session's other capabilities. `BindPlatformResumeTrustedTimeRuntime` is the only `platform_phase_b` mint and binds the same source/profile/native-client/provider projections to the sealed resume successor. Both return a sealed view and `ExchangeNativeTrustedTime` accepts only bounded canonical request bytes, first records one successor-bound unique request intent, consumes the runtime on every return, measures monotonic elapsed time inside the native runtime and yields a sealed transcript result only after the bounded response、its digest、the finite positive elapsed duration and a runtime/request/response-bound measurement digest are durably recorded. Negative、zero、overflowed、over-cap or provenance-spliced durations reject. For Phase B, `RecoverPlatformResumeTrustedTimeExchange` accepts only that same resume successor and returns the already-recorded exact transcript and identical monotonic measurement after response loss；it makes no provider/network call, cannot choose request bytes and rejects an absent/incomplete intent so the caller can only downgrade to cleanup. It never retries an ambiguous external operation or creates a second nonce. P09 c12evidence wraps the runtime to implement its private native-identity/provider markers and reuses `LoadTrustedTimeProviderProfile`/`NewTrustedTimeClient`; no cmd package can implement those markers or receive a handle/endpoint/credential. It Inspect-validates the sealed response and then either consumes the exchange result for Phase A/authority or passes that same original-or-recovered result to `PromoteFixedPlatformResumeSuccessor` after the full Phase-B time/manifest gate. Promotion exact-matches result/runtime/successor and consumes both, so a semantically rejected response can only downgrade to cleanup and a second exchange cannot promote. `TestNativeRunnerPlatformTrustedTimeExchangeCrashRecoveryNeverRefetches` injects before/after intent fsync、provider return、response fsync and result delivery, proving recovery returns the unique durable response plus byte-identical elapsed/provenance or fails closed to cleanup with the provider call count unchanged；P09 can therefore reapply the same fixed maximum-round-trip gate without recreating an impossible same-process stopwatch proof.

The promoted `PlatformResumeWorkloadSuccessor` exposes no raw provider/NV/WAL/root handle. `RunFixedPlatformResumeWorkload` owns the exact declared-resource open plus fixed Phase-B counter/provider/host-conformance child flow and returns a sealed bounded token carrying the canonical result (≤64 KiB)、scan-receipt projection (≤64 KiB) and supply-chain-bundle projection (≤4 MiB), all bound to the manifest/run/generation/provider session and aggregate-capped by `MaxFixedPostCleanupPublicationInputsBytes`; P09 reads defensive copies and their digests only through `InspectFixedPlatformResumeWorkloadResult`, applies the owning strict semantic validators and exact-cross-checks their repeated scanner/bundle facts before success completion. Semantically successful workload transfers the matching result and successor together through `CompleteFixedPlatformResumeWorkloadToCleanup`; any provider/semantic failure uses `AbortFixedPlatformResumeWorkloadToCleanup`. Both select a no-workload cleanup-only variant, while a crash in the active workload recovers only through `RecoverFixedPlatformResumeWorkloadSuccessor`. The success-origin CAS copies all three exact projections into the selected cleanup record and its full state digest；abort/downgrade carries none. That record durably carries the closed origin (`complete=>success`; `abort|downgrade=>abort`) and, only for success, the exact bounded canonical pre-clean workload result plus digest. `InspectFixedPlatformResumeCleanupSuccessor` returns that origin/result and the original resume projection in one sealed defensive view, so fresh-process recovery never guesses a finalizer or loses the evidence input；unknown origin、abort with nonzero result or success with absent/mismatched result fails closed. `DescribeFixedPlatformResumeCleanup` mirrors the completion-inner descriptor boundary：it returns a defensive canonical cleanup-capsule envelope plus nonzero policy/run/generation/successor/slot bindings and the same conditionally exact workload-result digest, but no path/key/deleter/create/sign/send authority. P09's cleanup adapter strict-decodes that envelope and derives/recomputes the exact `CleanupOnlyPolicy`, but only guarded `OpenFixedPlatformResumeCleanupCapability` may reread selection and privately open the existing leaf capability.

After `ExecuteCleanup`, c12evidence prepares one sealed leaf receipt bound to capability/WAL/absence/key. Only success-origin may pass it to `FinalizeFixedPlatformResumeCleanupToPublication`, which first selects `platform_cleanup_finalization_pending(receipt,platform_post_cleanup_publication)`. It then leaf Recover/Commits, proves the key absent, retires the tombstone with unlink+parent-directory fsync+exact absence while pending, and finally selects the ready post-clean state. Before the first CAS, Prepare replays the same receipt；after it, selector-free recovery reads only the stored view/target and repeats any incomplete leaf/key/tombstone step before final CAS. The successor binds manifest/workload/absence receipt plus the three platform attestors and sender slot, and has no post-target filesystem work.

An abort/downgrade-origin successor fails the special origin/binding check without consuming the prepared receipt and must use c12evidence `FinalizeCleanedOwnership(ctx,cap,proof)`. That wrapper prepares the same leaf receipt and calls runnerprofile `FinalizeFixedCleanedOwnership`, which analogously selects an origin-bound `cleanup_finalization_pending` whose sole target is ordinary terminal-inactive, recovers/commits the receipt and then advances the generation. `RecoverFixedCleanedOwnership` is selector-free and completes only that pending target. The first selected origin-specific pending state wins；a receipt cannot cross from ordinary to publication or vice versa, and failure-origin cleanup can never mint publication authority. Inspect returns only the bounded digests/scalars above. `AttestFixedPlatformNestedEvidence` and `AttestFixedPlatformScopeEvidence` each accept the corresponding c12evidence canonical unsigned projection and first durably record its exact input/domain intent. They invoke the retained one-shot hardware attestor at most once, self-verify and durably record the bounded public result before returning it；response-loss recovery returns only the byte-identical recorded result, never re-signs, and different bytes conflict. Wrong order、second use、domain crossover or projection not bound to the retained manifest/result/cleanup facts rejects. Task 7 later adds the post-clean sender binder/finalizer：only after both attestations are consumed may it return the profile-fixed sender, cross-persist the sole nonce/envelope/member identities and advance the role generation after exact committed-state inspection. Tests owned here include `TestNativeTrustedTimeRuntimeRolePurposeAndProvenanceAreClosed`, `TestNativeTrustedTimeRuntimeExchangeIsBoundedOneUse`, `TestNativeRunnerPlatformPromotionRequiresMatchingValidatedExchange`, `TestNativeRunnerPlatformWorkloadFacadeHasNoRawHandle`, `TestNativeRunnerPlatformCleanupDescriptorIsBoundedAndSealed`, `TestNativeRunnerPlatformSuccessCleanupMintsPublicationOnlyAfterAbsence`, `TestNativeRunnerPlatformAbortCleanupCannotMintPublication`, `TestNativeRunnerPlatformCleanupFinalizationReceiptCrashSeams`, `TestNativeRunnerPlatformPostCleanupAttestationResponseLossNeverResigns` and `TestNativeRunnerCleanupFinalizationFirstSelectedOriginWins`; Task 7 owns the exact sender-state recovery test.

For the non-resumable `authority` run, `CompleteFixedAuthorityWorkloadToCleanup` and `AbortFixedAuthorityWorkloadToCleanup` race through one guarded origin CAS. The success method is callable only by P09 c12evidence's fixed parent with its package-private validated pre-clean workload-result token；the package-private token supplies exactly its canonical workload result (≤64 KiB)、scan-receipt projection (≤64 KiB) and supply-chain-bundle projection (≤4 MiB) to the sole three-slice raw bridge. Runnerprofile generic-canonical-checks and defensively copies the fixed-order aggregate, records `PublicationInputsDigest` in the success-origin full state and transfers the three authority nested/scope/channel attestors into an inaccessible success-cleanup state before returning. The projection binds every authoritative matrix、receipt/bundle、trusted-time and resource-plan fact available before cleanup but deliberately contains no claimed terminal cleanup result or evidence signature；those are derived only from the later leaf receipt/post-clean view. The abort method accepts no reason、bytes、status or caller origin, records `abort` and irreversibly destroys those attestors. The first selected origin wins；same-origin replay is exact-idempotent and the opposite call conflicts. If the active parent crashes before the success CAS, `InspectFixedAuthorityRecoveryDispatch` first authenticates the guard/cell/reservation and atomically closes it to the already declared abort cleanup；otherwise it returns exactly `cleanup_only|ordinary_finalization_pending|post_cleanup_publication`. Unknown/corrupt/reservation-mismatched state returns no route. The ordinary-pending target calls only c12evidence `RecoverCleanedOwnership` and returns no session/capability. Each matching pathless recovery atomically rechecks that phase before doing work, so the fixed parent can neither rerun workload、choose success、probe by error、fall through nor remint cleanup authority after the first finalization CAS. `RecoverFixedAuthorityCleanupSuccessor` returns only the selected cleanup authority. Its Inspect/Describe views are bounded and the canonical workload result is present/nonzero only for success.

Both authority origins obtain the same leaf cleanup capability only after P09's descriptor adapter validates the sealed view and calls guarded `OpenFixedAuthorityCleanupCapability` on that same successor. After Execute+Prepare, success alone may call `FinalizeFixedAuthorityCleanupToPublication`, which selects pending with the assignment-identical receipt and sole same-generation post-clean target, leaf Recover/Commits, proves key absent, unlinks/fsyncs/verifies the tombstone absent while pending, then performs the final target CAS. Abort uses ordinary terminal-inactive/next；both finalizers reject the opposite origin. Recovery reads only stored receipt/target and finishes leaf/key/tombstone work before CAS, never accepts a selector. Every authority attestor/sender rejects before the successor, and target selection has no later filesystem work.

`AttestFixedAuthorityNestedEvidence` and `AttestFixedAuthorityScopeEvidence` accept only the corresponding P09 canonical unsigned projection bound to the retained pre-clean workload result plus terminal cleanup proof. Each fixed one-shot attestor durably records the exact input/domain intent before hardware invocation, invokes hardware at most once, self-verifies and durably records the result before returning；on response loss it returns only the byte-identical recorded signature；different bytes conflict and recovery never re-signs. Task 7 later adds the authority post-clean sender binder/finalizer and closes the direct generic binder to Windows. Every production reference/call to the raw authority complete/abort/dispatch/recover/inspect/describe/finalize/attest/sender bridge is statically confined to `internal/c12evidence/authority.go`; commands、runner helpers and alternate packages have zero references. P09's exact AST/JSON gate owns that rule. B10 exercises these neutral algorithms only with package-private synthetic role/session/profile projections and fixed non-authoritative attestator/channel fakes；it neither reads P09's future authority profile nor mints reusable authority evidence. Tests owned here are `TestNativeRunnerAuthoritySuccessCleanupMintsPublicationOnlyAfterAbsence`, `TestNativeRunnerAuthorityAbortCleanupCannotMintPublication`, `TestNativeRunnerAuthorityCleanupFailureCannotReachAttestors`, `TestNativeRunnerAuthorityActiveCrashDefaultsToAbortCleanup`, `TestNativeRunnerAuthorityRecoveryDispatchIsClosedAndGuarded`, `TestNativeRunnerAuthorityCleanupRecoveryIsSelectorFree`, `TestNativeRunnerAuthorityCleanupFinalizationReceiptCrashSeams`, `TestNativeRunnerAuthorityPostCleanupAttestationResponseLossNeverResigns` and `TestNativeRunnerAuthorityFinalizationFirstSelectedOriginWins`; Task 7 owns sender binding/recovery/final-generation tests.

`DescribeCompletionInnerCleanup` is a read-only, non-consuming bridge for the sole c12evidence adapter. It accepts only a currently authenticated inner handle and returns a bounded defensive capsule-envelope copy plus nonzero capsule/policy/launch-generation/machine-seal/coordinator bindings；the private seal rejects a zero/fabricated descriptor. It returns no path、WAL key、provider handle、deleter、adapter set or create authority. `internal/c12evidence.OpenCompletionInnerCleanupForExactCleanup` is the only production consumer：it calls Describe, strict-decodes the existing capsule schema, recomputes every descriptor/policy binding, and then passes only the same handle to one-argument guarded `OpenCompletionInnerCleanupCapability(ctx,handle)`. That runnerprofile method rereads the current nested selection, reconstructs the fixed declaration-bound adapters from handle-private runtime capabilities, leaf-binds the set and opens the capability；no policy/set crosses the package boundary. The dependency remains `c12evidence -> c12runnerprofile`; runnerprofile imports no c12evidence type. `TestCompletionInnerCleanupDescriptorIsBoundedAndSealed` and `TestOpenCompletionInnerCleanupForExactCleanupIsSoleDescriptorConsumer` freeze the exact `{DescribeCompletionInnerCleanup,OpenCompletionInnerCleanupCapability}` pair in `internal/c12evidence/cleanup.go` and statically reject any other production caller、unbounded envelope、zero/fabricated seal or descriptor/capability escape from `internal/c12completionpublication/operator.go`.

Completion inner cleanup never uses the ordinary role finalizer. The P09 opaque session exclusively owns `{CompletionInnerCleanupHandle,c12cleanup.Capability}`；the handle is returned only after its private selected WAL bootstrap/actual is durable, while the capability is delete-only. After `ExecuteCleanup`, leaf `PrepareFinalization` yields the assignment-identical receipt. `FinalizeInnerCleanup`、`FinalizeRetainedRevalidation` and both abort variants exact-match handle/receipt/run/generation/terminal-WAL/absence/key bindings, then first guard-select `finalization_pending` carrying the complete receipt view、authenticated tombstone binding and one fixed target. While that pending state remains selected they recover/commit the leaf receipt, destroy/verify the key, unlink the now-powerless tombstone with parent-directory fsync, verify its exact absence, and **only then** perform+reread the second guarded target CAS. There is no post-target filesystem retirement, so create `complete` can remain actionless. Fresh recovery never probes by error：the five-way router maps only genuine pre-manifest `none` to `begin_required`, selected pre-manifest states to `session_validation_recovery`, create capsule plus set-first/outer manifest intent to `session_cleanup_recovery`, pending receipt state to `finalization_recovery`, and create target-completed `none` plus manifest-intent-or-later to `complete`. Begin/Recover return a nonzero WAL-ready handle only after their fixed intent/restart/capsule/WAL loop；validation-required alone may bind the runtime/Stage, cleanup-only may only exact-clean/finalize after any private state-only manifest barrier. `RecoverInnerFinalization` is selector-free and returns no handle/capability/receipt；from stored values it idempotently closes leaf Commit、key absence、tombstone unlink+directory-fsync/absence and finally the second target CAS. A crash at unlink、directory fsync or CAS response loss therefore remains in the selected pending route until all pre-target retirement is proven. No target accepts a replacement receipt、outcome or selector；a retained target ends the active owner and response loss uses only outcome audit. Tests `TestCompletionInnerFinalizersRequireExactLeafReceipt`, `TestCompletionInnerFinalizationReceiptCrashSeams` and `TestCompletionRetainedOutcomeConsumesSameCleanupReceipt` cover create、completed、pre-time abort and with-head abort at every pending/leaf/key/unlink/directory-fsync/target-CAS seam.

`BindCompletionValidationChildRuntime` is the B10-frozen runnerprofile bridge from a fresh/pathlessly recovered coordinator plus its exact validation-required inner handle to the installed private-child/provider/process capability. Before it exists, it authenticates install、mode/phase、run/generation、coordinator/inner、runner/provider、tree/object-database and execution-reservation bindings and rechecks that the handle's private selected WAL has durable bootstrap/actual；cleanup-only handles reject. The concrete runtime borrows that handle's sole recorder and is the only completion create-side consumer of it. `RunCompletionValidationChild` accepts only the two closed purposes plus nonzero composite-derived input/view digests；before each declared work-root/provider-session/pipe/private-process create it appends+file/parent-fsyncs the exact intent, and immediately after create it records the observed identity actual before any dependent use. Missing/failed WAL append prevents the resource operation. It then generates the nonce, opens the installed child, places it in the bound Job, transfers only fixed inherited handles, waits/reaps and consumes the runtime on every return. It returns no recorder、path、numeric handle、credential、client、request bytes or spawn primitive. The result path length-bounds/hashes the response, authenticates PID/start、executable、profile/provider/tree/request/nonce/pipe facts and returns one immutable sealed result；`ConsumeCompletionValidationChildResult` consumes it once and returns only defensive non-authoritative response/evidence/retained-manifest copies plus a sealed read-only tree reader. `internal/c12runnerprofile` neither imports c12evidence nor decodes its DTO；only `internal/c12completionpublication/operator.go` calls this parent bridge and then strict-decodes through c12evidence. The `!windows` adapter and any runtime call without the WAL-ready validation handle reject before provider/process access.

Inside the exact installed direct child, zero-argument `OpenInheritedCompletionValidationChildSession` authenticates the inherited parent/runtime binding、child PID/start/executable、fixed pipe and provider/evidence/tree-reader handle table and returns one sealed child-side session；direct/go-run/unbound invocation rejects before handle access. Its read-only Inspect returns only the non-authoritative digest/nonce view above. `ReadInheritedCompletionValidationEvidence` reads only the runtime-bound no-follow evidence handles, rechecks the materialized binding, role exact-cover and existing per-member/aggregate caps, and returns defensive copies of the ordered eight canonical members plus the retained manifest sibling only for revalidation；create requires absent/zero retained-manifest fields, while revalidate requires nonzero bytes/digest bound to the selected immutable payload. It returns no root/path/open handle and cannot read a ninth caller-selected member. `ReadInheritedCompletionValidationTrackedTree` returns one sealed read-only reader whose Inspect binds the typed committed locator、canonical tracked-tree/inventory digest and exact snapshot/object-database identities to the same runtime. Its inventory method returns the complete bounded canonical relative-path/mode/blob-digest inventory needed to independently recompute `TrackedTreeDigest`; its blob method accepts only the closed role plus a profile-bound ordinal for the spec-set/members、catalog、toolchain、receipt-declared obligation documents、four completion contract documents and two provider/runner profiles. It returns defensive bytes and canonical inventory metadata, never an absolute/root path、FD/HANDLE、arbitrary token or write/create capability. Unknown/out-of-range/reused roles、inventory/blob splice or a binding mismatch rejects. `ExchangeInheritedCompletionTrustedTime` is a bounded one-use transport over that session's already inherited fixed provider handle；it accepts no endpoint、credential、profile、numeric handle or transport option and returns only bounded response bytes. Only the c12evidence-owned inherited completion adapter may call these child-side functions；the child `main` package cannot implement either sealed reader or a c12evidence provenance marker. Static tests `TestCompletionValidationChildRuntimeBindsCoordinatorInnerAndInstalledProjection`, `TestCompletionValidationChildRuntimeResultIsBoundedSealedAndSingleUse`, `TestCompletionValidationChildTrackedTreeReaderIsPathlessAndExact`, `TestCompletionValidationChildRuntimeHasNoEvidenceSemanticDependency` and `TestCompletionValidationChildRuntimeOwnedOnlyByCompositePackage` freeze both sides of this route and reject zero/fabricated/mutated result reuse、double-Bind、second exchange、evidence/member/manifest/tree splice、direct launcher/command access or alternate child spawn.

There is deliberately no exported store, raw `CompareAndSwap`, canonical-business byte getter/setter or caller-mintable mutation type. `AuthenticatedCompletionRoleStateCoordinator` is the only cross-package mutable completion-state bridge and exposes only the fixed semantic methods above；each internally mints and consumes one sealed permit after deriving the complete next business projection from the guard-selected state and its retained sealed capabilities. `BeginPublishPending` alone performs `active_unstarted→publish_pending` and derives every stable root/manifest/marker/set identity from the session. Ordinary inner semantic permits change only the nested cleanup substate, and retained permits read the immutable payload and change only their active/cleanup envelope. The sole fixed exception is not a caller-selectable inner mutation：before the create `session_cleanup_recovery` target returns a handle, it invokes package-private `recoverCompletionManifestIntentBarrier` with the already bound concrete set/coordinator. That function consumes only the set-bound outer publication permit for the previously persisted exact set-first old→`manifest_intent` transition, so it may update `PublishBindingDigest` to that one stored projection but cannot choose bytes、set、progress、filesystem work or any later outer delta. `InspectInnerRecoveryDispatch` is non-authorizing and its five cases each recheck the exact owner/phase/reservation/outer-progress/set-pending combination before work；`BeginInnerCleanup` is the only `begin_required` target and `RecoverInnerFinalization` is the only no-handle route across leaf-commit、key-destroy and second-CAS response loss. `InspectFixedCompletionRetainedFinalization` is deliberately not a coordinator or recovery surface：it yields only an opaque, bounded terminal audit receipt under the fixed create-mode terminal-audit reservation described below. Task 7 extends the same concrete coordinator with the opaque channel begin/adopt/publication-sink/ACK/finalize-ready/finalization operations frozen below. Every permit fixes current epoch/full/publish digests and the exact next semantic delta, is coordinator/set/run/generation bound and is consumed whether the guarded mutation succeeds or returns an exact already-applied result. The concrete coordinator and permit types live in `internal/c12runnerprofile/completion_state.go` plus `channel.go`; P09 `internal/c12completionpublication/operator.go` is the sole production holder/caller of the coordinator and every parent-side raw completion recovery/runtime/audit API. `internal/c12evidence/completion.go` and runner-helper have zero such references. The only separate raw completion exception set is exactly `{DescribeCompletionInnerCleanup,OpenCompletionInnerCleanupCapability}`, both called solely by the already-frozen descriptor adapter in `internal/c12evidence/cleanup.go`; they return the defensive descriptor and guarded delete-only capability respectively, never a coordinator、policy、adapter or set. The set-bound publication sink's typed root/materializer/manifest/marker methods mint and consume each closed progress value internally and return only success/error. P09's exact all-build-tag AST/type gate and compile-fail tests reject a raw/generic mutation method, direct state bytes, a caller-defined permit/progress value, any raw bridge caller outside those exact owners and any transition outside these semantic operations.

`C12NativeRunnerProfileV1` has exact schema `talenro-c12-native-runner-profile/v1`. It has four final roles plus one nonpublishing diagnostic role, each with one non-overridable tracked token: `windows_scopes` → `testdata/c12/windows/runner-profile.v1.json`, `platform` → `testdata/c12/platform/runner-profile.v1.json`, `authority` → `testdata/c12/authority/runner-profile.v1.json`, `completion` → `testdata/c12/completion/runner-profile.v1.json`, and `staged_diagnostic` → `testdata/c12/diagnostic/runner-profile.v1.json`. `BootstrapLauncherCatalogRole` is mandatory and closed by OS：both Windows roles use `c12_runner_launcher_windows_amd64` and all three Linux roles use `c12_runner_launcher_linux_amd64`; it is distinct from `HelperCatalogRole` and never appears in `InstallReceiptPolicy.RequiredReceiptRoles`. B10 creates/relocks/installs the `windows_scopes` and `staged_diagnostic` role generations after validating their already machine-base-installed launchers；Plan 09 owns only the other three final profile instances.

`ChannelBindings` is role-dependent and closed: `windows_scopes` has exactly one `sender/windows_scopes` binding, `platform` exactly one `sender/platform_scope`, `authority` exactly one `sender/authority_scope`, `completion` exactly three receiver bindings in fixed order `windows_scopes`, `platform_scope`, `authority_scope`, and `staged_diagnostic` exactly zero bindings. Every final-role binding uses its fixed symbolic `SlotPurpose` and `SequencePolicy="next"`. The corresponding actual namespace/slot identity/next sequence remain only in the sealed machine bootstrap/session and must exact-cover this ordered policy projection. Thus completion can adopt three expected slots without caller kinds/slots/sequences or enumeration, each producer gets only its sender, and diagnostic code cannot open or inspect any channel.

The Windows and diagnostic instances are policy-only/digest-free: neither contains an absolute machine path, repo commit/tree, installed executable digest, evidence/output digest, machine namespace/slot identity/sequence, channel transaction nonce or secret. The diagnostic instance additionally requires `AllowedTrackedScriptTokens=[]`, exact Linux-only `RequiredToolRoles` for the static runner helper, locked Git and reproducible Go/OCI build tools, `HelperCatalogRole="c12_runner_helper_linux_amd64"`, `ProviderCredentialPurpose="none"`, `EvidenceRootPurpose="nonpublishing"`, and zero `ChannelBindings`; it still requires the canonical object-database, owner-only external-root, recovery-slot and exact install-receipt policies. It contains no scope/channel/provider/result signer purpose and can never publish receipt, bundle, scope or evidence. Strict decoding rejects unknown/duplicate/null/defaulted fields, noncanonical order/path tokens, missing/extra/duplicate/reordered channel binding, any diagnostic script/provider/channel/publishing capability, another role's tool purpose and a role/token mismatch. `CanonicalNativeRunnerProfile` derives `SHA-256(ASCII("TALENRO-C12-NATIVE-RUNNER-PROFILE-V1") || 0x00 || JCS(profile))`; the digest is external and is never embedded back into the profile.

`C12NativeRunnerInstallReceiptV1` has exact schema `talenro-c12-native-runner-install-receipt/v1`; its unsigned projection excludes only `Attestation`, which is verified with the fixed machine-install authority over `TALENRO-C12-NATIVE-RUNNER-INSTALL-RECEIPT-V1\x00 || JCS(unsigned)`, and its complete envelope digest uses `TALENRO-C12-NATIVE-RUNNER-INSTALL-RECEIPT-ENVELOPE-V1\x00`. A receipt is machine-local and binds exactly one canonical absolute **role-generation native executable** path—runner helper、interpreter、Git/Docker/scanner/build tool or private completion child—plus catalog role, executable digest, OS/arch, no-follow regular-file identity and application-control policy；it never covers the machine-base launcher or a tracked snapshot script and is never tracked. `InstallReceiptPolicy.RequiredReceiptRoles` therefore contains only those role-generation native executable roles. Unknown/duplicate/null fields, relative/device/UNC/path-alias input, a launcher/script token/file presented as an install receipt, role/OS/arch/hash/file-identity replacement, stale install time or self-selected install authority rejects.

`C12FixedRunnerLauncherAttestationV1` has exact schema `talenro-c12-fixed-runner-launcher-attestation/v1`. It is produced only by the existing attested target-machine image/deployment baseline when that external OS package manager installs the reviewed fixed launcher, and its unsigned projection excludes only `Attestation`. The baseline authority signs `TALENRO-C12-FIXED-RUNNER-LAUNCHER-ATTESTATION-V1\x00 || JCS(unsigned)` and the envelope digest uses the matching `...-ENVELOPE-V1` domain. It exact-binds one of the two launcher catalog roles、fixed absolute path、PE/ELF digest、OS/arch、no-follow file identity、machine-image identity、application-control policy and the toolchain launcher projection. The two launchers are reproducibly built、SBOM/provenance scanned and frozen as native toolchain inputs before the B10 five-output relock；the B10 lock records both identities, and the B11 nine-output relock must revalidate them byte-identical rather than upgrade them. P09's final catalog exact-covers both roles without making the catalog a B10 dependency. This external base-install step is the explicit root prerequisite, not a hidden self-install：a host with absent/wrong launcher attestation cannot join B10/B11, and role provision commands cannot repair it. The launcher attestation is nonsecret and separate from every role guard/cell/install receipt；copying it to another machine、changing its path/digest/policy or substituting a helper fails preflight.

`VerifyFixedMachineStateGuardProvider` is the non-consuming production-provider preflight. Its compile-time OS adapter accepts no role/path/handle/provider/options, authenticates the fixed TPM/vTPM-backed compare-and-advance implementation、machine/application-control identity、non-resettable/nonexportable policy and absence of any file/registry/DPAPI-only fallback, and returns only `nil|error` without allocating a role guard or advancing state. The same preflight authenticates a distinct machine-base `talenro-c12-runner-launcher` installed by the OS package manager/application-control baseline at exactly `C:\Program Files\Talenro\C12\talenro-c12-runner-launcher.exe` or `/usr/libexec/talenro-c12/talenro-c12-runner-launcher`. This fixed launcher is a prerequisite/root-of-trust peer of the guard provider, not a role helper、install generation or `C12NativeRunnerInstallReceiptV1` subject；none of the five role provision modes can install、replace or retarget it, so first installation has no dependency on an already installed role helper.

`OpenFixedMachineInstallerSession` is the sole package-owned mint for `ProtectedMachineInstallerSession`：it accepts no inputs, re-runs that preflight, authenticates the short-lived OS machine-install authority、the exact fixed launcher/provision-mode process identity and one deployment-service-supplied nonexportable candidate-package handle, and returns one opaque nonserializable session that retains that handle. For the two B10 modes the handle is bound only to the committed B10 toolchain native-role projection、matching Windows/diagnostic profile projection and frozen launcher attestation—explicitly not to the future production catalog, which does not exist at B10. For the five B11 modes it additionally exact-binds the final catalog role projection and the final relock projection commit. A nil/fake interface、ordinary/versioned runner process、wrong fixed mode、expired/replayed installer authority、path/argv/environment package source or direct package call without the OS deployment context fails. There is no generic install command or caller-selected role/path；only the fixed launcher exposes five no-extra-argument protected modes `provision-windows-scopes-runner|provision-staged-diagnostic-runner|provision-platform-runner|provision-authority-runner|provision-completion-runner`, each compile-time bound to exactly one role and OS. `InstallFixedNativeRunnerProfile` linearly consumes the supplied installer session on every success or error return and internally defers destruction of both the session and its candidate handle before returning. Each mode therefore calls provider preflight → installer-session open → consuming install for its one role → non-consuming `VerifyFixedNativeRunnerInstallation` and returns bounded status only；there is no caller-visible destroy step or reusable session.

`RunFixedNativeRunnerMode` is the sole cross-package launcher bridge for the seven machine-native public modes. It authenticates that the caller process is the fixed base launcher at its attested identity, requires the literal argv to be exactly one matching zero-extra-argument mode and enforces the closed mode→role→OS map. Inside `internal/c12runnerprofile`—without returning a path, descriptor or capability—it authenticates the role guard、selected cell/install record/launcher projection, no-follow opens the selected versioned helper by receipt identity and creates that child initially suspended/stopped inside a role-private OS containment. Before the child can execute, one guarded whole-record mutation installs a private execution reservation containing only mode、reservation purpose、install-record digest、launcher PID/start、helper PID/start and containment identity；it does not change/consume the next launch generation or business phase. The child is resumed only after that reservation is selected and reread, consumes the inherited-only one-use binding plus matching reservation, and then alone calls the mode/phase-permitted `OpenInstalledNativeRunnerSession`, fixed active recovery API or read-only terminal-audit API. A direct package call from another process、unknown/wrong-OS mode、stale record/helper、candidate handle in a run mode、helper replacement or a child whose PID/start token does not match the binding fails before session/completion access. Thus `cmd/talenro-c12-runner-launcher` is a thin closed dispatcher and never parses the guard/store itself；there is no duplicated selector or launcher-owned active pointer.

Only one protected launcher invocation lock exists per role. On Windows the launcher pre-creates a dedicated `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE` Job and uses `PROC_THREAD_ATTRIBUTE_JOB_LIST` (or the audited equivalent atomic create-in-job primitive) so no helper process can exist outside that Job；it also creates the child suspended and keeps the sole controlling Job handle until exact wait/reap. On Linux it uses `clone3(CLONE_PIDFD|CLONE_INTO_CGROUP)` into a pre-created supervised transient cgroup/systemd scope with `KillMode=control-group`, or an equivalent parent/child gate in which the child sets `PDEATHSIG=SIGKILL`, proves cgroup attachment and ACKs before it can pass the stopped pre-exec barrier. No reservation is staged and no helper instruction executes until the containment-attached ACK；parent death before ACK kills the child without leaving an unrecorded process. The exact scope is then bound in the execution reservation and no descendant may escape it. If the launcher dies later, the whole bound helper tree is killed. Before any restart may replace or clear a stale reservation, the fixed bridge exact-inspects that recorded PID/start/Job-or-cgroup identity, kills only that bound containment if needed, waits/reaps and proves every member absent. For an active business phase it guarded-replaces the reservation before resuming a successor that uses only the same-generation recovery API and never Open. For an `idle` reservation abandoned before Open, or any terminal-inactive phase carrying an `exiting` reservation, it guarded-clears only that exact reservation after the absence proof without changing business state、payload or next generation；it cannot resume/reuse it. Any later fresh Open or terminal audit must create and guard-select its own distinct purpose-bound reservation. On normal terminal success the helper's terminal whole-record semantic mutation advances only the business phase/generation and marks its matching execution reservation `exiting`; it cannot remove that reservation or make the next invocation launchable while it is still alive. The launcher then exact-waits/reaps, proves the containment empty and performs one guarded mutation that removes only the `exiting` reservation without altering terminal business state/generation/payload；only after reread may it return zero or allow the next invocation. A pre-Open nonzero exit follows the same absence-before-clear rule without consuming a generation. If the launcher dies after helper terminalization but before reservation removal, the next launcher must prove exact absence and clear that same reservation before starting anything. Concurrent launchers、an unbound surviving helper/tree、reservation replay or recovery/next-run/audit before absence proof rejects. Launcher death before/after atomic create-in-containment/ACK、reservation fsync/guard advance/reread、helper Open、outer-state mutation、provider/private-child spawn、business terminal mutation and reservation clear can therefore leave recovery or audit work, but never an unrecorded child or two live parents with authority.

Each launcher build's operator help exact-enumerates only its OS-valid subset. Windows lists `run-windows-final-scopes|completion-consolidate-create|completion-revalidate` plus `provision-windows-scopes-runner|provision-completion-runner`; Linux lists `capture-staged-and-build-closed-set|run-platform-phase-a|resume-platform-phase-b|run-authority-final` plus `provision-staged-diagnostic-runner|provision-platform-runner|provision-authority-runner`. The union is seven run and five provision modes, but a wrong-OS name is absent from help and rejects before bridge access. The versioned helper's unbound help exposes none of the seven machine-native modes；its internal inherited dispatcher accepts one only with the protected launcher binding and rejects direct operator use. This restriction does not cover the separate nested-Linux verifier ingress already frozen in the verifier image：the fixed `verifier-entrypoint.sh` may directly exec only `validate-capsule-and-build-closed-set` under its compile-time image/toolchain identity and authenticated capsule/object-database contract, never a machine role mode or `OpenInstalledNativeRunnerSession`. Fixture-only/task-local helper modes retain their separately frozen authenticated test ingress and cannot mint a production role session.

Every execution reservation and inherited launch binding also contains a nonzero attested `BootIdentityDigest` (Windows boot/session epoch or Linux boot ID under the machine-base attestor). Same-boot recovery may exact-inspect/kill only the recorded PID/start plus Job/cgroup identity. A different authenticated boot proves the old process namespace ended and forbids acting on any reused numeric PID/cgroup name. It may **replace** the old reservation only for a guard-selected active variant that already has an explicit fixed recovery API：platform `resume_only_successor`；completion create `active_unstarted|publish_pending`；completion revalidate `retained_complete_active`；or an ordinary role's exact cleanup/tombstone/immutable-sender-envelope `Commit|InspectSender` resolution state. Each successor retains the exact same role/run/generation/install/business/publish binding and cannot expand that variant's authority. Separately, for `idle|retained_complete(active=nil)` or an ordinary terminal-inactive business phase, an authenticated boot change plus process-namespace-absence proof may only **clear** an exact pre-Open or `exiting` reservation by reservation-only guarded mutation；it cannot turn that reservation into a successor, consume a generation or call recovery. Platform Phase B binds the new boot identity while preserving the Phase-A successor and never calls Open；completion uses only the mode-matching active recovery API, while retained create success requires a later distinct read-only terminal-audit reservation after clear；ordinary senders can only finish their already recorded cleanup/envelope resolution. Fresh work after a terminal clear still requires a distinct purpose-bound reservation followed by the one permitted mode→phase Open or audit. A boot-identity splice、same numeric PID after reboot、cross-boot kill attempt or clear-and-Open/audit without a fresh matching reservation rejects before process/provider access.

`InspectFixedCompletionRoleState` is a non-consuming fixed-role router used only by the two installed completion parent modes；it authenticates the current process image、launcher mode/bootstrap、guard and selected cell and returns only the closed phase, with no bytes/handle/counter advance. The mode→phase map is closed. `completion-consolidate-create` may fresh-Open only `idle`, may use `RecoverFixedCompletionRoleState` only for `active_unstarted|publish_pending`, and for `retained_complete(active=nil)` may use only the distinct terminal audit below. `completion-revalidate` never infers recovery from an error or the coarse phase：it calls `InspectFixedCompletionRevalidationDispatch` before every target. That non-authorizing enum authenticates the selected `completion-revalidate` reservation purpose、guard/current cell and containment state and returns exactly three values. Exact `retained_complete_active` returns only `active_recovery` after the old helper containment is absence-proved and a same-generation `revalidation_recovery` replacement reservation is installed+reread；it cannot return fresh or audit. At `retained_complete(active=nil)`, durable `outcome_delivery_pending` returns only `outcome_audit` and requires `revalidation_outcome_audit`, while pending absent returns only `fresh_open` and requires the distinct fresh-work reservation. A reservation/pending/generation mismatch、unknown value、live-or-ambiguous prior containment or state error returns no dispatch and must not fall through. The subsequent Recover、audit or Open API repeats its own guarded exact match before touching state, so the enum cannot itself mint authority and an error never authorizes another target. This selector is package-owned durable state, never an argv/boolean/caller choice. Every other mode/phase pairing rejects, and the guarded recheck closes the inspect/use race. `OpenCompletionRoleStateForSession` accepts only a nonzero one-use completion session already guarded-claimed by that exact process and matching mode；a session for another role、mode or generation rejects. `RecoverFixedCompletionRoleState` has no role/path/store/options argument, re-authenticates the fixed installed completion process and returns a restricted coordinator only for the current launcher mode's guarded same-generation recovery route：create mode requires exact `active_unstarted|publish_pending` plus its publish-recovery replacement reservation, while revalidate mode requires exact `retained_complete_active` plus `active_recovery`; every other pairing rejects. It cannot be an alternate Open or claim a generation. `TestNativeRunnerCompletionRevalidationDispatchIsClosedAndGuarded` exact-covers all three values、the active-crash absence/replacement seam、corrupted/mismatched reservation and proves every target rejects the other two values while errors never fall through.

After E2 selects `retained_complete`, the matching create execution reservation is `exiting` and has no recovery successor. The launcher must first wait/reap or, after a crash/reboot, prove the exact old containment/process namespace absent and clear only that old reservation. A later `completion-consolidate-create` invocation may then install a distinct read-only `terminal_audit` execution reservation bound to the selected retained payload and immutable finalization-audit/set identity；it consumes no launch generation and cannot coexist with a revalidation reservation. Only that selected helper may call `InspectFixedCompletionRetainedFinalization(ctx)`. The function authenticates the terminal cell's immutable retained-payload/finalization-set audit digest, the powerless original finalization audit/set identity and the closed authenticated lineage from original E2 through zero or more closed `completed|aborted` revalidation edges plus reservation-only successors to the currently selected `retained_complete(active=nil)` cell；that audit digest excludes mutable `next`、`active`、execution-reservation and inner-cleanup fields. It returns only `CompletionRetainedFinalizationReceipt`. It does not return a coordinator/store/set/view bytes, call Open, mutate business/channel state, acquire adoption/ACK authority or make a same-generation recovery claim. The launcher marks the terminal-audit reservation `exiting` by a reservation-only mutation after the helper result, then clears it only after wait/reap and absence proof. A create/revalidate splice, audit before old-reservation clear, audit against `retained_complete_active`, second live audit, missing/unknown/rewritten lineage outcome, broken generation chain or any immutable retained-payload/finalization-set drift rejects.

The revalidation terminal CAS has a separate response-loss boundary. The same guarded terminal mutation that selects either `completed` or `aborted` also installs a package-owned `outcome_delivery_pending` record bound to that exact edge generation/outcome、immutable payload and latest trusted-time head. The old revalidation handle is intentionally dead and cannot answer recovery. After the matching `exiting` reservation is reaped、absence-proved and cleared, the fixed `completion-revalidate` launcher must inspect this pending bit before any fresh Open. While it exists, the only legal action is to install exactly one distinct read-only `revalidation_outcome_audit` reservation；only that selected helper may call `InspectLastFixedCompletionRevalidationOutcome(ctx)`. The returned sealed receipt is inspected only through `InspectCompletionRetainedRevalidationOutcomeReceipt` and exposes the pending closed outcome、its consumed launch generation、the byte-identical immutable payload digest and the guard-authenticated latest trusted-time head. It cannot Open or recover a generation、rerun validation、choose/change an outcome、mutate state or expose manifest/store bytes. The receipt must link the immediately preceding terminal edge to the currently selected retained cell and therefore makes post-CAS response loss distinguishable from pre-CAS failure without resurrecting the active handle. No prior edge may be selected, and a create `terminal_audit` reservation cannot call this API.

`PersistLastFixedCompletionRevalidationOutcomeDelivery` accepts only that exact sealed outcome receipt and writes the bounded canonical outcome view into the fixed launcher's generation-bound protected result slot selected in the `revalidation_outcome_audit` execution reservation. It O_EXCL-creates or exact-recovers the same slot record, caps/writes/fsyncs/reopens/verifies it and fsyncs the fixed parent directory before returning a sealed delivery receipt bound to slot identity、generation、outcome、payload and latest head. It accepts no path、bytes、sink、status or selector；a different slot/receipt or same generation with different bytes rejects. The durable slot is the consumer's acceptance record and remains replay-readable after helper/launcher response loss, so an ACK can never erase the only copy of an outcome.

`AcknowledgeLastFixedCompletionRevalidationOutcome` requires both the exact outcome receipt and its assignment-identical delivery receipt before performing the guarded reservation/pending-only ACK；it cannot alter the outcome/head/payload/next generation. Same-pair ACK replay is idempotent, while a missing、different or older receipt rejects. The launcher then marks the audit reservation `exiting` and clears it only after wait/reap/absence proof；the protected result record remains generation-addressed for response-loss delivery and is never overwritten by a later generation. A fresh revalidation Open is forbidden until the durable delivery receipt、pending ACK and audit-reservation clear are all selected；afterward the identical `retained_complete(active=nil)` business shape is unambiguous fresh work. Thus the zero-argument mode never chooses between audit and Open from an identical state. `TestNativeRunnerCompletionRevalidationPostTerminalResponseLossAudit` and `TestNativeRunnerCompletionRevalidationOutcomeDeliveryMustAckBeforeNextOpen` cover completed and both aborted routes across CAS/result-slot write+file fsync+directory fsync/result delivery/ACK/launcher-clear seams and prove no new Open/provider call/state mutation occurs before durable acceptance+ACK.

The opaque coordinator is the only cross-package completion-state bridge used exclusively by P09 `internal/c12completionpublication/operator.go`. `ReadCurrent` returns only the bounded authenticated header fields in `AuthenticatedCompletionRoleStateView`—epoch/full-state digest/publish-binding/install/phase/launch/immutable-payload digest；the coordinator internally validates the complete business projection before every semantic operation. Raw canonical business bytes、guard/cell/install paths、keys、channel/provider capabilities、execution-reservation bytes and sealed continuation bytes never leave it. `StateDigest` covers the entire selected cell including private execution-reservation and inner-cleanup substates. Nonzero `PublishBindingDigest` uses `SHA-256(ASCII("TALENRO-C12-COMPLETION-PUBLISH-BINDING-V1") || 0x00 || JCS(canonical outer-publish business))`. The projection exact-binds role/run/launch generation、outer retained-root/manifest/materialization/completed-marker identities and state、ordered channel reservations、the immutable `AdoptionSetIdentityDigest`、latest `LastSetTransitionDigest` and every semantic adoption/publication/ACK/finalize-ready intent+actual field. It explicitly excludes every mutable adoption-set pending/actual/full-record digest、any set field that copies the new `PublishBindingDigest` or new full `StateDigest`、launcher/helper PID/start/containment、execution-reservation status、StateEpoch/PreviousStateDigest、install-selection metadata and the entire same-record inner union `none|protected_key_intent|inner_bootstrap_restart_pending|cleanup_capsule|finalization_pending`. Execution-reservation-only or ordinary inner-only permits therefore change the full digest but preserve `PublishBindingDigest` byte-identically.

The reciprocal digest graph is a frozen DAG, never a fixed point. `AdoptionSetIdentityDigest I = SHA-256(ASCII("TALENRO-C12-EVIDENCE-ADOPTION-SET-IDENTITY-V1") || 0x00 || JCS(receiver/run/generation/profile/set-reservation/ordered-expected-bindings))` and contains no epoch、state/publish digest or mutable set record. For every ordinary begin/adopt/publication/ACK set-bound semantic mutation, `SetTransitionDigest T = SHA-256(ASCII("TALENRO-C12-EVIDENCE-SET-TRANSITION-V1") || 0x00 || JCS({AdoptionSetIdentityDigest,PreviousSetTransitionDigest,SourcePublishBindingDigest,TransitionKind,CanonicalSemanticDelta}))`. `T` contains no new publish/state/set-record digest. The private strict-canonical set records have exactly two ordinary V1 projections and reject unknown、missing、null or default fields. `EvidenceAdoptionSetPendingV1` contains exactly `{I,T,PreviousT,SourceP,TransitionKind,CanonicalSemanticDelta}` plus its own `RecordDigest Sp = SHA-256(ASCII("TALENRO-C12-EVIDENCE-ADOPTION-SET-PENDING-V1") || 0x00 || JCS(the six fields without RecordDigest))`. `EvidenceAdoptionSetActualV1` contains exactly `{PendingRecordDigest=Sp,I,T,NewStateEpoch,NewStateDigest,NewPublishBindingDigest,CanonicalSemanticResult}` plus its own `RecordDigest Sa = SHA-256(ASCII("TALENRO-C12-EVIDENCE-ADOPTION-SET-ACTUAL-V1") || 0x00 || JCS(the seven fields without RecordDigest))`. The fixed durable order is: (1) persist+file/parent-fsync and reread the exact pending record `Sp`；(2) reauthenticate the current guard/full digest, apply only the stored delta, store T in the outer business, compute `Pnew` and the new full digest, guard-CAS+reread；(3) persist+file/parent-fsync and reread the exact actual record `Sa` carrying `Sp` and `Pnew`. Neither T nor `Pnew` nor any outer/view digest contains `Sp`、`Sa` or another mutable/full set-record digest. Recovery has exactly two authenticated ordinary branches：if current P equals `SourceP`, an intervening execution/ordinary-inner-only CAS may rebase step 2 onto the newer epoch/full digest and apply the stored delta；if current P already equals the deterministically recomputed `Pnew` and the selected outer business carries the same I/T/delta/result, step 2 is already complete and an absent Sa binds the current selected epoch/full/Pnew after any such permitted successor. If Sa was already durable, a later permitted execution/ordinary-inner-only successor does not rewrite Sa and only remints the opaque proof against the current selected epoch/full state. Any other P、outer semantic、I/T/delta/result or disallowed inner drift rejects. The hash edges are `T → Sp`, `T → Pnew` and `{Sp,T,Pnew,new epoch,new full,result} → Sa`, while the durable sequence is `Sp selected → Pnew selected → Sa selected`; no later digest is fed back into an earlier preimage.

`finalize_ready` is the sole terminal set-bound exception and does not use `EvidenceAdoptionSetActualV1`. Its T uses the same transition domain with the exact canonical delta `{E2TargetTransitionDigest,ImmutableRetainedPayloadAuditDigest}`. `EvidenceAdoptionSetFinalizeReadyPendingV1` contains exactly `{I,T,PreviousT,SourceStateEpoch,SourceStateDigest,SourcePublishBindingDigest,E2TargetTransitionDigest,ImmutableRetainedPayloadAuditDigest}` plus `RecordDigest Fp = SHA-256(ASCII("TALENRO-C12-EVIDENCE-ADOPTION-SET-FINALIZE-READY-PENDING-V1") || 0x00 || JCS(those eight fields))`. After that pending record is durable, the guarded outer prepare CAS stores the same T and canonical delta and produces the post-prepare anchor `E1/D1/P1`. The source-only marker `EvidenceAdoptionSetFinalizeReadyActualV1` contains exactly `{PendingRecordDigest=Fp,I,T,SourceStateEpoch,SourceStateDigest,SourcePublishBindingDigest,E2TargetTransitionDigest,ImmutableRetainedPayloadAuditDigest}` plus `RecordDigest Fa = SHA-256(ASCII("TALENRO-C12-EVIDENCE-ADOPTION-SET-FINALIZE-READY-ACTUAL-V1") || 0x00 || JCS(those eight fields without RecordDigest))`. It deliberately contains no `E1/D1/P1` and no future E2/full-cell digest. `RecoverCompletionRetainedFinalization` authenticates Fp and any present Fa, converges a missing source-only Fa, and requires the selected outer state to carry the same I/T/delta/outer business with inner `none` and byte-identical P1. It may be the immediate E1/D1/P1 anchor or, after old-containment absence proof and guard-selected replacement reservation, a zero-or-more execution-reservation-only successor `E'/D'/P1`; that is the only terminal rebase. Recovery reconstructs the sealed proof from the **current selected** epoch/full/P1 and Commit rechecks that current proof before constructing E2. Any inner mutation、outer semantic drift、different T/delta/P1、unselected cell or caller-supplied anchor rejects. No I、T、outer P/full digest、materialized-view binding or terminal target/audit digest includes Fp or Fa.

While phase remains `publish_pending`, every outer publication-business mutation is possible only through this set-bound semantic permit. The sole terminal exception is the guard CAS preauthorized by the reciprocal set's immutable `finalize_ready` semantic transition. Its T binds the source publish digest、exact E2 target-transition digest and immutable retained-payload/finalization-set audit digest. The ready record additionally records the source epoch/full/publish values but deliberately contains neither post-prepare `D1/P1` nor a future full-cell digest；the returned proof alone carries actual `E1/D1/P1`. The target-transition digest binds the one permitted `retained_complete(active=nil,next=N+1)` E2 candidate；the separate audit digest binds only the immutable retained payload plus stable finalization/set identity and excludes `next`、`active`、execution reservation and inner cleanup. After Prepare, only `CommitCompletionRetainedFinalization` may consume the proof and internally construct/fsync the exact target terminal cell as `E2=E1+1, PreviousStateDigest=D1`, compute its full digest, compare-and-advance the guard and consume the binding. The set then becomes a powerless immutable audit record and requires no retained-phase reciprocal publish digest. The private state engine accepts only the closed graph `active_unstarted→publish_pending`, set-bound `publish_pending→publish_pending`, proof-bound `publish_pending→retained_complete` and `retained_complete_active→retained_complete_active|retained_complete`; the last edge must carry the package-selected sealed `completed|aborted` outcome, and no generic mutation entry exists. Private `OpenInstalledNativeRunnerSession` alone may perform `idle→active_unstarted` or `retained_complete→retained_complete_active`. Same-phase inner/execution permits preserve launch generation；inner-only changes preserve publish digest, while set-bound outer-publish permits update it；the proof-taking publish→retained hook sets next launch exactly `N+1` and immutable payload once；revalidation terminal sets next exactly `N+1` relative to its active generation, appends exactly one authenticated closed outcome edge and keeps that payload/audit digest byte-identical. Backward phase、generation skip/reuse、outcome switch、retained payload drift、caller epoch/previous/counter/install field、raw capability bytes or reused view/coordinator fails before staging. Tests `TestEvidenceChannelReciprocalDigestGraphIsAcyclic` and `TestEvidenceChannelSetActualDigestNeverFeedsPublishBinding` exact-recompute I、every T、Pnew and set-actual digest from their declared preimages, reject any self/new-P/full-set-digest inclusion and prove crash recovery at all three DAG steps. The returned view is bound to the new selected full and publish-business digests and cannot be fabricated or used as mutation input.

The two reciprocal-DAG tests must additionally exact-recompute `Sp/Sa/Fp/Fa` with the four declared domains and exact field sets, mutate every field/domain separator independently, and cover ordinary pending → outer CAS → actual plus terminal pending → outer prepare CAS → source-only marker → in-memory-proof loss → proof reconstruction. Any implementation that places post-prepare `E1/D1/P1` in the finalize-ready marker, places `Sp/Sa/Fp/Fa` in an outer/view digest, or accepts an actual/marker whose authenticated pending predecessor does not match must fail the gate.

`InstallFixedNativeRunnerProfile` is the only installer boundary. It takes no profile/path/tool/catalog argument, resolves the role's fixed tracked token and the session-authorized relock projection—B10 toolchain/profile/launcher inputs without a future catalog, or B11 toolchain/profile/catalog/launcher inputs—only through the installer session's authenticated candidate-package handle, stages/exact-opens only the declared native executables beneath an installer-derived immutable generation path under fixed application-control policy, emits one receipt per exact native role, and runs the same strict verification core used by `VerifyFixedNativeRunnerInstallation` against the staged install record. It linearly consumes the session before any observable work and defers destruction of the session plus candidate handle on every return path, including preflight failure、candidate fsync failure、guard response loss and success；a reused session or surviving handle is invalid. The machine-base launcher is deliberately absent from the candidate package's role receipt set. The installer never overwrites or retargets a currently selected executable/path/receipt set. The candidate role-state cell binds the new install generation and exact staged install-record digest；one role-guard compare-and-advance selects both that install record and the preserved/new runtime state and is the sole activation point—HKLM/Linux bootstrap keeps stable store/guard/cell/root identities and has no mutable active-install pointer or stable helper symlink. It never installs a snapshot script：the final parent resolves each `AllowedTrackedScriptToken` only inside its authenticated committed read-only snapshot, verifies the token's exact Git blob SHA-256/mode against that tree plus snapshot-root and no-follow opened-file identities, holds the root/file identities through spawn and executes only that derived snapshot path under the receipt-pinned interpreter. A fixed installed reference-script path, caller path or byte-equal file outside that snapshot cannot substitute. B10 invokes the installer for `windows_scopes` and `staged_diagnostic` only after the five-output relock; B11 invokes it for all five runner roles only after the nine-output relock. On a role's first installation, the authenticated initial guard projection is `(epoch=0,digest=zero)` and the role remains uninstalled/retryable until activation. The protected installer provisions one role-unique non-resettable rollback-resistant guard plus two fixed nonenumerable no-follow sealed state cells and stable bootstrap identities, stages/fsyncs the immutable generation-1 native files+receipts+install record and exact inactive runtime record (`idle(next=1)` for completion), then advances the guard `(epoch=0,digest=zero) -> (epoch=1,digest=initial-state)` and reopens/verifies the selected cell/install record before returning. A first-install pre-advance crash therefore leaves no selected generation or launchable helper. A reinstall is allowed only when the ordinary role is terminal-inactive or completion is exactly `idle|retained_complete(active=nil)`；it stages a new immutable generation without modifying the selected one, preserves the same guard/cell identities、next launch generation and any immutable retained payload, and selects the new install record plus preserved runtime state with one epoch advance. Installation with an active/pending state, missing prior guard, changed guard identity or a reset counter rejects. Before a reinstall guard advance, its candidate has zero runtime authority and the launcher resolves and starts only the byte-identical old selected helper；after any successful activation, it resolves and starts only the new selected helper and exact-forward recovery may finish verification/retire an old generation, never fall back. A missing/corrupt post-advance selected install quarantines that role. A diagnostic capability expansion or pre-advance candidate/bootstrap fsync failure leaves an existing selected generation unchanged and, on first install, leaves the role uninstalled；it never changes another role. There is no helper-hosted installer CLI, stable helper path or caller-selected destination.

The private authenticated install record also carries exactly one closed `ProjectionVariant=b10_pre_catalog|b11_final_catalog`, selected only from the protected candidate session and covered by the install-record digest. `b10_pre_catalog` binds the exact B10 toolchain native-role、tracked profile and frozen launcher projections and requires catalog absent/zero；`b11_final_catalog` binds those values plus the nonzero final catalog exact cover and final relock projection commit. Neither Verify nor any public caller can choose or upgrade this variant.

`VerifyFixedNativeRunnerInstallation(context.Context, NativeRunnerRole) error` is the exact non-consuming audit boundary. It has no profile/path/store/options/projection-variant parameter, opens the role's fixed tracked token and fixed machine store, authenticates the unique selected install record and dispatches solely on that record's sealed `ProjectionVariant`. The B10 branch requires exact profile/toolchain-native-role/frozen-launcher projections and catalog absent/zero；the B11 branch additionally requires the nonzero final catalog exact cover and matching relock projection commit. B10 with any catalog projection、B11 with absent/zero/mismatched catalog、an unknown variant or a caller attempt to select the branch rejects. Both branches strict-validate the active install generation、complete native receipt set、application-control policy、diagnostic zero-capability projection and the role guard's current epoch/digest against the unique selected state cell, reopen every native executable by identity, and return only `nil|error`. The function cannot mint a session/handle、read a channel/provider secret、change an active launch state、write either cell or consume/advance the guard/launch generation. B10/B11 post-install gates use this function；relock itself never uses it because no matching installation exists until after the rewrite.

Production `OpenInstalledNativeRunnerSession` has no path/profile/store/options parameter and reads no argv, environment, package global, caller cwd or default TEMP/home. An operator invokes only a fixed launcher mode. The launcher authenticates the guard-selected cell/install record, no-follow opens the exact versioned helper named by that record and directly spawns it with an inherited-only, one-use protected launch binding over role、public mode、install generation、install-record digest、helper identity and child PID/start token. The helper's compile-time OS adapter consumes that binding internally, reloads the same selected record, and only then loads the profile through its fixed tracked token and opens the fixed machine-sealed bootstrap location (`HKLM\\SOFTWARE\\Talenro\\C12\\RunnerBootstrapV1` with machine DPAPI/application-control protection on Windows; `/var/lib/talenro/c12/runner-bootstrap.v1.sealed` with owner-only OS-keystore sealing on Linux). The binding is never an API argument、numeric argv/env value or serializable byte string；direct invocation of a versioned helper, a stale-generation helper, a launcher-selected role/mode mismatch or a post-open replacement fails before session/state access. Each fixed store is a strict role-indexed sealed map：every role has an independent install generation and monotonic launch generation with at most one active run；installing/opening/terminalizing one role cannot consume、replace or roll back another role's record. Each role record binds the canonical Git object-database locator plus observed database identity, one owner-only external-root capability, the fixed recovery-authority slot, the allowed tracked-script token policy, exact native tool/helper/private-child install-receipt set, zero/one/three ordered actual channel namespace/next-slot identity/sequence capabilities matching `ChannelBindings`, plus only the role-allowed provider/evidence-root capabilities. It also binds the freshly validated tracked profile digest, toolchain projection and catalog projection. Absolute versioned native paths and machine identities exist only in that selected record/install receipts and are never returned by the launcher；tracked script paths are derived afresh from the authenticated snapshot as above, and the tracked profile contains role/purpose tokens, never machine values. The `completion` role alone specializes its fixed recovery slot as one logical authenticated `C12CompletionRoleStateV1` record that is the sole source of truth for that role's next/active launch generation、outer publication state and nested temporary-cleanup substate；the HKLM/bootstrap install record keeps only install receipts/projections plus the guard/cell/slot authority identities and has no second completion active bit、launch counter or publication record.

Every role's fixed adapter implements its launch/install runtime record with two role-private A/B sealed cells and one distinct machine-provisioned rollback-resistant monotonic CAS guard. On Windows the guard is TPM-backed through the fixed native platform adapter；on Linux it is a dedicated TPM/vTPM NV or equivalently attested non-rollbackable platform register. A filesystem/registry value、DPAPI/MAC alone、wall clock、install generation、trusted-time sequence or another agent/supervisor/latch role's counter cannot substitute. The guard's complete protected value is only `(Role, CounterIdentityDigest, StateEpoch, CurrentStateDigest)` and its bootstrap binding names the fixed A/B cell identities；it contains no phase、run、launch generation、next、resource、channel、publication or retained payload and therefore is freshness/selection metadata, not a second role-state truth. Every sealed cell includes the same nonzero `CounterIdentityDigest`, its `StateEpoch`, `PreviousStateDigest`, exact install-record digest and the complete logical runtime record. For completion that record directly contains the entire `C12CompletionRoleStateV1`; ordinary roles instead bind the exact identity+digest+closed variant of their separate fixed cleanup/resume/tombstone recovery-slot payload, so replaying that payload alone grants nothing. Epoch 1 alone has `PreviousStateDigest=zero`. Loading current state reads/authenticates the guard and verifies only the parity-selected no-follow cell's role/counter identity、epoch and complete digest plus its selected install/recovery payloads；it does not require the inactive cell to retain epoch E-1 history. `PreviousStateDigest` is checked only when a well-formed E+1 candidate is compared with expected selected `(E,D)` and again after a successful advance. An earlier valid selected cell/payload replay, copied selected cell, wrong counter, no matching selected cell, same-epoch digest mismatch, counter reset/reprovision or guard/provider unavailability fails closed before any session/recovery/channel/provider/root authority；an incomplete/unselected future candidate has zero authority and cannot invalidate an intact guard-selected E state.

All install/open/recovery/inner/adoption/ACK/finalize/terminal mutations use one private guarded whole-record CAS. With expected `(E,D)`, the adapter exclusively stages and file+directory-fsyncs the complete candidate in the inactive parity cell as `(StateEpoch=E+1, PreviousStateDigest=D)`, including the exact digest/identity of any already-durable ordinary-role recovery-payload or install-record candidate, reopens/verifies its seal、cell identity and digest, then atomically compare-and-advances the rollback-resistant guard `(E,D) -> (E+1,D')`; that guard update is the sole linearization/selection point. Only after rereading the guard and unique selected candidate and its exact bound install/recovery payloads may an API return a capability or success；a pointer/cache update is optional and never authoritative. A normal crash before the guard advance leaves `(E,D)` fully usable and makes every staged/partial inactive candidate grant zero authority；the fixed adapter may discard an incomplete candidate or exact-forward a complete one because a later guard digest selects only the actually advanced candidate. A normal crash after advance makes only `(E+1,D')` authoritative and exact-forward recovery finishes that selected candidate. Only malicious corruption/missing selected bytes after advance or stale expected epoch/digest quarantines；normal candidate/fsync/guard crash seams resolve to old or new without quarantine, fallback or a same-epoch fork. External idempotent channel operations remain preceded by an anchored intent；if their subsequent actual-state guard advance fails, recovery inspects/replays only that exact intent. Thus replaying any formerly valid install record、ordinary resume payload or completion `active_unstarted`、`publish_pending`、`retained_complete` or active-revalidation cell cannot restore authority or reuse a launch generation.

Opening a session repeats the non-consuming installation checks, proves the object database resolves the bound locator, exact-covers the role's zero/one/three ordered channel bindings, and atomically CAS-claims that role's next monotonic launch generation as one active run. A second parent for the **same role while that run is active** rejects；this is not a permanent one-run installation. Ordinary roles use the cleanup/resume/tombstone recovery transitions below and, after terminal cleanup, atomically advance only their own sealed launch generation without reinstalling unchanged files. For `completion`, Open specifically replaces `C12CompletionRoleStateV1 idle(next=N)` with its recoverable `active_unstarted(generation=N,run,session_digest)` variant. That variant is itself machine-authenticated recovery authority, so a parent crash immediately after Open cannot strand an unopenable opaque in-memory session：only the fixed zero-argument completion mode may pathlessly authenticate and resume or no-resource terminalize that exact run/generation, never call Open again、select a record or claim a new generation.

`completion-consolidate-create` uses the closed same-record phase transition. P09 `BeginCompletionPublishPending` consumes/invalidates the in-memory active session and uses one guarded whole-record mutation to change `active_unstarted(N)` into `publish_pending(N)` containing the authenticated outer publish state and all launch/session capabilities needed for exact recovery. A crash before the guard linearization leaves the recoverable `active_unstarted(N)` record；after it leaves only the pathlessly recoverable `publish_pending(N)` record. The completion inner `C12CleanupOnlyOwnershipCapsuleV1` and cleaned-tombstone payload remain logically separate least-authority values：the capsule is held only by the `cleanup_capsule` tag, while the selected tombstone binding exists only inside `finalization_pending`, never as its own tag、file、slot or recovery authority. Those tags are authenticated **nested substates of this same top-level record** alongside the other frozen inner tags. Each inner intent/restart/capsule/finalization-pending update advances `StateEpoch` and compare-and-replaces the whole guarded role-state digest；the fixed-target CAS may remove only that inner pending substate after WAL-key destruction and cannot alter unrelated outer publish fields or advance the launch generation. After manifest durability and all three ACK actuals, P09 `handle.FinalizeRetainedComplete(ctx)` performs one logical terminal transaction with two guarded state advances. Receiver Prepare cross-persists `finalize_ready` and advances `publish_pending(E)→publish_pending-finalize-ready(E1)` without constructing a terminal cell；receiver Commit alone consumes the returned E1/D1/P1 proof, constructs+fsyncs the opposite-cell candidate and advances `E1→E2` to `retained_complete(immutable_payload,next=N+1,active=nil)` while marking the matching execution reservation `exiting`. E2 is the sole terminal linearization and the only point that removes mutation/recovery authority and makes the next generation business-available；before it, no terminal cell exists. The launcher must still reap/prove absence and clear the exiting reservation before any terminal audit or next revalidation reservation. The later `completion-revalidate` actual parent guarded-claims `retained_complete(next=N+1,active=nil) -> retained_complete_active(generation=N+1,immutable_payload,inner_cleanup_substate)` without changing the immutable payload or reinstalling；its terminal inner cleanup replaces only the launch/inner envelope with `retained_complete(next=N+2,active=nil)` and the byte-identical payload while marking its reservation exiting. Cross-role records remain usable throughout. A final-role session returns only its opaque one-run object-database, external-root, recovery-slot, exact native-executable, derived-snapshot-script and exact channel/provider/evidence capabilities. A `staged_diagnostic` session returns only canonical object-database, owner-only external-root, recovery-slot and exact Linux helper/Git/build-tool handles: it has no script, channel sender/receiver, provider credential, evidence root, scope/result/channel attestor or publishing handle. No accessor returns an absolute native path, raw channel key, provider secret, WAL key or generic directory capability; only direct native spawn/typed operations accept those handles. The authenticated final session binds each channel namespace plus **next** slot identity/sequence, never a future transaction nonce.

For `retained_complete_active(generation=G)`, cleanup alone never chooses outcome. After the paired set is exact-absent and both halves hold the same prepared receipt, completed and the two abort methods race through the first CAS into sole `finalization_pending`：success fixes `completed` only with the validated token；abort fixes non-affirming `aborted` and accepts no caller outcome bytes. The only extra is the sealed P08 trusted-time-head candidate bound to this child/result/generation. Pending stores the complete receipt、tombstone、fixed target and candidate according to the frozen rule；same-target replay matches, opposite returns `ErrRetainedRevalidationOutcomeConflict`. Once selected, only no-handle `RecoverInnerFinalization` reconstructs them and completes leaf Recover/Commit、key absence、tombstone unlink+parent-directory fsync/exact absence and then the final target CAS. There is no handle/session retained-outcome recovery. That final CAS consumes G, preserves immutable payload/audit digest, appends the authenticated completed/aborted edge, installs `outcome_delivery_pending`, selects inactive next G+1 and marks reservation exiting；after it there is no tombstone work and response loss uses only read-only outcome audit→protected delivery→ACK. A valid candidate advances the latest trusted-time head on completed or later fail-closed aborted outcome；pre-time abort preserves the prior head. The next runtime must be strictly newer than the latest head, not merely creation. Cleanup failure cannot select pending/edge；aborted never affirms freshness or rewrites the manifest.

The abort API is deliberately split rather than nullable：`AbortRetainedRevalidation(ctx,handle,receipt)` is valid only before a trusted-time head has been semantically validated and preserves the previous head；`AbortRetainedRevalidationWithTrustedTimeHead(ctx,handle,receipt,candidate)` is mandatory after that validation barrier and advances the head while recording the same `aborted` outcome. `FinalizeRetainedRevalidation(ctx,handle,receipt,candidate)` always requires the matching candidate. All three terminal routes require the exact leaf finalization receipt stored in the pending intent before leaf commit/final CAS. A nil/typed-nil candidate or receipt is never a valid argument, and static call-site tests reject the P09 composite dropping an already validated candidate into the preserve-head abort route or using a different cleanup receipt. The two abort methods are same-outcome idempotent replays of one stored intent and both conflict with a durable completed intent.

For the one cross-reboot `platform` run, the Phase-A active session is not terminalized at reboot and Phase B must not call `OpenInstalledNativeRunnerSession` again. The guarded cleanup-only-capsule → authenticated resume-manifest transition stages/fsyncs the exact resume payload, binds its identity+digest in the inactive role-state cell and advances that role's guard once, thereby changing only the same role/run/launch generation from ordinary `active_run` to a sealed `resume_only_successor` bound to the exact manifest envelope、slot identity、run ID、role、runner/provider identities and retained capabilities. Before guard advance the staged resume payload grants zero authority and only the previously guard-selected delete-only state can be used；after advance only the resume state can be used, while a crash in between is exact-forwarded by the fixed adapter without a second Open. Phase B recovers capabilities solely through P09 `OpenResumeManifest`'s protected bootstrap adapter, which authenticates the guard、selected role cell and its bound successor before any handle access. The successor cannot start a new run、change role、reopen a launch generation or outlive its exact cleanup/publication lifecycle；expiry/provider outage still permits only its delete-only abort cleanup path. Abort/downgrade terminal cleanup advances that role's next launch generation exactly once with no signing authority. Successful workload cleanup instead uses the exact leaf receipt to select same-generation `platform_post_cleanup_publication`; only the exact nested/outer/channel committed-sender finalizer advances the next generation once. A role splice, copied tracked profile at another path, sealed-store/cell/recovery-payload rollback, caller-supplied absolute path, global/context injection, install-receipt replay, second same-role active parent, cross-role resume successor or sender/receiver/provider/publishing-capability cross-use fails before tree capture or resource creation.

Task 2 is also the sole owner of the public Windows→Linux bridge envelope and validator:

```go
type C12TreeFactsCapsuleV1 struct {
	SchemaVersion              string
	RunID                      uuid.UUID
	CommandNonce               [32]byte
	TreeFactsAttestorIdentity  contracts.Digest
	SnapshotIdentityDigest      contracts.Digest
	ObjectDatabaseIdentityDigest contracts.Digest
	ObjectDatabaseClosureDigest  contracts.Digest
	TreeLocator                 GitTreeLocatorV1
	TrackedTreeDigest           contracts.Digest
	CaptureKind                 string
	OwnershipWALDigest          contracts.Digest
	IssuedAt                    time.Time
	ExpiresAt                   time.Time
	Attestation                 []byte
}

type TreeFactsCapsulePolicy struct {
	ExpectedRunID                       uuid.UUID
	ExpectedCommandNonce                [32]byte
	ExpectedTreeFactsAttestorIdentity   contracts.Digest
	ExpectedSnapshotIdentityDigest      contracts.Digest
	ExpectedObjectDatabaseIdentityDigest contracts.Digest
	ExpectedObjectDatabaseClosureDigest  contracts.Digest
	ExpectedTreeLocator                 GitTreeLocatorV1
	ExpectedTrackedTreeDigest           contracts.Digest
	ExpectedCaptureKind                 string
	ExpectedOwnershipWALDigest          contracts.Digest
	ExpectedCapsuleFileIdentityDigest   contracts.Digest
	RunNotBefore                        time.Time
	RunNotAfter                         time.Time
}

// Opaque zero-value-invalid sealed state; only the authenticated private
// parent channel can mint one, and no public/raw-nonce constructor exists.
type AuthenticatedTreeFactsParentSession struct {
	sealedParentSessionState privateTreeFactsParentSessionState
}

type ProtectedTreeFactsHandle interface {
	protectedTreeFactsHandle() // opaque one-use child-inheritance input
}

type ProtectedTreeFactsParentBinding interface {
	protectedTreeFactsParentBinding() // capture-tree-minted run/WAL/snapshot/object-database facts; no caller fields
}

type FixedTreeFactsParentExchange interface {
	fixedTreeFactsParentExchange() // one fixed direct child, hidden challenge and one parent consume
}

type FixedTreeFactsCapsuleBuilder interface {
	fixedTreeFactsCapsuleBuilder() // inherited child-only attestor/challenge/binding; one use
}

type C12TreeFactsCapsuleTemplateV1 struct {
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type ProtectedTreeFactsBlobRole string

const (
	ProtectedTreeFactsOuterImageBuildContext ProtectedTreeFactsBlobRole = "outer_image_build_context"
	ProtectedTreeFactsOperationsBuildPolicy  ProtectedTreeFactsBlobRole = "operations_build_policy"
)

type ProtectedTreeFactsBlobRequest struct {
	Role  ProtectedTreeFactsBlobRole
	Index uint8 // outer contexts 0..2; operations policy exactly 0
}

type ProtectedTreeFactsReaderView struct {
	RunID                        uuid.UUID
	Locator                      GitTreeLocatorV1
	TrackedTreeDigest            contracts.Digest
	SnapshotFileIdentityDigest   contracts.Digest
	ObjectDatabaseIdentityDigest contracts.Digest
	ObjectDatabaseClosureDigest  contracts.Digest
	CaptureKind                  string
	ChildProcessBindingDigest    contracts.Digest
}

type ProtectedTreeFactsBlobView struct {
	Request          ProtectedTreeFactsBlobRequest
	RepoRelativePath string
	GitMode          string
	BlobSHA256       contracts.Digest
	CanonicalBytes   []byte
}

type ProtectedTreeFactsReader interface {
	protectedTreeFactsReader() // sealed child-only fixed-token reader
	Inspect(context.Context) (ProtectedTreeFactsReaderView, error)
	ReadFixedBlob(context.Context, ProtectedTreeFactsBlobRequest) (ProtectedTreeFactsBlobView, error)
}

func CanonicalTreeFactsCapsuleUnsignedPayload(C12TreeFactsCapsuleV1) ([]byte, error)
func CanonicalTreeFactsCapsule(C12TreeFactsCapsuleV1) ([]byte, contracts.Digest, error)
func (*AuthenticatedTreeFactsParentSession) ValidateConsumeAndOpenTreeFactsCapsule(C12TreeFactsCapsuleV1, TreeFactsCapsulePolicy) (ProtectedTreeFactsHandle, error)
func CaptureFixedTreeFactsParentBinding(context.Context, *c12runnerprofile.AuthenticatedNativeRunnerSession, SelectedOwnershipWAL) (ProtectedTreeFactsParentBinding, error)
func BeginFixedTreeFactsParentExchange(context.Context, *c12runnerprofile.AuthenticatedNativeRunnerSession, ProtectedTreeFactsParentBinding) (FixedTreeFactsParentExchange, error)
func OpenInheritedFixedTreeFactsCapsuleBuilder(context.Context) (FixedTreeFactsCapsuleBuilder, error)
func SealFixedTreeFactsCapsule(context.Context, FixedTreeFactsCapsuleBuilder, C12TreeFactsCapsuleTemplateV1) (C12TreeFactsCapsuleV1, error)
func ValidateAndOpenProtectedTreeFactsForFixedParent(context.Context, FixedTreeFactsParentExchange, C12TreeFactsCapsuleV1) (ProtectedTreeFactsHandle, error)
func ConsumeProtectedTreeFactsHandle(context.Context, ProtectedTreeFactsHandle) (ProtectedTreeFactsReader, error)
```

The exact schema is `talenro-c12-tree-facts-capsule/v1`. Strict decoding rejects unknown/duplicate/null/defaulted fields；`CaptureKind` is closed to `staged|committed`; the nonce is a fresh 256-bit CSPRNG value; all digests、locator members and times are nonzero/canonical with `RunNotBefore <= IssuedAt <= ExpiresAt <= RunNotAfter`. The unsigned projection excludes only `Attestation`; the distinct role-bound `TreeFactsAttestor` signs `TALENRO-C12-TREE-FACTS-CAPSULE-V1\x00 || JCS(unsigned)`, and the complete envelope digest is `SHA-256(ASCII("TALENRO-C12-TREE-FACTS-CAPSULE-ENVELOPE-V1") || 0x00 || JCS(envelope))`. `TreeFactsAttestor` is an OS-keystore/runner-profile locked, non-exportable one-purpose identity distinct from the scope `RunnerAttestor`; each rejects the other's EKU/domain and its one-use handle is destroyed immediately after capsule signing. `CaptureFixedTreeFactsParentBinding` is the named sole capture facade：it accepts only the authenticated native session and already selected sealed WAL recorder, derives committed/staged capture kind from the fixed role/mode, creates no caller-selected path, performs the locked capture plus minimal-object-database closure, and appends+file/directory-fsyncs their exact intents/actuals. Only then does it return a private-concrete `ProtectedTreeFactsParentBinding`; the existing recorder remains the same session-bound append surface for later declared resources. `BeginFixedTreeFactsParentExchange` exact-matches that binding to the same authenticated native session, consumes the binding, generates the independent 256-bit challenge and creates one `AuthenticatedTreeFactsParentSession` plus one child-only inherited builder in the fixed direct-spawn table；it returns only the opaque parent exchange. The command sees neither challenge、policy、builder handle nor attestor. The exact started child alone calls `OpenInheritedFixedTreeFactsCapsuleBuilder`; `SealFixedTreeFactsCapsule` accepts only `{IssuedAt,ExpiresAt}`, fills every run/WAL/snapshot/object-database/locator/capture/file-identity field and `CommandNonce` from that builder, signs with its one-use attestor, consumes builder+attestor on every return and returns the bounded public capsule. Thus the challenge exists before signing but is never argv/env/stdin/stdout or caller data.

`TreeFactsCapsulePolicy.ExpectedCommandNonce` is only a constant-time internal projection copied from the sealed parent session；it is neither replay state nor authority, and the lower method requires it to equal the hidden challenge. `ValidateAndOpenProtectedTreeFactsForFixedParent` accepts the matching opaque exchange plus the returned capsule, derives the complete policy from the exchange's sealed run/WAL/snapshot/object-database/capsule-file facts and immediately calls `ValidateConsumeAndOpenTreeFactsCapsule`; it accepts no raw native session、nonce、policy、path、digest、session literal or callback. The lower method is referenced in production only by that wrapper and first rejects nil、zero、foreign or already-consumed session state, then requires the fixed toolchain/runner trust policy and constant-time exact run/nonce、attestor identity、snapshot file identity、minimal object-database identity/closure digest、typed locator、canonical digest、capture kind、WAL digest、time interval and outer WAL-recorded capsule file identity；only after all validation succeeds does one atomic compare-and-swap consume the sealed session bit and return a nonzero opaque one-use `ProtectedTreeFactsHandle`. The fixed command private producer is the Capture/Begin/Validate chain's sole production caller；it passes the returned handle once to artifactscan, while every other command has zero reference. That handle can only be inserted by the same exchange's private direct-spawn inheritance table and cannot be opened、serialized or converted from a nil error/path by the caller. Inside that exact child, `ConsumeProtectedTreeFactsHandle` linearly consumes it and returns a sealed `ProtectedTreeFactsReader`; Inspect exposes only the authenticated locator/digest/identity/process view, while `ReadFixedBlob` accepts exactly outer-context ordinals 0..2 or the sole operations-policy ordinal 0 and returns a bounded defensive copy plus tracked mode/digest. It exposes no root、arbitrary path/token、FD/HANDLE or write authority. `artifactscan.BindAuthenticatedTrackedBuildInputReader` must call this sole bridge and wrap that reader；it cannot type-assert a c12evidence private concrete. Failure or a second consume returns a zero reader plus error. A second Capture/Begin、builder Open/Seal、wrapper/lower-method call, copied exchange/capsule, zero/caller-fabricated session or binding, cross-session/child replay, boolean/nil-error substitution, self-selected chain、caller policy/digest、valid capsule for another run/snapshot/object-database/WAL or scope attestation never yields a child input. `TestTreeFactsCapsuleStrictValidation` uses `go/types` over the production graph to prove the exact Capture→Begin→inherited Builder→Seal→Validate→lower-method positive chain, sole command caller and zero alternate mint/literal/caller-policy path in addition to its full semantic matrix.

- [ ] **Step 1: RED — freeze four-final-plus-diagnostic profiles, install receipt and bootstrap rejection matrices**

Add `TestNativeRunnerProfileStrictFiveRoleTokens`, `TestNativeRunnerProfileChannelBindingsAreRoleClosed`, `TestStagedDiagnosticProfileIsNonpublishing`, `TestNativeRunnerInstallReceiptBindsNativeExecutableIdentityAndRejectsScript`, `TestTrackedSnapshotScriptIdentityIsDerivedNotInstalled`, `TestInstallFixedNativeRunnerProfileIsAtomic`, `TestInstallFixedNativeRunnerProfileConsumesSessionAndDestroysCandidateOnEveryReturn`, `TestVerifyFixedNativeRunnerInstallationIsNonConsuming`, `TestNativeRunnerLaunchGenerationsAreRoleIndexed`, `TestNativeRunnerSameRoleActiveRunIsExclusiveAndRearmsAfterCleanup`, `TestNativeRunnerPlatformResumeSuccessorUsesSameGeneration`, `TestNativeRunnerCompletionRoleStateUsesSingleRecoverySlot`, `TestNativeRunnerCompletionRetainedTransitionUsesPrepareAndCommitGuardAdvances`, `TestNativeRunnerCompletionTerminalFinalizationAuditIsPathlessReadOnly`, `TestNativeRunnerCompletionModePhaseCrossClaimRejects`, `TestNativeRunnerRoleStateRejectsAuthenticatedRollback`, `TestNativeRunnerRoleStateGuardCrashSeams`, `TestNativeRunnerFirstInstallInitializesRoleState`, `TestNativeRunnerReinstallPreservesLaunchStateAndRetainedPayload` and `TestOpenInstalledNativeRunnerSessionRejectsAmbientAndCrossRoleInputs`. The vectors strict-mutate schema/role/script/tool/helper/recovery/external-root/channel/provider/evidence-root/install-policy fields, fixed tracked token, canonical JCS and domain; require exact one sender binding for each final producer, exact three ordered receiver bindings for completion, and exact zero bindings for diagnostic; mutate kind/access/slot purpose/`next` policy/count/order. Diagnostic vectors require empty scripts, Linux helper/Git/build-tool exact cover, `none/nonpublishing` provider/evidence purposes and absence of every channel/provider/evidence/signer/publisher handle; a Windows binary, authority profile, extra shell, provider credential or fake output handle rejects. Receipt vectors mutate native executable path/hash/OS/arch/file identity/application-control/attestation and require any tracked script token/file to be rejected as a receipt subject；snapshot-script vectors mutate committed blob/mode、snapshot-root/file identity、no-follow open and post-open replacement, and prove only the parent-derived snapshot path can execute under the receipt-pinned interpreter. Fixed-store vectors mutate role index、install generation、the sole recovery-slot record、next/active launch generation、run/session digest、Git object-database identity+locator、exact external root/recovery slot、toolchain/catalog projection、each next channel slot/sequence and sender/receiver/provider capability. Guard vectors mutate/cross-swap/reset/replay counter identity、epoch、current/previous/state/install digests、A/B parity/cell identity and provider class；copy every earlier valid ordinary/completion cell over the selected cell and require zero returned authority, inject before/after candidate file+directory fsync、guard compare-and-advance、post-advance reread and optional pointer/cache seams, and require every normal reinstall crash to recover exactly the old or new digest selected by the guard, never quarantine/fallback/same-epoch fork. Installer vectors inject every immutable generation-file/receipt/install-record/internal-verify/stable-bootstrap/cell/guard/fsync seam and every return path；first install must remain authenticated `(0,zero)` uninstalled/retryable before its sole guard advance, then create/select epoch-1 inactive state, while reinstall preserves guard identity、next generation and retained payload and rejects every active/pending state. They assert the installer session is consumed and its candidate handle destroyed before every return, response loss cannot reuse either, an existing selected generation remains byte-identical before a reinstall advance, a first-install pre-advance crash has no selected helper, only the complete new generation is usable afterward, bootstrap has no active pointer, and no crash exposes an old/new mismatch window or changes another role. `VerifyFixedNativeRunnerInstallation` must repeat strict checks while leaving every role's guard/cells/launch record byte-identical；the first actual `Open` claims one role/run, a second same-role active parent rejects, another role still opens, and terminal cleaned-tombstone completion advances only that role so a later run can open without reinstall. Platform vectors guarded-transition the Phase-A authority to a same-role/run resume-only successor, recover Phase B only through that successor, and reject a second `Open`、new-run/role splice or successor use after terminal cleanup. Completion vectors inject before/after Open、`active_unstarted→publish_pending`、E→E1 finalize-ready、proof-bound E1→E2 terminalization、reservation exiting/absence/clear/terminal-audit selection and retained→revalidation-active replacement；they require exactly one active same-run/generation recovery authority before E2, no recovery authority after E2, no separate active bit/publication record, a read-only create-mode terminal audit only after old reservation clear, strict mode→phase rejection and immutable retained bytes across revalidation. They must prove argv/environment/cwd/TEMP/home/package-global/caller path cannot select any input, and that `windows_scopes` plus `staged_diagnostic` have B10 tracked instances while the other three final tokens are recognized-but-required-later.

Also add `TestVerifyFixedNativeRunnerInstallationClosesB10AndB11ProjectionVariants` and `TestNativeRunnerCompletionTerminalAuditSurvivesCompletedRevalidations`. The first mutates the authenticated selected record across both closed variants and proves Verify has no caller-selected branch；the second completes multiple retained revalidations, advances `next` each time, and still audits the byte-identical immutable payload/finalization-set digest while rejecting an active revalidation or broken lineage.

Also add `TestNativeRunnerCompletionRevalidationAbortClosesGeneration`, `TestNativeRunnerCompletionRevalidationOppositeOutcomeReturnsConflict`, `TestNativeRunnerCompletionRevalidationPendingUsesNoHandleInnerFinalization`, `TestNativeRunnerCompletionRevalidationPostTerminalResponseLossAudit`, `TestNativeRunnerCompletionRevalidationOutcomeDeliveryMustAckBeforeNextOpen`, `TestNativeRunnerCompletionTrustedTimeHeadNeverRollsBackAcrossAbortedLineage`, `TestNativeRunnerCompletionTerminalAuditSurvivesAbortedRevalidations`, `TestCompletionValidationChildRuntimeBindsCoordinatorInnerAndInstalledProjection`, `TestCompletionValidationChildRuntimeResultIsBoundedSealedAndSingleUse`, `TestCompletionValidationChildTrackedTreeReaderIsPathlessAndExact`, `TestCompletionValidationChildRuntimeHasNoEvidenceSemanticDependency`, `TestCompletionValidationChildRuntimeOwnedOnlyByCompositePackage`, `TestCompletionInnerFinalizersRequireExactLeafReceipt`, `TestCompletionInnerFinalizationReceiptCrashSeams`, `TestCompletionRetainedOutcomeConsumesSameCleanupReceipt`, `TestNativeRunnerWindowsRecoveryDispatchIsClosedAndGuarded`, `TestNativeRunnerPlatformResumeTransitionIsGuardedAndOneUse`, `TestNativeRunnerPlatformResumeManifestUsesSinkBoundAttestor`, `TestNativeRunnerPlatformPhaseBUsesPathlessSuccessorNotOpen`, `TestNativeRunnerPlatformValidationDispatchSeparatesFreshFromAmbiguousExchange`, `TestNativeRunnerPlatformPromotionRecoveryDispatchIsClosedAndGuarded`, `TestNativeRunnerPlatformResumeDowngradeCannotRegainWorkloadAuthority`, `TestNativeRunnerPlatformTrustedTimeExchangeCrashRecoveryNeverRefetches`, `TestNativeRunnerPlatformSuccessCleanupMintsPublicationOnlyAfterAbsence`, `TestNativeRunnerPlatformAbortCleanupCannotMintPublication`, `TestNativeRunnerPlatformCleanupFinalizationReceiptCrashSeams`, `TestNativeRunnerPlatformPostCleanupAttestationResponseLossNeverResigns` and `TestNativeRunnerCleanupFinalizationFirstSelectedOriginWins`. They cover failed/provider-rejected and crash-recovered revalidation, exact-clean-before-abort, exactly one outcome-class success、selector-free no-handle `finalization_pending` recovery、post-terminal delivery ACK、latest valid trusted-time-head monotonicity across completed/aborted edges, immutable digest preservation, sealed fresh/recovery runtime binding, generated nonce、bounded result/tree reads, receipt-bound two-CAS leaf finalization, the one legal cross-package caller and guarded Windows plus platform closed-dispatch/exact-clean/post-clean publication routes. The platform matrix distinguishes no-intent fresh validation from intent-present recovery and never refetches an ambiguous exchange. Task 7 adds the sender-state recovery test after `channel.go` exists.

Also add `TestNativeRunnerCompletionRevalidationDispatchIsClosedAndGuarded`, `TestNativeRunnerAuthoritySuccessCleanupMintsPublicationOnlyAfterAbsence`, `TestNativeRunnerAuthorityAbortCleanupCannotMintPublication`, `TestNativeRunnerAuthorityCleanupFailureCannotReachAttestors`, `TestNativeRunnerAuthorityActiveCrashDefaultsToAbortCleanup`, `TestNativeRunnerAuthorityRecoveryDispatchIsClosedAndGuarded`, `TestNativeRunnerAuthorityCleanupRecoveryIsSelectorFree`, `TestNativeRunnerAuthorityCleanupFinalizationReceiptCrashSeams`, `TestNativeRunnerAuthorityPostCleanupAttestationResponseLossNeverResigns` and `TestNativeRunnerAuthorityFinalizationFirstSelectedOriginWins`. They freeze the non-authorizing completion and authority dispatches, authority first-origin-wins/default-abort rule, exact receipt/absence barrier and response-loss-safe one-shot attestations. Task 7 owns the channel/sender continuation.

Task 7 adds `TestB10ClosedHostGateExactCoversDeclaredTests`; it is intentionally neither defined nor run during Task 2 GREEN. Once added, it uses `go list` plus `go/parser`/`go/types` over the literal eight B10 gate manifests—the common post-clean/receipt/dispatch/key-bootstrap gate、cleanup-adapter/staged-diagnostic contract gate、completion inner-bootstrap crash gate、closed launcher/containment gate、full OS-neutral closed-set gate、Windows wrapper contract gate and the separate Windows/Linux production-adapter target gates—requires every named `Test*` to have exactly one owning definition and rejects `Skip`、`Skipf` or `SkipNow` in each test plus every package-local helper transitively reachable from it. The cleanup/staged manifest freezes both declaration-bound adapter roots and both guarded staged-diagnostic recovery roots. The wrapper manifest exact-covers every top-level `TestPowerShellWrapper*|TestRunnerHelperWindowsFinalScopes*|TestGitBashWrapper*|TestWrapperParity*` definition against its six-name literal owner table. The completion manifest freezes the session-free inner-key crash matrix；the two target manifests each include the matching OS-tagged state-guard、launcher-adapter and protected-key production-adapter roots plus the meta-test. A missing/renamed/duplicate/extra prefixed test、cross-OS zero-match or weakened literal manifest therefore fails instead of becoming a successful gate. Each matching-host runner independently checks JSON events：every declared top-level name must emit exactly one `run`、one `pass` and zero `skip`, and any descendant `TestParent/...` skip event fails.

Task 7 adds and finalizes this meta-test with a committed independent owner specification, `testdata/c12/b10-exact-owner-map.v1.json`, strict schema `talenro-c12-b10-exact-owner-map/v1`. Its manifest IDs are exactly ordered `common|cleanup_contract|completion_bootstrap|launcher|full|windows_wrapper|windows_adapter|linux_adapter`; their member counts including the shared meta root are exactly `43|5|2|32|76|7|4|4`, and their duplicate-free union is exactly 147 roots (one meta plus 146 substantive). The top level contains exactly `{SchemaVersion,ManifestOrder,Entries}`；entries are bytewise ordered by `Package|OwnerFile|Test`, each entry is exactly `{ManifestIDs,GOOS,BuildTags,Package,OwnerFile,Test}`, and each `ManifestIDs` subsequence follows `ManifestOrder`. Membership is the exact join of the eight literal member lists below with the following normative owner table. `GOOS` is empty except the six Windows/Linux state-guard、launcher and protected-key substantive entries, where it is exactly `windows` or `linux`；`BuildTags` is always the empty array because these use native OS constraints rather than optional tags. No path may be inferred from the current AST:

```text
PACKAGE	OWNER_FILE	TESTS
./internal/c12runnerprofile	internal/c12runnerprofile/profile_test.go	TestB10ClosedHostGateExactCoversDeclaredTests|TestNativeRunnerProfileStrictFiveRoleTokens|TestStagedDiagnosticProfileIsNonpublishing
./internal/c12cleanup	internal/c12cleanup/core_test.go	TestCleanupCoreAliasesAreAssignmentIdentical|TestCleanupCoreDependencyGraphIsAcyclic|TestCleanupInspectorDeleterSetExactCoverAndNoRegistry|TestRecoverExactUsesBoundAdapterAndWritesActualBeforeDelete
./internal/c12cleanup	internal/c12cleanup/protected_key_windows_test.go	TestProtectedWALKeyWindowsProductionAdapter
./internal/c12cleanup	internal/c12cleanup/protected_key_linux_test.go	TestProtectedWALKeyLinuxProductionAdapter
./internal/c12evidence	internal/c12evidence/cleanup_test.go	TestProtectedWALKeyIntentClosesEveryPreCapsuleCrashSeam|TestProtectedWALKeyBootstrapRawRunnerprofileCallsOwnedOnlyByCleanupFacade|TestCleanupOnlyCapsuleClosesEveryPreManifestPowerLossSeam|TestCleanedTombstoneClosesKeyAndSlotCrashSeams|TestOpenCompletionInnerCleanupForExactCleanupIsSoleDescriptorConsumer
./internal/c12runnerprofile	internal/c12runnerprofile/staged_diagnostic_cleanup_test.go	TestNativeRunnerStagedDiagnosticCleanupRecoveryIsClosedAndGuarded|TestNativeRunnerStagedDiagnosticCleanupCapabilityUsesExactAdapterSet
./internal/c12runnerprofile	internal/c12runnerprofile/bootstrap_test.go	TestVerifyFixedNativeRunnerInstallationIsNonConsuming|TestVerifyFixedNativeRunnerInstallationClosesB10AndB11ProjectionVariants|TestNativeRunnerOpenFixedMachineInstallerSessionRequiresProtectedProvisioner|TestNativeRunnerProvisionOutsideDeploymentTransactionRejects|TestNativeRunnerCandidatePackageHandleIsRoleProjectionAndHostBound|TestInstallFixedNativeRunnerProfileConsumesSessionAndDestroysCandidateOnEveryReturn|TestNativeRunnerFixedLauncherAttestationBindsMachineBase
./internal/c12runnerprofile	internal/c12runnerprofile/state_guard_test.go	TestNativeRunnerLaunchGenerationsAreRoleIndexed|TestNativeRunnerSameRoleActiveRunIsExclusiveAndRearmsAfterCleanup|TestNativeRunnerRoleStateRejectsAuthenticatedRollback|TestNativeRunnerRoleStateGuardCrashSeams|TestNativeRunnerFirstInstallInitializesRoleState|TestNativeRunnerReinstallPreservesLaunchStateAndRetainedPayload
./cmd/talenro-c12-runner-launcher	cmd/talenro-c12-runner-launcher/main_test.go	TestNativeRunnerFixedProvisionModesAreRoleAndOSClosed|TestNativeRunnerBootstrapLauncherClosesFirstInstallAndReinstall|TestNativeRunnerBootstrapLauncherUsesGuardSelectedInstallRecord|TestNativeRunnerLaunchReservationCrashSeams|TestNativeRunnerLauncherDeathCannotLeaveTwoLiveParents|TestNativeRunnerRecoveryProvesBoundParentAbsentBeforeSuccessor|TestNativeRunnerCrossBootClearsPreOpenReservationBeforeFreshOpen|TestNativeRunnerCrossBootClearsExitingReservationBeforeFreshOpen|TestNativeRunnerLauncherImportClosureB10Frozen
./internal/c12runnerprofile	internal/c12runnerprofile/state_guard_windows_test.go	TestNativeRunnerWindowsStateGuardProductionAdapterRejectsFallback
./internal/c12runnerprofile	internal/c12runnerprofile/state_guard_linux_test.go	TestNativeRunnerLinuxStateGuardProductionAdapterRejectsFallback
./cmd/talenro-c12-runner-launcher	cmd/talenro-c12-runner-launcher/launcher_windows_test.go	TestNativeRunnerWindowsBootstrapLauncherProductionAdapterRejectsFallback
./cmd/talenro-c12-runner-launcher	cmd/talenro-c12-runner-launcher/launcher_linux_test.go	TestNativeRunnerLinuxBootstrapLauncherProductionAdapterRejectsFallback
./internal/c12runnerprofile	internal/c12runnerprofile/completion_state_test.go	TestCompletionInnerCleanupDescriptorIsBoundedAndSealed|TestCompletionInnerCleanupProtectedWALKeyIntentCrashSeams|TestCompletionInnerFinalizationReceiptCrashSeams|TestCompletionInnerFinalizersRequireExactLeafReceipt|TestCompletionRetainedOutcomeConsumesSameCleanupReceipt|TestNativeRunnerCompletionModePhaseCrossClaimRejects|TestNativeRunnerCompletionRetainedTransitionUsesPrepareAndCommitGuardAdvances|TestNativeRunnerCompletionRevalidationAbortClosesGeneration|TestNativeRunnerCompletionRevalidationDispatchIsClosedAndGuarded|TestNativeRunnerCompletionRevalidationOppositeOutcomeReturnsConflict|TestNativeRunnerCompletionRevalidationOutcomeDeliveryMustAckBeforeNextOpen|TestNativeRunnerCompletionRevalidationPendingUsesNoHandleInnerFinalization|TestNativeRunnerCompletionRevalidationPostTerminalResponseLossAudit|TestNativeRunnerCompletionRoleStateCoordinatorRejectsIllegalPhaseGenerationAndPayload|TestNativeRunnerCompletionRoleStateCoordinatorRequiresSessionOrFixedRecovery|TestNativeRunnerCompletionRoleStateInspectIsNonConsuming|TestNativeRunnerCompletionRoleStateUsesSingleRecoverySlot|TestNativeRunnerCompletionTerminalAuditSurvivesAbortedRevalidations|TestNativeRunnerCompletionTerminalAuditSurvivesCompletedRevalidations|TestNativeRunnerCompletionTerminalFinalizationAuditIsPathlessReadOnly|TestNativeRunnerCompletionTrustedTimeHeadNeverRollsBackAcrossAbortedLineage
./internal/c12runnerprofile	internal/c12runnerprofile/completion_child_runtime_test.go	TestCompletionValidationChildRuntimeBindsCoordinatorInnerAndInstalledProjection|TestCompletionValidationChildRuntimeHasNoEvidenceSemanticDependency|TestCompletionValidationChildRuntimeOwnedOnlyByCompositePackage|TestCompletionValidationChildRuntimeResultIsBoundedSealedAndSingleUse|TestCompletionValidationChildTrackedTreeReaderIsPathlessAndExact
./internal/c12runnerprofile	internal/c12runnerprofile/platform_resume_test.go	TestNativeRunnerPlatformAbortCleanupCannotMintPublication|TestNativeRunnerPlatformCleanupFinalizationReceiptCrashSeams|TestNativeRunnerPlatformPhaseBUsesPathlessSuccessorNotOpen|TestNativeRunnerPlatformPostCleanupAttestationResponseLossNeverResigns|TestNativeRunnerPlatformPostCleanupPublicationRecoversExactSenderState|TestNativeRunnerPlatformPromotionRecoveryDispatchIsClosedAndGuarded|TestNativeRunnerPlatformResumeDowngradeCannotRegainWorkloadAuthority|TestNativeRunnerPlatformResumeManifestUsesSinkBoundAttestor|TestNativeRunnerPlatformResumeSuccessorUsesSameGeneration|TestNativeRunnerPlatformResumeTransitionIsGuardedAndOneUse|TestNativeRunnerPlatformSuccessCleanupMintsPublicationOnlyAfterAbsence|TestNativeRunnerPlatformTrustedTimeExchangeCrashRecoveryNeverRefetches|TestNativeRunnerPlatformValidationDispatchSeparatesFreshFromAmbiguousExchange|TestNativeRunnerCleanupFinalizationFirstSelectedOriginWins
./internal/c12runnerprofile	internal/c12runnerprofile/authority_cleanup_test.go	TestNativeRunnerAuthorityAbortCleanupCannotMintPublication|TestNativeRunnerAuthorityActiveCrashDefaultsToAbortCleanup|TestNativeRunnerAuthorityCleanupFailureCannotReachAttestors|TestNativeRunnerAuthorityCleanupFinalizationReceiptCrashSeams|TestNativeRunnerAuthorityCleanupRecoveryIsSelectorFree|TestNativeRunnerAuthorityFinalizationFirstSelectedOriginWins|TestNativeRunnerAuthorityPostCleanupAttestationResponseLossNeverResigns|TestNativeRunnerAuthorityPostCleanupPublicationRecoversExactSenderState|TestNativeRunnerAuthorityRecoveryDispatchIsClosedAndGuarded|TestNativeRunnerAuthoritySuccessCleanupMintsPublicationOnlyAfterAbsence
./internal/c12runnerprofile	internal/c12runnerprofile/windows_cleanup_test.go	TestNativeRunnerWindowsAbortCleanupCannotMintPublication|TestNativeRunnerWindowsCleanupFailureCannotReachAttestors|TestNativeRunnerWindowsCleanupFinalizationReceiptCrashSeams|TestNativeRunnerWindowsNextGenerationBlockedUntilCommittedPublication|TestNativeRunnerWindowsPostCleanupAttestationResponseLossNeverResigns|TestNativeRunnerWindowsPostCleanupPublicationRecoversExactSenderState|TestNativeRunnerWindowsRecoveryDispatchIsClosedAndGuarded|TestNativeRunnerWindowsSuccessCleanupMintsPublicationOnlyAfterAbsence|TestWindowsRawRunnerprofileCallsOwnedOnlyByFinalScopesFacade
./internal/c12runnerprofile	internal/c12runnerprofile/channel_test.go	TestCompletionCoordinatorHasNoRawMutationSurface|TestCompletionCoordinatorRejectsCallerDefinedPermit|TestCompletionPublicationSinkBindsExactCoordinatorAndSet|TestCompletionPublicationSinkHasNoCallerPathKindOrDigest|TestCompletionPublicationSinkRejectsViewManifestBindingMismatch|TestCompletionPublicationSinkStrictlyDecodesFrozenManifestProjection|TestEvidenceChannelAcknowledgePersistsOuterActualBeforeReturn|TestEvidenceChannelAdoptCommittedIsNonEnumerable|TestEvidenceChannelAdoptionSetBindsCompletionPublishPending|TestEvidenceChannelAdoptionSetIsDurableBeforeFirstAdopt|TestEvidenceChannelBindingMutationCrossBindsBeforeReturn|TestEvidenceChannelBindingMutationRejectsStaleStateEpoch|TestEvidenceChannelBindingRejectsBusinessDriftWithFreshStateEpoch|TestEvidenceChannelBindingSurvivesExecutionReservationOnlyCAS|TestEvidenceChannelBindingSurvivesInnerCleanupOnlyCAS|TestEvidenceChannelCrashSeams|TestEvidenceChannelDirectMutationCannotBypassCoordinator|TestEvidenceChannelFinalizeReadyBindsImmutableRetainedPayloadDigest|TestEvidenceChannelFinalizeReadyHasNoDigestCycleOrPrematureCell|TestEvidenceChannelPublicationProgressCrossPersistsPublishBinding|TestEvidenceChannelReciprocalDigestGraphIsAcyclic|TestEvidenceChannelRecoverAcknowledgeReadsAuthenticatedPendingActual|TestEvidenceChannelRecoverAdoptionSetClosesEveryCrashSeam|TestEvidenceChannelRecoverFinalizationBeforeAndAfterGuardCAS|TestEvidenceChannelRecoverPublicationProgressClosesCrashSeams|TestEvidenceChannelSeparateSenderReceiverCapabilities|TestEvidenceChannelSetActualDigestNeverFeedsPublishBinding|TestNativeRunnerDirectSenderBindRejectsAllActiveSessions
./cmd/talenro-c12-runner-helper	cmd/talenro-c12-runner-helper/main_test.go	TestCaptureStagedAndBuildClosedSetCleanupBarrier|TestNativeVerifierParentCoreSmokeFD3AdoptionSequence|TestNativeVerifierParentRejectsCoreSmokeHandoffBypass|TestRunnerHelperContainerVerifierIngressDoesNotRequireMachineLauncher|TestRunnerHelperWindowsFinalScopesClosedCommandSurface
./internal/c12evidence	internal/c12evidence/runnerhelper_test.go	TestB10FixtureComparatorUsesRunMaterializer|TestRunnerHelperCoreSmokeFD3Only|TestTreeFactsCapsuleStrictValidation|TestWindowsFinalScopesSealedProducerPathIsCallableAndSoleOwned|TestWindowsOuterParentCannotReachCoreSmokeSlotOrTrustedSource
./internal/artifactscan	internal/artifactscan/buildinputs_test.go	TestFixedOperationsBuildPolicySchemaAndAcyclic|TestFixedOuterImageBuildContextSchemaAndExactCover
./internal/c12evidence	internal/c12evidence/compose_test.go	TestOuterImageBuildContextsAreClosedAndAcyclic
./internal/c12evidence	internal/c12evidence/powershell_wrapper_test.go	TestPowerShellWrapperFullLifecycleCleanupAndZeroPrematurePublication|TestPowerShellWrapperUsesLockedSnapshotIdentityAndNoAmbientTools
./internal/c12evidence	internal/c12evidence/bash_wrapper_test.go	TestGitBashWrapperFullLifecycleParityAndTamperHasZeroChannelSideEffect|TestGitBashWrapperUsesLockedSnapshotIdentityAndNoAmbientTools
./internal/c12evidence	internal/c12evidence/wrapper_contract_test.go	TestWrapperParityExactCoversPowerShellAndGitBashStages
```

`TestB10ClosedHostGateExactCoversDeclaredTests` is itself fixed at `internal/c12runnerprofile/profile_test.go`. Its Go source independently hard-codes the exact grouped table above、the eight IDs/counts and all manifest memberships, strict-decodes the JSON and exact-compares every field before scanning source. It then parses each fixed owner file independent of host/build tag, requires exactly one top-level definition of every listed test at that exact package/file, rejects a same-name definition elsewhere and applies the transitive `go/types` no-skip/unresolved-dispatch rule from the declared owner. `scripts/invoke-exact-b10-gates.ps1` is a no-argument/no-environment-override committed runner with its own literal copies of the first seven manifest tables；`scripts/invoke-exact-b10-linux-adapter-gate.sh` is the no-argument Linux runner with an independent literal eighth table and no `jq`/optional skip path. The Go meta exact-compares both script literals to the JSON and its own table；each runner exact-compares `go test -list` and JSON run/pass/root-or-descendant-skip events. Expected names never come from the tested regex、actual list or a caller argument. A changed count with equal union、test move、package move、owner-file splice、coordinated map-only deletion、zero-match OS tag or manifest membership drift is therefore RED.
Also add `TestNativeRunnerOpenFixedMachineInstallerSessionRequiresProtectedProvisioner`, `TestNativeRunnerProvisionOutsideDeploymentTransactionRejects`, `TestNativeRunnerCandidatePackageHandleIsRoleProjectionAndHostBound`, `TestInstallFixedNativeRunnerProfileConsumesSessionAndDestroysCandidateOnEveryReturn`, `TestNativeRunnerFixedProvisionModesAreRoleAndOSClosed`, `TestNativeRunnerBootstrapLauncherClosesFirstInstallAndReinstall`, `TestNativeRunnerBootstrapLauncherUsesGuardSelectedInstallRecord`, `TestNativeRunnerLaunchReservationCrashSeams`, `TestNativeRunnerLauncherDeathCannotLeaveTwoLiveParents`, `TestNativeRunnerRecoveryProvesBoundParentAbsentBeforeSuccessor`, `TestNativeRunnerCrossBootClearsPreOpenReservationBeforeFreshOpen`, `TestNativeRunnerCrossBootClearsExitingReservationBeforeFreshOpen`, `TestNativeRunnerCompletionRoleStateCoordinatorRequiresSessionOrFixedRecovery`, `TestNativeRunnerCompletionRoleStateCoordinatorRejectsIllegalPhaseGenerationAndPayload`, `TestNativeRunnerCompletionRoleStateInspectIsNonConsuming`, `TestNativeRunnerWindowsStateGuardProductionAdapterRejectsFallback`, `TestNativeRunnerLinuxStateGuardProductionAdapterRejectsFallback`, `TestNativeRunnerWindowsBootstrapLauncherProductionAdapterRejectsFallback` and `TestNativeRunnerLinuxBootstrapLauncherProductionAdapterRejectsFallback`. The coordinator matrix calls every forbidden backward/skip/reuse semantic operation, changes retained immutable payload, supplies an old/fabricated view、caller-defined permit、raw capability bytes or the wrong installed process/mode, proves no raw store/CAS/business setter is exported and requires zero candidate write/guard advance. The installer/launcher matrix proves only the matching fixed provision mode running inside a live authenticated OS deployment transaction with its nonexportable role/projection/host-bound candidate handle can mint the session；ordinary-shell direct launch, handle replay/cross-host/cross-role substitution and a caller-defined interface fail before store access. `InstallFixedNativeRunnerProfile` consumes the session and destroys its handle on every success/error/response-loss path before returning, so Verify follows with no surviving installer capability and reuse always rejects. It begins with no role helper installed, completes generation 1 through the base launcher, then injects every reinstall and launcher/child/containment/reservation seam and proves the launcher starts the old exact helper before guard advance and the new exact helper afterward. It kills the launcher before/after child create、reservation mutation、Open、outer semantic permit、provider/private-child spawn and terminal cleanup, then requires the bound Job/cgroup empty before one pre-terminal same-generation recovery successor can execute. The two cross-boot tests reboot at reservation→Open and terminal→reservation-clear respectively；they require an absence-proved reservation-only clear in terminal-inactive business state followed by a distinct new reservation before any fresh Open or terminal audit. Direct/stale versioned-helper execution、stable helper symlink、launcher-as-role-receipt、candidate path/argv/env、wrong role/mode/PID/start/containment token、concurrent launcher、escaped descendant and post-open replacement all fail before state/session access or leave at most one live authoritative parent.

Profile RED vectors additionally remove/empty/duplicate `BootstrapLauncherCatalogRole`, alias it to `HelperCatalogRole`, swap the Windows/Linux launcher roles or put either launcher in `RequiredToolRoles`/`RequiredReceiptRoles`. Launcher-attestation vectors mutate schema、role、path、digest、OS/arch、file/machine-image identity、application-control/toolchain projection and base-authority signature；an absent launcher, helper receipt presented as launcher proof or same bytes copied to another machine must fail before candidate/session/store access.

Add `TestNativeRunnerFixedLauncherAttestationBindsMachineBase` for the exact external-baseline schema/domain and machine-copy rejection. The launcher source/build identity remains a locked supply-chain input；the machine-local signed attestation remains untracked and is never accepted as a role install receipt.

Add `TestNativeRunnerPlatformRecoveryUsesAttestedBootIdentity`: Phase A reservation → controlled reboot → guard-selected resume successor → Phase B is positive；boot-ID splice、PID/start/cgroup reuse and any cross-boot kill are negative. It also proves the fixed launcher never converts a new boot into a fresh Open for that generation.

- [ ] **Step 2: Run runner-profile tests and verify RED**

Run: `go test ./internal/c12runnerprofile ./cmd/talenro-c12-runner-launcher -run '^TestNativeRunner|^TestStagedDiagnostic|^TestTrackedSnapshotScript|^TestInstallFixedNativeRunnerProfile|^TestVerifyFixedNativeRunnerInstallation|^TestOpenInstalledNativeRunnerSession' -count=1`

Expected: FAIL because the package, fixed token and sealed session do not exist.

- [ ] **Step 3: GREEN — implement neutral profile/bootstrap and create both B10 policy instances**

Implement the strict schemas, canonical domains, atomic fixed-profile installer, exact non-consuming `VerifyFixedNativeRunnerInstallation`, compile-time fixed-store adapters, per-role rollback-resistant guard+A/B-cell protocol, role-indexed independent install/launch generations, same-role active-run exclusion, terminal rearm, Phase-A→resume-only successor transition and opaque one-run session above. The production adapter must use a fixed TPM/vTPM-backed or equivalently non-rollbackable platform `CompareAndAdvance(expectedEpoch, expectedDigest, newEpoch, newDigest)` provider and expose no file/DPAPI-only fallback；tests use only the package-private deterministic guard fake and exercise the identical state algorithm. Create `testdata/c12/windows/runner-profile.v1.json` as the canonical policy-only `windows_scopes` instance with the exact tracked PowerShell/Git-Bash tokens, closed Git/Go/Docker/Compose/cygpath/scanner/helper roles, recovery/external-root purposes, exactly one `C12NativeRunnerChannelBindingV1{TransactionKind:"windows_scopes",Access:"sender",SlotPurpose:"windows_scopes_next",SequencePolicy:"next"}` and no machine path/digest, secret, run/output identity or future nonce. Create `testdata/c12/diagnostic/runner-profile.v1.json` as the canonical `staged_diagnostic` instance with no scripts, no channel bindings, no provider/evidence/signing/publishing purpose, and only the Linux static helper/Git/reproducible-build roles plus canonical object-database/external-root/recovery/install policies. Receipt policies exact-cover only native executables；final profiles authenticate their tracked scripts through the committed snapshot derivation contract. Tests use package-private in-memory adapters but production exposes only protected-session `InstallFixedNativeRunnerProfile`, non-consuming `VerifyFixedNativeRunnerInstallation(ctx, role)` and pathless `OpenInstalledNativeRunnerSession(ctx, role)`; none permits injection.

`state_guard.go` owns the private provider interface, guarded A/B algorithm and test-only fake injection point. `state_guard_windows.go` has literal first line `//go:build windows` and binds only the fixed Windows TPM/TBS-backed application-control provider；`state_guard_linux.go` has literal first line `//go:build linux` and binds only the fixed TPM/vTPM resource-manager/provider identity. Neither production file imports or reads environment/config/default paths, and no non-test setter/factory exists. Their same-tag tests compile the real adapter, validate provider attestation/nonresettable policy and reject software/file/registry/DPAPI-only fallback, wrong machine/provider identity and cross-role counter use without consuming a production role guard. `state_guard_test.go` owns deterministic crash/CAS model tests only. `cmd/talenro-c12-runner-launcher` is the distinct machine-base dispatcher；its Windows/Linux tagged adapters accept only OS deployment/inherited protected handles and implement the five fixed provision modes plus the closed role-specific public-run modes. Unsupported role/OS modes are absent from help and reject before installer/session/store access. `cmd/talenro-c12-runner-helper` contains only the versioned public-parent/private-child implementations selected by an authenticated launcher handoff and cannot provision itself or dispatch without that handoff.

The Windows policy instance sets `BootstrapLauncherCatalogRole=c12_runner_launcher_windows_amd64`; the diagnostic instance sets `BootstrapLauncherCatalogRole=c12_runner_launcher_linux_amd64`. Both canonical digests cover that field. Build the launcher PE/ELF twice from independent absolute roots, require byte-identical identities, create their SBOM/provenance inputs and bind them into the toolchain native-role projection consumed later by B10. The launcher tagged adapters validate only the OS baseline attestation and protected deployment/launch handles；the helper's machine-native dispatcher validates only launcher-inherited bindings. The nested container-verifier and task-local fixture dispatchers remain separately authenticated and cannot call a machine role store.

Run: `go test ./internal/c12runnerprofile ./cmd/talenro-c12-runner-launcher -count=1`

Expected: PASS; copied profile bytes at a nonfixed token, a valid native install receipt for another absolute file identity, any script-as-receipt or snapshot-root/blob/mode/file-identity splice, another role's sender/receiver/provider capability, authenticated old-cell/install-record replay、same-epoch digest fork、guard reset/cross-role swap、sealed install/launch generation rollback and every ambient override fail before an opaque session is returned. Every injected reinstall candidate/fsync/guard crash seam resolves to exactly the guard-selected old or new state without quarantine；a first-install pre-advance crash remains authenticated `(0,zero)` and uninstalled/retryable, while only tampered/missing selected bytes quarantine. Successful first install yields guarded epoch-1 inactive state, reinstall never overwrites the selected native files and cannot reset next generation or immutable retained payload, and repeated `Verify` is byte-identical/non-consuming. Only a same-role concurrent active parent is rejected, cross-role sessions remain independent, terminal cleanup rearms the next generation, and platform Phase B uses only the same-run resume successor.

The package result must also prove an absent/wrong launcher attestation or launcher-profile role fails before first install, generation 1 succeeds with no preinstalled role helper, a reinstall uses the old helper before guard advance and the new helper afterward, and the selected helper cannot be launched directly. The fixed container verifier still passes its capsule-bound ingress without a machine launcher and has zero reachability to a role store.

- [ ] **Step 4: RED — add hash-chain and intent-before-create tests**

```go
func TestWALRequiresFsyncedIntentBeforeActual(t *testing.T) {
	w := newTestWAL(t)
	resource := ResourceRef{ResourceDeclarationDigest: testResourceDeclarationDigest(), Type: ResourceContainer, StableName: "talenro-c12-00112233445566778899aabbccddeeff-verifier"}
	if err := w.AppendActual(resource, "sha256:actual", testCreatedAt()); !errors.Is(err, ErrMissingIntent) {
		t.Fatal("WAL accepted actual without intent")
	}
	if err := w.AppendIntent(resource, expectedImageDigest(), testCreatedAt()); err != nil {
		t.Fatal("WAL rejected intent")
	}
	if err := w.AppendActual(resource, "sha256:actual", testCreatedAt().Add(time.Second)); err != nil {
		t.Fatal("WAL rejected matching actual")
	}
}
```

Add `TestTreeFactsCapsuleStrictValidation` vectors for every field, unknown/duplicate/null/defaulted member, zero/replayed/wrong independently expected nonce, identical signed bytes plus identical policy/session method-called twice, copied policy/nonce with zero-value or caller-fabricated session, cross-parent-session replay, expired/non-UTC interval, wrong runner trust root/EKU/domain, scope-attestation-as-capsule, capsule-attestation-as-scope, snapshot/file-identity、minimal-object-database identity/closure digest、Git locator/canonical digest、capture-kind、WAL/run splice and same signed bytes at another path. The method must return zero handle+error before snapshot/object-database open or FD creation except that one valid call atomically consumes the sealed parent-held challenge as it yields the opaque handle；a nil-error/boolean authorization cannot construct private-child input, and a valid capsule's attestor identity and complete envelope digest must change `runner_facts_digest`, with neither value caller-selectable.

- [ ] **Step 5: Run WAL tests and verify RED**

Run: `go test ./internal/c12cleanup ./internal/c12evidence -run '^TestWAL|^TestCleanupCore|^TestTreeFactsCapsuleStrictValidation$' -count=1`

Expected: FAIL because `OpenSelectedOwnershipWAL` and the sealed record methods are undefined.

- [ ] **Step 6: GREEN — implement the canonical WAL record**

```go
package c12cleanup

type WALPhase string

const (
	WALPhaseIntent   WALPhase = "intent"
	WALPhaseActual   WALPhase = "actual"
	WALPhaseCleaned  WALPhase = "cleaned"
	WALPhaseNotFound WALPhase = "not_found"
)

type WALRecordV1 struct {
	SchemaVersion             string
	RunID                     uuid.UUID
	Number                    uint64
	PreviousDigest            contracts.Digest
	Phase                     WALPhase
	ResourceDeclarationDigest contracts.Digest
	ResourceType              ResourceType
	StableName                string
	ActualID                  string
	ExpectedIdentity          contracts.Digest
	CreatedAt                 time.Time
	RecordHMAC                contracts.Digest
}
```

`internal/c12cleanup/core.go` is the sole canonical/HMAC codec owner for `WALPhase` and `WALRecordV1`. `internal/c12evidence/wal.go` exposes assignment-identical aliases only:

```go
package c12evidence

type WALPhase = c12cleanup.WALPhase

const (
	WALPhaseIntent   = c12cleanup.WALPhaseIntent
	WALPhaseActual   = c12cleanup.WALPhaseActual
	WALPhaseCleaned  = c12cleanup.WALPhaseCleaned
	WALPhaseNotFound = c12cleanup.WALPhaseNotFound
)

type ResourceType = c12cleanup.ResourceType

const (
	ResourceOwnershipWAL        = c12cleanup.ResourceOwnershipWAL
	ResourceGitObjectDatabase   = c12cleanup.ResourceGitObjectDatabase
	ResourceContainer           = c12cleanup.ResourceContainer
	ResourceNetwork             = c12cleanup.ResourceNetwork
	ResourceVolume              = c12cleanup.ResourceVolume
	ResourceProcess             = c12cleanup.ResourceProcess
	ResourceWindowsJob          = c12cleanup.ResourceWindowsJob
	ResourceScheduledTask       = c12cleanup.ResourceScheduledTask
	ResourceTrackedTreeRoot     = c12cleanup.ResourceTrackedTreeRoot
	ResourceBuildRoot           = c12cleanup.ResourceBuildRoot
	ResourceRunRoot             = c12cleanup.ResourceRunRoot
	ResourceResumeBlob          = c12cleanup.ResourceResumeBlob
	ResourceExternalVM          = c12cleanup.ResourceExternalVM
	ResourceTPMNVHandle         = c12cleanup.ResourceTPMNVHandle
	ResourcePITRDatabase        = c12cleanup.ResourcePITRDatabase
	ResourcePITRTimeline        = c12cleanup.ResourcePITRTimeline
	ResourceProviderCredential = c12cleanup.ResourceProviderCredential
	ResourceCertificate        = c12cleanup.ResourceCertificate
	ResourceBoundedRunPath      = c12cleanup.ResourceBoundedRunPath
)

type ResourceRef = c12cleanup.ResourceRef
type WALRecordV1 = c12cleanup.WALRecordV1
```

Linux cross-reboot ownership also freezes a separate pre-resource recovery authority. The remaining declarations in the next block are implemented in dependency-leaf `internal/c12cleanup`; `internal/c12evidence/cleanup.go` re-exports the data types as exact aliases and exposes only the named forwarding functions, preserving the existing public package while ensuring `internal/c12runnerprofile` can consume the sealed core without importing c12evidence:

```go
package c12cleanup

type ResourceType string

const (
	ResourceOwnershipWAL        ResourceType = "ownership_wal"
	ResourceGitObjectDatabase   ResourceType = "git_object_database"
	ResourceContainer           ResourceType = "container"
	ResourceNetwork             ResourceType = "network"
	ResourceVolume              ResourceType = "volume"
	ResourceProcess             ResourceType = "process"
	ResourceWindowsJob          ResourceType = "windows_job"
	ResourceScheduledTask       ResourceType = "scheduled_task"
	ResourceTrackedTreeRoot     ResourceType = "tracked_tree_root"
	ResourceBuildRoot           ResourceType = "build_root"
	ResourceRunRoot             ResourceType = "run_root"
	ResourceResumeBlob          ResourceType = "resume_blob"
	ResourceExternalVM          ResourceType = "external_vm"
	ResourceTPMNVHandle         ResourceType = "tpm_nv_handle"
	ResourcePITRDatabase        ResourceType = "pitr_database"
	ResourcePITRTimeline        ResourceType = "pitr_timeline"
	ResourceProviderCredential ResourceType = "provider_credential"
	ResourceCertificate        ResourceType = "certificate"
	ResourceBoundedRunPath      ResourceType = "bounded_run_path"
)

type ResourceRef struct {
	ResourceDeclarationDigest contracts.Digest
	Type                      ResourceType
	StableName                string
}

type CleanupOnlyResourceDeclarationV1 struct {
	ResourceType                 ResourceType
	StableParentIdentityDigest   contracts.Digest
	StableNameOrReservation      string
	ExpectedProviderIdentity     contracts.Digest
	ExpectedCreationPolicyDigest contracts.Digest
}

type CleanupOnlyResourceReservationV1 struct {
	ResourceType                 ResourceType
	StableParentIdentityDigest   contracts.Digest
	StableNameOrReservation      string
	ExpectedCreationPolicyDigest contracts.Digest
}

type CleanupOnlyOwnershipPlanTemplateV1 struct {
	OwnershipWALReservationCoreDigest contracts.Digest
	ResourceReservations               []CleanupOnlyResourceReservationV1
}

type CleanupOnlyOwnershipPlanV1 struct {
	SchemaVersion                         string
	RunID                                 uuid.UUID
	Role                                  string
	LaunchGeneration                      uint64
	RunnerIdentity                        contracts.Digest
	ProviderIdentitySet                   contracts.Digest
	RecoveryAuthoritySlotIdentity         contracts.Digest
	MachineSealLocatorDigest              contracts.Digest
	OwnershipWALReservationCoreDigest     contracts.Digest
	CleanupPlanDigest                     contracts.Digest
	DeclaredResources                     []CleanupOnlyResourceDeclarationV1
}

func CanonicalCleanupOnlyOwnershipPlan(CleanupOnlyOwnershipPlanV1) ([]byte, error)
func CleanupOnlyOwnershipPlanBindingDigest(CleanupOnlyOwnershipPlanV1) (contracts.Digest, error)

type C12CleanupOnlyOwnershipCapsuleTemplateV1 struct {
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type C12CleanupOnlyOwnershipCapsuleV1 struct {
	SchemaVersion                 string
	RunID                         uuid.UUID
	Role                          string
	LaunchGeneration              uint64
	RunnerIdentity                contracts.Digest
	ProviderIdentitySet           contracts.Digest
	RecoveryAuthoritySlotIdentity contracts.Digest
	MachineSealLocatorDigest      contracts.Digest
	OwnershipWALReservationDigest contracts.Digest
	WALKeyProtectedHandleDigest   contracts.Digest
	CleanupPlanDigest             contracts.Digest
	DeclaredResources             []CleanupOnlyResourceDeclarationV1
	IssuedAt                      time.Time
	ExpiresAt                     time.Time
	SealAttestation               []byte
}

type SelectedOwnershipWALBinding interface {
	selectedOwnershipWALBinding() // leaf-sealed capsule/key/reservation continuation; no path/raw key/general create
}

type SelectedOwnershipWAL interface {
	selectedOwnershipWAL() // leaf-sealed exact selected reservation; no path/key/general filesystem authority
	AppendIntent(ResourceRef, contracts.Digest, time.Time) error
	AppendActual(ResourceRef, string, time.Time) error
	AppendCleaned(ResourceRef, time.Time) error
	AppendNotFound(ResourceRef, time.Time) error
}

type ProtectedWALKeyHandle interface {
	protectedWALKeyHandle() // leaf-sealed nonexportable per-run key handle
}

type ProtectedWALKeyIntent interface {
	protectedWALKeyIntent() // leaf-sealed same-process create-or-recover authority for one guard-selected backend reservation
}

type ProtectedWALKeyBindingCoreV1 struct {
	SchemaVersion                 string
	RunID                         uuid.UUID
	Role                          string
	LaunchGeneration              uint64
	RecoveryAuthoritySlotIdentity contracts.Digest
	MachineRecoveryAuthorityDigest contracts.Digest
	BackendObjectReservationDigest contracts.Digest
}

type ProtectedWALKeyBindingV1 struct {
	Core                           ProtectedWALKeyBindingCoreV1
	BootstrapIntentDigest          contracts.Digest
	BindingDigest                 contracts.Digest
	Attestation                   []byte
}

type CleanupOnlyPolicy struct {
	ExpectedRunID                         uuid.UUID
	ExpectedRole                          string
	ExpectedLaunchGeneration              uint64
	ExpectedRunnerIdentity                contracts.Digest
	ExpectedProviderIdentitySet           contracts.Digest
	ExpectedRecoveryAuthoritySlotIdentity contracts.Digest
	ExpectedCapsuleEnvelopeDigest          contracts.Digest
	ExpectedOwnershipWALReservationDigest contracts.Digest
	ExpectedWALKeyProtectedHandleDigest   contracts.Digest
	ExpectedCleanupPlanDigest             contracts.Digest
	ExpectedDeclaredResourceSetDigest     contracts.Digest
	ExpectedMachineSealLocatorDigest      contracts.Digest
}

func CanonicalCleanupOnlyPolicyBinding(CleanupOnlyPolicy) ([]byte, error)
func CleanupOnlyPolicyBindingDigest(CleanupOnlyPolicy) (contracts.Digest, error)

type ExactResourceInspectionStatus string

const (
	ExactResourceMatched  ExactResourceInspectionStatus = "matched"
	ExactResourceNotFound ExactResourceInspectionStatus = "not_found"
)

type ExactResourceAdapterBindingV1 struct {
	ResourceDeclarationDigest     contracts.Digest
	ResourceType                  ResourceType
	ExpectedProviderIdentity      contracts.Digest
	ExpectedCreationPolicyDigest contracts.Digest
	ImplementationIdentityDigest contracts.Digest
	MachineSealLocatorDigest     contracts.Digest
}

type ExactResourceInspectionV1 struct {
	Status                    ExactResourceInspectionStatus
	ObservedIdentity          string
	ObservedIdentityDigest    contracts.Digest
	ObservedCreationTime      time.Time
	InspectionBindingDigest   contracts.Digest
}

type ExactResourceInspectorDeleter interface {
	AdapterBinding() ExactResourceAdapterBindingV1
	InspectExact(context.Context, CleanupOnlyResourceDeclarationV1, string) (ExactResourceInspectionV1, error)
	DeleteExact(context.Context, CleanupOnlyResourceDeclarationV1, string, ExactResourceInspectionV1) error
}

type CleanupInspectorDeleterSet interface {
	cleanupInspectorDeleterSet() // leaf-sealed exact declaration-to-production-adapter cover
}

func BindCleanupInspectorDeleterSet([]byte, CleanupOnlyPolicy, []ExactResourceInspectorDeleter) (CleanupInspectorDeleterSet, error)

type Capability interface {
	cleanupOnlyCapability() // leaf-sealed exact inspect/delete/finalize authority only
}

type TerminalProof interface {
	terminalCleanupProof() // leaf-sealed, capability/run/terminal-WAL/absence-set bound
}

type FinalizationReceipt interface {
	cleanupFinalizationReceipt() // leaf-sealed prepared exact-clean receipt; one terminal commit
}

type FinalizationReceiptView struct {
	RunID                         uuid.UUID
	AuthorityBindingDigest        contracts.Digest
	TerminalWALDigest             contracts.Digest
	ResourceAbsenceSetDigest      contracts.Digest
	WALKeyProtectedHandleDigest   contracts.Digest
	FinalizationReceiptDigest     contracts.Digest
}

type C12CleanedOwnershipTombstoneV1 struct {
	SchemaVersion               string
	RunID                       uuid.UUID
	PredecessorAuthorityDigest  contracts.Digest
	TerminalWALDigest           contracts.Digest
	ResourceAbsenceSetDigest    contracts.Digest
	WALKeyProtectedHandleDigest contracts.Digest
	CleanedAt                   time.Time
	Attestation                 []byte
}

func OpenProtectedWALKeyIntent(ProtectedWALKeyBindingV1) (ProtectedWALKeyIntent, error)
func CreateOrRecoverProtectedWALKey(context.Context, ProtectedWALKeyIntent) (ProtectedWALKeyHandle, error)
func DestroyProtectedWALKeyIntent(context.Context, ProtectedWALKeyIntent) error
func SealOwnershipCapsule(C12CleanupOnlyOwnershipCapsuleV1, ProtectedWALKeyHandle) ([]byte, contracts.Digest, SelectedOwnershipWALBinding, error)
func MarshalSelectedOwnershipWALBinding(SelectedOwnershipWALBinding) ([]byte, error)
func RecoverSelectedOwnershipWALBinding([]byte, CleanupOnlyPolicy) (SelectedOwnershipWALBinding, error)
func OpenSelectedOwnershipWAL(context.Context, SelectedOwnershipWALBinding, CleanupOnlyPolicy) (SelectedOwnershipWAL, error)
func OpenOwnershipCapsuleForExactCleanup([]byte, CleanupOnlyPolicy, CleanupInspectorDeleterSet) (Capability, error)
func RecoverExact(context.Context, Capability, ResourceRef) error
func Execute(context.Context, Capability) (TerminalProof, error)
func PrepareFinalization(Capability, TerminalProof) (FinalizationReceipt, error)
func InspectFinalizationReceipt(FinalizationReceipt) (FinalizationReceiptView, error)
func RecoverFinalization(context.Context, FinalizationReceiptView) (FinalizationReceipt, error)
func CommitFinalization(FinalizationReceipt) error
```

`CanonicalCleanupOnlyPolicyBinding` strict-validates every nonzero policy member and returns JCS of exactly `{SchemaVersion:"talenro-c12-cleanup-only-policy-binding/v1",RunID,Role,LaunchGeneration,RunnerIdentity,ProviderIdentitySet,RecoveryAuthoritySlotIdentity,CapsuleEnvelopeDigest,OwnershipWALReservationDigest,WALKeyProtectedHandleDigest,CleanupPlanDigest,DeclaredResourceSetDigest,MachineSealLocatorDigest}` with no `Expected` prefixes in the wire names. `CleanupOnlyPolicyBindingDigest` is exactly `SHA-256(ASCII("TALENRO-C12-CLEANUP-ONLY-POLICY-BINDING-V1") || 0x00 || CanonicalCleanupOnlyPolicyBinding(policy))`. These are the sole leaf implementations：Seal derives the policy from the authenticated capsule/selected plan, while every descriptor and recovery path recomputes through the same functions；no facade may hand-roll or accept this digest.

`CleanupInspectorDeleterSet` closes the runtime adapter dependency without a global registry. A declaration digest is exactly `SHA-256(ASCII("TALENRO-C12-CLEANUP-RESOURCE-DECLARATION-V1") || 0x00 || JCS({ResourceType,StableParentIdentityDigest,StableNameOrReservation,ExpectedProviderIdentity,ExpectedCreationPolicyDigest}))`. `ResourceRef` and every `WALRecordV1` carry that nonzero `ResourceDeclarationDigest` plus redundant `ResourceType|StableName`; Append、recover and execute first use the digest to select exactly one capsule declaration and then exact-match the redundant type/name, so equal type+name under different stable parents never alias. `ExactResourceAdapterBindingV1` includes that nonzero `ResourceDeclarationDigest` and canonicalizes with schema `talenro-c12-exact-resource-adapter-binding/v1` and domain `TALENRO-C12-EXACT-RESOURCE-ADAPTER-BINDING-V1\x00`; therefore two same-type resources with the same provider/policy but different parent or stable reservation have distinct bindings. The cleanup plan is exactly `SHA-256(ASCII("TALENRO-C12-CLEANUP-PLAN-V1") || 0x00 || JCS(bytewise-sorted({declaration,adapter_binding_digest}) exact cover))`. `BindCleanupInspectorDeleterSet` strict-decodes the authenticated capsule, exact-matches the policy and requires a bijection of exactly one binding-identical production adapter per declaration, with no extra adapter or reusable adapter instance. It rejects missing/extra/duplicate declaration digest、type/provider/creation-policy/implementation/machine-seal drift, a local/provider role mismatch and any adapter whose binding does not recompute the capsule's `CleanupPlanDigest`；the returned set is opaque、nonserializable、capsule/policy-bound and linearly consumed by a guarded Open. Production c12evidence callers cannot supply a slice or implement an adapter. Only each of runnerprofile's five guarded cleanup Open methods, after its current-state reread, may reconstruct the fixed adapter instances from the same sealed successor/handle's pathless installed executable、provider and OS runtime capabilities and call `BindCleanupInspectorDeleterSet`; tests may use package-local fakes. Static all-build-tag tests exact-enumerate those five construction sites and every concrete production adapter implementation in runnerprofile, require each implementation to consume only a successor-private capability, and reject registration/init/global maps、locator/path lookup、reflection/plugin/dynamic dispatch or an unresolved implementation.

Leaf `OpenOwnershipCapsuleForExactCleanup` accepts that sealed set and embeds it privately in `Capability`; it never discovers or constructs an adapter. `RecoverExact(ctx,cap,ref)` accepts only the unique declaration selected by the ref's nonzero declaration digest, requires the ref and every authenticated WAL intent/actual to repeat that declaration's exact type/name, and rejects a missing/unknown digest or digest/type/name/cross-parent splice before inspection. With no actual it invokes exactly the declaration-digest-bound adapter's deterministic inspector, rejects mismatch or creation before intent, and appends+file/directory-fsyncs the recovered actual before returning；an exact absence appends terminal `not_found`. With an actual it re-inspects only that ID and never lists. `Execute` calls this same function/adapter set, exact-deletes only a matched observation, re-inspects absence and appends `cleaned|not_found`; the capability and set authorize no create/list/shell operation. Thus cold recovery reconstructs adapters inside the guard-selected successor from fixed authenticated runtime capabilities before leaf Open and needs no facade path、callback、registry or caller-selected deleter. Tests require two same-type/same-name/same-provider/same-policy resources under distinct stable parents to produce distinct declaration digests and independent WAL chains/adapters and to bind and clean successfully, while digest omission、adapter reuse、declaration swap and cross-parent/digest/type/name binding fail before inspection.

The package-private selected-WAL binding encoding is frozen as strict canonical `talenro-c12-selected-ownership-wal-binding/v1`, with `MaxSelectedOwnershipWALBindingBytes = 16 << 10`. Its unsigned projection contains exactly `{SchemaVersion,RunID,Role,LaunchGeneration,RecoveryAuthoritySlotIdentity,CapsuleEnvelopeDigest,WALKeyProtectedHandleDigest,OwnershipWALReservationDigest,DeclaredResourceSetDigest,MachineSealLocatorDigest,CleanupPolicyBindingDigest}`；it contains no raw HMAC/key、backend secret、absolute/relative path、file handle or caller option. The machine authority authenticates `TALENRO-C12-SELECTED-OWNERSHIP-WAL-BINDING-V1\x00 || JCS(unsigned)` and the bounded sealed encoding contains exactly that projection plus attestation. Seal computes `CleanupPolicyBindingDigest` only from the authenticated capsule/selected plan；Recover uses its supplied policy only as external expected state, strict-recomputes its canonical digest and constant-time compares it to the machine-authenticated binding field before any WAL open/create. Raw recovery rejects unknown/missing/null/default/duplicate/noncanonical fields、oversize/truncation、wrong domain/EKU and every machine/run/role/generation/runner/provider/slot/capsule/key/reservation/cleanup-plan/resource/policy splice；the upper guarded coordinator separately proves that the encoded capsule is still the current selected owner immediately before it calls Recover. `MarshalSelectedOwnershipWALBinding` linearly consumes/downgrades the live binding on every return；the pre-marshal value can never Open afterward. `RecoverSelectedOwnershipWALBinding` returns one live binding only for an assignment-identical selected policy, and leaf `OpenSelectedOwnershipWAL` consumes that binding on every return；response-loss replay must recover from the authenticated stored encoding and exact-reopen the same WAL, never reuse a live value or create a second recorder. The encoding is nonsecret and raw leaf verification alone is not current-selection authority. Tests exact-recompute both domains/projections, enforce the hard cap/canonical decoder and mutate every field/owner/replay seam, including each of the twelve policy members independently while preserving the capsule digest and all other inputs；upper control-flow tests additionally reject a pre-CAS candidate and the old capsule after resume adoption or terminal selection before invoking leaf Recover/Open.

```go
package c12evidence

type CleanupOnlyResourceDeclarationV1 = c12cleanup.CleanupOnlyResourceDeclarationV1
type C12CleanupOnlyOwnershipCapsuleV1 = c12cleanup.C12CleanupOnlyOwnershipCapsuleV1
type C12CleanupOnlyOwnershipCapsuleTemplateV1 = c12cleanup.C12CleanupOnlyOwnershipCapsuleTemplateV1
type CleanupOnlyPolicy = c12cleanup.CleanupOnlyPolicy
type CleanupOnlyCapability = c12cleanup.Capability
type TerminalCleanupProof = c12cleanup.TerminalProof
type CleanupFinalizationReceipt = c12cleanup.FinalizationReceipt
type C12CleanedOwnershipTombstoneV1 = c12cleanup.C12CleanedOwnershipTombstoneV1
type SelectedOwnershipWAL = c12cleanup.SelectedOwnershipWAL
type StagedDiagnosticRecoveryDispatch = c12runnerprofile.StagedDiagnosticRecoveryDispatch

type FinalClosedSetProjection interface {
	finalClosedSetProjection() // sealed bounded canonical receipt+bundle pair, bound to one authenticated native session
}

type FinalClosedSetProjectionSink interface {
	finalClosedSetProjectionSink() // opaque one-use producer ingress; callers cannot implement it
	SealValidatedClosedSet(scanReceiptCanonical, supplyChainBundleCanonical []byte) (FinalClosedSetProjection, error)
}

type ProtectedWALKeyBootstrapRecoveryDispatch string

const (
	ProtectedWALKeyBootstrapRecoveryRoleDispatch   ProtectedWALKeyBootstrapRecoveryDispatch = "role_dispatch"
	ProtectedWALKeyBootstrapRecoveryAbortUnstarted ProtectedWALKeyBootstrapRecoveryDispatch = "abort_unstarted"
	ProtectedWALKeyBootstrapRecoveryAbortIntent     ProtectedWALKeyBootstrapRecoveryDispatch = "abort_intent"
)

type ProtectedWALKeyBootstrap interface {
	protectedWALKeyBootstrap() // wraps one runnerprofile guarded continuation plus exact plan; no leaf intent/handle/binding/backend method
}

type CleanupOnlyOwnershipPlan interface {
	cleanupOnlyOwnershipPlan() // facade-sealed exact resource/name/parent/provider/WAL-reservation projection
}

type SelectedCleanupOnlyOwnership interface {
	selectedCleanupOnlyOwnership() // wraps one runnerprofile fixed-selected continuation; no raw binding/key/path/general create
}

type SelectedCleanupOnlyOwnershipView = c12runnerprofile.FixedCleanupOnlySelectionView

func BindFinalClosedSetProjectionSink(context.Context, *c12runnerprofile.AuthenticatedNativeRunnerSession) (FinalClosedSetProjectionSink, error)
func BindCleanupOnlyOwnershipPlan(context.Context, *c12runnerprofile.AuthenticatedNativeRunnerSession, c12cleanup.CleanupOnlyOwnershipPlanTemplateV1) (CleanupOnlyOwnershipPlan, error)
func BeginProtectedWALKeyBootstrap(context.Context, *c12runnerprofile.AuthenticatedNativeRunnerSession, CleanupOnlyOwnershipPlan) (ProtectedWALKeyBootstrap, error)
func CreateOrRecoverProtectedWALKeyBootstrap(context.Context, ProtectedWALKeyBootstrap) error
func SealAndSelectProtectedWALKeyBootstrap(context.Context, ProtectedWALKeyBootstrap, C12CleanupOnlyOwnershipCapsuleTemplateV1) (SelectedCleanupOnlyOwnership, error)
func InspectSelectedCleanupOnlyOwnership(SelectedCleanupOnlyOwnership) (SelectedCleanupOnlyOwnershipView, error)
func OpenSelectedOwnershipWAL(context.Context, SelectedCleanupOnlyOwnership) (SelectedOwnershipWAL, error)
func InspectProtectedWALKeyBootstrapRecoveryDispatch(context.Context) (ProtectedWALKeyBootstrapRecoveryDispatch, error)
func AbortUnstartedProtectedWALKeyBootstrapRecovery(context.Context) error
func AbortProtectedWALKeyBootstrapRecovery(context.Context) error
func InspectStagedDiagnosticCleanupRecoveryDispatch(context.Context) (StagedDiagnosticRecoveryDispatch, error)
func RecoverStagedDiagnosticCleanup(context.Context) (CleanupOnlyCapability, error)
func OpenCompletionInnerCleanupForExactCleanup(context.Context, c12runnerprofile.CompletionInnerCleanupHandle) (CleanupOnlyCapability, error)
func ExecuteCleanup(context.Context, CleanupOnlyCapability) (TerminalCleanupProof, error)
func PrepareCleanedOwnershipFinalization(CleanupOnlyCapability, TerminalCleanupProof) (CleanupFinalizationReceipt, error)
func FinalizeCleanedOwnership(context.Context, CleanupOnlyCapability, TerminalCleanupProof) error
func RecoverCleanedOwnership(context.Context) error
```

`FinalClosedSetProjectionSink` is the only import-cycle-free receipt/bundle handoff. `BindFinalClosedSetProjectionSink` carves exactly one session/run/role/tree/scanner-bound, nonserializable one-use sink from an authenticated native session without consuming its cleanup/workload authority；a second bind or cross-session/role use fails. The command package may hold the opaque sink only long enough to pass it once to `artifactscan.ValidateAndSealFinalClosedSetProjection`. Only `internal/artifactscan/scan.go`, after strict typed receipt/bundle validation、canonicalization、reference/body/digest cross-check and the parent-owned two-root equality proof, may call the exported `SealValidatedClosedSet` method with the two canonical byte projections. The method rechecks generic canonical JCS、64-KiB/4-MiB/aggregate bounds and the sink's fixed run/tree/scanner bindings, consumes the sink on every terminal return and returns a private-concrete `FinalClosedSetProjection` that exposes no bytes or method. `cmd/**` has zero direct method calls；c12evidence never imports artifactscan, and neither a byte slice、DTO、callback、boolean permit nor a caller implementation can mint the projection. The projection is data provenance only, not cleanup authority, and becomes useful only when a role-specific c12evidence semantic mint exact-matches and linearly consumes it with the same authenticated session and its own sealed unsigned/workload result.

The staged-diagnostic role is fully closed in Task 2 rather than borrowing a later final-role facade. Every restart first runs the shared three-way bootstrap router；a session-claimed active generation that crashed before the first intent CAS is consumed by `abort_unstarted`, while a selected intent is consumed by `abort_intent`, so neither may reach the staged router. Only after the shared router returns `role_dispatch` may `InspectStagedDiagnosticCleanupRecoveryDispatch` map authenticated state exactly as `cleanup_only→RecoverStagedDiagnosticCleanup` and `ordinary_finalization_pending→RecoverCleanedOwnership`; unknown、incorrectly forwarded active-without-selected-capsule、terminal、wrong reservation or any other role returns no route. The cleanup target pathlessly obtains `RecoverFixedStagedDiagnosticCleanupSuccessor`, Inspect/Describe-validates its fixed role/run/generation/slot/capsule/policy and calls only guarded one-argument `OpenFixedStagedDiagnosticCleanupCapability`. That guarded method reconstructs and binds the exact production adapter set from successor-private runtime capabilities before returning a high-level delete-only capability；the facade never receives the set. The ordinary target returns no successor/capability. Both targets recheck the same dispatch immediately before work；an error never authorizes probing or fallthrough. The live process uses the same successor route after its selected-WAL work is finished, so a reboot at every session-open/plan-bind/intent/capsule/WAL/resource/finalization seam converges without an in-memory `SelectedCleanupOnlyOwnership` and can never regain staged build/workload authority. `TestNativeRunnerStagedDiagnosticCleanupRecoveryIsClosedAndGuarded` and `TestNativeRunnerStagedDiagnosticCleanupCapabilityUsesExactAdapterSet` own this matrix, including zero key/WAL/resource counts for every pre-intent abort.

The compile-time split is exact. `internal/c12cleanup` owns every type in its blocks, including the sole `WALPhase|WALRecordV1` canonical/HMAC codec、protected-key binding/intent/handle、the sealed selected-WAL binding and recorder、`Capability`、`TerminalProof`、`FinalizationReceipt`, plus the exact constructors/operations shown there. Its opaque interfaces have leaf-private markers and no other package can implement them. `ProtectedWALKeyBindingCoreV1|ProtectedWALKeyBindingV1` are exported only so runnerprofile can pass the machine-authenticated value to the leaf；neither is aliased or returned by c12evidence, runnerprofile, a command or any other production API. `internal/c12evidence` re-exports the listed leaf data/recorder types only as assignment-identical aliases and implements the opaque high-level `CleanupOnlyOwnershipPlan`、`ProtectedWALKeyBootstrap` and `SelectedCleanupOnlyOwnership` wrappers；it does not implement a second WAL. For non-completion roles the high-level selected wrapper privately holds only runnerprofile's opaque `FixedSelectedCleanupOnlyOwnership`, inside which runnerprofile retains its post-CAS selection facts plus leaf binding. High-level Inspect passes that one token to `InspectFixedSelectedCleanupOnlyOwnership`; high-level `OpenSelectedOwnershipWAL` may pass it once to one-argument `OpenFixedSelectedOwnershipWAL`, whose same-call current guard reread dominates the raw leaf Open. Neither witness nor binding crosses into c12evidence. For completion, `internal/c12runnerprofile` stores only the bounded sealed binding encoding inside the authenticated nested capsule state and calls the same leaf Marshal/Recover/Open engine internally only after its own current-selection reread；it never imports c12evidence and never exposes the encoding or recorder. Raw leaf selected-WAL Recover/Open does not independently read the upper guard；the exact owner/control-flow gate therefore limits raw Open to `OpenFixedSelectedOwnershipWAL` plus the completion coordinator, limits raw Recover to that coordinator, requires current phase/epoch/full-digest/capsule/policy success to dominate each call, and removes the impossible claim that raw bytes alone prove current selection. Cleanup capability Open is stricter still：there is no exported c12evidence raw-bytes+policy opener. Only runnerprofile's five sealed-successor/handle methods `OpenFixedStagedDiagnosticCleanupCapability|OpenFixedWindowsCleanupCapability|OpenFixedPlatformResumeCleanupCapability|OpenFixedAuthorityCleanupCapability|OpenCompletionInnerCleanupCapability` may call leaf `BindCleanupInspectorDeleterSet` and `OpenOwnershipCapsuleForExactCleanup`; each rereads and exact-matches current selection, reconstructs declaration-bound adapters from its private runtime capabilities, and linearly consumes the adapters/set/cleanup-open claim on every return. The named c12evidence APIs otherwise map shared bootstrap/diagnostic dispatch and forward to guarded runnerprofile Open or leaf Execute/Prepare, while `FinalizeCleanedOwnership(ctx,cap,proof)` prepares the same receipt and passes it to runnerprofile `FinalizeFixedCleanedOwnership` for the role-state commit. `InspectProtectedWALKeyBootstrapRecoveryDispatch` maps the raw fixed enum exactly and the two zero-capability abort facades below are owned only by `internal/c12evidence/cleanup.go`. `InspectStagedDiagnosticCleanupRecoveryDispatch` maps exactly `cleanup_only|ordinary_finalization_pending`; its cleanup target pathlessly recovers/describes the sealed diagnostic successor and calls only guarded `OpenFixedStagedDiagnosticCleanupCapability`, while the ordinary target is the same zero-capability `RecoverCleanedOwnership`. After the first ordinary pending CAS has consumed any cleanup successor, selector-free `RecoverCleanedOwnership(ctx)` delegates only to runnerprofile `RecoverFixedCleanedOwnership`；it returns no capability/session/receipt and rechecks the current authenticated role reservation plus stored target. `OpenCompletionInnerCleanupForExactCleanup` is only a c12evidence delete-only adapter：it obtains/authenticates the bounded descriptor and invokes the guarded runnerprofile handle method, never a raw leaf opener or recorder opener. Compile tests `TestCleanupCoreAliasesAreAssignmentIdentical` and `TestCleanupCoreDependencyGraphIsAcyclic` require `WALPhase|WALRecordV1|ResourceRef` plus every other declared alias/receipt/selected-WAL type to pass directly among c12evidence、c12cleanup and c12runnerprofile without conversion, exact-cover the sole WAL canonical/HMAC domain implementation in `internal/c12cleanup/core.go`, reject any duplicate implementation or leaf import of either upper package, and preserve the dependency DAG.

The exact capsule schema is `talenro-c12-cleanup-only-ownership-capsule/v1`; its unsigned projection excludes only `SealAttestation` and is sealed over `TALENRO-C12-CLEANUP-ONLY-OWNERSHIP-CAPSULE-V1\x00 || JCS(unsigned)`. The projection contains exactly the fields of `C12CleanupOnlyOwnershipCapsuleV1` above, including nonempty `Role`, nonzero `LaunchGeneration` and nonzero `MachineSealLocatorDigest`; those three are derived from and exact-matched to the selected plan/protected-key/machine authority before Seal, enter the envelope digest and may never be supplied by a cleanup caller. Protected-key bootstrap is intent-first and digest-acyclic. `CanonicalCleanupOnlyOwnershipPlan` strict-validates schema `talenro-c12-cleanup-only-ownership-plan/v1`, every nonzero scalar/digest, a nonempty declaration set in bytewise order of each canonical declaration, and rejects duplicate `ResourceType|StableParentIdentityDigest|StableNameOrReservation` or duplicate declaration digest. It returns exactly the JCS fields shown in `CleanupOnlyOwnershipPlanV1`; `CleanupOnlyOwnershipPlanBindingDigest` is exactly `SHA-256(ASCII("TALENRO-C12-CLEANUP-ONLY-OWNERSHIP-PLAN-V1") || 0x00 || CanonicalCleanupOnlyOwnershipPlan(plan))`. These are the sole canonicalizer/digest functions used by Bind、Begin、capsule Seal and cleanup recovery, and tests independently mutate every field、reorder/duplicate declarations and splice a plan with the same cleanup-plan digest.

For every non-completion role, the unexported c12evidence role facade first generates only stable parents/names/reservations and creation-policy digests, without creating a resource, and constructs `CleanupOnlyOwnershipPlanTemplateV1`; the template deliberately has no run、role、generation、runner、provider、slot、machine or cleanup-plan field. The fixed role builders may call only high-level `BindCleanupOnlyOwnershipPlan(ctx,session,template)` in `internal/c12evidence/cleanup.go`; that function alone calls raw runnerprofile `BindFixedCleanupOnlyOwnershipPlan`, immediately calls read-only `InspectFixedCleanupOnlyOwnershipPlan`, exact-matches the returned view/template and returns only a sealed high-level wrapper. Runnerprofile fills the full plan's run/role/generation/runner/provider/recovery-slot/machine fields, maps each resource type to the session's sealed fixed provider capability, derives the exact per-declaration production adapter binding/implementation identity, computes the cleanup-plan digest and the plan binding digest, and returns only opaque `FixedCleanupOnlyOwnershipPlan`. The Inspect view exposes no installed/provider/OS handle. The high-level `CleanupOnlyOwnershipPlan` wraps that fixed plan plus the assignment-identical view, so callers cannot implement or mutate one. `BeginProtectedWALKeyBootstrap` passes the same session and hidden fixed plan only to runnerprofile `BeginFixedProtectedWALKeyBootstrap`, which rechecks their sealed session identity and every plan digest before consuming the plan. It derives canonical `ProtectedWALKeyBindingCoreV1` with one deterministic backend-object reservation and binds the plan digest into its private continuation. The `protected_key_intent` candidate contains only that core、plan binding digest、the exact predecessor full-state digest and origin-fixed abort target；it contains no `BootstrapIntentDigest`、`BindingDigest` or attestation. Its digest is `SHA-256(ASCII("TALENRO-C12-PROTECTED-WAL-KEY-INTENT-V1") || 0x00 || JCS(candidate))`. Runnerprofile writes/file+directory-fsyncs that candidate, guard-CAS-selects its digest and rereads the selected phase、epoch、full state、role/run/generation/slot/core/plan. Only afterward and within the same call does it construct `ProtectedWALKeyBindingV1` with the exact core and `BootstrapIntentDigest=selected candidate digest`; `BindingDigest = SHA-256(ASCII("TALENRO-C12-PROTECTED-WAL-KEY-BINDING-V1") || 0x00 || JCS(Core) || BootstrapIntentDigest)`, and the machine attestation covers exactly `Core|BootstrapIntentDigest|BindingDigest` while excluding only `Attestation`. Thus no field hashes itself. It calls leaf `OpenProtectedWALKeyIntent` and stores the resulting leaf intent only inside the runnerprofile-private concrete `FixedProtectedWALKeyBootstrap`; c12evidence receives that outer continuation, which has no methods and exposes neither raw binding nor leaf intent/handle. A crash before the guard CAS returns no continuation and therefore has backend create count zero；an unselected candidate has no key authority.

Every later role-specific shorthand saying that a facade “derives/builds the sealed cleanup plan” expands to this exact high-level `BindCleanupOnlyOwnershipPlan` call followed by `BeginProtectedWALKeyBootstrap`; it never authorizes `runnerhelper.go`、`platform.go` or `authority.go` to reference raw `BindFixedCleanupOnlyOwnershipPlan|InspectFixedCleanupOnlyOwnershipPlan` directly.

Plan Bind is non-consuming only with respect to the authenticated session：it returns no session copy or new authority, and the exact same held session plus sealed plan may next enter only one `BeginProtectedWALKeyBootstrap` attempt. Begin consumes the plan on every return；a second Bind/Begin、cross-plan/session splice or any intervening role operation rejects before guard or leaf access.

The same live process calls only `CreateOrRecoverProtectedWALKeyBootstrap(ctx,bootstrap)`, which forwards the hidden runnerprofile continuation to `CreateOrRecoverFixedProtectedWALKeyBootstrap`. That guarded method rereads and exact-matches the still-current intent phase、epoch、full state、role/run/generation/slot/core/plan before any leaf/backend call. Its first invocation passes the privately held intent to leaf `CreateOrRecoverProtectedWALKey`, which consumes the raw intent on every return；if delivery is ambiguous, a retry under the same outer continuation reconstructs the same binding and a fresh raw intent only after the same current-state reread, then exact-reopens the one deterministic object. The raw handle remains inside runnerprofile. The method returns only `FixedProtectedWALKeyBootstrapView` with the exact plan、backend reservation and nonsecret handle digests；a copied outer continuation cannot call a backend directly, and after any phase replacement every method rejects before raw Open/Create. The Windows DPAPI/TPM and Linux kernel-keyring/TPM adapters accept no path、raw key、backend、role or slot option. The retained plan already contains `OwnershipWALReservationCoreDigest` but not the handle-bound final reservation digest.

After key creation the caller may supply only `C12CleanupOnlyOwnershipCapsuleTemplateV1{IssuedAt,ExpiresAt}` to `SealAndSelectProtectedWALKeyBootstrap`; the high-level wrapper passes that two-field value and its hidden continuation directly to runnerprofile `SealAndSelectFixedCleanupOnlyOwnershipCapsule`. In one guarded call runnerprofile rereads the same intent state and private plan/view, derives/fills schema、run、role、launch-generation、runner、provider、recovery-slot、machine-seal locator、WAL core、cleanup-plan and declared-resource projections itself, and computes `OwnershipWALReservationDigest = SHA-256(ASCII("TALENRO-C12-OWNERSHIP-WAL-RESERVATION-V1") || 0x00 || OwnershipWALReservationCoreDigest || WALKeyProtectedHandleDigest)`. It strict-validates the complete internally constructed capsule, calls leaf `SealOwnershipCapsule` with its private handle (consumed on every return), writes/file+directory-fsyncs the exact envelope into the role's fixed no-follow recovery slot, recomputes its digest, guard-selects the exact `cleanup_only` capsule from the selected intent, rereads the new phase/epoch/full state and consumes the outer bootstrap on every return. No caller can supply a capsule identity field or `SealAttestation`. A pre-selection response-loss retry may reconstruct only an assignment-identical raw intent/handle after the current-intent reread；after selection it cannot reach raw Open/Create/Seal. The method returns only one runnerprofile opaque `FixedSelectedCleanupOnlyOwnership`, which privately contains the current selection facts、leaf binding and three-digest view. The high-level selected wrapper retains only that token；`InspectSelectedCleanupOnlyOwnership` delegates to its read-only Inspect and exposes only the capsule、handle and final reservation digests, never envelope bytes、binding、raw intent/handle、path or state selector. `OpenSelectedOwnershipWAL(ctx,selected)` is the sole owning consumer：it passes the token once to runnerprofile `OpenFixedSelectedOwnershipWAL`, which rereads and exact-matches the current guard selection, derives the current policy, consumes the token and hidden binding on every return, and only then invokes leaf Open. The leaf exact-opens or `O_EXCL`-creates only the selected plan's reserved WAL child and writes+file/directory-fsyncs the reservation bootstrap and observed no-follow one-link regular-file actual before runnerprofile returns the sealed recorder. That interface can append the four exact phases for a plan-declared `ResourceRef` but exposes no path、key、open/create primitive or arbitrary record；no other API consumes selected ownership. A candidate write/rename/fsync without the guard transition returns no fixed-selected token and has zero cleanup/workload authority；an old token becomes unusable as soon as another selected state replaces its bound full digest.

Every non-completion zero-argument parent restart runs `InspectProtectedWALKeyBootstrapRecoveryDispatch` before its role-specific phase router. After prior containment absence proof and installation+reread of the fixed replacement reservation, it exact-maps a guard-selected active generation with no selected key intent/capsule/WAL/resource to `abort_unstarted`, a selected protected-key intent to `abort_intent`, and only an authenticated role-specific selected state with neither abort condition to `role_dispatch`. Candidate files without guard selection never affect the result. The first target alone calls selector-free `AbortUnstartedProtectedWALKeyBootstrapRecovery`；`AbortFixedUnstartedProtectedWALKeyBootstrap` rereads the same active phase/epoch/full state/session/plan-candidate absence, proves key/backend/WAL/ordinary-resource create counts zero, removes only a powerless unselected plan/intent candidate if present, no-follow unlinks the exact candidate/recovery-slot payload while that abort predecessor is still guard-selected, fsyncs its parent and exact-rereads absence, and only then compare-and-advances+rereads the fixed non-affirming abort target. It closes crashes after `OpenInstalledNativeRunnerSession`, during fixed-plan Bind and before/during the first intent candidate CAS. The second target alone calls `AbortProtectedWALKeyBootstrapRecovery`；`AbortFixedProtectedWALKeyIntent` rereads the current intent phase/epoch/full state/core/plan, reconstructs and opens one raw leaf intent without returning it, first calls `DestroyProtectedWALKeyIntent` and proves the backend object absent, then no-follow unlinks the exact intent/capsule candidate and recovery-slot payload、fsyncs the parent and exact-rereads absence while the same intent predecessor remains selected；only after both backend and slot absence are durable does it compare-and-advance+reread the intent's immutable abort target. In both routes the predecessor cell itself retains every digest/core/plan/target fact needed to reverify absence, so the target-CAS precondition never depends on bytes already unlinked. The target CAS is the sole final linearization：a crash before it still dispatches the same idempotent abort route；response loss after it only rereads that target and performs no filesystem operation. Leaf Open/Create/Destroy/Seal are owned only by the guarded runnerprofile bootstrap/abort methods plus the separately guarded completion coordinator；c12evidence has zero raw production reference. Live-process setup failure uses the matching unstarted/intent target. The role-dispatch target and every role-specific router recheck absence of both abort conditions；each abort target rechecks its exact phase/reservation. An inspect/abort error is terminal, never permission to try role recovery, create a new key or fall through. Raw leaf intents are consumed on every Create/Destroy return and never cross runnerprofile；a copied outer continuation has no backend method, and an old internal binding can be reopened only after a current-intent reread. Capsule、abort、resume、finalization and terminal selection therefore reject old-continuation and destroy→recreate replay before leaf/backend access. Completion nested intent recovery is owned separately by the coordinator below and cannot call this shared outer abort route. No global key-store or slot enumeration is permitted.

Completion inner cleanup is the sole session-free bootstrap exception because `BeginCompletionPublishPending` has already consumed the native launch session. The coordinator has retained only the sealed role/install/run/recovery-slot projections from that consumed session；they cannot be caller-supplied or reread from argv/environment. `AuthenticatedCompletionRoleStateCoordinator.BeginInnerCleanup` first rechecks the five-way dispatch and is callable only for `begin_required`, then internally owns the same intent-first sequence against its exact closed temporary-resource plan：inside the guarded `publish_pending|retained_complete_active` owner it derives the binding from those retained projections, exact-recovers or removes/rebuilds only an assignment-identical unselected candidate left before the first inner CAS, selects a nested `protected_key_intent`, calls the leaf Open/CreateOrRecover/Seal operations, stores the bounded sealed selected-WAL binding encoding inside the authenticated capsule substate, and selects+rereads that nested cleanup capsule. It then privately calls leaf `RecoverSelectedOwnershipWALBinding` and `OpenSelectedOwnershipWAL`, exact-creates or reopens the one reserved WAL, and verifies its bootstrap plus observed-identity actual are file+parent-directory durable. Only after that barrier does it return `CompletionInnerCleanupHandle`, whose concrete unexported value retains the sole leaf recorder；the public interface exposes no append method、binding bytes、path、key or general create authority. The none-state proof requires backend-key and WAL create counts zero, so candidate convergence cannot duplicate either object. No coordinator caller receives a binding、intent、key handle、capsule selector or recorder.

`InspectInnerRecoveryDispatch` is the sole fresh-process inner router and never reports a branch through success/error shape. It combines the authenticated inner tag with owner-specific outer progress and the same adoption set's authenticated pending transition：inner `none` before create's first set-bound `manifest_intent` with no set-first pending transition, or under initial `retained_complete_active`, is only `begin_required`; create inner `none` with reciprocal `manifest_intent` or later recoverable publication progress is only `complete`. Selected pre-handle/capsule states with no manifest intent are `session_validation_recovery`; create `cleanup_capsule` with a set-first pending manifest intent or outer `manifest_intent|manifest_actual` is instead `session_cleanup_recovery`. Both session routes share a private bootstrap subdispatch closed to `restart_pending|intent_selected|capsule_selected`. Under `intent_selected`, `RecoverInnerCleanup` exact-removes the candidate、destroys/verifies the provisional key and guard-selects `inner_bootstrap_restart_pending` while preserving all owner business；under `restart_pending` it begins a fresh deterministic key intent for the same assignment；under `capsule_selected` it authenticates the stored sealed WAL binding, exact-recovers/opens that one WAL and rechecks durable bootstrap+observed identity before rehydrating the handle. Only validation recovery may traverse pre-capsule states. Cleanup recovery requires an existing capsule；when only set-first manifest intent is durable it first completes that reciprocal state transition from stored bytes without filesystem work, then rereads the route and returns the same capsule only after the WAL barrier is also revalidated. The loop returns `(nonzero handle,nil)` exclusively after capsule selection **and** WAL bootstrap/actual durability, with `validation_required|cleanup_only` sealed into the high-level session. The validation handle's unexported concrete owns the only recorder；cleanup-only state disables every new intent/actual append and permits only delete-side cleaned/not-found records through the paired leaf capability. `finalization_pending` maps only to `finalization_recovery`; `RecoverInnerCleanup` rejects every other branch. `RecoverInnerFinalization` reads stored receipt/tombstone/fixed target and idempotently closes leaf Commit、key destruction/absence and terminal retirement/target selection without constructing a cleanup capability or returning a handle. A crash/error leaves exactly one authenticated route；no action returns `(nil,nil)`, infers a branch from error, Opens another launch session or changes owner/generation/assignment. Before first intent CAS only rechecked `begin_required` may Begin with both key/WAL create counts zero；after the create fixed-target CAS only `complete` remains and Begin plus both session recoveries reject. A retained target removes the active owner. Thus create and revalidate use the same leaf key/WAL protocol without the session-bound high-level bootstrap, post-key-destroy recovery needs no key reopen, and post-manifest recovery cannot refetch or regenerate bytes.

Only after the capsule payload is durable and the intent→`cleanup_only` guard transition rereads successfully may the leaf selected-WAL engine create/exact-reopen the reserved WAL. Non-completion reaches it solely through high-level `OpenSelectedOwnershipWAL(ctx,selected)` after that method rereads the opaque selected owner；completion reaches the same engine solely inside coordinator Begin/Recover after its current-state reread and before a handle/session is returned. Both require bootstrap+observed-identity actual file+parent-directory fsync before any run-root、ordinary resource intent/actual、producer、workload、provider or private-child create. The selected WAL binds reservation and observed identity into later WAL/manifest/state digests；replacement/file-ID drift rejects. A reboot with guard-selected `cleanup_only` may open the key for deletion only through the matching `OpenFixedStagedDiagnosticCleanupCapability|OpenFixedWindowsCleanupCapability|OpenFixedPlatformResumeCleanupCapability|OpenFixedAuthorityCleanupCapability|OpenCompletionInnerCleanupCapability` sealed-successor/handle method. Each method revalidates the live selected guard state before privately invoking leaf Bind/Open and returns only exact inspect/recover/delete for the declared type/name/parent/provider set, with no raw key、recorder、create/resume/workload capability or enumeration. A pre-selection candidate、a descriptor copy or an old capsule after resume/terminal selection cannot enter any of those methods and is rejected before leaf Open. `ExpiresAt` ends create/resume/workload/signing authority but never disables delete-only recovery, including trusted-time/provider outage. Missing/partial capsule with no selected intent proves key/WAL/ordinary resource create counts zero；a selected intent permits only its deterministic provisional key and abort route；a selected capsule mismatch quarantines. `CleanupOnlyPolicy` is derived only from authenticated guard/slot/capsule projections and accepts no caller time、expiry bypass、resource selector or outcome. `ExecuteCleanup` uses the leaf capability to append only cleaned/not-found, runs declared inspectors/deleters and returns a matching proof only after terminal WAL and full absence set are durable. `PrepareFinalization` durably stages one exact receipt；Inspect/Recover exact-open only that receipt and never enumerate；Commit is idempotent only for the same receipt. Runnerprofile stores the complete receipt view plus fixed target before Commit, so response loss reconstructs from selected pending state. No commit produces create/read authority, aliases remain assignment-identical, and a crash before terminal state replays only the same cleanup/receipt/target route.

`WALPhase` is closed to `intent`, `actual`, `cleaned`, and `not_found`. Apart from the capsule-authorized WAL bootstrap projection above, append intent and `fsync` file+parent before create；then append actual with the observed identity. Every record has a nonzero declaration digest that uniquely selects one capsule declaration and redundantly repeats its exact type/name. HMAC covers the domain `TALENRO-C12-OWNERSHIP-WAL-V1\x00`, canonical record—including `ResourceDeclarationDigest`—without `RecordHMAC`, previous digest and record number. Reject gaps, rewrites, duplicate actual, actual time before its capsule/WAL intent, missing/unknown/cross-parent declaration digest, declaration-digest/type/name/identity changes, invalid HMAC and trailing partial records. The WAL key is a per-run CSPRNG value inherited through DPAPI/OS-keystore protected handle. For Linux cold reboot, the helper reopens it only after authenticating the role guard/current state and the fixed recovery-authority payload as either the cleanup-only capsule above or its bound `CrossRebootResumeManifestV1`. Manifest adoption is one guarded logical slot replacement：write/cap/fsync a manifest candidate that embeds the capsule envelope digest、`OwnershipWALReservationDigest`、observed WAL file-identity/terminal-prefix digest and identical run/runner/provider/WAL-key/declared-resource/cleanup-plan projections, bind its exact identity+digest in the inactive role-state cell, then compare-and-advance/reread the role guard. Thus every selected durable state has exactly one authority；before guard advance the candidate has none and the old cleanup authority remains selected, while after advance only resume authority is selected. A crash at either side is exact-forwarded by the adapter and cannot choose a different payload. Exact-clean destroys/unseals the key only through the guarded terminal tombstone transition below.

Every cleanup-only, resume, staged producer and final producer uses the same leaf absence/WAL/key finalization protocol；none may unlink its authority slot or WAL key directly. Ordinary and abort-origin cleanup target terminal-inactive/next；Windows/platform/authority success-origin target only their same-generation post-clean publication successor and advance next only after nested/outer/channel finalization. After the declared set is absent and terminal `cleaned|not_found` WAL is file+directory durable, leaf Prepare creates the assignment-identical receipt and runnerprofile first selects a pending state carrying that full receipt plus one origin-fixed target. The authenticated `talenro-c12-cleaned-ownership-tombstone/v1` grants only idempotent destroy/verify-destroyed for its exact key handle and no WAL open、resource inspect/delete、resume、provider or signing authority. While the pending state is still guard-selected, recovery must complete leaf Commit、prove the key absent、no-follow unlink the tombstone、fsync its parent and verify the exact payload absent；only then may the second guarded CAS select the fixed terminal/post-clean target. Thus target selection is the final linearization and has no later filesystem work；`complete`/post-clean dispatch may be actionless. A crash before the first CAS retains predecessor authority and a powerless candidate；after it retains the stored receipt/target；after key destruction or either unlink/fsync response loss it stays pending and only re-verifies the same facts；after the second CAS the tombstone is already absent. Candidate write/fsync、both guard advances/rereads、key destroy/absence、unlink/parent-fsync/absence and target-CAS seams are resumable and can never recreate a key/resource, change target or substitute a receipt.

- [ ] **Step 7: Run WAL tests and verify GREEN**

Run: `go test ./internal/c12cleanup ./internal/c12evidence -run '^TestWAL|^TestCleanupCore|^TestTreeFactsCapsuleStrictValidation$' -count=1`

Expected: PASS, including crash after intent, crash after create, partial write, wrong key, record reorder and cross-run replay cases.

- [ ] **Step 8: RED — add exact recovery and cleanup precedence tests**

```go
func TestExitBitsPreservePrimaryAndCleanupFailures(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		primary bool
		cleanup bool
		want    int
	}{{false, false, 0}, {true, false, 1}, {false, true, 2}, {true, true, 3}} {
		if got := ExitBits(tc.primary, tc.cleanup); got != tc.want {
			t.Fatalf("ExitBits(%t,%t)=%d want %d", tc.primary, tc.cleanup, got, tc.want)
		}
	}
}
```

Add `TestCleanupInspectorDeleterSetExactCoverAndNoRegistry` and `TestRecoverExactUsesBoundAdapterAndWritesActualBeforeDelete`. The first constructs two declarations with the same resource type、provider、creation policy and implementation but the identical stable name under different stable parents, requires their declaration digests and adapter bindings to remain distinct, and proves both bind/clean；missing/extra/duplicate/reused adapter、declaration swap、parent/declaration-digest/type/name splice、global registry/init map or a c12evidence/cmd adapter implementation rejects before inspection. The second exercises absent-actual exact recovery through the capability-held adapter, requires the observed actual plus file/parent durability before return or delete, and proves a mismatch、pre-intent creation time、wrong adapter/declaration or attempted list fallback yields no delete. Both roots are literal members of the Task 2 and final B10 exact manifests.

Add `TestProtectedWALKeyIntentClosesEveryPreCapsuleCrashSeam`: crash immediately after session Open, before/after fixed-plan Bind/Inspect, before/after sealed-plan/session cross-match、core construction、bootstrap-intent candidate canonicalization/write/file+directory fsync、guard compare-and-advance/reread、post-selection binding digest/attestation construction、backend key create、provider response、same-binding CreateOrRecover response loss、capsule candidate create/write/file+directory fsync and intent→`cleanup_only` guard compare-and-advance/reread. Independently mutate/cross-splice every plan field/order、session、core、selected intent digest、binding digest and attestation and prove the exact construction order has no digest fixed point or caller override. Compile/AST assertions require the public capsule template to contain exactly `IssuedAt|ExpiresAt`, the plan template to contain only reservation core plus resource reservations, prove every machine/run/slot/provider/parent/resource/WAL projection is filled inside runnerprofile from the sealed plan/session, and reject any alternate plan/bootstrap implementer or caller field. Before the first intent CAS, restart returns only `abort_unstarted`, the zero-capability target proves key/backend/WAL/ordinary resource counts zero and terminalizes the claimed generation without role dispatch. Its crash matrix covers before/after powerless candidate/recovery-slot unlink、parent-directory fsync、exact absence reread、final target CAS+reread and target-CAS response loss, and requires the predecessor to remain selected until slot absence is durable. While intent is selected, at most its deterministic one key object may exist；a restart returns only `abort_intent`, first exact-destroys or verifies absent that key, then exact-removes the reservation-bound intent/capsule candidate and recovery-slot payload, and while the intent predecessor remains selected covers before/after key destroy/absence、intent/capsule-candidate and recovery-slot unlink、parent-directory fsync、exact absence reread、final target CAS+reread and target-CAS response loss. Both routes require zero filesystem work after target selection and prove their final CAS uses only predecessor-cell facts, not unlinked bytes. After capsule selection, only `role_dispatch` is legal and copied outer-bootstrap/internal-binding/raw-intent test values cannot destroy or recreate the key. Repeat those replay attempts after abort、resume adoption、finalization pending and terminal selection, including destroy→recreate, and require rejection before leaf/backend call. It then injects every `OpenSelectedOwnershipWAL` O_EXCL/exact-reopen/bootstrap/actual/file+directory-fsync seam and proves the selected token is consumed on every return, exactly one sealed WAL is reachable, its first durable actual matches the planned reservation and no alternate API can use selected ownership. Mutated/cross-run binding、backend reservation、intent/source/abort-target digest、live-or-ambiguous old containment、second key or WAL object、enumeration、selected-token replay、abort→role error fallback and role→abort fallback all reject. Same-binding provider response loss remains exact-idempotent and every normal seam ends with zero provisional key/candidate residue or exactly one selected capsule plus its one authenticated WAL authority.

Add `TestCompletionInnerCleanupProtectedWALKeyIntentCrashSeams`: run the plan/core/candidate/intent/backend/capsule/selected-WAL matrix once under create `publish_pending` and once under revalidate `retained_complete_active` after launch-session consumption. Only coordinator `BeginInnerCleanup|RecoverInnerCleanup` may call the leaf protocol. Before first inner CAS, backend-key/WAL/resource/provider/child counts stay zero and only rechecked Begin may converge the no-authority candidate. After intent selection and before any manifest transition, inject abort/key-destroy/restart/replacement-intent/capsule seams plus selected-WAL binding Seal/Marshal/store-in-authenticated-cell、selection CAS/reread、binding Recover、WAL O_EXCL/exact-reopen、bootstrap、observed-identity actual and every file/parent-directory fsync/response-loss seam. Recovery returns exactly one nonzero validation-required handle only after WAL readiness；mutated/truncated/oversize/noncanonical/cross-owner binding、policy/reservation/file-ID splice、live-binding replay or a second WAL/recorder rejects. This Task 2 version exercises only pre-channel `begin_required|session_validation_recovery|finalization_recovery` states and keeps the two future enum values unreachable, so it compiles without `channel.go`. Task 7, when it creates the adoption-set/channel engine, explicitly edits this same unique test to add each set-first reciprocal manifest-intent seam、`session_cleanup_recovery` state-only convergence and post-target `complete` assertions；only then may production `InspectInnerRecoveryDispatch/RecoverInnerCleanup` reference the private manifest barrier. The final B10 root exact-covers all five values/classes and rejects nil-handle success、cleanup-only validation/Stage、error fallback、cross-owner/generation or any raw binding/encoding/recorder escape.

Add `TestProtectedWALKeyBootstrapRawRunnerprofileCallsOwnedOnlyByCleanupFacade`. Across all production files/build tags it freezes the exact high-level c12evidence facade set `BindCleanupOnlyOwnershipPlan|BeginProtectedWALKeyBootstrap|CreateOrRecoverProtectedWALKeyBootstrap|SealAndSelectProtectedWALKeyBootstrap|InspectSelectedCleanupOnlyOwnership|OpenSelectedOwnershipWAL|InspectProtectedWALKeyBootstrapRecoveryDispatch|AbortUnstartedProtectedWALKeyBootstrapRecovery|AbortProtectedWALKeyBootstrapRecovery|InspectStagedDiagnosticCleanupRecoveryDispatch|RecoverStagedDiagnosticCleanup` to `internal/c12evidence/cleanup.go`. Only the literal fixed role-builder call sites in `internal/c12evidence/runnerhelper.go|platform.go|authority.go` may call high-level Bind；commands and every other production file have zero reference. `cleanup.go` is the sole c12evidence caller of runnerprofile `BindFixedCleanupOnlyOwnershipPlan|InspectFixedCleanupOnlyOwnershipPlan|BeginFixedProtectedWALKeyBootstrap|CreateOrRecoverFixedProtectedWALKeyBootstrap|SealAndSelectFixedCleanupOnlyOwnershipCapsule|InspectFixedSelectedCleanupOnlyOwnership|OpenFixedSelectedOwnershipWAL|InspectFixedProtectedWALKeyBootstrapDispatch|AbortFixedUnstartedProtectedWALKeyBootstrap|AbortFixedProtectedWALKeyIntent`; command packages have zero references. It exact-checks the reservation-only plan template、two-field capsule template、opaque outer bootstrap/plan/fixed-selected owners, the assignment-identical plan/selection views and the selected wrapper's one consumer；Inspect methods are read-only and non-authorizing.

The complete exported leaf function set is literal-exact：`CanonicalCleanupOnlyOwnershipPlan|CleanupOnlyOwnershipPlanBindingDigest|CanonicalCleanupOnlyPolicyBinding|CleanupOnlyPolicyBindingDigest|BindCleanupInspectorDeleterSet|OpenProtectedWALKeyIntent|CreateOrRecoverProtectedWALKey|DestroyProtectedWALKeyIntent|SealOwnershipCapsule|MarshalSelectedOwnershipWALBinding|RecoverSelectedOwnershipWALBinding|OpenSelectedOwnershipWAL|OpenOwnershipCapsuleForExactCleanup|RecoverExact|Execute|PrepareFinalization|InspectFinalizationReceipt|RecoverFinalization|CommitFinalization`. C12evidence has zero raw leaf key Open/Create/Destroy/Seal or selected-WAL Open/Recover reference. Raw key Open/Create/Destroy/Seal call sites are literal-exact inside runnerprofile's guarded fixed bootstrap/abort methods plus the completion coordinator；each raw call is dominated by current intent phase、epoch、full-state、slot、core and plan success, and old outer/bootstrap/binding/intent copies after capsule、abort、resume、finalization or terminal selection fail before leaf/backend access. Raw selected-WAL Open is owned only by `OpenFixedSelectedOwnershipWAL` plus the completion coordinator；Marshal/Recover are completion-only, and every call is dominated by current capsule/policy selection. At the Task 2 GREEN point raw cleanup Bind/Open has exactly four paired guarded owners：`OpenFixedStagedDiagnosticCleanupCapability|OpenFixedPlatformResumeCleanupCapability|OpenFixedAuthorityCleanupCapability|OpenCompletionInnerCleanupCapability`. Task 6 edits this same unique table when `windows_cleanup.go` is created, adding only `OpenFixedWindowsCleanupCapability`; the final B10 form requires exactly those five and reruns the root. Each owner constructs the declaration-bound concrete adapters and calls both leaf Bind and Open under one current-state branch；c12evidence/cmd have zero Bind/adapter/set reference. Go/types exact-enumerates every allowed concrete `ExactResourceInspectorDeleter` implementation plus declaration/resource type；a caller-defined implementation、registration、unlisted constructor or unresolved dynamic target fails. `RecoverExact` has zero upper-package production reference and is invoked only by leaf Execute/recovery control flow against the capability-held set.

The same root freezes finalization raw ownership. `PrepareFinalization` is called only by `internal/c12evidence/cleanup.go`; `InspectFinalizationReceipt` is called only by the fixed selection helpers in `internal/c12runnerprofile/cleanup_finalization.go` and `completion_state.go`. Raw `RecoverFinalization|CommitFinalization` have exactly two production call owners：private `commitGuardSelectedCleanupFinalization` in `cleanup_finalization.go` and private `commitGuardSelectedCompletionInnerFinalization` in `completion_state.go`. Each helper is reachable only from the named fixed finalizer/recovery methods, and a successful reread of the exact current `*_finalization_pending` phase、epoch、full digest、stored receipt view and immutable target dominates both raw calls. Pre-pending、cross-role/run/generation、copied/mutated view、post-target and second-commit paths reject before leaf access；no c12evidence/cmd function may Recover or Commit directly. The plan/policy canonicalizers are the sole shared helpers used by Bind、Begin、Seal、descriptors and guarded recovery. AST plus go/types control-flow assertions prohibit error fallback and every other name/call/alias/forward/function value. The test strict-validates the complete plan and all twelve `CleanupOnlyPolicy` fields、all canonical domains、adapter/cleanup-plan exact cover、the 16-KiB selected-WAL envelope and fixed-selected-token/hidden-binding/bootstrap/set/finalization linear consumption；encoding stays only in the authenticated completion inner cell and recorder stays only in c12evidence's live return or concrete completion handle/runtime. It exercises a byte-valid pre-CAS candidate、selection→Open state replacement、post-resume old capsule and post-terminal old capsule and requires rejection before any leaf Open/create. The sole exported runnerprofile recorder-returning exception is one-argument `OpenFixedSelectedOwnershipWAL`, callable only from `internal/c12evidence/cleanup.go` and returning the sealed recorder after its same-call current-state reread；all other exported runnerprofile signatures and every composite/command signature return no selected-token internals、binding、encoding、raw intent、raw handle、recorder、set、path or key and accept no full capsule/policy/adapter authority.

The two build-tagged protected-key files each own one and only one explicit aggregate root：Windows `TestProtectedWALKeyWindowsProductionAdapter` with literal subcases `machine_dpapi_scope_and_acl|exact_selected_binding_create_recover|role_generation_machine_splice_rejected|destroy_exact_and_verify_absence|no_caller_path_or_secret_export`, and Linux `TestProtectedWALKeyLinuxProductionAdapter` with `owner_only_os_keystore_scope_and_mode|exact_selected_binding_create_recover|role_generation_machine_splice_rejected|destroy_exact_and_verify_absence|no_caller_path_or_secret_export`. Each aggregate exercises the production adapter rather than a fake and rejects an environment、argv、caller path、unprotected alternate backend or exported raw key. The B10 meta-test parses both tagged files independent of the planning OS, requires those exact literal subcase tables, a unique top-level owner and a statically reachable helper graph with no `Skip|Skipf|SkipNow`；the matching host gate must additionally list and run/pass the root and every descendant with zero skip.

Add `TestCleanupOnlyCapsuleClosesEveryPreManifestPowerLossSeam`: begin with the complete key-intent test above, then inject cold reboot immediately after the selected capsule reread, WAL `O_EXCL` create/bootstrap/actual fsync, producer return, every ordinary resource intent, every actual, resume-manifest candidate create/write/fsync, guard compare-and-advance and selected-state reread, plus restart after capsule/manifest `ExpiresAt` and trusted-time-provider outage. Before bootstrap-intent selection all create counts are zero；between intent selection and capsule selection only the one exact provisional key is possible and must converge through the abort route；after capsule selection only the guard-selected authenticated cleanup-only capability may exact-inspect/recover/delete the predeclared set；after manifest adoption only the guard-selected authenticated resume authority may resume or exact-clean. A candidate rename/write without guard selection has zero authority. A missing/partial capsule with no intent proves zero create calls, while a selected intent、WAL reservation/actual file-identity/replacement splice、capsule/resource/parent/provider/WAL-key/cleanup-plan splice、enumeration、workload/sign attempt or reuse after adoption rejects or exact-aborts as frozen above. Expiry/outage must reject resume/workload/signing yet still exact-clean through the authenticated declared delete-only authority. Every normal seam recovers exactly the guard-selected intent、capsule or manifest state and ends with exact not-found/cleaned records and no provisional key/candidate/WAL/snapshot/build/run/VM/NV/PITR/credential residue.

Add `TestCleanedTombstoneClosesKeyAndSlotCrashSeams`: start once from ordinary cleanup-only and once from an abort/downgraded resume authority, prove all declared resources absent plus a terminal fsynced WAL, then crash before/after tombstone candidate write/fsync、first guard CAS/reread、key destroy/absence verification、tombstone unlink、parent-directory fsync、absence verification and the **final** guarded target CAS/reread. Recovery sees exactly predecessor、selected pending tombstone or terminal-inactive；until unlink durability is proven it remains pending, and after target selection the payload is already absent with no work left. An unselected candidate has zero authority；a selected tombstone may only commit the receipt、destroy/verify the key and retire itself, never reopen WAL/delete another resource/create/resume/sign. Wrong predecessor/WAL/absence/key/state digest、partial/foreign tombstone、key recreation、slot enumeration or target CAS before tombstone absence rejects. Every seam selects old or new without quarantine；success leaves key and obsolete payload absent. Windows/platform/authority success-origin targets are covered by their post-clean tests.

- [ ] **Step 9: Run cleanup tests and verify RED**

Run: `go test ./internal/c12cleanup ./internal/c12evidence ./internal/c12runnerprofile -run '^TestCleanupCore|^TestCleanupInspectorDeleterSet|^TestCleanupInspectorDeleterSetExactCoverAndNoRegistry$|^TestProtectedWALKey|^TestCompletionInnerCleanupProtectedWALKey|^TestNativeRunnerStagedDiagnosticCleanupRecoveryIsClosedAndGuarded$|^TestNativeRunnerStagedDiagnosticCleanupCapabilityUsesExactAdapterSet$|^TestExitBits|^TestRecoverExact|^TestRecoverExactUsesBoundAdapterAndWritesActualBeforeDelete$|^TestExecuteCleanup|^TestCleanupOnlyCapsule|^TestCleanedTombstone' -count=1`

Expected: FAIL because recovery and cleanup functions are undefined.

- [ ] **Step 10: GREEN — implement exact-only recovery and bounded cleanup**

`RecoverExact` may inspect only a stable name already present in an authenticated intent. It accepts a candidate only when exact name/type/run label/expected image or executable identity match and actual creation time is not before intent; it writes recovered `actual` before deletion. `not found` writes a terminal record. A mismatch returns `ErrIsolationRequired` without deletion or broader lookup.

`ExecuteCleanup` stops exact PID+start-token/Job records first, then exact container IDs, network ID, volumes, external VM IDs, NV handle+provider-identity pairs, PITR database/timeline IDs, ephemeral credential/certificate IDs, scheduled task, and finally repository-external snapshot/build/run paths that re-resolve to the WAL-recorded parent+file identity. Every deleter first invokes its typed exact inspector and accepts only the recorded run label、creation/start token、provider identity and expected image/toolchain identity；append-only provider/journal/archive evidence is a retained record, never a deletable resource type. It applies a single 120-second context and never calls global process/Docker/TPM/provider/database/list APIs.

Implement the Task 2 helper as a bounded native CLI over only the tracked-tree/WAL/cleanup Go APIs above. `capture-tree` is the sole Git boundary：it validates the absolute locked Git executable, captures the typed commit/tree locator separately from the canonical tree digest, proves HEAD/index/closed-source/untracked-zero invariants, materializes the exact object tree into one WAL-owned repository-external directory and makes it read-only before returning an inherited protected descriptor/facts document. For committed captures it also walks the exact commit→tree→subtree→blob object graph, verifies every object with the locked Git implementation, and writes an independently WAL-owned **minimal** read-only Git object database containing exactly that closed graph (plus only the required object-format metadata), never refs, reflogs, index, alternates, worktree config or unrelated reachable objects. It computes a no-follow directory/file-identity digest and `SHA-256(ASCII("TALENRO-C12-MINIMAL-GIT-OBJECT-DATABASE-CLOSURE-V1") || 0x00 || JCS(sorted({object_oid,type,size,content_sha256})))`; both are bound into `C12TreeFactsCapsuleV1` and its policy. For staged diagnostic captures, the analogous closure contains the synthetic index tree and every referenced object but cannot become authoritative because `CommitOID` remains empty.

A later public parent mode calls this API and retains both snapshot and object-database handles in-process; it creates/fsyncs the two opaque-root intents/actuals itself, then launches only the exact hash-verified same helper binary's private producer child with Windows `PROC_THREAD_ATTRIBUTE_HANDLE_LIST` or an exact Unix duplicated FD and private authenticated parent-session handle. For the Windows→nested-Linux boundary it additionally writes the bounded public signed `C12TreeFactsCapsuleV1` once to a repository-external WAL-owned direct-child file using no-follow `O_CREAT|O_EXCL`、one-link regular-file/owner-only ACL、32 KiB cap and file+parent fsync；the file is never beneath or copied into the tracked snapshot. The helper records capsule and minimal-object-database path/file identities/digests as actual before Compose, supplies those paths only in the native Docker/Compose child environment as `C12_TREE_FACTS_CAPSULE` and `C12_GIT_OBJECT_DATABASE`, and fixed read-only binds expose them at `/run/c12-input/tree-facts.capsule` and `/run/c12-input/object-db`. The fixed verifier parent validates capsule file identity/envelope/policy, snapshot closure and minimal object-database identity/closure, re-resolves `GitTreeLocatorV1` from that database, creates the protected handle and directly launches the private producer child through that inheritance table. Neither ordinary shell nor any shell environment receives either host/container object-database path, locator, closure facts or handle. Capsule/object-database path/signature/run/snapshot/locator/closure/WAL splice, extra/missing object, alternates/ref escape and post-bind identity drift fail before the producer. A decimal handle passed through a public CLI flag、`go run`/`go tool`/temporary executable、PATH lookup or an unverified intermediate process is always invalid. Every command accepts typed facts through one strict, regular, non-symlink input document or inherited protected handle and emits only one finite result document/category. The HMAC key is inherited through an OS-protected handle/file descriptor or DPAPI-backed helper store and is never accepted in argv, environment, stdin or stdout. Tests prove the Task 2 executable rejects the not-yet-owned public parent modes、private child、`scope-from-verifier-result` and `run-inner-integration` commands rather than stubbing a second schema/domain.

Freeze the helper and launcher sources into four locked build outputs. The fixed recipe uses `CGO_ENABLED=0`, `-trimpath -buildvcs=false -buildmode=exe` and the locked empty/nonvarying Go build-ID policy so a later commit outside a binary's import closure cannot stamp VCS revision、absolute root or ambient time into it. `GOOS=windows GOARCH=amd64` produces the hash-locked PE payload roles `c12_runner_helper_windows_amd64` and `c12_runner_launcher_windows_amd64`; their application-control/signature envelope is separately deterministic/frozen by the locked Windows signing policy and cannot add a timestamped rebuild drift. `GOOS=linux GOARCH=amd64` produces static ELF roles `c12_runner_helper_linux_amd64` and `c12_runner_launcher_linux_amd64`, but only the helper is embedded in the verifier image. Reproducibility tests build every role from two absolute roots, require byte-identical per-role payload/final-file digests and matching SBOM/provenance projections, and prove helper/launcher plus PE/ELF role/OS/arch cross-use rejects. They also rebuild the launcher from a later synthetic commit that changes only packages outside its import closure and require byte identity, closing the B10→B11 case rather than only same-commit two-root reproducibility. The helper shared command schema/canonical vectors remain identical across helper builds；launcher help/dispatch vectors instead exact-cover the OS-specific subsets above. All four identities become toolchain native-role inputs at B10, while the two launcher identities must remain byte-identical inputs at B11.

- [ ] **Step 11: Run WAL/cleanup tests and verify GREEN**

Run: `go test ./internal/c12cleanup ./internal/c12evidence ./internal/c12runnerprofile ./cmd/talenro-c12-runner-helper -run '^TestCleanupCore|^TestCleanupInspectorDeleterSet|^TestCleanupInspectorDeleterSetExactCoverAndNoRegistry$|^TestProtectedWALKey|^TestCompletionInnerCleanupProtectedWALKey|^TestNativeRunnerStagedDiagnosticCleanupRecoveryIsClosedAndGuarded$|^TestNativeRunnerStagedDiagnosticCleanupCapabilityUsesExactAdapterSet$|^TestWAL|^TestExitBits|^TestRecoverExact|^TestRecoverExactUsesBoundAdapterAndWritesActualBeforeDelete$|^TestExecuteCleanup|^TestCleanupOnlyCapsule|^TestCleanedTombstone|^TestRunnerHelperWAL' -count=1`

Expected: PASS; injected cleanup failure produces bit 2, primary+cleanup produces 3, no test deletes the pre-existing collision fixture, unavailable future subcommands fail closed, and no secret/HMAC material appears in argv/stdout.

- [ ] **Step 12: REFACTOR — run race and Windows path validation tests**

Run: `go test -race ./internal/c12cleanup ./internal/c12evidence -count=1`

Expected: PASS; WAL concurrent append attempts serialize by record number and path checks reject junction/reparse escape fixtures.

Run on the attested Windows/amd64 host with `CGO_ENABLED=0` and fail if the tagged root is absent、renamed or skipped:

```powershell
$env:CGO_ENABLED = '0'
$adapterPattern = '^TestProtectedWALKeyWindowsProductionAdapter$'
$listed = @(& go test ./internal/c12cleanup -list $adapterPattern)
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$actual = @($listed | Where-Object { $_ -cmatch '^Test[A-Za-z0-9_]+$' })
if ($actual.Count -ne 1 -or $actual[0] -cne 'TestProtectedWALKeyWindowsProductionAdapter') { throw 'missing or extra Windows protected-key adapter root' }
$jsonLines = @(& go test ./internal/c12cleanup -json -run $adapterPattern -count=1)
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$events = @($jsonLines | ForEach-Object { $_ | ConvertFrom-Json })
$run = @($events | Where-Object { $_.Test -ceq 'TestProtectedWALKeyWindowsProductionAdapter' -and $_.Action -ceq 'run' })
$pass = @($events | Where-Object { $_.Test -ceq 'TestProtectedWALKeyWindowsProductionAdapter' -and $_.Action -ceq 'pass' })
$skip = @($events | Where-Object { $_.Action -ceq 'skip' -and $_.Test -and ($_.Test -ceq 'TestProtectedWALKeyWindowsProductionAdapter' -or $_.Test.StartsWith('TestProtectedWALKeyWindowsProductionAdapter/', [System.StringComparison]::Ordinal)) })
if ($run.Count -ne 1 -or $pass.Count -ne 1 -or $skip.Count -ne 0) { throw 'Windows protected-key adapter did not run and pass exactly once without skip' }
```

Run independently on the attested Linux/amd64 host with `CGO_ENABLED=0`:

```bash
set -euo pipefail
export CGO_ENABLED=0
adapter_pattern='^TestProtectedWALKeyLinuxProductionAdapter$'
adapter_expected='TestProtectedWALKeyLinuxProductionAdapter'
adapter_actual="$(go test ./internal/c12cleanup -list "$adapter_pattern" | grep -E '^Test[A-Za-z0-9_]+$')"
[[ "$adapter_actual" == "$adapter_expected" ]]
adapter_json="$(mktemp)"
trap 'rm -f -- "$adapter_json"' EXIT
go test ./internal/c12cleanup -json -run "$adapter_pattern" -count=1 >"$adapter_json"
[[ "$(awk -v n="$adapter_expected" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"run\""){c++} END{print c+0}' "$adapter_json")" == 1 ]]
[[ "$(awk -v n="$adapter_expected" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"pass\""){c++} END{print c+0}' "$adapter_json")" == 1 ]]
[[ "$(awk -v n="$adapter_expected" 'index($0,"\"Test\":\"" n) && index($0,"\"Action\":\"skip\""){c++} END{print c+0}' "$adapter_json")" == 0 ]]
rm -f -- "$adapter_json"
trap - EXIT
```

Expected: both target adapters compile and PASS under the same fixed no-CGO build environment；the exact root and its frozen literal subcases execute without skip, neither target is inferred from the planning host and neither OS-specific test can be silently deleted or counted as coverage for the other.

- [ ] **Step 13: Commit ownership semantics**

```bash
git add internal/c12runnerprofile/authority_cleanup.go internal/c12runnerprofile/authority_cleanup_test.go internal/c12runnerprofile/staged_diagnostic_cleanup.go internal/c12runnerprofile/staged_diagnostic_cleanup_test.go internal/c12runnerprofile/cleanup_finalization.go
git add internal/c12cleanup/core.go internal/c12cleanup/core_test.go internal/c12cleanup/protected_key_windows.go internal/c12cleanup/protected_key_windows_test.go internal/c12cleanup/protected_key_linux.go internal/c12cleanup/protected_key_linux_test.go internal/c12runnerprofile/profile.go internal/c12runnerprofile/profile_test.go internal/c12runnerprofile/bootstrap.go internal/c12runnerprofile/bootstrap_test.go internal/c12runnerprofile/completion_state.go internal/c12runnerprofile/completion_state_test.go internal/c12runnerprofile/completion_child_runtime.go internal/c12runnerprofile/completion_child_runtime_test.go internal/c12runnerprofile/completion_child_runtime_windows.go internal/c12runnerprofile/completion_child_runtime_windows_test.go internal/c12runnerprofile/completion_child_runtime_unsupported.go internal/c12runnerprofile/platform_resume.go internal/c12runnerprofile/platform_resume_test.go internal/c12runnerprofile/native_trusted_time_runtime.go internal/c12runnerprofile/native_trusted_time_runtime_test.go internal/c12runnerprofile/native_trusted_time_runtime_linux.go internal/c12runnerprofile/native_trusted_time_runtime_linux_test.go internal/c12runnerprofile/native_trusted_time_runtime_unsupported.go internal/c12runnerprofile/state_guard.go internal/c12runnerprofile/state_guard_test.go internal/c12runnerprofile/state_guard_windows.go internal/c12runnerprofile/state_guard_windows_test.go internal/c12runnerprofile/state_guard_linux.go internal/c12runnerprofile/state_guard_linux_test.go testdata/c12/windows/runner-profile.v1.json testdata/c12/diagnostic/runner-profile.v1.json internal/c12evidence/wal.go internal/c12evidence/wal_test.go internal/c12evidence/cleanup.go internal/c12evidence/cleanup_test.go internal/c12evidence/runnerhelper.go internal/c12evidence/runnerhelper_test.go cmd/talenro-c12-runner-helper/main.go cmd/talenro-c12-runner-helper/main_test.go cmd/talenro-c12-runner-launcher/main.go cmd/talenro-c12-runner-launcher/main_test.go cmd/talenro-c12-runner-launcher/launcher_windows.go cmd/talenro-c12-runner-launcher/launcher_windows_test.go cmd/talenro-c12-runner-launcher/launcher_linux.go cmd/talenro-c12-runner-launcher/launcher_linux_test.go
git commit -m "feat: add exact C1.2 ownership recovery"
```

### Task 3: Closed-set Go, executable and OCI artifact scanner

**Files:**
- Create: `internal/artifactscan/rules.go`
- Create: `internal/artifactscan/buildinputs.go`
- Create: `internal/artifactscan/buildinputs_test.go`
- Create: `internal/artifactscan/godeps.go`
- Create: `internal/artifactscan/executable.go`
- Create: `internal/artifactscan/oci.go`
- Create: `internal/artifactscan/scan.go`
- Create: `internal/artifactscan/scan_test.go`
- Create: `cmd/talenro-artifact-scan/main.go`
- Create: `cmd/talenro-artifact-scan/main_test.go`
- Modify: `cmd/talenro-c12-runner-helper/main.go`
- Modify: `cmd/talenro-c12-runner-helper/main_test.go`
- Create: `testdata/c12/scanner/renamed-core.json`
- Create: `testdata/c12/scanner/archive-core.json`
- Create: `testdata/c12/scanner/base-layer-core.json`
- Create: `testdata/c12/scanner/forged-module.json`
- Create: `testdata/c12/scanner/artifact-catalog-valid.v1.json`
- Create: `testdata/c12/scanner/outer-image-build-context-valid.v1.json`
- Create: `testdata/c12/scanner/operations-build-policy-valid.v1.json`
- Test: `internal/artifactscan/scan_test.go`
- Test: `internal/artifactscan/buildinputs_test.go`
- Test: `cmd/talenro-artifact-scan/main_test.go`

**Interfaces:**
- Consumes during B10: the strict catalog schema and synthetic exact-cover fixture `testdata/c12/scanner/artifact-catalog-valid.v1.json`, three proprietary fixture binaries, a separately classified authority-operations fixture binary/image, both reproducible fixed-launcher binary/SBOM/provenance role triples and the remaining fixed scanner inputs. The production loader path is already frozen to the future sole tracked `testdata/c12/artifact-catalog.v1.json`, but absence of that B11-owned file is expected during B10 and no authoritative scan/scope may be minted. Plan 09 Task 6A must create the final instance and prove its exact cover before Task 8 can run the first authoritative scan. A catalog freezes logical role、tracked build source/policy/license path、canonical run-owned relative output path and required supply-chain record kinds；it contains no final tree/output digest. The synthetic launcher declarations validate the schema/exact-cover rules and B10-frozen byte identities；they do not turn the absent final production catalog into a B10 output.
- Produces: scanner version `talenro-artifact-scan/v1`, `RulesetDigest() contracts.Digest`, the sole strict fixed-token loaders for `C12OuterImageBuildContextV1` and `C12OperationsBuildPolicyV1`, strict `C12ArtifactCatalogV1` loading plus `ArtifactCatalogDigest`, `ScanClosedSet(context.Context, ScanRequest) (ScanReceiptV1, SupplyChainEvidenceBundleV1, error)`, the sole domain-separated `AuthorityOperationsBinaryAndImageDigest` derivation, deterministic typed supply-chain records/bundle, and the declared closed public `talenro-artifact-scan` CLI command enum `spec-digest`, `lock-core`, `lock-outer-images`, `verify-go-deps`, `scan-release`, `scan-authority-operations`, `scan-oci`, `scan-closed-set`, and `verify-scope`. CLI dispatch/help tests must enumerate that exact set；an implemented-but-undeclared or declared-but-undispatched public command fails. Plan 09 later adds only a parent-authenticated private completion child mode absent from public help. `scan-authority-operations` is a focused preflight with pass/fail diagnostics only；`scan-closed-set` is the one-root scanner used by fixtures and invoked twice through the package API by the native helper's in-process `build-closed-set` producer. Only that helper subcommand may atomically publish the authoritative canonical `ScanReceiptV1` plus bounded auditable bundle that include the auxiliary tuple/record set/catalog and every release/image/asset set；`talenro-artifact-scan build-closed-set` is deliberately nonexistent. `spec-digest` is only a thin CLI over `BuildTrackedC12SpecSet`, not a second builder.

`lock-outer-images` is the sole declared relock command. B10 requires both `--windows-runner-profile`/`--rewrite-windows-runner-profile` and `--diagnostic-runner-profile`/`--rewrite-diagnostic-runner-profile` alongside Compose/context/integration/lock flags. Before relock it validates only the freshly built PE/ELF candidate identities, tracked profile/install policy, exact toolchain native-role projection and diagnostic nonpublishing projection, then atomically rewrites exactly Compose, integration manifest, toolchain lock, Windows profile and diagnostic profile input projections. B10 pre-relock neither authenticates/opens a production bootstrap or runner session nor requires an install receipt, fixed-store generation, reusable evidence or final artifact catalog；those do not exist for the candidate outputs yet. The fixed future B11 all-or-none group is `--artifact-catalog`, `--trusted-time-provider-profile`, and all five `--{windows,diagnostic,platform,authority,completion}-runner-profile` flags plus the six matching provider/runner `--rewrite-*profile` flags. It validates the five candidate runner roles, native executable/client/helper/private-child mappings, diagnostic zero-capability projection and catalog exact cover, then atomically rewrites Compose, integration manifest, toolchain lock, provider profile and all five runner-profile projections while leaving the digest-free catalog and build policies byte-identical. B11 pre-relock likewise consumes no production bootstrap/session/install receipt. A partial group or candidate/profile path override rejects. No executable embeds any final lock/profile digest, so this projection order is acyclic.

`internal/artifactscan/buildinputs.go` freezes these production interfaces independently of Task 4/P09 instances:

```go
type TrackedBuildContextMemberV1 struct {
	RepoRelativePath string
	Mode             string
	BlobSHA256       contracts.Digest
}

type C12OuterImageBuildContextV1 struct {
	SchemaVersion  string
	ImageRole      string
	OrderedMembers []TrackedBuildContextMemberV1
}

type ReproducibleGoOCIBuildRecipeV1 struct {
	SourceDateEpoch       int64
	TrimpathPolicy        string
	BuildIDPolicy         string
	TimestampPolicy       string
	TarOrderPolicy        string
	UIDGIDModeXAttrPolicy string
	CompressionPolicy     string
}

type C12OperationsBuildPolicyV1 struct {
	SchemaVersion          string
	SourceRole             string
	BinaryCatalogRole      string
	ImageCatalogRole       string
	DockerfileToken        string
	SourceTokens           []string
	BuildRecipe            ReproducibleGoOCIBuildRecipeV1
	RequiredRecordKinds    [9]SupplyChainRecordKind
}

type AuthenticatedTrackedBuildInputReader interface {
	authenticatedTrackedBuildInputReader() // opaque exact-tree reader
}

type ValidatedOuterImageBuildContext struct {
	sealed privateValidatedOuterImageBuildContext
}

type ValidatedOperationsBuildPolicy struct {
	sealed privateValidatedOperationsBuildPolicy
}

func BindAuthenticatedTrackedBuildInputReader(context.Context, c12evidence.ProtectedTreeFactsHandle) (AuthenticatedTrackedBuildInputReader, error)
func LoadFixedOuterImageBuildContexts(AuthenticatedTrackedBuildInputReader) ([3]ValidatedOuterImageBuildContext, error)
func LoadFixedOperationsBuildPolicy(AuthenticatedTrackedBuildInputReader) (ValidatedOperationsBuildPolicy, error)
func CanonicalOperationsBuildPolicy(C12OperationsBuildPolicyV1) ([]byte, contracts.Digest, error)
```

The outer-context loader recognizes exactly `deploy/c12/context-init.build-context.v1.json`, `deploy/c12/inner-daemon.build-context.v1.json`, and `deploy/c12/verifier.build-context.v1.json`, with roles `context_init|inner_daemon|verifier` in that order. Strict decode rejects unknown/duplicate/null/default fields, noncanonical/absolute/escaping members, missing/extra/duplicate token, wrong Git mode/blob digest, symlink/reparse, role swap and any rewrite output member. It exact-covers each authenticated tracked context against the ordered path/mode/blob digest set and derives `SHA-256(ASCII("TALENRO-C12-OUTER-IMAGE-BUILD-CONTEXT-V1") || 0x00 || JCS(instance))`; there is no token/root/digest/member-list override.

The operations-policy loader reads only fixed future token `testdata/c12/authority/operations-build-policy.v1.json`; P08 owns its schema/loader and synthetic fixture, while Plan 09 Task 6A owns the final tracked instance. Strict validation closes source/binary/image roles, Dockerfile/source tokens, the reproducible Go/OCI recipe (`SOURCE_DATE_EPOCH`, trimpath/build-ID, config/history/layer timestamps, tar order/uid/gid/mode/xattrs and compression) and the exact nine supply-chain record kinds. Its digest is `SHA-256(ASCII("TALENRO-C12-OPERATIONS-BUILD-POLICY-V1") || 0x00 || JCS(policy))`. It contains no repo commit/tree, built binary/image/receipt/catalog/profile digest or output path, and its source/context set excludes itself plus every relock output, so B10's already locked helper can parse the later P09 instance without a source/output digest cycle. `BindAuthenticatedTrackedBuildInputReader` is artifactscan's sole reader mint：inside the already authenticated direct child it linearly consumes one nonzero `c12evidence.ProtectedTreeFactsHandle`, rechecks the handle's committed/staged locator、canonical-tree、snapshot/object-database and child-process binding, and copies only the fixed-token bounded read methods into an artifactscan-private sealed reader. A nil/fabricated/reused/cross-child handle fails before any blob read. The reader exposes no root/path/FD/HANDLE or arbitrary token, and callers cannot substitute a filesystem reader or implement the interface. The production dependency remains one-way `artifactscan -> c12evidence`; c12evidence production never imports artifactscan.

- [ ] **Step 1: RED — add negative core-concealment tests**

```go
func TestScannerRejectsEveryCoreConcealment(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"renamed-core", "archive-core", "base-layer-core", "forged-module"} {
		t.Run(name, func(t *testing.T) {
			request := materializeNegativeFixture(t, name)
			if _, _, err := ScanClosedSet(t.Context(), request); !errors.Is(err, ErrForbiddenCoreMaterial) {
				t.Fatal("scanner accepted concealed core material")
			}
		})
	}
}
```

Add `TestAuthorityOperationsArtifactIsReproducibleAcrossRoots`: materialize the same tracked tree in two different absolute roots, build the auxiliary binary and OCI layout independently with the locked toolchain, and require byte-identical binary digest, OCI manifest/config/layer digests and final auxiliary digest. Mutate source epoch, tar order, uid/gid/mode or config/layer timestamp and require the build-receipt validator to fail rather than mint a different accepted digest.

Add `TestTrackedArtifactCatalogRequiresExactCover` against `artifact-catalog-valid.v1.json`: strict-decode the fixed schema and require the three ordered proprietary release roles, one operations auxiliary binary plus its dedicated image, distinct `c12_runner_helper_windows_amd64` PE and `c12_runner_helper_linux_amd64` static-ELF roles built from the same source, distinct `c12_runner_launcher_windows_amd64` PE and `c12_runner_launcher_linux_amd64` ELF roles plus each launcher's exact SBOM/provenance declarations, the prebuilt `talenro_artifact_scan_windows_amd64` private-completion-child role, every synthetic production/outer OCI declaration, PostgreSQL 18.4、Redis 8.8.1、NATS 2.14.3 integration-only OCI layout, both external-core locks, verifier/core assets and every required supply-chain record role exactly once. Scan both directions over the fixture root. Missing/extra/duplicate/unknown role, role swap, cross-OS helper/launcher substitution, launcher binary without its matching SBOM/provenance, helper aliased as launcher, untracked declaration, absolute/escaping path, symlink/reparse target or catalog/path-list override returns `ErrArtifactCatalogMismatch` before any clean receipt. A separate absent-production-path test proves B10 returns the fixed deferred/not-authoritative result rather than silently substituting the fixture. Plan 09 Task 6A acceptance repeats this exact-cover test against the actual tracked declarations and final catalog；Plan 09 Task 8 is the first positive authoritative production load.

Add `TestLockOuterImagesAtomicRewrite`: invoke `lock-outer-images` against three synthetic outer image layouts and the three integration dependency layouts, plus the fixed `windows_scopes`/`staged_diagnostic` profile tokens, synthetic candidate helper/Git/build-tool identities, the already frozen Windows/Linux launcher binary+SBOM+provenance identities and their toolchain native-role projections. Deliberately provide no production bootstrap/session/install receipt and require any attempt to open one during pre-relock to fail the test. Require one atomic transaction to rewrite exactly `deploy/c12/compose.outer.yaml`, `testdata/c12/integration-manifest.v1.json`, `testdata/c12/toolchain-lock.v1.json`, `testdata/c12/windows/runner-profile.v1.json` and `testdata/c12/diagnostic/runner-profile.v1.json` with matching verified context/helper/launcher/install-policy projections. The diagnostic projection must retain empty scripts/channels, `none/nonpublishing` provider/evidence purposes and Linux helper/Git/build-tool exact cover；launcher roles remain toolchain/profile inputs and never receipt roles or rewrite outputs. A second B11-mode vector supplies the fixed `--artifact-catalog`、`--trusted-time-provider-profile`、all five `--{windows,diagnostic,platform,authority,completion}-runner-profile` flags and all six provider/runner profile rewrite flags, requires strict exact-cover including both candidate helper roles, both byte-identical B10-frozen launcher role triples and the candidate prebuilt completion child, and requires the same transaction to bind the canonical catalog/role projection into the lock、rewrite the provider profile `ClientBuildDigest` and rewrite all five runner-profile provider/helper/launcher/install/lock projections while leaving the digest-free catalog and both build policies byte-identical；it likewise has no production bootstrap/session/install receipt before its nine-output rewrite. Neither helper/launcher/completion child embeds any final lock/profile digest, so the projection graph is acyclic. Failure before/after any staged rewrite, wrong old digest, concurrent file/catalog/profile change, missing launcher/SBOM/provenance or other vulnerability/license/notice/source/secret record, candidate/install-policy mismatch, diagnostic capability expansion or fsync/rename failure must leave Compose、integration manifest、toolchain lock、provider/five runner profiles、catalog and policies byte-identical and retain no partial/temp file. CLI/help tests freeze the B10 required flags `--compose`, `--context-root`, `--integration-manifest`, `--output-lock`, `--windows-runner-profile`, `--diagnostic-runner-profile`, `--rewrite-compose`, `--rewrite-integration-manifest`, `--rewrite-windows-runner-profile`, `--rewrite-diagnostic-runner-profile` plus the later all-or-none fixed flags `--artifact-catalog`, `--trusted-time-provider-profile`, `--platform-runner-profile`, `--authority-runner-profile`, `--completion-runner-profile`, `--rewrite-trusted-time-provider-profile`, `--rewrite-platform-runner-profile`, `--rewrite-authority-runner-profile`, `--rewrite-completion-runner-profile`; a partial group, duplicate B10 profile flag or unknown alias rejects.

Add `TestAuthorityOperationsTupleCannotBeSpliced`: mutate each of `binary_sha256`, `image_manifest_digest`, `build_receipt_digest` and `dependency_closure_digest` while retaining the old combined digest；then try same binary/different image, same image/different binary and a caller-selected combined digest. Every case must fail recomputation before receipt output.

Add `TestFixedOuterImageBuildContextSchemaAndExactCover` and `TestFixedOperationsBuildPolicySchemaAndAcyclic` against the two synthetic fixtures owned by this task. The first test verifies the three hard-coded tokens/roles, strict schema/JCS/domain, ordered path/mode/blob exact cover and rewrite-output exclusion. The second verifies the one fixed future operations-policy token, closed recipe/roles/nine records, unknown/duplicate/null/tamper rejection and proves that policy source closure excludes itself, final outputs, catalog and every relock output. A caller token/root/digest/member override must fail before build.

Add `TestCaptureStagedAndBuildClosedSetCleanupBarrier` and `TestCaptureStagedAndBuildClosedSetCommandSurface`. The zero-argument command must call exactly `OpenInstalledNativeRunnerSession(ctx, NativeRunnerStagedDiagnostic)` and reject an otherwise valid `authority`, platform, Windows or completion session before any Git/resource operation. Its unexported diagnostic facade uses only that session's canonical object-database、owner-only external-root、recovery-slot and installed Linux helper/Git/build-tool handles to build the sealed cleanup plan, then must execute `BeginProtectedWALKeyBootstrap→CreateOrRecoverProtectedWALKeyBootstrap→SealAndSelectProtectedWALKeyBootstrap→OpenSelectedOwnershipWAL` **before** the first run-root/minimal-object-database/snapshot/build-root/producer create. The template supplies only trusted `IssuedAt|ExpiresAt`; every machine/slot/parent/provider/resource/WAL projection is filled internally. It then uses only the sealed WAL and the same tombstone finalizer on every exit. Inject every plan/intent/key/capsule/selected-WAL reservation/create/bootstrap/actual fsync, two-root/producer and tombstone replacement/key/unlink seam, including post-expiry and provider/trusted-time outage；recovery sees only intent-abort, cleanup-only, successor resume, cleaned tombstone or absent as selected by the authenticated guard and never probes a second route. The diagnostic session has no script/channel/provider/evidence/signer/publisher handle, so the command can emit only bounded nonpublishing diagnostic status and can never sign or publish a receipt/bundle/scope. Public help contains the zero-argument parent exactly once; every path/profile/provider/channel/attestor/handle/root/output flag or legacy positional form rejects.

- [ ] **Step 2: Run scanner tests and verify RED**

Run: `go test ./internal/artifactscan ./cmd/talenro-c12-runner-helper -run '^TestScannerRejectsEveryCoreConcealment$|^TestFixedOuterImageBuildContextSchemaAndExactCover$|^TestFixedOperationsBuildPolicySchemaAndAcyclic$|^TestCaptureStagedAndBuildClosedSet' -count=1`

Expected: FAIL because scanner/build-input packages, fixtures and staged parent do not exist.

- [ ] **Step 3: GREEN — implement Go dependency and embed inventory**

Decode concatenated `go list -deps -json` records and require `ImportPath`, `Module.Path`, `GoFiles`, `CgoFiles`, `CompiledGoFiles`, `EmbedFiles`, `TestEmbedFiles`, and `XTestEmbedFiles` to be parseable. Reject either upstream module path or child path, any nonempty CgoFiles for a release or the authority-operations auxiliary build, any embedded file matching an approved core executable/layout/layer digest, or embedded PE/ELF/archive payload. Cross-check package/build settings against `go version -m` and the build receipt.

- [ ] **Step 4: Run Go dependency tests and verify GREEN**

Run: `go test ./internal/artifactscan -run '^TestGoDeps|^TestEmbed' -count=1`

Expected: PASS for repository packages and FAIL for the forged-module fixture.

- [ ] **Step 5: GREEN — implement executable and OCI traversal**

Use `debug/elf`, `debug/pe`, `debug/buildinfo`, `archive/tar` and OCI descriptor digests. For Linux releases and the authority-operations auxiliary binary require static ELF with no `PT_INTERP`, `DT_NEEDED`, dynamic imports or external-link receipt. The locked build fixes `SOURCE_DATE_EPOCH`, Go trimpath/build IDs, OCI config/history timestamps, layer tar path order, uid/gid, mode, xattrs and gzip parameters；the receipt rejects any unset/ambient value. Walk every OCI manifest/config/layer by verified digest；require the dedicated operations image to contain exactly the scanned operations binary plus its approved finite runtime files and to appear in the production image set；inventory regular files, symlinks, whiteouts, nested archives and executable headers. Reject a core digest/module namespace/executable payload in any proprietary or auxiliary binary, production image or unaccounted outer layer. The auxiliary classification cannot increase `ReleaseInputDigests[3]` or inherit a release/license exemption.

Generate and verify a closed typed operations record set only in a run-owned external directory after the final tree exists. Every record carries `kind`、`status=pass`、`record_digest` and a strict subject containing the exact `AuthorityOperationsArtifactTupleV1`, repo commit/tracked-tree and toolchain/build-policy digests；kind-specific fields have a closed present/null matrix. Derive `subject_digest` from the strict subject, then `record_digest = SHA-256(ASCII("TALENRO-C12-SUPPLY-CHAIN-RECORD-V1") || 0x00 || JCS(record_body_without_record_digest))`; the record digest is never part of its own preimage. Binary/image SBOM records exact-cover parsed Go dependency plus OCI file/package closures. Source and provenance bind the tracked source set, build recipe, toolchain and full tuple. Vulnerability records bind the exact SBOM digests and toolchain-pinned vulnerability database digest. License and notice/source-obligation records exact-cover every SBOM component and the tracked approved obligation document. Secret-scan records bind the exact source set, binary and image subjects. Cross-artifact clean records, wrong subject/tree/toolchain, stale vulnerability DB, missing/extra SBOM dependency, wrong/self-included record digest, missing/extra/duplicate/unknown record kind or any non-pass status fail before the record-set digest is derived. No tracked lock/provenance file embeds the final repo tree/output digest, so the final post-commit record generation has no self-reference.

- [ ] **Step 6: Run executable/OCI tests and verify GREEN**

Run: `go test ./internal/artifactscan -run '^TestELF|^TestPE|^TestOCI|^TestScannerRejectsEveryCoreConcealment$' -count=1`

Expected: PASS for clean synthetic layouts; renamed, wrapper-loaded, archive, base-layer and forged-module fixtures all return `ErrForbiddenCoreMaterial`.

- [ ] **Step 7: GREEN — add the bounded CLI and machine-readable receipt**

```go
type SupplyChainRecordKind string

const (
	SupplyChainRecordBinarySBOM             SupplyChainRecordKind = "binary_sbom"
	SupplyChainRecordImageSBOM              SupplyChainRecordKind = "image_sbom"
	SupplyChainRecordSource                 SupplyChainRecordKind = "source"
	SupplyChainRecordProvenance             SupplyChainRecordKind = "provenance"
	SupplyChainRecordVulnerability          SupplyChainRecordKind = "vulnerability"
	SupplyChainRecordLicense                SupplyChainRecordKind = "license"
	SupplyChainRecordNoticeSourceObligation SupplyChainRecordKind = "notice_source_obligation"
	SupplyChainRecordSecretScan             SupplyChainRecordKind = "secret_scan"
	SupplyChainRecordCrossArtifactClean     SupplyChainRecordKind = "cross_artifact_clean"
)

type SupplyChainRecordStatus string

const SupplyChainRecordStatusPass SupplyChainRecordStatus = "pass"

type SupplyChainRecordSubjectV1 struct {
	AuthorityOperationsArtifact c12evidence.AuthorityOperationsArtifactTupleV1
	RepoCommit                   string
	TrackedTreeDigest            contracts.Digest
	ToolchainDigest              contracts.Digest
	BuildPolicyDigest            contracts.Digest
}

// Optional digest pointers implement the closed kind-specific present/null
// matrix.  Unknown fields and a non-nil field outside its kind are invalid.
type SupplyChainRecordV1 struct {
	SchemaVersion               string
	Kind                        SupplyChainRecordKind
	Status                      SupplyChainRecordStatus
	Subject                     SupplyChainRecordSubjectV1
	SubjectDigest               contracts.Digest
	DependencyClosureDigest     *contracts.Digest
	SourceSetDigest             *contracts.Digest
	BuildRecipeDigest           *contracts.Digest
	BinarySBOMDigest            *contracts.Digest
	ImageSBOMDigest             *contracts.Digest
	VulnerabilityDatabaseDigest *contracts.Digest
	LicenseClosureDigest        *contracts.Digest
	ObligationDocumentDigest    *contracts.Digest
	SecretScanInputDigest       *contracts.Digest
	CrossArtifactCleanDigest    *contracts.Digest
	RecordDigest                contracts.Digest
}

type SupplyChainRecordReferenceV1 struct {
	Kind                     SupplyChainRecordKind
	SubjectDigest            contracts.Digest
	RecordDigest             contracts.Digest
	Status                   string
	ObligationDocumentDigest contracts.Digest
}

type SupplyChainEvidenceBundleV1 struct {
	SchemaVersion               string
	ArtifactCatalogDigest       contracts.Digest
	AuthorityOperationsArtifact c12evidence.AuthorityOperationsArtifactTupleV1
	Records                     [9]SupplyChainRecordV1
}

type ScanReceiptV1 struct {
	SchemaVersion        string
	ScannerVersion       string
	RulesetDigest        contracts.Digest
	RepoCommit           string
	TrackedTreeDigest    contracts.Digest
	SpecDigest           contracts.Digest
	ToolchainDigest      contracts.Digest
	ReleaseInputDigests  [3]contracts.Digest
	AuthorityOperationsArtifact c12evidence.AuthorityOperationsArtifactTupleV1
	AuthorityOperationsBinaryAndImageDigest contracts.Digest
	AuthorityOperationsSupplyChainRecords [9]SupplyChainRecordReferenceV1
	AuthorityOperationsSupplyChainRecordSetDigest contracts.Digest
	SupplyChainEvidenceBundleDigest contracts.Digest
	ArtifactCatalogDigest contracts.Digest
	ImageSetDigest        contracts.Digest
	ProductionImageSet   contracts.Digest
	OuterImageSet        contracts.Digest
	VerifierAssetSet     contracts.Digest
	CoreAssetSet         contracts.Digest
	ExpectedProcessImagePolicyDigest contracts.Digest
	Result               string
}

func ValidateAndSealFinalClosedSetProjection(context.Context, ScanReceiptV1, SupplyChainEvidenceBundleV1, c12evidence.FinalClosedSetProjectionSink) (c12evidence.FinalClosedSetProjection, error)
```

The only successful scan result is `clean_closed_set`. Strict-load the catalog from the exact tracked path and compute `ArtifactCatalogDigest = SHA-256(ASCII("TALENRO-C12-ARTIFACT-CATALOG-V1") || 0x00 || JCS(catalog))` only after exact-cover succeeds. Populate `AuthorityOperationsArtifact` from measured data, then recompute `AuthorityOperationsBinaryAndImageDigest = SHA-256(ASCII("TALENRO-AUTHORITY-OPERATIONS-ARTIFACT-V1") || 0x00 || JCS({binary_sha256,image_manifest_digest,build_receipt_digest,dependency_closure_digest,ruleset_digest}))` only after binary build-info/dependency, reproducibility and dedicated OCI membership checks pass；there is no binary-only or caller-digest shortcut. Populate the fixed nine record references in closed kind order: `binary_sbom`, `image_sbom`, `source`, `provenance`, `vulnerability`, `license`, `notice_source_obligation`, `secret_scan`, `cross_artifact_clean`. `ObligationDocumentDigest` is nonzero only for the license and notice/source-obligation kinds and must name their exact tracked approved documents；all other kinds require zero. Independently compute `AuthorityOperationsSupplyChainRecordSetDigest = SHA-256(ASCII("TALENRO-AUTHORITY-OPERATIONS-SUPPLY-CHAIN-RECORD-SET-V1") || 0x00 || JCS(AuthorityOperationsSupplyChainRecords))` only after the exact subject/body/cross-record validations above. Serialize the full bounded `SupplyChainEvidenceBundleV1` (maximum 4 MiB, digests/finite enums/dependency and file identities only, no file contents、secret matches、credentials or absolute paths) and derive `SupplyChainEvidenceBundleDigest = SHA-256(ASCII("TALENRO-C12-SUPPLY-CHAIN-EVIDENCE-BUNDLE-V1") || 0x00 || JCS(bundle))`. Every reference must be the exact projection of its corresponding bundle record. This retained canonical bundle lets completion and later revalidation prove the same scanned set instead of a parallel clean set. `ExpectedProcessImagePolicyDigest` is a deterministic catalog projection `{scope -> ordered required_roles, ordered allowed_roles, exact executable/image tuples}` for all four closed scopes and contains no PID、run ID、boot ID、time、runner or observed process receipt. Actual per-run process-image receipts remain in each scope's `TestResultDigest`, `PlatformEvidenceV1.PostStartReceiptSetDigest` or authority result digest；each gate selects only its named scope policy, requires every required role exactly once and no role outside the allowed set, then signs the observed receipt-set digest plus expected scope-policy digest. A missing/extra/role-swapped process or another scope's otherwise valid tuple rejects, while run-specific bytes never change the canonical scan receipt.

Derive `ImageSetDigest = SHA-256(ASCII("TALENRO-C12-IMAGE-SET-V1") || 0x00 || JCS(sorted({catalog_role,image_or_layout_manifest_digest})))` over every catalog entry whose closed kind is OCI image/layout；the role is part of the preimage, every role occurs once, and the operations tuple image manifest must be the exact operations-role member. `ProductionImageSet` and `OuterImageSet` are deterministic subset projections only and cannot substitute for this full set. The receipt binds the tuple, record/bundle values, catalog digest, exact repo commit/typed Git tree locator/canonical tracked-tree digest, Task 1 builder-derived `SpecDigest`, locked toolchain, three release inputs and all image/asset/expected-process-policy sets. `CanonicalScanReceipt` JCS-encodes every field shown above and hashes `ASCII("TALENRO-C12-SCAN-RECEIPT-V1") || 0x00 || JCS(receipt)` to return the `ArtifactScanDigest` used by completion；the digest is not embedded back into the receipt. `ValidateAndSealFinalClosedSetProjection` is the sole typed bridge to c12evidence：it repeats strict validation and canonicalization of the clean receipt and complete bundle, exact-matches all nine references/bodies、catalog/tree/toolchain/result and digest fields, then and only then invokes `sink.SealValidatedClosedSet(canonicalReceipt,canonicalBundle)` exactly once. Its `go/types` owner gate requires that method call to exist only in `internal/artifactscan/scan.go`, requires the sink/result private markers to have no external concrete, and rejects a callback、interface implementation、raw-byte overload or path that reaches Seal without the validator. `spec-digest` accepts only canonical repo root plus the native helper's protected typed tree locator and separately captured canonical tracked-tree digest, calls `BuildTrackedC12SpecSet`, recomputes/compares the latter, and emits the resulting digest；it has no manifest/member/hash/tree-OID/tracked-tree/spec-digest override. `verify-scope` recomputes the same values and rejects old two-member/single-file spec digests or any tree/tuple/combined/supply-chain-bundle/catalog/image/full-receipt splice.

Freeze the only operator-callable committed parent form, plus one staged diagnostic parent form. Both are zero-argument native modes：the helper obtains the canonical object-database locator、owner-only external root capability and WAL/bootstrap identities only from the locked runner profile/native bootstrap, then captures the tree and creates/WAL-records its independently opaque children itself. No repo/WAL/run/output/tree/build-root path、digest or handle is accepted from argv/environment:

```text
<absolute-toolchain-locked-native-talenro-c12-runner-helper> capture-and-build-closed-set
<absolute-toolchain-locked-native-talenro-c12-runner-helper> capture-staged-and-build-closed-set
```

Task 3 fully implements and B10 locks `capture-staged-and-build-closed-set`; Plan 09 consumes it and may not add, upgrade or replace its barrier. This parent opens only `NativeRunnerStagedDiagnostic`; its role facade first derives the sealed complete reservation/cleanup plan, selects the protected-key intent, create-or-recovers its one deterministic nonexportable key, seals+guard-selects the complete cleanup-only capsule with the two-field time template, and linearly consumes selected ownership through `OpenSelectedOwnershipWAL`. Only after that method has exact-created/reopened the fixed WAL and file/directory-fsynced its bootstrap+actual may the parent create a run root、staged snapshot、minimal object database、either build root or private producer. Every success, failure, process crash and cold-reboot recovery follows the guard-selected intent-abort or cleanup-only/resume-to-cleaned-tombstone route, including expiry-tolerant delete-only recovery；no error authorizes fallback. The staged parent can return only bounded `deferred_not_authoritative` diagnostics with reusable receipt/bundle/scope outputs absent; it has no final-runner, channel, provider, evidence-root or signing capability. `TestCaptureStagedAndBuildClosedSetCleanupBarrier` is part of this task's GREEN command, helper command-surface test, Task 7 B10 lock validation and final `go test ./...`; a later plan cannot be the first test or implementation owner.

Task 4 adds exactly one nested parent invocation, fixed in the verifier image entrypoint and never operator-composed: `/usr/libexec/talenro-c12/talenro-c12-runner-helper validate-capsule-and-build-closed-set --tree-facts-capsule /run/c12-input/tree-facts.capsule --git-object-database /run/c12-input/object-db --workspace /workspace --integration-manifest /workspace/testdata/c12/integration-manifest.v1.json`. Those four paths and the executable identity are compile-time toolchain/image-locked constants, not caller-selected flags；dispatch rejects another value or argument shape. There is deliberately no authority-script argument: after reopening the minimal object database, verifying its capsule-bound identity/closure and privately re-resolving the typed locator, the parent derives and reopens the script only from its authenticated snapshot handle. It obtains output/root/WAL destinations from authenticated run/capsule/manifest facts and owns the complete fixed stage state machine, not environment or a shell protocol.

Inside that parent session only, the exact same prebuilt helper may be spawned with the private transcript `build-closed-set --repo-root <object-db> --tracked-tree-digest <parent-computed-64hex> --tree-facts-handle <explicitly-inherited-handle> --build-root-a <parent-created-opaque-a> --build-root-b <parent-created-opaque-b> --receipt-output ... --supply-chain-bundle-output ... --parent-session-handle <explicitly-inherited-one-use-capability>`. The private subcommand is absent from public help, rejects direct/operator invocation and requires both inherited handles to bind the parent PID/start token、helper executable identity、run/WAL/snapshot/root facts. The executable is the prebuilt PE/ELF helper itself, resolved by absolute locked path and verified against the exact catalog/toolchain OS/arch/build digest before launch；the parent never goes through `go run`、`go tool`、a temporary binary、shell or PATH. The authenticated tree handle contains exact snapshot directory identity、`GitTreeLocatorV1`、canonical `TrackedTreeDigest`、capture kind (`staged` or `committed`) and WAL record digest. The child consumes it once and rejects a non-inherited/reopened/path-backed handle, unknown field, signature/MAC/run/WAL mismatch, snapshot identity drift, parent digest disagreement or a repo-root worktree fallback. `--repo-root` only locates the canonical object database needed to re-resolve the typed locator；it is never a source tree.

The two build roots must have independently generated 128-bit lowercase-hex opaque child names (never literal shared `build-a`/`build-b` names), be different direct children of the helper-created run root, repository-external, empty, exclusive, regular-directory/no-reparse, have distinct file identities, and be recorded by helper WAL before materialization. `build-closed-set` is the sole producer: it uses only the exact snapshot proven by that handle, fixed catalog/toolchain/integration manifest, locked builders and already present digest-addressed OCI dependency layouts；all build/container network is disabled and any registry pull、tag lookup or ambient dependency download fails. It builds every role twice, calls `ScanClosedSet` once per root, compares every catalog-relative regular file/OCI descriptor plus canonical receipt/bundle byte-for-byte, and publishes only the first identical pair after file+directory fsync. A `staged` capture with empty `CommitOID` always returns `deferred_not_authoritative` with final outputs absent even when both roots compare clean；authoritative output requires `capture_kind=committed`, a nonempty commit OID resolving to the exact tree OID, matching snapshot/canonical digest and a clean committed source closure. Before the B11 final catalog roles exist it likewise remains deferred. Direct `scan-closed-set` accepts a single read-only fixture root and emits only run-internal temporary comparison files when its authenticated parent is `build-closed-set`; a wrapper/runbook/direct caller cannot use it to publish a receipt. Neither command exposes catalog、per-role path-list、expected digest、receipt override、tree-facts file/path or alternate builder. All output/build paths must be distinct bounded non-repo non-symlink/reparse paths. Other focused commands cannot emit an authoritative receipt or retained bundle. Reject stdin for binary/OCI data, cap diagnostics at 4 KiB of finite categories, and never print file content, certificate, key or core output.

At this task only, extend `cmd/talenro-c12-runner-helper` with the two zero-argument public parent modes、private `build-closed-set` child and parent-private `scope-from-verifier-result` unsigned projection builder. Public help/dispatch tests exact-enumerate the parents and reject any legacy path-bearing form；the projection builder cannot be operator-invoked. The command package may import both `internal/artifactscan` and `internal/c12evidence`: the parent owns capture/root/WAL/session capabilities；the private producer consumes the authenticated tree handle, invokes `ScanClosedSet` twice through the package API and owns authoritative receipt/bundle publication；the projection builder strictly validates/recomputes that receipt/bundle and builds the closed expected policy before c12evidence validates the signed `C12VerifierResultV1` and returns an opaque unsigned-scope-facts handle to the same long-lived parent. The helper library itself remains artifactscan-free. It requires a terminal-clean authenticated lifecycle WAL, bounded receipt/bundle/result bytes retained through helper-owned handles and strict outer facts, but accepts **no attestor handle** and cannot serialize/sign/publish a scope. Its `runner_facts_digest` is a domain-separated canonical projection that necessarily includes the distinct `TreeFactsAttestorIdentity`, complete verified `C12TreeFactsCapsuleV1` envelope digest and minimal-object-database identity/closure in addition to locked runner/helper facts；it derives `TestResultDigest = SHA-256(ASCII("TALENRO-C12-WINDOWS-SCOPE-TEST-RESULT-V1") || 0x00 || verifier_result_envelope_digest || runner_facts_digest || daemon_facts_digest || job_or_bash_portability_digest || wal_terminal_digest)`. Task 7 may sign the two returned projections only after its sealed both-lifecycles-complete/equal/full-clean barrier. Scope-attestation-as-capsule、capsule-attestation-as-scope and a valid capsule/object-database identity/digest splice each fail while projections remain unsigned. It accepts no caller policy/digest/canonical bytes/projection/signature and emits no scope on cleanup failure.

- [ ] **Step 8: Run CLI tests and verify GREEN**

Run: `go test ./cmd/talenro-artifact-scan ./cmd/talenro-c12-runner-helper ./internal/artifactscan -run '^Test|^Example' -count=1`

Expected: PASS; manifest order/path/hash/tracked-tree/second-amendment tamper, caller digest override, tuple/combined mismatch, same-binary/different-image, same-image/different-binary, cross-artifact clean supply-chain record, wrong subject, wrong/self-included record digest, wrong obligation-document digest, bundle/reference/digest mismatch, missing/same/inside-repo/symlink output or build-root path, stale vulnerability DB, SBOM missing/extra dependency, missing/duplicate operations binary/image/runner-helper/runner-launcher or launcher-SBOM/provenance/integration layout, mismatched embedded binary, binary-only auxiliary digest, catalog missing/extra/duplicate/role-swap/untracked/symlink/path override, bad digest, ruleset mismatch and unparseable tool output each return nonzero with neither output retained. Both fixed build-input loaders, the complete staged cleanup-capsule/tombstone crash matrix, the atomic `lock-outer-images` rewrite matrix and locked CLI surface pass. Two roots and four distinct run/boot identities produce byte-identical canonical receipts and bundles；mutating an actual runtime process receipt does not alter either, but a scope missing a required role, adding/swapping a role or using another scope's valid process tuple fails later scope validation.

- [ ] **Step 9: REFACTOR — fuzz OCI metadata and run static gates**

Run: `go test ./internal/artifactscan -run '^$' -fuzz '^FuzzOCIDescriptor$' -fuzztime=10s -timeout 30s`

Run: `go vet ./internal/artifactscan ./cmd/talenro-artifact-scan ./cmd/talenro-c12-runner-helper`

Expected: PASS with no panic or unbounded allocation.

- [ ] **Step 10: Commit scanner independently**

```bash
git add internal/artifactscan/rules.go internal/artifactscan/buildinputs.go internal/artifactscan/buildinputs_test.go internal/artifactscan/godeps.go internal/artifactscan/executable.go internal/artifactscan/oci.go internal/artifactscan/scan.go internal/artifactscan/scan_test.go cmd/talenro-artifact-scan/main.go cmd/talenro-artifact-scan/main_test.go cmd/talenro-c12-runner-helper/main.go cmd/talenro-c12-runner-helper/main_test.go testdata/c12/scanner/renamed-core.json testdata/c12/scanner/archive-core.json testdata/c12/scanner/base-layer-core.json testdata/c12/scanner/forged-module.json testdata/c12/scanner/artifact-catalog-valid.v1.json testdata/c12/scanner/outer-image-build-context-valid.v1.json testdata/c12/scanner/operations-build-policy-valid.v1.json
git commit -m "feat: add closed-set C1.2 artifact scanner"
```

### Task 4: Digest-pinned outer Compose and isolated inner daemon

**Files:**
- Create: `deploy/c12/compose.outer.yaml`
- Create: `deploy/c12/context-init.Dockerfile`
- Create: `deploy/c12/inner-daemon.Dockerfile`
- Create: `deploy/c12/verifier.Dockerfile`
- Create: `deploy/c12/context-init-entrypoint.sh`
- Create: `deploy/c12/inner-daemon-entrypoint.sh`
- Create: `deploy/c12/verifier-entrypoint.sh`
- Create: `deploy/c12/context-init.build-context.v1.json`
- Create: `deploy/c12/inner-daemon.build-context.v1.json`
- Create: `deploy/c12/verifier.build-context.v1.json`
- Create: `testdata/c12/integration-manifest.v1.json`
- Create: `testdata/c12/toolchain-lock.v1.json`
- Create: `internal/c12evidence/compose_test.go`
- Create (first line `//go:build c12_docker`): `internal/c12evidence/compose_c12_docker_test.go`
- Modify: `internal/c12evidence/runnerhelper.go`
- Modify: `internal/c12evidence/runnerhelper_test.go`
- Modify: `cmd/talenro-c12-runner-helper/main.go`
- Modify: `cmd/talenro-c12-runner-helper/main_test.go`
- Modify: `deploy/c12/compose.outer.yaml`
- Modify: `testdata/c12/integration-manifest.v1.json`
- Modify: `testdata/c12/toolchain-lock.v1.json`
- Modify: `testdata/c12/windows/runner-profile.v1.json`
- Modify: `testdata/c12/diagnostic/runner-profile.v1.json`
- Modify: `internal/c12evidence/compose_test.go`
- Test: `internal/c12evidence/compose_test.go`
- Test: `internal/c12evidence/compose_c12_docker_test.go`

Both Compose test files use the external package declaration `package c12evidence_test`. They may import `internal/artifactscan` for the cross-package loader/command contract while production dependency remains `artifactscan -> c12evidence`; no in-package c12evidence test may create the reverse test import cycle.

**Interfaces:**
- Consumes: Task 3's fixed `C12OuterImageBuildContextV1` loader/digest, scanner ruleset/version, B08/B09 OCI layouts and locks, PostgreSQL 18.4、Redis 8.8.1、NATS 2.14.3 exact integration layouts and license records, the Task 2 Windows and staged-diagnostic runner profiles/bootstrap/install-receipt policies, a validated run ID, and an empty `inner-data` volume.
- Produces: one audited outer Compose, three exact outer image digests, strict `C12IntegrationManifestV1`, a per-run mTLS inner Docker context, a stable verifier launcher and a verifier image test-assets inventory. This task is the sole owner of `testdata/c12/integration-manifest.v1.json` and the three tracked strict `C12OuterImageBuildContextV1` instances at `deploy/c12/context-init.build-context.v1.json`、`deploy/c12/inner-daemon.build-context.v1.json` and `deploy/c12/verifier.build-context.v1.json`. The schema is `{schema_version:"talenro-c12-outer-image-build-context/v1", image_role, ordered_members:[{repo_relative_path,mode,blob_sha256}]}` with unknown/duplicate/noncanonical paths rejected；each closed role occurs once, exact-covers only its Dockerfile、entrypoint and expressly required static assets, and rejects `compose.outer.yaml` plus integration/toolchain/catalog/profile/other rewrite outputs. `C12ToolchainLockV1.OuterImageBuildContexts[3]` stores the corresponding ordered role/manifest/digest projections. `lock-outer-images` resolves only those three fixed tracked tokens beneath its authenticated context root—there is no manifest-path override—strict-validates/exact-covers them before any build and binds each digest into image source/provenance、catalog and toolchain projections. Task 5 only consumes them. Because Tasks 6–7 still complete Windows parent source, this task's image/helper/Compose/toolchain rewrite is a provisional B10 fixture lock and cannot mint reusable scope evidence. Task 7 and Plan 09 B11 relock must revalidate but never rewrite these manifests after rebuilding the final images.

- [ ] **Step 1: RED — add a strict Compose resource/command test**

```go
func TestOuterComposeHasOnlyApprovedResources(t *testing.T) {
	model := parseComposeConfig(t, "../../deploy/c12/compose.outer.yaml")
	assertExactSet(t, maps.Keys(model.Services), "c12-context-init", "c12-inner-daemon", "c12-verifier")
	assertExactSet(t, maps.Keys(model.Networks), "c12-internal")
	assertExactSet(t, maps.Keys(model.Volumes), "inner-data", "context-secrets", "verifier-output")
	for name, service := range model.Services {
		if service.PullPolicy != "never" || !digestPinned(service.Image) || len(service.Ports) != 0 {
			t.Fatalf("outer service %s is not offline and digest pinned", name)
		}
	}
	assertContextInitStaysRunningAfterReady(t, model)
	assertExactTrackedSnapshotBind(t, model, "${C12_TRACKED_TREE_SNAPSHOT:?required}", true, false)
	assertExactTreeFactsCapsuleBind(t, model, "${C12_TREE_FACTS_CAPSULE:?required}", "/run/c12-input/tree-facts.capsule", true, false)
	assertExactGitObjectDatabaseBind(t, model, "${C12_GIT_OBJECT_DATABASE:?required}", "/run/c12-input/object-db", true, false)
	assertExactComposeEnvironmentAllowlist(t, "C12_RUN_ID", "C12_TRACKED_TREE_SNAPSHOT", "C12_TREE_FACTS_CAPSULE", "C12_GIT_OBJECT_DATABASE")
	assertContextSecretsReadOnly(t, model, "c12-inner-daemon", "c12-verifier")
	assertVerifierExportPathIsNotAMount(t, model, "/run/c12-export")
	assertContextInitPublicPathIsNotAMount(t, model, "/run/c12-public")
}

func TestOuterImageBuildContextsAreClosedAndAcyclic(t *testing.T) {
	contexts := loadFixedOuterImageBuildContexts(t, "../../deploy/c12")
	assertExactContextRoles(t, contexts, "context_init", "inner_daemon", "verifier")
	assertExactTrackedMembersAndDigests(t, contexts)
	assertNoRewriteOutputMember(t, contexts,
		"deploy/c12/compose.outer.yaml", "testdata/c12/integration-manifest.v1.json",
		"testdata/c12/toolchain-lock.v1.json")
}
```

- [ ] **Step 2: Run the Compose contract and verify RED**

Run: `go test ./internal/c12evidence -run '^TestOuterComposeHasOnlyApprovedResources$|^TestOuterImageBuildContextsAreClosedAndAcyclic$' -count=1`

Expected: FAIL because `deploy/c12/compose.outer.yaml` and the three fixed context instances do not exist.

- [ ] **Step 3: GREEN — create the exact outer topology**

Compose must set `pull_policy: never`, no host ports, one internal network, and only these mounts:

```yaml
services:
  c12-context-init:
    pull_policy: never
    networks: [c12-internal]
    volumes:
      - context-secrets:/run/c12-context
  c12-inner-daemon:
    pull_policy: never
    networks: [c12-internal]
    volumes:
      - inner-data:/var/lib/docker
      - context-secrets:/run/c12-context:ro
  c12-verifier:
    pull_policy: never
    networks: [c12-internal]
    depends_on:
      c12-context-init: {condition: service_healthy}
      c12-inner-daemon: {condition: service_healthy}
    volumes:
      - type: bind
        source: ${C12_TRACKED_TREE_SNAPSHOT:?required}
        target: /workspace
        read_only: true
        bind:
          create_host_path: false
      - type: bind
        source: ${C12_TREE_FACTS_CAPSULE:?required}
        target: /run/c12-input/tree-facts.capsule
        read_only: true
        bind:
          create_host_path: false
      - type: bind
        source: ${C12_GIT_OBJECT_DATABASE:?required}
        target: /run/c12-input/object-db
        read_only: true
        bind:
          create_host_path: false
      - context-secrets:/run/c12-context:ro
      - verifier-output:/run/c12-work
networks:
  c12-internal:
    name: talenro-c12-${C12_RUN_ID:?required}-network
    internal: true
volumes:
  inner-data:
    name: talenro-c12-${C12_RUN_ID:?required}-inner-data
  context-secrets:
    name: talenro-c12-${C12_RUN_ID:?required}-context-secrets
  verifier-output:
    name: talenro-c12-${C12_RUN_ID:?required}-verifier-output
```

The committed file must contain generated immutable `name@sha256:<64 lowercase hex>` image references, not tags or interpolation. The network and all three volumes have the exact explicit stable names shown above and Compose ownership labels；resource names and ownership labels interpolate only the validated `C12_RUN_ID`，so helper WAL can inspect/recover those exact names without any list. The only other Compose interpolation inputs are helper-issued `C12_TRACKED_TREE_SNAPSHOT`, `C12_TREE_FACTS_CAPSULE` and `C12_GIT_OBJECT_DATABASE`, installed by the native parent only in the locked Docker child environment and removed immediately afterward; none is inherited by an ordinary shell. The capsule is a bounded public signed one-run native-bridge document binding outer snapshot ownership/file identity、minimal object-database identity/closure、typed Git locator、canonical tree digest、capture kind and WAL digest；it contains no secret/key and cannot name another snapshot/database. All three binds use long syntax with `read_only: true` and `bind.create_host_path: false`；the context-secrets mount remains read-only. `/run/c12-export` is a fixed directory in the verifier container writable layer and is not a bind/tmpfs/named-volume mount；`verifier-output` is transient bounded workspace only and no receipt/bundle/result is copied from it.

- [ ] **Step 4: GREEN — implement per-run mTLS and rootless daemon entrypoints**

The three images run under three fixed non-overlapping UID/GID identities; only context-init begins with the narrow setup capability needed to create/chown the volume tree, then permanently drops it before readiness. Context-init opens the empty volume and every child with no-follow/anti-hardlink checks, creates fixed `public`, `daemon`, `verifier-client`, and `verifier-result-signer` directories, and never leaves a shared-readable private directory. Each principal directory is mode `0700` and owned by its exact consumer UID/GID; private leaf keys are one-link regular files mode `0400`, public cert/chain/profile files are mode `0444`, and every file/parent is file+directory fsynced before use. Daemon can read only its server key plus public CA; verifier can read only client/result-signer keys plus public trust; neither can traverse/read the other's private directory, and context-init after privilege drop cannot read either consumer private key.

Context-init creates a per-run mTLS CA, daemon server leaf, verifier client leaf and verifier-result signing leaf, validates exact EKU/SAN/profile, closes and zeroes the one-shot issuer handles, no-follow unlinks and fsyncs every CA/issuer private working file before publishing readiness, and proves no private issuer key remains in the volume or writable container layer. Only the pinned public issuer chain is retained. It writes no key to stdout, atomically publishes health readiness, then remains as its unprivileged fixed UID in a signal-only bounded idle loop until Compose stops it. A successful initialization must never exit by itself, so `--abort-on-container-exit` cannot stop the verifier prematurely；an initialization, ownership, mode, hardlink/symlink or issuer-destruction failure exits nonzero before readiness and correctly aborts the project. The result signer opens its one-link `0400` leaf key once into locked memory, permits exactly one signature request, zeroes its memory immediately after signing and retains no second handle/copy；the read-only volume and its exact private leaf are destroyed by outer exact volume cleanup before the run can succeed. Inner daemon refuses a nonempty data root, binds only its internal-network TLS endpoint, requires client cert, disables registry access and content trust fallback, and writes readiness without secret material. Verifier uses only that mTLS context and has no host socket or inner storage mount.

`deploy/c12/verifier-entrypoint.sh` is frozen here as a stable, content-scanned launcher: after validating the read-only tracked tree, exact Linux helper identity, fixed `/run/c12-input/tree-facts.capsule` and fixed read-only `/run/c12-input/object-db`, it `exec`s the absolute toolchain-locked static Linux helper's public `validate-capsule-and-build-closed-set` parent mode with the four fixed paths above. That long-lived native parent verifies the TreeFactsAttestor signature/run/WAL/locator/digest and outer WAL-recorded capsule/snapshot/minimal-object-database identities, recomputes both closures, privately re-resolves the locator, and alone retains every tree/spec/receipt/bundle/core-smoke/output FD or opaque handle. It owns a closed monotonically ordered stage machine：parent-only `spec-set`、parent-only double-root `artifact-scan`、a fixed list of ordinary one-shot Bash stages、parent-only integration、parent-only real-core, then parent-only verifier-result/sign/export.

For each ordinary stage the parent resolves the authority script only from its retained authenticated snapshot handle, opens and verifies that exact file identity, derives its canonical absolute snapshot path itself, sets cwd to the locked read-only snapshot root and constructs a fixed sanitized environment. The **only** allowed process argv is `/usr/bin/bash --noprofile --norc <parent-derived-absolute-snapshot-script>`; the script path is never caller-supplied and no stage/path/digest/handle follows it. The parent writes exactly one closed stage enum to the child's parent-installed bounded stdin and accepts exactly one closed bounded result record from parent-installed stdout; no other stdin/stdout frame or inherited descriptor exists. The child cannot select a later stage or return evidence facts. All external commands spawned by Bash therefore inherit no session/tree/spec/receipt/bundle/object-database endpoint, cannot hold EOF open and cannot impersonate a peer. The parent runs spec/scan/integration/core-smoke/Docker/WAL/result/export boundaries itself and carries their private results directly into result construction；no `SCM_RIGHTS`、socket endpoint、path token or scalar digest ever crosses into Bash. Reordered/duplicate/skipped stage transitions, argv/env/cwd drift, direct internal-stage invocation as evidence, child/grandchild FD injection or a child that outlives its process group fails and triggers parent-owned exact cleanup. The entrypoint and script contain no evidence builder or duplicate WAL/JCS/hash implementation and are not modified after their lock without relock. This signed-capsule/native-helper boundary is required because a Windows HANDLE cannot cross Docker directly, while the locked Linux parent can keep native handles wholly in-process.

This task also adds strict `C12IntegrationManifestV1` decoding plus final helper subcommands `create-verifier-trust` and `run-inner-integration`. The exact locked Linux helper binary is present in both context-init and verifier images. `create-verifier-trust` is the only context-init path that generates the per-run result issuer/leaf, applies the fixed UID/GID/mode/no-follow layout, calls Task 1 `CanonicalVerifierTrustHandoff`, writes the nonmount public handoff, destroys issuer private material and publishes readiness; shell does not serialize or hash that schema. `run-inner-integration` consumes only the tracked manifest, exact Linux helper identity, inner mTLS context and WAL handle. For each closed manifest group, it serially creates one exact inner network and the PostgreSQL 18.4、Redis 8.8.1、NATS 2.14.3 dependency containers from already streamed OCI layout digests, proves `ImagePull` is denied, probes each service, sets only the child test process's `TALENRO_DATABASE_URL`、`TALENRO_REDIS_ADDRESS`、`TALENRO_NATS_URL`, runs the registered migration profile and tagged integration package set, then cleans exact child/container/network IDs in reverse order before the next group. An absent/extra manifest role, wrong image digest, failed probe/migration/test or residue fails the result and preserves both exit bits. No artifactscan or handoff schema/domain is copied into shell/c12evidence.

- [ ] **Step 5: Build, scan and lock the three outer images**

Run in the audited offline image-build runner:

```text
go run ./cmd/talenro-artifact-scan lock-outer-images --compose deploy/c12/compose.outer.yaml --context-root deploy/c12 --integration-manifest testdata/c12/integration-manifest.v1.json --output-lock testdata/c12/toolchain-lock.v1.json --windows-runner-profile testdata/c12/windows/runner-profile.v1.json --diagnostic-runner-profile testdata/c12/diagnostic/runner-profile.v1.json --rewrite-compose --rewrite-integration-manifest --rewrite-windows-runner-profile --rewrite-diagnostic-runner-profile
```

Expected: exit 0 only after each outer image/base layer, both OS-specific runner-helper builds and all three integration-only OCI layouts have a verified digest, SBOM, provenance, vulnerability result, license/notice/source obligation and secret scan; one directory-fsynced atomic rewrite transaction updates exactly Compose, integration manifest, toolchain-lock, Windows runner-profile and diagnostic runner-profile input projections. Both context-init and verifier images must contain the same exact Linux helper digest while the lock separately names the Windows PE helper digest. The command validates the candidate PE/ELF/Git/build-tool identities against tracked install/application-control policy and exact toolchain native-role projections, requires the diagnostic zero-script/zero-channel/none-provider/nonpublishing projection, and refuses a tracked absolute machine path；it does not authenticate/open a B10 bootstrap/session, consume a sealed generation or validate an install receipt before relock. It must read the three exact tracked build-context manifest tokens named above from the authenticated context root, verify their path/mode/blob digest exact cover, and bind their canonical digests into each image's provenance、catalog source role, `OuterImageBuildContexts` lock projection and both profile helper/tool projections without rewriting the manifests. Any undeclared context member、manifest/profile override、diagnostic capability expansion、rewrite-output member or post-build context drift fails. Compose consumes already built OCI layouts outside every image context, so no image-digest↔Compose cycle exists. B11's authority-operations Dockerfile lives only in separate `deploy/c12-authority/` context. This provisional lock is superseded by Task 7's same-command relock. Missing review evidence, stale input or partial rename exits 1 with all five transaction outputs byte-identical and no residue. `C12IntegrationManifestV1` strict-decodes one ordered group with exact roles `postgresql_18_4`, `redis_8_8_1`, `nats_2_14_3`, immutable OCI manifest/layout digests, closed probe and child-environment keys, registered migration profile and test package set；unknown/duplicate group, role or environment key rejects.

- [ ] **Step 6: Run offline Compose and scanner contracts and verify GREEN**

The owning Go CLI test builds the current helper PE into a newly created helper-owned repository-external root, validates that task-build recipe/hash/OS/arch/file identity, and directly starts its zero-argument `compose-config-contract` mode；Task 4 does not assume a final installed helper or expose a Go object/handle to PowerShell. The test's protected fixture bootstrap gives that child the canonical repository locator、fixed Compose token and randomized external fixture-parent capability. The native parent captures the committed fixture tree, creates one unique WAL-owned fixture root and internally materializes the legal run ID、no-follow read-only snapshot, minimal read-only object database and one-link capsule. It saves its own process copies of `C12_RUN_ID`、`C12_TRACKED_TREE_SNAPSHOT`、`C12_TREE_FACTS_CAPSULE`、`C12_GIT_OBJECT_DATABASE`, overwrites all four only in the explicitly constructed locked-Docker child environment, invokes the absolute locked Docker executable, checks `compose config --quiet` exit/normalized binds, and restores/clears its environment plus exact-cleans the fixture in a native `defer` before exit. Neither the caller nor an ordinary shell receives a fixture/object-database path from ambient state. Missing fixture facts, any ambient pre-set value observed instead of overwritten, run-ID/path/file-identity/object-closure splice, symlink/reparse/hardlink, `create_host_path` fallback, wrong helper/Docker identity, config failure, restore failure or fixture cleanup failure is RED. Direct/private invocation and a caller repo/Compose/fixture/run/snapshot/capsule/object-database flag are absent from help.

Run: `go test ./internal/c12evidence -run '^TestOuterCompose|^TestOuterImageBuildContextsAreClosedAndAcyclic$' -count=1`

Run: `go test ./cmd/talenro-c12-runner-helper -run '^TestComposeConfigContractParent|^Test' -count=1`

Run: `go run ./cmd/talenro-artifact-scan scan-oci --lock testdata/c12/toolchain-lock.v1.json --compose deploy/c12/compose.outer.yaml`

Expected: all PASS; no service has a host port/extra mount, context-init remains healthy until verifier termination, each service runs under its fixed distinct UID/GID and can read only its own private-key directory/public trust, CA/issuer private material is destroyed before readiness, wrong mode/owner/hardlink/symlink fails, `/run/c12-export` is not a mount, all outer/integration images resolve to locked IDs, and verifier core assets are fully inventoried but absent from context-init/daemon and all production inputs.

- [ ] **Step 7: REFACTOR — prove inner registry pull is unavailable**

Run within the verifier contract fixture: `go test -tags=c12_docker ./internal/c12evidence -run '^TestInnerDaemonRejectsImagePull$' -count=1 -timeout 2m`

Run: `go vet ./internal/c12evidence ./cmd/talenro-c12-runner-helper`

Run on Windows before that tagged command:

```powershell
if ((Get-Content -LiteralPath 'internal/c12evidence/compose_c12_docker_test.go' -TotalCount 1) -cne '//go:build c12_docker') { throw 'compose_c12_docker_test.go must start with //go:build c12_docker' }
```

Expected: PASS only when the owning file has the exact first-line tag, streaming `ImageLoad` succeeds for every locked application/integration layout, `ImagePull` returns the fixed denied category, the integration manifest rejects role/digest/group drift, and the inner network cannot resolve or connect to a registry.

- [ ] **Step 8: Commit the outer substrate**

```bash
git add deploy/c12/compose.outer.yaml deploy/c12/context-init.Dockerfile deploy/c12/inner-daemon.Dockerfile deploy/c12/verifier.Dockerfile deploy/c12/context-init-entrypoint.sh deploy/c12/inner-daemon-entrypoint.sh deploy/c12/verifier-entrypoint.sh deploy/c12/context-init.build-context.v1.json deploy/c12/inner-daemon.build-context.v1.json deploy/c12/verifier.build-context.v1.json testdata/c12/integration-manifest.v1.json testdata/c12/toolchain-lock.v1.json testdata/c12/windows/runner-profile.v1.json testdata/c12/diagnostic/runner-profile.v1.json internal/c12evidence/compose_test.go internal/c12evidence/compose_c12_docker_test.go internal/c12evidence/runnerhelper.go internal/c12evidence/runnerhelper_test.go cmd/talenro-c12-runner-helper/main.go cmd/talenro-c12-runner-helper/main_test.go
git commit -m "build: add isolated C1.2 verifier substrate"
```

### Task 5: Canonical inner Bash authority script

**Files:**
- Create: `scripts/c12-authority.sh`
- Create: `internal/c12evidence/authority_script_test.go`
- Create (first line `//go:build c12_docker`): `internal/c12evidence/authority_script_c12_docker_test.go`
- Modify: `internal/c12evidence/runnerhelper.go`
- Modify: `internal/c12evidence/runnerhelper_test.go`
- Modify: `cmd/talenro-c12-runner-helper/main.go`
- Modify: `cmd/talenro-c12-runner-helper/main_test.go`
- Create (first line `//go:build linux`): `cmd/talenro-c12-runner-helper/runnerhelper_linux.go`
- Create (first line `//go:build linux`): `cmd/talenro-c12-runner-helper/runnerhelper_linux_test.go`
- Create (first line `//go:build windows`): `cmd/talenro-c12-runner-helper/runnerhelper_windows.go`
- Create (first line `//go:build windows`): `cmd/talenro-c12-runner-helper/runnerhelper_windows_test.go`
- Test: `internal/c12evidence/authority_script_test.go`
- Test: `internal/c12evidence/authority_script_c12_docker_test.go`
- Test: `internal/c12evidence/runnerhelper_test.go`
- Test: `cmd/talenro-c12-runner-helper/main_test.go`

**Interfaces:**
- Consumes: exactly one parent-selected closed ordinary-stage enum on bounded stdin, the parent-derived authenticated absolute `scripts/c12-authority.sh` identity, cwd fixed to the read-only tracked snapshot and a closed sanitized environment. It receives no Docker/mTLS context, Git/object-database/tree/spec/scanner/catalog/integration/core-smoke/WAL/result/export path, protected handle, canonical bytes, digest, signer or deleter authority. During B10 its ordinary command graph is exercised with fixed fake tools；the first real authority-operations build/scan and reusable scope outputs occur only in the native parent from the final B11 tree in Plan 09 Task 8.
- Produces: exactly one bounded `C12OrdinaryStageResultV1{stage,status,finite_category}` record on stdout and an exit status for that one invocation. It never constructs a receipt, bundle, `C12VerifierResultV1`, `C12ScopeEvidenceV1`, WAL record, Docker resource or export file. Task 4's long-lived native parent alone owns those boundaries and consumes this result as one finite portability/test-stage fact.

- [ ] **Step 1: RED — add exact stage-order and fail-closed tests**

```go
func TestAuthorityScriptRunsEveryRequiredStageInOrder(t *testing.T) {
	transcript := runNativeVerifierParentWithFakeOrdinaryStages(t)
	assertOrdered(t, transcript,
		"parent:spec-set", "parent:artifact-scan",
		"shell:check-tools", "shell:generated-diff", "shell:unit", "shell:fuzz", "shell:race", "shell:vet", "shell:lint",
		"shell:migration-roundtrip", "parent:integration", "shell:mtls-contract", "shell:controlled-process",
		"shell:xray-smoke", "shell:sing-box-smoke", "shell:load-1000", "shell:privacy",
		"parent:real-core", "parent:inner-ownership-cleanup", "parent:tracked-snapshot-revalidation", "parent:verifier-result")
	assertEveryShellLaunch(t, transcript,
		[]string{"/usr/bin/bash", "--noprofile", "--norc", parentDerivedSnapshotScript(t)},
		"one enum on stdin", "one result on stdout")
}
```

Add `TestNativeVerifierParentCoreSmokeFD3AdoptionSequence`, `TestNativeVerifierParentRejectsCoreSmokeHandoffBypass` and `TestRunnerHelperContainerVerifierIngressDoesNotRequireMachineLauncher` in the helper command tests, plus `TestRunnerHelperCoreSmokeFD3Only` and `TestWindowsOuterParentCannotReachCoreSmokeSlotOrTrustedSource` in `internal/c12evidence/runnerhelper_test.go`. `main.go` calls only a private OS facade whose signature contains no Plan 07 slot/source/set/result types. The unique production FD3 constructors/adoption imports and call sites live in `runnerhelper_linux.go`；its same-tag test `TestRunnerHelperLinuxOwnsSoleFD3AdoptionCallSite` proves the fixed container ingress is the only route. `runnerhelper_windows.go` is a fail-closed facade with no Plan 07 FD3 import；`TestRunnerHelperWindowsContainerIngressFacadeRejectsBeforeFD3Types` and the static import matrix prove both the Windows machine-base launcher and Windows versioned outer helper cannot import or reference the slot/source/set types, constructors or semantic result validator. Common sources may dispatch an OS facade but cannot mention those types or methods. The positive Linux fixture must enter only through the fixed container entrypoint/capsule/object-database contract and use the exact Plan 07 calls and order `NewCoreSmokeResultParentSession` → place only the returned child duplicate at literal FD 3 in the exact direct child's inheritance table → close the parent copy of that child duplicate → `BindStartedChild` → wait and reap the exact child → `AdoptAfterChildExit` → `AdoptCoreSmokeResultFromProtectedSlot`; only the last call may mint `TrustedCoreSmokeReceiptSource` and return the parent-revalidated result/set. RED rows must reject bytes/path/digest/FD-number/reopened-FD/second-memfd/socket/pipe/`SCM_RIGHTS`/stdout/env handoff, adoption before reap, missing bind, wrong child/PID/start/executable/session, missing final seals, provisional or old set and a second adoption/source use. The launcher may receive only role/mode/install-record launch binding plus bounded child exit status and can never receive a `C12VerifierResultV1` body.

- [ ] **Step 2: Run authority-script contracts and verify RED**

Run: `go test ./internal/c12evidence ./cmd/talenro-c12-runner-helper ./cmd/talenro-c12-runner-launcher -run '^TestAuthorityScript|^TestNativeVerifierParentCoreSmoke|^TestRunnerHelperContainerVerifierIngressDoesNotRequireMachineLauncher$|^TestRunnerHelperCoreSmokeFD3Only$|^TestWindowsOuterParentCannotReachCoreSmokeSlotOrTrustedSource$' -count=1`

Run on Linux: `go test ./cmd/talenro-c12-runner-helper -run '^TestRunnerHelperLinuxOwnsSoleFD3AdoptionCallSite$' -count=1`

Run on Windows: `go test ./cmd/talenro-c12-runner-helper ./cmd/talenro-c12-runner-launcher -run '^TestRunnerHelperWindowsContainerIngressFacadeRejectsBeforeFD3Types$|^TestWindowsOuterParentCannotReachCoreSmokeSlotOrTrustedSource$' -count=1`

Expected: FAIL because `scripts/c12-authority.sh` is missing and the nested Linux parent has not yet frozen the exact protected FD 3 adoption sequence or Windows-outer compile-time exclusion.

- [ ] **Step 3: GREEN — implement bounded stage execution**

The script uses `set -euo pipefail` and is never itself an operator command, stage scheduler, timeout owner or cleanup trap. For each invocation the long-lived native parent verifies `/usr/bin/bash` and the script file by locked identity, derives the script's absolute path from its retained snapshot handle, sets cwd to that read-only snapshot (which contains no ambient `.git`), builds the fixed sanitized environment, and launches exactly `/usr/bin/bash --noprofile --norc <parent-derived-absolute-snapshot-script>`. No caller path or fixed-stage argv exists. The script strict-reads one closed enum from parent-installed bounded stdin, runs only that ordinary stage, writes one closed bounded result to parent-installed stdout and exits; unknown/trailing/duplicate stdin, extra stdout, stderr beyond the finite category cap or an outliving process group rejects. It receives no tree/spec/receipt/bundle/object-database/control FD、locator、tracked-tree digest、`SpecDigest`、manifest/member/hash or stage environment input, never invokes Docker、`spec-digest`、scanner、runner helper、WAL、signer、deleter, and cannot request or reorder stages. The parent—not Bash—runs all spec/artifact/integration/core/result/export/cleanup boundaries and preserves `ExitBits`.

Unit/fuzz/race/vet/lint/migration ordinary commands read only the locked cwd and use only fixed parent-prepared bounded work/cache locations from the image policy; they never receive run/output authority and cannot delete those locations. For generated-diff, every generator must either expose a frozen tested no-write `--check` mode or run in the native parent's private generation-check copy；the parent compares the exact declared generated-file closure byte-for-byte and mode-for-mode to the snapshot, rejects missing/extra/undeclared output, and exact-cleans the copy. No shell command writes the snapshot, relies on `.git`, opens `/run/c12-input/object-db`, or falls back to an ambient repo.

Integration runs only through the locked long-lived parent's in-process `run-inner-integration` boundary and tracked `C12IntegrationManifestV1`: it serially creates the exact inner network/PostgreSQL 18.4/Redis 8.8.1/NATS 2.14.3 group from already present digest-verified streaming layouts, probes all three, runs the registered migration and child test packages with only the manifest-allowed environment, and exact-cleans before returning；any `ImagePull`, registry lookup or dependency download is the fixed denied category. Tests require the parent to execute exactly the declared state order, retain the parent-owned spec result before artifact-scan/verifier-result, and prove reordered/duplicate/skipped/extra stage transitions or fabricated/replayed/cross-session internal result state fails before later stages. The real-core stage is likewise owned by this locked nested Linux parent and invokes only the exact catalog/hash/OS/arch-verified prebuilt `c12-core-smoke-runner`. For each adapter/run the verifier parent creates one fresh `CoreSmokeResultParentSession`, installs only its returned child duplicate at literal inherited FD 3 of that exact direct child, closes the duplicate after start, binds the started PID/start/executable, waits and reaps, then calls `AdoptAfterChildExit` followed immediately by the sole semantic bridge `AdoptCoreSmokeResultFromProtectedSlot`. Inside the core-smoke runner the test child still receives `LiveReceiptSource` only on its distinct FD 9 and returns only an opaque provisional handle to the core-smoke parent；only after that parent exact-cleans and independently proves all 11 same-run WAL resource identities absent may its same `RunnerReceiptFinalizer` call `FinalizeCoreSmokeResultToInheritedSlot`, seal FD 3 and return no result/source/path. The verifier parent can receive the full result/set only from its one-use opaque adoption and revalidates the complete set/run/WAL/cleanup binding plus `CoreSmokeBuildReceiptV1` digest、recipe-only manifest/same-committed-tree/helper identity and broker/test/test2json/test-host actual hashes. Bytes/path/digest/FD scalar/reopen/duplicate/second memfd/socket/pipe/`SCM_RIGHTS`/shell/stdout、an old full set、manifest-carried final hash、provisional、cross-run or child-only success all reject；there is no parallel “native private channel”. The Windows outer parent sees only the signed bounded `C12VerifierResultV1` exported after this validation and is structurally unable to obtain the slot、adoption、`TrustedCoreSmokeReceiptSource` or receipt set. At the scanner stage the long-lived locked Linux parent directly invokes its private `build-closed-set` child in the one frozen internal form with two independently random run-scoped repository-external roots; it reproducibly builds the three releases and, only when the final B11 source/Dockerfile/catalog/build-policy/license record are present, the static `nodecontrol-authority-operations` binary plus dedicated OCI layout, scans both roots and requires byte-identical complete receipt/bundle. All authenticated spec/receipt/bundle handles remain in the native parent；the script cannot accept or parse a spec/member/hash/catalog/path-list/artifact-digest/alternate-builder/tree fact and contains no duplicate WAL/JCS/hash implementation. Xray and sing-box remain separate fixed ordinary stages. The parent loads inner images only by streaming verified OCI layouts, records every resource intent/actual before use, and proves every `ImagePull` attempt returns the fixed denied category. Tests inject a malicious Bash grandchild that writes/holds inherited descriptors and require it to find no protected/control FD while the parent still terminates and cleans on schedule.

- [ ] **Step 4: GREEN — prove the native parent alone emits three bounded fixed-path export files**

After all ordinary children have exited and their bounded results are validated, the **native parent** sets `C12VerifierResultV1.SpecDigest` only from the Task 1 builder result；sets `AuthorityOperationsArtifact`, both authority-operations digests, `SupplyChainEvidenceBundleDigest`, `ArtifactCatalogDigest`, process policy, image set and `ArtifactScanDigest` only from `CanonicalScanReceipt` plus the required clean Task 3 bundle；and requires that receipt/bundle repo/tree/toolchain/ruleset, embedded operations binary/image tuple, closed supply-chain record set, catalog, runner-helper/integration roles and every other release/image/asset/expected-process-policy field match the current closed set. Actual inner process receipts enter `VerifierTestResultDigest`/`InnerProcessReceiptSetDigest` and must exact-cover the Windows container-deterministic policy projection；the canonical full same-run `ValidatedCoreSmokeReceiptSetV1` digest、its separately validated `CoreSmokeBuildReceiptV1` digest and its final cleanup/11-resource-absence digest are mandatory distinct ordered members of that test-result preimage. A legacy/old-full-set result、manifest final-output hash、build-receipt/tree/helper/recipe splice、provisional child handle or digest-only path cannot enter it. These run facts do not enter or mutate the canonical scan receipt/bundle. The parent revalidates all against the same tracked snapshot/artifacts immediately before signing the strict result, signs exactly once and zeros the in-memory result signing key handle as frozen in Task 4, then exports. Bash has already exited and has no result/export handle.

The parent creates `/run/c12-export` as an empty verifier-container-layer directory proven by mountinfo/device+inode checks not to be any bind/tmpfs/named-volume mount；opens it by directory FD with no-follow, creates three random same-directory `O_CREAT|O_EXCL` one-link regular files with fixed modes, writes/caps/fsyncs, re-opens and hashes, then atomically `renameat`s exactly `/run/c12-export/closed-set-receipt.json`, `/run/c12-export/supply-chain-evidence-bundle.json`, and `/run/c12-export/verifier-result.json` and fsyncs the directory. Preexisting destination, source/directory identity change, mount appearance or extra fourth entry fails. The receipt/result are at most 64 KiB and the bundle at most 4 MiB；all three reject symlink/hardlink/non-regular targets and contain no private key, credential, file content, secret match, config, core output, absolute path or node ID. The sole permitted certificate bytes are the bounded public leaf/issuer chain inside `VerifierAttestation`; receipt and bundle contain none, and unknown/additional/self-signed trust material rejects. No scope or reusable evidence is written to `verifier-output`, stdout, stderr or a named volume. The script can generate/sign/export nothing, including upgrade/staging/recovery/Down evidence; its only output remains the finite stage result.

- [ ] **Step 5: Run script contract and injected-failure matrix and verify GREEN**

Run: `go test ./internal/c12evidence ./cmd/talenro-c12-runner-helper ./cmd/talenro-c12-runner-launcher -run '^TestAuthorityScript|^TestNativeVerifierParentCoreSmoke|^TestRunnerHelperContainerVerifierIngressDoesNotRequireMachineLauncher$|^TestRunnerHelperCoreSmokeFD3Only$|^TestWindowsOuterParentCannotReachCoreSmokeSlotOrTrustedSource$' -count=1`

Run the same-tag Linux and Windows ownership commands from Step 2 on their respective hosts.

Expected: PASS; each injected stage failure stops later test stages, cleanup always runs, primary and cleanup bits are preserved, missing Xray/sing-box/integration dependency result fails, provisional/old-path/cross-run core-smoke input、every FD3/reap/adoption ordering or identity bypass and missing any of the 11 final absence proofs fail, the Windows outer parent has no compile-time/runtime route to the protected slot/source/set, context-init readiness cannot terminate verifier, and no canary/result body appears in captured output.

- [ ] **Step 6: REFACTOR — run shell syntax and Docker-tagged smoke contract**

Run: `bash -n scripts/c12-authority.sh deploy/c12/verifier-entrypoint.sh`

Run on Windows:

```powershell
if ((Get-Content -LiteralPath 'internal/c12evidence/authority_script_c12_docker_test.go' -TotalCount 1) -cne '//go:build c12_docker') { throw 'authority_script_c12_docker_test.go must start with //go:build c12_docker' }
```

Run: `go test -tags=c12_docker ./internal/c12evidence -run '^TestAuthorityScriptAgainstInnerDaemon$' -count=1 -timeout 15m`

Run: `go vet ./internal/c12evidence ./cmd/talenro-c12-runner-helper`

Expected: PASS with the exact first-line owning tag, serial integration manifest lifecycle, exact P07 FD3 adoption call graph, three safe export files, exact inner resource cleanup and unchanged repo baseline.

- [ ] **Step 7: Commit canonical authority logic**

```bash
git add scripts/c12-authority.sh internal/c12evidence/authority_script_test.go internal/c12evidence/authority_script_c12_docker_test.go internal/c12evidence/runnerhelper.go internal/c12evidence/runnerhelper_test.go cmd/talenro-c12-runner-helper/main.go cmd/talenro-c12-runner-helper/main_test.go cmd/talenro-c12-runner-helper/runnerhelper_linux.go cmd/talenro-c12-runner-helper/runnerhelper_linux_test.go cmd/talenro-c12-runner-helper/runnerhelper_windows.go cmd/talenro-c12-runner-helper/runnerhelper_windows_test.go
git commit -m "test: add canonical C1.2 Docker authority"
```

### Task 6: Locked PowerShell wrapper and Windows Job Object scope

**Files:**
- Create: `internal/c12runnerprofile/windows_cleanup.go`
- Create: `internal/c12runnerprofile/windows_cleanup_test.go`
- Create: `scripts/verify-c12.ps1`
- Create: `internal/c12evidence/powershell_wrapper_test.go`
- Create: `internal/c12evidence/wrapper_contract_test.go`
- Modify: `internal/c12evidence/runnerhelper.go`
- Modify: `internal/c12evidence/runnerhelper_test.go`
- Modify: `internal/c12evidence/cleanup_test.go`
- Modify: `cmd/talenro-c12-runner-helper/main.go`
- Modify: `cmd/talenro-c12-runner-helper/main_test.go`
- Test: `internal/c12evidence/powershell_wrapper_test.go`

**Interfaces:**
- Consumes: parent-supplied finite nonsecret portability facts and the exact tracked PowerShell script identity. The shell consumes no repo/tool/output path、Docker/WAL/tree/attestor handle or canonical evidence bytes.
- Produces: only one bounded PowerShell/Job-Object portability stage result on the child process's parent-installed standard-output handle. The handle number is never present in argv/environment：the native parent installs its anonymous pipe write end as stdout at spawn, binds the reader to the exact child PID/start token/stage transcript, caps one closed result record, then kills/waits the whole Job before accepting EOF. Task 7's sole `run-windows-final-scopes` parent consumes it and alone produces the production-capable `windows_powershell_docker` envelope. B10 exercises the same parent with non-authoritative fixtures；Plan 09 Task 8 performs the first real final-tree run.

- [ ] **Step 1: RED — add PowerShell path, command allowlist and cleanup tests**

The fake transcript belongs to the task-built `run-windows-final-scopes` native parent. It requires the exact Docker order：context inspect/version/info、three image inspect、config、seven fsynced intents、up、three `compose ps --all --quiet` service calls、three exact-ID plus four exact-name inspections/actuals、four fixed copies、post-copy re-inspection、down and seven not-found records. The PowerShell child transcript is closed to one exact parent launch, Job-list child/grandchild checks and one bounded result frame；it must contain zero Docker/helper/Git/scanner/signer/deleter invocation and no output/path/handle/canonical bytes. Parent tests inject every Docker/identity/tree/spec/result/handoff/capsule/cleanup/publication failure, PATH alias, wrong shell/script identity, child escape, fake stage frame, account/TEMP/hard-coded path and crash seam. Every preflight failure precedes Docker/signing/output；every later failure exact-cleans the parent-owned identities and leaves the protected channel unchanged.

Task 6 defines exactly three wrapper roots：`TestPowerShellWrapperUsesLockedSnapshotIdentityAndNoAmbientTools` and `TestPowerShellWrapperFullLifecycleCleanupAndZeroPrematurePublication` in `internal/c12evidence/powershell_wrapper_test.go`, plus `TestRunnerHelperWindowsFinalScopesClosedCommandSurface` in `cmd/talenro-c12-runner-helper/main_test.go`. Their literal subcase tables exact-cover wrong snapshot/script/tool identity、PATH/drive/account/TEMP substitution、closed stage/stdio framing、Job descendant escape、every parent Docker/WAL/cleanup seam and zero sign/channel side effect before the post-clean successor. No other top-level `TestPowerShellWrapper*` or `TestRunnerHelperWindowsFinalScopes*` definition is permitted；Task 7's B10 meta and Windows wrapper manifest parse both owner files independent of host selection and make a missing、renamed、duplicate、extra or skipped root fail closed.

Task 6 also edits the unique Task 2 `TestProtectedWALKeyBootstrapRawRunnerprofileCallsOwnedOnlyByCleanupFacade` literal table from four to five cleanup Bind/Open owners by adding exactly `OpenFixedWindowsCleanupCapability`, and no other raw owner. The test must fail RED until `windows_cleanup.go` implements the same guarded one-argument current-state/adapter-bind path.

Capsule-specific wrapper tests require one helper create/validate/clean lifecycle and inject every capsule path/file/signature/nonce/run/snapshot/object-database/WAL/expiry splice plus attestor-domain substitution；all fail before producer/signing and leave owned objects absent. No scope signature is requested until both Windows Docker lifecycles finish, unsigned facts compare, and cleanup follows one sequence：declared identities absent → terminal WAL file+directory fsync → leaf Prepare → tombstone stage/file+directory fsync + first `windows_cleanup_finalization_pending(receipt,target)` CAS/reread → leaf Recover/Commit → exact key absence → tombstone unlink+parent-directory fsync+exact absence while pending → final CAS/reread to same-generation `windows_post_cleanup_publication`. The target has no later tombstone work. No scope/channel attestation or channel transaction may precede that successor；terminal-inactive/next remains unavailable until both scope self-verifications and committed-sender inspection.

- [ ] **Step 2: Run PowerShell wrapper tests and verify RED**

Run on Windows: `go test ./internal/c12evidence -run '^TestPowerShellWrapper' -count=1`

Expected: FAIL because `scripts/verify-c12.ps1` does not exist.

- [ ] **Step 3: GREEN — implement locked preflight and run ownership**

The native parent resolves and locks PowerShell、Docker/Compose、Git、scanner and helper identities and validates the runner profile plus native-executable install receipts. Before tree capture or any WAL/run/snapshot/object-database/output create, the unexported Windows role facade generates every stable reservation, derives the sealed `CleanupOnlyOwnershipPlan` from those values plus the authenticated `windows_scopes` session, and executes `BeginProtectedWALKeyBootstrap(ctx,session,plan) → CreateOrRecoverProtectedWALKeyBootstrap → SealAndSelectProtectedWALKeyBootstrap({IssuedAt,ExpiresAt}) → OpenSelectedOwnershipWAL(ctx,selected)`. The composite alone fills machine/run/slot/provider/parent/resource projections, selects the intent before the one DPAPI/TPM-protected key, seals+selects the cleanup capsule and consumes selected ownership while exact-creating/reopening the WAL and file/directory-fsyncing its bootstrap+observed identity. It returns only sealed `SelectedOwnershipWAL`, never the DPAPI key/HMAC、binding、intent、path or raw handle. Only after that return may the parent append through the sealed WAL, capture the committed tree, derive/authenticate `scripts/verify-c12.ps1` from its committed blob/mode and held snapshot-root/no-follow file identities, and materialize the snapshot/minimal object database/tree-facts capsule while retaining `TreeFactsAttestor`、two distinct scope-role `RunnerAttestor` handles、one channel-sender attestor and sealed run/output capabilities. Each attestor is non-exportable、one-use and domain/EKU-bound；neither scope attestor can sign the other scope or a channel transaction, and the channel attestor cannot sign scope bytes. It directly launches only that parent-derived absolute snapshot script under the receipt-pinned PowerShell interpreter after those facts are sealed；no install receipt covers the script. PowerShell validates the parent-provided finite portability projection and runs the Job test, but receives none of those handles or paths and cannot invoke Git/Docker/helper. The parent consumes the bounded stdout record and continues only when child PID/start token、script/executable identity and expected portability digest match. Every zero-argument restart maps the shared `abort_unstarted|abort_intent|role_dispatch` bootstrap enum before the Windows phase router；either abort exact-terminalizes its own pre-resource state and cannot fall through.

- [ ] **Step 4: GREEN — run native Job Object portability before Docker**

The native parent invokes the B06 fixture with `PROC_THREAD_ATTRIBUTE_JOB_LIST` and `EXTENDED_STARTUPINFO_PRESENT`, and gives PowerShell only the fixed test role. Failure to use the attribute、wrong child identity、escaped grandchild or residue fails the stage. The result remains Windows portability evidence only.

- [ ] **Step 5: GREEN — invoke only the outer Compose and canonical verifier**

The native parent—not PowerShell—executes the exact Docker lifecycle frozen above, including config, seven resources, all `ps --all`/inspect/copy/re-inspect/down/not-found operations and strict result/handoff validation. It creates every destination under its own absent identity and never passes a deletion/output path to shell.

The parent completes the first exact Docker lifecycle, revalidates original repository closure, verifies its terminal inner WAL/spec/result facts, combines the validated PowerShell stage digest with runner/daemon/Job/capsule/object-database facts, and retains only canonical receipt/bundle plus **unsigned** PowerShell-scope facts in private locked memory. Task 6 must not request a scope signature, channel attestation or channel mutation yet. Task 7 runs the independent Git-Bash lifecycle and owns the single terminal barrier for both.

The frozen terminal order is: both Windows lifecycles complete → compare canonical receipt/bundle/unsigned facts → `CompleteFixedWindowsScopesToCleanup` (or byte-free abort) → all resources absent + terminal WAL durable → leaf Prepare → tombstone stage/fsync + select `windows_cleanup_finalization_pending(receipt,target)` → leaf Recover/Commit → key absent → tombstone unlink+parent-directory fsync+exact absence → **final** CAS/reread to same-generation `windows_post_cleanup_publication` → revalidate facts/repository → durable at-most-once PowerShell then Git-Bash signing → bind sender、consume/self-verify channel attestor → `Prepare` → `Commit` → `InspectSender` → final terminal-inactive/next CAS. No sign-before-post-clean、post-target filesystem work、ordinary terminal advance on success、second lifecycle after signing or channel before clean. Abort/cleanup failure uses the same pre-target retirement order with ordinary terminal target and never retains attestor/sender.

- [ ] **Step 6: Run PowerShell wrapper tests and verify GREEN**

Run: `go test ./internal/c12evidence ./cmd/talenro-c12-runner-helper -run '^TestPowerShellWrapper|^TestRunnerHelperWindowsFinalScopes|^TestProtectedWALKeyBootstrapRawRunnerprofileCallsOwnedOnlyByCleanupFacade$' -count=1`

Run: `go vet ./internal/c12evidence ./cmd/talenro-c12-runner-helper`

Expected: PASS; both packages and vet pass. Every forbidden Docker command, PATH alias, daemon drift, missing image, malformed evidence, worktree drift, spec-set order/path/hash/tracked-tree/second-amendment tamper and cleanup residue returns nonzero with finite sanitized output. Barrier tests pause before the first scope signature, inject each handoff/export/object-database/snapshot/work-root/run-root absence proof、terminal-WAL append/file-fsync/directory-fsync、tombstone replace/directory-fsync、key destruction/verification and leaf receipt、`windows_cleanup_finalization_pending`、post-clean CAS、tombstone unlink/final-directory-fsync seam, and prove neither scope signing nor channel activity occurred before post-clean and no final/non-final output filesystem identity ever existed；helper memory is zeroed and arbitrary-path `verify-scope` has nothing to open.

- [ ] **Step 7: REFACTOR — freeze the parent-owned PowerShell stage and validate the B10 fixture lifecycle**

Freeze the PowerShell wrapper as one internal portability/contract stage of Task 7's sole native `run-windows-final-scopes` parent, not as an operator-callable function or a Go object exposed to PowerShell. The prebuilt parent first completes the sealed plan→intent→key→cleanup-capsule→`OpenSelectedOwnershipWAL` barrier above；only then may it append the run/snapshot/output intents, capture the committed implementation snapshot, validate the tracked `scripts/verify-c12.ps1` file identity and absolute PowerShell executable against catalog/toolchain policy, materialize those exact identities and directly launch that script with a closed internal stage enum and no tree、WAL-key、attestor、receipt/bundle FD or deletion authority. At spawn the parent installs its anonymous result-pipe write end as the child's standard output；no handle number appears in argv/environment. The script emits one closed bounded record, while the parent validates PID/start token/exit, kills/waits the full Job and independently runs Docker、receipt/result/scope/signing/publication/cleanup state machines. Caller repo/script/executable/output paths, environment overrides and a public per-shell finalizer do not exist. Task 7 freezes the only exact final PowerShell-host command after both script stages are present.

At B10, run the existing PowerShell fake-tool/Job Object/cleanup contract tests with a task-scoped build of the current helper plus a fixed non-authoritative tuple/catalog/receipt fixture and assert no reusable evidence path is created. Because this task modifies helper source after Task 4's provisional lock, it must not treat the old PE/ELF/verifier-image digest as current or mint evidence；Task 7's final relock is mandatory first. In Plan 09 Task 8, the frozen command must pass all canonical verifier stages using that final lock, exact cleanup in 120 seconds and byte-equal repository baseline, and output unexpired `windows_powershell_docker` evidence with the final scanner-derived tuple/combined/record-set/catalog/full-scan bindings and `provider_class=container_deterministic`.

- [ ] **Step 8: Commit the PowerShell wrapper independently**

```bash
git add internal/c12runnerprofile/windows_cleanup.go internal/c12runnerprofile/windows_cleanup_test.go scripts/verify-c12.ps1 internal/c12evidence/powershell_wrapper_test.go internal/c12evidence/wrapper_contract_test.go internal/c12evidence/runnerhelper.go internal/c12evidence/runnerhelper_test.go internal/c12evidence/cleanup_test.go cmd/talenro-c12-runner-helper/main.go cmd/talenro-c12-runner-helper/main_test.go
git commit -m "test: add locked PowerShell C1.2 verifier"
```

### Task 7: Repository-external Git Bash wrapper, fixture comparator, and B10 exit

**Files:**
- Create: `scripts/verify-c12.sh`
- Create: `internal/c12evidence/bash_wrapper_test.go`
- Create: `internal/c12runnerprofile/channel.go`
- Create: `internal/c12runnerprofile/channel_test.go`
- Modify: `internal/c12runnerprofile/windows_cleanup.go`
- Modify: `internal/c12runnerprofile/windows_cleanup_test.go`
- Modify: `internal/c12runnerprofile/platform_resume.go`
- Modify: `internal/c12runnerprofile/platform_resume_test.go`
- Modify: `internal/c12runnerprofile/authority_cleanup.go`
- Modify: `internal/c12runnerprofile/authority_cleanup_test.go`
- Modify: `internal/c12runnerprofile/completion_state.go`
- Modify: `internal/c12runnerprofile/completion_state_test.go`
- Modify: `internal/c12runnerprofile/profile_test.go`
- Modify: `internal/c12evidence/wrapper_contract_test.go`
- Modify: `internal/c12evidence/runnerhelper.go`
- Modify: `internal/c12evidence/runnerhelper_test.go`
- Modify: `cmd/talenro-c12-runner-helper/main.go`
- Modify: `cmd/talenro-c12-runner-helper/main_test.go`
- Modify: `cmd/talenro-c12-runner-launcher/main_test.go`
- Modify: `deploy/c12/compose.outer.yaml`
- Modify: `testdata/c12/integration-manifest.v1.json`
- Modify: `testdata/c12/toolchain-lock.v1.json`
- Modify: `testdata/c12/windows/runner-profile.v1.json`
- Modify: `testdata/c12/diagnostic/runner-profile.v1.json`
- Create: `testdata/c12/launcher-import-closure.v1.json`
- Create: `testdata/c12/b10-exact-owner-map.v1.json`
- Create: `scripts/invoke-exact-b10-gates.ps1`
- Create: `scripts/invoke-exact-b10-linux-adapter-gate.sh`
- Modify: `internal/c12evidence/compose_test.go`
- Test: `internal/c12evidence/bash_wrapper_test.go`

**Interfaces:**
- Consumes: one fixed Git-Bash portability stage launched from the same native-parent-captured snapshot and final lock as the PowerShell stage；Bash receives only finite nonsecret assertions and no helper/Docker/tree/WAL/output handle.
- Produces: the sole production `run-windows-final-scopes` parent mode, its distinct `windows_git_bash_docker` scope, in-process same-input comparator and atomic four-member attested transaction. This task is the sole schema/domain/state-machine owner of generic `C12AttestedEvidenceTransactionV1`, separate opaque sender/receiver/`EvidenceAdoptionSet`/adoption-lease APIs and their exact slot/sequence semantics；B10 exercises only a `windows_scopes` sender, while `staged_diagnostic` has zero channel capability and Plan 09's reserved `platform_scope`/`authority_scope` senders and completion-only receiver reuse the same types without redefining them. Plan 09 must supply the P08-owned opaque `EvidencePublishPendingBinding` from its authenticated durable `completion_publish_pending` state；it may not recreate adoption-set persistence or ACK semantics. This task also owns the final B10 relock over both helper builds/verifier image/Compose/integration/toolchain plus Windows and diagnostic runner profile policies. B10 uses non-reusable fixtures；Plan 09 Task 8 uses the final B11 relock.

This task also freezes `testdata/c12/launcher-import-closure.v1.json` as schema `talenro-c12-launcher-import-closure/v1`. It contains the module path and exactly two ordered target records, `{GOOS:"windows",GOARCH:"amd64",GOAMD64:"v1",CGOEnabled:false,BuildTags:[],GOFLAGS:"",GOWORK:"off",GOENV:"off",GOEXPERIMENT:"",GOTOOLCHAIN:"local",Packages,ClosureDigest}` then the identical environment record for `GOOS:"linux"`. Each `Packages` value is the bytewise-sorted duplicate-free complete module-local `go list -deps` package list for `cmd/talenro-c12-runner-launcher` under that explicit environment, and each `ClosureDigest = SHA-256(ASCII("TALENRO-C12-LAUNCHER-IMPORT-CLOSURE-V1") || 0x00 || JCS(schema,module_path,goos,goarch,goamd64,cgo_enabled,build_tags,goflags,gowork,goenv,goexperiment,gotoolchain,packages))`. `TestNativeRunnerLauncherImportClosureB10Frozen` starts the locked Go tool from an allowlist-sanitized environment and invokes it once with explicit `GOOS=windows GOARCH=amd64 GOAMD64=v1 CGO_ENABLED=0 GOFLAGS= GOWORK=off GOENV=off GOEXPERIMENT= GOTOOLCHAIN=local` and once with explicit `GOOS=linux` plus the same remaining values. It passes no `-tags`, requires exact per-target list/digest equality, explicitly rejects `internal/c12evidence` and `internal/c12completionpublication` in both closures, and rejects unknown/duplicate/defaulted fields、target reorder/duplication、ambient GOOS/GOARCH/GOAMD64/CGO/GOFLAGS/GOWORK/GOENV/GOEXPERIMENT/GOTOOLCHAIN/build-tag substitution or any other ambient `GO*` build selector. The committed two-target manifest is the reviewed B10 expected value—not a test-generated golden—and may change only by repeating the B10 source/rebuild/relock review. Plan 09 reads this exact file and cannot learn an expected list from its current graph、current host or ambient tool environment.

- [ ] **Step 1: RED — add repository-external Git Bash contract tests**

The native parent creates the independently random Git-Bash cwd beneath its validated invocation root. Require exact tracked script/Bash/cygpath identities, no hard-coded repo/account/drive path and no direct Bash call to Git、PowerShell、cmd、Docker、helper、scanner、signer、deleter、`eval`、`go run` or PATH aliases. The Bash transcript contains only the one fixed stage and bounded result. Parent transcript parity covers both full Docker lifecycles after replacing scope/run/portability facts. Repeat every parent/root/tree/spec/result/handoff/export tamper and child/grandchild inherited-handle attempt；all fail with zero channel side effects and exact-cleaned parent identities.

Task 7 adds exactly three roots：`TestGitBashWrapperUsesLockedSnapshotIdentityAndNoAmbientTools` and `TestGitBashWrapperFullLifecycleParityAndTamperHasZeroChannelSideEffect` in `internal/c12evidence/bash_wrapper_test.go`, plus `TestWrapperParityExactCoversPowerShellAndGitBashStages` in `internal/c12evidence/wrapper_contract_test.go`. The first two own the fixed Bash/cygpath/script identities、no ambient lookup/forbidden child command matrix and every full-lifecycle tamper/cleanup/channel-side-effect seam；the parity root compares the two parent-owned stage schemas、Docker lifecycle projections、failure bits and cleanup outcomes while requiring their scope-specific facts to remain distinct. Together with Task 6's three names they are the complete six-entry Windows wrapper table；no other top-level test with any of the four frozen prefixes is allowed.

Task 7 also adds `TestB10ClosedHostGateExactCoversDeclaredTests` in `internal/c12runnerprofile/profile_test.go` alongside the independent owner map and the two no-argument exact-gate scripts. Its first RED proves that an absent、incomplete or mutually inconsistent map/script/owner tuple cannot pass；the Task 7 GREEN supplies and exact-compares the final 147-root contract. It is not part of any Task 2 test command or commit.

Add `TestEvidenceChannelSeparateSenderReceiverCapabilities`, `TestEvidenceChannelAdoptCommittedIsNonEnumerable`, `TestEvidenceChannelAdoptionSetIsDurableBeforeFirstAdopt`, `TestEvidenceChannelRecoverAdoptionSetClosesEveryCrashSeam`, `TestEvidenceChannelAdoptionSetBindsCompletionPublishPending`, `TestEvidenceChannelBindingMutationCrossBindsBeforeReturn`, `TestEvidenceChannelAcknowledgePersistsOuterActualBeforeReturn` and `TestEvidenceChannelCrashSeams`. Start from the four final profiles plus the nonpublishing diagnostic profile and prove only completion receives receiver capability, Windows、platform and authority may bind their one sender only from their exact success-origin post-clean successor, and diagnostic receives no channel handle at all；signed namespace/slot/sequence mutation, direct active-session bind for any role, pre-clean/post-abort bind, sender read/adopt, platform/authority inspection of predecessor nonce, diagnostic sender/receiver fabrication, receiver list/scan, expected-slot/sequence mismatch, lease/set serialization or replay, publish-pending cross-bind and ACK with anything beyond the same set's durable actual reject. `BeginAdoptionSet` must durably intent+file/directory-fsync the exact three authenticated profile bindings and use only the package-private binding proof/closed-mutation hooks to establish reciprocal outer/set digests before returning or any channel slot mutation. For every member inject before/after outer+set adoption-intent fsync、idempotent slot mutation and outer+set nonce/envelope/protected-object actual fsync；`RecoverAdoptionSet` must converge only that exact set without enumeration or duplicate consumption. ACK tests require the binding proof to show manifest/`completed` plus all adoption actuals, then inject before/after outer ACK intent、channel mutation and outer ACK actual/receipt；`Acknowledge` returns nil only after the role-state actual is durable, while `RecoverAcknowledge(set)` reads only its authenticated pending intent/actual and accepts no selector. Inject every sign/self-verify/hardware-consume/prepare/install/fsync/commit/adopt/materialize/ACK seam and require a sender-visible exact state plus a receiver recoverable atomic adoption set, never partial members or an ACK detached from P09's matching durable completion marker.

Also add `TestNativeRunnerWindowsSuccessCleanupMintsPublicationOnlyAfterAbsence`, `TestNativeRunnerWindowsAbortCleanupCannotMintPublication`, `TestNativeRunnerWindowsCleanupFailureCannotReachAttestors`, `TestNativeRunnerWindowsRecoveryDispatchIsClosedAndGuarded`, `TestNativeRunnerWindowsCleanupFinalizationReceiptCrashSeams`, `TestNativeRunnerWindowsPostCleanupAttestationResponseLossNeverResigns`, `TestNativeRunnerWindowsPostCleanupPublicationRecoversExactSenderState`, `TestNativeRunnerWindowsNextGenerationBlockedUntilCommittedPublication`, `TestWindowsRawRunnerprofileCallsOwnedOnlyByFinalScopesFacade`, `TestNativeRunnerPlatformPostCleanupPublicationRecoversExactSenderState`, `TestNativeRunnerAuthorityPostCleanupPublicationRecoversExactSenderState` and `TestNativeRunnerDirectSenderBindRejectsAllActiveSessions`. They inject every origin、closed dispatch、leaf receipt、post-clean CAS、attestation intent/result、sender bind、Prepare、Commit、Inspect and final-generation seam, prove response loss recovers only the exact transaction, and prove abort/pre-clean/direct-session paths cannot bind or sign.

Also add RED ownership/state tests `TestCompletionCoordinatorHasNoRawMutationSurface`, `TestCompletionCoordinatorRejectsCallerDefinedPermit`, `TestCompletionPublicationSinkHasNoCallerPathKindOrDigest`, `TestCompletionPublicationSinkBindsExactCoordinatorAndSet`, `TestCompletionPublicationSinkStrictlyDecodesFrozenManifestProjection`, `TestCompletionPublicationSinkRejectsViewManifestBindingMismatch`, `TestEvidenceChannelBindingSurvivesInnerCleanupOnlyCAS`, `TestEvidenceChannelPublicationProgressCrossPersistsPublishBinding`, `TestEvidenceChannelReciprocalDigestGraphIsAcyclic`, `TestEvidenceChannelSetActualDigestNeverFeedsPublishBinding`, `TestEvidenceChannelRecoverPublicationProgressClosesCrashSeams`, `TestEvidenceChannelFinalizeReadyBindsImmutableRetainedPayloadDigest`, `TestEvidenceChannelFinalizeReadyHasNoDigestCycleOrPrematureCell`, `TestEvidenceChannelDirectMutationCannotBypassCoordinator`, `TestEvidenceChannelRecoverFinalizationBeforeAndAfterGuardCAS`, `TestNativeRunnerCompletionTerminalFinalizationAuditIsPathlessReadOnly`, `TestNativeRunnerCompletionTerminalAuditSurvivesCompletedRevalidations`, `TestNativeRunnerCompletionRevalidationAbortClosesGeneration`, `TestNativeRunnerCompletionTerminalAuditSurvivesAbortedRevalidations`, `TestNativeRunnerLauncherImportClosureB10Frozen` and `TestNativeRunnerCompletionModePhaseCrossClaimRejects`. They compile the actual `internal/c12runnerprofile` coordinator plus `internal/c12evidence` consumers, prove no exported store/raw-CAS/business-byte setter、caller-visible progress value or second binding/sink implementation exists, exact-recompute the immutable set identity、every T、outer Pnew and set-actual digest while rejecting any self/new-P/full-set-digest edge, freeze the committed launcher import-closure manifest, and inject every set/outer/file/guard A/B seam including prepare at E→E1, proof-consuming terminal construction at E1→E2, E2 reservation `exiting`, old-containment absence/clear、closed completed/aborted revalidation lineage and the distinct read-only terminal-audit reservation.

The same ownership block runs `TestCompletionInnerCleanupDescriptorIsBoundedAndSealed` and `TestOpenCompletionInnerCleanupForExactCleanupIsSoleDescriptorConsumer` against the real package graph；the first exact-covers the frozen descriptor field set including nonzero launch generation and machine-seal locator, mutates each value plus the fixed completion role and requires `CleanupPolicyBindingDigest` recomputation to fail on every splice. Only c12evidence may consume the bounded sealed descriptor, and only the composite may retain the paired opaque session.

- [ ] **Step 2: Run Bash wrapper tests and verify RED**

Run on Windows: `go test ./internal/c12runnerprofile ./internal/c12evidence ./cmd/talenro-c12-runner-launcher -run '^TestB10ClosedHostGateExactCoversDeclaredTests$|^TestNativeRunnerLauncherImportClosureB10Frozen$|^TestGitBashWrapper|^TestEvidenceChannel|^TestCompletionCoordinator|^TestCompletionPublicationSink|^TestCompletionInnerCleanupDescriptor|^TestOpenCompletionInnerCleanupForExactCleanup|^TestNativeRunnerWindowsSuccessCleanupMintsPublicationOnlyAfterAbsence$|^TestNativeRunnerWindowsAbortCleanupCannotMintPublication$|^TestNativeRunnerWindowsCleanupFailureCannotReachAttestors$|^TestNativeRunnerWindowsRecoveryDispatchIsClosedAndGuarded$|^TestNativeRunnerWindowsCleanupFinalizationReceiptCrashSeams$|^TestNativeRunnerWindowsPostCleanupAttestationResponseLossNeverResigns$|^TestNativeRunnerWindowsPostCleanupPublicationRecoversExactSenderState$|^TestNativeRunnerWindowsNextGenerationBlockedUntilCommittedPublication$|^TestWindowsRawRunnerprofileCallsOwnedOnlyByFinalScopesFacade$|^TestNativeRunnerPlatformPostCleanupPublicationRecoversExactSenderState$|^TestNativeRunnerAuthorityPostCleanupPublicationRecoversExactSenderState$|^TestNativeRunnerDirectSenderBindRejectsAllActiveSessions$' -count=1`

Run the two newly closed producer/mint roots exactly: `go test ./internal/c12evidence -run '^(TestTreeFactsCapsuleStrictValidation|TestWindowsFinalScopesSealedProducerPathIsCallableAndSoleOwned)$' -count=1`

Expected: FAIL because `scripts/verify-c12.sh`、the separate channel capability types and the exact owner-map/script tuple do not exist or are not yet mutually consistent.

- [ ] **Step 3: GREEN — implement repository-external resolution and locked preflight**

Use `set -euo pipefail`; the script may use `BASH_SOURCE`/`pwd -P` only to prove its parent-supplied exact snapshot identity and fixed internal stage, never to locate or choose an ambient repo/tool. The native parent—not Bash—validates Git Bash/cygpath/native helper/Docker/Compose identities, uses locked absolute `git.exe` to capture the typed tree locator plus separate canonical digest, rejects dirty/index/untracked drift, materializes the repository-external read-only snapshot and calls the sole spec builder before any create. Bash receives no manifest/member/hash/tree-OID/tracked-tree/spec-digest/canonical/signature option or protected/control FD. It cannot invoke another helper、Docker、scanner、signer or deleter；all such operations remain stages of the same native parent.

- [ ] **Step 4: GREEN — mirror the exact outer lifecycle without copying test logic**

The native parent invokes the same `compose.outer.yaml` and verifier image for each scope with only its own tracked snapshot/capsule/minimal-object-database binds；offline normalized config must byte-match after path normalization and reject repo-relative/create-host-path fallback. It owns both complete Docker transcripts：seven fsynced intents, collision probes, exact-project `compose ps --all --quiet` for all three services, exact-ID/name inspections, actuals, four fixed copies, re-inspection, down and not-found cleanup. PowerShell and Git Bash each run only their distinct one-shot portability/assertion stage against parent-supplied finite nonsecret facts and return a bounded result digest；neither owns Docker/WAL/output paths. After each lifecycle the parent validates receipt/bundle/result/handoff and retains only canonical receipt/bundle plus unsigned scope facts and the role-bound attestors in locked memory; it signs and publishes nothing yet.

Only after **both** lifecycles finish does the parent compare canonical receipt/bundle bytes and every shared repo/tree/spec/toolchain/release/operations/catalog/image/result input. On equality its sole c12evidence facade calls raw `CompleteFixedWindowsScopesToCleanup` with exactly three fixed-order defensive slices：the bounded canonical two-scope pre-clean result、the exact scan-receipt projection and the exact supply-chain-bundle projection；any comparison/workload failure calls only the byte-free abort route. It then executes the exact leaf absence/WAL/key protocol and success receipt target frozen in Task 6, reaching same-generation `windows_post_cleanup_publication` before any signature. That successor alone revalidates the repository baseline、durably signs/self-verifies PowerShell then Git-Bash with at-most-once response-loss replay、binds the fixed sender, constructs/signs the namespace/slot/sequence-bound four-member transaction and converges `Prepare→Commit→InspectSender(transaction_nonce)`. Only `FinalizeFixedWindowsPostCleanupPublication` may then select terminal-inactive/next. Any comparison/cleanup/WAL/tombstone/key/repository failure leaves both scopes unsigned and channel unchanged；signing failure leaves channel unchanged；post-Commit failure retains only the exact immutable transaction. No sign-before-post-clean, success-to-ordinary-terminal, commit-before-clean, shell output flag, digest override, alternate script/tool path or deletion authority exists.

Freeze the command-callable sealed final boundary in `internal/c12evidence/runnerhelper.go`:

```go
package c12evidence

type WindowsUnsignedScopeFacts interface {
	windowsUnsignedScopeFacts() // existing parent-private scope-from-verifier-result output; no bytes or attestor
}

type WindowsFinalScopesUnsignedResult interface {
	windowsFinalScopesUnsignedResult() // exact two distinct scopes plus shared-fact equality proof
}

type WindowsFinalScopesValidatedResult interface {
	windowsFinalScopesValidatedResult() // session-bound unsigned result + sealed FinalClosedSetProjection
}

func ValidateAndSealWindowsPowerShellUnsignedScopeFacts(context.Context, *c12runnerprofile.AuthenticatedNativeRunnerSession, C12VerifierResultV1, VerifierResultPolicy) (WindowsUnsignedScopeFacts, error)
func ValidateAndSealWindowsGitBashUnsignedScopeFacts(context.Context, *c12runnerprofile.AuthenticatedNativeRunnerSession, C12VerifierResultV1, VerifierResultPolicy) (WindowsUnsignedScopeFacts, error)
func CompareAndSealWindowsFinalScopesUnsignedResult(context.Context, WindowsUnsignedScopeFacts, WindowsUnsignedScopeFacts) (WindowsFinalScopesUnsignedResult, error)
func BindWindowsFinalScopesValidatedResult(context.Context, *c12runnerprofile.AuthenticatedNativeRunnerSession, WindowsFinalScopesUnsignedResult, FinalClosedSetProjection) (WindowsFinalScopesValidatedResult, error)
func CompleteWindowsFinalScopesToTerminal(context.Context, WindowsFinalScopesValidatedResult) error
func AbortWindowsFinalScopesToTerminal(context.Context, *c12runnerprofile.AuthenticatedNativeRunnerSession) error
func RecoverWindowsFinalScopesToTerminal(context.Context) error
```

`ValidateAndSealWindowsPowerShellUnsignedScopeFacts` and `ValidateAndSealWindowsGitBashUnsignedScopeFacts` are the two named role-fixed parent-private verifier-result builders. Each accepts only the authenticated session plus concrete c12evidence-owned `C12VerifierResultV1`/`VerifierResultPolicy`, derives the corresponding observed runner/daemon/portability/WAL facts from that session's fixed completed stage, strict-validates every signed result and policy binding, and returns the private-concrete `WindowsUnsignedScopeFacts`; there is no scope enum、bytes、path、digest override、callback or caller implementation. The command can retain and pass a token but cannot inspect or serialize it. `CompareAndSealWindowsFinalScopesUnsignedResult` consumes exactly one PowerShell and one Git-Bash fact token, requires their scope/run distinction and all shared semantic facts equal, and returns the canonical bounded pre-clean result only behind another private marker. The fixed command then binds one `FinalClosedSetProjectionSink` to the same session, passes it once with the byte-equal typed receipt/bundle to `artifactscan.ValidateAndSealFinalClosedSetProjection`, and gives the returned sealed projection plus the unsigned-result token and same session to `BindWindowsFinalScopesValidatedResult`. That mint exact-matches all run/tree/spec/toolchain/scanner/bundle digests, consumes both inputs and the active session into its one private concrete, and exposes no canonical bytes、callback、permit、raw P08 successor or cleanup handle. `CompleteWindowsFinalScopesToTerminal` consumes only that token and internally owns raw Complete、exact cleanup、post-clean attest/send/finalize through terminal-inactive/next；`AbortWindowsFinalScopesToTerminal` is the sole pre-mint failure route and accepts no data/reason, while `RecoverWindowsFinalScopesToTerminal` accepts no selector/capability and follows only the authenticated closed dispatch. The real positive call chain is therefore `cmd fixed parent → two role-fixed semantic builders → Compare → Bind sink → artifactscan typed Validate+Seal → c12evidence semantic Bind → Complete`, with no import cycle and no production caller-defined interface.

Add `TestWindowsFinalScopesSealedProducerPathIsCallableAndSoleOwned` in `internal/c12evidence/runnerhelper_test.go`. Using `go/packages`/`go/types`, it must compile the actual command path above, require `artifactscan -> c12evidence`、command→both and c12evidence↛artifactscan, require exactly one private concrete for every sink/projection/Windows token, require `SealValidatedClosedSet` calls only in `internal/artifactscan/scan.go`, and require both role-fixed builders plus Compare/Bind/Complete/Abort/Recover calls only in the fixed `cmd/talenro-c12-runner-helper` parent. It rejects bytes/DTO/callback/permit/raw successor in the semantic Bind or terminal signatures, any command raw runnerprofile selector and any missing/dynamic/unresolved call edge. `TestWindowsRawRunnerprofileCallsOwnedOnlyByFinalScopesFacade` remains the complementary inner-owner proof：every raw Windows selector occurs only in `runnerhelper.go` and is reachable from the sealed Complete/Abort/Recover facade, never directly from the command.

The same GREEN adds the Task 7-only `TestB10ClosedHostGateExactCoversDeclaredTests` implementation、the independent 147-root JSON owner map and both no-argument exact-gate scripts. The Go meta-test hard-codes and exact-compares the normative owner table、all eight memberships and both script literals before applying its owner/no-skip analysis；the PowerShell and Bash runners retain their own literal package/name tables and fail their exact list/JSON gates. Task 2 source or tests are not retroactively changed.

Freeze the shared channel types and guarded-state coordinator extension in `internal/c12runnerprofile/channel.go`:

```go
type C12AttestedEvidenceMemberV1 struct {
	Role                          string
	Length                        uint32
	SHA256                        contracts.Digest
	ProtectedObjectIdentityDigest contracts.Digest
	CanonicalMediaType            string
}

type C12EvidenceChannelAttestationV1 struct {
	SchemaVersion  string
	SignerIdentity contracts.Digest
	SignerEKU      string
	Signature      []byte
}

type C12AttestedEvidenceTransactionV1 struct {
	SchemaVersion        string
	TransactionKind      string
	TransactionNonce     [32]byte
	ChannelNamespaceDigest contracts.Digest
	ChannelSlotIdentityDigest contracts.Digest
	SlotSequence         uint64
	SourceRunnerIdentity contracts.Digest
	RepoCommit           string
	TrackedTreeDigest    contracts.Digest
	Members              [4]C12AttestedEvidenceMemberV1
	CreatedAt            time.Time
	Attestation          C12EvidenceChannelAttestationV1
}

type EvidenceChannelSenderPhase string

const (
	EvidenceChannelSenderAbsent      EvidenceChannelSenderPhase = "absent"
	EvidenceChannelSenderPrepared    EvidenceChannelSenderPhase = "prepared"
	EvidenceChannelSenderCommitted   EvidenceChannelSenderPhase = "committed"
	EvidenceChannelSenderQuarantined EvidenceChannelSenderPhase = "quarantined"
)

type EvidenceChannelSenderState struct {
	Phase                            EvidenceChannelSenderPhase
	TransactionNonce                 [32]byte
	CanonicalEnvelopeIdentityDigest  contracts.Digest
	ProtectedMemberIdentitySetDigest contracts.Digest
}

type EvidenceChannelSender interface {
	evidenceChannelSender() // sealed bootstrap-minted sender; package-local tests only
	Prepare(C12AttestedEvidenceTransactionV1) error
	Commit([32]byte) error
	InspectSender([32]byte) (EvidenceChannelSenderState, error)
}

func BindFixedWindowsPostCleanupEvidenceSender(context.Context, WindowsPostCleanupPublicationSuccessor) (EvidenceChannelSender, error)
func FinalizeFixedWindowsPostCleanupPublication(context.Context, WindowsPostCleanupPublicationSuccessor) error
func BindFixedPlatformPostCleanupEvidenceSender(context.Context, PlatformPostCleanupPublicationSuccessor) (EvidenceChannelSender, error)
func FinalizeFixedPlatformPostCleanupPublication(context.Context, PlatformPostCleanupPublicationSuccessor) error
func BindFixedAuthorityPostCleanupEvidenceSender(context.Context, AuthorityPostCleanupPublicationSuccessor) (EvidenceChannelSender, error)
func FinalizeFixedAuthorityPostCleanupPublication(context.Context, AuthorityPostCleanupPublicationSuccessor) error

type EvidenceAdoptionLease interface {
	evidenceAdoptionLease() // opaque, one receiver session and one slot/sequence
}

type evidencePublishPendingStateProof struct {
	StateDigest              contracts.Digest
	StateEpoch               uint64
	PublishBindingDigest     contracts.Digest
	RunIdentityDigest        contracts.Digest
	ReceiverIdentityDigest   contracts.Digest
	LaunchGeneration         uint64
	Phase                    string
	ExpectedBindingSetDigest contracts.Digest
	AdoptionSetIdentityDigest contracts.Digest // immutable; excludes state/publish/set-record digests
	LastSetTransitionDigest   contracts.Digest // latest acyclic semantic transition T, zero only before set begin
	ManifestDigest           contracts.Digest
	CompletedMarkerDigest    contracts.Digest
	AdoptionActualCount      uint8
	ACKActualCount           uint8
}

type evidencePublishPendingMutation interface {
	evidencePublishPendingMutation() // closed begin/adopt/publication-progress/ACK/finalize-ready mutation minted only by channel.go
}

type completionPublicationProgress interface {
	completionPublicationProgress() // package-private closed materialize/manifest/completed intent-or-actual proof
}

type CompletionRetainedFinalizationProof interface {
	completionRetainedFinalizationProof() // opaque set/binding/post-prepare epoch/digests/exact-E2-target plus immutable-audit proof
}

type CompletionMaterializedEvidenceView interface {
	completionMaterializedEvidenceView() // opaque, immutable, exact coordinator/set/run/generation/profile/root; no path/write handle
	CanonicalMember(context.Context, CompletionEvidenceMemberRole) ([]byte, contracts.Digest, error)
	BindingDigest() contracts.Digest
}

type CompletionPublicationSink interface {
	completionPublicationSink() // opaque low-level sink bound once to one coordinator/binding/adoption set
	MaterializeAdoptedEvidence(context.Context) (CompletionMaterializedEvidenceView, error)
	StageCanonicalManifest(context.Context, CompletionMaterializedEvidenceView, []byte) error
	PublishManifest(context.Context) error
	PublishCompletedMarker(context.Context) error
	RecoverPublication(context.Context) error
}

type EvidencePublishPendingBinding interface {
	evidencePublishPendingBinding() // opaque, P09 matching durable completion_publish_pending state
	evidencePublishPendingState() (evidencePublishPendingStateProof, error)
	applyEvidencePublishPendingMutation(expectedEpoch uint64, expectedDigest contracts.Digest, expectedPublishDigest contracts.Digest, mutation evidencePublishPendingMutation) (evidencePublishPendingStateProof, error)
	commitEvidenceRetainedFinalization(CompletionRetainedFinalizationProof) error // sole proof-taking terminal guarded mutation
}

type EvidenceAdoptionSet interface {
	evidenceAdoptionSet() // opaque, one receiver/run/publish-pending generation and exact three bindings
}

type EvidenceChannelReceiver interface {
	BeginAdoptionSet(EvidencePublishPendingBinding) (EvidenceAdoptionSet, error)
	RecoverAdoptionSet(EvidencePublishPendingBinding) (EvidenceAdoptionSet, error)
	AdoptCommitted(EvidenceAdoptionSet, kind string, expectedSlot contracts.Digest, expectedSequence uint64) ([32]byte, EvidenceAdoptionLease, error)
	BindCompletionPublicationSink(EvidenceAdoptionSet, EvidencePublishPendingBinding) (CompletionPublicationSink, error)
	commitCompletionPublicationProgress(EvidenceAdoptionSet, EvidencePublishPendingBinding, completionPublicationProgress) error
	recoverCompletionPublicationProgress(EvidenceAdoptionSet, EvidencePublishPendingBinding) error
	Acknowledge(EvidenceAdoptionSet, EvidenceAdoptionLease) error
	RecoverAcknowledge(EvidenceAdoptionSet) error
	PrepareCompletionRetainedFinalization(EvidenceAdoptionSet, EvidencePublishPendingBinding) (CompletionRetainedFinalizationProof, error)
	RecoverCompletionRetainedFinalization(EvidenceAdoptionSet, EvidencePublishPendingBinding) (CompletionRetainedFinalizationProof, error)
	CommitCompletionRetainedFinalization(EvidenceAdoptionSet, EvidencePublishPendingBinding, CompletionRetainedFinalizationProof) error
}

func BindCompletionEvidenceReceiver(AuthenticatedCompletionRoleStateCoordinator) (EvidencePublishPendingBinding, EvidenceChannelReceiver, error)
```

These types and the sole concrete binding/receiver/publication-sink coordinator live in `internal/c12runnerprofile/channel.go`, in the same package as the private guarded-state engine. `BindCompletionEvidenceReceiver` is the only initial cross-package adapter：it accepts only the already authenticated coordinator in exact `publish_pending`, consumes its sealed receiver capability, exact-covers the three profile bindings and returns opaque interfaces whose unexported methods are implemented in that same package. After all three adoption actuals, `EvidenceChannelReceiver.BindCompletionPublicationSink` is the only mint for a `CompletionPublicationSink`; it binds that sink once to the same concrete coordinator、binding and adoption set and returns the same authenticated binding on recovery. The sink is deliberately lower-level than P09 semantic validation：it owns exact filesystem identities/progress but imports no `internal/c12evidence` type. Its immutable view exposes only copied bounded canonical member bytes+digests under the eight closed roles and one digest binding the exact coordinator、set、run、launch generation、profile、reserved root and member identities, never a path、write handle、lease or enumeration API.

The view's `BindingDigest V` is frozen independently as `SHA-256(ASCII("TALENRO-C12-COMPLETION-MATERIALIZED-EVIDENCE-BINDING-V1") || 0x00 || JCS(completionMaterializedEvidenceBindingV1))`. That private strict projection contains exactly `{CounterIdentityDigest,AdoptionSetIdentityDigest,MaterializationIntentSetTransitionDigest,MaterializationIntentPublishBindingDigest,RunID,LaunchGeneration,ProfileDigest,RootIdentity,EvidenceMembers}`；`EvidenceMembers` is the exact ordered eight-entry array of `{Role,ProtectedObjectIdentity,CanonicalDigest}` obtained after the intent-bound writes. It rejects unknown、missing、null、default、duplicate or reordered fields/members. The counter/set/run/generation/profile、root and every protected-object identity are the stable values already bound before materialization；the two intent values must be the exact durable materialization-intent T and its selected post-intent outer P for this set. After materialized bytes are reopened/hashed/fsynced, V is computed from only those predecessor values plus the exact-eight result；the later materialization-actual T uses V and that result in its semantic delta. Hence the order is `materialization-intent T → intent P → V → materialization-actual T → actual P → actual Sa`. V deliberately excludes every current/future epoch/full `StateDigest`、materialization-actual-or-later T/P、`PreviousStateDigest`、all `Sp/Sa/Fp/Fa` or mutable/full adoption-set record digests、manifest bytes/digest and later progress. Tests exact-recompute the domain and all projection fields, mutate each independently, and reject a view from another coordinator/set/root/intent transition or role ordering；the reciprocal-DAG test also proves no materialization-actual result can feed V backward.

Before B10, the same file freezes a private, non-exported `completionPublicationManifestProjectionV1` wire decoder used only by `StageCanonicalManifest`. It accepts a nonempty canonical manifest of at most exactly 64 KiB, strict-decodes canonical JCS with the exact future `C12CompletionManifestV1` top-level field set, including nonzero `EvidenceSetBindingDigest` and the ordered exact-eight `EvidenceMembers[{Role,CanonicalDigest}]`; rejects unknown/duplicate/null/defaulted/reordered members；and compares those nine binding values to the supplied opaque view before any temp intent/write. Nested scope/trusted-time/operations values are retained as bounded canonical raw projections here and receive no P09 semantic acceptance at this layer. This is a frozen structural sink guard, not a second exported manifest schema or semantic validator, and it imports no `internal/c12evidence` package. P09 Task 7 defines the matching public semantic struct and identical 64 KiB limit in `internal/c12evidence`, must keep its canonical wire shape byte-for-byte compatible with this already-frozen decoder, and performs all signature/time/policy semantics.

P09 Task 7's sole `internal/c12completionpublication` composite operator imports both packages, obtains this sink/view, constructs semantic inputs only from the view's exact eight members plus fixed policy/provider dependencies, passes them to `c12evidence.ValidateCompletionInputs`, and alone may call `StageCanonicalManifest` with the canonical bytes extracted from the resulting opaque validated-manifest token. This dependency direction is `c12completionpublication -> {c12evidence,c12runnerprofile}` and `c12evidence -> c12runnerprofile`, with no runnerprofile→c12evidence edge or Go import cycle. Neither that new composite package nor later c12evidence sources enter the frozen launcher's import closure, and P09 may not edit runnerprofile/launcher source to add them；B11 rebuilds the launcher byte-identically from the B10 graph. AST ownership tests reject every `StageCanonicalManifest` production call site outside that one composite operator；P09 command/runner code receives only the high-level operator, never the sink/view. No package can implement、wrap with a caller-defined concrete、mint a semantic permit/private progress value or access the underlying coordinator/store.

Schema is exactly `talenro-c12-attested-evidence-transaction/v1`; kind is closed to `windows_scopes|platform_scope|authority_scope` with one fixed four-role order per kind. Windows is receipt、bundle、PowerShell scope、Git-Bash scope；the two reserved production kinds are receipt、bundle、their outer scope、their nested envelope. Member roles are closed, length is nonzero/role-capped, SHA-256/object identity nonzero, media type exact canonical JSON, and all protected object identities distinct. `ChannelNamespaceDigest`, `ChannelSlotIdentityDigest` and nonzero `SlotSequence` are mandatory members of the signed unsigned projection; they must equal the authenticated runner session's role-bound namespace plus **next** immutable slot/sequence. The tracked profile/bootstrap never contains or predicts `TransactionNonce`, which is freshly generated only after those facts and all four protected members exist.

The strict channel attestation schema/EKU are `talenro-c12-evidence-channel-attestation/v1` and `urn:talenro:c12:evidence-channel-sender:v1`; its separate non-exportable hardware attestor signs only `TALENRO-C12-EVIDENCE-CHANNEL-TRANSACTION-V1\x00 || JCS(unsigned transaction)`. The sender must self-verify the signature/envelope and then irreversibly consume the one-shot hardware attestor **before** `Prepare`; a surviving/reusable attestor, prepare-before-self-verify or scope/capsule/result attestation in this domain rejects. Scope/capsule/result attestations reject in this domain and channel attestation cannot sign those domains.

Sender and receiver capabilities are different opaque sealed types minted only by the role bootstrap. There is no active-session sender accessor. Windows、platform and authority may bind only through `BindFixedWindowsPostCleanupEvidenceSender`、`BindFixedPlatformPostCleanupEvidenceSender` or `BindFixedAuthorityPostCleanupEvidenceSender` after their assignment-identical success-origin exact-clean successor exists；completion and staged diagnostic have no sender route. A pre-clean/abort successor、second bind、cross-role successor or any active session rejects before channel access. A sender can call only `Prepare`, `Commit` and exact nonce-bound `InspectSender` for its own namespace/slot/sequence. `EvidenceChannelSenderState.Phase` is closed to `absent|prepared|committed|quarantined`; absent requires zero nonce/envelope/member identities, prepared/committed require the queried nonce plus nonzero exact envelope/member-set identities, and quarantined returns only the authenticated known identity projection without member bytes. Inspect never returns member bytes. A receiver cannot inspect/enumerate a namespace or choose a nonce. Before its first slot mutation, completion must obtain P09's opaque `EvidencePublishPendingBinding` only after the exact run/output generation's `completion_publish_pending` role-state and its **manifest-identity reservation intent** are durable, then call `BeginAdoptionSet`. That reservation fixes only the future temp/final/marker object identities in the initial barrier；it creates no file, claims no materialization fact and is distinct from the later closed publication-progress state named `manifest_intent`. The two additional binding methods are package-private proof/closed-mutation hooks, not caller APIs：the unexported proof dynamically carries only bounded nonsecret current `StateEpoch`、full `StateDigest`、`PublishBindingDigest`、immutable `AdoptionSetIdentityDigest`、latest `LastSetTransitionDigest` and run/receiver/generation/phase/binding/manifest/marker/count projections. `channel.go` alone mints the closed mutation for ordinary begin/adopt/publication-progress/ACK transitions and a distinct source-only finalize-ready mutation；it passes the expected state proof only to the matching private hook. Every ordinary hook verifies current epoch/full digest against the guard and set record, then follows the frozen two-branch DAG recovery：from source P it applies T/delta and selects Pnew；from an already selected deterministic Pnew with identical I/T/delta/result it skips the outer CAS；an absent Sa binds the current selected epoch/full/Pnew, while a present Sa remains immutable and only the proof is reminted after a permitted execution/ordinary-inner-only successor. Neither Pnew nor outer state includes Sp、Sa or another set-record digest. Finalize-ready instead fsyncs source-only Fp, performs/rereads the outer prepare CAS to E1/D1/P1, then fsyncs source-only Fa；after launcher death the recovery hook may cross only guard-selected execution-reservation replacements with identical I/T/delta/outer business、inner `none` and P1, and reconstructs the proof from current E'/D'/P1. A stale caller view、authenticated old cell、inner drift or outer publication drift fails. `BeginAdoptionSet` authenticates the proof against its run/role/generation and exact three ordered profile bindings and executes the ordinary three steps. A crash at any step is converged only by `RecoverAdoptionSet` with that same binding、identity、T、source P and semantic delta；it never lists a namespace, selects a kind, creates a replacement set or hashes a set-actual digest back into Pnew.

Mandatory DAG crash coverage includes launcher death after ordinary outer Pnew but before Sa、after Sa、after finalize outer prepare、after Fa、before proof delivery and immediately before Commit. Recovery must bind or rebase only to the current selected epoch/full record under the two ordinary branches or terminal execution-only branch above；all other states reject.

Every `AdoptCommitted(set, kind, expectedSlot, expectedSequence)` first uses the binding proof/closed-mutation hook to derive the per-binding adoption-intent T from immutable set identity、previous T、source publish digest and exact slot/sequence delta, then performs set-pending → outer-intent Pnew → set-actual(T,new epoch/full/Pnew) in the fixed DAG order before the expected channel-slot mutation. It next derives a distinct actual T from that new source P and the returned transaction nonce、canonical envelope digest and protected-object identities, persists the same three-step outer actual transition and only then returns its opaque one-use lease. No P09 caller separately fabricates an identity、T、intent/actual/publish digest, and no set pending/actual/full-record digest enters Pnew. A crash before the matching set actual is durable returns nothing；recovery re-authenticates the guard, tolerates only byte-identical source publish business across intervening execution-reservation-only or ordinary inner-cleanup-only guarded mutations, and replays only the T-bound exact slot/sequence or actual result. It treats an already matching T/Pnew/set actual as success and quarantines an unrepresented outer-business mismatch rather than enumerating or consuming another slot. The lease gives only the completion native parent fixed-role reads for that transaction and cannot open another kind/slot/sequence or be serialized. The set exact-covers all three bindings and cannot be reused for another receiver/run/output generation or rebound to a different P09 publish-pending record.

The only legal outer filesystem-progress sequence is closed to `materialization_intent → materialization_actual → manifest_intent → manifest_actual → manifest_durable → completed_intent → completed_actual → completed_durable`. The set-bound `CompletionPublicationSink` owns the typed completion root/materializer/manifest/marker filesystem operations and is the only code that can mint the corresponding package-private `completionPublicationProgress` values. `MaterializeAdoptedEvidence` obtains the exact three already-adopted leases from its bound set, writes/reopens/hashes/fsyncs the deduplicated exact-eight members and returns the opaque immutable view only after the first two reciprocal edges are durable. `StageCanonicalManifest(view,canonical)` first requires that exact view, rejects zero or more than 64 KiB, strict-decodes the frozen low-level projection, exact-matches its set-binding digest and ordered eight roles/digests, defensively copies the exact bytes and computes their digest. **Before touching the reserved temp object**, it uses the receiver's closed mutation to cross-persist and reread a reciprocal `manifest_intent` in both the guard-selected outer role record and the same adoption-set record. Both authenticated copies carry byte-identical canonical bytes、their digest、the exact view/set binding and reserved temp/final identities；the full state digest、`PublishBindingDigest` and reciprocal set digest cover every byte. The unique state order is set pending old→new、outer guard CAS/reread、set actual/reread；outer-first is illegal. A crash during this barrier performs no filesystem write. Package-private `recoverCompletionManifestIntentBarrier` is the sole state-only convergence primitive：it receives only the already bound concrete coordinator/binding/set reservation, accepts no caller set/selector/bytes, calls the receiver's private progress hook and returns only after the three fixed steps reread equal. It is called only by `StageCanonicalManifest`, the coordinator's create `session_cleanup_recovery` branch and sink `RecoverPublication`; the coordinator branch may finish this barrier but cannot touch any filesystem fact or later progress. The sink then opens or exact-recovers only the reserved no-follow temp identity, truncates that object to zero, writes the stored bytes, reopens and hashes them, file-fsyncs, and cross-persists `manifest_actual` **without removing either byte copy**. `PublishManifest` owns the reserved sibling rename plus parent-directory fsync and only afterward cross-persists `manifest_durable`; the byte copies may be erased solely as part of that reciprocal durable transition, which retains the same digest and identities. Thus both `manifest_intent` and `manifest_actual` carry the exact bytes, and every partial transition before reciprocal `manifest_durable` leaves an authenticated copy. If power loss drops the newly created temp directory entry after both actuals, or loses/rolls back the rename before its directory fsync, recovery recreates the reserved temp from those bytes and repeats the same rename. `PublishCompletedMarker` owns all three marker edges. Each method internally calls the receiver's unexported `commitCompletionPublicationProgress` for every edge and returns only after the matching outer/set reciprocal transition is durable. A crash returns no progress token. `RecoverPublication` remains the sole interface entry that may recover a filesystem fact or any publication progress after the reciprocal intent barrier；for `manifest_intent|manifest_actual`, it obtains bytes only from those reciprocal records, requires their byte equality、digest、view/set binding and reserved identities, and exact-rewrites/revalidates the temp/rename as above. It never calls a provider、private child、semantic validator、canonicalizer or caller, and never refetches/regenerates a nonce、time envelope or manifest. Zero/partial/full temp contents and a power-loss-missing temp entry are therefore all reconstructed or verified against the same durable bytes；a foreign identity、oversize value、byte/digest mismatch、one-sided unrepresented mutation or preexisting temp splice quarantines. No sink operation accepts a progress kind、digest、root、temp/final path or marker path, and no caller can skip、reorder、regenerate or choose an object. P09's high-level composite operator invokes these sink methods；the private progress values and commit/recover hooks are unnameable outside `internal/c12runnerprofile`, so no raw outer mutation method exists to bypass them. Static call-site tests require the `StageCanonicalManifest` byte slice to come directly from `CanonicalValidatedCompletionManifest` on an opaque successful `ValidateCompletionInputs` token, require its view to be the same operator-retained materialized view, and reject any alternate production call site、writer or preexisting temp.

`Acknowledge(set, lease)` first calls the coordinator proof hook and is permitted only when it proves the exact publication-progress sequence terminal at `completed_durable`, P09's matching manifest and `completed` marker durable for the same binding/generation and every set/outer adoption actual durable at the guard-selected epoch/full digest with matching reciprocal publish digest. It then uses a set-bound semantic permit to persist the ACK intent plus new publish digest, performs the exact channel ACK idempotently, and advances the guarded role state again with the ACK actual/receipt plus matching set actual/new publish digest before returning `nil`. The third ACK actual makes the set `fully_acked` but grants no finalization until two distinct canonical values have been frozen：the exact one-use E2 target transition and the immutable retained-payload/finalization-set audit projection. `PrepareCompletionRetainedFinalization(set,binding)` derives the exact `retained_complete(active=nil,next=N+1)` target exclusively from authenticated coordinator/set/outer actuals；separately it derives an immutable audit digest over the retained payload and finalization/set identity while excluding mutable `next`、`active`、execution-reservation and inner-cleanup fields. It cross-persists one `finalize_ready` record containing only the **source** epoch/full/publish digests, exact target-transition digest and immutable audit digest. It accepts no caller digest or bytes, does not stage a future full cell and contains neither post-prepare `D1/P1` nor a target `StateDigest`. The returned one-use proof instead carries the actual post-prepare `E1/D1/P1` plus both frozen digests. `RecoverCompletionRetainedFinalization` accepts no digest and converges only that pending reciprocal prepare transition. `CommitCompletionRetainedFinalization(set,binding,proof)` is the sole proof consumer：it rechecks current `E1/D1/P1`, internally constructs+file/directory-fsyncs exactly the target opposite-cell terminal candidate as `E2=E1+1, PreviousStateDigest=D1`, computes its full digest, advances+rereads the guard and marks the matching execution reservation `exiting`. No earlier terminal candidate is staged or overwritten by prepare. A crash before E2 repeats only that proof-bound target projection；after E2 there is no same-generation recovery. Only after the old exiting reservation's containment absence-proof+clear may a distinct `terminal_audit` reservation call `InspectFixedCompletionRetainedFinalization` to prove the selected immutable audit digest、powerless original finalization audit/set identity and authenticated closed lineage through zero or more closed `completed|aborted` revalidation edges to the current `retained_complete(active=nil)` state. The old set remains a powerless immutable audit record with no binding/channel/mutation authority；a current `retained_complete_active` state always rejects audit. `RecoverAcknowledge(set)` is the only ACK response-loss path：it accepts no kind、slot、sequence、nonce、digest or lease, reads the set's sole authenticated pending ACK intent plus matching adoption actual internally, re-authenticates the completed marker and converges that exact epoch/full/publish-digest sequence. If no ACK is pending, the caller must idempotently reacquire the exact fixed binding lease through `AdoptCommitted` and call `Acknowledge`;recovery never guesses which binding to start. No member path/key/lease bytes are retained and there is no caller-visible ACK receipt constructor. A different set、publish-pending generation、stale caller view、authenticated replayed state、unrepresented publish-business drift、missing completed marker or caller-invented selector rejects without ACK. Prepare/install/fsync/commit/adoption-set/adopt/materialize/ACK/finalize/audit seams are durable and idempotent by exact identity. Tests include `TestEvidenceChannelRecoverAcknowledgeReadsAuthenticatedPendingActual`, `TestEvidenceChannelBindingMutationRejectsStaleStateEpoch`, `TestEvidenceChannelBindingSurvivesExecutionReservationOnlyCAS`, `TestEvidenceChannelBindingSurvivesInnerCleanupOnlyCAS`, `TestEvidenceChannelPublicationProgressCrossPersistsPublishBinding`, `TestEvidenceChannelRecoverPublicationProgressClosesCrashSeams`, `TestEvidenceChannelFinalizeReadyBindsImmutableRetainedPayloadDigest`, `TestEvidenceChannelFinalizeReadyHasNoDigestCycleOrPrematureCell`, `TestEvidenceChannelDirectMutationCannotBypassCoordinator`, `TestEvidenceChannelRecoverFinalizationBeforeAndAfterGuardCAS`, `TestNativeRunnerCompletionTerminalFinalizationAuditIsPathlessReadOnly`, `TestNativeRunnerCompletionTerminalAuditSurvivesCompletedRevalidations`, `TestNativeRunnerCompletionRevalidationAbortClosesGeneration`, `TestNativeRunnerCompletionTerminalAuditSurvivesAbortedRevalidations`, `TestNativeRunnerCompletionModePhaseCrossClaimRejects` and `TestEvidenceChannelBindingRejectsBusinessDriftWithFreshStateEpoch` at every begin/adopt/progress/ACK/finalize hook, plus launcher death after Begin and between every intent/actual pair；only the exact pre-E2 same-generation successor may recover, while post-E2 create uses audit only. `TestEvidenceChannelRecoverPublicationProgressClosesCrashSeams` must additionally crash after each side of the reciprocal `manifest_intent`, with temp absent、zero-length、each partial prefix、full-but-unflushed、file-fsynced and immediately before/after each `manifest_actual` side；every recovery writes or verifies byte-identical intent bytes, preserves the original canonical digest/nonce/time envelope, leaves provider/private-child/validator/canonicalizer call counts unchanged from intent selection onward, and rejects oversize、one-side bytes、digest、view/set-binding、reserved-identity or temp-content splices.

Only the `completion` runner profile has receiver access. Windows/platform/authority final profiles have sender access and can neither adopt/read a transaction nor inspect any predecessor/other sender slot. `staged_diagnostic` has neither sender nor receiver and cannot construct, sign, prepare, inspect, begin/recover an adoption set, adopt or acknowledge a transaction. Receipt/bundle byte equality across all three transactions is therefore owned solely by completion after one exact three-member adoption set；no producer or diagnostic command asserts it by reading previous channel state. Tests strict-mutate every field/kind/namespace/slot/sequence/nonce/role/order/length/hash/object/media/EKU/domain and every prepare/install/fsync/commit/begin/recover-set/adopt/materialize/ACK seam, plus sender-read, receiver-enumeration, diagnostic-capability fabrication, cross-role, publish-pending cross-bind and set/lease replay attacks. Plan 09 may compose final producers/consumer policy and provide the authenticated `EvidencePublishPendingBinding`, but cannot redefine JCS/domain/set/WAL/ACK semantics or add a generic Inspect/read API.

The reciprocal `manifest_intent` order is uniquely set-first：write+file/parent-fsync the adoption-set pending old→new record carrying the bounded exact bytes/digest/bindings, then guard-CAS+reread the outer role record, then write+file/parent-fsync+reread the set actual with the new reciprocal publish digest. Outer-first is illegal. No temp filesystem operation occurs before all three steps agree；a crash after the set pending but before the outer CAS is exactly the create `session_cleanup_recovery` case and may only finish this state barrier from those bytes before cleanup.

The manifest branch of `TestEvidenceChannelRecoverPublicationProgressClosesCrashSeams` additionally power-cuts after each of those three fixed intent-order steps, after both `manifest_actual` records while simulating loss of the un-fsynced temp directory entry, and before/after rename、parent-directory fsync and each reciprocal `manifest_durable` side. It requires exact reconstruction from the still-retained bytes, proves both copies remain through reciprocal actual and may disappear only after reciprocal durable, keeps the original provider/private-child/validator/canonicalizer call counts unchanged and rejects an outer-first trace.

- [ ] **Step 5: Run Bash wrapper and cross-wrapper tests and verify GREEN**

Run: `go test ./internal/c12runnerprofile ./internal/c12evidence ./cmd/talenro-c12-runner-launcher -run '^TestB10ClosedHostGateExactCoversDeclaredTests$|^TestNativeRunnerLauncherImportClosureB10Frozen$|^TestGitBashWrapper|^TestWrapperParity|^TestEvidenceChannel|^TestCompletionCoordinator|^TestCompletionPublicationSink|^TestNativeRunnerWindowsSuccessCleanupMintsPublicationOnlyAfterAbsence$|^TestNativeRunnerWindowsAbortCleanupCannotMintPublication$|^TestNativeRunnerWindowsCleanupFailureCannotReachAttestors$|^TestNativeRunnerWindowsRecoveryDispatchIsClosedAndGuarded$|^TestNativeRunnerWindowsCleanupFinalizationReceiptCrashSeams$|^TestNativeRunnerWindowsPostCleanupAttestationResponseLossNeverResigns$|^TestNativeRunnerWindowsPostCleanupPublicationRecoversExactSenderState$|^TestNativeRunnerWindowsNextGenerationBlockedUntilCommittedPublication$|^TestWindowsRawRunnerprofileCallsOwnedOnlyByFinalScopesFacade$|^TestNativeRunnerPlatformPostCleanupPublicationRecoversExactSenderState$|^TestNativeRunnerAuthorityPostCleanupPublicationRecoversExactSenderState$|^TestNativeRunnerDirectSenderBindRejectsAllActiveSessions$' -count=1`

Run: `go test ./internal/c12evidence -run '^(TestTreeFactsCapsuleStrictValidation|TestWindowsFinalScopesSealedProducerPathIsCallableAndSoleOwned)$' -count=1`

Expected: PASS; injected primary/cleanup failures preserve bits, fake output canaries remain absent, either wrapper rejects the other scope enum, and every adoption-set crash seam recovers only the exact three durable bindings associated with the matching P09 publish-pending generation. No adopt returns before its actual fsync and no ACK occurs before the matching completed marker is durable.

- [ ] **Step 6: REFACTOR — freeze the final external Git Bash invocation and validate path handling**

Freeze the entire Plan 09 Task 8 Windows lifecycle behind the one prebuilt native public parent mode below. PowerShell owns no Go object/handle or output path；it only uses the call operator on the fixed installed helper and propagates `$LASTEXITCODE`:

```powershell
& 'C:\Program Files\Talenro\C12\talenro-c12-runner-launcher.exe' run-windows-final-scopes
$finalScopesExit = $LASTEXITCODE
if ($finalScopesExit -ne 0) { exit $finalScopesExit }
```

The only stable absolute operator path is the machine-base launcher, whose OS/application-control attestation is verified before it guard-selects and no-follow opens the immutable versioned helper by install-record receipt. The launcher installs+rereads the suspended-child execution reservation and then resumes that exact helper；the helper revalidates Authenticode/hash/OS/arch/catalog/toolchain identity plus its machine-local `C12NativeRunnerInstallReceiptV1` before its inherited dispatcher calls `OpenInstalledNativeRunnerSession(ctx, NativeRunnerWindowsScopes)`. That session supplies opaque canonical implementation object-database、owner-only external invocation-parent、recovery-slot、exact native-executable receipts and one `sender/windows_scopes` next-slot capability. No repo、script、tool、parent、channel、slot、sequence、output or digest flag/environment/global exists. Before capture or any ordinary resource create, the c12evidence Windows facade generates the run/invocation/snapshot/object-database/cwd/output/Docker reservations, derives its sealed plan and executes the exact `BeginProtectedWALKeyBootstrap(ctx,session,plan) → CreateOrRecoverProtectedWALKeyBootstrap → SealAndSelectProtectedWALKeyBootstrap({IssuedAt,ExpiresAt}) → OpenSelectedOwnershipWAL(ctx,selected)` chain. Only the unique final call may exact-create/reopen the planned WAL and returns only a sealed append surface after bootstrap+observed-identity file/directory fsync；the helper receives no raw key、binding、intent、selected token、path or signer. Only afterward may the long-lived parent append intents, capture/materialize the clean committed snapshot and minimal object database, derive each tracked PowerShell/Git-Bash script only from its committed blob SHA-256/mode plus held snapshot-root and no-follow opened-file identities, and execute only that derived absolute snapshot path under the receipt-pinned interpreter. Install receipts cover the helper/interpreters/native tools only, never the launcher or either script；a stable/direct helper path、fixed installed reference script、caller path、byte-equal external copy or post-open replacement rejects. It validates/pins PowerShell、Git Bash、cygpath、Docker and scanner executables, exact-creates the one 128-bit unique invocation root plus internal cwd/output children through the sealed WAL and directly launches the two fixed one-shot shell stages with only fixed stdio records. Every restart first dispatches shared `abort_unstarted|abort_intent|role_dispatch`; each abort route handles only its own pre-resource state and cannot enter Windows recovery. Its native `defer` covers every identity-validation/child/compare/cleanup/sign/handoff exception, restores child cwd state and kills exact Job/process trees；shell never receives a cleanup handle.

The parent completes both Docker lifecycles and retains only the two canonical receipt/bundle pairs、two unsigned scope fact sets and three role-bound attestor handles inside the bounded pre-clean composite. It compares all shared facts, selects success-origin cleanup, proves every Docker/export/minimal-object-database/capsule/snapshot/work/run-root resource absent, appends+file/directory-fsyncs terminal WAL, prepares the exact leaf receipt and drives `windows_cleanup_finalization_pending→leaf Recover/Commit→destroy/verify the bound WAL key→windows_post_cleanup_publication` without advancing the generation. That successor verifies the already-bound absence、terminal receipt and retired tombstone facts plus repository baseline, durably records each scope-signing intent, invokes each hardware attestor at most once, self-verifies and records the exact result, then binds only the fixed post-clean sender. It constructs a fresh-nonce transaction whose signed namespace/slot/sequence equal that sender, durably consumes/self-verifies the channel attestor and executes `Prepare→Commit→InspectSender(transaction_nonce)`. Commit atomically installs and fsyncs the complete immutable four-member transaction；response loss replays only the stored signature or own-nonce sender state. `FinalizeFixedWindowsPostCleanupPublication` alone marks the reservation exiting and advances terminal-inactive/next after exact committed inspection. Tests inject wrong implementation/script/tool identity and every origin、dispatch、cleanup、receipt、key-destroy、post-clean CAS、two attestation intent/result、sender/nonce、Prepare/Commit/Inspect/final-CAS seam；outcomes are unchanged or the exact transaction, never partial, and no next generation/signing route appears on abort or failure. B10 uses non-authoritative materializer fixtures；Plan 09 Task 8 is the first authoritative invocation.

Before the source freeze, add `TestC12DockerTaggedGateExactCoversDeclaredTests` to `internal/c12evidence/compose_test.go` with the exact four-root/two-tagged-file owner and transitive no-skip contract specified in the B10 exit gate below. This source/meta-test is committed now even though its owning-tag execution waits for the new relock projection；no source test may be added after the relock.

- [ ] **Step 7: Commit the final Task 7 source, tests and exact gates**

The five relock outputs must still be byte-identical to their preceding committed projection at this point. Stage and commit exactly the other 24 Task 7 paths, including the Docker meta-test、launcher-closure manifest、independent owner map and two no-argument gate scripts:

```bash
task7_source_files=(
  cmd/talenro-c12-runner-helper/main.go
  cmd/talenro-c12-runner-helper/main_test.go
  cmd/talenro-c12-runner-launcher/main_test.go
  internal/c12evidence/bash_wrapper_test.go
  internal/c12evidence/compose_test.go
  internal/c12evidence/runnerhelper.go
  internal/c12evidence/runnerhelper_test.go
  internal/c12evidence/wrapper_contract_test.go
  internal/c12runnerprofile/authority_cleanup.go
  internal/c12runnerprofile/authority_cleanup_test.go
  internal/c12runnerprofile/channel.go
  internal/c12runnerprofile/channel_test.go
  internal/c12runnerprofile/completion_state.go
  internal/c12runnerprofile/completion_state_test.go
  internal/c12runnerprofile/platform_resume.go
  internal/c12runnerprofile/platform_resume_test.go
  internal/c12runnerprofile/profile_test.go
  internal/c12runnerprofile/windows_cleanup.go
  internal/c12runnerprofile/windows_cleanup_test.go
  scripts/invoke-exact-b10-gates.ps1
  scripts/invoke-exact-b10-linux-adapter-gate.sh
  scripts/verify-c12.sh
  testdata/c12/b10-exact-owner-map.v1.json
  testdata/c12/launcher-import-closure.v1.json
)
task7_relock_outputs=(
  deploy/c12/compose.outer.yaml
  testdata/c12/integration-manifest.v1.json
  testdata/c12/toolchain-lock.v1.json
  testdata/c12/windows/runner-profile.v1.json
  testdata/c12/diagnostic/runner-profile.v1.json
)
git diff --exit-code -- "${task7_relock_outputs[@]}"
git diff --check
git add -- "${task7_source_files[@]}"
task7_source_expected="$(printf '%s\n' "${task7_source_files[@]}" | LC_ALL=C sort)"
task7_source_actual="$(git diff --cached --name-only | LC_ALL=C sort)"
[[ "$task7_source_actual" == "$task7_source_expected" ]]
git diff --cached --check
git commit -m "test: add final B10 verifier source gates"
git diff --exit-code
git diff --cached --exit-code
[[ -z "$(git ls-files --others --exclude-standard)" ]]
```

The new commit is the only source tree allowed to feed B10. Its `HEAD^{tree}`、index tree and working tree must agree；all relevant untracked files are absent. Any later change to these 24 paths invalidates B10 and restarts this source-commit step.

- [ ] **Step 8: Rebuild and atomically relock from the clean source commit**

Starting only from that clean Task 7 source commit and before any comparator、provisioning or B10 exit, rebuild both helper roles、both already-frozen launcher roles and the verifier image and run the sole atomic relock command one final time:

```text
go run ./cmd/talenro-artifact-scan lock-outer-images --compose deploy/c12/compose.outer.yaml --context-root deploy/c12 --integration-manifest testdata/c12/integration-manifest.v1.json --output-lock testdata/c12/toolchain-lock.v1.json --windows-runner-profile testdata/c12/windows/runner-profile.v1.json --diagnostic-runner-profile testdata/c12/diagnostic/runner-profile.v1.json --rewrite-compose --rewrite-integration-manifest --rewrite-windows-runner-profile --rewrite-diagnostic-runner-profile
```

`TestB10FixtureComparatorUsesRunMaterializer` is defined only in `internal/c12evidence/runnerhelper_test.go`. It builds both non-authoritative wrapper fixtures through the same task-scoped run materializer used by the native parent, exact-compares their common receipt/bundle/repository/spec/toolchain projections while preserving distinct scope/run/portability facts, and injects identity、result、cleanup and cross-scope splices. AST/go-types assertions require zero call or alias to any final P09 authoritative evidence facade、provider、attestor、channel sender or reusable output writer；the fixture can return only bounded unsigned comparison facts and always exact-cleans its run-materialized identities. Its unique definition、reachable-helper no-skip closure and single JSON run/pass are frozen by the full B10 manifest.

That transaction treats `c12_runner_launcher_windows_amd64` and `c12_runner_launcher_linux_amd64` as frozen native toolchain inputs alongside the two helper roles. Before rewrite it requires their independent-root reproducible builds、SBOM/provenance and OS-specific profile `BootstrapLauncherCatalogRole` projections to exact-match；the Linux launcher is not embedded in the verifier image. After the five outputs are durable **and the Step 9 projection commit names exactly those bytes**, each target host is eligible only if its existing OS image/deployment baseline supplies a valid `C12FixedRunnerLauncherAttestationV1` for the corresponding relocked launcher role. The baseline install is external to the five role installers and cannot be synthesized by a provision mode. Missing/wrong launcher bytes、path、machine image、application-control or toolchain projection blocks B10 before any role guard/cell is created. This adds frozen inputs, not rewrite outputs, so exact-five replacement atomicity remains unchanged.

This is the only authoritative B10 lock point. Before relock, require the freshly rebuilt Windows PE helper and Linux static helper/Git/build-tool **candidate identities** to exact-match the tracked `windows_scopes`/`staged_diagnostic` install and application-control policies plus toolchain native-role projections；require the diagnostic script/channel/provider/evidence/signing/publishing projection to remain empty/nonpublishing, the same candidate Linux helper digest in the rebuilt verifier image, and all three outer image references/integration roles to match the candidate toolchain projection. This pre-relock phase deliberately has no production bootstrap/session, install receipt, sealed install/launch generation, reusable evidence or final catalog dependency. The atomic rewrite failure matrix leaves all five tracked outputs unchanged. The lock also freezes the generic Go/OCI builder identities and future authority-operations reproducibility inputs (`SOURCE_DATE_EPOCH`, trimpath/build-id policy, OCI config/history/layer timestamps, tar order/uid/gid/mode/xattrs and compression parameters) without any future-tree/output digest; Plan 09 Task 6A creates the fixed-token `C12OperationsBuildPolicyV1` instance under this already locked schema/loader and may select those values but may not modify the lock or loader. Re-run Compose/helper/catalog scans after the rewrite, including the three exact per-image context manifests that exclude Compose/integration/toolchain/profile outputs, all three long-syntax read-only snapshot/capsule/minimal-object-database binds with `create_host_path:false`, the explicit stable network/volume names, the native-parent-owned stage state machine/no-shell-FD contract, exact P07 FD3 adoption tests, `TestCaptureStagedAndBuildClosedSetCleanupBarrier`, the diagnostic no-publisher capability test and the public-parent/private-child helper command surface. Tests rebuild, rewrite projections, then recompute every image context digest unchanged to prove the graph is acyclic. Plan 09's authority image uses only separate context `deploy/c12-authority/`. No helper/c12evidence/artifactscan/buildinputs source or context-manifest/static asset may differ from the Step 7 source commit；any such drift restarts Step 7. After the atomic command, exactly the five declared relock outputs may differ and no relevant untracked path may exist.

- [ ] **Step 9: Commit exactly the five B10 relock outputs**

Before distribution or any host provisioning, require `git diff HEAD --name-only` to equal the five-member `task7_relock_outputs` list, stage only that list, have the locked helper prove the index bytes equal the worktree bytes and record the candidate index tree, then commit it:

```bash
task7_relock_outputs=(
  deploy/c12/compose.outer.yaml
  testdata/c12/integration-manifest.v1.json
  testdata/c12/toolchain-lock.v1.json
  testdata/c12/windows/runner-profile.v1.json
  testdata/c12/diagnostic/runner-profile.v1.json
)
task7_relock_expected="$(printf '%s\n' "${task7_relock_outputs[@]}" | LC_ALL=C sort)"
task7_relock_actual="$(git diff HEAD --name-only -- | LC_ALL=C sort)"
[[ "$task7_relock_actual" == "$task7_relock_expected" ]]
[[ -z "$(git ls-files --others --exclude-standard)" ]]
git diff --check -- "${task7_relock_outputs[@]}"
git add -- "${task7_relock_outputs[@]}"
task7_relock_cached="$(git diff --cached --name-only | LC_ALL=C sort)"
[[ "$task7_relock_cached" == "$task7_relock_expected" ]]
git diff --exit-code
git diff --cached --check
git commit -m "build: relock final B10 execution identities"
git diff --exit-code
git diff --cached --exit-code
[[ -z "$(git ls-files --others --exclude-standard)" ]]
```

The locked helper's pre-commit candidate tree must equal the new `HEAD^{tree}` exactly. The projection commit may change only those five outputs, all of which are excluded from the three image build contexts；both rebuilt helper/launcher identities and every source/context digest must remain equal to the Step 7 source commit evidence. Any mismatch restarts the rebuild/relock rather than amending either commit.

- [ ] **Step 10: Run the committed B10 host gates, provisioning and fixture comparator**

Immediately after the five-output projection commit is distributed byte-identically, the Windows-scope runner host invokes only its installed/protected zero-extra-argument `provision-windows-scopes-runner` mode and the distinct Linux diagnostic host invokes only `provision-staged-diagnostic-runner`; no process/session crosses hosts and neither command accepts a role/path/profile/receipt/provider option. Each fixed mode internally performs `VerifyFixedMachineStateGuardProvider` → `OpenFixedMachineInstallerSession` → consuming `InstallFixedNativeRunnerProfile` for its compile-time role, which destroys the installer session/candidate handle before every return → `VerifyFixedNativeRunnerInstallation`, and returns one bounded role/status/install-generation/state-epoch/install-record-digest projection with no path/receipt/key bytes. At B10 the candidate handle binds the committed toolchain native-role projection、matching role profile and frozen launcher attestation, with no dependency on the future production catalog. Each first install begins at `(0,zero)` uninstalled/retryable and selects the epoch-1 inactive state plus immutable install record only at the sole guard advance；an existing role follows the guarded inactive-only reinstall protocol and preserves its next launch generation. B10 remains failed until both host modes exit zero, their bounded roles/projection commit match exactly, and the diagnostic verification proves its zero-capability projection. The internal audit calls do not call `OpenInstalledNativeRunnerSession`, mint runtime/channel/provider handles, change an active launch state or consume/advance a launch generation. A partial candidate on one host leaves an existing selected generation unchanged or, on first install, leaves that role uninstalled；it never affects the other host.

Before those two B10 provision modes, both hosts run the OS-neutral launcher and cleanup/staged manifests below from the relocked tree；Windows additionally runs the exact six-root wrapper manifest. Each host then runs its own production-adapter target：Windows exact-runs `TestB10ClosedHostGateExactCoversDeclaredTests|TestNativeRunnerWindowsStateGuardProductionAdapterRejectsFallback|TestNativeRunnerWindowsBootstrapLauncherProductionAdapterRejectsFallback|TestProtectedWALKeyWindowsProductionAdapter`, while Linux runs the same meta-test plus the three corresponding Linux names. Every command exact-compares `go test -list` output and requires one JSON `run`、one `pass`、zero root/descendant `skip` for every literal name；a tagged test absent on the current OS therefore fails rather than passing as zero-match. Then, inside the matching OS deployment transaction that supplies the protected candidate-package handle, the exact host commands are respectively `& 'C:\Program Files\Talenro\C12\talenro-c12-runner-launcher.exe' provision-windows-scopes-runner` and `/usr/libexec/talenro-c12/talenro-c12-runner-launcher provision-staged-diagnostic-runner`; each caller checks the exit immediately. Both host gate records are mandatory. The launcher is already present from the attested machine base；generation-1 provisioning therefore does not invoke a nonexistent helper, and neither command reveals a candidate/versioned-helper path.

Both hosts must additionally run `TestVerifyFixedNativeRunnerInstallationClosesB10AndB11ProjectionVariants` from `./internal/c12runnerprofile` before provisioning；the B10 fixtures require the selected B10 record's catalog projection to be absent/zero and exercise synthetic B11 records only inside the package test harness.

After Plan 09's last completion/helper/client source commit, B11 must repeat this transaction with the complete all-or-none group `--artifact-catalog testdata/c12/artifact-catalog.v1.json --trusted-time-provider-profile testdata/c12/platform/trusted-time-provider-profile.v1.json --windows-runner-profile testdata/c12/windows/runner-profile.v1.json --diagnostic-runner-profile testdata/c12/diagnostic/runner-profile.v1.json --platform-runner-profile testdata/c12/platform/runner-profile.v1.json --authority-runner-profile testdata/c12/authority/runner-profile.v1.json --completion-runner-profile testdata/c12/completion/runner-profile.v1.json` and `--rewrite-trusted-time-provider-profile --rewrite-windows-runner-profile --rewrite-diagnostic-runner-profile --rewrite-platform-runner-profile --rewrite-authority-runner-profile --rewrite-completion-runner-profile`, after rebuilding both helpers, verifier image and prebuilt completion child. Before relock it validates only those candidate native identities、tracked policies、toolchain native-role projections、catalog exact cover and diagnostic zero-capability projection, not production bootstrap/session/install receipts. That one atomic transaction updates exactly nine outputs—Compose、integration、toolchain、provider and all five runner projections—never the catalog/context manifests/operations policy；failure at any replacement leaves every input/output byte-identical. Only after all nine are durable and distributed from that same projection commit do the five target role hosts independently enter their matching authenticated OS deployment transactions and invoke their exact fixed protected provision modes: Windows `provision-windows-scopes-runner` and `provision-completion-runner`; Linux `provision-staged-diagnostic-runner`, `provision-platform-runner` and `provision-authority-runner`. Each mode owns one host/role/final-catalog+projection-bound candidate handle and local installer session and performs the same preflight/consuming-install/non-consuming Verify chain；the install call destroys the session/handle on every return, and there is no cross-host all-five session or runbook call to a sealed Go API. First installation remains `(0,zero)` uninstalled before its sole activation and then selects guarded epoch-1 inactive state plus install record；reinstall preserves guard/cells、next launch generation and any completion retained payload and rejects active/pending state. Task 8 requires five zero exits and exact role/projection-commit status cover before any evidence parent starts. Neither Verify nor relock calls `OpenInstalledNativeRunnerSession` or consumes/advances a launch generation. Only that clean relock/distribution/per-host-provision/Verify projection may feed authoritative scopes.

B11 must rebuild/recheck both launcher roles from the frozen Task 2 source and require byte identity with B10's launcher projections and every target's still-valid base attestation；it may not rewrite、reinstall or generation-select a launcher. P09's final catalog exact-covers both launcher binary/SBOM/provenance roles, while each final profile binds the matching `BootstrapLauncherCatalogRole`. These checks are inputs to the existing exact-nine relock and do not add a tenth output.

At B10 run only, first run the common origin/receipt/post-clean/delivery gate. It is the first literal manifest consumed by `TestB10ClosedHostGateExactCoversDeclaredTests`:

```text
function Invoke-ExactB10TestGate {
  param(
    [Parameter(Mandatory = $true)][ValidateSet('common','cleanup_contract','completion_bootstrap','launcher','full','windows_wrapper','windows_adapter')][string]$ManifestID,
    [Parameter(Mandatory = $true)][string[]]$Packages,
    [Parameter(Mandatory = $true)][string]$Pattern
  )
  $ownerMapPath = 'testdata/c12/b10-exact-owner-map.v1.json'
  $ownerSpec = Get-Content -LiteralPath $ownerMapPath -Raw | ConvertFrom-Json
  if ($ownerSpec.SchemaVersion -cne 'talenro-c12-b10-exact-owner-map/v1') { throw 'wrong B10 owner-map schema' }
  $expected = @($ownerSpec.Entries | Where-Object { @($_.ManifestIDs) -ccontains $ManifestID } | ForEach-Object Test | Sort-Object)
  if ($expected.Count -eq 0 -or @($expected | Sort-Object -Unique).Count -ne $expected.Count) { throw "invalid B10 owner-map manifest: $ManifestID" }
  $terms = @($Pattern -split '\|')
  if ($terms.Count -eq 0 -or @($terms | Where-Object { $_ -cnotmatch '^\^Test[A-Za-z0-9_]+\$$' }).Count -ne 0) {
    throw 'B10 pattern is not an exact top-level Test alternation'
  }
  $patternMembers = @($terms | ForEach-Object { $_.Substring(1, $_.Length - 2) } | Sort-Object)
  $patternDelta = @(Compare-Object -CaseSensitive $expected $patternMembers)
  if ($patternDelta.Count -ne 0) { $patternDelta | Out-String | Write-Error; throw "B10 pattern/owner-map mismatch: $ManifestID" }
  $listed = @(& go test @Packages -list $Pattern)
  if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
  $actual = @($listed | Where-Object { $_ -cmatch '^Test[A-Za-z0-9_]+$' } | Sort-Object)
  $listDelta = @(Compare-Object -CaseSensitive $expected $actual)
  if ($listDelta.Count -ne 0) { $listDelta | Out-String | Write-Error; exit 1 }
  $jsonLines = @(& go test @Packages -json -run $Pattern -count=1)
  if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
  $events = @($jsonLines | ForEach-Object { $_ | ConvertFrom-Json })
  $topLevelNames = @($events | Where-Object { $_.Test -and $_.Test -cnotmatch '/' } | ForEach-Object Test | Sort-Object -Unique)
  $eventDelta = @(Compare-Object -CaseSensitive $expected $topLevelNames)
  if ($eventDelta.Count -ne 0) { $eventDelta | Out-String | Write-Error; exit 1 }
  foreach ($name in $expected) {
    $run = @($events | Where-Object { $_.Test -ceq $name -and $_.Action -ceq 'run' })
    $pass = @($events | Where-Object { $_.Test -ceq $name -and $_.Action -ceq 'pass' })
    $skip = @($events | Where-Object { $_.Test -ceq $name -and $_.Action -ceq 'skip' })
    if ($run.Count -ne 1 -or $pass.Count -ne 1 -or $skip.Count -ne 0) {
      Write-Error "B10 test did not run+pass exactly once without skip: $name"
      exit 1
    }
    $descendantSkip = @($events | Where-Object {
      $_.Action -ceq 'skip' -and $_.Test -and $_.Test.StartsWith("$name/", [System.StringComparison]::Ordinal)
    })
    if ($descendantSkip.Count -ne 0) {
      Write-Error "B10 descendant test skipped under declared parent: $name"
      exit 1
    }
  }
}

$commonPackages = @('./internal/c12runnerprofile', './internal/c12evidence', './internal/c12cleanup', './cmd/talenro-c12-runner-launcher', './cmd/talenro-c12-runner-helper')
$commonPattern = '^TestB10ClosedHostGateExactCoversDeclaredTests$|^TestProtectedWALKeyIntentClosesEveryPreCapsuleCrashSeam$|^TestProtectedWALKeyBootstrapRawRunnerprofileCallsOwnedOnlyByCleanupFacade$|^TestCleanupCoreAliasesAreAssignmentIdentical$|^TestCleanupCoreDependencyGraphIsAcyclic$|^TestCleanupOnlyCapsuleClosesEveryPreManifestPowerLossSeam$|^TestCleanedTombstoneClosesKeyAndSlotCrashSeams$|^TestNativeRunnerWindowsSuccessCleanupMintsPublicationOnlyAfterAbsence$|^TestNativeRunnerWindowsAbortCleanupCannotMintPublication$|^TestNativeRunnerWindowsCleanupFailureCannotReachAttestors$|^TestNativeRunnerWindowsRecoveryDispatchIsClosedAndGuarded$|^TestNativeRunnerWindowsCleanupFinalizationReceiptCrashSeams$|^TestNativeRunnerWindowsPostCleanupAttestationResponseLossNeverResigns$|^TestNativeRunnerWindowsPostCleanupPublicationRecoversExactSenderState$|^TestNativeRunnerWindowsNextGenerationBlockedUntilCommittedPublication$|^TestWindowsRawRunnerprofileCallsOwnedOnlyByFinalScopesFacade$|^TestNativeRunnerPlatformResumeManifestUsesSinkBoundAttestor$|^TestNativeRunnerPlatformValidationDispatchSeparatesFreshFromAmbiguousExchange$|^TestNativeRunnerPlatformPromotionRecoveryDispatchIsClosedAndGuarded$|^TestNativeRunnerPlatformTrustedTimeExchangeCrashRecoveryNeverRefetches$|^TestNativeRunnerPlatformSuccessCleanupMintsPublicationOnlyAfterAbsence$|^TestNativeRunnerPlatformAbortCleanupCannotMintPublication$|^TestNativeRunnerPlatformCleanupFinalizationReceiptCrashSeams$|^TestNativeRunnerPlatformPostCleanupAttestationResponseLossNeverResigns$|^TestNativeRunnerCleanupFinalizationFirstSelectedOriginWins$|^TestNativeRunnerPlatformPostCleanupPublicationRecoversExactSenderState$|^TestNativeRunnerAuthoritySuccessCleanupMintsPublicationOnlyAfterAbsence$|^TestNativeRunnerAuthorityAbortCleanupCannotMintPublication$|^TestNativeRunnerAuthorityCleanupFailureCannotReachAttestors$|^TestNativeRunnerAuthorityActiveCrashDefaultsToAbortCleanup$|^TestNativeRunnerAuthorityRecoveryDispatchIsClosedAndGuarded$|^TestNativeRunnerAuthorityCleanupRecoveryIsSelectorFree$|^TestNativeRunnerAuthorityCleanupFinalizationReceiptCrashSeams$|^TestNativeRunnerAuthorityPostCleanupAttestationResponseLossNeverResigns$|^TestNativeRunnerAuthorityPostCleanupPublicationRecoversExactSenderState$|^TestNativeRunnerDirectSenderBindRejectsAllActiveSessions$|^TestNativeRunnerAuthorityFinalizationFirstSelectedOriginWins$|^TestNativeRunnerCompletionRevalidationDispatchIsClosedAndGuarded$|^TestNativeRunnerCompletionRevalidationPostTerminalResponseLossAudit$|^TestNativeRunnerCompletionRevalidationOutcomeDeliveryMustAckBeforeNextOpen$|^TestCompletionInnerFinalizersRequireExactLeafReceipt$|^TestCompletionInnerFinalizationReceiptCrashSeams$|^TestCompletionRetainedOutcomeConsumesSameCleanupReceipt$'
Invoke-ExactB10TestGate -ManifestID common -Packages $commonPackages -Pattern $commonPattern
```

Then run the second literal cleanup-adapter/staged-diagnostic contract manifest:

```text
$cleanupContractPackages = @('./internal/c12runnerprofile', './internal/c12evidence', './internal/c12cleanup')
$cleanupContractPattern = '^TestB10ClosedHostGateExactCoversDeclaredTests$|^TestCleanupInspectorDeleterSetExactCoverAndNoRegistry$|^TestRecoverExactUsesBoundAdapterAndWritesActualBeforeDelete$|^TestNativeRunnerStagedDiagnosticCleanupRecoveryIsClosedAndGuarded$|^TestNativeRunnerStagedDiagnosticCleanupCapabilityUsesExactAdapterSet$'
Invoke-ExactB10TestGate -ManifestID cleanup_contract -Packages $cleanupContractPackages -Pattern $cleanupContractPattern
```

Then run the third literal completion inner-bootstrap crash manifest:

```text
$completionBootstrapPackages = @('./internal/c12runnerprofile', './internal/c12evidence', './internal/c12cleanup')
$completionBootstrapPattern = '^TestB10ClosedHostGateExactCoversDeclaredTests$|^TestCompletionInnerCleanupProtectedWALKeyIntentCrashSeams$'
Invoke-ExactB10TestGate -ManifestID completion_bootstrap -Packages $completionBootstrapPackages -Pattern $completionBootstrapPattern
```

Then run the fourth literal closed launcher/guard/containment and nested-verifier separation gate:

```text
$launcherPackages = @('./internal/c12runnerprofile', './internal/c12evidence', './cmd/talenro-c12-runner-launcher', './cmd/talenro-c12-runner-helper')
$launcherPattern = '^TestB10ClosedHostGateExactCoversDeclaredTests$|^TestNativeRunnerFixedLauncherAttestationBindsMachineBase$|^TestNativeRunnerLauncherImportClosureB10Frozen$|^TestNativeRunnerOpenFixedMachineInstallerSessionRequiresProtectedProvisioner$|^TestNativeRunnerProvisionOutsideDeploymentTransactionRejects$|^TestNativeRunnerCandidatePackageHandleIsRoleProjectionAndHostBound$|^TestInstallFixedNativeRunnerProfileConsumesSessionAndDestroysCandidateOnEveryReturn$|^TestNativeRunnerBootstrapLauncherClosesFirstInstallAndReinstall$|^TestNativeRunnerBootstrapLauncherUsesGuardSelectedInstallRecord$|^TestNativeRunnerLaunchReservationCrashSeams$|^TestNativeRunnerLauncherDeathCannotLeaveTwoLiveParents$|^TestNativeRunnerRecoveryProvesBoundParentAbsentBeforeSuccessor$|^TestNativeRunnerCrossBootClearsPreOpenReservationBeforeFreshOpen$|^TestNativeRunnerCrossBootClearsExitingReservationBeforeFreshOpen$|^TestNativeRunnerPlatformResumeTransitionIsGuardedAndOneUse$|^TestNativeRunnerPlatformPhaseBUsesPathlessSuccessorNotOpen$|^TestNativeRunnerPlatformResumeDowngradeCannotRegainWorkloadAuthority$|^TestNativeRunnerCompletionTerminalFinalizationAuditIsPathlessReadOnly$|^TestNativeRunnerCompletionTerminalAuditSurvivesCompletedRevalidations$|^TestNativeRunnerCompletionRevalidationAbortClosesGeneration$|^TestNativeRunnerCompletionRevalidationOppositeOutcomeReturnsConflict$|^TestNativeRunnerCompletionRevalidationPendingUsesNoHandleInnerFinalization$|^TestNativeRunnerCompletionTrustedTimeHeadNeverRollsBackAcrossAbortedLineage$|^TestNativeRunnerCompletionTerminalAuditSurvivesAbortedRevalidations$|^TestCompletionValidationChildRuntimeBindsCoordinatorInnerAndInstalledProjection$|^TestCompletionValidationChildRuntimeResultIsBoundedSealedAndSingleUse$|^TestCompletionValidationChildTrackedTreeReaderIsPathlessAndExact$|^TestCompletionValidationChildRuntimeHasNoEvidenceSemanticDependency$|^TestCompletionValidationChildRuntimeOwnedOnlyByCompositePackage$|^TestNativeRunnerCompletionModePhaseCrossClaimRejects$|^TestRunnerHelperContainerVerifierIngressDoesNotRequireMachineLauncher$|^TestWindowsOuterParentCannotReachCoreSmokeSlotOrTrustedSource$'
Invoke-ExactB10TestGate -ManifestID launcher -Packages $launcherPackages -Pattern $launcherPattern
```

Then run the fifth literal full B10 closed-set gate:

```text
$fullPackages = @('./internal/c12runnerprofile', './internal/c12evidence', './internal/artifactscan', './cmd/talenro-artifact-scan', './cmd/talenro-c12-runner-launcher', './cmd/talenro-c12-runner-helper')
$fullPattern = '^TestB10ClosedHostGateExactCoversDeclaredTests$|^TestProtectedWALKeyIntentClosesEveryPreCapsuleCrashSeam$|^TestProtectedWALKeyBootstrapRawRunnerprofileCallsOwnedOnlyByCleanupFacade$|^TestNativeRunnerProfileStrictFiveRoleTokens$|^TestStagedDiagnosticProfileIsNonpublishing$|^TestVerifyFixedNativeRunnerInstallationIsNonConsuming$|^TestVerifyFixedNativeRunnerInstallationClosesB10AndB11ProjectionVariants$|^TestNativeRunnerLaunchGenerationsAreRoleIndexed$|^TestNativeRunnerSameRoleActiveRunIsExclusiveAndRearmsAfterCleanup$|^TestNativeRunnerPlatformResumeSuccessorUsesSameGeneration$|^TestNativeRunnerCompletionRoleStateUsesSingleRecoverySlot$|^TestNativeRunnerCompletionRetainedTransitionUsesPrepareAndCommitGuardAdvances$|^TestNativeRunnerCompletionTerminalFinalizationAuditIsPathlessReadOnly$|^TestNativeRunnerCompletionTerminalAuditSurvivesCompletedRevalidations$|^TestNativeRunnerCompletionRevalidationAbortClosesGeneration$|^TestNativeRunnerCompletionTerminalAuditSurvivesAbortedRevalidations$|^TestCompletionValidationChildRuntimeBindsCoordinatorInnerAndInstalledProjection$|^TestCompletionValidationChildRuntimeResultIsBoundedSealedAndSingleUse$|^TestCompletionValidationChildRuntimeHasNoEvidenceSemanticDependency$|^TestCompletionValidationChildRuntimeOwnedOnlyByCompositePackage$|^TestNativeRunnerCompletionModePhaseCrossClaimRejects$|^TestNativeRunnerRoleStateRejectsAuthenticatedRollback$|^TestNativeRunnerRoleStateGuardCrashSeams$|^TestNativeRunnerFirstInstallInitializesRoleState$|^TestNativeRunnerReinstallPreservesLaunchStateAndRetainedPayload$|^TestNativeRunnerOpenFixedMachineInstallerSessionRequiresProtectedProvisioner$|^TestNativeRunnerProvisionOutsideDeploymentTransactionRejects$|^TestNativeRunnerCandidatePackageHandleIsRoleProjectionAndHostBound$|^TestInstallFixedNativeRunnerProfileConsumesSessionAndDestroysCandidateOnEveryReturn$|^TestNativeRunnerFixedProvisionModesAreRoleAndOSClosed$|^TestNativeRunnerBootstrapLauncherClosesFirstInstallAndReinstall$|^TestNativeRunnerBootstrapLauncherUsesGuardSelectedInstallRecord$|^TestNativeRunnerFixedLauncherAttestationBindsMachineBase$|^TestNativeRunnerCompletionRoleStateCoordinatorRequiresSessionOrFixedRecovery$|^TestNativeRunnerCompletionRoleStateCoordinatorRejectsIllegalPhaseGenerationAndPayload$|^TestNativeRunnerCompletionRoleStateInspectIsNonConsuming$|^TestNativeVerifierParentCoreSmokeFD3AdoptionSequence$|^TestNativeVerifierParentRejectsCoreSmokeHandoffBypass$|^TestRunnerHelperCoreSmokeFD3Only$|^TestTreeFactsCapsuleStrictValidation$|^TestWindowsFinalScopesSealedProducerPathIsCallableAndSoleOwned$|^TestWindowsOuterParentCannotReachCoreSmokeSlotOrTrustedSource$|^TestCompletionCoordinatorHasNoRawMutationSurface$|^TestCompletionCoordinatorRejectsCallerDefinedPermit$|^TestCompletionInnerCleanupDescriptorIsBoundedAndSealed$|^TestOpenCompletionInnerCleanupForExactCleanupIsSoleDescriptorConsumer$|^TestCompletionPublicationSinkHasNoCallerPathKindOrDigest$|^TestCompletionPublicationSinkBindsExactCoordinatorAndSet$|^TestCompletionPublicationSinkStrictlyDecodesFrozenManifestProjection$|^TestCompletionPublicationSinkRejectsViewManifestBindingMismatch$|^TestEvidenceChannelSeparateSenderReceiverCapabilities$|^TestEvidenceChannelAdoptCommittedIsNonEnumerable$|^TestEvidenceChannelAdoptionSetIsDurableBeforeFirstAdopt$|^TestEvidenceChannelRecoverAdoptionSetClosesEveryCrashSeam$|^TestEvidenceChannelAdoptionSetBindsCompletionPublishPending$|^TestEvidenceChannelBindingMutationCrossBindsBeforeReturn$|^TestEvidenceChannelBindingMutationRejectsStaleStateEpoch$|^TestEvidenceChannelBindingSurvivesExecutionReservationOnlyCAS$|^TestEvidenceChannelBindingSurvivesInnerCleanupOnlyCAS$|^TestEvidenceChannelBindingRejectsBusinessDriftWithFreshStateEpoch$|^TestEvidenceChannelPublicationProgressCrossPersistsPublishBinding$|^TestEvidenceChannelReciprocalDigestGraphIsAcyclic$|^TestEvidenceChannelSetActualDigestNeverFeedsPublishBinding$|^TestEvidenceChannelRecoverPublicationProgressClosesCrashSeams$|^TestEvidenceChannelFinalizeReadyBindsImmutableRetainedPayloadDigest$|^TestEvidenceChannelFinalizeReadyHasNoDigestCycleOrPrematureCell$|^TestEvidenceChannelDirectMutationCannotBypassCoordinator$|^TestEvidenceChannelRecoverFinalizationBeforeAndAfterGuardCAS$|^TestEvidenceChannelRecoverAcknowledgeReadsAuthenticatedPendingActual$|^TestEvidenceChannelAcknowledgePersistsOuterActualBeforeReturn$|^TestEvidenceChannelCrashSeams$|^TestB10FixtureComparatorUsesRunMaterializer$|^TestCaptureStagedAndBuildClosedSetCleanupBarrier$|^TestFixedOuterImageBuildContextSchemaAndExactCover$|^TestFixedOperationsBuildPolicySchemaAndAcyclic$|^TestOuterImageBuildContextsAreClosedAndAcyclic$'
Invoke-ExactB10TestGate -ManifestID full -Packages $fullPackages -Pattern $fullPattern
```

Run the sixth literal Windows wrapper manifest on the Windows target in the **same PowerShell session** that defined `Invoke-ExactB10TestGate`:

```powershell
$windowsWrapperPackages = @('./internal/c12runnerprofile', './internal/c12evidence', './cmd/talenro-c12-runner-helper')
$windowsWrapperPattern = '^TestB10ClosedHostGateExactCoversDeclaredTests$|^TestPowerShellWrapperUsesLockedSnapshotIdentityAndNoAmbientTools$|^TestPowerShellWrapperFullLifecycleCleanupAndZeroPrematurePublication$|^TestRunnerHelperWindowsFinalScopesClosedCommandSurface$|^TestGitBashWrapperUsesLockedSnapshotIdentityAndNoAmbientTools$|^TestGitBashWrapperFullLifecycleParityAndTamperHasZeroChannelSideEffect$|^TestWrapperParityExactCoversPowerShellAndGitBashStages$'
Invoke-ExactB10TestGate -ManifestID windows_wrapper -Packages $windowsWrapperPackages -Pattern $windowsWrapperPattern
```

Run the seventh literal production-adapter manifest on the Windows target in that same PowerShell session:

```powershell
$windowsAdapterPackages = @('./internal/c12runnerprofile', './internal/c12cleanup', './cmd/talenro-c12-runner-launcher')
$windowsAdapterPattern = '^TestB10ClosedHostGateExactCoversDeclaredTests$|^TestNativeRunnerWindowsStateGuardProductionAdapterRejectsFallback$|^TestNativeRunnerWindowsBootstrapLauncherProductionAdapterRejectsFallback$|^TestProtectedWALKeyWindowsProductionAdapter$'
Invoke-ExactB10TestGate -ManifestID windows_adapter -Packages $windowsAdapterPackages -Pattern $windowsAdapterPattern
```

Run the eighth literal production-adapter manifest independently on the Linux target. This Bash gate exact-compares the listed top-level names and then checks the JSON event stream without `jq` or a host-language skip path:

```bash
set -euo pipefail
linux_adapter_packages=(./internal/c12runnerprofile ./internal/c12cleanup ./cmd/talenro-c12-runner-launcher)
linux_adapter_pattern='^TestB10ClosedHostGateExactCoversDeclaredTests$|^TestNativeRunnerLinuxStateGuardProductionAdapterRejectsFallback$|^TestNativeRunnerLinuxBootstrapLauncherProductionAdapterRejectsFallback$|^TestProtectedWALKeyLinuxProductionAdapter$'
linux_adapter_expected=(
  TestB10ClosedHostGateExactCoversDeclaredTests
  TestNativeRunnerLinuxBootstrapLauncherProductionAdapterRejectsFallback
  TestNativeRunnerLinuxStateGuardProductionAdapterRejectsFallback
  TestProtectedWALKeyLinuxProductionAdapter
)
linux_expected_text="$(printf '%s\n' "${linux_adapter_expected[@]}" | LC_ALL=C sort)"
linux_actual_text="$(go test "${linux_adapter_packages[@]}" -list "$linux_adapter_pattern" | grep -E '^Test[A-Za-z0-9_]+$' | LC_ALL=C sort)"
if [[ "$linux_actual_text" != "$linux_expected_text" ]]; then
  diff -u <(printf '%s\n' "$linux_expected_text") <(printf '%s\n' "$linux_actual_text")
  exit 1
fi
linux_json_file="$(mktemp)"
trap 'rm -f -- "$linux_json_file"' EXIT
go test "${linux_adapter_packages[@]}" -json -run "$linux_adapter_pattern" -count=1 >"$linux_json_file"
for name in "${linux_adapter_expected[@]}"; do
  run_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"run\""){c++} END{print c+0}' "$linux_json_file")"
  pass_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"pass\""){c++} END{print c+0}' "$linux_json_file")"
  skip_count="$(awk -v n="$name" 'index($0,"\"Test\":\"" n "\"") && index($0,"\"Action\":\"skip\""){c++} END{print c+0}' "$linux_json_file")"
  if [[ "$run_count" != 1 || "$pass_count" != 1 || "$skip_count" != 0 ]]; then
    printf 'B10 Linux adapter test did not run+pass exactly once without skip: %s\n' "$name" >&2
    exit 1
  fi
  descendant_skip_count="$(awk -v n="$name/" 'index($0,"\"Test\":\"" n) && index($0,"\"Action\":\"skip\""){c++} END{print c+0}' "$linux_json_file")"
  if [[ "$descendant_skip_count" != 0 ]]; then
    printf 'B10 Linux adapter descendant skipped under declared parent: %s\n' "$name" >&2
    exit 1
  fi
done
rm -f -- "$linux_json_file"
trap - EXIT
```

The five OS-neutral manifests、Windows wrapper manifest and Windows production-adapter manifest run in one pinned Windows PowerShell process；the Linux production-adapter manifest runs on the attested Linux host. The meta-test parses all eight frozen member lists regardless of build tag, while the matching-host list/JSON checks prove the OS-specific members and every non-skipped descendant really executed. The seven PowerShell blocks above are the literal manifest bodies committed in the no-argument `scripts/invoke-exact-b10-gates.ps1`; the Linux block is the literal body committed in no-argument `scripts/invoke-exact-b10-linux-adapter-gate.sh`. At B10 execute exactly `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/invoke-exact-b10-gates.ps1` on the pinned Windows host and `bash scripts/invoke-exact-b10-linux-adapter-gate.sh` on the attested Linux host. The PowerShell helper precompares each literal name list with the committed owner map；both mandatory meta-test executions exact-compare the full script/package/owner/membership tuple before either command can succeed. Both scripts refuse positional arguments、pattern/package/environment overrides or a dirty/missing owner-map file.

The test-only materializer derives both independently random non-authoritative verifier-result/receipt/bundle fixture roots beneath the same canonicalized、no-follow、owner-only repository-external parent capability already authenticated by the native session. It O_EXCL-creates them, verifies distinct file identities, passes its retained handles directly to the comparator, and finally removes only those exact WAL-owned roots. Neither side reads `$env:TEMP`/account defaults or recomputes a path；it may not require `C:/Users/runner/...`, `D:/c12-evidence/...` or any future final output. A missing materializer return, same root/name/identity, fixed external literal, repo-local path、ambient temp or reusable output is RED.

The final-tree comparator and four-file attested handoff are private terminal stages of `run-windows-final-scopes`, not separately callable fragments. The native parent uses its already pinned scanner identity and exact in-memory/output handles；it never recomputes an account/TEMP path or executes a literal/PATH scanner. The materializer test requires byte-identical canonical receipts/bundles from two independent roots, distinct unsigned scopes, equal daemon/repo/tree/spec/toolchain/three-release/authority-operations/helper/integration/image inputs, fresh spec validation, constant-time auxiliary equality, and complete cleanup before either signature. It proves the terminal transaction contains signed namespace/slot/sequence plus the PowerShell receipt/bundle and both scope identities, that the Git Bash scope remains available to completion, and that failure before cleanup/sign/Prepare leaves the channel unchanged while acknowledgement loss converges by exact `InspectSender(transaction_nonce)` to either unchanged/absent or that exact immutable committed transaction. No seam leaves partial channel state or an invocation root. An old two-member set, changed member order/path/hash, different tracked tree, second-amendment tamper, receipt/bundle mismatch, operations binary/image splice or binary-only receipt fails. Neither the fixture result nor this comparator generates upgrade/staging/recovery evidence or reads a predecessor transaction; three-way receipt/bundle equality remains completion-only.

- [ ] **Step 11: Run full non-provider regressions**

Run: `go test ./... -count=1 -timeout 30m`

Run: `go test -race ./... -count=1 -timeout 30m`

Run: `go vet ./...`

Run: `go tool golangci-lint run ./...`

Run: `git diff --check`

Expected: all PASS.

## B10 exit gate

B10 closes when `BuildTrackedC12SpecSet` accepts the fixed tracked manifest with the exact three approved members above, computes the one post-v7 `SpecDigest`, the scanner/evidence/wrapper tests prove a required separately classified authority-operations binary+image input, and both Windows wrappers pass their full deterministic lifecycle against fixed test receipts. Because B11 has not yet produced that auxiliary binary, B10 does not mint reusable authoritative Windows scope evidence；Plan 09 Task 8 must build/scan the final operations artifact and run both real wrappers from the final tracked tree before completion. The B10 exit record includes the builder-derived `SpecDigest`, states that nested-Docker evidence is only `container_deterministic`, that Xray and sing-box each passed its own real protocol smoke, that final closed-set scope generation is deferred to the B11 artifact, and that no host SELinux/sysctl/TPM/secure-time/operator-trust/authority-fence, upgrade, staging, recovery or Down evidence has been generated.

Before declaring that exit, rerun the Step 7-committed `TestC12DockerTaggedGateExactCoversDeclaredTests`. It parses all package test files regardless of tag, requires exactly one definition at the fixed owner for itself、`TestOuterComposeHasOnlyApprovedResources`、`TestInnerDaemonRejectsImagePull` and `TestAuthorityScriptAgainstInnerDaemon`, rejects duplicate/renamed definitions and, with `go/types` call resolution, rejects `Skip`/`Skipf`/`SkipNow` in all four tests plus every package-local helper transitively reachable from them. It exact-compares its three substantive-name/owner table plus the two tagged-file/first-line table to the literal manifest below. Then run this exact owning-tag static/list/JSON gate against the clean projection commit:

```powershell
$expected = '//go:build c12_docker'
$files = @('internal/c12evidence/compose_c12_docker_test.go','internal/c12evidence/authority_script_c12_docker_test.go')
$tracked = & 'C:\Program Files\Git\cmd\git.exe' ls-files --error-unmatch -- @files
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
if ((@($tracked).Count -ne 2) -or (Compare-Object -CaseSensitive ($files | Sort-Object) (@($tracked) | Sort-Object))) { throw 'c12_docker tracked exact set mismatch' }
foreach ($file in $files) { if ((Get-Content -LiteralPath $file -TotalCount 1) -cne $expected) { throw "$file must start with $expected" } }
$tagPattern = '^(TestAuthorityScriptAgainstInnerDaemon|TestC12DockerTaggedGateExactCoversDeclaredTests|TestInnerDaemonRejectsImagePull|TestOuterComposeHasOnlyApprovedResources)$'
$tagExpected = @(
  'TestAuthorityScriptAgainstInnerDaemon',
  'TestC12DockerTaggedGateExactCoversDeclaredTests',
  'TestInnerDaemonRejectsImagePull',
  'TestOuterComposeHasOnlyApprovedResources'
)
$tagListed = @(& 'C:\Program Files\Go\bin\go.exe' test -tags=c12_docker ./internal/c12evidence -list $tagPattern)
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$tagActual = @($tagListed | Where-Object { $_ -cmatch '^Test[A-Za-z0-9_]+$' } | Sort-Object)
$tagDelta = @(Compare-Object -CaseSensitive $tagExpected $tagActual)
if ($tagDelta.Count -ne 0) { $tagDelta | Out-String | Write-Error; exit 1 }
$tagJSONLines = @(& 'C:\Program Files\Go\bin\go.exe' test -tags=c12_docker ./internal/c12evidence -json -run $tagPattern -count=1 -timeout 15m)
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$tagEvents = @($tagJSONLines | ForEach-Object { $_ | ConvertFrom-Json })
foreach ($name in $tagExpected) {
  $run = @($tagEvents | Where-Object { $_.Test -ceq $name -and $_.Action -ceq 'run' })
  $pass = @($tagEvents | Where-Object { $_.Test -ceq $name -and $_.Action -ceq 'pass' })
  $skip = @($tagEvents | Where-Object { $_.Test -ceq $name -and $_.Action -ceq 'skip' })
  if ($run.Count -ne 1 -or $pass.Count -ne 1 -or $skip.Count -ne 0) {
    Write-Error "c12_docker test did not run+pass exactly once without skip: $name"
    exit 1
  }
  $descendantSkip = @($tagEvents | Where-Object {
    $_.Action -ceq 'skip' -and $_.Test -and $_.Test.StartsWith("$name/", [System.StringComparison]::Ordinal)
  })
  if ($descendantSkip.Count -ne 0) {
    Write-Error "c12_docker descendant test skipped under declared parent: $name"
    exit 1
  }
}
```

Before this block, the two exact non-consuming `VerifyFixedNativeRunnerInstallation` calls validate the installed Windows and diagnostic native executable identities/receipts/projections；diagnostic verification additionally proves the zero channel/provider/evidence/signing/publishing projection. Neither role is opened and neither launch generation is consumed by this B10 exit audit. The exact list rejects zero/partial/extra matches, JSON requires one run+pass and zero skip for all four names and rejects any descendant skip below them, and the meta-test prevents weakening the owner/name list or hiding a skip in a reachable helper. An untracked correctly tagged file, missing/extra `ls-files` output, wrong first line, profile/install-projection mismatch or tagged-test nonzero exit is fail-closed.

Both files must already be created and tracked by their owning tasks. An absent file, tag on any line other than the first, untagged duplicate Docker test, wrong helper OS/arch role, context-init premature exit, failed safe-export copy contract or integration-group residue blocks B10.

- [ ] **Step 12: Verify the clean committed B10 exit tree**

Only after the exact owning-tag static/list/JSON gate above has passed, prove that neither host execution nor any gate changed source、projection or untracked state. There is no post-gate source commit or output restage:

```bash
git diff --check
git diff --exit-code
git diff --cached --exit-code
[[ -z "$(git ls-files --others --exclude-standard)" ]]
```

The final B10 exit names the Step 9 projection commit and its parent Step 7 source commit. A changed source/context digest、dirty index/worktree、recreated output or relevant untracked path invalidates the exit and requires repeating Steps 7–12.
