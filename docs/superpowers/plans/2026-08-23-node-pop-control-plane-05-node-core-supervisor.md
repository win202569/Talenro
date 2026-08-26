# Talenro C1.2 Node Core Supervisor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付 root-owned `node-core-supervisor`，通过严格本地协议、独立 rollback state、single-use seals、有限 lease 和 Linux sandbox，仅管理预装且摘要批准的外部 process。

**Architecture:** Supervisor 不持有 node certificate/private key，也不连接 control plane；它通过 Unix peer credentials、pidfd 和长度前缀 canonical Protobuf 接受一个 exact agent session。每次 prepare 独立验证完整 trust/state、manifest/map/resource envelope 和 memory-policy readback，随后以 supervisor-owned compiler 生成 sealed config，并在 child create 前设置/read-back aggregate reservation、cgroup、namespace、seccomp、nftables 和 rlimit。B05 的 Docker/Linux tests 只使用 deterministic host-memory fake；真实 attested Linux memory/platform gate留给分册 09。

**Tech Stack:** Go 1.26.5、Protobuf deterministic wire、Unix domain sockets/SCM_RIGHTS、SO_PEERCRED、pidfd、`execveat(AT_EMPTY_PATH)`、cgroup v2、Linux namespaces、seccomp、nftables、SELinux readback abstraction、RLIMIT。

**Spec:** [Approved C1.2 node and POP control-plane design](../specs/2026-08-23-node-pop-control-plane-design.md), especially §§5.3, 11.1, 12, 13 and 17.1/17.4; suite index [C1.2 implementation plan](2026-08-23-node-pop-control-plane.md).

## Global Constraints

- 本册只实现 `C1.2-B05`，消费完整 B04 agent/client contract；不得加入 fixture behavior、health reducer、real core adapter、Docker verifier、真实 Linux provider gate或 C1.3/C1.4 功能。
- Production target 只有 `linux/amd64`。非 Linux build只提供显式 `ErrUnsupportedPlatform` stub，不能声明 supervisor support。
- Supervisor 只能接受一个由 `SO_PEERCRED`、pidfd/start token 与 configured exact agent UID绑定的持久 session；socket backlog=1，每 slot最多一个 in-flight，全局最多8个。
- Local frame是最多1 MiB的4-byte big-endian length-prefixed deterministic Protobuf；unknown field、noncanonical encoding、oversize、truncated、trailing bytes和 `MSG_CTRUNC` 全部在 state advance 前拒绝。
- Supervisor request不得包含 executable path、argv、environment、shell、PID、signal、任意 config bytes、任意 file或production credential。
- Production supervisor independently strict-verifies the B02 canonical build-embedded `ApprovedReleaseManifestV1` and only the fixed `/etc/talenro/releases/installed-release-map.v1.json`; neither source/path/version/digest is accepted from the agent、caller、config or env. Missing/fixture/tampered bytes and any non-root-owned、writable or no-follow/immutable-root failure reject before the local socket or child creation.
- SCM_RIGHTS只允许 test-profile `prepare`/`probe` 各一个 sealed anonymous memfd，role精确为 server/client，大小1–4096 bytes；其他operation或production profile接收任何FD都拒绝并关闭全部FD。
- Supervisor RollbackGuard 使用独立于 agent main/latch 的 counter/sealer identity。Lease deadline、heartbeat和probe sample只在 boot-nonce-bound volatile table中，不写 counter。
- `prepare` 独立验证 root、metadata、desired、authority checkpoint、node audience、trusted time、manifest/map/resource envelope、memory policy 和 process spec；agent 的验证结果不是授权。
- Config只能由 supervisor-owned `AdapterCompilerV1` 从 signed process spec和build内profile生成；agent preview只用于digest比较，不能成为可启动config。
- Native check、running core和test smoke client是不可信 executable；child create 前必须设置并read-back finite CPU/memory/pids/FD/tmpfs reservation、`memory.swap.max=0`、`memory.oom.group=1`、namespace、seccomp、nftables、no-new-privs与hard `RLIMIT_CORE=0`。
- Production memory policy要求真实SELinux/Yama/BPF/core-pattern/collector证据，但本册的 test只能使用精确名称 `DeterministicHostMemoryIsolationPolicyFakeV1` 和 `scope=container_deterministic`；不得调用或模拟分册09的 attested gate输出。
- Supervisor restart不收养旧PID/lease；boot nonce改变使旧 config/lease/rollback seal失效，parent-death/cgroup cleanup必须先回收旧child。
- 普通lease最多延至 `min(effective_authorization_deadline,supervisor_trusted_now+60s)`；agent至少20秒renew。普通 expiry排空最多10分钟；security/integrity fault最多5秒强杀。
- Pending supervisor security fault必须先进入独立 rollback state，最多16个按 subtype聚合，额外fault设置一次 overflow；agent ACK不能自行清除它。
- 每个 task 必须执行 RED → GREEN → REFACTOR、focused/package/race test和独立 commit；不得 stage用户 `.cache/`、`.superpowers/`、`.task19-go/`。

---

## Repository map

```text
api/proto/talenro/nodesupervisor/v1/protocol.proto
gen/go/talenro/nodesupervisor/v1/protocol.pb.go
cmd/node-core-supervisor/
internal/nodesupervisor/
├── wire/       # framing, peer/session, FD and dispatch
├── state/      # rollback high-water, boot leases, security faults
├── releasecatalog/ # independent no-argument embedded-manifest/fixed-map loader
├── adapter/    # manifest/map/compiler and config seals
├── sandbox/    # resource reservation, cgroup, namespace, exec readback
└── runtime/    # operation state machine, lease/drain/rollback
```

### Task B05-T01: Generate the typed protocol and reject noncanonical local frames

**Files:**
- Create: `api/proto/talenro/nodesupervisor/v1/protocol.proto`
- Create generated: `gen/go/talenro/nodesupervisor/v1/protocol.pb.go`
- Modify: `internal/nodesupervisor/wire/types.go`
- Modify: `internal/nodesupervisor/wire/client.go`
- Create: `internal/nodesupervisor/wire/codec.go`
- Create: `internal/nodesupervisor/wire/server.go`
- Create: `internal/nodesupervisor/wire/peer_linux.go`
- Create: `internal/nodesupervisor/wire/fd_linux.go`
- Create (first line `//go:build !linux`): `internal/nodesupervisor/wire/peer_stub.go`
- Create (first line `//go:build !linux`): `internal/nodesupervisor/wire/fd_stub.go`
- Test: `internal/nodesupervisor/wire/codec_test.go`
- Test: `internal/nodesupervisor/wire/peer_linux_test.go`
- Test (first line `//go:build !linux`): `internal/nodesupervisor/wire/platform_stub_test.go`
- Test: `internal/nodesupervisor/wire/fuzz_test.go`

**Interfaces:**
- Consumes: B04 `wire.Client` DTOs and exact agent UID/build/protocol compatibility.
- Produces: generated `nodesupervisorv1.SupervisorRequestV1/SupervisorResponseV1`, `wire.Dial`, `wire.Serve`, strict codec, peer identity and credential-FD validation; B05-T06 supplies the operation handler.

- [ ] **Step 1: Write the failing frame and FD rejection table**

```go
func TestDecodeRejectsUnsafeFrames(t *testing.T) {
	for _, fixture := range []FrameFixture{
		OversizedFrame(1_048_577), TruncatedFrame(), TrailingFrame(),
		UnknownFieldFrame(), NonCanonicalFrame(), DuplicateOperationFrame(),
	} {
		if _, err := DecodeRequest(bytes.NewReader(fixture.Bytes)); !errors.Is(err, ErrProtocol) {
			t.Fatalf("%s error = %v", fixture.Name, err)
		}
	}
}
```

- [ ] **Step 2: Run the wire tests and verify RED**

Run: `go test ./internal/nodesupervisor/wire -run 'TestDecodeRejectsUnsafeFrames|TestCredentialFDMatrix' -count=1`

Expected: FAIL because the generated protocol and codec do not exist.

- [ ] **Step 3: Define the complete oneof protocol**

```proto
syntax = "proto3";
package talenro.nodesupervisor.v1;
option go_package = "talenro.local/platform/gen/go/talenro/nodesupervisor/v1;nodesupervisorv1";

message BindingV1 {
  bytes node_id = 1;
  string slot_id = 2;
  uint64 generation = 3;
  uint64 authority_sequence = 4;
  bytes desired_digest = 5;
  bytes supervisor_boot_id = 6;
}

message HandshakeRequestV1 { string protocol_version = 1; bytes agent_build_digest = 2; }
message PrepareRequestV1 { BindingV1 binding = 1; bytes signed_desired = 2; repeated bytes root_chain = 3; bytes metadata = 4; bytes resource_envelope = 5; bytes memory_policy = 6; }
message CheckRequestV1 { BindingV1 binding = 1; bytes config_seal_id = 2; }
message StartRequestV1 { BindingV1 binding = 1; bytes config_seal_id = 2; bytes transition_id = 3; }
message ProbeRequestV1 { BindingV1 binding = 1; bytes lease_id = 2; }
message DrainRequestV1 { BindingV1 binding = 1; bytes lease_id = 2; }
message StopRequestV1 { BindingV1 binding = 1; bytes lease_id = 2; bool all_owned = 3; }
message RenewRequestV1 { BindingV1 binding = 1; bytes lease_id = 2; bytes signed_desired = 3; repeated bytes root_chain = 4; bytes metadata = 5; }
message RollbackRequestV1 { BindingV1 binding = 1; bytes transition_id = 2; bytes rollback_seal_id = 3; }
message ListFaultsRequestV1 { bytes node_id = 1; bytes supervisor_boot_id = 2; }
message ClearFaultRequestV1 { bytes node_id = 1; bytes recovery_snapshot = 2; repeated bytes root_chain = 3; bytes metadata = 4; bytes remediation_evidence = 5; bytes local_latch_attestation = 6; }

message SupervisorRequestV1 {
  uint64 request_id = 1;
  oneof operation {
    HandshakeRequestV1 handshake = 10;
    PrepareRequestV1 prepare = 11;
    CheckRequestV1 check = 12;
    StartRequestV1 start = 13;
    ProbeRequestV1 probe = 14;
    DrainRequestV1 drain = 15;
    StopRequestV1 stop = 16;
    RenewRequestV1 renew = 17;
    RollbackRequestV1 rollback = 18;
    ListFaultsRequestV1 list_faults = 19;
    ClearFaultRequestV1 clear_fault = 20;
  }
}

enum SupervisorErrorCodeV1 {
  SUPERVISOR_ERROR_CODE_UNSPECIFIED = 0;
  SUPERVISOR_ERROR_CODE_PROTOCOL = 1;
  SUPERVISOR_ERROR_CODE_PEER_IDENTITY = 2;
  SUPERVISOR_ERROR_CODE_DEADLINE = 3;
  SUPERVISOR_ERROR_CODE_REJECTED = 4;
  SUPERVISOR_ERROR_CODE_UNAVAILABLE = 5;
}

enum ProbeHealthV1 {
  PROBE_HEALTH_UNSPECIFIED = 0;
  PROBE_HEALTH_HEALTHY = 1;
  PROBE_HEALTH_DEGRADED = 2;
  PROBE_HEALTH_UNHEALTHY = 3;
}

enum ProbeHandshakeV1 {
  PROBE_HANDSHAKE_UNSPECIFIED = 0;
  PROBE_HANDSHAKE_NOT_REQUIRED = 1;
  PROBE_HANDSHAKE_SUCCEEDED = 2;
  PROBE_HANDSHAKE_FAILED = 3;
}

enum RenewDispositionV1 {
  RENEW_DISPOSITION_UNSPECIFIED = 0;
  RENEW_DISPOSITION_SAME_GENERATION = 1;
  RENEW_DISPOSITION_LEASE_REFRESH_REBOUND = 2;
}

enum SupervisorFaultSubtypeV1 {
  SUPERVISOR_FAULT_SUBTYPE_UNSPECIFIED = 0;
  SUPERVISOR_FAULT_SUBTYPE_IDENTITY_COMPROMISE = 1;
  SUPERVISOR_FAULT_SUBTYPE_ONLINE_SIGNER_EQUIVOCATION = 2;
  SUPERVISOR_FAULT_SUBTYPE_METADATA_ROLLBACK = 3;
  SUPERVISOR_FAULT_SUBTYPE_ROOT_ROLLBACK = 4;
  SUPERVISOR_FAULT_SUBTYPE_ROOT_EQUIVOCATION = 5;
  SUPERVISOR_FAULT_SUBTYPE_UNVERIFIED_CLIENT_HIGHWATER_CONFLICT = 6;
  SUPERVISOR_FAULT_SUBTYPE_CLIENT_HIGHWATER_AHEAD = 7;
  SUPERVISOR_FAULT_SUBTYPE_SERVER_TRUST_BUNDLE_CONFLICT = 8;
  SUPERVISOR_FAULT_SUBTYPE_TRUSTED_TIME_ROLLBACK_OR_UNAVAILABLE = 9;
  SUPERVISOR_FAULT_SUBTYPE_LOCAL_STATE_CORRUPTION_OR_ROLLBACK = 10;
  SUPERVISOR_FAULT_SUBTYPE_RELEASE_OR_PROCESS_INTEGRITY = 11;
  SUPERVISOR_FAULT_SUBTYPE_PROFILE_BINDING_MISMATCH = 12;
  SUPERVISOR_FAULT_SUBTYPE_INCIDENT_OVERFLOW = 13;
}

message SupervisorErrorV1 {
  SupervisorErrorCodeV1 code = 1;
  uint32 retry_after_ms = 2;
}

message HandshakeResponseV1 {
  string protocol_version = 1;
  bytes supervisor_build_digest = 2;
  bytes supervisor_boot_id = 3;
}

message PrepareResponseV1 {
  BindingV1 binding = 1;
  bytes config_seal_id = 2;
  bytes config_semantic_digest = 3;
  uint64 effective_authorization_deadline_unix_ms = 4;
  bytes resource_envelope_digest = 5;
  bytes memory_policy_digest = 6;
}

message CheckResponseV1 {
  BindingV1 binding = 1;
  bytes config_seal_id = 2;
  bytes check_receipt_digest = 3;
  uint64 completed_at_unix_ms = 4;
}

message StartResponseV1 {
  BindingV1 binding = 1;
  bytes lease_id = 2;
  bytes rollback_seal_id = 3;
  uint64 lease_deadline_unix_ms = 4;
  bytes process_identity_digest = 5;
  uint64 started_at_unix_ms = 6;
}

message ProbeResponseV1 {
  BindingV1 binding = 1;
  bytes lease_id = 2;
  ProbeHealthV1 health = 3;
  ProbeHandshakeV1 handshake = 4;
  uint64 observed_at_unix_ms = 5;
  bytes observation_digest = 6;
}

message DrainResponseV1 {
  BindingV1 binding = 1;
  bytes lease_id = 2;
  uint64 drain_started_at_unix_ms = 3;
  uint64 hard_stop_at_unix_ms = 4;
}

message StopResponseV1 {
  BindingV1 binding = 1;
  bool all_owned = 2;
  bool owned_cgroups_empty = 3;
  uint64 completed_at_unix_ms = 4;
}

message RenewResponseV1 {
  BindingV1 binding = 1;
  bytes lease_id = 2;
  RenewDispositionV1 disposition = 3;
  bytes lease_refresh_seal_id = 4;
  uint64 lease_deadline_unix_ms = 5;
  bytes semantic_digest = 6;
}

message RollbackResponseV1 {
  BindingV1 restored_binding = 1;
  bytes lease_id = 2;
  bytes consumed_rollback_seal_id = 3;
  uint64 lease_deadline_unix_ms = 4;
  bytes process_identity_digest = 5;
  uint64 completed_at_unix_ms = 6;
}

message SupervisorFaultV1 {
  bytes supervisor_fault_id = 1;
  SupervisorFaultSubtypeV1 subtype = 2;
  bytes evidence_digest = 3;
  uint64 first_seen_at_unix_ms = 4;
  uint64 last_seen_at_unix_ms = 5;
  uint64 occurrence_count = 6;
}

message ListFaultsResponseV1 {
  bytes node_id = 1;
  bytes supervisor_boot_id = 2;
  repeated SupervisorFaultV1 faults = 3;
  bool overflow = 4;
  bytes overflow_digest = 5;
}

message ClearFaultResponseV1 {
  bytes node_id = 1;
  bytes recovery_snapshot_digest = 2;
  uint64 supervisor_rollback_state_counter = 3;
  bytes supervisor_rollback_state_digest = 4;
  repeated bytes cleared_supervisor_fault_ids = 5;
  bool owned_cgroups_empty = 6;
  uint64 completed_at_unix_ms = 7;
}

message SupervisorResponseV1 {
  uint64 request_id = 1;
  oneof outcome {
    SupervisorErrorV1 error = 2;
    HandshakeResponseV1 handshake = 10;
    PrepareResponseV1 prepare = 11;
    CheckResponseV1 check = 12;
    StartResponseV1 start = 13;
    ProbeResponseV1 probe = 14;
    DrainResponseV1 drain = 15;
    StopResponseV1 stop = 16;
    RenewResponseV1 renew = 17;
    RollbackResponseV1 rollback = 18;
    ListFaultsResponseV1 list_faults = 19;
    ClearFaultResponseV1 clear_fault = 20;
  }
}
```

Reserve field numbers after any removal; never add path/argv/environment/config/PID/signal fields. `request_id` is nonzero and exactly one outcome is present. A success outcome must match the request operation; an error is the only legal non-success outcome. Bytes representing node/boot/seal/lease/transition/fault IDs are exactly 16 bytes, digests are exactly 32 bytes, timestamps are positive Unix milliseconds no greater than `MaxInt64`, and `faults`/`cleared_supervisor_fault_ids` are sorted unique and capped at 16. `rollback_seal_id` is empty exactly when no running-to-running rollback capsule exists; `lease_refresh_seal_id` is exactly 16 bytes only for `RENEW_DISPOSITION_LEASE_REFRESH_REBOUND` and empty for same-generation renewal; `overflow_digest` is 32 bytes iff `overflow=true`.

`StopRequestV1`/B04 `wire.StopRequest` validation is one closed union. For `all_owned=false`, `lease_id` is exactly 16 nonzero bytes and every `BindingV1` field is present/canonical/nonzero: it must equal one current boot-lease row including configured node、slot、generation、authority sequence、desired digest and current supervisor boot ID. For `all_owned=true`, `lease_id` is empty and the binding is node-scoped: exact configured nonzero node ID and current supervisor boot ID, with empty slot and zero generation/authority sequence/desired digest. Partial/mixed forms、cross-node、old/future boot and complete lease bindings combined with `all_owned=true` return only `REJECTED` before any ownership lookup. After verified `SO_PEERCRED`/pidfd peer and configured node/current boot checks, `all_owned=true` may enumerate only the supervisor's current-boot ownership ledger and configured node service cgroup root；it never scans global processes/cgroups and never accepts PID、path、name、label or pattern.

`SupervisorErrorV1.code` is never `UNSPECIFIED`; `retry_after_ms` is zero except for `UNAVAILABLE`, where it is `1..60_000`. The wire layer maps the five codes one-to-one to B04 `ErrProtocol`, `ErrPeerIdentity`, `ErrDeadline`, `ErrRejected`, and `ErrUnavailable`; all internal trust, ordering, resource, integrity, and fault-pending failures collapse to `REJECTED`. No error or operation response contains a free-form result/reason, provider text, decoded input, stdout/stderr, executable identity beyond a digest, PID, path, argv, environment, config, signal, or credential. Contract tests enumerate all eleven request/response pairs, mutate every fixed-width field, reject the wrong success oneof for a request, and assert that generated descriptors contain neither an opaque generic payload nor a string result field.

- [ ] **Step 4: Generate and verify deterministic output**

Run the first generation, stage exactly its source/output baseline, run generation again, then compare worktree against that index baseline and reject any untracked file in the exact generated closure:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1
if ($LASTEXITCODE -ne 0) { throw 'first supervisor generation failed' }
git add -- api/proto/talenro/nodesupervisor/v1/protocol.proto gen/go/talenro/nodesupervisor/v1/protocol.pb.go
if ($LASTEXITCODE -ne 0) { throw 'failed to stage exact supervisor protocol generation baseline' }
$expectedSupervisorGenerated = @('api/proto/talenro/nodesupervisor/v1/protocol.proto','gen/go/talenro/nodesupervisor/v1/protocol.pb.go')
$stagedSupervisorGenerated = @(git diff --cached --name-only --diff-filter=ACMR -- api/proto/talenro/nodesupervisor/v1 gen/go/talenro/nodesupervisor/v1)
if (@(Compare-Object $expectedSupervisorGenerated $stagedSupervisorGenerated).Count -ne 0) { throw 'supervisor generated baseline is not the exact proto/pb.go pair' }
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1
if ($LASTEXITCODE -ne 0) { throw 'second supervisor generation failed' }
git diff --exit-code -- api/proto/talenro/nodesupervisor/v1/protocol.proto gen/go/talenro/nodesupervisor/v1/protocol.pb.go
if ($LASTEXITCODE -ne 0) { throw 'second supervisor generation differs from staged baseline' }
$untrackedSupervisorGenerated = @(git ls-files --others --exclude-standard -- api/proto/talenro/nodesupervisor/v1 gen/go/talenro/nodesupervisor/v1)
if ($LASTEXITCODE -ne 0 -or $untrackedSupervisorGenerated.Count -ne 0) { throw 'untracked supervisor generated artifact' }
```

Expected: the first run creates the reviewed proto/`pb.go` pair and puts exactly those two paths in the index；the second run has zero worktree-versus-index or untracked generated delta, after which Step 5 proceeds and Step 9 commits normally.

- [ ] **Step 5: Implement canonical framing**

Read exactly four length bytes, reject zero or values above 1 MiB before allocation, read exactly the declared payload, reject unknown fields, deterministic-remarshal and require byte equality. Encode responses with deterministic marshal and a four-byte big-endian size.

- [ ] **Step 6: Bind the one-session peer before dispatch**

Use `SO_PEERCRED` to require exact agent UID, open pidfd, record PID/start token/image digest, and recheck it before every mutating operation. Backlog is one; a second authenticated session is rejected while the first is alive.

`peer_stub.go` and `fd_stub.go` define the complete same exported constructor/method surface on every `!linux` target and return only `ErrUnsupportedPlatform` before opening a socket、accepting an FD or dispatching a request. `platform_stub_test.go` compile-assigns every Linux/stub method expression and proves each constructor returns that sentinel without side effects.

- [ ] **Step 7: Implement the exact SCM_RIGHTS matrix**

Accept one `CLOEXEC`, read-only, fully sealed anonymous memfd only for test-profile prepare/server or probe/client. Require 1–4096 bytes and exact run/slot/lease/schema role. Close every received FD on success, error, timeout and disconnect; any `MSG_CTRUNC` rejects the frame.

- [ ] **Step 8: Run focused, fuzz and race tests**

Run: `go test ./internal/nodesupervisor/wire -count=1`

Run: `go test ./internal/nodesupervisor/wire -run=FuzzDecodeRequest -fuzz=FuzzDecodeRequest -fuzztime=10s`

Run: `go test -race ./internal/nodesupervisor/wire -count=1`

Expected: PASS with no allocation above the cap, FD leak, second session or state dispatch from malformed input.

- [ ] **Step 9: Commit the protocol**

```bash
git add api/proto/talenro/nodesupervisor/v1/protocol.proto gen/go/talenro/nodesupervisor/v1/protocol.pb.go internal/nodesupervisor/wire/types.go internal/nodesupervisor/wire/client.go internal/nodesupervisor/wire/codec.go internal/nodesupervisor/wire/server.go internal/nodesupervisor/wire/peer_linux.go internal/nodesupervisor/wire/fd_linux.go internal/nodesupervisor/wire/peer_stub.go internal/nodesupervisor/wire/fd_stub.go internal/nodesupervisor/wire/codec_test.go internal/nodesupervisor/wire/peer_linux_test.go internal/nodesupervisor/wire/platform_stub_test.go internal/nodesupervisor/wire/fuzz_test.go
git commit -m "feat: add strict supervisor protocol"
```

### Task B05-T02: Persist supervisor high-water, boot leases and pending faults

**Files:**
- Create: `internal/nodesupervisor/state/providers.go`
- Create: `internal/nodesupervisor/state/deterministic_fake.go`
- Create: `internal/nodesupervisor/state/rollback.go`
- Create: `internal/nodesupervisor/state/boot_leases.go`
- Create: `internal/nodesupervisor/state/faults.go`
- Test: `internal/nodesupervisor/state/rollback_test.go`
- Test: `internal/nodesupervisor/state/boot_leases_test.go`
- Test: `internal/nodesupervisor/state/faults_test.go`

**Interfaces:**
- Consumes: a supervisor-only counter/sealer identity, verified state high-waters, independently derived B02 manifest/map `contracts.LocalVersionedDigestV1` tuples and supervisor boot nonce.
- Produces: `OpenRollbackGuard`, `BootLeaseTable`, `RecordSecurityFault`, `ListSecurityFaults`, and a persisted `SupervisorRollbackStateV1`; B05-T06 uses these as the sole seal/lease authority.

- [ ] **Step 1: Write the failing restart and pending-fault crash tests**

```go
func TestRestartNeverAdoptsOldBootLease(t *testing.T) {
	guard := newSupervisorGuard(t)
	oldBoot := fixedOpaqueID(1)
	leases := NewBootLeaseTable(oldBoot)
	leases.Put(testLease(oldBoot))
	newBoot := fixedOpaqueID(2)
	restarted := NewBootLeaseTable(newBoot)
	if restarted.Contains(testLease(oldBoot).LeaseID) {
		t.Fatal("old boot lease was adopted")
	}
	if err := guard.RequireOwnedCgroupsEmpty(context.Background()); err == nil {
		t.Fatal("restart did not require old cgroup cleanup")
	}
}
```

- [ ] **Step 2: Run state tests and verify RED**

Run: `go test ./internal/nodesupervisor/state -run 'TestRestartNeverAdoptsOldBootLease|TestPendingFaultCrashRecovery' -count=1`

Expected: FAIL because state package is absent.

- [ ] **Step 3: Define the persisted supervisor payload**

```go
type SupervisorRollbackStateV1 struct {
	SchemaVersion           string
	StateEpoch              uint64
	CounterIdentity         string
	ControlPlaneAuthority   contracts.AuthorityVersion
	NodeCheckpoint          uint64
	Root                    contracts.VersionedDigest
	Metadata                contracts.VersionedDigest
	HighestVerifiedDesired  contracts.VersionedDigest
	HighestAppliedDesired   contracts.VersionedDigest
	ApprovedManifest        contracts.LocalVersionedDigestV1
	InstalledMap            contracts.LocalVersionedDigestV1
	RevokedKeyLedgerDigest  contracts.Digest
	ResourceEnvelope        contracts.VersionedDigest
	HostMemoryPolicy        contracts.VersionedDigest
	PendingFaults           []SupervisorSecurityFaultV1
	FaultOverflow           bool
	FaultOverflowDigest     contracts.Digest
	TransitionCapsuleDigest contracts.Digest
}
```

Use the same fsync/counter/double-slot order as B04 but a separate purpose string and provider identity. The guard accepts manifest/map tuples only from B05-T03's opaque independently verified catalog, never as caller fields. On first startup or a verified local catalog advance it compares both with `CompareLocalVersionedDigest` and persists both atomically in one supervisor candidate/file+directory fsync/counter/pointer transaction before socket creation；later prepare may advance them only in the same transaction as root/metadata/desired and cumulative revoked-key-ledger digest. Same version with another digest、any version rollback、ledger disappearance/change without an authorized metadata advance, a half-persisted manifest/map pair or a restart slot whose tuple differs from the counter-selected state is a security fault and cannot reach prepare/renew/rollback.

- [ ] **Step 4: Add the deterministic supervisor guard fake**

Name the concrete test provider `DeterministicRollbackGuardFakeV1`; its `EvidenceScope()` returns only `container_deterministic`. It supports candidate-sync, counter-increment, pointer and `NV_RATE` injection without representing a TPM.

- [ ] **Step 5: Implement volatile boot leases**

`BootLeaseTableV1` binds lease ID, slot, generation/digest, PID/start token, image/cgroup identity, deadline and boot nonce. It has no serializer. Restart creates a new random boot nonce, kills/verifies empty old owned cgroups, and starts with an empty table.

- [ ] **Step 6: Implement bounded pending faults**

Aggregate by subtype, cap at 16, saturate occurrence count and set one overflow digest thereafter. A synced candidate containing a pending fault cannot be discarded as an ordinary future slot; startup finishes its exact counter commit or permits only stop/list-faults.

- [ ] **Step 7: Run crash, restart and wear tests**

Run: `go test ./internal/nodesupervisor/state -count=1`

Expected: PASS; renew/probe do not increment the counter, while high-water, transition capsule and fault create/clear do; old boot leases are never adopted.

- [ ] **Step 8: Commit supervisor state**

```bash
git add internal/nodesupervisor/state/providers.go internal/nodesupervisor/state/deterministic_fake.go internal/nodesupervisor/state/rollback.go internal/nodesupervisor/state/boot_leases.go internal/nodesupervisor/state/faults.go internal/nodesupervisor/state/rollback_test.go internal/nodesupervisor/state/boot_leases_test.go internal/nodesupervisor/state/faults_test.go
git commit -m "feat: persist supervisor rollback state"
```

### Task B05-T03: Validate release installation and seal compiled configuration

**Files:**
- Create: `internal/nodesupervisor/releasecatalog/catalog.go`
- Test: `internal/nodesupervisor/releasecatalog/catalog_test.go`
- Modify: `internal/nodesupervisor/state/rollback.go`
- Modify: `internal/nodesupervisor/state/rollback_test.go`
- Create: `internal/nodesupervisor/adapter/manifest.go`
- Create: `internal/nodesupervisor/adapter/installed_map.go`
- Create: `internal/nodesupervisor/adapter/types.go`
- Create: `internal/nodesupervisor/adapter/compiler.go`
- Create: `internal/nodesupervisor/adapter/seal.go`
- Test: `internal/nodesupervisor/adapter/manifest_test.go`
- Test: `internal/nodesupervisor/adapter/compiler_test.go`
- Test: `internal/nodesupervisor/adapter/seal_test.go`

**Interfaces:**
- Consumes: B04 verified canonical `contracts.ProcessSpecV1`; B02-owned canonical `ApprovedReleaseManifestV1`/`InstalledReleaseMapV1` types、strict validators and build-embedded approved-manifest source bytes；the fixed root-owned installed-map file；an exact role-bound sealed test-credential reader；and B05-T02 guard.
- Produces and uniquely owns: `releasecatalog.LoadProduction() (releasecatalog.Verified, error)` whose handle has no exported fields；`RollbackGuard.CompareAndCommitReleaseCatalog(context.Context, releasecatalog.Verified) error`；`ValidateInstallation(releasecatalog.Verified, contracts.ProcessSpecV1)` with no caller tuple constructor；`CredentialRole`、`RoleBoundCredentialReader`、`CompileRequestV1`、`CompiledConfigV1`、`CompileFunc`、`AdapterCompilerV1` and `RegisterCompiler(profileID string, CompileFunc) error` in `internal/nodesupervisor/adapter`; `ConfigSealV1`; one-shot `MarkChecked`/`ConsumeStart`; and secret-free preview digest comparison. The leaf `internal/nodesupervisor/releasecatalog` package imports only B01/B02 contracts and bounded OS readers；both `state` and `adapter` import it, avoiding an import cycle. Plan 07 only consumes the adapter surface. There is no positional compile overload or second process-spec/config/release DTO.

- [ ] **Step 1: Write failing source, path, digest and seal-replay tests**

```go
func TestConfigSealIsSingleUseAndBootBound(t *testing.T) {
	store := newSealStore(t, fixedOpaqueID(1))
	seal := store.Create(validCompileResult())
	if err := store.MarkChecked(seal.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ConsumeStart(seal.ID, fixedOpaqueID(1)); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ConsumeStart(seal.ID, fixedOpaqueID(1)); !errors.Is(err, ErrSealReplay) {
		t.Fatalf("replay error = %v", err)
	}
}
```

- [ ] **Step 2: Run adapter tests and verify RED**

Run: `go test ./internal/nodesupervisor/adapter -run 'Test(ConfigSeal|ValidateInstallation|Compiler|ReleaseCatalog)' -count=1`

Expected: FAIL because manifest/compiler/seal code is absent.

- [ ] **Step 3: Independently load, verify and persist the manifest/map intersection**

`internal/nodesupervisor/releasecatalog.LoadProduction()` takes no argument. It obtains a fresh defensive copy only from B02 `internal/localrelease`'s package-local build-embedded approved-manifest source；there is no runtime manifest path/provider or constructor accepting bytes/version/digest/decoded DTO. On every production startup it independently calls B02 `VerifyApprovedReleaseManifestV1` and then `RequireC12ProductionReleaseSet`, so absent/empty/tampered/noncanonical bytes and the fixture-only/partial embedded development instance reject. It opens only literal `/etc/talenro/releases/installed-release-map.v1.json` from a pre-opened UID-0 `/etc/talenro/releases` directory with `openat2(RESOLVE_BENEATH|RESOLVE_NO_SYMLINKS|RESOLVE_NO_MAGICLINKS|RESOLVE_NO_XDEV)` and `O_RDONLY|O_CLOEXEC|O_NOFOLLOW`; require regular file、one link、UID 0、no group/other write、size at most 16 KiB and identical pre/post-read device/inode/size/mode/owner, then independently call B02 `VerifyInstalledReleaseMapV1(mapBytes,verifiedApproved)`. Every mapped root/component below `/opt/talenro/releases/` and every manifest-pinned file/dependency is pinned through root-owned、service-unwritable、no-follow/no-junction/no-reparse/no-mount-escape handles; no current directory、`PATH` or caller path participates.

Only after all independent filesystem checks succeed does the loader retain the two B02 immutable verified handles and permit its private commit adapter to call their sole tuple accessors. Those accessors derive the exact domain-separated manifest/map `contracts.LocalVersionedDigestV1` values frozen by B02；P05 never recomputes a raw JCS digest. The loader returns one defensive `releasecatalog.Verified` whose fields are unexported. No other production constructor exists, and no exported API accepts either tuple、path、bytes、filesystem-success boolean or caller-decoded DTO. `RollbackGuard.CompareAndCommitReleaseCatalog` accepts only that handle, extracts/compares both tuples internally and atomically commits a first/advanced pair before use. `ValidateInstallation` consumes the same handle and requires exact adapter/release ID, immutable absolute release root, exact executable/dependency SHA-256, no setuid/setgid/file capability and exact loader closure. Installed map cannot override digest、argv、adapter、dependency closure、sandbox、provenance or license. Production has no fixture/test byte-source injection seam and rejects absent/empty source、fixture-only/partial embedded set、unknown/duplicate/trailing/noncanonical fields、embedded/map tamper、rollback/fork、permission/root mutation and every pre/post-open substitution before compiler/reservation/child calls. Tests `TestReleaseCatalogHasNoCallerTuple`、`TestReleaseCatalogRejectsRawJCSDigestSubstitution`、`TestProductionReleaseCatalogRejectsAbsentFixtureAndTamper`、`TestReleaseCatalogImmutableRoots` and `TestReleaseCatalogHighWaterCommitsAtomically` mutate both sources and every file identity boundary, substitute a consumer-computed raw-JCS SHA-256 for each B02 domain-separated tuple and prove the substitution cannot enter the opaque handle or persisted guard high-water.

- [ ] **Step 4: Implement the deterministic compiler**

```go
type CredentialRole string

const (
	CredentialRoleServer CredentialRole = "server"
	CredentialRoleClient CredentialRole = "client"
)

type RoleBoundCredentialReader interface {
	io.Reader
	Role() CredentialRole
	DeclaredSize() uint16
}

type CompileRequestV1 struct {
	NodeID            uuid.UUID
	Generation        uint64
	SlotID            string
	ProcessSpec       contracts.ProcessSpecV1
	ProcessSpecDigest contracts.Digest
	ReleaseID         string
	ProfileID         string
	Credential        RoleBoundCredentialReader
}

type CompiledConfigV1 struct {
	CanonicalJSON []byte
	ConfigDigest  contracts.Digest
	Listen        netip.AddrPort
	Target        netip.AddrPort
}

type AdapterCompilerV1 interface {
	Compile(CompileRequestV1) (CompiledConfigV1, error)
}

type CompileFunc func(CompileRequestV1) (CompiledConfigV1, error)

func RegisterCompiler(profileID string, compile CompileFunc) error
```

`ProcessSpecDigest` must equal SHA-256 of the canonical B04-verified `ProcessSpec`, and node/generation/slot/release/profile must equal that same signed desired binding and the selected registered profile. `Credential` is nil for production and for profiles that do not require test credentials. Otherwise its sealed FD implementation exposes only the fixed role and declared size `1..4096`; compiler checks the registered template's expected role plus run/node/slot/lease credential header before the first payload read and reads exactly `DeclaredSize()` bytes once. The role cannot be supplied as an independent string that disagrees with the reader. `RegisterCompiler` accepts one non-nil function for one canonical profile ID exactly once；duplicate、alias、case-folded、nil and post-start registration fail without replacing an existing function. Compiler selects only build-registered fixture/xray/sing-box template IDs, returns canonical JSON capped at 64 KiB, recomputes `ConfigDigest`, and rejects unknown JSON fields, trailing credential bytes, short reads, reader reuse, preview semantic mismatch, or any request mutation. `CompiledConfigV1.CanonicalJSON` remains supervisor-owned and never crosses the wire; `Listen` and `Target` must be exact registered loopback endpoints. An external-package compile test assigns every type/method/function expression and proves P06/P07 cannot redeclare a wire-compatible shadow.

- [ ] **Step 5: Seal generation-isolated config**

Write compiler output to a supervisor-owned generation directory, file-sync, directory-sync, atomically seal it read-only for the target core UID, and record `ConfigSealV1={node,generation,desired_digest,slot,process_spec_digest,compiler_version,release_id,config_digest,boot_nonce,lease_nonce}`. Agent preview digest mismatch is a security fault.

- [ ] **Step 6: Enforce check-before-start and one-shot use**

Only native-check success plus verified empty check cgroup can mark a seal checked. Start requires same boot/slot/generation/digest and consumes it once. Disconnect, timeout, security fault and later generation invalidate unused seals.

- [ ] **Step 7: Run package and race tests**

Run: `go test ./internal/nodesupervisor/adapter -count=1`

Run: `go test -race ./internal/nodesupervisor/adapter -count=1`

Expected: PASS; catalog absence/fixture/tamper/rollback/fork、path substitution, dependency swap, preview mismatch, cross-slot/boot/generation use and replay all fail closed, and only an atomic independently verified manifest/map tuple pair reaches compiler state.

- [ ] **Step 8: Commit compiler and seals**

```bash
git add internal/nodesupervisor/releasecatalog/catalog.go internal/nodesupervisor/releasecatalog/catalog_test.go internal/nodesupervisor/state/rollback.go internal/nodesupervisor/state/rollback_test.go internal/nodesupervisor/adapter/manifest.go internal/nodesupervisor/adapter/installed_map.go internal/nodesupervisor/adapter/types.go internal/nodesupervisor/adapter/compiler.go internal/nodesupervisor/adapter/seal.go internal/nodesupervisor/adapter/manifest_test.go internal/nodesupervisor/adapter/compiler_test.go internal/nodesupervisor/adapter/seal_test.go
git commit -m "feat: seal supervisor core configuration"
```

### Task B05-T04: Enforce aggregate resource reservation before child creation

**Files:**
- Create: `internal/nodesupervisor/sandbox/limits.go`
- Create: `internal/nodesupervisor/sandbox/reservation.go`
- Create: `internal/nodesupervisor/sandbox/cgroup_linux.go`
- Create: `internal/nodesupervisor/sandbox/rlimit_linux.go`
- Create (first line `//go:build !linux`): `internal/nodesupervisor/sandbox/platform_stub.go`
- Test: `internal/nodesupervisor/sandbox/limits_test.go`
- Test: `internal/nodesupervisor/sandbox/reservation_test.go`
- Test: `internal/nodesupervisor/sandbox/cgroup_linux_test.go`

**Interfaces:**
- Consumes: Plan 02 Task 5 canonical `contracts.NodeResourceEnvelopeV1` through the independently verified `contracts.NodeResourceEnvelopePackageV1`, immutable capacity profiles, at most 8 running/draining/candidate slots and the one allowed ephemeral check/probe reservation. This package must not define a second envelope DTO.
- Produces: `ReservationLedger.Acquire/Release`, `PrepareCoreParent`, `PrepareSlotCgroup`, and exact readback receipts consumed by B05-T05/T06.

- [ ] **Step 1: Write failing aggregate and overflow tests**

```go
func TestReservationRejectsIndividuallyValidAggregate(t *testing.T) {
	ledger := NewReservationLedger(testEnvelope())
	for index := 0; index < 7; index++ {
		if _, err := ledger.Acquire(slotReservation(index)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := ledger.Acquire(slotReservation(7)); !errors.Is(err, ErrAggregateLimit) {
		t.Fatalf("eighth reservation error = %v", err)
	}
}
```

- [ ] **Step 2: Run reservation tests and verify RED**

Run: `go test ./internal/nodesupervisor/sandbox -run 'TestReservation|TestCheckedLimitArithmetic' -count=1`

Expected: FAIL because the sandbox resource ledger is absent.

- [ ] **Step 3: Define exact finite limits and checked arithmetic**

Represent CPU millicores, memory bytes, tasks, per-process FD, tmpfs bytes/inodes and slot count with unsigned integers. Reject zero, spec bounds violations, `max`, unsupported required metrics, addition/multiplication overflow and any limit absent from the signed envelope/profile.

- [ ] **Step 4: Implement one aggregate reservation lock**

Account for all running/draining slots, current candidate and exactly one ephemeral check or probe client. Acquire before child creation; release only after pidfd closure and verified empty cgroup/config view. FD uses worst-case per-process reservations because cgroup has no FD controller.

- [ ] **Step 5: Prepare and read back parent/slot cgroups**

Set parent and child `cpu.max`, `memory.max`, `memory.swap.max=0`, `memory.oom.group=1`, `pids.max`; set exact hard/soft `RLIMIT_NOFILE` and `RLIMIT_CORE=0`. Any controller missing, value `max`, late application or readback mismatch returns `ErrIsolationReadback` before exec.

- [ ] **Step 6: Add explicit non-Linux failure**

The `!linux` stub returns `ErrUnsupportedPlatform` from every child/resource constructor. It does not emulate cgroup or allow a production profile.

- [ ] **Step 7: Run boundary and race tests**

Run: `go test ./internal/nodesupervisor/sandbox -run 'Test(Reservation|Checked|Cgroup|Rlimit)' -count=1`

Run: `go test -race ./internal/nodesupervisor/sandbox -run TestReservation -count=1`

Expected: PASS; old+candidate+ephemeral and eight-slot sums never oversubscribe or leak reservations.

- [ ] **Step 8: Commit resource enforcement**

```bash
git add internal/nodesupervisor/sandbox/limits.go internal/nodesupervisor/sandbox/reservation.go internal/nodesupervisor/sandbox/cgroup_linux.go internal/nodesupervisor/sandbox/rlimit_linux.go internal/nodesupervisor/sandbox/platform_stub.go internal/nodesupervisor/sandbox/limits_test.go internal/nodesupervisor/sandbox/reservation_test.go internal/nodesupervisor/sandbox/cgroup_linux_test.go
git commit -m "feat: enforce supervisor resource envelopes"
```

### Task B05-T05: Create the Linux sandbox and pinned process launcher

**Files:**
- Create: `internal/nodesupervisor/sandbox/memory_policy.go`
- Create: `internal/nodesupervisor/sandbox/deterministic_memory_policy_fake.go`
- Create: `internal/nodesupervisor/sandbox/namespaces_linux.go`
- Create: `internal/nodesupervisor/sandbox/seccomp_linux.go`
- Create: `internal/nodesupervisor/sandbox/network_linux.go`
- Create: `internal/nodesupervisor/sandbox/exec_linux.go`
- Create: `internal/nodesupervisor/sandbox/ownership_linux.go`
- Create (first line `//go:build !linux`): `internal/nodesupervisor/sandbox/exec_stub.go`
- Test: `internal/nodesupervisor/sandbox/memory_policy_test.go`
- Test: `internal/nodesupervisor/sandbox/exec_linux_test.go`
- Test: `internal/nodesupervisor/sandbox/network_linux_test.go`
- Test (first line `//go:build !linux`): `internal/nodesupervisor/sandbox/exec_stub_test.go`

**Interfaces:**
- Consumes: B05-T03 validated executable/config handles, B05-T04 reservation/cgroup receipt, Plan 04 B04-T04 canonical `contracts.HostMemoryIsolationPolicyPackageV1`, and exact target core UID/SELinux domain. Supervisor validation imports the shared policy/package types and transcript constant; it must not define a wire-compatible shadow.
- Produces: `Launcher.Prepare`, `Launcher.Start`, `OwnedProcess`, `VerifyPostStart`, and `KillOwnedCgroup`; B05-T06 never signals a raw PID.

- [ ] **Step 1: Write failing unsafe-policy and pre-exec ordering tests**

```go
func TestUnsafeMemoryPolicyStopsBeforeChildCreate(t *testing.T) {
	for _, mutation := range deterministicUnsafePolicyMutations() {
		launcher := newLauncherWithMemoryPolicy(t, mutation)
		_, err := launcher.Prepare(context.Background(), validLaunchRequest())
		if !errors.Is(err, ErrHostMemoryPolicy) {
			t.Fatalf("%s error = %v", mutation.Name, err)
		}
		if launcher.ChildCreateCount() != 0 {
			t.Fatal("unsafe policy reached child creation")
		}
	}
}
```

- [ ] **Step 2: Run sandbox ordering tests and verify RED**

Run: `go test ./internal/nodesupervisor/sandbox -run 'TestUnsafeMemoryPolicyStopsBeforeChildCreate|TestPinnedExecOrdering' -count=1`

Expected: FAIL because launcher and deterministic policy evaluator are absent.

- [ ] **Step 3: Add only the deterministic memory-policy fake for this plan**

```go
type DeterministicHostMemoryIsolationPolicyFakeV1 struct {
	Scope    string
	Readback MemoryPolicyReadback
}

func (fake DeterministicHostMemoryIsolationPolicyFakeV1) Verify(context.Context, contracts.HostMemoryIsolationPolicyPackageV1) (MemoryPolicyReceipt, error) {
	if fake.Scope != "container_deterministic" {
		return MemoryPolicyReceipt{}, ErrHostMemoryPolicy
	}
	return verifyDeterministicReadback(fake.Readback)
}
```

The fake exercises the canonical Plan 04 parser, `HostMemoryIsolationPolicyTranscriptV1`, role=`host_remediation` signature and high-water rules before substituting deterministic readback. Tests cover five sysctls, SELinux enforcing/policy/domain map, perf/BPF deny set and collector mask as parsed values only. This type cannot emit real platform evidence. Production verifies that same canonical package and direct readback before child creation and every lease renewal.

- [ ] **Step 4: Implement private mount/PID/network namespaces**

Expose only approved release/dependencies, sealed config, minimal read-only hidepid proc, `/dev/null`, `/dev/urandom`, and optional private tmpfs capped at 16 MiB/1,024 inodes. No host `/etc`, `/tmp`, agent state, supervisor socket, other slot or host namespace FD is visible.

- [ ] **Step 5: Implement seccomp and nftables policy**

Reject fork/vfork/clone3, non-exact thread clone flags, helper exec, mount/pivot_root/unshare/setns, ptrace/process-vm/pidfd-getfd/perf/BPF, rlimit increase, raw/netlink sockets and non-allowlisted connect/listen. Network namespace contains loopback only, no veth/default route, and exact manifest loopback endpoints.

- [ ] **Step 6: Launch from a pinned executable FD**

After UID/GID/filesystem-ID/SELinux transition and pre-exec readbacks, call `execveat(AT_EMPTY_PATH)` on the digest-verified `O_PATH` FD. Close all FDs except null-connected standard streams and supervisor-owned config FD; use only manifest allowlisted locale/loader/config environment values.

- [ ] **Step 7: Verify the actual process after start**

Record pidfd/start token, then read back `/proc/<pid>/exe`, UID/GID, capabilities, SELinux domain, `NoNewPrivs`, seccomp, cgroup and prlimit. Any mismatch kills the exact cgroup within five seconds, records `release_or_process_integrity`, and returns no owned process.

- [ ] **Step 8: Run deterministic sandbox tests**

Run: `go test ./internal/nodesupervisor/sandbox -run 'Test(MemoryPolicy|Namespace|Seccomp|Network|PinnedExec|PostStart)' -count=1`

Expected: PASS on supported Linux test hosts or explicit skip for kernel features unavailable to the unit runner; no result is labeled real Linux platform evidence.

On `!linux`, `exec_stub.go` supplies every launcher/owned-process/namespace/network/ownership constructor referenced by generic runtime code and returns only `ErrUnsupportedPlatform` before child creation. `exec_stub_test.go` compile-assigns the Linux/stub surface and checks zero child/FD/config-view side effects. It does not emulate any sandbox feature.

- [ ] **Step 9: Commit the launcher**

```bash
git add internal/nodesupervisor/sandbox/memory_policy.go internal/nodesupervisor/sandbox/deterministic_memory_policy_fake.go internal/nodesupervisor/sandbox/namespaces_linux.go internal/nodesupervisor/sandbox/seccomp_linux.go internal/nodesupervisor/sandbox/network_linux.go internal/nodesupervisor/sandbox/exec_linux.go internal/nodesupervisor/sandbox/ownership_linux.go internal/nodesupervisor/sandbox/exec_stub.go internal/nodesupervisor/sandbox/memory_policy_test.go internal/nodesupervisor/sandbox/exec_linux_test.go internal/nodesupervisor/sandbox/network_linux_test.go internal/nodesupervisor/sandbox/exec_stub_test.go
git commit -m "feat: sandbox supervised node processes"
```

### Task B05-T06: Implement prepare/check/start/lease/drain/rollback and dual-fault clear

**Files:**
- Create: `internal/nodesupervisor/trust/types.go`
- Create: `internal/nodesupervisor/trust/verifier.go`
- Test: `internal/nodesupervisor/trust/verifier_test.go`
- Create: `internal/nodesupervisor/runtime/types.go`
- Create: `internal/nodesupervisor/runtime/service.go`
- Create: `internal/nodesupervisor/runtime/check.go`
- Create: `internal/nodesupervisor/runtime/lease.go`
- Create: `internal/nodesupervisor/runtime/rollback.go`
- Create: `internal/nodesupervisor/runtime/faults.go`
- Test: `internal/nodesupervisor/runtime/service_test.go`
- Test: `internal/nodesupervisor/runtime/lease_test.go`
- Test: `internal/nodesupervisor/runtime/rollback_test.go`
- Test: `internal/nodesupervisor/runtime/faults_test.go`

**Interfaces:**
- Consumes: B01 exact `DesiredReasonV1` value `lease_refresh`; Plan 02 canonical root-set、trust-metadata、desired/recovery/resource-envelope DTOs; a preprovisioned C1.2 root anchor and exact deployment-role key; Plan 04 canonical `contracts.HostMemoryIsolationPolicyV1`; `TrustedTimeSource`; B05-T02 guard/leases/faults; B05-T03 compiler/seals; B05-T04 reservation; and B05-T05 launcher. It never consumes an agent `Verified*` object or trusts an agent high-water decision.
- Produces: supervisor-owned immutable `trust.VerifiedPrepareV1`、`VerifiedRenewV1` and `VerifiedRecoveryClearV1`; a complete `wire.Handler` implementation for all eleven operations, `SupervisorSecurityFaultV1` bindings, the sole canonical `LeaseRefreshSealV1`/`RollbackSealV1` definitions in `internal/nodesupervisor/runtime/types.go`, and only the finite typed B05-T01 operation results. Agent code can retain returned opaque seal IDs but cannot construct, open, or deserialize any verified value or sealed payload.

- [ ] **Step 1: Write the failing operation-state matrix**

```go
func TestOperationOrderIsClosed(t *testing.T) {
	service := newRuntimeFixture(t)
	binding := validBinding()
	if _, err := service.Start(context.Background(), wire.StartRequest{Binding: binding}); !errors.Is(err, ErrOperationOrder) {
		t.Fatalf("start-before-check error = %v", err)
	}
	prepared := service.MustPrepare(t, binding)
	checked := service.MustCheck(t, prepared)
	started := service.MustStart(t, checked)
	if started.LeaseID == (wire.OpaqueID{}) {
		t.Fatal("empty lease ID")
	}
}
```

- [ ] **Step 2: Run runtime tests and verify RED**

Run: `go test ./internal/nodesupervisor/runtime -run 'TestOperationOrderIsClosed|TestLeaseDeadline|TestLeaseRefreshSeal|TestRollbackSeal|TestSealIsSupervisorOnlyAndSingleUse|TestStopRequestClosedUnion|TestDualFaultClear' -count=1`

Expected: FAIL because runtime service is absent.

- [ ] **Step 3: Define exact supervisor-only lease-refresh and rollback seals**

In `internal/nodesupervisor/runtime/types.go`, freeze these payloads:

```go
type LeaseBindingV1 struct {
	NodeID                         uuid.UUID
	ControlPlaneAuthorityEpoch     uint64
	NodeAuthorityCheckpoint        uint64
	SlotID                         string
	Generation                     uint64
	AuthoritySequence              uint64
	DesiredDigest                  contracts.Digest
	LeaseID                        wire.OpaqueID
	ConfigSealID                   wire.OpaqueID
	SemanticDigest                 contracts.Digest
	ReleaseID                      string
	ReleaseDigest                  contracts.Digest
	ProcessDigest                  contracts.Digest
	ProfileID                      string
	ProfileVersion                 uint64
	CapacityBindingDigest          contracts.Digest
	ResourceEnvelope               contracts.VersionedDigest
	HostMemoryPolicy               contracts.VersionedDigest
	EffectiveAuthorizationDeadline time.Time
}

type LeaseRefreshSealV1 struct {
	OldLeaseBinding  LeaseBindingV1
	NewGeneration    uint64
	NewDesiredDigest contracts.Digest
	SemanticDigest   contracts.Digest
	NewValidUntil    time.Time
	BootNonce        wire.OpaqueID
}

type AuthorizingKeyStatusV1 string

const AuthorizingKeyStatusActive AuthorizingKeyStatusV1 = "active"

type RollbackSealV1 struct {
	PriorHighestAppliedDesired          contracts.VersionedDigest
	PriorConfigSealID                   wire.OpaqueID
	PriorConfigSemanticDigest           contracts.Digest
	CandidateDesired                    contracts.VersionedDigest
	CandidateConfigSealID               wire.OpaqueID
	CandidateConfigSemanticDigest       contracts.Digest
	SlotID                              string
	TransitionID                        wire.OpaqueID
	BootNonce                           wire.OpaqueID
	PriorEffectiveAuthorizationDeadline time.Time
	AuthorizingRootDigest               contracts.Digest
	AuthorizingMetadataDigest           contracts.Digest
	AuthorizingKeyID                    contracts.Digest
	AuthorizingKeyStatus                AuthorizingKeyStatusV1
	PriorReleaseDigest                  contracts.Digest
	PriorProcessDigest                  contracts.Digest
	CandidateReleaseDigest              contracts.Digest
	CandidateProcessDigest              contracts.Digest
	CreatedAt                           time.Time
	ExpiresAt                           time.Time
}
```

Every UUID/opaque ID/digest is nonzero, slot/release/profile IDs use their canonical B01 bounds, versions/sequences/checkpoints are in `1..MaxInt64`, and all times are UTC whole-millisecond. A refresh payload is valid only for a signed desired whose B01 reason is exactly `lease_refresh`, lifecycle remains `running`, new generation strictly advances, and node/slot/process spec/release/profile/capacity binding/resource-envelope/config semantic digest are byte-identical to `OldLeaseBinding`. The canonical Plan 04 memory policy may advance only after independent safe-policy verification; any change that affects compiler or sandbox semantics rejects the rebind. `NewValidUntil` is after trusted now and no later than the new composite effective authorization deadline. Changed fields, `stopped`/`draining`, a pending security fault, a non-`lease_refresh` reason, or an old-generation renewal never creates a refresh seal.

The rollback payload requires `CandidateDesired.Version > PriorHighestAppliedDesired.Version`, distinct config seal IDs, exact root/metadata/key bindings, `AuthorizingKeyStatus=active`, and `ExpiresAt=min(CreatedAt+2m,PriorEffectiveAuthorizationDeadline,metadata/key validity)` with `CreatedAt < ExpiresAt`. It is created before the prior process is stopped and is destroyed by candidate success, any newer prepare/desired, stopped/draining/security intent, authorization/key expiry or revocation, lease loss, fault creation, or boot change.

The store generates a random 16-byte external seal ID, canonical-encodes the payload, and authenticated-seals it with the B05-T02 supervisor-only identity. Use purpose `talenro-supervisor-lease-refresh-seal-v1` or `talenro-supervisor-rollback-seal-v1` and associated data `(schema version, supervisor boot nonce, node ID, slot ID, seal ID, guard counter)`. Plaintext never crosses the socket or reaches disk; the agent receives only the opaque ID in the typed response. Open and consume run only inside the single-writer runtime after peer/binding/trusted-time checks. Consumption and rollback-guard/high-water/boot-lease mutation are one atomic guarded transition; an exact request retry returns the stored typed result, while any second semantic consumption, cross-node/slot/transition/boot use, ciphertext/AAD mutation, or post-expiry use returns `ErrRejected` without revealing which binding failed.

Add table tests `TestLeaseRefreshSealBindsEveryField`, `TestLeaseRefreshRequiresExactReason`, `TestRollbackSealBindsEveryField`, and `TestSealIsSupervisorOnlyAndSingleUse`. Crash tests stop before seal persist, after persist/before guard mutation, after guard mutation/before response, and after response loss; recovery must expose exactly the old lease or the committed new binding, never two usable seals or a downgraded high-water.

- [ ] **Step 4: Independently verify every supervisor authorization input**

`internal/nodesupervisor/trust` strict-decodes the canonical B02 DTOs itself, verifies one continuous threshold root chain from the preprovisioned C1.2 anchor, exact deployment role/signature domain, complete metadata, authorizing online key status, cumulative revoked-key ledger, desired/recovery signature, canonical JSON, node/audience, control-plane authority epoch, node checkpoint, stream version/authority-sequence/digest, resource-envelope and memory-policy bindings, trusted UTC whole-millisecond validity and minimum supervisor version. It compares all high-waters against `SupervisorRollbackStateV1`, using `contracts.LocalVersionedDigestV1` only for manifest/map and authority-bearing `VersionedDigest` for signed control-plane streams, plus `RevokedKeyLedgerDigest`; rollback、same-value fork、ledger drop、unknown schema/role/key、insufficient threshold、wrong audience、future/expired time and unsupported minimum version return a finite trust error before compiler、reservation、launcher or latch-clear mutation.

Only this package can construct immutable defensive-copy `VerifiedPrepareV1`、`VerifiedRenewV1` and `VerifiedRecoveryClearV1`; no exported constructor accepts a digest、boolean or agent `Verified*` value. `Prepare` consumes only `VerifiedPrepareV1`; `Renew` consumes only `VerifiedRenewV1`; `ClearFault` consumes only `VerifiedRecoveryClearV1`. The clear verifier additionally requires stopped recovery action、exact sorted local/supervisor fault bindings、remediation digest、all-slots-stopped and current certificate/identity audience. Tests mutate every root/metadata/desired/recovery/audience/authority/time/ledger/min-version field, cover threshold/root fork and restart, and assert zero compiler/launcher/clear calls on rejection.

Run: `go test ./internal/nodesupervisor/trust -run 'TestVerify(Prepare|Renew|RecoveryClear)|TestTrustRestartHighWater' -count=1`

Expected: PASS; the supervisor reaches the same decision from canonical signed bytes and its own providers/state, without accepting an agent verdict.

- [ ] **Step 5: Implement prepare/check/start**

Prepare accepts only the Step 4 `VerifiedPrepareV1`, acquires aggregate reservation, compiles and seals config. Check launches one ephemeral no-network check cgroup, drains stdout/stderr to null/bounded byte counter, requires success and verified empty cleanup, then marks the same seal checked. Start consumes that seal once and commits a boot-bound lease only after post-start verification.

- [ ] **Step 6: Implement probe and bounded snapshots**

Probe checks process identity, loopback management and typed real-handshake result. Snapshot returns finite enum/integer measurements only; it never returns stdout/stderr, path, config, credential or raw upstream error.

- [ ] **Step 7: Implement renew, drain and stop deadlines**

Renew first consumes only a newly independently verified `VerifiedRenewV1` and moves the volatile deadline no later than trusted-now+60s or effective authorization deadline. Ordinary agent death/EOF/expiry starts drain with `hard_stop_at=min(drain_started+10m,lease_deadline+10m,effective_deadline+10m)`. Security/integrity/explicit stopped intent terminates and kills the exact cgroup within five seconds.

Stop validates the B05-T01 union before mutation. `AllOwned=false` requires the complete exact live lease binding and terminates only that ledger/cgroup entry. `AllOwned=true` requires the already pidfd-verified configured agent peer plus exact configured node/current supervisor boot node-scoped binding and empty lease/slot/generation/authority-sequence/desired fields；it snapshots only the supervisor's sealed ownership ledger, cross-checks each entry is beneath the configured node service cgroup root, terminates that finite set, proves each exact cgroup empty and returns a node-scoped response. With no lease it is an idempotent empty proof；with multiple slots it removes every and only owned current-node entry. Any cross-node、old/future boot、partial binding、global cgroup/process enumeration or ledger/root disagreement returns `ErrRejected` and changes nothing. Add runtime tables `TestStopRequestClosedUnion` and `TestStopAllOwnedNoLeaseMultiSlotCrossNodeAndBoot`.

- [ ] **Step 8: Implement single-use rollback and lease refresh seals**

Implement the Step 3 payloads and guarded store. Only a supervisor-observed candidate check/start/probe failure within the sealed two-minute/composite deadline may consume rollback and mint a new one-time start seal from the prior capsule; an agent-declared failure is insufficient. Lease refresh creates and consumes its seal within one guarded rebind, advances highest-verified and highest-applied to the new desired binding, and leaves the process identity unchanged only when the semantic digest and every process/release/profile/capacity field match. Both paths return only the B05-T01 typed IDs and fields.

- [ ] **Step 9: Implement supervisor-fault reporting and dual clear**

Fault detection first makes leases non-accepting, kills owned cgroups and persists pending fault. `ListFaults` is bounded. `ClearFault` consumes only `VerifiedRecoveryClearV1`, then requires exact local-latch attestation and empty owned cgroups; agent ACK alone cannot remove the supervisor fault.

- [ ] **Step 10: Run crash, deadline and race suites**

Run: `go test ./internal/nodesupervisor/trust ./internal/nodesupervisor/runtime -count=1`

Run: `go test -race ./internal/nodesupervisor/trust ./internal/nodesupervisor/runtime -count=1`

Expected: PASS; every crash point preserves either the old valid lease or fail-closed state, no seal replays across boot/slot/transition, and clear requires both latch authorities.

- [ ] **Step 11: Commit runtime semantics**

```bash
git add internal/nodesupervisor/trust/types.go internal/nodesupervisor/trust/verifier.go internal/nodesupervisor/trust/verifier_test.go internal/nodesupervisor/runtime/types.go internal/nodesupervisor/runtime/service.go internal/nodesupervisor/runtime/check.go internal/nodesupervisor/runtime/lease.go internal/nodesupervisor/runtime/rollback.go internal/nodesupervisor/runtime/faults.go internal/nodesupervisor/runtime/service_test.go internal/nodesupervisor/runtime/lease_test.go internal/nodesupervisor/runtime/rollback_test.go internal/nodesupervisor/runtime/faults_test.go
git commit -m "feat: manage supervised process leases"
```

### Task B05-T07: Compose the root-owned supervisor server

**Files:**
- Create: `internal/nodesupervisor/config/config.go`
- Test: `internal/nodesupervisor/config/config_test.go`
- Create: `cmd/node-core-supervisor/main.go`
- Create: `cmd/node-core-supervisor/run.go`
- Test: `cmd/node-core-supervisor/main_test.go`
- Test: `cmd/node-core-supervisor/run_test.go`
- Modify: `.env.example`

**Interfaces:**
- Consumes: B05-T01 `wire.Serve`, B05-T02–T06 providers/state/runtime, preprovisioned root-anchor/deployment-role-key paths and other root-owned deployment paths.
- Produces: `node-core-supervisor`, one local socket, bounded shutdown and local/test deterministic composition used by B06.

- [ ] **Step 1: Write the failing profile, catalog and startup stop-union tests**

```go
func TestProductionRejectsDeterministicMemoryPolicy(t *testing.T) {
	lookup := validSupervisorProductionLookup()
	lookup["TALENRO_SUPERVISOR_MEMORY_POLICY_PROVIDER"] = "DeterministicHostMemoryIsolationPolicyFakeV1"
	if _, err := config.Load(mapLookup(lookup)); !errors.Is(err, config.ErrUnsafeProvider) {
		t.Fatalf("error = %v", err)
	}
}
```

- [ ] **Step 2: Run config/composition tests and verify RED**

Run: `go test ./internal/nodesupervisor/config ./cmd/node-core-supervisor -run 'TestProductionRejectsDeterministicMemoryPolicy|TestProductionRejectsCatalogOverrideOrFixture|TestStartupFailClosed|TestStartupStopAllOwned(NoLease|MultiSlot|RejectsCrossNode|RejectsOldBoot)' -count=1`

Expected: FAIL because config and process entrypoint are absent.

- [ ] **Step 3: Implement exact configuration validation**

Require canonical node ID, exact agent/core UIDs, socket/state roots, C1.2 root-anchor and deployment-role-key paths/identities, envelope/memory-policy package paths, provider identities, protocol/build digests and operation deadlines. Approved-manifest source bytes are B02-owned and build-embedded；installed-map path is the non-overridable literal `/etc/talenro/releases/installed-release-map.v1.json`, so configuration/env contains neither catalog bytes nor manifest/map path/version/digest. Each trust-anchor/key/deployment parent is absolute、root-owned、non-writable、no-follow and distinct from agent mutable state. Production rejects deterministic providers、fixture/test catalog sources、catalog override keys、missing/tampered fixed map、writable/non-root paths, overlapping agent/supervisor/latch identities and non-linux/amd64 runtime.

- [ ] **Step 4: Compose startup before accepting the socket**

Startup order is provider identity/time → load and pin exact root anchor/deployment-role key → construct the independent trust verifier → independently strict-verify the build-embedded B02 approved manifest and fixed root-owned no-follow installed map plus immutable release roots → memory-policy verification → state guard → atomically compare/persist the two derived local catalog tuples with resource-envelope/revoked-ledger high-water → use only the configured node service cgroup root to clean/prove empty all old-boot cgroups → create new boot nonce/empty volatile ownership ledger/lease table → socket ownership/mode → accept. Old boot leases are never reconstructed. Any failure leaves no socket and no child.

`TestStartupStopAllOwnedNoLease` starts an exact verified peer/current-node/current-boot server with an empty ownership ledger and requires idempotent empty success. `TestStartupStopAllOwnedMultiSlot` seeds two current-node ledger/cgroup entries and one baseline foreign cgroup, requires exactly the two owned entries empty, and proves the baseline untouched. `TestStartupStopAllOwnedRejectsCrossNode` and `TestStartupStopAllOwnedRejectsOldBoot` submit otherwise canonical node-scoped requests and require `REJECTED`、zero enumeration and zero mutation. All four are process-level server tests and also prove a complete lease binding cannot be mixed with `AllOwned=true`.

- [ ] **Step 5: Implement bounded shutdown**

Stop accept, reject new prepare/check/start/renew, make leases non-accepting, kill security-fault leases within five seconds, drain ordinary leases within their existing hard-stop deadline, verify all cgroups/config views empty, close socket and guard. Raw errors and paths never reach output.

- [ ] **Step 6: Record explicit local/test provider values**

```dotenv
TALENRO_SUPERVISOR_ROLLBACK_PROVIDER=DeterministicRollbackGuardFakeV1
TALENRO_SUPERVISOR_TIME_PROVIDER=DeterministicTrustedTimeFakeV1
TALENRO_SUPERVISOR_MEMORY_POLICY_PROVIDER=DeterministicHostMemoryIsolationPolicyFakeV1
```

The adjacent comment states `local/test only; scope=container_deterministic`.

- [ ] **Step 7: Run all B05 tests**

Run: `go test ./internal/nodesupervisor/... ./cmd/node-core-supervisor -count=1`

Run: `go test -race ./internal/nodesupervisor/... ./cmd/node-core-supervisor -count=1`

Expected: PASS; startup failure creates no socket/child, production catalog absence/override/fixture/tamper fails before accept, the no-lease/multi-slot/cross-node/boot stop matrix touches only exact owned cgroups, and shutdown leaves no owned cgroup or received FD.

The untagged gate runs on Windows and must compile every generic package against the exact `!linux` wire/resource/launcher stubs. `TestUnsupportedPlatformSurface` invokes all exported platform constructors and requires `ErrUnsupportedPlatform` plus zero socket、FD、child、config-view and cgroup effects.

- [ ] **Step 8: Build and inspect the fixed target**

Run:

```powershell
$supervisorBuildRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("talenro-c12-b05-" + [guid]::NewGuid().ToString('N'))
$supervisorBuildPath = Join-Path $supervisorBuildRoot 'node-core-supervisor'
$supervisorOldCGO = [Environment]::GetEnvironmentVariable('CGO_ENABLED', 'Process')
$supervisorOldGOOS = [Environment]::GetEnvironmentVariable('GOOS', 'Process')
$supervisorOldGOARCH = [Environment]::GetEnvironmentVariable('GOARCH', 'Process')
New-Item -ItemType Directory -LiteralPath $supervisorBuildRoot | Out-Null
try {
    $env:CGO_ENABLED = '0'
    $env:GOOS = 'linux'
    $env:GOARCH = 'amd64'
    go build -trimpath -buildmode=exe -o $supervisorBuildPath ./cmd/node-core-supervisor
    if ($LASTEXITCODE -ne 0 -or -not (Test-Path -LiteralPath $supervisorBuildPath -PathType Leaf)) { throw 'supervisor linux build failed' }
    $supervisorProductionDeps = @(go list -deps ./cmd/node-core-supervisor)
    if ($LASTEXITCODE -ne 0) { throw 'supervisor linux/amd64 dependency listing failed' }
    if (@($supervisorProductionDeps | Where-Object { $_ -match '^github\.com/(xtls/xray-core|sagernet/sing-box)(/|$)' }).Count -ne 0) { throw 'supervisor production dependency exclusion failed' }
} finally {
    foreach ($supervisorEnv in @(@('CGO_ENABLED',$supervisorOldCGO),@('GOOS',$supervisorOldGOOS),@('GOARCH',$supervisorOldGOARCH))) {
        if ($null -eq $supervisorEnv[1]) { Remove-Item -LiteralPath ("Env:" + $supervisorEnv[0]) -ErrorAction SilentlyContinue } else { [Environment]::SetEnvironmentVariable($supervisorEnv[0], $supervisorEnv[1], 'Process') }
    }
    $resolvedSupervisorBuildRoot = [System.IO.Path]::GetFullPath($supervisorBuildRoot)
    $supervisorTempSeparators = [char[]]@([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
    $resolvedSupervisorTempRoot = [System.IO.Path]::GetFullPath([System.IO.Path]::GetTempPath()).TrimEnd($supervisorTempSeparators)
    if ([System.IO.Path]::GetDirectoryName($resolvedSupervisorBuildRoot) -ne $resolvedSupervisorTempRoot -or [System.IO.Path]::GetFileName($resolvedSupervisorBuildRoot) -notmatch '^talenro-c12-b05-[0-9a-f]{32}$') { throw 'unsafe supervisor build cleanup target' }
    Remove-Item -LiteralPath $resolvedSupervisorBuildRoot -Recurse -Force
}
```

Expected: the production release target build succeeds and its exact Linux/amd64 dependency set contains neither core module. The wrapper snapshots/restores any prior three environment values in `finally`, verifies the resolved build root is an immediate uniquely named child of the OS temp root, and removes only that exact root. It never writes a repository-root or user cache output.

- [ ] **Step 9: Commit the composition root**

```bash
git add internal/nodesupervisor/config/config.go internal/nodesupervisor/config/config_test.go cmd/node-core-supervisor/main.go cmd/node-core-supervisor/run.go cmd/node-core-supervisor/main_test.go cmd/node-core-supervisor/run_test.go .env.example
git commit -m "feat: compose node core supervisor"
```

## B05 exit gate

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1
if ($LASTEXITCODE -ne 0) { throw 'B05 generation failed' }
git diff --exit-code HEAD -- api gen internal/store
if ($LASTEXITCODE -ne 0) { throw 'B05 generated tree differs from committed tree' }
$untrackedB05Generated = @(git ls-files --others --exclude-standard -- api gen internal/store)
if ($LASTEXITCODE -ne 0 -or $untrackedB05Generated.Count -ne 0) { throw 'untracked B05 generated artifact' }
go test ./internal/nodesupervisor/... ./cmd/node-core-supervisor -count=1
go test -race ./internal/nodesupervisor/... ./cmd/node-core-supervisor -count=1
go vet ./internal/nodesupervisor/... ./cmd/node-core-supervisor
go tool golangci-lint run ./internal/nodesupervisor/... ./cmd/node-core-supervisor
git diff --check
```

Expected: every command succeeds; protocol generation is byte-equal to the committed tree and the complete generated closure has no untracked output；frame/FD/lease/seal/fault/resource tests are bounded; deterministic memory-policy evidence is only `container_deterministic`; no test or document claims the real Linux platform/operator-trust gate from分册09 has passed.
