# Ordinary identity ledger — stopped unfinished implementation checkpoint

Date:2026-10-17. Base89cb1ad09bf831216c5999ddcb4409e4cea4d18d.
User approved the [plan](../superpowers/plans/2026-10-17-ordinary-member-identity-ledger.md),
maximum30 suite invocations,and Native/inline with one final independent review.
Status:Task1 incomplete;stopped under the approved unexpected-failure rule.
No accepted implementation,final regression or independent review is claimed.

## Delivered so far

Four new portable files under docs/diagnostics/isolated-fuzz:
ordinary-ledger.cs,ordinary-ledger-fixture.cs,ordinary-ledger-fixture.ps1 and
ordinary-ledger.tests.ps1. Acquisition/temporary ownership/disposal code exists;
finalization deliberately refuses until Task2. Existing process-native.cs,
run.ps1,diagnostic registry,preflight,Go source and callers are unchanged.
No schema4 path is wired or enabled. This checkpoint is not a green checkout.

Existing isolated worktree/branch reused;no dependency download/install or
baseline wildcard suite. Windows PowerShell5.1 and PowerShell7 hosts exist.
The fixture loader compiles declarations/fakes without invoking native APIs.
Shells/Add-Type compiler processes are not native target workload validation.

## Exact execution result

Only ordinary-ledger.tests.ps1 ran;the original metadata/probe identities were
not touched. All three invocations completed,nothing remains running.

| Invocation | Host/intent | Result | Elapsed |
| --- | --- | --- | --- |
|1 |PowerShell7/RED |exit1,expected missing ordinary-ledger.cs assertion |753ms |
|2 |Windows PowerShell5.1/RED |exit1,same expected missing implementation |538ms |
|3 |PowerShell7/first GREEN attempt |exit1,unexpected ASSERTION FAILED:Cap32 |2073ms |

Invocation3 compiled and emitted PASS for AcquisitionOwnsOnce,
OwnHandleReleasedOnce,Open87Sticky,ForeignReleased and CrossRunIdentity,then
stopped at Cap32. No InvalidCandidate/CloseFailureOnce result or PS5.1 GREEN
result exists. Do not call Task1 complete from those five PASS lines.
No task2–4 suite,legacy final regression,native adapter/new entry,Go build,
target runtime,metadata continuation or final review ran.
Used3/30 suite invocations;27 remain unused,not permission to retry this failure.

## Readonly cause trace and proposed bounded next decision

The Cap32 test assigns Candidates=2..34:33 numeric IDs. Its fake QueryMembers
returns Assigned=Returned=33 but hard-codes Capacity=32,BufferBytes=264.
The production ValidList condition requires Capacity>=Ids.Length,so this
malformed envelope is refused before any acquisition can exercise the32-attempt
limit. The assertion expecting32 opens therefore does not test the intended path.
This is a source-level trace of the saved failure,not a successful repair/retest.

The indicated correction is confined to the fake's list capacity/byte envelope
so it can represent all returned IDs. Preserve the33 candidates,32-open assertion
and production malformed-list refusal. Do not loosen ValidList or cap checks.
No correction/retry was performed after the stop. Production behavior beyond
the executed assertions remains unverified and may reveal other issues later.

Proposed authorization:make that fixture-only correction and increase suite
budget30→31 by one invocation,then run replacement GREEN7 and the first GREEN5
before continuing the remaining original plan. Arithmetic:3 already used +2
Task1 GREEN calls +12 Task2–4 RED/GREEN calls +14 final allowlist calls=31.
Any further unexpected failure again stops;no automatic extra retry allowance.
No native workload,new rights,wait extension or acceptance activation proposed.

Saved-only fingerprints (SHA256,after stop):

| Ignored artifact | SHA256 |
| --- | --- |
| invocations.json | B29AC0FD786269AE1219B7D33A04137D8A6B885E95FEE867E5CCBCC2743208A5 |
| run-3.stdout.txt | EC89FD07460012A8B59BF1234FED5AA040896D4295CD6DCF35533EA1CCFA862A |
| run-3.stderr.txt | 027CCCE4AF698EDD33B611353BD82E450A8895B6A94CEAA4CA4CFB177F79E68B |

## Execution rulings preserved from the ignored ledger

- PowerShell equivalents replace Bash-only skill bookkeeping scripts because
  this project requires PowerShell. Cost if wrong:manual bookkeeping reconciliation.
- Task GREEN is task-done evidence without an additional helper rerun;no broad
  baseline/install. The approved30-call scope controls. Cost:broader regressions
  are not evidenced;only the eventual explicit allowlist can be reported.
- Task3 permits run.ps1 refusal-only edits despite the generic file map saying
  Task4 only;the task-specific instruction/spec controls. Cost:earlier refusal
  diff requires review. Task3 has not started.

## Recovery and gates

Portable source/tests/spec/plan and this failure report are committed on
codex/c12-b01-task9-coordinator. Ignored execution root is
.superpowers/sdd/2026-10-17-ordinary-member-identity-ledger;it holds immutable
invocation records and raw output. Do not rewrite/delete its records or replay
earlier invocations. A new computer restores portable code/report,not these logs.

Resume at the bounded fixture-correction/budget decision above. Task1 remains
incomplete and Task2–4 unstarted;final independent review still required.
Original metadata remains1 build/1 runtime used,RED unaccepted;retained probe
remains0 builds/1 runtime consumed. No ordinary rollout,metadata/main/production
acceptance or merge. Preserve all original failure gates and consumed identities.
