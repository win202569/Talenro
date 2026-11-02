# Pin-Binding Recovery Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans natively after explicit plan/allowance approval. No implementation agents, extra tests or early reviewer. Steps use checkbox syntax.

**Goal:** Repair the observed pin-array collision, independently prove the missing call descriptor, then resume the original bounded contract/runtime chain.

**Architecture:** Freeze a focused actual-source pin regression, change only the local variable, and freeze that intermediate identity. A versioned call-model regression retains the original20 assertions but binds the new recovery epoch. The controller preserves all historical artifacts and produces the original still-unspent repair manifest.

**Tech Stack:** Existing Windows PowerShell7.6.5 Core x64, AST/JSON/SHA256; no installs.

**Spec:** [Approved pin-binding design](../specs/2026-11-02-offline-pin-binding-recovery-design.md), SHA BA143139835158A959ECBB1FD2D04655AF31C8E442D3BA0C2A988BDE1CB38736. User “是” after977b039b approves the written design. This implementation plan and execution allowance remain PENDING_APPROVAL.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; existing branch codex/c12-b01-task9-coordinator; origin https://github.com/win202569/Talenro.git.
- L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2; P=.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner; Q=.superpowers/sdd/2026-11-01-offline-controller-recovery.
- Keep exactly three runtime files in scope: L/source-model.ps1,observation-contract.ps1,driver.ps1. Other28of31 runtime sources and existing tests/oracles unchanged. New auxiliaries remain outside runtime inventory.
- Preserve all eight stops,155snapshots, old tests/inputs/results/archives. Failed computation10 remains exit2/0PASS/0FAIL/20NotRun; no refund, replay or reclassification. Accepted writer stays bound to its original source epoch and is never rerun.
- Request3additional auxiliary computations; ceiling14, consumed10. Four remaining computations are pin RED11, pin GREEN12, descriptor RED13, descriptor GREEN14 (the old unused GREEN allowance is reassigned explicitly).
- No additional runtime allowance: remaining4hosts/7targets/0compiles/31snapshots/conditional1review. Cumulative ceilings7hosts/3compiles/186snapshots. Actual consumed3hosts/3compiles/155snapshots/0targets/0reviews.
- Auxiliaries: existing tool context only; actual extracted definitions, pinned readonly file/AST access and explicit test-only manifest data. No module/profile/control/writer body, child host, compilation, target or exploratory probe.
- First unexpected result creates P/pin-binding.stop.json and blocks all later stages. No patch/retest, public IgnoreStop, arbitrary callback, default full-suite run, cleanup or additional review.
- Keep all process ownership, timeouts, source/oracle/recorder predicates and five false-capability flags. No main merge, force push, deletion, install, security-setting change or product acceptance from auxiliary success.

## Review Focus

1. A regex-valid pin overwrites the pin collection: ValidPinnedHash exercises the actual old/repaired function with the real source.
2. Renaming accidentally loosens validation: MissingPinRejected,AbsentPinRejected,MalformedHashRejected,ConflictingPinsRejected,ActualHashMismatchRejected retain exact rejection semantics.
3. Duplicate pins become incorrectly forbidden: IdenticalPinsAccepted compares actual returned hash to an independent file hash.
4. Descriptor RED still fails during setup rather than missing metadata: versioned RED requires exact first group/signature and19NotRun, never an arbitrary exception.
5. Recovery silently borrows a spent test or old source epoch: historical hash bindings, unique new intents, strict facts and original20group guards retain provenance.

## Task1 — focused pin RED and rename-only GREEN

**Files:** create L/pin-binding.tests.ps1; P/pin-binding.approval.json,
pin-binding.input.json,pin-binding-intermediate.input.json,pin-binding-red.*,
pin-binding-green.*; Q/pin-binding-progress.md. Only after expected RED, modify
the local pin-array variable in L/source-model.ps1.

**Interfaces:** Invoke-ODPinBindingRegression -Phase Red|Green. Loading defines
functions only. Seven ordered groups: ValidPinnedHash,MissingPinRejected,
AbsentPinRejected,MalformedHashRejected,ConflictingPinsRejected,
ActualHashMismatchRejected,IdenticalPinsAccepted.
Receipt ordered keys Schema,Phase,TestSHA256,InputSHA256,SourceSHA256,
PostSourceSHA256,Groups,PASS,FAIL,NotRun,ExitCode,RootError,RecordingError.
Schema int1; source hashes map repository-relative paths; errors preserve actual
Type,Message,ErrorId,Position,Stack. Setup/recording error exit2 never passes.

- [ ] Read plan/spec and confirm clean documentation-checkpoint HEAD. Verify parent writer manifest182B85886304BFE3CBFB248DDA70F2EB99DF3E2158F574DBE299519DE7CE0461 and2251pins, using writer-projection snapshots for operational sources. Verify155snapshots,eight stops and new paths absent. Check original test/input/final/tool/intent and accepted writer evidence/streams against preserved hashes. Record actual baseline three-file identities; no diagnostic execution.
- [ ] Create approval with ordered keys Schema,BaseCommit,UserReply,Method,DesignSHA256,PlanSHA256,ParentSHA256,HistoricalStopSHA256,AddedAuxiliary,AuxiliaryCeiling,RemainingHosts,RemainingTargets,RemainingCompilations,RemainingSnapshots,RemainingReviewers,HistoricalFiles,BaselineSources. Values1,actual HEAD/reply,Native,actual hashes,parent182B…0461,stop8581A9DA9A4CA3E7F2367757BE7616992889B434F42B1C4F35922B37921633ED,3,14,4,7,0,31,1. HistoricalFiles uses repository-relative paths for every bound old record, including original call-model intent and full tool record.
- [ ] Author focused regression. Extract uniquely the actual Get-ODPinnedSHA256 definition via AST, without evaluating other source bodies. Use actual pinned record-boundary.tests.ps1 and its manifest pin. ValidPinnedHash requires returned uppercase64hex hash equal an independently computed file hash. RED must preserve actual PropertyNotFoundException/PropertyNotFoundStrict with exact message “The property 'SHA256' cannot be found on this object. Verify that the property exists.” Result0PASS/1FAIL/6NotRun/exit1. Any other error/pass stops.
- [ ] Freeze input ordered keys Schema,ApprovalSHA256,DesignSHA256,PlanSHA256,ParentSHA256,HistoricalStopSHA256,TestSHA256,BaselineSources,HistoricalFiles. No future hash dependency. Freeze test/input identities and push a new preparation archive checkpoint before execution. Phase intent ordered keys Schema,Phase,AttemptConsumed,Computation,ComputationCeiling,TestSHA256,InputSHA256,SourceSHA256; values1,Red,true,11,14 and actual identities. CreateNew before Invoke-ODPinBindingRegression -Phase Red once; preserve full tool chunks and actual receipt.
- [ ] After exact RED only, rename every reference to local $matches inside Get-ODPinnedSHA256 to $pinRows. No other semantic change. Static AST/function-text comparison modulo this identifier must establish rename-only repair and all other function texts unchanged; do not execute a separate probe. Preserve intermediate complete source bytes in Git, not a new runtime snapshot.
- [ ] Create immutable pin-binding-intermediate.input.json ordered keys Schema,InputSHA256,SourceSHA256,HistoricalFiles; int1,focused input hash,three current source hashes,old artifact map. Freeze/push rename-only checkpoint before GREEN. This input is future-independent and anchors both pin GREEN and subsequent descriptor RED.
- [ ] Consume pin-binding-green intent12of14, run Invoke-ODPinBindingRegression -Phase Green once. Require7PASS/0FAIL/0NotRun/exit0, null errors and unchanged source/test/input hashes. Missing pin → exact 'OD: source model input not pinned'; Absent,malformed hash,conflicting duplicate → exact 'OD: source model pin conflict'; wrong well-formed hash → exact 'OD: source model input changed'; valid and identical duplicate pins return real hash. Use cloned in-memory manifests; never edit actual source/manifest to simulate bad data. Preserve all actual errors, no broad rejection catch.

## Task2 — versioned descriptor RED, three-source implementation and GREEN

**Files:** create L/contract-call-model-v2.tests.ps1; P/contract-call-model-v2.input.json,
contract-call-model-v2-candidate.input.json,contract-call-model-v2-red.*,
contract-call-model-v2-green.*. Modify only the same three runtime sources.

**Interfaces:** Invoke-ODCallModelRegressionV2 -Phase Red|Green; original20group
receipt schema unchanged. Read-ODSourceModel,Get-ODIntentCallBinding,
New-ODIntentUnitContext and OC-SourceError interfaces/descriptor fields are exactly
[previous plan Task2](2026-11-02-offline-contract-call-model.md#task2--implement-three-source-amendment-and-one-green).
The rename is retained in addition to that descriptor work.

- [ ] Copy frozen original regression to v2 without editing the original. Rebind entry name/file/input/candidate/phase prefixes; ExpectedPinCount2276; new recovery facts helper/stop/preparation identities below. Preserve all20behavioral group bodies except these explicit namespace/hash/epoch constants. The old named-call comparison uses the historical Read-ODSourceModel definition with the now-repaired pin helper; record this accurately, not as successful execution of the old broken helper.
- [ ] Freeze v2 input ordered keys Schema,ApprovalSHA256,DesignSHA256,PlanSHA256,ParentSHA256,HistoricalStopSHA256,TestSHA256,BaselineSources,HistoricalFiles. BaselineSources MUST be rename-only intermediate three-source hashes, not pre-repair hashes. Bind intermediate input and focused RED/GREEN finals in HistoricalFiles. Freeze/commit v2 test/input before descriptor RED; do not add future candidate hashes.
- [ ] CreateNew descriptor RED intent13of14 with same intent schema, then Invoke-ODCallModelRegressionV2 -Phase Red once. Require0PASS/1FAIL/19NotRun/exit1 and RuntimeException Message/FQID 'OD: fixed owner-control descriptor missing'. Its setup must pass. Original failed computation10 stays untouched and is not evidence for this assertion.
- [ ] Implement approved descriptor and pure bridge exactly per previous Task2: separate DynamicCalls, unique fixed controls/Name literal block, exact AST-derived source/hash/owner/variable/invocation/body extent, explicit unit-only binding, unchanged named Calls/aliases and first-error operation checks. Keep runtime constructors without unit fallback.
- [ ] Implement Driver-PinBindingGate and Assert-ODPinBindingFacts([Collections.IDictionary]$Facts). Facts ordered keys Schema,ParentSHA256,HistoricalStopSHA256,ApprovalValid,AttemptConsumed,NewStopPresent; exact values int1,parent182B…0461,stop8581…33ED,typed bool true,false,false. Other inputs throw exact RuntimeException Message/FQID 'OD: pin binding recovery mismatch'. Gate checks all four actual receipts/intents and both prior stops, approval/test/input hashes and old writer originals. Historical old source comparisons use existing snapshot epochs, not edited current files.
- [ ] Add sole Prepare grant pin-binding-rebind; output remains repair-green.frozen.json and snapshots/repair-green. Block all earlier preparation routes/writer replay; RunOne retains contract-green/core/lifecycle/continuation. New outer stop pin-binding.stop.json uses CreateNew. Preserve old recovery maps and add PinBindingRecovery separately; do not pretend the proposed earlier CallModelRecovery manifest existed.
- [ ] Freeze v2 candidate input ordered keys Schema,InputSHA256,CandidateSources,HistoricalFiles,ExpectedPinCount; int1,v2 input hash,exact three-source hashes,immutable history,2276. Driver may embed earlier approval/test/input hashes but never its own candidate hash or the containing candidate-input hash. Keep strict source-window and original process guard definitions. Commit/push source/test/input checkpoint before GREEN.
- [ ] Consume versioned descriptor GREEN intent14of14; run Invoke-ODCallModelRegressionV2 -Phase Green once. Require original20PASS/0FAIL/0NotRun/exit0, null errors, pre/post hashes unchanged. No new separate scope computation. Real saved unit error and negative tests run through actual pure consumer; no runtime manifest fabrication. First unexpected result stops.

## Task3 — freeze2276pins, conditional original runtime chain and handoff

**Files:** P/repair-green.intent.json/frozen.json,snapshots/repair-green and original
unspent runtime stage records; new L/portable-pin-binding-preparation.json,
portable-pin-binding-intermediate.json,portable-pin-binding-candidate.json,
portable-pin-binding-freeze.json,portable-pin-binding.json as each checkpoint is
actually reached. New execution report docs/roadmap/2026-11-02-offline-pin-binding-recovery.md;
update current-status.md and repository-recovery.md.

- [ ] Inherit2251pins and change exactly three operational source hashes. Append exactly25unique paths listed below; require2276pins/31sources. Planning confirmed none of these25paths is already a parent pin; recheck before freeze. Preserve InheritedPins2482/InheritedUniquePins2172, source list and historical OwnerScopeRecovery/ErrorProjectionRecovery data unchanged.
- [ ] Add PinBindingRecovery ordered fields Schema,ParentSHA256,ApprovalSHA256,DesignSHA256,PlanSHA256,PinTestSHA256,PinInputSHA256,IntermediateInputSHA256,PinRedFinalSHA256,PinGreenFinalSHA256,ModelTestSHA256,ModelInputSHA256,CandidateInputSHA256,ModelRedFinalSHA256,ModelGreenFinalSHA256,HistoricalStopSHA256,HistoricalArtifacts,Consumed,Ceilings. Schema1; actual file hashes, eight-stop/failed-attempt/writer artifact map; consumed auxiliary14/hosts3/compiles3/snapshots155/targets0/reviews0 before freeze. Ceilings auxiliary14/hosts7/compiles3/snapshots186; current remaining MaxHosts4/MaxChildren7/MaxCompilations0/MaxReviewers1. Never reset consumed counters.
- [ ] Run driver -Action Prepare -Grant pin-binding-rebind -PreviousFrozenSHA256 182B85886304BFE3CBFB248DDA70F2EB99DF3E2158F574DBE299519DE7CE0461 once. CreateNew repair-green.intent before outputs; require2276pins/exact31new snapshots/total186 and exact three-source window. No partial-freeze retry.
- [ ] Verify and push NEW freeze portable archive before runtime: actual P/Q files, directories and zero-byte files, exact decoded bytes/lengths/hashes. Explicit-file staging, staged raw bytes equal disk, diff checks, commit/push current branch, remoteSHA==HEAD. Archive generation is mechanical preservation, not an extra diagnostic computation.
- [ ] Run original four hosts once each only after preceding exact acceptance: contract-green52PASS; core25PASS/22rows/0targets; lifecycle17PASS/8rows/3targets; continuation16PASS+12seamPASS/15rows/4targets. All use actual repair manifest and original recorder/source/ownership/capability gates. Keep120000ms host cap except continuation150000; sequence140000/cleanup4000; exit/shared drain2000; hold45000/receipt2000/timeout200/seam110000; session/sinks32,Ack2,observations32,secondaries8.
- [ ] Only after all actual runtime stages pass dispatch original conditional single readonly runtime reviewer. No early/extra review or compile. Preserve no-main-merge boundary and report actual acceptance only.
- [ ] On success OR stop preserve full actual tool chunks/first error/partial records with no reconstruction; create NEW portable-pin-binding.json, verify decoded equality, update actual counts/status/recovery and push explicit necessary files with remote equality. Never overwrite an old archive, stop or frozen test.

### Exact25new provenance pins

1. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/writer-projection-repair.frozen.json`
2. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/error-projection.stop.json`
3. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model.stop.json`
4. `docs/superpowers/specs/2026-11-02-offline-contract-call-model-design.md`
5. `docs/superpowers/plans/2026-11-02-offline-contract-call-model.md`
6. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model.approval.json`
7. `docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/contract-call-model.tests.ps1`
8. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model.input.json`
9. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model-red.final.json`
10. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model-red.tool-record.json`
11. `docs/superpowers/specs/2026-11-02-offline-pin-binding-recovery-design.md`
12. `docs/superpowers/plans/2026-11-02-offline-pin-binding-recovery.md`
13. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/pin-binding.approval.json`
14. `docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/pin-binding.tests.ps1`
15. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/pin-binding.input.json`
16. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/pin-binding-intermediate.input.json`
17. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/pin-binding-red.final.json`
18. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/pin-binding-green.final.json`
19. `docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/contract-call-model-v2.tests.ps1`
20. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model-v2.input.json`
21. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model-v2-candidate.input.json`
22. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model-v2-red.final.json`
23. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/contract-call-model-v2-green.final.json`
24. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/writer-projection-green.final.json`
25. `.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/error-projection-writer.tool-record.json`

Actual intents and complete tool records for all four new phases are additionally
bound in HistoricalArtifacts and portable records; they must match their finals
and computation numbers. Receipt bindings do not replace actual raw evidence.

## Self-review and execution approval

Task1 isolates the observed bug without descriptor changes; Task2 establishes the
separate descriptor RED and preserves original20assertions; Task3 covers provenance,
finite runtime and recovery. All five review-focus risks have tests/gates. Inputs
precede outputs; intermediate identity prevents RED from accidentally expecting
pre-rename source;2276includes all25new pins, not obsolete2263metadata. No tests,
runtime edits or computation occurred in this planning turn. Approval must cover
this plan,3additional auxiliary computations and conditional use of the original
remaining runtime chain. Native execution retained.
