# Ordinary leaf check — implementation verified, fake-only

User approved the [written design](../superpowers/specs/2026-10-16-leaf-ordinary-check-design.md)
and selected Native/inline execution of the [three-task plan](../superpowers/plans/2026-10-16-leaf-ordinary-check.md)
after `8fa1363c0f3241600d2eeb0c7ca1503666f8d457`. Work stays on the existing
development branch/worktree. This record grants no real invocation.

## Task1 and Task2 evidence

Planner checkpoint `2240adbe`: `PlannerExists` first failed on both hosts;
subsequent planner suite passed44 assertions per host, exit0. Planning launches
no workload, creates no output root and inventories neither SDK nor product.
It freezes13 helper and four input bindings, exact default-vet arguments,
20 environment fields and600s/150s budgets; changed values/types/paths are refused.

Flow checkpoint `b3a0b52f`: `ControllerExists` first failed on both hosts.
Initial positive testing exposed two fixture defects: null collection fields
instead of real empty arrays, and a runtime-only forced-termination flag applied
during compilation. Both were traced and corrected in fixtures, not ordinary
completion predicates. Final Task2 flow selection passed31 assertions per host;
planner rerun passed44 each. All four final Task2 child commands exited0.

The fake seam verifies executable/arguments/environment/cwd/budgets and durable
attempt order. Tests exercise original controller/file/receipt/output behavior,
but replace native execution and expensive SDK/product inventory. Their data-only
executables are never run. Boundary source drift is modeled by altering the
captured binding, not mutating a production helper in place.

## Task3 execution and review repairs

Saved-reader RED failed `SavedReaderExists`; full controller RED failed
`SuccessHasDurableCompletedMarker`, on both hosts. Independent final audits,
explicit15-file manifest, completion marker and saved-only evaluator are now
implemented. Saved suite passed22 assertions per host and full controller passed45
per host, exit0. Verified hosts:PS7.6.5 and Windows PS5.1.26100.9711, both64-bit.
These were initial Task3 results,not the final repaired-source regression below.
The implementation is now verified in the approved fake-only scope;overall
acceptance remains false and real execution has not occurred.

The first final-regression orchestration stopped after planner PS7 exit0/44 and
Windows PS5.1 exit1/1 because Get-FileHash was unavailable. A controlled host
probe found direct invocation uses Windows module paths and resolves the command,
whereas Start-Process inherits PS7 module paths and fails resolution. No product
defect or gate failure was suppressed:the failed logs remain, and the corrected
serial launcher uses a new log identity with unchanged production source.

The final allowlist is exactly eight scripts per PS7/Windows PS5.1 host, as named
in the plan. Historical build-only or matrix totals are not substituted for this
increment;the complete post-fix results are recorded below.

One fresh read-only review of8fa1363c..ab30df3c found0 Critical,2 Important and
1 Minor status-document issue. Important1:leaf receipt validation delegated to
truthiness-based shared predicates,allowing string flags/scalar collections or
nonordinary live-sampling metadata in resealed saved evidence. Important2:the
shared recorded-process wrapper can replace a native exception if its own
Set-Content error persistence fails,before the leaf controller receives it.
Both were independently reproduced with fake inputs:the string-flag refusal
assertion failed,and the native-preservation assertion observed the inner
persistence error instead of the original native error. The one repair pass has
now produced covering RED/GREEN evidence and a fresh complete final compatibility
batch. No second reviewer substitutes for those tests. The minor
stale-footer issue was updated in already-planned status/recovery closeout.

Important1:the tracked receipt-flag test first failed on both hosts after a
positive populated nested receipt. Native-model field/collection type validation
and typed-false ordinary-mode checks are leaf-local;shared guards are unchanged.
PS7 default JSON parsing promotes ISO strings to DateTime,a behavior reproduced
with a controlled parser probe. The leaf reader preserves JSON strings through
DateKind=String where supported rather than weakening native-field checks.
The resulting saved suite passed35 assertions per host,exit0,covering nested
flags,scalar collections,nonordinary mode and timestamp strings.

Important2:the tracked inner-wrapper-persistence test failed on both hosts.
A leaf-owned recorded adapter now calls the unchanged ordinary
Invoke-MatrixNative boundary and independently preserves native primary and
secondary recording errors. The targeted test passed2 assertions per host,
exit0,with one native seam attempt and all eligible final audits. This is not
an independent rereview of the fixes;the executor verified RED/GREEN.
Repair checkpoint:`b54a2c6be899714f656f713b432862041a2b0a1a`. Use the latest
development tip for subsequent preservation/final verification metadata;this
hash alone does not certify a regression batch;the measured final results below do.

The pre-fix corrected serial final batch was interrupted after the findings,
with8 completed commands exit0 (planner44/controller45/saved22/mapping36 per host).
Its interrupted legacy-matrix execution is not a completed result. The next
full final batch uses a new identity and pins the repaired source;old/partial
results are preserved,not presented as post-fix verification.

## Final repaired-source verification

Frozen implementation:`b54a2c6be899714f656f713b432862041a2b0a1a`.
Exactly16 fresh child commands completed,all exit0,with empty stderr;8 suites
per host. Hosts:PS7.6.5 and Windows PS5.1.26100.9711,both64-bit. Each host passed
407 assertions,including1 explicitly labelled static check (not407 native tests).

| Script | PS7 assertions | PS5.1 assertions | Static included per host |
| --- | ---: | ---: | ---: |
| leaf-ordinary-check-plan.tests.ps1 |44|44|0|
| leaf-ordinary-check.tests.ps1 |46|46|0|
| leaf-ordinary-check-saved.tests.ps1 |35|35|0|
| mapping-single-check.tests.ps1 |36|36|0|
| matrix-check.tests.ps1 |60|60|0|
| leaf-stage-build-only-plan.tests.ps1 |60|60|0|
| leaf-stage-build-only-scope.tests.ps1 |41|41|0|
| leaf-stage-build-only.tests.ps1 |85|85|1|

All71 immediate .ps1/.cs source/test fingerprints matched after the batch.
Aggregate SHA256:`AF83F1057EA2B71E31BCB38BD59F6518C2C68424BE77F1B1FD62A04D17340869`.
Definition:form `Name|UPPERCASE_SHA256`,sort with StringComparer.Ordinal,join LF
with one final LF,hash UTF8 without BOM. This binds current worktree bytes,not a
promise of identical checkout line endings on another computer.
Raw outcomes SHA256:`3418EF20B6E8F67F57B3BE441FA59D0A72F986BC7ADFA92D34419DA739E50730`;
raw source-pin JSON SHA256:`CAF56008B22B28DDE5B2D1F7E82042470C2E4127BEE6397F2B25C33453D084A1`.
Raw records remain local/ignored;this portable report carries the measured facts.

One fresh review found the two Important issues above;one fix pass reproduced
each RED then GREEN and passed the final16-command batch. Fixes were verified by
the executor,not independently rereviewed. The minor status-document finding was
closed as planned recovery housekeeping;no deferred minors. Existing shared
guard/collector/matrix/mapping/build-only and product sources were not changed.

## Decisions and remaining gates

1. Use PowerShell bookkeeping instead of Bash skill scripts:Windows-only user
   constraint;cost is manual brief/ledger checking.
2. Do not install/download or run a whole-Go baseline:explicit approved fake-only
   scope;cost is no new full-repository regression evidence.
3. Run final hosts serially through direct invocation:preserve legal host module
   loading without global environment/security changes;cost is longer wall time.
4. Overlap one read-only implementation review with final frozen-source
   verification:independent activities,not a test-result certification;cost is
   that any review fix still requires fresh final covering verification.
5. Use a leaf-owned recorded adapter rather than change the shared wrapper:
   primary-error preservation conflicts with its unguarded persistence;cost is
   one local conversion/persistence path requiring covering fake tests.
6. Update stale plan text as required status/recovery closeout,not optional code
   polish;cost if missed is duplicate approval/execution during recovery.
7. Executor verifies final compatibility outcomes after review;reviewer did not
   certify future results;cost if wrong is a false regression claim.
8. Initial counts/host exits use executor-read preserved logs,not reviewer reruns;
   cost if wrong is an incorrect evidence summary.
9. Real Go/default-vet/package success remains unjudged under fake-only scope;
   cost is no real package acceptance.
10. Actual SDK/reference/module readiness remains separately gated without real
    inventories here;cost is no environment-readiness conclusion.
11. Windows native shutdown/error87 cause remains unresolved;synthetic receipts
    do not establish OS behavior;cost is no historical root-cause claim.
12. Full-repository regressions and unchanged internals beyond directly reused
    contracts remain outside the allowlist;cost is possible unexercised bugs.
13. Do not infer exhaustive test-body/analyzer/process coverage from observed
    finite metadata;cost is remaining coverage limits.
14. Bounded file checks do not promise hostile-writer atomicity or power-loss
    guarantees;cost is that these external threats remain unproved.
15. Hashes bind integrity,not the authenticity of arbitrarily fabricated audit
    history;cost is that untrusted raw records alone cannot prove execution.
16. Cross-machine replay needs new preparation/bindings because raw state is not
    cloned;cost is that cloning restores progress,not a ready environment.
17. Remote synchronization/final documents are executor handoff checks;cost if
    missed is a false preservation claim.
18. Main/security/production/C11/C12/Task9/Task10/Batch01 acceptance stays open
    with overall Acceptance=false;cost is no deployment/main-ready conclusion.

No real Go compile, generated-binary runtime, native workload, PID query,
SDK inventory, Docker, matrix replay, security change or main merge occurred.
The new controller does not resolve historical error87, old timeouts/fuzz failure,
physical PITR/Batch01 or broader C12/serving/production/main acceptance.

Recover from the latest `codex/c12-b01-task9-coordinator` tip, not automatically
main. Necessary source/tests/plans/progress are portable; fake/raw evidence and
machine SDK/module/cache/binary data stay ignored. A real run requires a later
separate concrete approval, even after all fake tests pass.

## Preservation and workflow archive

Completed implementation/report checkpoint
`224780a3dc741984ec2384942f68b8e5a962e3a1` was pushed with remote SHA=local and
clean worktree. The latest tip also carries this follow-up preservation metadata.
Only this plan's SDD scratch was removed after copying and hash/length-verifying
all106 files into the ignored local archive
`.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/leaf-ordinary-implementation-records-001`.
The worktree,other plans,diagnostic histories and fake fixture outputs remain.
The archive retains the failed launcher,interrupted pre-fix batch,RED/GREEN logs,
review,briefs and final records;its manifest excludes itself and all SDK/cache/tmp
or separately preserved fake fixture directories. Manifest SHA256:
`1DF76E06DED1C8D8ED824D8ADDA054AF1839AA84A501C3EE45FA02052C4C9C61`.
The deleted scratch is locally recoverable from that verified copy,not cloned
from Git. All18 execution rulings/costs are above;no deferred minors remain.
