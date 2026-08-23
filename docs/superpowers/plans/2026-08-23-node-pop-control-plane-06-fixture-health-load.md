# Talenro C1.2 Fixture, Health, and Load Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 以独立 deterministic fixture 完成 node-agent→supervisor controlled-process/crash 验收，并交付可重放的健康容量 reducer、有限 observation、隐私门和 1,000-agent reference load acceptance。

**Architecture:** B06 使用独立 fixture executable和固定 typed profile，在真实 agent/supervisor process boundary上确定性触发 check、start、probe、drain、crash、resource和cleanup路径。B07 让 agent与control plane复用同一纯 `CapacityReducerV1`，服务端根据 PostgreSQL profile、原始 observation与受信 arrival interval独立重放并只收紧 `accepting_new`。Container/Linux fixture只使用 deterministic anti-rollback/time/memory-policy fake；真实 attested Linux platform/memory gate留给分册09。

**Tech Stack:** Go 1.26.5、subprocess fixtures、Unix process groups/cgroup v2/pidfd/seccomp/nftables、Windows Job Objects portability、PostgreSQL/sqlc integration、Prometheus、deterministic load generation。

**Spec:** [Approved C1.2 node and POP control-plane design](../specs/2026-08-23-node-pop-control-plane-design.md), especially §§13–18 and §21 batches 6–7; suite index [C1.2 implementation plan](2026-08-23-node-pop-control-plane.md).

## Global Constraints

- 本册严格分成 `C1.2-B06` 和 `C1.2-B07` 两个 exit；B06通过不能冒充B07通过，两个 batch各自有独立 RED/GREEN/commit与验收。
- Fixture 是独立 executable，不是 agent/supervisor 内嵌测试分支。Agent/supervisor production code不得依据 fixture mode绕过验证、sandbox、lease、probe或cleanup。
- Fixture、native-check和smoke-client仍是不可信 child；任何测试模式都必须走 B05 digest-pinned manifest、compiler、config seal、cgroup、namespace、seccomp、nftables、pidfd和lease边界。
- B06/B07 container tests只允许精确 provider名 `DeterministicRollbackGuardFakeV1`、`DeterministicSecurityLatchGuardFakeV1`、`DeterministicTrustedTimeFakeV1`、`DeterministicHostMemoryIsolationPolicyFakeV1`，evidence scope固定 `container_deterministic`。
- 本册不得读取、安装或声称控制 outer host SELinux policy、Yama/BPF one-way sysctl、`core_pattern`、systemd collector、TPM或secure-time provider；真实 forced-crash/memory-attach/platform gate只属于分册09。
- Fixture configuration只能由 signed typed profile经 supervisor compiler产生；desired state不得携带 fixture command、path、argv、environment或任意 config。
- Controlled-process必须覆盖 enrollment、gen1 convergence、gen2 candidate failure+supervisor rollback seal恢复gen1、更高generation修复、lease-refresh不重启、changed-spec正常transition、LKG/expiry和security stop。
- Fixture slot状态只允许 `stopped/preparing/starting/healthy/degraded/draining/restarting/quarantined`；5分钟内连续5次失败进入slot quarantine，只有更高 `clear_slot_quarantine` generation可清除。
- Observation约每5秒一份，最大64 KiB、最多8 slot；metric state仅 `valid/unknown/unsupported`，只有valid可携带整数value。
- Base required metrics固定按bytewise排序：`cpu_basis_points/egress_bps/memory_bytes/open_file_descriptors/task_count`；packet loss和queue depth固定unsupported。
- Sample window只接受4–10秒；新boot首份rate unknown；同node每4秒最多一份report推进reducer；20秒gap使health offline并清空EWMA。
- `CapacityReducerV1` 固定30秒time-decay EWMA；unknown→healthy需要3个连续valid且所有required ratio<70%；healthy连续3窗>=75%降级，任一>=90%立即capacity-blocked；恢复需要6窗全部<70%。
- 服务端权威 `accepting_new` 是 operator enabled、security normal、health healthy/degraded、非capacity-blocked、agent candidate true、authorization未过期的AND；服务端不得把agent false放宽为true。
- Metrics/log/error/report不得使用node ID、slot ID、address、path、credential、raw core/provider error或raw sample作为label；所有label来自有限enum。
- Reference load runner固定Linux amd64、8 dedicated vCPU/16 GiB；control-api 4 vCPU/4 GiB、PostgreSQL 2 vCPU/4 GiB、load generator 2 vCPU/2 GiB。资源或digest不符只产生diagnostic。
- Load固定seed `0xC12A6E17`、1,000 agent、25秒均匀错峰、120秒warm-up、600秒measurement、measurement第60秒发布generation；p95<10秒、p99<20秒、30秒内100% ACK。
- 每 task执行 RED → GREEN → REFACTOR、focused/package/race test和独立commit；不得stage/删除用户 `.cache/`、`.superpowers/`、`.task19-go/`。

---

## Repository map

```text
cmd/c12-fixture/                         # independent deterministic child
cmd/c12-load/                            # 1,000-agent generator and report
internal/nodeagent/adapter/fixture/       # preview/profile/probe client
internal/nodesupervisor/adapter/fixture/  # supervisor-owned compiler
internal/nodeagent/observation/           # cgroup/adapter metric collection
internal/nodecontrol/observation/         # shared reducer + authoritative ingest
internal/c12test/                         # exact-process test harness, not B10 ownership WAL
internal/e2e/c12_*                        # controlled process/privacy tagged suites
testdata/c12/fixture/                     # typed configs and deterministic vectors
testdata/c12/load/                        # runner profile and expected report schema
```

## C1.2-B06: Deterministic fixture and controlled process

### Task B06-T01: Build the independent deterministic fixture executable

**Files:**
- Create: `cmd/c12-fixture/main.go`
- Create: `internal/c12test/fixture/config.go`
- Create: `internal/c12test/fixture/process.go`
- Create: `internal/c12test/fixture/management.go`
- Test: `cmd/c12-fixture/main_test.go`
- Test: `internal/c12test/fixture/process_test.go`
- Test: `internal/c12test/fixture/management_test.go`
- Create: `testdata/c12/fixture/valid-config.v1.json`

**Interfaces:**
- Consumes: a supervisor-created read-only canonical `FixtureConfigV1` file and fixed manifest argv; no desired/user-provided command data.
- Produces: `c12-fixture` with deterministic check/run behavior, loopback management/probe counters and finite exit/result codes used by the fixture adapter.

- [ ] **Step 1: Write the failing mode and output-privacy table**

```go
func TestFixtureModesAreClosedAndSilent(t *testing.T) {
	for _, mode := range []fixture.Mode{
		fixture.ModeHealthy, fixture.ModeCheckFailure, fixture.ModeStartupDelay,
		fixture.ModeHandshakeFailure, fixture.ModeImmediateExit, fixture.ModeIgnoreTerminate,
		fixture.ModeDrainTimeout, fixture.ModeThreadTaskBomb, fixture.ModeMemoryPressure,
		fixture.ModeFDExhaustion, fixture.ModeInvalidMetric,
	} {
		result := runFixtureMode(t, mode)
		if bytes.Contains(result.Stdout, []byte("fixture-canary")) || bytes.Contains(result.Stderr, []byte("fixture-canary")) {
			t.Fatalf("mode %s leaked canary", mode)
		}
	}
}
```

- [ ] **Step 2: Run fixture tests and verify RED**

Run: `go test ./cmd/c12-fixture ./internal/c12test/fixture -run 'TestFixtureModesAreClosedAndSilent|TestFixtureConfigStrict' -count=1`

Expected: FAIL because fixture packages do not exist.

- [ ] **Step 3: Define the exact canonical config**

```go
type ConfigV1 struct {
	SchemaVersion       string `json:"schema_version"`
	Mode                Mode   `json:"mode"`
	StartupDelayMS      uint32 `json:"startup_delay_ms"`
	DrainDelayMS        uint32 `json:"drain_delay_ms"`
	ManagementPort      uint16 `json:"management_port"`
	ExpectedProbeNonce  string `json:"expected_probe_nonce"`
	ConnectionBaseline uint32 `json:"connection_baseline"`
}
```

Require schema `c12-fixture-config.v1`, one closed mode, delays 0–30,000 ms, loopback port 1024–65535, a 43-character unpadded base64url 32-byte nonce and no unknown/duplicate/trailing JSON.

- [ ] **Step 4: Implement explicit check and run entry modes**

Fixed argv permits only `check --config-fd=3` or `run --config-fd=3`; the config FD is supervisor-owned, read-only and already sealed. Check parses/validates and exits `0` or fixed `20`. Run uses fixed exit classes `0 clean`, `30 immediate_exit`, `31 check_bypass`, `32 invalid_config`; it never echoes input or raw errors.

- [ ] **Step 5: Implement loopback management and deterministic counters**

Expose only a length-prefixed loopback management socket inside the slot namespace. Operations are `health`, `probe`, `snapshot`, `begin_drain`, `stop`; replies contain finite enum/integer fields. Healthy probe returns the exact nonce; handshake-failure mode returns `probe_failed` without text.

- [ ] **Step 6: Implement fault modes without host-wide effects**

Thread task-bomb uses only exact allowed thread clone behavior; FD exhaustion stays within process `RLIMIT_NOFILE`; memory pressure allocates until slot `memory.max`; ignore-terminate ignores only the graceful signal; no mode changes sysctl, host policy, mount, global process or another cgroup.

- [ ] **Step 7: Run package and subprocess tests**

Run: `go test ./cmd/c12-fixture ./internal/c12test/fixture -count=1`

Expected: PASS; each mode terminates within its fixed test deadline and stdout/stderr contain no config, nonce or injected canary.

- [ ] **Step 8: Commit the fixture executable**

```bash
git add cmd/c12-fixture/main.go cmd/c12-fixture/main_test.go internal/c12test/fixture/config.go internal/c12test/fixture/process.go internal/c12test/fixture/management.go internal/c12test/fixture/process_test.go internal/c12test/fixture/management_test.go testdata/c12/fixture/valid-config.v1.json
git commit -m "test: add deterministic core fixture"
```

### Task B06-T02: Add typed fixture preview, compiler, manifest and probe

**Files:**
- Create: `internal/nodeagent/adapter/registry.go`
- Create: `internal/nodeagent/adapter/fixture/profile.go`
- Create: `internal/nodeagent/adapter/fixture/preview.go`
- Create: `internal/nodeagent/adapter/fixture/probe.go`
- Test: `internal/nodeagent/adapter/fixture/preview_test.go`
- Test: `internal/nodeagent/adapter/fixture/probe_test.go`
- Create: `internal/nodesupervisor/adapter/fixture/compiler.go`
- Test: `internal/nodesupervisor/adapter/fixture/compiler_test.go`
- Create: `testdata/c12/fixture/approved-release-manifest.v1.json`
- Create: `testdata/c12/fixture/installed-release-map.v1.json`
- Create: `testdata/c12/fixture/resource-envelope.v1.json`
- Create: `testdata/c12/fixture/memory-policy-fake-package.v1.json`

**Interfaces:**
- Consumes: B04 `PreviewAdapter`, B05 `AdapterCompilerV1`, signed fixture process spec and build/test manifest pins.
- Produces: the closed `adapter.Register(ProfileV1)` / `adapter.Preview(contracts.ProcessSpecV1)` registry consumed by B08/B09, fixture adapter kind/profile `fixture-controlled-process-v1`, secret-free preview digest, supervisor canonical config bytes and finite probe/snapshot values.

- [ ] **Step 1: Write the failing cross-compiler digest test**

```go
func TestPreviewMatchesSupervisorCompiler(t *testing.T) {
	spec := validFixtureProcessSpec()
	preview, err := fixtureagent.New().Preview(spec)
	if err != nil {
		t.Fatal(err)
	}
	compiled, err := fixturesupervisor.New().Compile(adapter.CompileRequestV1{
		NodeID:            validFixtureNodeID(),
		Generation:        1,
		SlotID:            spec.SlotID,
		ProcessSpec:       spec,
		ProcessSpecDigest: mustProcessSpecDigest(t, spec),
		ReleaseID:         spec.ReleaseID,
		ProfileID:         spec.TestProfile,
		Credential:        nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	if preview.SemanticDigest != compiled.SemanticDigest {
		t.Fatalf("preview/compiler digest mismatch")
	}
}
```

- [ ] **Step 2: Run adapter tests and verify RED**

Run: `go test ./internal/nodeagent/adapter/fixture ./internal/nodesupervisor/adapter/fixture -run TestPreviewMatchesSupervisorCompiler -count=1`

Expected: FAIL because fixture adapters are absent.

- [ ] **Step 3: Freeze the typed fixture profile**

`fixture-controlled-process-v1` maps signed enum fields for behavior, fixed loopback management endpoint, probe/restart profiles and exact resource limits. It contains no path, argv, environment, free-form string, credential, URL or raw config.

- [ ] **Step 4: Implement independent preview and compiler**

Agent preview validates typed semantics and hashes canonical secret-free config. Supervisor repeats validation, generates `FixtureConfigV1`, hashes the same semantic projection, seals actual bytes and uses only manifest-fixed check/run argv.

- [ ] **Step 5: Implement finite probe/snapshot mapping**

Map management replies to `healthy/probe_failed/process_missing/invalid_response/timeout`; map counters to integer metric states. Reject unknown enum, wrong nonce, oversized/truncated frame, negative/overflow value and any response text.

- [ ] **Step 6: Add exact deterministic manifests**

The approved manifest fixes fixture build digest, executable/dependency digests, check/start argv, compiler version, sandbox profile, capability list and license ID. Installed map contains only release ID and immutable absolute test-host root. Memory package names `DeterministicHostMemoryIsolationPolicyFakeV1` and scope `container_deterministic`.

- [ ] **Step 7: Run cross-compiler, manifest and probe tests**

Run: `go test ./internal/nodeagent/adapter/fixture ./internal/nodesupervisor/adapter/fixture -count=1`

Expected: PASS; agent cannot inject compiled bytes, manifest/map cannot override each other, and probe never returns raw fixture output.

- [ ] **Step 8: Commit typed fixture adapters**

```bash
git add internal/nodeagent/adapter/registry.go internal/nodeagent/adapter/fixture/profile.go internal/nodeagent/adapter/fixture/preview.go internal/nodeagent/adapter/fixture/probe.go internal/nodeagent/adapter/fixture/preview_test.go internal/nodeagent/adapter/fixture/probe_test.go internal/nodesupervisor/adapter/fixture/compiler.go internal/nodesupervisor/adapter/fixture/compiler_test.go testdata/c12/fixture/approved-release-manifest.v1.json testdata/c12/fixture/installed-release-map.v1.json testdata/c12/fixture/resource-envelope.v1.json testdata/c12/fixture/memory-policy-fake-package.v1.json
git commit -m "test: add typed fixture adapter"
```

### Task B06-T03: Prove controlled convergence, rollback, refresh and expiry

**Files:**
- Create: `internal/c12test/process_harness.go`
- Create: `internal/c12test/process_harness_linux.go`
- Create: `internal/c12test/process_harness_windows.go`
- Test: `internal/c12test/process_harness_test.go`
- Create: `internal/e2e/c12_controlled_process_test.go`
- Create: `internal/e2e/c12_lkg_expiry_test.go`

**Interfaces:**
- Consumes: B02/B03 in-process control-plane applications, B04 agent, B05 supervisor, B06-T01/T02 fixture and deterministic providers.
- Produces: tagged `c12fixture` happy-path/expiry acceptance and an exact-process harness that B10 later wraps with ownership WAL.

- [ ] **Step 1: Write the failing end-to-end scenario**

```go
func TestC12ControlledProcessConvergesAndRollsBack(t *testing.T) {
	harness := c12test.NewControlledProcessHarness(t, c12test.ContainerDeterministicProviders())
	node := harness.EnrollNode(t)
	harness.Publish(t, node, generation(1, fixture.ModeHealthy))
	harness.RequireApplied(t, node, 1)
	firstPID := harness.FixturePID(t, node)
	harness.Publish(t, node, generation(2, fixture.ModeCheckFailure))
	harness.RequireSeenApplied(t, node, 2, 1)
	harness.RequireFixturePID(t, node, firstPID)
	harness.Publish(t, node, generation(3, fixture.ModeHealthy))
	harness.RequireApplied(t, node, 3)
}
```

- [ ] **Step 2: Run the tagged scenario and verify RED**

Run: `go test -tags=c12fixture ./internal/e2e -run TestC12ControlledProcessConvergesAndRollsBack -count=1 -timeout 2m`

Expected: FAIL because the controlled-process harness is absent.

- [ ] **Step 3: Implement exact process ownership in the harness**

Use an OS temp-directory child, record exact PID/start token/image digest/process group or Job Object, and register cleanup before start. Never enumerate by name, wildcard or label. Linux uses an owned cgroup; Windows creates kill-on-close Job Object and assigns at process creation with `PROC_THREAD_ATTRIBUTE_JOB_LIST`.

- [ ] **Step 4: Compose deterministic node enrollment and signed generations**

Use B01–B03 application services with deterministic fence/issuer/signer/time providers. Generate the agent key locally, claim the bound grant, verify receipt, poll full signed state and report seen/applied. The harness may not bypass HTTP/mTLS authorization after enrollment.

- [ ] **Step 5: Implement gen1/gen2/gen3 and refresh assertions**

Assert gen1 converges; gen2 check failure consumes a supervisor rollback seal and leaves highest-seen=2/applied=1; gen3 recovers. A semantic-identical `lease_refresh` advances generation without PID/start-token change; changed-spec refresh takes the normal transition and changes process identity.

- [ ] **Step 6: Implement disconnect and deadline assertions**

Stop control-plane poll/report, keep a still-valid LKG, then advance deterministic trusted time to effective deadline. Require immediate non-accepting/draining and exact kill no later than deadline+10 minutes. Metadata/key emergency revocation and integrity faults use the five-second kill path.

- [ ] **Step 7: Run happy-path, refresh and expiry tests**

Run: `go test -tags=c12fixture ./internal/e2e -run 'TestC12(ControlledProcess|LeaseRefresh|LKGExpiry)' -count=1 -timeout 5m`

Expected: PASS; post-run exact processes/groups/cgroups/Job Objects and temp credentials are absent.

- [ ] **Step 8: Commit controlled-process acceptance**

```bash
git add internal/c12test/process_harness.go internal/c12test/process_harness_linux.go internal/c12test/process_harness_windows.go internal/c12test/process_harness_test.go internal/e2e/c12_controlled_process_test.go internal/e2e/c12_lkg_expiry_test.go
git commit -m "test: prove controlled node process convergence"
```

### Task B06-T04: Exercise crash, dual-latch, sandbox and aggregate resource faults

**Files:**
- Create: `internal/e2e/c12_crash_recovery_test.go`
- Create: `internal/e2e/c12_security_latch_test.go`
- Create: `internal/e2e/c12_resource_faults_test.go`
- Create: `internal/e2e/c12_sandbox_faults_test.go`
- Create: `internal/e2e/c12_cleanup_test.go`
- Modify: `internal/c12test/process_harness.go`
- Modify: `internal/c12test/process_harness_linux.go`

**Interfaces:**
- Consumes: B06-T03 harness and deterministic provider crash hooks from B04/B05.
- Produces: B06 fault matrix proving no process adoption, no forgotten fault, bounded resource isolation and exact cleanup; it emits only `container_deterministic` evidence.

- [ ] **Step 1: Write the failing crash-point table**

```go
func TestC12CrashPointsRemainFailClosed(t *testing.T) {
	for _, point := range []c12test.CrashPoint{
		c12test.AgentCandidateSync, c12test.AgentCounterIncrement,
		c12test.SupervisorCandidateSync, c12test.SupervisorCounterIncrement,
		c12test.LocalLatchCandidateSync, c12test.LocalLatchCounterIncrement,
		c12test.FaultBeforeAgentACK, c12test.ClearBetweenSupervisorAndAgentLatch,
	} {
		t.Run(point.String(), func(t *testing.T) {
			harness := c12test.NewCrashHarness(t, point)
			harness.RunAndRestart(t)
			harness.RequireNoCoreBeforeAuthoritativeClear(t)
		})
	}
}
```

- [ ] **Step 2: Run crash tests and verify RED**

Run: `go test -tags=c12fixture ./internal/e2e -run TestC12CrashPointsRemainFailClosed -count=1 -timeout 5m`

Expected: FAIL because fault hooks and assertions are not wired into the harness.

- [ ] **Step 3: Add agent/supervisor/socket failure cases**

Kill agent, close socket, kill supervisor, replay old boot lease and damage main guard. Require supervisor drain/reap, no adoption, exact stop-all-owned scope and a durable local/supervisor fault before any restart.

- [ ] **Step 4: Add seal/frame/FD attack cases**

Inject cross-slot/boot/generation config seal, rollback/refresh replay, production credential FD, swapped test roles, regular path FD, socket FD, oversize/truncated frame and `MSG_CTRUNC`. Require no core start, every received FD closed and goroutine count return to baseline.

- [ ] **Step 5: Add process and aggregate resource cases**

Exercise fork/vfork/clone3 denial, exact thread task-bomb, memory.max group OOM, FD hard limit, tmpfs byte/inode exhaustion and eight individually valid slots whose CPU/memory/task/FD/tmpfs sum exceeds envelope. Assert failure before spawn or isolation to the exact slot cgroup; agent/supervisor stay responsive.

- [ ] **Step 6: Add sandbox escape cases with deterministic memory policy only**

Attempt host `/etc`/agent-state read, host `/tmp` write, other-slot read, ptrace/FD copy, mount/setns/helper exec, non-allowlisted listen/connect and metadata/control-plane access. Use `DeterministicHostMemoryIsolationPolicyFakeV1`; assert scope `container_deterministic` and do not record host SELinux/Yama/BPF/dump proof.

- [ ] **Step 7: Add cleanup interruption cases**

Interrupt before create, after create/before owned record, after owned record and during cleanup. Recover only by exact PID/start token/cgroup/Job Object/temp root recorded by this harness; collision or ownership mismatch stops without deleting the candidate.

- [ ] **Step 8: Run the complete B06 fault suite**

Run: `go test -tags=c12fixture ./internal/e2e -run 'TestC12(Crash|SecurityLatch|Resource|Sandbox|Cleanup)' -count=1 -timeout 10m`

Expected: PASS; no child/cgroup/Job Object/temp credential remains and no result carries a real-platform scope.

- [ ] **Step 9: Commit B06 fault acceptance**

```bash
git add internal/c12test/process_harness.go internal/c12test/process_harness_linux.go internal/e2e/c12_crash_recovery_test.go internal/e2e/c12_security_latch_test.go internal/e2e/c12_resource_faults_test.go internal/e2e/c12_sandbox_faults_test.go internal/e2e/c12_cleanup_test.go
git commit -m "test: close controlled process fault matrix"
```

## B06 exit gate

Run:

```powershell
go test ./cmd/c12-fixture ./internal/c12test/... ./internal/nodeagent/adapter/fixture ./internal/nodesupervisor/adapter/fixture -count=1
go test -tags=c12fixture ./internal/e2e -run 'TestC12(ControlledProcess|LeaseRefresh|LKGExpiry|Crash|SecurityLatch|Resource|Sandbox|Cleanup)' -count=1 -timeout 10m
go test -race ./internal/c12test/... ./internal/nodeagent/adapter/fixture ./internal/nodesupervisor/adapter/fixture -count=1
git diff --check
```

Expected: all pass; controlled process and every failure path clean only exact run-owned resources; evidence scope is only `container_deterministic`; B07 remains unclaimed.

## C1.2-B07: Health, capacity, privacy and 1,000-agent load

### Task B07-T01: Implement the shared pure CapacityReducerV1

**Files:**
- Create: `internal/nodecontrol/observation/types.go`
- Create: `internal/nodecontrol/observation/reducer.go`
- Create: `internal/nodecontrol/observation/ruleset.go`
- Test: `internal/nodecontrol/observation/reducer_test.go`
- Test: `internal/nodecontrol/observation/reducer_vectors_test.go`
- Create: `testdata/c12/load/capacity-reducer-v1.json`

**Interfaces:**
- Consumes: ordered raw metric/probe/slot states, previous reducer state and trusted server arrival delta.
- Produces: `Reduce(Input) (Output, error)`, fixed `ReducerVersion="capacity-reducer.v1"`, ruleset digest, EWMA/hysteresis state, health, capacity-blocked and local candidate; both agent and server import this package.

- [ ] **Step 1: Write the failing threshold and arrival-spacing table**

```go
func TestCapacityReducerThresholds(t *testing.T) {
	state := observation.NewState(fixedBootID())
	state = applyWindows(t, state, 3, 6*time.Second, ratios(6_900))
	if state.Health != observation.HealthHealthy || !state.LocalAcceptingNew {
		t.Fatalf("healthy state = %#v", state)
	}
	state = applyWindows(t, state, 3, 6*time.Second, ratios(7_500))
	if state.Health != observation.HealthDegraded {
		t.Fatalf("degraded state = %#v", state)
	}
	state = applyWindows(t, state, 6, 6*time.Second, ratios(6_900))
	if state.Health != observation.HealthHealthy {
		t.Fatalf("recovered state = %#v", state)
	}
}
```

- [ ] **Step 2: Run reducer tests and verify RED**

Run: `go test ./internal/nodecontrol/observation -run 'TestCapacityReducerThresholds|TestArrivalSpacing' -count=1`

Expected: FAIL because reducer types are absent.

- [ ] **Step 3: Define closed metric/input/output types**

```go
type Metric struct {
	Kind           MetricKind
	State          MetricState
	Value          uint64
	Source         MetricSource
	SampleWindowNS uint64
}

type Input struct {
	BootID                [16]byte
	Sequence              uint64
	ServerArrivalDeltaNS  uint64
	RequiredSlotStates    []SlotState
	OptionalSlotStates    []SlotState
	Metrics               []Metric
	Previous              State
}
```

Reject more than8slots, duplicate metric kinds, nonvalid metric with value, sample outside4–10seconds, nonincreasing sequence/sample end and arrival below4seconds for a state-advancing report.

- [ ] **Step 4: Implement exact EWMA arithmetic and priority**

Use `alpha=1-exp(-elapsed/30s)` with deterministic float-to-fixed conversion frozen by vector tests; ratios derive from immutable profile limits with checked integers. Apply security, slot/probe/offline, unknown, 70/75/90 thresholds and 3/6-window hysteresis in spec priority order.

- [ ] **Step 5: Implement boot/gap/reset semantics**

New boot requires sequence1 and starts unknown. Gap>20seconds resets EWMA/counters and returns unknown before evaluating three valid windows. A required unknown immediately clears local candidate; third consecutive unknown makes health degraded.

- [ ] **Step 6: Freeze vectors and ruleset digest**

Store canonical input/output vectors for threshold edges, exact zero, counter rollback, overflow, unsupported required metric, optional fault, slot quarantine, security quarantine and 3/6 instant floods. Hash canonical rules/vectors into the compiled ruleset digest.

- [ ] **Step 7: Run reducer and deterministic vector tests**

Run: `go test ./internal/nodecontrol/observation -count=1`

Expected: PASS with byte-identical outputs across repeated runs and no instant sequence flood advancing recovery.

- [ ] **Step 8: Commit the reducer**

```bash
git add internal/nodecontrol/observation/types.go internal/nodecontrol/observation/reducer.go internal/nodecontrol/observation/ruleset.go internal/nodecontrol/observation/reducer_test.go internal/nodecontrol/observation/reducer_vectors_test.go testdata/c12/load/capacity-reducer-v1.json
git commit -m "feat: add node capacity reducer"
```

### Task B07-T02: Collect bounded agent metrics and build observations

**Files:**
- Create: `internal/nodeagent/observation/collector.go`
- Create: `internal/nodeagent/observation/cgroup_linux.go`
- Create: `internal/nodeagent/observation/report.go`
- Test: `internal/nodeagent/observation/collector_test.go`
- Test: `internal/nodeagent/observation/cgroup_linux_test.go`
- Test: `internal/nodeagent/observation/report_test.go`
- Modify: `cmd/node-agent/run.go`

**Interfaces:**
- Consumes: owned lease/PID/start token, adapter finite snapshot, immutable profile limits, monotonic sample clock and B07-T01 reducer.
- Produces: `Collector.Sample`, a strict maximum-64-KiB `ObservationV1`, local reducer output and replacement of B04's one-element report queue.

- [ ] **Step 1: Write the failing metric state/source table**

```go
func TestCollectorDoesNotInventUnsupportedMetrics(t *testing.T) {
	collector := newCollector(t, capabilitySet("cpu_basis_points", "memory_bytes"))
	report, err := collector.Sample(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if metric(report, observation.MetricEgressBPS).State != observation.MetricUnsupported {
		t.Fatal("unsupported egress was invented")
	}
	if metric(report, observation.MetricCPU).State != observation.MetricValid {
		t.Fatal("supported CPU was not sampled")
	}
}
```

- [ ] **Step 2: Run collector tests and verify RED**

Run: `go test ./internal/nodeagent/observation -run 'TestCollectorDoesNotInventUnsupportedMetrics|TestCPUBasisPoints' -count=1`

Expected: FAIL because collector package is absent.

- [ ] **Step 3: Implement exact cgroup/process sources**

Read `cpu.stat usage_usec`, `memory.current`, `pids.current`, owned process-tree open FD count and exact cgroup/rlimit readbacks. Verify PID/start token/lease ownership before each sample. Packet loss and queue depth remain unsupported for base profiles.

- [ ] **Step 4: Implement checked rates and CPU basis points**

Compute delta CPU and egress with monotonic elapsed nanoseconds, checked multiplication and round-half-up; clamp CPU at10,000. First boot sample, counter rollback, elapsed outside4–10seconds, quota `max`/mismatch and overflow yield `unknown`, never zero/default/negative.

- [ ] **Step 5: Build the closed observation**

Include boot/sequence, build/capability digests, seen/applied generation+digest, at most8 sorted slot snapshots, profile ID/version, reducer version/digest/output and finite failure reason. Strict marshal must stay at or below64KiB.

- [ ] **Step 6: Wire the 5-second sample and replacement queue**

The agent samples by monotonic timer, replaces any unsent observation and never blocks reconciler or writes historical samples to disk. Shutdown sends at most the latest value within its existing deadline.

- [ ] **Step 7: Run package and race tests**

Run: `go test ./internal/nodeagent/observation ./cmd/node-agent -run 'Test(Collector|CPU|ReportQueue|Observation)' -count=1`

Run: `go test -race ./internal/nodeagent/observation -count=1`

Expected: PASS; all values have explicit state/source/unit and queue length never exceeds one.

- [ ] **Step 8: Commit agent observation collection**

```bash
git add internal/nodeagent/observation/collector.go internal/nodeagent/observation/cgroup_linux.go internal/nodeagent/observation/report.go internal/nodeagent/observation/collector_test.go internal/nodeagent/observation/cgroup_linux_test.go internal/nodeagent/observation/report_test.go cmd/node-agent/run.go
git commit -m "feat: collect bounded node observations"
```

### Task B07-T03: Persist and independently replay authoritative observations

**Files:**
- Create: `internal/nodecontrol/observation/repository.go`
- Create: `internal/nodecontrol/observation/service.go`
- Test: `internal/nodecontrol/observation/service_test.go`
- Test: `internal/nodecontrol/observation/postgres_integration_test.go`
- Create: `internal/nodeagentapi/observations.go`
- Test: `internal/nodeagentapi/observations_test.go`
- Modify: `internal/nodeagentapi/handler.go`
- Create: `db/queries/nodecontrol_observation.sql`
- Create generated: `internal/store/nodecontrol_observation.sql.go`

**Interfaces:**
- Consumes: B03 exact active-certificate request authority, B01 observation queries/schema, B07-T01 reducer and B07-T02 strict observation DTO.
- Produces: `observation.Service.Accept(context.Context, Authority, ObservationV1) (Ack, error)`, latest snapshot/state transition/outbox writes and node-agent HTTP observation handler.

- [ ] **Step 1: Write failing duplicate/reorder/fork integration tests**

```go
func TestObservationIdempotencyAndFork(t *testing.T) {
	service := newObservationService(t)
	authority := activeNodeAuthority()
	report := validObservation(1, 1)
	first, err := service.Accept(context.Background(), authority, report)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Accept(context.Background(), authority, report)
	if err != nil || second != first {
		t.Fatalf("duplicate ack = %#v, err = %v", second, err)
	}
	if _, err := service.Accept(context.Background(), authority, mutateDigest(report)); !errors.Is(err, ErrObservationFork) {
		t.Fatalf("fork error = %v", err)
	}
}
```

- [ ] **Step 2: Run service/integration tests and verify RED**

Run: `go test ./internal/nodecontrol/observation -run TestObservationIdempotencyAndFork -count=1`

Expected: FAIL because service/repository are absent.

- [ ] **Step 3: Add exact lock/read/update queries**

Lock node inventory/latest observation, verify node/identity epoch/profile/generation/digest, insert or update only the latest `(boot_id,sequence,digest)` snapshot, and append only meaningful state transitions/outbox events. Old boot or lower sequence returns stale ACK without state mutation.

- [ ] **Step 4: Enforce server arrival cadence before reducer advance**

Rate-limit a new state-advancing report when trusted server arrival delta is below4seconds; return fixed retry-after and do not store its sequence. Gap>20seconds clears reducer state. New boot requires sequence1 and atomically supersedes old boot.

- [ ] **Step 5: Replay raw facts and only tighten accepting_new**

Load immutable profile/inventory/desired limits, run the shared reducer, compare reducer version, health, capacity-blocked and agent candidate. Mismatch records finite `agent_reducer_mismatch`, sets authority false and never treats node healthy. Final authority is the required AND expression; no branch turns agent false into true.

- [ ] **Step 6: Implement bounded HTTP handling**

Require exact active current-epoch credential and normal allowed node state at admission and commit. Strictly decode at most64KiB with no compression/unknown fields. Same report returns same ACK; same key/different digest returns conflict and security transition without raw report text.

- [ ] **Step 7: Generate and run PostgreSQL integration**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `go test -tags=integration ./internal/nodecontrol/observation ./internal/nodeagentapi -run 'TestObservation(Postgres|HTTP)' -count=1 -timeout 2m`

Expected: PASS against a run-owned migrated database; latest snapshot and transitions are atomic and raw five-second history is not appended.

- [ ] **Step 8: Run package tests and commit**

Run: `go test ./internal/nodecontrol/observation ./internal/nodeagentapi -count=1`

```bash
git add internal/nodecontrol/observation/repository.go internal/nodecontrol/observation/service.go internal/nodecontrol/observation/service_test.go internal/nodecontrol/observation/postgres_integration_test.go internal/nodeagentapi/observations.go internal/nodeagentapi/observations_test.go internal/nodeagentapi/handler.go db/queries/nodecontrol_observation.sql internal/store/nodecontrol_observation.sql.go
git commit -m "feat: reduce authoritative node observations"
```

### Task B07-T04: Enforce privacy, bounded metrics and API cardinality

**Files:**
- Create: `internal/e2e/c12_privacy_test.go`
- Create: `internal/e2e/c12_api_bounds_test.go`
- Create: `internal/nodecontrol/observation/privacy_test.go`
- Modify: `internal/observability/metrics.go`
- Modify: `internal/observability/metrics_test.go`
- Modify: `internal/errorreport/reporter.go`
- Modify: `internal/errorreport/reporter_test.go`

**Interfaces:**
- Consumes: B06 fault paths and B07 observation/API results.
- Produces: finite C1.2 metric/reporter enums, canary scan and bounded request/series acceptance used by B07 load and B10 verifier.

- [ ] **Step 1: Write the failing canary and series-count tests**

```go
func TestC12OutputsContainNoNodeCanaries(t *testing.T) {
	capture := runC12CanaryScenario(t, []string{
		"node-id-canary", "grant-canary", "cert-canary", "path-canary",
		"core-output-canary", "provider-error-canary", "credential-canary",
	})
	for _, canary := range capture.Canaries {
		if bytes.Contains(capture.AllBytes(), []byte(canary)) {
			t.Fatalf("output contains %s", canary)
		}
	}
}
```

- [ ] **Step 2: Run privacy tests and verify RED**

Run: `go test -tags=c12fixture ./internal/e2e -run 'TestC12OutputsContainNoNodeCanaries|TestC12MetricSeriesBounded' -count=1 -timeout 2m`

Expected: FAIL because C1.2 capture and finite metric registry are incomplete.

- [ ] **Step 3: Add closed metric and reporter dimensions**

Permit only component, operation, result, reason, adapter and health enums. Never label node/slot/operator/incident/certificate/boot IDs, addresses, paths, release IDs or provider strings. Fold unknown values to one finite `unknown` enum before observation/log/metric/report emission.

- [ ] **Step 4: Add wire/decoded and collection bounds**

Test exactly64KiB accepted and64KiB+1 rejected; 8 slots accepted and9 rejected; unknown/duplicate/trailing JSON rejected; oversized chunked/uncompressed bodies stop before application call; nonidentity content encoding rejected.

- [ ] **Step 5: Capture all observable surfaces**

Capture HTTP body/header, structured log, Prometheus gather, errorreport DTO, panic recovery, fixture drain byte counter and command output. Inject canaries into every forbidden input class and require zero matches without printing the matched input.

- [ ] **Step 6: Prove series count does not grow with nodes**

Gather after 1, 10, 100 and1,000 synthetic node reports; exact descriptor/series count must stay constant for the same finite state mix. Any metric containing node/slot/boot/release IDs fails the descriptor test.

- [ ] **Step 7: Run privacy and race suites**

Run: `go test -tags=c12fixture ./internal/e2e -run 'TestC12(Outputs|API|Metric)' -count=1 -timeout 5m`

Run: `go test -race ./internal/nodecontrol/observation ./internal/observability ./internal/errorreport -count=1`

Expected: PASS with zero canary and constant series cardinality.

- [ ] **Step 8: Commit privacy/boundedness**

```bash
git add internal/e2e/c12_privacy_test.go internal/e2e/c12_api_bounds_test.go internal/nodecontrol/observation/privacy_test.go internal/observability/metrics.go internal/observability/metrics_test.go internal/errorreport/reporter.go internal/errorreport/reporter_test.go
git commit -m "test: bound node control observability"
```

### Task B07-T05: Build and run the 1,000-agent reference load acceptance

**Files:**
- Create: `cmd/c12-load/main.go`
- Create: `internal/c12test/load/config.go`
- Create: `internal/c12test/load/generator.go`
- Create: `internal/c12test/load/statistics.go`
- Create: `internal/c12test/load/report.go`
- Test: `cmd/c12-load/main_test.go`
- Test: `internal/c12test/load/generator_test.go`
- Test: `internal/c12test/load/statistics_test.go`
- Test: `internal/c12test/load/adversarial_test.go`
- Create: `testdata/c12/load/reference-runner.v1.json`
- Create: `testdata/c12/load/report-schema.v1.json`

**Interfaces:**
- Consumes: real B03 agent mTLS endpoints, B02 desired publish, B07 observation handler/reducer and a digest-locked reference runner.
- Produces: `c12-load`, canonical `C12LoadReportV1`, nearest-rank latency/resource results and B07 authoritative/diagnostic classification used by B10.

- [ ] **Step 1: Write failing schedule/statistics tests**

```go
func TestLoadScheduleAndNearestRank(t *testing.T) {
	config := load.ReferenceConfig()
	schedule := load.BuildSchedule(config)
	if len(schedule.Agents) != 1_000 || schedule.Seed != 0xC12A6E17 {
		t.Fatalf("schedule = %#v", schedule)
	}
	values := []time.Duration{time.Second, 2 * time.Second, 3 * time.Second, 4 * time.Second}
	if got := load.NearestRank(values, 95); got != 4*time.Second {
		t.Fatalf("p95 = %s", got)
	}
}
```

- [ ] **Step 2: Run load unit tests and verify RED**

Run: `go test ./cmd/c12-load ./internal/c12test/load -run 'TestLoadScheduleAndNearestRank|TestReferenceProfile' -count=1`

Expected: FAIL because load packages are absent.

- [ ] **Step 3: Freeze reference profile and authoritative classification**

Require Linux amd64, 8 dedicated vCPU,16GiB, locked Docker/Go/PostgreSQL/image digests and exact container limits. Mismatch sets report `classification=diagnostic` before opening agent connections; only an exact profile can emit `authoritative`.

- [ ] **Step 4: Implement the deterministic agent schedule**

Create1,000 isolated identities, evenly stagger first poll over25seconds, poll for25seconds, report every5seconds, warm up120seconds, measure600seconds and publish one generation at measurement second60. ACK time is the control-plane commit time of an observation whose applied generation equals target; missing after30seconds is infinite.

- [ ] **Step 5: Implement latency and resource assertions**

Sort all1,000 results and use nearest rank `ceil(p*N)`. Require p95<10seconds, p99<20seconds and100% finite by30seconds. Sample heap/goroutines every5seconds; baseline is the last12 warm-up samples' median; OLS projected300-second growth must be at most5%, heap<2GiB and goroutines<baseline+2,500.

- [ ] **Step 6: Implement the three adversarial phases**

Flood time nonce and prove pre-provider rate limit bounds provider/signer calls while normal poll p99 remains<20seconds. Drive999 other nodes and prove selected-node guard/latch/checkpoint unchanged with zero ordinary report/attestation counter writes. Burst3/6 sequences and prove only reports separated by at least4seconds advance hysteresis.

- [ ] **Step 7: Run unit/adversarial tests**

Run: `go test ./cmd/c12-load ./internal/c12test/load -count=1`

Run: `go test -race ./internal/c12test/load -count=1`

Expected: PASS with deterministic schedule/report bytes and exact authoritative/diagnostic classification.

- [ ] **Step 8: Run the reference acceptance on an eligible runner**

Run:

```powershell
$loadOutput = Join-Path ([System.IO.Path]::GetTempPath()) 'talenro-c12-load-report-v1.json'
go run ./cmd/c12-load --profile testdata/c12/load/reference-runner.v1.json --output $loadOutput
```

Expected: exit0 with `classification=authoritative`,1,000 ACKs, p95<10s, p99<20s,100% within30s, resource growth within bounds and all three adversarial phases passing. On an ineligible local runner the command exits3 after writing a `diagnostic` report and makes no completion claim.

- [ ] **Step 9: Commit load acceptance**

```bash
git add cmd/c12-load/main.go cmd/c12-load/main_test.go internal/c12test/load/config.go internal/c12test/load/generator.go internal/c12test/load/statistics.go internal/c12test/load/report.go internal/c12test/load/generator_test.go internal/c12test/load/statistics_test.go internal/c12test/load/adversarial_test.go testdata/c12/load/reference-runner.v1.json testdata/c12/load/report-schema.v1.json
git commit -m "test: add node control load acceptance"
```

## B07 exit gate

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1
go test ./internal/nodecontrol/observation ./internal/nodeagent/observation ./internal/nodeagentapi ./internal/observability ./internal/errorreport ./cmd/c12-load ./internal/c12test/load -count=1
go test -tags=c12fixture ./internal/e2e -run 'TestC12(Outputs|API|Metric)' -count=1 -timeout 5m
go test -race ./internal/nodecontrol/observation ./internal/nodeagent/observation ./internal/c12test/load -count=1
go vet ./internal/nodecontrol/observation ./internal/nodeagent/observation ./internal/nodeagentapi ./cmd/c12-load
go tool golangci-lint run ./internal/nodecontrol/observation ./internal/nodeagent/observation ./internal/nodeagentapi ./cmd/c12-load
git diff --check
```

Expected: all local gates pass. B07 is authoritative only when the separately executed reference-runner command returns exit0 and an exact-profile `authoritative` report; local resource mismatch remains diagnostic. Neither result claims the real Linux platform/memory/provider gate reserved for分册09.
