# Staged integrity implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move the writer-local post-task full audit to the existing parent post-exit pass while keeping final observation acceptance fail-closed.

**Architecture:** New versioned PowerShell entries reuse the unchanged compiled supervisor and adapter. Writer evidence explicitly delegates its after-check; a separately persisted parent audit and strict four-phase trace become mandatory inputs to a new observation verifier. Historical scripts and receipts remain immutable.

**Tech Stack:** Windows PowerShell 7.6.5/Core/x64, existing pinned DLLs, UTF8/LF JSON, memory fixtures and AST checks.

**Spec:** `docs/superpowers/specs/2026-11-05-runtime-staged-integrity-design.md`, approved by user reply 是 after checkpoint `2067616c1082703fa5cbc22d542212e9b7d75bfa`. This records approval without rewriting that historical document. Preserve native inline/no-agent execution for this finite gate.

## Global Constraints

- Current counters 43/32/6/217. Proposed one computation44, at most2starts, zero compilation/snapshots; ceiling44/34/6/217. Execution awaits approval of this plan.
- Keep120000ms execution,5000ms readiness,2000ms shared normal drain/finalization and2000ms shared failure exit-confirmation budgets. Parent synchronous audit has no hard wall deadline.
- Keep parent/writer before checks, producer before/after checks, all input/absence pins, original guards, two groups, cutoff and62events. No deduplication, cached hash, skipped archive, or second parent full pass.
- Normal writer phases: task.returned, role.save.begin, role.save.end, exit.ready. At most one terminal error, maximum5records/384bytes each/1920total. Role identity writer-staged-timing. Old nine-phase helper remains unchanged.
- Preserve raw supervision before audit. Unknown exits/pending resources stop postprocessing. Failed transport never authorizes reading/hashing unavailable output or upgrading transport status.
- No PID adoption/tree kill, disposal of pending resources, compile/install/security changes, automatic retries or full suite. Missing pinned runtime/DLL stops.
- Memory/static tests run in the current shell, without workloads, child processes, assembly loads, background tasks or filesystem timing experiments. One expected RED and one GREEN invocation per test script below. Unexpected results stop for review; do not silently expand the budget.
- All five capability flags false; final project status NOT_ACCEPTED even if this narrow observation passes. No main merge or original-timeout-root-cause/fix claim.
- Temporal guarantee is intentionally changed: parent reads later in a different process. Neither scheme proves an atomic snapshot, absence of transient changes, or loaded-module identity.

## File map

L = `docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2`.
D = `.superpowers/sdd/2026-11-05-runtime-staged-integrity-v1`.
All names below are relative to those exact directories; no historical source edits.

- Create L/runtime-staged-integrity.ps1: definition-only trace helpers and pure final-decision gate.
- Create L/runtime-staged-integrity.tests.ps1: finite memory tests.
- Create L/runtime-staged-integrity-static.tests.ps1: allowed-delta/AST checks.
- Create L/runtime-staged-integrity-role.ps1 and -driver.ps1: versioned writer/delegation and parent audit owner.
- Create L/runtime-staged-integrity-capture.ps1 and -common.ps1: binding-only producer versions.
- Create D/prepare.ps1, verify.ps1, preserve.ps1, progress.md and exclusive immutable records.
- Create L/portable-runtime-staged-integrity-preparation.json and -final.json for recovery; do not depend on ignored D records alone.
- Update docs/roadmap/current-status.md, docs/runbooks/repository-recovery.md and docs/roadmap/2026-11-04-runtime-isolated-writer-result.md at execution handoff.

## Review Focus

1. Successful transport plus missing/unpersistable audit must never produce a passing observation (Task1 G02, Task2 static persistence gate).
2. Stale audit from another input/intent or tampered map reference must fail identity/map validation (Task1 G03, Task2 verifier).
3. Late trace/role publication failure must fail even if an earlier saved role-result says exit0 (Task1 G04, Task2 actual-exit checks).
4. Old nine-phase evidence must not satisfy the new four-phase contract or weaken an old verifier (Task1 T04, Task2 historical immutability).
5. Parent after-check can hang or encounter a missing file: preserve raw transport, withhold final success, never start a substitute audit/worker (Task2 catch and source checks).

## Task1: Pure trace and observation contracts

**Files:** L/runtime-staged-integrity.ps1, L/runtime-staged-integrity.tests.ps1.

**Interfaces:**
- `New-SIState([string]$Token,[string]$InputSHA256,[long]$Frequency)` -> dictionary, Seq=0, Bytes=0, Failed=false, Terminal=false, LastTicks=0.
- `ConvertTo-SIRecord($State,[string]$Phase,[long]$ElapsedTicks)` -> byte[], same eight JSON fields as WPT but new role/four-phase contract, no state mutation.
- `Write-SIRecord($State,[IO.Stream]$Stream,[string]$Phase,[long]$ElapsedTicks)` -> bool; admit before one Write/Flush; latch failure, no retry/refund.
- `Read-SITrace([byte[]]$Bytes,[string]$Token,[string]$InputSHA256)` -> dictionary Status COMPLETE/PARTIAL/ERROR/INVALID, Records, RawTailBase64, Reason, Intervals. Four normal records with no tail required for COMPLETE.
- `Test-SIObservation([System.Collections.IDictionary]$Facts)` -> dictionary Passed(bool), Reasons(string[]). Pure gate; exact required true Boolean facts: TransportVerified, TerminalVerified, BusinessVerified, AuditPersisted, AuditIdentityVerified, AuditMapsVerified, WriterDelegationVerified, RoleVerified, TimingVerified, OutputVerified. Missing/nonboolean/false values reject. The file-reading verifier, not this helper, establishes each fact.

- [ ] Write tests first, retaining raw test receipts. T01 four-phase roundtrip with exact role and eight fields; T02 Int64 limits/384byte line/1920total and refusal of extra emission; T03 nonwritable MemoryStream failure latch/no later bytes; T04 old role/nine-phase input, wrong token/hash/schema/order, duplicate/extra/missing keys, numeric strings/bools/floats, decreasing ticks and changing/nonpositive frequency reject; T05 empty/prefix PARTIAL and truncated tail preserved; T06 malformed UTF8/CRLF/oversize completed line INVALID; T07 terminal error once and no normal records thereafter. Use JsonDocument duplicate-key detection before conversion.
- [ ] Add gate tests G01 all ten exact true booleans passes; G02 each fact independently false/missing/nonboolean rejects (especially AuditPersisted); G03 wrong input/intent/map evidence represented by false corresponding verified fact rejects; G04 transport true with late role failure or incomplete timing rejects. Example assertion: `$facts.AuditPersisted=$false; if((Test-SIObservation $facts).Passed){throw 'audit bypass'}`. Real evidence validation is tested statically in Task2, not assumed from this fixture.
- [ ] Run `& ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-staged-integrity.tests.ps1` once; expect RED missing helper, no side effects.
- [ ] Implement interfaces by versioning the definition-only WPT helper; do not modify it. Preserve strict integer parsing, raw tail, fail latch and overhead-qualified intervals. Implement the conjunctive gate without coercion.
- [ ] Run the same script once; expect `STAGED_INTEGRITY_MEMORY_VERIFIED`, all11named cases passing. Preserve results in progress/preparation archive and checkpoint the explicit new files.

## Task2: Versioned roles, parent audit and frozen verifier

**Files:** remaining new scripts in File map plus Task2 static test.

**Interfaces:** Role takes Role/Scenario/InputSHA256/IntentSHA256; Scenario fixed StagedIntegrity44. Driver takes InputSHA256. Intent Schema1, Computation44, Ceiling44, MaxStarts2, MaxProcesses2, AttemptConsumed=true, fresh Token. Capture retains Grant/FrozenSHA256/Epoch arguments. No compiled interface changes.

Audit record D/integrity-review.json: Schema1, Scenario, InputSHA256, IntentSHA256, Before(map reference), After(map reference or null), State VERIFIED/FAILED, ElapsedMilliseconds(number or null), Error(null or description). References retain Path/Count/Length/SHA256. CreateNew persistence only. Missing record is failure, never inferred success.

- [ ] Write static tests first against exact computation43 baselines. Assert capture differs only in imports/directory; common only new bindings/computation/ceiling; producer body and guards unchanged. Assert writer-before remains, writer-after call/publication removed only in new writer, After=null and AfterVerificationMode=PARENT_POST_EXIT_REQUIRED. Assert four emission sites, error latch, actual nonzero late exit, unchanged old helper/verifiers and DLL hashes. First run `& ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-staged-integrity-static.tests.ps1` expects RED missing new role.
- [ ] Implement new role from phase-timing role. Remove its post-task Map-IW finally audit and associated five scan markers; preserve error handling, before-map, producer route and role publication. Use Task1 trace helper. Record explicit delegation, check trace latch before saving role and after final emission. Failure to save/emit exits nonzero; never rewrite prior saved role evidence.
- [ ] Implement driver from phase-timing driver. Preserve one Supervisor.RunAsync invocation and raw supervision save. Require terminal/no pending/no unconfirmed exits before audit. Wrap the existing single parent Map-IW after-pass in Stopwatch/try-catch; produce VERIFIED only after full map creation, otherwise FAILED with After=null. Persist audit exclusively before any success output. Audit save failure propagates to stop record; transport receipt remains unchanged. No timeout wrapper or second full scan.
- [ ] Implement D/verify.ps1 before preparation: verify input/intent and exclusive audit identity; validate Before/After references, bytes/hashes/counts and complete expected pin coverage/absence results. Reject malformed/missing fields, FAILED, missing map or wrong binding. Require writer explicit delegation/After=null, actual normal owned exits, role exit0/no error, complete new trace, original business checks and exact48/0byte eligible output. Populate all ten Task1 facts from checked evidence; gate produces separate verification.json with ObservationPassed and Status=NOT_ACCEPTED. Never rewrite supervision. For terminal failed transport, diagnostics/audit remain diagnostic only and outputs unavailable; for pending/unknown exits, stop without further inspection. Do not repeat full source hashing in verifier: validate saved map artifacts against the pinned manifest; the driver owns the one post-exit source pass.
- [ ] Implement D/prepare.ps1 and preserve.ps1 by versioning computation43 patterns with explicit historical paths (no broad search/replace across old archive names). Pin old final archive SHA256 `98278C9EBA3AE834220D81E4D880B51F5C4AA52610091B007B90C95E6CEB2E4C`, all original checked sets, new scripts and unchanged DLLs. Reject occupied claims/intent/stops/data paths. Portable archives exclude binaries and unavailable data. Preparation does not rerun tests.
- [ ] Extend static assertions to audit persistence-before-success, exactly one parent after-pass after terminal gate, catch/null failure handling, strict identity/reference validation, all ten fact assignments, output access only in successful eligible branch, actual exit checks, unchanged budgets and one invocation/no retry. Parse all new scripts. Run static tests once; expect `STAGED_INTEGRITY_STATIC_VERIFIED`. Record limitations: these source checks are not real disk-fault injection or proof of hard filesystem deadlines.
- [ ] Run preparation once in the current shell. Expect pinned-file/absence verification and portable preparation archive; decode-check path/length/hash and preserve test receipts. If preparation fails, stop without runtime. Review explicit staged files, checkpoint/push existing development branch and verify remote SHA equals local before launch.

## Task3: One observation and portable handoff

**Files:** immutable D run/audit/review records, portable final archive, result/status/recovery documentation.
**Interfaces:** Consume frozen Task2 entries and prepared InputSHA256; supervision.json stays raw, integrity-review.json reports parent audit, verification.json reports conjunctive observation result.

- [ ] Invoke `& ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-staged-integrity-driver.ps1 -InputSHA256 <prepared hash>` exactly once; consume computation44 even on startup failure, maximum2starts. A pass is not promised. No retry, repair, compile or alternate workload.
- [ ] After confirmed terminal supervision, run the frozen read-only verifier once; expect either all gates satisfied or explicit refusal. Unknown exits/pending resources prohibit postprocessing and further runtime work. If audit has not returned, report incomplete without restarting or claiming acceptance.
- [ ] Preserve raw outcomes and separate transport/audit/observation statuses. Report measured audit time separately from writer drain, trace overhead caveat and changed temporal guarantee. Do not infer original timeout root cause from a successful run.
- [ ] Archive eligible stable records, excluding unavailable outputs/binaries; decode/source-check archive. Update result/status/recovery with actual counters/starts, failures, missing evidence and fresh-computer steps. Necessary source/tests and portable archives must suffice without ignored session data; runtime/cache reinstallation is not authorized here.
- [ ] Review explicit file list, verify staged/raw equality and diff checks, commit/push existing branch, confirm remote SHA equals local and clean status. Stop this gate regardless of outcome; no main merge or additional runtime budget.

## Self-review and handoff

Design coverage: changed audit owner/temporal limits -> Global Constraints/Task2; full pins/producer and protocol -> Task2 static deltas; trace -> Task1/Task2; failure/persistence/identity -> G01–G04 and Task2 verifier/static checks; bounded authority and recovery -> Task3. Review Focus items each name their owning tests/checks. No hidden runtime probes or duplicate parent source scan. The memory gate tests validate decision logic; static checks validate wiring, not runtime fault tolerance under every filesystem failure.

Status PLAN_READY_FOR_REVIEW. Design approved; implementation and proposed44/34/6/217 ceiling await user approval. Native inline method retained, no new agents or broad suite. This plan does not authorize changing the previously consumed computation43.
