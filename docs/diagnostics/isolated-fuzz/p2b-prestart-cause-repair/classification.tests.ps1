# Same test in both phases: no fabricated ErrorRecord or lifecycle observation.
function Invoke-LSClassificationProbe {
    $s=LS-New classification
    $r=LS-Step $s 0 $null
    $v=Get-LSView $s
    LS-Need ($r.Accepted -and $null -eq $r.Observation -and $null -eq $r.Ack -and $r.Error -is [Management.Automation.ErrorRecord]) 'real classification step error'
    LS-Need ($r.Error.FullyQualifiedErrorId -ceq 'Host target budget exhausted before Process creation' -and $r.Error.Exception.Message -ceq 'Host target budget exhausted before Process creation') 'exact unchanged budget gate'
    LS-Need ($v.FailureLatched -and $v.Reserved -eq 1 -and $v.StartCalls -eq 0 -and $v.AckIssued -eq 0 -and $v.Sinks.Start.LogicalIOAttempts -eq 1 -and $v.Sinks.Out0.LogicalIOAttempts -eq 0) 'classification no process or open'
    $same=[object]::ReferenceEquals($r.Error,$script:HostPrimary)
    $sessionSame=[object]::ReferenceEquals($r.Error,$v.PrimaryError)
    LS-Result UnexpectedStepPrimarySame ($same -and $sessionSame) ([ordered]@{PrimarySame=$same;SessionPrimarySame=$sessionSame;StartCalls=$v.StartCalls;AckIssued=$v.AckIssued})
    $script:ProbeObserved=$r.Error;$script:ProbePrimarySame=$same;$script:ProbeReady=$true
    $script:CurrentGroup='ClassificationProbeAbort'
    [Console]::Out.WriteLine('PROBE: ClassificationBeforeAssertion')
    LS-Need $false 'ClassificationProbeAbort'
}
