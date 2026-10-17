# Administrator reader-paths verification — stopped on receipt gate

## Approved scope

After ee0c4cb2,the user approved NEW two sequential administrator-context
children:PowerShell7 then5.1,reader-paths only,120000ms each,45 ordered PASS
labels required. Stop on any abnormal result,no retry,product changes,ACL or
security changes. These are separate from old frozen slots. No full-suite run,
independent review or acceptance expansion was authorized.

An ignored parent runner pins the original test/product hashes,checks expected
user and administrator role,and checks the same identity/role in each child
before invoking the unchanged test. The parent was parsed before execution.
Each child uses NoProfile/NonInteractive and an encoded bootstrap solely for
UTF8 output,context confirmation and the fixed reader-paths invocation. The
PS5 child receives its WindowsPowerShell module path. No RunAs or settings edits.

## Observed results

|New slot|Host|Child exit|Elapsed ms|PASS|Stderr chars|Receipt disposition|
|---|---|---:|---:|---:|---:|---|
|1|7.6.5 Core x64|0|2372|45|0|ExpectedResult=true|
|2|5.1.26100.9711 Desktop x64|0|2083|45|584|ExpectedResult=false;STOP|

Both parent and child administrator guards passed. Neither child timed out.
Both output complete,identical ordered45-label sequences:43 content assertions
plus required API and partition completion sentinels. Both include the LIMIT
marker. This includes both hardlink rejection assertions,junction rejection,
exclusive-writer rejection,and both cross-format checks that the old normal
slot7 did not reach. No test-body or product-code change was needed for these
observed assertions. This does not retrospectively pass the old failed slot.

PS5 emitted584 decoded characters on stderr:CLIXML with two Obj records whose
stream attributes are both progress. Their progress records show Completed;
the localized activity text is damaged by decoding and is not reconstructed.
Readonly XML inspection found no non-progress Obj records in that saved text.
The child exited0;the parent/tool exited1 because its strict empty-stderr gate
failed. Do not conflate the two exit codes or describe this as an observed
archive assertion failure. Nevertheless,the packet is NOT accepted under its
original receipt gate. Both2/2 slots are spent. No suppression,parser change,
receipt rewrite,post-hoc ExpectedResult promotion or rerun occurred.

## Next action and boundaries

Recommended next step:approve a readonly disposition of the saved PS5 CLIXML
receipt,with a separate derived assessment distinguishing progress from actual
errors while preserving the original failed gate/index. No new test child is
needed for that assessment. This recommendation is not permission to change
the old gate or resume tests. Future runners should explicitly specify stream
policy before launch rather than retroactively claim empty stderr.

Old partition7/10 remains spent,8..10 frozen;original6/12 spent,7..12 frozen.
Format verification and dependent regressions remain unrun in the new scope.
Both old reviews remain unstarted;the old all10-success review prerequisite
has not been met by these separate runs. No main merge,module/full C12
acceptance,live capture,Stage B/I2/I3 or security/vendor clearance follows.

## Recovery and integrity

Ignored local runner,index,stdout/stderr and fixtures are retained without
cleanup under `.superpowers/sdd/2026-10-17-synthetic-capture-reader-admin`
and the existing synthetic fixture root. Git transfers this portable report
and unchanged tests/product,not raw sessions or fixture state. Cloning grants
no execution permission. This is an unfinished development checkpoint only.

Readonly rechecks matched raw receipts to the two index records,exit codes,
timeout flags,stderr lengths and all ordered labels. SHA256 values:

- New index:080455EE56A74498B09099061C53FC0CF46801C4ED614BEA37774D72629D8211.
- Product unchanged:508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C.
- Tests unchanged:406C7C661A181988C7F0AE90E577D1F948148C2B19BE5CD411F92D23E7FECBFF.
- Old partition index unchanged:B61DD587834978007736566C0951A8C91FAA9B5FADE1760B903455F932A19AB1.
- Original index unchanged:ADF5012858A53A20B371BD9D187054C5DB003D4F070EF743A9171E4AE338827A.
