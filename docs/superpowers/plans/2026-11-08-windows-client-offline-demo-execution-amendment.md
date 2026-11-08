# Windows Offline Demo Execution Amendment

> **For agentic workers:** REQUIRED SUB-SKILL: superpowers:executing-plans for
> proposed inline execution. Await user review before implementation.

**Goal:** Make Tasks2–4 of the parent plan executable with pinned inputs and a finite allowance.

**Architecture:** Retain the pure in-memory reducer and native four-view shell.
Use an unpackaged x64 app with local runtime files, not MSIX installation.
Model tests remain independent of WinUI and do not launch the app.

**Tech Stack:** C#, .NET SDK10.0.403, VS MSBuild18.11.1.51303, WinUI3, VSTest/xUnit.

**Spec:** [Approved design](../specs/2026-11-07-windows-client-four-screen-design.md).
Read with [parent plan](2026-11-07-windows-client-offline-demo.md); its interfaces,
file ownership and behavior requirements remain authoritative.

Status2026-11-08: approved by user after5c61627c for Tasks2–4. First R1 failed;
separately approved retry and L1 passed. First SessionTests invocation aborted
on testhost connection timeout; no behavioral RED. Execution stopped. See
[execution record](../../roadmap/2026-11-08-windows-client-offline-demo-execution.md).
The following original proposal describes intended work, not completed work.
This is documentation, not restore,
build, test or launch evidence. Tool metadata is verified; this package combination
has NOT been restored or compiled. Supersedes the parent's unresolved command recipe,
not its acceptance gates. No new SDK installation is included.

## Global constraints

- Four native offline pages; fixed demo@example.com, read-only masked password.
- Permanent notice: “离线演示，不会连接 VPN 或提交账号信息”. No runtime network,
  persistence, telemetry, account changes, VPN, IPC, driver or payment actions.
- No Developer Mode/security changes, IDE launch, MSIX registration or elevation
  for app execution. Dependency restoration alone may use NuGet network access.
- Task11/12 and44/34/6/217,NOT_ACCEPTED remain unchanged; no main merge.
- Native launch allowance is **zero** here. Task5 needs a separate reviewed allowance.

## Inputs and proposed pins

Paths below are relative to `clients/windows/`, created only after approval.

|File / input|Decision|
|---|---|
|global.json|sdk.version10.0.403, rollForward=disable, allowPrerelease=false|
|Directory.Build.props|Nullable/ImplicitUsings enabled; RestorePackagesWithLockFile=true; deterministic builds|
|Talenro.Demo.Core/Talenro.Demo.Core.csproj|net10.0; no direct packages|
|Talenro.Demo.Tests/Talenro.Demo.Tests.csproj|net10.0; IsTestProject=true; Core project reference|
|Test direct packages|Microsoft.NET.Test.Sdk18.11.0; xunit2.9.3; xunit.runner.visualstudio3.1.5 with PrivateAssets=all|
|Talenro.Demo.App/Talenro.Demo.App.csproj|net10.0-windows10.0.19041.0; SupportedOSPlatformVersion10.0.19041.0; Core reference|
|App direct package|Microsoft.WindowsAppSDK2.6.5|
|App properties|OutputType=WinExe; UseWinUI=true; WindowsPackageType=None; RuntimeIdentifier=win-x64; PlatformTarget=x64; SelfContained=true; WindowsAppSDKSelfContained=true; PublishSingleFile=false|
|NuGet.Config|clear packageSources; only https://api.nuget.org/v3/index.json; no credentials or fallback feeds|
|Each project/packages.lock.json|Generate on first successful restore, review exact transitive graph/content hashes, commit; subsequent restoration locked|
|.gitignore|Local bin/, obj/, .nuget/, .dotnet-home/, TestResults/ only; do not ignore source or lockfiles|

Use exact NuGet version ranges `[version]` for direct pins, not wildcards. SDK
reference/runtime packs are first-restore dependencies too; report their resolved
versions. Installed SDK BuildTools10.0.26100.7705 is evidence of tools, not an
invented direct PackageReference. If the proposed app needs another direct package
or SDK/TFM change, stop and document it instead of silently changing these pins.

Sources checked2026-11-08: [WindowsAppSDK2.6.5](https://www-1.nuget.org/packages/Microsoft.WindowsAppSDK/2.6.5),
[Test SDK18.11.0](https://www.nuget.org/packages/microsoft.net.test.sdk/18.11.0),
[xunit2.9.3](https://www.nuget.org/packages/xunit/2.9.3),
[adapter3.1.5](https://www.nuget.org/packages/xunit.runner.visualstudio/3.1.5).
xUnit2 is explicitly legacy/security-fixes-only; this proposal deliberately uses
its small VSTest-compatible assertion API, not a claim of latest supported feature
development. Revisit before production reuse. No extra coverage package is requested.
[Self-contained guidance](https://learn.microsoft.com/en-us/windows/apps/package-and-deploy/self-contained-deploy/deploy-self-contained-apps)
requires both Windows App SDK and .NET self-contained settings; native runtime
files stay beside the executable. Do not use APIs requiring additional MSIX packages.

## Execution recipes and allowance

Run PowerShell from the checked-out `clients/windows` directory; resolve these
tool paths anew on another computer. Set only process-local environment values:
DOTNET_CLI_TELEMETRY_OPTOUT=1, DOTNET_NOLOGO=1,
DOTNET_GENERATE_ASPNET_CERTIFICATE=false, DOTNET_ADD_GLOBAL_TOOLS_TO_PATH=false,
DOTNET_CLI_HOME=`$PWD/.dotnet-home`, NUGET_PACKAGES=`$PWD/.nuget/packages`.
Never overwrite HOME/CODEX_HOME or machine/user environment settings.

Post-R1 correction: the original DOTNET_SKIP_FIRST_TIME_EXPERIENCE is unsupported
on modern .NET. The explicit certificate/PATH opt-outs above prevent unwanted
first-use setup; they have NOT yet been exercised here. No retry is authorized
by this correction. See the execution record before any further command.

```powershell
$demoDotnet = 'C:\Program Files\dotnet\dotnet.exe'
$demoMSBuild = 'D:\DeveloperTools\VisualStudio\2026\MSBuild\Current\Bin\MSBuild.exe'
$demoTests = 'Talenro.Demo.Tests/Talenro.Demo.Tests.csproj'
$demoApp = 'Talenro.Demo.App/Talenro.Demo.App.csproj'
# R1: once after test/core project creation; includes Core restore
& $demoDotnet restore $demoTests --configfile NuGet.Config --use-lock-file
# R2: once after app project creation
& $demoMSBuild $demoApp /t:Restore /p:RestoreConfigFile=NuGet.Config /p:RestorePackagesWithLockFile=true /m:1 /nr:false
# L1/L2: once each after inspecting the generated lockfiles
& $demoDotnet restore $demoTests --configfile NuGet.Config --locked-mode
& $demoMSBuild $demoApp /t:Restore /p:RestoreConfigFile=NuGet.Config /p:RestoreLockedMode=true /m:1 /nr:false
# T: substitute only the approved class names listed below
& $demoDotnet test $demoTests --no-restore -c Debug --filter 'FullyQualifiedName~SessionTests'
# G: full suite, no restore
& $demoDotnet test $demoTests --no-restore -c Debug
# B: app compile only, no /restore and no launch
& $demoMSBuild $demoApp /t:Build /p:Configuration=Debug /m:1 /nr:false
```

Total maximum:4 restores (R1,R2,L1,L2),14 filtered T calls (RED/GREEN for seven
classes),3 full G calls (end of Tasks2,3,4),2 B calls (end of Tasks3,4).
No retries beyond these counts. Each restore/build has a15-minute observation
budget; each test call5 minutes. Record start/end, exit code, tests discovered,
failures and consumed allowance. Poll running commands in intervals<=60seconds.
These are stop/escalation budgets, not an OS hard-deadline guarantee: on overrun
stop subsequent work, request cooperative cancellation and verify exit; if unresolved,
report the running session. No arbitrary PID/name kill, borrowed C12 runner, or
repeat-until-pass. Unexpected failure stops execution for diagnosis, not gate removal.

## Task A: Parent Task2, pinned core and behavioral tests

**Files/interfaces:** Parent Task2 owns Core and the five test classes; preserve
`DemoFixtures.Initial(): DemoState` and `DemoReducer.Reduce(DemoState, DemoEvent): DemoState`.

- [x] Create local build configuration, Core and Tests projects; run R1, inspect
  lockfiles, then L1. Package/setup failure is not behavioral RED.
- [x] For each class SessionTests, ConnectionTests, NodeSelectionTests,
  CountryLockTests, EntitlementTests: add the parent's assertions, run T once
  against a compiling minimal implementation to observe assertion RED.
- [x] Implement only that class's required transitions; run its T once for GREEN.
- [x] Run G once; require all discovered tests pass, none skipped. Commit explicit
  configuration/source/test/lockfile paths and a short result/remaining-budget log.

Completion note2026-11-08: G1 passed110/110,0 skipped,exit0 in a verified
non-admin command context. Class cycles and separately approved environment
diagnostic calls are detailed in the execution record; do not infer that aborted
test-host runs counted as behavioral results. Task B remains unstarted.

## Task B: Parent Tasks3–4, native shell and remaining views

**Files/interfaces:** Parent UI ownership unchanged; add
`Talenro.Demo.Tests/ShellContractTests.cs` and `ViewContractTests.cs`.
They inspect XAML/resources as XML and copy fixtures, not launch WinUI; locate
sources through an explicit test-project content link copied to output. No absolute
machine paths. `DemoViewModel.State` remains read-only; `Dispatch(DemoEvent)` is sole mutation entry.

- [ ] ShellContractTests asserts permanent notice/EnterDemo/read-only demo fields,
  nonfocusable decorative glyph and no editable password control. Run T for RED.
- [ ] Implement parent Task3 shell/views/resources; run R2, inspect locks, L2;
  run ShellContractTests T GREEN, G and B once each. Build failure is not UI success.
- [ ] ViewContractTests asserts four real page destinations, Protocol label,
  unknown-value display and excluded-action explanations; run T for RED.
- [ ] Implement parent Task4 nodes/account/dialog bindings. Run ViewContractTests
  T GREEN, G and B once each. Commit explicit source/resources and updated evidence.

## Review focus / tests assigned to owners

1. ConnectionTests: cancel then stale completion preserves Disconnected; logout
   then completion preserves the entire initial state.
2. NodeSelectionTests: connected confirmation Cancel preserves old node/lock;
   Confirm commits both atomically and enters Connecting, never partial state.
3. EntitlementTests: unknown/stale each refuse admission without showing zero;
   entitlement invalidated before confirmation also refuses reconnect.
4. ViewContractTests: each placeholder leaves domain state unchanged except notice;
   resource checks never claim these static checks prove click wiring.
5. Task5 manual checklist: enlarged/RTL text, dialog focus return, keyboard and
   screen-reader names remain explicitly unverified until a native session.

For item4, reducer assertions belong in SessionTests; ViewContractTests covers copy
and routing declarations. Extend the existing class before its RED/GREEN cycle.
Task5 is deferred, not waived: at1920x1200/125%, nominal1536x960 logical display
cannot fit a1440x1024 client area. Record that target as unavailable unless separately
resolved; do not silently change display settings or claim full viewport coverage.

## Self-review and handoff

Author checked spec coverage against parent Tasks2–5, consistent shared interfaces,
all five failure classes, bounded commands and source/lockfile recovery. No task
asserts native runtime/accessibility acceptance from static checks or compilation.
This amendment adds execution inputs, not a second implementation design.
Next: user review/approval for Tasks2–4 including NuGet downloads and test execution;
then inline implementation. Task5/native launch stays separately gated.
