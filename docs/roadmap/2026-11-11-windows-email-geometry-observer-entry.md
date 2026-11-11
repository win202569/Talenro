# Fixed email geometry entry execution ledger

Plan:docs/superpowers/plans/2026-11-11-windows-email-geometry-observer-entry.md.
User confirmed inline implementation plus one independent review; no app launch.
BASE89579621. Existing linked coordinator worktree/branch verified clean at start.

Pre-flight: Task2 documents/reviews Task1 contracts; no shared-interface conflict.
Ruling: PowerShell/manual plan ledger substitutes for bundled Bash helpers because
the repository requires PowerShell; risk is bookkeeping drift, mitigated by explicit
base, test commands and tracked records.

## Task1

Added default-inert observe-email-geometry.ps1 and its synthetic test suite. Reuses
the unchanged Invoke-TalenroC5Observation loop and shared native/managed/identity
modules; thin gate/backend/writer follows established C4 contracts without changing
old entries or modules. One internal canonical run ID binds CLI/context/manifest and
app geometry path. Normal process exit always leaves geometry Unreviewed/NotAccepted.

RED: new suite failed because entry missing. During implementation, a text-adaptation
replacement interpreted a dollar-anchor incorrectly and caused a parser error;
literal replacement corrected it before runtime tests. A test incorrectly counted
Windows newline length and omitted the two new geometry-destination checks from its
expected event order; corrected those fixtures. First GREEN170 assertions; added
actual preflight observer/app-directory/claim/reparse collision cases, then178 passed.

Full scoped command: each of observe-email-geometry,observe-c4,observe-c5,
c5-observation-session,observation-session-guard,process-observation,
native-process-observation `.tests.ps1` via separate `pwsh -NoProfile -File`.
All exit0:178 new +152/32/120/48/32/78 existing =640 assertions.
No real launch/CIM/native-helper/sleep/file creation in synthetic fixtures. No App
build/restore/package changes. Old/shared source diff empty against89579621.
New and old artifact hashes remain exactly as recorded in the build report.

## Boundaries

Record roots differ: email-geometry-observations/<run> for manifest/process samples;
email-geometry-records/<run>/geometry.jsonl for app snapshots, plus app `<run>.claim`.
Entry checks app paths but never writes them. All owned observer writes create-new.
Trusted directory precondition remains; metadata checks do not defeat hostile races.
Operator startup30s/normal-close by300s/exit-assessment30s remain human-observation
requirements; shared engine enforces360s cap/2s maximum sleeps, not a UI timer.
Geometry parsing, sample completeness and visual correlation are not inferred from
process exit; require separately scoped evidence review after a separately approved run.

Task1 verification complete; Task2 independent review pending. Live allowance0.
