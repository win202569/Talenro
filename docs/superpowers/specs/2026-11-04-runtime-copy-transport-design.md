# Bounded CopyToAsync transport comparison design

Status: DRAFT / PENDING_DESIGN_APPROVAL / NO_EXECUTION_AUTHORITY.
User selected option1 with 是: retain the successful traced host prefix and compare
bounded asynchronous-copy transport. This small diagnostic design grants no run.

## Intent and approaches

Diagnostic28 completed the real-host-derived prefix in13.234seconds using polled
ReadAsync output. Original driver used two CopyToAsync tasks followed by a timed
WaitForExit, and its earlier host timed out. Root cause remains unconfirmed.

Selected: run the same bounded host workload with asynchronous-copy transport
and fixed-capacity destinations. It covers the workload missing from a toy probe.
Alternative: minimal output-only child. Simpler but omits real setup, guards and
wrapper behavior, so not selected. Exact old unbounded FileStream copying is
excluded because it removes the retained output hard limit.

Reference is the saved diagnostic28 result, not a newly rerun control arm.
One sample at a later time is not a controlled causal experiment or benchmark.

## Immutable workload and new namespace

Current28aux/7hosts/3compiles/217snapshots. Propose exactly computation29 and at
most one owned process counted as host8; maximum29/8/3/217.
Use fresh runtime-copy-transport-v1 namespace and separately named helpers.
Keep31operational sources, old diagnostics/inputs/claims/stops/archives unchanged.
Never execute consumed parent28 or write into its evidence directory.

Keep diagnostic28's host setup, guards, original three contexts, first baseline
wrapper/result writes and cutoff. No later tests, caching, weakened assertions
or added per-result flush. Preserve its stage markers and journal bounds.
The host is not literally byte-identical: fixed namespace/helper references and
computation binding must be versioned for29. Record exact mechanical edits and
prove all other workload statements unchanged before execution. Derived helpers
remain beside originals to retain PSScriptRoot directory/import behavior.

Use the same fixed7.6.5/Core/x64 runtime, non-admin identity and repository root.
Before/after checks bind original inputs plus successful28 artifacts and new
helper identities. All five capability flags remain false; NOT_ACCEPTED.

## Parent transport proposal

Start exactly two CopyToAsync tasks on the owned process's stdout/stderr before
the finite process wait, matching the original broad start-copy-then-wait order.
Use explicit8192-byte copy buffers and no PowerShell worker-thread callbacks,
compiled stream classes or new dependencies.

Assistant choice: destination per stream is a writable, non-expandable MemoryStream
backed by one fixed65536-byte array, initially at position0. A capacity-only
constructor is forbidden because it may expand. Capture only bytes actually
written (terminal Position), NOT the initial array-backed Length or the entire
zero-filled backing array. Do not represent unwritten padding as captured output.

The destination must refuse writes beyond its capacity, never truncate silently
and continue. Record copy task failures independently from child status; a
failed chunk may leave fewer than65536bytes retained. Do not invent missing bytes
or claim a complete prefix on overflow. No unbounded disk output or memory queue.

Use one WaitForExit with the remaining execution budget, as in the original
ordering, not the prior polling/draining loop. A copy fault may only be observed
after that wait; the destination still enforces the byte bound immediately.
Do not claim immediate fault-driven termination during the blocking wait.

Execution deadline120000ms begins before buffer/process setup. On deadline/error,
terminate only this invocation-local associated Process, never PID lookup,
adoption or a process tree. Keep owned-exit confirmation2000ms and total shared
stream-drain2000ms; record actual elapsed separately.

After exit/termination, inspect both tasks with a bounded wait, preserving each
success/fault/cancellation state rather than only an aggregate WaitAll exception.
Do not read/copy/dispose a destination while a writer may still be mutating it.
If a task is not terminal by the shared drain limit, report its output unavailable,
retain the child journal and task state, and mark the entire attempt incomplete.
Do not fake an empty stream or wait indefinitely. The implementation plan must
spell out source-stream shutdown and safe release without extending these limits.

For terminal tasks only, save retained bytes to exclusive output files and record
actual length/hash. A saved zero-byte stderr is valid only when its actual task
terminated with zero captured bytes. Preserve task fault evidence even when the
child exited0. Parent archival time is not child execution time.

## Static preparation and one-shot evidence

Before execution, statically check namespace-only workload edits, original guard/
wrapper/cutoff preservation, fixed buffer construction, both-copy-before-wait order,
one process start, bounded ownership/cleanup, input identity checks and output
extraction based on written position. No smoke tests, repeated controls, additional
target children, compiles, operational snapshots or reviewers. No installs or
security-setting changes. Diagnostic code performs no network activity.

First unexpected preparation/execution result stops without automatic repair or
retry. CreateNew intent consumes29 before validation/start; a failed start still
consumes the computation, but host8 counts only if started. Existing artifacts
forbid reuse. Successful completion also ends the scope.

Success requires the actual first wrapper return, exactly two successful groups/
SchemaSuite rows, complete62-event expected structure (or explicitly justified
unchanged-workload schema in the plan), matching pre/post identities, correct
48-byte stdout, empty stderr, both copy tasks successful, owned exit/streams
complete, exit0 and no timeout/overflow/recording/cleanup errors.
The child journal remains bounded4096events/1MiB and flushed on each event.

Preserve actual tool initial/polls/final, intent/claim, source-transform map,
input/helper identities, child/parent outcomes, task errors and raw available
streams. Archive with decode/length/hash/path verification; update result/status/
recovery and push the existing development branch. No main merge.

## Interpretation and limitations

This changes both transfer mechanism and destination type: bounded memory replaces
original direct FileStream writes. Copy buffer size, later run conditions, helper
locations and added tracing also differ. There is no same-time paired control.
Success does NOT clear the original disk transport, later suite/publication path,
or historical transient conditions. Failure does not by itself blame CopyToAsync.
Do not extend timeout, optimize sources or claim a product fix from either result.

## Sources and approval boundary

API feasibility references, not evidence of behavior in this fixed local runtime:
- Microsoft MemoryStream constructors document byte-array-backed non-resizable
  instances: https://learn.microsoft.com/en-us/dotnet/api/system.io.memorystream.-ctor
- Microsoft Stream.CopyToAsync documents asynchronous copying and task-reported
  operation exceptions: https://learn.microsoft.com/en-us/dotnet/api/system.io.stream.copytoasync

Design approval permits a concrete implementation plan only. That plan must pin
exact source versions/namespace substitutions, buffer constructor overload,
terminal-task extraction, cleanup/error routing and output checks. If bounded
safe handling cannot be established, stop before execution rather than weakening it.
Execution requires approval of that plan and its proposed29/8 allowance.
No helper implemented, probe or diagnostic run performed by this document.
