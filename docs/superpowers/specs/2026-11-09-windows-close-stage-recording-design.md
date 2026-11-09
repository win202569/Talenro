# Windows close-stage recording — short design for approval

Status: DESIGN APPROVED by user “是” after the79611414 design handoff.
Proceed to implementation planning; execution method and plan review remain pending.
No runtime approval is implied. This is diagnostic instrumentation, not a shutdown
fix or product feature. Earlier “proposed” wording records the design's provenance.

## Intent and scope

Determine which close-handler/disposal boundaries positively completed in the
original demo path. Preserve current event order, exception propagation, UI,
services and shutdown semantics. Preserve all existing demo/prototype artifacts.
Previous experiments did not reproduce the stall and do not establish its cause.

## Approaches

The recommended stage-recording direction was accepted for further design by
“继续下一步”; transport below is the agent's proposal, not a prior user decision.

- Remove/disable observer: coarse comparison changes subscriptions, property reads
  and title writes together, without showing the stalled boundary. Not selected.
- Synchronous file/stdout stage writes: simple but may block the observed UI close
  path. Not selected.
- Proposed: preallocated bounded in-memory queue plus background local writer.
  Close markers do not wait for disk or writer completion. Positive records are
  useful; tail loss is explicitly possible. No claim of a reliable last-executed
  stage from a truncated trace. A shared-memory external reader could improve
  survival, but introduces cross-process publication/lifetime complexity; deferred.

## Isolation and activation

Opt-in Debug/layout-diagnostics-only instrumentation, default off. Separate
intermediate/output directories; never overwrite existing diagnostic binaries.
Reject Release or incompatible flags before build. Normal build must exclude
recorder and added call sites. Runtime requires a fresh run identifier and a new
local output file in a dedicated ignored diagnostics folder; reject path escape,
existing-file overwrite and invalid initialization before creating the UI window.
The implementation plan must specify exact flags/paths and verify isolation.
No arbitrary text, command or path accepted from UI input.

## Producer and stages

Only the UI-thread close/dispose path produces stage records; initialize before
window creation. Reserve a fixed64-record queue. Each record contains schema,
launch correlation, monotonically increasing sequence, fixed stage identifier,
Stopwatch timestamp and its frequency. No search text, clipboard, exception
messages/stacks, account data or arbitrary string payloads.

Mark close-entry, dispose-entry, already-disposed-return (when applicable),
before/after each of the six root/pageHost removals, window.Closed removal,
conditional XamlRoot/UISettings removals (explicit skipped marker for null),
root-reference clear, dispose-return and close-return. Preserve original operation
order and disposed assignment. Never emit a return marker from a finally block
when the underlying operation throws. Do not catch/suppress application failures.

Producer must not allocate per marker, lock, spin/retry, perform file/network/COM
operations, wait, flush, join or dispatch UI work. Use tested single-producer/
single-consumer publication; overflow drops new records and increments a health
counter rather than wrapping/overwriting earlier entries. Unexpected producer
thread invalidates recording and does not change application behavior.

## Writer and interpretation

Background thread, not a foreground lifetime keeper. Drain finite records to the
preopened local file, bounded periodic polling; no close-time flush/join. Initial
ready record confirms transport initialization; subsequent health records include
drop/failure state when possible. Writer errors disable recording, never throw
onto the UI thread. File writes may block the writer; do not describe the whole
system as nonblocking or zero-overhead. Background exit can lose pending records.

External retained-handle process observations remain independent. Readers reject
wrong correlation/schema, noncontiguous sequences, invalid stages/timestamps,
truncated records or reported recording failure. A valid prefix proves only its
published stages. Before-without-after does NOT prove that operation blocked;
missing close-entry does NOT prove callback absence. All completed close markers
plus continued native process liveness would positively narrow investigation to
outside/after this handler. Lost/ambiguous traces remain INCONCLUSIVE, not PASS.

## Verification and next gate

After design approval, write implementation plan before code. Plan behavioral
tests for queue ordering/capacity, overflow/failure isolation, truncated/mismatched
receipts, exception paths and isolation gates; no source-marker-only substitute
for runtime behavior. Build verification must preserve old artifact hashes.
Tests/build are not WinUI runtime acceptance. Any instrumented app launch needs a
separate bounded approval; none authorized here. Recorder overhead and tail-loss
limitations must be reported, not waived to obtain a result.

Non-goals: no repair, event removal changes, forced exit, debugger/dump, service or
security changes, production telemetry, new launch budget or main merge. Necessary
source/tests/design/recovery files may later be committed; raw logs remain ignored.
