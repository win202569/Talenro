# Talenro C1.2 Operations and Completion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付完整 C1.2 threat model/runbooks、真实 Linux platform 与 OperatorClientTrustGuard conformance、真实 ControlPlaneAuthorityFence/PITR conformance，以及只在四个同构建 scope 全部有效时才能生成的完成清单和路线图推进。

**Architecture:** 文档门先把安全事件、有限操作、验证与停止条件冻结；Linux-native gate 在 attested `linux/amd64` TPM/vTPM runner 上跨 cold reboot 验证三个独立 anti-rollback identity、production secure time、host memory policy和真实外部攻击，再验证 production operator trust guard。独立 authority gate 使用 production provider class 的隔离 conformance tenant 和真实 PostgreSQL PITR；最终 validator 重验四份 scope 的签名、时效、输入一致性、license 与 cleanup，不能拼接或提升 fake evidence。

**Tech Stack:** Go 1.26.5、Linux `/usr/bin/bash`、TPM/vTPM NV counter与keystore、production TrustedTimeSource、SELinux enforcing、Yama 3、BPF lockdown、systemd collector mask、pidfd/perf/BPF/ptrace negative harness、PostgreSQL 18.4 PITR、ControlPlaneAuthorityFence、OperatorClientTrustGuard、RFC 8785 JCS、Ed25519 runner attestation、SHA-256。

**Spec:** [Approved C1.2 node and POP control-plane design](../specs/2026-08-23-node-pop-control-plane-design.md), especially sections 6.1, 9.1, 12, 17.4, 18–21.

## Global Constraints

- 本册只实现 `C1.2-B11`；只有在 B01–B10 全部关闭后执行。
- Production completion 恰好需要四个 scope：`windows_powershell_docker`、`windows_git_bash_docker`、`linux_platform_operator_trust`、`authority_fence_pitr`。
- 四个 scope 必须有不同 run ID，但 repo commit、tracked-tree、approved spec、toolchain、三个 release binary和全部 image digest逐字相同；任一证据最长 72 小时。
- Container deterministic fake、Windows Job Object、Git Bash 与 Docker test-host 都不能满足真实 Linux/platform/provider 或 authority-fence scope。
- Linux platform gate 必须在固定、attested `linux/amd64` runner 上由 native `/usr/bin/bash` 执行；production target不扩展到Windows。
- Agent main、supervisor 与 LocalSecurityLatch 分配三个独立 run-scoped NV counter/keystore identity；任一 identity 相等、handle cross-swap、blob/counter不匹配或 rollback 都 fail closed。
- Cross-reboot resume manifest 在访问任何 NV handle 前验证 authenticated seal、runner/provider identity、phase、run ID、exact handle digest、counter/digest floor与toolchain hash；禁止 TPM 全局 enumeration。
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
- External VM、NV handles、PITR数据库、临时cert/credential/state都写B10 ownership WAL并按exact ID清理；provider digest record按policy保留且不得含secret。
- Threat model/runbook不能要求直接修改数据库、运行远程shell、降低host policy或绕过provider；root/kernel/hypervisor/deployment主体失陷时唯一支持恢复是外部隔离、重建和reenrollment。
- Xray MPL-2.0、sing-box GPLv3+、三个outer image与verifier test-assets obligation必须处于通过状态；Redis 8.8.1 production block不得被C1.2完成绕过。
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
scripts/
├── verify-c12-platform.sh
└── verify-c12-authority-fence.sh
testdata/c12/platform/runner-profile.v1.json
testdata/c12/authority/runner-profile.v1.json
docs/roadmap/implementation-sequence.md
```

The production evidence boundaries are:

```go
package c12evidence

type PlatformEvidenceV1 struct {
	SchemaVersion              string
	RunID                      uuid.UUID
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
	ReleaseAndToolchainDigest  contracts.Digest
	CleanupResultDigest        contracts.Digest
}

type AuthorityFenceEvidenceV1 struct {
	SchemaVersion                    string
	RunID                            uuid.UUID
	StartedAt                        time.Time
	FinishedAt                       time.Time
	ExpiresAt                        time.Time
	ProviderIdentityPolicyVersion    string
	ConformanceTenant                string
	ControlAPIBinaryAndImageDigest   contracts.Digest
	AuthorityClientProtocolVersion   string
	AuthorityRulesetDigest           contracts.Digest
	DBSystemTimelineDigest           contracts.Digest
	DBSchemaOrderedMigrationDigest   contracts.Digest
	PITRBackupRestoreDigest          contracts.Digest
	RepoCommitAndTrackedTreeDigest   contracts.Digest
	ToolchainLockDigest              contracts.Digest
	TestCaseResultDigest             contracts.Digest
	TerminalProviderRecordDigest     contracts.Digest
	CleanupResultDigest              contracts.Digest
	RunnerAttestation                []byte
}
```

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
- Consumes: threat IDs C12-T01–T09, node/operator certificate APIs, root/metadata workflows, host-deployed trust packages, OperatorClientTrustGuard and ControlPlaneAuthorityFence recovery APIs.
- Produces: bounded procedures for certificate compromise, signer/root/CA/trust-package/operator-guard/deployment-authority incidents, fence outage/PITR and administrative restore.

- [ ] **Step 1: RED — add runbook scenario and section tests**

Require these exact procedure IDs: `C12-R01 node certificate compromise`, `R02 operator certificate compromise`, `R03 signer rotation`, `R04 root emergency ceremony`, `R05 CA issuer outage`, `R06 server CA rotation/removal`, `R07 operator guard outage/rollback`, `R08 deployment authority compromise`, `R09 fence outage/PITR`, `R10 administrative disable/restore`.

- [ ] **Step 2: Run the runbook contract and verify RED**

Run: `go test ./internal/c12evidence -run '^TestIdentityTrustAuthorityRunbook$' -count=1`

Expected: FAIL because the runbook is absent.

- [ ] **Step 3: GREEN — write the ten bounded procedures**

Every procedure uses exact headings `Trigger signals`, `Required role and scope`, `Finite operator actions`, `Verification`, `Stop conditions`, `Escalation`, and `Evidence retained`. Actions call only approved operator/security-admin APIs or out-of-band deployment provider operations; they never issue SQL, mutate a database row, run remote shell, reuse old certificates, accept a lower trust version or bypass uncached authorization.

- [ ] **Step 4: GREEN — encode destructive restore and two-person reauthorization**

`C12-R09` must keep listeners/signers closed after PITR until the external fence proves a nondecreasing authority head; destructive restore disables every node and discards old resume/desired intent. `C12-R10` requires completed stopped reenrollment, fresh host evidence, two distinct uncached security-admin credentials within 15 minutes, same epoch/effect and a new higher desired generation.

- [ ] **Step 5: Run runbook tests and verify GREEN**

Run: `go test ./internal/c12evidence -run '^TestIdentityTrustAuthorityRunbook|^TestRunbooksForbidDirectMutation$' -count=1`

Expected: PASS; all required roles/actions/stops/evidence fields exist and forbidden direct-mutation patterns are absent.

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
- Create: `internal/c12evidence/platform.go`
- Create: `internal/c12evidence/platform_test.go`
- Create: `testdata/c12/platform/runner-profile.v1.json`
- Test: `internal/c12evidence/platform_test.go`

**Interfaces:**
- Consumes: `C12ScopeEvidenceV1`, B10 WAL, three production counter/sealer identities, TrustedTimeSource identity, runner attestor and host policy package.
- Produces: `PlatformEvidenceV1`, `CrossRebootResumeManifestV1`, `ValidatePlatformEvidence`, `SealResumeManifest`, and `OpenResumeManifest`.

- [ ] **Step 1: RED — add three-identity and pre-handle seal tests**

```go
func TestResumeManifestRejectsCounterIdentityReuse(t *testing.T) {
	t.Parallel()
	manifest := validResumeManifest()
	manifest.CounterIdentityDigests[1] = manifest.CounterIdentityDigests[0]
	if _, err := SealResumeManifest(testContext(), manifest, testGateSealer()); !errors.Is(err, ErrIdentityReuse) {
		t.Fatal("resume manifest reused an anti-rollback identity")
	}
}

func TestOpenResumeManifestAuthenticatesBeforeHandleAccess(t *testing.T) {
	provider := newHandleAuditProvider(t)
	_, err := OpenResumeManifest(testContext(), tamperedResumeBlob(t), provider)
	if err == nil || provider.HandleAccesses() != 0 {
		t.Fatal("tampered resume manifest reached an NV handle")
	}
}
```

- [ ] **Step 2: Run platform-contract tests and verify RED**

Run: `go test ./internal/c12evidence -run '^TestResumeManifest|^TestPlatformEvidence' -count=1`

Expected: FAIL because platform evidence and resume types are undefined.

- [ ] **Step 3: GREEN — implement the cross-reboot manifest**

```go
type CrossRebootResumeManifestV1 struct {
	SchemaVersion          string
	RunID                  uuid.UUID
	RunnerIdentity         contracts.Digest
	ProviderIdentitySet    contracts.Digest
	Phase                   string
	CounterHandleDigests    [3]contracts.Digest
	CounterIdentityDigests  [3]contracts.Digest
	ExpectedCounterFloors   [3]uint64
	ExpectedStateDigests    [3]contracts.Digest
	TrustedTimeFloor        time.Time
	ToolchainDigest         contracts.Digest
	PhaseABootID            string
}
```

Only phase `awaiting_cold_reboot` is serializable. Seal with a gate-specific keystore identity and domain `TALENRO-C12-PLATFORM-RESUME-V1\x00`; validate runner/provider/toolchain and three pairwise-distinct identities before any handle lookup.

- [ ] **Step 4: GREEN — implement full platform evidence validation**

Require schema `talenro-c12-platform-evidence/v1`, distinct VM and boot IDs, three increasing counter floors, nondecreasing trusted time, exact signed host policy, all five sysctl/unit/policy digests, replay/attacker/post-start/operator-guard/cleanup digests, production provider set and runner attestation. Missing any negative case is diagnostic-only and returns `ErrIncompletePlatformGate`.

- [ ] **Step 5: Run platform tests and verify GREEN**

Run: `go test ./internal/c12evidence -run '^TestResumeManifest|^TestPlatformEvidence' -count=1`

Expected: PASS; missing replay, same boot, swapped counter, fake provider, missing attacker, stale policy and failed cleanup cases all reject.

- [ ] **Step 6: REFACTOR — fuzz resume/evidence strict decoding**

Run: `go test ./internal/c12evidence -run '^$' -fuzz '^FuzzPlatformEvidence$' -fuzztime=10s -timeout 30s`

Expected: PASS without panic, input echo or handle access on unauthenticated input.

- [ ] **Step 7: Commit production platform evidence types**

```bash
git add internal/c12evidence/platform.go internal/c12evidence/platform_test.go testdata/c12/platform/runner-profile.v1.json
git commit -m "feat: define C1.2 platform evidence"
```

### Task 5: Real Linux platform, memory isolation and OperatorClientTrustGuard gate

**Files:**
- Create: `scripts/verify-c12-platform.sh`
- Create: `internal/c12evidence/platform_gate_test.go`
- Create: `internal/nodecontrol/hostevidence/operator_guard_conformance_test.go`
- Modify: `cmd/c12-fixture/main.go`
- Test: `internal/c12evidence/platform_gate_test.go`
- Test: `internal/nodecontrol/hostevidence/operator_guard_conformance_test.go`

**Interfaces:**
- Consumes: signed runner profile, production counter/sealer/time/host-policy/operator-guard providers, B10 WAL, approved releases and `PlatformEvidenceV1`.
- Produces: attested `linux_platform_operator_trust` `C12ScopeEvidenceV1` with `provider_class=production`.

- [ ] **Step 1: RED — add phase, replay and external-attacker contract tests**

Require cases `unsafe-preflight`, `phase-a`, `phase-b`, `same-uid-attacker`, `different-uid-attacker`, `canary-crash`, `operator-guard`, and `cleanup`. The fake production-provider fixture must be rejected before any test case runs.

- [ ] **Step 2: Run tagged platform contracts and verify RED**

Run on the attested Linux runner: `/usr/bin/bash -c 'go test -tags=c12_platform ./internal/c12evidence ./internal/nodecontrol/hostevidence -count=1 -timeout 20m'`

Expected: FAIL because `verify-c12-platform.sh` and production cases are absent.

- [ ] **Step 3: GREEN — implement unsafe negative VM preflight**

Run each unsafe profile as a separate fresh boot: Yama absent/non-3, unprivileged BPF non-1, suid dumpable nonzero, nonempty regular core pattern, pipe core pattern, core_uses_pid nonzero, collector unmasked, SELinux permissive, wrong policy digest, wrong domain map and perf/BPF permission. Assert agent/supervisor fail before secret read, credential generation or child create; record only finite case/result digests, then destroy the VM and WAL-record its exact destroy receipt.

- [ ] **Step 4: GREEN — implement Phase A and authenticated cold-reboot handoff**

Allocate three handles only from the reserved conformance namespace, verify pairwise distinct provider identities, write agent/supervisor/latch state, active local/supervisor faults and transition capsule, record intent/actual, seal the resume manifest, fsync it and request a controlled cold reboot. Do not enumerate TPM handles and do not emit key/state bytes.

- [ ] **Step 5: GREEN — implement Phase B replay and time tests**

After reboot, authenticate manifest/runner/provider/toolchain before exact handle access. Reject old agent/supervisor/latch blob replay, counter/blob mismatch, every pairwise cross-swap, highest-verified/highest-applied/rollback-capsule replay, power-loss window and latch damage. Exercise wall-clock rollback, permitted VM snapshot rollback and production time-provider unavailable; no core may run before both latches clear and server finalization/explicit resume.

- [ ] **Step 6: GREEN — implement positive host policy and post-start receipts**

Create a fresh positive VM from the same attested image/toolchain. Before secrets, require SELinux enforcing, exact policy/domain-map, Yama 3, unprivileged BPF disabled, suid dumpable 0, empty core pattern, core_uses_pid 0 and collector mask; compare before/after reboot readbacks. For approved core, native-check and smoke-client record pinned executable FD, post-start pidfd image, UID/GID/capability, SELinux domain, NoNewPrivs, seccomp, cgroup and prlimit receipts; crash each canary and prove no handler/artifact/secret result.

- [ ] **Step 7: GREEN — implement external attacker cases outside target seccomp**

Launch one attacker with exact target core UID/domain and another with different UID/allowed core domain, both different PID/cgroup and neither inheriting target seccomp. Attempt `PERF_SAMPLE_STACK_USER`, user registers/high-frequency sampling, ptrace, both `process_vm_*`, `/proc/<pid>/mem`, pidfd_getfd, BPF map and program load. Each must fail at LSM/syscall boundary, produce no perf sample/BPF object, and leave target running.

- [ ] **Step 8: GREEN — implement production OperatorClientTrustGuard conformance**

Against the real provider, test lower package, same-version fork, client disk rollback then restart, provider restart, normal CA overlap, emergency removal, provider unavailable, nonce replay and expired attestation. Current attested digest is the only accepted state; both existing and new transports close on unavailable/replay/expiry.

- [ ] **Step 9: Run the production platform gate across cold reboot**

Phase A command:

```text
sudo /usr/bin/bash scripts/verify-c12-platform.sh phase-a --runner-profile testdata/c12/platform/runner-profile.v1.json --provider-config /run/secrets/talenro-c12-platform/provider.json --resume-root /var/lib/talenro-c12-platform
```

After the controlled cold reboot, Phase B command:

```text
sudo /usr/bin/bash scripts/verify-c12-platform.sh phase-b --runner-profile testdata/c12/platform/runner-profile.v1.json --provider-config /run/secrets/talenro-c12-platform/provider.json --resume-root /var/lib/talenro-c12-platform --evidence-output /var/lib/talenro-c12-platform/linux-platform-scope.json
```

Expected: both phases exit 0, boot IDs differ, provider/VM identity remains valid, all negative/positive/attacker/operator-guard cases pass, and the output validates as unexpired production `linux_platform_operator_trust`.

- [ ] **Step 10: REFACTOR — verify exact cleanup and no pending handle**

Run: `/usr/bin/bash scripts/verify-c12-platform.sh verify-cleanup --evidence /var/lib/talenro-c12-platform/linux-platform-scope.json`

Expected: PASS only after exact owner revalidation, undefine of all three run handles, run-key/state deletion, both VM destroy receipts, no pending WAL intent and cleanup digest success. The two one-way sysctls are not lowered online.

- [ ] **Step 11: Commit the Linux/operator production gate**

```bash
git add scripts/verify-c12-platform.sh internal/c12evidence/platform_gate_test.go internal/nodecontrol/hostevidence/operator_guard_conformance_test.go cmd/c12-fixture/main.go
git commit -m "test: add production Linux C1.2 platform gate"
```

### Task 6: Real ControlPlaneAuthorityFence and PostgreSQL PITR gate

**Files:**
- Create: `internal/c12evidence/authority.go`
- Create: `internal/c12evidence/authority_test.go`
- Create: `internal/nodecontrol/authority/production_conformance_test.go`
- Create: `scripts/verify-c12-authority-fence.sh`
- Create: `testdata/c12/authority/runner-profile.v1.json`
- Test: `internal/c12evidence/authority_test.go`
- Test: `internal/nodecontrol/authority/production_conformance_test.go`

**Interfaces:**
- Consumes: frozen `authority.Provider`, real production provider class/policy in an isolated tenant, PostgreSQL PITR harness, B10 WAL and runner attestor.
- Produces: exact `AuthorityFenceEvidenceV1`, validator, and attested `authority_fence_pitr` scope evidence.

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
```

- [ ] **Step 2: Run authority evidence tests and verify RED**

Run: `go test ./internal/c12evidence -run '^TestAuthorityEvidence' -count=1`

Expected: FAIL because authority evidence validator is undefined.

- [ ] **Step 3: GREEN — implement strict authority evidence validation**

Require schema `talenro-authority-fence-evidence/v1`, production provider identity/policy/version, isolated tenant, exact control API binary/image/client protocol/ruleset/DB system+timeline/schema+migration/PITR/repo/tree/toolchain/test/terminal/cleanup digests, `ExpiresAt <= FinishedAt+72h`, and trusted runner attestation over `TALENRO-AUTHORITY-FENCE-EVIDENCE-V1\x00 || JCS(payload)`.

- [ ] **Step 4: Run authority evidence tests and verify GREEN**

Run: `go test ./internal/c12evidence -run '^TestAuthorityEvidence' -count=1`

Expected: PASS; fake provider, production tenant, stale build, old protocol/ruleset, failed cleanup and tampered attestation reject.

- [ ] **Step 5: RED — add production provider crash/PITR conformance cases**

Cases must cover concurrent total order/no duplicate sequence, same-operation idempotency, changed scope/effect conflict, receipt replay/tamper, illegal abort, provider outage, crash before/after DB commit, crash before/after provider finalize, and pre-revoke/pre-disable PITR.

- [ ] **Step 6: Run tagged provider conformance and verify RED**

Run on the authority runner:

```text
/usr/bin/bash -c 'go test -tags=c12_authority ./internal/nodecontrol/authority -run "^TestProductionAuthority" -count=1 -timeout 30m'
```

Expected: FAIL because production conformance harness is absent.

- [ ] **Step 7: GREEN — implement isolated tenant and crash matrix**

Read provider configuration only from `/run/secrets/talenro-c12-authority/provider.json`; require conformance tenant marker and reject production namespace. Create run-scoped credentials and WAL records. Drive exact `Reserve/Finalize/Abort/Inspect/Head` operations through every crash point, asserting DB visibility matches finalized external receipts and no changed effect reuses an operation.

- [ ] **Step 8: GREEN — implement real PITR refusal proof**

Create a backup before a real node/operator revoke and desired disable, commit/finalize those effects, restore the old backup into an exact run-owned PostgreSQL instance, and point the control plane at it while retaining the external provider head. Assert all agent/operator listeners and signers remain closed, old node/operator certs reject and old desired is not served. Record backup/restore/system/timeline/migration digests; never alter provider head to match the old DB.

- [ ] **Step 9: Run the real authority/PITR gate**

Run:

```text
/usr/bin/bash scripts/verify-c12-authority-fence.sh --runner-profile testdata/c12/authority/runner-profile.v1.json --provider-config /run/secrets/talenro-c12-authority/provider.json --evidence-output /var/lib/talenro-c12-authority/authority-fence-scope.json
```

Expected: PASS only when every provider record for this run is terminal committed/aborted, PITR refusal succeeds, provider retention attestation is valid, cleanup succeeds and output is unexpired production `authority_fence_pitr` evidence.

- [ ] **Step 10: REFACTOR — verify exact cleanup without deleting append-only evidence**

Run: `/usr/bin/bash scripts/verify-c12-authority-fence.sh --verify-cleanup /var/lib/talenro-c12-authority/authority-fence-scope.json`

Expected: PASS after exact PITR DB/cert/credential/state cleanup; provider records remain retained as non-secret terminal digests, and no non-run operation was inspected or changed.

- [ ] **Step 11: Commit the authority/PITR gate**

```bash
git add internal/c12evidence/authority.go internal/c12evidence/authority_test.go internal/nodecontrol/authority/production_conformance_test.go scripts/verify-c12-authority-fence.sh testdata/c12/authority/runner-profile.v1.json
git commit -m "test: add production authority fence PITR gate"
```

### Task 7: Four-scope completion manifest and license gate

**Files:**
- Create: `internal/c12evidence/completion.go`
- Create: `internal/c12evidence/completion_test.go`
- Modify: `cmd/talenro-artifact-scan/main.go`
- Modify: `cmd/talenro-artifact-scan/main_test.go`
- Test: `internal/c12evidence/completion_test.go`

**Interfaces:**
- Consumes: exactly four `C12ScopeEvidenceV1` documents, nested `PlatformEvidenceV1` and `AuthorityFenceEvidenceV1`, scanner receipt, Xray/sing-box/outer/verifier license records and current trusted time.
- Produces: `C12CompletionManifestV1`, `ValidateCompletionInputs`, and CLI `verify-completion`.

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

func TestCompletionRejectsCrossBuildSplice(t *testing.T) {
	t.Parallel()
	inputs := validCompletionInputs()
	inputs.Scopes[2].TrackedTreeDigest[0] ^= 1
	if _, err := ValidateCompletionInputs(inputs); !errors.Is(err, ErrBuildMismatch) {
		t.Fatal("completion accepted cross-build evidence")
	}
}
```

- [ ] **Step 2: Run completion tests and verify RED**

Run: `go test ./internal/c12evidence -run '^TestCompletion' -count=1`

Expected: FAIL because completion types and validator are undefined.

- [ ] **Step 3: GREEN — implement fixed-order completion references**

```go
type ScopeReferenceV1 struct {
	Scope         Scope
	EvidenceDigest contracts.Digest
}

type C12CompletionManifestV1 struct {
	SchemaVersion          string
	CreatedAt              time.Time
	ExpiresAt              time.Time
	Scopes                 [4]ScopeReferenceV1
	RepoCommit             string
	TrackedTreeDigest      contracts.Digest
	SpecDigest             contracts.Digest
	ToolchainDigest        contracts.Digest
	ReleaseDigests         [3]contracts.Digest
	ImageSetDigest         contracts.Digest
	ArtifactScanDigest     contracts.Digest
	LicenseRecordSetDigest contracts.Digest
	DocumentationDigest    contracts.Digest
}
```

The fixed scope order is PowerShell, Git Bash, Linux platform/operator trust, authority fence/PITR. Revalidate every nested signature/attestation and expiry at creation time; require same repo/tree/spec/toolchain/releases/images, successful cleanup, production provider class for last two scopes, and `ExpiresAt` no later than the earliest scope expiry.

- [ ] **Step 4: GREEN — implement non-bypassable scanner/license/document gates**

Require scanner version/ruleset and clean closed-set receipt; distinct Xray MPL-2.0, sing-box GPLv3+, context-init, inner-daemon, verifier and verifier-test-assets records; Redis production status `blocked_pending_written_approval_or_commercial_license`; threat model and all three runbooks must pass their contract digest. Any missing/changed/license-bypassed record rejects completion.

- [ ] **Step 5: Run completion tests and verify GREEN**

Run: `go test ./internal/c12evidence -run '^TestCompletion' -count=1`

Expected: PASS for one same-build four-scope fixture; duplicate/missing scope, cross-build, expired evidence, fake production scope, cleanup failure, stale scanner, missing core license and Redis bypass all fail.

- [ ] **Step 6: GREEN — expose `verify-completion` CLI**

The command accepts four repeated `--scope-evidence` paths plus explicit scanner/license/document paths and `--output`; it emits one canonical manifest only after full validation. Reject stdin, glob, directory, duplicate path, symlink/reparse file, unknown flag and output inside the repo.

- [ ] **Step 7: Run CLI tests and verify GREEN**

Run: `go test ./cmd/talenro-artifact-scan ./internal/c12evidence -run '^TestCompletion|^TestVerifyCompletionCLI' -count=1`

Expected: PASS with sanitized finite errors and no partial output on failure.

- [ ] **Step 8: REFACTOR — fuzz completion decoding and run static gates**

Run: `go test ./internal/c12evidence -run '^$' -fuzz '^FuzzCompletionManifest$' -fuzztime=10s -timeout 30s`

Run: `go vet ./internal/c12evidence ./cmd/talenro-artifact-scan`

Expected: PASS.

- [ ] **Step 9: Commit completion validation before claiming completion**

```bash
git add internal/c12evidence/completion.go internal/c12evidence/completion_test.go cmd/talenro-artifact-scan/main.go cmd/talenro-artifact-scan/main_test.go
git commit -m "feat: validate four-scope C1.2 completion"
```

### Task 8: Run final four-scope acceptance and advance the roadmap

**Files:**
- Create: `docs/security/c12-completion-record.md`
- Modify: `docs/roadmap/implementation-sequence.md`
- Test: `docs/security/c12-completion-record.md`
- Test: `docs/roadmap/implementation-sequence.md`

**Interfaces:**
- Consumes: four fresh same-build evidence files, final scanner receipt, license/document digests and `verify-completion` from Task 7.
- Produces: one external `C12CompletionManifestV1`, a digest-only completion record, C1.2=`complete`, and C1.3=`current design`.

- [ ] **Step 1: RED — prove completion fails with only B10 Windows scopes**

Run:

```text
go run ./cmd/talenro-artifact-scan verify-completion --scope-evidence C:/c12-evidence/powershell.json --scope-evidence D:/c12-evidence/git-bash.json --scanner-receipt C:/c12-evidence/scanner.json --license-root docs/licenses --documentation-root docs --output C:/c12-evidence/completion.json
```

Expected: FAIL with finite category `scope_set_incomplete`; no output manifest exists.

- [ ] **Step 2: Run both Windows wrapper scopes from the final tracked tree**

Run PowerShell from the repository root and Git Bash from a repository-external temp directory using the exact commands in plan 08 Tasks 6–7.

Expected: two fresh, attested, successful `container_deterministic` scope files with distinct run IDs and the same daemon/repo/tree/spec/toolchain/release/image inputs.

- [ ] **Step 3: Run the final Linux platform/operator-trust scope**

Run Task 5 Phase A, perform the controlled cold reboot, run Phase B, then verify cleanup.

Expected: fresh production `linux_platform_operator_trust` evidence with all replay/time/host-policy/attacker/post-start/operator-guard cases and cleanup successful.

- [ ] **Step 4: Run the final authority-fence/PITR scope**

Run Task 6 production gate and cleanup verification.

Expected: fresh production `authority_fence_pitr` evidence, real pre-revoke PITR refusal, all provider records terminal and cleanup successful.

- [ ] **Step 5: GREEN — create the four-scope completion manifest**

Run:

```text
go run ./cmd/talenro-artifact-scan verify-completion --scope-evidence C:/c12-evidence/powershell.json --scope-evidence D:/c12-evidence/git-bash.json --scope-evidence /var/lib/talenro-c12-platform/linux-platform-scope.json --scope-evidence /var/lib/talenro-c12-authority/authority-fence-scope.json --scanner-receipt C:/c12-evidence/scanner.json --license-root docs/licenses --documentation-root docs --output C:/c12-evidence/completion.json
```

Expected: PASS; the manifest has exactly four ordered scope digests, equal build inputs, all evidence unexpired and no fake promoted to a production scope.

- [ ] **Step 6: GREEN — write the digest-only completion record**

`docs/security/c12-completion-record.md` records final repo commit/tracked-tree/spec/toolchain/release/image/scanner/license/document digests, four scope evidence digests, manifest digest, run completion timestamps/expiry, cleanup success and the exact constrained scanner claim. Do not copy attestation, provider credential, cert/key, core asset, config, absolute runner path or raw test output into Git.

- [ ] **Step 7: GREEN — update the roadmap only now**

Change C1.2 from `current design` to `complete`; change C1.3 to `current design`; preserve the C1.2 production target, Redis production block and Xray/sing-box license/release caveats. Do not claim C1.3 implementation.

- [ ] **Step 8: REFACTOR — run final clean-tree verification**

Run: `go test ./... -count=1 -timeout 10m`

Run: `go test -race ./... -count=1 -timeout 10m`

Run: `go vet ./...`

Run: `go tool golangci-lint run ./...`

Run: `git diff --check`

Run: `go run ./cmd/talenro-artifact-scan verify-completion --manifest C:/c12-evidence/completion.json --revalidate-all`

Expected: all PASS, current repository status equals the captured baseline except the two intended documentation files, and the manifest remains unexpired.

- [ ] **Step 9: Commit the completion record and roadmap transition**

```bash
git add docs/security/c12-completion-record.md docs/roadmap/implementation-sequence.md
git commit -m "docs: record C1.2 control plane completion"
```

## B11 and C1.2 final exit gate

B11 closes only after Tasks 1–7 are committed and Task 8 has produced a valid, fresh four-scope manifest from one build. A blocked provider, unavailable runner, expired evidence, failed cleanup, incomplete license record, Redis license bypass, local resource shortage or fake-only result leaves C1.2 at `current design`; none is a reason to weaken or relabel the gate.
