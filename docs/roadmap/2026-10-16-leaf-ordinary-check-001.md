# Ordinary check001 — stopped before preflight execution

User explicitly approved the [concrete plan](../superpowers/plans/2026-10-16-leaf-ordinary-check-001.md)
with “是” after proposal checkpoint2734cad9da2a82670be5cdd7d1c261850526d72c.
No real ordinary package result was obtained.

## Observed failure

The executor's first preflight command contained an unquoted Bash-style path
brace expression: `docs/diagnostics/isolated-fuzz/{isolation,envelope-sample,provenance}.ps1`.
PowerShell reported ParserError / “Missing argument in parameter list” at the
comma and exited1 before executing the command body. This was an executor
command-construction error, not a product/Go/collector/test failure. The preceding
read-only skill/plan/spec reads succeeded; they did not invoke the controller.

No preflight Git/helper command in that failed command body ran. No control
wrapper or output root was created, no controller was called, no native build
or runtime was attempted, and no SDK/product audit or package acceptance was
obtained. Post-stop read-only inspection confirmed both proposed directories
leaf-ordinary-check-001 and leaf-ordinary-check-001-control remain absent and the
development worktree was clean before this documentation handoff.

There is no controller stopped.json/receipt to claim: the failure occurred before
the controller/control directory existed. This portable report records the
actual outer-tool failure, rather than inventing native evidence.

## Scope and next decision

The approved plan explicitly says even a failed preflight returns for a new
decision. No corrected preflight or controller call was attempted. Task1 is
incomplete;Task2 never started;Task3 records this stopped preparation rather than
a completed native attempt. Checkboxes are not marked complete for unperformed
work. Historical error87, real matrix003 failure, timeouts/fuzz and broader
C12/main/production gates remain open;PackagePassed is unestablished and overall
Acceptance remains false.

Next recommendation: authorize correcting the executor's path list to explicit
PowerShell-compatible file arguments, parse-checking the complete wrapper before
any workload boundary, then restarting preflight under a revised single-run
decision. Because001 never created a root or crossed the call boundary, its name
can only be reused if that new decision explicitly permits it;the previous
approval must not silently be replayed. Do not change product/helper source or
weaken evidence gates. No correction or live retry is claimed here.

Ruling: retain the existing worktree and PowerShell-only bookkeeping;no Bash
skill scripts, installation or whole-project baseline — the approved operational
scope permits only this bounded run — cost: no broader regression evidence.
Ruling: stop on the outer parser failure, with no implementation/reviewer/fix
cycle — no implementation changed and the operational stop rule governs — cost:
the intended real validation remains undone pending a new decision.

Recovery: use latest codex/c12-b01-task9-coordinator tip, not main. This report,
plan and status/recovery documentation are portable;machine inputs still need
fresh validation. No raw SDK/cache/binary data is committed.
