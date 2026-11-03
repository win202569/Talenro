# GREEN24-backed Runtime Recovery Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Authenticate historical GREEN24 and a newly tested candidate, then freeze and run one original52group host.
**Architecture:** One enumerated epoch in the existing four binding consumers. Separate historical receipt authentication from live candidate verification; retain private early intent acquisition and the original process owner.
**Tech Stack:** Windows PowerShell, pinned runtime, JSON/SHA256, native inline execution.
**Spec:** ../specs/2026-11-03-offline-green24-runtime-recovery-design.md

Status:PENDING_EXECUTION_APPROVAL. User approved design067f05a7 with “是”.
Original design bytes remain unchanged; this plan does not authorize its own execution.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; branch codex/c12-b01-task9-coordinator.
- L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2.
- P=.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner.
- Q=.superpowers/sdd/2026-11-01-offline-controller-recovery.
- Current24aux/4hosts/3compiles/186snapshots; additional RED25/GREEN26 only, ceiling26.
- One new Prepare; reactivate only unused1host/31snapshots. Maximum26/5/3/217.
- Zero children/compilations/reviewers. Legacy downstream3hosts/7targets/conditional1review frozen.
- Epoch history-contract-runtime; Prepare grant history-contract-runtime-rebind.
- Semantic Grant contract-green; physical slot history-contract-runtime-host.
- Prepare intent/frozen/stop prefix history-contract-runtime; snapshots/history-contract-runtime.
- Auxiliary slots history-contract-runtime-binding-red/green, disjoint from host.
- All13stops and prior artifacts immutable; no22/23/24 replay, install, main merge or cleanup.
- Preserve120000ms host cap,2000ms exit/drain, evidence bounds, five false capability flags and private owned-process cleanup.
- First unexpected preparation/execution/drift/recording result stops without patch/retry.
- Even success ends after one host and preservation; product NOT_ACCEPTED.

## Review Focus

- GREEN24 source epoch must not be relabelled as the edited live candidate.
- Nested AST arrays must not turn a single-call assertion into collection comparison.
- Historical absolute external paths must not be joined to the repository.
- Existing attempt tokens must not be adopted after a failed invocation.
- Parent must remain actual repair-green, not a nonexistent intervening freeze.

## Files and interfaces

Only four of31operational files may change:L/driver.ps1,capture.ps1,
source-model.ps1,observation-contract.tests.ps1. Preserve other27byte identities.
Add L/history-contract-runtime.tests.ps1 and Q/prepare-history-contract-runtime.ps1,
run-history-contract-runtime.ps1,preserve-history-contract-runtime.ps1,
history-contract-runtime-progress.md. Create the P immutable inputs/actual receipts
listed below; portable archives use L/portable-history-contract-runtime-
{preparation,candidate,freeze,final}.json. Update result/status/recovery docs.

- Extend Get-ODAttemptSlot(Grant,Epoch) and Get-ODEpochManifestName(Epoch) with
  exact new cases; reject unknown/case variants and new epoch/non-contract grants.
- Driver-ValidateRuntimeInputs([string]Repo,[IDictionary]Approval,
  [IDictionary]Binding,[IDictionary]Candidate) -> ordered Parent,WriterArtifacts,
  HistoricalGreen24,CandidateSources. Read-only; real history verifier unchanged.
- Driver-RequireRuntimeReceipts(Repo,Binding,Candidate) -> RedSHA256,GreenSHA256:
  actual25/26 intents, ordered results, pre/post maps and complete tool records.
- Driver-RuntimeRecoveryMap(Inputs,Receipts) -> ParentSHA256,WriterSHA256,
  ApprovalSHA256,InputSHA256,CandidateSHA256,HistoricalGreen24,RedSHA256,
  GreenSHA256,WriterArtifacts. Add PrepareToken after successful owned acquisition.
- Driver-AcquireRuntimeAttempt(Action,Grant) -> private Path,Token,SHA256 using
  fixed destinations and CreateNew; no caller path/token parameter.
- Driver-RuntimeGate(Ownership) -> Approval,Input,Candidate,Validated,Receipts,
  Ownership. Only invocation-local ownership accepted.
- Driver-PrepareRuntime(Info) writes31snapshots and final manifest; no second intent.
- Assert-ODRuntimeLineage(Manifest,Parent,ParentSHA256,WriterSHA256) -> void;
  validates original repair->writer and new RuntimeRecovery map.
- Invoke-ODRuntimeBindingRegression(Phase Red|Green) -> ordered receipt with
 20groups, source/history pre/post hashes, RootError/SetupError/RecordingError.

Approval schema1 binds approved commits/spec/plan, exact finite allowances and
consumed counters. Binding schema1 contains ApprovalSHA256,TestSHA256,HelperSHA256,
BaselineSources31,BaselineSourceBytes(four base64 entries),HistoricalFiles,
Green24Artifacts,AppendPaths68,ExpectedPinCount2344. Candidate schema1 contains
InputSHA256,RedFinalSHA256,CandidateSources31,AppendPaths,ExpectedPinCount.
No mutable progress, self-intent, manifest or future host result in its own pins.

## Task 1: Freeze complete fixtures and run RED25

- [ ] Verify branch, clean baseline, new namespaces absent and31sources equal
  GREEN24 amendment. Preserve actual13stops and22/23/24 records.
- [ ] Author all20groups before freezing:
  NewEpochBinding,Green24Identity,HistoricalEpoch,LiveCandidate,ApprovalIdentity,
  PriorFailurePreserved,ExternalHistory,RelativeHistory,ParentLineage,WriterIdentity,
  InvalidReceiptRejected,InvalidCandidateRejected,EpochRejected,DisjointSlots,
  ScalarCallSelection,AttemptOwnership,ActualGateCall,SourceWindow,
  OriginalGuardsUnchanged,OriginalSuiteUnchanged.
- [ ] First group extracts actual Get-ODAttemptSlot; assert
  `(Get-ODAttemptSlot contract-green history-contract-runtime) -ceq 'history-contract-runtime-host'`.
  Expected RED25:0PASS/1FAIL/19NotRun/exit1; RuntimeException message/ErrorId
  'OD: invalid explicit epoch'; setup/recording null, immutable hashes unchanged.
- [ ] Green24Identity authenticates real amendment SHA
  FA9A9A82DFC135A2284B62C341F1D1B16B47ABFC3ADCDDC42DF39D765F1A99E2
  and receipt AB86CCC7042089B0D705E9FB5933B4ADD92045514B5B976068315693F7CCFC8A,
  actual intent24/ceiling24,20orderedPASS/exit0/errorsnull, all2391preservation
  checks and full initial/poll/final tool records. HistoricalEpoch compares saved
  receipt sources to saved candidate/baseline bytes, never edited live files.
- [ ] LiveCandidate validates all31current hashes. ApprovalIdentity checks exact
  allowance/schema and rejects one-field mutations. PriorFailurePreserved requires
  original failedGREEN23 and all stops without interpreting them as successes.
- [ ] ExternalHistory uses actual three pinned external identities; RelativeHistory
  uses L/expected-output.json from BaselineSources. No mocked hashing.
- [ ] ParentLineage/WriterIdentity authenticate actual original files and reject
  swapped/zeroed identities. InvalidReceiptRejected mutates a copy's computation,
  source epoch, PASS count and tool identity one field at a time. InvalidCandidateRejected
  rejects missing/wrong source hashes. All rejections must be explicit validation
  errors, not setup/parser errors.
- [ ] EpochRejected tests empty/unknown/case variants and non-contract grant.
  DisjointSlots checks exact new paths against occupied old paths.
  ScalarCallSelection enumerates individual CommandAst nodes and requires one
  actual validator call and one acquisition call with integer offsets.
- [ ] AttemptOwnership checks real read-only occupancy decisions and structural
  CreateNew/token checks; no fake attempt writes. ActualGateCall executes only the
  extracted read-only validator call with real fixtures and checks actual returned
  parent/writer/receipt/source values. Do not invoke the full driver in auxiliaries.
- [ ] SourceWindow checks exactly four allowed changes; OriginalGuardsUnchanged
  preserves history verifier, extent logic, recorder and process tail;
  OriginalSuiteUnchanged compares49+3groups/order/assertions after binding-only
  normalization. These tests must not manufacture new success receipts/manifests.
- [ ] Parse all authored scripts and check fixture keys before preparation writes.
  Recheck68unique append paths and no parent overlap. Freeze approval/input/test.
- [ ] CreateNew preparation archive; decode-verify all files/directories/zero bytes.
  Commit explicit files, push current branch and verify remote equality.
- [ ] Run Q/run-history-contract-runtime.ps1 -Phase Red -InputSHA256 <fixed>
  -TestSHA256 <fixed> once; preserve full actual call/result. Proceed only on
  the exact RED25 disposition above.

## Task 2: Bind the new route and run GREEN26

- [ ] Extend exact selectors/epoch propagation in the four files only.
- [ ] Implement shared read-only input validation using real immutable historical
  records and current candidate maps. Preserve old validator and stopped-route
  behavior; do not call its live old-candidate check against new source.
- [ ] Authenticate GREEN24 in a separately named historical map; new25/26 receipts
  independently bind edited source. Preserve baseline bytes and hashes so prior
  receipts remain meaningful. No accepted-fact flags or copied success receipts.
- [ ] Implement early acquisition before engine/resource checking, locally held
  ownership, refusal of other preexisting outputs/new stop and no duplicate
  host intent writer. Prepare token is later bound by completed manifest.
- [ ] Implement RuntimeRecovery lineage and original2276pin inheritance; preserve
  old maps, replace only four operational hashes, append exactly68listed entries.
- [ ] Parse candidate, inspect binding-only changes, fix candidate input with31hashes.
  Archive/checkpoint/push and verify remote before GREEN26.
- [ ] Run Q/run-history-contract-runtime.ps1 -Phase Green -InputSHA256 <fixed>
  -TestSHA256 <fixed> once. Expected20PASS/0FAIL/0NotRun/exit0; all errors null;
  source/history pre/post equal. Preserve actual call. Any other result stops.

## Exact appended provenance

Read-only planning check:68unique paths, zero inherited overlap;2276+68=2344.
Recheck before preparation; disagreement stops rather than changing the count.
The first53entries exist as historical evidence; remaining15are this new scope.

1. docs/superpowers/specs/2026-11-03-offline-history-contract-recovery-design.md
2. docs/superpowers/plans/2026-11-03-offline-history-contract-recovery.md
3. docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/history-contract.tests.ps1
4. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract.approval.json
5. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract.input.json
6. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-candidate.input.json
7. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-binding-red.intent.json
8. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-binding-red.final.json
9. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-binding-red.tool-record.json
10. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-binding-green.intent.json
11. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-binding-green.final.json
12. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-binding-green.tool-record.json
13. .superpowers/sdd/2026-11-01-offline-controller-recovery/prepare-history-contract.ps1
14. .superpowers/sdd/2026-11-01-offline-controller-recovery/run-history-contract.ps1
15. .superpowers/sdd/2026-11-01-offline-controller-recovery/preserve-history-contract.ps1
16. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/extent-contract.approval.json
17. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/extent-contract.input.json
18. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/extent-contract-candidate.input.json
19. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/extent-contract-binding-red.intent.json
20. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/extent-contract-binding-red.final.json
21. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/extent-contract-binding-red.tool-record.json
22. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/extent-contract-binding-green.intent.json
23. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/extent-contract-binding-green.final.json
24. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/extent-contract-binding-green.tool-record.json
25. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/extent-contract-prepare.tool-record.json
26. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/pin-binding.stop.json
27. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/extent-contract.stop.json
28. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-path.stop.json
29. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-path-v2.stop.json
30. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-path.input.json
31. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-path-red.intent.json
32. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-path-red.final.json
33. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-path-red.tool-record.json
34. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-path-v3.approval.json
35. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-path-v3.input.json
36. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-path-v3-candidate.input.json
37. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-path-v3-green.intent.json
38. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-path-v3-green.final.json
39. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-path-v3-green.tool-record.json
40. docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/extent-contract.tests.ps1
41. docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/history-path.tests.ps1
42. docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/history-path-v2.tests.ps1
43. docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/history-path-v3.tests.ps1
44. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract.stop.json
45. docs/roadmap/2026-11-03-offline-history-contract-v2-approval.md
46. docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/history-contract-v2.tests.ps1
47. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-v2.input.json
48. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-v2-green.intent.json
49. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-v2-green.final.json
50. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-v2-green.tool-record.json
51. .superpowers/sdd/2026-11-01-offline-controller-recovery/prepare-history-contract-v2.ps1
52. .superpowers/sdd/2026-11-01-offline-controller-recovery/run-history-contract-v2.ps1
53. .superpowers/sdd/2026-11-01-offline-controller-recovery/preserve-history-contract-v2.ps1
54. docs/superpowers/specs/2026-11-03-offline-green24-runtime-recovery-design.md
55. docs/superpowers/plans/2026-11-03-offline-green24-runtime-recovery.md
56. docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/history-contract-runtime.tests.ps1
57. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-runtime.approval.json
58. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-runtime.input.json
59. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-runtime.candidate.input.json
60. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-runtime-binding-red.intent.json
61. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-runtime-binding-red.final.json
62. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-runtime-binding-red.tool-record.json
63. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-runtime-binding-green.intent.json
64. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-runtime-binding-green.final.json
65. .superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/history-contract-runtime-binding-green.tool-record.json
66. .superpowers/sdd/2026-11-01-offline-controller-recovery/prepare-history-contract-runtime.ps1
67. .superpowers/sdd/2026-11-01-offline-controller-recovery/run-history-contract-runtime.ps1
68. .superpowers/sdd/2026-11-01-offline-controller-recovery/preserve-history-contract-runtime.ps1

## Task 3: One freeze, one host, preserve

- [ ] Run driver.ps1 -Action Prepare -Grant history-contract-runtime-rebind
  -Epoch history-contract-runtime -PreviousFrozenSHA256
  A9949DD46C7F928CFE7DB33F092AF0497E6F89556E8F359A8A55CB4484BB1556 once.
  Require actualGREEN26, exclusive intent,2344pins/31sources/31snapshots;
  MaxHosts1/MaxChildren0/MaxCompilations0/MaxReviewers0. Writer parent remains
  182B85886304BFE3CBFB248DDA70F2EB99DF3E2158F574DBE299519DE7CE0461.
- [ ] Verify all pins and snapshot bytes before final manifest write. Record actual
  tool output including partial failure. No retry or completion of partial freeze.
- [ ] Archive successful freeze, decode-verify, commit/push and confirm remote.
- [ ] Run driver.ps1 -Action RunOne -Grant contract-green
  -Epoch history-contract-runtime -FrozenSHA256 <actual-new-hash> once.
  Gate validates owned intent plus completed Prepare token/hash/map/snapshots.
- [ ] Require parent Exact/EvidenceValid,52orderedPASS, exact original streams/exit,
  zero children, intact post-pins, no timeout/termination/overflow/rescue.
  Record attempts separately from actual host starts; exit0 alone is insufficient.
- [ ] Always preserve actual calls, receipts, partials and stop in final archive.
  Verify decoded bytes/hashes/lengths/directories/empty files. Update result/status/
  recovery and counters; explicitly stage necessary files, verify raw staged bytes,
  commit/push and check remote HEAD equality. Stop even on success.

## Self-review and handoff

Task1 covers immutable fixtures and the five Review Focus risks; Task2 binds real
history/current source and ownership; Task3 tests completed freeze and actual host,
which auxiliary structural checks cannot prove. No extra suite/reviewer/cleanup
is authorized. Scope is native inline; this plan awaits execution approval.
