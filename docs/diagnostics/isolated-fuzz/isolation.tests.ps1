$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
. (Join-Path $PSScriptRoot 'isolation.ps1')
function Assert-True([bool]$Value,[string]$Name) {
    if (-not $Value) { throw "ASSERTION FAILED: $Name" }
    Write-Output "PASS: $Name"
}
function Assert-Throws([scriptblock]$Action,[string]$Name) {
    $caught=$false
    try { & $Action | Out-Null } catch { $caught=$true }
    Assert-True $caught $Name
}
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$allowed=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
$testRoot=Join-Path $allowed ('tests-'+[guid]::NewGuid().ToString('N'))
$source=Join-Path $testRoot 'source'
New-Item -ItemType Directory -Path $source -Force | Out-Null
[IO.File]::WriteAllText((Join-Path $source 'one.txt'),'source-one')
New-Item -ItemType Directory -Path (Join-Path $source 'nested') | Out-Null
[IO.File]::WriteAllText((Join-Path $source 'nested/two.txt'),'source-two')
$target=Join-Path $testRoot 'copy'
try {
    Assert-IsolatedPaths $source $target $allowed
    Assert-True $true 'valid independent target accepted'
} catch { throw "ASSERTION FAILED: valid independent target was rejected: $_" }
Assert-Throws { Assert-IsolatedPaths $source $source $allowed } 'same path rejected'
Assert-Throws { Assert-IsolatedPaths $source (Join-Path $source 'inside') $allowed } 'descendant rejected'
Assert-Throws { Assert-IsolatedPaths $source $testRoot $allowed } 'ancestor rejected'
Assert-Throws { Assert-IsolatedPaths $source ($allowed+'-outside/copy') $allowed } 'prefix-confusable outside rejected'
Assert-True (-not (Test-Path -LiteralPath $target)) 'validation writes nothing'
$before=@(Get-TreeManifest $source)
$copy=New-IsolatedToolchain $source $target
Assert-True (Test-Path -LiteralPath (Join-Path $target 'copy-receipt.json')) 'fresh verified copy saves required recovery receipt'
$saved=Get-Content -Raw -LiteralPath (Join-Path $target 'copy-receipt.json')|ConvertFrom-Json
Assert-True ($saved.Toolchain -eq $copy.Toolchain -and $saved.BeforeManifest.Count -eq 2 -and @((Compare-Object $saved.CopyManifest $copy.CopyManifest -Property Path,SHA256,Length)).Count -eq 0) 'saved receipt preserves copy identity and verified manifests'
Assert-True ($before.Count -eq 2) 'manifest contains exact two files'
Assert-True (($before | ConvertTo-Json -Compress) -ceq ($copy.CopyManifest | ConvertTo-Json -Compress)) 'copy manifest equality'
Assert-True (($before | ConvertTo-Json -Compress) -ceq (@(Get-TreeManifest $source) | ConvertTo-Json -Compress)) 'source unchanged'
[IO.File]::WriteAllText((Join-Path $copy.Toolchain 'one.txt'),'target-only')
Assert-True ([IO.File]::ReadAllText((Join-Path $source 'one.txt')) -ceq 'source-one') 'copied bytes independent'
Assert-Throws { New-IsolatedToolchain $source $target } 'existing destination rejected'
Assert-Throws { Assert-IsolatedPaths $source (Join-Path $testRoot 'copy/file:stream') $allowed } 'alternate stream path rejected'
Assert-Throws { Get-TreeManifest (Join-Path $testRoot 'missing') } 'missing source rejected'
$linkSource=Join-Path $testRoot 'link-source'
New-Item -ItemType Directory -Path $linkSource | Out-Null
$junction=Join-Path $linkSource 'junction'
$created=$false
try { New-Item -ItemType Junction -Path $junction -Target $source | Out-Null; $created=$true } catch { Write-Output 'UNRUN: real junction creation unavailable' }
if($created) {
    Assert-Throws { Get-TreeManifest $linkSource } 'real junction rejected'
    Assert-Throws { Assert-IsolatedPaths $source (Join-Path $junction 'copy') $allowed } 'destination ancestor junction rejected'
}
$hardSource=Join-Path $testRoot 'hard-source'
New-Item -ItemType Directory -Path $hardSource | Out-Null
$created=$false
try { New-Item -ItemType HardLink -Path (Join-Path $hardSource 'alias.txt') -Target (Join-Path $source 'one.txt') | Out-Null; $created=$true } catch { Write-Output 'UNRUN: real hardlink creation unavailable' }
if($created) { Assert-Throws { Get-TreeManifest $hardSource } 'real multiple hardlink rejected' }
Write-Output "Fixtures retained: $testRoot"
