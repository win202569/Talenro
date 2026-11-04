# Bounded disk-copy adapter: short design and feasibility review

Status: DRAFT / PENDING_DESIGN_APPROVAL / NO_IMPLEMENTATION_OR_EXECUTION_AUTHORITY.
User selected option1: design and feasibility review of a capacity-limited disk
adapter, not compilation or another diagnostic. Current counters29/8/3/217.
This is a narrow diagnostic extension, not a product fix or acceptance.

## Intent and approaches

Saved28 succeeded with polled ReadAsync followed by disk writes; saved29 succeeded
with CopyToAsync into fixed memory. The original timed-out host copied directly
to FileStream. Neither success isolates the old failure or clears the full suite.

Selected: a small compiled C# Stream adapter owning an exclusive new FileStream,
enforcing a per-channel65536-byte write allowance before delegating any write.
It retains streaming to a file instead of collecting the whole output in memory.
Cost: new code, compilation and independent boundary tests; behavior differs from
the original unwrapped FileStream. Selection authorizes this design only.

Alternative discussed: stop runtime investigation and retain the acceptance block.
No added tooling risk, but the disk-copy combination remains unobserved.
Rejected: unbounded original replay, periodic size polling, truncate-after-copy,
or SetLength(65536) as a supposed cap; these do not enforce a pre-write allowance.
Memory mapping or memory-then-file would change the destination semantics again.

## Proposed adapter contract (assistant design choices)

- One private file per stdout/stderr, new path in a fresh namespace, initial length0,
  CreateNew/Write/FileShare.Read, matching the original opening flags. No external
  raw stream, handle or path-changing interface. No preallocation, padding or
  per-write flush. FileShare.Read permits readers, not another writer.
- Derive from Stream, not FileStream. Write-only and non-seekable; reject read,
  Seek, SetLength and Position setters. Cover byte-array and span synchronous
  writes, both asynchronous write overloads and WriteByte; audit inherited entry
  points so none bypasses the common gate. Do not use PowerShell callbacks on
  worker threads or queue an unbounded collection of writes.
- Gate one operation at a time; concurrent write/flush/close attempts fail rather
  than wait indefinitely. Validate arguments and pre-cancellation before reserving
  capacity. Reject count > limit - reserved BEFORE touching the file, and latch
  failure. Reject the whole overflowing chunk; never truncate it and report success.
- Reserve admitted bytes before delegating, including the asynchronous lifetime.
  Do not refund a failed/canceled reservation: an underlying operation may have
  partially written. Latch faults/cancellation and refuse further writes. Track
  admitted bytes separately from confirmed writes and final observed file length.
  This conservative accounting plus no seeks bounds logical output length.
- Await the actual underlying write before releasing its gate; propagate real
  exceptions and cancellation. Recognize capacity failure through a dedicated
  adapter error, not by guessing from an arbitrary IOException.
- The cap concerns these output files, not filesystem allocation granularity,
  runtime memory, other evidence files or adversarial filesystem modification.
  Use the existing trusted local-worktree model; no security/quota changes.

## Lifecycle and evidence boundary

Future parent would retain the saved29 workload, guards, wrapper, result writes,
cutoff and62-event skeleton, with mechanically reversible namespace/identity edits.
Keep two CopyToAsync calls with8192-byte buffers before one process wait.
Retain120000ms execution budget,2000ms owned-exit confirmation and one2000ms
shared stream-cleanup budget; no per-channel fresh window or timeout increase.
No later groups, new target children, caching, weakened assertions or main merge.

Disk output files exist before completion, unlike saved29 memory output. Existence
or zero length does NOT mean available output. Observe copy and finalization tasks
separately. Only after terminal copying and successful flush/close may the parent
hash and classify stable output; save actual file bytes, not admitted-byte counters.
File finalization must share the cleanup deadline, not follow it unboundedly.

If copying/finalization remains pending, do not read/hash/dispose concurrently or
claim stable output. Record OutputAvailable=false, null length/hash, pending task
state and deferred release. Retain references through parent exit, preserve the
path, and stop. Do not fabricate an empty output, truncate, delete, or publish a
later background success. Portable archival must explicitly distinguish existing
unstable/unavailable files from absent files; no byte-stability claim while pending.
Any later post-exit read-only salvage must be separately identified and scoped.

Feasibility caveat: .NET async APIs and cancellation are NOT a hard OS I/O deadline.
Opening, write initiation, flush, disposal, hashing and archival can encounter slow
or unresponsive storage. The plan must examine synchronous prefixes and pending
operation ownership before promising bounded parent responsiveness. A timeout on
a task only bounds waiting, not underlying I/O. If this cannot be handled under
the approved constraints, STOP BEFORE implementation/run and request a revised
scope; do not silently add a supervisor process or weaken cleanup.

## Validation and approval sequence

1. Approve this design, then write an implementation plan. No code in this stage.
2. Plan a separately bounded adapter-only verification gate BEFORE any business
   host: exact-limit/one-byte-over/oversized-first-write; all write entry points;
   invalid arguments; pre/in-flight cancellation; injected partial failure;
   concurrent calls; prohibited seek/length changes; finalization faults/pending
   state; normal48/0-byte output. Deterministic test helpers, temporary paths,
   file counts and maximum bytes must be enumerated, not open-ended smoke tests.
3. The plan must explicitly budget new computations, compilation(s), any test
   process(es), durations and evidence. Existing no-compile/no-extra-run approval
   is exhausted and cannot cover this work. Prefer one frozen assembly reused
   for validation and a later host, but compilation/loading feasibility is not
   yet established; count every actual compile and process, including helpers.
4. Adapter validation failure stops that approved attempt. No automatic repair,
   recompile or retry. Successful adapter validation alone does NOT authorize the
   business host. Report evidence and obtain its separate one-shot allowance.
5. Only then may a separately approved disk-host observation be prepared, pinned,
   archived and pushed before invocation. Preserve31operational sources and all
   previous claims/stops/raw records. Never rerun consumed parents or edit their
   inputs. Fixed runtime/non-admin identity and original ownership rules remain.

No compilation, adapter test, runtime probe, helper implementation or new host was
performed for this review. Current29/8/3/217 unchanged. Exact future budget belongs
in the implementation plan, not an inferred reuse of computation29.

## Success and interpretation

Adapter gate success requires measured boundary/failure behavior, not static
presence of guard code. Host success additionally requires62complete events,
two actual successful groups/schema rows, four consistent identity maps, owned
exit0, both copies and finalizations successful, stable exact stdout48bytes and
stderr0bytes, and no timeout, overflow, error or deferred release. Keep all five
acceptance flags false and NOT_ACCEPTED.

Even a successful host would establish only this bounded adapter/file combination
for one traced prefix. New wrapper/compiled assembly, copy buffer, tracing,
cleanup and later machine conditions prevent claiming an exact original replay
or causal A/B result. Full publication, later groups and historical transient
conditions remain unresolved. A failure also does not alone prove a disk defect.

## Read-only feasibility sources

Local evidence: driver.ps1:1428 onward (FileStream/CopyToAsync/wait/cleanup);
runtime-host-prefix-trace-parent-v3.ps1; runtime-copy-transport-parent.ps1;
roadmap reports2026-11-04-runtime-host-prefix-trace-result.md and
2026-11-04-runtime-copy-transport-result.md. No original entry point was invoked.

Microsoft API contracts consulted on2026-11-04, not proof of fixed-runtime behavior:
- [Stream](https://learn.microsoft.com/en-us/dotnet/api/system.io.stream):
  subclass surface and unsupported-operation design.
- [WriteAsync](https://learn.microsoft.com/en-us/dotnet/api/system.io.stream.writeasync?view=net-10.0):
  byte-array/memory overloads and task-based cancellation/errors.
- [SetLength](https://learn.microsoft.com/en-us/dotnet/api/system.io.filestream.setlength?view=net-10.0):
  adjusts length; not a write quota.
- [DisposeAsync](https://learn.microsoft.com/en-us/dotnet/api/system.io.filestream.disposeasync?view=net-10.0):
  finalization includes backing-store changes and closing.

Conclusion: a conservative pre-write logical-size bound is design-feasible;
bounded disk finalization and exact runtime compatibility remain validation gates.
Approval of this document permits planning only, not implementation or execution.
