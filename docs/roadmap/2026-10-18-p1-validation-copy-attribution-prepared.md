# Validation/copy attribution — PREPARED, NOT RUN

After9ebfe9e4,user confirmed preparation only:independent copy,strict summary
classifier,offline synthetic tests and static fidelity. No target run/controller,
warmup,retry,PS5,frozen-slot release or new independent review.

## Scope and exact changes

New folder:docs/diagnostics/isolated-fuzz/integer-candidate-validation-copy-attribution/.
Eleven files:eight PS1 copies,inventory,README and leaf .gitattributes.
Six PS1 sources byte-identical to ownership-marker copy. The other two change:
- phase-evaluator.ps1:rename only diagnostic entry to
  Get-OfflineSupervisionDecisionAttributionProbe,add exactly eight local
  instrumentation lines. Removing them/restoring entry recovers whole old file.
- offline-supervision-policy.tests.ps1:one caller command rename only. Whole
  prior file recovered by restoring that command name. Startup load unchanged.

Six collision-checked local longs initialize between RootSetup END and
ValidationCopy START. Per-row t0 precedes original failure-index/input assignment;
t1 immediately precedes original ordered snapshot;original six-field copy and
list Add remain untouched; t2 immediately follows Add. Validation/copy sums and
completed-row counter update after t2. No per-row output,new wrappers/shared
helper edits,global state,finally ends,guard/copy/assertion/order changes.

A single invariant ATTRIBUTION: ValidationCopy summary follows the existing
ValidationCopy END,before Decision START. It reports rows,validationTicks and
copyTicks. No summary is fabricated on exception/unfinished loop.
106 ordered nonempty lines/64 PASS/22 markers/11 pairs;summary is line76,not a
timing pair. Removing its rule restores every old105 inventory record exactly.
All old tools/inventories and412 frozen file bytes remain unchanged.

## Offline classifier semantics

Separate candidate-ownership-attribution-capture-verifier.ps1 defines only
Test-CandidateAttributionCapture,reads the pinned inventory and supplied text/
receipt facts,never imports/invokes target sources or writes receipts.
Retains all previous ordinary parent/child context,PS7 host,typed receipt,
newline/order/ticks/nesting/exit/30000ms gates and ownership frontiers.

Summary requires exactly4096 rows,nonnegative invariant Int64 values,proper
closed ValidationCopy parent/order,one summary,and exact sum <= parent ticks.
Decimal exact arithmetic prevents Int64 sum overflow or Double precision from
accepting a one-tick excess beyond2^53. Attribution milliseconds use Decimal
nearest ties-to-even;no fixed frequency assumed.
No summary:AttributionAvailable=false,Attribution=null (not zero).
Valid later partial prefix may retain totals;invalid capture never publishes
them. Hypothesis remains not-assessed unless ALL diagnostic completion and
execution gates pass. Complete validation>copy supports only this instrumented
ordering;copy>=validation falsifies it. FormalProjectAcceptance always false.
Not receipt authenticity,clock/token/source provenance or owned-process proof.

## Fresh verification and failures preserved

Existing pinned PS7.6.5 runtime,NoLogo/NoProfile/File. Final seven offline/static
scripts all exited0:
- candidate-ownership-attribution-capture-verifier.tests.ps1:160 cases,including
  all104 supported incomplete106-format prefixes (2..105 records),literal sums,
  absent versus zero,wrong rows/name/format/order/duplicates,Int64 overflow,
  exact2^53 boundary and one-tick excess,partial/error/timeout/full-exit gates.
- candidate-ownership-attribution-preparation.tests.ps1:21 cases:eight AST parses,
  six unchanged hashes,whole clone/caller restoration,unique definition/single
  caller,timer adjacency/three reads/no row output,typed locals/summary placement,
  entire inventory restoration and eleven-file scope.
- unchanged ownership classifier122 and ownership static20.
- unchanged boundary47,API-equality63,three-phase76.
Total509 cases;prefix subchecks are not additional cases. Only offline utilities
executed;target sources were read/AST-parsed only. Instrumented target runtime
behavior/output ordering has NOT been tested in this preparation-only scope.

RED before implementation:new fixture against unchanged105 classifier failed
exit1 ASSERTION FAILED: NormalPS7 for expected rejection of106 records,not import
or syntax failure. Initial new classifier GREEN158,later160 with exact large-
integer and whole-execution-gate coverage.
Initial static utility failed after18 successful checks:unused reference to
ScriptBlockAst.TryStatements (property does not exist). Read-only AST property
inspection confirmed ScriptBlockAst/EndBlock and zero source syntax errors;
removed ONLY that unused utility read. All21 static cases then passed. No target
fix,assertion weakening or target allowance used to address it.

Three new utilities AST-parse without errors. Engine SHA256:
362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139.
Prior412 frozen pins plus manifest/controller/index/raw streams checked unchanged;
single prior invocation remains. Latest old index:
7F8D8368B2F2263A85FF352ECB8756952DD4F151A13D5F344EF66993ECB9A292.

## Portable file identities (SHA256)

All14 artifacts below must be saved with this report/status/recovery.

| File(relative to docs/diagnostics/isolated-fuzz) | SHA256 |
|---|---|
| `integer-candidate-validation-copy-attribution/offline-supervision-policy.ps1` | `3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3` |
| `integer-candidate-validation-copy-attribution/offline-supervision-policy.ownership.tests.ps1` | `2A1BF87DC436DA1AE7232B937EC0DA1DDD375953FDF7CD71EE127909693EA43F` |
| `integer-candidate-validation-copy-attribution/offline-supervision-policy.reflection-reuse.tests.ps1` | `09BBAFD33ECED7DE1D98E00290A6B0D5E357E80F74A8766DBCC8409131C9B14F` |
| `integer-candidate-validation-copy-attribution/offline-supervision-policy.reflection-reuse-regression.ps1` | `0B7692FE426407AA9431D54BDB1AA964216FB695CD5A202045ABDDCEB5C61069` |
| `integer-candidate-validation-copy-attribution/offline-supervision-policy.shape-lookup.tests.ps1` | `9DAA38EF1F19CA2320D17FCA35CDA87A1B4DFBDFCE960AF40EFB223B9E245B6B` |
| `integer-candidate-validation-copy-attribution/offline-supervision-policy.shape-lookup-regression.ps1` | `12D6F5460F97B9E09453B0387C30E3D20CE19B2EDF1CDF4DBF4E77DA31561422` |
| `integer-candidate-validation-copy-attribution/phase-evaluator.ps1` | `95260FC356B0C6637BE1FFBC8E25388138A4E91464A161A232B7F8AD2B24B11E` |
| `integer-candidate-validation-copy-attribution/offline-supervision-policy.tests.ps1` | `6CF2455942D5D47DD92FFCD764F874FF8885A9B6FB8C01E3C28CA5AA56DD725B` |
| `integer-candidate-validation-copy-attribution/expected-output.json` | `6F65F6E195CF208A05F30ED4F6D1D2F168851C677E45A7FA84DC52190834B753` |
| `integer-candidate-validation-copy-attribution/.gitattributes` | `4403EB157DBAF72FA21AD9241F23ECAD1EE2472A47F0BA5BCC8EBE3A56D8123B` |
| `integer-candidate-validation-copy-attribution/README.md` | `81ED688C00EE6D2341563A18EF08A8E47F83BF4F6EAB263A6767077FDB4A1158` |
| `candidate-ownership-attribution-capture-verifier.ps1` | `A62683B07934486590DF8FDBBAAD6A34A3333FDF84C383FDED5F2C50D77A3B60` |
| `candidate-ownership-attribution-capture-verifier.tests.ps1` | `2CCFA9202B4F210AF8896C462855155C47B0B9AFABB440E9F53D70B12A166D36` |
| `candidate-ownership-attribution-preparation.tests.ps1` | `2EC573BE2BAFD5FB3B41F3712213CBEA4DFC787E861F5A8287F548F956156896` |

Recovery choice:old diagnostic .gitattributes is already pinned,so it was NOT
edited. New utility names match its existing candidate-ownership-*.ps1 LF rule;
the new folder's own .gitattributes fixes LF for its copies/JSON/README.
This adds one leaf attribute file versus the previous ten-file copy,not a new
runtime source/helper or global/root Git setting change.
Git cat-file --filters on staged blobs with core.autocrlf=true reproduced26
working-file hashes:14 new artifacts,8 prior source dependencies,4 prior
inventory dependencies. Bytes hashed in memory;no temporary checkout or target
execution. These26 checks are separate from509 offline/static cases.
A whole fresh-machine/runtime end-to-end restoration is not claimed.

## Review / limitations / next decision

Author self-review only:approved placements/restoration,strict row/sum/order/
overflow/prefix semantics,unchanged old tools/bytes,nullable availability and no
false hypothesis/acceptance inspected. No unresolved in-scope correctness finding.
Independent review0/1 remains unused/ineligible. Project/target suite deliberately
not run because approval is PREPARATION ONLY;offline results do not clear runtime
or project acceptance. No new plan/spec/controller/ignored runtime packet added.

Three timestamp reads per completed row imply12288 extra reads at4096,plus
assignments/accumulation/compiler/history/allocation effects. Timer/assignment
work is asymmetrically inside measured subintervals;loop/accounting also remains
outside them within parent. Parent-minus-sums is merely unattributed time,not
calibrated instrumentation cost. No empty-loop/baseline subtraction or pure
helper/CPU/original-run shares. Prior28983ms headroom was1017ms;instrumentation
could cause a timeout at the unchanged30000ms limit. No result is promised.

Preparation complete only. If desired,next requires separate approval before
creating/freezing a new bounded controller and ONE ordinary fresh PS7 target,
30000ms,no warmup/retry/fallback. No PS5 or frozen-slot reuse,guard weakening,
cache,second validation traversal,copy removal,timeout increase or main merge.

All historical budgets unchanged:ownership-marker/three-phase/split/boundary
each1/1spent,candidate-full1/2spent/PS5FROZEN;review0/1unused/ineligible.
Candidate/I1/P1/main/C12/security/runtime/release UNACCEPTED. Original timeout
unresolved;no actual106-record target output or attribution result exists.
Git restores portable sources/tests and this report,not ignored raw/controllers,
cached runtime,credentials or replay authority.
