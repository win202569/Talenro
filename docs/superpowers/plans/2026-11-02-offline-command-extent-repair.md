# Command Extent Local Repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans natively after execution approval. No agents or reviewer. Steps use checkbox syntax.

**Goal:** Correct the identified ConvertFrom-Json fault extent and prove its local acceptance/rejection boundary in exactly two auxiliary runs.

**Architecture:** Freeze a definitions-only ten-group regression before RED. After the exact expected failure, insert one narrow early return in Get-ODFailureExtent and bind the candidate independently for GREEN. Preserve the old runtime manifest and all historical outcomes unchanged.

**Tech Stack:** Existing Windows PowerShell, AST, JSON and SHA256; no installation.

**Spec:** [Approved local design](../specs/2026-11-02-offline-command-extent-repair-design.md), SHA7456BBBBF349B87B17C51000612B6F5EE599747E5E73C79B9BBB0BDF3035798F. User “是” after8e8b2147 approves the written design. This plan/execution allowance remains PENDING_APPROVAL; retain original spec bytes.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; branch codex/c12-b01-task9-coordinator; origin https://github.com/win202569/Talenro.git. Require original absolute root; do not silently relocate saved error paths.
- L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2; P=.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner; Q=.superpowers/sdd/2026-11-01-offline-controller-recovery; H=.superpowers/sdd/2026-10-31-offline-diagnostic-runtime-migration.
- Only L/source-model.ps1:Get-ODFailureExtent may change operational behavior. All other30operational sources, old tests/oracles, manifests, stops and archives remain byte-identical. No production hardcoded coordinates, extra fault candidates or relaxed predicates.
- Request2additional auxiliary computations: RED16/GREEN17, ceiling17. Consumed15auxiliary/4hosts/3compiles/186snapshots/0targets/0reviews. Added hosts/targets/compiles/runtime snapshots/reviewers ZERO. Original remaining3hosts/7targets/conditional1review stay frozen.
- Native execution only, no module/profile/control/writer bodies, arbitrary callbacks from data, child hosts, separate probes, full-suite runs, diagnosis15 replay, contract-green replay or automatic retry. Generic extra verification/review/cleanup would exceed this expressly bounded scope.
- First unexpected result freezes all later work in new P/command-extent.stop.json. Each intent consumes its attempt even on setup/recording failure. Expected RED permits only the planned repair and GREEN; successful GREEN also ends the allowance.
- Candidate source will intentionally mismatch its old runtime pin. Do not modify the manifest or run/bypass the driver. A separate future runtime-freeze proposal is required. Product remains NOT_ACCEPTED.

## Review Focus

1. GREEN accidentally reuses the old helper: load baseline and candidate in separate local scopes, bind both files independently and assert exact command extent.
2. Too-broad correction changes methods/throws: compare their full modeled records against frozen baseline in groups8–10; statically preserve old ancestor-walk bytes.
3. Wrong coordinates start passing: groups3–7 call the unchanged predicate on clones and require its exact rejection signature.
4. Expected RED is actually setup failure: classify setup separately and require first-group location rejection, with all remaining groups NotRun.
5. Candidate becomes a fake runtime freeze: separate immutable-history and current-source maps; preserve old manifest bytes and explicitly report its mismatch.

## Task1 — immutable regression and one RED

**Create:** L/command-extent.tests.ps1; P/command-extent.approval.json;
P/command-extent.input.json; P/command-extent-red.intent.json/final.json/tool-record.json;
Q/command-extent-progress.md; L/portable-command-extent-preparation.json.

**Interfaces:** Invoke-ODCommandExtentRegression -Phase Red|Green returns a single
ordered receipt; importing the new test defines functions only. Test helpers use
CE- prefix and must not replace OC- predicates. Get-CEContext([string]$ModelPath,
[string]$Repo,[Collections.IDictionary]$Manifest) evaluates the allowlist below in
a new local scope and returns only the historical context. The phase's actual
OC-SourceError calls use a separate local scope containing unchanged predicates.

- [ ] Verify clean approved checkpoint, original root, nine stops, no new command-extent outputs and the source31inventory. Re-read the saved diagnosis rather than invoking it. Confirm all anchors below; a discrepancy blocks before RED.
- [ ] Create approval ordered fields Schema,BaseCommit,UserReply,DesignSHA256,PlanSHA256,AddedAuxiliary,ComputationCeiling,Method; values1,actual approval checkpoint/reply,actual hashes,2,17,Native. Write ledger with this plan identity and task statuses.
- [ ] Author the ten groups below and real receipt/error recording. Preserve extracted helper function text; prefix newlines when evaluating definitions to retain original OC-SourceError call-site line33. AST parsing must find exactly one top-level definition per allowed helper. Static inspection only before RED; no trial test invocation.
- [ ] Freeze input ordered fields Schema,ApprovalSHA256,DesignSHA256,PlanSHA256,TestSHA256,ParentSHA256,DiagnosisFinalSHA256,HistoricalFiles,BaselineSources,HelperNames. HistoricalFiles is an explicit relative-path/hash map. BaselineSources has exactly the31current operational paths. Keep L/source-model.ps1 ONLY in BaselineSources, not HistoricalFiles: GREEN must not expect its old hash at the edited live path. Its preserved snapshot remains an immutable historical input. No output/self-hash cycle.
- [ ] Preserve preparation P/Q plus needed H originals in a NEW portable archive, verify decoded bytes/hashes/directories/zero-byte entries and push explicit files before RED. Freeze test/input bytes for both phases.
- [ ] CreateNew RED intent with Schema,Phase,AttemptConsumed,Computation,Ceiling,TestSHA256,InputSHA256,PhaseInputSHA256,SourceSHA256; values1,Red,true,16,17,actual hashes,null,31source map. Invoke once in the existing tool context:

```powershell
. ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/command-extent.tests.ps1
$receipt=Invoke-ODCommandExtentRegression -Phase Red
```

- [ ] Persist actual returned receipt and complete tool input/results, without reconstructing exception text. Require0PASS/1FAIL/9NotRun/ExitCode1, SavedErrorAccepted failed, actual RuntimeException with Message/FQID 'OD: observation contract mismatch' and OC-SourceError call site33. SetupError/RecordingError null and all pre/post hashes identical. Unexpected pass, wrong failure or setup error creates stop and forbids Task2.

### Exact input sources and helper boundary

Bind these existing anchors verbatim:

- P/repair-green.frozen.json: A9949DD46C7F928CFE7DB33F092AF0497E6F89556E8F359A8A55CB4484BB1556.
- P/fault-location.final.json: F911770E1FC860017BD7EF1114F88BBD7700173F465F7B7E9571D714177137CE.
- P/fault-location.input.json: C7B18E568EDA29F5DDAEE6FABB795378E4206FADE64B742F3E15AA0364FC6B61.
- L/portable-fault-location.json: EB5B1E3743889DB1DD695D23CA2131E5E49380799423665CA9D5942FA2825BB3.
- H/runtime-green.evidence.json: 58B2D81047526EFB2BEE3FD8155DFAF493FC3FD6ECAAE32B0E7ED19C7DB66EA0.
- Baseline live source-model and P/snapshots/repair-green/L/source-model.ps1: 1C7897E486D43601D4D0E963A566737C97E33FC8C88BF7E4ECF83348A5A61418.
- L/observation-contract.ps1: B68A9D8C9674692C8B75264CDB687E92063911254F64F5D69124E03265E3CB12.

Materialize the remainder from the verified fault-location.input.json Files map,
excluding only live L/source-model.ps1 from immutable history. This binds the
historical14source inventory, expected-observations/output, historical manifest,
old final/tool/streams, nine stops, helper snapshots and old planning artifacts.
Add this design/plan/approval and P/fault-location.intent.json/tool-record.json,
hashing their actual bytes. Frozen prior input maps are provenance, not permission
to replay their old computations or recursively demand their old live-source hash.
Every currently read historical source must still match the real repair manifest
Present pin; duplicate pins must agree. No in-memory replacement runtime pins.

Allowed source-model functions: Get-ODPinnedSHA256, Get-ODSourceOwner,
Get-ODFailureExtent, Read-ODSourceModel, New-ODFixedContext,
New-ODHistoricalContext. Allowed pure functions: OC-Need, OC-Keys, OC-Integer,
OC-Expand, OC-SourceError. No OC-Match/full-row suite, unit/runtime constructor,
driver or profile execution. Read historical source bodies only as AST data.

Construct the candidate historical entry-migration-core context with actual
candidate helpers and the unchanged repair manifest. This context reads only
unchanged historical sources, not the edited live helper as a runtime source.
Extract row8.Value.View.PrimaryError from the unique saved ProfileReturn, using
the original Depth60 JSON roundtrip; extract expected $error from the literal
context row8 at the same path. Verify both equal diagnosis FirstMismatch data.
For GREEN groups8–10 construct the baseline context from the preserved helper
snapshot in a different local scope. Never copy the candidate model to manufacture
the baseline. Treat unexpected setup errors as exit2, not a valid RED.

### Ten frozen groups, in order

1. **SavedErrorAccepted:** OC-SourceError(actualError,expectedError,candidateContext)
   emits nothing and returns without exception. No shortened/rebuilt error stack.
2. **CommandExtent:** independently parse the historical sinks AST, select the
   unique Read-LSLedger ConvertFrom-Json CommandAst matching saved36:50. Require
   one modeled candidate, equality of Line/Column/EndLine/EndColumn/Text with this
   AST, and no duplicate old36:9 candidate. Literal coordinates are assertions
   about the frozen fixture only, never production logic.
3. **WrongColumnRejected:** clone actual error; replace only the initial Position
   header's char50 with char9. Require the exact rejection signature, no output.
4. **WrongLineRejected:** clone actual error; change initial Position line36 to35
   and first Stack frame line36 to35, leaving all other data unchanged. Require
   exact rejection, not merely a stack/position disagreement.
5. **WrongOperationRejected:** clone expected error; add Operation='ReadAllText',
   keep actual saved error unchanged and require exact rejection.
6. **WrongFunctionRejected:** clone expected error; Function='Write-LSLedger',
   keep actual unchanged and require exact rejection.
7. **WrongSourceRejected:** clone expected error; Source='Core' (existing pinned
   different source), keep actual unchanged and require exact rejection.
8. **MethodsUnchanged:** for each of14historical ASTs, enumerate fault nodes in
   the same traversal order as Read-ODSourceModel. For indexes whose node is
   InvokeMemberExpressionAst, compare the complete baseline/candidate fault
   records exactly; assert at least one such node exists (no vacuous pass).
9. **ThrowsUnchanged:** same ordinal comparison for ThrowStatementAst, requiring
   at least one throw. Preserve every field including extent Text.
10. **ModelBoundaryUnchanged:** identical source keys/SHA256/Functions/Calls/
    DynamicCalls and fault counts/order for all14models. Non-command faults must
    be identical; each ConvertFrom-Json fault's Function/Operation stays equal
    while its extent equals the independent raw CommandAst. Derive AST owner
    independently, not by calling the helper under test. Assert this fixture has
    at least one command extent that actually differs from baseline.

All negative cases require Type=System.Management.Automation.RuntimeException
and Message/FQID='OD: observation contract mismatch'. A different exception is a
failure with the original error preserved. Stop evaluating groups after first
failure and mark every later group NotRun.

## Task2 — narrow repair, one GREEN and final preservation

**Modify:** only L/source-model.ps1 inside Get-ODFailureExtent.
**Create:** P/command-extent-candidate.input.json; P/command-extent-green.intent.json/
final.json/tool-record.json; L/portable-command-extent-candidate.json;
L/portable-command-extent.json; docs/roadmap/2026-11-02-offline-command-extent-repair.md.
**Update:** Q/command-extent-progress.md, current-status.md, repository-recovery.md.

- [ ] Only after exact RED, insert an early return when Node is CommandAst and
  GetCommandName() -ceq 'ConvertFrom-Json', returning Node.Extent. Leave the old
  ancestor walk and all other bytes unchanged. Static removal of this insertion
  must recover the complete baseline source bytes; all other30source hashes
  unchanged. Parse the candidate, but do not run an extra probe.
- [ ] Create immutable candidate input ordered fields Schema,InputSHA256,
  RedFinalSHA256,CandidateSources; values1,original input hash,actual RED final
  hash,31source map with exactly one changed entry. Do not update the original
  input/manifest or snapshots. Preserve candidate source in Git, candidate input
  and RED/tool records in new candidate portable archive, verify and push before
  GREEN. This checkpoint is explicitly not compatible with the old runtime freeze.
- [ ] CreateNew GREEN intent with the same schema as RED, values Green/17/17 and
  PhaseInputSHA256=candidate-input hash. Verify immutable test/input/history and
  candidate sources, then invoke once:

```powershell
. ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/command-extent.tests.ps1
$receipt=Invoke-ODCommandExtentRegression -Phase Green
```

- [ ] Require10PASS/0FAIL/0NotRun/ExitCode0 with RootError/SetupError/RecordingError
  null, exact ten-group order, zero unexpected output and all pre/post hashes
  equal. Save actual final and tool output. Any failure creates the new stop;
  no patch/retest. Even success ends this scope; no full22row/52group verification.
- [ ] Preserve actual final P/Q plus required H inputs in NEW final portable
  archive; verify decoded bytes/lengths/hashes/directories/zero-byte entries.
  Update report/status/recovery with consumed counts, local result and unchanged
  runtime blockers. Explicit staging, staged raw-byte equality, diff checks,
  commit/push existing branch, remoteSHA==HEAD; no main merge or acceptance claim.

### Receipt and error contract (both phases)

Ordered fields: Schema,Phase,TestSHA256,InputSHA256,PhaseInputSHA256,
SourceSHA256,PostSourceSHA256,HistoricalSHA256,PostHistoricalSHA256,Groups,
PASS,FAIL,NotRun,ExitCode,RootError,SetupError,RecordingError.
Schema int1; sources/maps use repository-relative paths; Groups entries Name,
Status (PASS/FAIL/NotRun) in fixed order. Errors preserve Type,Message,ErrorId,
Position,Stack from actual ErrorRecord; no synthesized exception signatures.
Retain original first group failure separately if subsequent bookkeeping fails.
RED expected failure exit1; successful GREEN exit0; setup/recording errors exit2;
any other failed group exit1 but is unexpected and stops further execution.
Do not refill missing hashes after a failed computation. Outer tool capture must
preserve partial output if the receipt cannot be written. All output files are
CreateNew; new stop retains Phase,actual errors,consumed count,Remainder=FROZEN.

## Self-review and handoff

Task1 covers fixed real inputs, exact RED and immutable history. Task2 covers the
single insertion, independent candidate binding, all ten checks and preservation.
Review-focus risks are covered by setup/phase checks and groups2–10. Baseline and
candidate constructors never share helper scopes; neither changes runtime pins.
No test, source edit or computation occurred during planning. Approve this plan
and two new auxiliary computations before native implementation. All broader
runtime work remains separately gated.
