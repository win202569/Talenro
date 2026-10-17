# Isolated fuzz diagnosis

Latest:[timing coverage completed](../../roadmap/2026-10-17-ordinary-ledger-timing-coverage.md).
Test-only late-drift/finalization-lock checks catch both in-memory regressions;
21 per host/final350 passed,70/70 cumulative calls consumed. Product unchanged;
real adapters,Stage B and security gates are not certified or authorized.

Latest:[fixed inbox policy implemented/fake-verified](../../roadmap/2026-10-17-ordinary-ledger-inbox-policy.md).
Schema2 pins shell signature/hash/identity/link count under held read binding;
source/evidence still require one link.64 checks per host/final350;54/54 calls
used. See documented index anomaly and deferred timing-regression gap. Real OS
adapters and Stage B remain unverified/unapproved;do not invoke native probes.

Latest:[native-probe Stage A review fixes fake-verified](../../roadmap/2026-10-17-ordinary-ledger-native-probe-repair.md).
Load snapshot,held executable binding and fail-closed fake protections pass47
checks per host/final350.36/36 Stage A calls used. Child result explicitly does
not prove exit ordering. Stage B remains unapproved:do not invoke real probes
without current prerequisites/exact-input approval. Earlier review blockers below
describe the pre-repair source,not a new native clearance.

Latest:new ordinary-ledger-native-probe.ps1 is an unfinished Stage A checkpoint.
[Independent review](../../roadmap/2026-10-17-ordinary-ledger-native-probe-review.md)
found source/executable binding gaps and an overstrong child-ordering scenario.
Do not invoke its real/native boundary. Fake39 per host/final350 passed,but
18/18 used and no fixes/retests after review;Stage B not authorized.

Latest:[initial snapshot follow-up](../../roadmap/2026-10-17-ordinary-ledger-initial-snapshot-fix.md)
fixed the provenance gap below,21 new fake checks per host and final350 passed.
Review repairs47/47 used. No native/workload or acceptance authority;historical
blocker statements below remain as the record of the pre-fix checkpoint.

Closeout blocker:failed initial shutdown queries can reuse runtime member facts
in InitialMembers;sticky refusal remains,but provenance is inaccurate. No fix or
rerun after29/29. The green scope below is not implementation completion.

2026-10-17 latest:the six review findings now have scoped corrective regressions;
A/B/C passed94 checks across two hosts and final14 allowlisted runs passed350.
29/29 repair calls used. No native/workload acceptance or replay authorization;
earlier broken-writer warning below describes the pre-repair checkpoint.
See [corrective-pass evidence/PS5 host setup](../../roadmap/2026-10-17-ordinary-ledger-review-repair.md).

2026-10-17 current warning:ordinary ledger final review found6 blocking defects,
including null new-mode fields causing existing matrix/leaf ordinary receipts to
be rejected. Do not execute those workflows from this unfinished checkpoint.
No repairs/tests after review;34/34 calls consumed. See the
[blocking review](../../roadmap/2026-10-17-ordinary-ledger-final-review.md).
The350 fake/static assertions below are not an implementation-complete claim.

2026-10-17:ordinary identity ledger adds an explicit isolated native entry and
schema4 evidence evaluator. Existing Run overloads/callers and diagnostic modes
remain unchanged;Assert-ProcessCompletion refuses schema4 and new-mode markers.
Verification scope is fake/reflection/static only,not native adapter validation
or Error87 repair. No workload/rollout is enabled. See the
[implementation record](../../roadmap/2026-10-17-ordinary-ledger-implementation.md)
and [approved design](../../superpowers/specs/2026-10-17-ordinary-member-identity-ledger-design.md).

2026-10-16:after renewed approval,ordinary check001 completed one default-vet
build exit0/73.927s and original runtime exit0/7.287s;33 top-level/135 subtest
PASS,no SKIP. Ordinary guards/full SDK/product audits and15-file completion
binding pass;PackagePassed=true,Acceptance=false. No retry or source changes.
Read the [actual result](../../roadmap/2026-10-16-leaf-ordinary-check-001.md).
Earlier stopped-preflight entry below is historical;001 must not be replayed.

2026-10-16:ordinary check001 was approved but stopped before preflight execution
because the executor's command failed PowerShell parsing. No controller/build/
runtime call or output/control root creation occurred;no retry or acceptance.
Read the [stopped preparation record](../../roadmap/2026-10-16-leaf-ordinary-check-001.md)
before any new execution decision. Historical diagnostics below stay separate.

2026-10-16:separately approved build-only001 completed one real compile,exit0/
331.219s,5 retained admissions below32,scope/trigger/core/lifetime and final
audits complete. Completed marker/10 artifacts verified;11536-file SDK/701
product comparisons unchanged,66 source/test+16 helper+4 input pins matched.
Outer24m17s includes15m45s pre-native work;no binary runtime,retry or ordinary
vet/test/main acceptance,error87 remains unresolved. Read the
[actual result and recovery](../../roadmap/2026-10-16-leaf-stage-build-only-001.md).

2026-10-16:user selected separate build-only diagnostic option1 (fixed -vet=off
only in the new separate entry,ordinary vet gates unchanged). The
[written spec](../../superpowers/specs/2026-10-16-leaf-stage-build-only-design.md)
is approved;the [three-task plan](../../superpowers/plans/2026-10-16-leaf-stage-build-only.md)
was approved for Native/inline fake-only implementation. Planner/scope/refusal
and the separate single-call controller are implemented. One fresh review/two
Important repairs and final21-suite fake regressions are complete:884 checks per
PS7/5.1 host (11 static included each),all42 commands exit0;66 source/test hashes
matched. Controller checkpoint:a4bfe80a;no deferred minors. Read the
[implementation record](../../roadmap/2026-10-16-leaf-stage-build-only-implementation.md).
Implementation does not authorize a real diagnostic or grant ordinary acceptance.

2026-10-16:saved-only stage compile001 budget analysis confirms cumulative32
slots,27 vet admissions in about0.95s,and225 logged vet command/config pairs
after the trigger. No new Run/PID lookup/filter/cap or gate change. Read the
[analysis and next design boundary](../../roadmap/2026-10-16-compile-stage-budget-analysis.md).

2026-10-16:separately approved leaf-stage-retained-compile-001 compiled once,
exit0/66.574s,trigger/core complete and final audits unchanged; deferred registry
still capped at32 admissions (27 recorded vet.exe). Diagnostic STOPPED,no binary
runtime/retry/acceptance. Read [actual result and saved-only next scope](../../roadmap/2026-10-16-leaf-stage-retained-compile-001.md).

2026-10-16:approved compile-stage retained diagnostic implementation passed final
fake-only verification after one independent review/fix pass:18 suites/698 checks
per PS7/5.1 host (10 static included each), all36 commands exit0. It adds a bounded
target-package log trigger and a separate fixed controller, not default/matrix
enablement or live execution. Read the
[implementation record](../../roadmap/2026-10-16-compile-stage-retained-implementation.md).

2026-10-16:saved-only compile001 chronology explains cumulative32 admissions
stop new queries at23.42s; retained root-wait/final reads still occur. No new
execution/cap change; the sample remains stopped and error87 unexplained. Read
[analysis and bounded next decision](../../roadmap/2026-10-16-leaf-retained-cap-analysis.md).

2026-10-16:separately approved leaf-retained-compile-001 real compile returned0
but diagnostic observation STOPPED at32 admissions. No timeout/forced cleanup;
compiler evidence/core cleanup verified, lifetime sampling incomplete. All final
audits unchanged,11 artifacts verified, stopped/no completed marker. No binary
execution/retry/acceptance. Read [actual result and recovery](../../roadmap/2026-10-16-leaf-retained-compile-001.md).

2026-10-16:fixed nodeleafbinding retained-only compile wrapper prepared with
portable fake tests. No real Go compile authorized/executed in this increment;
never runs the generated binary and cannot grant package/ordinary acceptance.
Counter32 stays capped; independent final audits and persisted SDK identity are
required. Read [entry and recovery](../../roadmap/2026-10-16-leaf-retained-compile-entry.md)
before any separately approved live compile.

2026-10-16:separately approved retained-child-observe-002 single real observation
PASS as sampled diagnostics only. Preserved root-wait descendants2, final0/empty;
3 admissions/releases,0 failures, no timeout/forced cleanup. Completion marker
and9 artifact hashes verified. No acceptance/error87 fix/matrix claim;001 remains
stopped. Read [actual sample and recovery](../../roadmap/2026-10-16-retained-child-observation-002.md).

2026-10-16:bounded observer phase fix removes the mistaken schema3 root-wait count
zero requirement, without changing final cleanup/acceptance gates. New realistic
fixture and negative final-residual controls:70 entry checks; five adjacent suites
148 PASS perPS7/5.1. No live rerun;001 stays stopped. Read
[fix and recovery](../../roadmap/2026-10-16-retained-child-phase-fix.md).

2026-10-16:one separately approved retained-child-observe-001 live observation
STOPPED. Retained child/core natural cleanup facts are complete, but the new
observer mistakes the root-wait DescendantsRemaining2 for a final residual and
refuses it. Final active0/empty; saved result remains false/no completion marker.
No retry/fix/Go/matrix. Read
[actual evidence and bounded next step](../../roadmap/2026-10-16-retained-child-observation-001.md).

2026-10-16:dedicated retained-child observation entry prepared with fake-only
tests. Dot-sourcing/planning does not launch; explicit controller consumes one
new retained-child-observe-NNN identity, fixes20s and preserves diagnostic-only
refusal. No live observation is approved/executed by this increment.
Scoped fake-only verification:13 suites/389 assertions perPS7/5.1, including61
new wrapper checks; two review findings fixed with RED/GREEN. Read
[entry, evidence checks and recovery](../../roadmap/2026-10-16-retained-child-entry.md)
before any separately approved real execution. Existing short-child helper and
normal acceptance remain unchanged.

2026-10-16:test-only populated nested receipt coverage closes the prior M1 gap.
Writer-depth6 roundtrip preserves32-entry metadata and nulls; in-memory corruption
controls verify test sensitivity. Three portable suites55 PASS perPS7/5.1, no
collector/guard change or real process run. Read
[coverage and recovery](../../roadmap/2026-10-16-retained-receipt-coverage.md).

Retained-handle implementation now has explicit opt-in native entry, bounded
registry, optional receipt fields and unconditional diagnostic-mode refusal.
Scoped verification PASS:11 suites/298 checks perPS7/5.1 plus2 PS7 supplemental
checks at the2026-10-15 checkpoint; two review blockers fixed through RED/GREEN.
Its deferred nested-receipt enhancement is covered by the2026-10-16 test increment
above. No real native Run/probe or matrix rerun/acceptance.
Read [implementation and limits](../../roadmap/2026-10-15-retained-lifetime-implementation.md).

Retained-handle written spec and [three-task implementation
plan](../../superpowers/plans/2026-10-15-retained-handle-diagnostics.md) were approved
for inline fake-only implementation. No live experiment is approved by that scope.

Approved retained-handle diagnostic design authorizes planning only:
explicit default-off entry,32 admission attempts, owned-handle lifetime/teardown,
sampled non-acceptance evidence and process-free test requirements. Read
[written spec](../../superpowers/specs/2026-10-15-retained-handle-diagnostics-design.md).

Latest read-only matrix003 analysis separates empty-list success, successful
owned-member acquisition/final waiting and failed initial OpenProcess87. No root
cause/fix established; no test/process/PID query/gate change. Existing live sampling
is not a retained-handle registry; notification-only evidence is insufficient.
Next scope is lifecycle-evidence design review, not matrix repetition. Read
[comparison and recovery](../../roadmap/2026-10-15-matrix003-acquisition-analysis.md).

Latest matrix-check-003:2PASS/1FAIL/23UNRUN. First two packages PASS; check03
nodeleafbinding compile root exit0, but listed-member OpenProcess87 caused the
unchanged natural-completion gate to refuse it. Final cleanup verified0/empty;
no runtime03/later checks/fuzz/retry/main merge. Copied SDK/product/artifact/source
checks unchanged; cause unresolved, Acceptance=false. Read
[matrix003 evidence and recovery](../../roadmap/2026-10-15-limited-matrix-003.md).

Latest mapping-check-001: one approved fixed nodeissuermapping build/run PASS,
35 top-level/65 subtests; native natural cleanup/final0, no timeout/termination.
Runtime initial members empty, so no residual-member OpenProcess path was tested
and historical87 is not resolved.36+19 no-process checks PASS perPS7/5.1 host;
review found no actionable finding; copied SDK/product/seven artifacts unchanged.
No matrix rerun/main merge. Read
[single-package results and recovery](../../roadmap/2026-10-15-mapping-single-check.md).

Latest short-child-observe-001: one approved20s-bounded observation PASS in3388ms.
Ready/release markers matched the intended native owned running child; same-handle
final exit and empty natural job verified within existing500ms/shared5s gates.
New query boundaries retained; no timeout/termination/retry/Go/matrix.23 new+19
existing no-process checks PASS perPS7/5.1 host; independent review no blocker.
It does not explain or fix historical matrix002/OpenProcess87. Read
[actual evidence, interpretation and recovery](../../roadmap/2026-10-15-short-child-observation.md).

Latest diagnostic-only increment: process-native records the existing initial
post-root-wait accounting/member-list API clocks and raw result/error, without
new queries/retries or predicate changes. New initial-job-query.tests.ps1 uses
only an empty job and invalid job handles, no child/process probes;19 assertions
plus214 existing process-free checks PASS on each PS7/5.1 host (233 each).
Independent review found no blocker;701 recorded product hashes unchanged.
No Go/matrix retry or underlying failure fix. Read
[implementation, persistence limits and recovery](../../roadmap/2026-10-15-initial-job-query-diagnostics.md).

Latest fresh matrix0022026-10-15:0PASS/1FAIL/25UNRUN, stopped at check01 completion
gate. Tests output35 top-level/65 subtest PASS and exit0, but one listed member's
OpenProcess error87 leaves state/ownership unverified. Natural shutdown refused,
termination requested, final cleanup0/empty, no timeout. No retry/check02/fuzz.
701 source hashes/four compile artifacts unchanged; copied and final original
SDK11536-file audits passed.
No root-cause or broad acceptance claim. Read
[result, evidence limits and recovery](../../roadmap/2026-10-15-limited-matrix-002.md).

Latest bounded scanner repair2026-10-15: only repository-root
.superpowers/diagnostics is excluded by the test-only contract consumer scan.
RED reproduced two false positives; GREEN entire statecontract package33 tests/
96 subtests PASS, including ten portable scope cases. Copied-toolchain audit PASS;
other700 source inventory entries unchanged; independent review found no blocker.
No full matrix rerun or main merge; the stopped001 result below is unchanged.
Read [bounded repair and recovery](../../roadmap/2026-10-15-statecontract-scan-scope.md).

Latest new limited matrix2026-10-15: matrix-check.ps1 provides a separate26-check
flow,24 builds/25 runs maximum, preserving package cwd and strict native gates.
214 scoped assertions PASS perPowerShell5.1/7 host; review closed a runtime-evidence
binding gap before execution. One actual matrix-check-001 stopped12PASS/1FAIL/
13UNRUN at TestContractBoundary: its whole-repository scan includes .superpowers
and parses an intentional-invalid SDK syntax fixture. No timeout/retry/old guard
change/main merge. All26 native attempts cleaned naturally;701 product hashes and
88 bound files unchanged. MatrixPassed=false, Acceptance=false. Do not delete SDK
evidence or replay001 to bypass this gate. The separately approved scanner repair
is recorded above and does not change this historical result. Read
[scope, failure, actual hashes and recovery](../../roadmap/2026-10-15-limited-matrix-validation.md).
New matrix-check.tests.ps1 uses portable process-free fixtures; optional
-RuntimeEvidenceOnly selects a synthetic integrity regression, not a real matrix.

Latest separate Pair observation2026-10-15: pair-sample.ps1 validates the retained
uninstrumented005 compile bundle and invokes only FuzzIssuerPair once with a
fresh identity/cache. No copy/recompile/version/list/Envelope process.154 scoped
assertions PASS perPowerShell5.1/7 host; review found no blocking defect. One
actual pair-observe-001 PASS,6097.5363ms, natural final0/empty, no timeout/retry.
Only5 executions were reported after4/4 baseline; coverage is very limited, not
full fuzz/product acceptance. All eight retained artifacts and701 product hashes
unchanged; Acceptance=false. Historical matrix3 and wider gates remain open.
Read [evidence, review limits and recovery](../../roadmap/2026-10-15-pair-single-observation.md).
New pair-sample.tests.ps1 synthesizes its own prerequisites and never launches its
unit executable. Actual sampling requires a separately approved fresh identity
and independently prepared bound compile; do not replay pair-observe-001.

Latest uninstrumented validation2026-10-15: separate envelope-uninstrumented.ps1
uses one independent unpatched fixed toolchain with full11536-file pin/manifest
binding, fresh offline caches, unchanged budgets and native cleanup gates.
109 scoped assertions PASS perPowerShell5.1/7 host; independent review found no
blocking issue. One actual005 build and short fuzz validation PASS:74867
executions, natural verified final0/empty, no timeout/retry,701 product hashes
unchanged. Acceptance=false; old deadline/origin87/Pair/wider gates remain open.
Read [implementation, measured evidence, limits and recovery](../../roadmap/2026-10-15-envelope-uninstrumented-validation.md).
Unit tests accept explicit -Toolchain/-ModuleCache; independently prepare those
prerequisites on another computer. Tests synthesize the native boundary and never
execute the unit binary. Do not replay consumed uninstrumented-001/observe-005,
infer another real invocation, or upload the ignored local toolchain/cache/binary.

Latest readonly comparison2026-10-15: historical matrix3 and004 share all701
product hashes after separator normalization.004 captured child deadline and
suppressed the budget error; controlled parent-first captured child nil and kept
it. Historical operands are absent. Different runtime hooks/corpus/cache/entry
prevent causal or repair claims.001/87 is pre-build evidence, not a fuzz trace.
No new run or code change. See [comparison and next authorization boundary](../../roadmap/2026-10-15-envelope-evidence-comparison.md).

Latest original Envelope observation2026-10-15: new envelope-sample.ps1 validates
compiled inputs/source/product and invokes once without rebuilding.81 scoped
checks PASS per5.1/7 host; one authorized004 sample from003 PASS,73921 executions,
one worker, no watchdog/timeout, natural final0/empty,701 product hashes unchanged.
ObservationValidationPassed is diagnostic only; Acceptance=false. Old deadline/
origin87 and Pair/wider gates remain unresolved. No replay/retry/main merge.
Read [binding, evidence and recovery](../../roadmap/2026-10-15-envelope-single-observation.md).
Tests accept -Toolchain/-ModuleCache for independently prepared prerequisites;
they synthesize records and never run their unit binary. Real sample invocation
requires separate authorization; cloning does not restore the local003 artifact.

Latest Envelope compile-only2026-10-15: optional -CompileOnly conflicts with
-PreparationOnly before planning. It verifies version/origin, compiles once with
the existing offline readonly600s bound and stops before executing any test
binary. Final product/source/binary checks precede success JSON and output.
59 scoped checks PASS per5.1/7 host. One authorized003 build PASS in111914ms,
natural shutdown/final0/empty, compiler/source evidence verified,701 hashes
unchanged. No fuzz/retry/acceptance/main merge. Generated binary remains local and
ignored; recover source/progress from the dev checkpoint, not the old binary.
See [compile evidence, limits and proposed next gate](../../roadmap/2026-10-15-envelope-compile-only.md).

Latest Envelope preparation-only2026-10-15: Invoke-EnvelopeObservation accepts
optional -PreparationOnly and stops after version/origin, before build/fuzz.
Default remains auto-build/observation and must not be used under preparation-only
approval. Success JSON/output publish only after final inventory validation.
New19+existing21 scoped checks PASS on5.1/7; one approved envelope-observe-002
preparation PASS, natural exits, final0/empty,701 product hashes unchanged.
No build/fuzz/retry/acceptance/main merge.001/002 remain consumed; source and
progress recover from the dev branch, not ignored local artifacts. See
[scope, finalization tests, actual evidence and next gate](../../roadmap/2026-10-15-envelope-preparation-only.md).

Latest full runner regression2026-10-15: one approved default run.tests.ps1 run
on WindowsPowerShell5.1 passed31/31 assertions, exit0. Previously unrun16 checks
passed; seven actual fixtures all ended with verified final0/empty owned jobs.
No retry/source/security change/Go/fuzz/main merge. This is the diagnostic runner
script only, not full project acceptance; origin87/Envelope/Pair remain open.
See [full run, evidence hashes and recovery](../../roadmap/2026-10-15-full-runner-regression.md).

Latest revised descendant case2026-10-15: approved migration of only the original
timeout case to explicit readable-file mode retains2s/count/cleanup gates and adds
intended-child nonce/owned-running proof. `run.tests.ps1 -DescendantTimeoutOnly`
returns before all other actual fixtures. One approved WindowsPowerShell5.1 run
PASS, verified empty cleanup;137 helper checks PASS per5.1/7 host. No full suite,
Go/fuzz, retry or acceptance upgrade. The16 later checks remain UNRUN. Read
[scope, selected-run evidence and recovery](../../roadmap/2026-10-15-descendant-timeout-regression.md).
This command launches a real timed fixture; another run/full suite needs separate
approval. Do not replay the consumed runner-tests identity in that report.

Latest readable-file diagnostic2026-10-15: explicit UseScriptFiles=true on the
child-readiness helper requires fresh file-child-readiness-NNN, materializes
exclusive UTF8-BOM scripts and uses normal -NoProfile -File arguments. Default
encoded mode and native2s/security/acceptance rules remain unchanged. One approved
file-child-readiness-001 sample passed diagnostic readiness/timeout/cleanup; creation
bracket21.2584ms and child marker UTC offset615.7873ms. Not an original-test pass or
causal security conclusion.118 helper checks pass on5.1/7; no policy bypass, retry,
WPR or Go/fuzz. See [scope, evidence and recovery](../../roadmap/2026-10-15-file-readiness-diagnostic.md).
Planning writes nothing; invocation requires a separately approved fresh identity.
Do not replay consumed001 or silently switch old tests to file mode.

Latest local trace2026-10-15: one approved recording/new child-readiness-003 run
still FAILED readiness with verified cleanup. Offline runner/WdFilter waiting
stack and MsMpEng/MpRtp readier evidence cover ~98.98% of the2082.4830ms creation
bracket. This is sample-specific participation evidence, not a security defect or
exact scan attribution. Stop requested at10s, but ETL actually covers46s; own
session/collectors ended. Raw ETL/CSVs stay local/ignored; no retry, security change,
Go/fuzz or acceptance upgrade. See [trace evidence and limits](../../roadmap/2026-10-15-local-launch-performance-trace.md).

Latest native launch timing2026-10-15: additive LaunchBoundaries records UTC,
monotonic ticks/frequency and original budget-watch position around creation,
job assignment and thread resume. No API/security/timeout/acceptance change.
One approved child-readiness-002 sample FAILED readiness: creation bracket2097.7060ms,
resume ended at2105.4985ms on the original2s clock. Owned cleanup verified empty.
No retry/Go/fuzz;99 helper checks pass on5.1/7, not the full blocked suite. Read
[timing evidence, failure-path limits and recovery](../../roadmap/2026-10-15-launch-timing-diagnostic.md).
Neither a phase record nor empty forced cleanup is acceptance. Do not replay002.

Latest independent readiness/cleanup fixture2026-10-15: child-readiness.ps1 keeps
the original2s runner bound and separately correlates a unique nonce/child PID
record with non-root owned-running evidence. One actual child-readiness-001 run
FAILED readiness (no record), while owned cleanup verified empty. No retry or
original-test acceptance. New28 and existing33 helper checks pass on5.1/7. Read
[design, actual failure, timing limits and recovery](../../roadmap/2026-10-15-child-readiness-fixture.md).
Dot-source the helper for non-launching planning/evaluation; Invoke-ChildReadinessFixture
is the explicit process-launch entry and requires a separately authorized fresh
identity. FixturePassed is diagnostic only and must be checked by its caller;
it is not the original test/product acceptance. Do not replay consumed001.

Latest full-parent diagnostic2026-10-15: one2s probe observed flushed ENTRY and
BEFORE_START plus their files, but no AFTER_START or child-entry marker. It timed
out with native accounting3; owned cleanup verified empty. Process presence alone
does not prove child-script readiness or explain prior failures. New12 checks and
existing9+12 pass on5.1/7; no Go/fuzz, retry or acceptance upgrade. See
[full-parent evidence and recovery](../../roadmap/2026-10-15-full-parent-probe.md).
Explicit SynchronousStdout mode on the existing tree helper requires a fresh
full-parent-probe-NNN identity; default tree identity cannot be reused for it.

Latest minimal-entry diagnostic2026-10-15: one2s probe observed both flushed stdout
and file marker, naturally completed in265.4917ms with verified empty job. It has
no explicit child launch, and its native TotalProcesses2 does not prove the old
descendant scenario. Original failures/acceptance stay open. Twelve unit checks
pass on5.1/7; no Go/fuzz or retry. See [entry observation and recovery](../../roadmap/2026-10-15-entry-startup-probe.md).

Latest diagnostic2026-10-15: one approved2s stage probe timed out with one recorded
process and no markers, including no first parent-script marker. Cleanup verified
empty; exact cause remains unknown. Nine non-launching checks pass on5.1/7; no
Go/fuzz or retry. See [one-run evidence and recovery](../../roadmap/2026-10-15-tree-timeout-stage-probe.md).
Do not replay probe001 or treat its safely cleaned timeout as original-test success.

Latest implementation2026-10-15: optional ProcessOpen/JobMembership API/timing/error
records diagnose initial acquisition without changing acceptance or cleanup.
New12-check isolated suite passes on PowerShell5.1/7. Related regression is not
green: run.tests stopped on the descendant-timeout fixture, recording only one
process despite verified cleanup. No Go/fuzz or preparation retry. See [diagnostics,
independent review and coverage limits](../../roadmap/2026-10-15-member-query-diagnostics.md).

Latest actual preparation2026-10-15: Envelope001 stopped at the origin command's
natural-completion gate before build/fuzz. A listed member lacked initial
ownership/state proof (QueryError87); owned cleanup succeeded but is not acceptance.
Sampling remains UNRUN; preparation001 must not be replayed. See [current stop and
recovery report](../../roadmap/2026-10-15-envelope-preparation-stop.md).

Latest2026-10-15: one final-exit diagnostic observation did not reproduce the prior
failure. The fixture now saves returned rows before assertions and separately
records cleanup requests, without changing native policy or time limits. Fresh
scoped coverage184 PASS (173 distinct); the old cause remains unknown. No Go or
original Envelope build/fuzz ran. See [diagnostic evidence and recovery](../../roadmap/2026-10-15-final-exit-diagnostic.md).

Original-project Envelope continuation: the user approved one observation sample,
not Pair or full matrix. `envelope-observe.ps1` provides readonly planning and a
fresh-identity, offline, source-bound one-shot build/run entry point. Preparation
and remaining results are documented in the [Envelope execution/recovery report](../../roadmap/2026-10-14-envelope-observation-execution.md).
Source identity now uses the recorded literal six-path order, avoiding different
PowerShell5.1/7 collation while preserving historical identities and hash checks.
Recovery does not authorize a repeated sample or acceptance upgrade.

Current outcome2026-10-14: all six finite actual-framework arms completed and
their expected diagnostic checks passed. Baseline/observe/synchronized exited0;
parent-first/crash/cancel deliberately exited1 with the prescribed complete trace.
Every instrumented arm returned activeWorkers0/watchdogfalse, and every arm ended
with an empty owned job and no forced cleanup. Completed arms were not rerun.
The earlier failed preparations remain failures, not retrospective successes.

Final remaining-arm authorization used synchronized-prep-002, then fresh crash
and cancel once each. synchronized recorded326344 executions; crash injected the
warmup worker panic marker and retained the same opaque error identity through
crasher/worker/return; cancel stopped early after22914 executions and retained exact
canceled identity. Latest three receipts use schema3/v2, earlier results retain
their original formats. Source-bound diagnostic binary reused without rebuilding.

Fresh full scoped suite183 assertions PASS (11 reused duplicates;172 distinct).
Final original11536-file and701 product-hash inventories unchanged. This establishes
bounded actual-framework controls, not the historical Envelope cause, a product
fix, material-error integration, toolchain promotion, main merge or acceptance.
The one earlier independent review concerned an unfinished checkpoint; subsequent
policy changes and integration results are author-verified, not a new review.
See [complete six-arm report](../../roadmap/2026-10-14-isolated-fuzz-coordinator-execution.md).
Fresh-PC recovery uses tracked code/report on the development branch; do not
repeat completed arms merely to recover progress. Copies/caches/raw receipts are
local ignored data; no additional execution or deletion is implied.

Sections below preserve earlier stages and policy explanations, not current
UNRUN status. All former3/6 and UNRUN references below are historical.

Latest approved completion policy: ordinary runs now produce ShutdownSchema3 /
owned-final-signaled-exit-image-optional-v2. A running member is not accepted merely
because its image is readable or accounting later becomes empty: its same retained,
job-verified handle must signal and its exit-code query succeed within the shared
500ms window. Immediate verified exited snapshots remain sufficient. Unknown
ownership/state, query failure, truncation, still-running/late final evidence,
timeout, nonempty job or forced cleanup still fails; shared5000ms cleanup is unchanged.

FinalShutdown is separate from DiagnosticShutdown and all earlier core/initial/
confirmation fields. It records member identity, same-handle proof, query/read
result and shared-clock budget/times. All bounded batch handles/core snapshots are
acquired before any final wait, preventing our own first wait from making a later
listed PID vanish before identity capture. Every retained handle closes in finally;
no unknown-ID or image-name exemption is introduced. Diagnostic-only mode still
produces legacy schema2, never final evidence, and is refused for acceptance.

Schema3 validates per-member final proof and initial-member coverage after JSON
serialization. Legacy schema2 retains its old rules and cannot carry v2 policy or
FinalShutdown evidence. Old failed/diagnostic receipts remain refused unchanged.
final-exit.tests.ps1 passed52 assertions (11 reused/41 new); the full scoped suite
passed183 assertions including11 duplicate reused checks,172 distinct checks.
Controlled missing-image rows used real owned handles/waits; actual native short-
child schema3 completion/serialized receipt passed. No remaining framework fuzz,
version probe or product suite ran in this increment. Actual-arm progress stays3/6,
and the historical image-query failure/root cause remains unproven.

Latest bounded diagnostic-only increment: opt-in `CaptureShutdownDiagnostic` on
Invoke-RecordedProcess records a separate DiagnosticShutdown member observation
on the same retained, live-job-verified handle. Only running-after-confirmation
members with missing image metadata qualify; unknown ownership/core queries skip.
One wait uses the remainder of the shared500ms shutdown window, not a fresh window
per member. Core state/exit/image/first/confirm fields stay untouched. Live pre-root
sampling and ordinary arm execution do not enable this mode. Member cap32 and
shared5000ms cleanup remain unchanged; scheduler latency is not hard real-time.

Every opt-in receipt carries ShutdownDiagnosticOnly=true; Assert-ProcessCompletion
refuses it even if the process naturally finished with an empty job. Diagnostic
observations never substitute for core acceptance evidence or launch an arm.
shutdown-observation.tests.ps1 runs real short/long owned process waits plus
ownership/query/window/no-upgrade controls. It reuses11 same-handle assertions;
22 additional assertions give33 in that script and131 in the full scoped suite.
CoreObservedUTC separately captures the core snapshot before any later diagnostic
wait; legacy ObservedUTC still describes inspection completion. A real native
live-child test protects this ordering and owned cleanup.
Literal image-error boundary rows are controlled inputs, not proof that those test
children naturally produced Windows image errors. No product full-suite/fuzz ran.

One authorized copied-Go version probe, shutdown-probe-001, ran2026-10-14 UTC17:51:32.
It did not hit the running+missing-image target: the owned running member had a
readable conhost.exe image, so observation was skipped as not-needed. Exit0,
bounded owned cleanup verified; acceptance deliberately refused. This cannot
identify old failed members or explain their image errors. No second probe,
remaining fuzz execution, guard relaxation or acceptance follows from a no-hit.
Probe receipts/caches remain local ignored data, not recoverable from Git.
The single probe preceded the CoreObservedUTC addition; no probe was repeated to
populate that extra field. Final native/serialization behavior is component-tested.

Current checkpoint2026-10-14: baseline passed once; framework copy now instrumented.
The approved exited-state/image split and one same-handle confirmation are verified.
observe-prep-002 and parent-first actual fuzz passed their diagnostic checks;
synchronized stopped during version preparation with the confirming wait still
running and missing image metadata. synchronized/crash/cancel actual fuzz remain
UNRUN. No retries, old failure reclassification or main merge.
[Review/dispositions](../../roadmap/2026-10-14-isolated-fuzz-checkpoint-review.md).
Entries below preserve historical checkpoints, not the latest state.

The latest continuation passed109 PowerShell assertions. Member receipts retain
InitialState/InitialExitCode/InitialExitCodeRead and ConfirmationState/UTC/attempt;
the producer performs only one zero-wait on the same retained, job-verified handle
before closing it. A signaled confirmation refreshes the exit code; running with
missing image, failed queries and unknown ownership remain failures. No new delay
or budget is introduced. same-handle.tests.ps1 uses real owned process handles.

The explicitly approved call was `Invoke-DiagnosticArm <verified-run-root> observe
-PreparationAttempt prep-002`. Preparation identities must match prep-NNN and
require an existing failed nonbaseline preparation. Original or any prep execute
directory prevents another execution; duplicate identities are rejected. This is
not authorization to generate retries. Failed synchronized version preparation
is immutable; further execution needs a separate bounded investigation/approval.
Fresh-PC recovery reads the execution report before creating any new run root;
never repeat completed baseline/observe/parent-first merely to restore progress.

Latest bounded revision: exact verified ownership + signaled exited state +
successful exit-code query permits missing image metadata. Unknown ownership,
state/query/unread exit still fails; missing-image running members do not pass.
ShutdownImageEvidenceComplete is recorded separately from ownership/liveness
completeness, with policy tag owned-signaled-exit-image-optional-v1. Budgets and
forced-cleanup refusal unchanged.9 state regressions RED→GREEN,90 PS assertions PASS.
New component gate8 logic tests and completion passed, including an exited owned
member with ImageError31 whose metadata stayed missing. Old failures not rewritten.
The first observe preparation stopped at go-env origin: initial wait state running,
later exit code0, image error5. Non-atomic reads do not prove a signaled exit; current
guard refused/owned cleanup verified. No fixture build or actual fuzz execution.
Five actual arms remain UNRUN; observe now has a failed preparation directory.

Use PowerShell on Windows. Read the approved2026-10-14 spec/plan and AGENTS.md.
This directory currently contains the verified isolation utility, not a complete
framework reproducer. Never run scripts against a project/cache root as target.

Checkpoint: context gate and trace templates passed7 Go tests; materialization
passed8 assertions. coordinator.patch now contains both context and actual
internal/fuzz tracing changes. The runner is implemented but its real Go build
handoff is blocked: baseline compilation exited0 while job accounting reported1
remaining process at root exit, cleaned to0 afterward. No actual fuzz arm ran.
Do not retry baseline or run later arms without new bounded authorization.
The user subsequently approved bounded runner repair and one new baseline
preparation, identity retry-001. Historical receipts remain unchanged. Use
`Invoke-DiagnosticArm <verified-run-root> baseline retry-001` only under that
approval: it creates arms/baseline-retry-001 and binaries/plain-retry-001 with
fresh caches, refuses duplicate identities, and cannot retry other fuzz arms.
Before reusing an unmodified copy, revalidate its complete manifest and links
against the original copy receipt; never adopt the failed build's binary.
Do not treat these component tests as a full reproduction.
The diagnostic gate deliberately pauses propagation and is never an official
toolchain patch. Trace IDs preserve equality for comparable errors without
printing payloads; non-comparable errors are explicitly unsupported, not silently
classified. Observation mode releases the propagation gate immediately.
Final return marks child/old state unread and stoppingRead=false; stop captures
the original operands. Source parent-done names the actual ctx.Done branch;
incoming deadline/canceled distinguishes the cause without extra Err reads.

run.tests.ps1 currently passes16 assertions, including actual descendant cleanup,
wrong-mode/missing-worker evidence rejection and wrapped child-only environment.
process-native.cs launches suspended into an owned, non-breakaway Windows job,
whitelists inherited std handles and records group accounting. Its baseline
remaining-count interpretation still needs investigation; unit passes are not
proof of real Go integration. That checkpoint also recorded a compile-log
path-escaping incompatibility (raw doubled backslashes), repaired below without
bypassing provenance guards. Keep the failed launch/build receipts unchanged.

Current repair: live owned-job membership is recorded at root wait and after
cleanup. Descendants exclude only the exact root creation ID, while raw active
accounting is retained separately. Cleanup still requires both accounting and
member list empty, within the existing5s cleanup bound; query failure is fatal.
Observed genuine descendants remain a failed gate even when cleanup succeeds.
This does not establish what the historical remaining1 represented.
Compile evidence now requires exact decoded Windows tokens for the copied
compiler, correct package and source in the same command, not loose substrings.
Run runner-regression.tests.ps1 (13 assertions) and run.tests.ps1 (18 assertions,
including a root that exits while an owned child remains) to verify the repairs.
Raw membership uses [JobObjectBasicProcessIdList](https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_basic_process_id_list).

retry-001 outcome: pristine copy revalidated11536 files, baseline build and
actual5s fuzz passed,238579 executions, no remaining members and no termination
needed. Original full11536-file manifest and701 product hashes remained unchanged.
A fresh controls-only model run printed11 preserved controls but failed the
process gate: after the root exited, the owned job still listed one different
member. Owned cleanup reached empty/zero. Member image/liveness was not captured;
do not query recycled receipt PIDs or treat cleanup as acceptance. No further
arms ran. Task3 remains blocked and final review pending. Once any baseline
retry has launched fuzz, the retry helper refuses additional attempts.

Bounded live-identity investigation now records MemberEvidence before cleanup:
open only IDs from the current job list with limited query/synchronize rights,
verify exact-handle job membership, then read image and zero-wait state/exit code.
Query failures and ownership loss remain unknown; they never clear existing gates.
At most32 identity records are collected. Optional final boolean argument to
Invoke-RecordedProcess enables MembersDuringRun with10ms polling; default remains
off/100ms. This timing-perturbed observation is not a natural fuzz/acceptance arm.
State and exit code are separate reads, not an atomic snapshot; exit code0 does
not prove the prior wait observation was exited. No archived PID is queried.

Two finite model-only observations were recorded, no baseline rerun: first had
no members available at capture; opt-in live observation captured an owned
C:/Windows/System32/conhost.exe during root execution and at root exit. It again
failed DescendantsRemaining1, then owned cleanup reached zero/empty. All11 model
controls were preserved. This identifies this new failure's member, not the old
ones, and does not establish why the console host was created or its exit timing.
See [Microsoft's console-host definition](https://learn.microsoft.com/en-us/windows/console/definitions).
Current tests:20 runner assertions+13 regression assertions passed. Later arms
and final review remain pending. Any change to natural shutdown waiting, launch
flags or member acceptance needs a separately approved bounded design.

Approved bounded shutdown revision (schema2): after root exit, observe for at
most500ms within the existing shared5000ms shutdown clock, including evidence
collection and forced cleanup. No named-process exemption; unknown or truncated
shutdown evidence, timeout, or any requested termination cannot pass. Scheduling
and native-call latency are not hard real-time guarantees; actual elapsed times
are recorded and an over-budget successful receipt is refused.
Initial DescendantsRemaining/ActiveAtRootWait/MembersAtRootWait remain raw facts.
NaturalShutdownVerified plus zero accounting/membership at the natural decision
and final check determine completion; forced CleanupVerified is only cleanup.
Schema-less historical receipts retain the old strict remainder gate, not a
retroactive reinterpretation. Version/origin/build/arm checks share this validator.

Real short-lived child and natural-decision validation RED→GREEN; long-lived
children/timeouts still fail. Final31 runner+13 regression+8 materialization
assertions passed. One new controls-natural-001 run preserved all11 model controls;
initial owned conhost remainder1 became zero/empty naturally in25ms, no termination
requested, schema2 completion passed. It does not rewrite previous failures or
explain console-host creation/historical fuzz failure. Baseline was not rerun;
later five arms and final review still pending. Necessary source/report files
are recovery artifacts; machine receipts/caches/binaries remain excluded.

Fresh-computer static prerequisite check (no toolchain execution, copying,
downloads or generated receipts): dot-source preflight.ps1, then call
`Test-DiagnosticPrerequisites <absolute-source-toolchain-path>`. It checks Windows,
independent pinned go.exe/context.go/fuzz.go, VERSION and portable materials.
Ready only means these prerequisites match; full manifest/link validation during
New-IsolatedToolchain and actual version/GOROOT/compiler evidence remain required.
preflight.tests.ps1 passed4 behavioral assertions, including no generated files.
On another PC, pass its pinned toolchain location with `preflight.tests.ps1 -Source
<absolute-path>`; the test does not require a Lenovo user profile.
The test uses a tiny source fixture; it is not a complete11536-file toolchain.

Current framework continuation stopped before observe: patch installed only in
framework-001. Context5 and trace3 logic tests passed; both compiles passed. The
trace test process exited0 but its shutdown gate failed because a verified owned
member was already signaled/exited with code0, while image lookup returned31.
Current policy treats that missing evidence as failure and forced owned cleanup
then verified0/empty. Image was not captured; do not infer conhost from previous
runs. All five actual diagnostic fuzz arms remain UNRUN; no gate/arm retry or
baseline rerun. Full original11536-file manifest and701 product hashes unchanged.
Resolving known exit state versus missing image identity requires a bounded
policy/design decision, not bypassing the guard. Independent checkpoint review
does not turn these unrun arms into acceptance.

Windows API references: [CreateProcessW](https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-createprocessw),
[AssignProcessToJobObject](https://learn.microsoft.com/en-us/windows/win32/api/jobapi2/nf-jobapi2-assignprocesstojobobject),
[job limits](https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_extended_limit_information).

After verifying a new independent copy and the pinned context.go hash
765F99F7CC0710B0117FFA183E432524F429C21496FCBAD9B63FB9886FDBAAC7,
materialize context-diag.go.txt/context-diag_test.go.txt as diagnostic_fuzz.go/
diagnostic_fuzz_test.go inside copied src/context, and apply the context patch
only there. This hash-pinned patch uses zero-context hunks: require the exact
source hash and Git's --unidiff-zero flag (never apply it to drifted sources).
The reverse check against the tested copy succeeds. Build context's test executable with the copy's explicit GOROOT,
independent GOCACHE and offline process environment from the approved plan.
Run with quoted native arguments '-test.run=^TestDiagnosticFuzz', '-test.count=1',
'-test.timeout=30s'. In PowerShell, unquoted dotted arguments can be misparsed.
Registration rejects unsupported/canceled/duplicate contexts; a forgotten gate
releases after1s and records watchdog expiry. These5 context tests are not
coordinator/worker/RPC or original project acceptance results.

Run `powershell -NoProfile -File docs/diagnostics/isolated-fuzz/isolation.tests.ps1`
from the development worktree. Tests generate unique tiny fixture directories
inside the owned diagnostic root and retain them. Real link tests report UNRUN
if privileges prevent creation; UNRUN is not PASS. No external test library needed.
Run materialize.tests.ps1 for actual patch replay/source-drift guard tests using
tiny source-only fixtures. Install-DiagnosticPatch requires both exact source
pins, rejects official/outside targets and existing template files, then installs
all four templates and the patch only into the validated copy.

Dot-source isolation.ps1, then use New-IsolatedToolchain with an existing source
directory and a nonexistent unique RunRoot beneath
.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/. The function checks
paths, links and independent source/copy manifests before returning. It does not
run or patch the copy. Do not substitute another toolchain or download dependencies.
It now saves copy-receipt.json in the new RunRoot after verification, via CreateNew.
Do not manually overwrite it. Saved/returned Source/Toolchain/Cache and manifests
match; Invoke-DiagnosticArm consumes that receipt. An incomplete copy is unusable.
The fixed version is Go1.26.5 windows/amd64, independently prepared on a fresh PC.

Generated copies, caches, fixtures and receipts are not recovery artifacts and
must not be committed. Rebuild from tracked scripts/templates; source hashes,
actual executable/GOROOT and runtime results still require verification. Original
matrix3 failure and Pair-unrun remain open; no product acceptance or main merge.

provenance.ps1 requires the exact installed6-file set/current4 templates/patch and
binds binary receipts to canonical source identity. Missing/stale identity refuses
reuse without rebuilding or retrying an existing arm.7 provenance assertions PASS
including template-only drift. Old baseline results remain historical facts, not
permission to reuse unbound binaries.

# Ordinary leaf verification entry — fake implementation verified

The [approved design](../../superpowers/specs/2026-10-16-leaf-ordinary-check-design.md)
and [fake-only implementation plan](../../superpowers/plans/2026-10-16-leaf-ordinary-check.md)
introduce a fixed default-vet ordinary nodeleafbinding entry. Read the
[actual implementation record](../../roadmap/2026-10-16-leaf-ordinary-check-implementation.md)
for final8-suite407 checks per PS7/5.1 host (1 static included each),all16 commands
exit0,71 source/test pins,one review/two verified repairs and remaining real gates.
No real invocation is authorized;never reuse the build-only binary.
