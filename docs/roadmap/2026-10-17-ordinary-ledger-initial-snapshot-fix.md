# Ordinary ledger initial snapshot — bounded fake regression complete

Date:2026-10-17. Base776f2d3e78850044c8b861daf22969c1a00d90dc.
User approved18 additional fake calls,total47 review-repair calls. Scope:
stale initial-shutdown snapshot provenance only;no native workload or rollout.

## Cause and minimal correction

Previously LastMembers retained the last successful runtime list when a later
shutdown poll failed,truncated or exceeded its deadline. Finish combined that
old list with new failed query facts. Each Poll now resets its transient member
snapshot and CurrentPollQuery before any early return. A completed query records
its own facts there;only validated,in-budget list data fills LastMembers.
Finish takes InitialMemberQuery from this per-invocation slot,not the historical
Report.LastQuery. First/last historical report facts remain intact. Valid list
facts remain available even if later admission fails;no failure is reclassified.
The initial list is cloned,so subsequent final polls cannot overwrite it.

No new API rights,retries,timeout windows,acceptance schemas or diagnostic joins.
Existing sticky failures,owned-handle release and shared500ms budget remain.
No second independent review;this narrow follow-up was explicitly user approved.

## RED/GREEN evidence

New suite:ordinary-ledger-initial-snapshot.tests.ps1. Real ledger/session with
existing fake API/clock;no P/Invoke or Run entry invocation.

| Call | Host | Phase | Exit | ms | Evidence |
| --- | --- | --- | --- | --- | --- |
| snapshot1 | PS7 | RED | 1 | 1332 | RuntimeSeed passes;NoStaleInitialMembers-error234 fails as planned |
| snapshot2 | PS5.1 | RED | 1 | 1226 | Same planned assertion |
| snapshot3 | PS7 | GREEN | 0 | 1307 | 21 PASS,empty stderr |
| snapshot4 | PS5.1 | GREEN | 0 | 1185 | 21 PASS,empty stderr |

Both GREENs seed runtime PID2,then shutdown PID3 with Error234,truncation or
501ms fake overrun. They require no stale initial members,current raw error/count/
timing facts,sticky refusal,no retry/reopen and one release of the runtime owner.
Success controls preserve current initial PID3 despite later empty polls and
release both admitted owners. Initial-accounting failure leaves the list query
unattempted rather than substituting runtime facts. This verifies fake behavior,
not real Windows timing or Error87 remediation.

## Final allowlist

| Suite | PS7 call/ms | PS5.1 call/ms | PASS per host |
| --- | --- | --- | --- |
| ordinary-ledger.tests.ps1 | 5/1342 | 12/1382 | 9 |
| ordinary-ledger-session.tests.ps1 | 6/1063 | 13/1176 | 14 |
| ordinary-ledger-receipt.tests.ps1 | 7/1533 | 14/1895 | 88 |
| ordinary-ledger-integration.tests.ps1 | 8/1197 | 15/1401 | 9 |
| retained-lifetime.tests.ps1 | 9/860 | 16/822 | 27 |
| retained-lifetime-integration.tests.ps1 | 10/1073 | 17/1129 | 12 |
| retained-lifetime-receipt.tests.ps1 | 11/1389 | 18/1721 | 16 |

All14 final calls exited0 with empty stderr,350 assertions. New GREENs add42
checks separately. 18/18 calls used;combined29+18=47/47. Original implementation
34 calls are a separate older budget,not part of47. No further execution granted.
Fresh64-bit hosts,-NoLogo -NoProfile -NonInteractive -File,120s outer limit;
PS5 child-only inbox PSModulePath correction from the prior report retained.

## Evidence preservation and recovery

Separate snapshot-invocations.json and snapshot-N stdout/stderr preserve the
prior29-call repair index byte-for-byte. Raw artifacts stay ignored in the plan's
.superpowers/sdd directory;tracked source/tests/report suffice to resume elsewhere.
SHA256 saved-file reads:

- snapshot index:5AEED8AA20DFDFF31563655ECA6FCC93489983EDC30C87F6745144199B2ADEDA
- frozen29 repair index:ED2DF76BEE9886E1A58BBD86230DBA4880EDD69E22C0043042356829C0DF55BA
- frozen34 original index:38F718E5D7BABA66B8D55AB6630EA2F5543C91A7CA951390E4E8B769092D0340

Ruling:use a separate18-call index to preserve the consumed29-call record;
cost:new-machine readers combine29+18,explicitly reported above. No other new
scope rulings or deferred minors. Worktree/raw records retained under the user's
standing progress-preservation instruction;no deletion,merge or PR inferred.

Author closeout:checked transient reset before early returns,query/member origin,
historical report preservation,initial clone lifetime and unchanged refusal/owner
paths against the diff. No additional known source blocker found in this narrow
follow-up. Existing six-review-finding scoped regressions and their limitations
remain in the [repair report](2026-10-17-ordinary-ledger-review-repair.md).

Next:separately plan remaining real-environment/native validation if desired.
No full project/controller suite ran;mechanically updated old count fixtures
remain unexecuted. Native adapter,OS scheduling/handle behavior,Error87,metadata,
main and production/acceptance gates are not cleared. Do not replay consumed
metadata/probe identities or run a workload based on these fake results.
