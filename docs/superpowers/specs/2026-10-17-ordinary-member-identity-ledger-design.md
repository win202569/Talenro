# Ordinary member identity ledger — proposed written design

Date: 2026-10-17. Base: 5f5da3a0d5ebf5f4f43c4cd292a699a26d459eb5.
Status: conversational direction and boundaries approved; written spec awaiting
user review. This is design only, not implementation or execution authority.

## 1. Purpose and evidence boundary

The user selected option 1: one ordinary-observer identity ledger that acquires
job-member handles earlier and uses those exact verified handles through final
exit checks. The goal is continuity of identity, not suppressing Error87 or
declaring the current metadata RED accepted. Unknown evidence must still refuse.

The [saved comparison](../../roadmap/2026-10-17-metadata-runtime-retained-probe-001.md#authorized-readonly-timeline-comparison)
found ordinary acquisition failed around57ms after resume, while the separate
diagnostic acquired a member around21ms and observed it exited around41ms.
Different modes, timing and retained handles prevent causal attribution. The
original PID16616 remains unidentified. The new sample is not an accepted RED,
an A/B result, proof of complete coverage or a demonstrated Error87 repair.

Success for the later implementation means a separately versioned, testable
identity/evidence path with conservative failure handling and unchanged bounds.
It does not mean ordinary callers are switched or a live workload is accepted.

## 2. Alternatives and scope

- Selected: a single run-local ledger owns acquisition and final observation.
  It avoids joining an early diagnostic row with a later core row by numeric PID.
- Not selected: add more samples to the existing diagnostic registry and use
  those rows to override ordinary failed-open evidence. That leaves two identity
  authorities and contradicts diagnostic-only mode isolation.
- Not selected: a new event-notification subsystem. It expands the change and
  still needs explicit identity, missing-event and teardown contracts.

No new rights, retries, larger waits, executable whitelist, SDK change, Go
metadata implementation, dependency installation or security-setting change.
Existing diagnostic entries and ordinary Run overloads retain their behavior.
No matrix/wrapper enables this path automatically. Historical receipts stay frozen.

## 3. Components and explicit entry

Introduce a separate native entry named RunWithOrdinaryIdentityLedger with the
existing executable/arguments/environment/cwd/streams/limit input shape. It is
not an ambient environment flag and cannot combine with live-sampling,
retained-diagnostic or compile-stage modes. Share launch/cleanup mechanics only
where their existing behavior can be preserved and tested.

Separate responsibilities:

1. A bounded ledger owns candidate records and admitted handles. Its operations
   are Poll, ObserveFinal and Dispose, through injected native/clock interfaces.
2. The native adapter performs current-job queries, limited-rights open,
   exact-handle ownership checks, wait/exit reads and release. Fakes must not
   call Windows APIs; no archived-PID lookup is part of either adapter.
3. A pure evidence evaluator checks the new receipt contract. It does not launch
   anything or treat an empty final list as a substitute for identity evidence.
4. The runner owns root creation, assignment, clocks and fail-safe job cleanup.
   The root is represented by its original creation handle, not a second open.

The existing diagnostic registry may supply patterns, not runtime records or an
acceptance shortcut. Do not silently change its sampled/diagnostic semantics.

## 4. Identity and ownership

After successful assignment and resume, poll the current job before each root
wait slice, using up to10ms slices in this new mode only. Every non-root positive
UInt32 candidate receives at most one acquisition attempt during the whole run.
Use existing0x101000 rights,inherit=false; capture last-error immediately.
Verify IsProcessInJob on the returned handle against this exact job before
transferring the handle into ledger ownership. Failed and foreign admissions
have no owned identity and their temporary handle gets one release attempt.

The runner creates a fresh RunIdentity before launch and binds the ledger to
that run, root creation and current job. Each admitted entry carries RunIdentity,
a monotonic run-local EntryId and one owning handle.
All later state/exit reads use this entry directly, never OpenProcess again.
PID indexes discover candidates only; they cannot repair failed admissions,
replace a handle, attach independent rows, or establish cross-run identity.
Retain admitted handles until their final observation/teardown; no eviction.
Repeated discovery of an already represented PID does not create a new entry.
Any inconsistency that cannot be resolved through the existing retained entry
is sticky ambiguity and refusal, not a new successful identity.

State is unknown, running or exited. An exited observation needs a signaled
exact owned handle and a successful exit-code read from that handle. Exit code0
alone is not exit proof; running code259 is not exit proof either. A member's
nonzero exit code is preserved; workload-specific output/root-exit predicates
remain separate from the process-group completion decision.

Image path is optional metadata. Image failure cannot establish ownership or
exit, but does not erase an independently verified signaled/exit-code result.
No conhost-specific exception exists.

## 5. Sampling, bounds and conservative reconciliation

The ledger remains sampled, not a complete event recorder. A successful query
must have valid counts, complete bounded output and successful native status.
Failed/truncated queries, invalid candidates or impossible counter order mark
evidence incomplete. Store bounded first/last query facts and counters, not an
unbounded log. Maximum32 non-root acquisition attempts/records/retained handles;
failures consume attempts. Reaching32 marks cap-exhausted and refuses, even if
the observed batch happened to end there. No retry or fresh budget.

Required coverage checks include every candidate discovered in runtime and
shutdown lists, all admitted entries (including ones absent at root exit), and
fresh final job accounting. Require final TotalProcesses = 1 + Admitted, with
successful root assignment, no invalid/unowned/failed candidate and no query,
cap or ambiguity failure. Preserve the raw count and the reconciliation result.
A mismatch refuses; it does not manufacture identities for missing members.
This is a necessary conservative check, not proof of a complete event trace.
Job accounting can count association-limit failures; false refusals in such
cases are intentional. Do not label sampled coverage complete when counts match.

All runtime polling/acquisition charges the original runtime clock. At root-wait
completion, start one shutdown clock; all final acquisition, observation,
release and final job queries charge that same clock. The natural window remains
500ms and total cleanup ceiling5000ms. Check remaining budget between calls;
record nonpreemptible-call overruns instead of promising OS hard deadlines.

## 6. Final observation and failure flow

At root-wait end, use retained entries directly and admit newly discovered job
members under the same limits. Acquire a bounded batch before waiting on a member.
Within the remaining shared500ms, verify final signaled/exit-code state for every
admitted member; preserve initial snapshots. No per-member fresh500ms window.
Any new member discovered during shutdown is subject to the same acquisition and
final-exit requirements. A disappearance before acquisition remains a failure.

When all required states are verified, stop admission, release child handles
through their owners, and perform final accounting/list checks within the same
natural window. This ordering avoids assuming retained references cannot affect
job accounting. Require successful releases, active count0, empty final list,
no limit-terminated processes and the count reconciliation in section5.
A new/unrepresented member at this final check refuses; do not reopen a released
entry or start a second acquisition pass to obtain a better result.

Natural completion additionally requires root exited, no runtime timeout, no
termination requested and no sticky evidence failure. Once any acquisition,
ownership, required read, budget, count or release failure occurs, later empty
lists or successful reads cannot erase it. Optional image errors stay separate.
Stop optional evidence work after fatal incompleteness, but always execute the
existing owned-job cleanup under the remaining shared5s; retain raw errors.
Forced termination followed by a signaled handle is post-termination evidence,
never natural success. Cleanup success and workload acceptance remain distinct.

Outer finally attempts release of every still-owned child once, even if another
release or core cleanup fails. Disposal is idempotent; do not retry numeric raw
handles after release failure. Preserve original launch/cleanup exceptions and
report disposal errors without replacing the primary failure. Root/job/stream
resource cleanup retains its existing responsibilities. Missing partial receipts
after an exception cannot be treated as success.

## 7. New receipt and rollout boundary

Use ShutdownSchema4, policy ordinary-owned-ledger-final-exit-v1, and a separate
MemberIdentityLedger object with Schema1, Mode=ordinary-identity-ledger-v1,
Coverage=sampled. This is a new contract, not schema3 data with flags removed.
Keep diagnostic entry flags/receipt contracts unchanged and always rejected.
The new mode explicitly records both diagnostic flags as false and contains no
LifetimeDiagnostics, CompileStageTrigger or diagnostic shutdown observations.
Absence of diagnostic flags is not sufficient to qualify as the new mode.

The new object records caps, attempts, admissions, query counts, first/last query
evidence, sticky error/ambiguity/budget flags, final accounting/reconciliation,
teardown attempts/failures and a bounded Entries array. Ledger and entries bind
to the receipt's RunIdentity. Entry records contain EntryId/PID, source-query
identity/bounds, acquisition/membership outcomes,
initial and final reads, termination-at-observation and release status.
Final reads reference EntryId and SameHandleVerified, not PID equality alone.
Use monotonic bounds/frequency plus UTC metadata; never serialize raw handles.
Unattempted/missing/default values are not affirmative evidence.

Serialization/evaluation must reject missing fields, unknown versions/policy,
wrong types, duplicate entries, inconsistent counts/references, impossible timing,
diagnostic contamination and claims inconsistent with raw observations.
The schema4 evaluator requires its full envelope. The existing gate must reject
any new-mode/ledger marker under a missing or older shutdown schema, rather than
fall through to legacy acceptance. Tests remove/change version markers one at
a time and require refusal. This is contract validation, not cryptographic
authentication of arbitrarily rewritten JSON. Old receipts cannot be enriched
after the fact or submitted as new-mode results.

The first implementation scope exposes a separate pure evaluator for schema4;
Assert-ProcessCompletion continues rejecting schema4 and all diagnostics.
Existing callers/drivers do not switch. Activating schema4 as an ordinary
acceptance path is a later explicit rollout decision after scoped verification
and separately authorized live evidence. An evaluator result is not approval of
a workload or permission to resume metadata001. No accepted runtime is created
by changing a flag, copying a diagnostic object or replaying a consumed identity.

## 8. Verification requirements for a future plan

No tests ran for this document. Later implementation starts with deterministic
RED/GREEN tests using injected fake handles, job queries and monotonic clocks.
They require no native workload, SDK/cache, network or historical receipt.
Required cases:

1. Existing Run modes/cadence/contracts unchanged; new entry explicit and isolated.
2. Early owned admission followed by absent-at-root member still needs that
   entry's final exit; empty job with a running/unknown entry refuses.
3. Error87 preserved before membership/state reads; no retry or late override.
   Failed ownership/nonmembership releases the temporary handle and refuses.
4. Duplicate candidates never reopen; ambiguous identity, independent PID-only
   row substitution and cross-run entries refuse. Root is not admitted twice.
5. All sampled/late candidates accounted; missed-member count mismatch refuses;
   matching counts cannot erase failed queries or assert full event coverage.
6. Cap32 includes failures; no33rd open, eviction, budget reset or unbounded log.
7. Shared natural/cleanup/runtime budgets include collection and release work;
   expiry, clock errors and native-call overrun refuse without skipping cleanup.
8. Exit-code0 without signaled state refuses; image loss alone cannot pass or
   invalidate a separately valid exit; missing required exit read refuses.
9. Forced cleanup, timeout, release failure and new member at final checks refuse;
   release exceptions attempt remaining entries once and preserve the core error.
10. Receipt round trips retain nested evidence; missing/ill-typed/unknown/downgraded
    schema or contradictory counts/timings refuse; legacy and diagnostics remain
    isolated; Assert-ProcessCompletion rejects even a fully populated schema4.

The plan must list exact process-free suites and hosts (PowerShell7 and Windows
PowerShell5.1), test/build boundaries and costs. Suites that create real processes
are not incidental fake regression. Native integration, scheduling/handle effects,
reconciliation behavior and real Error87 recurrence remain unverified until a
separately approved bounded experiment. No matrix replay or metadata run is
authorized by fake tests, design approval or plan preparation.

## 9. Files, recovery and approval gates

Expected later code surface: process-native.cs entry/wiring, a dedicated ordinary
ledger/adapter/evaluator source, explicit run.ps1 serialization/refusal handling,
preflight source inventory, dedicated fake fixtures/tests and documentation.
Avoid unrelated refactoring or modifying the diagnostic registry's contract.
No implementation plan has been created by this document.

Next gate is user review of this written spec. Approval permits writing-plans
only. A written implementation plan then requires user approval and execution
method selection. Live execution and acceptance activation remain separate gates.
No test execution budget is granted or consumed here. Existing metadata usage
remains1 build/1 runtime; its RED is unaccepted. Independent retained probe usage
remains0 builds/1 runtime, consumed. Preserve both identities; no replay.

Commit portable design/status/recovery docs on codex/c12-b01-task9-coordinator per
AGENTS.md. Raw receipts, binaries, SDKs/caches and credentials stay ignored. New
machines read current status and repository-recovery.md; Git does not restore raw
evidence or prior process state. Main/production/C12 and metadata Task1/2 gates stay
open. This spec supersedes no historical failure or accepted diagnostic design.

Author self-review: checked scope, identity ownership, bounded storage/clocks,
sticky refusal, sampled/count limitations, release-before-final-query ordering,
version isolation and approval stages. No independent review/test success claimed.

## References

- [Retained diagnostic design](2026-10-15-retained-handle-diagnostics-design.md).
- [Current project status](../../roadmap/current-status.md).
- [Recovery](../../runbooks/repository-recovery.md).
- [Process handles](https://learn.microsoft.com/en-us/windows/win32/procthread/process-handles-and-identifiers): handle validity is separate from numeric discovery.
- [Process information](https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/ns-processthreadsapi-process_information): object/identifier lifetime and handle release.
- [Job accounting](https://learn.microsoft.com/windows/win32/api/winnt/ns-winnt-jobobject_basic_accounting_information): cumulative and active counts, including special association-limit cases.
