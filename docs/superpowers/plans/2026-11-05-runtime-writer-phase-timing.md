# Writer exit-phase timing implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Capture bounded writer exit-phase evidence in one observation without relaxing checks or claiming a timeout fix.

**Architecture:** A versioned PowerShell writer role emits small timing records through its existing stderr pipe; the unchanged supervisor captures them. A definition-only helper implements encoding, emission and strict parsing. Business capture, ownership and compiled protocol remain unchanged apart from new invocation bindings.

**Tech Stack:** Windows PowerShell7.6.5/Core/x64, pinned existing DLLs, UTF8/LF JSON, in-memory streams and PowerShell AST checks.

**Spec:** `docs/superpowers/specs/2026-11-05-runtime-writer-phase-timing-design.md`, approved by user reply 是 after checkpointf46d4236. Design approval authorizes this plan, not its execution. Preserve the native inline/no-agent/no-broad-suite method of this finite diagnostic work.

## Global Constraints

- Current42/30/6/217; proposed one computation43, at most2starts, zero compilation/snapshots; ceiling43/32/6/217. No runtime invocation before plan approval.
- Definition-only helper tests operate on memory streams/fixture bytes in the current shell; no assembly load, business entry, benchmark, process, background task or disk-output experiment. They do not consume runtime computation slots. Retain RED/GREEN results; unexpected results stop for review, not automatic repair/retry.
- Keep120000ms execution,5000ms readiness,2000ms shared normal drain/finalization, and2000ms shared failure exit confirmation. These are not hard OS wall-time guarantees.
- Preserve all original source/absence pins, full checks, business groups/cutoff/62events, original error/cleanup paths and actual48/0-byte business output requirements.
- Normal timing sequence9records; maximum10records/384bytes each including LF/3840bytes total. Existing4096byte writer-stderr cap unchanged.
- No source/evidence edits for consumed computation42; new namespace and bindings only. Missing pinned DLL stops, never compiles implicitly.
- Original Process associations only, no PID adoption/tree kill; no disposal of pending resources. Unknown exits/pending resources stop further runtime work. Failure outputs unavailable/null, never read/hash/salvage.
- All five flags false; NOT_ACCEPTED. No full suite, main merge, install, changed safety settings or original-timeout-fix claim.

## File map

L = `docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2`; D = `.superpowers/sdd/2026-11-05-runtime-writer-phase-timing-v1`.

- Create L/runtime-writer-phase-timing.ps1: definition-only encoder, bounded emitter and parser.
- Create L/runtime-writer-phase-timing.tests.ps1: finite in-memory tests, no workload/child invocation.
- Create L/runtime-writer-phase-timing-static.tests.ps1: AST and exact allowed-delta checks.
- Create L/runtime-writer-phase-timing-role.ps1: version of business-role with after-task instrumentation only.
- Create L/runtime-writer-phase-timing-driver.ps1: version of business-driver, single computation43.
- Create L/runtime-writer-phase-timing-capture.ps1 and common.ps1: binding-only versions of business capture/common.
- Create D/prepare.ps1, verify.ps1, preserve.ps1, progress.md and immutable run records. Portable archives: L/portable-runtime-writer-phase-timing-preparation.json and final.json.
- Update current-status.md, repository-recovery.md and 2026-11-04-runtime-isolated-writer-result.md at preparation/final handoff. Leave approved design and this plan immutable after pinning; progress lives in D.

## Review Focus

1. Writer diagnostics mistaken for producer stderr: keep separate acceptance rules and exact business0stderr.
2. Killed writer leaves partial JSON: interpret only valid complete prefix, preserve tail raw, never infer success.
3. Instrumentation fails inside cleanup: latch failure without skipping original checks or replacing original errors.
4. Wrong identity/order/numeric types or duplicate JSON keys: reject rather than coerce or silently overwrite.
5. Trace timing misreported as COMPLETE/exit time or pure hash cost: report only local phase boundaries with overhead caveats.

## Task1: Bounded trace helpers and finite memory checks

**Interfaces (definition-only script, no top-level I/O):**
- `New-WPTState([string]$Token,[string]$InputSHA256,[long]$Frequency)` returns dictionary with identity, Seq=0, Bytes=0, Failed=false, Terminal=false, LastTicks=0.
- `ConvertTo-WPTRecord($State,[string]$Phase,[long]$ElapsedTicks)` returns byte[] with exactly the eight spec fields and LF; validates identity, numbers, allowed phase, next sequence/size without mutating state.
- `Write-WPTRecord($State,[IO.Stream]$Stream,[string]$Phase,[long]$ElapsedTicks)` returns bool, admits budget before I/O, performs one Write and Flush, latches Failed on construction/I/O failure and disables later emission; no throw to caller. Valid error is terminal. Failed writes are not refunded or retried.
- `Read-WPTTrace([byte[]]$Bytes,[string]$Token,[string]$InputSHA256)` returns dictionary Status=COMPLETE/PARTIAL/ERROR/INVALID, Records, RawTailBase64, Reason, Intervals. No file access or modification of supplied bytes. Exactly nine normal records yields COMPLETE; valid shorter prefix yields PARTIAL; final error yields ERROR; malformed completed record yields INVALID. Empty input is PARTIAL. Structural parsing rejects duplicate keys using JsonDocument property enumeration before conversion.

- [ ] Write named tests first: T01 nine-record exact roundtrip; T02 maximum Int64 ticks/frequency with36char token/64char hash fits384; T03 no11th write and total<=3840; T04 invalid token/hash/phase/negative ticks rejected; T05 nonwritable MemoryStream latches failure and subsequent emission leaves Seq/Bytes unchanged; T06 empty/valid prefix is PARTIAL; T07 truncated final record retains raw tail without manufacturing event; T08 malformed UTF8/CRLF/oversize line or total is INVALID; T09 duplicate/extra/missing fields and wrong JSON types rejected; T10 wrong token/hash/role/schema rejected; T11 missing/repeated/reordered sequence/phase, decreasing ticks or changed/nonpositive frequency rejected; T12 terminal error allowed once/no later records, COMPLETE requires nine phases and no tail.
- [ ] Run `& ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-writer-phase-timing.tests.ps1`. Expected RED: missing helper, no workload side effects. Save actual result.
- [ ] Implement the interfaces, preserving strict integer JSON types (not strings/bools/floats), sequence1..10 and spec phase order. Permit error as the terminal record after any valid normal prefix; no normal records after error. T03 rejection may be order/cap rejection but must prove no extra output.
- [ ] Run the same finite tests once. Expected GREEN: `WRITER_PHASE_MEMORY_TESTS_VERIFIED` with12named cases. No extra runtime probes. Commit necessary helper/tests with progress evidence in the preparation archive before launch.

## Task2: Versioned role and single-run preparation

**Interfaces:** Role accepts existing Role/Scenario/InputSHA256/IntentSHA256, Scenario fixed WriterPhase43. Driver takes InputSHA256 and uses existing ScenarioOptions/Supervisor.RunAsync exactly once. Capture retains original Grant/FrozenSHA256/Epoch parameters. Intent Schema1/Computation43/Ceiling43/MaxStarts2/MaxProcesses2/AttemptConsumed=true with fresh token.

- [ ] Write static tests first: only new directory/common-import changes in capture; only43/ceiling43/two-process intent values in common; role preserves original finally/identity/maps/business invocation; driver has one supervisor invocation and no retry/compile/alternate start. Assert instrumentation restricted to Writer after-task/after-map, never producer or before-map. Check all nine phase call sites and trace-failure propagation through exit after publication. Run expecting missing new role/driver RED.
- [ ] Create versioned entries. Start the local Stopwatch immediately before Writer.RunAsync. Use its ticks only after terminal GetResult returns. Pass actual Console.OpenStandardError stream to Write-WPTRecord; do not close it early to manufacture EOF. Instrument pins.begin/end around original Assert-HTInputs, map.save.begin/end around original Save-IW and map.meta.end after original metadata return, then role.save.begin/end and exit.ready. Keep original work order and checked data identical.
- [ ] Ensure error path attempts at most one error record if trace state remains usable, retains original error, executes original finally checks and exits nonzero. Check the latch before role-result publication and again after final trace emission; a late failure must set process exit nonzero without rewriting the already saved role-result. No trace error may bypass original finally work.
- [ ] Build D preparation/verification/preservation scripts by versioning computation42 patterns. Implement the verifier contract specified in Task3 now, before the final static run; Task3 consumes this frozen verifier. Pin original source maps, helper identities, both DLLs and prior final archive SHA25693F36D09D25F18D4ED46D01B1B809822F5D5A7357AC4C67AC690FD4701927DAD; never include excluded failed data paths. Preserve original runtime/account guards. Reject occupied intent/claim/stop paths before starts.
- [ ] Run static check once, including Task3 verifier assertions; expected `WRITER_PHASE_STATIC_VERIFIED`. Parse all new scripts. Unexpected failure stops preparation; do not change frozen prior files or run workload to debug it.
- [ ] Prepare once in current shell, record file/hash/absence checks, test records and approved ledger; no extra test invocation inside preparation (parse/hash existing test results instead). Decode/source-check portable preparation archive. Commit/push explicit files on codex/c12-b01-task9-coordinator and confirm remote SHA equals local before runtime.

## Task3: One observation, separate timing verdict and recovery

**Interfaces:** D/supervision.json remains the unchanged transport result. D/timing-review.json records parser outcome and qualified intervals; D/verification.json records conjunctive observation verdict. No review overwrites supervision or original evidence.

- [ ] Verifier contract implemented/frozen in Task2: after confirmed terminal supervision, decode only captured DiagnosticBase64 and call Read-WPTTrace. Preserve raw bytes and parser result even if transport failed. Do not attempt reading capture files for failed or uncertain supervision. On unknown exits/pending resources record that stop only, with no further interpretation/file inspection.
- [ ] In successful transport branch retain every computation42 success check except writer-diagnostics-empty; replace that with strict complete nine-phase timing validation. Business stderr remains0, stdout exact48bytes, producer/child/schema/cutoff/62events/maps/claims/EOF/hashes still required. A trace failure makes observation fail without rewriting transport success.
- [ ] Verify the separation statically before freezing: failure branch cannot call stable output inspection; diagnostics parsed only from saved supervisor bytes; business stderr assertion remains0; interval labels never claim COMPLETE emission or exit time. Use source/AST checks in Task2's static run, not an additional business test.
- [ ] Invoke `& ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-writer-phase-timing-driver.ps1 -InputSHA256 <prepared hash>` exactly once. Expected result is a terminal recorded observation, not a promised pass; consume computation43 even if startup fails. Maximum2starts; no retries, repair, wider suite or compile.
- [ ] Run the read-only verifier once. Successful observation requires all transport/business/timing conditions; failed observation preserves its valid trace prefix or parser rejection. Report last observed boundary, only paired local-clock intervals, and unknowns. No cross-pipe ordering assumptions or pure hash-duration claim.
- [ ] Archive stable eligible files and raw diagnostics; exclude binaries and unavailable data. Decode-check path/length/hash. Update result/status/recovery with actual starts/counters, trace completeness and NOT_ACCEPTED scope. Commit/push explicit necessary files, verify remote equality/clean worktree, then stop this finite gate regardless of outcome.

## Self-review / handoff

Spec coverage: unchanged guards/budgets -> Global Constraints/Task2; bounded emission and errors -> Task1/Task2; partial trace and strict success -> Task1/Task3; evidence and authority -> Task2/Task3. Review Focus1/3 covered by role/verifier static checks and T05/T12; Focus2/4 by T06–T12; Focus5 by Task3 labels and report requirements. No changes to compiled code or protocol; no hidden process tests.

Status PLAN_READY_FOR_REVIEW. User-approved design is recorded here without rewriting its historical pending label. Execution awaits plan approval. Native finite inline method retained; no extra agents, broad regression runs or cleanup of evidence. This plan does not diagnose or fix the original timeout merely by adding instrumentation.
