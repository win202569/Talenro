# Fixed inbox executable policy — implemented, fake-verified only

The user approved the short design and a separate18-call fake-only allowance
after development checkpoint `5ea1cef71f8b1194e25ab24db8f84db3a6d9c792`.
This change addresses the single-link incompatibility described in
[prerequisite evidence](2026-10-17-ordinary-ledger-native-prerequisites.md).
It does not approve Stage B or certify the real OS integrations.

## Implementation and policy boundary

- Source/evidence files still require exactly one hardlink. The generic helper
  delegates only the existing OS link-count operation to a small boundary.
- Only the fixed `SystemRoot/System32/WindowsPowerShell/v1.0/powershell.exe`
  uses `fixed-inbox-shell-binding-v1`. It is not an arbitrary-path exception.
- An open FileStream with FileShare.Read is held while collecting SHA256,
  length, volume/file-index identity, link count and Authenticode evidence.
  Signature status must be Valid and subject exactly Microsoft Windows /
  Microsoft Corporation / Redmond / Washington / US; signer certificate raw
  bytes are SHA256-pinned. Numeric identity members must be nonnegative uint32
  values; volume, file index and link count must be present/nonzero.
- Multiple links are accepted only under this executable-specific policy;
  their actual count is pinned, not assumed to be two. Missing/drifting identity,
  signature, hash, signer, length, count or fixed path refuses.
- `IsolatedFuzzFileInfo.Snapshot` adds an additive held-handle metadata query;
  existing Links behavior and other callers are unchanged.
- Plan schema2 / `ordinary-ledger-native-probe-v2` includes ShellBinding.
  Old schema1 plans are deliberately rejected, not silently migrated or replayed.
- The initial source/executable bindings remain held through finalization and
  saved-result validation. The fixed executable path is reopened and compared
  with the approved identity immediately before the sole native entry; this
  additional binding is held across that entry too. Both release in finally.

No system file/link was changed, copied or substituted. A read sharing contract
is not protection against hostile administrators/kernel actors, preexisting
mapped writes, or every directory-namespace race. Authenticode/identity evidence
and a caller hash are not external security clearance. Existing source-load,
process-safety, cleanup, timeout, no-retry and Acceptance=false rules remain.

## Tests, budget and evidence

Production changes were made only after both RED calls failed specifically at
`DedicatedInboxShellValidatorRequired`. Later assertions were not individually
observed failing in RED. Static parsing succeeded for all three edited scripts.
New suite uses the real planner, binding/validation, runner and saved evaluator;
signature, held-handle identity, general link-count and workload OS boundaries
are fake. The fake native boundary is installed before import, unarmed by
default and armed once for each intended fake call. No native entry/real child.

| Call | Host/phase | Exit | ms | Result |
| --- | --- | --- | --- | --- |
| 1 | PS7 RED | 1 | 3487 | Expected missing-validator assertion |
| 2 | PS5.1 RED | 1 | 4157 | Same expected assertion |
| 3 | PS7 GREEN | 0 | 84193 | 64 PASS,empty stderr;see bookkeeping note |
| 4 | PS5.1 GREEN | 0 | 45818 | 64 PASS,empty stderr |

New cases cover multi-link executable binding, invalid/other-subject signature
refusal, missing identity, wrong-path refusal, source multi-link refusal, plan
identity/link/signer mutation and pre-consumption identity/link/signer drift.
The existing real managed lock check operates on owned plaintext fixture only;
it never attempts to modify a system executable. Real Authenticode/PInvoke
adapter integration is unverified; compiling the adapter is not runtime proof.

| Final suite | PS7 call/ms | PS5.1 call/ms | PASS per host |
| --- | --- | --- | --- |
| ordinary-ledger.tests.ps1 | 5/1570 | 12/1238 | 9 |
| ordinary-ledger-session.tests.ps1 | 6/1140 | 13/1278 | 14 |
| ordinary-ledger-receipt.tests.ps1 | 7/2294 | 14/2376 | 88 |
| ordinary-ledger-integration.tests.ps1 | 8/1989 | 15/1518 | 9 |
| retained-lifetime.tests.ps1 | 9/1062 | 16/833 | 27 |
| retained-lifetime-integration.tests.ps1 | 10/1513 | 17/1226 | 12 |
| retained-lifetime-receipt.tests.ps1 | 11/1592 | 18/1795 | 16 |

Final14 all exited0 with empty stderr,350 assertions;new suite128 separate.
Fresh64-bit PS7/5.1 test hosts used NoLogo/NoProfile/NonInteractive/File,
120s outer deadline and PS5 child-only inbox module path. No retries or deadline
extension. New allowance18/18 consumed;Stage A cumulative54/54 including earlier
36/36. Full project/controller suites were not run and are not claimed passing.

### Bookkeeping anomaly — preserve, do not repair original raw index

Calls3/4 overlapped under a non-concurrent index writer. Call4 retained call3's
pending state and overwrote its completion fields. The final raw index therefore
has18 numbered entries but only17 completed fields;call3 remains null/false.
Do not claim the index alone proves all18 completed, or rewrite it as clean.
Call3's captured runner output independently reported exit0/ms84193;its separate
saved stdout contains64 PASS and its stderr is empty. This is the basis of the
table's call3 disposition, not inference from a pending index. Calls5–18 ran
sequentially after both GREEN hosts ended. No invocation was repeated or omitted
from budget. The raw records/logs remain ignored locally;this portable report
preserves their limits and observed reconciliation for a new computer.

Raw directory: `.superpowers/sdd/2026-10-17-ordinary-ledger-native-probe-preparation`.
SHA256 values from saved-file reads:

- Original invocations.json unchanged: `24BF6872D5CCF071BB92B0E24CDA4761B847DCAF5B34E788797939DA8D91B52C`
- Earlier repair-invocations.json unchanged: `D056F2217E8920BA9578433A20062849261CAF4A11E3EC780B34B0C1A037C40D`
- policy-invocations.json with documented pending entry: `64ED3BB625C7332274ED268F9F2985F07529931998EE40F10F68F816A3A25486`
- policy-3.stdout.txt: `B68A5B65BF6C0B9E960A8D2BFCCFACA71F21D7ED65AC8D923EC76DA50EEC0ED9`
- policy-3.stderr.txt (empty): `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855`

## Independent read-only review and disposition

One reviewer inspected these three uncommitted source/test diffs against the
base checkpoint,without edits/tests/helper launches or another reviewer. No
Critical/Important source defect found in this bounded fake preparation.
Minor accepted/deferred:tests do not independently catch removal of the
immediate-entry recheck or premature binding disposal during finalization.
Author verified both operations exist in current source,but existing early-drift
and native-boundary lock assertions do not prove those timing guarantees.
No additional test allowance is inferred;no second review or coverage fix run.

Reviewer declined runtime OS integration,Stage B/child containment,safety
clearance/hostile-admin resistance,and recertification of inherited loading
limitations. Author retains each as unverified/outside this change,not dropped.
Test totals above come from captured execution and saved-log inspection,not
the reviewer. No product/source edits after GREENs,only documentation.

## Resume safely

All four actual ledger-native-root/child-001 run/control paths remain absent.
No real native probe,Go/Docker workload,install,security-setting change,metadata
continuation,main merge or acceptance activation. Recover from the development
branch;do not reuse hashes/identity/signatures as another host's clearance.

Next recommended step:separately authorize the two missing timing regressions
and real read-only adapter/prerequisite evidence. Budget is exhausted;Stage B
remains unapproved and historical Defender/I2/I3 gates remain unresolved. This
fake-green source checkpoint does not itself clear the real hardlink prerequisite.
