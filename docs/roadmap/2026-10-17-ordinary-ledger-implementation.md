# Ordinary identity ledger — stopped unfinished implementation checkpoint

Latest:[review repair execution](2026-10-17-ordinary-ledger-review-repair.md)
stopped3/26 after partial A edits;no correction/rerun after unexpected failure.
All review findings remain open;do not execute ordinary matrix/leaf workflows.
Earlier no-post-review-repair and current statements below are historical.

Date:2026-10-17. Base89cb1ad09bf831216c5999ddcb4409e4cea4d18d.
Current final review:[six accepted blocking findings](2026-10-17-ordinary-ledger-final-review.md).
NOT implementation complete;no fixes or tests after review. Existing ordinary
matrix/leaf receipt workflows are broken by new null fields:do not run them.
Earlier fake-PASS/pending-review/current statements below are historical.
Current:all approved fake/static regression passed;34/34 used,independent review
pending. The following stop/current histories are superseded by the final
regression section at the end. No native/metadata acceptance is claimed.
Latest:Task1/2 fake checks complete;Task3 stopped on NestedRoundtrip failure at
invocation13/32. Detailed follow-up below supersedes earlier current/next states.
Latest follow-up:after468709b6,user approved pure diagnosis/cause-specific repair,
budget34. Run14 isolated String type StartUTC rejection;UTC metadata now accepts
the host JSON parser's UTC DateTime representation only. Identity/boolean/numeric
checks remain strict. Runs15 GREEN7/16 GREEN5 passed all receipt checks
(1791/1981ms). Task3 complete in fake scope,16/34 used;Task4 next. Earlier stop
states are historical;no native/metadata acceptance claimed.
Current:after8dafbfe3,user approved script UInt64 aliases/budget32. Invocation6
GREEN5 exit0/1196ms,9 PASS assertions. Task1 is now complete against scoped
fake tests;Task2 next. Earlier stops below are historical,not current.
Subsequently Task2 completed:invocations7/8 expected missing-session RED;
9 GREEN7 exit0/1317ms and10 GREEN5 exit0/1112ms,14 assertions each. Added
shared-budget finalization/count reconciliation;existing native runner untouched.
Used10/32;Task3 next. This remains fake-only,not native or metadata acceptance.
Latest:the follow-up below records5/31 calls and a new PS5.1 compatibility stop;
the earlier3/30 stop and its proposed next step are historical,not current.
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

## Approved capacity correction;new PS5.1 stop after557c2249

User approved correcting the fake list envelope and increasing the maximum30→31,
then continuing the original plan. Only fake QueryMembers changed:capacity is
max(32,returned IDs),buffer bytes8+8*capacity. Candidate set,assertions and
ordinary-ledger.cs production checks remained unchanged.

| Invocation | Host/intent | Result | Elapsed |
| --- | --- | --- | --- |
|4 |PowerShell7/replacement GREEN |exit0;all9 PASS lines,including Cap32 and remaining cases |1344ms |
|5 |Windows PowerShell5.1/first GREEN |exit1,TypeNotFound [ulong] at line4 Candidates cast |1155ms |

The PS5.1 host loaded/compiled the source and reached the first test body,but
its script type name [ulong] could not be resolved. No PASS assertion was emitted
there. This is an author-introduced test-script compatibility error,not evidence
of a production cap defect. No fix/retry followed the failure;Task1 remains
incomplete,Tasks2–4/final regression/review have not started. Total5/31 used.

Readonly comparison finds the affected aliases only in ordinary-ledger.tests.ps1;
existing run.ps1/member-query.tests.ps1 use [UInt64]. Proposed correction is
script casts [ulong]/[ulong[]]→[UInt64]/[UInt64[]],without changing values,
assertions or production C# types. No additional baseline/native probe is needed.

Proposed next approval:that script-only compatibility correction and cap31→32,
then one replacement GREEN5 and the remaining original plan. Budget arithmetic:
5 used +1 Task1 replacement +12 Task2–4 RED/GREEN +14 final allowlist=32.
PowerShell7 will cover the compatible test text again in the planned final run;
its present9-PASS result belongs to the pre-alias-correction script. A new
unexpected failure still stops. Nothing authorizes native target execution,
ordinary rollout or metadata acceptance.

Current saved-only SHA256 fingerprints:

| Ignored artifact | SHA256 |
| --- | --- |
| invocations.json at5 calls |9E23462EA0BAB7F2BCF0753979614AD9DF120B992AC66165096B4466BE79B0D4 |
| run-4.stdout.txt |0B6F42053070FAB01451739266263589DD1D71B822078F63C7A2C4BF8E37CD94 |
| run-5.stderr.txt |2C302415CB8E51EAA1E88F99340D87BD246C65AB07CC77CDF13298FABF3BE892 |

All5 suite hosts completed. Raw records1–3 retained byte-identical;the invocation
index appends new rows rather than rewriting earlier results. No old metadata/
probe workload was started. Fresh-computer resume begins at this new approval
decision,not the earlier Cap32 proposal or any automatic test replay.

## UInt64 correction verified;Task2 complete;Task3 JSON stop

User approved script-only aliases/budget32 after8dafbfe3. Invocation6 on PS5.1
passed9 assertions (1196ms). Task1 checkpoint442053bf preserves that correction;
production acquisition checks were not weakened. Task2 checkpointeb610080 adds
same-handle finalization/shared clocks/conservative accounting with fake tests.

| Calls | Scope | Actual result |
| --- | --- | --- |
|6 |Task1 GREEN5 |exit0/1196ms;9 PASS |
|7/8 |Task2 RED7/RED5 |exit1/1448ms and986ms;expected missing session |
|9/10 |Task2 GREEN7/GREEN5 |exit0/1317ms and1112ms;14 PASS each |
|11/12 |Task3 RED7/RED5 |exit1/1351ms and1020ms;expected missing evaluator |
|13 |Task3 first GREEN7 |exit1/1756ms;2 PASS then NestedRoundtrip failure |

Task2 tests cover absent-at-root entries,late admission,running/unknown exit,
count mismatch,release before final accounting,new member after sealing,shared
500ms deadline,overflow/backward clocks,forced exit,lifetime regression,failed
release and optional image loss. These are simulated adapter results only.
Ruling:the shared budget measures actual elapsed wait,not the sum of maximum
wait arguments when a prior wait returns early;the next member gets200ms after
300ms elapsed. Cost if wrong:timing interpretation needs review. Ruling:added
shutdown start/end/frequency to the outcome so receipt validation can check time
bounds. Cost:extra serialized fields,not acceptance activation.

Task3 created ordinary-ledger-receipt.ps1 and its tests,and added only an early
new-ledger-marker refusal to Assert-ProcessCompletion. In invocation13,the
direct typed fake envelope passed pure evaluation with Acceptance=false,and
the existing acceptance gate correctly rejected it. JSON conversion/readback
then failed the NestedRoundtrip assertion. No malformed-input test below that
assertion,Task3 GREEN5,Task4 wiring,final allowlist or independent review ran.

The evaluator returns Errors,but this test's failing assertion did not print
that value. Saved output does not establish the specific serialized field/type
that was rejected. Do not guess a correction,relax type validation,or call
Task3 verified. No change/retry followed this unexpected failure.

Current use13/32;19 calls unused. Proposed next decision:one pure simulated
diagnostic call that emits the rejection reason (no native API/target),then a
cause-specific roundtrip correction and replacement GREEN7/GREEN5,with total
cap34. Arithmetic:13 used +1 diagnostic +2 Task3 GREEN +4 Task4 RED/GREEN +14
final allowlist=34. Preserve all refusal rules/old records;any additional
unexpected failure stops. This proposal has not been approved or executed.

Saved-only SHA256 at this stop:

| Artifact | SHA256 |
| --- | --- |
| invocations.json at13 calls |987401123D16C51ACDC48F81AF0808D7090AEBBB5E7AE9D1DA5E3C6B3CF6C2A1 |
| run-13.stdout.txt |DFCD4BC2B26D9A6BE399ACAD641E8DAAC5BA84F114B27006B64A8D20CBBC33A0 |
| run-13.stderr.txt |48E865F366FC512963EB4EC14033620793C4DA3934AD1D563ABEFD8E44D235B4 |

All13 suite hosts completed. The checkpoint includes unfinished Task3 code and
is not green or accepted. Existing native runner/SDK/Go/callers unchanged;
schema4 still refused and no real target ran. Git carries portable progress
through Tasks1/2 and this Task3 stop,not raw logs/SDK/cache or acceptance.

## Approved UTC diagnosis and final fake/static regression

After468709b6,user approved one pure diagnostic,cause-specific repair and cap34.
Invocation14 DIAGNOSTIC7 exit0/1589ms emitted rejection String type StartUTC.
The JSON host can materialize ISO timestamps as UTC DateTime;validator expected
string metadata. Corrected only UTC metadata validation to recognize UTC
DateTime/DateTimeOffset or a valid UTC string. Other field types remain strict;
invalid UTC is rejected. Added diagnostic output mode with Acceptance=false,
not a workload or acceptance shortcut. No historical output was rewritten.

Task3 replacement GREEN7 run15 exit0/1791ms and GREEN5 run16 exit0/1981ms passed
all88 receipt assertions each. Task4 expected RED17/18 failed ExplicitEntryOnly;
GREEN19/20 passed9 each (1464/1353ms). Native adapter/entry exists but was never
invoked. Static/reflection hooks do not prove Windows adapter behavior.

Final regression,each in a fresh64-bit host,one invocation per suite/host:

| Suite | PS7 call/PASS/ms | PS5.1 call/PASS/ms |
| --- | --- | --- |
| ordinary-ledger.tests.ps1 |21/9/1448 |22/9/1144 |
| ordinary-ledger-session.tests.ps1 |23/14/1290 |24/14/1402 |
| ordinary-ledger-receipt.tests.ps1 |25/88/2508 |26/88/1869 |
| ordinary-ledger-integration.tests.ps1 |27/9/1200 |28/9/1247 |
| retained-lifetime.tests.ps1 |29/27/1094 |30/27/736 |
| retained-lifetime-integration.tests.ps1 |31/12/1065 |32/12/981 |
| retained-lifetime-receipt.tests.ps1 |33/16/1342 |34/16/1554 |

All14 final invocations exit0,empty stderr,350 PASS assertions total. All34
budgeted invocations consumed,including historical expected/unexpected failures
and the one approved diagnostic. No Go build/native target/old controller run,
SDK modification or metadata continuation. This is the exact allowlist only,
not a full project or native integration acceptance claim.

Task4 adds explicit RunWithOrdinaryIdentityLedger,conditional new-mode fields,
ordinary adapter/shared clocks/finalization and guaranteed owning-ledger teardown.
Existing Run/diagnostic modes retain their entry selection. New schema4 and
new-marker downgrade attempts remain refused by Assert-ProcessCompletion.
Preflight checks new portable files;legacy integration compile list expanded
only to resolve new source dependencies. No caller/matrix driver adopts the entry.

Independent whole-change readonly review is next and still pending. This is
not the final implementation completion declaration. No additional execution
is authorized;review findings needing repairs/tests require a bounded decision.
Original metadata RED and consumed runtime identities remain unaccepted/frozen.

Final independent review returned one P1/five P2 findings;all accepted after
author source checks. Task4 and the overall slice remain incomplete. No Critical
or Minor reported. No post-review corrective code/test was authorized/performed.
The review report contains exact paths/effects,all declined-area rulings and a
proposed26-invocation fake-only repair pass awaiting user approval/specific scope.
Current34-row invocation index SHA256:
38F718E5D7BABA66B8D55AB6630EA2F5543C91A7CA951390E4E8B769092D0340.
No review re-run or extra workload. Preserve this plan's ignored evidence directory
because final review is not clean;do not apply the clean-finish deletion step.
