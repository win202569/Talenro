# DIAGNOSTIC ENTRY CLONE: PREPARED, NOT RUN, NOT A PRODUCT REPLACEMENT.
# Six sparse phase markers only. Fresh entry/body/history can perturb timing.
function Get-OfflineSupervisionDecisionAttributionProbe([object]$Scenario){
    [Console]::Out.WriteLine('PRODUCTPHASE: RootSetup START ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
    Set-StrictMode -Version Latest
    $result=[pscustomobject]@{Schema=1;Mode='policy-only';EvidenceClass='synthetic';Evaluated=$false;Reason='invalid-scenario';State='Refused';FailureObservationIndex=$null;Decision=$null;OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false;LiveActionsPerformed=$false;CleanupStatus='unknown'}
    try{
        $Metadata=New-SupervisionPolicyMetadata
        Assert-SupervisionPolicyShape $Scenario @('Schema','Mode','SessionLabel','ClockLabel','StartMs','Observations') -Metadata $Metadata -ShapeKey root
        Assert-SupervisionPolicyInteger $Scenario.Schema 1 1 -Metadata $Metadata;Assert-SupervisionPolicyToken $Scenario.Mode @('policy-only') -Metadata $Metadata
        Assert-SupervisionPolicyToken $Scenario.ClockLabel @('synthetic-ms') -Metadata $Metadata
        if((Get-SupervisionPolicyType $Scenario.SessionLabel -Metadata $Metadata) -ne [string]){throw 'Label string required'}
        if(-not [regex]::IsMatch($Scenario.SessionLabel,'\A[A-Za-z0-9][A-Za-z0-9_-]{0,63}\z',[Text.RegularExpressions.RegexOptions]::CultureInvariant)){throw 'Label bound'}
        Assert-SupervisionPolicyInteger $Scenario.StartMs 0 ([long]::MaxValue-120000) -Metadata $Metadata
        $inputObservations=$Scenario.Observations
        if((Get-SupervisionPolicyType $inputObservations -Metadata $Metadata) -ne [object[]]){throw 'Exact zero-based object array required'}
        Assert-SupervisionPolicyPassive $inputObservations -Metadata $Metadata
        foreach($p in $inputObservations.PSObject.Properties){if($p.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty){throw 'Decorated array'}}
        if($inputObservations.Length -gt 4096){throw 'Observation bound'}
        $start=[long]$Scenario.StartMs;$session=[string]$Scenario.SessionLabel;$clock=[string]$Scenario.ClockLabel
        $result.Reason='invalid-observation';$observations=New-Object 'System.Collections.Generic.List[object]'
        [Console]::Out.WriteLine('PRODUCTPHASE: RootSetup END ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
        [long]$vcDiagValidationTicks=0;[long]$vcDiagCopyTicks=0;[long]$vcDiagCompletedRows=0;[long]$vcDiagVStart=0;[long]$vcDiagCStart=0;[long]$vcDiagCEnd=0
        [Console]::Out.WriteLine('PRODUCTPHASE: ValidationCopy START ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
        for($i=0;$i -lt $inputObservations.Length;$i++){
            $vcDiagVStart=[Diagnostics.Stopwatch]::GetTimestamp()
            $result.FailureObservationIndex=$i;$o=$inputObservations[$i]
            $type=Get-SupervisionPolicyType $o -Metadata $Metadata;if($null -eq $type -or $type.FullName -cne 'System.Management.Automation.PSCustomObject'){throw 'Plain observation required'}
            foreach($p in $o.PSObject.Properties){if($p.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty){throw 'Passive observation required'}}
            if($null -eq $o.PSObject.Properties['Kind']){throw 'Missing kind'}
            Assert-SupervisionPolicyToken $o.Kind @('arm','release','progress','control','time','fault','cleanup','persist') -Metadata $Metadata
            $extra=@(switch -CaseSensitive ($o.Kind){
                arm{'LauncherOwnsControl';'SupervisorOwnsControl';'ContainerBound';'HandleInheritanceBlocked';'ReadersReady'}
                release{'InputBindingMatches'}
                progress{'StdoutRetainedBytes';'StderrRetainedBytes';'DataEventCount';'StdoutEOF';'StderrEOF';'ExitObserved';'ExitCode'}
                control{'SummaryCount';'MaxSummaryBytes';'QueueDepth'}
                fault{'Fault'} cleanup{'Verified';'NaturalCompletion';'CountsReconciled'} persist{'ReadbackMatches';'Classification'}
            })
            Assert-SupervisionPolicyShape $o (@('Sequence','SessionLabel','ClockLabel','AtMs','Actor','Kind')+$extra) -Metadata $Metadata -ShapeKey $o.Kind
            Assert-SupervisionPolicyInteger $o.Sequence -Metadata $Metadata;Assert-SupervisionPolicyInteger $o.AtMs -Metadata $Metadata
            foreach($name in @('SessionLabel','ClockLabel')){if((Get-SupervisionPolicyType $o.$name -Metadata $Metadata) -ne [string] -or $o.$name.Length -gt 64){throw 'Bounded declaration required'}}
            Assert-SupervisionPolicyToken $o.Actor @('L','W','C') -Metadata $Metadata
            switch -CaseSensitive ($o.Kind){
                arm{foreach($name in $extra){Assert-SupervisionPolicyBool $o.$name -Metadata $Metadata}}
                release{Assert-SupervisionPolicyBool $o.InputBindingMatches -Metadata $Metadata}
                progress{
                    foreach($name in @('StdoutRetainedBytes','StderrRetainedBytes','DataEventCount')){Assert-SupervisionPolicyInteger $o.$name -Metadata $Metadata}
                    foreach($name in @('StdoutEOF','StderrEOF','ExitObserved')){Assert-SupervisionPolicyBool $o.$name -Metadata $Metadata}
                    if($o.ExitObserved){Assert-SupervisionPolicyInteger $o.ExitCode ([int]::MinValue) ([int]::MaxValue) -Metadata $Metadata}elseif($null -ne $o.ExitCode){throw 'Unknown exit must be null'}
                }
                control{foreach($name in $extra){Assert-SupervisionPolicyInteger $o.$name -Metadata $Metadata}}
                fault{Assert-SupervisionPolicyToken $o.Fault @('read-error','capacity-exceeded','control-overflow','identity-drift','control-lost','supervisor-lost','launcher-lost') -Metadata $Metadata}
                cleanup{foreach($name in $extra){Assert-SupervisionPolicyBool $o.$name -Metadata $Metadata}}
                persist{Assert-SupervisionPolicyBool $o.ReadbackMatches -Metadata $Metadata;Assert-SupervisionPolicyToken $o.Classification @('clean','refused','unknown') -Metadata $Metadata}
            }
            $vcDiagCStart=[Diagnostics.Stopwatch]::GetTimestamp()
            $copy=[ordered]@{};foreach($p in $o.PSObject.Properties){$copy[$p.Name]=$p.Value};$observations.Add([pscustomobject]$copy)
            $vcDiagCEnd=[Diagnostics.Stopwatch]::GetTimestamp()
            $vcDiagValidationTicks=[long]($vcDiagValidationTicks+($vcDiagCStart-$vcDiagVStart))
            $vcDiagCopyTicks=[long]($vcDiagCopyTicks+($vcDiagCEnd-$vcDiagCStart))
            $vcDiagCompletedRows++
        }
        [Console]::Out.WriteLine('PRODUCTPHASE: ValidationCopy END ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
        [Console]::Out.WriteLine('ATTRIBUTION: ValidationCopy rows='+$vcDiagCompletedRows.ToString([Globalization.CultureInfo]::InvariantCulture)+' validationTicks='+$vcDiagValidationTicks.ToString([Globalization.CultureInfo]::InvariantCulture)+' copyTicks='+$vcDiagCopyTicks.ToString([Globalization.CultureInfo]::InvariantCulture))
        [Console]::Out.WriteLine('PRODUCTPHASE: Decision START ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
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
        [Console]::Out.WriteLine('PRODUCTPHASE: Decision END ticks='+[Diagnostics.Stopwatch]::GetTimestamp().ToString([Globalization.CultureInfo]::InvariantCulture))
        return $result
    }catch{
        if($result.Evaluated){$result.Evaluated=$false;$result.Reason='internal-error';$result.State='Refused';$result.Decision=$null;$result.FailureObservationIndex=$null}
        return $result
    }
}
