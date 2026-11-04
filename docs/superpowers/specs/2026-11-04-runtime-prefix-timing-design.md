# Bounded runtime-prefix timing diagnostic

Status: DRAFT / PENDING_DESIGN_APPROVAL / NO_EXECUTION_AUTHORITY.

The user selected option1 after the read-only timeout investigation: measure
historical loading, context construction and the first assertion, not a full host
rerun. This document proposes the bounds; selection did not approve execution.

## Intent and approaches

Locate the last completed stage and measured cost in the silent interval after
ContractAPIAvailable. Existing logs do not distinguish preparation from the first
assertion; repeated work counts are hypotheses, not a proven timeout cause.

Selected: one isolated prefix diagnostic. It is smaller and stops immediately
after the real BaselineCore assertion, whether accepted or rejected. It cannot
establish equivalence to full-host timing or prove52group acceptance.

Alternative: instrument and rerun the whole host. More representative, but runs
unrelated later cases and consumes a larger execution scope; not selected.
Another static review cannot recover timestamps that were never recorded.

## Exact scope

Use the existing PowerShell runtime and original frozen input identities. Read
manifest186FE2C2B44A31A98EA5406109AC888A625FD1B0061C5338D0382268015D9BBA,
the three literal historical evidence files, and the same historical source
set used by the current suite. Retain their original ordering and all pin/hash
checks. No caching, algorithm optimization, skipped verification or reordering.

Measure manifest hash/read/JSON separately; each evidence load; each of the three
historical contexts; pin lookup/hash and source parse/model construction within
each context; first baseline data-copy and actual Assert-ODProfileObservations.
No subsequent mutation, report, unit, recorder or runtime test groups execute.

All31operational sources, existing tests, inputs, snapshots, receipts and15stops
remain byte-identical. Create only separately named diagnostic helper files and
new evidence. Never invoke the stopped driver/capture/suite or a historical grant.
Helpers extract only explicitly allowlisted definitions from hash-bound sources;
do not dot-source the full operational entry points. Instrumented definitions
must differ only by bounded timing events; an implementation plan must enumerate
the allowlist, transformations and checks before any execution can be authorized.
Do not substitute mocks, reconstructed success data or weaker assertions.

## Proposed finite budget and ownership

Current counters26aux/5hosts/3compiles/217snapshots. Propose exactly one new
diagnostic computation27 and at most one owned diagnostic PowerShell process,
counted conservatively as host6 (not a product acceptance host). No additional
test runs, target children, compilations, snapshots, independent reviewers,
network operations, installations or security-setting changes. Parse/static
inspection and artifact byte verification do not execute the diagnostic body.

Use fixed fresh namespace runtime-prefix-timing-v1. Exclusive CreateNew intent
consumes the attempt before validation/start; existing intent, stop or partial
output forbids reuse. Failed setup also consumes it; no automatic retry.
The parent retains only its invocation-local Process object, starts hidden with
redirected streams, and never adopts a PID or kills a process tree.

Execution cap120000ms begins before stream/process setup. Keep separate2000ms
owned-exit confirmation and2000ms stream-drain caps; these are cleanup overhead,
not extra diagnostic execution time. Terminate only that owned process on timeout.
The parent preflight and final archival duration are recorded separately and
must not be reported as diagnostic execution duration.

## Evidence, integrity and stopping

Monotonic elapsed timing, ordered sequence IDs, stage identifiers and begin/end
events must be flushed to a fresh journal. Bound journal to1MiB/4096events and
cap captured streams; overflow or recording failure stops, never drops evidence
silently or permits unbounded output. Record clock/flush overhead limitations.
No secrets or complete input payloads in timing events. Exact event schema,
output bounds and parent recording mechanics belong in the implementation plan.

Preserve actual invocation, initial/poll/final outputs, partial journal, process
status, exit/timeout/drain outcomes, input hashes before/after and stop reason.
An unmatched begin identifies an unfinished interval, not proof of a specific
CPU/I/O cause. Nested timings overlap; never add them as independent costs.
Completion within the cap only describes this instrumented prefix. Rejection,
timeout, drift or missing evidence is not a pass and cannot authorize a repair.

Stop after this one attempt even if successful. Archive/decode-verify all evidence,
update status/recovery and push an ordinary development-branch checkpoint.
Keep the old stopped scope closed and all prior raw records immutable.

## Approval boundary

This is a small diagnostic design, not an optimization or recovery implementation.
Design approval permits writing a concrete implementation plan only. That plan
must pin commands/files, helper verification, timing schema and exclusive budget;
execution requires approval of that plan. No new allowance has been consumed.
No diagnostic result constitutes product acceptance; overall NOT_ACCEPTED.
