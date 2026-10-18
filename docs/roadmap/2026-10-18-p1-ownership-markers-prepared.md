# Ownership markers — PREPARED, NOT RUN; acceptance remains blocked

After3e58ecfc,user confirmed independent preparation and offline/static checks only.
Approved spec:[bounded design](2026-10-18-p1-ownership-group-marker-proposal.md).
Execution record:[portable ledger](2026-10-18-p1-ownership-preparation-ledger.md).

## Prepared changes

New folder:`docs/diagnostics/isolated-fuzz/integer-candidate-ownership-markers/`.
Ten files:eight source copies,105-record inventory,README. Seven sources byte-
identical to three-phase copies. Ownership source adds exactly four standalone
Console timestamp lines:first/last statements in SequentialIsolation and
PlainControls bodies. Removing them restores the entire original C4905925 source.
AST confirms body-edge placement;shared helpers,fixtures,API arguments/call sites,
metadata/result mutations,assertions,order,product/loading code unchanged.
END is inside each body;original P1I1-Case emits PASS only after body return.

Distinct external classifier/test/static-fidelity files do not modify old tools.
Inventory:105 ordered nonempty records,64 PASS,22 markers,11 pairs. New empty-
Parent OwnershipCase pairs occur after all prior timing scopes close.
New one-based positions:Sequential START96/END97/PASS98;Plain START99/END100/PASS101.
Removing the four new rules restores every prior101 inventory record exactly.

Classifier returns closed pairs and OpenScopes,plus OwnershipProgress states
unobserved/entered/ended-awaiting-pass/passed. Termination distinguishes receipt
timeout,completed and incomplete-exit;invalid input stays invalid with no inferred
ownership progression or fabricated end. In particular,body END is not PASS or
acceptance,and an error exit is not assumed timeout.
Host/context/typed receipt/exit/30000ms gates,Int64 ticks,monotonicity,exact record
order and parent/child nesting remain. FormalProjectAcceptance always false.
Classifier is supplied-text analysis,not proof of receipt authenticity,owned
processes,source provenance or real environment safety.

## Fresh evidence in this preparation session

All five scripts ran with the existing pinned PowerShell7.6.5 runtime using
-NoLogo -NoProfile -File;each final run exited0:

- candidate-ownership-capture-verifier.tests.ps1:122 synthetic offline cases,
  including all103 supported incomplete ordered prefixes (2..104 records),eight
  explicit ownership frontier snapshots and malformed/exit/receipt/tick cases.
- candidate-ownership-preparation.tests.ps1:20 static fidelity cases;eight source
  AST parses,seven unchanged hashes,exact whole ownership-source restoration,
  AST body edges,inventory restoration/positions and ten-file scope.
- candidate-boundary-capture-verifier.tests.ps1:47 unchanged offline cases.
- candidate-api-equality-capture-verifier.tests.ps1:63 unchanged offline cases.
- candidate-three-phase-capture-verifier.tests.ps1:76 unchanged offline cases.

Total328 passed checks/cases;prefix subchecks are not counted as extra cases.
Only classifier/test utilities ran. Target sources were read/AST-parsed only;
no target entry,module import,probe run,controller,warmup,retry,PS5 or Docker run.

RED evidence:before implementing the new classifier,the new synthetic test ran
against the unchanged101-line classifier. Exit1 ASSERTION FAILED: NormalPS7 was
expected rejection of105 records,not a missing-module or syntax failure.
GREEN initially110 cases,later122 with full prefix/receipt coverage.

Preparation helper initially failed AST/Parse/offline-supervision-policy.ps1:
its Case parameter Name shadowed the file-loop name under PowerShell dynamic
scope,creating a nonexistent Parse/file path. Direct source parse had zero
errors;minimal scope reproduction confirmed the wrong resolved filename.
Renaming only the helper label parameter fixed it;all20 fidelity cases passed.
No target source fix,assertion weakening or target allowance was consumed.

The existing388 frozen pins,manifest,controller and latest invocation index
were independently hash-checked unchanged. Latest index remains
02A8907BE3E1D12B3013EEF77ECABE4B6BAE314D0EBF034633D4EF07526A56D9.
Engine SHA256:
362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139.

## Portable identities (SHA256)

Paths below are relative to the repository;Git must carry all13 files plus the
scoped .gitattributes described below.

| File | SHA256 |
|---|---|
| `integer-candidate-ownership-markers/expected-output.json` | `C8C578530ECFFE2691FA2A62EB11C5C4B6DE4553984BDFE28F46E643EA1431AC` |
| `integer-candidate-ownership-markers/offline-supervision-policy.ownership.tests.ps1` | `2A1BF87DC436DA1AE7232B937EC0DA1DDD375953FDF7CD71EE127909693EA43F` |
| `integer-candidate-ownership-markers/offline-supervision-policy.ps1` | `3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3` |
| `integer-candidate-ownership-markers/offline-supervision-policy.reflection-reuse-regression.ps1` | `0B7692FE426407AA9431D54BDB1AA964216FB695CD5A202045ABDDCEB5C61069` |
| `integer-candidate-ownership-markers/offline-supervision-policy.reflection-reuse.tests.ps1` | `09BBAFD33ECED7DE1D98E00290A6B0D5E357E80F74A8766DBCC8409131C9B14F` |
| `integer-candidate-ownership-markers/offline-supervision-policy.shape-lookup-regression.ps1` | `12D6F5460F97B9E09453B0387C30E3D20CE19B2EDF1CDF4DBF4E77DA31561422` |
| `integer-candidate-ownership-markers/offline-supervision-policy.shape-lookup.tests.ps1` | `9DAA38EF1F19CA2320D17FCA35CDA87A1B4DFBDFCE960AF40EFB223B9E245B6B` |
| `integer-candidate-ownership-markers/offline-supervision-policy.tests.ps1` | `66EAC1E47D9540D92AE58D0AD94576F7EB1A53042A09857BFA3D81205714AEA2` |
| `integer-candidate-ownership-markers/phase-evaluator.ps1` | `6610F6F4396B9C0B1DF6AA202A63572F45EA037C175F30DF85973146F1881AE6` |
| `integer-candidate-ownership-markers/README.md` | `1CB342F58C93009892F06BF494E6B8EE4B7B9C3588298C4C57B43E66AC8B4903` |
| `candidate-ownership-capture-verifier.ps1` | `7570D7692A78F6F39893DCB6D4E2654DB6DAE31395BC153C6F6AA1FAE9DFEC51` |
| `candidate-ownership-capture-verifier.tests.ps1` | `B3A07FEE6EC20966BAAE5A5BC9D82C042F84C1C6019B9D099F9420FAF5E6474D` |
| `candidate-ownership-preparation.tests.ps1` | `5CC19431F76733A449FBEDD0505EB272F3E39EEDA51D1369F9EF8C960B7F4F2C` |

## Scope review and recovery

Staging exposed a fresh-checkout defect:root .gitattributes forces PS1 CRLF and
leaves JSON eol unspecified,which would invalidate physical SHA256 identities.
RED check-attr confirmed new source eol=crlf. Added only
docs/diagnostics/isolated-fuzz/.gitattributes:LF for these new utilities/copy,
their prior three-phase source/inventory fidelity dependencies,and the two prior
inventory JSON files used by offline regressions. No old file content,root
attribute,global Git config or security setting changed. New attribute SHA256:
4CEEF7E2230274253364BD477B550B26D8225412250C509446CD7DA488661DCA.
GREEN:Git cat-file --filters on staged blobs with core.autocrlf=true reproduced
working-tree SHA256 bytes for25 files (14 preparation artifacts including this
attribute,8 old source dependencies,3 old inventories). Raw stdout bytes were
hashed in memory;no temporary checkout/target execution. One initial helper
launch failed because Get-Command returned two Git paths;selecting the first
installed executable removed that ambiguity. These25 byte checks are additional
to the328 test cases,not target runs. A whole fresh-machine end-to-end restore
is still not claimed;unrelated historical files/packets remain outside scope.

Author self-review only against approved design and reviewer checklist:body-edge
placement,exact restoration,old-tool isolation,inventory/hash identity,distinct
frontiers,non-timeout error handling and no false formal acceptance inspected.
No unresolved in-scope correctness finding;not an independent review or merge
approval. Existing review0/1 remains unused/ineligible.

Rulings:PowerShell inline/portable ledger replaces Bash workflow helpers;manual
bookkeeping is less automated. Explicit offline-only approval overrides generic
full-target-suite/reviewer/cleanup steps;cost is unverified runtime ordering,
performance and missing fresh review. Historical raw packets are retained.
Ruling:scoped LF attributes preserve pinned bytes on fresh checkouts without
global/root config changes;cost if wrong is a checkout-byte policy difference
on the named diagnostic dependencies. No other expansion or deferred minor.

On a fresh computer use the development branch,read current status/recovery,
and obtain an existing authorized PowerShell7 runtime. The five offline scripts
above are portable from the repository;their synthetic fixtures need no ignored
raw records,credentials or old runtime cache. Do not invoke the copied target
as part of preparation. Git cannot recover ignored raw streams or execution
authority;this table records their prior identity check,not those raw files.

Next decision:separate approval would be needed before preparing/freezing a new
bounded controller and ONE ordinary fresh PS7 target run,30000ms,no warmup/retry/
fallback. Pin all source/controller/inventory/engine/receipt identities before
that run. Do not reuse frozen PS5 or increase timeout. Four output calls and new
source path/body/history can perturb timing;runtime order/span is UNVERIFIED.

All prior allowances persist:three-phase/split/boundary each1/1 spent;
candidate-full1/2 spent with PS5 FROZEN;other historical budgets unchanged.
Original timeout root cause unresolved. Candidate/I1/P1/main/C12/security/
runtime/release UNACCEPTED. No main merge or product promotion.
