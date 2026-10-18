# Strict duplicate JSON keys — checker-only offline regression

Approved in chat after2d020e50: preserve all old frozen files and RED evidence,
create a revised checker and one offline batch.No candidate import/run or old
slot reuse. Max1new ordinary PS7 host,30s+2s owned exit+2s drain;failure saves
results without repair/retry. Original37cases plus7newcases=44.

inspect-capture.ps1 uses the existing runtime's System.Text.Json document parser
to enumerate decoded ROOT property names before ConvertFrom-Json can collapse
them. Root name uniqueness uses OrdinalIgnoreCase to reject aliases in the
PowerShell property namespace. TIME and FACT both use this path. The existing
flat schema/type/order/counter/tick/acceptance checks and converter stay intact.
No regex extraction of JSON names. Unicode escape spellings are compared by
their decoded names;unique legal escaped names retain the original schema.
JSON document is disposed even on rejection;no new dependency is installed.
The fixed schemas are flat;this does not introduce a general nested JSON API.

New regressions:TIME case/escaped duplicates;FACT exact/case/escaped duplicates;
legal single escaped TIME/FACT names.Duplicate variants require the explicit
duplicate-json-property reason,not any unrelated exception. Existing37cases
and the old exact-duplicate RED case remain in the copied harness.
Historical36/37RED is retained,not rerun against the old checker.New assertions
have no fresh old-checker negative-control run;report that limitation.

Original output-inventory and three logger helpers are read from the old stage
leaf. Only AST-selected logger definitions execute;the ownership test's top-level
code and product functions never execute. Synthetic protocol records cannot
grant product/runtime acceptance. This is not an actual instrumented candidate
run and says nothing new about the original30s ownership timeout.

capture.ps1 pins597inherited entries,four latest packet anchors,and these six
portable files before launch/consumption. The old controller/packets remain
untouched. Raw streams and synthetic files stay ignored. Read the roadmap
result before any future action. No reviewer,main merge,I1/security/runtime/
C12/release/PS5/performance acceptance is authorized by this isolated batch.
