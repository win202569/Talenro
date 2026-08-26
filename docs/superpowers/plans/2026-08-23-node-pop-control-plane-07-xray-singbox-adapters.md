# Talenro C1.2 Xray and sing-box Adapters Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付彼此独立的 Xray 与 sing-box fixed-profile adapter、受审查 release lock、真实 loopback protocol smoke，并证明核心始终是 run-owned external release volume 中的独立 OS 进程。

**Architecture:** Agent 只注册类型化、无 secret 的 preview profile；supervisor 从同一 signed process spec、批准 release manifest 与 single-use credential FD 确定性编译 server/client config，执行原生 check，再在隔离 slot namespace 中启动核心。B08 与 B09 各自拥有 release、license、RED/GREEN、真实握手和提交；公共 smoke result 只汇总有限证据，不能让一个 adapter 的成功冒充另一个。

**Tech Stack:** Go 1.26.5、标准库 `encoding/json`、Xray `run -test`、sing-box `check`、Linux cgroup v2/pidfd/seccomp/nftables、OCI image layout、Docker Engine、RFC 8785 JCS、SHA-256。

**Spec:** [Approved C1.2 node and POP control-plane design](../specs/2026-08-23-node-pop-control-plane-design.md), especially sections 11.1–11.3, 13.2–13.5, 17.5, 20, and 21 batches 8–9.

## Global Constraints

- 本册只实现 `C1.2-B08` 与 `C1.2-B09`；必须先完成分册 01–06。
- Xray profile 只允许 `xray-loopback-vless-tcp-v1`；sing-box profile 只允许 `sing-box-loopback-hysteria2-v1`。
- Xray 与 sing-box 必须拥有不同 release ID、source OCI digest、server/client executable digest、dependency closure、profile、license record、smoke result 与 commit。
- Desired state、OpenAPI、PostgreSQL 与 NATS 不得出现 path、argv、environment、原始 core JSON、test credential 或 production credential。
- Agent renderer 只能返回 secret-free preview；只有 supervisor compiler 能读取 1–4096 byte、sealed、single-use、role-bound credential FD。
- Native check、server core、ephemeral client 与 echo endpoint 都是独立受限 process/cgroup；成功和失败路径都必须证明 cgroup empty。
- Xray 原生检查固定使用 `run -test`；sing-box 原生检查固定使用 `check`。所有 argv 来自 build-pinned manifest，不从 desired state、环境或 PATH 取得。
- Xray loopback 握手 deadline 固定 5 秒；sing-box 固定 10 秒；成功必须经隧道往返匹配本次运行的 32-byte random nonce。
- 仅端口打开、process alive、config check 成功或 agent 自报 healthy 都不能满足 real-core smoke。
- Docker 只是 Linux node-host test substrate；`DeterministicHostMemoryIsolationPolicyFakeV1` 结果不能提升为 SELinux、host sysctl、secure-time 或 production anti-rollback 证据。
- Core OCI layout、executable 与 dependency 不得进入 Git、三个 proprietary release binary、production image 或 `node-agent-test-host` image。
- Xray-core 按 MPL-2.0、sing-box 按 GPLv3+ 分开记录；process boundary 不是法律结论，任一版本、来源、patch、argv 或 image 变化都重触发审查。
- Production target 仍仅为 `linux/amd64`；Windows Job Object fixture 不能满足本册 real-core gate。
- 普通 Go build/test 使用 `CGO_ENABLED=0`，race 使用 `CGO_ENABLED=1`；不得 import `github.com/xtls/xray-core`、`github.com/sagernet/sing-box` 或子路径。
- 每个 task 只 stage 明列路径；不得使用 `git add .`、`git add -A`，不得触及 `.cache/`、`.superpowers/`、`.task19-go/`。

---

## File structure and frozen adapter surfaces

```text
internal/nodeagent/adapter/
├── registry.go                                  # 既有 closed adapter/profile registry
├── xray/profile.go                              # Xray typed preview and fixed endpoints
├── xray/profile_test.go
├── singbox/profile.go                           # sing-box typed preview and fixed endpoints
└── singbox/profile_test.go
internal/nodesupervisor/adapter/
├── compiler.go                                  # 既有 supervisor-only compiler registry
├── xray.go                                      # Xray server/client deterministic JSON
├── xray_test.go
├── singbox.go                                   # sing-box server/client deterministic JSON
└── singbox_test.go
internal/c12acceptance/coresmoke/
├── result.go                                    # adapter-bound finite smoke result
├── result_test.go
├── xray_test.go                                 # real Xray loopback tagged test
└── singbox_test.go                              # real sing-box loopback tagged test
testdata/c12/releases/
├── xray-approved-v1.lock.json                   # generated from reviewed external OCI layout
└── sing-box-approved-v1.lock.json
testdata/c12/adapters/
├── xray-preview-v1.json
└── sing-box-preview-v1.json
docs/licenses/
├── c12-xray-core.md
└── c12-sing-box.md
```

This plan consumes these exact interfaces from batches 4–6. The declarations below are normative references, not new ownership: P06 uniquely owns `internal/nodeagent/adapter` `ProfileV1`/`PreviewV1`/`Register`/`Preview`, P05 uniquely owns `internal/nodesupervisor/adapter` compile types and `RegisterCompiler`, and P04 owns only its narrower verified preview wrapper. No file in this plan redeclares or replaces them.

```go
package adapter

type ProfileV1 struct {
	Adapter             contracts.Adapter
	ProfileID           string
	ServerListen        netip.AddrPort
	ClientListen        netip.AddrPort
	EchoTarget          netip.AddrPort
	NativeCheckDeadline time.Duration
	StartupDeadline     time.Duration
	HandshakeDeadline   time.Duration
}

type PreviewV1 struct {
	Adapter        contracts.Adapter
	ProfileID      string
	ReleaseID      string
	SemanticDigest contracts.Digest
	EndpointSet    []netip.AddrPort
}

func Register(ProfileV1) error
func Preview(contracts.ProcessSpecV1) (PreviewV1, error)
```

Supervisor compilation consumes only the sealed FD reader and returns bytes that remain supervisor-owned:

```go
package adapter

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

type CompileFunc func(CompileRequestV1) (CompiledConfigV1, error)

func RegisterCompiler(profileID string, compile CompileFunc) error
```

### Task 1: B08 Xray typed profile and supervisor compiler

**Files:**
- Create: `internal/nodeagent/adapter/xray/profile.go`
- Create: `internal/nodeagent/adapter/xray/profile_test.go`
- Create: `internal/nodesupervisor/adapter/xray.go`
- Create: `internal/nodesupervisor/adapter/xray_test.go`
- Create: `internal/c12acceptance/adapter_surface_test.go`
- Create: `testdata/c12/adapters/xray-preview-v1.json`
- Test: `internal/nodeagent/adapter/xray/profile_test.go`
- Test: `internal/nodesupervisor/adapter/xray_test.go`

**Interfaces:**
- Consumes: `adapter.Register(ProfileV1) error`, `adapter.Preview(contracts.ProcessSpecV1) (PreviewV1, error)`, `nodesupervisoradapter.RegisterCompiler(string, CompileFunc) error`, and the exact `ProfileV1`/`CompileRequestV1` types frozen above.
- Produces: `xray.Profile() adapter.ProfileV1`, `xray.CompileServer(adapter.CompileRequestV1) (adapter.CompiledConfigV1, error)`, and `xray.CompileClient(adapter.CompileRequestV1) (adapter.CompiledConfigV1, error)`.

An external-package test compile-assigns the P06 agent and P05 supervisor declarations/method expressions, registers Xray through both sole registries, rejects duplicate/replacement/case aliases, and statically scans the new packages for shadow declarations named `ProfileV1`、`PreviewV1`、`CompileRequestV1` or `CompiledConfigV1`.

- [ ] **Step 1: RED — add the closed Xray profile test**

```go
func TestProfileIsExactAndSecretFree(t *testing.T) {
	t.Parallel()
	p := Profile()
	if p.Adapter != contracts.AdapterXray || p.ProfileID != "xray-loopback-vless-tcp-v1" {
		t.Fatal("xray profile identity mismatch")
	}
	if p.ServerListen.String() != "127.0.0.1:32081" ||
		p.ClientListen.String() != "127.0.0.1:32082" ||
		p.EchoTarget.String() != "127.0.0.1:32080" ||
		p.HandshakeDeadline != 5*time.Second {
		t.Fatal("xray profile bounds mismatch")
	}
	preview, err := adapter.Preview(validXrayProcessSpec())
	if err != nil {
		t.Fatal("xray preview rejected valid process spec")
	}
	encoded, err := json.Marshal(preview)
	if err != nil || bytes.Contains(encoded, []byte("credential")) || bytes.Contains(encoded, []byte("uuid")) {
		t.Fatal("xray preview exposed credential material")
	}
}
```

- [ ] **Step 2: Run the focused profile test and verify RED**

Run: `go test ./internal/nodeagent/adapter/xray -run '^TestProfileIsExactAndSecretFree$' -count=1`

Expected: FAIL because package `internal/nodeagent/adapter/xray` and `Profile` do not exist.

- [ ] **Step 3: GREEN — register the immutable Xray profile**

```go
const ProfileID = "xray-loopback-vless-tcp-v1"

func Profile() adapter.ProfileV1 {
	return adapter.ProfileV1{
		Adapter:             contracts.AdapterXray,
		ProfileID:           ProfileID,
		ServerListen:        netip.MustParseAddrPort("127.0.0.1:32081"),
		ClientListen:        netip.MustParseAddrPort("127.0.0.1:32082"),
		EchoTarget:          netip.MustParseAddrPort("127.0.0.1:32080"),
		NativeCheckDeadline: 5 * time.Second,
		StartupDeadline:     5 * time.Second,
		HandshakeDeadline:   5 * time.Second,
	}
}
```

Register exactly once from `registry.go`; reject duplicate profile ID, adapter mismatch, non-loopback endpoint, zero deadline and any endpoint outside the three values above.

- [ ] **Step 4: Run the profile package and verify GREEN**

Run: `go test ./internal/nodeagent/adapter/xray -count=1`

Expected: PASS, including wrong adapter, unknown profile, remote endpoint and preview-canary cases.

- [ ] **Step 5: RED — add server/client compiler tests**

```go
func TestCompileRequiresExactRoleAndSingleCredential(t *testing.T) {
	t.Parallel()
	server := validXrayCompileRequest(adapter.CredentialRoleServer, `{"schema_version":"xray-smoke-credential.v1","vless_uuid":"018f25e8-6cb0-7d3a-9e70-123456789abc"}`)
	compiled, err := CompileServer(server)
	if err != nil {
		t.Fatal("xray server compile rejected valid credential")
	}
	if compiled.Listen.String() != "127.0.0.1:32081" || compiled.Target.String() != "127.0.0.1:32080" {
		t.Fatal("xray server compile changed fixed endpoints")
	}
	server.Credential = newTestCredentialReader(adapter.CredentialRoleClient, `{"schema_version":"xray-smoke-credential.v1","vless_uuid":"018f25e8-6cb0-7d3a-9e70-123456789abc"}`)
	if _, err := CompileServer(server); !errors.Is(err, adapter.ErrCredentialRole) {
		t.Fatal("xray server accepted client credential role")
	}
}
```

- [ ] **Step 6: Run compiler tests and verify RED**

Run: `go test ./internal/nodesupervisor/adapter -run '^TestCompile.*Xray|^TestXray.*' -count=1`

Expected: FAIL because the Xray compiler functions are undefined.

- [ ] **Step 7: GREEN — implement deterministic Xray JSON compilation**

Use typed structs and `json.Marshal`; reject unknown/duplicate JSON members, invalid UUID, trailing bytes, a second read, input outside 1–4096 bytes, role mismatch, release ID other than `xray-approved-v1`, profile mismatch, zero node/generation and non-canonical slot ID. Emit fixed argv metadata through the approved manifest only:

```go
var XrayArgv = struct {
	Check  []string
	Start  []string
	Client []string
}{
	Check:  []string{"run", "-test", "-config", "/run/talenro/config.json"},
	Start:  []string{"run", "-config", "/run/talenro/config.json"},
	Client: []string{"run", "-config", "/run/talenro/client.json"},
}
```

The server config exposes VLESS only on `127.0.0.1:32081` and routes to `127.0.0.1:32080`; the client exposes SOCKS only on `127.0.0.1:32082` and routes solely through that VLESS server. Disable access/error logs and API/stats listeners. No field accepts a URL, path from input, DNS name, environment value or alternate outbound.

- [ ] **Step 8: Run compiler and registry tests and verify GREEN**

Run: `go test ./internal/nodeagent/adapter/xray ./internal/nodesupervisor/adapter -count=1`

Expected: PASS; canonical JSON digest is stable across 100 runs, no synthetic credential appears in preview or error text, and role replay is rejected.

- [ ] **Step 9: REFACTOR — format and run affected race/static gates**

Run: `gofmt -w internal/nodeagent/adapter/xray internal/nodesupervisor/adapter/xray.go internal/nodesupervisor/adapter/xray_test.go`

Run: `go test -race ./internal/nodeagent/adapter/xray ./internal/nodesupervisor/adapter -count=1`

Expected: PASS with no race report.

- [ ] **Step 10: Commit B08 compiler independently**

```bash
git add internal/nodeagent/adapter/xray/profile.go internal/nodeagent/adapter/xray/profile_test.go internal/nodesupervisor/adapter/xray.go internal/nodesupervisor/adapter/xray_test.go internal/c12acceptance/adapter_surface_test.go testdata/c12/adapters/xray-preview-v1.json
git commit -m "feat: add fixed Xray loopback adapter"
```

### Task 2: B08 Xray release review and real loopback smoke exit

**Files:**
- Create: `cmd/talenro-core-lock/main.go`
- Create: `cmd/c12-core-smoke-runner/main.go`
- Create: `cmd/c12-core-smoke-broker/main.go`
- Create: `internal/coreartifactlock/lock.go`
- Create: `internal/coreartifactlock/deps.go`
- Test: `internal/coreartifactlock/lock_test.go`
- Test: `internal/coreartifactlock/deps_test.go`
- Create: `internal/c12test/coresmokerunner/types.go`
- Create (first line `//go:build linux`): `internal/c12test/coresmokerunner/runner_linux.go`
- Create (first line `//go:build linux`): `internal/c12test/coresmokerunner/channel_linux.go`
- Create (first line `//go:build !linux`): `internal/c12test/coresmokerunner/runner_stub.go`
- Test: `internal/c12test/coresmokerunner/runner_test.go`
- Test (first line `//go:build linux`): `internal/c12test/coresmokerunner/channel_linux_test.go`
- Create: `internal/c12test/coresmokebroker/protocol.go`
- Create (first line `//go:build linux`): `internal/c12test/coresmokebroker/broker_linux.go`
- Create (first line `//go:build !linux`): `internal/c12test/coresmokebroker/broker_stub.go`
- Test: `internal/c12test/coresmokebroker/protocol_test.go`
- Test (first line `//go:build linux`): `internal/c12test/coresmokebroker/broker_linux_test.go`
- Create: `scripts/run-c12-core-smoke.sh`
- Test: `scripts/tests/run-c12-core-smoke.Tests.ps1`
- Create: `deploy/c12-core-smoke/Dockerfile`
- Create: `deploy/c12-core-smoke/.dockerignore`
- Create: `testdata/c12/core-smoke-runner.v1.json`
- Create: `testdata/c12/releases/xray-approved-v1.lock.json`
- Modify: `internal/localrelease/approved-release-manifest.v1.json`
- Create: `internal/c12acceptance/coresmoke/harness.go`
- Create: `internal/c12acceptance/coresmoke/result.go`
- Create: `internal/c12acceptance/coresmoke/receipts.go`
- Test: `internal/c12acceptance/coresmoke/result_test.go`
- Test: `internal/c12acceptance/coresmoke/receipts_test.go`
- Create (first line `//go:build c12_core_smoke && linux`): `internal/c12acceptance/coresmoke/xray_test.go`
- Create: `docs/licenses/c12-xray-core.md`
- Test: `internal/c12acceptance/coresmoke/xray_test.go`

**Interfaces:**
- Consumes: `xray.Profile()`, supervisor `Prepare/Check/Start/Probe/Drain/Stop`, B02 canonical local-release verifier/build-embedded manifest source, B06 `c12test.CoreSmokeCredentialProvider.New(...)(CoreSmokeCredentialSession,error)`, and the already frozen B06 `run --config-fd=3` challenge-echo endpoint. It does not modify the fixture executable or add an argv/mode.
- Produces: the offline deterministic `talenro-core-lock` (`lock-core` and `verify-go-deps` only), the independent Linux `run-c12-core-smoke.sh` substrate, helper-authenticated `CoreSmokeBuildReceiptV1`, the common harness and canonical receipt/result types plus validator later frozen in this task, tagged `TestXrayLoopbackSmoke`, release ID `xray-approved-v1`, license record `docs/licenses/c12-xray-core.md`, and—only in the parent after cleanup—an adapter-bound validated `CoreSmokeResultV1` with `Adapter=xray`.

- [ ] **Step 1: RED — write the Xray real-smoke test**

```go
func TestXrayLoopbackSmoke(t *testing.T) {
	lock := requireReleaseLock(t, "xray-approved-v1")
	provisional, err := runRealCoreSmoke(t, SmokeRequest{
		Adapter:           contracts.AdapterXray,
		ProfileID:         "xray-loopback-vless-tcp-v1",
		ReleaseLock:       lock,
		NonceBytes:        32,
		HandshakeDeadline: 5 * time.Second,
	})
	if err != nil {
		t.Fatal("xray live smoke failed")
	}
	if err := handoffOpaqueProvisionalCoreSmoke(t, provisional); err != nil {
		t.Fatal("xray provisional handoff failed")
	}
}
```

The tagged child can observe only the live challenge/process/volume/credential facts through inherited FD 9 and hand the one-use opaque handle back over that same authenticated broker session. It has no final result accessor and makes no cleanup/cgroup-absence assertion；the parent helper performs every final result assertion only after test exit and its cleanup barrier.

- [ ] **Step 2: Run the tagged test inside the approved Linux test-host and verify RED**

Run through the P07-owned Linux substrate:

```text
bash scripts/run-c12-core-smoke.sh --phase candidate --adapter xray --expect release-lock-unavailable --timeout 2m
```

Expected: FAIL before child creation with `release lock unavailable`.

- [ ] **Step 3: Generate the exact Xray release lock from the reviewed external OCI layout**

Run in the isolated review runner:

```text
go run ./cmd/talenro-core-lock lock-core --adapter xray --release-id xray-approved-v1 --profile xray-loopback-vless-tcp-v1 --source-layout /run/talenro-c12/review/xray-approved-v1.oci --write testdata/c12/releases/xray-approved-v1.lock.json
```

Expected: exit 0 and one canonical lock containing exact upstream version/source URI/image manifest digest, server/client executable and dependency SHA-256, the three fixed argv arrays, provenance digest, vulnerability result digest, MPL-2.0 license record digest and sandbox profile. In the same fixed two-artifact transaction, the tool regenerates B02's package-local `internal/localrelease/approved-release-manifest.v1.json` from the immutable fixture entry plus every then-present reviewed core lock, increments its manifest version and verifies its domain-separated canonical digest. `talenro-core-lock` is repository-owned、offline、network-disabled and digest/toolchain-pinned；it accepts only an immutable canonical-absolute reviewed OCI input plus one of two literal caller-selectable repository lock tokens: `testdata/c12/releases/xray-approved-v1.lock.json` or `testdata/c12/releases/sing-box-approved-v1.lock.json`. The derived embedded-manifest target is fixed in the binary and has no CLI override. It resolves all targets beneath a no-follow helper-authenticated repository root, stages/fsyncs both candidates, and failure-injection tests prove recovery leaves either the exact old pair or exact new pair；an incomplete review exits 1 without changing either. At this task the embedded set is fixture+Xray and is valid only for the test profile；`RequireC12ProductionReleaseSet` must still reject it until Task 4 adds sing-box. P08 later consumes and exact-covers both locks and the final embedded manifest；P07 does not call the P08 scanner and therefore has no forward dependency.

- [ ] **Step 4: GREEN — connect the fixture to the Xray real-core path**

Use only the B06 fixture's existing `run --config-fd=3` challenge-echo configuration. The smoke harness generates a random 32-byte challenge and obtains one role-bound VLESS `CoreSmokeCredentialSession`, calls `Prepare → Check → Start → Probe`, compares the returned challenge, snapshots finite health/capacity, calls `Drain → Stop`, consumes the credential cleanup receipt, and asserts all three exact owned cgroups and handles are absent. It never passes lock/map/output paths to the fixture and never prints core stdout/stderr、config、challenge or credential.

`run-c12-core-smoke.sh` accepts only closed adapter、expectation、timeout and `--phase candidate|authoritative`, then launches the deployment-profile hash-checked `c12-core-smoke-runner`; the Go helper, not shell, owns the substrate and all receipt authority. Candidate phase may use only the current exact task-path worktree for RED/GREEN proof, marks every result candidate/non-authoritative and cannot close a gate. Authoritative phase first requires those exact paths present in a clean committed tree, captures its commit/tree digest itself, rejects staged/worktree/untracked deltas, builds only that tree and binds the digest into every receipt. The tracked `core-smoke-runner.v1.json` pins only the Linux/amd64/cgroup-v2/kernel/Docker/toolchain identities, closed package-root/dependency source-selection policy and deterministic build/runtime recipe (trimpath/build-id/time/order/uid/gid/mode/compression)；it contains neither a materialized source list/digest nor any final broker、test binary、`test2json` or test-host image hash that would go stale when a later task changes the selected source. Before any build or container, the helper creates a random `0700` run root, an authenticated append-only fsynced ownership WAL and two distinct empty build roots, recording each intent/actual. It applies the locked selection policy to the exact phase-authorized tree, rejects any selected path outside the closed package roots or undeclared generated/vendor input, freezes the resulting exact path/mode/blob closure digest in the build receipt, materializes the recipe twice in those roots, and requires byte-identical complete outputs. It then mints `CoreSmokeBuildReceiptV1` with the exact commit/tree/recipe/toolchain/source-closure plus actual `c12-core-smoke-broker`, `c12-core-smoke.test` (`go test -c -tags=c12_core_smoke`), pinned `test2json` and test-host image/config/rootfs digests. The receipt is authenticated by that helper session/WAL, has no caller-digest constructor and is the sole materialized-source/output-hash authority. The helper copies only the receipt-matched runtime artifacts into the immutable test-host context, then exact-cleans both build roots and appends/verifies their absence before freezing the receipt's `OwnershipWALPrefixDigest`; final run-root absence therefore covers no hidden build child. Candidate receipts are non-reusable；an authoritative receipt requires the clean committed tree. The source-free nonroot runtime layer contains only the receipt-matched artifacts plus fixed CA/config data；it never executes `go test` or carries Go source. The helper binds the build-receipt digest into every process/final receipt and P08 input, then creates/WAL-records a uniquely named private network、external-core volume、delegated cgroup subtree、exact source/init containers and one exact-ID test-host container. The init container copies only digest-verified server/client executables and dependency closure into the empty volume; the test-host mounts that volume read-only, has a read-only root/private mount+network namespace and only the exact delegated cgroup. The helper generates the installed map from the independently verified embedded manifest/release locks, verifies it through the same B02 canonical contract, and starts the broker; the broker directly runs the precompiled test with FD9 and streams its `-test.v=test2json` output into the separately receipt-matched `test2json` converter, so no shell/intermediate process is inserted on the FD path. The host helper parses the canonical JSON events and requires exactly one terminal pass with zero skip/missing/extra events.

The host helper never attempts to pass a native FD through Docker. Instead it bind-mounts one uniquely created run-owned socket directory and a read-only one-use broker bootstrap secret into the exact test-host container. The broker creates the private `SOCK_SEQPACKET` socketpair inside the container/PID namespace, directly `exec`s the precompiled test binary with one end as fixed inherited FD 9, and retains the other. Test and broker validate local `SO_PEERCRED` plus PID/start-token/image/run/adapter/session binding. The broker exposes only a separate Unix control socket in that run-owned directory to the host helper；their closed request/response protocol authenticates container actual ID、run/adapter/session、monotonic sequence and transcript with the one-use secret, and relays only opaque provisional-handle/readback requests—not receipt preimages or FD numbers. After mutual authentication the secret file is unlinked and locked copies are zeroized at close. `NewInheritedLiveReceiptSource(9,ExpectedRunBinding)` is the sole test constructor；there is no path、bytes、digest or environment-value constructor. Malformed/replayed/cross-run frames、wrong broker/helper/test peer、socket/FD substitution、child crash and response loss cannot create a result. Broker tests run inside a Linux namespace fixture and host-helper tests mock only the authenticated control protocol, never claim Docker FD inheritance.

After the test passes, the core-smoke runner parent reads back process image/volume/cgroup/credential identities, stops/removes only WAL-matching resources (source/init/test-host containers、network、external volume、three cgroups、credential handles), removes and verifies absence of the exact run root, and appends final absence records. It then finalizes the opaque provisional handle against the final WAL and constructs/validates the complete receipt set in locked memory. In ordinary standalone P07 mode only, it may then atomically materialize a UUID-named result in a newly created separate evidence root and print the path last. When the exact runner itself is a P08 native verifier child with the Task 5 protected slot on inherited FD 3, that terminal sink is mutually exclusive: the same finalizer writes/seals the bounded protected slot, returns no result/source to the child, creates no evidence root and prints no path；only the P08 native parent may one-use adopt and revalidate. No caller path、Docker ID、port、cgroup or result path is accepted. Any failure leaves no evidence root/publishable result and cannot close B08/B09.

Before Step 6, implement the Task 5 receipt/result schemas and two-phase live-source/finalizer exactly as frozen below, plus the common `runRealCoreSmoke` harness. Xray's test may hand off only an opaque provisional handle over FD 9；the parent helper alone may turn it into a full `ValidatedCoreSmokeReceiptSetV1` after its cleanup barrier. A bare boolean/digest result is never publishable. Task 5 later adds the two-adapter aggregate/non-substitution matrix and the protected cross-process sink/adopter；it does not introduce the semantic validator after an adapter has already closed.

- [ ] **Step 5: Write the Xray license record with fixed evidence headings**

`docs/licenses/c12-xray-core.md` must contain exact sections `Reviewed release`, `OCI/source provenance`, `MPL-2.0 covered files`, `Source availability`, `Notices`, `Vulnerability result`, `Executable/dependency digests`, `Distribution boundary`, and `Re-review triggers`. Each section records the digest copied from the generated lock; the document must state that the OCI layout is verifier-only test inventory and is not a proprietary release input.

- [ ] **Step 6: Run the tagged Xray smoke and verify GREEN**

Run: `bash scripts/run-c12-core-smoke.sh --phase candidate --adapter xray --expect pass --timeout 2m`.

Expected: candidate PASS within 2 minutes with only fixed enums/digests/timestamps, `nonce_matched=true`, native check success, healthy probe, process image in the run-owned external-core volume and three empty-cgroup observations. It is not an authoritative/published gate result before Step 8 commit and Step 9 rerun.

- [ ] **Step 7: REFACTOR — run Xray package, privacy and dependency checks**

Run: `go test ./internal/nodeagent/adapter/xray ./internal/nodesupervisor/adapter ./internal/c12acceptance/coresmoke -count=1`

Run: `go list -deps -json ./cmd/node-agent ./cmd/node-core-supervisor | go run ./cmd/talenro-core-lock verify-go-deps --deny-module github.com/xtls/xray-core --deny-module github.com/sagernet/sing-box`

Expected: both commands PASS and no core module is in either proprietary dependency graph.

- [ ] **Step 8: Commit B08 real-core acceptance independently**

```bash
git add cmd/talenro-core-lock/main.go cmd/c12-core-smoke-runner/main.go cmd/c12-core-smoke-broker/main.go internal/coreartifactlock/lock.go internal/coreartifactlock/deps.go internal/coreartifactlock/lock_test.go internal/coreartifactlock/deps_test.go internal/c12test/coresmokerunner/types.go internal/c12test/coresmokerunner/runner_linux.go internal/c12test/coresmokerunner/channel_linux.go internal/c12test/coresmokerunner/runner_stub.go internal/c12test/coresmokerunner/runner_test.go internal/c12test/coresmokerunner/channel_linux_test.go internal/c12test/coresmokebroker/protocol.go internal/c12test/coresmokebroker/broker_linux.go internal/c12test/coresmokebroker/broker_stub.go internal/c12test/coresmokebroker/protocol_test.go internal/c12test/coresmokebroker/broker_linux_test.go scripts/run-c12-core-smoke.sh scripts/tests/run-c12-core-smoke.Tests.ps1 deploy/c12-core-smoke/Dockerfile deploy/c12-core-smoke/.dockerignore testdata/c12/core-smoke-runner.v1.json testdata/c12/releases/xray-approved-v1.lock.json internal/localrelease/approved-release-manifest.v1.json internal/c12acceptance/coresmoke/harness.go internal/c12acceptance/coresmoke/result.go internal/c12acceptance/coresmoke/receipts.go internal/c12acceptance/coresmoke/result_test.go internal/c12acceptance/coresmoke/receipts_test.go internal/c12acceptance/coresmoke/xray_test.go docs/licenses/c12-xray-core.md
git commit -m "test: accept Xray as an external core"
```

- [ ] **Step 9: Run the owning Xray gate from the new clean commit**

Immediately require the Task 2 exact paths/index/worktree/untracked state clean, then run `bash scripts/run-c12-core-smoke.sh --phase authoritative --adapter xray --expect pass --timeout 2m`.

Expected: PASS and a post-cleanup result bound to the helper-captured new commit/tree digest. Do not begin Task 3 until this authoritative real Xray handshake has passed；the Step 6 candidate、unit-only or config-check-only result does not close B08.

### Task 3: B09 sing-box typed profile and supervisor compiler

**Files:**
- Create: `internal/nodeagent/adapter/singbox/profile.go`
- Create: `internal/nodeagent/adapter/singbox/profile_test.go`
- Create: `internal/nodesupervisor/adapter/singbox.go`
- Create: `internal/nodesupervisor/adapter/singbox_test.go`
- Modify: `internal/c12acceptance/adapter_surface_test.go`
- Create: `testdata/c12/adapters/sing-box-preview-v1.json`
- Test: `internal/nodeagent/adapter/singbox/profile_test.go`
- Test: `internal/nodesupervisor/adapter/singbox_test.go`

**Interfaces:**
- Consumes: the same closed `ProfileV1`, `CompileRequestV1` and registry signatures frozen at the top of this plan; it does not consume any Xray config type or release lock.
- Produces: `singbox.Profile() adapter.ProfileV1`, `singbox.CompileServer(adapter.CompileRequestV1) (adapter.CompiledConfigV1, error)`, and `singbox.CompileClient(adapter.CompileRequestV1) (adapter.CompiledConfigV1, error)`.

Extend the external surface test to register sing-box through the same P05/P06 registries, prove Xray and sing-box cannot replace each other's profile/compiler, and repeat the shadow-declaration scan.

- [ ] **Step 1: RED — add exact sing-box profile tests**

```go
func TestProfileIsExactAndDistinctFromXray(t *testing.T) {
	t.Parallel()
	p := Profile()
	if p.Adapter != contracts.AdapterSingBox || p.ProfileID != "sing-box-loopback-hysteria2-v1" {
		t.Fatal("sing-box profile identity mismatch")
	}
	if p.ServerListen.String() != "127.0.0.1:32181" ||
		p.ClientListen.String() != "127.0.0.1:32182" ||
		p.EchoTarget.String() != "127.0.0.1:32180" ||
		p.HandshakeDeadline != 10*time.Second {
		t.Fatal("sing-box profile bounds mismatch")
	}
	if p.ProfileID == xray.ProfileID || p.ServerListen == xray.Profile().ServerListen {
		t.Fatal("sing-box profile aliases Xray")
	}
}
```

- [ ] **Step 2: Run the focused profile test and verify RED**

Run: `go test ./internal/nodeagent/adapter/singbox -run '^TestProfileIsExactAndDistinctFromXray$' -count=1`

Expected: FAIL because package `internal/nodeagent/adapter/singbox` and `Profile` do not exist.

- [ ] **Step 3: GREEN — register the immutable sing-box profile**

```go
const ProfileID = "sing-box-loopback-hysteria2-v1"

func Profile() adapter.ProfileV1 {
	return adapter.ProfileV1{
		Adapter:             contracts.AdapterSingBox,
		ProfileID:           ProfileID,
		ServerListen:        netip.MustParseAddrPort("127.0.0.1:32181"),
		ClientListen:        netip.MustParseAddrPort("127.0.0.1:32182"),
		EchoTarget:          netip.MustParseAddrPort("127.0.0.1:32180"),
		NativeCheckDeadline: 10 * time.Second,
		StartupDeadline:     10 * time.Second,
		HandshakeDeadline:   10 * time.Second,
	}
}
```

- [ ] **Step 4: Run the profile package and verify GREEN**

Run: `go test ./internal/nodeagent/adapter/singbox -count=1`

Expected: PASS, including profile alias, wrong adapter, remote endpoint and preview secret-canary cases.

- [ ] **Step 5: RED — add sing-box credential and compiler tests**

The credential schema is closed to `schema_version`, `hysteria2_password`, `certificate_pem`, and `private_key_pem`; server requires all four, client requires schema/password/certificate and rejects a private key.

```go
func TestSingBoxClientRejectsServerPrivateKey(t *testing.T) {
	t.Parallel()
	request := validSingBoxClientRequest()
	request.Credential = strings.NewReader(validSingBoxClientCredentialWithPrivateKey())
	if _, err := CompileClient(request); !errors.Is(err, adapter.ErrCredentialShape) {
		t.Fatal("sing-box client accepted server private key")
	}
}
```

- [ ] **Step 6: Run compiler tests and verify RED**

Run: `go test ./internal/nodesupervisor/adapter -run '^TestSingBox' -count=1`

Expected: FAIL because the sing-box compiler functions are undefined.

- [ ] **Step 7: GREEN — implement deterministic sing-box JSON compilation**

```go
var SingBoxArgv = struct {
	Check  []string
	Start  []string
	Client []string
}{
	Check:  []string{"check", "-c", "/run/talenro/config.json"},
	Start:  []string{"run", "-c", "/run/talenro/config.json"},
	Client: []string{"run", "-c", "/run/talenro/client.json"},
}
```

Emit a Hysteria2 server bound only to `127.0.0.1:32181`, direct outbound only to the echo endpoint, and a client with SOCKS bound only to `127.0.0.1:32182`. Fix TLS server name to `c12-loopback.invalid`; trust only the per-run certificate from the client credential. Disable external DNS, remote rule sets, API, log files and telemetry. Reject unknown JSON, trailing input, empty/oversize password, malformed PEM, certificate/key mismatch, role replay, release ID other than `sing-box-approved-v1`, and any alternate endpoint.

- [ ] **Step 8: Run compiler and registry tests and verify GREEN**

Run: `go test ./internal/nodeagent/adapter/singbox ./internal/nodesupervisor/adapter -count=1`

Expected: PASS with deterministic config digests and no PEM/password bytes in preview, errors or test output.

- [ ] **Step 9: REFACTOR — run race/static checks for the new adapter**

Run: `gofmt -w internal/nodeagent/adapter/singbox internal/nodesupervisor/adapter/singbox.go internal/nodesupervisor/adapter/singbox_test.go`

Run: `go test -race ./internal/nodeagent/adapter/singbox ./internal/nodesupervisor/adapter -count=1`

Expected: PASS with no race report.

- [ ] **Step 10: Commit B09 compiler independently**

```bash
git add internal/nodeagent/adapter/singbox/profile.go internal/nodeagent/adapter/singbox/profile_test.go internal/nodesupervisor/adapter/singbox.go internal/nodesupervisor/adapter/singbox_test.go internal/c12acceptance/adapter_surface_test.go testdata/c12/adapters/sing-box-preview-v1.json
git commit -m "feat: add fixed sing-box loopback adapter"
```

### Task 4: B09 sing-box release review and real loopback smoke exit

**Files:**
- Create: `testdata/c12/releases/sing-box-approved-v1.lock.json`
- Modify: `internal/localrelease/approved-release-manifest.v1.json`
- Create (first line `//go:build c12_core_smoke && linux`): `internal/c12acceptance/coresmoke/singbox_test.go`
- Create: `docs/licenses/c12-sing-box.md`
- Test: `internal/c12acceptance/coresmoke/singbox_test.go`

**Interfaces:**
- Consumes: `singbox.Profile()`, supervisor `Prepare/Check/Start/Probe/Drain/Stop`, the unchanged B06 fixture echo endpoint, the P06 credential session with `AdapterSingBox`, B02 canonical embedded-manifest verifier, and the P07-owned core-smoke runner/core-lock tool.
- Produces: tagged `TestSingBoxLoopbackSmoke`, release ID `sing-box-approved-v1`, GPLv3+ evidence record, the final exact three-entry build-embedded approved manifest, and—only in the post-cleanup parent—`CoreSmokeResultV1{Adapter: sing_box}`.

- [ ] **Step 1: RED — write the sing-box real-smoke test**

```go
func TestSingBoxLoopbackSmoke(t *testing.T) {
	lock := requireReleaseLock(t, "sing-box-approved-v1")
	provisional, err := runRealCoreSmoke(t, SmokeRequest{
		Adapter:           contracts.AdapterSingBox,
		ProfileID:         "sing-box-loopback-hysteria2-v1",
		ReleaseLock:       lock,
		NonceBytes:        32,
		HandshakeDeadline: 10 * time.Second,
	})
	if err != nil {
		t.Fatal("sing-box live smoke failed")
	}
	if err := handoffOpaqueProvisionalCoreSmoke(t, provisional); err != nil {
		t.Fatal("sing-box provisional handoff failed")
	}
}
```

As for Xray, the tagged child has no final result/cleanup API. It hands only the one-use live-source-created opaque provisional handle to the authenticated broker；the parent asserts nonce/process/build receipt and all absence facts only after test exit and complete cleanup.

- [ ] **Step 2: Run the tagged test inside the approved Linux test-host and verify RED**

Run through the P07-owned Linux substrate:

```text
bash scripts/run-c12-core-smoke.sh --phase candidate --adapter sing_box --expect release-lock-unavailable --timeout 2m
```

Expected: FAIL before child creation with `release lock unavailable`.

- [ ] **Step 3: Generate the exact sing-box release lock from the reviewed external OCI layout**

Run:

```text
go run ./cmd/talenro-core-lock lock-core --adapter sing_box --release-id sing-box-approved-v1 --profile sing-box-loopback-hysteria2-v1 --source-layout /run/talenro-c12/review/sing-box-approved-v1.oci --write testdata/c12/releases/sing-box-approved-v1.lock.json
```

Expected: exit 0 with exact version/source/image/executable/dependency/argv/provenance/vulnerability/GPLv3+ obligation/sandbox digests. The same fixed crash-safe pair transaction updates B02's embedded manifest from fixture+Xray to the exact sorted fixture+Xray+sing-box set, advances its version and requires `RequireC12ProductionReleaseSet` plus manifest/lock digest equality before either candidate replaces the old pair. An incomplete review、missing Xray lock、partial update or manifest mismatch exits 1 with the prior lock/manifest pair byte-identical.

- [ ] **Step 4: GREEN — connect the fixture to the sing-box real-core path**

Use the P06 provider to generate one per-run P-256 loopback certificate/key and random Hysteria2 password, split opaque server/client handles by role, then execute the same B06 `run --config-fd=3` echo lifecycle used by the test. The client handle contains no private key, cleanup must yield the exact run-bound receipt, and output contains no PEM/password/challenge/core text. No fixture source or argv changes in this task.

- [ ] **Step 5: Write the sing-box GPLv3+ record**

`docs/licenses/c12-sing-box.md` must contain `Reviewed release`, `OCI/source provenance`, `GPLv3+ license and notices`, `Complete corresponding source`, `Build materials and modifications`, `Vulnerability result`, `Executable/dependency digests`, `Verifier-only distribution boundary`, and `Re-review triggers`. Bind every record to the generated lock digest; do not claim that test-only distribution removes obligations.

- [ ] **Step 6: Run the tagged sing-box smoke and verify GREEN**

Run: `bash scripts/run-c12-core-smoke.sh --phase candidate --adapter sing_box --expect pass --timeout 2m`.

Expected: candidate PASS within 2 minutes with nonce match, native check, healthy probe, external-volume process-image observation, empty server/client/echo cgroups and no credential residue. It cannot close B09 before Step 8 commit and Step 9 rerun.

- [ ] **Step 7: REFACTOR — run sing-box package, privacy and dependency checks**

Run: `go test ./internal/nodeagent/adapter/singbox ./internal/nodesupervisor/adapter ./internal/c12acceptance/coresmoke -count=1`

Run: `go list -deps -json ./cmd/node-agent ./cmd/node-core-supervisor | go run ./cmd/talenro-core-lock verify-go-deps --deny-module github.com/xtls/xray-core --deny-module github.com/sagernet/sing-box`

Expected: PASS and neither core module occurs in a proprietary dependency graph.

- [ ] **Step 8: Commit B09 real-core acceptance independently**

```bash
git add testdata/c12/releases/sing-box-approved-v1.lock.json internal/localrelease/approved-release-manifest.v1.json internal/c12acceptance/coresmoke/singbox_test.go docs/licenses/c12-sing-box.md
git commit -m "test: accept sing-box as an external core"
```

- [ ] **Step 9: Run the owning sing-box gate from the new clean commit**

Immediately require the Task 4 exact paths/index/worktree/untracked state clean, then run `bash scripts/run-c12-core-smoke.sh --phase authoritative --adapter sing_box --expect pass --timeout 2m`.

Expected: PASS and a post-cleanup result bound to the helper-captured new commit/tree digest. Only this post-commit command closes B09 provisionally pending Task 5's aggregate revalidation.

### Task 5: Bind smoke evidence to one adapter and close B08/B09 separately

**Files:**
- Modify: `cmd/c12-core-smoke-runner/main.go`
- Modify: `internal/c12test/coresmokerunner/types.go`
- Modify (first line `//go:build linux`): `internal/c12test/coresmokerunner/channel_linux.go`
- Modify (first line `//go:build !linux`): `internal/c12test/coresmokerunner/runner_stub.go`
- Modify: `internal/c12test/coresmokerunner/runner_test.go`
- Modify (first line `//go:build linux`): `internal/c12test/coresmokerunner/channel_linux_test.go`
- Modify: `internal/c12acceptance/coresmoke/result.go`
- Modify: `internal/c12acceptance/coresmoke/receipts.go`
- Modify: `internal/c12acceptance/coresmoke/result_test.go`
- Modify: `internal/c12acceptance/coresmoke/receipts_test.go`
- Modify: `internal/c12acceptance/coresmoke/xray_test.go`
- Modify: `internal/c12acceptance/coresmoke/singbox_test.go`
- Test: `internal/c12acceptance/coresmoke/result_test.go`

**Interfaces:**
- Consumes: exact release-lock and final embedded-manifest digests, complete helper-authenticated build/process-image/ownership/cleanup receipt preimages emitted by the P07 runner, inherited-FD `LiveReceiptSource`, parent-only `RunnerReceiptFinalizer` backed by its authenticated WAL/readbacks, and on Linux only the parent-created `memfd_create(MFD_CLOEXEC|MFD_ALLOW_SEALING)` one-use protected result slot described below.
- Produces: `CoreSmokeBuildReceiptV1`, `CoreProcessImageReceiptV1`, `CoreCleanupReceiptV1`, `CoreSmokeResultV1`; `ValidateLiveCoreSmoke(context.Context, LiveCoreSmokeObservationV1, contracts.Adapter, contracts.Digest, LiveReceiptSource) (OpaqueProvisionalCoreSmokeHandle,error)` for the test child；the unchanged single-process `FinalizeCoreSmokeResult(context.Context, OpaqueProvisionalCoreSmokeHandle, RunnerReceiptFinalizer) (CoreSmokeResultV1,ValidatedCoreSmokeReceiptSetV1,error)`；`FinalizeCoreSmokeResultToInheritedSlot(context.Context, OpaqueProvisionalCoreSmokeHandle, RunnerReceiptFinalizer, *coresmokerunner.InheritedCoreSmokeResultSlotWriter) error` for an exact native child；`AdoptCoreSmokeResultFromProtectedSlot(context.Context, *coresmokerunner.AdoptedCoreSmokeResultSlot, contracts.Adapter, contracts.Digest) (CoreSmokeResultV1,ValidatedCoreSmokeReceiptSetV1,error)` for its native parent；and the sole semantic validator `ValidateCoreSmokeResult(context.Context, CoreSmokeResultV1, contracts.Adapter, contracts.Digest, TrustedCoreSmokeReceiptSource) (ValidatedCoreSmokeReceiptSetV1,error)`. `TrustedCoreSmokeReceiptSource` remains opaque and non-serializable：it is minted only inside `coresmoke` from the live same-process finalizer or a successfully consumed protected-slot adoption, never from caller bytes、path、FD number or digest. P08 `TestResultDigest` consumes the complete parent-revalidated same-run receipt set including the build receipt, not bare caller digests or an old result path.

- [ ] **Step 1: RED — add cross-adapter substitution tests**

```go
func TestResultCannotSatisfyAnotherAdapter(t *testing.T) {
	t.Parallel()
	result := validCoreSmokeResult(contracts.AdapterXray)
	if _, err := ValidateCoreSmokeResult(context.Background(), result, contracts.AdapterSingBox, singBoxLockDigest(), trustedReceiptSource(t, result.RunID));
		!errors.Is(err, ErrAdapterEvidenceMismatch) {
		t.Fatal("Xray result satisfied sing-box gate")
	}
	result.Adapter = contracts.AdapterSingBox
	if _, err := ValidateCoreSmokeResult(context.Background(), result, contracts.AdapterSingBox, singBoxLockDigest(), trustedReceiptSource(t, result.RunID));
		!errors.Is(err, ErrReleaseEvidenceMismatch) {
		t.Fatal("rewritten adapter bypassed release binding")
	}
}
```

Also write the native handoff RED matrix before implementation. `internal/c12test/coresmokerunner/channel_linux_test.go` owns `TestCoreSmokeResultSlotExactChildRoundTrip`、`TestCoreSmokeResultSlotRejectsCopyReopenAndCrossSession`、`TestCoreSmokeResultSlotRejectsWrongChildAndMutableSlot` and `TestCoreSmokeResultSlotWriterIsOneUse`; `internal/c12acceptance/coresmoke/result_test.go` owns `TestProtectedSlotAdoptionMintsTrustedSourceOnce`、`TestProtectedSlotAdoptionRejectsFinalizerAttestationSplice`、`TestProtectedSlotParentRevalidatesCompleteReceiptSet`、`TestRunnerReceiptFinalizerRejectsCrossSinkReuse` and `TestProtectedSlotWriterHasSoleSemanticCaller`. The positive fixture must fork/exec the test binary as an exact child with the slot only in the explicit inherited-FD table；negative rows must cover copied canonical bytes、`/proc/self/fd` reopen/new open-file-description、same slot under another parent session、another memfd with identical bytes、wrong PID/start token/executable、another adapter/run/release/WAL prefix or final WAL、missing `F_SEAL_WRITE|F_SEAL_GROW|F_SEAL_SHRINK|F_SEAL_SEAL`、child/grandchild FD inheritance、write after seal、single-process-then-slot and slot-then-single-process reuse、second writer use、second adoption and second trusted-source use. The static call-site test exact-covers one non-test `AttestSealAndConsume` call inside `FinalizeCoreSmokeResultToInheritedSlot` and one non-test `ConsumeCanonicalFinalizerEnvelope` call inside `AdoptCoreSmokeResultFromProtectedSlot`；owning unit tests may invoke them only through explicit negative fixtures.

- [ ] **Step 2: Run the result test and verify RED**

Run: `go test ./internal/c12acceptance/coresmoke -run '^TestResultCannotSatisfyAnotherAdapter$' -count=1`

Run on Linux: `go test ./internal/c12test/coresmokerunner ./internal/c12acceptance/coresmoke -run '^(TestCoreSmokeResultSlot|TestProtectedSlot|TestRunnerReceiptFinalizer)' -count=1`

Expected: FAIL because the Task 2 validator exists for one adapter/run, but the two-adapter aggregate/non-substitution matrix、complete cross-run/cross-volume mutation table and protected result-slot APIs are not yet implemented. A non-Linux build must compile only the existing fail-closed stub and must never emulate an authoritative memfd handoff.

- [ ] **Step 3: GREEN — implement the finite adapter-bound result**

```go
type CoreSmokeResultV1 struct {
	SchemaVersion          string
	RunID                  uuid.UUID
	Adapter                contracts.Adapter
	ProfileID              string
	ReleaseLockDigest      contracts.Digest
	ApprovedManifestDigest contracts.Digest
	BuildReceipt           CoreSmokeBuildReceiptV1
	BuildReceiptDigest     contracts.Digest
	ServerImageDigest      contracts.Digest
	ClientImageDigest      contracts.Digest
	ProcessImageReceipt    CoreProcessImageReceiptV1
	ProcessImageReceiptDigest contracts.Digest
	NonceMatched           bool
	NativeCheckPassed      bool
	HealthyProbePassed     bool
	ServerCgroupEmpty      bool
	ClientCgroupEmpty      bool
	EchoCgroupEmpty        bool
	CleanupReceipt         CoreCleanupReceiptV1
	CleanupReceiptDigest   contracts.Digest
	StartedAt              time.Time
	FinishedAt             time.Time
}
```

Freeze the embedded receipt preimages:

```go
type CoreSmokeBuildReceiptV1 struct {
	SchemaVersion            string
	RunID                    uuid.UUID
	Phase                    string
	CommitOID                string
	TrackedTreeDigest        contracts.Digest
	RunnerRecipeDigest       contracts.Digest
	ToolchainDigest          contracts.Digest
	SourceClosureDigest      contracts.Digest
	ApprovedManifestDigest   contracts.Digest
	BrokerExecutableDigest   contracts.Digest
	TestExecutableDigest     contracts.Digest
	Test2JSONDigest          contracts.Digest
	TestHostImageDigest      contracts.Digest
	TestHostConfigDigest     contracts.Digest
	TestHostRootFSDigest     contracts.Digest
	OwnershipWALPrefixDigest contracts.Digest
}

type CoreProcessImageReceiptV1 struct {
	SchemaVersion              string
	RunID                      uuid.UUID
	Adapter                    contracts.Adapter
	ReleaseID                  string
	ProfileID                  string
	ExternalCoreVolumeActualID string
	OwnershipWALPrefixDigest   contracts.Digest
	BuildReceiptDigest         contracts.Digest
	ServerImageDigest          contracts.Digest
	ClientImageDigest          contracts.Digest
	EchoImageDigest            contracts.Digest
	ServerProcessIdentity      contracts.Digest
	ClientProcessIdentity      contracts.Digest
	EchoProcessIdentity        contracts.Digest
	StartedAt                  time.Time
	FinishedAt                 time.Time
}

type CoreOwnedResourceAbsenceV1 struct {
	Kind                  CoreOwnedResourceKindV1
	ActualIdentityDigest  contracts.Digest
	AbsenceReceiptDigest  contracts.Digest
}

type CoreCleanupReceiptV1 struct {
	SchemaVersion              string
	RunID                      uuid.UUID
	Adapter                    contracts.Adapter
	ReleaseID                  string
	ProfileID                  string
	ExternalCoreVolumeActualID string
	OwnershipWALPrefixDigest   contracts.Digest
	FinalOwnershipWALDigest    contracts.Digest
	SortedResourceAbsences     []CoreOwnedResourceAbsenceV1
	FinalResourceSetDigest     contracts.Digest
	FinalAbsenceDigest         contracts.Digest
	EmptyCgroupIDs             []contracts.Digest
	CredentialHandleSetDigest  contracts.Digest
	CredentialAbsenceDigest    contracts.Digest
	CompletedAt                time.Time
}
```

Freeze the Linux native handoff without adding another package or build-tag file. `internal/c12test/coresmokerunner/types.go` owns the platform-neutral declarations below；the existing `channel_linux.go` owns `memfd`、seal、`pidfd`/`/proc` no-follow identity and inherited-FD mechanics；the existing `runner_stub.go` returns only `ErrProtectedCoreSmokeResultSlotUnsupported` on non-Linux. The entire `internal/c12test/coresmokerunner` package is a payload transport/substrate capability layer and must not import `internal/c12acceptance/coresmoke`；`cmd/c12-core-smoke-runner/main.go` is the sole composition root allowed to import both. `internal/c12acceptance/coresmoke/result.go` remains the sole canonical payload、finalizer-attestation interpretation、trusted-source mint and semantic-validator owner, so no import cycle or second receipt parser is permitted.

`result.go` also owns `RunnerReceiptFinalizer` as a sealed interface with the sole unexported method `consumeFinalState(context.Context, OpaqueProvisionalCoreSmokeHandle) (finalizedCoreSmokeState,error)`. The private state contains the result、complete receipt set、run/WAL final binding and one-use finalizer authorization but no exported source constructor. `FinalizeCoreSmokeResult` and `FinalizeCoreSmokeResultToInheritedSlot` are the only two terminal consumers；an atomic state transition makes them mutually exclusive, and a second/cross-sink call returns `ErrCoreSmokeFinalizerConsumed`.

```go
const CoreSmokeResultSlotFDV1 = 3
const CoreSmokeResultSlotCapacityV1 int64 = 64 << 10

type CoreSmokeResultSlotExpectationV1 struct {
	SchemaVersion          string
	Adapter                contracts.Adapter
	ReleaseLockDigest      contracts.Digest
	CommitOID              string
	RunnerExecutableDigest contracts.Digest
}

type CoreSmokeResultSlotFinalBindingV1 struct {
	RunID                    uuid.UUID
	OwnershipWALPrefixDigest contracts.Digest
	FinalOwnershipWALDigest  contracts.Digest
}

type CoreSmokeResultSlotBoundRunV1 struct {
	ParentSessionBindingDigest contracts.Digest
	RunID                      uuid.UUID
	Adapter                    contracts.Adapter
	ReleaseLockDigest          contracts.Digest
	CommitOID                  string
	RunnerExecutableDigest     contracts.Digest
}

type CoreSmokeResultSlotAttestationV1 struct {
	SchemaVersion              string
	ParentSessionBindingDigest contracts.Digest
	SlotDevice                 uint64
	SlotInode                  uint64
	ChildPID                   uint32
	ChildStartTokenDigest      contracts.Digest
	ChildExecutableDigest      contracts.Digest
	RunID                      uuid.UUID
	OwnershipWALPrefixDigest   contracts.Digest
	FinalOwnershipWALDigest    contracts.Digest
	CanonicalPayloadDigest     contracts.Digest
	MAC                        contracts.Digest
}

type CoreSmokeResultParentSession struct{ /* opaque, non-copyable */ }
type InheritedCoreSmokeResultSlotWriter struct{ /* opaque, non-copyable */ }
type AdoptedCoreSmokeResultSlot struct{ /* opaque, non-copyable */ }

func NewCoreSmokeResultParentSession(context.Context, CoreSmokeResultSlotExpectationV1) (*CoreSmokeResultParentSession, *os.File, error)
func OpenInheritedCoreSmokeResultSlotWriter() (*InheritedCoreSmokeResultSlotWriter, error)
func (*CoreSmokeResultParentSession) BoundRun() CoreSmokeResultSlotBoundRunV1
func (*InheritedCoreSmokeResultSlotWriter) BoundRun() CoreSmokeResultSlotBoundRunV1
func (*CoreSmokeResultParentSession) BindStartedChild(*os.Process) error
func (*InheritedCoreSmokeResultSlotWriter) AttestSealAndConsume([]byte, CoreSmokeResultSlotFinalBindingV1) error
func (*CoreSmokeResultParentSession) AdoptAfterChildExit(context.Context, *os.ProcessState) (*AdoptedCoreSmokeResultSlot, error)
func (*AdoptedCoreSmokeResultSlot) ConsumeCanonicalFinalizerEnvelope() (CoreSmokeResultSlotBoundRunV1, []byte, CoreSmokeResultSlotAttestationV1, error)
func (*CoreSmokeResultParentSession) Close() error
func (*InheritedCoreSmokeResultSlotWriter) Close() error
func (*AdoptedCoreSmokeResultSlot) Close() error
```

`NewCoreSmokeResultParentSession` validates the literal schema `talenro-core-smoke-result-slot-expectation/v1`, exact adapter/release/clean-commit/executable expectations, generates a random RunID、parent-session ID and 256-bit one-use HMAC key internally, records the exact parent PID/start token/executable, and creates one anonymous `memfd_create(MFD_CLOEXEC|MFD_ALLOW_SEALING)` slot. It writes a fixed private bootstrap header in state `created`, preallocates exactly 64 KiB, applies `F_SEAL_GROW|F_SEAL_SHRINK`, takes an exclusive `F_OFD_SETLK` lock on the reserved identity byte range, retains that original open-file-description as the read/adoption capability plus the key in locked parent memory, and returns exactly one child-side duplicate sharing the same open-file-description. The native parent makes itself non-dumpable before creating the slot, places that handle at literal FD 3 only through the direct child's explicit inherited-FD table, closes its child-side duplicate immediately after successful start, then calls `BindStartedChild`；the binding opens/stores a `pidfd`, reads the no-follow start token and `/proc/<pid>/exe` identity itself, requires the expected prebuilt runner digest, and fsyncs the exact child identity plus authenticated transition `created → child_bound`. `OpenInheritedCoreSmokeResultSlotWriter` waits at most five seconds for `child_bound`, requires that identity equal its own PID/start/executable, and proves FD 3 shares the parent's open-file-description by successfully reasserting the reserved OFD lock；a `/proc/self/fd` reopen/new open-file-description receives the conflicting-lock failure while an inherited duplicate succeeds. It then sets child non-dumpable, moves the bootstrap key into locked memory and zeroizes the bootstrap copy before returning. Any timeout、unsupported/denied `pidfd`/OFD-lock operation、state skip/replay or identity change fails before Docker/build/test work. The handle/session/key is never represented in argv、environment、filesystem path、stdout/stderr or `SCM_RIGHTS`, and FD 3 is marked close-on-exec before the runner can create any broker/container/grandchild. The slot's authenticated on-memfd states are exactly `created → child_bound → final_sealed`; the first two transitions are fsynced and `final_sealed` is fsynced before the immutable seals. The later `adopted` bit is an atomic one-use state held only by the opaque parent session, because `F_SEAL_WRITE` deliberately makes any post-final slot mutation impossible. The slot attestation schema is exactly `talenro-core-smoke-result-slot-attestation/v1`; its bounded record has no unknown、duplicate、null or omitted member.

`OpenInheritedCoreSmokeResultSlotWriter` takes no caller argument and accepts only literal FD 3 with the exact private header、parent PID/start/executable、session/run/expectation、device/inode/size and initial seals. Before exposing the key it enumerates its own descriptor table only to reject any second descriptor whose device/inode identifies this slot (including `dup` and `/proc/self/fd` reopen), then marks FD 3 close-on-exec；unrelated Go runtime descriptors do not satisfy that identity and are not treated as capabilities. It consumes the bootstrap key into locked memory and yields one writer；the two `BoundRun` accessors return the same immutable typed binding so the composition root must use the parent-generated RunID and cannot substitute caller state. After the exact 11-resource cleanup and authenticated WAL terminal state, `FinalizeCoreSmokeResultToInheritedSlot`, under the authorization returned by the same `RunnerReceiptFinalizer` terminal transition that backs `FinalizeCoreSmokeResult`, constructs and validates the bounded payload `JCS({"core_smoke_result":CoreSmokeResultV1,"validated_receipt_set":ValidatedCoreSmokeReceiptSetV1})` and makes the sole repository call to `AttestSealAndConsume` with the finalizer-owned run/WAL binding. `ValidatedCoreSmokeReceiptSetV1` in this payload is finite canonical data only and contains no source、FD or capability. The method computes `CoreSmokeResultSlotAttestationV1.MAC = HMAC-SHA-256(one_use_parent_key, ASCII("TALENRO-C12-CORE-SMOKE-FINALIZER-HANDOFF-V1") || 0x00 || JCS(attestation_without_mac) || 0x00 || canonical_payload)`, writes a strict length-prefixed record containing only the canonical payload plus attestation, zero-fills the unused fixed capacity, calls `fsync`, applies `F_SEAL_WRITE|F_SEAL_SEAL`, verifies the exact final seal set `F_SEAL_GROW|F_SEAL_SHRINK|F_SEAL_WRITE|F_SEAL_SEAL`, zeroizes the child key and consumes/closes every writer capability. The terminal API returns only `error` to the child：in protected-slot mode the runner cannot obtain a `TrustedCoreSmokeReceiptSource`, materialize an evidence root, print a result path or publish the result by any other sink. This host-runner FD 3 is in a different process from the already frozen container test-child FD 9 and the fixture process's own `--config-fd=3`; no descriptor crosses or aliases those boundaries.

Only after the bound child has exited and been reaped may `AdoptAfterChildExit` verify the stored `pidfd`/PID/start token/executable、successful exact process state、parent session、expected adapter/release/commit/run、slot device/inode/size、exact final seals and HMAC, zeroize the parent HMAC key, and return one opaque read-only adoption. Its one-use `ConsumeCanonicalFinalizerEnvelope` returns the immutable bound-run projection、canonical bytes and attestation together；it is not a constructor and no bytes/FD can recreate an adoption. `AdoptCoreSmokeResultFromProtectedSlot` is the sole non-test consumer of that method, strictly decodes the bounded canonical payload and attestation, requires every run/build/WAL/adapter/release member byte-identical to the adopted bound run and receipt set, mints `TrustedCoreSmokeReceiptSource` only in the adopting parent process, and calls `ValidateCoreSmokeResult` again as the sole semantic validator before returning the result/set. Its adapter and release-digest parameters are verifier expectations only and cannot mint a source without the opaque adoption. All three native handles have idempotent cleanup-only `Close` methods；every parent/child path defers them, and failure/close zeroizes keys、closes pidfd/memfd and destroys adoption state without producing evidence. Copied canonical bytes、an FD number、a reopened/duplicated FD、an identical second memfd、a different slot/session/child/adapter/run/WAL、missing seal、noncanonical/trailing data、child crash、pre-cleanup write、second write/adoption/source use all fail closed；there is no exported trusted-source constructor from bytes、path、digest、attestation or native handle scalar.

P08's long-lived native verifier parent is the only cross-process consumer contract: before starting each exact catalog/hash/OS/arch-locked prebuilt `c12-core-smoke-runner`, it creates and defers `Close` on one fresh `CoreSmokeResultParentSession`, supplies the returned child handle only as inherited FD 3 to that exact direct child, binds the started process, waits/reaps it, adopts once, and uses only the returned parent-revalidated complete `ValidatedCoreSmokeReceiptSetV1`. It never receives a Go object from the child and cannot substitute a result file、serialized `TrustedCoreSmokeReceiptSource`、old set、digest-only value or shell-mediated channel. Each adapter/run has a distinct session and slot.

`CoreOwnedResourceAbsenceV1={kind,actual_identity_digest,absence_receipt_digest}` uses the exact declaration-order kinds `source_container`、`init_container`、`test_host_container`、`private_network`、`external_core_volume`、`server_cgroup`、`client_cgroup`、`echo_cgroup`、`server_credential_handle`、`client_credential_handle`、`run_root`, each exactly once. Validation requires schemas `talenro-core-smoke-build-receipt/v1`、`talenro-core-process-image-receipt/v1`、`talenro-core-cleanup-receipt/v1` and `talenro-core-smoke-result/v1`; nonzero and byte-identical run/adapter/release/profile/actual-volume/WAL-prefix bindings；exact final embedded-manifest/lock membership；a helper-authenticated build receipt whose recipe/toolchain/source closure matches the tracked recipe and whose actual broker/test/test2json/image tuple came from byte-equal double builds of the same committed tree；and equality of its digest in the process/result records. It also requires proof that `FinalOwnershipWALDigest` is the authenticated append-only successor of `OwnershipWALPrefixDigest`; the exact 11-resource set and recomputed absence digest; exactly three sorted unique nonzero cgroup IDs matching their absence entries; UTC whole-millisecond `StartedAt < FinishedAt <= CompletedAt`; exact server/client image digests from the reviewed lock; exact process identities observed in that external-core volume; all booleans true; and recomputed canonical receipt digests equal the redundant fields. A tracked runner manifest that contains or overrides a final artifact hash is invalid；only the same-run build receipt may carry those hashes.

The Go test's `LiveReceiptSource` comes only from inherited FD 9 and can validate live process/volume plus the P06 credential cleanup, then emit a one-use opaque provisional handle to the parent helper；it cannot create `CoreCleanupReceiptV1` or a final result. After test exit and full cleanup, the same parent-held `RunnerReceiptFinalizer` appends/reads the final WAL, verifies every exact absence including run-root removal, and constructs the cleanup receipt/full `ValidatedCoreSmokeReceiptSetV1` in locked memory. In ordinary P07 single-process mode, the unchanged `FinalizeCoreSmokeResult` returns those values to the same runner parent and materialization is allowed only after its final validation. In inherited protected-slot mode, the mutually exclusive terminal call is `FinalizeCoreSmokeResultToInheritedSlot`; the child receives no result/source and only the adopting native parent may publish after its own second validation. Neither interface has a digest/caller-bytes/path constructor. Random nonzero digests、cross-run/volume/WAL-prefix receipts、unlinked final WAL、missing/duplicate resource kind、old cleanup、fabricated cgroup emptiness、time reorder、FD/source reuse and publish-before-cleanup all fail.

- [ ] **Step 4: Run the result package and verify GREEN**

Run: `go test ./internal/c12acceptance/coresmoke -run '^(TestResultCannotSatisfyAnotherAdapter|TestBuildReceiptRequiresSameCommittedTreeDoubleBuild|TestReceiptSourceRejectsFabrication|TestReceiptBindingsAndTimes)$' -count=1`

Run on Linux: `go test ./internal/c12test/coresmokerunner ./internal/c12acceptance/coresmoke -run '^(TestCoreSmokeResultSlot|TestProtectedSlot|TestRunnerReceiptFinalizer)' -count=1`

Expected: PASS for exact Xray and sing-box records plus one exact fork/exec protected-slot round trip；all cross-adapter、cross-release、recipe/tree/manifest/build-artifact splice、single-build-only、missing-cleanup、copied/reopened/cross-session slot、wrong child/run/WAL、mutable/unsealed and replay mutations fail.

- [ ] **Step 5: REFACTOR — execute both independent real-core commands**

Run on the approved Linux test-host through the owning substrate, as two commands rather than one filtered aggregate:

```text
bash scripts/run-c12-core-smoke.sh --phase candidate --adapter xray --expect pass --timeout 2m
bash scripts/run-c12-core-smoke.sh --phase candidate --adapter sing_box --expect pass --timeout 2m
```

Expected: both candidate runs PASS independently；deleting either provisional result causes its own candidate validation to fail, neither validates under the other adapter, and neither closes a gate before Task 5 commit/Step 8 authoritative reruns.

- [ ] **Step 6: Run repository-wide non-Docker regressions**

Run: `go test ./... -count=1 -timeout 10m`

Run: `go test -race ./internal/nodeagent/adapter/... ./internal/nodesupervisor/adapter ./internal/c12acceptance/coresmoke -count=1`

Run on Linux: `go test -race ./internal/c12test/coresmokerunner ./internal/c12acceptance/coresmoke -run '^(TestCoreSmokeResultSlot|TestProtectedSlot|TestRunnerReceiptFinalizer)' -count=1`

Run: `go vet ./...`

Run: `go tool golangci-lint run ./...`

Expected: all commands PASS.

- [ ] **Step 7: Commit the non-substitution guard**

```bash
git add cmd/c12-core-smoke-runner/main.go internal/c12test/coresmokerunner/types.go internal/c12test/coresmokerunner/channel_linux.go internal/c12test/coresmokerunner/runner_stub.go internal/c12test/coresmokerunner/runner_test.go internal/c12test/coresmokerunner/channel_linux_test.go internal/c12acceptance/coresmoke/result.go internal/c12acceptance/coresmoke/receipts.go internal/c12acceptance/coresmoke/result_test.go internal/c12acceptance/coresmoke/receipts_test.go internal/c12acceptance/coresmoke/xray_test.go internal/c12acceptance/coresmoke/singbox_test.go
git commit -m "test: keep external core evidence adapter bound"
```

- [ ] **Step 8: Re-run both owning gates from the Task 5 clean commit**

Require every Task 5 exact path/index/worktree/untracked state clean, revalidate the unchanged exact final three-entry embedded manifest against both reviewed locks, and run the two adapters separately with `--phase authoritative`. Each result must bind the same helper-captured Task 5 commit/tree digest and a fresh helper-authenticated build receipt produced from that commit, and validate only under its own adapter/release.

## B08 and B09 exit gates

`C1.2-B08` closes only when Tasks 1–2 pass, the reviewed Xray lock/license record exists, `TestXrayLoopbackSmoke` proves a 32-byte nonce round trip within 5 seconds, and cleanup/process-image evidence validates.

`C1.2-B09` closes only when Tasks 3–4 pass, the reviewed sing-box lock/license record exists, `TestSingBoxLoopbackSmoke` proves a 32-byte nonce round trip within 10 seconds, and cleanup/process-image evidence validates.

Task 5 is required before B10: it proves the two completed gates remain distinct inputs to the canonical verifier. Neither Docker orchestration nor a config-check-only result may retroactively replace either real protocol smoke.

Owning positive exits are exactly:

```bash
bash scripts/run-c12-core-smoke.sh --phase authoritative --adapter xray --expect pass --timeout 2m
bash scripts/run-c12-core-smoke.sh --phase authoritative --adapter sing_box --expect pass --timeout 2m
```

Each command requires its one exact `c12_core_smoke && linux` top-level test to terminate `pass`, validates the complete trusted receipt set and publishes only after exact cleanup. A Windows/local untagged package pass、empty tagged run、`t.Skip`、ambient `/run` directory or later P08 scope cannot close either gate retroactively.
