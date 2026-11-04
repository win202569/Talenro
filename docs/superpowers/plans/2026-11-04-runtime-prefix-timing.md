# Runtime Prefix Timing Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Obtain bounded stage timings through the first real BaselineCore assertion without reopening a stopped route.
**Architecture:** Separate parent owns one hidden diagnostic process. Child loads allowlisted, hash-bound definitions with timing-only insertions and runs the original prefix once; immutable evidence survives timeout.
**Tech Stack:** Windows PowerShell, existing pinned pwsh, .NET monotonic clock/asynchronous stream reads, JSON/SHA256.
**Spec:** ../specs/2026-11-04-runtime-prefix-timing-design.md

Status: PENDING_EXECUTION_APPROVAL. User approved design72f722ec with 是.
The original design bytes remain unchanged. This plan grants no execution itself.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; existing codex/c12-b01-task9-coordinator branch only.
- L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2.
- P=.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner; H=.superpowers/sdd/2026-10-31-offline-diagnostic-runtime-migration.
- D=.superpowers/sdd/2026-11-04-runtime-prefix-timing-v1, a fresh fixed directory; do not reuse if occupied.
- Current26aux/5hosts/3compiles/217snapshots. One new computation27, at most one diagnostic process counted as host6; maximum27/6/3/217. No targets/compiles/snapshots/reviewers or additional test runs.
- All31operational sources, existing tests/inputs/receipts/snapshots and15stops remain byte-identical. Do not execute driver/capture/full suite, cache/optimize checks, modify timeouts, install or access network from diagnostic code.
- Execution cap120000ms; owned-exit2000ms; stream-drain2000ms. Only the invocation-local Process object can be terminated; no PID/name lookup, process tree kill or adoption.
- First unexpected preparation/identity/recording result stops; no retry or repair of a consumed attempt. End after one diagnostic even on success. NOT_ACCEPTED throughout.
- Native PowerShell, no generic skill test wrappers, full-suite runs, workspace cleanup or extra reviewers: the explicit finite diagnostic scope overrides those defaults.

## Review Focus

- Instrumentation must not change extracted statements, pipeline output or source-model content; inspect reversible insertion map in Task1.
- Input drift during extraction/model construction must not become a timing success; pre/post identities and original pin checks remain in Tasks1–2.
- Timeout or journal failure must leave an identifiable incomplete interval, not a fabricated end event; schema/status rules in Tasks1–2.
- Redirected streams must not deadlock or grow without bound; explicit polling/drain limits in Task1.
- Repeated/cross-invocation execution must not adopt an intent or process; exclusive parent/child claims in Task1, inspected before Task2.

## Task 1: Prepare immutable helpers and inspect without execution

**Files:** Create L/runtime-prefix-timing-parent.ps1, L/runtime-prefix-timing-child.ps1,
L/runtime-prefix-timing-common.ps1; D/approval.json,input.json,instrumentation.json;
L/portable-runtime-prefix-timing-v1-preparation.json. No operational edits.

**Interfaces:**
- Parent entry `-InputSHA256 <64 uppercase hex>`; repo/D/runtime locations are fixed, not caller-selected.
- Child entry `-InputSHA256 <hash> -IntentSHA256 <hash>`; validates fixed intent/input and uses an exclusive child.claim.json. It cannot start another process or call parent.
- Common definitions: `Write-PTEvent(Stage,Kind,Profile,Source)` emits no pipeline output; `Get-PTDefinitions(SourcePath,Names)` returns only exact allowlisted function text; `Write-PTOnce(Path,Value)` uses CreateNew.
- input schema1: ApprovedDesign/PlanSHA256, FrozenSHA256, Engine/SMA identities, SourceSHA256(31), PreservedSHA256, HistoricalEvidenceSHA256(3), HelperSHA256(3), DefinitionSHA256, InstrumentationSHA256, Limits. Never bind a future/self-generated receipt.

- [ ] Confirm branch/origin and absence of D plus all three helper files and preparation archive. Read current candidate and manifest; verify all2344pin states,31sources,15stops and original final archive11EFF0959FA2AA181A44FF1D38E9FAA5C59C6CAA6B7C2F7AF56CEE11EE51C5F0. Record all immutable input hashes. A conflict stops, not an overwrite.
- [ ] Bind manifest186FE2C2B44A31A98EA5406109AC888A625FD1B0061C5338D0382268015D9BBA and the three evidence identities from the approved investigation. Bind existing pwshSHA362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139 and SMASHA2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B. Existing runtime path is the one in driver.ps1, not discovered/replaced by another install.
- [ ] Implement the exact allowlist below. Parse original source as data, require exactly one FunctionDefinitionAst for each selected name. Reject unexpected call dependencies; do not evaluate top-level code. Capture original definition text/hashes before instrumentation.

From source-model.ps1: Get-ODEpochManifestName, Get-ODPinnedSHA256,
Get-ODSourceOwner, Get-ODFailureExtent, Read-ODSourceModel, New-ODFixedContext,
New-ODHistoricalContext. From observation-contract.ps1: OC-Need, OC-Keys,
OC-Integer, OC-Expand, OC-SourceError, OC-Match, Assert-ODProfileObservations.
No runtime/unit constructors, driver, suite, report or process functions are extractable.

- [ ] Instrument only Get-ODPinnedSHA256 and Read-ODSourceModel; leave other12definitions byte-identical. Insert calls around exact AST statement extents, never global text replacement. For pin helper: bracket pin-row lookup and Get-FileHash assignment. For source model: bracket ParseFile, and the block building functions/calls/faults/dynamicCalls. Retain pre/post pin calls, variable names, order and returned values. Event calls use unique script-scoped PT names and no success-stream output. Store each insertion offset/text and original/transformed hashes. Removing insertions in reverse offset order must reproduce the exact original text; parser must report zero errors and no duplicate functions.
- [ ] Implement child initialization: require original7.6.5/Core/x64 and fixed non-admin identity from capture.ps1, verify common/helper/input hashes, exact parent intent schema/identity, then CreateNew child.claim.json and journal.jsonl. No fallback runtime or privileges. Record initialization and input-verification intervals; load only the validated definitions.
- [ ] Implement prefix in exact order: manifest hash, read text, ConvertFrom-Json; then runtime-green/core, lifecycle/owned, continuation/live, each hash/read/JSON, require one canonical ProfileReturn, append actual observations and call New-ODHistoricalContext. Preserve the suite's checks/order. After all three, copy frames[0] through ConvertTo-Json -Depth60/ConvertFrom-Json -AsHashtable, then call actual Assert-ODProfileObservations for core/context0 once. No later groups. Stop if emitted output is nonempty; do not manufacture LS-Result or suite pass receipts.
- [ ] Bracket top-level stages `manifest.hash/read/json`, `evidence.hash/read/json/extract`, `context`, `baseline.copy`, `baseline.assert`; nested stages `pin.lookup`, `pin.hash`, `source.parse`, `source.model`. Profile is empty or one of the three fixed profiles; Source is a bound repository-relative path or empty. Preserve a stage stack; an exception logs an error for the active interval, not a successful end.
- [ ] Implement journal schema: Schema1, Seq(long starting1), EventId(long), ParentId(long/null), Stage, Kind(begin/end/error), Profile, Source, ElapsedTicks(long), Frequency(long). Begin allocates EventId; end/error reuses it; Seq increases, ticks never decrease. Time is Stopwatch ticks since child start; parent duration is separate. Encode UTF8 no BOM, one compact JSON per LF, write+flush each event. Maximum4096events and1048576bytes, checked before write; exceed -> recording failure, do not silently omit and continue. Flush cost is included in enclosing intervals; do not sum nested inclusive durations.
- [ ] Child exit0 only if actual assertion returns silently and input post-hashes match; exit1 for actual assertion rejection; exit2 for setup/drift/recording/other exception. result.json includes status `PREFIX_COMPLETED`, `ASSERTION_REJECTED` or `INCOMPLETE`, actual error details, last event, pre/post maps and all five capability flags false. No product PASS label. If killed, missing child result stays missing; parent records timeout separately.
- [ ] Implement parent: refuse any D intent/claim/journal/streams/result/stop, CreateNew intent before validation with computation27/ceiling27, MaxProcesses1, random invocation token and input/helper hashes. Then verify immutable inputs/runtime. Failed validation consumes attempt27 but does not count a process as started. Never reuse existing intent; child claim prevents duplicate direct child execution.
- [ ] Start one local Process hidden, UseShellExecute=false, redirect both streams, no shell command string. Stopwatch starts before output-file creation/Start. Use .NET ReadAsync byte buffers8192bytes, one pending read per stream; parent polls tasks and process in at most10ms wait slices, concurrently draining both. Do not put PowerShell callbacks on worker threads or compile helper code. Each output capped65536bytes: preserve capped prefix, record overflow and terminate only owned process. Polling deadline remains120000ms; do not extend for startup/output. No unbounded CopyToAsync-to-file path.
- [ ] Reuse original owned-process association/exit discipline without calling old driver: reflection on this Process.Id confirms association, no lookup; timeout/error/overflow triggers Kill only if not exited. Bound WaitForExit2000ms and simultaneous drain grace2000ms total; unfinished tasks/streams are reported, not retried. Record parent final once even after partial setup, with Started/PID/OwnedExit/StreamsComplete/ExitCode/TimedOut/Overflow/errors and preflight/execution/cleanup durations separately. Stop file is CreateNew on any incomplete outcome. Keep child's result distinct from parent's final.
- [ ] Perform static inspection only: parse all helpers, verify allowlist/dependency closure, unique anchors and reversible transformations, literal limits/path containment, bounded stream loop, exclusive claims, no invocation of prohibited entries, exact prefix order. Do NOT execute extracted functions, smoke-test child, warm up or simulate a successful result. Any gap stops before invocation and is reported.
- [ ] Save approval referring to actual user execution approval, immutable input and insertion map. Archive exact D files plus three helpers into preparation archive; decode-verify SHA/length/path containment. Update status/recovery, commit explicit files, push existing branch, confirm remote equals HEAD. No attempt consumed by successful preparation alone.

## Task 2: One diagnostic and evidence-only interpretation

**Files:** D/intent.json,child.claim.json,journal.jsonl,stdout.bin,stderr.bin,
result.json (child, if produced),final.json (parent),stop.json (if needed),
tool-record.json; L/portable-runtime-prefix-timing-v1-final.json;
docs/roadmap/2026-11-04-runtime-prefix-timing-result.md and status/recovery.

**Consumes:** Task1 immutable input hash and remote preparation checkpoint.
**Produces:** actual bounded process evidence; no product acceptance or repair authority.

- [ ] Invoke exactly once from the fixed worktree:
  `& ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-prefix-timing-parent.ps1 -InputSHA256 <actual frozen Task1 hash>`.
  Preserve full tool input/initial/polls/final, including partial output. Do not invoke child directly. Failure before Start still forbids retry.
- [ ] Read actual terminal status; validate saved journal bytes, sequence/stack ordering/tick monotonicity and bounds without rerunning functions. Keep a torn final JSONL line as raw evidence and mark incomplete. Exit0 alone is insufficient: require silent real assertion completion, matching identities, valid complete events, no overflow/timeout/cleanup error before saying this prefix completed.
- [ ] Report measured inclusive durations for top-level intervals and separate nested breakdowns. For interruption, report last completed and unmatched begin plus parent elapsed limit; do not infer CPU/I/O cause or exact blocked instruction. Compare static work counts only as expectations, not fabricated completed work. Do not extrapolate full52group success or claim optimization needed from one sample.
- [ ] Independently recheck immutable inputs31sources/15stops/original final archive and bound historical files; retain original pre/post records even if checks fail. Stop whether diagnosis is conclusive or not. Record computation27 consumed; add host6 only if actual process started. No fix/retry/new suite.
- [ ] Preserve all actual D originals and helper bytes in exclusive final archive, decode-check hashes/lengths/empty files/directories. Update report/status/recovery and counters; commit explicit necessary files, push and verify remote HEAD. No overwrite/delete/merge. Future steps need a new decision based on this evidence.

## Plan self-review and approval

Scope/identities/budget/ownership/evidence requirements map to Task1 and actual
one-shot interpretation/preservation to Task2. The two tasks share only frozen
input/helper identities and actual artifacts; no future receipt is pinned in
itself. No unsupported performance claim or benchmark repetition is required.
Timing events can perturb performance; this is an instrumented prefix, not an
equivalent full-host replay. No diagnostic code has been written or executed.

Await user approval of this plan before implementation or budget consumption.
