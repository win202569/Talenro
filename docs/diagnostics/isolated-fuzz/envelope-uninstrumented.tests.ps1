param([string]$Toolchain='',[string]$ModuleCache='')
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'envelope-observe.ps1')
$helper=Join-Path $PSScriptRoot 'envelope-uninstrumented.ps1'
if(Test-Path -LiteralPath $helper){. $helper}
function Check-Plain([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
function Reject-Plain([scriptblock]$Action,[string]$Name) {$failed=$false;try{& $Action|Out-Null}catch{$failed=$true};Check-Plain $failed $Name}
Check-Plain ($null -ne (Get-Command Invoke-UninstrumentedEnvelope -CommandType Function -ErrorAction SilentlyContinue)) 'independent uninstrumented Envelope entry exists'
$manifest=@([pscustomobject]@{Path='bin/go.exe';SHA256='D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F';Length=1},[pscustomobject]@{Path='src/context/context.go';SHA256='765F99F7CC0710B0117FFA183E432524F429C21496FCBAD9B63FB9886FDBAAC7';Length=1},[pscustomobject]@{Path='src/internal/fuzz/fuzz.go';SHA256='D9DE12569D5C18FB629C8E79A4D629EB2985B719E4FBFE30C3F3CE23FC44FEB5';Length=1})+@(for($i=3;$i -lt 11536;$i++){[pscustomobject]@{Path=('fixture/f{0:D5}.go' -f $i);SHA256=('A'*64);Length=1}})
Assert-UninstrumentedEnvelopeManifest $manifest $manifest
Check-Plain $true 'complete unique pinned manifest accepted'
Reject-Plain {Assert-UninstrumentedEnvelopeManifest $manifest @($manifest[0..11534])} 'missing toolchain member refused'
foreach($change in @('hash','duplicate','hook','path','pin')) {
    $copy=@($manifest|ForEach-Object {[pscustomobject]@{Path=$_.Path;SHA256=$_.SHA256;Length=$_.Length}})
    switch($change){'hash'{$copy[-1].SHA256='B'*64};'duplicate'{$copy[-1].Path=$copy[0].Path};'hook'{$copy[-1].Path='src/context/diagnostic_fuzz.go'};'path'{$copy[-1].Path='../escaped.go'};'pin'{$copy[0].SHA256='B'*64}}
    if($change -ceq 'pin'){Reject-Plain {Assert-UninstrumentedEnvelopeManifest $copy $copy} 'self-consistent but wrong executable pin refused'}else{Reject-Plain {Assert-UninstrumentedEnvelopeManifest $manifest $copy} ('toolchain '+$change+' refused')}
}
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
$tool=if($Toolchain){$Toolchain}else{Join-Path $scope 'framework-001/toolchain'}
$modules=if($ModuleCache){$ModuleCache}else{Join-Path ([Environment]::GetFolderPath('UserProfile')) 'go/pkg/mod'}
function Test-PlainCase([string]$Scenario) {
    $target=Join-Path $scope ('plain-unit-'+[guid]::NewGuid().ToString('N')+'/envelope-observe-901')
    $plan=Get-EnvelopeObservationPlan $target $tool $modules
    $product=[pscustomobject]@{Path='fixture.go';SHA256='ABC'}
    function Get-EnvelopeProductManifest {[pscustomobject]@{Path=$product.Path;SHA256=$product.SHA256}}
    $binding=[pscustomobject]@{Calls=0}
    # Full real manifest comparison is tested above. Replace expensive source
    # enumeration and external Go execution in branch tests; never run unit binary.
    function Assert-UninstrumentedEnvelopeToolchain([string]$Toolchain,[string]$ReferenceReceipt) {$binding.Calls++;if($Scenario -ceq 'post-tool' -and $binding.Calls -eq 3){throw 'Controlled source drift'};return ('F'*64)}
    $calls=New-Object 'System.Collections.Generic.List[string]'
    function Invoke-RecordedProcess([string]$Executable,[string[]]$Arguments,[hashtable]$Environment,[string]$ReceiptDir,[int]$LimitSeconds) {
        $stage=Split-Path -Leaf $ReceiptDir;$calls.Add($stage)
        if($Environment.TALENRO_FUZZ_DIAG_MODE -cne '' -or $Environment.TALENRO_FUZZ_FIXTURE -cne '' -or $Environment.GODEBUG -cne '' -or $Environment.GOPROXY -cne 'off' -or $Environment.GOFLAGS -cne '-mod=readonly'){throw 'Unexpected uninstrumented environment'}
        $text='';$errors=''
        if($stage -ceq 'version' -and $Executable -ceq $plan.Executable -and $LimitSeconds -eq 60 -and ($Arguments -join '|') -ceq 'version'){$text='go version go1.26.5 windows/amd64'}
        elseif($stage -ceq 'origin' -and $Executable -ceq $plan.Executable -and $LimitSeconds -eq 60 -and ($Arguments -join '|') -ceq 'env|GOROOT|GOTOOLDIR|GOMOD'){$text=$tool+"`r`n"+(Join-Path $tool 'pkg/tool/windows_amd64')+"`r`n"+(Join-Path $workspace 'go.mod');if($Scenario -ceq 'origin-output'){$text='wrong origin'}}
        elseif($stage -ceq 'build' -and $Executable -ceq $plan.Executable -and $LimitSeconds -eq 600 -and ($Arguments -join '|') -ceq ('test|-c|-x|-p=1|-count=1|-run=^$|-fuzz=^FuzzIssuerEnvelope$|-fuzztime=5s|-parallel=1|-timeout=30s|-o='+$plan.Binary+'|talenro.local/platform/internal/nodecontrol/nodeissuer')) {
            [IO.File]::WriteAllText($plan.Binary,'never-executed-unit-binary');$compiler=Join-Path $tool 'pkg/tool/windows_amd64/compile.exe'
            $errors='"'+$compiler+'" -p context -pack "'+(Join-Path $tool 'src/context/context.go')+'"'+"`r`n"+'"'+$compiler+'" -p internal/fuzz -pack "'+(Join-Path $tool 'src/internal/fuzz/fuzz.go')+'"'+"`r`n"+'"'+$compiler+'" -p talenro.local/platform/internal/nodecontrol/nodeissuer -pack "'+(Join-Path $workspace 'internal/nodecontrol/nodeissuer/fuzz_test.go')+'" "'+(Join-Path $workspace 'internal/nodecontrol/nodeissuer/der.go')+'" "'+(Join-Path $workspace 'internal/nodecontrol/nodeissuer/signature.go')+'"'
            if($Scenario -ceq 'build-log'){$errors='missing compiler evidence'}
        } elseif($stage -ceq 'execute' -and $Executable -ceq $plan.Binary -and $LimitSeconds -eq 60 -and ($Arguments -join '|') -ceq ('-test.run=^$|-test.fuzz=^FuzzIssuerEnvelope$|-test.fuzztime=5s|-test.parallel=1|-test.timeout=30s|-test.count=1|-test.fuzzcachedir='+(Join-Path $target 'fuzz-cache'))){$text='PASS';$errors="fuzz: elapsed: 0s, gathering baseline coverage: 5/5 completed, now fuzzing with 1 workers`r`nfuzz: elapsed: 6s, execs: 32 (5/sec), new interesting: 0 (total: 5)";if($Scenario -ceq 'sample-diag'){$errors+="`r`nTALENRO_FUZZ_DIAG unexpected"};if($Scenario -ceq 'sample-no-worker'){$errors='no worker evidence'}}
        else{throw 'Unexpected uninstrumented process contract'}
        New-Item -ItemType Directory -Path $ReceiptDir -Force|Out-Null
        $stdout=Join-Path $ReceiptDir 'stdout.txt';$stderr=Join-Path $ReceiptDir 'stderr.txt';[IO.File]::WriteAllText($stdout,$text);[IO.File]::WriteAllText($stderr,$errors)
        $receipt=[pscustomobject]@{ExitCode=0;TimedOut=$false;RootExited=$true;CleanupVerified=$true;ActiveAfterCleanup=0;DescendantsRemaining=0;TotalProcesses=2;RootId=17;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;TerminationRequested=$false;ActiveBeforeTermination=0;MembersBeforeTermination=@();MembersAfterCleanup=@();MembersAtRootWait=@();NaturalObservationMs=0;ShutdownElapsedMs=0;ShutdownSchema=3;ShutdownStatePolicy='owned-final-signaled-exit-image-optional-v2';MemberEvidenceTruncated=$false;ShutdownMemberEvidence=@();StdoutPath=$stdout;StderrPath=$stderr}
        if($stage -ceq 'version' -and $Scenario -ceq 'version-gate'){$receipt.NaturalShutdownVerified=$false}
        if($stage -ceq 'build' -and $Scenario -ceq 'build-exit'){$receipt.ExitCode=7}
        if($stage -ceq 'execute') {if($Scenario -ceq 'sample-exit'){$receipt.ExitCode=7};if($Scenario -ceq 'sample-gate'){$receipt.TimedOut=$true};if($Scenario -ceq 'post-binary'){[IO.File]::WriteAllText($plan.Binary,'changed')};if($Scenario -ceq 'post-product'){$product.SHA256='CHANGED'}}
        return $receipt
    }
    $observed=New-Object 'System.Collections.Generic.List[object]';$failure=$null
    try{Invoke-UninstrumentedEnvelope $target $tool (Join-Path $scope 'reference.json') $modules|ForEach-Object {$observed.Add($_)}}catch{$failure=$_.Exception.Message}
    [pscustomobject]@{Root=$target;Failure=$failure;Calls=$calls.Count;Observed=@($observed.ToArray())}
}
$success=Test-PlainCase 'success'
Check-Plain ($null -eq $success.Failure -and $success.Calls -eq 4 -and $success.Observed.Count -eq 1 -and $success.Observed[0].ObservationValidationPassed) 'one uninstrumented version origin build execute sequence succeeds'
$saved=Get-Content -LiteralPath (Join-Path $success.Root 'sample-result.json') -Raw|ConvertFrom-Json
Check-Plain ($saved.Uninstrumented -and $saved.SamplingExecuted -and -not $saved.Acceptance) 'uninstrumented sample cannot imply product acceptance'
foreach($case in @(@('version-gate',1),@('origin-output',2),@('build-exit',3),@('build-log',3),@('sample-exit',4),@('sample-gate',4),@('sample-diag',4),@('sample-no-worker',4),@('post-binary',4),@('post-product',4),@('post-tool',4))) {
    $failed=Test-PlainCase $case[0];$stopped=Get-Content -LiteralPath (Join-Path $failed.Root 'stopped.json') -Raw|ConvertFrom-Json
    Check-Plain ($null -ne $failed.Failure -and $failed.Calls -eq $case[1] -and $failed.Observed.Count -eq 0 -and -not $stopped.Acceptance) ($case[0]+' stops without retry or premature success')
    if($case[1] -eq 4){$result=Get-Content -LiteralPath (Join-Path $failed.Root 'sample-result.json') -Raw|ConvertFrom-Json;Check-Plain (-not $result.ObservationValidationPassed -and -not $result.Acceptance) ($case[0]+' preserves failed observation')}
}
