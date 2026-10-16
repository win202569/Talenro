# Retained-child observer phase fix — 2026-10-16

Subsequent separately approved [real observation002](2026-10-16-retained-child-observation-002.md)
passes sampled diagnostic checks after this fix, not original/product acceptance.
It does not relabel001 or establish an error87 fix. Details below record the
earlier fake-only implementation checkpoint.

## Approved scope and change

User approved only a bounded observer field-semantic fix with fake regressions;
no live experiment, gate change, matrix, Go, Docker or historical reclassification.
Base commit:a68263ee71d971525d4e23b86446c8943dbcf7d8.

retained-child-observation.ps1 no longer requires the schema3 root-wait snapshot
DescendantsRemaining to be0 as a final cleanup condition. The value is preserved
unchanged in the receipt; final ActiveAfterCleanup/MembersAfterCleanup and the
existing pretermination/natural/shared-cleanup checks still decide completion.
No collector, registry, ordinary acceptance guard, native entry, budget, receipt
schema, diagnostic flag, core-row coverage or finalization behavior was changed.

## TDD and regression scope

Existing entry suite passed61 onPS7 before edits. The synthetic native fixture now
sets literal TotalProcesses2/ActiveAtRootWait1/DescendantsRemaining1 with root-wait
member23 and verified final core evidence, but final accounting0/empty. The new
named assertion `root wait descendants do not imply final residual` failed against
the old observer, then passed after removing only the mistaken zero requirement.
No raw historical receipt/private cache is needed to reproduce this regression.

Two new positive checks cover that lifecycle and preservation of both phases.
The preservation assertion inspects the exact deserialized input after the
observer returns, not an unrelated original DTO, so input normalization cannot
hide a lost root-wait count.
Seven new negative mutations still reject final active/list residuals,
pretermination active/list residuals, natural501ms/shared5001ms and unverified
cleanup. Existing missing/ambiguous/forced/timeout/error87/core-row/diagnostic
refusal and finalization-failure tests remain. The controller success assertion
also verifies that its persisted receipt retains the nonzero root-wait count.
Expected values are literal; observation/conversion/persistence are real, while
only native process/fixture launch and sleep boundaries are faked.

Final fresh-host scoped verification passed on both PowerShell7 and Windows
PowerShell5.1, all ten commands exit0:

- retained-child-observation.tests.ps1:70 per host.
- retained-lifetime.tests.ps1:27 per host.
- retained-lifetime-integration.tests.ps1:12 per host (four static checks).
- retained-lifetime-receipt.tests.ps1:16 per host.
- short-child-observation.tests.ps1:23 per host.

Total148 per host/296 primary assertions. Original61 baseline and intentional RED
are separate evidence, not added to the final totals. This is five adjacent
fake-only suites, NOT the earlier13-suite389-per-host batch or full project suite.
Eight unrelated legacy synthetic wrapper suites were not rerun because they do
not import this helper. Actual-process/Go/full-project runs were excluded by the
approved scope. No native lifetime/timing/error87 resolution is newly established.

One independent read-only review found no Critical/Important/Minor issues in the
fix/test diff. It confirmed the native assignment phase, retained final gates,
literal positive fixture and seven negative controls. Review did not run tests or
judge live behavior, historical classification or broader subsystem guarantees;
those exclusions are accepted as outside this bounded fix. After review the
coordinator tightened the preservation assertion to inspect the observer's exact
input, then reran all five adjacent suites on both hosts; no product change or
second independent review is claimed for that test-only tightening.

## Preservation and continuation

Ignored local evidence is under
.superpowers/diagnostics/2026-10-16-retained-child/phase-fix-validation-001.
It retains the intentional RED, ten initial GREEN logs, ten final test logs,
source/context binding and summary; only the ten final commands count toward the
296 primary assertions. Final manifest/hash verification is recorded below. Old raw
observation001 remains stopped, result=false/no completed marker. Its original
artifact bytes are not rewritten or re-evaluated under the new helper.

The24-file local evidence manifest SHA256 is
`1EF004DDE543DFCE91B73520EE098C975F198E91CA9758B1233A99D097CD64F7`;
all24 hashes/lengths rechecked. Twelve final source pins bind this verification:
only the new observer/helper test change, the other ten match the earlier binding.
All ten original observation001 artifacts still match its unchanged manifest.

On a fresh machine recover the existing development branch and read current
status, [repository recovery](../runbooks/repository-recovery.md) and
[the stopped observation](2026-10-16-retained-child-observation-001.md). The portable
new regression needs no ignored receipts, SDK or private Go cache. Raw evidence
and local fixture directories are not committed or restored by cloning.

The wrapper defect is corrected in code/fake tests only. Observation001 stays
historically failed; another real observation would need a separately approved
new identity and scope, with no automatic replay. Error87's cause/fix is still
unestablished. Matrix003 remains2PASS/1FAIL/23UNRUN; broader C11/C12/Task10/B01/
production/main gates remain open. No main merge is authorized by this checkpoint.
