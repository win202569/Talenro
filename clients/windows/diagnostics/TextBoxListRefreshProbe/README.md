# TextBox / list refresh diagnostic spike

Experimental, build-only checkpoint; not a product fix or confirmed reproducer.
Independent of Talenro.Demo.App and the unchanged TextBoxShutdownProbe baseline.

Both modes use the same window, TextBox, ListView, initial three synthetic rows,
default control resources and normal window lifetime. No argument (or
`--mode=static`) leaves rows untouched after initialization. `--mode=rebuild`
synchronously clears and recreates the same three TextBlocks on TextChanged.
Both attach the same handler; only the guarded rebuild differs. Initialization
events are ignored until both controls and initial rows exist. Unknown/multiple
arguments are rejected before a window is created. Mode is not shown in the UI;
a future observer must record and verify the launch arguments.

This deliberately does NOT filter rows, update selection/header/placeholder,
load demo fixtures/resources, navigate pages, or install a layout observer.
It is not equivalent to the complete NodesPage.Refresh path. No clipboard,
service, settings, shutdown interception or force-exit code is included.

## Build without launching (PowerShell, clients/windows directory)

Prerequisites: the repository-pinned .NET SDK and Windows build dependencies.
On a fresh computer restore dependencies from the configured trusted sources;
the existing-machine offline commands below require a populated .nuget cache.

```powershell
$env:DOTNET_CLI_HOME = Join-Path (Get-Location) '.dotnet-home'
$env:DOTNET_SKIP_FIRST_TIME_EXPERIENCE = '1'
$env:DOTNET_CLI_TELEMETRY_OPTOUT = '1'
dotnet restore ./diagnostics/TextBoxListRefreshProbe/TextBoxListRefreshProbe.csproj --locked-mode --source ./.nuget/packages --packages ./.nuget/packages -p:NuGetAudit=false --disable-parallel
dotnet build ./diagnostics/TextBoxListRefreshProbe/TextBoxListRefreshProbe.csproj --no-restore -c Debug -p:Platform=x64
```

Offline NuGetAudit=false is not a vulnerability audit. Bin/obj, caches and raw
observations are ignored, not portable deliverables. No runtime tests or launches
are authorized by this build checkpoint. A separately approved bounded runtime
comparison is required; previous launch budgets do not transfer to this variant.
All existing NOT_ACCEPTED gates remain.
