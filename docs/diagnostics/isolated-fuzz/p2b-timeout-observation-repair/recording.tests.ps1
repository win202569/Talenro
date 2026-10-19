function Invoke-LSRecordingProbe {
    $script:CurrentGroup='RecordingProbeAbort'
    $s=LS-New recording
    $r=LS-Read $s Ledger
    $v=Get-LSView $s
    LS-Need (-not $r.Ok -and $r.Error -is [Management.Automation.ErrorRecord]) 'real recording probe file error'
    LS-Need ($r.Error.FullyQualifiedErrorId -ceq 'FileNotFoundException') 'recording probe exact error type'
    LS-Need ($v.StartCalls -eq 0 -and $v.AckIssued -eq 0) 'recording probe no authority'
    Capture-Failure $r.Error
    $script:ProbePrimarySame=[object]::ReferenceEquals($r.Error,$script:HostPrimary)
    $script:ProbeObserved=$r.Error
    $null=Close-LSSession $s Stop
    LS-Need $script:ProbePrimarySame 'probe preserves actual first reference'
    $script:ProbeReady=$true
    $script:CurrentGroup='RecordingProbeAbort'
    [Console]::Out.WriteLine('PROBE: RecordingBeforeAssertion')
    # Actual test assertion, not a fabricated IO/OS ErrorRecord or callback.
    LS-Need $false 'RecordingProbeAbort'
}
