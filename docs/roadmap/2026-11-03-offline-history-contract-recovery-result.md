# History-contract execution progress

Status: IN_PROGRESS / UNFINISHED / NOT_ACCEPTED.

## Latest: RED22 exact; four-file candidate ready for GREEN23

RED22 ran once after verified remote89bb7824:0PASS/1FAIL/19NotRun/exit1,
exact RuntimeException message/ErrorId OD: invalid explicit epoch. Setup and
recording errors null; pre/post source/history hashes unchanged. Actual call saved.
Four production binding files now implement the new route and parse. Exactly4of31
source identities changed;2341historical identities remain unchanged. Candidate
input is fixed. New attempts are exclusively acquired before resource checking;
shared read-only validation binds old receipts to old source epochs.
GREEN23 has not run; no new Prepare or host. Counters22/4/3/186.
Next: push candidate archive, verify remote, invoke GREEN23 once. First unexpected
result stops the remaining scope without edits/retries. Earlier sections are history.

## Latest: complete Task 1 fixtures prepared, RED22 not yet invoked

All20regression groups and the preparation helper are authored. Four new scripts
parse. Mechanical preparation succeeded:31sources,2341historical files,
2307original history,43append paths,2319expected pins. Fixture-map presence was
checked before writing. Production remains unchanged.
InputSHA256:3F7B73785FB601B2E988E14796C43D2085DC7B741E5E6B18A03D6049586D41D8.
TestSHA256:C9F2B924E84B000B4894D492DDA023C9EDA5825CA9273079F73D3951D407A5A9.
The preparation archive supersedes the two-file helper draft for recovery.
Next: after remote checkpoint verification, invoke RED22 exactly once.
The older paragraphs below describe the preceding draft checkpoint.

The user approved execution of plan commit
e2fcf01a13540a6f4b6658006bedf4d637359bb4 with “是”.
The original spec and plan bytes retain their historical draft labels; this
record supersedes only their pending-approval status, not their constraints.

Read-only baseline check: all 31 current candidate sources and 2326 historical
files match the path-v3 saved inventories. The new history-contract P namespace
was absent. No production source has been changed.

Authored Q/run-history-contract.ps1 and Q/preserve-history-contract.ps1.
Both parse with zero errors. These are unfinished helper drafts, not a frozen
test fixture or an executed preparation. The runner separates auxiliary slots
from the host slot and uses computations 22/23 with ceiling 23. The archive
helper adds the distinct freeze stage. Neither script has been invoked.

Portable helper bytes are in
../diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/portable-history-contract-helper-draft.json.
Restore only its two explicitly listed paths after verifying lengths and SHA256;
do not overwrite differing existing files. Prior complete history remains in
portable-history-path-v3-final.json. This small draft archive does not replace it.

Remaining under the existing approval: author all 20 test groups and the
preparation helper; mechanically check definitions and fixture keys; freeze the
complete inputs and preparation archive; checkpoint before RED22. Only after
exact expected RED may the four production files change. GREEN23, one Prepare
and one host must follow the approved sequence, with first unexpected execution
failure stopping the scope. No additional approval is needed merely to resume
this already approved plan.

No new auxiliary, Prepare, snapshot or host was executed. Counters remain
21 auxiliary / 4 hosts / 3 compilations / 186 snapshots. All 12 old stops remain.
The pending 2 auxiliaries and reactivated 1 host/31 snapshots are unconsumed.
No acceptance, main merge, or downstream execution is claimed.
