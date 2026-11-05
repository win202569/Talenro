# Parent audit persistence validation implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline. Steps use checkbox syntax. This document proposes tests only; await user approval before implementation.

**Goal:** Exercise the existing parent post-supervision audit and persistence failure paths without starting business processes.

**Architecture:** Extract exact definitions and the post-supervision tail from the pinned diagnostic driver into isolated in-process runspaces. Supply tiny synthetic input files and terminal supervision data; use real hashing and exclusive filesystem writes, not mocked Save-IW/Map-IW. Do not invoke the original driver entry or modify its decisions.

**Tech Stack:** Existing PowerShell/.NET, AST, SHA256, local filesystem; no dependencies or compilation.

**Spec:** `docs/superpowers/specs/2026-11-05-runtime-staged-integrity-design.md`, specifically separate transport/audit/verdict and persistence refusal. Readiness evidence: `docs/roadmap/2026-11-05-staged-integrity-integration-gap-review.md` latest v4 section.

## Global constraints

- Preserve all consumed scripts, claims, archives and verdicts. Use only new `.superpowers/sdd/2026-11-05-parent-audit-persistence-v1` descendants; refuse occupied cases, traversal and reparse paths.
- Zero business, OS children, Go/Docker, runtime launch, full suite, installs, compilation, snapshot, ACL/security changes or main merge. Counters44/34/6/217 unchanged.
- One RED and one GREEN bootstrap invocation, then one ten-case parent matrix; stop first unexpected result, no automatic repair/retry. Expected refusals count as matching cases. At most ten parent-tail runspaces plus two harmless bootstrap runspaces.
- No full observation44 rescan: hash only tiny synthetic files. Original source identity checks are allowed. No v4 evidence matrix rerun or final verifier invocation.
- Preserve120000/5000/2000/2000ms runtime budgets without invoking them. Synchronous parent filesystem calls still lack a hard wall-clock bound; no hung-disk test or forced process termination.
- All five capability flags remain false; NOT_ACCEPTED even if every test passes. This proves selected existing parent-tail behavior, not live supervisor integration or production disk-fault resilience.

## Source and integration findings

- Driver SHA256:5031ED56375B5DB1194567E6D7D8B95405DD36595324D9BC6B4AED7A89BFECDC. Its Save-IW uses CreateNew,262144-byte ordinary record and2097152-byte map bounds.
- Driver lines67–83 preserve supervision, append journal, gate terminal state, perform Map-IW, persist audit, gate audit success, persist result and check transport success. Outer catch lines85–89 saves execution.stop when absent then exits1.
- `runtime-copy-transport-common.ps1` Get-HTHash/Assert-HTInputs perform actual hashes and absent-pin checks. Extract these definitions only; never import operational functions.
- `scripts/run-c12-integration.ps1` Focused/Suite entry has closed base/authority-v7/authority-v7-pitr profiles,3/3/8-minute setup allowances, prepared executable receipts, owned resource lifecycle and group/suite deadlines. Invoke-C12Group starts at5970. No direct diagnostic-driver wiring is proposed.

## File map and interfaces

L = `docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2`.
Create L/`runtime-parent-audit-fixtures.ps1`, L/`runtime-parent-audit-fixtures.tests.ps1`, L/`runtime-parent-audit-matrix.ps1`, L/`portable-runtime-parent-audit-persistence-v1.json`.
No modifications to existing driver, common helpers, review-v4 or formal entry.

- `Get-ParentAuditSlice([string]$DriverSource,[string]$CommonSource)` returns exact helper source plus a wrapped try/catch tail. Require pinned driver hash; pin common hash before tests. Select top-level Need-IW, Save-IW, Map-IW, Journal-IW and common Get-HTHash/Assert-HTInputs by AST. Tail starts at unique `Save-IW supervision.json $retained` statement and ends at existing success marker; use original outer catch unchanged. Reject missing/duplicate anchors or unexpected commands. Compare every extracted extent byte-for-byte with source; never replace decision code.
- `Invoke-ParentAuditCase([string]$CaseId,[string]$Workspace,[string]$Slice)` returns Terminal(bool), Output(string[]), Errors(string[]), and persisted-file hashes. Each runspace sets both PowerShell location and .NET current directory to its case root. Restore directory and dispose only after terminal state; retain/stop if uncertain. Reject unknown/occupied case IDs.
- Synthetic context binds HTRepo/base/folder to case root, relative='.', part='', InputSHA256='A'*64, options.IntentSHA256='B'*64. Typed terminal result has NextAllowed=true, PendingResources/ExitUnconfirmed=false, SupervisionOutcome='SUCCESS', Writer/Producer OwnedExit=true/ExitCode0. Before reference is explicitly synthetic. Binding uses one source file with fixed UTF8 bytes `fixture`, empty other maps and one absent relative path. These are not observation44 identities or genuine receipts.

## Review focus

1. Accidental runtime launch: bootstrap verifies extraction excludes supervisor/assembly/engine setup and runs harmless exit isolation only.
2. Writes escaping case root: bootstrap rejects traversal/reparse/occupied destinations; pin context and inspect all created files per case.
3. Mocked persistence hiding failure: P04–P07 preoccupy actual filesystem destinations with sentinel bytes, using unchanged CreateNew behavior.
4. Postprocessing despite pending exits: P08–P09 use a deliberately drifted source, require terminal refusal and absence of any after-map/audit/result.
5. Transport failure upgraded by successful audit: P10 requires terminal failure despite a VERIFIED audit and NOT_ACCEPTED result.

## Task1: Exact extraction and isolated fixture bootstrap

- [ ] Write tests first: missing helper RED; source-anchor/hash rejection; byte-identical extents; safe-root rejection; occupied case refusal. Two harmless runspaces (`exit 1`, then success) prove controller survival and terminal cleanup, not actual audit behavior.
- [ ] Run `& ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-parent-audit-fixtures.tests.ps1` once RED; retain receipt.
- [ ] Implement interfaces above; no substitution for Save-IW, Map-IW or Assert-HTInputs. Create each file exclusively; seed a tiny journal begin record outside the slice, retain it as test setup. Save fixture input/context and slice before invocation.
- [ ] Run same test command once GREEN, require `PARENT_AUDIT_BOOTSTRAP_VERIFIED`. Verify source hashes unchanged and commit helper/tests before matrix. Stop unexpected bootstrap result.

## Task2: One finite parent-tail matrix

Every case starts from an independent fresh synthetic baseline; no mutations to previous cases.

|Case|Single change|Required outcome|
|---|---|---|
|P01|None|Supervision saved; real after-map hash matches fixture; audit VERIFIED; result NOT_ACCEPTED/BusinessVerified false; success marker; no stop|
|P02|Source content drift after binding|Audit FAILED/After null with drift error; stop; no result or success marker|
|P03|Create absent-pinned file|Audit FAILED/After null with absent-pin error; stop; no result|
|P04|Preoccupy supervisor.after-map.json|Sentinel unchanged; audit FAILED/After null; stop; no result|
|P05|Preoccupy integrity-review.json|Sentinel unchanged; after-map may exist; stop; no result or success marker|
|P06|Preoccupy result.json|Sentinel unchanged; audit VERIFIED; stop; no success marker; sentinel never interpreted as result|
|P07|Preoccupy supervision.json|Sentinel unchanged; stop; no after-map/audit/result|
|P08|PendingResources=true; drift source as ordering probe|Terminal refusal; supervision saved; no after-map/audit/result; stop reason pending/unknown, not input drift|
|P09|ExitUnconfirmed=true; same drift probe|Same ordering guarantees as P08|
|P10|SupervisionOutcome='FAILED', other terminal fields unchanged|Audit VERIFIED allowed; result stays NOT_ACCEPTED; transport refusal/stop; no success marker|

- [ ] Implement literal assertions on file bytes/state/error categories and success-marker absence/presence, not merely shell exit code. Preserve original slice catch and errors. Require terminal runspace for each case; expected failure remains a matching test, never a passed observation.
- [ ] Validate all source pins, extraction deltas and destination paths without executing tail; save exclusive matrix.claim.json with MaxEvaluations10.
- [ ] Run `& ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-parent-audit-matrix.ps1` once. Expected: `PARENT_AUDIT_MATRIX_VERIFIED`,10 matches. First mismatch: persist actual files/outcome and remaining NOT_RUN, stop without repairs.

## Task3: Evidence and integration handoff

- [ ] Preserve input bytes, context, exact slices, actual outputs/errors and matrix result in an exclusive portable archive; decode/length/hash check each entry and compare originals. Tiny fixtures only; exclude runtime binaries/caches/secrets. Do not delete evidence.
- [ ] Update status/recovery/gap review; explicitly distinguish real exclusive-open collisions and synthetic hash faults from untested write/flush failures, disk-full, ACL failures, hung filesystem, and failure to persist the stop marker itself.
- [ ] Commit explicit necessary files, push existing development branch and verify remote equality/clean status. No main merge.
- [ ] If all cases match, propose separate formal-integration design covering repository-relative configuration, artifact ownership, source receipts, caller error/result contract, parent-audit deadline semantics and clean-machine prerequisites. Do not modify or invoke formal entry under this plan.

## Self-review and approval

Tasks cover selected spec refusal/persistence paths with real existing code and tiny files; untested failure classes remain explicit, not silently claimed. Source extraction is a test seam, not production refactoring. Inline/no-agent method retained under finite scope. PLAN_READY_FOR_REVIEW; implementation and tests pending user approval.
