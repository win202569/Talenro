# Offline raw capture coordinator — execution record

Status: **STOPPED AT SLOT5, NOT ACCEPTED OR REVIEWED.** User selected Native option1 after
5915bc732a658973b368b40a681250388911854e, approving the
[plan](../superpowers/plans/2026-10-17-offline-raw-capture.md)'s at most10
sequential synthetic test hosts and one conditional readonly reviewer.
No actual collector/process/watchdog or Stage B authority was granted.

## Frozen input and source

New test SHA256:7010CBCF535A431D62AEA7C3392971E96F46769801DF2075651CA87B3B8D88D7.
New implementation SHA256:4C0E6C10DAB31C63E550416260B0672EB0D721A8210B31CCABEA05648DCFB673.
All12 prior [dependency source/test pins](2026-10-17-synthetic-capture-dependency-regression.md)
matched before execution. Existing source/tests were not edited.

## Observed slots

|Slot|Engine|Exit|Elapsed ms|Observed|Disposition|
|---|---|---:|---:|---|---|
|1|PS7|1|1678|exact RawCaptureApiRequired assertion, no stderr|expected RED|
|2|PS5.1|1|4044|same missing-API assertion, no stderr|expected RED|
|3|PS7|0|53686|exact28 ordered PASS labels, no stderr|GREEN|
|4|PS5.1|0|111570|exact28 ordered PASS labels, no stderr|GREEN|
|5|PS7 ownership mutation|1|1085|exact expected ownership assertion; controller rejected label comparison|ExpectedResult=false, STOP|

5/10 slots consumed; slots6..10 frozen/unperformed. Review0/1, not dispatched;
its original all-results-match prerequisite is unmet. All five children used
DESKTOP-H4SVCLU\Lenovo, admin=false. No timeout/elevation or automatic retry.
PS5 first GREEN111570ms remained within120000ms, with no deadline extension.

## Slot5 readonly assessment and restoration

The one authorized temporary mutation replaced the result Actions snapshot
with Scenario.Actions. Mutant SHA256:
5705E855413C5D97070113CC3F5C4D7AA76D30337A63348DDD2D248B28B308DB.
Saved stdout is exactly context, PASS:RawCaptureApiRequired, HOST metadata,
ASSERTION FAILED:RawCaptureCopyOwnership; stderr empty, exit1. Thus the intended
test caught the ownership break, but the recorded controller verdict is false.

Controller source line31 assigns an if-expression returning a single-element
array to expectedLabels. PowerShell pipeline unrolling produces a scalar string;
line70 indexes that string by character, comparing the full actual label with R.
The count is one but orderedEqual becomes false. The28-label GREEN cases are
arrays and do not take that scalar comparison path. This is a controller defect,
not evidence the mutant passed. No test/diagnostic rerun was used to assess it.

The mutation was immediately restored with apply_patch; module hash again equals
4C0E6C10DAB31C63E550416260B0672EB0D721A8210B31CCABEA05648DCFB673,
and the test remains7010CBCF535A431D62AEA7C3392971E96F46769801DF2075651CA87B3B8D88D7.
All12 prior pins still match. No product/test/controller repair followed the stop.
Original index and receipts were not edited or promoted; index SHA256:
82B3593D3526EFCA688D0D4ABB0DD404E22D073DC320611FC4C96AC4CE129D39.

Proposed only, requiring explicit user approval: preserve the false original
verdict, allow the readonly assertion/label assessment as the specific slot5
admission exception, normalize future expected-label collections to arrays in
the controller, then use only slots6..10. Do not repeat slot5 or change source,
tests, assertions, stream policy or time limits. Record that continuation
authorization explicitly; allow the existing single reviewer only if all five
remaining calls meet their frozen expectations and pins still match. Without
this new direction, no correction, continuation or review is authorized.

## Execution rulings and limits

- PowerShell-native ledger/controller management replaces Bash-only skill scripts.
  Existing isolated worktree is reused. No extra baseline/full-suite/task-done
  invocation is permitted outside the approved ten slots; this is narrower evidence.
- Exact assertion failures are retained as stdout markers with exit1 by the test
  harness; unrelated exceptions remain engine errors, never accepted as expected RED.
- The approved plan overrides automatic repair/retest, whole-branch review and
  scratch deletion defaults. Preserve all receipts; any unexpected failure freezes
  remaining slots and conditional review. No fixes within a stopped packet.

This tests actual owned MemoryStream reads, not real pipes or scheduling.
Time is virtual, commands/UTC are declarations, stopping is intention only.
Created means internally consistent record, not collection/exit success;
OriginVerified/Acceptance/RuntimeApproved/LiveActionsPerformed stayfalse and
CleanupStatus staysunknown. New Actions snapshots are not stored by the old
archive interface. The old archive's approved limited acceptance does not
automatically accept this new module or close any Stage B/I2/I3/full C12 gate.

Ignored runner/frozen manifest/index/stdout/stderr/progress are retained locally
under `.superpowers/sdd/2026-10-17-offline-raw-capture/`. They are test-harness
text receipts, not a real raw-pipe collector or origin-authenticated capture.
Clone transfers portable source/tests and this report, not those raw receipts
or authority to reuse consumed/frozen slots. A new computer cannot validate raw
slot5 facts from this report alone or automatically resume this stopped packet.
