# Fixed C5 observer entry execution

Plan docs/superpowers/plans/2026-11-09-windows-c5-observer-entry.md approved after
0f87ad74; implementation baseline0f87ad74. Native launch budget0.

Task1: complete (synthetic111/111; helper48/32/78 exit0). Task2: in progress.
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
