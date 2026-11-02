# Pin-binding recovery amendment

Status: PROPOSED_DESIGN. User “是” after ca8486df authorized preparing this
recovery revision, not this written design, additional allowance or execution.
Small amendment to the approved call-model design; native method retained.

## Evidence and goal

[Actual stopped execution](../../roadmap/2026-11-02-offline-contract-call-model.md):
computation10of11 exited2 before all20groups (0PASS/0FAIL/20NotRun).
Actual PropertyNotFoundException/PropertyNotFoundStrict names SHA256 in
Get-ODPinnedSHA256. The selected source pin exists uniquely with a SHA256 field.
The function uses $matches for pin rows, then -cnotmatch before reading
$matches[0].SHA256. Collision with PowerShell automatic $Matches is the inferred
cause; no separately executed reproduction or verified fix exists yet.

Preserve stopSHA8581A9DA9A4CA3E7F2367757BE7616992889B434F42B1C4F35922B37921633ED
and actual finalSHA900141E27B3A9201AD3F7552B779DEEADC989632498432F474670BD9EE2297F9.
The failed attempt is not an expected descriptor RED and is never reclassified.
Goal: prove and repair pin-array corruption, then independently establish the
original missing-descriptor failure before implementing that descriptor.

## Approaches

Use the narrow local-variable repair direction discussed with the user: rename
the pin-array local to pinRows throughout Get-ODPinnedSHA256. Keep selection,
Present state, uppercase64hex validation, duplicate-hash consistency, actual
file hash and failure messages unchanged. No new API or error normalization.

Replacing the regex with another validator would add semantic changes unrelated
to the collision. Skipping validation or treating arbitrary setup errors as
expected RED would remove the evidence boundary. Neither fits this repair.
The rename is proposed until the focused actual-source regression confirms it.

## Verification sequence and allowance request

Request THREE additional finite auxiliary computations, cumulative ceiling14.
Actual consumed remains10; one previously unspent GREEN allowance is explicitly
reassigned to the versioned final call-model GREEN. Four future computations:

1. New pin-binding RED: actual extracted old Get-ODPinnedSHA256 on the real
   pinned record source; require the exact observed missing-SHA256 failure.
   This is a new focused regression, not a rerun of the consumed20group attempt.
2. Pin-binding GREEN after rename only: valid real pin returns its independent
   file hash; missing pin, Absent state, malformed hash, conflicting duplicates
   and mismatched actual hash reject under their original exact checks. Duplicate
   identical pins remain accepted. Pin mutations are explicitly test-only data;
   no historical manifest/file is edited. Failures are not broadly caught as PASS.
3. New versioned descriptor RED after pin GREEN, before descriptor changes:
   actual mapper now reaches the unique-descriptor assertion; require intended
   0PASS/1FAIL/19NotRun/exit1 with exact descriptor-missing signature. Preserve
   this intermediate source identity in immutable provenance, not a runtime epoch.
4. Versioned final call-model GREEN after the original three-source work:
   retain the original20 behavioral groups, exact actual saved stack acceptance,
   negative bindings, historical/source-window/guard checks. Require20PASS/exit0.

Freeze focused and versioned test/input identities before their first execution.
Original contract-call-model.tests.ps1 and its inputs/failed receipt stay immutable.
Versioning is necessary: the old test hardcodes proposed2263pin metadata and old
input/epoch selectors; new recovery provenance changes those identities. Change
only those bindings plus the separately specified focused pin tests, not original
call-model assertions. The later plan must define exact paths, group counts,
schemas, new provenance count and checks before any computation.

No profile/control body, module probe, record writer, child host, fixture compile
or target in auxiliaries. No exploratory computation outside these four stages.
First unexpected result stops without patch/retest, refunds or extra review.

## Source, historical and stop boundaries

Keep the same three runtime files: source-model.ps1 (pin local repair plus
previously approved descriptor), observation-contract.ps1 (strict pure bridge),
driver.ps1 (new recovery identity/provenance/stop gates). Other28sources and
existing runtime tests/oracles remain frozen. No generic dynamic-call resolver,
anonymous-source runtime fallback, relaxed operation/source/stack acceptance,
or ownership/termination/timing/recorder changes.

Preserve all eight stops,155snapshots, original test bytes, receipts and archives.
New fixed pin-binding recovery stop always blocks. The previous call-model stop
is recoverable only by exact approved history hashes and all new actual receipts;
no public bypass or arbitrary callback. Old preparations and attempts remain spent.

Accepted writer-projection-repair manifest and writer result remain immutable,
validated against their existing snapshots, not new source bytes. Preserve both
historical recovery maps and add separate new provenance. Keep still-absent
repair-green.frozen.json, snapshots/repair-green and the original four runtime
slots; no selector change in entry/session/runtime tests. Avoid driver/input
hash cycles; derive exact new pin count from a fixed unique list in the plan.

No additional runtime hosts/compiles/targets/snapshots/reviewers requested.
Consumed3hosts/3compiles/155snapshots/0targets/0reviews stays unchanged during
drafting. Original remainder4hosts/7targets/0compiles/31snapshots/conditional1review
remains gated by all auxiliary results, then original contract52PASS/core25PASS/
lifecycle17PASS/continuation16PASS+12seamPASS chain. Cumulative ceilings7hosts/
3compiles/186snapshots; five capability flags false. Do not rerun accepted writer.

## Delivery and approval

Success or stop gets a NEW actual-record portable archive with decoded-byte
verification and ordinary necessary-file commit/push to the current development
branch, remoteSHA equality verified. Do not overwrite old archives or stage raw
ignored records separately. No main merge, deletion, install or security change.

Written design approval precedes a separately approved implementation plan and
allowance. This design spends nothing and authorizes no source/test edit or run.
Whole I1/I2/product remains NOT_ACCEPTED.
