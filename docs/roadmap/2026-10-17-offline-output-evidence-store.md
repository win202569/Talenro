# Offline evidence persistence — scoped verification and preserved history

## Authority and outcome

User approved the bounded persistence design and at most four fresh synthetic/local-file test hosts after commit6ce0bbd6. No live collector, ordinary ledger identity, Stage B, original suite replay or wider acceptance was authorized. Existing package4/4,classifier8/8,ledger70/70 budgets remain separate/frozen.

Latest approved elevated continuation:PS7 and PS5 each passed50 checks,exit0,
empty stderr,with unchanged source/tests. This is scoped synthetic/local-file
verification,not origin authentication,Stage B or full-project acceptance.
Two review minors remain deferred. Earlier failed and partial runs below are
preserved history,not overwritten or converted into passes.

## Implemented behavior (scoped offline verification only)

`output-evidence-store.ps1` exposes Write-/Read-OfflineOutputEvidenceStore. Fixed worktree-local `.superpowers/diagnostics/offline-output-evidence-store`, fresh direct child only. Caller prepares the root. Writer checks an existing package, constructs bounded primitive fields, creates fresh files, closes/reopens payload, checks bytes, and publishes completion marker last. Reader holds single-link read bindings, checks marker length/hash, parses strict bounded fields/exact EOF, then rebuilds and verifies the package. A refused classification can still be retained as internally consistent evidence, never authenticity or successful acceptance.

Format: `TLROEV01`,little-endian Int32 exit/encoding(1=utf8,2=utf16le),Int32 strictUTF8 origin length/bytes,Int32 stdout length/raw bytes,Int32 stderr length/raw bytes. Marker `TLROEC01`,Int64 payload length,raw SHA25632;exact48 bytes. Stream limit1MiB each,origin1..1024 chars/4096 bytes,payload<=3MiB. No generic text deserialization. File failures preserve partial state and best-effort `stopped`; no overwrite,delete or retry.

Dependency: unchanged pure package/classifier and isolation.ps1 file metadata helpers; import defines other isolation functions but invokes none of their toolchain/patch operations. No runtime process helper is imported. File metadata P/Invoke is not a live process workload. No hostile-admin/concurrent namespace or power-loss durability guarantee.

## Fresh execution record

Ignored runner acquired a separate Store switch/index,budget4,only this test suite,120s sequential hosts. No earlier index rewritten.

| Call | Host / phase | Exit | Elapsed ms | PASS lines | Result |
|---|---|---|---|---|---|
|1|PS7 RED|1|1582|0|Expected OfflineEvidenceStoreRequired assertion; missing API|
|2|PS5 RED|1|2084|0|Same expected missing API assertion|
|3|PS7 GREEN attempted|1|1402|1|Bounded integer required before store entry|
|4|PS5 GREEN|—|—|—|Not started; remaining slot frozen at stop|

Parent runner itself prints child errors and can exit0; recorded child exit is authoritative. The surrounding GREEN command checked the saved exit and threw on call3, so call4 did not start. No retry/cleanup or fixture/source correction followed. Root was created; readonly listing found it empty. Tests and source passed readonly PowerShell parser inspection, which is not execution or PS5 compatibility proof.

Index `store-invocations.json` (three completed entries) SHA256:
`83839B575F5A80490ECBF1D43696D78AAA2B710B2243268BD63D4BE6E151D033`.
Saved stdout/stderr and index remain ignored under `.superpowers/sdd/2026-10-17-ordinary-ledger-native-probe-preparation`; they are local audit artifacts, not prerequisites for running the committed fixture suite on a fresh computer. No real raw payload is committed.

## Readonly diagnosis and next step

The parser AST for the first `New-OfflineOutputEvidencePackage` invocation in `New-StorePackage` shows bare `-ExitCode -1` as **StringConstantExpressionAst / System.String**. The unchanged package validator only accepts Int32/Int64 and throws `Bounded integer required`. This explains call3 before directory creation/store entry; it does not prove the unexecuted store implementation is correct. The earlier package tests used an integer loop variable for negative exit values.

Proposed smallest correction: explicitly numeric `-ExitCode ([int]-1)` in the fixture. No production validator relaxation. Do not apply/retry under an assumption that the stopped allowance renews. Seek approval for fixture correction and two new sequential bounded verification hosts (PS7 thenPS5,120s each,no retry), while retaining all original three calls and freezing unused call4. Any new failure must remain visible; source/tests after correction need their own results.

## Independent readonly review

One fresh scoped reviewer reported0Critical,1Important,2Minor;no additional confirmed production correctness defect. Important is the same fixture integer-string failure above and remains unresolved. This is not a functioning-store or merge verdict.

Deferred minors (no fixes/runs this turn): empty-origin mutation retains its original bytes,so generic refusal may test the following malformed length instead of isolating empty-origin validation;UTF16LE persistence fixtures contain only empty streams,so nonempty UTF16LE reconstruction/classification is untested. Future correction needs targeted assertions,not broader success claims.

Ruling on declined-to-judge areas: hostile namespace/hardlink races,concurrent caller mutation,ACL ownership/authenticity,actual mid-write OS faults,crash/fsync durability,real collection/runtime authority,wider repository correctness and acceptance remain unsupported/unverified. File bindings protect opened files,not the whole namespace;the root is a dedicated local scope,not ACL ownership attestation. These exclusions match the approved scope,not evidence that the stronger properties hold. Cost if wrongly relied upon: false trust,readable completed artifacts after failed stop publication,or unsafe reuse as runtime evidence. Do not rely on them.

Reviewer noted a complete marker surviving later OS failure plus failed stop-record creation is not guaranteed unreadable;best-effort stopped marking is intentionally weaker than guaranteed crash-failure atomicity. No source/test changes followed review.

Reviewed source SHA256:`7710A1149B42CFD0C2505988CEE6748F399C3FB5E8C34D795B35FBCB8558BB81`.
Reviewed tests SHA256:`C6B399D5B517226A15A9F45AB9F13391E269755176E921D27DC96A09AAF6B8B9`.
Saved decoded stderr artifact sizes for calls1/2/3:363/551/337 bytes (runner text re-encoding,not raw child byte capture).

Formal Stage B/safety disposition and I2/I3 full acceptance remain OPEN. This offline persistence extension cannot resolve missing original raw bytes or historical security disposition.

## Approved numeric-fixture continuation — partial, blocked on hardlink creation

After f2c562a4,the user approved numeric fixture correction plus at most2 new
sequential hosts,PS7 thenPS5,120s each,no retry. Applied only bare `-1` to
`([int]-1)` at New-StorePackage;production source/validators and two deferred
minor coverage gaps remain unchanged. No repeat independent review for this
single known correction;prior source review/limitations still apply.

Separate ignored `store-repair-invocations.json` contains1 completed call:
PS7 GREEN attempted,exit1,7373ms,45 PASS lines,decoded stderr artifact338 bytes.
PS5 was not launched after first-host failure;remaining repair slot frozen.
Original Store index still exactly3 calls and unchanged SHA83839B575F5A80490ECBF1D43696D78AAA2B710B2243268BD63D4BE6E151D033;
the original unused fourth remains frozen too. No retries or cleanup.

New index SHA256:`514F1512188C77376657E4FFB518ACA22B32B9C3BBD23E7C3D5A0B072890725F`.
Unchanged source SHA256:`7710A1149B42CFD0C2505988CEE6748F399C3FB5E8C34D795B35FBCB8558BB81`.
Corrected tests SHA256:`E54BE44B09391A02068CF2E4B203B1908C019766217AA4CDF368DB40A1821072`.

Executed PS7 checks reached full binary roundtrip (including nonzero exit/raw
binary),independent literal layout,marker layout,non-authentication flags,
caller-object independence,no overwrite,empty UTF8/UTF16,UTF8 progress,
missing/partial/stopped/hash mismatch refusal,valid-hash malformed fields,
both1MiB streams/origin1024 bound,bad package/surrogate rejection,and unsafe
path read/write refusal. This verifies those45 assertions on PS7 only,not the
whole suite. Prior numeric fixture blocker no longer reproduces.

Failure at tests line94: `New-Item -ItemType HardLink` reported access denied
while creating the first hardlink fixture. Neither hardlink rejection assertion,
junction/reparse checks,nor final exclusive-writer check ran. No implementation
bug is established by failed fixture setup;the access-denial cause is not yet
proven to be sandbox versus filesystem/ACL policy. Do not silently skip it.

Readonly inspection retained `fixture-678f75e7f1634263b9049353073823ae-30`,with
evidence.bin49 bytes and complete.bin48 bytes,ordinary Archive attributes.
Directory owner BUILTIN\\Administrators,inherited ACL (AreAccessRulesProtected
false);these observations alone do not establish why link creation was denied.
All fixtures remain ignored/local;no real raw output was collected/committed.

Next proposal: obtain new authority for at most2 fresh bounded local-fixture
hosts (PS7 thenPS5),executed with scoped sandbox elevation to permit hardlink
fixture creation. Preserve checks and all failure artifacts;no ACL/security
changes,no outside-project links,no automatic retries. Elevation may still
fail and is not a guarantee of a green run. Full-project/Stage B authority and
old budgets remain unchanged. No further source/test fixes in this continuation.

## Approved elevated continuation — 50 checks per host passed

After fa4f2094,the user approved at most2 new sandbox-elevated local-fixture
hosts,PS7 thenPS5,120s each,no retry or security-setting/ACL change. An independent
StoreElevated index/budget2 was added to the ignored runner. No source or test
changes occurred in this continuation;no extra review loop was dispatched.

| Call | Host | Exit | Elapsed ms | PASS | Decoded stderr artifact bytes |
|---|---|---|---|---|---|
|1|PowerShell7|0|3433|50|0|
|2|Windows PowerShell5.1|0|1853|50|0|

Both saved records Completed=true;parent checked the child's exit before the
next host. All existing assertions ran,including both multiply-linked file
refusals,reparse directory read/write refusal,and exclusive-writer refusal.
No checks skipped,no automatic retry,no fixture deletion. This demonstrates
the hardlink setup succeeded under scoped sandbox elevation;it does not by
itself prove the exact cause of the earlier access denial or Windows UAC state.

Elevated index SHA256:
`C9D7AE456A7DF8F8FD685B6C6EC7F5DE9C57F1B492CA17A223CCBF0F6AFE4701`.
Source remains SHA7710A1149B42CFD0C2505988CEE6748F399C3FB5E8C34D795B35FBCB8558BB81;
tests remain SHAE54BE44B09391A02068CF2E4B203B1908C019766217AA4CDF368DB40A1821072.
Readonly hash checks confirm both older Store/StoreRepair indexes unchanged.
Elevated2/2 consumed;older3/4 and1/2 unused slots stay frozen,not reusable.

Scope verdict:existing offline storage suite green on both hosts;known numeric
fixture blocker fixed;no unresolved Important finding from the single source
review. Not a claim that every assertion was individually observed RED:original
RED stopped at the missing API guard,and earlier GREEN stopped at fixture setup.
Deferred minors still apply:empty-origin refusal fixture is not isolated and
nonempty UTF16 persistence is not exercised. No full-project suite,live capture,
historical raw-byte reconstruction,crash durability/authenticity,Stage B,I2/I3
or security-disposition acceptance. Next bounded improvement would address the
two minor coverage gaps with its own approved verification allowance.
