# Offline Integrated I1/I2 Repair Implementation Plan — pending approval

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans natively after explicit approval. No implementation agents or per-task reviewers; only the single conditional readonly reviewer below. Steps use checkbox syntax for tracking.

**Goal:** Integrate the verified no-export recording boundary into an actual new offline entry, then reproduce and repair I1/I2 and verify three real profiles without changing frozen history.

**Architecture:** Mechanically migrate the frozen v2 entry/private/test chain to a separately approved v3 namespace. Case modules use the tested no-members export construction, while normal evidence publication calls the same production writer source inside its own private, non-injected module. A separate bounded journal preserves actual failures; it never substitutes for failed evidence.

**Tech Stack:** Windows PowerShell7.6.5/Core/x64; .NET Stream/Process; two in-memory fixture compilations inside the writer hosts only; existing pinned dependencies, no install/download.

**Spec:** The [approved offline design](../specs/2026-10-19-offline-diagnostic-entry-migration-design.md), [original I1/I2 repair plan](2026-10-31-offline-diagnostic-review-repair.md) sections2–4 and exact lifecycle contracts remain normative except for the explicit paths, inventory, recorder additions and budgets below. The [complete review](../../roadmap/2026-10-31-offline-diagnostic-runtime-review.md) defines I1/I2; [actual isolation RED/GREEN](../../roadmap/2026-10-31-offline-record-isolation-repair.md) proves only the recording seam, not this integration. The local integration decisions below remain proposed until approval.

Status: PLANNING_ONLY. Baseline `c99380dbab5adc96138d627b41ac0048891f07d1`, existing `codex/c12-b01-task9-coordinator` linked worktree. Current consent is drafting only: no new source, Prepare, module import, test, compilation, process, agent or review in this turn.

## Global Constraints

- “只涉及离线诊断，不做生产接入”; no product/CI/Docker/default verification chain changes, installs, security changes, main merge or release.
- “默认为 Inspect”; Inspect must not import execution modules, create a runtime root/Session or start a target.
- “Session/Ack 不序列化，不从文件加载，不跨 shell/进程转移”; preserve actual reference/step/budget authority, not a disk Outcome.
- “首次实际失败后，后续逻辑 Read/Parse/Save/Complete/Publish/重新打开在文件访问前拒绝”; no failed sink retry, fallback evidence, resource reopen or success inferred from PASS output.
- “不加入 PID/名称重查、杀树、Job/watchdog、无限等待、启动预热或额外 readiness 等待”.
- All old sources/tests/oracles/manifests/raw/snapshots/reports/portable archives remain byte-identical. Do not edit or replay old drivers, including the now-complete isolation harness.
- Old repair remainder6hosts/7targets/1compile/conditional1review remains FROZEN. Diagnostic1host and isolation2hosts are exhausted. All proposed allowances here are NEW, never refunds or resumes.
- Historical writer-red's first internal ErrorRecord stays UNKNOWN. New actual errors do not reconstruct that missing event. C1/C2 remain open, R2/Minor2 deferred; all five capability flagsfalse until separately scoped evidence supports a decision.

## Review Focus

1. Normal publication must actually use the tested production writer, not bypass I1 with instrumentation: trusted-module/probe/source-binding checks in every host and Task3 review.
2. Test factories/collectors must remain case-local during execution and abort: zero exports, explicit state and four-reference guards in Task1 and every later host.
3. Compound operation/cleanup errors must preserve actual first references/order: both real writer boundaries, all ten fixed vectors in Task1, not just healthy publication.
4. Non-Step values and real error signatures must be validated recursively: the original23schema vectors and exact22/8/15profile observations in Task2/3.
5. Journal/evidence/parent validation failures must stop instead of fabricating continuation: independent predeclared journal, exact parent contracts and no retries in every task.

## 1. Integration choice and exact ownership

Recommended A: new v3 entry/private/test ownership, tested no-export case construction and private normal publisher bound to the actual v3 writer. Cost: another explicit source namespace and new finite verification, but historical source/evidence identities remain intact.

Excluded B: copy only the standalone instrumentation publisher's direct .NET write loop into normal capture and stop invoking the production capture writer. That disconnects a reviewed boundary from its real consumer, so unit success would not establish actual integration. Excluded C: modify frozen v2/test/oracle or resume its stopped allowance. That destroys provenance and is not authorized.

New runtime tree `scripts/offline-diagnostics-v3/` has the same10files as v2: `.gitattributes`, `diagnose-offline.ps1`, `private/catalog.json`, `session.psm1`, `sinks.ps1`, `owned-step.ps1`, `sequence.ps1`, `profiles.ps1`, `observations.ps1`, `write-boundary.ps1` (the last eight filenames are under private/).

New leaf `docs/diagnostics/isolated-fuzz/p2b-integrated-review-repair/` has18files:

- Existing15roles migrated mechanically: `.gitattributes`, `README.md`, `expected-output.json`, `test-helpers.ps1`, `entry.tests.ps1`, `core.tests.ps1`, `prefix.tests.ps1`, `owned.tests.ps1`, `live.tests.ps1`, `capture.ps1`, `driver.ps1`, `record-boundary.tests.ps1`, `observation-contract.tests.ps1`, `fault-stream.cs`, `observation-contract.ps1`.
- `case-module.ps1`: fixed internal WriterPath/CollectorText; dot-source in the active writer-test function, assign the actual module reference; no-members Export-ModuleMember from the initial seed, not a later repair.
- `publisher.psm1`: dedicated normal evidence module, described below; no exported members, no unit factory replacement.
- `recording.tests.ps1`: Start-ODRecorderGuards() returns an in-memory guard reference; Complete-ODRecorderGuards(Guard,Publisher,CaseModules) appends the three fixed recorder groups and complete actual observations.

New packet `.superpowers/sdd/2026-10-31-offline-integrated-review-repair/`; report `docs/roadmap/2026-10-31-offline-integrated-review-repair.md`. Public action/profile IDs, fixed3profile definitions and sole runtime export Invoke-ODProfile remain unchanged; no arbitrary script/path/callback/resume parameters. Repository root is `../..` from the runtime entry directory, `../../..` from private/, and `../../../..` from the leaf. Local `.gitattributes` forces LF for all new sources and portable JSON.

### 1.1 Do not make production I1 a test-only helper

The normal capture publisher imports the fixed v3 `write-boundary.ps1` into its own private module state, along with the one frozen Capture-Failure definition extracted as source from the new capture (do not execute capture during import). Explicitly initialize that module's HostPrimary=null, empty typed HostSecondary and HostOverflow=false. It exports nothing.

Capture retains the actual publisher module reference and invokes Write-ODCaptureRecord inside that module with its fixed internal evidence path and serialized bytes. Therefore normal publication and the ten unit vectors execute the SAME v3 source file/function. Trusted publisher's New-ODRecordStream is the default factory from that same file; it is never injected, patched or selected by test arguments. The profile intent caller still invokes Write-ODProfileIntentRecord from v3/private/write-boundary.ps1. No direct .NET write shortcut replaces either production call site.

The fixed case module loads the same writer and frozen collector, initializes explicit case state, replaces only its own factory with the fixture, and exports nothing. Snapshot all four outer function ScriptBlock references before cases, immediately after creation/replacement, and in a finally observation path if a case aborts. Check each live module's export counts before any case operation; a leak is an unexpected recording failure, not expected I1 RED. Never restore global commands or re-import to conceal contamination.

Capture declares a separate1MiB-bounded direct journal stream before risky module/test imports and fixture compilation, using the proven retained-stream/local-state pattern. Each journal write/flush is attempted once; its first failure latches before any catch could write to it again. Dispose held resources once. Capture startup before journal availability can still fail: preserve parent raw streams/exit and classify UNKNOWN, never invent a missing internal record. Evidence publication is consumed once before opening; actual publisher failure cannot trigger another filename, alternate sink, rebuilt JSON or journal-based success. Parent requires healthy original evidence AND journal. Healthy journal is exactly the ordered JSONL events `begin`, `tests-complete`, `published` with LF delimiters and no extra fields. An actual-error event serializes the original Type/Message/ErrorId/Position/Stack for diagnosis, never a successful journal verdict.

### 1.2 Three recorder groups in every actual host

Append these three PASS groups after that grant's original groups, never as extra hosts:

1. `RecorderIdentityIsolation`: all four outer references still identical, no case mutation escaped; complete before/created/replaced/finally observations are checked where cases occur.
2. `RecorderPrivateState`: normal publisher has zero exported functions/aliases/variables/cmdlets and explicit clean state; all ten writer-case modules likewise export0, with their actual error state kept separate. Other grants have no writer-case modules; do not count an empty list as dynamic case coverage.
3. `RecorderProductionBinding`: actual publisher writer/factory ScriptBlock.File and source SHA bind v3/private/write-boundary.ps1; inside that trusted module perform one healthy probe write of exact UTF8 `{"Recorder":"v3","Healthy":true}` at the fixed capture/<Grant>/publisher-probe.json, after case operations. Parent checks original bytes, default factory binding, module identity/state and then the actual one-shot final evidence. The probe is not evidence fallback or a substitute for final publication.

New expected-output.json must contain literal complete group facts/observation schemas for these roles and all original vectors before seed freeze. Actual source/module identities and references are measured in the host, not inferred from equal printed text. Parent constrains dynamic module IDs by actual role-consistency; all error Type/Message/ErrorId and frozen valid call sites remain independently bound. Cases do not receive a publisher reference or its state.

## 2. Frozen inputs, changes and retained I1/I2 contract

Before seed, all28sources must exist, retaining the original I1 ordering defect and Step-only I2 validator; isolation/journal corrections are already applied. The two true REDs must reveal those reviewed defects, not missing functions/compilation/record failures. From the first freeze, only two production edit windows exist:

- writer-green: v3/private/write-boundary.ps1 only, first-error-before-secondary and original reference rethrow for BOTH normal capture writer and profile intent writer; cleanup must not replace the earlier ErrorRecord.
- repair-green: new leaf/observation-contract.ps1 only, complete recursive observations/errors validation, exact mismatch signature `OD: observation contract mismatch` (RuntimeException/Message/FQID).

All other27sources, tests, capture/publisher/builder/recorder, oracle and parent remain frozen. Preserve all ten original writer vectors and counts/messages/reference checks; the real-IO delegating Stream fixture remains explicitly a controlled fixture, not an OS late-write/real FileStream.Dispose-fault claim. Compile it once in each writer host, never before/elsewhere.

Retain the original three historical positive schema frames plus20single mutations:23groups. Use the exact vector definitions and fixed type/message templates in original plan sections2/3, including four Step positive controls; arbitrary exceptions do not count as valid rejection. Original source-bound historical frames are read-only data, not scripts/modules to execute and not an oracle generated from measurements.

The same Assert-ODProfileObservations function must validate ALL actual ProfileReturn observations in the new driver, including ordered22core/8owned/15live inventories, full recursive View/Sinks/IORoot/Error/summary fields/types/unknown-key rejection and deterministic error signatures. Preserve the independent literal models and new core error assertions before seed. No fallback Step-only validation. Adapt contexts and paths to v3/new namespace while preserving explicit historical contexts for the three saved positives; no global context switch or broader dynamic whitelist.

### 2.1 Exact inventory

Inherit all2394entries from the verified isolation GREEN manifest, SHA `93B1D18EE6C4FF944FEB45AA8CE8E91C7011BDD3C0374FFA9A6587C2AAF1D447`. If the new schema uses VerificationPath/ExpectedState, map old Path→VerificationPath and Present explicitly in the NEW manifest only; preserve every actual path/hash and consistent duplicate identity, never edit the old manifest.

Add exactly88Present mappings: isolation packet47actual files; isolation leaf9files (8sources+portable); isolation execution report1; this plan1; new packet approved spec/plan snapshots2; new sources28. Total2482Present. Verify prior portable SHA `83FE6E002281837B4E841F334A591B6D936EE2D90D107441529AF4D123DFFD2D` against all47actual original files/6directories, not reconstructed originals.

Three single Prepare phases seed/writer-green/repair-green snapshot28sources each, total84. Each later Prepare requires the actual previous manifest SHA and exactly its single permitted changed source; other2481pins unchanged. Runtime session/profile code must bind the NEW schema/2482count/repair-green/source paths, not old2317/v2/old packet constants. The new portable export is not a runtime input.

Before first Prepare, inventory permits precisely progress.md and explicitly named author static-result receipts, recorded in the ledger; no arbitrary file count assumptions or silently ignored unknown files. Actual AST/LF/JSON/full source-binding checks and all original tool objects are preserved. An unexpected real preparation/static/runtime/record failure stops rather than becoming an excuse for a new generation or rerun.

## 3. New budget request and exact sequence

Proposed NEW **7capture hosts + 7controlled PS7 targets + 2in-memory compilations + conditional1fresh readonly review**. Native implementation, no implementation agents. Original allowances remain frozen/exhausted even if this plan succeeds.

| Grant / order | Hosts | Targets | Exact complete output and capture exit |
| --- | ---: | ---: | --- |
| writer-red | 1 | 0 | Original4FAIL/6PASS then3recorderPASS =4FAIL/9PASS; exit1 |
| contract-red | 1 | 0 | Original16FAIL/7PASS then3recorderPASS =16FAIL/10PASS; exit1 |
| writer-green | 1 | 0 | Original10PASS then3recorderPASS =13PASS; exit0 |
| contract-green | 1 | 0 | Original23PASS then3recorderPASS =26PASS; exit0 |
| core | 1 | 0 | Original22PASS then3recorderPASS =25PASS; exit0;22private observations fully checked |
| lifecycle | 1 | 3 | Original14PASS then3recorderPASS =17PASS; exit0;8private observations fully checked |
| continuation | 1 | 4 | Original13PASS then3recorderPASS =16PASS, original12seamPASS additionally; exit0;15private observations fully checked |

First two grants use actual seed SHA; writer-green uses its new freeze; remaining four use repair-green. Both RED signatures require complete original observations, no extra exception/host error, empty stderr, valid evidence/journal, exact exit/drain and intact pins. Writer/contract test abort, compilation error, recorder group failure or unknown evidence is not expected RED.

Use unchanged engine/SMA paths/hashes and DESKTOP-H4SVCLU\Lenovo/adminFalse from original repair plan section4. Each noncontinuation host120000ms, continuation150000ms; internal sequence140000ms/cleanup reserve4000ms; exact owned exit2000ms/shared stream drain2000ms. Original hold45000ms, normal/receipt2000ms, timeout200ms and seam110000ms remain unchanged. Original Session≤32/sink≤32/Ack≤2/observations≤32/distinct secondaries≤8 bounds remain; capture's separate recorder observations have a separately frozen bounded schema, not arbitrary extra private-session rows.

Only after all seven grants and source/record counts are exact may the one NEW fresh-context readonly reviewer inspect the complete new entry→profiles→writers/validator/recorder chain, original I1/I2, actual results and unproved cases. Use the most capable then-available approved reviewer model with explicit metadata, separate context; no execution/import/network/write/Git/agents by reviewer. No review now, second review, automatic fix pass or test budget is implied. Important/Critical/incomplete review stops; preserve findings and unfinished checkpoint. Even a clean review supports only a proposed narrow new-v3 finding disposition, not rewritten old history/full C12/main acceptance.

First unexpected actual failure freezes ALL remaining grants/review. Preserve full original errors/streams/tool-return chunks before parsing; only wait the actual returned session ID. No retry/refreeze/source patch/new packet/budget refund, no unknown→PASS reconstruction, no old test replays. Explicitly scoped pure read-only display mistakes are not runtime failures, but genuine record/verification failures cannot be relabelled display errors.

## Task 1 — actual recording integration and I1 RED→GREEN

**Files:** before seed create the exact10runtime+18leaf sources above; after RED edit v3/private/write-boundary.ps1 only.
**Interfaces:** consumes immutable v2 sources and verified no-export mechanism; produces same source-bound normal capture/profile writer calls, private publisher/case modules and exact13-group writer verification.

- [ ] Record explicit approval of this plan/paths/new budget before creating code. Verify existing linked branch/clean state, all2394inherited entries and prior byte archive; manually create only this plan's ledger/explicit static receipts. No baseline/suite or historical import.
- [ ] Write complete tests/literal oracle first; mechanically adapt v3 paths/counts, integrate fixed case builder/private normal publisher/three recorder groups and bounded independent journal. Statically trace normal capture and profile intent to the SAME tested production file; verify Inspect remains import/IO/target-free. Freeze all28inputs once with Prepare seed, expected2482Present/28snapshots/no changes.
- [ ] RunOne writer-red once, then contract-red once, with actual seed SHA; expected table signatures and complete recorder/actual-error contracts. If either differs, preserve and stop; do not modify frozen assertions to match results.
- [ ] Repair both production writer orderings in the one allowed file: locally preserve first ErrorRecord; after cleanup register first→secondary once, retain existing primary, deduplicate references, rethrow original first only. No finally replacement, reopen/rewrite/dispose retry or factory API expansion.
- [ ] Prepare writer-green once with actual previous SHA; exactly1source change/28new snapshots. RunOne writer-green once; expected13PASS and healthy actual source-bound default publisher/probe/final,0targets. Preserve all actual records and necessary source checkpoint; no I2 acceptance claim yet.

## Task 2 — I2 full contract consumed by actual parent

**Files:** only new leaf/observation-contract.ps1 changes; all tests/oracle/capture/driver and production writer stay fixed.
**Interfaces:** Assert-ODProfileObservations([object[]]Rows,[string]Profile,[Collections.IDictionary]Context) success emits nothing, mismatch only the exact declared RuntimeException; the actual profile driver always calls it.

- [ ] Read exact saved contract-red, original23vectors and full independent literal models, without rerunning RED or importing old runtime. Confirm failures are the16unrejected mutations rather than arbitrary crashes.
- [ ] Implement full ordered/recursive22/8/15observation validation, reject missing/extra/duplicate/wrong type/error signatures and scope dynamic fields only to original explicit rules. Preserve exact Step controls and all non-Step IORoot/View/summary behavior.
- [ ] Prepare repair-green once with actual writer-green SHA; exactly1validator change/28new snapshots, other2481pins unchanged. RunOne contract-green once,26PASS/0targets/healthy records and3recorder groups. Preserve checkpoint; schema-unit success is not actual profile success.

## Task 3 — actual new public profiles, sole review and recoverable handoff

**Files:** runtime/test sources now immutable; only actual packet/review/new execution+operator/recovery documents and portable export are produced.
**Interfaces:** consumes actual repair-green and fixed new public entry; returns complete observed profile/capture evidence, not permission tokens or historical closures.

- [ ] RunOne core once25PASS, verifying actual default/explicit Inspect,14rejections, private16original/core conditions and full22observations through repaired parent validator; record source bindings and3recorder groups. No targets.
- [ ] RunOne lifecycle once17PASS/3targets: exact nonassociation/prestart, one receipt failure, finite ownership/drain, normal and timeout tuples; full8observations and3recorder groups. Preserve actual raw and parent errors before interpretation.
- [ ] RunOne continuation once16PASS plus original12seamPASS/4targets: independent missing/wrong Ack negative while Active/steps/budget remain valid, six invalid requests/replay, real C-output failure without contaminating B, actual A1seam continuation once; full15observations and3recorder groups. No extra FINAL/authority test host.
- [ ] Only all seven grants/7targets/2compilations/2482pins/84snapshots exact: checkpoint all necessary files, then the sole fresh readonly review above. Save complete findings/declined-to-judge items and author dispositions; Important/unknown stops without code changes, retest or second review.
- [ ] Export every actual packet file/directory/raw zero byte/partial/unknown/snapshot/tool object/approval/review as portable base64; compare exact original bytes. Preserve old records untouched. Record actual consumed budgets, what passed versus unproved OS write/close/host-death/classifier faults, and reviewer limits.
- [ ] Update current-status/repository-recovery and a new `docs/runbooks/offline-diagnostics-v3.md` (do not edit frozen older runbooks). Explicit necessary source/doc/portable list only; verify staged bytes, commit/push existing dev branch and verify remote SHA. No main merge, deletion, credentials/cache/raw commit or restored Session/Ack.

## Self-review and approval gate

I1 both real callers and ten compound/single vectors→Task1; independent tested recording mechanism versus actual normal production binding→all three recorder groups and source-bound final publication in every host; I2 complete values/errors and same validator actual consumption→Task2/3; original lifecycle/Ack constraints and reviewer limitations→Task3.28×3=84snapshots;2394+47+9+1+1+2+28=2482pins;4unit+3profile=7hosts,3+4=7targets,2writer-host compilations,conditional1review. All five review risks have named checks and retained limitations.

Drafting checks read all2394existing final pins and matched them, confirmed47original packet/9leaf files and the recorded portable hash; no dynamic verification, runtime source or reviewer was used. This plan proposes integration, not a claim that a trusted publisher, compound fixture or full validator already works in v3.

Please review the proposed Native integration and explicitly approve its NEW7hosts/7targets/2in-memory compilations/conditional1readonly-review ceiling before execution. A planning “yes” is not execution consent; old6/7/1/review stays frozen regardless of new approval.
