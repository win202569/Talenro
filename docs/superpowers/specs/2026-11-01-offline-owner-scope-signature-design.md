# Owner scope auxiliary signature supplement

Status: PROPOSED; design/execution/additional computation approval pending.
This is a small diagnostic-test amendment, not another runtime architecture.

## Evidence and intent

Read [stopped attempt](../../roadmap/2026-11-01-offline-owner-scope-repair.md) and
[approved scope design](2026-11-01-offline-owner-scope-repair-design.md).
At baseline a8f4c46eff71c22e59fa1dd79e5ac7639e074f5a the frozen auxiliary
owner-scope.tests.ps1 lines94–95 asserts the error signature before saving it.
Actual output proves that assertion threw, not which component was unequal.
The original caught record was not saved; do not regrade the failed attempt.

Readonly comparison with PowerShell7.6.5 shows a boundary mismatch: the test uses
Get-Command but expects a bare CommandNotFoundException error identifier.
[GetCommandCommand](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/GetCommandCommand.cs#L868)
reports missing-command errors via WriteError; [ErrorRecord](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/ErrorPackage.cs#L1401)
appends the invocation type to the identifier. The corrected query-boundary
expectation is CommandNotFoundException,Microsoft.PowerShell.Commands.GetCommandCommand.
This is a source-derived expectation for a NEW attempt, not recovered old data.

## Minimal approach

Recommended: preserve the actual caught record before ANY assertion, then match
the exact query-boundary identifier and existing type/name requirements.
If predicate evaluation fails or differs, return a distinct unexpected diagnostic
containing the retained original plus any check error; outer recording exits2,
never labels it a valid RED final. A matching intended absence returns the same
nine-group1FAIL/8NotRun/exit1 model. Unexpected success also stops.

Do not directly call Invoke-ODProfile to imitate the old bare identifier: after
repair that could execute the profile body outside its runtime allowance. Do not
use prefix/substring/regex acceptance of arbitrary error IDs or suppress errors.
The original nine groups, actual initializer/function texts, parameter/body and
privacy checks remain unchanged. Only error preservation, exact signature and
version/provenance bindings change in a separately named auxiliary.

## Constraints and budget

Keep old auxiliary/input/intent/unexpected/tool record/stop/archive byte-exact.
Request1additional existing-tool replacement RED computation; it is not a refund.
Reuse the previously unused1GREEN only after the new RED is exact. Scope total
ceiling3computations includes the already failed1. Each new computation remains
at most2private modules, no child/profile body/record writer/Add-Type/target.
No additional runtime host, compile, target, snapshot or reviewer. Original
runtime remainder5hosts/7targets/1compile/conditional1review and snapshot ceiling155
stay frozen until their original gates pass. Native implementation, no new agent.
First unexpected result stops without automatic edit, refreeze or repeat.

The five-source runtime repair window and fixed writer-scope-green attempt/new
manifest remain the approved scope design's choices. Append exact supplement
provenance and preserve all four historical stops; a distinct new stop always
blocks. No main merge, broad acceptance, deletion or history rewrite.
