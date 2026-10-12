# Main-state contract L0 execution — 2026-10-10

Status: implemented locally; one fresh source review completed and its three Important findings reproduced and fixed. Post-fix full scoped regression passed. Two review Minors deferred. Not committed or pushed.

Subsequent 2026-10-10 follow-up closes both deferred Minors; see `2026-10-10-main-state-contract-followup.md` for the classification RED→GREEN, all-collection ownership coverage and fresh scoped regressions. The original review/execution history below is preserved. No second independent review.

Approved design/plan: `docs/superpowers/specs/2026-10-10-task10-main-state-contract-design.md` and `docs/superpowers/plans/2026-10-10-task10-main-state-contract.md`.
Native execution in existing `codex/c12-b01-task9-coordinator` worktree. HEAD baseline `6752b9612a00b9d6b0bf87ae8dd95063f6b056ee`.

New isolated `internal/nodecontrol/statecontract` models twelve fixed domains, distinct guard/key identities, explicit coverage, bounded typed references, local cross-bindings, previous/proposed monotonic checks, bounded root chains, owned snapshots, fixed outstanding L1–L5 responsibilities and exact rechecks. It performs no material parsing, signatures, hashing, I/O, platform verification or authorization. No production consumer or proposal bridge.

## Author evidence so far

- Baseline: eight existing allowed packages passed, exit 0.
- Task 1: missing-interface compilation RED, followed by shape suite GREEN2.238s and expanded package GREEN1.654s, exit0.
- Task 2: missing-validator compilation RED, full package GREEN1.692s, exit0.
- Task 3: missing-transition compilation RED, full package GREEN1.893s, exit0.
- Task 4: missing-model compilation RED, full package GREEN1.894s, exit0.
- Task 5 boundary: missing-helper compilation RED observed; implemented helper; new-package suite including boundary passed.
- Final pre-review-fix suite, all exit0: statecontract8.765s, staterecovery8.647s, statetime6.471s, stateformat6.510s, stateverify10.512s, statepolicy7.087s, serving7.327s, contracts7.817s, readiness6.265s; selected authority6.489s; store compile-only8.219s (no tests).
- All 387 preexisting Go/SQL/module files match pre-edit SHA256 baseline. All new product/test changes are confined to the new package.

Compilation RED establishes absent interfaces, not individual behavioral mutation coverage. Independent review is source-only and is not independent execution of author tests.

## Rulings

1. PowerShell-only host: native commands and apply_patch ledger replace Bash skill scripts. Same task/evidence bookkeeping is retained; cost is manual audit of bookkeeping.
2. Approved no-commit/no-cleanup and scoped-only test rules override generic skill defaults. Existing work is retained; integration outside the approved tests remains unknown.

## Remaining gates

One final fresh gpt-6-astra high read-only source review found 0 Critical / 3 Important / 2 Minor. No second independent review.

- Visible desired identity conflicts across fields/snapshots: reproduced through EvaluateModel and fixed by reconciling all visible desired bindings before other cross-bindings.
- Nonadjacent root payload reuse and repeated rotation material payload: reproduced and fixed by whole-chain identity checks.
- Plan defect requiring historical authorizers to remain current: corrected explicitly. Persisted Desired/Recovery/resource/metadata references may retain older materials, while visible version/payload conflicts and relabeling original snapshot authorizers remain rejected. Their current authorization remains unproved.

All three public-entry regression tests failed before fixes (2.099s, exit1); full new package then passed2.003s, exit0.

Post-fix full approved suite, all exit0: statecontract9.761s, staterecovery8.210s, statetime6.910s, stateformat6.863s, stateverify6.801s, statepolicy6.140s, serving8.400s, contracts6.780s, readiness4.126s; selected authority6.350s; store compile-only9.907s (no tests). All 387 prior Go/SQL/module files still unchanged, fixed gofmt clean. No second independent review.

Final test-only coverage audit explicitly checked every persisted time field against both out-of-range bounds and valid inclusive endpoints; full new package passed1.804s, exit0. Existing shape checks were already correct; no artificial behavioral RED was claimed. No product changes after the full post-fix suite.

Final checklist also added explicit zero-threshold, zero-epoch and non-RFC4122 UUID cases; latest full statecontract run passed1.852s, exit0. This was test-only coverage of already-correct checks, with no product changes.

Deferred Minors: ownership regression tests do not populate/assert every collection though copies appear correct; nested coverage failures sometimes return InvalidInput instead of IncompleteCoverage (still safely rejected). The plan retains the unchecked exhaustive ownership-coverage step rather than claiming complete case-by-case coverage. HEAD unchanged; no staging, commits, pushes, merge or cleanup.

L1–L5, real material closure, cryptography/domain policy, physical seal/counter/durability, security locks/startup, Linux/hardware/reboot, production compatibility and overall Task10/B01/C12/main acceptance remain unresolved. No whole-repository, Docker/database, race, or platform tests were authorized or run.

## Review exclusions and executor rulings

- Raw parsing/signature/digest/closure and actual history/tombstones remain L1 obligations; unknown material authenticity is not inferred. Other unchanged-digest declarations remain deferred material binding.
- Certificate/CA/recovery/supervisor/LKG semantics, deadlines and authorization remain L2/L5 obligations; no execution/startup capability results.
- Physical isolation/time-provider, coding/seal/disk/counter/reboot/startup remain platform/L3–L4 gates; labels are not proof.
- Initialization/migration/reenrollment/reset/key replacement and concurrent mutation/Clear/complete erasure remain excluded; use outside this contract is unsupported.
- AST guards are development checks, not a hostile-code sandbox. Review did not independently run tests/hash checks or audit unrelated work or overall main acceptance; runtime evidence is author-only.
