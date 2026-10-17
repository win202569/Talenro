# Pure-memory capture lifecycle simulation — proposed design

Status: **WRITTEN DESIGN AWAITING REVIEW**. User approved writing this design
after834ace2f,not implementation,tests or runtime entry. Architecture path:
written-spec approval precedes an implementation plan and execution-method
selection. No new validation allowance is granted here.

## Intent and success criteria

Prepare a deterministic model of separate stdout/stderr byte accumulation and
completion/timeout decisions before considering a real collector. A reviewer
must be able to tell complete capture from incomplete,late or invalid sequences,
while recovering every admitted byte without decoding/transcoding it. Existing
classification,package and persistence helpers remain unchanged.

This models caller-supplied events,not actual collection or a watchdog. No
process start/handle/PID access,stream I/O,OS clock,sleep,thread/task,timer,file
I/O,network,security query,nonce generation,or real probe identity. It cannot
establish command identity,UTC provenance,cleanup or security clearance.

## Alternatives and selected boundary

1. Selected:pure event-array evaluator with virtual elapsed milliseconds.
   Deterministic ordering/bounds,minimum OS exposure;does not test concurrent
   pipe draining or scheduling.
2. Background-thread simulated streams:closer to asynchronous reads,but adds
   nondeterminism and wall-clock timing;deferred to a separately designed adapter.
3. Real process collector:can observe actual pipes,but requires capture/watchdog
   integration and separate prerequisites/authority;not part of this design.

Only a new `output-capture-simulation.ps1` and corresponding `.tests.ps1` under
`docs/diagnostics/isolated-fuzz` are contemplated. Import may load the existing
pure evidence-package/classifier helpers,not isolation/store/native helpers.
No existing runtime interface or evidence schema is modified.

## Proposed interface and typed input

`Get-OfflineCaptureSimulation -Events <object> -EncodingName <object>
-OriginDeclaration <object>` returns one result. Inputs are not coerced into
valid values. Ordinary invalid input returns a refused result,not a thrown
validation exception. Resource exhaustion/runtime failures are not guaranteed
recoverable or equivalent to a valid result.

Events must be an actual one-dimensional System.Object[] with0..4096 elements;
no lazy enumeration,callbacks,streams or caller executable properties. The array
and contained values must remain unchanged during evaluation. Each event is a
plain PSCustomObject with exactly these case-sensitive NoteProperties:

| Kind literal | Exact properties besides Kind | Meaning |
|---|---|---|
|chunk|ElapsedMs,Channel,Bytes|Channel exactly stdout/stderr;Bytes actual byte[],length0..65536|
|eof|ElapsedMs,Channel|End of the named channel|
|exit|ElapsedMs,ExitCode|One observed simulated exit;integer within Int32 range|
|tick|ElapsedMs|Advance virtual observation time without byte/lifecycle changes|

No unknown fields/kinds/channels,duplicate EOF/exit,null byte arrays or alternate
property types. ElapsedMs must be Int32/Int64 in0..Int64.MaxValue;bool,float,
numeric strings and automatic conversions refused. Times must be nondecreasing;
equal-time events use supplied array order,never sorting/reordering. ExitCode
uses the same strict integer rule as existing packages,including parenthesized
numeric negative fixtures. stdout/stderr have independent accumulated caps of
1048576 bytes. Reject an entire exceeding chunk before appending;never clip it
and then claim complete capture. Maximum event/chunk counts bound parsing and
admitted work,not the caller's prior allocations or overall CLR memory usage.

EncodingName exactly utf-8/utf-16le,OriginDeclaration string1..1024 characters
with strict UTF8 representation at most4096 bytes. This compatible metadata is
an unverified caller declaration,not automatic encoding detection or identity.

## State machine and deadline

Initial state:collecting,both EOF flags false,no exit,empty streams,virtual
time0. Fixed deadline120000 milliseconds;not configurable and never measured
against a real clock. This simulation cannot enforce a wall-clock deadline.

Process in array order:

1. Validate outer array and metadata before admitting events. Failure returns
   refused with empty admitted streams. More than4096 events refuses the whole
   input before enumeration,not a truncated prefix run.
2. If an earlier event completed collection,any additional event (even tick)
   refuses as trailing-after-complete. Do not ignore trailing input or report
   success for only a prefix.
3. Validate the current event's exact shape,kind-specific value types/bounds,
   and nondecreasing valid time. Malformed/unknown/backward events refuse.
4. A valid event at or after120000 yields timed-out BEFORE applying its bytes,
   EOF or exit. Boundary equality is timeout. All previously admitted bytes and
   lifecycle facts remain;this event and later ones are not admitted.
5. For chunk,refuse if its channel already EOF or its total would exceed1MiB;
   otherwise append an owned copy of its complete byte array. Exit may precede
   remaining chunks/EOF because simulated buffered output can arrive afterward.
   For EOF/exit,refuse duplicate observations;otherwise record that fact. Tick
   changes only the observed time. No synthetic EOF/exit is invented.
6. Mark complete only after BOTH channel EOFs and exactly one exit,all before
   the deadline. Completion says lifecycle observations are present,not exit0
   or successful output. If the supplied array ends earlier,return incomplete.

First refusal/timeout stops interpretation immediately. Keep a reason and the
admitted prefix;do not retry,inspect later events to improve the result,or allow
a subsequent valid event to overwrite the failure. Invalid event data is not
part of admitted bytes;the caller retains the original input. This deliberately
does not promise archival of rejected/oversized bytes.

Examples (all channels initially empty):

| Supplied observations | Outcome |
|---|---|
|exit0 at5,eof stdout at6,eof stderr at7|complete;empty output classification|
|both EOF at1,no exit|incomplete;no manufactured exit code|
|exit0 at1,eof stdout at2,eof stderr at120000|timed-out;last EOF not admitted|
|exit-1 and both EOF at3|complete lifecycle;package classification refused|
|complete at3,then tick at4|refused;complete package not published|
|chunk after EOF,duplicate exit,or backward time|refused;admitted prefix retained|

## Result and existing evidence boundary

Result fields:Schema=1,Policy=offline-capture-simulation-v1,Status
(complete/incomplete/timed-out/refused),Reason,AcceptedEventCount,FailureEventIndex,
LastObservedElapsedMs,StdoutEOF,StderrEOF,ExitObserved,ExitCode,Stdout,Stderr,
Package,OriginVerified=false,Acceptance=false,RuntimeApproved=false.

AcceptedEventCount counts fully applied events,including ticks/EOF/exit;an
offending/timed-out event is not accepted. FailureEventIndex is zero-based for
event refusal/timeout,otherwise null (including header refusal/incomplete).
LastObservedElapsedMs is the latest structurally valid nondecreasing time
observed,including a timeout or valid-shaped trailing event;otherwise retain
the previous time (initial0). Reading a trailing event is limited to validating
its time for this diagnostic field;trailing-after-complete remains the reason
and the event is not applied. Do not inspect any further event after that.
ExitCode is null unless a valid exit was admitted;never use zero as unknown.

Stdout/Stderr use existing immutable Base64/ByteLength/SHA256 stream snapshots,
including empty admitted streams on early refusal. Post-return caller mutation
cannot change them. Package is null except when final Status=complete. For
complete,construct the existing package from admitted bytes and strict metadata,
recompute classification and verify it;an internal construction inconsistency
refuses with no package. Nonzero exits/invalid text may yield a complete capture
simulation with a refused output classification;neither means accepted runtime.

No automatic persistence. Existing file store can serialize a complete package
as bytes,but it does not retain this lifecycle result/trace or establish timeout,
cleanup,command/time provenance. No claim of durable simulation evidence follows
from separately writing only Package. Partial/failed simulations never masquerade
as existing complete packages;schema migration/integration is a future decision.

## Required future verification (not authorized yet)

Future implementation plan must specify a finite fresh synthetic-only allowance
and execution method. No old budget may be transferred. Cover literal byte
interleavings,UTF8/UTF16 multibyte splits,empty streams,EOF/exit permutations,
nonzero exit/invalid-text retention,deadline119999/120000/120001,no exit/EOF,
trailing events,duplicates,backward/equal times,unknown/malformed/plain-property
types,array/event/chunk/stream bounds,and post-return caller mutation.

Independently assert known Base64/length/hash values and conservative flags;
verify refusal preserves precisely the admitted prefix and failure index.
Include early exit followed by buffered chunks before EOF. Never replace input
bytes with decoded strings or use a fake that only asserts its own calls.
TDD/red-green and one scoped review belong to that later approved plan;no new
implementation,fixtures,hosts or tests execute during spec writing/review.

## Remaining external gates

Stage B remains NOT READY/NOT APPROVED. Historical safety disposition,real raw
collector/owned-host watchdog,exact runtime packet and separate finite runtime
approval are unresolved. I2/I3 and full C12/ordinary acceptance remain OPEN.
No main merge,security changes,OS scheduling/cleanup guarantees or real command
provenance follows even from a future green simulation suite.

## Spec self-review and handoff

Checked input types/units/bounds,exact event fields,deadline equality,completion
versus classification,trailing input,prefix retention and unchanged dependency
schemas. No executable example,live identity,nonce or launch hash issued.
No implementation-plan placeholders are implied:allowance/method require their
own later review. Request human review of this written spec before writing an
implementation plan. This document itself grants no test/runtime authority.
