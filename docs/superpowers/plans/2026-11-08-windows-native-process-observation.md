# Native Process Observation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans in the existing worktree, task-by-task. Preserve the current sequential execution method.

**Goal:** Add an independent same-handle native receipt recorder, verified without native invocation.

**Architecture:** A PowerShell module validates and normalizes per-call records. A C# helper protects the borrowed handle and captures native return/error values immediately. Synthetic readers exercise the same normalization without loading the helper.

**Tech Stack:** Existing Windows PowerShell7 host, C# via Add-Type, built-in JSON/reflection/AST; no dependencies or installs.

**Spec:** docs/superpowers/specs/2026-11-08-windows-native-process-observation-design.md (approved after3bf6a2d5).

Status: prepared for user review; no implementation or tests yet.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; branch codex/c12-b01-task9-coordinator; PowerShell only.
- Existing process-observation.psm1 and its tests remain unchanged. No app source edits.
- Import and synthetic tests must not load the helper or call Windows APIs.
- No packages or installs. No launch/kill/attach/dump, polling, sleep, retries, infinite waits, file writes or settings changes inside the module.
- Only WaitForSingleObject(handle,0), GetExitCodeProcess and GetProcessTimes are allowed, once each in that order.
- Any actual native invocation, including against the current process or a child fixture, requires the later explicit runtime scope.
- Only compile/reflect the helper in a separate explicit check; no native method or live export execution. No app build, restore or full suite.
- Preserve Task11/12 and44/34/6/217 NOT_ACCEPTED; no new launch allowance, main merge or root-cause claim.

## Review Focus

1. UInt32 high-bit exit values and FILETIME halves survive validation/JSON without signed coercion (Task1).
2. Partial output followed by an exception is discarded; later readers still run (Task1).
3. Stale error values on successful calls never become current failures (Task1 and Task2).
4. Closed handles, failed reference acquisition and concurrent disposal never cause fallback opening or owned-handle closure (Task2 static review; runtime race unverified).
5. Import, synthetic tests and repeated helper loading cannot silently invoke native code or accept an unrelated preloaded helper type (Task2).

## File map and contracts

Create under clients/windows/tests:
- native-process-observation.psm1: shared normalization and two exports.
- native-process-observation.cs: namespace Talenro.Diagnostics.NativeObservationV1, static NativeProbe helper and typed DTOs.
- native-process-observation.tests.ps1: synthetic/static assertions only.
- native-process-observation.compile.tests.ps1: explicit helper compilation and signature reflection only.
Create docs/roadmap/2026-11-08-windows-native-process-observation-execution.md.
Update current-status.md and repository-recovery.md at handoff.

Exports:
- Get-TalenroNativeProcessObservation -ProcessHandle <Microsoft.Win32.SafeHandles.SafeProcessHandle> -CorrelationId <string>
- Get-TalenroSyntheticNativeProcessObservation -Readers <System.Collections.IDictionary> -CorrelationId <string>
Require nonempty correlation (maximum128 characters), no inferred PID/ownership.
Readers has exactly Wait, ExitCode, Times scriptblocks; reject other dictionaries
before any execution. Trusted in-memory readers only, no evaluated strings.

Snapshot: SchemaVersion=1, Mode='native-live'|'synthetic', CorrelationId,
ObservationStartedUtc/ObservationEndedUtc (UTC round-trip strings), SnapshotError
(null or exception type), Calls={Wait,ExitCode,Times}. No acceptance/verdict field.
Each call: Status='ok'|'api-error'|'exception'|'invalid'|'not-run', StartedUtc,
EndedUtc, ApiSucceeded (nullable Boolean), RawResult (nullable UInt32; wait only),
WaitState (signaled/timeout/failed/unexpected, otherwise null), ExitCode (nullable
UInt32), Times (null or Creation/Exit/Kernel/User each {High:UInt32,Low:UInt32}),
LastError (nullable Int32), ExceptionType (nullable string). Inapplicable fields null.

Synthetic reader payloads use only raw fields: Wait={RawResult,LastError};
ExitCode={ApiSucceeded,ExitCode,LastError}; Times={ApiSucceeded,Times,LastError}.
Exactly one scalar payload; require exact raw value types before normalization.
On success ignore stale LastError and normalize it to null. On failed BOOL calls
require Int32 LastError, ignore output buffers and normalize outputs to null.
For wait0/258 set ApiSucceeded=true, ok and signaled/timeout; 4294967295 means
false/api-error/failed with Int32 LastError; other UInt32 means null/invalid/
unexpected but retain RawResult. Exceptions/invalid payloads have null outputs;
exceptions preserve caught type only. No timestamp or numeric string coercion.

## Task1 — Synthetic receipt contract, test first

**Files:** module, synthetic tests, execution ledger from file map.
**Interfaces:** implement synthetic export and private normalization above; live
export is added in Task2. Output contract is shared by both exports.

- [ ] Record branch/HEAD, clean or explained baseline, approved spec and no runtime allowance in ledger. Read applicable test-first instructions.
- [ ] Write self-contained tests that import the missing module, print each assertion and exit nonzero on failure. Define fixture builders using explicit UInt32/Boolean types, never coerce inputs within production validation.
- [ ] Pin assertions: Wait0=>ok/signaled,258=>ok/timeout,UInt32.MaxValue=>api-error with error5,128=>invalid/unexpected retaining128. Exit0/7/259/2147483648/4294967295 retain exact UInt32 values. BOOL false with buffer0=>null output/error5; true with stale error5=>null LastError.
- [ ] Pin Times assertions: all-zero and all-UInt32.MaxValue halves retained without date conversion; failed call discards buffers; malformed/missing halves invalid. JSON round-trip retains numeric halves and nulls without loss (deserialized integer width need not stay UInt32).
- [ ] Pin each reader throwing, partial output then throw, null, wrong scalar types, multi-output and nonenumerated arrays; preserve later calls exactly once in Wait/ExitCode/Times order. Missing/extra/non-scriptblock dictionary executes nothing. Reject invalid correlation before reads. Assert mode and UTC intervals, no verdict, and helper type absent before/after import and synthetic execution.
- [ ] Run `& ./clients/windows/tests/native-process-observation.tests.ps1`; expect RED due to missing module/function, not test syntax. Record actual failure.
- [ ] Implement shared validation and synthetic export; independently catch each reader and timestamp each call. Null values never become success0. No native helper loading in this path.
- [ ] Rerun same command; require all assertions pass. Record exact count and limitations; inspect diff, commit explicit Task1 files as synthetic-only checkpoint.

## Task2 — Borrowed-handle helper and live adapter, compile only

**Files:** C# helper, module, synthetic/static tests, compile tests and ledger.
**Interfaces:** NativeProbe.Capture(SafeProcessHandle handle) returns typed snapshot
DTO with snapshot error, call timestamps and raw payloads consumed by Task1
normalization. Live export supplies correlation/mode; DTO cannot carry arbitrary
exception messages. No static constructor or initializer invokes native code.

- [ ] Write failing AST/source checks for the two exports and private loader, three allowed DllImports, zero wait, protected handle lifetime and no lifecycle operations. Add compile/reflection script expecting NativeProbe.Capture signature and DTO property types; never invoke Capture or P/Invoke methods.
- [ ] Run synthetic/static and compile scripts separately; record expected missing-helper/signature RED. Ensure compile test fails for missing source, not by launching a live probe.
- [ ] Implement private static P/Invokes with SetLastError=true, correct Win32 BOOL marshalling, UInt32 DWORD outputs and sequential two-UInt32 FILETIME struct. Capture Marshal.GetLastPInvokeError immediately on documented failure before timestamps/other operations. Keep successful LastError null.
- [ ] Implement Capture: reject null/closed/invalid handle; DangerousAddRef with success flag, recheck validity, obtain pointer only while held; try/finally releases reference only when acquired. Never Dispose/CloseHandle. Acquisition failure sets SnapshotError and three not-run calls; per-call exceptions retain type and continue. All native calls occur within retained reference lifetime.
- [ ] Add live export with Windows guard, typed handle/correlation inputs and private lazy loader using repository C# source. Cache loaded Type only in module scope; on foreign/preloaded same-name type without that cache, fail closed instead of adopting it. Repeat use of this module instance reuses its Type. Forced reimport into a host with preloaded type is an explicit error; use a fresh host later, never silently trust stale code. Import itself only declares functions/exports.
- [ ] Expand static checks for error-capture ordering, no constructors invoking APIs, balanced reference finally, no fallback OpenProcess/duplication/privilege operations and loader isolation. Manually inspect control flow: static text checks are not proof of concurrency safety.
- [ ] Run synthetic script then explicit `& ./clients/windows/tests/native-process-observation.compile.tests.ps1`. Compile helper with Add-Type in memory; reflect exact signatures, SetLastError/marshalling, FILETIME size8/offsets0,4 and DTO types. Do not instantiate handles or call native functions. Record counts separately; compilation does not prove ABI/runtime behavior. A fresh PowerShell7 host may be used only for these tests, not a process-observation fixture.
- [ ] Verify original recorder/test Git diffs are empty; rerun original `& ./clients/windows/tests/process-observation.tests.ps1` as synthetic regression. Commit explicit helper/module/test/ledger files, excluding binaries or raw receipts.

## Task3 — Review and portable handoff

**Files:** Task1/2 files read; execution/current-status/recovery docs updated.
**Interfaces:** consume test receipts and diff; produce saved checkpoint with live NOT_RUN.

- [ ] Apply the preserved inline execution workflow's final independent-review requirement; review only this component and relevant original contract. Reviewer must not run live export/native probe. Resolve findings with scoped tests, or record blockers without broader changes.
- [ ] Rerun synthetic, compile/reflection and old synthetic tests after fixes. Record actual results and review findings; no root-cause or app-shutdown claim.
- [ ] Update recovery with test commands, helper stale-type limitation, sequential/non-atomic semantics, caller identity responsibility and live NOT_RUN. Next is separately scoped same-handle integration, retaining30s/360s limits, not automatic launch.
- [ ] Run git diff --check, inspect explicit file list, commit/push existing branch, confirm remote SHA=HEAD and clean/explained status. No raw receipts, compiled helpers, secrets or main merge.

## Self-review and handoff

Task1 covers all receipt/type/serialization cases; Task2 covers handle and interop
structure via static review/compilation only; Task3 preserves review and recovery.
Real ABI correctness, concurrent disposal behavior and OS state remain unverified.
All review-focus points map to named steps; no real native tests are smuggled into
synthetic verification. Preserve sequential execution in this session. User review
of this implementation/testing scope is required before implementation begins.
