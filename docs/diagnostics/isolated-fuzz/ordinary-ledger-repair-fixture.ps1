# Pure root-only envelope used by review regressions. Never invokes native APIs.
. (Join-Path $PSScriptRoot 'ordinary-ledger-fixture.ps1')
. (Join-Path $PSScriptRoot 'run.ps1')
function New-OrdinaryRootEnvelope {
    $f=[OrdinaryLedgerFixture]::Create();$f.Api.Candidates=[UInt64[]]@();$f.Api.Total=1
    $s=New-Object OrdinaryLedgerSession($f.Ledger,[IntPtr]1,$f.Api,$f.Clock)
    $outcome=$s.Finish($true,$false,100)
    [pscustomobject]@{RunIdentity='run-a';RootId=1;RootAssigned=$true;RootExited=$true;TimedOut=$false;TerminationRequested=$false;CleanupVerified=$true;ActiveAfterCleanup=0;MembersAfterCleanup=@();ActiveBeforeTermination=0;MembersBeforeTermination=@();ShutdownSchema=4;ShutdownStatePolicy='ordinary-owned-ledger-final-exit-v1';LifetimeDiagnosticOnly=$false;ShutdownDiagnosticOnly=$false;LiveSampling=$false;MemberEvidenceTruncated=$false;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;NaturalObservationMs=0;ShutdownElapsedMs=0;MemberIdentityLedger=$f.Ledger.Report;LedgerFinalOutcome=$outcome}
}
function Copy-OrdinaryRepairEnvelope($r){$r|ConvertTo-Json -Depth 16|ConvertFrom-Json}
