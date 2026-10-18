# Isolated integer candidate complete regression — PS7 TIMEOUT, PS5 FROZEN

After 1b1144a7, user approved the recommended NEW two-host allowance:
ordinary fresh PS7 then PS5 once each, 30000ms each, stop/freeze later host
on any failure/timeout. No retry/warmup, original replacement, reviewer,
main merge or old-budget reuse. One of two slots spent; remaining PS5 FROZEN.

## Source and observed result

Seven files in `docs/diagnostics/isolated-fuzz/integer-type-reuse-full-regression/`:
candidate module byte-identical to isolated candidate3537E9F5, plus six
byte-identical original entry/wrapper/test scripts. No test assertion/order,
relative loading path or product-internal diagnostic change. Original shape6,
reflection3, fixture8, original35, ownership10 sequence requires83 ordered
nonempty lines/64 PASS. Full regression copies do not include the separate
14-group direct integer suite, already measured in the preceding checkpoint.

PS7 7.6.5 Core x64 started2026-10-18T09:14:22.3775366Z:
wall30090ms, CPU36171.875ms, timedOut=true, exit-1 after owned-process termination.
Parent and child admin=False. Empty stderr; owned-host-exited; streams saved.
65 ordered lines/48 PASS, last `PASS: P1PassiveDescriptors`.
CaptureValid=true (exact prefix), ExpectedResult=false, CandidateOnly=true,
FullRegressionCopy=true, FormalProjectAcceptance=false.

PS5 was NOT launched. No retry or warmup. Controller has one final receipt and
no PS5 stream files. The original module was not replaced. All327 frozen pins,
controller/manifest, engines, saved stream hashes and ordered prefix were
independently rechecked read-only after execution. No saved assertion/fixture
error is present, but timeout is a failed complete-regression verdict.

## What this does and does not establish

The next expected output is `PASS: P1ArrayAndIntegerBounds`. That source group
contains invalid integer/array cases, a decorated-array case, construction and
evaluation of4096 observations, and4097 rejection. This uninstrumented capture
does NOT locate which statement was active at timeout or prove that4096
evaluation itself consumed the time. No stack/profile evidence was captured.

Earlier original PS7 completed the full entry while original PS5 timed out.
This candidate PS7 timed out at the same output frontier as prior original PS5.
Different runs/hosts/history are not a matched performance experiment.
Do not infer a stable regression, root cause, speedup or reduced dynamic query
count. Integer-local type reuse has NOT demonstrated a complete-flow solution.
Its prior direct-semantic PASS remains valid only for those direct vectors.

## Immutable identities

| Artifact | SHA256 |
|---|---|
| Original product, unchanged | 953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA |
| Candidate module/copy | 3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3 |
| Copied top entry, unchanged bytes | 12D6F5460F97B9E09453B0387C30E3D20CE19B2EDF1CDF4DBF4E77DA31561422 |
| frozen.json | E42F5CB53E683D6EFA4FBFFE42050D0372740D1F9C12A7C0A51C6869A708D0C3 |
| run.ps1 | E3738929978A48B3A4C7294E5783AA47F320CF00CB4BB60533CDE70CA333D520 |
| invocations.json | 581472A810A6351BCC78C4E85C031C513F581FC94460119CEBBAB196E1496DC9 |
| PS7 stdout | 0349D1AE13D3BE5CCD7A3D93514FA35722B1E5CA41A8325E86915BD9A78C0B7B |
| PS7 empty stderr | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

327 pins inherit312, add seven prior semantic packet anchors, seven new source
copies and the preceding semantic report. Local ignored packet:
`.superpowers/sdd/2026-10-18-p1-integer-candidate-full-regression/`.
Raw/controller/session files are preserved locally, not committed or required
to clone portable source and this evidence summary.

## Gates and next recommendation

UNACCEPTED development checkpoint. Current allowance1/2 spent, PS5 FROZEN;
all prior spent/frozen budgets and review0/1 unused/ineligible persist.
No main/C12/I1/P1/security/runtime/release acceptance or original timeout fix.
No independent reviewer dispatched.

Next recommend read-only comparison of the boundary group's call paths and
existing successful/timeout records to define one falsifiable diagnostic
hypothesis. Do not change product/assertions, repeat a full run, increase timeout
or release frozen PS5 without a separately agreed scope. Preserve this result,
not another speculative optimization. Fresh-computer recovery uses the current
development branch, not main; cloning does not restore ignored records,
runtime caches, credentials or authority to replay tests.
