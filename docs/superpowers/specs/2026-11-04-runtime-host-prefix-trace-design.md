# Real-host bounded prefix trace design

Status: DRAFT / PENDING_DESIGN_APPROVAL / NO_EXECUTION_AUTHORITY.

User selected option1 (可以): retain real-host setup, guards, case wrapper and
result output, stopping after the first baseline case. This is a small diagnostic
extension, not a product fix. Selection permits writing this design only.

## Intent and approaches

Diagnostic27 completed in14.458seconds but bypassed original host setup and
Invoke-ODContractCase/LS-Result. The original host timed out without stage timing.
Goal: identify the last completed boundary in that more representative path.

Selected: versioned host-derived prefix with the original setup order, recorder
identity observations, publisher import, suite bookkeeping, real case wrapper
and first result output. Smaller than a full replay; does not cover later cases.
Alternative: trace all52groups. Broader coverage and execution effects, rejected
for this decision. Repeating the14-definition standalone prefix would leave the
same wrapper/output coverage gap and is not selected.

## Scope and preservation

Keep all31operational files, original manifests, prior intents, stops and raw
evidence byte-identical. Never resume the stopped driver or reuse a capture slot.
Author separately named diagnostic helpers and a fresh namespace
runtime-host-prefix-trace-v1. Use hash-bound original sources and a documented,
mechanically checked transformation map; never hand-reconstruct successful data.

Preserve original runtime7.6.5/Core/x64 and fixed non-admin identity, host setup
order, guard identity semantics, publisher/private scopes, manifest/evidence
order, three contexts, original first-baseline copy and predicate, wrapper
pipeline, suite/details updates and LS-Result behavior. No algorithm changes,
caching, disabled guards, relaxed checks or output substitution.

Host-derived code must resolve original source-relative imports correctly.
The implementation plan must enumerate changed source spans, extraction/import
bindings and scope effects, including scriptblock source locations. Source
identity must still refer to original pinned bytes, never relabel transformed
code as original. If scope/guard semantics cannot be preserved and checked
statically, stop and revise the design before execution.

## Proposed trace and stopping boundary

Record flushed monotonic begin/end/error events around setup/imports, availability
output, source loading, manifest, each historical evidence/context, baseline copy,
case-wrapper entry, body entry/return, suite update, result write, wrapper return,
and diagnostic finalization. Record original exception entry before projection
where possible without replacing its error. Keep event helpers silent.

Normal cutoff is immediately after the original first BaselineCore
Invoke-ODContractCase returns (including LS-Result and LS-Need); no second case,
mutation, report, unit or recorder-completion test may execute.
Do not throw a synthetic failure to reach the normal cutoff. Use a statically
bounded prefix and explicit diagnostic completion path. An assertion rejection,
wrapper error or output error is a failure, not successful cutoff.

No additional per-result Flush inside LS-Result: that would change the path under
investigation. Retain its original WriteLine; trace a separate terminal flush
after cutoff, clearly labelled diagnostic finalization. Parent verifies actual
captured bytes, not just a write-end event.

Before cutoff, preserve the applicable original catch/failure-collection,
journal/error-summary behavior in the new evidence namespace. Never fabricate
the old52group capture evidence or run its full-suite publication/completion.
Save diagnostic errors/partial events instead. Thus full publication remains
outside coverage even if the prefix succeeds.

## Proposed budget and process bounds

Current27aux/6hosts/3compiles/217snapshots. Propose one computation28 and at most
one owned diagnostic process counted as host7; ceilings28/7/3/217.
No additional test runs, child targets, compiles, source snapshots, reviewers,
installs, security changes or diagnostic network activity. Source transformation
metadata and portable evidence packaging are not new operational source snapshots.

Exclusive CreateNew intent consumes the attempt before validation/start; existing
intent, stop or partial artifacts forbid reuse. First unexpected preparation or
execution result stops; no automatic repair/retry. No running smoke test.
Execution120000ms begins before stream/process setup, followed only by bounded
owned-exit2000ms and stream-drain2000ms cleanup. Preflight recorded separately.
Only the local Process object may be terminated; no PID lookup or tree kill.

Assistant proposal: reuse the bounded diagnostic parent stream discipline from
diagnostic27 (8192-byte reads,65536bytes per stream), not the old unbounded
CopyToAsync parent. This is an explicit representativeness limitation: success
will not clear the original parent's transport behavior. Do not change parent
strategies within this one attempt.

## Evidence and interpretation

Journal at most4096events/1MiB, UTF8, monotonic ticks/frequency, ordered sequence
and interval IDs. Flush each event, preserve partial/torn records, fail on overflow.
Do not log input payloads or secrets. Event overhead is included in timings.
Keep actual tool invocation/initial/polls/final, before/after hashes, child/parent
results, stdout/stderr, exit/timeout/cleanup state and exact stop reason.

Only report prefix completion when the actual baseline case returned successfully,
expected availability/baseline lines were captured, streams/owned exit completed,
all required events are valid/closed, identities match and no recording error
occurred. Exit0 alone is insufficient. No full-suite PASS or product acceptance.
Unmatched events locate an unfinished interval, not a proven CPU/I/O root cause.

Stop after one attempt even on success. Decode-verify the portable archive, update
status/recovery, commit explicit necessary files and push the existing development
branch. Preserve every historical stop. No merge or further execution implied.

## Approval boundary and remaining planning work

Design approval permits a concrete implementation plan only, not execution.
Plan must specify exact files/commands, hash inventory, transformation/guard
equivalence checks, error routing, cutoff enforcement, event schema, expected
output and recovery archive. Any extra required work must be surfaced before run.
Execution needs approval of that plan and its new finite allowance.
Counters unchanged by this document; overall NOT_ACCEPTED.
