# Offline Supervision Policy P1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. Execution method and finite allowance below require user approval first.

**Goal:** Implement only a passive synthetic supervision-decision model, preserving a single absolute budget and never turning modeled facts into runtime authority.

**Architecture:** One new standalone pure PowerShell module validates a bounded trace, then reduces control observations into hypothetical supervision decisions. It performs no reads, clocks, processes or persistence and imports no existing module; the prior accepted byte coordinator remains unchanged. A separate portable suite supplies independent literal expectations.

**Tech Stack:** Windows PowerShell5.1/PowerShell7; plain objects, arrays, checked-range integer arithmetic and ordinal strings; no installation, Add-Type or new dependency.

**Spec:** [Joint capture/supervision design](../specs/2026-10-17-raw-capture-supervision-design.md), written atd9dcd740113ad2f7c8eacf48e3c043ae6e92e099 and explicitly approved by the user's yes to written-design approval/P1-plan writing.

Status: **NATIVE EXECUTION APPROVED; ALL8 HOSTS MATCHED, SCOPED REVIEW PENDING.** User selected1 afterc2fd4b5063e97fc1eebb46a6279d3ba24848b184, approving parent implementation, eight sequential synthetic hosts and one conditional scoped readonly reviewer. All8 consumed, no tests remain. See [execution evidence](../../roadmap/2026-10-17-offline-supervision-policy-execution.md). P2/P3/live runtime excluded. Earlier proposal wording is historical, not extra allowance.

## Global Constraints

- Reuse `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`, branch `codex/c12-b01-task9-coordinator`; inspect clean state/AGENTS before implementation, no new worktree by default.
- P1 only: no Process/Job/pipe/Stream/real clock/timer/sleep/file/network/security API, callbacks, scriptblocks, external handles or arbitrary command in the product. No module imports or load-time effects.
- Exact Mode=`policy-only`, EvidenceClass=`synthetic`; LiveActionsPerformed/OriginVerified/Acceptance/RuntimeApproved=false; CleanupStatus=`unknown` even when a hypothetical branch says verified.
- One absolute120000ms budget, stop at113500ms, internal shared natural500ms/cleanup5000ms/publication1000ms; equality is expired, no per-member reset or added time.
- Model only normalized integer synthetic milliseconds. ClockLabel=`synthetic-ms` is not QPC evidence; real QPC frequency/conversion, process/control authentication and OS compatibility belong to P2.
- Per-channel hypothetical retained bytes1048576, total2097152, data events4096; control summaries16, each1024 bytes; freeze control queue bound16 in this P1 contract. No byte arrays or byte-transfer simulation here.
- Preserve the accepted raw module FB48642D4A4F7CA8337DEDD61C4202C3C1C182168BEF23E574B22313386291C7, its test2D44F9C9EB68450866A92D43384E25C5DF4133F60150FB1DC3BE20EDBE1D914B, and all12 [prior dependency pins](../../roadmap/2026-10-17-synthetic-capture-dependency-regression.md). Freeze these14 before slot1; no dependency-suite reruns.
- No source repair/retest after an unexpected result, no automatic budget transfer, no new review cycle. Restore an authorized temporary mutant even on failure; otherwise preserve stopped state and ask direction.
- Historical detection/safety disposition, exact runtime packet/nonces, Stage B/I2/I3/Task10/11/full C12/main/release remain separate. No Go/Docker/services/native fixture/security queries or protection changes.
- Test hosts are ordinary inherited-token harnesses, not the future collector/supervisor; no elevation or sandbox relaunch to obtain a pass. Git preservation permission is separate.

## Review Focus

1. Late cleanup/persist claims must not erase an earlier timeout/loss: FailureSticky plus its specified mutant.
2. A delayed/relabelled message must not restart an absolute budget or forge actor direction: AbsoluteDeadline/StartupConsumesBudget/SessionClockRefusal/ActorDirection.
3. Executable descriptors, numeric coercion or shadow array bounds must not be evaluated/admitted: PassiveDescriptors/ArrayAndIntegerBounds.
4. Overflow or an exact phase boundary must not grant extra milliseconds: StopBoundary/TotalDeadline/PhaseBudgets, including nonzero starts and Int64-edge cases.
5. Hypothetical verified cleanup/output, nonzero exit, or caller mutation must not turn into real authority or modify another result: NonzeroExit/PersistRefused/CallerMutation/ConservativeFlags.

---

## File Structure and Frozen Interfaces

- Create `docs/diagnostics/isolated-fuzz/offline-supervision-policy.ps1`: public `Get-OfflineSupervisionDecision([object]$Scenario)` only; private helpers use `SupervisionPolicy` prefix. Pure scalar/result snapshots, no import dependency.
- Create `docs/diagnostics/isolated-fuzz/offline-supervision-policy.tests.ps1`: fixtures/comparators only in tests, Group=`all`/`cutoff-mutation`/`sticky-mutation`, default all. Optional import only when new module exists, then explicit required-API assertion.
- Create `docs/roadmap/2026-10-17-offline-supervision-policy-execution.md` during execution; update this plan's checkboxes, current status/recovery and necessary spec disposition, not unrelated code.
- Retain new ignored controller/index/streams under `.superpowers/sdd/2026-10-17-offline-supervision-policy/`; do not overwrite old packets or commit session/raw data. Portable tests must not require that directory.

### Input contract

Scenario is an exact plain PSCustomObject of six NoteProperties: Schema=1 (Int32/Int64), Mode=`policy-only`, SessionLabel (ASCII token `[A-Za-z0-9][A-Za-z0-9_-]{0,63}`), ClockLabel=`synthetic-ms`, StartMs and Observations. StartMs is exact Int32/Int64 in0..(Int64.MaxValue-120000). Observations is an actual zero-based rank1 object[] of0..4096 items, no shadow Length/Rank. Labels are declarations, not generated runtime identities.

Every observation has exact common NoteProperties Sequence,SessionLabel,ClockLabel,AtMs,Actor,Kind. Sequence/AtMs are exact nonnegative Int32/Int64; Sequence must equal its zero-based index, AtMs must be nondecreasing and >=StartMs. Session/clock match the frozen root ordinally. Actor is ordinal L/W/C; Kind and actor/payload combinations are exactly:

|Kind|Actor|Additional exact fields|
|---|---|---|
|arm|W|LauncherOwnsControl,SupervisorOwnsControl,ContainerBound,HandleInheritanceBlocked,ReadersReady: actual bool, all true required|
|release|W|InputBindingMatches: bool, true required|
|progress|C|StdoutRetainedBytes,StderrRetainedBytes,DataEventCount: exact nonnegative integers; StdoutEOF,StderrEOF,ExitObserved: bool; ExitCode: null iff ExitObserved=false, otherwise exact Int32-range integer|
|control|W|SummaryCount,MaxSummaryBytes,QueueDepth: exact nonnegative integers|
|time|L or W|none|
|fault|see below|Fault: exact ordinal token|
|cleanup|L or W|Verified,NaturalCompletion,CountsReconciled: bool|
|persist|W|ReadbackMatches: bool; Classification: ordinal clean/refused/unknown|

Fault/actor pairs: read-error or capacity-exceeded/C; control-overflow or identity-drift/W; control-lost/L or W; supervisor-lost/L; launcher-lost/W. No caller-provided exception/string payload, real path, command, process identifier, callback or source bytes.

All structures/descriptors/types are preflighted across the entire trace before reduction. Do not invoke ScriptProperty/CodeProperty/getters, overridden GetType/ToString or numeric conversion. Exact field names/tokens use ordinal comparison, not PowerShell case-insensitive hashtable matching. Refuse extra/missing/case-variant fields, subclasses/decorated arrays and executable descriptors; never reconstruct a favorable value by clipping. Bounds/clock arithmetic cannot overflow.

### Output contract

Return exact13 fields: Schema=1,Mode=`policy-only`,EvidenceClass=`synthetic`,Evaluated,Reason,State,FailureObservationIndex,Decision,OriginVerified=false,Acceptance=false,RuntimeApproved=false,LiveActionsPerformed=false,CleanupStatus=`unknown`.

Preflight failure: Evaluated=false, State=`Refused`, Decision=null, Reason=`invalid-scenario` or `invalid-observation`, index first bad observation or null for root. Unexpected recoverable internal exception: same conservative descriptor with Reason=`internal-error`, not exception content; no catastrophic recovery guarantee.

Valid preflight: Evaluated=true even for semantic refusal; State is Prepared/Armed/Running/Stopping/Sealed/Persisted/Refused/Aborted. Decision is a fresh plain object with exactly22 fields: DeadlineMs,StopAtMs,NaturalUntilMs,CleanupUntilMs,PublishUntilMs,LastAtMs,LastSequence,StdoutRetainedBytes,StderrRetainedBytes,DataEventCount,ControlSummaryCount,ControlSummaryByteMax,ControlQueueDepth,WouldCaptureStatus,WouldCleanupStatus,WouldStopOwnedSession,WouldNaturalCompletion,WouldEvidenceComplete,WouldPositiveTechnicalCandidate,WouldExitObserved,WouldExitCode,WouldClassification. Phase Until fields are null until stopping. No returned reference to input objects/arrays; no mutable cached empty-array exposure.

Initial hypothetical values: capture=incomplete, cleanup=unknown, classification=unknown, exit code=null, all Would booleans false; counters0, LastSequence=null/LastAtMs=StartMs. Capture status enum is incomplete/complete/timed-out/refused/fault; hypothetical cleanup is unknown/verified only. DeadlineMs=StartMs+120000, StopAtMs=StartMs+113500. Root flags never change with any modeled outcome.

### Reduction rules, priority and complete decision boundaries

1. Validate binding, actor, Sequence and nondecreasing AtMs before applying an observation. Exact failure reasons respectively binding-mismatch/wrong-actor/sequence-mismatch/time-went-backward; time before StartMs is also time-went-backward. Common-valid observations advance LastAtMs/LastSequence even if payload later refuses; bad common fields do not replace those two admitted values. First semantic failure sets reason/index once. After failure never clear it or re-enter positive stages; later time/cleanup/persist may add hypothetical failure-evidence facts only, never evidence-complete/positive-technical booleans. A later cleanup may set hypothetical verified facts only before CleanupUntil/D and natural=true only before NaturalUntil too; otherwise leave the corresponding fact unknown/false. Other later events cannot restart the run or replace the primary cause.
2. After binding/time checks, Persisted followed by any observation is trailing-after-terminal: Aborted, positive/evidence-complete false. For nonfailed nonpersisted states, AtMs>=DeadlineMs is total-deadline before any purported cleanup/persist success: State=Aborted, capture=timed-out/cleanup=unknown, WouldStopOwnedSession reflects whether arming had succeeded. Remaining phase/cutoff checks precede event application.
3. Prepared permits arm or time/control; all arm prerequisites must be true. Armed permits one release with binding true, time/control. Release before arm, repeat arm/release, progress before release or wrong-phase cleanup/persist fails unexpected-transition. A bad declared capability/binding fails prerequisites-unmet, without inventing ownership.
4. Progress is allowed only in Running. Time/control/fault are allowed in every nonfailed Prepared/Armed/Running/Stopping/Sealed state, subject to the phase deadline checks; cleanup only in Stopping and persist only in Sealed. Progress byte/event counters cannot decrease; EOF/ExitObserved cannot revert, known ExitCode cannot change: failure reason progress-regressed. Repeating an identical snapshot is allowed with next Sequence. Counter limits are the Global Constraints; over-limit progress/control fails capacity-exceeded before replacing admitted counters. SummaryCount/MaxSummaryBytes cannot decrease (progress-regressed); QueueDepth may decrease. Negative/coerced counters fail preflight.
5. In Prepared/Armed/Running, AtMs>=StopAtMs triggers capture-cutoff before arm/release/progress: capture=timed-out, State=Aborted, no positive candidate. At113499ms after start a valid noncomplete trace may remain Running; equality113500 may not. An already completed capture in Stopping/Sealed uses its phase budgets instead of the capture cutoff, still bounded by D.
6. Running progress with both EOF and ExitObserved sets WouldCaptureStatus=complete and enters Stopping at s=AtMs; no other snapshot infers EOF/exit. Freeze data counters then. NaturalUntil=min(s+500,D-6000), CleanupUntil=min(NaturalUntil+5000,D-1000), PublishUntil=min(CleanupUntil+1000,D), computed without intermediate overflow. Earlier failure also freezes stop s (clamped to D for safe arithmetic) and uses those bounded phase windows, but never regains success.
7. In a nonfailed Stopping trace, first check AtMs>=CleanupUntil (phase-deadline), then any observation at/after NaturalUntil without prior cleanup confirmation sets natural-window-expired, WouldStopOwnedSession=true, State=Aborted; a late cleanup message cannot revive a natural-completion claim. Cleanup before both limits may arrive early. Verified AND CountsReconciled true yields hypothetical cleanup=verified and State=Sealed; false yields cleanup-unverified/Aborted. Preserve declared NaturalCompletion separately; false can never yield a positive technical candidate. Once Sealed the natural/cleanup windows no longer invalidate completed facts; only the publication window/D bounds the next persist.
8. In nonfailed Sealed, persist strictly before PublishUntil with ReadbackMatches=true yields Persisted/WouldEvidenceComplete=true. WouldPositiveTechnicalCandidate=true requires complete capture, verified cleanup, NaturalCompletion=true, observed ExitCode=0, Classification=clean and no failure. Refused/unknown classification or nonzero exit keeps evidence-complete true but positive false, without falsely marking raw capture incomplete. Readback mismatch fails persist-unverified. Cleanup at/after CleanupUntil or persist at/after PublishUntil fails phase-deadline (unless global D already expired); unknown cleanup cannot be promoted by persist.
9. Faults retain their exact token as primary reason and index; after successful arming use Aborted and WouldStopOwnedSession=true (including failure while Armed), otherwise Refused/false. Semantic protocol/binding/time failures follow the same state/stop rule. For an incomplete capture read-error sets hypothetical capture=fault; semantic refusal sets refused; other faults retain incomplete. Already complete bytes remain complete except a total-deadline requires timed-out. Persist updates WouldClassification only when its fields are admitted. Only supplied virtual observations are evaluated; ending an incomplete trace does not advance time or fabricate faults/EOF.

Normal fixtures use StartMs=0: arm0, release1, progress2 (both EOF, exit0, data events3), cleanup3 (all true), persist4 (readback true/clean). Expected final: Persisted, Reason=trace-evaluated, index=null, D120000/Stop113500/Natural502/Cleanup5502/Publish6502, capture complete, hypothetical cleanup verified, observed exit=true/code0/classification clean, evidence/technical true, actual flags false/cleanup unknown. No-byte snapshot is intentional: P1 does not verify actual data events or provenance. Other no-failure traces use Reason=trace-evaluated; first failure reasons above remain sticky.

## Approved Finite Verification — consumed8/8, no renewal

At most8 sequential fresh children,120000ms each. No exploratory imports, extra baseline/full-suite/task-done commands, retry or automatic repair outside these slots. Static parsing/hashing/readonly receipts do not execute product and may be used. Missing engine/permission/timeout/unexpected stderr or assertion stops remaining slots and the reviewer.

|Slot|Engine/group|Frozen required result|
|---|---|---|
|1|PS7 all, product absent|exit1, exactly ASSERTION FAILED: P1ApiRequired; no PASS labels|
|2|PS5.1 all, still absent|same RED|
|3|PS7 all, new product|exit0, exact35 ordered labels below|
|4|PS5.1 all, same files|same GREEN|
|5|PS7 cutoff-mutation|exit1 at P1StopBoundary only, API PASS only|
|6|PS7 sticky-mutation|exit1 at P1FailureSticky only, API PASS only|
|7|PS7 restored all|exit0, exact35 labels|
|8|PS5.1 restored all|same final GREEN|

Freeze the complete test/hash/inventory/protected14 before1; product/hash before3. Cutoff mutant only changes the Running/Prepared/Armed cutoff comparison from `>=` to `>`; sticky mutant only clears prior failure reason/index and fatal latch on a later verified/reconciled cleanup. Change no test/expectation. Record mutant hashes/diffs, restore exact product immediately after each slot even if unexpected. These mutations are synthetic policy sensitivity checks, not real process behavior. Any unexpected result freezes all remaining work after restoration.

Controller resolves fixed existing x64 engines `C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe` and `C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe`; no install/fallback/elevation. Invocation is engine `-NoProfile -NonInteractive -File <test> -Group <group>` inside an owned ProcessStartInfo timeout controller, with PS5 inbox PSModulePath=`C:\WINDOWS\system32\WindowsPowerShell\v1.0\Modules`. This process use is harness-only, forbidden in product. Preconsume each slot, separate stdout/stderr, record context/version/admin role, exit/elapsed/timeout, source/test/output hashes; check protected14 and earlier receipt hashes before/after each slot. Normalize expected labels as object arrays for0/1/many. Timeout may stop only the exact owned test handle; report cleanup unknown if unproved, never arbitrary PID termination. Keep old raw unchanged.

GREEN stderr policy: empty or exclusively structurally recognized CLIXML Completed progress objects, with DTD/external resolution prohibited, no object deserialization; any Error/unknown/trailing payload fails. RED/mutants require exact named assertion, exit1, expected label inventory and the same strict nonerror stream policy, not any nonzero exit. No output/progress suppression or retrospective relaxation. Full stdout contains only CONTEXT/HOST/PASS/LIMIT metadata plus exact expected assertion for RED; no unexplained output.

## Task1: Passive policy module and independently expected suite

**Files:** new module/test named above; execution report and necessary status/recovery updates. No existing product/test edit.

**Interfaces:** consumes exactly the Scenario/observation contract above; produces Get-OfflineSupervisionDecision's13-field result and22-field hypothetical Decision. Tests own New-P1Scenario(StartMs,Observations), P1-Observation(Sequence,AtMs,Kind,Actor,payload), P1-Need(bool), P1-Case(name,body) and a recursive exact-value comparator. Helpers never generate expected decisions with product code. P1-Case emits one PASS only after all assertions; catch exact P1 assertion to stdout/exit1, unrelated errors fail normally.

- [x] **Step1 — write the entire suite before the module.** Full ordered labels: P1ApiRequired, P1EmptyPrepared, P1NormalStages, P1AbsoluteDeadline, P1StopBoundary, P1TotalDeadline, P1StartupConsumesBudget, P1EqualTimeSequence, P1BackwardTime, P1SequenceRefusal, P1SessionClockRefusal, P1ActorDirection, P1ArmPrerequisites, P1ReleaseOnce, P1ProgressMonotonic, P1ExitBeforeEOF, P1UnknownExit, P1NonzeroExit, P1DataBounds, P1ControlBounds, P1ReadFault, P1SupervisorLost, P1LauncherLost, P1FailureSticky, P1PhaseBudgets, P1CleanupUnknown, P1PersistRefused, P1TerminalTrailing, P1InvalidScenario, P1InvalidObservation, P1PassiveDescriptors, P1ArrayAndIntegerBounds, P1CallerMutation, P1ConservativeFlags, P1Complete. cutoff-mutation selects API/StopBoundary/Complete; sticky-mutation selects API/FailureSticky/Complete. Scope output LIMIT explicitly excludes OS/real clock/cleanup proof.
- [x] **Step2 — pin literal expected outcomes and negative vectors.** NormalStages compares every field of the complete normal result above; EmptyPrepared compares the entire initial result. AbsoluteDeadline shifts StartMs and every AtMs by1000: D121000/Stop114500, no rebasing on release/progress. StopBoundary checks113499/113500/113501 after start: last two timed-out/capture-cutoff with retained admitted counters. TotalDeadline attempts cleanup/persist exactly120000: total-deadline and no positive flag. StartupConsumesBudget releases at113499, then time113500; never gets another120000ms. EqualTimeSequence preserves supplied order. BackwardTime/SequenceRefusal/binding/actor cases give the exact first index and conservative result rather than sorting/coercing.

Representative assertions (inside their named P1-Case bodies; `$normal` is the five-event literal fixture described above, `$cutoff` is arm0/release1/time113500, not generated from product output):

```powershell
P1-Case ApiRequired { P1-Need ($null -ne (Get-Command Get-OfflineSupervisionDecision -CommandType Function -ErrorAction SilentlyContinue)) }
$r=Get-OfflineSupervisionDecision (New-P1Scenario 0 $normal)
P1-Need ($r.State -ceq 'Persisted' -and $r.Reason -ceq 'trace-evaluated')
P1-Need ($r.Decision.DeadlineMs -eq 120000 -and $r.Decision.StopAtMs -eq 113500)
P1-Need ($r.Decision.NaturalUntilMs -eq 502 -and $r.Decision.CleanupUntilMs -eq 5502 -and $r.Decision.PublishUntilMs -eq 6502)
P1-Need ($r.Decision.WouldPositiveTechnicalCandidate -and -not $r.RuntimeApproved -and $r.CleanupStatus -ceq 'unknown')
$r=Get-OfflineSupervisionDecision (New-P1Scenario 0 $cutoff)
P1-Need ($r.Reason -ceq 'capture-cutoff' -and $r.FailureObservationIndex -eq 2)
P1-Need ($r.Decision.WouldCaptureStatus -ceq 'timed-out' -and -not $r.Decision.WouldEvidenceComplete)
```

These selected lines do not replace the complete13+22-field literal comparisons required by Step2. Define each test helper only in the portable test file; New-P1Scenario returns the exact six root fields with fixed SessionLabel=`s`/ClockLabel=`synthetic-ms`, and P1-Observation returns common fields plus only the supplied case's payload. No production helpers construct expected results.

- [x] **Step3 — pin all Review Focus and remaining contract branches.** ArmPrerequisites flips each of five bools; release-before-arm/duplicate release fails. ProgressMonotonic separately regresses counters/EOF/exit/value and accepts identical snapshots; ExitBeforeEOF keeps Running until both EOF, UnknownExit stays incomplete/null, NonzeroExit retains complete evidence but no positive candidate. DataBounds admit1048576 per channel/2097152 total/events4096, reject+1 retaining prior counters; ControlBounds admit16/1024/16, reject+1 and counter regression. Fault cases preserve exact primary reason. FailureSticky: arm/release, fault supervisor-lost/L at2, then verified cleanup3 and clean persist4; first reason/index2 remains, State=Aborted, real flags and both positive/evidence-complete false. PhaseBudgets cover natural502, cleanup5502, publish6502 equal expiry plus one millisecond before, early cleanup, and Int64 maximum safe root start. CleanupUnknown/PersistRefused/TerminalTrailing cannot promote unknown/failure; refused classification permits evidence only. PassiveDescriptors includes getter/ToString/GetType counters that must remain0, case/NUL token variants, and decorated bounds; ArrayAndIntegerBounds includes null/coerced/bool/float/negative/oversize/2D/nonzero-lower-bound arrays and4096/4097 observations. InvalidObservation places a structural bad item after a synthetic timeout: full preflight still refuses, not partial evaluation. CallerMutation changes root/observations and returned Decision independently without affecting an earlier/separate result. ConservativeFlags asserts all13 root fields and bool/string types across normal/refused/fault results. Use literal full expectations for each first failure and admitted-counter snapshot, not product-generated expectations.
- [x] **Step4 — run frozen RED slots1/2.** Module absent; exact API failure required. Any other result stops before implementation.
- [x] **Step5 — implement Get-OfflineSupervisionDecision in the new standalone module.** Preflight descriptors/shapes/types before values, snapshot admitted scalars, then use the single ordered reduction rules above; no actual actions/clock/byte lifecycle recomputation or old imports. Keep first failure in one immutable latch; never call caller methods to inspect types. Return independently owned scalar objects. Trace overflow and all return/error paths statically; no invented fault-injection seam.
- [x] **Step6 — run GREEN slots3/4.** Exact35 labels, conservative flags/output policy and frozen pins must match. No product/test repair after unexpected failure under this allowance.
- [x] **Step7 — perform the two specified mutants in slots5/6 and restore.** Use apply_patch, save before/mutant/restored hashes, exact expected assertion, preserve original streams. No other source/test manipulation to manufacture sensitivity.
- [x] **Step8 — run restored final slots7/8, reconcile and checkpoint.** Recheck source/test/protected14, ordered inventories and every receipt hash; record actual results and untested limits in execution report. Review explicit file list/parse/whitespace, preserve a clearly labelled development checkpoint, commit/push ordinary to verified existing origin and confirm remote=local/clean. No acceptance/main claim.

## Task2: Conditional independent review and durable handoff

**Files:** execution report/status/recovery/checked plan only; no automatic code repair.

**Interfaces:** consumes Task1's frozen module/tests/pins/receipts/spec/plan; produces a bounded review disposition, not runtime authority.

- [ ] **Step1 — after all eight precise expectations hold, request ONE fresh independent readonly reviewer.** Scope only this P1 increment and relevant evidence; inspect passive checks, first-failure latch, absolute/phase arithmetic, complete literal expectations, both mutants, flags and pins. Reviewer may not import/run tests, edit, query security/network or delegate. Use a currently available most-capable reviewer configuration, not an unverified hardcoded model name. No whole historical-branch audit or main review is authorized by this allowance.
- [ ] **Step2 — reconcile every finding and declined-to-judge limit.** Important/Critical means unfinished/stopped, not a repair cycle; retain Minor explicitly and request direction if material. Do not call parent checks independent review. Declare real QPC/process containment/concurrent pipes/termination/durability/catastrophic/concurrent caller or hostile runspace behavior unproved; named synthetic tests are not a project-wide test pass.
- [ ] **Step3 — save final recovery state and ordinary commit/push.** Include source/tests/approved documents and portable reports, not raw/controller/cache/binaries/secrets. Preserve old and new receipts even on stop. Verify expected origin/branch, explicit stage list, remote SHA=local and clean worktree. Next work needs its own decision.

## Self-Review and Execution Handoff

Spec§1..3/4..8 are architecture context and exclusion boundaries; this plan implements only§9 P1 and respects§10 gates. No task implements P2/P3, QPC adapters, real arming/readers/control queues, recording or cleanup. Normalized virtual-ms snapshots only exercise the new control-decision policy, not a new byte simulator. All five Review Focus classes map to named Task1 assertions;35 labels include API/completion sentinels, not coverage percentage. Full-suite means this new suite only; old protected tests are not rerun with spent budgets.

Selected **Native** afterc2fd4b50: parent implementation, eight hosts, one conditional scoped fresh reviewer. All8 expected results now hold; review remains. No Subagent-driven implementation or historical whole-branch audit is selected, and no automatic extra review/fix cycle follows.
