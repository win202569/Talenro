# Test-only factories. Importing this file creates no files or processes.
function Check-BuildOnly([bool]$Value,[string]$Name) {
    if(-not $Value){throw ('ASSERTION FAILED: '+$Name)}
    Write-Output ('PASS: '+$Name)
}
function New-BuildOnlyTestFiles([string]$FixtureRoot) {
    $tool=Join-Path $FixtureRoot 'uninstrumented-901/toolchain'
    $modules=Join-Path $FixtureRoot 'modules'
    foreach($dir in @((Join-Path $tool 'bin'),(Join-Path $modules 'github.com/google/uuid@v1.6.0'))){New-Item -ItemType Directory -Path $dir -Force -ErrorAction Stop|Out-Null}
    [IO.File]::WriteAllText((Join-Path $tool 'bin/go.exe'),'fake file; never executed')
    [IO.File]::WriteAllText((Join-Path $modules 'github.com/google/uuid@v1.6.0/go.mod'),'module fixture')
    $reference=Join-Path $FixtureRoot 'reference.json';[IO.File]::WriteAllText($reference,'{}')
    [pscustomobject]@{Root=(Join-Path $FixtureRoot 'leaf-stage-build-only-901');Toolchain=$tool;ReferenceReceipt=$reference;ModuleCache=$modules}
}
