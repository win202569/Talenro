# Synthetic capture verifier internal timing — exploratory diagnosis

## Approved question and scope

After checkpoint9772976c,user approved a separately presented probe to identify
which internal verifier phases consume time. NEW maximum2 hosts,PS7 thenPS5,
120000ms each,no retry,stop on abnormal/timeout and preserve partial records.
This does not renew the consumed coarse probe or unfreeze original slots7..12.

The [diagnostic](../diagnostics/isolated-fuzz/synthetic-capture-internal-timing-diagnostic.ps1)
imports unchanged product helpers,builds one literal4096tick packet in memory,
constructs its record once,then follows the successful verifier helper sequence
once. Fields are s/x/o,empty arguments,ElapsedMs=0 and fixed UTC metadata.
It times header/schema/policy/false flags,snapshot/event reading,packet rebuilding
and Get-ContractPacket,New-ContractValidatedRecord,binding comparison,
Assert-ContractEquivalent,then UTC order and extra fixture facts.

No product function is replaced or wrapped. No extra public verifier call,
archive/public store I/O,codec roundtrip,old full-suite replay,product/test edit,
limit change or check deletion. Static parsing completed before either launch.
The internal path retains the checks but lets exceptions escape rather than
returning the public verifier verdict. This trusted-fixture exploratory artifact
is not a public API test,hostile-helper certification,benchmark or supported
runtime entry. It is retained only to explain/reproduce the probe definition;
cloning the repository is not permission to launch it again.

## Observed results

Both new slots completed normally,exit0,empty childstderr,ten ordered start/end
pairs and the terminal diagnostic message. PS7 was7.6.5Core x64,total18273ms;
PS5 was5.1.26100.9711Desktop x64,total37274ms.2/2 consumed,no retry or repair.

|Stage|PS7 ms|PS5 ms|
|---|---:|---:|
|import|734.139|527.510|
|packet-build|30.962|30.458|
|construct-record|3551.020|5991.950|
|record-header|15.766|29.632|
|read-snapshot-events|2882.286|5477.174|
|rebuild-packet|4645.068|10697.763|
|rebuild-expected-record|1735.317|4311.709|
|binding-compare|3.536|4.868|
|recursive-equivalence|4228.221|9888.525|
|utc-final-facts|0.484|0.604|

The seven internal-path stage durations sum to13510.678ms onPS7 and
30410.275ms onPS5,excluding import,fixture setup and initial construction.
Rebuild-packet plus recursive-equivalence account for65.68%/67.70% of those
respective sums. These are shares of measured stage durations,not CPU samples.
Timing/JSON overhead is included;parent total also includes launch/bookkeeping.
Each host had only one sample,no warmup/repeats or stable latency guarantee.

## What this narrows and what remains unknown

The two largest internal stages in both observations are packet rebuilding
and recursive equivalence. Source tracing places the former's per-event work
inside Get-ContractPacket:passive property inspection,exact shape/token/type
checks,simulator event validation and defensive copying. The latter recursively
checks every stored field/event with type/shape/string/integer checks. Both
therefore traverse4096 events;the probe has NOT measured individual primitives
or proved reflection,property enumeration,comparison,copying or function-call
overhead to be the specific root cause. Other measured phases are nonzero too.

The previous coarse PS5 public verification took56114.443ms;this internal-path
sum is not a speedup claim. No product change occurred. These are separate fresh
hosts and different call contexts,with no repeatability control;the internal
path neither executes the public catch/verdict wrapper nor verifies a decoded
record. The original writer/reader and full boundary suite do additional work.
Do not extrapolate this single-path completion to passing that suite or infer
that the old timeout vanished. No public verifier result was produced here.

Next recommended work is a readonly audit of repeated operations in those two
loops,then a narrowly scoped unchanged-semantics optimization proposal with
finite RED/GREEN and dependent coverage. Do not remove duplicate validations,
passive-descriptor/ordinal guards or defensive ownership merely for speed.
No optimization or additional execution is authorized by this report.

## Preservation and acceptance boundaries

Launch/post-run diagnostic SHA256:
`05E764444D641874F566BDEE99A90F6F6298F6339603C4C67DC010D056CD89E2`.
Product store unchanged:
`508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C`.
Original tests unchanged:
`185FA2569A3364CD8DC4E894490794A5B6C6799F84B323F6BB6D66B94468335F`.
Contract verifier unchanged:
`116FCAAA6506ADC6969D718AC61F8C79908BDED48ADA1204C2E8A60F33A35BB2`.
New two-entry index:
`80784F02FE395FE6A57BA4813332224172AAB514708245BB2D83B6716E4BF7F8`.
Original six-entry index unchanged:
`ADF5012858A53A20B371BD9D187054C5DB003D4F070EF743A9171E4AE338827A`.
Prior coarse two-entry index unchanged:
`DA2CEFA720F55A00423872E54CED6C3369F5506D6F5B6E7AFE706E835246716F`.

All new raw receipts/index/runner/progress retained locally under ignored
`.superpowers/sdd/2026-10-17-synthetic-capture-internal-timing`;previous records
retained untouched. Portable diagnostic/report contain the fixture,phase order,
scope and observations;raw local sessions/caches are not transferred by Git.

The [original archive timeout](2026-10-17-synthetic-capture-store.md) remains an
unresolved acceptance blocker:6/12 spent,7..12 frozen,dependent regression and
independent review not started. The [coarse diagnostic](2026-10-17-synthetic-capture-timing.md)
remains2/2 spent with a PS5 timeout. This new2/2 diagnostic completion is not
GREEN,product acceptance,review clearance or authority to merge main. No real
capture/watchdog/cleanup,security/vendor clearance,Stage B/I2/I3/full C12 or
extra permission follows. Save only an unfinished development checkpoint.
