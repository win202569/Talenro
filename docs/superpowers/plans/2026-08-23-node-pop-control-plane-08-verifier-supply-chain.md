# Talenro C1.2 Verifier and Supply Chain Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付唯一 canonical nested-Docker verifier、PowerShell 与 repository-external Git Bash 两个受锁 wrapper、可恢复 ownership WAL、封闭制品 scanner，以及不能冒充真实 Linux/provider 证据的两个 Windows scope evidence。

**Architecture:** 两个 Windows wrapper 只验证本机 toolchain/daemon、建立 run identity、驱动同一份 digest-pinned outer Compose 并执行 exact cleanup；所有 PostgreSQL、Linux fixture、real-core smoke、load 与扫描判定都在 `c12-verifier` 中由同一个 Bash authority script完成。Outer rootless inner daemon 与 host daemon 隔离；append-only WAL 在资源创建前写 intent、创建后写 actual，scanner 对三个 proprietary release、production/outer OCI 和 verifier test assets 做封闭集合比较。

**Tech Stack:** Go 1.26.5、PowerShell 5.1+、Git for Windows Bash 5+、Linux `/usr/bin/bash`、Docker Engine/Compose v2、rootless Docker-in-Docker、mTLS、OCI image layout、`debug/elf`、`debug/pe`、`debug/buildinfo`、RFC 8785 JCS、SHA-256、HMAC-SHA-256、Windows DPAPI。

**Spec:** [Approved C1.2 node and POP control-plane design](../specs/2026-08-23-node-pop-control-plane-design.md), especially sections 17.5–17.6, 18, 20, and 21 batch 10.

## Global Constraints

- 本册只实现 `C1.2-B10`，并消费已独立通过的 B08 Xray 与 B09 sing-box result；缺任一结果时 verifier fail closed。
- Windows scope 恰好为 `windows_powershell_docker` 与 `windows_git_bash_docker`；Git for Windows Bash 不能称为 Linux platform scope。
- Nested Docker provider 名称必须精确为 `DeterministicRollbackGuardFakeV1`、`DeterministicSecurityLatchGuardFakeV1`、`DeterministicTrustedTimeFakeV1`、`DeterministicHostMemoryIsolationPolicyFakeV1` 与 `DeterministicControlPlaneAuthorityFenceFakeV1`。
- Container fake evidence 的 provider class 固定为 `container_deterministic`；completion validator 必须拒绝它满足 `linux_platform_operator_trust` 或 `authority_fence_pitr`。
- Wrapper 先解析并验证 PowerShell/Git Bash/docker/Compose 的 canonical absolute path、SHA-256、签名/版本与 OS/arch；禁止 PATH alias、shim 和自动升级。
- 两个 wrapper 显式使用同一个 locked Docker context/endpoint；`DOCKER_HOST`、`DOCKER_CONTEXT`、`COMPOSE_FILE`、`COMPOSE_PROJECT_NAME` 必须为空或逐字等于 lock。
- Outer Compose 恰好三个 service：`c12-context-init`、`c12-inner-daemon`、`c12-verifier`；恰好一个 internal network 和三个 named volume：`inner-data`、`context-secrets`、`verifier-output`。
- 三个 outer image 必须以 OCI digest 固定并声明 `pull_policy: never`；wrapper 使用 `--pull never --no-build`，禁止 fallback pull/build/tag replacement。
- Host Docker allowlist 仅包含 locked context inspect/version/info、三个 exact image inspect、offline compose config、exact project up/ps/stop/rm/down `--volumes` 和 WAL exact name/ID inspect。禁止 pull/build/load/save/run/exec/cp/logs/events/system prune、list 与模糊枚举。
- Verifier 不挂 host Docker socket，只通过 per-run mTLS 连接 inner daemon；repo 只读、output volume 可写，inner daemon 独占空 data volume。
- Core/base OCI layouts 只存在 verifier image 的 test-assets layer；inner daemon 无 registry 网络，只允许 streaming `ImageLoad`，`ImagePull` 必须被策略和测试拒绝。
- 每次入口生成 128-bit random run ID，project 固定为 `talenro-c12-<32 lowercase hex>`，临时根必须重新解析为 OS temp 的直接子目录。
- WAL 是 append-only、hash-chained、HMAC-authenticated；intent 与 actual 都必须 file+directory fsync。HMAC key 使用 DPAPI 或 mode `0600` 文件保护且不写日志。
- Recovery 只查询 WAL 中 exact deterministic name/actual ID，禁止 label/name list；pre-existing collision、creation time/label/image mismatch 都停止并要求人工隔离。
- Cleanup deadline 固定 120 秒；exit bit 1 表示主测试失败，bit 2 表示 cleanup/ownership failure，两者可以同时保留。
- 入口保存 `git status --porcelain=v2 -z` baseline；结束要求逐字节相同，并拒绝新凭据/core artifact 未跟踪文件。
- Scanner version 固定 `talenro-artifact-scan/v1`；ruleset digest写入 toolchain lock。缺命令、不可解析输出、输入集合不全或 scanner/ruleset漂移都失败。
- 三个 proprietary release 固定 `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 -buildmode=exe`，禁止 `-linkshared`、external linker 和自定义 extldflags；Linux release不得有 `PT_INTERP`、`DT_NEEDED` 或动态 import closure。
- Scanner 不能用自由文本 `strings` 作为结论；必须解析 Go deps/embed、PE/ELF、OCI manifest/config/layer 与 archive payload。
- Xray/sing-box test asset、三个 outer image各自的SBOM/provenance/vulnerability/license/notice/source obligations必须完整入账；Redis production block保持不变。
- 每个 scope evidence 最长 72 小时并绑定 repo/tracked tree/spec/toolchain、三个 release、全部 image、test result 与 cleanup digest。
- 每个 task 使用独立 RED/GREEN/REFACTOR 与 commit；只 stage 明列文件，绝不 stage 用户未跟踪目录。

---

## File structure and frozen B10 interfaces

```text
internal/c12evidence/
├── scope.go                         # scope/provider class/evidence validation
├── scope_test.go
├── canonical.go                     # strict JSON + JCS digest
├── wal.go                           # intent/actual/cleaned append-only WAL
├── wal_test.go
├── cleanup.go                       # exact cleanup plan and exit bits
├── cleanup_test.go
└── wrapper_contract_test.go         # fake Docker and two-shell command contracts
internal/artifactscan/
├── rules.go                         # talenro-artifact-scan/v1 closed rules
├── godeps.go                        # go list/buildinfo/embed inventory
├── executable.go                    # ELF/PE/static-link inventory
├── oci.go                           # OCI manifest/config/layer traversal
├── scan.go                          # closed-set comparison and receipt
└── *_test.go
cmd/talenro-artifact-scan/main.go
deploy/c12/
├── compose.outer.yaml
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
├── toolchain-lock.v1.json
├── evidence/container-scope-valid.json
└── scanner/*.json
```

The evidence types introduced here are frozen for plan 09:

```go
package c12evidence

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
	ImageSetDigest     contracts.Digest
	TestResultDigest   contracts.Digest
	CleanupDigest      contracts.Digest
	Attestation        []byte
}

func ValidateScopeEvidence(C12ScopeEvidenceV1, TrustPolicy, time.Time) error
func CanonicalScopeEvidence(C12ScopeEvidenceV1) ([]byte, contracts.Digest, error)
```

### Task 1: Scope evidence schema and deterministic/production boundary

**Files:**
- Create: `internal/c12evidence/scope.go`
- Create: `internal/c12evidence/scope_test.go`
- Create: `internal/c12evidence/canonical.go`
- Create: `testdata/c12/evidence/container-scope-valid.json`
- Test: `internal/c12evidence/scope_test.go`

**Interfaces:**
- Consumes: `contracts.Digest`, strict JSON/JCS support from C1.1, and approved-spec digest from the suite index.
- Produces: `Scope`, `ProviderClass`, `C12ScopeEvidenceV1`, `TrustPolicy`, `ValidateScopeEvidence`, and `CanonicalScopeEvidence` exactly as frozen above.

- [ ] **Step 1: RED — write scope completeness and anti-substitution tests**

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
```

- [ ] **Step 2: Run the focused tests and verify RED**

Run: `go test ./internal/c12evidence -run '^TestContainerEvidence|^TestScopeEvidence' -count=1`

Expected: FAIL because scope evidence types and validator are undefined.

- [ ] **Step 3: GREEN — add the closed enums and strict validator**

Validation must require schema `talenro-c12-scope-evidence/v1`, one of the four exact scopes, nonzero run/times/digests, `StartedAt <= FinishedAt <= ExpiresAt <= FinishedAt+72h`, a trusted runner attestation over `TALENRO-C12-SCOPE-EVIDENCE-V1\x00 || JCS(payload)`, and exact expected repo/tree/spec/toolchain/release/image inputs. Windows scopes require `container_deterministic`; production scopes require `production`. Unknown JSON fields, duplicate fields, non-UTC timestamps and a scope/attestation mismatch fail closed.

- [ ] **Step 4: Run all evidence tests and verify GREEN**

Run: `go test ./internal/c12evidence -count=1`

Expected: PASS for the canonical fixture; field removal, field duplication, expiry, cross-scope rewrite, stale build and fake-provider mutations all fail.

- [ ] **Step 5: REFACTOR — fuzz strict scope decoding**

```go
func FuzzScopeEvidence(f *testing.F) {
	f.Add(readFixture(f, "container-scope-valid.json"))
	f.Fuzz(func(t *testing.T, input []byte) {
		_, _ = DecodeScopeEvidence(bytes.NewReader(input), 64<<10)
	})
}
```

Run: `go test ./internal/c12evidence -run '^$' -fuzz '^FuzzScopeEvidence$' -fuzztime=10s -timeout 30s`

Expected: PASS with no panic, unbounded allocation or raw-input error echo.

- [ ] **Step 6: Commit the evidence boundary**

```bash
git add internal/c12evidence/scope.go internal/c12evidence/scope_test.go internal/c12evidence/canonical.go testdata/c12/evidence/container-scope-valid.json
git commit -m "feat: define C1.2 scope evidence"
```

### Task 2: Append-only ownership WAL, exact recovery and cleanup exit bits

**Files:**
- Create: `internal/c12evidence/wal.go`
- Create: `internal/c12evidence/wal_test.go`
- Create: `internal/c12evidence/cleanup.go`
- Create: `internal/c12evidence/cleanup_test.go`
- Test: `internal/c12evidence/wal_test.go`
- Test: `internal/c12evidence/cleanup_test.go`

**Interfaces:**
- Consumes: validated 128-bit run ID, OS-keystore protected HMAC key, exact resource inspectors/deleters supplied by wrappers, and `contracts.Digest`.
- Produces: `OpenWAL`, `AppendIntent`, `AppendActual`, `AppendCleaned`, `RecoverExact`, `ExecuteCleanup`, and `ExitBits(primaryFailed, cleanupFailed bool) int`.

- [ ] **Step 1: RED — add hash-chain and intent-before-create tests**

```go
func TestWALRequiresFsyncedIntentBeforeActual(t *testing.T) {
	w := newTestWAL(t)
	resource := ResourceRef{Type: ResourceContainer, StableName: "talenro-c12-00112233445566778899aabbccddeeff-verifier"}
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

- [ ] **Step 2: Run WAL tests and verify RED**

Run: `go test ./internal/c12evidence -run '^TestWAL' -count=1`

Expected: FAIL because `OpenWAL` and record methods are undefined.

- [ ] **Step 3: GREEN — implement the canonical WAL record**

```go
type WALRecordV1 struct {
	SchemaVersion    string
	RunID            uuid.UUID
	Number           uint64
	PreviousDigest   contracts.Digest
	Phase            WALPhase
	ResourceType     ResourceType
	StableName       string
	ActualID         string
	ExpectedIdentity contracts.Digest
	CreatedAt        time.Time
	RecordHMAC       contracts.Digest
}
```

`WALPhase` is closed to `intent`, `actual`, `cleaned`, and `not_found`. Append one canonical line, `fsync` file, `fsync` parent directory, then return. HMAC covers the domain `TALENRO-C12-OWNERSHIP-WAL-V1\x00`, canonical record without `RecordHMAC`, previous digest and record number. Reject gaps, rewrites, duplicate actual, actual time before intent, stable-name/type/identity changes, invalid HMAC and trailing partial records.

- [ ] **Step 4: Run WAL tests and verify GREEN**

Run: `go test ./internal/c12evidence -run '^TestWAL' -count=1`

Expected: PASS, including crash after intent, crash after create, partial write, wrong key, record reorder and cross-run replay cases.

- [ ] **Step 5: RED — add exact recovery and cleanup precedence tests**

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

- [ ] **Step 6: Run cleanup tests and verify RED**

Run: `go test ./internal/c12evidence -run '^TestExitBits|^TestRecoverExact|^TestExecuteCleanup' -count=1`

Expected: FAIL because recovery and cleanup functions are undefined.

- [ ] **Step 7: GREEN — implement exact-only recovery and bounded cleanup**

`RecoverExact` may inspect only a stable name already present in an authenticated intent. It accepts a candidate only when exact name/type/run label/expected image or executable identity match and actual creation time is not before intent; it writes recovered `actual` before deletion. `not found` writes a terminal record. A mismatch returns `ErrIsolationRequired` without deletion or broader lookup.

`ExecuteCleanup` stops exact PID+start-token records first, then exact container IDs, network ID, volumes, scheduled task, and finally paths that re-resolve beneath the run root. It applies a single 120-second context and never calls global process/Docker/list APIs.

- [ ] **Step 8: Run WAL/cleanup tests and verify GREEN**

Run: `go test ./internal/c12evidence -run '^TestWAL|^TestExitBits|^TestRecoverExact|^TestExecuteCleanup' -count=1`

Expected: PASS; injected cleanup failure produces bit 2, primary+cleanup produces 3, and no test deletes the pre-existing collision fixture.

- [ ] **Step 9: REFACTOR — run race and Windows path validation tests**

Run: `go test -race ./internal/c12evidence -count=1`

Expected: PASS; WAL concurrent append attempts serialize by record number and path checks reject junction/reparse escape fixtures.

- [ ] **Step 10: Commit ownership semantics**

```bash
git add internal/c12evidence/wal.go internal/c12evidence/wal_test.go internal/c12evidence/cleanup.go internal/c12evidence/cleanup_test.go
git commit -m "feat: add exact C1.2 ownership recovery"
```

### Task 3: Closed-set Go, executable and OCI artifact scanner

**Files:**
- Create: `internal/artifactscan/rules.go`
- Create: `internal/artifactscan/godeps.go`
- Create: `internal/artifactscan/executable.go`
- Create: `internal/artifactscan/oci.go`
- Create: `internal/artifactscan/scan.go`
- Create: `internal/artifactscan/scan_test.go`
- Create: `cmd/talenro-artifact-scan/main.go`
- Create: `cmd/talenro-artifact-scan/main_test.go`
- Create: `testdata/c12/scanner/renamed-core.json`
- Create: `testdata/c12/scanner/archive-core.json`
- Create: `testdata/c12/scanner/base-layer-core.json`
- Create: `testdata/c12/scanner/forged-module.json`
- Test: `internal/artifactscan/scan_test.go`
- Test: `cmd/talenro-artifact-scan/main_test.go`

**Interfaces:**
- Consumes: three final proprietary binaries, production and outer OCI layouts, verifier asset inventory, `go list -deps -json`, `go version -m`, `git ls-files`, B08/B09 release locks and license records.
- Produces: scanner version `talenro-artifact-scan/v1`, `RulesetDigest() contracts.Digest`, `ScanClosedSet(context.Context, ScanRequest) (ScanReceiptV1, error)`, and CLI commands `lock-core`, `verify-go-deps`, `scan-release`, `scan-oci`, `scan-closed-set`, `verify-scope`, and later `verify-completion`.

- [ ] **Step 1: RED — add negative core-concealment tests**

```go
func TestScannerRejectsEveryCoreConcealment(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"renamed-core", "archive-core", "base-layer-core", "forged-module"} {
		t.Run(name, func(t *testing.T) {
			request := materializeNegativeFixture(t, name)
			if _, err := ScanClosedSet(t.Context(), request); !errors.Is(err, ErrForbiddenCoreMaterial) {
				t.Fatal("scanner accepted concealed core material")
			}
		})
	}
}
```

- [ ] **Step 2: Run scanner tests and verify RED**

Run: `go test ./internal/artifactscan -run '^TestScannerRejectsEveryCoreConcealment$' -count=1`

Expected: FAIL because scanner package and fixtures do not exist.

- [ ] **Step 3: GREEN — implement Go dependency and embed inventory**

Decode concatenated `go list -deps -json` records and require `ImportPath`, `Module.Path`, `GoFiles`, `CgoFiles`, `CompiledGoFiles`, `EmbedFiles`, `TestEmbedFiles`, and `XTestEmbedFiles` to be parseable. Reject either upstream module path or child path, any nonempty CgoFiles for the release build, any embedded file matching an approved core executable/layout/layer digest, or embedded PE/ELF/archive payload. Cross-check package/build settings against `go version -m` and the build receipt.

- [ ] **Step 4: Run Go dependency tests and verify GREEN**

Run: `go test ./internal/artifactscan -run '^TestGoDeps|^TestEmbed' -count=1`

Expected: PASS for repository packages and FAIL for the forged-module fixture.

- [ ] **Step 5: GREEN — implement executable and OCI traversal**

Use `debug/elf`, `debug/pe`, `debug/buildinfo`, `archive/tar` and OCI descriptor digests. For Linux releases require static ELF with no `PT_INTERP`, `DT_NEEDED`, dynamic imports or external-link receipt. Walk every OCI manifest/config/layer by verified digest; inventory regular files, symlinks, whiteouts, nested archives and executable headers. Reject a core digest/module namespace/executable payload in proprietary binary, production image or any unaccounted outer layer.

- [ ] **Step 6: Run executable/OCI tests and verify GREEN**

Run: `go test ./internal/artifactscan -run '^TestELF|^TestPE|^TestOCI|^TestScannerRejectsEveryCoreConcealment$' -count=1`

Expected: PASS for clean synthetic layouts; renamed, wrapper-loaded, archive, base-layer and forged-module fixtures all return `ErrForbiddenCoreMaterial`.

- [ ] **Step 7: GREEN — add the bounded CLI and machine-readable receipt**

```go
type ScanReceiptV1 struct {
	SchemaVersion        string
	ScannerVersion       string
	RulesetDigest        contracts.Digest
	ReleaseInputDigests  [3]contracts.Digest
	ProductionImageSet   contracts.Digest
	OuterImageSet        contracts.Digest
	VerifierAssetSet     contracts.Digest
	CoreAssetSet         contracts.Digest
	ProcessImageEvidence contracts.Digest
	Result               string
}
```

The only successful result is `clean_closed_set`. Limit each CLI input to an explicit repeated path flag, reject stdin for binary/OCI data, emit one canonical JSON receipt to the named output file, cap diagnostics at 4 KiB of finite categories, and never print file content, certificate, key or core output.

- [ ] **Step 8: Run CLI tests and verify GREEN**

Run: `go test ./cmd/talenro-artifact-scan ./internal/artifactscan -count=1`

Expected: PASS; missing input class, duplicate path, bad digest, ruleset mismatch and unparseable tool output each return nonzero without a success receipt.

- [ ] **Step 9: REFACTOR — fuzz OCI metadata and run static gates**

Run: `go test ./internal/artifactscan -run '^$' -fuzz '^FuzzOCIDescriptor$' -fuzztime=10s -timeout 30s`

Run: `go vet ./internal/artifactscan ./cmd/talenro-artifact-scan`

Expected: PASS with no panic or unbounded allocation.

- [ ] **Step 10: Commit scanner independently**

```bash
git add internal/artifactscan/rules.go internal/artifactscan/godeps.go internal/artifactscan/executable.go internal/artifactscan/oci.go internal/artifactscan/scan.go internal/artifactscan/scan_test.go cmd/talenro-artifact-scan/main.go cmd/talenro-artifact-scan/main_test.go testdata/c12/scanner/renamed-core.json testdata/c12/scanner/archive-core.json testdata/c12/scanner/base-layer-core.json testdata/c12/scanner/forged-module.json
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
- Create: `testdata/c12/toolchain-lock.v1.json`
- Create: `internal/c12evidence/compose_test.go`
- Test: `internal/c12evidence/compose_test.go`

**Interfaces:**
- Consumes: scanner ruleset/version, B08/B09 OCI layouts and locks, a validated run ID, and an empty `inner-data` volume.
- Produces: one audited outer Compose, three exact outer image digests, a per-run mTLS inner Docker context, and a verifier image test-assets inventory.

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
}
```

- [ ] **Step 2: Run the Compose contract and verify RED**

Run: `go test ./internal/c12evidence -run '^TestOuterComposeHasOnlyApprovedResources$' -count=1`

Expected: FAIL because `deploy/c12/compose.outer.yaml` does not exist.

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
    volumes:
      - ../..:/workspace:ro
      - context-secrets:/run/c12-context:ro
      - verifier-output:/run/c12-output
networks:
  c12-internal:
    internal: true
volumes:
  inner-data: {}
  context-secrets: {}
  verifier-output: {}
```

The committed file must contain generated immutable `name@sha256:<64 lowercase hex>` image references, not tags or interpolation. Resource names and ownership labels interpolate only the validated `C12_RUN_ID` supplied by the wrapper.

- [ ] **Step 4: GREEN — implement per-run mTLS and rootless daemon entrypoints**

Context-init creates a per-run CA, daemon server leaf and verifier client leaf in `context-secrets`, validates exact EKU/SAN/profile, writes no key to stdout, then exits. Inner daemon refuses a nonempty data root, binds only its internal-network TLS endpoint, requires client cert, disables registry access and content trust fallback, and writes readiness without secret material. Verifier uses only that mTLS context and has no host socket or inner storage mount.

- [ ] **Step 5: Build, scan and lock the three outer images**

Run in the audited offline image-build runner:

```text
go run ./cmd/talenro-artifact-scan lock-outer-images --compose deploy/c12/compose.outer.yaml --context-root deploy/c12 --output-lock testdata/c12/toolchain-lock.v1.json --rewrite-compose
```

Expected: exit 0 only after each image/base layer has a verified digest, SBOM, provenance, vulnerability result, license/notice/source obligation and secret scan; the command rewrites exact image digests and toolchain-lock digests atomically. Missing review evidence exits 1 without changing either file.

- [ ] **Step 6: Run offline Compose and scanner contracts and verify GREEN**

Run: `docker --context talenro-c12-locked compose -f deploy/c12/compose.outer.yaml config --quiet`

Run: `go test ./internal/c12evidence -run '^TestOuterCompose' -count=1`

Run: `go run ./cmd/talenro-artifact-scan scan-oci --lock testdata/c12/toolchain-lock.v1.json --compose deploy/c12/compose.outer.yaml`

Expected: all PASS; no service has a host port/extra mount, all images resolve to locked IDs, and verifier core assets are fully inventoried but absent from context-init/daemon and all production inputs.

- [ ] **Step 7: REFACTOR — prove inner registry pull is unavailable**

Run within the verifier contract fixture: `go test -tags=c12_docker ./internal/c12evidence -run '^TestInnerDaemonRejectsImagePull$' -count=1 -timeout 2m`

Expected: PASS only when streaming `ImageLoad` succeeds for locked layouts, `ImagePull` returns the fixed denied category, and the inner network cannot resolve or connect to a registry.

- [ ] **Step 8: Commit the outer substrate**

```bash
git add deploy/c12/compose.outer.yaml deploy/c12/context-init.Dockerfile deploy/c12/inner-daemon.Dockerfile deploy/c12/verifier.Dockerfile deploy/c12/context-init-entrypoint.sh deploy/c12/inner-daemon-entrypoint.sh deploy/c12/verifier-entrypoint.sh testdata/c12/toolchain-lock.v1.json internal/c12evidence/compose_test.go
git commit -m "build: add isolated C1.2 verifier substrate"
```

### Task 5: Canonical inner Bash authority script

**Files:**
- Create: `scripts/c12-authority.sh`
- Create: `internal/c12evidence/authority_script_test.go`
- Modify: `deploy/c12/verifier-entrypoint.sh`
- Test: `internal/c12evidence/authority_script_test.go`

**Interfaces:**
- Consumes: inner mTLS Docker context, B01–B09 code/tests, B08/B09 result validators, toolchain lock, scanner, run WAL and output volume.
- Produces: one canonical container-deterministic `C12ScopeEvidenceV1` payload plus digest, shared by both Windows wrappers.

- [ ] **Step 1: RED — add exact stage-order and fail-closed tests**

```go
func TestAuthorityScriptRunsEveryRequiredStageInOrder(t *testing.T) {
	transcript := runAuthorityWithFakeTools(t)
	assertOrdered(t, transcript,
		"check-tools", "generated-diff", "unit", "fuzz", "race", "vet", "lint",
		"migration-roundtrip", "integration", "mtls-contract", "controlled-process",
		"xray-smoke", "sing-box-smoke", "load-1000", "privacy", "artifact-scan",
		"ownership-cleanup", "worktree-baseline", "scope-evidence")
}
```

- [ ] **Step 2: Run authority-script contracts and verify RED**

Run: `go test ./internal/c12evidence -run '^TestAuthorityScript' -count=1`

Expected: FAIL because `scripts/c12-authority.sh` is missing.

- [ ] **Step 3: GREEN — implement bounded stage execution**

The script uses `set -euo pipefail`, a validated run root, GNU `timeout --kill-after=5s`, bounded stdout/stderr sinks, and one trap preserving `ExitBits`. It runs exact fixed-tool/generation/unit/fuzz/race/vet/lint/migration/integration/mTLS/fixture/Xray/sing-box/load/privacy/scanner commands. Xray and sing-box remain separate commands and result files. It loads inner images only by streaming verified OCI layouts, records every resource intent/actual before use, and never calls `ImagePull`.

- [ ] **Step 4: GREEN — emit one bounded canonical evidence line**

On success, write signed canonical evidence to `/run/c12-output/scope-evidence.json`, then emit exactly one line `C12_SCOPE_EVIDENCE_B64=<base64url-without-padding>`. The payload is at most 64 KiB and contains no key, cert bytes, credentials, config, core output, path or node ID. All other service output is empty or fixed finite stage names.

- [ ] **Step 5: Run script contract and injected-failure matrix and verify GREEN**

Run: `go test ./internal/c12evidence -run '^TestAuthorityScript' -count=1`

Expected: PASS; each injected stage failure stops later test stages, cleanup always runs, primary and cleanup bits are preserved, missing Xray or sing-box result fails, and no canary appears in captured output.

- [ ] **Step 6: REFACTOR — run shell syntax and Docker-tagged smoke contract**

Run: `bash -n scripts/c12-authority.sh deploy/c12/verifier-entrypoint.sh`

Run: `go test -tags=c12_docker ./internal/c12evidence -run '^TestAuthorityScriptAgainstInnerDaemon$' -count=1 -timeout 15m`

Expected: PASS with exact inner resource cleanup and unchanged repo baseline.

- [ ] **Step 7: Commit canonical authority logic**

```bash
git add scripts/c12-authority.sh deploy/c12/verifier-entrypoint.sh internal/c12evidence/authority_script_test.go
git commit -m "test: add canonical C1.2 Docker authority"
```

### Task 6: Locked PowerShell wrapper and Windows Job Object scope

**Files:**
- Create: `scripts/verify-c12.ps1`
- Create: `internal/c12evidence/powershell_wrapper_test.go`
- Create: `internal/c12evidence/wrapper_contract_test.go`
- Test: `internal/c12evidence/powershell_wrapper_test.go`

**Interfaces:**
- Consumes: toolchain lock, outer Compose, WAL/cleanup functions, Windows RunnerAttestor, exact Docker context and canonical evidence line.
- Produces: authoritative `windows_powershell_docker` `C12ScopeEvidenceV1` and a Windows-native Job Object portability result bound into its test digest.

- [ ] **Step 1: RED — add PowerShell path, command allowlist and cleanup tests**

The fake `docker.exe` transcript must require exact ordered calls: context inspect/version/info, three image inspect calls, compose config, exact-project up/ps/stop/rm/down `--volumes`, and exact-ID inspect. Tests inject PATH aliases, daemon changes, pre-existing names, primary failure and cleanup failure.

- [ ] **Step 2: Run PowerShell wrapper tests and verify RED**

Run on Windows: `go test ./internal/c12evidence -run '^TestPowerShellWrapper' -count=1`

Expected: FAIL because `scripts/verify-c12.ps1` does not exist.

- [ ] **Step 3: GREEN — implement locked preflight and run ownership**

PowerShell resolves its own executable, `docker.exe`, Compose plugin and repo root without PATH fallback; compares path/hash/signature/version/OS/arch to `toolchain-lock.v1.json`; clears or validates Docker/Compose overrides; records `git status --porcelain=v2 -z`; creates a DPAPI-protected HMAC key and direct-child temp root; writes WAL intent before each outer resource command.

- [ ] **Step 4: GREEN — run native Job Object portability before Docker**

Invoke the B06 Windows fixture using `PROC_THREAD_ATTRIBUTE_JOB_LIST` and `EXTENDED_STARTUPINFO_PRESENT` so the child enters a kill-on-close Job Object at `CreateProcess`. Failure to use that attribute, handle inheritance, escaped grandchild or cleanup residue fails the scope. The result is Windows portability evidence only.

- [ ] **Step 5: GREEN — invoke only the outer Compose and canonical verifier**

Use `docker --context <locked-name> compose --project-name talenro-c12-<run-id> -f <absolute-locked-compose> up --abort-on-container-exit --exit-code-from c12-verifier --pull never --no-build`; parse one bounded canonical evidence line, validate it, add PowerShell/runner/daemon/Job Object digests, attest it, then execute exact cleanup and byte-equal worktree check.

- [ ] **Step 6: Run PowerShell wrapper tests and verify GREEN**

Run: `go test ./internal/c12evidence -run '^TestPowerShellWrapper' -count=1`

Expected: PASS; every forbidden Docker command, PATH alias, daemon drift, missing image, malformed evidence, worktree drift and cleanup residue returns nonzero with finite sanitized output.

- [ ] **Step 7: REFACTOR — run the real PowerShell Docker scope**

Run from the repository root:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify-c12.ps1 -EvidenceOutput "$env:TEMP\talenro-c12-powershell-scope.json"
```

Expected: all canonical verifier stages PASS; exact cleanup completes in 120 seconds; repository status equals baseline; output validates as unexpired `windows_powershell_docker` with `provider_class=container_deterministic`.

- [ ] **Step 8: Commit the PowerShell wrapper independently**

```bash
git add scripts/verify-c12.ps1 internal/c12evidence/powershell_wrapper_test.go internal/c12evidence/wrapper_contract_test.go
git commit -m "test: add locked PowerShell C1.2 verifier"
```

### Task 7: Repository-external Git Bash wrapper and B10 two-scope exit

**Files:**
- Create: `scripts/verify-c12.sh`
- Create: `internal/c12evidence/bash_wrapper_test.go`
- Modify: `internal/c12evidence/wrapper_contract_test.go`
- Test: `internal/c12evidence/bash_wrapper_test.go`

**Interfaces:**
- Consumes: the same toolchain lock, outer Compose, canonical authority script and evidence validator as Task 6; it may not invoke PowerShell wrapper logic.
- Produces: authoritative `windows_git_bash_docker` evidence and the B10 assertion that both wrapper evidences share exact daemon/repo/tree/spec/toolchain/release/image inputs while having distinct run IDs/scopes.

- [ ] **Step 1: RED — add repository-external Git Bash contract tests**

Test invocation directory must be a temp directory outside the repository. Require `BASH_SOURCE` canonical repo resolution, Git-for-Windows path conversion only for native executable arguments, fixed `bash.exe` identity from the lock, exact Docker transcript parity with PowerShell, and no call to `powershell`, `cmd /c`, `eval` or a PATH alias.

- [ ] **Step 2: Run Bash wrapper tests and verify RED**

Run on Windows: `go test ./internal/c12evidence -run '^TestGitBashWrapper' -count=1`

Expected: FAIL because `scripts/verify-c12.sh` does not exist.

- [ ] **Step 3: GREEN — implement repository-external resolution and locked preflight**

Use `set -euo pipefail`; resolve `script_dir`/`repo_root` via `pwd -P`; validate Git Bash executable identity and native Docker/Compose paths from the same lock; use `cygpath -w` only after rejecting control characters; validate overrides, daemon identity, temp root, run ID and Git baseline before any create command.

- [ ] **Step 4: GREEN — mirror the exact outer lifecycle without copying test logic**

Invoke the same `compose.outer.yaml` and the same verifier image/canonical script. Bash owns only preflight, WAL, outer commands, evidence attestation and cleanup. Its allowed Docker command-class transcript must byte-normalize to the PowerShell transcript after run ID/path replacement.

- [ ] **Step 5: Run Bash wrapper and cross-wrapper tests and verify GREEN**

Run: `go test ./internal/c12evidence -run '^TestGitBashWrapper|^TestWrapperParity' -count=1`

Expected: PASS; injected primary/cleanup failures preserve bits, fake output canaries remain absent, and either wrapper rejects the other scope enum.

- [ ] **Step 6: REFACTOR — execute Git Bash from outside the repository**

Run from an OS temp directory using the locked Git for Windows Bash:

```text
"C:\Program Files\Git\bin\bash.exe" /d/Projects/Talenro/scripts/verify-c12.sh --evidence-output /d/c12-evidence/talenro-c12-git-bash-scope.json
```

Expected: canonical verifier PASS, exact cleanup, unchanged worktree baseline, and valid `windows_git_bash_docker` evidence with a run ID distinct from PowerShell.

- [ ] **Step 7: Validate the two B10 scope inputs together**

Run:

```text
go run ./cmd/talenro-artifact-scan verify-scope --evidence C:/Users/runner/AppData/Local/Temp/talenro-c12-powershell-scope.json --evidence D:/c12-evidence/talenro-c12-git-bash-scope.json --require-scope windows_powershell_docker --require-scope windows_git_bash_docker --require-same-build
```

Expected: PASS only when scopes are distinct, daemon/repo/tree/spec/toolchain/release/image inputs are equal, both are unexpired/attested, cleanup succeeded, and provider class is `container_deterministic`. This command does not claim C1.2 completion.

- [ ] **Step 8: Run full non-provider regressions**

Run: `go test ./... -count=1 -timeout 10m`

Run: `go test -race ./internal/c12evidence ./internal/artifactscan -count=1`

Run: `go vet ./...`

Run: `go tool golangci-lint run ./...`

Run: `git diff --check`

Expected: all PASS.

- [ ] **Step 9: Commit the Git Bash wrapper and B10 exit guard**

```bash
git add scripts/verify-c12.sh internal/c12evidence/bash_wrapper_test.go internal/c12evidence/wrapper_contract_test.go
git commit -m "test: add locked Git Bash C1.2 verifier"
```

## B10 exit gate

B10 closes only when both real Windows wrapper runs complete independently from the same tracked tree/toolchain/build and `verify-scope` accepts them together. The exit record must state that nested-Docker provider evidence is `container_deterministic`, that Xray and sing-box each passed its own real protocol smoke, that scanner inputs form the complete closed set, and that no host SELinux/sysctl/TPM/secure-time/operator-trust/authority-fence claim has been made.
