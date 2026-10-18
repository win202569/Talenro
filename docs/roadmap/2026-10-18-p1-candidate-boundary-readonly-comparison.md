# Candidate boundary read-only comparison — 2026-10-18

After8b7eb852,user approved read-only boundary/call-path comparison. ZERO target
test/diagnostic launches,product/test/controller/raw edits,retries,warmups or
slot releases. AST parsing and file/receipt reads only. This is implementer
analysis,not independent review,fix,benchmark or acceptance.

## Source trace and unchanged boundaries

Paths relative to docs/diagnostics/isolated-fuzz/integer-type-reuse-full-regression/.
tests.ps1 below abbreviates offline-supervision-policy.tests.ps1.

- tests.ps1:6–10 P1-Case invokes the case body and emits PASS only AFTER return.
  Therefore PassiveDescriptors PASS proves that case returned,not that the
  next case started or which instruction later consumed time.
- tests.ps1:168–175 ArrayAndIntegerBounds runs invalid integer/array cases,
  adds/removes a shadow Length,builds4096 ordinary time observations,checks4096,
  then constructs/checks4097. Its single PASS follows all these operations.
- tests.ps1:54 P1-Check calls Get-OfflineSupervisionDecision,compares the full
  result through P1-Equal/P1-Need,then returns. The boundary call:173 includes
  scenario/expected argument evaluation,API evaluation and equality.
- Product:114–131 performs root validation and the4096 maximum-array gate.
  Product:132–163 validates each record and snapshots its six fields.
  Product:168 onward performs a separate decision loop over owned observations.
  The4097 test should be refused at the root length gate before per-row work;
  that source expectation is not a new measured duration or PASS.
- All six original test/wrapper copies remain byte-identical. AST comparison
  finds13 product functions;12 full function texts,including the evaluator,
  are identical to original. Only Assert-SupervisionPolicyInteger differs.
  No target script was dot-sourced or invoked during this comparison.

Conditional clean-time-row model: candidate makes16 SDK type queries rather
than18,while retaining10 SDK extended-Members reads per row (eight Passive
calls plus two inline Integer guards). A completed4096 validation would give
65536 type queries and40960 member reads. These are source-model counts,NOT
dynamic counts from the timeout,measured costs or latency guarantees. Shape
still performs full passive checks on all six values;Kind/Actor token checks
still invoke Passive. Metadata/shape allowlists were already evaluation-local.
No further guard removal or input-acceptance cache is justified.

## Existing outcomes are not a matched experiment

| Saved path | Engine | Wall ms | Observed frontier |
|---|---|---:|---|
| Original complete entry | PS7 | 10528 | Complete83 lines/64 PASS |
| Original complete entry | PS5 | 30079 timeout |65 lines/48 PASS;last PassiveDescriptors |
| Progress-instrumented original copy | PS5 | 30062 timeout |ValidationCopy open;1536 completed-row snapshot |
| Region-instrumented original copy | PS5 | 18993 |Complete diagnostic;4096 evaluation8783.485ms |
| Unchanged original local reload diagnostic | PS5 | 30081 timeout |Evaluate4096 open despite fresh definitions |
| Integer candidate complete entry | PS7 | 30090 timeout |65 lines/48 PASS;last PassiveDescriptors |

Compared saved invocation receipts against preceding reports;latest327 pins and
candidate streams held. Historical packet receipts are inherited pinned inputs.
Region diagnostic includes extra instrumentation and special helper loading;
it is NOT original-entry acceptance. Its assertion interval7.773ms and4097
interval5.482ms apply only to that run. Progress copy reached1536 completed rows,
so that run was making progress,not demonstrated permanent deadlock.
The latest candidate has no boundary/internal timing or row-count markers.
Older PS5 observations cannot identify its active PS7 instruction or row count.

Counterevidence narrows claims:
local reload is NOT sufficient for completion;integer type-query reduction is
NOT a demonstrated full-flow solution;one original PS7 completion and one
candidate PS7 timeout do NOT establish stable regression or causal effect.
Product/body/loading/history/instrumentation and ambient runtime conditions
remain confounded. CPU covers all process threads and may exceed wall.
No profiler/stack,GC,JIT,dispatch or exact SDK-helper attribution is available.

Sources:
[full candidate result](2026-10-18-p1-integer-candidate-full-regression.md),
[original full](2026-10-17-p1-original-full-verification.md),
[progress](2026-10-18-p1-ps5-validation-copy-progress.md),
[regions](2026-10-18-p1-ps5-validation-regions.md),
[local reload](2026-10-18-p1-loading-context-comparison.md),
[original static model](2026-10-18-p1-hot-loop-readonly-cost-model.md).

## One falsifiable localization hypothesis

H-boundary: in a prospective fresh candidate PS7 diagnostic preserving the full
test order,the4096 P1-Check region is the longest COMPLETED boundary subregion,
or is the region left open at timeout after invalid-input and construction
regions have closed. This is a localization hypothesis,NOT a unique root cause.

Proposed minimal evidence: start/end markers around four unchanged statement
blocks inside the existing ArrayAndIntegerBounds body:InvalidInputs,Build4096,
Check4096 and Check4097. Keep P1-Case/P1-Check/P1-Equal and every product function
byte-identical;no special helper,no product reload,no inner-loop timestamps,
counters,changed assertions,changed inputs or lowered safety checks.
Record outer process timing and marker timestamps;do not call the Check4096
interval API-only time because arguments/equality are included.

Interpret prospectively:
- InvalidInputs/Build open,or timeout before first marker: H not supported for
  that run;investigate the observed earlier frontier,do not assume4096 evaluation.
- Check4096 start without end after earlier closes: supports localization only;
  cannot separate API,arguments,equality,engine/runtime work or a permanent hang.
- All regions close: compare completed spans;Check4096 not longest falsifies
  this hypothesis. A complete fast run is diagnostic evidence,not reproduction
  of the earlier failure or acceptance of unchanged entry.
- Check4097/later path open: retain actual frontier;do not relabel as4096 hang.
- Malformed/truncated markers,assertion errors or source drift: invalid/incomplete
  diagnostic;no inferred tail,repair-and-rerun or automatic fallback.

## Next scoped recommendation; not executed

Prepare ONLY a separately labeled portable candidate diagnostic copy and static
fidelity/output checks for this outer-marker design. Freeze expected merged
output and marker order BEFORE seeking execution approval. Do not prepare an
autorunning controller or launch a target as part of preparation.
A subsequent proposal may request ONE fresh ordinary PS7 run,30000ms,no retry/
warmup;that would be a NEW allowance,not reuse of frozen PS5. No run approved here.

Original product953CC2A8/candidate3537E9F5 and all existing test bytes unchanged.
Current full allowance1/2 spent,PS5 FROZEN;all prior spent/frozen budgets and
review0/1 unused/ineligible remain. Candidate/I1/P1/main/C12/security/runtime/
release acceptance blocked. Preserve this analysis/status/recovery on the existing
development branch;Git does not restore ignored raw/caches/credentials/replay
authority. No speculative optimization or timeout increase recommended.
