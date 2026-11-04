# Bounded Copy Transport Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Obtain one bounded CopyToAsync transport observation using the successful diagnostic28 workload.
**Architecture:** Mechanically rebind four prior child-side helper files to a fresh namespace, preserving workload statements. A new parent copies both output streams to fixed arrays, waits for the owned child and records each task independently.
**Tech Stack:** Existing PowerShell7.6.5/Core/x64, .NET Process/MemoryStream/Task/CancellationTokenSource, JSON/SHA256.
**Spec:** ../specs/2026-11-04-runtime-copy-transport-design.md

Status: PENDING_PLAN_APPROVAL / NO_IMPLEMENTATION_OR_EXECUTION.
User approved designb062c319 with 是. DesignSHA256
77C9D7C28004E13B1FCEA539E9D07A8BDFEF8C5843E920B8427618E4761A6A78.
The design's pending label is historical; this document records approval.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; existing codex/c12-b01-task9-coordinator branch/origin only.
- L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2; D=.superpowers/sdd/2026-11-04-runtime-copy-transport-v1.
- Previous D28=.superpowers/sdd/2026-11-04-runtime-host-prefix-trace-v1, immutable.
- Current28aux/7hosts/3compiles/217snapshots; proposed computation29/at most host8, ceiling29/8/3/217. No second control run.
- Exactly one hidden owned process, no target children, compiles, operational snapshots, reviewers, installs, security changes or diagnostic network.
-31operational sources and all previous helpers/stops/inputs/raw records remain byte-identical.
- Execution120000ms before buffers/start; owned-exit2000ms; shared drain2000ms. Two fixed65536-byte destination arrays,8192-byte copy buffer per task.
- Child trace remains4096events/1048576bytes; expected successful structure62events, original two groups, stdout48bytes and stderr0.
- No smoke tests/full-suite runs or repeated probes. Explicit approved finite scope overrides generic skill RED/GREEN/review/cleanup steps; static data checks are not behavioral proof.
- First unexpected preparation/execution result stops, no automatic repair/retry. Exclusive intent consumes29 before validation; process8 only counted if actually started.
- All results NOT_ACCEPTED. No merge or product fix claim.

## Review Focus

1. Capacity constructor or saving Length can create growth/padding; Task1 checks byte-array overload and terminal Position slicing, Task2 verifies exact bytes.
2. Faulted/canceled/incomplete tasks can look like empty output; Task1 defines distinct records/guards, Task2 checks task-specific availability/status/error.
3. Reading/disposal while copying risks corrupt evidence; Task1 proves terminal-task branches and retained references, Task2 records cleanup state.
4. Fresh namespace rebinding must not change workload or bypass original cutoff/guards; Task1 reverses edits, Task2 compares62event skeleton and real groups.
5. Child exit0 cannot mask transport/recording/drift failure; Task2 independently checks all inputs, task results, cleanup and output.

## Files and interfaces

Create L/runtime-copy-transport-common.ps1, -capture.ps1, -contract.tests.ps1,
-test-helpers.ps1, -parent.ps1, portable-runtime-copy-transport-preparation.json
and -final.json.
Create D/prepare.ps1, preserve.ps1, progress.md, approval.json, input.json,
transformations.json, preparation-checks.json, preparation.tool-record.json;
then intent.json, child.claim.json, stages.jsonl, capture/journal.jsonl,
child-result.json, stdout.bin/stderr.bin WHEN safely available, final.json,
tool-record.json, verification.json and stop.json on failure.
Update result docs/roadmap/2026-11-04-runtime-copy-transport-result.md,
docs/roadmap/current-status.md and docs/runbooks/repository-recovery.md.

Parent entry: -InputSHA256 <64uppercase hex>; fixed root/runtime/child/outputs.
Child/common retain HT function names and original signatures, scoped to the
new process. Common Initialize-HT binds computation29/ceiling29.
Parent helpers:
- Save-CTOnce([string]Name, Value): exclusive JSON under D, no output.
- Get-CTError(Record): actual exception type/message/ErrorId/stack, no fabricated labels.
- Complete-CTChannel(Channel): called ONLY after Task.IsCompleted; records task
  disposition/error and extracts written prefix. No task wait inside this helper.
Channel holds Name,Source,Destination,Backing,Task,CancelSource and Record.
Record fields: Name,TaskCreated,Terminal,TaskStatus,CopySucceeded,CopyError,
OutputAvailable,RetainedBytes,OutputSHA256,Capacity,CapacityFault,ReleaseDeferred.
RetainedBytes/hash stay null while unavailable, not zero/empty-string.
No synchronous output destination read/dispose in a nonterminal branch.

## Task1: Immutable preparation and static validation

- [ ] Verify correct branch/origin and all new paths absent. Read spec and actual
  successful28 helpers/input/result/verification and original driver transport
  as data. Bind D28/input.json SHA
  7E731A84FCC87999EB16F5A84DAEE527E37872728B59B04C45D16DF6E1260B57 and final archive
  8E98345719003801FF4DC7935490360EFD3D9ED9D0629029292D3F84FCE9067E.
  Expected: no drift/consumed new namespace; do not run old entry points.

- [ ] Write mechanical transformation/check machinery before generating helpers.
  Input contains Schema1, approved design/plan/approval hashes, SourceSHA256,
  PreservedSHA256, HistoricalEvidenceSHA256, HelperSHA256, AbsentPaths,
  EngineSHA256,SMASHA256,FrozenSHA256,Limits,ExpectedEventSkeleton.
  Carry forward all28bound original paths, pin its active helpers as history and
  its actual intent/claim/stages/child-result/final/tool-record/verification.
  Bind the old final archive, every old stop and new transform/approval metadata.
  Never pin a future receipt or the input to itself.

- [ ] Generate child-side artifacts by explicit AST literal-span edits only:
  * capture: namespace D28->D and imported common/contract/test-helper filenames
    host-prefix-trace->copy-transport. No function, condition or stage change.
  * common: only Initialize-HT computation/ceiling literals28->29.
  * contract.tests/test-helpers: byte-identical copies, no statement edits.
  Keep all four in L; HTRepo fixed original worktree; HTD computed new namespace.
  Record original/generated hashes and each removed/inserted span. Reverse every
  map to exact prior bytes. Use named fields for edit/rule records, explicit
  integer offset sorting and shape validation before accessing fields.
  Expected: only enumerated namespace/identity changes; original workload order,
  guards, wrapper, cutoff, terminal flush and62event structure unchanged.

- [ ] Derive ExpectedEventSkeleton from verified28 journal as data: sequence,
  EventId/ParentId,Stage,Kind,Case only; exclude elapsed values. Require62events
  and31closed intervals before freezing it; do not treat expected timings as data.

- [ ] Author new parent bootstrap and input checks. Refuse stop/intent/claim/
  partial outputs or capture directory. CreateNew intent29/ceiling29/maxprocess1
  with token and helper/input hashes before runtime validation. Verify all bound
  paths, absent pins, fixed runtime7.6.5/Core/x64 and same non-admin child context.
  Child args retain Grant contract-green, Epoch history-contract-runtime and freeze
  186FE2C2B44A31A98EA5406109AC888A625FD1B0061C5338D0382268015D9BBA.
  Use ArgumentList, fixed working directory, UseShellExecute=false,
  CreateNoWindow=true and redirect both streams. One Process.Start only.

- [ ] Implement exact bounded destination setup before start:
  allocate one byte[65536] per stream and use
  MemoryStream(byte[] buffer, int index, int count, bool writable, bool publiclyVisible)
  with index0,count65536,writable=true,publiclyVisible=false.
  Retain backing arrays privately; Position initially0. Never use capacity-only
  constructor, SetLength, expandable buffering or ToArray of the entire stream.
  Limits cover retained stream data, not an absolute bound on all runtime memory.

- [ ] After Start, launch stdout then stderr
  Source.CopyToAsync(Destination,8192,CancelSource.Token).
  Each channel owns a separate CancellationTokenSource; no cancellation scheduled
  during ordinary execution. Call one WaitForExit with nonnegative remaining
  120000ms, not a polling drain loop. Set TimedOut on false; copy task faults
  may remain unobserved during this wait. Destination still cannot expand.
  Failure while starting either copy leads to cleanup of any actual prior task.

- [ ] Cleanup ownership: resolve association using only local Process.Id
  reflection; on a still-running associated process call Kill() (not tree kill),
  then WaitForExit(2000) exactly once for exit confirmation. Preserve failure/
  uncertainty; no process lookup, adoption, retries or second process.
  Record execution elapsed before cleanup, cleanup elapsed separately.

- [ ] Shared stream cleanup: start one2000ms drain clock after exit confirmation.
  For existing nonterminal tasks, schedule CancelAfter(1500) once at cleanup entry;
  this allows most of the drain interval for EOF, then requests cancellation.
  This assistant-selected failure-path detail is not ordinary-run cancellation.
  Inspect IsCompleted with short sleeps bounded by the same remaining drain time;
  no Task.Result/Wait/GetResult until terminal. No fresh per-stream2000ms window.

  On each terminal task, independently retrieve its result/caught exception and
  save TaskStatus (RanToCompletion/Faulted/Canceled). An observed cancellation is
  failure even if child exited0. Capture every channel error even if the other
  task faults first; at most bounded error records.

  At deadline, freeze a single final terminal snapshot. Nonterminal channel:
  OutputAvailable=false, RetainedBytes/hash=null, ReleaseDeferred=true; do not
  read Position/backing bytes or create its output file. Keep source/destination/
  cancellation/task/Process references alive until this parent invocation exits;
  do not synchronously close an in-flight source merely to force completion.
  Parent returns explicit incomplete and exits; OS handle release is not claimed
  as successful stream drain. No later background evidence publication.

  Terminal channel: safe to inspect Position (0..65536) and persist exactly that
  backing-array prefix via exclusive file creation, then dispose source,
  destination and cancellation source. Dispose Process only when no channel still
  owns a nonterminal task; otherwise mark release deferred. Buffers without a
  created task can be disposed without producing synthetic captured output.

- [ ] Capacity failure interpretation: preserve actual copy exception first.
  Set CapacityFault only for an exception attributable to the non-expandable
  destination; unknown NotSupportedException remains a copy failure, not a
  guessed overflow. Never report success after a task fault. Retained prefix may
  stop before65536because a whole write failed; never pad or invent missing data.

- [ ] Static checks: named assertions NamespaceOnlyEdits, SameDirectoryImports,
  GuardWrapperCutoffUnchanged, EventSkeleton62, NonExpandableDestinations,
  PositionOnlyExtraction, BothCopiesBeforeWait, OneStart,
  TerminalOnlyExtractionAndDisposal, SharedCleanupDeadline, IncompleteNotEmpty,
  TaskErrorsPreserved, InputIdentities, NoExtraExecution.
  Inspect actual AST/edits/control flow, not just presence of strings.
  Parse all helpers and preparation script without importing or invoking them.
  Any unknown property fails preparation. No executable destination smoke probe.

- [ ] Run preparation once:
  `& ./.superpowers/sdd/2026-11-04-runtime-copy-transport-v1/prepare.ps1`
  Expected: all static checks pass,31sources and prior records unchanged,
  reversible edits, input hash, DiagnosticInvoked=false. Save actual tool call/
  result. Any unexpected result stops; preserve drafts and stop without repair.

- [ ] Independently recheck generated hashes/reversal. Archive preparation sources
  and D records with decode/length/hash/path checks; outer JSON LF, raw bytes intact.
  Update status/recovery, commit explicit files and push verified existing branch.
  Require remote==HEAD before Task2. Static success is not runtime acceptance.

## Task2: One observation and evidence-only interpretation

- [ ] Invoke exactly once:
  `& ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-copy-transport-parent.ps1 -InputSHA256 <Task1 actual hash>`
  Preserve input/initial/polls/final. Poll existing session; never replay.
  Record29consumed, host8 only if actual Start. Stop after outcome regardless.

- [ ] Parent final includes Started,OwnedExit,ExitCode,TimedOut,TerminationRequested,
  preflight/execution/cleanup/drain durations, per-channel records, errors,
  input before/after and child status. StreamsComplete only when BOTH actual
  copy tasks succeeded; OutputAvailable alone does not imply successful copying.
  Save stop on any incomplete/fault/timeout/drift/recording outcome. Missing child
  result stays missing. Do not label all copy failures as capacity overflow.

- [ ] Independently verify actual input hashes and all four recorded before/after
  maps;31sources and old stops unchanged. Validate UTF8/LF event bytes, bounds,
  monotonic ticks, complete stack, exact62event skeleton and original two actual
  successful Groups/SchemaSuite rows. Require real wrapper Cutoff and no errors.
  Require both task statuses RanToCompletion, output available48/0bytes, stdout
  exactly original CRLF availability/BaselineCore lines, empty stderr and hashes.
  Require exit0, owned exit, no deferred release, timeout, termination, overflow,
  copy/child/recording/identity/cleanup error. Only then COPY_PREFIX_COMPLETED.

- [ ] Report measured stage/parent durations beside saved28 reference, labelled
  different-run observations, not causal A/B or speedup. Both transport mechanism
  and destination type differ; original disk transport/full publication remain
  outside coverage. If unavailable output exists, report which channel and why.
  Preserve all contradictory/failure evidence. No optimization or retry.

- [ ] Archive actual D files/directories and helpers; list intentionally absent
  unavailable outputs explicitly, never create fake files to complete inventory.
  Decode-check lengths/hashes/containment and zero-length available files.
  Update report/status/recovery and counters, commit/push explicit files, verify
  remote==HEAD and clean worktree. No new run or main merge.

## Self-review and approval handoff

Task1 covers immutable rebinding, bounded transport, task lifetime/cleanup and
static proof; Task2 covers the single actual run, independent success gates,
limitations and portable recovery. Task1 input/helper hashes are Task2's interface.
All five review risks have named static checks and actual outcome gates.
Static checks do not exercise runtime failure branches; this limitation remains.
Counters28/7/3/217 unchanged. Await approval of this plan and29/8 allowance.
