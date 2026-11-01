# Offline Controller Recovery Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans natively after explicit approval. No implementation agents, per-task review, extra runtime reviewer, default-suite reruns or workspace deletion. Track checkbox steps in this plan's own ledger.

**Goal:** Repair the confirmed manifest type boundary and resume the unstarted writer validation through an explicit append-only source revision, without rewriting historical evidence.

**Architecture:** Keep both existing freezes and stop in place. A new fixed writer manifest binds exactly three source changes and the recovery provenance; existing writer/intent/prior-final consumers select this same manifest. Runtime manifest decoding yields IDictionary before validation, with all original gates retained.

**Tech Stack:** Existing Windows PowerShell7.6.5 Core x64, .NET JSON/AST/hash APIs and existing bounded diagnostic process code. No installs.

**Spec:** [Selected readonly audit/proposal](../../roadmap/2026-11-01-offline-controller-readonly-audit.md), ca715395a94ab5a795d2f2ced772429a667329c9, SHA256 EC4FF0038A10BF613A7F99D264953F81B6B4BB20B0B872023C74281E2537C2EB. User selected its recommended approach and requested this plan. Also read [original implementation plan](2026-11-01-offline-i1-i2-single-owner.md) and its approved single-owner design. This amendment changes source ownership/provenance and finite recovery only, not I1/I2 assertions.

Status: PLAN_PENDING_EXECUTION_APPROVAL. This document is not a run approval.

## Global Constraints

- Existing worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; branch codex/c12-b01-task9-coordinator; origin https://github.com/win202569/Talenro.git. No main merge, deletion, overwrite of old evidence or history rewrite.
- L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2; P=.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner; Q=.superpowers/sdd/2026-11-01-offline-controller-recovery. Q owns this execution ledger; this plan explicitly permits new named append-only records in P, not edits to its old ledger/receipts.
- Preserve old seed SHA4E4342C389C507F6148AE029140814FBCB4046AD62D1EFE3BFD451863C9DC1BA, writer SHA9AC3EAD775DB2BE7D21DE3B1CD3FE46342B1B176BA6B07C03F082D0F63B0C548, stop SHAF19687B1628DC1C48CC771A07DA4308D48E89B8D51921DDACDB6C4CBFAFDE9D2, both REDs,62snapshots and all previous archives at their original paths/bytes. No seed/old writer Prepare or old RED rerun.
- Existing31runtime sources stay enumerated exactly. Only L/driver.ps1, L/record-boundary.tests.ps1 and L/source-model.ps1 may change in this amendment. Other28unchanged. Later original Task3 still changes only L/observation-contract.ps1 after exact writer success.
- New auxiliary L/controller-recovery.tests.ps1 is a separately pinned regression script, NOT a32nd runtime SourcePath and never imported by capture/profile. It may not modify fixture/test oracles or replace production logic.
- Current source snapshots62of93; request31extra runtime snapshots for this revision. Revised total ceiling124=62existing+31controller+31original repair. No new runtime snapshot stage thereafter.
- Request exactly2additional existing-tool-context regression computations, once RED then once GREEN. No new capture host, target, Add-Type, publisher/module fixture or real-profile invocation in these computations. Original2availabilityREDs remain consumed, distinct and unchanged.
- Original remaining5capture hosts/7targets/1fixture compile/conditional1readonly runtime reviewer remain the only runtime budget. Recovery expressly permits one new RunOne writer invocation after gates, not an automatic retry. Later stages keep original order/expectations/limits; no refunds, spares or extra final host.
- All five capability flags remain false. Session/Ack never restored from disk. Preserve unchanged ErrorRecord-reference collection, source AST checks, exact bytes/streams and exact Process ownership; no PID lookup/kill-tree, Force/IgnoreStop, arbitrary source/manifest roots, installs or security relaxation.
- First unexpected authoring/source/provenance/record/regression/transport/runtime failure stops; the newly specified exact root-cause RED is the only additional intended failure. No fix/retest/refreeze after failure without new authority. Archive on success or stop; necessary checkpoint commit/push is already authorized.

## Review Focus

1. A helper-only fix must not leave the runtime context receiving PSCustomObject: ProductionManifestDictionary covers actual decoder AST, flags and typed boundary.
2. A historical stop must not become a general bypass: RecoveryExact/MissingApproval/ChangedStop/WrongParent/NewStop controls and fixed-path production checks.
3. Rebinding a writer path must not mix source epochs: WriterBindings checks writer unit, intent context, prior final and future repair parent against the same manifest.
4. Remaining budget must not permit an attempted grant to run twice: ExistingAttempt control plus original per-grant intents/output absence and prior-final gates.
5. Updated provenance must not masquerade as old frozen tests: exact3source diff, six explicit additional pins, unchanged assertions/oracles and preserved original RED labels.

## Task1 — freeze regression and observe the exact root failure once

**Files:** create L/controller-recovery.tests.ps1; Q/progress.md; P/controller-repair.approval.json, P/controller-regression.input.json, P/controller-regression-red.intent.json and later actual RED records. No existing source edits yet.

**Interfaces:** Test script accepts only -Phase Red|Green. It reads fixed repository/P paths and extracts actual AST definitions of Driver-Need and Driver-FalseFlags, not driver top-level execution. It also extracts the unique assignment to variable manifest whose RHS pipeline reads -LiteralPath $manifestPath and calls ConvertFrom-Json. Require one AST match, no string fallback. Evaluate that exact assignment in a private script scope with manifestPath fixed to the original writer manifest, then assert IDictionary and call the actual flags helper. No driver invocation/import or new process.

- [ ] Verify branch/origin/clean baseline, all2209writer pins and31writer snapshots; hash old seed/writer/stop and selected audit. Confirm all five grants' intent/final/stdout/stderr/evidence and capture/unit/profiles/repair manifest absent. Stop on mismatch; never recreate originals from portable.
- [ ] Record explicit approval of THIS plan, its actual SHA, selected audit SHA,3source window,2regression computations,31extra snapshots, original runtime remainder and fixed recovered old stop/parent in CreateNew approval. Initialize Q ledger. Disk approval records provenance, not a runtime capability.
- [ ] Author and commit the auxiliary regression script before RED. Freeze its exact SHA plus input identities in controller-regression.input.json (CreateNew), then verify bytes before each phase and after completion. No runtime snapshots/compilation for this auxiliary file; its exact source is tracked in Git and preserved in the eventual portable packet.
- [ ] Fix13ordered group names and independent expectations in that script, before RED: ProductionManifestDictionary, FiveFalseFlags, TrueFlagRejected, StringFlagRejected, MissingFlagRejected, NullFlagRejected, RecoveryExact, MissingApproval, ChangedStop, WrongParent, ExistingAttempt, NewStop, WriterBindings.
- [ ] Consume RED intent before computation; run `. L/controller-recovery.tests.ps1; Invoke-ODControllerRecoveryRegression -Phase Red` in the existing tool context once. ProductionManifestDictionary must catch actual CannotIndex with the saved error message from actual original decoder/helper; then mark its type/flags assertion FAIL, remaining12NotRun, exit1. Evaluate actual flags helper before the separate IDictionary assertion so the original CannotIndex remains visible. Unexpected signature, exception, helper absence or passing root case stops. Do not run the later cases in RED.
- [ ] Save complete original command/tool chunks after terminal return into controller-regression-red.tool-record.json and exact final receipt. It proves this root cause only, not I1 behavior, availability RED replacement or a complete old driver run.

## Task2 — exact three-source repair, one GREEN and one append-only freeze

**Files:** modify exactly the3existing leaf sources; auxiliary regression remains byte-frozen. Create P/controller-regression-green.intent.json/records, P/controller-repair.intent.json, P/writer-controller-repair.frozen.json, P/snapshots/controller-rebind/<31source paths>. New failures use P/controller-repair.stop.json, never old stop.json.

**Interfaces / fixed provenance:**

- Add only fixed preparation Grant `controller-rebind` to driver; RunOne still permits only original five grants. Prepare seed/writer-green cannot be used again. No public recovery/manifest/path/ignore switch.
- The new writer file is `writer-controller-repair.frozen.json`, Schema2/Phase `writer-green`; its PreviousFrozenSHA256 is the original writer SHA. SourcePaths/SourceCount31 and inherited2482/2172counts stay unchanged. Replace only3source pin hashes; other original2206pins unchanged. Add exactly6new unique Present pins: original writer manifest, old stop, selected audit, THIS plan, controller-repair.approval.json, auxiliary regression script. Result PinCount2215. No mutable ledger/tool output becomes a pin.
- Preserve original manifest PlanSHA256/DesignSHA256 and limits. Add ControllerRecovery ordered map: Schema(1), HistoricalWriterSHA256, HistoricalStopSHA256, ApprovalSHA256, AmendmentPlanSHA256, AuditSHA256, RegressionTestSHA256, RegressionRedSHA256, RegressionGreenSHA256. Last two hash exact phase-final receipts, validated explicitly at each invocation; no circular self-hash or mutable success file.
- driver's writer RunOne, Driver-RequireFinal(writer-green), and Prepare repair-green prior selection use the new writer path. Runtime phase comparison stays `writer-green`, not the filename. Repair stays `repair-green.frozen.json` and its PreviousFrozenSHA256 must equal the actual NEW writer hash.
- Use private `Get-ODWriterManifestPath()` → one string, the fixed Join-Path of packet and writer-controller-repair.frozen.json, at those three driver sites; it accepts no parameter and has no old-writer fallback. New freeze creation uses this same fixed result. WriterBindings verifies the three actual caller ASTs select this function, not merely that the new filename appears somewhere in driver text.
- record-boundary.tests.ps1: replace ONLY its fixed intent ProfileManifestPath filename with writer-controller-repair.frozen.json. Keep ProfilePhase writer-green. source-model.ps1: replace ONLY New-ODIntentUnitContext's fixed writer filename. Its existing prior-SHA check remains. No oracle, group, exception, factory, row/count/reference or source-location expectation changes.

**Recovery gate:** Add private pure `Assert-ODControllerRecoveryFacts([Collections.IDictionary]$Facts)` → no output; mismatch RuntimeException Message/FQID `OD: controller recovery mismatch`. Exact ordered facts: Schema(int1), HistoricalWriterSHA256(string), HistoricalStopSHA256(string), ApprovalValid(bool), ForbiddenOutputsPresent(bool), NewStopPresent(bool). Require fixed historical hashes, ApprovalValid true, both other booleans false; missing/wrong keys/order/types reject. Facts are internally built from fixed paths, not caller-supplied runtime authority.

Before Prepare controller-rebind and writer RunOne, compute forbidden outputs from all original grant outputs/capture/unit/profiles plus unexpected repair outputs; before later stages, use current/future stage guards and require exact completed prior finals. Earlier valid writer outputs are not forbidden for a later contract stage. Old stop is accepted only after its exact hash, explicit approval/plan/test input hashes and applicable phase receipts match. Before new freeze exists, use fixed approval + regression input identities; afterward require ControllerRecovery and supplied FrozenSHA256. Always reject a present new stop. File reads/hash validation remain outside the pure facts helper. Never infer authorization merely from an approval boolean/file.

- [ ] Apply the RunOne decoder `ConvertFrom-Json -AsHashtable` at its single actual boundary; retain the inline AST assignment used by regression. Do not change Prepare's other object-style loaders without necessity. Implement strict gate/provenance/unique preparation path and append-only catch destination; retain all original constraints and per-grant guards.
- [ ] Apply the two literal binding-only edits; inspect exact diff against writer snapshots. Verify other28sources and old originals unchanged; PowerShell/JSON/LF static checks only, no spare runtime probe.
- [ ] Consume GREEN intent, run same frozen auxiliary script with -Phase Green once. Require13PASS/0FAIL/0NotRun/exit0 and no emitted production-helper values. Extract actual new private facts helper AST only inside this regression scope; do not execute driver top level.

| GREEN groups | Exact input / assertion |
| --- | --- |
| ProductionManifestDictionary | actual production decoder AST reads old fixed manifest into IDictionary; actual flags helper accepts; static source-model signatures still require IDictionary |
| FiveFalseFlags | literal ordered fivefalse bool map; helper emits nothing |
| True/String/Missing/Null flag groups | independently change SourceVerified to true, string false, absent or null; actual helper must reject; no coercion |
| RecoveryExact | literal ordered facts with int1/fixed historical hashes/true,false,false; no output |
| MissingApproval / ChangedStop / WrongParent | one mutation each: ApprovalValid false; stop all-zero64hex; writer all-zero64hex; exact recovery mismatch |
| ExistingAttempt / NewStop | one mutation each: ForbiddenOutputsPresent true or NewStopPresent true; exact mismatch |
| WriterBindings | AST/text literals select ONLY new writer filename at the2consumers; driver RunOne/prior/final fixed selection agree; repair filename unchanged; no old-writer fallback; existing lifecycle/source/output checks retained |

- [ ] Save complete GREEN command/chunks and exact receipt before Prepare; no synthesized PASS/final. Exact phase receipts: Schema,Phase,TestSHA256,InputSHA256,SourceSHA256,PASS,FAIL,NotRun,ExitCode,Groups,RootError; groups ordered Name/Status. SourceSHA256 is an ordered map of the3actual source-relative paths to hashes, captured before/after each phase and required equal within that phase. RED RootError retains actual ErrorId/Message/Position/Stack; GREEN null. Counters numeric, statuses literal PASS/FAIL/NotRun. GREEN may differ from the baseline source hashes only in the approved3files; validate baseline input hashes against preserved original writer snapshots, not against the deliberately edited GREEN disk sources. Prepare must match all3GREEN source hashes before freezing; no post-GREEN edit is authorized.
- [ ] Invoke `L/driver.ps1 -Action Prepare -Grant controller-rebind -PreviousFrozenSHA256 <original-writer-SHA>` once. Validate old writer mapping with only3allowed changed pins, original RED receipts against seed, exact new regression receipts, approval and all prelaunch absences. Create controller-repair.intent.json BEFORE new snapshots/manifest, with Schema1,AttemptConsumed true,old parent/stop/test/plan/approval hashes. Existing intent/snapshot root/new manifest rejects; new failure writes new stop.
- [ ] Snapshot all31runtime sources once, CreateNew, hash each; create new2215pin manifest once. Require Prepared controller-rebind, Phase writer-green,31snapshots/exact3ChangedInputs and actual new SHA; save terminal tool record outside inventory-sensitive execution. Never overwrite old writer or stop, even if a new freeze fails partially.
- [ ] Commit/push explicit source/regression/docs and byte-labelled portable records, staged raw bytes equal disk, remote SHA equals HEAD. Do not declare original Task2 complete yet.

## Task3 — bounded resume and original conditional review/handoff

**Consumes:** exact new freeze/GREEN/provenance. **Produces:** only original finite runtime evidence, or a clearly unaccepted stop checkpoint.

- [ ] Once only: `L/driver.ps1 -Action RunOne -Grant writer-green -FrozenSHA256 <actual-new-writer-SHA>`. Require26PASS/exit0 AND exact original transport, recorder, private references, healthy default probe/final publication, pins and parent final;0targets/1in-memory fixture compile. Full original tool chunks retained. A yield is not completion; wait only returned session identifier. Any unexpected result freezes all remaining stages/review, no automatic edit or rerun.
- [ ] Only if writer exact, execute original plan Task3: implement ONLY observation-contract.ps1, other30runtime sources fixed to new controller epoch; Prepare repair-green once against NEW writer SHA,31CreateNew snapshots, then contract-green once52PASS/exit0/0targets/0compile. Actual unit prior context must bind new writer SHA; no historical fallback. Source snapshots reach124total.
- [ ] Only if contract exact, original core once25PASS/22rows/0targets, lifecycle once17PASS/8rows/3targets, continuation once16PASS+12seamPASS/15rows/4targets. Keep original fixed120000/150000capture,140000sequence,4000cleanup,2000owned-exit/shared-drain,45000hold,200timeout,110000seam limits and original bounds.
- [ ] Only after original two exact REDs, amended regression RED/GREEN, new provenance and ALL five exact actual host finals, use the already remaining conditional1fresh readonly runtime reviewer under original Task5 restrictions. Include amendment, three-source diff, auxiliary results and old/new identity chain. This is not an extra review. Important/Critical/incomplete stops without fix/retest/second review.
- [ ] On success OR any stop, preserve every actual original/partial/empty output, intent, receipt, old/new stop, manifest and snapshot from P and this plan's Q ledger in a NEW portable-controller-recovery.json with explicit root-relative namespaces/directories/base64/length/SHA. Verify decoded equality, never reconstruct missing originals. Preserve integer text/raw streams without JS-number normalization.
- [ ] Update current-status, original execution report and repository-recovery with actual consumption, next blockers and narrow verdicts. Commit/push explicit necessary files to existing dev branch, check staged raw blobs and remote equality. No I1/I2 historical closure or main/product acceptance inferred.

## Self-review / handoff gate

Coverage: decoder/type chain→Task1+2ProductionManifestDictionary; strict flags/recovery→Task2literal matrix; same writer SHA in three consumers→WriterBindings+new freeze+actual host; immutable history and finite snapshots→Task2mapping/new stop; original evidence/lifecycle/conditional review→Task3. No independent implementation/per-task agent requested.

This plan intentionally changes the original frozen test source's filename binding and source-snapshot ceiling; it does NOT claim the old tests are byte-identical or that old availability REDs validate this amendment. The extra2regression computations are explicit, not hidden reruns. Runtime budget remains5/7/1/conditional1. Please approve this entire amendment, Native method,3existing-source window plus auxiliary regression script,31extra snapshots and2regression computations before any implementation or renewed RunOne.
