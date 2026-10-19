function Invoke-LSSequenceTests {
    $a=LS-New sequence-a
    $b=LS-New sequence-b
    $c=LS-New sequence-output-fault
    $a0=Invoke-LSStep $a 0 $null
    $b0=Invoke-LSStep $b 0 $null
    LS-Need ($a0.Accepted -and $b0.Accepted -and $null -eq $a0.Error -and $null -eq $b0.Error -and $null -ne $a0.Ack -and $null -ne $b0.Ack) 'actual live acknowledgements'
    $saved=[Console]::Out
    $closed=LS-ClosedWriter
    try{[Console]::SetOut($closed);$c0=Invoke-LSStep $c 0 $null}finally{[Console]::SetOut($saved)}
    $cv=Get-LSView $c
    $cnext=Invoke-LSStep $c 1 $c0.Ack
    LS-Need ($c0.Observation.ExitCode -eq 0 -and $c0.Observation.OwnedExit -and $c0.Observation.StreamsComplete) 'C real successful lifecycle before actual output failure'
    $outFault=$cv.FailureLatched -and $null -eq $c0.Ack -and $null -ne $c0.Error -and -not $cnext.Accepted -and $cv.StartCalls -eq 1

    $before=Get-LSView $a
    $rejected=0
    $copy=[pscustomobject]@{Verified=$true;Ordinal=0;Grant=$Grant;Case='sequence-a';FrozenSHA256=$FrozenSHA256}
    foreach($ack in @($null,$copy,($copy | ConvertTo-Json | ConvertFrom-Json),$b0.Ack)){
        LS-Need (-not (Test-LSAdvance $a 1 $ack)) 'invalid ack pre-gate'
        $r=Invoke-LSStep $a 1 $ack
        LS-Need (-not $r.Accepted) 'invalid ack consumer'
        $rejected++
    }
    foreach($ordinal in @(0,2)){
        LS-Need (-not (Test-LSAdvance $a $ordinal $a0.Ack)) 'wrong ordinal'
        LS-Need (-not (Invoke-LSStep $a $ordinal $a0.Ack).Accepted) 'wrong ordinal consumer'
        $rejected++
    }
    $after=Get-LSView $a
    LS-Need ($after.StartCalls -eq $before.StartCalls -and $after.AckConsumed -eq 0 -and (Test-LSAdvance $a 1 $a0.Ack)) 'bad requests do not consume genuine identity'
    LS-Need (-not (Invoke-LSStep $b 1 $b0.Ack).Accepted) 'B fixed budget never creates second target'
    $a1=Invoke-LSStep $a 1 $a0.Ack
    LS-Need ($a1.Accepted -and $null -eq $a1.Error -and $a1.Observation.ExitCode -eq 0 -and $null -eq $a1.Ack) 'positive A1 continuation'
    $replay=Invoke-LSStep $a 1 $a0.Ack
    $av=Close-LSSession $a Finish
    $bv=Close-LSSession $b Stop
    $cv=Close-LSSession $c Stop
    $fresh=LS-New fresh
    [IO.File]::WriteAllText((LS-Path $fresh ledger.json),'{"Outcome":"EXPECTED_GREEN"}')
    $freshAdvance=Test-LSAdvance $fresh 1 $null
    $fv=Close-LSSession $fresh Stop
    LS-Details sequence ([ordered]@{A=$av;B=$bv;C=$cv;Fresh=$fv;A0=$a0.Observation;B0=$b0.Observation;C0=$c0.Observation;A1=$a1.Observation})
    LS-Result LiveAckBindingsAndReplay ($rejected -eq 6 -and -not $replay.Accepted -and $av.AckConsumed -eq 1 -and $av.StartCalls -eq 2 -and $bv.StartCalls -eq 1) ([ordered]@{InvalidRequests=$rejected;AckConsumed=$av.AckConsumed;ReplayAccepted=$replay.Accepted;AStarts=$av.StartCalls;BStarts=$bv.StartCalls})
    LS-Result OutputFailureNoContinuation $outFault ([ordered]@{Failed=$cv.FailureLatched;AckIssued=$cv.AckIssued;StartCalls=$cv.StartCalls;NextAccepted=$cnext.Accepted;OwnedExit=$c0.Observation.OwnedExit;StreamsComplete=$c0.Observation.StreamsComplete})
    LS-Result PositiveOwnedContinuation ($av.Status -ceq 'Closed' -and -not $av.FailureLatched -and $av.AckIssued -eq 1 -and $av.AckConsumed -eq 1) ([ordered]@{Status=$av.Status;Failed=$av.FailureLatched;AckIssued=$av.AckIssued;AckConsumed=$av.AckConsumed;SeamExact=$a1.Observation.SeamExact})
    LS-Result SessionClosedNoResume (-not (Test-LSAdvance $a 1 $a0.Ack) -and -not (Test-LSAdvance $b 1 $b0.Ack) -and -not $freshAdvance -and $fv.StartCalls -eq 0) ([ordered]@{AClosed=$av.Status;BClosed=$bv.Status;CClosed=$cv.Status;FreshAdvance=$freshAdvance;FreshStartCalls=$fv.StartCalls})
}
