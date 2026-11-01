# BOM guard repair — scoped validation passed

Status: BOM_REPAIR_VERIFIED; FORMAL_DATA_ASSESSMENT_NOT_RUN.
Explicit user approval “是” after exact comparison/regression/one complete
validation proposal. Baselinef7fec2261cd11be06d58ddbe638905efa0f89b96.

Changed only the BOM comparison in saved-data-contract.ps1:44 by specifying
StringComparison.Ordinal. Retained strict UTF8, LF/framing, JSON, array/type,
source attribution and matrix checks. An exact source-text delta guard passed
before testing. No production I1/I2 or receipt implementation changed.

Added bom-regression.tests.ps1 as a separate supplement; original frozen42group
source/expected model remain byte-identical. Under scoped zh-CN it checks real
validator behavior: original valid BOM-free data starting with literal ASCII {
is accepted with six exact expected summaries, and original bytes prepended
with literal EF BB BF are rejected with the exact contract mismatch. Culture
restored in finally, verified by actual results.

One focused RED before repair:1PASS/1FAIL/exit1, valid data rejected at original
line44 and genuine BOM rejected as expected. One combined validation after
repair:44PASS/0FAIL/0NotRun/exit0 = original42PASS + supplement2PASS. No extra
suite, hidden fix pass, expectation correction, diagnostic host or review.

Post-validation integrity: original16files/9directories/archive byte equality,
8bound inputs, original frozen tests/model and receipt all matched. Prior
stopped portable-final.json and root-cause-evidence.json unchanged. All prior
failed outcomes remain historical failures, not retroactively regraded.

Scope limits: this is finite saved-data checker validation, not full repository
or production/integration acceptance. Formal assessment and receipt invocation
were NOT run; compound receipt I/O remains dynamically unverified. No new
diagnostic/capture hosts, targets, compiles, reviewers, main merge or budget
release. Integration6/7/1/conditional1 remains FROZEN.

Recording packet: .superpowers/sdd/2026-11-01-offline-checker-bom-repair.
Portable records: new leaf/portable-bom-repair.json, separate from all previous
archives. Contains actual approval/ledger, before/validation source/test
snapshots, consumed intents and complete RED/validation/integrity tool records.
Original36negative fixtures remain transient in-memory; no individual byte
snapshot preservation or universal invalid-input coverage is claimed.

Next optional step is a separately approved, source-bound formal saved-data
assessment with the corrected checker. Passing tests alone do not authorize
that step or resume stopped integration/diagnostic attempts.

Portable archive17actual files/1root directory; decoded lengths/SHA256/bytes
matched originals. Archive SHA256:
9D1E05C9D8E0336D85123860F289BA59AD4D303FD14810059893709FB4C92105.
