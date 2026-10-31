# Offline record-chain diagnosis — proposed plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for native execution only after explicit approval. No subagents or fresh review are authorized. Steps use checkbox syntax for tracking.

**Goal:** Determine whether the writer-test module changes command resolution or collector state outside its intended boundary, without replaying writer-red or repairing I1/I2.

**Architecture:** A single disposable diagnostic host observes the original module-creation topology using the frozen writer and collector definitions. Its observer holds a separate direct .NET stream and local state, not the functions/state under examination. This is evidence gathering, not a new RED/GREEN attempt or replacement acceptance test.

**Tech Stack:** Windows PowerShell orchestration; pinned PowerShell 7.6.5/Core/x64 diagnostic host; existing .NET types only.

**Spec:** The approved [repair plan](2026-10-31-offline-diagnostic-review-repair.md), its stop rule, and the [actual stopped execution](../../roadmap/2026-10-31-offline-diagnostic-review-repair.md). This proposal requests a separate, smaller diagnostic scope; it does not amend either historical record.

## Status and authority

Draft only. The user's current consent covers read-only investigation and this written proposal, not source creation, module loading, compilation, Prepare or execution. Baseline commit: `09c709d5ecc1e237850bef495f3aff6b85a2003d` on `codex/c12-b01-task9-coordinator`.

## Confirmed evidence versus hypothesis

- The original capture exited 2. Original stdout, stderr and evidence are each zero bytes; parent final says `EvidenceValid=false`, `EvidenceState=Unknown`. No internal ErrorRecord was saved. Its historical first cause remains UNKNOWN.
- `record-boundary.tests.ps1:16–21` creates a dynamic module with the writer and collector definitions, without an explicit export restriction. It then resolves those same command names during and after the case. Its outer-factory identity assertion is after case completion (`:61–62`); an earlier throw skips it.
- Microsoft's [New-Module documentation](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/new-module?view=powershell-7.6) states that functions are exported into the current session by default. Private module execution therefore does not by itself guarantee unaltered caller command resolution. The [scope documentation](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_scopes?view=powershell-7.6) distinguishes module session state from caller scope.
- `capture.ps1:49–70,90–93` publishes through `Write-ODCaptureRecord` and records failures through name-resolved `Capture-Failure`, while checking its own script-level collector state. This is not an independent failure channel if those names resolve to module-bound functions.
- Hypothesis: exported writer/collector functions can shadow the capture functions; collector state and injected factory resolution then diverge from the outer capture's assumptions. This is a specific mechanism to investigate, NOT a recovered original exception or proof that every missing-output symptom has that cause. Caller-scriptblock binding and other early errors remain alternatives.

## Approaches and recommendation

1. **Read-only explanation alone:** no runtime risk, but cannot establish actual command binding on the pinned engine. This turn stops here unless further permission is given.
2. **One scope-only diagnostic host (recommended):** observes bindings and state without compiling the fault fixture or invoking writer IO. Smallest proposed next step that can distinguish module export from caller-scriptblock/state assumptions.
3. **Rebuild and rerun the full writer test:** could expose more errors, but exceeds the necessary question, risks another dependent evidence channel, and conflicts with frozen-run handling. Not requested by this proposal.

Recommendation is the author's, not an already approved user choice.

## Global constraints

- Preserve all 2317 seed pins, all 25 frozen sources/snapshots, the empty evidence/raw files, historical reports and portable-stopped archive. No edits, overwrites, replay, refreeze, fabricated errors or budget refunds.
- Original spent allowance remains 1 host / 0 targets / 1 in-memory compilation / 0 reviews; remaining 6 hosts / 7 targets / 1 compilation / conditional 1 review stays FROZEN, including after a successful diagnosis.
- Proposed NEW ceiling: **1 diagnostic host, 0 target children, 0 compilations, 0 reviews**. No Docker, network execution, installation, Session/Ack, RunApproved, product profile or fault-stream fixture.
- Use the exact engine path, engine/SMA hashes, non-admin identity and owned-process safety checks specified in section 4 of the approved repair plan. Missing/mismatched context stops before launch, without substitutions.
- No run in this planning turn. No repair of the frozen test, production writer or validator under this proposal, even if the diagnostic mechanism is confirmed.

## Review focus

- Exported command shadowing: record function identities and module associations immediately after module creation, before any case assertion.
- Collector-state separation: observe outer and module state independently; do not conclude from a single scope's null/non-null value.
- Observer contamination: retain local observer state/stream and use direct .NET calls; never call the tested factory, collector or writer to publish diagnostics.
- Unexpected abort or partial evidence: preserve all original bytes; incomplete trace is UNKNOWN, not a negative result or permission to retry.
- Cross-machine recovery: commit diagnostic sources and portable actual records only if later authorized/executed; cloning never resumes frozen allowances.

## Task 1 — prepare and perform one binding observation (NOT AUTHORIZED YET)

**Future files:** create only `docs/diagnostics/isolated-fuzz/p2b-record-chain-diagnosis/{README.md,observe.ps1,driver.ps1}` and new ignored packet `.superpowers/sdd/2026-10-31-offline-record-chain-diagnosis/`. No existing runtime source changes. The new namespace is solely the explicitly proposed diagnostic, not a way to replay a stopped grant.

**Interfaces:** `observe.ps1 -PacketRoot <absolute-path> -SeedSHA256 <actual-seed-sha>` consumes the immutable seed definitions as text. `driver.ps1 -Action RunOnce` owns exactly one diagnostic host. No generic Grant/profile/retry interface.

- [ ] Read this proposal and the original approved plan; record explicit user authorization before creating diagnostic sources. Inspect existing ownership/drain logic as text; do not import/run the frozen driver.
- [ ] Implement a scope-only observer. Define the outer writer/collector using the frozen definitions, retain their ScriptBlock references, initialize outer collector state, and create the module with the exact frozen `New-Module` construction. Do not invoke either writer or `New-ODRecordStream`.
- [ ] Record ordered stages `outer-before`, `module-created`, `module-scriptblock`, `outer-after`, `complete`. At each applicable stage include all four command names (`New-ODRecordStream`, both writers, `Capture-Failure`), ScriptBlock reference equality versus original, module association, outer/module primary-null and secondary-count facts. Keep actual identities separate from stable boolean projections; do not treat dynamic module names as fixed expected literals.
- [ ] Inside the same `& $module { ... }` topology, set only a uniquely named script-scope sentinel and observe its visibility from module and outer scope. Do not replace factories, construct fake ErrorRecords, invoke the collector with invented errors, or run a unit case. This observes binding/state ownership without reenacting the failed operation.
- [ ] Open one independent observer stream with CreateNew before module creation; write each bounded ordered JSONL stage once and flush directly through its retained .NET reference. Limit total evidence to 1 MiB. Catch actual unexpected ErrorRecords locally and serialize Type/Message/ErrorId/Position/Stack independently; preserve publication failures as UNKNOWN, no fallback that overwrites/reconstructs a stage. Dispose the held stream once; no reopen or retry of a failed write/flush/close.
- [ ] Driver validates source AST/declared dependencies and records hashes of the three new files plus seed/loaded-source hashes before any launch. Existing sources remain immutable. Create packet/intent only once; reject existing paths. Snapshot new diagnostic inputs once, not another repair Prepare. Verify hashes before and after launch.
- [ ] Launch at most one owned pinned host with redirected stdout/stderr, 120000 ms host deadline, exact-exit check 2000 ms and shared drain 2000 ms. Preserve association/ownership checks and no broad process killing. Every start attempt consumes the single launch allowance; any failure stops. No self-test host or automatic rerun. Parent saves actual raw streams, exit, trace, pins, intent/final and complete tool-return chunks before interpreting results.
- [ ] Classify separately: trace structurally complete or UNKNOWN; command shadowing observed/not observed; sentinel outer-visible/module-only/other. Unexpected fields, missing/duplicate/out-of-order stages, changed pins or incomplete exit/drain are UNKNOWN and stop. A complete diagnostic may support/refute the scope hypothesis, but is never writer-red PASS, I1/I2 repair, or recovery of the original ErrorRecord.

## Task 2 — evidence-only handoff and separate repair decision

**Future files:** new `docs/roadmap/2026-10-31-offline-record-chain-diagnosis.md`, diagnostic `portable-final.json`, and status/recovery updates. No old report edits.

- [ ] Preserve actual diagnostic packet bytes and directory inventory, including zero-byte/partial files and unknown states. Check portable bytes against originals. Commit only explicit necessary sources/docs/portable data, not ignored raw packet/cache/secrets.
- [ ] Explain whether export, script-scope visibility, both or neither were observed. List what remains unproven about the historical failure. Do not silently apply `Export-ModuleMember`, relocate state, change error assertions, or alter the failed evidence.
- [ ] Update current status/recovery, push the existing development branch and verify remote SHA equals local HEAD. No main merge. Stop and request a separately written repair/verification adjustment; do not resume the original 6/7/1/review remainder.

## Self-review and approval gate

The five review risks map to Task 1's stage/identity/state/independent-publication/UNKNOWN checks and Task 2's portable handoff. One proposed host, no nested target, no compilation and no review; neither I1 nor I2 is modified. Static observations above are source/documentation findings, not new runtime results. This planning checkpoint creates only this document and updates status/recovery; it does not create or run the future diagnostic files.

Planning-only preservation checks: all 2317 existing pin hashes were read and matched; seed manifest still hashes to `9091825652E469001E51D8A510B12E0F84013C38509D740C649942DC2EAAFF6B`. This byte-integrity check did not execute Prepare, a test/module, compilation or a diagnostic host.

Please approve or revise the **single scope-only diagnostic stage** before implementation/execution. Native execution is recommended; no agent or independent review allowance is requested.
