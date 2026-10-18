# Offline timing logger/checker validation

Approved after a90555f6: validate synthetic saved records and ONLY the three
AST-selected logger function definitions. Never dot-source the instrumented
test or import/invoke the candidate. One bounded ordinary PS7 batch,37cases,
30s wall+2s owned exit+2s drain;no repair/retry after capture. Collect all fixed
offline case outcomes,including failures,within that batch. This is not a
candidate diagnostic allowance and does not unfreeze any old slots.

Independent protocol fixtures test valid complete text,closed/open/empty
prefixes,missing/duplicate/reordered records,clock/accounting/counter errors,
malformed/truncated/extra/duplicate JSON fields,false/missing ownership facts,
complete-mode rejection and logger output identity/nesting/end-stage guards.
Synthetic complete text can only yield COMPLETE_TEXT_ONLY,never acceptance.
Fixture construction uses the declared inventory as protocol data,not product
helpers or their values. No product-mutant sensitivity or ownership proof.

capture.ps1 pins all582prior entries,four prior packet anchors,seven prepared
files and this leaf before consuming its single offline attempt. Raw streams,
synthetic input files and receipts remain ignored. Preserve failures without
patching a90555f6 sources merely to rerun. Read the roadmap result before acting.
