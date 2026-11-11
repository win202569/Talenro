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

Task1 committed d6da5ec9. Task2 review completed; handoff BLOCKED on one Important.
Live allowance0. This is an unfinished development checkpoint, not entry readiness.

## Independent review and disposition

Fresh read-only reviewer inspected89579621..d6da5ec9 and independently reran640
synthetic assertions. Result:0 Critical/2 Important/0 Minor, ready only with fixes.

Important2 fixed: both Get-Item mocks incorrectly declared Force as an ordinary
parameter, so reparse cases passed on parameter-binding failure. Added exact
reparse-error assertion and existing-ordinary-parent success case: RED observed.
Changed both mocks to switch parameters: GREEN; added direct geometry preflight
assertion of the same intended error. Final new180+existing462=642 assertions pass.
Production destination checker did not change. No second review dispatched.

Important1 UNRESOLVED: New-EGTextWriter suppresses a cleanup error; shared C5 catches
replace aggregated failures with generic exceptions, and entry catch sanitizes again.
Reviewer injected managed-observation plus sample-writer-disposal failures: caller
and abort publication retain only generic session failure, not both categories.
Current behavior fails closed and never accepts geometry, but does not meet the
plan's end-to-end diagnostic requirement. A generic throw is not proof of retention.
Next repair must introduce allowlisted entry-owned failure bookkeeping and clarify
resource ownership (including process disposal), preserving the original Process
and SafeHandle and frozen shared modules. Do not simply expose raw exception text.
Add combined primary/cleanup and final-surface tests before implementation.

Final: Ruling: retain Important1 as an explicit readiness blocker, not a deferred
Minor or silent acceptance — resource wrappers alter the specified direct original
resource transfer contract and need a bounded design decision — cost is delayed
entry readiness; current generic failures remain insufficient for diagnosis.

### Reviewer boundaries ruled on

Final: Ruling: real launch/CIM/native/UI timing/rendering/durability remain untested
under the no-live instruction — synthetic tests establish only synthetic contracts
— cost if mistaken: real environment faults may remain.
Final: Ruling: reviewer did not rehash artifacts; implementer hashes were rechecked
against the build record — later execution must recheck again — cost: artifacts
can drift after this checkpoint.
Final: Ruling: geometry completeness/parsing/readability/diagnosis remain separate
post-run work and Unreviewed/NotAccepted remains mandatory — cost: no clipping
conclusion is available from this entry alone.
Final: Ruling: startup/close timings remain operator duties, not automated timers
— cost: a later operator can miss the route even when process monitoring works.
Final: Ruling: trusted-workspace/filesystem assumption remains; no hostile-race or
function-substitution resistance claimed — cost: this entry is unsuitable for an
untrusted workspace.
Final: Ruling: a synthetic writer disposal failure before delegated stream cleanup
does not prove a real .NET handle leak — retain it as fault-injection evidence only
— cost: actual cleanup behavior requires separate evidence if disputed.
Final: Ruling: PowerShell/manual bookkeeping stands as recorded above — cost remains
bookkeeping drift, mitigated by tracked evidence.
Final: Ruling: frozen shared modules/old entries stay unchanged; their category-loss
interaction is not dismissed and remains Important1 — cost: repair must be designed
at the entry boundary or require a separately approved scope expansion.

Deferred minors: none. No merge, PR, launch, build, restore or installation performed.

## Follow-up design,2026-11-11

User selected entry-owned sanitized failure state and resource adapters instead of
changing shared modules. Short addendum written at
docs/superpowers/specs/2026-11-11-windows-email-geometry-failure-recording-design.md;
awaits approval of the written contract. Author proposals: finite deduplicated category
sets, explicit SessionUnclassified for inaccessible shared failures, one-time adapter
cleanup and information-channel/safe-exception summary without a new file sink.
These are proposed design details, not implemented behavior. No tests rerun for this
documentation-only step;642 remains previous checkpoint evidence. Important1 open.
