# Ordinary leaf check — implementation in progress, fake-only

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

## Task3 status

Saved-reader RED failed `SavedReaderExists`; full controller RED failed
`SuccessHasDurableCompletedMarker`, on both hosts. Independent final audits,
explicit15-file manifest, completion marker and saved-only evaluator are now
implemented. Saved suite passed22 assertions per host and full controller passed45
per host, exit0. Verified hosts:PS7.6.5 and Windows PS5.1.26100.9711, both64-bit.
Final compatibility regression/review remain in progress; no completed
implementation or overall acceptance is claimed by this checkpoint.

The first final-regression orchestration stopped after planner PS7 exit0/44 and
Windows PS5.1 exit1/1 because Get-FileHash was unavailable. A controlled host
probe found direct invocation uses Windows module paths and resolves the command,
whereas Start-Process inherits PS7 module paths and fails resolution. No product
defect or gate failure was suppressed:the failed logs remain, and the corrected
serial launcher uses a new log identity with unchanged production source.

The final allowlist is exactly eight scripts per PS7/Windows PS5.1 host, as named
in the plan. Current results/counts, source pins, final review and remote checkpoint
will be recorded only after they are actually verified. Historical build-only
or matrix totals are not substituted for this increment.

One fresh read-only review of8fa1363c..ab30df3c found0 Critical,2 Important and
1 Minor status-document issue. Important1:leaf receipt validation delegated to
truthiness-based shared predicates,allowing string flags/scalar collections or
nonordinary live-sampling metadata in resealed saved evidence. Important2:the
shared recorded-process wrapper can replace a native exception if its own
Set-Content error persistence fails,before the leaf controller receives it.
Both were independently reproduced with fake inputs:the string-flag refusal
assertion failed,and the native-preservation assertion observed the inner
persistence error instead of the original native error. The one repair pass has
now produced covering RED/GREEN evidence;fresh final compatibility verification
remains pending. No second reviewer substitutes for those tests. The minor
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

The pre-fix corrected serial final batch was interrupted after the findings,
with8 completed commands exit0 (planner44/controller45/saved22/mapping36 per host).
Its interrupted legacy-matrix execution is not a completed result. The next
full final batch uses a new identity and pins the repaired source;old/partial
results are preserved,not presented as post-fix verification.

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

No real Go compile, generated-binary runtime, native workload, PID query,
SDK inventory, Docker, matrix replay, security change or main merge occurred.
The new controller does not resolve historical error87, old timeouts/fuzz failure,
physical PITR/Batch01 or broader C12/serving/production/main acceptance.

Recover from the latest `codex/c12-b01-task9-coordinator` tip, not automatically
main. Necessary source/tests/plans/progress are portable; fake/raw evidence and
machine SDK/module/cache/binary data stay ignored. A real run requires a later
separate concrete approval, even after all fake tests pass.
