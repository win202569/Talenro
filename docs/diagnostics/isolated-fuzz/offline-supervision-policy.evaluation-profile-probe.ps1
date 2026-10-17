# Diagnostic spike: instrumented independent copy only; NOT a product module.
$ErrorActionPreference='Stop'
function Profile-Ms([long]$Ticks){[Math]::Round(($Ticks*1000.0/[Diagnostics.Stopwatch]::Frequency),3)}
function Profile-Snapshot([string]$Phase){
    $s=$script:profileStats
    $vTicks=$s.ValidationTicks;$vOwn=$s.ValidationOwnTicks;$dTicks=$s.DecisionTicks
    if($s.Active -ceq 'ValidationLoop'){
        $open=[Diagnostics.Stopwatch]::GetTimestamp()-$s.OpenAt
        $vTicks+=$open
        $vOwn+=($open-($s.TypeTicks-$s.OpenType)-($s.ExtTicks-$s.OpenExt)-($s.ShapeOwnTicks-$s.OpenShapeOwn))
    }elseif($s.Active -ceq 'DecisionLoop'){$dTicks+=([Diagnostics.Stopwatch]::GetTimestamp()-$s.OpenAt)}
    $categories=@(
        [pscustomobject]@{Name='Type';Calls=$s.TypeCalls;InclusiveMs=(Profile-Ms $s.TypeTicks);OwnMs=(Profile-Ms $s.TypeTicks)},
        [pscustomobject]@{Name='ExtendedMembers';Calls=$s.ExtCalls;InclusiveMs=(Profile-Ms $s.ExtTicks);OwnMs=(Profile-Ms $s.ExtTicks)},
        [pscustomobject]@{Name='Shape';Calls=$s.ShapeCalls;InclusiveMs=(Profile-Ms $s.ShapeTicks);OwnMs=(Profile-Ms $s.ShapeOwnTicks)},
        [pscustomobject]@{Name='ValidationLoop';Calls=$s.ValidationCalls;InclusiveMs=(Profile-Ms $vTicks);OwnMs=(Profile-Ms $vOwn)},
        [pscustomobject]@{Name='DecisionLoop';Calls=$s.DecisionCalls;InclusiveMs=(Profile-Ms $dTicks);OwnMs=(Profile-Ms $dTicks)}
    )
    $script:profileSeq++
    $frame=[pscustomobject]@{Schema=1;Seq=$script:profileSeq;Phase=$Phase;ValidatedRows=$s.ValidationRows;DecisionRows=$s.DecisionRows;Active=$s.Active;ProductCalls=$script:profileCalls;AtMs=$script:profileClock.ElapsedMilliseconds;Categories=$categories}
    [Console]::Out.WriteLine('PROFILE: '+($frame|ConvertTo-Json -Depth 6 -Compress));[Console]::Out.Flush()
}
# BEGIN INSTRUMENTED MODULE COPY
# Pure synthetic supervision decisions. No imports, I/O, clocks or live adapters.
function Get-SupervisionPolicyType($Value){
$profTypeAt=[Diagnostics.Stopwatch]::GetTimestamp()
try{if($null -eq $Value){return $null};return [object].GetMethod('GetType').Invoke($Value.PSObject.BaseObject,$null)}finally{
$script:profileStats.TypeTicks+=([Diagnostics.Stopwatch]::GetTimestamp()-$profTypeAt)
$script:profileStats.TypeCalls++
}
}
function Assert-SupervisionPolicyPassive($Value){
    if($null -eq $Value){return}
    $type=Get-SupervisionPolicyType $Value
    if($type -eq [int] -or $type -eq [long] -or $type -eq [bool] -or $type -eq [string]){
$profExtAt=[Diagnostics.Stopwatch]::GetTimestamp()
try{
        # Use the fixed SDK getter so an added member named Members cannot shadow it.
        $view=$Value.psextended
        $members=[Management.Automation.PSMemberSet].GetProperty('Members').GetValue($view.PSObject.BaseObject,$null)
        foreach($member in $members){throw 'Decorated scalar'}
    }finally{
$script:profileStats.ExtTicks+=([Diagnostics.Stopwatch]::GetTimestamp()-$profExtAt)
$script:profileStats.ExtCalls++
}
}
    foreach($p in $Value.PSObject.Properties){
        if($p.MemberType -in @([Management.Automation.PSMemberTypes]::ScriptProperty,[Management.Automation.PSMemberTypes]::CodeProperty)){throw 'Executable descriptor'}
        if(([StringComparer]::OrdinalIgnoreCase.Equals($p.Name,'Length') -or [StringComparer]::OrdinalIgnoreCase.Equals($p.Name,'Rank')) -and $p.MemberType -ne [Management.Automation.PSMemberTypes]::Property){throw 'Shadow bound'}
    }
}
function Assert-SupervisionPolicyShape($Value,[string[]]$Names){
$profShapeType=$script:profileStats.TypeTicks;$profShapeExt=$script:profileStats.ExtTicks
$profShapeAt=[Diagnostics.Stopwatch]::GetTimestamp()
try{
    $type=Get-SupervisionPolicyType $Value
    if($null -eq $type -or $type.FullName -cne 'System.Management.Automation.PSCustomObject'){throw 'Plain object required'}
    $props=@($Value.PSObject.Properties);if($props.Count -ne $Names.Count){throw 'Exact fields required'}
    foreach($p in $props){
        $found=$false;foreach($name in $Names){if([string]::Equals($p.Name,$name,[StringComparison]::Ordinal)){$found=$true;break}}
        if(-not $found -or $p.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty){throw 'Exact passive fields required'}
        Assert-SupervisionPolicyPassive $p.Value
    }
}finally{
$profShapeElapsed=[Diagnostics.Stopwatch]::GetTimestamp()-$profShapeAt
$script:profileStats.ShapeTicks+=$profShapeElapsed
$script:profileStats.ShapeOwnTicks+=($profShapeElapsed-($script:profileStats.TypeTicks-$profShapeType)-($script:profileStats.ExtTicks-$profShapeExt))
$script:profileStats.ShapeCalls++
}
}
function Assert-SupervisionPolicyInteger($Value,[long]$Min=0,[long]$Max=[long]::MaxValue){
    $type=Get-SupervisionPolicyType $Value;if($type -ne [int] -and $type -ne [long]){throw 'Exact integer required'}
    Assert-SupervisionPolicyPassive $Value;if($Value -lt $Min -or $Value -gt $Max){throw 'Integer bound'}
}
function Test-SupervisionPolicyToken($Value,[string[]]$Allowed){
    foreach($token in $Allowed){if([string]::Equals($Value,$token,[StringComparison]::Ordinal)){return $true}};return $false
}
function Assert-SupervisionPolicyToken($Value,[string[]]$Allowed){
    if((Get-SupervisionPolicyType $Value) -ne [string]){throw 'Exact string required'}
    Assert-SupervisionPolicyPassive $Value;if(-not (Test-SupervisionPolicyToken $Value $Allowed)){throw 'Ordinal token required'}
}
function Assert-SupervisionPolicyBool($Value){if((Get-SupervisionPolicyType $Value) -ne [bool]){throw 'Exact bool required'};Assert-SupervisionPolicyPassive $Value}
function Get-SupervisionPolicyActors($Observation){
    switch -CaseSensitive ($Observation.Kind){
        progress{return @('C')}
        time{return @('L','W')}
        cleanup{return @('L','W')}
        fault{
            switch -CaseSensitive ($Observation.Fault){
                'read-error'{return @('C')} 'capacity-exceeded'{return @('C')}
                'supervisor-lost'{return @('L')} 'control-lost'{return @('L','W')}
                default{return @('W')}
            }
        }
        default{return @('W')}
    }
}
function Add-SupervisionPolicyBounded([long]$Start,[long]$Amount,[long]$Bound){
    if($Start -ge ($Bound-$Amount)){return $Bound};return [long]($Start+$Amount)
}
function Set-SupervisionPolicyWindows($Decision,[long]$At){
    if($null -ne $Decision.NaturalUntilMs){return}
    $s=[Math]::Min($At,$Decision.DeadlineMs)
    $Decision.NaturalUntilMs=Add-SupervisionPolicyBounded $s 500 ($Decision.DeadlineMs-6000)
    $Decision.CleanupUntilMs=Add-SupervisionPolicyBounded $Decision.NaturalUntilMs 5000 ($Decision.DeadlineMs-1000)
    $Decision.PublishUntilMs=Add-SupervisionPolicyBounded $Decision.CleanupUntilMs 1000 $Decision.DeadlineMs
}
function Set-SupervisionPolicyFailure($Result,$Context,[string]$Reason,[int]$Index,[long]$At,[string]$Capture='refused'){
    if($Context.Fatal){return}
    $Context.Fatal=$true;$Result.Reason=$Reason;$Result.FailureObservationIndex=$Index
    $Result.State=if($Context.Armed){'Aborted'}else{'Refused'}
    $d=$Result.Decision;$d.WouldStopOwnedSession=$Context.Armed;$d.WouldEvidenceComplete=$false;$d.WouldPositiveTechnicalCandidate=$false
    if($Capture -ceq 'timed-out'){$Result.State='Aborted';$d.WouldCaptureStatus='timed-out';$d.WouldCleanupStatus='unknown';$d.WouldNaturalCompletion=$false}
    elseif($Capture -cne 'preserve' -and $d.WouldCaptureStatus -cne 'complete'){$d.WouldCaptureStatus=$Capture}
    Set-SupervisionPolicyWindows $d $At
}
function Get-OfflineSupervisionDecision([object]$Scenario){
    Set-StrictMode -Version Latest
    $result=[pscustomobject]@{Schema=1;Mode='policy-only';EvidenceClass='synthetic';Evaluated=$false;Reason='invalid-scenario';State='Refused';FailureObservationIndex=$null;Decision=$null;OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false;LiveActionsPerformed=$false;CleanupStatus='unknown'}
    try{
        Assert-SupervisionPolicyShape $Scenario @('Schema','Mode','SessionLabel','ClockLabel','StartMs','Observations')
        Assert-SupervisionPolicyInteger $Scenario.Schema 1 1;Assert-SupervisionPolicyToken $Scenario.Mode @('policy-only')
        Assert-SupervisionPolicyToken $Scenario.ClockLabel @('synthetic-ms')
        if((Get-SupervisionPolicyType $Scenario.SessionLabel) -ne [string]){throw 'Label string required'}
        if(-not [regex]::IsMatch($Scenario.SessionLabel,'\A[A-Za-z0-9][A-Za-z0-9_-]{0,63}\z',[Text.RegularExpressions.RegexOptions]::CultureInvariant)){throw 'Label bound'}
        Assert-SupervisionPolicyInteger $Scenario.StartMs 0 ([long]::MaxValue-120000)
        $inputObservations=$Scenario.Observations
        if((Get-SupervisionPolicyType $inputObservations) -ne [object[]]){throw 'Exact zero-based object array required'}
        Assert-SupervisionPolicyPassive $inputObservations
        foreach($p in $inputObservations.PSObject.Properties){if($p.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty){throw 'Decorated array'}}
        if($inputObservations.Length -gt 4096){throw 'Observation bound'}
        $start=[long]$Scenario.StartMs;$session=[string]$Scenario.SessionLabel;$clock=[string]$Scenario.ClockLabel
        $result.Reason='invalid-observation';$observations=New-Object 'System.Collections.Generic.List[object]'
        $script:profileStats.Active='ValidationLoop'
$script:profileStats.OpenType=$script:profileStats.TypeTicks;$script:profileStats.OpenExt=$script:profileStats.ExtTicks;$script:profileStats.OpenShapeOwn=$script:profileStats.ShapeOwnTicks
$script:profileStats.ValidationCalls++
$script:profileStats.OpenAt=[Diagnostics.Stopwatch]::GetTimestamp()
Profile-Snapshot 'ValidationStart'
try{for($i=0;$i -lt $inputObservations.Length;$i++){
            $result.FailureObservationIndex=$i;$o=$inputObservations[$i]
            $type=Get-SupervisionPolicyType $o;if($null -eq $type -or $type.FullName -cne 'System.Management.Automation.PSCustomObject'){throw 'Plain observation required'}
            foreach($p in $o.PSObject.Properties){if($p.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty){throw 'Passive observation required'}}
            if($null -eq $o.PSObject.Properties['Kind']){throw 'Missing kind'}
            Assert-SupervisionPolicyToken $o.Kind @('arm','release','progress','control','time','fault','cleanup','persist')
            $extra=@(switch -CaseSensitive ($o.Kind){
                arm{'LauncherOwnsControl';'SupervisorOwnsControl';'ContainerBound';'HandleInheritanceBlocked';'ReadersReady'}
                release{'InputBindingMatches'}
                progress{'StdoutRetainedBytes';'StderrRetainedBytes';'DataEventCount';'StdoutEOF';'StderrEOF';'ExitObserved';'ExitCode'}
                control{'SummaryCount';'MaxSummaryBytes';'QueueDepth'}
                fault{'Fault'} cleanup{'Verified';'NaturalCompletion';'CountsReconciled'} persist{'ReadbackMatches';'Classification'}
            })
            Assert-SupervisionPolicyShape $o (@('Sequence','SessionLabel','ClockLabel','AtMs','Actor','Kind')+$extra)
            Assert-SupervisionPolicyInteger $o.Sequence;Assert-SupervisionPolicyInteger $o.AtMs
            foreach($name in @('SessionLabel','ClockLabel')){if((Get-SupervisionPolicyType $o.$name) -ne [string] -or $o.$name.Length -gt 64){throw 'Bounded declaration required'}}
            Assert-SupervisionPolicyToken $o.Actor @('L','W','C')
            switch -CaseSensitive ($o.Kind){
                arm{foreach($name in $extra){Assert-SupervisionPolicyBool $o.$name}}
                release{Assert-SupervisionPolicyBool $o.InputBindingMatches}
                progress{
                    foreach($name in @('StdoutRetainedBytes','StderrRetainedBytes','DataEventCount')){Assert-SupervisionPolicyInteger $o.$name}
                    foreach($name in @('StdoutEOF','StderrEOF','ExitObserved')){Assert-SupervisionPolicyBool $o.$name}
                    if($o.ExitObserved){Assert-SupervisionPolicyInteger $o.ExitCode ([int]::MinValue) ([int]::MaxValue)}elseif($null -ne $o.ExitCode){throw 'Unknown exit must be null'}
                }
                control{foreach($name in $extra){Assert-SupervisionPolicyInteger $o.$name}}
                fault{Assert-SupervisionPolicyToken $o.Fault @('read-error','capacity-exceeded','control-overflow','identity-drift','control-lost','supervisor-lost','launcher-lost')}
                cleanup{foreach($name in $extra){Assert-SupervisionPolicyBool $o.$name}}
                persist{Assert-SupervisionPolicyBool $o.ReadbackMatches;Assert-SupervisionPolicyToken $o.Classification @('clean','refused','unknown')}
            }
            $copy=[ordered]@{};foreach($p in $o.PSObject.Properties){$copy[$p.Name]=$p.Value};$observations.Add([pscustomobject]$copy)
        
$script:profileStats.ValidationRows++
if(($script:profileStats.ValidationRows%128) -eq 0){Profile-Snapshot 'ValidationProgress'}
}}finally{
$profLoopElapsed=[Diagnostics.Stopwatch]::GetTimestamp()-$script:profileStats.OpenAt
$script:profileStats.ValidationTicks+=$profLoopElapsed
$script:profileStats.ValidationOwnTicks+=($profLoopElapsed-($script:profileStats.TypeTicks-$script:profileStats.OpenType)-($script:profileStats.ExtTicks-$script:profileStats.OpenExt)-($script:profileStats.ShapeOwnTicks-$script:profileStats.OpenShapeOwn))
$script:profileStats.Active='None'
}
Profile-Snapshot 'ValidationDone'
        $result.Reason='internal-error';$result.FailureObservationIndex=$null
        $d=[pscustomobject]@{DeadlineMs=[long]($start+120000);StopAtMs=[long]($start+113500);NaturalUntilMs=$null;CleanupUntilMs=$null;PublishUntilMs=$null;LastAtMs=$start;LastSequence=$null;StdoutRetainedBytes=[long]0;StderrRetainedBytes=[long]0;DataEventCount=[long]0;ControlSummaryCount=[long]0;ControlSummaryByteMax=[long]0;ControlQueueDepth=[long]0;WouldCaptureStatus='incomplete';WouldCleanupStatus='unknown';WouldStopOwnedSession=$false;WouldNaturalCompletion=$false;WouldEvidenceComplete=$false;WouldPositiveTechnicalCandidate=$false;WouldExitObserved=$false;WouldExitCode=$null;WouldClassification='unknown'}
        $result.Evaluated=$true;$result.State='Prepared';$result.Reason='trace-evaluated';$result.Decision=$d
        $context=[pscustomobject]@{Fatal=$false;Armed=$false;StdoutEOF=$false;StderrEOF=$false}
        $script:profileStats.Active='DecisionLoop'
$script:profileStats.OpenType=$script:profileStats.TypeTicks;$script:profileStats.OpenExt=$script:profileStats.ExtTicks;$script:profileStats.OpenShapeOwn=$script:profileStats.ShapeOwnTicks
$script:profileStats.DecisionCalls++
$script:profileStats.OpenAt=[Diagnostics.Stopwatch]::GetTimestamp()
Profile-Snapshot 'DecisionStart'
try{for($i=0;$i -lt $observations.Count;$i++){
            $o=$observations[$i];$at=[long]$o.AtMs;$reason=$null
            if(-not [string]::Equals($o.SessionLabel,$session,[StringComparison]::Ordinal) -or -not [string]::Equals($o.ClockLabel,$clock,[StringComparison]::Ordinal)){$reason='binding-mismatch'}
            elseif(-not (Test-SupervisionPolicyToken $o.Actor @(Get-SupervisionPolicyActors $o))){$reason='wrong-actor'}
            elseif($o.Sequence -ne $i){$reason='sequence-mismatch'}
            elseif($at -lt $d.LastAtMs){$reason='time-went-backward'}
            if($null -ne $reason){Set-SupervisionPolicyFailure $result $context $reason $i $at;continue}
            $d.LastAtMs=$at;$d.LastSequence=[long]$o.Sequence
            if($context.Fatal){
                if($o.Kind -ceq 'cleanup' -and $o.Verified -and $o.CountsReconciled){
                    if($at -lt $d.CleanupUntilMs -and $at -lt $d.DeadlineMs){$d.WouldCleanupStatus='verified';$d.WouldNaturalCompletion=($o.NaturalCompletion -and $at -lt $d.NaturalUntilMs)}
                }
                if($o.Kind -ceq 'persist'){$d.WouldClassification=$o.Classification}
                continue
            }
            if($result.State -ceq 'Persisted'){Set-SupervisionPolicyFailure $result $context trailing-after-terminal $i $at;continue}
            if($at -ge $d.DeadlineMs){Set-SupervisionPolicyFailure $result $context total-deadline $i $at timed-out;continue}
            if($result.State -ceq 'Stopping'){
                if($at -ge $d.CleanupUntilMs){Set-SupervisionPolicyFailure $result $context phase-deadline $i $at;continue}
                if($at -ge $d.NaturalUntilMs){Set-SupervisionPolicyFailure $result $context natural-window-expired $i $at;continue}
            }elseif($result.State -ceq 'Sealed'){
                if($at -ge $d.PublishUntilMs){Set-SupervisionPolicyFailure $result $context phase-deadline $i $at;continue}
            }elseif($at -ge $d.StopAtMs){Set-SupervisionPolicyFailure $result $context capture-cutoff $i $at timed-out;continue}
            switch -CaseSensitive ($o.Kind){
                arm{
                    if($result.State -cne 'Prepared'){$reason='unexpected-transition'}
                    elseif(-not ($o.LauncherOwnsControl -and $o.SupervisorOwnsControl -and $o.ContainerBound -and $o.HandleInheritanceBlocked -and $o.ReadersReady)){$reason='prerequisites-unmet'}
                    else{$context.Armed=$true;$result.State='Armed'}
                }
                release{
                    if($result.State -cne 'Armed'){$reason='unexpected-transition'}elseif(-not $o.InputBindingMatches){$reason='prerequisites-unmet'}else{$result.State='Running'}
                }
                progress{
                    if($result.State -cne 'Running'){$reason='unexpected-transition';break}
                    if($o.StdoutRetainedBytes -gt 1048576 -or $o.StderrRetainedBytes -gt 1048576 -or $o.DataEventCount -gt 4096){$reason='capacity-exceeded';break}
                    if($o.StdoutRetainedBytes -lt $d.StdoutRetainedBytes -or $o.StderrRetainedBytes -lt $d.StderrRetainedBytes -or $o.DataEventCount -lt $d.DataEventCount -or ($context.StdoutEOF -and -not $o.StdoutEOF) -or ($context.StderrEOF -and -not $o.StderrEOF) -or ($d.WouldExitObserved -and (-not $o.ExitObserved -or $o.ExitCode -ne $d.WouldExitCode))){$reason='progress-regressed';break}
                    $d.StdoutRetainedBytes=[long]$o.StdoutRetainedBytes;$d.StderrRetainedBytes=[long]$o.StderrRetainedBytes;$d.DataEventCount=[long]$o.DataEventCount
                    $context.StdoutEOF=$o.StdoutEOF;$context.StderrEOF=$o.StderrEOF;$d.WouldExitObserved=$o.ExitObserved;$d.WouldExitCode=$o.ExitCode
                    if($context.StdoutEOF -and $context.StderrEOF -and $d.WouldExitObserved){$d.WouldCaptureStatus='complete';$result.State='Stopping';Set-SupervisionPolicyWindows $d $at}
                }
                control{
                    if($o.SummaryCount -gt 16 -or $o.MaxSummaryBytes -gt 1024 -or $o.QueueDepth -gt 16){$reason='capacity-exceeded';break}
                    if($o.SummaryCount -lt $d.ControlSummaryCount -or $o.MaxSummaryBytes -lt $d.ControlSummaryByteMax){$reason='progress-regressed';break}
                    $d.ControlSummaryCount=[long]$o.SummaryCount;$d.ControlSummaryByteMax=[long]$o.MaxSummaryBytes;$d.ControlQueueDepth=[long]$o.QueueDepth
                }
                fault{
                    $capture=if($o.Fault -ceq 'read-error'){'fault'}else{'preserve'}
                    Set-SupervisionPolicyFailure $result $context $o.Fault $i $at $capture
                }
                cleanup{
                    if($result.State -cne 'Stopping'){$reason='unexpected-transition'}
                    elseif(-not ($o.Verified -and $o.CountsReconciled)){$reason='cleanup-unverified'}
                    else{$d.WouldCleanupStatus='verified';$d.WouldNaturalCompletion=$o.NaturalCompletion;$result.State='Sealed'}
                }
                persist{
                    if($result.State -cne 'Sealed'){$reason='unexpected-transition';break}
                    $d.WouldClassification=$o.Classification
                    if(-not $o.ReadbackMatches){$reason='persist-unverified';break}
                    $result.State='Persisted';$d.WouldEvidenceComplete=$true
                    $d.WouldPositiveTechnicalCandidate=($d.WouldCaptureStatus -ceq 'complete' -and $d.WouldCleanupStatus -ceq 'verified' -and $d.WouldNaturalCompletion -and $d.WouldExitObserved -and $d.WouldExitCode -eq 0 -and $d.WouldClassification -ceq 'clean')
                }
                time{}
            }
            if($null -ne $reason){Set-SupervisionPolicyFailure $result $context $reason $i $at}
        
$script:profileStats.DecisionRows++
if(($script:profileStats.DecisionRows%512) -eq 0){Profile-Snapshot 'DecisionProgress'}
}}finally{
$profLoopElapsed=[Diagnostics.Stopwatch]::GetTimestamp()-$script:profileStats.OpenAt
$script:profileStats.DecisionTicks+=$profLoopElapsed

$script:profileStats.Active='None'
}
Profile-Snapshot 'DecisionDone'
        return $result
    }catch{
        if($result.Evaluated){$result.Evaluated=$false;$result.Reason='internal-error';$result.State='Refused';$result.Decision=$null;$result.FailureObservationIndex=$null}
        return $result
    }
}

# END INSTRUMENTED MODULE COPY
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
function Profile-Assert($r,$Expected){P1-Need (P1-Equal $r $Expected)}
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=[Security.Principal.WindowsPrincipal]::new($identity)
    Write-Output ('CONTEXT: '+$identity.Name+' admin='+$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
    Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
    $script:p1Case='P1ArrayAndIntegerBounds'
    $many=New-Object object[] 4096;for($i=0;$i -lt 4096;$i++){$many[$i]=P1-Observation $i 0 time}
    $scenario=New-P1Scenario -Observations $many;P1-Need ($scenario.Observations.Length -eq 4096)
    $expected=P1-Want @{} @{LastAtMs=0;LastSequence=4095}
    $script:profileStats=@{TypeTicks=0L;TypeCalls=0L;ExtTicks=0L;ExtCalls=0L;ShapeTicks=0L;ShapeOwnTicks=0L;ShapeCalls=0L;ValidationTicks=0L;ValidationOwnTicks=0L;ValidationCalls=0L;DecisionTicks=0L;DecisionCalls=0L;ValidationRows=0;DecisionRows=0;Active='None';OpenAt=0L;OpenType=0L;OpenExt=0L;OpenShapeOwn=0L}
    $script:profileSeq=0;$script:profileCalls=0;$script:profileClock=[Diagnostics.Stopwatch]::StartNew()
    if($script:profileCalls -ne 0){throw 'PROFILE ERROR: call bound'};$script:profileCalls=1
    Profile-Snapshot Begin
    $evaluation=[Diagnostics.Stopwatch]::StartNew()
    $result=Get-OfflineSupervisionDecision $scenario
    $evaluation.Stop()
    Profile-Assert $result $expected
    Profile-Snapshot Complete
    Write-Output 'PASS: EvaluationProfile'
    $summary=[pscustomobject]@{Schema=1;ProductCalls=1;ValidatedRows=$script:profileStats.ValidationRows;DecisionRows=$script:profileStats.DecisionRows;EvaluationMs=[Math]::Round($evaluation.Elapsed.TotalMilliseconds,3);OriginalAssertionPassed=$true;OriginalModuleRun=$false;OriginalTestsRun=$false;ProductChanged=$false}
    Write-Output ('PROFILE-SUMMARY: '+($summary|ConvertTo-Json -Compress))
    Write-Output 'LIMIT: independent evaluation profile only; instrumentation overhead, no formal acceptance or budget renewal.'
}catch{
    [Console]::Out.WriteLine('PROBE-ERROR: '+([pscustomobject]@{Message=$_.Exception.Message}|ConvertTo-Json -Compress));exit 1
}
