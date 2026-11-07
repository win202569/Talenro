# Windows client developer tools — installation proposal only

Date:2026-11-07. User authorized preparation of this proposal, not downloads,
installation, account creation, security-setting changes, build or launch.
Scope: prerequisites for the [offline WinUI demo plan](../superpowers/plans/2026-11-07-windows-client-offline-demo.md).

## Recommended route

Use the stable Visual Studio 2026 installer with **WinUI application development**,
selecting C# support and its required managed desktop/.NET, MSBuild and Windows SDK
dependencies. Prefer Community only if the user confirms eligibility and accepts
its actual terms; otherwise use an existing properly licensed edition. This note
does not determine licensing eligibility or promise account-free long-term use.
The user previously lacked a Microsoft account: do not create one or bypass a
sign-in/license requirement. Stop for their decision if one prevents progress.

Microsoft's [WinUI quickstart](https://learn.microsoft.com/en-us/windows/apps/get-started/start-here)
offers both Visual Studio and winapp CLI. Its CLI workflow uses .NET10+ and
winapp0.6+, currently public preview. The IDE route is recommended here to avoid
adding a preview orchestration tool to this first native demo. This is a choice,
not a claim that CLI cannot work. Do not install both routes automatically.

## Components and sources

|Item|Reason / proposed scope|Official source|
|---|---|---|
|Visual Studio 2026 stable + Installer|C# WinUI development and debug tools; no preview channel|[Downloads](https://visualstudio.microsoft.com/downloads/)|
|WinUI application development workload|Only C# support and required dependencies; omit optional C++, mobile, web, Azure, game, AI assistants and unrelated extensions unless a documented required dependency exists|[WinUI setup](https://learn.microsoft.com/en-us/windows/apps/get-started/start-here)|
|.NET SDK and Windows SDK|Use compatible versions actually selected by the stable workload; do not double-install an independent SDK by guess|[Install and select features](https://learn.microsoft.com/en-us/visualstudio/install/install-visual-studio)|
|Windows App SDK NuGet package|Later project-specific stable package, separately pinned/restored; SDK install is not the same as a project restore|[Windows App SDK](https://learn.microsoft.com/en-us/windows/apps/windows-app-sdk/)|

Before installation, capture exact installer/channel/product version, publisher,
component list, disk estimates and intended paths from the actual installer.
Verify valid Microsoft signature and record SHA256; a self-computed hash alone
does not authenticate an arbitrary download. Use Microsoft's linked distribution
endpoints only; stop on an unverified publisher, unexpected redirect or alert.
No exact patch/component ID is invented from an uninspected installer catalog.

## Local observations and disk plan

Prior bounded inventory found no dotnet/MSBuild on PATH or standard .NET/Visual
Studio directories; Windows Kits/10 exists, completeness unknown. Custom paths
may exist. No installed-package or security configuration query was performed.
This turn found winget by command lookup, but did not invoke it.

Free-space snapshot: C:68,970,418,176 bytes (about64.2GiB),
D:508,217,958,400 bytes (about473.3GiB). This does not certify RAM, CPU, OS edition,
architecture or full compatibility; those must be checked before installing.
[Microsoft requirements](https://learn.microsoft.com/en-us/visualstudio/releases/2026/vs-system-requirements)
describe typical20–50GB installations, with actual space dependent on components.
Do not present that range as this workload's measured size.

Proposed locations: IDE `D:\DeveloperTools\VisualStudio\2026`, download cache
`D:\DeveloperTools\VisualStudio\Packages`; keep existing shared-component paths
if discovered. These are outside the repository and require installation approval.
Do not relocate existing Windows Kits or .NET files by hand. Some tools remain
on C: even with D: selected, as explained in Microsoft's
[location guidance](https://learn.microsoft.com/en-us/visualstudio/install/change-installation-locations).
Author guardrail: require installer-estimated C: use plus20GiB headroom to fit
current C: free space; otherwise stop and revise, never delete user files.

## System impact and explicit exclusions

Installation downloads from Microsoft, writes machine-wide tools, registration,
shared runtimes/caches and possibly environment integration; initial setup requires
administrator permission per Microsoft's installation guide. The installer may
request reboot. No automatic reboot; pause and report. Run normal development
later without elevation. Approval of installation does not approve admin app runs.

The WinUI quickstart includes Developer Mode for its deployment workflow. **Do
not enable it as part of this proposal's installation scope.** Evaluate the exact
packaged/unpackaged launch requirements separately. No broad `winget configure`
recipe: it can bundle system-setting changes. No execution-policy changes,
Defender exclusions, quarantine restore, firewall edits, certificates, Windows
Update installation or account sign-in on the user's behalf.

## Finite execution sequence, only after explicit approval

1. Confirm edition/license choice and existing custom installation paths, supported
   OS/architecture and free space. If unavailable, report the prerequisite rather
   than trying alternate installers repeatedly.
2. Download and validate the official bootstrapper; open Installer to inspect its
   stable version, selected components and C:/D: estimates. This already writes
   installer/cache state and is not a read-only operation.
3. Present the exact selection and any new terms/sign-in/UAC prompt to the user
   before starting the workload installation. Do not accept agreements for them.
4. Install the approved selection once; use the installer's supported progress
   and cancellation, not a hard process kill or repeated repair attempts. No silent
   optional-feature expansion, concurrent installer or automatic restart.
5. After success, record installed versions/paths and export a portable workload
   configuration without secrets. With a separately specified check allowance,
   query SDK/MSBuild metadata only; do not create/build/run a project here.
6. Update readiness and the implementation plan with actual SDK/package choices.
   Project restore, compilation, tests and demo launch remain separately scoped.

Unexpected security alert, publisher mismatch, unsupported OS, insufficient disk,
unaccepted terms, unexpected components, install failure or reboot requirement
stops continuation. Preserve errors; no security/permission workaround.

## Removal and recovery

Use Visual Studio Installer's Modify to remove this workload, or More→Uninstall
for this specific instance, following [Microsoft removal guidance](https://learn.microsoft.com/en-us/visualstudio/install/uninstall-visual-studio).
Do not recursively delete DeveloperTools, Windows Kits, dotnet or shared caches.
Shared runtimes/Installer/cache may remain; uninstall is not a byte-perfect rollback.
Do not run InstallCleanup automatically: it can affect other instances/components.
No removal is authorized by this document. Keep source, workload configuration,
pins and this report in Git, never SDK executables or installation caches.

## Current result

### Installation completed after subsequent approval — 2026-11-08

See [installation result](2026-11-08-windows-client-tool-installation-result.md).
User confirmed Community eligibility, reported manually changing scaling from150%
to125%, then explicitly approved the reviewed installation. Installer now shows
Community2026 18.11.3 installed. No SDK/MSBuild commands, IDE/project launch or tests
were performed. Historical pending-install sections below are retained as history.

### Dependency and host review — 2026-11-07

Resolved the prior SDK-selection question without installing or running any SDK.
Read local Installer catalog at
`C:\Users\Lenovo\AppData\Local\Microsoft\VisualStudio\Packages\_Channels\433fad92\catalog.json`.
Catalog SHA256: `417B8D3FEE577BDB0E192835692446DE51B4626526E4B555162FD9C2FCA7DFF1`.
Identity: VisualStudio/18.11.3+12328.437, stable, RTW, not prerelease.
Bootstrapper metadata identifies Community / VisualStudio.18.Release.

Required dependency edges observed in that catalog:
`Microsoft.VisualStudio.Workload.Universal`
→ `Microsoft.VisualStudio.ComponentGroup.WindowsAppDevelopment.Prerequisites`
→ `Microsoft.VisualStudio.Component.WindowsAppSdkSupport.CSharp`
→ `Microsoft.VisualStudio.PackageGroup.MSIX.Shared`
→ `Microsoft.Windows.SDK.BuildTools_10.0.26100.7705` (Nupkg).
This is bundled build tooling, not the full standalone Windows SDK. The three
standalone Windows11 SDK components are explicitly Optional. Thus their unchecked
boxes alone do NOT prove missing C# WinUI prerequisites. Recommend keeping the
reviewed C# selection rather than adding a full SDK without a concrete need.
Later project targeting/package restore and build must still be verified.

Catalog .NET SDK dependency identifies toolset10.0.403-servicing.26523.112;
component version18.11.12324.64 is not the .NET CLI version. MSBuild component
18.11.12301.404 resolves Microsoft.Build package18.11.1.2651303. These are proposed
catalog versions, not installed-command results. C# tooling also depends on
HostX64.TargetX64/TargetX86 native tools internally; excluding optional C++ workload
does not imply no native dependencies. Preserve required dependencies.

Official cross-checks (documentation has older package revisions than local catalog):
- [VS component directory](https://learn.microsoft.com/en-us/visualstudio/install/workload-component-id-vs-community): WinUI workload requires MSBuild/.NET and lists standalone Windows SDK as optional; do not copy its older versions into current selection.
- [WinUI quickstart](https://learn.microsoft.com/en-us/windows/apps/get-started/start-here): VS2026 WinUI route; no broad WinGet configuration executed.
- [VS requirements](https://learn.microsoft.com/en-us/visualstudio/releases/2026/vs-system-requirements) and [Windows lifecycle](https://learn.microsoft.com/en-us/lifecycle/products/windows-11-home-and-pro).

Read-only CIM queries initially unavailable in restricted execution; approved
read-only execution returned Windows11 Pro64-bit, build26200 (25H2 family), x64,
Intel Core7 240H10 cores, physical RAM16,699,781,120 bytes, display1920x1200 pixels.
OS family/architecture, CPU and RAM meet documented baseline. Physical display
exceeds minimum, but effective logical resolution after scaling remains unverified;
do not claim complete display compatibility or change scaling automatically.
Free bytes C:67,922,391,040, D:508,217,372,672. Previous14.39GB selection plus20GiB
headroom fits even if charged entirely to C. No UI changes this turn; previous
selection estimate must be refreshed immediately before any approved install.

Next: user confirms Community eligibility/terms and effective display suitability,
then grants explicit installation approval. Refresh selection and stop on drift.
Install only reviewed WinUI required set, retain excluded AI/optional C++ choices,
product/cache D paths and unchanged shared path. No automatic reboot, sign-in,
Developer Mode/security changes, IDE/project launch or tests. Metadata verification
after installation should be separately stated; actual toolchain readiness still
requires installed versions and later bounded project checks. Task11/12 gates unchanged.

### Subsequent selection review — workload installation not started

After launch approval, bootstrapper was opened and stopped at the license screen.
User then reported manually reaching selection. Observed Installer executable is
now under `C:\Program Files (x86)\Microsoft Visual Studio\Installer\setup.exe`.
This supersedes the earlier not-launched status, not the no-workload-install gate.

Observed product: Visual Studio Community2026, October2026 feature update18.11.3.
Initial core-only selection estimated1.37GB. Selecting WinUI application
development estimated15.69GB. Included required group lists .NET WinUI application
development prerequisites, .NET Native/.NET Standard, NuGet package manager and
Blend for Visual Studio. Deselected IntelliCode (confirmed removal of dependent
GitHub Copilot from the pending selection) and .NET Copilot Skills; three AI boxes
visually unchecked. Optional C++/UWP/DirectX items were not selected.

Current provisional estimate14.39GB; location breakdown observed:
- Product2.97GB: `D:\DeveloperTools\VisualStudio\2026` entered and read back.
- Download cache3.44GB: `D:\DeveloperTools\VisualStudio\Packages` entered and read back;
  retain-download-cache option remains selected.
- Shared components66MB: unchanged `C:\Program Files (x86)\Microsoft Visual Studio\Shared`.
- System cache/tools/fixed-location SDKs7.92GB.

These are rounded UI values, not measured installed usage. Current free bytes:
C:67,921,747,968; D:508,217,819,136. Even conservatively allocating all14.39GiB
plus20GiB headroom to C fits this snapshot, but additions require recalculation.

Individual-components view visually showed .NET SDK and .NET Framework4.8 SDK
selected, and .NET10/.NET8 runtimes selected. It did not expose exact .NET SDK or
MSBuild patch versions. Windows11 SDK10.0.22621.0,10.0.26100.9457 and10.0.28000.2705
boxes were all unchecked. Do not infer that merely listed SDKs are included.
The contents of the required WinUI prerequisite group and whether a standalone
Windows SDK must be added remain to be resolved. A subsequent Windows10 SDK filter
showed stale/disconnected list items, so it supplies no reliable Windows10 inventory.

Stop before installation: final dependency selection, OS/architecture/RAM support
and Community license eligibility remain unverified. Current selection is NOT a
complete accepted toolchain configuration. No Install action, developer-mode or
security change, project build/test/launch occurred. Preserve existing project gates.

### Subsequent approved download record — 2026-11-07

User subsequently approved official download/verification/selection inspection,
not workload installation. Download completed once; bootstrapper not launched.
The official thank-you page's `c2r_downloadResult.url` supplied this endpoint:
https://c2rsetup.officeapps.live.com/c2r/downloadVS.aspx?sku=community&channel=stable&version=VS18&source=VSLandingPage&cid=2500

- Source page: https://visualstudio.microsoft.com/thank-you-downloading-visual-studio/?channel=Stable&cid=2500&passive=false&sku=Community&source=VSLandingPage&version=VS18
- File description: Visual Studio Installer; product version: Visual Studio 2026.
- Length: 5,690,136 bytes.
- Authenticode: Valid; signer CN/O Microsoft Corporation, Redmond, Washington, US.
- SHA256: `4FF68902AAC5BA8890B062919D5651B1ABC80E5FB40796B6B3F6A4F4E9ECD180`.
- Local-only file: `C:\Users\Lenovo\AppData\Local\Temp\Talenro-VS-review-f52407fed809427f93d59b5a606ed754\vs_community.exe`.

Binary intentionally not committed. Initial restricted-network DNS failed;
approved network access retrieved the official page and file. No alternate
unofficial source used. Computer-use action-time confirmation is required before
first launch of the downloaded program. No terms accepted, UAC acted on, workload
installed, security setting changed or project built/tested. Exact workload
versions/selection/disk costs and machine compatibility remain unverified.
Task11/12 gates and44/34/6/217,NOT_ACCEPTED remain unchanged.

### Historical proposal-only result

Proposal ready for route selection. Zero installer downloads/launches, installs,
SDK commands, builds or tests this turn. Only web documentation, path/free-space
metadata and repository documents were inspected. Task11/12 gates unchanged;
44/34/6/217,NOT_ACCEPTED. Next decision: approve the Visual Studio route through
download/selection review, with installation only after exact choices and terms
are shown, or supply an existing custom toolchain path.
