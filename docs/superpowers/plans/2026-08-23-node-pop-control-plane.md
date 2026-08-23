# Talenro C1.2 Node and POP Control Plane Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有 Talenro Go 模块化单体中交付可审计的节点/POP 库存、外部权威防回滚、仅出站 mTLS node agent、签名 desired/recovery state、独立 node-core-supervisor、fixture/Xray/sing-box 外部进程适配、健康容量事实，以及四作用域权威完成证据。

**Architecture:** PostgreSQL 保存库存、身份、签名意图、恢复、审计与最新 observation 等权威事实；独立 rollback-resistant provider 以 epoch/sequence 封住 PostgreSQL PITR 回放。`control-api` 增加 bootstrap、agent、operator 三个隔离 TLS listener；`node-agent` 验证 host-deployed trust、签名状态和本地防回滚高水位；root-owned `node-core-supervisor` 以严格本地协议、租约、密封配置、cgroup/namespace/seccomp/SELinux 边界管理预装的外部 core。Docker 只提供确定性容器证据，真实 Linux platform/provider 与 authority-fence/PITR 由独立 attested gate 证明。

**Tech Stack:** Go 1.26.5、PostgreSQL 18.4、Redis 8.8.1、NATS 2.14.3、pgx 5.10.0、OpenAPI 3.0.3、oapi-codegen 2.8.0、Protobuf Go 1.36.11、Buf 1.72.0、sqlc 1.31.1、goose 3.27.1、Prometheus Go client 1.23.2、Ed25519、ECDSA P-256、RFC 8785 JCS、TLS 1.3、Linux cgroup v2/pidfd/seccomp/nftables/SELinux、Docker Compose。

**Spec:** [Approved C1.2 node and POP control-plane design](../specs/2026-08-23-node-pop-control-plane-design.md), approved working-tree SHA-256 `B0B0DDBBD11546FC07A25CB76992E375481192A3261270FF442095501CC2B4B5`.

## Global Constraints

- 本计划只实现 C1.2。禁止加入 C1.3 账本、权益、配额租约，禁止加入 C1.4 调度、生产隧道凭据、客户端候选集或真实公网代理流量。
- `node-agent` 和 `node-core-supervisor` 的 production support matrix 只含 `linux/amd64`；Windows 只运行控制面开发入口与 Job Object portability fixture。
- `control-api` 继续是 modular monolith，但 public、bootstrap、agent、operator、metrics 五个 listener 使用独立 server、端口、认证和 resource cap。
- PostgreSQL 是唯一数据库权威；Redis 只能保存可重建的短期状态；NATS JetStream 仍是至少一次投递并按 `event_id` 去重。
- 所有授权、撤销、identity epoch、trust/root/metadata、desired/recovery activation、operator transition 与 authorizer record 都经过 `ControlPlaneAuthorityFence`。Observation 与纯读不消耗 sequence。
- Production fence、certificate issuer、NodeStateSigner、root-share provider、OperatorAuthorizer、OperatorClientTrustGuard、RollbackGuard、SecurityLatchGuard、TrustedTimeSource 和 host evidence verifier 都必须是外部 provider；local/test provider 名称必须显式含 `Deterministic` 或 `LocalTest`，不能进入 production profile。
- C1.1 `internal/trust`、C1.1 root schema、config signer key 与 device enrollment token domain 不得复用为 C1.2 node-state、node certificate 或 node grant authority。
- C1.2 canonical JSON 使用现有严格 JSON/JCS 基础，但每个 schema 使用自己的固定 domain-separated transcript；V1 不协商算法。
- Node、operator、server leaf 固定 ECDSA P-256、精确 SAN/EKU/KeyUsage/BasicConstraints/extension profile；TLS 仅 1.3 + `h2`，禁止 tickets、0-RTT、renegotiation和 `InsecureSkipVerify`。
- Bootstrap listener 不请求客户端证书；agent/operator listener 必须 `RequireAndVerifyClientCert`，并在每个 request 入场和 response commit 前重新执行 exact certificate row 与状态授权。
- Node desired/recovery/root/metadata、trust bundle、resource envelope 与 host memory policy 都执行 authority epoch/sequence、同流 version、digest、cumulative revoke 与 same-value fork 检查。
- Agent 的 TLS、snapshot validity 与 lease deadline 全部来自 production `TrustedTimeSource`；HTTP `Date`、signed time evidence、wall clock 和 RollbackGuard counter 都不能设置可信时间。
- Agent main RollbackGuard、supervisor RollbackGuard 与 LocalSecurityLatch 使用三个独立 counter/keystore identity；任何文件、MAC、数据库 sequence 或 wall clock 都不能冒充 production anti-rollback。
- Host-deployed `DeploymentAuthorityKeySetV1` 至少四把不同 Ed25519 key，角色精确为 `trust_bundle`、`host_remediation`、`operator_trust_guard`、`node_resource_envelope`，一把 key 不能占两个角色。
- `ApprovedReleaseManifestV1` 随 agent/supervisor build 固定；`InstalledReleaseMapV1` 只把 release ID 映射到不可写 absolute release root，不能覆盖 digest、argv、adapter、dependency closure 或 license。
- Desired state 最多 8 个 slot、canonical payload 最大 64 KiB；API 普通 body 最大 64 KiB，明确的 trust-conflict artifact endpoint 才允许 1 MiB/两个 artifact。
- Agent 永不接收 path、argv、environment、shell、原始 core config、production credential、下载 URL 或动态插件。Supervisor 本地协议也不接受这些字段。
- Xray 与 sing-box 始终是独立 OS 进程、独立上游制品和独立发布批次；不得 import、vendor、link 或复制进三个专有 release binary/production image。
- C1.2 core profile 仅是固定 loopback test profile；真实 ingress/egress/TUN/生产 credential 属于后续已批准规格。
- Supervisor 在 child create 前验证 resource envelope、aggregate reservation、manifest/map、memory policy、exact executable FD 与全部 finite cgroup/rlimit；任何 read-back 差异 fail closed。
- Production Linux V1 必须 SELinux enforcing，并要求 Yama 3、unprivileged BPF disabled、空 core pattern、suid dumpable 0、core_uses_pid 0 与 collector mask。Docker fake 不得声称证明 host policy。
- 所有 stdout/stderr、HTTP 错误、日志、metrics、outbox、observation 与 evidence 使用有限 enum/allowlist，禁止 grant、CSR、cert DER/serial、key、signed payload、core output、credential、path、IP 原文和 node ID metric label。
- 所有生成物提交仓库；`scripts/generate.*` 后 `api/`、`gen/`、`internal/store/` 必须无 drift。
- 每个实施任务遵循 RED → GREEN → REFACTOR：先运行 focused failing test，再做最小实现，再运行 focused、package 和受影响 integration test，最后独立提交。
- 普通 Go build/test 固定 `CGO_ENABLED=0`，race 固定 `CGO_ENABLED=1`；三个 release 固定 `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 -buildmode=exe`，禁止 external linker、`-linkshared` 和自定义 extldflags。
- PowerShell 与 repository-external Git-for-Windows Bash wrapper 必须调用同一个 canonical Docker verifier；Linux platform/provider 与 authority-fence 是另外两个作用域。
- 每个 authority evidence 最长 72 小时，绑定 repo commit、tracked-tree、approved spec、toolchain、三个 release binary、全部 image digest 与 cleanup digest。最终 manifest 要求四个作用域恰好各一且输入逐字相同。
- Test harness 只清理本 run 的 exact deterministic name/ID 与已验证临时根，不枚举或删除既有用户资源；主失败与 cleanup 失败分别保留 exit bit。
- 用户现有未跟踪目录 `.cache/`、`.superpowers/`、`.task19-go/` 不属于本计划，不得 stage、删除、扫描凭据或用作 test root。
- Redis 8.8.1 production license block、sing-box GPLv3+ gate 与 Xray MPL-2.0 gate 保持有效；进程隔离不是法律结论。

---

## Scope check and plan suite

设计第 21 节包含数据库权威、身份信任、节点 TCB、主机隔离、两个独立 core adapter 和四作用域验收，不能作为一个可审查提交序列实现。本索引冻结跨分册类型、文件责任与依赖；九个分册保留全部十一个实施批次：

| Batch | Plan | Independent exit |
| --- | --- | --- |
| `C1.2-B01` | [01 contracts, schema, authority](2026-08-23-node-pop-control-plane-01-contracts-schema-authority.md) | 三份 OpenAPI、nodecontrol events、完整 schema 与 deterministic fence crash/PITR fake 通过 |
| `C1.2-B02` | [02 operator state and signing](2026-08-23-node-pop-control-plane-02-operator-state-signing.md) | inventory/cursor/audit、desired/recovery signer saga、root threshold publisher 通过 |
| `C1.2-B03` | [03 node identity and mTLS](2026-08-23-node-pop-control-plane-03-node-identity-mtls.md) | enrollment/rotation/receipt、bounded recovery/remediation/two-person restore、trust package、operator guard 和三个 TLS listener 通过 |
| `C1.2-B04` | [04 node agent](2026-08-23-node-pop-control-plane-04-node-agent.md) | agent guards、signed pull/recovery、resource envelope 与 single-writer reconcile 通过 |
| `C1.2-B05` | [05 node-core supervisor](2026-08-23-node-pop-control-plane-05-node-core-supervisor.md) | strict socket、lease/seal、双 latch、Linux sandbox/aggregate reservation 通过 |
| `C1.2-B06–B07` | [06 fixture, health, load](2026-08-23-node-pop-control-plane-06-fixture-health-load.md) | controlled process、capacity reducer、privacy 与 1,000-agent reference load 通过 |
| `C1.2-B08–B09` | [07 Xray and sing-box adapters](2026-08-23-node-pop-control-plane-07-xray-singbox-adapters.md) | 两个独立 fixed-profile real loopback smoke 依次通过 |
| `C1.2-B10` | [08 verifier and supply chain](2026-08-23-node-pop-control-plane-08-verifier-supply-chain.md) | 双 wrapper、nested Docker、ownership WAL、scanner 与 scope evidence 通过 |
| `C1.2-B11` | [09 operations and completion](2026-08-23-node-pop-control-plane-09-operations-completion.md) | threat/runbooks、真实 Linux/operator guard、authority PITR 与四作用域 manifest 通过 |

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
├── c12-fixture/                                  # 独立 controlled process
├── c12-load/                                     # 1,000-agent load generator
└── talenro-artifact-scan/                        # versioned binary/OCI scanner
db/
├── migrations/00006_nodecontrol.sql              # 全部 C1.2 authority semantics
├── schema/nodecontrol.v1.yaml                    # 25-table exact catalog manifest; tests compare every object
└── queries/nodecontrol_{authority,inventory,identity,state,recovery,observation}.sql
internal/
├── nodecontrol/
│   ├── contracts/                                # pure canonical DTO/enums/validation
│   ├── authority/                                # fence provider/fake/repository/recovery
│   ├── inventory/                                # POP/node/slot/capacity/resource envelope
│   ├── operator/                                 # exact credential authorization/cursor/audit
│   ├── state/                                    # root/metadata/desired/recovery/signer workflows
│   ├── identity/                                 # grant/x509/issuance/receipt/certificate auth
│   ├── hostevidence/                             # trust package/guard/remediation verification
│   ├── recovery/                                 # incident/session/clear/resume/restore state machines
│   └── observation/                              # report/reducer/health transitions
├── nodebootstrapapi/                             # generated bootstrap adapter
├── nodeagentapi/                                 # generated agent adapter and waiter cancelation
├── nodeoperatorapi/                              # generated operator adapter
├── nodeagent/
│   ├── localstate/                               # RollbackGuard/latch/double-slot/time
│   ├── transport/                                # trust-guarded TLS claim/rotate/poll/report
│   ├── trust/                                    # independent C1.2 verifier
│   ├── reconcile/                                # single writer and LKG transition
│   └── adapter/{fixture,xray,singbox}/            # typed previews and probes
├── nodesupervisor/
│   ├── wire/                                     # strict canonical Protobuf framing + SCM_RIGHTS
│   ├── state/                                    # independent guard, lease and seals
│   ├── sandbox/                                  # Linux execution/isolation/read-back
│   └── adapter/                                  # deterministic config compiler
├── c12evidence/                                  # WAL, evidence schema and completion validator
└── artifactscan/                                 # Go deps, ELF/PE and OCI layout scanner
deploy/c12/                                       # audited outer Compose/images/policies
scripts/
├── verify-c12.ps1                                # fixed PowerShell wrapper
├── verify-c12.sh                                 # repository-external Git Bash wrapper
├── c12-authority.sh                              # canonical inner Bash verifier
├── verify-c12-platform.sh                        # Linux-native attested host gate
└── verify-c12-authority-fence.sh                 # provider/PITR gate
testdata/c12/                                     # locks, manifests, vectors and negative scanner fixtures
docs/security/c12-threat-model.md
docs/runbooks/c12-*.md
```

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

func (v AuthorityVersion) Validate() error
func CompareVersionedDigest(current, candidate VersionedDigest) (Comparison, error)
```

`Comparison` is a closed enum with only `same`, `advance`, `rollback`, and `fork`; callers must switch exhaustively.

The external authority boundary is created in `internal/nodecontrol/authority/provider.go`:

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

All domain mutations consume the Batch 01 coordinator rather than constructing provider database coordinates:

```go
type CoordinatorFinalizeRequest struct {
	OperationID  uuid.UUID
	EffectDigest contracts.Digest
}

func NewCoordinator(Provider, Repository, EffectResolver, securitykit.Clock) (*Coordinator, error)
func (c *Coordinator) Reserve(context.Context, ReserveRequest) (Reservation, error)
func (c *Coordinator) Finalize(context.Context, CoordinatorFinalizeRequest) (Receipt, error)
func (c *Coordinator) Abort(context.Context, AbortRequest) (Receipt, error)
func (c *Coordinator) Recover(context.Context, uuid.UUID) (Receipt, error)
func (c *Coordinator) CheckReady(context.Context) (Readiness, error)
```

Identity, state/root publishing, and recovery repositories expose read-only immutable effect resolvers. The composition root dispatches the finite `EffectKind` to exactly one resolver and builds one coordinator. Only that coordinator calls `CaptureDatabasePoint`, constructs provider `FinalizeRequest`, or activates/aborts an authority receipt; raw `Provider` remains available only to the coordinator, readiness, node-checkpoint time attestation, and the production authority gate.

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

package operator

type OperatorAuthorizer interface {
	Authorize(context.Context, Credential, Action, Target) (Authorization, error)
}
```

Issuer requests contain an immutable issuance ID, exact issuer ID, public key, server-constructed leaf template and template digest. Authorization credentials contain issuer ID, serial bytes, exact leaf DER/public-key digests, URI SAN and authority epoch; security-admin actions are never served from cache.

The host-remediation boundary is created in `internal/nodecontrol/hostevidence/remediation.go` and consumed only by the recovery application:

```go
package hostevidence

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

The verifier accepts only the exact role=`host_remediation` Ed25519 transcript and returns a defensive, digest-bound `VerifiedHostRemediation`; the recovery repository enforces 15-minute freshness, one-time use and exact node/incident/action binding.

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

No request includes executable path, argv, environment, shell, arbitrary config bytes, PID or signal. `CredentialFD` is valid only for exact test-profile prepare/probe roles; its zero value means no FD.

## Canonical artifact ownership and design traceability

These names are normative; implementations must not shorten, alias or redefine them in a second package:

| Canonical artifact or field | Owning file/task | Required consumers |
| --- | --- | --- |
| `contracts.Adapter`, `contracts.DesiredReasonV1`, `contracts.Nonce32`, `contracts.ProcessSpecV1`, `contracts.CapacityLimitsV1` | `internal/nodecontrol/contracts/process_spec.go`, Plan 02 Task 5 | desired/time signer, recovery attestation, agent preview/reconciler, supervisor compiler, Xray and sing-box adapters |
| `NodeTimeAttestationV1` | `internal/nodecontrol/state/contracts.go`, Plan 02 Tasks 5/8 | bootstrap and poll responses, agent trusted-time verifier and rollback guard |
| `NodeResourceEnvelopeV1`, `NodeResourceEnvelopePackageV1` | `internal/nodecontrol/contracts/resource_envelope.go`, Plan 04 Task B04-T04 | inventory pointer, agent trust verifier, supervisor reservation/sandbox |
| `HostMemoryIsolationPolicyV1`, `HostMemoryIsolationPolicyPackageV1` | `internal/nodecontrol/contracts/host_memory_policy.go`, Plan 04 Task B04-T04 | agent verifier, supervisor sandbox, Linux platform evidence |
| `RecoveryStateSnapshotV1` | `internal/nodecontrol/state/contracts.go`, Plan 02 Task 5 | recovery signer saga, agent recovery verifier/latch, API authorizers |
| `SecurityFaultReceiptV1` | `internal/nodecontrol/recovery/types.go`, Plan 03 Task 7 | agent recovery transport/latch and specialized receipt commit gate |
| `HostRemediationEvidenceV1`, `HostRemediationVerifier` | `internal/nodecontrol/hostevidence/remediation.go`, Plan 03 Task 7 | incident registration, reenrollment completion, typed clear, restore reauthorization |
| `LeaseRefreshSealV1`, `RollbackSealV1` | `internal/nodesupervisor/runtime/types.go`, Plan 05 Task B05-T06 | supervisor runtime only; agent may receive only typed result IDs |
| `C12ScopeEvidenceV1` | `internal/c12evidence/scope.go`, Plan 08 Task 1 | both Windows wrappers and both production-provider gates |
| `AuthorityFenceEvidenceV1`, `C12CompletionManifestV1` | `internal/c12evidence/authority.go` and `completion.go`, Plan 09 Tasks 6–7 | final four-scope validator |
| `authority.NodeCheckpoint` / `node_authority_checkpoint_sequence` | `internal/nodecontrol/authority/provider.go` and `repository.go`, Plan 01 Tasks 5–6; signed field in Plan 02 Tasks 5/8 | poll high-water and time attestation, desired/recovery audience, agent and supervisor rollback state |

Spec coverage remains explicit across the suite:

| Approved design sections | Plan ownership |
| --- | --- |
| §2–5 boundary, decision and architecture | this index plus Plan 01 Tasks 1–3 |
| §6 identity, rotation, application authorization and recovery | Plan 03 Tasks 1–7 |
| §7 signing, metadata and root threshold workflow | Plan 02 Tasks 5–9 |
| §8 isolated APIs, endpoint matrix, bounds and events | Plan 01 Tasks 2–3; Plan 03 Tasks 6, 8–10; Plan 06 Task B07-T03 |
| §9 PostgreSQL authority and PITR fence | Plan 01 Tasks 4–7; Plan 09 Task 6 |
| §10 inventory and two-level quarantine | Plan 02 Tasks 1–4; Plan 03 Task 7; Plans 04–06 |
| §11 desired/recovery snapshots | Plan 02 Tasks 5–8; Plans 04–05 |
| §12 local guards, trusted time and host policy | Plan 04 Tasks B04-T01–T04; Plan 09 Task 5 |
| §13 reconciler, supervisor and adapters | Plans 04–05; Plan 07 |
| §14–16 observation, health, capacity, privacy | Plan 06 Tasks B07-T01–T05 |
| §17 deterministic, integration, mTLS, process and load tests | every plan's exit gate; Plans 06–08 own the composed harness |
| §18 four-scope authority completion | Plans 08–09 |
| §19 threat model and runbooks | Plan 09 Tasks 1–3 |
| §20 license and supply chain | Plans 07–09 |

## Generation, verification, and commit discipline

Every task uses these exact local commands unless its plan specifies a narrower package first:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1
go test ./... -count=1 -timeout 10m
go test -race ./... -count=1 -timeout 10m
go vet ./...
go tool golangci-lint run ./...
```

Integration tasks first let the C1.2 verifier create an ephemeral migrated PostgreSQL, then run:

```text
go test -tags=integration ./... -count=1 -timeout 10m
```

Do not run tagged integration/E2E against fixed user services. Each plan names the owned harness entrypoint. A task commit stages only its explicit file list; never use `git add .`, `git add -A`, or a path that includes user-owned untracked directories.

## Completion semantics

Implementation completion is not inferred from unit tests or Docker alone. `C12CompletionManifestV1` must validate four unexpired, attested and build-identical scope digests:

1. `windows_powershell_docker`;
2. `windows_git_bash_docker`;
3. `linux_platform_operator_trust`;
4. `authority_fence_pitr`.

The first two may contain `container_deterministic` fake evidence but cannot satisfy the third or fourth scope. Only after all four pass, all licenses remain non-bypassed, cleanup/worktree equality succeeds, and the C1.2 runbooks are complete may Batch 11 change the roadmap from C1.2 `current design` to `complete` and C1.3 to `current design`.
