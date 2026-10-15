# Independent child-readiness/cleanup fixture — 2026-10-15

## Approved bounded design

The user approved an additive independent fixture, not replacement of the old
descendant test: a target child writes its unique nonce and own process ID, which
must match one non-root owned/running member record. Report readiness and cleanup
independently. Only matching readiness, timeout and verified empty cleanup can
pass this diagnostic fixture. Acceptance and OriginalTestPassed always stay false.

The native runner, run.tests.ps1, security policies and original whole-process2s
budget are unchanged. There is no extra startup period, timer reset or automatic
retry. The approved execution scope was one new actual test after helper checks;
no Go/fuzz, old probe replay, original-test acceptance or main merge.

## Implementation and test-first verification

New child-readiness.ps1 exposes a readonly planner, evidence evaluator and explicit
invoker. Only a nonexistent child-readiness-NNN root beneath the owned diagnostic
scope is allowed. Planning creates a nonce without creating a directory/process.
The hidden PowerShell child creates child-ready.json with CreateNew, flushes and
closes it, then sleeps20s; the parent also sleeps20s after its hidden child launch.
The unchanged runner's2s budget includes setup and startup, interrupting that work.

The evaluator refuses a wrong nonce, root/zero/unobserved ID, duplicate matching
members, truncated evidence, unowned/unknown/non-running member, query/exit-code
failure or a timestamp outside StartUTC..StartUTC+2s or after member CoreObservedUTC.
The readiness file is size-bounded and link-checked under the validated scope;
missing/partial records do not become readiness. Cleanup requires its boolean
proof, final accounting0 and an empty member list, independently of readiness.
Invocation saves plan/result/native receipts and refuses unsafe cleanup. Reusing
the consumed root is rejected before launch. No historical PID queries are used.

The new test suite first failed at the missing guarded planner assertion. After
implementation,28 helper assertions passed on PowerShell7 and Windows PowerShell5.1.
They exercise real file IO by running only the child's preparation/write prefix
in the existing test host, excluding the sleep and any child launch. Overwriting
that file fails. Literal synthetic receipts test the real evidence evaluator;
their positive case is not a claim about an actual native child or product test.
Wrong identities, timestamps, ownership/state/query/exit evidence, partial/missing
records and invalid timeout/cleanup/truncation results are covered.

The existing full-parent12, tree-stage9 and minimal-entry12 helper checks also
passed separately on both hosts:61 checks per host including the new28. No full
native runner/product suite was run in this scope; the historical descendant
regression and its16 unrun checks remain open, not silently retried or passed.

## Single actual execution — FAILED readiness, verified cleanup

child-readiness-001 ran ONCE at UTC10:14:00.5965486–10:14:02.6365229
(Asia/Tbilisi14:14). The test invocation returned exit1 because FixturePassed=false.

- ReadinessRecordObserved=false, ChildReady=false: child-ready.json absent.
- TimedOut=true, root not exited; native TotalProcesses3. Saved member evidence
  has three owned/running records, two powershell.exe images and one conhost.exe
  image, with QueryError0/read exit259. This does not identify which script stage
  the intended child reached or substitute for its missing nonce/ID record.
- TerminationRequested=true, NaturalShutdownVerified=false. CleanupVerified=true,
  ActiveAfterCleanup0, MembersAfterCleanup empty; ShutdownElapsedMs30.
- Stdout empty; stderr11 bytes, only '#< CLIXML', not a causal exception.
- FixturePassed=false, Acceptance=false, OriginalTestPassed=false. No retry.

Raw ignored worktree-relative evidence:
.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/child-readiness-001/.
SHA256 process/receipt.json:
E057193C8A1D396F82DA5B791EE2A1BD8E5CBA4094504701012D78E0B31BC6D7;
result.json:
56FA5DDAAC768B2F4571346ED052C144DFA91FA664868DBF90D0447E14F1ED78.

This proves empty owned-job cleanup for this sample, not target-child readiness.
Missing readiness could mean the child did not reach/complete the write before
termination; this observation does not distinguish startup, scheduling, conversion
or IO delays. It is not proof of deadlock, repaired old fixture or origin87 cause.

## Independent review and evidence limits

Independent read-only review before the actual run found no Critical/Important
defects under the diagnostic interpretation and supported the one authorized run
and an unfinished checkpoint only. The reviewer ran no tests/processes or edits.
Minor coverage gaps: duplicate member, marker UTC after CoreObservedUTC, malformed
field types, cleanup-refusal invocation, parent syntax and embedded child encoding
are not independently exercised. Guards were inspected statically, not all tested.

Child UTC is recorded before its file IO. Member evidence is collected after the
native timeout. Even a future ChildReady/FixturePassed=true would mean a correlated
script-entry record and owned-running snapshot, NOT proof that durable marker
flush completed before the monotonic2s deadline or the full20s sleep executed.
UTC interval checks assume comparable clocks; no exact monotonic readiness timing
was added. This actual record is missing, so no readiness claim is made at all.
The review's exclusions (strict timing/durable-ready proof, full sleep completion,
original repair/acceptance/main, runtime prediction and unauthorized extra runs)
are retained, not silently promoted to success.

## Recovery and remaining blockers

### Read-only event-log follow-up

No new fixture or process ran for this follow-up. Queried Windows PowerShell and
Microsoft-Windows-PowerShell/Operational only from one second before the saved
child-readiness001 StartUTC to one second after EndUTC. Ordinary access was
denied; the same bounded readonly query succeeded with approved sandbox escalation.
No logging/security configuration was changed. The classic log returned no
matching events; the operational log returned three, all matching the saved root
process ID. No matching intended-child startup event was found in this window.
This is not proof that its engine or script never started: event coverage/delivery
is not a complete execution trace. No historical process handles were reopened.

Operational event40961 described console startup at10:14:01.8665331Z
(record3367984), event53504 described IPC listener startup at10:14:01.9078139Z
(record3367985), and event40962 described console ready at10:14:01.9313207Z
(record3367986). Offsets from saved native StartUTC are1269.9845,1311.2653 and
1334.7721ms, respectively. These are local provider messages and UTC differences,
not proof of encoded-command/child readiness or monotonic phase measurements.

The same bounded root-ID-only operational query on two earlier saved samples
produced the following comparison; it did not replay those samples:

| Saved sample | Console-start event40961 offset | Console-ready event40962 offset | Record IDs (40961 / 53504 / 40962) |
| --- | ---: | ---: | --- |
| entry-startup-probe-001 | 86.3687ms | 151.8186ms | 3367906 / 3367907 / 3367908 |
| full-parent-probe-001 | 1323.9237ms | 1400.0098ms | 3367938 / 3367939 / 3367940 |
| child-readiness-001 | 1269.9845ms | 1334.7721ms | 3367984 / 3367985 / 3367986 |

Full-parent console-start/IPC/ready event timestamps are10:05:10.4162555Z /
10:05:10.4686522Z /10:05:10.4923416Z.
Minimal-entry counterparts are09:55:10.3898471Z /09:55:10.4310082Z /
09:55:10.4552970Z. The varying elapsed time is observational, not a controlled
payload-size, cold-start, scanning or scheduling experiment. Native setup/create/
resume phase times were not saved, so none of those causes can be isolated yet.
Original native/result hashes above were rechecked unchanged. Event record IDs
are local lookup aids, not portable event logs; raw system logs were not committed.

Recover the latest codex/c12-b01-task9-coordinator checkpoint, not automatically
main, then read current-status.md and repository-recovery.md. Scripts/tests and
this report are portable tracked artifacts; raw receipts, caches and live process
state are not committed or restored by cloning. Helper tests use fresh owned
fixtures and need no previous Go cache. Do not replay child-readiness-001 after
recovery: that authorization and identity are consumed even when raw files are absent.

Original origin87, descendant-test failure/16 unrun checks, Envelope sampling and
Pair remain unresolved; C11/C12/Task10/B01/production/main gates stay open. Further
actual execution or fixture/timing changes need separately agreed scope. Do not
repeat until passing, enlarge the timeout or replace old failure with this result.
