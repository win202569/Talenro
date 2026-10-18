# Original versus diagnostic loading/scope audit — 2026-10-18

After2697b28e, the user approved the recommended read-only source comparison.
ZERO test/diagnostic hosts were launched, no source/controller/raw was edited,
no allowance consumed or renewed. This is implementer analysis, not independent
review, a runtime provenance attestation, a fix or acceptance.

## Key finding

The successful diagnostic copies do more than add clocks: their dedicated4096
check reloads the entire product function set inside P1-BoundaryCheck4096.
The original check calls the product previously loaded at test-script level.
Thus timing instrumentation, function-definition loading context and invocation
history are confounded. This difference is present in BOTH diagnostic A and B,
so their reverse-order comparison does not isolate it.

## Static call and loading paths

All paths below are relative to docs/diagnostics/isolated-fuzz/.
Table basenames omit the common offline-supervision-policy. prefix where applicable.

| Step | Original complete entry | Diagnostic A/B |
|---|---|---|
| Entry | shape-lookup-regression.ps1 | probe.ps1 in its diagnostic folder |
| Shape stage | child shape tests load original product | copied child shape tests load ../original product |
| Next wrapper | reflection-reuse-regression.ps1 | existing.ps1 with boundary timers |
| Reflection stage | child reflection tests load original product | copied child reflection tests load ../original product |
| Ownership runner | ownership.tests.ps1 -Group all | copied ownership.tests.ps1 -Group all |
| Before original35 | fixture8, then child original35 tests | same order, copied scripts plus timers |
| Original35 startup | tests.ps1:60 dot-sources original product | copied tests.ps1:75 dot-sources ../original product |
|4096 check | tests.ps1:173 calls P1-Check | copied tests.ps1:193 calls P1-BoundaryCheck4096 |
| Product load at4096 | none inside P1-Check | copied tests.ps1:68 dot-sources phase-copy inside helper before Evaluate timer |
| After original35 | ownership runner loads original product at:187 | copied runner loads ../original product at:197 |

Wrapper invocations use &; product loads use dot-sourcing. Original P1-Check
(tests.ps1:54) calls Get-OfflineSupervisionDecision, compares its full result with
P1-Equal/P1-Need and returns it. The diagnostic special helper retains those
operations and argument order but first loads phase-copy, then brackets evaluation
and equality separately. The helper returns after that call;4097 refusal and later
ordinary checks use unchanged P1-Check. Static scoping predicts local definitions
for that dedicated invocation, not permanent replacement of the script-level
functions. Actual command binding/definition identity before and after was NOT
captured dynamically, so this audit does not certify runtime isolation.

The dedicated reload defines all13 original product functions, not only the main
entry, plus one snapshot helper. Original main contains Set-StrictMode Latest;
it creates per-evaluation metadata/shape sets. No script/global cache, import,
Add-Type or process launch was found in that original product. This excludes
those explicit source mechanisms, not engine caches, inherited scope or ambient
runtime effects. No engine internals or measured cache-reset behavior is asserted.

## Checks performed without executing target scripts

- Parsed source ASTs only; did not dot-source or invoke product/tests.
- Both product copies contain the same12 original helper function bodies, byte
  for byte. Removing diagnostic header, VC/REGION blocks and ProductPhase marker
  lines restores the FULL original product text after newline/end normalization.
- All16 original35 helper definitions except the deliberately timed P1-Case are
  byte-identical. Diff shows added event calls, dedicated boundary helper,
  adjusted relative product path and the single4096 call substitution; original
  invalid-input construction,4096 array construction, expected4095 last sequence,
  equality and4097 refusal remain present.
- Original35 case names/order35 and ownership fixture+regression names/order18
  match via AST command extraction. Shape6/reflection3 case assignment order also
  matches (these scripts use literal assignments, not case-helper commands).
  This is static inventory, not fresh PASS evidence.
- Six shared A/B scripts are byte-identical; only their product phase copies
  differ. Their common special helper has the same reload behavior.
- All270 latest frozen pins held. Fresh-original manifest/controller/index/raw
  hashes match the preceding report. Latest invocation counts remain original1,
  reverse pair2; no new test slots were launched.

Product SHA953CC2A8 and original entry12D6F546 remain unchanged (full hashes in
the [fresh original PS5 report](2026-10-18-p1-original-ps5-fresh.md)). A product
copy37935F0E and B product copyD3D06344 also match their retained full hashes.

## Hypothesis, counterevidence and next boundary

H-load: fresh local function definitions/loading context at the4096 call may
contribute to the diagnostic-versus-original timing gap. Source supports a real
structural difference; it does NOT prove the performance mechanism or cause.
Freshness, local versus inherited resolution, compiled body history, source
path/body changes and instrumentation cannot yet be distinguished.

Important counterevidence: the historical A diagnostic ALSO performed this local
reload and still timed out. Therefore reload is not established as sufficient for
completion, nor is inherited scope established as necessary for timeout. The
recent B/A diagnostic pair completed13.280/13.425s while the unchanged original
fresh run timed out30.070s; these are not a matched single-variable experiment.
Original timeout's last PASS cannot prove it reached the dedicated4096 call.

Next recommendation is to design a bounded single-variable loading-context
comparison using the UNCHANGED original product in both arms, identical external
boundary-only markers/full test order/assertions, no product timing counters and
explicit binding provenance. Compare inherited loading against local reload;
do not mix it with helper optimization or extra warmup. Predefine expected
output separately rather than reusing diagnostic inventories containing inner
product events. Any launch needs a separately approved fresh allowance with the
existing30s cap and stop-on-failure rule. No such experiment was implemented or
launched here; this audit grants no replay authority.

PS5/I1/P1/candidate acceptance remains blocked. All spent/frozen budgets and
review0/1 unused/ineligible remain unchanged. No main/C12/runtime/release/security
acceptance. Recover this report, preceding evidence summaries and committed
source on codex/c12-b01-task9-coordinator; ignored raw/controllers/runtime caches
are not portable Git prerequisites.
