# History Contract Recovery Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Freeze the repaired current source under a new identity and run one original52group contract host.
**Architecture:** Extend four existing binding consumers with history-contract; share real read-only input validation between regression and production gate. Separate immutable intent acquisition from validation so early failures remain recorded.
**Tech Stack:** Windows PowerShell7.6.5, existing pinned runtime, JSON/SHA256, native inline execution.
**Spec:** ../specs/2026-11-03-offline-history-contract-recovery-design.md

Status: PENDING_EXECUTION_APPROVAL. User approved design06676374 with “是”.
Original design bytes remain unchanged. No execution during planning.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; branch codex/c12-b01-task9-coordinator.
- L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2; P=.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner; Q=.superpowers/sdd/2026-11-01-offline-controller-recovery.
- TWO added auxiliaries RED22/GREEN23, ceiling23; ONE new Prepare. Explicitly reactivate only the unconsumed1host/31snapshots, not a second allowance.
- New epoch history-contract; Prepare grant history-contract-rebind; semantic Grant contract-green; host slot history-contract-host; auxiliary slots history-contract-binding-red/green.
- Manifest P/history-contract.frozen.json; Prepare intent history-contract.intent.json; snapshots/history-contract; stop history-contract.stop.json.
- Current21aux/4hosts/3compiles/186snapshots; maximum23/5/3/217. No child targets, compilations or reviewers. Legacy downstream3hosts/7targets/conditional1review stay frozen.
- All12old stops/failed records unchanged. Old routes blocked. No arbitrary path/epoch, fallback, install, security change, cleanup or main merge.
- Retain120000ms cap,2000ms exit/drain, existing evidence/journal bounds, five false capability flags and private owned-process cleanup.
- First unexpected execution/preparation/freeze result, drift or recording error stops; no patch/retry. Even full success ends this scope, product NOT_ACCEPTED.

## Review Focus

- Absolute external history is not repo-relative: RealHistory and ExternalIdentity.
- Old receipts describe old source epochs: PriorReceiptIdentity and CandidateIdentity.
- Attempt ownership cannot be replayed or supplied by a caller: IntentOwnership and ConsumedAttempt.
- Parent is original repair, not failed extent or writer: ParentLineage and WriterIdentity.
- An unused correct helper gives false confidence: ActualInputCall and ActualSlotFlow.

## File map and interfaces

Modify only L/driver.ps1,L/capture.ps1,L/observation-contract.tests.ps1,
L/source-model.ps1 among31operational files. Create L/history-contract.tests.ps1.
Create Q/prepare-history-contract.ps1,Q/run-history-contract.ps1,
Q/preserve-history-contract.ps1,Q/history-contract-progress.md.
Create P/history-contract.approval.json,input.json and candidate.input.json using
the same prefix; actual phase intent/final/tool records use binding-red/green.
Create L/portable-history-contract-{preparation,candidate,freeze,final}.json,
docs/roadmap/2026-11-03-offline-history-contract-recovery-result.md; update
current-status/repository-recovery. Ignored raw artifacts are embedded, not staged.

Production interfaces (no new operational file):
- Get-ODAttemptSlot(Grant,Epoch='legacy') gains only history-contract/contract-green
  -> history-contract-host. Other new-epoch Grants and unknown/case variants reject.
- Get-ODEpochManifestName(Epoch='legacy') gains only history-contract ->
  history-contract.frozen.json. Existing cases unchanged.
- Driver-ValidateHistoryInputs([string]Repo,[IDictionary]Approval,
  [IDictionary]Binding,[IDictionary]Candidate) -> ordered dictionary with
  Parent,WriterArtifacts,CandidateSources,PriorBinding,PathRepair.
  Read-only; calls unchanged Driver-VerifyHistory for real authenticated history.
  Verify approval identity/finite values, candidate hash inventory, actual parent/
  writer artifacts and historical receipts. No reliance on caller true/false flags.
- Driver-RequireHistoryReceipts([string]Repo,[IDictionary]Binding,
  [IDictionary]Candidate) -> RedSHA256,GreenSHA256. Require actual22/23 intents,
  ordered group results, test/input/candidate SHA, full tool records, pre/post
  identities and exact expected dispositions; no fake receipt for testing.
- Driver-HistoryRecoveryMap(Inputs,Receipts) -> HistoryContractRecovery with
  ParentSHA256,WriterSHA256,ApprovalSHA256,InputSHA256,CandidateSHA256,
  RedSHA256,GreenSHA256,PriorBinding,PathRepair,WriterArtifacts.
- Assert-ODHistoryLineage(Manifest,Parent,ParentSHA256,WriterSHA256) -> void.
  Require exact original parent and writer SHA below, both parent links and new
  HistoryContractRecovery identities. Leave Assert-ODExtentLineage unchanged.
- Driver-AcquireHistoryAttempt(Action,Grant) -> private ownership object with
  Path,Token,SHA256. Derive fixed path internally; CreateNew JSON Schema1 with
  Action,Grant,Epoch,AttemptConsumed=true,Token(random GUID). No passed path/token.
- Driver-HistoryGate(Ownership) -> authenticated Inputs/Receipts for new route.
  Validate actual owned intent bytes/token, all other attempt paths absent and
  no new stop. Existing file alone is never treated as owned. No public ownership
  parameter. In RunOne additionally validate completed Prepare/frozen map/snapshots.
- Driver-PrepareHistory(Info) -> existing style prepared JSON,31snapshots and new
  manifest. It does not create/rewrite already acquired intent.
- Invoke-ODHistoryContractRegression -Phase Red|Green -> ordered test receipt.
  Auxiliary runner alone creates new phase intent; test emits no side effects.

Approval schema1: BaseCommit,UserReply,SpecSHA256,PlanSHA256,AddedAuxiliary=2,
AuxiliaryCeiling=23,PrepareAttempts=1,ReactivatedHosts=1,ReactivatedSnapshots=31,
AddedChildren=0,AddedCompilations=0,AddedReviewers=0,ConsumedAuxiliary=21,
ConsumedHosts=4,ConsumedCompilations=3,ConsumedSnapshots=186,HistoricalFiles.
Binding schema1: ApprovalSHA256,SpecSHA256,PlanSHA256,TestSHA256,HelperSHA256,
BaselineSources(31),BaselineSourceBytes(four base64 entries, not extra snapshots),
OriginalHistory,HistoricalFiles,AppendPaths,ExpectedPinCount=2319,FixtureBindings.
Candidate schema1: InputSHA256,RedFinalSHA256,CandidateSources(31),AppendPaths,
ExpectedPinCount=2319. Immutable history referenced by InputSHA256, not recopied.

## Task 1: Freeze identities and run RED22

**Consumes:** current source, actual prior receipts and preserved snapshots.
**Produces:** complete immutable fixtures/test, exact RED22 without source edits.

- [ ] Record execution approval and verify branch/baseline. Read spec/plan and
  existing ledger before action; new namespaces must be absent. Parse every newly
  authored helper/test and check fixture-map keys before mechanical preparation.
  Authoring syntax checks are not regression invocations or acceptance claims.
- [ ] Fix parent repair-green SHA
  A9949DD46C7F928CFE7DB33F092AF0497E6F89556E8F359A8A55CB4484BB1556;
  writer SHA182B85886304BFE3CBFB248DDA70F2EB99DF3E2158F574DBE299519DE7CE0461.
  Bind current driver0F46DF4AA5CF442535426A17FA65672AB21A34514D6ACE42CFDF71226472E10D
  and other30source hashes; old operational pins are checked against old snapshots.
- [ ] Freeze original history from history-path-v3 input, all12stops, writer
  manifest/final/evidence/streams, failed extent Prepare, actual18/19/20/21
  receipts and full tool records. Separate live source hashes from historical
  snapshot keys; preserve historical driver bytes rather than compare edited live
  driver against old receipt. Every fixture must name its source map explicitly.
- [ ] Create definitions-only regression with these20ordered groups:
  NewEpochBinding,RealHistory,ExternalIdentity,RepositoryFixture,ApprovalIdentity,
  CandidateIdentity,PriorReceiptIdentity,ParentLineage,WriterIdentity,
  WrongHashRejected,InvalidPathRejected,UnknownEpochRejected,DisjointNamespaces,
  ConsumedAttempt,IntentOwnership,ActualInputCall,ActualSlotFlow,SourceWindow,
  OriginalGuardsUnchanged,OriginalSuiteUnchanged.
- [ ] RED group1 extracts actual Get-ODAttemptSlot and invokes contract-green with
  history-contract, asserting literal history-contract-host. Current function
  throws RuntimeException message/ErrorId 'OD: invalid explicit epoch'.
  Require0PASS/1FAIL/19NotRun/exit1, null setup/recording errors and unchanged input
  hashes. Missing input/function/parser errors are not expected RED.
- [ ] Test extraction allowlist: Get-ODAttemptSlot,Get-ODEpochManifestName,
  Driver-Hash,Driver-CheckHash,Driver-Need,Driver-EqualJSON,Driver-Keys,
  Driver-VerifyHistory,Driver-ValidateHistoryInputs,Assert-ODHistoryLineage,
  Driver-RequireFinal,Get-ODWriterManifestPath,Driver-FalseFlags.
  Extract AST definitions with preserved offsets and isolated repo/packet context;
  never dot-source full driver/capture or invoke receipt/attempt/host orchestration.
- [ ] GREEN groups2–4 invoke real Driver-ValidateHistoryInputs and real
  Driver-VerifyHistory on complete pinned history; additionally exercise the
  three external literals and L/expected-output.json using BaselineSources hash,
  not snapshot-keyed history. Require silent verifier success and exact returned
  parent/writer/source identity values. No hashing mocks.
- [ ] Groups5–12 use in-memory one-field mutations of actual approval/binding/
  candidate or lineage maps: wrong approval/schema/budget; wrong/missing candidate
  source hash; misbound prior receipt hash; parent swapped to writer; writer hash
  zeroed; real relative file wrong hash; unknown absolute or escaping relative
  path; empty/unknown/case-variant epoch and new epoch with non-contract Grant.
  Assert precise validation rejection, not generic setup/I/O errors except the
  deliberate real hash mismatch. Do not write fixtures or alter actual receipts.
- [ ] Groups13–15 enumerate fixed artifacts and check all namespaces disjoint.
  Verify actual filesystem decision expressions from gate/attempt AST against
  occupied existing historical files/directories and new absent paths without
  creating attempts. Inspect CreateNew use and ownership-token comparison on the
  actual new path. These are read-only/structural checks, not claims that attempt
  writing has executed; real acquisition is verified in Task3.
- [ ] Group16 invokes the actual extracted Driver-ValidateHistoryInputs call from
  the new gate using real fixture objects; compare actual return, not fact flags.
  Group17 asserts every physical path consumer carries explicit epoch and matches
  literal history-contract-host. Groups18–20 compare27unchanged files, unchanged
  Driver-VerifyHistory/Get-ODFailureExtent/owned-process/recorder functions, and
  original49suite+3recorder cases/order/assertions after binding-only normalization.
- [ ] Prepare/run/preserve helpers use CreateNew and record full actual calls,
  first error, source/history pre/post and finite counters. Freeze test/helper/
  input hashes before RED. Confirm43append paths unique and absent in parent.
- [ ] Archive preparation, decode-verify bytes/hashes/lengths/directories/empty
  files; commit/push explicit test/archive/docs before invoking RED22 once via
  Q/run-history-contract.ps1 -Phase Red with frozen input/test SHA arguments.

## Task 2: Coordinate new route and run GREEN23

**Consumes:** exact RED22 and frozen test/input.
**Produces:** four-file candidate and actual GREEN23, no freeze or host yet.

- [ ] Extend selector cases, entry validation and explicit epoch flow in four
  files. Suite change is binding plumbing only. New source-model lineage uses
  HistoryContractRecovery and original repair->writer chain; historical and
  extent-contract branches retain existing behavior.
- [ ] Implement shared read-only input validator and new map/receipt functions.
  Check old receipts against their own source hashes; current candidate against
  current31source hashes. Preserve original recovery maps, source-path order,
  bounded runtime/privilege checks and acceptedwriter final validation.
- [ ] Implement private new-attempt acquisition immediately after fixed root and
  writer definitions are available, BEFORE engine hash/resource/gate validation.
  Only exact new Action/Grant/Epoch dispatch enters it. Catch preserves new stop
  and acquired intent. Old routes retain old stop behavior. Track acquisition
  object in local scope; never infer ownership just from existing intent bytes.
- [ ] Prepare intent has no green/self/future hash requirements at creation.
  Later manifest recovery map binds authenticated approval/input/receipts; no
  circular pins. RunOne keeps acquired host intent; do not invoke old duplicate
  intent writer. Validate ownership separately from absence of all other host
  outputs. Existing Prepare intent is allowed for RunOne only with verified
  completed manifest/snapshots and original token identity recorded therein.
- [ ] Freeze candidate input; parse four files and inspect binding-only changes.
  Preserve/push candidate checkpoint and verify remote before GREEN23.
- [ ] Invoke Q/run-history-contract.ps1 -Phase Green once. Require20PASS/0FAIL/
  0NotRun/exit0, all errors null, unchanged source/history inputs. Unexpected
  result ends scope without fixes, retries or use of host allowance.
  Actual GREEN23 authentication occurs after receipt exists, never a mock during it.

## Exact provenance append set

Parent2276 +43unique new entries =2319pins, SourceCount31. Planning read-only
comparison found no inherited overlap. Recheck before RED; any disagreement stops
preparation rather than silently changing count. Replace four parent operational
hashes, preserve all other entries. Use consistent duplicate checking.

1. docs/superpowers/specs/2026-11-03-offline-history-contract-recovery-design.md
2. docs/superpowers/plans/2026-11-03-offline-history-contract-recovery.md
3. L/history-contract.tests.ps1
4. P/history-contract.approval.json
5. P/history-contract.input.json
6. P/history-contract-candidate.input.json
7. P/history-contract-binding-red.intent.json
8. P/history-contract-binding-red.final.json
9. P/history-contract-binding-red.tool-record.json
10. P/history-contract-binding-green.intent.json
11. P/history-contract-binding-green.final.json
12. P/history-contract-binding-green.tool-record.json
13. Q/prepare-history-contract.ps1
14. Q/run-history-contract.ps1
15. Q/preserve-history-contract.ps1
16. P/extent-contract.approval.json
17. P/extent-contract.input.json
18. P/extent-contract-candidate.input.json
19. P/extent-contract-binding-red.intent.json
20. P/extent-contract-binding-red.final.json
21. P/extent-contract-binding-red.tool-record.json
22. P/extent-contract-binding-green.intent.json
23. P/extent-contract-binding-green.final.json
24. P/extent-contract-binding-green.tool-record.json
25. P/extent-contract-prepare.tool-record.json
26. P/pin-binding.stop.json
27. P/extent-contract.stop.json
28. P/history-path.stop.json
29. P/history-path-v2.stop.json
30. P/history-path.input.json
31. P/history-path-red.intent.json
32. P/history-path-red.final.json
33. P/history-path-red.tool-record.json
34. P/history-path-v3.approval.json
35. P/history-path-v3.input.json
36. P/history-path-v3-candidate.input.json
37. P/history-path-v3-green.intent.json
38. P/history-path-v3-green.final.json
39. P/history-path-v3-green.tool-record.json
40. L/extent-contract.tests.ps1
41. L/history-path.tests.ps1
42. L/history-path-v2.tests.ps1
43. L/history-path-v3.tests.ps1

The inherited set already pins the other8stops. New immutable input binds
additional historical test/approval/result files transitively, including original
local command repair evidence. Do not append mutable progress/archives, the new
manifest, its own intent or future host result to its own pins.

## Task 3: One freeze, one host, preserve outcome

**Consumes:** exact GREEN23, current committed candidate, actual prior records.
**Produces:** new frozen source and actual host disposition, never product acceptance.

- [ ] Invoke driver -Action Prepare -Grant history-contract-rebind
  -Epoch history-contract -PreviousFrozenSHA256 with fixed parent SHA once.
  Early acquisition must exist even if validation later fails. Validate actual
  receipts and inputs, inherit2276pins, replace four source entries, append43,
  preserve maps and add HistoryContractRecovery. MaxHosts1/MaxChildren0/
  MaxCompilations0/MaxReviewers0; SourceCount31/PinCount2319.
- [ ] Copy exactly31snapshots using CreateNew, verify bytes/hash and full source/
  provenance before final manifest write. Record PrepareToken in manifest and
  map; do not pin intent/manifest to themselves. Partial error ends scope.
- [ ] Save full Prepare tool record, build freeze portable archive and verify it;
  commit/push and confirm remote SHA before host invocation.
- [ ] Invoke driver -Action RunOne -Grant contract-green -Epoch history-contract
  -FrozenSHA256 with actual new manifest SHA once. Early host intent acquisition
  precedes validation; existing pinned owned-process tail launches at most one
  host. Preserve semantic original oracle. No child targets.
- [ ] Read actual parent final. Require52ordered PASS, exact original streams/exit,
  Exact/EvidenceValid true, postpins intact, zero children, no timeout/termination/
  overflow/rescue. Raw child exit alone never passes. Failed invocation is
  consumed even if process never started; record actual host starts separately.
- [ ] Always save full tool input/output/partial evidence/stop. Verify final archive
  bytes/length/hash/directories/zero-byte inventory. Update result/status/recovery
  with exact consumed counters and remaining frozen allowances.
- [ ] Explicitly stage necessary files, check diff and raw staged hashes, commit/
  ordinary push to existing verified origin/branch; confirm remote equality.
  Stop after preservation even if all52pass. No downstream runtime or main merge.

## Self-review and handoff

All design sections map to Task1(identity/tests),Task2(four bindings/ownership),
Task3(real freeze/host/preservation). Review Focus cases have named tests.
Attempt ownership is only statically/read-only checked before real execution;
post-freeze/receipt acceptance is not claimed exercised before artifacts exist.
Actual shared validation replaces fact-only coverage. Exact pin count was checked.
Native execution, no additional reviewer, under proposed finite scope.
Await execution approval; this planning turn consumes no allowance.
