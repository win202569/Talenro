# API/equality split — PREPARED, target NOT RUN

After2a369070,user confirmed the bounded preparation design. Added an isolated
diagnostic copy plus separate95-line nested offline verifier/tests.
ZERO target/product tests,diagnostic launches,retries,warmups,controllers,
reviewers,old-record edits or frozen-slot releases. Preparation only.

## Prepared source and fidelity evidence

Folder: docs/diagnostics/isolated-fuzz/integer-candidate-api-equality/.
Seven PowerShell sources,expected-output.json and README.

Six sources,including3537E9F5 candidate module,are byte-identical to the measured
outer-marker folder. Only original35 tests adds P1-SplitCheck4096 at script level
and substitutes its name at the single4096 call. Same parameters,local r,
API command/arguments,full P1-Equal/P1-Need assertion and return;four direct
SUBPHASE records bracket API assignment and comparison/assertion.

Static AST checks:
- Seven sources parse;all original helper function texts remain byte-identical.
- Exactly one extra helper and one invocation. Its two parameters and three
  non-marker statement texts equal those of original P1-Check,in order.
- Remove that helper and restore one caller name:entire test text equals the
  measured outer-marker copy exactly.
- No product reload,changed input/assertion,common-helper branch or per-row probe.

These are static claims,not runtime scope/binding/function-history equivalence.
The helper's fresh identity/body/history and added output can perturb performance.
Product source/load declarations and all safety checks remain unchanged.

## Frozen prospective output and separate classifier

New inventory:95 ordered nonempty records,64 PASS,12 timing markers,six pairs.
The91 preexisting records remain in order;four inner records are added inside
Check4096 (Evaluate start/end,Assert start/end). No test-count change.

New function Test-CandidateAPIEqualityCapture in the separate
candidate-api-equality-capture-verifier.ps1 pins ONLY the new inventory.
It validates ordered host/context/literal/timing records,positive integer
frequency,monotonic nonnegative Int64 ticks,and exact parent/child stack:
outer Check4096 contains Evaluate and Assert siblings. Only closed pairs get
durations;OpenScopes retains all unfinished parents/children. Full completion
also requires91+4 records/64 PASS/12 markers/six closed pairs,empty stderr,
ordinary context/typed receipt facts,owned exit,exit0,no timeout,wall<=30000ms.
FormalProjectAcceptance is always false.

The prior91-line verifier and47-case suite are unchanged. No generic loosening,
inventory replacement,old receipt rewriting or post-timeout tail reconstruction.
This classifier is not an authenticity/security boundary:caller must separately
verify saved file hashes,source/engine/controller/receipt count/provenance and
actual token/clock/ownership evidence. Valid text cannot authenticate forged facts.

## Offline test-first results

PS7.6.5 only. Wrote new synthetic tests before implementation;first run exited1
ASSERTION FAILED: VerifierAvailable (missing-feature red,not63 separate reds).
Implemented distinct classifier;full new suite exit0,63/63 PASS.

Inherited47 vectors cover normal host,wrong host/context/literals/marker names/
edges/order/absence/duplication/partial output,invalid/overflow/backward ticks,
stderr,extra output,open prefix,missing tail,timeout/exit/budget,missing/invalid
receipt fields,frequency/timing/boolean/cleanup gates,newline and CRLF.
Sixteen additional cases cover inner order,parent containment,sibling chronology,
premature parent end,duplicate inner end,inner overflow,three parent/child open
frontiers,zero-width interval,legacy91 rejection,and four invalid CPU values.
Hand-fixed synthetic tick values yield literal closed durations10/20/30/5/60/10ms
in close order.60ms is the synthetic outer Check4096 parent,NOT additive work.
No dynamic4096 fixture or policy evaluator is invoked by these offline tests.

PS5 compatibility and real target output/binding/timing remain UNVERIFIED.
There is no actual95-line saved-output integration because no target has run.
Repository/product suites deliberately NOT run under preparation-only scope.
Old47-case suite rerun separately to verify it remains intact;results below
describe offline tools only,not the project suite.

## Identities

| File | SHA256 |
|---|---|
| Candidate product/copy |3537E9F5F203E181042DD9CCE27C2962096B72A7D25842B968145D2066255BD3|
| New marked original35 tests |FC85C3851FCE229C6326658D3480407A0B68C2D7CF14F37C3DA291D2F8EDD820|
| New95-line inventory |224A640A30F0B4F0AC891DB81AC545F3B2D621AA1865733EF88530942C807940|
| New nested verifier |803F55C79F40D481C4AC07E5FDEF4C7D024CC34A7C4C79CFBD57A5E29D39B1AA|
| New63-case tests |4C87237E421891E7D553ECEC4F9BF18930AB0E8F00679B70843C729F60B4BF6F|

Old344 pins and prior invocation index remain unchanged. Original product953CC2A8,
old91-line verifier CE8B9C98/tests B98DFB11 and historical false classifier remain
intact. No preparation authority silently renews run budgets.

## Next scoped recommendation

Separately approve ONE NEW ordinary fresh PS7 diagnostic,30000ms,no retry/warmup/
PS5 fallback,freeze source/inventory/controller identities before launch,preserve
raw and owned-process limits. Do NOT run directly or reuse spent/frozen slots.
No controller prepared in this step. A future split can localize API versus
assertion wall intervals,not evaluator internals or the root cause uniquely.
Do not claim speedup,fix or add parent and child intervals together.

Current boundary-run1/1 spent,candidate-full1/2 spent/PS5 FROZEN,older budgets
unchanged;review0/1 unused/ineligible. Candidate/I1/P1/main/C12/security/runtime/
release remain unaccepted. Recover source/inventory/tools/tests/report/status/
runbook from development branch,not ignored raw/cache/credentials or replay
authority. Preparation README/source UNVERIFIED wording is intentional.
