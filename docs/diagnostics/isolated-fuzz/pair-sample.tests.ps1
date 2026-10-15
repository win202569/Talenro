$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'envelope-uninstrumented.ps1')
$helper=Join-Path $PSScriptRoot 'pair-sample.ps1'
if(Test-Path -LiteralPath $helper){. $helper}
function Check-Pair([bool]$Value,[string]$Name){if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
Check-Pair ($null -ne (Get-Command Invoke-PairCompiledSample -ErrorAction SilentlyContinue)) 'Pair entry exists without changing Envelope entry'
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
function Save-PairFixture([string]$Path,[object]$Value){$Value|ConvertTo-Json -Depth 9|Set-Content -LiteralPath $Path}
function New-PairReceipt([string]$Executable,[string[]]$Arguments,[hashtable]$Environment,[string]$Directory){
    [pscustomobject]@{ExitCode=0;TimedOut=$false;RootExited=$true;CleanupVerified=$true;ActiveAfterCleanup=0;DescendantsRemaining=0;TotalProcesses=3;RootId=17;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;TerminationRequested=$false;ActiveBeforeTermination=0;MembersBeforeTermination=@();MembersAfterCleanup=@();MembersAtRootWait=@();NaturalObservationMs=0;ShutdownElapsedMs=0;ShutdownSchema=3;ShutdownStatePolicy='owned-final-signaled-exit-image-optional-v2';MemberEvidenceTruncated=$false;ShutdownMemberEvidence=@();Executable=$Executable;Arguments=$Arguments;Environment=$Environment;StdoutPath=(Join-Path $Directory 'stdout.txt');StderrPath=(Join-Path $Directory 'stderr.txt')}
}
function Test-PairCase([string]$Scenario){
    $unit=Join-Path $scope ('pair-unit-'+[guid]::NewGuid().ToString('N'))
    $compiled=Join-Path $unit 'envelope-observe-901';$target=Join-Path $unit 'pair-observe-901'
    $tool=Join-Path $unit 'uninstrumented-901/toolchain';$modules=Join-Path $unit 'modules'
    foreach($directory in @($compiled,(Join-Path $compiled 'build'),(Join-Path $tool 'bin'),(Join-Path $modules 'github.com/google/uuid@v1.6.0'))){New-Item -ItemType Directory -Path $directory -Force|Out-Null}
    [IO.File]::WriteAllText((Join-Path $tool 'bin/go.exe'),'never executed')
    [IO.File]::WriteAllText((Join-Path $modules 'github.com/google/uuid@v1.6.0/go.mod'),'module fixture')
    $old=Get-EnvelopeObservationPlan (Join-Path $unit 'envelope-observe-902') $tool $modules
    # Build an independent prior Envelope plan for an already existing compile root.
    $old.RunRoot=$compiled;$old.Binary=Join-Path $compiled 'nodeissuer.test.exe';$old.Mode='uninstrumented'
    $old.BuildArguments=@('test','-c','-x','-p=1','-count=1','-run=^$','-fuzz=^FuzzIssuerEnvelope$','-fuzztime=5s','-parallel=1','-timeout=30s',('-o='+$old.Binary),'talenro.local/platform/internal/nodecontrol/nodeissuer')
    foreach($name in @('GOCACHE','TEMP','TMP','GOTMPDIR')){$old.Environment[$name]=Join-Path $compiled $(if($name -eq 'GOCACHE'){'go-cache'}else{'tmp'})}
    foreach($name in @('TALENRO_FUZZ_DIAG_MODE','TALENRO_FUZZ_FIXTURE','GODEBUG')){$old.Environment[$name]=''}
    [IO.File]::WriteAllText($old.Binary,'never-executed-unit-binary')
    $hash=(Get-FileHash -LiteralPath $old.Binary).Hash;$identity=('F'*64)+':'+('E'*64)
    $product=[pscustomobject]@{Path='fixture.go';SHA256=('A'*64)}
    function Get-EnvelopeProductManifest {[pscustomobject]@{Path=$product.Path;SHA256=$product.SHA256}}
    $sourceChecks=[pscustomobject]@{Count=0}
    # Full toolchain scanning and native execution are replaced; product enumeration
    # is fixture-backed. JSON, hashes, compiler parsing, manifest comparisons,
    # planner, completion and finalization stay real. No machine cache is required.
    function Assert-UninstrumentedEnvelopeToolchain([string]$Toolchain,[string]$ReferenceReceipt){$sourceChecks.Count++;if($Scenario -eq 'post-tool' -and $sourceChecks.Count -gt 1){throw 'controlled tool drift'};return $identity}
    $compiler=Join-Path $tool 'pkg/tool/windows_amd64/compile.exe'
    $log='"'+$compiler+'" -p context -pack "'+(Join-Path $tool 'src/context/context.go')+'"'+"`r`n"+'"'+$compiler+'" -p internal/fuzz -pack "'+(Join-Path $tool 'src/internal/fuzz/fuzz.go')+'"'+"`r`n"+'"'+$compiler+'" -p talenro.local/platform/internal/nodecontrol/nodeissuer -pack "'+(Join-Path $workspace 'internal/nodecontrol/nodeissuer/fuzz_test.go')+'" "'+(Join-Path $workspace 'internal/nodecontrol/nodeissuer/der.go')+'" "'+(Join-Path $workspace 'internal/nodecontrol/nodeissuer/signature.go')+'"'
    [IO.File]::WriteAllText((Join-Path $compiled 'build/stderr.txt'),$log)
    $build=New-PairReceipt $old.Executable $old.BuildArguments $old.Environment (Join-Path $compiled 'build')
    if($Scenario -eq 'build-args'){$build.Arguments=@('test','different')};if($Scenario -eq 'build-env'){$build.Environment.GOPROXY='https://invalid.example'}
    Save-PairFixture (Join-Path $compiled 'build/receipt.json') $build
    Save-PairFixture (Join-Path $compiled 'plan.json') $old
    $prior=[pscustomobject]@{Uninstrumented=$true;SamplingExecuted=$true;ObservationValidationPassed=$true;Acceptance=$false;ToolchainReference=$identity;BinaryHash=$hash;Receipt=$build}
    if($Scenario -eq 'prior-failed'){$prior.ObservationValidationPassed=$false}
    Save-PairFixture (Join-Path $compiled 'sample-result.json') $prior
    $binding=[pscustomobject]@{Uninstrumented=$true;ToolchainReference=$identity;BinaryHash=$hash;CompilerEvidence=(Join-Path $compiled 'build/stderr.txt');Project=$workspace}
    if($Scenario -eq 'binary-binding'){$binding.BinaryHash='B'*64}
    Save-PairFixture (Join-Path $compiled 'binary-receipt.json') $binding
    foreach($file in @('before-product.json','after-product.json')){Save-PairFixture (Join-Path $compiled $file) @($product)}
    if($Scenario -eq 'build-log'){[IO.File]::WriteAllText((Join-Path $compiled 'build/stderr.txt'),'missing evidence')}
    if($Scenario -eq 'existing'){New-Item -ItemType Directory -Path $target|Out-Null}
    if($Scenario -eq 'wrong-target'){$target=Join-Path $unit 'envelope-observe-903'}
    $calls=New-Object 'System.Collections.Generic.List[string]'
    function Invoke-RecordedProcess([string]$Executable,[string[]]$Arguments,[hashtable]$Environment,[string]$ReceiptDir,[int]$LimitSeconds){
        $calls.Add($Executable)
        if($Executable -cne $old.Binary -or $LimitSeconds -ne 60 -or ($Arguments -join '|') -cne ('-test.run=^$|-test.fuzz=^FuzzIssuerPair$|-test.fuzztime=5s|-test.parallel=1|-test.timeout=30s|-test.count=1|-test.fuzzcachedir='+(Join-Path $target 'fuzz-cache'))){throw 'Unexpected process; no rebuild/Envelope/listing allowed'}
        foreach($name in @('TALENRO_FUZZ_DIAG_MODE','TALENRO_FUZZ_FIXTURE','GODEBUG')){if($Environment[$name] -cne ''){throw 'Diagnostic environment inherited'}}
        if($Environment.GOCACHE -cne (Join-Path $target 'go-cache') -or $Environment.TEMP -cne (Join-Path $target 'tmp') -or $Environment.GOPROXY -cne 'off' -or $Environment.GOFLAGS -cne '-mod=readonly'){throw 'Fresh offline environment missing'}
        New-Item -ItemType Directory -Path $ReceiptDir|Out-Null
        $r=New-PairReceipt $Executable $Arguments $Environment $ReceiptDir
        $stdout='PASS';$stderr="fuzz: elapsed: 0s, gathering baseline coverage: 4/4 completed, now fuzzing with 1 workers`r`nfuzz: elapsed: 6s, execs: 23 (4/sec), new interesting: 0 (total: 4)"
        switch($Scenario){'exit'{$r.ExitCode=7};'timeout'{$r.TimedOut=$true};'gate'{$r.NaturalShutdownVerified=$false};'no-pass'{$stdout=''};'no-worker'{$stderr='no worker'};'zero'{$stderr=$stderr.Replace('execs: 23 ','execs: 0 ')};'diag'{$stderr+="`r`nTALENRO_FUZZ_DIAG unexpected"};'post-binary'{[IO.File]::WriteAllText($old.Binary,'changed')};'post-record'{[IO.File]::WriteAllText((Join-Path $compiled 'binary-receipt.json'),'{}')};'post-product'{$product.SHA256='CHANGED'};'throw'{throw 'controlled native wrapper exception'}}
        [IO.File]::WriteAllText($r.StdoutPath,$stdout);[IO.File]::WriteAllText($r.StderrPath,$stderr)
        return $r
    }
    $seen=New-Object 'System.Collections.Generic.List[object]';$failure=$null
    try{Invoke-PairCompiledSample $target $compiled $tool (Join-Path $unit 'reference.json') $modules|ForEach-Object {$seen.Add($_)}}catch{$failure=$_.Exception.Message}
    [pscustomobject]@{Failure=$failure;Calls=$calls.Count;Seen=@($seen.ToArray());Root=$target;CompileRoot=$compiled}
}
$good=Test-PairCase 'success'
Check-Pair ($null -eq $good.Failure -and $good.Calls -eq 1 -and $good.Seen.Count -eq 1 -and $good.Seen[0].ObservationValidationPassed -and $good.Seen[0].RecordedExecutions -eq 23) 'one bound Pair invocation succeeds without compile or extra process'
$saved=Get-Content -LiteralPath (Join-Path $good.Root 'sample-result.json') -Raw|ConvertFrom-Json
Check-Pair ($saved.Target -ceq 'FuzzIssuerPair' -and $saved.Uninstrumented -and $saved.SamplingExecuted -and -not $saved.Acceptance) 'saved result distinguishes Pair from Envelope and acceptance'
foreach($scenario in @('prior-failed','build-args','build-env','binary-binding','build-log','existing','wrong-target')){
    $bad=Test-PairCase $scenario
    Check-Pair ($null -ne $bad.Failure -and $bad.Calls -eq 0 -and $bad.Seen.Count -eq 0) ($scenario+' refuses before native execution')
}
foreach($scenario in @('exit','timeout','gate','no-pass','no-worker','zero','diag','post-binary','post-record','post-product','post-tool','throw')){
    $bad=Test-PairCase $scenario
    Check-Pair ($null -ne $bad.Failure -and $bad.Calls -eq 1 -and $bad.Seen.Count -eq 0) ($scenario+' stops after one invocation without streaming success')
    $stopped=Get-Content -LiteralPath (Join-Path $bad.Root 'stopped.json') -Raw|ConvertFrom-Json
    Check-Pair ($stopped.SamplingAttempted -and -not $stopped.Acceptance) ($scenario+' preserves consumed attempt and false acceptance')
    if($scenario -ne 'throw'){$failed=Get-Content -LiteralPath (Join-Path $bad.Root 'sample-result.json') -Raw|ConvertFrom-Json;Check-Pair (-not $failed.ObservationValidationPassed -and -not $failed.Acceptance) ($scenario+' saves failed result')}
}
