# C12 Secure Execution Architecture Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace C12's prepared-worker and cleanup paths with a fail-closed Windows controller that proves process containment, protocol authenticity, immutable inputs, and exact resource ownership.

**Architecture:** PowerShell remains the policy and evidence orchestrator while one embedded C# block supplies auditable Win32 primitives for suspended processes, Job Objects, named pipes, exact handles, and write-through files. Monotonic PowerShell state records connect those primitives to Go graph receipts, artifact/PITR ledgers, Docker custody, and retryable top-level cleanup.

**Tech Stack:** PowerShell 7.6.5, embedded C# via `Add-Type`, Win32 process/job/pipe/file APIs, Go 1.26.5 tests, Docker Engine 29.6.2, PostgreSQL 18.4.

**Spec:** `docs/superpowers/specs/2026-09-04-c12-secure-execution-architecture-design.md`

## Global Constraints

- Do not skip, delete, or weaken an existing C12 security assertion.
- The prepared-worker execution path must not call `Start-Job`.
- A native process must remain suspended until exact Job membership and the six-frame protocol are verified.
- Cleanup must use retained exact identities; it must never adopt or delete by name alone.
- External Docker volumes must never be mounted, relabelled, adopted, or deleted.
- Every child `go test` must inherit the approved overlay through `GOFLAGS`.
- Elevated Windows tests must explicitly set `GOOS=windows`, `GOARCH=amd64`, and `CGO_ENABLED=0`.
- Each task uses the existing failing acceptance test as RED, implements one cohesive boundary, and runs the broader affected package before commit.

---

### Task 1: Deterministic Test and Overlay Environment

**Files:**
- Modify: `internal/nodecontrol/authority/v7_migration_grant_test.go`
- Modify: `internal/store/nodecontrol_v7_catalog_test.go`
- Modify: `internal/testinfra/c12_integration_manifest_test.go`

**Interfaces:**
- Consumes: approved overlay at `.superpowers/sdd/task-8-corrective-implementation-plan/task-2-overlay-gate/overlay.json`.
- Produces: `task8ChildGoCommand(t, directory, arguments...) *exec.Cmd`, which carries the authenticated parent overlay and explicit Windows target into semantic probe subprocesses.

- [ ] **Step 1: Preserve the current RED evidence**

Run:

```powershell
$overlay=(Resolve-Path '.superpowers\sdd\task-8-corrective-implementation-plan\task-2-overlay-gate\overlay.json').Path
go test "-overlay=$overlay" ./internal/nodecontrol/authority ./internal/store -run 'TestNodeControlV7(SealedGrantTypes|OpaqueCrossPackageUse|ProviderScopedMigration)|TestStoredFencePersistedOutcomeProjection' -count=1
```

Expected: subprocess probes fail because `-overlay` is not inherited, and guarded-file authentication reports the obsolete repository hashes.

- [ ] **Step 2: Add a child-Go environment test**

Add a table-driven test that starts a temporary `go env GOOS GOARCH GOFLAGS` child through `task8ChildGoCommand` and asserts literal `windows`, `amd64`, and one absolute `-overlay=<approved path>` token. The production change that makes it fail is dropping the overlay or target when constructing a semantic-probe subprocess.

- [ ] **Step 3: Implement the shared child command**

Construct commands with an environment derived from `os.Environ()` after removing existing `GOOS`, `GOARCH`, `CGO_ENABLED`, and `GOFLAGS`, then append:

```go
"GOOS=windows",
"GOARCH=amd64",
"CGO_ENABLED=0",
"GOFLAGS=-overlay=" + approvedOverlayPath,
"GOPROXY=off",
"GOWORK=off",
```

Use the helper for opaque-boundary, repository-embedding, provider-scope, and integration-factory probes. Keep the overlay authentication check before spawning the child.

- [ ] **Step 4: Bind the current guarded files**

Retain the already observed literal SHA-256 values:

```text
internal/nodecontrol/authority/postgres_repository.go = 4F759A5B17648A75DF8930939282D51D63F1EC4D8563C254BFE8755321B9AFF3
.superpowers/.../postgres_repository.go = 8F3447CA7D6E32C41FF80057750A6172DB1C2207673BCADB8CDFD9752AE5B5DB
```

- [ ] **Step 5: Verify GREEN**

Run the command from Step 1 with the overlay supplied through `GOFLAGS`. Expected: all selected tests pass.

- [ ] **Step 6: Commit**

```powershell
git add internal/nodecontrol/authority/v7_migration_grant_test.go internal/store/nodecontrol_v7_catalog_test.go internal/testinfra/c12_integration_manifest_test.go
git commit -m "test(testinfra): propagate authenticated overlay to probes"
```

### Task 2: Suspended Process and Job Containment

**Files:**
- Modify: `scripts/run-c12-integration.ps1`
- Test: `internal/testinfra/c12_integration_manifest_test.go`

**Interfaces:**
- Produces: C# `C12SuspendedProcessController.Create(string executable, string arguments, string workingDirectory, IDictionary<string,string> environment)`.
- Produces: methods `AssignToJob()`, `VerifyMembership()`, `Resume()`, `Terminate(uint exitCode)`, `WaitForActiveProcessZero(int timeoutMilliseconds)`, and `Dispose()`.
- Produces: PowerShell `New-C12SuspendedPreparedWorker` returning a state record with `ProcessId`, `ProcessHandle`, `PrimaryThreadHandle`, `JobHandle`, `CompletionPortHandle`, and `Phase`.

- [ ] **Step 1: Run the existing RED containment test**

```powershell
go test ./internal/testinfra -run '^TestC12PreparedProcessIsSuspendedUntilVerifiedJobMembership$' -count=1
```

Expected: missing `CreateProcessW → AssignProcessToJobObject → IsProcessInJob → ResumeThread`, forbidden `Start-Job`, and missing retained handle state.

- [ ] **Step 2: Add behavioral failure tests**

Extend the existing sentinel harness to prove that a worker cannot create its marker before membership verification and that forced assignment failure terminates the suspended process. Assert the worker PID is absent and the Job reports zero active processes.

- [ ] **Step 3: Implement the embedded C# controller**

Add exact constants and P/Invoke declarations for `CREATE_SUSPENDED`, `CREATE_NO_WINDOW`, `CREATE_UNICODE_ENVIRONMENT`, `CreateProcessW`, `CreateJobObjectW`, `SetInformationJobObject`, `AssignProcessToJobObject`, `IsProcessInJob`, `QueryInformationJobObject`, `CreateIoCompletionPort`, `GetQueuedCompletionStatus`, `ResumeThread`, `TerminateProcess`, `TerminateJobObject`, `CloseHandle`, and `JOB_OBJECT_MSG_ACTIVE_PROCESS_ZERO`.

Set `STARTUPINFO.cb`, pass `bInheritHandles=false`, create the Job before resuming, bind its completion port, and set kill-on-job-close. Reject zero/invalid handles and any PID mismatch.

- [ ] **Step 4: Replace the prepared-worker `Start-Job` path**

Replace the call near the current prepared worker with `New-C12SuspendedPreparedWorker`. Transition only:

```text
CreatedSuspended → AssignedToJob → MembershipVerified
```

Do not call `Resume()` yet; Task 3 owns release.

- [ ] **Step 5: Implement bounded failure convergence**

On any pre-release failure, record `CleanIntent`, call `TerminateProcess` and `TerminateJobObject`, wait for `JOB_OBJECT_MSG_ACTIVE_PROCESS_ZERO`, and retain handles/state if confirmation times out.

- [ ] **Step 6: Verify GREEN and regression scope**

```powershell
go test ./internal/testinfra -run '^(TestC12PreparedProcessIsSuspendedUntilVerifiedJobMembership|TestC12BaseRunnerWatchdogKillsNativeDescendantsAndContinuesCleanup)$' -count=1
```

- [ ] **Step 7: Commit**

```powershell
git add scripts/run-c12-integration.ps1 internal/testinfra/c12_integration_manifest_test.go
git commit -m "feat(testinfra): contain prepared workers before execution"
```

### Task 3: Authenticated Six-Frame Private Pipe and Release Gate

**Files:**
- Modify: `scripts/run-c12-integration.ps1`
- Test: `internal/testinfra/c12_integration_manifest_test.go`

**Interfaces:**
- Extends C# controller with `CreatePrivatePipe`, `VerifyClientProcessId`, `VerifyServerProcessId`, `ReadMessage`, and `WriteMessage`.
- Produces: `Invoke-C12PreparedProtocol` returning the six verified frame receipts and `ProjectionDigest`.
- Produces: `Release-C12PreparedWorker`, the only function allowed to resume the thread and signal the execution gate.

- [ ] **Step 1: Run the existing RED protocol test**

```powershell
go test ./internal/testinfra -run '^TestC12PreparedPipeProtocolAndProjectionAreClosed$' -count=1
```

Expected: missing message-mode pipe primitives, missing ordered frames, forbidden worker receipt key, and incomplete protocol trace.

- [ ] **Step 2: Add malformed/replay behavioral cases**

For duplicate sequence, reordered frame, foreign PID, wrong nonce, oversized frame, incorrect previous digest, and bad HMAC, assert no gate signal, no worker marker, and exact process/Job convergence.

- [ ] **Step 3: Implement private pipe creation**

Create one instance with:

```text
PIPE_ACCESS_DUPLEX | FILE_FLAG_FIRST_PIPE_INSTANCE
PIPE_TYPE_MESSAGE | PIPE_READMODE_MESSAGE | PIPE_WAIT | PIPE_REJECT_REMOTE_CLIENTS
nMaxInstances = 1
```

Bind both endpoint PIDs using `GetNamedPipeClientProcessId` and `GetNamedPipeServerProcessId`.

- [ ] **Step 4: Implement canonical authenticated frames**

Frame fields are `schema`, `run`, `worker_nonce`, `sequence`, `type`, `payload`, `payload_digest`, `previous_frame_digest`, and `hmac_sha256`. Enforce UTF-8, canonical JSON, a 131072-byte frame maximum, monotonically increasing sequences, and constant-time HMAC comparison.

- [ ] **Step 5: Implement the ordered handshake**

Require exactly:

```text
HELLO → VERIFICATION_PROJECTION → JOB_MEMBER_READY → CAPABILITIES → CAPABILITIES_INSTALLED → GATE_WAITING
```

Keep `ReceiptKeyHex`, cleanup handles, and controller WAL keys out of the worker projection.

- [ ] **Step 6: Implement formal release**

After the sixth frame, start the release stopwatch, resume the primary thread, signal the gate, and append `FORMAL_RELEASE`, `EXEC_BEGIN`, and `EXEC_END` to the controller trace. Reject any execution marker timestamp preceding release.

- [ ] **Step 7: Verify GREEN**

```powershell
go test ./internal/testinfra -run '^(TestC12PreparedPipeProtocolAndProjectionAreClosed|TestC12PreparedArtifactExecutionAndTamperConverge)$' -count=1
```

- [ ] **Step 8: Commit**

```powershell
git add scripts/run-c12-integration.ps1 internal/testinfra/c12_integration_manifest_test.go
git commit -m "feat(testinfra): authenticate prepared worker release"
```

### Task 4: Purpose-Labelled Go Graph Receipts

**Files:**
- Modify: `scripts/run-c12-integration.ps1`
- Test: `internal/testinfra/c12_integration_manifest_test.go`

**Interfaces:**
- Produces: `New-C12GoGraphReceipt -Purpose $purpose -Packages $packages -Deadline $deadline`, where the arguments are respectively a non-empty string, a non-empty string array, and an absolute UTC `DateTime`.
- Produces receipt fields `Schema`, `Purpose`, `Toolchain`, `ModuleRoot`, `Environment`, `ModuleFiles`, `BuildArgv`, `Packages`, `Files`, and `CandidateTreeDigest`.

- [ ] **Step 1: Run the existing RED graph test**

```powershell
go test ./internal/testinfra -run '^TestC12PreparedGoGraphsBindEverySelectedInput$' -count=1
```

- [ ] **Step 2: Add graph mutation tests**

Mutate one selected source byte, one embed file, `go.mod`, `go.sum`, one module replacement, one package import, and one build argument. Each mutation must change the literal candidate-tree digest or fail validation before compilation.

- [ ] **Step 3: Implement bounded graph discovery**

Run `go list -deps -json -test -tags=integration` under the inherited deadline. Decode the stream with an 8192-byte read buffer, a 67108864-byte total limit, and a 16777216-byte individual JSON-object limit. Reject duplicate import paths and directory identities outside the module/cache roots.

- [ ] **Step 4: Bind exact package and file fields**

Record `GoFiles`, `CgoFiles`, `CFiles`, `CXXFiles`, `MFiles`, `HFiles`, `FFiles`, `SFiles`, `SwigFiles`, `SwigCXXFiles`, `SysoFiles`, all embed/test/external-test fields, imports, dependencies, module metadata, relative origin, length, and SHA-256.

- [ ] **Step 5: Bind toolchain and module policy**

Record exact Go executable/GOROOT/module-root/module-cache identities, `GOOS`, `GOARCH`, `CGO_ENABLED`, `go.mod`, `go.sum`, `GOSUMDB`, `-mod=readonly`, `go mod verify`, and the exact `cmd/test2json` invocation.

- [ ] **Step 6: Attach purpose-labelled receipts**

Set `Receipt.GoGraphReceipts` independently for validator and initializer artifacts. Validate the selected graph again immediately before executing each artifact.

- [ ] **Step 7: Verify GREEN**

```powershell
go test ./internal/testinfra -run '^(TestC12PreparedGoGraphsBindEverySelectedInput|TestC12Batch01SuiteRejectsTrackedAndUntrackedGroupMutationBeforeLaterCompilation)$' -count=1
```

- [ ] **Step 8: Commit**

```powershell
git add scripts/run-c12-integration.ps1 internal/testinfra/c12_integration_manifest_test.go
git commit -m "feat(testinfra): bind C12 Go dependency graphs"
```

### Task 5: Prepared Artifact Direct-Leaf Ledger

**Files:**
- Modify: `scripts/run-c12-integration.ps1`
- Test: `internal/testinfra/c12_integration_manifest_test.go`

**Interfaces:**
- Produces: `New-C12DirectLeafLedger`, `Register-C12DirectLeafIntent`, `Bind-C12DirectLeaf`, and `Converge-C12DirectLeafLedger`.
- Ledger entry fields: `Name`, `Kind`, `Expected`, `CreateAttempted`, `Identity`, `NumberOfLinks`, `RefCount`, `Lifecycle`, and `LastCleanupError`.

- [ ] **Step 1: Run the existing RED ledger test**

```powershell
go test ./internal/testinfra -run '^TestC12PreparedArtifactDirectLeafLedgerIsClosed$' -count=1
```

- [ ] **Step 2: Add unknown-sibling and link-splice tests**

Create an unknown direct child, hard-link splice, reparse leaf, and replaced directory identity. Assert cleanup refuses the affected root and retains the exact ledger entry without deleting the foreign object.

- [ ] **Step 3: Implement monotonic lifecycle transitions**

Allow only:

```text
NeverAttempted → CreateAttempted → Bound → CleanIntent → Removed → Absent
```

Support `exact_file` and `owned_ephemeral_subtree`. Capture file ID, volume serial, reparse status, link count, and directory identity immediately after creation.

- [ ] **Step 4: Replace bounded-directory cleanup for artifact roots**

Converge registered direct leaves individually, reject unknown siblings, and remove the root only after every leaf is confirmed absent. Never call `Remove-C12BoundedDirectory` on a prepared artifact root.

- [ ] **Step 5: Preserve retry state after deadlines**

Do not dispose the owning directory/handle before the final verification. Store `CleanIntent` and the last exact error so the top-level finalizer can retry using the same identity.

- [ ] **Step 6: Verify GREEN**

```powershell
go test ./internal/testinfra -run '^(TestC12PreparedArtifactDirectLeafLedgerIsClosed|TestC12PreparedArtifactCleanupRetainsOwnershipAfterDeadline|TestC12BaseRunnerCapturesNativeFailuresBeforeCleanup)$' -count=1
```

- [ ] **Step 7: Commit**

```powershell
git add scripts/run-c12-integration.ps1 internal/testinfra/c12_integration_manifest_test.go
git commit -m "feat(testinfra): close prepared artifact ownership ledger"
```

### Task 6: PITR Five-Leaf Ledger and Controller Ownership WAL

**Files:**
- Modify: `scripts/run-c12-integration.ps1`
- Modify: `internal/testinfra/c12_authority_pitr_integration.go`
- Test: `internal/testinfra/c12_integration_manifest_test.go`

**Interfaces:**
- Produces: `New-C12PITRRunLedger` with exactly five named leaves.
- Produces: `Open-C12ControllerOwnershipWAL`, `Append-C12ControllerOwnershipRecord`, `Read-C12ControllerOwnershipWAL`, and `Verify-C12ControllerOwnershipWAL`.
- WAL record fields: `schema`, `version`, `run`, `profile`, `nonce_digest`, `docker_executable_digest`, `docker_endpoint_identity_digest`, `sequence`, `previous_record_digest`, `event`, `timestamp_utc`, `payload`, `payload_digest`, `record_digest`, and `hmac_sha256`.

- [ ] **Step 1: Run both existing RED tests**

```powershell
go test ./internal/testinfra -run '^(TestC12PITRRunRootHasExactFiveLeafLedger|TestC12ControllerOwnershipWALUsesExactBytes)$' -count=1
```

- [ ] **Step 2: Add WAL tamper tests**

Cover truncation, duplicate sequence, reordered record, wrong previous digest, payload alteration, record alteration, wrong HMAC, noncanonical timestamp, and unknown event. Each must prevent destructive cleanup.

- [ ] **Step 3: Implement the exact five-leaf registry**

Register only `controller-ownership-v1.wal`, `ownership.wal`, `tlsgen.go`, `server.crt`, and `server.key` before creation. Reuse Task 5 lifecycle and exact identity semantics.

- [ ] **Step 4: Implement write-through WAL storage**

Open with write-through semantics, append one UTF-8 canonical JSON record plus LF, call `FlushFileBuffers`, and then update the in-memory verified head. Format UTC timestamps as `yyyy-MM-ddTHH:mm:ss.fffffffZ`.

- [ ] **Step 5: Implement the hash/HMAC chain**

Use canonical JSON to derive `payload_digest`, then `record_digest` over all non-HMAC fields, then `hmac_sha256`. Verify from `BOOTSTRAP` through `INTENT`, `ACTUAL`/`NOT_FOUND`, `CLEAN_INTENT`, and `CLEAN_RESULT` without trusting partial in-memory state.

- [ ] **Step 6: Correct pre-acquisition cleanup boundaries**

Before the PITR run root is acquired, skip PITR candidates entirely. After acquisition, cleanup reads only the authenticated registry and never accesses absent properties such as `WALPath` on pre-acquisition state.

- [ ] **Step 7: Verify GREEN**

```powershell
go test ./internal/testinfra -run '^(TestC12PITRRunRootHasExactFiveLeafLedger|TestC12ControllerOwnershipWALUsesExactBytes|TestC12PITRGroupCleanupSkipsCandidatesBeforeRunRootAcquisition)$' -count=1
```

- [ ] **Step 8: Commit**

```powershell
git add scripts/run-c12-integration.ps1 internal/testinfra/c12_authority_pitr_integration.go internal/testinfra/c12_integration_manifest_test.go
git commit -m "feat(testinfra): authenticate PITR ownership recovery"
```

### Task 7: Docker Endpoint Receipt and Retryable Exact Cleanup

**Files:**
- Modify: `scripts/run-c12-integration.ps1`
- Test: `internal/testinfra/c12_integration_manifest_test.go`

**Interfaces:**
- Produces: `Get-C12DockerEndpointReceipt`, `Register-C12DockerIntent`, `Confirm-C12DockerActual`, `Inspect-C12ExactDockerObject`, and `Converge-C12DockerRegistry`.
- Resource states: `NeverAttempted`, `CreateAttempted`, `Created`, `Verified`, `CleanIntent`, `Removed`, and `Absent`.

- [ ] **Step 1: Run existing RED Docker/cleanup tests**

```powershell
go test ./internal/testinfra -run '^(TestC12PITRDockerReceiptAndAbsenceAreExact|TestC12CleanupRetryStateRetainsExactOwnership|TestC12PITRBaseCleanupSkipsNeverAttemptedVolumes|TestC12PITRBaseCleanupReinspectsAttemptedUnconfirmedVolume)$' -count=1
```

- [ ] **Step 2: Add exact-identity behavioral cases**

Cover same-name foreign object, wrong role, wrong image reference, wrong image ID, wrong run nonce, create timeout with object actually created, and `No such container`/`no such volume`. Assert foreign resources survive and owned resources converge or retain retry state.

- [ ] **Step 3: Freeze the Docker endpoint receipt**

Bind context name, endpoint, engine ID, server version, OS type, architecture, `DOCKER_HOST`, `DOCKER_CONTEXT`, `DOCKER_CONFIG`, `DOCKER_TLS_VERIFY`, `DOCKER_CERT_PATH`, and `DOCKER_API_VERSION`, plus exact CLI bytes and SHA-256.

- [ ] **Step 4: Write intent and actual WAL transitions**

Set `CreateAttempted` before invoking Docker. On success, record returned object ID and full labels/image identity as `ACTUAL`. On uncertain failure, re-inspect the exact name and adopt only if every frozen identity field matches the intent.

- [ ] **Step 5: Implement exact absence and retained cleanup**

Use structured `container inspect` and `volume inspect` argument arrays. Treat only exact not-found responses as absence. On stop/remove timeout or identity mismatch, retain the registry, root handle, process handle, WAL handle, and ledger for the top-level finalizer.

- [ ] **Step 6: Share one dependency deadline**

Pass one absolute deadline through PostgreSQL, NATS, Redis, and Docker probes. Validate loopback-only mapped ports and reject malformed mappings before use.

- [ ] **Step 7: Verify GREEN**

```powershell
go test ./internal/testinfra -run '^(TestC12PITRDockerReceiptAndAbsenceAreExact|TestC12CleanupRetryStateRetainsExactOwnership|TestC12PITRBaseCleanupSkipsNeverAttemptedVolumes|TestC12PITRBaseCleanupReinspectsAttemptedUnconfirmedVolume|TestC12BaseRunnerBoundsDependencyProbesToOneDeadline|TestC12BaseRunnerRejectsMutableDockerIdentityDimensions)$' -count=1
```

- [ ] **Step 8: Commit**

```powershell
git add scripts/run-c12-integration.ps1 internal/testinfra/c12_integration_manifest_test.go
git commit -m "feat(testinfra): bind Docker lifecycle to exact custody"
```

### Task 8: Replace Stale Source-Shape Assertions with Stronger Behavior

**Files:**
- Modify: `internal/testinfra/c12_integration_manifest_test.go`
- Modify only if behavior is wrong: `scripts/run-c12-integration.ps1`

**Interfaces:**
- Consumes: Task 1–7 observable runner behavior.
- Produces: behavior-level tests for package authorization and Docker label identity, with no whitespace or occurrence-count dependency.

- [ ] **Step 1: Preserve the two current RED results**

```powershell
go test ./internal/testinfra -run '^(TestC12Task4AllowedPackagesStaySynchronizedWithRunner|TestC12BaseRunnerUsesPowerShellSafeDockerLabelTemplate)$' -count=1
```

- [ ] **Step 2: Replace whitespace matching with authorization behavior**

Invoke the real selector with each of the six literal package/import pairs and assert acceptance. Invoke it with wrong case, unknown package, and mismatched import path and assert rejection before any Go command runs.

- [ ] **Step 3: Replace template count with identity behavior**

Run the real Docker identity parser against literal inspect output containing labels with spaces/quotes. Assert exact run and role extraction for every call site, and assert rejection for missing or altered labels. Keep the explicit prohibition on the unsafe double-quoted template form.

- [ ] **Step 4: Verify GREEN**

```powershell
go test ./internal/testinfra -run '^(TestC12Task4AllowedPackagesStaySynchronizedWithRunner|TestC12BaseRunnerUsesPowerShellSafeDockerLabelTemplate)$' -count=1
```

- [ ] **Step 5: Commit**

```powershell
git add internal/testinfra/c12_integration_manifest_test.go scripts/run-c12-integration.ps1
git commit -m "test(testinfra): assert C12 behavior instead of formatting"
```

### Task 9: Full Verification, Node-Control Oracle, and Delivery

**Files:**
- Modify only for verified regressions: files implicated by failing tests.
- Create ignored evidence under: `.superpowers/sdd/task-8-corrective-implementation-plan/`

**Interfaces:**
- Consumes: all task outputs.
- Produces: a clean full-suite result, a fresh immutable authoritative oracle, reviewed Git index, final commit(s), and pushed branch.

- [ ] **Step 1: Run focused C12 package with correct platform and overlay**

```powershell
$env:GOOS='windows'; $env:GOARCH='amd64'; $env:CGO_ENABLED='0'
$overlay=(Resolve-Path '.superpowers\sdd\task-8-corrective-implementation-plan\task-2-overlay-gate\overlay.json').Path
$env:GOFLAGS="-overlay=$overlay"
go test ./internal/testinfra -count=1
```

Expected: PASS with no skipped security tests.

- [ ] **Step 2: Run the full repository suite**

```powershell
go test ./... -count=1
```

Expected: PASS for every package.

- [ ] **Step 3: Run static hygiene checks**

```powershell
gofmt -w internal/nodecontrol/authority/v7_migration_grant_test.go internal/store/nodecontrol_v7_catalog_test.go internal/testinfra/c12_integration_manifest_test.go
git diff --check
git diff --cached --name-only
```

Expected: no formatting errors and no unintentionally staged files.

- [ ] **Step 4: Build fresh immutable oracle inputs**

Clone the latest sealed build harness to a new never-before-used round, update only source closure hashes, isolated cache paths, artifact names, and exact expected product hashes, then execute it once. Record harness, lease, receipt, binary, test-list, and build-info hashes.

- [ ] **Step 5: Run a fresh one-shot authoritative Docker oracle**

Create a new never-before-used oracle bound to Step 4. It must report:

```text
PRE V7 reversibility 63/63
POST V7 authority 52/52
Legacy shapes 6/6
AUTHORITATIVE_ORACLE=PASS
```

Verify exact container/network absence and byte-identical normalized Docker custody. Verify the external volume count/bytes/digest are unchanged.

- [ ] **Step 6: Review the complete unstaged diff**

```powershell
git status --short
git diff --stat
git diff --name-only
git diff --check
```

Exclude ignored caches, binaries, leases, receipts, logs, and Docker snapshots. Include only reviewed source, generated files, tests, spec, and plan.

- [ ] **Step 7: Stage and inspect the final index**

```powershell
git add -- api/openapi/node-operator-api.v1.yaml db/migrations db/queries/nodecontrol_authority.sql db/schema/nodecontrol.v1.yaml docs/superpowers internal/nodecontrol internal/store internal/testinfra gen/go/talenro/nodeoperator/v1/server.gen.go go.mod scripts/run-c12-integration.ps1 sqlc.yaml
git diff --cached --stat
git diff --cached --check
git diff --cached --name-only
```

- [ ] **Step 8: Commit implementation**

```powershell
git commit -m "feat(testinfra): complete secure C12 execution controller"
```

- [ ] **Step 9: Push without force**

```powershell
git push -u origin codex/node-pop-control-plane
```

Expected: the remote branch advances by fast-forward. If rejected, stop and inspect remote history; do not force-push.
