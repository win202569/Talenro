# Ordinary observer implementation — unfinished checkpoint

Date2026-11-13; base4ec59956; branch codex/c12-b01-task9-coordinator.
User approved entry implementation and synthetic tests, no application launch.

## Changes

Added separate observe-ordinary-account-center.ps1 and .tests.ps1 under
clients/windows/tests. Entry adapts C4 wrapper: six artifact hashes, fixed build
commit/path-scoped diff, reparse artifact checks, no ArgumentList, ordinary
manifest and separate receipt directory. Existing product, shared observers,
original entries/tests remain unchanged. This entry is NOT validated for live use.

## Actual tests and stop

PowerShell7.6.9, fresh host each invocation:
`pwsh -NoProfile -File clients/windows/tests/observe-ordinary-account-center.tests.ps1`.

1. RED: exit1, expected `FAIL: RED: Ordinary entry missing` before implementation.
2. First GREEN attempt: exit1, unexpected `Get-Item: Missing an argument for
   parameter 'Force'`. Stopped without correction/retry as checklist requires.

Read-only cause inspection: test-local Get-Item mocks declare untyped `$Force`,
which requires a value, while real artifact preflight correctly invokes the
switch `-Force`. Replace with `[switch]$Force` in the new test's mocks in an
approved bounded continuation; do not weaken production reparse checks or edit
the original C4 suite. No claim that this is the only remaining defect.

Failure occurs at initial synthetic preflight before mock process acquisition.
No App launch, real CIM/native helper/sleep, restore or build executed. No raw
candidate/receipt changed. Tests have not reached all assertions; no PASS count.

## Remaining work

Task1 incomplete. Finish checklist coverage including per-source/per-product-path
faults, default side-effect coverage, CreateNew verification, combined cleanup
failures, strict ordinary command-line negative cases and fake-clock deadline
integration. Existing test cases beyond the first failure are unverified.
Task2 five-suite regression and independent scoped review have NOT run.
No new live allowance; C4/security/Release/runtime acceptance unchanged.

## Execution rulings

- Bash-only skill bookkeeping replaced with PowerShell/read-only brief plus
  apply_patch ledger to honor the user's PowerShell constraint; risk: bookkeeping
  omission, not altered safety checks.
- No full-project/build test command: explicit synthetic-only approval takes
  precedence; broader regression coverage remains unverified.
- Retain plan workspace/evidence at unfinished handoff; no cleanup or original
  evidence deletion. Fresh clone restores source/docs, not ignored evidence.

Next requested scope: correct the new test mocks, complete pending assertions,
one focused synthetic attempt, then only if successful the five planned regression
suites and scoped review; stop on unexpected failure, no application launch.
