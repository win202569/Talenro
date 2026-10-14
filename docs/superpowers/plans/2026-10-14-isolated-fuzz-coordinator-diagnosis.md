# Isolated Fuzz Coordinator Diagnosis Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Obtain bounded actual-coordinator diagnostic evidence without changing the official toolchain, products or acceptance gates.

**Architecture:** Copy the pinned toolchain into a validated independent diagnostic directory. Apply trace and targeted context-propagation hooks only to that copy, then execute individually recorded control arms through its real fuzz coordinator and worker. Preserve the difference between observation, forced scheduling and historical attribution.

**Tech Stack:** Windows PowerShell, existing Go1.26.5 windows/amd64, standard library only; no installs or downloads.

**Spec:** [Approved design](../specs/2026-10-14-isolated-fuzz-coordinator-diagnosis-design.md), approved after commit05e5a17c. Native execution selected. Baseline and new natural-shutdown model passed; five-arm preparation now stops on an exited member's missing image metadata. One independent unfinished-checkpoint review completed; source/binary binding and recovery fixes verified, shutdown-policy revision still pending. See the execution record.

## Global Constraints

- Use existing worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator and branch codex/c12-b01-task9-coordinator. Read AGENTS.md and spec before execution.
- Fixed source: C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64; never write it or overlay it.
- Owned run root: .superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/<unique-run-id>; fail if that root already exists. No sibling cache access.
- Source/target cannot coincide or contain one another. Reject reparse points and links. No hardlinks/junctions, global environment/config changes or automatic cleanup.
- Preserve -fuzztime=5s -parallel=1 -p=1 -timeout=30s and count=1. Each listed fuzz arm runs once, no retry-until-pass.
- Build process upper bound600s; each test process upper bound60s, internal test timeout30s. Record cold build separately by prebuilding the test executable.
- Child-process env only: GOTOOLCHAIN=local, GOPROXY=off, GOSUMDB=off, GOAUTH=off, GOVCS=all:off, GOOS=windows, GOARCH=amd64, GO111MODULE=off, GOFLAGS empty; explicit copy GOROOT/GOCACHE, GOENV=off. Do not change the parent/global environment.
- No product Go/SQL/module/test changes, product full matrix, Envelope/Pair acceptance, actual environment acceptance or main merge.
- Commit necessary scripts/templates/tests/reports only. Never commit toolchain tree, binaries, cache, receipts or secrets. Preserve failed/unfinished checkpoints honestly and verify remote SHA.

## Review Focus

1. Prefix-confusable and linked paths must be rejected before copying (Task1).
2. Copy go.exe silently using original GOROOT must fail provenance checks (Task1/3).
3. Unregistered contexts and early failures must not hang or be paused (Task2).
4. Short-circuit error reads must not become fabricated observed states (Task2/3).
5. Nonzero expected crashes must not hide absent traces, watchdog expiry or unfinished workers (Task3/4).

## File structure and interfaces

Tracked files, all under docs/diagnostics/isolated-fuzz/ unless noted:

- isolation.ps1: path validation, ordinary copying, provenance manifests.
- isolation.tests.ps1: standalone behavioral assertions; no Pester dependency.
- run.ps1: bounded processes, per-arm directories, recording and result validation.
- run.tests.ps1: runner tests with harmless local child-process fixtures.
- provenance.ps1 / provenance.tests.ps1: current-template/complete installed-source identity and binary binding; stale receipts refuse reuse, no automatic rebuild.
- preflight.ps1 / preflight.tests.ps1: readonly static prerequisite checks, not a full copy/origin substitute; test source path configurable for another PC.
- context-diag.go.txt / context-diag_test.go.txt: diagnostic-only context registry/gate and tests, materialized inside copied src/context.
- coordinator.patch: exact minimal patch for copied src/internal/fuzz/fuzz.go and src/context/context.go; source hash must match before applying.
- fixture_test.go.txt: independent standard-library fuzz fixture; materialized in run root.
- README.md: preparation and reproduction instructions, no machine-cache dependency.
- docs/roadmap/2026-10-14-isolated-fuzz-coordinator-execution.md: evidence and limits.
- docs/roadmap/current-status.md: current approval/results/blockers.

Materialized files never enter ordinary project packages. Existing portable model probe remains unchanged.

## Task1: Verified isolated copy

**Files:** isolation.ps1, isolation.tests.ps1, README.md.

**Interfaces:** `Assert-IsolatedPaths([string]$Source,[string]$Destination,[string]$AllowedRoot)` throws on unsafe paths; `Get-TreeManifest([string]$Root)` returns sorted relative Path/SHA256/Length objects; `New-IsolatedToolchain([string]$Source,[string]$RunRoot)` returns Source,Toolchain,Cache,BeforeManifest,CopyManifest. An existing run root is an error, not reuse authorization.

Test assertion shape (test utilities `Assert-Throws`/`Assert-Equal` belong in the standalone test file, not product code):

```powershell
Assert-Throws { Assert-IsolatedPaths $source $source $allowedRoot }
Assert-Throws { Assert-IsolatedPaths $source ($allowedRoot + '-outside/copy') $allowedRoot }
$before = Get-TreeManifest $source
$copy = New-IsolatedToolchain $source $newRunRoot
Assert-Equal ($before | ConvertTo-Json) ($copy.CopyManifest | ConvertTo-Json)
Assert-Equal ($before | ConvertTo-Json) ((Get-TreeManifest $source) | ConvertTo-Json)
```

- [x] Write failing path/copy tests using tiny ordinary fixture trees in a newly created owned test directory. Assert normal copy has equal relative manifests and independent bytes after editing only the target. Assert source=target, ancestor/descendant, prefix-confusable outside root, existing destination and reparse flags are rejected before write. On a real existing reparse point assert rejection; if real-link creation is unavailable, record that integration case unrun, not passing.
- [x] Run `powershell -NoProfile -File docs/diagnostics/isolated-fuzz/isolation.tests.ps1`; confirm behavioral RED (unsafe copy allowed or missing independent copy), not merely missing-file/parser error. Start with safe throwing stubs if needed to obtain meaningful assertion failures.
- [x] Implement validation and copying with native LiteralPath operations. Resolve existing ancestors, validate each component and every source entry; inspect HardLink count for source ordinary files and reject multiple links. Require destination nonexistent, then create it and copy ordinary bytes. Do not delete fixtures on failure.
- [x] Run the full standalone tests; expect exit0 with named passed assertions and explicit unrun integration entries. Include before/after source equality. Verify actual fixed source manifests only after execution authorization.
- [x] Update execution/status checkpoint and commit explicit tracked files; push and verify SHA. Do not copy the real toolchain yet if tests fail.

## Task2: Targeted propagation gate and truthful trace

**Files:** context-diag.go.txt, context-diag_test.go.txt, coordinator.patch; isolation.ps1 adds exact-hash patch materialization.

**Interfaces inside copied context only:** `DiagnosticFuzzRegister(ctx Context) *DiagnosticFuzzGate`, `(*DiagnosticFuzzGate).Entered() <-chan struct{}`, `.Propagated() <-chan struct{}`, `.Release()`, `.Unregister()`, `.Cancel()`. Only one registration is allowed; registration must find the exact parent cancelCtx, reject unsupported/already-canceled contexts, and never change their Err/Cause values while registering. Cancel explicitly invokes the registered real parent's existing cancellation path with Canceled (including timer cleanup), only in cancel mode; never synthesize Err. Ordinary context API is unchanged.

**Protocol:** register the actual timeout parent after fuzzCtx creation and before worker start. A matching cancelCtx hook immediately after Done closure signals Entered, then waits Release with a1s watchdog; after propagation and parent unlock it signals Propagated. Use sync.Once for release/signals and an atomic published gate pointer; the hook must not acquire another registry mutex while holding the parent lock. Unregister always releases then removes only its own registration. A watchdog event makes the diagnostic arm inconclusive/failing even if it later exits0.

**Coordinator modes:** environment selector `TALENRO_FUZZ_DIAG_MODE=observe|parent-first|synchronized|cancel`. Observe records only. Parent-first keeps gate closed through original classification, then releases before any stopping early return/cancelWorkers. Synchronized releases and waits for Propagated before entering the deadline stop branch. Cancel uses the same registered coordinator parent, explicit cancellation500ms after registration, and parent-first release discipline. Coordinator registration cleanup is deferred on every return.

**Trace:** JSONL prefix `TALENRO_FUZZ_DIAG ` to stderr, emitted by internal/fuzz only. Context hooks use channels/atomic state, not os/log imports. Record seq, mode, event, source (`deadline|worker|crasher|return`), stopping, incoming type/identity, childErr, parentErr plus parentRead boolean, old/new fuzzErr, gate/watchdog state. Capture the exact once-read values at existing expression evaluation points. Preserve short-circuit order, interrupt predicate, incoming errors and replacement rule; mark unread operands unread, not nil. Map known sentinels by equality, no text matching.

Core assertion in copied package context's `TestDiagnosticFuzzSelectedParent` (use a100ms deadline and register before creating child; keep cleanup deferred):

```go
<-gate.Entered()
if parent.Err() != DeadlineExceeded || child.Err() != nil {
    t.Fatal("selected parent window not established")
}
gate.Release()
<-gate.Propagated()
if child.Err() != DeadlineExceeded { t.Fatal("child did not inherit exact deadline") }
```

- [x] Write diagnostic tests asserting: unrelated canceled context never enters gate; selected parent Done closes while child Err is nil; release produces child's exact DeadlineExceeded and closes Propagated; duplicate release/unregister are harmless; watchdog releases a forgotten gate but marks expiry; early unregister cannot leave a blocked cancellation. Explicitly test the lock order by releasing before cancelChild and requiring completion under30s. Tests use real contexts, not fake Err implementations.
- [x] Materialize tests in the verified copy and run copied `go test context -run '^TestDiagnosticFuzz' -count=1 -p=1 -timeout=30s`; observe a behavioral RED against safe no-gate stubs. Build separate if cold compilation would consume external test guard.
- [x] Implement the protocol and exact source patch, rejecting source drift (fuzz.go SHA D9DE12569D5C18FB629C8E79A4D629EB2985B719E4FBFE30C3F3CE23FC44FEB5; context.go SHA765F99F7CC0710B0117FFA183E432524F429C21496FCBAD9B63FB9886FDBAAC7). Add copied-source trace tests for unread parent operand and exact same-text/wrapped/typed error identities; do not extract a new classifier that replaces actual stop.
- [x] Run all added copied context/trace tests, inspect complete output, then source hash and diff audits. Confirm no suppression/replacement semantic change and no import cycle. Failure to implement protocol faithfully stops implementation for design revision, not fabricated context fallback.
- [x] Save tests/patch and checkpoint with explicit observed versus unrun status; commit/push necessary files only.

## Task3: Runner, receipts and single-execution fixtures

**Files:** run.ps1, run.tests.ps1, fixture_test.go.txt, README.md.

**Interfaces:** `Invoke-RecordedProcess([string]$Executable,[string[]]$Arguments,[hashtable]$Environment,[string]$ReceiptDir,[int]$LimitSeconds)` returns ExitCode,TimedOut,DescendantsRemaining,StartUTC,EndUTC,StdoutPath,StderrPath. `Assert-ArmResult([string]$Arm,[object]$Receipt,[object[]]$Trace)` validates complete trace/expected outcome; `Invoke-DiagnosticArm([string]$RunRoot,[string]$Arm)` refuses existing receipts and invokes the arm exactly once.

Fixture `FuzzDiagnostic` has one []byte seed and a no-op callback; with child-process env `TALENRO_FUZZ_FIXTURE=crash`, callback deterministically panics with diagnostic marker `isolated-fixture-crash`. This is a deliberately injected warmup worker crasher, not a project material error. `-test.run=^$` excludes ordinary seed-test execution. Preserve worker callback entry/crash evidence and identify whether it traveled through result/crasher rather than errC.

Runner assertion rejecting exit-code-only evidence:

```powershell
$receipt = [pscustomobject]@{ ExitCode=1; TimedOut=$false; DescendantsRemaining=0 }
Assert-Throws { Assert-ArmResult -Arm 'parent-first' -Receipt $receipt -Trace @() }
```

- [x] Write runner tests: nonzero harmless child preserves exit and complete streams; timeout marked separately from exit; invalid GOROOT refused before execution; duplicate arm refused; expected nonzero with empty/incomplete trace refused; watchdog/remaining descendants rejected. Use literal assertions, e.g. `Assert-ArmResult parent-first` must reject a receipt with ExitCode1 but no stop/return trace.
- [x] Run `powershell -NoProfile -File docs/diagnostics/isolated-fuzz/run.tests.ps1`; confirm behavioral RED against minimal runner/validator stubs, not command syntax errors.
- [ ] Implement using hidden child processes, explicit working directory, redirected streams and process-specific environment. Track child descendants while root is alive; terminate only verified task-owned processes on timeout, record and report it, do not kill unrelated processes. Fail closed when ownership/cleanup cannot be verified.
- [ ] Build fixture test binary using copy `go test -c -o <arm>/fixture.exe <fixture_test.go>` once per source variant, limit600s. Record actual GOROOT/version, compile trace proving copied context/internal/fuzz source, fixture/patch hashes and env. Execute prebuilt binary with `-test.run=^$ -test.fuzz=^FuzzDiagnostic$ -test.fuzztime=5s -test.parallel=1 -test.timeout=30s -test.count=1`; the build retains `-p=1`. No build/test downloads.
- [ ] Run all runner tests and existing11 model controls (controls-only, no stochastic rerun needed). Check known error identity/field/unwrap cases retain their model-level label. Commit/push scripts/templates and docs checkpoint.

## Task4: Finite actual-framework arms

**Files:** execution report/status; receipts only in run root. Consumes Tasks1–3 APIs.

Arm order is fixed; use separate empty arm corpus/cache directories, no old input cache. Each arm once:

| Arm | Toolchain/mode | Required result |
| --- | --- | --- |
| baseline | Unmodified isolated copy | version/source match; normal no-op fuzz exit0 |
| observe | Patched copy/observe | full trace; natural state labelled timing-perturbed; exit0 or faithfully recorded unexpected failure |
| parent-first | Patched copy/parent-first | gate entered; actual incoming DeadlineExceeded/child nil; original classification preserves deadline; release before worker cancellation; final nonzero with exact deadline identity trace |
| synchronized | Patched copy/synchronized | propagation complete before stop; incoming/child deadline equal; no stored fuzz error; exit0 |
| crash | Patched copy/observe + crash fixture | actual worker/crasher evidence, final failure not swallowed; no claim of product material integration |
| cancel | Patched copy/cancel | selected parent early cancellation500ms; trace distinguishes cancellation from deadline; preserves original decision outcome, releases gate and workers finish |

- [x] Create verified real isolated copy using Task1; retain pristine manifest and baseline toolchain version/source evidence. Run baseline once before patch. If baseline cannot launch correctly, stop all later arms. Approved retry-001 fresh build and actual baseline passed; original failed preparation remains preserved. Later arms still stopped by Task3 model process-group gate.
- [x] Apply verified Task2 patch and run its gate/trace tests; record any failures before proceeding. Do not rerun completed baseline after patch. Context5 logic/completion PASS; trace3 logic PASS but process completion FAIL on ImageError31 of an owned exited member. STOP five later arms, no retry.
- [ ] Execute observe once, recording outcome without treating its possible deadline failure as product failure or passing acceptance.
- [ ] Execute parent-first and synchronized once each; reject false success if trace does not establish the prescribed state or watchdog expired. Mechanism reproduced means actual original stop/return path reached, not merely model assertion passed.
- [ ] Execute crash and cancel once each; require complete worker shutdown. Unexpected outcome stops that arm; safety/provenance/cleanup failure stops subsequent arms. A controlled nonzero expected result may count as a valid diagnostic control only after full trace checks.
- [ ] Recompute original toolchain and product manifests; require equality. Report all arms, including failed/unrun, with counts, timing, source identity, release and worker evidence. Commit/push report/status, no raw/private files.

## Task5: Recovery handoff and review

**Files:** README.md, execution report, current-status; update this plan's checkboxes from actual evidence only.

- [x] Verify committed scripts/templates independently materialize all required files without old receipts/cache; readonly preflight mode checks prerequisites without executing/copying. Record any real integration test still unrun. Materialization8 + preflight4 + provenance7 + real-link isolation18 assertions PASS; five actual arms remain UNRUN.
- [ ] Self-review spec coverage: isolation Task1; deadlock/trace Task2; bounded processes Task3; six finite arms Task4; recovery/limits Task5. Distinguish real-worker failure, model material fixtures, forced scheduling and natural observations throughout.
- [ ] Use chosen execution method's final review gate. Do not create reviewers until that method is selected. Fix scoped findings with tests; no repeated fuzz arms to hide a failure—rerunning an arm requires a new bounded authorization and fresh receipt identity.
- [ ] Confirm exact staged file list excludes binaries/toolchain/cache/receipts; verify source unchanged and documentation checks; commit/push and confirm remote SHA. If blocked, preserve/report local-only checkpoint without claiming remote recovery.
- [ ] Final handoff states reproduced/not-reproduced/failed/unrun per arm; original matrix3/Pair and all production gates remain open. No fix, toolchain promotion or main merge is authorized by completing this plan.

## Plan self-review and approval gate

Original plan self-review covered all seven spec sections and five Review Focus risks. Execution is partial: component tests do not close real integration risk. Baseline compilation exit0 was rejected by a remaining-process gate; actual fuzz arms were not launched and will not be retried without new bounded authorization. No product fix, acceptance clearance or main merge.
Latest checkpoint: approved new baseline did execute/PASS; historical failed preparation is immutable. Source-pinned patch/component runs completed, but trace-process shutdown failed despite three logic tests passing. Single independent review0C/2I/1M (recovery re-graded Important) produced two tested fixes and one unresolved approved-policy issue. No remaining actual arm ran or was retried; full plan completion is not claimed.
