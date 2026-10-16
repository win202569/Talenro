# Leaf stage build-only implementation — 2026-10-16

Status: approved Native/inline fake-only implementation,one fresh review/two
Important repairs and final full regression complete. Final preservation in
progress;not live/ordinary acceptance.

## Implemented boundary

The [approved spec](../superpowers/specs/2026-10-16-leaf-stage-build-only-design.md)
and [three-task plan](../superpowers/plans/2026-10-16-leaf-stage-build-only.md)
are implemented in a separate fixed planner,scope validator and controller.
Only this diagnostic requests one -vet=off argument. Existing default/matrix,
eager/stage controllers,collectors,C# bodies,32 cumulative admissions,600s native
deadline and cleanup policies are unchanged. The common ordinary completion
guard rejects every present BuildOnlyDiagnostic property,even null/false.

The controller has one replaceable native boundary,never a binary runtime or
second vet invocation. It saves scope metadata,independent trigger/core/lifetime
facts,SDK/product/source/artifact audits,provisional result and bounded artifact
manifest. Only a successfully persisted manifest-bound completed marker establishes
diagnostic finalization. VetEvidenceComplete,Acceptance,PackagePassed,
OriginalTestPassed and RuntimeAttempted remain false in every outcome.

Completion is written/closed and byte-verified as completed.pending.json before
a same-directory,non-overwriting move to completed.json. A failed candidate may
remain locally;it is excluded from the bounded manifest and is never a completion
marker. Do not accept it or provisional result.json as finalized evidence.

## Evidence and history

Base8a25712f816ee6f8ffe6d3f0695e7df436f2160c. Task1
1c13f5fb33bc7d5b87214adc2815252457228575 adds planner/source identity;
Task2 4bd4054f adds fixed scope/ordinary refusal. Controller and final repair
checkpoint:a4bfe80ab898e0b6ada5a033a44534d0ff05a9e6. Both first tasks had named RED tests.
Task1:60 checks per PS7/5.1 host. Task2:four suites/133 checks per host
(6 static included each),all8 commands exit0. Task3 named
BuildOnlyControllerExists RED exited1 before implementation. Focused controller
tests now pass79 checks per host (one static included each),both commands exit0.

One fresh whole-increment reviewer found2 Important,0 Critical and0 Minor issues:
secondary persistence masked primary errors,and direct marker creation could
leave completed.json after failure. Both were reproduced with named RED tests;
one repair pass preserves primary/secondary failures,attempts result even after
stopped-write failure,and stages/verifies completion before publication.
Focused native/final/marker persistence tests passed on both hosts (6 commands
exit0);marker faults cover creation,partial write and post-close failure. No
second review was dispatched and no minors were deferred. The first incomplete
final batch was cancelled before source edits and retained separately;it is not
a final passing batch. Post-fix final results:

| Host | Suites / exit0 commands | Checks | Static included |
| --- | ---: | ---: | ---: |
| PowerShell7 | 21 / 21 | 884 | 11 |
| Windows PowerShell5.1 | 21 / 21 | 884 | 11 |

All42 commands exited0. New suites contribute60 planner,41 scope/guard and85
controller checks per host (one controller static included);existing18 suites
contribute698 checks (10 static included). Final66 ps1/cs source/test pins matched
after both batches. Six changed PowerShell files parsed;154 local documentation
targets exist;Git whitespace checks passed. Run each listed test in a fresh
NoProfile/NonInteractive host,with no focus/live switches,using the plan's exact
21-file list. No whole-product Go baseline or live test is represented by these counts.

Raw verification logs/bindings remain ignored locally;archive finalization and
remote checkpoint verification are the remaining preservation work.
All tests use fake native/SDK/product boundaries;portable new fixtures
are tiny files and synthetic DTO/logs,not ignored prior receipts or caches.
Legacy planning tests may depend on this machine's existing SDK/module paths.

## Execution rulings

1. Use PowerShell equivalents of the workflow's Bash bookkeeping scripts:
   Windows-only execution;cost is manual ledger/brief evidence checking.
2. Run only the approved fake baseline/final batch,without setup/download/full
   product Go or real SDK inventory;cost is unverified whole-product/live behavior.
3. Keep current source verification inside saved scope validation. There are
   four verifier calls including three lifecycle audit phases,not the plan's
   original three;cost is extra file checks and corrected spy expectations.
4. Add test-only writer fault injection for create/write/close failures while
   retaining real IO for normal writes:directory obstruction alone cannot cover
   post-create faults;cost is that simulation cannot establish real disk reliability.
5. Keep actual vet scheduling,capture completeness,timing,error87 and live SDK/
   product behavior outside the review/verification boundary;cost is unverified
   live reliability and broader acceptance.
6. Do not claim receipt authentication or hostile-concurrency/OS sandbox
   protection,as explicitly excluded by design;cost is inability to authenticate
   a hostile host.
7. The executor verifies final test/remote state after review,which observed
   them pending;cost if checked incorrectly is a false green/recovery claim.

## Recovery and remaining gates

Recover from the latest codex/c12-b01-task9-coordinator tip,not automatically main.
Git carries necessary source/tests/spec/plan/status/recovery,not raw logs,toolchains,
modules,caches,binaries or conversations. Dependencies and any real diagnostic
require separate preparation and approval. No new Go/native Run,PID query,SDK/
product inventory,Docker,installation,binary runtime,main merge or live vet occurred
in this increment. Historical compile001 remains STOPPED;error87 and broader
C11/C12/Task10/B01/production/main acceptance gates remain open. Do not replay001,
reuse its cache or infer that vet-off completion grants required vet acceptance.
