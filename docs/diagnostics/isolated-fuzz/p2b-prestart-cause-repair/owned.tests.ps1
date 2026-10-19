function Invoke-LSOwnedTests {
    $s=LS-New prestart
    $null=[IO.Directory]::CreateDirectory((LS-Path $s stdout-0.bin))
    $r=LS-Step $s 0 $null
    $o=$r.Observation
    $v=Get-LSView $s
    LS-Need ($null -ne $o -and $v.FailureLatched) 'real prestart open failure'
    LS-Details prestart ([ordered]@{View=$v;Observation=$o})
    LS-Result PreStartFaultNoStart (-not $o.Started -and $v.StartCalls -eq 0) ([ordered]@{Started=$o.Started;StartCalls=$v.StartCalls;Failed=$v.FailureLatched})
    LS-Result UnassociatedNoKillWait ($o.Association -ceq 'Unassociated' -and $null -eq $o.PID -and $o.KillCalls -eq 0 -and $o.WaitCalls -eq 0) ([ordered]@{Association=$o.Association;PID=$o.PID;KillCalls=$o.KillCalls;WaitCalls=$o.WaitCalls})
    $null=Close-LSSession $s Stop

    $s=LS-New receipt-fault
    $null=[IO.Directory]::CreateDirectory((LS-Path $s receipt-0.json))
    $r=LS-Step $s 0 $null
    $v=Get-LSView $s
    $o=$r.Observation
    LS-Need ($null -ne $o -and $o.Started -and $o.OwnedExit -and $o.StreamsComplete -and $null -ne $r.Error) 'fault must run, fail receipt and clean exact object'
    LS-Need ($o.Association -ceq 'Associated' -and $o.TerminationRequested -and -not $o.TimedOut) 'fault exact owned cleanup'
    LS-Details receipt-fault ([ordered]@{View=$v;Observation=$o})
    $facts=[ordered]@{Started=$o.Started;OwnedExit=$o.OwnedExit;StreamsComplete=$o.StreamsComplete;ReceiptAttempts=$v.Sinks.Receipt0.LogicalIOAttempts;SecondaryCount=$v.SecondaryErrors.Count;AckIssued=$v.AckIssued}
    $good=$facts.ReceiptAttempts -eq 1 -and $facts.SecondaryCount -eq 0 -and $facts.AckIssued -eq 0
    if($Grant -ceq 'owned-red'){
        LS-Result ReceiptSinkMustNotBeRetried $good $facts
        $null=Close-LSSession $s Stop
        return
    }
    LS-Result ReceiptFaultNoRetry $good $facts
    $same=[object]::ReferenceEquals($r.Error,$v.PrimaryError)
    LS-Result FirstCauseSurvivesCleanup ($same -and $o.OwnedExit -and $o.StreamsComplete) ([ordered]@{PrimarySame=$same;OwnedExit=$o.OwnedExit;StreamsComplete=$o.StreamsComplete})
    $null=Close-LSSession $s Stop

    foreach($case in @('normal','timeout')){
        $s=LS-New $case
        $r=LS-Step $s 0 $null
        $o=$r.Observation
        LS-Need ($null -ne $o -and $null -eq $r.Error -and $o.Started -and $o.Association -ceq 'Associated' -and $o.OwnedExit -and $o.StreamsComplete) ('owned '+$case)
        $out=[IO.File]::ReadAllText((LS-Path $s stdout-0.bin))
        $err=[IO.File]::ReadAllText((LS-Path $s stderr-0.bin))
        $streams=$out -ceq 'P2B-CONTAINMENT-OUT' -and $err -ceq 'P2B-CONTAINMENT-ERR'
        # Independent byte-prefix oracle, not the controller's helper.
        $prefix=$true
        foreach($pair in @(@('stdout','P2B-CONTAINMENT-OUT'),@('stderr','P2B-CONTAINMENT-ERR'))){
            [byte[]]$bytes=[IO.File]::ReadAllBytes((LS-Path $s ($pair[0]+'-0.bin')))
            [byte[]]$expected=[Text.Encoding]::ASCII.GetBytes($pair[1])
            if($bytes.Length -gt 19){$prefix=$false}
            for($j=0;$j -lt $bytes.Length;$j++){if($j -ge 19 -or $bytes[$j] -ne $expected[$j]){$prefix=$false}}
        }
        $facts=[ordered]@{Started=$o.Started;OwnedExit=$o.OwnedExit;StreamsComplete=$o.StreamsComplete;TimedOut=$o.TimedOut;TerminationRequested=$o.TerminationRequested;ExitCode=$o.ExitCode;StreamsMatch=$(if($case -ceq 'timeout'){$prefix}else{$streams})}
        LS-Details $case ([ordered]@{View=(Get-LSView $s);Observation=$o})
        if($case -ceq 'normal'){LS-Result NormalExit ($streams -and -not $o.TimedOut -and -not $o.TerminationRequested -and $o.ExitCode -eq 0) $facts}
        else{LS-Result Timeout ($prefix -and $o.TimedOut -and $o.TerminationRequested -and $o.ExitCode -eq -1) $facts}
        $null=Close-LSSession $s Stop
    }
}
