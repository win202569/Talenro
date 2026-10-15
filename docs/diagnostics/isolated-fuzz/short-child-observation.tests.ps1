$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'run.ps1')
$helper=Join-Path $PSScriptRoot 'short-child-observation.ps1'
if(Test-Path -LiteralPath $helper){. $helper}
function Check-ShortChild([bool]$Value,[string]$Name){if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
Check-ShortChild ($null -ne (Get-Command Get-ShortChildObservationPlan -ErrorAction SilentlyContinue)) 'bounded short-child entry exists'
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
$unit=Join-Path $scope ('short-child-unit-'+[guid]::NewGuid().ToString('N'))
$root=Join-Path $unit 'short-child-observe-901'
$plan=Get-ShortChildObservationPlan $root
Check-ShortChild (-not (Test-Path -LiteralPath $root) -and $plan.LimitSeconds -eq 20 -and $plan.ChildDelayMs -eq 150 -and $plan.ReadyLimitMs -eq 10000) 'planning creates nothing and fixes approved limits'
Check-ShortChild ($plan.Arguments[0] -ceq '-NoProfile' -and $plan.Arguments[1] -ceq '-NonInteractive' -and $plan.Arguments[2] -ceq '-EncodedCommand') 'fixed noninteractive parent invocation'
Check-ShortChild ([Text.Encoding]::Unicode.GetString([Convert]::FromBase64String($plan.Arguments[3])) -ceq $plan.ParentScript) 'encoded native argument preserves actual parent script'
$caught=$false;try{Get-ShortChildObservationPlan (Join-Path $unit 'matrix-check-901')|Out-Null}catch{$caught=$true}
Check-ShortChild $caught 'wrong identity refused before launch'
New-Item -ItemType Directory -Path $root -Force|Out-Null
$caught=$false;try{Get-ShortChildObservationPlan $root|Out-Null}catch{$caught=$true}
Check-ShortChild $caught 'consumed identity cannot replay'
$calls=New-Object 'System.Collections.Generic.List[object]'
function Start-Process {
    param($FilePath,$ArgumentList,$WindowStyle)
    $calls.Add([pscustomobject]@{Executable=$FilePath;Arguments=$ArgumentList;WindowStyle=$WindowStyle})
    [pscustomobject]@{Nonce=$plan.Nonce;ProcessId=23;UTC=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json -Compress|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'child-ready.json')
}
& ([scriptblock]::Create($plan.ParentScript))
Check-ShortChild ($calls.Count -eq 1 -and $calls[0].Executable -ceq $plan.Executable -and $calls[0].WindowStyle -ceq 'Hidden') 'parent performs one fixed hidden child launch through fixture boundary'
Check-ShortChild ([Text.Encoding]::Unicode.GetString([Convert]::FromBase64String(($calls[0].Arguments -split ' ')[-1])) -ceq $plan.ChildScript) 'child invocation carries the tested generated child body'
$release=Get-Content (Join-Path $root 'release.json') -Raw|ConvertFrom-Json
$parentExit=Get-Content (Join-Path $root 'parent-exit.json') -Raw|ConvertFrom-Json
Check-ShortChild ($release.Nonce -ceq $plan.Nonce -and $parentExit.Nonce -ceq $plan.Nonce -and $release.ProcessId -eq $PID -and $parentExit.ProcessId -eq $PID) 'parent releases only after readiness and retains own stage identity'
$childRoot=Join-Path $unit 'short-child-observe-902';$childPlan=Get-ShortChildObservationPlan $childRoot
New-Item -ItemType Directory -Path $childRoot|Out-Null
[pscustomobject]@{Nonce=$childPlan.Nonce;ProcessId=$PID;UTC=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json -Compress|Set-Content -LiteralPath (Join-Path $childRoot 'release.json')
$sleeps=New-Object 'System.Collections.Generic.List[int]'
function Start-Sleep {param([int]$Milliseconds);$sleeps.Add($Milliseconds)}
& ([scriptblock]::Create($childPlan.ChildScript))
Check-ShortChild ($sleeps.Count -eq 1 -and $sleeps[0] -eq 150) 'released child requests exactly one approved short delay'
$ready=Get-Content (Join-Path $childRoot 'child-ready.json') -Raw|ConvertFrom-Json
$exit=Get-Content (Join-Path $childRoot 'child-exit.json') -Raw|ConvertFrom-Json
Check-ShortChild ($ready.Nonce -ceq $childPlan.Nonce -and $exit.Nonce -ceq $childPlan.Nonce -and $ready.ProcessId -eq $PID -and $exit.ProcessId -eq $PID) 'child stage payloads preserve nonce and self identity without PID queries'
# Literal fixture receipt: no native execution or child process. Marker PIDs are
# current-host values from evaluating scripts, not claims of actual job ownership.
$receipt=[pscustomobject]@{RootId=$PID;StartUTC=[DateTime]::UtcNow.AddMinutes(-1).ToString('o');EndUTC=[DateTime]::UtcNow.AddMinutes(1).ToString('o');MemberEvidence=@([pscustomobject]@{Id=23;Owned=$true;QueryError=0;ExitCodeRead=$true});InitialAccountingQuery=[pscustomobject]@{Succeeded=$true;Error=0;ValuesRead=$true;Frequency=10;StartTimestamp=1;EndTimestamp=2};InitialMemberQueries=@([pscustomobject]@{Succeeded=$true;Error=0;HeaderRead=$true;Frequency=10;StartTimestamp=3;EndTimestamp=4});ExitCode=0;TimedOut=$false;CleanupVerified=$true;ActiveAfterCleanup=0;DescendantsRemaining=0;RootExited=$true;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;TerminationRequested=$false;ActiveBeforeTermination=0;MembersBeforeTermination=@();MembersAfterCleanup=@();NaturalObservationMs=0;ShutdownElapsedMs=0}
$shutdownRow=[pscustomobject]@{Id=23;Owned=$true;QueryError=0;ExitCodeRead=$true;ExitCode=0;State='exited';ConfirmationAttempted=$false;ConfirmationState=$null;DiagnosticShutdown=$null;FinalShutdown=$null}
$receipt|Add-Member ShutdownSchema 3
$receipt|Add-Member ShutdownStatePolicy 'owned-final-signaled-exit-image-optional-v2'
$receipt|Add-Member MemberEvidenceTruncated $false
$receipt|Add-Member MembersAtRootWait @(23)
$receipt|Add-Member ShutdownMemberEvidence @($shutdownRow)
[pscustomobject]@{Nonce=$plan.Nonce;ProcessId=23;UTC=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json -Compress|Set-Content -LiteralPath (Join-Path $root 'child-exit.json')
$good=Get-ShortChildObservationEvidence $plan $receipt
Check-ShortChild ($good.StageEvidenceComplete -and $good.ChildMatchedOwnedEvidence -and $good.QueryEvidenceComplete -and $good.CompletionPassed -and $good.ObservationPassed -and -not $good.Acceptance) 'complete fixture observation stays diagnostic only'
$validReceiptJSON=$receipt|ConvertTo-Json -Depth 8
$receipt.MemberEvidence=@()
$evidence=Get-ShortChildObservationEvidence $plan $receipt
Check-ShortChild (-not $evidence.ChildMatchedOwnedEvidence -and -not $evidence.ObservationPassed) 'stage files and aggregate success cannot identify a missed member'
$receipt.MemberEvidence=@([pscustomobject]@{Id=23;Owned=$false;QueryError=87;ExitCodeRead=$false})
$evidence=Get-ShortChildObservationEvidence $plan $receipt
Check-ShortChild (-not $evidence.ChildMatchedOwnedEvidence -and -not $evidence.ObservationPassed) 'unknown-member error never becomes owned evidence'
$receipt.InitialMemberQueries=@()
$evidence=Get-ShortChildObservationEvidence $plan $receipt
Check-ShortChild (-not $evidence.QueryEvidenceComplete) 'missing query metadata remains incomplete'
$receipt.TerminationRequested=$true;$receipt.NaturalShutdownVerified=$false
$evidence=Get-ShortChildObservationEvidence $plan $receipt
Check-ShortChild (-not $evidence.CompletionPassed -and -not $evidence.ObservationPassed) 'forced cleanup cannot pass observation'
[pscustomobject]@{Nonce=('0'*32);ProcessId=23;UTC=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json -Compress|Set-Content -LiteralPath (Join-Path $root 'child-exit.json')
$evidence=Get-ShortChildObservationEvidence $plan $receipt
Check-ShortChild (-not $evidence.StageEvidenceComplete) 'wrong marker nonce refused'
function Test-ShortChildController([bool]$Refuse,[string]$Identity) {
    $controllerRoot=Join-Path $unit $Identity
    $nativeCalls=New-Object 'System.Collections.Generic.List[object]'
    function Invoke-RecordedProcess($Executable,$Arguments,$Environment,$ReceiptDir,$LimitSeconds) {
        $actualPlan=Get-Content (Join-Path $controllerRoot 'plan.json') -Raw|ConvertFrom-Json
        if($Executable -cne $actualPlan.Executable -or $LimitSeconds -ne 20 -or ($Arguments -join '|') -cne ($actualPlan.Arguments -join '|')){throw 'Controller changed fixed launch contract'}
        $nativeCalls.Add($actualPlan)
        $r=$validReceiptJSON|ConvertFrom-Json
        if($Refuse){$r.NaturalShutdownVerified=$false;$r.TerminationRequested=$true}
        foreach($name in @('child-ready.json','release.json','parent-exit.json','child-exit.json')){
            $id=if($name.StartsWith('child')){23}else{$r.RootId}
            [pscustomobject]@{Nonce=$actualPlan.Nonce;ProcessId=$id;UTC=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json -Compress|Set-Content -LiteralPath (Join-Path $controllerRoot $name)
        }
        return $r
    }
    $failure=$null;$returned=$null
    try{$returned=Invoke-ShortChildObservation $controllerRoot}catch{$failure=$_}
    Check-ShortChild ($nativeCalls.Count -eq 1) ('controller makes exactly one recorded invocation: '+$Identity)
    $saved=Get-Content (Join-Path $controllerRoot 'result.json') -Raw|ConvertFrom-Json
    if($Refuse){
        $stopped=Get-Content (Join-Path $controllerRoot 'stopped.json') -Raw|ConvertFrom-Json
        Check-ShortChild ($null -ne $failure -and -not $saved.ObservationPassed -and $stopped.NoRetry -and -not $stopped.Acceptance) 'controller saves refused result before stopped exception'
    }else{
        Check-ShortChild ($null -eq $failure -and $returned.ObservationPassed -and $saved.ObservationPassed -and -not $saved.Acceptance) 'controller persists complete diagnostic result without product acceptance'
    }
    $duplicate=$false;try{Invoke-ShortChildObservation $controllerRoot|Out-Null}catch{$duplicate=$true}
    Check-ShortChild ($duplicate -and $nativeCalls.Count -eq 1) ('duplicate controller identity cannot launch again: '+$Identity)
}
Test-ShortChildController $false 'short-child-observe-903'
Test-ShortChildController $true 'short-child-observe-904'
