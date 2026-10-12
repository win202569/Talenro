# SDD ledger — plan: docs/superpowers/plans/2026-10-08-task10-bound-epoch.md

User selected Native execution. Base: 6752b9612a00b9d6b0bf87ae8dd95063f6b056ee.
Existing uncommitted changes belong to earlier increments and are preserved.

Pre-flight: Task 1 produces BoundReadEpoch(store.DBTX) (uint64,error);
Task 2 consumes exactly that signature. Task 3 consumes their scoped evidence.
No interface conflict. Task 1 in progress; Task 2/3 pending.

Task 1: complete (local only). Isolated baseline PASS 2.946s. Missing symbols
observed, then fail-closed behavioral RED 3.053s for binding/range/expiry and
concurrent expiry; implemented source wrapper and scoped authority GREEN 3.100s.
The existing empty normalized snapshot is exercised with matching positive epochs;
zero database epoch is rejected rather than repaired here. gofmt applied.
Task 2: in progress. Exact producer/consumer signature unchanged.

Task 2: complete (local only). Missing reader symbols observed; fail-closed
behavioral RED 2.859s (success, cancellation, postcheck/query counts, panic).
Whole serving GREEN 2.647s. No exported source factory or reader added.
Task 3: in progress; fresh independent review and final regression pending.

Task 3 regression PASS: serving 3.493s, contracts 4.292s, readiness 3.198s,
authority selected TestBoundRead/TestCoordinator 3.468s, store compile-only
1.894s (no tests run), all exit 0. gofmt, tracked diff whitespace and new Go
file trailing-whitespace checks clean; all four plan lock hashes unchanged.
No test processes remain. Independent fresh-context review is running.

Task 3: complete (local only). Fresh independent review found no Critical or
Important issues in inspected read-boundary code and confirmed real cleanup
wiring; no fixes needed. Reviewer did not run tests. It inspected the complete
read-boundary implementation/tests, SQL and pertinent readiness context, not
every historical branch change. Final test evidence remains the executor runs
above; no product code changed after those runs.

Final: minor (deferred): distinct certificate/lineage/operation UUID fixture
values would catch more field swaps; current mappings correct, same prior minor.

Final: Ruling: actual PostgreSQL semantics/failover/pool mismatch remain unproved
— live DB is outside permission — cost: no physical integration acceptance.
Final: Ruling: no race-detector result claimed — atomic state and inspected
interleaving tests are narrower evidence — cost: dynamic races not ruled out.
Final: Ruling: end-to-end Coordinator-to-reader provenance remains unproved
— package-separated tests intentionally provide layered evidence — cost: a
future production integration still needs independent proof.
Final: Ruling: historical branch changes outside the read boundary are not
accepted — reviewer read relevant context, not all historical tooling/Coordinator
paths — cost: this is not whole-branch approval despite whole-branch review scope.
Final: Ruling: v7/production/HTTP/X.509/time/policy gates remain open — absent or
separate obligations — cost: private facts cannot authorize endpoints.
Final: Ruling: no SQL access revocation or secure-erasure claim — expiry is only
epoch metadata lifetime — cost: misuse of retained DBTX/runtime copies is not
prevented by this mechanism.
Final: Ruling: Defender/full suite/C11/C12/bootstrap/Docker/PITR/main/release
remain blocked or unaccepted — explicit safety scope — cost: no complete Task 10
or production acceptance. Preserve worktree and all local files; no Git writes.

Task 2: Ruling: observe staged aliases in the real guard/adapter/discard
composition tests, while the private reader tests verify zero publication,
query counts and panic propagation — its intentional encapsulation has no
staged-facts observer; adding a production leak hook solely for tests is worse.
Cost: reader-specific wiring of discard also requires explicit code review;
the alias assertions are compositional evidence, not end-to-end reader evidence.

Ruling: use this portable ledger and direct PowerShell test commands instead of
Bash-only task-start/task-done/workspace scripts; preserve the ledger without
committing or deleting it — PowerShell-only and no-commit user scope governs.
Cost if wrong: bookkeeping differs from skill automation, not product behavior.
Ruling: only the explicit isolated test allowlist, no dependency setup/full suite
or branch finishing/merge operations — unresolved security gate and approved
scope govern. Cost: no full-system or production acceptance evidence.
