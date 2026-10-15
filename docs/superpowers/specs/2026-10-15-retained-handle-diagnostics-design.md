# Retained-handle lifetime diagnostics — proposed design

Date:2026-10-15. Base:26bb962a1c7c17aebe12f3e0b9774212f9682d92.
Status:written design for user review; NOT approved for planning, implementation,
live execution, matrix replay, acceptance or main merge.

## 1. Intent and evidence boundary

The user approved writing option1: an explicitly enabled diagnostic mode that
retains handles acquired and verified against the runner's own job before root
exit. The goal is better same-process lifetime evidence while preserving current
rights, budgets, fail-closed completion gates and existing default behavior.

Matrix003 stopped2PASS/1FAIL/23UNRUN because check03's listed member OpenProcess
failed87 before ownership/state evidence. A0.1961ms list-to-open interval does not
prove the member exited. Empty-list successes and successful short-child waits
do not explain the failed acquisition. This design is NOT an error87 fix or a
guarantee of observing every short-lived member. Retention itself changes object
lifetime/timing; any future sample must explicitly disclose that observer effect.

## 2. Alternatives and selected scope

- Selected: bounded own-job polling plus a diagnostic-only retained-handle registry.
  Reuses existing membership and zero-wait state APIs without new privileges.
- Not selected: completion-port notifications as sole identity/exit authority.
  Ordinary messages may be omitted and their IDs may no longer identify a live
  process. Adding a port/consumer thread is outside this increment.
- Not selected: repeat matrix until PASS, treat87/empty final job as natural
  success, extend waits or increase rights. None supplies missing identity proof.

The default runner, current live-sampling mode, original shutdown inspection and
matrix invocation remain behaviorally unchanged. No production Go/SDK/test callback
change, dependency install, security setting change or real-process probe belongs
to implementation of this design.

## 3. Entry point and mode isolation

Add an explicit native `RunWithRetainedHandleDiagnostics` entry with the same
executable/arguments/environment/cwd/streams/limit inputs as Run. Refactor shared
mechanics internally only as necessary; existing Run overloads pass no registry
and preserve their present mode selection, sampling cadence and receipt semantics.
Do not enable the feature through ambient environment variables or an existing
boolean whose meaning changes. Do not combine it with captureLiveMembers or
captureShutdownDiagnostic; the new entry chooses its one diagnostic mode itself.

The new result marks both LifetimeDiagnosticOnly=true and ShutdownDiagnosticOnly=true,
without changing ShutdownSchema3 or its current policy to a different schema.
Assert-ProcessCompletion must reject either diagnostic flag before success checks.
Legacy entries do not set either flag merely because the registry code exists.
The PowerShell serializer preserves optional lifetime fields; old receipt support
must not infer registry evidence or upgrade acceptance. Matrix scripts do not call
the new entry. A future live wrapper requires separately reviewed bounded scope.

## 4. Ownership and storage responsibilities

A run-local registry owns admitted child handles. It has no static/global cache,
background thread, completion port, archived-PID lookup or cross-run recovery.
Root remains represented by its exact CreateProcess handle, not a second registry
admission. Only IDs returned by the current own-job list are candidates.

Admission uses the existing0x101000 rights (limited query plus synchronize),
inherit=false. Validate positive UInt32/non-root ID; OpenProcess once; capture
last-error immediately; verify the exact handle using IsProcessInJob against the
current job. Admit only successful owned handles. Failed open, failed membership
and successful nonmembership are distinct records; none receives an owned token.
Close every opened-but-unadmitted handle. Transfer an admitted handle to registry
ownership exactly once; temporary admission code must not also close it.

Each admission has a run-local monotonic EntryId and the observed numeric PID.
Only EntryId plus its retained handle is an identity token. Later observations
never reopen that PID, replace a stored handle, merge independent records by PID,
or attach an initial shutdown failure to a registry row using PID equality alone.
Seeing an already represented numeric PID does not create another handle; any
inconsistent evidence is labelled ambiguous, not silently merged or accepted.
Do not claim complete process coverage or guaranteed PID-reuse detection.

## 5. Bounds and observation flow

Fixed limits:32 admission attempts per run, at most32 retained child handles,
one admission attempt per candidate numeric PID, no retry after failure, and no
eviction/reopening. Attempts count failures as well as successes. When exhausted,
stop admission polling and mark AdmissionBudgetExhausted; do not imply an exact
count of unobserved members. Previously retained handles remain until teardown.

While the root is being waited on, this explicit mode polls the own job at the
existing live-sampling cadence (root wait slices up to10ms), charging all work to
the existing root runtime limit. Check remaining time before every diagnostic
query/admission; stop new diagnostic work when exhausted. Retained/failed IDs are
not retried. Never start a process, sleep extra, wait on a child or reset a clock
to obtain better evidence. Cadence is not a guaranteed scheduler deadline.

At admission, record zero-wait handle state and exit code where available. Image
lookup is optional metadata; image failure is not ownership or exit evidence.
At root-wait completion, start the existing shared shutdown clock as today. Take
one zero-wait snapshot of admitted handles, charged to that clock, then continue
the original accounting/list/InspectMembers and completion logic without using
registry results as replacements. Root wait can end by timeout: record that fact,
never label this snapshot unconditionally as root-exit evidence.

After original natural/forced cleanup, take at most one final zero-wait snapshot
per retained entry using only remaining shared5s shutdown time. Record whether
termination had already been requested. An exit observed after forced cleanup
must not be described as natural. No registry snapshot waits for a process.
Skip remaining snapshots with a budget-exhausted reason when no time remains;
include diagnostic work in actual elapsed-time reporting. The existing500ms
natural window and shared5s shutdown ceiling are not enlarged. Native calls are
not preemptible; check budgets between calls, report any overrun honestly, and
never claim a hard OS scheduling/call-duration guarantee.

Store at most32 candidate admission records and two later snapshots per admitted
entry. Do not retain an unbounded per-poll log: use bounded first/last query facts,
query totals and explicit cap/budget/error flags. These are not a full event log.

## 6. Receipt contract and failure handling

Add a separate optional `LifetimeDiagnostics` object with diagnostic schema1:

- Mode marker, cap/attempt/admitted/query counts, root-wait outcome and budget/
  truncation/error flags. Coverage is always labelled sampled, never complete.
- Candidate EntryId, numeric PID, source list-query identity and monotonic/UTC API
  bounds, original open and membership outcomes/errors; no serialized handle value.
- For admitted entries, admission/root-wait-end/final observations: same-entry
  marker, zero-wait status, exit-code availability/value, image availability/error,
  monotonic/UTC bounds, skip reason and termination-requested-at-observation flag.
- Diagnostic collection errors and teardown outcome, separate from original core
  receipts. Null/unattempted/default values are never positive evidence.

Original MemberEvidence/ShutdownMemberEvidence and API failures remain unchanged;
the registry is neither their replacement nor an automatic join by PID. A consumer
may compare samples descriptively but cannot say a failed-open row was the same
object as a retained entry without additional identity evidence. This increment
does not introduce such a cross-path identity proof or acceptance rule.

A diagnostic acquisition/read/query error marks diagnostics incomplete and stops
further admission work; it must not bypass or interrupt original owned-job cleanup.
Do not swallow a core launch/cleanup failure. If Run throws before a native result
can be returned, partial in-memory registry records need not be persisted in this
increment: existing launch-error behavior remains and this limitation is explicit.
No success claim may be based on missing partial data.

Registry teardown belongs to the outer finally and must run after original job
cleanup/job-handle closure, even on launch/admission/read/serialization exceptions.
Every admitted handle gets one release attempt via an owning disposable wrapper;
temporary failed admissions own their release attempts. Disposal is idempotent and
attempts every entry even if another disposal reports an error. Close failures
are diagnostic errors, never permission to retry a raw numeric handle that may
have been reused. Teardown precedes successful native return/serialization so a
returned receipt can state whether all release attempts completed. Release-call
success is not equivalent to proving no other OS/user handles exist.

## 7. Planned verification, not executed here

The later implementation plan must use RED/GREEN on process-free deterministic
tests with an injected native-adapter boundary. Fake IDs/handles must never reach
real Windows APIs. Pure tests need no private SDK, cache, receipts, network,
process creation or historical PID lookup. Required coverage:

1. Old entries leave registry disabled and preserve existing modes/cadence/output.
2. Explicit diagnostics set both diagnostic flags and cannot pass completion gates,
   even if root/core cleanup and all registry observations appear successful.
3. Open failure87 retains the original API error; membership is not called.
   Open success/membership failure/nonmembership release temporary handles once.
4. Owned admission transfers exactly one handle; snapshots use it without reopen.
   Duplicate numeric candidates, invalid/root IDs and ambiguous observations do
   not manufacture another identity or silently overwrite evidence.
5.32-attempt cap includes failed admissions;33rd candidate is not opened; no retry,
   eviction or unbounded logging. Existing members remain queryable by EntryId.
6. Root timeout, zero remaining runtime/shutdown budget, nonpreemptible-call overrun,
   image loss, missing exit-code read and final-snapshot failure remain explicit.
7. Forced-cleanup then signaled handle is labelled post-termination, never natural.
   Original failed-open/core evidence stays refused regardless of registry output.
8. All exception paths attempt disposal of owned/temporary handles exactly once; disposal
   continues after another release fails; no raw-handle retry/double close.
9. Serialization round-trips nested fields/caps without raw handles and preserves
   optional legacy fields; missing data never upgrades ownership/exit/acceptance.

Existing process-free regression suites must pass on PowerShell7 and Windows
PowerShell5.1. Do not invoke suites that create processes as an incidental part of
this increment. Real ownership, native adapter correctness and observer effects
remain unverified until a separately approved controlled live experiment. Passing
fake-adapter tests is not proof that error87 is fixed or that the matrix passes.

## 8. Files, recovery and approval gates

Expected later implementation surface: process-native.cs, a small dedicated
registry/native-adapter source if needed for separation, run.ps1 optional field
serialization/diagnostic refusal, dedicated process-free tests and documentation.
No matrix-driver enablement, Go code, SDK patch or experiment runner is included.
Keep registry ownership testable without loading a real-process harness.

This written spec must be reviewed/approved by the user before writing-plans is
invoked. The resulting written plan then needs approval and an execution-method
choice before implementation. Live capture/matrix execution remains a separate
decision after implementation verification. Do not collapse those approvals.

Save spec/status/index on the existing development branch under AGENTS.md's
ordinary checkpoint authorization. Fresh-machine recovery starts there and reads
docs/runbooks/repository-recovery.md. Only portable source/tests/docs are committed;
SDKs, receipts, caches, binaries, credentials and raw OS handle values are not.
Historical matrix003 and broader C11/C12/Task10/B01/production/main gates stay open.

Author self-review: no placeholders or broken local document links; checked mode
isolation,32-attempt/storage bounds, clock charging, ownership transfer/disposal,
sample incompleteness and approval-stage consistency. Close failure language
describes release attempts, not guaranteed OS release. Ten reviewed tooling pins
remain unchanged. No independent implementation review or test execution is claimed.

## References

- [Saved acquisition analysis](../../roadmap/2026-10-15-matrix003-acquisition-analysis.md).
- [Process handles and identifiers](https://learn.microsoft.com/en-us/windows/win32/procthread/process-handles-and-identifiers): retained handles and numeric IDs have different lifetime contracts.
- [OpenProcess](https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-openprocess): requested access, returned handle and failure semantics.
- [Job completion-port association](https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_associate_completion_port): notification delivery and PID-lifetime limitations.
