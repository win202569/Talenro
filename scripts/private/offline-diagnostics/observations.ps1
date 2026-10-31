function Capture-Failure([Management.Automation.ErrorRecord]$Record){
    if($null -eq $Record){return}
    if($null -eq $script:HostPrimary){$script:HostPrimary=$Record;return}
    if([object]::ReferenceEquals($script:HostPrimary,$Record)){return}
    foreach($prior in $script:HostSecondary){if([object]::ReferenceEquals($prior,$Record)){return}}
    if($script:HostSecondary.Count -lt 8){$script:HostSecondary.Add($Record)}else{$script:HostOverflow=$true}
}
function Capture-StepError($Session,$Result){
    if($null -eq $Result.Error){return}
    # Classification is a test-consumer policy, never a lifecycle/ack setter.
    # No file reads, runtime switches or caller-provided expected-error flags.
    $expected=$false
    try{
        $v=Get-LSView $Session;$o=$Result.Observation;$errorRecord=$Result.Error
        $common=$null -ne $v -and $Result.Accepted -and $null -eq $Result.Rejection -and $null -eq $Result.Ack -and $v.FailureLatched -and [object]::ReferenceEquals($v.PrimaryError,$errorRecord) -and $v.SecondaryErrors.Count -eq 0 -and -not $v.Overflow -and $v.Reserved -eq 1 -and $v.NextOrdinal -eq 0 -and $v.AckIssued -eq 0 -and $v.AckConsumed -eq 0 -and $v.Sinks.Start.State -ceq 'Healthy' -and $v.Sinks.Start.LogicalIOAttempts -eq 1
        if($Grant -ceq 'entry-migration-core' -and $v.Root -ceq (Join-Path $packet ('profiles/'+$Grant+'/preconsume'))){
            $failedSinks=@($v.Sinks.PSObject.Properties | Where-Object {$_.Value.State -ceq 'Failed'} | ForEach-Object {$_.Name})
            $expected=$Result.Accepted -and $null -eq $Result.Rejection -and $null -eq $Result.Ack -and $null -eq $o -and $v.FailureLatched -and [object]::ReferenceEquals($v.PrimaryError,$errorRecord) -and $v.SecondaryErrors.Count -eq 0 -and -not $v.Overflow -and $v.Reserved -eq 1 -and $v.NextOrdinal -eq 0 -and $v.AckIssued -eq 0 -and $v.AckConsumed -eq 0 -and $v.StartCalls -eq 0 -and ($failedSinks -join ',') -ceq 'Start' -and $v.Sinks.Start.LogicalIOAttempts -eq 1 -and $v.Sinks.Out0.LogicalIOAttempts -eq 0 -and $errorRecord.Exception.GetType().FullName -ceq 'System.Management.Automation.MethodInvocationException' -and $errorRecord.FullyQualifiedErrorId -ceq 'UnauthorizedAccessException'
        }
        if($common -and $errorRecord.Exception.GetType().FullName -ceq 'System.Management.Automation.MethodInvocationException' -and $null -ne $o){
            $case=if($v.Root -ceq (Join-Path $packet ('profiles/'+$Grant+'/prestart'))){'prestart'}elseif($v.Root -ceq (Join-Path $packet ('profiles/'+$Grant+'/receipt-fault'))){'receipt-fault'}elseif($v.Root -ceq (Join-Path $packet ('profiles/'+$Grant+'/sequence-output-fault'))){'sequence-output-fault'}else{''}
            $failedSinks=@($v.Sinks.PSObject.Properties | Where-Object {$_.Value.State -ceq 'Failed'} | ForEach-Object {$_.Name})
            if($Grant -ceq 'entry-migration-owned' -and $case -ceq 'prestart'){
                $expected=$errorRecord.FullyQualifiedErrorId -ceq 'UnauthorizedAccessException' -and ($failedSinks -join ',') -ceq 'Out0' -and $v.Sinks.Out0.LogicalIOAttempts -eq 1 -and $v.Sinks.Err0.LogicalIOAttempts -eq 0 -and $v.Sinks.Receipt0.LogicalIOAttempts -eq 0 -and $v.StartCalls -eq 0 -and -not $o.Started -and $o.Association -ceq 'Unassociated' -and $null -eq $o.PID -and $o.KillCalls -eq 0 -and $o.WaitCalls -eq 0 -and -not $o.TerminationRequested -and -not $o.TimedOut -and -not $o.OwnedExit -and -not $o.StreamsComplete
            }elseif($Grant -ceq 'entry-migration-owned' -and $case -ceq 'receipt-fault'){
                $expected=$errorRecord.FullyQualifiedErrorId -ceq 'UnauthorizedAccessException' -and ($failedSinks -join ',') -ceq 'Receipt0' -and $v.Sinks.Receipt0.LogicalIOAttempts -eq 1 -and $v.StartCalls -eq 1 -and $o.Started -and $o.Association -ceq 'Associated' -and $o.PID -gt 0 -and $o.OwnedExit -and $o.StreamsComplete -and $o.TerminationRequested -and -not $o.TimedOut -and $o.KillCalls -eq 1 -and $o.WaitCalls -eq 1 -and $o.ExitCode -eq -1
            }elseif($Grant -ceq 'entry-migration-live' -and $case -ceq 'sequence-output-fault'){
                $expected=$errorRecord.FullyQualifiedErrorId -ceq 'ObjectDisposedException' -and ($failedSinks -join ',') -ceq 'ConsoleOut' -and $v.Sinks.ConsoleOut.LogicalIOAttempts -eq 1 -and $v.StartCalls -eq 1 -and $o.Started -and $o.Association -ceq 'Associated' -and $o.PID -gt 0 -and $o.OwnedExit -and $o.StreamsComplete -and -not $o.TerminationRequested -and -not $o.TimedOut -and $o.KillCalls -eq 0 -and $o.ExitCode -eq 0
            }
        }
    }catch{
        # A classifier fault cannot displace the already-observed actual cause.
        Capture-Failure $Result.Error
        Capture-Failure $_
        return
    }
    if(-not $expected){Capture-Failure $Result.Error}
}
