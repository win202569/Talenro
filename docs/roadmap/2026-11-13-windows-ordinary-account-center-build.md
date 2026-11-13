# Ordinary account-center preparation — Task1 execution

Date2026-11-13, baseline80db7a35b6e9a0bbd424005a64c4f6bf7e62834d.
User approved isolation configuration and managed tests only. No App restore,
native build or launch authorized/executed; Tasks2/3 remain gated.

## Implementation

OrdinaryBuildVariant=account-center selects App-only bin/ordinary-account-center/
and obj/ordinary-account-center/, with absolute NuGet extensions path. The new
first initial target ValidateOrdinaryBuildVariant rejects unknown/path-like values,
non-Debug, enabled/malformed diagnostics, email variant or close-verification mode
with ORDINARYBUILD001. Empty variant preserves existing behavior; old guards stay.
No App XAML, business logic, observer, lockfile or launch entry changed.

## Fresh verification

Each test command used an ordinary token and successful related-process query0;
--no-restore, no App project reference build. App was evaluated or ran validation
targets only. Core and Tests compiled in normal paths as explicitly authorized.

| Run | Result | Command wall time | Test time |
| --- | --- | --- | --- |
| New OrdinaryBuildIsolationTests RED |16 total,4 pass,12 fail,exit1|19.773s|12.083s|
| Same tests GREEN |16/16,exit0|13.315s|11.268s|
| OrdinaryBuildIsolationTests + EmailGeometryIsolationTests |34/34,exit0|14.988s|13.882s|

RED failures: missing selected output path (actual bin/), and11 invalid-mode cases
lacking required ORDINARYBUILD001. Baseline/dependency comparisons passed. No
infrastructure failure, timeout, retry or process termination. GREEN/regression
had no failed/skipped tests. This is scoped regression, not a full-suite result.
Tests exercise real MSBuild evaluation; selected App diagnostic types/resources
are excluded, reserved-argument code retained; Core empty RID and dependency
paths preserved. Existing geometry tests unchanged.

## Preservation and environment

Linked development worktree, no superproject, clean starting sources. Candidate
bin/obj absent before and after. Recursive preflight rejected reparse entries.
Inventoried4278 App bin/obj/receipt/lock files; after all tests every hash unchanged,
no missing or additional App files. Saved205 Core/Tests bin/obj files and verified
every snapshot hash before testing.24 existing mutable files changed as expected
in Core/Tests compiled output, debug metadata/cache and copied App test sources;
not a claim all workspace outputs stayed unchanged. Snapshots remain local.
Postflight successful process query:0 relevant processes. git diff --check passed
(Git noted normal LF-to-CRLF normalization on the two configuration files).

Local ignored evidence: .superpowers/sdd/2026-11-13-windows-ordinary-account-center-build/.

| Evidence | SHA256 |
| --- | --- |
| red.log | DD9DF37563C12C1FD91CC63609260BC295B71E592A4CDE7A71042ADFA3BE72E5 |
| green.log | 949B87D5A6DB693670E173254BF425D538CC4DC316CD9D89A745B9CD1949DC2A |
| regression.log | 346A111910405C5031B1516F75888B99AC84CB6A018D70AC1F6E028EFB6A5F68 |
| evidence-task1/protected.json | 48BD5814C59626B0F8AD6E3284409A72400CFC262F5E4CFA634784C58F4E009F |
| evidence-task1/mutable-before.json | 6D9F2C643143E7CD80EF3876798D6C1F1DEBE83D46923A61E9017EA1D37BD008 |

## Rulings and scope

- PowerShell/apply_patch replaced Bash-only skill bookkeeping; same plan identity,
  BASE and ledger retained. Cost: manual bookkeeping rather than helper execution.
- Approved RED/GREEN/combined scope controls tests; no extra setup restore/install,
  baseline/full-suite rerun. Cost: broader suite is not freshly verified.
- Preserve plan workspace/snapshots rather than deleting them. Cost: local disk use;
  cloning does not recover this ignored evidence.
- Task1 only is authorized; no automatic continuation to native build or UI tests.

## Independent scoped review

Fresh reviewer ordinary_isolation_review inspected the Task1 diff/tests/recorded
logs against the approved design:0 Critical,0 Important,1 Minor. No reviewer
mutations or test/build/restore runs. Fixed App-only paths, first guard ordering,
predicate, existing guards and empty-variant behavior judged correct.

Deferred Minor: valid configuration is tested via evaluation, while direct guard
invocation covers only invalid configurations. Add a valid-target acceptance
case in a separately authorized future test scope; current predicate is correct
on inspection. Cost: a future always-rejecting guard mutation is not directly
caught by this positive test. No extra tests were run to hide this limitation.

Reviewer declined-to-judge dispositions: actual restore/cache/path preflight and
generated XAML/XBF/binary inspection remain Tasks2/3, not accepted here; native UI,
shutdown, Release/security and overall completion remain outside scope. Full
artifact/snapshot rehash relies on coordinator evidence, not independent reviewer
reproduction. Git delivery is verified separately by coordinator after commit.

Task1 scoped implementation/testing checkpoint is complete with the deferred
Minor, not whole-plan completion. No main merge or C4 acceptance. Ordinary
candidate binary/hash does not exist yet. Next gate is Task2 explicit one local
locked restore plus conditional one build; no application launch included.
