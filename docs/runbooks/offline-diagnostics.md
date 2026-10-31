# Offline diagnostics — current operator boundary

The normal documented entry is scripts/diagnose-offline.ps1, not any historical
driver/capture. Current status: IMPLEMENTED BUT NOT ACCEPTED. Independent review
found2Important issues; do not run profiles as a released diagnostic workflow.
See [execution and blockers](../roadmap/2026-10-31-offline-diagnostic-runtime-migration.md)
and [full review](../roadmap/2026-10-31-offline-diagnostic-runtime-review.md).

## Default: inspection only

In the supported PowerShell7.6.5 Core x64 environment, from the repository root:

```powershell
& .\scripts\diagnose-offline.ps1
```

Default Action is Inspect; explicit -Action Inspect is equivalent. It reads
catalog/dependency/history data, shows missing dependencies and matching pinned
identities, and does not import the runtime module, create sessions/profile
outputs or start targets. Authority is None and all five capability flagsfalse.
Inspection and matching files are not permission to execute. This example is
documentation only; writing this runbook did not invoke the entry again.

Supported inspection runtime is PS7.6.5, not a promised Windows PowerShell5.1
fallback. On a new machine without the pinned interpreter/cache, read static
recovery instructions first; do not automatically install/download anything.
An existing PS7 interpreter does not establish the pinned machine/identity/
privilege/runtime prerequisites for execution.

## Execution is not currently released

RunApproved has exactly three hard-coded profiles: entry-migration-core,
entry-migration-owned and entry-migration-live. Profile names and catalog data
are not authorization tokens. There is no arbitrary ScriptPath/Command/
Arguments/OutputRoot/RunId or resume interface. Unknown action/profile, old
Grant/Slot/Resume/Retry/Force/SkipValidation and positional arguments reject
before runtime preparation. CommonParameters never supply execution authority.

The current fixed output namespace and all6hosts/7targets/1review are consumed.
No auto retry, new directory name, deletion, reinitialization or replacement
manifest may reclaim them. Any future execution or repair requires a separately
approved fixed plan/inputs/allowance; do not run historical or current snapshots
to recover authority. The two review blockers require a new revision, not edits
to this frozen test oracle to match its existing results.

Session/Ack are private live references, not files, report fields or serialized
JSON. They cannot be copied, restored after a crash, or transferred to another
shell/computer. A successful outcome/intent/final/portable file cannot resume
execution; a failure or unknown record stops further authorized work.

## History and recovery

Old leaf/packet scripts, their raw outputs, approvals and portable exports remain
byte-preserved evidence. The new entry never forwards old run/slot/grant calls.
They are unsupported for automatic replay, not technically disabled by the OS;
this migration does not repair or certify all historical/product paths.

Recover from the latest codex/c12-b01-task9-coordinator development commit and
[repository recovery](repository-recovery.md), not automatically main. Read
current status and full review before considering a new plan. portable-final
contains original bytes for handoff, not credentials/runtime caches/live refs
or execution permission; never use it to recreate missing evidence as an
asserted original. This operator doc/export is author-written after review,
not an independently reviewed or production-accepted deliverable.

Historical C1/C2 remain open, R2/Minor2 deferred; all five capability flagsfalse.
No production integration, default-suite/CI/Docker change, main merge, release,
formal acceptance or overall C12 completion is claimed.
