# Ordinary ledger review repair — scoped regressions green, rollout blocked

Latest:[initial snapshot follow-up verified](2026-10-17-ordinary-ledger-initial-snapshot-fix.md).
User approved18 extra calls after776f2d3e. Both hosts RED→GREEN21 checks each;
final14 runs/350 assertions green. The closeout provenance finding below is
resolved in fake scope;47/47 consumed. Native/workload acceptance still blocked.
Historical stop/proposed/unapproved statements below do not describe current state.

## Author closeout finding:failed initial snapshot can reuse runtime members

After final tests,readonly source inspection found an additional provenance gap
introduced by repair B. OrdinaryMemberLedger.LastMembers changes only after a
valid list. If runtime polling previously succeeded with members,then the first
shutdown list fails/truncates/overruns,session Finish copies LastMembers into
InitialMembers anyway. Thus InitialMemberQuery describes the failed new call but
InitialMembers can describe an older runtime call. Current B failure fixtures
start without prior runtime polling,so all29 tests miss this path. This is a
source-established stale-snapshot risk,not a newly executed native observation.

Sticky Incomplete remains true and natural/evidence success remains false;this
does not establish an acceptance bypass. It nevertheless violates accurate raw
initial fact reporting. Implementation completion remains BLOCKED. Do not claim
all findings closed or merge/enable workflows merely because350 final checks pass.
No code correction,test fixture addition or rerun after identifying this gap.

Proposed new bounded decision:four fake calls (new stale-initial-snapshot RED
and GREEN on both hosts),then14 original final regressions;18 extra,total47.
Regression must seed a valid runtime list first,then initial shutdown failure
234/truncation/overrun;initial list facts must never be represented by stale
runtime members. Preserve failure records,sticky refusal,deadline and all rights;
no native workload. This proposal is not approved. Existing29-call index frozen.

## Final corrective-pass result

After5b4dc515,user approved one readonly PS5 diagnostic,cause-supported child-host
configuration correction and cap29. Diagnostic6 (exit0/1275ms) showed Desktop5.1
had loaded the bundled Core-only Microsoft.PowerShell.Utility manifest,which
did not export Get-FileHash. Inbox Desktop manifest did export it. Correction:
only the PS5 child ProcessStartInfo.EnvironmentVariables['PSModulePath'] is set
to SystemRoot/System32/WindowsPowerShell/v1.0/Modules before launch. PS7,parent
environment,installed modules,production helpers and security settings unchanged.
Replacement GREEN5 call7 passed28 assertions,exit0/1930ms,empty stderr.

The approved corrective-pass execution is exhausted with its scoped tests green;
the additional closeout finding above prevents implementation completion.
No second independent review was dispatched;the six existing findings have
covering grouped regressions and a green final allowlist. This is not full
project testing,OS/native adapter validation,Error87 repair or workload acceptance.
Metadata remains blocked;no existing workflow replay,main merge,PR or rollout.

| Finding | Correction | Evidence |
| --- | --- | --- |
| 1 legacy compatibility | Common matrix/leaf field selection;legacy shape excludes additive fields | A RED1/2;GREEN4/7,28 per host |
| 6 source bindings | Three loaded sources required/hash-bound by five producers/consumers | A GREEN4/7;named/count fixtures mechanically +3 |
| 2 initial queries | Ordinary session captures initial accounting/list within shared deadline;legacy retry helper only in else branch | B RED8/9;GREEN10/11,9 per host |
| 3 stale final timing | Accounting/list start no earlier than shutdown,including zero entries | B GREEN10/11 |
| 4 policy downgrade | New policy refused before legacy schema dispatch | C RED12/13;GREEN14/15,10 per host |
| 5 diagnostic rows | Old observation containers must be empty if present in schema4 | C GREEN14/15 |

B expected RED:StaleRootOnly-FinalAccounting (1645/1638ms). GREEN10/11:
1614/1581ms. C expected RED:PolicyOnlyDowngrade-0 (1405/1440ms).
GREEN14/15:1568/1835ms. All successful grouped runs had empty stderr.
Fixture initial-accounting errors/truncated lists/234/overrun are sticky and
single-attempt;fake initial facts retained. Native branch examined statically,
never invoked. Existing source-count/controller tests were updated mechanically
but deliberately NOT run outside the approved allowlist. Their full controller
behavior remains unverified. Optional initial facts are unattempted on early
fatal/root-unfinished paths;no affirmative evidence is manufactured from them.

### Final allowlist:14 invocations,350 assertions

| Suite | PS7 call/ms | PS5 call/ms | PASS per host |
| --- | --- | --- | --- |
| ordinary-ledger.tests.ps1 | 16/1331 | 23/1231 | 9 |
| ordinary-ledger-session.tests.ps1 | 17/1050 | 24/1128 | 14 |
| ordinary-ledger-receipt.tests.ps1 | 18/1848 | 25/1873 | 88 |
| ordinary-ledger-integration.tests.ps1 | 19/1301 | 26/1443 | 9 |
| retained-lifetime.tests.ps1 | 20/922 | 27/749 | 27 |
| retained-lifetime-integration.tests.ps1 | 21/1103 | 28/1239 | 12 |
| retained-lifetime-receipt.tests.ps1 | 22/1442 | 29/1653 | 16 |

Every final call exited0 with empty stderr. A/B/C add94 assertions across six
successful runs (28+9+10 per host);not silently counted in the original350.
29/29 repair calls consumed;no additional host/suite executions authorized.
Original34-call index SHA remains38F718E5D7BABA66B8D55AB6630EA2F5543C91A7CA951390E4E8B769092D0340.
Final repair index SHA256:ED2DF76BEE9886E1A58BBD86230DBA4880EDD69E22C0043042356829C0DF55BA.
Raw failed and successful logs are retained unchanged and ignored. Two attempted
saved-file summary commands failed (PowerShell pipeline syntax/null byte count);
the corrected readonly summary verified all29 records without running a host or
changing evidence. No test outcome was reclassified because of these errors.

### Portable recovery and decisions

Recover tracked tests/sources from the development branch. The ignored runner
is not required:with future explicit test authorization,use a fresh64-bit
PS7 or PS5 -NoLogo -NoProfile -NonInteractive -File host per named suite,120s
outer deadline,record before launch,stop on unexpected failure. For PS5 only,
set that child's PSModulePath to the inbox directory above,not the global one.
Do not automatically rerun the diagnostic or original REDs. No caches or raw
receipts are required by these fake fixtures;diagnostic script is tracked.

Rulings made in this corrective pass:

1. Refine approved repair fixtures instead of another design cycle;cost if
   wrong:compatibility/binding surface expands. Exact surface is recorded.
2. AST-extract actual producer copy logic without running wrappers;cost:full
   driver integration remains unverified.
3. Pin inbox modules only in PS5 test child;cost:third-party modules unavailable
   there,none needed by allowed suites. No production/module fallback.
4. Existing release-order assertion now selects last (final) accounting because
   initial accounting is newly explicit;cost:wrong selection could hide ordering,
   retained assertion still requires final accounting after release.
5. Keep branch/worktree/raw records for existing progress-preservation authority;
   no integration decision inferred. No deferred minors from original review.

Next work is a separately scoped decision about environment/native evidence,
not enabling acceptance or resuming consumed metadata identities. All earlier
stop/proposal statements below are historical.

## Latest: approved fixture correction passed PS7, PS5 host lookup stopped5/27

After53bca818,user approved only the two original array initializations and cap27.
MemberEvidence/MembersDuringRun now initialize as empty arrays;production checks
unchanged. Repair4 GREEN7 exited0 in1668ms,28 PASS,empty stderr. Repair5 GREEN5
exited1 in1634ms after5 PASS: Get-FileHash was not recognized at fixture line34.
The original shape failure is cleared on both hosts, but source-binding checks
did not run on PS5. Group A is not complete;count-fixture updates still pending.
B/C and final regression remain unstarted;all six findings remain open.

Stopped immediately;no module import, host diagnostic, fix or rerun followed.
Readonly inspection finds the Windows Utility manifest exports Get-FileHash and
its psm1 defines it;the parent host resolves the command. Parent PSModulePath
includes both Core and Desktop module directories. Why the child cannot resolve
it is NOT established;module-path/loaded-module collision is only a hypothesis.
No installed module, environment security setting or production helper changed.

Saved-file SHA256:

- repair4 stdout:03F70141055A3B9598BA0D1AB4606EA5A93B571A5767404E03205906E86D5BAB
- repair5 stdout:25BE838E2929A6DDCEFBEA1C25C80136F06912F7DAF783677DFE91562EC33C06
- repair5 stderr:8B2D91449613BAA1CE7EF51131DB71F79F49B8A7C3D09CF256A04A11FD744593
- repair index after5:E9D7A9A9ACBA1A5D04E8543AABE6A8989A50562DF0D5F2CDB0A51C1CF2DD018A

Proposed next bounded decision: one readonly fresh PS5 host diagnostic (module
paths/loaded Utility/command discovery only, no native workload), then a
cause-specific test-host setup correction only if supported by that evidence,
replacement A GREEN5 and original remaining plan. Cap29:5 used +1 diagnostic
+1 replacement GREEN5 +8 B/C +14 final. If diagnosis does not support a scoped
host-only correction, stop and report;no security changes/install or production
hash fallback. This proposal is not approved. Earlier stop states below historical.

Date:2026-10-17. Base9c89149efb12e8c95488501640c74c817008ee8f.
User approved six-finding corrective pass A/B/C, Native/inline execution and
26 additional fake/static invocations. Original34 records remain frozen.

## Execution and stop

| Repair call | Host | Phase | Result |
| --- | --- | --- | --- |
| 1 | PowerShell7 | A RED | Expected failure: matrix-check.ps1 LegacyReceipt;exit1,1624ms |
| 2 | PowerShell5.1 | A RED | Same expected failure;exit1,1482ms |
| 3 | PowerShell7 | A GREEN | Unexpected failure after two PASS lines;exit1,1606ms |

Call3 passed the actual extracted matrix/leaf property-copy checks, then failed
the existing shape validator: `Receipt collection must retain its array/list shape`.
Readonly source comparison identifies the newly constructed fake Result's
MemberEvidence and MembersDuringRun as uninitialized null arrays. The fixture
sets other array fields; normal runner initialization supplies these original
arrays. No fixture correction, further suite invocation or production shape
validator relaxation occurred after the stop. Used3/26;all three hosts completed.

## Partial changes, not verified fixes

Group A adds shared legacy receipt field selection, schema-specific legacy shape
handling and three loaded dependencies to five source-binding producers/consumers.
Only two producer assertions passed before the stop. Shape compatibility,
new-mode refusal, source bindings and their regressions remain unverified.
Existing source-count test expectations still require the planned mechanical
updates and have not run. Findings1/6 are NOT closed;B/C findings2–5 are untouched.
No native adapter/workload, Go build, metadata continuation, rollout, main merge
or second independent review occurred. Do not run ordinary matrix/leaf workflows
until repair and scoped verification finish. The original350 assertions do not
validate these new edits.

## Preserved evidence and recovery

Raw logs stay ignored under
`.superpowers/sdd/2026-10-17-ordinary-member-identity-ledger/`.
SHA256, verified by saved-file reads only:

- Original invocations.json:38F718E5D7BABA66B8D55AB6630EA2F5543C91A7CA951390E4E8B769092D0340
- repair-invocations.json:5B88FB6EF4490AC484C64FB66A33CFF25339090B89C059448ED9BBB884795FC8
- repair-3.stdout.txt:C0B56BA961536D0F3ACB061C75AA32E45B4CECC5004CC8C3E9EAAC355E0EBE6B
- repair-3.stderr.txt:A8E29C142FBF9108D6DD214539A50DE23BADBD49B2A4E610D217ABB954D43AA8

Tracked sources, fixture and this summary support recovery without raw logs.
Resume only after a new bounded decision: proposed fixture-only initialization
of the two original arrays, no change to production shape checks, and cap26→27.
This permits replacement GREEN7/GREEN5, then the original B/C and final allowlist:
3 consumed +2 A GREEN +8 B/C RED/GREEN +14 final =27. No rerun of RED1/2.
This proposal is not yet approved at this checkpoint. Retain all original gates
and stop again on any unexpected failure.
