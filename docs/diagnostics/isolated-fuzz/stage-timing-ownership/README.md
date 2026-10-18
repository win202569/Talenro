# Stage-boundary ownership timing — PREPARED, NOT RUN

Design approved in chat after fb38f5b8. Approval covers code/checklist preparation
ONLY: no target host, candidate import/helper call, benchmark, runtime test or
new execution allowance. Old remaining slots stay frozen. Nothing is accepted.

## Files and preserved subject

owned-return.tests.ps1 is a standalone instrumented copy of the frozen view-aware
test. It imports the existing scalar-owned-return-candidate module, pinned
1971F808. No candidate changes; the original frozen test/controller/oracle/raw
records remain unchanged. Original 15 scalar vectors, 39 behavioral cases,
independent 13-root/22-Decision oracles, both mutation phases, all current and
retained prior-unselected view/reread checks remain. No per-leaf timing or scan
optimization. Historical selected contexts remain excluded, as before.

Timing is inserted inline around existing operations, not around new scriptblock
execution wrappers. No subject, wrapper, result, descriptor or payload is passed
to the timing helpers. They consume only stage names and scalar diagnostic state.
Removing timing-only code and normalizing separators preserves the original
4910 significant PowerShell tokens; this is a static preservation check, NOT
proof of identical runtime identity/pipeline behavior.

output-inventory.json preserves the original 153-line non-TIME inventory and
53 fact schemas unchanged. Its explicit ordered StagePlan has 670 pairs:
1340 TIME records / 1493 total nonempty success lines. No per-leaf records.
The stage-entry/exit order is independently fixed before any possible launch.

inspect-capture.ps1 is an OFFLINE text checker. It reads saved UTF-8 text and the
inventory; no launch, imports, candidate calls or fixture evaluation. Prefix mode
can report PREFIX_ONLY, never a full run pass. Complete text is COMPLETE_TEXT_ONLY
and still needs the separate process receipt. The checker has only been parsed
and inspected, NOT behaviorally exercised. Its exit status alone cannot be used
as suite success or acceptance.

## Clock and accounting semantics

GetTimestamp/Frequency supplies one monotonic tick scale. Durations are flat,
non-overlapping stage-body intervals: module load, sharing controls, direct
helpers, fixtures, decisions, literal oracles, view capture/selection, selected
mutation, current scans, prior scans, and record/retain. Each stage publishes
entry before the body and leave after successful completion.

TIME records use Console.Out directly so logging cannot add objects to the
candidate result pipeline. They share stdout with normal records; validators
check the two logical ordered inventories separately. Do not assume a fixed
cross-format merge order from PowerShell pipeline versus direct-console output.

Entry serialization/output/flush completes BEFORE body start. Body end is read
BEFORE leave serialization/output/flush. T-Emit measures its own record building,
serialization, write and explicit flush, accumulating that separately. Each
record's EmitTicksBefore contains completed prior TIME emission intervals,
NOT the duration of the current emission. The final emission duration is
unreported; this is not an exact total-logging-overhead measurement.

Boundary bookkeeping outside T-Emit, original output not inside a timed body,
script setup and any other gaps remain unclassified. This remainder must not
be relabeled as candidate time or removed to produce a "corrected" benchmark.
Original FACT output is inside record-retain; original PASS/BEGIN output is
outside these intervals. Parent launch/parse time is outside the child's origin.
No claim that logging is free or that the instrumented run measures baseline
performance. It may time out earlier due to instrumentation.

Counters are cumulative observations at the last emitted boundary. A complete
decision stage increments a separate completed count after the original call.
Scalar/metadata counts are similarly bounded (80 decisions,29 direct scalar,
1 direct metadata). Counter data in a prefix can lag work after its last record;
do not infer final interrupted counts from it. Internal candidate helper calls
inside decisions are not counted as direct helper calls.

An unmatched entry means the stage was ANNOUNCED, not that its body definitely
started. Killing during boundary output or before body start is possible.
A truncated/malformed line, missing event, sequence mismatch, nesting, changed
frequency, negative/overlapping accounting or wrong counts is invalid/incomplete,
not a skipped event or an ownership finding. No monotonic clock -> no fallback.

## Before any execution

Read CHECKLIST.md. No launch controller or new raw packet is prepared by this
checkpoint. A separately authorized controller must pin this inventory and
sources, preserve all582old entries plus four prior packet anchors, preconsume a
NEW allowance and retain the existing30s +2s owned-exit +2s drain safety policy.
It must validate receipt AND both logical output inventories, not just process
exit or inspect-capture's status. Never call the old frozen controller's later
slots. No retry, repair-after-launch, warmup, extra probe/reviewer, main merge,
I1/security/runtime/C12/release/PS5/performance acceptance is implied.

Only parse/data/static-preservation checks were performed during preparation;
there is no RED/GREEN/runtime proof of this instrumentation or offline checker.
