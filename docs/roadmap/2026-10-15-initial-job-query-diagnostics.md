# Initial job-query diagnostics — 2026-10-15

## Approved bounded increment

The user approved the short design: record only the initial accounting and
process-ID-list query boundaries/raw results; keep every acceptance predicate,
query count, retry bound, right and timeout unchanged; test using an empty owned
job and invalid job handles without launching child processes. No live-child
probe, Go/fuzz/matrix retry, archived-PID query, SDK write or main merge.

`process-native.cs` annotates the existing first shutdown snapshot after the
root-wait loop. This also occurs when that loop exits on timeout; metadata alone
does not establish RootExited=true. No separate query path or extra API calls are
introduced. Subsequent live/natural/forced-cleanup queries remain unrecorded.

## Data and persistence

- InitialAccountingQuery: API, UTC/monotonic start/end/frequency, success/error,
  ValuesRead, total/active/terminated process counts.
- InitialMemberQueries: one row per actual existing bounded enumeration attempt,
  API and clocks, success/error, attempt/capacity/buffer bytes, HeaderRead and
  assigned/returned native counts. Maximum8 attempts, existing capacity/bounds.
- Native error is read immediately before timing/diagnostic bookkeeping. Success
  records error0 rather than stale last-error. Fatal invalid-handle queries throw
  unchanged and do not interpret invalid output as valid zero counts. The existing
  ERROR_MORE_DATA/header/capacity-growth policy is unchanged, not a new retry.

The normal receipt mapping is extracted unchanged into
ConvertTo-RecordedProcessReceipt and extended with the two optional fields;
Invoke-RecordedProcess uses that converter. Matrix's existing all-field adapter
already carries these fields without modification. JSON depth6 retains this
shape. The schema3/v2 policy and validators do not consult metadata to grant exit,
ownership, cleanup, or natural-completion success. Legacy receipts may omit it.

Native initial-query exceptions still abort Run: observations exist in the passed
Result during helper execution, but the caller's launch-error.json does not retain
that partial result. Invalid-handle tests verify in-memory evidence and original
exceptions, not durable failed-launch receipt persistence. Successful initial
queries followed by member-open refusal can return the ordinary populated Result.
No such live invocation is made in this increment.

Instrumentation adds clock/object/serialization overhead inside existing clocks;
it does not reset budgets. The two queries are ordered, not an atomic snapshot.
This data cannot retroactively identify matrix002's unknown member or prove its
OS/process cause. No change in a previous immutable receipt is inferred.

## Tests and independent review

Initial RED: the new test failed at missing ReadInitialJobQueries diagnostic path.
GREEN:19 new assertions pass on PowerShell7.6.5 and Windows PowerShell5.1. One
test invocation initially exposed PowerShell PSObject wrapping during reflection;
passing BaseObject fixed only the fixture argument, not native logic or gates.

Real empty job queries test zero native counts, one enumeration attempt, ordered
clocks, raw buffer/header facts and preserved return values. Invalid job handles
exercise real error6 with no fabricated values, no enumeration after accounting
failure, and no retry after fatal member-query failure. Actual receipt conversion
and depth6 roundtrip retain metadata/context. Synthetic complete and incomplete
receipts verify legacy acceptance compatibility and rejection of forced cleanup.
No test assigns/starts a process or uses a historical PID.

Existing process-free adapter/driver regressions also completed:214 assertions
per host plus19 new =233 PASS on each host (466 total). Counts were independently
rechecked from16 saved suite logs; no suite failure remains in this scope.

| Suite | PS7.6.5 PASS | PS5.1 PASS |
| --- | --- | --- |
| initial-job-query |19 |19 |
| matrix-check (synthetic only) |60 |60 |
| pair-sample (synthetic only) |45 |45 |
| envelope-uninstrumented |28 |28 |
| envelope-sample |22 |22 |
| envelope-compile |19 |19 |
| envelope-preparation |19 |19 |
| envelope-observe |21 |21 |

Local ignored logs are under
`.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/initial-job-query-validation-001`.
Its validation-manifest.json binds all16 logs to SHA256 values and counts; manifest
SHA256 `CBA558B51B460F0A1183FCD9BC00E5BA78E98290116DF5CC58DE2C3DFA8DF82F`.
The701 previously recorded product source entries also retain their matrix002
post-run hashes. Changes are diagnostic tooling/tests and progress docs only.
Full runner suites requiring child processes and the product matrix are deliberately
UNRUN under the approved scope; scoped checks are not full-suite acceptance.

Independent read-only review found no Critical/Important issue. Its two Minor
documentation qualifications (timeout-loop snapshots and lack of durable partial
results on query exceptions) are recorded above. Reviewer did not run tests or
judge live timing/membership/capacity growth/cleanup, actual matrix002 acceptance,
main/deployment/fuzz readiness, or these not-yet-written recovery docs. Those
exclusions are retained as limits, not passing results. No collector redesign.
The same reviewer subsequently read this report/current-status/README and closed
both documentation follow-ups with no new blocker. It ran no tests in that pass.

## Recovery and next boundary

Source/tests and this report are saved on codex/c12-b01-task9-coordinator, not main.
Read current-status.md and docs/runbooks/repository-recovery.md. The new test uses
only the current Windows PowerShell/.NET host and its own empty job object; it
needs no copied SDK/cache, private receipts or previous conversation. Run
docs/diagnostics/isolated-fuzz/initial-job-query.tests.ps1 in PowerShell. Cloning
does not restore ignored raw artifacts or dependencies for actual product tests.

Matrix002 stays0PASS/1FAIL/25UNRUN. This is diagnostic readiness only, not a fix or
authority to rerun it. A controlled short-lived-process observation would require
a separately approved scope/fresh identity; no live capture is authorized here.
All broader historical/product/C11/C12/Task10/B01/production/main gates remain open.
