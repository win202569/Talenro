$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
. (Join-Path $PSScriptRoot 'isolation.ps1')
function Assert-True([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
function Assert-Throws([scriptblock]$Action,[string]$Name) {$caught=$false;try{& $Action|Out-Null}catch{$caught=$true};Assert-True $caught $Name}
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$root=Join-Path $workspace ('.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/materialize-'+[guid]::NewGuid().ToString('N'))
$fixed='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64'
function New-TestTree([string]$Name) {
    $tree=Join-Path $root $Name
    foreach($path in @('src/context/context.go','src/internal/fuzz/fuzz.go')) {
        New-Item -ItemType Directory -Path (Split-Path -Parent (Join-Path $tree $path)) -Force | Out-Null
        Copy-Item -LiteralPath (Join-Path $fixed $path) -Destination (Join-Path $tree $path)
        [IO.File]::SetAttributes((Join-Path $tree $path),[IO.FileAttributes]::Normal)
    }
    return $tree
}
$tree=New-TestTree 'positive'
Install-DiagnosticPatch $tree
Assert-True (Test-Path -LiteralPath (Join-Path $tree 'src/context/diagnostic_fuzz.go')) 'context helper installed'
Assert-True (Test-Path -LiteralPath (Join-Path $tree 'src/internal/fuzz/diagnostic_fuzz.go')) 'trace helper installed'
Assert-True ((Get-FileHash -LiteralPath (Join-Path $tree 'src/internal/fuzz/diagnostic_fuzz.go')).Hash -eq (Get-FileHash -LiteralPath (Join-Path $PSScriptRoot 'fuzz-diag.go.txt')).Hash) 'exact trace helper bytes'
$relative=$tree.Substring($workspace.Length+1).Replace('\','/')
git -c core.fsmonitor=false -C $workspace apply --check --reverse --unidiff-zero --ignore-space-change --directory=$relative (Join-Path $PSScriptRoot 'coordinator.patch')
Assert-True ($LASTEXITCODE -eq 0) 'actual source patch matches replay'
Assert-Throws {Install-DiagnosticPatch $tree} 'duplicate patch refused'
Assert-Throws {Install-DiagnosticPatch $fixed} 'official toolchain refused'
$drift=New-TestTree 'drift'
$before=(Get-FileHash -LiteralPath (Join-Path $drift 'src/context/context.go')).Hash
[IO.File]::WriteAllText((Join-Path $drift 'src/internal/fuzz/fuzz.go'),'drift')
Assert-Throws {Install-DiagnosticPatch $drift} 'source drift rejected'
Assert-True ($before -eq (Get-FileHash -LiteralPath (Join-Path $drift 'src/context/context.go')).Hash) 'all source guards precede mutation'
Write-Output "Fixtures retained: $root"
