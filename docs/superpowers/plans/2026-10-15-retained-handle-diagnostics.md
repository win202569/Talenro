# Retained-handle Diagnostics Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add explicitly enabled, bounded own-job retained-handle diagnostics without upgrading missing evidence or changing default acceptance.

**Architecture:** A separately testable registry owns verified handles through one run. A small native adapter connects it to an opt-in runner entry, while original shutdown inspection remains independent. Optional diagnostic receipt fields are preserved but always refused as acceptance.

**Tech Stack:** Windows PowerShell7/5.1, existing Add-Type C#/.NET and kernel32 adapter; no packages or SDK changes.

**Spec:** [Approved written design](../specs/2026-10-15-retained-handle-diagnostics-design.md).

**Execution record:** User approved inline execution afterc4069cc6. Tasks1/2 are
committed bf0c8376/658571f5; Task3, the one review fix pass and final scoped checks
are complete pending ordinary checkpoint preservation. The original checklist
below is the approved work definition; actual RED/GREEN, coverage limits, deferred
Minor and verification sequence are in the
[implementation record](../../roadmap/2026-10-15-retained-lifetime-implementation.md).
No live experiment or acceptance is authorized by these completed unit steps.

## Global Constraints

- Working directory:D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; existing branch codex/c12-b01-task9-coordinator. Planning base8c81feb30be99ac09a0be0a17519b2ce6d4728e1.
- Explicit native RunWithRetainedHandleDiagnostics; no ambient environment enablement, matrix-driver enablement or changed meanings of old booleans.
- Fixed limits:32 admission attempts per run, at most32 retained child handles, one attempt per numeric candidate, no retry, eviction or reopening.
- Existing0x101000 rights, inherit=false, own-job verification; root uses its CreateProcess handle, not another admission.
- Diagnostic root wait slices up to10ms; existing root runtime,500ms natural window and shared5s shutdown clock are not enlarged. Calls are nonpreemptible; check deadlines between calls and report overruns.
- LifetimeDiagnosticOnly=true and ShutdownDiagnosticOnly=true; unchanged ShutdownSchema3/policy. Neither flags nor registry records grant acceptance; missing evidence stays unknown.
- Run-local bounded sampled records only; no static cache, thread, port, archived-PID query, raw handle serialization or identity join by PID alone.
- Cleanup always owns precedence; finally attempts all releases once, even after diagnostic errors. Core exceptions are not swallowed; partial diagnostics on native throw need not persist.
- No real-process creation/test/Go/matrix/SDK audit/download/security change/main merge in this implementation. Unit tests must use fake native adapters and portable fixtures.
- Ordinary necessary checkpoints may commit/push under AGENTS.md; explicit file review and remote-SHA confirmation required. Never stage SDK/cache/binaries/raw evidence/secrets.

## Review Focus

1. Diagnostic work at an expired deadline must not consume a fresh budget or mask root timeout (Task1 clock tests; Task2 wiring checks).
2. Enumeration changes or reused numeric IDs must not attach another object's state to an earlier EntryId (Task1 duplicate/ambiguity tests).
3. A native read/close exception during cleanup must not leak remaining release attempts or replace a core exception (Task1 disposal; Task2 outer-finally tests).
4. Hosts loading C# through run.ps1/preflight must include the new source, while old receipts still parse without lifetime fields (Task2 loader; Task3 roundtrip tests).
5. A fully successful-looking diagnostic receipt, including stripped legacy diagnostic flag, must still be rejected if LifetimeDiagnosticOnly is true (Task3 gate tests).

## File map and locked interfaces

Create under docs/diagnostics/isolated-fuzz:

- retained-lifetime.cs: portable DTOs, interfaces, owning wrapper and registry; no P/Invoke or process creation.
- retained-lifetime-fixture.cs: deterministic fake API/clock for tests only; fake handles never call native code.
- retained-lifetime.tests.ps1: admission, bounds, observations, budgets, disposal contracts.
- retained-lifetime-integration.tests.ps1: compile/reflect shared wiring, fake lifecycle ordering and adapter contract checks; never invokes Run.
- retained-lifetime-receipt.tests.ps1: actual converter/guard with synthetic receipts and JSON.

Modify process-native.cs (opt-in entry/adapter/hooks), run.ps1 (multi-source load,
optional serialization/refusal), preflight.ps1 (new required source inventory),
README/current-status and add a portable implementation-results report. Do not
restructure unrelated collector logic or modify matrix-check.ps1.

Public portable C# contracts (Task1 produces; Tasks2/3 consume):

```csharp
public interface IRetainedLifetimeClock {
    long Ticks { get; } long Frequency { get; } string UtcNow { get; }
}
public interface IRetainedLifetimeApi {
    LifetimeCandidates QueryMembers(IntPtr job, long deadlineTicks);
    LifetimeOpen Open(uint pid, uint access, bool inherit);
    LifetimeMembership IsOwned(IntPtr handle, IntPtr job);
    LifetimeRead Read(IntPtr handle, long deadlineTicks);
    LifetimeRelease Close(IntPtr handle);
}
public sealed class RetainedLifetimeRegistry : IDisposable {
    public RetainedLifetimeRegistry(IntPtr job, uint rootId,
        IRetainedLifetimeApi api, IRetainedLifetimeClock clock);
    public LifetimeReport Report { get; }
    public void Poll(long runtimeDeadlineTicks);
    public void Snapshot(string phase, long shutdownDeadlineTicks,
        bool rootExited, bool timedOut, bool terminationRequested);
    public void Dispose();
}
```

DTOs are plain public fields compatible with both Add-Type hosts. LifetimeCandidates
contains bounded query facts and returned ulong IDs; LifetimeOpen carries the
temporary handle and original error; LifetimeMembership carries success/owned/error;
LifetimeRead carries zero-wait state, exit/image availability/errors and call clocks;
LifetimeRelease carries success/error. All API results distinguish unattempted,
failed and successful. Native adapter must capture last-error before clock updates.
LifetimeReport fields:Schema=1, Coverage="sampled", Attempts, Admitted, QueryCount,
AdmissionBudgetExhausted, Incomplete, Errors, Entries, FirstQuery, LastQuery,
RootExited, TimedOut, TeardownAttempted, ReleaseAttempts, ReleaseFailures.
Entries store EntryId/Pid/source-query facts, scalar Open/Membership evidence and Admission/
RootWaitEnd/Final observations; report must not expose IntPtr/handle values.
LifetimeOpen is a transient adapter reply:copy only clocks/success/error into
report Open evidence, never the reply's handle or object reference. Error records
use bounded stage/code/message values, never serialize native exceptions or adapter
objects. Internal handle storage stays private, separate from report DTOs. Tick-deadline arithmetic
is checked against overflow; clocks use one injected monotonic frequency per run.

---

### Task 1: Portable registry and ownership contracts

**Files:** Create retained-lifetime.cs, retained-lifetime-fixture.cs and retained-lifetime.tests.ps1 in the directory above.

**Interfaces:** Consumes no collector types/native APIs; produces all portable
contracts above. Fake API supplies queued list/open/membership/read/close results,
call counters and order; fake clock advances only when the fixture says so.
PowerShell New-LifetimeFixture returns Api,Clock,Registry for fake job1/root1;
Check-Lifetime throws on false and emits one PASS line otherwise.

- [ ] Write failing named cases AdmissionOwnsExactHandle, Open87StopsBeforeMembership,
  MembershipFailureReleasesTemporary, NonmemberNeverAdmitted, DuplicateAndRootIgnored,
  NumericAmbiguityNeverMerged, Cap32No33rdOpen, DiagnosticErrorStopsAdmission,
  SameHandleSnapshots, ExpiredBudgetNoCalls, CallOverrunNoClockReset,
  ForcedExitNeverNatural, MissingImageOrExitUnknown, DisposeAllOnce and CloseFailureContinues.
  Queue fake IDs2–34 for cap; make the32nd open fail87 to prove failures count;
  separate first-error case verifies later Poll makes no new admission calls.
  Pin assertions using the actual report and fake counters, for example:

```powershell
Check-Lifetime ($f.Registry.Report.Attempts -eq 32 -and $f.Api.OpenCalls -eq 32) 'cap includes failed admission'
Check-Lifetime ($f.Api.MembershipCalls -eq 0 -and $f.Registry.Report.Entries[0].Open.Error -eq 87) 'open error remains original'
Check-Lifetime ($f.Api.ReopenedHandles -eq 0 -and $f.Api.PositiveWaitCalls -eq 0) 'same handle and zero wait only'
Check-Lifetime ($f.Api.ReleaseCalls -eq $expectedOwnedAndTemporary -and $f.Api.DoubleCloseCalls -eq 0) 'one release attempt per owned resource'
```

- [ ] Run `pwsh -NoProfile -File docs/diagnostics/isolated-fuzz/retained-lifetime.tests.ps1`.
  Expected:RED because registry contracts do not exist, not fixture syntax failure.
- [ ] Implement the locked contracts. Use an internal idempotent owned-handle
  wrapper with injected Close; transfer only after successful membership. Keep
  candidate IDs for deduplication separate from same-entry handle observations.
  Mark invalid/root/duplicate cases without creating identities; record at most32
  attempt rows. At cap stop admission queries. Unexpected adapter exceptions and
  API failures mark Incomplete and stop admission, never close a transferred handle
  prematurely. Snapshot phases are exactly root-wait-end and post-cleanup; each
  phase records at most one observation per owned entry. Repeated calls cannot
  overwrite earlier evidence. Dispose records failures and continues, never throws
  over a core exception or retries a failed raw close.
- [ ] Run the Task1 command on fresh PS7 and PS5.1 hosts. Expected:all named cases
  PASS/exit0, no native adapter/PInvoke/process/SDK path used. Add fake adapter throws
  during membership/read/close, stale clock/deadline overflow, repeated Dispose,
  and snapshot-after-dispose tests; expected explicit failure/skip, never reopened
  handles or fabricated successful observations.
- [ ] Review exactly the three new source/test files; commit `diagnostics: add bounded retained lifetime registry`.

### Task 2: Opt-in native adapter and lifecycle wiring

**Files:** Modify process-native.cs, retained-lifetime.cs, run.ps1 loader and preflight.ps1 inventory;
create retained-lifetime-integration.tests.ps1; reuse Task1 fake fixture.

**Interfaces:** Consumes Task1 registry/API/clock. Produces existing Run overloads
unchanged plus exact new public signature:

```csharp
public static Result RunWithRetainedHandleDiagnostics(string executable,
    string[] args, IDictionary overrides, string cwd, string stdout,
    string stderr, int limitSeconds);
```

Result adds bool LifetimeDiagnosticOnly and optional LifetimeReport LifetimeDiagnostics.
Private RunCore accepts existing two capture booleans plus bool retainLifetime;
old overloads pass false, new entry passes false,false,true. Reject mixed modes
before launch. Native adapter is nested RetainedLifetimeNativeApi; real clock is
nested RetainedLifetimeClock. No production global adapter replacement hook.
Also produce portable public static RetainedLifetimeTeardown.Run(Action coreCleanup,
RetainedLifetimeRegistry registry):try coreCleanup, finally dispose registry when
non-null. RunCore's actual outer finally calls this helper with its existing
resource-cleanup body. This is the shared seam tested for core-exception priority,
not a second fake-only approximation of cleanup.

- [ ] Write failing tests NewEntryExists, LegacyModeSelection, LoaderHasBothSources,
  InventoryIncludesRegistry, HookOrderAndBudgetIdentity and CoreFailureStillDisposes.
  Compile both source files in one Add-Type call, reflect signatures/fields, and
  call RetainedLifetimeTeardown.Run with fake cleanup actions and real portable
  registry/fake API to verify ordering/preserved core error. Inspect actual
  RunCore source wiring as a supplementary static check; do not label static/fake
  checks as end-to-end native validation. Expected initial RED:new entry absent.
- [ ] Run `pwsh -NoProfile -File docs/diagnostics/isolated-fuzz/retained-lifetime-integration.tests.ps1`.
  Expected:RED at missing integration contract, without calling Run/CreateProcess.
- [ ] Implement native adapter using current own-job list, OpenProcess0x101000,
  IsProcessInJob, WaitForSingleObject(handle,0), exit/image APIs and CloseHandle.
  QueryMembers uses one new diagnostic-only process-ID-list call with capacity32
  (buffer8+32*IntPtr.Size), no resize/retry; retain raw counts/error and reject
  failed/partial/oversized headers as incomplete without admitting partial IDs.
  Root may occupy a returned slot but is never admitted. Existing Members/
  MembersWithEvidence remain unchanged. Add fake incomplete-list/header tests;
  native parsing/wiring is statically reviewed, not certified by fake tests.
  Read checks the passed deadline before each API; optional image error stays
  metadata. Capture original native errors immediately. Only the adapter touches
  real handles; fake tests never instantiate/call it. Add both source paths to
  run.ps1's single Add-Type compilation and new source to preflight inventory.
- [ ] Wire registry only for retainLifetime. Poll while root waits with up-to10ms
  slices; baseline modes retain their original cadence/branches. Start the shared
  cleanup clock at the original point; root-wait-end snapshot precedes original
  queries/inspection and consumes the same clock. Post-cleanup snapshot consumes
  only remaining5s; final elapsed includes diagnostics. Preserve original native
  member evidence, counts, waits, errors and cleanup predicates without registry
  substitution. Set both diagnostic flags while keeping schema3/policy.
- [ ] Put registry.Dispose in an outer finally after the existing job-handle
  cleanup via the shared teardown helper, even if cleanup itself fails. Attach
  mutable diagnostic report before returning so final disposal outcomes survive
  successful native return. Do not persist partial native exception results or
  replace the original exception. Add fake fault tests at admission, root-wait-end,
  post-cleanup, core failure and disposal; assert core error identity unchanged,
  all release attempts made and forced post-cleanup exit labelled correctly.
- [ ] Run Task1 and Task2 test files in separate fresh PS7/5.1 hosts. Expected:all
  PASS/exit0; compile/reflection/static/fake lifecycle evidence only, no real launch.
  Review old Run call sites and diff to ensure matrix/default callers never select
  retained mode. Commit explicit Task2 files as `diagnostics: wire isolated opt-in lifetime capture`.

### Task 3: Receipt refusal, compatibility and scoped handoff

**Files:** Modify run.ps1 converter/guard; create retained-lifetime-receipt.tests.ps1;
update README/current-status; create docs/roadmap/2026-10-15-retained-lifetime-implementation.md.

**Interfaces:** Consumes Task2 Result fields. ConvertTo-RecordedProcessReceipt keeps
its existing signature and conditionally copies optional LifetimeDiagnosticOnly/
LifetimeDiagnostics fields. Assert-ProcessCompletion retains its existing signature
and rejects either true diagnostic flag before success/legacy checks. No new live
PowerShell invocation wrapper is added.

- [ ] Write failing cases LifetimeFlagAloneRejected, BothFlagsRejected,
  LegacyReceiptWithoutFieldsUnchanged, NestedRoundtripNoHandles,
  CoreOpen87NeverUpgraded and MissingSnapshotNeverBecomesSuccess using the real
  converter/guard with in-memory synthetic receipts. Expected assertions:

```powershell
Check-Lifetime ($saved.LifetimeDiagnostics.Schema -eq 1 -and $saved.LifetimeDiagnostics.Coverage -ceq 'sampled') 'nested diagnostics retained'
# Build a clean synthetic schema3 receipt, set LifetimeDiagnosticOnly=true,
# ShutdownDiagnosticOnly=false; Assert-ProcessCompletion MUST throw.
# Reuse the saved scalar87 fields in an in-memory fixture, never query its PID.
```

- [ ] Run `pwsh -NoProfile -File docs/diagnostics/isolated-fuzz/retained-lifetime-receipt.tests.ps1`.
  Expected:RED at lost lifetime metadata or missing lifetime-only rejection.
- [ ] Implement optional serialization and explicit refusal. Preserve existing
  initial-query/member evidence and legacy behavior. No raw handle property may
  enter report DTOs/JSON; check nested field names/values with a fake distinctive
  handle, not a broad rejection of legitimate numeric IDs. Missing fields do not
  imply coverage, ownership or exit proof.
- [ ] Run all three new suites and the approved process-free regression allowlist
  below on fresh PS7/PS5.1 hosts. Expected:every suite exit0/no failed assertions.
  Capture separate ignored logs with host/version, command and count. A failing
  unit check requires investigation, not a live probe or relaxed assertions.
- [ ] Write actual implementation/RED/GREEN/results/limitations and recovery in
  the report/status/index. Distinguish fake-adapter/static wiring tests from real
  ownership/native timing, which remain unverified. Do not claim87 fixed/matrix
  accepted; preserve historical2/1/23 and all broader acceptance blockers.
- [ ] Run fresh diff checks, source-scope review and one independent whole-increment
  review under the selected execution method; address blocking findings with
  RED/GREEN. Review explicit files, commit `diagnostics: preserve lifetime receipt isolation and verification`,
  ordinary push to verified origin/existing branch, confirm local/remote SHA and
  clean worktree. No main merge or experiment is a completion step.

## Verification commands and allowlist

All commands are planned, NOT run during plan writing. Run from the worktree.
Use fresh hosts to avoid stale Add-Type definitions. Each task's PS5.1 equivalent:

```powershell
& 'C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe' -NoProfile -File 'docs/diagnostics/isolated-fuzz/retained-lifetime.tests.ps1'
```

Substitute the exact test filename for each suite. Existing allowed suites:
matrix-check.tests.ps1, pair-sample.tests.ps1, envelope-uninstrumented.tests.ps1,
envelope-sample.tests.ps1, envelope-compile.tests.ps1, envelope-preparation.tests.ps1,
envelope-observe.tests.ps1, mapping-single-check.tests.ps1. Read their transitive
imports before execution; their native boundary must remain synthetic/default
fixture mode, with no optional real toolchain inputs. If inspection finds any
real-process path, stop that suite and document it rather than invoking it.

Do NOT run final-exit, same-handle, member-query, descendant-timeout, shutdown,
run/runner/probe or arbitrary wildcard suites: several start real processes.
initial-job-query creates empty job objects and is not a fake-only suite; omit it
from this stricter implementation scope. Do not recreate earlier233/55 counts as
if rerun. Current source/tooling pins intentionally change only approved files;
do not modify old frozen receipt pins to conceal changes.

## Plan self-review and handoff

Spec mapping:sections3/5 mode+budgets ->Task2;4/5 ownership/caps ->Task1;
6 records/disposal ->Tasks1–3;7 test matrix ->all three;8 recovery/approval ->Task3.
Five review-focus cases have owning tests above. Contracts/types/names are shared
once in the locked-interface block; no task assumes a registry in the default
mode. Unit-only evidence does not certify real native adapter behavior.

Author self-review:checked spec coverage, step/type consistency, file loader and
preflight dependencies, bounded fixture isolation, review-focus coverage and
document links. No implementation/test/native execution is claimed. User plan
approval and execution-method choice are still required. Recommended method:
inline/native execution for these three tightly coupled tasks, with one fresh
whole-increment reviewer at the end. Alternative:per-task implementer/reviewer
subagents, at higher context/review cost. Either method excludes live experiments.
