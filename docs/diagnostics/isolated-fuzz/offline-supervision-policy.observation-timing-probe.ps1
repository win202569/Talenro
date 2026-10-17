# Diagnostic spike only. No original test-script execution or candidate edits.
$ErrorActionPreference='Stop'
function P1I1-Need([bool]$Value){if(-not $Value){throw ('ASSERTION FAILED: '+$script:p1i1Case)}}
function P1I1-Case([string]$Name,[scriptblock]$Body){$script:p1i1Case='P1I1'+$Name;& $Body;Write-Output ('PASS: '+$script:p1i1Case)}
function P1I1-Equal($Actual,$Expected,[int]$Depth=0){
    if($Depth -gt 4){return $false};if($null -eq $Expected){return ($null -eq $Actual)}
    if($Expected -is [pscustomobject]){
        if($Actual -isnot [pscustomobject] -or @($Actual.PSObject.Properties).Count -ne @($Expected.PSObject.Properties).Count){return $false}
        foreach($p in $Expected.PSObject.Properties){$q=@($Actual.PSObject.Properties|Where-Object {[string]::Equals($_.Name,$p.Name,[StringComparison]::Ordinal)});if($q.Count -ne 1 -or $q[0].MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty -or -not (P1I1-Equal $q[0].Value $p.Value ($Depth+1))){return $false}};return $true
    }
    if($Expected -is [string]){return ($Actual -is [string] -and [string]::Equals($Actual,$Expected,[StringComparison]::Ordinal))}
    if($Expected -is [bool]){return ($Actual -is [bool] -and $Actual -eq $Expected)}
    if($Expected -is [int] -or $Expected -is [long]){return (($Actual -is [int] -or $Actual -is [long]) -and $Actual -eq $Expected)}
    return $false
}
function P1I1-Scenario([object[]]$Observations=[object[]]@()){
    [pscustomobject]@{Schema=1;Mode='policy-only';SessionLabel='s';ClockLabel='synthetic-ms';StartMs=0;Observations=$Observations}
}
function P1I1-Observation($Sequence,$At,[string]$Kind,[string]$Actor,[hashtable]$Payload=@{}){
    $v=[ordered]@{Sequence=$Sequence;SessionLabel='s';ClockLabel='synthetic-ms';AtMs=$At;Actor=$Actor;Kind=$Kind};foreach($k in $Payload.Keys){$v[$k]=$Payload[$k]};[pscustomobject]$v
}
function P1I1-Normal {
    P1I1-Scenario ([object[]]@(
        (P1I1-Observation 0 0 arm W @{LauncherOwnsControl=$true;SupervisorOwnsControl=$true;ContainerBound=$true;HandleInheritanceBlocked=$true;ReadersReady=$true}),
        (P1I1-Observation 1 1 release W @{InputBindingMatches=$true}),
        (P1I1-Observation 2 2 progress C @{StdoutRetainedBytes=0;StderrRetainedBytes=0;DataEventCount=3;StdoutEOF=$true;StderrEOF=$true;ExitObserved=$true;ExitCode=0}),
        (P1I1-Observation 3 3 cleanup W @{Verified=$true;NaturalCompletion=$true;CountsReconciled=$true}),
        (P1I1-Observation 4 4 persist W @{ReadbackMatches=$true;Classification='clean'})
    ))
}
function P1I1-Refusal([string]$Reason='invalid-observation',$Index=2){
    [pscustomobject]@{Schema=1;Mode='policy-only';EvidenceClass='synthetic';Evaluated=$false;Reason=$Reason;State='Refused';FailureObservationIndex=$Index;Decision=$null;OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false;LiveActionsPerformed=$false;CleanupStatus='unknown'}
}
function P1I1-Type($Value){[object].GetMethod('GetType').Invoke($Value.PSObject.BaseObject,$null)}
function P1I1-WrapPrimitive([object]$Value){
    if($null -eq $Value){throw 'FIXTURE ERROR: null is not a wrapped primitive'}
    # Decorate owned copies, never cached literals used by plain control scenarios.
    $type=P1I1-Type $Value
    if($type -eq [string]){$fresh=[string]::Copy($Value.PSObject.BaseObject)}
    elseif($type -eq [int] -or $type -eq [long] -or $type -eq [bool]){
        $storage=[Array]::CreateInstance($type,1);$storage.SetValue($Value.PSObject.BaseObject,0);$fresh=$storage.GetValue(0)
    }else{throw 'FIXTURE ERROR: unsupported primitive'}
    $ctor=[Management.Automation.PSObject].GetConstructor([type[]]@([object]))
    if($null -eq $ctor -or -not $ctor.IsPublic){throw 'FIXTURE ERROR: object constructor unavailable'}
    $parameters=@($ctor.GetParameters())
    if($parameters.Count -ne 1 -or $parameters[0].ParameterType -ne [object]){throw 'FIXTURE ERROR: exact object signature required'}
    return ,($ctor.Invoke([object[]]@($fresh)))
}
function P1I1-Decorate($Owner,[string]$Name){
    $original=$Owner.$Name;$type=P1I1-Type $original
    $wrapped=P1I1-WrapPrimitive $original
    $metadata=[object[]]@('initial')
    $wrapped.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Metadata',$metadata))
    $Owner.$Name=$wrapped;$field=$Owner.$Name;$property=$field.PSObject.Properties['Metadata']
    if((P1I1-Type $field) -ne $type -or $field.PSObject.BaseObject -cne $original.PSObject.BaseObject -or $null -eq $property -or $property.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty -or -not [object]::ReferenceEquals($property.Value,$metadata)){throw 'FIXTURE ERROR: decorated scalar was not preserved'}
    return ,$metadata
}
function ObservationTiming-Core($Scenario,$Owner,[string]$Field,[string]$Reason='invalid-observation',$Index=2){
    $null=P1I1-Decorate $Owner $Field
    $result=Get-OfflineSupervisionDecision $Scenario;P1I1-Need (P1I1-Equal $result (P1I1-Refusal $Reason $Index))
}

function P1I1-CheckDecorated($Scenario,$Owner,[string]$Field,[string]$Reason='invalid-observation',$Index=2){
    $script:obsTimingStarted++
    if($script:obsTimingStarted -gt 65){throw 'TIMING ERROR: unexpected check count'}
    $id=$script:obsTimingStarted;$kind=[string]$Owner.Kind
    $start=[pscustomobject]@{Schema=1;Event='start';Id=$id;ObservationKind=$kind;Field=$Field;Index=$Index;Reason=$Reason;AtMs=$script:obsClock.ElapsedMilliseconds;ElapsedMs=$null}
    Write-Output ('TIMING: '+($start|ConvertTo-Json -Compress))
    $timer=[Diagnostics.Stopwatch]::StartNew()
    ObservationTiming-Core $Scenario $Owner $Field $Reason $Index
    $timer.Stop();$script:obsTimingCompleted++
    $end=[pscustomobject]@{Schema=1;Event='end';Id=$id;ObservationKind=$kind;Field=$Field;Index=$Index;Reason=$Reason;AtMs=$script:obsClock.ElapsedMilliseconds;ElapsedMs=[Math]::Round($timer.Elapsed.TotalMilliseconds,3)}
    Write-Output ('TIMING: '+($end|ConvertTo-Json -Compress))
}
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=[Security.Principal.WindowsPrincipal]::new($identity)
    Write-Output ('CONTEXT: '+$identity.Name+' admin='+$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
    Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
    $script:obsTimingStarted=0;$script:obsTimingCompleted=0;$script:obsClock=[Diagnostics.Stopwatch]::StartNew()
    . (Join-Path $PSScriptRoot 'offline-supervision-policy.ps1')
P1I1-Case ObservationScalars {
        foreach($index in 0..4){$template=P1I1-Normal;foreach($p in $template.Observations[$index].PSObject.Properties){$s=P1I1-Normal;P1I1-CheckDecorated $s $s.Observations[$index] $p.Name invalid-observation $index}}
        foreach($kind in @('control','fault')){
            $payload=if($kind -ceq 'control'){@{SummaryCount=0;MaxSummaryBytes=0;QueueDepth=0}}else{@{Fault='supervisor-lost'}}
            $actor=if($kind -ceq 'control'){'W'}else{'L'};$template=P1I1-Observation 0 0 $kind $actor $payload
            foreach($p in $template.PSObject.Properties){$o=P1I1-Observation 0 0 $kind $actor $payload;$s=P1I1-Scenario ([object[]]@($o));P1I1-CheckDecorated $s $o $p.Name invalid-observation 0}
        }
        $s=P1I1-Normal;$s.Observations[2].AtMs=113500;P1I1-CheckDecorated $s $s.Observations[4] Classification invalid-observation 4
    }

    $summary=[pscustomobject]@{Schema=1;Expected=65;Started=$script:obsTimingStarted;Completed=$script:obsTimingCompleted;ElapsedMs=$script:obsClock.ElapsedMilliseconds;OriginalTestsRun=$false;FixturePreflightRun=$false;ProductChanged=$false}
    Write-Output ('TIMING-SUMMARY: '+($summary|ConvertTo-Json -Compress))
    Write-Output 'LIMIT: ObservationScalars timing only; no full suite, retries, repairs or budget renewal.'
}catch{
    $failure=[pscustomobject]@{Message=$_.Exception.Message;Started=$script:obsTimingStarted;Completed=$script:obsTimingCompleted}
    Write-Output ('PROBE-ERROR: '+($failure|ConvertTo-Json -Compress));exit 1
}
