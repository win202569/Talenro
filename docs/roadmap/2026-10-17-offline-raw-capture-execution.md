# Offline raw capture coordinator — execution record

Status: **FINITE VERIFICATION COMPLETE WITH EXPLICIT SLOT5 EXCEPTION; REVIEW PENDING, NOT ACCEPTED.** User selected Native option1 after
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
|6|PS7 stop mutation|1|1155|exact RawCaptureDeadlineEdges assertion, stderr0|expected mutation failure|
|7|PS7 final|0|38324|exact28 ordered PASS labels, stderr0|GREEN|
|8|PS5.1 final|0|101197|exact28 ordered PASS labels, stderr0|GREEN|
|9|PS7 contract regression|0|32813|unchanged suite, exact110 ordered PASS occurrences, stderr0|GREEN|
|10|PS5.1 contract regression|0|52254|same110 occurrences, stderr0|GREEN|

At the initial stop5/10 slots were consumed and6..10 frozen. The
[explicit continuation approval](2026-10-17-offline-raw-capture-continuation.md)
now permits only those remaining slots, with the original slot5 verdict retained.
All10/10 slots now consumed; no remaining test allowance. Review0/1 pending;
its prerequisite is now met under that explicit exception, not by changing
the old verdict. All ten children used
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
All12 prior pins still match. No product/test/controller repair preceded the explicit continuation approval.
Original index and receipts were not edited or promoted; index SHA256:
82B3593D3526EFCA688D0D4ABB0DD404E22D073DC320611FC4C96AC4CE129D39.

The user subsequently approved preserving the false original verdict, using the
readonly assessment as that specific admission exception, and correcting future
expected-label collections before slots6..10. No slot5 replay occurred. The
original index was copied verbatim before new rows were appended, and the
controller checks rows1..5 against it plus raw slot5 hashes before every new call.
Other failed rows still stop. Corrected controller hash:
C823F2893F511C3B2ABBA9521561F3B16379F76BA499AE44FD7F5447562B702F.

Slot6's sole temporary mutation replaced deadline-read-to-tick with an actual
read/chunk path; mutant67FC764E9C9946D838D528901B6F2D4BBF386464075D5BCBD23FC16898A51773
was caught by the exact frozen DeadlineEdges assertion. Source was immediately
restored to4C0E6C10, with no permanent product/test change. Conditional reviewer
may now proceed because slots7..10 met expectations and all source pins reconcile.

Final invocation index SHA256:
8EB29C6E95ADBF0D1491833303E970FF1A55C6619AC932514F065B6BE084F4D6.
Frozen source/inventory manifest SHA256:
B40871DB054BDFBAF61862C0CAF8F7C9CAAE8446FCA0913DCD9D10812A4B067B.
Parent independently checked all raw output hashes, ordered inventories, elapsed
limits and unchanged first five rows against the stopped snapshot. All12 prior
source/test pins and new module/test hashes still match. No permanent source/test
change or extra run followed the original implementation checkpoint.

Index TestSHA256 is the frozen new-suite pin on every row, not a claim that
slots9/10 execute that suite. Those two entries are the unchanged contract test,
SHA256=5A6975AAF9998D62304BA3DDB7483A6CE1A1382FBE9D05EBDA209BD7AE71566F,
bound by the frozen controller's Number>=9 target selection, manifest and exact
110-label inventories.28 labels include API/completion sentinels, not28 unique
requirements or a coverage percentage. No whole-project/full-suite run occurred.

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
