# P2-A declared-integer timebase plan

Approved Native execution: at most five ordinary PS7 targets in the order red,
full, phase-mutant, expiry-mutant, final; one conditional readonly review.
This frozen portable ledger describes the route, not a successful execution.
Actual consumption/results belong in ../../../../roadmap/2026-10-18-p2a-timebase-execution.md.

The public interface is New-OfflineTimebasePlan(Request). It accepts declared
integers only. Created does not establish a real clock, source identity, runtime
authority or acceptance. No live adapter or process supervision is implemented.
Helpers are dot-source visible, not a security boundary. No PS5 evidence.

Tests have 24 ordered groups and independent full literal comparisons. Two
specific temporary mutants must fail at their designated assertions. No retries
or repairs are authorized after an unexpected outcome; restore the approved
temporary mutant even on failure. CodeProperty rejection and internal-error
handling are static-only checks. Raw String held-view mutation is distinct from
fresh property reread persistence. No concurrency/type-data/catastrophic guarantee.

Controller: Prepare (module absent), Run -Slot red, SealProduct, then Run for the
remaining four slots. Each invocation requires the printed FrozenSHA256. The
controller creates a new ignored local packet, preserves raw separate byte streams,
checks inherited pins and persists consumption before launch. Test hosts are never
elevated. Git-only permission is separate. Do not reuse or overwrite packets.

Fresh machines recover source/tests/oracles/reports from the development branch,
not ignored raw receipts, caches, sessions or automatically main. The historical
pin chain is local evidence, not a prerequisite for reading the pure source or
portable tests. Re-execution needs its own explicit allowance and prepared host;
this packet's finite allowance is not transferable to a fresh machine.

Skill adaptation: native PowerShell ledger/controller replaces Bash helpers;
retain all raw packets. The exact approved finite route overrides generic extra
baseline/full-project/task-done, fix/retest/review loops. No runner acceptance.
