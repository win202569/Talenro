# Candidate outer boundary markers — PREPARED, NOT RUN

After56a7df4d,user approved preparation ONLY of the bounded outer-marker design.
ZERO target launches,warmups,retries,old-slot use,PS5 unfreezing,product changes,
reviewers or main merges. No autorunning controller was prepared.

## Portable artifacts

Folder: docs/diagnostics/isolated-fuzz/integer-candidate-boundary-markers/.

Seven source files copied from integer-type-reuse-full-regression. Six,including
the candidate module,are byte-identical. Only offline-supervision-policy.tests.ps1
adds eight standalone Console.Out.WriteLine calls inside the existing
ArrayAndIntegerBounds case body:START/END for InvalidInputs,Build4096,Check4096,
Check4097. Each samples Stopwatch.GetTimestamp as an invariant decimal integer.
No marker helper,product reload,product/internal-loop instrumentation,new
branches in common helpers,changed inputs or changed assertions.

P1-Case/P1-Check/P1-Equal and ALL other original test helper definitions are
byte-identical. The original statement blocks and case order remain in place.
Console output avoids contributing marker values to the PowerShell success
pipeline;actual runtime behavior and output ordering are NOT verified here.
Existing source headers' isolated/unverified warnings are retained.

README.md labels preparation/no replay authority and explains evidence limits.
expected-output.json records the complete91-line inventory:83 original ordered
nonempty records plus eight boundary records inserted at positions66–73
(one-based),after PassiveDescriptors and before ArrayAndIntegerBounds PASS.
64 PASS markers are required for a future complete result;NONE were newly run.

## Static verification performed

- Parsed all seven PowerShell source ASTs without dot-sourcing/invoking targets.
- Checked exact hash equality for six source files.
- Removed only eight matching marker lines in memory;remaining test text
  equals the entire prior test byte-for-byte (not just assertion substrings).
- Compared every original test function definition;none added/removed/changed.
- Independently checked eight source marker names/edges/order against JSON.
- Compared all83 non-marker expected entries with the prior frozen inventory.
- Rechecked327 existing pins and unchanged prior full-regression invocation index.
  No source/controller/raw record from previous measurements changed.

These are static preparation checks,NOT PS7/PS5 behavior,marker parsing runtime
tests,full regression PASS,root-cause confirmation,performance or acceptance.

| Identity | SHA256 |
|---|---|
| Original product, unchanged | 953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA |
| Candidate module/copy | 3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3 |
| New marked tests.ps1 | EB4BA718BC960061F948FEB184DE18676110B264E34B1600BC01EADD11F87018 |
| expected-output.json | 6FF2D34D86C43ABFCDA28A6502ED7735D539C113665B902FC936C202E4F8FCF3 |
| Unchanged copied top entry | 12D6F5460F97B9E09453B0387C30E3D20CE19B2EDF1CDF4DBF4E77DA31561422 |

## Future interpretation and next recommendation

The [read-only comparison](2026-10-18-p1-candidate-boundary-readonly-comparison.md)
defines H-boundary:Check4096 is longest among completed boundary regions,or
left open after earlier regions closed. This remains an UNTESTED localization
hypothesis. Check4096 includes scenario/expected arguments,API and equality;
it cannot identify which operation is slow. No product/helper cause is asserted.

A future capture must pin source/engine/output/controller BEFORE launch,record
Stopwatch.Frequency outside the91-line test stream,parse tick values as
nonnegative invariant Int64 and enforce globally nondecreasing marker times.
Calculate milliseconds only for complete matching START/END pairs. No inferred
tail,deadline-subtraction timings,missing-end repair or automatic retry.
A timeout before markers cannot support the4096 hypothesis;an open earlier
region must retain its actual frontier. A complete fast diagnostic does not
reproduce the prior failure or accept the uninstrumented entry. Marker/body-size/
output/runtime/JIT perturbation is real and uncalibrated.

Next recommend a separately approved NEW ONE-host ordinary fresh PS7 diagnostic,
30000ms cap,no warmup/retry,frozen source/output,unchanged safety checks,
bounded owned-process handling,raw preservation and stop on failure/timeout.
There is no PS5 fallback and no use of its frozen slot. An execution controller
and that run still require the execution-stage scope;neither was created here.
No direct invocation to bypass the prior stop rule.

Existing full-regression allowance remains1/2 spent with PS5 FROZEN;all older
spent/frozen allowances and review0/1 unused/ineligible persist. Candidate/I1/P1/
main/C12/security/runtime/release remain unaccepted. Clone the development
branch to recover nine portable preparation files plus this report/status/
runbook,not ignored receipts/controllers/runtime caches,credentials or execution
authority. Historical PASS/timeout verdicts remain unchanged.
