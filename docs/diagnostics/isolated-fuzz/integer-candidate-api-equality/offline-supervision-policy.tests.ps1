param([ValidateSet('all','cutoff-mutation','sticky-mutation')][string]$Group='all')
$ErrorActionPreference='Stop'
# Breaks caught: budget rebasing/equality, lost first failure, forged direction,
# passive-input execution, counter regression and hypothetical-to-real promotion.
function P1-Need([bool]$Value){if(-not $Value){throw ('ASSERTION FAILED: '+$script:p1Case)}}
function P1-Case([string]$Name,[scriptblock]$Body){
    if($Group -eq 'cutoff-mutation' -and $Name -cnotin @('ApiRequired','StopBoundary','Complete')){return}
    if($Group -eq 'sticky-mutation' -and $Name -cnotin @('ApiRequired','FailureSticky','Complete')){return}
    $script:p1Case='P1'+$Name;& $Body;Write-Output ('PASS: '+$script:p1Case)
}
function New-P1Scenario($StartMs=0,[object[]]$Observations=[object[]]@()){
    [pscustomobject]@{Schema=1;Mode='policy-only';SessionLabel='s';ClockLabel='synthetic-ms';StartMs=$StartMs;Observations=$Observations}
}
function P1-Observation($Sequence,$AtMs,[string]$Kind,[string]$Actor='W',[hashtable]$Payload=@{}){
    $v=[ordered]@{Sequence=$Sequence;SessionLabel='s';ClockLabel='synthetic-ms';AtMs=$AtMs;Actor=$Actor;Kind=$Kind}
    foreach($key in $Payload.Keys){$v[$key]=$Payload[$key]};[pscustomobject]$v
}
function P1-Arm($Time=0){P1-Observation 0 $Time arm W @{LauncherOwnsControl=$true;SupervisorOwnsControl=$true;ContainerBound=$true;HandleInheritanceBlocked=$true;ReadersReady=$true}}
function P1-Release($Time=1){P1-Observation 1 $Time release W @{InputBindingMatches=$true}}
function P1-Progress($Seq=2,$Time=2,$Out=0,$Err=0,$Count=3,$OutEOF=$true,$ErrEOF=$true,$Exit=$true,$Code=0){
    P1-Observation $Seq $Time progress C @{StdoutRetainedBytes=$Out;StderrRetainedBytes=$Err;DataEventCount=$Count;StdoutEOF=$OutEOF;StderrEOF=$ErrEOF;ExitObserved=$Exit;ExitCode=$Code}
}
function P1-Cleanup($Seq=3,$Time=3,$Verified=$true,$Natural=$true,$Counts=$true){P1-Observation $Seq $Time cleanup W @{Verified=$Verified;NaturalCompletion=$Natural;CountsReconciled=$Counts}}
function P1-Persist($Seq=4,$Time=4,$Readback=$true,$Class='clean'){P1-Observation $Seq $Time persist W @{ReadbackMatches=$Readback;Classification=$Class}}
function P1-Normal {return ,([object[]]@((P1-Arm),(P1-Release),(P1-Progress),(P1-Cleanup),(P1-Persist)))}
# Literal expected defaults, never computed by the product or its helpers.
function P1-Want([hashtable]$Root=@{},[hashtable]$Decision=@{}){
    $d=[ordered]@{DeadlineMs=120000;StopAtMs=113500;NaturalUntilMs=$null;CleanupUntilMs=$null;PublishUntilMs=$null;LastAtMs=0;LastSequence=$null;StdoutRetainedBytes=0;StderrRetainedBytes=0;DataEventCount=0;ControlSummaryCount=0;ControlSummaryByteMax=0;ControlQueueDepth=0;WouldCaptureStatus='incomplete';WouldCleanupStatus='unknown';WouldStopOwnedSession=$false;WouldNaturalCompletion=$false;WouldEvidenceComplete=$false;WouldPositiveTechnicalCandidate=$false;WouldExitObserved=$false;WouldExitCode=$null;WouldClassification='unknown'}
    foreach($k in $Decision.Keys){$d[$k]=$Decision[$k]}
    $r=[ordered]@{Schema=1;Mode='policy-only';EvidenceClass='synthetic';Evaluated=$true;Reason='trace-evaluated';State='Prepared';FailureObservationIndex=$null;Decision=[pscustomobject]$d;OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false;LiveActionsPerformed=$false;CleanupStatus='unknown'}
    foreach($k in $Root.Keys){$r[$k]=$Root[$k]};[pscustomobject]$r
}
function P1-WantNormal([hashtable]$Root=@{},[hashtable]$Decision=@{}){
    $r=@{State='Persisted'};foreach($k in $Root.Keys){$r[$k]=$Root[$k]}
    $d=@{LastAtMs=4;LastSequence=4;NaturalUntilMs=502;CleanupUntilMs=5502;PublishUntilMs=6502;DataEventCount=3;WouldCaptureStatus='complete';WouldCleanupStatus='verified';WouldNaturalCompletion=$true;WouldEvidenceComplete=$true;WouldPositiveTechnicalCandidate=$true;WouldExitObserved=$true;WouldExitCode=0;WouldClassification='clean'}
    foreach($k in $Decision.Keys){$d[$k]=$Decision[$k]};P1-Want $r $d
}
function P1-WantFailure([string]$Reason,[hashtable]$Root=@{},[hashtable]$Decision=@{}){
    $r=@{Reason=$Reason;State='Aborted';FailureObservationIndex=2};foreach($k in $Root.Keys){$r[$k]=$Root[$k]}
    $d=@{LastAtMs=2;LastSequence=2;NaturalUntilMs=502;CleanupUntilMs=5502;PublishUntilMs=6502;WouldCaptureStatus='refused';WouldStopOwnedSession=$true}
    foreach($k in $Decision.Keys){$d[$k]=$Decision[$k]};P1-Want $r $d
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
function P1-Check($Scenario,$Expected){$r=Get-OfflineSupervisionDecision $Scenario;P1-Need (P1-Equal $r $Expected);return $r}
function P1-SplitCheck4096($Scenario,$Expected){
    [Console]::Out.WriteLine('SUBPHASE: Evaluate4096 START ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
    $r=Get-OfflineSupervisionDecision $Scenario
    [Console]::Out.WriteLine('SUBPHASE: Evaluate4096 END ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
    [Console]::Out.WriteLine('SUBPHASE: Assert4096 START ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
    P1-Need (P1-Equal $r $Expected)
    [Console]::Out.WriteLine('SUBPHASE: Assert4096 END ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
    return $r
}
function P1-Bad($Scenario,[string]$Reason='invalid-scenario',$Index=$null){$null=P1-Check $Scenario (P1-Want @{Evaluated=$false;Reason=$Reason;State='Refused';FailureObservationIndex=$Index;Decision=$null})}
function P1-Running($Tail){New-P1Scenario -Observations ([object[]]@((P1-Arm),(P1-Release)) + [object[]]@($Tail))}
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=New-Object Security.Principal.WindowsPrincipal($identity)
    Write-Output ('CONTEXT: '+$identity.Name+' admin='+$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
    $module=Join-Path $PSScriptRoot 'offline-supervision-policy.ps1';if(Test-Path -LiteralPath $module){. $module}
    P1-Case ApiRequired {P1-Need ($null -ne (Get-Command Get-OfflineSupervisionDecision -CommandType Function -ErrorAction SilentlyContinue))}
    Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
    P1-Case EmptyPrepared {$null=P1-Check (New-P1Scenario) (P1-Want)}
    P1-Case NormalStages {$null=P1-Check (New-P1Scenario -Observations (P1-Normal)) (P1-WantNormal)}
    P1-Case AbsoluteDeadline {
        $a=P1-Normal;foreach($o in $a){$o.AtMs+=1000}
        $null=P1-Check (New-P1Scenario 1000 $a) (P1-WantNormal @{} @{DeadlineMs=121000;StopAtMs=114500;NaturalUntilMs=1502;CleanupUntilMs=6502;PublishUntilMs=7502;LastAtMs=1004})
    }
    P1-Case StopBoundary {
        $null=P1-Check (P1-Running (P1-Observation 2 113499 time)) (P1-Want @{State='Running'} @{LastAtMs=113499;LastSequence=2})
        foreach($t in @(113500,113501)){$null=P1-Check (P1-Running (P1-Observation 2 $t time)) (P1-WantFailure capture-cutoff @{} @{LastAtMs=$t;NaturalUntilMs=114000;CleanupUntilMs=119000;PublishUntilMs=120000;WouldCaptureStatus='timed-out'})}
    }
    P1-Case TotalDeadline {
        foreach($o in @((P1-Cleanup 2 120000),(P1-Persist 2 120000))){$null=P1-Check (P1-Running $o) (P1-WantFailure total-deadline @{} @{LastAtMs=120000;NaturalUntilMs=114000;CleanupUntilMs=119000;PublishUntilMs=120000;WouldCaptureStatus='timed-out'})}
    }
    P1-Case StartupConsumesBudget {
        $s=New-P1Scenario -Observations @((P1-Arm),(P1-Release 113499),(P1-Observation 2 113500 time))
        $null=P1-Check $s (P1-WantFailure capture-cutoff @{} @{LastAtMs=113500;NaturalUntilMs=114000;CleanupUntilMs=119000;PublishUntilMs=120000;WouldCaptureStatus='timed-out'})
    }
    P1-Case EqualTimeSequence {$a=P1-Normal;foreach($o in $a){$o.AtMs=0};$null=P1-Check (New-P1Scenario -Observations $a) (P1-WantNormal @{} @{LastAtMs=0;NaturalUntilMs=500;CleanupUntilMs=5500;PublishUntilMs=6500})}
    P1-Case BackwardTime {$null=P1-Check (P1-Running (P1-Observation 2 0 time)) (P1-WantFailure time-went-backward @{} @{LastAtMs=1;LastSequence=1;NaturalUntilMs=500;CleanupUntilMs=5500;PublishUntilMs=6500})}
    P1-Case SequenceRefusal {$null=P1-Check (P1-Running (P1-Observation 7 2 time)) (P1-WantFailure sequence-mismatch @{} @{LastAtMs=1;LastSequence=1})}
    P1-Case SessionClockRefusal {
        foreach($name in @('SessionLabel','ClockLabel')){$o=P1-Observation 2 2 time;$o.$name='changed';$null=P1-Check (P1-Running $o) (P1-WantFailure binding-mismatch @{} @{LastAtMs=1;LastSequence=1})}
    }
    P1-Case ActorDirection {
        $o=P1-Cleanup 2 2;$o.Actor='C';$null=P1-Check (P1-Running $o) (P1-WantFailure wrong-actor @{} @{LastAtMs=1;LastSequence=1})
    }
    P1-Case ArmPrerequisites {
        foreach($p in @('LauncherOwnsControl','SupervisorOwnsControl','ContainerBound','HandleInheritanceBlocked','ReadersReady')){$o=P1-Arm;$o.$p=$false;$null=P1-Check (New-P1Scenario -Observations @($o)) (P1-WantFailure prerequisites-unmet @{State='Refused';FailureObservationIndex=0} @{LastAtMs=0;LastSequence=0;NaturalUntilMs=500;CleanupUntilMs=5500;PublishUntilMs=6500;WouldStopOwnedSession=$false})}
    }
    P1-Case ReleaseOnce {
        $o=P1-Release 0;$o.Sequence=0;$null=P1-Check (New-P1Scenario -Observations @($o)) (P1-WantFailure unexpected-transition @{State='Refused';FailureObservationIndex=0} @{LastAtMs=0;LastSequence=0;NaturalUntilMs=500;CleanupUntilMs=5500;PublishUntilMs=6500;WouldStopOwnedSession=$false})
        $o=P1-Release 2;$o.Sequence=2;$null=P1-Check (P1-Running $o) (P1-WantFailure unexpected-transition)
        $o=P1-Arm 2;$o.Sequence=2;$null=P1-Check (P1-Running $o) (P1-WantFailure unexpected-transition)
        $o=P1-Release;$o.InputBindingMatches=$false;$null=P1-Check (New-P1Scenario -Observations @((P1-Arm),$o)) (P1-WantFailure prerequisites-unmet @{FailureObservationIndex=1} @{LastAtMs=1;LastSequence=1;NaturalUntilMs=501;CleanupUntilMs=5501;PublishUntilMs=6501})
    }
    P1-Case ProgressMonotonic {
        $first=P1-Progress 2 2 4 4 5 $true $false $true 7
        foreach($field in @('StdoutRetainedBytes','StderrRetainedBytes','DataEventCount','StdoutEOF','ExitObserved','ExitCode')){
            $next=P1-Progress 3 3 4 4 5 $true $false $true 7
            switch($field){StdoutRetainedBytes{$next.$field=3} StderrRetainedBytes{$next.$field=3} DataEventCount{$next.$field=4} StdoutEOF{$next.$field=$false} ExitObserved{$next.$field=$false;$next.ExitCode=$null} ExitCode{$next.$field=8}}
            $null=P1-Check (P1-Running @($first,$next)) (P1-WantFailure progress-regressed @{FailureObservationIndex=3} @{LastAtMs=3;LastSequence=3;NaturalUntilMs=503;CleanupUntilMs=5503;PublishUntilMs=6503;StdoutRetainedBytes=4;StderrRetainedBytes=4;DataEventCount=5;WouldExitObserved=$true;WouldExitCode=7})
        }
        $same=P1-Progress 3 3 4 4 5 $true $false $true 7
        $null=P1-Check (P1-Running @($first,$same)) (P1-Want @{State='Running'} @{LastAtMs=3;LastSequence=3;StdoutRetainedBytes=4;StderrRetainedBytes=4;DataEventCount=5;WouldExitObserved=$true;WouldExitCode=7})
    }
    P1-Case ExitBeforeEOF {
        $a=@((P1-Arm),(P1-Release),(P1-Progress 2 2 0 0 1 $false $false),(P1-Progress 3 3 0 0 2 $true $false),(P1-Progress 4 4 0 0 3))
        $null=P1-Check (New-P1Scenario -Observations $a) (P1-Want @{State='Stopping'} @{LastAtMs=4;LastSequence=4;DataEventCount=3;WouldCaptureStatus='complete';WouldExitObserved=$true;WouldExitCode=0;NaturalUntilMs=504;CleanupUntilMs=5504;PublishUntilMs=6504})
    }
    P1-Case UnknownExit {$null=P1-Check (P1-Running (P1-Progress 2 2 0 0 2 $true $true $false $null)) (P1-Want @{State='Running'} @{LastAtMs=2;LastSequence=2;DataEventCount=2})}
    P1-Case NonzeroExit {$a=P1-Normal;$a[2].ExitCode=7;$null=P1-Check (New-P1Scenario -Observations $a) (P1-WantNormal @{} @{WouldExitCode=7;WouldPositiveTechnicalCandidate=$false})}
    P1-Case DataBounds {
        $null=P1-Check (P1-Running (P1-Progress 2 2 1048576 1048576 4096 $false $false $false $null)) (P1-Want @{State='Running'} @{LastAtMs=2;LastSequence=2;StdoutRetainedBytes=1048576;StderrRetainedBytes=1048576;DataEventCount=4096})
        foreach($field in @('StdoutRetainedBytes','StderrRetainedBytes','DataEventCount')){$o=P1-Progress 2 2 0 0 0 $false $false $false $null;$o.$field=if($field -eq 'DataEventCount'){4097}else{1048577};$null=P1-Check (P1-Running $o) (P1-WantFailure capacity-exceeded)}
    }
    P1-Case ControlBounds {
        $p=@{SummaryCount=16;MaxSummaryBytes=1024;QueueDepth=16};$o=P1-Observation 2 2 control W $p
        $null=P1-Check (P1-Running $o) (P1-Want @{State='Running'} @{LastAtMs=2;LastSequence=2;ControlSummaryCount=16;ControlSummaryByteMax=1024;ControlQueueDepth=16})
        foreach($field in $p.Keys){$bad=@{SummaryCount=16;MaxSummaryBytes=1024;QueueDepth=16};$bad[$field]++;$null=P1-Check (P1-Running (P1-Observation 2 2 control W $bad)) (P1-WantFailure capacity-exceeded)}
        $next=P1-Observation 3 3 control W @{SummaryCount=15;MaxSummaryBytes=1024;QueueDepth=0}
        $null=P1-Check (P1-Running @($o,$next)) (P1-WantFailure progress-regressed @{FailureObservationIndex=3} @{LastAtMs=3;LastSequence=3;NaturalUntilMs=503;CleanupUntilMs=5503;PublishUntilMs=6503;ControlSummaryCount=16;ControlSummaryByteMax=1024;ControlQueueDepth=16})
    }
    P1-Case ReadFault {$null=P1-Check (P1-Running (P1-Observation 2 2 fault C @{Fault='read-error'})) (P1-WantFailure read-error @{} @{WouldCaptureStatus='fault'})}
    P1-Case SupervisorLost {$null=P1-Check (P1-Running (P1-Observation 2 2 fault L @{Fault='supervisor-lost'})) (P1-WantFailure supervisor-lost @{} @{WouldCaptureStatus='incomplete'})}
    P1-Case LauncherLost {$null=P1-Check (P1-Running (P1-Observation 2 2 fault W @{Fault='launcher-lost'})) (P1-WantFailure launcher-lost @{} @{WouldCaptureStatus='incomplete'})}
    P1-Case FailureSticky {
        $a=@((P1-Arm),(P1-Release),(P1-Observation 2 2 fault L @{Fault='supervisor-lost'}),(P1-Cleanup),(P1-Persist))
        $null=P1-Check (New-P1Scenario -Observations $a) (P1-WantFailure supervisor-lost @{} @{LastAtMs=4;LastSequence=4;WouldCaptureStatus='incomplete';WouldCleanupStatus='verified';WouldNaturalCompletion=$true;WouldClassification='clean'})
    }
    P1-Case PhaseBudgets {
        foreach($row in @(@{At=501;Reason='trace-evaluated';State='Stopping';Index=$null},@{At=502;Reason='natural-window-expired';State='Aborted';Index=3},@{At=5501;Reason='natural-window-expired';State='Aborted';Index=3},@{At=5502;Reason='phase-deadline';State='Aborted';Index=3})){
            $want=P1-Want @{State=$row.State;Reason=$row.Reason;FailureObservationIndex=$row.Index} @{LastAtMs=$row.At;LastSequence=3;NaturalUntilMs=502;CleanupUntilMs=5502;PublishUntilMs=6502;DataEventCount=3;WouldCaptureStatus='complete';WouldExitObserved=$true;WouldExitCode=0;WouldStopOwnedSession=($null -ne $row.Index)}
            $null=P1-Check (New-P1Scenario -Observations @((P1-Arm),(P1-Release),(P1-Progress),(P1-Observation 3 $row.At time))) $want
        }
        $a=P1-Normal;$a[4].AtMs=6501;$null=P1-Check (New-P1Scenario -Observations $a) (P1-WantNormal @{} @{LastAtMs=6501})
        $a[4].AtMs=6502;$null=P1-Check (New-P1Scenario -Observations $a) (P1-WantFailure phase-deadline @{FailureObservationIndex=4} @{LastAtMs=6502;LastSequence=4;DataEventCount=3;WouldCaptureStatus='complete';WouldCleanupStatus='verified';WouldNaturalCompletion=$true;WouldExitObserved=$true;WouldExitCode=0})
        $start=[long]::MaxValue-120000;$a=P1-Normal;foreach($o in $a){$o.AtMs=[long]($start+$o.AtMs)}
        $null=P1-Check (New-P1Scenario $start $a) (P1-WantNormal @{} @{DeadlineMs=[long]::MaxValue;StopAtMs=([long]::MaxValue-6500);NaturalUntilMs=([long]::MaxValue-119498);CleanupUntilMs=([long]::MaxValue-114498);PublishUntilMs=([long]::MaxValue-113498);LastAtMs=([long]::MaxValue-119996)})
    }
    P1-Case CleanupUnknown {
        foreach($field in @('Verified','CountsReconciled')){$o=P1-Cleanup;$o.$field=$false;$null=P1-Check (New-P1Scenario -Observations @((P1-Arm),(P1-Release),(P1-Progress),$o)) (P1-WantFailure cleanup-unverified @{FailureObservationIndex=3} @{LastAtMs=3;LastSequence=3;DataEventCount=3;WouldCaptureStatus='complete';WouldExitObserved=$true;WouldExitCode=0})}
        $a=P1-Normal;$a[3].NaturalCompletion=$false;$null=P1-Check (New-P1Scenario -Observations $a) (P1-WantNormal @{} @{WouldNaturalCompletion=$false;WouldPositiveTechnicalCandidate=$false})
    }
    P1-Case PersistRefused {
        foreach($c in @('refused','unknown')){$a=P1-Normal;$a[4].Classification=$c;$null=P1-Check (New-P1Scenario -Observations $a) (P1-WantNormal @{} @{WouldClassification=$c;WouldPositiveTechnicalCandidate=$false})}
        $a=P1-Normal;$a[4].ReadbackMatches=$false;$null=P1-Check (New-P1Scenario -Observations $a) (P1-WantFailure persist-unverified @{FailureObservationIndex=4} @{LastAtMs=4;LastSequence=4;DataEventCount=3;WouldCaptureStatus='complete';WouldCleanupStatus='verified';WouldNaturalCompletion=$true;WouldExitObserved=$true;WouldExitCode=0;WouldClassification='clean'})
    }
    P1-Case TerminalTrailing {$a=[object[]](P1-Normal)+@((P1-Observation 5 5 time));$null=P1-Check (New-P1Scenario -Observations $a) (P1-WantFailure trailing-after-terminal @{FailureObservationIndex=5} @{LastAtMs=5;LastSequence=5;DataEventCount=3;WouldCaptureStatus='complete';WouldCleanupStatus='verified';WouldNaturalCompletion=$true;WouldExitObserved=$true;WouldExitCode=0;WouldClassification='clean'})}
    P1-Case InvalidScenario {
        P1-Bad $null;$s=New-P1Scenario;$s.Mode='live';P1-Bad $s;$s=New-P1Scenario;$s|Add-Member NoteProperty Extra 1;P1-Bad $s
        foreach($label in @('',('x'*65),"s`n",('s'+[char]0))){$s=New-P1Scenario;$s.SessionLabel=$label;P1-Bad $s}
        $s=New-P1Scenario;$s.PSObject.Properties.Remove('StartMs');P1-Bad $s
    }
    P1-Case InvalidObservation {
        $s=New-P1Scenario -Observations @((P1-Observation 0 120000 time),[pscustomobject]@{Kind='bad'});P1-Bad $s invalid-observation 1
        foreach($kind in @('Time',('time'+[char]0))){$s=New-P1Scenario -Observations @((P1-Observation 0 0 $kind));P1-Bad $s invalid-observation 0}
        $o=P1-Progress;$o.ExitCode=$null;P1-Bad (P1-Running $o) invalid-observation 2
        $o=P1-Observation 0 0 time;$o|Add-Member NoteProperty Extra 1;P1-Bad (New-P1Scenario -Observations @($o)) invalid-observation 0
    }
    P1-Case PassiveDescriptors {
        $script:p1Getter=0;$s=New-P1Scenario;$s.PSObject.Properties.Remove('StartMs');$s|Add-Member ScriptProperty StartMs {$script:p1Getter++;0};P1-Bad $s;P1-Need ($script:p1Getter -eq 0)
        $o=P1-Observation 0 0 time;$o.PSObject.Properties.Remove('AtMs');$o|Add-Member ScriptProperty AtMs {$script:p1Getter++;0};P1-Bad (New-P1Scenario -Observations @($o)) invalid-observation 0
        $s=New-P1Scenario;$s|Add-Member ScriptMethod GetType {$script:p1Getter++;throw 'not-called'} -Force;$null=P1-Check $s (P1-Want)
        $s=New-P1Scenario;$v=[pscustomobject]@{};$v|Add-Member ScriptMethod ToString {$script:p1Getter++;'0'} -Force;$s.StartMs=$v;P1-Bad $s;P1-Need ($script:p1Getter -eq 0)
    }
    P1-Case ArrayAndIntegerBounds {
        [Console]::Out.WriteLine('BOUNDARY: InvalidInputs START ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
        foreach($v in @('0',$true,1.0,-1,[long]::MaxValue)){$s=New-P1Scenario;$s.StartMs=$v;P1-Bad $s}
        foreach($v in @($null,(New-Object 'object[,]' 1,1),([Array]::CreateInstance([object],[int[]]@(1),[int[]]@(1))))){$s=New-P1Scenario;$s.Observations=$v;P1-Bad $s}
        $s=New-P1Scenario;$null=Add-Member -InputObject $s.Observations -MemberType NoteProperty -Name Length -Value 0 -Force;try{P1-Bad $s}finally{$s.Observations.PSObject.Properties.Remove('Length')}
        [Console]::Out.WriteLine('BOUNDARY: InvalidInputs END ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
        [Console]::Out.WriteLine('BOUNDARY: Build4096 START ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
        $many=New-Object object[] 4096;for($i=0;$i -lt 4096;$i++){$many[$i]=P1-Observation $i 0 time}
        [Console]::Out.WriteLine('BOUNDARY: Build4096 END ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
        [Console]::Out.WriteLine('BOUNDARY: Check4096 START ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
        $null=P1-SplitCheck4096 (New-P1Scenario -Observations $many) (P1-Want @{} @{LastAtMs=0;LastSequence=4095})
        [Console]::Out.WriteLine('BOUNDARY: Check4096 END ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
        [Console]::Out.WriteLine('BOUNDARY: Check4097 START ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
        P1-Bad (New-P1Scenario -Observations ([object[]]($many+@((P1-Observation 4096 0 time)))))
        [Console]::Out.WriteLine('BOUNDARY: Check4097 END ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
    }
    P1-Case CallerMutation {
        $s=New-P1Scenario -Observations (P1-Normal);$r=P1-Check $s (P1-WantNormal);$s.StartMs=9;$s.Observations[0].ReadersReady=$false
        P1-Need (P1-Equal $r (P1-WantNormal));$r.Decision.WouldPositiveTechnicalCandidate=$false;$next=Get-OfflineSupervisionDecision (New-P1Scenario -Observations (P1-Normal));P1-Need (P1-Equal $next (P1-WantNormal));P1-Need (-not [object]::ReferenceEquals($r.Decision,$next.Decision))
    }
    P1-Case ConservativeFlags {
        foreach($s in @((New-P1Scenario),(New-P1Scenario -Observations (P1-Normal)),(P1-Running (P1-Observation 2 2 fault C @{Fault='read-error'})))){
            $r=Get-OfflineSupervisionDecision $s;P1-Need (@($r.PSObject.Properties).Count -eq 13 -and $r.Mode -ceq 'policy-only' -and $r.EvidenceClass -ceq 'synthetic' -and $r.CleanupStatus -ceq 'unknown')
            foreach($name in @('OriginVerified','Acceptance','RuntimeApproved','LiveActionsPerformed')){P1-Need ($r.$name -is [bool] -and -not $r.$name)}
        }
    }
    P1-Case Complete {P1-Need $true}
    Write-Output 'LIMIT: synthetic supervision policy only;no OS,real clock,pipes,containment or cleanup proof.'
}catch{
    if($_.Exception.Message -cmatch '^ASSERTION FAILED: P1[A-Za-z0-9]+$'){Write-Output $_.Exception.Message;exit 1}
    Write-Error $_;exit 1
}
