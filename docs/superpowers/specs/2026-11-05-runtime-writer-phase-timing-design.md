# Bounded writer exit-phase timing design

Status: DESIGN_PENDING_APPROVAL. User selected the existing diagnostic-channel
approach; this document does not authorize implementation or another observation.
Current counters42/30/6/217 remain unchanged. Small instrumentation extension to
the existing isolated-writer design, not a new transport or timeout repair.

## Intent and approaches

Determine which post-task-return phase is last observed before writer termination.
Observation42 received COMPLETE but timed out before normal writer exit. Existing
producer input.after took2579.3967ms; that is not a measurement of writer hashing.

Selected by user: bounded records on the writer's existing stderr pipe. The
supervisor already drains that pipe into a4096byte buffer and preserves raw bytes.
No new process, endpoint, compiled code, trace file or permission is needed.
Alternative: separate trace file; rejected because additional disk writes would
interfere more directly with the suspected filesystem work. Buffering only in
writer memory cannot preserve progress if the writer is terminated, so is unsuitable.

## Unchanged boundaries

Version new PowerShell role/driver/capture bindings and evidence namespace; never
edit computation42 sources, receipts or claims. Reuse the pinned helper and adapter
DLLs. Preserve original business guards, two groups, cutoff and62-event structure.
Do not remove/reorder any full input verification, mapping, or original error path.
Keep120000ms execution,5000ms readiness,2000ms shared normal drain/finalization,
and2000ms shared failure exit confirmation. Preserve original process ownership,
termination, pending-resource retention, no-retry and unavailable-output policies.

Writer stderr is diagnostics only; producer stdout/stderr remain the business
payload, requiring exact48/0bytes. Writer stdout remains READY/COMPLETE/ERROR only.
The existing supervisor's protocol, caps and success conjunction do not change.
All five capability flags remain false; overall NOT_ACCEPTED even if this limited
observation passes. No claim about the original full-host timeout or full suite.

## Proposed record contract (assistant design choices)

Emit compact UTF8 JSON followed by LF to the writer's standard-error stream.
Exact fields: Schema=1, Token, InputSHA256, Role="writer-timing", Seq, Phase,
ElapsedTicks, Frequency. Token/input hash bind records to the same attempt;
no paths, contents, credentials, exception messages or per-file records.
Seq begins1, increments by1; monotonic elapsed ticks use a single Stopwatch started
immediately before invoking Writer.RunAsync. Frequency is positive and constant.
This clock is local to the writer; do not subtract it from supervisor timestamps.

Maximum10records,384bytes per record including LF,3840bytes total, leaving256bytes
below the existing4096byte diagnostic cap. Validate encoded size/count before each
write. One write and flush attempt per record; no retries, periodic heartbeat,
background tasks, process launches or buffer growth. The plan must test bounds
with maximum-width supported numeric values and actual token/hash lengths.

Normal sequence (nine records):

|Phase|Emission boundary|
|---|---|
|task.returned|After terminal task GetResult returns, before entering existing finally work|
|pins.begin|Immediately before Assert-HTInputs in the writer after-map path|
|pins.end|Immediately after that full verification returns|
|map.save.begin|Before existing map serialization/CreateNew/write/flush/close|
|map.save.end|After existing Save-IW returns|
|map.meta.end|After original map metadata/length/hash calculation returns|
|role.save.begin|Before original role-result serialization/publication|
|role.save.end|After original role-result publication returns|
|exit.ready|Immediately before the original exit statement|

Instrument only the after-map path, not writer before-map or producer operations.
At most one phase="error" record may terminate a failed trace, using only the
same fields. A failed original operation must still follow its original error
and cleanup path. Trace construction/write errors latch instrumentation failure,
disable subsequent trace attempts, and must not bypass original cleanup/checks;
they prevent an accepted observation (nonzero role exit where possible). Never
retry output or swallow a business/identity failure to obtain a timing record.
Publication failure after role-result is still failure; do not rewrite that file.

## Interpretation and acceptance

Keep transport outcome, business verdict and timing completeness separate.
Version the read-only verifier: replace its old writer-diagnostics-empty assertion
with strict schema/identity/size/order/tick validation. Never relax producer stderr0.
Validate exactly the nine-record sequence for a successful timed observation,
normal owned exits, all tasks successful and all original business/output checks.
Trace error/missing/malformed data prevents accepting the timed observation even
if the original transport result says SUCCESS. Do not rewrite that transport result.

On a failed observation, parse only already captured diagnostic bytes after
confirmed terminal supervision. A complete valid prefix may be reported as partial;
a truncated final line stays raw evidence, not a fabricated record. An error record
ends interpretation. Never inspect unavailable capture data or treat COMPLETE as
success. Unknown exits/pending resources retain their original stop conditions.

A start without its corresponding end means only "end not observed". A measured
interval includes trace writes/flushes, scheduling and original work; it is not a
pure hashing/disk benchmark. No records after COMPLETE might mean task return,
instrumentation or scheduling did not progress. task.returned is NOT the timestamp
of COMPLETE emission. exit.ready is NOT proof of process exit. No cross-pipe arrival
ordering inference. Synchronous diagnostic output can itself stall; byte limits
bound volume, not wall time, and this uncertainty must be included in the report.

## Planning and authority boundary

After design approval, write a finite implementation/validation plan with exact
new paths, static/record-parser tests, preserved raw evidence, preparation checkpoint
before launch and explicit stop rules. Proposed runtime scope for that future plan:
one computation43, at most writer+producer two starts, zero compilation/snapshots;
proposed ceiling43/32/6/217. This is a proposal, not a currently approved allowance.
Any test that starts a process or executes workload needs a ledger slot; no hidden
smoke test, automatic repair/retry, benchmark or extra reviewer process.

Preserve/push necessary source, design/plan and portable records on the existing
development branch; never commit binaries/caches or merge main. If evidence remains
inconclusive, report that rather than changing deadlines or dropping verification.
No instrumentation or runtime work is performed by saving this design.
