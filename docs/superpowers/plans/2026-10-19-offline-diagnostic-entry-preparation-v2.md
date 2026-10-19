# Offline Diagnostic Entry Preparation V2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline. Do not dispatch agents or reviewers. Track execution in a new ledger; do not edit this approved plan during execution.

**Goal:** Correct the confirmed optional-Role preparation defect without changing failed originals or reviving frozen runtime allowances.

**Architecture:** A new preparation-only script validates explicit source mappings and writes one new packet. It has no RunOne action or capture/target launcher. The historical runtime driver remains unchanged and cannot consume this packet automatically.

**Tech Stack:** Existing Windows PowerShell7 tool host, no nested interpreter, dependencies or installs.

**Spec:** [Approved migration design](../specs/2026-10-19-offline-diagnostic-entry-migration-design.md), [original plan](2026-10-19-offline-diagnostic-entry-migration.md), [actual stopped execution](../../roadmap/2026-10-19-offline-diagnostic-entry-migration.md).

Status: REVISION_PENDING_USER_REVIEW. User's latest continue authorized continuing the proposed correction/preparation work; this document makes the newly discovered directory-count and generation boundaries explicit. No implementation or new Prepare has run.

## Global Constraints

- Reuse clean linked development worktree at921289059f37938472fa72dba4b85a95d9cd508b; no main merge, new worktree, installation or security change.
- Original failed leaf, packet, portable summary, approved plan/design and historical evidence remain byte-identical. Old6hosts/7children/conditional1review remain frozen.
- New scope: at most two in-tool regression invocations (one RED, one GREEN) and ONE new preparation invocation;0capture hosts,0children,0reviewers. These are not restored historical allowances.
- No public/private runtime implementation, import, capture execution, broad suite, Bash helpers, extra setup, retry, deletion, reconstructed ErrorRecord or additional review. This explicit scope overrides skill defaults requiring those extra operations.
- Any unexpected outcome or recording failure stops all remaining work except readonly diagnosis and clearly unfinished checkpoint preservation.
- New packet is preparation evidence only, not permission to execute any runtime grant. Later runtime integration needs a separately approved plan/allowance.

## Review Focus

1. Missing optional Role must not throw; existing string or null Role must remain unchanged; never invent a classification.
2. Path/SHA and1930inherited file checks remain mandatory; absent Role cannot suppress identity checks.
3. The old leaf now has12files including portable summary. Check explicit source names, not the old eleven-file directory-count assumption; never delete the summary or count it as a runtime source.
4. Existing or partially written new packet blocks another Prepare; no retry, replacement path or deletion.
5. A prepared manifest must not become runtime authority. New script has no runtime action, and old runtime packet/path and frozen allowance stay unchanged.

## Task 1: Isolated metadata correction and one preparation

**Files:** Create new leaf `docs/diagnostics/isolated-fuzz/p2b-entry-preparation-v2/` with `.gitattributes`, `README.md`, `prepare.ps1`, `prepare.tests.ps1`. Final portable summary is recovery-only, not an input. Create new execution report and update current-status/repository-recovery at handoff.

**Interfaces:** Private `Convert-ODInheritedPin($Pin)` returns ordered VerificationPath/SHA256/Role/ExpectedState; optional missing Role becomes null, existing Role preserved. Required path/hash cannot be omitted or empty. `prepare.ps1` accepts no runtime action, arbitrary paths or caller callbacks. Tests evaluate only this real helper's AST function definition, never the script body or historical Prepare.

- [ ] Write regression before correction. Initial helper retains the exact failing direct Role access. Test real inherited1930pins (295missing Role), plus literal present-string/null cases; assert no input mutation, exact path/hash preservation and required-field rejection. Removing the optional-property guard must reproduce the same missing-Role failure.
- [ ] Invoke regression once in the Native tool host; require the confirmed PropertyNotFound failure before any preparation side effect. Save actual result without fabricating prior raw records.
- [ ] Correct only optional-property access using PSObject.Properties lookup. Keep required-field checks and all inherited SHA verification.
- [ ] Invoke the same regression once; require all cases pass. No extra dynamic test or broad suite.
- [ ] Implement fixed preparation body: verify original identities and exact explicit inputs; require new packet pristine except its new ledger. Use CreateNew writes and original failure-first behavior. Check old failed packet against its preserved portable bytes; do not restore files from it.
- [ ] Invoke Prepare once in tool host, no nested PowerShell. Require2075mappings:1930inherited+92old raw+16old leaf+7dependencies+25inputs+5failed-generation records;2067Present/8Absent. Require17actual source snapshots and0runtime processes. Save exact manifest SHA and original preparation result. Any mismatch stops, no correction/retry.
- [ ] Readonly verify bytes/lengths/hashes and unchanged historical records, preserve portable data, then explicitly stage necessary files, commit and push existing dev branch; verify remote SHA equals HEAD. No acceptance claim.

### Fixed mapping details

New packet: `.superpowers/sdd/2026-10-19-offline-diagnostic-entry-preparation-v2/`.
New manifest name: `preparation.frozen.json`; deliberately not a runtime grant name.

The7dependencies are original prestart execution report, closure map, approved design/plan originals, their two fresh approved-document snapshots and a snapshot of this reviewed revision. The25inputs are the original20planned source paths,4new leaf sources and this revision original. The5failed-generation records are its actual3packet files, original portable summary and stopped execution report. Snapshot only present inputs: original12+new4+revision1=17. Missing8runtime paths remain explicit Absent/null SHA, not fake files. All repeated path mappings must agree on state/hash.

Keep original design/plan identity checks from the failed driver; bind this revision original and snapshot to the digest recorded on human approval before execution. No runtime max-budget fields: manifest marks all capture/child/reviewer allowance0 and old allowances frozen. Static checks may parse source/JSON but must not import or invoke runtime files.

## Execution handoff

Review this preparation-only revision before execution. Native method is retained. Approval does not authorize later six-grant runtime migration or unfreeze the old remainder.
