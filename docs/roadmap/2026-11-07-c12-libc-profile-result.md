# Fixed libc profile — implementation progress

Source plan approved after e7e2a9b9. Global44/34/6/217,NOT_ACCEPTED.

## Final scoped result — Build and libc Live passed

User approved resuming only Build180s then Live120s after publication correction.
Source checkpoint d5b2714c63e2f665255f1d8fa9c8c89d93469620 matched remote before
launch. Build4a947b passed in11619.698ms;Live3d8226 passed in66887.2084ms.
Both exited0 with confirmed exit,no timeout/owner failure,and unchanged four module
files. Build stdout/stderr empty;Live stdout1597 bytes,stderr empty,raw output read.
Complete six events/three OK phases for modernc.org/libc v1.72.1:

| Phase | Seconds |
| --- | ---: |
| PREPARE | 2.0218729 |
| ZIP | 0.3651346 |
| DIRECTORY | 64.0042556 |

ZIP and DIRECTORY both match root go.sum/ziphash:
`h1:db1xwJ6u1kE3KHTFTTbe2GCrczHPKzlURP0aDC4NGD0=`.
Binary SHA256 E195ACAE63ED26DF5AC23EC6E9C8257BE81B4AEFE0910A8F7781AD694FFA6F5B.
Launcher checked tested Go/owner sources,18-case owner receipt,complete Build/Live
manifest equality and binary binding. No Go/owner tests rerun;no runtime retry.
Directory hashing dominated this selected run;this does not attribute the earlier
root verifier timeout,prove all dependencies intact,or establish C12 acceptance.

portable-c12-libc-profile-live.json:60 entries,194815 decoded bytes,all lengths and
SHA256 checked against originals;archive SHA256
9A60F7E4A74C87BDB61329C6301D06DFC96A88208B4216AB726141C9FAFDF9E4.
It preserves exact pre-handoff spec/plan,source snapshots,claims,receipts,raw output,
corrected manifest record and packaging recipe. Prior archives remain unchanged and
are referenced by hash. Binary/caches/real module contents excluded;binary hash is
retained. Expected local paths are present;no credential/private payload observed.

Inline author review only:five plan focus areas checked against existing tests and
new runtime binding/output;no new independent review claimed. All finite stage
slots consumed. No cleanup of original evidence,no main merge,no acceptance change.
Next safe step:read-only synthesis of saved text/sqlite/libc timings and coverage
gaps before proposing any further bounded validation;no automatic new run.
Historical checkpoint sections below retain the original sequence,not current status.

## Go selector checkpoint

GoRed08a212/891406:29.693s,exit1,only TestProfileLibc assertion failed with PROFILE
against the declared legacy-only stub. GoGreen5075e6:2.428s,exit0,diagnostic-local
package suite passed (existing16 and new5 top-level tests). Both owned exits
confirmed,no timeout/failure,four module hashes unchanged. Raw stdout/stderr read.
New fixed resolver/CLI parser preserves paired-v1 default and rejects unapproved
names without filesystem access. main resolves options before cache candidates.
No phase hashing or safety code changed. No real libc/build/owner tests yet.
portable-c12-libc-profile-go.json preserves exact RED/GREEN sources and receipts.
This is a partial implementation checkpoint,not final profile or C12 acceptance.

## Owner checkpoint

OwnerRed7e8499 rejected libc with ORDER as expected (0.869s,exit1).
OwnerGreen374e41 passed18 cases (6.251s,exit0),including exactly4 owned synthetic
children:libc completion,wrong-profile stop,3s partial cutoff,native raw mode.
All exits confirmed;paired12 events remains compatible,libc6 events bound explicitly;
unknown profile rejects before I/O/launch. Module hashes unchanged.
portable-c12-libc-profile.json preserves snapshots and fixtures.

## Stopped before Build — manifest record destination exceeds quota

Toold16513's five in-memory manifest comparisons reached publication without a case
error,but Write-C12WaitRecord threw TOTAL_LIMIT at common helper line39. The new
manifest-check.ps1 passed its entire plan workspace as the receipt root;the writer
recursively totals that root,including isolated compilation caches and synthetic
fixtures. This is a diagnostic orchestration error,not libc/cache corruption or
a failing profile test. No durable manifest-check success receipt was created.
The full manifest gate is incomplete;do not promote the in-memory comparisons to
a completed verification. Read-only854810 confirms Build/Live roots absent.

Stopped immediately:Build and Live NOT_RUN,no retry,quota increase,cache deletion
or repair. Go/owner RED/GREEN slots consumed. New launcher/manifest code is saved
as unfinished orchestration,not as an end-to-end verified execution path.
Existing vendored provenance remains unchanged because vendored code was unchanged.

Recommendation:authorize only correcting the receipt destination to a dedicated
fresh root separated from cache/fixture subtrees,and bounded offline revalidation
of manifest publication. Preserve original failure,keep4MiB cap,do not rerun Go or
owner tests. Resume unused Build/Live only after that correction is verified and
the resumed scope is approved;no automatic continuation from this checkpoint.

Archive includes original RED/GREEN snapshots,claims/receipts,synthetic fixture
inputs and failure note/recipe. The4MiB+1 all-zero quota fixture is not included
(synthetic generated filler,not module data);test source records exact construction
and its PASS result. Real caches,binaries,module contents excluded. Expected local
workspace paths occur in receipts;no credential or private input content observed.
Author inline review only. Global NOT_ACCEPTED and44/34/6/217 remain unchanged.
