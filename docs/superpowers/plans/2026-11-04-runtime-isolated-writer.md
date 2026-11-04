# Isolated Writer Synthetic Gate Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Verify a bounded framed relay and sibling-process writer lifecycle with synthetic data only.
**Architecture:** One new assembly supplies protocol, relay, writer, supervisor policy and test routines, referencing the existing frozen adapter DLL. A single owned compilation worker runs18in-process protocol/policy checks; then9fixed scenarios launch writer/producer siblings directly from the supervisor.
**Tech Stack:** Pinned PowerShell7.6.5/Core/x64, existing .NET/C# compiler, redirected standard streams, SHA256, bounded JSON receipts.
**Spec:** ../specs/2026-11-04-runtime-isolated-writer-design.md

Status: PENDING_PLAN_APPROVAL / NO_IMPLEMENTATION_OR_EXECUTION.
Design168a05f821e03fd4facc0b118b98aca82fde125b approved by user with 是.
DesignSHA2567A603CB6DE009F2234D3B4EFCF692C224958A8F09E7784D4F1B86AD115358B70.
The design's pending label is historical. This plan still requires approval.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator, existing
  codex/c12-b01-task9-coordinator branch, github.com/win202569/Talenro.git only.
- L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2.
  D=.superpowers/sdd/2026-11-04-runtime-isolated-writer-v1.
- Current30/9/4/217. PROPOSED computations31..40 (10), at most18new child starts,
  one source compilation5, no snapshots: ceiling40/27/5/217. Do not infer authority
  from design approval. The table below is the complete launch allowance.
- Computation31: one owned compiler/check worker, one source compile. Computations
  32..40: nine synthetic scenarios,17starts maximum. Existing supervisor is not
  an extra spawned process. No child starts grandchildren; no separate shell,
  external compiler, service, listener, network or business-host process.
- Missing/mismatched old adapter DLL is a preflight blocker, not permission to
  recompile it. Expected50176bytes/SHA256
  4C1470B94B3DEAA87635F1375A05161BE0E1F4C4D2CF65595214A8E39CDBCCDA at
  .superpowers/sdd/2026-11-04-runtime-bounded-disk-adapter-v1/adapter.dll.
  Pin old source/compile receipt and final archive
  BD10DDFB8D2523ACDBCB4BBD3177776DC5F283387ECF76A6A911B995B78294E9.
- Every computation consumes its exclusive intent before validation/start.
  Record compile intent before Add-Type. Count actual starts/compile attempts,
  including failed attempts. No retries, recompiles, omitted failures or spending
  unused slots on alternative scenarios. First UNEXPECTED result stops the gate.
- Execution120000ms before setup; readiness5000ms within execution. On normal
  producer exit, one shared2000ms relay/finalization window; on failure, one
  shared2000ms owned-exit confirmation window. No per-process fresh windows.
  These bound programmed waits, not OS calls, scheduling, startup or persistence.
- Data per channel<=65536bytes; read buffers2x8192 plus one8208-byte send buffer.
  Single in-flight serialized send, no unbounded queue. DATA1..8192, at most16
  per channel and35frames total. Coalesce full8192chunks except final EOF partial.
- Writer control record<=4096UTF8bytes, stdout<=8192, stderr<=4096. Exact LF JSON,
  no BOM; one READY and one terminal COMPLETE/ERROR. Producer stderr is DATA,
  never confused with writer diagnostics. No PowerShell worker-thread callbacks.
- Each scenario owns at most2data files (131072bytes); all9<=18files/1179648bytes.
  Synthetic capacity-overflow producer emits exactly65537stdout bytes (one over
  the admission limit) and0stderr; no file may exceed65536.
- Source/evidence/old stops unchanged. No timeout extension, PID lookup/adoption,
  tree kill, job object, unsafe-code/PInvoke, dependency install or security change.
  No business code/imports/62-event host or full suite. All five flags false,
  NOT_ACCEPTED; no main merge.
- Tests are authored first. Proposed finite exception: one compiled test run,
  no separate executable RED build, mutation run, reviewer or broad suite.
  No executed-TDD claim. Static checks do not prove runtime behavior.

## Review Focus

1. Malformed lengths/EOF and duplicate frames: U01-09 exercise actual decoder;
   it must reject before allocating/writing beyond the bounds.
2. Receipt forgery/duplicates and success before real EOF: U10-13 and S09 must
   reject false success even with exit0.
3. Backpressure and unfair stderr scheduling: U14 plus S01 exercise both channels
   with single-buffer ownership and exact bytes.
4. Cancellation/late success/termination uncertainty: U15-18 and S03-07 verify
   failure latch, shared clocks and no success after timeout.
5. File existence mistaken for stable output: U17 and S04-09 require unavailable
   state unless both channels finalized, COMPLETE validated and exits confirmed.

## File map and interfaces

Create in L:
- runtime-isolated-writer-protocol.cs: frame codec, receipt validator, bounds.
- runtime-isolated-writer-runtime.cs: pure-.NET relay, sibling ownership/state,
  writer and synthetic producer roles; no compilation in runtime.
- runtime-isolated-writer.tests.cs:18fixed in-process protocol/policy tests.
- runtime-isolated-writer-role.ps1: fixed compiler/writer/producer entry dispatcher.
- runtime-isolated-writer-driver.ps1: preparation binding, compiler ownership,
  sequential scenario ledger, observation/preservation orchestration.
- portable-runtime-isolated-writer-preparation.json,
  portable-runtime-isolated-writer-compiled.json and -final.json.

D contains prepare.ps1, preserve.ps1, progress.md, approval.json, input.json,
static-checks.json, actual preparation tool/verification records; compile/ and
scenarios/S01..S09/. Per directory: intent, claim, bounded journal, result, stop
on unexpected failure, actual tool records, stdout/stderr when safely available.
Each scenario writer output lives only under its own data/stdout.bin and stderr.bin.
Generated isolated-writer.dll is local only. Exclude DLLs, runtime/cache binaries
from archives; retain source/hash/length/compile command. Binary post-size check
<=1MiB is not a hard disk quota. Scenario journal<=128records/65536bytes,
result<=262144bytes. Four identity maps separately capped2MiB each, hash-referenced.
Driver gate ledger<=262144bytes, no rolling deletion of evidence.
Four maps per computation: supervisor.before/after and worker.before/after for31;
supervisor.before/after and writer.before/after for32..40. A terminated writer may
lack its after-map: record absence, never synthesize it. Expected failure cases
require unchanged actual pins and all available authenticated maps; a missing
writer after-map is allowed only when the pinned fault/termination prevents it.
Update result docs/roadmap/2026-11-04-runtime-isolated-writer-result.md,
current-status.md and repository-recovery.md at each checkpoint/handoff.

New namespace Talenro.Diagnostics.IsolatedWriterV1:
- FrameCodec.Encode(Frame) -> byte[]; ReadAsync(Stream,ProtocolState,CancellationToken)
  -> Task<Frame>; bounded exact-read loop, validates header before payload allocate.
- ProtocolState.Accept(Frame): monotonic sequence, channel totals/END/CLOSE state.
- ReceiptValidator.Accept(ReadOnlySpan<byte> line,ReceiptState): exact schema/token/
  input hash/role/order; Reject trailing records/data after terminal.
- Relay.RunAsync(Stream stdout,Stream stderr,Stream writerInput,CancellationToken)
  -> Task<RelayResult>: counters, incremental sent hashes, two real EOF flags.
- Writer.RunAsync(WriterOptions) -> Task<int>; Producer.RunAsync(ProducerOptions)
  -> Task<int>. Roles receive frozen scenario ID, token, input hash and fixed root.
- Supervisor.RunAsync(ScenarioOptions) -> Task<ScenarioResult>: owns both Process
  objects directly, no filesystem capture reads/writes inside active supervision.
- SupervisorPolicy.Step(State,Observation,long ticks) -> Decision: deterministic
  failure latch/deadlines/actions. Runtime must actually use this policy.
- ProtocolTests.RunAsync() -> Task<UnitReport>, exactly U01..U18.
- DTOs use explicit scalar fields: IDs, status/reason, token/hash, timestamps,
  read/admitted/sent counts, task states/errors, process associations/exits,
  output availability, nullable lengths/hashes, pending resources and false flags.
  Bounds errors are typed ProtocolViolation, ReceiptViolation, CapacityExceeded,
  ReadyTimeout, DrainTimeout, ProducerFailed, WriterExited, ExitUnconfirmed.

## Fixed verification ledger

Compiler worker31 runs U01..U18 in-process after its sole compilation; no role
processes within it. Each test records all stated assertions and no extra inputs.

|ID|Behavior and independent expected result|
|---|---|
|U01|Golden16-byte header: TLW1/version1/DATA/channel1/seq1/len1 encodes exact literal bytes; round-trip payload0x2A|
|U02|Wrong magic,version,reserved byte,channel,type each rejected before payload read (five literal vectors)|
|U03|Lengths0 and8193 on DATA rejected;8192 accepted; END/CLOSE require0|
|U04|Truncated header15bytes and payload declared2/actual1 reject at EOF|
|U05|Sequence0,duplicate1,gap3 reject after valid sequence1|
|U06|DATA after END,duplicate END,early CLOSE,duplicate CLOSE reject; valid two END+CLOSE accepted|
|U07|Per-channel65536 accepted; next1 rejected without sink delegation; counters unchanged for rejected frame|
|U08|17thDATA on one channel and36thframe reject; use small payloads to distinguish frame cap from byte cap|
|U09|CLOSE without actual EOF not complete; extra byte after CLOSE rejected; actual EOF permits finalize|
|U10|READY with wrong token/hash/role/schema or unknown field rejected (five fixed vectors)|
|U11|Duplicate READY/terminal,terminal-before-READY and line after terminal rejected|
|U12|4097-byte line,8193-byte control total,missing LF,invalid UTF8/BOM rejected before unbounded accumulation|
|U13|COMPLETE length/hash mismatch,exit0 without COMPLETE,COMPLETE without both exit0 reject; good pair accepted|
|U14|Deterministic two controlled input streams: partial reads coalesce to8192, final tails exact; both ready => alternating channel preference; at most one send, no buffer reuse while held|
|U15|Simulated clock: readiness and execution deadlines consume same origin; normal shared drain never resets per channel; failure exit budget shared across processes|
|U16|Pending cancellation task: failure stays latched, no forced concurrent dispose, late COMPLETE cannot change failure|
|U17|Existing zero-length but unclosed/pending output => unavailable/null hash/length; confirmed finalized successful output may be empty|
|U18|Simulated Start associates then throws: cleanup targets only that Process association; unconfirmed exit forbids success or next scenario; no PID/name lookup|

U08 tests protocol-state counters directly for the global cap where earlier
per-channel/state rejection prevents constructing36otherwise-valid frames.
U18 tests injectable process facade/policy; it does NOT prove actual OS behavior.
Fake streams/controllers are test utilities, not product fault-repair code.

Each S row consumes one new computation. Fixed synthetic pattern byte(index%251);
S01 stdout8193bytes/stderr48bytes, S02 both empty. Mode is input-pinned and permitted
only in this synthetic gate, never user-selected arbitrary commands.
Unless a row specifies otherwise, its producer emits48stdout bytes (the two
original CRLF PASS lines as synthetic fixture bytes) and0stderr, then exits0.
S01 stderr uses48pattern bytes, not those text lines. Writer-controlled holds
are deterministic test branches before READY, inside first write, or before
close completion; they do not claim to emulate every OS I/O hang.

|Scenario|Computation|max starts|Injected condition|Expected observation|
|---|---:|---:|---|---|
|S01|32|2|normal dual-stream|two exit0, valid COMPLETE, exact8193/48bytes, all tasks done|
|S02|33|2|normal empty|two exit0, real EOF/close, exact0/0bytes, not missing files|
|S03|34|1|writer holds before READY/open completion|ReadyTimeout, producer never starts, writer owned exit after termination|
|S04|35|2|writer holds before first disk write|producer emits8193/48 then exits0; DrainTimeout, writer terminated, outputs unavailable|
|S05|36|2|writer holds during finalization|producer emits48/0 then exits0; DrainTimeout, no COMPLETE, unavailable|
|S06|37|2|writer exits23 after first DATA|WriterExited/transport failure retained, both owned exits; not success|
|S07|38|2|producer exits23 without payload|ProducerFailed, writer stopped, no COMPLETE-based success|
|S08|39|2|producer65537/0|relay CapacityExceeded, forwarded/file totals<=65536; both owned exits|
|S09|40|2|writer finalizes but suppresses COMPLETE and exits0|MissingReceipt, no success despite files/exit0|

Compiler worker1 + scenario starts17 =18. Totals at ceiling40/27/5/217.
S03-05 controlled holds never self-release; supervisor must terminate its own
writer. S04 producer writes small payload and exits; tests cannot depend on a
particular OS pipe-buffer size: if it stays blocked, ExecutionTimeout is an
unexpected result and stops, not silently relabeled DrainTimeout.
S06 may observe pipe failure before exit23; record both, accept either first
reason only with the pinned writer exit23 and confirmed exits. No other scenario
allows alternative failure reasons. Expected induced failure counts as that case
PASS only when its actual assertions match; gate PASS is not transport SUCCESS.

## Task1: Author and statically freeze the gate

**Consumes:** approved design and this plan, saved30 evidence/adapter.
**Produces:** source-only preparation checkpoint; no runtime validation.

- [ ] Verify isolated branch/origin, absence of new paths, existing DLL source/
  receipt/hash match. Read saved parent-v2, adapter and verification as data.
  Bind all previous source/history/absence pins plus old final archive and this
  spec/plan/approval. Existing inputs/claims/stops unchanged.
- [ ] Write tests.cs U01..U18 first with literal expectations above; fake I/O only
  below real codec/relay/policy. Record mental mutation targets: bypassed cap,
  reused buffer, reset deadline, accepted late COMPLETE, inferred empty output.
  Do not execute test stubs or compile an extra RED assembly.
- [ ] Implement protocol.cs surfaces; validate little-endian16-byte header before
  bounded payload allocation. One sequence across channels. EOF after CLOSE must
  be observed, never inferred from declaration. Validate exact control schema.
- [ ] Implement runtime.cs relay with two reusable8192-byte input buffers,
  one8208-byte encoder buffer and one in-flight write. Await terminal results
  before reuse; fair alternate selection when both ready. Accumulate source read
  counters and incremental hashes of successful DATA writes, not rejected bytes.
  Source overflow detection may read one bounded extra chunk but forwards none
  beyond65536. Do not treat CopyToAsync's buffer argument as a hard allocation cap.
- [ ] Implement writer role: verify token/input/assembly and fixed paths, open
  adapter sinks, emit READY, consume valid frames to matching adapter. After
  CLOSE+actual EOF, close both once, compute actual closed-file hashes, emit
  COMPLETE and exit0. Fault => bounded ERROR where possible and nonzero exit;
  never silently finish after capacity/protocol fault. Disk I/O lives here.
- [ ] Implement synthetic producer modes/table; use raw standard streams for
  deterministic bytes, no text encoding side effects or descendants. Final
  controlled failure modes are explicit pinned branches, not random sleeps.
- [ ] Implement SupervisorPolicy and runtime supervisor. Start writer once; begin
  bounded control readers immediately. Valid READY gates producer start. Start
  both producer readers immediately. Track process association even when Start
  throws. Never block on Result/GetResult before terminal; no capture file access.
  On normal producer exit run one shared2000ms drain/finalization window; on
  failure latch/cancel/kill associated processes and share2000ms exit confirmation.
  On unconfirmed exit stop the entire gate, retain references and never start
  the next scenario. No process-tree operations or restart.
- [ ] Separate SupervisionOutcome from EvidencePersisted. Raw output tasks that
  remain pending must stay unavailable; no concurrent dispose/read/hash. Only
  after confirmed exits may the independent verifier inspect SUCCESS files.
  Failed-case output remains unavailable unless a separately approved salvage
  follows; do not inspect it just to fill an archive inventory.
- [ ] Implement role.ps1 compiler: authenticate fixed runtime/non-admin state and
  claims, record compile intent, Add-Type exactly once for3new .cs files referencing
  the frozen adapter DLL, load the emitted DLL by explicit path, run18unit tests
  once. C# compilation itself must not spawn an external compiler process.
  Retain full actual warnings/errors; no IgnoreWarnings, fallback or retry.
- [ ] Implement driver.ps1 preparation/compile/scenario stages with explicit
  -Stage Compile|Scenarios and mandatory InputSHA256. Compile stage creates31intent,
  supervises one compiler worker using bounded memory output and120000/2000/2000
  waits. Scenarios stage refuses missing/failed compile receipt or existing gate
  intent/any scenario intent; creates a gate claim and runs table sequentially,
  each with fresh intent/token. Driver checkpoints every completed scenario before
  allowing the next; recording error stops. No per-case manual rerun entry.
- [ ] Author prepare.ps1 and static-checks.json: PowerShell ParseFile only, no
  imports/compiles; validate18U/9S names, exact budget sum, fixed paths, private
  original Process references, no business imports/launches, one compile site,
  bounded queues/receipts and unavailable-state branches. C# code is uncompiled.
  Use safe enumerated parser-error formatting, never empty-array.Message access.
  Run data-only preparation once; first unexpected result stops.
- [ ] Archive original drafts, preparation call/result, pins and explicit check
  findings; decode-check containment/hash/bytes. Update status/result/recovery,
  commit explicit files and push. Require remote==HEAD before Task2. No binaries.

## Task2: Compile once and run the fixed synthetic gate

**Consumes:** Task1 input SHA; existing approved source/runtime/adapter hashes.
**Produces:** actual compiler/unit/scenario evidence, not business-host permission.

- [ ] Invoke once:
  & ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-isolated-writer-driver.ps1 -Stage Compile -InputSHA256 <actual>
  Expected: one31intent/oneworker/compile5;18unit checks PASS, no pending resources.
  Capture actual initial/polls/final; poll sessions, never replay. Unexpected
  result stops scope. Source files stay frozen after this invocation.
- [ ] Independently read receipts/check original identities and unit outcomes.
  Verify generated DLL hash/size and reference identity. Archive compiled checkpoint
  with DLL metadata only, push and confirm remote equality before scenarios.
  No fresh test/compile to compensate for missing receipts.
- [ ] Invoke once:
  & ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-isolated-writer-driver.ps1 -Stage Scenarios -InputSHA256 <actual>
  Expected: S01..S09 match their exact success/failure assertions, actual starts
  <=17, no orphan/unknown exit and no pending parent resources at scenario handoff.
  Intent claims prevent accidental replay. Unexpected outcome stops remaining
  rows and marks them NOT_RUN. Do not repair/retry or use their unused allowance.
- [ ] Independent read-only verification: actual claim/counters <=40/27/5/217;
  all source/input identities and available referenced maps match (missing fault-
  path maps handled under File map rules);18unit and9scenario
  rows exact, per-case times/ownership/control order/data counters valid. For
  S01/S02 only, compare stable real files to deterministic expected bytes/hash.
  Failure cases require null availability fields and correct reasons, not empty
  substitute data. Check every started process's recorded owned-exit result;
  evidence records are observations, not a new process-lookup probe.
- [ ] Validate journal bounds/strict UTF8/LF/sequence and balanced case records.
  Compare raw control bytes to parsed receipts and protocol state, truncation
  flags and pending status; missing/contradictory evidence blocks gate success.
  Do not read mutating/unavailable writer files. Preserve actual partial records.
- [ ] Archive only stable sources/evidence with path/length/hash/decode checks.
  List excluded DLLs and unavailable/absent output separately. No deletes or
  synthetic completion. Update report/status/recovery and actual used counters,
  stage explicit list, check raw/index equality, commit/push, verify remote==HEAD
  and clean worktree. End on either gate completion or unexpected stop.
  Gate success remains NOT_ACCEPTED; no business-prefix preparation/execution.

## Self-review and handoff

All design protocol/lifecycle requirements map to Task1 surfaces and U01-18/S01-09.
Malformed protocol and receipt branches are exercised in-process through actual
decoder/validator; selected sibling IPC/hang paths receive actual process tests.
Not every malformed vector is an end-to-end process test; this limitation is
explicit, not full fault coverage. Supervisor abrupt death and OS kernel hangs
remain outside guarantees. Expected failure does not mean actual file success.

Native inline execution retained; this plan requests no reviewer process/agent.
Single-compile exception omits separately executed RED cycle. Tests/source/receipts
and budgets are frozen before execution. Existing adapter assembly reused unchanged;
missing assembly or any need for another compiler requires a new scope.

Approval requested: Task1 plus Task2, exactly the fixed31..40 ledger, at most18
new child starts and one compilation, total ceiling40/27/5/217. No business host,
automatic retry, full-suite acceptance or merge. No implementation/run this turn.
