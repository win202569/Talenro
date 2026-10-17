# Ordinary ledger review repair — stopped, not accepted

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
