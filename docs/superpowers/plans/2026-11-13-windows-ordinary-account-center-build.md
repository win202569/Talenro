# Ordinary Account-Center Build Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for the preserved inline execution method. Steps use checkbox syntax for tracking. Read the approved design first.

**Goal:** Produce a statically inspected ordinary Debug candidate with Center, preserving old App artifacts and observations.

**Architecture:** Add a fixed App-only output variant with diagnostics disabled. Validate isolation first; separately authorize one local locked restore and one conditional build. Inspect resulting files as data, without launching them.

**Tech Stack:** Existing PowerShell7, .NET10/MSBuild/WinUI, xUnit and PEReader; no new dependencies.

**Spec:** [Approved design](../specs/2026-11-13-windows-ordinary-account-center-build-design.md).

Status2026-11-13: checklist confirmed; Task1 implementation and scoped managed
tests executed (RED4/16, GREEN16/16, combined34/34); Task1 checkpoint complete.
Independent review:0 Critical/Important,1 deferred Minor (valid guard acceptance
not directly invoked by tests; predicate correct on inspection).
See [execution report](../../roadmap/2026-11-13-windows-ordinary-account-center-build.md).
Tasks2/3 remain unexecuted and no native build/live allowance exists. Historical
checkbox list below records the prescribed steps, not further run authority.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; existing codex/c12-b01-task9-coordinator branch. Command examples run from clients/windows.
- OrdinaryBuildVariant empty or account-center only; selected variant requires Debug, all three diagnostic switches false, EmailGeometryBuildVariant and CloseStageVerificationMode empty.
- Selected App paths bin/ordinary-account-center/ and obj/ordinary-account-center/; absolute MSBuildProjectExtensionsPath under selected App obj.
- App RID stays win-x64; Core RID stays empty. No global RID/output override, copied assets, dependency/lock update, downloads, installs or settings changes.
- No product XAML/business code changes, observer changes, launch entry, UI input, force-kill, Release acceptance or main merge.
- Before execution: ordinary token, successful relevant process inventory, preserve unrelated work and every preexisting App output/receipt. No destructive cleanup or destination reuse.
- Configuration tests can compile Core/Tests and update their normal outputs; they must not build/restore App. This bounded write scope requires Task1 approval.
- Test commands:120s observation budget each; evaluation child60s; restore/build300s each. Poll at most30s. Timeout/failure stops downstream work, preserve sessions/results; never interpret timeout as exit or automatically retry/kill.
- Managed RED then GREEN is the planned test cycle, not permission for unlimited retries. Unexpected infrastructure errors stop the cycle.
- Raw inventories/snapshots/binaries remain ignored/local; only sanitized source/tests/docs committed. Fresh clone does not restore them.

## Review Focus

1. Property leakage into Core/Tests: Task1 compares evaluated default and selected paths.
2. Invalid configuration reaching restore: Task1 explicitly checks ORDINARYBUILD001 and initial-target ordering.
3. Existing/reparse destination or missing cache: Task2 stops before writes, never repairs by deletion/download.
4. Stale layout paired with a new DLL or hidden diagnostic payload: Task3 verifies generated/deployed XBF, source identity and metadata together.
5. Preserved artifact damage or silent Core writes: Tasks1–3 inventory/snapshot/recheck and separately report allowed Core regeneration.

## Task1 — Fixed App-only variant and configuration tests

**Files:** Modify clients/windows/Directory.Build.props and Talenro.Demo.App/Talenro.Demo.App.csproj; create Talenro.Demo.Tests/OrdinaryBuildIsolationTests.cs.
**Interface:** OrdinaryBuildVariant; initial target ValidateOrdinaryBuildVariant with error ORDINARYBUILD001, placed before existing initial targets. Produces committed configuration/test evidence, not native binaries.

- [ ] Obtain explicit Task1 implementation and managed-test approval. Record baseline HEAD/status, ordinary token and no App/prototype/dotnet/MSBuild/testhost/vstest/XamlCompiler processes. Failed enumeration is not zero processes.
- [ ] Inventory preexisting App bin/obj regular files and all raw receipts, plus App/Core lockfiles. Check reparse entries before recursive traversal; stop rather than follow outside trees. Store hashed inventories outside protected trees in a fresh verified ignored evidence directory. Snapshot existing Core bin/obj and Tests output/restore records before planned test writes; verify copies against originals. No overwrite or deletion.
- [ ] Add focused tests using the existing EmailGeometryIsolationTests process/evaluation pattern without refactoring its tests. Helpers RunProject(string project, params string[] args) return (int Exit, string Text), consume stdout/stderr concurrently, never request Build/Restore, and do not kill on timeout.
- [ ] OrdinaryCenterPathsAndPayload: evaluate selected App BaseOutputPath, BaseIntermediateOutputPath, MSBuildProjectExtensionsPath, MSBuildProjectDirectory, TargetPath, RuntimeIdentifier, DefineConstants, Compile and EmbeddedResource. Assert fixed paths, absolute extensions path, win-x64, TargetPath under selected Debug tree, no LAYOUT_DIAGNOSTICS/EMAIL_GEOMETRY_RECORDING/CLOSE_STAGE_RECORDING, no Diagnostics/ compile items or LayoutFixtures/LayoutStrings resources; ReservedLayoutArguments remains compiled.
- [ ] EmptyVariantPreservesDefaults: compare omitted versus explicit empty property for evaluated properties/items under ordinary defaults and existing geometry account-center configuration. Existing geometry isolation tests must remain unchanged.
- [ ] VariantLeavesDependenciesUnchanged: Core and Tests selected versus empty OrdinaryBuildVariant yield identical evaluated output/extensions/TargetPath/RID properties, under otherwise identical flags. Compare normalized properties rather than unordered JSON. Core empty RID is separately asserted.
- [ ] InvalidOrdinaryVariantFailsEarly: unknown, ../account-center, Release, layout=true, geometry=true, close=true, nonempty email variant, and nonempty close-verification mode. Invoke only ValidateOrdinaryBuildVariant; require nonzero plus ORDINARYBUILD001. Include malformed diagnostic booleans to prove only literal false is accepted for this variant.
- [ ] Run RED once (command below). Expected new path/payload/validation assertions fail; baseline/dependency cases may pass. A missing target is an expected preimplementation symptom but must never satisfy the eventual rejection assertion. Infrastructure/compile failure blocks implementation.
- [ ] Add the fixed App-only property group after existing variant groups. Add ValidateOrdinaryBuildVariant first in InitialTargets; empty variant is inert, selected variant enforces every contract above. Keep all existing validation targets unchanged. No additional generic path parameter.
- [ ] Run GREEN once with the same filter, then the combined focused regression once. Require exit0, no failed/skipped tests. No native App outputs may appear; compare protected inventory. Report Core/Tests output changes separately.
- [ ] Review exact source/test diff, git diff --check, record actual counts/commands/outcomes; commit/push explicit necessary files and verify remote SHA. Stop before Task2 regardless of Task1 success.

Commands ONLY AFTER Task1 approval:

```powershell
dotnet test .\Talenro.Demo.Tests\Talenro.Demo.Tests.csproj --no-restore --filter 'FullyQualifiedName~OrdinaryBuildIsolationTests' --logger 'console;verbosity=normal'
# After implementation, repeat the above once for GREEN, then:
dotnet test .\Talenro.Demo.Tests\Talenro.Demo.Tests.csproj --no-restore --filter 'FullyQualifiedName~OrdinaryBuildIsolationTests|FullyQualifiedName~EmailGeometryIsolationTests' --logger 'console;verbosity=normal'
```

## Task2 — Separately authorized local restore and native build

**Files:** Create ignored clients/windows/.nuget/ordinary-account-center-offline.config with apply_patch; generated candidate App bin/obj and bounded default Core outputs only. No lockfile edits.
**Consumes:** committed Task1 result. **Produces:** one build receipt and exact candidate paths for Task3.

- [ ] Obtain new explicit one-restore/one-conditional-build approval including local config creation, snapshots and possible Core default bin/obj updates. No launch included.
- [ ] Refresh clean source/token/process/tool preflight; no download/install. Resolve both candidate paths under App; both absent, ancestors non-reparse. Recheck old files against preserved inventories; snapshot current Core bin/obj and verify hashes before restore. Do not reuse a failed-run destination automatically.
- [ ] Evaluate App/Core with the fixed flags below using dotnet msbuild -nologo -getProperty:BaseOutputPath,BaseIntermediateOutputPath,MSBuildProjectExtensionsPath,TargetPath,RuntimeIdentifier,DefineConstants. No targets/build/restore. Require exact selected App and unchanged default Core properties. Stop on mismatch.
- [ ] Resolve existing cache and locked packages; verify metadata contentHash against lock and archive SHA512 against sidecar separately. Missing/inconsistent package stops; do not force-evaluate lockfiles. Create local-only config with packageSources/auditSources cleared then existing cache as sole source; globalPackagesFolder identical. Retain audits and offline warnings.
- [ ] Run exactly one locked restore. Require exit0; parse App/Core assets for expected paths, locked mode and only approved cache/config/source. Require lock hashes unchanged. Candidate obj is now owned by this restore, but candidate bin must still be absent before build.
- [ ] Run exactly one no-restore build only after all restore checks pass. Record exit/warnings and keep original terminal session. Any unexpected writes, failure or timeout stops subsequent execution; preserve partial outputs and request direction, not automatic recovery.

Commands ONLY AFTER Task2 approval and checks:

```powershell
$ordinaryArgs = @('-p:Configuration=Debug', '-p:OrdinaryBuildVariant=account-center', '-p:EnableLayoutDiagnostics=false', '-p:EnableEmailGeometryRecording=false', '-p:EnableCloseStageRecording=false', '-p:EmailGeometryBuildVariant=', '-p:CloseStageVerificationMode=')
dotnet restore .\Talenro.Demo.App\Talenro.Demo.App.csproj --locked-mode --configfile .\.nuget\ordinary-account-center-offline.config @ordinaryArgs
# STOP to verify restore/assets/locks before proceeding:
dotnet build .\Talenro.Demo.App\Talenro.Demo.App.csproj --no-restore @ordinaryArgs
```

No global RuntimeIdentifier or Platform property is added. Expected App output:
Talenro.Demo.App/bin/ordinary-account-center/Debug/net10.0-windows10.0.19041.0/win-x64/.

## Task3 — Static evidence and recovery checkpoint

**Files:** Create docs/roadmap/2026-11-13-windows-ordinary-account-center-build.md (execution report); update current-status and repository-recovery. No new executable verifier is required.
**Consumes:** Task2 actual output, source commit and prior inventories. **Produces:** inspected/unaccepted candidate identity or explicit blocker.

- [ ] Read files with PEReader/metadata only; do not Assembly.Load, execute EXE/DLL, or invoke any launch/observation script. Record EXE/DLL x64 identity, SHA256/size, copied Core hash, deps/runtimeconfig target/framework and effective build properties/tool versions.
- [ ] Enumerate all defined types and resources. Require no LayoutPresentation/LayoutOptions/LayoutWindowObserver, geometry/close-stage recorder payload or LayoutFixtures/LayoutStrings embedded resources. Confirm ordinary ReservedLayoutArguments presence and correlate its rejection branch with the build's source; runtime rejection remains unverified.
- [ ] Parse generated Account XAML and require Center on AccountGeometryPanel. Verify obj and deployed Views/AccountPage.xbf hashes identical and correlate with this fresh build's generated-file list; inventory alone or differing XBF hashes is not proof of correct runtime layout.
- [ ] Inspect App generated-file paths all remain in selected App namespaces (inputs/package references are not outputs). Require copied Core matches the dependency output. Rehash every protected old file/receipt/lock; require no changed/missing files, report unexpected additions separately. Preserve full original records regardless of result.
- [ ] Record Core pre/post differences against verified snapshots without overwriting old App dependency copies. No claim that all workspace outputs stayed unchanged.
- [ ] Record per-check PASS/FAIL/UNVERIFIED and warnings, plus successful residual process-query result. A static candidate pass is not native UI, shutdown, Release, dependency-security or overall C4 acceptance. Launch allowance remains0.
- [ ] Review explicit doc/source/test list, links and git diff --check. Commit/push necessary files to existing branch, verify remote==local; do not add raw evidence, caches or binaries. No main merge.

## Self-review and handoff

All approved design constraints map to Tasks1–3; no new product behavior is planned.
Task1 intentionally requires managed-test authority although it does not build App.
Task2 is separately gated even after configuration succeeds. Static inspection
does not generate a launch allowance or weaken existing process/hash guards.
Recommended next authorized scope after checklist confirmation: Task1 only.
