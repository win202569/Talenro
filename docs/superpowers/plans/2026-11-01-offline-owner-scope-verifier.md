# Owner Scope Verifier Boundary Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans natively after explicit approval. No implementation agents, extra reviewer, suite rerun or cleanup.

**Goal:** Correct typed function comparison and retain complete failed-case evidence before another bounded scope GREEN.

**Architecture:** Test the existing broken comparison first, then version the full auxiliary with a pure typed checker and a protected case runner. Bind their independently verified results into the existing controller and source epoch without rewriting history.

**Tech Stack:** Existing Windows PowerShell7.6.5 Core x64, AST/JSON/hash APIs, no installs.

**Spec:** [Verifier design](../specs/2026-11-01-offline-owner-scope-verifier-design.md), [scope plan](2026-11-01-offline-owner-scope-repair.md), [signature supplement](2026-11-01-offline-owner-scope-signature.md). Only verifier boundaries, current receipt/provenance bindings and the explicitly requested allowance change.

Status: PENDING_DESIGN_PLAN_AND_THREE_COMPUTATION_APPROVAL. No execution yet.

## Global Constraints

- Same existing isolated worktree/branch/origin and L/P/Q. Preserve old sources/tests/approved docs, all five stops, actual receipts and archives. No regrading old failures or missing-field reconstruction.
- New auxiliaries outside31runtime inventory, never imported by capture/profile. Only further runtime source edit is L/driver.ps1 for strict binding/provenance/new stop; four other candidate files byte-identical to0e17eba9. Original five-source window relative to historical parent remains exact.
- Additional3computations only: verifier RED once, verifier GREEN once, full scope GREEN once. Old3consumed, cumulative6; max2private modules each new computation. No extra host/compile/target/snapshot/reviewer. Keep5hosts/7targets/1compile/conditional1review and93of155snapshots, original process/time/capability/oracle gates.
- No profile body or record writer call, Add-Type, child launch or exploratory module probe in auxiliary computations. Literal AST fixtures and controlled runner fault bodies are permitted only inside the approved computations.
- First unexpected result stops; no source/test correction, repeat, re-freeze or conditional reviewer before all actual runtime stages pass. Save actual chunks/diagnostics. No main merge, security/install change, deletion or history rewrite.

## Review Focus

1. Correct function body is falsely rejected by comparing different AST levels: SameDefinitionsAccepted.
2. Body normalization hides changed inline parameters/name/node kind: ChangedBodyRejected/ChangedInlineParameterRejected/UnexpectedAstKindRejected/UnexpectedNameRejected.
3. An assertion masks the real first error: FailurePreservesOriginal retains actual reference until projection.
4. Early exit loses PASS history or executes later groups: CompletedGroupsPreserved/LaterCasesNotExecuted.
5. New GREEN borrows old test identity or ignores a newer stop: fixed per-phase identities, historical-source validation and new-stop refusal.

## Task1 — freeze verifier regression and observe the old checker fail

**Files:** create L/owner-scope-verifier-regression.tests.ps1;
P/owner-scope-verifier.approval.json, owner-scope-verifier-regression.input.json,
owner-scope-verifier-red.intent.json/actual records; Q/owner-scope-verifier-progress.md.

**Interface:** Invoke-ODOwnerVerifierRegression -Phase Red|Green, fixed paths only.
Normal receipt: Schema,Phase,TestSHA256,InputSHA256,CodeSHA256,FixtureSHA256,Groups,
PASS,FAIL,NotRun,ExitCode,RootError. Unexpected output is a distinct diagnostic,
not a normal final; source hashes must match before/after phase. FixtureSHA256
binds the six actual sources used: record-boundary.tests.ps1,capture.ps1,
case-module.ps1 plus runtime private/write-boundary.ps1,record-owner.ps1,profiles.ps1.
No unit-test dependency on the later edited driver bytes.

- [ ] Verify baseline/history/candidate hashes and absence of all new paths. Record actual approval/design/plan hashes and added3computations, no runtime addition. Establish the explicit ledger; no Bash/default task-done rerun.
- [ ] Author and commit regression before RED. Freeze eight ordered groups: SameDefinitionsAccepted,ChangedBodyRejected,ChangedInlineParameterRejected,UnexpectedAstKindRejected,UnexpectedNameRejected,FailurePreservesOriginal,CompletedGroupsPreserved,LaterCasesNotExecuted. Freeze actual test hash and six fixture hashes in input; old frozen auxiliary SHAFA11EF9E…C3513 and old stopped GREEN are explicit historical bindings.
- [ ] RED initializes actual private fixture once from current candidate initializer and actual four source definitions, exactly as earlier scope harness, without invoking those functions. Uniquely extract the ORIGINAL DefinitionsExact case body and ODS-Need from frozen owner-scope-signature.tests.ps1 AST. Execute that exact case body with actual definitions/commands, not an imitation old predicate.
- [ ] Consume RED intent and run the regression once. Require actual RuntimeException Message/FQID 'Owner scope regression: actual body unchanged Capture-Value',1FAIL/7NotRun/exit1, no later cases. Save actual original exception and both comparison operands/node kinds before invoking old checker. Different failure or unexpected pass stops; no mutation of old checker or replay.

## Task2 — implement new verifier boundaries; one verifier GREEN

**Files:** create L/owner-scope-verifier.tests.ps1 and
P/owner-scope-verifier-scope.input.json; old auxiliaries/input stay frozen.
Unit GREEN prefix owner-scope-verifier-green. No driver edits before this GREEN.

**Interfaces:** Assert-ODImportedDefinition(Expected,Loaded,Comparison) and protected
case runner defined only in new full auxiliary; definitions-only loading performs
no computation. Checker and runner behavior are specified by the design. Full
entry retains Invoke-ODOwnerScopeRegression, but accepts only -Phase Green.
Successful full receipt retains original nine-group schema/order; unexpected
diagnostic adds FailedGroup,CompletedGroups,RootError,Comparison,RecordingError,
TestSHA256,InputSHA256,SourceSHA256 and PostSourceSHA256, with actual values.

Checker parameters are objects with explicit node-kind checks (not coercing typed
parameter binding); Comparison is a private holder with Value initially null.
Runner signature: Invoke-ODVerifierCase([string]Name,[scriptblock]Body,[object]State)
returns no output. Private State fields: CurrentGroup,Groups (ordered result list),
RootError (retained live ErrorRecord or null),Comparison (the holder),RecordingError
(live ErrorRecord or null),Stopped (bool). Project references only for diagnostics;
do not serialize/restore them for identity assertions or substitute wrapper errors.

- [ ] Implement checker with strict FunctionDefinitionAst inputs, exact body/inline-parameter comparisons and allowlisted script-qualified name. Capture operands/kinds BEFORE comparison. Do not replace source code under test or use stripped strings as an AST surrogate.
- [ ] Implement case runner and protected phase setup. State owns ordered actual group results, first actual ErrorRecord and comparison context. On first failure retain it, mark failing group, prevent later bodies; phase produces a diagnostic exit2. Diagnostic/projection failure retains both original and recording errors. Keep nine original assertions/semantics and private-module inventory; only comparison mechanics/result preservation change.
- [ ] Static inspect, commit/freeze new full auxiliary and its input before unit GREEN. Full input binds new auxiliary hash, regression test/input hashes, old exact scope RED/test/input, old failed GREEN/stop and original baseline source map. Do not bind the not-yet-existing unit GREEN receipt here: that would create a cycle.
- [ ] Consume unit GREEN intent with full-input identity. Run frozen regression -Phase Green once:8PASS/0FAIL/0NotRun/exit0. SameDefinitionsAccepted uses all four actual loaded/source functions. Negative AST inputs are independently literal: changed return body; inline parameter type/default; non-function ScriptBlockAst; foreign name/qualifier. Each rejects only the exact definition-mismatch exception.
- [ ] Runner controls use fixed scriptblocks: a preceding passing case, then a retained real controlled throw, then a side-effect counter case. Require original ErrorRecord reference before projection, preserved prior PASS/current FAIL/later NotRun and counter0. Retain first error if diagnostic projection is deliberately faulted; do not disguise synthetic controls as real writer events. Complete actual tool records retained; no post-GREEN auxiliary edits.

## Task3 — strict binding, one full scope GREEN and original bounded resume

**Files:** only L/driver.ps1 runtime edit; frozen new auxiliaries unchanged.
New full phase prefix owner-scope-verifier-scope-green; failures use fixed
P/owner-scope-verifier.stop.json. New portable L/portable-owner-scope-verifier.json.

- [ ] Replace current signature-GREEN requirement with exact new full nine-group GREEN and require both new verifier phase receipts. Keep historical exact scope RED separately bound to FA11EF9E…C3513/input511274F0…AE792/final307D8BC0…79C1A; never require the nonexistent old signature GREEN final. Bind old GREEN diagnostic/tool record/stop as failed history, not as success.
- [ ] Validate new approval, auxiliary/input identities and historical signature stop SHA1D8844C21D1F0559CCD705D4394519E30D31BC72524F9671D821F46FF4118F38. Preserve existing source/attempt/prior-final/process/recorder predicates. New verifier stop always blocks. Other four candidate sources must remain baseline-exact; no public IgnoreStop/path/callback.
- [ ] Consume full scope GREEN intent and run new full auxiliary once in existing tool context. Require9PASS/0FAIL/0NotRun/exit0, no Unexpected, null RootError, exact five current source hashes before/after and all original isolation/consumer/gate assertions. Any unknown/partial diagnostic stops; no repair or repeat.
- [ ] Scope source freeze retains the thirteen previous intended provenance pins and adds exactly eight: THIS design,THIS plan,new approval,regression script,regression input,new full auxiliary,new full input,historical signature stop. New count2236=2215+13+8, exactly original five source changes relative to parent. OwnerScopeRecovery keeps actual old RED and failed-GREEN identities, adds VerifierRecovery map binding new design/plan/approval, each test/input, verifier RED/GREEN finals and full scope GREEN final. Replace old nonexistent GreenFinalSHA256 with an explicitly historical failed diagnostic hash; never fabricate it. Map declares cumulative6auxiliary computations; phase finals are explicit hash-bound records, not source pins.
- [ ] Only then original unique Prepare writer-scope-rebind/31snapshots, checkpoint, one writer-scope-green host/one compile, later contract freeze/31snapshots and four original later hosts/7targets under unchanged acceptance. No host/snapshot/review addition. Single runtime reviewer remains conditional on every actual stage passing.
- [ ] Success or stop: archive all actual P/Q records/directories/zero-byte files to NEW portable, LF container/decoded byte equality; old archives unchanged. Update execution/current-status/recovery. Explicit necessary-file commit/push, staged raw bytes==disk, remote SHA==HEAD, clean worktree. Candidate/auxiliary success is not I1/I2/product acceptance.

## Self-review and approval gate

Same-level AST/inline parameter rules, retained first error/partial groups,
independent negative controls and historical/current phase identities are explicit.
Unit input precedes implementation; full input precedes unit GREEN and references
unit input but not its future receipt; later provenance binds all actual finals,
avoiding cycles. No source/test change or verification computation occurred during
planning. Approval must cover design/plan,3additional auxiliary computations and
conditional resumption of existing unchanged runtime budget. Native method retained.
