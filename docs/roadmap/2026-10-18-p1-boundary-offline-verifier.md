# Offline boundary capture verifier — 47 synthetic cases and saved-output check

After1c9a31d7,user approved correcting a SEPARATE offline verifier using saved
output and synthetic anomalies. ZERO policy/product tests,diagnostic target
launches,warmups,retries or frozen-slot releases. No old controller/index/raw
was edited. PS5 remains frozen.

## Fix and test-first evidence

Added three portable files beside the existing isolated-fuzz utilities:
candidate-boundary-capture-verifier.ps1,its .tests.ps1 and .md usage guide.
The function reads only the approved pinned output inventory and classifies
supplied text/receipt facts. It never launches/imports the product or writes
records. It uses the inventory HostRule rather than copying the overescaped
frozen-controller regex.

Test-first sequence on PS7.6.5:
- Wrote the synthetic suite before implementation. First execution exit1 with
  ASSERTION FAILED: VerifierAvailable because the new verifier was absent.
  This was a missing-feature red,not a product test or all47 individual reds.
- Implemented the standalone verifier. Full offline suite exit0,47 cases PASS.
  The normal7.6.5 host case,incorrect-dot/suffix/wrong-host vectors catch the
  classifier admission defect. No mocked product or process invocation.
- Actual saved-output integration calls the new verifier only,using unchanged
  prior stdout/stderr/receipt with separately checked identities.

47 cases include normal completion with literal10/20/30/40ms expected pairs,
wrong host/literal-dot/suffix,child/later administrator context,changed ordinary
output,wrong/missing/duplicate/reordered/partial markers,negative/overflow/
noninteger/backward ticks within/across pairs,stderr,extra output,open pair,
missing tail,timeout,exit failure,over-budget,missing receipt fields,bad frequency,
nonfinite/negative/string elapsed,parent administrator,string boolean,unknown
cleanup,unterminated output and CRLF. Tests use real verifier behavior;fixed
synthetic ticks and expected durations are independent of verifier calculation.

Normal fixtures are generated from the committed91-line inventory with synthetic
identity and literal ticks. They are not copied private raw receipts. A fresh
clone can execute this offline suite without ignored packets or runtime caches
from this computer (PowerShell7 must be available). PS5 compatibility NOT tested.

## Saved-output integration, not a rerun

Read existing packet:
.superpowers/sdd/2026-10-18-p1-candidate-boundary-marker-run/.
Separately checked all344 pins,frozen manifest/controller/index,both streams and
one invocation before passing saved text/receipt to the function.

New result:Status=complete-diagnostic,CaptureValid=true,Complete=true,
FormalProjectAcceptance=false,LineCount91,PassCount64,MarkerCount8,
ClosedPairs4,OpenBoundary=null. Pair wall values:
InvalidInputs245.072ms,Build4096=646.361ms,Check4096=16860.926ms,
Check4097=14.711ms. These match the earlier independent read-only arithmetic.

The receipt still records CaptureValid=false,ExpectedResult=false because of
the old controller line2 host-regex defect. Those original fields,hashes and
run count remain unchanged. New offline assessment is separate,not historical
verdict rewriting or a renewed allowance. No target was run to test this tool.

## Contract and limitations

Valid complete classification requires ordered inventory,ordinary context,
typed receipt facts,positive integer frequency,finite nonnegative timing/CPU,
owned exit,empty stderr,exit0,no timeout and elapsed<=30000ms. Parse ticks as
Int64/nonnegative/nondecreasing;only emit complete START/END durations. Valid
partial output exposes an open boundary without inventing an end. Malformed
or unsupported captures are rejected;valid full text with failed execution
gates is incomplete,not complete-diagnostic.

This tool is NOT a hostile-input evaluator or authenticity/security boundary:
a caller can supply forged text/receipt facts. It does not independently prove
actual token,process ownership,clock frequency/CPU provenance or saved file
identity. Caller-side hash/source/engine/controller/receipt checks remain
mandatory. FormalProjectAcceptance is always false. No product-timeout repair,
stable performance,original full PASS or main/C12/I1/P1 acceptance follows.

Only the offline suite ran;the repository/product suites were deliberately NOT
run because this approved scope excludes product replay and prior budgets are
spent/frozen.47/47 refers exclusively to the new offline verifier cases.

## Identities and next step

| File | SHA256 |
|---|---|
| Verifier | CE8B9C988337C8D0B444D9CD10AA7CF45B037ADB5A8A70E4B6ACD11482BA8597 |
| Tests | B98DFB1172F84B328E38DD1FFEEB6ED2BDEFB198951F9F7DA616C92260780B1F |
| Pinned inventory,unchanged |6FF2D34D86C43ABFCDA28A6502ED7735D539C113665B902FC936C202E4F8FCF3|
| Prior frozen controller,unchanged |18EF13E927D43D59ADCBB97A209DAF2873FD82F53827883928B209ECBC3CB518|
| Prior index,unchanged |8FC69424C59CD9B58856820C93AEADEC979102822248B9CD50776A3231BB8CFE|

Current boundary-run1/1 spent;prior candidate-full1/2 spent/PS5 FROZEN.
All older budgets and review0/1 unused/ineligible remain. No additional execution
allowance or independent reviewer dispatch was used for this tool-only work.

Next recommend a bounded design of how to distinguish API evaluation from full
result comparison inside the observed Check4096 span while explicitly accounting
for scope/body/instrumentation changes. Do not prepare or launch another product
diagnostic,change safety guards or infer a culprit from this outer interval.
Keep the original timeout as unresolved. Recover tool/tests/guide/report/status/
runbook from the development branch,not ignored receipts/caches/credentials or
authority to replay product tests.
