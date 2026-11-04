# Supervised isolated writer design

Status: DRAFT / PENDING_DESIGN_APPROVAL / NO_EXECUTION_AUTHORITY.
User selected an independent writer process, design only. Current30/9/4/217.
This is a protocol/lifecycle design, not another original-transport comparison.
The adapter's24-case success is retained evidence, not approval of this topology.

## Intent, alternatives and tradeoff

Selected: an outer supervisor directly starts and owns TWO sibling processes:
a producer (synthetic first, business prefix only under later approval) and a
writer. Neither child may spawn another process. The supervisor relays producer
stdout/stderr to the writer and never opens the two capture data files.

Alternative: in-process disk adapter. Less machinery, but synchronous file opening
and pending finalization can block the controller; not selected.
Rejected topology: writer starts producer as its own child. If the writer hangs,
the outer supervisor lacks its own original producer Process association. Do not
solve that by PID adoption, name lookup, tree killing or silently adding job objects.

Assistant choice: redirected standard input carries framed bytes into the writer;
its stdout carries bounded receipts and stderr bounded diagnostic text. No named
pipe endpoint, network listener, service, credential or security-setting change.
This avoids a new endpoint discovery/authentication surface. Pipes are created by
the supervisor's exact Process.Start associations, not accepted from arbitrary
senders. A per-attempt token binds receipts to the pinned manifest and role.

The relay changes buffering, backpressure and scheduling; framed writes replace
the old direct CopyToAsync-to-FileStream path. A successful run demonstrates only
this containment protocol, NOT the old timeout root cause, exact original path,
speed improvement, full suite or publication correctness.

## Ownership and bounded communication

Supervisor owns both original Process objects, redirected pipe handles, relay
tasks, cancellation sources and bounded control buffers. Writer exclusively owns
two CreateNew output files and the frozen bounded adapter. Producer owns its normal
workload/journal, not those output files. No shared raw file handles.
Use the same fixed runtime, non-admin account, worktree and immutable source pins.
Every run has fresh files/intent/token; never reuse the consumed30 namespace.

Binary stdin protocol (assistant choice):
-16-byte little-endian header: magic4bytes "TLW1", version1byte=1, type1byte,
  channel1byte, reserved1byte=0, sequence uint32, payloadLength uint32.
- DATA(type1): channel1stdout or2stderr, payload1..8192bytes.
- END(type2): channel1or2, payload0; exactly once per channel after its real EOF.
- CLOSE(type3): channel0,payload0, only after both END frames; exactly once.
- Sequence starts1 and increases by1 across the single serialized pipe stream.
  Reject wrong magic/version/reserved/channel, out-of-order/duplicate END/CLOSE,
  DATA after END, length>8192, truncated frame or unexpected EOF.
- Per-channel DATA total<=65536, enforced before relay AND by writer before
  adapter write. At most16DATA frames per channel; at most35frames total.
  Relay coalesces small reads into8192-byte chunks, except final partial chunk:
  no per-result flush promise, and this additional buffering is a stated difference.
  On actual EOF send remaining partial chunk then END. CLOSE then close stdin
  only after its asynchronous send completed. The writer requires actual pipe EOF
  immediately after CLOSE; it must not finalize/acknowledge success before that EOF.
- Exactly one serialized send may be in flight. Each producer channel has one
  reusable8192-byte buffer until sent, plus one send buffer<=8208bytes. No unbounded
  queue, whole-output retention or PowerShell callback on worker threads.
  Both channels have active asynchronous readers; bounded fairness selects ready
  chunks without starving stderr. Counters distinguish read/admitted/sent bytes.
- Writer receives one bounded frame and applies DATA to the matching adapter,
  preserving bytes exactly. Disk writes may block inside writer only. It must not
  report success for overflow or skipped frames. ACKs do not substitute for actual
  process exit and finalization state.

Writer stdout: UTF8/LF JSON control records, maximum4096bytes each and8192bytes
total: one READY then one COMPLETE or ERROR, no duplicate terminal record.
Fields: Schema1,Token,InputSHA256,Role=writer,State; COMPLETE also carries each
channel's actual length/SHA256, copy-input EOF, close-success and fault state.
READY only after pinned assembly/input validation and both new files opened.
No stdout payload mixing. Stderr bound4096bytes; preserve actual bounded bytes.
Supervisor authenticates token/hash/role/order/schema and refuses malformed,
truncated, over-limit or contradictory receipts. Silence is not an empty success.
Start both writer control-output readers immediately after its Start, before
waiting for READY, and retain them throughout supervision. Producer readers start
immediately after producer Start. Incremental SHA256 and length of successfully
sent DATA per channel must match COMPLETE after real producer EOF; control
receipts alone are not evidence that all producer bytes were forwarded.

## Timing, normal exit and failure

These are proposed algorithmic wait budgets, not OS real-time guarantees.
Execution clock120000ms starts before writer startup/pipe setup, includes its
readiness and producer startup; readiness ceiling5000ms consumes that same budget.
Start producer only after valid READY. At most two actual starts, no restarts.

Supervisor drives all task/exit observations without blocking on a synchronous
pipe read/write or task.Result/GetResult before terminal state. No parent-side
capture-file open/flush/hash during active supervision. File I/O, source loading,
Process.Start/Kill/handle operations, scheduler delays and a hung OS still lack
hard real-time guarantees; independent writer improves containment, not certainty.

On normal producer exit, begin ONE shared2000ms drain/finalization window:
finish reading its two streams, deliver END/CLOSE/pipe EOF, observe writer
finalization/COMPLETE, and confirm writer exit0 inside the remaining window.
No new2000ms per stream, close or ACK. Writer flushes/closes both adapters once,
hashes actual closed files, then emits COMPLETE; that work is inside the window.
Large/hash-stalled finalization is failure, not an unlimited archival phase.

Any deadline, protocol/transport/copy/capacity error or unexpected exit:
- latch failure; do not start producer if it has not started; no success afterward;
- request cancellation of outstanding relay/control operations, once;
- request Kill() on each still-running invocation-local associated process,
  never Kill(true), PID lookup, adoption or broad process selection;
- share one2000ms exit-confirmation budget across both associated processes;
- retain references while asynchronous operations remain pending; never dispose
  a stream concurrently to manufacture completion, retry close, truncate/delete
  output, or perform late background evidence publication.

The failure exit-confirmation window is separate from the normal2000ms
drain/finalization window; an execution timeout skips normal drain and goes directly
to termination. Thus programmed waits are <=120000+2000+2000ms on the longest
normal-drain-then-failure route, excluding noninterruptible calls/recording.
Log actual times, not an asserted124second wall-clock guarantee.

If either associated process exit is unconfirmed, record its ownership/exit
uncertainty; no orphan-free claim. Supervisor abrupt death is also outside the
current guarantee: siblings can survive it. A job-object/OS containment solution
would need a separate design and approval; it is not smuggled into this one.

## Stable evidence versus unavailable files

Success requires valid COMPLETE, two normal owned exits, all forwarding/control
tasks terminal and successful, no timeout/termination/error, exact payloads and
matching identities. Actual closed writer files are the output evidence.
Exit0 without COMPLETE or COMPLETE without exit0 is incomplete.

On failure preserve diagnostics and mark unfinalized files OutputAvailable=false,
length/hash=null, even if a file exists or appears empty. Do not open/hash those
files during supervision. After confirmed writer exit, any salvage is a separate
read-only evidence action, explicitly not success or a retroactive COMPLETE.
If writer exit is unknown, no inspection of potentially mutating output files.

Supervisor state remains in bounded memory until terminal supervision outcome.
Recording a final receipt/archive may itself block on disk; report the supervision
outcome separately from EvidencePersisted. Do not promise portable recovery when
recording failed. Preparation pins/checkpoints are saved before launch; successful
post-run preservation uses exclusive files, decoded length/hash/path checks and
the existing branch. No binary/runtime/cache files in Git. Existing generated
adapter DLL may be loaded only after its recorded hash and source identity match;
missing DLL on a new computer requires an explicit compilation budget.

## Validation gates and budget boundary

No new computation/process/compile is authorized by this document. Current
30/9/4/217 is unchanged and exhausted. The implementation plan must enumerate:
- new coordinator/relay/protocol source, writer entry and synthetic producer;
- any necessary single compiled helper and exact reuse of the frozen adapter;
- fixed finite scenarios: normal dual-stream bytes and empty channel; capacity
  overflow; malformed/truncated/extra protocol bytes; writer blocked before READY;
  writer blocked in write; writer blocked in close; unexpected writer exit;
  producer early failure; missing/wrong/duplicate receipts; pending cancellation;
- exact per-scenario process count (zero to two starts), compile count, computation
  IDs, files/bytes, clocks, deterministic release/kill behavior and stop rules.
  Do not label intentional hung-worker scenarios as unbudgeted smoke tests.
  Expected injected failures are scenario expectations; unexpected results stop
  the gate without repair/retry or running later scenarios.
- tests of state/timeout policy without launching a business workload first;
  transport helper validation before using it to supervise a business process.

After the finite synthetic gate, report and STOP. A later separately approved
business-prefix observation would preserve original guards/wrapper/cutoff and
62-event structure, require actual48/0-byte output and all original success
conditions, and budget writer plus producer explicitly. No such host is prepared
or run under this design approval. No tests beyond the prefix or main merge.

## Scope and local evidence

Reference: runtime-bounded-disk-adapter.cs OpenNew:22-25, CloseAsync:104-118;
runtime-bounded-disk-adapter-parent-v2.ps1:127-140 stops its drain timer before
memory-output persistence; it cannot simply substitute the disk adapter.
Adapter result checkpoint e53b6d41 records24cases/111assertions and compilation4,
not supervisor/IPC verification. Original31sources and all stopped records stay
byte-identical. Five capability flags remain false; NOT_ACCEPTED.

[Microsoft Process.Kill](https://learn.microsoft.com/en-us/dotnet/api/system.diagnostics.process.kill?view=net-10.0)
documents asynchronous termination and the need to confirm exit; the design
therefore never equates requesting termination with successful cleanup.
Implementation must verify synchronous API prefixes on the pinned runtime rather
than inferring boundedness from Async names.

This document is complete enough to plan the protocol gate; code/file names and
the exact finite validation ledger belong to that plan. Design approval permits
planning only. No source implementation, compilation, process or probe this turn.

Design-only builder check: a fresh read-only agent found no material unanswered
planning questions. This was document review only, not an operational reviewer
process, runtime verification or a safety guarantee. The exact finite synthetic
gate and any business observation still require separate approval.
