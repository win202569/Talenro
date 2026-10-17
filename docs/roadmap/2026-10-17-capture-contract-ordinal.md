# Capture contract exact string matching — bounded correction

## Authority and scope

User approved short in-chat bounded design after970a5e2c15a165b932e805490d3ba1d1744d1b30:
only capture-adapter-contract.ps1/tests and preservation docs;four NEW synthetic
PS7RED/PS5RED/PS7GREEN/PS5GREEN hosts,120s each,no retries,then one readonly
review. If reproduction absent,stop/no production change. No formal spec/plan
document for bounded path;existing synthetic contract design remains authority.
No unchanged dependency/interface/schema/live collector/runtime/security changes.

## Confirmed failure and correction

Both RED hosts reproduced the same6 accepted non-exact inputs:embedded-null
encoding,soft-hyphen encoding,raw event kind,raw channel,stored channel,and an
altered equal-length canonical-equivalent derived OriginDeclaration. This
changes the earlier unverified comparison concern into observed contract failures.

[PowerShell documentation](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_comparison_operators)
describes InvariantCulture comparison;case-sensitive does not mean ordinal.
The issue is linguistic equivalence,not merely a machine's current UI culture.
The equal-length derived fixture swaps composed/decomposed accented sequences
while leaving the genuine manifest/hash untouched;the old verifier accepted it.
Expectations use literal input/refusal and ordinal comparisons,not production
encoder or helper results. Ordinary Unicode metadata is still allowed and kept
distinct without normalization;both original declarations verify with different
binding hashes. Invalid token objects are refused without ToString callbacks.

Implementation:strict-string ordinal equality/membership helpers,exact shape
field-name membership,SourceMode/EncodingName/Policy checks,raw event kind/channel
and field preflight before unchanged simulator validation,stored kind preflight,
snapshot/hash comparisons,and full recursive derived-string equality. Remaining
internal ASCII switch/comparisons consume already validated exact tokens,not raw
caller text. CLR type-name checks are trusted metadata. No string normalization,
case folding or substitution. Passive member/type/byte bounds and independent
empty-array snapshots unchanged. No modifications to simulator/package/classifier/
store;their standalone comparison behavior is not newly certified by this task.

## Fresh finite verification —4/4 consumed

|Call|Host/phase|Exit|Elapsed ms|PASS|FAIL|Stderr artifact bytes|
|---|---|---|---|---|---|---|
|1|PS7 RED|1|1236|11|6|375|
|2|PS5 RED|1|1282|11|6|572|
|3|PS7 GREEN|0|12580|110|0|0|
|4|PS5 GREEN|0|50924|110|0|0|

REDs ended at expected OrdinalContractMustRejectNonExactText aggregate assertion,
not fixture errors. Observed7.6.5Core x64 and5.1.26100.9711Desktop x64. Actual
child records/full outputs checked. Each GREEN includes original93 checks plus
16 ordinal checks and aggregate;all110 passed. Ten new checks already passed
before correction;only six failure cases are claimed RED->GREEN. No source/test
edit between GREEN hosts or after verification. No full-project/old-suite replay,
baseline,dependency installation or additional diagnostic host. Decoded test text
receipts are not real raw capture/provenance evidence. All old budgets/frozen
slots and earlier failures retained,not renewed or transferred.

Source SHA256:`116FCAAA6506ADC6969D718AC61F8C79908BDED48ADA1204C2E8A60F33A35BB2`.
Tests:`5A6975AAF9998D62304BA3DDB7483A6CE1A1382FBE9D05EBDA209BD7AE71566F`.
Four-entry index:`5AA97E7D77BAF523B4B81D1F38607421A14E42CD07F74A8A516646630C65698F`.
Ignored audit artifacts under .superpowers/sdd/2026-10-17-capture-contract-ordinal
remain local;portable source/tests require neither these files nor private caches.

## Review focus and disposition — scoped checkpoint accepted

One fresh scoped readonly review completed:0Critical/0Important/0Minor findings.
Verdict acceptable for bounded synthetic-contract development checkpoint,not
merge/production readiness. It confirmed exact admission,derived-string matching,
Unicode preservation,noncoercing guards,and that remaining internal linguistic
branches consume exact admitted/internal tokens. Prior passive member/length/
rank and owned-empty-array protections remain intact. Complete source/tests,
spec/authority/ledger,preservation diff and relevant dependency paths inspected;
reviewer independently matched source/test hashes. No reviewer execution,
source loading,network,edits,Git mutation or subdelegation. Parent owns execution
evidence;no second reviewer or extra fix/test reserve used.

Declined judgments and parent rulings:

- Independent platform execution/full-project regressions:static review and
  exhausted cap. Accepted limit;parent's two complete110-check runs are actual
  evidence,not independent re-execution or whole-project certification.
- Standalone simulator/package/classifier/store,including unchanged XML string
  comparisons:explicit unchanged-dependency boundary. Consumed paths checked,
  not exhaustively certified. Cost:legacy standalone/Unicode XML variants remain
  unmeasured;this checkpoint does not claim every project string comparison fixed.
- Arbitrary direct calls to internal helpers:only constructor/verifier are the
  supported contract boundary. Accepted;functions are not a security boundary or
  separate guarantee for bypassing public preflight. Cost:unsupported calls may
  not enforce the public contract's full validation sequence.
- Concurrent mutation,hostile process-wide tampering/resource exhaustion:
  expressly excluded synchronous synthetic guarantee. Bounds/callback tests do
  not create a sandbox. Cost:no stronger adversarial/process-wide resilience claim.
- Live capture/provenance/owned stopping/OS cleanup:synthetic declarations cannot
  prove actual observations. Accepted;cost:real integration/gates remain missing.
- Durability,security/vendor disposition,Stage B,I2/I3/full C12:not implemented
  or authorized. Accepted;cost:no deployment/merge/acceptance progress on these
  gates from a green synthetic suite. No declined judgment is treated as a pass.

No confirmed minor deferred. Original comparison concern is now reproduced and
fixed at this contract boundary;unchanged dependency comparison concerns above
are unmeasured,not silently certified. Preserve development branch and audit
records under standing rule;no PR/main merge,force push or worktree deletion.

Ruling:PowerShell-native bookkeeping and finite full-contract suites instead of
Bash helpers/extra baseline/task-done/full-project runs;retain ignored audit
workspace. Approved cap/preservation prevails. Cost:no full-project regression
certificate;raw receipts not restored by cloning. No new guarantee for live
capture,clock/command provenance,owned stopping/cleanup,concurrent caller mutation,
hostile process-wide tampering,resource exhaustion,durability,security/vendor
disposition,Stage B,I2/I3 or full C12 acceptance. All authority/action flags false.
