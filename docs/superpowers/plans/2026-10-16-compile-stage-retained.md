# Compile-stage Retained Diagnostics Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a fixed target-package-log-triggered retained diagnostic sample without changing ordinary acceptance or rerunning compilation.

**Architecture:** A bounded incremental parser proposes a trigger; a portable session gates activation using an exact-root check and owns one unchanged retained registry. An explicit native entry and separate compile-only controller preserve trigger provenance, independent cleanup evidence and durable diagnostic-only results.

**Tech Stack:** Windows PowerShell7/5.1, existing Add-Type C#/.NET/kernel32 adapter; no packages, SDK edits or installation.

**Spec:** [Approved written design](../specs/2026-10-16-compile-stage-retained-design.md).

**Status:** User approved the written spec after54dc50188c23a885fd5f1ba4a7d2ba90b17f0bcb.
User then approved this plan and selected Native/inline execution. Tasks1/2
are committed e0e971b2/b3c42aa2; Task3 and one review fix pass are complete.
Post-fix fake verification passed18 suites/698 checks per PS7/5.1 host
(10 static included per host), all36 commands exit0. Read the
[implementation record](../../roadmap/2026-10-16-compile-stage-retained-implementation.md)
for actual evidence/limits; no real execution is authorized by these steps.

## Global Constraints

- Use the existing isolated worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator and branch codex/c12-b01-task9-coordinator; inspect/reuse it with using-git-worktrees at execution, never reset unrelated work.
- Fixed target:talenro.local/platform/internal/nodecontrol/nodeleafbinding; fresh leaf-stage-retained-compile-NNN root in the existing diagnostic scope.
- 32 cumulative admission attempts, exact owned handles,600s native deadline,500ms natural/shared5s cleanup; unchanged0x101000 rights/inherit=false, no eviction/reopening/retry.
- Fixed limits:16MiB total bytes scanned,64KiB per runner iteration,64KiB per raw line including terminator, and at most60000 reader polls.
- Trigger policy leaf-test-compile-log-v1, schema1; error list at most16 entries, each at most512 characters. Paths compare as fixed Windows paths, package ordinal/case-sensitive.
- Both diagnostic flags true; unchanged shutdown schema3/policy. Acceptance/PackagePassed/OriginalTestPassed/RuntimeAttempted always false. Registry data never replaces missing independent core evidence.
- No Go/native Run/process/PID lookup/full SDK audit/Docker/download/security change/main merge during implementation tests. Fake sources/native boundaries and tiny portable fixtures only.
- Separate optional trigger metadata; old native entry signatures/default behavior unchanged. Ordinary guard additionally refuses any present CompileStageTrigger property, including null/malformed/cleared-flags receipts.
- A log match is observed command text, not OS identity, exact execution timing or guaranteed end-phase coverage. Cap/missing trigger/zero admissions remain incomplete.
- Preserve explicit necessary files only under AGENTS.md; no raw receipts/SDK/cache/binaries/secrets. Verify push SHA and clean worktree at handoff.

## Review Focus

1. Read-ahead beyond a matching line must not enter that line's prefix digest; final appended log data must not break a correct binding (Task1 PrefixStopsAtMatch; Task3 AppendedSuffixAllowed).
2. A candidate split at the exact byte/line limit may match at the bound; an unterminated or nonmatching record at the bound must fail without another read (Task1 ExactLimitCandidate).
3. Root can signal after parsing but before activation; no factory/open may occur after the exact-root check reports signaled (Task2 RootExitBetweenReadAndCheck).
4. Newly required C# sources can break old direct Add-Type consumers or leave source pin gaps (Task2 LoaderInventoryCompatibility; Task3 CompleteSourcePins).
5. Reader-disposal failure must not skip registry releases or replace an existing core failure (Task2 TeardownPreservesCoreFailure).

## File map and interfaces

All implementation files below are under docs/diagnostics/isolated-fuzz/.

Create compile-stage-trigger.cs (reader/parser/report), compile-stage-fixture.cs
(test-only fake bytes/clock/root probe), compile-stage-trigger.tests.ps1.
Create compile-stage-session.cs (portable activation/registry ownership) and
compile-stage-session.tests.ps1. Create compile-stage-receipt.tests.ps1 and
leaf-stage-retained-compile.ps1/.tests.ps1 (fixed controller and fake fixtures).

Modify process-native.cs only for explicit mode/hooks; run.ps1 for four-source
loading, optional conversion and refusal; preflight.ps1 for the two new production
C# files; retained-lifetime-integration.tests.ps1 for its direct loader. Update
the source binding lists in leaf-retained-compile.ps1 and
retained-child-observation.ps1 to include both new transitive native dependencies,
without changing their policy/identity/arguments. Do not change registry policy,
matrix dispatch, product Go files, SDK or historical receipts.

### Task 1: Bounded log trigger and provenance

**Files:** Create compile-stage-trigger.cs, compile-stage-fixture.cs and
compile-stage-trigger.tests.ps1. No native entry or process creation in this task.

**Interfaces produced:**

```csharp
public interface ICompileStageByteSource : IDisposable {
    long Length { get; }
    int Read(byte[] buffer, int offset, int count);
}
public sealed class CompileStageRootCheck : LifetimeApiEvidence {
    public uint WaitResult;
}
public sealed class CompileStageTrigger : IDisposable {
    public CompileStageTrigger(ICompileStageByteSource source,
        IRetainedLifetimeClock clock, string expectedCompiler, string sourceDirectory);
    public CompileStageTriggerReport Report { get; }
    public bool HasCandidate { get; }
    public void Observe(long runtimeDeadlineTicks);
    public bool ResolveCandidate(CompileStageRootCheck rootCheck, long runtimeDeadlineTicks);
    public void RootExited();
    public void Fail(string error);
}
```

CompileStageTriggerReport is a plain-field DTO:Schema/Policy/State; TotalByteLimit,
IterationByteLimit/LineByteLimit/PollLimit; BytesConsumed/PollCount; nullable
LineStart/LineEnd/LineNumber; RawLineBase64/LineSHA256/PrefixSHA256; nullable
CandidateObserved/Activation (LifetimeApiEvidence clock intervals), RootCheck
(CompileStageRootCheck), FirstRegistryQuery (LifetimeQueryEvidence); bounded Errors.
State strings exactly Waiting/Activated/Failed/RootExitedWithoutActivation.
LineEnd is exclusive including LF. Nullable fields stay null before observation.
Report fields hold no byte-source/clock/handles. Production FileCompileStageByteSource
constructor(string path) opens once with read access/FileShare.ReadWrite (not Delete),
implements this interface and closes once. Fixed package is internal, not an input.

- [x] **Step 1: Write named failing parser/provenance tests.** Add fake clock and
  appendable byte-source with read/dispose counters; no OS process adapters.
  Encode these assertions in executable tests:

```text
FragmentedLine: before LF HasCandidate=false; after LF true; Report.State=Waiting.
ExactGrammar: exact compiler/-p/package/-pack/source suffix including *_test.go matches;
  wrong compiler/package, duplicate -p/-pack, flag/value masquerading as source do not.
PrefixStopsAtMatch: line SHA/base64 and prefix SHA equal independently computed raw bytes
  through matching LF even when a read contains trailing lines; offsets/line number exact.
ExactLimitCandidate: LF-complete eligible record at byte/line bound may match;
  no-match/partial at bound -> Failed, no extra read or growing buffer.
```

Also assert CRLF, quote/backslash/spaced and UTF-8 paths, split multibyte sequences,
vet/link/non-command lines, malformed compiler candidate, short reads/0-byte EOF,
length regression/read exception,16MiB/64KiB/60000 limits, backwards clock/deadline,
error truncation16x512, duplicate observations after terminal state and dispose-once.
Unsupported escaping fails candidates rather than falling back to shell semantics.

- [x] **Step 2: Observe named RED.** Run
  `pwsh -NoProfile -File docs/diagnostics/isolated-fuzz/compile-stage-trigger.tests.ps1`.
  Expected:named missing CompileStageTrigger assertion, not a syntax/load accident.
- [x] **Step 3: Implement the interfaces/DTO/production reader.** Strict UTF-8,
  narrow quoted tokenizer, exact grammar from spec section4; no variable expansion.
  Keep bounded line and read buffers; incremental prefix digest ends at candidate
  LF even with read-ahead. Do not finalize/invent activation in Observe.
  ResolveCandidate accepts only a successful unsignaled258 root check, ordered
  clocks and remaining deadline; signaled0 becomes RootExitedWithoutActivation,
  failures become Failed. Capture candidate/root-check/activation intervals.
- [x] **Step 4: Observe GREEN on both hosts in fresh processes.** Run the command
  above, then `& 'C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe' -NoProfile -NonInteractive -File docs/diagnostics/isolated-fuzz/compile-stage-trigger.tests.ps1`.
  Require exit0/all named assertions PASS, no live calls; publish actual counts only.
- [x] **Step 5: Inspect the explicit file list before commit.** This task has
  exactly three new files listed above; stage those only and commit
  `diagnostics: add bounded compile-stage log trigger`. Correct any accidental
  source/test dependency or fixture coupling before moving to Task2.

### Task 2: One deferred registry, native mode and receipt isolation

**Files:** Create compile-stage-session.cs/.tests.ps1 and
compile-stage-receipt.tests.ps1; extend compile-stage-fixture.cs for fake root
checks. Modify process-native.cs/run.ps1/preflight.ps1,
retained-lifetime-integration.tests.ps1 and the two existing source-binding lists.

**Consumes:** Task1 trigger/interfaces/report; existing RetainedLifetimeRegistry,
IRetainedLifetimeClock and FakeLifetimeApi/FakeLifetimeClock in retained-lifetime-fixture.cs.
**Produces:**

```csharp
public interface ICompileStageRootProbe { CompileStageRootCheck Check(); }
public sealed class CompileStageRetainedSession : IDisposable {
    public CompileStageRetainedSession(CompileStageTrigger trigger,
        ICompileStageRootProbe rootProbe, Func<RetainedLifetimeRegistry> factory);
    public CompileStageTriggerReport TriggerReport { get; }
    public LifetimeReport LifetimeReport { get; }
    public void Tick(long runtimeDeadlineTicks);
    public void RootExited();
    public void Snapshot(string phase, long deadlineTicks,
        bool rootExited, bool timedOut, bool terminationRequested);
}
public static class CompileStageRetainedTeardown {
    public static void Run(Action coreCleanup, CompileStageRetainedSession session);
}
```

LifetimeReport is null until activation. Session owns trigger/reader and at most
one registry, never resets it. Tick orders Observe -> exact-root Check ->
ResolveCandidate -> factory once -> Poll; subsequent activated ticks only Poll.
Copy the registry's first query clock metadata into TriggerReport.FirstRegistryQuery.
RootProbe is a private native adapter around the exact CreateProcess handle,
WaitForSingleObject(handle,0) with immediate error/clock capture, not a PID lookup.
IsolatedFuzzProcess.Result gains nullable CompileStageTriggerReport CompileStageTrigger.
The new native signature is Result RunWithCompileStageRetainedDiagnostics(string
executable,string[] args,IDictionary overrides,string cwd,string stdout,string
stderr,int limitSeconds); production controller always supplies600.

- [x] **Step 1: Write named failing fake session/receipt tests.** Assert:

```text
WaitingCreatesNothing: factory/query/open counts all0 across unmatched/partial reads.
RootExitBetweenReadAndCheck: completed candidate + signaled root -> no registry/factory.
ActivateOnce: unsignaled candidate -> exactly1 factory, first query after activation;
  no more log reads or second cohort; attempt32 remains capped.
TeardownPreservesCoreFailure: core throws; reader disposal throws; all admitted handles
  still released once; primary core error preserved, secondary error recorded.
TriggerRoundtrip: populated report/nulls/base64/hashes/clocks survive actual converter/JSON;
  present trigger property (even null) is refused with false/missing diagnostic flags.
```

Add root signals immediately after activation, reader failure with continued root
supervision/cleanup, no late activation, cap snapshots/releases, zero samples,
native mode signature/selection static checks and LoaderInventoryCompatibility.
Use fake session/root/API for dynamic ordering; do not call a real Run entry.

- [x] **Step 2: Run RED.** Execute each new session/receipt test file in a fresh
  pwsh process; require named missing-session/optional-report assertions.
- [x] **Step 3: Implement session and explicit native wiring.** Keep old RunCore
  branches eager/default as today. Only the new mode defers registry allocation,
  opens stderr reader after native stream creation and uses session Tick/Snapshot.
  Waiting at root exit must be terminal. A root-check native failure records its
  original error; do not convert it into unsignaled success or mask core failures.
  Teardown tries core cleanup, reader disposal and every registry release; collect
  secondary disposal errors without replacing a primary core exception.
- [x] **Step 4: Update all four-source loaders/inventories/pins and receipt guard.**
  Production Add-Type and old integration loader include process-native.cs,
  retained-lifetime.cs, compile-stage-trigger.cs, compile-stage-session.cs.
  Converter omits a missing or null native CompileStageTrigger (the new typed
  Result has a default-null field on legacy paths), but preserves every non-null
  supplied value, including false/malformed values for downstream refusal. Guard
  refuses any present property on an input receipt, including null; no property
  means legacy behavior. Add LegacyDefaultNullOmitted:convert an ordinary typed
  Result, assert no trigger property and unchanged ordinary guard behavior.
  Include
  both new C# dependencies in preflight and existing child/leaf source bindings;
  leave eager entry algorithms and registry source unchanged.
- [x] **Step 5: Run GREEN/new and adjacent fake regressions on both fresh hosts.**
  Use Task1's host commands for compile-stage-session.tests.ps1,
  compile-stage-receipt.tests.ps1, retained-lifetime.tests.ps1,
  retained-lifetime-integration.tests.ps1, retained-lifetime-receipt.tests.ps1,
  retained-child-observation.tests.ps1 and leaf-retained-compile.tests.ps1.
  Require exit0/all assertions PASS; label static checks separately.
- [x] **Step 6: Review explicit task diff and commit**
  `diagnostics: wire isolated compile-stage retained mode`. No real process test,
  historical receipt reclassification or product/SDK edits.

### Task 3: Fixed controller, final provenance checks and recovery

**Files:** Create leaf-stage-retained-compile.ps1/.tests.ps1 and
docs/roadmap/2026-10-16-compile-stage-retained-implementation.md; update diagnostic
README, current-status.md and repository-recovery.md with actual verification.

**Consumes:** Task2 native entry/report plus existing Get-LeafRetainedCompileEvidence,
SDK/product/compiler-token/binding validators and safe JSON writer.
**Produces:** PowerShell Get-LeafStageRetainedCompilePlan(string RunRoot,string
Toolchain,string ReferenceReceipt,string ModuleCache) -> plan; Assert-LeafStageRetainedCompileSources(plan);
Get-LeafStageRetainedCompileEvidence(receipt,plan,string StderrPath) -> independent
TriggerEvidenceComplete/TriggerError/CoreNaturalComplete/CoreError/
LifetimeEvidenceComplete/LifetimeError/ObservationPassed with Acceptance=false;
Invoke-LeafStageRetainedCompileNative(plan,string Stdout,string Stderr) -> native
Result (only external boundary); Invoke-LeafStageRetainedCompile with planner's
four inputs -> finalized diagnostic state or stopped throw.

- [x] **Step 1: Write named controller RED tests with tiny portable SDK/module
  fixtures and fake native/SDK-audit/product inventory boundaries.** Assert:

```text
FixedPlan: new owned leaf-stage-retained-compile-NNN,600s, fixed package/offline args,
  no overlap/reparse/links/replay; imports/plans launch0; runtime/acceptance false.
CompleteSourcePins: full transitive production helper/parser/session/converter/collector
  file list and actual go hash present; source change before/after stops.
AppendedSuffixAllowed: exact saved prefix/line binding succeeds with later log suffix;
  changed consumed prefix, offsets, base64, line/prefix SHA or clocks stops.
OneFakeCompileOnly: one boundary invocation, no binary runtime, all independent audits;
  completed marker only after matching manifest; result alone provisional.
```

Require complete trigger/root-check/activation/first-query chronology within
receipt and matching registry query metadata; missing/null/forged/error reports,
zero samples/attempt32/core missing rows/nonzero/timeout/87/forced cleanup/release
failure all stop. Reuse ordinary validator without clearing flags or mutating
the receipt. Test product/SDK/source/artifact drift, primary plus multiple audit
errors, result/manifest/marker persistence failures, consumed identity/no retry.

- [x] **Step 2: Run RED in fresh pwsh.** Command:
  `pwsh -NoProfile -File docs/diagnostics/isolated-fuzz/leaf-stage-retained-compile.tests.ps1`;
  require named missing stage-controller assertion.
- [x] **Step 3: Implement the fixed independent controller.** Keep eager leaf
  controller's policy/identity unchanged. Bind trigger compiler/package/source
  directory, fixed limits and full transitive source inventory in the new plan.
  Verify prefix/line from a bounded streaming read of at most16MiB, not a second
  unbounded parser or another matching line. Join trigger and registry by clocked
  first-query evidence, never by an inferred process identity from text.
  Compose independent evidence results; source/core/lifetime failures remain
  distinct. Persist all original fixed-controller audits/manifest semantics and
  raw optional trigger metadata, never cache/tmp inventories or repeated native calls.
- [x] **Step 4: Run GREEN and final fake-only batch on both hosts.** One fresh
  process per suite; use Task1's two host command forms for these18 files:
  compile-stage-trigger, compile-stage-session, compile-stage-receipt,
  leaf-stage-retained-compile, matrix-check, pair-sample, envelope-uninstrumented,
  envelope-sample, envelope-compile, envelope-preparation, envelope-observe,
  mapping-single-check, retained-lifetime, retained-lifetime-integration,
  retained-lifetime-receipt, retained-child-observation, leaf-retained-compile,
  short-child-observation (all names suffixed .tests.ps1). Run only their existing
  fake/default modes; inspect any changed mode before running. Require all commands
  exit0; record exact host/suite/assertion/static counts and failures, not estimated
  or historical counts. Verify parsing/whitespace/local document links separately.
- [x] **Step 5: Request independent whole-change review.** Follow requesting-code-review;
  review against spec, interfaces, limits, default isolation, error precedence,
  fake-only coverage and source provenance. Address findings with focused RED/GREEN
  and scoped regression reruns; disclose review/test exclusions and unresolved gaps.
- [x] **Step 6: Record actual evidence/limits and preserve checkpoint.** Save raw
  logs/source bindings only in a new ignored evidence identity; commit necessary
  production/test files and portable report/current-status/recovery updates.
  Check explicit staged file list; commit `diagnostics: prepare compile-stage retained controller`,
  push existing branch, verify remote SHA/clean worktree. Do not claim native
  capture, error87 fix, acceptance or main readiness from fake tests.

## Self-review and handoff

Spec sections1/2/8 map to global constraints/recovery; sections3/5 to Tasks1/2;
section4 to Task1 grammar/provenance; section6 to Tasks2/3; section7 to each RED/
GREEN batch and final review. Review Focus cases each have a named owning test.
Portable DTOs/interfaces above are shared contracts, not optional name suggestions.
If a requirement cannot be met without changing the approved design, stop and
seek a design revision rather than quietly relaxing it.

Recommended execution:Native/inline, because three tasks depend closely on the
same parser/report/session interfaces; one independent whole-change review after
implementation preserves a separate check without per-task fresh contexts.
Alternative:subagent-driven implementation with fresh task reviews. User must
review this plan and select the method before either starts. Later real execution
remains a separate scope/approval even after every fake test passes.
