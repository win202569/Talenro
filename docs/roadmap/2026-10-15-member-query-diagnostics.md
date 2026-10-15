# Member-acquisition API diagnostics — 2026-10-15

## Approved scope and implementation

The user approved recording the exact API, before/after timing and original error,
with isolated tests. No process rights, ownership/exit requirement,500ms natural
window,5000ms cleanup window, acceptance predicate or cleanup strategy changed.
No Go, Envelope build/fuzz, Pair, full product matrix or preparation retry ran.

`process-native.cs` adds optional `ProcessOpen` and `JobMembership` records:
API name, UTC start/end, monotonic start/end timestamps/frequency, success and raw
error. Last-error is captured immediately after failure, before clocks/bookkeeping,
and propagated through the existing QueryError/catch path. Success records error0,
not a stale last-error. API success is distinct from actual membership. Failed open
leaves the membership record null/unattempted. These cover initial acquisition
only, not every later state/image/confirmation API.

Rights stay QUERY_LIMITED_INFORMATION|SYNCHRONIZE. API order, acquired-handle
lifetime/closure, enumeration bounds and all policy checks remain unchanged.
Recording overhead stays within existing outer clocks; no budget reset/extension.
The optional records fit the existing receipt JSON depth6/schema3/v2. Acceptance
does not consult them to forgive unknown ownership/state or forced cleanup.

## Test-first and compatibility evidence

New portable `member-query.tests.ps1` creates a live PowerShell child, retains its
creation handle, assigns its job, and uses another empty job plus an invalid job
handle. It invokes the real private inspection path by reflection, without test-only
production APIs, mocked native calls, arbitrary/archived PID queries or ACL changes.
Before implementation it preserved native error6 but failed the missing-phase
assertion: genuine RED. Afterward all12 assertions passed on PowerShell7.6.5 and
Windows PowerShell5.1. An initial5.1 test-only `[ulong]` alias error was corrected
to `[UInt64]`; the final test passed freshly on both runtimes.

Coverage: real handle-open success followed by ownership-query failure, original
error, UTC/monotonic timing/order, correct membership, API-success/nonmembership,
JSON roundtrip and unchanged native rejection. Readonly inspection also verified
both records in a saved native-producer receipt and rejection of the unchanged old
origin87 receipt. No old record was rewritten; that87 remains unattributed.

## Related regression — not an overall pass

| Script | Fresh result |
| --- | --- |
| member-query | 12 PASS on each runtime |
| final-exit | 52 PASS |
| shutdown-observation | 33 PASS |
| shutdown-state | 9 PASS |
| run | 14 PASS; FAIL `real owned descendant cleaned on timeout`; remaining16 checks UNRUN |
| runner-regression | 21 PASS |
| materialize | 8 PASS |
| provenance | 8 PASS |
| preflight | 4 PASS |
| isolation | 18 PASS, including actual junction/hardlink rejection |
| envelope-observe planning/provenance | 21 PASS, no Go launch |

The aggregate command aborted at run.tests.ps1. Remaining independent scripts were
run separately, without retrying that failed fixture or increasing its2s deadline.
Existing elevated fixture permissions enabled real link tests; no install or
security change. Full product tests were outside the approved no-Go scope.

The failed tree-timeout receipt, UTC09:36:14.077–09:36:16.132
(Asia/Tbilisi13:36:14–13:36:16), records TotalProcesses1, TimedOuttrue,
RootExitedfalse, TerminationRequestedtrue, CleanupVerifiedtrue, final accounting0
and empty member list. Only TotalProcesses>=2 failed in the compound assertion:
the recorded job did not establish the intended descendant scenario. Cleanup was
verified; why the scenario was absent is unknown, not proof of startup latency,
scheduling or a relationship to the new metadata. Root acquisition APIs both
succeeded with error0. No timing-test retry or policy workaround was performed.

Ignored raw evidence is worktree-relative
`.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/runner-tests-5799af87146d4dea9e9ec55c154afeae/tree-timeout/receipt.json`,
SHA256 `F36BE13805AFCA6E29094D8B9F88DFCB4BAC90C2A067CE1866E9F6ECE306B283`.

## Independent review and coverage limits

An independent read-only review inspected the implementation/test delta and
surrounding validators, without rerunning tests or spawning further reviewers.
It found no blocking implementation defect and supported an unaccepted development
checkpoint, not main merge/product acceptance. Its nonblocking limits are retained:

- Native OpenProcess failure was not forcibly exercised; no arbitrary/archived PID
  probes or ACL changes were used to force it. Historical87 is not reproduced.
- Tests assert native rejection/row JSON, not a dedicated full-validator matrix
  with diagnostic metadata attached to failed and legacy synthetic receipts.
  Static validation remains unchanged; the actual old origin receipt still fails.
- New fixture cleanup requests termination and closes handles, but does not verify
  termination results or whole-job emptiness. A passing12-check test must not be
  described as independently verified fixture cleanup. This differs from the saved
  native-producer and failed timeout receipts, which do verify empty-job cleanup.

All three limits remain explicit; none grants permission to weaken a guard or
claims an original failure is fixed. No additional native probe was dispatched.

## Blocked state and recovery

The old origin87 and new descendant-fixture coverage failure remain open. This
increment does not reopen consumed preparation001, establish an Envelope result,
or clear matrix3/Pair/C11/C12/Task10/B01/production/main gates. Next work should
diagnose the fixture's missing descendant scenario or define a separately approved
probe; do not retry Go preparation or enlarge limits.

Necessary code/tests/reports are saved on the latest
`codex/c12-b01-task9-coordinator`; commit/push and remote SHA equality are verified
at handoff. Recover that branch and read current status/repository recovery.
The new12-check fixture uses current PowerShell and new process/job objects, not
old toolchains, caches or raw receipts. Cloning does not restore live/raw state.
Existing Envelope integration prerequisite tests still require separately prepared
verified copy/cache inputs; running their checks does not authorize Go execution.
