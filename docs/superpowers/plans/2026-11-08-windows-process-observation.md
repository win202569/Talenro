# Lightweight Process Observation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans in the existing worktree. Implement sequentially; preserve the current inline workflow.

**Goal:** Build a typed, same-object observation recorder and verify it with synthetic readers only.

**Architecture:** One shared field reader validates results and isolates getter exceptions. Separate live/synthetic entrypoints supply readers to one snapshot assembler. Neither entrypoint controls process lifetime or decides acceptance.

**Tech Stack:** Existing PowerShell7, built-in JSON and AST parser; no new dependencies.

**Spec:** docs/superpowers/specs/2026-11-08-windows-process-observation-design.md (approved afterd1644653).

Status: proposed implementation plan; no implementation or tests run at preparation.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; branch codex/c12-b01-task9-coordinator. Preserve unrelated edits; PowerShell only.
- No app launch, process enumeration/termination/disposal, native interop, debugger, dump, settings change, build/restore or application-suite run. Live adapter remains unexecuted.
- No module filesystem writes or import-time process actions; no loops that poll, sleep, indefinite waits or retry. Trusted synthetic readers only.
- Keep error messages/paths/stacks/command lines out of records. Preserve all original evidence and NOT_ACCEPTED gates; no main merge.
- Record tests actually run; synthetic PASS means record logic only, not native behavior or root cause.

## Review Focus

1. Null/false/zero coercion: invalid values must never become successful exit0.
2. Field exceptions and partial output: discard partial results, preserve errors and later reads.
3. Disagreeing wait/HasExited: retain both, never manufacture an overall verdict.
4. Accidental live effects: test adapter declaration statically and never execute it.
5. Misleading identity/time/mode: retain field errors, UTC intervals and synthetic labels.

## Task1 — Implement recorder with synthetic test-first evidence

**Files:** Create clients/windows/tests/process-observation.psm1 and clients/windows/tests/process-observation.tests.ps1. Create docs/roadmap/2026-11-08-windows-process-observation-execution.md.

**Interfaces:** Export only `Get-TalenroProcessObservation -ProcessObject <System.Diagnostics.Process>` and `Get-TalenroSyntheticProcessObservation -Readers <System.Collections.IDictionary>`.
Synthetic Readers must contain exactly six scriptblock values named ProcessId, StartTime, WaitSignaled, HasExited, ExitCode, ExitTime; reject missing/extra/non-scriptblock inputs before invoking any reader.

Snapshot shape: SchemaVersion=1; Mode='live'|'synthetic'; ObservationStartedUtc and ObservationEndedUtc strings; Fields with those six named records. Field shape: Status='ok'|'error'|'invalid', Value, ExceptionType. Invalid has null Value/type; error has null Value and caught exception type name only. No extra Verdict/Accepted/Success property.
Read order: ProcessId, StartTime, WaitSignaled, HasExited, ExitCode, ExitTime. Each invoked exactly once, separately guarded. Live readers use explicit getters on the supplied object and WaitForExit(0); do not look up PID again. Private helpers must not be exported.

- [ ] Verify expected branch/clean baseline and initialize execution ledger with design approval, HEAD, no runtime allowance and test scope. Load applicable testing/verification instructions before implementing.
- [ ] Write self-contained assertion script using Import-Module and in-memory readers. It exits nonzero on assertion failure, prints each case and a final synthetic-only count. Tests must be repeatable without session-directory claims or external fixtures.
- [ ] Cover cases: exact exit codes0/7/259; Boolean false; both exit getters throwing while others succeed; wait=false/HasExited=true; each of six readers throwing in turn with full ordered call log; partial output then throw; zero output/null; wrong types including string0, Int64 exit code and PID0; empty/singleton/multiple arrays emitted without enumeration plus multiple pipeline outputs; invalid reader dictionary; failing identity fields; UTC timestamps and JSON round-trip preserving null/false/integer0/mode. Do not coerce input to declared field types before validation.
- [ ] Add AST checks for module import structure (only requires/definitions/exports), exactly the two exports, live adapter explicit getter names, WaitForExit argument0, and absence of process start/lookup/kill/dispose/native/file-write/Invoke-Expression operations. Manually review call paths too: an AST check alone is not a universal safety proof. The test script must never invoke the live export.
- [ ] Run `& ./clients/windows/tests/process-observation.tests.ps1` before adding the module. Expected RED from missing module/export, not a test syntax error. Save concise observed failure in ledger.
- [ ] Implement field reader and snapshot assembler, then both exports. Capture reader output without losing null/multiple-output distinctions; reject non-scalar types instead of taking the first item. Catch each reader exception, discard partial success output, continue. Convert valid DateTime to UTC round-trip strings. Timestamp interval covers all reads. Explicitly document sequential/non-atomic nature and caller ownership.
- [ ] Run the same synthetic command. Expected every assertion PASS and successful script completion; inspect full result. On failure diagnose only this component, record the failure, fix and rerun synthetic tests; do not broaden into live probes.
- [ ] Review implementation against all design bullets, record RED/GREEN evidence and limitations. Run git diff --check; commit explicit module/test/ledger files. Do not claim live adapter runtime verification from AST coverage.

## Task2 — Review and preserve portable handoff

**Files:** Read Task1 files; update execution ledger, docs/roadmap/current-status.md and docs/runbooks/repository-recovery.md. No app files.

**Interfaces:** Consume Task1 snapshot contract/test receipts; produce a recoverable development checkpoint with synthetic-vs-live limits explicit.

- [ ] Review diff for all five Review Focus points, module import effects and live adapter ownership. Use the inline execution workflow's final review requirement; no reviewer may run live adapter or apps. Any finding must be addressed within approved module/synthetic-test scope or reported as unresolved.
- [ ] Run `& ./clients/windows/tests/process-observation.tests.ps1` after any review fix. Expected all assertions pass; record actual final count, no invented budget or native acceptance.
- [ ] Update progress/recovery: module/tests available, exact test command/results, live adapter unexecuted, no app-root-cause conclusion. Next step is separately scoped launcher integration/runtime evidence, not immediate launch.
- [ ] Run git diff --check; stage only necessary explicit files, commit/push existing development branch, then confirm remote SHA equals HEAD and working tree status. Exclude raw receipts/caches/binaries.

## Self-review and handoff

Task1 exports are the only interface; Task2 consumes their tests without changing application behavior. All eight design test categories map to Task1 assertions; runtime API behavior is explicitly outside coverage. Existing PowerShell is sufficient; no package acquisition or new process is needed. Retain sequential execution by the current implementer. Please confirm this implementation/test scope before code changes.
