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

Proposal ready for route selection. Zero installer downloads/launches, installs,
SDK commands, builds or tests this turn. Only web documentation, path/free-space
metadata and repository documents were inspected. Task11/12 gates unchanged;
44/34/6/217,NOT_ACCEPTED. Next decision: approve the Visual Studio route through
download/selection review, with installation only after exact choices and terms
are shown, or supply an existing custom toolchain path.
