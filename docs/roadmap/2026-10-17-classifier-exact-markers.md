# Classifier exact XML and encoding markers — bounded correction

## Authority and boundary

User approved short classifier-only design after4078706e34947a7b6d3b7e0daf0c1164bf8e513b:
four NEW sequential synthetic hosts PS7RED/PS5RED/PS7GREEN/PS5GREEN,120s each,
no retries,then one readonly review. Reproduction required before source change;
unexpected result stops. This is a bounded change to existing flow,no new formal
spec/implementation plan. Only classifier/tests and preservation docs changed.
No package,store,simulation,capture-contract,format,public signature or schema
changes. No real collection/runtime/security or merge authority follows.

## Observed defect and minimal correction

Both REDs reproduced22 false acceptances:9 legal XML ignorable-character cases
in each output encoding,and4 encoding-name variants. U+00AD in channel,SourceId/
Record markers,Completed/Processing,type names,version and namespace was treated
as equivalent to an approved marker. Fixtures use legal XML character references,
not NUL inside XML. Encoding tests separately cover NUL/soft hyphen in utf-8/
utf-16le names. Aggregate assertion failed as expected,not fixture setup.

Classifier now uses string-type-guarded ordinal equality/membership for XML
element/prefix/namespace/attribute identities,envelope/version/channel,type and
progress-state markers,and encoding selection. No normalization,case folding or
blanket ban on Unicode display content. The remaining Nil comparison uses an
internal constant field-name list,not untrusted text. Canonical numeric/reference
regexes,strict decoders,DTD/resolver prohibition,XML bounds and original byte
hashes unchanged. Public typed EncodingName binding is unchanged;this is not a
new certification of arbitrary-object coercion/callback or all hostile inputs.

Tests exercise actual public classifier,no mocks.12 XML refusal cases per
encoding,4 valid Unicode display/state cases and6 encoding-name cases give34
new checks plus aggregate. Independent SHA computation checks hashes/lengths
even for refused bytes;positive display fixtures include composed/decomposed
accents,soft hyphen and CJK text in Completed/Processing states. Caller bytes
also checked unchanged. Only protocol marker identity is tightened.

## Fresh finite verification —4/4 consumed

|Call|Host/phase|Exit|Elapsed ms|PASS|FAIL|Stderr artifact bytes|
|---|---|---|---|---|---|---|
|1|PS7 RED|1|1162|13|22|369|
|2|PS5 RED|1|1041|13|22|559|
|3|PS7 GREEN|0|1013|87|0|0|
|4|PS5 GREEN|0|998|87|0|0|

Expected RED assertion ExactClassifierMarkersRequired. GREEN includes previous
52 classifier checks and35 additions;each complete87-check suite exit0/empty
stderr.22 assertions observed RED->GREEN on each host;12 new checks already
passedRED,not claimed failing first. No source/test edits between GREEN hosts
or after verification. Launcher selects pwsh.exe and inbox Desktop PowerShell
with child-local inbox module path;this suite prints no host-version line,so
no fresh full version attestation is claimed. Prior observed version records
remain historical. Parent checks saved child exit/results,not launcher exit alone.

Source SHA256:`F323C9C46190DE06F25BFA55A2B7983707B1FF5877548F9703870F9BB771A0BF`.
Tests:`817B2DDB62BB0F0C4AB9681597E4B0ADB4B4B2825D648A86356FE9A7E6EC0FBB`.
Four-entry index:`0B8A64E0514727A59252B3B90FBAEA31E59DC6B85A5F9652ED7CA34FC8441F03`.
Ignored raw receipts/index under .superpowers/sdd/2026-10-17-classifier-exact-markers
retained,not committed or required by source/tests. Decoded/re-encoded stdout/
stderr artifacts are synthetic test output,not live raw collection evidence.

## Review focus and disposition — bounded checkpoint accepted

One fresh scoped readonly review completed:0Critical/0Important/0Minor findings;
suitable for preservation as bounded classifier-correction checkpoint. It
confirmed exact marker admission,strict namespace/prefix/attribute counts,TNRef
fallback identity/reference checks,Unicode display preservation,unchanged
numeric/XML limits and prior52 assertions. No actionable defect or repair pass.
No reviewer execution,source loading,edits,network,Git mutation or subdelegation.
Run timings/counts/stderr/hashes are parent evidence,not independent execution
or receipt attestation. No second review or extra test/fix reserve used.

Declined judgments and parent rulings:

- Independent execution/receipt or hash attestation:passive review only. Parent
  read saved outputs/child records and hashes. Cost:no independent rerun proof.
- Package/store/contract correctness/fresh integration:outside two-file change,
  no dependent suite rerun. Cost:prior110-check record cannot certify changed
  dependency;future integrated verification needs separate finite authority.
- Arbitrary-object coercion/callback/internal-helper misuse:existing typed
  public API retained,not a hostile-caller redesign. Cost:no stronger caller
  boundary;standalone package GetType issue stays open.
- Universal CLIXML/numeric redesign:restricted subset and established grammar
  intentionally retained. Cost:unsupported but legitimate CLIXML may refuse;
  no broader compatibility promise or freshly tested culture variation.
- Exhaustive adversarial/resource certification:finite fixtures/static review
  insufficient. Cost:no sandbox/resource-exhaustion resilience guarantee.
- Live provenance/clocks/concurrent tampering,durability,watchdog/OS cleanup,
  security/vendor clearance,Stage B/I2/I3/full C12:no implementation/evidence or
  authority here. Cost:those gates remain blocked,not treated as passed.

No confirmed minor deferred. No merge/production readiness approval. Preserve
existing dev branch and raw audit workspace;no PR/main merge/worktree deletion.

Ruling:PowerShell-native bookkeeping and finite classifier suite instead of
Bash helpers/extra baseline/task-done/full-project runs;retain audit workspace.
Approved cap/preservation prevails. Cost:no fresh package/store/contract/full-
project regression claim. Prior110-check contract result is historical and NOT
rerun against this changed dependency. This task closes classifier findings
from the readonly audit only;standalone package exact-label/GetType concerns
remain unresolved and require separate design/tests/authority.

All earlier budgets/frozen slots/failures preserved,not transferred. Origin
authenticity,actual clocks,concurrent hostile caller/process tampering,resource
exhaustion,live collector/watchdog,OS stop/cleanup,durability,security/vendor
clearance,Stage B,I2/I3 and full C12 acceptance remain unproved/unauthorized.
Classification and exact matching never grant Acceptance or RuntimeApproved.
