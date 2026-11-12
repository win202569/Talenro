# Account alignment: isolated native build and fixed binding proposal

Status: written design approved; implementation plan pending confirmation. No build, restore,
source/entry edit or application launch performed to prepare this document.
Independent builder check: no important unanswered questions. Exact commands,
deadlines and test organization remain implementation-plan decisions as stated below.

## Goal and evidence

Prepare a separately identifiable native artifact for the reviewed Center candidate
e6c77ddf (checkpoint6829e82a), then a default-inert observer bound to its measured
hashes. Offline19/19 and97/97 verify scope, not rendered layout. Preserve the old
email-geometry binary, observer and raw run55550244cd6449eab7d870733015eb23.

## Approaches

User selected new isolated output and a separate hash-bound observation entry.
Proposed implementation: one App-only, fixed-name build variant; new entry copied
from the reviewed observer with a narrowly verified binding delta. This retains
the established project reference paths and guards, at the cost of a second fixed
entry whose behavior must remain in parity with the original.

Rejected alternatives: replacing old email-geometry artifacts loses the preserved
baseline; global BaseOutputPath/BaseIntermediateOutputPath overrides can propagate
into Core; a new worktree adds branch/path handling to the existing branch-bound
observer. A general arbitrary-path/hash argument is not needed and is excluded.

## Build isolation (proposed, not executed)

Add optional `EmailGeometryBuildVariant` with only empty or `account-center` accepted.
Empty preserves every current path/behavior. `account-center` is valid only for
App Debug, EnableLayoutDiagnostics=true, EnableEmailGeometryRecording=true,
EnableCloseStageRecording=false and empty CloseStageVerificationMode. Reject all
other nonempty values/combinations before build/restore; do not weaken existing gates.

App-only conditional paths in Directory.Build.props, after existing geometry paths:
- BaseOutputPath: `bin/email-geometry-account-center/`
- BaseIntermediateOutputPath: `obj/email-geometry-account-center/`
- MSBuildProjectExtensionsPath: absolute App project path plus that new obj suffix.

Core and Tests retain existing paths and must not receive App path assignments.
No arbitrary output override is accepted as a substitute. Tests must verify default
and selected App paths, unchanged Core paths, required defines/Compile items and
invalid variant/configuration rejection, using evaluation without native compilation.

Before any future build: clean committed source checkpoint; ordinary token; no
relevant running application/build process; both candidate directories absent;
resolved paths within App's expected bin/obj roots and no reparse ancestors.
If a destination exists, stop—no delete, overwrite or automatic suffix/retry.
Inventory old EXE/DLL/receipt and lockfile hashes before/after.

New obj requires new assets. Never copy old project.assets.json. Propose a separately
approved cache-only locked restore, existing local cache only, no remote package
source/download/install/version changes. Retain warnings/audit reporting; offline
audit availability is not a vulnerability clearance. Missing cache stops the work.
Restore/build must use identical fixed properties. Build is Debug win-x64 with
--no-restore after validated assets. Exact commands and bounded observation deadlines
belong in the implementation plan after this design is approved; no implicit rebuild
retry. Any timeout stops subsequent stages, without forced termination.

## Artifact inspection and provenance

Read files/PE metadata without loading/executing App code. Confirm candidate target
path, AMD64, expected geometry/layout types, absence of close-stage types, deps and
runtimeconfig presence. Review generated file list/restore metadata for path leakage;
no new writes into preserved email-geometry/layout artifact namespaces are acceptable.
Do not assume unchanged Core outputs merely because its paths are separate; record
any normal dependency build outputs while preserving old native evidence hashes.

Record source commit, tool versions, exact properties, lockfile hashes, full build
results/warnings, sizes and SHA256 of new EXE/DLL. Source tree and build inputs must
be unchanged across capture/build; fail on unexpected tracked changes. An EXE may
remain identical if it is only an app host: verify composition rather than require
both hashes to differ. No new hashes may be filled in before actual inspection.

## Separate fixed observer binding

Proposed files: `clients/windows/tests/observe-email-geometry-account-center.ps1`
and corresponding `.tests.ps1`. Preserve original entry/tests and all five shared
observation/native files unchanged. New public interface remains only RunApproved;
without it return NotStarted before imports, native compilation, CIM or writes.

Allowed clone delta: candidate executable path in policy AND manifest; measured
EXE/DLL hash literals; the new entry's own filename in Sources. Keep arguments,
branch/clean-tree/source checks, ordinary-token check, process absence/identity,
reparse/create-new destinations, run correlation, failure retention, getter fixes,
cleanup and360000ms cap/2000ms sleep unchanged. Receipts remain under the candidate
artifact directory, never the preserved old directory. Commit provenance and new
binding together in a sanitized build record; a fresh clone is not a binary restore.

Offline tests must exercise default inert behavior, reject public overrides and
wrong artifacts, verify manifest/policy/path/hash agreement, and replay the old
observer's failure/cleanup cases with external effects stubbed. A parity check must
allow only the enumerated clone differences, not blanket removal of path/hash text.
Run old regression unchanged; do not reduce its assertions to accommodate the clone.
Any helper needed to share tests must be test-only and preserve old coverage.
Once bound, commit and review the entry before considering a later live request.

## Success, limits and handoff

This stage succeeds only with isolated build evidence, preserved baseline hashes,
new measured binding, offline regression and independent review. Build/entry approval
does not grant RunApproved. No application launch, UI input, process termination,
system/security change, main merge or C4 acceptance. Live allowance remains0.

Implementation plan must separate: isolation/test changes; approved local restore
and native build/inspection; fixed-entry preparation/regression/review. A failure in
any phase blocks the next; retain records rather than auto-retry. Subsequent live
observation needs its own user approval and new run budget. Native RTL/LTR/narrow
layout quality and full-project suite remain unverified. Clone restores source,
tests and sanitized provenance only; dependencies/artifacts need local preparation.
