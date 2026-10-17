# Synthetic archive boundary workload and budget — readonly inventory

## Scope and evidence

Aftere99d0b8a,user requested readonly workload/budget accounting. No test,
diagnostic child,product/test edit,fixture creation or frozen-slot continuation
occurred. Sources:synthetic-capture-store.tests.ps1,store/contract code,the
original finite runner/six-entry index and saved slot6 stdout. Counts below are
source-path accounting,NOT newly measured dynamic call counts or timings.

The parent calls WaitForExit(120000) once for the entire child running Group
boundary. It does not reset for each Check-CaptureStore,fixture or API call.
Imports/Add-Type,common literal/hash setup,all preceding cases,large byte/event
fixtures,codec and filesystem work share that one wait budget. Launch/termination
and bookkeeping contribute to recorded parent elapsed120180ms;that is not a
raised120000ms wait. The contract's virtual event deadline120000ms is a separate
model bound,not the parent's wall-clock budget or a measurement of API latency.

## Ordered checks and observed stopping boundary

|Section|Expected PASS rows|Cumulative PASS|
|---|---:|---:|
|Required API sentinel|1|1|
|Preflight,clock,11 mutations,methods/getters,shadows|26|27|
|ExactFieldAndTraceBounds:2MiB trace/64KiB arguments|1|28|
|ExactEventCountBound:4096 ticks|1|29|
|Malformed primitive/derived/authority input|18|47|
|Marker/partial/stopped cases|8|55|
|Existing destination refusal|1|56|
|Five unsafe paths,read/write each|10|66|
|Two hardlink cases|2|68|
|Junction|1|69|
|Exclusive writer|1|70|
|Old/new cross-read cases|2|72|
|Boundary completion sentinel|1|73|

Thus the73 target includes71 content checks and2 sentinels,not73 independent
performance measurements. Slot6 saved28PASS,ending ExactFieldAndTraceBounds.
ExactEventCountBound and all subsequent checks lack completion evidence:45
target PASS rows absent,including the completion sentinel. stderr was empty,
exit-1,ExpectedResult=false. Index Completed=true means parent bookkeeping
completed,NOT a passing child. Original PS7 boundary73PASS/54429ms remains
historical scoped evidence,not proof of a PS5 timing bound.

## What a successful archive roundtrip actually does

For the4096 block at test lines110..111,the intended call sequence is:

1. Invoke Fixture-Tick4096 times and build the Object[] packet.
2. New-SyntheticCaptureRecord normalizes/validates it,simulates events,creates
   owned snapshots/stored events and computes binding. It does NOT call the
   public Test-SyntheticCaptureRecord verifier.
3. Fixture-Write calls the public writer. The writer validates/encodes every
   field into memory before calling Test-SyntheticCaptureRecord on the input.
4. Writer creates the archive,new-only writes payload,rereads and compares
   hashes,then publishes the completion marker.
5. Writer calls the public reader before returning. Reader checks path/handles,
   marker/payload integrity,decodes the full record,then calls the same public
   Test-SyntheticCaptureRecord on the reconstruction and rechecks stopped.
6. Fixture-Result checks the returned shape/reason/flags,then the test asserts
   the4096-event facts. Only then is ExactEventCountBound printed.

There are TWO full public verifier calls in this case's intended successful
path,not one. Before the missing PASS,the child could be in fixture construction,
encoding,first verifier,publication/readback or second verifier. Slot6 has no
subphase timestamps,so its exact interrupted operation cannot be inferred.
The prior one-verifier diagnostics omit much of this case and all earlier cases.
Do not multiply a later28.022s sample by two and claim an authoritative case
duration. Other PS5 observations varied(including56.114s original-verifier in
the coarse probe);no stable latency/remaining-case budget follows.

## Static API workload on the intended full boundary path

This inventory assumes fixture operations succeed and expected refusals occur
at their source-indicated guards. Environmental/earlier errors may stop execution
so these are not claims that slot6 executed every listed call.

|API/work|Whole boundary intended path|Prefix through last recorded PASS|
|---|---:|---:|
|New-SyntheticCaptureRecord calls|31|24|
|New-store public writer attempts|34|27|
|Successful writer roundtrips|5|4|
|New-store public reader calls,incl writer-final reads|41|4|
|Public synthetic verifier entries,intended refusal paths included|15|11|
|Old-store reader calls,separate API|1|0|

Constructor31 includes one explicit null-packet refusal descriptor;the rest are
fixture constructions,not31 certified records. Constructor arguments execute
even for later invalid/existing-path writer calls. Writer34 comprises five
positive roundtrips(three method fixtures,two exact-bound cases) and29 expected
refusals. Reader41 comprises36 explicit negative reader attempts plus five
writer-final reads. The old-format reader is not a synthetic-verifier call.

Verifier15 consists of10 positive entries(two per successful writer),three
writer-side expected refusals(clock reversal,derived mutation,binding mutation),
and two malformed archive reader paths(binding/derived). Most other malformed
inputs are rejected by local preflight/codec/path/marker checks before reaching
the verifier. The authority mutation is rejected by the codec,not by a full
verifier. Therefore neither writer34 nor reader41 can simply be multiplied by
one full-verifier sample. Prefix11 includes the two large-byte roundtrip verifier
calls,not just small empty-record checks. Large-byte setup also allocates32
65536-byte chunks,64KiB arguments,and admitted1MiB per channel.

## Conclusion and recommended next design

Confirmed:the budget is whole-group;the missing4096 case requires two full
verifiers plus substantial other work;it follows26 admission checks and a large
byte roundtrip. The group combines different workload scales. This makes
cumulative workload versus grouping a credible hypothesis to investigate.
NOT confirmed:an individual API exceeds120s,which exact operation caused the
old timeout,or that regrouping/any optimization would pass. No per-case times
exist in the failed run;successful later diagnostics cannot fill that gap.

Recommend designing a coverage-preserving split of admission,large-byte,
large-event and reader/filesystem checks before more micro-optimization. Any
such design must map every original71 content checks plus required sentinels,
retain all fixture/precondition/ownership and filesystem guards,retain a120s
per-child cap,and request a NEW finite allowance. This would be a change to
verification organization,NOT a proven product speedup or retroactive pass of
the original monolithic run. Original7..12 remain frozen;do not implement or
launch groups until that separate design/allowance is approved. No partition
or new runner was created this turn.

## Preservation and boundaries

### Pending bounded test-organization proposal

Following this audit,user asked to continue with design. Short design is being
presented in chat for approval;no test edits/new runner/hosts yet. Add explicit
Group choices admission,byte-bound,event-bound,reader-paths to the existing suite.
Guard the four original contiguous sections without changing fixture/assertion
bodies or their internal ordering. Retain format,boundary and default all;legacy
boundary still runs the complete original sequence with73PASS,default all retains
its combined sequence(expected90PASS). These legacy full runs are NOT proposed
for rerun and preservation of code is not new runtime verification of them.

Map original71 content checks exactly once across groups:admission26,byte-bound1,
event-bound1,reader-paths43. Each new group retains the required API sentinel and
gets its own distinctly named partition completion sentinel,NOT the full boundary
completion label. Expected PASS28/3/3/45 respectively,total79 per host;the extra6
versus73 are duplicated sentinels only,not extra coverage. Root/literal/hash setup
and fixture/result/passivity/ownership guards remain. Reader-paths owns its offset
and destination initialization;no dependence on another group's process state.

Propose NEW10 fresh hosts in order:admission7/5,byte-bound7/5,event-bound7/5,
reader-paths7/5,then existing format7/5(expected18 each). Every child120000ms,
no retries or repair reserve,stop on first abnormal/timeout and retain all records.
Before launch,static source/coverage comparison must establish all original test
bodies/ordered labels preserved and product hashes unchanged. This is a test-only
organization refactor,not a new product RED/GREEN repair claim. One NEW independent
readonly test-organization/coverage review only after all10 complete;findings stop
handoff without an automatic fix/retest reserve. Save source/mapping/evidence and
recovery docs to the existing development branch,not main.

This proposal needs user approval before changing the previously frozen test
source. Original six-entry index/raw outputs and frozen7..12 remain untouched.
No old dependent regressions or original implementation review are resumed.
Separate-process results do not prove original same-process state interactions,
clear the monolithic timeout or establish a product speedup. Even all proposed
groups passing would be scoped coverage evidence only,not full acceptance.

Store/original tests unchanged SHA256:
`508F853F23EB84CB2CFC54D57D94BEE6635000E7D03D3A97994660722173BA3C` /
`185FA2569A3364CD8DC4E894490794A5B6C6799F84B323F6BB6D66B94468335F`.
Original six-entry index unchanged:
`ADF5012858A53A20B371BD9D187054C5DB003D4F070EF743A9171E4AE338827A`.
Latest four-entry diagnostic index unchanged:
`07C14EA6AF3A562881DDB54B8322CACCE2DE56C7CA0951755C2E46740E501DD9`.
All receipts retained. Original6/12 spent,7..12 frozen,review unstarted;coarse2/2
with PS5 timeout,internal2/2,lookup2/2,names2/2,context4/4 spent. No new host,
root-cause/fix/GREEN/acceptance claim,main merge,real capture/security/vendor
clearance,Stage B/I2/I3/full C12 or expanded permission. Documents-only checkpoint.
