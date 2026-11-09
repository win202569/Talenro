# Close-stage recording (build-only handoff)

Opt-in diagnostic instrumentation, not a shutdown fix. Default off; Debug plus
layout diagnostics required. No new packages. UI launch budget for implementation: 0.

## Reproduce builds (PowerShell 7, clients/windows)

Use the pinned SDK and existing Windows SDK/WinUI build prerequisites. A fresh
computer must prepare dependencies separately; ignored caches/binaries are not
restored by Git. With the existing local package cache:

```powershell
$env:DOTNET_CLI_HOME = Join-Path (Get-Location) '.dotnet-home'
$env:DOTNET_SKIP_FIRST_TIME_EXPERIENCE = '1'
$env:DOTNET_CLI_TELEMETRY_OPTOUT = '1'
$project = './Talenro.Demo.App/Talenro.Demo.App.csproj'
foreach ($mode in @('ordinary','layout','trace')) {
    $layout = if ($mode -eq 'ordinary') { 'false' } else { 'true' }
    $trace = if ($mode -eq 'trace') { 'true' } else { 'false' }
    $flags = @('-p:Configuration=Debug','-p:Platform=x64',
        "-p:EnableLayoutDiagnostics=$layout", "-p:EnableCloseStageRecording=$trace")
    if ($mode -ne 'trace') { $flags += "-p:CloseStageVerificationMode=$mode" }
    dotnet restore $project --locked-mode --source ./.nuget/packages --packages ./.nuget/packages -p:NuGetAudit=false --disable-parallel @flags
    if ($LASTEXITCODE -ne 0) { throw "Restore failed: $mode" }
    dotnet build $project --no-restore @flags
    if ($LASTEXITCODE -ne 0) { throw "Build failed: $mode" }
}
./tests/close-stage-build-gates.tests.ps1
$suffix = 'x64/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.dll'
./tests/verify-close-stage-artifacts.ps1 -OrdinaryAssembly "Talenro.Demo.App/bin/close-stage-verification/ordinary/$suffix" -LayoutAssembly "Talenro.Demo.App/bin/close-stage-verification/layout/$suffix" -CloseStageAssembly "Talenro.Demo.App/bin/close-stage-recording/$suffix"
dotnet test ./Talenro.Demo.Tests/Talenro.Demo.Tests.csproj --no-restore
```

Tests need their own locked restore on a fresh machine. Audit disabled above is
an offline constraint, not evidence of a successful security audit. Do not pass
global relative MSBuildProjectExtensionsPath or DefaultItemExcludes overrides:
the former loses NuGet imports, and relocating Core exposes stale obj sources.
The App-only verification mode keeps absolute import paths and Core unchanged.

## Activation and interpretation (not authorization to launch)

Trace builds require `--layout-test` and exactly one
`--layout-close-run=<fresh lowercase nonzero Guid in N form>` plus valid layout
arguments. Ready must be flushed within 2 seconds before MainWindow creation.
Output is `<AppContext.BaseDirectory>/close-stage-records/<run>.jsonl`, CreateNew
only; detected reparse ancestors are rejected. This is not protection against a
hostile concurrent filesystem replacement. Do not commit raw receipts.

The UI producer writes only fixed stage/sequence/timestamp records to 64 slots.
No close-time disk operation, flush or join. The background writer may block or
lose pending tail records when the process exits. Schema/run/frequency are in the
ready header; final health reports drop/invalid-producer/failure state.

Read using CloseStageReceiptReader with the expected run identifier. ValidPrefix
proves only published stages; a missing after marker does not prove a stall.
CompletedClose requires the full valid sequence and final healthy record, and
means handler completion only, NOT process exit or acceptance. Invalid/truncated/
cross-run/dropped traces remain Inconclusive. Pure tests do not validate native
event unsubscription; future live observation needs separate bounded approval.
