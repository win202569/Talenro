# Synthetic capture 4096-event stage timing — exploratory diagnosis

## Authority and scope

User approved the short probe after unfinished checkpoint
`1c71ac6987acf36ab8dae26144ee24531cd0ce1e`:NEW maximum2 hosts PS7 thenPS5,
120000ms each,no retry,stop on abnormal/timeout and retain partial timings.
No product/original-test change,no archive writes,no public file-store calls,
no failed full-suite replay or use of frozen original slots7..12.

The [exploratory diagnostic](../diagnostics/isolated-fuzz/synthetic-capture-timing-diagnostic.ps1)
uses one literal4096tick packet,all ElapsedMs=0,empty arguments and declared
s/x/o plus fixed UTC metadata. It times imports,fixture setup,public record
construction,public original verification,actual private memory codec and
public decoded verification. It neither wraps/replaces functions nor changes
validation. Memory payload includes actual TLRSYN01/version1 envelope and is
checked against independent size308+4096*9=37172bytes. No payload/command is run.

This diagnostic artifact is not a supported hostile-helper interface,product
feature,benchmark or acceptance suite. Private codecs are called only with a
trusted literal fixture. The script is preserved for reproducibility,not wired
into tests/runtime or authorization to rerun. Packet setup directly fills a
fixed array rather than replaying all old Fixture-Tick calls/previous cases.
No physical file-I/O/marker checks are timed;the original public writer/reader
have extra filesystem/hash work not present in this probe.

## Observations

PS7 completed exit0/33514ms/empty stderr,observed7.6.5Core x64.
PS5 observed5.1.26100.9711Desktop x64,reached120000ms outer cap;testhost
terminated,exit-1/120104ms/empty childstderr. It completed six stages and
started verify-decoded,but no end/DIAGNOSTIC COMPLETE was printed.2/2 consumed,
no retry/source change or original regression/review continuation.

|Stage|PS7 ms|PS5 ms|
|---|---|---|
|import|570.660|514.724|
|packet-build|32.346|71.036|
|construct-record|3704.356|7135.423|
|verify-original|14943.332|56114.443|
|encode-memory|5123.660|19444.190|
|decode-memory|313.560|1523.865|
|verify-decoded|8398.781|incomplete;started at childTotal84834.218ms|

Stage times include small timing/JSON-output overhead and are not rigorous
microbenchmarks. Outer elapsed includes launch/termination/bookkeeping too.
One sample per host,no warm-up/repeats,confidence intervals or stable latency
guarantee. Any timeout is an incomplete observation,never a passing suite.

verify-original onPS5 returned Verified=true before its end timestamp. The
decoded4096event count,complete memory consumption,binding equality and37172-byte
payload checks also completed. Final decoded public verification did NOT return;
do not claim its verdict or infer its exact duration. Parent total uses a
different start point from childTotal;subtracting84834.218 from120104 does not
produce a precisely measured final-phase duration. Parent termination overhead
does not extend the approved120000ms wait cap or prove broader OS cleanup.

## Interpretation — bottleneck narrowed,not a verified fix

The largest completed PS5 stage is the public record verifier(56.114s),not packet
construction(0.071s) or decoding(1.524s). A second call to that same verifier is
the last active stage at termination. This memory-only path also exhausts120s;
physical archive I/O is therefore not necessary for high latency. It does NOT
prove disk work has zero cost or exactly reproduce old whole-suite timings.
PS7's two verifier calls total23.342s;this one sample cannot prove stable host
ratios or attribute the PS5/PS7 difference to a specific engine/JIT/cache cause.

Static call tracing of Test-SyntheticCaptureRecord shows four relevant parts:
Read-ContractSnapshot/Read-ContractStoredEvents;Get-ContractPacket;
New-ContractValidatedRecord;Assert-ContractEquivalent and final UTC/order verdict.
The writer invokes the full verifier before publication;the public reader invokes
it after reconstruction. The probe preserved those two checks in separate stages,
but did not time the verifier's internal parts individually. Repeated full-record
work is observed by the call chain;which internal loop/type/property/compare
operation dominates remains unmeasured. Do not call this an exact root cause.

No checks were bypassed to measure these stages;no function wrapping/replacement,
limit increase,case deletion or diagnostic-generated acceptance token. The coarse
timings justify investigating the verifier's internal phases,not deleting one
verification,weakening passive/ordinal guards or blaming only the environment.
Next proposal is separately bounded internal-phase timing/readonly source audit;
only after identifying the hotspot should an unchanged-semantics optimization
design and finite RED/GREEN scope be proposed. No such host or repair is granted
by this report. Old timeout remains unresolved as an acceptance gate.

## Preservation and boundaries

Probe SHA256 at launch:
`A5516669EDEC2170F7A5423E78318A4C5929BF75BF10A31982BC9639C32812A4`.
Product store:
`508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C`.
Original tests:
`185FA2569A3364CD8DC4E894490794A5B6C6799F84B323F6BB6D66B94468335F`.
Unchanged contract verifier:
`116FCAAA6506ADC6969D718AC61F8C79908BDED48ADA1204C2E8A60F33A35BB2`.
Two-entry new probe index:
`DA2CEFA720F55A00423872E54CED6C3369F5506D6F5B6E7AFE706E835246716F`.
Original six-entry store index rechecked unchanged:
`ADF5012858A53A20B371BD9D187054C5DB003D4F070EF743A9171E4AE338827A`.
Ignored new probe ledger/index/stdout/stderr under
`.superpowers/sdd/2026-10-17-synthetic-capture-timing`;old12-slot ledger/index
untouched. All receipts retained,not committed as live/raw capture evidence.
Tracked diagnostic/report can recover the probe definition on another computer;
preparing hosts and authorizing any new launch are separate from cloning.

The [original timeout](2026-10-17-synthetic-capture-store.md) remains a failure,
6/12 spent,7..12 frozen;no acceptance or review continuation. No fix,new GREEN,
integration/full-project certificate,real capture/watchdog/cleanup,security/
vendor clearance,Stage B/I2/I3/full C12/main merge or extra permission follows.
