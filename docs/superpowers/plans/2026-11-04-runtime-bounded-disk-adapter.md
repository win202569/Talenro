# Bounded Disk Adapter Verification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build and verify only a pre-write-capacity-limited file adapter; do not run a business host.
**Architecture:** One C# assembly contains the adapter and deterministic verification harness. One separately owned PowerShell worker compiles it once and runs the fixed matrix; a parent observes bounded in-memory output and enforces the worker deadline. Sources and evidence are portable; the generated DLL remains local.
**Tech Stack:** Existing pinned PowerShell7.6.5/Core/x64, .NET Stream/FileStream/Task, C# via Add-Type, JSON/SHA256.
**Spec:** ../specs/2026-11-04-runtime-bounded-disk-adapter-design.md

Status: PENDING_PLAN_APPROVAL / NO_IMPLEMENTATION_OR_EXECUTION.
User approved design commit dcc3b5866003b6a5d784f11ea20725ccbe9afaeb with 是.
DesignSHA256:1CE2438076B4BDF3D5A8C4D6EB6E1A3D8BFB5AB860EBAD01665CE8DFDC1EF421.
Its pending label is historical; approval permits this plan, not execution.

## Global Constraints

- Existing worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator,
  branch codex/c12-b01-task9-coordinator, origin github.com/win202569/Talenro.git.
- L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2.
  D=.superpowers/sdd/2026-11-04-runtime-bounded-disk-adapter-v1.
- Current29aux/8hosts/3compiles/217snapshots. PROPOSED exactly computation30,
  at most one owned verification worker (conservatively counted as host9),
  at most one compilation (compile4): ceiling30/9/4/217.
  No business host, grandchildren, auxiliary process, external compiler launch,
  operational snapshot, reviewer, install, network-in-diagnostic or security change.
- Parent invocation creates exclusive intent before validation; computation30 is
  consumed even if validation/start fails. Count process only on actual Start.
  Worker records exclusive compile.intent before Add-Type; failure consumes compile
  allowance, never retry. No standalone compile, test or smoke invocation.
- Execution120000ms before buffer/start; owned-exit2000ms; shared transport drain
  2000ms. Parent uses fixed65536-byte arrays and8192-byte copy buffers, not the
  new disk adapter. Retain saved29 terminal-task evidence and ownership rules.
- Adapter limit65536bytes per sink. Fixed24cases, maximum2 newly created files per
  case,48data files total/3145728logical bytes maximum. Each under D/cases/NN/.
  Fake sinks also retain at most65536bytes. No cleanup/deletion of evidence.
- Child journal at most256events/262144UTF8bytes; bounded48case begin/end events
  plus compile/verification/finalization markers. Results at most262144bytes;
  each error message/stack at most4096characters with explicit truncation flags.
-31operational sources, all saved diagnostics/inputs/stops remain byte-identical.
  Do not run consumed entry points. Reuse fixed runtime/non-admin identity checks.
- First unexpected static preparation, compile, assertion, deadline or recording
  failure stops the attempt. Expected exceptions in named cases are assertions,
  not unexpected stops. No automatic repairs, reruns or recompile.
- Tests are authored before implementation; the one-compile gate deliberately
  omits a separately executed RED cycle. Do not claim executed TDD or add runs.
  This finite native scope supersedes generic extra review/full-suite workflows.
- NOT_ACCEPTED and all five capability flags false. No main merge/product fix.
  Adapter success never authorizes a subsequent business-host observation.

## Review Focus

1. Alternate write entry points bypassing the byte cap: cases01-10 audit all routes.
2. Partial failed/canceled writes refunded incorrectly: cases14-16 retain reservation and latch failure.
3. Concurrent write/flush/dispose corrupting state: cases17-19 reject contention without closing the active sink.
4. Pending/failed finalization mistaken for stable empty output: cases20-22 deny availability until successful closure.
5. Real filesystem/parent shutdown behaving differently from fake sinks: cases23-24 cover actual files; worker timeout and unresponsive-storage limitations remain explicit.

## Files and interfaces

Create in L:
- runtime-bounded-disk-adapter.cs: sealed BoundedFileSink and CapacityExceededException.
- runtime-bounded-disk-adapter.tests.cs: controlled sinks and fixed Verification.RunAsync.
- runtime-bounded-disk-adapter-worker.ps1: authenticate, compile/load, run once, save receipts.
- runtime-bounded-disk-adapter-parent.ps1: owned-process deadline and bounded transport only.
- portable-runtime-bounded-disk-adapter-preparation.json and -final.json.

Create in D: prepare.ps1, preserve.ps1, progress.md, approval.json, input.json,
preparation-checks.json, preparation.tool-record.json, preparation.verification.json;
then intent.json, worker.claim.json, compile.intent.json, compile.result.json,
adapter.dll (local only), stages.jsonl, cases/01..24/, verification-result.json,
stdout.bin/stderr.bin when safe, final.json, stop.json on failure, tool-record.json,
verification.json and completion.md. DLL is never committed or base64-archived;
record its hash/length and reproducible source/command. Compiler/runtime binaries
and caches are excluded. Assembly size postcheck<=16MiB, not a filesystem quota.

Update docs/roadmap/2026-11-04-runtime-bounded-disk-adapter-result.md,
docs/roadmap/current-status.md and docs/runbooks/repository-recovery.md.

Public namespace Talenro.Diagnostics.BoundedDiskV1:
- BoundedFileSink.OpenNew(string absolutePath) -> BoundedFileSink.
- long ReservedBytes, long ConfirmedBytes; bool Faulted, bool Closed.
- ordinary Stream write/flush/disposal overrides, CanRead=false, CanSeek=false.
- Task CloseAsync(): single-flight successful closure or actual failure.
- CapacityExceededException : IOException, unique to the capacity gate.
Internal BoundedFileSink(Stream sink) permits harness-controlled sinks only;
no public raw stream/handle accessor. Production factory owns its new FileStream.

Harness: Verification.RunAsync(string casesRoot) -> Task<VerificationReport>.
Report: Schema, Status, CaseResults[24] with Id/Name/Passed/Error/Observed,
Files with relativePath/Length/SHA256 for stable files, UnavailablePaths,
PendingOperations, Error, Truncated, counters and all five false flags.
No dynamic discovery, generated test cases or parameter fuzzing.

### Task1: Author test matrix, adapter and immutable preparation

**Consumes:** approved spec and this plan; saved29 input/helper/final archive.
**Produces:** frozen input SHA and preparation archive, no compiled code or test run.

- [ ] Check branch/origin/clean overlap and new paths absent. Read complete saved29
  parent and common checks as data. Pin saved29 input
  E45AAD367F9AF3C30012FF870186682552FD4AB30DE344A885A9A8FE2CE1A845 and final archive
  4887A84DD04E8D9D6FA2E1B40803C0E7EE9296C6C90AE0741F81A323C52B10D7.
  Carry forward its source/preserved/history/helper maps and absent pins; add its
  actual final/tool-record/verification, this spec/plan/approval and new sources.
  Do not bind future receipts or input to itself.

- [ ] Write tests.cs first. Fixed ordered matrix below; every case requires actual
  assertions and emits exactly one result. Exceptions outside the named expectation
  stop remaining cases. Data patterns are deterministic byte(index modulo251).
  Each case creates at most two files named a.bin/b.bin, without external temp paths.

| ID | Test | Required assertions |
| --- | --- | --- |
|01|ArrayWrite|Write65536; next1 throws capacity before delegation; exact original65536 retained|
|02|SpanWrite|Same boundary through Write(ReadOnlySpan<byte>)|
|03|WriteByte|Array fill65535, WriteByte succeeds, next byte rejected; length65536|
|04|ArrayAsyncToken|Same boundary through WriteAsync(byte[],offset,count,token)|
|05|MemoryAsyncToken|Same boundary through WriteAsync(ReadOnlyMemory<byte>,token)|
|06|ArrayAsyncConvenience|3argument inherited WriteAsync routes through gate; same boundary|
|07|BeginEndWrite|Inherited BeginWrite/EndWrite cannot bypass gate; same boundary; complete async result before assertions|
|08|OversizedFirst|65537-byte write rejected whole; delegated calls0, reserved0, fault latched|
|09|CopyExact|MemoryStream source65536, CopyToAsync(sink,8192,token); exact bytes, close successful|
|10|CopyOverflow|source65537, copy faults; actual prefix<=65536, never output-complete; no assumed chunk boundary|
|11|InvalidArguments|null/negative/out-of-range array arguments: no reservation/delegation; subsequent valid1 succeeds|
|12|UnsupportedOperations|Read/ReadAsync, Seek, SetLength, Position setter reject without touching sink; CanRead/CanSeek false|
|13|PreCanceled|already canceled async token: canceled result, no reservation/delegation; failure latch refuses later write|
|14|PartialFault|controlled sink stores3 of requested8 then throws; reserved8, confirmed0, retained3; subsequent write refused|
|15|PartialCancellation|controlled sink stores3 of8, signals entered, cancellation then releases as canceled; reserved8, confirmed0; latch|
|16|SynchronousFault|sink throws on write entry; reserved8 not refunded; confirmed0; real exception preserved|
|17|ConcurrentWrite|hold first8-byte async write; second1 rejected immediately, reservation stays8; release first -> confirmed8|
|18|ConcurrentFlush|held write; Flush and FlushAsync reject; no underlying flush; release then flush succeeds|
|19|ConcurrentClose|held write; Dispose and CloseAsync reject without underlying close; release then CloseAsync succeeds|
|20|CloseFault|controlled flush/dispose failure recorded; Closed=false and OutputAvailable=false, no stable bytes/hash claim|
|21|ClosePending|controlled finalization stays pending; observe unavailable/null length/hash with zero-length sink, no concurrent read; release deterministically, then closure succeeds|
|22|ClosedAndZero|zero-byte write no size increase; successful repeated close is idempotent; writes after close throw; stable empty file valid only after close|
|23|ActualFiles|real factory pair: exact48-byte two CRLF PASS lines plus empty stderr; CopyToAsync8192, CloseAsync, read/hash exact48/0|
|24|ExclusiveCreation|existing a.bin sentinel must remain unchanged on OpenNew failure; fresh b.bin writes/finalizes normally; no path escape|

Controlled sinks use TaskCompletionSource with asynchronous continuations and
explicit release, not sleeps, worker spawning or PowerShell callbacks. Pending
tests must resolve their owned tasks before case end; per awaited test operation
budget2000ms, but worker deadline remains authoritative if a synchronous call
blocks. A test timeout stops all remaining cases. Never abandon a fake task and
proceed to later cases. Case20 has two controlled subchecks (flush and disposal);
case12 covers all exposed read overloads. Each named subcheck must be reported.

- [ ] Implement adapter.cs with one fail-fast interlocked operation gate across
  write/flush/close lifetimes. Validate input and pre-cancellation before capacity
  reservation; invalid arguments do not latch, pre-cancellation does latch.
  Concurrent-call rejection must not poison the already-owned operation.
  Reserve using count > 65536 - reserved (no addition overflow); reservation is
  never refunded. Mark confirmed bytes only on complete underlying write success.
  Latch I/O fault/cancellation/capacity rejection, preserve actual exceptions.
  After latch allow closure but no further writes. No direct handle exposure,
  seeking, padding, queued writes, per-write flush or worker-thread scriptblocks.

- [ ] Implement sync and async Stream entry points consistently, including span,
  memory and disposal. Length/Position getters must not consult a busy sink;
  reject unsupported state inspection, use explicit counters for accounting.
  CloseAsync gates finalization, flushes and disposes the owned sink without a
  second concurrent disposal; Closed only after complete success. For a flush
  failure, still attempt owned disposal once and preserve both errors. A faulted
  or pending finalization never advertises stable output. After successful close
  subsequent close is a no-op. Audit inherited async/APM forwarding as static
  source reasoning plus cases06-07, not assumptions about which overload is used.

- [ ] Write worker.ps1 with InputSHA256/IntentSHA256 arguments and exclusive claim.
  Authenticate runtime/identity/source hashes before any Add-Type. Create compile
  intent, then exactly one Add-Type -LiteralPath with both frozen .cs paths and
  -OutputAssembly D/adapter.dll. No IgnoreWarnings or fallback compiler. Load the
  resulting DLL by exact path using Add-Type -LiteralPath (assembly load, not a
  second source compilation); reject existing type/assembly before compilation.
  Record hash/length and compiler outcome. Run Verification.RunAsync once.
  No original business imports or test groups. On unexpected failure save stop
  and actual partial results; no assembly replacement.

- [ ] Write parent.ps1 by deliberate reuse of saved29 ownership/in-memory transport,
  NOT its business completion predicate. New worker entry only, no Grant/freeze/
  business capture. Entry -InputSHA256 <64hex>; one hidden Process.Start, fixed
  workdir/runtime, ArgumentList, redirects, own intent30 before validation.
  Both CopyToAsync tasks precede one finite WaitForExit. Stop only the associated
  local Process, not lookup/tree killing. Preserve terminal tasks/partial bytes,
  errors and deferred references independently. Success requires worker exit0,
  worker report24/24, no pending ops/errors and intact pre/post identities.

- [ ] Examine disk blocking boundaries before approving the authored implementation:
  all adapter open/write/flush/dispose and compilation are inside the sole worker.
  The parent cannot guarantee OS process termination or filesystem receipt latency;
  record failure if owned exit cannot be confirmed and do not salvage pending files.
  No additional supervisor or child may be silently introduced. This containment
  validates the adapter only; it does NOT establish safe future in-parent disk
  transport. Any unresolved adapter safety issue stops BEFORE compilation/run.

- [ ] Author prepare.ps1: exclusive outputs; data-only hash maps, source inventory,
  expected24case matrix, runtime identities, limits, path allowlist and approval.
  Parse PowerShell without import. Inspect actual C# entry points and control flow
  as text; do not call a compiler for a purported syntax-only check. Freeze
  findings for GateCoverage, ReservationNoRefund, FailureLatch, NoRawHandle,
  ExclusiveFiles, NoBusinessCalls, OneStartOneCompile, WorkerOnlyDiskOperations,
  TerminalOnlyEvidence, FiniteMatrix, NoRetry. C# compilation remains unverified.
  Run preparation once after static review:
  & ./.superpowers/sdd/2026-11-04-runtime-bounded-disk-adapter-v1/prepare.ps1
  Expected all named checks and original hashes match, DiagnosticInvoked=false.
  Unexpected result => preserve stopped drafts, no fix/retry.

- [ ] Independently check frozen hashes/matrix and archive all text/raw receipts,
  excluding DLL/cache/runtime binaries. Decode-check path containment, lengths
  and bytes, normalize only outer JSON to LF. Update result/status/recovery;
  commit explicit files and push existing branch, verify remote==HEAD before Task2.

### Task2: One compile-and-verification observation and preservation

**Consumes:** Task1 input SHA and remotely saved checkpoint.
**Produces:** actual adapter-only outcome and portable evidence; no business-host authority.

- [ ] Invoke exactly once after plan approval and Task1 remote check:
  & ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-bounded-disk-adapter-parent.ps1 -InputSHA256 <actual>
  Save actual initial/polls/final tool returns. Poll only an existing session.
  Never rerun for missing output. Parent full run includes single worker's compile.

- [ ] Worker writes one bounded result with all completed cases and unexecuted cases
  explicitly listed. Print one summary line only: ADAPTER_VERIFIED on24/24 else
  ADAPTER_FAILED. Parent output cap remains65536 per stream. Journal/receipt bounds
  are checked before writes; do not pad/truncate successful data. On unexpected
  assertion/compile/runtime result stop and preserve; do not advance to a host.

- [ ] Independently verify input maps against actual files and parent/worker before/
  after maps; inspect real compile intent/result and assembly hash (if stable).
  Require exactly24 ordered case IDs, expected negative outcomes, no unexpected
  errors, zero pending operations, each real file<=65536 and inventory<=48 files.
  Verify available data bytes and hashes, absent versus unstable paths and actual
  worker exit. Parse journal as strict UTF8/LF, bounded/ordered complete pairs.
  Count actual computation/process/compile consumption. Do not label a missing
  DLL/result as success or retry to recover it.

- [ ] Archive stable files and actual source/commands/results with decode/length/hash/
  containment checks. Do not read unstable pending files; list them as unavailable.
  DLL is excluded with explicit reason/hash/length when known. A fresh computer
  can restore source/progress/evidence but cannot reuse an unavailable local DLL;
  recompilation requires renewed scope and is not automatically authorized.

- [ ] Report ADAPTER_VERIFIED or actual stop, always NOT_ACCEPTED. Separate observed
  local filesystem behavior from deterministic fake-sink cases; no hard OS latency
  promise. Update status/recovery/report, stage explicit list, check raw/index
  blobs and diff, commit/push, confirm remote==HEAD and clean. End regardless.
  A future disk-copy business-host design must resolve finalization ownership and
  receive its own budget; do not prepare or invoke it under this plan.

## Self-review and handoff

Spec adapter requirements map to Task1 implementation and cases01-24; worker
containment is an explicit proposed validation-process use, not business-host
approval. All five review risks have named tests. Future host conditions remain
requirements for a later scope, not omitted tests silently declared passing.
Native execution is retained; no new reviewer/agent allowance. One compile means
no independent executed RED cycle, stated above. File/byte/test/process budgets
are finite. Hard OS I/O latency is not claimed. No implementation/run this turn.

Approval requested: Task1 authoring/preparation and Task2 exactly one adapter-only
verification, ceiling30/9/4/217. No business host or automatic retry.

API reference for compilation/loading syntax, not a local runtime validation:
[Microsoft Add-Type](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.utility/add-type?view=powershell-7.6).
