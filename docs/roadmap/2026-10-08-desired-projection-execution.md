# SDD ledger — plan: docs/superpowers/plans/2026-10-08-task10-desired-projection.md

Approved Native execution; base 6752b9612a00b9d6b0bf87ae8dd95063f6b056ee.
Pre-flight: Task 1 produces private types/constructor/snapshot/discard; Task 2
consumes only its evidence, no shared interface conflict. Task 1 in progress;
Task 2 pending. Serving baseline PASS 2.918s. Existing dirty files preserved.

Ruling: use portable ledger and PowerShell instead of Bash-only bookkeeping;
no commit, scratch deletion, full suite or dependency setup — approved platform
and safety scope prevail. Cost: manually recorded evidence, no full acceptance.

Baseline locks match prior recorded four hashes. Unchanged-file controls:
SQL 11A0A62B225468C0FF246EFE6ECBC621CEF172D38B7ED48BD1D7EDC78A06F298;
generated query AF5D164980E9847C55561BDD713497A3F8D9BF4FF8073342A024CE6B7C9F31BC;
Querier 966AA32E417EC67D52B7D985A3CF4EF1EBD25A226182839116336EA7A0BB63EF.

Task 1: complete (local). Missing symbols observed; fail-closed behavior RED
3.518s (positive structure/boundaries and invalid-argument classification).
Initial whole-serving GREEN 1.410s. Discard no-op RED 2.382s (retained bytes).
Temporarily removing desired payload clone made Ownership fail in 3.134s;
restored immediately. Full serving GREEN after actual cleanup 2.605s.
No production parser/query/reader or unrelated file edits introduced.
Task 2: in progress; fresh-context review and final regression pending.

Final scoped regression PASS: serving 2.527s, contracts 3.188s, readiness 2.537s,
authority selected TestBoundRead/TestCoordinator groups 4.240s, store compile-only
2.501s (no tests run); all exit 0. Four dependency locks and three SQL/generated
control hashes match baseline. gofmt/whitespace checks clean. No test processes
remain; reviewer final report pending.

Task 2: complete, local only. Independent review found no Critical/Important
issues in this increment. Final scoped regression above independently run by
executor, not reviewer. No code changes after that regression.

Final: minor (deferred): M1, Ownership test XORs shared input tail/ActiveRoot
buffers twice, canceling those mutations; explicit internal alias check covers
payload but not envelope. Strengthen by mutating/asserting once and checking
envelope isolation. Implementation currently clones both correctly.
Final: minor (deferred): previous certificate fixture reuses UUIDs; unchanged.

Final: Ruling: historical branch work is not rereviewed — ownership protection
prevented reviewer Git access, parent supplied inventory — cost: not whole-branch
approval. Do not change Git security configuration to work around this.
Final: Ruling: final regressions/docs are executor evidence — reviewer ran no
tests — cost: review and execution are separate rather than duplicated proof.
Final: Ruling: SQL/fence/active intent, trusted chain window, canonical decoding,
signatures and current-time authorization remain absent — phase 1 is structural
only — cost: private projection cannot be served as authenticated facts.
Final: Ruling: live DB/PITR/v7/full C12/race/security/production/main gates remain
unaccepted — outside permission — cost: no Task 10 or release completion claim.
Final: Ruling: runtime-copy secure erasure and concurrent caller mutation are
not guaranteed — owned staging contract only — cost: callers must not race
mutation with construction; clearing is not a secure memory erasure guarantee.
Worktree and portable ledger retained; no commits, push, merge or cleanup.
