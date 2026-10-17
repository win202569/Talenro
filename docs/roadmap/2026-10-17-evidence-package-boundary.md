# Evidence package exact/passive input boundary — bounded correction

## Approved scope

User approved in-chat bounded design after64de23b048f9cdf6863e7059ec33a1760ae50d17:
package/tests only plus preservation docs;exact labels/fields,non-dispatching
outer/nested type validation,executable-property and bound-shadow rejection,
unchanged raw bytes/Unicode origin/schema/format. Four NEW synthetic hosts
PS7RED/PS5RED/PS7GREEN/PS5GREEN,120s each,no retry,then one readonly review.
No classifier/store/simulation/contract source edits,reverse import,format
migration,live workload/security change,Stage B or main merge authority.

## Reproduced failures and correction

Both REDs produced the same18 failed new assertions:4 non-exact label mutations,
2 constructor encoding variants,4 caller GetType dispatch checks,2 executable
scalar-property checks,2 raw-byte bound shadows,an origin bound shadow,2 stream
snapshot bound shadows and a raw-byte Length getter. Each ended at expected
EvidenceBoundaryRequired aggregate,not fixture error. Six other new assertions
already passed before correction;no claim that every new assertion failedRED.

New local helpers invoke Object.GetType directly on PSObject.BaseObject rather
than caller GetType methods. Plain outer/nested NoteProperty shapes are checked
before field access. Scalar/byte validation inspects descriptors without reading
ScriptProperty/CodeProperty values,and refuses non-CLR Length/Rank descriptors
before intrinsic bounds are read. Valid unused GetType ScriptMethods can remain
without execution;wrong containers cannot spoof their CLR type through them.
Primitive integers/flags also get passive checks. Constructor/stream helper
check actual byte arrays/bounds before Base64/hash processing.

Policy,encoding,classification,field names,canonical snapshots and recomputed
labels use ordinal comparisons. Unicode origin declarations remain unchanged
and unverified,not normalized or authenticated. Existing diagnostic text for
canonical Base64 and invalid origin remains compatible with established tests.
No classifier private-helper dependency or reverse simulator import introduced.
Public constructor byte[] parameter binding remains unchanged;arbitrary coercion
before function entry and direct misuse of internal helpers are NOT newly
certified hostile-caller boundaries. Concurrent/process-wide tampering remains
excluded;descriptor checks are not a sandbox.

## Fresh verification —4/4 consumed

|Call|Host/phase|Exit|Elapsed ms|PASS|FAIL|Stderr artifact bytes|
|---|---|---|---|---|---|---|
|1|PS7 RED|1|1309|7|18|361|
|2|PS5 RED|1|979|7|18|545|
|3|PS7 GREEN|0|1667|74|0|0|
|4|PS5 GREEN|0|1305|74|0|0|

Observed7.6.5Core x64 and5.1.26100.9711Desktop x64 in the counted hosts.
Both complete suites include previous49 checks plus24 boundary checks and one
aggregate. All74 passed with exit0/empty stderr. Callback fixtures count actual
ETS method/getter execution;final required counts0. Unicode origin and literal
abc/raw00FF0D0A snapshots remain unchanged;normal JSON object roundtrip,failed
child/refused output retention,canonical padbits,1MiB limit and false authority
flags still pass. No source/test edit between GREEN hosts or after verification.

Source SHA256:`857543A9DDA9B9C849B2507C6EE419C1F73AA0D0C8599CD9D88738841D021198`.
Tests:`114F2F66A4D96EFC30E4975D1467CFDA47C349397221E00AF17C21BF1A649ED3`.
Four-entry index:`3ED40888CF9870184E21564F0E1673062BDDDC1EDA8B16EAC7E58CA58A0B7DB0`.
Ignored receipts/index under .superpowers/sdd/2026-10-17-evidence-package-boundary
retained,not required by portable tracked source/tests. Redirected text receipts
are synthetic host output,not actual raw collection/provenance evidence.

## Review disposition — standalone checkpoint accepted

One independent scoped readonly review completed:0Critical/0Important/0Minor.
Verdict acceptable bounded standalone-package development checkpoint,not fresh
integration/merge readiness. Reviewer confirmed reflected base-type query,
descriptor-before-bound/primitive validation,exact nested shapes and ordinal
matching,byte/Unicode/schema preservation,and no cycles/private classifier
coupling. Existing49 checks unchanged. Branch/base and source/test hashes
independently matched;classifier/store/simulation/contract source unchanged.
No reviewer execution,source loading,edits,network,Git mutation or subdelegation.
Recorded run results remain parent evidence,not independent re-execution.

CodeProperty,Rank and decorated-primitive branches were source-traced,not each
separately demonstrated by a new fixture. Do not turn generic guarded code into
an exhaustive dynamic coverage claim. No confirmed minor deferred or second
review/fix pass. All4 calls consumed with no source/test edits afterGREEN.

Declined judgments and parent rulings:

- Arbitrary typed byte[] pre-entry conversion:unchanged binding behavior,not
  this validation-stage correction. Cost:no hostile pre-binding conversion claim.
- Direct hostile helper misuse:only public constructor/verifier supported.
  Cost:internal functions are not separately hardened public/security interfaces.
- Concurrent/process-wide tampering/sandbox:excluded from synchronous contract.
  Cost:descriptor validation does not supply process isolation or race resilience.
- Origin/command/clock/capture authenticity or coordinated data/hash changes:
  no external trust anchor. Cost:consistent fabricated evidence can still verify;
  authority flags remainfalse.
- Raw JSON/duplicate keys,durable serialization/atomicity/fsync/full capture
  persistence:in-memory object scope. Cost:no disk durability/complete archive
  guarantee or format migration;store still serializes Package only.
- Fresh classifier/store/simulation/contract/full-project certification:those
  suites not run;shared helpers changed. Cost:historical consumer results are
  not current integrated acceptance;new finite regression authority required.
- Live collection/watchdog/cleanup,security/vendor,Stage B/I2/I3/full C12,
  production/main merge:neither authorized nor evidenced. Cost:all gates remain
  blocked,not counted as passed from a green package suite.

These limits are accepted as limits,not waived guarantees. Preserve existing
development branch/audit artifacts;no PR/main merge or worktree deletion.

Ruling:finite package suites/PowerShell-native bookkeeping rather than extra
Bash/baseline/task-done/full-project runs;retain ignored audit artifacts. Approved
cap/preservation prevails. Cost:no fresh classifier/store/simulation/contract or
full-project regression certificate. Prior54-store/110-contract and other
dependent results are historical,not rerun against this changed package helper.

Next safe proposal after review:separately approved dependent pure-chain
regression,then reassess store writer inheritance/local-file integration. Do not
rewrite store format or claim fixed upstream code proves unexecuted store paths.
No new allowance follows from this report. All old/frozen budgets/failures
preserved. No live capture/watchdog/OS cleanup,command/clock provenance,durable
complete capture-record schema,security/vendor clearance,Stage B,I2/I3/full C12
acceptance or main merge. OriginVerified/Acceptance/RuntimeApproved remainfalse.
