# Probe timing regressions — fake coverage completed

User approved the short test-only design and new16-host allowance after
`3bba6b7c0728c02968555e2ebd9586e0d91fb386`. This addresses the two prior Minor
coverage gaps in [inbox policy review](2026-10-17-ordinary-ledger-inbox-policy.md),
not a new production behavior fix. No product code changed.

## Test design and observed behavior

Existing probe tests add an optional SetupOnly return after the fake-boundary
bootstrap. New `ordinary-ledger-native-probe-timing.tests.ps1` reuses that
bootstrap,including verified source loading and preinstalled unarmed fake native
boundary. Signature/identity/link OS seams remain fake. Only owned plaintext
fixtures are opened for write-access checks;never the real system executable.

1. Late identity drift is injected after plan.json is written:initial binding,
   full plan verification and consumption have already happened. The real entry
   recheck refuses the new identity,zero fake native entries occur,and consumed/
   stopped artifacts remain without completed.json. This is not an early-drift
   test or a claim that an already consumed identity becomes reusable.
2. Real managed FileShare.Read write denial is observed at receipt,result,
   evidence-manifest and completion writes,and at entry to saved-result
   validation. Writable baseline and post-return access demonstrate that the
   fixture is not intrinsically read-only. The normal fake run finalizes once.

Before normal scenarios,the suite runs two test-only in-memory runner mutants:
one removes the entry recheck;one disposes held bindings before receipt writing.
Each must fail its exact behavioral assertion,not merely throw any error. The
original function is restored in finally. No on-disk product mutation or native
entry is used. Both mutants were caught on both hosts,then the unmodified
production runner passed both scenarios. These are caught mutation failures
within successful suites,not separate RED host calls or failing project tests.

## Verification and budget

All16 hosts ran sequentially,fresh64-bit PS7/5.1,NoLogo/NoProfile/NonInteractive/
File,120s deadline each. PS5 used only child-local inbox PSModulePath. Every
invocation was recorded before launching;no retries,timeout extension or overlap.
Both edited test scripts passed static parsing before host launch.

| Focused call | Host | Exit | ms | PASS |
| --- | --- | --- | --- | --- |
| 1 | PS7 | 0 | 17449 | 21 |
| 2 | PS5.1 | 0 | 10680 | 21 |

Each21 includes9 bootstrap checks,2 caught-mutant checks and10 normal timing
checks. These42 are separate from final350. Focused and final stderr all empty.

| Final suite | PS7 call/ms | PS5.1 call/ms | PASS per host |
| --- | --- | --- | --- |
| ordinary-ledger.tests.ps1 | 3/1312 | 10/1374 | 9 |
| ordinary-ledger-session.tests.ps1 | 4/1140 | 11/1265 | 14 |
| ordinary-ledger-receipt.tests.ps1 | 5/1625 | 12/2071 | 88 |
| ordinary-ledger-integration.tests.ps1 | 6/1226 | 13/1546 | 9 |
| retained-lifetime.tests.ps1 | 7/897 | 14/871 | 27 |
| retained-lifetime-integration.tests.ps1 | 8/1194 | 15/1646 | 12 |
| retained-lifetime-receipt.tests.ps1 | 9/1411 | 16/1838 | 16 |

All final14 exited0. New allowance16/16 consumed;Stage A cumulative70/70.
The full64-check probe suite and full project/controller suite were not rerun;
their prior results are historical,not new results of this allowance.

Raw index/logs remain ignored under
`.superpowers/sdd/2026-10-17-ordinary-ledger-native-probe-preparation`.
timing-invocations.json has16 numbered,completed entries with exit0;SHA256
`BBC919F4AA31ABAFADBE2AEE5F6BA0248FDDA7BD390CE12CCD99E388C1DE644B`.
Saved-file reads confirmed older hashes unchanged:

- policy-invocations.json: `64ED3BB625C7332274ED268F9F2985F07529931998EE40F10F68F816A3A25486`
- invocations.json: `24BF6872D5CCF071BB92B0E24CDA4761B847DCAF5B34E788797939DA8D91B52C`
- repair-invocations.json: `D056F2217E8920BA9578433A20062849261CAF4A11E3EC780B34B0C1A037C40D`

The older policy index's documented pending entry is not rewritten or cleared
by this clean sequential index. Portable evidence is this report;raw receipts
are not needed to recover source/tests,but cloning cannot restore local raw data.

## Read-only review and remaining scope

The same reviewer performed one read-only pass on this new test-only follow-up,
without edits,helper launches,tests or further delegation. No blocking/actionable
findings;both prior Minor gaps are meaningfully addressed in fake scope. Author
accepts this disposition against the actual tests and captured results.
No test/product edits after the two focused runs,only documentation.

Reviewer declined real OS identity/signature/process behavior and continuous
adversarial interleavings within writers/saved validation. Author retains both
as unverified:these tests observe stage boundaries,not every instruction or
hostile-admin behavior. Execution totals were checked by the coordinator,not
inferred from the review. This is not native readiness or security clearance.

All four actual ledger-native-root/child-001 run/control paths remain absent.
No real probe,Go/Docker workload,install,security-setting change,metadata replay,
main merge or acceptance activation. Next safe step is separately scoped
read-only OS adapter/prerequisite evidence,without invoking the workload entry.
Stage B remains unapproved and Defender/I2/I3 gates unresolved. Do not reuse
another machine's signature/hash/identity facts as current clearance.
