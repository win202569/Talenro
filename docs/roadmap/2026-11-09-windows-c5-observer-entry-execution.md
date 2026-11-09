# Fixed C5 observer entry execution

Plan docs/superpowers/plans/2026-11-09-windows-c5-observer-entry.md approved after
0f87ad74; implementation baseline0f87ad74. Native launch budget0.

Task1: complete (checkpoint synthetic111/111; helper48/32/78 exit0).
Task2: complete (final entry32, orchestration120, helper48/32/78; all exit0).
Task1 callback/receipt contracts feed Task2;
no shared-interface conflict found at preflight. Existing linked worktree/branch
verified; no dependency installation needed for PowerShell synthetic checks.

Ruling: bundled workspace/task scripts require Bash, while this workspace requires
PowerShell only. Use this portable ledger and direct PowerShell verification rather
than invoke Bash scripts; task scope/RED/GREEN/review evidence remain recorded here.
Cost: no automatic task-brief bookkeeping; plan is read directly before each task.

Task1 RED: explicit missing orchestration assertion. During GREEN, a PowerShell
closure could not resolve the module-local imported guard; bound its CommandInfo
before closure creation. A nested synthetic fixture closure lacked its start-time
capture; corrected the fixture to capture local values. Diagnostic exception output
was enabled only during synthetic investigation and restored to sanitized errors.
Final first-task checks: c5-observation-session111, session-guard48, managed32,
native synthetic/static78, all exit0, native helper not loaded. No real launch.

Task1 checkpoint80a0106c. Task2 RED: missing fixed-entry assertion; GREEN32 entry
assertions using AST-extracted private definitions and recording command doubles.
Additional orchestration RED: cleanup overrun incorrectly preserved timely success;
GREEN113 after post-cleanup clock validation. Raw outcomes are explicitly provisional,
final completion only after cleanup. This refines receipt metadata without changing
the ten callback names or fixed scope. Live adapter behavior remains unverified.

## Final independent review and one fix pass

Fresh-context reviewer c5_entry_final_review inspected baseline0f87ad74..80a0106c
plus Task2 working files, reused adapters, plan/spec and this ledger. Read-only;
reran orchestration113 and two in-memory reproductions. No live operations.
Verdict: no Critical, two Important, one Minor; ready with fixes, no merge approval.

Final: fixed stale failure timing — delayed-CIM test RED (recorded0 instead of
360050), then GREEN; validated clock refresh before failure receipt and after cleanup,
including failed cleanup. Invalid clock keeps failure, with TimingVerified=false.
Tests verify cleanup failure reports360060ms/60ms overrun and both disposal attempts.

Final: fixed native evidence lost on managed exception — signaled-native regression
RED, then GREEN; incomplete raw pair survives managed/clock acquisition failure,
without retry or acceptance. Both original resources still disposed.

Final: minor (deferred): observe-c5.tests.ps1 native-helper absence assertion names
the wrong type. It is redundant with the orchestration and native suites' correct
type assertions; no runtime code defect. Keep as a documented small test gap, per
the approved one-pass workflow (Minor findings are deferred, not included in fixes).

Final: Ruling: real Windows launch/CIM/native behavior and process races remain
unverified — current authority is synthetic-only; separate live approval required.
Cost if wrong: OS-specific integration defects may surface on the later bounded run.

Final: Ruling: C5 UI, RTL, shutdown root cause and operator close timing stay outside
implementation acceptance — no UI/product behavior changed. Cost if wrong: observer
checks alone cannot establish functional or shutdown acceptance; existing gates stay.

Final: Ruling: no hard real-time or adversarial filesystem-race guarantee — synchronous
OS/writer/cleanup/transport may block; destination protection assumes a trusted
workspace as designed. Cost if wrong: a delayed call can overrun, or a hostile racer
can defeat path checks; do not use this entry as a security boundary.

Final: Ruling: current-status/recovery docs reviewed by executor, not independent
reviewer — reviewer explicitly set pending documentation aside. Updated to match
verified counts and remaining boundaries. Cost if wrong: handoff text may contain
an author-only oversight; source/tests and this evidence remain authoritative.

## Verification and preservation

Separate PowerShell7 hosts, command pattern:
`pwsh -NoProfile -File clients/windows/tests/<name>.tests.ps1`.
Final names/counts: observe-c5=32; c5-observation-session=120;
observation-session-guard=48; process-observation=32; native-process-observation=78.
Total310; all exit0. Four new runtime/test files parse; git diff --check passes.
Default and private-function tests use doubles, not live adapters. No .NET test,
build, native helper load, app launch, settings change or product/binary edits.
No raw files, credentials or ignored outputs in the checkpoint. No scratch workspace
was created/deleted. Preserve existing worktree and dev branch; no main merge/PR.
Commit/push follows AGENTS standing instruction; remote SHA is verified separately
after commit. C5 remains NOT_ACCEPTED, native launch budget0.
