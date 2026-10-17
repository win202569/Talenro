# Supplemental synthetic-store coverage — four approved runs complete

Closing disposition: the separately approved [independent closing review](2026-10-17-synthetic-capture-closing-review.md)
is now complete. No confirmed supplemental/evidence defects; targeted Important2/3
are addressed within the approved cases. Historical references below to no second
review describe the earlier snapshot, not current review status. Overall
acceptance and other listed limitations remain open.

Subsequent work: [new dependency regressions](2026-10-17-synthetic-capture-dependency-regression.md)
completed all six separately approved runs. The next-work proposal below is
historical, not an unused allowance. Supplemental source remains unchanged
and has not received another independent review; all acceptance limits remain.

## Approved boundary

After fe77ddeb, the user approved a separate bounded coverage extension: add an
independent supplemental test file, leave product and original tests unchanged,
two groups on normal-user PS7/PS5.1, at most four new children,120000ms each,
stop on abnormal,no automatic fix/retry. Preserve all records, use the previously
approved progress-classification rule, commit/push necessary portable files.
This does not resume frozen slots or authorize dependency regressions or another
review. Both new scripts parsed before the first launch; expected label lists
and test hash were fixed before any child. Execution order below was preserved.

New portable test: [synthetic-capture-store.supplemental.tests.ps1](../diagnostics/isolated-fuzz/synthetic-capture-store.supplemental.tests.ps1).
Required explicit groups are event-decoder and writer-failure; no implicit all
entry and no dependency on ignored receipt files. No production functions/files
or original test bodies were edited. This is characterization of existing
behavior, not a product fix or a claimed newly observed RED-to-GREEN cycle.
No extra full-project test or mutation-test hosts were run outside the cap.

## Test design and observed coverage

### Event decoder —9 PASS per engine

Seven content assertions plus required/completion sentinels:

- Independently hand-encoded one-tick archive is accepted, preserving its event
  and accepted count1/incomplete lifecycle. Binding and marker hashes are built
  with test-local SHA256 over manually constructed bytes, not product encoders.
- Event kind0 and5 are rejected; EOF channel0 and3 are rejected.
- Tick elapsed time-1 is rejected.
- A33-chunk wire trace has individually valid65536-byte chunks but aggregate
 2162688 bytes, exceeding2097152; decoder rejects the cumulative bound.

Each negative fixture exercises the real private event decoder and asserts the
specific rejection branch, then the public reader on a complete file/marker
pair with independently recomputed hash. Public result must be invalid-payload,
Verified=false,Record=null and all authority flags false. Private checks prevent
later semantic failures from being mistaken for evidence of the targeted guard.
The positive control establishes that the hand-encoded event-bearing path works.
These fixtures are not a general independent populated-package wire oracle.

### Writer failure —6 PASS per engine

Four content assertions plus required/completion sentinels:

- After target directory creation, a record.bin collision causes real CreateNew
  failure; collision bytes remain, no complete marker appears, stopped byte1 is
  published and reader refuses the archive.
- A complete.bin collision after real payload writing preserves original payload
  bytes and collision bytes, publishes stopped and reader refuses.
- The same marker collision plus a stopped directory makes sentinel file creation
  itself fail. The original marker error still propagates; the secondary stopped
  failure is captured separately, artifacts remain and reader refuses.
- After local injection scopes unwind, the original helper is restored/visible;
  normal writer+reader succeed with the independent minimal payload bytes and
  no stopped sentinel. This checks that injection did not escape its scope.

Injection is a function-local forwarding wrapper around the original new-file
helper. It creates only the designated collision in a fresh GUID fixture beneath
the fixed synthetic root, then calls the real helper. It does not return fake
success, fabricate exceptions, mock the codec/contract/reader, alter permissions,
add a production test hook, rely on a concurrent race or terminate processes.
The caught primary message,FullyQualifiedErrorId and base HResult match the
captured original; in the double-failure case they are not replaced by the
stopped error. All created artifacts are retained, including deliberate conflicts.

## Execution evidence

|New slot|Normal-user engine/group|Child exit|Elapsed ms|PASS|Stderr chars|Progress objects|Result|
|---|---|---:|---:|---:|---:|---:|---|
|1|PS7.6.5 event-decoder|0|4140|9|0|0|complete|
|2|PS5.1.26100.9711 event-decoder|0|2602|9|584|2|complete|
|3|PS7.6.5 writer-failure|0|2175|6|0|0|complete|
|4|PS5.1.26100.9711 writer-failure|0|2871|6|584|2|complete|

Both engines x64; parent and child non-admin guards confirmed. No timeout or
retry. PS5 stderr contains two parseable progress objects per child, accepted
under the rule fixed before launch; no blanket stderr ignore or suppression.
Raw stdout must match exact ordered labels plus context/host/LIMIT, and unknown
output/errors/parsing failures stop the runner. All four ExpectedResult=true.
NEW4/4 allowance is spent. Readonly follow-up compared raw labels to index and
fixed inventory, stderr lengths and progress-only structure; hashes below match.

## Review finding disposition and remaining gates

The [independent review](2026-10-17-synthetic-capture-readonly-review.md) Important2
event-decoder and Important3 writer-failure gaps now have direct bounded dynamic
evidence for the cases above on both engines. Its historical findings are not
rewritten, and this new test file has NOT received a second independent review.
These results do not validate every possible writer fault or decode condition.

Important1 acceptance/regressions remains open. Original PS5 boundary timeout,
ordinary reader fixture failure and administrator PS5 strict-stderr receipt
retain their original failed dispositions. Missing-root PS5 evidence, file-level
reparse, full independent populated-package oracle/max-package trace and other
recorded limitations remain. All old frozen slots and conditional reviews stay
gated. No full-suite, main, live capture, Stage B/I2/I3/full C12, provenance,
power-loss or security/vendor clearance follows.

Next recommended work: agree a NEW bounded dependency-regression packet for
old store, simulation and contract on both engines, with exact inventories,
prior stream policy and fail-stop limits. Not executed or granted by this report.

## Recovery and integrity

Raw runner,index,stdout/stderr remain ignored/local under
`.superpowers/sdd/2026-10-17-synthetic-capture-supplemental`; fixtures remain in
the existing synthetic-store root. Git carries the new standalone supplemental
test and report/status/recovery notes, not local raw receipts or fixture state.
New-computer users can recover the test definitions and actual result/limits
from the development branch; cloning is not permission to rerun spent slots.

SHA256:

- Supplemental test:243B4673BF22614718D61FF92BB421F7EE09D993691D0803BB88EC8D11C4B87A.
- New index:33D8DFC4A53C94176B4FD0EBF60B970D4FAD3F500859A0692994257A9A6E0AD1.
- Product unchanged:508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C.
- Original test unchanged:406C7C661A181988C7F0AE90E577D1F948148C2B19BE5CD411F92D23E7FECBFF.
- Partition index unchanged:B61DD587834978007736566C0951A8C91FAA9B5FADE1760B903455F932A19AB1.
- Reader index unchanged:080455EE56A74498B09099061C53FC0CF46801C4ED614BEA37774D72629D8211.
- Format index unchanged:5FA944DF33467C084ED50260CAA2201C77EF4341F482CFBD3E12163D71D52600.
