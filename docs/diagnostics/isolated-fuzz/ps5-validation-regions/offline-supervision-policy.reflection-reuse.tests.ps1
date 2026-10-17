# Regression: metadata reuse must never cache input acceptance or result objects.
# Run separately from old suites; these assertions apply before/after optimization.
$ErrorActionPreference='Stop'
# BEGIN COMBINED TIMING HELPERS
# Diagnostic boundary samples only; no pipeline objects or product instrumentation.
function Combined-Event([string]$Kind,[string]$Name,[string]$Event){
    $stamp=[Diagnostics.Stopwatch]::GetTimestamp()
    $cpuProcess=[Diagnostics.Process]::GetCurrentProcess()
    try{$cpuTicks=$cpuProcess.TotalProcessorTime.Ticks}finally{$cpuProcess.Dispose()}
    $record=[pscustomobject]@{Schema=2;Event=$Event;Kind=$Kind;Name=$Name;Ticks=$stamp;Frequency=[Diagnostics.Stopwatch]::Frequency;CPUTicks=$cpuTicks}
    [Console]::Out.WriteLine('CTIME: '+($record|ConvertTo-Json -Compress));[Console]::Out.Flush()
}
# END COMBINED TIMING HELPERS
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
try{
    . (Join-Path $PSScriptRoot '../offline-supervision-policy.ps1')
    Combined-Event Reflection InputMutation start
    $script:p1Case='P1ReflectionInputMutation'
    $o=P1-Observation 0 0 time;$s=New-P1Scenario -Observations ([object[]]@($o))
    $a=Get-OfflineSupervisionDecision $s
    P1-Need (P1-Equal $a (P1-Want @{} @{LastAtMs=0;LastSequence=0}))
    $o.AtMs=1
    $b=Get-OfflineSupervisionDecision $s
    P1-Need (P1-Equal $b (P1-Want @{} @{LastAtMs=1;LastSequence=0}))
    P1-Need (P1-Equal $a (P1-Want @{} @{LastAtMs=0;LastSequence=0}))
    P1-Need (-not [object]::ReferenceEquals($a,$b) -and -not [object]::ReferenceEquals($a.Decision,$b.Decision))
    Write-Output ('PASS: '+$script:p1Case)
    Combined-Event Reflection InputMutation end
    Combined-Event Reflection DecorationAfterClean start
    $script:p1Case='P1ReflectionDecorationAfterClean'
    $storage=[Array]::CreateInstance([int],1);$plain=$storage.GetValue(0)
    $wrapped=[Management.Automation.PSObject].GetConstructor([type[]]@([object])).Invoke([object[]]@($plain))
    $wrapped.PSObject.Members.Add([Management.Automation.PSNoteProperty]::new('Members','untrusted'))
    $o.Sequence=$wrapped
    P1-Need ($null -ne $o.Sequence.PSObject.Members['Members'])
    $r=Get-OfflineSupervisionDecision $s
    P1-Need (P1-Equal $r (P1-Want @{Evaluated=$false;Reason='invalid-observation';State='Refused';FailureObservationIndex=0;Decision=$null}))
    $o.Sequence=0
    P1-Need (P1-Equal (Get-OfflineSupervisionDecision $s) (P1-Want @{} @{LastAtMs=1;LastSequence=0}))
    Write-Output ('PASS: '+$script:p1Case)
    Combined-Event Reflection DecorationAfterClean end
    Combined-Event Reflection DescriptorAfterClean start
    $script:p1Case='P1ReflectionDescriptorAfterClean'
    $script:reflectionCallbacks=0
    $bad=New-P1Scenario
    $bad.PSObject.Properties.Remove('Schema')
    $bad.PSObject.Properties.Add([Management.Automation.PSScriptProperty]::new('Schema',{ $script:reflectionCallbacks++;return 1 }))
    P1-Need (P1-Equal (Get-OfflineSupervisionDecision $bad) (P1-Want @{Evaluated=$false;Reason='invalid-scenario';State='Refused';FailureObservationIndex=$null;Decision=$null}))
    P1-Need ($script:reflectionCallbacks -eq 0)
    P1-Need (P1-Equal (Get-OfflineSupervisionDecision (New-P1Scenario)) (P1-Want))
    Write-Output ('PASS: '+$script:p1Case)
    Combined-Event Reflection DescriptorAfterClean end
    Write-Output 'LIMIT: reflection reuse regression only; no timing guarantee or runtime acceptance.'
}catch{[Console]::Out.WriteLine('REFLECTION-ERROR: '+$_.Exception.Message);exit 1}
