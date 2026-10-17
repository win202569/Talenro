# Breaks caught: dropping absent members, accepting unknown exits/count gaps,
# resetting member budgets, accounting before release and late identity substitution.
. (Join-Path $PSScriptRoot 'ordinary-ledger-fixture.ps1')
if(-not('OrdinaryLedgerSession' -as [type])){throw 'ASSERTION FAILED: planned ordinary finalization session missing'}
function Finish-Ordinary($f){$s=New-Object OrdinaryLedgerSession($f.Ledger,[IntPtr]1,$f.Api,$f.Clock);$s.Finish($true,$false,$f.Clock.Now)}
function New-ExitFixture {
    $f=[OrdinaryLedgerFixture]::Create();$f.Ledger.Poll(1000);$f.Api.Candidates=[UInt64[]]@();$f.Api.Exited=$true;return $f
}
$f=New-ExitFixture;$r=Finish-Ordinary $f
Check-Ordinary ($r.EvidenceComplete -and $r.NaturalShutdownVerified -and $f.Ledger.Report.Entries[0].Final.Signaled -and $f.Api.OpenCalls -eq 1) 'AbsentAtRootStillVerified'
Check-Ordinary ($f.Api.Calls.IndexOf('close:900002') -lt $f.Api.Calls.IndexOf('accounting') -and $f.Api.ReleaseCalls -eq 1) 'ReleaseBeforeAccounting'
$f=New-ExitFixture;$f.Api.Total=3;$f.Api.Lists.Enqueue([UInt64[]]@(2,3));$r=Finish-Ordinary $f
Check-Ordinary ($r.EvidenceComplete -and $f.Ledger.Report.Admitted -eq 2 -and $f.Api.OpenCalls -eq 2) 'LateCandidate'
$f=New-ExitFixture;$f.Api.Exited=$false;$r=Finish-Ordinary $f
Check-Ordinary (-not $r.EvidenceComplete -and -not $r.NaturalShutdownVerified) 'EmptyButRunning';$f.Ledger.Dispose()
$f=New-ExitFixture;$f.Api.Total=3;$r=Finish-Ordinary $f
Check-Ordinary (-not $r.CountReconciled -and -not $r.EvidenceComplete) 'MissedMemberCount';$f.Ledger.Dispose()
$f=New-ExitFixture;$f.Api.Lists.Enqueue([UInt64[]]@());$f.Api.Lists.Enqueue([UInt64[]]@());$f.Api.Lists.Enqueue([UInt64[]]@(3));$r=Finish-Ordinary $f
Check-Ordinary (-not $r.EvidenceComplete -and $f.Api.OpenCalls -eq 1 -and $f.Api.ReleaseCalls -eq 1) 'NewMemberAfterSeal'
$f=[OrdinaryLedgerFixture]::Create();$f.Api.Candidates=[UInt64[]]@(2,3);$f.Ledger.Poll(1000);$f.Api.Candidates=[UInt64[]]@();$f.Api.Total=3;$f.Api.Exited=$true;$f.Api.AdvanceOnRead=300;$r=Finish-Ordinary $f
Check-Ordinary (-not $r.EvidenceComplete -and $f.Clock.Now -eq 600 -and $f.Api.Waits[$f.Api.Waits.Count-1] -eq 200) 'Shared500ms';$f.Ledger.Dispose()
$f=New-ExitFixture;$s=New-Object OrdinaryLedgerSession($f.Ledger,[IntPtr]1,$f.Api,$f.Clock);$r=$s.Finish($true,$false,[long]::MaxValue)
Check-Ordinary (-not $r.EvidenceComplete -and $f.Ledger.Report.Incomplete) 'DeadlineOverflow';$f.Ledger.Dispose()
$f=New-ExitFixture;$f.Clock.Now=50;$r=Finish-Ordinary $f
Check-Ordinary (-not $r.EvidenceComplete -and $f.Ledger.Report.Incomplete) 'BackwardClock';$f.Ledger.Dispose()
$f=New-ExitFixture;$f.Ledger.ObserveFinal(1000,$true)
Check-Ordinary ($f.Ledger.Report.Entries[0].Final.TerminationRequested -and $f.Ledger.Report.Incomplete) 'ForcedExitRefused';$f.Ledger.Dispose()
$f=[OrdinaryLedgerFixture]::Create();$f.Api.Exited=$true;$f.Ledger.Poll(1000);$f.Api.Exited=$false;$f.Ledger.ObserveFinal(1000,$false)
Check-Ordinary ($f.Ledger.Report.Ambiguous -and $f.Ledger.Report.Incomplete) 'ExitedThenRunning';$f.Ledger.Dispose()
$f=New-ExitFixture;$f.Api.ThrowClose=$true;$r=Finish-Ordinary $f;$f.Ledger.Dispose()
Check-Ordinary (-not $r.EvidenceComplete -and $f.Api.ReleaseCalls -eq 1 -and $f.Ledger.Report.ReleaseFailures -eq 1) 'CloseFailureSticky'
foreach($mode in @('MissingExit','MissingImage')){$f=New-ExitFixture;$f.Api.$mode=$true;$r=Finish-Ordinary $f;Check-Ordinary ($r.EvidenceComplete -eq ($mode -ceq 'MissingImage')) $mode;$f.Ledger.Dispose()}
Write-Output 'LIMIT: fake finalization only; no native adapter invoked.'
