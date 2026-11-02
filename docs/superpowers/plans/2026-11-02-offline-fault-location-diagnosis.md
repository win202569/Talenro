# Fault-Location Diagnosis Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans natively after execution approval. No agents or reviewer. Steps use checkbox syntax.

**Goal:** Locate the first mismatch in saved BaselineCore data without repairing or replaying the original contract host.

**Architecture:** A separate diagnostic reads pinned historical data and extracts an explicit helper allowlist. A structural walker follows the frozen predicate's order and delegates atomic comparisons to its unchanged body. New records preserve the finding separately from original failed acceptance.

**Tech Stack:** Existing Windows PowerShell, AST, JSON and SHA256; no installations.

**Spec:** [Fault-location recovery addendum](../specs/2026-11-02-offline-fault-location-diagnosis-design.md), SHA D3A5EF530B57B682E11DC030A4C6143FFC556510C1AFDD8BC823F55CE20731EE. User “是” after dac6f473 approves the written design. This plan and execution allowance are PENDING_APPROVAL; retain the approved spec bytes.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; branch codex/c12-b01-task9-coordinator; origin https://github.com/win202569/Talenro.git.
- L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2; P=.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner; Q=.superpowers/sdd/2026-11-01-offline-controller-recovery; H=.superpowers/sdd/2026-10-31-offline-diagnostic-runtime-migration. All are repository-relative.
- Request exactly one additional auxiliary computation15, ceiling15. Existing14auxiliary/4hosts/3compiles/186snapshots/0targets/0reviews remain consumed. Added hosts/targets/compiles/runtime snapshots/reviewers ZERO; remaining runtime chain stays frozen.
- No existing source/test/oracle/manifest/stop/archive edits. No full-script imports, module/profile/control/writer execution, child hosts, runtime retries, new source freeze, installation or security changes. No main merge.
- Only the new diagnostic and mechanical preservation files may be authored. Loading the diagnostic defines functions only. No trial invocation, RED/GREEN pair, extra probe or default test suite: the specifically approved single computation supersedes generic implementation test cycles.
- Hash/AST/file inspection and document checks are preparation, not permission to invoke extracted helpers before intent15. First attempted computation consumes the slot, even on setup failure.
- Original failed final remains authoritative, EvidenceValid=false/EvidenceSHA256=null. Product NOT_ACCEPTED regardless of diagnostic result.

## Review Focus

1. Wrong epoch: bind the actual repair manifest, frozen helper bytes and historical core input before computing.
2. Misidentifying a later row: preserve dictionary order, array order, structural checks and shared dynamic-value state; stop at first failure.
3. Helper extraction executes unrelated code: allowlist unique top-level function definitions only; fail closed on ambiguity or dependencies outside the allowlist.
4. Path changes masquerade as coordinate bugs: require original worktree path for this single attempt; never rewrite original positions or manifest paths.
5. Partial output is interpreted as success: separate diagnosis status, actual error, recording error and source-integrity result; never emit contract PASS.

## Task1 — freeze the bounded diagnostic and inputs

**Create:** L/fault-location.diagnostic.ps1; P/fault-location.approval.json;
P/fault-location.input.json; Q/fault-location-progress.md;
L/portable-fault-location-preparation.json.

**Interface:** Invoke-ODFaultLocationDiagnosis -Repo [string] -InputPath [string].
No caller-supplied code, helper names, profile or alternate manifest. Fixed profile
entry-migration-core. All output paths fixed below. Function definitions only at
file top level; use explicit call within the existing tool context.

- [ ] Confirm approved plan/design, clean documentation checkpoint, nine stops and all new output paths absent. Verify the four spec anchors verbatim. Record original worktree root; if it differs, stop before intent and seek a separately reviewed relocation design. No automatic path rewriting.
- [ ] Bind the exact input inventory below in an ordered repository-relative path→SHA256 map. Require every applicable manifest pin Present and duplicate pins identical. Verify current operational helper bytes equal corresponding P/snapshots/repair-green/L copies. Historical source files must match their repair-manifest pins; do not evaluate their bodies. Verify all31 current operational SourcePaths against the manifest before and after the computation; freeze this inventory too.
- [ ] Write approval ordered fields Schema,BaseCommit,UserReply,DesignSHA256,PlanSHA256,AddedAuxiliary,Computation,Ceiling,Method; values int1,actual approval checkpoint/reply,actual hashes,1,15,15,Native. This is a new approval, not a rewrite of prior ones.
- [ ] Author the helper extraction and traversal described below. Static-parse the diagnostic and review its AST/call sites only; do not invoke it or extracted helpers. Verify unique allowed definitions and byte-identical extracted function text. If an additional helper or behavior is needed, stop rather than expanding scope.
- [ ] Write input ordered fields Schema,ApprovalSHA256,DesignSHA256,PlanSHA256,DiagnosticSHA256,Repo,Profile,Files,HelperNames,OperationalSources. Values int1,actual hashes,original absolute root,entry-migration-core and frozen inventories. No output hashes or self-reference. Hash and commit diagnostic/input preparation before any execution. Preserve P/Q in a NEW preparation portable archive; include required H inputs by verified reference to their existing tracked recovery archive or explicit portable entries, never rely on a private cache.

### Exact input inventory and identities

- Approved spec and this plan; approval and diagnostic frozen during Task1.
- P/repair-green.frozen.json and P/pin-binding.stop.json: exact spec hashes.
- L/portable-pin-binding.json and P/contract-green.evidence.json: exact spec hashes.
- P/contract-green.final.json: 85F28B2BFE5EE3722B1A30E865889012F6DFA5A4019C011E2F0ADA72C23AAADB.
- P/pin-binding-contract.tool-record.json: 64CB9B9D5BDD3F23AF08507B308B3CE528FF27E8313F849A7A875CD0E932C5D0.
- Original stdout/stderr paths named by that final: resolve within P and bind their actual bytes to final hashes; do not guess filenames.
- H/runtime-green.evidence.json: 58B2D81047526EFB2BEE3FD8155DFAF493FC3FD6ECAAE32B0E7ED19C7DB66EA0.
- H/runtime-green.frozen.json: exact Present pin in the anchored repair manifest.
- L/source-model.ps1: 1C7897E486D43601D4D0E963A566737C97E33FC8C88BF7E4ECF83348A5A61418.
- L/observation-contract.ps1: B68A9D8C9674692C8B75264CDB687E92063911254F64F5D69124E03265E3CB12.
- L/observation-contract.tests.ps1 (read only, never invoke): B459F5122E26042782FB44EAE45C8F45D1CFD94892E7A7B74945518454389906.
- L/expected-observations.json: 7D681DE1110D43561B2B148B12BE1821460A4F27734953F23D490E5B363E52B0.
- L/expected-output.json: 0FA169E9C7D69EB4D53AFF6C20378ADE4DB43E6767880F2991F26A7B87C69BBF.
- Historical model inputs, each at its exact repair-manifest pin: scripts/diagnose-offline.ps1; scripts/private/offline-diagnostics/{session.psm1,sinks.ps1,owned-step.ps1,sequence.ps1,profiles.ps1,observations.ps1}; docs/diagnostics/isolated-fuzz/p2b-entry-runtime-migration/{test-helpers.ps1,core.tests.ps1,owned.tests.ps1,live.tests.ps1,prefix.tests.ps1,entry.tests.ps1,capture.ps1}. Braces mean the enumerated filenames, not wildcard discovery.
- Existing nine stop paths from the previous recovery record, bound individually during preparation, plus the31 operational sources and two helper snapshots described above. The materialized input map must list each path/hash explicitly before execution.

### Helper allowlist and traversal

From source-model.ps1 only: Get-ODPinnedSHA256, Get-ODSourceOwner,
Get-ODFailureExtent, Read-ODSourceModel, New-ODFixedContext,
New-ODHistoricalContext. These are readonly construction helpers, not all pure;
their filesystem reads are confined to the frozen input inventory.
From observation-contract.ps1 only: OC-Need, OC-Keys, OC-Integer, OC-Expand,
OC-SourceError, OC-Match. No runtime/unit context, full suite, driver, report
predicate or profile body. Evaluate only inspected function-definition text in a
local scope, never whole source files or arbitrary callbacks from input data.

Read the historical evidence's unique ProfileReturn with matching Profile, take
its Observations, then perform the original JSON roundtrip at Depth60. Construct
New-ODHistoricalContext once from the anchored repair manifest. Require exactly
22actual/expected rows and the original supported-profile conditions. Do not
construct owned/live contexts or mutate any fixture.

Implement Trace-ODSavedValue($Actual,$Expected,$Context,$State,[string]$Path):
for a non-marker expected dictionary call unchanged OC-Keys first, then recurse
in Expected.Keys order; for an expected array check actual array/type/count
with unchanged OC-Need, then recurse by ascending index. For null, scalar,
'$error' or '$dynamic' nodes, invoke unchanged OC-Match once on that node using
one shared state (Pids and Elapsed ordered dictionaries). Marker dispatch order
must match OC-Match ($error before $dynamic). Do not invoke OC-Match again on
containers or run the whole predicate separately. This is one traced traversal,
not a contract suite replay or multiple diagnostic attempts.

Record structural/atomic visits (Path,Kind,Matched,Error) in order. First error
ends traversal. Preserve actual exception Type,Message,ErrorId,Position,Stack.
If it is an error-object leaf, derive diagnostic details from that SAME object:
expected source/function/operation, original Position/Stack, parsed coordinates,
model candidates for function+operation and exact-coordinate count, raw matching
AST node/ancestor extents. Details are data inspection within this computation;
do not rerun the predicate, change values or execute source commands. Report a
location mismatch only if actual error stack/predicate site and candidates
establish it; earlier structural or other rejection means HypothesisContradicted.
Missing/ambiguous detail means Inconclusive, never invent a row or coordinate.

## Task2 — one attempt, then preserve and stop

**Create:** P/fault-location.intent.json, P/fault-location.final.json,
P/fault-location.tool-record.json; P/fault-location.stop.json on failure;
L/portable-fault-location.json; docs/roadmap/2026-11-02-offline-fault-location-diagnosis.md.
**Update:** Q/fault-location-progress.md, current-status.md, repository-recovery.md.

- [ ] Require the preparation checkpoint pushed and input/script hashes unchanged. CreateNew intent BEFORE invocation, fields Schema,AttemptConsumed,Computation,Ceiling,InputSHA256,DiagnosticSHA256; values1,true,15,15,actual hashes. Existing intent or output means no invocation. Never reuse or remove it.
- [ ] In the existing PowerShell tool context, load only the new definitions-only diagnostic and invoke once:

```powershell
. ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/fault-location.diagnostic.ps1
Invoke-ODFaultLocationDiagnosis -Repo 'D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator' -InputPath './.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner/fault-location.input.json'
```

- [ ] CreateNew final ordered fields Schema,Computation,Status,InputSHA256,DiagnosticSHA256,PreHashes,PostHashes,Visits,FirstMismatch,DiagnosticError,RecordingError,OriginalAcceptance,ExitCode. Schema1/computation15; status LocatedMismatch, HypothesisContradicted or Inconclusive; OriginalAcceptance NOT_ACCEPTED. FirstMismatch includes Path,zero-based RowIndex,ActualError,ExpectedError,Source,Function,Operation,Line,Column,Candidates,ASTExtents,PredicateSite; unavailable fields null. DiagnosticError is infrastructure error, distinct from the observed comparison error. Exit0 means a recorded located mismatch only, not contract PASS; contradiction/inconclusive exit1, setup/recording error exit2. No all-matched acceptance: if traversal finishes without reproducing a rejection, report Inconclusive.
- [ ] On unexpected result, infrastructure error or drift, CreateNew diagnostic stop with actual error/partial status and RemainderFROZEN. On a located mismatch also stop work normally: no fix or second computation. Preserve complete actual tool input/results, not a reconstructed transcript; if final creation fails, retain outer tool error and partial records without retrying execution.
- [ ] Preserve all new P/Q artifacts in a new final portable archive; verify decoded bytes/lengths/hashes and directory/zero-byte entries. Never overwrite older archives. State actual count15consumed if intent/invocation occurred, and that remaining runtime chain is still frozen. Report the first mismatch and limitations without claiming all causes found.
- [ ] Review explicit necessary-file list, diff checks and staged byte equality; commit/push existing branch, verify remoteSHA==HEAD and clean worktree. No main merge. Handoff asks for a separately scoped repair only if diagnosis supports one.

## Self-review / approval boundary

Task1 covers identity, allowlist, provenance and ordered traversal; Task2 covers
single-use intent, actual outputs, finite stop and fresh-computer preservation.
The five review-focus risks have explicit checks within the single diagnostic
and static preparation, not extra computations. Generic RED/GREEN or fresh review
requirements would exceed the selected diagnostic-only scope and are not used.
No diagnostic code or execution occurred in this planning turn. Approve this plan
and its single added auxiliary allowance before Task1 implementation. Native only.
