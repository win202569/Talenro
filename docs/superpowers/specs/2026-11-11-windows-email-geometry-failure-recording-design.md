# Email geometry entry failure recording — short design addendum

Status: approach and written design approved by user,2026-11-11.
Implementation plan:../plans/2026-11-11-windows-email-geometry-failure-recording.md;
plan confirmed and repair implemented; final744 synthetic assertions pass. Independent
review getter finding fixed RED→GREEN. No live run authorized; see execution ledger.
This is a bounded repair to Important1 in the entry review, not a new observer.
Parent:2026-11-10-windows-c4-email-geometry-design.md. Existing entry implementation
and review disposition:docs/roadmap/2026-11-11-windows-email-geometry-observer-entry.md.

## Intent and approaches

Retain both operation and cleanup failure categories when the unchanged shared
session replaces exceptions with a generic failure. Never expose raw exceptions.

Selected: entry-owned in-memory failure state plus callback/resource wrappers.
This isolates changes to the new entry and tests, at the cost of explicit ownership
adapters. Alternative: change shared exception propagation; simpler centrally, but
affects frozen C4/C5 consumers and is outside scope. Outer-catch-only rewriting is
not viable: it cannot recover categories already discarded by the shared session.

## Proposed contract

One state object per invocation, correlated to the existing run ID. No global state,
extra process acquisition, retry, polling loop, file sink or background worker.
Record failures at their actual entry-owned operation boundary before rethrowing.
Keep ordered, deduplicated fixed category tokens, separated into OperationFailures
and CleanupFailures; do not call the first recorded error the proven root cause.
No exception message, stack, user text, arbitrary path or raw error object enters
the public summary. A bounded allowlist supplies all tokens; unknown tokens reject.

Implementation-plan choice proposed by author: use finite operation/resource names
(preflight, import, writer construction/open/write/flush, process acquisition,
identity, native/CIM/managed read, clock, sleep, publication, session) and cleanup
names (manifest writer/stream, sample writer/stream, process). Fix the exact token
enumeration in the plan; arrays are bounded by that enumeration and repeated errors
do not grow memory. Preserve each distinct operation and cleanup category, not just
one generic failure. Internal shared validation errors inaccessible to these wrappers
must be reported as SessionUnclassified, never guessed from exception text or timing.
Existing sanitized shared outcome reasons may remain evidence but are not exact causes.

## Resource ownership and unchanged evidence

Shared engine keeps its exact ten callbacks and owns disposable adapters instead
of the bare resources. This narrowly supersedes the old direct-transfer assertion.
The process adapter owns exactly the one Process returned by the launch callback;
Id forwards to it, identity/managed callbacks unwrap it, and native observation uses
its original SafeHandle. No reopening by PID, replacement Process, handle cloning,
forced termination or independent handle disposal. Adapter Dispose attempts original
Process.Dispose once, records its category if it throws, then rethrows. Repeated
adapter disposal does not retry an already attempted release.

Writer adapter owns its writer and underlying stream. Explicit ownership uses
leaveOpen where needed so writer disposal and stream disposal each have one owner.
Attempt both releases even when the first throws; record both failures. Mark each
release attempted before calling it. Creation failure cleans up resources already
acquired; failure during adapter construction cannot orphan the raw resource.
Manifest acquisition cleanup and text-writer construction cleanup use the same state,
including the currently suppressed constructor-path disposal error. Normal successful
ownership transfer clears factory ownership; no duplicate cleanup by factory/adapter.
Do not assert a real handle leak based solely on a throwing synthetic fixture.

## Final failure surface

The outer entry builds a sanitized summary after the shared scope has attempted
cleanup: SchemaVersion, RunId (null if not generated), Outcome=Failed,
OperationFailures, CleanupFailures, GeometryEvidenceStatus=Unreviewed,
Acceptance=NotAccepted. Surface it on the information channel and attach the same
allowlisted data to the terminal sanitized exception; exit remains unsuccessful.
Do not attach original exceptions as inner exceptions. Test this actual finalization
helper with the real shared Synthetic session, not only callbacks in isolation.
Publication failure adds its fixed token before throwing the safe terminal error;
no retry and no fallback file. If all output channels fail or the host crashes,
delivery is not guaranteed; missing output is not success or durable evidence.
The state must survive generic shared catches. A failure token must never coexist
with a claimed successful entry result. Successful process exit still cannot accept
geometry. Shared deadline/overrun logic and provisional receipt semantics are unchanged.

## Verification and boundaries

RED→GREEN cases: simultaneous managed-read + sample-writer cleanup + process cleanup
failures retained at the final surface; constructor/open/flush failures with cleanup
failures; stream release after writer failure; one-time ownership on success, partial
acquisition and repeated Dispose; original Process/SafeHandle reference equality;
unknown shared validation failure explicitly unclassified; publication failure;
private payload absence; bounded state and per-run isolation. Retain default-inert,
gate/collision and Unreviewed/NotAccepted tests. Run new and all six existing synthetic
suites, plus unchanged-source/hash checks. No real CIM/native probes or app launch.

Only new entry/tests and necessary documentation may change. Freeze old entries,
shared modules, App source, hashes, arguments and observation limits. No build,
restore, install, main merge or acceptance promotion. Live allowance remains0.
Important1 remains open until implementation and regression evidence satisfy this
contract. After written approval, prepare exact implementation steps/tests; approval
of this design does not itself authorize a live run.
