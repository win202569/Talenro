# One short-child observation — 2026-10-15

## Approved scope

The user approved one controlled parent/short-lived-child observation after
diagnostic checkpoint `4b24b2eb432673de453db84a81fc3a74d9e7f419`: no-process entry
tests first, then one fresh identity, readiness before release, parent finishes
after release, child requests150ms before finishing. Whole native body budget20s;
existing500ms natural observation/shared5s shutdown-cleanup clocks unchanged.
No retry on failure/inconclusive evidence, Go/fuzz/matrix/security or guard change,
SDK invocation, historical PID query or main merge.

The new short-child-observation.ps1 plans without creating files or launching.
Parent and child use the existing fixed Windows PowerShell executable; the child
launch is hidden and occurs exactly once. Each side's readiness/release wait is
bounded10s inside the outer20s job budget. Stage payloads carry a fresh nonce,
self-reported process ID and UTC, published by moving a completed temporary file.
Stage names ending in exit are pre-exit intent markers, not OS process-exit proof.
Requested150ms is not a guaranteed parent-exit/child-exit gap.

Controller uses the unchanged native process boundary and completion guard once.
It preserves the receipt, stages and result before refusing incomplete evidence.
A diagnostic ObservationPassed requires all stage identities/order, a matching
owned queried child member, initial-query metadata and existing natural-completion
success. Markers/aggregate counts alone cannot establish child ownership. It never
grants product Acceptance or OriginalTestPassed; both remain false.

## Entry validation and review

Initial RED: missing bounded entry. An intermediate fixture omitted schema3 and
therefore did not model actual forced-cleanup refusal; the test was corrected to
use the real schema3 policy and complete literal member shape, without changing
any validator. Initial17 new assertions PASS on PowerShell7.6.5 and Windows
PowerShell5.1; existing19 initial-job-query assertions also PASS on each.
Entry units execute generated parent/child bodies in the host with Start-Process
and Start-Sleep replaced at the external boundaries; they start no child process.
They test actual stage payloads/order, the150ms request and encoded bodies, not
native scheduling/delay duration or real hidden-launch behavior. Synthetic
schema3 receipts cover missing/unknown child, metadata absence, forced cleanup
and nonce mismatch. Six additional controller assertions replace only the recorded
native invocation; they exercise real planning/result/stopped persistence, one
call on success/refusal, and duplicate-identity rejection before any second call.
Final count is23 new +19 existing =42 per host. No live collector/process capture
occurs in these unit tests.

The701 recorded product source hashes were rechecked unchanged. Existing broad
driver fixtures/full runner/product suites are not rerun in this narrow increment;
their earlier233-per-host results remain historical, not newly re-executed here.
Independent review found no Critical/Important blocker. Interpretation limits:
success does not prove actual parent-before-child termination or a live-member
final-wait branch; inspect native state/final evidence before describing that
branch. QueryEvidenceComplete means successful ordered accounting/list metadata,
not exhaustive raw-count/ID consistency validation. The same reviewer inspected
the added controller assertions and closed that coverage follow-up; no new blocker.
Review was read-only, without execution/PID queries or independent test-result,
environment, actual cleanup or production-readiness validation.

## One actual observation — PASS, diagnostic only

`short-child-observe-001` ran once. UTC14:17:41.9784983Z–14:17:45.3660737Z,
3387.5754ms; root exit0, TimedOut=false. StageEvidenceComplete,
ChildMatchedOwnedEvidence, QueryEvidenceComplete, CompletionPassed and
ObservationPassed=true. Acceptance/OriginalTestPassed=false. No stopped record,
retry, later identity or additional native invocation.

The saved root ID is19084. Child-ready/child-exit self-report18452 with the same
fresh nonce; native evidence independently matches that member as owned, queried
successfully, powershell.exe, initially and once-confirmed running/exit259.
RootExited=true precedes its inspection, so this sample did observe a live owned
child after the root had exited, not merely intent markers. Its same retained
handle was then wait-signaled/exited and read exit0: FinalShutdown attempted,
SameHandleVerified=true, shared-clock start22ms/end203ms, remaining budget478ms.
These are saved historical identities; no current process lookup was performed.

A second member22564 is exact-handle owned conhost.exe, also initially/confirmed
running/259, finally signaled/exited0 on the same handle at shared-clock203ms,
remaining budget297ms. Both members' core acquisitions completed before the first
final wait. TotalProcesses4 is lifetime aggregate only; it does not identify the
other already-absent process or establish that only parent/child existed.
No identity for matrix002's unknown member can be transferred from this sample.

NaturalShutdownVerified/ShutdownEvidenceComplete/CleanupVerified=true;
TerminationRequested=false, final active0/member list empty. Natural observation
and total shutdown209ms, within unchanged500ms/5s bounds. The initial accounting
query succeeded with total4/active2/terminated0. One member-list query succeeded:
assigned2/returned2/capacity16/buffer136 bytes. Both raw API errors0; UTC and
monotonic boundaries are retained. Derived monotonic intervals: accounting0.0305ms,
enumeration2.5146ms, enumeration end to intended-child OpenProcess start4.7067ms.
Those are this run's recorded boundary intervals, not pure kernel execution costs
or an explanation for earlier missing-member races.

Stages (UTC): ready14:17:44.8420370, release14:17:45.1152786,
parent-exit-intent14:17:45.1353057, child-exit-intent14:17:45.3333061.
Intent timing is not exact OS exit timing or proof of an exact150ms delay.
Stdout is empty; stderr596 bytes contains CLIXML progress payloads, not empty stderr.
Completion is established by native evidence/gates, not by stream emptiness.

All four reviewed helper/test/runner source hashes remain unchanged across this
invocation, and701 recorded product source hashes remain unchanged afterward.
The existing Go SDK was neither invoked nor written; no new full SDK audit was
performed or claimed for this PowerShell-only sample.

| Ignored local evidence under short-child-observe-001 | SHA256 |
| --- | --- |
| plan.json |2FFF5C491B4C2B14A7DF145CC4FA3C65EA30BAF8F2F7EA4486C2B8B40F48B1D7 |
| result.json |00E067F0464614FF673AB93DEA0B20577DBF5CA30D237F72136D634DCFAD7137 |
| process/receipt.json |786A64E1B9C3412DBF1DBF8696BC20D27F2FA9D9EC8980C7B1A22A10CBF581A4 |
| process/stdout.txt (empty) |E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| process/stderr.txt |BE2A813C5E264E6E4B05C9816B17DC6EB039A1CC247767C8927D259FA5541E12 |
| child-ready.json |1ECD505DFDDD4C54B5329222A7B33FBDE941D7A6979AE43164A2D49639041233 |
| release.json |6BB49E5AD928507C66E1B863114FDF644F7357026908C8FBC9D15B35AA70E992 |
| parent-exit.json |4C0BE2C5AAB5E861957373D3D2B6A71603A15013FFA85A028AE3E08546D6AA58 |
| child-exit.json |45EFA20C7176B9B98B3CBD90D46E565998772203A5F750811AC834C22D61BA93 |

Unit logs/source-binding.json are separately retained under ignored
short-child-validation-001. They bind the actual helper bytes and final23-per-host
entry logs plus19-per-host metadata logs, not an additional process observation.
Its final-evidence-manifest.json binds16 process/stage/test/source-binding files;
SHA256 `170E559DD78DCC816073480E487A680AB4C4E4B5B054B9A409838DA7AC0F60E4`.

The same independent reviewer subsequently compared this report/status/README
against the saved native receipt and stage files, finding no factual overclaim.
It confirmed the live-child/same-handle-final-wait interpretation for this sample
only. It did not independently certify invocation uniqueness, unit totals, source
hashes, absence of Go/SDK activity or repository state from those receipts alone;
the coordinator's execution/source-binding/product/Git checks separately support
those claims. No additional process was launched during evidence validation.

## Recovery and remaining boundaries

Save necessary helper/tests, this report and current-status.md on the existing
development branch. Read docs/runbooks/repository-recovery.md. Unit fixtures are
generated from tracked code using the current Windows/.NET host. Raw process
receipts/stages, logs, caches and SDK copies stay ignored and are not cloned.
Do not replay a consumed observation identity on this or another computer.
Any further live observation needs separately approved scope and a fresh identity.

This controlled fixture is not matrix002 or the historical origin87 execution.
Even success cannot prove why that earlier OpenProcess87 occurred or that it was
fixed. Matrix002 remains0PASS/1FAIL/25UNRUN; broad C11/C12/Task10/B01/production/main
acceptance gates remain open.
