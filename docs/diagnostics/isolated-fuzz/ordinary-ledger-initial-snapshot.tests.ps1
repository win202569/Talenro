# Break caught: failed/current-unattempted shutdown polls reuse prior runtime facts.
# Real ledger/session, injected fake API only; no native entry or workload.
. (Join-Path $PSScriptRoot 'ordinary-ledger-fixture.ps1')
foreach($scenario in @('error234','truncated','overrun')){
    $f=[OrdinaryLedgerFixture]::Create();$f.Ledger.Poll(1000)
    Check-Ordinary ($f.Ledger.Report.Admitted -eq 1 -and $f.Ledger.Report.LastQuery.Returned -eq 1 -and -not $f.Ledger.Report.Incomplete) ('RuntimeSeed-'+$scenario)
    $f.Clock.Now=200;$f.Api.Candidates=[UInt64[]]@(3)
    switch($scenario){error234{$f.Api.Partial=$true} truncated{$f.Api.Truncated=$true} overrun{$f.Api.AdvanceOnQuery=501}}
    $session=New-Object OrdinaryLedgerSession($f.Ledger,[IntPtr]1,$f.Api,$f.Clock)
    $result=$session.Finish($true,$false,200)
    Check-Ordinary ($result.InitialMembers.Count -eq 0) ('NoStaleInitialMembers-'+$scenario)
    Check-Ordinary ($null -ne $result.InitialMemberQuery -and $result.InitialMemberQuery.StartTimestamp -ge 200) ('CurrentFailedQueryRecorded-'+$scenario)
    switch($scenario){
        error234{Check-Ordinary (-not $result.InitialMemberQuery.Succeeded -and $result.InitialMemberQuery.Error -eq 234) 'Raw234Preserved'}
        truncated{Check-Ordinary ($result.InitialMemberQuery.Assigned -eq 2 -and $result.InitialMemberQuery.Returned -eq 1) 'RawTruncationPreserved'}
        overrun{Check-Ordinary ($result.InitialMemberQuery.EndTimestamp -eq 701 -and $result.EndTimestamp -eq 701) 'RawOverrunPreserved'}
    }
    Check-Ordinary (-not $result.EvidenceComplete -and -not $result.NaturalShutdownVerified -and $f.Ledger.Report.Incomplete -and $f.Api.OpenCalls -eq 1 -and @($f.Api.Calls|Where-Object {$_ -ceq 'list'}).Count -eq 2) ('StickyNoRetry-'+$scenario)
    $f.Ledger.Dispose()
    Check-Ordinary ($f.Api.ReleaseCalls -eq 1) ('RuntimeOwnerReleasedOnce-'+$scenario)
}
$f=[OrdinaryLedgerFixture]::Create();$f.Ledger.Poll(1000)
$f.Clock.Now=200;$f.Api.Candidates=[UInt64[]]@();$f.Api.Lists.Enqueue([UInt64[]]@(3));$f.Api.Total=3;$f.Api.Exited=$true
$session=New-Object OrdinaryLedgerSession($f.Ledger,[IntPtr]1,$f.Api,$f.Clock);$result=$session.Finish($true,$false,200)
Check-Ordinary ($result.InitialMembers.Count -eq 1 -and $result.InitialMembers[0] -eq 3 -and $result.InitialMemberQuery.StartTimestamp -eq 200) 'SuccessfulCurrentInitialSnapshotPreserved'
Check-Ordinary ($result.EvidenceComplete -and $f.Ledger.Report.Admitted -eq 2 -and $f.Api.ReleaseCalls -eq 2 -and $result.FinalMembers.Count -eq 0) 'FinalPollDoesNotOverwriteInitialSnapshot'
$f=[OrdinaryLedgerFixture]::Create();$f.Ledger.Poll(1000);$f.Clock.Now=200;$f.Api.AccountingError=234
$session=New-Object OrdinaryLedgerSession($f.Ledger,[IntPtr]1,$f.Api,$f.Clock);$result=$session.Finish($true,$false,200)
Check-Ordinary ($null -eq $result.InitialMemberQuery -and $result.InitialMembers.Count -eq 0 -and $result.InitialAccounting.Error -eq 234 -and -not $result.EvidenceComplete) 'UnattemptedInitialListNotReplacedByRuntime'
$f.Ledger.Dispose()
Write-Output 'LIMIT: deterministic fake API/clock only; no real process or native adapter.'
