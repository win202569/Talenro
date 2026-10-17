# Ordinary Member Identity Ledger Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an isolated, fake-testable ordinary identity-ledger path without enabling workload acceptance or changing existing callers.

**Architecture:** One run-local ledger owns each admitted handle from early acquisition through final exit. A shutdown session charges all work to shared clocks and a separate pure receipt evaluator refuses incomplete evidence. Existing ordinary/diagnostic modes stay isolated; schema4 remains refused by the existing acceptance gate.

**Tech Stack:** Windows PowerShell7/5.1, C# compatible with both hosts' Add-Type, injected fake native/clock interfaces, existing limited-rights Windows adapter.

**Spec:** [Approved written design](../specs/2026-10-17-ordinary-member-identity-ledger-design.md), approved after a3abff3305248df151d0357a75d0aba7d5923b7c.

Status: user approved plan/30-invocation budget and selected Native/inline after
89cb1ad0. Task1 stopped on unexpected Cap32 test failure after3 invocations;
Tasks2–4 pending. See the execution report. Planning itself ran no tests/workload.

Execution addendum after557c2249: user approved fixture-only capacity correction
and total budget31. Replacement GREEN7 passed;first GREEN5 stopped at unsupported
script type [ulong],5/31 used. No further fix/run authorized yet. Original30-call
arithmetic below is historical;remaining plan is unchanged. Proposal:script-only
UInt64 compatibility correction and total32,then replacement GREEN5 and remaining
12 task RED/GREEN +14 final calls. User approved this correction/budget32 after
8dafbfe3;replacement GREEN5 passed9 assertions at invocation6. Task1 complete,
Task2 next. The earlier stop/proposal statements above are historical.
Latest execution state:Task2 completed on both hosts (14 assertions each);
Task3 expected REDs completed,but first GREEN7 failed NestedRoundtrip after2
PASS lines at invocation13. Stopped13/32;Task4/final review unstarted. The JSON
rejection reason was not emitted;no repair/retry after the stop. See report.
Current:after468709b6,user approved diagnostic/repair,total34. UTC metadata
representation corrected;Task3 passed both hosts. Task4 GREEN and all14 final
allowlist runs passed;350 final assertions.34/34 consumed. Independent readonly
review pending;further repair/reverification needs a new bounded decision.
Earlier stop/budget statements above and original30-call arithmetic are historical.

## Global Constraints

- Use existing0x101000 rights,inherit=false; capture last-error immediately.
- Maximum32 non-root acquisition attempts/records/retained handles; failures consume attempts. Reaching32 marks cap-exhausted and refuses.
- The natural window remains500ms and total cleanup ceiling5000ms.
- Use ShutdownSchema4, policy ordinary-owned-ledger-final-exit-v1, and MemberIdentityLedger Schema1, Mode=ordinary-identity-ledger-v1, Coverage=sampled.
- Require final TotalProcesses = 1 + Admitted; count alignment is necessary, not a complete event trace.
- RunIdentity plus EntryId binds a retained handle; no reopen, cross-run/PID-only join, eviction, retry or archived-PID lookup.
- Existing Run overloads, diagnostic contracts, callers and acceptance behavior stay unchanged except explicit refusal of new markers. Assert-ProcessCompletion rejects schema4.
- No native workload, Go build, SDK modification, installs, matrix replay, metadata continuation or acceptance activation in this implementation stage.
- No old receipt rewrite/replay. Metadata1-build/1-runtime and independent probe0-build/1-runtime remain consumed and unaccepted for metadata purposes.

## Review Focus

- Counter mismatch after an otherwise empty final job must refuse (Task2, MissedMemberCount).
- A failed close followed by Dispose must not retry a possibly reused raw handle (Task1, CloseFailureOnce).
- Huge/overflowing/backward clock inputs must refuse without a fresh deadline (Task2, DeadlineOverflow).
- JSON strings/numbers/null substituted for booleans must not be coerced into successful evidence (Task3, StrictTypes).
- A stripped schema with a retained new-mode marker must not fall through legacy acceptance (Task3, MarkerDowngrade).

## Workspace, files and contracts

Reuse D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator on branch
codex/c12-b01-task9-coordinator; do not create another checkout or merge main.
At execution start read both spec/plan, AGENTS.md and applicable worktree skill,
then inspect dirty state. Preserve unrelated edits; stop on overlapping ambiguity.
Below, H = docs/diagnostics/isolated-fuzz. Paths prefixed H are relative to this
workspace. Test scripts resolve sources from PSScriptRoot, not private caches.

Create H/ordinary-ledger.cs: ordinary DTOs/interfaces/acquisition/disposal.
Create H/ordinary-ledger-session.cs: shared-budget finalization only.
Create H/ordinary-ledger-receipt.ps1: strict pure schema4 evaluation.
Create H/ordinary-ledger-fixture.cs and H/ordinary-ledger-fixture.ps1: deterministic
fake adapter/clock/factories and source loader, no Windows calls or process starts.
Create four suites named below; modify process-native.cs, run.ps1, preflight.ps1,
retained-lifetime-integration.tests.ps1 source loader and README.md in Task4 only.
Do not refactor retained-lifetime.cs or alter its diagnostic registry semantics.

Reuse existing transport DTOs only: LifetimeCandidates, LifetimeOpen,
LifetimeMembership, LifetimeRelease and LifetimeApiEvidence from retained-lifetime.cs.
Do not reuse LifetimeReport/registry entries as ordinary evidence.
Use IRetainedLifetimeClock for Ticks/Frequency/UtcNow, without changing it.

New exact public contracts:

```csharp
public interface IOrdinaryLedgerApi {
    LifetimeCandidates QueryMembers(IntPtr job, long deadlineTicks);
    LifetimeOpen Open(uint pid, uint access, bool inherit);
    LifetimeMembership IsOwned(IntPtr handle, IntPtr job);
    OrdinaryRead Read(IntPtr handle, uint waitMs, long deadlineTicks);
    LifetimeRelease Close(IntPtr handle);
    OrdinaryAccounting QueryAccounting(IntPtr job, long deadlineTicks);
}
public sealed class OrdinaryMemberLedger : IDisposable {
    public OrdinaryMemberLedger(string runIdentity, IntPtr job, uint rootId,
        IOrdinaryLedgerApi api, IRetainedLifetimeClock clock);
    public OrdinaryLedgerReport Report { get; }
    public void Poll(long deadlineTicks);
    public void ObserveFinal(long naturalDeadlineTicks, bool terminationRequested);
    public void SealAndRelease();
    public void Dispose();
}
public sealed class OrdinaryLedgerSession {
    public OrdinaryLedgerSession(OrdinaryMemberLedger ledger, IntPtr job,
        IOrdinaryLedgerApi api, IRetainedLifetimeClock clock);
    public OrdinaryFinalOutcome Finish(bool rootExited, bool timedOut,
        long shutdownStartTicks);
}
```

OrdinaryRead fields: RunIdentity:string, EntryId:int (assigned by ledger),
SameHandleVerified/Attempted/Signaled/ExitCodeRead/TerminationRequested:bool,
ExitCode:uint, Error/ImageError:int, State/ImagePath/SkipReason:string,
StartTimestamp/EndTimestamp/Frequency:long, StartUTC/EndUTC:string.
OrdinaryAccounting extends LifetimeApiEvidence with
TotalProcesses/ActiveProcesses/TotalTerminatedProcesses:uint.
OrdinaryLedgerReport exposes Schema:int=1, RunIdentity/Mode/Coverage:string,
RootId:uint, Attempts/Admitted/QueryCount/ReleaseAttempts/ReleaseFailures:int,
Incomplete/Ambiguous/CapExhausted/TeardownAttempted/Sealed:bool,
Errors:List<string> (maximum96), FirstQuery/LastQuery:LifetimeQueryEvidence,
Entries:List<OrdinaryEntry> (maximum32).
OrdinaryEntry contains RunIdentity:string, EntryId/SourceQuery:int, Pid:ulong,
SourceQueryEvidence:LifetimeQueryEvidence, Open:LifetimeApiEvidence,
Membership:LifetimeMembership, Initial/Final:OrdinaryRead,
ReleaseAttempted/ReleaseSucceeded:bool, ReleaseError:int.
OrdinaryFinalOutcome holds EvidenceComplete/NaturalShutdownVerified:bool,
FinalAccounting:OrdinaryAccounting, FinalMembers:ulong[], CountReconciled:bool,
FinalMemberQuery:LifetimeQueryEvidence, NaturalObservationMs:long and
Errors:List<string>; no Acceptance=true field.

These signatures are API decisions; write implementations only after plan approval.

## Verification envelope proposed for approval

Run each suite in a fresh64-bit host from the workspace:

```powershell
pwsh.exe -NoLogo -NoProfile -NonInteractive -File docs/diagnostics/isolated-fuzz/ordinary-ledger.tests.ps1
& "$env:SystemRoot/System32/WindowsPowerShell/v1.0/powershell.exe" -NoLogo -NoProfile -NonInteractive -File docs/diagnostics/isolated-fuzz/ordinary-ledger.tests.ps1
```

Substitute only the explicit suite filename in each task. Hosts must already
exist; if unavailable, stop/report, do not install. No wildcard suite execution.
Read the entire selected suite and transitive loaders before its first execution
to confirm fake-only scope. Shell/Add-Type compilation is allowed after approval;
it is not a Go build or a native target workload. Do not claim zero OS processes.

Proposed maximum30 suite invocations:4 tasks × RED/GREEN ×2 hosts=16, followed
by one final run of4 new suites plus3 listed legacy suites ×2 hosts=14.
Each invocation has120s wall limit; stop on timeout, unexpected error, wrong RED,
scope violation or exhausted budget. No automatic rerun/extra diagnostic loop.
Expected RED is only a named planned assertion/missing planned type or member;
infrastructure/compiler syntax failure is not accepted RED. Fixing an unexpected
test/infrastructure failure and rerunning needs a new bounded decision. Capture
command/host/result/elapsed/counts in the execution report; raw output stays ignored.
This budget is proposed, not granted by the already-approved design.

Final legacy allowlist (read-only/fake sources inspected before use):
retained-lifetime.tests.ps1, retained-lifetime-integration.tests.ps1,
retained-lifetime-receipt.tests.ps1. Do not invoke run.tests.ps1,
member-query.tests.ps1, shutdown-observation.tests.ps1, descendant/child fixtures
or any controller. Existing loaders may compile P/Invoke declarations but tests
must never invoke native adapters/Run/RunWithOrdinaryIdentityLedger.

### Task 1: Bounded acquisition and owning ledger

**Files:** Create H/ordinary-ledger.cs, H/ordinary-ledger-fixture.cs,
H/ordinary-ledger-fixture.ps1, H/ordinary-ledger.tests.ps1.
**Interfaces:** Consumes existing transport/clock DTOs above. Produces
IOrdinaryLedgerApi, OrdinaryRead/Accounting/Entry/Report and OrdinaryMemberLedger.
Fixture factory OrdinaryLedgerFixture.Create() returns Clock, Api and Ledger;
defaults: run-a, job handle1, root1, candidate2, owned handle900002, running state.
Fake Api exposes OpenError:int, OpenCalls/MembershipCalls/ReleaseCalls:int,
ThrowClose:bool, Candidates:ulong[] and Calls:List<string>; Clock is settable.
The fixture loader compiles all existing native/transport/compile-stage sources
plus the new ledger/session sources as they become available in one Add-Type
batch, then fixture sources in that same batch, without calling run.ps1. It does
not invoke any compiled native method. Load this batch before dot-sourcing
run.ps1 in Tasks3/4, avoiding cross-assembly references to in-memory types.

- [ ] Write AcquisitionOwnsOnce, Open87Sticky, ForeignReleased, CrossRunIdentity,
  Cap32, InvalidCandidate and CloseFailureOnce assertions. Pin
  rights0x101000/inherit=false and zero membership calls after open87. Example:
  `$f=[OrdinaryLedgerFixture]::Create(); $f.Api.OpenError=87; $f.Ledger.Poll(1000); $f.Ledger.Poll(1000)`
  must yield OpenCalls=1, MembershipCalls=0, Incomplete=true, Open.Error=87.
  Duplicate candidate2 opens once; candidate1 never opens;32 attempts refuse and
  candidate33 is not opened. Close failure followed by Dispose calls Close once.
- [ ] Run ordinary-ledger.tests.ps1 RED once on each host; record planned failure.
- [ ] Implement constructor/Poll/SealAndRelease/Dispose with temporary→owned
  transfer, sticky bounded errors, exact-handle initial reads and idempotent
  release. Constructor validates nonempty RunIdentity, own-job/root and clock.
  Reserve ObserveFinal implementation for Task2; it must not emit success yet.
- [ ] Run the same suite GREEN once per host; expected exit0 and all named PASS.
- [ ] Review explicit file diff, diff-check and checkpoint:
  `git add -- <the four files above>`; commit `feat: add bounded ordinary identity ledger`.

### Task 2: Same-handle finalization and count reconciliation

**Files:** Modify H/ordinary-ledger.cs/fixture.cs; create
H/ordinary-ledger-session.cs and H/ordinary-ledger-session.tests.ps1.
**Interfaces:** Consumes Task1 ledger/API; implements ObserveFinal and produces
OrdinaryLedgerSession/OrdinaryFinalOutcome. Fake adapter supplies queued
OrdinaryRead/accounting/list results and advances Clock inside calls, with no sleep.

- [ ] Write AbsentAtRootStillVerified, LateCandidate, EmptyButRunning,
  MissedMemberCount, ReleaseBeforeAccounting, NewMemberAfterSeal, Shared500ms,
  DeadlineOverflow, BackwardClock, ForcedExitRefused and ExitedThenRunning tests.
  A signaled member with total3/admitted1 refuses CountReconciled; total2/admitted1
  and all valid exits may yield EvidenceComplete but never Acceptance=true.
  At ticks frequency1000, shutdownStart0/deadline500, two member waits must sum
  to at most the remaining500ms, not500ms each. Overflow/negative elapsed refuses.
  Calls must show final reads before child close before final accounting/list.
  A fake close error cannot erase an earlier open/read failure or double-close.
- [ ] Run ordinary-ledger-session.tests.ps1 RED once per host.
- [ ] Implement ObserveFinal using retained entries and remaining deadline only.
  Finish polls/acquires each bounded shutdown batch before member waits, retaining
  initial facts; newly observed members obey Task1 rules. Once all required
  states verify, seal/release then query accounting/list, require empty/zero,
  no limit termination and TotalProcesses=1+Admitted. No admission after sealing.
  Deadlines use checked arithmetic; every API completion checks clock/budget.
  Root timeout/not-exited skips natural success but never drops release ownership.
  Failure result tells runner to perform existing forced cleanup, not to retry.
- [ ] Run the session suite GREEN once per host.
- [ ] Review/checkpoint exact task files; commit `feat: finalize ledger evidence within shared shutdown budget`.

### Task 3: Strict schema4 evaluation and downgrade refusal

**Files:** Create H/ordinary-ledger-receipt.ps1 and
H/ordinary-ledger-receipt.tests.ps1; modify H/run.ps1 refusal checks only.
**Interfaces:** Consumes Task1/2 report/outcome DTOs. Produces pure PowerShell
`Test-OrdinaryLedgerEvidence([object]$Receipt)` returning a single object with
EvidenceComplete:bool, Acceptance:bool=false, Errors:string[]. It never launches
or coerces untrusted field types into DTO defaults. Receipt fixture factory
New-OrdinaryEnvelope in this suite constructs the full schema4 envelope from
fake Task2 output with matching RunIdentity at root/ledger/entry/read levels.

- [ ] Write CompleteButNotAccepted, StrictTypes, NestedRoundtrip, MarkerDowngrade,
  MissingRead, WrongRunIdentity, DuplicateEntry, ContradictoryTiming and
  DiagnosticContamination tests. A valid fixture yields EvidenceComplete=true,
  Acceptance=false but Assert-ProcessCompletion still throws. Mutate each required
  boolean to null,0,1,'true','false'; remove/version-change each marker one at a
  time; all refuse. Wrong/missing counts, zero/negative frequency, out-of-order
  call bounds, post-termination exit, raw-handle fields and duplicate IDs refuse.
  A legacy schema3 fixture remains unchanged; a new ledger marker under it refuses.
- [ ] Run ordinary-ledger-receipt.tests.ps1 RED once per host.
- [ ] Implement evaluator validation before conversions: schema/policy/mode,
  exact boolean/numeric/string/array shapes, root/run/entry references, API success,
  signaled+exit read, monotonic/elapsed bounds, all sticky flags, releases and
  reconciliation. Return errors bounded96. Add early new-marker refusal to
  Assert-ProcessCompletion; never add schema4 to accepted schemas. No legacy
  behavior change for receipts with no new-mode markers; no diagnostic enrichment.
- [ ] Run the receipt suite GREEN once per host.
- [ ] Review/checkpoint exact files; commit `feat: validate isolated ledger receipts without acceptance activation`.

### Task 4: Explicit native wiring and portable regression handoff

**Files:** Modify H/process-native.cs, H/run.ps1, H/preflight.ps1,
H/retained-lifetime-integration.tests.ps1 (compile list only), H/README.md;
create H/ordinary-ledger-integration.tests.ps1. Update current-status.md,
repository-recovery.md and create docs/roadmap/2026-10-17-ordinary-ledger-implementation.md.
**Interfaces:** Consumes Tasks1–3. Produces
`IsolatedFuzzProcess.RunWithOrdinaryIdentityLedger(string executable,string[] args,IDictionary overrides,string cwd,string stdout,string stderr,int limitSeconds):Result`.
Result gains nullable MemberIdentityLedger:OrdinaryLedgerReport,
LedgerFinalOutcome:OrdinaryFinalOutcome and RunIdentity:string. Serialize these
only when this new mode is active; legacy null defaults must not create markers.

- [ ] Write ExplicitEntryOnly, LegacyDefaults, ModeIsolation, LoaderInventory,
  SessionFakeCleanup and SerializationRefusal tests. Use reflection/static source
  checks for entry/signature/cadence/hook order; use fake ledger/session and an
  injected throwing cleanup Action for exception/disposal order. Do not invoke
  the new entry or native adapter. Assert old defaults schema3/no ledger,
  two existing Run overloads and existing diagnostic flags remain unchanged.
  New result serialization preserves schema4/run binding but cannot pass the gate.
- [ ] Run ordinary-ledger-integration.tests.ps1 RED once per host.
- [ ] Add nested native IOrdinaryLedgerApi adapter using existing limited-rights
  declarations and immediate native-error capture. Wire only the new entry to
  early Poll,10ms root wait slices and Task2 finalization with shared clocks;
  expose final facts but let failure take the unchanged owned-job cleanup path.
  Fresh RunIdentity is created before launch. Preserve primary exceptions and
  release outstanding owners in finally. Do not use diagnostic rows as inputs.
  Extend loaders/inventory and serializer conditionally; no Invoke-RecordedProcess,
  matrix or metadata controller migration. Update the legacy integration compile
  list only as needed, keeping its existing assertions unchanged.
- [ ] Run the integration suite GREEN once per host.
- [ ] Run all4 new suites and the3 legacy suites in the exact final allowlist,
  once per host. Expected14 invocations exit0; report fake/static limitations,
  not live adapter correctness or Error87 repair. Stop on any unexpected failure.
- [ ] Review full diff/spec coverage and obtain final independent review before
  claiming implementation complete. If Native chosen, request one fresh final
  reviewer; do not silently spawn implementers. Review must not run workloads or
  spend extra suite budget. Required code repairs/reverification need a separately
  bounded decision when the approved30 invocations have been consumed.
- [ ] Record counts, all RED/GREEN outcomes, review findings, unresolved issues and
  native/rollout gates in the execution report; update recovery/status. Checkpoint
  only explicit relevant files, push the existing branch, verify remote SHA=local
  and working state. Commit `feat: wire isolated ordinary ledger and preserve blocked rollout`.

## Self-review and approval handoff

Spec coverage: acquisition/identity→Task1; shared clocks/final exits/release/counts→
Task2; strict schema/isolation→Task3; adapter/entry/portability/recovery→Task4.
Every Review Focus item has a named owning test. Interfaces/types match across
tasks; DTO reuse does not reuse diagnostic observations. Static tests are labelled
static, fakes are not native validation. This plan makes no test success claim.

Implementation recommendation: Native/inline because these4 tasks share tight
contracts and the scope is fake-only; one fresh final review remains required.
Subagent-driven is an optional user-selected alternative, not authorized by this
plan alone. No agent has been dispatched during planning.

Next gate: user reviews this plan, approves its proposed30-invocation envelope
and selects Native/inline or Subagent-driven. Approval of the spec did not approve
this plan/budget. Later live experiments, ordinary acceptance activation, metadata
resumption and main/production integration each remain outside this plan.
