# Contract Call-Model Amendment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans natively after plan/allowance approval. No implementation agents or extra reviews. Steps use checkbox syntax.

**Goal:** Bind the fixed owner-control module call from pinned AST data, then resume only the previously unspent contract/runtime stages.

**Architecture:** Add separate dynamic-call metadata without changing named Calls. Explicit unit context supplies the sole bridge binding; the pure contract checks it against the independently parsed source model. A new controller preparation gate preserves the accepted writer epoch and freezes exactly three changed runtime sources.

**Tech Stack:** Existing Windows PowerShell7.6.5 Core x64; AST, JSON and SHA256 APIs; no installs.

**Spec:** [Approved scope amendment](../specs/2026-11-02-offline-contract-call-model-design.md), SHA256 2F9A0948D0C245919F94D8703E1B5AE51EDC43019C3CF424F52EBBE1CE91B5A5. User reply “是” after checkpoint1df88952 approves that written design. This plan and its allowance remain PENDING_APPROVAL.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; branch codex/c12-b01-task9-coordinator; origin https://github.com/win202569/Talenro.git. Native method retained.
- L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2; P=.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner; Q=.superpowers/sdd/2026-11-01-offline-controller-recovery. Paths below use these exact aliases.
- Runtime source window exactly L/source-model.ps1, L/observation-contract.ps1, L/driver.ps1. Other28of31 sources, existing tests/oracles, all seven stops,155snapshots and archives remain unchanged. New auxiliary is outside runtime inventory.
- Request TWO auxiliary computations, RED then GREEN, cumulative ceiling11 with9 consumed. No extra runtime hosts/compiles/targets/snapshots/reviewers. Existing remainder4hosts/7targets/0compiles/31snapshots/conditional1review; cumulative runtime ceilings7hosts/3compiles/186snapshots. Consumed3hosts/3compiles/155snapshots/0targets/0reviews.
- Accepted writer-projection-green remains accepted only at its original boundary; do not rerun, refund, regrade against new sources or use its success as product acceptance. Contract candidate remains untested until actual verification.
- Auxiliary runs in the existing tool context, no child host, private module, profile/control body, writer, compilation or target. Only actual extracted definitions, pinned file reads, AST and pure predicates.
- First unexpected result stops, preserves actual first error and partial groups; no patch/retest, extra computation or partial-freeze retry. P/contract-call-model.stop.json always blocks. No IgnoreStop, arbitrary path/callback or old attempt replay.
- Keep original process ownership, timing, source/oracle/recorder predicates and all five false-capability flags. No main merge, force push, deletion, install, security change or broader acceptance.

## Review Focus

1. Initializer/cleanup resembles the case call: ExactDescriptor and OtherControlsExcluded distinguish literal parameter list Name.
2. Forged or ambiguous bindings authorize unrelated code: BindingIdentityRejected, MissingBindingRejected and AmbiguousBindingRejected require exact independent descriptor identity.
3. Stack lines drift within an enclosing function: WrongCallerLineRejected and BodyExtentRejected require exact invocation and literal body extent, not function range.
4. Unit authority leaks to runtime: RuntimeBindingAbsent and RuntimeBridgeRejected keep runtime context free of unit metadata and special fallback.
5. New controller validates old receipts against new bytes or replays a spent stage: HistoricalEpoch and RecoveryGate check snapshot provenance and strict current-stage facts.

## Task1 — freeze a finite regression and observe old-model RED

**Files:** create L/contract-call-model.tests.ps1; P/contract-call-model.approval.json,
contract-call-model.input.json, contract-call-model-red.intent.json and actual
receipt/tool records; Q/contract-call-model-progress.md. No runtime edits yet.

**Interfaces:** Invoke-ODCallModelRegression -Phase Red|Green, definitions only
when loaded. Receipt ordered keys Schema,Phase,TestSHA256,InputSHA256,SourceSHA256,
PostSourceSHA256,Groups,PASS,FAIL,NotRun,ExitCode,RootError,RecordingError.
Schema int1; Groups ordered name/status records; actual ErrorRecord projection
for errors, never synthetic success. Setup failures use exit2 and preserve the
actual first error, completed groups and recording error. Each phase's fixed
prefix is contract-call-model-red or contract-call-model-green.

- [ ] Verify clean documentation-checkpoint HEAD and accepted writer manifest SHA182B85886304BFE3CBFB248DDA70F2EB99DF3E2158F574DBE299519DE7CE0461. Validate its2251pins using snapshots/writer-projection-rebind for operational sources, original paths for other pins. Separately record current three-file hashes (contract candidate is deliberately different). Verify seven stops,155snapshots, actual writer final/streams/evidence/probe/journal, and absence of all new paths/repair-green outputs. Do not execute diagnostic definitions during preflight.
- [ ] Record approval with ordered keys Schema,BaseCommit,UserReply,Method,DesignSHA256,PlanSHA256,ParentSHA256,HistoricalStopSHA256,AddedAuxiliary,AuxiliaryCeiling,RemainingHosts,RemainingTargets,RemainingCompilations,RemainingSnapshots,RemainingReviewers,HistoricalFiles,BaselineSources. Values int1, actual HEAD/reply, Native, exact file hashes, parent182B8588…0461, stopAC592FFF3F7FE2671692352CF68C819967CC066527A937E0816FC5E9F07C71EB,2,11,4,7,0,31,1. HistoricalFiles binds original artifacts by path/hash; BaselineSources binds the three actual starting files.
- [ ] Author/freeze the regression and input before RED. Input ordered keys Schema,ApprovalSHA256,DesignSHA256,PlanSHA256,ParentSHA256,HistoricalStopSHA256,TestSHA256,BaselineSources,HistoricalFiles; no future candidate hash or receipt dependency. Actual extracted model functions must uniquely match AST definitions; do not dot-source the controller top level. Read actual saved IntentFailureNoWork data without executing its producer.
- [ ] Freeze these20ordered groups: ExactDescriptor, OtherControlsExcluded, NamedCallsUnchanged, UnknownVariableExcluded, ActualUnitStackAccepted, WrongSourceRejected, WrongFunctionRejected, WrongCallerLineRejected, BodyExtentRejected, MissingBindingRejected, BindingIdentityRejected, AmbiguousBindingRejected, RuntimeBindingAbsent, RuntimeBridgeRejected, FaultOperationPreserved, HistoricalEpoch, RecoveryGate, SourceWindow, FixedSlots, OriginalGuardsUnchanged. First group on the old model requires one descriptor, treating absent DynamicCalls as zero rather than triggering a missing-property error. Its explicit assertion throws exact RuntimeException Message/FQID 'OD: fixed owner-control descriptor missing'. Expected RED0PASS/1FAIL/19NotRun/exit1. Any other result stops.
- [ ] Commit/push frozen regression/input through a new portable preparation archive if needed for raw input recovery; never stage ignored raw files directly. CreateNew phase intent records computation10of11 before invoking Invoke-ODCallModelRegression -Phase Red once. Save all actual tool chunks and receipt; no missing-output reconstruction. RED consumes its slot even if unexpectedly interrupted.

## Task2 — implement three-source amendment and one GREEN

**Files:** modify only the three runtime files. Create P/contract-call-model-candidate.input.json and contract-call-model-green.*. Frozen regression unchanged.

**Interfaces:** Read-ODSourceModel keeps its signature and existing Functions,
Calls and Faults shapes; adds DynamicCalls array. Each descriptor has ordered
keys Source,SHA256,Owner,Variable,Line,Column,BodyStartLine,BodyStartColumn,
BodyEndLine,BodyEndColumn. Positions are integers from AST, Source absolute pinned
path, SHA256 its pinned hash, Owner Invoke-ODOwnerControls, Variable controls.
New Get-ODIntentCallBinding([Collections.IDictionary]$SourceModel,[string]$Source)
returns the unique descriptor or throws 'OD: unit call binding mismatch'.
New-ODIntentUnitContext adds UnitBindings.OwnerControlCall; runtime context does
not. Existing function signatures and public contract mismatch stay unchanged.

- [ ] In Read-ODSourceModel recognize only fixed record-boundary.tests.ps1, owner Invoke-ODOwnerControls, variable controls, literal case ScriptBlockExpressionAst with ordered parameter names exactly Name. Do not evaluate variable/module identity. Derive invocation and literal body extent from AST. Preserve all named Calls and existing aliases unchanged; unknown calls add no descriptor. Get-ODIntentCallBinding rejects zero/multiple descriptors and wrong source/hash/owner/variable; use it in the explicit unit constructor only.
- [ ] In OC-SourceError track the previous frame line as well as function/source. Only when ordinary named/factory resolution finds no call, inspect the explicit OwnerControlCall binding. Require exact descriptor keys/types, equality with the unique independently modeled descriptor, caller source/function/invocation line, previous function <ScriptBlock>, previous source equal descriptor source, previous line within literal body extent. Missing/foreign/ambiguous binding rejects with exact existing RuntimeException Message/FQID 'OD: observation contract mismatch'. Keep first fault/operation and all named-chain checks; do not catch arbitrary implementation errors as legitimate negatives.
- [ ] Add Driver-CallModelGate and Assert-ODCallModelFacts([Collections.IDictionary]$Facts). Facts ordered keys Schema,ParentSHA256,HistoricalStopSHA256,ApprovalValid,AttemptConsumed,NewStopPresent; values int1,parent182B8588…0461,stopAC592FFF…71EB,typed booleans true,false,false. All mismatches throw exact RuntimeException Message/FQID 'OD: contract call model recovery mismatch'. Gate binds approval/input/regression/receipts and immutable accepted writer originals. Historical projection/scope receipts use writer-projection snapshots for old source hashes; preserve older recovery maps and earlier snapshot rules, not current edited bytes.
- [ ] Add sole new Prepare grant contract-call-model-rebind, output still repair-green.frozen.json and snapshots/repair-green. Reject old writer preparations and old repair-green preparation route; keep RunOne contract-green/core/lifecycle/continuation slots unchanged. Preserve accepted writer namespace. New outer stop is contract-call-model.stop.json, CreateNew only; all seven old stops remain. Historical error-projection stop is recoverable only by exact approval/hash binding, not a generic bypass.
- [ ] Freeze candidate input ordered keys Schema,InputSHA256,CandidateSources,HistoricalFiles,ExpectedPinCount; int1, original input hash, three actual candidate hashes, immutable history map,2263. Driver must not embed its own candidate hash or containing input hash: candidate input is externally frozen and later manifest-pinned; regression freezes/checks it before/after GREEN. Add no future receipt references to it. Commit source/test/input checkpoint before GREEN; do not alter source/test bytes after GREEN.
- [ ] GREEN groups use actual mapper and pure consumer. ExactDescriptor compares every field to independently parsed pinned AST. OtherControlsExcluded enumerates initializer/cleanup and excludes both; NamedCallsUnchanged compares entire ordered Calls to the old model; UnknownVariableExcluded proves all other variable commands unauthorized. Pure binding helper rejects zero and duplicate descriptor copies without changing source files.
- [ ] For ActualUnitStackAccepted construct an explicitly test-only context from actual mapper data and the same pinned profile AST-relative call model used by New-ODIntentUnitContext, then call actual OC-SourceError with the untouched saved error and independently fixed IntentFailure expectation. Do not invoke New-ODRuntimeContext before repair-green exists or fabricate that manifest. Independently clone only test data for source/function/caller-line/body-line/missing/foreign/duplicate binding negatives; require exact contract mismatch, not arbitrary exception. FaultOperationPreserved mutates initial operation/position and rejects. RuntimeBindingAbsent inspects actual runtime constructor AST for no unit assignment/fallback; RuntimeBridgeRejected removes only OwnerControlCall from an otherwise valid test context and rejects. This auxiliary does not claim full runtime-context execution; the original contract host will exercise real constructors after freeze.
- [ ] HistoricalEpoch validates saved receipts against old snapshots; RecoveryGate executes actual extracted facts predicate with valid facts and each field independently invalid; SourceWindow compares all31runtime sources against accepted writer snapshots allowing only three; FixedSlots checks exact preparation/output namespace and spent-writer refusal; OriginalGuardsUnchanged compares ownership/termination/timing and strict error-location function definitions against the accepted snapshot. Do not invoke driver/profile bodies. Preserve actual first errors and stop on any unexpected result.
- [ ] CreateNew GREEN intent computation11of11; invoke Invoke-ODCallModelRegression -Phase Green exactly once. Require20PASS/0FAIL/0NotRun/exit0, null RootError/RecordingError and unchanged input/test/source hashes before/after. No additional full-scope computation, runtime host or compiler is authorized here.

## Task3 — one contract freeze, original bounded chain, recoverable handoff

**Files:** create P/repair-green.intent.json/frozen.json, snapshots/repair-green,
actual original stage records; new L/portable-contract-call-model-freeze.json and
L/portable-contract-call-model.json; docs/roadmap/2026-11-02-offline-contract-call-model.md,
current-status.md and repository-recovery.md.

- [ ] Prepare inherits2251pins, changes exactly three operational source pins and appends exactly12unique paths listed below, yielding2263pins/31sources. Preserve inherited metadata and both historical recovery maps byte-equivalent as data. Add ContractCallModelRecovery ordered fields Schema,ParentSHA256,ApprovalSHA256,DesignSHA256,PlanSHA256,TestSHA256,InputSHA256,CandidateInputSHA256,RedFinalSHA256,GreenFinalSHA256,HistoricalStopSHA256,Consumed,Ceilings. Counters disclose auxiliary11/hosts3/compiles3/snapshots155 before preparation and ceilings11/7/3/186; new current chain MaxHosts4/MaxChildren7/MaxCompilations0/MaxReviewers1. Do not reset consumed history.

The12new pin paths (check each absent from parent before preparation):
P/writer-projection-repair.frozen.json; P/error-projection.stop.json;
docs/superpowers/specs/2026-11-02-offline-contract-call-model-design.md;
docs/superpowers/plans/2026-11-02-offline-contract-call-model.md;
P/contract-call-model.approval.json; L/contract-call-model.tests.ps1;
P/contract-call-model.input.json; P/contract-call-model-candidate.input.json;
P/contract-call-model-red.final.json; P/contract-call-model-green.final.json;
P/writer-projection-green.final.json; P/error-projection-writer.tool-record.json.
Actual writer streams/evidence/journal remain bound by their existing final and
explicit HistoricalFiles map. No current source references a containing input hash.

- [ ] Run driver -Action Prepare -Grant contract-call-model-rebind -PreviousFrozenSHA256 182B85886304BFE3CBFB248DDA70F2EB99DF3E2158F574DBE299519DE7CE0461 once. Require2263pins, exactly31new snapshots,total186, exact three-source window. CreateNew intent precedes outputs. Stop on missing/partial artifacts, never retry. Verify actual returned manifest hash, preserve source bytes.
- [ ] Before runtime, create NEW portable-contract-call-model-freeze.json covering all actual P/Q files/directories/zero-byte files with LF container and exact embedded original bytes. Verify decoded lengths/SHA256/bytes and inventory. Update status/recovery, explicit-file stage, staged raw-byte equality and diff checks, commit/push existing branch and verify remoteSHA==HEAD.
- [ ] Run driver -Action RunOne -Grant contract-green -FrozenSHA256 <actual repair manifest SHA> once. Require original52PASS including actual saved incomplete-unit report and real context construction, exit0, Exact/EvidenceValid, actual recorder and source checks. Auxiliary20PASS is not a substitute. First unexpected result blocks all subsequent stages.
- [ ] Only after exact contract acceptance run core once (25PASS/22rows/0targets), then lifecycle once (17PASS/8rows/3targets), then continuation once (16PASS plus12seamPASS/15rows/4targets), same manifest and original grants. Preserve120000ms host cap except continuation150000, sequence140000/cleanup4000, owned exit/shared drain2000, hold45000/normal receipt2000/timeout200/seam110000; session/sinks32,Ack2,observations32,secondaries8. No extra compile or rerun.
- [ ] Only after all actual runtime stages pass use the original conditional single readonly runtime reviewer; no earlier or extra review. Product/I1/I2 status changes only if original acceptance criteria truly pass, not from documentation or auxiliary success.
- [ ] On success OR stop preserve complete actual tool chunks/errors/partial records, create NEW portable-contract-call-model.json and verify decoded equality. Keep all old archives/stops. Update execution/status/recovery with actual counts and remaining blockers; commit/push necessary files only, verify remote equality and clean worktree. No main merge.

## Self-review and approval boundary

Tasks1–2 cover fixed AST identity, unchanged named calls, strict negative tests,
unit/runtime separation and immutable historical provenance. Task3 covers exact
pin count, original output selectors, finite remaining runtime chain and recovery.
All five review-focus cases have named tests; no source hash/input cycle or new
runtime namespace is required. New tests and inputs freeze before computations;
unexpected outcomes stop rather than expanding the plan. This planning turn
does not execute any listed checkboxes. Approval must cover this plan and its
two auxiliary computations plus conditional use of the original remaining chain.
