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

This plan consumes these exact interfaces from batches 4–6:

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
- Create: `testdata/c12/adapters/xray-preview-v1.json`
- Modify: `internal/nodeagent/adapter/registry.go`
- Modify: `internal/nodesupervisor/adapter/compiler.go`
- Test: `internal/nodeagent/adapter/xray/profile_test.go`
- Test: `internal/nodesupervisor/adapter/xray_test.go`

**Interfaces:**
- Consumes: `adapter.Register(ProfileV1) error`, `adapter.Preview(contracts.ProcessSpecV1) (PreviewV1, error)`, `nodesupervisoradapter.RegisterCompiler(string, CompileFunc) error`, and the exact `ProfileV1`/`CompileRequestV1` types frozen above.
- Produces: `xray.Profile() adapter.ProfileV1`, `xray.CompileServer(adapter.CompileRequestV1) (adapter.CompiledConfigV1, error)`, and `xray.CompileClient(adapter.CompileRequestV1) (adapter.CompiledConfigV1, error)`.

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
git add internal/nodeagent/adapter/registry.go internal/nodeagent/adapter/xray/profile.go internal/nodeagent/adapter/xray/profile_test.go internal/nodesupervisor/adapter/compiler.go internal/nodesupervisor/adapter/xray.go internal/nodesupervisor/adapter/xray_test.go testdata/c12/adapters/xray-preview-v1.json
git commit -m "feat: add fixed Xray loopback adapter"
```

### Task 2: B08 Xray release review and real loopback smoke exit

**Files:**
- Create: `testdata/c12/releases/xray-approved-v1.lock.json`
- Create: `internal/c12acceptance/coresmoke/xray_test.go`
- Create: `docs/licenses/c12-xray-core.md`
- Modify: `cmd/c12-fixture/main.go`
- Test: `internal/c12acceptance/coresmoke/xray_test.go`

**Interfaces:**
- Consumes: `xray.Profile()`, supervisor `Prepare/Check/Start/Probe/Drain/Stop`, `CoreSmokeCredentialProvider.New(context.Context, contracts.Adapter, uuid.UUID, string) (ServerCredentialHandle, ClientCredentialHandle, error)`, and the B06 fixture-owned echo endpoint.
- Produces: tagged `TestXrayLoopbackSmoke`, release ID `xray-approved-v1`, license record `docs/licenses/c12-xray-core.md`, and an adapter-bound `CoreSmokeResultV1` with `Adapter=xray`.

- [ ] **Step 1: RED — write the Xray real-smoke test**

```go
func TestXrayLoopbackSmoke(t *testing.T) {
	lock := requireReleaseLock(t, "xray-approved-v1")
	result := runRealCoreSmoke(t, SmokeRequest{
		Adapter:           contracts.AdapterXray,
		ProfileID:         "xray-loopback-vless-tcp-v1",
		ReleaseLock:       lock,
		NonceBytes:        32,
		HandshakeDeadline: 5 * time.Second,
	})
	if result.Adapter != contracts.AdapterXray || !result.NonceMatched ||
		!result.ServerCgroupEmpty || !result.ClientCgroupEmpty || !result.EchoCgroupEmpty {
		t.Fatal("xray real-core smoke did not prove handshake and cleanup")
	}
}
```

- [ ] **Step 2: Run the tagged test inside the approved Linux test-host and verify RED**

Run inside `node-agent-test-host`:

```text
go test -tags=c12_core_smoke ./internal/c12acceptance/coresmoke -run '^TestXrayLoopbackSmoke$' -count=1 -timeout 2m -args -release-map /run/talenro-c12/installed-release-map.v1.json -result /run/talenro-c12/output/xray-result.json
```

Expected: FAIL before child creation with `release lock unavailable`.

- [ ] **Step 3: Generate the exact Xray release lock from the reviewed external OCI layout**

Run in the isolated review runner:

```text
go run ./cmd/talenro-artifact-scan lock-core --adapter xray --release-id xray-approved-v1 --profile xray-loopback-vless-tcp-v1 --source-layout /run/talenro-c12/review/xray-approved-v1.oci --write testdata/c12/releases/xray-approved-v1.lock.json
```

Expected: exit 0 and one canonical JSON record containing exact upstream version/source URI/image manifest digest, server/client executable and dependency SHA-256, the three fixed argv arrays, provenance digest, vulnerability result digest, MPL-2.0 license record digest and sandbox profile. If any field cannot be derived or verified, expected exit is 1 and the lock file is not created.

- [ ] **Step 4: GREEN — connect the fixture to the Xray real-core path**

Add a closed `core-smoke` fixture mode that accepts only the lock path, installed-map path and output path supplied by the test harness. It must generate a random 32-byte nonce and single-use VLESS credential, call `Prepare → Check → Start → Probe`, compare the returned nonce, snapshot finite health/capacity, call `Drain → Stop`, clear credential buffers, and assert all three owned cgroups and credential files are absent. It must never print core stdout/stderr, config, nonce or credential.

- [ ] **Step 5: Write the Xray license record with fixed evidence headings**

`docs/licenses/c12-xray-core.md` must contain exact sections `Reviewed release`, `OCI/source provenance`, `MPL-2.0 covered files`, `Source availability`, `Notices`, `Vulnerability result`, `Executable/dependency digests`, `Distribution boundary`, and `Re-review triggers`. Each section records the digest copied from the generated lock; the document must state that the OCI layout is verifier-only test inventory and is not a proprietary release input.

- [ ] **Step 6: Run the tagged Xray smoke and verify GREEN**

Run the same tagged command from Step 2.

Expected: PASS within 2 minutes; `xray-result.json` contains only fixed enums/digests/timestamps, `nonce_matched=true`, native check success, healthy probe, process-image path under the run-owned external-core volume, and three empty-cgroup receipts.

- [ ] **Step 7: REFACTOR — run Xray package, privacy and dependency checks**

Run: `go test ./internal/nodeagent/adapter/xray ./internal/nodesupervisor/adapter ./internal/c12acceptance/coresmoke -count=1`

Run: `go list -deps -json ./cmd/node-agent ./cmd/node-core-supervisor | go run ./cmd/talenro-artifact-scan verify-go-deps --deny-module github.com/xtls/xray-core --deny-module github.com/sagernet/sing-box`

Expected: both commands PASS and no core module is in either proprietary dependency graph.

- [ ] **Step 8: Commit B08 real-core acceptance independently**

```bash
git add testdata/c12/releases/xray-approved-v1.lock.json internal/c12acceptance/coresmoke/xray_test.go cmd/c12-fixture/main.go docs/licenses/c12-xray-core.md
git commit -m "test: accept Xray as an external core"
```

Do not begin Task 3 until the Step 6 real Xray handshake has passed; a unit-only or config-check-only result does not close B08.

### Task 3: B09 sing-box typed profile and supervisor compiler

**Files:**
- Create: `internal/nodeagent/adapter/singbox/profile.go`
- Create: `internal/nodeagent/adapter/singbox/profile_test.go`
- Create: `internal/nodesupervisor/adapter/singbox.go`
- Create: `internal/nodesupervisor/adapter/singbox_test.go`
- Create: `testdata/c12/adapters/sing-box-preview-v1.json`
- Modify: `internal/nodeagent/adapter/registry.go`
- Modify: `internal/nodesupervisor/adapter/compiler.go`
- Test: `internal/nodeagent/adapter/singbox/profile_test.go`
- Test: `internal/nodesupervisor/adapter/singbox_test.go`

**Interfaces:**
- Consumes: the same closed `ProfileV1`, `CompileRequestV1` and registry signatures frozen at the top of this plan; it does not consume any Xray config type or release lock.
- Produces: `singbox.Profile() adapter.ProfileV1`, `singbox.CompileServer(adapter.CompileRequestV1) (adapter.CompiledConfigV1, error)`, and `singbox.CompileClient(adapter.CompileRequestV1) (adapter.CompiledConfigV1, error)`.

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
git add internal/nodeagent/adapter/registry.go internal/nodeagent/adapter/singbox/profile.go internal/nodeagent/adapter/singbox/profile_test.go internal/nodesupervisor/adapter/compiler.go internal/nodesupervisor/adapter/singbox.go internal/nodesupervisor/adapter/singbox_test.go testdata/c12/adapters/sing-box-preview-v1.json
git commit -m "feat: add fixed sing-box loopback adapter"
```

### Task 4: B09 sing-box release review and real loopback smoke exit

**Files:**
- Create: `testdata/c12/releases/sing-box-approved-v1.lock.json`
- Create: `internal/c12acceptance/coresmoke/singbox_test.go`
- Create: `docs/licenses/c12-sing-box.md`
- Modify: `cmd/c12-fixture/main.go`
- Test: `internal/c12acceptance/coresmoke/singbox_test.go`

**Interfaces:**
- Consumes: `singbox.Profile()`, supervisor `Prepare/Check/Start/Probe/Drain/Stop`, the B06 fixture echo endpoint, and `CoreSmokeCredentialProvider.New` with `AdapterSingBox`.
- Produces: tagged `TestSingBoxLoopbackSmoke`, release ID `sing-box-approved-v1`, GPLv3+ evidence record, and `CoreSmokeResultV1{Adapter: sing_box}`.

- [ ] **Step 1: RED — write the sing-box real-smoke test**

```go
func TestSingBoxLoopbackSmoke(t *testing.T) {
	lock := requireReleaseLock(t, "sing-box-approved-v1")
	result := runRealCoreSmoke(t, SmokeRequest{
		Adapter:           contracts.AdapterSingBox,
		ProfileID:         "sing-box-loopback-hysteria2-v1",
		ReleaseLock:       lock,
		NonceBytes:        32,
		HandshakeDeadline: 10 * time.Second,
	})
	if result.Adapter != contracts.AdapterSingBox || !result.NonceMatched ||
		!result.ServerCgroupEmpty || !result.ClientCgroupEmpty || !result.EchoCgroupEmpty {
		t.Fatal("sing-box real-core smoke did not prove handshake and cleanup")
	}
}
```

- [ ] **Step 2: Run the tagged test inside the approved Linux test-host and verify RED**

Run inside `node-agent-test-host`:

```text
go test -tags=c12_core_smoke ./internal/c12acceptance/coresmoke -run '^TestSingBoxLoopbackSmoke$' -count=1 -timeout 2m -args -release-map /run/talenro-c12/installed-release-map.v1.json -result /run/talenro-c12/output/sing-box-result.json
```

Expected: FAIL before child creation with `release lock unavailable`.

- [ ] **Step 3: Generate the exact sing-box release lock from the reviewed external OCI layout**

Run:

```text
go run ./cmd/talenro-artifact-scan lock-core --adapter sing_box --release-id sing-box-approved-v1 --profile sing-box-loopback-hysteria2-v1 --source-layout /run/talenro-c12/review/sing-box-approved-v1.oci --write testdata/c12/releases/sing-box-approved-v1.lock.json
```

Expected: exit 0 with exact version/source/image/executable/dependency/argv/provenance/vulnerability/GPLv3+ obligation/sandbox digests. An incomplete review exits 1 without a lock file.

- [ ] **Step 4: GREEN — connect the fixture to the sing-box real-core path**

Generate one per-run P-256 loopback certificate/key and random Hysteria2 password, split server/client handles by role, then execute the same bounded lifecycle used by the test. The client path must consume no private key, all credential buffers must be zeroed, and output must contain no PEM/password/nonce/core text.

- [ ] **Step 5: Write the sing-box GPLv3+ record**

`docs/licenses/c12-sing-box.md` must contain `Reviewed release`, `OCI/source provenance`, `GPLv3+ license and notices`, `Complete corresponding source`, `Build materials and modifications`, `Vulnerability result`, `Executable/dependency digests`, `Verifier-only distribution boundary`, and `Re-review triggers`. Bind every record to the generated lock digest; do not claim that test-only distribution removes obligations.

- [ ] **Step 6: Run the tagged sing-box smoke and verify GREEN**

Run the same tagged command from Step 2.

Expected: PASS within 2 minutes with nonce match, native check, healthy probe, external-volume process-image receipt, empty server/client/echo cgroups and no credential residue.

- [ ] **Step 7: REFACTOR — run sing-box package, privacy and dependency checks**

Run: `go test ./internal/nodeagent/adapter/singbox ./internal/nodesupervisor/adapter ./internal/c12acceptance/coresmoke -count=1`

Run: `go list -deps -json ./cmd/node-agent ./cmd/node-core-supervisor | go run ./cmd/talenro-artifact-scan verify-go-deps --deny-module github.com/xtls/xray-core --deny-module github.com/sagernet/sing-box`

Expected: PASS and neither core module occurs in a proprietary dependency graph.

- [ ] **Step 8: Commit B09 real-core acceptance independently**

```bash
git add testdata/c12/releases/sing-box-approved-v1.lock.json internal/c12acceptance/coresmoke/singbox_test.go cmd/c12-fixture/main.go docs/licenses/c12-sing-box.md
git commit -m "test: accept sing-box as an external core"
```

### Task 5: Bind smoke evidence to one adapter and close B08/B09 separately

**Files:**
- Create: `internal/c12acceptance/coresmoke/result.go`
- Create: `internal/c12acceptance/coresmoke/result_test.go`
- Modify: `internal/c12acceptance/coresmoke/xray_test.go`
- Modify: `internal/c12acceptance/coresmoke/singbox_test.go`
- Test: `internal/c12acceptance/coresmoke/result_test.go`

**Interfaces:**
- Consumes: exact release-lock digest, process-image digest, profile ID and cleanup receipts emitted by Tasks 2 and 4.
- Produces: `CoreSmokeResultV1`, `ValidateCoreSmokeResult(CoreSmokeResultV1, contracts.Adapter, contracts.Digest) error`, and two non-interchangeable acceptance result files.

- [ ] **Step 1: RED — add cross-adapter substitution tests**

```go
func TestResultCannotSatisfyAnotherAdapter(t *testing.T) {
	t.Parallel()
	result := validCoreSmokeResult(contracts.AdapterXray)
	if err := ValidateCoreSmokeResult(result, contracts.AdapterSingBox, singBoxLockDigest());
		!errors.Is(err, ErrAdapterEvidenceMismatch) {
		t.Fatal("Xray result satisfied sing-box gate")
	}
	result.Adapter = contracts.AdapterSingBox
	if err := ValidateCoreSmokeResult(result, contracts.AdapterSingBox, singBoxLockDigest());
		!errors.Is(err, ErrReleaseEvidenceMismatch) {
		t.Fatal("rewritten adapter bypassed release binding")
	}
}
```

- [ ] **Step 2: Run the result test and verify RED**

Run: `go test ./internal/c12acceptance/coresmoke -run '^TestResultCannotSatisfyAnotherAdapter$' -count=1`

Expected: FAIL because `CoreSmokeResultV1` and its validator do not exist.

- [ ] **Step 3: GREEN — implement the finite adapter-bound result**

```go
type CoreSmokeResultV1 struct {
	SchemaVersion          string
	RunID                  uuid.UUID
	Adapter                contracts.Adapter
	ProfileID              string
	ReleaseLockDigest      contracts.Digest
	ServerImageDigest      contracts.Digest
	ClientImageDigest      contracts.Digest
	ProcessImageReceipt    contracts.Digest
	NonceMatched           bool
	NativeCheckPassed      bool
	HealthyProbePassed     bool
	ServerCgroupEmpty      bool
	ClientCgroupEmpty      bool
	EchoCgroupEmpty        bool
	CredentialCleanupProof contracts.Digest
}
```

Validation requires schema `talenro-core-smoke-result/v1`, nonzero run ID/digests, exact adapter/profile/release tuple, all booleans true, server/client image digests equal the reviewed lock entries, and a process-image receipt under the run-owned external-core volume. Reject unknown adapter, cross-profile data, zero cleanup proof and any additional JSON field.

- [ ] **Step 4: Run the result package and verify GREEN**

Run: `go test ./internal/c12acceptance/coresmoke -count=1`

Expected: PASS for exact Xray and sing-box records; all cross-adapter, cross-release and missing-cleanup mutations fail.

- [ ] **Step 5: REFACTOR — execute both independent real-core commands**

Run inside the approved Linux test-host, as two commands rather than one filtered aggregate:

```text
go test -tags=c12_core_smoke ./internal/c12acceptance/coresmoke -run '^TestXrayLoopbackSmoke$' -count=1 -timeout 2m -args -release-map /run/talenro-c12/installed-release-map.v1.json -result /run/talenro-c12/output/xray-result.json
go test -tags=c12_core_smoke ./internal/c12acceptance/coresmoke -run '^TestSingBoxLoopbackSmoke$' -count=1 -timeout 2m -args -release-map /run/talenro-c12/installed-release-map.v1.json -result /run/talenro-c12/output/sing-box-result.json
```

Expected: both PASS independently; deleting either result causes its own gate to fail, and neither result validates under the other adapter.

- [ ] **Step 6: Run repository-wide non-Docker regressions**

Run: `go test ./... -count=1 -timeout 10m`

Run: `go test -race ./internal/nodeagent/adapter/... ./internal/nodesupervisor/adapter ./internal/c12acceptance/coresmoke -count=1`

Run: `go vet ./...`

Run: `go tool golangci-lint run ./...`

Expected: all commands PASS.

- [ ] **Step 7: Commit the non-substitution guard**

```bash
git add internal/c12acceptance/coresmoke/result.go internal/c12acceptance/coresmoke/result_test.go internal/c12acceptance/coresmoke/xray_test.go internal/c12acceptance/coresmoke/singbox_test.go
git commit -m "test: keep external core evidence adapter bound"
```

## B08 and B09 exit gates

`C1.2-B08` closes only when Tasks 1–2 pass, the reviewed Xray lock/license record exists, `TestXrayLoopbackSmoke` proves a 32-byte nonce round trip within 5 seconds, and cleanup/process-image evidence validates.

`C1.2-B09` closes only when Tasks 3–4 pass, the reviewed sing-box lock/license record exists, `TestSingBoxLoopbackSmoke` proves a 32-byte nonce round trip within 10 seconds, and cleanup/process-image evidence validates.

Task 5 is required before B10: it proves the two completed gates remain distinct inputs to the canonical verifier. Neither Docker orchestration nor a config-check-only result may retroactively replace either real protocol smoke.
