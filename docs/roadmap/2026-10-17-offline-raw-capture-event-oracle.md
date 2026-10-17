# Offline raw capture — M1 independent event oracle

## Approval and boundary

After checkpoint81afee106650c0a9528acbbc80272d093b2b1119, the user approved
the short M1 design: tests only, independent complete event expectations,
deliberately incorrect counterexamples and at most four NEW sequential test
hosts,120000ms each. Order: PS7 targeted, PS5.1 targeted, PS7 full, PS5.1 full.
Unexpected results stop without repair/retry/deadline extension. The product
and12 dependencies must remain unchanged. Commit/push and recovery documentation
use the standing preservation authority. No main merge or real collection.

## Test change

Only offline-raw-capture.tests.ps1 changes executable content. DerivedConsistency
now compares eight complete ordered stored-event arrays to literal fixtures:

|Fixture|Hand-derived events|Expected disposition|
|---|---|---|
|complete|stdout byte01, stderr byte02, stdout EOF, stderr EOF, exit0 at times1..5|complete; no failure index|
|incomplete|stdout byte01 at1, tick2, exit7 at3; unread tails retained|incomplete; no failure index|
|deadline-read|stdout byte01 at1, tick120000 instead of stderr read; later exit omitted|timed-out; index1|
|backward-read|stdout byte01 at2, tick1 instead of backward stderr read|refused/time-went-backward; index1|
|after-eof-read|stdout EOF1, empty stdout chunk2; later exit omitted|refused/chunk-after-eof; index1|
|duplicate-exit|exit0 at1, exit7 at2; later tick omitted|refused/duplicate-exit; index1|
|trailing-read|stdout EOF1, stderr EOF2, exit0 at3, tick4 instead of trailing read|refused/trailing-after-complete; index3|
|empty|empty array, not null|incomplete; no failure index|

Expected values are not derived with the coordinator, constructor, verifier or
their helpers. Expected chunk Base64/byte lengths/hashes are literals; one-byte
SHA256 values were separately checked with .NET SHA256 over literal01/02 bytes,
not product output. The recursive test-only comparator checks exact property
sets/names, array order/length, ordinal strings and integer values (Int32/Int64
accepted without string coercion). Data includes all Base64/ByteLength/SHA256
fields. Positive event comparison precedes the existing public record verifier;
status/reason/intent and both failure indices are checked independently too.

New EventOracleRejectsAlterations starts from genuine verified captures and
copies their returned arrays/events/Data into test-owned snapshots. It rejects
extra/missing/null events, reordered events, wrong kind/time, missing/extra
event fields, channel case changes, wrong exit code, and altered Base64/length/
hash or missing/extra Data fields. Original captures are rechecked unchanged.
These are negative inputs to the test comparator, NOT product mutations or
an observed production defect. No external failing-host RED was proposed or
run; the expected false comparisons execute inside each passing negative case.

The new event-oracle group selects ApiRequired,DerivedConsistency,
EventOracleRejectsAlterations,Complete (four labels). Full suite keeps the old29
labels/order, inserting EventOracleRejectsAlterations after DerivedConsistency,
for30 labels. All other old case bodies, including EmptyOwnership, are unchanged.

## Finite verification

Status: **M1 ADDRESSED BY PARENT'S BOUNDED TEST EVIDENCE; FOUR OF FOUR NEW
SLOTS COMPLETE.** No additional independent review or broader acceptance.

|New slot|Engine/group|Exit|Elapsed ms|Ordered labels|
|---|---|---:|---:|---:|
|1|PS7 event-oracle|0|2833|4|
|2|PS5.1 event-oracle|0|2513|4|
|3|PS7 all|0|15402|30|
|4|PS5.1 all|0|115437|30|

All four ExpectedResult=true, stderr empty, no timeout/retry or deadline
extension. PS5.1 full115437ms is within120000ms, not a worst-case performance
guarantee. Final readonly reconciliation verified every new receipt hash,
unchanged product/test pins and all12 dependencies, plus original10 and I1-fix5
indices/receipt pairs. Static parser checks for amended test/controller and Git
whitespace check passed; those checks did not execute extra test hosts.

Hosts are PS7.6.5 Core/PS5.1.26100.9711 Desktop, x64, inherited normal token
DESKTOP-H4SVCLU\Lenovo admin=False. No install/elevation/security/trust change.
The new separate controller preconsumes slots and checks frozen test/product/
dependency pins before and after each child. It retains separate stdout/stderr,
exact ordered label inventories, exit/elapsed/timeout and actual context. Any
unexpected result freezes later slots. Same strict empty-or-Completed-CLIXML
stderr policy; DTD/external resolution disabled, no object deserialization.
These are harness receipts, not collector raw-byte evidence.

|Pinned item|SHA256|
|---|---|
|Unchanged product|FB48642D4A4F7CA8337DEDD61C4202C3C1C182168BEF23E574B22313386291C7|
|Previous test|47808D6CF2E591FFF39325F238E505B1111E5A45DBCDF924279B52C7DBAEB978|
|Amended frozen test|2D44F9C9EB68450866A92D43384E25C5DF4133F60150FB1DC3BE20EDBE1D914B|
|New controller|58FFBAAC04CCECE0FC17B129A8968011A7F5A176158DBCD2261548ADB5A44007|
|Final four-slot index|CD2B3BB13997FEE9C6AA2238C6A22F2ABF3F69996A73F24A6FBF4BA13F4CF901|

All12 old [dependency pins](2026-10-17-synthetic-capture-dependency-regression.md)
remain protected. Local raw/controller/index are retained separately at
`.superpowers/sdd/2026-10-17-offline-raw-capture-event-oracle/` and not committed.
Hashes describe local execution bytes, not line-ending-normalized Git blobs.

|Slot stdout|SHA256|
|---|---|
|1|37452487410FDB43010FEB2FB939C75486899644662CF33458067BF758A1550B|
|2|B3B68F2BB1CFF7E345A3C310953C55CBE14D0F902FF23C5871A8A5D0A177D997|
|3|E1E05813BC4EF49AA267DF73CFB2977E1868F6E95C6A05FB0B410F7C21AF1EA2|
|4|6EEE0420B23D90A3DA59627E19C0E74FD2B8FDE44291150FF112D19569471BCA|

Each stderr has empty-file SHA256
E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855.

## Disposition and recovery limits

This task addresses the original review's M1 promise, not every possible mapping,
malformed-input combination or hostile/concurrent runtime. No additional
independent review, product-mutant run, dependency-suite rerun or repository-wide
test run is part of the four-host budget. Never present named-suite results as
whole-project verification. New4/4 are spent. Original10/reviewer1 and I1-fix5 remain spent and
their raw history/slot5 false verdict/admission exception remain unchanged.

No change to real pipes/concurrency/read faults/watchdog/timeout cleanup proof,
durability, execution attestation/provenance, security/Stage B/I2/I3/Task10/11,
full C12/main/release acceptance. RuntimeApproved/OriginVerified/Acceptance stay
false and CleanupStatus stays unknown. Existing archive acceptance is separate.
Nonempty CLIXML stderr and controller timeout cleanup are unexercised limits.

Fresh-computer recovery uses codex/c12-b01-task9-coordinator, not main, and the
portable source/tests/status/runbook, not ignored session caches. A later
limited-acceptance decision is separate; this test change does not grant it.
