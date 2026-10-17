# Breaks caught: duplicate opens, ownership bypass, retry after87, overflow admissions,
# cross-run read adoption and double release after a failed Close.
. (Join-Path $PSScriptRoot 'ordinary-ledger-fixture.ps1')
$f=[OrdinaryLedgerFixture]::Create();$f.Api.Candidates=[ulong[]]@(1,2,2);$f.Ledger.Poll(1000);$f.Ledger.Poll(1000)
Check-Ordinary ($f.Api.OpenCalls -eq 1 -and $f.Ledger.Report.Admitted -eq 1 -and $f.Ledger.Report.Entries[0].Membership.Owned -and $f.Ledger.Report.Entries[0].Initial.RunIdentity -ceq 'run-a') 'AcquisitionOwnsOnce'
$f.Ledger.SealAndRelease();$f.Ledger.Dispose();Check-Ordinary ($f.Api.ReleaseCalls -eq 1) 'OwnHandleReleasedOnce'
$f=[OrdinaryLedgerFixture]::Create();$f.Api.OpenError=87;$f.Ledger.Poll(1000);$f.Ledger.Poll(1000)
Check-Ordinary ($f.Api.OpenCalls -eq 1 -and $f.Api.MembershipCalls -eq 0 -and $f.Ledger.Report.Incomplete -and $f.Ledger.Report.Entries[0].Open.Error -eq 87) 'Open87Sticky';$f.Ledger.Dispose()
$f=[OrdinaryLedgerFixture]::Create();$f.Api.Owned=$false;$f.Ledger.Poll(1000);$f.Ledger.Dispose()
Check-Ordinary ($f.Ledger.Report.Incomplete -and $f.Ledger.Report.Admitted -eq 0 -and $f.Api.ReleaseCalls -eq 1) 'ForeignReleased'
$f=[OrdinaryLedgerFixture]::Create();$f.Api.ReadRunIdentity='other-run';$f.Ledger.Poll(1000)
Check-Ordinary ($f.Ledger.Report.Incomplete -and $f.Ledger.Report.Ambiguous) 'CrossRunIdentity';$f.Ledger.Dispose()
$f=[OrdinaryLedgerFixture]::Create();$f.Api.Candidates=[ulong[]](2..34);$f.Ledger.Poll(1000)
Check-Ordinary ($f.Api.OpenCalls -eq 32 -and $f.Ledger.Report.Attempts -eq 32 -and $f.Ledger.Report.CapExhausted -and $f.Ledger.Report.Incomplete) 'Cap32';$f.Ledger.Dispose()
foreach($candidate in @([ulong]0,[ulong]4294967296)){$f=[OrdinaryLedgerFixture]::Create();$f.Api.Candidates=[ulong[]]@($candidate);$f.Ledger.Poll(1000);Check-Ordinary ($f.Ledger.Report.Incomplete -and $f.Api.OpenCalls -eq 0) 'InvalidCandidate';$f.Ledger.Dispose()}
$f=[OrdinaryLedgerFixture]::Create();$f.Ledger.Poll(1000);$f.Api.ThrowClose=$true;$f.Ledger.Dispose();$f.Ledger.Dispose()
Check-Ordinary ($f.Api.ReleaseCalls -eq 1 -and $f.Ledger.Report.ReleaseFailures -eq 1 -and $f.Ledger.Report.Incomplete) 'CloseFailureOnce'
Write-Output 'LIMIT: fake adapter only; no native target invoked.'
