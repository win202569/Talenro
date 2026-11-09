# TextBox shutdown diagnostic prototype

Experimental, non-production reduction for the input-associated shutdown report.
Not a fix and not yet shown to reproduce the stall. Source is retained only for
diagnostic reproducibility; no product integration or promotion is implied.

One WinUI Window with one empty TextBox (`ProbeInput`). No text-change handlers,
filtering, clipboard code, diagnostic observer, shutdown handler, forced exit,
network/account code or reference to Talenro projects. Native framework behavior
is deliberately left intact. The placeholder is not an initial text value.

Uses the existing Windows SDK pin through `../../global.json`, shared build
properties, and the same declared WindowsAppSDK package as the demo. Built outputs
are separate and ignored. It is not added to the application solution or CI.

Build from `clients/windows` in ordinary PowerShell with the pinned .NET SDK and
existing compatible Visual Studio C++/Windows build prerequisites:

```powershell
$env:DOTNET_CLI_HOME = Join-Path (Get-Location) '.dotnet-home'
$env:DOTNET_SKIP_FIRST_TIME_EXPERIENCE = '1'
$env:DOTNET_CLI_TELEMETRY_OPTOUT = '1'
dotnet restore ./diagnostics/TextBoxShutdownProbe/TextBoxShutdownProbe.csproj --locked-mode --packages ./.nuget/packages
dotnet build ./diagnostics/TextBoxShutdownProbe/TextBoxShutdownProbe.csproj --no-restore -c Debug -p:Platform=x64
```

The verified local-cache restore additionally used `--source ./.nuget/packages
-p:NuGetAudit=false --disable-parallel`; this avoided a network restore and was not
a vulnerability audit. A fresh machine needs the pinned dependencies available.

Output directory: `bin/x64/Debug/net10.0-windows10.0.19041.0/win-x64/`.

Restore may need network access on a fresh computer. Do not upgrade dependencies
or change system settings to bypass build failures. The checked-in lockfile binds
NuGet resolution; outputs are newly identified artifacts, not the previous demo.

Current approval covers creation/compilation only: do not launch from this README.
Any future automatic/manual runtime comparison needs its own attempt limit,
observer binding to this executable (not the demo hash/name), no-leftover checks,
normal-close criteria and no-kill rules. The existing demonstration's observations
do not establish this prototype's startup, layout, input or shutdown behavior.
