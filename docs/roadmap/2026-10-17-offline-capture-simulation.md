# Offline capture simulation — execution and scoped review

User approved [spec](../superpowers/specs/2026-10-17-offline-capture-simulation-design.md),
[plan](../superpowers/plans/2026-10-17-offline-capture-simulation.md),inline
implementation and4 fresh synthetic-only verification hosts after88d623d3.
This report does not authorize a real collector,watchdog,Stage B or acceptance.

## Implementation and boundaries

New pure `output-capture-simulation.ps1` exposes Get-OfflineCaptureSimulation,
loading only unchanged evidence-package/classifier helpers. Supplied plain
events carry bounded channel bytes,EOF,exit and virtual milliseconds. Both EOFs
and one exit before120000ms are required;boundary equality times out. Early exit
may precede buffered bytes. Invalid/backward/duplicate/trailing/overlimit events
refuse without appending the offending chunk;admitted prefix is retained as
immutable Base64,length/hash snapshots. Failed/incomplete simulations have no
Package. Complete lifecycle may still carry refused output classification.

No source-side file I/O,real clocks/sleeps,threads/tasks,process/handle/PID APIs,
external streams,security queries,network,nonces or native identities. Existing
file store/runtime/evaluators unchanged. No automatic persistence/provenance,
OS watchdog enforcement,cleanup or real scheduling proof. Conservative flags
OriginVerified/Acceptance/RuntimeApproved alwaysfalse. Caller mutation DURING
evaluation/resource exhaustion not promised safe;post-return snapshot isolation
is tested. In-memory work bounds do not cap caller allocations/overall CLR RSS.

## Four-call record (allowance exhausted)

|Call|Host/phase|Child exit|Elapsed ms|PASS|Decoded stderr artifact bytes|
|---|---|---|---|---|---|
|1|PS7 RED|1|1730|0|371|
|2|PS5 RED|1|2323|0|563|
|3|PS7 GREEN|0|2394|100|0|
|4|PS5 GREEN|0|1936|100|0|

Both REDs reached expected OfflineCaptureSimulationRequired missing API guard,
not syntax/infrastructure failure. Each GREEN completed all100 assertions.
Observed GREEN hosts:7.6.5 Core x64 and5.1.26100.9711 Desktop x64. Harness hosts
are counted test processes,not collection/probe workloads. Each had120s outer
wall-clock cap,sequential,no retries. Simulation's120000ms is separately virtual.

Important bookkeeping preservation:after correct first RED,the outer caller
incorrectly tested stale global LASTEXITCODE and threw Unexpected RED harness
failure. The raw outer failure remains. Readonly inspection of the saved single
completed record and exact missing-API stderr established expected childexit1;
no first-run replay or index rewriting. Subsequent calls rely on runner's explicit
child-result checks as the plan requires. No inference from parent exit alone.

Independent expected-value coverage includes abc Base64/hash,UTF8/UTF16 split
codepoints/progress,invalid binary bytes,all6 EOF/exit permutations,early buffered
chunks,missing observations,119999/120000/120001 and Int64.MaxValue virtual time,
trailing/duplicate/backward events,ScriptProperty getter not executed,strict header
and field types,4096/4097 events,65536/65537 chunks,independent1MiB channel caps,
whole-chunk refusal and post-return mutation. Common result checks verify exact
field set,schema,flags,snapshot metadata and package presence/consistency.

Do not claim each individual assertion was independently observed RED:both REDs
stopped at the missing API guard. No full-project suite or spent old suites ran.

## Hashes and portable recovery

Source SHA256:`8E0ABE718DAE9AB51D646D9685D2C2FEA6CA6638C70880C0751CDB5D5599F7A4`.
Tests SHA256:`B6DC1F3C4FAE1F24DA0BBECB92FE019DB11CF48EE7B7C3E4B1922C52EC5A908B`.
Ignored four-entry invocation index SHA256:
`F4FEA30F1728D27A48ED446BFAA88BAB9633E640938E02DC5403F4E07B4BF3BD`.
Audit runner/progress/stdout/stderr/index reside under ignored
`.superpowers/sdd/2026-10-17-offline-capture-simulation`;text redirects are decoded
and re-encoded audit output,not raw byte capture from a real diagnostic. Portable
source/tests require no ignored artifacts. No actual raw payload is committed.
Original ledger70/70,classifier8/8,package4/4,store/repair/elevated/coverage records
and frozen unused slots remain unchanged. New simulation4/4 is separate.

## Execution rulings

1. PowerShell-native ledger/runner instead of Bash skill scripts;no task-done
   extra run,baseline/install/full-suite run or scratch deletion. Approved cap
   and preservation take priority. Cost:no full-project regression claim.
2. Resolve stale outer LASTEXITCODE against explicit saved child evidence,then
   continue remaining approved calls without replay. Cost if wrong:false
   continuation;mitigated by exact count/completed/exit/assertion checks and raw
   outer failure preservation. The incorrect parent command is not called clean.

## Review disposition and next boundary

One fresh scoped readonly reviewer returned0Critical,1Important,0Minor. Required
fix:output-capture-simulation.ps1 lines4/36 invoke `.GetType()` on unvalidated
caller objects. An ETS ScriptMethod named GetType can execute before event
descriptor validation or outer-array refusal;catching afterward cannot undo
effects. This is a static source finding,no reproduction executed. Existing
ElapsedMs ScriptProperty counter fixture does not cover this dispatch. Source/
test hashes match the reviewed files;no post-GREEN edits or second review.

Disposition:Important accepted,Task1 overall INCOMPLETE despite100PASS/host.
The callback-free boundary is a promised behavior,not excluded hostile concurrent
mutation or process-wide function tampering. Do not call this safe/completed.
No fixes/new tests under exhausted4/4 cap. Next proposal:counter-only failing
fixtures for a hostile outer value and event,non-dispatching type/shape checks
(including validation through unchanged helpers),then GREEN/full new suite in
at most4 separately approved synthetic hosts,PS7/PS5 RED thenGREEN,120s each,
sequential,no retry. Preserve original tests/results;no relaxation or new runtime
authority. No second independent review loop planned for the one fix pass.

### Rulings on declined review scopes

3. Reviewer did not independently execute hosts,validate raw receipts/runner or
   individual RED sensitivity. Coordinator inspected saved index/outputs/counts
   and documented the stale outer failure;not an independent rerun/certification.
   Cost if over-relied upon:undetected harness defect/false coverage confidence.
4. Real pipes/process ownership/wall clocks/watchdogs/scheduling/cleanup/command
   UTC provenance absent by design. Remain unsupported;cost if reused as real
   evidence:false runtime safety conclusions.
5. Concurrent mutation/process-wide command or type tampering/resource exhaustion/
   peak-memory/performance excluded or unmeasured. No sandbox/latency proof;
   cost if relied upon:callbacks/resource failure or timing misinterpretation.
   This does not excuse the admitted Important ordinary-input callback defect.
6. Encoding detection/authenticated origin not supplied. Caller metadata only;
   cost if assumed:misclassification or false provenance trust.
7. Persistence/lifecycle durability/store integration not implemented here.
   Package alone omits trace/status;cost if assumed:loss of lifecycle evidence.
8. Dependency correctness reviewed only as consumed,not independently certified.
   Existing package/classifier results remain scoped;cost if assumed:unseen
   dependency defect mistaken for a universally certified pipeline.
9. Whole-project tests/installations/host security/historical clearance/I2/I3/
   Stage B/full acceptance outside review. All remain blocked/unverified;cost
   if inferred:unsafe or unjustified launch/acceptance.
10. Reviewer did not verify Git ancestry/status/push/handoff. Coordinator must
    verify explicit staged files,remote SHA and clean tree independently;cost
    if omitted:incomplete fresh-machine recovery.

No main merge or Stage B authority follows from100PASS. Historical safety
disposition,actual collector/watchdog,exact runtime packet and I2/I3 unresolved.
