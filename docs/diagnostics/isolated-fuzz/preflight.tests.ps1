param([string]$Source='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64')
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'preflight.ps1')
function Check-Preflight([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
function Reject-Preflight([scriptblock]$Action,[string]$Name) {$caught=$false;try{& $Action|Out-Null}catch{$caught=$true};Check-Preflight $caught $Name}
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$root=Join-Path $workspace ('.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/preflight-'+[guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $root | Out-Null
$original=$Source
foreach($path in @('bin/go.exe','VERSION','src/context/context.go','src/internal/fuzz/fuzz.go')) {
    $target=Join-Path $root $path
    New-Item -ItemType Directory -Path (Split-Path -Parent $target) -Force | Out-Null
    Copy-Item -LiteralPath (Join-Path $original $path) -Destination $target
    (Get-Item -LiteralPath $target).IsReadOnly=$false
}
$before=@(Get-TreeManifest $root)
$ready=$null;try{$ready=Test-DiagnosticPrerequisites $root}catch{}
Check-Preflight ($null -ne $ready -and $ready.Ready) 'valid pinned prerequisites accepted without launching Go'
$after=@(Get-TreeManifest $root)
Check-Preflight (@(Compare-Object $before $after -Property Path,SHA256,Length).Count -eq 0 -and $after.Count -eq 4) 'preflight writes no toolchain receipt cache or fixture'
Set-Content -LiteralPath (Join-Path $root 'src/context/context.go') -Value 'drift'
Reject-Preflight {Test-DiagnosticPrerequisites $root} 'drifted source rejected before execution'
Reject-Preflight {Test-DiagnosticPrerequisites (Join-Path $root 'missing')} 'missing toolchain rejected without download'
