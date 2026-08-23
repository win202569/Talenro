# Talenro C1.2 Node Agent Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付只出站连接的 `node-agent`，使其用独立防回滚状态验证 C1.2 trust、desired/recovery state 与 host-deployed resource envelope，并通过单写入者 reconciler 安全驱动 typed supervisor client。

**Architecture:** Agent 以 `RollbackGuard`、独立 `LocalSecurityLatch` 和 `TrustedTimeSource` 为本地权威，严格验证 B01–B03 产出的 root、metadata、证书、trust bundle 与 signed state。Transport 只连接 bootstrap/agent mTLS listener；typed reconciler 只接收已验证 IR，并通过不含 path、argv、environment、PID 或任意 config 的 supervisor client surface 执行 transition。B04 只提供 provider-neutral 接口和名称明确的 deterministic fault-injection fake；真实 Linux TPM/secure-time/host-memory 证明属于分册 09。

**Tech Stack:** Go 1.26.5、ECDSA P-256 mTLS、Ed25519、RFC 8785 JCS、OpenAPI generated clients、Protobuf message types、Linux atomic rename/fsync、provider-neutral monotonic counter/sealer/secure time。

**Spec:** [Approved C1.2 node and POP control-plane design](../specs/2026-08-23-node-pop-control-plane-design.md), especially §§5.3, 8.2, 10.3, 11–13, 15–17; suite index [C1.2 implementation plan](2026-08-23-node-pop-control-plane.md).

## Global Constraints

- 本册只实现 `C1.2-B04`，并消费已完成的 B01–B03 contracts、authority、state、identity 和三个 generated API packages；不得加入 supervisor server、fixture/core adapter、容量 reducer、Xray、sing-box 或 C1.3/C1.4 功能。
- Production support matrix 只有 `linux/amd64`。Windows 代码只可用于后续 portability fixture，不能成为 production provider 或 production acceptance。
- Agent 不接受入站控制连接，不连接 PostgreSQL、Redis 或 NATS，不下载、安装、升级或修改 core。
- Agent main RollbackGuard 与 LocalSecurityLatch 必须使用不同 counter/sealer identity；构造时 identity 相同立即返回 `ErrProviderIdentityConflict`。
- Production `MonotonicCounter`、`AuthenticatedSealer` 和 `TrustedTimeSource` 必须来自外部 provider。单元测试只能使用精确名称 `DeterministicRollbackGuardFakeV1`、`DeterministicSecurityLatchGuardFakeV1` 和 `DeterministicTrustedTimeFakeV1`。
- Docker 或普通文件 fake 只能产生 `scope=container_deterministic` 的诊断/状态机证据；本册不得声明 TPM、跨 reboot secure time、SELinux、Yama、BPF、host dump policy 或真实 memory-attach gate 已通过。
- Agent 私钥、grant、certificate、credential、signed payload、raw provider error、path 和 core output不得进入日志、metric、observation 或 public error。
- Agent 的 TLS 验证通过 `tls.Config.Time` 使用 `TrustedTimeSource`；wall clock、HTTP `Date`、signed time attestation 与诊断 hint 都不能设置可信时间。
- Desired state 最多 8 个 slot，canonical payload 最大 64 KiB；普通 API wire/decoded body 最大 64 KiB，poll/rotate response 最大 1 MiB；client 只发送 `Accept-Encoding: identity` 并拒绝其他 content encoding。
- Desired/recovery/root/metadata 每条流分别比较 version/generation、authority sequence 与 digest；低值是 rollback，同值异 digest 是 security fault，同值同 digest 才幂等。
- `effective_authorization_deadline=min(issued_at+24h,metadata.valid_until,key.not_after)`；到期后立即 non-accepting，最多排空 10 分钟后强制停止，security/integrity fault 最多 5 秒强杀。
- Recovery snapshot 永远不能进入普通 running reconciler；它只能更新 recovery high-water、保持全部 slot stopped、处理精确 latch binding，并提交 recovery attestation。
- Agent 只向 supervisor 发送完整 signed trust/state、slot/generation、opaque lease/transition/fault ID 和 test-only typed credential FD；不得发送 path、argv、environment、shell、PID、signal 或任意 config bytes。
- 每个 task 必须执行 RED → GREEN → REFACTOR、focused test、受影响 package test和独立 commit。普通 test 固定 `CGO_ENABLED=0`；race gate固定 `CGO_ENABLED=1`。
- 禁止 stage 或清理用户的 `.cache/`、`.superpowers/`、`.task19-go/`；commit 必须逐项列出文件。

---

## Repository map

```text
cmd/node-agent/                         # composition root; no inbound server
internal/nodeagent/config/              # agent-only profile and bounded settings
internal/nodeagent/localstate/          # main guard, latch, secure-time interfaces/fakes
internal/nodeagent/trust/               # C1.2 root/metadata/state/package verification
internal/nodeagent/transport/           # claim/rotate/poll/report/recovery clients
internal/nodeagent/reconcile/           # one writer, typed plan, LKG and recovery coordinator
internal/nodesupervisor/wire/            # agent-facing frozen client/types; B05 adds codec/server
```

### Task B04-T01: Freeze the agent-facing supervisor contract and compatibility gate

**Files:**
- Create: `internal/nodesupervisor/wire/types.go`
- Create: `internal/nodesupervisor/wire/client.go`
- Create: `internal/nodesupervisor/wire/client_test.go`
- Create: `internal/nodeagent/reconcile/compatibility.go`
- Test: `internal/nodeagent/reconcile/compatibility_test.go`

**Interfaces:**
- Consumes: `contracts.Digest`, canonical `uuid.UUID`, and B02 signed root/metadata/desired/recovery envelope bytes.
- Produces: the suite-index `wire.Client` interface, closed request/response DTOs, `CredentialFD`, and `reconcile.VerifySupervisorCompatibility(Compatibility, wire.HandshakeResponse, wire.PeerIdentity) error`; B05 implements this client over the local socket.

- [ ] **Step 1: Write the failing closed-surface reflection test**

```go
func TestSupervisorRequestsContainOnlyClosedFields(t *testing.T) {
	requests := []any{
		wire.HandshakeRequest{}, wire.PrepareRequest{}, wire.CheckRequest{},
		wire.StartRequest{}, wire.ProbeRequest{}, wire.DrainRequest{},
		wire.StopRequest{}, wire.RenewRequest{}, wire.RollbackRequest{},
		wire.ListFaultsRequest{}, wire.ClearFaultRequest{},
	}
	for _, request := range requests {
		typeOf := reflect.TypeOf(request)
		for index := 0; index < typeOf.NumField(); index++ {
			name := strings.ToLower(typeOf.Field(index).Name)
			for _, forbidden := range []string{"path", "argv", "environment", "shell", "configbytes", "pid", "signal"} {
				if strings.Contains(name, forbidden) {
					t.Fatalf("%s contains forbidden field %s", typeOf.Name(), name)
				}
			}
		}
	}
}
```

- [ ] **Step 2: Run the contract test and verify RED**

Run: `go test ./internal/nodesupervisor/wire -run TestSupervisorRequestsContainOnlyClosedFields -count=1`

Expected: FAIL because `internal/nodesupervisor/wire` and its request types do not exist.

- [ ] **Step 3: Add the closed operation bindings and credential role**

```go
package wire

type OpaqueID [16]byte

type Binding struct {
	NodeID            [16]byte
	SlotID            string
	Generation        uint64
	AuthoritySequence uint64
	DesiredDigest     [32]byte
	SupervisorBootID  OpaqueID
}

type CredentialRole uint8

const (
	CredentialNone CredentialRole = iota
	CredentialServerTest
	CredentialClientTest
)

type CredentialFD struct {
	Descriptor int
	Role       CredentialRole
}
```

Define every request with only `Binding`, complete signed root/metadata/desired/recovery byte slices, opaque seal/lease/transition/fault IDs, booleans with one closed meaning, and the exact test credential role. `StopRequest` contains only `Binding`, `LeaseID`, and `AllOwned`; `AllOwned` never accepts a PID or pattern.

- [ ] **Step 4: Add the frozen client interface**

```go
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

Use sentinel errors `ErrProtocol`, `ErrPeerIdentity`, `ErrDeadline`, `ErrRejected`, and `ErrUnavailable`; none retains provider text or decoded request values.

- [ ] **Step 5: Write and run the compatibility failure matrix**

Test exact protocol version, supervisor build digest, root-owned peer UID, executable digest, peer PID/start token, and boot identity stability. Run: `go test ./internal/nodeagent/reconcile -run TestVerifySupervisorCompatibility -count=1`

Expected: FAIL because `VerifySupervisorCompatibility` is undefined.

- [ ] **Step 6: Implement the compatibility gate before any prepare call**

```go
type Compatibility struct {
	ProtocolVersion         string
	AllowedSupervisorBuilds map[[32]byte]struct{}
	ExpectedSupervisorUID   uint32
	ExpectedImageDigest     [32]byte
}

func VerifySupervisorCompatibility(expected Compatibility, got wire.HandshakeResponse, peer wire.PeerIdentity) error {
	if got.ProtocolVersion != expected.ProtocolVersion || peer.UID != expected.ExpectedSupervisorUID ||
		peer.ImageDigest != expected.ExpectedImageDigest || got.BootID != peer.BootID {
		return ErrSupervisorIntegrity
	}
	if _, ok := expected.AllowedSupervisorBuilds[got.BuildDigest]; !ok {
		return ErrSupervisorIntegrity
	}
	return nil
}
```

- [ ] **Step 7: Run focused and package tests**

Run: `go test ./internal/nodesupervisor/wire ./internal/nodeagent/reconcile -count=1`

Expected: PASS; reflection rejects forbidden fields and every compatibility mismatch returns only `ErrSupervisorIntegrity`.

- [ ] **Step 8: Commit the frozen client surface**

```bash
git add internal/nodesupervisor/wire/types.go internal/nodesupervisor/wire/client.go internal/nodesupervisor/wire/client_test.go internal/nodeagent/reconcile/compatibility.go internal/nodeagent/reconcile/compatibility_test.go
git commit -m "feat: freeze node supervisor client contract"
```

### Task B04-T02: Implement deterministic providers and the main rollback guard

**Files:**
- Create: `internal/nodeagent/localstate/providers.go`
- Create: `internal/nodeagent/localstate/deterministic_fakes.go`
- Create: `internal/nodeagent/localstate/main_state.go`
- Create: `internal/nodeagent/localstate/double_slot.go`
- Test: `internal/nodeagent/localstate/main_state_test.go`
- Test: `internal/nodeagent/localstate/double_slot_fuzz_test.go`

**Interfaces:**
- Consumes: suite-index `MonotonicCounter`, `AuthenticatedSealer`, `TrustedTimeSource`, `contracts.AuthorityVersion`, and `contracts.VersionedDigest`.
- Produces: `OpenMainGuard(Config) (*MainGuard, error)`, `(*MainGuard).Load(context.Context) (MainStateV1, error)`, and `(*MainGuard).Commit(context.Context, MainStateV1) error`; later transport/reconciler calls `Commit` only for listed security high-water changes.

- [ ] **Step 1: Write the failing provider-identity and crash-table tests**

```go
func TestMainGuardCrashTable(t *testing.T) {
	for _, point := range []CrashPoint{CrashAfterCandidateSync, CrashAfterCounterIncrement, CrashBeforePointerReplace} {
		t.Run(point.String(), func(t *testing.T) {
			fixture := NewDeterministicMainGuardFixture(t, point)
			if err := fixture.Guard.Commit(context.Background(), fixture.Next); !errors.Is(err, ErrInjectedCrash) {
				t.Fatalf("commit error = %v", err)
			}
			reopened := fixture.Reopen(t)
			got, err := reopened.Load(context.Background())
			fixture.AssertCrashOutcome(t, point, got, err)
		})
	}
}
```

- [ ] **Step 2: Run the local-state tests and verify RED**

Run: `go test ./internal/nodeagent/localstate -run 'Test(MainGuardCrashTable|ProviderIdentity)' -count=1`

Expected: FAIL because the provider interfaces, deterministic fixtures and guard do not exist.

- [ ] **Step 3: Define the exact provider interfaces and safe errors**

```go
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

Add only `ErrProviderUnavailable`, `ErrProviderIdentityConflict`, `ErrStateCorrupt`, `ErrStateRollback`, `ErrStateFork`, `ErrStateFuture`, and `ErrInjectedCrash` as exported classifications.

- [ ] **Step 4: Add deterministic fakes with exact names and scope**

```go
type DeterministicRollbackGuardFakeV1 struct {
	identity string
	value    uint64
	failAt   CrashPoint
}

type DeterministicTrustedTimeFakeV1 struct {
	identity string
	now      time.Time
	floor    time.Time
}

func (DeterministicTrustedTimeFakeV1) EvidenceScope() string {
	return "container_deterministic"
}
```

Fake constructors require nonempty run-scoped identities; production config rejects these concrete types in B04-T07.

- [ ] **Step 5: Define the complete main-state payload**

```go
type MainStateV1 struct {
	SchemaVersion             string
	StateEpoch                uint64
	CounterIdentity           string
	ControlPlaneAuthority     contracts.AuthorityVersion
	NodeAuthorityCheckpoint   uint64
	Root                      contracts.VersionedDigest
	Metadata                  contracts.VersionedDigest
	Desired                   contracts.VersionedDigest
	Recovery                  contracts.VersionedDigest
	IdentityEpoch             uint64
	CertificateDigest         contracts.Digest
	ServerBundle              contracts.VersionedDigest
	ApprovedManifestVersion   uint64
	InstalledMapVersion       uint64
	ResourceEnvelope          contracts.VersionedDigest
	HostMemoryPolicy          contracts.VersionedDigest
	SupervisorBuildDigest     contracts.Digest
	SupervisorProtocolVersion string
	LKGGeneration             uint64
	LKGDigest                 contracts.Digest
	RevokedKeyLedgerDigest    contracts.Digest
}
```

Encode with strict canonical JSON, cap the sealed plaintext at 64 KiB, and bind sealer purpose `talenro-node-agent-main-state-v1`.

- [ ] **Step 6: Implement the inactive-slot, fsync, counter, pointer sequence**

Implement `candidate write → file sync → directory sync → counter increment(expected+1) → pointer atomic replace`. Startup ignores the pointer for authority and accepts exactly one sealed slot whose `StateEpoch` equals the counter; a synced future slot is removable only when no security fault is embedded.

- [ ] **Step 7: Run crash, corruption and fuzz tests**

Run: `go test ./internal/nodeagent/localstate -run 'Test(MainGuard|ProviderIdentity)' -count=1`

Run: `go test ./internal/nodeagent/localstate -run=FuzzDoubleSlot -fuzz=FuzzDoubleSlot -fuzztime=10s`

Expected: PASS; old replay, same-epoch fork, truncation, trailing bytes, unknown schema and counter/blob mismatch all fail closed before returning an LKG.

- [ ] **Step 8: Run race and commit**

Run: `go test -race ./internal/nodeagent/localstate -count=1`

Expected: PASS with no race in concurrent read-versus-commit tests.

```bash
git add internal/nodeagent/localstate/providers.go internal/nodeagent/localstate/deterministic_fakes.go internal/nodeagent/localstate/main_state.go internal/nodeagent/localstate/double_slot.go internal/nodeagent/localstate/main_state_test.go internal/nodeagent/localstate/double_slot_fuzz_test.go
git commit -m "feat: add node agent rollback guard"
```

### Task B04-T03: Add the independent local security latch

**Files:**
- Create: `internal/nodeagent/localstate/latch.go`
- Create: `internal/nodeagent/localstate/latch_store.go`
- Test: `internal/nodeagent/localstate/latch_test.go`
- Test: `internal/nodeagent/localstate/latch_fuzz_test.go`
- Modify: `internal/nodeagent/localstate/providers.go`

**Interfaces:**
- Consumes: a counter/sealer identity distinct from B04-T02, canonical node/identity epoch, bounded fault subtype, `SecurityFaultReceiptV1`, and verified `RecoveryStateSnapshotV1`.
- Produces: `OpenSecurityLatch(Config) (*SecurityLatch, error)`, `RecordFault`, `BindReceipt`, `AuthorizeClear`, and `Snapshot`; B04-T06 uses any active/overflow result as an unconditional start prohibition.

- [ ] **Step 1: Write the failing 64/65-fault and clear-authorization tests**

```go
func TestSecurityLatchCapacityAndClear(t *testing.T) {
	latch := newLatchFixture(t)
	for index := 0; index < 64; index++ {
		fault := testFault(uint64(index + 1))
		if err := latch.RecordFault(context.Background(), fault); err != nil {
			t.Fatalf("record %d: %v", index, err)
		}
	}
	if err := latch.RecordFault(context.Background(), testFault(65)); err != nil {
		t.Fatal(err)
	}
	snapshot, err := latch.Snapshot(context.Background())
	if err != nil || len(snapshot.Active) != 64 || !snapshot.Overflow {
		t.Fatalf("snapshot = %#v, err = %v", snapshot, err)
	}
	if err := latch.AuthorizeClear(context.Background(), unsignedClear()); !errors.Is(err, ErrClearUnauthorized) {
		t.Fatalf("clear error = %v", err)
	}
}
```

- [ ] **Step 2: Run the latch test and verify RED**

Run: `go test ./internal/nodeagent/localstate -run TestSecurityLatchCapacityAndClear -count=1`

Expected: FAIL because `OpenSecurityLatch` and latch methods are undefined.

- [ ] **Step 3: Define the latch payload and mutation surface**

```go
type LocalFaultV1 struct {
	LocalFaultID   [16]byte
	Subtype        FaultSubtype
	EvidenceDigest contracts.Digest
	IdentityEpoch  uint64
	BootID         [16]byte
	IncidentID     [16]byte
}

type LocalSecurityLatchV1 struct {
	SchemaVersion          string
	LatchEpoch             uint64
	CounterIdentity        string
	NodeID                 [16]byte
	IdentityEpoch          uint64
	Active                 []LocalFaultV1
	Overflow               bool
	OverflowEvidenceDigest contracts.Digest
	LastClearAuthorization *ClearAuthorizationV1
}
```

`FaultSubtype` is a closed enum copied from B01 contracts; `Active` is sorted by `LocalFaultID`, unique and capped at 64.

- [ ] **Step 4: Implement record and receipt binding**

`RecordFault` is idempotent only when subtype/evidence/epoch/boot match. The 65th distinct fault sets overflow once and records only its first evidence digest. `BindReceipt` requires exact local fault ID, request digest, current identity epoch, finalized authority sequence and incident ID.

- [ ] **Step 5: Implement signed-clear consumption and tombstone retention**

`AuthorizeClear` requires `required_action=clear_security_latches`, exact recovery ID, exact snapshot digest, exact sorted fault bindings, non-null remediation evidence digest and `all_slots_stopped=true`. It clears only the listed exact faults, preserves `LastClearAuthorization`, and never changes desired/LKG fields.

- [ ] **Step 6: Run corruption, replay and independent-provider tests**

Run: `go test ./internal/nodeagent/localstate -run 'TestSecurityLatch|TestGuardIdentitySeparation' -count=1`

Expected: PASS; main-state corruption still permits recording `local_state_corruption_or_rollback` into the intact latch, while two unverifiable guards keep the agent stopped.

- [ ] **Step 7: Run fuzz and package tests**

Run: `go test ./internal/nodeagent/localstate -run=FuzzLatchStore -fuzz=FuzzLatchStore -fuzztime=10s`

Run: `go test ./internal/nodeagent/localstate -count=1`

Expected: PASS with no accepted duplicate field, reordered fault set, old latch epoch or same-epoch fork.

- [ ] **Step 8: Commit the latch**

```bash
git add internal/nodeagent/localstate/providers.go internal/nodeagent/localstate/latch.go internal/nodeagent/localstate/latch_store.go internal/nodeagent/localstate/latch_test.go internal/nodeagent/localstate/latch_fuzz_test.go
git commit -m "feat: add node security latch"
```

### Task B04-T04: Verify node-state trust, resource envelopes and deterministic memory policy

**Files:**
- Create: `internal/nodecontrol/contracts/resource_envelope.go`
- Test: `internal/nodecontrol/contracts/resource_envelope_test.go`
- Create: `internal/nodecontrol/contracts/host_memory_policy.go`
- Test: `internal/nodecontrol/contracts/host_memory_policy_test.go`
- Create: `internal/nodeagent/trust/types.go`
- Create: `internal/nodeagent/trust/verifier.go`
- Create: `internal/nodeagent/trust/resource_envelope.go`
- Create: `internal/nodeagent/trust/memory_policy.go`
- Create: `internal/nodeagent/trust/deterministic_memory_policy_fake.go`
- Test: `internal/nodeagent/trust/verifier_test.go`
- Test: `internal/nodeagent/trust/resource_envelope_test.go`
- Test: `internal/nodeagent/trust/memory_policy_test.go`
- Test: `internal/nodeagent/trust/fuzz_test.go`

**Interfaces:**
- Consumes: B01 closed `DesiredReasonV1` (including the exact value `lease_refresh`), B02 root/metadata/desired/recovery canonical bytes and signatures, B03 deployment key set/trust packages/current certificate identity, B04-T02 main high-water, and `TrustedTimeSource`.
- Produces: the sole canonical definitions of `contracts.NodeResourceEnvelopeV1`, `contracts.NodeResourceEnvelopePackageV1`, `contracts.HostMemoryIsolationPolicyV1`, and `contracts.HostMemoryIsolationPolicyPackageV1`; `Verifier.VerifyDesired`, `Verifier.VerifyRecovery`, `Verifier.VerifyResourceEnvelope`, `Verifier.VerifyMemoryPolicy`; immutable `VerifiedDesired`/`VerifiedRecovery`; and `EffectiveAuthorizationDeadline`. Agent trust code, Plan 05 supervisor code, and the Linux platform gate import these shared contracts and must not redefine wire-compatible shadows.

- [ ] **Step 1: Write the failing fork/audience/deadline table**

```go
func TestVerifyDesiredRejectsUnsafeBindings(t *testing.T) {
	for _, mutation := range []string{
		"node_id", "authority_epoch", "authority_sequence", "generation",
		"digest", "root", "metadata", "resource_envelope", "profile", "valid_until",
	} {
		t.Run(mutation, func(t *testing.T) {
			fixture := validDesiredFixture(t)
			fixture.Mutate(mutation)
			_, err := fixture.Verifier.VerifyDesired(context.Background(), fixture.Input)
			if !errors.Is(err, ErrTrustRejected) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}
```

- [ ] **Step 2: Run verifier tests and verify RED**

Run: `go test ./internal/nodeagent/trust -run TestVerifyDesiredRejectsUnsafeBindings -count=1`

Expected: FAIL because the verifier package does not exist.

- [ ] **Step 3: Define the canonical signed resource-envelope contract**

In `internal/nodecontrol/contracts/resource_envelope.go`, make these the only exported envelope wire types:

```go
package contracts

type ResourceLimitsV1 struct {
	CPUMillicores uint64 `json:"cpu_millicores"`
	MemoryBytes   uint64 `json:"memory_bytes"`
	TaskLimit     uint64 `json:"task_limit"`
	FDLimit       uint64 `json:"fd_limit"`
}

type NodeResourceEnvelopeV1 struct {
	SchemaVersion                   string           `json:"schema_version"`
	NodeID                          uuid.UUID        `json:"node_id"`
	ControlPlaneAuthorityEpoch      uint64           `json:"control_plane_authority_epoch"`
	AuthoritySequence               uint64           `json:"authority_sequence"`
	EnvelopeVersion                 uint64           `json:"envelope_version"`
	MaxSlots                        uint8            `json:"max_slots"`
	AgentLimits                     ResourceLimitsV1 `json:"agent_limits"`
	SupervisorLimits                ResourceLimitsV1 `json:"supervisor_limits"`
	CoreParentLimits                ResourceLimitsV1 `json:"core_parent_limits"`
	AggregateSlotFDReservationLimit uint64           `json:"aggregate_slot_fd_reservation_limit"`
	AggregateTmpfsBytes             uint64           `json:"aggregate_tmpfs_bytes"`
	AggregateTmpfsInodes            uint64           `json:"aggregate_tmpfs_inodes"`
	DetectedHostCapacityDigest      Digest           `json:"detected_host_capacity_digest"`
	IssuedAt                        time.Time         `json:"issued_at"`
}

type NodeResourceEnvelopePackageV1 struct {
	Envelope                     NodeResourceEnvelopeV1 `json:"envelope"`
	DeploymentKeyID              Digest                 `json:"deployment_key_id"`
	Algorithm                    string                 `json:"algorithm"`
	DeploymentAuthoritySignature [64]byte               `json:"deployment_authority_signature"`
}

const NodeResourceEnvelopeTranscriptV1 = "TALENRO-NODE-RESOURCE-ENVELOPE-V1\x00"
```

`schema_version` is exactly `node-resource-envelope.v1`, `algorithm` is exactly `ed25519`, UUID is canonical, every authority/version value is in `1..MaxInt64`, `max_slots` is exactly 8, every digest is nonzero, and `issued_at` is UTC with whole-second precision. Each `ResourceLimitsV1` uses CPU `100..64_000`, memory `67_108_864..1_099_511_627_776`, tasks `32..4_096`, and FDs `64..1_000_000`. Aggregate FD reservation is `64..9_000_000`, aggregate tmpfs bytes is `1..9_895_604_649_984`, and aggregate tmpfs inodes is `1..9_000_000`; checked sums must fit both the declared aggregate and `uint64`. Host-capacity validation requires agent + supervisor + core-parent + fixed OS headroom of at least one CPU/1 GiB/256 tasks to fit the trusted readback before any child operation.

Strict-decode the package with unknown/duplicate/trailing rejection, cap canonical JCS at 64 KiB, and verify only a B03 deployment key whose role is exactly `node_resource_envelope`. The signed bytes are `NodeResourceEnvelopeTranscriptV1 || JCS(package.envelope)`; package metadata is not part of the payload. Lower authority/version, same authority/version with another envelope digest, wrong node/key role, zero digest, overflow, or partial atomic install returns `ErrResourceEnvelope` and raises a security fault. The unit test freezes JSON field names, exact transcript bytes, all numeric edges, same-value fork behavior, and defensive-copy behavior.

- [ ] **Step 4: Define the canonical signed host-memory-policy contract**

In `internal/nodecontrol/contracts/host_memory_policy.go`, use closed nested types rather than an open `map[string]any`:

```go
package contracts

type HostMemorySysctlsV1 struct {
	YamaPtraceScope       uint8  `json:"kernel.yama.ptrace_scope"`
	UnprivilegedBPF       uint8  `json:"kernel.unprivileged_bpf_disabled"`
	SUIDDumpable          uint8  `json:"fs.suid_dumpable"`
	CorePattern           string `json:"kernel.core_pattern"`
	CoreUsesPID           uint8  `json:"kernel.core_uses_pid"`
}

type HostServiceDomainBindingV1 struct {
	ServiceUID uint32 `json:"service_uid"`
	Domain     string `json:"domain"`
}

type HostMemoryIsolationPolicyV1 struct {
	SchemaVersion                  string                       `json:"schema_version"`
	NodeID                         uuid.UUID                    `json:"node_id"`
	ControlPlaneAuthorityEpoch     uint64                       `json:"control_plane_authority_epoch"`
	AuthoritySequence              uint64                       `json:"authority_sequence"`
	PolicyVersion                  uint64                       `json:"policy_version"`
	KernelReleaseAndConfigDigest   Digest                       `json:"kernel_release_and_config_digest"`
	SELinuxBinaryPolicySHA256      Digest                       `json:"selinux_binary_policy_sha256"`
	SELinuxPolicyVersion           uint32                       `json:"selinux_policy_version"`
	RequiredEnforcing              bool                         `json:"required_enforcing"`
	AllowedServiceUIDToDomainMap   []HostServiceDomainBindingV1 `json:"allowed_service_uid_to_domain_map"`
	ClosedServiceDomainSet         []string                     `json:"closed_service_domain_set"`
	CoreDomainSet                  []string                     `json:"core_domain_set"`
	DenyUnknown                    bool                         `json:"deny_unknown"`
	Sysctls                        HostMemorySysctlsV1           `json:"sysctls"`
	CoredumpUnitMaskDigest         Digest                       `json:"coredump_unit_mask_digest"`
	IssuedAt                       time.Time                    `json:"issued_at"`
}

type HostMemoryIsolationPolicyPackageV1 struct {
	Policy                       HostMemoryIsolationPolicyV1 `json:"policy"`
	DeploymentKeyID              Digest                      `json:"deployment_key_id"`
	Algorithm                    string                      `json:"algorithm"`
	DeploymentAuthoritySignature [64]byte                    `json:"deployment_authority_signature"`
}

const HostMemoryIsolationPolicyTranscriptV1 = "TALENRO-HOST-MEMORY-ISOLATION-POLICY-V1\x00"
```

`schema_version` is exactly `host-memory-isolation-policy.v1`; authority/version values are in `1..MaxInt64`; `SELinuxPolicyVersion` is positive; all digests are nonzero; `RequiredEnforcing` and `DenyUnknown` are true; and `Algorithm` is `ed25519`. The service map has `1..32` entries sorted by strictly increasing nonzero UID, every mapped domain is in the closed set, both domain sets have `1..32` sorted unique SELinux type names matching `^[a-z][a-z0-9_]{0,126}_t$`, and the core set is a subset of the closed set. The five sysctls are exactly `{3,1,0,"",0}` in field order above. Canonical package size is at most 64 KiB and `IssuedAt` is UTC whole-second.

Only a B03 key with role exactly `host_remediation` verifies `HostMemoryIsolationPolicyTranscriptV1 || JCS(package.policy)`. The package validator rejects a trust-bundle key, lower policy/authority value, same-value different digest, wrong node, partial install, permissive SELinux, an open/unknown UID-domain mapping, a perf/BPF-capable domain, or collector-mask mismatch. Both agent and supervisor retain `(policy_version,authority_sequence,SHA256(JCS(policy)))` in their independent rollback state and directly read back the package plus all five sysctls, enforcing state, binary-policy digest, own domain, closed UID/domain map, and coredump unit mask before secrets, child prepare, and every lease renewal. Contract tests freeze field order-independent JCS, exact transcript bytes, role isolation, list bounds, domain grammar/subset rules, sysctl values, high-water fork rules, and signature mutation rejection.

- [ ] **Step 5: Define immutable verified outputs**

```go
type VerifiedDesired struct {
	NodeID                         [16]byte
	Generation                     uint64
	AuthoritySequence              uint64
	Digest                         contracts.Digest
	InventoryVersion               uint64
	ResourceEnvelope               contracts.VersionedDigest
	Processes                      []VerifiedProcessSpec
	EffectiveAuthorizationDeadline time.Time
}

type VerifiedRecovery struct {
	NodeID               [16]byte
	RecoveryID           [16]byte
	RecoveryGeneration   uint64
	AuthoritySequence    uint64
	Digest               contracts.Digest
	OpenIncidentIDs      [][16]byte
	RequiredAction       RecoveryAction
	AllSlotsStopped      bool
	RemediationDigest    contracts.Digest
}
```

Return defensive copies; constructors reject more than 8 processes, payloads above 64 KiB, noncanonical UUIDs, unsorted sets and any unknown enum.

- [ ] **Step 6: Implement trust ordering and time checks**

Verify continuous root chain, complete metadata, revoked-key cumulative ledger, JCS, domain-separated Ed25519 signature, current node/certificate audience, authority epoch/checkpoint, per-stream sequence/digest and minimum agent version. Compute the deadline using trusted time and reject `issued_at > trusted_now+5m`, expired metadata/key/state and any wall-clock fallback.

- [ ] **Step 7: Implement exact resource-envelope validation**

Parse `contracts.NodeResourceEnvelopePackageV1` and return an immutable verified view of its canonical `contracts.NodeResourceEnvelopeV1`; do not redeclare either shape in `internal/nodeagent/trust`. Validate role=`node_resource_envelope`, node ID, authority/version/digest monotonicity, `max_slots=8`, all agent/supervisor/core-parent limits, aggregate FD/tmpfs bounds and detected-host-capacity digest. Checked addition/multiplication overflow returns `ErrResourceEnvelope`; no default limit is substituted.

- [ ] **Step 8: Add deterministic host-memory policy evaluation**

```go
type DeterministicHostMemoryIsolationPolicyFakeV1 struct {
	Scope       string
	Readback    HostMemoryPolicyReadback
	PackageHash contracts.Digest
}

func NewDeterministicHostMemoryIsolationPolicyFakeV1(readback HostMemoryPolicyReadback) (*DeterministicHostMemoryIsolationPolicyFakeV1, error) {
	return &DeterministicHostMemoryIsolationPolicyFakeV1{Scope: "container_deterministic", Readback: readback}, nil
}
```

The fake accepts the canonical `contracts.HostMemoryIsolationPolicyPackageV1`, parses/verifies the test-key package with the same transcript and bounds, and simulates the five sysctls, enforcing state, SELinux policy/domain digest and collector mask. `internal/nodeagent/trust/memory_policy.go` consumes the shared contract and must not define another policy/package DTO. Its result type has no field that can claim `linux_platform_operator_trust`.

- [ ] **Step 9: Run contract, verifier, envelope, policy and fuzz tests**

Run: `go test ./internal/nodecontrol/contracts -run 'TestNodeResourceEnvelopeContract|TestHostMemoryIsolationPolicyContract' -count=1`

Run: `go test ./internal/nodeagent/trust -count=1`

Run: `go test ./internal/nodeagent/trust -run=FuzzSignedNodeState -fuzz=FuzzSignedNodeState -fuzztime=10s`

Expected: PASS; rollback/fork/profile mismatch and unsafe policy readback fail before a typed state is returned, and fake evidence remains `container_deterministic`.

- [ ] **Step 10: Commit the verifier boundary**

```bash
git add internal/nodecontrol/contracts/resource_envelope.go internal/nodecontrol/contracts/resource_envelope_test.go internal/nodecontrol/contracts/host_memory_policy.go internal/nodecontrol/contracts/host_memory_policy_test.go internal/nodeagent/trust/types.go internal/nodeagent/trust/verifier.go internal/nodeagent/trust/resource_envelope.go internal/nodeagent/trust/memory_policy.go internal/nodeagent/trust/deterministic_memory_policy_fake.go internal/nodeagent/trust/verifier_test.go internal/nodeagent/trust/resource_envelope_test.go internal/nodeagent/trust/memory_policy_test.go internal/nodeagent/trust/fuzz_test.go
git commit -m "feat: verify node agent trust state"
```

### Task B04-T05: Build bounded bootstrap, poll, recovery and report transport

**Files:**
- Create: `internal/nodeagent/transport/client.go`
- Create: `internal/nodeagent/transport/tls.go`
- Create: `internal/nodeagent/transport/poll.go`
- Create: `internal/nodeagent/transport/report.go`
- Test: `internal/nodeagent/transport/client_test.go`
- Test: `internal/nodeagent/transport/tls_test.go`
- Test: `internal/nodeagent/transport/poll_test.go`
- Test: `internal/nodeagent/transport/report_test.go`

**Interfaces:**
- Consumes: B03 generated `nodebootstrapv1`/`nodeagentv1` typed clients, host-deployed purpose-specific server trust bundle, current certificate slot, B04-T04 verifier and `TrustedTimeSource`.
- Produces: `transport.Client` claim/rotate/desired-poll/recovery-poll/report/security-fault/recovery-attestation methods, `LatestObservationQueue` with capacity one, and sanitized sentinel errors.

- [ ] **Step 1: Write the failing TLS and response-bound tests**

```go
func TestClientRejectsUntrustedResponseShapes(t *testing.T) {
	for _, test := range []struct {
		name     string
		encoding string
		bytes    int
	}{
		{name: "gzip", encoding: "gzip", bytes: 64},
		{name: "oversize", encoding: "identity", bytes: 1_048_577},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := newAgentTLSServer(t, test.encoding, test.bytes)
			client := newTestClient(t, server)
			_, err := client.PollDesired(context.Background(), validPollRequest())
			if !errors.Is(err, transport.ErrProtocol) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}
```

- [ ] **Step 2: Run transport tests and verify RED**

Run: `go test ./internal/nodeagent/transport -run 'TestClientRejectsUntrustedResponseShapes|TestTLSUsesTrustedTime' -count=1`

Expected: FAIL because client and trusted TLS configuration are absent.

- [ ] **Step 3: Implement the exact trusted TLS configuration**

Create separate bootstrap and agent transports with TLS 1.3, `h2`, disabled compression, disabled session tickets, exact DNS hostname and purpose-specific CA pool. Set `tls.Config.Time` from `TrustedTimeSource.Now`; reconnect on server-bundle change and cap connection lifetime at `min(server_leaf.NotAfter,connected_at+30m)`.

- [ ] **Step 4: Implement claim and rotation candidate slots**

Claim/rotate writes certificate, chain and `CertificateAuthorizationReceiptV1` only to a candidate slot; verify exact CSR/SPKI, URI SAN, leaf profile, chain, authority high-water, trusted time and fresh private-key challenge before calling the main guard and atomically switching the active slot. Any failure deletes the candidate and retains the old valid identity.

- [ ] **Step 5: Implement desired and recovery long-poll**

Use a 30-second request deadline around the server's 25-second wait. Send complete per-stream high-water tuples and `Accept-Encoding: identity`. Treat `204` as no state. A `200` desired response goes through `VerifyDesired`; a recovery response goes only through `VerifyRecovery` and never returns a `VerifiedDesired` value.

- [ ] **Step 6: Implement report queue and retry policy**

```go
type LatestObservationQueue struct {
	mu      sync.Mutex
	pending *Observation
}

func (queue *LatestObservationQueue) Replace(value Observation) {
	queue.mu.Lock()
	copyOfValue := value.Clone()
	queue.pending = &copyOfValue
	queue.mu.Unlock()
}
```

Retry starts at 1 second, caps at 30 seconds and uses crypto-random ±20% jitter. A new observation replaces the unsent value; no disk history is created.

- [ ] **Step 7: Prove report, security-fault and recovery separation**

Run: `go test ./internal/nodeagent/transport -run 'Test(Poll|Report|SecurityFault|Recovery)' -count=1`

Expected: PASS; ordinary report cannot use recovery credentials, security-fault returns only an exact finalized receipt, and recovery response cannot reach the desired callback.

- [ ] **Step 8: Run package tests and commit**

Run: `go test ./internal/nodeagent/transport -count=1`

```bash
git add internal/nodeagent/transport/client.go internal/nodeagent/transport/tls.go internal/nodeagent/transport/poll.go internal/nodeagent/transport/report.go internal/nodeagent/transport/client_test.go internal/nodeagent/transport/tls_test.go internal/nodeagent/transport/poll_test.go internal/nodeagent/transport/report_test.go
git commit -m "feat: add bounded node agent transport"
```

### Task B04-T06: Implement the single-writer reconciler and recovery coordinator

**Files:**
- Create: `internal/nodeagent/reconcile/types.go`
- Create: `internal/nodeagent/reconcile/planner.go`
- Create: `internal/nodeagent/reconcile/service.go`
- Create: `internal/nodeagent/reconcile/recovery.go`
- Create: `internal/nodeagent/reconcile/supervisor_adapter.go`
- Test: `internal/nodeagent/reconcile/planner_test.go`
- Test: `internal/nodeagent/reconcile/service_test.go`
- Test: `internal/nodeagent/reconcile/recovery_test.go`

**Interfaces:**
- Consumes: `trust.VerifiedDesired`, `trust.VerifiedRecovery`, B04-T01 `wire.Client`, B04-T02 `MainGuard`, B04-T03 `SecurityLatch`, and a no-side-effect `PreviewAdapter` registry.
- Produces: `reconcile.Service.SubmitDesired`, `SubmitRecovery`, `SupervisorFaultsChanged`, `Run`; persistent `AppliedState`; and finite `TransitionResult`/`SecurityFault` outputs used by B06/B07.

- [ ] **Step 1: Write the failing supersession and critical-section tests**

```go
func TestSingleWriterSupersedesOnlyCancelableWork(t *testing.T) {
	supervisor := newRecordingSupervisor()
	service := newReconcilerFixture(t, supervisor)
	service.SubmitDesired(desiredGeneration(1))
	supervisor.BlockAt(wire.OperationCheck)
	service.SubmitDesired(desiredGeneration(2))
	supervisor.Release(wire.OperationCheck)
	service.WaitIdle(t)
	if got := supervisor.Generations(); !slices.Equal(got, []uint64{1, 2}) {
		t.Fatalf("generations = %v", got)
	}
	if supervisor.MaxConcurrentMutations() != 1 {
		t.Fatal("more than one process mutation ran")
	}
}
```

- [ ] **Step 2: Run reconciler tests and verify RED**

Run: `go test ./internal/nodeagent/reconcile -run TestSingleWriterSupersedesOnlyCancelableWork -count=1`

Expected: FAIL because the reconciler service is undefined.

- [ ] **Step 3: Define the typed adapter and state interfaces**

```go
type PreviewAdapter interface {
	Adapter() AdapterKind
	ValidateProcess(trust.VerifiedProcessSpec) error
	Preview(trust.VerifiedProcessSpec) (PreviewDigest, error)
}

type AppliedState struct {
	HighestSeenGeneration uint64
	HighestSeenDigest     contracts.Digest
	AppliedGeneration     uint64
	AppliedDigest         contracts.Digest
	Slots                 []AppliedSlot
}
```

The registry contains only `fixture`, `xray`, and `sing_box`; B04 tests use a deterministic in-memory preview adapter and B06 adds the fixture implementation.

- [ ] **Step 4: Implement one writer and cancellation boundaries**

One goroutine owns prepare/check/start/rollback/stop and local applied-state mutation. A newer generation cancels preview/prepare/check and cleans staging; after stop/start/rollback begins, the service completes that critical section before planning the newest generation.

- [ ] **Step 5: Implement validation-first transition and lease refresh**

Validate all slots, resource sums, preview digests and supervisor compatibility before touching the old process. `lease_refresh` calls renew/rebind without restart only when lifecycle, process spec, release/profile/capacity/config semantic digest are byte-identical; any changed field enters a normal transition.

- [ ] **Step 6: Implement LKG and fail-closed boundaries**

Running-to-running candidate failure may consume only the supervisor-issued rollback seal within two minutes and before the prior effective deadline. `stopped`, `draining`, security fault, expired authorization, trust failure and latch failure never roll back to running. Highest-seen remains advanced when apply fails.

- [ ] **Step 7: Implement recovery-state coordination**

Recovery input stops all slots, seals recovery high-water, imports exact receipt/fault bindings into the latch and emits a recovery attestation. Only a verified `clear_security_latches` snapshot can call supervisor `ClearFault` and latch `AuthorizeClear`; even after local clear, no process starts until server finalization and a later ordinary resume generation.

- [ ] **Step 8: Run transition, expiry and recovery suites**

Run: `go test ./internal/nodeagent/reconcile -run 'Test(SingleWriter|Transition|LeaseRefresh|LKG|Recovery)' -count=1`

Expected: PASS; no canceled staging remains, recovery never invokes prepare/start, and security/expiry paths stop rather than restore LKG.

- [ ] **Step 9: Run race tests and commit**

Run: `go test -race ./internal/nodeagent/reconcile -count=1`

Expected: PASS with one mutation owner and no queue/report races.

```bash
git add internal/nodeagent/reconcile/types.go internal/nodeagent/reconcile/planner.go internal/nodeagent/reconcile/service.go internal/nodeagent/reconcile/recovery.go internal/nodeagent/reconcile/supervisor_adapter.go internal/nodeagent/reconcile/planner_test.go internal/nodeagent/reconcile/service_test.go internal/nodeagent/reconcile/recovery_test.go
git commit -m "feat: reconcile signed node state"
```

### Task B04-T07: Compose the fail-closed node-agent process

**Files:**
- Create: `internal/nodeagent/config/config.go`
- Test: `internal/nodeagent/config/config_test.go`
- Create: `cmd/node-agent/main.go`
- Create: `cmd/node-agent/run.go`
- Test: `cmd/node-agent/main_test.go`
- Test: `cmd/node-agent/run_test.go`
- Modify: `.env.example`

**Interfaces:**
- Consumes: B04-T02–T06 constructors, B03 enrollment material/provider configuration and one outbound supervisor socket path supplied only by trusted local deployment config.
- Produces: the `node-agent` executable, bounded shutdown, production-profile provider rejection, and a deterministic local/test composition used by B05/B06.

- [ ] **Step 1: Write the failing production-profile matrix**

```go
func TestProductionRejectsDeterministicProviders(t *testing.T) {
	for _, provider := range []string{
		"DeterministicRollbackGuardFakeV1",
		"DeterministicSecurityLatchGuardFakeV1",
		"DeterministicTrustedTimeFakeV1",
		"DeterministicHostMemoryIsolationPolicyFakeV1",
	} {
		lookup := validProductionLookup()
		lookup["TALENRO_NODE_AGENT_PLATFORM_PROVIDER"] = provider
		if _, err := config.Load(mapLookup(lookup)); !errors.Is(err, config.ErrUnsafeProvider) {
			t.Fatalf("provider %s error = %v", provider, err)
		}
	}
}
```

- [ ] **Step 2: Run config/composition tests and verify RED**

Run: `go test ./internal/nodeagent/config ./cmd/node-agent -run 'TestProductionRejectsDeterministicProviders|TestRunFailsClosed' -count=1`

Expected: FAIL because agent config and composition root are absent.

- [ ] **Step 3: Implement bounded configuration**

Require explicit profile, canonical node UUID, bootstrap/agent origins, exact server DNS names, protected identity/state/latch roots, supervisor socket, provider IDs, request deadlines, poll=25s, report=5s, lease-renew=20s and shutdown deadline. Production rejects local key bytes, deterministic provider names, non-HTTPS origins, non-linux/amd64 runtime and overlapping state/latch paths.

- [ ] **Step 4: Compose startup in fail-closed order**

Startup order is: parse config → instantiate external providers → verify distinct identities → read trusted time → verify file permissions/trust packages/memory-policy package → open latch → open main guard → load identity → handshake supervisor → build transport → start reconciler → start poll/report workers. Any failure before the last step starts no worker and requests supervisor `Stop(AllOwned=true)` only after exact peer verification.

- [ ] **Step 5: Implement bounded shutdown**

Cancel poll/report intake, stop accepting desired work, complete an active critical transition, request drain/stop according to the current effective deadline, flush only the latest observation within the shutdown deadline, close transport and clear in-memory credentials. Shutdown never prints raw errors or state paths.

- [ ] **Step 6: Add explicit local/test fake names to `.env.example`**

```dotenv
TALENRO_NODE_AGENT_PLATFORM_PROVIDER=DeterministicRollbackGuardFakeV1
TALENRO_NODE_AGENT_LATCH_PROVIDER=DeterministicSecurityLatchGuardFakeV1
TALENRO_NODE_AGENT_TIME_PROVIDER=DeterministicTrustedTimeFakeV1
TALENRO_NODE_AGENT_MEMORY_POLICY_PROVIDER=DeterministicHostMemoryIsolationPolicyFakeV1
```

Keep them under a comment stating local/test only and `scope=container_deterministic`; production startup rejects each value.

- [ ] **Step 7: Run focused and all B04 package tests**

Run: `go test ./internal/nodeagent/... ./internal/nodesupervisor/wire ./cmd/node-agent -count=1`

Expected: PASS; provider loss, corrupt latch/main state, invalid time and supervisor mismatch start no reconciler or transport worker.

- [ ] **Step 8: Build the fixed target and run dependency exclusion**

Run: `go build -trimpath -buildmode=exe -o .task19-go/node-agent-b04.exe ./cmd/node-agent`

Run: `go list -deps ./cmd/node-agent | rg 'github.com/(xtls/xray-core|sagernet/sing-box)'`

Expected: build succeeds; the dependency search exits 1 with no match. The output path is an existing user-owned ignored build cache and must not be staged or deleted by this task.

- [ ] **Step 9: Commit the node-agent composition**

```bash
git add internal/nodeagent/config/config.go internal/nodeagent/config/config_test.go cmd/node-agent/main.go cmd/node-agent/run.go cmd/node-agent/main_test.go cmd/node-agent/run_test.go .env.example
git commit -m "feat: compose fail-closed node agent"
```

## B04 exit gate

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1
go test ./internal/nodeagent/... ./internal/nodesupervisor/wire ./cmd/node-agent -count=1
go test -race ./internal/nodeagent/... ./internal/nodesupervisor/wire ./cmd/node-agent -count=1
go vet ./internal/nodeagent/... ./internal/nodesupervisor/wire ./cmd/node-agent
go tool golangci-lint run ./internal/nodeagent/... ./internal/nodesupervisor/wire ./cmd/node-agent
git diff --check
```

Expected: every command succeeds; generated trees are unchanged; deterministic evidence is labeled only `container_deterministic`; no test claims real Linux TPM, secure-time, SELinux, Yama/BPF or host memory isolation; `git status --short` contains no files outside B04 commits and pre-existing user-owned untracked directories.
