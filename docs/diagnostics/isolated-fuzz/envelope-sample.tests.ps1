param([string]$Toolchain='',[string]$ModuleCache='')
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'envelope-observe.ps1')
$helper=Join-Path $PSScriptRoot 'envelope-sample.ps1'
if(Test-Path -LiteralPath $helper){. $helper}
function Check-Sample([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
Check-Sample ($null -ne (Get-Command Invoke-EnvelopeCompiledSample -CommandType Function -ErrorAction SilentlyContinue)) 'independent compiled observation entry exists'
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
$tool=if($Toolchain){$Toolchain}else{Join-Path $scope 'framework-001/toolchain'}
$modules=if($ModuleCache){$ModuleCache}else{Join-Path ([Environment]::GetFolderPath('UserProfile')) 'go/pkg/mod'}
function New-SampleReceipt([string]$Executable,[string[]]$Arguments,[hashtable]$Environment) {
    [pscustomobject]@{Executable=$Executable;Arguments=$Arguments;Environment=$Environment;ExitCode=0;TimedOut=$false;RootExited=$true;CleanupVerified=$true;ActiveAfterCleanup=0;DescendantsRemaining=0;TotalProcesses=2;RootId=17;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;TerminationRequested=$false;ActiveBeforeTermination=0;MembersBeforeTermination=@();MembersAfterCleanup=@();MembersAtRootWait=@();NaturalObservationMs=0;ShutdownElapsedMs=0;ShutdownSchema=3;ShutdownStatePolicy='owned-final-signaled-exit-image-optional-v2';MemberEvidenceTruncated=$false;ShutdownMemberEvidence=@()}
}
function Test-CompiledSample([string]$Scenario) {
    $unit=Join-Path $scope ('envelope-sample-tests-'+[guid]::NewGuid().ToString('N'))
    $compiledRoot=Join-Path $unit 'envelope-observe-901';$target=Join-Path $unit 'envelope-observe-902'
    $buildPlan=Get-EnvelopeObservationPlan $compiledRoot $tool $modules
    $identity=Assert-EnvelopeRunProvenance $buildPlan
    New-Item -ItemType Directory -Path (Join-Path $compiledRoot 'build') -Force|Out-Null
    [IO.File]::WriteAllText($buildPlan.Binary,'never-executed-unit-binary')
    $binaryHash=(Get-FileHash -LiteralPath $buildPlan.Binary).Hash
    $product=[pscustomobject]@{Path='fixture.go';SHA256='ABC'}
    function Get-EnvelopeProductManifest {if($Scenario -ceq 'pre-product'){[pscustomobject]@{Path='fixture.go';SHA256='CHANGED'}}else{$product}}
    foreach($name in @('before-product.json','after-product.json')){@($product)|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $compiledRoot $name)}
    $compiler=Join-Path $tool 'pkg/tool/windows_amd64/compile.exe'
    $log='"'+$compiler+'" -p context -pack "'+(Join-Path $tool 'src/context/context.go')+'"'+"`r`n"+'"'+$compiler+'" -p internal/fuzz -pack "'+(Join-Path $tool 'src/internal/fuzz/fuzz.go')+'"'+"`r`n"+'"'+$compiler+'" -p talenro.local/platform/internal/nodecontrol/nodeissuer -pack "'+(Join-Path $workspace 'internal/nodecontrol/nodeissuer/fuzz_test.go')+'" "'+(Join-Path $workspace 'internal/nodecontrol/nodeissuer/der.go')+'" "'+(Join-Path $workspace 'internal/nodecontrol/nodeissuer/signature.go')+'"'
    [IO.File]::WriteAllText((Join-Path $compiledRoot 'build/stderr.txt'),$log)
    $buildReceipt=New-SampleReceipt $buildPlan.Executable $buildPlan.BuildArguments $buildPlan.Environment
    if($Scenario -ceq 'pre-gate'){$buildReceipt.NaturalShutdownVerified=$false}
    $buildReceipt|ConvertTo-Json -Depth 6|Set-Content -LiteralPath (Join-Path $compiledRoot 'build/receipt.json')
    $binding=[pscustomobject]@{SourceIdentity=$identity;BinaryHash=$binaryHash;OriginalFuzzHash='318CF8DE04BA5D95E71E625ABF6AAED8DFD54B854A9A925B7A3004D4132D989E';ProductBefore=(Join-Path $compiledRoot 'before-product.json');CompilerEvidence=(Join-Path $compiledRoot 'build/stderr.txt');GOROOT=$tool;Project=$workspace}
    $saved=[pscustomobject]@{CompileOnly=$true;BuildPassed=$true;BuildExecuted=$true;SamplingExecuted=$false;Acceptance=$false;SourceIdentity=$identity;BinaryHash=$binaryHash;RunRoot=$compiledRoot;Binary=$buildPlan.Binary;BinaryReceipt=(Join-Path $compiledRoot 'binary-receipt.json');CompilerEvidence=(Join-Path $compiledRoot 'build/stderr.txt')}
    if($Scenario -ceq 'pre-mode'){$saved.CompileOnly=$false}
    if($Scenario -ceq 'pre-root'){$saved.RunRoot=$target}
    if($Scenario -ceq 'pre-binding'){$binding.SourceIdentity='0'*64}
    $binding|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $compiledRoot 'binary-receipt.json')
    $saved|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $compiledRoot 'compile-result.json')
    if($Scenario -ceq 'pre-binary'){[IO.File]::WriteAllText($buildPlan.Binary,'changed')}
    if($Scenario -ceq 'pre-log'){[IO.File]::WriteAllText((Join-Path $compiledRoot 'build/stderr.txt'),'missing compiler evidence')}
    if($Scenario -ceq 'existing'){New-Item -ItemType Directory -Path $target|Out-Null}
    $calls=New-Object 'System.Collections.Generic.List[string]'
    function Invoke-RecordedProcess([string]$Executable,[string[]]$Arguments,[hashtable]$Environment,[string]$ReceiptDir,[int]$LimitSeconds) {
        $calls.Add($Executable)
        $expected='-test.run=^$|-test.fuzz=^FuzzIssuerEnvelope$|-test.fuzztime=5s|-test.parallel=1|-test.timeout=30s|-test.count=1|-test.fuzzcachedir='+(Join-Path $target 'fuzz-cache')
        if($calls.Count -ne 1 -or $Executable -cne $buildPlan.Binary -or ($Arguments -join '|') -cne $expected -or $LimitSeconds -ne 60 -or $Environment.GOPROXY -cne 'off' -or $Environment.GOFLAGS -cne '-mod=readonly'){throw 'Wrong single observation process contract'}
        New-Item -ItemType Directory -Path $ReceiptDir -Force|Out-Null
        $receipt=New-SampleReceipt $Executable $Arguments $Environment
        $receipt|Add-Member -NotePropertyName StdoutPath -NotePropertyValue (Join-Path $ReceiptDir 'stdout.txt')
        $receipt|Add-Member -NotePropertyName StderrPath -NotePropertyValue (Join-Path $ReceiptDir 'stderr.txt')
        if($Scenario -ceq 'sample-gate'){$receipt.TimedOut=$true;$receipt.NaturalShutdownVerified=$false}
        if($Scenario -ceq 'sample-exit'){$receipt.ExitCode=7}
        $trace=@([pscustomobject]@{schema=1;seq=1;mode='observe';event='stop';source='worker';incoming='nil';childErr='nil';parentRead=$false;parentErr='unread';oldError='nil';newError='nil';stopping=$true;stoppingRead=$true;watchdog=$false;activeWorkers=1},[pscustomobject]@{schema=1;seq=2;mode='observe';event='return';source='return';incoming='nil';childErr='unread';parentRead=$false;parentErr='unread';oldError='unread';newError='nil';stopping=$false;stoppingRead=$false;watchdog=$false;activeWorkers=0})
        if($Scenario -ceq 'sample-watchdog'){$trace[0].watchdog=$true}
        $lines=@($trace|ForEach-Object {'TALENRO_FUZZ_DIAG '+($_|ConvertTo-Json -Compress)})
        [IO.File]::WriteAllText($receipt.StdoutPath,'controlled observation')
        [IO.File]::WriteAllLines($receipt.StderrPath,$lines)
        $receipt|ConvertTo-Json -Depth 6|Set-Content -LiteralPath (Join-Path $ReceiptDir 'receipt.json')
        if($Scenario -ceq 'post-binary'){[IO.File]::WriteAllText($buildPlan.Binary,'changed after observation')}
        if($Scenario -ceq 'post-record'){[IO.File]::WriteAllText((Join-Path $compiledRoot 'compile-result.json'),'changed after observation')}
        return $receipt
    }
    $observed=New-Object 'System.Collections.Generic.List[object]';$failure=$null
    try{Invoke-EnvelopeCompiledSample $target $compiledRoot $tool $modules|ForEach-Object {$observed.Add($_)}}catch{$failure=$_.Exception.Message}
    [pscustomobject]@{Root=$target;CompileRoot=$compiledRoot;Failure=$failure;Calls=$calls.Count;Observed=@($observed.ToArray())}
}
$success=Test-CompiledSample 'success'
Check-Sample ($null -eq $success.Failure -and $success.Calls -eq 1 -and $success.Observed.Count -eq 1 -and $success.Observed[0].ObservationValidationPassed) 'validated compiled artifact executes once without rebuilding'
$result=Get-Content -LiteralPath (Join-Path $success.Root 'sample-result.json') -Raw|ConvertFrom-Json
Check-Sample ($result.SamplingExecuted -and -not $result.Acceptance -and $result.CompileRoot -ceq $success.CompileRoot -and $result.ObservationValidationPassed) 'saved observation is bound and never product acceptance'
Check-Sample ((Test-Path -LiteralPath (Join-Path $success.Root 'after-product.json')) -and -not (Test-Path -LiteralPath (Join-Path $success.Root 'build'))) 'observation preserves final inventory without a new build'
foreach($scenario in @('pre-mode','pre-root','pre-binding','pre-binary','pre-log','pre-product','pre-gate','existing')) {
    $failed=Test-CompiledSample $scenario
    Check-Sample ($null -ne $failed.Failure -and $failed.Calls -eq 0 -and $failed.Observed.Count -eq 0 -and ($scenario -ceq 'existing' -or -not (Test-Path -LiteralPath $failed.Root))) ($scenario+' refuses before a process or new observation root')
}
foreach($scenario in @('sample-exit','sample-gate','sample-watchdog','post-binary','post-record')) {
    $failed=Test-CompiledSample $scenario
    $stopped=Get-Content -LiteralPath (Join-Path $failed.Root 'stopped.json') -Raw|ConvertFrom-Json
    Check-Sample ($null -ne $failed.Failure -and $failed.Calls -eq 1 -and $failed.Observed.Count -eq 0 -and -not $stopped.Acceptance) ($scenario+' stops after one observation without streaming success or retry')
    $saved=Get-Content -LiteralPath (Join-Path $failed.Root 'sample-result.json') -Raw|ConvertFrom-Json
    Check-Sample (-not $saved.ObservationValidationPassed -and $saved.SamplingExecuted -and -not $saved.Acceptance) ($scenario+' preserves a failed rather than successful result')
}
