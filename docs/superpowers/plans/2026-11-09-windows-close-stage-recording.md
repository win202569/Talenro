# Windows Close-Stage Recording Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add opt-in diagnostic close-stage evidence without waiting for logging on the UI close path.

**Architecture:** A preallocated64-slot single-producer/single-consumer queue feeds a background local writer. Fixed stage records surround existing disposal operations; a conservative reader accepts only trustworthy published prefixes. Instrumentation and artifacts are isolated from existing ordinary/layout builds.

**Tech Stack:** Existing pinned .NET10/C#, WinUI, xUnit and PowerShell7; no new packages.

**Spec:** docs/superpowers/specs/2026-11-09-windows-close-stage-recording-design.md (approved).

## Global Constraints

- “Reserve a fixed64-record queue.”
- “No search text, clipboard, exception messages/stacks, account data or arbitrary string payloads.”
- “Producer must not allocate per marker, lock, spin/retry, perform file/network/COM operations, wait, flush, join or dispatch UI work.”
- “No close-time flush/join.”
- “Do not catch/suppress application failures.”
- “Lost/ambiguous traces remain INCONCLUSIVE, not PASS.”
- “Normal build must exclude recorder and added call sites.”
- Preserve original artifacts; no app launch, kill, services/settings changes or main merge.
- This plan is pending user review/execution-method selection. Tests/build do not authorize UI runtime runs.

## Review Focus

1. Producer on wrong thread or reentry must not corrupt publication or affect shutdown (Task1).
2. Slow/failing storage must not block producer or keep process alive (Task2).
3. Partial/cross-run/duplicate records must not masquerade as a completed close (Task2).
4. Path alias/reparse point or existing destination must not escape/overwrite (Task3).
5. Flag combinations/output reuse must not contaminate preserved artifacts (Task4).

## File map and execution context

Worktree: D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; current development branch only.
All paths below relative to this worktree. Inspect any closer AGENTS.md before edits.
New recorder files: clients/windows/Talenro.Demo.App/Diagnostics/CloseStages/.
Pure files link into clients/windows/Talenro.Demo.Tests, following its existing AppLogic pattern;
do not pull live WinUI observer into net10.0 tests. Each task owns named behavioral tests.
Add only necessary files to commits; update status/recovery with verified versus unverified state.

PowerShell test setup, working directory clients/windows:

```powershell
$env:DOTNET_CLI_HOME = Join-Path (Get-Location) '.dotnet-home'
$env:DOTNET_SKIP_FIRST_TIME_EXPERIENCE = '1'
$env:DOTNET_CLI_TELEMETRY_OPTOUT = '1'
dotnet test ./Talenro.Demo.Tests/Talenro.Demo.Tests.csproj --no-restore
```

Before no-restore commands, verify assets exist; if needed use existing cache with locked
restore, --source ./.nuget/packages --packages ./.nuget/packages -p:NuGetAudit=false.
Do not download/install to resolve failures without authority. Offline audit disabled is not audit success.
For each task run the named filter RED then GREEN, followed by full command above.

### Task1: Typed bounded producer and stage operations

**Files:** Create CloseStages/CloseStageRecord.cs, CloseStageBuffer.cs, CloseStageScope.cs;
create Talenro.Demo.Tests/CloseStageBufferTests.cs; modify Talenro.Demo.Tests.csproj links.

**Interfaces:** CloseStage enum; readonly record struct CloseStageRecord(int Sequence,
CloseStage Stage,long Ticks); CloseStageBuffer(int producerThreadId), void Mark(CloseStage),
bool TryRead(out CloseStageRecord), int Dropped, bool InvalidProducer, bool Disabled,
void Disable(); CloseStageScope(CloseStageBuffer), void Run(CloseStage before,
CloseStage after,Action operation). Sequence starts1; frequency/correlation are immutable
session-header data, not repeated arbitrary strings in producer entries.

- [ ] Write failing BufferPreserves64ThenDrops65th: enqueue64, assert65th increments Dropped,
  drain exact ordered64 with no overwritten entries. Publication test runs producer/consumer
  concurrently using barriers, verifies unique contiguous sequences and intact stage/tick pairs.
- [ ] Add WrongThreadInvalidatesWithoutThrowing and DisabledProducerReturnsWithoutEnqueue tests;
  producer identity bound before launch, no thread-local capture per Mark. Warm producer first;
  AllocationFreeMark asserts current-thread allocated-byte delta0 over a no-overflow64-record batch.
- [ ] Add ScopePreservesOriginalException: before recorded, operation throws a sentinel exception,
  same instance propagates and after absent. Cached delegate supplied by caller, never constructed
  on close path. Reentrant markers remain ordered on same producer thread; no lock/spin.
- [ ] Run --filter FullyQualifiedName~CloseStageBufferTests and observe RED for missing behavior.
- [ ] Implement slot write then release publication; consumer acquire before slot read, release
  consumed position before reuse. Drop on full, atomic health invalidation, no producer wait.
  Use vetted .NET memory-ordering documentation at implementation time, not assumed visibility.
- [ ] Run filtered GREEN and full suite; record results and commit explicit Task1 files.

### Task2: Background transport and conservative receipt reader

**Files:** Create CloseStages/CloseStageSession.cs, CloseStageReceiptReader.cs;
create Talenro.Demo.Tests/CloseStageTransportTests.cs; extend pure-source links.

**Interfaces:** CloseStageSession(Guid runId,CloseStageBuffer buffer,Stream ownedStream),
void Start(), bool Ready, bool Failed; background worker never joined by app close.
CloseStageReceiptReader.Read(Stream input,Guid expectedRun) returns CloseStageReceipt
with Status (ValidPrefix/Inconclusive), immutable Records and CompletedClose bool.
CompletedClose means only a published matching CloseReturn, never process exit/acceptance.

- [ ] Add ReadyPrecedesStages (schema1, runId N form, frequency>0); StageLinesRoundTrip;
  terminal CloseReturn causes drain/health flush and worker completion when writer can progress.
  Poll every10ms off UI thread; one initial ready and one final health record, no unbounded heartbeat.
- [ ] Add BlockedWriterDoesNotBlockMark using a controllable blocking Stream and a test deadline;
  release it in test cleanup. Assert worker IsBackground. FailureStream disables recording and
  reports Failed without changing producer exception behavior. Never use sleep as test correctness.
- [ ] Add RejectTruncatedLine, WrongRun, DuplicateSequence, SequenceGap, BackwardTicks,
  UnknownStage, NonpositiveFrequency, InvalidStageOrder, ReportedDropOrFailure.
  BeforeWithoutAfter yields ValidPrefix only, CompletedClose=false; missing ready is Inconclusive.
  Missing final health makes complete trace inconclusive even if partial records were parseable.
- [ ] Run --filter FullyQualifiedName~CloseStageTransportTests; observe RED, implement JSONL
  writer/reader, then GREEN/full suite. Reader caps input64KiB/128 lines; tests reject excess input.
  Serialize only on writer, never producer. Fixed health codes only; no error messages or stacks.
- [ ] Commit explicit Task2 files with tail-loss/blocked-writer limitations documented.

### Task3: Safe activation and original close-path hooks

**Files:** Create CloseStages/CloseStageActivation.cs; modify App.xaml.cs,
Diagnostics/LayoutWindowObserver.cs, Layout/ReservedLayoutArguments.cs;
create Talenro.Demo.Tests/CloseStageActivationTests.cs; update links.

**Interfaces:** CloseStageActivation.Parse(IReadOnlyList<string>) returns extracted
optional Guid RunId, remaining layout arguments and fixed error code. Exactly one
--layout-close-run=<32 lowercase hex digits> is required when tracing is compiled;
reject empty Guid, duplicate/case-altered flag, malformed value or missing --layout-test.
CloseStageActivation.OpenNew(string trustedRoot,Guid runId) returns owned Stream;
file name runId.ToString("N") + ".jsonl", FileMode.CreateNew semantics, reader sharing.

- [ ] Write ParseAcceptsOneRun, RejectsDuplicateOrMalformedOrMissingLayout, CreatesNewOnly,
  RejectsExistingDestination, RejectsReparseAncestors and RootContainment tests.
  Trusted root is fixed relative to AppContext.BaseDirectory: close-stage-records/ beneath
  the isolated ignored output directory. No user-supplied output path, no cwd-relative escape.
  Existing parent components must be checked; recheck after creation/open and fail closed on
  detected changes. Do not claim this defeats hostile concurrent filesystem replacement.
- [ ] Run --filter FullyQualifiedName~CloseStageActivationTests RED. Implement activation,
  initialize buffer/session on UI thread before MainWindow; require writer ready before window
  creation with a bounded2s startup wait (never close-time wait). Startup failure exits2 using
  existing rejection style. Ordinary builds already reject --layout-prefixed arguments.
- [ ] Under CLOSE_STAGE_RECORDING only, pass recorder into observer. Add CloseEntry,
  DisposeEntry, AlreadyDisposedReturn, before/after all nine removal boundaries, explicit
  XamlRootSkipped/UISettingsSkipped, before/after root-reference clear, DisposeReturn,
  CloseReturn. Keep disposed assignment and removal order exactly unchanged.
  Prebind removal Action delegates at construction for CloseStageScope.Run; conditional
  native-object checks occur at their original positions, not at construction.
- [ ] Add named stage-order examples covering normal, already-disposed, null conditional
  subscriptions and each operation throwing. Use actual scope helper and sentinel operations;
  explicitly label these pure-flow tests, not proof of WinUI event removal or live callback wiring.
- [ ] Run GREEN/full suite, inspect actual observer diff for order/exception preservation,
  compile via Task4 before claiming hooks build. Commit only scoped changes, no runtime verdict.

### Task4: Artifact isolation and build-only handoff

**Files:** Modify clients/windows/Directory.Build.props and Talenro.Demo.App.csproj;
create tests/verify-close-stage-artifacts.ps1 and tests/close-stage-build-gates.tests.ps1;
update diagnostics documentation, docs/roadmap/current-status.md, docs/runbooks/repository-recovery.md.

**Interfaces:** EnableCloseStageRecording=false default; true requires Configuration=Debug
and EnableLayoutDiagnostics=true. Define CLOSE_STAGE_RECORDING only in accepted combination.
Dedicated BaseIntermediateOutputPath=obj/close-stage-recording/ and BaseOutputPath=
bin/close-stage-recording/ (MSBuildProjectExtensionsPath agrees); precedence over layout paths.
Exclude CloseStages/**/*.cs when off; all hook sites inside conditional compilation.
Metadata verifier parameters OrdinaryAssembly, LayoutAssembly, CloseStageAssembly, checks
three distinct paths and recorder types absent in first two/present only in third; execute no DLL.

- [ ] Before builds, record hashes/paths of original demo layout EXE/DLL and both prototype
  EXE/DLL pairs from existing ledgers. Never build directly into those original paths.
- [ ] Write build-gate tests evaluating MSBuild properties and invalid combinations; Release+on
  and Debug+on/layout-off must fail with CLOSESTAGE001 before compilation/restore side effects.
  Assert off retains existing output selection; on uses dedicated paths. Observe RED.
- [ ] Implement gates and run GREEN. Build ordinary and layout-off-recorder verification copies
  with command-line BaseOutputPath/BaseIntermediateOutputPath/MSBuildProjectExtensionsPath
  directed to fresh ignored obj/close-stage-verification/{ordinary,layout}/ and matching bin
  directories; evaluate effective paths before building. Preserve canonical original outputs.
- [ ] Locked restore and build instrumented App with -c Debug -p:Platform=x64
  -p:EnableLayoutDiagnostics=true -p:EnableCloseStageRecording=true; --no-restore for build.
  Run three-artifact metadata verifier, full .NET suite, build-gate tests and git diff --check.
  Require original six hashes unchanged, no binary/cache/log staged. Do not launch App.
- [ ] Review all changes (including test limitations), update portable recovery commands and
  remaining runtime gates, commit explicit file list and push existing development branch;
  verify remote SHA equals local HEAD and worktree clean. Report all failures, no acceptance promotion.

## Self-review and handoff

Design coverage: producer/stages Task1+3; writer/interpretation Task2; activation Task3;
build isolation Task4. Review Focus mapped to named tests. Runtime callback behavior remains
unverified by design, and a full marker sequence is not a process-exit claim. Allocation tests
need warmup, blocked-writer tests need guaranteed release, and path checks are scoped to the
ordinary trusted workspace rather than hostile filesystem protection.

Recommended execution: native sequential implementation in this session because all four
tasks share the record/session contract, followed by one independent whole-change review.
Alternative: separate implementer/reviewer per task. User chooses method before execution.
Any later WinUI launch requires a separate bounded approval; zero launch budget in this plan.
