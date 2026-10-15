# Limited matrix003 — 2026-10-15

## Approved execution and ledger

User separately approved one new26-check limited matrix after mapping-check-001
passed. Base commit:5afd84e903cf884ba23de62e9ece558c96d41963. Execute the existing
matrix-check.ps1 unchanged, with current initial-query metadata, retained
uninstrumented Go1.26.5 copy, new matrix-check-003 identity and independent offline
cache. No historical identity replay or automatic retry. First failure stops.

Scope matches matrix002:22 complete package tests, selected authority tests
matching `^(TestBoundRead|TestCoordinator)`, store compilation only, and original
Envelope/Pair fuzz samples. Maximum24 builds/25 runtimes. Build600s; ordinary
test120s/native150s; fuzz5s/parallel1/test30s/native60s. No dependency download,
SDK patch, Docker/database integration operation, changed security/completion
gates, main merge or broader acceptance. Acceptance stays false even if26 pass.

This is execution-only, not a new implementation plan: no production/helper/test
change or TDD cycle is needed. Existing isolated worktree is reused. Read-only
preflight confirmed the expected development branch/origin and a clean worktree.
Execution results, final audits and recovery notes will be recorded below before
the ordinary development checkpoint is committed and pushed.

## Result

Stopped at check03/nodeleafbinding compile completion gate. Result:
**2PASS/1FAIL/23UNRUN**, MatrixPassed=false, Acceptance=false,
FinalizationFailed=false. Exactly three builds and two package runtimes occurred;
no check03 runtime, check04, later fuzz or retry. Native host PowerShell7.6.5.

| Check | Build ms | Runtime ms | Top-level/subtest PASS | Completion |
| --- | ---: | ---: | --- | --- |
|01 nodeissuermapping |71120.5165 |11282.461 |35/65 |Natural, final0/empty |
|02 nodeissuerbinding |3893.6371 |15287.2763 |39/128 |Natural, final0/empty |
|03 nodeleafbinding |3825.8501 |UNRUN |UNRUN |Build root exit0; completion refused |

All five native roots returned0 without timeout. First four receipts verified
natural shutdown, complete evidence, final active0/empty, no termination. Check03
build UTC14:44:51.978452Z–14:44:55.8043021Z returned0 and copied a test binary,
but natural shutdown and shutdown evidence were false. Thus this is not an
accepted build and does not establish runtime test success for nodeleafbinding.

## Saved failure evidence and limits

Check03 initial job accounting succeeded: aggregate TotalProcesses10, active1,
terminated0. Its single list query succeeded, capacity16/buffer136, assigned1/
returned1. One member's OpenProcess then failed87 before ownership, exit code or
image could be verified; JobMembership was not attempted. Owned=false is the
unverified default, not evidence of foreign ownership. Aggregate10 is lifetime
accounting, not ten simultaneous processes or an identity inventory.

Accounting API UTC14:44:55.8014597Z–.8014665Z; list API
14:44:55.801484Z–.8014925Z; OpenProcess14:44:55.8016874Z–.8020098Z.
Monotonic list-end1405659343398 to open-start1405659345359 at10,000,000Hz
is0.1961ms. These are observation/API boundaries, not member OS exit timestamps.
They locate a failed acquisition after a successful list, but do not prove that
the process exited during this interval, identify it as a compiler/conhost, or
explain the OS-level cause of error87. No historical PID was queried.

The collector refuses unverified member evidence before entering its natural-wait
loop. TerminationRequested=true; ShutdownElapsedMs/NaturalObservationMs=2, not an
exhausted500ms natural wait. Final cleanup verified active0/empty. No surviving
product child, ownership loss, business assertion failure or exact relation to
matrix002's failed runtime member is established. The error87 symptom recurred
in a different phase/package; the underlying cause remains unresolved. No fix,
extra probe, changed wait/rights/predicate or retry was attempted.

## Finalization and evidence preservation

Driver finalization passed:701 recorded product files unchanged; full copied
11536-file SDK matched the frozen reference;14 successfully bound build/runtime
artifacts unchanged. Read-only postchecks revalidated product inventories, those
14 bindings and ten reviewed tooling pins. No new unit suite or full installed
original-SDK audit is claimed. The driver and process collector remain unchanged.

Reference receipt SHA256:
`AA86DADA2C7EF1BB591C78504FB8B5C168D35D836493A574ABC37DDA84880B55`;
copy receipt:`DD62FDC0704E8ADEE49F77AD7E2A91D3A1CDFC6E430ABD9AADBFC74D74F04D10`.

Ignored matrix-validation-003/final-evidence-manifest.json binds25 files: plan,
final result/stopped record, both product inventories, all five native receipts
and their streams, all three build binaries (including refused check03), source
binding and host-context note. SHA256:
`39BE20634E15042482FEE2DF1B05D34B1A88C33BA4F23A14416B232E1BABE030`.
The host-context note is a controller record, not native host attestation.

Independent read-only review found no Critical/Important/Minor issue: checkpoint
documentation ready, not merge/acceptance ready. Reviewer checked all25 manifest
hashes, ten tooling pins,701 current recorded product hashes,14 bound artifacts,
five receipts/streams, query timing arithmetic, collector invariant and recovery
limits. No Git operations, native/PID queries, tests or full SDK audit were run by
the reviewer. Thus branch/base/clean preflight remain controller-verified, final
SDK audit relies on the saved result/pinned driver, and saved records are not
global proof that an unrecorded invocation could never have occurred. No OS cause,
historical member identity or acceptance determination was requested or inferred.

| Local file under matrix-check-003 | SHA256 |
| --- | --- |
| plan.json |8257EFBE22B335D71501CBBECBBC4CFE227C3EE6719FA4E5F7A4679329BD1874 |
| matrix-result.json |88355AB8AC04C04D3535D7A116B546A13B5C47FD572CDF7E8F43D96C2C72E7A7 |
| stopped.json |EB4626B959BF0F3F7B0A1B776B31DA19E76E0EFC945A8DFCE71A20845F2BF981 |
| before/after-product.json |7E328EB5D4E1A43D372F8C86DD2FE938556F226323587143BAEF3E25BD10374A |
|03-nodeleafbinding/build/receipt.json |244646AAE8F4D7C27755152AB207150C389BBD22070A91ACEE060D7EBBD0B844 |
|03-nodeleafbinding/build/stderr.txt |3CB12CCEAE77C9309019B8BFEC031FEF9688FF6DC598BB120B933A6DA0E1E86B |
|03-nodeleafbinding/package.test.exe |EE91187039EFBB142E57D86200C7675672ED2B85BCCD3E7BDB4FEFF2259B29E4 |

## Recovery and remaining work

Preserve this report, current status and diagnostic index on the existing
development branch; read docs/runbooks/repository-recovery.md on a fresh machine.
SDKs, caches, binaries and raw records stay ignored and are not restored by clone.
Use this portable report for progress, not an assumption that local evidence is
available elsewhere. Do not replay matrix-check-003 or continue its skipped steps.

Next useful scope is read-only comparison of saved successful/failed acquisition
paths and collector invariants before proposing another diagnostic or fix. Any
new real process experiment/matrix run needs its own bounded approval. Historical
matrix00112PASS/1FAIL/13UNRUN and matrix0020PASS/1FAIL/25UNRUN are not overwritten.
C11/C12/Task10/B01, database/full acceptance/production/main gates remain open.
