# Windows offline-demo toolchain readiness

Updated2026-11-08. Filename retained from the approved plan's named output.
Status: installed tool metadata VERIFIED; project execution NOT YET QUALIFIED.

## Metadata-only checks

All three commands completed successfully without a project argument:
- Installer `vswhere.exe -products Microsoft.VisualStudio.Product.Community -format json -utf8`:
  Community18.11.3, installationVersion18.11.12328.437, stable/not prerelease,
  isComplete=true, isLaunchable=true, isRebootRequired=false.
- `C:\Program Files\dotnet\dotnet.exe --list-sdks`:10.0.403 at
  `C:\Program Files\dotnet\sdk`.
- `D:\DeveloperTools\VisualStudio\2026\MSBuild\Current\Bin\MSBuild.exe -version -nologo`:
  18.11.1.51303. Distinguish this executable version from catalog package
  Microsoft.Build18.11.1.2651303.

Instance path `D:\DeveloperTools\VisualStudio\2026`; instance IDc3a17ac6.
Installed `state.packages.json` includes SDK BuildTools10.0.26100.7705,
the10.0.403 toolset, Microsoft.Build and WindowsAppSDK.Cs.Dev17 VSIX2.0.260206002.
The VSIX version is NOT an app's Windows App SDK NuGet pin. No project/package
restore has established such a pin yet.

Selected component inventory contains25 Component/Workload IDs. Required C# WinUI,
MSBuild, .NET SDK and Universal workload are present. Optional IntelliCode,
GitHub Copilot and .NET Copilot Skills selection IDs are absent. Installed internal
packages do include Copilot.Contracts, Copilot.Testing.UI and Copilot.Build.Analyzer;
absence of optional selection is not absence of all AI-named files. No removal,
sign-in or AI feature use was performed or inferred.

## Portable configuration and provenance

[Configuration](../toolchains/windows-winui.vsconfig) is reconstructed from actual
selectedPackages, with Product excluded; it is not native-export output.
The documented Installer `export --installPath ... --config ... --quiet` command
failed in restricted execution with exit-532462766. A separately approved normal
environment invocation returned0 but no target file existed; native export remains
unverified. No repair/update/modify fallback was used. Exact cause not diagnosed.

Snapshot source SHA256 values (private raw instance files are not committed):
- state.json:CA81C61D47DD34FDA837B5E02096A337B40401257D6B61C5FED82A5741CC2C81
- state.packages.json:5E29EC95DD504C084DBC30DE0E54273F2A56C504F082EAE313E95221FA4016F0

Configuration must pass JSON/schema-shape, uniqueness, exact selected-ID set and
required/excluded-ID comparisons before checkpoint. This does not prove import
behavior; no new-machine install/import was run. See toolchains README for limits.

## Remaining implementation gate

Host/tool absence is resolved. Next prepare the plan amendment with exact Windows
App SDK/test package versions, TFMs, deployment approach, approved restore sources,
lockfiles and finite restore/build/test/launch commands. Then obtain implementation
approval. Existing design remains native offline-only; no real account/VPN/service,
networking or payments. No client scaffold, restore, build, test, IDE or app launch
occurred in this metadata pass. Task11/12 and44/34/6/217,NOT_ACCEPTED unchanged.
