param([ValidateSet('regression','all')][string]$Group='regression')
$ErrorActionPreference='Stop'
# Catches scalar metadata admission, callback evaluation and cross-call aliasing.
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
function P1I1-Want([string]$Kind='normal'){
    $d=[pscustomobject]@{DeadlineMs=120000;StopAtMs=113500;NaturalUntilMs=502;CleanupUntilMs=5502;PublishUntilMs=6502;LastAtMs=4;LastSequence=4;StdoutRetainedBytes=0;StderrRetainedBytes=0;DataEventCount=3;ControlSummaryCount=0;ControlSummaryByteMax=0;ControlQueueDepth=0;WouldCaptureStatus='complete';WouldCleanupStatus='verified';WouldStopOwnedSession=$false;WouldNaturalCompletion=$true;WouldEvidenceComplete=$true;WouldPositiveTechnicalCandidate=$true;WouldExitObserved=$true;WouldExitCode=0;WouldClassification='clean'}
    $r=[pscustomobject]@{Schema=1;Mode='policy-only';EvidenceClass='synthetic';Evaluated=$true;Reason='trace-evaluated';State='Persisted';FailureObservationIndex=$null;Decision=$d;OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false;LiveActionsPerformed=$false;CleanupStatus='unknown'}
    if($Kind -ceq 'prepared'){$r.State='Prepared';$d.NaturalUntilMs=$null;$d.CleanupUntilMs=$null;$d.PublishUntilMs=$null;$d.LastAtMs=0;$d.LastSequence=$null;$d.DataEventCount=0;$d.WouldCaptureStatus='incomplete';$d.WouldCleanupStatus='unknown';$d.WouldNaturalCompletion=$false;$d.WouldEvidenceComplete=$false;$d.WouldPositiveTechnicalCandidate=$false;$d.WouldExitObserved=$false;$d.WouldExitCode=$null;$d.WouldClassification='unknown'}
    if($Kind -ceq 'sticky'){$r.State='Aborted';$r.Reason='supervisor-lost';$r.FailureObservationIndex=2;$d.DataEventCount=0;$d.WouldCaptureStatus='incomplete';$d.WouldStopOwnedSession=$true;$d.WouldEvidenceComplete=$false;$d.WouldPositiveTechnicalCandidate=$false;$d.WouldExitObserved=$false;$d.WouldExitCode=$null}
    if($Kind -ceq 'unknown'){$r.State='Running';$d.NaturalUntilMs=$null;$d.CleanupUntilMs=$null;$d.PublishUntilMs=$null;$d.LastAtMs=2;$d.LastSequence=2;$d.WouldCaptureStatus='incomplete';$d.WouldCleanupStatus='unknown';$d.WouldNaturalCompletion=$false;$d.WouldEvidenceComplete=$false;$d.WouldPositiveTechnicalCandidate=$false;$d.WouldExitObserved=$false;$d.WouldExitCode=$null;$d.WouldClassification='unknown'}
    return $r
}
function P1I1-Refusal([string]$Reason='invalid-observation',$Index=2){
    [pscustomobject]@{Schema=1;Mode='policy-only';EvidenceClass='synthetic';Evaluated=$false;Reason=$Reason;State='Refused';FailureObservationIndex=$Index;Decision=$null;OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false;LiveActionsPerformed=$false;CleanupStatus='unknown'}
}
function P1I1-Type($Value){[object].GetMethod('GetType').Invoke($Value.PSObject.BaseObject,$null)}
function P1I1-Decorate($Owner,[string]$Name){
    $original=$Owner.$Name;$type=P1I1-Type $original
    $wrapped=[Management.Automation.PSObject]::new($original)
    $metadata=[object[]]@('initial')
    $wrapped.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new('Metadata',$metadata))
    $Owner.$Name=$wrapped;$field=$Owner.$Name;$property=$field.PSObject.Properties['Metadata']
    if((P1I1-Type $field) -ne $type -or $field.PSObject.BaseObject -cne $original.PSObject.BaseObject -or $null -eq $property -or $property.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty -or -not [object]::ReferenceEquals($property.Value,$metadata)){throw 'FIXTURE ERROR: decorated scalar was not preserved'}
    return ,$metadata
}
function P1I1-CheckDecorated($Scenario,$Owner,[string]$Field,[string]$Reason='invalid-observation',$Index=2){
    $null=P1I1-Decorate $Owner $Field
    $result=Get-OfflineSupervisionDecision $Scenario;P1I1-Need (P1I1-Equal $result (P1I1-Refusal $Reason $Index))
}
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=New-Object Security.Principal.WindowsPrincipal($identity)
    Write-Output ('CONTEXT: '+$identity.Name+' admin='+$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
    Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
    if($Group -ceq 'all'){& (Join-Path $PSScriptRoot 'offline-supervision-policy.tests.ps1') -Group all;if(-not $?){throw 'Prior P1 suite failed'}}
    . (Join-Path $PSScriptRoot 'offline-supervision-policy.ps1')
    P1I1-Case ApiRequired {P1I1-Need ($null -ne (Get-Command Get-OfflineSupervisionDecision -CommandType Function -ErrorAction SilentlyContinue))}
    P1I1-Case DecoratedInteger {
        foreach($code in @([int]0,[long]0)){$s=P1I1-Normal;$s.Observations[2].ExitCode=$code;P1I1-CheckDecorated $s $s.Observations[2] ExitCode}
    }
    P1I1-Case DecoratedBoolean {
        foreach($natural in @($true,$false)){$s=P1I1-Normal;$s.Observations[3].NaturalCompletion=$natural;P1I1-CheckDecorated $s $s.Observations[3] NaturalCompletion invalid-observation 3}
    }
    P1I1-Case DecoratedString {
        foreach($class in @('clean','refused','unknown')){$s=P1I1-Normal;$s.Observations[4].Classification=$class;P1I1-CheckDecorated $s $s.Observations[4] Classification invalid-observation 4}
    }
    P1I1-Case RootScalars {
        foreach($name in @('Schema','Mode','SessionLabel','ClockLabel','StartMs')){$s=P1I1-Normal;P1I1-CheckDecorated $s $s $name invalid-scenario $null}
    }
    P1I1-Case ObservationScalars {
        foreach($index in 0..4){$template=P1I1-Normal;foreach($p in $template.Observations[$index].PSObject.Properties){$s=P1I1-Normal;P1I1-CheckDecorated $s $s.Observations[$index] $p.Name invalid-observation $index}}
        foreach($kind in @('control','fault')){
            $payload=if($kind -ceq 'control'){@{SummaryCount=0;MaxSummaryBytes=0;QueueDepth=0}}else{@{Fault='supervisor-lost'}}
            $actor=if($kind -ceq 'control'){'W'}else{'L'};$template=P1I1-Observation 0 0 $kind $actor $payload
            foreach($p in $template.PSObject.Properties){$o=P1I1-Observation 0 0 $kind $actor $payload;$s=P1I1-Scenario ([object[]]@($o));P1I1-CheckDecorated $s $o $p.Name invalid-observation 0}
        }
        $s=P1I1-Normal;$s.Observations[2].AtMs=113500;P1I1-CheckDecorated $s $s.Observations[4] Classification invalid-observation 4
    }
    P1I1-Case PassiveDescriptors {
        $script:p1i1Callbacks=0
        foreach($member in @('ScriptProperty','AliasProperty','ScriptMethod')){
            $s=P1I1-Normal;$w=[Management.Automation.PSObject]::new([int]0)
            switch($member){
                ScriptProperty {$m=[Management.Automation.PSScriptProperty]::new('Probe',{ $script:p1i1Callbacks++;return 'called' })}
                AliasProperty {$m=[Management.Automation.PSAliasProperty]::new('Probe','MissingTarget')}
                ScriptMethod {$m=[Management.Automation.PSScriptMethod]::new('Probe',{ $script:p1i1Callbacks++;return 'called' })}
            }
            $w.PSObject.Members.Add($m);$s.Observations[2].ExitCode=$w
            $descriptor=$s.Observations[2].ExitCode.PSObject.Members['Probe']
            if($null -eq $descriptor -or $descriptor.MemberType.ToString() -cne $member){throw 'FIXTURE ERROR: executable descriptor lost'}
            $r=Get-OfflineSupervisionDecision $s;P1I1-Need (P1I1-Equal $r (P1I1-Refusal));P1I1-Need ($script:p1i1Callbacks -eq 0)
        }
    }
    P1I1-Case SequentialIsolation {
        foreach($case in @(@{Index=2;Field='ExitCode'},@{Index=3;Field='NaturalCompletion'},@{Index=4;Field='Classification'})){
            $s=P1I1-Normal;$meta=P1I1-Decorate $s.Observations[$case.Index] $case.Field
            $want=P1I1-Refusal invalid-observation $case.Index;$a=Get-OfflineSupervisionDecision $s;$b=Get-OfflineSupervisionDecision $s
            P1I1-Need ((P1I1-Equal $a $want) -and (P1I1-Equal $b $want) -and -not [object]::ReferenceEquals($a,$b))
            $meta[0]='changed';P1I1-Need ((P1I1-Equal $a $want) -and (P1I1-Equal $b $want))
            $a.Reason='caller-edited';P1I1-Need (P1I1-Equal $b $want);P1I1-Need (P1I1-Equal (Get-OfflineSupervisionDecision (P1I1-Normal)) (P1I1-Want))
        }
    }
    P1I1-Case PlainControls {
        $s=P1I1-Normal;P1I1-Need ($s.SessionLabel.Length -eq 1);P1I1-Need (P1I1-Equal (Get-OfflineSupervisionDecision $s) (P1I1-Want))
        $s=P1I1-Normal;$s.Schema=[long]1;$s.StartMs=[long]0;$s.Observations[2].ExitCode=[long]0;P1I1-Need (P1I1-Equal (Get-OfflineSupervisionDecision $s) (P1I1-Want))
        P1I1-Need (P1I1-Equal (Get-OfflineSupervisionDecision (P1I1-Scenario)) (P1I1-Want prepared))
        $s=P1I1-Normal;$s.Observations=$s.Observations[0..2];$s.Observations[2].ExitObserved=$false;$s.Observations[2].ExitCode=$null;P1I1-Need (P1I1-Equal (Get-OfflineSupervisionDecision $s) (P1I1-Want unknown))
        $s=P1I1-Normal;$s.Observations[2]=P1I1-Observation 2 2 fault L @{Fault='supervisor-lost'};P1I1-Need (P1I1-Equal (Get-OfflineSupervisionDecision $s) (P1I1-Want sticky))
    }
    P1I1-Case Complete {P1I1-Need $true}
    Write-Output 'LIMIT: decorated-scalar refusal and sequential isolation only; no live supervision, concurrency or runtime acceptance.'
}catch{
    if($_.Exception.Message -cmatch '^ASSERTION FAILED: P1I1[A-Za-z0-9]+$'){Write-Output $_.Exception.Message;exit 1}
    Write-Error $_;exit 1
}
