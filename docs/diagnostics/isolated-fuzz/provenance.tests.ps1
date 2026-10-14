$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'isolation.ps1')
function Check-Provenance([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
function Reject-Provenance([scriptblock]$Action,[string]$Name) {$caught=$false;try{& $Action|Out-Null}catch{$caught=$true};Check-Provenance $caught $Name}
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$root=Join-Path $workspace ('.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/provenance-'+[guid]::NewGuid().ToString('N'))
$portable=Join-Path $root 'portable';$tool=Join-Path $root 'toolchain'
New-Item -ItemType Directory -Path $portable,$tool | Out-Null
foreach($name in @('provenance.ps1','coordinator.patch','context-diag.go.txt','context-diag_test.go.txt','fuzz-diag.go.txt','fuzz-diag_test.go.txt')) {Copy-Item -LiteralPath (Join-Path $PSScriptRoot $name) -Destination (Join-Path $portable $name)}
$map=@{'context-diag.go.txt'='src/context/diagnostic_fuzz.go';'context-diag_test.go.txt'='src/context/diagnostic_fuzz_test.go';'fuzz-diag.go.txt'='src/internal/fuzz/diagnostic_fuzz.go';'fuzz-diag_test.go.txt'='src/internal/fuzz/diagnostic_fuzz_test.go'}
foreach($name in $map.Keys) {New-Item -ItemType Directory -Path (Split-Path -Parent (Join-Path $tool $map[$name])) -Force|Out-Null;Copy-Item -LiteralPath (Join-Path $portable $name) -Destination (Join-Path $tool $map[$name])}
Set-Content -LiteralPath (Join-Path $tool 'src/context/context.go') -Value 'patched-context'
Set-Content -LiteralPath (Join-Path $tool 'src/internal/fuzz/fuzz.go') -Value 'patched-coordinator'
$hashes=@{};foreach($path in @($map.Values)+@('src/context/context.go','src/internal/fuzz/fuzz.go')) {$hashes[$path]=(Get-FileHash -LiteralPath (Join-Path $tool $path)).Hash}
$record=[pscustomobject]@{Toolchain=$tool;PatchHash=(Get-FileHash -LiteralPath (Join-Path $portable 'coordinator.patch')).Hash;InstalledHashes=[pscustomobject]$hashes}
. (Join-Path $portable 'provenance.ps1')
$identity=Get-DiagnosticSourceIdentity $tool $record
Check-Provenance (-not [string]::IsNullOrWhiteSpace($identity)) 'matched installed and portable source identity accepted'
$template=Join-Path $portable 'fuzz-diag.go.txt';$original=[IO.File]::ReadAllText($template)
Set-Content -LiteralPath $template -Value 'portable-template-drift'
Reject-Provenance {Get-DiagnosticSourceIdentity $tool $record} 'portable template drift cannot reuse old installed helper'
[IO.File]::WriteAllText($template,$original)
$incomplete=$record|ConvertTo-Json -Depth 5|ConvertFrom-Json
$incomplete.InstalledHashes.PSObject.Properties.Remove('src/context/context.go')
Reject-Provenance {Get-DiagnosticSourceIdentity $tool $incomplete} 'missing installed source entry refused'
$extra=$record|ConvertTo-Json -Depth 5|ConvertFrom-Json
$extra.InstalledHashes|Add-Member -NotePropertyName unexpected -NotePropertyValue 'irrelevant'
Reject-Provenance {Get-DiagnosticSourceIdentity $tool $extra} 'extra installed source entry refused'
Assert-DiagnosticBinarySource ([pscustomobject]@{SourceIdentity=$identity}) $identity
Check-Provenance $true 'binary with exact source identity accepted'
Reject-Provenance {Assert-DiagnosticBinarySource ([pscustomobject]@{SourceIdentity='other-source'}) $identity} 'binary from different source identity refused'
Reject-Provenance {Assert-DiagnosticBinarySource ([pscustomobject]@{}) $identity} 'unbound historical binary receipt refused without rebuild'
