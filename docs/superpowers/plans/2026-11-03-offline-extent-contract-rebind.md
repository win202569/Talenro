# Extent Contract Rebind Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bind the repaired source to one new immutable epoch and consume at most one full contract host without reopening old attempts.

**Architecture:** Keep semantic Grant contract-green and separate its physical attempt slot through an explicit extent-contract epoch. Coordinate four binding files, preserve parent/writer lineage, then freeze once and run once.

**Tech Stack:** Windows PowerShell orchestration; existing pinned PowerShell 7.6.5 host; JSON/SHA256 receipts.

**Spec:** ../specs/2026-11-03-offline-extent-contract-rebind-design.md

Status: PLAN_PENDING_EXECUTION_APPROVAL. User approved design at checkpoint
2c58fb47a988d7527bf87af5d16e05031d68d9a5 with “是”. Original design bytes remain
unchanged. This plan grants no execution by itself. Native, no delegated reviewer.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; branch codex/c12-b01-task9-coordinator.
- P=.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner; Q=.superpowers/sdd/2026-11-01-offline-controller-recovery; L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2.
- Added allowance TWO auxiliary runs RED18/GREEN19, ONE Prepare, ONE contract host,31snapshots; zero targets/compilations/reviewers. Current17aux/4hosts/3compiles/186snapshots; maximum after scope19/5/3/217. Downstream3hosts/7targets/conditional1review remain frozen.
- New epoch extent-contract; Prepare grant extent-contract-rebind; runtime semantic Grant contract-green; attempt slot extent-contract-green.
- New manifest P/extent-contract.frozen.json; snapshots/extent-contract; stop extent-contract.stop.json. Every action uses CreateNew intent; no retries or overwrites.
- Parent repair-green SHA A9949DD46C7F928CFE7DB33F092AF0497E6F89556E8F359A8A55CB4484BB1556.
- Accepted writer manifest SHA182B85886304BFE3CBFB248DDA70F2EB99DF3E2158F574DBE299519DE7CE0461; never replay it.
- Preserve49suite cases plus3recorder groups, original oracles/predicates, nine stops, all raw records,120000ms cap,2000ms exit/drain and five false flags.
- No install, privilege/environment change, arbitrary path selector, fallback, main merge or product acceptance. First unexpected result ends execution; preserve partial records.
- Only four operational files may change. Tests/mechanical preservation helpers below are additional non-operational artifacts, not added SourcePaths.

## Review Focus

- Missing/invalid epoch cannot silently select old or new runtime: groups EpochRejected and OldRouteBlocked.
- Partial files/directories consume an attempt even without final receipt: SpentRejected and StopRejected.
- Writer identity is not the new direct parent: ParentRejected and WriterRejected.
- Duplicate pins must agree before deduplication: PinConflictsRejected.
- Nested evidence/journal consumers must use the same slot as capture: SlotAgreement and GuardsUnchanged.

## File map and interfaces

Modify only L/driver.ps1, L/capture.ps1, L/observation-contract.tests.ps1,
L/source-model.ps1 among31operational files. Create L/extent-contract.tests.ps1.
Create Q/prepare-extent-contract.ps1, Q/run-extent-contract.ps1,
Q/preserve-extent-contract.ps1 and Q/extent-contract-progress.md; preserve these
ignored helpers through portable archives, never stage raw Q/P files.
Create L/portable-extent-contract-{preparation,candidate,freeze,final}.json.
Create docs/roadmap/2026-11-03-offline-extent-contract-rebind.md; update current-status
and repository-recovery at each handoff.

Exact new/extended interfaces:

- driver Get-ODAttemptSlot([string]$Grant,[string]$Epoch='legacy') -> string.
  legacy behavior unchanged; extent-contract accepts only contract-green and
  returns extent-contract-green; all other epoch/grant pairs throw.
- source-model Get-ODEpochManifestName([string]$Epoch='legacy') -> string:
  legacy=repair-green.frozen.json; extent-contract=extent-contract.frozen.json;
  unknown/case variants throw. No filesystem probing to choose.
- source-model Assert-ODExtentLineage([Collections.IDictionary]$Manifest,
  [Collections.IDictionary]$Parent,[string]$ParentSHA256,[string]$WriterSHA256) -> void.
  Validate fixed parent SHA, parent.PreviousFrozenSHA256=writer SHA,
  new PreviousFrozenSHA256=parent SHA and recovery map identities; throw otherwise.
- driver Assert-ODExtentContractFacts([Collections.IDictionary]$Facts) -> void:
  require Schema1, ApprovalValid=true, HistoryValid=true, CandidateValid=true,
  ReceiptsValid=true, AttemptConsumed=false, NewStopPresent=false.
- driver Driver-ExtentContractGate -> validated info object (Approval,Input,
  Candidate,Parent,RedSHA256,GreenSHA256); reads actual artifacts, constructs facts.
  Never use caller-supplied true flags as the filesystem gate.
- Append [string]$Epoch='legacy' to New-ODFixedContext, New-ODRuntimeContext,
  New-ODIntentUnitContext; pass it explicitly through runtime/unit paths.
  New-ODHistoricalContext retains its signature and fixed historical selection.
- Add validated Epoch=legacy|extent-contract to driver/capture entry params and
  Invoke-ODContractTests. Reject extent-contract for other Grants before side effects.
- Propagate Epoch through Driver-UnitEvidence, Driver-Recorder, Driver-Evidence
  and Driver-RequireFinal as a trailing defaulted parameter and their calls.
  Every Get-ODAttemptSlot call in these consumers receives Epoch; accepted writer
  checks explicitly use legacy. Capture passes Epoch into suite; suite passes it
  into manifest selector and New-ODIntentUnitContext.
- tests Invoke-ODExtentContractRegression -Phase Red|Green -> recorded ordered
  result with Groups,RootError,SetupError,RecordingError,pre/post hashes,exit code.
  It does not launch processes, compile, invoke whole driver/capture/suite or call
  the old Invoke-ODCommandExtentRegression.

## Task 1: Freeze regression inputs and obtain expected RED18

**Files:** New test/helper/input/approval/preparation archive above.
**Consumes:** Approved spec/plan; actual local16/17 and diagnosis15 artifacts.
**Produces:** Immutable test/input identities and actual RED18 receipt.

- [ ] Confirm execution approval specifically covers this plan and new finite allowances; record reply, spec/plan hashes, baseline commit and counters in P/extent-contract.approval.json.
- [ ] Read31source hashes, all inherited2276pins using old snapshots for changed operational bytes, nine stops and actual15/16/17 receipts. Build P/extent-contract.input.json with baseline hashes, HistoricalFiles, immutable test/helper hashes and exact append list below. No diagnostic function invocation during preparation.
- [ ] Implement definitions-only regression with these22ordered groups:
  BindingAvailable; ManifestAgreement; SlotAgreement; OldRouteBlocked;
  ApprovalRejected; SpentRejected; StopRejected; HashRejected; ParentRejected;
  WriterRejected; EpochRejected; PinConflictsRejected; SourceWindow;
  GuardsUnchanged; CommandExtentRegression; HistoricalContext;
  RuntimeNoFallback; UnitLineage; ReceiptIdentity; BudgetIdentity;
  CreateNewIdentity; OriginalSuiteUnchanged.
- [ ] BindingAvailable parses actual function definitions and asserts the new selector exists; expected RED exact assertion message “Extent contract regression: missing Get-ODEpochManifestName”. Missing input/parse errors are setup failures, not expected RED.
- [ ] Tests extract allowlisted function AST definitions only into isolated scope:
  Get-ODAttemptSlot, Assert-ODExtentContractFacts, Get-ODEpochManifestName,
  Assert-ODExtentLineage, Get-ODPinnedSHA256, Get-ODSourceOwner,
  Get-ODFailureExtent, Read-ODSourceModel, Get-ODIntentCallBinding,
  New-ODFixedContext, New-ODRuntimeContext, New-ODHistoricalContext,
  New-ODIntentUnitContext, Driver-AddPin. Preserve source line metadata where
  existing source-error checks depend on it. Never dot-source driver top level.
- [ ] Use actual selectors with fixed epoch/grant values; actual fact/lineage
  predicates with one-field negative mutations. Wrong hashes/parents/writer
  must reject, not select another epoch. Spent group covers each intent/final/
  stream/evidence file plus capture/unit/snapshot directory.
- [ ] Before a real new freeze exists, test proposed runtime/unit wiring by
  AST/call-argument assertions and lineage predicates over explicitly labelled
  in-memory candidate data. Do not create a fake frozen file or stub hash checks.
  Actual runtime/unit constructors on the new manifest are exercised by the
  single host later, not falsely claimed proven by auxiliary tests.
- [ ] CommandExtentRegression copies the ten existing invariant bodies into this
  versioned test with new input/result bindings; assert AST-equivalent bodies
  after that binding-only normalization. Preserve ten subgroups SavedErrorAccepted,
  CommandExtent,WrongColumnRejected,WrongLineRejected,WrongOperationRejected,
  WrongFunctionRejected,WrongSourceRejected,MethodsUnchanged,ThrowsUnchanged,
  ModelBoundaryUnchanged. Read actual historical row8 and original source snapshot;
  use unchanged OCSourceError predicate extraction as in the previous test.
  Copy only CE-Need,CE-Error,CE-Equal,CE-Copy,CE-Case,CE-Reject,CE-Tree,
  Get-CEContext,CE-FaultNodes,CE-Owner,CE-ExtentMatches and OC-Need,OC-Keys,
  OC-Integer,OC-Expand,OC-SourceError definitions unchanged. No old phase runner
  invocation or16/17 intent reuse.
- [ ] Static checks compare27unchanged source bytes; allow only declared binding
  edits in four files. Compare suite case bodies/order and owned-process tail
  after explicit epoch argument substitutions; pure predicates/oracles/recorder
  remain byte-identical. Record allowed spans and original hashes in input.
- [ ] Pin group assertions: ManifestAgreement checks the two exact filenames;
  SlotAgreement checks driver/capture/suite/context argument flow and new journal
  paths; OldRouteBlocked checks old gate remains on every legacy branch with
  historical stop present. Approval/Spent/Stop/Hash negative groups independently
  falsify the corresponding actual gate facts; Parent/Writer negative groups
  mutate each of the lineage hashes. EpochRejected tests empty,unknown,case variant
  and extent-contract with writer-green/core/lifecycle/continuation. PinConflicts
  requires same path/different hash rejection and same path/same hash single row.
  HistoricalContext builds actual historical context without the new selector;
  RuntimeNoFallback asserts no catch-driven selection and bad selector rejection.
  UnitLineage asserts new parent->writer chain acceptance and reversed chain
  rejection. ReceiptIdentity checks actual RED18 identity/candidate link; GREEN19
  receipt validation itself is independently enforced by driver after the run.
  BudgetIdentity asserts exact new1/0/0/0 ceilings and cumulative counters;
  CreateNewIdentity checks all new artifact writers use CreateNew and reject
  occupied paths without writing during the test; OriginalSuiteUnchanged compares
  all49case bodies/order plus3recorder groups to baseline. SourceWindow and
  GuardsUnchanged enforce byte/allowed-span comparisons described above.
- [ ] Implement Q runner -Phase Red|Green with exclusive intent, exact identity
  checks, full tool-call recording, expected receipt checks and first-error stop.
  Prepare script only hashes/parses/copies metadata, no test calls.
- [ ] Preserve preparation archive; decode-check bytes/lengths/SHA/directory and
  empty-file inventory. Commit explicit test/archive/docs and push verified branch.
- [ ] Invoke Q/run-extent-contract.ps1 -Phase Red once in existing tool context.
  Require0PASS/1FAIL/21NotRun,exit1, exact BindingAvailable failure, null setup/
  recording errors and unchanged31sources/history. Any other outcome stops.

## Task 2: Coordinate bindings and obtain GREEN19

**Files:** Four operational files, candidate input/archive; no other source edits.
**Consumes:** Exact RED18 and immutable test/input.
**Produces:** Four-file candidate, exact GREEN19, freeze-ready provenance.

- [ ] Implement interfaces above. Route only explicit new Prepare/RunOne to
  Driver-ExtentContractGate; old routes still call old gate and remain blocked.
  Catch writes new stop only for the explicit new route; old stop behavior remains.
- [ ] Implement fixed epoch manifest selection and parent/writer chain checks.
  Preserve legacy direct-writer check; new unit context anchors repair parent
  before using writer fixture. Preserve historical model and command extent fix.
- [ ] New gate checks actual approval, all historical hashes/old snapshots,
  candidate hashes, exact RED/GREEN identities and no spent/new-stop state.
  Prepare permits no new manifest/intent/snapshot tree; RunOne requires valid
  completed Prepare intent/manifest but no host files/capture/unit directory.
- [ ] Ensure semantic oracle lookup remains contract-green while all physical
  output/journal/unit paths and error-path expectations use extent-contract-green.
  Capture loads source-model selector through existing load order before suite use.
- [ ] Freeze P/extent-contract-candidate.input.json with31candidate hashes,
  InputSHA256, RedFinalSHA256, exact append paths/count and unchanged-history hashes.
  Freeze identities do not contain GREEN hash yet (avoid circular pin dependencies).
- [ ] Preserve/push candidate archive and source checkpoint, verify remote equality.
- [ ] Invoke Q/run-extent-contract.ps1 -Phase Green once. Require22PASS/0FAIL/
  0NotRun, ten command subgroups PASS,exit0,null errors and exact pre/post hashes.
  Unexpected failure ends scope without patch/retry or use of remaining host.

## Exact new provenance pins

Append exactly the following30unique paths to inherited2276pins. Replace only
four operational pin hashes (no additional operational entries). Expected
PinCount2306, SourceCount31. Before RED, compare this list against parent paths;
any existing overlap or unexpected count requires plan correction, not silent
deduplication. Driver-AddPin still rejects conflicting duplicates.

1. docs/superpowers/specs/2026-11-03-offline-extent-contract-rebind-design.md
2. docs/superpowers/plans/2026-11-03-offline-extent-contract-rebind.md
3. L/extent-contract.tests.ps1
4. L/command-extent.tests.ps1
5. P/fault-location.input.json
6. P/fault-location.final.json
7. P/fault-location.intent.json
8. P/fault-location.tool-record.json
9. P/command-extent.approval.json
10. P/command-extent.input.json
11. P/command-extent-candidate.input.json
12. P/command-extent-red.intent.json
13. P/command-extent-red.final.json
14. P/command-extent-red.tool-record.json
15. P/command-extent-green.intent.json
16. P/command-extent-green.final.json
17. P/command-extent-green.tool-record.json
18. P/extent-contract.approval.json
19. P/extent-contract.input.json
20. P/extent-contract-candidate.input.json
21. P/extent-contract-red.intent.json
22. P/extent-contract-red.final.json
23. P/extent-contract-red.tool-record.json
24. P/extent-contract-green.intent.json
25. P/extent-contract-green.final.json
26. P/extent-contract-green.tool-record.json
27. Q/prepare-extent-contract.ps1
28. Q/run-extent-contract.ps1
29. Q/preserve-extent-contract.ps1
30. P/pin-binding.stop.json

Approval HistoricalFiles additionally anchors all nine stops, failed oldcontract
final/evidence/streams, parent manifest and acceptedwriter manifest/final/evidence/
streams; input and approval pins transitively bind these exact bytes. Do not add
mutable progress documents, archives, new manifest or its future host receipt to
its own pin set. No circular self-hash. ExtentContractRecovery stores exact hashes
of Approval/Input/Candidate/RedFinal/GreenFinal, parent and all acceptedwriter
artifacts; preserve existing recovery maps byte-equivalently.

## Task 3: Freeze once, run once, preserve actual outcome

**Files:** Driver's new Prepare route; new P manifest/snapshots/host records;
portable freeze/final archives and result/status/recovery documents.
**Consumes:** Exact GREEN19, committed candidate, original52group oracle.
**Produces:** One frozen epoch and actual single-host result, never broader acceptance.

- [ ] From approved existing tool context invoke driver -Action Prepare
  -Grant extent-contract-rebind -Epoch extent-contract exactly once. CreateNew
  P/extent-contract.intent.json precedes writes. Copy31snapshots, verify every
  byte/hash and inherited pins; write manifest only when all inputs are valid.
  Schema remains existing schema with SourceCount31, PinCount2306,
  PreviousFrozenSHA256 fixed parent, MaxHosts1/MaxChildren0/
  MaxCompilations0/MaxReviewers0 and ExtentContractRecovery map.
- [ ] Reject partial freeze on any exception, preserve partial files and stop;
  never finish or overwrite it in a second attempt.
- [ ] Build freeze archive containing all P/Q originals plus the two historical
  H/runtime-green.{frozen,evidence}.json inputs, directory/empty-file inventory.
  Verify decoded copies; commit/push archive and docs; remote SHA must equal HEAD
  before host launch. Mechanical preservation cannot invoke tests or processes.
- [ ] Invoke driver -Action RunOne -Grant contract-green -Epoch extent-contract
  once. Use existing pinned host executable/assembly identities and privilege
  checks; retain private process ownership and bounded output handling.
  New intent marks host consumed before launch. No child targets.
- [ ] Read actual parent final and full tool record. Accept only original52ordered
  PASS groups, original exact streams/expected exit, actual EvidenceValid=true,
  Exact=true, postpins intact, zero children and no timeout/termination/overflow/
  rescue. Never infer acceptance from child exit alone or fill missing evidence.
- [ ] Always preserve actual output/error/partial records and new stop if any.
  Write report with consumed counters and separate unused/frozen allowances.
  Even success remains NOT_ACCEPTED for product; no core/lifecycle/continuation.
- [ ] Create final portable archive and decode-verify; update status/recovery with
  exact hashes/counts and next blocked step. Explicitly stage only source/test/
  archives/docs, inspect staged diff and raw staged hashes, commit and ordinary
  push to existing verified origin/branch. Verify remote SHA and clean worktree.

## Self-review and execution handoff

Scope maps to tasks1(inputs/tests),2(four bindings),3(freeze/host/preservation).
All five Review Focus cases have named tests. Runtime constructors are deliberately
not claimed dynamically verified before real freeze. Legacy failures remain failed.
No new allowances consumed by this plan. Approval of execution is still required;
then use native executing-plans under this finite budget, not an extra reviewer or
generic additional validation loop.
