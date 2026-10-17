# Offline Raw Capture Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for the selected Native execution. Steps use checkbox (`- [ ]`) syntax for tracking. Method and finite scope are approved; do not repeat completed or consumed steps.

**Goal:** Build only the synthetic, in-memory byte-transfer coordinator from the approved design, preserving bytes and conservative lifecycle/authority semantics.

**Architecture:** A new module privately owns two exact MemoryStreams, follows bounded supplied read/exit/tick actions, and delegates the derived event trace to the existing public synthetic record constructor/verifier. It does not implement real pipes, processes, independent supervision or automatic persistence.

**Tech Stack:** Windows PowerShell 5.1 and PowerShell 7, existing .NET MemoryStream/SHA256 and repository PowerShell contracts; no installation, Add-Type compilation or external dependencies.

**Spec:** [Approved raw capture/outer control design](../specs/2026-10-17-raw-capture-outer-control-design.md), approved by the user's explicit yes after `92af8224ca24f9be1eac6d9650c3a9da9aa061fc`.

Status: **APPROVED FOR NATIVE EXECUTION.** User selected1 after5915bc732a658973b368b40a681250388911854e, explicitly choosing parent implementation, up to10 synthetic tests and one conditional readonly review. No real runtime/Stage B authority. Progress is recorded in the execution report/checked steps below; historical proposal wording describes the approved scope, not extra budget.

Latest follow-up: **I1 CLOSED; SEPARATELY APPROVED M1 TEST FOLLOW-UP VERIFIED.**
After81afee10 the user approved tests-only literal event expectations and
alteration-rejection cases. Four NEW hosts completed: targeted and full30/30
on both engines, with product/dependencies unchanged. See the
[M1 record](../../roadmap/2026-10-17-offline-raw-capture-event-oracle.md).
Parent bounded evidence addresses M1; no independent rereview/acceptance follows.
New4/4 are spent; original budgets/history below are not reopened.

Earlier I1 disposition: **I1 CLOSED BY SEPARATELY APPROVED BOUNDED REPAIR; M1 DEFERRED AT THAT CHECKPOINT.**
After f8657855 the user approved two clones, one regression and five NEW hosts.
The [repair record](../../roadmap/2026-10-17-offline-raw-capture-empty-ownership-fix.md)
records precise PS7 RED, targeted GREEN on both engines and full29/29 each,
no timeout/retry or dependency change. New5/5 spent, no new independent reviewer.
Original10/review1, slot5 false verdict/exception and all historical evidence
below remain unchanged. No broader acceptance or further budget is implied.

Disposition at original review: **ALL10 SLOTS AND REVIEW1/1 CONSUMED; IMPORTANT OWNERSHIP DEFECT OPEN, INCREMENT UNFINISHED.**
See [execution/readonly assessment](../../roadmap/2026-10-17-offline-raw-capture-execution.md).
The user approved the [specific slot5 admission exception and remaining-budget continuation](../../roadmap/2026-10-17-offline-raw-capture-continuation.md) after3243d55b. All five remaining calls matched;
original slot5 ExpectedResult=false is preserved. Both first and final new-suite
engines had28PASS; unchanged contract regression110 each. The
[completed review](../../roadmap/2026-10-17-offline-raw-capture-readonly-review.md)
found Important1 empty-array ownership and Minor1 incomplete independent event
expectations. Important1 blocks completion; Minor1 is disclosed/deferred.
No test/review allowance remains; no product/test repair or additional run followed review.

## Global Constraints

- Worktree: `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`; existing branch `codex/c12-b01-task9-coordinator`; inspect clean state and applicable instructions before execution. Reuse this isolated worktree, do not create another by default.
- Only synthetic input; exact `SourceMode='synthetic'`; no live-mode flag, external Stream, callback, arbitrary command, PID/handle or process API in the product module.
- stdout/stderr each <=1048576 bytes; one read Count in1..65536; Actions actual one-dimensional object[],0..4096; all derived chunk bytes <=2097152.
- Virtual deadline120000ms, equality means timeout; no real clock, sleep, timer, concurrent tasks or watchdog in the new module.
- OriginVerified=false, Acceptance=false, RuntimeApproved=false, LiveActionsPerformed=false and CleanupStatus='unknown'; stream disposal is not process cleanup proof.
- Preserve old contract/simulator/classifier/package/store source and tests byte-for-byte; no schema migration or synthetic archive acceptance expansion.
- Stage B, historical safety disposition, real runtime packet, I2/I3/Task10/11/full C12/main/release remain separate. No old runs replayed and no spent/frozen budget reused.
- Filesystem writes during tests limited to saved test receipts outside the product module; new source itself is pure in-memory. No archives, native fixtures, security queries/settings, Go/Docker/services/network workloads or installs.

## Review Focus

These five implied edge classes are explicitly pinned in Task1 rather than left to generic happy-path checks:

1. A short nonzero read must not fabricate EOF or expose unused buffer tail: ShortRead.
2. A semantic stop must not read the offending source or lose the admitted prefix: DeadlineEdges/AfterEOF/Trailing and stop mutation.
3. Caller mutation after return must not alter Actions or record snapshots: CopyOwnership and ownership mutation.
4. Preflight must inspect passive descriptors before values and reject decorated arrays/byte arrays without executing members: PassiveProperties/ArrayAndByteBounds.
5. A complete lifecycle with refused output classification, or a clock-inconsistent declaration, must not become accepted runtime: RefusedClassification/InvalidHeader/ResultShapeFlags.

## File Structure and Interfaces

- Create `docs/diagnostics/isolated-fuzz/offline-raw-capture.ps1`: only public entry `New-OfflineRawCaptureRecord([object]$Scenario)`; private helpers use `RawCapture` prefix. Import only `capture-adapter-contract.ps1`, transitively its existing pure dependencies. Do not import isolation/store/native helpers.
- Create `docs/diagnostics/isolated-fuzz/offline-raw-capture.tests.ps1`: self-contained fixtures/assertions; parameter `Group` exactly all/ownership-mutation/stop-mutation, default all. Group changes selected assertions, never product behavior.
- Create `docs/roadmap/2026-10-17-offline-raw-capture-execution.md` during implementation: pins, per-slot results, limitations, review and recovery.
- Update current-status.md, repository-recovery.md and this plan's checked steps truthfully. No unrelated files staged.
- Ignored receipts/controller, if needed: `.superpowers/sdd/2026-10-17-offline-raw-capture/`. No raw receipt or controller committed; portable tests must run without that ignored directory.

Scenario fields/types are exactly spec§5, including Header's seven fields and owned byte inputs. Return is exactly ten fields: Schema,SourceMode,Created,Reason,Record,Actions,FailureActionIndex,OriginVerified,Acceptance,RuntimeApproved. Reasons and null/false behavior are spec§6, not new enums.

Consumes `New-SyntheticCaptureRecord -Packet <object>` and `Test-SyntheticCaptureRecord -Record <object>`. Record remains the existing20-field object, not a wrapper extension. Preserve canonical argument bytes and UTC strings without normalization. The new Actions snapshot belongs to the returned coordinator result, not the existing archive format.

## Approved Finite Verification — consumed, no renewal

At most10 fresh child test hosts, sequential,120000ms each, no automatic retry/repair/timeout extension. Expected TDD/mutation failures count as consumed calls and permit continuation only for their exact frozen assertion. Any other failure stops all remaining slots and the review. No separate exploratory imports/test invocations outside this list; static parsing, hashing and receipt inspection do not invoke product code.

|Slot|Engine|Entry/Group|Required result|
|---|---|---|---|
|1|PS7|new test/all, no implementation|exit1; only expected ASSERTION FAILED: RawCaptureApiRequired|
|2|PS5.1|same frozen test/all, still no implementation|same expected RED|
|3|PS7|new test/all, implementation present|exit0; exact28 ordered labels|
|4|PS5.1|same frozen files/all|same GREEN|
|5|PS7|new test/ownership-mutation, one specified temporary mutation|exit1 at RawCaptureCopyOwnership only|
|6|PS7|new test/stop-mutation, restored baseline then one specified mutation|exit1 at RawCaptureDeadlineEdges only|
|7|PS7|restored new test/all|exit0; exact28 ordered labels|
|8|PS5.1|same restored files/all|same final GREEN|
|9|PS7|unchanged capture-adapter-contract.tests.ps1|exit0; existing110 ordered PASS occurrences|
|10|PS5.1|same unchanged contract test|same regression GREEN|

No test quota is transferable. If a slot times out, save output and timeout/cleanup facts and stop; no claim all processes naturally ended. Infrastructure must hold the exact launched test-host handle; never kill an arbitrary PID or reuse an old diagnostic controller/identity. A later exception/repair needs separate user direction.

These are synthetic test hosts, NOT native collection or Stage B cases. Use current inherited token, record actual identity/admin role; do not request elevation or change roles to pass. Permission/ownership refusal stops that call, not a justification to relaunch outside the sandbox. Git preservation permissions are separate from testing authority.

Host commands are `& $engine -NoProfile -NonInteractive -File $testPath -Group $group`, and omit -Group for the unchanged contract suite. Resolve and record x64 engine paths before the first launch; current PS7 resolves to `C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe`, PS5 to `C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe`. Missing engine on another computer stops preparation, no download/fallback. Use a process controller for the stated timeout, not an unbounded direct invocation. PS5 receives the existing fixed Windows inbox module path policy, without execution-policy changes.

Freeze test source/hash and ordered labels before slot1. Freeze implementation and all12 prior dependency pins before slot3. Preserve stdout/stderr separately with lengths/hashes, actual engine/version/role, exit, elapsed and timeout; never rewrite receipts. For GREEN stderr accept empty or exclusively structurally recognized CLIXML Completed progress objects under the already approved progress policy: DTD/external resolution disabled, no object deserialization, reject Error/unknown/trailing data. RED must match its exact assertion and expected engine error envelope; any unrelated stderr record fails. Do not suppress progress/errors or retroactively relax the policy.

Original prerequisite was all10 results matching the table. The explicit later approval adopts the readonly slot5 assessment as that specific admission exception while retaining its false verdict; all remaining results and pins must still match. Those conditions now hold: ONE fresh independent read-only reviewer is allowed, scoped to the new increment, relevant dependencies and saved evidence. No reviewer tests, edits, network or further delegation. No automatic second review/fix cycle and no whole-branch/main approval.

### Task1: Implement and verify the offline coordinator

**Files:** create the two source/test files listed above; do not modify dependencies.

**Interfaces:** consumes Scenario/spec§5 and the existing constructor/verifier; produces the exact ten-field result above. Fixture helpers live only in the test: `New-RawTestScenario` returns fixed Header plus owned supplied bytes/Actions; `New-RawReadAction(Channel,Count,ElapsedMs)` produces the exact read shape; `Assert-Raw(Value,Label)` throws `ASSERTION FAILED: <Label>` or prints one PASS; `Assert-RawRecord(Result)` checks ten-field wrapper,20-field verified Record and conservative flags. Helpers must not invoke the new API to construct expected bytes/events/hashes.

- [x] **Step1 — write tests before the module.** Import only when its file exists, then assert API presence so RED fails for the intended reason, not a missing-file parser error. Group all emits exactly the following28 labels in order; each label means all assertions in that case passed, not one independent requirement:

```text
RawCaptureApiRequired
RawCaptureNormalBytes
RawCaptureUtf8Split
RawCaptureUtf16Split
RawCaptureShortRead
RawCaptureEmptyEOF
RawCaptureEarlyExit
RawCaptureUnknownExit
RawCaptureUnreadTail
RawCaptureDeadlineEdges
RawCaptureEqualTime
RawCaptureBackwardTime
RawCaptureAfterEOF
RawCaptureDuplicateExit
RawCaptureTrailing
RawCaptureInvalidScenario
RawCaptureInvalidHeader
RawCaptureInvalidActions
RawCapturePassiveProperties
RawCaptureArrayAndByteBounds
RawCaptureActionCountBounds
RawCaptureCountBounds
RawCaptureCopyOwnership
RawCaptureResultShapeFlags
RawCaptureDeclaredBinding
RawCaptureRefusedClassification
RawCaptureDerivedConsistency
RawCaptureComplete
```

Pin these exact case expectations before launching:

|Case suffix|Independent assertions|
|---|---|
|NormalBytes/Utf8Split/Utf16Split|stdout hex000102 -> Base64AAEC, stderrFF -> /w== with separate hashes; splits E4B8AD and2D4E reconstruct exact original bytes. Classification may refuse; do not replace binary with text|
|ShortRead/EmptyEOF|source0102 with Count4 returns two bytes, no EOF until a later read; empty source requires read0 on each channel, not an omitted EOF|
|EarlyExit/UnknownExit/UnreadTail|exit before reads can complete after both EOF; no exit stays null/incomplete; reading one byte of0102 retains only01 and stays incomplete when actions end|
|DeadlineEdges/EqualTime/BackwardTime|119999 read admitted;120000/120001 read maps to tick with no Data, timed-out; same-time order preserved; decreasing read time maps to tick, refused with unchanged prefix|
|AfterEOF/DuplicateExit/Trailing|after-EOF read maps to empty chunk/refusal; second exit refused; a read after complete maps to tick/trailing; preserve failing index and admitted bytes, never execute offending read|
|InvalidScenario/InvalidHeader/InvalidActions|null/wrong shape/mode/version; seven Header fields and exact argument/UTC/encoding bounds including reversed UTC; unknown fields/kinds, malformed action even after an earlier semantic stop cause whole-input preflight refusal with Actions=null|
|PassiveProperties/ArrayAndByteBounds|ScriptProperty getter increments a counter if invoked; rejection leaves counter0. Reject decorated Length/Rank, object[] instead of byte[], multidimensional arrays. Accept each exact1MiB source; refuse1MiB+1; mutated GetType ScriptMethod must not execute|
|ActionCountBounds/CountBounds|4096 ticks allowed vs4097 refused; Count1/65536 accepted,0/65537/bool/string/floating refused; negative/overflow integer and case/culture variant tags refused, no coercion|
|CopyOwnership|mutate input Actions,Header,argument/source bytes after return; returned snapshot/Record unchanged; mutate returned Actions and verify Record still unchanged; output has no stream/reference to input containers|
|ResultShapeFlags/DeclaredBinding|exact wrapper/nullability; all flagsfalse/cleanupunknown; changing executable/argument/UTC/session declaration changes corresponding binding; known literal TLRCAP01 bytes independently hash to Record.BindingSHA256|
|RefusedClassification/DerivedConsistency|complete exit-1 with both EOF has refused Package classification yet remains internally verified; compare exact expected mapped Events and full public verifier results for all four lifecycle states, not just a constructor self-check|

Representative assertions (fixtures expanded using the helper signatures above):

```powershell
# Each assert below is part of its case, not an additional PASS label.
if ($r.Record.Lifecycle.Stdout.Base64 -cne 'AQI=') { throw 'ASSERTION FAILED: RawCaptureShortRead' }
if ($r.Record.Events[0].Kind -cne 'tick') { throw 'ASSERTION FAILED: RawCaptureDeadlineEdges' }
if ($r.Record.CleanupStatus -cne 'unknown' -or $r.Acceptance) { throw 'ASSERTION FAILED: RawCaptureResultShapeFlags' }
```

ownership-mutation runs only ApiRequired,CopyOwnership,Complete; stop-mutation runs only ApiRequired,DeadlineEdges,Complete. Full tests use the same case bodies. HOST/context metadata separate from PASS labels. Freeze exact expected assertion prefix and error envelope; never accept any nonzero exit as RED.

- [x] **Step2 — run expected RED slots1/2.** No source module/stub yet; both must fail only ApiRequired. Otherwise stop without using later slots to diagnose/retry.
- [x] **Step3 — implement `New-OfflineRawCaptureRecord([object]$Scenario)` minimally.** Use passive exact-shape checks and original CLR types, validate Scenario then Header via a zero-event public packet then all Actions. Snapshot inputs before owned MemoryStream construction. Validate source byte limits in Scenario preflight; Header failure maps invalid-header; action shape failure maps invalid-action. Preserve only admitted plain scalars in returned Actions. Use direct byte Read once per allowed action; copy exactly returned count and Dispose/clear temporary holders in finally. No StreamReader/PowerShell pipeline text conversions or public runtime seam.
- [x] **Step4 — implement stop mapping and public reconciliation in the same module.** Apply spec§5/6 priorities. Maintain only position/EOF/exit/completion bookkeeping needed to avoid invalid reads; do not implement a second full lifecycle validator. Offending read becomes the specified empty chunk/tick. Construct the derived packet once after stopping, call the existing public constructor and verifier, check failure index/status against bookkeeping; inconsistency returns adapter-inconsistent, not a favorable result. Unknown exit stays null and source tails are never implicitly drained. Unexpected catch returns adapter-error with Record=null, retained validated Actions and no exception content; process-level failures have no recovery promise. Static review must trace all disposal/error paths because real I/O/resource failures cannot be truthfully exercised by exact MemoryStream success fixtures without broadening the interface.
- [x] **Step5 — run GREEN slots3/4.** Exact28 labels, exit0 and stream policy required. No product/test repair after an unexpected failure under this allowance.
- [x] **Step6 — run the two prescribed sensitivity checks, restoring immediately.** Slot5 temporarily replace the owned Actions snapshot assignment with the original Scenario.Actions reference; CopyOwnership must fail. Restore via apply_patch and verify baseline hash. Slot6 temporarily replace deadline-read-to-tick handling with the actual read/chunk branch; DeadlineEdges must fail on expected Events kind. Restore immediately even if a call fails; record diff and before/mutant/restored hashes. Slot5 completion uses the explicit readonly-admission exception, not a rewritten verdict. No git reset/checkout or dependency changes.
- [x] **Step7 — final new-suite and dependency verification slots7..10.** Exact restored implementation/test hashes and all12 prior pins must match before/after. No store or simulator full rerun added: unchanged public-contract regression is the explicit dependency scope. Capture all actual results, including limitations of memory-only reads and virtual time.
- [x] **Step8 — checkpoint.** Review explicit source/test/report/status/recovery/plan file list, whitespace/parse checks and receipts; preserve a clearly labelled implementation checkpoint on the existing branch. No acceptance/main claim before the separate review and no untracked raw receipts committed.

### Task2: Independent review and durable handoff

**Files:** execution report, current status, recovery and checked plan only; no product edits in this task.

**Interfaces:** consumes frozen Task1 source/tests, receipts and this spec/plan; produces a bounded review disposition and portable recovery state, not new runtime or acceptance authority.

- [x] **Step1 — review only after every earlier prerequisite.** One fresh independent reviewer checks byte copying, exact mapping/indices, passive preflight, ownership/disposal, two mutations, dependency pins and actual output policies. Include the five Review Focus classes; explicitly decline real-pipe scheduling, real timeout/cleanup and whole-branch security/acceptance claims. Reviewer must not import scripts or run tests.
- [x] **Step2 — reconcile findings.** Parent checks citations and source pins. Confirmed Important1 leaves the increment unfinished and stops; no automatic repair/review/retest budget. Minor1 is disclosed/deferred. No generic review pass converts synthetic facts into origin/cleanup proof.
- [x] **Step3 — save evidence and recoverability.** Record all10 consumed slots, expected versus actual outcomes and reviewer disposition in the execution report; preserve old failures and budgets unchanged. State whether new increment is verified/reviewed or incomplete. Full new module acceptance is not inherited from the old archive's approved verdict.
- [x] **Step4 — commit and push necessary files.** Verify expected origin and branch, stage explicit files, push ordinarily, confirm remote SHA=local and clean worktree. No main merge, force push, branch deletion, secret/cache/binary/raw receipt inclusion.

## Self-Review and User Handoff

Spec§1..4 map to scope/file/import boundaries;§5/6 to Task1;§7/8 remain explicit unimplemented future contracts, not missing Task1 work;§9 to the finite table and review;§10 to global constraints and Task2. All case labels map to a defined helper/interface; 28 labels include API/completion sentinels, not coverage percentage. Static disposal checks are not claimed runtime-fault proof.

Selected execution is **Native**: one coupled module and suite, implemented by the parent, then one fresh scoped read-only review. The original writing approval did not grant execution; the later option1 selection did, followed by the specific slot5 continuation. No unselected Subagent-driven implementation or extra review/repair budget follows.
