# Account Alignment Native Isolation and Binding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for the preserved inline execution method, or superpowers:subagent-driven-development if the user changes that choice. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce isolated, inspected native candidate artifacts and a separately fixed, default-inert observation entry without launching the application.

**Architecture:** Add one fixed App-only build variant, retaining default paths and Core paths. After separately authorized local restore/build, inspect artifacts as data and bind a reviewed clone of the observer to measured hashes. Shared runtime/observation code is not refactored.

**Tech Stack:** PowerShell7, existing .NET10/WinUI SDK/MSBuild, xUnit, PEReader; no new dependencies.

**Spec:** [Approved design](../specs/2026-11-12-windows-account-alignment-native-binding-design.md).

## Global Constraints

Approved execution revision2026-11-12: do not pass RuntimeIdentifier globally;
App already defines win-x64, while Core must retain empty RID/default output.
The original global RID command caused NU1004; one separately approved corrected
restore/build succeeded after preserving failed App/Core restore records. The
known-owned obj reuse exception was specific to that approved recovery; fresh
destinations still require absence checks. See the execution record for evidence.

- Variant: empty or `account-center` only. Empty retains all current behavior.
- Selected variant requires App Debug, layout=true, geometry=true, close-stage=false, empty CloseStageVerificationMode.
- App paths: bin/email-geometry-account-center/ and obj/email-geometry-account-center/; absolute MSBuildProjectExtensionsPath under App.
- No arbitrary output override, old artifact/entry replacement, copied project.assets.json, package download, dependency version change or security change.
- No application launch, UI input, process termination, main merge or C4 acceptance. Live allowance remains0.
- Task2 cache-only locked restore/native build requires explicit separate user authorization, even after plan confirmation. Task3 depends on successful Task2, never guessed hashes.
- Run commands in ordinary-token approved external context (restricted context previously denied localhost test communication); reject administrator token or relevant existing processes before execution.
- Test/evaluation commands:120s observation budget each; restore/build:300s each. Poll at most30s at a time. On failure/timeout stop downstream tasks, no automatic retry or force-kill; record outstanding process/session state.
- Working directory for commands below: clients/windows in the coordinator worktree. Use discovered absolute pwsh when not on PATH.

## Review Focus

1. Variant property leaking paths into Core/Tests: Task1 comparison of each project's evaluated default/variant paths.
2. Invalid variant entering restore: Task1 initial-target rejection plus Task2 preflight; no network/restore in negative tests.
3. Existing/reparse candidate destination or missing local cache: Task2 fail before writes, never delete/reuse old output.
4. Policy path and manifest/hash disagree: Task3 executable-path consistency and negative hash tests.
5. Clone drift silently weakening failure/cleanup/default-inert behavior: Task3 exact parity and complete synthetic cases, old tests unchanged.

## Task 1: Fixed App-only variant with failing tests first

**Files:** Modify clients/windows/Directory.Build.props; clients/windows/Talenro.Demo.App/Talenro.Demo.App.csproj; clients/windows/Talenro.Demo.Tests/EmailGeometryIsolationTests.cs. Update execution/status/recovery documents alongside each checkpoint.

**Interface:** `EmailGeometryBuildVariant` MSBuild property; new App initial target `ValidateEmailGeometryBuildVariant`, error `EMAILGEOMETRY002`. It checks nonempty variants before existing initial targets. Core/Tests do not acquire App-only path overrides or validation targets.

- [ ] Record clean branch/source and old artifact/lock/receipt inventory. Read the prior build record and approved design; preserve unrelated work. Do not repeat the completed alignment test cycle.
- [ ] Add `AccountCenterPathsStayInApp` and `VariantLeavesCoreAndTestsPathsUnchanged` using existing MSBuild property-evaluation test helper (extend helper to accept project path, keep all old calls). Assert selected App BaseOutputPath and BaseIntermediateOutputPath equal the fixed paths above, extensions path equals its absolute obj path, TargetPath remains under selected output, EMAIL_GEOMETRY_RECORDING/layout definitions present and close-stage types absent. Compare Core/Tests evaluated paths to their empty-variant results, not hardcoded host paths.
- [ ] Add `InvalidAccountCenterVariantFailsEarly`: unknown value, path-like value, Release, layout=false, geometry=false, close-stage=true, and nonempty close verification. Target only ValidateEmailGeometryBuildVariant; assert exit!=0 and EMAILGEOMETRY002. Require empty variant's original tests unchanged. Existing unknown-target errors are not a valid assertion of the new guard.
- [ ] RED run once:
  `dotnet test .\Talenro.Demo.Tests\Talenro.Demo.Tests.csproj --no-restore --filter 'FullyQualifiedName~EmailGeometryIsolationTests' --logger 'console;verbosity=normal'`
  Expected new path/guard assertions fail, old tests stay green. Infrastructure/compile failures block continuation.
- [ ] Add the fixed conditional App paths after existing geometry group; add initial validation target before existing initial targets. No public arbitrary-directory property. No App layout/observer/source change beyond this build configuration.
- [ ] GREEN same command, then `dotnet test .\Talenro.Demo.Tests\Talenro.Demo.Tests.csproj --no-restore --filter 'FullyQualifiedName~EmailGeometry' --logger 'console;verbosity=normal'`. Require exit0/no failures/skips. Inspect no native output was produced; preserve default-mode checks. Commit explicit files and sanitized execution record.

## Task 2: Separately approved local restore, native build and inspection

**Files:** Create ignored clients/windows/.nuget/account-center-offline.config via apply_patch; native generated files only under selected App namespace plus normal Core dependency paths. Create docs/roadmap/2026-11-12-windows-account-alignment-native-binding.md for sanitized provenance. Do not edit package lock files.

**Interface:** Produces source commit + measured CandidateExeSha256/CandidateDllSha256 and artifact inventory for Task3. Binary path is App/bin/email-geometry-account-center/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe. Hash values do not exist until inspected.

- [ ] STOP for explicit cache-only restore/build authorization. Approval includes one restore and one subsequent no-restore build only, conditional on preflight; no launch.
- [ ] Resolve candidate bin/obj paths inside App and inspect ancestors for reparse points; both absent before restore. Check ordinary token, no relevant build/app processes, clean committed sources. Snapshot old email-geometry/layout EXE/DLL, four prior run records and App/Core lock hashes. Record dotnet/installed native tool versions without installing anything.
- [ ] Evaluate selected App and Core properties using `dotnet msbuild <project> -nologo -getProperty:BaseOutputPath,BaseIntermediateOutputPath,MSBuildProjectExtensionsPath,TargetPath,DefineConstants` plus the fixed properties below. Require selected App/default Core paths. Evaluation must not build or restore.
- [ ] Verify required locked App/Core package versions exist in the existing project cache. Create the new local config with packageSources and auditSources cleared and only that absolute cache directory, globalPackagesFolder identical. Reject missing cache rather than download; retain audits/warnings (offline audit is not security clearance).
- [ ] Use identical properties for restore/build; these are commands to execute only after the above gate, not instructions to run during planning:

```powershell
$variantArgs = @('-p:Configuration=Debug', '-p:EnableLayoutDiagnostics=true', '-p:EnableEmailGeometryRecording=true', '-p:EnableCloseStageRecording=false', '-p:CloseStageVerificationMode=', '-p:EmailGeometryBuildVariant=account-center')
dotnet restore .\Talenro.Demo.App\Talenro.Demo.App.csproj --locked-mode --configfile .\.nuget\account-center-offline.config @variantArgs
# Inspect restore result/assets before executing the next command.
dotnet build .\Talenro.Demo.App\Talenro.Demo.App.csproj --no-restore @variantArgs
```

- [ ] After restore, require exit0, unchanged lockfiles and asset metadata naming only approved cache/config/sources and new App obj namespace. Candidate obj now legitimately exists from this restore; do not apply the pre-restore absence check again. Require candidate bin still absent before first build. Stop if unexpected writes/configuration appear.
- [ ] After build exit0, inspect source git status and generated file list/paths. Read EXE/DLL as bytes with PEReader using the pattern in tests/verify-layout-artifacts.ps1 (do not run that script unmodified against unrelated modes). Require AMD64 and expected EmailGeometryObserver/targets/evaluator/coordinator/receipt/LayoutOptions types, no CloseStage types, valid deps/runtimeconfig. Dispose readers/streams; never Assembly.Load or execute App. Verify fixed target and Core provenance, not merely file presence.
- [ ] Record EXE/DLL bytes/hashes, source commit, fixed properties, tool versions, warnings and preserved old hashes. Equal app-host EXE hash is allowed; source-bound DLL/composition still checked. Record Core outputs if rebuilt. Commit/push sanitized record, not binaries/config/cache/logs. Failure blocks Task3 and preserves all evidence.

## Task 3: Fixed candidate observer and full synthetic parity

**Files:** Create clients/windows/tests/observe-email-geometry-account-center.ps1 and observe-email-geometry-account-center.tests.ps1. Preserve old entry/tests and all shared modules/native sources. Update execution/status/recovery docs.

**Interface:** New entry retains only `[switch]$RunApproved`; no switch returns one NotStarted result before any external effects. Get-EGPolicy/manifest retain original shapes and use Task2 measured literals and fixed candidate path. Sources includes new entry name and same five shared files.

- [ ] Write candidate synthetic tests first, copying existing full observer test harness with exact binding literal changes only; original harness stays untouched. Use Task2 measured hashes as independent expected literals. Add parity tests before entry exists: default no-start, public overrides rejected, candidate policy/manifest/receipt roots agree, old/wrong hashes rejected, source list names candidate entry. Expected initial run fails for missing candidate entry, not a dependency error.
- [ ] Define parity by cloning old entry text in-memory and applying only four exact token substitutions: old executable-relative path (policy/manifest), old EXE hash, old DLL hash, old own-script filename. Assert expected occurrence counts, exact expected old tokens, no accidental substring replacements, and equality to new entry after consistent newline normalization. Add negative parity fixture changing360000ms or removing a preflight call; it must fail. Do not normalize all paths/hashes away or execute comparison fixtures.
- [ ] Create candidate entry with only those approved differences. Existing arguments,360000ms cap/2000ms sleep, clean-tree/token/CIM/identity/reparse/create-new checks, error retention, explicit getter invocation and cleanup unchanged. Never insert placeholder hashes; never invoke -RunApproved.
- [ ] Run with absolute pwsh `-NoProfile -File` separately for candidate harness then old harness. Both must pass, exit0; old count remains282 assertions (per recorded baseline), candidate includes all282 plus new parity/binding checks. External effects replaced before entry evaluation. Run unchanged c5-observation-session.tests.ps1 (120 assertions; historical462 was six suites combined, not this script); record actual counts. No native compile/loader/live tests included.
- [ ] Verify full diff/source hashes and old/native artifacts unchanged; commit source and measured provenance together. Request fresh read-only final review of all three tasks against spec/five focus items. Resolve important findings within scope with regression evidence; do not launch to satisfy review. Update current-status and repository-recovery, push current dev branch and confirm remote SHA. Leave branch/worktree/old evidence intact.

## Handoff and self-review

Self-review: source isolation, provenance and fixed binding are three dependent
deliverables; Task2 supplies Task3 hashes, Task1 supplies Task2 paths. All five focus
items have explicit gates/tests. Asset-directory existence changes after restore
are distinguished from preflight collisions. No task assumes native build success
or visual repair. Complete commands require locally checked paths/tool availability.

Preserve inline execution with final independent review. Plan confirmation pending;
after confirmation execute Task1 first, then request the explicit Task2 authority.
Task3 remains conditional on verified artifacts. No implementation or command above
was executed while writing this plan; live allowance0 and C4 NOT_ACCEPTED.
