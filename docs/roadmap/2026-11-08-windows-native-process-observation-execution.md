# Native recorder execution ledger

Plan: docs/superpowers/plans/2026-11-08-windows-native-process-observation.md.
Approved after85169c27; baseline85169c27, clean linked worktree, no submodule.
PowerShell7.6.9; original recorder baseline32/32 synthetic PASS.
Task1 in progress; Task2/3 pending. No live/native execution allowance.
Pre-flight: Task1 raw payload normalization feeds Task2 DTO adapter; Task3 uses
both test receipts. DTO time/exception metadata must not be replaced by adapter
timestamps or rethrown wrappers. Preserve native helper timing/type verbatim.
Ruling: use this portable ledger instead of Bash-only workspace/task scripts —
Windows PowerShell-only requirement and recoverability take precedence; manual
task completion/commit evidence replaces script bookkeeping.
Ruling: run only approved synthetic/static/compile checks, not the full app suite
or dependency setup — explicit scope forbids live calls/builds; broader behavior
remains unverified. Static safety checks supplement manual review, not runtime proof.

Task1: complete. Missing-module RED (exit1) then48/48 synthetic GREEN (exit0).
Commands: & ./clients/windows/tests/native-process-observation.tests.ps1.
Only normalization/synthetic export exists at this checkpoint; native helper not loaded.
Task2 in progress; Task3 pending. No native API/application calls.
