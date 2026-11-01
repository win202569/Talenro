# Owner scope repair — stopped during auxiliary RED

Status: NOT_ACCEPTED. Task1 incomplete; Task2/3 not started.

User approved [design](../superpowers/specs/2026-11-01-offline-owner-scope-repair-design.md),
[plan](../superpowers/plans/2026-11-01-offline-owner-scope-repair.md) and its finite
increment with reply “是”. Baseline f8514dd7a035abd413c2928daa6f882f9d1c85cd.
Before authoring: branch/origin/clean isolation verified,2215present pins and
all93snapshots across three epochs matched; all new attempt paths absent.

## Actual execution

Auxiliary owner-scope.tests.ps1 committed BEFORE computation at
385f338b64f714de91c8b1679cc1e0c4335efa09; SHA256
BDA424DCC24805F6FFDDEB30DDB67F502CDE7FD34B4B32E1E760D156507C4DCE.
Static parser reported zero errors. Approval/input/RED intent created append-only.
The sole RED computation created one actual private module and ran the actual
initializer, then attempted function resolution in a separate module invocation.

Actual tool exit2, Unexpected=true. Error: RuntimeException,
`Owner scope regression: only intended absent actual profile`, assertion at
auxiliary line94 (ODS-Need throws at line4). It did NOT return the specified
1FAIL/8NotRun/exit1 receipt. No owner-scope-red.final.json exists. The assertion
threw BEFORE saving the caught underlying ErrorRecord, so its exact Type,
ErrorId, message, position and stack cannot be recovered from this saved output.
Do not infer or synthesize those missing fields from the historical writer error.
Get-Command error-signature mismatch is a hypothesis, not a measured root cause.

I wrote this auxiliary and its signature check/late error preservation;
this is a verification-script failure, not evidence of a production repair,
platform refusal or changed runtime sources. No frozen script edit, replay,
GREEN, new freeze, host, compile, profile body invocation, target or review after
the unexpected result. All existing runtime sources remain at the old epoch.
Complete actual tool input/output and separate unexpected JSON retained in P.

New writer-scope.stop.json SHA256
D82E7A269C8982EB6AA7BDECCC779D8A3E674CC3F01EBE6484A4145B35CF3236.
All three historical stops and old outputs remain in place. Scope regression
tool record SHA256
8E6CD1BFF8B2C6F2B9B6467F36D4EADD9F9790C8CFCFF80C69EE891D6007EAA2.

## Actual budget and recovery

- Scope computations1of2 consumed; remaining1was GREEN, now frozen, not spare RED.
- This turn0capture hosts/0compiles/0targets/0reviews/0source snapshots.
- Overall1historical writer host/1compile consumed; runtime remainder5hosts/
  7targets/1compile/conditional1review frozen after approved increment.
- Source snapshots93of155; new scope manifest/31snapshots not created.

Read new portable-owner-scope-repair.json for all actual P/Q originals, including
the new approval/input/intent/unexpected/tool record/stop/ledger. Old portable
archives remain unchanged. Auxiliary source and this report are ordinary Git
files. Never replay commands automatically after cloning or treat absence of
runtime source changes as a reason to refund the computation.

Next recommendation: readonly review of the auxiliary's error preservation and
signature boundary before proposing a corrected, separately versioned test and
any replacement RED allowance. Do not request another writer host/compile merely
to diagnose the auxiliary. Existing unused runtime allowance stays frozen.

Stopped portable archive:278actual files/75directories/10zero-byte files; every
decoded length/hash/raw byte sequence matched actual originals. SHA256
1EE8B6A21A1D0A6B56C465E5F1454CDD1B59602FF1761C6050FE3D4E8740474E.
Post-stop readonly verification:2215pins,31runtime source/snapshot byte pairs,
93snapshots,12historical artifact identities and old portable hash unchanged;
new runtime stages absent. Frozen auxiliary hash also unchanged.
Pre-commit raw-byte gate found CRLF/LF normalization of the newly generated
portable container. No commit occurred at that point. Container formatting
only was normalized to LF, then all278embedded actual files were byte-verified
again; original P/Q contents, older archives and all tests stayed unchanged.
