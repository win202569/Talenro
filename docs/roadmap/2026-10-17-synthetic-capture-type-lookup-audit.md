# Synthetic verifier repeated type lookup — readonly audit and pending probe

## Evidence and scope

After internal-timing checkpoint0d01d6f6,user requested continued readonly
loop audit and an optimization proposal. No product/test changes or new hosts
were executed in this audit. The [measured phases](2026-10-17-synthetic-capture-internal-timing.md)
remain evidence of aggregate phase cost,not individual operation cost.

Get-OfflineCaptureValueType in output-capture-simulation.ps1 resolves
`[object].GetMethod('GetType')` on every nonnull call,then invokes that method
on Value.PSObject.BaseObject. This deliberately avoids caller ETS GetType
dispatch. The invocation and per-input type check must remain;only lookup of
the same fixed CLR method is a candidate for reuse.

For each tick in Get-ContractPacket,the safe type helper is called by the direct
event query,Assert-ContractString Kind,Assert-ContractShape event and
Assert-OfflineCaptureEvent:four lookups per event,16384 across4096 ticks.
Assert-ContractEquivalent on each stored tick calls it for expected event,
actual event shape,expected Kind,actual Kind string and expected ElapsedMs:
five per event,20480 across4096 ticks. Together these event paths alone imply
36864 repeated resolutions of the fixed method. This is a static call count,
not runtime profiling;metadata/other phases add calls not counted here.

The recursive object branch also creates a ForEach-Object pipeline to collect
expected field names per object. Get-ContractPacket repeatedly inspects passive
properties and validates/copies events. Those are further candidates for later
investigation,not permission to remove checks or combine multiple changes now.
No source audit can establish that method lookup dominates the measured time.

## Proposed next step — diagnostic spike,NOT approved or executed

Question:does reusing only the fixed Object.GetType MethodInfo materially reduce
the safe type helper's repeated-call cost while returning identical base types
and never dispatching forged caller GetType methods?

Propose NEW maximum2 fresh hosts,PS7 thenPS5,120000ms each,no retry,stop on any
unexpected result/timeout and preserve partial output. This proposal is separate
from all consumed/frozen allowances and needs explicit user approval.

Use two differently named exploratory functions with otherwise identical helper
bodies/call structure:baseline resolves the fixed method each call;candidate
uses one locally prepared fixed MethodInfo but invokes it on every input.
Do not redefine/wrap product functions or edit product/original tests. Use only
trusted in-memory literals and a forged GetType ScriptMethod with a zero-callback
assertion;check null and actual fixture base types without accepting input as a
valid record. Time equal36864-call groups for baseline and candidate,baseline
first onPS7,candidate first onPS5. Keep fixture/setup outside group timers and
report both orders. One sample per host is not a stable benchmark or end-to-end
latency guarantee. No archive/store API,public full verification or suite replay.

No input object,type result,validation verdict or caller field is cached. This
probe reuses metadata only;all Invoke operations remain. It cannot certify
passive-property admission or all public security boundaries. Any real cache
initialization/lifetime/tamper assumptions and shared-consumer regression scope
must be addressed in a later product design,not silently introduced here.

If no useful effect is observed,do not implement this candidate merely because
the call count is high. If useful,it supports a separate narrowly bounded product
design and finite semantic/regression verification;it does not prove resolution
of the full-suite timeout. Pipeline extraction remains unchanged in this probe.

## Recovery and boundaries

No new code or host result exists. Resume from this proposal plus prior timing
report,not by automatically launching a new diagnostic. Original archive6/12
spent,7..12 frozen,dependent regression/review unstarted;coarse timing2/2 spent
with PS5 timeout;internal timing2/2 spent with complete manual-path observations.
All existing product/test/contract hashes remain at the previous checkpoint.
No repair,GREEN,acceptance,main merge,real capture/security/vendor clearance or
Stage B/I2/I3/full C12 authority follows. Save documents-only development progress.
