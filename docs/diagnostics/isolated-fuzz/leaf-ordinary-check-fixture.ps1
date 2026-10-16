# Fake-only utilities. No workload launches at import or from fixture executables.
function Check-LeafOrdinary([bool]$Value,[string]$Name){if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
function Reject-LeafOrdinary([scriptblock]$Action,[string]$Name){$failed=$false;try{& $Action|Out-Null}catch{$failed=$true};Check-LeafOrdinary $failed $Name}
function New-LeafOrdinaryTestFiles([string]$FixtureRoot){
    $tool=Join-Path $FixtureRoot 'uninstrumented-901/toolchain';$modules=Join-Path $FixtureRoot 'modules'
    foreach($dir in @((Join-Path $tool 'bin'),(Join-Path $modules 'github.com/google/uuid@v1.6.0'))){New-Item -ItemType Directory -Path $dir -Force|Out-Null}
    foreach($path in @((Join-Path $tool 'bin/go.exe'),(Join-Path $FixtureRoot 'uninstrumented-901/copy-receipt.json'),(Join-Path $FixtureRoot 'reference.json'),(Join-Path $modules 'github.com/google/uuid@v1.6.0/go.mod'))){[IO.File]::WriteAllText($path,'fixture data; never executed')}
    [pscustomobject]@{Root=(Join-Path $FixtureRoot 'leaf-ordinary-check-901');Toolchain=$tool;ReferenceReceipt=(Join-Path $FixtureRoot 'reference.json');ModuleCache=$modules}
}
