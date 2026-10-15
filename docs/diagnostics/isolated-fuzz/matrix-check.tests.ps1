param([switch]$RuntimeEvidenceOnly)
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'envelope-uninstrumented.ps1')
$helper=Join-Path $PSScriptRoot 'matrix-check.ps1'
if(Test-Path -LiteralPath $helper){. $helper}
function Check-Matrix([bool]$Value,[string]$Name){if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
Check-Matrix ($null -ne (Get-Command Invoke-LimitedMatrix -ErrorAction SilentlyContinue)) 'bounded matrix entry exists without altering existing drivers'
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
$expectedPackages=@('internal/nodecontrol/nodeissuermapping','internal/nodecontrol/nodeissuerbinding','internal/nodecontrol/nodeleafbinding','internal/nodecontrol/nodecerttime','internal/nodecontrol/nodeissuer','internal/nodecontrol/nodeca','internal/nodecontrol/nodeleaf','internal/nodecontrol/nodeleafkeys','internal/nodecontrol/contracts','internal/nodecontrol/serving','internal/nodecontrol/statebinding','internal/nodecontrol/statecomposition','internal/nodecontrol/statecontract','internal/nodecontrol/stateformat','internal/nodecontrol/statematerials','internal/nodecontrol/statepolicy','internal/nodecontrol/statepolicybinding','internal/nodecontrol/staterecovery','internal/nodecontrol/statetime','internal/nodecontrol/statetimebinding','internal/nodecontrol/stateverify','internal/readiness','internal/nodecontrol/authority','internal/store')
function New-MatrixNativeReceipt {
    $receipt=New-Object 'IsolatedFuzzProcess+Result'
    $fields=@{ExitCode=0;TimedOut=$false;RootExited=$true;CleanupVerified=$true;ActiveAfterCleanup=0;DescendantsRemaining=0;TotalProcesses=3;RootId=17;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;TerminationRequested=$false;ActiveBeforeTermination=0;MembersBeforeTermination=@();MembersAfterCleanup=@();MembersAtRootWait=@();NaturalObservationMs=0;ShutdownElapsedMs=0;ShutdownSchema=3;ShutdownStatePolicy='owned-final-signaled-exit-image-optional-v2';MemberEvidenceTruncated=$false;ShutdownMemberEvidence=@()}
    foreach($name in $fields.Keys){$receipt.$name=$fields[$name]}
    return $receipt
}
function Test-MatrixCase([string]$Scenario){
    $unit=Join-Path $scope ('matrix-unit-'+[guid]::NewGuid().ToString('N'));$root=Join-Path $unit 'matrix-check-901'
    $tool=Join-Path $unit 'uninstrumented-901/toolchain';$modules=Join-Path $unit 'modules'
    foreach($dir in @((Join-Path $tool 'bin'),(Join-Path $modules 'github.com/google/uuid@v1.6.0'))){New-Item -ItemType Directory -Path $dir -Force|Out-Null}
    [IO.File]::WriteAllText((Join-Path $tool 'bin/go.exe'),'never executed');[IO.File]::WriteAllText((Join-Path $modules 'github.com/google/uuid@v1.6.0/go.mod'),'module fixture')
    if($Scenario -eq 'existing'){New-Item -ItemType Directory -Path $root|Out-Null}
    if($Scenario -eq 'wrong-root'){$root=Join-Path $unit 'pair-observe-901'}
    $product=[pscustomobject]@{Path='fixture.go';SHA256=('A'*64)};$source=[pscustomobject]@{Calls=0}
    function Get-EnvelopeProductManifest {[pscustomobject]@{Path=$product.Path;SHA256=$product.SHA256}}
    function Assert-UninstrumentedEnvelopeToolchain([string]$Toolchain,[string]$ReferenceReceipt){$source.Calls++;if($Scenario -eq 'pre-tool' -or ($Scenario -eq 'post-tool' -and $source.Calls -gt 1)){throw 'controlled tool drift'};return (('F'*64)+':'+('E'*64))}
    # Only native execution/full toolchain scanning are replaced; product enumeration
    # is fixture-backed. The wrapper, planner, files, hashes, parsing and gates run.
    $calls=New-Object 'System.Collections.Generic.List[object]'
    function Invoke-MatrixNative([string]$Executable,[string[]]$Arguments,[hashtable]$Environment,[string]$WorkingDirectory,[string]$Stdout,[string]$Stderr,[int]$LimitSeconds){
        $isBuild=$Executable -ceq (Join-Path $tool 'bin/go.exe')
        $checkDir=Split-Path -Parent (Split-Path -Parent $Stdout)
        $index=[int]([IO.Path]::GetFileName($checkDir).Substring(0,2))
        $calls.Add([pscustomobject]@{Build=$isBuild;Index=$index;Executable=$Executable;Arguments=$Arguments;WorkingDirectory=$WorkingDirectory;Limit=$LimitSeconds})
        foreach($key in @('TALENRO_FUZZ_DIAG_MODE','TALENRO_FUZZ_FIXTURE','GODEBUG')){if($Environment[$key] -cne ''){throw 'Diagnostic environment inherited'}}
        if($Environment.GOPROXY -cne 'off' -or $Environment.GOFLAGS -cne '-mod=readonly' -or $Environment.GOCACHE -cne (Join-Path $root 'go-cache')){throw 'Fresh offline environment missing'}
        $r=New-MatrixNativeReceipt;$stdoutText="=== RUN   TestFixture`r`n--- PASS: TestFixture (0.00s)`r`nPASS";$stderrText=''
        if($isBuild){
            if($index -eq 2 -and $Scenario -eq 'runtime-evidence'){[IO.File]::WriteAllText((Join-Path $root '01-nodeissuermapping/execute/stdout.txt'),'changed earlier runtime evidence')}
            if($index -gt 24 -or $LimitSeconds -ne 600 -or $WorkingDirectory -cne $workspace){throw 'Unexpected build count/budget/cwd'}
            $package=$expectedPackages[$index-1];$binary=Join-Path $checkDir 'package.test.exe'
            $expected=@('test','-c','-x','-p=1','-count=1','-timeout=120s',('-o='+$binary),('talenro.local/platform/'+$package))
            if($index -eq 5){$expected=@('test','-c','-x','-p=1','-count=1','-run=^$','-fuzz=^FuzzIssuerEnvelope$','-fuzztime=5s','-parallel=1','-timeout=30s',('-o='+$binary),('talenro.local/platform/'+$package))}
            if(($Arguments -join '|') -cne ($expected -join '|')){throw 'Unauthorized compile recipe'}
            [IO.File]::WriteAllText($binary,'never-executed-'+$index)
            $testFile=@(Get-ChildItem -LiteralPath (Join-Path $workspace $package) -Filter '*_test.go' -File)[0].FullName
            $compiler=Join-Path $tool 'pkg/tool/windows_amd64/compile.exe'
            $stderrText='"'+$compiler+'" -p talenro.local/platform/'+$package+' -pack "'+$testFile+'"'
            if($index -eq 1){$stderrText+="`r`n"+'"'+$compiler+'" -p context -pack "'+(Join-Path $tool 'src/context/context.go')+'"'+"`r`n"+'"'+$compiler+'" -p internal/fuzz -pack "'+(Join-Path $tool 'src/internal/fuzz/fuzz.go')+'"'}
            if($index -eq 5){$stderrText='"'+$compiler+'" -p talenro.local/platform/internal/nodecontrol/nodeissuer -pack "'+(Join-Path $workspace 'internal/nodecontrol/nodeissuer/fuzz_test.go')+'" "'+(Join-Path $workspace 'internal/nodecontrol/nodeissuer/der.go')+'" "'+(Join-Path $workspace 'internal/nodecontrol/nodeissuer/signature.go')+'"'}
            if($index -eq 1){switch($Scenario){'build-exit'{$r.ExitCode=7};'build-gate'{$r.NaturalShutdownVerified=$false};'build-evidence'{$stderrText='missing compile evidence'};'native-throw'{throw 'controlled native exception'}}}
        }else{
            if($index -eq 24){throw 'Store must never execute'}
            $package=if($index -le 23){$expectedPackages[$index-1]}else{'internal/nodecontrol/nodeissuer'}
            if($WorkingDirectory -cne (Join-Path $workspace $package)){throw 'Package test cwd not preserved'}
            if($index -le 23){
                $run=if($index -eq 23){'^(TestBoundRead|TestCoordinator)'}else{'.'}
                if($LimitSeconds -ne 150 -or ($Arguments -join '|') -cne ('-test.count=1|-test.timeout=120s|-test.v=true|-test.run='+$run)){throw 'Ordinary/authority restriction changed'}
                if($index -eq 23){$stdoutText="=== RUN   TestBoundReadFixture`r`n--- PASS: TestBoundReadFixture (0.00s)`r`nPASS"}
                if($index -eq 1){switch($Scenario){'run-exit'{$r.ExitCode=7};'run-timeout'{$r.TimedOut=$true};'no-pass'{$stdoutText='FAIL'};'no-tests'{$stdoutText='PASS'};'post-binary'{[IO.File]::WriteAllText($Executable,'changed')};'post-binding'{[IO.File]::WriteAllText((Join-Path $checkDir 'build/receipt.json'),'{}')}}}
            }else{
                $target=if($index -eq 25){'FuzzIssuerEnvelope'}else{'FuzzIssuerPair'}
                $node=Join-Path $root '05-nodeissuer/package.test.exe'
                if($Executable -cne $node -or $LimitSeconds -ne 60 -or ($Arguments -join '|') -cne ('-test.run=^$|-test.fuzz=^'+$target+'$|-test.fuzztime=5s|-test.parallel=1|-test.timeout=30s|-test.count=1|-test.fuzzcachedir='+(Join-Path $checkDir 'fuzz-cache'))){throw 'Fuzz target/cache/budget/binary changed'}
                $stdoutText='PASS';$stderrText="fuzz: elapsed: 0s, gathering baseline coverage: 4/4 completed, now fuzzing with 1 workers`r`nfuzz: elapsed: 6s, execs: 17 (3/sec), new interesting: 0 (total: 4)"
                if($index -eq 25){switch($Scenario){'fuzz-exit'{$r.ExitCode=7};'fuzz-zero'{$stderrText=$stderrText.Replace('execs: 17 ','execs: 0 ')};'fuzz-diag'{$stderrText+="`r`nTALENRO_FUZZ_DIAG unexpected"}}}
                if($index -eq 26){switch($Scenario){'pair-exit'{$r.ExitCode=7};'post-product'{$product.SHA256='CHANGED'};'persist'{New-Item -ItemType Directory -Path (Join-Path $root 'matrix-result.json')|Out-Null}}}
            }
        }
        [IO.File]::WriteAllText($Stdout,$stdoutText);[IO.File]::WriteAllText($Stderr,$stderrText);return $r
    }
    $seen=New-Object 'System.Collections.Generic.List[object]';$failure=$null
    try{Invoke-LimitedMatrix $root $tool (Join-Path $unit 'reference.json') $modules|ForEach-Object {$seen.Add($_)}}catch{$failure=$_.Exception.Message}
    [pscustomobject]@{Root=$root;Calls=@($calls.ToArray());Seen=@($seen.ToArray());Failure=$failure}
}
if($RuntimeEvidenceOnly){$bad=Test-MatrixCase 'runtime-evidence';Check-Matrix ($null -ne $bad.Failure -and $bad.Calls.Count -eq 49 -and $bad.Seen.Count -eq 0) 'earlier runtime evidence mutation prevents final matrix success';exit}
$good=Test-MatrixCase 'success'
Check-Matrix ($null -eq $good.Failure -and $good.Seen.Count -eq 1 -and $good.Seen[0].MatrixPassed -and $good.Seen[0].Passed -eq 26) 'all26 checks complete before one final success'
Check-Matrix ($good.Calls.Count -eq 49 -and @($good.Calls|Where-Object {$_.Build}).Count -eq 24 -and @($good.Calls|Where-Object {-not $_.Build}).Count -eq 25) '24 builds and25 runs include no store execution or extra query'
$saved=Get-Content -LiteralPath (Join-Path $good.Root 'matrix-result.json') -Raw|ConvertFrom-Json
Check-Matrix ($saved.Unrun -eq 0 -and $saved.Failed -eq 0 -and -not $saved.Acceptance -and $saved.Checks[24].Target -ceq 'FuzzIssuerEnvelope' -and $saved.Checks[25].Target -ceq 'FuzzIssuerPair') 'new matrix labels cannot overwrite historical or production acceptance'
foreach($scenario in @('existing','wrong-root','pre-tool')){$bad=Test-MatrixCase $scenario;Check-Matrix ($null -ne $bad.Failure -and $bad.Calls.Count -eq 0 -and $bad.Seen.Count -eq 0) ($scenario+' refuses before process launch')}
foreach($case in @(@('build-exit',1,0),@('build-gate',1,0),@('build-evidence',1,0),@('native-throw',1,0),@('run-exit',2,0),@('run-timeout',2,0),@('no-pass',2,0),@('no-tests',2,0),@('post-binary',2,0),@('post-binding',2,0),@('fuzz-exit',48,24),@('fuzz-zero',48,24),@('fuzz-diag',48,24),@('pair-exit',49,25),@('post-product',49,26),@('post-tool',49,26),@('runtime-evidence',49,26),@('persist',49,26))){
    $bad=Test-MatrixCase $case[0]
    Check-Matrix ($null -ne $bad.Failure -and $bad.Calls.Count -eq $case[1] -and $bad.Seen.Count -eq 0) ($case[0]+' stops at first failed gate without retry or streamed success')
    $stopped=Get-Content -LiteralPath (Join-Path $bad.Root 'stopped.json') -Raw|ConvertFrom-Json
    Check-Matrix ($stopped.NoRetry -and -not $stopped.Acceptance) ($case[0]+' retains stopped attempt')
    if($case[0] -ne 'persist'){$failed=Get-Content -LiteralPath (Join-Path $bad.Root 'matrix-result.json') -Raw|ConvertFrom-Json;Check-Matrix (-not $failed.MatrixPassed -and -not $failed.Acceptance -and $failed.Passed -eq $case[2]) ($case[0]+' cannot save successful matrix')}
}
