# Pure synthetic supervision decisions. No imports, I/O, clocks or live adapters.
function Get-SupervisionPolicyType($Value){if($null -eq $Value){return $null};return [object].GetMethod('GetType').Invoke($Value.PSObject.BaseObject,$null)}
function Assert-SupervisionPolicyPassive($Value){
    if($null -eq $Value){return}
    $type=Get-SupervisionPolicyType $Value
    if($type -eq [int] -or $type -eq [long] -or $type -eq [bool] -or $type -eq [string]){
        # Use the fixed SDK getter so an added member named Members cannot shadow it.
        $view=$Value.psextended
        $members=[Management.Automation.PSMemberSet].GetProperty('Members').GetValue($view.PSObject.BaseObject,$null)
        foreach($member in $members){throw 'Decorated scalar'}
    }
    foreach($p in $Value.PSObject.Properties){
        if($p.MemberType -in @([Management.Automation.PSMemberTypes]::ScriptProperty,[Management.Automation.PSMemberTypes]::CodeProperty)){throw 'Executable descriptor'}
        if(([StringComparer]::OrdinalIgnoreCase.Equals($p.Name,'Length') -or [StringComparer]::OrdinalIgnoreCase.Equals($p.Name,'Rank')) -and $p.MemberType -ne [Management.Automation.PSMemberTypes]::Property){throw 'Shadow bound'}
    }
}
function Assert-SupervisionPolicyShape($Value,[string[]]$Names){
    $type=Get-SupervisionPolicyType $Value
    if($null -eq $type -or $type.FullName -cne 'System.Management.Automation.PSCustomObject'){throw 'Plain object required'}
    $props=@($Value.PSObject.Properties);if($props.Count -ne $Names.Count){throw 'Exact fields required'}
    foreach($p in $props){
        $found=$false;foreach($name in $Names){if([string]::Equals($p.Name,$name,[StringComparison]::Ordinal)){$found=$true;break}}
        if(-not $found -or $p.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty){throw 'Exact passive fields required'}
        Assert-SupervisionPolicyPassive $p.Value
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
        for($i=0;$i -lt $inputObservations.Length;$i++){
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
        }
        $result.Reason='internal-error';$result.FailureObservationIndex=$null
        $d=[pscustomobject]@{DeadlineMs=[long]($start+120000);StopAtMs=[long]($start+113500);NaturalUntilMs=$null;CleanupUntilMs=$null;PublishUntilMs=$null;LastAtMs=$start;LastSequence=$null;StdoutRetainedBytes=[long]0;StderrRetainedBytes=[long]0;DataEventCount=[long]0;ControlSummaryCount=[long]0;ControlSummaryByteMax=[long]0;ControlQueueDepth=[long]0;WouldCaptureStatus='incomplete';WouldCleanupStatus='unknown';WouldStopOwnedSession=$false;WouldNaturalCompletion=$false;WouldEvidenceComplete=$false;WouldPositiveTechnicalCandidate=$false;WouldExitObserved=$false;WouldExitCode=$null;WouldClassification='unknown'}
        $result.Evaluated=$true;$result.State='Prepared';$result.Reason='trace-evaluated';$result.Decision=$d
        $context=[pscustomobject]@{Fatal=$false;Armed=$false;StdoutEOF=$false;StderrEOF=$false}
        for($i=0;$i -lt $observations.Count;$i++){
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
        }
        return $result
    }catch{
        if($result.Evaluated){$result.Evaluated=$false;$result.Reason='internal-error';$result.State='Refused';$result.Decision=$null;$result.FailureObservationIndex=$null}
        return $result
    }
}
