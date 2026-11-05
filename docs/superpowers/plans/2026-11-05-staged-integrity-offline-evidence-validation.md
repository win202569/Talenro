# Staged integrity offline evidence validation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Test real saved-evidence rejection paths before considering formal C12 integration.

**Architecture:** Restore minimal fixture copies from portable archives into a new owned directory. Run a path-rebound copy of the existing v2 verifier in a fresh in-process PowerShell runspace per case; mutate only fixture evidence, never original files or verifier decision logic. This is a proposed diagnostic harness, not a new runtime gate.

**Tech Stack:** Existing PowerShell7/Core, .NET PowerShell runspaces, JSON/UTF8/SHA256; no added dependencies.

**Spec:** Existing approved docs/superpowers/specs/2026-11-05-runtime-staged-integrity-design.md refusal requirements, narrowed by docs/roadmap/2026-11-05-staged-integrity-integration-gap-review.md. New harness approach and budget below await user approval. Native inline/no-agent method retained.

## Global Constraints

- No implementation or tests authorized merely by saving this plan. Actual44/34/6/217 unchanged.
- Zero business computations, child processes, compilation, snapshots, Docker/Go, installs, full suite or runtime replay. In-process runspaces are new offline fixture evaluations, not operating-system processes; report them separately.
- Preserve all original scripts, archived bytes, claims and verdicts. New workspace `.superpowers/sdd/2026-11-05-staged-evidence-fixtures-v1`; refuse an occupied case directory, never delete/reuse it.
- One RED and one GREEN bootstrap test invocation (memory/path-rebinding only), then one matrix invocation with at most16verifier evaluations. Stop on first unexpected result; no automatic repair/retry. Expected refusals are successful tests, not reasons to stop the matrix.
- No weakened schema, transport/audit/ownership/output gate or widened timeout. Rebinding is limited to fixture root and output paths; no decision-code substitution.
- No claim of bounded filesystem wall time or real disk-fault coverage. If runspace completion is uncertain, retain ownership/resources and stop; do not dispose an active pipeline or launch another case.
- NOT_ACCEPTED/all five flagsfalse/main unchanged even when matrix passes. Diagnostic counters stay44/34/6/217; record actual fixture invocation count separately.

## File map / interfaces

L = docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2.
Create L/runtime-staged-evidence-fixtures.ps1 (definition-only fixture and runspace helpers), L/runtime-staged-evidence-fixtures.tests.ps1 (bootstrap tests), L/runtime-staged-evidence-matrix.ps1 (one bounded matrix).
Create workspace progress.md, per-case inputs/rebound-verifier/outputs and matrix-result.json; portable output L/portable-runtime-staged-evidence-matrix-v1.json.
Update current-status, repository-recovery and the integration gap review at handoff.
Do not modify runtime-staged-integrity-review-v2.ps1, any consumed input, or scripts/run-c12-integration.ps1.

## Review Focus

1. Fixture rebinding accidentally reads original data: enforce exact replacements, root containment and absence of old absolute root (bootstrap B01/B02).
2. Invalid audit wrongly accepted through coercion: matrix C04/C06/C07/C08 uses serialized artifacts, not fabricated Boolean facts.
3. Hash mismatch masks map-content validation: C10 updates reference digest/length to match the mutated map but retains original expected pins.
4. Script exit terminates the controller or success is inferred from text: B03 establishes in-process isolation, matrix requires persisted typed verdict plus terminal invocation state.
5. Missing output makes unsafe branches appear covered: C12/C13 also remove fixture output, but this only verifies refusal ordering/reason, not a general I/O interception proof.

## Task1: Bounded fixture harness bootstrap

**Interfaces:**
- `ConvertTo-SEFixtureVerifier([string]$Source,[string]$FixtureRoot,[string]$ReviewRelative)` -> string, exact allowlisted replacements of root and review directory only; data-relative paths remain unchanged.
- `Invoke-SEFixture([string]$Script,[string]$FixtureRoot,[string]$InputSHA256)` -> dictionary Terminal(bool), HadErrors(bool), Output(string[]), Errors(string[]). Fresh `[PowerShell]::Create()` runspace, set its location, invoke supplied script with InputSHA256, retain streams; dispose only after terminal invocation. No Process/Start-Job/assembly compilation.
- `New-SEFixture([string]$CaseId,[string]$Workspace)` -> dictionary Root, ReviewRelative, InputSHA256. Only allow CaseId C01..C16 and canonical descendants of the new workspace, refuse existing roots.

- [ ] Write bootstrap tests before helper: B01 rebind actual v2 source and inverse-transform to byte-equivalent decision code; B02 reject unsafe/occupied roots and unexpected source anchors; B03 run harmless `param($InputSHA256) 'fixture'; exit 1` in a runspace and assert controller survives/Terminal=true, then harmless success in a separate runspace. No verifier/artifact workload during bootstrap.
- [ ] Run `& ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-staged-evidence-fixtures.tests.ps1` once; expected missing helper RED. Save receipt.
- [ ] Implement those interfaces. Restore archive entries only after canonical path validation and byte/hash checks; reject absolute/traversal/reparse destinations. Never dot-source archive scripts during restore. Verify observation-final SHA256 EEBC1498D8DD1760D36B37CB79EB7C56199AF5B62AFFBB8F8D501D17033DC905 and review-v2 SHA256 5943F342A08B6ADF65A1605267876905F2A3E0934BCA31B9BDEB4BE3ADD1AB87.
- [ ] Fixture inputs: original observation records/maps/helper/output from those archives; original comparison stages.jsonl from portable-runtime-copy-transport-final.json, verify its entry against the observation manifest historical pin. Exclude original review outputs so new exclusive verdict files can be created. Copy no DLL, executable, runtime driver or producer into executable fixture locations. Keep manifest paths/hash values unchanged: verifier compares maps, not a new full source scan.
- [ ] Run bootstrap once GREEN; require `STAGED_EVIDENCE_BOOTSTRAP_VERIFIED` and three case results. Preserve unexpected errors and stop without expanding scope. Commit necessary helper/tests/progress before matrix.

## Task2: One sixteen-case evidence matrix

Each case gets an independent baseline fixture. Never compound mutations across cases.
Run the identical rebound v2 decision code; mutation code belongs only to test utilities.

|Case|Single mutation relative to baseline|Expected|
|---|---|---|
|C01|None|ObservationPassed true, STAGED_PREFIX_VERIFIED, ten typed true facts|
|C02|Remove integrity-review.json|Reject, no passing verdict|
|C03|Audit State FAILED/After null/Error fixture failure|Reject|
|C04|Audit malformed JSON|Reject|
|C05|Audit InputSHA256 becomes a different64hex value|Reject identity|
|C06|Audit IntentSHA256 becomes a different64hex value|Reject identity|
|C07|Audit duplicate State keys: FAILED then VERIFIED, raw JSON|Reject malformed/ambiguous audit|
|C08|Writer Associated becomes JSON string "false" in supervision and matching result.Result|Reject wrong Boolean type, not just row-copy mismatch|
|C09|Remove supervisor.after-map.json|Reject|
|C10|Change one map hash, recalculate map ref hash/length in audit and result|Reject expected-pin mismatch|
|C11|Writer actual ExitCode1 in supervision and matching result.Result|Reject actual exit even if role-result says0|
|C12|PendingResources true in supervision/result, omit output files|Reject terminal state before output-file error|
|C13|ExitUnconfirmed true in supervision/result, omit output files|Reject terminal state before output-file error|
|C14|Trace truncated to three complete records|Reject incomplete trace|
|C15|Replace empty stderr with one byte|Reject length/output|
|C16|Preoccupy new review/verification.json with sentinel non-verdict bytes|Reject, sentinel bytes unchanged|

- [ ] Implement matrix runner with these literal expected results and per-case mutation descriptions. Reject unknown cases. Persist each input snapshot and actual output, including failure reasons; never convert exceptions to passing observations.
- [ ] Before matrix, validate paths and allowed source delta once without invoking the verifier. Pin bootstrap code/receipts and archive original input identities; do not run the original source or recalculate a new input identity to hide an unexpected mutation.
- [ ] Run `& ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-staged-evidence-matrix.ps1` once. For C01 require a new typed passing verdict and terminal/no errors; C02–C16 must not produce or preserve a passing verdict, must match intended rejection category, and must not write outside their case root. Existing sentinel is not a verdict. Stop immediately on first unexpected acceptance/error; record remaining cases NOT_RUN. No code repair authorized by this matrix.
- [ ] Record results as evidence of the offline verifier only. If C07/C08 expose parser/coercion defects, open a separate minimal repair proposal; do not loosen expectations to obtain green. Case16 does not simulate disk-full/ACL failure, and no fixture proves actual parent audit/save behavior.

## Task3: Handoff gate

- [ ] Verify original source/archive hashes unchanged; preserve fixture inputs, outcomes and bootstrap receipts in an exclusive portable archive with containment/length/hash checks. Do not archive caches, binaries, credentials or unrelated files; no cleanup/deletion of evidence.
- [ ] Update gap review with PASS/FAIL/NOT_RUN by case and remaining blockers. All16expected outcomes only allow planning parent-failure tests/formal integration, not executing them. Failure requires a new scoped repair decision.
- [ ] Commit explicit necessary files and push existing development branch; verify remote equality and clean status. Stop; no main merge, new runtime budget or full-suite acceptance.

## Self-review / approval

Identity/map/late-exit/partial-trace refusal requirements map to C02–C15; persistence refusal to C16; fixture contamination/isolation to B01–B03. Real parent write failures, hung disk, dynamic aliases, clean-machine setup and formal C12 lifecycle remain deliberately outside this gate. The source review is not a reproduced vulnerability finding. PLAN_READY_FOR_REVIEW; native inline execution only after user approval.
