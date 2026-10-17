# Ordinary ledger native probe — Stage A execution record

Latest:[covering review-fix pass complete in fake scope](2026-10-17-ordinary-ledger-native-probe-repair.md).
47 checks per host/final350 passed;Stage A36/36 used. Prior blocked/pending states
below are historical. No Stage B/native/acceptance authority or second review.

Final disposition:[independent review found binding/claim gaps](2026-10-17-ordinary-ledger-native-probe-review.md).
Stage A NOT complete;Stage B not ready/unapproved.Three Important findings
accepted,one Minor deferred,no repairs/tests after review.18/18 consumed.
Earlier review-pending states below describe the pre-review checkpoint.

Date:2026-10-17. Base6a9ade2e5fbb961e826233430df9485107f40f3c.
Implementation checkpoint636e489092051d9f8523215da55cce978adbc79e.
User approved Stage A design/18 fake invocations,Native/inline. Stage B has NOT
been approved. Current status:scoped tests green,independent readonly review
pending;not yet Stage A completion or permission for real probes.

## Implemented surface

New ordinary-ledger-native-probe.ps1 imports existing run.ps1 without launch.
Planner fixes Root/Child identities,system Windows PowerShell executable and
encoded fixture bodies,nonce,empty environment overrides,5s/20s root limits,
eleven exact loaded-source/hash bindings and Acceptance=false/NoRetry=true.
Plan reconstruction refuses mutation,current source/executable drift and existing
run/control roots. Separate control directory consumes the identity before any
boundary call;child case requires a verified completed root case first.

The sole native boundary selects RunWithOrdinaryIdentityLedger,never diagnostic
mode. Default boundary requires64-bit Desktop5 control host. Runner additionally
requires explicit RuntimeApproved and exact plan SHA256. These caller flags/pins
are not security clearance or approval generation:the separate Stage B current
safety/prerequisite review remains mandatory and unperformed. No input plan or
runtime approval pin for actual Stage B has been generated/approved this turn.

Evidence evaluation checks the unchanged strict schema4 evaluator,root exit and
invocation binding,nonce/PID/time/order markers,controlled-child admission running
and final same-handle exit. Markers corroborate only;they cannot repair ownership
or join historical identities. Result remains Acceptance=false/OriginalTestPassed=false.
Provisional result is not final:manifest hashes bind artifacts and completed.json
binds the manifest/plan;the saved validator recomputes raw evidence and hashes.
Stopped/missing/corrupted/finalization-failed records are refused. A failure does
not erase consumed identity;raw stopped evidence remains on disk.

## Verification scope and outcomes

New suite ordinary-ledger-native-probe.tests.ps1 uses real planning,parameter
validation,evaluator,filesystem JSON/hash writes,manifest and saved-result paths.
It replaces only the workload boundary and native independent-file/link check.
No native entry/adapter,file-link P/Invoke,actual child,Go build/runtime,Docker,
network or security-setting change was invoked. Existing C# declarations compile
as part of fixture loading,which is not a native workload or Go build.

| Call | Host | Phase | Exit | ms | Outcome |
| --- | --- | --- | --- | --- | --- |
| 1 | PS7 | RED | 1 | 6987 | NativeProbePlannerExists expected missing-planner assertion |
| 2 | PS5.1 | RED | 1 | 5192 | Same expected assertion |
| 3 | PS7 | GREEN | 0 | 11660 | 39 PASS,empty stderr |
| 4 | PS5.1 | GREEN | 0 | 6524 | 39 PASS,empty stderr |

Named coverage:fixed shell/args/limits/source count;seven plan mutations;
wrong identity;missing runtime approval/wrong pin;child-before-root;root and
child fake completion;saved revalidation;schema4 refusal;consumed/replay refusal;
saved source drift/byte mutation;boundary throw/missing result/wrong nonce/cleanup
failure;missing saved receipt;completion write failure versus provisional result.
The source/file-link boundary implementation itself and real host/OS behavior
are not proven by the mocks. No full controller/project suite was authorized.

### Final seven-suite allowlist,14 invocations

| Suite | PS7 call/ms | PS5.1 call/ms | PASS per host |
| --- | --- | --- | --- |
| ordinary-ledger.tests.ps1 | 5/1352 | 12/1200 | 9 |
| ordinary-ledger-session.tests.ps1 | 6/1136 | 13/1200 | 14 |
| ordinary-ledger-receipt.tests.ps1 | 7/1687 | 14/1884 | 88 |
| ordinary-ledger-integration.tests.ps1 | 8/1241 | 15/1439 | 9 |
| retained-lifetime.tests.ps1 | 9/920 | 16/812 | 27 |
| retained-lifetime-integration.tests.ps1 | 10/1120 | 17/1158 | 12 |
| retained-lifetime-receipt.tests.ps1 | 11/1406 | 18/1732 | 16 |

Every final call exited0 with empty stderr:350 assertions. New suite GREENs
add78,not silently included in350.18/18 Stage A calls consumed;not an extension
of original34 or prior47 review-repair calls. No more suite/native calls granted.
Fresh64-bit hosts,-NoLogo -NoProfile -NonInteractive -File,120s outer deadline;
PS5 child-only inbox PSModulePath retained. All18 test hosts completed.

## Preservation and remaining gates

Tracked helper/tests/this report support recovery without ignored artifacts.
This plan's raw records remain under
.superpowers/sdd/2026-10-17-ordinary-ledger-native-probe-preparation/.
Index SHA256:24BF6872D5CCF071BB92B0E24CDA4761B847DCAF5B34E788797939DA8D91B52C.
Prior consumed indices/logs were not edited. Test-only fake evidence roots are
ignored and retained;none is an actual Stage B ledger-native-root/child-001 run.

Rulings:

1. Use Windows PowerShell bookkeeping and fake workload/file-link seams instead
   of Bash/native execution;cost:OS/file-link/native boundary remains unverified.
2. Explicit approval+pin and completed root before child are caller safeguards,
   not a security attestation;cost:Stage B needs separately reviewed current
   prerequisites and exact-input approval,not just setting those parameters.
3. Preserve ignored raw/test records and existing worktree under standing user
   recovery instruction;do not infer cleanup,merge,PR or external publication.

Independent readonly final review requested under the execution/review skills
for6a9ade2e..636e4890;no edits,tests,native calls or further delegation allowed.
Its findings will be appended before handoff;budget exhaustion means any needed
fix/reverification requires a new bounded decision. Stage B/native,full ordinary/
C11/C12/security,metadata,main and production/acceptance gates remain unchanged.
