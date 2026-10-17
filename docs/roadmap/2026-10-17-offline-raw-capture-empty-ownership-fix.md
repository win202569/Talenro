# Offline raw capture — bounded empty-array ownership repair

Subsequent M1 status: a separately approved tests-only follow-up adds independent
complete event expectations and passes four new hosts. See the
[M1 evidence](2026-10-17-offline-raw-capture-event-oracle.md). Deferred-M1 wording
below describes this earlier I1 checkpoint; its receipts/budget are unchanged.

## Authority and scope

After checkpoint f8657855722927630d17a0b23c04f8e8767dc66a, the user explicitly
answered yes to the [bounded I1 proposal](2026-10-17-offline-raw-capture-readonly-review.md).
Only the two array clones, one cross-call regression and five NEW sequential
synthetic children were approved, each capped at120000ms. Any unexpected result
stops without retry. No new reviewer, dependency edits or M1 implementation.
The original ten slots, slot5 false verdict/exception and sole review remain
unchanged and spent. This is not an expansion of old archive acceptance.

## Change and regression

`offline-raw-capture.ps1` now explicitly clones both List<object>.ToArray()
snapshots: returned Actions and outbound Events. No interface, bounds, stop
mapping or dependency changes. Empty arrays no longer expose the shared runtime
cache through the returned Actions. Uniform cloning follows the existing public
contract pattern without adding an empty-only branch.

`offline-raw-capture.tests.ps1` adds only RawCaptureEmptyOwnership and its
empty-ownership group. Original28 case bodies are unchanged. The new case adds
Length/Rank ETS properties to returned empty Actions, observes whether the
original Record still verifies and a fresh valid nonempty capture still creates
a record, then removes only its own decorations in finally even on RED. It
checks the fresh byte result and subsequent independent empty snapshot too.
The public-behavior assertion is primary, not merely ReferenceEquals.
Full inventory is the old28 with EmptyOwnership immediately before Complete;
targeted inventory is ApiRequired/EmptyOwnership/Complete.

## Separate finite evidence

Status: **I1 CLOSED BY PARENT RED/GREEN EVIDENCE; M1 DEFERRED; NO NEW
INDEPENDENT REVIEW OR BROADER ACCEPTANCE.** All five new slots consumed and
matched their precise expectations. No extra hosts, reruns or deadline extension.

|New slot|Engine/group|Exit|Elapsed ms|Observed|Verdict|
|---|---|---:|---:|---|---|
|1|PS7 targeted, unchanged source|1|1086|API PASS, exact RawCaptureEmptyOwnership assertion|expected RED|
|2|fixed PS7 targeted|0|1246|exact3 ordered PASS|GREEN|
|3|fixed PS5.1 targeted|0|1218|exact3 ordered PASS|GREEN|
|4|fixed PS7 full|0|48343|exact29 ordered PASS|GREEN|
|5|fixed PS5.1 full|0|109628|exact29 ordered PASS|GREEN|

All five stderr files are empty, no timeouts, all ExpectedResult=true. Final
readonly reconciliation verified every new receipt hash and all ten original
receipt pairs; original index/stopped-index hashes remain unchanged. Static
PowerShell parser checks for source/test/controller report zero errors. No
additional test import/execution was used for those checks.

Child contexts: DESKTOP-H4SVCLU\Lenovo, admin=False, x64. PS7.6.5 Core and
PS5.1.26100.9711 Desktop. Engines were resolved at fixed existing paths; no
installation, elevation, token switch, security/trust change or progress/error
suppression. Hosts run -NoProfile -NonInteractive -File; PS5.1 uses the inbox
module path. Preconsumed slots, separate redirected streams, owned process
handle, immutable prior-receipt hashes and exact ordered arrays guard execution.
Expected-label arrays are explicitly normalized, including the single RED label.
The controller checks test/source/dependency pins before and after every child.
These are test-harness stdout/stderr receipts, not a real collector's raw bytes.

## Frozen hashes

All hashes here identify local execution bytes, not Git's line-ending-normalized
blob format. Source was unchanged for RED1; fixed source frozen before GREEN2;
amended test frozen before RED1 and unchanged throughout.

|Item|SHA256|
|---|---|
|Source before repair|4C0E6C10DAB31C63E550416260B0672EB0D721A8210B31CCABEA05648DCFB673|
|Fixed source|FB48642D4A4F7CA8337DEDD61C4202C3C1C182168BEF23E574B22313386291C7|
|Original test|7010CBCF535A431D62AEA7C3392971E96F46769801DF2075651CA87B3B8D88D7|
|Amended test|47808D6CF2E591FFF39325F238E505B1111E5A45DBCDF924279B52C7DBAEB978|
|New controller|63BF7C2CAA89DFB9C54588F2A947673606155B1B0A9645F906ADC3B10EDB4E37|
|New final five-slot index|1A660022166C51BF42DDE67E8A91CB7B44D528DCC40A341515B76B85866B9B50|
|Original ten-slot index, retained|8EB29C6E95ADBF0D1491833303E970FF1A55C6619AC932514F065B6BE084F4D6|
|Original stopped-five index, retained|82B3593D3526EFCA688D0D4ABB0DD404E22D073DC320611FC4C96AC4CE129D39|

All12 [dependency pins](2026-10-17-synthetic-capture-dependency-regression.md)
remain unchanged. Original receipts stay in their original ignored packet;
new receipts/controller/index are separately retained under
`.superpowers/sdd/2026-10-17-offline-raw-capture-empty-ownership-fix/`.
Do not commit ignored raw/session data or depend on it for fresh-computer tests.

|New slot stdout|SHA256|
|---|---|
|1|E8D66A63EE072F133AED2F516D20B333B82305C3FD5DC2FBD7CE7BF39BA5FE2D|
|2|C2C86746075937671B9E999C958E8673CFE38A3731B6172C07ED04B95BD33E5A|
|3|C9A1B7EEA7D468117E35FAC268FA9083638FAC9147B953F2AE24418EBE8E8DE9|
|4|9CD6F23BC45ED4580CFE05ADE00F706A8E41D7012304A8F5C3DF02E863F145B4|
|5|78055A89696A3C5BD84CB6289FDE8B775744520DFA50A140359325F9E4D55219|

Every new stderr SHA256 is the empty-file hash
E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855.

## Limits and recovery

I1 dynamic reproduction is established on PS7 by the precise new RED followed
by the same frozen regression's GREEN after only the two clones. The old
PS5.1 manifestation was not tested and is not claimed; post-fix compatibility
GREEN is a different fact. No additional independent review was requested or
performed. M1 (full independent mapped-event expectations) remains disclosed
and deferred; the old DerivedConsistency case is not strengthened or claimed
to provide that oracle. All other review limits remain in force.

Do not infer real pipe/concurrency/read-fault/watchdog/timeout cleanup proof,
durability, execution attestation, provenance, security/Stage B/I2/I3/Task10/11,
full C12/release/main acceptance. The nonempty CLIXML stderr branch and timeout
cleanup path are not dynamically exercised by successful empty-stderr runs.
RuntimeApproved/OriginVerified/Acceptance remain false; CleanupStatus unknown.

Recover from the existing development branch `codex/c12-b01-task9-coordinator`,
not main. Source and tests are ordinary repository files and need no old session
cache. Read this record, current status, original review/execution and recovery
runbook before further work. Spent budgets are not permission to replay runs.
Any next M1 improvement or broader work needs its own scoped decision.
