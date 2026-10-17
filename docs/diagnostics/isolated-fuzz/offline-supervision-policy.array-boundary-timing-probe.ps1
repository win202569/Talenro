# Diagnostic spike only: isolated 4096/4097 boundary path, no original suite run.
$ErrorActionPreference='Stop'
function P1-Need([bool]$Value){if(-not $Value){throw ('ASSERTION FAILED: '+$script:p1Case)}}
function New-P1Scenario($StartMs=0,[object[]]$Observations=[object[]]@()){
    [pscustomobject]@{Schema=1;Mode='policy-only';SessionLabel='s';ClockLabel='synthetic-ms';StartMs=$StartMs;Observations=$Observations}
}
function P1-Observation($Sequence,$AtMs,[string]$Kind,[string]$Actor='W',[hashtable]$Payload=@{}){
    $v=[ordered]@{Sequence=$Sequence;SessionLabel='s';ClockLabel='synthetic-ms';AtMs=$AtMs;Actor=$Actor;Kind=$Kind}
    foreach($key in $Payload.Keys){$v[$key]=$Payload[$key]};[pscustomobject]$v
}
function P1-Want([hashtable]$Root=@{},[hashtable]$Decision=@{}){
    $d=[ordered]@{DeadlineMs=120000;StopAtMs=113500;NaturalUntilMs=$null;CleanupUntilMs=$null;PublishUntilMs=$null;LastAtMs=0;LastSequence=$null;StdoutRetainedBytes=0;StderrRetainedBytes=0;DataEventCount=0;ControlSummaryCount=0;ControlSummaryByteMax=0;ControlQueueDepth=0;WouldCaptureStatus='incomplete';WouldCleanupStatus='unknown';WouldStopOwnedSession=$false;WouldNaturalCompletion=$false;WouldEvidenceComplete=$false;WouldPositiveTechnicalCandidate=$false;WouldExitObserved=$false;WouldExitCode=$null;WouldClassification='unknown'}
    foreach($k in $Decision.Keys){$d[$k]=$Decision[$k]}
    $r=[ordered]@{Schema=1;Mode='policy-only';EvidenceClass='synthetic';Evaluated=$true;Reason='trace-evaluated';State='Prepared';FailureObservationIndex=$null;Decision=[pscustomobject]$d;OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false;LiveActionsPerformed=$false;CleanupStatus='unknown'}
    foreach($k in $Root.Keys){$r[$k]=$Root[$k]};[pscustomobject]$r
}
function P1-Equal($Actual,$Expected,[int]$Depth=0){
    if($Depth -gt 4){return $false};if($null -eq $Expected){return ($null -eq $Actual)}
    if($Expected -is [pscustomobject]){
        if($Actual -isnot [pscustomobject] -or @($Actual.PSObject.Properties).Count -ne @($Expected.PSObject.Properties).Count){return $false}
        foreach($p in $Expected.PSObject.Properties){$q=@($Actual.PSObject.Properties|Where-Object {[string]::Equals($_.Name,$p.Name,[StringComparison]::Ordinal)});if($q.Count -ne 1 -or $q[0].MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty -or -not (P1-Equal $q[0].Value $p.Value ($Depth+1))){return $false}};return $true
    }
    if($Expected -is [string]){return ($Actual -is [string] -and [string]::Equals($Actual,$Expected,[StringComparison]::Ordinal))}
    if($Expected -is [bool]){return ($Actual -is [bool] -and $Actual -eq $Expected)}
    if($Expected -is [int] -or $Expected -is [long]){return (($Actual -is [int] -or $Actual -is [long]) -and $Actual -eq $Expected)}
    return $false
}
function BoundaryTiming-Assert($r,$Expected){
    P1-Need (P1-Equal $r $Expected)
}
function BoundaryTiming-Stage([string]$Phase,[int]$Rows,[scriptblock]$Body){
    $script:boundaryStarted++;$id=$script:boundaryStarted
    if($id -gt 6){throw 'TIMING ERROR: phase bound'}
    $start=[pscustomobject]@{Schema=1;Event='start';Id=$id;Phase=$Phase;Rows=$Rows;ProductCalls=$script:boundaryCalls;AtMs=$script:boundaryClock.ElapsedMilliseconds;ElapsedMs=$null}
    [Console]::Out.WriteLine('TIMING: '+($start|ConvertTo-Json -Compress));[Console]::Out.Flush()
    $timer=[Diagnostics.Stopwatch]::StartNew()
    $answer=& $Body
    $timer.Stop();$script:boundaryCompleted++
    $end=[pscustomobject]@{Schema=1;Event='end';Id=$id;Phase=$Phase;Rows=$Rows;ProductCalls=$script:boundaryCalls;AtMs=$script:boundaryClock.ElapsedMilliseconds;ElapsedMs=[Math]::Round($timer.Elapsed.TotalMilliseconds,3)}
    [Console]::Out.WriteLine('TIMING: '+($end|ConvertTo-Json -Compress));[Console]::Out.Flush()
    return ,$answer
}
function BoundaryTiming-Decide($Scenario){
    if($script:boundaryCalls -ge 2){throw 'TIMING ERROR: product call bound'}
    $script:boundaryCalls++
    Get-OfflineSupervisionDecision $Scenario
}
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=[Security.Principal.WindowsPrincipal]::new($identity)
    Write-Output ('CONTEXT: '+$identity.Name+' admin='+$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
    Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
    $script:boundaryCalls=0;$script:boundaryStarted=0;$script:boundaryCompleted=0
    $script:p1Case='P1ArrayAndIntegerBounds';$script:boundaryClock=[Diagnostics.Stopwatch]::StartNew()
    . (Join-Path $PSScriptRoot 'offline-supervision-policy.ps1')
    $input4096=BoundaryTiming-Stage Build4096 4096 {
        $many=New-Object object[] 4096;for($i=0;$i -lt 4096;$i++){$many[$i]=P1-Observation $i 0 time}
        $scenario=New-P1Scenario -Observations $many
        P1-Need ($scenario.Observations.Length -eq 4096)
        [pscustomobject]@{Scenario=$scenario;Expected=(P1-Want @{} @{LastAtMs=0;LastSequence=4095})}
    }
    $result4096=BoundaryTiming-Stage Evaluate4096 4096 {BoundaryTiming-Decide $input4096.Scenario}
    $null=BoundaryTiming-Stage Assert4096 4096 {BoundaryTiming-Assert $result4096 $input4096.Expected}
    $input4097=BoundaryTiming-Stage Build4097 4097 {
        $scenario=New-P1Scenario -Observations ([object[]]($input4096.Scenario.Observations+@((P1-Observation 4096 0 time))))
        P1-Need ($scenario.Observations.Length -eq 4097)
        [pscustomobject]@{Scenario=$scenario;Expected=(P1-Want @{Evaluated=$false;Reason='invalid-scenario';State='Refused';FailureObservationIndex=$null;Decision=$null})}
    }
    $result4097=BoundaryTiming-Stage Evaluate4097 4097 {BoundaryTiming-Decide $input4097.Scenario}
    $null=BoundaryTiming-Stage Assert4097 4097 {BoundaryTiming-Assert $result4097 $input4097.Expected}
    Write-Output 'PASS: ArrayBoundaryTiming'
    $summary=[pscustomobject]@{Schema=1;ExpectedPhases=6;CompletedPhases=$script:boundaryCompleted;ProductCalls=$script:boundaryCalls;ElapsedMs=$script:boundaryClock.ElapsedMilliseconds;OriginalTestsRun=$false;FixturePreflightRun=$false;ProductChanged=$false}
    Write-Output ('TIMING-SUMMARY: '+($summary|ConvertTo-Json -Compress))
    Write-Output 'LIMIT: isolated array-boundary timing only; no full suite, retries, repairs or budget renewal.'
}catch{
    $failure=[pscustomobject]@{Message=$_.Exception.Message;StartedPhases=$script:boundaryStarted;CompletedPhases=$script:boundaryCompleted;ProductCalls=$script:boundaryCalls}
    Write-Output ('PROBE-ERROR: '+($failure|ConvertTo-Json -Compress));exit 1
}
