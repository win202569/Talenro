# Format-only verification — both engines complete under prior stream policy

## Approved scope and policy

After 79a8f7a6, the user approved NEW two normal-user children: PS7 then PS5.1,
format only, one run each, 120000ms cap, exact 18 ordered PASS labels. Stop on
missing results, errors, unknown output or parsing failure; no retry, product
edit, security change or old frozen-slot consumption. Preserve raw output.

Before either launch, the new ignored parent runner pinned product/test hashes,
checked expected username and non-administrator role in parent and child, and
defined the output policy. Stdout must contain only the 18 exact PASS labels,
one context line, matching host line, root-precondition line and exact LIMIT
line. Stderr must be empty or entirely parseable CLIXML whose direct objects
are explicitly progress records with recognizable PR Record/type. XML DTDs
and external resolution are disabled; no object deserialization/evaluation.
Any other stream, malformed serialization or unrecognized output fails the run.
No progress suppression or output deletion was applied.

The runner was parsed before execution. Children used NoProfile/NonInteractive
with encoded bootstrap only for UTF8 output, context guard and fixed format
invocation; the PS5 module path selected WindowsPowerShell modules. No elevation
or alternate context route was requested. Product and test bodies stayed fixed.

## Results

|New slot|Normal-user engine|Child exit|Elapsed ms|PASS|Stderr chars|Progress objects|Result|
|---|---|---:|---:|---:|---:|---:|---|
|1|7.6.5 Core x64|0|8323|18|0|0|complete|
|2|5.1.26100.9711 Desktop x64|0|2939|18|584|2|complete|

Neither timed out. Both parent/child non-admin guards passed, all 18 labels
matched the prior inventory in exact order, and host/context/LIMIT lines matched.
PS5 stderr consists of two progress objects, both Completed; no non-progress
object was present. Localized activity text is already damaged by decoding;
its exact wording or emitting module is not reconstructed. Both ExpectedResult
values are true under this packet's previously approved policy, not a rule
invented after observing the output. Both NEW2/2 slots are spent, no retries.

Coverage includes independent literal payload/marker comparison, twenty-field
roundtrip, four lifecycle states, nullable exit/index distinction, Int64 time,
late-event retention, UTF8/UTF16 Unicode and byte fidelity, independent ownership,
caller-mutation isolation and archive text never executed. The 18 labels include
API and completion sentinels, not 18 distinct content scenarios.

Both ROOT_PRECONDITION lines report existed=True. The missing-root branch was
not exercised; no existing root was deleted to fabricate that precondition.
These results do not establish live capture, origin trust, atomicity, power-loss
safety or full acceptance. This was format only, not all/boundary or regression.

## Integrity, gates and recovery

Readonly follow-up rechecked saved stdout/index labels, normal-user guards,
exit/timeouts, stderr length and structured progress classification. SHA256:

- New index:5FA944DF33467C084ED50260CAA2201C77EF4341F482CFBD3E12163D71D52600.
- Product unchanged:508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C.
- Test unchanged:406C7C661A181988C7F0AE90E577D1F948148C2B19BE5CD411F92D23E7FECBFF.
- Prior reader index unchanged:080455EE56A74498B09099061C53FC0CF46801C4ED614BEA37774D72629D8211.
- Old partition index unchanged:B61DD587834978007736566C0951A8C91FAA9B5FADE1760B903455F932A19AB1.
- Original index unchanged:ADF5012858A53A20B371BD9D187054C5DB003D4F070EF743A9171E4AE338827A.

Original reader packet's PS5 strict empty-stderr failure is not rewritten. Old
partition 8..10/original 7..12 remain frozen; this separate packet does not turn
their indexes into success or satisfy the old all10-success review prerequisite.
Dependent regressions and independent reviews remain pending. No main merge,
Stage B/I2/I3/full C12 acceptance or security/vendor clearance follows.

Next recommendation: consolidate the evidence/remaining coverage across the
separate packets before proposing a newly scoped independent review. Do not
automatically dispatch the old review or rerun old suites.

Raw receipts, runner and retained fixtures remain ignored/local under
`.superpowers/sdd/2026-10-17-synthetic-capture-format` and the existing fixture
root. Git preserves this result, limits, unchanged portable source/tests and
recovery notes on the development branch, not ignored raw state. Cloning does
not grant new execution authority. Overall implementation remains unaccepted.
