# Offline record-chain diagnosis — scoped observation complete

Status: COMPLETE_SCOPE_OBSERVATION_ONLY. Baseline `aac17117`; user explicitly
approved the [single-host diagnostic plan](../superpowers/plans/2026-10-31-offline-record-chain-diagnosis.md),
SHA `953AEDFEFA0C2BF3C48ED8E3064696C76729764D512AD09AE4D6982FDCC4D76A`.
No product/test repair or I1/I2 acceptance was authorized or performed.

## Actual result

One pinned PowerShell 7.6.5/Core/x64 non-admin diagnostic host, PID11488,
exited0. Exact owned exit and both-stream drain completed; no timeout,
termination, rescue or parent error. Original stdout/stderr are both empty.
Trace contains all five ordered stages, 4762 original JSONL bytes; parent
classification COMPLETE. Full original tool return is preserved, not only
the author's summary. Three new source inputs were snapshotted before launch.

Before module creation, all four command ScriptBlock references matched the
outer originals with no module association. Immediately after the exact frozen
New-Module construction, and through the final in-function stage, each resolved
to a different ScriptBlock associated with the same actual dynamic module:

- New-ODRecordStream
- Write-ODCaptureRecord
- Write-ODProfileIntentRecord
- Capture-Failure

The unique script-scope sentinel set inside `& $module { ... }` was visible
in module state and absent from outer script state. Outer collector remained
present/null with zero secondary errors; module collector variables were
absent throughout this scope-only observation. No collector was invoked and
no factory/writer IO operation or fault-stream fixture was invoked.

This establishes command shadowing during the diagnostic caller function and
separate script-state ownership on the pinned engine. It supports the proposed
binding/state mismatch mechanism. It does NOT establish which ErrorRecord was
first in the stopped writer-red, whether every missing-output symptom shares
this cause, or how imported bindings persist after the caller function unwinds.
No ErrorRecord was invented to answer those questions. The original capture
failure's historical first cause stays UNKNOWN and its zero-byte evidence
remains unchanged. The original error assertions may have independent defects;
this observation neither executes nor validates them.

## Integrity, budget and portable handoff

New allowance consumed: 1 diagnostic host / 0 target children / 0 compilations /
0 reviews; exhausted, no retries. Original repair remainder 6 hosts / 7 targets /
1 compilation / conditional 1 review stays FROZEN. I1/I2 remain unresolved,
C1/C2 open, R2/Minor2 deferred and all five capability flags false. No main merge.

The parent verified 2365 pin entries before and after launch: all original2317,
three new diagnostic sources, seed/engine/SMA/approved plan/stopped portable,
and all40original stopped-packet files. Original seed manifest remains SHA
`9091825652E469001E51D8A510B12E0F84013C38509D740C649942DC2EAAFF6B`.
No original frozen source, raw/evidence, snapshot, report or approved plan was
edited. Pin verification and archive byte comparison are author integrity
checks, not another runtime test or independent review.

`docs/diagnostics/isolated-fuzz/p2b-record-chain-diagnosis/portable-final.json`
archives actual diagnostic packet bytes/directories, including empty streams,
source/plan snapshots, intent, start consumption, pins, final/assessment, full
tool return and ledger. Actual export: 15 files / 2 directories, all bytes
compared exactly against originals, including both zero-byte raw streams.
Portable SHA (LF JSON formatting): `F47946FC985FE6F80B98D20057E38944FC8AE46DEA0F03EFC4BB1E56CD0459DB`.
JSON layout was normalized to LF for Git preservation; decoded original bytes
were re-compared unchanged. Source snapshots preserve original bytes regardless
of another computer's checkout line-ending conversion; replay is not authorized.
It confers no execution permission or restored budget.
Use the latest existing development-branch commit, not automatically main,
when recovering on another computer. No cache, identity or live state is restored.

## Ruling and next boundary

Approved finite native diagnosis replaced skill-default setup/Bash/TDD/suite/
task-done reruns, reviewers and workspace deletion. Cost: this author-written
observer is not regression-verified, and the result is limited to one observed
binding topology; it is not production, full-repository or full-C12 validation.

Next recommended work is a separately written record-isolation repair and
verification adjustment, preserving the stopped seed/test/oracle and making
normal capture publication independent of injected test functions/state. Do not
silently apply an export restriction, change error assertions or resume old
grants. This report authorizes no additional implementation or execution.
