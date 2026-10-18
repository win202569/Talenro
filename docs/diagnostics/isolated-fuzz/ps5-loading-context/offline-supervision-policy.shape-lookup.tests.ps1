# Contract: ordinal field admission and per-evaluation schema isolation.
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
function Shape-Refusal([string]$Reason='invalid-scenario',$Index=$null){P1-Want @{Evaluated=$false;Reason=$Reason;State='Refused';FailureObservationIndex=$Index;Decision=$null}}
function Shape-Rename($Owner,[string]$From,[string]$To){$value=$Owner.PSObject.Properties[$From].Value;$Owner.PSObject.Properties.Remove($From);$Owner.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new($To,$value))}
function Shape-Mixed {
    New-P1Scenario -Observations ([object[]]@(
        (P1-Observation 0 0 time),
        (P1-Observation 1 1 control W @{SummaryCount=0;MaxSummaryBytes=0;QueueDepth=0}),
        (P1-Observation 2 2 time),
        (P1-Observation 3 3 control W @{SummaryCount=0;MaxSummaryBytes=0;QueueDepth=0})
    ))
}
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=[Security.Principal.WindowsPrincipal]::new($identity)
    Write-Output ('CONTEXT: '+$identity.Name+' admin='+$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
    Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
    . (Join-Path $PSScriptRoot '../offline-supervision-policy.ps1')
    $script:p1Case='P1ShapeRootOrdinal'
    foreach($name in @('Schema','Mode','SessionLabel','ClockLabel','StartMs','Observations')){
        $s=New-P1Scenario;Shape-Rename $s $name $name.ToLowerInvariant()
        P1-Need (P1-Equal (Get-OfflineSupervisionDecision $s) (Shape-Refusal))
    }
    Write-Output ('PASS: '+$script:p1Case)
    $script:p1Case='P1ShapeFieldBounds'
    $s=New-P1Scenario;$s.PSObject.Properties.Remove('StartMs');P1-Need (P1-Equal (Get-OfflineSupervisionDecision $s) (Shape-Refusal))
    $s=New-P1Scenario;$s.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Extra',0));P1-Need (P1-Equal (Get-OfflineSupervisionDecision $s) (Shape-Refusal))
    $s=New-P1Scenario;Shape-Rename $s StartMs Unknown;P1-Need (P1-Equal (Get-OfflineSupervisionDecision $s) (Shape-Refusal))
    Write-Output ('PASS: '+$script:p1Case)
    $script:p1Case='P1ShapeObservationOrdinal'
    foreach($name in @('Sequence','SessionLabel','ClockLabel','AtMs','Actor','Kind')){
        $o=P1-Observation 0 0 time;Shape-Rename $o $name $name.ToLowerInvariant()
        P1-Need (P1-Equal (Get-OfflineSupervisionDecision (New-P1Scenario -Observations ([object[]]@($o)))) (Shape-Refusal invalid-observation 0))
    }
    Write-Output ('PASS: '+$script:p1Case)
    $script:p1Case='P1ShapeKindIsolation'
    P1-Need (P1-Equal (Get-OfflineSupervisionDecision (Shape-Mixed)) (P1-Want @{} @{LastAtMs=3;LastSequence=3}))
    $s=Shape-Mixed;Shape-Rename $s.Observations[2] ClockLabel Unknown;P1-Need (P1-Equal (Get-OfflineSupervisionDecision $s) (Shape-Refusal invalid-observation 2))
    $s=Shape-Mixed;Shape-Rename $s.Observations[3] SummaryCount Unknown;P1-Need (P1-Equal (Get-OfflineSupervisionDecision $s) (Shape-Refusal invalid-observation 3))
    Write-Output ('PASS: '+$script:p1Case)
    $script:p1Case='P1ShapeRepeatedCalls'
    $s=Shape-Mixed;$a=Get-OfflineSupervisionDecision $s;$b=Get-OfflineSupervisionDecision $s;$want=P1-Want @{} @{LastAtMs=3;LastSequence=3}
    P1-Need ((P1-Equal $a $want) -and (P1-Equal $b $want) -and -not [object]::ReferenceEquals($a.Decision,$b.Decision))
    Shape-Rename $s.Observations[2] AtMs Wrong
    P1-Need (P1-Equal (Get-OfflineSupervisionDecision $s) (Shape-Refusal invalid-observation 2))
    P1-Need ((P1-Equal $a $want) -and (P1-Equal $b $want))
    P1-Need (P1-Equal (Get-OfflineSupervisionDecision (Shape-Mixed)) $want)
    Write-Output ('PASS: '+$script:p1Case)
    $script:p1Case='P1ShapeDescriptorSafety'
    $script:shapeCallbacks=0;$s=New-P1Scenario;$s.PSObject.Properties.Remove('Schema')
    $s.PSObject.Properties.Add([Management.Automation.PSScriptProperty]::new('Schema',{ $script:shapeCallbacks++;return 1 }))
    P1-Need (P1-Equal (Get-OfflineSupervisionDecision $s) (Shape-Refusal))
    $s=Shape-Mixed;$s.Observations[2].PSObject.Properties.Remove('AtMs')
    $s.Observations[2].PSObject.Properties.Add([Management.Automation.PSScriptProperty]::new('AtMs',{ $script:shapeCallbacks++;return 2 }))
    P1-Need (P1-Equal (Get-OfflineSupervisionDecision $s) (Shape-Refusal invalid-observation 2))
    P1-Need ($script:shapeCallbacks -eq 0)
    Write-Output ('PASS: '+$script:p1Case)
    Write-Output 'LIMIT: shape lookup contracts only; no full regression or runtime acceptance.'
}catch{[Console]::Out.WriteLine('SHAPE-ERROR: '+$_.Exception.Message);exit 1}
