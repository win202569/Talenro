# Synthetic capture contract — bounded execution and review

## Latest correction attempt — stopped at fixture error; UNFINISHED

After75a6cabdc93ba33940af19ede69e8433226232a1,user approved correction plus
four NEW hosts,PS7RED/PS5RED/PS7GREEN/PS5GREEN,120s each,no retries. Existing
original4/4 remains consumed. Repair call1 PS7 exited1 after2280ms on7.6.5Core
x64. It printed1PASS(API guard) and6FAIL public-API regressions:nonempty arguments
clipped,oversized arguments accepted,nonempty chunks clipped,oversized chunks
accepted,4097-event preflight bypass and passive Rank shadow accepted. These
reproduce the Important boundary problem;no production source change was made.

The run then failed at tests:33,Shadow fixture not attached,while preparing the
ExecutableDeclaration string fixture. The expected aggregate assertion was not
reached;runner recorded ExpectedResult=false and stopped. Do NOT call this a
clean expected RED or a green78-check suite. The new helper's Add-Member call
does not use PassThru;string wrapper/attachment handling is a candidate test
harness cause,not yet proven. No fixture correction,replay or second host ran.
Current revised tests are NOT GREEN. Repair1/4 used;unused2-4 FROZEN pending
explicit continuation authority. All original logs/indexes retained unchanged.

Source SHA256 unchanged:
`09D2AF7970D67780526988A0B1409F862E12E2185B22AC7FA4F8718864A349D2`.
Revised tests:`2F016B10DAD68883FE701A5C80EFE5F75E9731E776C6523C9904E977D6C20C5E`.
One-entry repair index:
`2E6398EE4C95A6495C106E514A2E7E639AC4E1F2EDE5CDB38DDC1EB5A2C56DF5`.
Repair receipts remain local ignored artifacts under this plan's workspace
`length-shadow-fix`;tracked tests use no ignored runner/cache dependency.

Next proposed authority:preserve call1 and accept its six actual API failures as
the PS7 reproduction evidence;correct only fixture attachment/parent phase
continuation bookkeeping,then use remaining3 hosts for PS5RED,PS7GREEN,PS5GREEN
with the approved production correction. No first-call replay,no extra slot,
no second review,no retries. This continuation is NOT approved yet. Any unexpected
result stops again;do not silently renew or reset the index.

Rulings carried:PowerShell-native audit instead of Bash helpers;no additional
baseline/full-project/task-done run,install or receipt deletion because finite
authority/preservation prevails(cost:no whole-project regression certificate).
Reject shadow descriptors rather than normalize them in the proposed fix
(cost:extended values with Length/Rank are refused);no production fix executed.
No second review for the original review's single fix pass,and unverified
culture-token concern stays deferred(cost:no re-review/culture certification).
No runtime,security,Stage B,I2/I3 or broader acceptance changed.

## Original implementation and review — historical record

User approved written spec/plan,inline method and4 synthetic hosts after
30aa84f9. This report covers pure New-/Test-SyntheticCaptureRecord only,not a
real collector,watchdog,security clearance or Stage B/ordinary/C12 acceptance.

## Implemented behavior — incomplete; review-blocked

New capture-adapter-contract.ps1 imports unchanged corrected simulator/pure
dependencies only. Implements passive packet/record shapes,uncoerced CLR types,
UTF16 argument validation,UTC declarations and trace bounds. The static review
below found passive ETS length shadows defeat parts of these checks;do not rely
on this revision as a correct bounded byte-preserving contract.
It creates owned Base64/length/hash argument/chunk snapshots,derives simulator
lifecycle and stop intent,and hashes TLRCAP01 primitive manifest/full-event bytes.
Verifier reconstructs trusted values,recomputes binding/derived state and rejects
tampered labels or authority/cleanup claims. No caller containers go directly to
dependency GetType-based shape helpers. Counter fixtures cover nested methods/
properties. No universal hostile-process/sandbox guarantee is claimed.

UTC reversal retains record/bytes but refuses verification. Complete lifecycle
with nonzero/invalid output may still be internally consistent;incomplete or
timed-out lifecycle may verify as a truthful synthetic record,not runtime success.
CleanupStatus remainsunknown,all action/authority flagsfalse. Invalid preflight
returns Created=false with no fabricated binding/trace. All supplied chunks count
toward2MiB packet cap,including those after simulated failure;admitted streams
still1MiB each. No file I/O,clock queries,threads/tasks,network,process/handle/PID,
native identities or real adapter path;no persistence/schema migration.

## Fresh verification record — cap4/4 consumed

|Call|Host/phase|Child exit|Elapsed ms|PASS|Decoded stderr artifact bytes|
|---|---|---|---|---|---|
|1|PS7 RED|1|2111|0|370|
|2|PS5 RED|1|2161|0|562|
|3|PS7 GREEN|0|19255|78|0|
|4|PS5 GREEN|0|30109|78|0|

Both REDs expected missing SyntheticCaptureContractRequired guard. Each GREEN
completed78 checks;observed versions7.6.5Core x64 and5.1.26100.9711Desktop x64.
Sequential,120s outer cap per verification host,no retry/elevation/extra run.
Parent checked actual saved child exit,not stale LASTEXITCODE. Verification
hosts are counted test processes,not source-side collection/probe workloads.
No claim that each assertion individually ran RED;RED stopped at API guard.
No old-suite replay/full-project/baseline/dependency installation performed.

Coverage includes independent literal canonical-byte hash oracle,interleaving
binding,argument/raw-output preservation,all lifecycle stop-intent classes,
buffered output after exit,clock reversal,strictUTC/refusal descriptors,metadata/
UTF16/event/chunk/2MiB bounds,bytes after timeout,admitted-prefix versus retained
rejected trace,post-return mutation,25 tamper cases,nested callback counters and
UTF8/UTF16 split progress. Tests use actual constructors/verifier,not mocks.

## Hashes and portable handoff

Source SHA256:`09D2AF7970D67780526988A0B1409F862E12E2185B22AC7FA4F8718864A349D2`.
Tests SHA256:`8C2819126252B2662DE416F93EAE9E4082EF99E81707C8B0C262795B66228188`.
Four-entry invocation index SHA256:
`EFFA2A18A7AA5BDD828EA7DEA0CFBFD973E3B386A7893FE3C80E026EB415F60A`.
Ignored runner/ledger/stdout/stderr/index under
`.superpowers/sdd/2026-10-17-capture-adapter-contract` remain local audit artifacts.
Decoded/re-encoded test text is not raw capture from an actual diagnostic.
Tracked source/tests need no ignored files or prior-machine private cache.
All old budgets/records/frozen slots and prior106-check simulator evidence remain
unchanged;no source/test change after GREEN. No actual raw payload committed.

## Execution ruling and review disposition

Ruling:PowerShell-native audit bookkeeping rather than Bash skill scripts;no
task-done extra run,baseline/full-project replay/install or scratch deletion.
Approved finite cap and preservation take priority. Cost:no fresh whole-project
regression claim;local raw audit artifacts retained,not restored by cloning.

One fresh scoped readonly review concluded0Critical/1Important/0Minor confirmed,
verdict unfinished. It identified a required Important defect in
capture-adapter-contract.ps1:35-36 (related19,42,68). Passive NoteProperty members
can shadow intrinsic Length/Rank while passing Assert-ContractPassiveProperties.
For a real byte[] carrying Length=0,Copy-ContractBytes checks,allocates and copies
using the shadow value:nonempty or oversized input can become an empty snapshot
instead of refusal. Object[] length shadows bypass event-count preflight;string
length shadows defeat character/Base64 predecode bounds. This is ordinary input,
not excluded process-wide hostile tampering. Current callback fixtures omit
passive shadows. Parent independently confirmed the source path;no reproduction
or fix was executed. Existing 78PASS results do not establish this requirement.

Overall UNFINISHED. Four-call cap4/4 spent;source/tests remain exactly the GREEN
revision with hashes above. No fix,extra tests or second review under this cap.
Proposed next action,requiring separate approval:counter/byte-preservation
regressions for passive length/rank shadows,non-dispatching intrinsic checks and
copy lengths,and four new PS7/PS5 RED/GREEN hosts,120s each,no retries. Review
token/field equality for ordinal semantics during that bounded correction;
culture-sensitive comparison concern is unverified,not a reproduced failure.

Excluded/unmeasured:real acquisition,clock/command provenance,concurrency,
hostile process-wide mutation,resource exhaustion,durability,old/full-project
suite,real machine security clearance and Stage B. All prior frozen budgets and
runtime/security/acceptance gates stay unchanged. Recovery uses this explicitly
unfinished development checkpoint,not main or a claim of acceptance.

Review declined independent RED/GREEN confirmation and platform-specific
reproduction because execution was prohibited;parent owns recorded test evidence.
It declined exhaustive unchanged-dependency audit/full-project regressions,
live collection/provenance/draining/stop/cleanup,concurrent caller mutation,
hostile process-wide tampering/resource exhaustion,durable storage/migration/
fresh-machine exercise,and security/vendor/runtime/Stage B/merge acceptance as
outside scope or unimplemented. Final commit/push integrity is parent-owned and
not reviewer-certified. No reviewer edits,execution,network,Git mutation or
subdelegation. These exclusions are accepted as limits,not waived guarantees.
