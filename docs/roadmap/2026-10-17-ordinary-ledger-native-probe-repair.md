# Native probe Stage A — approved covering repair, fake verification green

Date:2026-10-17. Base2feca69b53cbf82ed65a0fb23d69aedf01038c8d.
User approved binding/test-protection fixes,narrower child claim and18 extra
fake calls. This is the single covering review-fix pass,not Stage B approval.
No second independent reviewer or real probe was invoked.

## Corrections and evidence boundary

1. **Verified core load snapshot.** Before importing run.ps1,the initializer
   rejects an already loaded IsolatedFuzzProcess type and opens all11 bound
   source files readonly with FileShare.Read. It checks its own parsed script
   text against the locked file,loads dependencies while locked,and retains
   source hashes plus the freshly loaded native type identity. Required pure
   dependency functions are explicitly exported from the initializer scope.
   New plans and saved validation reject later source drift against that load
   snapshot,not merely against whichever bytes happen to be on disk now.
2. **Executable/source launch window.** Runner acquires verified source and
   executable read bindings before plan validation and keeps them through root
   prerequisite checks,identity consumption,the boundary and finalization.
   FileShare.Read excludes write/delete sharing. All handles close in finally,
   including partial acquisition failures. The test uses an owned plaintext
   fake executable:write access succeeds before/after the run but is refused
   inside the fake boundary. No real executable is modified or launched.
3. **Narrowed claim.** User approved same-run owned-handle natural completion,
   not proof the child outlived root. Results now explicitly record
   ObservationClaim='same-run-owned-handle-natural-completion' and
   ChildOutlivedRootVerified=false. Intent markers remain corroboration only;
   existing ownership/final-exit and conservative marker checks are not relaxed.
4. **Fail-closed fake.** Test installs its unarmed boundary before helper import;
   helper preserves an existing boundary. Each intended fake call is armed once
   and immediately disarms on entry. Unarmed/extra calls throw. File-link and
   owned-scope test seams are restored after import;they cannot intercept actual
   workload creation because the workload seam is already installed.

The probe test now uses a tracked hand-derived PowerShell receipt fixture rather
than precompiling the native/fake C# batch before import. The real helper therefore
performs its verified fresh core-type load during the test. Ledger/session
behavior remains exercised by the existing final allowlist,not inferred from
hand-derived receipt facts. Native process APIs and the independent-file/link
P/Invoke boundary are still not exercised;current signatures,security readiness,
host freshness and real runtime behavior remain Stage B prerequisites.

## Verification outcomes

Same combined suite ordinary-ledger-native-probe.tests.ps1,once per phase/host.
Expected RED is specifically ImportMustPreserveFailClosedBoundary;the suite stops
there,so do not claim each later assertion independently failed in RED.

| Repair call | Host | Phase | Exit | ms | Outcome |
| --- | --- | --- | --- | --- | --- |
| 1 | PS7 | RED | 1 | 1801 | Planned import-boundary assertion |
| 2 | PS5.1 | RED | 1 | 1467 | Same planned assertion |
| 3 | PS7 | GREEN | 0 | 62437 | 47 PASS,empty stderr |
| 4 | PS5.1 | GREEN | 0 | 35948 | 47 PASS,empty stderr |

GREENs cover previous39 behaviors plus import preservation,fresh-type guard,
preloaded-type refusal,unarmed boundary refusal,loaded-source drift before plan,
explicit no-exit-ordering claim,held executable binding and release afterward.
These are94 checks across both hosts,not part of final350. The measured suite
times remain below120s;no timeout extension or performance guarantee is inferred.

### Final allowlist

| Suite | PS7 repair call/ms | PS5.1 repair call/ms | PASS per host |
| --- | --- | --- | --- |
| ordinary-ledger.tests.ps1 | 5/1314 | 12/1307 | 9 |
| ordinary-ledger-session.tests.ps1 | 6/1081 | 13/1226 | 14 |
| ordinary-ledger-receipt.tests.ps1 | 7/1615 | 14/2134 | 88 |
| ordinary-ledger-integration.tests.ps1 | 8/1432 | 15/1588 | 9 |
| retained-lifetime.tests.ps1 | 9/971 | 16/957 | 27 |
| retained-lifetime-integration.tests.ps1 | 10/1225 | 17/1266 | 12 |
| retained-lifetime-receipt.tests.ps1 | 11/1479 | 18/1863 | 16 |

All14 calls exited0 with empty stderr,350 assertions. All18 repair hosts completed.
Fresh64-bit PS7/5.1,-NoLogo -NoProfile -NonInteractive -File,120s outer deadline,
PS5 child-only inbox module path unchanged. Repair18/18 consumed;Stage A18+18=36/36.
Older47 review-repair and original34 budgets are separate and remain frozen.

## Preservation, rulings and remaining gates

Separate repair-invocations.json keeps the first18 Stage A records unchanged.
Ignored raw index/logs remain under the plan's native-probe-preparation directory.
SHA256 from saved-file reads:

- Original Stage A18 index:24BF6872D5CCF071BB92B0E24CDA4761B847DCAF5B34E788797939DA8D91B52C
- Repair18 index:D056F2217E8920BA9578433A20062849261CAF4A11E3EC780B34B0C1A037C40D

Rulings:use pure PS fixture to permit verified fresh core loading (cost:receipt
fixtures do not prove ledger scheduling);use a separate18-call index (cost:combine
18+18 totals);retain owned worktree/raw fixtures per standing recovery instruction
(cost:local ignored disk usage). Prior review's declined-to-judge limitations
remain in its report;the one deferred Minor was explicitly included by the user
and is no longer deferred. No new minor fixes or scope expansion were inferred.

The reviewed binding gaps have covering fake regressions;the child claim is
resolved by the user's explicit narrowing. This is author disposition after one
fix pass,not a second independent review or real-environment certification.
No source changes after the GREENs/final regression,only documentation.

Stage A preparation is complete within this bounded fake scope. Stage B remains
unapproved and must separately review current safety/prerequisites,exact plan/
source/executable hashes,outer watchdog and finite native allowance. No actual
ledger-native-root/child-001 identity consumed,Go/Docker workload,metadata replay,
full project/controller suite,main merge,security-setting change or acceptance
activation occurred. Next safe step is readonly Stage B prerequisite preparation,
not executing the helper merely because its caller switches exist.
