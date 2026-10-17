# Synthetic archive test partitions — STOPPED,unfinished checkpoint

## Approved bounded scope

User approved the short grouping design after2cd4eed12dd6312857358fdcb561fab28f603cc9:
test-only selectors for four original contiguous boundary sections,NEW10 fresh
hosts in fixed order,120000ms each,no retry or repair reserve,stop on abnormal.
One NEW independent readonly coverage review only after all10 complete. This
does not resume original slots7..12 or the original implementation review.

Only synthetic-capture-store.tests.ps1 execution guards/selector validation and
partition completion labels changed. No product/module/dependency/test-body
change,mocking,validation deletion,cache optimization,timeout increase or full
project test run. This is an organization refactor,not a new product RED/GREEN
repair claim. Approved finite scope excludes automatic fixes or extra tests.

## Preservation and coverage map

|Selector|Original content checks|Required+completion|Expected PASS|
|---|---:|---:|---:|
|admission|26|2|28|
|byte-bound|1|2|3|
|event-bound|1|2|3|
|reader-paths|43|2|45|

The71 content checks remain exactly once in original order when groups are
concatenated. New79 PASS target per engine includes8 sentinels;the6 extra versus
old73 are repeated sentinels,NOT added content coverage. Admission retains all
preflight/clock/mutation/method/getter/shadow cases;byte-bound retains exact2MiB
trace/64KiB arguments;event-bound retains4096 ticks;reader-paths retains18
malformed payload,8 marker/partial,existing target,10 unsafe read/write,2
hardlink,junction,exclusive writer and2 cross-format checks.

All fixture/helpers,literal oracle,root preparation,callbacks,ownership and
result checks stay in place. Reader-paths initializes its own offsets/destination.
Legacy format/boundary/default all remain;boundary still selects the entire
original ordered sequence and its completion marker(expected73),all retains
format then boundary(expected90). These aggregate entries were NOT rerun.
New selectors emit only their partition-specific completion,not a false full
boundary completion. Format remains expected18 but its new runs did not occur.

Before any launch,removing only lines marked '# PARTITION' and restoring the
two selector header expressions reproduced every baseline line exactly,after
normalizing line endings for comparison. The test and runner parsed. An ordered
label inventory checked all71 original content labels against the earlier
successful PS7 receipt,including repeated labels. This is source/coverage
preservation evidence,not a behavioral test or a new full-suite pass. The saved
runner also checks exact ordered label lists,count,terminal,exit and stderr.

## New execution results —7/10 consumed,8..10 FROZEN

|Slot|Host/group|Exit|Elapsed ms|PASS|Child stderr chars|Disposition|
|---|---|---:|---:|---:|---:|---|
|1|PS7 admission|0|10843|28|0|complete|
|2|PS5 admission|0|2517|28|0|complete|
|3|PS7 byte-bound|0|18480|3|0|complete|
|4|PS5 byte-bound|0|4367|3|0|complete|
|5|PS7 event-bound|0|47442|3|0|complete|
|6|PS5 event-bound|0|80106|3|0|complete|
|7|PS7 reader-paths|1|2948|38 partial|328|unexpected hardlink fixture error;STOP|
|8|PS5 reader-paths|—|—|—|—|FROZEN,unstarted|
|9|PS7 format|—|—|—|—|FROZEN,unstarted|
|10|PS5 format|—|—|—|—|FROZEN,unstarted|

No run timed out. The original4096 case now has a completed PS5 public archive
roundtrip in a fresh independent group at80.106s. This does NOT retrospectively
pass the old same-process boundary run,prove a stable latency or establish that
grouping was the sole cause of its120s timeout. Original payload checks and both
public verifications remained on the path. Large timing differences between
hosts/groups are observations,not attributed to an engine/cache cause.

Slot7 completed required API+37 reader content checks through the five unsafe
read/write pairs. At test line140,the first New-Item -ItemType HardLink fixture
creation failed with an access-denial diagnostic. The first hardlink assertion,
second hardlink,junction,exclusive writer and both cross-format assertions did
not execute to completion;six content outcomes plus completion sentinel absent.
This is a fixture-operation failure,not an observed failed archive rejection
assertion. It is still a FAILED group,not38/45 accepted coverage or a skipped
hardlink check. The runner stopped;no retry,elevated test run,permission change,
fixture cleanup,skip or broader alternate execution occurred.

Localized child stderr was already damaged by process text decoding;original
saved text remains untouched. Operation/location/exit and missing outcomes are
clear;no reliable native error code/exception subtype is claimed. A readonly
root ACL/attribute check observed administrator ownership plus modify grants.
That alone does NOT establish whether ACLs,sandbox policy or another filesystem
condition caused the denial. Do not infer that taking ownership,admin execution
or an ACL rewrite is necessary/safe. No such action is authorized here.

## Review,next action and recovery

The all10-success prerequisite was not met,so the new independent readonly review
was NOT dispatched. No finding counts or review clearance exist. No fix/retest
reserve;new8..10 remain frozen. Source/tests were unchanged after first new
launch;postrun receipts/ordered labels/test hashes match the saved index. Any
further hardlink diagnosis or rerun needs a separately bounded proposal and new
authority,not silent continuation of these slots. Product acceptance remains open.

Keep this as an unfinished development checkpoint on the existing branch. Git
preserves the updated portable test/source mapping and this report/status/runbook.
Actual raw receipts,expected label files,runner and fixture artifacts remain
ignored/local under `.superpowers/sdd/2026-10-17-synthetic-capture-partitions`
and the existing archive fixture root. No files/links/old receipts removed.
Tracked test definitions depend on no ignored receipts;cloning does not transfer
fixture state,caches,raw sessions or authorization to rerun.

## Hashes and retained gates

Original test SHA256 before grouping:
`185FA2569A3364CD8DC4E894490794A5B6C6799F84B323F6BB6D66B94468335F`.
New test SHA256 at all7 launches and postrun:
`406C7C661A181988C7F0AE90E577D1F948148C2B19BE5CD411F92D23E7FECBFF`.
Unchanged product store:
`508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C`.
Contract/simulator/package/classifier hashes rechecked unchanged;no product diff.
New seven-entry index:
`B61DD587834978007736566C0951A8C91FAA9B5FADE1760B903455F932A19AB1`.
Original six-entry index unchanged:
`ADF5012858A53A20B371BD9D187054C5DB003D4F070EF743A9171E4AE338827A`.
Prior context diagnostic index unchanged:
`07C14EA6AF3A562881DDB54B8322CACCE2DE56C7CA0951755C2E46740E501DD9`.

Original archive6/12 spent,original7..12 frozen,dependent regression/original
review unstarted;all prior diagnostic allowances spent. NEW partition7/10 spent,
new8..10 frozen,new review unstarted. No root-cause/fix/GREEN/full acceptance,
main merge,real capture/security/vendor clearance,Stage B/I2/I3/full C12 or extra
permission follows from partial results.
