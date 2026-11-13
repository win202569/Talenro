# Ordinary observer implementation — unfinished checkpoint

## Latest: two receipt verification gates closed by approved synthetic run

Continuation after3935c670: user approved only the two missing tests and one
synthetic run, no App launch. New tests invoke the actual Open-OrdinaryNewStream
for both manifest.json and samples.jsonl in a fresh GUID fixture directory under
ignored .superpowers/sdd/2026-11-13-windows-ordinary-minimal-observation/fixtures.
Each creates one byte42, closes it, attempts reopening, then verifies rejection
and unchanged byte content. Two tiny fixtures retained; no original receipt or
candidate modified. This is real local fixture I/O, not an entirely memory-only
test; process/native/CIM/sleep effects remain simulated.

Four direct factory cases inject simultaneous failures: manifest write or flush
plus writer and stream disposal; manifest or samples writer construction plus
stream disposal. Assertions check AggregateException counts2/3, primary first,
each cleanup error retained, cleanup continues after writer failure, all acquired
resources receive disposal attempts, no successful writer and no acquisition.

Single fresh PowerShell7.6.9 invocation of observe-ordinary-account-center.tests.ps1:
exit0,598 assertions PASS (previous564 plus34). No second run, broad regression,
review dispatch, App launch, native compile, restore or build in this continuation.
Production entry/shared files unchanged; five-suite results below are prior-run
evidence, not repeated today. Both Important review coverage gates now have
direct passing tests. This does not fix or dismiss the four deferred Minors,
especially the inherited AutoFlush initialization/disposal diagnostic limitation.
No new production behavior was added, so this is test-gap closure, not a claimed
RED/GREEN production fix. Previous failures remain recorded below.

Next: separate live approval, with current binary/source/path/process preflight
still required, at most one visible run/zero retries/360000ms, close by300s.
No current live allowance or runtime/C4/security/Release acceptance. Exact viewport
and host command-line serialization still require their own evidence.

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

### Authorized continuation after65e545bc

User approved correction, one focused attempt and conditional regression/review,
still no App launch. Corrected three new-test Get-Item mocks to `[switch]$Force`;
production entry unchanged. Added per-observer-source and per-product-path faults,
untracked dirty state, fixed source/diff invocation assertions, public parameter
and ArgumentList AST checks, strict trailing/extra command-line rejection and
exact-deadline integration with finalization/cleanup assertions.

Fresh PowerShell7.6.9 results, each once, all exit0:

| Suite | Result |
| --- | --- |
| observe-ordinary-account-center.tests.ps1 |564 assertions PASS|
| observe-c4.tests.ps1 |152 assertions PASS|
| c5-observation-session.tests.ps1 |120 assertions PASS|
| observation-session-guard.tests.ps1 |48 assertions PASS|
| process-observation.tests.ps1 |32 cases PASS|
| native-process-observation.tests.ps1 |78 assertions PASS|

Only simulated process callbacks/memory writers used. No App, restore, build,
native loader/compiler tests, real CIM query or live entry invocation. Prior
failure retained above. The original entry/shared sources remain unchanged.
Independent static review requested against4ec59956..65e545bc plus new test diff.
Still require review disposition and explicit reconciliation of remaining
CreateNew/combined wrapper-cleanup coverage before calling preparation complete.

### Independent review disposition

Reviewer ordinary_observer_review, static only, reviewed baseline4ec59956 to
65e545bc plus continuation test diff. No Critical findings or demonstrated
Important production regression. Two Important verification gates accepted:

1. Real Open-OrdinaryNewStream CreateNew mode is not exercised/asserted; mocked
   factory tests would not catch replacing CreateNew with Create (entry89,
   tests149/165). Current implementation is correct but proof is missing.
2. New wrapper factory primary+cleanup failure aggregation is not tested:
   scalar Fault cannot combine errors and mock stream Dispose never throws
   (tests156/205). Shared guard aggregate tests do not close this wrapper gate.

Four deferred Minors (not repaired or retested this turn):
- Entry99 inherits swallowed writer-disposal error after AutoFlush setup failure;
  diagnostic limitation, not a new launch defect. Do not claim unrestricted
  retention of every cleanup exception.
- Tests32 lack independent executable/receipt-root literal expectations.
- Tests51 stub Git helper; real nonzero handling and explicit documentation-only
  HEAD success are not directly demonstrated.
- Tests266 cover exact expiry but not a progressing ordinary loop/disagreement;
  shared suites cover those behaviors, so exact-plan wrapper coverage differs.

Ruling: preserve both Important gates as blockers, no live-ready claim — the
bounded focused/regression allowance is spent — cost is delayed readiness, not
weakened checks. Next continuation should add only the missing receipt tests,
then bounded synthetic verification; no App launch or unrelated Minor fixes.

Reviewer exclusions accepted: actual command-line serialization/UI/normal close
and real exit require separately approved live observation; current candidate
presence/hashes were not rechecked; hostile filesystem races/full dependency
signatures exceed the six-file trusted-workspace contract; broader C4/security/
Release/viewport acceptance remains excluded. Reviewer did not rerun tests;
counts above are main-agent tool results. These exclusions leave corresponding
runtime/security claims unverified, not implicitly passed.

The following paragraph records the earlier65e545bc handoff, not current test results:

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
