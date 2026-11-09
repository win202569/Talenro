# SDD ledger — plan: docs/superpowers/plans/2026-11-09-windows-c4-observer-entry.md

User approved inline execution after5a13d26b. Baseline5a13d26b, Task1 in progress.
Pre-flight: one integrated task; no shared interfaces between tasks. Existing linked
worktree verified, expected development branch, no superproject, clean baseline.
Fresh baseline: observe-c5=32, c5-observation-session=120, guard=48, managed=32,
native=78; all exit0 in separate PowerShell hosts (310 assertions).

Ruling: bundled workspace/task helpers require Bash; use this portable ledger and
read the task directly in PowerShell instead. Cost: manual task bookkeeping.
Ruling: verification stays within the approved synthetic-only boundary, overriding
generic full-project build/setup guidance. Cost: real Windows/product integration
remains unverified. No installs, .NET build/test or live observation authorized.

No scratch workspace created. No native launch allowance.

2026-11-10 implementation: default/policy/preflight RED missing-entry assertion,
then GREEN34. Receipt factory RED missing behavior, then GREEN105. Adapter/integration
RED missing backend, then GREEN152. All runs separate PowerShell7 hosts, exit1 on
each intended RED and exit0 on GREEN; no actual native/CIM/launch/sleep.
Fixture corrections: PowerShell array multiplication needed parentheses; shared
failure uses Kind=outcome, Outcome=observation-failed (not Kind=failure). Corrected
expectations to the unchanged shared contract. No production changes for these.

Manifest serialization and cleanup precede process acquisition; tests exercise
collision, invalid IDs/context, path/reparse checks, acquisition/write/flush/disposal
failures, exact manifest values and correlated shared synthetic evidence. Boundary
wrappers isolate filesystem writes; OS behavior and adversarial races not proven.
Final independent review pending. Checkpoint is not live/UI acceptance.

Task1 implementation complete (base5a13d26b; final verification command pattern
`pwsh -NoProfile -File clients/windows/tests/<name>.tests.ps1`): observe-c4=152,
observe-c5=32, c5-observation-session=120, observation-session-guard=48,
process-observation=32, native-process-observation=78; total462, all exit0.
Both new scripts parse with zero errors; git diff --check passes; C5/shared runtime
and tests have empty diff against5a13d26b. Eight explicit necessary source/test/doc
files only in checkpoint. Independent review range is5a13d26b..checkpoint HEAD;
review this ledger's Ruling lines, full plan/spec, guide and status/recovery changes.
