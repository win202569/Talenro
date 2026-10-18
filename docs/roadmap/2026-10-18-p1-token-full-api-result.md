# Token full API and ownership regression result — 2026-10-18

Status: FOUR SCOPED SUITES PASSED / FORMAL ACCEPTANCE WITHHELD.
Parent checkpoint:bef133274503b174bd74469970ff18252401e9b2.
User approved preparation and up to four attempts in the
[bounded plan](../superpowers/plans/2026-10-18-token-full-api-bounded-regression.md).
All four ran exactly once in approved order;no retries or source repair.

## Actual receipts

|Slot|Start UTC|PID|Wall ms|CPU ms|Ordered lines|
|---|---|---|---|---|---|
|baseline-full|2026-10-18T10:35:11.3676503Z|33968|25681|33031.25|38/38|
|baseline-ownership|2026-10-18T10:35:43.6550207Z|32268|4404|5656.25|22/22|
|candidate-full|2026-10-18T10:35:54.4792448Z|47000|18883|23734.375|38/38|
|candidate-ownership|2026-10-18T10:36:18.7522441Z|25080|7228|10765.625|22/22|

Every host:ordinary parent/inherited token,pinned PS7 7.6.5 Core x64,
fresh no-profile noninteractive process,exit0,no timeout,owned-host-exited,
complete concurrently captured original-byte streams,zero stderr,
no ControllerError,pre/post pins intact,SuitePassed=true,FormalAcceptance=false.
Both baseline suites passed before either candidate launch.

Full suite35 named PASS groups plus context/host/limit=38 lines.
Ownership suite8 fixture+10 regression PASS groups plus framing/limits=22 lines.
These counts describe named groups and framing,not every nested assertion.
Each output was checked against an independently fixed ordered inventory;
the baseline/candidate stdout hashes also agree within each suite,but that
agreement alone was NOT the success criterion.

## What this establishes

Unchanged recursive equality and literal expected result builders checked the
full synthetic API output fields/type categories and refusal positions in the
covered scenarios. No summary-only comparison or oracle learned from baseline.
ArrayAndIntegerBounds retained4096 acceptance and4097 rejection;no shrinking.
InvalidObservation and ObservationScalars retained late-validation/refusal-index
checks. CallerMutation retained input-mutation protection and distinct Decision
references;SequentialIsolation retained distinct result roots,metadata mutation
and edited-result isolation. All ownership primitive/descriptor/identity fixture
checks ran before the ownership import. All callback-zero assertions retained.

The previously passed Token direct suite remains separate evidence,including
CodeProperty cases. This full API suite does not add universal CodeProperty/
concurrency/OS coverage;passing tested inputs is not proof of all possible inputs.
No product/candidate modification,promotion or trusted input bypass was made.

## Portable source preparation and exact-copy evidence

New tree:docs/diagnostics/isolated-fuzz/integer-token-api-regression/.
Each baseline/ and candidate/ directory contains exactly one selected module
and the same two byte-identical existing tests. No wrapper/import/assertion edits.
Tests copied first,then modules;all six physical SHA256 values match source
pins and PS syntax parsing has zero errors. Relative imports route to only the
local module in each fresh host. Full uses -Group all;ownership uses -Group
regression,which keeps fixture checks without running/loading full again.

|Artifact|SHA256|
|---|---|
|both offline-supervision-policy.tests.ps1 copies|61A3BACF26EDE1F8B137CBF77CC4DBC35E4D6E6FB8FE6365953CDA9F41CEDEA3|
|both offline-supervision-policy.ownership.tests.ps1 copies|C4905925B309B00E94834B2FDA3EE8034265B23CDF274276C9ECA5511250C6E7|
|baseline/offline-supervision-policy.ps1|3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3|
|candidate/offline-supervision-policy.ps1|F9B4E408D9A4FF9C7A36EFD6E1F2472F17DACB3F76F463FF4973DAB4139CC82A|
|.gitattributes|A79691A93B46E49CE460C26EF22AFCC03D6ECA1E63BF2EDBC20E96159510F6C9|
|expected-output.json|32A6C13F77635D4B768DC1637BBE885641AAAE96912121A2C44301F64CA09F07|
|source-pins.json|4C5D7C85C04D0390ABA01756B96365721F732F831CC7A8C6E9FCF0333C7F59BF|
|README.md|A1B3887D69E6C291887C86D5FF4B82ABA4D53448DE7AC24D3921DBCE9C23FB24|

Leaf LF rules preserve physical source pins on fresh Windows checkout.
source-pins.json gives repository-relative provenance,no ignored-session
dependency. expected-output.json contains fixed literal lines with only
{Context} resolved to ordinary parent identity at freeze time;host version fixed.
No dynamic target-derived oracle,clock instrumentation or new support compilation.

## Frozen evidence (ignored,not committed)

Packet:.superpowers/sdd/2026-10-18-p1-token-full-api-run/

|File/capture|SHA256|
|---|---|
|run.ps1|9CD2CE4E15828F5A7B5A302DF3079305ADEFE7461E1AEDE19B01466789BE5719|
|frozen.json|116C9A42CA26321B1E8ECCCCF89D2A418B0F60852E87898A2B2BD63AA4CD2694|
|invocations.json|249D30A70D006B60BAFFC13193FF8F480CB65607480C3F45A833EA0C3B638706|
|both full stdout files|3FE31C43E6947167D06A1C03C94A382856B249C36A1FAA0DBF2BBD6C99342D76|
|both ownership stdout files|0571AD4FEE3BA556B3D5A29D0E36A4C7AA5C3CAA95E97F2FF0556DA8D53464B8|
|all four stderr files|E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855|

All476 frozen pins verified before/after,including all457 inherited identities,
latest direct manifest/index/streams,source provenance,new files/controller/plan.
Engine SHA256362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139.
Completed current receipts/streams were re-read and hash/oracle-checked as data,
without another host. Prior failed and successful records remain unchanged.

## Accounting,limits and next step

NEW full API allowance4/4SPENT,one per slot. Prior direct2/2SPENT,older direct
failed baseline1/1SPENT and candidate0/1UNUSED/FROZEN,all historical diagnostics
and legacy candidate-full1/2/PS5 freeze unchanged. No slot reset or retry.
No independent reviewer launched;review0/1unused/ineligible unchanged.
Author self-review covers source-copy/controller/static safety only.

Ruling:approved bounded terms override generic skill defaults for additional
whole-project/test hosts,review,installs,Bash helpers,retries and scratch deletion.
Cost:weaker independent/whole-project evidence;wider acceptance withheld.
No production bugfix/RED-GREEN claim:unchanged selected modules passed tests.
No skipped4096/4097 bound,assertion/fixture relaxation,pre-import,warmup,reload,
PS5,alternative engine,native/live supervision,whole-project test run or merge.

The four elapsed/CPU receipts are NOT a benchmark. No speedup or previous
diagnostic-timeout root-cause/repair is established by them;ownership candidate
elapsed is even larger than baseline in these individual hosts. Do not select
favorable timings or convert this to performance acceptance.

For this PS7 tested-case scope,full API literal equality/refusal-index/result
ownership regressions are now passed in addition to the earlier direct suite.
Candidate stays isolated/unpromoted;I1/P1/main/C12/security/runtime/release/
whole-project acceptance remains withheld. Reflection/shape extra suites,
independent review and environment/performance gates are not automatically
cleared by these results. No main merge.

Next recommend READ-ONLY promotion-gate inventory to identify remaining evidence
and independent-review eligibility before proposing promotion or more runs.
No further target execution or reviewer is authorized by this exhausted allowance.

Recover from latest codex/c12-b01-task9-coordinator commit,not main. Git preserves
the ten portable source/routing/oracle files and report/status/recovery,not ignored
controller/raw records,caches,credentials or replay rights. Missing raw on another
computer does not reset spent/frozen slots. Engine/dependencies need their own
preparation on a new computer;cloning alone is not runtime acceptance.
